package main

import "fmt"
import "net"
import "net/http"
import "net/url"
import "os"
import "path/filepath"
import "strings"
import "testing"
import "time"

// gui_test.go - the visual design tool server: editor page + in-process
// compile-and-preview endpoint.

func startGui(t *testing.T, keep bool) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	if keep {
		t.Cleanup(func() { _ = ln.Close() })
	}
	var lastHtml string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", guiPage)
	})
	mux.HandleFunc("POST /__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		ok, html, msgs := renderSui(src)
		if ok {
			lastHtml = html
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
	})
	mux.HandleFunc("POST /__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		model, msgs, ok := guiOpen(src)
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		body := StringifyJson(model)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,`+body[1:])
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", lastHtml)
	})
	go func() {
		_ = http.Serve(ln, mux)
	}()
	return "http://" + ln.Addr().String()
}

func TestGuiEditorPage(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatalf("get editor: %v", err)
	}
	body := readAllClose(t, resp)
	for _, want := range []string{"SimpleUI Design Tool", "addEl", "__gui__/render", "generate()", "id=\"prev\"", "__gui__/open", "draggable", "addPageFlow", "saveBlockFlow", "pagetabs", "duplicatePage", "delPageFlow", "savePageTemplateFlow", "filein", "stateslist", "fnslist", "updStateName", "classlist", "renderPalette", "PALETTE", "openPrefs", "applyPrefs", "\"table\"", "\"section\"", "\"list\"", "\"textarea\"", "\"input\"", "listControls"} {
		if !strings.Contains(body, want) {
			t.Errorf("editor page missing %q", want)
		}
	}
}

func TestGuiEditorPreferencesPanel(t *testing.T) {
	for _, want := range []string{
		`id="pref-tab-interface"`,
		`id="pref-tab-editor"`,
		`id="pref-tab-sui"`,
		`id="pref-tab-shortcuts"`,
		`id="pref-tab-history"`,
		`id="pf-tip-enabled"`,
		`id="pf-tip-delay"`,
		`id="pf-tip-place"`,
		`id="pf-shortcut-list"`,
		`id="pf-hist-interval"`,
		`id="pf-hist-keep"`,
		`id="pf-data-dir"`,
		`id="edtip"`,
		`Editor preferences.`,
		`Project preferences.`,
		`EDITOR_SHORTCUTS`,
		`function renderShortcutList`,
		`function clearProjectHistory`,
		`function applyEditorTipPrefs`,
		`editorTipPlace`,
		`historyInterval`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("editor preferences panel missing %q", want)
		}
	}
}

func TestGuiHistorySettingsParsing(t *testing.T) {
	auto, interval, keep := parseGuiHistorySettings(`{}`)
	if !auto || interval != time.Minute || keep != guiVersionKeep {
		t.Fatalf("defaults = %v %v %d", auto, interval, keep)
	}
	auto, interval, keep = parseGuiHistorySettings(`{"historyAuto":false,"historyInterval":120,"historyKeep":12}`)
	if auto || interval != 2*time.Minute || keep != 12 {
		t.Fatalf("custom = %v %v %d", auto, interval, keep)
	}
	_, interval, keep = parseGuiHistorySettings(`{"historyInterval":1,"historyKeep":9999}`)
	if interval != 15*time.Second || keep != 500 {
		t.Fatalf("clamped = %v %d", interval, keep)
	}
}

func TestGuiPruneVersionsKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 8; i++ {
		name := fmt.Sprintf("project-2026010%d-000000.sui", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("view main {}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	guiPruneVersions(dir, 5)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("kept %d versions, want 5", len(entries))
	}
	if _, err := os.Stat(filepath.Join(dir, "project-20260108-000000.sui")); err != nil {
		t.Error("newest version was removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "project-20260101-000000.sui")); !os.IsNotExist(err) {
		t.Error("oldest version was not removed")
	}
}

func TestGuiRenderOk(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.PostForm(base+"/__gui__/render", url.Values{"sui": {"view main {\n\th1 \"Hello\"\n}\n"}})
	if err != nil {
		t.Fatalf("post render: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Errorf("render ok expected:\n%v", body)
	}
	// The preview page must contain the generated renderer.
	prev, err := http.Get(base + "/__gui__/preview")
	if err != nil {
		t.Fatalf("get preview: %v", err)
	}
	pb := readAllClose(t, prev)
	if !strings.Contains(pb, "render_main") || !strings.Contains(pb, "sui-view") || !strings.Contains(pb, "Hello") && !strings.Contains(pb, "__SUI_views") {
		t.Errorf("preview must carry the compiled app:\n%v", pb[:300])
	}
}

func TestGuiRenderError(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.PostForm(base+"/__gui__/render", url.Values{"sui": {"view main {\n\th1 \"{nope}\"\n}\n"}})
	if err != nil {
		t.Fatalf("post render: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":false`) || !strings.Contains(body, "unknown variable nope") {
		t.Errorf("render error diagnostics missing:\n%v", body)
	}
}

// Every root-view element (containers included) must get a unique dfs-order
// data-sui-id, and the preview must embed the click-to-select picker.
func TestGuiRenderPreviewSelection(t *testing.T) {
	ok, html, _ := renderSui("view main {\n\th1 \"Hi\"\n\trow {\n\t\tp \"a\"\n\t\tp \"b\"\n\t}\n}\n")
	if !ok {
		t.Fatal("renderSui must compile")
	}
	for i, want := range []string{
		`data-sui-id="0"`, // h1
		`data-sui-id="1"`, // row
		`data-sui-id="2"`, // p a
		`data-sui-id="3"`, // p b
	} {
		if !strings.Contains(html, want) {
			t.Errorf("preview missing %q (dfs id %d)", want, i)
		}
	}
	if !strings.Contains(html, "postMessage") || !strings.Contains(html, "data-sui-id") {
		t.Errorf("preview must embed the click-to-select picker script")
	}
}

func TestGuiOpenOk(t *testing.T) {
	base := startGui(t, true)
	src := "state steps = 2\nstate name = \"\"\n" +
		"view main {\n" +
		"\ttitle \"My panel\"\n" +
		"\trow class: \"main\" {\n" +
		"\t\tcol {\n" +
		"\t\t\tbutton \"Refresh\" onClick: refresh\n" +
		"\t\t\tinput placeholder: \"Name\" value: name\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n" +
		"fn refresh() -> null {\n}\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("open expected ok:\n%v", body)
	}
	// Recursive tree: row (with children col -> button/input).
	for _, want := range []string{
		`"title":"My panel"`,
		`"kind":"row"`, `"class":"main"`,
		`"kind":"col"`, `"kind":"button"`, `"text":"Refresh"`, `"onClick":"refresh"`,
		`"kind":"input"`, `"placeholder":"Name"`, `"value":"name"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("open tree missing %q in:\n%v", want, body)
		}
	}
	// View model: global states + helper function round-tripped with init/type.
	// (Key order in the JSON object is not guaranteed, so assert per-field.)
	for _, want := range []string{`"steps"`, `"init":"2"`, `"name"`, `"refresh"`, `"params":""`, `"ret":"null"`} {
		if !strings.Contains(body, want) {
			t.Errorf("open view-model missing %q in:\n%v", want, body)
		}
	}
}

// TestGuiOpenSwitchBinding guards the designer round-trip for control value
// bindings. The "checked" attribute used to be dropped in BOTH directions
// (guiElementFrom never read it, attrsOf never emitted it), so a switch bound
// to a state through the inspector silently lost its binding and never
// reflected state changes.
func TestGuiOpenSwitchBinding(t *testing.T) {
	base := startGui(t, true)
	src := "state active = false\n" +
		"view main {\n" +
		"\tswitch \"\" checked: active onChange: toggleActive\n" +
		"\tswitch \"\"" +
		"\n" +
		"\tnumber value: qty min: 0 max: 10 step: 1\n" +
		"}\n" +
		"fn toggleActive() -> null {\n" +
		"\tactive = active == true ? false : true\n" +
		"}\n" +
		"state qty = 0\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("open expected ok:\n%v", body)
	}
	for _, want := range []string{
		`"checked":"active"`, `"onChange":"toggleActive"`,
		`"value":"qty"`, `"min":"0"`, `"max":"10"`, `"step":"1"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("open tree missing %q in:\n%v", want, body)
		}
	}
}

func TestGuiOpenTitleAndState(t *testing.T) {
	base := startGui(t, true)
	src := "state n: int = 5\nview main {\n\ttitle \"T\"\n\th2 \"Count: {n}\"\n}\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	for _, want := range []string{`"title":"T"`, `"name":"n"`, `"type":"int"`, `"init":"5"`, `"text":"Count: {n}"`} {
		if !strings.Contains(body, want) {
			t.Errorf("open missing %q in:\n%v", want, body)
		}
	}
}

func TestGuiOpenComplexExpr(t *testing.T) {
	base := startGui(t, true)
	src := "view main {\n\tstate n = 0\n\th2 \"Count: {n}\"\n}\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("open expected ok:\n%v", body)
	}
	if !strings.Contains(body, `"text":"Count: {n}"`) {
		t.Errorf("interpolated text should keep literal part:\n%v", body)
	}
}

func TestGuiOpenError(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {"view main {\n\th1 \"{missing}\"\n}\n"}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":false`) {
		t.Errorf("open expected error:\n%v", body)
	}
}

// The exporter emits source with title, state, params, nested containers and
// handler references. That shape must always compile.
func TestGuiGeneratedSourceCompiles(t *testing.T) {
	src := "view main {\n" +
		"\ttitle \"My panel\"\n" +
		"\tstate name = \"\"\n" +
		"\tstate steps = 2\n" +
		"\trow class: \"main\" {\n" +
		"\t\tcol {\n" +
		"\t\t\tbutton \"Refresh\" onClick: refresh\n" +
		"\t\t\tinput placeholder: \"Name\" value: name\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n" +
		"fn refresh() -> null {\n}\n" +
		"fn copy(text: string) -> null {\n}\n"
	ok, _, msgs := renderSui(src)
	if !ok {
		t.Errorf("exported shape must compile, got diagnostics: %v", StringifyJson(msgs))
	}
}

// exprSource must round-trip: whatever it emits has to re-parse cleanly.
func TestGuiExprSourceRoundTrip(t *testing.T) {
	exprs := []*Expr{
		{kind: "int", i: 7},
		{kind: "float", f: 2.5},
		{kind: "bool", b: true},
		{kind: "str", parts: []StrPart{{kind: "text", text: "a{bc}"}}},
		{kind: "null"},
		{kind: "ident", name: "x"},
		{kind: "member", obj: &Expr{kind: "ident", name: "fs"}, memb: "read"},
		{kind: "call", target: &Expr{kind: "ident", name: "fn1"}, args: []*Expr{{kind: "int", i: 1}, {kind: "str", parts: []StrPart{{kind: "text", text: "s"}}}}},
		{kind: "array", elements: []*Expr{{kind: "int", i: 1}, {kind: "int", i: 2}}},
		{kind: "map", keys: []string{"a"}, vals: []*Expr{{kind: "int", i: 1}}},
		{kind: "index", obj: &Expr{kind: "ident", name: "arr"}, idx: &Expr{kind: "int", i: 0}},
		{kind: "binary", op: "add", l: &Expr{kind: "int", i: 1}, r: &Expr{kind: "int", i: 2}},
		{kind: "binary", op: "eq", l: &Expr{kind: "ident", name: "a"}, r: &Expr{kind: "ident", name: "b"}},
	}
	for i := 0; i < len(exprs); i++ {
		src := exprSource(exprs[i])
		prog, dl := ParseSource("fn _t() -> null {\n\tlet chk = "+src+"\n}\n", "t.sui")
		if dl.HasErrors() {
			t.Errorf("exprSource(%q) did not parse: %v", src, dl.String())
			continue
		}
		if len(prog.funcs) == 0 {
			t.Errorf("exprSource(%q) produced no function", src)
		}
	}
}

func TestGuiPreviewHasNoRuler(t *testing.T) {
	for _, unwanted := range []string{"sui-ruler", "drawRuler", "rTop", "rLeft"} {
		if strings.Contains(guiPreviewPicker, unwanted) {
			t.Errorf("preview picker still contains removed ruler reference %q", unwanted)
		}
	}
	if !strings.Contains(guiPreviewPicker, "sui-grid") {
		t.Error("preview picker must keep the guide grid")
	}
}
