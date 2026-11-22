package core

import "strings"
import "testing"

// maps_test.go - map iteration (loop k, v in m / list k, v in m) and the map
// builtins (keys / values / merge / has).

func TestMapIteration(t *testing.T) {
	src := "state counts: map[string]int = { \"a\": 1, \"b\": 2 }\n" +
		"\n" +
		"fn total() -> int {\n" +
		"\tlet sum = 0\n" +
		"\tloop k, v in counts {\n" +
		"\t\tsum += v\n" +
		"\t}\n" +
		"\treturn sum\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tlist k, v in counts {\n" +
		"\t\tp \"{k}={v}\"\n" +
		"\t}\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("map program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "m.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"for (const __sui_k of Object.entries(counts))",
		"const k = __sui_k[0];",
		"const v = __sui_k[1];",
		"${__sui_esc(k)}=${__sui_esc(v)}",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("map iteration jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestMapIterationErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn f(m: map[string]int) -> null {\n\tloop k in m {\n\t}\n}", "iterating a map requires two names"},
		{"fn f() -> null {\n\tloop k, v in [1, 2] {\n\t}\n}", "two-name iteration requires a map"},
		{"view main {\n\tlist k in {\"a\": 1} {\n\t}\n}", "requires two names"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}

func TestMapBuiltins(t *testing.T) {
	src := "state m: map[string]int = { \"a\": 1, \"b\": 2 }\n" +
		"\n" +
		"fn probe() -> int {\n" +
		"\tlet ks = keys(m)\n" +
		"\tlet vs = values(m)\n" +
		"\tlet merged = merge(m, { \"c\": 3 })\n" +
		"\tlet found = has(m, \"a\")\n" +
		"\tlet n = len(ks) + len(vs)\n" +
		"\tif found and has(merged, \"c\") {\n" +
		"\t\tn += 1\n" +
		"\t}\n" +
		"\treturn n\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("map builtins:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "mb.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"Object.keys(m)",
		"Object.values(m)",
		"Object.assign({}, m, (",
		"Object.prototype.hasOwnProperty.call(m, `a`)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("map builtin jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestMapBuiltinErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn f() -> array[string] { return keys(5) }", "keys expects a map"},
		{"fn f(m: map[string]int) -> map[string]int { return merge(m, 3) }", "merge expects two maps"},
		{"fn f(m: map[string]int) -> bool { return has(m, 3) }", "has expects a string key"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
