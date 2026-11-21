package sui

import "fmt"
import "os"
import "strconv"
import "strings"

// parser.go - the SimpleUI recursive-descent parser (SPEC sections 2.2-2.5
// and 4). Owned by Agent A (front-end). All user-facing failures are
// accumulated into a DiagList; this file never panics on user input.

// Parser is the recursive-descent state: the token stream, the current
// position and the shared diagnostic list.
type Parser struct {
	toks []Token
	pos  int
	file string
	dd   *DiagList
}

// ---------------------------------------------------------------------------
// Small parser helpers
// ---------------------------------------------------------------------------

// mergeDiags copies diagnostics from src into dst. It routes through
// DiagList.Extend so each diagnostic keeps its kind and hint (e.g. lexer
// diagnostics stay "lexical" even after a parser merge).
func mergeDiags(dst *DiagList, src DiagList) {
	dst.Extend(&src)
}

func newParser(toks []Token, file string, dd *DiagList) *Parser {
	var st Parser
	st.toks = toks
	st.pos = 0
	st.file = file
	st.dd = dd
	return &st
}

// tokLabel renders a token kind for diagnostics. The lexer's err tokens carry
// a human message in text; surface that instead of the bare "err" kind, and
// name end-of-file explicitly.
func tokLabel(k string, text string) string {
	if k == "err" {
		return "invalid character (" + text + ")"
	}
	if k == "eof" {
		return "end of file"
	}
	return k
}

func (p *Parser) cur() Token {
	if p.pos < len(p.toks) {
		return p.toks[p.pos]
	}
	return p.toks[len(p.toks)-1]
}

func (p *Parser) look(k int) Token {
	idx := p.pos + k
	if idx < len(p.toks) {
		return p.toks[idx]
	}
	return p.toks[len(p.toks)-1]
}

func (p *Parser) at(k string) bool {
	return p.cur().kind == k
}

func (p *Parser) adv() Token {
	t := p.cur()
	if p.pos < len(p.toks) {
		p.pos++
	}
	return t
}

func (p *Parser) accept(k string) bool {
	if p.at(k) {
		p.adv()
		return true
	}
	return false
}

// expectKind consumes a token of kind k, or records a diagnostic and returns
// the current token unconsumed so the caller can resynchronize.
func (p *Parser) expectKind(k string, what string) Token {
	if p.at(k) {
		return p.adv()
	}
	t := p.cur()
	p.dd.Add(p.file, t.line, t.col, fmt.Sprintf("expected %v %v, found %v", what, k, tokLabel(t.kind, t.text)))
	return t
}

func (p *Parser) span() Span {
	t := p.cur()
	return NewSpan(p.file, t.line, t.col)
}

func (p *Parser) spanAt(t Token) Span {
	return NewSpan(p.file, t.line, t.col)
}

func (p *Parser) diag(msg string) {
	t := p.cur()
	p.dd.Add(p.file, t.line, t.col, msg)
}

func skipNewlines(p *Parser) {
	for p.at("newline") {
		p.adv()
	}
}

// resyncToLine skips tokens until the next newline or EOF. It is the
// statement-level recovery strategy: stop at the first bad construct and pick
// up again at the next statement boundary.
func resyncToLine(p *Parser) {
	for !p.at("newline") && !p.at("eof") {
		p.adv()
	}
}

// stmtEndCheck enforces that a statement is terminated by a newline, a
// closing brace or EOF (SPEC section 4). Anything else is a diagnostic
// followed by a resync to the next statement boundary.
func stmtEndCheck(p *Parser) {
	k := p.cur().kind
	if k == "newline" || k == "}" || k == "eof" {
		return
	}
	t := p.cur()
	p.dd.Add(p.file, t.line, t.col, fmt.Sprintf("expected newline after statement, found %v", tokLabel(t.kind, t.text)))
	resyncToLine(p)
}

// dirOf returns everything in file up to the last '/' or '\' (or "." when
// there is none).
func dirOf(file string) string {
	last := -1
	n := len(file)
	for i := 0; i < n; i++ {
		c := file[i]
		if c == '/' || c == '\\' {
			last = i
		}
	}
	if last < 0 {
		return "."
	}
	return file[:last]
}

func pathSep() string {
	if os.PathSeparator == '\\' {
		return "\\"
	}
	return "/"
}

func pathJoin(dir string, rel string) string {
	if dir == "" || dir == "." {
		return rel
	}
	return strings.Join([]string{dir, rel}, pathSep())
}

// normPath canonicalizes separators and resolves '.'/'..' segments so the
// same physical file always produces the same string (used for import dedupe
// and cycle detection).
func normPath(path string) string {
	sep := pathSep()
	var canon string
	if sep == "\\" {
		canon = strings.ReplaceAll(path, "/", "\\")
	} else {
		canon = strings.ReplaceAll(path, "\\", "/")
	}
	var segs []string
	if sep == "\\" {
		segs = strings.Split(canon, "\\")
	} else {
		segs = strings.Split(canon, "/")
	}
	var out []string
	for i := 0; i < len(segs); i++ {
		s := segs[i]
		if s == "" || s == "." {
			continue
		}
		if s == ".." {
			if len(out) > 0 && out[len(out)-1] != ".." && !strings.HasSuffix(out[len(out)-1], ":") {
				var tmp []string
				for k := 0; k < len(out)-1; k++ {
					tmp = append(tmp, out[k])
				}
				out = tmp
			} else {
				out = append(out, s)
			}
		} else {
			out = append(out, s)
		}
	}
	return strings.Join(out, sep)
}

// ---------------------------------------------------------------------------
// Entry points (SPEC section 4)
// ---------------------------------------------------------------------------

// ParseSource parses a single file into a Prog. Every node carries a 1-based
// Span; Prog.dir is the directory of `file`.
func ParseSource(source string, file string) (*Prog, DiagList) {
	RegisterSource(file, source)
	var dl DiagList
	dl.Kind("syntax")
	toks, ldx := Lex(source, file)
	mergeDiags(&dl, ldx)
	p := newParser(toks, file, &dl)
	var prog Prog
	prog.file = file
	prog.dir = dirOf(file)
	prog.security = defaultSecurityPolicy
	skipNewlines(p)
	for !p.at("eof") {
		k := p.cur().kind
		if k == "newline" {
			p.adv()
		} else if k == "import" {
			imp := parseImport(p)
			if imp != nil {
				prog.imports = append(prog.imports, *imp)
			}
		} else if k == "state" {
			sd := parseState(p)
			if sd != nil {
				prog.states = append(prog.states, *sd)
			}
		} else if k == "const" {
			cd := parseConst(p)
			if cd != nil {
				prog.consts = append(prog.consts, *cd)
			}
		} else if k == "computed" {
			cd := parseComputed(p)
			if cd != nil {
				prog.computed = append(prog.computed, *cd)
			}
		} else if k == "type" {
			td := parseTypeDecl(p)
			if td != nil {
				prog.types = append(prog.types, *td)
			}
		} else if k == "fn" || k == "async" {
			fd := parseFunc(p)
			if fd != nil {
				prog.funcs = append(prog.funcs, *fd)
			}
		} else if k == "view" {
			vd := parseView(p)
			if vd != nil {
				prog.views = append(prog.views, *vd)
			}
		} else if k == "css" {
			start := p.cur()
			text := parseCss(p)
			if msg := cssExternalReferenceError(text); msg != "" {
				p.dd.Add(p.file, start.line, start.col, msg)
			}
			if text != "" {
				prog.css = append(prog.css, text)
				prog.cssSpans = append(prog.cssSpans, NewSpan(p.file, start.line, start.col))
			}
		} else if k == "ident" && p.cur().text == "enum" {
			if decl := parseEnum(p); decl != nil {
				prog.enums = append(prog.enums, *decl)
			}
		} else if k == "ident" && p.cur().text == "i18n" {
			if decl := parseI18n(p); decl != nil {
				prog.i18n = append(prog.i18n, *decl)
			}
		} else {
			t := p.cur()
			p.dd.Add(p.file, t.line, t.col, fmt.Sprintf("unexpected %v at top level", tokLabel(t.kind, t.text)))
			resyncToLine(p)
			if p.at("newline") {
				p.adv()
			}
		}
		skipNewlines(p)
	}
	return &prog, dl
}

// ---------------------------------------------------------------------------
// Top-level declarations
// ---------------------------------------------------------------------------

// parseImport: 'import' '{' ident { ',' ident } '}' 'from' String
func parseImport(p *Parser) *ImportDecl {
	sp := p.span()
	p.adv() // import
	if !p.accept("{") {
		p.diag("expected '{' after import")
		resyncToLine(p)
		return nil
	}
	var names []string
	skipNewlines(p)
	for !p.at("}") {
		if p.at("eof") {
			p.diag("unexpected end of file in import list")
			break
		}
		t := p.cur()
		if t.kind != "ident" {
			p.diag(fmt.Sprintf("expected identifier in import list, found %v", tokLabel(t.kind, t.text)))
			resyncToLine(p)
			break
		}
		p.adv()
		names = append(names, t.text)
		if !p.accept(",") {
			break
		}
		skipNewlines(p)
	}
	if !p.accept("}") {
		p.diag("expected '}' after import list")
	}
	skipNewlines(p)
	if !p.accept("from") {
		p.diag("expected 'from' after import list")
	}
	skipNewlines(p)
	s := p.cur()
	if s.kind != "str" {
		p.diag("expected a string after 'from' (module name or file path)")
		resyncToLine(p)
		return nil
	}
	p.adv()
	kind := "file"
	if s.text == "std" {
		kind = "std"
	}
	return &ImportDecl{names: names, kind: kind, path: s.text, span: sp}
}

// i18nPluralCategories is the fixed CLDR order used in source and generated
// catalogs.
var i18nPluralCategories = []string{"zero", "one", "two", "few", "many", "other"}

func isI18nPluralCategory(s string) bool {
	for _, c := range i18nPluralCategories {
		if c == s {
			return true
		}
	}
	return false
}

// parseI18n parses one top-level translation catalog:
//
//	i18n "it" {
//	  "save": "Salva"
//	  "items": { one: "{count} elemento", other: "{count} elementi" }
//	}
//
// The locale and keys are strings; values are raw strings (message placeholders
// such as {name} are kept literally) or a plural block. Entries may be
// separated by commas and/or newlines.
func parseI18n(p *Parser) *I18nDecl {
	sp := p.span()
	p.adv() // i18n
	loc := p.cur()
	if loc.kind != "str" {
		p.diag("expected a locale string after 'i18n'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	locale := strings.TrimSpace(suiUnescape(loc.text, p.file, loc.line, loc.col, p.dd))
	decl := &I18nDecl{locale: locale, span: sp}
	if !p.accept("{") {
		p.diag("expected '{' after the i18n locale")
		resyncToLine(p)
		return nil
	}
	seen := map[string]bool{}
	skipNewlines(p)
	for !p.at("}") && !p.at("eof") {
		keyTok := p.cur()
		if keyTok.kind != "str" {
			p.diag("expected a quoted message key in the i18n block")
			resyncToLine(p)
			skipNewlines(p)
			continue
		}
		p.adv()
		key := strings.TrimSpace(suiUnescape(keyTok.text, p.file, keyTok.line, keyTok.col, p.dd))
		if key == "" {
			p.dd.Add(p.file, keyTok.line, keyTok.col, "i18n message key must not be empty")
		}
		if seen[key] {
			p.dd.Add(p.file, keyTok.line, keyTok.col, "duplicate i18n message key "+key)
		}
		seen[key] = true
		skipNewlines(p)
		if !p.accept(":") {
			p.diag("expected ':' after the i18n message key")
			resyncToLine(p)
			skipNewlines(p)
			continue
		}
		skipNewlines(p)
		msg := I18nMessage{key: key, span: Span{file: p.file, line: keyTok.line, col: keyTok.col}}
		if p.at("{") {
			p.adv()
			msg.plural = map[string]string{}
			skipNewlines(p)
			for !p.at("}") && !p.at("eof") {
				cat := p.cur()
				if cat.kind != "ident" {
					p.diag("expected a plural category (zero, one, two, few, many, other)")
					resyncToLine(p)
					skipNewlines(p)
					continue
				}
				p.adv()
				if !isI18nPluralCategory(cat.text) {
					p.dd.Add(p.file, cat.line, cat.col, "unknown plural category \""+cat.text+"\" (use zero, one, two, few, many or other)")
				}
				skipNewlines(p)
				if !p.accept(":") {
					p.diag("expected ':' after the plural category")
					resyncToLine(p)
					skipNewlines(p)
					continue
				}
				skipNewlines(p)
				val := p.cur()
				if val.kind != "str" {
					p.diag("expected a quoted string after the plural category")
					resyncToLine(p)
					skipNewlines(p)
					continue
				}
				p.adv()
				msg.plural[cat.text] = suiUnescape(val.text, p.file, val.line, val.col, p.dd)
				if !p.accept(",") {
					skipNewlines(p)
				}
			}
			if !p.accept("}") {
				p.diag("expected '}' after the plural forms")
			}
		} else {
			val := p.cur()
			if val.kind != "str" {
				p.diag("expected a quoted message string")
				resyncToLine(p)
				skipNewlines(p)
				continue
			}
			p.adv()
			msg.text = suiUnescape(val.text, p.file, val.line, val.col, p.dd)
		}
		decl.messages = append(decl.messages, msg)
		if !p.accept(",") {
			skipNewlines(p)
		}
		skipNewlines(p)
	}
	if !p.accept("}") {
		p.diag("expected '}' after the i18n block")
		resyncToLine(p)
	}
	return decl
}

// parseCss: 'css' String  — a raw CSS snippet appended verbatim to style.css.
// The string follows the usual SUI escaping (\n \t \\ \" \{ \}), so CSS braces
// are written escaped in the source and unescaped here.
func parseCss(p *Parser) string {
	p.adv() // css
	t := p.cur()
	if t.kind != "str" {
		p.diag("expected a string after 'css' (the raw CSS text)")
		resyncToLine(p)
		return ""
	}
	p.adv()
	return suiUnescape(t.text, p.file, t.line, t.col, p.dd)
}

// suiUnescape decodes the SUI string escapes: \n \t \\ \" \{ \}. Invalid
// escapes are reported as diagnostics and kept literally.
func suiUnescape(raw string, file string, line, col int, dl *DiagList) string {
	var sb strings.Builder
	n := len(raw)
	i := 0
	for i < n {
		c := raw[i]
		if c == '\\' {
			if i+1 >= n {
				dl.Add(file, line, col, "dangling backslash at end of string")
				sb.WriteString("\\")
				i++
				continue
			}
			e := raw[i+1]
			if e == 'n' {
				sb.WriteString("\n")
			} else if e == 't' {
				sb.WriteString("\t")
			} else if e == '\\' {
				sb.WriteString("\\")
			} else if e == '"' {
				sb.WriteString("\"")
			} else if e == '{' {
				sb.WriteString("{")
			} else if e == '}' {
				sb.WriteString("}")
			} else {
				dl.Add(file, line, col, fmt.Sprintf("invalid escape sequence '\\%c'", e))
				sb.WriteString("\\")
				sb.WriteByte(e)
			}
			i += 2
		} else {
			sb.WriteByte(c)
			i++
		}
	}
	return sb.String()
}

// parseState: 'state' ident [ ':' TypeName ] '=' Expr
func parseState(p *Parser) *StateDecl {
	sp := p.span()
	p.adv() // state
	name := p.cur()
	if name.kind != "ident" {
		p.diag("expected identifier after 'state'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	typename := ""
	if p.accept(":") {
		typename = parseTypeName(p)
	}
	if !p.accept("=") {
		p.diag("expected '=' in state declaration")
		resyncToLine(p)
		return nil
	}
	init := parseExpr(p)
	stmtEndCheck(p)
	return &StateDecl{name: name.text, typename: typename, init: init, span: sp}
}

// parseConst: 'const' ident [ ':' TypeName ] '=' Expr
func parseConst(p *Parser) *ConstDecl {
	sp := p.span()
	p.adv() // const
	name := p.cur()
	if name.kind != "ident" {
		p.diag("expected identifier after 'const'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	typename := ""
	if p.accept(":") {
		typename = parseTypeName(p)
	}
	if !p.accept("=") {
		p.diag("expected '=' in const declaration")
		resyncToLine(p)
		return nil
	}
	init := parseExpr(p)
	stmtEndCheck(p)
	return &ConstDecl{name: name.text, typename: typename, init: init, span: sp}
}

// parseComputed: 'computed' ident [ ':' TypeName ] '=' Expr
func parseComputed(p *Parser) *ComputedDecl {
	sp := p.span()
	p.adv() // computed
	name := p.cur()
	if name.kind != "ident" {
		p.diag("expected identifier after 'computed'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	typename := ""
	if p.accept(":") {
		typename = parseTypeName(p)
	}
	if !p.accept("=") {
		p.diag("expected '=' in computed declaration")
		resyncToLine(p)
		return nil
	}
	init := parseExpr(p)
	stmtEndCheck(p)
	return &ComputedDecl{name: name.text, typename: typename, init: init, span: sp}
}

// parseEnum: 'enum' ident '{' ident (','|newline)* '}'
func parseEnum(p *Parser) *EnumDecl {
	sp := p.span()
	p.adv() // enum
	name := p.cur()
	if name.kind != "ident" {
		p.diag("expected an enum name after 'enum'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	if !p.accept("{") {
		p.diag("expected '{' to open enum " + name.text)
		resyncToLine(p)
		return nil
	}
	var members []string
	seen := map[string]bool{}
	skipNewlines(p)
	for !p.at("}") {
		if p.at("eof") {
			p.diag("unexpected end of file: enum " + name.text + " not closed")
			break
		}
		if p.at("newline") || p.at(",") {
			p.adv()
			continue
		}
		m := p.cur()
		if m.kind != "ident" {
			p.diag("expected an enum member name")
			resyncToLine(p)
			continue
		}
		if seen[m.text] {
			p.dd.Add(p.file, m.line, m.col, "duplicate enum member "+m.text+" in "+name.text)
		}
		seen[m.text] = true
		members = append(members, m.text)
		p.adv()
		if !p.accept(",") {
			skipNewlines(p)
		}
	}
	if !p.accept("}") {
		p.diag("expected '}' to close enum " + name.text)
	}
	stmtEndCheck(p)
	return &EnumDecl{name: name.text, members: members, span: sp}
}

// parseTypeDecl: 'type' ident 'struct' '{' (ident ':' TypeName) (','|newline)* '}'
func parseTypeDecl(p *Parser) *TypeDecl {
	sp := p.span()
	p.adv() // type
	name := p.cur()
	if name.kind != "ident" {
		p.diag("expected a type name after 'type'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	if !p.accept("struct") {
		p.diag("expected 'struct' in type declaration")
		resyncToLine(p)
		return nil
	}
	if !p.accept("{") {
		p.diag("expected '{' to open a struct type")
		resyncToLine(p)
		return nil
	}
	skipNewlines(p)
	var fields []Field
	seen := map[string]bool{}
	for !p.at("}") {
		if p.at("eof") {
			p.diag("unexpected end of file: struct type not closed")
			break
		}
		if p.at("newline") || p.at(",") {
			p.adv()
			continue
		}
		f := p.cur()
		if f.kind != "ident" {
			p.diag(fmt.Sprintf("expected a field name, found %v", tokLabel(f.kind, f.text)))
			resyncToLine(p)
			continue
		}
		readonly := false
		if f.text == "readonly" {
			readonly = true
			p.adv()
			f = p.cur()
			if f.kind != "ident" {
				p.diag("expected a field name after 'readonly'")
				resyncToLine(p)
				continue
			}
		}
		p.adv()
		fsp := p.span()
		if seen[f.text] {
			p.dd.Add(p.file, fsp.line, fsp.col, "duplicate field "+f.text+" in struct "+name.text)
		}
		seen[f.text] = true
		if !p.accept(":") {
			p.diag("expected ':' after field name")
			resyncToLine(p)
			continue
		}
		tn := parseTypeName(p)
		fields = append(fields, Field{name: f.text, typename: tn, readonly: readonly, span: fsp})
	}
	if !p.accept("}") {
		p.diag("expected '}' to close struct type")
	}
	stmtEndCheck(p)
	return &TypeDecl{name: name.text, fields: fields, span: sp}
}

// parseTypeName parses a type with unions (`string | int`), the `?` suffix
// (`string?` = `string | null`) and literal types (`"success" | "error"`).
func parseTypeName(p *Parser) string {
	first := parseTypeAtom(p)
	if first == "" {
		return ""
	}
	var r strings.Builder
	r.WriteString(first)
	for p.accept("|") {
		r.WriteString(" | ")
		r.WriteString(parseTypeAtom(p))
	}
	for p.accept("?") {
		r.WriteString(" | null")
	}
	return r.String()
}

// parseTypeAtom parses one non-union type name: primitives, user structs or
// enums, array[T], map[string]V, fn, null, or a literal type.
func parseTypeAtom(p *Parser) string {
	t := p.cur()
	if t.kind == "str" {
		p.adv()
		return strconv.Quote(suiUnescape(t.text, p.file, t.line, t.col, p.dd))
	}
	if t.kind == "-" {
		p.adv()
		n := p.cur()
		if n.kind != "int" && n.kind != "float" {
			p.diag("expected a number after '-' in a literal type")
			resyncToLine(p)
			return ""
		}
		p.adv()
		return "-" + n.text
	}
	if t.kind == "int" || t.kind == "float" || t.kind == "true" || t.kind == "false" {
		p.adv()
		return t.text
	}
	if t.kind != "ident" && t.kind != "null" && t.kind != "fn" {
		p.diag(fmt.Sprintf("expected a type name, found %v", tokLabel(t.kind, t.text)))
		resyncToLine(p)
		return ""
	}
	p.adv()
	var r strings.Builder
	r.WriteString(t.text)
	if p.accept("[") {
		r.WriteString("[")
		if t.text == "map" {
			// The key is parsed as a full type so `map[string | int]...` is
			// rejected by the type checker; the value stays a single atom so a
			// top-level union applies to the whole map (`map[string]int | null`).
			r.WriteString(parseTypeName(p))
			if !p.accept("]") {
				p.diag("expected ']' inside map type")
			}
			r.WriteString("]")
			r.WriteString(parseTypeAtom(p))
		} else {
			r.WriteString(parseTypeName(p))
			if !p.accept("]") {
				p.diag("expected ']' inside array type")
			}
			r.WriteString("]")
		}
	}
	return r.String()
}

// parseFunc: [ 'async' ] 'fn' ident '(' [ Params ] ')' [ '->' TypeName ] Block
// v0.1 requires explicit annotations: every parameter needs ':' TypeName and
// the return type must be given with '->' (SPEC sections 2.2 and 4).
func parseFunc(p *Parser) *FuncDecl {
	sp := p.span()
	isAsync := false
	if p.accept("async") {
		isAsync = true
		if !p.accept("fn") {
			p.diag("expected 'fn' after 'async'")
			resyncToLine(p)
			return nil
		}
	} else {
		p.adv() // fn
	}
	name := p.cur()
	if name.kind != "ident" {
		p.diag("expected function name after 'fn'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	var typeParams []string
	if p.accept("<") {
		seenTP := map[string]bool{}
		for !p.at(">") {
			if p.at("eof") {
				p.diag("unexpected end of file in generic parameter list")
				break
			}
			tp := p.cur()
			if tp.kind != "ident" {
				p.diag("expected a generic type parameter name")
				resyncToLine(p)
				break
			}
			if seenTP[tp.text] {
				p.diag("duplicate generic type parameter " + tp.text)
			}
			seenTP[tp.text] = true
			typeParams = append(typeParams, tp.text)
			p.adv()
			if !p.accept(",") {
				break
			}
			skipNewlines(p)
		}
		if !p.accept(">") {
			p.diag("expected '>' to close the generic parameter list")
		}
	}
	if !p.accept("(") {
		p.diag("expected '(' after function name")
		resyncToLine(p)
		return nil
	}
	var params []Param
	skipNewlines(p)
	for !p.at(")") {
		if p.at("eof") {
			p.diag("unexpected end of file in parameter list")
			break
		}
		pt := p.cur()
		psp := p.span()
		isRest := false
		if p.at("...") {
			p.adv()
			isRest = true
			pt = p.cur()
		}
		if pt.kind != "ident" {
			p.diag("expected a parameter name")
			resyncToLine(p)
			break
		}
		p.adv()
		if !p.accept(":") {
			t := p.cur()
			p.dd.Add(p.file, t.line, t.col, fmt.Sprintf("parameter %v requires a ':' type annotation", pt.text))
		}
		ptype := parseTypeName(p)
		var def *Expr
		if p.accept("=") {
			def = parseExpr(p)
		}
		if isRest && def != nil {
			p.diag("rest parameter " + pt.text + " cannot have a default value")
		}
		params = append(params, Param{name: pt.text, typename: ptype, dflt: def, rest: isRest, span: psp})
		if !p.accept(",") {
			break
		}
		skipNewlines(p)
	}
	if !p.accept(")") {
		p.diag("expected ')' to close parameter list")
	}
	returnType := ""
	var returnTypes []string
	if p.accept("->") {
		// Multi-return: '->' '(' TypeName ',' TypeName ')'
		if p.at("(") {
			p.adv()
			skipNewlines(p)
			for !p.at(")") {
				if p.at("eof") {
					break
				}
				tn := parseTypeName(p)
				returnTypes = append(returnTypes, tn)
				skipNewlines(p)
				if !p.accept(",") {
					break
				}
				skipNewlines(p)
			}
			if !p.accept(")") {
				p.diag("expected ')' to close the multi-return type list")
				resyncToLine(p)
			}
			if len(returnTypes) >= 1 {
				returnType = returnTypes[0]
			}
		} else {
			returnType = parseTypeName(p)
			if returnType != "" {
				returnTypes = append(returnTypes, returnType)
			}
		}
	} else {
		t := p.cur()
		p.dd.Add(p.file, t.line, t.col, fmt.Sprintf("function %v must declare a return type with '->'", name.text))
	}
	body := parseBlock(p)
	return &FuncDecl{name: name.text, async: isAsync, typeParams: typeParams, params: params, returnType: returnType, returnTypes: returnTypes, body: body, span: sp}
}

// parseView: 'view' ident '{' ViewBody '}'
func parseView(p *Parser) *ViewDecl {
	sp := p.span()
	p.adv() // view
	name := p.cur()
	if name.kind != "ident" {
		p.diag("expected view name after 'view'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	var typeParams []string
	if p.accept("<") {
		seenTP := map[string]bool{}
		for !p.at(">") {
			if p.at("eof") {
				p.diag("unexpected end of file in generic parameter list")
				break
			}
			tp := p.cur()
			if tp.kind != "ident" {
				p.diag("expected a generic type parameter name")
				resyncToLine(p)
				break
			}
			if seenTP[tp.text] {
				p.diag("duplicate generic type parameter " + tp.text)
			}
			seenTP[tp.text] = true
			typeParams = append(typeParams, tp.text)
			p.adv()
			if !p.accept(",") {
				break
			}
			skipNewlines(p)
		}
		if !p.accept(">") {
			p.diag("expected '>' to close the generic parameter list")
		}
	}
	// Component parameters: view Card(title: string, accent: string) { ... }
	var params []Param
	if p.at("(") {
		p.adv()
		skipNewlines(p)
		seen := map[string]bool{}
		if !p.at(")") {
			for true {
				pt := p.cur()
				if pt.kind != "ident" {
					p.diag(fmt.Sprintf("expected a view parameter name, found %v", tokLabel(pt.kind, pt.text)))
					resyncToLine(p)
					break
				}
				p.adv()
				psp := p.span()
				if seen[pt.text] {
					p.dd.Add(p.file, psp.line, psp.col, "duplicate view parameter "+pt.text)
				}
				seen[pt.text] = true
				if !p.accept(":") {
					p.diag("expected ':' after view parameter name")
					resyncToLine(p)
					break
				}
				tn := parseTypeName(p)
				params = append(params, Param{name: pt.text, typename: tn, span: psp})
				skipNewlines(p)
				if !p.accept(",") {
					break
				}
				skipNewlines(p)
			}
		}
		if !p.accept(")") {
			p.diag("expected ')' to close view parameter list")
			resyncToLine(p)
		}
	}
	if !p.accept("{") {
		p.diag("expected '{' to open a view")
		resyncToLine(p)
		return nil
	}
	skipNewlines(p)
	var title []StrPart
	titleSeen := false
	itemCount := 0
	var body []ViewNode
	var onLoad string
	onLoadSeen := false
	var instState []StateDecl
	var methods []FuncDecl
	for !p.at("}") {
		if p.at("eof") {
			p.diag("unexpected end of file: view not closed")
			break
		}
		if p.at("newline") {
			p.adv()
			continue
		}
		k := p.cur().kind
		if k == "state" {
			sd := parseState(p)
			if sd != nil {
				instState = append(instState, *sd)
			}
			itemCount++
			skipNewlines(p)
			continue
		}
		if k == "fn" || k == "async" {
			md := parseFunc(p)
			if md != nil {
				methods = append(methods, *md)
			}
			itemCount++
			skipNewlines(p)
			continue
		}
		if k == "title" {
			p.adv()
			if titleSeen {
				p.diag("duplicate 'title' in view")
			} else if itemCount > 0 {
				p.diag("view 'title' must be the first item")
			}
			titleSeen = true
			st := p.cur()
			if st.kind == "str" {
				p.adv()
				title = parseStrInto(p, st.text, p.spanAt(st))
			} else {
				p.diag("expected a string after 'title'")
				resyncToLine(p)
			}
			itemCount++
			skipNewlines(p)
			continue
		}
		if k == "ident" && p.cur().text == "onLoad" {
			p.adv()
			if onLoadSeen {
				p.diag("duplicate 'onLoad' in view")
			}
			onLoadSeen = true
			if !p.accept(":") {
				p.diag("expected ':' after onLoad")
				resyncToLine(p)
			} else {
				nf := p.cur()
				if nf.kind == "ident" {
					onLoad = nf.text
					p.adv()
				} else {
					p.diag(fmt.Sprintf("expected a function name after onLoad:, found %v", tokLabel(nf.kind, nf.text)))
					resyncToLine(p)
				}
			}
			itemCount++
			skipNewlines(p)
			continue
		}
		if k == "if" {
			n := parseViewIf(p)
			if n != nil {
				body = append(body, *n)
			}
		} else if isListKw(p) {
			n := parseViewList(p)
			if n != nil {
				body = append(body, *n)
			}
		} else if k == "ident" {
			n := parseViewElem(p)
			if n != nil {
				body = append(body, *n)
			}
		} else {
			t := p.cur()
			p.diag(fmt.Sprintf("unexpected %v in view body", tokLabel(t.kind, t.text)))
			resyncToLine(p)
		}
		itemCount++
		skipNewlines(p)
	}
	if !p.accept("}") {
		p.diag("expected '}' to close view")
	}
	return &ViewDecl{name: name.text, typeParams: typeParams, params: params, inst_state: instState, methods: methods, title: title, body: body, onLoad: onLoad, span: sp}
}

// isListKw reports whether the current token is the `list` view keyword.
// (`list` is not a lexer keyword, so the parser distinguishes it by text.)
func isListKw(p *Parser) bool {
	return p.at("ident") && p.cur().text == "list"
}

// structLitAhead reports whether the token stream after a '{' starts a struct
// literal field list: an `ident` followed by ':' (newlines/commas allowed in
// between).
func structLitAhead(p *Parser) bool {
	i := p.pos + 1 // p.cur() is '{'
	for i < len(p.toks) {
		k := p.toks[i].kind
		if k == "newline" || k == "," {
			i++
			continue
		}
		if k == "ident" {
			j := i + 1
			for j < len(p.toks) && (p.toks[j].kind == "newline" || p.toks[j].kind == ",") {
				j++
			}
			return j < len(p.toks) && p.toks[j].kind == ":"
		}
		return false
	}
	return false
}

// compoundOpFor maps a compound-assignment token to the binary op NAME the
// parser uses for arithmetic (add/sub/mul/div/mod, matching the AST contract)
// so desugared `x += e` behaves exactly like `x = x + e`.
func compoundOpFor(k string) string {
	if k == "+=" {
		return "add"
	}
	if k == "-=" {
		return "sub"
	}
	if k == "*=" {
		return "mul"
	}
	if k == "/=" {
		return "div"
	}
	if k == "%=" {
		return "mod"
	}
	return ""
}

// parseViewBlock parses '{' ViewBody '}' for if/list bodies.
func parseViewBlock(p *Parser) []*ViewNode {
	if !p.accept("{") {
		p.diag("expected '{' to open a view block")
		resyncToLine(p)
		return []*ViewNode{}
	}
	return parseViewBodyRest(p)
}

// parseViewBodyFromOpen parses the children of an element whose '{' has just
// been consumed.
func parseViewBodyFromOpen(p *Parser) []*ViewNode {
	return parseViewBodyRest(p)
}

// parseViewBodyRest continues parsing view nodes until the closing '}' of an
// already-open view block.
func parseViewBodyRest(p *Parser) []*ViewNode {
	skipNewlines(p)
	var nodes []*ViewNode
	for !p.at("}") {
		if p.at("eof") {
			p.diag("unexpected end of file in view block")
			break
		}
		if p.at("newline") {
			p.adv()
			continue
		}
		k := p.cur().kind
		if k == "if" {
			n := parseViewIf(p)
			if n != nil {
				nodes = append(nodes, n)
			}
		} else if isListKw(p) {
			n := parseViewList(p)
			if n != nil {
				nodes = append(nodes, n)
			}
		} else if k == "ident" {
			n := parseViewElem(p)
			if n != nil {
				nodes = append(nodes, n)
			}
		} else {
			t := p.cur()
			p.diag(fmt.Sprintf("unexpected %v in view block", tokLabel(t.kind, t.text)))
			resyncToLine(p)
		}
		skipNewlines(p)
	}
	if !p.accept("}") {
		p.diag("expected '}' to close view block")
	}
	return nodes
}

// parseViewIf: 'if' Expr '{' ViewBody '}' [ 'else' '{' ViewBody '}' ]
func parseViewIf(p *Parser) *ViewNode {
	sp := p.span()
	p.adv() // if
	cond := parseExpr(p)
	then_ := parseViewBlock(p)
	var res ViewNode
	res = ViewNode{kind: "if", cond: cond, then_: then_, span: sp}
	if p.accept("else") {
		res.else_ = parseViewBlock(p)
	}
	return &res
}

// parseViewList: 'list' ident 'in' Expr '{' ViewBody '}'
func parseViewList(p *Parser) *ViewNode {
	sp := p.span()
	p.adv() // list
	item := p.cur()
	if item.kind != "ident" {
		p.diag("expected identifier after 'list'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	item2 := ""
	if p.at(",") {
		p.adv()
		v := p.cur()
		if v.kind != "ident" {
			p.diag("expected a second identifier after ',' (map value)")
			resyncToLine(p)
			return nil
		}
		item2 = v.text
		p.adv()
	}
	if !p.accept("in") {
		p.diag("expected 'in' in view list")
		resyncToLine(p)
		return nil
	}
	iterable := parseExpr(p)
	var key *Expr
	if p.cur().kind == "ident" && p.cur().text == "key" {
		p.adv()
		p.accept(":")
		key = parseExpr(p)
	}
	body := parseViewBlock(p)
	return &ViewNode{kind: "list", item: item.text, item2: item2, iterable: iterable, key: key, body: body, span: sp}
}

// parseViewElem: Tag [ String ] { Attr } [ '{' ViewBody '}' ]
func parseViewElem(p *Parser) *ViewNode {
	t := p.cur()
	sp := p.spanAt(t)
	p.adv() // tag
	var elem Element
	elem.tag = t.text
	elem.span = sp
	if p.at("str") {
		ts := p.cur()
		p.adv()
		elem.text = parseStrInto(p, ts.text, p.spanAt(ts))
	}
	for p.at("ident") {
		ts := p.cur()
		asp := p.spanAt(ts)
		p.adv() // attribute key
		if !p.accept(":") {
			p.diag("expected ':' after attribute name")
			resyncToLine(p)
			break
		}
		var a Attr
		a.key = ts.text
		a.span = asp
		// Attribute values are full expressions. A bare string literal that
		// carries no {interpolation} is kept as plain text (hasExpr=false) so
		// existing code paths that read a.text keep working; anything else
		// (concatenation, conditionals, identifiers, ...) is stored as an expr.
		expr := parseExpr(p)
		if expr != nil && expr.kind == "str" && !strPartsHaveExpr(expr.parts) {
			a.text = expr.parts
			a.hasExpr = false
		} else {
			a.expr = expr
			a.hasExpr = true
		}
		elem.attrs = append(elem.attrs, a)
	}
	if p.accept("{") {
		childPts := parseViewBodyFromOpen(p)
		var cv []ViewNode
		for i := 0; i < len(childPts); i++ {
			cv = append(cv, *childPts[i])
		}
		elem.children = cv
	}
	return &ViewNode{kind: "element", elem: &elem, span: sp}
}

// strPartsHaveExpr reports whether parsed string parts contain any dynamic
// {expression} segment (as opposed to plain literal text).
func strPartsHaveExpr(parts []StrPart) bool {
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "expr" {
			return true
		}
	}
	return false
}

func parseStrInto(p *Parser, raw string, sp Span) []StrPart {
	parts, dl := parseStringParts(raw, sp, p)
	mergeDiags(p.dd, dl)
	return parts
}

// ---------------------------------------------------------------------------
// Statements (SPEC section 2.3)
// ---------------------------------------------------------------------------

func parseBlock(p *Parser) []*Stmt {
	if !p.accept("{") {
		p.diag("expected '{' to open a block")
		resyncToLine(p)
		return []*Stmt{}
	}
	skipNewlines(p)
	var stmts []*Stmt
	for !p.at("}") {
		if p.at("eof") {
			p.diag("unexpected end of file: block not closed")
			break
		}
		if p.at("newline") {
			p.adv()
			continue
		}
		before := p.pos
		st := parseStmt(p)
		if st != nil {
			stmts = append(stmts, st)
		}
		if p.pos == before {
			p.adv() // guarantee progress on pathological errors
		}
		skipNewlines(p)
	}
	if !p.accept("}") {
		p.diag("expected '}' to close block")
	}
	return stmts
}

func parseStmt(p *Parser) *Stmt {
	if p.at("let") {
		s := parseLet(p)
		stmtEndCheck(p)
		return s
	}
	if p.at("return") {
		s := parseReturn(p)
		stmtEndCheck(p)
		return s
	}
	if p.at("break") {
		sp := p.span()
		p.adv()
		stmtEndCheck(p)
		return &Stmt{kind: "break", span: sp}
	}
	if p.at("continue") {
		sp := p.span()
		p.adv()
		stmtEndCheck(p)
		return &Stmt{kind: "continue", span: sp}
	}
	if p.at("if") {
		s := parseIf(p)
		stmtEndCheck(p)
		return s
	}
	if p.at("loop") {
		s := parseLoop(p)
		stmtEndCheck(p)
		return s
	}
	if p.at("while") {
		s := parseWhile(p)
		stmtEndCheck(p)
		return s
	}
	if p.cur().kind == "ident" && p.cur().text == "switch" && p.look(1).kind != "(" {
		s := parseSwitch(p)
		stmtEndCheck(p)
		return s
	}
	if p.at("}") {
		p.diag("unexpected '}'")
		p.adv()
		return nil
	}
	s := parseAssignOrExpr(p)
	stmtEndCheck(p)
	return s
}

// parseLet: 'let' (ident | '{' ident, ... '}' | '[' ident, ... ']') '=' Expr
func parseLet(p *Parser) *Stmt {
	sp := p.span()
	p.adv() // let
	var names []string
	patternKind := ""
	if p.at("{") || p.at("[") {
		close := "}"
		patternKind = "object"
		if p.at("[") {
			close = "]"
			patternKind = "array"
		}
		p.adv()
		skipNewlines(p)
		for !p.at(close) {
			if p.at("eof") {
				p.diag("unexpected end of file: destructuring pattern not closed")
				break
			}
			if p.at("newline") || p.at(",") {
				p.adv()
				continue
			}
			n := p.cur()
			if n.kind != "ident" {
				p.diag("expected a name in the destructuring pattern")
				resyncToLine(p)
				break
			}
			names = append(names, n.text)
			p.adv()
			if !p.accept(",") {
				skipNewlines(p)
				break
			}
			skipNewlines(p)
		}
		if !p.accept(close) {
			p.diag("expected '" + close + "' to close the destructuring pattern")
		}
	} else {
		name := p.cur()
		if name.kind != "ident" {
			p.diag("expected identifier after 'let'")
			resyncToLine(p)
			return nil
		}
		p.adv()
		names = append(names, name.text)
	}
	if !p.accept("=") {
		p.diag("expected '=' in let declaration")
		resyncToLine(p)
		return nil
	}
	init := parseExpr(p)
	if patternKind != "" {
		return &Stmt{kind: "let", pattern: names, patternKind: patternKind, init: init, span: sp}
	}
	return &Stmt{kind: "let", name: names[0], init: init, span: sp}
}

// parseReturn: 'return' [ Expr ]
func parseReturn(p *Parser) *Stmt {
	sp := p.span()
	p.adv() // return
	var val *Expr
	k := p.cur().kind
	if k != "newline" && k != "}" && k != "eof" {
		val = parseExpr(p)
	}
	// Multi-return: `return a, b` (tuple) parses a comma-separated list.
	var vals []*Expr
	if val != nil && p.at(",") {
		vals = append(vals, val)
		for p.at(",") {
			p.adv()
			e := parseExpr(p)
			if e == nil {
				break
			}
			vals = append(vals, e)
		}
	}
	if len(vals) > 0 {
		return &Stmt{kind: "return", values: vals, span: sp}
	}
	return &Stmt{kind: "return", value: val, span: sp}
}

// parseIf: 'if' Expr Block { 'else' 'if' Expr Block } [ 'else' Block ]
// parseSwitch: 'switch' Expr '{' { 'case' Expr Block | 'default' Block } '}'
func parseSwitch(p *Parser) *Stmt {
	sp := p.span()
	p.adv() // switch
	subject := parseExpr(p)
	if !p.accept("{") {
		p.diag("expected '{' after the switch value")
		resyncToLine(p)
		return nil
	}
	skipNewlines(p)
	var cases []SwitchCase
	var defaultBody []*Stmt
	hasDefault := false
	for !p.at("}") && !p.at("eof") {
		if p.at("newline") {
			p.adv()
			continue
		}
		if p.cur().kind == "ident" && p.cur().text == "case" {
			csp := p.span()
			p.adv()
			e := parseExpr(p)
			if !p.at("{") {
				p.diag("expected '{' after the case value")
				resyncToLine(p)
				skipNewlines(p)
				continue
			}
			body := parseBlock(p)
			cases = append(cases, SwitchCase{expr: e, body: body, span: csp})
			skipNewlines(p)
			continue
		}
		if p.cur().kind == "ident" && p.cur().text == "default" {
			p.adv()
			hasDefault = true
			if !p.at("{") {
				p.diag("expected '{' after default")
				resyncToLine(p)
				skipNewlines(p)
				continue
			}
			defaultBody = parseBlock(p)
			skipNewlines(p)
			continue
		}
		p.diag("expected 'case' or 'default' in switch")
		resyncToLine(p)
		skipNewlines(p)
	}
	if !p.accept("}") {
		p.diag("expected '}' to close switch")
	}
	return &Stmt{kind: "switch", init: subject, cases: cases, default_: defaultBody, hasDefault: hasDefault, span: sp}
}

func parseIf(p *Parser) *Stmt {
	sp := p.span()
	p.adv() // if
	cond := parseExpr(p)
	then_ := parseBlock(p)
	var res Stmt
	res = Stmt{kind: "if", cond: cond, then_: then_, span: sp}
	if p.at("else") {
		p.adv()
		if p.at("if") {
			inner := parseIf(p)
			res.else_ = append(res.else_, inner)
		} else {
			res.else_ = parseBlock(p)
		}
	}
	return &res
}

// parseLoop: 'loop' ident 'in' Expr [ '..' Expr ] Block
func parseLoop(p *Parser) *Stmt {
	sp := p.span()
	p.adv() // loop
	item := p.cur()
	if item.kind != "ident" {
		p.diag("expected identifier after 'loop'")
		resyncToLine(p)
		return nil
	}
	p.adv()
	item2 := ""
	if p.at(",") {
		p.adv()
		v := p.cur()
		if v.kind != "ident" {
			p.diag("expected a second identifier after ',' (map value)")
			resyncToLine(p)
			return nil
		}
		item2 = v.text
		p.adv()
	}
	if !p.accept("in") {
		p.diag("expected 'in' in loop")
		resyncToLine(p)
		return nil
	}
	iterable := parseExpr(p)
	if p.accept("..") {
		hi := parseExpr(p)
		iterable = &Expr{kind: "range", lo: iterable, hi: hi, span: sp}
	}
	body := parseBlock(p)
	return &Stmt{kind: "loop", item: item.text, item2: item2, iterable: iterable, body: body, span: sp}
}

// parseWhile: 'while' Expr Block
func parseWhile(p *Parser) *Stmt {
	sp := p.span()
	p.adv() // while
	cond := parseExpr(p)
	body := parseBlock(p)
	return &Stmt{kind: "while", cond: cond, body: body, span: sp}
}

// parseAssignOrExpr handles the statement-level ambiguity between a plain
// expression statement, a single assignment `a = e`, and a tuple assignment
// `a, b = e` (the two-target form requires a std intrinsic call on the RHS;
// the typechecker enforces that).
func parseAssignOrExpr(p *Parser) *Stmt {
	sp := p.span()
	if p.at("ident") {
		saved := p.pos
		var targets []string
		t := p.adv()
		targets = append(targets, t.text)
		for p.at(",") && p.look(1).kind == "ident" {
			p.adv() // ,
			t2 := p.adv()
			targets = append(targets, t2.text)
		}
		if p.at("=") {
			p.adv()
			rhs := parseExpr(p)
			return &Stmt{kind: "assign", targets: targets, rhs: rhs, span: sp}
		}
		// Compound assignment `x += e` desugars to `x = x + e` (single target
		// only), so the checker and the code generator keep one rule each.
		if len(targets) == 1 {
			compound := compoundOpFor(p.cur().kind)
			if compound != "" {
				p.adv()
				rhs := parseExpr(p)
				lhs := &Expr{kind: "ident", name: targets[0], span: sp}
				bin := &Expr{kind: "binary", op: compound, l: lhs, r: rhs, span: sp}
				return &Stmt{kind: "assign", targets: targets, rhs: bin, span: sp}
			}
		}
		p.pos = saved

		// Struct field / index assignment `u.name = e`, `arr[i] = e`,
		// `m[k] += e`. Only a single lvalue target is allowed.
		lv := &Expr{kind: "ident", name: p.cur().text, span: p.spanAt(p.cur())}
		p.adv()
		ok := true
		for true {
			if p.at(".") && p.look(1).kind == "ident" {
				p.adv()
				m := p.adv()
				lv = &Expr{kind: "member", obj: lv, memb: m.text, span: sp}
			} else if p.at("[") {
				p.adv()
				ix := parseExpr(p)
				if !p.accept("]") {
					ok = false
					break
				}
				lv = &Expr{kind: "index", obj: lv, idx: ix, span: sp}
			} else {
				break
			}
		}
		if ok {
			op := p.cur().kind
			co := ""
			isEq := op == "="
			if !isEq {
				co = compoundOpFor(op)
			}
			if isEq || co != "" {
				p.adv()
				rhs := parseExpr(p)
				final := rhs
				if co != "" {
					final = &Expr{kind: "binary", op: co, l: lv, r: rhs, span: sp}
				}
				return &Stmt{kind: "assign", lvalue: lv, rhs: final, span: sp}
			}
		}
		p.pos = saved
	}
	e := parseExpr(p)
	if e == nil {
		return nil
	}
	// Postfix `x++` / `x--` sugar on a bare identifier.
	if (p.at("++") || p.at("--")) && e.kind == "ident" {
		op := "add"
		if p.at("--") {
			op = "sub"
		}
		p.adv()
		one := &Expr{kind: "int", i: 1, span: sp}
		bin := &Expr{kind: "binary", op: op, l: e, r: one, span: sp}
		return &Stmt{kind: "assign", targets: []string{e.name}, rhs: bin, span: sp}
	}
	// NOTE: expression statements store their expression in `init` (the only
	// free *Expr field in the shared Stmt shape). Consumers of Stmt kind
	// "expr" should read `init`.
	return &Stmt{kind: "expr", init: e, span: sp}
}

// ---------------------------------------------------------------------------
// Expressions (SPEC section 2.4, precedence low to high)
// ---------------------------------------------------------------------------

func parseExpr(p *Parser) *Expr {
	c := parseOr(p)
	if p.at("?") && c != nil {
		p.adv()
		a := parseExpr(p)
		var b *Expr
		if p.accept(":") {
			b = parseExpr(p)
		} else {
			p.diag("expected ':' in conditional expression")
		}
		return &Expr{kind: "cond", cond: c, l: a, r: b, span: c.span}
	}
	return c
}

func parseOr(p *Parser) *Expr {
	l := parseAnd(p)
	for p.at("or") && l != nil {
		p.adv()
		r := parseAnd(p)
		if r == nil {
			break
		}
		l = &Expr{kind: "binary", op: "or", l: l, r: r, span: l.span}
	}
	return l
}

func parseAnd(p *Parser) *Expr {
	l := parseCmp(p)
	for p.at("and") && l != nil {
		p.adv()
		r := parseCmp(p)
		if r == nil {
			break
		}
		l = &Expr{kind: "binary", op: "and", l: l, r: r, span: l.span}
	}
	return l
}

func parseCmp(p *Parser) *Expr {
	l := parseAdd(p)
	for l != nil {
		k := p.cur().kind
		if k == "is" {
			p.adv()
			t := parseTypeAtom(p)
			l = &Expr{kind: "is", l: l, name: t, span: l.span}
			continue
		}
		if k != "==" && k != "!=" && k != "<" && k != "<=" && k != ">" && k != ">=" {
			break
		}
		p.adv()
		r := parseAdd(p)
		if r == nil {
			break
		}
		l = &Expr{kind: "binary", op: k, l: l, r: r, span: l.span}
	}
	return l
}

func parseAdd(p *Parser) *Expr {
	l := parseMul(p)
	for l != nil {
		k := p.cur().kind
		if k != "+" && k != "-" {
			break
		}
		p.adv()
		op := "add"
		if k == "-" {
			op = "sub"
		}
		r := parseMul(p)
		if r == nil {
			break
		}
		l = &Expr{kind: "binary", op: op, l: l, r: r, span: l.span}
	}
	return l
}

func parseMul(p *Parser) *Expr {
	l := parseUnary(p)
	for l != nil {
		k := p.cur().kind
		if k != "*" && k != "/" && k != "%" {
			break
		}
		p.adv()
		op := "mul"
		if k == "/" {
			op = "div"
		} else if k == "%" {
			op = "mod"
		}
		r := parseUnary(p)
		if r == nil {
			break
		}
		l = &Expr{kind: "binary", op: op, l: l, r: r, span: l.span}
	}
	return l
}

func parseUnary(p *Parser) *Expr {
	if p.at("not") {
		sp := p.span()
		p.adv()
		operand := parseUnary(p)
		return &Expr{kind: "unary", op: "not", operand: operand, span: sp}
	}
	if p.at("-") {
		sp := p.span()
		p.adv()
		operand := parseUnary(p)
		return &Expr{kind: "unary", op: "neg", operand: operand, span: sp}
	}
	return parsePostfix(p)
}

func parsePostfix(p *Parser) *Expr {
	e := parsePrimary(p)
	for e != nil {
		if p.at(".") {
			p.adv()
			m := p.cur()
			if m.kind != "ident" {
				p.diag("expected identifier after '.'")
				resyncToLine(p)
				break
			}
			p.adv()
			e = &Expr{kind: "member", obj: e, memb: m.text, span: e.span}
		} else if p.at("[") {
			p.adv()
			idx := parseExpr(p)
			if !p.accept("]") {
				p.diag("expected ']' to close index expression")
			}
			e = &Expr{kind: "index", obj: e, idx: idx, span: e.span}
		} else if p.at("(") {
			p.adv()
			args, names := parseArgs(p)
			e = &Expr{kind: "call", target: e, args: args, argNames: names, span: e.span}
		} else {
			break
		}
	}
	return e
}

func parseArgs(p *Parser) ([]*Expr, []string) {
	var args []*Expr
	var names []string
	if p.accept(")") {
		return args, names
	}
	for true {
		name := ""
		if p.cur().kind == "ident" && p.look(1).kind == ":" {
			name = p.cur().text
			p.adv()
			p.adv()
		}
		a := parseExpr(p)
		args = append(args, a)
		names = append(names, name)
		if !p.accept(",") {
			break
		}
	}
	if !p.accept(")") {
		p.diag("expected ')' to close argument list")
	}
	return args, names
}

func parsePrimary(p *Parser) *Expr {
	t := p.cur()
	if t.kind == "int" {
		p.adv()
		n, e := strconv.Atoi(t.text)
		if e != nil {
			n = 0
		}
		return &Expr{kind: "int", i: n, span: p.spanAt(t)}
	}
	if t.kind == "float" {
		p.adv()
		f, e := strconv.ParseFloat(t.text, 64)
		if e != nil {
			f = 0
		}
		return &Expr{kind: "float", f: f, span: p.spanAt(t)}
	}
	if t.kind == "str" {
		p.adv()
		parts := parseStrInto(p, t.text, p.spanAt(t))
		return &Expr{kind: "str", parts: parts, span: p.spanAt(t)}
	}
	if t.kind == "true" || t.kind == "false" {
		p.adv()
		return &Expr{kind: "bool", b: t.kind == "true", span: p.spanAt(t)}
	}
	if t.kind == "null" {
		p.adv()
		return &Expr{kind: "null", span: p.spanAt(t)}
	}
	if t.kind == "ident" {
		p.adv()
		// Struct constructor: `User{ id: "1", name: "Alice" }`. Ambiguous with
		// a block ("if x {", "list r in rows {", an element with children), so
		// only treat it as a struct literal when the braces start a field
		// list (`ident ':' ...`). Empty `{}` after an identifier is always a
		// block, never an (empty) constructor.
		if p.at("{") && structLitAhead(p) {
			var keys []string
			var vals []*Expr
			p.adv()
			skipNewlines(p)
			for !p.at("}") {
				if p.at("eof") {
					p.diag("unexpected end of file: struct literal not closed")
					break
				}
				if p.at("newline") || p.at(",") {
					p.adv()
					continue
				}
				k := p.cur()
				if k.kind != "ident" {
					p.diag(fmt.Sprintf("expected a field name in struct literal, found %v", tokLabel(k.kind, k.text)))
					resyncToLine(p)
					continue
				}
				p.adv()
				if !p.accept(":") {
					p.diag("expected ':' after field name in struct literal")
					resyncToLine(p)
					continue
				}
				v := parseExpr(p)
				keys = append(keys, k.text)
				vals = append(vals, v)
			}
			p.accept("}")
			return &Expr{kind: "structlit", name: t.text, keys: keys, vals: vals, span: p.spanAt(t)}
		}
		return &Expr{kind: "ident", name: t.text, span: p.spanAt(t)}
	}
	if t.kind == "(" {
		p.adv()
		e := parseExpr(p)
		if !p.accept(")") {
			p.diag("expected ')' to close parenthesized expression")
		}
		return e
	}
	if t.kind == "[" {
		p.adv()
		var els []*Expr
		var spreads []bool
		skipNewlines(p)
		if p.accept("]") {
			return &Expr{kind: "array", elements: els, spreads: spreads, span: p.spanAt(t)}
		}
		for true {
			spread := p.accept("...")
			e := parseExpr(p)
			els = append(els, e)
			spreads = append(spreads, spread)
			if !p.accept(",") {
				break
			}
			skipNewlines(p)
		}
		skipNewlines(p)
		if !p.accept("]") {
			p.diag("expected ']' to close array literal")
		}
		return &Expr{kind: "array", elements: els, spreads: spreads, span: p.spanAt(t)}
	}
	if t.kind == "{" {
		p.adv()
		skipNewlines(p)
		var keys []string
		var vals []*Expr
		var spreads []bool
		if p.accept("}") {
			return &Expr{kind: "map", keys: keys, vals: vals, spreads: spreads, span: p.spanAt(t)}
		}
		for true {
			skipNewlines(p)
			if p.accept("...") {
				v := parseExpr(p)
				keys = append(keys, "")
				vals = append(vals, v)
				spreads = append(spreads, true)
				skipNewlines(p)
				if !p.accept(",") {
					break
				}
				skipNewlines(p)
				continue
			}
			kt := p.cur()
			if kt.kind != "str" {
				p.diag("expected a string map key")
				resyncToLine(p)
				break
			}
			p.adv()
			kp := parseStrInto(p, kt.text, p.spanAt(kt))
			keys = append(keys, joinTextParts(p, kp))
			if !p.accept(":") {
				p.diag("expected ':' after map key")
				resyncToLine(p)
				break
			}
			skipNewlines(p)
			v := parseExpr(p)
			vals = append(vals, v)
			spreads = append(spreads, false)
			skipNewlines(p)
			if !p.accept(",") {
				break
			}
			skipNewlines(p)
		}
		if !p.accept("}") {
			p.diag("expected '}' to close map literal")
		}
		return &Expr{kind: "map", keys: keys, vals: vals, spreads: spreads, span: p.spanAt(t)}
	}
	p.diag(fmt.Sprintf("unexpected %v in expression", tokLabel(t.kind, t.text)))
	resyncToLine(p)
	return nil
}

// ---------------------------------------------------------------------------
// Strings and interpolation (SPEC sections 2.1 and 4)
// ---------------------------------------------------------------------------

// parseStringParts decodes the raw inner text of a string token into text
// parts and interpolated expression parts. Recognized escapes: \n \t \\ \"
// \{ \}. Any other backslash sequence is a diagnostic (the characters are
// kept literally). A `{...}` span is re-lexed and parsed as an expression;
// string literals inside interpolation (t("save"), map values, nested
// templates) are supported and skipped while finding the closing brace.
func parseStringParts(raw string, sp Span, p *Parser) ([]StrPart, DiagList) {
	var dl DiagList
	var parts []StrPart
	var sb strings.Builder
	n := len(raw)
	i := 0
	for i < n {
		c := raw[i]
		if c == '\\' {
			if i+1 >= n {
				dl.Add(p.file, sp.line, sp.col, "dangling backslash at end of string")
				sb.WriteString("\\")
				i++
				continue
			}
			e := raw[i+1]
			if e == 'n' {
				sb.WriteString("\n")
			} else if e == 't' {
				sb.WriteString("\t")
			} else if e == '\\' {
				sb.WriteString("\\")
			} else if e == '"' {
				sb.WriteString("\"")
			} else if e == '{' {
				sb.WriteString("{")
			} else if e == '}' {
				sb.WriteString("}")
			} else {
				dl.Add(p.file, sp.line, sp.col, fmt.Sprintf("invalid escape sequence '\\%c'", e))
				sb.WriteString("\\")
				sb.WriteByte(e)
			}
			i += 2
		} else if c == '{' {
			if sb.Len() > 0 {
				parts = append(parts, StrPart{kind: "text", text: sb.String()})
				sb.Reset()
			}
			close := scanInterpEnd(raw, i+1, &dl, p)
			if close < 0 {
				i = n
				continue
			}
			sub := raw[i+1 : close]
			if sub == "" {
				dl.Add(p.file, sp.line, sp.col, "empty interpolation '{}' in string")
			} else {
				e, ed := parseInterpExpr(sub, p)
				mergeDiags(&dl, ed)
				parts = append(parts, StrPart{kind: "expr", expr: e})
			}
			i = close + 1
		} else if c == '\n' {
			dl.Add(p.file, sp.line, sp.col, "unescaped newline inside string literal")
			sb.WriteByte(c)
			i++
		} else {
			sb.WriteByte(c)
			i++
		}
	}
	if sb.Len() > 0 {
		parts = append(parts, StrPart{kind: "text", text: sb.String()})
	}
	if len(parts) == 0 {
		parts = append(parts, StrPart{kind: "text", text: ""})
	}
	return parts, dl
}

// scanInterpEnd finds the matching '}' of an interpolation starting after the
// opening '{', tracking { ( [ nesting. A '"' starts a nested string literal
// that is skipped (with escapes) so its braces do not affect the depth.
// Returns the index of the closing '}' or -1 when unterminated.
func scanInterpEnd(raw string, start int, dl *DiagList, p *Parser) int {
	depth := 0
	j := start
	n := len(raw)
	for j < n {
		c := raw[j]
		if c == '"' {
			// A string literal inside an interpolation (t("save"), map values,
			// nested templates) is lexed by parseInterpExpr. Skip it here so
			// braces inside it do not affect the interpolation depth.
			j++
			closed := false
			for j < n {
				if raw[j] == '\\' {
					j += 2
					continue
				}
				if raw[j] == '"' {
					closed = true
					j++
					break
				}
				j++
			}
			if !closed {
				dl.Add(p.file, p.span().line, p.span().col, "unterminated string inside interpolation")
				return -1
			}
			// The nested-string scan already advanced past its closing quote;
			// continue at the character after it.
			continue
		} else if c == '{' || c == '(' || c == '[' {
			depth++
		} else if c == '}' || c == ')' || c == ']' {
			if depth == 0 {
				if c == '}' {
					return j
				}
			} else {
				depth--
			}
		}
		j++
	}
	dl.Add(p.file, p.span().line, p.span().col, "unterminated interpolation in string literal")
	return -1
}

// parseInterpExpr re-lexes and parses the inside of a `{...}` interpolation.
func parseInterpExpr(sub string, p *Parser) (*Expr, DiagList) {
	toks, dl := Lex(sub, p.file)
	ip := newParser(toks, p.file, &dl)
	skipNewlines(ip)
	e := parseExpr(ip)
	skipNewlines(ip)
	if !ip.at("eof") {
		t := ip.cur()
		dl.Add(p.file, p.span().line, p.span().col, fmt.Sprintf("unexpected %v at end of interpolation", t.kind))
	}
	return e, dl
}

// joinTextParts concatenates the text parts of a string; interpolation is a
// diagnostic here (used for map keys, which must be literal strings).
func joinTextParts(p *Parser, parts []StrPart) string {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "expr" {
			p.diag("string interpolation is not allowed in a map key")
		} else {
			sb.WriteString(parts[i].text)
		}
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// LoadProgram: entry file + recursive import merging (SPEC section 4)
// ---------------------------------------------------------------------------

// Loader accumulates the merged program while walking the import graph.
type Loader struct {
	prog     *Prog
	loaded   []string
	stack    []string
	stdNames []string
	dd       *DiagList
}

// addStd merges a std import, deduplicating module names so a module is never
// recorded twice in the final Prog.
func (l *Loader) addStd(imp ImportDecl) {
	var fresh []string
	for i := 0; i < len(imp.names); i++ {
		already := false
		for j := 0; j < len(l.stdNames); j++ {
			if l.stdNames[j] == imp.names[i] {
				already = true
				break
			}
		}
		if !already {
			fresh = append(fresh, imp.names[i])
			l.stdNames = append(l.stdNames, imp.names[i])
		}
	}
	if len(fresh) > 0 {
		l.prog.imports = append(l.prog.imports, ImportDecl{names: fresh, kind: imp.kind, path: imp.path, span: imp.span})
	}
}

// loadFile parses one file (once per physical path) and merges its
// declarations depth-first: imported files first, then the file's own.
func (l *Loader) loadFile(path string) {
	norm := normPath(path)
	for i := 0; i < len(l.stack); i++ {
		if l.stack[i] == norm {
			l.dd.Add(path, 1, 1, fmt.Sprintf("import cycle detected at %v", path))
			return
		}
	}
	for i := 0; i < len(l.loaded); i++ {
		if l.loaded[i] == norm {
			return // already parsed and merged
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		l.dd.AddEx(path, 1, 1, "io", fmt.Sprintf("cannot read file: %v", path), "check that the file exists and that the path in the import is correct")
		return
	}
	src := string(data)
	pp, d2 := ParseSource(src, path)
	mergeDiags(l.dd, d2)
	if pp == nil {
		return
	}
	l.stack = append(l.stack, norm)
	for i := 0; i < len(pp.imports); i++ {
		imp := pp.imports[i]
		if imp.kind == "std" {
			l.addStd(imp)
		} else {
			l.loadFile(normPath(pathJoin(pp.dir, imp.path)))
		}
	}
	for i := 0; i < len(pp.types); i++ {
		l.prog.types = append(l.prog.types, pp.types[i])
	}
	for i := 0; i < len(pp.enums); i++ {
		l.prog.enums = append(l.prog.enums, pp.enums[i])
	}
	for i := 0; i < len(pp.css); i++ {
		l.prog.css = append(l.prog.css, pp.css[i])
		if i < len(pp.cssSpans) {
			l.prog.cssSpans = append(l.prog.cssSpans, pp.cssSpans[i])
		}
	}
	for i := 0; i < len(pp.i18n); i++ {
		l.prog.i18n = append(l.prog.i18n, pp.i18n[i])
	}
	for i := 0; i < len(pp.consts); i++ {
		l.prog.consts = append(l.prog.consts, pp.consts[i])
	}
	for i := 0; i < len(pp.computed); i++ {
		l.prog.computed = append(l.prog.computed, pp.computed[i])
	}
	for i := 0; i < len(pp.states); i++ {
		l.prog.states = append(l.prog.states, pp.states[i])
	}
	for i := 0; i < len(pp.funcs); i++ {
		l.prog.funcs = append(l.prog.funcs, pp.funcs[i])
	}
	for i := 0; i < len(pp.views); i++ {
		l.prog.views = append(l.prog.views, pp.views[i])
	}
	l.loaded = append(l.loaded, norm)
}

// trackSym records a top-level symbol name; a duplicate name is reported at
// both declaration sites.
func trackSym(name string, sp Span, names []string, spans []Span, dl *DiagList) ([]string, []Span) {
	for i := 0; i < len(names); i++ {
		if names[i] == name {
			dl.Add(sp.file, sp.line, sp.col, fmt.Sprintf("duplicate top-level symbol %v (already declared elsewhere)", name))
			dl.Add(spans[i].file, spans[i].line, spans[i].col, fmt.Sprintf("first declaration of %v (duplicated here)", name))
			return names, spans
		}
	}
	names = append(names, name)
	spans = append(spans, sp)
	return names, spans
}

// LoadProgram reads entryFile, parses it, and resolves every file import
// ({ ident } from "./rel.sui") recursively, merging states/functions/views.
// Physical files are parsed once (deduped by normalized path); duplicate
// top-level symbols, import cycles and unreadable files are diagnostics.
func LoadProgram(entryFile string) (*Prog, DiagList) {
	var dl DiagList
	var prog Prog
	prog.file = entryFile
	prog.dir = dirOf(entryFile)
	l := &Loader{prog: &prog, dd: &dl}
	l.loadFile(entryFile)
	var names []string
	var spans []Span
	for i := 0; i < len(prog.consts); i++ {
		names, spans = trackSym(prog.consts[i].name, prog.consts[i].span, names, spans, &dl)
	}
	for i := 0; i < len(prog.computed); i++ {
		names, spans = trackSym(prog.computed[i].name, prog.computed[i].span, names, spans, &dl)
	}
	for i := 0; i < len(prog.states); i++ {
		names, spans = trackSym(prog.states[i].name, prog.states[i].span, names, spans, &dl)
	}
	for i := 0; i < len(prog.funcs); i++ {
		names, spans = trackSym(prog.funcs[i].name, prog.funcs[i].span, names, spans, &dl)
	}
	for i := 0; i < len(prog.views); i++ {
		names, spans = trackSym(prog.views[i].name, prog.views[i].span, names, spans, &dl)
	}
	cfgPath := securityPathOverride
	if cfgPath == "" {
		cfgPath = SecurityPathFor(dirOf(entryFile))
	}
	cfg, warns, serr := SecurityForEntry(entryFile, securityPathOverride)
	if serr != nil {
		dl.AddEx(cfgPath, 1, 1, "security", "security config: "+serr.Error(),
			"fix simpleui.security.json or pass --security <file>")
	} else {
		prog.security = cfg
		for _, w := range warns {
			dl.AddEx(cfgPath, 1, 1, "warning", "security config: "+w, "")
		}
	}
	return &prog, dl
}
