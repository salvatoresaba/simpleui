package main

import (
	"bytes"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

// Hyphenation dictionaries are the largest single part of the rich text
// runtime (about 1.9 MB of TeX patterns). A .sui file can restrict them with
// the `hyphenLangs` attribute on a richtext control; this file resolves the
// designer's language list, maps language tags to pattern sets and emits a
// runtime containing only the selected dictionaries.

const richTextRuntimeMarker = "/* SimpleUI rich text editor runtime"

const (
	hyphenPatternsStart = "  var HYPHEN_PATTERNS = "
	hyphenPatternsEnd   = ";  var hyphenationLanguage"
)

// hyphenLangOption is one embedded dictionary. Labels come from the Document
// language selector; dictionaries that the selector does not expose (for
// example Sanskrit) are still listed so every embedded pattern set can be
// selected individually.
type hyphenLangOption struct {
	Code     string   `json:"code"`
	Label    string   `json:"label"`
	Bytes    int      `json:"bytes"`
	Patterns []string `json:"patterns"`
}

var (
	richHyphenPatterns = parseHyphenPatterns()
	richHyphenOptions  = parseHyphenOptions()
)

// splitRichTextRuntimePatterns returns the runtime around the single-line
// HYPHEN_PATTERNS literal. Keeping the original suffix preserves the
// language variable declaration that shares that line.
func splitRichTextRuntimePatterns() (head, raw, tail string, ok bool) {
	i := strings.Index(suiRichTextRuntime, hyphenPatternsStart)
	if i < 0 {
		return "", "", "", false
	}
	start := i + len(hyphenPatternsStart)
	e := strings.Index(suiRichTextRuntime[start:], hyphenPatternsEnd)
	if e < 0 {
		return "", "", "", false
	}
	end := start + e
	return suiRichTextRuntime[:start], suiRichTextRuntime[start:end], suiRichTextRuntime[end:], true
}

func parseHyphenPatterns() map[string]string {
	_, raw, _, ok := splitRichTextRuntimePatterns()
	if !ok {
		return nil
	}
	var patterns map[string]string
	if err := json.Unmarshal([]byte(raw), &patterns); err != nil {
		return nil
	}
	return patterns
}

// parseHyphenOptions builds one designer entry per embedded dictionary. The
// labels come from the Document language selector in the runtime, so the
// picker uses the editor's own names. Pattern keys absent from that selector
// get a stable English fallback label; the keys themselves come from
// HYPHEN_PATTERNS, which is the exhaustive source of truth.
func parseHyphenOptions() []hyphenLangOption {
	selectorLabels := map[string]string{}
	const marker = "[['', 'Automatic'],"
	if i := strings.Index(suiRichTextRuntime, marker); i >= 0 {
		if j := strings.Index(suiRichTextRuntime[i:], "].forEach(function (o) {"); j >= 0 {
			chunk := suiRichTextRuntime[i : i+j+1]
			re := regexp.MustCompile(`\['([^']*)', '([^']*)'\]`)
			for _, m := range re.FindAllStringSubmatch(chunk, -1) {
				if m[1] == "" {
					continue // "Automatic" is the absence of a choice
				}
				codes := hyphenPatternCodes([]string{m[1]})
				if len(codes) == 1 && selectorLabels[codes[0]] == "" {
					selectorLabels[codes[0]] = m[2]
				}
			}
		}
	}
	fallbackLabels := map[string]string{
		"el-polyton": "Ελληνικά (polytonic)",
		"fur":        "Friulian",
		"hsb":        "Upper Sorbian",
		"kmr":        "Northern Kurdish",
		"pms":        "Piedmontese",
		"rm":         "Romansh",
		"sa":         "Sanskrit",
	}
	keys := make([]string, 0, len(richHyphenPatterns))
	for key := range richHyphenPatterns {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	opts := make([]hyphenLangOption, 0, len(keys))
	for _, key := range keys {
		label := selectorLabels[key]
		if label == "" {
			label = fallbackLabels[key]
		}
		if label == "" {
			label = key
		}
		opts = append(opts, hyphenLangOption{
			Code: key, Label: label, Bytes: len(richHyphenPatterns[key]), Patterns: []string{key},
		})
	}
	return opts
}

// hyphenPatternAliases mirrors the runtime's HYPHEN_LANG_ALIAS table. The
// compiler needs it because a language tag such as `el` or `de-CH` must keep
// the pattern set the editor will actually request at runtime.
var hyphenPatternAliases = map[string]string{
	"sr":      "sr-cyrl",
	"me":      "sr-cyrl",
	"sr-latn": "sh-latn",
	"bs":      "sh-latn",
	"sh":      "sh-latn",
	"el":      "el-monoton",
	"gr":      "el-monoton",
	"no":      "nn",
	"nb":      "nn",
	"mn":      "mn-cyrl",
	"de-ch":   "de",
	"de-at":   "de",
	"de-1901": "de",
}

// hyphenPatternCodes resolves a list of language tags to the embedded pattern
// keys they use. Unknown tags are dropped: their documents fall back to the
// runtime's Latin heuristic, exactly as they do with the full dictionary set.
func hyphenPatternCodes(codes []string) []string {
	seen := map[string]bool{}
	out := []string{}
	add := func(code string) {
		if code == "" || seen[code] {
			return
		}
		if _, ok := richHyphenPatterns[code]; !ok {
			return
		}
		seen[code] = true
		out = append(out, code)
	}
	for _, raw := range codes {
		code := strings.ToLower(strings.TrimSpace(raw))
		if code == "" || code == "none" || code == "all" || code == "*" {
			continue
		}
		if _, ok := richHyphenPatterns[code]; ok {
			add(code)
			continue
		}
		base := strings.Split(code, "-")[0]
		if _, ok := richHyphenPatterns[base]; ok {
			add(base)
			continue
		}
		if alias, ok := hyphenPatternAliases[code]; ok {
			add(alias)
			continue
		}
		if alias, ok := hyphenPatternAliases[base]; ok {
			add(alias)
		}
	}
	sort.Strings(out)
	return out
}

// richTextRuntimeSubset replaces the embedded pattern object with the selected
// dictionaries. An empty list produces an empty object (heuristic only); the
// caller keeps the full runtime when the attribute is absent.
func richTextRuntimeSubset(codes []string) string {
	head, _, tail, ok := splitRichTextRuntimePatterns()
	if !ok || richHyphenPatterns == nil {
		return suiRichTextRuntime
	}
	selected := map[string]string{}
	for _, p := range hyphenPatternCodes(codes) {
		selected[p] = richHyphenPatterns[p]
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(selected); err != nil {
		return suiRichTextRuntime
	}
	return head + strings.TrimSpace(buf.String()) + tail
}

// hyphenLanguagesJSON feeds the designer's language picker.
func hyphenLanguagesJSON() string {
	data, err := json.Marshal(richHyphenOptions)
	if err != nil {
		return "[]"
	}
	return string(data)
}
