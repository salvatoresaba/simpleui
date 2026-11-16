package main

// External appearance catalog.
//
// components/appearance.json holds every structure, style and animation that
// the designer exposes for each control kind. The file is validated atomically:
// a schema error, an unsafe CSS value or a missing Dark/Light style rejects the
// whole file and the editor keeps the built-in catalog. The JSON is read by the
// GUI server and handed to the page through /__gui__/appearance.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	appearanceLibraryFormat    = "simpleui-appearance-library"
	appearanceLibraryVersion   = 1
	appearanceLibraryMaxSize   = 1 << 20 // 1 MB
	appearanceMaxControls      = 120
	appearanceMaxVariants      = 80
	appearanceMaxVars          = 60
	appearanceMaxAttrs         = 20
	appearanceMaxCSSValueBytes = 300
)

// AppearanceLibraryJSON is the on-disk and wire shape.
type AppearanceLibraryJSON struct {
	Format   string                            `json:"format"`
	Version  int                               `json:"version"`
	Feedback map[string]AppearanceFeedbackJSON `json:"feedback,omitempty"`
	Controls []AppearanceControlJSON           `json:"controls"`
}

type AppearanceFeedbackJSON struct {
	Hover string `json:"hover"`
	Press string `json:"press"`
	Focus string `json:"focus"`
	State string `json:"state"`
}

type appearanceCatalogData struct {
	Controls []AppearanceControlJSON
	Feedback map[string]AppearanceFeedbackJSON
}

type AppearanceControlJSON struct {
	Kind       string                    `json:"kind"`
	Label      string                    `json:"label,omitempty"`
	Structures []AppearanceVariantJSON   `json:"structures"`
	Styles     []AppearanceVariantJSON   `json:"styles"`
	Animations []AppearanceAnimationJSON `json:"animations"`
}

type AppearanceVariantJSON struct {
	ID    string            `json:"id"`
	Label string            `json:"label"`
	Vars  map[string]string `json:"vars,omitempty"`
	Attrs map[string]any    `json:"attrs,omitempty"`
}

type AppearanceAnimationJSON struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

var (
	appearanceKindRe      = regexp.MustCompile(`^[a-z][a-z0-9]{0,39}$`)
	appearanceVariantIDRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,39}$`)
	appearanceVarKeyRe    = regexp.MustCompile(`^--[A-Za-z0-9_-]{1,60}$`)
	appearanceCSSPropRe   = regexp.MustCompile(`^[a-z][a-z0-9-]{0,60}$`)
	appearanceAttrKeyRe   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]{0,40}$`)
)

// appearanceLibraryKinds mirrors APPEARANCE_KINDS in the designer script.
var appearanceLibraryKinds = map[string]bool{
	"button": true, "input": true, "textarea": true, "select": true,
	"combobox": true, "datepicker": true, "number": true, "slider": true,
	"checkbox": true, "radio": true, "switch": true, "color": true,
	"date": true, "time": true, "file": true, "progress": true,
	"spinner": true, "badge": true, "alert": true, "card": true,
	"tabs": true, "accordion": true, "menubar": true, "dropdown": true,
	"popover": true, "modal": true, "toast": true, "tree": true,
	"commandbar": true, "drawer": true, "splitter": true, "pane": true,
	"dropzone": true, "timeline": true, "carousel": true, "stepper": true,
}

// appearanceAttrKeys limits structure attributes to the ones the preview and
// the editor understand. Arbitrary keys could rewrite unrelated node fields.
var appearanceAttrKeys = map[string]bool{
	"shape": true, "size": true, "variant": true, "tone": true,
}

func appearanceLibraryPath(projectDir string) string {
	if projectDir != "" {
		p := filepath.Join(projectDir, "components", "appearance.json")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	p := filepath.Join("components", "appearance.json")
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}

func loadAppearanceLibraryPayload(projectDir string) []byte {
	path := appearanceLibraryPath(projectDir)
	if path == "" {
		return appearancePayload(false, "", nil, nil, []string{"no external appearance library found; using the built-in catalog"})
	}
	catalog, warnings, err := loadAppearanceLibraryFile(path)
	if err != nil {
		b, _ := json.Marshal(map[string]any{
			"ok":     false,
			"source": path,
			"error":  err.Error(),
		})
		return b
	}
	return appearancePayload(true, path, catalog.Controls, catalog.Feedback, warnings)
}

func appearancePayload(ok bool, source string, controls []AppearanceControlJSON, feedback map[string]AppearanceFeedbackJSON, warnings []string) []byte {
	if controls == nil {
		controls = []AppearanceControlJSON{}
	}
	if feedback == nil {
		feedback = map[string]AppearanceFeedbackJSON{}
	}
	if warnings == nil {
		warnings = []string{}
	}
	b, _ := json.Marshal(map[string]any{
		"ok":       ok,
		"source":   source,
		"controls": controls,
		"feedback": feedback,
		"warnings": warnings,
	})
	return b
}

func loadAppearanceLibraryFile(path string) (appearanceCatalogData, []string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return appearanceCatalogData{}, nil, fmt.Errorf("cannot read %s: %v", path, err)
	}
	if st.Size() > appearanceLibraryMaxSize {
		return appearanceCatalogData{}, nil, fmt.Errorf("%s is larger than %d bytes", path, appearanceLibraryMaxSize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return appearanceCatalogData{}, nil, fmt.Errorf("cannot read %s: %v", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var lib AppearanceLibraryJSON
	if err := dec.Decode(&lib); err != nil {
		return appearanceCatalogData{}, nil, fmt.Errorf("invalid JSON: %v", err)
	}
	if dec.More() {
		return appearanceCatalogData{}, nil, fmt.Errorf("unexpected data after the appearance object")
	}
	if lib.Format != appearanceLibraryFormat {
		return appearanceCatalogData{}, nil, fmt.Errorf("format must be %q", appearanceLibraryFormat)
	}
	if lib.Version != appearanceLibraryVersion {
		return appearanceCatalogData{}, nil, fmt.Errorf("unsupported version %d (expected %d)", lib.Version, appearanceLibraryVersion)
	}
	if len(lib.Controls) == 0 {
		return appearanceCatalogData{}, nil, fmt.Errorf("the appearance catalog has no controls")
	}
	if len(lib.Controls) > appearanceMaxControls {
		return appearanceCatalogData{}, nil, fmt.Errorf("too many controls (max %d)", appearanceMaxControls)
	}
	for variant, fb := range lib.Feedback {
		if !appearanceVariantIDRe.MatchString(variant) {
			return appearanceCatalogData{}, nil, fmt.Errorf("feedback %q has an invalid id", variant)
		}
		for phase, value := range map[string]string{"hover": fb.Hover, "press": fb.Press, "focus": fb.Focus, "state": fb.State} {
			if strings.TrimSpace(value) == "" {
				return appearanceCatalogData{}, nil, fmt.Errorf("feedback %q needs a %s value", variant, phase)
			}
			if err := validateAppearanceCSSValue(value); err != nil {
				return appearanceCatalogData{}, nil, fmt.Errorf("feedback %q %s: %v", variant, phase, err)
			}
		}
	}
	seenControls := map[string]bool{}
	warnings := []string{}
	for ci := range lib.Controls {
		c := &lib.Controls[ci]
		if !appearanceKindRe.MatchString(c.Kind) || !appearanceLibraryKinds[c.Kind] {
			return appearanceCatalogData{}, nil, fmt.Errorf("control %d: unknown kind %q", ci+1, c.Kind)
		}
		if seenControls[c.Kind] {
			return appearanceCatalogData{}, nil, fmt.Errorf("duplicate control kind %q", c.Kind)
		}
		seenControls[c.Kind] = true
		if c.Label != "" && (len(c.Label) > 80 || strings.ContainsAny(c.Label, "<>\n\r")) {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: invalid label", c.Kind)
		}
		if len(c.Structures) == 0 {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: at least one structure is required", c.Kind)
		}
		if len(c.Styles) == 0 {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: at least one style is required", c.Kind)
		}
		if len(c.Animations) == 0 {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: at least one animation is required", c.Kind)
		}
		if len(c.Structures) > appearanceMaxVariants || len(c.Styles) > appearanceMaxVariants || len(c.Animations) > appearanceMaxVariants {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: too many variants (max %d)", c.Kind, appearanceMaxVariants)
		}
		seen := map[string]bool{}
		for vi := range c.Structures {
			if err := validateAppearanceVariant(c.Kind, "structure", &c.Structures[vi], seen); err != nil {
				return appearanceCatalogData{}, nil, err
			}
		}
		seen = map[string]bool{}
		hasDark, hasLight := false, false
		for vi := range c.Styles {
			if err := validateAppearanceVariant(c.Kind, "style", &c.Styles[vi], seen); err != nil {
				return appearanceCatalogData{}, nil, err
			}
			if c.Styles[vi].ID == "dark" {
				hasDark = true
			}
			if c.Styles[vi].ID == "light" {
				hasLight = true
			}
		}
		if !hasDark || !hasLight {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: styles must include Dark and Light (ids \"dark\" and \"light\")", c.Kind)
		}
		seen = map[string]bool{}
		hasNone := false
		for ai := range c.Animations {
			a := &c.Animations[ai]
			if !appearanceVariantIDRe.MatchString(a.ID) {
				return appearanceCatalogData{}, nil, fmt.Errorf("%s: animation %d has invalid id %q", c.Kind, ai+1, a.ID)
			}
			if seen[a.ID] {
				return appearanceCatalogData{}, nil, fmt.Errorf("%s: duplicate animation id %q", c.Kind, a.ID)
			}
			seen[a.ID] = true
			if a.Label == "" || len(a.Label) > 60 || strings.ContainsAny(a.Label, "<>\n\r") {
				return appearanceCatalogData{}, nil, fmt.Errorf("%s: animation %q has an invalid label", c.Kind, a.ID)
			}
			if err := validateAppearanceCSSValue(a.Value); err != nil {
				return appearanceCatalogData{}, nil, fmt.Errorf("%s: animation %q: %v", c.Kind, a.ID, err)
			}
			if a.ID == "none" || a.Value == "none" {
				hasNone = true
			}
		}
		if !hasNone {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: animations must include a None option", c.Kind)
		}
	}
	for _, kind := range sortedAppearanceKinds() {
		if !seenControls[kind] {
			warnings = append(warnings, "appearance catalog has no entry for "+kind+"; the built-in catalog is used for that control")
		}
	}
	return appearanceCatalogData{Controls: lib.Controls, Feedback: lib.Feedback}, warnings, nil
}

func validateAppearanceVariant(kind, layer string, v *AppearanceVariantJSON, seen map[string]bool) error {
	if !appearanceVariantIDRe.MatchString(v.ID) {
		return fmt.Errorf("%s: %s has invalid id %q", kind, layer, v.ID)
	}
	if seen[v.ID] {
		return fmt.Errorf("%s: duplicate %s id %q", kind, layer, v.ID)
	}
	seen[v.ID] = true
	if v.Label == "" || len(v.Label) > 60 || strings.ContainsAny(v.Label, "<>\n\r") {
		return fmt.Errorf("%s: %s %q has an invalid label", kind, layer, v.ID)
	}
	if len(v.Vars) > appearanceMaxVars {
		return fmt.Errorf("%s: %s %q has too many variables (max %d)", kind, layer, v.ID, appearanceMaxVars)
	}
	for k, val := range v.Vars {
		if !appearanceVarKeyRe.MatchString(k) && !appearanceCSSPropRe.MatchString(k) {
			return fmt.Errorf("%s: %s %q has an invalid CSS key %q", kind, layer, v.ID, k)
		}
		if err := validateAppearanceCSSValue(val); err != nil {
			return fmt.Errorf("%s: %s %q: %s: %v", kind, layer, v.ID, k, err)
		}
	}
	if len(v.Attrs) > appearanceMaxAttrs {
		return fmt.Errorf("%s: %s %q has too many attributes (max %d)", kind, layer, v.ID, appearanceMaxAttrs)
	}
	for k, val := range v.Attrs {
		if !appearanceAttrKeyRe.MatchString(k) || !appearanceAttrKeys[k] {
			return fmt.Errorf("%s: %s %q has an unsupported attribute %q", kind, layer, v.ID, k)
		}
		switch tv := val.(type) {
		case string:
			if len(tv) > 200 || strings.ContainsAny(tv, "<>\n\r") {
				return fmt.Errorf("%s: %s %q has an invalid value for %q", kind, layer, v.ID, k)
			}
		case float64, bool:
		default:
			return fmt.Errorf("%s: %s %q attribute %q must be a string, number or boolean", kind, layer, v.ID, k)
		}
	}
	return nil
}

func validateAppearanceCSSValue(v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("value must not be empty")
	}
	if len(v) > appearanceMaxCSSValueBytes {
		return fmt.Errorf("value is longer than %d bytes", appearanceMaxCSSValueBytes)
	}
	if strings.ContainsAny(v, ";{}<>\\\n\r") {
		return fmt.Errorf("value contains a forbidden character")
	}
	low := strings.ToLower(v)
	for _, bad := range []string{"url(", "expression(", "javascript:", "@import", "!important", "/*", "*/"} {
		if strings.Contains(low, bad) {
			return fmt.Errorf("value contains forbidden text %q", bad)
		}
	}
	return nil
}

func sortedAppearanceKinds() []string {
	out := make([]string, 0, len(appearanceLibraryKinds))
	for k := range appearanceLibraryKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
