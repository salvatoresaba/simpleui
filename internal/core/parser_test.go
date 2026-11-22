package core

import "testing"
import "strings"

// parser_test.go — tests for ParseSource / LoadProgram (Agent A, front-end).

func diagMsgs(dl DiagList) string {
	var sb strings.Builder
	for i := 0; i < len(dl.list); i++ {
		sb.WriteString("|")
		sb.WriteString(dl.list[i].msg)
	}
	return sb.String()
}

func TestParseFullProgram(t *testing.T) {
	src := "state counter: int = 0\n" +
		"fn incr() -> int {\n" +
		"\tif counter < 10 {\n" +
		"\t\tcounter = counter + 1\n" +
		"\t}\n" +
		"\tloop i in 0..5 {\n" +
		"\t\tcounter = counter + i\n" +
		"\t}\n" +
		"\treturn counter\n" +
		"}\n" +
		"async fn load() -> null {\n" +
		"\tbody, err = http.get(\"https://x\")\n" +
		"\tif err == null {\n" +
		"\t\tcounter = body\n" +
		"\t}\n" +
		"}\n" +
		"view main {\n" +
		"\ttitle \"Count {counter}\"\n" +
		"\th1 \"Top\" class: \"big\"\n" +
		"\tbutton \"Go\" onClick: incr\n" +
		"\tlist x in stuff {\n" +
		"\t\tp \"{x}\"\n" +
		"\t}\n" +
		"\tif counter > 5 {\n" +
		"\t\ttext \"big\"\n" +
		"\t} else {\n" +
		"\t\ttext \"small\"\n" +
		"\t}\n" +
		"}\n"
	prog, dl := ParseSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("full program: %v diagnostics: %v", dl.Count(), diagMsgs(dl))
		return
	}
	if len(prog.states) != 1 || prog.states[0].name != "counter" || prog.states[0].typename != "int" {
		t.Errorf("state: got %v states, name %q type %q", len(prog.states), prog.states[0].name, prog.states[0].typename)
	}
	if len(prog.funcs) != 2 {
		t.Errorf("funcs: got %v, want 2", len(prog.funcs))
		return
	}
	var incr *FuncDecl
	var load *FuncDecl
	for i := 0; i < len(prog.funcs); i++ {
		if prog.funcs[i].name == "incr" {
			incr = &prog.funcs[i]
		}
		if prog.funcs[i].name == "load" {
			load = &prog.funcs[i]
		}
	}
	if incr == nil || incr.async {
		t.Errorf("incr missing or async flag wrong")
	}
	if load == nil || !load.async {
		t.Errorf("load missing or async flag wrong")
	}
	if load != nil {
		if len(load.body) != 2 {
			t.Errorf("load body: got %v stmts, want 2", len(load.body))
		} else {
			ta := load.body[0]
			if ta.kind != "assign" || len(ta.targets) != 2 || ta.targets[0] != "body" || ta.targets[1] != "err" {
				t.Errorf("tuple assign targets: got %v", strings.Join(ta.targets, ","))
			}
			if ta.rhs == nil || ta.rhs.kind != "call" || ta.rhs.target == nil || ta.rhs.target.kind != "member" ||
				ta.rhs.target.memb != "get" || len(ta.rhs.args) != 1 {
				t.Errorf("tuple rhs should be http.get(url) call")
			}
		}
	}
	if incr != nil {
		if len(incr.body) != 3 {
			t.Errorf("incr body: got %v stmts, want 3 (if, loop, return)", len(incr.body))
		} else {
			is := incr.body[0]
			if is.kind != "if" || is.cond == nil || is.cond.kind != "binary" || is.cond.op != "<" ||
				len(is.then_) != 1 || is.then_[0].kind != "assign" {
				t.Errorf("incr if statement malformed")
			}
			ls := incr.body[1]
			if ls.kind != "loop" || ls.item != "i" || ls.iterable == nil || ls.iterable.kind != "range" {
				t.Errorf("range loop missing")
			} else if ls.iterable.lo == nil || ls.iterable.hi == nil || ls.iterable.lo.kind != "int" ||
				ls.iterable.lo.i != 0 || ls.iterable.hi.i != 5 {
				t.Errorf("range bounds: got %v..%v", ls.iterable.lo.i, ls.iterable.hi.i)
			}
			rs := incr.body[2]
			if rs.kind != "return" || rs.value == nil || rs.value.kind != "ident" || rs.value.name != "counter" {
				t.Errorf("return stmt malformed")
			}
		}
	}
	if len(prog.views) != 1 {
		t.Errorf("views: got %v, want 1", len(prog.views))
		return
	}
	v := &prog.views[0]
	if v.name != "main" {
		t.Errorf("view name: %q", v.name)
	}
	if len(v.title) != 2 || v.title[0].kind != "text" || v.title[0].text != "Count " ||
		v.title[1].kind != "expr" || v.title[1].expr == nil || v.title[1].expr.name != "counter" {
		t.Errorf("view title parts wrong")
	}
	if len(v.body) != 4 {
		t.Errorf("view body: got %v nodes, want 4", len(v.body))
		return
	}
	if v.body[0].kind != "element" || v.body[0].elem == nil || v.body[0].elem.tag != "h1" {
		t.Errorf("view body[0] should be h1 element")
	} else {
		e := v.body[0].elem
		if len(e.text) != 1 || e.text[0].text != "Top" {
			t.Errorf("h1 text wrong")
		}
		if len(e.attrs) != 1 || e.attrs[0].key != "class" || e.attrs[0].hasExpr || len(e.attrs[0].text) != 1 ||
			e.attrs[0].text[0].text != "big" {
			t.Errorf("h1 class attr wrong")
		}
	}
	if v.body[1].kind != "element" || v.body[1].elem == nil || v.body[1].elem.tag != "button" {
		t.Errorf("view body[1] should be button element")
	} else {
		attrs := v.body[1].elem.attrs
		if len(attrs) != 1 || attrs[0].key != "onClick" || !attrs[0].hasExpr || attrs[0].expr == nil ||
			attrs[0].expr.kind != "ident" || attrs[0].expr.name != "incr" {
			t.Errorf("button onClick attr wrong")
		}
	}
	if v.body[2].kind != "list" || v.body[2].item != "x" || v.body[2].iterable == nil ||
		v.body[2].iterable.kind != "ident" || v.body[2].iterable.name != "stuff" {
		t.Errorf("view list node wrong")
	}
	if v.body[3].kind != "if" || v.body[3].cond == nil || v.body[3].cond.op != ">" ||
		len(v.body[3].then_) != 1 || v.body[3].then_[0].elem == nil || v.body[3].then_[0].elem.tag != "text" ||
		len(v.body[3].else_) != 1 || v.body[3].else_[0].elem == nil || v.body[3].else_[0].elem.tag != "text" {
		t.Errorf("view if/else node wrong")
	}
}

func TestParseInterpolation(t *testing.T) {
	src := "fn f(n: int) -> string {\n\treturn \"a{n}b\"\n}\n"
	prog, dl := ParseSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("interp: %v diagnostics: %v", dl.Count(), diagMsgs(dl))
		return
	}
	rs := prog.funcs[0].body[0]
	if rs.kind != "return" || rs.value == nil || rs.value.kind != "str" {
		t.Errorf("expected str return")
		return
	}
	parts := rs.value.parts
	if len(parts) != 3 {
		t.Errorf("parts: got %v, want 3", len(parts))
		return
	}
	if parts[0].kind != "text" || parts[0].text != "a" ||
		parts[1].kind != "expr" || parts[1].expr == nil || parts[1].expr.kind != "ident" || parts[1].expr.name != "n" ||
		parts[2].kind != "text" || parts[2].text != "b" {
		t.Errorf("interpolation parts wrong: %v %v %v", strings.Join(partKinds(parts), ","), parts[0].text, parts[2].text)
	}
}

func partKinds(ps []StrPart) []string {
	var ks []string
	for i := 0; i < len(ps); i++ {
		ks = append(ks, ps[i].kind)
	}
	return ks
}

func TestParseBracedEscapes(t *testing.T) {
	src := "fn f() -> string {\n\treturn \"lit\\{x\\}done\"\n}\n"
	prog, dl := ParseSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("braced escapes: %v diagnostics: %v", dl.Count(), diagMsgs(dl))
		return
	}
	rs := prog.funcs[0].body[0]
	parts := rs.value.parts
	if len(parts) != 1 || parts[0].kind != "text" || parts[0].text != "lit{x}done" {
		t.Errorf("brace escapes: want one text part %q, got %v parts text %q", "lit{x}done", len(parts), parts[0].text)
	}
}

func TestParseErrors(t *testing.T) {
	// Missing function annotations must be parse diagnostics.
	src := "fn bad(x) -> int {\n\treturn x\n}\n"
	prog, dl := ParseSource(src, "t.sui")
	if dl.Count() == 0 {
		t.Errorf("missing param annotation should be a diagnostic")
	}
	// Statements must end at a newline.
	src2 := "fn bad2() -> int {\n\ta = 1 b = 2\n\treturn a\n}\n"
	_, dl2 := ParseSource(src2, "t.sui")
	if dl2.Count() == 0 {
		t.Errorf("missing newline between statements should be a diagnostic")
	}
	// Unbalanced braces must not panic and must produce a diagnostic.
	src3 := "view v {\n\th1 \"x\"\n"
	_, dl3 := ParseSource(src3, "t.sui")
	if dl3.Count() == 0 {
		t.Errorf("unbalanced braces should be a diagnostic")
	}
	_ = prog
}

func TestLoadProgramMerge(t *testing.T) {
	prog, dl := LoadProgram("Test/testdata/entry.sui")
	if dl.Count() != 0 {
		t.Errorf("merge: %v diagnostics: %v", dl.Count(), diagMsgs(dl))
		return
	}
	if prog.file != "Test/testdata/entry.sui" {
		t.Errorf("prog.file: %q", prog.file)
	}
	names := ""
	count := 0
	for i := 0; i < len(prog.funcs); i++ {
		names = names + prog.funcs[i].name + " "
		if prog.funcs[i].name == "scale" {
			count++
		}
	}
	if count != 1 || len(prog.funcs) != 2 {
		t.Errorf("merged funcs: [%v] want run, scale", names)
	}
	hasView := false
	for i := 0; i < len(prog.views); i++ {
		if prog.views[i].name == "card" {
			hasView = true
		}
	}
	if !hasView {
		t.Errorf("imported view 'card' not merged")
	}
	// The std import must be kept (deduplicated); the file import is folded
	// into the merged decls.
	if len(prog.imports) != 1 || prog.imports[0].kind != "std" {
		t.Errorf("merged imports: got %v entries, want one std import", len(prog.imports))
		return
	}
	if len(prog.imports[0].names) != 1 || prog.imports[0].names[0] != "http" {
		t.Errorf("std import names: got %v", strings.Join(prog.imports[0].names, ","))
	}
}

func TestLoadProgramDuplicate(t *testing.T) {
	prog, dl := LoadProgram("Test/testdata/dup_entry.sui")
	if dl.Count() < 2 {
		t.Errorf("duplicate symbol: got %v diagnostics: %v", dl.Count(), diagMsgs(dl))
	}
	found := false
	for i := 0; i < len(dl.list); i++ {
		if strings.Contains(dl.list[i].msg, "duplicate top-level symbol") {
			found = true
		}
	}
	if !found {
		t.Errorf("no duplicate-symbol message in: %v", diagMsgs(dl))
	}
	_ = prog
}

func TestLoadProgramMissing(t *testing.T) {
	prog, dl := LoadProgram("Test/testdata/missing_entry.sui")
	if dl.Count() == 0 {
		t.Errorf("missing import should be a diagnostic")
		return
	}
	if !strings.Contains(diagMsgs(dl), "cannot read file") {
		t.Errorf("unexpected message: %v", diagMsgs(dl))
	}
	_ = prog
}

func TestLoadProgramCycle(t *testing.T) {
	prog, dl := LoadProgram("Test/testdata/cycle_a.sui")
	if dl.Count() == 0 {
		t.Errorf("import cycle should be a diagnostic")
		return
	}
	if !strings.Contains(diagMsgs(dl), "import cycle") {
		t.Errorf("unexpected message: %v", diagMsgs(dl))
	}
	_ = prog
}

func TestParseMapAndArray(t *testing.T) {
	src := "state m: map[string]int = {\"a\": 1}\nstate arr: array[int] = [1, 2, 3]\n"
	prog, dl := ParseSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("map/array: %v diagnostics: %v", dl.Count(), diagMsgs(dl))
		return
	}
	if prog.states[0].typename != "map[string]int" {
		t.Errorf("map typename: %q", prog.states[0].typename)
	}
	if prog.states[1].typename != "array[int]" {
		t.Errorf("array typename: %q", prog.states[1].typename)
	}
	if prog.states[0].init == nil || prog.states[0].init.kind != "map" || len(prog.states[0].init.keys) != 1 ||
		prog.states[0].init.keys[0] != "a" {
		t.Errorf("map literal wrong")
	}
	if prog.states[1].init == nil || prog.states[1].init.kind != "array" || len(prog.states[1].init.elements) != 3 {
		t.Errorf("array literal wrong")
	}
}

// TestNestedStringInInterpolation covers string literals inside {...}: the
// lexer keeps the outer token together and the parser re-lexes the nested
// expression, including escaped quotes.
func TestNestedStringInInterpolation(t *testing.T) {
	src := `view main {
	p "{upper("hello")}"
	p "{upper("say \"hi\"")}"
}`
	prog, dl := ParseSource(src, "nested.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, ".toUpperCase()") || !strings.Contains(js, "hello") {
		t.Errorf("nested string expression was not generated")
	}
}
