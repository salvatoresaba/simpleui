package accordion_test

import . "example.com/simpleui"

import (
	"strings"
	"testing"
)

const accordionTestSource = `view main {
	accordion "FAQ" {
		accordionitem "Question one" {
			p "First answer"
		}
		accordionitem "Question two" {
			p "Second answer"
		}
	}
}
`

func TestAccordionRender(t *testing.T) {
	prog, dl := ParseSource(accordionTestSource, "accordion.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}

	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-accordion"`,
		`aria-label="FAQ"`,
		`data-sui-accordion-key="sui-accordion-`,
		`class="sui-accordion-item open"`,
		`class="sui-accordion-head"`,
		`class="sui-accordion-panel"`,
		`aria-expanded="true"`,
		`Question one`,
		`First answer`,
		`sui-accordion-head`,
		`applyStored`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}

	css := GenerateCss(prog)
	for _, want := range []string{".sui-accordion", ".sui-accordion-item", ".sui-accordion-head", ".sui-accordion-panel"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}

	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-accordion"`,
		`class="sui-accordion-item open"`,
		`class="sui-accordion-head"`,
		`aria-expanded="true"`,
		`class="sui-accordion-panel"`,
		`Question one`,
		`First answer`,
		`hidden`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
}

func TestAccordionDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(accordionTestSource, "accordion.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	modelText := StringifyJson(GuiModelJSON(prog))
	if !strings.Contains(modelText, `"kind":"accordion"`) || !strings.Contains(modelText, `"kind":"accordionitem"`) {
		t.Fatalf("designer model lost accordion: %s", modelText)
	}
	src := ViewNodesSource(prog.Views()[0].Body(), 1)
	for _, want := range []string{`accordion "FAQ"`, `accordionitem "Question one"`, `accordionitem "Question two"`} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{"newAccordionNode", "newAccordionItemNode", "insertAccordionItem", `"Add accordion item"`, `"Add item below"`} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer accordion editor missing %q", want)
		}
	}
}
