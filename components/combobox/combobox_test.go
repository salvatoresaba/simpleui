package combobox_test

import . "example.com/simpleui"

import (
	"strings"
	"testing"
)

const comboboxTestSource = `state country = ""
fn countryChanged() -> null {
}
view main {
	combobox "Country" options: "Italy,France,Spain=es" value: country onChange: countryChanged placeholder: "Search..." empty: "Nothing found" maxOptionLength: 12 maxOptions: 5
}
`

func TestComboboxRender(t *testing.T) {
	prog, dl := ParseSource(comboboxTestSource, "combobox.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-combobox"`,
		`data-sui-combobox="1"`,
		`data-sui-options="Italy,France,Spain=es"`,
		`data-sui-placeholder="Search..."`,
		`data-sui-empty="Nothing found"`,
		`data-sui-max-option-length="12"`,
		`data-sui-max-options="5"`,
		`sui-combobox-error`,
		`data-sui-combobox-setter="setCountry"`,
		`data-sui-combobox-onchange="countryChanged"`,
		`sui-combobox-input`,
		`sui-combobox-popup`,
		`role="combobox"`,
		`role="listbox"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-combobox"`,
		`data-sui-options="Italy,France,Spain=es"`,
		`data-sui-empty="Nothing found"`,
		`data-sui-max-option-length="12"`,
		`data-sui-max-options="5"`,
		`sui-combobox-error`,
		`sui-combobox-label`,
		`role="combobox"`,
		`role="listbox"`,
		`Country`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
	css := GenerateCss(prog)
	for _, want := range []string{".sui-combobox", ".sui-combobox-input", ".sui-combobox-popup", ".sui-combobox-option"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}
}

const comboboxDynamicSource = `state country = ""
state countryOptions = ""
fn loadCountries() -> null {
	countryOptions = "Italy,France,Spain=es"
}
view main {
	combobox "Country" options: countryOptions value: country placeholder: "Search..." empty: "Nothing found" maxOptionLength: 12 maxOptions: 5
}
`

func TestComboboxDynamicOptions(t *testing.T) {
	prog, dl := ParseSource(comboboxDynamicSource, "combobox.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `data-sui-options="${__sui_esc(countryOptions)}"`) {
		t.Fatalf("dynamic options expression missing from generated JS:\n%s", js)
	}
	if !strings.Contains(js, `data-sui-combobox-setter="setCountry"`) {
		t.Fatalf("dynamic value setter missing from generated JS")
	}
	if !strings.Contains(js, `data-sui-max-option-length="12"`) {
		t.Fatalf("dynamic max option length missing from generated JS")
	}
	if !strings.Contains(js, `data-sui-max-options="5"`) {
		t.Fatalf("dynamic max options missing from generated JS")
	}
	model := StringifyJson(GuiModelJSON(prog))
	for _, want := range []string{`"kind":"combobox"`, `"options":"countryOptions"`, `"optionsExpr":true`} {
		if !strings.Contains(model, want) {
			t.Errorf("designer dynamic model missing %q: %s", want, model)
		}
	}
	src := ViewNodesSource(prog.Views()[0].Body(), 1)
	if !strings.Contains(src, `options: countryOptions`) {
		t.Errorf("designer dynamic source round-trip missing options expression:\n%s", src)
	}
	for _, want := range []string{"options state", "Options in ", "At runtime assign a new comma-separated string"} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer runtime options support missing %q", want)
		}
	}
}

func TestComboboxDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(comboboxTestSource, "combobox.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	model := StringifyJson(GuiModelJSON(prog))
	for _, want := range []string{`"kind":"combobox"`, `"options":"Italy,France,Spain=es"`, `"value":"country"`, `"empty":"Nothing found"`, `"maxOptionLength":"12"`, `"maxOptions":"5"`} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
	src := ViewNodesSource(prog.Views()[0].Body(), 1)
	for _, want := range []string{`combobox "Country"`, `options: "Italy,France,Spain=es"`, `value: country`, `onChange: countryChanged`, `empty: "Nothing found"`, `maxOptionLength: 12`, `maxOptions: 5`} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{"comboboxControls", "value state", "combobox disabled", "empty text", "max option length", "max options", "comboboxSplitRaw", "comboboxDisplayText"} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer combobox support missing %q", want)
		}
	}
}

func TestComboboxTextIsEscaped(t *testing.T) {
	src := "view main {\n\tcombobox \"<b>Pick</b>\" options: \"A,B\"\n}\n"
	prog, dl := ParseSource(src, "combobox.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	if strings.Contains(html, "<b>Pick</b>") {
		t.Errorf("combobox label must be escaped: %s", html)
	}
	if !strings.Contains(html, "&lt;b&gt;Pick&lt;/b&gt;") {
		t.Errorf("escaped combobox label missing: %s", html)
	}
}
