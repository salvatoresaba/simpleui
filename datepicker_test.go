package main

import (
	"strings"
	"testing"
)

const datepickerTestSource = `state when = ""
state at = ""
fn changed() -> null {
}
view main {
	datepicker "Start" mode: "datetime" value: when min: "2026-01-01T09:00" max: "2026-12-31T18:00" step: 60 onChange: changed required: true
	datepicker "Day" mode: "date" value: when min: "2026-01-01" variant: "calendar"
	datepicker "Alarm" mode: "time" value: at step: 60 disabled: true
}
`

func TestDatepickerRender(t *testing.T) {
	prog, dl := ParseSource(datepickerTestSource, "datepicker.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-datepicker"`,
		`data-sui-datepicker="1"`,
		`data-sui-mode="datetime"`,
		`type="datetime-local"`,
		`type="time"`,
		`min="2026-01-01T09:00"`,
		`max="2026-12-31T18:00"`,
		`step="60"`,
		`data-sui-datepicker-setter="setWhen"`,
		`data-sui-datepicker-onchange="changed"`,
		`sui-datepicker-input`,
		`sui-datepicker-error`,
		`variant="calendar"`,
		`input.readOnly = false`,
		`input.addEventListener('input'`,
		`input.addEventListener('paste'`,
		`data-sui-calendar-open`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-datepicker"`,
		`data-sui-datepicker="1"`,
		`type="date"`,
		`type="datetime-local"`,
		`data-sui-mode="datetime"`,
		`variant="calendar"`,
		`sui-datepicker-label`,
		`sui-datepicker-error`,
		`Start`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
	css := GenerateCss(prog)
	for _, want := range []string{
		".sui-datepicker",
		".sui-datepicker-input",
		".sui-datepicker-error",
		".sui-datepicker-calendar",
		".sui-datepicker-day",
		`.sui-datepicker[data-sui-calendar-open="1"] { z-index: 4100; }`,
		`cursor: text`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}
}

func TestDatepickerDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(datepickerTestSource, "datepicker.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"kind":"datepicker"`,
		`"mode":"datetime"`,
		`"min":"2026-01-01T09:00"`,
		`"max":"2026-12-31T18:00"`,
		`"step":"60"`,
		`"value":"when"`,
		`"variant":"calendar"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
	src := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{
		`datepicker "Start"`,
		`mode: "datetime"`,
		`min: "2026-01-01T09:00"`,
		`max: "2026-12-31T18:00"`,
		`step: 60`,
		`value: when`,
		`variant: "calendar"`,
		`onChange: changed`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{"datepickerControls", "Date/Time picker", "value state", "date/time picker", "required"} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer datepicker support missing %q", want)
		}
	}
}

func TestDatepickerRejectsInvalidMode(t *testing.T) {
	src := "view main {\n\tdatepicker \"When\" mode: \"month\"\n}\n"
	prog, dl := ParseSource(src, "datepicker.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatalf("expected invalid mode diagnostic")
	}
	if !strings.Contains(cd.String(), "mode must be one of") {
		t.Fatalf("unexpected diagnostic: %s", cd.String())
	}
}

func TestDatepickerRejectsInvalidVariant(t *testing.T) {
	src := "view main {\n\tdatepicker \"When\" variant: \"floating\"\n}\n"
	prog, dl := ParseSource(src, "datepicker.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatalf("expected invalid datepicker variant diagnostic")
	}
	if !strings.Contains(cd.String(), "variant must be one of: native, calendar") {
		t.Fatalf("unexpected diagnostic: %s", cd.String())
	}
}

func TestDatepickerLabelIsEscaped(t *testing.T) {
	src := "view main {\n\tdatepicker \"<b>When</b>\"\n}\n"
	prog, dl := ParseSource(src, "datepicker.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	if strings.Contains(html, "<b>When</b>") {
		t.Errorf("datepicker label must be escaped")
	}
	if !strings.Contains(html, "&lt;b&gt;When&lt;/b&gt;") {
		t.Errorf("escaped datepicker label missing")
	}
}
