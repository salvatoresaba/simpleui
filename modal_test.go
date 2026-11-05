package main

import (
	"strings"
	"testing"
)

const modalTestSource = `state name = ""
fn openDialog() -> null {
	modal.open("edit-name")
}
fn closeDialog() -> null {
	modal.close("edit-name")
}
fn saveDialog() -> null {
	name = "saved"
}
view main {
	button "Edit" onClick: openDialog
	modal "Edit name" name: "edit-name" {
		p "Update the name."
		input "Name" value: name
		modalbutton "Cancel" onClick: closeDialog
		modalbutton "Save" variant: "primary" onClick: saveDialog
	}
}
`

func TestModalRender(t *testing.T) {
	prog, dl := ParseSource(modalTestSource, "modal.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}

	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-modal"`,
		`data-sui-modal="edit-name"`,
		`aria-hidden="true"`,
		`class="sui-modal-panel"`,
		`class="sui-modal-title"`,
		`class="sui-modal-body"`,
		`class="sui-modal-actions"`,
		`class="sui-modal-button primary"`,
		`data-sui-modal-dismiss="1"`,
		`onclick="saveDialog()"`,
		`__sui_modal_open(`,
		`__sui_modal_close(`,
		`MutationObserver`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}

	css := GenerateCss(prog)
	for _, want := range []string{".sui-modal", ".sui-modal-panel", ".sui-modal-button", "sui-modal-open"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}

	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-modal"`,
		`data-sui-modal="edit-name"`,
		`hidden`,
		`class="sui-modal-actions"`,
		`class="sui-modal-button primary"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
}

func TestModalDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(modalTestSource, "modal.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	modelText := StringifyJson(guiModelJson(prog))
	if !strings.Contains(modelText, `"kind":"modal"`) || !strings.Contains(modelText, `"kind":"modalbutton"`) {
		t.Fatalf("designer model lost modal nodes: %s", modelText)
	}
	src := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{`modal "Edit name"`, `name: "edit-name"`, `modalbutton "Cancel"`, `variant: "primary"`} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{"newModalNode", "newModalButtonNode", "insertModalButton", `"Add modal button"`, `"Add button below"`, "modalControls"} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer modal editor missing %q", want)
		}
	}
}

func TestModalBuiltinTypes(t *testing.T) {
	src := `fn go() -> null {
	modal.open(1)
}
view main {
	button "x" onClick: go
}
`
	prog, dl := ParseSource(src, "badmodal.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	cd := Check(prog)
	if !strings.Contains(cd.String(), "modal name string") {
		t.Fatalf("modal.open(1) was not rejected")
	}
}
