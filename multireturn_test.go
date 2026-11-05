package main

import "strings"
import "testing"

// multireturn_test.go - user functions returning multiple values (v0.8):
//   fn divmod(a: int, b: int) -> (int, int) { return a / b, a % b }
//   ...  q, r = divmod(7, 3)

func TestMultiReturnSync(t *testing.T) {
	src := "fn divmod(a: int, b: int) -> (int, int) {\n" +
		"\treturn a / b, a % b\n" +
		"}\n" +
		"\n" +
		"fn use() -> int {\n" +
		"\tq, r = divmod(7, 3)\n" +
		"\treturn q + r\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("multi-return sync:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "mr.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"return [(Math.trunc(a / b)), (a % b)];",
		"const __sui_r = divmod(7, 3);",
		"const q = __sui_r[0];",
		"const r = __sui_r[1];",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("multi-return jsgen missing %q\n----\n%v", want, js)
		}
	}
	// The sync call must NOT be awaited (valid inside a non-async function).
	if strings.Contains(js, "await divmod") {
		t.Errorf("sync multi-return call must not await:\n%v", js)
	}
}

func TestMultiReturnAsync(t *testing.T) {
	src := "async fn both() -> (string, int) {\n" +
		"\treturn \"hi\", 7\n" +
		"}\n" +
		"\n" +
		"async fn use() -> int {\n" +
		"\ts, n = both()\n" +
		"\treturn n\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("multi-return async:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "ma.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "async function both()") {
		t.Errorf("async fn missing:\n%v", js)
	}
	if !strings.Contains(js, "const __sui_r = await both();") {
		t.Errorf("async multi call must await:\n%v", js)
	}
}

func TestMultiReturnErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn d() -> (int, int) { return 1, 2 }\nfn f() -> int { x = d()\nreturn x }", "returns 2 values"},
		{"fn d() -> (int, int) { return 1, 2 }\nfn f() -> int { x, y, z = d()\nreturn x }", "supports exactly two targets"},
		{"fn d() -> (int, int) { return 1 }\nfn f() -> int { a, b = d()\nreturn a }", "use \"return a, b\""},
		{"fn d() -> (int, int) { return 1, \"x\" }\nfn f() -> int { a, b = d()\nreturn a }", "cannot return a value of type string"},
		{"fn d() -> int { return 1 }\nfn f() -> int { a, b = d()\nreturn a }", "expected a two-value call on the right-hand side"},
		{"async fn d() -> (int, int) { return 1, 2 }\nfn f() -> int { a, b = d()\nreturn a }", "cannot be called from a non-async function"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
