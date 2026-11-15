package main

import "strings"
import "testing"

func jsContains(js string, frag string, t *testing.T) {
	if !strings.Contains(js, frag) {
		t.Errorf("jsgen output missing %q\n----full output----\n%v\n--------------------", frag, js)
	}
}

func countChar(s string, c byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

func jsBalanced(js string, t *testing.T) {
	if countChar(js, '`')%2 != 0 {
		t.Errorf("generated JS has an odd number of backticks")
	}
	if countChar(js, '{') != countChar(js, '}') {
		t.Errorf("generated JS brace mismatch: %v open, %v close", countChar(js, '{'), countChar(js, '}'))
	}
	if countChar(js, '(') != countChar(js, ')') {
		t.Errorf("generated JS paren mismatch: %v open, %v close", countChar(js, '('), countChar(js, ')'))
	}
	if countChar(js, '[') != countChar(js, ']') {
		t.Errorf("generated JS bracket mismatch: %v open, %v close", countChar(js, '['), countChar(js, ']'))
	}
}

func TestJsgenCounterApp(t *testing.T) {
	src := "state counter: int = 0\n" +
		"\n" +
		"fn increment() -> int {\n" +
		"\tcounter = counter + 1\n" +
		"\treturn counter\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\ttitle \"Count {counter}\"\n" +
		"\th1 \"Count {counter}\"\n" +
		"\tbutton \"Go\" onClick: increment\n" +
		"}\n"
	prog, dl := ParseSource(src, "counter.sui")
	if dl.Count() != 0 {
		t.Errorf("counter app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "\"use strict\";", t)
	jsContains(js, "let counter = 0;", t)
	jsContains(js, "function setCounter(v) { counter = v; __sui_react(\"counter\"); }", t)
	jsContains(js, "e1.onclick = increment;", t)
	jsContains(js, "__SUI_views[\"main\"] = render_main;", t)
	jsContains(js, "function __sui_replace_view(el, html) {", t)
	jsContains(js, "__sui_replace_view(el, __viewHtml);", t)
	jsContains(js, "__sui_bindView(\"counter\", \"main\");", t)
	jsContains(js, "${__sui_esc(counter)}", t)
	jsContains(js, "__sui_render();", t)
	jsBalanced(js, t)
}

func TestJsgenAsyncTupleHttp(t *testing.T) {
	src := "import { http } from \"std\"\n" +
		"\n" +
		"state result: string = \"\"\n" +
		"\n" +
		"async fn load() -> null {\n" +
		"\tbody, err = http.get(\"https://example.com\")\n" +
		"\tif err == null {\n" +
		"\t\tresult = body\n" +
		"\t}\n" +
		"}\n"
	prog, dl := ParseSource(src, "load.sui")
	if dl.Count() != 0 {
		t.Errorf("http app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "async function load() {", t)
	jsContains(js, "await __SUI_STD.http_get(", t)
	jsContains(js, "const body = __sui_r[0];", t)
	jsContains(js, "const err = __sui_r[1];", t)
	jsContains(js, "http_get: async (url) => {", t)
	jsContains(js, "function setResult(v) { result = v; __sui_react(\"result\"); }", t)
	jsContains(js, "(err === null)", t)
	jsBalanced(js, t)
}

func TestJsgenViewStructural(t *testing.T) {
	src := "state counter: int = 0\n" +
		"state items: array[string] = [\"a\", \"b\"]\n" +
		"\n" +
		"fn incr() -> int {\n" +
		"\tcounter = counter + 1\n" +
		"\treturn counter\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\th1 \"Top\"\n" +
		"\trow {\n" +
		"\t\tbutton \"Go\" onClick: incr\n" +
		"\t\tp \"row text\"\n" +
		"\t}\n" +
		"\tlist x in items {\n" +
		"\t\tp \"{x}\"\n" +
		"\t}\n" +
		"\tif counter > 5 {\n" +
		"\t\ttext \"big\"\n" +
		"\t} else {\n" +
		"\t\ttext \"small\"\n" +
		"\t}\n" +
		"}\n"
	prog, dl := ParseSource(src, "view.sui")
	if dl.Count() != 0 {
		t.Errorf("view app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "class=\"row\"", t)
	jsContains(js, "data-sui-id=\"0\"", t)
	jsContains(js, "e2.onclick = incr;", t)
	jsContains(js, "for (const x of (items))", t)
	jsContains(js, "((counter > 5) ?", t)
	jsContains(js, "__sui_bindView(\"counter\", \"main\");", t)
	jsContains(js, "__sui_bindView(\"items\", \"main\");", t)
	jsContains(js, "let items = [`a`, `b`];", t)
	jsBalanced(js, t)
}

func TestJsgenNoViews(t *testing.T) {
	src := "state counter: int = 0\n" +
		"\n" +
		"fn bump() -> int {\n" +
		"\tcounter = counter + 1\n" +
		"\treturn counter\n" +
		"}\n"
	prog, dl := ParseSource(src, "noview.sui")
	if dl.Count() != 0 {
		t.Errorf("no-views app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "\"use strict\";", t)
	jsContains(js, "__sui_render();", t)
	jsContains(js, "function setCounter(v) { counter = v; __sui_react(\"counter\"); }", t)
	if strings.Contains(js, "render_") {
		t.Errorf("no-views app must not emit a view renderer")
	}
	jsBalanced(js, t)
}

func TestJsgenWsInputBinding(t *testing.T) {
	src := "import { ws } from \"std\"\n" +
		"\n" +
		"state msg: string = \"\"\n" +
		"\n" +
		"view main {\n" +
		"\tinput \"Type here\" value: msg\n" +
		"}\n"
	prog, dl := ParseSource(src, "ws.sui")
	if dl.Count() != 0 {
		t.Errorf("ws app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "class __SUI_WS {", t)
	jsContains(js, "this.sock = new WebSocket(url);", t)
	jsContains(js, "onMessage(fn) { this.sock.onmessage = (ev) => fn(ev.data); }", t)
	jsContains(js, "ws_connect: async (url) => {", t)
	jsContains(js, "value=\"${__sui_esc(msg)}\"", t)
	jsContains(js, "placeholder=\"Type here\"", t)
	jsContains(js, "setMsg(ev.target.value);", t)
	// value: is a write-only binding when the state is not read elsewhere, so
	// the view is not rebound and typing does not re-render it.
	if strings.Contains(js, "__sui_bindView(\"msg\", \"main\");") {
		t.Errorf("input value binding must not rebind the view")
	}
	jsBalanced(js, t)
}

// TestJsgenSwitchCheckedBinding guards the fix for a switch/checkbox whose
// "checked" attribute was emitted as checked="${state}". "checked" is a boolean
// HTML attribute, so checked="false" still renders as ON and the control never
// visibly toggles. It must be emitted only when the expression is truthy.
func TestJsgenSwitchCheckedBinding(t *testing.T) {
	src := "state on = false\n" +
		"\n" +
		"view main {\n" +
		"\tswitch checked: on\n" +
		"}\n"
	prog, dl := ParseSource(src, "sw.sui")
	if dl.Count() != 0 {
		t.Errorf("switch app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	// The checked attribute must be conditional, never checked="${...}".
	jsContains(js, "${(on) ? \"checked\" : \"\"}", t)
	if strings.Contains(js, "checked=\"${") {
		t.Errorf("checked attribute must not be emitted as a plain boolean attr:\n%s", js)
	}
	// A lone "checked:" binding (no explicit handler) must still write back to
	// state: the control has to be treated as interactive.
	jsContains(js, "e0.onchange = (ev) => setOn(ev.target.checked);", t)
	jsBalanced(js, t)
}

// TestJsgenExplicitHandlerOwnsState checks that when the user declares an
// onChange/onInput handler, the automatic state setter is NOT emitted. Both
// running would fight each other: the setter coerces the bound state to
// ev.target.checked / ev.target.value, destroying whatever the handler did
// (e.g. an int counter incremented by the handler).
func TestJsgenExplicitHandlerOwnsState(t *testing.T) {
	src := "state count = 0\n" +
		"\n" +
		"fn onSwitch() -> null {\n" +
		"\tcount = count + 1\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tswitch checked: count onChange: onSwitch\n" +
		"}\n"
	prog, dl := ParseSource(src, "swh.sui")
	if dl.Count() != 0 {
		t.Errorf("switch+handler app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "e0.onchange = onSwitch;", t)
	if strings.Contains(js, "setCount(ev.target.checked)") {
		t.Errorf("explicit onChange must not be followed by the state setter:\n%s", js)
	}
	if strings.Contains(js, "__u0(ev)") {
		t.Errorf("no handler chaining expected when the user owns the event:\n%s", js)
	}
	jsBalanced(js, t)

	// Same for value + onInput on a text input.
	src2 := "state msg = \"\"\n" +
		"\n" +
		"fn onType() -> null {\n" +
		"\tmsg = msg\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tinput \"x\" value: msg onInput: onType\n" +
		"}\n"
	prog2, dl2 := ParseSource(src2, "inh.sui")
	if dl2.Count() != 0 {
		t.Errorf("input+handler app: parse diagnostics: %v", dl2.Count())
		return
	}
	js2 := GenerateJs(prog2)
	jsContains(js2, "e0.oninput = onType;", t)
	if strings.Contains(js2, "setMsg(ev.target.value)") {
		t.Errorf("explicit onInput must not be followed by the state setter:\n%s", js2)
	}
	jsBalanced(js2, t)
}

func TestJsgenLiterals(t *testing.T) {
	src := "state f: float = 1.5\n" +
		"state m: map[string]int = {\"a\": 1}\n" +
		"state b: bool = true\n" +
		"state n: null = null\n" +
		"state xs: array[int] = [1, 2, 3]\n"
	prog, dl := ParseSource(src, "lit.sui")
	if dl.Count() != 0 {
		t.Errorf("literals app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "let f = 1.5;", t)
	jsContains(js, "let m = ({\"a\": 1});", t)
	jsContains(js, "let b = true;", t)
	jsContains(js, "let n = null;", t)
	jsContains(js, "let xs = [1, 2, 3];", t)
	jsBalanced(js, t)
}

func TestJsgenEscapes(t *testing.T) {
	if EscapeJsStr("a\"b\\c") != "a\\\"b\\\\c" {
		t.Errorf("EscapeJsStr: got %q", EscapeJsStr("a\"b\\c"))
	}
	if EscapeJsStr("x\ny\tz\r") != "x\\ny\\tz\\r" {
		t.Errorf("EscapeJsStr controls: got %q", EscapeJsStr("x\ny\tz\r"))
	}
	if EscapeJsTemplate("a`b${c}\\d") != "a\\`b\\${c}\\\\d" {
		t.Errorf("EscapeJsTemplate: got %q", EscapeJsTemplate("a`b${c}\\d"))
	}
	if EscapeJsTemplate("nl\n") != "nl\\n" {
		t.Errorf("EscapeJsTemplate newline: got %q", EscapeJsTemplate("nl\n"))
	}
}

func TestJsgenStatementsToJs(t *testing.T) {
	var e Expr
	e.kind = "int"
	e.i = 5
	var l Stmt
	l.kind = "let"
	l.name = "x"
	l.init = &e
	var body []*Stmt
	body = append(body, &l)
	var a Stmt
	a.kind = "assign"
	a.targets = append(a.targets, "x")
	var v Expr
	v.kind = "int"
	v.i = 6
	a.rhs = &v
	body = append(body, &a)
	s := StatementsToJs(body)
	if !strings.Contains(s, "let x = 5;") {
		t.Errorf("StatementsToJs let: got %q", s)
	}
	if !strings.Contains(s, "x = 6;") {
		t.Errorf("StatementsToJs assign: got %q", s)
	}
}

func TestJsgenTwoTupleAssignsInOneFunction(t *testing.T) {
	src := "import { http } from \"std\"\n" +
		"\n" +
		"async fn go() -> null {\n" +
		"\ta, e = http.get(\"https://x\")\n" +
		"\tb, f = http.post(\"https://y\", a)\n" +
		"}\n"
	prog, dl := ParseSource(src, "two.sui")
	if dl.Count() != 0 {
		t.Errorf("two-tuple app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "const __sui_r = await __SUI_STD.http_get(", t)
	jsContains(js, "const a = __sui_r[0];", t)
	jsContains(js, "const e = __sui_r[1];", t)
	jsContains(js, "const __sui_r2 = await __SUI_STD.http_post(", t)
	jsContains(js, "const b = __sui_r2[0];", t)
	jsContains(js, "const f = __sui_r2[1];", t)
	jsContains(js, "http_post: async (url, data) => {", t)
	jsBalanced(js, t)
}

func TestJsgenRangeLoop(t *testing.T) {
	src := "fn sum() -> int {\n" +
		"\tlet acc = 0\n" +
		"\tloop i in 0..10 {\n" +
		"\t\tacc = acc + i\n" +
		"\t}\n" +
		"\treturn acc\n" +
		"}\n"
	prog, dl := ParseSource(src, "sum.sui")
	if dl.Count() != 0 {
		t.Errorf("range loop app: parse diagnostics: %v", dl.Count())
		return
	}
	js := GenerateJs(prog)
	jsContains(js, "for (let i = 0; i < 10; i++) {", t)
	jsContains(js, "let acc = 0;", t)
	jsContains(js, "acc = (acc + i);", t)
	jsBalanced(js, t)
}

func TestBindingOnlyValueDoesNotRebindView(t *testing.T) {
	src := "state q = \"x\"\nview main {\n\tinput \"Q\" value: q\n}\n"
	prog, dl := ParseSource(src, "bind.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.Count())
	}
	js := GenerateJs(prog)
	if strings.Contains(js, `__sui_bindView("q"`) {
		t.Errorf("write-only value binding must not bind the view:\n%v", js)
	}
}

func TestReadStateStillRebindsView(t *testing.T) {
	src := "state q = \"x\"\nview main {\n\tinput \"Q\" value: q\n\tp \"Value: {q}\"\n}\n"
	prog, dl := ParseSource(src, "bind.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.Count())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `__sui_bindView("q"`) {
		t.Errorf("state read by the view must bind it:\n%v", js)
	}
}
