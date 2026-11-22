package table_test

import . "github.com/salvatoresaba/simpleui"

import (
	"os"
	"strings"
	"testing"
)

// table_api_test.go covers the core `table` module: it is always available
// (no import), its calls are synchronous, and the generated app carries the
// runtime that fills/reads the table identified by its name attribute.

func TestTableApiTypechecksAndGenerates(t *testing.T) {
	src := "state rows: array[array[string]] = []\n" +
		"\n" +
		"fn fill(data: any) -> null {\n" +
		"\tn, err = table.display(\"users\", data)\n" +
		"\tif err != null { rows = [[\"error\", err]] }\n" +
		"}\n" +
		"\n" +
		"fn readAll() -> null {\n" +
		"\tr, e1 = table.data(\"users\")\n" +
		"\th, e2 = table.header(\"users\")\n" +
		"\tpart, e3 = table.select(\"users\", [0, 2])\n" +
		"\tblock, e3b = table.slice(\"users\", 0, 10)\n" +
		"\tn2, e3c = table.importText(\"users\", \"a,b\\n1,2\", \"csv\")\n" +
		"\tcsv, e3d = table.exportText(\"users\", \"csv\")\n" +
		"\tbody, e4 = table.json(\"users\")\n" +
		"\tn, e5 = table.clear(\"users\")\n" +
		"\trows = r\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\ttable name: \"users\" {\n" +
		"\t\tthead {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"ID\" head: true\n" +
		"\t\t\t\ttcell \"Name\" head: true\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t\ttbody {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"1\"\n" +
		"\t\t\t\ttcell \"Alice\"\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
	prog, pdl := ParseSource(src, "table_api.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check (table module must need no import and no async): %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"const __SUI_TBL",
		"table_display:",
		"table_data:",
		"table_header:",
		"table_select:",
		"table_slice:",
		"table_importtext:",
		"table_exporttext:",
		"table_json:",
		"table_clear:",
		"__SUI_TBL.afterRender();",
		`name="users"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated app.js missing %q", want)
		}
	}
	if strings.Contains(js, "await __SUI_STD.table_display") {
		t.Errorf("table.display must be synchronous (no await)")
	}
}

func TestTableApiWrongArity(t *testing.T) {
	src := "fn bad() -> null {\n\tn, e = table.display(\"users\")\n}\n" +
		"view main {\n\tp \"x\"\n}\n"
	prog, pdl := ParseSource(src, "bad.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	dl := Check(prog)
	if !dl.HasErrors() {
		t.Fatalf("expected an arity error for table.display with one argument")
	}
	if !strings.Contains(dl.String(), "expects 2 arguments") {
		t.Errorf("unexpected diagnostics: %v", dl.String())
	}
}

// TestTableDataExample loads the shipped example so it cannot rot.
func TestTableDataExample(t *testing.T) {
	raw, err := os.ReadFile(RepoPath("examples/table_data.sui"))
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	prog, pdl := ParseSource(string(raw), "examples/table_data.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse example: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check example: %v", dl.String())
	}
}

// TestSpreadsheetTableExample keeps the XLSX/ODS demo app compilable.
func TestSpreadsheetTableExample(t *testing.T) {
	raw, err := os.ReadFile(RepoPath("examples/spreadsheet_table/app.sui"))
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	prog, pdl := ParseSource(string(raw), "examples/spreadsheet_table/app.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse example: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check example: %v", dl.String())
	}
}
