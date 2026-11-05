package main

import (
	"strings"
	"testing"
)

const requiredControlsSource = `state name = ""
state role = ""
state age = 0
state day = ""
state at = ""
view main {
	input "Name" value: name required: true
	textarea "Notes" required: true
	select "Role" value: role options: "Admin,User" required: true
	radio "Choice" name: "choice" value: "a" required: true
	number "Age" value: age required: true
	date "Day" value: day required: true
	time "At" value: at required: true
	combobox "Country" options: "Italy,France" required: true
	datepicker "When" mode: "date" required: true
}
`

func TestRequiredAttributeGeneration(t *testing.T) {
	prog, dl := ParseSource(requiredControlsSource, "required.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if strings.Count(js, " required") < 8 {
		t.Errorf("expected required attributes in generated JS, got:\n%s", js)
	}
	for _, want := range []string{
		`data-sui-required="1"`,
		`data-sui-combobox="1"`,
		`data-sui-datepicker="1"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	if strings.Count(html, " required") < 8 {
		t.Errorf("expected required attributes in static HTML, got:\n%s", html)
	}
	if !strings.Contains(html, `data-sui-required="1"`) {
		t.Errorf("combobox/datepicker required state missing in static HTML")
	}
	model := StringifyJson(guiModelJson(prog))
	if !strings.Contains(model, `"required":true`) {
		t.Errorf("designer model missing required:true")
	}
	for _, want := range []string{"requiredControls", "Native required validation"} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer required support missing %q", want)
		}
	}
}
