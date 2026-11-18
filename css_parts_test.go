package main

import (
	"os"
	"strings"
	"testing"
)

// cssMarkers maps an optional component to a selector that only its CSS slice
// contains. They are used to prove that unused component styles stay out of
// style.css and that the full base stylesheet still keeps every slice.
var cssMarkers = map[string]string{
	"chart":      ".sui-chart {",
	"tree":       ".sui-tree-row {",
	"table":      ".sui-table-settings",
	"combobox":   ".sui-combobox {",
	"datepicker": ".sui-datepicker {",
	"spinner":    ".sui-spinner {",
	"tabs":       ".sui-tabs {",
	"accordion":  ".sui-accordion {",
	"toast":      ".sui-toast-icon",
	"tooltip":    ".sui-tooltip-bubble",
	"modal":      ".sui-modal {",
	"popover":    ".sui-popover-panel",
	"drawer":     ".sui-drawer {",
	"splitter":   ".sui-splitter {",
	"stepper":    ".sui-stepper {",
	"carousel":   ".sui-carousel {",
	"commandbar": ".sui-commandbar {",
	"menu":       ".sui-menubar {",
	"richtext":   ".sui-rich-editor",
	"dropzone":   ".sui-dropzone {",
	"timeline":   ".sui-timeline {",
	"emptystate": ".sui-empty {",
	"skeleton":   ".sui-skeleton {",
	"anim":       "sui-anim-fade",
}

// TestBaseCssKeepsAllSlices guards the split: suiCssBase must still expose the
// complete stylesheet for callers and tests that inspect it.
func TestBaseCssKeepsAllSlices(t *testing.T) {
	for name, marker := range cssMarkers {
		if name == "richtext" {
			// The editor stylesheet lives in its own conditionally emitted const.
			if !strings.Contains(suiRichTextCss, marker) {
				t.Errorf("suiRichTextCss lost the editor styles (%q)", marker)
			}
			continue
		}
		if !strings.Contains(suiCssBase, marker) {
			t.Errorf("suiCssBase lost the %s styles (%q)", name, marker)
		}
	}
}

// TestGeneratedCssOmitsUnusedStyles is the CSS counterpart of the runtime
// test: a text-only app must not ship component styles.
func TestGeneratedCssOmitsUnusedStyles(t *testing.T) {
	css := GenerateCss(nil)
	for name, marker := range cssMarkers {
		if strings.Contains(css, marker) {
			t.Errorf("unused %s styles embedded in a text-only app", name)
		}
	}
	if !strings.Contains(css, ".sui-view {") {
		t.Error("core layout styles missing from a text-only app")
	}
	if !strings.Contains(css, ".sui-button {") {
		t.Error("core button styles missing from a text-only app")
	}
}

// TestGeneratedCssEmitsComponentStyles checks that every component pulls in
// its own styles while unrelated, large slices (above all the chart
// stylesheet) stay out.
func TestGeneratedCssEmitsComponentStyles(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		marker string
	}{
		{"table", "examples/table_data.sui", cssMarkers["table"]},
		{"chart", "examples/chart/app.sui", cssMarkers["chart"]},
		{"tree", "examples/tree/app.sui", cssMarkers["tree"]},
		{"richtext", "examples/richtext_editor/app.sui", cssMarkers["richtext"]},
		{"dropzone", "examples/dropzone/app.sui", cssMarkers["dropzone"]},
		{"menu", "examples/menu_bar.sui", cssMarkers["menu"]},
		{"tabs", "examples/tabs.sui", cssMarkers["tabs"]},
		{"accordion", "examples/accordion.sui", cssMarkers["accordion"]},
		{"toast", "examples/toast.sui", cssMarkers["toast"]},
		{"tooltip", "examples/tooltip.sui", cssMarkers["tooltip"]},
		{"combobox", "examples/combobox.sui", cssMarkers["combobox"]},
		{"datepicker", "examples/datepicker/app.sui", cssMarkers["datepicker"]},
		{"modal", "examples/modal.sui", cssMarkers["modal"]},
		{"popover", "examples/animation_lab/app.sui", cssMarkers["popover"]},
		{"drawer", "examples/drawer/app.sui", cssMarkers["drawer"]},
		{"splitter", "examples/splitter/app.sui", cssMarkers["splitter"]},
		{"stepper", "examples/stepper/app.sui", cssMarkers["stepper"]},
		{"carousel", "examples/carousel/app.sui", cssMarkers["carousel"]},
		{"commandbar", "examples/commandbar/app.sui", cssMarkers["commandbar"]},
		{"spinner", "examples/spinner_image/app.sui", cssMarkers["spinner"]},
		{"timeline", "examples/timeline/app.sui", cssMarkers["timeline"]},
		{"emptystate", "examples/emptystate/app.sui", cssMarkers["emptystate"]},
		{"skeleton", "examples/skeleton/app.sui", cssMarkers["skeleton"]},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			css := generateCssForTest(t, tc.path)
			if !strings.Contains(css, tc.marker) {
				t.Fatalf("%s: generated style.css is missing its styles (%q)", tc.name, tc.marker)
			}
			if tc.name != "chart" && strings.Contains(css, cssMarkers["chart"]) {
				t.Fatalf("%s: generated style.css embeds the chart stylesheet", tc.name)
			}
		})
	}
}

// generateCssForTest parses and checks a SUI source file, then returns its
// generated stylesheet.
func generateCssForTest(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	prog, pdl := ParseSource(string(raw), path)
	if pdl.HasErrors() {
		t.Fatalf("parse %s: %v", path, pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check %s: %v", path, dl.String())
	}
	return GenerateCss(prog)
}
