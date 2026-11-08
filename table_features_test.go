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
	table name: "users" columns: "id,name,status" columnLabels: "ID,Name,Status" columnWidths: "50,120,90" columnAlign: "right,left,center" columnTypes: "number,text,select" columnFormats: "text,text,text" columnOptions: "status=active|paused" sortable: true filterable: true pageSize: 2 selection: "row" multiSelect: true showAddTable: true structureEditable: true cellText: "wrap" onCellChange: changed onRowChange: changed onTableChange: changed onRowSelect: changed onSort: changed {
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
		`data-sui-show-add-table="1"`, `data-sui-structure-editable="1"`, `data-sui-cell-text="wrap"`,
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
		`"showAddTable":true`, `"structureEditable":true`, `"cellText":"wrap"`,
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
		`showAddTable: true`, `structureEditable: true`, `cellText: "wrap"`,
		`onCellChange: changed`, `onTableChange: changed`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table source missing %q:\n%s", want, out)
		}
	}
	css := GenerateCss(prog)
	if !strings.Contains(css, `table[data-sui-cell-text="wrap"]`) {
		t.Errorf("generated CSS missing wrapping cell text mode")
	}
	if !strings.Contains(css, `.sui-table-sort-rank`) || strings.Contains(css, `th[data-sort-rank]::before`) {
		t.Errorf("sort rank must live in the header toolbar, not in a th pseudo-element")
	}
	for _, want := range []string{
		`function sort(`, `function filter(`, `function filterText(`, `function page(`,
		`function setPageSize(`, `function selectedRows(`, `function selectedCells(`,
		`function lastChange(`, `sui-table-filter`, `sui-table-pager`, `sui-table-struct-menu`, `sui-table-add`, `sui-tab-add`, `table_sort`,
		`sui-table-settings`, `sui-table-cfg`, `sui-table-filterbtn`, `sui-table-sortbtn`, `sui-table-fpop`, `sortKeys`,
		`columnValuePairs`, `colValueFilters`, `Filter on column headers`, `(Select All)`, `Clear Filter`,
		`Sort levels`, `+ Add level`, `Clear filters`, `Clear sort`, `rowIds`, `headCells`, `sui-table-sort-rank`,
		`if (editCell ||`, `c !== editCell`, `commitCellEdit`, `editTimer`, `__suiInputGuard`,
		`sui-table-topbar`, `sui-table-title`, `data-sui-table-topbar-for`,
		`bar.nextElementSibling !== anchor`, `function ensureHbtns`, `headLabel`,
		`st.header[col] = raw`, `targetLabel`, `runtimeLabelsStr`,
		`if (cfg2.editable) return;`, `editTextOf`, `setCellEditText`, `syncHeaderLabelsAttr`,
		`contenteditable', 'false`, `data-sui-editable', '1'`, `data-sui-columns', headKeys.join(',')`,
		`normalizeWorkbookRuntime`, `data-sui-show-add-table`, `data-sui-selection`,
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
	if !strings.Contains(guiPreviewPicker, `sui-hl-table`) {
		t.Errorf("preview picker missing whole-table highlight")
	}
	if !strings.Contains(guiPreviewPicker, `sui-table-select`) || !strings.Contains(guiPreviewPicker, `gutterHit`) {
		t.Errorf("preview picker missing table range selection")
	}
	// Body rows must map to the designer model, whose row indices include the
	// header rows; otherwise a click selects the cell above the pointer.
	if !strings.Contains(guiPreviewPicker, `headCount + bodyRow`) {
		t.Errorf("preview picker must offset body rows by the header row count")
	}
	// Row/column selection strips and the select-all corner must sit outside
	// the table and be driven by geometry, not by the clicked element.
	if !strings.Contains(guiPreviewPicker, `sui-tzone`) || !strings.Contains(guiPreviewPicker, `TABLE_BORDER_ZONE`) {
		t.Errorf("preview picker must show the row/column gutter strip")
	}
	if !strings.Contains(guiPreviewPicker, `sui-tcorner`) || !strings.Contains(guiPreviewPicker, `TABLE_CORNER`) {
		t.Errorf("preview picker must show the select-all corner")
	}
	if !strings.Contains(guiPreviewPicker, `rowAtY`) || !strings.Contains(guiPreviewPicker, `colAtX`) {
		t.Errorf("preview picker must resolve gutter rows/columns geometrically")
	}
	if !strings.Contains(guiPreviewPicker, `focusEditableCell`) {
		t.Errorf("preview picker must focus editable cells for in-preview typing")
	}
	for _, want := range []string{
		`"Columns"`, `"Data view"`, `"from header"`, `"Use first row"`,
		`sheet tabs`, `"Bottom (Excel-like)"`, `function tabsControls`,
		`onCellChange`, `table.lastChange`, `id="tablesettings"`, `Table settings...`,
		`tableAddSheetAction`, `tableDeleteSheetAction`, `showAddTable`,
		`structureEditable`, `+ Table`, `sui-table-select`, `applyTableSelection`,
		`tableSelCells`, `tableSelRows`, `tableSelCols`, `resize table`,
		`uniqueTableName`, `tableNameTaken`,
		`#tablesettings-body .acc.open > .accbody`, `grid-template-columns:repeat(2,minmax(0,1fr))`,
		`width:720px`, `onlyTables`, `flex:0 1 auto`,
		`capturePreviewTabs`, `restorePreviewTabs`, `copyTableSheetLook`,
		`color-mix(in srgb, var(--t-zebra, transparent) 14%, transparent)`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer table/tabs support missing %q", want)
		}
	}
	// Direct preview selection replaced the inspector mode buttons.
	if strings.Contains(guiHTML, "Table selectors") {
		t.Errorf("inspector must not show the table selectors")
	}
}

func TestTableZebraTintCSS(t *testing.T) {
	const src = `css "table.suitbl tbody tr:nth-child(even) td \{ background: var(--t-zebra, transparent); \}"
view main {
	table name: "t" columns: "a,b" {
	}
}
`
	prog, dl := ParseSource(src, "zebra.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	css := GenerateCss(prog)
	if !strings.Contains(css, "color-mix(in srgb, var(--t-zebra, transparent) 14%, transparent)") {
		t.Errorf("generated CSS must tint legacy zebra colors:\n%s", css)
	}
	if !strings.Contains(css, "color-mix(in srgb, var(--t-hover, transparent) 14%, transparent)") {
		t.Errorf("generated CSS must tint legacy hover colors")
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

func TestDesignerDataTablePreservesAttributes(t *testing.T) {
	src := `view main {
	table name: "orders" columns: "name,qty" columnLabels: "Product,Qty" columnTypes: "text,number" columnFormats: "text,number" editable: true sortable: true filterable: true showAddTable: true {
	}
}
`
	prog, dl := ParseSource(src, "data_table.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"dataTable":true`, `"editable":true`, `"columns":"name,qty"`,
		`"columnLabels":"Product,Qty"`, `"columnTypes":"text,number"`,
		`"sortable":true`, `"filterable":true`, `"showAddTable":true`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer data-table model missing %q: %s", want, model)
		}
	}
	// The designer must serialize a columns-declared table without a synthetic
	// thead/tbody, and must not let the type-to-edit bridge reload the preview.
	for _, want := range []string{"if (n.dataTable)", "n.dataTable || n.columns", "dst.editable = true"} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer table handling missing %q", want)
		}
	}
}

func TestDesignerNestedTableIsNotDataTable(t *testing.T) {
	src := `view main {
	table name: "plain" {
		thead {
			trow {
				tcell "A" head: true
			}
		}
		tbody {
			trow {
				tcell "x"
			}
		}
	}
}
`
	prog, dl := ParseSource(src, "nested_table.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	model := StringifyJson(guiModelJson(prog))
	if !strings.Contains(model, `"dataTable":false`) {
		t.Errorf("nested table must not be marked as a data table: %s", model)
	}
}

func TestDesignerWorkbookAuthoring(t *testing.T) {
	for _, want := range []string{
		"workbookEntriesForTabs", "renderWorkbookTreeRow", "workbookChosenTablePath",
		"wrapTableInWorkbook", "normalizeWorkbooks", "syncWorkbookWrapperSize",
		"previewAddWorkbookSheet", "applyPreviewCellEdit", "sui-table-add",
		"sui-cell-edit", "__forcePreviewActive", "dst.editable = true",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer workbook handling missing %q", want)
		}
	}
	for _, want := range []string{"sui-table-add", "sui-cell-edit", "postCellEdit", ".sui-tab-add"} {
		if !strings.Contains(guiPreviewPicker, want) {
			t.Errorf("preview picker missing %q", want)
		}
	}
}
