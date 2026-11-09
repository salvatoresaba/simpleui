package main

import (
	"strings"
	"testing"
)

func TestLegacyChartAttributesRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"state chart", `view main { chart "C" chartType: "state" }`, "chartType"},
		{"statemachine chart", `view main { chart "C" chartType: "statemachine" }`, "chartType"},
		{"interactive", `view main { chart "C" interactive: true }`, "interactive"},
		{"stateMode", `view main { chart "C" stateMode: "static" }`, "stateMode"},
		{"tableEdit", `view main { graph "G" tableEdit: true }`, "tableEdit"},
		{"tree icon", `view main { tree "T" icon: "x" }`, "icon"},
		{"tree editable", `view main { tree "T" editable: true }`, "editable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, pdl := ParseSource(tc.src, "legacy.sui")
			if pdl.HasErrors() {
				t.Fatalf("parse: %s", pdl.String())
			}
			dl := Check(prog)
			if !dl.HasErrors() {
				t.Fatalf("legacy source must be rejected, got no diagnostics")
			}
			if !strings.Contains(dl.String(), tc.want) {
				t.Fatalf("diagnostic should mention %q: %s", tc.want, dl.String())
			}
		})
	}
}

func TestLegacyRuntimeAliasesRemoved(t *testing.T) {
	for _, legacy := range []string{
		`["start", "from"`,
		`["end", "to"`,
		`["task", "label", "name"]`,
		`["name", "label"]`,
		`["initial", "isInitial"`,
		`["type", "kind"`,
		`attrs.tooltip`,
		`attrs.note`,
		`attrs.colour`,
		`ca.source`,
		`ca.target`,
		`xa.trigger`,
		`ra.condition`,
		`ca.effect`,
	} {
		if strings.Contains(suiChartRuntime, legacy) {
			t.Errorf("chart runtime still accepts legacy alias %q", legacy)
		}
	}
	if !strings.Contains(suiChartRuntime, `if (typeof v === "number") return null;`) {
		t.Error("ganttDate must reject numeric positions")
	}
}
