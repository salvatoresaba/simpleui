package sui

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

func TestAppearanceGalleryHasOneNativeDatePicker(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("examples", "appearance_gallery.html"))
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(body), `kind: "native"`); count != 1 {
		t.Fatalf("appearance gallery has %d native date picker entries, want 1", count)
	}
	if !strings.Contains(string(body), `label: "Native picker", kind: "native"`) {
		t.Fatal("appearance gallery must retain the named Native picker example")
	}
	for _, structure := range []string{"Vertical tab rail", "Separated accordion cards", "Glass toast", "Compact toast"} {
		if !strings.Contains(string(body), structure) {
			t.Errorf("appearance gallery is missing %q", structure)
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

func TestAppearanceLibraryRejectsMotionForTextEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "controls": [{
    "kind": "input",
    "structures": [{"id": "default", "label": "Default"}],
    "styles": [{"id": "dark", "label": "Dark"}, {"id": "light", "label": "Light"}],
    "animations": [
      {"id": "none", "label": "None", "value": "none"},
      {"id": "lift", "label": "Lift", "value": "sui-fx-lift .18s ease-out forwards"}
    ]
  }]
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadAppearanceLibraryFile(path)
	if err == nil || !strings.Contains(err.Error(), "text-entry controls only allow the None animation") {
		t.Fatalf("text-entry motion should be rejected, got %v", err)
	}
}

func TestAppearanceLibraryRejectsMotionStaticFeedback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "feedback": {"static": {"hover": "sui-fx-lift .2s ease-out", "press": "none", "focus": "none", "state": "none"}},
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
	if err == nil || !strings.Contains(err.Error(), "feedback static hover must not move") {
		t.Fatalf("static feedback motion should be rejected, got %v", err)
	}
}

func TestMCPAppearanceTool(t *testing.T) {
	text, isErr := mcpCallTool("sui_appearance", JObj("path", JStr(filepath.Join("components", "appearance.json"))))
	if isErr {
		t.Fatalf("sui_appearance failed: %s", text)
	}
	if !strings.Contains(text, `"ok":true`) || !strings.Contains(text, `"controls":36`) {
		t.Fatalf("unexpected sui_appearance result: %s", text)
	}
}

func TestAppearanceLibraryRejectsUnknownKeyframe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "controls": [{
    "kind": "accordion",
    "structures": [{"id": "default", "label": "Default"}],
    "styles": [{"id": "dark", "label": "Dark"}, {"id": "light", "label": "Light"}],
    "animations": [
      {"id": "none", "label": "None", "value": "none"},
      {"id": "expand", "label": "Expand", "value": "sui-accordion-expand .3s ease-out"}
    ]
  }]
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadAppearanceLibraryFile(path)
	if err == nil || !strings.Contains(err.Error(), "unknown keyframe") {
		t.Fatalf("unknown keyframe should be rejected, got %v", err)
	}
}

func TestBundledAppearanceLibraryDefinitions(t *testing.T) {
	catalog, _, err := loadAppearanceLibraryFile(filepath.Join("components", "appearance.json"))
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Definitions == nil {
		t.Fatal("bundled catalog is missing the self-contained definitions")
	}
	if catalog.Definitions.BaseCSS == "" || catalog.Definitions.PhaseCSS == "" || catalog.Definitions.GenericKeyframesCSS == "" {
		t.Fatal("bundled definitions must include baseCss, phaseCss and genericKeyframesCss")
	}
	if len(catalog.Definitions.Keyframes) < 100 {
		t.Fatalf("bundled definitions list only %d keyframes", len(catalog.Definitions.Keyframes))
	}
	for _, c := range catalog.Controls {
		if c.Markup == "" || len(c.Classes) == 0 || len(c.Keyframes) == 0 {
			t.Fatalf("%s: markup, classes and keyframes are required", c.Kind)
		}
		if c.Kind == "switch" {
			found := false
			for _, st := range c.Structures {
				if st.ID == "ios" && st.Class == "sui-sw-shape-ios" {
					found = true
				}
			}
			if !found {
				t.Fatal("switch iOS structure must declare its shape class")
			}
		}
	}
}

func TestAppearanceLibraryRejectsUndefinedClass(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearance.json")
	body := `{
  "format": "simpleui-appearance-library",
  "version": 1,
  "definitions": {
    "baseCss": ".sui-button { color: red; }",
    "phaseCss": ".sui-anim-hover:hover { animation: var(--sui-hover-animation, none); }",
    "genericKeyframesCss": "@keyframes sui-fx-ring { 0% { box-shadow: none; } 100% { box-shadow: 0 0 0 4px transparent; } }",
    "phaseClasses": ["sui-anim-hover"],
    "keyframes": {"sui-fx-ring": "@keyframes sui-fx-ring { 0% { box-shadow: none; } 100% { box-shadow: 0 0 0 4px transparent; } }"}
  },
  "feedback": {"static": {"hover": "sui-fx-ring .2s", "press": "none", "focus": "none", "state": "none"}},
  "controls": [{
    "kind": "button",
    "markup": "<button class=\"sui-button sui-missing\">X</button>",
    "classes": ["sui-button", "sui-missing"],
    "css": "",
    "keyframes": ["sui-fx-ring"],
    "structures": [{"id": "rect", "label": "Rect"}],
    "styles": [{"id": "dark", "label": "Dark"}, {"id": "light", "label": "Light"}],
    "animations": [{"id": "none", "label": "None", "value": "none"}]
  }]
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadAppearanceLibraryFile(path)
	if err == nil || !strings.Contains(err.Error(), "sui-missing") {
		t.Fatalf("undefined class should be rejected, got %v", err)
	}
}

func TestAppearanceKeyframeAllowListMatchesCSS(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`@keyframes\s+(sui-[A-Za-z0-9_-]+)`)
	defined := map[string]bool{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range re.FindAllSubmatch(data, -1) {
			defined[string(match[1])] = true
		}
	}
	for name := range defined {
		if !appearanceKeyframes[name] {
			t.Errorf("keyframe %q is defined but missing from appearanceKeyframes", name)
		}
	}
	for name := range appearanceKeyframes {
		if !defined[name] {
			t.Errorf("appearanceKeyframes lists %q but no CSS defines it", name)
		}
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
	for kind := range scriptKinds {
		if !appearanceLibraryKinds[kind] {
			t.Fatalf("designer lists unsupported appearance kind %q", kind)
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
