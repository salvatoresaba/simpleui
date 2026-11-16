package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLibraryFile(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "components"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "components", "library.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestComponentLibraryValid(t *testing.T) {
	body := `{
  "format": "simpleui-component-library",
  "version": 1,
  "groups": [{
    "id": "custom",
    "label": "Custom",
    "components": [{
      "id": "metric",
      "label": "Metric",
      "kind": "card",
      "class": "lib-metric",
      "css": ".lib-metric { padding: 12px; border-radius: 10px; background: #111827; color: #ffffff; }",
      "children": [{"kind": "p", "text": "Revenue"}]
    }]
  }]
}`
	groups, warnings, err := loadComponentLibraryFile(writeLibraryFile(t, body))
	if err != nil {
		t.Fatalf("valid library rejected: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Components) != 1 {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestBundledComponentLibraryValid(t *testing.T) {
	groups, _, err := loadComponentLibraryFile(filepath.Join("components", "library.json"))
	if err != nil {
		t.Fatalf("bundled components/library.json is invalid: %v", err)
	}
	if len(groups) == 0 {
		t.Fatal("bundled library has no groups")
	}
}

func TestComponentLibraryRejectsUnknownField(t *testing.T) {
	body := `{"format":"simpleui-component-library","version":1,"unexpected":true,"groups":[]}`
	if _, _, err := loadComponentLibraryFile(writeLibraryFile(t, body)); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field was not rejected: %v", err)
	}
}

func TestComponentLibraryRejectsBadFormat(t *testing.T) {
	body := `{"format":"other","version":1,"groups":[]}`
	if _, _, err := loadComponentLibraryFile(writeLibraryFile(t, body)); err == nil {
		t.Fatal("bad format was accepted")
	}
}

func TestComponentLibraryRejectsUnsupportedKind(t *testing.T) {
	body := `{"format":"simpleui-component-library","version":1,"groups":[{"id":"g","label":"G","components":[{"id":"c","label":"C","kind":"script"}]}]}`
	if _, _, err := loadComponentLibraryFile(writeLibraryFile(t, body)); err == nil || !strings.Contains(err.Error(), "unsupported kind") {
		t.Fatalf("unsupported kind was not rejected: %v", err)
	}
}

func TestComponentLibraryRejectsUnsafeCss(t *testing.T) {
	for _, css := range []string{
		".x { position: fixed; }",
		".x { background: url(https://example.com/x.png); }",
		".x { content: \"x\"; }",
		".x { color: #fff; } .x { color: #000; }",
	} {
		body := `{"format":"simpleui-component-library","version":1,"groups":[{"id":"g","label":"G","components":[{"id":"c","label":"C","kind":"card","class":"x","css":"` + strings.ReplaceAll(css, `"`, `\"`) + `"}]}]}`
		if _, _, err := loadComponentLibraryFile(writeLibraryFile(t, body)); err == nil {
			t.Fatalf("unsafe css was accepted: %s", css)
		}
	}
}

func TestComponentLibraryRejectsDuplicateIDs(t *testing.T) {
	body := `{"format":"simpleui-component-library","version":1,"groups":[{"id":"g","label":"G","components":[{"id":"dup","label":"A","kind":"card"},{"id":"dup","label":"B","kind":"card"}]}]}`
	if _, _, err := loadComponentLibraryFile(writeLibraryFile(t, body)); err == nil || !strings.Contains(err.Error(), "duplicate component id") {
		t.Fatalf("duplicate id was not rejected: %v", err)
	}
}

func TestComponentLibraryRejectsTooDeep(t *testing.T) {
	inner := `{"kind":"p","text":"x"}`
	for i := 0; i < 7; i++ {
		inner = `{"kind":"div","children":[` + inner + `]}`
	}
	body := `{"format":"simpleui-component-library","version":1,"groups":[{"id":"g","label":"G","components":[{"id":"c","label":"C","kind":"div","children":[` + inner + `]}]}]}`
	if _, _, err := loadComponentLibraryFile(writeLibraryFile(t, body)); err == nil {
		t.Fatal("deeply nested component was accepted")
	}
}

func TestComponentLibraryMissingFileKeepsFallback(t *testing.T) {
	dir := t.TempDir()
	// The payload loader must not panic or fail the GUI when there is no file.
	payload := string(loadComponentLibraryPayload(dir))
	if !strings.Contains(payload, `"ok":true`) {
		t.Fatalf("missing library should be a soft fallback, got %s", payload)
	}
}
