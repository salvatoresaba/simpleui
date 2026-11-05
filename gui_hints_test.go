package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gui_hints_test.go covers the designer state-chart hint controls: the panel
// markup, the serializer entries and the security-relevant colour helpers.
func TestGuiSuiQuote(t *testing.T) {
	cases := []struct{ in, want string }{
		{`a"b`, `"a\"b"`},
		{`a{b}`, `"a\{b\}"`},
		{"a\nb", `"a\nb"`},
		{`a\b`, `"a\\b"`},
	}
	for _, c := range cases {
		if got := guiSuiQuote(c.in); got != c.want {
			t.Errorf("guiSuiQuote(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestGuiStateHintPanelMarkup(t *testing.T) {
	for _, want := range []string{
		`id="chintscfg"`, `id="chintscfgbox"`, `id="chints-enabled"`,
		`id="chints-bg"`, `id="chints-opacity"`, `id="chints-text"`,
		`id="chints-border"`, `id="chints-font"`, `id="chints-fontsize"`,
		`id="chints-borderwidth"`, `id="chints-padding"`, `id="chints-radius"`,
		`id="chints-preview"`, "HINTS settings", "openChartHintsCfg",
		"applyChartHintsCfg", "resetChartHintsCfg", "updateChartHintsPreview",
		`id="chints-groups"`, `class="chints-group"`, `id="chints-opacity-label"`, `class="chints-bg-field"`,
		"Hints and colors", "Size and spacing", "chartHintHex",
		`id="cs-state-edit"`, "Edit mode",
		`id="celemhints"`, `id="celemhints-rows"`, "openChartElementHints",
		"applyChartElementHints", "Element hints...", "chartElementRows",
		"chartElementCsvDraft", "chartElementXmlDraft", "chartElementDraft",
		`id="cextrattrs"`, "openChartExtraCfg", "applyChartExtraCfg", "Chart attributes...", "cextra-row",
		"newGraphNode", "isGraphKind", "chartTypeChangePlan", "chartTypeWarningText",
		"chartSampleDataset", "chartSampleFormatFor", `graph: "Graph"`, `CHART_GRAPH_SAMPLE_JSON`,
		"Block diagram (legacy chart)", "Replace chart data?",
		"openGraphRuntimeSettings",
		"function graphControls", "openGraphStyleEditor", "applyGraphStyleEditor", "closeGraphStyleEditor",
		`id="gstyle"`, "Graph style editor", "gs-head-bg", "gs-title-font", "gs-title-weight",
		"showGraphStyleTab", `data-gtab="title"`,
		"show table", "zoom buttons", "add element button", "undo/redo buttons", "edit data",
		"showZoom", "showAdd", "showHistory", "settingsButton",
		"gs-border-width", "gs-btn-shadow", "gs-table-headfg", "gs-scroll-thumb",
		"updateGraphStylePreview", "graphStyleInputValues",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer HTML missing %q", want)
		}
	}
	if strings.Contains(guiHTML, `["state", "State machine"]`) {
		t.Error("chart type menu still offers State machine")
	}
	if strings.Contains(guiHTML, `id="gs-series-1"`) || strings.Contains(guiHTML, "Component 1") {
		t.Error("component color controls should be removed from the graph style editor")
	}
	for _, want := range []string{
		`chartParamAttr(n, "stateHintBg"`,
		`chartParamAttr(n, "stateHintOpacity"`,
		`chartParamAttr(n, "stateHintTextColor"`,
		`chartParamAttr(n, "stateHintBorderColor"`,
		`chartParamAttr(n, "stateHintFont"`,
		`chartParamAttr(n, "stateHintFontSize"`,
		`chartParamAttr(n, "stateHintBorderWidth"`,
		`chartParamAttr(n, "stateHintPadding"`,
		`chartParamAttr(n, "stateHintRadius"`,
		`if (n.chartExtra && typeof n.chartExtra === "object")`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer serializer missing %q", want)
		}
	}
}

func TestGuiScriptParses(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed; skipping designer script syntax test")
	}
	start := strings.Index(guiHTML, "<script>")
	end := strings.LastIndex(guiHTML, "</script>")
	if start < 0 || end <= start {
		t.Fatal("designer script block not found")
	}
	script := guiHTML[start+len("<script>") : end]
	path := filepath.Join(t.TempDir(), "gui.js")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write gui script: %v", err)
	}
	if out, err := exec.Command(nodePath, "--check", path).CombinedOutput(); err != nil {
		t.Fatalf("designer script syntax: %v\n%s", err, out)
	}
}

func TestGuiElementHintRows(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed; skipping designer element hint test")
	}
	start := strings.Index(guiHTML, "function chartElementLineText(")
	end := strings.Index(guiHTML, "function openChartElementHints(")
	if start < 0 || end <= start {
		t.Fatal("chart element hint helpers not found")
	}
	script := guiHTML[start:end] + `
const rows = chartElementRows({
  states: [{ id: 'a', hint: 'A|B' }],
  transitions: [{ from: 'a', to: 'b', hint: 'X|Y' }]
});
const literal = chartElementLiteralText('"A\\nB"');
const plain = chartElementLiteralText('"plain"');
const expr = chartElementLiteralText('computed()');
console.log(JSON.stringify([rows.map(function(r) { return [r.kind, r.item.id || r.item.from, r.item.hint]; }), literal, plain, expr]));
`
	path := filepath.Join(t.TempDir(), "element_hints.js")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write element hint test: %v", err)
	}
	out, err := exec.Command(nodePath, path).CombinedOutput()
	if err != nil {
		t.Fatalf("run element hint test: %v\n%s", err, out)
	}
	got := strings.TrimSpace(string(out))
	want := `[[["state","a","A|B"],["transition","a","X|Y"]],"A\nB","plain",null]`
	if got != want {
		t.Fatalf("element hint rows = %s, want %s", got, want)
	}
}

func TestGuiHintHex(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed; skipping designer hint colour test")
	}
	start := strings.Index(guiHTML, "function chartHintHex(")
	end := strings.Index(guiHTML, "function chartHintsPreviewColor(")
	if start < 0 || end <= start {
		t.Fatal("chartHintHex helper not found")
	}
	script := guiHTML[start:end] + `
console.log(JSON.stringify([
  chartHintHex("#abc", "#111111"),
  chartHintHex("transparent", "#111111"),
  chartHintHex("#112233", "#111111")
]));
`
	path := filepath.Join(t.TempDir(), "hint_hex.js")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write hint hex test: %v", err)
	}
	out, err := exec.Command(nodePath, path).CombinedOutput()
	if err != nil {
		t.Fatalf("run hint hex test: %v\n%s", err, out)
	}
	got := strings.TrimSpace(string(out))
	want := `["#aabbcc","#111111","#112233"]`
	if got != want {
		t.Fatalf("chartHintHex = %s, want %s", got, want)
	}
}

func TestGuiHintColourHelpers(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed; skipping designer helper test")
	}
	start := strings.Index(guiHTML, "function chartHintsColor(")
	end := strings.Index(guiHTML, "function updateChartHintsPreview(")
	if start < 0 || end <= start {
		t.Fatal("chart hint colour helpers not found")
	}
	helpers := guiHTML[start:end]
	script := helpers + `
const out = [
  chartHintsPreviewColor('#ff0000', 0.5, '#111111'),
  chartHintsPreviewColor('url(https://example.invalid/x)', 1, '#111111'),
  chartHintsPreviewColor('transparent', 1, '#111111'),
  chartHintsColor('', 'mono'),
  chartHintsColor('Georgia, serif', 'mono')
];
console.log(JSON.stringify(out));
`
	path := filepath.Join(t.TempDir(), "hint_colors.js")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write helper test: %v", err)
	}
	out, err := exec.Command(nodePath, path).CombinedOutput()
	if err != nil {
		t.Fatalf("run helper test: %v\n%s", err, out)
	}
	got := strings.TrimSpace(string(out))
	want := `["rgba(255,0,0,0.5)","#111111","transparent","mono","Georgia, serif"]`
	if got != want {
		t.Fatalf("hint colour helpers = %s, want %s", got, want)
	}
}
