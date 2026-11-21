package sui

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// sourcemap.go - SUI -> generated JavaScript/CSS source maps.
//
// The generators can prepend invisible markers such as
//
//	/*@sui-src:<base64-file>:<line>:<column>*/
//
// immediately before code derived from an AST node. SourceMapFromMarked
// removes the markers and turns them into a standard Source Map v3 document.
// Markers are emitted only when a source-map build is requested, so normal
// output is byte-for-byte unchanged.
//
// Browsers support source maps for JavaScript and CSS; HTML itself has no
// native source-map format. The SUI -> HTML path is served by the JavaScript
// map because the generated HTML view code is built by app.js.

const sourceMarkerPrefix = "/*@sui-src:"

// SourceMapOptions controls source names and embedded sourcesContent.
type SourceMapOptions struct {
	GeneratedFile  string
	SourceBase     string
	IncludeContent bool
}

type sourceMapping struct {
	genLine int
	genCol  int
	file    string
	srcLine int
	srcCol  int
}

// srcMarker returns the source-map marker for a span, or "" when markers are
// disabled or the span is unknown.
func srcMarker(span Span) string {
	if !genEmitSourceMarkers {
		return ""
	}
	return rawSrcMarker(span)
}

// rawSrcMarker builds a marker regardless of the JS generator flag; CSS
// generation controls markers with its own boolean.
func rawSrcMarker(span Span) string {
	if span.line <= 0 || span.file == "" {
		return ""
	}
	file := base64.RawURLEncoding.EncodeToString([]byte(span.file))
	return sourceMarkerPrefix + file + ":" + strconv.Itoa(span.line) + ":" + strconv.Itoa(span.col) + "*/"
}

// SourceMapFromMarked strips source markers from generated code and returns
// the clean code plus a Source Map v3 JSON document.
func SourceMapFromMarked(marked string, opts SourceMapOptions) (string, string) {
	var out strings.Builder
	out.Grow(len(marked))
	var mappings []sourceMapping
	posIndex := map[[2]int]int{}
	line, col := 0, 0
	for i := 0; i < len(marked); {
		if strings.HasPrefix(marked[i:], sourceMarkerPrefix) {
			endRel := strings.Index(marked[i:], "*/")
			if endRel >= 0 {
				body := marked[i+len(sourceMarkerPrefix) : i+endRel]
				if file, srcLine, srcCol, ok := parseSourceMarker(body); ok {
					key := [2]int{line, col}
					m := sourceMapping{genLine: line, genCol: col, file: file, srcLine: srcLine, srcCol: srcCol}
					if idx, exists := posIndex[key]; exists {
						mappings[idx] = m
					} else {
						posIndex[key] = len(mappings)
						mappings = append(mappings, m)
					}
					i += endRel + 2
					continue
				}
			}
		}
		c := marked[i]
		out.WriteByte(c)
		if c == '\n' {
			line++
			col = 0
		} else {
			col++
		}
		i++
	}
	clean := out.String()
	return clean, buildSourceMapJSON(clean, mappings, opts)
}

func parseSourceMarker(body string) (string, int, int, bool) {
	first := strings.IndexByte(body, ':')
	if first <= 0 {
		return "", 0, 0, false
	}
	second := strings.IndexByte(body[first+1:], ':')
	if second < 0 {
		return "", 0, 0, false
	}
	second += first + 1
	rawFile := body[:first]
	line, ok1 := parsePositiveInt(body[first+1 : second])
	col, ok2 := parsePositiveInt(body[second+1:])
	if !ok1 || !ok2 {
		return "", 0, 0, false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(rawFile)
	if err != nil {
		return "", 0, 0, false
	}
	return string(decoded), line, col, true
}

func parsePositiveInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
		if n > 1<<30 {
			return 0, false
		}
	}
	return n, true
}

type sourceMapJSON struct {
	Version        int           `json:"version"`
	File           string        `json:"file"`
	SourceRoot     string        `json:"sourceRoot,omitempty"`
	Sources        []string      `json:"sources"`
	SourcesContent []interface{} `json:"sourcesContent,omitempty"`
	Names          []string      `json:"names"`
	Mappings       string        `json:"mappings"`
}

func buildSourceMapJSON(clean string, mappings []sourceMapping, opts SourceMapOptions) string {
	baseAbs := ""
	if opts.SourceBase != "" {
		if abs, err := filepath.Abs(opts.SourceBase); err == nil {
			baseAbs = abs
		}
	}
	sourceIndex := map[string]int{}
	var sources []string
	var originalPaths []string
	for i := 0; i < len(mappings); i++ {
		name := sourceMapName(mappings[i].file, baseAbs)
		if _, ok := sourceIndex[name]; !ok {
			sourceIndex[name] = len(sources)
			sources = append(sources, name)
			originalPaths = append(originalPaths, mappings[i].file)
		}
		mappings[i].file = name
	}
	sort.SliceStable(mappings, func(i, j int) bool {
		if mappings[i].genLine != mappings[j].genLine {
			return mappings[i].genLine < mappings[j].genLine
		}
		return mappings[i].genCol < mappings[j].genCol
	})
	var encoded strings.Builder
	prevSrc, prevOrigLine, prevOrigCol := 0, 0, 0
	prevGenCol := 0
	currentLine := 0
	firstInLine := true
	lastGenLine, lastGenCol := -1, -1
	for i := 0; i < len(mappings); i++ {
		m := mappings[i]
		if m.genLine == lastGenLine && m.genCol == lastGenCol {
			continue
		}
		for currentLine < m.genLine {
			encoded.WriteByte(';')
			currentLine++
			firstInLine = true
			prevGenCol = 0
		}
		if !firstInLine {
			encoded.WriteByte(',')
		}
		encoded.WriteString(encodeVLQ(m.genCol - prevGenCol))
		encoded.WriteString(encodeVLQ(sourceIndex[m.file] - prevSrc))
		encoded.WriteString(encodeVLQ(m.srcLine - 1 - prevOrigLine))
		encoded.WriteString(encodeVLQ(m.srcCol - 1 - prevOrigCol))
		prevGenCol = m.genCol
		prevSrc = sourceIndex[m.file]
		prevOrigLine = m.srcLine - 1
		prevOrigCol = m.srcCol - 1
		firstInLine = false
		lastGenLine, lastGenCol = m.genLine, m.genCol
	}
	_ = clean
	doc := sourceMapJSON{
		Version:  3,
		File:     opts.GeneratedFile,
		Sources:  sources,
		Names:    []string{},
		Mappings: encoded.String(),
	}
	if doc.Sources == nil {
		doc.Sources = []string{}
	}
	if opts.IncludeContent {
		content := make([]interface{}, len(sources))
		for i := 0; i < len(sources); i++ {
			path := originalPaths[i]
			if !filepath.IsAbs(path) {
				if baseAbs != "" {
					path = filepath.Join(baseAbs, path)
				} else if abs, err := filepath.Abs(path); err == nil {
					path = abs
				}
			}
			data, err := os.ReadFile(path)
			if err == nil {
				content[i] = string(data)
			} else {
				content[i] = nil
			}
		}
		doc.SourcesContent = content
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return `{"version":3,"file":"` + opts.GeneratedFile + `","sources":[],"names":[],"mappings":""}`
	}
	return string(data)
}

func sourceMapName(path, baseAbs string) string {
	if path == "" {
		return ""
	}
	abs := path
	if !filepath.IsAbs(abs) {
		if a, err := filepath.Abs(path); err == nil {
			abs = a
		}
	}
	if baseAbs != "" {
		if rel, err := filepath.Rel(baseAbs, abs); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(abs)
}

const vlqAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func encodeVLQ(value int) string {
	var out strings.Builder
	v := value << 1
	if value < 0 {
		v = ((-value) << 1) | 1
	}
	for {
		digit := v & 31
		v >>= 5
		if v > 0 {
			digit |= 32
		}
		out.WriteByte(vlqAlphabet[digit])
		if v == 0 {
			break
		}
	}
	return out.String()
}

// AppendInlineSourceMap adds an inline sourceMappingURL comment. kind is
// "js" or "css".
func AppendInlineSourceMap(code, mapJSON, kind string) string {
	data := base64.StdEncoding.EncodeToString([]byte(mapJSON))
	url := "data:application/json;charset=utf-8;base64," + data
	if kind == "css" {
		return code + "\n/*# sourceMappingURL=" + url + " */\n"
	}
	return code + "\n//# sourceMappingURL=" + url + "\n"
}
