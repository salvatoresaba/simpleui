package sui

import (
	"strings"
	"testing"
)

// TestCheckToggleNoCheckedWarning verifies the warning emitted when a value
// control has an onChange handler but no checked: binding. The full re-render
// that follows a state change would otherwise reset the control, which is a
// confusing trap: the click changes state, yet the switch never moves.
func TestCheckToggleNoCheckedWarning(t *testing.T) {
	src := "state active = false\n" +
		"\n" +
		"view main {\n" +
		"\tswitch \"\" onChange: toggleActive\n" +
		"}\n" +
		"\n" +
		"fn toggleActive() -> null {\n" +
		"\tactive = active == true ? false : true\n" +
		"}\n"
	prog, dl := ParseSource(src, "sw.sui")
	if dl.HasErrors() {
		t.Fatalf("parse errors: %v", dl.String())
	}
	cd := Check(prog)
	if cd.HasErrors() {
		t.Fatalf("the missing checked: binding must be a warning, not an error: %v", cd.String())
	}
	if !hasMsg(cd, "handler but no checked: binding, so it will not reflect state") {
		t.Errorf("expected warning about missing checked: binding, got: %v", cd.String())
	}
	if cd.WarningCount() != 1 {
		t.Errorf("expected exactly 1 warning, got %v (%v)", cd.WarningCount(), cd.String())
	}
	if !strings.Contains(cd.String(), "1 warning.") {
		t.Errorf("report should say \"1 warning.\", got: %v", cd.String())
	}

	// Binding checked: must silence the warning.
	ok := "state active = false\n" +
		"\n" +
		"view main {\n" +
		"\tswitch \"\" checked: active onChange: toggleActive\n" +
		"}\n" +
		"\n" +
		"fn toggleActive() -> null {\n" +
		"\tactive = active == true ? false : true\n" +
		"}\n"
	prog2, dl2 := ParseSource(ok, "sw_ok.sui")
	if dl2.HasErrors() {
		t.Fatalf("parse errors: %v", dl2.String())
	}
	if cd2 := Check(prog2); cd2.Count() != 0 {
		t.Errorf("checked: binding should produce no diagnostics, got: %v", cd2.String())
	}

	// A control with neither a state-binding nor a handler is inert: nothing
	// reads its value and re-renders reset it. This must be a warning (per the
	// "dead control" contract) for every control family.
	loose := "state on = false\n" +
		"\n" +
		"view main {\n" +
		"\tswitch \"\"\n" +
		"\tswitch \"\" checked: true\n" +
		"\tinput \"x\"\n" +
		"\tselect \"\" options: \"A,B\"\n" +
		"\tbutton \"Save\"\n" +
		"\tbutton \"x\" onClick: bump\n" +
		"}\n" +
		"\n" +
		"fn bump() -> null {\n" +
		"\ton = on == true ? false : true\n" +
		"}\n"
	prog3, dl3 := ParseSource(loose, "sw3.sui")
	if dl3.HasErrors() {
		t.Fatalf("parse errors: %v", dl3.String())
	}
	cd3 := Check(prog3)
	if cd3.HasErrors() {
		t.Errorf("unbound controls must warn, not error: %v", cd3.String())
	}
	// Dead-control warnings: loose switch, unbound input, unbound select.
	// (the "checked: true" switch is bound-but-static: it warns as non-reflecting
	// only if it had a handler; here it has none, so it counts as a dead control.)
	if counted := strings.Count(tcDiagMsgs(cd3), "nothing reads its value and re-renders reset it"); counted != 4 {
		t.Errorf("expected 4 dead-control warnings, got %v:\n%v", counted, cd3.String())
	}
	// A handlerless button is inert too: clicking does nothing.
	if !hasMsg(cd3, "button has no handler, so nothing reacts to it") {
		t.Errorf("expected handlerless-button warning, got: %v", cd3.String())
	}
	if !strings.Contains(tcDiagHints(cd3), "add onClick: <fn> to react to the click") {
		t.Errorf("expected onClick hint for the handlerless button, got: %v", cd3.String())
	}
}
