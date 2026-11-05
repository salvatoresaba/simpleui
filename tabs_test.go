package main

import (
	"strings"
	"testing"
)

const tabsTestSource = `state name = ""
view main {
	tabs "Settings" {
		tab "General" {
			p "General settings"
			input "Name" value: name
		}
		tab "Advanced" {
			p "Advanced settings"
		}
	}
}
`

func TestTabsRender(t *testing.T) {
	prog, dl := ParseSource(tabsTestSource, "tabs.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}

	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-tabs"`,
		`aria-label="Settings"`,
		`data-sui-tabs-key="sui-tabs-`,
		`class="sui-tablist"`,
		`role="tablist"`,
		`class="sui-tab active"`,
		`role="tab"`,
		`aria-selected="true"`,
		`class="sui-tabpanels"`,
		`class="sui-tabpanel active"`,
		`General settings`,
		`Advanced settings`,
		`MutationObserver`,
		`activate`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}

	css := GenerateCss(prog)
	for _, want := range []string{".sui-tabs", ".sui-tablist", ".sui-tab", ".sui-tabpanel"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}

	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-tabs"`,
		`class="sui-tablist"`,
		`class="sui-tab active"`,
		`class="sui-tabpanels"`,
		`class="sui-tabpanel active"`,
		`General settings`,
		`hidden`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
}

func TestTabsDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(tabsTestSource, "tabs.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	modelText := StringifyJson(guiModelJson(prog))
	if !strings.Contains(modelText, `"kind":"tabs"`) || !strings.Contains(modelText, `"kind":"tab"`) {
		t.Fatalf("designer model lost tabs: %s", modelText)
	}
	src := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{`tabs "Settings"`, `tab "General"`, `tab "Advanced"`} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{"newTabsNode", "newTabNode", "insertTabNode", `"Add tab"`, `"Add tab below"`} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer tabs editor missing %q", want)
		}
	}
}
