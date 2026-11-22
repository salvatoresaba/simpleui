package core

import "strings"
import "testing"

// inststate_test.go - per-instance state & view-local methods (v0.12):
// each component instance owns its own `state`, mutated through a method
// dispatched via __sui_vm against the per-site store.

func TestPerInstanceState(t *testing.T) {
	src := "view Counter(start: int) {\n" +
		"\tstate count = start\n" +
		"\tfn inc() -> null {\n" +
		"\t\tcount += 1\n" +
		"\t}\n" +
		"\tbutton \"{count}\" onClick: inc\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tCounter start: 0\n" +
		"\tCounter start: 10\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("per-instance program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "is.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"function __html_Counter({ __inst: inst, __key, start }) {",
		"if (!(\"count\" in inst)) inst.count = start;",
		"let count = inst.count;",
		"function inc() {",
		"inst.inc = inc;",
		"onclick=\"__sui_vm(${JSON.stringify(__key)}, 'inc')\"",
		"__SUI_inst[\"0\"] ||= {})",
		"__SUI_inst[\"1\"] ||= {})",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("per-instance jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestPerInstanceStateErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		// A method is only callable inside its component view.
		{"state n = 0\nview C() {\n\tstate count = 0\n\tfn inc() -> null {\n\t\tcount += 1\n\t}\n}\nfn top() -> null {\n\tinc()\n}\nview main {\n\tC\n}", "can only be called inside its component view"},
		{"view C() {\n\tstate x = 0\n\tstate x = 1\n}\nview main {\n\tC\n}", "duplicate instance state"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
