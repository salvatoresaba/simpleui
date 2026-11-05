package main

import "strings"
import "testing"

// css_test.go - the `css "…"` top-level declaration: raw user CSS stored on
// the program and appended to the generated style.css.

func TestCssDeclarationAndCodegen(t *testing.T) {
	src := "css \".card \\{ background: #111; padding: 12px; \\}\"\n" +
		"css \".btn:hover { opacity: .8; }\"\n" +
		"view main {\n\th1 \"Hi\" class: \"card\"\n}\n"
	prog, dl := ParseSource(src, "css.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if len(prog.css) != 2 || prog.css[0] != ".card { background: #111; padding: 12px; }" || prog.css[1] != ".btn:hover { opacity: .8; }" {
		t.Errorf("prog.css wrong: %q", prog.css)
	}
	css := GenerateCss(prog)
	for _, want := range []string{".card { background: #111; padding: 12px; }", ".btn:hover { opacity: .8; }", ".card { /* generated */ }"} {
		if !strings.Contains(css, want) {
			t.Errorf("style.css missing %q\n%v", want, css)
		}
	}
	dl2 := Check(prog)
	if dl2.HasErrors() {
		t.Errorf("check: %v", dl2.String())
	}
}

func TestCssInvalid(t *testing.T) {
	_, dt := ParseSource("css 42\nview main {\n}\n", "bad.sui")
	if !dt.HasErrors() || !hasMsg(dt, "expected a string after 'css'") {
		t.Errorf("css number should be rejected, got: %v", dt.String())
	}
}

func TestCssThroughGuiOpen(t *testing.T) {
	src := "css \".hero { color: #fff; }\"\nview main {\n\th1 \"T\"\n}\n"
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen failed: %v", StringifyJson(msgs))
	}
	cssStr := asStr(get(model, "css"))
	if !strings.Contains(cssStr, ".hero { color: #fff; }") {
		t.Errorf("gui model css missing: %q", cssStr)
	}
}

// The AST JSON dump must carry the css snippets.
func TestCssInAstJson(t *testing.T) {
	prog, dl := ParseSource("css \".x { a: b; }\"\nview main {\n}\n", "j.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	j := ProgToJson(prog)
	if !strings.Contains(j, `"css": [`) || !strings.Contains(j, ".x { a: b; }") {
		t.Errorf("ProgToJson missing css field:\n%v", j)
	}
}

// The designer's CSS engine endpoints.
func TestGuiCssParse(t *testing.T) {
	css := ".used { color: red; }\n.unused { color: blue; }\n"
	res := guiCssParse(css, `["used"]`)
	if !asBool(get(res, "ok")) {
		t.Fatalf("parse not ok: %v", StringifyJson(res))
	}
	if asStr(get(res, "generated")) == "" {
		t.Errorf("expected generated css")
	}
	probs := get(res, "problems")
	if probs.kind != "arr" || len(probs.arr) == 0 {
		t.Fatalf("expected unused-class problem, got: %v", StringifyJson(probs))
	}
	foundUnused := false
	for i := 0; i < len(probs.arr); i++ {
		if asStr(get(probs.arr[i], "kind")) == "unused" {
			foundUnused = true
		}
	}
	if !foundUnused {
		t.Errorf("expected an 'unused' problem, got: %v", StringifyJson(probs))
	}
}

func TestGuiCssSet(t *testing.T) {
	css := ".card { padding: 4px; }\n"
	res := guiCssSet(css, "card", "padding", "16px")
	if !asBool(get(res, "ok")) {
		t.Fatalf("set not ok")
	}
	out := asStr(get(res, "css"))
	if !strings.Contains(out, "padding: 16px;") {
		t.Errorf("set should update padding:\n%v", out)
	}
	// create a missing rule
	res2 := guiCssSet(css, "hero", "color", "red")
	out2 := asStr(get(res2, "css"))
	if !strings.Contains(out2, ".hero") || !strings.Contains(out2, "color: red;") {
		t.Errorf("set should create .hero rule:\n%v", out2)
	}
	// remove by empty value
	res3 := guiCssSet(out, "card", "padding", "")
	out3 := asStr(get(res3, "css"))
	if strings.Contains(out3, "padding:") {
		t.Errorf("empty value should remove the declaration:\n%v", out3)
	}
}

func asBool(v JVal) bool {
	if v.kind == "bool" {
		return v.b
	}
	return false
}

// The special "body" styling target maps to the bare body element selector.
func TestGuiCssSetBody(t *testing.T) {
	css := ".card { padding: 4px; }\n"
	res := guiCssSetBatch(css, "body", `[{"prop":"background-color","value":"#202020"}]`)
	if !asBool(get(res, "ok")) {
		t.Fatalf("setbatch body not ok: %v", StringifyJson(res))
	}
	out := asStr(get(res, "css"))
	if !strings.Contains(out, "body {") || !strings.Contains(out, "background-color: #202020;") {
		t.Errorf("body rule should be created with the bare selector:\n%v", out)
	}
	if !strings.Contains(out, ".card") {
		t.Errorf("existing rules must be preserved:\n%v", out)
	}
}

// Parameterized views (components) must round-trip: the root view is the one
// not invoked, and component views/invocations are preserved verbatim.
func TestGuiOpenComponentViews(t *testing.T) {
	src := "state likes = 0\n" +
		"fn like() -> null {\n\tlikes += 1\n}\n" +
		"view Card(heading: string, note: string) {\n\trow {\n\t\th2 \"{heading}\"\n\t\tp \"{note}\"\n\t\tbutton \"Like\" onClick: like\n\t}\n}\n" +
		"view main {\n\ttitle \"Components\"\n\th1 \"Parameterized views\"\n\tCard heading: \"First card\" note: \"a reusable view\"\n\tCard heading: \"Second card\" note: \"each instance\"\n\tp \"Likes: {likes}\"\n}\n"
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen failed: %v", StringifyJson(msgs))
	}
	if asStr(get(model, "title")) != "Components" {
		t.Errorf("title = %q, want Components", asStr(get(model, "title")))
	}
	elems := get(model, "elements")
	if elems.kind != "arr" || len(elems.arr) != 4 {
		t.Fatalf("expected 4 root elements, got %v", StringifyJson(elems))
	}
	kinds := []string{asStr(get(elems.arr[0], "kind")), asStr(get(elems.arr[1], "kind")), asStr(get(elems.arr[2], "kind")), asStr(get(elems.arr[3], "kind"))}
	if kinds[0] != "text" || kinds[1] != "Card" || kinds[2] != "Card" || kinds[3] != "text" {
		t.Errorf("root element kinds wrong: %v", kinds)
	}
	// h1/p collapse into rich text nodes with the right block tag + run text.
	if asStr(get(elems.arr[0], "tag")) != "h1" || asStr(get(elems.arr[3], "tag")) != "p" {
		t.Errorf("text block tags wrong: h1=%q p=%q", asStr(get(elems.arr[0], "tag")), asStr(get(elems.arr[3], "tag")))
	}
	// component invocations carry raw source
	if asStr(get(elems.arr[1], "raw")) != "Card heading: \"First card\" note: \"a reusable view\"" {
		t.Errorf("Card invocation raw wrong: %q", asStr(get(elems.arr[1], "raw")))
	}
	if !asBool(get(elems.arr[1], "component")) {
		t.Errorf("Card invocation should be marked component")
	}
	// component view preserved
	comps := get(model, "components")
	if comps.kind != "arr" || len(comps.arr) != 1 {
		t.Fatalf("expected 1 component view, got %v", StringifyJson(comps))
	}
	csrc := asStr(comps.arr[0])
	if !strings.Contains(csrc, "view Card(heading: string, note: string)") || !strings.Contains(csrc, "button \"Like\" onClick: like") {
		t.Errorf("component source wrong:\n%v", csrc)
	}
}

// guiFuncsJson must preserve function bodies so the SUI round-trips.
func TestGuiFuncBodyRoundTrip(t *testing.T) {
	src := "state counter = 0\n" +
		"fn increment() -> null {\n" +
		"\tcounter = counter + 1\n" +
		"\tif counter > 10 {\n" +
		"\t\tcounter = 0\n" +
		"\t}\n" +
		"}\n" +
		"view main {\n\tbutton \"Go\" onClick: increment\n}\n"
	prog, dl := ParseSource(src, "f.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	fn := guiFuncsJson(prog.funcs)
	if fn.kind != "arr" || len(fn.arr) != 1 {
		t.Fatalf("expected 1 func, got %v", StringifyJson(fn))
	}
	body := asStr(get(fn.arr[0], "body"))
	if !strings.Contains(body, "counter = counter + 1") || !strings.Contains(body, "if counter > 10") {
		t.Errorf("function body lost:\n%v", body)
	}
	// re-serialize into a compilable source
	out := "state counter = 0\nfn increment() -> null {\n" + body + "\n}\nview main {\n\tbutton \"Go\" onClick: increment\n}\n"
	_, dl2 := ParseSource(out, "f2.sui")
	if dl2.HasErrors() {
		t.Errorf("re-serialized body does not parse:\n%v", dl2.String())
	}
}

func TestGuiCssSetBatch(t *testing.T) {
	css := ".card { padding: 4px; }\n"
	res := guiCssSetBatch(css, "card", `[{"prop":"-webkit-border-radius","value":"8px"},{"prop":"-moz-border-radius","value":"8px"},{"prop":"border-radius","value":"8px"}]`)
	if !asBool(get(res, "ok")) {
		t.Fatalf("setbatch not ok: %v", StringifyJson(res))
	}
	out := asStr(get(res, "css"))
	for _, want := range []string{"-webkit-border-radius: 8px;", "-moz-border-radius: 8px;", "border-radius: 8px;", "padding: 4px;"} {
		if !strings.Contains(out, want) {
			t.Errorf("setbatch should keep %q:\n%v", want, out)
		}
	}
}

// A table in the source must round-trip to the designer's matrix model.
func TestGuiOpenTableMatrix(t *testing.T) {
	src := "view main {\n" +
		"\ttable {\n" +
		"\t\tthead {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"A\" head: true\n" +
		"\t\t\t\ttcell \"B\" head: true\n" +
		"\t\t\t\ttcell \"C\" head: true\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t\ttbody {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"1\"\n" +
		"\t\t\t\ttcell \"2\" colspan: 2\n" +
		"\t\t\t}\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"x\"\n" +
		"\t\t\t\ttcell \"y\"\n" +
		"\t\t\t\ttcell \"z\"\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen failed: %v", StringifyJson(msgs))
	}
	elems := get(model, "elements")
	if elems.kind != "arr" || len(elems.arr) == 0 {
		t.Fatalf("expected a table element")
	}
	tbl := elems.arr[0]
	if asStr(get(tbl, "kind")) != "table" {
		t.Fatalf("expected table, got %q", asStr(get(tbl, "kind")))
	}
	if asInt(get(tbl, "cols")) != 3 {
		t.Errorf("cols = %d, want 3", asInt(get(tbl, "cols")))
	}
	if asInt(get(tbl, "rows")) != 3 {
		t.Errorf("rows = %d, want 3", asInt(get(tbl, "rows")))
	}
	if asInt(get(tbl, "header")) != 1 {
		t.Errorf("header = %d, want 1", asInt(get(tbl, "header")))
	}
	// cells[1][1] should be "2" with colspan 2
	cells := get(tbl, "cells")
	cellRow := cells.arr[1]
	if asStr(cellRow.arr[1]) != "2" {
		t.Errorf("cells[1][1] = %q, want \"2\"", asStr(cellRow.arr[1]))
	}
	spans := get(tbl, "spans")
	sp := spans.arr[1].arr[1]
	if asInt(get(sp, "cs")) != 2 {
		t.Errorf("spans[1][1].cs = %d, want 2", asInt(get(sp, "cs")))
	}
}
