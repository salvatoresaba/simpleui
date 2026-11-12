package main

import (
	"strings"
	"testing"
)

func TestTableLayoutRendersStructuralCells(t *testing.T) {
	src := `view main {
	tablelayout columns: 2 class: "layout" {
		card "A"
		card "B"
		card "C"
	}
}`
	prog, dl := ParseSource(src, "table_layout.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"sui-layout-table layout",
		`role="presentation"`,
		`<tbody><tr><td class="sui-layout-cell" data-sui-layout-row="0"`,
		`</td><td class="sui-layout-cell" data-sui-layout-row="0"`,
		`</tr><tr>`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("table layout JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`<table class="sui-layout-table layout"`,
		`<td class="sui-layout-cell" data-sui-layout-row="0"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("table layout static HTML missing %q", want)
		}
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{`"kind":"tablelayout"`, `"layoutColumns":"2"`} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
}

func TestTableLayoutSelectionSyncsElementTree(t *testing.T) {
	// Preview cell selection must highlight the layoutcell row in the element
	// tree, not the owning table, and the table outline must drop away.
	for _, want := range []string{
		"layoutSelectedTreePaths",
		"isCellPrimary",
		"if (selLayoutCells.length) ids = [];",
		"expandTreePath",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("layout selection tree sync missing %q", want)
		}
	}
}

func TestTableLayoutPreviewExcludesDataTableInteraction(t *testing.T) {
	// Layout cells are design-time containers. The data-table cell/row/column
	// selection must not run inside them: its mousedown handler suppresses the
	// following click, which made cards and texts inside layout cells
	// unselectable. The select-all corner must not be injected either.
	for _, want := range []string{
		`t.closest('table.sui-layout-table')`,
		`table.classList.contains('sui-layout-table')`,
		`table:not(.sui-layout-table)`,
	} {
		if !strings.Contains(guiPreviewPicker, want) {
			t.Errorf("layout table preview guard missing %q", want)
		}
	}
}

func TestTableLayoutColumnsValidation(t *testing.T) {
	for _, tc := range []struct {
		attr  string
		value string
		ok    bool
	}{
		{"columns", "1", true},
		{"columns", "12", true},
		{"columns", "0", false},
		{"columns", "13", false},
		{"rows", "1", true},
		{"rows", "50", true},
		{"rows", "0", false},
		{"rows", "51", false},
		{"columns", "two", false},
	} {
		src := `view main { tablelayout ` + tc.attr + `: ` + tc.value + ` }`
		prog, pdl := ParseSource(src, "table_layout.sui")
		if pdl.HasErrors() {
			t.Fatalf("%s %s parse: %s", tc.attr, tc.value, pdl.String())
		}
		dl := Check(prog)
		if gotOK := !dl.HasErrors(); gotOK != tc.ok {
			t.Errorf("%s %s: check ok=%v, want %v", tc.attr, tc.value, gotOK, tc.ok)
		}
	}
}

func TestTableLayoutGridSpansSizesAndBorders(t *testing.T) {
	src := `view main {
	tablelayout rows: 3 columns: 3 width: "80%" height: "300px" columnWidths: "120px, 30%, auto" rowHeights: "80px, 120px" gapX: "10px" gapY: "14px" borderStyle: "solid" borderWidth: "2px" borderColor: "#ff0000" style: "outline:1px solid red" responsive: true semantic: true label: "Main grid" {
		card "A" colspan: 2
		card "B"
		card "C" rowspan: 2 cellPadding: "8px" cellAlign: "center" cellValign: "middle" cellBackground: "#112233" cellBorderStyle: "dotted" cellBorderWidth: "3px" cellBorderColor: "#00ff00"
	}
}`
	prog, dl := ParseSource(src, "table_layout.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`colspan="2"`,
		`rowspan="2"`,
		`style="width:120px"`,
		`style="height:80px"`,
		`border:2px solid #ff0000;`,
		`border-spacing:10px 14px`, `outline:1px solid red`, `sui-layout-responsive`, `role="table"`, `aria-label="Main grid"`, `padding:8px`, `text-align:center`, `background:#112233`, `border:3px dotted #00ff00;`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("table layout JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`colspan="2"`,
		`rowspan="2"`,
		`style="width:120px"`,
		`style="height:80px"`,
		`border:2px solid #ff0000;`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("table layout static HTML missing %q", want)
		}
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"layoutRows":"3"`, `"layoutColumns":"3"`,
		`"layoutColumnWidths":"120px, 30%, auto"`,
		`"layoutRowHeights":"80px, 120px"`,
		`"layoutGapX":"10px"`, `"layoutGapY":"14px"`,
		`"layoutBorderStyle":"solid"`, `"layoutBorderWidth":"2px"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
}

func TestTableLayoutExplicitCells(t *testing.T) {
	src := `view main {
	tablelayout rows: 2 columns: 2 {
		layoutcell row: 0 col: 0 {
			card "A"
		}
		layoutcell row: 0 col: 1 rowspan: 2
		layoutcell row: 1 col: 1 {
			card "B"
		}
	}
}`
	prog, dl := ParseSource(src, "table_layout.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-layout-row="0" data-sui-layout-col="1" rowspan="2"`,
		`data-sui-layout-row="1" data-sui-layout-col="1"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("explicit table layout JS missing %q", want)
		}
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{`"kind":"layoutcell"`, `"layoutRow":"0"`, `"layoutCol":"1"`} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
}

func TestTableLayoutInspectorStructurePanel(t *testing.T) {
	for _, want := range []string{
		"Structure table settings",
		"layoutstruct-cols",
		"layoutstruct-rows",
		"function layoutSizeRow",
		"Reset to auto",
		"LAYOUT_CELL_PROP_ATTR",
		"function selectedLayoutCellNode",
		"function layoutApplyCounts",
		"function layoutSetCellSpans",
		"span would cut through another cell",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("table layout structure panel missing %q", want)
		}
	}
	// Width/height live in the generic Size section and the table border lives
	// in the generic Border section: the table layout panel must not duplicate
	// them.
	for _, unwanted := range []string{
		`textRow("width", "layoutWidth"`,
		`textRow("height", "layoutHeight"`,
		`textRow("border width", "layoutBorderWidth"`,
		`textRow("border color", "layoutBorderColor"`,
		`cellTextRow("padding"`,
		`cellSelectRow("align"`,
		`cellSelectRow("vertical"`,
		`cellTextRow("background"`,
		`cellSelectRow("border"`,
	} {
		if strings.Contains(guiHTML, unwanted) {
			t.Errorf("table layout panel duplicates %q", unwanted)
		}
	}
}

func TestTableLayoutCellClassRendering(t *testing.T) {
	src := `view main {
	tablelayout columns: 1 {
		layoutcell row: 0 col: 0 class: "pad" {
			card "A"
		}
	}
}`
	prog, dl := ParseSource(src, "layout_cell_class.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `class="sui-layout-cell pad"`) {
		t.Errorf("generated JS must carry the layout cell class: %s", js)
	}
	html := GenerateHtmlPage(prog)
	if !strings.Contains(html, `class="sui-layout-cell pad"`) {
		t.Errorf("generated HTML must carry the layout cell class: %s", html)
	}
}

func TestTableLayoutCellAlignment(t *testing.T) {
	// Alignment inside a layout cell must not turn the <td> into a flex
	// container: that collapses the cell and breaks the grid. It uses the
	// cell's vertical-align / text-align and auto margins on the child.
	for _, want := range []string{
		`parent.kind === "layoutcell"`,
		`"vertical-align"`,
		`"text-align"`,
		`"margin-left": h === "r" ? "auto"`,
		`sn.kind === "tablelayout" || sn.kind === "layoutcell"`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("layout cell alignment missing %q", want)
		}
	}
	if !strings.Contains(guiPreviewPicker, "sui-layout-responsive") {
		t.Errorf("designer preview must keep layout cells in grid mode")
	}
}

func TestDesignerCssWritesAreSerialized(t *testing.T) {
	// Rapid inspector edits (padding then margin, border style/width/color)
	// must not overwrite each other with stale CSS: the writer is queued.
	for _, want := range []string{"__cssWriteQueue", "Promise.resolve()"} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer CSS write queue missing %q", want)
		}
	}
}

func TestTableLayoutPaletteEntry(t *testing.T) {
	for _, want := range []string{
		`"tablelayout"`,
		`tablelayout: "Table layout"`,
		`function newTableLayoutNode`,
		`function layoutTableControls`,
		`function layoutCellControls`,
		`layoutCellParentPath`,
		`function materializeLayoutCells`,
		`function splitLayoutSelection`,
		`function layoutMergeSelection`,
		`function layoutSetCount`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer table layout integration missing %q", want)
		}
	}
	if !strings.Contains(guiPreviewPicker, `sui-layout-table > tbody > tr > td`) {
		t.Errorf("designer preview must outline table layout cells")
	}
	if !strings.Contains(guiPreviewPicker, `sui-layout-split-v`) || !strings.Contains(guiPreviewPicker, `sui-layout-resize`) {
		t.Errorf("designer preview must expose draggable split handles")
	}
	if !strings.Contains(guiPreviewPicker, `sui-layout-pick-cell`) || !strings.Contains(guiPreviewPicker, `sui-layout-pick-table`) || !strings.Contains(guiPreviewPicker, `sui-layout-split-on`) {
		t.Errorf("designer preview must expose cell/table pick triangles and split selection")
	}
	if !strings.Contains(guiHTML, `function openLayoutStructure`) || !strings.Contains(guiHTML, `dropFromPaletteLayout`) {
		t.Errorf("designer must support empty-cell selection and direct drops")
	}
	if !strings.Contains(suiCssBase, "@media (max-width: 640px)") {
		t.Errorf("table layout responsive CSS missing")
	}
}
