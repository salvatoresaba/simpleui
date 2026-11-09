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

func TestTableLayoutPaletteEntry(t *testing.T) {
	for _, want := range []string{
		`"tablelayout"`,
		`tablelayout: "Table layout"`,
		`function newTableLayoutNode`,
		`function layoutTableControls`,
		`function layoutCellControls`,
		`layoutCellParentPath`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer table layout integration missing %q", want)
		}
	}
	if !strings.Contains(guiPreviewPicker, `sui-layout-table > tbody > tr > td`) {
		t.Errorf("designer preview must outline table layout cells")
	}
	if !strings.Contains(guiHTML, `function layoutEmptyCellControls`) || !strings.Contains(guiHTML, `dropFromPaletteLayout`) {
		t.Errorf("designer must support empty-cell selection and direct drops")
	}
	if !strings.Contains(suiCssBase, "@media (max-width: 640px)") {
		t.Errorf("table layout responsive CSS missing")
	}
}
