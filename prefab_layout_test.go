package main

import (
	"strings"
	"testing"
)

func TestOverlayPlacementAndAnimationAttrs(t *testing.T) {
	src := `state show = false

view main {
	dropdown "Azioni" position: "right" placement: "top" animation: "slide" duration: "220ms" easing: "ease-in-out" delay: "40ms" {
		dropdownitem "Uno"
	}
	popover "Info" position: "auto" animation: "scale" duration: "180ms" easing: "ease-out" {
		p "Contenuto"
	}
	combobox "Paese" options: "Italia,Francia" placement: "top" animation: "fade" duration: "200ms" {
	}
	drawer "Filtri" open: show duration: "350ms" easing: "ease-out" {
		p "Filtri"
	}
	carousel value: 0 duration: "500ms" easing: "ease-in-out" {
		slide { p "A" }
		slide { p "B" }
	}
	skeleton lines: 2 duration: "2s"
	commandbar "Cerca" open: show animation: "slide" duration: "200ms" {
		command "Nuovo"
	}
}`
	prog, dl := ParseSource(src, "overlay.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`data-sui-dropdown-placement="top"`,
		`data-sui-animation="slide"`,
		`data-sui-duration="220ms"`,
		`data-sui-easing="ease-in-out"`,
		`data-sui-delay="40ms"`,
		`data-sui-popover-position="auto"`,
		`data-sui-animation="scale"`,
		`data-sui-combobox-placement="top"`,
		`--sui-drawer-duration:350ms`,
		`--sui-drawer-easing:ease-out`,
		`--sui-carousel-duration:500ms`,
		`--sui-carousel-easing:ease-in-out`,
		`--sui-sk-duration:2s`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("overlay JS missing %q", want)
		}
	}
	for _, want := range []string{
		`data-sui-dropdown-placement="top"`,
		`data-sui-popover-position="auto"`,
		`data-sui-combobox-placement="top"`,
		`--sui-sk-duration:2s`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("overlay static HTML missing %q", want)
		}
	}
}

func TestTogglePresetAndMotionAttrs(t *testing.T) {
	src := `state on = true
state shape = "pill"
state theme = "accent"
state anim = "morph"

view main {
	switch "Modalità" shape: shape theme: theme checked: on animation: anim duration: "360ms" easing: "ease-in-out" stagger: "0.12" onChange: noop
	checkbox "Accetta" shape: "ios" theme: "success" checked: on animation: "morph" onChange: noop
}

fn noop() -> null {
}`
	prog, dl := ParseSource(src, "toggle.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`__sui_preset(shape, "sui-sw-shape-",`,
		`__sui_preset(theme, "sui-sw-theme-",`,
		`data-sui-motion`,
		`data-sui-motion-key="0"`,
		`data-sui-motion-effect="${__sui_esc(anim)}"`,
		`data-sui-motion-effect="morph"`,
		`data-sui-motion-vars="--sui-sw-w`,
		`data-sui-motion-stagger="0.12"`,
		`sui-cb-shape-ios`,
		`sui-cb-theme-success`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("toggle preset JS missing %q", want)
		}
	}
	for _, want := range []string{
		`sui-cb-shape-ios`,
		`sui-cb-theme-success`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("toggle preset static HTML missing %q", want)
		}
	}
}

func TestDrawerRender(t *testing.T) {
	src := `state show = false

fn openDrawer() -> null {
	show = true
}

view main {
	button "Open" onClick: openDrawer
	drawer "Filters" side: "left" width: "280px" open: show {
		p "Drawer body"
	}
}`
	prog, dl := ParseSource(src, "drawer.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-drawer`,
		`data-sui-drawer-open="${(show) ? "1" : "0"}"`,
		`data-sui-drawer-set="setShow"`,
		`class="sui-drawer-backdrop"`,
		`class="sui-drawer sui-drawer-left"`,
		`Drawer body`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("drawer JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`data-sui-drawer`,
		`sui-drawer-left`,
		`Drawer body`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("drawer static HTML missing %q", want)
		}
	}
}

func TestSplitterRender(t *testing.T) {
	src := `view main {
	splitter direction: "vertical" initial: "40%" min: "90px" {
		pane { p "Top" }
		pane { p "Bottom" }
	}
}`
	prog, dl := ParseSource(src, "splitter.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-splitter sui-splitter-vertical"`,
		`data-sui-splitter`,
		`data-sui-splitter-direction="vertical"`,
		`data-sui-splitter-min="90px"`,
		`class="sui-split-pane" style="flex-basis:40%"`,
		`class="sui-split-bar" role="separator" tabindex="0" aria-orientation="horizontal"`,
		`Top`,
		`Bottom`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("splitter JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`sui-splitter-vertical`,
		`sui-split-bar`,
		`flex-basis:40%`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("splitter static HTML missing %q", want)
		}
	}
}

func TestSplitterRequiresTwoPanes(t *testing.T) {
	src := `view main {
	splitter {
		pane { p "Only" }
	}
}`
	prog, dl := ParseSource(src, "splitter.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatal("expected splitter with one pane to fail")
	}
	if !strings.Contains(cd.String(), "exactly two pane") {
		t.Fatalf("unexpected diagnostics: %s", cd.String())
	}
}

func TestDropzoneRender(t *testing.T) {
	src := `state files = ""

view main {
	dropzone "CSV files" dropFiles: files accept: ".csv" multiple: true hint: "Drop CSV files" {
		p "Extra"
	}
}`
	prog, dl := ParseSource(src, "dropzone.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-dropzone"`,
		`data-sui-dropzone="1"`,
		`data-sui-dropfiles="setFiles"`,
		`data-sui-pickfiles="setFiles"`,
		`data-sui-pick-accept=".csv"`,
		`data-sui-pick-multiple="1"`,
		`Drop CSV files`,
		`sui-dropzone-status`,
		`Extra`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("dropzone JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`sui-dropzone`,
		`data-sui-dropfiles="setFiles"`,
		`Drop CSV files`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("dropzone static HTML missing %q", want)
		}
	}
}
