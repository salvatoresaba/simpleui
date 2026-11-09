package main

import (
	"strings"
	"testing"
)

func TestEditModeDesignerThreeStateControls(t *testing.T) {
	for _, want := range []string{
		`id="cs-edit-mode"`,
		`<option value="none">Non-editable</option>`,
		`<option value="editable">Editable</option>`,
		`<option value="persistent">Editable persistent</option>`,
		`["non-editable", "none"]`,
		`["editable persistent", "persistent"]`,
		`n.chartEditMode = "none"`,
		`v.chartEditMode = document.getElementById("cs-edit-mode").value`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer missing 3-state edit mode support %q", want)
		}
	}
}

func TestEditModeRuntimeSupportsSessionAndPersistent(t *testing.T) {
	for _, want := range []string{
		`function chartEditMode(root)`,
		`function chartEditingAllowed(root)`,
		`function chartPersists(root)`,
		`window.__SUI_CHART_SESSION`,
		`function chartStoreGet(root, key)`,
		`function chartStoreSet(root, key, value)`,
		`sui.chart.data.`,
		`persistentBootBlocks`,
		`function persistTableState(name, st, cfg)`,
		`function loadPersistedTable(name)`,
		`saveTableBookDefs`,
	} {
		if !strings.Contains(suiTableRuntime+suiChartRuntime, want) {
			t.Errorf("runtime missing edit-mode support %q", want)
		}
	}
}

func TestEditModeSourceRoundTripAndMarkup(t *testing.T) {
	src := `view main {
	table name: "persisted" columns: "name,amount" columnLabels: "Name,Amount" editMode: "persistent" {
	}
	table name: "readonly" columns: "id" editMode: "none" {
	}
	chart "C" name: "c" data: "Month,Revenue\nJan,12\nFeb,18" chartType: "bar" editMode: "editable"
	graph "G" name: "g" editMode: "none"
}
`
	prog, dl := ParseSource(src, "edit_mode.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`name="persisted"`,
		`data-sui-edit-mode="persistent"`,
		`data-sui-edit-mode="none"`,
		`data-sui-chart-edit-mode="editable"`,
		`data-sui-chart-edit-mode="none"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated app missing %q", want)
		}
	}
	page := GenerateHtmlPage(prog)
	for _, want := range []string{
		`data-sui-edit-mode="persistent"`,
		`data-sui-chart-edit-mode="editable"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("static page missing %q", want)
		}
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"editMode":"persistent"`,
		`"editMode":"none"`,
		`"chartEditMode":"editable"`,
		`"chartEditMode":"none"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
}
