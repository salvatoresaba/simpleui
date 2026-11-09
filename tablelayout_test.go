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
		`<tbody><tr><td class="sui-layout-cell">`,
		`</td><td class="sui-layout-cell">`,
		`</tr><tr>`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("table layout JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`<table class="sui-layout-table layout"`,
		`<td class="sui-layout-cell">`,
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
		columns string
		ok      bool
	}{
		{"1", true},
		{"6", true},
		{"0", false},
		{"7", false},
		{"two", false},
	} {
		src := `view main { tablelayout columns: ` + tc.columns + ` }`
		prog, pdl := ParseSource(src, "table_layout.sui")
		if pdl.HasErrors() {
			t.Fatalf("columns %s parse: %s", tc.columns, pdl.String())
		}
		dl := Check(prog)
		if gotOK := !dl.HasErrors(); gotOK != tc.ok {
			t.Errorf("columns %s: check ok=%v, want %v", tc.columns, gotOK, tc.ok)
		}
	}
}

func TestTableLayoutPaletteEntry(t *testing.T) {
	for _, want := range []string{
		`"tablelayout"`,
		`tablelayout: "Table layout"`,
		`function newTableLayoutNode`,
		`function layoutTableControls`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer table layout integration missing %q", want)
		}
	}
}
