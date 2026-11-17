package main

import "fmt"
import "os"
import "strings"

// formatter.go - `simpleui fmt`: a canonical token-level formatter.
//
// It re-emits the token stream with a consistent style: 2-space indent per
// block, one statement per line, normalized operator/punctuation spacing, and
// at most one blank line between top-level declarations. Comments (`#...`)
// are parsed separately and re-attached (standalone comments stay on their
// own line at the current indent before the following item). Because it works
// on tokens, formatting never changes semantics - it only normalizes layout.

type FmtComment struct {
	line int
	col  int
	text string
}

// collectComments returns the standalone comments (with position) found in
// the source. Used by the token re-emitter so formatting does not lose them.
func collectComments(source string) []FmtComment {
	var out []FmtComment
	line := 1
	col := 1
	i := 0
	n := len(source)
	for i < n {
		c := source[i]
		if c == '\n' {
			line++
			col = 1
			i++
			continue
		}
		if c == '"' {
			// Skip string literals: a # inside a CSS string, color literal or
			// URL is not a comment.
			i++
			col++
			for i < n {
				if source[i] == '\\' {
					i += 2
					col += 2
					continue
				}
				if source[i] == '"' {
					i++
					col++
					break
				}
				if source[i] == '\n' {
					i++
					line++
					col = 1
					continue
				}
				i++
				col++
			}
			continue
		}
		if c == '#' {
			start := i
			startLine := line
			startCol := col
			for i < n && source[i] != '\n' {
				i++
				col++
			}
			out = append(out, FmtComment{line: startLine, col: startCol, text: source[start:i]})
			continue
		}
		col++
		i++
	}
	return out
}

// FormatSource returns the canonical form of `source`, or ("", errors) when
// the source does not parse.
func FormatSource(source string, file string) (string, *DiagList) {
	_, dl := ParseSource(source, file)
	if dl.HasErrors() {
		return "", &dl
	}
	toks, dlx := Lex(source, file)
	if dlx.HasErrors() {
		return "", &dlx
	}
	comments := collectComments(source)
	f := &fmtState{comments: comments, commentIdx: 0, curLine: 1}
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t.kind == "newline" {
			if f.wantBreak {
				// This newline fulfils the requested break after "{".
				f.wantBreak = false
				f.pendingNL = 1
			} else {
				f.pendingNL++
			}
			continue
		}
		if t.kind == "eof" {
			break
		}
		// Break after an opening brace so one-line blocks are expanded. If
		// the very next token closes the block ("{}" or "{ }"), keep it
		// inline and skip the break.
		if f.wantBreak {
			f.wantBreak = false
			if t.kind != "}" && f.pendingNL == 0 {
				f.pendingNL = 1
			}
		}
		// A closing brace that still shares its line with preceding content
		// starts a new line, so multi-line blocks close cleanly. An empty
		// block ("{}") is left untouched.
		if t.kind == "}" && f.pendingNL == 0 && !f.atLineStart && f.prevKind != "{" && f.prevKind != "" {
			f.pendingNL = 1
		}
		// Standalone comments that precede this token on earlier lines.
		for f.commentIdx < len(f.comments) && f.comments[f.commentIdx].line < t.line {
			f.flushNewlines(1)
			f.emitIndent()
			f.sb.WriteString(f.comments[f.commentIdx].text)
			f.sb.WriteString("\n")
			f.atLineStart = true
			f.curLine++
			f.commentIdx++
		}
		if f.pendingNL > 0 || t.line > f.curLine {
			// A newline token already accounts for the physical line break,
			// so only fall back to the line delta when no newline token was
			// produced (e.g. the very first token). Counting both would
			// double the break and insert spurious blank lines.
			if f.pendingNL == 0 && t.line > f.curLine {
				f.pendingNL += t.line - f.curLine
			}
			// Collapse blank runs: at most one blank line between items.
			f.flushNewlines(1)
		}
		if f.atLineStart {
			// A closing brace that begins a line belongs to the outer block,
			// so dedent before writing its indentation. Mark it so emitToken
			// does not decrement a second time.
			if t.kind == "}" && f.depth > 0 {
				f.depth--
				f.preDedented = true
			}
			f.emitIndent()
			// A new line starts fresh: no spacing decision may leak from the
			// last token of the previous line (which would add a stray
			// leading space after the indent).
			f.prevKind = ""
			f.prevText = ""
			f.preTight = false
		}
		f.emitToken(t)
		f.curLine = t.line
	}
	// Trailing comments.
	for f.commentIdx < len(f.comments) {
		f.flushNewlines(1)
		f.emitIndent()
		f.sb.WriteString(f.comments[f.commentIdx].text)
		f.sb.WriteString("\n")
		f.curLine++
		f.commentIdx++
	}
	f.sb.WriteString("\n")
	return f.sb.String(), nil
}

type fmtState struct {
	sb          strings.Builder
	comments    []FmtComment
	commentIdx  int
	prevKind    string
	prevText    string
	depth       int
	atLineStart bool
	curLine     int
	pendingNL   int
	preTight    bool
	wantBreak   bool
	preDedented bool
	declName    bool // previous token was fn/view and next ident is its name
	generic     int  // >0 while inside <T, U> after a declaration name
}

func (f *fmtState) emitIndent() {
	for i := 0; i < f.depth; i++ {
		f.sb.WriteString("  ")
	}
	f.atLineStart = false
}

func (f *fmtState) flushNewlines(max int) {
	n := f.pendingNL
	if n > max+1 {
		n = max + 1
	}
	for i := 0; i < n; i++ {
		f.sb.WriteString("\n")
		f.curLine++
	}
	f.pendingNL = 0
	f.atLineStart = true
}

func isOp(kind string) bool {
	if kind == "=" || kind == "==" || kind == "!=" || kind == "<" || kind == "<=" || kind == ">" ||
		kind == ">=" || kind == "->" || kind == "+" || kind == "-" || kind == "*" || kind == "/" || kind == "%" ||
		kind == "&&" {
		return true
	}
	return false
}

func unaryContext(kind string) bool {
	if kind == "" {
		return true
	}
	if kind == "(" || kind == "[" || kind == "," {
		return true
	}
	return isOp(kind)
}

// noSpaceAfter reports whether a token of kind `prev` must not be followed by
// a space (open parentheses/brackets, dots, or tokens that already emitted a
// trailing space like "," and ":").
func noSpaceAfter(prev string) bool {
	if prev == "" || prev == "(" || prev == "[" || prev == "." || prev == ".." || prev == "..." || prev == "," || prev == ":" || prev == "{" {
		return true
	}
	return false
}

func (f *fmtState) emitToken(t Token) {
	kind := t.kind
	text := t.text
	if kind == "newline" || kind == "eof" {
		return
	}
	unary := false
	if kind == "-" || kind == "+" {
		unary = unaryContext(f.prevKind)
	}
	openBreak := false
	startGeneric := kind == "<" && f.declName && f.generic == 0
	endGeneric := kind == ">" && f.generic == 1
	if startGeneric {
		f.generic = 1
		f.declName = false
	}
	if endGeneric {
		f.generic = 0
	}
	// Decide whether a single space is needed BEFORE this token.
	switch kind {
	case "<", ">":
		if f.generic > 0 || startGeneric || endGeneric {
			// No space inside a declaration's <T, U> list.
		} else if !noSpaceAfter(f.prevKind) {
			f.space()
		}
	case ")":
		// no space before
	case "]":
		// no space before
	case "[":
		// No space before an index, but keep a space before an array literal
		// after let/=/return/operator (let[first] -> let [first]).
		if f.prevKind == "let" || f.prevKind == "return" || isOp(f.prevKind) {
			f.space()
		}
	case ",", ".", "..", ":", "(":
		// no space before
	default:
		if f.preTight {
			// tight to a unary +/- before
		} else if !noSpaceAfter(f.prevKind) {
			f.space()
		}
	}
	switch kind {
	case "{":
		f.sb.WriteString("{")
		f.depth++
		// Request a line break after an opening brace (so one-line blocks are
		// expanded), but let a real newline token in the source satisfy the
		// request instead of stacking on top of it.
		openBreak = true
	case "}":
		f.sb.WriteString("}")
		if f.preDedented {
			f.preDedented = false
		} else {
			f.depth--
		}
		if f.depth < 0 {
			f.depth = 0
		}
	case "(", "[":
		f.sb.WriteString(kind)
	case ")", "]":
		f.sb.WriteString(kind)
	case ",":
		f.sb.WriteString(", ")
	case ":":
		f.sb.WriteString(": ")
	case ".":
		f.sb.WriteString(".")
	case "..":
		f.sb.WriteString("..")
	case "->":
		f.sb.WriteString("->")
	case "=", "==", "!=", "<", "<=", ">=", ">", "*", "/", "%", "-", "+":
		f.sb.WriteString(kind)
	default:
		if kind == "str" {
			f.sb.WriteString("\"")
			f.sb.WriteString(text)
			f.sb.WriteString("\"")
		} else {
			f.sb.WriteString(text)
		}
	}
	f.preTight = unary
	if startGeneric {
		f.preTight = true
	}
	if endGeneric {
		f.preTight = false
	}
	f.prevKind = kind
	f.prevText = text
	f.wantBreak = openBreak
	if kind == "fn" || kind == "view" {
		f.declName = true
	} else if kind == "ident" && f.declName {
		// Keep the flag for the following '<'; any other token clears it.
	} else if kind != "<" && kind != "," && kind != ">" {
		f.declName = false
	}
}

func (f *fmtState) space() {
	f.sb.WriteString(" ")
}

// CmdFmt implements `simpleui fmt <file.sui>` (writes to stdout, or --write in
// place).
func CmdFmt(args []string) {
	writeBack := false
	input := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--write" || a == "-w" {
			writeBack = true
			continue
		}
		if strings.HasPrefix(a, "-") {
			UsageError("unknown flag " + a)
		} else if input == "" {
			input = a
		} else {
			UsageError("unexpected argument " + a)
		}
	}
	if input == "" {
		UsageError("missing input file")
	}
	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
		os.Exit(1)
	}
	formatted, dl := FormatSource(string(data), input)
	if dl != nil && dl.HasErrors() {
		dl.Print()
		os.Exit(1)
	}
	if writeBack {
		if err := os.WriteFile(input, []byte(formatted), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
			os.Exit(1)
		}
		return
	}
	os.Stdout.WriteString(formatted)
}
