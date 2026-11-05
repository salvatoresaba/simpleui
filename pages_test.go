package main

import (
	"strings"
	"testing"
)

const pagesTestSource = `state visits = 0
fn goSettings() -> null {
	visits = visits + 1
	navigate("settings")
}
fn goHome() -> null {
	navigate("main")
}
view main {
	title "Pages"
	h1 "Main"
	button "Settings" onClick: goSettings
}
view settings {
	title "Settings"
	h1 "Settings"
	button "Back" onClick: goHome
}
`

func TestPageNavigation(t *testing.T) {
	prog, dl := ParseSource(pagesTestSource, "pages.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`__sui_navigate(`,
		"__sui_navigate(`settings`)",
		"__sui_navigate(`main`)",
		`__sui_page`,
		`window.__SUI_PAGE`,
		`data-sui-pagebox`,
		`settings`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
}

func TestNavigateTypecheck(t *testing.T) {
	src := `view main {
	button "Go" onClick: go
}
fn go() -> null {
	navigate(1)
}
`
	prog, dl := ParseSource(src, "bad.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatal("navigate(1) must be rejected")
	}
	if !strings.Contains(cd.String(), "page name string") {
		t.Fatalf("missing navigate type error: %s", cd.String())
	}
}

func TestNavigateSinglePageIsNoop(t *testing.T) {
	src := `view main {
	button "Go" onClick: go
}
fn go() -> null {
	navigate("other")
}
`
	prog, dl := ParseSource(src, "single.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "__sui_navigate(`other`)") {
		t.Fatalf("navigate wrapper missing")
	}
	if strings.Contains(js, "function __sui_page") {
		t.Fatalf("single-page app must not define __sui_page")
	}
}
