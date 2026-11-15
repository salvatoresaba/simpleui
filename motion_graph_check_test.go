package main

import (
	"strings"
	"testing"
)

func TestMotionGraphGuardValidation(t *testing.T) {
	valid := []string{
		"gate",
		"!gate",
		"gate && editing",
		"(gate && editing) || mode == \"edit\"",
		"level >= 2 && level < 10",
		"name != \"\" && name === 'x'",
		"true",
		"!false && (count == 0 || count != 3)",
	}
	for _, src := range valid {
		if msg := motionGraphGuardError(src); msg != "" {
			t.Errorf("guard %q should be valid, got %q", src, msg)
		}
	}
	invalid := map[string]string{
		"gate &&":       "ends with an operator",
		"(gate":         "unbalanced parentheses",
		"gate)":         "unbalanced parentheses",
		"gate + 1":      "invalid character",
		"!":             "ends with an operator",
		"'":             "unterminated string",
		"gate gate":     "expected an operator",
		"&& gate":       "expected a value",
		"gate == edge)": "unbalanced parentheses",
	}
	for src, want := range invalid {
		msg := motionGraphGuardError(src)
		if msg == "" {
			t.Errorf("guard %q should be invalid", src)
			continue
		}
		if !strings.Contains(msg, want) {
			t.Errorf("guard %q: got %q, want substring %q", src, msg, want)
		}
	}
}

func TestMotionGraphCheckDiagnostics(t *testing.T) {
	if msgs := motionGraphDiagnostics(`{"initial":"a","states":[{"id":"a"},{"id":"b"}],"links":[{"from":"a","to":"b","event":"click","guard":"gate && edit"}]}`); len(msgs) != 0 {
		t.Fatalf("valid graph produced diagnostics: %v", msgs)
	}
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"bad json", `{"states":[`, "not valid JSON"},
		{"no states", `{"initial":"a","states":[]}`, "at least one state"},
		{"duplicate state", `{"states":[{"id":"a"},{"id":"a"}]}`, "duplicate state id"},
		{"unknown initial", `{"initial":"b","states":[{"id":"a"}]}`, "initial state"},
		{"unknown source", `{"states":[{"id":"a"}],"links":[{"from":"x","to":"a","event":"click"}]}`, "source state"},
		{"unknown target", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"x","event":"click"}]}`, "target state"},
		{"missing trigger", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a"}]}`, "missing trigger"},
		{"unknown trigger", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"hover"}]}`, "unknown trigger"},
		{"bad guard", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"click","guard":"a &&"}]}`, "guard: guard ends with an operator"},
		{"bad exit", `{"states":[{"id":"a"}],"exit":{"state":"gone"}}`, "exit: state"},
	}
	for _, tc := range cases {
		msgs := motionGraphDiagnostics(tc.raw)
		if len(msgs) == 0 {
			t.Errorf("%s: expected diagnostics", tc.name)
			continue
		}
		joined := strings.Join(msgs, "\n")
		if !strings.Contains(joined, tc.want) {
			t.Errorf("%s: got %q, want substring %q", tc.name, joined, tc.want)
		}
	}
}

func TestMotionGraphEventDiagnostics(t *testing.T) {
	valid := `{"initial":"a","states":[{"id":"a"},{"id":"b"}],"links":[
		{"id":"c","from":"a","to":"b","event":"custom","customEvent":"scene-open"},
		{"id":"s","from":"a","to":"b","event":"state","when":"!flag"},
		{"id":"e","from":"a","to":"b","event":"elementState","sourceId":"src","whenState":"b"},
		{"id":"t","from":"a","to":"b","event":"timer","timer":"1.5s"},
		{"id":"w","from":"a","to":"b","event":"wheel","wheel":"down"},
		{"id":"sw","from":"a","to":"b","event":"swipe","axis":"left","distance":45},
		{"id":"l","from":"a","to":"b","event":"longpress","hold":600},
		{"id":"v","from":"a","to":"b","event":"visible","threshold":0.6}
	],"exit":{"state":"b","duration":"180ms","delay":"20ms","spring":{"stiffness":180,"damping":14,"mass":1}}}`
	if msgs := motionGraphDiagnostics(valid); len(msgs) != 0 {
		t.Fatalf("valid event graph produced diagnostics: %v", msgs)
	}
	invalid := []struct {
		name string
		raw  string
		want string
	}{
		{"custom missing", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"custom"}]}`, "customEvent"},
		{"custom bad", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"custom","customEvent":"1bad"}]}`, "invalid custom event"},
		{"state missing", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"state"}]}`, "needs when"},
		{"state bad", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"state","when":"a b"}]}`, "invalid state name"},
		{"element missing source", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"elementState","whenState":"b"}]}`, "sourceId"},
		{"element missing state", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"elementState","sourceId":"x"}]}`, "whenState"},
		{"timer missing", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"timer"}]}`, "non-negative time"},
		{"timer bad", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"timer","timer":"soon"}]}`, "non-negative time"},
		{"wheel bad", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"wheel","wheel":"left"}]}`, "wheel must be"},
		{"swipe axis", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"swipe","axis":"diagonal"}]}`, "swipe axis"},
		{"swipe distance", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"swipe","distance":0}]}`, "positive number"},
		{"hold", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"longpress","hold":-1}]}`, "non-negative number"},
		{"threshold", `{"states":[{"id":"a"}],"links":[{"from":"a","to":"a","event":"visible","threshold":1.5}]}`, "between 0 and 1"},
		{"duplicate link", `{"states":[{"id":"a"}],"links":[{"id":"x","from":"a","to":"a","event":"click"},{"id":"x","from":"a","to":"a","event":"click"}]}`, "duplicate link id"},
		{"exit duration", `{"states":[{"id":"a"}],"exit":{"state":"a","duration":"later"}}`, "duration must be"},
		{"exit spring", `{"states":[{"id":"a"}],"exit":{"state":"a","spring":{"stiffness":5000}}}`, "stiffness must be"},
	}
	for _, tc := range invalid {
		msgs := motionGraphDiagnostics(tc.raw)
		if len(msgs) == 0 {
			t.Errorf("%s: expected diagnostics", tc.name)
			continue
		}
		joined := strings.Join(msgs, "\n")
		if !strings.Contains(joined, tc.want) {
			t.Errorf("%s: got %q, want substring %q", tc.name, joined, tc.want)
		}
	}
}

func TestMotionGraphCheckIntegration(t *testing.T) {
	graph := map[string]any{
		"initial": "a",
		"states":  []any{map[string]any{"id": "a", "css": "width: 100px;"}},
		"links": []any{map[string]any{
			"from": "a", "to": "a", "event": "click", "guard": "gate &&",
		}},
	}
	src := "view main {\n\tcard motionGraph: \"" + suiGraphLiteral(graph) + "\" { p \"G\" }\n}\n"
	prog, dl := ParseSource(src, "motion_check.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatal("invalid guard should fail Check")
	}
	if !strings.Contains(cd.String(), "guard: guard ends with an operator") {
		t.Fatalf("missing guard diagnostic:\n%s", cd.String())
	}
}
