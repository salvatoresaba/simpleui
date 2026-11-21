package sui

import "testing"
import "strings"

// astjson_test.go — tests for ProgToJson / AstJsonFromSource (Agent A,
// front-end).

func TestJsonKeyFields(t *testing.T) {
	src := "state s: int = 1\nfn f(x: int) -> int {\n\treturn x\n}\nview v {\n\th1 \"hi\"\n}\n"
	js, dl := AstJsonFromSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("json: %v diagnostics: %v", dl.Count(), jsDiagJson(dl))
	}
	if !strings.Contains(js, "\"kind\"") {
		t.Errorf("json missing \"kind\" key")
	}
	if !strings.Contains(js, "\"name\"") {
		t.Errorf("json missing \"name\" key")
	}
	if !strings.Contains(js, "\"body\"") {
		t.Errorf("json missing \"body\" key")
	}
	if !strings.Contains(js, "\"span\"") {
		t.Errorf("json missing \"span\" key")
	}
}

func jsDiagJson(dl DiagList) string {
	var sb strings.Builder
	for i := 0; i < len(dl.list); i++ {
		sb.WriteString("|")
		sb.WriteString(dl.list[i].msg)
	}
	return sb.String()
}

func TestJsonMatchesPipeline(t *testing.T) {
	src := "state counter: int = 0\nfn incr() -> int {\n\treturn counter\n}\nview main {\n\th1 \"Hi {counter}\"\n}\n"
	js1, dl1 := AstJsonFromSource(src, "t.sui")
	prog, dl2 := ParseSource(src, "t.sui")
	js2 := ProgToJson(prog)
	if js1 != js2 {
		t.Errorf("AstJsonFromSource != ParseSource+ProgToJson")
	}
	if dl1.Count() != 0 || dl2.Count() != 0 {
		t.Errorf("expected no diagnostics, got %v / %v", dl1.Count(), dl2.Count())
	}
}

func TestJsonDeterministic(t *testing.T) {
	src := "state n: int = 2\n"
	js1, _ := AstJsonFromSource(src, "t.sui")
	js2, _ := AstJsonFromSource(src, "t.sui")
	if js1 != js2 {
		t.Errorf("json output is not deterministic")
	}
}

func TestJsonScalarsEscaping(t *testing.T) {
	src := "fn f() -> string {\n\treturn \"a\\\"b\\\\c\\n\"\n}\n"
	js, dl := AstJsonFromSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("escaping: %v diagnostics: %v", dl.Count(), jsDiagJson(dl))
	}
	// The decoded text "a\"b\\c<newline>" must be JSON-escaped.
	if !strings.Contains(js, "\"a\\\"b\\\\c\\n\"") {
		t.Errorf("escaped text not found in json:\n%v", js)
	}
}
