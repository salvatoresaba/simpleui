package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SecurityConfig is the compiler's security policy for an application.
// Defaults are strict; a project can relax selected *compile-time* policies
// with simpleui.security.json (or --security <file>).
//
// Non-configurable invariants (always enforced):
//   - HTML escaping of interpolated text/attributes;
//   - table cells written with textContent;
//   - the runtime URL-scheme filter for href/src/link;
//   - runtime iframe URL validation and its visible error message;
//   - bridge token, loopback Host check, symlink-safe allow-dirs.
type SecurityConfig struct {
	StyleDynamic      string // safe (require cssSafe) | warn | allow
	ClassDynamic      string // filter | allow
	URLDynamic        string // safe (require *Safe + allow-list) | warn | allow
	StaticURLSchemes  string // error | warn | allow
	MaxRows           int
	MaxCols           int
	MaxText           int
	CSVFormulaEscape  bool
	IframeSandbox     string // default (auto sandbox="") | require | allow
	IframeReferrer    string // referrerpolicy default, or unset
	IframeScriptsSame string // error | warn | allow (allow-scripts + allow-same-origin)
	IframeDynamicSrc  string // allowlist | sandbox | allow
	PDFDynamicSrc     string // allowlist | allow (dynamic pdfviewer src)
	PDFUnsandboxed    string // allow | deny (the pdf viewer frame is never sandboxed)
	PDFBlobSrc        string // allow | deny (blob: sources for the pdf viewer)
}

// DefaultSecurity returns the strict, recommended policy.
func DefaultSecurity() SecurityConfig {
	return SecurityConfig{
		StyleDynamic:      "safe",
		ClassDynamic:      "filter",
		URLDynamic:        "safe",
		StaticURLSchemes:  "error",
		MaxRows:           100000,
		MaxCols:           10000,
		MaxText:           20000000,
		CSVFormulaEscape:  false,
		IframeSandbox:     "default",
		IframeReferrer:    "no-referrer",
		IframeScriptsSame: "error",
		IframeDynamicSrc:  "allowlist",
		PDFDynamicSrc:     "allowlist",
		PDFUnsandboxed:    "allow",
		PDFBlobSrc:        "allow",
	}
}

// Normalized replaces empty/invalid values with strict defaults. The config
// loader rejects malformed files/values before this point; this method is the
// final safety net for programs built without ParseSource/LoadProgram.
func (s SecurityConfig) Normalized() SecurityConfig {
	d := DefaultSecurity()
	switch s.StyleDynamic {
	case "safe", "warn", "allow":
	default:
		s.StyleDynamic = d.StyleDynamic
	}
	switch s.ClassDynamic {
	case "filter", "allow":
	default:
		s.ClassDynamic = d.ClassDynamic
	}
	switch s.URLDynamic {
	case "safe", "warn", "allow":
	default:
		s.URLDynamic = d.URLDynamic
	}
	switch s.StaticURLSchemes {
	case "error", "warn", "allow":
	default:
		s.StaticURLSchemes = d.StaticURLSchemes
	}
	if s.MaxRows <= 0 {
		s.MaxRows = d.MaxRows
	}
	if s.MaxCols <= 0 {
		s.MaxCols = d.MaxCols
	}
	if s.MaxText <= 0 {
		s.MaxText = d.MaxText
	}
	switch s.IframeSandbox {
	case "default", "require", "allow":
	default:
		s.IframeSandbox = d.IframeSandbox
	}
	switch s.IframeReferrer {
	case "no-referrer", "no-referrer-when-downgrade", "origin", "origin-when-cross-origin",
		"same-origin", "strict-origin", "strict-origin-when-cross-origin", "unsafe-url", "unset":
	default:
		s.IframeReferrer = d.IframeReferrer
	}
	switch s.IframeScriptsSame {
	case "error", "warn", "allow":
	default:
		s.IframeScriptsSame = d.IframeScriptsSame
	}
	switch s.IframeDynamicSrc {
	case "allowlist", "sandbox", "allow":
	default:
		s.IframeDynamicSrc = d.IframeDynamicSrc
	}
	switch s.PDFDynamicSrc {
	case "allowlist", "allow":
	default:
		s.PDFDynamicSrc = d.PDFDynamicSrc
	}
	switch s.PDFUnsandboxed {
	case "allow", "deny":
	default:
		s.PDFUnsandboxed = d.PDFUnsandboxed
	}
	switch s.PDFBlobSrc {
	case "allow", "deny":
	default:
		s.PDFBlobSrc = d.PDFBlobSrc
	}
	return s
}

// securityFile mirrors the JSON schema. Unknown fields are rejected: a typo in
// a security policy must never be silently ignored.
type securityFile struct {
	Style struct {
		Dynamic string `json:"dynamic"`
	} `json:"style"`
	Class struct {
		Dynamic string `json:"dynamic"`
	} `json:"class"`
	URLs struct {
		Dynamic             string `json:"dynamic"`
		StaticActiveSchemes string `json:"staticActiveSchemes"`
	} `json:"urls"`
	Table struct {
		MaxRows int `json:"maxRows"`
		MaxCols int `json:"maxCols"`
		MaxText int `json:"maxText"`
	} `json:"table"`
	CSV struct {
		FormulaEscapeByDefault bool `json:"formulaEscapeByDefault"`
	} `json:"csv"`
	Iframe struct {
		Sandbox           string `json:"sandbox"`
		ReferrerPolicy    string `json:"referrerPolicy"`
		ScriptsSameOrigin string `json:"scriptsSameOrigin"`
		DynamicSrc        string `json:"dynamicSrc"`
	} `json:"iframe"`
	PDF struct {
		DynamicSrc       string `json:"dynamicSrc"`
		UnsandboxedFrame string `json:"unsandboxedFrame"`
		BlobSrc          string `json:"blobSrc"`
	} `json:"pdf"`
}

// LoadSecurityFile reads a JSON security policy. Missing file is an error when
// the path was explicit; auto-discovery handles missing files separately.
func LoadSecurityFile(path string) (SecurityConfig, []string, error) {
	cfg := DefaultSecurity()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil, err
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return cfg, nil, fmt.Errorf("invalid JSON: the security config must be an object")
	}
	var raw securityFile
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		return cfg, nil, fmt.Errorf("invalid JSON: %v", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return cfg, nil, fmt.Errorf("invalid JSON: trailing data")
	}
	var problems []string
	if raw.Style.Dynamic != "" {
		cfg.StyleDynamic = pickMode(raw.Style.Dynamic, cfg.StyleDynamic,
			[]string{"safe", "warn", "allow"}, "style.dynamic", &problems)
	}
	if raw.Class.Dynamic != "" {
		cfg.ClassDynamic = pickMode(raw.Class.Dynamic, cfg.ClassDynamic,
			[]string{"filter", "allow"}, "class.dynamic", &problems)
	}
	if raw.URLs.Dynamic != "" {
		cfg.URLDynamic = pickMode(raw.URLs.Dynamic, cfg.URLDynamic,
			[]string{"safe", "warn", "allow"}, "urls.dynamic", &problems)
	}
	if raw.URLs.StaticActiveSchemes != "" {
		cfg.StaticURLSchemes = pickMode(raw.URLs.StaticActiveSchemes, cfg.StaticURLSchemes,
			[]string{"error", "warn", "allow"}, "urls.staticActiveSchemes", &problems)
	}
	if raw.Table.MaxRows != 0 {
		cfg.MaxRows = pickPositive(raw.Table.MaxRows, cfg.MaxRows, "table.maxRows", &problems)
	}
	if raw.Table.MaxCols != 0 {
		cfg.MaxCols = pickPositive(raw.Table.MaxCols, cfg.MaxCols, "table.maxCols", &problems)
	}
	if raw.Table.MaxText != 0 {
		cfg.MaxText = pickPositive(raw.Table.MaxText, cfg.MaxText, "table.maxText", &problems)
	}
	cfg.CSVFormulaEscape = raw.CSV.FormulaEscapeByDefault
	if raw.Iframe.Sandbox != "" {
		cfg.IframeSandbox = pickMode(raw.Iframe.Sandbox, cfg.IframeSandbox,
			[]string{"default", "require", "allow"}, "iframe.sandbox", &problems)
	}
	if raw.Iframe.ReferrerPolicy != "" {
		cfg.IframeReferrer = pickMode(raw.Iframe.ReferrerPolicy, cfg.IframeReferrer,
			[]string{"no-referrer", "no-referrer-when-downgrade", "origin", "origin-when-cross-origin",
				"same-origin", "strict-origin", "strict-origin-when-cross-origin", "unsafe-url", "unset"},
			"iframe.referrerPolicy", &problems)
	}
	if raw.Iframe.ScriptsSameOrigin != "" {
		cfg.IframeScriptsSame = pickMode(raw.Iframe.ScriptsSameOrigin, cfg.IframeScriptsSame,
			[]string{"error", "warn", "allow"}, "iframe.scriptsSameOrigin", &problems)
	}
	if raw.Iframe.DynamicSrc != "" {
		cfg.IframeDynamicSrc = pickMode(raw.Iframe.DynamicSrc, cfg.IframeDynamicSrc,
			[]string{"allowlist", "sandbox", "allow"}, "iframe.dynamicSrc", &problems)
	}
	if raw.PDF.DynamicSrc != "" {
		cfg.PDFDynamicSrc = pickMode(raw.PDF.DynamicSrc, cfg.PDFDynamicSrc,
			[]string{"allowlist", "allow"}, "pdf.dynamicSrc", &problems)
	}
	if raw.PDF.UnsandboxedFrame != "" {
		cfg.PDFUnsandboxed = pickMode(raw.PDF.UnsandboxedFrame, cfg.PDFUnsandboxed,
			[]string{"allow", "deny"}, "pdf.unsandboxedFrame", &problems)
	}
	if raw.PDF.BlobSrc != "" {
		cfg.PDFBlobSrc = pickMode(raw.PDF.BlobSrc, cfg.PDFBlobSrc,
			[]string{"allow", "deny"}, "pdf.blobSrc", &problems)
	}
	if len(problems) > 0 {
		return cfg.Normalized(), nil, fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return cfg.Normalized(), nil, nil
}

func pickMode(v, def string, allowed []string, name string, problems *[]string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	*problems = append(*problems, fmt.Sprintf("%s: invalid value %q", name, v))
	return def
}

func pickPositive(v, def int, name string, problems *[]string) int {
	if v > 0 {
		return v
	}
	*problems = append(*problems, fmt.Sprintf("%s: must be positive", name))
	return def
}

// SecurityPathFor returns the conventional config path for a project dir.
func SecurityPathFor(dir string) string {
	return filepath.Join(dir, "simpleui.security.json")
}

// SecurityForEntry resolves the policy for an entry .sui file:
// explicit override > simpleui.security.json in the entry directory > defaults.
func SecurityForEntry(entryFile, override string) (SecurityConfig, []string, error) {
	if override != "" {
		return LoadSecurityFile(override)
	}
	p := SecurityPathFor(dirOf(entryFile))
	if _, err := os.Stat(p); err == nil {
		return LoadSecurityFile(p)
	} else if !os.IsNotExist(err) {
		return DefaultSecurity(), nil, err
	}
	return DefaultSecurity(), nil, nil
}

// ParseSource uses this default; CmdGui loads the project config once at
// startup. LoadProgram overrides it with the entry directory's policy.
var defaultSecurityPolicy = DefaultSecurity()

// SetDefaultSecurity changes the policy used by ParseSource (designer, tests).
func SetDefaultSecurity(c SecurityConfig) {
	defaultSecurityPolicy = c.Normalized()
}

// DefaultSecurityPolicy returns the current ParseSource default (tests reset it).
func DefaultSecurityPolicy() SecurityConfig {
	return defaultSecurityPolicy
}

// securityPathOverride is the --security flag. LoadProgram applies it.
var securityPathOverride string

// SetSecurityPath sets/clears the explicit --security file for LoadProgram.
func SetSecurityPath(path string) {
	securityPathOverride = path
}

// SecurityPathOverride returns the current explicit path (tests).
func SecurityPathOverride() string {
	return securityPathOverride
}

// ApplyDefaultSecurityFile loads simpleui.security.json from dir (when present)
// into the ParseSource default. Used by the GUI/LSP/MCP entry points, which do
// not have a single entry .sui file.
func ApplyDefaultSecurityFile(dir string) ([]string, error) {
	p := SecurityPathFor(dir)
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	cfg, warns, err := LoadSecurityFile(p)
	if err != nil {
		return nil, err
	}
	SetDefaultSecurity(cfg)
	return warns, nil
}
