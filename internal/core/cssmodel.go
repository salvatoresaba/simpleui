package core

import "fmt"
import "sort"
import "strconv"
import "strings"

// cssmodel.go - a structured CSS engine for the design tool. CSS is NOT
// treated as a string: it is parsed into an AST (CssNode), edited as data and
// re-serialized on demand. The parser is tolerant: malformed rules are
// isolated (recovery) and reported as non-blocking diagnostics, so a broken
// stylesheet never crashes the tool or loses the rest of the document.

// CssSpan is a 1-based position in the CSS source.
type CssSpan struct {
	line int
	col  int
}

// CssDiag is a recoverable problem in the CSS source.
type CssDiag struct {
	span CssSpan
	msg  string
}

// CssDecl is a single `property: value [!important]`.
type CssDecl struct {
	prop      string
	value     string
	important bool
	span      CssSpan
}

// CssNode is one node of the stylesheet: a rule (`kind == "rule"`) or an
// at-rule such as @media/@keyframes (`kind == "at"`). At-rules may nest.
type CssNode struct {
	kind      string
	selectors []string  // for rules
	decls     []CssDecl // for rules
	name      string    // for at-rules, includes '@'
	prelude   string    // for at-rules, the part before '{' (may be empty)
	nodes     []*CssNode
	span      CssSpan
}

// CssSheet is the whole stylesheet plus any diagnostics collected while
// parsing.
type CssSheet struct {
	nodes []*CssNode
	diags []CssDiag
}

// CssProblem is a report from AnalyzeCss: an unused class, an overridden
// declaration, or a duplicate selector.
type CssProblem struct {
	kind    string // "unused" | "override" | "duplicate" | "syntax"
	message string
	line    int
	sel     string
	winner  string // for "override": the selector that wins
	reason  string // for "override": why it wins
}

// ---------------------------------------------------------------------------
// Parser (text -> AST)
// ---------------------------------------------------------------------------

type cssParser struct {
	src   string
	i     int
	line  int
	col   int
	diags []CssDiag
}

// ParseCss parses a stylesheet into a CssSheet, recovering from errors.
func ParseCss(src string) *CssSheet {
	p := &cssParser{src: src, line: 1, col: 1}
	s := &CssSheet{}
	s.nodes = p.parseNodes(false)
	s.diags = p.diags
	return s
}

func (p *cssParser) eof() bool { return p.i >= len(p.src) }

func (p *cssParser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.src[p.i]
}

func (p *cssParser) peekAt(off int) byte {
	if p.i+off >= len(p.src) {
		return 0
	}
	return p.src[p.i+off]
}

func (p *cssParser) span() CssSpan { return CssSpan{line: p.line, col: p.col} }

func (p *cssParser) adv() byte {
	c := p.src[p.i]
	p.i++
	if c == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
	return c
}

func (p *cssParser) errf(sp CssSpan, format string, args ...interface{}) {
	p.diags = append(p.diags, CssDiag{span: sp, msg: fmt.Sprintf(format, args...)})
}

func (p *cssParser) skipWsComments() {
	for !p.eof() {
		c := p.peek()
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			p.adv()
			continue
		}
		if c == '/' && p.peekAt(1) == '*' {
			p.adv()
			p.adv()
			for !p.eof() {
				if p.peek() == '*' && p.peekAt(1) == '/' {
					p.adv()
					p.adv()
					break
				}
				p.adv()
			}
			continue
		}
		break
	}
}

// parseNodes reads a sequence of rules/at-rules until '}' (when closing a
// block) or EOF.
func (p *cssParser) parseNodes(closing bool) []*CssNode {
	var nodes []*CssNode
	for {
		p.skipWsComments()
		if p.eof() {
			return nodes
		}
		if p.peek() == '}' {
			if closing {
				p.adv()
			}
			return nodes
		}
		if p.peek() == '@' {
			if n := p.parseAtRule(); n != nil {
				nodes = append(nodes, n)
			}
		} else {
			if n := p.parseRule(); n != nil {
				nodes = append(nodes, n)
			}
		}
	}
}

func (p *cssParser) parseAtRule() *CssNode {
	start := p.span()
	p.adv() // '@'
	var name strings.Builder
	for !p.eof() {
		c := p.peek()
		if isCssIdentByte(c) {
			name.WriteByte(p.adv())
		} else {
			break
		}
	}
	if name.Len() == 0 {
		p.errf(start, "expected an at-rule name after '@'")
		p.recoverLine()
		return nil
	}
	prelude := p.readPrelude()
	node := &CssNode{kind: "at", name: "@" + name.String(), prelude: prelude, span: start}
	if p.peek() == '{' {
		p.adv()
		node.nodes = p.parseNodes(true)
	} else if p.peek() == ';' {
		p.adv()
	} else if !p.eof() {
		// e.g. a malformed at-rule; recover to next ';' or '}'
		p.recoverDecl()
	}
	return node
}

// readPrelude consumes the text between the at-rule name and the opening '{'
// or terminating ';', respecting quotes/parens/brackets.
func (p *cssParser) readPrelude() string {
	var sb strings.Builder
	depth := 0
	quote := byte(0)
	for !p.eof() {
		c := p.peek()
		if quote != 0 {
			sb.WriteByte(p.adv())
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			sb.WriteByte(p.adv())
			continue
		}
		if c == '(' || c == '[' {
			depth++
			sb.WriteByte(p.adv())
			continue
		}
		if c == ')' || c == ']' {
			if depth > 0 {
				depth--
			}
			sb.WriteByte(p.adv())
			continue
		}
		if depth == 0 && (c == '{' || c == ';') {
			break
		}
		sb.WriteByte(p.adv())
	}
	return strings.TrimSpace(sb.String())
}

func (p *cssParser) parseRule() *CssNode {
	start := p.span()
	raw := p.readUntilBrace()
	if p.peek() != '{' {
		p.errf(start, "expected '{' to open a rule (found a malformed selector or stray text)")
		p.recoverLine()
		return nil
	}
	selectors := splitCssSelectors(raw)
	if len(selectors) == 0 {
		p.errf(start, "empty selector list")
	}
	p.adv() // '{'
	decls := p.parseDecls()
	node := &CssNode{kind: "rule", selectors: selectors, decls: decls, span: start}
	return node
}

// readUntilBrace consumes raw text up to an opening '{' at top level.
func (p *cssParser) readUntilBrace() string {
	var sb strings.Builder
	depth := 0
	quote := byte(0)
	for !p.eof() {
		c := p.peek()
		if quote != 0 {
			sb.WriteByte(p.adv())
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			sb.WriteByte(p.adv())
			continue
		}
		if c == '(' || c == '[' {
			depth++
			sb.WriteByte(p.adv())
			continue
		}
		if c == ')' || c == ']' {
			if depth > 0 {
				depth--
			}
			sb.WriteByte(p.adv())
			continue
		}
		if depth == 0 && c == '{' {
			break
		}
		if depth == 0 && c == '}' {
			break
		}
		sb.WriteByte(p.adv())
	}
	return strings.TrimSpace(sb.String())
}

func (p *cssParser) parseDecls() []CssDecl {
	var decls []CssDecl
	for {
		p.skipWsComments()
		if p.eof() || p.peek() == '}' {
			if p.peek() == '}' {
				p.adv()
			}
			return decls
		}
		if p.peek() == ';' {
			p.adv()
			continue
		}
		start := p.span()
		prop := p.readIdent()
		if prop == "" {
			p.errf(start, "expected a property name")
			p.recoverDecl()
			continue
		}
		p.skipWsComments()
		if p.peek() != ':' {
			p.errf(start, "expected ':' after property %q", prop)
			p.recoverDecl()
			continue
		}
		p.adv() // ':'
		p.skipWsComments()
		value := p.readValue()
		important := false
		trimmed := strings.TrimSpace(value)
		if strings.HasSuffix(trimmed, "!important") {
			important = true
			trimmed = strings.TrimSpace(strings.TrimSuffix(trimmed, "!important"))
		}
		decls = append(decls, CssDecl{prop: prop, value: trimmed, important: important, span: start})
		p.skipWsComments()
		if p.peek() == ';' {
			p.adv()
		} else if p.peek() == '}' {
			// last declaration, no trailing semicolon
		} else {
			p.errf(start, "expected ';' after declaration for %q", prop)
			p.recoverDecl()
		}
	}
}

// readValue consumes the value up to a top-level ';' or '}'.
func (p *cssParser) readValue() string {
	var sb strings.Builder
	depth := 0
	quote := byte(0)
	for !p.eof() {
		c := p.peek()
		if quote != 0 {
			sb.WriteByte(p.adv())
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			sb.WriteByte(p.adv())
			continue
		}
		if c == '(' || c == '[' {
			depth++
			sb.WriteByte(p.adv())
			continue
		}
		if c == ')' || c == ']' {
			if depth > 0 {
				depth--
			}
			sb.WriteByte(p.adv())
			continue
		}
		if depth == 0 && (c == ';' || c == '}') {
			break
		}
		sb.WriteByte(p.adv())
	}
	return strings.TrimSpace(sb.String())
}

// readIdent reads a CSS identifier (property name): letters, digits, '-', '_'.
func (p *cssParser) readIdent() string {
	p.skipWsComments()
	var sb strings.Builder
	for !p.eof() {
		c := p.peek()
		if isCssIdentByte(c) {
			sb.WriteByte(p.adv())
		} else {
			break
		}
	}
	return sb.String()
}

// recoverDecl skips to the next ';' or '}' so the parser can continue.
func (p *cssParser) recoverDecl() {
	for !p.eof() {
		c := p.adv()
		if c == ';' || c == '}' {
			return
		}
	}
}

// recoverLine skips to the end of the current line.
func (p *cssParser) recoverLine() {
	for !p.eof() {
		c := p.adv()
		if c == '\n' {
			return
		}
		if c == '}' {
			return
		}
	}
}

func isCssIdentByte(c byte) bool {
	return c == '-' || c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func splitCssSelectors(raw string) []string {
	var out []string
	depth := 0
	quote := byte(0)
	start := 0
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			continue
		}
		if c == '(' || c == '[' {
			depth++
			continue
		}
		if c == ')' || c == ']' {
			depth--
			continue
		}
		if c == ',' && depth == 0 {
			out = append(out, strings.TrimSpace(raw[start:i]))
			start = i + 1
		}
	}
	last := strings.TrimSpace(raw[start:])
	if last != "" {
		out = append(out, last)
	}
	return out
}

// ---------------------------------------------------------------------------
// Generator (AST -> text)
// ---------------------------------------------------------------------------

// GenCss renders the sheet as cleanly formatted CSS.
func GenCss(s *CssSheet) string {
	if s == nil {
		return ""
	}
	var sb strings.Builder
	writeCssNodes(&sb, s.nodes, 0)
	return strings.TrimRight(sb.String(), "\n")
}

func writeCssNodes(sb *strings.Builder, nodes []*CssNode, ind int) {
	for _, n := range nodes {
		if n == nil {
			continue
		}
		writeIndent(sb, ind)
		if n.kind == "at" {
			sb.WriteString(n.name)
			if n.prelude != "" {
				sb.WriteString(" ")
				sb.WriteString(n.prelude)
			}
			if len(n.nodes) > 0 {
				sb.WriteString(" {\n")
				writeCssNodes(sb, n.nodes, ind+1)
				writeIndent(sb, ind)
				sb.WriteString("}\n")
			} else {
				sb.WriteString(";\n")
			}
			continue
		}
		// rule
		for i, sel := range n.selectors {
			if i > 0 {
				sb.WriteString(",\n")
				writeIndent(sb, ind)
			}
			sb.WriteString(sel)
		}
		sb.WriteString(" {\n")
		for _, d := range n.decls {
			writeIndent(sb, ind+1)
			sb.WriteString(d.prop)
			sb.WriteString(": ")
			sb.WriteString(d.value)
			if d.important {
				sb.WriteString(" !important")
			}
			sb.WriteString(";\n")
		}
		writeIndent(sb, ind)
		sb.WriteString("}\n")
	}
}

func writeIndent(sb *strings.Builder, ind int) {
	for i := 0; i < ind; i++ {
		sb.WriteString("  ")
	}
}

// ---------------------------------------------------------------------------
// Specificity
// ---------------------------------------------------------------------------

// CssSpecificity is the (a, b, c) triple: a = id count, b = class/attr/
// pseudo-class count, c = type/pseudo-element count.
type CssSpecificity struct {
	a, b, c int
}

func (s CssSpecificity) String() string {
	return strconv.Itoa(s.a) + "," + strconv.Itoa(s.b) + "," + strconv.Itoa(s.c)
}

// compareSpecificity returns >0 if s wins, <0 if o wins, 0 if equal.
func compareSpecificity(s, o CssSpecificity) int {
	if s.a != o.a {
		return s.a - o.a
	}
	if s.b != o.b {
		return s.b - o.b
	}
	return s.c - o.c
}

// Specificity computes the specificity of a single selector.
func Specificity(sel string) CssSpecificity {
	var sp CssSpecificity
	i := 0
	n := len(sel)
	for i < n {
		c := sel[i]
		if c == '#' {
			sp.a++
			i++
			i = skipCssIdent(sel, i)
			continue
		}
		if c == '.' {
			sp.b++
			i++
			i = skipCssIdent(sel, i)
			continue
		}
		if c == '[' {
			sp.b++
			i++
			for i < n && sel[i] != ']' {
				i++
			}
			if i < n {
				i++
			}
			continue
		}
		if c == ':' {
			if i+1 < n && sel[i+1] == ':' {
				sp.c++ // pseudo-element
				i += 2
			} else {
				sp.b++ // pseudo-class
				i++
			}
			i = skipCssIdent(sel, i)
			if i < n && sel[i] == '(' {
				depth := 0
				for i < n {
					if sel[i] == '(' {
						depth++
					} else if sel[i] == ')' {
						depth--
						if depth == 0 {
							i++
							break
						}
					}
					i++
				}
			}
			continue
		}
		if c == '*' {
			i++
			continue
		}
		if isCssIdentByte(c) {
			sp.c++ // type selector
			i = skipCssIdent(sel, i)
			continue
		}
		i++
	}
	return sp
}

func skipCssIdent(s string, i int) int {
	for i < len(s) && isCssIdentByte(s[i]) {
		i++
	}
	return i
}

// selectorClasses extracts the class names referenced in a selector.
func selectorClasses(sel string) []string {
	var out []string
	for i := 0; i < len(sel); i++ {
		if sel[i] == '.' {
			j := i + 1
			for j < len(sel) && isCssIdentByte(sel[j]) {
				j++
			}
			if j > i+1 {
				out = append(out, sel[i+1:j])
			}
			i = j - 1
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Analysis (dead code + overrides + tokens)
// ---------------------------------------------------------------------------

// CssToken is a design token declared as a custom property: --name: value.
type CssToken struct {
	name  string
	value string
	line  int
}

// ExtractTokens walks the sheet collecting `--name` custom properties.
func ExtractTokens(s *CssSheet) []CssToken {
	var toks []CssToken
	var walk func([]*CssNode)
	walk = func(nodes []*CssNode) {
		for _, n := range nodes {
			if n.kind == "rule" {
				for _, d := range n.decls {
					if strings.HasPrefix(d.prop, "--") {
						toks = append(toks, CssToken{name: d.prop, value: d.value, line: d.span.line})
					}
				}
			}
			if len(n.nodes) > 0 {
				walk(n.nodes)
			}
		}
	}
	if s != nil {
		walk(s.nodes)
	}
	sort.Slice(toks, func(i, j int) bool { return toks[i].name < toks[j].name })
	return toks
}

// cssRuleAt holds a flattened rule with its source position.
type cssRuleAt struct {
	node *CssNode
	idx  int    // global order
	ctx  string // at-rule context ("@media (max-width: 760px)"), empty at the top level
}

// flattenRules returns every rule in source order (including nested at-rules).
// Keyframe steps (0%, to, ...) are not selectors and are intentionally skipped.
// The at-rule context is kept: a rule inside a media query is a conditional
// refinement of the same selector, not a duplicate of the top-level one.
func flattenRules(s *CssSheet) []cssRuleAt {
	var out []cssRuleAt
	idx := 0
	var walk func([]*CssNode, string)
	walk = func(nodes []*CssNode, ctx string) {
		for _, n := range nodes {
			if n.kind == "rule" {
				out = append(out, cssRuleAt{node: n, idx: idx, ctx: ctx})
				idx++
			}
			if len(n.nodes) > 0 {
				if n.kind == "at" && strings.HasSuffix(strings.ToLower(n.name), "keyframes") {
					continue
				}
				childCtx := ctx
				if n.kind == "at" {
					childCtx = n.name
					if n.prelude != "" {
						childCtx = n.name + " " + n.prelude
					}
					if ctx != "" {
						childCtx = ctx + " / " + childCtx
					}
				}
				walk(n.nodes, childCtx)
			}
		}
	}
	if s != nil {
		walk(s.nodes, "")
	}
	return out
}

// onlyCustomProps reports whether a rule declares only CSS custom properties.
// Repeated :root blocks with custom properties are a normal layering technique,
// not a duplicate-selector mistake.
func onlyCustomProps(n *CssNode) bool {
	if n == nil || len(n.decls) == 0 {
		return false
	}
	for _, d := range n.decls {
		if !strings.HasPrefix(d.prop, "--") {
			return false
		}
	}
	return true
}

// sameSelectorOverlap reports whether two rules for the same selector set at
// least one of the same properties. Repeating a selector to add unrelated
// properties (for example media sizing after media colour rules) is normal.
func sameSelectorOverlap(a, b *CssNode) bool {
	if a == nil || b == nil {
		return false
	}
	props := map[string]bool{}
	for _, d := range a.decls {
		props[d.prop] = true
	}
	for _, d := range b.decls {
		if props[d.prop] {
			return true
		}
	}
	return false
}

// runtimeCssClass lists classes driven by the generated app runtime. They are
// not authored markup and should not be reported as unused.
func runtimeCssClass(cls string) bool {
	if strings.HasPrefix(cls, "sui-") {
		return true
	}
	switch cls {
	case "active", "selected", "open", "on", "off", "checked", "disabled", "enabled", "hidden", "show", "suitbl":
		return true
	}
	return false
}

// runtimeSelector reports whether a selector targets a runtime-generated class
// or state class. Such rules intentionally layer on top of base rules.
func runtimeSelector(sel string) bool {
	for _, cls := range selectorClasses(sel) {
		if runtimeCssClass(cls) {
			return true
		}
	}
	return false
}

// interactiveSelector reports whether a selector targets a state or generated
// pseudo-element. Hover/focus/checked/pseudo-element rules intentionally layer
// on top of the base rule; reporting them as overrides is noise.
func interactiveSelector(sel string) bool {
	if strings.Contains(sel, "::") {
		return true
	}
	for _, state := range []string{
		":hover", ":focus", ":focus-visible", ":focus-within", ":active",
		":checked", ":disabled", ":enabled", ":invalid", ":valid", ":target", ":open",
	} {
		if strings.Contains(sel, state) {
			return true
		}
	}
	return false
}

// refinementSelector reports whether a selector targets a structural
// pseudo-class (`:last-child`, `:nth-child(2)`, ...). Such a rule refines the
// base rule on purpose: the cascade is the expected outcome, not a mistake.
func refinementSelector(sel string) bool {
	for _, state := range []string{
		":first-child", ":last-child", ":only-child", ":first-of-type",
		":last-of-type", ":only-of-type", ":nth-child", ":nth-last-child",
		":nth-of-type", ":nth-last-of-type", ":not(", ":empty",
	} {
		if strings.Contains(sel, state) {
			return true
		}
	}
	return false
}

// AnalyzeCss reports problems: unused classes, overridden declarations,
// duplicate selectors, and any parse diagnostics.
func AnalyzeCss(s *CssSheet, usedClasses map[string]bool) []CssProblem {
	var probs []CssProblem
	if s == nil {
		return probs
	}
	for _, d := range s.diags {
		probs = append(probs, CssProblem{kind: "syntax", message: d.msg, line: d.span.line})
	}

	rules := flattenRules(s)
	seen := map[string]int{} // selector -> first rule idx

	// gather declarations per "target key" for override detection.
	type declRec struct {
		rule cssRuleAt
		decl CssDecl
		sel  string
	}
	var byProp []declRec

	for _, r := range rules {
		for _, sel := range r.node.selectors {
			key := strings.TrimSpace(sel)
			// The key carries the at-rule context: the same selector in
			// another context is a conditional refinement (a media query, a
			// print sheet), not a duplicate.
			seenKey := key + "\x00" + r.ctx
			if first, ok := seen[seenKey]; ok {
				layeredRoot := key == ":root" && onlyCustomProps(r.node) && onlyCustomProps(rules[first].node)
				if !layeredRoot && sameSelectorOverlap(r.node, rules[first].node) {
					probs = append(probs, CssProblem{kind: "duplicate", message: "selector " + key + " is defined more than once (see line " + strconv.Itoa(rules[first].node.span.line) + ")", line: r.node.span.line, sel: key})
				}
			} else {
				seen[seenKey] = r.idx
			}

			// unused classes. Runtime `sui-*` and state classes are generated
			// by the designer theme and may be used by controls not present
			// yet, so they are not actionable author warnings.
			for _, cls := range selectorClasses(sel) {
				if !usedClasses[cls] && !strings.HasPrefix(cls, "class") && !runtimeCssClass(cls) {
					probs = append(probs, CssProblem{kind: "unused", message: "class ." + cls + " is never used by any element", line: r.node.span.line, sel: key})
				}
			}

			for _, d := range r.node.decls {
				byProp = append(byProp, declRec{rule: r, decl: d, sel: key})
			}
		}
	}

	// override detection: for each declaration, find a LATER declaration of the
	// same property on a selector that can match the same element with >= specificity
	// (or !important).
	for i := range byProp {
		rec := byProp[i]
		sp := Specificity(rec.sel)
		for j := i + 1; j < len(byProp); j++ {
			other := byProp[j]
			if other.decl.prop != rec.decl.prop {
				continue
			}
			if interactiveSelector(rec.sel) || interactiveSelector(other.sel) {
				continue
			}
			if refinementSelector(rec.sel) || refinementSelector(other.sel) {
				continue
			}
			if runtimeSelector(rec.sel) || runtimeSelector(other.sel) {
				continue
			}
			// A declaration in another at-rule context (a media query, a print
			// sheet) does not override the base rule: it applies under its own
			// condition.
			if rec.rule.ctx != other.rule.ctx {
				continue
			}
			if !overlapSelectors(rec.sel, other.sel) {
				continue
			}
			// !important rules
			aImp := rec.decl.important
			bImp := other.decl.important
			var reason string
			if aImp && !bImp {
				continue // earlier important still wins
			}
			if !aImp && bImp {
				reason = "!important beats a normal declaration"
			} else {
				osp := Specificity(other.sel)
				cmp := compareSpecificity(osp, sp)
				if cmp > 0 {
					reason = "higher specificity (" + osp.String() + " > " + sp.String() + ")"
				} else if cmp == 0 {
					reason = "same specificity, later in the source wins"
				} else {
					continue // lower specificity does not override
				}
			}
			probs = append(probs, CssProblem{
				kind:    "override",
				message: rec.decl.prop + " (line " + strconv.Itoa(rec.decl.span.line) + ") is overridden by " + other.sel + " — " + reason,
				line:    rec.decl.span.line,
				sel:     rec.sel,
				winner:  other.sel,
				reason:  reason,
			})
			break
		}
	}

	return probs
}

// overlapSelectors reports whether two selectors may target the same element.
//
// The subject of a selector is its last compound (".card.active p" has the
// subject "p"): the classes of the ancestors only describe the context, not
// the element. Two rules conflict only when they share the context, can match
// the same kind of element and their subjects share a class whose sets are
// comparable.
//
// This keeps the useful warnings (".card" overridden by ".card.active") and
// drops the noise:
//   - mutually exclusive variants (".row.part-row" and ".row.file-row"),
//   - descendant refinements (".monitor" and ".monitor h1").
func overlapSelectors(a, b string) bool {
	if strings.TrimSpace(a) == strings.TrimSpace(b) {
		return true
	}
	ctxA, subjA := selectorParts(a)
	ctxB, subjB := selectorParts(b)
	if ctxA != ctxB {
		return false
	}
	typeA, typeB := selectorType(subjA), selectorType(subjB)
	if typeA != "" && typeB != "" && typeA != typeB {
		return false
	}
	classesA := selectorClasses(subjA)
	classesB := selectorClasses(subjB)
	shared := false
	for _, ca := range classesA {
		for _, cb := range classesB {
			if ca == cb {
				shared = true
			}
		}
	}
	if !shared {
		return false
	}
	if !classSetSubset(classesA, classesB) && !classSetSubset(classesB, classesA) {
		return false
	}
	return true
}

// selectorParts splits a selector into its ancestor context and its subject
// compound: ".card.active p" becomes (".card.active", "p").
func selectorParts(sel string) (context, subject string) {
	s := strings.TrimSpace(sel)
	depth := 0
	last := -1
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ' ', '\t', '\n':
			if depth == 0 {
				last = i
			}
		}
	}
	if last < 0 {
		return "", s
	}
	return strings.TrimSpace(s[:last]), strings.TrimSpace(s[last+1:])
}

// selectorType returns the element type of a compound selector ("p" for
// "p.card"); it is empty for a universal selector.
func selectorType(compound string) string {
	i := 0
	for i < len(compound) {
		c := compound[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			i++
			continue
		}
		break
	}
	return compound[:i]
}

// classSetSubset reports whether every class of a is also in b.
func classSetSubset(a, b []string) bool {
	for _, ca := range a {
		found := false
		for _, cb := range b {
			if ca == cb {
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
