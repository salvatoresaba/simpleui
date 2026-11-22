package core

import "fmt"
import "strconv"
import "strings"

// astjson.go — dumps an AST as pretty-printed JSON (SPEC §4: keys in struct
// field order with "kind" first for tagged nodes). Owned by Agent A
// (front-end). There is no json package in std, so this file is a small
// self-contained writer.

// ---------------------------------------------------------------------------
// Low-level writer helpers
// ---------------------------------------------------------------------------

func jsonIndent(sb *strings.Builder, ind int) {
	for i := 0; i < ind; i++ {
		sb.WriteString("  ")
	}
}

// writeJsonString writes s as a JSON string literal. The quote, backslash,
// newline and tab characters are escaped; other control bytes become \u00XX.
func writeJsonString(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	n := len(s)
	for i := 0; i < n; i++ {
		c := s[i]
		if c == '"' {
			sb.WriteString("\\\"")
		} else if c == '\\' {
			sb.WriteString("\\\\")
		} else if c == '\n' {
			sb.WriteString("\\n")
		} else if c == '\t' {
			sb.WriteString("\\t")
		} else if c == '\r' {
			sb.WriteString("\\r")
		} else if c < 0x20 {
			sb.WriteString(fmt.Sprintf("\\u00%02x", c))
		} else {
			sb.WriteByte(c)
		}
	}
	sb.WriteByte('"')
}

func jsonBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// wfield writes the start of a JSON object field: indentation, "name": .
func wfield(sb *strings.Builder, ind int, name string) {
	jsonIndent(sb, ind)
	sb.WriteString("\"")
	sb.WriteString(name)
	sb.WriteString("\": ")
}

// ---------------------------------------------------------------------------
// Node emitters. Each emitX(sb, val, ind) writes the full JSON value for val;
// `ind` is the indentation nesting at which the value starts.
// ---------------------------------------------------------------------------

func emitSpan(sb *strings.Builder, sp Span, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "file")
	writeJsonString(sb, sp.file)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "line")
	sb.WriteString(strconv.Itoa(sp.line))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "col")
	sb.WriteString(strconv.Itoa(sp.col))
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitStrPart(sb *strings.Builder, sp StrPart, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "kind")
	writeJsonString(sb, sp.kind)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "text")
	writeJsonString(sb, sp.text)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "expr")
	emitExpr(sb, sp.expr, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitStrPartArray(sb *strings.Builder, a []StrPart, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitStrPart(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitStringArray(sb *strings.Builder, a []string, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		writeJsonString(sb, a[i])
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitExprArray(sb *strings.Builder, a []*Expr, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitExpr(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitStmtArray(sb *strings.Builder, a []*Stmt, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitStmt(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitExpr(sb *strings.Builder, e *Expr, ind int) {
	if e == nil {
		sb.WriteString("null")
		return
	}
	sb.WriteString("{\n")
	wfield(sb, ind+1, "kind")
	writeJsonString(sb, e.kind)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "i")
	sb.WriteString(strconv.Itoa(e.i))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "f")
	sb.WriteString(fmt.Sprintf("%v", e.f))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "b")
	sb.WriteString(jsonBool(e.b))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "parts")
	emitStrPartArray(sb, e.parts, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, e.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "obj")
	emitExpr(sb, e.obj, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "memb")
	writeJsonString(sb, e.memb)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "target")
	emitExpr(sb, e.target, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "args")
	emitExprArray(sb, e.args, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "argNames")
	emitStringArray(sb, e.argNames, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "elements")
	emitExprArray(sb, e.elements, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "spreads")
	emitBoolArray(sb, e.spreads, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "keys")
	emitStringArray(sb, e.keys, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "vals")
	emitExprArray(sb, e.vals, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "idx")
	emitExpr(sb, e.idx, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "op")
	writeJsonString(sb, e.op)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "operand")
	emitExpr(sb, e.operand, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "l")
	emitExpr(sb, e.l, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "r")
	emitExpr(sb, e.r, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "cond")
	emitExpr(sb, e.cond, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "lo")
	emitExpr(sb, e.lo, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "hi")
	emitExpr(sb, e.hi, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, e.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitBoolArray(sb *strings.Builder, a []bool, ind int) {
	sb.WriteString("[")
	for i := 0; i < len(a); i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(jsonBool(a[i]))
	}
	sb.WriteString("]")
}

func emitStmt(sb *strings.Builder, st *Stmt, ind int) {
	if st == nil {
		sb.WriteString("null")
		return
	}
	sb.WriteString("{\n")
	wfield(sb, ind+1, "kind")
	writeJsonString(sb, st.kind)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, st.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "pattern")
	emitStringArray(sb, st.pattern, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "patternKind")
	writeJsonString(sb, st.patternKind)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "init")
	emitExpr(sb, st.init, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "targets")
	emitStringArray(sb, st.targets, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "rhs")
	emitExpr(sb, st.rhs, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "cond")
	emitExpr(sb, st.cond, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "then_")
	emitStmtArray(sb, st.then_, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "else_")
	emitStmtArray(sb, st.else_, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "item")
	writeJsonString(sb, st.item)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "iterable")
	emitExpr(sb, st.iterable, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "body")
	emitStmtArray(sb, st.body, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "cases")
	emitSwitchCaseArray(sb, st.cases, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "default_")
	emitStmtArray(sb, st.default_, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "hasDefault")
	sb.WriteString(jsonBool(st.hasDefault))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "value")
	emitExpr(sb, st.value, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, st.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitParam(sb *strings.Builder, pa Param, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, pa.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "typename")
	writeJsonString(sb, pa.typename)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "rest")
	sb.WriteString(jsonBool(pa.rest))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, pa.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitParamArray(sb *strings.Builder, a []Param, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitParam(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitFuncDecl(sb *strings.Builder, f FuncDecl, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, f.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "typeParams")
	emitStringArray(sb, f.typeParams, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "async")
	sb.WriteString(jsonBool(f.async))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "params")
	emitParamArray(sb, f.params, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "returnType")
	writeJsonString(sb, f.returnType)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "body")
	emitStmtArray(sb, f.body, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, f.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitFuncArray(sb *strings.Builder, a []FuncDecl, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitFuncDecl(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitStateDecl(sb *strings.Builder, s StateDecl, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, s.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "typename")
	writeJsonString(sb, s.typename)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "init")
	emitExpr(sb, s.init, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, s.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitStateArray(sb *strings.Builder, a []StateDecl, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitStateDecl(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitImportDecl(sb *strings.Builder, im ImportDecl, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "names")
	emitStringArray(sb, im.names, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "kind")
	writeJsonString(sb, im.kind)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "path")
	writeJsonString(sb, im.path)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, im.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitImportArray(sb *strings.Builder, a []ImportDecl, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitImportDecl(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitViewNode(sb *strings.Builder, v *ViewNode, ind int) {
	if v == nil {
		sb.WriteString("null")
		return
	}
	sb.WriteString("{\n")
	wfield(sb, ind+1, "kind")
	writeJsonString(sb, v.kind)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "elem")
	emitElement(sb, v.elem, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "cond")
	emitExpr(sb, v.cond, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "then_")
	emitViewNodeArray(sb, v.then_, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "else_")
	emitViewNodeArray(sb, v.else_, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "item")
	writeJsonString(sb, v.item)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "iterable")
	emitExpr(sb, v.iterable, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "body")
	emitViewNodeArray(sb, v.body, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, v.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitViewNodeArray(sb *strings.Builder, a []*ViewNode, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitViewNode(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitAttr(sb *strings.Builder, a Attr, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "key")
	writeJsonString(sb, a.key)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "text")
	emitStrPartArray(sb, a.text, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "hasExpr")
	sb.WriteString(jsonBool(a.hasExpr))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "expr")
	emitExpr(sb, a.expr, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, a.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitAttrArray(sb *strings.Builder, a []Attr, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitAttr(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitElement(sb *strings.Builder, e *Element, ind int) {
	if e == nil {
		sb.WriteString("null")
		return
	}
	sb.WriteString("{\n")
	wfield(sb, ind+1, "tag")
	writeJsonString(sb, e.tag)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "text")
	emitStrPartArray(sb, e.text, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "attrs")
	emitAttrArray(sb, e.attrs, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "children")
	emitViewNodeArray(sb, childPts(e.children), ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, e.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

// childPts converts a []ViewNode value slice into the []*ViewNode shape used
// by the JSON view-node arrays (values are stored in the AST as values, but
// the JSON dumper reuses the pointer-based view-node array writer).
func childPts(a []ViewNode) []*ViewNode {
	var r []*ViewNode
	for i := 0; i < len(a); i++ {
		r = append(r, &a[i])
	}
	return r
}

func emitViewDecl(sb *strings.Builder, v ViewDecl, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "kind")
	writeJsonString(sb, "view")
	sb.WriteString(",\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, v.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "typeParams")
	emitStringArray(sb, v.typeParams, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "params")
	emitParamArray(sb, v.params, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "inst_state")
	emitStateArray(sb, v.inst_state, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "methods")
	emitFuncArray(sb, v.methods, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "title")
	emitStrPartArray(sb, v.title, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "body")
	emitViewNodeArray(sb, childPts(v.body), ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, v.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitViewArray(sb *strings.Builder, a []ViewDecl, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitViewDecl(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitField(sb *strings.Builder, f Field, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, f.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "typename")
	writeJsonString(sb, f.typename)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "readonly")
	sb.WriteString(jsonBool(f.readonly))
	sb.WriteString(",\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, f.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitTypeDecl(sb *strings.Builder, td TypeDecl, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "kind")
	writeJsonString(sb, "type")
	sb.WriteString(",\n")
	wfield(sb, ind+1, "name")
	writeJsonString(sb, td.name)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "fields")
	sb.WriteString("[")
	for i := 0; i < len(td.fields); i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		emitField(sb, td.fields[i], ind+1)
	}
	sb.WriteString("],\n")
	wfield(sb, ind+1, "span")
	emitSpan(sb, td.span, ind+1)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

func emitTypeArray(sb *strings.Builder, a []TypeDecl, ind int) {
	if len(a) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	for i := 0; i < len(a); i++ {
		jsonIndent(sb, ind+1)
		emitTypeDecl(sb, a[i], ind+1)
		if i+1 < len(a) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	jsonIndent(sb, ind)
	sb.WriteString("]")
}

func emitI18nArray(sb *strings.Builder, decls []I18nDecl, ind int) {
	sb.WriteString("[")
	if len(decls) > 0 {
		sb.WriteString("\n")
	}
	for i := 0; i < len(decls); i++ {
		if i > 0 {
			sb.WriteString(",\n")
		}
		jsonIndent(sb, ind+1)
		sb.WriteString("{\n")
		wfield(sb, ind+2, "locale")
		writeJsonString(sb, decls[i].locale)
		sb.WriteString(",\n")
		wfield(sb, ind+2, "messages")
		sb.WriteString("[")
		if len(decls[i].messages) > 0 {
			sb.WriteString("\n")
		}
		for j := 0; j < len(decls[i].messages); j++ {
			if j > 0 {
				sb.WriteString(",\n")
			}
			m := decls[i].messages[j]
			jsonIndent(sb, ind+3)
			sb.WriteString("{\n")
			wfield(sb, ind+4, "key")
			writeJsonString(sb, m.key)
			if len(m.plural) > 0 {
				sb.WriteString(",\n")
				wfield(sb, ind+4, "plural")
				sb.WriteString("{\n")
				first := true
				for _, cat := range i18nPluralCategories {
					v, ok := m.plural[cat]
					if !ok {
						continue
					}
					if !first {
						sb.WriteString(",\n")
					}
					first = false
					jsonIndent(sb, ind+5)
					writeJsonString(sb, cat)
					sb.WriteString(": ")
					writeJsonString(sb, v)
				}
				sb.WriteString("\n")
				jsonIndent(sb, ind+4)
				sb.WriteString("}")
			} else {
				sb.WriteString(",\n")
				wfield(sb, ind+4, "text")
				writeJsonString(sb, m.text)
			}
			sb.WriteString("\n")
			jsonIndent(sb, ind+3)
			sb.WriteString("}")
		}
		if len(decls[i].messages) > 0 {
			sb.WriteString("\n")
			jsonIndent(sb, ind+2)
		}
		sb.WriteString("]\n")
		jsonIndent(sb, ind+1)
		sb.WriteString("}")
	}
	if len(decls) > 0 {
		sb.WriteString("\n")
		jsonIndent(sb, ind)
	}
	sb.WriteString("]")
}

func emitEnumArray(sb *strings.Builder, a []EnumDecl, ind int) {
	sb.WriteString("[")
	if len(a) > 0 {
		sb.WriteString("\n")
	}
	for i := 0; i < len(a); i++ {
		if i > 0 {
			sb.WriteString(",\n")
		}
		jsonIndent(sb, ind+1)
		sb.WriteString("{\n")
		wfield(sb, ind+2, "name")
		writeJsonString(sb, a[i].name)
		sb.WriteString(",\n")
		wfield(sb, ind+2, "members")
		emitStringArray(sb, a[i].members, ind+2)
		sb.WriteString("\n")
		jsonIndent(sb, ind+1)
		sb.WriteString("}")
	}
	if len(a) > 0 {
		sb.WriteString("\n")
		jsonIndent(sb, ind)
	}
	sb.WriteString("]")
}

func emitSwitchCaseArray(sb *strings.Builder, a []SwitchCase, ind int) {
	sb.WriteString("[")
	if len(a) > 0 {
		sb.WriteString("\n")
	}
	for i := 0; i < len(a); i++ {
		if i > 0 {
			sb.WriteString(",\n")
		}
		jsonIndent(sb, ind+1)
		sb.WriteString("{\n")
		wfield(sb, ind+2, "expr")
		emitExpr(sb, a[i].expr, ind+2)
		sb.WriteString(",\n")
		wfield(sb, ind+2, "body")
		emitStmtArray(sb, a[i].body, ind+2)
		sb.WriteString("\n")
		jsonIndent(sb, ind+1)
		sb.WriteString("}")
	}
	if len(a) > 0 {
		sb.WriteString("\n")
		jsonIndent(sb, ind)
	}
	sb.WriteString("]")
}

func emitComputedArray(sb *strings.Builder, a []ComputedDecl, ind int) {
	sb.WriteString("[")
	if len(a) > 0 {
		sb.WriteString("\n")
	}
	for i := 0; i < len(a); i++ {
		if i > 0 {
			sb.WriteString(",\n")
		}
		jsonIndent(sb, ind+1)
		sb.WriteString("{\n")
		wfield(sb, ind+2, "name")
		writeJsonString(sb, a[i].name)
		sb.WriteString(",\n")
		wfield(sb, ind+2, "type")
		writeJsonString(sb, a[i].typename)
		sb.WriteString(",\n")
		wfield(sb, ind+2, "init")
		emitExpr(sb, a[i].init, ind+2)
		sb.WriteString("\n")
		jsonIndent(sb, ind+1)
		sb.WriteString("}")
	}
	if len(a) > 0 {
		sb.WriteString("\n")
		jsonIndent(sb, ind)
	}
	sb.WriteString("]")
}

func emitConstArray(sb *strings.Builder, a []ConstDecl, ind int) {
	sb.WriteString("[")
	if len(a) > 0 {
		sb.WriteString("\n")
	}
	for i := 0; i < len(a); i++ {
		if i > 0 {
			sb.WriteString(",\n")
		}
		jsonIndent(sb, ind+1)
		sb.WriteString("{\n")
		wfield(sb, ind+2, "name")
		writeJsonString(sb, a[i].name)
		sb.WriteString(",\n")
		wfield(sb, ind+2, "type")
		writeJsonString(sb, a[i].typename)
		sb.WriteString(",\n")
		wfield(sb, ind+2, "init")
		emitExpr(sb, a[i].init, ind+2)
		sb.WriteString("\n")
		jsonIndent(sb, ind+1)
		sb.WriteString("}")
	}
	if len(a) > 0 {
		sb.WriteString("\n")
		jsonIndent(sb, ind)
	}
	sb.WriteString("]")
}

func emitProg(sb *strings.Builder, prog *Prog, ind int) {
	sb.WriteString("{\n")
	wfield(sb, ind+1, "imports")
	emitImportArray(sb, prog.imports, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "types")
	emitTypeArray(sb, prog.types, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "enums")
	emitEnumArray(sb, prog.enums, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "consts")
	emitConstArray(sb, prog.consts, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "computed")
	emitComputedArray(sb, prog.computed, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "states")
	emitStateArray(sb, prog.states, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "i18n")
	emitI18nArray(sb, prog.i18n, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "funcs")
	emitFuncArray(sb, prog.funcs, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "views")
	emitViewArray(sb, prog.views, ind+1)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "css")
	sb.WriteString("[")
	for i := 0; i < len(prog.css); i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		writeJsonString(sb, prog.css[i])
	}
	sb.WriteString("]")
	sb.WriteString(",\n")
	wfield(sb, ind+1, "file")
	writeJsonString(sb, prog.file)
	sb.WriteString(",\n")
	wfield(sb, ind+1, "dir")
	writeJsonString(sb, prog.dir)
	sb.WriteString("\n")
	jsonIndent(sb, ind)
	sb.WriteString("}")
}

// ProgToJson renders a Prog as 2-space-indented JSON.
func ProgToJson(prog *Prog) string {
	var sb strings.Builder
	emitProg(&sb, prog, 0)
	sb.WriteString("\n")
	return sb.String()
}

// AstJsonFromSource runs the whole front-end pipeline (lex + parse) and
// returns the AST as JSON plus any diagnostics.
func AstJsonFromSource(source string, file string) (string, DiagList) {
	prog, dl := ParseSource(source, file)
	return ProgToJson(prog), dl
}
