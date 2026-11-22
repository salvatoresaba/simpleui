package dropdown_test

import . "github.com/salvatoresaba/simpleui"

import (
	"strings"
	"testing"
)

func TestDropdownRender(t *testing.T) {
	src := `fn edit() -> null {
}

view main {
	dropdown "Actions" position: "right" {
		dropdownitem "Edit" onClick: edit
		dropdownitem separator: true
		dropdownitem "Delete" disabled: true
	}
}`
	prog, dl := ParseSource(src, "dropdown.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-menu"`,
		`data-sui-dropdown-right`,
		`class="sui-menu-button"`,
		`class="sui-menu-popup"`,
		`class="sui-menuitem"`,
		`class="sui-dropdown-sep"`,
		`>Edit<`,
		` disabled`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("dropdown JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-menu`,
		`sui-menu-button`,
		`sui-dropdown-sep`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("dropdown static HTML missing %q", want)
		}
	}
}

func TestPopoverRender(t *testing.T) {
	src := `view main {
	popover "Info" position: "bottom" {
		p "Popover content"
		button "Action"
	}
}`
	prog, dl := ParseSource(src, "popover.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-popover`,
		`data-sui-popover-position="bottom"`,
		`class="sui-popover-trigger"`,
		`class="sui-popover-panel"`,
		`Popover content`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("popover JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-popover"`,
		`sui-popover-trigger`,
		`sui-popover-panel`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("popover static HTML missing %q", want)
		}
	}
}
