package core

import "fmt"
import "strconv"
import "strings"

// jsonutil.go - a tiny JSON value model + parser + serializer for the LSP
// protocol (there is no JSON package in the standard library).

type JVal struct {
	kind string // "null" "bool" "int" "float" "str" "obj" "arr"
	s    string
	i    int
	f    float64
	b    bool
	obj  map[string]JVal
	arr  []JVal
}

func JNull() JVal         { return JVal{kind: "null"} }
func JBool(b bool) JVal   { return JVal{kind: "bool", b: b} }
func JInt(i int) JVal     { return JVal{kind: "int", i: i} }
func JNum(f float64) JVal { return JVal{kind: "float", f: f} }
func JStr(s string) JVal  { return JVal{kind: "str", s: s} }
func JObj(vals ...interface{}) JVal {
	m := map[string]JVal{}
	for i := 0; i < len(vals); i += 2 {
		m[vals[i].(string)] = vals[i+1].(JVal)
	}
	return JVal{kind: "obj", obj: m}
}
func JArr(items ...interface{}) JVal {
	var a []JVal
	for i := 0; i < len(items); i++ {
		a = append(a, items[i].(JVal))
	}
	return JVal{kind: "arr", arr: a}
}

// get is a convenience object member lookup (returns JNull when missing).
func get(o JVal, key string) JVal {
	if o.kind != "obj" {
		return JNull()
	}
	return o.obj[key]
}

// pushArr appends to a JVal array, returning the (unboxed) value so callers
// can re-bind their variable (Go value semantics).
func pushArr(a JVal, v JVal) JVal {
	if a.kind != "arr" {
		return a
	}
	a.arr = append(a.arr, v)
	return a
}

// ParseJson converts a JSON document into a JVal.
func ParseJson(s string) (JVal, error) {
	state := &JsonParse{source: s, i: 0}
	state.skipWs()
	v, err := state.parseValue()
	if err != nil {
		return JNull(), err
	}
	return v, nil
}

type JsonParse struct {
	source string
	i      int
}

func (p *JsonParse) skipWs() {
	n := len(p.source)
	for p.i < n && (p.source[p.i] == ' ' || p.source[p.i] == '\t' || p.source[p.i] == '\n' || p.source[p.i] == '\r') {
		p.i++
	}
}

func (p *JsonParse) parseValue() (JVal, error) {
	p.skipWs()
	if p.i >= len(p.source) {
		return JNull(), fmt.Errorf("json: empty value")
	}
	c := p.source[p.i]
	if c == '{' {
		return p.parseObject()
	}
	if c == '[' {
		return p.parseArray()
	}
	if c == '"' {
		str, err := p.parseString()
		return JStr(str), err
	}
	if c == 't' {
		p.i += 4
		return JBool(true), nil
	}
	if c == 'f' {
		p.i += 5
		return JBool(false), nil
	}
	if c == 'n' {
		p.i += 4
		return JNull(), nil
	}
	if c == '-' || (c >= '0' && c <= '9') {
		return p.parseNumber()
	}
	return JNull(), fmt.Errorf("json: unexpected value")
}

func (p *JsonParse) parseObject() (JVal, error) {
	p.i++ // {
	res := JObj()
	first := true
	for {
		p.skipWs()
		if p.i >= len(p.source) {
			return JNull(), fmt.Errorf("json: unterminated object")
		}
		if p.source[p.i] == '}' {
			p.i++
			return res, nil
		}
		if !first {
			if p.source[p.i] != ',' {
				return JNull(), fmt.Errorf("json: expected ,")
			}
			p.i++
			p.skipWs()
		}
		first = false
		if p.source[p.i] != '"' {
			return JNull(), fmt.Errorf("json: expected string key")
		}
		key, kerr := p.parseString()
		if kerr != nil {
			return JNull(), kerr
		}
		p.skipWs()
		if p.i >= len(p.source) || p.source[p.i] != ':' {
			return JNull(), fmt.Errorf("json: expected :")
		}
		p.i++
		val, verr := p.parseValue()
		if verr != nil {
			return JNull(), verr
		}
		res.obj[key] = val
	}
}

func (p *JsonParse) parseArray() (JVal, error) {
	p.i++ // [
	res := JArr()
	first := true
	for {
		p.skipWs()
		if p.i >= len(p.source) {
			return JNull(), fmt.Errorf("json: unterminated array")
		}
		if p.source[p.i] == ']' {
			p.i++
			return res, nil
		}
		if !first {
			if p.source[p.i] != ',' {
				return JNull(), fmt.Errorf("json: expected ,")
			}
			p.i++
			p.skipWs()
		}
		first = false
		v, err := p.parseValue()
		if err != nil {
			return JNull(), err
		}
		res.arr = append(res.arr, v)
	}
}

func (p *JsonParse) parseString() (string, error) {
	p.i++ // "
	var sb strings.Builder
	n := len(p.source)
	for p.i < n {
		c := p.source[p.i]
		if c == '"' {
			p.i++
			return sb.String(), nil
		}
		if c == '\\' {
			p.i++
			if p.i >= n {
				return "", fmt.Errorf("json: bad escape")
			}
			e := p.source[p.i]
			if e == '"' {
				sb.WriteString("\"")
			} else if e == '\\' {
				sb.WriteString("\\")
			} else if e == '/' {
				sb.WriteString("/")
			} else if e == 'n' {
				sb.WriteString("\n")
			} else if e == 't' {
				sb.WriteString("\t")
			} else if e == 'r' {
				sb.WriteString("\r")
			} else {
				return "", fmt.Errorf("json: bad escape")
			}
		} else {
			sb.WriteByte(c)
		}
		p.i++
	}
	return "", fmt.Errorf("json: unterminated string")
}

func (p *JsonParse) parseNumber() (JVal, error) {
	start := p.i
	n := len(p.source)
	if p.source[p.i] == '-' {
		p.i++
	}
	for p.i < n && p.source[p.i] >= '0' && p.source[p.i] <= '9' {
		p.i++
	}
	isFloat := false
	if p.i < n && p.source[p.i] == '.' {
		isFloat = true
		p.i++
		for p.i < n && p.source[p.i] >= '0' && p.source[p.i] <= '9' {
			p.i++
		}
	}
	if p.i < n && (p.source[p.i] == 'e' || p.source[p.i] == 'E') {
		isFloat = true
		p.i++
		if p.i < n && (p.source[p.i] == '+' || p.source[p.i] == '-') {
			p.i++
		}
		for p.i < n && p.source[p.i] >= '0' && p.source[p.i] <= '9' {
			p.i++
		}
	}
	text := p.source[start:p.i]
	if isFloat {
		fv, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return JNum(0), err
		}
		return JNum(fv), nil
	}
	iv, err := strconv.Atoi(text)
	if err != nil {
		return JInt(0), err
	}
	return JInt(iv), nil
}

// StringifyJson serializes a JVal as compact JSON.
func StringifyJson(v JVal) string {
	var sb strings.Builder
	writeJson(&sb, v)
	return sb.String()
}

func writeJson(sb *strings.Builder, v JVal) {
	switch v.kind {
	case "null":
		sb.WriteString("null")
	case "bool":
		if v.b {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case "int":
		sb.WriteString(strconv.Itoa(v.i))
	case "float":
		sb.WriteString(fmt.Sprintf("%v", v.f))
	case "str":
		jsonStringOut(sb, v.s)
	case "obj":
		sb.WriteString("{")
		first := true
		for key, val := range v.obj {
			if !first {
				sb.WriteString(",")
			}
			first = false
			jsonStringOut(sb, key)
			sb.WriteString(":")
			writeJson(sb, val)
		}
		sb.WriteString("}")
	case "arr":
		sb.WriteString("[")
		for i := 0; i < len(v.arr); i++ {
			if i > 0 {
				sb.WriteString(",")
			}
			writeJson(sb, v.arr[i])
		}
		sb.WriteString("]")
	}
}

func jsonStringOut(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for i := 0; i < len(s); i++ {
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
		} else {
			sb.WriteByte(c)
		}
	}
	sb.WriteByte('"')
}

// asStr / asInt are accessor helpers for callers.
func asStr(v JVal) string {
	if v.kind == "str" {
		return v.s
	}
	return ""
}

func asInt(v JVal) int {
	if v.kind == "int" {
		return v.i
	}
	return 0
}
