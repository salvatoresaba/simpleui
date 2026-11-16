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
	Format      string                            `json:"format"`
	Version     int                               `json:"version"`
	Definitions *AppearanceDefinitionsJSON        `json:"definitions,omitempty"`
	Feedback    map[string]AppearanceFeedbackJSON `json:"feedback,omitempty"`
	Controls    []AppearanceControlJSON           `json:"controls"`
}

// AppearanceDefinitionsJSON is the self-contained CSS/keyframe layer. baseCss
// is the full base stylesheet; phaseCss defines the sui-anim-* phase classes;
// genericKeyframesCss is the shared sui-fx-*/sui-appearance-* keyframe block;
// keyframes maps every animation name to its full @keyframes rule.
type AppearanceDefinitionsJSON struct {
	BaseCSS             string            `json:"baseCss"`
	PhaseCSS            string            `json:"phaseCss"`
	GenericKeyframesCSS string            `json:"genericKeyframesCss"`
	PhaseClasses        []string          `json:"phaseClasses"`
	Keyframes           map[string]string `json:"keyframes"`
}

type AppearanceFeedbackJSON struct {
	Hover string `json:"hover"`
	Press string `json:"press"`
	Focus string `json:"focus"`
	State string `json:"state"`
}

type appearanceCatalogData struct {
	Controls    []AppearanceControlJSON
	Feedback    map[string]AppearanceFeedbackJSON
	Definitions *AppearanceDefinitionsJSON
}

type AppearanceControlJSON struct {
	Kind       string                    `json:"kind"`
	Label      string                    `json:"label,omitempty"`
	Markup     string                    `json:"markup,omitempty"`
	Classes    []string                  `json:"classes,omitempty"`
	CSS        string                    `json:"css,omitempty"`
	Keyframes  []string                  `json:"keyframes,omitempty"`
	Structures []AppearanceVariantJSON   `json:"structures"`
	Styles     []AppearanceVariantJSON   `json:"styles"`
	Animations []AppearanceAnimationJSON `json:"animations"`
}

type AppearanceVariantJSON struct {
	ID    string            `json:"id"`
	Label string            `json:"label"`
	Class string            `json:"class,omitempty"`
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
	appearanceTokenRe     = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	appearanceClassRe     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]{0,59}$`)
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

// appearanceKeyframes is the allow-list of keyframes defined by the base and
// component CSS. The catalog may only reference these names; typos and stale
// options from an older catalog are rejected.
var appearanceKeyframes = map[string]bool{
	"sui-alert-bounce-in":        true,
	"sui-alert-breathe":          true,
	"sui-alert-flash":            true,
	"sui-alert-in":               true,
	"sui-alert-pulse":            true,
	"sui-alert-shake":            true,
	"sui-alert-slide":            true,
	"sui-alert-slide-right":      true,
	"sui-alert-wobble":           true,
	"sui-alert-zoom":             true,
	"sui-anim-fade":              true,
	"sui-anim-scale":             true,
	"sui-anim-slide-down":        true,
	"sui-anim-slide-left":        true,
	"sui-anim-slide-right":       true,
	"sui-anim-slide-up":          true,
	"sui-appearance-fade":        true,
	"sui-appearance-float":       true,
	"sui-appearance-glow":        true,
	"sui-appearance-pulse":       true,
	"sui-badge-blink":            true,
	"sui-badge-bounce":           true,
	"sui-badge-float":            true,
	"sui-badge-glow":             true,
	"sui-badge-jelly":            true,
	"sui-badge-ping":             true,
	"sui-badge-pop":              true,
	"sui-badge-pulse":            true,
	"sui-badge-sheen":            true,
	"sui-badge-tilt":             true,
	"sui-button-bounce":          true,
	"sui-button-float":           true,
	"sui-button-glow":            true,
	"sui-button-jelly":           true,
	"sui-button-press":           true,
	"sui-button-pulse":           true,
	"sui-button-ring":            true,
	"sui-button-sheen":           true,
	"sui-button-tilt":            true,
	"sui-card-drift":             true,
	"sui-card-float":             true,
	"sui-card-halo":              true,
	"sui-card-in":                true,
	"sui-card-rise":              true,
	"sui-card-tilt-in":           true,
	"sui-card-unfold":            true,
	"sui-card-zoom":              true,
	"sui-chart-cell-flash":       true,
	"sui-chart-table-flash":      true,
	"sui-chart-task-flash":       true,
	"sui-check-bounce":           true,
	"sui-check-jelly":            true,
	"sui-check-pop":              true,
	"sui-check-ring":             true,
	"sui-field-glow":             true,
	"sui-field-lift":             true,
	"sui-field-ring":             true,
	"sui-fx-bounce":              true,
	"sui-fx-glow":                true,
	"sui-fx-grow":                true,
	"sui-fx-jelly":               true,
	"sui-fx-lift":                true,
	"sui-fx-pop":                 true,
	"sui-fx-press":               true,
	"sui-fx-ring":                true,
	"sui-fx-shake":               true,
	"sui-fx-slide":               true,
	"sui-fx-tilt":                true,
	"sui-progress-aurora":        true,
	"sui-progress-breathe":       true,
	"sui-progress-glow":          true,
	"sui-progress-indeterminate": true,
	"sui-progress-pulse":         true,
	"sui-progress-shimmer":       true,
	"sui-progress-spark":         true,
	"sui-progress-stripes":       true,
	"sui-progress-ticker":        true,
	"sui-progress-wave":          true,
	"sui-sk":                     true,
	"sui-spin":                   true,
	"sui-spin-reverse":           true,
	"sui-spinner-bars":           true,
	"sui-spinner-bounce":         true,
	"sui-spinner-comet":          true,
	"sui-spinner-dots":           true,
	"sui-spinner-flip":           true,
	"sui-spinner-orbit":          true,
	"sui-spinner-pulse":          true,
	"sui-spinner-ripple":         true,
	"sui-spinner-swing":          true,
	"sui-spinner-wobble":         true,
	"sui-switch-bounce":          true,
	"sui-switch-fade":            true,
	"sui-switch-jelly":           true,
	"sui-switch-morph":           true,
	"sui-switch-pop":             true,
	"sui-switch-pulse":           true,
	"sui-switch-scale":           true,
	"sui-tab-glow":               true,
	"sui-tab-pop":                true,
	"sui-tab-slide":              true,
	"sui-toast-bounce":           true,
	"sui-toast-fade":             true,
	"sui-toast-flip":             true,
	"sui-toast-in":               true,
	"sui-toast-left":             true,
	"sui-toast-pop":              true,
	"sui-toast-rise":             true,
	"sui-toast-slide":            true,
	"sui-toast-squash":           true,
	"sui-toast-swing":            true,
	"sui-toast-zoom":             true,
	"sui-tooltip-in":             true,
}

// appearanceTextKinds mirrors APPEARANCE_TEXT_KINDS. These controls accept
// typing, so the catalog must not move or resize them.
var appearanceTextKinds = map[string]bool{
	"input": true, "textarea": true, "select": true, "combobox": true,
	"datepicker": true, "number": true, "color": true, "date": true,
	"time": true, "file": true,
}

var appearanceMotionRe = regexp.MustCompile(`sui-fx-(lift|grow|tilt|press|pop|jelly|bounce|shake|slide)|sui-appearance-(pulse|float|fade)|sui-field-lift`)

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
		return appearancePayload(false, "", nil, nil, nil, []string{"no external appearance library found; using the built-in catalog"})
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
	return appearancePayload(true, path, catalog.Controls, catalog.Feedback, catalog.Definitions, warnings)
}

func appearancePayload(ok bool, source string, controls []AppearanceControlJSON, feedback map[string]AppearanceFeedbackJSON, definitions *AppearanceDefinitionsJSON, warnings []string) []byte {
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
		"ok":          ok,
		"source":      source,
		"controls":    controls,
		"feedback":    feedback,
		"definitions": definitions,
		"warnings":    warnings,
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
			if err := validateAppearanceAnimationName(value); err != nil {
				return appearanceCatalogData{}, nil, fmt.Errorf("feedback %q %s: %v", variant, phase, err)
			}
			if variant == "static" && appearanceMotionRe.MatchString(value) {
				return appearanceCatalogData{}, nil, fmt.Errorf("feedback static %s must not move or resize the control: %s", phase, value)
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
			if err := validateAppearanceAnimationName(a.Value); err != nil {
				return appearanceCatalogData{}, nil, fmt.Errorf("%s: animation %q: %v", c.Kind, a.ID, err)
			}
			if a.ID == "none" || a.Value == "none" {
				hasNone = true
			}
		}
		if !hasNone {
			return appearanceCatalogData{}, nil, fmt.Errorf("%s: animations must include a None option", c.Kind)
		}
		if appearanceTextKinds[c.Kind] {
			if len(c.Animations) != 1 || c.Animations[0].ID != "none" || c.Animations[0].Value != "none" {
				return appearanceCatalogData{}, nil, fmt.Errorf("%s: text-entry controls only allow the None animation; use feedback.static for phase feedback", c.Kind)
			}
		}
	}
	if err := validateAppearanceDefinitions(lib.Definitions, lib.Controls, lib.Feedback); err != nil {
		return appearanceCatalogData{}, nil, err
	}
	if lib.Definitions == nil {
		warnings = append(warnings, "appearance catalog has no self-contained definitions (markup/classes/css/keyframes)")
	}
	for _, kind := range sortedAppearanceKinds() {
		if !seenControls[kind] {
			warnings = append(warnings, "appearance catalog has no entry for "+kind+"; the built-in catalog is used for that control")
		}
	}
	return appearanceCatalogData{Controls: lib.Controls, Feedback: lib.Feedback, Definitions: lib.Definitions}, warnings, nil
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
		if strings.Contains(strings.ToLower(k), "animation") {
			if err := validateAppearanceAnimationName(val); err != nil {
				return fmt.Errorf("%s: %s %q: %s: %v", kind, layer, v.ID, k, err)
			}
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

// validateAppearanceAnimationName checks the first token of an animation
// shorthand against the keyframes the base/component CSS actually defines.
// A valid-looking but undefined name (for example a stale sui-accordion-expand
// option) would otherwise be accepted and do nothing at runtime.
func validateAppearanceAnimationName(value string) error {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return fmt.Errorf("animation value must not be empty")
	}
	name := fields[0]
	if name == "none" {
		return nil
	}
	if !appearanceKeyframes[name] {
		return fmt.Errorf("unknown keyframe %q (not defined by SimpleUI; use a keyframe from docs/APPEARANCE_LIBRARY.md)", name)
	}
	return nil
}

// validateAppearanceCSSBlock validates a full stylesheet fragment (control CSS,
// base CSS, phase CSS or a keyframe block). These are definitions, not single
// declaration values, so semicolons and braces are allowed; executable markup
// and remote/script URLs are not.
func validateAppearanceCSSBlock(label, css string) error {
	if strings.TrimSpace(css) == "" {
		return fmt.Errorf("%s must not be empty", label)
	}
	if len(css) > 1<<20 {
		return fmt.Errorf("%s is larger than 1 MB", label)
	}
	low := strings.ToLower(css)
	for _, bad := range []string{"<script", "</style", "javascript:", "expression(", "@import", "behavior:"} {
		if strings.Contains(low, bad) {
			return fmt.Errorf("%s contains forbidden text %q", label, bad)
		}
	}
	return nil
}

func appearanceClassDefined(corpus, class string) bool {
	re := regexp.MustCompile(`\.` + regexp.QuoteMeta(class) + `([^A-Za-z0-9_-]|$)`)
	return re.MatchString(corpus)
}

func validateAppearanceDefinitions(defs *AppearanceDefinitionsJSON, controls []AppearanceControlJSON, feedback map[string]AppearanceFeedbackJSON) error {
	if defs == nil {
		return nil
	}
	if err := validateAppearanceCSSBlock("definitions.baseCss", defs.BaseCSS); err != nil {
		return err
	}
	if err := validateAppearanceCSSBlock("definitions.phaseCss", defs.PhaseCSS); err != nil {
		return err
	}
	if err := validateAppearanceCSSBlock("definitions.genericKeyframesCss", defs.GenericKeyframesCSS); err != nil {
		return err
	}
	if len(defs.PhaseClasses) == 0 {
		return fmt.Errorf("definitions.phaseClasses must not be empty")
	}
	for _, class := range defs.PhaseClasses {
		if !appearanceClassDefined(defs.PhaseCSS, class) {
			return fmt.Errorf("definitions.phaseClasses: %q is not defined in phaseCss", class)
		}
	}
	if len(defs.Keyframes) == 0 {
		return fmt.Errorf("definitions.keyframes must not be empty")
	}
	for name, block := range defs.Keyframes {
		if !appearanceTokenRe.MatchString(name) {
			return fmt.Errorf("definitions.keyframes: invalid name %q", name)
		}
		if !appearanceKeyframes[name] {
			return fmt.Errorf("definitions.keyframes: %q is not a SimpleUI keyframe", name)
		}
		if err := validateAppearanceCSSBlock("definitions.keyframes."+name, block); err != nil {
			return err
		}
		if !strings.HasPrefix(strings.TrimSpace(block), "@keyframes "+name) {
			return fmt.Errorf("definitions.keyframes.%s must start with @keyframes %s", name, name)
		}
	}
	for ci := range controls {
		c := &controls[ci]
		if strings.TrimSpace(c.Markup) == "" {
			return fmt.Errorf("%s: definitions.markup is required", c.Kind)
		}
		if len(c.Markup) > 100000 {
			return fmt.Errorf("%s: markup is larger than 100 KB", c.Kind)
		}
		lowMarkup := strings.ToLower(c.Markup)
		for _, bad := range []string{"<script", "</style", "javascript:"} {
			if strings.Contains(lowMarkup, bad) {
				return fmt.Errorf("%s: markup contains forbidden text %q", c.Kind, bad)
			}
		}
		if len(c.Classes) == 0 {
			return fmt.Errorf("%s: definitions.classes is required", c.Kind)
		}
		if c.CSS != "" {
			if err := validateAppearanceCSSBlock(c.Kind+".css", c.CSS); err != nil {
				return err
			}
		}
		corpus := defs.BaseCSS + "\n" + c.CSS + "\n" + defs.PhaseCSS
		seenClasses := map[string]bool{}
		for _, class := range c.Classes {
			if !appearanceClassRe.MatchString(class) {
				return fmt.Errorf("%s: invalid class %q", c.Kind, class)
			}
			if seenClasses[class] {
				return fmt.Errorf("%s: duplicate class %q", c.Kind, class)
			}
			seenClasses[class] = true
			if !appearanceClassDefined(corpus, class) {
				return fmt.Errorf("%s: class %q is not defined by baseCss, css or phaseCss", c.Kind, class)
			}
		}
		for _, st := range c.Structures {
			if st.Class == "" {
				continue
			}
			if !appearanceClassRe.MatchString(st.Class) {
				return fmt.Errorf("%s: structure %q has invalid class %q", c.Kind, st.ID, st.Class)
			}
			if !appearanceClassDefined(corpus, st.Class) {
				return fmt.Errorf("%s: structure %q class %q is not defined", c.Kind, st.ID, st.Class)
			}
		}
		if len(c.Keyframes) == 0 {
			return fmt.Errorf("%s: definitions.keyframes is required", c.Kind)
		}
		seenKeyframes := map[string]bool{}
		for _, name := range c.Keyframes {
			if seenKeyframes[name] {
				return fmt.Errorf("%s: duplicate keyframe %q", c.Kind, name)
			}
			seenKeyframes[name] = true
			if _, ok := defs.Keyframes[name]; !ok {
				return fmt.Errorf("%s: keyframe %q is missing from definitions.keyframes", c.Kind, name)
			}
		}
		used := map[string]bool{}
		for _, a := range c.Animations {
			fields := strings.Fields(strings.TrimSpace(a.Value))
			if len(fields) > 0 && fields[0] != "none" {
				used[fields[0]] = true
			}
		}
		variants := []string{"static"}
		if !appearanceTextKinds[c.Kind] {
			variants = make([]string, 0, len(feedback))
			for variant := range feedback {
				variants = append(variants, variant)
			}
		}
		for _, variant := range variants {
			fb := feedback[variant]
			for _, value := range []string{fb.Hover, fb.Press, fb.Focus, fb.State} {
				fields := strings.Fields(strings.TrimSpace(value))
				if len(fields) > 0 && fields[0] != "none" {
					used[fields[0]] = true
				}
			}
		}
		for name := range used {
			if !seenKeyframes[name] {
				return fmt.Errorf("%s: keyframe %q is used but missing from definitions.keyframes", c.Kind, name)
			}
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
