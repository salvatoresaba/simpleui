package tooltip_test

import . "github.com/salvatoresaba/simpleui"

import (
	"strings"
	"testing"
)

const tooltipPropertyTestSource = `view main {
	button "Save" tooltip: "Saves the current document"
	input placeholder: "Name" tooltip: "Your display name"
}
`

func TestTooltipPropertyRender(t *testing.T) {
	prog, dl := ParseSource(tooltipPropertyTestSource, "tooltip.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}

	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-tooltip="Saves the current document"`,
		`data-sui-tooltip="Your display name"`,
		`sui-tooltip-bubble`,
		`pointerover`,
		`focusin`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	if strings.Contains(js, `class="sui-tooltip `) || strings.Contains(js, `class="sui-tooltip"`) {
		t.Errorf("tooltip must be a property, not a wrapper control:\n%s", js)
	}

	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`data-sui-tooltip="Saves the current document"`,
		`data-sui-tooltip="Your display name"`,
		`Save`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
	if strings.Contains(html, ` tooltip="`) || strings.Contains(html, ` tooltip='`) {
		t.Errorf("tooltip source attribute leaked into HTML: %s", html)
	}

	css := GenerateCss(prog)
	for _, want := range []string{".sui-tooltip-bubble", "position: fixed", "--sui-tooltip-max-width"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}
}

func TestTooltipPropertyDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(tooltipPropertyTestSource, "tooltip.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	modelText := StringifyJson(GuiModelJSON(prog))
	if !strings.Contains(modelText, `"tooltip":"Saves the current document"`) {
		t.Fatalf("designer model lost tooltip property: %s", modelText)
	}
	src := ViewNodesSource(prog.Views()[0].Body(), 1)
	for _, want := range []string{`button "Save"`, `tooltip: "Saves the current document"`, `input`, `tooltip: "Your display name"`} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{
		"tooltipPropertyControls", "Tooltip settings", "openTooltipCfg",
		"applyTooltipCfg", "--sui-tooltip-place", "Global defaults",
	} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer tooltip support missing %q", want)
		}
	}
}

func TestTooltipPropertyTextIsEscaped(t *testing.T) {
	src := "view main {\n\tbutton \"Go\" tooltip: \"<b>not markup</b>\"\n}\n"
	prog, dl := ParseSource(src, "tooltip.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	if strings.Contains(html, "<b>not markup</b>") {
		t.Errorf("tooltip text must be escaped, got raw markup: %s", html)
	}
	if !strings.Contains(html, "&lt;b&gt;not markup&lt;/b&gt;") {
		t.Errorf("escaped tooltip text missing: %s", html)
	}
}
