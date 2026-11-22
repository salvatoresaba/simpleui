package core

import (
	"os"
	"strings"
	"testing"
)

func resetKnownStateNames() {
	for n := range KnownStateNames {
		delete(KnownStateNames, n)
	}
}

func PipelineWeb(source string, file string) (bool, string, string, string, DiagList) {
	resetKnownStateNames()
	var dl DiagList
	if err := os.MkdirAll("Test/testdata", 0o755); err != nil {
		dl.Add(file, 1, 1, "cannot create Test/testdata dir: "+err.Error())
		return false, "", "", "", dl
	}
	if err := os.WriteFile(file, []byte(source), 0o644); err != nil {
		dl.Add(file, 1, 1, "cannot write fixture: "+err.Error())
		return false, "", "", "", dl
	}
	prog, perrs := LoadProgram(file)
	mergeDiags(&dl, perrs)
	if !dl.HasErrors() {
		cerrs := Check(prog)
		mergeDiags(&dl, cerrs)
	}
	if dl.HasErrors() {
		_ = os.Remove(file)
		return false, "", "", "", dl
	}
	appJs := GenerateJs(prog)
	indexHtml := GenerateIndexHtml(prog)
	css := GenerateCss(prog)
	_ = os.Remove(file)
	return true, appJs, indexHtml, css, dl
}

func TestPipelineCounter(t *testing.T) {
	data, err := os.ReadFile("examples/counter.sui")
	if err != nil {
		t.Fatalf("cannot read examples/counter.sui: %v", err)
	}
	ok, appJs, indexHtml, css, errs := PipelineWeb(string(data), "Test/testdata/tmp_counter.sui")
	if !ok {
		errs.Print()
		t.Errorf("counter pipeline did not succeed")
	}
	if !strings.Contains(appJs, "setCounter") {
		t.Errorf("app.js missing setCounter")
	}
	if !strings.Contains(appJs, `__sui_react("counter")`) {
		t.Errorf("app.js missing __sui_react binding")
	}
	if !strings.Contains(appJs, "e3.onclick = increment;") {
		t.Errorf("app.js missing onclick wiring for the increment button")
	}
	if !strings.Contains(appJs, "__sui_render();") {
		t.Errorf("app.js missing __sui_render()")
	}
	if !strings.Contains(indexHtml, "<title>Counter App</title>") {
		t.Errorf("index.html missing counter title")
	}
	if !strings.Contains(css, ".sui-view") {
		t.Errorf("style.css missing .sui-view")
	}
	if !strings.Contains(css, ".row") {
		t.Errorf("style.css missing .row")
	}
}

func TestPipelineImportMerge(t *testing.T) {
	data, err := os.ReadFile("Test/testdata/entry.sui")
	if err != nil {
		t.Fatalf("cannot read Test/testdata/entry.sui: %v", err)
	}
	ok, appJs, _, _, errs := PipelineWeb(string(data), "Test/testdata/tmp_import_entry.sui")
	if !ok {
		errs.Print()
		t.Errorf("import merge pipeline did not succeed")
	}
	if !strings.Contains(appJs, "function run(") {
		t.Errorf("app.js missing entry function run")
	}
	if !strings.Contains(appJs, "function scale(") {
		t.Errorf("app.js missing imported function scale")
	}
}

func TestPipelineImportDuplicate(t *testing.T) {
	data, err := os.ReadFile("Test/testdata/dup_entry.sui")
	if err != nil {
		t.Fatalf("cannot read Test/testdata/dup_entry.sui: %v", err)
	}
	ok, _, _, _, errs := PipelineWeb(string(data), "Test/testdata/tmp_dup_entry.sui")
	if ok {
		t.Errorf("duplicate-symbol fixture should not compile")
	}
	if !errs.HasErrors() {
		t.Errorf("duplicate-symbol fixture should report errors")
	}
}

func TestPipelineBadType(t *testing.T) {
	src := "state badnum = 0\nfn f() -> null { badnum = \"s\" }\n"
	ok, _, _, _, errs := PipelineWeb(src, "Test/testdata/tmp_badtype.sui")
	if ok {
		t.Errorf("bad-type program (int state assigned a string) should not compile")
	}
	if !errs.HasErrors() {
		t.Errorf("bad-type program should report errors")
	}
}

func TestPipelineStdOutsideAsync(t *testing.T) {
	src := "import { http } from \"std\"\nfn f() -> null { a, b = http.get(\"u\") }\n"
	ok, _, _, _, errs := PipelineWeb(src, "Test/testdata/tmp_std_sync.sui")
	if ok {
		t.Errorf("std call outside async should not compile")
	}
	if !errs.HasErrors() {
		t.Errorf("std call outside async should report errors")
	}
}

func TestPipelineAsyncJsonTuples(t *testing.T) {
	src := "import { http, json } from \"std\"\n" +
		"state data = 0\n" +
		"async fn fetchAll() -> null {\n" +
		"	a, e = http.get(\"/x\")\n" +
		"	if e == null {\n" +
		"		d, je = json.parse(a)\n" +
		"		data = d\n" +
		"	}\n" +
		"}\n"
	ok, appJs, _, _, errs := PipelineWeb(src, "Test/testdata/tmp_async.tmp")
	if !ok {
		errs.Print()
		t.Errorf("async+json pipeline did not succeed")
	}
	if !strings.Contains(appJs, "await __SUI_STD.http_get(") {
		t.Errorf("app.js missing http_get tuple call")
	}
	if !strings.Contains(appJs, "await __SUI_STD.json_parse(") {
		t.Errorf("app.js missing json_parse tuple call")
	}
	if !strings.Contains(appJs, "const __sui_r = await __SUI_STD.http_get(") {
		t.Errorf("app.js missing first tuple binding const __sui_r")
	}
	if !strings.Contains(appJs, "const __sui_r2 = await __SUI_STD.json_parse(") {
		t.Errorf("app.js missing second tuple binding const __sui_r2")
	}
}

func TestPipelineViewBindings(t *testing.T) {
	src := "state a = 1\nstate b = \"x\"\nview main {\n\tp \"{a} {b}\"\n}\n"
	ok, appJs, _, _, errs := PipelineWeb(src, "Test/testdata/tmp_refs.tmp")
	if !ok {
		errs.Print()
		t.Errorf("view-bindings fixture did not compile")
	}
	if n := strings.Count(appJs, `__sui_bindView("`); n != 2 {
		t.Errorf("expected 2 __sui_bindView lines, got %v", n)
	}
	if !strings.Contains(appJs, `__sui_bindView("a", "main")`) {
		t.Errorf("missing binding for state a")
	}
	if !strings.Contains(appJs, `__sui_bindView("b", "main")`) {
		t.Errorf("missing binding for state b")
	}
	if err := os.MkdirAll("Test/testdata", 0o755); err != nil {
		t.Fatalf("cannot create Test/testdata: %v", err)
	}
	if err := os.WriteFile("Test/testdata/tmp_refs.tmp", []byte(src), 0o644); err != nil {
		t.Fatalf("cannot write fixture: %v", err)
	}
	prog, perrs := LoadProgram("Test/testdata/tmp_refs.tmp")
	if perrs.HasErrors() {
		perrs.Print()
		t.Fatalf("reload of bindings fixture failed")
	}
	_ = Check(prog)
	refs := []string{}
	for i := 0; i < len(prog.views); i++ {
		if prog.views[i].name == "main" {
			refs = ViewStateRefs(&prog.views[i])
		}
	}
	if len(refs) != 2 || refs[0] != "a" || refs[1] != "b" {
		t.Errorf("ViewStateRefs(main) = %v, want [a b]", refs)
	}
	_ = os.Remove("Test/testdata/tmp_refs.tmp")
}

func TestAstJsonParse(t *testing.T) {
	src := "state a = 1\nview main {\n\tp \"hi\"\n}\n"
	j, errs := AstJsonFromSource(src, "Test/testdata/parse_fixture.sui")
	if errs.HasErrors() {
		errs.Print()
		t.Errorf("parse JSON should have no diagnostics")
	}
	if !strings.Contains(j, `"name": "main"`) {
		t.Errorf("AST JSON missing view name main")
	}
	if !strings.Contains(j, `"kind": "element"`) {
		t.Errorf("AST JSON missing element node kind")
	}
	if !strings.Contains(j, `"tag": "p"`) {
		t.Errorf("AST JSON missing element tag p")
	}
}

func TestPipelineReservedNames(t *testing.T) {
	src := "state cmd = \"x\"\nfn setCmd(v: string) -> null {\n\tcmd = v\n}\n"
	ok, _, _, _, errs := PipelineWeb(src, "Test/testdata/tmp_reserved.tmp")
	if ok {
		t.Errorf("function colliding with a generated state setter should not compile")
	}
	if !errs.HasErrors() {
		t.Errorf("reserved setter name should report errors")
	}

	src2 := "fn render_thing() -> null {}\n"
	ok2, _, _, _, errs2 := PipelineWeb(src2, "Test/testdata/tmp_reserved2.tmp")
	if ok2 {
		t.Errorf("function with render_ prefix should not compile")
	}
	if !errs2.HasErrors() {
		t.Errorf("reserved render_ name should report errors")
	}
}

func TestExamplesCompile(t *testing.T) {
	examples := []string{
		"examples/monitor.sui",
		"examples/wrapper.sui",
		"examples/chat.sui",
		"examples/todo.sui",
		"examples/motion_features/app.sui",
		"examples/motion_sequence/app.sui",
	}
	for i := 0; i < len(examples); i++ {
		data, err := os.ReadFile(examples[i])
		if err != nil {
			t.Errorf("cannot read %v: %v", examples[i], err)
			continue
		}
		ok, _, _, _, errs := PipelineWeb(string(data), "Test/testdata/tmp_example.tmp")
		if !ok {
			errs.Print()
			t.Errorf("example %v did not compile through the pipeline", examples[i])
		}
	}
}
