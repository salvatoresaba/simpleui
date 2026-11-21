package sui

import "testing"
import "strings"

// lexer_test.go — tests for Lex (Agent A, front-end).

func kindsOf(toks []Token) []string {
	var ks []string
	for i := 0; i < len(toks); i++ {
		ks = append(ks, toks[i].kind)
	}
	return ks
}

func TestLexBasic(t *testing.T) {
	src := "# a comment\nstate counter: int = 42\nfn greet(name: string) -> string {\n\treturn \"hi {name}\"\n}\n"
	toks, dl := Lex(src, "t.sui")
	got := strings.Join(kindsOf(toks), " ")
	want := "newline state ident : ident = int newline fn ident ( ident : ident ) -> ident { newline return str newline } newline eof"
	if got != want {
		t.Errorf("basic lex kinds:\n got: %q\nwant: %q", got, want)
	}
	if dl.Count() != 0 {
		t.Errorf("basic lex produced %v diagnostics, want 0", dl.Count())
	}
}

func TestLexKeywords(t *testing.T) {
	src := "state fn async if else loop in return break continue let import from view title true false null and or not\n"
	toks, dl := Lex(src, "t.sui")
	got := strings.Join(kindsOf(toks), " ")
	want := "state fn async if else loop in return break continue let import from view title true false null and or not newline eof"
	if got != want {
		t.Errorf("keyword lex kinds:\n got: %q\nwant: %q", got, want)
	}
	if dl.Count() != 0 {
		t.Errorf("keywords produced %v diagnostics, want 0", dl.Count())
	}
}

func TestLexNumbers(t *testing.T) {
	src := "3 \n 1.5 \n 1. \n 0..10 \n 2.5.5"
	toks, dl := Lex(src, "t.sui")
	got := strings.Join(kindsOf(toks), " ")
	want := "int newline float newline err newline int .. int newline float . int eof"
	if got != want {
		t.Errorf("number lex kinds:\n got: %q\nwant: %q", got, want)
	}
	if dl.Count() != 1 {
		t.Errorf("number lex produced %v diagnostics, want 1 (the '1.' literal)", dl.Count())
	}
}

func TestLexStrings(t *testing.T) {
	src := "\"a\\\"b\\\\c\\n\\t\"\n\"unterminated\n"
	toks, dl := Lex(src, "t.sui")
	if len(toks) != 4 {
		t.Errorf("string lex: got %v tokens, want 4", len(toks))
		return
	}
	if toks[0].kind != "str" {
		t.Errorf("string lex: first token kind %v, want str", toks[0].kind)
	}
	if toks[0].text != "a\\\"b\\\\c\\n\\t" {
		t.Errorf("string lex: raw inner text %q, want raw escapes kept", toks[0].text)
	}
	if toks[1].kind != "newline" {
		t.Errorf("string lex: second token kind %v, want newline", toks[1].kind)
	}
	if toks[2].kind != "err" {
		t.Errorf("string lex: unterminated string token kind %v, want err", toks[2].kind)
	}
	if toks[3].kind != "eof" {
		t.Errorf("string lex: last token %v, want eof", toks[3].kind)
	}
	if dl.Count() != 1 {
		t.Errorf("string lex produced %v diagnostics, want 1 (unterminated)", dl.Count())
	}
}

func TestLexOps(t *testing.T) {
	src := "a -> b .. c == d != e <= f >= g < h > i + j - k * l / m % n . o , p : q\n"
	toks, dl := Lex(src, "t.sui")
	got := strings.Join(kindsOf(toks), " ")
	want := "ident -> ident .. ident == ident != ident <= ident >= ident < ident > ident + ident - ident * ident / ident % ident . ident , ident : ident newline eof"
	if got != want {
		t.Errorf("operator lex kinds:\n got: %q\nwant: %q", got, want)
	}
	if dl.Count() != 0 {
		t.Errorf("operator lex produced %v diagnostics, want 0", dl.Count())
	}
}

func TestLexIllegalChar(t *testing.T) {
	src := "a $ b\n"
	toks, dl := Lex(src, "t.sui")
	got := strings.Join(kindsOf(toks), " ")
	want := "ident err ident newline eof"
	if got != want {
		t.Errorf("illegal char lex kinds:\n got: %q\nwant: %q", got, want)
	}
	if dl.Count() != 1 {
		t.Errorf("illegal char produced %v diagnostics, want 1", dl.Count())
	}
	if toks[1].text != "unexpected character" {
		t.Errorf("err token text %q, want the diagnostic message", toks[1].text)
	}
}

func TestLexPositions(t *testing.T) {
	src := "a\n  b"
	toks, _ := Lex(src, "t.sui")
	if toks[0].line != 1 || toks[0].col != 1 || toks[0].offset != 0 {
		t.Errorf("token a: got %v:%v@%v, want 1:1@0", toks[0].line, toks[0].col, toks[0].offset)
	}
	if toks[2].line != 2 || toks[2].col != 3 {
		t.Errorf("token b: got %v:%v, want 2:3", toks[2].line, toks[2].col)
	}
}
