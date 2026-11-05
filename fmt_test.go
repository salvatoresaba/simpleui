package main

import "strings"
import "testing"

// fmt_test.go - the token-level formatter: style, stability and comment
// preservation.

func TestFmtBasic(t *testing.T) {
	ugly := "state counter = 0\nfn inc() -> null{\ncounter=counter+1\n}\nview main{\nh1 \"{counter}\"\n}\n# trailing note\n"
	out, dl := FormatSource(ugly, "x.sui")
	if dl != nil && dl.HasErrors() {
		t.Fatalf("fmt errors: %v", dl)
	}
	for _, want := range []string{
		"state counter = 0",
		"fn inc() -> null {",
		"counter = counter + 1",
		"view main {",
		"h1 \"{counter}\"",
		"# trailing note", // comments must survive
	} {
		if !strings.Contains(out, want) {
			t.Errorf("fmt output missing %q\n---\n%v", want, out)
		}
	}
	// Idempotent: formatting twice is stable.
	out2, _ := FormatSource(out, "x.sui")
	if out != out2 {
		t.Errorf("fmt not stable:\n---first---\n%v\n---second---\n%v", out, out2)
	}
}

func TestFmtOperators(t *testing.T) {
	src := "fn f(a: int,b:int)->int{\nreturn a+b*2\n}\n"
	out, dl := FormatSource(src, "x.sui")
	if dl != nil && dl.HasErrors() {
		t.Fatalf("fmt errors: %v", dl)
	}
	for _, want := range []string{
		"fn f(a: int, b: int) -> int {", // spaces normalized
		"a + b * 2",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("fmt output missing %q\n---\n%v", want, out)
		}
	}
}

func TestFmtInvalidSource(t *testing.T) {
	_, dl := FormatSource("fn f( {\n", "x.sui")
	if dl == nil || !dl.HasErrors() {
		t.Errorf("invalid source must fail to format")
	}
}

// TestFmtBlankLines guards the formatter against inserting blank lines that
// were not in the source, while still collapsing intentional blank runs to a
// single blank line and expanding one-line blocks.
func TestFmtBlankLines(t *testing.T) {
	src := "state a = 1\nstate b = 2\nfn f() -> null {\n  a = 1\n}\nview App {\n  p \"x\"\n  p \"y\"\n}\n"
	out, dl := FormatSource(src, "x.sui")
	if dl != nil && dl.HasErrors() {
		t.Fatalf("fmt errors: %v", dl)
	}
	if strings.Contains(out, "\n\n") {
		t.Errorf("fmt added blank lines to a compact source:\n%q", out)
	}
	want := "state a = 1\nstate b = 2\nfn f() -> null {\n  a = 1\n}\nview App {\n  p \"x\"\n  p \"y\"\n}\n"
	if out != want {
		t.Errorf("fmt compact mismatch:\n got %q\nwant %q", out, want)
	}

	// An intentional blank line survives as exactly one.
	withBlank := "state a = 1\n\nstate b = 2\n"
	out2, _ := FormatSource(withBlank, "x.sui")
	if out2 != withBlank {
		t.Errorf("intentional blank line not preserved:\n got %q\nwant %q", out2, withBlank)
	}

	// One-line blocks are expanded; "{}" stays inline; "else" keeps a space.
	inline := "fn e() -> null {}\nfn f() -> null { return 1 }\nview A {\n  if a > 0 { p \"y\" } else { p \"n\" }\n}\n"
	out3, dl3 := FormatSource(inline, "x.sui")
	if dl3 != nil && dl3.HasErrors() {
		t.Fatalf("fmt errors: %v", dl3)
	}
	for _, wantFrag := range []string{
		"fn e() -> null {}",
		"fn f() -> null {\n  return 1\n}",
		"} else {",
	} {
		if !strings.Contains(out3, wantFrag) {
			t.Errorf("fmt inline output missing %q\n%v", wantFrag, out3)
		}
	}
}
