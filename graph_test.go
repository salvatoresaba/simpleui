package main

import (
	"strings"
	"testing"
)

// graph_test.go covers the dedicated Graph control: it compiles through the
// shared chart runtime, always renders the block-diagram geometry, rejects the
// legacy chartType attribute and survives the designer round trip.

const graphTestSource = `state flow = "\{\"states\":[\{\"id\":\"a\",\"label\":\"Alpha\"\},\{\"id\":\"b\",\"label\":\"Beta\"\}],\"transitions\":[\{\"from\":\"a\",\"to\":\"b\",\"label\":\"go\"\}]\}"

view main {
	graph "Flow" name: "flow" data: flow stateGrid: true stateGridGap: 24 tableView: "both"
}
`

func TestGraphControlCompiles(t *testing.T) {
	prog, dl := ParseSource(graphTestSource, "graph.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-chart"`,
		`data-sui-chart="1"`,
		`data-sui-chart-type="state"`,
		`data-sui-chart-state-grid="1"`,
		`data-sui-chart-state-grid-gap="24"`,
		`data-sui-chart-table-view="both"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("graph JS missing %q", want)
		}
	}
	if strings.Contains(js, `data-sui-chart-type="bar"`) {
		t.Errorf("graph must not default to a bar chart")
	}
}

func TestGraphControlStaticHtml(t *testing.T) {
	prog, dl := ParseSource(graphTestSource, "graph.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	if !strings.Contains(html, `data-sui-chart-type="state"`) {
		t.Errorf("static graph HTML missing state geometry: %s", html)
	}
	if !strings.Contains(html, `aria-label="Flow"`) {
		t.Errorf("static graph HTML missing accessible label")
	}
}

func TestGraphControlValidation(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"chartType", `view main { graph "G" chartType: "bar" }`, "chartType is not allowed on <graph>"},
		{"children", `view main { graph "G" { p "nested" } }`, "graph cannot contain child elements"},
		{"format", `view main { graph "G" format: "yaml" }`, "format must be one of"},
	}
	for _, tc := range cases {
		prog, dl := ParseSource(tc.src, "graph.sui")
		if dl.HasErrors() {
			t.Fatalf("%s parse: %s", tc.name, dl.String())
		}
		cd := Check(prog)
		if !cd.HasErrors() || !strings.Contains(cd.String(), tc.want) {
			t.Errorf("%s: want %q, got %s", tc.name, tc.want, cd.String())
		}
	}
}

func TestGraphVisibilityAttributes(t *testing.T) {
	src := `view main {
	graph "G" settingsButton: false showZoom: false showAdd: false showHistory: false
}`
	prog, dl := ParseSource(src, "graph_visibility.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-chart-settings-button="0"`,
		`data-sui-chart-show-zoom="0"`,
		`data-sui-chart-show-add="0"`,
		`data-sui-chart-show-history="0"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("graph JS missing %q", want)
		}
	}
	if strings.Contains(js, `class="sui-chart-menu-btn"`) {
		t.Errorf("settingsButton: false must not render the gear")
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`data-sui-chart-settings-button="0"`,
		`data-sui-chart-show-zoom="0"`,
		`data-sui-chart-show-add="0"`,
		`data-sui-chart-show-history="0"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static graph HTML missing %q", want)
		}
	}
	if strings.Contains(html, `class="sui-chart-menu-btn"`) {
		t.Errorf("static graph must omit the gear")
	}
}

func TestGraphFullSizeAndSplitterMarkup(t *testing.T) {
	src := `view main {
	graph "G" width: "100%" height: "100%" tableView: "chart"
}`
	prog, dl := ParseSource(src, "graph_fullsize.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-chart-width="100%"`,
		`data-sui-chart-height="100%"`,
		`class="sui-chart-splitter"`,
		`class="sui-chart-table" role="region"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("graph JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`data-sui-chart-width="100%"`,
		`data-sui-chart-height="100%"`,
		`class="sui-chart-splitter"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static graph HTML missing %q", want)
		}
	}
}

func TestGraphTableHeightLayout(t *testing.T) {
	for _, want := range []string{
		".sui-chart-table-both > .sui-chart-table",
		"--sui-chart-table-height",
		"max-height: calc(100% - var(--sui-chart-head-height, 38px) - 80px)",
		".sui-chart-table-both .sui-chart-viewport { flex: 1 1 auto;",
		".sui-chart-state-table thead th { position: sticky; top: 0;",
		"> .sui-chart-table-tools",
		"--sui-chart-table-head-top",
		"> .sui-chart-gantt-group-toggles",
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("table layout CSS missing %q", want)
		}
	}
	for _, want := range []string{
		`"sui.chart.table.height."`,
		`dragStartHeight + (dragStartY - clientY)`,
		`if (tableView === "both") applyChartTableHeight(root);`,
		`height = Math.max(80, Math.min(chartTableHeightMax(root), height));`,
		`function syncChartTableSticky(root)`,
		`"--sui-chart-table-tools-top"`,
		`--sui-chart-table-head-top`,
		`syncChartTableSticky(root);`,
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("table height runtime missing %q", want)
		}
	}
}

func TestGraphRuntimeThemePresets(t *testing.T) {
	for _, want := range []string{
		"STATE_THEME_PRESETS", "stateThemePreset(\"midnight\"", "stateThemePreset(\"paper\"",
		"stateThemePreset(\"sand\"", "stateThemePreset(\"contrast\"", "stateThemePreset(\"plum\"",
		"applyStateThemeOverrides", "STATE_THEME_FIELDS", "Title size", "Stripe opacity",
		"stateTagFieldNumber", "Preset", "Style",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("chart runtime theme support missing %q", want)
		}
	}
	if count := strings.Count(suiChartRuntime, "stateThemePreset(\""); count < 14 {
		t.Errorf("expected at least 14 theme presets, got %d", count)
	}
}

func TestGraphStateTableDeleteIcon(t *testing.T) {
	for _, want := range []string{
		"sui-chart-state-delete sui-chart-table-row-delete",
		`["M6 6l12 12", "M18 6L6 18"]`,
		`delCell.appendChild(del)`,
		`th.style.width = "88px"`,
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("state table delete action missing %q", want)
		}
	}
	if !strings.Contains(suiCssBase, "width: 88px !important") {
		t.Error("state table actions column is too narrow for all icons")
	}
}

func TestGraphStateTableToolbar(t *testing.T) {
	for _, gone := range []string{
		"openStateTableMenu",
		"sui-chart-state-table-settings",
		"sui-chart-state-table-tools",
	} {
		if strings.Contains(suiChartRuntime, gone) {
			t.Errorf("state table toolbar still contains %q", gone)
		}
	}
	if got := strings.Count(suiChartRuntime, `"Undo (Ctrl+Z)"`); got != 1 {
		t.Errorf("expected one undo button (graph toolbar), found %d", got)
	}
	for _, want := range []string{
		"sui-chart-state-table-tablist",
		"tabBar.appendChild(addRow)",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("state table toolbar missing %q", want)
		}
	}
	for _, want := range []string{
		".sui-chart-state-table-tabs .sui-chart-state-add-row { margin-left: auto;",
		".sui-chart-state-table-tablist { display: flex;",
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("state table toolbar CSS missing %q", want)
		}
	}
}

func TestGraphDesignerModel(t *testing.T) {
	prog, dl := ParseSource(graphTestSource, "graph.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"kind":"graph"`,
		`"chartType":"state"`,
		`"chartData":"flow"`,
		`"chartDataExpr":true`,
		`"chartTableView":"both"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
}
