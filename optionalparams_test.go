package sui

import "strings"
import "testing"

// optionalparams_test.go - function parameters with default values:
//   fn greet(name: string, excited: bool = true) -> string ...

func TestOptionalParams(t *testing.T) {
	src := "fn greet(name: string, excited: bool = true) -> string {\n" +
		"\tif excited {\n" +
		"\t\treturn name + \"!\"\n" +
		"\t}\n" +
		"\treturn name\n" +
		"}\n" +
		"\n" +
		"fn use() -> string {\n" +
		"\tlet a = greet(\"hi\")\n" +
		"\tlet b = greet(\"bye\", false)\n" +
		"\treturn a + b\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("optional params:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "op.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	// The omitted trailing default is appended at the call site.
	if !strings.Contains(js, "greet(`hi`, true)") {
		t.Errorf("default not appended at call site:\n%v", js)
	}
	if !strings.Contains(js, "greet(`bye`, false)") {
		t.Errorf("explicit argument lost:\n%v", js)
	}
}

func TestOptionalParamErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn f(a: int = 1, b: int) -> int {\n\treturn a + b\n}", "follows an optional parameter"},
		{"fn f(a: int, b: int = 1) -> int {\n\treturn a\n}\nfn g() -> int { return f() }", "missing argument for parameter 1"},
		{"fn f(a: int, b: bool = 1) -> int {\n\treturn a\n}", "default value of parameter \"b\" must be bool, got int"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
