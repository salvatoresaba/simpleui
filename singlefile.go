package main

import (
	"path/filepath"
	"strings"
)

// singlefile.go - self-contained HTML export.
//
// A single-file build inlines the generated CSS and JavaScript into the index
// document. It contains no external runtime/file references (apart from assets
// deliberately referenced by user CSS or data). `--minify` applies a
// conservative, semantics-preserving pass: CSS comments/whitespace and JS
// full-line comments/indentation. It is not a replacement for a general-purpose
// JS minifier.

// BuildSingleHTML compiles a program and returns one HTML document.
func BuildSingleHTML(prog *Prog, minify bool) string {
	return CombineSingleHTML(GenerateIndexHtml(prog), GenerateCss(prog), GenerateJs(prog), minify)
}

// SingleOutputPath maps --out to a single .html file. An explicit .html/.htm
// path is used as-is; anything else is treated as a directory and receives
// app.html.
func SingleOutputPath(out string) string {
	if out == "" {
		out = "dist"
	}
	low := strings.ToLower(out)
	if strings.HasSuffix(low, ".html") || strings.HasSuffix(low, ".htm") {
		return out
	}
	return filepath.Join(out, "app.html")
}

// CombineSingleHTML inlines the already generated index/CSS/JS. It is used by
// the CLI and by the designer (which adds uploaded-font @font-face rules to the
// CSS).
func CombineSingleHTML(index, css, js string, minify bool) string {
	if minify {
		// The generated index skeleton is pure markup; removing its newlines
		// does not touch text content/inline assets.
		index = strings.ReplaceAll(index, "\n", "")
		css = minifyCSS(css)
		js = minifyJS(js)
	}
	// A literal </style> or </script> inside inline data would terminate the
	// HTML element early. In CSS, \/ is an escaped slash; in a JS string or
	// regex, \/ also denotes '/', so semantics are preserved.
	css = escapeInlineEnd(css, "style")
	js = escapeInlineEnd(js, "script")
	index = strings.Replace(index,
		`<link rel="stylesheet" href="style.css">`,
		"<style>"+css+"</style>", 1)
	index = strings.Replace(index,
		`<script src="app.js"></script>`,
		"<script>"+js+"</script>", 1)
	return index
}

// escapeInlineEnd inserts a backslash before the slash of a case-insensitive
// closing tag inside inline CSS/JS, preserving the original case/content.
func escapeInlineEnd(s, tag string) string {
	needle := "</" + tag
	var b strings.Builder
	for i := 0; i < len(s); {
		if i+len(needle) <= len(s) && strings.EqualFold(s[i:i+len(needle)], needle) {
			b.WriteString("<\\")
			i++ // keep the original '/', tag and following text
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// minifyCSS removes comments and redundant whitespace while preserving quoted
// strings and functional whitespace (calc, selectors, media queries).
func minifyCSS(s string) string {
	var b strings.Builder
	pendingSpace := false
	lastByte := byte(0)
	lastPunct := func() bool {
		if b.Len() == 0 {
			return true
		}
		switch lastByte {
		case '{', '}', ';', ',':
			return true
		}
		return false
	}
	writeByte := func(c byte) {
		b.WriteByte(c)
		lastByte = c
	}
	for i := 0; i < len(s); {
		c := s[i]
		if c == '/' && i+1 < len(s) && s[i+1] == '*' {
			j := i + 2
			for j+1 < len(s) && !(s[j] == '*' && s[j+1] == '/') {
				j++
			}
			if j+1 < len(s) {
				j += 2
			} else {
				j = len(s)
			}
			pendingSpace = true
			i = j
			continue
		}
		if c == '\'' || c == '"' {
			if pendingSpace && b.Len() > 0 && !lastPunct() {
				writeByte(' ')
			}
			pendingSpace = false
			quote := c
			writeByte(c)
			i++
			for i < len(s) {
				writeByte(s[i])
				if s[i] == '\\' && i+1 < len(s) {
					i++
					writeByte(s[i])
					i++
					continue
				}
				if s[i] == quote {
					i++
					break
				}
				i++
			}
			continue
		}
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' {
			pendingSpace = true
			i++
			continue
		}
		if c == '{' || c == '}' || c == ';' || c == ',' {
			pendingSpace = false
			writeByte(c)
			i++
			continue
		}
		if pendingSpace && b.Len() > 0 && !lastPunct() {
			writeByte(' ')
		}
		pendingSpace = false
		writeByte(c)
		i++
	}
	return strings.TrimSpace(b.String())
}

// minifyJS is conservative: it removes full-line // comments (when not inside
// a string/template) and indentation/trailing whitespace outside strings and
// template literals. Newlines are preserved to keep ASI semantics intact.
func minifyJS(s string) string {
	var b strings.Builder
	inSingle, inDouble, inTemplate := false, false, false
	for _, line := range strings.Split(s, "\n") {
		startState := inSingle || inDouble || inTemplate
		if !startState {
			trimmed := strings.TrimLeft(line, " \t")
			if strings.HasPrefix(trimmed, "//") {
				// Keep the line break: it may participate in ASI.
				b.WriteByte('\n')
				continue
			}
			line = strings.TrimRight(trimmed, " \t")
			if line == "" {
				continue
			}
		}
		b.WriteString(line)
		b.WriteByte('\n')
		inSingle, inDouble, inTemplate = scanJSLine(line, inSingle, inDouble, inTemplate)
	}
	return b.String()
}

// scanJSLine updates string/template state for one physical line. Template
// interpolation is deliberately treated as template text: comments inside
// ${...} are preserved, which is safe and still compact enough.
func scanJSLine(line string, inSingle, inDouble, inTemplate bool) (bool, bool, bool) {
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case inTemplate:
			if c == '\\' {
				i++
				continue
			}
			if c == '`' {
				inTemplate = false
			}
		case inSingle:
			if c == '\\' {
				i++
				continue
			}
			if c == '\'' {
				inSingle = false
			}
		case inDouble:
			if c == '\\' {
				i++
				continue
			}
			if c == '"' {
				inDouble = false
			}
		default:
			if c == '\'' {
				inSingle = true
			} else if c == '"' {
				inDouble = true
			} else if c == '`' {
				inTemplate = true
			}
		}
	}
	return inSingle, inDouble, inTemplate
}
