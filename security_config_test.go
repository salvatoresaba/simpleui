package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withSecurityTestGlobals(t *testing.T) {
	t.Helper()
	oldCfg := DefaultSecurityPolicy()
	oldPath := SecurityPathOverride()
	t.Cleanup(func() {
		SetDefaultSecurity(oldCfg)
		SetSecurityPath(oldPath)
	})
}

func TestSecurityConfigDefaults(t *testing.T) {
	d := DefaultSecurity()
	if d.StyleDynamic != "safe" || d.ClassDynamic != "filter" || d.URLDynamic != "safe" || d.StaticURLSchemes != "error" {
		t.Fatalf("unexpected defaults: %#v", d)
	}
	if d.MaxRows != 100000 || d.MaxCols != 10000 || d.MaxText != 20000000 {
		t.Fatalf("unexpected table defaults: %#v", d)
	}
	if d.CSVFormulaEscape {
		t.Fatal("CSV formula escaping must be opt-in by default")
	}
	if d.IframeSandbox != "default" || d.IframeReferrer != "no-referrer" ||
		d.IframeScriptsSame != "error" || d.IframeDynamicSrc != "allowlist" {
		t.Fatalf("unexpected iframe defaults: %#v", d)
	}
}

func TestSecurityConfigLoad(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "simpleui.security.json")
	raw := `{
	  "style": {"dynamic": "warn"},
	  "class": {"dynamic": "allow"},
	  "urls": {"dynamic": "allow", "staticActiveSchemes": "warn"},
	  "table": {"maxRows": 7, "maxCols": 3, "maxText": 99},
	  "csv": {"formulaEscapeByDefault": true},
	  "iframe": {"sandbox": "require", "referrerPolicy": "strict-origin", "scriptsSameOrigin": "warn", "dynamicSrc": "sandbox"}
	}`
	if err := os.WriteFile(p, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, warns, err := LoadSecurityFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(warns) != 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
	if cfg.StyleDynamic != "warn" || cfg.ClassDynamic != "allow" || cfg.URLDynamic != "allow" ||
		cfg.StaticURLSchemes != "warn" || cfg.MaxRows != 7 || cfg.MaxCols != 3 ||
		cfg.MaxText != 99 || !cfg.CSVFormulaEscape ||
		cfg.IframeSandbox != "require" || cfg.IframeReferrer != "strict-origin" ||
		cfg.IframeScriptsSame != "warn" || cfg.IframeDynamicSrc != "sandbox" {
		t.Fatalf("loaded config = %#v", cfg)
	}
}

func TestSecurityConfigInvalidRejected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "simpleui.security.json")
	if err := os.WriteFile(p, []byte(`{"style":{"dynamic":"wat"},"table":{"maxRows":-2}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := LoadSecurityFile(p)
	if err == nil {
		t.Fatal("invalid security config must be rejected")
	}
	msg := err.Error()
	if !strings.Contains(msg, "style.dynamic") || !strings.Contains(msg, "table.maxRows") {
		t.Fatalf("error does not identify the malformed fields: %v", err)
	}
	if cfg.StyleDynamic != "safe" || cfg.MaxRows != 100000 {
		t.Fatalf("returned fallback config = %#v", cfg)
	}
}

func TestSecurityConfigUnknownFieldRejected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "simpleui.security.json")
	if err := os.WriteFile(p, []byte(`{"stile":{"dynamic":"allow"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadSecurityFile(p); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field must be rejected, got %v", err)
	}
}

func TestSecurityRelaxedStyle(t *testing.T) {
	withSecurityTestGlobals(t)
	src := "state c = \"red\"\nview main {\n\tdiv \"x\" style: c\n}\n"

	cfg := DefaultSecurity()
	cfg.StyleDynamic = "warn"
	SetDefaultSecurity(cfg)
	prog, dl := ParseSource(src, "style.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() || dl.WarningCount() == 0 {
		t.Fatalf("warn policy: errors=%v warnings=%d", dl.String(), dl.WarningCount())
	}

	cfg.StyleDynamic = "allow"
	SetDefaultSecurity(cfg)
	prog, dl = ParseSource(src, "style.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() || dl.WarningCount() != 0 {
		t.Fatalf("allow policy: errors=%v warnings=%d", dl.String(), dl.WarningCount())
	}
}

func TestSecurityRelaxedURL(t *testing.T) {
	withSecurityTestGlobals(t)
	src := "import { http } from \"std\"\n" +
		"state u = \"https://x\"\n" +
		"async fn f() -> null {\n\tb, e = http.get(u)\n}\n" +
		"view main {\n\tp \"x\"\n}\n"

	cfg := DefaultSecurity()
	cfg.URLDynamic = "warn"
	SetDefaultSecurity(cfg)
	prog, dl := ParseSource(src, "url.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() || dl.WarningCount() == 0 {
		t.Fatalf("warn policy: errors=%v warnings=%d", dl.String(), dl.WarningCount())
	}

	cfg.URLDynamic = "allow"
	SetDefaultSecurity(cfg)
	prog, dl = ParseSource(src, "url.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() || dl.WarningCount() != 0 {
		t.Fatalf("allow policy: errors=%v warnings=%d", dl.String(), dl.WarningCount())
	}
}

func TestSecurityRelaxedStaticURL(t *testing.T) {
	withSecurityTestGlobals(t)
	src := "view main {\n\ta \"x\" href: \"javascript:alert(1)\"\n}\n"

	cfg := DefaultSecurity()
	cfg.StaticURLSchemes = "warn"
	SetDefaultSecurity(cfg)
	prog, dl := ParseSource(src, "url.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() || dl.WarningCount() == 0 {
		t.Fatalf("warn policy: errors=%v warnings=%d", dl.String(), dl.WarningCount())
	}

	cfg.StaticURLSchemes = "allow"
	SetDefaultSecurity(cfg)
	prog, dl = ParseSource(src, "url.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() || dl.WarningCount() != 0 {
		t.Fatalf("allow policy: errors=%v warnings=%d", dl.String(), dl.WarningCount())
	}
}

func TestSecurityAllowClass(t *testing.T) {
	withSecurityTestGlobals(t)
	cfg := DefaultSecurity()
	cfg.ClassDynamic = "allow"
	SetDefaultSecurity(cfg)
	src := "state c = \"x\"\nview main {\n\tdiv \"x\" class: c\n}\n"
	prog, dl := ParseSource(src, "class.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatal(dl.String())
	}
	js := GenerateJs(prog)
	if strings.Contains(js, "class=\"${__sui_class(c)}\"") {
		t.Error("class filtering still active under allow policy")
	}
	if !strings.Contains(js, "class=\"${__sui_esc(c)}\"") {
		t.Error("escaped class expression missing under allow policy")
	}
}

func TestSecurityTableLimits(t *testing.T) {
	withSecurityTestGlobals(t)
	cfg := DefaultSecurity()
	cfg.MaxRows = 7
	cfg.MaxCols = 3
	cfg.MaxText = 99
	cfg.CSVFormulaEscape = true
	SetDefaultSecurity(cfg)
	prog, dl := ParseSource("view main {\n\tp \"x\"\n}\n", "limits.sui")
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "const __SUI_SEC = {maxRows:7,maxCols:3,maxText:99,csvFormulaEscape:true};") {
		t.Error("generated __SUI_SEC does not reflect the config")
	}
	if !strings.Contains(js, "__SUI_SEC && __SUI_SEC.maxRows") {
		t.Error("table runtime does not read the config")
	}
}

func TestSecurityAutoDiscovery(t *testing.T) {
	withSecurityTestGlobals(t)
	SetSecurityPath("")
	dir := t.TempDir()
	entry := filepath.Join(dir, "app.sui")
	src := "state c = \"red\"\nview main {\n\tdiv \"x\" style: c\n}\n"
	if err := os.WriteFile(entry, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "simpleui.security.json"),
		[]byte(`{"style":{"dynamic":"allow"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, dl := LoadProgram(entry)
	if dl.HasErrors() {
		t.Fatal(dl.String())
	}
	if prog.security.StyleDynamic != "allow" {
		t.Fatalf("auto-discovered policy = %#v", prog.security)
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("relaxed style should compile: %v", dl.String())
	}
}

func TestSecurityInvalidValueBlocksLoadProgram(t *testing.T) {
	withSecurityTestGlobals(t)
	SetSecurityPath("")
	dir := t.TempDir()
	entry := filepath.Join(dir, "app.sui")
	if err := os.WriteFile(entry, []byte("view main {\n\tp \"x\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "simpleui.security.json"),
		[]byte(`{"iframe":{"sandbox":"nope"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, dl := LoadProgram(entry)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "iframe.sandbox") {
		t.Fatalf("invalid security value was accepted: %v", dl.String())
	}
}

func TestSecurityMalformedConfig(t *testing.T) {
	withSecurityTestGlobals(t)
	SetSecurityPath("")
	dir := t.TempDir()
	entry := filepath.Join(dir, "app.sui")
	if err := os.WriteFile(entry, []byte("view main {\n\tp \"x\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "simpleui.security.json"), []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, dl := LoadProgram(entry)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "security config") {
		t.Fatalf("expected a security config error, got: %v", dl.String())
	}
}
