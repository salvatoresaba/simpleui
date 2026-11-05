package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// iframe.go - compile-time validation for <iframe> attributes. Iframe data is
// accepted only when it is well formed and can be sandboxed safely. Dynamic
// URLs are validated again at runtime (__sui_iframe_src) and replaced with a
// visible error message when they cannot be accepted.

var iframeSandboxTokens = map[string]bool{
	"allow-downloads":                          true,
	"allow-forms":                              true,
	"allow-modals":                             true,
	"allow-orientation-lock":                   true,
	"allow-pointer-lock":                       true,
	"allow-popups":                             true,
	"allow-popups-to-escape-sandbox":           true,
	"allow-presentation":                       true,
	"allow-same-origin":                        true,
	"allow-scripts":                            true,
	"allow-storage-access-by-user-activation":  true,
	"allow-top-navigation":                     true,
	"allow-top-navigation-by-user-activation":  true,
	"allow-top-navigation-to-custom-protocols": true,
}

var iframeReferrerPolicies = map[string]bool{
	"no-referrer":                     true,
	"no-referrer-when-downgrade":      true,
	"origin":                          true,
	"origin-when-cross-origin":        true,
	"same-origin":                     true,
	"strict-origin":                   true,
	"strict-origin-when-cross-origin": true,
	"unsafe-url":                      true,
}

var policyFeatureName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func hasDynamicParts(parts []StrPart) bool {
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "expr" {
			return true
		}
	}
	return false
}

func attrIsDynamic(a *Attr) bool {
	if a == nil {
		return false
	}
	return a.hasExpr || hasDynamicParts(a.text)
}

// validateIframeSandbox checks the literal sandbox token list.
func validateIframeSandbox(v string) ([]string, error) {
	fields := strings.Fields(v)
	for i := 0; i < len(fields); i++ {
		tok := strings.ToLower(fields[i])
		if !iframeSandboxTokens[tok] {
			return nil, fmt.Errorf("unknown sandbox token %q", fields[i])
		}
		fields[i] = tok
	}
	return fields, nil
}

// validateIframeURL accepts only http(s) and relative URLs. It rejects
// credentials, control characters, backslashes and oversized input.
func validateIframeURL(v string) error {
	s := strings.TrimSpace(v)
	if s == "" {
		return fmt.Errorf("empty URL")
	}
	if len(s) > 4096 {
		return fmt.Errorf("URL is too long (max 4096 characters)")
	}
	for _, r := range s {
		if r <= 0x20 || r == 0x7f {
			return fmt.Errorf("URL contains whitespace or control characters")
		}
	}
	if strings.Contains(s, "\\") {
		return fmt.Errorf("URL contains a backslash")
	}
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("malformed URL: %v", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		return nil // relative or protocol-relative
	}
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("scheme %q is not allowed", scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	if u.User != nil {
		return fmt.Errorf("URL credentials are not allowed")
	}
	return nil
}

// validateIframeSrcAllowList validates the developer-written allow-list. Each
// entry is an absolute http(s) origin (optionally with a path prefix) or a
// same-origin path prefix beginning with "/".
func validateIframeSrcAllowList(v string) ([]string, error) {
	items := strings.FieldsFunc(v, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	if len(items) == 0 {
		return nil, fmt.Errorf("empty allow-list")
	}
	var out []string
	for _, raw := range items {
		a := strings.TrimSpace(raw)
		if a == "" {
			continue
		}
		for _, r := range a {
			if r < 0x21 || r == 0x7f || r == '\\' || strings.ContainsRune("<>\"'", r) {
				return nil, fmt.Errorf("invalid character in allow-list entry %q", a)
			}
		}
		if strings.Contains(a, "..") {
			return nil, fmt.Errorf("dot segments are not allowed in allow-list entry %q", a)
		}
		if strings.HasPrefix(a, "/") {
			if strings.HasPrefix(a, "//") || strings.Contains(a, "\\") ||
				strings.ContainsAny(a, "?#") {
				return nil, fmt.Errorf("invalid same-origin prefix %q", a)
			}
			out = append(out, a)
			continue
		}
		if a == "*" || strings.Contains(a, "*") {
			return nil, fmt.Errorf("wildcards are not allowed in an iframe allow-list")
		}
		u, err := url.Parse(a)
		if err != nil {
			return nil, fmt.Errorf("invalid origin %q", a)
		}
		scheme := strings.ToLower(u.Scheme)
		if scheme != "http" && scheme != "https" {
			return nil, fmt.Errorf("allow-list entry %q must use http(s)", a)
		}
		if u.Host == "" || u.User != nil {
			return nil, fmt.Errorf("invalid origin %q", a)
		}
		if u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("allow-list entry %q must not contain a query or fragment", a)
		}
		out = append(out, a)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty allow-list")
	}
	return out, nil
}

// validatePermissionsPolicy performs a conservative syntax check on the
// `allow` attribute: feature-name directives and optional origins.
func validatePermissionsPolicy(v string) error {
	if len(v) > 4096 {
		return fmt.Errorf("value is too long")
	}
	for _, r := range v {
		if r < 0x20 || r == 0x7f || r == '<' || r == '>' || r == '"' {
			return fmt.Errorf("invalid character")
		}
	}
	directives := strings.Split(v, ";")
	for _, d := range directives {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		fields := strings.Fields(d)
		if len(fields) == 0 || !policyFeatureName.MatchString(strings.ToLower(fields[0])) {
			return fmt.Errorf("invalid feature directive %q", d)
		}
		for _, origin := range fields[1:] {
			if origin == "*" || origin == "'self'" || origin == "'src'" || origin == "'none'" {
				continue
			}
			u, err := url.Parse(origin)
			if err != nil {
				return fmt.Errorf("invalid origin %q", origin)
			}
			scheme := strings.ToLower(u.Scheme)
			if (scheme != "http" && scheme != "https") || u.Host == "" || u.User != nil ||
				u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
				return fmt.Errorf("invalid origin %q", origin)
			}
		}
	}
	return nil
}

func validateDimension(name, v string) error {
	s := strings.TrimSpace(v)
	if s == "" {
		return fmt.Errorf("empty value")
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 100000 {
		return fmt.Errorf("must be an integer between 0 and 100000")
	}
	return nil
}

// checkIframe validates all iframe-specific data at compile time.
func (c *Checker) checkIframe(e *Element) {
	if len(e.children) > 0 || partsContent(e.text) != "" {
		c.diagAt(e.span, "cannot accept iframe content: use the src attribute; srcdoc/fallback markup is not allowed")
	}
	src := findAttr(e, "src")
	if src == nil {
		c.diagAt(e.span, "cannot accept iframe data: src is required")
		return
	}
	srcAllow := findAttr(e, "srcAllow")
	allowLit := ""
	if srcAllow != nil {
		if attrIsDynamic(srcAllow) {
			c.diagAt(srcAllow.span, "cannot accept dynamic srcAllow data: the allow-list must be a literal string")
		} else if lit, ok := staticPartsLiteral(srcAllow.text); ok {
			entries, err := validateIframeSrcAllowList(lit)
			if err != nil {
				c.diagAt(srcAllow.span, "cannot accept iframe srcAllow data: "+err.Error())
			} else {
				allowLit = strings.Join(entries, " ")
			}
		}
	}
	srcDynamic := attrIsDynamic(src)
	if srcDynamic {
		if allowLit == "" && c.security.IframeDynamicSrc == "allowlist" {
			c.diagAt(src.span, "cannot accept dynamic iframe src data: add srcAllow: \"https://host\" (literal allow-list)")
		}
	} else if lit, ok := staticPartsLiteral(src.text); ok {
		if err := validateIframeURL(lit); err != nil {
			c.diagAt(src.span, "cannot accept iframe src data: "+err.Error())
		}
	}

	sb := findAttr(e, "sandbox")
	if sb == nil {
		if c.security.IframeSandbox == "require" {
			c.diagAt(e.span, "cannot accept iframe data: sandbox is required by the security policy")
		}
	} else if attrIsDynamic(sb) {
		c.diagAt(sb.span, "cannot accept dynamic sandbox data: sandbox must be a literal token list")
	} else if lit, ok := staticPartsLiteral(sb.text); ok {
		tokens, err := validateIframeSandbox(lit)
		if err != nil {
			c.diagAt(sb.span, "cannot accept iframe sandbox data: "+err.Error())
		} else {
			hasScripts, hasSame := false, false
			for _, t := range tokens {
				if t == "allow-scripts" {
					hasScripts = true
				}
				if t == "allow-same-origin" {
					hasSame = true
				}
			}
			if hasScripts && hasSame {
				c.securityDiag(sb.span, c.security.IframeScriptsSame,
					"iframe sandbox: allow-scripts + allow-same-origin can escape the sandbox")
			}
		}
	}

	if a := findAttr(e, "allow"); a != nil {
		if attrIsDynamic(a) {
			c.diagAt(a.span, "cannot accept dynamic allow data: permissions must be a literal string")
		} else if lit, ok := staticPartsLiteral(a.text); ok {
			if err := validatePermissionsPolicy(lit); err != nil {
				c.diagAt(a.span, "cannot accept iframe allow data: "+err.Error())
			}
		}
	}
	if a := findAttr(e, "referrerpolicy"); a != nil {
		if attrIsDynamic(a) {
			c.diagAt(a.span, "cannot accept dynamic referrerpolicy data: use a literal value")
		} else if lit, ok := staticPartsLiteral(a.text); ok {
			if !iframeReferrerPolicies[strings.ToLower(strings.TrimSpace(lit))] {
				c.diagAt(a.span, "cannot accept iframe referrerpolicy data: invalid value "+strconv.Quote(lit))
			}
		}
	}
	if a := findAttr(e, "loading"); a != nil {
		if attrIsDynamic(a) {
			c.diagAt(a.span, "cannot accept dynamic loading data: use lazy or eager")
		} else if lit, ok := staticPartsLiteral(a.text); ok {
			v := strings.ToLower(strings.TrimSpace(lit))
			if v != "lazy" && v != "eager" {
				c.diagAt(a.span, "cannot accept iframe loading data: use lazy or eager")
			}
		}
	}
	for _, key := range []string{"width", "height"} {
		if a := findAttr(e, key); a != nil {
			if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
				if err := validateDimension(key, strconv.Itoa(a.expr.i)); err != nil {
					c.diagAt(a.span, "cannot accept iframe "+key+" data: "+err.Error())
				}
			} else if attrIsDynamic(a) {
				c.diagAt(a.span, "cannot accept dynamic "+key+" data: use a literal integer")
			} else if lit, ok := staticPartsLiteral(a.text); ok {
				if err := validateDimension(key, lit); err != nil {
					c.diagAt(a.span, "cannot accept iframe "+key+" data: "+err.Error())
				}
			}
		}
	}
}
