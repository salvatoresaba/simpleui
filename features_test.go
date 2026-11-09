package main

import "strings"
import "testing"

// features_test.go - v0.2 language features: built-ins (len/append/string
// helpers/conversions), compound assignment and the while loop.

func TestBuiltinsTypecheckAndJs(t *testing.T) {
	src := "state items: array[string] = []\n" +
		"\n" +
		"fn addItem() -> null {\n" +
		"\titems = append(items, \"x\")\n" +
		"}\n" +
		"\n" +
		"fn main2() -> string {\n" +
		"\tlet s = upper(trim(\"  hi  \"))\n" +
		"\tlet parts = split(\"a,b,c\", \",\")\n" +
		"\tlet joined = join(parts, \"|\")\n" +
		"\tlet n = len(s) + len(parts)\n" +
		"\tlet ok = contains(s, \"HI\")\n" +
		"\tlet sub = substring(s, 0, 1)\n" +
		"\tlet c = str(3.5)\n" +
		"\tlet i = int(3.9)\n" +
		"\tlet f = float(3)\n" +
		"\tlet z = lower(joined) + c\n" +
		"\treturn z\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("builtins program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "b.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"[...items, `x`]",
		"`  hi  `.trim().toUpperCase()",
		"`a,b,c`.split(`,`)",
		"parts.join(`|`)",
		"s.length",
		"s.includes(`HI`)",
		"substring(0, 0+1)",
		"String(3.5)",
		"Math.trunc(3.9)",
		"Number(3)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestBuiltinsTypeErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn f() -> int { return len(1) }", "len requires a string or an array"},
		{"fn f() -> int { return len(\"a\", \"b\") }", "takes exactly 1 arguments"},
		{"fn f() -> array[int] { return append([1, 2], \"x\") }", "cannot add a value of type string to an array of int"},
		{"fn f() -> string { return upper(3) }", "upper expects a string"},
		{"fn f() -> bool { return contains(1, \"a\") }", "contains expects strings"},
		{"fn f() -> string { return join([1], \",\") }", "join expects an array of strings"},
		{"fn f() -> int { return substring(\"abc\", 0, 0.5) }", "offsets must be ints"},
		{"fn f() -> int { return int(\"abc\") }", "int expects a number"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}

func TestCompoundAssignDesugar(t *testing.T) {
	src := "state counter = 0\n" +
		"state name = \"\"\n" +
		"\n" +
		"fn bump() -> null {\n" +
		"\tcounter += 1\n" +
		"}\n" +
		"\n" +
		"fn rename() -> null {\n" +
		"\tname += \"x\"\n" +
		"\tcounter *= 2\n" +
		"\tcounter -= 1\n" +
		"\tcounter /= 2\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("compound program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "c.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"setCounter((counter + 1));",
		"setName((name + `x`));",
		"setCounter((counter * 2));",
		"setCounter((counter - 1));",
		"Math.trunc(counter / 2)",
		"Math.trunc(counter / 2)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("compound jsgen missing %q", want)
		}
	}
}

func TestWhileLoop(t *testing.T) {
	src := "fn sumTo(n: int) -> int {\n" +
		"\tlet total = 0\n" +
		"\tlet i = 0\n" +
		"\twhile i < n {\n" +
		"\t\ttotal += i\n" +
		"\t\ti += 1\n" +
		"\t}\n" +
		"\treturn total\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("while program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "w.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "while ((i < n)) {") {
		t.Errorf("while jsgen:\n%v", js)
	}
	// A non-bool condition must be rejected.
	bad := "fn f() -> null {\n\twhile 1 {\n\t}\n}\n"
	dlb := checkSrc(t, bad)
	if dlb.Count() == 0 || !hasMsg(dlb, "while condition must be a bool") {
		t.Errorf("non-bool while condition not rejected:\n%v", dlb.String())
	}
}

func TestBuiltinShadowingByUserFn(t *testing.T) {
	src := "fn len(x: string) -> int { return 42 }\n" +
		"\n" +
		"fn call() -> int {\n" +
		"\treturn len(\"anything\")\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("shadow program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "sh.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	// The user `len` must win at the call site: `len("anything")` compiles to
	// a call, not to the builtin's `.length` member expression.
	if strings.Contains(js, "`anything`.length") || strings.Contains(js, "\"anything\".length") {
		t.Errorf("user len must shadow the builtin: %v", js)
	}
	if !strings.Contains(js, "function len(x) {") {
		t.Errorf("user len missing: %v", js)
	}
	if !strings.Contains(js, "return len(`anything`);") && !strings.Contains(js, "return len(\"anything\");") {
		t.Errorf("call to user len missing: %v", js)
	}
}
