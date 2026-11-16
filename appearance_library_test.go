package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestBundledAppearanceLibraryValid(t *testing.T) {
	catalog, warnings, err := loadAppearanceLibraryFile(filepath.Join("components", "appearance.json"))
	if err != nil {
		t.Fatalf("bundled components/appearance.json is invalid: %v", err)
	}
	controls := catalog.Controls
	if len(controls) != len(appearanceLibraryKinds) {
		t.Fatalf("bundled catalog covers %d controls, want %d", len(controls), len(appearanceLibraryKinds))
	}
	if len(warnings) != 0 {
		t.Fatalf("bundled catalog should cover every control, got warnings: %v", warnings)
	}
	if len(catalog.Feedback) < 4 {
		t.Fatalf("bundled catalog needs the four feedback variants, got %d", len(catalog.Feedback))
	}
	for variant, fb := range catalog.Feedback {
		if fb.Hover == "" || fb.Press == "" || fb.Focus == "" || fb.State == "" {
			t.Fatalf("feedback %q must cover hover, press, focus and state", variant)
		}
	}
	for _, c := range controls {
		hasDark, hasLight, hasNone := false, false, false
		for _, s := range c.Styles {
			hasDark = hasDark || s.ID == "dark"
			hasLight = hasLight || s.ID == "light"
		}
		for _, a := range c.Animations {
			hasNone = hasNone || a.ID == "none"
		}
		if !hasDark || !hasLight {
			t.Fatalf("%s: Dark and Light styles are required", c.Kind)
		}
		if !hasNone {
			t.Fatalf("%s: None animation is required", c.Kind)
		}
	}
}

func TestAppearanceLibraryRejectsUnsafeCSS(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "controls": [{
    "kind": "button",
    "structures": [{"id": "rect", "label": "Rect", "vars": {"--sui-button-radius": "8px; background: url(javascript:alert(1))"}}],
    "styles": [
      {"id": "dark", "label": "Dark"},
      {"id": "light", "label": "Light"}
    ],
    "animations": [{"id": "none", "label": "None", "value": "none"}]
  }]
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadAppearanceLibraryFile(path)
	if err == nil {
		t.Fatal("unsafe CSS value was accepted")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppearanceLibraryRejectsMissingDarkLight(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "controls": [{
    "kind": "button",
    "structures": [{"id": "rect", "label": "Rect"}],
    "styles": [{"id": "dark", "label": "Dark"}],
    "animations": [{"id": "none", "label": "None", "value": "none"}]
  }]
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadAppearanceLibraryFile(path)
	if err == nil || !strings.Contains(err.Error(), "Dark and Light") {
		t.Fatalf("missing Light style should be rejected, got %v", err)
	}
}

func TestAppearanceLibraryRejectsUnknownAttribute(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "controls": [{
    "kind": "switch",
    "structures": [{"id": "round", "label": "Round", "attrs": {"onclick": "alert(1)"}}],
    "styles": [
      {"id": "dark", "label": "Dark"},
      {"id": "light", "label": "Light"}
    ],
    "animations": [{"id": "none", "label": "None", "value": "none"}]
  }]
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadAppearanceLibraryFile(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported attribute") {
		t.Fatalf("unknown attribute should be rejected, got %v", err)
	}
}

func TestAppearanceLibraryRejectsUnsafeFeedback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "feedback": {"default": {"hover": "sui-fx-lift .2s; url(javascript:alert(1))", "press": "none", "focus": "none", "state": "none"}},
  "controls": [{
    "kind": "button",
    "structures": [{"id": "rect", "label": "Rect"}],
    "styles": [{"id": "dark", "label": "Dark"}, {"id": "light", "label": "Light"}],
    "animations": [{"id": "none", "label": "None", "value": "none"}]
  }]
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadAppearanceLibraryFile(path)
	if err == nil || !strings.Contains(err.Error(), "feedback") {
		t.Fatalf("unsafe feedback should be rejected, got %v", err)
	}
}

func TestAppearanceKindsMatchDesignerScript(t *testing.T) {
	data, err := os.ReadFile("gui_components_script.go")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?s)var APPEARANCE_KINDS = \[(.*?)\];`)
	m := re.FindSubmatch(data)
	if m == nil {
		t.Fatal("APPEARANCE_KINDS not found in the designer script")
	}
	scriptKinds := map[string]bool{}
	for _, q := range regexp.MustCompile(`"([a-z0-9]+)"`).FindAllSubmatch(m[1], -1) {
		scriptKinds[string(q[1])] = true
	}
	if len(scriptKinds) != len(appearanceLibraryKinds) {
		t.Fatalf("designer has %d appearance kinds, Go validator has %d", len(scriptKinds), len(appearanceLibraryKinds))
	}
	for kind := range appearanceLibraryKinds {
		if !scriptKinds[kind] {
			t.Fatalf("Go validator lists %q but the designer does not", kind)
		}
	}
	for kind := range scriptKinds {
		if !appearanceLibraryKinds[kind] {
			t.Fatalf("designer lists %q but the Go validator does not", kind)
		}
	}
}

func TestAppearanceLibraryPayloadFallback(t *testing.T) {
	t.Chdir(t.TempDir())
	payload := loadAppearanceLibraryPayload("")
	if !strings.Contains(string(payload), `"ok":false`) {
		t.Fatalf("missing catalog should produce ok:false, got %s", payload)
	}
	if !strings.Contains(string(payload), "built-in catalog") {
		t.Fatalf("fallback payload should mention the built-in catalog, got %s", payload)
	}
}
