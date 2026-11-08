package main

import (
	"strings"
	"testing"
)

func TestTableDeclarativeFeatures(t *testing.T) {
	src := `state version = 0
fn changed() -> null {
	version = version + 1
}
fn tableApi() -> null {
	n, err = table.sort("users", "status", "asc")
	f, ferr = table.filterText("users", "active")
	c, cerr = table.lastChange("users")
}
view main {
	table name: "users" columns: "id,name,status" columnLabels: "ID,Name,Status" columnWidths: "50,120,90" columnAlign: "right,left,center" columnTypes: "number,text,select" columnFormats: "text,text,text" columnOptions: "status=active|paused" sortable: true filterable: true pageSize: 2 selection: "row" multiSelect: true onCellChange: changed onRowChange: changed onTableChange: changed onRowSelect: changed onSort: changed {
		thead {
			trow {
				tcell "ID" head: true
				tcell "Name" head: true
				tcell "Status" head: true
			}
		}
		tbody {
			trow {
				tcell ""
				tcell ""
				tcell ""
			}
		}
	}
}
`
	prog, dl := ParseSource(src, "table_features.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`name="users"`, `data-sui-columns="id,name,status"`, `data-sui-column-labels="ID,Name,Status"`,
		`data-sui-column-widths="50,120,90"`, `data-sui-column-align="right,left,center"`,
		`data-sui-column-types="number,text,select"`, `data-sui-column-options="status=active|paused"`,
		`data-sui-sortable="1"`, `data-sui-filterable="1"`, `data-sui-page-size="2"`,
		`data-sui-selection="row"`, `data-sui-multi="1"`,
		`data-sui-table-oncellchange="changed"`, `data-sui-table-onrowchange="changed"`,
		`data-sui-table-ontablechange="changed"`, `data-sui-table-onrowselect="changed"`,
		`data-sui-table-onsort="changed"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("table markup missing %q", want)
		}
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"columns":"id,name,status"`, `"columnLabels":"ID,Name,Status"`, `"sortable":true`,
		`"filterable":true`, `"pageSize":"2"`, `"selection":"row"`, `"multiSelect":true`,
		`"onCellChange":"changed"`, `"onTableChange":"changed"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("table model missing %q: %s", want, model)
		}
	}
	out := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{
		`columns: "id,name,status"`, `columnLabels: "ID,Name,Status"`, `sortable: true`,
		`filterable: true`, `pageSize: 2`, `selection: "row"`, `multiSelect: true`,
		`onCellChange: changed`, `onTableChange: changed`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table source missing %q:\n%s", want, out)
		}
	}
	for _, want := range []string{
		`function sort(`, `function filter(`, `function filterText(`, `function page(`,
		`function setPageSize(`, `function selectedRows(`, `function selectedCells(`,
		`function lastChange(`, `sui-table-filter`, `sui-table-pager`, `table_sort`,
	} {
		if want == "table_sort" {
			continue // jsKey mapping lives in jsgen, checked below
		}
		if !strings.Contains(suiTableRuntime, want) {
			t.Errorf("table runtime missing %q", want)
		}
	}
	if !strings.Contains(js, `__SUI_TBL.sort`) && !strings.Contains(js, `table_sort`) {
		t.Errorf("generated JS missing table.sort mapping")
	}
	for _, want := range []string{
		`"Columns"`, `"Data view"`, `"from header"`, `"Use first row"`,
		`sheet tabs`, `"Bottom (Excel-like)"`, `function tabsControls`,
		`onCellChange`, `table.lastChange`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer table/tabs support missing %q", want)
		}
	}
}

func TestTabsBottomSheetStyle(t *testing.T) {
	src := `view main {
	tabs "Workbook" position: "bottom" variant: "sheets" {
		tab "Sheet 1" {
			table name: "one" columns: "a,b" columnLabels: "A,B" {
				thead {
					trow {
						tcell "A" head: true
						tcell "B" head: true
					}
				}
				tbody {
					trow {
						tcell "1"
						tcell "2"
					}
				}
			}
		}
		tab "Sheet 2" {
			table name: "two" columns: "a,b" columnLabels: "A,B" {
				thead {
					trow {
						tcell "A" head: true
						tcell "B" head: true
					}
				}
				tbody {
					trow {
						tcell "3"
						tcell "4"
					}
				}
			}
		}
	}
}
`
	prog, dl := ParseSource(src, "tabs_bottom.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{`data-sui-tabs-position="bottom"`, `data-sui-tabs-variant="sheets"`, `name="one"`, `name="two"`} {
		if !strings.Contains(js, want) {
			t.Errorf("tabs/table markup missing %q", want)
		}
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{`"position":"bottom"`, `"variant":"sheets"`} {
		if !strings.Contains(model, want) {
			t.Errorf("tabs model missing %q", want)
		}
	}
	out := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{`position: "bottom"`, `variant: "sheets"`} {
		if !strings.Contains(out, want) {
			t.Errorf("tabs source missing %q", want)
		}
	}
	for _, want := range []string{`.sui-tabs[data-sui-tabs-position="bottom"]`, `.sui-tabs[data-sui-tabs-variant="sheets"]`} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("tabs CSS missing %q", want)
		}
	}
	if !strings.Contains(guiHTML, `function tabsControls`) || !strings.Contains(guiHTML, `"Bottom (Excel-like)"`) {
		t.Errorf("designer tabs settings missing")
	}
}
