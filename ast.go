package main

// ast.go defines the Abstract Syntax Tree for the SimpleUI language.
//
// This file is the SHARED CONTRACT for every compiler phase:
//   - the parser PRODUCES this tree
//   - the typechecker and both code generators CONSUME it
// Do not change these shapes without updating every phase.
//
// Design notes
// ------------
// Go (this toolchain) has no `union`/`enum` keywords, so every node kind is a
// single tagged struct: all node variants share one type, differentiated by
// the `kind` string field. Slices of structs copy the struct but share the
// pointer payloads, which is exactly what we want for child lists.
//
// `kind` strings are the source of truth for switch/debug output. The JSON
// dumper (astjson.go) also prints them verbatim.

// Span locates a node inside a source file. Positions are 1-based.
type Span struct {
	file string
	line int
	col  int
}

func NewSpan(file string, line, col int) Span {
	return Span{file: file, line: line, col: col}
}

// StrPart is one segment of an interpolated string literal, e.g.
// `"hello {name}!"` becomes [Text("hello "), ExprPart(name), Text("!")].
type StrPart struct {
	kind string // "text" or "expr"
	text string
	expr *Expr
}

func NewTextPart(text string) StrPart {
	return StrPart{kind: "text", text: text}
}

func NewExprPart(e *Expr) StrPart {
	return StrPart{kind: "expr", expr: e}
}

// ---------------------------------------------------------------------------
// Expressions
//
// kind values:
//
//	"int"    i holds the value
//	"float"  f holds the value
//	"str"    parts holds the literal (possibly interpolated)
//	"bool"   b holds the value
//	"null"   (no payload)
//	"ident"  name
//	"member" obj . memb            (e.g. http.get)
//	"call"   target(args)          target is usually member or ident
//	"array"  elements
//	"map"    keys[i] = vals[i]
//	"structlit" name holds the struct type name; keys[i] = vals[i] (ident keys)
//	"index"  obj[idx]
//	"unary"  op operand            op: "neg" | "not"
//	"binary" op l r
//	         op: "add" "sub" "mul" "div" "mod" "eq" "neq" "lt" "lte" "gt"
//	             "gte" "and" "or"
//	"cond"   cond ? l : r          conditional (ternary) expression
//	"range"  lo..hi                only valid as a loop iterable
//
// ---------------------------------------------------------------------------
type Expr struct {
	kind     string
	i        int
	f        float64
	b        bool
	parts    []StrPart
	name     string
	obj      *Expr
	memb     string
	target   *Expr
	args     []*Expr
	argNames []string // parallel to args; "" for positional
	elements []*Expr
	spreads  []bool // parallel to elements (array) or vals (map)
	keys     []string
	vals     []*Expr
	idx      *Expr
	op       string
	operand  *Expr
	l        *Expr
	r        *Expr
	cond     *Expr // "cond": condition of a conditional (ternary) expression
	lo       *Expr
	hi       *Expr
	span     Span
}

// ---------------------------------------------------------------------------
// Statements
//
// kind values:
//
//	"let"      name = init          (local variable; only inside functions)
//	"assign"   targets = rhs        (1 or 2 targets; 2 targets => tuple call),
//	           or lvalue = rhs      (struct field / index assignment)
//	"expr"     expression statement (e.g. a bare function call)
//	"if"       cond { then_ } else { else_ }
//	"loop"     loop item [ , item2 ] in iterable { body }
//	           (item2 is the value variable when iterating a map)
//	"while"    (cond) { body }
//	"return"   value                (value == nil for a bare `return`)
//	"break"    (no payload)
//	"continue" (no payload)
//
// ---------------------------------------------------------------------------
type Stmt struct {
	kind        string
	name        string
	init        *Expr
	pattern     []string // let destructuring names
	patternKind string   // "object" or "array"
	targets     []string
	rhs         *Expr
	lvalue      *Expr // struct field / index target for single-target assigns
	cond        *Expr
	then_       []*Stmt
	else_       []*Stmt
	item        string
	item2       string // map value variable ("" for single-var loops)
	iterable    *Expr
	body        []*Stmt
	value       *Expr
	values      []*Expr // multi-return values (return a, b)
	span        Span
}

// ---------------------------------------------------------------------------
// Program-level declarations
// ---------------------------------------------------------------------------

type Param struct {
	name     string
	typename string // "" means "infer from call sites" (v0.1: not implemented -> error)
	dflt     *Expr  // optional default value (nil = required)
	span     Span
}

type FuncDecl struct {
	name        string
	async       bool
	params      []Param
	returnType  string   // "" means inferred; the first type for multi-return
	returnTypes []string // multi-return type list (empty = single return)
	body        []*Stmt
	span        Span
}

type StateDecl struct {
	name     string
	typename string // "" means inferred from init
	init     *Expr
	span     Span
}

// Field is one `name: TypeName` line of a user struct type.
type Field struct {
	name     string
	typename string
	span     Span
}

// TypeDecl is a user-defined struct declaration:
//
//	type User struct { id: string, name: string }
type TypeDecl struct {
	name   string
	fields []Field
	span   Span
}

type ImportDecl struct {
	names []string
	kind  string // "std" | "file"
	path  string // for kind == "file" the (already escaped) source string
	span  Span
}

// ---------------------------------------------------------------------------
// Views (the UI declaration surface)
// ---------------------------------------------------------------------------

// Attr is a single attribute on an element, e.g. class: "x" or onClick: goUp.
// For plain string values `text` holds the (possibly interpolated) parts and
// hasExpr is false. For code values (onClick, value, src, ...) the parser sets
// hasExpr true and stores the parsed expression in `expr`.
type Attr struct {
	key     string
	text    []StrPart
	hasExpr bool
	expr    *Expr
	span    Span
}

type Element struct {
	tag string // "h1" "h2" "h3" "p" "text" "button" "input" "img"
	// "row" "col" "spacer" "a"
	text     []StrPart
	attrs    []Attr
	children []ViewNode
	span     Span
}

type ViewNode struct {
	kind     string // "element" | "if" | "list"
	elem     *Element
	cond     *Expr
	then_    []*ViewNode
	else_    []*ViewNode
	item     string
	item2    string // map value variable ("" for single-var lists)
	iterable *Expr
	body     []*ViewNode
	span     Span
}

type ViewDecl struct {
	name       string
	params     []Param     // component parameters (empty for plain views)
	inst_state []StateDecl // per-instance state (`state` inside a component body)
	methods    []FuncDecl  // per-instance methods (`fn` inside a component body)
	title      []StrPart
	body       []ViewNode
	onLoad     string // function name invoked once at boot ("" = none)
	span       Span
}

// ---------------------------------------------------------------------------
// Program
// ---------------------------------------------------------------------------

type Prog struct {
	imports  []ImportDecl
	types    []TypeDecl
	states   []StateDecl
	css      []string // raw user CSS snippets (`css "..."` declarations)
	i18n     []I18nDecl
	funcs    []FuncDecl
	views    []ViewDecl
	security SecurityConfig
	file     string
	dir      string // directory of the entry source file (for resolving imports)
}

// I18nDecl is a top-level translation catalog for one locale.
//
//	i18n "it" {
//	  "save": "Salva"
//	  "items": { one: "{count} elemento", other: "{count} elementi" }
//	}
type I18nDecl struct {
	locale   string
	messages []I18nMessage
	span     Span
}

// I18nMessage is one key. text is used for simple messages; plural holds the
// CLDR categories (zero, one, two, few, many, other) for plural messages.
type I18nMessage struct {
	key    string
	text   string
	plural map[string]string
	span   Span
}
