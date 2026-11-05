package main

import "strings"
import "testing"

// mutate_test.go - data mutation (v0.5): struct field assignment, array/map
// element assignment (both reactive through the state setter) and the
// postfix ++ / -- sugar.

func TestMutateStructFieldReactive(t *testing.T) {
	src := "type User struct { name: string, age: int }\n" +
		"\n" +
		"state me: User = User{ name: \"Ann\", age: 20 }\n" +
		"\n" +
		"fn birthday() -> null {\n" +
		"\tme.age += 1\n" +
		"}\n" +
		"\n" +
		"fn rename() -> null {\n" +
		"\tme.name = \"Bee\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("struct mutate program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "m.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"me.age = (me.age + 1);",
		"me.name = `Bee`;",
		"setMe(me);",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("struct mutate jsgen missing %q\n----\n%v", want, js)
		}
	}
	if strings.Count(js, "setMe(me);") != 2 {
		t.Errorf("expected both field mutations to re-sync the state (2 x setMe)")
	}
}

func TestMutateIndexAssign(t *testing.T) {
	src := "state notes: array[string] = [\"a\", \"b\"]\n" +
		"state scores: map[string]int = { \"x\": 1 }\n" +
		"\n" +
		"fn edit() -> null {\n" +
		"\tnotes[0] = \"z\"\n" +
		"\tscores[\"x\"] += 5\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("index mutate program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "i.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"notes[0] = `z`;",
		"setNotes(notes);",
		"scores[`x`] = (scores[`x`] + 5);",
		"setScores(scores);",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("index mutate jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestMutatePostfixSugar(t *testing.T) {
	src := "state counter = 0\n" +
		"\n" +
		"fn bump() -> null {\n" +
		"\tcounter++\n" +
		"}\n" +
		"\n" +
		"fn down() -> null {\n" +
		"\tcounter--\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("postfix program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "p.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"setCounter((counter + 1));",
		"setCounter((counter - 1));",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("postfix jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestMutateErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"type U struct { name: string }\nfn f(u: U) -> null {\n\tu.nme = \"x\"\n}", "unknown field"},
		{"type U struct { age: int }\nfn f(u: U) -> null {\n\tu.age = \"x\"\n}", "cannot assign a value of type string to a value of type int"},
		{"state n: array[int] = [1]\nfn f() -> null {\n\tn[0] = \"x\"\n}", "cannot assign a value of type string to a value of type int"},
		{"state n = 1\nfn f() -> null {\n\tn.nope = 2\n}", "member access is not allowed on a value of type int"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
