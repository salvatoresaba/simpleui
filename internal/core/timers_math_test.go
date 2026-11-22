package core

import "strings"
import "testing"

// timers_math_test.go - timers (after / interval) and math built-ins (v0.7):
// min/max/abs/round/floor/ceil/pow/sqrt/random/randint.

func TestTimersMathTypeAndJs(t *testing.T) {
	src := "fn tick() -> null {\n}\n" +
		"fn poll() -> null {\n}\n" +
		"\n" +
		"fn math1() -> int {\n" +
		"\treturn round(3.4) + floor(3.9) + ceil(2.1)\n" +
		"}\n" +
		"\n" +
		"fn math2() -> float {\n" +
		"\treturn sqrt(16.0) + pow(2.0, 8.0)\n" +
		"}\n" +
		"\n" +
		"fn math3() -> bool {\n" +
		"\treturn min(3, 5) == 3 and max(1, 9) == 9 and abs(-4) == 4\n" +
		"}\n" +
		"\n" +
		"fn boot() -> null {\n" +
		"\tafter(1000, tick)\n" +
		"\tinterval(500, poll)\n" +
		"\trandint(1, 9)\n" +
		"\trandom()\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("timers/math program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "tm.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"setTimeout(() => tick(), 1000)",
		"setInterval(() => poll(), 500)",
		"Math.round(3.4)",
		"Math.floor(3.9)",
		"Math.ceil(2.1)",
		"Math.sqrt(16)",
		"Math.pow(2, 8)",
		"Math.min(3, 5)",
		"Math.max(1, 9)",
		"Math.abs((-4))",
		"Math.random()",
		"(Math.floor(Math.random() * (9 - 1 + 1)) + 1)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("timers/math jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestTimersMathErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn f() -> null {\n}\nfn g() -> null {\n\tafter(1000, 42)\n}", "expects a declared function name"},
		{"fn f() -> null {\n}\nfn g() -> null {\n\tafter(\"x\", f)\n}", "expects a number"},
		{"fn f() -> null {\n}\nfn g() -> null {\n\tinterval(1000)\n}", "takes exactly 2 arguments"},
		{"fn f() -> null {\n\trandom(1)\n}", "takes exactly 0 arguments"},
		{"fn f() -> bool { return min(\"a\", 1) }", "expects numbers"},
		{"fn f() -> float { return sqrt(\"x\") }", "expects a number"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
