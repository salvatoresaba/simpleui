package main

import (
	"os"
	"strings"
	"testing"
)

// generateJsFromSource parses, type-checks and generates an app from a SUI
// source file, failing the test on any diagnostic.
func generateJsFromSource(t *testing.T, path string) string {
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
	return GenerateJs(prog)
}

func generateJsFromInline(t *testing.T, src string) string {
	t.Helper()
	prog, pdl := ParseSource(src, "feature_inline.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse inline: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check inline: %v", dl.String())
	}
	return GenerateJs(prog)
}

// TestGeneratedJsOmitsUnusedRuntimes is the contract of the conditional
// runtime emission: a program that only renders text and buttons must not
// carry any component runtime.
func TestGeneratedJsOmitsUnusedRuntimes(t *testing.T) {
	js := generateJsFromInline(t, `
state counter = 0

fn increment() -> null {
	counter = counter + 1
}

view main {
	title "Counter"
	h1 "Count: {counter}"
	button "Add" onClick: increment
}
`)
	runtimes := map[string]string{
		"table":        suiTableRuntime,
		"dnd":          suiDndRuntime,
		"motion":       suiMotionRuntime,
		"motion graph": suiMotionGraphRuntime,
		"overlay":      suiOverlayRuntime,
		"menu":         suiMenuRuntime,
		"tabs":         suiTabsRuntime,
		"accordion":    suiAccordionRuntime,
		"toast":        suiToastRuntime,
		"tooltip":      suiTooltipRuntime,
		"combobox":     suiComboboxRuntime,
		"datepicker":   suiDatePickerRuntime,
		"tree":         suiTreeRuntime,
		"richtext":     suiRichTextRuntime,
		"pdf fonts":    suiPdfFontsJS,
		"chart":        suiChartRuntime,
		"modal":        suiModalRuntime,
		"popover":      suiPopoverRuntime,
		"drawer":       suiDrawerRuntime,
		"splitter":     suiSplitterRuntime,
		"stepper":      suiStepperRuntime,
		"carousel":     suiCarouselRuntime,
		"commandbar":   suiCommandbarRuntime,
	}
	for name, body := range runtimes {
		if strings.Contains(js, body) {
			t.Errorf("unused %s runtime embedded in a text-only app", name)
		}
	}
	for _, hook := range []string{"__SUI_TBL.afterRender();", "__SUI_DND.afterRender();"} {
		if strings.Contains(js, hook) {
			t.Errorf("unused after-render hook %q emitted", hook)
		}
	}
}

// TestGeneratedJsEmitsRuntimePerComponent checks that each component pulls in
// its own runtime and that unrelated, large runtimes stay out.
func TestGeneratedJsEmitsRuntimePerComponent(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		want   []string
		absent []string
	}{
		{"table", "examples/table_data.sui",
			[]string{suiTableRuntime, "__SUI_TBL.afterRender();"},
			[]string{suiChartRuntime, suiRichTextRuntime, suiTreeRuntime}},
		{"chart", "examples/chart/app.sui",
			[]string{suiChartRuntime},
			[]string{suiTableRuntime, suiRichTextRuntime, suiTreeRuntime}},
		{"tree", "examples/tree/app.sui",
			[]string{suiTreeRuntime},
			[]string{suiChartRuntime, suiRichTextRuntime}},
		{"richtext", "examples/richtext_editor/app.sui",
			[]string{suiRichTextRuntime, suiPdfFontsJS},
			[]string{suiChartRuntime, suiTreeRuntime}},
		{"dropzone", "examples/dropzone/app.sui",
			[]string{suiDndRuntime, "__SUI_DND.afterRender();"},
			[]string{suiChartRuntime, suiTreeRuntime, suiRichTextRuntime}},
		{"motion", "examples/motion_layout/app.sui",
			[]string{suiMotionRuntime}, nil},
		{"motion graph", "examples/motion_graph/app.sui",
			[]string{suiMotionGraphRuntime}, nil},
		{"menu and overlay", "examples/menu_bar.sui",
			[]string{suiMenuRuntime, suiOverlayRuntime},
			[]string{suiChartRuntime}},
		{"tabs", "examples/tabs.sui",
			[]string{suiTabsRuntime}, nil},
		{"accordion", "examples/accordion.sui",
			[]string{suiAccordionRuntime}, nil},
		{"toast", "examples/toast.sui",
			[]string{suiToastRuntime}, nil},
		{"tooltip and overlay", "examples/tooltip.sui",
			[]string{suiTooltipRuntime, suiOverlayRuntime}, nil},
		{"combobox and overlay", "examples/combobox.sui",
			[]string{suiComboboxRuntime, suiOverlayRuntime}, nil},
		{"datepicker", "examples/datepicker/app.sui",
			[]string{suiDatePickerRuntime}, nil},
		{"modal", "examples/modal.sui",
			[]string{suiModalRuntime}, nil},
		{"popover and overlay", "examples/animation_lab/app.sui",
			[]string{suiPopoverRuntime, suiOverlayRuntime}, nil},
		{"drawer", "examples/drawer/app.sui",
			[]string{suiDrawerRuntime}, nil},
		{"splitter", "examples/splitter/app.sui",
			[]string{suiSplitterRuntime}, nil},
		{"stepper", "examples/stepper/app.sui",
			[]string{suiStepperRuntime}, nil},
		{"carousel", "examples/carousel/app.sui",
			[]string{suiCarouselRuntime}, nil},
		{"commandbar and overlay", "examples/commandbar/app.sui",
			[]string{suiCommandbarRuntime, suiOverlayRuntime}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			js := generateJsFromSource(t, tc.path)
			for _, want := range tc.want {
				if !strings.Contains(js, want) {
					t.Fatalf("%s: generated app.js is missing its runtime", tc.name)
				}
			}
			for _, absent := range tc.absent {
				if strings.Contains(js, absent) {
					t.Fatalf("%s: generated app.js embeds an unrelated runtime", tc.name)
				}
			}
		})
	}
}

// TestFeatureScanFindsApiAndToggleUsage covers triggers that do not appear as
// tags: intrinsic module calls and animation on switch/checkbox/radio.
func TestFeatureScanFindsApiAndToggleUsage(t *testing.T) {
	api := generateJsFromInline(t, `
fn probe() -> null {
	label, err = chart.getLabel("sales", "a")
}

view main {
	p "noop"
}
`)
	if !strings.Contains(api, suiChartRuntime) {
		t.Error("chart.* calls must embed the chart runtime even without a chart tag")
	}
	if strings.Contains(api, suiTableRuntime) {
		t.Error("chart.* calls must not embed the table runtime")
	}

	toggle := generateJsFromInline(t, `
state on = false

view main {
	switch "Enabled" checked: on animation: "morph"
}
`)
	if !strings.Contains(toggle, suiMotionRuntime) {
		t.Error("switch animation must embed the motion runtime")
	}
	if strings.Contains(toggle, suiMotionGraphRuntime) {
		t.Error("switch animation must not embed the motion graph runtime")
	}

	drag := generateJsFromInline(t, `
state key = drag.source()

view main {
	p "{key}"
}
`)
	if !strings.Contains(drag, suiDndRuntime) {
		t.Error("drag.* calls must embed the drag-and-drop runtime")
	}

	toast := generateJsFromInline(t, `
fn ping() -> null {
	toast.show("saved")
}

view main {
	button "Ping" onClick: ping
}
`)
	if !strings.Contains(toast, suiToastRuntime) {
		t.Error("toast.show must embed the toast runtime")
	}
	if strings.Contains(toast, suiOverlayRuntime) {
		t.Error("toast.show must not embed the overlay runtime")
	}

	modal := generateJsFromInline(t, `
fn open() -> null {
	modal.open("dialog")
}

view main {
	button "Open" onClick: open
}
`)
	if !strings.Contains(modal, suiModalRuntime) {
		t.Error("modal.open must embed the modal runtime")
	}
}
