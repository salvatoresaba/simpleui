package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSingleHTML(t *testing.T) {
	src := "state n = 1\nview main {\n\ttitle \"Single\"\n\th1 \"{n}\"\n}\n"
	prog, dl := ParseSource(src, "single.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatal(dl.String())
	}
	html := BuildSingleHTML(prog, false)
	for _, want := range []string{"<!DOCTYPE html>", "<title>Single</title>", "<style>", "<script>", "<div id=\"app\"></div>", "__SUI_views"} {
		if !strings.Contains(html, want) {
			t.Errorf("single HTML missing %q", want)
		}
	}
	for _, bad := range []string{`href="style.css"`, `src="app.js"`} {
		if strings.Contains(html, bad) {
			t.Errorf("single HTML still references external asset %q", bad)
		}
	}
}

func TestSingleHTMLInlineEndEscaping(t *testing.T) {
	index := `<!DOCTYPE html><html><head><link rel="stylesheet" href="style.css"></head><body><div id="app"></div><script src="app.js"></script></body></html>`
	css := `body { content: "</style>"; }`
	js := `var x = "</script>"; var y = "</SCRIPT>";`
	html := CombineSingleHTML(index, css, js, false)
	if strings.Contains(html, `"</style>"`) || strings.Contains(html, `"</script>"`) || strings.Contains(html, `"</SCRIPT>"`) {
		t.Errorf("inline data can terminate its HTML element:\n%s", html)
	}
	if !strings.Contains(html, `"<\/style>"`) || !strings.Contains(html, `"<\/script>"`) || !strings.Contains(html, `"<\/SCRIPT>"`) {
		t.Errorf("inline closing tags were not escaped:\n%s", html)
	}
}

func TestMinifyCSS(t *testing.T) {
	in := "a { color: red; /* comment */ }\n@font-face { src: url(\"data:x;base64,AAAA\"); }"
	out := minifyCSS(in)
	if strings.Contains(out, "comment") {
		t.Errorf("CSS comment survived minify: %q", out)
	}
	if strings.Contains(out, "\n") {
		t.Errorf("CSS minify kept newlines: %q", out)
	}
	if !strings.Contains(out, "data:x;base64,AAAA") {
		t.Errorf("CSS string was damaged: %q", out)
	}
	if !strings.Contains(out, "color: red") {
		t.Errorf("CSS functional whitespace was damaged: %q", out)
	}
}

func TestMinifyJS(t *testing.T) {
	in := "// full line comment\nvar s = \"// not a comment\";\nvar t = `line1\n// template line\n`;\nvar n = 1;\n"
	out := minifyJS(in)
	if strings.Contains(out, "full line comment") {
		t.Errorf("JS comment survived minify:\n%s", out)
	}
	if !strings.Contains(out, `"// not a comment"`) {
		t.Errorf("JS string was damaged:\n%s", out)
	}
	if !strings.Contains(out, "line1\n// template line") {
		t.Errorf("JS template content was damaged:\n%s", out)
	}
	if !strings.Contains(out, "var n = 1;") {
		t.Errorf("JS code was damaged:\n%s", out)
	}
}

func TestSingleOutputPath(t *testing.T) {
	cases := map[string]string{
		"":           filepath.Join("dist", "app.html"),
		"app.html":   "app.html",
		"APP.HTM":    "APP.HTM",
		"dist":       filepath.Join("dist", "app.html"),
		"dist/":      filepath.Join("dist", "app.html"),
		"custom.dir": filepath.Join("custom.dir", "app.html"),
	}
	for in, want := range cases {
		if got := SingleOutputPath(in); got != want {
			t.Errorf("SingleOutputPath(%q) = %q, want %q", in, got, want)
		}
	}
}
