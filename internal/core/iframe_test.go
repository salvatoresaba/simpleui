package core

import (
	"strings"
	"testing"
)

func mustIframeProgram(t *testing.T, src string) (*Prog, DiagList) {
	t.Helper()
	prog, dl := ParseSource(src, "iframe.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	return prog, Check(prog)
}

func TestIframeSecureDefaults(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	prog, dl := mustIframeProgram(t, "view main {\n\tiframe src: \"https://example.com/a\"\n}\n")
	if dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, " sandbox=\"\"") {
		t.Error("missing sandbox was not rendered as sandbox=\"\"")
	}
	if !strings.Contains(js, " referrerpolicy=\"no-referrer\"") {
		t.Error("missing referrerpolicy was not rendered as no-referrer")
	}
	if !strings.Contains(js, "function __sui_iframe_src(") {
		t.Error("iframe runtime guard missing")
	}
}

func TestIframeRequireSandboxPolicy(t *testing.T) {
	withSecurityTestGlobals(t)
	cfg := DefaultSecurity()
	cfg.IframeSandbox = "require"
	SetDefaultSecurity(cfg)
	_, dl := mustIframeProgram(t, "view main {\n\tiframe src: \"https://example.com/a\"\n}\n")
	if !dl.HasErrors() || !strings.Contains(dl.String(), "sandbox is required") {
		t.Fatalf("expected missing sandbox error, got: %v", dl.String())
	}
}

func TestIframeSandboxValidation(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	cases := []string{
		"iframe src: \"https://example.com\" sandbox: \"allow-everything\"",
		"iframe src: \"https://example.com\" sandbox: sandboxTokens",
	}
	src := "state sandboxTokens = \"allow-scripts\"\nview main {\n"
	for _, c := range cases {
		src += "\t" + c + "\n"
	}
	src += "}\n"
	_, dl := mustIframeProgram(t, src)
	if !dl.HasErrors() {
		t.Fatal("malformed dynamic/unknown sandbox data was accepted")
	}
	if !strings.Contains(dl.String(), "cannot accept iframe sandbox data") ||
		!strings.Contains(dl.String(), "cannot accept dynamic sandbox data") {
		t.Fatalf("missing iframe sandbox diagnostics: %v", dl.String())
	}
}

func TestIframeScriptsSameOriginPolicy(t *testing.T) {
	withSecurityTestGlobals(t)
	src := "view main {\n\tiframe src: \"https://example.com\" sandbox: \"allow-scripts allow-same-origin\"\n}\n"
	SetDefaultSecurity(DefaultSecurity())
	_, dl := mustIframeProgram(t, src)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "can escape the sandbox") {
		t.Fatalf("dangerous sandbox combination was accepted: %v", dl.String())
	}

	cfg := DefaultSecurity()
	cfg.IframeScriptsSame = "warn"
	SetDefaultSecurity(cfg)
	_, dl = mustIframeProgram(t, src)
	if dl.HasErrors() || dl.WarningCount() == 0 {
		t.Fatalf("warn mode should compile with a warning: %v (%d warnings)", dl.String(), dl.WarningCount())
	}

	cfg.IframeScriptsSame = "allow"
	SetDefaultSecurity(cfg)
	_, dl = mustIframeProgram(t, src)
	if dl.HasErrors() || dl.WarningCount() != 0 {
		t.Fatalf("allow mode should compile silently: %v (%d warnings)", dl.String(), dl.WarningCount())
	}
}

func TestIframeMalformedSrcRejected(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	bad := []string{
		"javascript:alert(1)",
		"data:text/html,<script>alert(1)</script>",
		"mailto:user@example.com",
		"file:///etc/passwd",
		"about:blank",
		"https://user:pass@example.com/",
		"https://[",
		"https://example.com/a b",
		"https:\\\\example.com\\\\a",
		"",
	}
	for _, u := range bad {
		src := "view main {\n\tiframe src: \"" + strings.ReplaceAll(u, "\"", "\\\"") + "\"\n}\n"
		prog, _ := ParseSource(src, "iframe.sui")
		if prog == nil {
			continue
		}
		dl := Check(prog)
		if !dl.HasErrors() || !strings.Contains(dl.String(), "cannot accept iframe src data") {
			t.Errorf("malformed src %q was accepted: %v", u, dl.String())
		}
	}
}

func TestIframeDynamicSrcRequiresAllowList(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	src := "state u = \"https://example.com/a\"\nview main {\n\tiframe src: u\n}\n"
	_, dl := mustIframeProgram(t, src)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "add srcAllow") {
		t.Fatalf("dynamic src without allow-list was accepted: %v", dl.String())
	}

	src = "state u = \"https://example.com/a\"\nview main {\n\tiframe src: u srcAllow: \"https://example.com\"\n}\n"
	prog, dl := mustIframeProgram(t, src)
	if dl.HasErrors() {
		t.Fatalf("dynamic src with allow-list rejected: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"__sui_iframe_src(u, \"https://example.com\", true)",
		"__sui_iframe_error(__g[1])",
		"Cannot accept the data:",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated iframe guard missing %q", want)
		}
	}
}

func TestIframeInvalidSrcAllowRejected(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	bad := []string{"*", "javascript:alert(1)", "https://user:pass@example.com", "/../secret", "https://example.com/?q=1", ""}
	for _, allow := range bad {
		src := "state u = \"https://example.com\"\nview main {\n\tiframe src: u srcAllow: \"" + allow + "\"\n}\n"
		prog, _ := ParseSource(src, "iframe.sui")
		if prog == nil {
			continue
		}
		dl := Check(prog)
		if !dl.HasErrors() || !strings.Contains(dl.String(), "cannot accept") {
			t.Errorf("invalid srcAllow %q was accepted: %v", allow, dl.String())
		}
	}
}

func TestIframeInvalidAttributesRejected(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	attrs := []string{
		`allow: "camera; bad<"`,
		`allow: policyState`,
		`referrerpolicy: "sometimes"`,
		`referrerpolicy: rpState`,
		`loading: "fast"`,
		`width: "wide"`,
		`height: "-1"`,
	}
	for _, a := range attrs {
		src := "state policyState = \"camera\"\nstate rpState = \"no-referrer\"\nview main {\n\tiframe src: \"https://example.com\" " + a + "\n}\n"
		prog, _ := ParseSource(src, "iframe.sui")
		if prog == nil {
			continue
		}
		dl := Check(prog)
		if !dl.HasErrors() || !strings.Contains(dl.String(), "cannot accept") {
			t.Errorf("invalid iframe attribute %q was accepted: %v", a, dl.String())
		}
	}
}

func TestIframeDuplicateAttributeRejected(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	src := "view main {\n\tiframe src: \"https://example.com\" sandbox: \"\" sandbox: \"allow-scripts\"\n}\n"
	_, dl := mustIframeProgram(t, src)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "duplicate attribute") {
		t.Fatalf("duplicate iframe attribute was accepted: %v", dl.String())
	}
}

func TestIframeNumericDimensions(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	prog, dl := mustIframeProgram(t, "view main {\n\tiframe src: \"https://example.com\" width: 640 height: 360\n}\n")
	if dl.HasErrors() {
		t.Fatalf("literal integer dimensions rejected: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "width=\"${__sui_esc(640)}\"") ||
		!strings.Contains(js, "height=\"${__sui_esc(360)}\"") {
		t.Error("integer dimensions were not rendered")
	}
}

func TestIframeContentRejected(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	src := "view main {\n\tiframe \"fallback\" src: \"https://example.com\"\n}\n"
	_, dl := mustIframeProgram(t, src)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "cannot accept iframe content") {
		t.Fatalf("iframe fallback content was accepted: %v", dl.String())
	}
}

func TestGuiIframeAttributesRoundTrip(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	src := "state u = \"https://example.com/a\"\n" +
		"view main {\n\tiframe src: u srcAllow: \"https://example.com\" sandbox: \"allow-scripts\" allow: \"camera\" referrerpolicy: \"no-referrer\" loading: \"lazy\" width: \"640\" height: \"360\"\n}\n"
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen failed: %v", StringifyJson(msgs))
	}
	raw := StringifyJson(model)
	for _, want := range []string{
		`"srcExpr":true`,
		`"srcAllow":"https://example.com"`,
		`"sandbox":"allow-scripts"`,
		`"allow":"camera"`,
		`"referrerpolicy":"no-referrer"`,
		`"loading":"lazy"`,
		`"width":"640"`,
		`"height":"360"`,
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("designer model lost iframe attribute %s: %s", want, raw)
		}
	}
}

func TestIframePolicyModes(t *testing.T) {
	withSecurityTestGlobals(t)
	cfg := DefaultSecurity()
	cfg.IframeDynamicSrc = "sandbox"
	cfg.IframeReferrer = "unset"
	SetDefaultSecurity(cfg)
	src := "state u = \"https://example.com/a\"\nview main {\n\tiframe src: u\n}\n"
	prog, dl := mustIframeProgram(t, src)
	if dl.HasErrors() {
		t.Fatalf("dynamicSrc=sandbox should not require srcAllow: %v", dl.String())
	}
	js := GenerateJs(prog)
	if strings.Contains(js, " referrerpolicy=") {
		t.Error("referrerPolicy=unset still emitted a referrerpolicy")
	}
	if !strings.Contains(js, "__sui_iframe_src(u, '', false)") {
		t.Error("unguarded dynamic iframe src did not use the runtime validator")
	}
}
