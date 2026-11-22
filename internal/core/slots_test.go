package core

import "strings"
import "testing"

// slots_test.go - component children/slots (v0.10): a component renders the
// caller's block via the `children` marker.

func TestComponentSlots(t *testing.T) {
	src := "state clicks = 0\n" +
		"\n" +
		"fn go() -> null {\n" +
		"\tclicks += 1\n" +
		"}\n" +
		"\n" +
		"view Panel(heading: string) {\n" +
		"\th2 \"{heading}\"\n" +
		"\tchildren\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tPanel heading: \"Box\" {\n" +
		"\t\tp \"inside {clicks}\"\n" +
		"\t\tbutton \"Go\" onClick: go\n" +
		"\t}\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("slots program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "sl.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"function __html_Panel({ heading, children }) {",
		"children: `<p",
		">inside ${__sui_esc(clicks)}</p>`",
		"data-sui-id=\"0\"",
		"${children}",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("slots jsgen missing %q\n----\n%v", want, js)
		}
	}
	// The caller's states (clicks) are bound on the root.
	if !strings.Contains(js, `__sui_bindView("clicks", "main");`) {
		t.Errorf("children state must be bound on the root:\n%v", js)
	}
	// The caller's button belongs to the root wiring (query-based, with id).
	if !strings.Contains(js, "document.querySelector('[data-sui-id=\"1\"]')") {
		t.Errorf("caller button should use root wiring:\n%v", js)
	}
}

func TestComponentSlotErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"view main {\n\tchildren\n}\n", "children is only valid inside a component view"},
		{"view A(x: int) {\n\tbutton \"b\" onClick: go\n\tchildren\n\tchildren\n}\nview main {\n\tA x: 1\n}\nfn go() -> null {\n}\n", "children can only appear once"},
		{"view A(x: int) {\n\tchildren \"text\"\n}\nview main {\n\tA x: 1\n}\n", "children cannot take text"},
		{"view A(x: int) {\n\tchildren onClick: go\n}\nview main {\n\tA x: 1\n}\nfn go() -> null {\n}\n", "children cannot take attribute"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}

func TestComponentNoSlotSignatureUnchanged(t *testing.T) {
	// Components that do not use children keep the plain signature.
	src := "view Badge(tone: string) {\n\tp \"{tone}\"\n}\nview main {\n\tBadge tone: \"x\"\n}\n"
	prog, pdl := ParseSource(src, "nb.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "function __html_Badge({ tone }) {") {
		t.Errorf("signature must not add children when unused:\n%v", js)
	}
}
