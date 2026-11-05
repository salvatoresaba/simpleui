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
