package sui

import "strings"
import "testing"

// errors_test.go - tests for the diagnostic pipeline: phase categories, hint
// rendering, source snippets and the "did you mean" helper.

func TestDiagLexicalKind(t *testing.T) {
	_, dl := Lex("state a = 1 @ 2", "l.sui")
	if dl.Count() == 0 {
		t.Errorf("expected a lexical diagnostic")
		return
	}
	if dl.list[0].kind != "lexical" {
		t.Errorf("kind = %q, want lexical", dl.list[0].kind)
	}
}

func TestDiagSyntaxKind(t *testing.T) {
	_, dl := ParseSource("fn f( { }", "s.sui")
	if !dl.HasErrors() {
		t.Errorf("expected a syntax diagnostic")
		return
	}
	all := dl.String()
	if !strings.Contains(all, "syntax error") {
		t.Errorf("missing 'syntax error' label in:\n%v", all)
	}
	if !strings.Contains(all, "s.sui:") {
		t.Errorf("missing source location in:\n%v", all)
	}
	if !strings.Contains(all, "|") {
		t.Errorf("missing snippet in:\n%v", all)
	}
}

func TestDiagTypeKindAndHint(t *testing.T) {
	src := "state counter = 0\n" +
		"\n" +
		"fn bad() -> null {\n" +
		"\tcounter = \"test\"\n" +
		"}\n"
	prog, pdl := ParseSource(src, "t.sui")
	if pdl.HasErrors() {
		t.Fatalf("unexpected parse diagnostics: %v", pdl.String())
	}
	dl := Check(prog)
	if dl.Count() != 1 {
		t.Fatalf("want exactly 1 diagnostic, got %v:\n%v", dl.Count(), dl.String())
	}
	d := dl.list[0]
	if d.kind != "type" {
		t.Errorf("kind = %q, want type", d.kind)
	}
	if !strings.Contains(d.msg, "cannot assign") {
		t.Errorf("msg = %q", d.msg)
	}
	if !strings.Contains(d.hint, "declared here: t.sui:1") {
		t.Errorf("hint = %q, want 'declared here: t.sui:1'", d.hint)
	}
	all := dl.String()
	if !strings.Contains(all, "type error") {
		t.Errorf("missing 'type error' label:\n%v", all)
	}
	if !strings.Contains(all, "counter = \"test\"") {
		t.Errorf("missing source line in render:\n%v", all)
	}
	if !strings.Contains(all, "^") {
		t.Errorf("missing caret in render:\n%v", all)
	}
	if !strings.Contains(all, "1 error.") {
		t.Errorf("missing count line in render:\n%v", all)
	}
}

func TestDiagRecommendationHint(t *testing.T) {
	src := "state counter = 0\n" +
		"\n" +
		"fn bad() -> null {\n" +
		"\tcountr = 1\n" +
		"}\n"
	prog, pdl := ParseSource(src, "r.sui")
	if pdl.HasErrors() {
		t.Fatalf("unexpected parse diagnostics: %v", pdl.String())
	}
	dl := Check(prog)
	found := false
	for i := 0; i < len(dl.list); i++ {
		if dl.list[i].kind == "type" && strings.Contains(dl.list[i].hint, "counter") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing did-you-mean hint in:\n%v", dl.String())
	}
}

func TestDiagFallbackWithoutSource(t *testing.T) {
	var dl DiagList
	dl.Add("nowhere.sui", 3, 4, "boom")
	out := dl.String()
	if !strings.Contains(out, "nowhere.sui:3:4: boom") {
		t.Errorf("fallback line = %q", out)
	}
	if strings.Contains(out, "-->") {
		t.Errorf("should not render a snippet without registered source: %q", out)
	}
}

func TestNearestName(t *testing.T) {
	els := []string{"counter", "status", "refresh"}
	if nearestName("countr", els) != "counter" {
		t.Errorf("close typo not found")
	}
	if nearestName("countrrr", els) != "" {
		t.Errorf("distant word should not match")
	}
	if nearestName("counter", els) != "" {
		t.Errorf("exact name should be ignored")
	}
}

func TestDiagListKindDefault(t *testing.T) {
	var dl DiagList
	dl.Kind("io")
	dl.Add("x", 1, 2, "missing")
	if dl.list[0].kind != "io" {
		t.Errorf("default kind not stamped: %q", dl.list[0].kind)
	}
	dl.AddEx("y", 1, 2, "syntax", "boom", "")
	if dl.list[1].kind != "syntax" {
		t.Errorf("explicit kind overridden: %q", dl.list[1].kind)
	}
}
