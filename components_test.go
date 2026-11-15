package main

import "strings"
import "testing"

// components_test.go - parameterized views / components (v0.4): reusable
// chunk views invoked from other views, inline event handlers, reactivity
// across component boundaries, and component errors (including cycles).

func TestComponentBasic(t *testing.T) {
	src := "state count = 0\n" +
		"state label = \"items\"\n" +
		"\n" +
		"fn inc() -> null {\n" +
		"\tcount += 1\n" +
		"}\n" +
		"\n" +
		"view Badge(prefix: string) {\n" +
		"\tbutton \"{prefix} {count}\" onClick: inc\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tBadge prefix: \"seen\"\n" +
		"\tp \"{label}\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("component program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "c.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"function __html_Badge({ prefix }) {",
		"__html_Badge({prefix: `seen`})",
		"onclick=\"inc()\"",
		"${__sui_esc(prefix)} ${__sui_esc(count)}",
		"function render_main()",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("component jsgen missing %q\n----\n%v", want, js)
		}
	}
	// Badge is a component: no root renderer, and no bind line for it.
	if strings.Contains(js, "render_Badge") {
		t.Errorf("component must not get a root renderer:\n%v", js)
	}
	if strings.Contains(js, `__SUI_views["Badge"]`) {
		t.Errorf("component must not register as root:\n%v", js)
	}
	// Reactivity crosses the component boundary: count is referenced inside
	// Badge but bound on the root view; the parameter is not a state.
	for _, want := range []string{
		`__sui_bindView("count", "main");`,
		`__sui_bindView("label", "main");`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("component binding missing %q\n----\n%v", want, js)
		}
	}
	if strings.Contains(js, `__sui_bindView("prefix"`) {
		t.Errorf("component parameter must not be bound as a state:\n%v", js)
	}
}

func TestComponentCallsIfListNested(t *testing.T) {
	src := "state names = [\"a\", \"b\"]\n" +
		"\n" +
		"view Row2(one: string, two: string) {\n" +
		"\trow {\n" +
		"\t\tp \"{one}\"\n" +
		"\t\tp \"{two}\"\n" +
		"\t}\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tif len(names) > 0 {\n" +
		"\t\tlist i in names {\n" +
		"\t\t\tRow2 one: i two: names[0]\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("component nested program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "n.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "__html_Row2({one: i, two: names[0]})") {
		t.Errorf("component within list missing:\n%v", js)
	}
	if strings.Contains(js, "render_Row2") {
		t.Errorf("Row2 must be a component, not a root:\n%v", js)
	}
}

func TestComponentErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"view Badge(tone: string) {\n}\nview main {\n\tBadge tone2: \"x\"\n}\n", "unknown parameter"},
		{"view Badge(tone: string) {\n}\nview main {\n\tBadge tone: \"x\" nope: \"y\"\n}\n", "unknown parameter"},
		{"view Badge(tone: string, label: string) {\n}\nview main {\n\tBadge tone: \"x\"\n}\n", "missing parameter"},
		{"view Badge(count: int) {\n}\nview main {\n\tBadge count: \"x\"\n}\n", "expects int, got string"},
		{"view main {\n\tIntruder a: \"x\"\n}\n", "unknown tag"},
		{"view A(x: int) { B y: x }\nview B(y: int) { A x: y }\nview main { A x: 1 }\n", "circular component reference"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}

func TestComponentMultipleInstancesInlineIds(t *testing.T) {
	// Two instances of the same component must not both carry a data-sui-id
	// (they use inline handlers), so roots keep the query wiring machine.
	src := "fn go() -> null {\n}\n" +
		"\n" +
		"view Badge(tone: string) {\n" +
		"\tbutton \"{tone}\" onClick: go\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tBadge tone: \"a\"\n" +
		"\tBadge tone: \"b\"\n" +
		"}\n"
	prog, pdl := ParseSource(src, "m.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	// The runtime prelude mentions data-sui-id for focus and animation
	// restoration; the check is about the generated component chunk.
	start := strings.Index(js, "function __html_Badge")
	if start < 0 {
		t.Fatalf("component chunk missing:\n%v", js)
	}
	chunk := js[start:]
	if next := strings.Index(chunk[1:], "function "); next >= 0 {
		chunk = chunk[:next+1]
	}
	if strings.Contains(chunk, `data-sui-id="`) {
		t.Errorf("component buttons must use inline handlers, not ids:\n%v", chunk)
	}
	if strings.Contains(js, `onclick="go()"`) {
		// ok
	} else {
		t.Errorf("inline handler missing:\n%v", js)
	}
}
func TestComponentFnParams(t *testing.T) {
	src := "state clicks = 0\n" +
		"\n" +
		"fn turnOn() -> null {\n" +
		"\tclicks += 1\n" +
		"}\n" +
		"\n" +
		"view Switch(label: string, action: fn) {\n" +
		"\tbutton \"{label}\" onClick: action\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tSwitch label: \"Go\" action: turnOn\n" +
		"\tp \"{clicks}\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("fn-param component:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "fp.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"action: \"turnOn\"",                 // handler passed by name (string)
		"onclick=\"${__sui_esc(action)}()\"", // inline handler interpolating the param
	} {
		if !strings.Contains(js, want) {
			t.Errorf("fn-param jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestComponentFnParamErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"fn f() -> null {\n}\nview B(a: fn) {\n\tbutton \"x\" onClick: a\n}\nview main {\n\tB a: 42\n}\n", "expects a function name"},
		{"fn f() -> null {\n}\nview B(a: fn) {\n\tbutton \"x\" onClick: a\n}\nview main {\n\tB a: \"f\"\n}\n", "expects a function name"},
		{"view B(a: fn) {\n\tbutton \"x\" onClick: a\n}\nview main {\n\tB a: notAThing\n}\n", "expects a function name"},
		{"view B(a: fn) {\n\tbutton \"x\" onClick: a\n}\nview main {\n\tB\n}\n", "missing parameter"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
