package main

import "strings"
import "testing"

// htmltags_test.go - the full HTML element palette: every tag must typecheck
// and produce the expected generated markup.

func TestHtmlTagsAllowedAndRendered(t *testing.T) {
	// Each entry: the SUI tag, and the HTML tag it must generate.
	cases := []struct{ tag, html string }{
		{"div", "<div"}, {"section", "<section"}, {"article", "<article"},
		{"main", "<main"}, {"header", "<header"}, {"footer", "<footer"},
		{"nav", "<nav"}, {"aside", "<aside"}, {"span", "<span"},
		{"h4", "<h4"}, {"h5", "<h5"}, {"h6", "<h6"},
		{"blockquote", "<blockquote"}, {"code", "<code"}, {"mark", "<mark"},
		{"strong", "<strong"}, {"em", "<em"}, {"small", "<small"},
		{"ul", "<ul"}, {"ol", "<ol"}, {"li", "<li"},
		{"form", "<form"}, {"label", "<label"}, {"textarea", "<textarea"},
		{"select", "<select"}, {"option", "<option"}, {"fieldset", "<fieldset"},
		{"legend", "<legend"},
		{"menubar", "<nav"}, {"menu", "<div"}, {"menuitem", "<div"},
		{"tabs", "<div"}, {"tab", "<div"}, {"modal", "<div"}, {"modalbutton", "<button"},
		{"tbody", "<tbody"}, {"tfoot", "<tfoot"},
		{"tablelayout", "<table"},
		{"video", "<video"}, {"audio", "<audio"}, {"iframe", "<iframe"},
		{"hr", "<hr"}, {"br", "<br"},
	}
	for _, c := range cases {
		src := "view main {\n\t" + c.tag + " \"x\"\n}\n"
		if c.tag == "iframe" {
			src = "view main {\n\tiframe src: \"https://example.com/\"\n}\n"
		}
		prog, dl := ParseSource(src, "t.sui")
		if dl.HasErrors() {
			t.Errorf("tag %q parse: %v", c.tag, dl.String())
			continue
		}
		cdl := Check(prog)
		if cdl.HasErrors() {
			t.Errorf("tag %q check rejected: %v", c.tag, cdl.String())
			continue
		}
		js := GenerateJs(prog)
		if !strings.Contains(js, c.html) {
			t.Errorf("tag %q should render %s, got:\n%v", c.tag, c.html, js)
		}
	}
}

func TestTextAreaValueBinding(t *testing.T) {
	src := "state msg: string = \"\"\nview main {\n\ttextarea value: msg\n}\n"
	prog, dl := ParseSource(src, "ta.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %v", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "oninput") || !strings.Contains(js, "setMsg") {
		t.Errorf("textarea value should bind oninput to a setter:\n%v", js)
	}
}

// GenerateIndexHtml renders the real index.html shown in the "HTML" tab.
func TestGenerateHtmlPage(t *testing.T) {
	src := "view main {\n\ttitle \"Demo\"\n\th1 \"Hello\"\n\trow {\n\t\tbutton \"Go\" onClick: go\n\t}\n}\nfn go() -> null {\n}\n"
	prog, dl := ParseSource(src, "h.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	html := GenerateIndexHtml(prog)
	for _, want := range []string{"<!DOCTYPE html>", "<title>Demo</title>", "<link rel=\"stylesheet\" href=\"style.css\">", "<div id=\"app\"></div>", "<script src=\"app.js\"></script>"} {
		if !strings.Contains(html, want) {
			t.Errorf("index.html missing %q:\n%v", want, html)
		}
	}
}
