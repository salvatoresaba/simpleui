package core

import "fmt"
import "os"
import "strings"
import "sync"

// Diag is a single compiler diagnostic (error or warning). Everything a
// phase wants to report is accumulated in a DiagList instead of panicking,
// so the CLI can print all problems in one pass.
//
// Fields:
//
//	file, line, col: 1-based source position.
//	kind: phase category, one of "", "lexical", "syntax", "type", "io",
//	      "codegen". Empty falls back to the DiagList default.
//	msg:  the human-readable explanation, kept short and precise.
//	hint: optional follow-up note ("did you mean ...?", "declared here: ...").
//	      May be blank.
type Diag struct {
	file string
	line int
	col  int
	kind string
	msg  string
	hint string
}

type DiagList struct {
	list []Diag
	kind string // default kind stamped onto Add()-ed diagnostics
}

// Kind sets the default category used by Add for this list.
func (dl *DiagList) Kind(k string) {
	dl.kind = k
}

// HasErrors reports whether any diagnostics were accumulated.
// HasErrors reports whether any diagnostic is an error (as opposed to a
// warning). Warnings are informational and must not block compilation.
func (dl *DiagList) HasErrors() bool {
	for i := 0; i < len(dl.list); i++ {
		if dl.list[i].kind == "warning" {
			continue
		}
		return true
	}
	return false
}

// Count returns the number of accumulated diagnostics.
func (dl *DiagList) Count() int {
	return len(dl.list)
}

// WarningCount returns the number of warning-kind diagnostics.
func (dl *DiagList) WarningCount() int {
	n := 0
	for i := 0; i < len(dl.list); i++ {
		if dl.list[i].kind == "warning" {
			n++
		}
	}
	return n
}

// Add appends a diagnostic with the list's default kind and no hint.
// Positions are 1-based (line and col).
func (dl *DiagList) Add(file string, line, col int, msg string) {
	dl.AddEx(file, line, col, "", msg, "")
}

// AddEx appends a diagnostic with an explicit kind and an optional hint.
func (dl *DiagList) AddEx(file string, line, col int, kind string, msg string, hint string) {
	entry := Diag{file: file, line: line, col: col, kind: kind, msg: msg, hint: hint}
	if entry.kind == "" {
		entry.kind = dl.kind
	}
	dl.list = append(dl.list, entry)
}

// AddHint adds a hint line to the most recent diagnostic (a no-op when the
// list is empty). Useful for enriching an error discovered later.
func (dl *DiagList) AddHint(hint string) {
	if len(dl.list) == 0 {
		return
	}
	last := &dl.list[len(dl.list)-1]
	if last.hint == "" {
		last.hint = hint
	} else {
		last.hint = last.hint + "; " + hint
	}
}

// Extend appends every diagnostic from src, preserving its kind and hint.
func (dl *DiagList) Extend(src *DiagList) {
	if src == nil {
		return
	}
	dl.list = append(dl.list, src.list...)
}

// Print renders every diagnostic to standard error: a one-line summary with
// the phase category, a source snippet with a caret under the offending
// position (when the source is registered), the optional hint, and a final
// error count. When the source cannot be located the classic
// `file:line:col: message` line is printed instead, so the fallback is always
// informative.
func (dl *DiagList) Print() {
	if len(dl.list) == 0 {
		return
	}
	os.Stderr.WriteString(dl.String())
}

// String renders the whole diagnostic report (see Print). It never writes to
// a stream, which makes it usable from tests and embedding tools alike.
func (dl *DiagList) String() string {
	if len(dl.list) == 0 {
		return ""
	}
	var sb strings.Builder
	for i := 0; i < len(dl.list); i++ {
		renderDiag(&dl.list[i], &sb)
	}
	errs := 0
	warns := 0
	for i := 0; i < len(dl.list); i++ {
		if dl.list[i].kind == "warning" {
			warns++
		} else {
			errs++
		}
	}
	if errs == 0 {
		if warns == 1 {
			sb.WriteString("1 warning.\n")
		} else {
			fmt.Fprintf(&sb, "%v warnings.\n", warns)
		}
		return sb.String()
	}
	if errs == 1 {
		sb.WriteString("1 error")
	} else {
		fmt.Fprintf(&sb, "%v errors", errs)
	}
	if warns == 1 {
		sb.WriteString(" and 1 warning")
	} else if warns > 1 {
		fmt.Fprintf(&sb, " and %v warnings", warns)
	}
	sb.WriteString(". fix the problem")
	if errs > 1 {
		sb.WriteString("s")
	}
	sb.WriteString(" above and recompile.\n")
	return sb.String()
}

// FullString is like String but never falls back to the condensed form; it is
// exported for diagnostics tooling. This iteration only uses String.
func (dl *DiagList) FullString() string {
	return dl.String()
}

// labelFor turns a Diag kind into a human category label.
func labelFor(kind string) string {
	switch kind {
	case "lexical":
		return "lexical error"
	case "syntax":
		return "syntax error"
	case "type":
		return "type error"
	case "io":
		return "io error"
	case "codegen":
		return "codegen error"
	case "warning":
		return "warning"
	}
	return "error"
}

func renderDiag(d *Diag, sb *strings.Builder) {
	line, ok := sourceLineAt(d.file, d.line)
	if !ok {
		fmt.Fprintf(sb, "%v:%v:%v: %v\n", d.file, d.line, d.col, d.msg)
		return
	}
	sb.WriteString(labelFor(d.kind))
	sb.WriteString(": ")
	sb.WriteString(d.msg)
	sb.WriteString("\n")
	fmt.Fprintf(sb, "   --> %v:%v:%v\n", d.file, d.line, d.col)
	sb.WriteString("    |\n")
	num := fmt.Sprintf("%v", d.line)
	pad := strings.Repeat(" ", len(num))
	expanded := expandTabs(line)
	fmt.Fprintf(sb, "%v | %v\n", num, expanded)
	fmt.Fprintf(sb, "%v | %v\n", pad, caretAt(expanded, d.col))
	if d.hint != "" {
		fmt.Fprintf(sb, "    = %v\n", d.hint)
	}
}

// caretAt builds a caret string pointing at 1-based column col, aligning
// with the display line (tabs already expanded). Tabs before the caret are
// counted as four columns to stay aligned with expandTabs.
func caretAt(line string, col int) string {
	width := 0
	i := 0
	for i < len(line) && i < col-1 {
		if line[i] == '\t' {
			width += 4
		} else {
			width++
		}
		i++
	}
	if width == 0 {
		return "^"
	}
	return strings.Repeat(" ", width) + "^"
}

func expandTabs(line string) string {
	if strings.IndexByte(line, '\t') < 0 {
		return line
	}
	var sb strings.Builder
	for i := 0; i < len(line); i++ {
		if line[i] == '\t' {
			sb.WriteString("    ")
		} else {
			sb.WriteByte(line[i])
		}
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// Source registry. Each diagnostic knows a file name and a line number; to
// render a snippet the CLI needs the actual source text. ParseSource
// registers every file it reads (RegisterSource), keyed by the same file
// name that appears in the tokens/spans, so lookups stay consistent.
// ---------------------------------------------------------------------------

var (
	sourceLinesMu sync.RWMutex
	sourceLines   = map[string][]string{}
)

// RegisterSource stores the source of a file for snippet rendering. The
// lines are split eagerly; line N lives at slice index N-1. The designer and
// preview handlers parse sources concurrently, so the registry is guarded.
func RegisterSource(file string, content string) {
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	sourceLinesMu.Lock()
	sourceLines[file] = lines
	sourceLinesMu.Unlock()
}

// sourceLineAt returns the raw (unexpanded) text of the 1-based line, or
// false when the file is not registered or the line is out of range.
func sourceLineAt(file string, line int) (string, bool) {
	sourceLinesMu.RLock()
	ls, ok := sourceLines[file]
	sourceLinesMu.RUnlock()
	if !ok {
		return "", false
	}
	if line < 1 || line > len(ls) {
		return "", false
	}
	return ls[line-1], true
}

// Fatal formats a message using fmt.Printf style %v/%q/etc. arguments and
// returns a non-nil error, suitable for reporting an unrecoverable phase
// failure (e.g. file not found). The resulting error message is plain text.
func Fatal(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
