package main

import (
	"os"
	"strings"
	"testing"
)

// chart_api_test.go covers the core `chart` module used by process-monitor
// views. It is always available (no import) and updates rendered state-chart
// tags synchronously.
func TestChartApiTypechecksAndGenerates(t *testing.T) {
	src := `fn paint() -> null {
	ok, e1 = chart.setTag("monitor", "a", "Running", "#22c55e", "#ffffff")
	ok2, e2 = chart.update("monitor", "edge_1", "\{\"text\":\"go\",\"tagTextColor\":\"#ffffff\"\}")
	n, e3 = chart.updateMany("monitor", "\{\"a\":\{\"tag\":\"Idle\"\},\"b\":\{\"tagBoxColor\":\"#f59e0b\"\}\}")
	info, e4 = chart.get("monitor", "a")
	all, e5 = chart.elements("monitor")
	found, e6 = chart.find("monitor", "tag:Running")
	attr, e7 = chart.getAttr("monitor", "a", "tag")
	ok6, e8 = chart.setAttr("monitor", "a", "text", "Running")
	events, e12 = chart.pollEvents("monitor")
	peek, e13 = chart.peekEvents("monitor")
	one, e14 = chart.nextEvent("monitor")
	ok7, e15 = chart.clearEvents("monitor")
	count, e16 = chart.eventCount("monitor")
	lbl, e17 = chart.getLabel("monitor", "a")
	ok8, e18 = chart.setLabel("monitor", "a", "A")
	txt, e19 = chart.getText("monitor", "a")
	ok9, e20 = chart.setText("monitor", "a", "Running")
	hint, e21 = chart.getHint("monitor", "a")
	ok10, e22 = chart.setHint("monitor", "a", "Tip")
	kind, e23 = chart.getKind("monitor", "a")
	w, e24 = chart.getWidth("monitor", "a")
	ok11, e25 = chart.setWidth("monitor", "a", 120)
	h, e26 = chart.getHeight("monitor", "a")
	ok12, e27 = chart.setHeight("monitor", "a", 60)
	x, e28 = chart.getX("monitor", "a")
	ok13, e29 = chart.setX("monitor", "a", 10)
	y, e30 = chart.getY("monitor", "a")
	ok14, e31 = chart.setY("monitor", "a", 20)
	rot, e32 = chart.getRotation("monitor", "a")
	ok15, e33 = chart.setRotation("monitor", "a", 90)
	dash, e34 = chart.getDashStyle("monitor", "edge_1")
	ok16, e35 = chart.setDashStyle("monitor", "edge_1", "dashed")
	sm, e36 = chart.getStartMarker("monitor", "edge_1")
	ok17, e37 = chart.setStartMarker("monitor", "edge_1", "triangle")
	ok18, e38 = chart.setFrom("monitor", "edge_1", "b")
	ok19, e39 = chart.setTo("monitor", "edge_1", "a")
	wp, e40 = chart.getWaypoints("monitor", "edge_1")
	ok20, e41 = chart.setWaypoints("monitor", "edge_1", "[]")
	tk, e42 = chart.getTagKind("monitor", "a")
	ok21, e43 = chart.setTagKind("monitor", "a", "external")
	focused, e44 = chart.focusRow("monitor", "a")
}

view main {
	chart "Monitor" name: "monitor" data: "" chartType: "state"
}
`
	prog, pdl := ParseSource(src, "chart_api.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check (chart module must need no import and no async): %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"chart_settag:",
		"chart_update:",
		"chart_updatemany:",
		"chart_get:",
		"chart_elements:",
		"chart_find:",
		"chart_getattr:",
		"chart_setattr:",
		"chart_pollevents:",
		"chart_peekevents:",
		"chart_nextevent:",
		"chart_clearevents:",
		"chart_eventcount:",
		"chart_getlabel:",
		"chart_setlabel:",
		"chart_gettext:",
		"chart_settext:",
		"chart_gethint:",
		"chart_sethint:",
		"chart_getkind:",
		"chart_getwidth:",
		"chart_setwidth:",
		"chart_getheight:",
		"chart_setheight:",
		"chart_getx:",
		"chart_sety:",
		"chart_getrotation:",
		"chart_setrotation:",
		"chart_getdashstyle:",
		"chart_setdashstyle:",
		"chart_setfrom:",
		"chart_setto:",
		"chart_getwaypoints:",
		"chart_setwaypoints:",
		"chart_gettagkind:",
		"chart_settagkind:",
		"chart_focusrow:",
		"__SUI_CHART.focusRow(",
		"__SUI_CHART.setTag(",
		"__SUI_CHART.update(",
		"__SUI_CHART.updateMany(",
		"__SUI_CHART.get(",
		"__SUI_CHART.elements(",
		"__SUI_CHART.pollEvents(",
		"__SUI_CHART.peekEvents(",
		"__SUI_CHART.nextEvent(",
		"__SUI_CHART.clearEvents(",
		"__SUI_CHART.eventCount(",
		"__SUI_CHART.getAttr(name, id, \"label\")",
		"__SUI_CHART.setAttr(name, id, \"tag\", value)",
		"__SUI_CHART.getAttr(name, id, \"width\")",
		"__SUI_CHART.setAttr(name, id, \"dashStyle\", value)",
		"__SUI_CHART.setAttr(name, id, \"from\", value)",
		"__SUI_CHART.find(",
		"__SUI_CHART.getAttr(",
		"__SUI_CHART.setAttr(",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated app.js missing %q", want)
		}
	}
	if strings.Contains(js, "await __SUI_STD.chart_") {
		t.Errorf("chart module must be synchronous (no await)")
	}
}

func TestStateModeAndEventAttributes(t *testing.T) {
	src := `fn clicked(payload: string) -> null {
}

view main {
	chart "Monitor" name: "monitor" data: "" chartType: "state" stateMode: "static"
}
`
	prog, pdl := ParseSource(src, "state_mode.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-chart-edit-mode="none"`,
		`function stateChartPollEvents(name)`,
		`stateChartEventQueue(root)`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated app.js missing %q", want)
		}
	}
}

func TestStateModeDesignerRoundTrip(t *testing.T) {
	src := `fn clicked(payload: string) -> null {
}

view main {
	chart "Monitor" name: "monitor" data: "" chartType: "state" stateMode: "static"
}
`
	prog, pdl := ParseSource(src, "state_mode_designer.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"chartEditMode":"none"`,
		`"kind":"graph"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
}

func TestStateElementHandlersAndUniqueIds(t *testing.T) {
	src := `fn processClicked(payload: string) -> null {
}

view main {
	chart "Monitor" name: "monitor" data: "" chartType: "state"
}
`
	prog, pdl := ParseSource(src, "state_element_handlers.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"window.__SUI_USER_FUNCS",
		`window.__SUI_USER_FUNCS["processClicked"] = 1;`,
		"every state needs an explicit unique",
		"duplicate state id",
		"shape and link ids must be unique",
		"function stateChartFireEvent(root, kind, item, part, eventName, ev)",
		"stateChartEventQueue(root)",
		"function stateChartPollEvents(name)",
		"function stateChartNextEvent(name)",
		"function stateChartClearEvents(name)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated app.js missing %q", want)
		}
	}
}

func TestStateHintsDesignerAndHtml(t *testing.T) {
	src := `view main {
	chart "C" name: "c" data: "" chartType: "state" stateHints: false stateHintBg: "#112233" stateHintOpacity: 0.4 stateHintTextColor: "#aabbcc" stateHintBorderColor: "#00ff99" stateHintFont: "Georgia, serif" stateHintFontSize: 15 stateHintBorderWidth: 2 stateHintPadding: 9 stateHintRadius: 12 stateGridGap: 40 stateRouting: "orthogonal" stateTagFont: "Georgia, serif" stateGridColor: "#cbd5e1"
}
`
	prog, pdl := ParseSource(src, "state_hints.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-chart-state-hints="0"`,
		`data-sui-chart-state-hint-bg="#112233"`,
		`data-sui-chart-state-hint-opacity="0.4"`,
		`data-sui-chart-state-hint-text-color="#aabbcc"`,
		`data-sui-chart-state-hint-border-color="#00ff99"`,
		`data-sui-chart-state-hint-font="Georgia, serif"`,
		`data-sui-chart-state-hint-font-size="15"`,
		`data-sui-chart-state-hint-border-width="2"`,
		`data-sui-chart-state-hint-padding="9"`,
		`data-sui-chart-state-hint-radius="12"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated chart missing %s", want)
		}
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"chartStateHints":false`,
		`"chartStateHintBg":"#112233"`,
		`"chartStateHintOpacity":"0.4"`,
		`"chartStateHintFont":"Georgia, serif"`,
		`"chartStateHintRadius":"12"`,
		`"chartExtra":`,
		`"stateGridGap"`,
		`"stateRouting"`,
		`"stateTagFont"`,
		`"stateGridColor"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %s: %s", want, model)
		}
	}
}

func TestProcessMonitorExample(t *testing.T) {
	raw, err := os.ReadFile("examples/process_monitor/app.sui")
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	prog, pdl := ParseSource(string(raw), "examples/process_monitor/app.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse example: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check example: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"chart_updatemany:",
		"chart_update:",
		"chart_pollevents:",
		"__SUI_CHART.pollEvents(",
		"pumpChartEvents",
		"processClicked",
		"window.__SUI_USER_FUNCS[\"processClicked\"] = 1;",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("process monitor app.js missing %q", want)
		}
	}
}

func TestChartApiWrongArity(t *testing.T) {
	src := `fn bad() -> null {
	ok, e = chart.setTag("monitor", "a", "Running")
}

view main {
	p "x"
}
`
	prog, pdl := ParseSource(src, "bad_chart_api.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	dl := Check(prog)
	if !dl.HasErrors() {
		t.Fatalf("expected an arity error for chart.setTag with three arguments")
	}
	if !strings.Contains(dl.String(), "expects 5 arguments") {
		t.Errorf("unexpected diagnostics: %v", dl.String())
	}
}
