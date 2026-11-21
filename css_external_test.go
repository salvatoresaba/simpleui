package sui

import (
	"strings"
	"testing"
)

func TestCSSExternalReferenceValidation(t *testing.T) {
	valid := []string{
		".a { background: url(/assets/a.png); }",
		".a { background: url('./assets/a.png'); }",
		".a { background: url(\"../assets/a.png\"); }",
		".a { background: url(#gradient); }",
		".a { background: url(data:image/png;base64,AAAA); }",
		".a { background-image: image-set(\"data:image/png;base64,AAAA\" 1x); }",
		".a { content: \"url(https://example.invalid/x)\"; }",
		".a { content: \"@import\"; }",
		"/* @import url(https://example.invalid/x.css); */ .a { color: red; }",
	}
	for _, css := range valid {
		if msg := cssExternalReferenceError(css); msg != "" {
			t.Errorf("CSS should be allowed: %q, got %q", css, msg)
		}
	}
	invalid := map[string]string{
		".a { background: url(https://example.invalid/x.png); }":                    "external URL",
		".a { background: url('HTTP://example.invalid/x.png'); }":                   "external URL",
		".a { background: url(//example.invalid/x.png); }":                          "external URL",
		".a { background: url(file:///etc/passwd); }":                               "external URL",
		".a { background: url(h\\74tps://example.invalid/x); }":                     "external URL",
		"@import url(/local.css);":                                                  "@import",
		"@import \"https://example.invalid/x.css\";":                                "@import",
		"@\\69mport \"https://example.invalid/x.css\";":                             "@import",
		".a { background: u\\72l(https://example.invalid/x.png); }":                 "external URL",
		".a { background-image: image-set(\"https://example.invalid/x.png\" 1x); }": "external URL",
	}
	for css, want := range invalid {
		msg := cssExternalReferenceError(css)
		if msg == "" {
			if want != "" {
				t.Errorf("CSS should be rejected: %q", css)
			}
			continue
		}
		if want != "" && !strings.Contains(msg, want) {
			t.Errorf("CSS %q: got %q, want substring %q", css, msg, want)
		}
	}
}

func TestCSSExternalReferenceCheckIntegration(t *testing.T) {
	src := "css \".a \\{ background: url(https://example.invalid/x.png); \\}\"\nview main {\n\tp \"X\"\n}\n"
	_, dl := ParseSource(src, "css_external.sui")
	if !dl.HasErrors() {
		t.Fatal("external URL in css should be rejected")
	}
	if !strings.Contains(dl.String(), "external URL") {
		t.Fatalf("missing external URL diagnostic:\n%s", dl.String())
	}

	styleSrc := "view main {\n\tdiv \"X\" style: \"background: url(https://example.invalid/x.png)\"\n}\n"
	prog, pdl := ParseSource(styleSrc, "css_style.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %s", pdl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() || !strings.Contains(cd.String(), "external URL") {
		t.Fatalf("static style should reject external URL:\n%s", cd.String())
	}

	graph := map[string]any{
		"initial": "a",
		"states": []any{map[string]any{
			"id": "a", "css": "background-image: url(https://example.invalid/x.png);",
		}},
		"links": []any{},
	}
	graphSrc := "view main {\n\tcard motionGraph: \"" + suiGraphLiteral(graph) + "\" { p \"G\" }\n}\n"
	gprog, gdl := ParseSource(graphSrc, "css_graph.sui")
	if gdl.HasErrors() {
		t.Fatalf("parse graph: %s", gdl.String())
	}
	gcd := Check(gprog)
	if !gcd.HasErrors() || !strings.Contains(gcd.String(), "external URL") {
		t.Fatalf("motionGraph state CSS should reject external URL:\n%s", gcd.String())
	}
}
