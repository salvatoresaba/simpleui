package main

import (
	"strings"
	"testing"
)

const toastTestSource = `view main {
	toast class: "demo-toasts" hAlign: "left" vAlign: "bottom" hOffset: "24px" vOffset: "12px" duration: 4000 max: 3 icons: false
	button "Show info" onClick: showInfo
	button "Show success" onClick: showSuccess
	button "Clear" onClick: clearToasts
}

fn showInfo() -> null {
	toast.show("This is an info toast", "info")
}

fn showSuccess() -> null {
	toast.show("Saved", "success", 2000)
}

fn clearToasts() -> null {
	toast.clear()
}
`

func TestToastRender(t *testing.T) {
	prog, dl := ParseSource(toastTestSource, "toast.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}

	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-toast-host sui-toast-h-left sui-toast-v-bottom demo-toasts"`,
		`data-sui-toast-host="1"`,
		`data-sui-toast-x="24px"`,
		`data-sui-toast-y="12px"`,
		`data-sui-toast-duration="4000"`,
		`data-sui-toast-max="3"`,
		`data-sui-toast-icons="false"`,
		`aria-live="polite"`,
		`__sui_toast_show(`,
		`__sui_toast_clear()`,
		`function show(`,
		`textContent`,
		`normalizeKind`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}

	css := GenerateCss(prog)
	for _, want := range []string{".sui-toast-host", ".sui-toast-host.sui-toast-h-left", ".sui-toast-host.sui-toast-v-bottom", "column-reverse", ".sui-toast", ".sui-toast-success", ".sui-toast-close"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}

	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-toast-host sui-toast-h-left sui-toast-v-bottom demo-toasts"`,
		`data-sui-toast-host="1"`,
		`data-sui-toast-x="24px"`,
		`data-sui-toast-y="12px"`,
		`data-sui-toast-duration="4000"`,
		`data-sui-toast-max="3"`,
		`data-sui-toast-icons="false"`,
		`aria-live="polite"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
}

func TestToastCssTargetsInternalParts(t *testing.T) {
	res := guiCssSetBatch(
		".demo-toasts .sui-toast { padding: 4px; }\n",
		"demo-toasts .sui-toast",
		`[{"prop":"width","value":"420px"},{"prop":"border-width","value":"3px"}]`,
	)
	if !asBool(get(res, "ok")) {
		t.Fatalf("setbatch not ok: %v", StringifyJson(res))
	}
	out := asStr(get(res, "css"))
	for _, want := range []string{".demo-toasts .sui-toast", "width: 420px;", "border-width: 3px;", "padding: 4px;"} {
		if !strings.Contains(out, want) {
			t.Errorf("part-scoped toast CSS missing %q:\n%s", want, out)
		}
	}
}

func TestToastDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(toastTestSource, "toast.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	modelText := StringifyJson(guiModelJson(prog))
	if !strings.Contains(modelText, `"kind":"toast"`) || !strings.Contains(modelText, `"toastIcons":false`) ||
		!strings.Contains(modelText, `"toastHAlign":"left"`) || !strings.Contains(modelText, `"toastVAlign":"bottom"`) ||
		!strings.Contains(modelText, `"toastHOffset":"24px"`) || !strings.Contains(modelText, `"toastVOffset":"12px"`) {
		t.Fatalf("designer model lost toast host options: %s", modelText)
	}
	src := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{
		`toast class: "demo-toasts"`, `hAlign: "left"`, `vAlign: "bottom"`,
		`hOffset: "24px"`, `vOffset: "12px"`, `duration: 4000`, `max: 3`, `icons: false`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{
		"Toast host", `"toast.show"`, `"toast.clear"`, "sui-toast-host",
		"toastControls", "previewToastSoon", "horizontal position", "vertical position",
		"horizontal offset", "vertical offset", "toastCardCss", "applyToastCardCss",
		"preview kind", "Show preview",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer toast support missing %q", want)
		}
	}
}
