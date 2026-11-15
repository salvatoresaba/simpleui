package main

import (
	"fmt"
	"strings"
)

// css_external.go - reject external resource references in author CSS.
//
// Static CSS is trusted developer code, but it must not silently turn into a
// network request. The compiler therefore rejects @import and url(...) values
// that use a network scheme. Relative URLs, fragments and data: URIs remain
// available; dynamic CSS already goes through cssSafe, which strips url().
func cssExternalReferenceError(css string) string {
	if strings.TrimSpace(css) == "" {
		return ""
	}
	clean := cssUnescape(stripCssComments(css))
	if strings.Contains(strings.ToLower(maskCssStrings(clean)), "@import") {
		return "CSS @import is not allowed (external stylesheets are blocked)"
	}
	lower := strings.ToLower(clean)
	for i := 0; i < len(clean); i++ {
		if clean[i] == '"' || clean[i] == '\'' {
			i = skipCssString(clean, i)
			continue
		}
		if lower[i] != 'u' || i+3 >= len(clean) || lower[i:i+4] != "url(" {
			continue
		}
		if i > 0 && cssIdentByte(clean[i-1]) {
			continue
		}
		raw, end := cssFunctionArgument(clean, i+4)
		if end < 0 {
			continue
		}
		value := strings.TrimSpace(cssUnescape(raw))
		value = strings.Trim(value, "\"'")
		if cssURLIsExternal(value) {
			return fmt.Sprintf("CSS external URL %q is not allowed (use a data URI or a local relative asset)", value)
		}
		i = end
	}
	if msg := cssImageSetExternalError(clean); msg != "" {
		return msg
	}
	return ""
}

// cssImageSetExternalError covers image-set("https://...") notation, where the
// URL is a string rather than a url(...) token.
func cssImageSetExternalError(src string) string {
	lower := strings.ToLower(src)
	from := 0
	for {
		idx := strings.Index(lower[from:], "image-set(")
		if idx < 0 {
			return ""
		}
		start := from + idx + len("image-set(")
		end, ok := cssMatchingParen(src, start)
		if !ok {
			return ""
		}
		arg := src[start:end]
		for i := 0; i < len(arg); i++ {
			if arg[i] != '"' && arg[i] != '\'' {
				continue
			}
			close := skipCssString(arg, i)
			if close <= i {
				break
			}
			value := strings.TrimSpace(cssUnescape(arg[i+1 : close]))
			if cssURLIsExternal(value) {
				return fmt.Sprintf("CSS external URL %q is not allowed (use a data URI or a local relative asset)", value)
			}
			i = close
		}
		from = end + 1
	}
}

func cssMatchingParen(src string, start int) (int, bool) {
	depth := 0
	for i := start; i < len(src); i++ {
		if src[i] == '"' || src[i] == '\'' {
			i = skipCssString(src, i)
			continue
		}
		if src[i] == '(' {
			depth++
		} else if src[i] == ')' {
			if depth == 0 {
				return i, true
			}
			depth--
		}
	}
	return 0, false
}

func cssIdentByte(c byte) bool {
	return c == '-' || c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// skipCssString returns the index of the closing quote (or the last byte when
// the string is unterminated).
func skipCssString(src string, open int) int {
	quote := src[open]
	for i := open + 1; i < len(src); i++ {
		if src[i] == '\\' && i+1 < len(src) {
			i++
			continue
		}
		if src[i] == quote {
			return i
		}
	}
	return len(src) - 1
}

// maskCssStrings replaces the contents of quoted strings with spaces so an
// @import mentioned inside a content string is not treated as a rule.
func maskCssStrings(src string) string {
	var b strings.Builder
	for i := 0; i < len(src); i++ {
		if src[i] != '"' && src[i] != '\'' {
			b.WriteByte(src[i])
			continue
		}
		end := skipCssString(src, i)
		b.WriteByte(src[i])
		for j := i + 1; j < end; j++ {
			b.WriteByte(' ')
		}
		if end < len(src) {
			b.WriteByte(src[end])
		}
		i = end
	}
	return b.String()
}

func stripCssComments(src string) string {
	var b strings.Builder
	for i := 0; i < len(src); i++ {
		if src[i] == '/' && i+1 < len(src) && src[i+1] == '*' {
			i += 2
			for i+1 < len(src) && !(src[i] == '*' && src[i+1] == '/') {
				i++
			}
			if i+1 < len(src) {
				i++
			}
			b.WriteByte(' ')
			continue
		}
		b.WriteByte(src[i])
	}
	return b.String()
}

// cssFunctionArgument returns the raw argument of a function whose opening
// parenthesis was already consumed, plus the index of the closing parenthesis.
func cssFunctionArgument(src string, start int) (string, int) {
	i := start
	for i < len(src) && (src[i] == ' ' || src[i] == '\t' || src[i] == '\n' || src[i] == '\r') {
		i++
	}
	if i >= len(src) {
		return "", -1
	}
	if src[i] == '"' || src[i] == '\'' {
		quote := src[i]
		i++
		begin := i
		for i < len(src) {
			if src[i] == '\\' && i+1 < len(src) {
				i += 2
				continue
			}
			if src[i] == quote {
				raw := src[begin:i]
				i++
				for i < len(src) && src[i] != ')' {
					i++
				}
				if i < len(src) {
					return raw, i
				}
				return raw, -1
			}
			i++
		}
		return "", -1
	}
	begin := i
	depth := 0
	for i < len(src) {
		if src[i] == '(' {
			depth++
		} else if src[i] == ')' {
			if depth == 0 {
				return src[begin:i], i
			}
			depth--
		}
		i++
	}
	return src[begin:], -1
}

func cssUnescape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		i++
		if i >= len(s) {
			break
		}
		if s[i] == '\n' || s[i] == '\r' || s[i] == '\f' {
			continue
		}
		if isCSSHex(s[i]) {
			j := i
			for j < len(s) && j-i < 6 && isCSSHex(s[j]) {
				j++
			}
			n := 0
			for k := i; k < j; k++ {
				n = n*16 + cssHexValue(s[k])
			}
			if n > 0 && n <= 0x10FFFF {
				b.WriteRune(rune(n))
			}
			i = j - 1
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isCSSHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func cssHexValue(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	default:
		return int(c-'A') + 10
	}
}

func cssURLIsExternal(value string) bool {
	v := strings.TrimSpace(value)
	if v == "" || strings.HasPrefix(v, "#") || strings.HasPrefix(v, "data:") || strings.HasPrefix(v, "blob:") {
		return false
	}
	low := strings.ToLower(v)
	if strings.HasPrefix(low, "//") {
		return true
	}
	for _, scheme := range []string{"http:", "https:", "ftp:", "file:", "ws:", "wss:"} {
		if strings.HasPrefix(low, scheme) {
			return true
		}
	}
	return false
}
