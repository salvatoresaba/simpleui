package core

import "strings"
import "testing"

// strings_test.go - the extended string & collection built-ins (v0.6):
// replace, startsWith, endsWith, indexOf/lastIndexOf, slice, repeat,
// padLeft/padRight, trimLeft/trimRight, chars, fmt, concat, remove, includes,
// sort, reverse, removeKey.

func TestStringBuiltinsTypeAndJs(t *testing.T) {
	src := "fn probe() -> string {\n" +
		"\tlet s = \"  Hello World  \"\n" +
		"\tlet a = trim(s)\n" +
		"\tlet b = replace(a, \"World\", \"SUI\")\n" +
		"\tlet sw = startsWith(b, \"Hello\")\n" +
		"\tlet ew = endsWith(b, \"SUI\")\n" +
		"\tlet i = indexOf(b, \"SUI\")\n" +
		"\tlet li = lastIndexOf(b, \"l\")\n" +
		"\tlet sub = slice(b, 0, 5)\n" +
		"\tlet r = repeat(\"ab\", 3)\n" +
		"\tlet l = padLeft(\"7\", 3, \"0\")\n" +
		"\tlet rr = padRight(\"x\", 3, \".\")\n" +
		"\tlet tl = trimLeft(\"  z\")\n" +
		"\tlet tr = trimRight(\"z  \")\n" +
		"\tlet cs = chars(\"héllo\")\n" +
		"\tlet n = fmt(3.14159, 2)\n" +
		"\treturn n + r + l + rr + tl + tr + sub\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("string builtins:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "st.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"s.trim()",
		".replaceAll(`World`, `SUI`)",
		"b.startsWith(`Hello`)",
		"b.endsWith(`SUI`)",
		"b.indexOf(`SUI`)",
		"b.lastIndexOf(`l`)",
		"b.slice(0, 5)",
		"`ab`.repeat(3)",
		"`7`.padStart(3, `0`)",
		"`x`.padEnd(3, `.`)",
		"`  z`.trimStart()",
		"`z  `.trimEnd()",
		"[...`héllo`]",
		"3.14159.toFixed(2)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("string builtin jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestCollectionBuiltinsTypeAndJs(t *testing.T) {
	src := "fn coll() -> int {\n" +
		"\tlet a = [1, 2, 3]\n" +
		"\tlet c = concat(a, [4])\n" +
		"\tlet rm = remove(c, 0)\n" +
		"\tlet srt = sort([3, 1, 2])\n" +
		"\tlet rev = reverse(a)\n" +
		"\tlet f = includes([1, 2], 2)\n" +
		"\tlet m = { \"a\": 1, \"b\": 2 }\n" +
		"\tlet n = removeKey(m, \"a\")\n" +
		"\treturn len(rm) + len(srt) + len(rev)\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("collection builtins:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "cl.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"[...a, ...",
		"[...c.slice(0, 0), ...c.slice(0+1)]",
		"[...[3, 1, 2]].sort()",
		"[...a].reverse()",
		"includes(",
		"Object.fromEntries(Object.entries(m).filter((e) => e[0] !== `a`))",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("collection builtin jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestStringBuiltinErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn f() -> string { return repeat(\"a\", \"x\") }", "expects an int"},
		{"fn f() -> string { return slice(1, 0, 2) }", "expects a string"},
		{"fn f() -> bool { return startsWith(1, \"a\") }", "expects a string"},
		{"fn f() -> string { return padLeft(\"x\", 3, 2) }", "padLeft expects a string"},
		{"fn f() -> string { return fmt(\"a\", 2) }", "fmt expects a number"},
		{"fn f() -> int { return indexOf(5, 1) }", "indexOf expects a string or an array"},
		{"fn f() -> array[int] { return sort(3) }", "sort expects an array"},
		{"fn f() -> array[int] { return remove(3, 0) }", "remove expects an array"},
		{"fn f() -> bool { return includes(3, 1) }", "includes expects an array"},
		{"fn f(m: map[string]int) -> map[string]int { return removeKey(m, 3) }", "removeKey expects a string key"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
