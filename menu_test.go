package main

import (
	"strings"
	"testing"
)

const menuTestSource = `state status = ""
view main {
	menubar "Main menu" {
		menu "File" {
			menuitem "New" onClick: choose
			menuitem "Docs" link: "/docs"
			divider
		}
		menu "Help" {
			menuitem "About" onClick: choose
			menu "Advanced" {
				menuitem "Privacy" link: "/privacy"
			}
		}
	}
	p "{status}"
}
fn choose() -> null {
	status = "chosen"
}
`

func TestMenuBarRenders(t *testing.T) {
	prog, dl := ParseSource(menuTestSource, "menu.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}

	js := GenerateJs(prog)
	for _, want := range []string{
		`<nav class="sui-menubar"`,
		`role="menubar"`,
		`aria-label="Main menu"`,
		`<div class="sui-menu"`,
		`class="sui-menu-button"`,
		`class="sui-menu-popup"`,
		`class="sui-menuitem"`,
		`aria-haspopup="true"`,
		`onclick = choose`,
		`__sui_url`,
		`closeMenus`,
		`closeBranch`,
		`directItems`,
		`Advanced`,
		`Privacy`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}

	css := GenerateCss(prog)
	for _, want := range []string{".sui-menubar", ".sui-menu-button", ".sui-menu-popup", ".sui-menuitem", ".sui-menu-popup .sui-menu > .sui-menu-popup"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}

	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`<nav class="sui-menubar"`,
		`role="menubar"`,
		`aria-label="Main menu"`,
		`<div class="sui-menu"`,
		`class="sui-menu-button"`,
		`class="sui-menu-popup"`,
		`class="sui-menuitem"`,
		`href="/docs"`,
		`Advanced`,
		`Privacy`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
}

func TestMenuEditorContextMenuWired(t *testing.T) {
	for _, want := range []string{
		"row.oncontextmenu",
		"showNodeContextMenu",
		"insertMenuNode",
		`"Add menu"`,
		`"Add menu item"`,
		`"Add submenu"`,
		`"Add item below"`,
		`"Add submenu below"`,
		`"Delete"`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer menu editor missing %q", want)
		}
	}
}

func TestMenuBarDesignerModelRoundTrip(t *testing.T) {
	prog, dl := ParseSource(menuTestSource, "menu.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	model := guiModelJson(prog)
	modelText := StringifyJson(model)
	if !strings.Contains(modelText, `"kind":"menubar"`) || !strings.Contains(modelText, `"kind":"menuitem"`) {
		t.Fatalf("designer model lost menu nodes: %s", modelText)
	}
	src := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{`menubar "Main menu"`, `menu "File"`, `menuitem "New" onClick: choose`, `menuitem "Docs" link: "/docs"`, `menu "Advanced"`, `menuitem "Privacy" link: "/privacy"`} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source round-trip missing %q:\n%s", want, src)
		}
	}
}
