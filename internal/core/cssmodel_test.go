package core

import "strings"
import "testing"

// cssmodel_test.go - tests for the structured CSS engine.

func TestParseAndGenerateRoundtrip(t *testing.T) {
	src := "/* comment */\n" +
		".card, .panel {\n  background: #111;\n  padding: 12px !important;\n}\n" +
		"@media (max-width: 600px) {\n  .card { margin: 0; }\n}\n"
	s := ParseCss(src)
	if len(s.diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", s.diags)
	}
	out := GenCss(s)
	for _, want := range []string{".card,\n.panel {", "background: #111;", "padding: 12px !important;", "@media (max-width: 600px) {", "margin: 0;"} {
		if !strings.Contains(out, want) {
			t.Errorf("generated css missing %q:\n%v", want, out)
		}
	}
	// reparsing the generated css must be stable.
	s2 := ParseCss(out)
	if len(s2.diags) != 0 {
		t.Fatalf("reparse diagnostics: %+v", s2.diags)
	}
	if GenCss(s2) != out {
		t.Errorf("round-trip not stable:\n%v\n---\n%v", out, GenCss(s2))
	}
}

func TestParseRecovery(t *testing.T) {
	src := ".good { color: red; }\n" +
		".broken { color red; }   /* missing colon */\n" +
		".also-good { display: flex; }\n"
	s := ParseCss(src)
	if len(s.diags) == 0 {
		t.Fatal("expected diagnostics for the broken rule")
	}
	out := GenCss(s)
	if !strings.Contains(out, ".also-good") {
		t.Errorf("parser should recover and keep .also-good:\n%v", out)
	}
}

func TestSpecificity(t *testing.T) {
	cases := []struct {
		sel     string
		a, b, c int
	}{
		{"div", 0, 0, 1},
		{".card", 0, 1, 0},
		{"#app", 1, 0, 0},
		{"#app .card", 1, 1, 0},
		{"a:hover", 0, 1, 1},
		{"ul li.active", 0, 1, 2},
		{"input[type=text]", 0, 1, 1},
		{"p::before", 0, 0, 2},
	}
	for _, c := range cases {
		got := Specificity(c.sel)
		if got.a != c.a || got.b != c.b || got.c != c.c {
			t.Errorf("Specificity(%q) = %v, want (%d,%d,%d)", c.sel, got.String(), c.a, c.b, c.c)
		}
	}
}

func TestAnalyzeUnusedAndOverride(t *testing.T) {
	src := ".used { color: red; }\n" +
		".unused { color: blue; }\n" +
		".used { color: green; }\n"
	s := ParseCss(src)
	used := map[string]bool{"used": true}
	probs := AnalyzeCss(s, used)
	var unused, override, duplicate bool
	for _, p := range probs {
		if p.kind == "unused" && strings.Contains(p.message, "unused") {
			unused = true
		}
		if p.kind == "override" && strings.Contains(p.message, "overridden") {
			override = true
		}
		if p.kind == "duplicate" {
			duplicate = true
		}
	}
	if !unused {
		t.Errorf("expected unused-class problem, got: %+v", probs)
	}
	if !override {
		t.Errorf("expected override problem, got: %+v", probs)
	}
	if !duplicate {
		t.Errorf("expected duplicate-selector problem, got: %+v", probs)
	}
}

func TestAnalyzeSkipsKeyframesAndLayeredRoot(t *testing.T) {
	src := ":root { --a: 1; }\n" +
		":root { --b: 2; }\n" +
		"@keyframes one { 0% { opacity: 0; } 100% { opacity: 1; } }\n" +
		"@keyframes two { 0% { opacity: 1; } 100% { opacity: 0; } }\n" +
		".sui-progress { color: red; }\n" +
		".active { color: green; }\n" +
		".base { color: red; }\n" +
		".base:hover { color: blue; }\n" +
		".sui-tab { color: red; }\n" +
		".sui-tab.active { color: blue; }\n" +
		"table.suitbl { color: red; }\n" +
		"table.suitbl thead th { color: blue; }\n" +
		"img, video { background: red; }\n" +
		"img, video { object-fit: cover; }\n" +
		".author-unused { color: blue; }\n"
	probs := AnalyzeCss(ParseCss(src), map[string]bool{})
	var duplicate, override, runtimeUnused, authorUnused bool
	for _, p := range probs {
		switch p.kind {
		case "duplicate":
			duplicate = true
		case "override":
			override = true
		case "unused":
			if strings.Contains(p.message, "sui-progress") || strings.Contains(p.message, ".active") {
				runtimeUnused = true
			}
			if strings.Contains(p.message, "author-unused") {
				authorUnused = true
			}
		}
	}
	if duplicate {
		t.Errorf("duplicate reported for layered :root, keyframes or additive selectors: %+v", probs)
	}
	if override {
		t.Errorf("state/pseudo override reported as actionable: %+v", probs)
	}
	if runtimeUnused {
		t.Errorf("runtime class reported as unused: %+v", probs)
	}
	if !authorUnused {
		t.Errorf("author unused class was not reported: %+v", probs)
	}
}

// TestAnalyzeOverrideOnlyForTheSameElement checks that an override is reported
// only when the two rules can really fight over the same element.
//
// Two mutually exclusive variants (".row.part-row" and ".row.file-row") and a
// descendant refinement (".monitor" and ".monitor h1") never apply to the same
// element: reporting them is noise. A more specific variant of the same
// element (".card" and ".card.primary") is a real cascade problem and must
// stay reported.
func TestAnalyzeOverrideOnlyForTheSameElement(t *testing.T) {
	src := ".row.part-row { display: grid; gap: 14px; }\n" +
		".row.file-row { display: grid; gap: 14px; }\n" +
		".monitor { margin: 0; }\n" +
		".monitor h1 { margin: 6px; }\n" +
		".card { color: red; }\n" +
		".card.primary { color: blue; }\n"
	used := map[string]bool{"row": true, "part-row": true, "file-row": true, "monitor": true, "card": true, "primary": true}
	probs := AnalyzeCss(ParseCss(src), used)
	var over []string
	for _, p := range probs {
		if p.kind == "override" {
			over = append(over, p.message)
		}
	}
	if len(over) != 1 {
		t.Fatalf("expected only the .card/.card.primary override, got %d: %v", len(over), over)
	}
	if !strings.Contains(over[0], ".card.primary") {
		t.Errorf("expected the .card.primary override, got: %s", over[0])
	}
}

// TestAnalyzeOverrideWithTheSameElement checks the case of two rules that
// declare the same property for the same element: the later one wins and the
// author should know it.
func TestAnalyzeOverrideWithTheSameElement(t *testing.T) {
	src := ".card { padding: 10px; }\n.card { padding: 20px; }\n"
	probs := AnalyzeCss(ParseCss(src), map[string]bool{"card": true})
	found := false
	for _, p := range probs {
		if p.kind == "override" && strings.Contains(p.message, "padding") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the padding override to be reported: %+v", probs)
	}
}

func TestExtractTokens(t *testing.T) {
	src := ":root {\n  --primary: #ff0000;\n  --gap: 12px;\n}\n.card { color: var(--primary); }\n"
	s := ParseCss(src)
	toks := ExtractTokens(s)
	if len(toks) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(toks))
	}
	if toks[0].name != "--gap" || toks[1].name != "--primary" {
		t.Errorf("tokens should be sorted: %+v", toks)
	}
}
