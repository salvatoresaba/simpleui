package main

import "strconv"

// migrate.go rewrites deprecated source syntax into its current equivalent.
// It runs after parsing and before type checking, so old .sui files keep
// compiling while the compiler, generators and runtime only implement the
// current parameters. The migration is intentionally silent: opening and
// saving a file in the designer writes the new syntax.
//
// Current migrations:
//   - table editable: true/false        -> editMode: "editable"/"none"
//   - table fixed: true                 -> cellText: "truncate"
//   - table cellText: "auto"            -> cellText: "truncate"
//   - chart/graph interactive           -> editMode (editable or persistent)
//   - chart/graph stateMode: "static"   -> editMode: "none"
//   - chart/graph tableEdit             -> editMode: "editable"
//   - chartType: "statemachine"         -> chartType: "state"
//   - chart with chartType "state"      -> graph control
func migrateLegacyProgram(prog *Prog) {
	for i := range prog.states {
		migrateLegacyExpr(prog.states[i].init)
	}
	for i := range prog.funcs {
		migrateLegacyStmts(prog.funcs[i].body)
	}
	for i := range prog.views {
		migrateLegacyStrParts(prog.views[i].title)
		migrateLegacyNodes(prog.views[i].body)
	}
}

func migrateLegacyStmts(stmts []*Stmt) {
	for _, s := range stmts {
		if s == nil {
			continue
		}
		migrateLegacyExpr(s.init)
		migrateLegacyExpr(s.rhs)
		migrateLegacyExpr(s.lvalue)
		migrateLegacyExpr(s.cond)
		migrateLegacyExpr(s.iterable)
		migrateLegacyExpr(s.value)
		for _, v := range s.values {
			migrateLegacyExpr(v)
		}
		migrateLegacyStmts(s.then_)
		migrateLegacyStmts(s.else_)
		migrateLegacyStmts(s.body)
	}
}

func migrateLegacyStrParts(parts []StrPart) {
	for i := range parts {
		if parts[i].kind == "expr" {
			migrateLegacyExpr(parts[i].expr)
		}
	}
}

func migrateLegacyExpr(e *Expr) {
	if e == nil {
		return
	}
	for i := range e.parts {
		if e.parts[i].kind == "expr" {
			migrateLegacyExpr(e.parts[i].expr)
		}
	}
	migrateLegacyExpr(e.obj)
	migrateLegacyExpr(e.target)
	for _, a := range e.args {
		migrateLegacyExpr(a)
	}
	for _, a := range e.elements {
		migrateLegacyExpr(a)
	}
	for _, a := range e.vals {
		migrateLegacyExpr(a)
	}
	migrateLegacyExpr(e.idx)
	migrateLegacyExpr(e.operand)
	migrateLegacyExpr(e.l)
	migrateLegacyExpr(e.r)
	migrateLegacyExpr(e.cond)
	migrateLegacyExpr(e.lo)
	migrateLegacyExpr(e.hi)
	if e.kind != "call" || e.target == nil {
		return
	}
	if e.target.kind == "ident" && e.target.name == "substr" {
		e.target.name = "substring"
	}
	if e.target.kind == "member" && e.target.obj != nil && e.target.obj.kind == "ident" &&
		e.target.obj.name == "timers" && e.target.memb == "clear" {
		e.target.memb = "cancel"
	}
}

func migrateLegacyNodes(nodes []ViewNode) {
	for i := range nodes {
		migrateLegacyNode(&nodes[i])
	}
}

func migrateLegacyNodePtrs(nodes []*ViewNode) {
	for _, n := range nodes {
		if n != nil {
			migrateLegacyNode(n)
		}
	}
}

func migrateLegacyNode(n *ViewNode) {
	if n.kind == "element" && n.elem != nil {
		migrateLegacyStrParts(n.elem.text)
		for i := range n.elem.attrs {
			migrateLegacyStrParts(n.elem.attrs[i].text)
			if n.elem.attrs[i].hasExpr {
				migrateLegacyExpr(n.elem.attrs[i].expr)
			}
		}
		migrateLegacyElement(n.elem)
		migrateLegacyNodes(n.elem.children)
		return
	}
	migrateLegacyExpr(n.cond)
	migrateLegacyExpr(n.iterable)
	migrateLegacyNodePtrs(n.then_)
	migrateLegacyNodePtrs(n.else_)
	migrateLegacyNodePtrs(n.body)
}

func migrateLegacyElement(e *Element) {
	switch e.tag {
	case "table":
		migrateLegacyTable(e)
	case "chart", "graph":
		migrateLegacyChart(e)
	case "tabs":
		migrateLegacyWorkbook(e)
	case "tree":
		migrateLegacyTree(e)
	case "treeitem":
		migrateLegacyTreeItem(e)
	}
}

func migrateLegacyTreeItem(e *Element) {
	if editable, ok := attrBoolValue(findAttr(e, "editable")); ok {
		setBoolAttr(e, "rename", editable)
	}
	removeAttr(e, "editable")
}

// migrateLegacyTree replaces the old tree-level icon/iconOpen defaults with the
// explicit leaf/parent icons introduced later.
func migrateLegacyTree(e *Element) {
	if mode, ok := attrText(findAttr(e, "editMode")); !ok || mode == "" {
		if editable, hasEditable := attrBoolValue(findAttr(e, "editable")); hasEditable {
			if editable {
				setLiteralAttr(e, "editMode", "editable")
			} else {
				setLiteralAttr(e, "editMode", "none")
			}
		}
	}
	removeAttr(e, "editable")
	if depth, ok := attrValueSource(findAttr(e, "editableDepth")); ok && depth != "" {
		if v, _ := attrValueSource(findAttr(e, "editDepth")); v == "" {
			if n, err := strconv.Atoi(depth); err == nil {
				setIntAttr(e, "editDepth", n)
			} else {
				setLiteralAttr(e, "editDepth", depth)
			}
		}
	}
	removeAttr(e, "editableDepth")

	icon, hasIcon := attrText(findAttr(e, "icon"))
	iconOpen, hasOpen := attrText(findAttr(e, "iconOpen"))
	removeAttr(e, "icon")
	removeAttr(e, "iconOpen")
	if hasIcon && icon != "" {
		if v, _ := attrText(findAttr(e, "iconLeaf")); v == "" {
			setLiteralAttr(e, "iconLeaf", icon)
		}
		if v, _ := attrText(findAttr(e, "iconParent")); v == "" {
			setLiteralAttr(e, "iconParent", icon)
		}
	}
	if hasOpen && iconOpen != "" {
		if v, _ := attrText(findAttr(e, "iconParentOpen")); v == "" {
			setLiteralAttr(e, "iconParentOpen", iconOpen)
		}
	}
}

// migrateLegacyWorkbook preserves the old data-entry behavior of a tabs
// control whose panels all contain a table. Sheets without an explicit mode
// become editable unless they declare a selection mode (which used to keep a
// sheet read-only).
func migrateLegacyWorkbook(e *Element) {
	var tables []*Element
	for i := range e.children {
		child := &e.children[i]
		if child.kind != "element" || child.elem == nil || child.elem.tag != "tab" {
			return
		}
		table := firstTableElement(&child.elem.children)
		if table == nil {
			return
		}
		tables = append(tables, table)
	}
	if len(tables) == 0 {
		return
	}
	for _, table := range tables {
		if a := findAttr(table, "editMode"); a != nil {
			continue
		}
		mode := "editable"
		if a := findAttr(table, "selection"); a != nil {
			if sel, ok := attrText(a); ok && sel != "" && sel != "none" {
				mode = "none"
			}
		}
		if editable, ok := attrBoolValue(findAttr(table, "editable")); ok && !editable {
			mode = "none"
		}
		setLiteralAttr(table, "editMode", mode)
	}
}

func firstTableElement(nodes *[]ViewNode) *Element {
	for i := range *nodes {
		n := &(*nodes)[i]
		if n.kind != "element" || n.elem == nil {
			continue
		}
		if n.elem.tag == "table" {
			return n.elem
		}
		if t := firstTableElement(&n.elem.children); t != nil {
			return t
		}
	}
	return nil
}

func migrateLegacyTable(e *Element) {
	existing, _ := attrText(findAttr(e, "editMode"))
	editable, hasEditable := attrBoolValue(findAttr(e, "editable"))
	if existing == "" && hasEditable {
		if editable {
			setLiteralAttr(e, "editMode", "editable")
		} else {
			setLiteralAttr(e, "editMode", "none")
		}
	}
	removeAttr(e, "editable")

	fixed, hasFixed := attrBoolValue(findAttr(e, "fixed"))
	removeAttr(e, "fixed")
	cellText, _ := attrText(findAttr(e, "cellText"))
	if cellText == "auto" || (cellText == "" && hasFixed && fixed) {
		setLiteralAttr(e, "cellText", "truncate")
	}
}

func migrateLegacyChart(e *Element) {
	chartType, _ := attrText(findAttr(e, "chartType"))
	chartType = lowerASCII(chartType)
	interactive, hasInteractive := attrBoolValue(findAttr(e, "interactive"))
	stateMode, _ := attrText(findAttr(e, "stateMode"))
	tableEdit, hasTableEdit := attrBoolValue(findAttr(e, "tableEdit"))
	existing, _ := attrText(findAttr(e, "editMode"))

	diagram := e.tag == "graph" || chartType == "state" || chartType == "statemachine"
	typeGantt := chartType == "gantt"

	if existing == "" {
		mode := "none"
		switch {
		case lowerASCII(stateMode) == "static":
			if hasTableEdit && tableEdit {
				mode = "editable"
			}
		case hasInteractive && interactive:
			if diagram || typeGantt {
				mode = "persistent"
			} else {
				mode = "editable"
			}
		case diagram:
			// Block diagrams persisted their geometry before editMode existed.
			mode = "persistent"
		case hasTableEdit && tableEdit:
			mode = "editable"
		}
		setLiteralAttr(e, "editMode", mode)
	}

	removeAttr(e, "interactive")
	removeAttr(e, "stateMode")
	removeAttr(e, "tableEdit")

	if e.tag == "chart" && (chartType == "state" || chartType == "statemachine") {
		// The dedicated graph control replaces the old state chart. The runtime
		// keeps its internal state renderer, but the source surface is graph.
		e.tag = "graph"
		removeAttr(e, "chartType")
		return
	}
	if chartType == "statemachine" {
		setLiteralAttr(e, "chartType", "state")
	}
}

// attrText returns the literal value of an attribute. Dynamic expressions return
// ok=false and are left untouched by the migration.
func attrText(a *Attr) (string, bool) {
	if a == nil {
		return "", false
	}
	if a.hasExpr {
		if a.expr != nil && a.expr.kind == "str" {
			return partsLiteral(a.expr.parts), true
		}
		return "", false
	}
	return partsLiteral(a.text), true
}

// attrBoolValue returns the literal boolean value of an attribute.
func attrBoolValue(a *Attr) (bool, bool) {
	if a == nil || !a.hasExpr || a.expr == nil || a.expr.kind != "bool" {
		return false, false
	}
	return a.expr.b, true
}

// attrValueSource returns the source text of a literal or scalar expression
// attribute (for example an integer depth).
func attrValueSource(a *Attr) (string, bool) {
	if a == nil {
		return "", false
	}
	if a.hasExpr && a.expr != nil {
		return exprSource(a.expr), true
	}
	return partsLiteral(a.text), true
}

func setIntAttr(e *Element, key string, value int) {
	if a := findAttr(e, key); a != nil {
		a.hasExpr = true
		a.text = nil
		a.expr = &Expr{kind: "int", i: value, span: a.span}
		return
	}
	e.attrs = append(e.attrs, Attr{
		key:     key,
		hasExpr: true,
		expr:    &Expr{kind: "int", i: value, span: e.span},
		span:    e.span,
	})
}

func setBoolAttr(e *Element, key string, value bool) {
	if a := findAttr(e, key); a != nil {
		a.hasExpr = true
		a.text = nil
		a.expr = &Expr{kind: "bool", b: value, span: a.span}
		return
	}
	e.attrs = append(e.attrs, Attr{
		key:     key,
		hasExpr: true,
		expr:    &Expr{kind: "bool", b: value, span: e.span},
		span:    e.span,
	})
}

// setLiteralAttr replaces an attribute with a plain string literal. It keeps the
// existing span so diagnostics still point at the original line.
func setLiteralAttr(e *Element, key, value string) {
	if a := findAttr(e, key); a != nil {
		a.hasExpr = false
		a.expr = nil
		a.text = []StrPart{NewTextPart(value)}
		return
	}
	e.attrs = append(e.attrs, Attr{
		key:  key,
		text: []StrPart{NewTextPart(value)},
		span: e.span,
	})
}

func removeAttr(e *Element, key string) {
	out := e.attrs[:0]
	for i := range e.attrs {
		if e.attrs[i].key != key {
			out = append(out, e.attrs[i])
		}
	}
	e.attrs = out
}

func lowerASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
