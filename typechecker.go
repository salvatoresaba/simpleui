package main

import "fmt"
import "strconv"
import "strings"
import "sync"

// typechecker.go - the SimpleUI static type checker (SPEC section 2.6).
//
// It infers and validates the types of every declaration in a Prog and
// produces a DiagList; it never panics on user input. It also exposes
// ViewStateRefs, a pure collector that returns the sorted, unique state
// variable names a view references.
//
// Public interface:
//   Check(prog *Prog) DiagList
//   ViewStateRefs(v *ViewDecl) []string
//
// Design notes
// ------------
// - The SType `kind` field uses the vocabulary in the contract ("int" ..
//   "error") plus one internal marker: "range" for `a..b` iterables, so loops
//   and views can distinguish a range from a value type.
// - Sentinel types ("error", "unknown") produced after a diagnostic never
//   cascade: assignments, arguments and operator checks treat them as "any".
// - viewParams/types are resolved against registries populated at the start
//   of Check.

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// SType is the internal static type. `name` holds the struct type name when
// kind == "struct".
type SType struct {
	kind    string
	name    string
	elem    *SType
	key     *SType
	val     *SType
	members []*SType // union members
}

func NewSType(kind string) *SType {
	return &SType{kind: kind}
}

// typeStr renders a type for human-readable diagnostics.
func typeStr(t *SType) string {
	if t == nil {
		return "unknown"
	}
	if t.kind == "struct" || t.kind == "enum" || t.kind == "literal" || t.kind == "typevar" {
		return t.name
	}
	if t.kind == "xmlnode" {
		return "xmlNode"
	}
	if t.kind == "xmldocument" {
		return "xmlDocument"
	}
	if t.kind == "array" {
		return "array[" + typeStr(t.elem) + "]"
	}
	if t.kind == "map" {
		return "map[string]" + typeStr(t.val)
	}
	if t.kind == "union" {
		var parts []string
		for i := 0; i < len(t.members); i++ {
			parts = append(parts, typeStr(t.members[i]))
		}
		return strings.Join(parts, " | ")
	}
	return t.kind
}

// typeEqual is structural equality. "any" accepts anything; "unknown" and
// "error" sentinels also match anything so an already reported problem never
// cascades.
func typeEqual(a, b *SType) bool {
	if a == nil || b == nil {
		return true
	}
	if a.kind == "any" || b.kind == "any" {
		return true
	}
	if a.kind == "unknown" || b.kind == "unknown" || a.kind == "error" || b.kind == "error" {
		return true
	}
	if a.kind != b.kind {
		return false
	}
	if (a.kind == "struct" || a.kind == "enum" || a.kind == "literal" || a.kind == "typevar") && a.kind == b.kind {
		return a.name == b.name
	}
	if a.kind == "union" || b.kind == "union" {
		if a.kind != "union" || b.kind != "union" || len(a.members) != len(b.members) {
			return false
		}
		for i := 0; i < len(a.members); i++ {
			found := false
			for j := 0; j < len(b.members); j++ {
				if typeEqual(a.members[i], b.members[j]) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	if a.kind == "array" {
		return typeEqual(a.elem, b.elem)
	}
	if a.kind == "map" {
		return typeEqual(a.key, b.key) && typeEqual(a.val, b.val)
	}
	return true
}

func isNumericType(t *SType) bool {
	return t != nil && (t.kind == "int" || t.kind == "float")
}

// Name registries shared across the free type/name helpers (populated at the
// start of Check, like KnownStateNames before them).
// unionOf flattens and deduplicates types into a union. A single member is
// returned as-is; any/unknown absorb the union.
func unionOf(types ...*SType) *SType {
	var flat []*SType
	var add func(t *SType)
	add = func(t *SType) {
		if t == nil {
			return
		}
		if t.kind == "union" {
			for i := 0; i < len(t.members); i++ {
				add(t.members[i])
			}
			return
		}
		if t.kind == "any" || t.kind == "unknown" || t.kind == "error" {
			flat = append(flat, t)
			return
		}
		for i := 0; i < len(flat); i++ {
			if typeEqual(flat[i], t) {
				return
			}
		}
		flat = append(flat, t)
	}
	for i := 0; i < len(types); i++ {
		add(types[i])
	}
	for i := 0; i < len(flat); i++ {
		if flat[i].kind == "any" || flat[i].kind == "unknown" || flat[i].kind == "error" {
			return flat[i]
		}
	}
	if len(flat) == 0 {
		return NewSType("unknown")
	}
	if len(flat) == 1 {
		return flat[0]
	}
	return &SType{kind: "union", members: flat}
}

// isNumericLiteralName reports whether a literal type name is an integer or
// float literal (used by assignable).
func isNumericLiteralName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && c != '.' && c != '-' && c != '+' && c != 'e' && c != 'E' {
			return false
		}
	}
	return true
}

// canonicalFloatName formats a float literal with a decimal point when it is
// integral, so 1e3 and 1000.0 stay float literals rather than becoming ints.
func canonicalFloatName(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}

// assignable reports whether a value of type src can be stored where dst is
// expected. Unions are covariant: every src member must fit dst, and a value
// may fit any member of a dst union. Literal types fit their base type.
func assignable(src, dst *SType) bool {
	if src == nil || dst == nil {
		return true
	}
	if src.kind == "any" || dst.kind == "any" || src.kind == "unknown" || dst.kind == "unknown" || src.kind == "error" || dst.kind == "error" {
		return true
	}
	if src.kind == "union" && dst.kind == "union" {
		for i := 0; i < len(src.members); i++ {
			if !assignable(src.members[i], dst) {
				return false
			}
		}
		return true
	}
	if dst.kind == "union" {
		for i := 0; i < len(dst.members); i++ {
			if assignable(src, dst.members[i]) {
				return true
			}
		}
		return false
	}
	if src.kind == "union" {
		for i := 0; i < len(src.members); i++ {
			if !assignable(src.members[i], dst) {
				return false
			}
		}
		return true
	}
	if src.kind == "func" && dst.kind == "fn" {
		return true
	}
	if src.kind == "xmldocument" && dst.kind == "xmlnode" {
		return true
	}
	if src.kind == "array" && dst.kind == "array" {
		return assignable(src.elem, dst.elem)
	}
	if src.kind == "map" && dst.kind == "map" {
		return assignable(src.val, dst.val)
	}
	if src.kind == "literal" {
		switch dst.kind {
		case "string":
			return strings.HasPrefix(src.name, "\"")
		case "bool":
			return src.name == "true" || src.name == "false"
		case "int":
			return isNumericLiteralName(src.name) && !strings.ContainsAny(src.name, ".eE")
		case "float":
			return isNumericLiteralName(src.name)
		}
	}
	return typeEqual(src, dst)
}

// overlaps reports whether a runtime value could satisfy both types. Used by
// type guards and equality comparisons against union members.
func overlaps(a, b *SType) bool {
	if a == nil || b == nil {
		return true
	}
	if a.kind == "any" || b.kind == "any" || a.kind == "unknown" || b.kind == "unknown" || a.kind == "error" || b.kind == "error" {
		return true
	}
	if a.kind == "union" {
		for i := 0; i < len(a.members); i++ {
			if overlaps(a.members[i], b) {
				return true
			}
		}
		return false
	}
	if b.kind == "union" {
		for i := 0; i < len(b.members); i++ {
			if overlaps(a, b.members[i]) {
				return true
			}
		}
		return false
	}
	return assignable(a, b) || assignable(b, a)
}

// narrowType returns the possible type after `value is target` (positive) or
// its negation. Unions lose the matching members; non-unions keep their type.
func narrowType(t, target *SType, positive bool) *SType {
	if t == nil {
		return target
	}
	if t.kind == "union" {
		var out []*SType
		for i := 0; i < len(t.members); i++ {
			match := overlaps(t.members[i], target)
			if (positive && match) || (!positive && !match) {
				out = append(out, t.members[i])
			}
		}
		if len(out) == 0 {
			if positive {
				return target
			}
			return t
		}
		return unionOf(out...)
	}
	if positive {
		return target
	}
	return t
}

var structRegistry = map[string][]Field{} // struct name -> fields
var enumRegistry = map[string][]string{}  // enum name -> members
var KnownStateNames = map[string]bool{}   // state names for ViewStateRefs
var viewInvoked = map[string]bool{}       // views invoked as components
// Check writes the registries above. The GUI can issue a live check and a
// preview render concurrently, so serialize Check instead of allowing two
// goroutines to write the same maps.
var checkMu sync.Mutex

func isTypeNameChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// parseTypeFrom parses one type name off the FRONT of s: flat names, bracket
// forms array[T] / map[string]V, user struct names (registry) and, when vars
// is non-nil, type parameters of the enclosing generic function.
func parseTypeFrom(s string) (*SType, string, string) {
	return parseTypeFromWith(s, nil)
}

func parseTypeFromWith(s string, vars map[string]bool) (*SType, string, string) {
	s = strings.TrimLeft(s, " \t")
	n := len(s)
	if n == 0 {
		return &SType{kind: "error"}, "", "missing type name"
	}
	if s[0] == '"' {
		// String literal type, e.g. "success" | "error".
		i := 1
		for i < n {
			if s[i] == '\\' {
				i += 2
				continue
			}
			if s[i] == '"' {
				i++
				break
			}
			i++
		}
		if i > n || s[i-1] != '"' {
			return &SType{kind: "error"}, "", "unterminated string literal type"
		}
		raw := s[:i]
		val, err := strconv.Unquote(raw)
		if err != nil {
			return &SType{kind: "error"}, "", "invalid string literal type " + raw
		}
		return &SType{kind: "literal", name: strconv.Quote(val)}, s[i:], ""
	}
	if s[0] == '-' || (s[0] >= '0' && s[0] <= '9') {
		i := 1
		for i < n && (s[i] == '.' || s[i] == 'e' || s[i] == 'E' || s[i] == '+' || s[i] == '-' || (s[i] >= '0' && s[i] <= '9')) {
			i++
		}
		lit := s[:i]
		if strings.ContainsAny(lit, ".eE") {
			f, err := strconv.ParseFloat(lit, 64)
			if err != nil {
				return &SType{kind: "error"}, "", "invalid numeric literal type " + lit
			}
			return &SType{kind: "literal", name: canonicalFloatName(f)}, s[i:], ""
		}
		iv, err := strconv.ParseInt(lit, 10, 64)
		if err != nil {
			return &SType{kind: "error"}, "", "invalid numeric literal type " + lit
		}
		return &SType{kind: "literal", name: strconv.FormatInt(iv, 10)}, s[i:], ""
	}
	if strings.HasPrefix(s, "true") && (n == 4 || !isTypeNameChar(s[4])) {
		return &SType{kind: "literal", name: "true"}, s[4:], ""
	}
	if strings.HasPrefix(s, "false") && (n == 5 || !isTypeNameChar(s[5])) {
		return &SType{kind: "literal", name: "false"}, s[5:], ""
	}
	j := 0
	for j < n && isTypeNameChar(s[j]) {
		j++
	}
	base := s[:j]
	rest := s[j:]
	if base == "int" || base == "float" || base == "string" || base == "bool" || base == "null" || base == "any" || base == "fn" {
		return &SType{kind: base}, rest, ""
	}
	if base == "xmlNode" || base == "xmlnode" {
		return &SType{kind: "xmlnode"}, rest, ""
	}
	if base == "xmlDocument" || base == "xmldocument" {
		return &SType{kind: "xmldocument"}, rest, ""
	}
	if base == "array" || base == "map" {
		if len(rest) < 2 || rest[0] != '[' {
			return &SType{kind: "error"}, "", base + " type name requires a bracket element type"
		}
		depth := 0
		found := -1
		k := 0
		for k < len(rest) {
			if rest[k] == '[' {
				depth++
			} else if rest[k] == ']' {
				depth--
				if depth == 0 {
					found = k
					break
				}
			}
			k++
		}
		if found < 0 {
			return &SType{kind: "error"}, "", "unclosed '[' in type name " + base
		}
		inner := rest[1:found]
		tail := rest[found+1:]
		if base == "array" {
			it, rest2, err2 := parseUnionFromWith(inner, vars)
			if err2 != "" {
				return it, "", err2
			}
			if rest2 != "" {
				return it, "", "invalid array element type name"
			}
			return &SType{kind: "array", elem: it}, tail, ""
		}
		kt, rest2, err2 := parseUnionFromWith(inner, vars)
		if err2 != "" {
			return kt, "", err2
		}
		if kt.kind != "string" {
			return &SType{kind: "error"}, "", "map keys must be of type string"
		}
		if rest2 != "" {
			return kt, "", "invalid map key type name"
		}
		vt, resto, err3 := parseTypeFromWith(tail, vars)
		if err3 != "" {
			return vt, "", err3
		}
		return &SType{kind: "map", key: &SType{kind: "string"}, val: vt}, resto, ""
	}
	if vars != nil && vars[base] {
		return &SType{kind: "typevar", name: base}, rest, ""
	}
	if structRegistry[base] != nil {
		return &SType{kind: "struct", name: base}, rest, ""
	}
	if enumRegistry[base] != nil {
		return &SType{kind: "enum", name: base}, rest, ""
	}
	return &SType{kind: "error"}, "", "unknown type name " + base
}

// parseUnionFrom parses one type followed by optional `|` members and `?`
// (which means `| null`).
func parseUnionFrom(s string) (*SType, string, string) {
	return parseUnionFromWith(s, nil)
}

func parseUnionFromWith(s string, vars map[string]bool) (*SType, string, string) {
	s = strings.TrimLeft(s, " \t")
	first, rest, err := parseTypeFromWith(s, vars)
	if err != "" {
		return first, "", err
	}
	members := []*SType{first}
	for {
		rest = strings.TrimLeft(rest, " \t")
		if strings.HasPrefix(rest, "|") {
			rest = strings.TrimLeft(rest[1:], " \t")
			next, r2, e := parseTypeFromWith(rest, vars)
			if e != "" {
				return next, "", e
			}
			members = append(members, next)
			rest = r2
			continue
		}
		if strings.HasPrefix(rest, "?") {
			members = append(members, NewSType("null"))
			rest = rest[1:]
			continue
		}
		break
	}
	if len(members) == 1 {
		return first, rest, ""
	}
	return unionOf(members...), rest, ""
}

func parseSUIType(s string) (*SType, string) {
	return parseSUITypeWith(s, nil)
}

func parseSUITypeWith(s string, vars map[string]bool) (*SType, string) {
	typ, rest, err := parseUnionFromWith(s, vars)
	if err != "" {
		return typ, err
	}
	if strings.TrimSpace(rest) != "" {
		return typ, "did not parse the complete type name"
	}
	return typ, ""
}

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

func capFirstName(s string) string {
	if len(s) == 0 {
		return s
	}
	c := s[0]
	if c >= 'a' && c <= 'z' {
		var buf []byte
		buf = append(buf, c-32)
		buf = append(buf, s[1:]...)
		return string(buf)
	}
	return s
}

func reservedStateName(name string) string {
	if strings.HasPrefix(name, "__sui") || strings.HasPrefix(name, "__SUI_") {
		return "state name \"" + name + "\" is reserved (runtime global)"
	}
	return ""
}

func reservedFuncName(name string, prog *Prog) string {
	if strings.HasPrefix(name, "render_") {
		return "function name \"" + name + "\" is reserved (generated view renderer)"
	}
	if strings.HasPrefix(name, "__sui") || strings.HasPrefix(name, "__SUI_") {
		return "function name \"" + name + "\" is reserved (runtime global)"
	}
	for i := 0; i < len(prog.states); i++ {
		if name == "set"+capFirstName(prog.states[i].name) {
			return "function name \"" + name + "\" is reserved (generated setter for state \"" + prog.states[i].name + "\")"
		}
	}
	return ""
}

// itoaStr renders a small integer without strconv.
func itoaStr(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var digits []byte
	for v > 0 {
		digits = append(digits, byte('0'+v%10))
		v = v / 10
	}
	var out []byte
	if neg {
		out = append(out, '-')
	}
	for i := len(digits) - 1; i >= 0; i-- {
		out = append(out, digits[i])
	}
	return string(out)
}

func ordinalNd(n string) string {
	if strings.HasSuffix(n, "1") {
		return n + "st"
	}
	if strings.HasSuffix(n, "2") {
		return n + "nd"
	}
	if strings.HasSuffix(n, "3") {
		return n + "rd"
	}
	return n + "th"
}

// editDistance is the Levenshtein string distance.
func editDistance(a, b string) int {
	la := len(a)
	lb := len(b)
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			best := del
			if ins < best {
				best = ins
			}
			if sub < best {
				best = sub
			}
			cur[j] = best
		}
		for j := 0; j <= lb; j++ {
			prev[j] = cur[j]
		}
	}
	return prev[lb]
}

func nearestName(name string, candidates []string) string {
	best := ""
	bestDist := 2
	for i := 0; i < len(candidates); i++ {
		cand := candidates[i]
		if cand == name {
			continue
		}
		d := editDistance(name, cand)
		if d < bestDist || (d == bestDist && best != "" && cand < best) {
			best = cand
			bestDist = d
		}
	}
	if bestDist <= 2 {
		return best
	}
	return ""
}

func sortedStrings(a []string) []string {
	for i := 1; i < len(a); i++ {
		j := i
		for j > 0 && a[j] < a[j-1] {
			tmp := a[j]
			a[j] = a[j-1]
			a[j-1] = tmp
			j--
		}
	}
	return a
}

// derefNodeList turns a []*ViewNode into a []ViewNode for the value-based
// walkers.
func derefNodeList(nodes []*ViewNode) []ViewNode {
	var out []ViewNode
	for i := 0; i < len(nodes); i++ {
		if nodes[i] != nil {
			out = append(out, *nodes[i])
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Checker state
// ---------------------------------------------------------------------------

type Scope struct {
	vars []VarEntry
}

type VarEntry struct {
	name string
	typ  *SType
	span Span
}

type FnSig struct {
	name       string
	params     []SType
	paramNames []string
	ret        *SType
	defaults   []*Expr  // per-param default value (nil = required)
	rets       []*SType // multi-return list (ret == rets[0])
	multi      bool
	async      bool
	rest       bool     // last parameter collects the remaining arguments
	typeParams []string // generic type parameters (empty for ordinary functions)
}

type Checker struct {
	dd             *DiagList
	states         map[string]*SType
	declaredStates map[string]*SType
	consts         map[string]*SType
	computed       map[string]*SType
	sigs           map[string]*FnSig
	symKinds       map[string]string
	symSpans       map[string]Span
	stdImported    map[string]bool
	scopes         []Scope
	loopDepth      int
	switchDepth    int
	fnDepth        int
	inFn           bool
	curRet         *SType
	curRets        []*SType
	curAsync       bool
	seenReturn     bool
	viewParams     map[string][]Param
	viewTypeParams map[string][]string
	viewIsComp     bool // current view is a component (invoked or has params)
	childrenSeen   bool
	curView        string            // name of the view currently being checked
	curTypeParams  map[string]bool   // type parameters of the current generic function/view
	methodNames    map[string]bool   // view-method names (globally unique)
	instTypes      map[string]*SType // current view's instance-state types
	security       SecurityConfig    // compile-time security policy
}

func newChecker(dd *DiagList) *Checker {
	var c Checker
	c.dd = dd
	c.states = map[string]*SType{}
	c.declaredStates = map[string]*SType{}
	c.consts = map[string]*SType{}
	c.computed = map[string]*SType{}
	c.sigs = map[string]*FnSig{}
	c.symKinds = map[string]string{}
	c.symSpans = map[string]Span{}
	c.stdImported = map[string]bool{}
	c.viewParams = map[string][]Param{}
	c.viewTypeParams = map[string][]string{}
	c.methodNames = map[string]bool{}
	c.instTypes = map[string]*SType{}
	return &c
}

func (c *Checker) diagAt(sp Span, msg string) {
	c.dd.AddEx(sp.file, sp.line, sp.col, "type", msg, "")
}

func (c *Checker) diagAtEx(sp Span, msg string, hint string) {
	c.dd.AddEx(sp.file, sp.line, sp.col, "type", msg, hint)
}

func (c *Checker) pushScope() {
	var sc Scope
	c.scopes = append(c.scopes, sc)
}

func (c *Checker) popScope() {
	if len(c.scopes) == 0 {
		return
	}
	var ns []Scope
	for i := 0; i < len(c.scopes)-1; i++ {
		ns = append(ns, c.scopes[i])
	}
	c.scopes = ns
}

func (c *Checker) lookupVar(name string) *VarEntry {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		for j := 0; j < len(c.scopes[i].vars); j++ {
			if c.scopes[i].vars[j].name == name {
				return &c.scopes[i].vars[j]
			}
		}
	}
	return nil
}

func (c *Checker) declareVar(name string, t *SType, sp Span) {
	if len(c.scopes) == 0 {
		c.pushScope()
	}
	top := len(c.scopes) - 1
	c.scopes[top].vars = append(c.scopes[top].vars, VarEntry{name: name, typ: t, span: sp})
}

func (c *Checker) declareSym(name string, kind string, sp Span) {
	if c.symKinds[name] != "" {
		c.diagAt(sp, "duplicate top-level symbol "+name+" (already declared elsewhere)")
		first := c.symSpans[name]
		if first.file != "" {
			c.diagAt(first, "first declaration of "+name+" (duplicated here)")
		}
		return
	}
	c.symKinds[name] = kind
	c.symSpans[name] = sp
}

func (c *Checker) allNames() []string {
	var out []string
	for name := range c.states {
		out = append(out, name)
	}
	for name := range c.sigs {
		out = append(out, name)
	}
	for name := range c.consts {
		out = append(out, name)
	}
	for name := range c.computed {
		out = append(out, name)
	}
	return out
}

func (c *Checker) funcNames() []string {
	var out []string
	for name := range c.sigs {
		out = append(out, name)
	}
	return out
}

// ---------------------------------------------------------------------------
// Check: the public entry point
// ---------------------------------------------------------------------------

// checkI18n validates translation catalogs: unique locales, unique keys,
// non-empty keys and the required plural forms.
func (c *Checker) checkI18n(decls []I18nDecl) {
	seenLocales := map[string]bool{}
	for i := 0; i < len(decls); i++ {
		d := decls[i]
		loc := strings.TrimSpace(d.locale)
		if loc == "" {
			c.diagAt(d.span, "i18n locale must not be empty")
		} else if strings.ContainsAny(loc, " \t\r\n") {
			c.diagAt(d.span, "i18n locale \""+loc+"\" must not contain spaces")
		}
		if seenLocales[loc] {
			c.diagAt(d.span, "duplicate i18n locale \""+loc+"\"")
		}
		seenLocales[loc] = true
		seenKeys := map[string]bool{}
		for j := 0; j < len(d.messages); j++ {
			m := d.messages[j]
			if m.key == "" {
				c.diagAt(m.span, "i18n message key must not be empty")
			}
			if seenKeys[m.key] {
				c.diagAt(m.span, "duplicate i18n message key \""+m.key+"\" in locale \""+loc+"\"")
			}
			seenKeys[m.key] = true
			if len(m.plural) > 0 {
				if _, ok := m.plural["other"]; !ok {
					c.diagAt(m.span, "plural message \""+m.key+"\" must define the \"other\" category")
				}
				for cat := range m.plural {
					if !isI18nPluralCategory(cat) {
						c.diagAt(m.span, "unknown plural category \""+cat+"\"")
					}
				}
			}
		}
	}
}

func Check(prog *Prog) DiagList {
	checkMu.Lock()
	defer checkMu.Unlock()
	var dl DiagList
	if prog == nil {
		return dl
	}
	c := newChecker(&dl)
	c.security = prog.security.Normalized()

	// Imports from "std".
	for i := 0; i < len(prog.imports); i++ {
		imp := prog.imports[i]
		for j := 0; j < len(imp.names); j++ {
			if imp.kind == "std" {
				if IsStdModule(imp.names[j]) {
					c.stdImported[imp.names[j]] = true
				} else {
					c.diagAt(imp.span, "unknown std module \""+imp.names[j]+"\"")
				}
			}
		}
	}

	c.checkI18n(prog.i18n)

	// User struct types first: names must be unique and resolve anywhere.
	structRegistry = map[string][]Field{}
	enumRegistry = map[string][]string{}
	for i := 0; i < len(prog.types); i++ {
		c.declareSym(prog.types[i].name, "type", prog.types[i].span)
		structRegistry[prog.types[i].name] = prog.types[i].fields
	}
	for i := 0; i < len(prog.enums); i++ {
		c.declareSym(prog.enums[i].name, "enum", prog.enums[i].span)
		enumRegistry[prog.enums[i].name] = prog.enums[i].members
	}
	// Eagerly validate every field type name.
	for i := 0; i < len(prog.types); i++ {
		td := prog.types[i]
		for j := 0; j < len(td.fields); j++ {
			fld := td.fields[j]
			_, err := parseSUIType(fld.typename)
			if err != "" {
				c.diagAt(fld.span, "struct "+td.name+" field \""+fld.name+"\": "+err)
			}
		}
	}

	// Symbol table pass: duplicate top-level names across types/states/funcs/
	// views are errors.
	for i := 0; i < len(prog.consts); i++ {
		c.declareSym(prog.consts[i].name, "const", prog.consts[i].span)
	}
	for i := 0; i < len(prog.computed); i++ {
		c.declareSym(prog.computed[i].name, "computed", prog.computed[i].span)
	}
	for i := 0; i < len(prog.states); i++ {
		c.declareSym(prog.states[i].name, "state", prog.states[i].span)
	}
	for i := 0; i < len(prog.funcs); i++ {
		c.declareSym(prog.funcs[i].name, "func", prog.funcs[i].span)
	}
	for i := 0; i < len(prog.views); i++ {
		c.declareSym(prog.views[i].name, "view", prog.views[i].span)
	}

	// Generated-name reservation: user declarations must not collide with the
	// JS runtime's setters/renderers/globals.
	for i := 0; i < len(prog.funcs); i++ {
		msg := reservedFuncName(prog.funcs[i].name, prog)
		if msg != "" {
			c.diagAt(prog.funcs[i].span, msg)
		}
	}
	for i := 0; i < len(prog.consts); i++ {
		msg := reservedStateName(prog.consts[i].name)
		if msg != "" {
			c.diagAt(prog.consts[i].span, msg)
		}
	}
	for i := 0; i < len(prog.computed); i++ {
		msg := reservedStateName(prog.computed[i].name)
		if msg != "" {
			c.diagAt(prog.computed[i].span, msg)
		}
	}
	for i := 0; i < len(prog.states); i++ {
		msg := reservedStateName(prog.states[i].name)
		if msg != "" {
			c.diagAt(prog.states[i].span, msg)
		}
	}

	// Register state placeholders (forward references resolve to "unknown").
	KnownStateNames = map[string]bool{}
	for i := 0; i < len(prog.states); i++ {
		c.states[prog.states[i].name] = &SType{kind: "unknown"}
		KnownStateNames[prog.states[i].name] = true
	}

	// Constants first: states, function defaults and other consts may use them.
	for i := 0; i < len(prog.consts); i++ {
		c.checkConst(&prog.consts[i])
	}
	// States.
	for i := 0; i < len(prog.states); i++ {
		c.checkState(&prog.states[i])
	}
	// Function signatures (all registered first, so calls resolve anywhere).
	for i := 0; i < len(prog.funcs); i++ {
		c.checkFuncSig(&prog.funcs[i])
	}
	// View-method signatures are also registered up front (names globally
	// unique; bodies are checked later inside their view's scope).
	for i := 0; i < len(prog.views); i++ {
		prevTP := c.curTypeParams
		c.curTypeParams = typeParamSet(prog.views[i].typeParams)
		for j := 0; j < len(prog.views[i].methods); j++ {
			md := &prog.views[i].methods[j]
			c.declareSym(md.name, "method", md.span)
			c.methodNames[md.name] = true
			c.checkFuncSig(md)
		}
		c.curTypeParams = prevTP
	}
	// Computed values may use states and function signatures (registered
	// above), plus earlier computed values.
	for i := 0; i < len(prog.computed); i++ {
		c.checkComputed(&prog.computed[i])
	}
	// Function bodies.
	for i := 0; i < len(prog.funcs); i++ {
		c.checkFuncBody(&prog.funcs[i])
	}

	// Component registry + invocation set + circular reference detection.
	for i := 0; i < len(prog.views); i++ {
		c.viewParams[prog.views[i].name] = prog.views[i].params
		c.viewTypeParams[prog.views[i].name] = prog.views[i].typeParams
	}
	viewInvoked = map[string]bool{}
	for i := 0; i < len(prog.views); i++ {
		var kids []string
		collectComponentNames(prog.views[i].body, &kids, c.viewParams)
		for j := 0; j < len(kids); j++ {
			viewInvoked[kids[j]] = true
		}
	}
	c.checkComponentCycles(prog)

	// Views.
	for i := 0; i < len(prog.views); i++ {
		c.checkView(&prog.views[i])
	}

	return dl
}

// register the standard module membership (from stdlib.go, re-exported)
// stdlib.go's IsStdModule is package visible.

// ---------------------------------------------------------------------------
// States
// ---------------------------------------------------------------------------

// isConstExpr reports whether an expression can initialize a const. Only
// literals, earlier consts, enum members and composite/operator expressions
// built from them are allowed, so a const never depends on runtime state.
func (c *Checker) isConstExpr(e *Expr) bool {
	if e == nil {
		return false
	}
	switch e.kind {
	case "int", "float", "bool", "null":
		return true
	case "str":
		for i := 0; i < len(e.parts); i++ {
			if e.parts[i].kind == "expr" && !c.isConstExpr(e.parts[i].expr) {
				return false
			}
		}
		return true
	case "ident":
		return c.consts[e.name] != nil
	case "member":
		_, _, ok := enumMemberOf(e)
		return ok
	case "array":
		for i := 0; i < len(e.elements); i++ {
			if i < len(e.spreads) && e.spreads[i] {
				return false
			}
			if !c.isConstExpr(e.elements[i]) {
				return false
			}
		}
		return true
	case "map":
		for i := 0; i < len(e.vals); i++ {
			if i < len(e.spreads) && e.spreads[i] {
				return false
			}
			if !c.isConstExpr(e.vals[i]) {
				return false
			}
		}
		return true
	case "structlit":
		for i := 0; i < len(e.vals); i++ {
			if !c.isConstExpr(e.vals[i]) {
				return false
			}
		}
		return true
	case "unary":
		return c.isConstExpr(e.operand)
	case "binary":
		return c.isConstExpr(e.l) && c.isConstExpr(e.r)
	case "cond":
		return c.isConstExpr(e.cond) && c.isConstExpr(e.l) && c.isConstExpr(e.r)
	case "index":
		return c.isConstExpr(e.obj) && c.isConstExpr(e.idx)
	}
	return false
}

func (c *Checker) checkComputed(cd *ComputedDecl) {
	it := c.inferExpr(cd.init)
	if cd.typename == "" {
		c.computed[cd.name] = it
		return
	}
	decl, err := parseSUIType(cd.typename)
	if err != "" {
		c.diagAt(cd.span, err)
		c.computed[cd.name] = it
		return
	}
	vt := c.valueTypeOf(cd.init, it)
	if !assignable(vt, decl) {
		c.diagAt(cd.span, "computed "+cd.name+" declared as "+typeStr(decl)+" but calculated with a value of type "+valueTypeStr(vt, decl))
	}
	c.computed[cd.name] = decl
}

func (c *Checker) checkConst(cd *ConstDecl) {
	if !c.isConstExpr(cd.init) {
		c.diagAtEx(cd.span, "const "+cd.name+" must be initialized with a compile-time value",
			"use literals, earlier consts, enum members, arrays/maps/structs and operators; runtime state and function calls are not allowed")
	}
	it := c.inferExpr(cd.init)
	if cd.typename == "" {
		c.consts[cd.name] = it
		return
	}
	decl, err := parseSUIType(cd.typename)
	if err != "" {
		c.diagAt(cd.span, err)
		c.consts[cd.name] = it
		return
	}
	vt := c.valueTypeOf(cd.init, it)
	if !assignable(vt, decl) {
		c.diagAt(cd.span, "const "+cd.name+" declared as "+typeStr(decl)+" but initialized with a value of type "+valueTypeStr(vt, decl))
	}
	c.consts[cd.name] = decl
}

func (c *Checker) checkState(sd *StateDecl) {
	if len(c.scopes) > 0 {
		c.diagAt(sd.span, "state "+sd.name+" must be declared at the top level")
	}
	it := c.inferExpr(sd.init)
	if sd.typename == "" {
		c.states[sd.name] = it
		c.declaredStates[sd.name] = it
		return
	}
	decl, err := parseSUIType(sd.typename)
	if err != "" {
		c.diagAt(sd.span, err)
		c.states[sd.name] = it
		return
	}
	vt := c.valueTypeOf(sd.init, it)
	if !assignable(vt, decl) {
		c.diagAt(sd.span, "state "+sd.name+" declared as "+typeStr(decl)+" but initialized with a value of type "+valueTypeStr(vt, decl))
	}
	c.states[sd.name] = decl
	c.declaredStates[sd.name] = decl
}

// ---------------------------------------------------------------------------
// Functions
// ---------------------------------------------------------------------------

func typeParamSet(names []string) map[string]bool {
	if len(names) == 0 {
		return nil
	}
	out := map[string]bool{}
	for i := 0; i < len(names); i++ {
		out[names[i]] = true
	}
	return out
}

func (c *Checker) checkFuncSig(fd *FuncDecl) {
	prevTP := c.curTypeParams
	vars := map[string]bool{}
	for name := range c.curTypeParams {
		vars[name] = true
	}
	for _, name := range fd.typeParams {
		vars[name] = true
	}
	c.curTypeParams = vars
	defer func() { c.curTypeParams = prevTP }()
	seenTP := map[string]bool{}
	for i := 0; i < len(fd.typeParams); i++ {
		if seenTP[fd.typeParams[i]] {
			c.diagAt(fd.span, "duplicate generic type parameter "+fd.typeParams[i])
		}
		seenTP[fd.typeParams[i]] = true
	}
	var params []SType
	var paramNames []string
	var defaults []*Expr
	sawOptional := false
	for i := 0; i < len(fd.params); i++ {
		pt, err := parseSUITypeWith(fd.params[i].typename, c.curTypeParams)
		if err != "" {
			if fd.params[i].typename == "" {
				c.diagAt(fd.params[i].span, "parameter "+fd.params[i].name+" requires a \":\" type annotation")
			} else {
				c.diagAt(fd.params[i].span, err)
			}
			pt = &SType{kind: "error"}
		}
		if fd.params[i].rest {
			if i != len(fd.params)-1 {
				c.diagAt(fd.params[i].span, "rest parameter \""+fd.params[i].name+"\" must be the last parameter")
			}
			if pt.kind != "array" && pt.kind != "error" {
				c.diagAt(fd.params[i].span, "rest parameter \""+fd.params[i].name+"\" must have an array type, got "+typeStr(pt))
			}
		}
		params = append(params, *pt)
		paramNames = append(paramNames, fd.params[i].name)
		// Optional parameters must be trailing: a required one after an
		// optional one is an error.
		if fd.params[i].dflt == nil && sawOptional {
			c.diagAt(fd.params[i].span, "parameter \""+fd.params[i].name+"\" follows an optional parameter; only trailing parameters may have defaults")
		}
		if fd.params[i].dflt != nil {
			sawOptional = true
		}
		defaults = append(defaults, fd.params[i].dflt)
	}
	rt, rerr := parseSUITypeWith(fd.returnType, c.curTypeParams)
	if rerr != "" {
		if fd.returnType == "" {
			c.diagAt(fd.span, "function "+fd.name+" must declare a return type with \"->\"")
		} else {
			c.diagAt(fd.span, rerr)
		}
		rt = NewSType("error")
	}
	var rets []*SType
	if len(fd.returnTypes) <= 1 {
		rets = append(rets, rt)
	} else {
		for i := 0; i < len(fd.returnTypes); i++ {
			t, terr := parseSUITypeWith(fd.returnTypes[i], c.curTypeParams)
			if terr != "" {
				c.diagAt(fd.span, "function "+fd.name+" return "+itoaStr(i+1)+": "+terr)
				t = NewSType("error")
			}
			rets = append(rets, t)
		}
	}
	rest := len(fd.params) > 0 && fd.params[len(fd.params)-1].rest
	c.sigs[fd.name] = &FnSig{name: fd.name, params: params, paramNames: paramNames, ret: rets[0], rets: rets, defaults: defaults, multi: len(rets) > 1, async: fd.async, rest: rest, typeParams: append([]string{}, fd.typeParams...)}
	// Type each default expression against its parameter.
	for i := 0; i < len(fd.params); i++ {
		if fd.params[i].dflt != nil && i < len(params) {
			dt := c.inferExpr(fd.params[i].dflt)
			dv := c.valueTypeOf(fd.params[i].dflt, dt)
			if dt != nil && dt.kind != "error" && dt.kind != "unknown" && !assignable(dv, &params[i]) {
				c.diagAt(fd.params[i].span, "default value of parameter \""+fd.params[i].name+"\" must be "+typeStr(&params[i])+", got "+valueTypeStr(dv, &params[i]))
			}
		}
	}
}

func (c *Checker) checkFuncBody(fd *FuncDecl) {
	sig := c.sigs[fd.name]

	prevRet := c.curRet
	prevRets := c.curRets
	prevAsync := c.curAsync
	prevSeen := c.seenReturn
	prevLoop := c.loopDepth
	prevFn := c.fnDepth
	prevInFn := c.inFn
	prevTP := c.curTypeParams
	vars := map[string]bool{}
	for name := range c.curTypeParams {
		vars[name] = true
	}
	for _, name := range fd.typeParams {
		vars[name] = true
	}
	c.curTypeParams = vars
	c.fnDepth = prevFn + 1
	c.inFn = true
	c.curRet = sig.ret
	c.curRets = sig.rets
	c.curAsync = fd.async
	c.seenReturn = false
	c.loopDepth = 0

	c.pushScope()
	for i := 0; i < len(fd.params); i++ {
		c.declareVar(fd.params[i].name, &sig.params[i], fd.params[i].span)
	}
	c.checkStmts(fd.body)
	c.popScope()

	if sig.ret != nil && sig.ret.kind != "null" && sig.ret.kind != "error" && sig.ret.kind != "unknown" && !c.seenReturn {
		c.diagAt(fd.span, "function "+fd.name+" with return type "+typeStr(sig.ret)+" must contain a return statement")
	}

	c.curRet = prevRet
	c.curRets = prevRets
	c.curAsync = prevAsync
	c.seenReturn = prevSeen
	c.loopDepth = prevLoop
	c.fnDepth = prevFn
	c.inFn = prevInFn
	c.curTypeParams = prevTP
}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

func (c *Checker) checkStmts(stmts []*Stmt) {
	for i := 0; i < len(stmts); i++ {
		c.checkStmt(stmts[i])
	}
}

func (c *Checker) checkStmt(s *Stmt) {
	if s.kind == "let" {
		c.checkLet(s)
	} else if s.kind == "assign" {
		c.checkAssign(s)
	} else if s.kind == "expr" {
		_ = c.inferExpr(s.init)
	} else if s.kind == "if" {
		c.checkIf(s)
	} else if s.kind == "switch" {
		c.checkSwitch(s)
	} else if s.kind == "loop" {
		c.checkLoop(s)
	} else if s.kind == "while" {
		c.checkWhile(s)
	} else if s.kind == "return" {
		c.checkReturn(s)
	} else if s.kind == "break" {
		if c.loopDepth == 0 && c.switchDepth == 0 {
			c.diagAt(s.span, "break outside a loop or switch")
		}
	} else if s.kind == "continue" {
		if c.loopDepth == 0 {
			c.diagAt(s.span, "continue outside a loop")
		}
	}
}

func (c *Checker) checkLet(s *Stmt) {
	if len(s.pattern) > 0 {
		c.checkDestructure(s)
		return
	}
	if c.states[s.name] != nil {
		c.diagAt(s.span, "let cannot shadow a state variable \""+s.name+"\"")
	}
	it := c.inferExpr(s.init)
	c.declareVar(s.name, it, s.span)
}

// checkDestructure types `let {a, b} = value` and `let [a, b] = values`.
// Object patterns take fields from a struct or values from a map; array
// patterns take the element type.
func (c *Checker) checkDestructure(s *Stmt) {
	it := c.inferExpr(s.init)
	seen := map[string]bool{}
	for i := 0; i < len(s.pattern); i++ {
		name := s.pattern[i]
		if seen[name] {
			c.diagAt(s.span, "duplicate name \""+name+"\" in the destructuring pattern")
		}
		seen[name] = true
		if c.states[name] != nil {
			c.diagAt(s.span, "let cannot shadow a state variable \""+name+"\"")
		}
		var vt *SType
		if s.patternKind == "array" {
			switch it.kind {
			case "array":
				vt = it.elem
			case "any", "unknown":
				vt = NewSType("any")
			case "error":
				vt = NewSType("error")
			default:
				c.diagAt(s.span, "array destructuring requires an array, got "+typeStr(it))
				vt = NewSType("error")
			}
		} else {
			switch it.kind {
			case "struct":
				vt = c.structFieldType(s.init, it.name, name)
			case "map":
				vt = it.val
			case "any", "unknown":
				vt = NewSType("any")
			case "error":
				vt = NewSType("error")
			default:
				c.diagAt(s.span, "object destructuring requires a struct or a map, got "+typeStr(it))
				vt = NewSType("error")
			}
		}
		c.declareVar(name, vt, s.span)
	}
}

// literalTypeOf returns the literal type of a compile-time literal
// expression. It is nil when the expression is not a single literal value;
// arrays and maps qualify when every element is itself a literal.
func literalTypeOf(e *Expr) *SType {
	if e == nil {
		return nil
	}
	switch e.kind {
	case "int":
		return &SType{kind: "literal", name: strconv.Itoa(e.i)}
	case "float":
		return &SType{kind: "literal", name: canonicalFloatName(e.f)}
	case "bool":
		if e.b {
			return &SType{kind: "literal", name: "true"}
		}
		return &SType{kind: "literal", name: "false"}
	case "str":
		if val, ok := literalString(e); ok {
			return &SType{kind: "literal", name: strconv.Quote(val)}
		}
	case "unary":
		if e.op == "neg" && e.operand != nil {
			switch e.operand.kind {
			case "int":
				return &SType{kind: "literal", name: strconv.Itoa(-e.operand.i)}
			case "float":
				return &SType{kind: "literal", name: canonicalFloatName(-e.operand.f)}
			}
		}
	case "array":
		if len(e.elements) == 0 {
			return nil
		}
		var parts []*SType
		for i := 0; i < len(e.elements); i++ {
			if i < len(e.spreads) && e.spreads[i] {
				return nil
			}
			lt := literalTypeOf(e.elements[i])
			if lt == nil {
				return nil
			}
			parts = append(parts, lt)
		}
		return &SType{kind: "array", elem: unionOf(parts...)}
	case "map":
		if len(e.vals) == 0 {
			return nil
		}
		var parts []*SType
		for i := 0; i < len(e.vals); i++ {
			if i < len(e.spreads) && e.spreads[i] {
				return nil
			}
			lt := literalTypeOf(e.vals[i])
			if lt == nil {
				return nil
			}
			parts = append(parts, lt)
		}
		return &SType{kind: "map", key: NewSType("string"), val: unionOf(parts...)}
	}
	return nil
}

// valueTypeOf reports the type a value contributes to an assignment-like
// check. Literal expressions contribute their literal type, and a state
// identifier contributes its declared type (the expression layer widens it so
// ordinary operators see the base type).
func (c *Checker) valueTypeOf(e *Expr, inferred *SType) *SType {
	if e != nil && e.kind == "ident" && c.declaredStates[e.name] != nil {
		return c.declaredStates[e.name]
	}
	if e != nil && e.kind == "ident" && c.computed[e.name] != nil {
		return c.computed[e.name]
	}
	if lit := literalTypeOf(e); lit != nil {
		return lit
	}
	return inferred
}

// declaredExprType returns the declared type of a state identifier. Switch
// subjects and `is` guards use it to keep literal unions visible even though
// ordinary expressions see the widened base type.
func (c *Checker) declaredExprType(e *Expr) *SType {
	if e != nil && e.kind == "ident" && c.declaredStates[e.name] != nil {
		return c.declaredStates[e.name]
	}
	return nil
}

// hasLiteralType reports whether a declared type contains a literal member.
// Diagnostics keep the exact literal for those destinations and fall back to
// the base type elsewhere, so ordinary messages stay "got string".
func hasLiteralType(t *SType) bool {
	if t == nil {
		return false
	}
	if t.kind == "literal" {
		return true
	}
	if t.kind == "union" {
		for i := 0; i < len(t.members); i++ {
			if hasLiteralType(t.members[i]) {
				return true
			}
		}
		return false
	}
	if t.kind == "array" {
		return hasLiteralType(t.elem)
	}
	if t.kind == "map" {
		return hasLiteralType(t.val)
	}
	return false
}

// valueTypeStr renders a source value for a diagnostic. When the destination
// accepts literal members, showing the literal value is more useful; otherwise
// the base type ("string", "int", ...) is shown.
func valueTypeStr(src, dst *SType) string {
	if hasLiteralType(dst) {
		return typeStr(src)
	}
	return typeStr(widenType(src))
}

// widenType maps literal types back to their base type. Operators and
// builtins use it so a value narrowed to a literal can still be concatenated
// or compared as its underlying string/int/bool.
func widenType(t *SType) *SType {
	if t == nil {
		return nil
	}
	if t.kind == "literal" {
		if strings.HasPrefix(t.name, "\"") {
			return NewSType("string")
		}
		if t.name == "true" || t.name == "false" {
			return NewSType("bool")
		}
		if strings.ContainsAny(t.name, ".eE") {
			return NewSType("float")
		}
		return NewSType("int")
	}
	if t.kind == "union" {
		var members []*SType
		for i := 0; i < len(t.members); i++ {
			members = append(members, widenType(t.members[i]))
		}
		return unionOf(members...)
	}
	if t.kind == "array" {
		return &SType{kind: "array", elem: widenType(t.elem)}
	}
	if t.kind == "map" {
		return &SType{kind: "map", key: t.key, val: widenType(t.val)}
	}
	return t
}

// literalTypeKey returns the switch-case key for a literal type, matching
// switchCaseKey so exhaustiveness can compare declared members with cases.
func literalTypeKey(t *SType) string {
	if t == nil || t.kind != "literal" {
		return ""
	}
	name := t.name
	if strings.HasPrefix(name, "\"") {
		val, err := strconv.Unquote(name)
		if err != nil {
			return ""
		}
		return "s:" + val
	}
	if name == "true" || name == "false" {
		return "b:" + name
	}
	if f, err := strconv.ParseFloat(name, 64); err == nil {
		if strings.ContainsAny(name, ".eE") {
			return "f:" + strconv.FormatFloat(f, 'g', -1, 64)
		}
		return "i:" + strconv.FormatInt(int64(f), 10)
	}
	return ""
}

// allLiteralUnion reports whether t is a union made only of literal types.
func allLiteralUnion(t *SType) bool {
	if t == nil || t.kind != "union" || len(t.members) == 0 {
		return false
	}
	for i := 0; i < len(t.members); i++ {
		if t.members[i].kind != "literal" {
			return false
		}
	}
	return true
}

// enumMemberOf reports whether e is Enum.Member and returns both parts.
func enumMemberOf(e *Expr) (string, string, bool) {
	if e == nil || e.kind != "member" || e.obj == nil || e.obj.kind != "ident" {
		return "", "", false
	}
	if enumRegistry[e.obj.name] == nil {
		return "", "", false
	}
	return e.obj.name, e.memb, true
}

// switchCaseKey returns a stable key for duplicate detection of literal and
// enum-member case values. Non-literal expressions return "".
func switchCaseKey(e *Expr) string {
	if e == nil {
		return ""
	}
	switch e.kind {
	case "int":
		return "i:" + strconv.Itoa(e.i)
	case "float":
		return "f:" + strconv.FormatFloat(e.f, 'g', -1, 64)
	case "bool":
		if e.b {
			return "b:true"
		}
		return "b:false"
	case "str":
		for i := 0; i < len(e.parts); i++ {
			if e.parts[i].kind != "text" {
				return ""
			}
		}
		return "s:" + partsContent(e.parts)
	case "member":
		if en, mem, ok := enumMemberOf(e); ok {
			return "e:" + en + "." + mem
		}
	case "unary":
		if e.op == "neg" && e.operand != nil {
			switch e.operand.kind {
			case "int":
				return "i:" + strconv.Itoa(-e.operand.i)
			case "float":
				return "f:" + strconv.FormatFloat(-e.operand.f, 'g', -1, 64)
			}
		}
	}
	return ""
}

// checkSwitch types the subject, case values and exhaustiveness. Switching on
// an enum requires every member unless a default exists.
func (c *Checker) checkSwitch(s *Stmt) {
	st := c.inferExpr(s.init)
	if d := c.declaredExprType(s.init); d != nil {
		st = d
	}
	switch st.kind {
	case "enum", "string", "int", "float", "bool", "union", "any", "unknown", "error":
	default:
		c.diagAt(s.span, "switch value must be an enum, string, number, bool or any, got "+typeStr(st))
	}
	seen := map[string]bool{}
	covered := map[string]bool{}
	c.switchDepth++
	for i := 0; i < len(s.cases); i++ {
		cse := s.cases[i]
		ct := c.inferExpr(cse.expr)
		cv := c.valueTypeOf(cse.expr, ct)
		if st.kind != "any" && st.kind != "unknown" && st.kind != "error" && ct.kind != "error" && ct.kind != "unknown" {
			if !assignable(cv, st) {
				c.diagAt(cse.expr.span, "case value must be "+typeStr(st)+", got "+valueTypeStr(cv, st))
			}
		}
		if st.kind == "enum" {
			if en, mem, ok := enumMemberOf(cse.expr); ok {
				if en != st.name {
					c.diagAt(cse.expr.span, "case value belongs to enum "+en+", but the switch is on "+st.name)
				} else {
					covered[mem] = true
				}
			} else if ct.kind != "error" && ct.kind != "unknown" {
				c.diagAt(cse.expr.span, "case value for enum "+st.name+" must be one of its members")
			}
		}
		if key := switchCaseKey(cse.expr); key != "" {
			if seen[key] {
				c.diagAt(cse.expr.span, "duplicate case value in switch")
			}
			seen[key] = true
			covered[key] = true
		}
		c.pushScope()
		c.checkStmts(cse.body)
		c.popScope()
	}
	if s.hasDefault {
		c.pushScope()
		c.checkStmts(s.default_)
		c.popScope()
	}
	c.switchDepth--
	if !s.hasDefault {
		if st.kind == "enum" {
			var missing []string
			for _, m := range enumRegistry[st.name] {
				if !covered[m] {
					missing = append(missing, m)
				}
			}
			if len(missing) > 0 {
				c.diagAt(s.span, "switch on "+st.name+" is not exhaustive; missing: "+strings.Join(missing, ", ")+" (add the cases or a default)")
			}
		} else if allLiteralUnion(st) {
			var missing []string
			for i := 0; i < len(st.members); i++ {
				if !covered[literalTypeKey(st.members[i])] {
					missing = append(missing, typeStr(st.members[i]))
				}
			}
			if len(missing) > 0 {
				c.diagAt(s.span, "switch on "+typeStr(st)+" is not exhaustive; missing: "+strings.Join(missing, ", ")+" (add the cases or a default)")
			}
		}
	}
}

func (c *Checker) checkIf(s *Stmt) {
	ct := c.inferExpr(s.cond)
	if ct.kind != "bool" && ct.kind != "error" && ct.kind != "unknown" {
		c.diagAt(s.span, "if condition must be a bool, got "+typeStr(ct))
	}
	c.pushScope()
	saved := c.applyTypeGuards(s.cond, true)
	c.checkStmts(s.then_)
	c.restoreTypeGuards(saved)
	c.popScope()
	c.pushScope()
	saved = c.applyTypeGuards(s.cond, false)
	c.checkStmts(s.else_)
	c.restoreTypeGuards(saved)
	c.popScope()
}

// typeGuard is one `value is Type` sub-expression in a condition.
type typeGuard struct {
	expr     *Expr
	positive bool
}

// savedTypeGuard restores a state type after a narrowed branch. Local
// variables use a scope shadow and need no explicit restore.
type savedTypeGuard struct {
	name string
	old  *SType
}

// collectTypeGuards extracts guards from the common boolean shapes:
// `v is T`, `not (v is T)`, `A and B` in the positive branch and `A or B` in
// the negative branch. Other shapes do not narrow.
func collectTypeGuards(e *Expr, positive bool, out *[]typeGuard) {
	if e == nil {
		return
	}
	if e.kind == "is" {
		*out = append(*out, typeGuard{expr: e, positive: positive})
		return
	}
	if e.kind == "unary" && e.op == "not" {
		collectTypeGuards(e.operand, !positive, out)
		return
	}
	if e.kind == "binary" {
		if (e.op == "and" && positive) || (e.op == "or" && !positive) {
			collectTypeGuards(e.l, positive, out)
			collectTypeGuards(e.r, positive, out)
		}
	}
}

func (c *Checker) applyTypeGuards(cond *Expr, positive bool) []savedTypeGuard {
	var guards []typeGuard
	collectTypeGuards(cond, positive, &guards)
	var saved []savedTypeGuard
	for i := 0; i < len(guards); i++ {
		g := guards[i]
		if g.expr == nil || g.expr.l == nil || g.expr.l.kind != "ident" {
			continue
		}
		name := g.expr.l.name
		target, err := parseSUIType(g.expr.name)
		if err != "" {
			continue
		}
		var current *SType
		isState := false
		if c.states[name] != nil {
			current = c.states[name]
			isState = true
		} else if v := c.lookupVar(name); v != nil {
			current = v.typ
		}
		if current == nil {
			continue
		}
		narrowed := narrowType(current, target, g.positive)
		if current.kind == narrowed.kind && typeEqual(current, narrowed) {
			continue
		}
		if isState {
			saved = append(saved, savedTypeGuard{name: name, old: current})
			c.states[name] = narrowed
		} else {
			c.declareVar(name, narrowed, g.expr.l.span)
		}
	}
	return saved
}

func (c *Checker) restoreTypeGuards(saved []savedTypeGuard) {
	for i := len(saved) - 1; i >= 0; i-- {
		c.states[saved[i].name] = saved[i].old
	}
}

func (c *Checker) checkLoop(s *Stmt) {
	var itemType *SType
	var item2Type *SType
	it := s.iterable
	if it != nil && it.kind == "range" {
		if it.lo != nil {
			lt := c.inferExpr(it.lo)
			if lt.kind != "int" && lt.kind != "error" && lt.kind != "unknown" {
				c.diagAt(it.lo.span, "loop range bounds must be ints, got "+typeStr(lt))
			}
		}
		if it.hi != nil {
			ht := c.inferExpr(it.hi)
			if ht.kind != "int" && ht.kind != "error" && ht.kind != "unknown" {
				c.diagAt(it.hi.span, "loop range bounds must be ints, got "+typeStr(ht))
			}
		}
		itemType = NewSType("int")
	} else {
		tt := c.inferExpr(it)
		if tt.kind == "array" {
			itemType = tt.elem
		} else if tt.kind == "map" {
			itemType = NewSType("string")
			item2Type = tt.val
		} else if tt.kind == "error" || tt.kind == "unknown" {
			itemType = NewSType("error")
		} else {
			c.diagAt(s.span, "loop iterable must be an array, a map or a range, got "+typeStr(tt))
			itemType = NewSType("error")
		}
	}
	if item2Type != nil && s.item2 == "" {
		c.diagAt(s.span, "iterating a map requires two names: loop k, v in <map>")
	} else if item2Type == nil && s.item2 != "" {
		c.diagAt(s.span, "two-name iteration requires a map, but this iterable is not a map")
	}
	c.pushScope()
	c.declareVar(s.item, itemType, s.span)
	if s.item2 != "" {
		if item2Type == nil {
			item2Type = NewSType("error")
		}
		c.declareVar(s.item2, item2Type, s.span)
	}
	c.loopDepth++
	c.checkStmts(s.body)
	c.loopDepth--
	c.popScope()
}

func (c *Checker) checkWhile(s *Stmt) {
	ct := c.inferExpr(s.cond)
	if ct.kind != "bool" && ct.kind != "error" && ct.kind != "unknown" {
		c.diagAt(s.span, "while condition must be a bool, got "+typeStr(ct))
	}
	c.pushScope()
	saved := c.applyTypeGuards(s.cond, true)
	c.loopDepth++
	c.checkStmts(s.body)
	c.loopDepth--
	c.restoreTypeGuards(saved)
	c.popScope()
}

func (c *Checker) checkReturn(s *Stmt) {
	if c.fnDepth == 0 {
		c.diagAt(s.span, "return outside a function")
		return
	}
	c.seenReturn = true
	multi := len(c.curRets) > 1
	if len(s.values) > 0 {
		if !multi {
			c.diagAt(s.span, "this function returns a single value; use \"return expr\"")
			for i := 0; i < len(s.values); i++ {
				_ = c.inferExpr(s.values[i])
			}
			return
		}
		if len(s.values) != len(c.curRets) {
			c.diagAt(s.span, "return provides "+itoaStr(len(s.values))+" values, expected "+itoaStr(len(c.curRets)))
		}
		for i := 0; i < len(s.values) && i < len(c.curRets); i++ {
			vt := c.inferExpr(s.values[i])
			if vt == nil || vt.kind == "error" || vt.kind == "unknown" {
				continue
			}
			vv := c.valueTypeOf(s.values[i], vt)
			if !assignable(vv, c.curRets[i]) {
				c.diagAt(s.span, "cannot return a value of type "+valueTypeStr(vv, c.curRets[i])+" for the "+ordinalNd(itoaStr(i+1))+" return value")
			}
		}
		return
	}
	if s.value == nil {
		if !multi && c.curRet.kind != "null" {
			c.diagAt(s.span, "bare return in a function whose return type is "+typeStr(c.curRet))
		}
		if multi {
			c.diagAt(s.span, "bare return in a multi-return function must provide all "+itoaStr(len(c.curRets))+" values")
		}
		return
	}
	if multi {
		c.diagAt(s.span, "this function returns "+itoaStr(len(c.curRets))+" values; use \"return a, b\"")
		_ = c.inferExpr(s.value)
		return
	}
	vt := c.inferExpr(s.value)
	if vt.kind == "error" || vt.kind == "unknown" {
		return
	}
	vv := c.valueTypeOf(s.value, vt)
	if !assignable(vv, c.curRet) {
		c.diagAt(s.span, "cannot return a value of type "+valueTypeStr(vv, c.curRet)+" from a function returning "+typeStr(c.curRet))
	}
}

// ---------------------------------------------------------------------------
// Assignments
// ---------------------------------------------------------------------------

func lvalueRootName(e *Expr) string {
	if e != nil && (e.kind == "member" || e.kind == "index") {
		return lvalueRootName(e.obj)
	}
	if e != nil && e.kind == "ident" {
		return e.name
	}
	return ""
}

func (c *Checker) checkAssign(s *Stmt) {
	if s.lvalue != nil {
		c.checkLvalueAssign(s)
		return
	}
	if len(s.targets) >= 2 {
		c.checkTupleAssign(s)
		return
	}
	if len(s.targets) == 0 {
		c.diagAt(s.span, "assignment requires a target name")
		_ = c.inferExpr(s.rhs)
		return
	}
	name := s.targets[0]
	var tgt *SType
	isState := false
	if c.states[name] != nil {
		tgt = c.states[name]
		if c.declaredStates[name] != nil {
			tgt = c.declaredStates[name]
		}
		isState = true
	} else {
		v := c.lookupVar(name)
		if v != nil {
			tgt = v.typ
		}
	}
	if tgt == nil {
		if c.symKinds[name] == "const" {
			c.diagAt(s.span, "cannot assign to const "+name)
		} else if c.symKinds[name] == "computed" {
			c.diagAt(s.span, "cannot assign to computed "+name)
		} else if c.symKinds[name] != "" {
			c.diagAt(s.span, "cannot assign to "+name)
		} else {
			hint := ""
			if cand := nearestName(name, c.allNames()); cand != "" {
				hint = "did you mean \"" + cand + "\"? "
			}
			c.diagAtEx(s.span, "unknown variable "+name, hint+"declare it with \"state\" or \"let\" first")
		}
		_ = c.inferExpr(s.rhs)
		return
	}
	if isState && !c.inFn {
		c.diagAtEx(s.span, "cannot assign to state \""+name+"\" outside a function", "move the assignment inside a function (for example an event handler)")
	}
	rt := c.inferExpr(s.rhs)
	vt := c.valueTypeOf(s.rhs, rt)
	if rt.kind != "error" && rt.kind != "unknown" && !assignable(vt, tgt) {
		hint := "change either the state or the assigned value to match the other"
		if isState {
			sp := c.symSpans[name]
			if sp.file != "" {
				hint = "the state was declared here: " + fmt.Sprintf("%v:%v", sp.file, sp.line)
			}
		}
		c.diagAtEx(s.span, "cannot assign a value of type "+valueTypeStr(vt, tgt)+" to \""+name+"\" of type "+typeStr(tgt), hint)
	}
}

func (c *Checker) checkLvalueAssign(s *Stmt) {
	root := lvalueRootName(s.lvalue)
	if root != "" && c.symKinds[root] == "const" {
		c.diagAt(s.span, "cannot mutate const "+root)
		return
	}
	if root != "" && c.symKinds[root] == "computed" {
		c.diagAt(s.span, "cannot mutate computed "+root)
		return
	}
	if root != "" && c.states[root] != nil && !c.inFn {
		c.diagAt(s.span, "cannot mutate state \""+root+"\" outside a function")
		return
	}
	c.checkReadonlyLvalue(s.lvalue)
	lt := c.inferExpr(s.lvalue)
	rt := c.inferExpr(s.rhs)
	vt := c.valueTypeOf(s.rhs, rt)
	if lt != nil && rt != nil && rt.kind != "error" && lt.kind != "error" && !assignable(vt, lt) {
		c.diagAt(s.span, "cannot assign a value of type "+valueTypeStr(vt, lt)+" to a value of type "+typeStr(lt))
	}
}

// checkReadonlyLvalue rejects writes through a struct field declared
// `readonly`. Nested fields and fields reached through indexes are checked by
// walking the object chain; the final assignment type check still runs.
func (c *Checker) checkReadonlyLvalue(e *Expr) {
	if e == nil {
		return
	}
	if e.kind == "member" {
		objT := c.inferExpr(e.obj)
		if objT != nil && objT.kind == "struct" {
			fields := structRegistry[objT.name]
			for i := 0; i < len(fields); i++ {
				if fields[i].name == e.memb && fields[i].readonly {
					c.diagAt(e.span, "cannot assign to readonly field \""+e.memb+"\" of "+objT.name)
					break
				}
			}
		}
		c.checkReadonlyLvalue(e.obj)
		return
	}
	if e.kind == "index" {
		c.checkReadonlyLvalue(e.obj)
	}
}

// ---------------------------------------------------------------------------
// Tuple assignment (std intrinsics and user multi-return functions)
// ---------------------------------------------------------------------------

// checkWsCallback validates fire-and-forget ws callback handler args.
func checkWsCallback(c *Checker, e *Expr, full string) {
	if (full != "ws.onMessage" && full != "ws.onOpen" && full != "ws.onClose" && full != "ws.onError") || len(e.args) != 2 {
		return
	}
	fn := e.args[1]
	if fn == nil || fn.kind != "ident" || c.sigs[fn.name] == nil {
		if full == "ws.onMessage" {
			c.diagAt(e.span, "ws.onMessage second argument must be a declared function that receives the message")
		} else {
			c.diagAt(e.span, full+" second argument must be a declared function")
		}
	}
}

func checkWsOnMessage(c *Checker, e *Expr, full string) {
	checkWsCallback(c, e, full)
}

func isStdCallShape(e *Expr) bool {
	if e == nil || e.kind != "call" {
		return false
	}
	t := e.target
	return t != nil && t.kind == "member" && t.obj != nil && t.obj.kind == "ident" && IsStdModule(t.obj.name)
}

func isUserMultiCall(c *Checker, e *Expr) bool {
	if e == nil || e.kind != "call" {
		return false
	}
	t := e.target
	if t == nil || t.kind != "ident" {
		return false
	}
	sig := c.sigs[t.name]
	return sig != nil && sig.multi
}

func userMultiCallRets(c *Checker, e *Expr) []*SType {
	sig := c.sigs[e.target.name]
	if sig == nil {
		return []*SType{}
	}
	var rets []*SType
	if len(sig.typeParams) > 0 {
		rets = c.checkGenericCall(e.target.name, sig, e.args, e.span)
	} else {
		c.checkArgs(e.target.name, sig.params, e.args, sig.defaults, sig.rest, e.span)
		rets = sig.rets
	}
	if sig.async && !c.curAsync {
		c.diagAt(e.span, "async function \""+e.target.name+"\" cannot be called from a non-async function")
	}
	return rets
}

// stdTypeName maps a stdlib parameter/return type name to its SType.
func stdTypeName(s string) *SType {
	if s == "any" {
		return NewSType("any")
	}
	if s == "null" {
		return NewSType("null")
	}
	if s == "int" {
		return NewSType("int")
	}
	if s == "float" {
		return NewSType("float")
	}
	if s == "bool" {
		return NewSType("bool")
	}
	if s == "string" {
		return NewSType("string")
	}
	if s == "array[string]" {
		return &SType{kind: "array", elem: NewSType("string")}
	}
	if s == "array[int]" {
		return &SType{kind: "array", elem: NewSType("int")}
	}
	if s == "array[array[string]]" {
		return &SType{kind: "array", elem: &SType{kind: "array", elem: NewSType("string")}}
	}
	if s == "array[any]" {
		return &SType{kind: "array", elem: NewSType("any")}
	}
	if s == "map[string]string" {
		return &SType{kind: "map", key: NewSType("string"), val: NewSType("string")}
	}
	if s == "map[string]any" {
		return &SType{kind: "map", key: NewSType("string"), val: NewSType("any")}
	}
	if s == "xmlNode" || s == "xmlnode" {
		return NewSType("xmlnode")
	}
	if s == "xmlDocument" || s == "xmldocument" {
		return NewSType("xmldocument")
	}
	if s == "array[xmlNode]" || s == "array[xmlnode]" {
		return &SType{kind: "array", elem: NewSType("xmlnode")}
	}
	return NewSType("error")
}

// safeURLVariant maps an unsafe std URL call to its guarded counterpart.
func safeURLVariant(full string) string {
	switch full {
	case "http.get":
		return "http.getSafe"
	case "http.getStatus":
		return "http.getStatusSafe"
	case "http.post":
		return "http.postSafe"
	case "http.postFiles":
		return "http.postFilesSafe"
	case "http.getWithHeaders":
		return "http.getWithHeadersSafe"
	case "http.postWithHeaders":
		return "http.postWithHeadersSafe"
	case "http.request":
		return "http.requestSafe"
	case "http.upload":
		return "http.uploadSafe"
	case "http.uploadWithHeaders":
		return "http.uploadWithHeadersSafe"
	case "ws.connect":
		return "ws.connectSafe"
	case "ws.connectWithOptions":
		return "ws.connectWithOptionsSafe"
	}
	return full + "Safe"
}

// literalString returns the value of a string literal with no interpolation.
func literalString(e *Expr) (string, bool) {
	if e == nil || e.kind != "str" {
		return "", false
	}
	var sb strings.Builder
	for i := 0; i < len(e.parts); i++ {
		if e.parts[i].kind != "text" {
			return "", false
		}
		sb.WriteString(e.parts[i].text)
	}
	return sb.String(), true
}

// checkStdURLSecurity enforces the SSRF policy at compile time:
//   - http.get/getStatus/post and ws.connect accept only literal URLs;
//   - dynamic URLs must use the *Safe variants;
//   - the allow-list of a *Safe call must be a developer-written literal.
func (c *Checker) checkStdURLSecurity(e *Expr, f *StdFunc) {
	if e == nil || f == nil {
		return
	}
	full := f.mod + "." + f.name
	switch full {
	case "http.get", "http.getStatus", "http.post", "http.postFiles", "http.getWithHeaders", "http.postWithHeaders", "http.upload", "http.uploadWithHeaders", "ws.connect", "ws.connectWithOptions":
		if len(e.args) == 0 {
			return
		}
		ws := full == "ws.connect" || full == "ws.connectWithOptions"
		if s, ok := literalString(e.args[0]); ok {
			c.checkLiteralURL(e.args[0], s, ws)
		} else {
			msg := "dynamic URL: use " + safeURLVariant(full) + " with an allow-list"
			if ws {
				msg = "dynamic WebSocket URL: use " + safeURLVariant(full) + " with an allow-list"
			}
			c.securityDiag(e.args[0].span, c.security.URLDynamic, msg)
		}
	case "http.request":
		if len(e.args) < 2 {
			return
		}
		if s, ok := literalString(e.args[1]); ok {
			c.checkLiteralURL(e.args[1], s, false)
		} else {
			msg := "dynamic URL: use http.requestSafe with an allow-list"
			c.securityDiag(e.args[1].span, c.security.URLDynamic, msg)
		}
	case "http.getSafe", "http.getStatusSafe", "http.postSafe", "http.postFilesSafe", "http.getWithHeadersSafe", "http.postWithHeadersSafe", "http.requestSafe", "http.uploadSafe", "http.uploadWithHeadersSafe", "xml.loadUrlSafe", "ws.connectSafe", "ws.connectWithOptionsSafe":
		idx := 1
		switch full {
		case "http.postSafe", "http.getWithHeadersSafe", "ws.connectWithOptionsSafe":
			idx = 2
		case "http.postWithHeadersSafe":
			idx = 3
		case "http.uploadSafe":
			idx = 4
		case "http.uploadWithHeadersSafe", "http.requestSafe":
			idx = 5
		}
		if len(e.args) > idx {
			if _, ok := literalString(e.args[idx]); !ok {
				c.diagAt(e.args[idx].span, "allow-list must be a literal string")
			}
		}
	}
}

func (c *Checker) checkLiteralURL(e *Expr, s string, ws bool) {
	var probe strings.Builder
	for _, r := range s {
		if r > ' ' {
			probe.WriteRune(r)
		}
	}
	low := strings.ToLower(probe.String())
	if strings.HasPrefix(low, "//") {
		if ws {
			c.securityDiag(e.span, c.security.StaticURLSchemes, "ws.connect URL must use ws:// or wss://")
		}
		return
	}
	i := strings.Index(low, ":")
	if i < 0 {
		if ws {
			c.securityDiag(e.span, c.security.StaticURLSchemes, "ws.connect URL must use ws:// or wss://")
		}
		return
	}
	scheme := low[:i]
	if ws {
		if scheme != "ws" && scheme != "wss" {
			c.securityDiag(e.span, c.security.StaticURLSchemes, "ws.connect URL must use ws:// or wss://")
		}
		return
	}
	if scheme != "http" && scheme != "https" {
		c.securityDiag(e.span, c.security.StaticURLSchemes, "URL scheme must be http or https")
	}
}

// stdCallIsAsync reports whether a std-intrinsic call targets an async
// intrinsic. Core modules such as `table` are synchronous DOM helpers, so they
// are callable from ordinary (non-async) functions.
func stdCallIsAsync(e *Expr) bool {
	if e == nil || e.kind != "call" || e.target == nil || e.target.kind != "member" {
		return false
	}
	if e.target.obj == nil {
		return false
	}
	if f, ok := StdLookup(e.target.obj.name + "." + e.target.memb); ok && f != nil {
		return f.async
	}
	return false
}

// stdCall validates a std-intrinsic call and returns the two value types.
func stdCall(c *Checker, e *Expr) (*SType, *SType) {
	if !isStdCallShape(e) {
		return nil, nil
	}
	target := e.target
	modName := target.obj.name
	full := modName + "." + target.memb
	var f *StdFunc
	ok := false
	if c.stdImported[modName] || IsCoreModule(modName) {
		f, ok = StdLookup(full)
		if !ok {
			c.diagAt(e.span, "unknown std function \""+full+"\"")
		}
	} else {
		c.diagAt(e.span, "std module \""+modName+"\" is not imported (add import { "+modName+" } from \"std\")")
	}
	if !ok || f == nil {
		return nil, nil
	}
	c.checkStdURLSecurity(e, f)
	minArgs := len(f.params) - f.optional
	if minArgs < 0 {
		minArgs = 0
	}
	if len(e.args) < minArgs || len(e.args) > len(f.params) {
		expect := fmt.Sprintf("%v", len(f.params))
		if minArgs != len(f.params) {
			expect = fmt.Sprintf("%v to %v", minArgs, len(f.params))
		}
		c.diagAt(e.span, "std function \""+full+"\" expects "+expect+" arguments, got "+fmt.Sprintf("%v", len(e.args)))
		for i := 0; i < len(e.args); i++ {
			if e.args[i] != nil {
				_ = c.inferExpr(e.args[i])
			}
		}
		return NewSType("error"), NewSType("error")
	}
	for i := 0; i < len(e.args); i++ {
		a := e.args[i]
		pt := stdTypeName(f.params[i])
		at := c.inferExpr(a)
		if at.kind == "error" || at.kind == "unknown" {
			continue
		}
		if !assignable(at, pt) {
			c.diagAt(a.span, "argument "+fmt.Sprintf("%v", i+1)+" of "+full+" must be "+typeStr(pt)+", got "+typeStr(at))
		}
	}
	if (full == "ws.onMessage" || full == "ws.onOpen" || full == "ws.onClose" || full == "ws.onError") && len(e.args) == 2 {
		checkWsCallback(c, e, full)
	}
	return stdTypeName(f.ret0), stdTypeName(f.ret1)
}

func (c *Checker) checkTupleAssign(s *Stmt) {
	if len(s.targets) > 2 {
		c.diagAt(s.span, "tuple assignment supports exactly two targets")
	}
	rhs := s.rhs
	var b0 *SType
	var b1 *SType
	if isStdCallShape(rhs) {
		t0, t1 := stdCall(c, rhs)
		if !c.curAsync && stdCallIsAsync(rhs) {
			c.diagAt(rhs.span, "std I/O calls must be inside an async function")
		}
		b0 = t0
		b1 = t1
	} else if isUserMultiCall(c, rhs) {
		rets := userMultiCallRets(c, rhs)
		if len(rets) >= 2 {
			b0 = rets[0]
			b1 = rets[1]
		}
	} else {
		c.diagAt(s.span, "expected a two-value call on the right-hand side (a std intrinsic or a function declared with \"-> (A, B)\")")
		if rhs != nil {
			_ = c.inferExpr(rhs)
		}
		return
	}
	if b0 == nil {
		b0 = NewSType("error")
	}
	if b1 == nil {
		b1 = NewSType("error")
	}
	for i := 0; i < len(s.targets) && i < 2; i++ {
		name := s.targets[i]
		if c.states[name] != nil || c.lookupVar(name) != nil || c.symKinds[name] != "" {
			c.diagAt(s.span, "tuple target \""+name+"\" must be a fresh name")
		}
	}
	c.declareVar(s.targets[0], b0, s.span)
	if len(s.targets) >= 2 {
		c.declareVar(s.targets[1], b1, s.span)
	}
}

// orderNamedArgs rewrites a call with named arguments into declaration order.
// Positional and named arguments cannot be mixed. Missing parameters before a
// named one are filled from their defaults; trailing optional parameters may
// still be omitted.
// hasNamedArgs reports whether a call has at least one named argument.
// Positional calls carry a parallel slice of empty strings.
func hasNamedArgs(names []string) bool {
	for i := 0; i < len(names); i++ {
		if names[i] != "" {
			return true
		}
	}
	return false
}

func (c *Checker) orderNamedArgs(e *Expr, sig *FnSig, fname string) []*Expr {
	if !hasNamedArgs(e.argNames) {
		return e.args
	}
	if sig.rest {
		c.diagAt(e.span, "named arguments cannot be used with the rest parameter of "+fname)
		return e.args
	}
	for i := 0; i < len(e.argNames); i++ {
		if e.argNames[i] == "" {
			c.diagAt(e.span, "cannot mix positional and named arguments in a call to "+fname)
			return e.args
		}
	}
	pos := map[string]int{}
	for i := 0; i < len(sig.paramNames); i++ {
		pos[sig.paramNames[i]] = i
	}
	ordered := make([]*Expr, len(sig.params))
	seen := map[string]bool{}
	max := -1
	for i := 0; i < len(e.argNames); i++ {
		name := e.argNames[i]
		idx, ok := pos[name]
		if !ok {
			c.diagAt(e.args[i].span, "unknown named argument \""+name+"\" for "+fname)
			continue
		}
		if seen[name] {
			c.diagAt(e.args[i].span, "duplicate named argument \""+name+"\" for "+fname)
			continue
		}
		seen[name] = true
		ordered[idx] = e.args[i]
		if idx > max {
			max = idx
		}
	}
	for i := 0; i < max; i++ {
		if ordered[i] == nil {
			if i < len(sig.defaults) && sig.defaults[i] != nil {
				ordered[i] = sig.defaults[i]
			} else {
				c.diagAt(e.span, "missing argument for parameter \""+sig.paramNames[i]+"\" of "+fname)
				return e.args
			}
		}
	}
	last := len(ordered) - 1
	for last >= 0 && ordered[last] == nil {
		last--
	}
	return ordered[:last+1]
}

// checkArgs validates provided arguments against the parameter types; missing
// trailing arguments are permitted when those parameters carry defaults.
func (c *Checker) checkArgs(fname string, params []SType, args []*Expr, defs []*Expr, rest bool, sp Span) {
	types := make([]*SType, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] != nil {
			types[i] = c.inferExpr(args[i])
		}
	}
	c.checkArgsTyped(fname, params, args, types, defs, rest, sp)
}

// checkArgsTyped validates already-inferred argument types. Generic calls
// infer the types once for unification and reuse them here.
func (c *Checker) checkArgsTyped(fname string, params []SType, args []*Expr, types []*SType, defs []*Expr, rest bool, sp Span) {
	fixed := len(params)
	if rest {
		fixed--
	}
	if !rest && len(args) > len(params) {
		c.diagAt(sp, "function "+fname+" expects "+fmt.Sprintf("%v", len(params))+" arguments, got "+fmt.Sprintf("%v", len(args)))
		return
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == nil {
			continue
		}
		at := types[i]
		if at == nil || at.kind == "error" || at.kind == "unknown" {
			continue
		}
		av := c.valueTypeOf(a, at)
		want := &params[0]
		if i < fixed {
			want = &params[i]
		} else if rest {
			want = params[len(params)-1].elem
		} else {
			return
		}
		if !assignable(av, want) {
			c.diagAt(a.span, "argument "+fmt.Sprintf("%v", i+1)+" of "+fname+" must be "+typeStr(want)+", got "+valueTypeStr(av, want))
		}
	}
	// Missing trailing arguments must have default values; a rest parameter
	// may be empty.
	for i := len(args); i < fixed; i++ {
		hasDefault := i < len(defs) && defs[i] != nil
		if !hasDefault {
			c.diagAt(sp, "missing argument for parameter "+fmt.Sprintf("%v", i+1)+" of "+fname+" (no default value)")
		}
	}
}

// substituteType replaces generic type variables with their inferred types.
func substituteType(t *SType, subst map[string]*SType) *SType {
	if t == nil {
		return nil
	}
	if t.kind == "typevar" {
		if bound := subst[t.name]; bound != nil {
			return bound
		}
		return t
	}
	if t.kind == "array" {
		return &SType{kind: "array", elem: substituteType(t.elem, subst)}
	}
	if t.kind == "map" {
		return &SType{kind: "map", key: substituteType(t.key, subst), val: substituteType(t.val, subst)}
	}
	if t.kind == "union" {
		var members []*SType
		for i := 0; i < len(t.members); i++ {
			members = append(members, substituteType(t.members[i], subst))
		}
		return unionOf(members...)
	}
	return t
}

// unifyType binds generic type variables from an argument type. It returns
// false when a bound parameter conflicts with the argument.
func (c *Checker) unifyType(param, arg *SType, subst map[string]*SType) bool {
	if param == nil || arg == nil {
		return true
	}
	if arg.kind == "error" || arg.kind == "unknown" || arg.kind == "any" {
		return true
	}
	if param.kind == "typevar" {
		if arg.kind == "typevar" && arg.name == param.name {
			return true
		}
		if bound := subst[param.name]; bound != nil {
			return assignable(arg, bound)
		}
		subst[param.name] = widenType(arg)
		return true
	}
	if param.kind == "array" && arg.kind == "array" {
		return c.unifyType(param.elem, arg.elem, subst)
	}
	if param.kind == "map" && arg.kind == "map" {
		return c.unifyType(param.val, arg.val, subst)
	}
	if param.kind == "union" && arg.kind == "union" {
		for i := 0; i < len(arg.members); i++ {
			if !assignable(arg.members[i], param) {
				return false
			}
		}
		return true
	}
	return assignable(arg, param)
}

// checkGenericCall infers a generic function's type arguments from the call,
// substitutes them into the signature, checks the arguments and returns the
// substituted return types.
func (c *Checker) checkGenericCall(name string, sig *FnSig, args []*Expr, sp Span) []*SType {
	types := make([]*SType, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] != nil {
			types[i] = c.inferExpr(args[i])
		}
	}
	subst := map[string]*SType{}
	fixed := len(sig.params)
	if sig.rest {
		fixed--
	}
	for i := 0; i < len(args); i++ {
		var param *SType
		if i < fixed {
			param = &sig.params[i]
		} else if sig.rest && len(sig.params) > 0 {
			param = sig.params[len(sig.params)-1].elem
		} else {
			break
		}
		if types[i] == nil || types[i].kind == "error" || types[i].kind == "unknown" {
			continue
		}
		// A false result means a bound type parameter conflicts with this
		// argument; checkArgsTyped below reports it against the substituted type.
		_ = c.unifyType(param, types[i], subst)
	}
	var missing []string
	for _, tp := range sig.typeParams {
		if subst[tp] == nil {
			missing = append(missing, tp)
		}
	}
	if len(missing) > 0 {
		c.diagAt(sp, "cannot infer generic type parameter "+strings.Join(missing, ", ")+" for "+name+" (use it in a parameter type)")
		for _, tp := range missing {
			subst[tp] = NewSType("any")
		}
	}
	params := make([]SType, len(sig.params))
	for i := 0; i < len(sig.params); i++ {
		params[i] = *substituteType(&sig.params[i], subst)
	}
	c.checkArgsTyped(name, params, args, types, sig.defaults, sig.rest, sp)
	rets := make([]*SType, len(sig.rets))
	for i := 0; i < len(sig.rets); i++ {
		rets[i] = substituteType(sig.rets[i], subst)
	}
	return rets
}

// A shim so pre-existing call sites with the old 4-argument signature still
// compile; new code passes defaults too.
func (c *Checker) checkArgs4(fname string, params []SType, args []*Expr, sp Span) {
	c.checkArgs(fname, params, args, nil, false, sp)
}

// ---------------------------------------------------------------------------
// Expressions
// ---------------------------------------------------------------------------

func (c *Checker) inferExpr(e *Expr) *SType {
	if e == nil {
		return NewSType("error")
	}
	k := e.kind
	if k == "int" {
		return NewSType("int")
	}
	if k == "float" {
		return NewSType("float")
	}
	if k == "bool" {
		return NewSType("bool")
	}
	if k == "null" {
		return NewSType("null")
	}
	if k == "str" {
		for i := 0; i < len(e.parts); i++ {
			if e.parts[i].kind == "expr" && e.parts[i].expr != nil {
				_ = c.inferExpr(e.parts[i].expr)
			}
		}
		return NewSType("string")
	}
	if k == "range" {
		c.diagAt(e.span, "range is only valid as a loop iterable")
		return NewSType("range")
	}
	if k == "array" {
		return c.inferArray(e)
	}
	if k == "map" {
		return c.inferMap(e)
	}
	if k == "structlit" {
		return c.inferStructLit(e)
	}
	if k == "ident" {
		return c.inferIdent(e)
	}
	if k == "member" {
		return c.inferMember(e)
	}
	if k == "call" {
		return c.inferCall(e)
	}
	if k == "index" {
		return c.inferIndex(e)
	}
	if k == "unary" {
		return c.inferUnary(e)
	}
	if k == "is" {
		lt := c.inferExpr(e.l)
		if d := c.declaredExprType(e.l); d != nil {
			lt = d
		}
		target, err := parseSUIType(e.name)
		if err != "" {
			c.diagAt(e.span, "invalid type guard: "+err)
			return NewSType("bool")
		}
		if target.kind == "struct" || target.kind == "map" {
			c.diagAt(e.span, "type guard for "+typeStr(target)+" is not supported at runtime; use an enum discriminator field or check null/array instead")
			return NewSType("bool")
		}
		if !overlaps(lt, target) {
			c.diagAt(e.span, "type guard "+typeStr(lt)+" is "+typeStr(target)+" can never succeed")
		}
		return NewSType("bool")
	}
	if k == "binary" {
		return c.inferBinary(e)
	}
	if k == "cond" {
		ct := c.inferExpr(e.cond)
		if ct.kind != "bool" && ct.kind != "error" && ct.kind != "unknown" {
			c.diagAt(e.cond.span, "condition of a conditional expression must be bool, got "+typeStr(ct))
		}
		lt := c.inferExpr(e.l)
		if e.r == nil {
			return lt
		}
		rt := c.inferExpr(e.r)
		if lt.kind == "error" || lt.kind == "unknown" {
			return rt
		}
		if rt.kind == "error" || rt.kind == "unknown" {
			return lt
		}
		if !typeEqual(lt, rt) {
			return unionOf(lt, rt)
		}
		return lt
	}
	return NewSType("error")
}

func (c *Checker) inferIdent(e *Expr) *SType {
	name := e.name
	if c.states[name] != nil {
		// Literal restrictions stay on the declaration; expression code uses
		// the base type so operators and builtins keep working.
		return widenType(c.states[name])
	}
	if c.consts[name] != nil {
		return widenType(c.consts[name])
	}
	if c.computed[name] != nil {
		return widenType(c.computed[name])
	}
	v := c.lookupVar(name)
	if v != nil {
		return v.typ
	}
	if enumRegistry[name] != nil {
		c.diagAt(e.span, "enum \""+name+"\" cannot be used as a value; use "+name+".Member")
		return NewSType("error")
	}
	if c.symKinds[name] == "func" {
		return NewSType("func")
	}
	if IsStdModule(name) {
		if !c.stdImported[name] && !IsCoreModule(name) {
			c.diagAt(e.span, "std module \""+name+"\" is not imported (add import { "+name+" } from \"std\")")
		}
		return NewSType("func")
	}
	hint := ""
	if cand := nearestName(name, c.allNames()); cand != "" {
		hint = "did you mean \"" + cand + "\"? "
	}
	c.diagAtEx(e.span, "unknown variable "+name, hint+"declare the variable with \"state\" or \"let\", or fix the typo")
	return NewSType("error")
}

func (c *Checker) inferArray(e *Expr) *SType {
	if len(e.elements) == 0 {
		return &SType{kind: "array", elem: NewSType("unknown")}
	}
	var common *SType
	for i := 0; i < len(e.elements); i++ {
		et := c.inferExpr(e.elements[i])
		if i < len(e.spreads) && e.spreads[i] {
			if et.kind == "array" {
				et = et.elem
			} else if et.kind == "any" || et.kind == "unknown" || et.kind == "error" {
				// keep the sentinel and let the common-type pass ignore it
			} else {
				c.diagAt(e.elements[i].span, "spread in an array literal requires an array, got "+typeStr(et))
				et = NewSType("error")
			}
		}
		if et.kind == "error" || et.kind == "unknown" {
			continue
		}
		if common == nil {
			common = et
			continue
		}
		if !typeEqual(common, et) {
			common = unionOf(common, et)
		}
	}
	if common == nil {
		common = NewSType("null")
	}
	return &SType{kind: "array", elem: common}
}

func (c *Checker) inferMap(e *Expr) *SType {
	var common *SType
	for i := 0; i < len(e.vals); i++ {
		vt := c.inferExpr(e.vals[i])
		if i < len(e.spreads) && e.spreads[i] {
			if vt.kind == "map" {
				vt = vt.val
			} else if vt.kind == "any" || vt.kind == "unknown" || vt.kind == "error" {
				// keep the sentinel and let the common-type pass ignore it
			} else {
				c.diagAt(e.vals[i].span, "spread in a map literal requires a map, got "+typeStr(vt))
				vt = NewSType("error")
			}
		}
		if vt.kind == "error" || vt.kind == "unknown" {
			continue
		}
		if common == nil {
			common = vt
			continue
		}
		if !typeEqual(common, vt) {
			common = unionOf(common, vt)
		}
	}
	if common == nil {
		common = NewSType("unknown")
	}
	return &SType{kind: "map", key: NewSType("string"), val: common}
}

func structFieldNames(fields []Field) []string {
	var out []string
	for i := 0; i < len(fields); i++ {
		out = append(out, fields[i].name)
	}
	return out
}

func (c *Checker) structFieldType(e *Expr, typeName string, field string) *SType {
	fields := structRegistry[typeName]
	if fields == nil {
		c.diagAt(e.span, "unknown struct type "+typeName)
		return NewSType("error")
	}
	for i := 0; i < len(fields); i++ {
		if fields[i].name == field {
			if typ, serr := parseSUIType(fields[i].typename); serr == "" {
				return typ
			}
			return NewSType("error")
		}
	}
	hint := ""
	if cand := nearestName(field, structFieldNames(fields)); cand != "" {
		hint = "did you mean \"" + cand + "\"? "
	}
	if hint == "" {
		c.diagAt(e.span, "unknown field \""+field+"\" on struct "+typeName+" (fields: "+strings.Join(structFieldNames(fields), ", ")+")")
	} else {
		c.diagAtEx(e.span, "unknown field \""+field+"\" on struct "+typeName, hint)
	}
	return NewSType("error")
}

func (c *Checker) inferStructLit(e *Expr) *SType {
	typeName := e.name
	fields := structRegistry[typeName]
	if fields == nil {
		c.diagAt(e.span, "unknown struct type \""+typeName+"\" (declare it with \"type "+typeName+" struct { ... }\")")
		for i := 0; i < len(e.vals); i++ {
			_ = c.inferExpr(e.vals[i])
		}
		return NewSType("error")
	}
	seen := map[string]bool{}
	for i := 0; i < len(e.keys); i++ {
		key := e.keys[i]
		if seen[key] {
			c.diagAt(e.span, "duplicate field \""+key+"\" in "+typeName+" literal")
		}
		seen[key] = true
		var field *Field
		for j := 0; j < len(fields); j++ {
			if fields[j].name == key {
				field = &fields[j]
				break
			}
		}
		if field == nil {
			hint := ""
			if cand := nearestName(key, structFieldNames(fields)); cand != "" {
				hint = "did you mean \"" + cand + "\"? "
			}
			c.diagAtEx(e.span, "unknown field \""+key+"\" in "+typeName+" literal", hint)
			_ = c.inferExpr(e.vals[i])
			continue
		}
		var fexp *SType
		if ft, ferr := parseSUIType(field.typename); ferr == "" {
			fexp = ft
		}
		vt := c.inferExpr(e.vals[i])
		vv := c.valueTypeOf(e.vals[i], vt)
		if fexp != nil && !assignable(vv, fexp) {
			c.diagAt(e.span, "field \""+key+"\" of "+typeName+" expects "+typeStr(fexp)+", got "+valueTypeStr(vv, fexp))
		}
	}
	for j := 0; j < len(fields); j++ {
		if !seen[fields[j].name] {
			c.diagAt(e.span, "missing field \""+fields[j].name+"\" in "+typeName+" literal")
		}
	}
	return &SType{kind: "struct", name: typeName}
}

func (c *Checker) inferMember(e *Expr) *SType {
	obj := e.obj
	if obj != nil && obj.kind == "ident" && enumRegistry[obj.name] != nil {
		members := enumRegistry[obj.name]
		for i := 0; i < len(members); i++ {
			if members[i] == e.memb {
				return &SType{kind: "enum", name: obj.name}
			}
		}
		hint := ""
		if cand := nearestName(e.memb, members); cand != "" {
			hint = "did you mean \"" + cand + "\"? "
		}
		c.diagAtEx(e.span, "unknown enum member \""+e.memb+"\" on "+obj.name, hint+"members: "+strings.Join(members, ", "))
		return NewSType("error")
	}
	if obj != nil && obj.kind == "ident" && IsStdModule(obj.name) {
		if c.stdImported[obj.name] || IsCoreModule(obj.name) {
			full := obj.name + "." + e.memb
			_, ok := StdLookup(full)
			if !ok {
				c.diagAt(e.span, "unknown std function \""+full+"\"")
			}
		} else {
			c.diagAt(e.span, "std module \""+obj.name+"\" is not imported (add import { "+obj.name+" } from \"std\")")
		}
		return NewSType("func")
	}
	ot := c.inferExpr(obj)
	if ot.kind == "struct" {
		return c.structFieldType(e, ot.name, e.memb)
	}
	if ot.kind == "error" || ot.kind == "unknown" {
		return NewSType("error")
	}
	c.diagAt(e.span, "member access is not allowed on a value of type "+typeStr(ot))
	return NewSType("error")
}

// ---------------------------------------------------------------------------
// Built-in function typing (routing)
// ---------------------------------------------------------------------------

// collectionCallback resolves the function argument of map/filter/reduce/... .
// Top-level functions carry a signature; an fn-typed component parameter is
// accepted without static parameter/return information.
func (c *Checker) collectionCallback(arg *Expr, label string, minParams int) *FnSig {
	if arg == nil {
		return nil
	}
	if arg.kind == "ident" {
		if sig := c.sigs[arg.name]; sig != nil {
			if sig.async {
				c.diagAt(arg.span, label+" callback must be synchronous")
				return sig
			}
			if len(sig.params) < minParams || len(sig.params) > minParams+1 {
				c.diagAt(arg.span, label+" callback takes "+itoaStr(minParams)+" or "+itoaStr(minParams+1)+" arguments")
			}
			return sig
		}
		if v := c.lookupVar(arg.name); v != nil && (v.typ.kind == "fn" || v.typ.kind == "func") {
			return nil
		}
	}
	_ = c.inferExpr(arg)
	c.diagAt(arg.span, label+" expects the name of a function")
	return nil
}

// callbackParamFits checks one callback parameter against the value the runtime
// passes to it. any/unknown/error types are permissive.
func (c *Checker) callbackParamFits(arg *Expr, sig *FnSig, index int, want *SType, label string) {
	if sig == nil || want == nil || index >= len(sig.params) {
		return
	}
	if want.kind == "any" || want.kind == "unknown" || want.kind == "error" {
		return
	}
	got := &sig.params[index]
	if got.kind == "any" || got.kind == "unknown" || got.kind == "error" {
		return
	}
	if !assignable(want, got) {
		c.diagAt(arg.span, label+" callback parameter "+itoaStr(index+1)+" must be "+typeStr(want)+", got "+typeStr(got))
	}
}

func (c *Checker) inferBuiltin(e *Expr) *SType {
	if hasNamedArgs(e.argNames) {
		c.diagAt(e.span, "named arguments are only supported for user functions")
	}
	name := e.target.name
	if e.target != nil && e.target.kind == "member" {
		name = memberFullName(e.target)
	}
	args := e.args
	n := len(args)
	if n == 0 && name != "len" && name != "str" && name != "int" && name != "float" &&
		name != "random" && name != "time.now" && name != "time.nowMs" &&
		name != "storage.clear" && name != "storage.keys" && name != "clipboard.readText" &&
		name != "i18n.locale" && name != "i18n.detect" && name != "i18n.available" && name != "i18n.dir" && name != "i18n.missing" &&
		name != "toast.clear" && name != "intl.timezone" && name != "storage.quota" && name != "storage.sync" && name != "crypto.uuid" && name != "crypto.available" && name != "crypto.status" && !isDndZeroArg(name) {
		c.diagAt(e.span, name+" takes at least one argument")
		return NewSType("error")
	}
	want := func(count int) bool {
		if n != count {
			c.diagAt(e.span, name+" takes exactly "+itoaStr(count)+" arguments, got "+itoaStr(n))
			return false
		}
		return true
	}

	if IsAsyncBuiltin(name) && !c.curAsync {
		c.diagAt(e.span, name+" is async; call it from an async function")
	}

	if strings.HasPrefix(name, "path.") {
		switch name {
		case "path.join":
			if n < 1 || n > 8 {
				c.diagAt(e.span, "path.join takes 1 to 8 arguments, got "+itoaStr(n))
				return NewSType("error")
			}
			for i := 0; i < n; i++ {
				t := c.inferExpr(args[i])
				if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
					c.diagAt(e.span, "path.join arguments must be strings, got "+typeStr(t))
				}
			}
			return NewSType("string")
		case "path.dirname", "path.basename", "path.extname":
			if !want(1) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, name+" requires a string, got "+typeStr(t0))
			}
			return NewSType("string")
		}
	}

	if name == "crypto.available" {
		if !want(0) {
			return NewSType("error")
		}
		return NewSType("bool")
	}
	if name == "crypto.status" {
		if !want(0) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "crypto.uuid" {
		if !want(0) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "crypto.hash" {
		if !want(2) {
			return NewSType("error")
		}
		for i := 0; i < 2; i++ {
			t := c.inferExpr(args[i])
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, "crypto.hash arguments must be strings, got "+typeStr(t))
			}
		}
		return NewSType("string")
	}
	if name == "crypto.randomBytes" {
		if !want(1) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "int" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "crypto.randomBytes requires an int length, got "+typeStr(t0))
		}
		return &SType{kind: "array", elem: NewSType("int")}
	}

	if name == "intl.timezone" {
		if !want(0) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "intl.date" || name == "intl.time" || name == "intl.datetime" {
		if n < 1 || n > 4 {
			c.diagAt(e.span, name+" takes 1 to 4 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "string" && t0.kind != "int" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, name+" requires an ISO string or epoch milliseconds, got "+typeStr(t0))
		}
		for i := 1; i < n; i++ {
			t := c.inferExpr(args[i])
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, name+" option arguments must be strings, got "+typeStr(t))
			}
		}
		return NewSType("string")
	}
	if name == "intl.number" || name == "intl.percent" {
		if n < 1 || n > 3 {
			c.diagAt(e.span, name+" takes 1 to 3 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "int" && t0.kind != "float" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, name+" requires a number, got "+typeStr(t0))
		}
		if n > 1 {
			t1 := c.inferExpr(args[1])
			if t1.kind != "int" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
				c.diagAt(e.span, name+" decimals must be an int, got "+typeStr(t1))
			}
		}
		if n > 2 {
			t2 := c.inferExpr(args[2])
			if t2.kind != "string" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
				c.diagAt(e.span, name+" locale must be a string, got "+typeStr(t2))
			}
		}
		return NewSType("string")
	}
	if name == "intl.currency" {
		if n < 2 || n > 3 {
			c.diagAt(e.span, "intl.currency takes 2 or 3 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "int" && t0.kind != "float" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "intl.currency requires a number, got "+typeStr(t0))
		}
		t1 := c.inferExpr(args[1])
		if t1.kind != "string" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
			c.diagAt(e.span, "intl.currency code must be a string, got "+typeStr(t1))
		}
		if n > 2 {
			t2 := c.inferExpr(args[2])
			if t2.kind != "string" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
				c.diagAt(e.span, "intl.currency locale must be a string, got "+typeStr(t2))
			}
		}
		return NewSType("string")
	}
	if name == "intl.relative" {
		if n < 2 || n > 4 {
			c.diagAt(e.span, "intl.relative takes 2 to 4 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "int" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "intl.relative offset must be an int, got "+typeStr(t0))
		}
		t1 := c.inferExpr(args[1])
		if t1.kind != "string" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
			c.diagAt(e.span, "intl.relative unit must be a string, got "+typeStr(t1))
		}
		for i := 2; i < n; i++ {
			t := c.inferExpr(args[i])
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, "intl.relative option arguments must be strings, got "+typeStr(t))
			}
		}
		return NewSType("string")
	}
	if name == "assert" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "bool" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "assert condition must be a bool, got "+typeStr(t0))
		}
		t1 := c.inferExpr(args[1])
		if t1.kind != "string" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
			c.diagAt(e.span, "assert message must be a string, got "+typeStr(t1))
		}
		return NewSType("null")
	}
	if name == "len" {
		if !want(1) {
			return NewSType("error")
		}
		t := c.inferExpr(args[0])
		if t.kind != "string" && t.kind != "array" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
			c.diagAt(e.span, "len requires a string or an array, got "+typeStr(t))
			return NewSType("error")
		}
		return NewSType("int")
	}
	if name == "append" {
		if !want(2) {
			return NewSType("error")
		}
		ta := c.inferExpr(args[0])
		te := c.inferExpr(args[1])
		if ta.kind == "array" && !assignable(te, ta.elem) {
			c.diagAt(e.span, "append: cannot add a value of type "+typeStr(te)+" to an array of "+typeStr(ta.elem))
			return NewSType("error")
		}
		return ta
	}
	if name == "map" || name == "filter" || name == "find" || name == "findIndex" || name == "any" || name == "all" || name == "sortBy" {
		if !want(2) {
			return NewSType("error")
		}
		ta := c.inferExpr(args[0])
		sig := c.collectionCallback(args[1], name, 1)
		var elem *SType
		if ta.kind == "array" {
			elem = ta.elem
		}
		if sig != nil && elem != nil {
			c.callbackParamFits(args[1], sig, 0, elem, name)
		}
		if name == "filter" || name == "any" || name == "all" {
			if sig != nil && sig.ret != nil && sig.ret.kind != "bool" && sig.ret.kind != "any" && sig.ret.kind != "unknown" && sig.ret.kind != "error" {
				c.diagAt(args[1].span, name+" callback must return bool, got "+typeStr(sig.ret))
			}
			if name == "filter" {
				if ta.kind == "array" {
					return ta
				}
				return &SType{kind: "array", elem: NewSType("any")}
			}
			return NewSType("bool")
		}
		if name == "findIndex" {
			return NewSType("int")
		}
		if name == "find" {
			return NewSType("any")
		}
		if name == "sortBy" {
			if ta.kind == "array" {
				return ta
			}
			return &SType{kind: "array", elem: NewSType("any")}
		}
		// map
		if sig != nil && sig.ret != nil && sig.ret.kind != "error" {
			return &SType{kind: "array", elem: sig.ret}
		}
		return &SType{kind: "array", elem: NewSType("any")}
	}
	if name == "reduce" {
		if !want(3) {
			return NewSType("error")
		}
		ta := c.inferExpr(args[0])
		ti := c.inferExpr(args[1])
		sig := c.collectionCallback(args[2], "reduce", 2)
		if sig != nil {
			if ta.kind == "array" {
				c.callbackParamFits(args[2], sig, 1, ta.elem, "reduce")
			}
			c.callbackParamFits(args[2], sig, 0, ti, "reduce")
			if sig.ret != nil && sig.ret.kind != "error" && sig.ret.kind != "any" && sig.ret.kind != "unknown" && !assignable(sig.ret, ti) {
				c.diagAt(args[2].span, "reduce callback must return "+typeStr(ti)+", got "+typeStr(sig.ret))
			}
		}
		return ti
	}
	if name == "unique" {
		if !want(1) {
			return NewSType("error")
		}
		ta := c.inferExpr(args[0])
		if ta.kind == "array" {
			return ta
		}
		if ta.kind == "any" || ta.kind == "unknown" || ta.kind == "error" {
			return &SType{kind: "array", elem: NewSType("any")}
		}
		c.diagAt(args[0].span, "unique expects an array, got "+typeStr(ta))
		return NewSType("error")
	}
	if name == "t" || name == "i18n.t" {
		if n < 1 || n > 2 {
			c.diagAt(e.span, name+" takes 1 or 2 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, name+" expects a string key, got "+typeStr(t0))
		}
		if n == 2 && args[1] != nil {
			t1 := c.inferExpr(args[1])
			if t1.kind != "map" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
				c.diagAt(e.span, name+" expects a map of placeholders, got "+typeStr(t1))
			}
		}
		return NewSType("string")
	}
	if name == "i18n.setLocale" {
		if !want(1) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "i18n.setLocale expects a locale string, got "+typeStr(t0))
		}
		return NewSType("bool")
	}
	if name == "i18n.locale" || name == "i18n.detect" || name == "i18n.dir" {
		return NewSType("string")
	}
	if name == "i18n.available" || name == "i18n.missing" {
		return &SType{kind: "array", elem: NewSType("string")}
	}
	if name == "i18n.has" {
		if !want(1) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "i18n.has expects a string key, got "+typeStr(t0))
		}
		return NewSType("bool")
	}
	if name == "str" || name == "int" || name == "float" {
		if !want(1) {
			return NewSType("error")
		}
		t := c.inferExpr(args[0])
		if name == "str" {
			return NewSType("string")
		}
		if t.kind != "int" && t.kind != "float" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
			c.diagAt(e.span, name+" expects a number, got "+typeStr(t))
			return NewSType("error")
		}
		if name == "int" {
			return NewSType("int")
		}
		return NewSType("float")
	}
	if name == "keys" || name == "values" {
		if !want(1) {
			return NewSType("error")
		}
		t := c.inferExpr(args[0])
		if t.kind == "error" || t.kind == "unknown" {
			return &SType{kind: "array", elem: NewSType("any")}
		}
		if t.kind == "map" {
			if name == "keys" {
				return &SType{kind: "array", elem: NewSType("string")}
			}
			return &SType{kind: "array", elem: t.val}
		}
		if t.kind == "any" {
			return &SType{kind: "array"}
		}
		c.diagAt(e.span, name+" expects a map, got "+typeStr(t))
		return NewSType("error")
	}
	if name == "merge" {
		if !want(2) {
			return NewSType("error")
		}
		t1 := c.inferExpr(args[0])
		t2 := c.inferExpr(args[1])
		if t1.kind == "map" && t2.kind != "map" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
			c.diagAt(e.span, "merge expects two maps, got "+typeStr(t2))
			return NewSType("error")
		}
		if t1.kind != "map" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
			c.diagAt(e.span, "merge expects two maps, got "+typeStr(t1))
			return NewSType("error")
		}
		if t1.kind == "map" {
			return t1
		}
		if t2.kind == "map" {
			return t2
		}
		return &SType{kind: "map", key: NewSType("string"), val: NewSType("any")}
	}
	if name == "has" {
		if !want(2) {
			return NewSType("error")
		}
		t1 := c.inferExpr(args[0])
		if t1.kind != "map" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
			c.diagAt(e.span, "has expects a map, got "+typeStr(t1))
			return NewSType("error")
		}
		t2 := c.inferExpr(args[1])
		if t2.kind != "string" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
			c.diagAt(e.span, "has expects a string key, got "+typeStr(t2))
			return NewSType("error")
		}
		return NewSType("bool")
	}
	if name == "upper" || name == "lower" || name == "trim" {
		if !want(1) {
			return NewSType("error")
		}
		t := c.inferExpr(args[0])
		if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
			c.diagAt(e.span, name+" expects a string, got "+typeStr(t))
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "contains" {
		if !want(2) {
			return NewSType("error")
		}
		for i := 0; i < 2; i++ {
			t := c.inferExpr(args[i])
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, "contains expects strings, got "+typeStr(t))
				return NewSType("error")
			}
		}
		return NewSType("bool")
	}
	if name == "split" {
		if !want(2) {
			return NewSType("error")
		}
		for i := 0; i < 2; i++ {
			t := c.inferExpr(args[i])
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, "split expects strings, got "+typeStr(t))
				return NewSType("error")
			}
		}
		return &SType{kind: "array", elem: NewSType("string")}
	}
	if name == "join" {
		if !want(2) {
			return NewSType("error")
		}
		ta := c.inferExpr(args[0])
		if ta.kind == "array" && !assignable(ta.elem, NewSType("string")) {
			c.diagAt(e.span, "join expects an array of strings, got "+typeStr(ta))
			return NewSType("error")
		}
		if ta.kind != "array" && ta.kind != "any" && ta.kind != "unknown" && ta.kind != "error" {
			c.diagAt(e.span, "join expects an array of strings, got "+typeStr(ta))
			return NewSType("error")
		}
		t2 := c.inferExpr(args[1])
		if t2.kind != "string" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
			c.diagAt(e.span, "join separator must be a string, got "+typeStr(t2))
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "substring" {
		if !want(3) {
			return NewSType("error")
		}
		s := c.inferExpr(args[0])
		if s.kind != "string" && s.kind != "any" && s.kind != "unknown" && s.kind != "error" {
			c.diagAt(e.span, "substring expects a string, got "+typeStr(s))
			return NewSType("error")
		}
		for i := 1; i <= 2; i++ {
			t := c.inferExpr(args[i])
			if t.kind != "int" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, "substring offsets must be ints, got "+typeStr(t))
				return NewSType("error")
			}
		}
		return NewSType("string")
	}
	if name == "navigate" {
		if !want(1) {
			return NewSType("error")
		}
		t := c.inferExpr(args[0])
		if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
			c.diagAt(e.span, "navigate expects a page name string, got "+typeStr(t))
			return NewSType("error")
		}
		return NewSType("null")
	}
	if name == "modal.open" || name == "modal.close" {
		if !want(1) {
			return NewSType("error")
		}
		t := c.inferExpr(args[0])
		if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
			c.diagAt(e.span, name+" expects a modal name string, got "+typeStr(t))
			return NewSType("error")
		}
		return NewSType("null")
	}
	if name == "toast.show" {
		if n < 1 || n > 3 {
			c.diagAt(e.span, "toast.show takes 1 to 3 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		_ = c.inferExpr(args[0])
		if n >= 2 {
			t := c.inferExpr(args[1])
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, "toast.show kind must be a string, got "+typeStr(t))
				return NewSType("error")
			}
		}
		if n >= 3 {
			t := c.inferExpr(args[2])
			if t.kind != "int" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(e.span, "toast.show duration must be an int, got "+typeStr(t))
				return NewSType("error")
			}
		}
		return NewSType("null")
	}
	if name == "toast.clear" {
		if !want(0) {
			return NewSType("error")
		}
		return NewSType("null")
	}
	if isTimerMath(name) {
		return c.inferTimersMath(e, name)
	}
	if isDndBuiltin(name) {
		return c.inferDndBuiltin(e, name)
	}
	if name == "moveItem" {
		return c.inferMoveItem(e)
	}
	return c.inferStrCollBuiltin(e, name)
}

func isDndZeroArg(name string) bool {
	switch name {
	case "drag.source", "drag.target", "drag.data", "drag.files", "drag.folder", "drag.error", "drag.active", "drag.clear":
		return true
	}
	return false
}

func isDndBuiltin(name string) bool {
	return isDndZeroArg(name)
}

// inferDndBuiltin types the drag-and-drop state accessors. Handlers receive no
// event object, so these accessors are the only supported way to read a drag
// payload from SUI code.
func (c *Checker) inferDndBuiltin(e *Expr, name string) *SType {
	if len(e.args) != 0 {
		c.diagAt(e.span, name+" takes no arguments")
		return NewSType("error")
	}
	switch name {
	case "drag.source", "drag.target", "drag.data", "drag.folder", "drag.error":
		return NewSType("string")
	case "drag.active":
		return NewSType("bool")
	case "drag.files":
		return &SType{kind: "array", elem: &SType{kind: "map", key: NewSType("string"), val: NewSType("any")}}
	case "drag.clear":
		return NewSType("null")
	}
	c.diagAt(e.span, "unknown built-in "+name)
	return NewSType("error")
}

// inferMoveItem types moveItem(array, from, to): it returns a new array with
// the element moved from one index to another. Both indices must be ints.
func (c *Checker) inferMoveItem(e *Expr) *SType {
	if len(e.args) != 3 {
		c.diagAt(e.span, "moveItem takes exactly 3 arguments, got "+itoaStr(len(e.args)))
		return NewSType("error")
	}
	arr := c.inferExpr(e.args[0])
	if arr.kind != "array" && arr.kind != "any" && arr.kind != "unknown" && arr.kind != "error" {
		c.diagAt(e.span, "moveItem expects an array, got "+typeStr(arr))
		return NewSType("error")
	}
	for i := 1; i <= 2; i++ {
		t := c.inferExpr(e.args[i])
		if t.kind != "int" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
			c.diagAt(e.span, "moveItem indices must be ints, got "+typeStr(t))
			return NewSType("error")
		}
	}
	return arr
}

func isTimerMath(name string) bool {
	if name == "after" || name == "interval" || name == "timers.cancel" || name == "timers.debounce" || name == "timers.throttle" {
		return true
	}
	if name == "min" || name == "max" || name == "abs" || name == "round" || name == "floor" ||
		name == "ceil" || name == "pow" || name == "sqrt" || name == "random" || name == "randint" {
		return true
	}
	if strings.HasPrefix(name, "time.") || strings.HasPrefix(name, "url.") || strings.HasPrefix(name, "encoding.") ||
		strings.HasPrefix(name, "storage.") || strings.HasPrefix(name, "clipboard.") || strings.HasPrefix(name, "validation.") {
		return true
	}
	return false
}

func (c *Checker) inferTimersMath(e *Expr, name string) *SType {
	args := e.args
	n := len(args)
	want := func(count int) bool {
		if n != count {
			c.diagAt(e.span, name+" takes exactly "+itoaStr(count)+" arguments, got "+itoaStr(n))
			return false
		}
		return true
	}
	isNum := func(t *SType) bool {
		return t.kind == "int" || t.kind == "float" || t.kind == "any" || t.kind == "unknown" || t.kind == "error"
	}
	if name == "after" || name == "interval" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if !isNum(t0) {
			c.diagAt(e.span, name+" expects a number, got "+typeStr(t0))
			return NewSType("error")
		}
		h := args[1]
		if h == nil || h.kind != "ident" || c.sigs[h.name] == nil {
			c.diagAt(e.span, name+" expects a declared function name as its second argument")
			_ = c.inferExpr(h)
			return NewSType("error")
		}
		_ = c.inferExpr(h)
		return NewSType("null")
	}
	if name == "timers.cancel" {
		if !want(1) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if !isNum(t0) {
			c.diagAt(e.span, name+" expects a number, got "+typeStr(t0))
			return NewSType("error")
		}
		return NewSType("null")
	}
	if name == "timers.debounce" || name == "timers.throttle" {
		if !want(2) {
			return NewSType("error")
		}
		h := args[0]
		if h == nil || h.kind != "ident" || c.sigs[h.name] == nil {
			c.diagAt(e.span, name+" expects a declared function name as its first argument")
			_ = c.inferExpr(h)
			return NewSType("error")
		}
		_ = c.inferExpr(h)
		t1 := c.inferExpr(args[1])
		if !isNum(t1) {
			c.diagAt(e.span, name+" expects a number, got "+typeStr(t1))
			return NewSType("error")
		}
		return NewSType("int")
	}
	// ---- math ----
	numOrErr := func(argName string) bool {
		t := c.inferExpr(args[0])
		if !isNum(t) {
			c.diagAt(e.span, argName+" expects a number, got "+typeStr(t))
			return false
		}
		return true
	}
	if name == "random" {
		if !want(0) {
			return NewSType("error")
		}
		return NewSType("float")
	}
	if name == "randint" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		t1 := c.inferExpr(args[1])
		if !isNum(t0) || !isNum(t1) {
			c.diagAt(e.span, "randint expects numbers, got "+typeStr(t0)+" and "+typeStr(t1))
			return NewSType("error")
		}
		return NewSType("int")
	}
	if name == "round" || name == "floor" || name == "ceil" {
		if !want(1) {
			return NewSType("error")
		}
		if !numOrErr(name) {
			return NewSType("error")
		}
		return NewSType("int")
	}
	if name == "abs" {
		if !want(1) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if !isNum(t0) {
			c.diagAt(e.span, name+" expects a number, got "+typeStr(t0))
			return NewSType("error")
		}
		if t0.kind == "float" {
			return NewSType("float")
		}
		return NewSType("int")
	}
	if name == "sqrt" {
		if !want(1) {
			return NewSType("error")
		}
		if !numOrErr(name) {
			return NewSType("error")
		}
		return NewSType("float")
	}
	if name == "pow" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		t1 := c.inferExpr(args[1])
		if !isNum(t0) || !isNum(t1) {
			c.diagAt(e.span, "pow expects numbers, got "+typeStr(t0)+" and "+typeStr(t1))
			return NewSType("error")
		}
		return NewSType("float")
	}
	if name == "min" || name == "max" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		t1 := c.inferExpr(args[1])
		if !isNum(t0) || !isNum(t1) {
			c.diagAt(e.span, name+" expects numbers, got "+typeStr(t0)+" and "+typeStr(t1))
			return NewSType("error")
		}
		if t0.kind == "float" || t1.kind == "float" {
			return NewSType("float")
		}
		return NewSType("int")
	}
	if strings.HasPrefix(name, "time.") {
		if name == "time.now" || name == "time.nowMs" {
			if !want(0) {
				return NewSType("error")
			}
			if name == "time.now" {
				return NewSType("any")
			}
			return NewSType("int")
		}
		if name == "time.format" {
			if n < 1 || n > 2 {
				c.diagAt(e.span, name+" takes 1 or 2 arguments, got "+itoaStr(n))
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			if n == 2 && args[1] != nil {
				t1 := c.inferExpr(args[1])
				if t1.kind != "string" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
					c.diagAt(e.span, "time.format expects a string format, got "+typeStr(t1))
					return NewSType("error")
				}
			}
			return NewSType("string")
		}
		if name == "time.parse" {
			if n < 1 || n > 2 {
				c.diagAt(e.span, name+" takes 1 or 2 arguments, got "+itoaStr(n))
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, "time.parse expects a string, got "+typeStr(t0))
				return NewSType("error")
			}
			if n == 2 && args[1] != nil {
				t1 := c.inferExpr(args[1])
				if t1.kind != "string" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
					c.diagAt(e.span, "time.parse expects a string format, got "+typeStr(t1))
					return NewSType("error")
				}
			}
			return NewSType("any")
		}
		if name == "time.add" {
			if n < 2 || n > 3 {
				c.diagAt(e.span, name+" takes 2 or 3 arguments, got "+itoaStr(n))
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			t1 := c.inferExpr(args[1])
			if !isNum(t1) {
				c.diagAt(e.span, "time.add expects a number, got "+typeStr(t1))
				return NewSType("error")
			}
			if n == 3 && args[2] != nil {
				t2 := c.inferExpr(args[2])
				if t2.kind != "string" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
					c.diagAt(e.span, "time.add expects a string unit, got "+typeStr(t2))
					return NewSType("error")
				}
			}
			return NewSType("any")
		}
		if name == "time.diff" {
			if n < 2 || n > 3 {
				c.diagAt(e.span, name+" takes 2 or 3 arguments, got "+itoaStr(n))
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			_ = c.inferExpr(args[1])
			if n == 3 && args[2] != nil {
				t2 := c.inferExpr(args[2])
				if t2.kind != "string" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
					c.diagAt(e.span, "time.diff expects a string unit, got "+typeStr(t2))
					return NewSType("error")
				}
			}
			return NewSType("int")
		}
		if name == "time.toISO" || name == "time.toMs" {
			if !want(1) {
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			if name == "time.toISO" {
				return NewSType("string")
			}
			return NewSType("int")
		}
		return NewSType("any")
	}
	if strings.HasPrefix(name, "url.") {
		if name == "url.encode" || name == "url.encodeComponent" {
			if !want(1) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, name+" expects a string, got "+typeStr(t0))
				return NewSType("error")
			}
			return NewSType("string")
		}
		if name == "url.decode" || name == "url.decodeComponent" {
			if !want(1) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, name+" expects a string, got "+typeStr(t0))
				return NewSType("error")
			}
			return NewSType("string")
		}
		return NewSType("string")
	}
	if strings.HasPrefix(name, "encoding.") {
		if name == "encoding.base64Encode" {
			if !want(1) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, name+" expects a string, got "+typeStr(t0))
				return NewSType("error")
			}
			return NewSType("string")
		}
		if name == "encoding.base64Decode" {
			if !want(1) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, name+" expects a string, got "+typeStr(t0))
				return NewSType("error")
			}
			return NewSType("string")
		}
		return NewSType("string")
	}
	if strings.HasPrefix(name, "storage.") {
		if name == "storage.get" {
			if n < 1 || n > 2 {
				c.diagAt(e.span, name+" takes 1 or 2 arguments, got "+itoaStr(n))
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, "storage.get expects a string key, got "+typeStr(t0))
				return NewSType("error")
			}
			if n == 2 && args[1] != nil {
				_ = c.inferExpr(args[1])
			}
			return NewSType("any")
		}
		if name == "storage.set" {
			if !want(2) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, "storage.set expects a string key, got "+typeStr(t0))
				return NewSType("error")
			}
			_ = c.inferExpr(args[1])
			return NewSType("bool")
		}
		if name == "storage.remove" {
			if !want(1) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, "storage.remove expects a string key, got "+typeStr(t0))
				return NewSType("error")
			}
			return NewSType("bool")
		}
		if name == "storage.clear" {
			if !want(0) {
				return NewSType("error")
			}
			return NewSType("bool")
		}
		if name == "storage.keys" {
			if !want(0) {
				return NewSType("error")
			}
			return &SType{kind: "array", elem: NewSType("string")}
		}
		if name == "storage.quota" {
			if !want(0) {
				return NewSType("error")
			}
			return &SType{kind: "map", key: NewSType("string"), val: NewSType("int")}
		}
		if name == "storage.sync" {
			if !want(0) {
				return NewSType("error")
			}
			return NewSType("bool")
		}
		return NewSType("any")
	}
	if strings.HasPrefix(name, "clipboard.") {
		if name == "clipboard.readText" {
			if !want(0) {
				return NewSType("error")
			}
			return NewSType("string")
		}
		if name == "clipboard.writeText" {
			if !want(1) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, "clipboard.writeText expects a string, got "+typeStr(t0))
				return NewSType("error")
			}
			return NewSType("bool")
		}
		return NewSType("any")
	}
	if strings.HasPrefix(name, "validation.") {
		if name == "validation.isEmail" || name == "validation.isUrl" || name == "validation.isNumber" || name == "validation.isBool" ||
			name == "validation.isString" || name == "validation.isArray" || name == "validation.isObject" {
			if !want(1) {
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			return NewSType("bool")
		}
		if name == "validation.isEmpty" || name == "validation.required" {
			if !want(1) {
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			return NewSType("bool")
		}
		if name == "validation.minLen" || name == "validation.maxLen" {
			if !want(2) {
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			t1 := c.inferExpr(args[1])
			if !isNum(t1) {
				c.diagAt(e.span, name+" expects a number, got "+typeStr(t1))
				return NewSType("error")
			}
			return NewSType("bool")
		}
		if name == "validation.length" {
			if n < 2 || n > 3 {
				c.diagAt(e.span, name+" takes 2 or 3 arguments, got "+itoaStr(n))
				return NewSType("error")
			}
			_ = c.inferExpr(args[0])
			t1 := c.inferExpr(args[1])
			if !isNum(t1) {
				c.diagAt(e.span, name+" expects a number, got "+typeStr(t1))
				return NewSType("error")
			}
			if n == 3 {
				t2 := c.inferExpr(args[2])
				if !isNum(t2) {
					c.diagAt(e.span, name+" expects a number, got "+typeStr(t2))
					return NewSType("error")
				}
			}
			return NewSType("bool")
		}
		if name == "validation.pattern" {
			if !want(2) {
				return NewSType("error")
			}
			t0 := c.inferExpr(args[0])
			if t0.kind != "string" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
				c.diagAt(e.span, "validation.pattern expects a string, got "+typeStr(t0))
				return NewSType("error")
			}
			t1 := c.inferExpr(args[1])
			if t1.kind != "string" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
				c.diagAt(e.span, "validation.pattern expects a regex string, got "+typeStr(t1))
				return NewSType("error")
			}
			return NewSType("bool")
		}
		return NewSType("bool")
	}
	c.diagAt(e.span, "unknown built-in "+name)
	return NewSType("error")
}

func (c *Checker) inferStrCollBuiltin(e *Expr, name string) *SType {
	args := e.args
	n := len(args)
	want := func(count int) bool {
		if n != count {
			c.diagAt(e.span, name+" takes exactly "+itoaStr(count)+" arguments, got "+itoaStr(n))
			return false
		}
		return true
	}
	isStrLike := func(t *SType) bool {
		return t.kind == "string" || t.kind == "any" || t.kind == "unknown" || t.kind == "error"
	}
	strArg := func(i int) bool {
		t := c.inferExpr(args[i])
		if !isStrLike(t) {
			c.diagAt(e.span, name+" expects a string, got "+typeStr(t))
			return false
		}
		return true
	}
	intArg := func(i int) bool {
		t := c.inferExpr(args[i])
		if t.kind != "int" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
			c.diagAt(e.span, name+" expects an int, got "+typeStr(t))
			return false
		}
		return true
	}

	if name == "cssSafe" {
		if !want(1) {
			return NewSType("error")
		}
		if !strArg(0) {
			return NewSType("error")
		}
		return NewSType("string")
	}

	if name == "replace" || name == "replaceFirst" {
		if !want(3) {
			return NewSType("error")
		}
		ok := strArg(0)
		for i := 1; i < 3; i++ {
			if !strArg(i) {
				ok = false
			}
		}
		if !ok {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "replaceRegex" {
		if n < 3 || n > 4 {
			c.diagAt(e.span, name+" takes 3 or 4 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		ok := strArg(0) && strArg(1) && strArg(2)
		if n == 4 && args[3] != nil && !strArg(3) {
			ok = false
		}
		if !ok {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "match" {
		if n < 2 || n > 3 {
			c.diagAt(e.span, name+" takes 2 or 3 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		ok := strArg(0) && strArg(1)
		if n == 3 && args[2] != nil && !strArg(2) {
			ok = false
		}
		if !ok {
			return NewSType("error")
		}
		return &SType{kind: "array", elem: NewSType("string")}
	}
	if name == "matchAll" {
		if n < 2 || n > 3 {
			c.diagAt(e.span, name+" takes 2 or 3 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		ok := strArg(0) && strArg(1)
		if n == 3 && args[2] != nil && !strArg(2) {
			ok = false
		}
		if !ok {
			return NewSType("error")
		}
		return &SType{kind: "array", elem: NewSType("any")}
	}
	if name == "search" || name == "test" {
		if !want(2) || !strArg(0) || !strArg(1) {
			return NewSType("error")
		}
		if name == "search" {
			return NewSType("int")
		}
		return NewSType("bool")
	}
	if name == "splitRegex" {
		if n < 2 || n > 3 {
			c.diagAt(e.span, name+" takes 2 or 3 arguments, got "+itoaStr(n))
			return NewSType("error")
		}
		ok := strArg(0) && strArg(1)
		if n == 3 && args[2] != nil {
			t2 := c.inferExpr(args[2])
			if t2.kind != "int" && t2.kind != "any" && t2.kind != "unknown" && t2.kind != "error" {
				c.diagAt(e.span, "splitRegex limit expects int, got "+typeStr(t2))
				ok = false
			}
		}
		if !ok {
			return NewSType("error")
		}
		return &SType{kind: "array", elem: NewSType("string")}
	}
	if name == "extract" {
		if !want(2) || !strArg(0) || !strArg(1) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "extractAll" {
		if !want(2) || !strArg(0) || !strArg(1) {
			return NewSType("error")
		}
		return &SType{kind: "array", elem: NewSType("string")}
	}
	if name == "startsWith" || name == "endsWith" {
		if !want(2) || !strArg(0) || !strArg(1) {
			return NewSType("error")
		}
		return NewSType("bool")
	}
	if name == "indexOf" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		t1 := c.inferExpr(args[1])
		if isStrLike(t0) && isStrLike(t1) {
			return NewSType("int")
		}
		if t0.kind != "array" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "indexOf expects a string or an array, got "+typeStr(t0))
			return NewSType("error")
		}
		return NewSType("int")
	}
	if name == "lastIndexOf" {
		if !want(2) || !strArg(0) || !strArg(1) {
			return NewSType("error")
		}
		return NewSType("int")
	}
	if name == "slice" {
		if !want(3) || !strArg(0) || !intArg(1) || !intArg(2) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "repeat" {
		if !want(2) || !strArg(0) || !intArg(1) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "padLeft" || name == "padRight" {
		if !want(3) || !strArg(0) || !intArg(1) || !strArg(2) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "trimLeft" || name == "trimRight" {
		if !want(1) || !strArg(0) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "chars" {
		if !want(1) || !strArg(0) {
			return NewSType("error")
		}
		return &SType{kind: "array", elem: NewSType("string")}
	}
	if name == "fmt" {
		if !want(2) {
			return NewSType("error")
		}
		n0 := c.inferExpr(args[0])
		if n0.kind != "int" && n0.kind != "float" && n0.kind != "any" && n0.kind != "unknown" && n0.kind != "error" {
			c.diagAt(e.span, "fmt expects a number, got "+typeStr(n0))
			return NewSType("error")
		}
		if !intArg(1) {
			return NewSType("error")
		}
		return NewSType("string")
	}
	if name == "concat" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		t1 := c.inferExpr(args[1])
		if t0.kind == "array" && t1.kind == "array" {
			return t0
		}
		if t0.kind != "array" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "concat expects two arrays, got "+typeStr(t0))
			return NewSType("error")
		}
		if t1.kind != "array" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
			c.diagAt(e.span, "concat expects two arrays, got "+typeStr(t1))
			return NewSType("error")
		}
		if t0.kind == "array" {
			return t0
		}
		if t1.kind == "array" {
			return t1
		}
		return NewSType("array")
	}
	if name == "remove" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind == "error" || t0.kind == "unknown" {
			return t0
		}
		if t0.kind != "array" && t0.kind != "any" {
			c.diagAt(e.span, "remove expects an array, got "+typeStr(t0))
			return NewSType("error")
		}
		if !intArg(1) {
			return NewSType("error")
		}
		if t0.kind == "array" {
			return t0
		}
		return NewSType("array")
	}
	if name == "includes" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind != "array" && t0.kind != "any" && t0.kind != "unknown" && t0.kind != "error" {
			c.diagAt(e.span, "includes expects an array, got "+typeStr(t0))
			return NewSType("error")
		}
		_ = c.inferExpr(args[1])
		return NewSType("bool")
	}
	if name == "sort" || name == "reverse" {
		if !want(1) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind == "error" || t0.kind == "unknown" {
			return t0
		}
		if t0.kind != "array" && t0.kind != "any" {
			c.diagAt(e.span, name+" expects an array, got "+typeStr(t0))
			return NewSType("error")
		}
		if t0.kind == "array" {
			return t0
		}
		return NewSType("array")
	}
	if name == "removeKey" {
		if !want(2) {
			return NewSType("error")
		}
		t0 := c.inferExpr(args[0])
		if t0.kind == "error" || t0.kind == "unknown" {
			return t0
		}
		if t0.kind != "map" && t0.kind != "any" {
			c.diagAt(e.span, "removeKey expects a map, got "+typeStr(t0))
			return NewSType("error")
		}
		t1 := c.inferExpr(args[1])
		if t1.kind != "string" && t1.kind != "any" && t1.kind != "unknown" && t1.kind != "error" {
			c.diagAt(e.span, "removeKey expects a string key, got "+typeStr(t1))
			return NewSType("error")
		}
		if t0.kind == "map" {
			return t0
		}
		return &SType{kind: "map", key: NewSType("string"), val: NewSType("any")}
	}
	c.diagAt(e.span, "unknown built-in "+name)
	return NewSType("error")
}

// ---------------------------------------------------------------------------
// Calls
// ---------------------------------------------------------------------------

func (c *Checker) inferCall(e *Expr) *SType {
	target := e.target
	fireAndForget := false
	// Namespaced builtins (time.now, storage.get, drag.source, ...) are parsed
	// as member calls. They are not std modules, so route them to the builtin
	// table before evaluating the receiver as an ordinary value.
	if target != nil && target.kind == "member" && !isStdCallShape(e) {
		full := memberFullName(target)
		if IsBuiltin(full) {
			return c.inferBuiltin(e)
		}
	}
	if isStdCallShape(e) {
		if hasNamedArgs(e.argNames) {
			c.diagAt(e.span, "named arguments are only supported for user functions")
		}
		var f *StdFunc
		if e.target.obj != nil && !c.stdImported[e.target.obj.name] && !IsCoreModule(e.target.obj.name) {
			c.diagAt(e.span, "std module \""+e.target.obj.name+"\" is not imported (add import { "+e.target.obj.name+" } from \"std\")")
		} else {
			full := ""
			if e.target.obj != nil {
				full = e.target.obj.name + "." + e.target.memb
			}
			ok := false
			f, ok = StdLookup(full)
			if !ok {
				c.diagAt(e.span, "unknown std function \""+full+"\"")
			} else if f.ret0 == "null" {
				// Fire-and-forget intrinsics (ws.send/ws.onMessage/ws.close,
				// fs.write/mkdir/remove, json.stringify): no value to bind.
				checkWsOnMessage(c, e, full)
				fireAndForget = true
			} else {
				c.diagAt(e.span, "call \""+full+"\" returns two values; bind both with a, b = ...")
			}
		}
		if !c.curAsync && f != nil && f.async {
			c.diagAt(e.span, "std I/O calls must be inside an async function")
		}
		if f != nil {
			c.checkStdURLSecurity(e, f)
		}
		for i := 0; i < len(e.args); i++ {
			if e.args[i] != nil {
				_ = c.inferExpr(e.args[i])
			}
		}
		if fireAndForget {
			return NewSType("null")
		}
		return NewSType("any")
	}
	if target != nil && target.kind == "ident" {
		name := target.name
		sig := c.sigs[name]
		if sig == nil {
			if IsBuiltin(name) {
				return c.inferBuiltin(e)
			}
			hint := ""
			if cand := nearestName(name, c.funcNames()); cand != "" {
				hint = "did you mean \"" + cand + "\"? "
			}
			c.diagAtEx(e.span, "unknown function "+name, hint+"define the function with \"fn "+name+"(...) -> ...\"")
			for i := 0; i < len(e.args); i++ {
				if e.args[i] != nil {
					_ = c.inferExpr(e.args[i])
				}
			}
			return NewSType("error")
		}
		args := c.orderNamedArgs(e, sig, name)
		ret := sig.ret
		if len(sig.typeParams) > 0 {
			rets := c.checkGenericCall(name, sig, args, e.span)
			if len(rets) > 0 {
				ret = rets[0]
			}
		} else {
			c.checkArgs(name, sig.params, args, sig.defaults, sig.rest, e.span)
		}
		if c.methodNames[name] && c.curView == "" {
			c.diagAt(e.span, "method \""+name+"\" can only be called inside its component view")
		}
		if sig.multi {
			c.diagAt(e.span, "call \""+name+"\" returns "+itoaStr(len(sig.rets))+" values; bind them with a, b = "+name+"(...)")
		}
		if sig.async && !c.curAsync {
			c.diagAt(e.span, "async function \""+name+"\" cannot be called from a non-async function")
		}
		return ret
	}
	ot := c.inferExpr(target)
	for i := 0; i < len(e.args); i++ {
		if e.args[i] != nil {
			_ = c.inferExpr(e.args[i])
		}
	}
	if ot.kind == "error" || ot.kind == "unknown" {
		return NewSType("error")
	}
	c.diagAt(e.span, "cannot call a value of type "+typeStr(ot))
	return NewSType("error")
}

func (c *Checker) inferIndex(e *Expr) *SType {
	ot := c.inferExpr(e.obj)
	it := c.inferExpr(e.idx)
	if ot.kind == "error" || ot.kind == "unknown" {
		return NewSType("error")
	}
	if ot.kind == "array" {
		return ot.elem
	}
	if ot.kind == "map" {
		if it.kind != "string" && it.kind != "any" && it.kind != "unknown" && it.kind != "error" {
			c.diagAt(e.span, "map index must be a string, got "+typeStr(it))
			return NewSType("error")
		}
		return ot.val
	}
	if ot.kind == "string" {
		c.diagAt(e.span, "cannot index a string value; use substring or chars")
		return NewSType("error")
	}
	if it.kind == "error" || it.kind == "unknown" || ot.kind == "any" {
		return NewSType("any")
	}
	c.diagAt(e.span, "cannot index a value of type "+typeStr(ot))
	return NewSType("error")
}

func (c *Checker) inferUnary(e *Expr) *SType {
	ot := widenType(c.inferExpr(e.operand))
	if e.op == "not" {
		if ot.kind != "bool" && ot.kind != "error" && ot.kind != "unknown" {
			c.diagAt(e.span, "operator not requires a bool operand, got "+typeStr(ot))
			return NewSType("error")
		}
		return NewSType("bool")
	}
	if ot.kind != "int" && ot.kind != "float" && ot.kind != "error" && ot.kind != "unknown" {
		c.diagAt(e.span, "operator - requires an int or float operand, got "+typeStr(ot))
		return NewSType("error")
	}
	return ot
}

func (c *Checker) reportOperatorError(op string, lt, rt *SType, sp Span) {
	if isNumericType(lt) && isNumericType(rt) {
		c.diagAt(sp, "cannot mix int and float")
		return
	}
	sym := op
	if op == "add" {
		sym = "+"
	} else if op == "sub" {
		sym = "-"
	} else if op == "mul" {
		sym = "*"
	} else if op == "div" || op == "fdiv" {
		sym = "/"
	} else if op == "mod" {
		sym = "%"
	}
	c.diagAt(sp, "operator "+sym+" cannot be applied to types "+typeStr(lt)+" and "+typeStr(rt))
}

func (c *Checker) inferBinary(e *Expr) *SType {
	op := e.op
	if op == "and" || op == "or" {
		lt := c.inferExpr(e.l)
		// Short-circuit narrowing: the right operand of `A and B` sees the
		// positive guards of A; the right operand of `A or B` sees the negative
		// guards (the left side was false). A temporary scope removes local
		// shadows after B; state types are restored explicitly.
		c.pushScope()
		saved := c.applyTypeGuards(e.l, op == "and")
		rt := c.inferExpr(e.r)
		c.restoreTypeGuards(saved)
		c.popScope()
		lw := widenType(lt)
		rw := widenType(rt)
		bad := (lw.kind != "bool" && lw.kind != "error" && lw.kind != "unknown") ||
			(rw.kind != "bool" && rw.kind != "error" && rw.kind != "unknown")
		if bad {
			c.diagAt(e.span, "operator "+op+" requires bool operands, got "+typeStr(lt)+" and "+typeStr(rt))
			return NewSType("error")
		}
		return NewSType("bool")
	}
	lt := c.inferExpr(e.l)
	rt := c.inferExpr(e.r)
	// Literal types are subtypes of their base type: concatenation, arithmetic
	// and comparisons see the widened base type.
	lt = widenType(lt)
	rt = widenType(rt)
	if op == "==" || op == "!=" {
		if lt.kind == "struct" && rt.kind == "struct" {
			c.diagAt(e.span, "comparison of struct values is not supported in v0.5; compare their fields instead (u.name == \"x\")")
			return NewSType("bool")
		}
		if lt.kind == "null" || rt.kind == "null" {
			return NewSType("bool")
		}
		if lt.kind == "error" || lt.kind == "unknown" || rt.kind == "error" || rt.kind == "unknown" {
			return NewSType("bool")
		}
		if !overlaps(lt, rt) {
			c.reportOperatorError(op, lt, rt, e.span)
		}
		return NewSType("bool")
	}
	if op == "<" || op == "<=" || op == ">" || op == ">=" {
		if lt.kind == "struct" && rt.kind == "struct" {
			c.diagAt(e.span, "comparison of struct values is not supported in v0.5; compare their fields instead")
			return NewSType("bool")
		}
		if lt.kind == "error" || lt.kind == "unknown" || rt.kind == "error" || rt.kind == "unknown" {
			return NewSType("bool")
		}
		if !overlaps(lt, rt) {
			c.reportOperatorError(op, lt, rt, e.span)
		}
		return NewSType("bool")
	}
	if op == "add" && lt.kind == "string" && rt.kind == "string" {
		return NewSType("string")
	}
	if op == "div" && lt.kind == "float" && rt.kind == "float" {
		e.op = "fdiv"
	}
	return c.inferArith(op, lt, rt, e.span)
}

func (c *Checker) inferArith(op string, lt, rt *SType, sp Span) *SType {
	if (lt.kind == "error" || lt.kind == "unknown") || (rt.kind == "error" || rt.kind == "unknown") {
		return NewSType("error")
	}
	if lt.kind == "int" && rt.kind == "int" {
		return NewSType("int")
	}
	if lt.kind == "float" && rt.kind == "float" {
		return NewSType("float")
	}
	c.reportOperatorError(op, lt, rt, sp)
	return NewSType("error")
}

// ---------------------------------------------------------------------------
// Views
// ---------------------------------------------------------------------------

func isAllowedTag(tag string) bool {
	switch tag {
	// layout & containers
	case "body", "div", "section", "article", "main", "header", "footer", "nav", "aside", "span",
		"row", "col", "spacer", "card", "tablelayout", "layoutcell",
		// headings & text
		"h1", "h2", "h3", "h4", "h5", "h6", "p", "text", "blockquote", "pre", "code",
		"mark", "strong", "em", "small",
		// forms
		"form", "label", "input", "textarea", "select", "option", "button", "fieldset", "legend",
		"checkbox", "radio", "switch", "number", "slider", "file", "color", "date", "time",
		// lists
		"ul", "ol", "li", "list",
		// tables
		"table", "caption", "thead", "tbody", "tfoot", "trow", "tcell",
		// media & embed
		"a", "img", "video", "audio", "iframe", "hr", "br",
		// feedback & nav
		"progress", "badge", "alert", "spinner", "divider", "combobox", "datepicker",
		"tabs", "tab", "accordion", "accordionitem", "tree", "treeitem", "richtext", "chart", "graph", "toast",
		"menubar", "menu", "menuitem", "modal", "modalbutton",
		"dropdown", "dropdownitem", "popover", "drawer", "splitter", "pane", "dropzone",
		"timeline", "timelineitem", "emptystate", "skeleton", "carousel", "slide",
		"stepper", "step", "commandbar", "command", "errorboundary":
		return true
	}
	return false
}

// eventAttr reports whether key names an event handler attribute
// (onClick, onChange, onMouseEnter, ...).
func eventAttr(key string) bool {
	switch key {
	case "onClick", "onDoubleClick", "onChange", "onInput", "onSubmit", "onReset",
		"onRename", "onCheck", "onMove", "onLoad", "onContextMenu", "onTreeChange", "onTaskChange", "onDataChange", "onCalendarChange", "onResourceChange", "onCommit", "onFocus", "onBlur", "onMouseEnter", "onMouseLeave", "onKeyDown", "onKeyUp", "onReady", "onSelectionChange",
		"onCellChange", "onRowChange", "onTableChange", "onRowSelect", "onSort", "onCellClick", "onRowClick",
		"onDragStart", "onDragEnter", "onDragOver", "onDragLeave", "onDrop", "onDragEnd":
		return true
	}
	return false
}

func attrAllowed(key string, tag string) bool {
	if key == "class" || key == "style" || key == "link" {
		return true
	}
	if key == "motion" || key == "motionKey" || key == "motionDuration" || key == "motionEasing" || key == "motionDelay" || key == "motionExit" || key == "motionGraph" {
		return true
	}
	if eventAttr(key) {
		return true
	}
	if key == "draggable" || key == "dragKey" || key == "dropKey" ||
		key == "dropFiles" || key == "pickFiles" || key == "pickMode" || key == "directory" {
		return true
	}
	if tag == "dropzone" && key == "hint" {
		return true
	}
	if tag == "errorboundary" && key == "fallback" {
		return true
	}
	if tag == "button" && key == "onClick" {
		return true
	}
	if tag == "input" && (key == "placeholder" || key == "value" || key == "onChange" ||
		key == "inputtype" || key == "pattern" || key == "minlength" || key == "maxlength" ||
		key == "min" || key == "max" || key == "step" || key == "required" || key == "options") {
		return true
	}
	if tag == "textarea" && (key == "placeholder" || key == "value" || key == "onChange" ||
		key == "pattern" || key == "minlength" || key == "maxlength" || key == "required") {
		return true
	}
	if tag == "select" && (key == "value" || key == "required") {
		return true
	}
	if tag == "radio" && key == "required" {
		return true
	}
	if (tag == "checkbox" || tag == "switch" || tag == "radio") &&
		(key == "shape" || key == "theme" || key == "animation" || key == "duration" || key == "easing" || key == "delay" || key == "stagger") {
		return true
	}
	if (tag == "number" || tag == "date" || tag == "time") && key == "required" {
		return true
	}
	if tag == "combobox" && (key == "value" || key == "options" || key == "placeholder" ||
		key == "empty" || key == "disabled" || key == "required" ||
		key == "maxOptionLength" || key == "maxOptions" ||
		key == "placement" || key == "animation" || key == "duration" || key == "easing" || key == "delay") {
		return true
	}
	if tag == "datepicker" && (key == "value" || key == "mode" || key == "min" || key == "max" ||
		key == "step" || key == "disabled" || key == "required" || key == "variant") {
		return true
	}
	if tag == "spinner" && (key == "variant" || key == "src" || key == "alt") {
		return true
	}
	if tag == "richtext" && (key == "name" || key == "value" || key == "placeholder" || key == "height" ||
		key == "readonly" || key == "toolbar" || key == "menu" ||
		key == "contextToolbar" || key == "properties" || key == "statusbar" || key == "theme" ||
		key == "toolbarItems" || key == "template" ||
		key == "onReady" || key == "onSelectionChange") {
		return true
	}
	if tag == "tree" && (key == "name" || key == "value" || key == "editMode" || key == "editDepth" ||
		key == "renameKey" || key == "renameValue" || key == "renameIcon" || key == "showRenameIcon" ||
		key == "guides" || key == "truncate" || key == "checkbox" || key == "counts" || key == "rowHeight" ||
		key == "iconLeaf" || key == "iconParent" ||
		key == "iconParentOpen" || key == "toggleClosed" || key == "toggleOpen" ||
		key == "multiSelect" || key == "selection" ||
		key == "search" || key == "drag" || key == "contextMenu" || key == "columns" || key == "loadKey" ||
		key == "virtual" || key == "columnLabels" || key == "columnWidths" || key == "sort" || key == "persistSelection" ||
		key == "tooltips" || key == "columnsEnabled" || key == "fontStep") {
		return true
	}
	if tag == "treeitem" && (key == "key" || key == "open" || key == "rename" ||
		key == "renameIcon" || key == "showRenameIcon" || key == "icon" || key == "iconOpen" ||
		key == "badge" || key == "badgeType" || key == "meta" || key == "checked" ||
		key == "lazy" || key == "size" || key == "date" || key == "status" || key == "cells") {
		return true
	}
	if (tag == "chart" || tag == "graph") && (key == "name" || key == "data" || key == "editMode" || key == "chartType" || key == "format" || key == "theme" || key == "colorPicker" || key == "settingsButton" || key == "showZoom" || key == "showAdd" || key == "showHistory" ||
		key == "x" || key == "y" || key == "width" || key == "height" || key == "overflow" ||
		key == "zoom" || key == "pointWidth" || key == "pointSize" || key == "bins" || key == "min" || key == "max" || key == "stateHints" ||
		key == "stateHintBg" || key == "stateHintOpacity" || key == "stateHintTextColor" || key == "stateHintTextAlign" || key == "stateHintBorderColor" ||
		key == "stateHintFont" || key == "stateHintFontSize" || key == "stateHintBorderWidth" || key == "stateHintPadding" || key == "stateHintRadius" ||
		key == "tableView" || key == "tableFilters" || key == "tableAddRow" || key == "tableDeleteRow" ||
		key == "tableSteppers" || key == "tableContextMenu" || key == "tableSelection" ||
		key == "criticalPath" || key == "holidays" || key == "weekend" || key == "timezone" || key == "showNonWorkingDays" || key == "progressColumn" ||
		key == "showGrid" || key == "showLegend" || key == "showValues" ||
		key == "showResources" || key == "showProgress" ||
		key == "showSlack" || key == "showCosts" || key == "statusColors" || key == "ganttScale" || key == "autoSchedule" ||
		key == "linkColor" || key == "linkWidth" || key == "linkGap" || key == "linkStyle" || key == "arrowType" || key == "arrowSize" || key == "barBorderColor" ||
		key == "stateRouting" || key == "stateTagStyle" || key == "stateTagLink" ||
		key == "stateLinkTagStyle" || key == "stateLinkTagLink" || key == "stateShapeTagStyle" || key == "stateShapeTagLink" ||
		key == "stateShapeTagKind" || key == "stateLinkTagKind" || key == "stateShapeTagAlign" || key == "stateLinkTagAlign" ||
		key == "stateShapeTagRotation" || key == "stateLinkTagRotation" ||
		key == "stateShapeBorderColor" || key == "stateSelectionColor" ||
		key == "stateShapeTagArrowSize" || key == "stateLinkTagArrowSize" ||
		key == "stateTagTextColor" || key == "stateShapeTagTextColor" || key == "stateLinkTagTextColor" ||
		key == "stateTagBoxColor" || key == "stateShapeTagBoxColor" || key == "stateLinkTagBoxColor" ||
		key == "stateTagBorderColor" || key == "stateShapeTagBorderColor" || key == "stateLinkTagBorderColor" ||
		key == "stateTagBorderWidth" || key == "stateShapeTagBorderWidth" || key == "stateLinkTagBorderWidth" ||
		key == "stateTagBorderStyle" || key == "stateShapeTagBorderStyle" || key == "stateLinkTagBorderStyle" ||
		key == "stateTagPadding" || key == "stateShapeTagPadding" || key == "stateLinkTagPadding" ||
		key == "stateTagLeaderColor" || key == "stateShapeTagLeaderColor" || key == "stateLinkTagLeaderColor" ||
		key == "stateTagArrowShape" || key == "stateShapeTagArrowShape" || key == "stateLinkTagArrowShape" ||
		key == "stateTagFontWeight" || key == "stateShapeTagFontWeight" || key == "stateLinkTagFontWeight" ||
		key == "stateTagFontStyle" || key == "stateShapeTagFontStyle" || key == "stateLinkTagFontStyle" ||
		key == "stateTagTextDecoration" || key == "stateShapeTagTextDecoration" || key == "stateLinkTagTextDecoration" ||
		key == "stateTagTextAlign" || key == "stateShapeTagTextAlign" || key == "stateLinkTagTextAlign" ||
		key == "stateTagLetterSpacing" || key == "stateShapeTagLetterSpacing" || key == "stateLinkTagLetterSpacing" ||
		key == "stateTagLineSpacing" || key == "stateShapeTagLineSpacing" || key == "stateLinkTagLineSpacing" ||
		key == "stateTagTextTransform" || key == "stateShapeTagTextTransform" || key == "stateLinkTagTextTransform" ||
		key == "stateTagTextShadow" || key == "stateShapeTagTextShadow" || key == "stateLinkTagTextShadow" ||
		key == "stateTagTextShadowColor" || key == "stateShapeTagTextShadowColor" || key == "stateLinkTagTextShadowColor" ||
		key == "stateTagFont" || key == "stateTagFontSize" || key == "stateShapeTagFont" || key == "stateShapeTagFontSize" || key == "stateLinkTagFont" || key == "stateLinkTagFontSize" || key == "stateLinkArcRadius" ||
		key == "stateTagsMovable" || key == "stateRoutesEditable" || key == "stateHandles" || key == "stateStartArrow" ||
		key == "stateGrid" || key == "stateGridGap" || key == "stateGridColor") {
		return true
	}
	if tag == "option" && key == "value" {
		return true
	}
	if (tag == "img" || tag == "video" || tag == "audio" || tag == "iframe") && (key == "src" || key == "alt") {
		return true
	}
	if tag == "iframe" && (key == "sandbox" || key == "allow" || key == "referrerpolicy" ||
		key == "loading" || key == "width" || key == "height" || key == "srcAllow") {
		return true
	}
	if tag == "a" && key == "href" {
		return true
	}
	if tag == "modal" && key == "name" {
		return true
	}
	if tag == "toast" && (key == "position" || key == "offset" ||
		key == "hAlign" || key == "vAlign" || key == "hOffset" || key == "vOffset" ||
		key == "duration" || key == "max" || key == "icons") {
		return true
	}
	if key == "tooltip" {
		return true
	}
	if tag == "modalbutton" && (key == "variant" || key == "close") {
		return true
	}
	if tag == "dropdown" && (key == "position" || key == "align" || key == "placement" || key == "animation" || key == "duration" || key == "easing" || key == "delay") {
		return true
	}
	if tag == "dropdownitem" && (key == "disabled" || key == "separator") {
		return true
	}
	if tag == "popover" && (key == "position" || key == "animation" || key == "duration" || key == "easing" || key == "delay") {
		return true
	}
	if tag == "drawer" && (key == "side" || key == "width" || key == "open" || key == "onClose" || key == "duration" || key == "easing") {
		return true
	}
	if tag == "splitter" && (key == "direction" || key == "initial" || key == "min") {
		return true
	}
	if tag == "timelineitem" && (key == "date" || key == "status") {
		return true
	}
	if tag == "emptystate" && (key == "icon" || key == "hint" || key == "compact") {
		return true
	}
	if tag == "skeleton" && (key == "lines" || key == "width" || key == "height" || key == "circle" || key == "animated" || key == "gap" || key == "duration") {
		return true
	}
	if tag == "carousel" && (key == "value" || key == "autoplay" || key == "interval" || key == "loop" || key == "controls" || key == "indicators" || key == "height" || key == "duration" || key == "easing") {
		return true
	}
	if tag == "slide" && key == "label" {
		return true
	}
	if tag == "stepper" && (key == "value" || key == "clickable" || key == "vertical") {
		return true
	}
	if tag == "step" && (key == "description" || key == "icon") {
		return true
	}
	if tag == "commandbar" && (key == "open" || key == "hotkey" || key == "placeholder" || key == "animation" || key == "duration" || key == "easing" || key == "delay") {
		return true
	}
	if tag == "command" && key == "shortcut" {
		return true
	}
	if tag == "label" && key == "for" {
		return true
	}
	if tag == "tcell" && (key == "head" || key == "colspan" || key == "rowspan") {
		return true
	}
	if tag == "layoutcell" && (key == "row" || key == "col") {
		return true
	}
	if tag == "tablelayout" && (key == "rows" || key == "columns" ||
		key == "width" || key == "height" ||
		key == "columnWidths" || key == "rowHeights" ||
		key == "gapX" || key == "gapY" ||
		key == "borderStyle" || key == "borderWidth" || key == "borderColor" ||
		key == "cellMinHeight" || key == "responsive" || key == "semantic" || key == "label") {
		return true
	}
	if key == "colspan" || key == "rowspan" {
		return true
	}
	if key == "cellPadding" || key == "cellBackground" ||
		key == "cellBorderStyle" || key == "cellBorderWidth" || key == "cellBorderColor" {
		return true
	}
	if tag == "table" && (key == "name" || key == "editMode" || key == "cellText" ||
		key == "columns" || key == "columnLabels" || key == "columnWidths" ||
		key == "columnAlign" || key == "columnTypes" || key == "columnFormats" || key == "columnOptions" ||
		key == "sortable" || key == "filterable" || key == "pageSize" || key == "selection" ||
		key == "multiSelect" || key == "virtual" || key == "showAddTable" || key == "structureEditable") {
		return true
	}
	if tag == "tabs" && (key == "position" || key == "variant") {
		return true
	}
	return false
}

func (c *Checker) checkViewParts(parts []StrPart) {
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "expr" && parts[i].expr != nil {
			_ = c.inferExpr(parts[i].expr)
		}
	}
}

func (c *Checker) checkView(vd *ViewDecl) {
	prevTP := c.curTypeParams
	c.curTypeParams = typeParamSet(vd.typeParams)
	defer func() { c.curTypeParams = prevTP }()
	for i := 0; i < len(vd.title); i++ {
		if vd.title[i].kind == "expr" {
			c.diagAt(vd.span, "view title cannot interpolate values")
		}
	}
	if vd.onLoad != "" && c.sigs[vd.onLoad] == nil {
		msg := "view onLoad must name a declared function"
		hint := ""
		if cand := nearestName(vd.onLoad, c.funcNames()); cand != "" {
			hint = "did you mean \"" + cand + "\"?"
		}
		c.diagAtEx(vd.span, msg, hint)
	}
	c.curView = vd.name
	c.viewIsComp = viewInvoked[vd.name] || len(vd.params) > 0
	c.childrenSeen = false
	c.instTypes = map[string]*SType{}
	c.pushScope()
	for i := 0; i < len(vd.params); i++ {
		if vd.params[i].typename == "fn" {
			c.declareVar(vd.params[i].name, NewSType("fn"), vd.params[i].span)
		} else if pt, perr := parseSUITypeWith(vd.params[i].typename, c.curTypeParams); perr == "" {
			c.declareVar(vd.params[i].name, pt, vd.params[i].span)
		}
	}
	// Per-instance state: locals initialized per component instance.
	seenInst := map[string]bool{}
	for i := 0; i < len(vd.inst_state); i++ {
		sd := &vd.inst_state[i]
		if seenInst[sd.name] {
			c.diagAt(sd.span, "duplicate instance state \""+sd.name+"\" in view")
		}
		seenInst[sd.name] = true
		it := c.inferExpr(sd.init)
		if sd.typename != "" {
			decl, derr := parseSUITypeWith(sd.typename, c.curTypeParams)
			if derr != "" {
				c.diagAt(sd.span, derr)
			} else if !assignable(it, decl) {
				c.diagAt(sd.span, "instance state "+sd.name+" declared as "+typeStr(decl)+" but initialized with "+typeStr(it))
			}
			it = decl
		}
		c.declareVar(sd.name, it, sd.span)
		c.instTypes[sd.name] = it
	}
	// Method bodies run with view params + instance state in scope.
	for i := 0; i < len(vd.methods); i++ {
		c.checkViewMethod(&vd.methods[i])
	}
	c.checkViewBody(vd.body)
	c.popScope()
	c.curView = ""
}

// checkViewMethod checks a view-local method body. Instance state and view
// parameters are the enclosing scope; the method's own params are locals.
func (c *Checker) checkViewMethod(fd *FuncDecl) {
	sig := c.sigs[fd.name]
	if sig == nil {
		return
	}
	prevRet := c.curRet
	prevRets := c.curRets
	prevAsync := c.curAsync
	prevSeen := c.seenReturn
	prevFn := c.fnDepth
	prevInFn := c.inFn

	c.fnDepth = prevFn + 1
	c.inFn = true
	c.curRet = sig.ret
	c.curRets = sig.rets
	c.curAsync = fd.async
	c.seenReturn = false

	c.pushScope()
	for i := 0; i < len(fd.params); i++ {
		c.declareVar(fd.params[i].name, &sig.params[i], fd.params[i].span)
	}
	c.checkStmts(fd.body)
	c.popScope()

	c.curRet = prevRet
	c.curRets = prevRets
	c.curAsync = prevAsync
	c.seenReturn = prevSeen
	c.fnDepth = prevFn
	c.inFn = prevInFn
}

func (c *Checker) checkViewBody(nodes []ViewNode) {
	for i := 0; i < len(nodes); i++ {
		c.checkViewNode(&nodes[i])
	}
}

func (c *Checker) checkViewNode(n *ViewNode) {
	if n == nil {
		return
	}
	if n.kind == "element" {
		if n.elem != nil {
			e := n.elem
			if e.tag == "children" {
				c.checkChildren(e)
			} else if !isAllowedTag(e.tag) && c.viewParams[e.tag] != nil {
				c.checkComponentCall(e, e.tag)
			} else {
				c.checkElement(e)
			}
		}
		return
	}
	if n.kind == "if" {
		ct := c.inferExpr(n.cond)
		if ct.kind != "bool" && ct.kind != "error" && ct.kind != "unknown" {
			c.diagAt(n.span, "view if condition must be a bool, got "+typeStr(ct))
		}
		c.pushScope()
		saved := c.applyTypeGuards(n.cond, true)
		c.checkViewBody(derefNodeList(n.then_))
		c.restoreTypeGuards(saved)
		c.popScope()
		c.pushScope()
		saved = c.applyTypeGuards(n.cond, false)
		c.checkViewBody(derefNodeList(n.else_))
		c.restoreTypeGuards(saved)
		c.popScope()
		return
	}
	if n.kind == "list" {
		var itemType *SType
		var item2Type *SType
		it := n.iterable
		if it != nil && it.kind == "range" {
			itemType = NewSType("int")
			if it.lo != nil {
				lt := c.inferExpr(it.lo)
				if lt.kind != "int" && lt.kind != "error" && lt.kind != "unknown" {
					c.diagAt(it.lo.span, "view list range bounds must be ints, got "+typeStr(lt))
				}
			}
			if it.hi != nil {
				ht := c.inferExpr(it.hi)
				if ht.kind != "int" && ht.kind != "error" && ht.kind != "unknown" {
					c.diagAt(it.hi.span, "view list range bounds must be ints, got "+typeStr(ht))
				}
			}
		} else {
			tt := c.inferExpr(it)
			if tt.kind == "array" {
				itemType = tt.elem
			} else if tt.kind == "map" {
				itemType = NewSType("string")
				item2Type = tt.val
			} else if tt.kind == "error" || tt.kind == "unknown" {
				itemType = NewSType("error")
			} else {
				c.diagAt(n.span, "view list iterable must be an array, a map or a range, got "+typeStr(tt))
				itemType = NewSType("error")
			}
		}
		if item2Type != nil && n.item2 == "" {
			c.diagAt(n.span, "iterating a map in a view requires two names: list k, v in <map>")
		} else if item2Type == nil && n.item2 != "" {
			c.diagAt(n.span, "two-name view list requires a map, but this iterable is not a map")
		}
		c.pushScope()
		c.declareVar(n.item, itemType, n.span)
		if n.item2 != "" {
			if item2Type == nil {
				item2Type = NewSType("error")
			}
			c.declareVar(n.item2, item2Type, n.span)
		}
		if n.key != nil {
			kt := c.inferExpr(n.key)
			if kt.kind != "string" && kt.kind != "int" && kt.kind != "float" && kt.kind != "bool" &&
				kt.kind != "enum" && kt.kind != "typevar" && kt.kind != "any" && kt.kind != "unknown" && kt.kind != "error" {
				c.diagAt(n.key.span, "list key must be a string, number, bool or enum, got "+typeStr(kt))
			}
		}
		c.checkViewBody(derefNodeList(n.body))
		c.popScope()
		return
	}
	c.diagAt(n.span, "unknown view node kind "+n.kind)
}

func (c *Checker) checkElement(e *Element) {
	if !isAllowedTag(e.tag) {
		c.diagAt(e.span, "unknown tag "+e.tag)
		return
	}
	c.checkViewParts(e.text)
	seenAttr := map[string]bool{}
	for i := 0; i < len(e.attrs); i++ {
		a := &e.attrs[i]
		if seenAttr[a.key] {
			c.diagAt(a.span, "cannot accept duplicate attribute \""+a.key+"\"")
		}
		seenAttr[a.key] = true
		c.checkAttr(a, e.tag)
	}
	if e.tag == "iframe" {
		c.checkIframe(e)
	}
	if (e.tag == "chart" || e.tag == "graph") && len(e.children) > 0 {
		kind := "chart"
		if e.tag == "graph" {
			kind = "graph"
		}
		c.diagAt(e.span, kind+" cannot contain child elements")
	}
	if e.tag == "splitter" {
		panes := 0
		for i := 0; i < len(e.children); i++ {
			ch := &e.children[i]
			if ch.kind == "element" && ch.elem != nil {
				if ch.elem.tag == "pane" {
					panes++
				} else {
					c.diagAt(ch.elem.span, "splitter can only contain pane children")
				}
			}
		}
		if panes != 2 {
			c.diagAt(e.span, "splitter requires exactly two pane children")
		}
	}
	if e.tag == "timeline" || e.tag == "carousel" || e.tag == "stepper" || e.tag == "commandbar" {
		want := map[string]string{
			"timeline":   "timelineitem",
			"carousel":   "slide",
			"stepper":    "step",
			"commandbar": "command",
		}[e.tag]
		count := 0
		for i := 0; i < len(e.children); i++ {
			ch := &e.children[i]
			if ch.kind != "element" || ch.elem == nil {
				continue
			}
			if ch.elem.tag != want {
				c.diagAt(ch.elem.span, e.tag+" can only contain "+want+" children")
				continue
			}
			count++
		}
		if count == 0 && (e.tag == "carousel" || e.tag == "stepper") {
			c.diagAt(e.span, e.tag+" requires at least one "+want+" child")
		}
	}
	if e.tag == "skeleton" && len(e.children) > 0 {
		c.diagAt(e.span, "skeleton cannot contain child elements")
	}
	c.checkControlBinding(e)
	if len(e.children) > 0 {
		c.checkViewBody(e.children)
	}
}

// checkControlBinding warns about controls that are not actually wired to
// anything. A view re-renders by rebuilding its HTML, so a value control
// (switch/checkbox/radio/input/select/slider/...) whose binding does not
// reference a declared state — and that declares no event handler — is inert:
// nothing can ever read or set its value, and each re-render resets it. That
// is almost always a mistake worth flagging.
//
// The binding is "real" when its expression (or an interpolated string part)
// reaches a declared state; a static literal like `checked: true` or
// `value: 5` does not count.
func (c *Checker) checkControlBinding(e *Element) {
	var bindKey string
	var hintKey string
	switch e.tag {
	case "switch", "checkbox", "radio":
		bindKey, hintKey = "checked", "checked"
	case "input", "textarea", "select", "number", "slider", "color", "date", "time":
		bindKey, hintKey = "value", "value"
	case "file":
		// A file control has no useful value binding; only handlers matter.
		bindKey = ""
	case "button":
		// A button only acts through an event handler; without one a click
		// does nothing, so it is inert the same way an unbound control is.
		bindKey = ""
	default:
		return
	}
	bound := false
	if bindKey != "" {
		if a := findAttr(e, bindKey); a != nil {
			if a.hasExpr && c.exprUsesState(a.expr) {
				bound = true
			}
			if !bound {
				for i := 0; i < len(a.text); i++ {
					if a.text[i].kind == "expr" && c.exprUsesState(a.text[i].expr) {
						bound = true
						break
					}
				}
			}
		}
	}
	hasHandler := false
	for i := 0; i < len(e.attrs); i++ {
		if (eventAttr(e.attrs[i].key) || e.attrs[i].key == "dropFiles" || e.attrs[i].key == "pickFiles") && e.attrs[i].hasExpr {
			hasHandler = true
			break
		}
	}
	if bound || (bindKey == "" && hasHandler) {
		return
	}
	if bindKey == "" {
		hint := "add onChange: <fn> to react to the chosen value"
		if e.tag == "button" {
			hint = "add onClick: <fn> to react to the click"
		}
		c.dd.AddEx(e.span.file, e.span.line, e.span.col, "warning",
			e.tag+" has no handler, so nothing reacts to it",
			hint)
		return
	}
	if hasHandler {
		c.dd.AddEx(e.span.file, e.span.line, e.span.col, "warning",
			e.tag+" has a handler but no "+bindKey+": binding, so it will not reflect state",
			"add "+bindKey+": <state>")
		return
	}
	c.dd.AddEx(e.span.file, e.span.line, e.span.col, "warning",
		e.tag+" control is not bound to any state and has no handler; nothing reads its value and re-renders reset it",
		"bind it with "+hintKey+": <state> to make it live")
}

// exprUsesState reports whether the expression references a declared state
// variable, recursively (including interpolated string parts and nested
// calls/composites).
func (c *Checker) exprUsesState(e *Expr) bool {
	if e == nil {
		return false
	}
	switch e.kind {
	case "ident":
		return c.states[e.name] != nil
	case "member", "index":
		return c.exprUsesState(e.obj) || c.exprUsesState(e.idx)
	case "unary":
		return c.exprUsesState(e.operand)
	case "call":
		if c.exprUsesState(e.target) {
			return true
		}
		for i := 0; i < len(e.args); i++ {
			if c.exprUsesState(e.args[i]) {
				return true
			}
		}
	case "array":
		for i := 0; i < len(e.elements); i++ {
			if c.exprUsesState(e.elements[i]) {
				return true
			}
		}
	case "map", "structlit":
		for i := 0; i < len(e.vals); i++ {
			if c.exprUsesState(e.vals[i]) {
				return true
			}
		}
	case "binary":
		return c.exprUsesState(e.l) || c.exprUsesState(e.r)
	case "cond":
		return c.exprUsesState(e.cond) || c.exprUsesState(e.l) || c.exprUsesState(e.r)
	case "range":
		return c.exprUsesState(e.lo) || c.exprUsesState(e.hi)
	case "str":
		for i := 0; i < len(e.parts); i++ {
			if e.parts[i].kind == "expr" && c.exprUsesState(e.parts[i].expr) {
				return true
			}
		}
	}
	return false
}

func (c *Checker) checkAttr(a *Attr, tag string) {
	if !attrAllowed(a.key, tag) {
		// Allow common attributes for new elements
		if a.key == "label" || a.key == "checked" || a.key == "disabled" || a.key == "name" ||
			a.key == "min" || a.key == "max" || a.key == "step" || a.key == "accept" ||
			a.key == "multiple" || a.key == "value" || a.key == "options" {
			// allow
		} else {
			c.diagAt(a.span, "attribute \""+a.key+"\" is not allowed on <"+tag+">")
			return
		}
	}
	if eventAttr(a.key) {
		if a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
			if c.symKinds[a.expr.name] == "func" {
				return
			}
			if c.methodNames[a.expr.name] && c.curView != "" {
				return
			}
			if v := c.lookupVar(a.expr.name); v != nil && v.typ.kind == "fn" {
				return
			}
		}
		c.diagAt(a.span, a.key+" must name a declared function")
		return
	}
	if a.key == "motionGraph" {
		// Dynamic graph expressions cannot be validated at compile time. A
		// literal graph is checked here so hand-written sources get the same
		// diagnostics as designer-authored ones.
		if a.hasExpr {
			return
		}
		for _, msg := range motionGraphDiagnostics(partsLiteral(a.text)) {
			c.diagAt(a.span, msg)
		}
		return
	}
	if a.key == "editMode" && (tag == "table" || tag == "chart" || tag == "graph" || tag == "tree") {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "editMode must be a string, got "+typeStr(t))
			}
			return
		}
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		v = strings.Trim(v, "\"'")
		if v != "none" && v != "editable" && v != "persistent" {
			c.diagAt(a.span, "editMode must be one of: none, editable, persistent")
		}
		return
	}
	if a.key == "dropFiles" || a.key == "pickFiles" {
		// These bind an element to a state that receives dropped/selected
		// paths. The runtime calls the state setter, so the target must be a
		// real, top-level string state (not a temporary).
		if !a.hasExpr || a.expr == nil || a.expr.kind != "ident" {
			c.diagAt(a.span, a.key+" must name a declared state")
			return
		}
		st := c.states[a.expr.name]
		if st == nil {
			c.diagAt(a.span, a.key+" must name a declared state")
			return
		}
		if st.kind != "string" && st.kind != "any" && st.kind != "unknown" && st.kind != "error" {
			c.diagAt(a.span, a.key+" state must be a string, got "+typeStr(st))
		}
		return
	}
	if (tag == "tree" || tag == "treeitem") && a.key == "showRenameIcon" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "showRenameIcon must be a bool, got "+typeStr(t))
			}
		}
		return
	}
	if (tag == "tree" || tag == "treeitem") && a.key == "renameIcon" {
		if a.hasExpr {
			c.inferExpr(a.expr)
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	if tag == "treeitem" && a.key == "rename" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "rename must be a bool, got "+typeStr(t))
			}
		}
		return
	}
	if tag == "tree" && a.key == "fontStep" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "int" && t.kind != "float" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "fontStep must be a number, got "+typeStr(t))
			}
		} else if n, err := strconv.ParseFloat(strings.TrimSpace(partsContent(a.text)), 64); err != nil || n < 0 {
			c.diagAt(a.span, "fontStep must be a non-negative number")
		}
		return
	}
	if tag == "tree" && a.key == "rowHeight" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "int" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "rowHeight must be an integer, got "+typeStr(t))
			}
		} else if n, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text))); err != nil || n < 12 || n > 120 {
			c.diagAt(a.span, "rowHeight must be an integer between 12 and 120")
		}
		return
	}
	if tag == "tree" && (a.key == "guides" || a.key == "truncate" || a.key == "checkbox" || a.key == "counts" ||
		a.key == "multiSelect" || a.key == "search" || a.key == "drag" || a.key == "contextMenu" || a.key == "virtual" ||
		a.key == "persistSelection" || a.key == "tooltips" || a.key == "columnsEnabled") {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, a.key+" must be a bool, got "+typeStr(t))
			}
		}
		return
	}
	if tag == "treeitem" && a.key == "lazy" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "lazy must be a bool, got "+typeStr(t))
			}
		}
		return
	}
	if tag == "tree" && (a.key == "selection" || a.key == "loadKey") {
		if !a.hasExpr || a.expr == nil || a.expr.kind != "ident" || c.states[a.expr.name] == nil {
			c.diagAt(a.span, a.key+" must name a declared state")
			return
		}
		st := c.states[a.expr.name]
		if st.kind != "string" && st.kind != "any" && st.kind != "unknown" && st.kind != "error" {
			c.diagAt(a.span, a.key+" state must be a string, got "+typeStr(st))
		}
		return
	}
	if tag == "tree" && a.key == "columns" {
		if a.hasExpr {
			c.inferExpr(a.expr)
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	if tag == "tree" && (a.key == "columnLabels" || a.key == "columnWidths") {
		if a.hasExpr {
			c.inferExpr(a.expr)
			return
		}
		v := partsContent(a.text)
		if a.key == "columnWidths" {
			for _, part := range strings.Split(v, ",") {
				if strings.TrimSpace(part) == "" {
					continue
				}
				n, err := strconv.Atoi(strings.TrimSpace(part))
				if err != nil || n < 20 || n > 600 {
					c.diagAt(a.span, "columnWidths must be comma-separated integers between 20 and 600")
					break
				}
			}
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	if tag == "tree" && a.key == "sort" {
		if a.hasExpr {
			c.inferExpr(a.expr)
			return
		}
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		if v != "" && v != "asc" && v != "desc" {
			c.diagAt(a.span, "sort must be asc or desc")
		}
		return
	}
	if tag == "treeitem" && (a.key == "size" || a.key == "date" || a.key == "status" || a.key == "cells") {
		if a.hasExpr {
			c.inferExpr(a.expr)
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	if tag == "treeitem" && a.key == "checked" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "checked must be a bool, got "+typeStr(t))
			}
		}
		return
	}
	if (tag == "tree" || tag == "treeitem") && (a.key == "icon" || a.key == "iconOpen" || a.key == "badge" || a.key == "meta") {
		if a.hasExpr {
			c.inferExpr(a.expr)
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	if tag == "tree" && (a.key == "iconLeaf" || a.key == "iconParent" || a.key == "iconParentOpen" || a.key == "toggleClosed" || a.key == "toggleOpen") {
		if a.hasExpr {
			c.inferExpr(a.expr)
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	if tag == "treeitem" && a.key == "badgeType" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "badgeType must be a string, got "+typeStr(t))
			}
		} else {
			v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
			if v != "neutral" && v != "info" && v != "success" && v != "warning" && v != "danger" {
				c.diagAt(a.span, "badgeType must be neutral, info, success, warning or danger")
			}
		}
		return
	}
	if tag == "tree" && a.key == "editDepth" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "int" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, "editDepth must be an integer, got "+typeStr(t))
			}
		} else if n, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text))); err != nil || n < 0 {
			c.diagAt(a.span, "editDepth must be a non-negative integer")
		}
		return
	}
	if tag == "tree" && (a.key == "renameKey" || a.key == "renameValue") {
		if !a.hasExpr || a.expr == nil || a.expr.kind != "ident" || c.states[a.expr.name] == nil {
			c.diagAt(a.span, a.key+" must name a declared state")
		}
		return
	}
	if tag == "tablelayout" && (a.key == "rows" || a.key == "columns") {
		n := 0
		if a.hasExpr {
			if a.expr == nil || a.expr.kind != "int" {
				c.diagAt(a.span, a.key+" must be a literal integer")
				return
			}
			n = a.expr.i
		} else {
			v, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text)))
			if err != nil {
				c.diagAt(a.span, a.key+" must be a literal integer")
				return
			}
			n = v
		}
		min, max := 1, 12
		if a.key == "rows" {
			max = 50
		}
		if n < min || n > max {
			c.diagAt(a.span, a.key+" must be an integer between "+strconv.Itoa(min)+" and "+strconv.Itoa(max))
		}
		return
	}
	if tag == "tablelayout" && a.key == "borderStyle" {
		if a.hasExpr {
			c.diagAt(a.span, "borderStyle must be a literal value")
			return
		}
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		switch v {
		case "none", "solid", "dashed", "dotted", "double":
		default:
			c.diagAt(a.span, "borderStyle must be one of: none, solid, dashed, dotted, double")
		}
		return
	}
	if tag == "tablelayout" && (a.key == "width" || a.key == "height" ||
		a.key == "columnWidths" || a.key == "rowHeights" ||
		a.key == "gapX" || a.key == "gapY" ||
		a.key == "borderWidth" || a.key == "borderColor" || a.key == "cellMinHeight" ||
		a.key == "label") {
		if a.hasExpr {
			c.diagAt(a.span, a.key+" must be a literal value")
		}
		return
	}
	if tag == "tablelayout" && (a.key == "responsive" || a.key == "semantic") {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, a.key+" must be a bool, got "+typeStr(t))
			}
		}
		return
	}
	if tag == "layoutcell" && (a.key == "row" || a.key == "col") {
		n := 0
		if a.hasExpr {
			if a.expr == nil || a.expr.kind != "int" {
				c.diagAt(a.span, a.key+" must be a literal integer")
				return
			}
			n = a.expr.i
		} else {
			v, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text)))
			if err != nil {
				c.diagAt(a.span, a.key+" must be a literal integer")
				return
			}
			n = v
		}
		max := 49
		if a.key == "col" {
			max = 11
		}
		if n < 0 || n > max {
			c.diagAt(a.span, a.key+" must be an integer between 0 and "+strconv.Itoa(max))
		}
		return
	}
	if a.key == "colspan" || a.key == "rowspan" {
		n := 0
		if a.hasExpr {
			if a.expr == nil || a.expr.kind != "int" {
				c.diagAt(a.span, a.key+" must be a literal integer")
				return
			}
			n = a.expr.i
		} else {
			v, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text)))
			if err != nil {
				c.diagAt(a.span, a.key+" must be a literal integer")
				return
			}
			n = v
		}
		if n < 1 || n > 50 {
			c.diagAt(a.span, a.key+" must be an integer between 1 and 50")
		}
		return
	}
	if a.key == "cellBorderStyle" {
		if a.hasExpr {
			c.diagAt(a.span, a.key+" must be a literal value")
			return
		}
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		allowed := map[string][]string{
			"cellBorderStyle": {"none", "solid", "dashed", "dotted", "double"},
		}[a.key]
		ok := false
		for _, want := range allowed {
			if v == want {
				ok = true
				break
			}
		}
		if !ok {
			c.diagAt(a.span, a.key+" must be one of: "+strings.Join(allowed, ", "))
		}
		return
	}
	if a.key == "cellPadding" || a.key == "cellBackground" || a.key == "cellBorderWidth" || a.key == "cellBorderColor" {
		if a.hasExpr {
			c.diagAt(a.span, a.key+" must be a literal value")
		}
		return
	}
	if tag == "graph" && a.key == "chartType" {
		c.diagAt(a.span, "graph always renders a block diagram; chartType is not allowed on <graph>")
		return
	}
	if tag == "chart" || tag == "graph" {
		switch a.key {
		case "chartType", "format", "overflow", "theme", "ganttScale", "linkStyle", "arrowType", "stateRouting", "stateTagStyle", "stateTagLink",
			"stateLinkTagStyle", "stateLinkTagLink", "stateShapeTagStyle", "stateShapeTagLink",
			"stateTagArrowShape", "stateShapeTagArrowShape", "stateLinkTagArrowShape",
			"stateTagFontWeight", "stateShapeTagFontWeight", "stateLinkTagFontWeight",
			"stateTagFontStyle", "stateShapeTagFontStyle", "stateLinkTagFontStyle",
			"stateTagTextDecoration", "stateShapeTagTextDecoration", "stateLinkTagTextDecoration",
			"stateTagTextAlign", "stateShapeTagTextAlign", "stateLinkTagTextAlign",
			"stateTagTextTransform", "stateShapeTagTextTransform", "stateLinkTagTextTransform",
			"stateTagTextShadow", "stateShapeTagTextShadow", "stateLinkTagTextShadow",
			"stateTagBorderStyle", "stateShapeTagBorderStyle", "stateLinkTagBorderStyle":
			allowed := map[string][]string{
				"chartType":                   {"bar", "barhorizontal", "barstacked", "bargrouped", "line", "step", "smooth", "area", "combo", "pie", "donut", "scatter", "histogram", "waterfall", "radar", "gauge", "funnel", "heatmap", "candlestick", "treemap", "boxplot", "bullet", "dumbbell", "gantt"},
				"format":                      {"auto", "json", "xml", "csv"},
				"overflow":                    {"scroll", "zoom", "fit"},
				"theme":                       {"dark", "light"},
				"ganttScale":                  {"auto", "day", "week", "month", "year"},
				"linkStyle":                   {"auto", "solid", "dashed", "dotted"},
				"arrowType":                   {"triangle", "open", "circle", "square", "none"},
				"stateRouting":                {"curved", "bezier", "arc", "circular", "orthogonal", "straight"},
				"stateTagStyle":               {"plain", "box", "pill", "rectangle", "rounded", "square", "circle", "ellipse", "line", "line-start", "line-end"},
				"stateTagLink":                {"none", "line", "arrow", "bezier", "bezier-arrow"},
				"stateLinkTagStyle":           {"plain", "box", "pill", "rectangle", "rounded", "square", "circle", "ellipse", "line", "line-start", "line-end"},
				"stateLinkTagLink":            {"none", "line", "arrow", "bezier", "bezier-arrow"},
				"stateShapeTagStyle":          {"plain", "box", "pill", "rectangle", "rounded", "square", "circle", "ellipse", "line", "line-start", "line-end"},
				"stateShapeTagLink":           {"none", "line", "arrow", "bezier", "bezier-arrow"},
				"stateTagArrowShape":          {"triangle", "open", "circle", "square"},
				"stateShapeTagArrowShape":     {"triangle", "open", "circle", "square"},
				"stateLinkTagArrowShape":      {"triangle", "open", "circle", "square"},
				"stateTagBorderStyle":         {"solid", "dashed", "dotted"},
				"stateShapeTagBorderStyle":    {"solid", "dashed", "dotted"},
				"stateLinkTagBorderStyle":     {"solid", "dashed", "dotted"},
				"stateTagFontWeight":          {"normal", "300", "400", "500", "600", "700", "800", "900"},
				"stateShapeTagFontWeight":     {"normal", "300", "400", "500", "600", "700", "800", "900"},
				"stateLinkTagFontWeight":      {"normal", "300", "400", "500", "600", "700", "800", "900"},
				"stateTagFontStyle":           {"normal", "italic", "oblique"},
				"stateShapeTagFontStyle":      {"normal", "italic", "oblique"},
				"stateLinkTagFontStyle":       {"normal", "italic", "oblique"},
				"stateTagTextDecoration":      {"none", "underline", "line-through", "overline"},
				"stateShapeTagTextDecoration": {"none", "underline", "line-through", "overline"},
				"stateLinkTagTextDecoration":  {"none", "underline", "line-through", "overline"},
				"stateTagTextAlign":           {"left", "center", "right"},
				"stateShapeTagTextAlign":      {"left", "center", "right"},
				"stateLinkTagTextAlign":       {"left", "center", "right"},
				"stateTagTextTransform":       {"none", "uppercase", "lowercase", "capitalize"},
				"stateShapeTagTextTransform":  {"none", "uppercase", "lowercase", "capitalize"},
				"stateLinkTagTextTransform":   {"none", "uppercase", "lowercase", "capitalize"},
				"stateTagTextShadow":          {"none", "soft", "medium", "strong"},
				"stateShapeTagTextShadow":     {"none", "soft", "medium", "strong"},
				"stateLinkTagTextShadow":      {"none", "soft", "medium", "strong"},
			}[a.key]
			if a.hasExpr {
				t := c.inferExpr(a.expr)
				if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
					c.diagAt(a.span, a.key+" must be a string, got "+typeStr(t))
				}
				return
			}
			v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
			ok := false
			for _, want := range allowed {
				if v == want {
					ok = true
					break
				}
			}
			if !ok {
				c.diagAt(a.span, a.key+" must be one of: "+strings.Join(allowed, ", "))
			}
			return
		case "showGrid", "showLegend", "showValues", "tableFilters", "tableAddRow", "tableDeleteRow", "tableSteppers", "tableContextMenu", "criticalPath", "colorPicker", "settingsButton", "showZoom", "showAdd", "showHistory",
			"showResources", "showProgress", "showSlack", "showCosts", "statusColors", "autoSchedule", "showNonWorkingDays", "progressColumn",
			"stateTagsMovable", "stateRoutesEditable", "stateHandles", "stateStartArrow":
			if a.hasExpr {
				t := c.inferExpr(a.expr)
				if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
					c.diagAt(a.span, a.key+" must be a bool, got "+typeStr(t))
				}
			}
			return
		case "tableSelection":
			allowed := []string{"none", "row", "column", "cell"}
			if a.hasExpr {
				t := c.inferExpr(a.expr)
				if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
					c.diagAt(a.span, a.key+" must be a string, got "+typeStr(t))
				}
			} else {
				v := strings.TrimSpace(partsContent(a.text))
				v = strings.Trim(v, "\"'")
				ok := false
				for _, allowedValue := range allowed {
					if v == allowedValue {
						ok = true
						break
					}
				}
				if !ok {
					c.diagAt(a.span, a.key+" must be one of: "+strings.Join(allowed, ", "))
				}
			}
			return
		case "holidays", "weekend", "timezone", "linkColor", "barBorderColor":
			if a.hasExpr {
				t := c.inferExpr(a.expr)
				if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
					c.diagAt(a.span, a.key+" must be a string, got "+typeStr(t))
				}
			}
			return
		case "zoom", "pointWidth", "pointSize", "bins", "min", "max", "linkWidth", "linkGap", "arrowSize", "stateTagFontSize", "stateShapeTagFontSize", "stateLinkTagFontSize", "stateLinkArcRadius", "stateGridGap":
			if a.hasExpr {
				t := c.inferExpr(a.expr)
				if t.kind != "int" && t.kind != "float" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
					c.diagAt(a.span, a.key+" must be a number, got "+typeStr(t))
				}
			} else if _, err := strconv.ParseFloat(strings.TrimSpace(partsContent(a.text)), 64); err != nil {
				c.diagAt(a.span, a.key+" must be a number")
			}
			return
		case "tableView":
			if a.hasExpr {
				t := c.inferExpr(a.expr)
				if t.kind != "string" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
					c.diagAt(a.span, a.key+" must be a string, got "+typeStr(t))
				}
				return
			}
			v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
			if v != "chart" && v != "table" && v != "both" {
				c.diagAt(a.span, a.key+" must be one of: chart, table, both")
			}
			return
		}
	}
	if tag == "datepicker" && a.key == "mode" {
		if a.hasExpr {
			c.diagAt(a.span, "mode must be one of: date, time, datetime")
			return
		}
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		if v != "date" && v != "time" && v != "datetime" {
			c.diagAt(a.span, "mode must be one of: date, time, datetime")
		}
		return
	}
	if tag == "datepicker" && a.key == "variant" {
		if a.hasExpr {
			c.diagAt(a.span, "variant must be one of: native, calendar")
			return
		}
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		if v != "native" && v != "calendar" {
			c.diagAt(a.span, "variant must be one of: native, calendar")
		}
		return
	}
	if tag == "spinner" && a.key == "variant" {
		if a.hasExpr {
			c.diagAt(a.span, "variant must be one of: ring-track, orbit-dots, multiple-dots, clessidra, hourglass, fill-wave, fill-flip")
			return
		}
		v := strings.TrimSpace(partsContent(a.text))
		if v != "ring-track" && v != "orbit-dots" && v != "multiple-dots" && v != "clessidra" &&
			v != "hourglass" && v != "fill-wave" && v != "fill-flip" {
			c.diagAt(a.span, "variant must be one of: ring-track, orbit-dots, multiple-dots, clessidra, hourglass, fill-wave, fill-flip")
		}
		return
	}
	if a.key == "pickMode" {
		if a.hasExpr {
			c.diagAt(a.span, "pickMode must be one of: native, fs, auto")
			return
		}
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		if v != "native" && v != "fs" && v != "auto" {
			c.diagAt(a.span, "pickMode must be one of: native, fs, auto")
		}
		return
	}
	if a.key == "draggable" || a.key == "directory" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "bool" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, a.key+" must be a bool, got "+typeStr(t))
			}
		}
		return
	}
	if a.key == "maxOptionLength" || a.key == "maxOptions" {
		if a.hasExpr {
			t := c.inferExpr(a.expr)
			if t.kind != "int" && t.kind != "any" && t.kind != "unknown" && t.kind != "error" {
				c.diagAt(a.span, a.key+" must be an integer, got "+typeStr(t))
			}
		} else {
			if n, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text))); err != nil || n < 0 {
				c.diagAt(a.span, a.key+" must be a non-negative integer")
			}
		}
		return
	}
	if a.key == "dragKey" || a.key == "dropKey" {
		if a.hasExpr {
			_ = c.inferExpr(a.expr)
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	if a.key == "style" {
		// Dynamic CSS is a security boundary: the default policy requires
		// cssSafe(...). A project may relax this to warn/allow explicitly.
		if a.hasExpr {
			if !isCssSafeCall(a.expr) {
				c.securityDiag(a.span, c.security.StyleDynamic,
					"dynamic style: must be wrapped in cssSafe(...) (untrusted CSS)")
			}
			_ = c.inferExpr(a.expr)
			return
		}
		for i := 0; i < len(a.text); i++ {
			if a.text[i].kind == "expr" {
				sp := a.span
				if a.text[i].expr != nil {
					sp = a.text[i].expr.span
				}
				c.securityDiag(sp, c.security.StyleDynamic,
					"style interpolation requires cssSafe(...) (untrusted CSS)")
			}
		}
		if lit, ok := staticPartsLiteral(a.text); ok {
			if msg := cssExternalReferenceError(lit); msg != "" {
				c.diagAt(a.span, msg)
			}
		}
		c.checkViewParts(a.text)
		return
	}
	if a.key == "href" || a.key == "src" || a.key == "link" {
		// Iframe src has a stricter validator in checkIframe (http(s), relative
		// and about:blank only) with a "cannot accept" diagnostic.
		if !(tag == "iframe" && a.key == "src") {
			if lit, ok := staticPartsLiteral(a.text); ok && hasActiveURLScheme(lit) {
				// Image data URIs are inert inside <img> (scripts do not run).
				// The designer asset manager embeds uploaded GIF/PNG/JPEG
				// files this way, so an image spinner needs them.
				imageSource := a.key == "src" && (tag == "img" || tag == "spinner") && isImageDataURL(lit)
				if !imageSource {
					c.securityDiag(a.span, c.security.StaticURLSchemes,
						"active URL scheme is not allowed in "+a.key+" (use http(s), mailto, tel or a relative URL)")
				}
			}
		}
	}
	if a.key == "class" || a.key == "placeholder" || a.key == "alt" || a.key == "for" {
		if a.hasExpr {
			_ = c.inferExpr(a.expr)
		} else {
			c.checkViewParts(a.text)
		}
		return
	}
	// value / src / href: loosely-typed expression (or text for href/src).
	if a.hasExpr {
		_ = c.inferExpr(a.expr)
	} else {
		c.checkViewParts(a.text)
	}
}

// securityDiag reports a security-policy violation as an error ("error"/"safe")
// or a warning ("warn"), according to simpleui.security.json. "allow" is silent.
func (c *Checker) securityDiag(sp Span, mode, msg string) {
	switch mode {
	case "error", "safe":
		c.diagAt(sp, msg)
	case "warn":
		c.dd.AddEx(sp.file, sp.line, sp.col, "warning", msg,
			"security policy relaxed in simpleui.security.json")
	}
}

func isCssSafeCall(e *Expr) bool {
	return e != nil && e.kind == "call" && e.target != nil &&
		e.target.kind == "ident" && e.target.name == "cssSafe"
}

// staticPartsLiteral returns the literal value of an interpolated string when
// it contains no expressions.
func staticPartsLiteral(parts []StrPart) (string, bool) {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		if parts[i].kind != "text" {
			return "", false
		}
		sb.WriteString(parts[i].text)
	}
	return sb.String(), true
}

// hasActiveURLScheme reports whether a literal URL uses a scheme that can
// execute code. Whitespace/control characters are ignored before the check.
func hasActiveURLScheme(s string) bool {
	var probe strings.Builder
	for _, r := range s {
		if r > ' ' {
			probe.WriteRune(r)
		}
	}
	low := strings.ToLower(probe.String())
	return strings.HasPrefix(low, "javascript:") ||
		strings.HasPrefix(low, "vbscript:") ||
		strings.HasPrefix(low, "data:")
}

// isImageDataURL reports whether a literal is an inline image data URI. Such a
// URI is safe in an <img> (or an image spinner) because scripts are disabled in
// that context; it is still rejected in href/link and for non-image media.
func isImageDataURL(s string) bool {
	var probe strings.Builder
	for _, r := range s {
		if r > ' ' {
			probe.WriteRune(r)
		}
	}
	return strings.HasPrefix(strings.ToLower(probe.String()), "data:image/")
}

func viewParamNameList(params []Param) []string {
	var out []string
	for i := 0; i < len(params); i++ {
		out = append(out, params[i].name)
	}
	return out
}

func (c *Checker) checkChildren(e *Element) {
	if !c.viewIsComp {
		c.diagAt(e.span, "children is only valid inside a component view (a view invoked by another view)")
	}
	if c.childrenSeen {
		c.diagAt(e.span, "children can only appear once in a component")
	}
	c.childrenSeen = true
	for i := 0; i < len(e.attrs); i++ {
		if e.attrs[i].key != "class" {
			c.diagAt(e.attrs[i].span, "children cannot take attribute \""+e.attrs[i].key+"\"")
		}
	}
	if len(e.text) > 0 {
		c.diagAt(e.span, "children cannot take text")
	}
	if len(e.children) > 0 {
		c.diagAt(e.span, "children cannot take nested content")
	}
}

func (c *Checker) checkComponentCall(e *Element, viewName string) {
	params := c.viewParams[viewName]
	if params == nil {
		c.diagAt(e.span, "unknown component "+viewName)
		return
	}
	if len(e.text) > 0 {
		c.diagAt(e.span, "component "+viewName+" does not take text; pass values as parameters")
	}
	// The caller block (if any) is the component's `children` slot.
	vars := typeParamSet(c.viewTypeParams[viewName])
	declared := map[string]*SType{}
	for i := 0; i < len(params); i++ {
		if params[i].typename == "fn" {
			declared[params[i].name] = NewSType("fn")
		} else if pt, perr := parseSUITypeWith(params[i].typename, vars); perr == "" {
			declared[params[i].name] = pt
		}
	}
	// First pass: infer generic type parameters from the provided values.
	subst := map[string]*SType{}
	provided := map[string]bool{}
	argTypes := map[string]*SType{}
	for i := 0; i < len(e.attrs); i++ {
		a := &e.attrs[i]
		pt := declared[a.key]
		if pt == nil || pt.kind == "fn" {
			continue
		}
		provided[a.key] = true
		if a.hasExpr && a.expr != nil {
			vt := c.inferExpr(a.expr)
			argTypes[a.key] = vt
			_ = c.unifyType(pt, vt, subst)
		} else {
			argTypes[a.key] = NewSType("string")
			_ = c.unifyType(pt, NewSType("string"), subst)
		}
	}
	var missingTP []string
	for _, tp := range c.viewTypeParams[viewName] {
		if subst[tp] == nil {
			missingTP = append(missingTP, tp)
		}
	}
	if len(missingTP) > 0 {
		c.diagAt(e.span, "cannot infer generic type parameter "+strings.Join(missingTP, ", ")+" for component "+viewName+" (use it in a parameter type)")
		for _, tp := range missingTP {
			subst[tp] = NewSType("any")
		}
	}
	// Second pass: validate each provided value against the substituted type.
	provided = map[string]bool{}
	for i := 0; i < len(e.attrs); i++ {
		a := &e.attrs[i]
		pt := declared[a.key]
		if pt == nil {
			hint := ""
			if cand := nearestName(a.key, viewParamNameList(params)); cand != "" {
				hint = "did you mean \"" + cand + "\"? "
			}
			c.diagAtEx(a.span, "unknown parameter \""+a.key+"\" for component "+viewName, hint)
			continue
		}
		provided[a.key] = true
		if pt.kind == "fn" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "ident" && c.sigs[a.expr.name] != nil {
				continue
			}
			c.diagAt(a.span, "parameter \""+a.key+"\" of component "+viewName+" expects a function name (a declared fn)")
			continue
		}
		want := substituteType(pt, subst)
		if !a.hasExpr {
			c.checkViewParts(a.text)
		}
		vt := argTypes[a.key]
		if vt == nil {
			vt = NewSType("string")
		}
		if vt.kind != "error" && !assignable(vt, want) {
			c.diagAt(a.span, "parameter \""+a.key+"\" of component "+viewName+" expects "+typeStr(want)+", got "+valueTypeStr(vt, want))
		}
	}
	for i := 0; i < len(params); i++ {
		if params[i].name == "" || provided[params[i].name] {
			continue
		}
		c.diagAt(e.span, "missing parameter \""+params[i].name+"\" for component "+viewName)
	}
}

func (c *Checker) checkComponentCycles(prog *Prog) {
	if len(prog.views) == 0 {
		return
	}
	adj := map[string][]string{}
	spans := map[string]Span{}
	for i := 0; i < len(prog.views); i++ {
		spans[prog.views[i].name] = prog.views[i].span
		var kids []string
		collectComponentNames(prog.views[i].body, &kids, c.viewParams)
		adj[prog.views[i].name] = kids
	}
	state := map[string]int{}
	var pathStack []string
	for i := 0; i < len(prog.views); i++ {
		name := prog.views[i].name
		if state[name] == 0 {
			c.dfsComponentCycle(name, adj, spans, state, pathStack)
		}
	}
}

func (c *Checker) dfsComponentCycle(name string, adj map[string][]string, spans map[string]Span, state map[string]int, path []string) {
	if state[name] == 2 {
		return
	}
	if state[name] == 1 {
		idx := 0
		for idx < len(path) && path[idx] != name {
			idx++
		}
		loop := strings.Join(path[idx:], " -> ")
		sp := spans[name]
		if sp.file == "" {
			c.dd.Add("", 0, 0, "circular component reference: "+loop+" (starts and ends at "+name+")")
		} else {
			c.diagAt(sp, "circular component reference: "+loop+" (starts and ends at "+name+")")
		}
		return
	}
	state[name] = 1
	kids := adj[name]
	np := append(path, name)
	for i := 0; i < len(kids); i++ {
		c.dfsComponentCycle(kids[i], adj, spans, state, np)
	}
	state[name] = 2
}

func collectComponentNames(nodes []ViewNode, out *[]string, viewParams map[string][]Param) {
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		if n.kind == "element" && n.elem != nil {
			if viewParams[n.elem.tag] != nil {
				*out = append(*out, n.elem.tag)
			}
			collectComponentNames(n.elem.children, out, viewParams)
		} else if n.kind == "if" {
			collectComponentNames(derefNodeList(n.then_), out, viewParams)
			collectComponentNames(derefNodeList(n.else_), out, viewParams)
		} else if n.kind == "list" {
			collectComponentNames(derefNodeList(n.body), out, viewParams)
		}
	}
}

// ---------------------------------------------------------------------------
// ViewStateRefs
// ---------------------------------------------------------------------------

type RefAcc struct {
	names map[string]bool
}

func collectPartsRefs(parts []StrPart, acc *RefAcc) {
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "expr" && parts[i].expr != nil {
			collectExprRefs(parts[i].expr, acc)
		}
	}
}

func collectExprRefs(e *Expr, acc *RefAcc) {
	if e == nil {
		return
	}
	k := e.kind
	if k == "ident" {
		acc.names[e.name] = true
		return
	}
	if k == "str" {
		collectPartsRefs(e.parts, acc)
		return
	}
	if k == "member" {
		collectExprRefs(e.obj, acc)
		return
	}
	if k == "call" {
		collectExprRefs(e.target, acc)
		for i := 0; i < len(e.args); i++ {
			collectExprRefs(e.args[i], acc)
		}
		return
	}
	if k == "array" {
		for i := 0; i < len(e.elements); i++ {
			collectExprRefs(e.elements[i], acc)
		}
		return
	}
	if k == "map" || k == "structlit" {
		for i := 0; i < len(e.vals); i++ {
			collectExprRefs(e.vals[i], acc)
		}
		return
	}
	if k == "index" {
		collectExprRefs(e.obj, acc)
		collectExprRefs(e.idx, acc)
		return
	}
	if k == "unary" {
		collectExprRefs(e.operand, acc)
		return
	}
	if k == "binary" {
		collectExprRefs(e.l, acc)
		collectExprRefs(e.r, acc)
		return
	}
	if k == "range" {
		collectExprRefs(e.lo, acc)
		collectExprRefs(e.hi, acc)
	}
}

func collectNodeBodyRefs(nodes []ViewNode, acc *RefAcc) {
	for i := 0; i < len(nodes); i++ {
		collectNodeRefs(&nodes[i], acc)
	}
}

func collectNodeRefs(n *ViewNode, acc *RefAcc) {
	if n == nil {
		return
	}
	if n.kind == "element" {
		e := n.elem
		if e == nil {
			return
		}
		collectPartsRefs(e.text, acc)
		for i := 0; i < len(e.attrs); i++ {
			if e.attrs[i].hasExpr {
				collectExprRefs(e.attrs[i].expr, acc)
			} else {
				collectPartsRefs(e.attrs[i].text, acc)
			}
		}
		collectNodeBodyRefs(e.children, acc)
		return
	}
	if n.kind == "if" {
		collectExprRefs(n.cond, acc)
		collectNodeBodyRefs(derefNodeList(n.then_), acc)
		collectNodeBodyRefs(derefNodeList(n.else_), acc)
		return
	}
	if n.kind == "list" {
		collectExprRefs(n.iterable, acc)
		collectNodeBodyRefs(derefNodeList(n.body), acc)
	}
}

// ViewStateRefs returns the sorted, unique names of the state variables
// referenced anywhere in the view: text interpolation, attribute
// interpolation and expressions, if/else conditions, list iterables and
// nested bodies. It resolves idents against the state names of the most
// recent Check run.
func ViewStateRefs(v *ViewDecl) []string {
	var acc RefAcc
	acc.names = map[string]bool{}
	if v != nil {
		collectPartsRefs(v.title, &acc)
		collectNodeBodyRefs(v.body, &acc)
	}
	var out []string
	for name, _ := range acc.names {
		if KnownStateNames[name] {
			out = append(out, name)
		}
	}
	return sortedStrings(out)
}
