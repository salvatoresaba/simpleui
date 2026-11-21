package sui

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sourceMapProgram(t *testing.T) (*Prog, string) {
	t.Helper()
	dir := t.TempDir()
	src := `css "p \{ color: red; \}"
state count = 0
fn inc() -> null {
	count = count + 1
}
view main {
	col {
		h1 "Counter"
		p "Value: {count}"
		button "Add" onClick: inc
	}
}
`
	path := filepath.Join(dir, "app.sui")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	prog, dl := LoadProgram(path)
	if dl.HasErrors() {
		t.Fatalf("load: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	return prog, dir
}

func TestSourceMapGeneration(t *testing.T) {
	prog, dir := sourceMapProgram(t)
	if js := GenerateJs(prog); strings.Contains(js, "@sui-src") {
		t.Fatal("normal build leaked source markers")
	}
	js, jsMap := GenerateJsWithSourceMap(prog, "web", dir)
	if strings.Contains(js, "@sui-src") {
		t.Fatal("clean JS still contains source markers")
	}
	if !strings.Contains(js, "function inc") {
		t.Fatal("generated JS lost the user function")
	}
	var doc struct {
		Version        int           `json:"version"`
		File           string        `json:"file"`
		Sources        []string      `json:"sources"`
		SourcesContent []interface{} `json:"sourcesContent"`
		Mappings       string        `json:"mappings"`
	}
	if err := json.Unmarshal([]byte(jsMap), &doc); err != nil {
		t.Fatalf("parse JS map: %v (%s)", err, jsMap[:minInt(len(jsMap), 200)])
	}
	if doc.Version != 3 || doc.File != "app.js" || len(doc.Sources) == 0 || doc.Mappings == "" {
		t.Fatalf("incomplete JS map: %+v", doc)
	}
	if doc.Sources[0] != "app.sui" {
		t.Fatalf("unexpected source name %q", doc.Sources[0])
	}
	if len(doc.SourcesContent) == 0 || doc.SourcesContent[0] == nil ||
		!strings.Contains(doc.SourcesContent[0].(string), "fn inc") {
		t.Fatalf("JS map lacks sourcesContent: %+v", doc.SourcesContent)
	}

	css, cssMap := GenerateCssWithSourceMap(prog, dir)
	if strings.Contains(css, "@sui-src") {
		t.Fatal("clean CSS still contains source markers")
	}
	var cssDoc struct {
		Version  int      `json:"version"`
		File     string   `json:"file"`
		Sources  []string `json:"sources"`
		Mappings string   `json:"mappings"`
	}
	if err := json.Unmarshal([]byte(cssMap), &cssDoc); err != nil {
		t.Fatalf("parse CSS map: %v", err)
	}
	if cssDoc.Version != 3 || cssDoc.File != "style.css" || len(cssDoc.Sources) == 0 || cssDoc.Mappings == "" {
		t.Fatalf("incomplete CSS map: %+v", cssDoc)
	}
}

func TestSourceMapServeAndSingle(t *testing.T) {
	prog, dir := sourceMapProgram(t)
	out := t.TempDir()
	hr := NewHotReloader()
	if err := BuildWebIntoOpts(filepath.Join(dir, "app.sui"), out, hr, false, true); err != nil {
		t.Fatalf("mapped build: %v", err)
	}
	for _, name := range []string{"app.js", "app.js.map", "style.css", "style.css.map"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	js, err := os.ReadFile(filepath.Join(out, "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(js), "//# sourceMappingURL=app.js.map") {
		t.Fatal("app.js lacks sourceMappingURL")
	}
	css, err := os.ReadFile(filepath.Join(out, "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(css), "sourceMappingURL=style.css.map") {
		t.Fatal("style.css lacks sourceMappingURL")
	}
	mapText, err := os.ReadFile(filepath.Join(out, "app.js.map"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]interface{}
	if err := json.Unmarshal(mapText, &doc); err != nil {
		t.Fatalf("app.js.map is not JSON: %v", err)
	}
	rec := httptest.NewRecorder()
	appServer(out, false, hr).ServeHTTP(rec, httptest.NewRequest("GET", "/app.js.map", nil))
	if rec.Code != 200 || !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("map serving: status=%d type=%q", rec.Code, rec.Header().Get("Content-Type"))
	}

	html := BuildSingleHTMLOpts(prog, false, true)
	if !strings.Contains(html, "sourceMappingURL=data:application/json;charset=utf-8;base64,") {
		t.Fatal("single build lacks inline source map")
	}
	if strings.Contains(html, "@sui-src") {
		t.Fatal("single build leaked source markers")
	}
	plain := BuildSingleHTMLOpts(prog, false, false)
	if strings.Contains(plain, "sourceMappingURL") {
		t.Fatal("source maps must be opt-in")
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestSourceMapMappingsDecode(t *testing.T) {
	prog, dir := sourceMapProgram(t)
	js, mapText := GenerateJsWithSourceMap(prog, "web", dir)
	var doc struct {
		Sources  []string `json:"sources"`
		Mappings string   `json:"mappings"`
	}
	if err := json.Unmarshal([]byte(mapText), &doc); err != nil {
		t.Fatalf("map JSON: %v", err)
	}
	segments := decodeMappings(t, doc.Mappings)
	if len(segments) == 0 {
		t.Fatal("no mappings decoded")
	}
	lines := strings.Split(js, "\n")
	foundState, foundFn := false, false
	for _, m := range segments {
		if m.genLine >= len(lines) || m.genCol > len(lines[m.genLine]) {
			t.Fatalf("mapping outside generated file: %+v", m)
		}
		if m.srcLine == 1 {
			foundState = true
		}
		if m.srcLine == 2 {
			foundFn = true
		}
	}
	if !foundState || !foundFn {
		t.Fatalf("missing SUI mappings (state=%v fn=%v): %+v", foundState, foundFn, segments[:minInt(len(segments), 8)])
	}
}

type decodedMapping struct {
	genLine, genCol int
	srcLine, srcCol int
}

func decodeMappings(t *testing.T, mappings string) []decodedMapping {
	t.Helper()
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var out []decodedMapping
	src, srcLine, srcCol := 0, 0, 0
	for lineIndex, line := range strings.Split(mappings, ";") {
		genCol := 0
		for _, segment := range strings.Split(line, ",") {
			if segment == "" {
				continue
			}
			values := []int{}
			value, shift := 0, 0
			for i := 0; i < len(segment); i++ {
				digit := strings.IndexByte(alphabet, segment[i])
				if digit < 0 {
					t.Fatalf("invalid VLQ character %q", segment[i])
				}
				value += (digit & 31) << shift
				if digit&32 != 0 {
					shift += 5
					continue
				}
				negative := value&1 != 0
				value >>= 1
				if negative {
					value = -value
				}
				values = append(values, value)
				value, shift = 0, 0
			}
			if len(values) != 4 {
				t.Fatalf("segment has %d fields: %q", len(values), segment)
			}
			genCol += values[0]
			src += values[1]
			srcLine += values[2]
			srcCol += values[3]
			out = append(out, decodedMapping{genLine: lineIndex, genCol: genCol, srcLine: srcLine, srcCol: srcCol})
		}
	}
	return out
}
