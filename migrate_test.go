package main

import (
	"strings"
	"testing"
)

func TestLegacyTableParametersMigrate(t *testing.T) {
	src := `view main {
	table name: "legacy" columns: "a,b" editable: true fixed: true {
	}
}`
	prog, dl := ParseSource(src, "legacy_table.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	out := viewNodesSource(prog.views[0].body, 1)
	if strings.Contains(out, "editable:") || strings.Contains(out, "fixed:") {
		t.Errorf("legacy table parameters were not removed:\n%s", out)
	}
	for _, want := range []string{`editMode: "editable"`, `cellText: "truncate"`} {
		if !strings.Contains(out, want) {
			t.Errorf("migrated table source missing %q:\n%s", want, out)
		}
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `data-sui-edit-mode="editable"`) || !strings.Contains(js, `data-sui-editable="1"`) {
		t.Error("migrated table markup is not editable")
	}
}

func TestLegacyChartParametersMigrate(t *testing.T) {
	src := `view main {
	chart "C" name: "c" data: "A,B\n1,2" chartType: "bar" interactive: true
}`
	prog, dl := ParseSource(src, "legacy_chart.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	out := viewNodesSource(prog.views[0].body, 1)
	if strings.Contains(out, "interactive:") {
		t.Errorf("legacy interactive was not removed:\n%s", out)
	}
	if !strings.Contains(out, `editMode: "editable"`) {
		t.Errorf("interactive chart did not migrate to editMode editable:\n%s", out)
	}
	js := GenerateJs(prog)
	if strings.Contains(js, `data-sui-chart-interactive=`) {
		t.Error("legacy interactive attribute is still emitted")
	}
	if !strings.Contains(js, `data-sui-chart-edit-mode="editable"`) {
		t.Error("migrated chart edit mode is missing")
	}
}

func TestLegacyStateChartMigratesToGraph(t *testing.T) {
	src := `view main {
	chart "Flow" name: "f" data: "\{\"states\":[]\}" chartType: "state" stateMode: "static"
}`
	prog, dl := ParseSource(src, "legacy_state.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	out := viewNodesSource(prog.views[0].body, 1)
	if !strings.HasPrefix(strings.TrimSpace(out), "graph ") {
		t.Errorf("state chart did not migrate to graph:\n%s", out)
	}
	if strings.Contains(out, "chartType:") || strings.Contains(out, "stateMode:") {
		t.Errorf("legacy state parameters were not removed:\n%s", out)
	}
	if !strings.Contains(out, `editMode: "none"`) {
		t.Errorf("static state chart did not migrate to editMode none:\n%s", out)
	}
}

func TestLegacyWorkbookSheetsMigrate(t *testing.T) {
	src := `view main {
	tabs "Book" variant: "sheets" {
		tab "Entry" {
			table name: "entry" columns: "a" {
			}
		}
		tab "Read" {
			table name: "read" columns: "a" selection: "row" {
			}
		}
	}
}`
	prog, dl := ParseSource(src, "legacy_workbook.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	out := viewNodesSource(prog.views[0].body, 1)
	if !strings.Contains(out, `name: "entry"`) || !strings.Contains(out, `editMode: "editable"`) {
		t.Errorf("data-entry workbook sheet did not migrate to editable:\n%s", out)
	}
	if !strings.Contains(out, `name: "read"`) || !strings.Contains(out, `editMode: "none"`) {
		t.Errorf("selection workbook sheet did not migrate to read-only:\n%s", out)
	}
}

func TestLegacyTreeIconsMigrate(t *testing.T) {
	src := `view main {
	tree "Files" name: "files" icon: "D" iconOpen: "O" {
		treeitem "src" icon: "F"
	}
}`
	prog, dl := ParseSource(src, "legacy_tree.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	out := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{`iconLeaf: "D"`, `iconParent: "D"`, `iconParentOpen: "O"`, `icon: "F"`} {
		if !strings.Contains(out, want) {
			t.Errorf("tree icon migration missing %q:\n%s", want, out)
		}
	}
	js := GenerateJs(prog)
	if strings.Contains(js, `data-sui-tree-icon="`) || strings.Contains(js, `data-sui-tree-icon-open="`) {
		t.Error("legacy tree icon attributes are still emitted")
	}
}

func TestEditModeRejectsInvalidValue(t *testing.T) {
	src := `view main {
	table name: "t" columns: "a" editMode: "persistant" {
	}
}`
	prog, dl := ParseSource(src, "invalid_mode.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatal("invalid editMode value was accepted")
	}
	if !strings.Contains(strings.ToLower(cd.String()), "editmode") {
		t.Errorf("diagnostic does not mention editMode: %s", cd.String())
	}
}
