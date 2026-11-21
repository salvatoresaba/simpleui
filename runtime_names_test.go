package main

import (
	"strings"
	"testing"
)

func checkRuntimeNamesSource(t *testing.T, src string) DiagList {
	t.Helper()
	prog, pdl := ParseSource(src, "runtime_names.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	return Check(prog)
}

func TestRuntimeNameDuplicatesAreRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"richtext", `state doc = ""
view main {
	richtext name: "editor" value: doc
	richtext name: "editor" value: doc
}`},
		{"table", `view main {
	table name: "users" { }
	table name: "users" { }
}`},
		{"tree", `view main {
	tree name: "files" { }
	tree name: "files" { }
}`},
		{"chart and graph", `state flow = []
view main {
	chart "A" name: "sales" data: flow
	graph "B" name: "sales" data: flow
}`},
		{"modal", `view main {
	modal "One" name: "dialog" { }
	modal "Two" name: "dialog" { }
}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dl := checkRuntimeNamesSource(t, tc.src)
			if !dl.HasErrors() {
				t.Fatalf("expected a duplicate name diagnostic")
			}
			if !strings.Contains(dl.String(), `duplicate runtime name "`) {
				t.Fatalf("unexpected diagnostics: %v", dl.String())
			}
		})
	}
}

func TestRuntimeNamesMayRepeatAcrossFamilies(t *testing.T) {
	dl := checkRuntimeNamesSource(t, `state doc = ""
state flow = []
view main {
	table name: "data" { }
	modal "Details" name: "data" { }
	richtext name: "data" value: doc
	chart "Sales" name: "data" data: flow
}`)
	if dl.HasErrors() {
		t.Fatalf("names from different runtime families must not collide: %v", dl.String())
	}
}

func TestDesignerExposesRuntimeNamesAndLiveChecks(t *testing.T) {
	if !strings.Contains(guiHTML, "function chartNameField(n)") {
		t.Error("designer has no chart/graph runtime name field")
	}
	if strings.Count(guiHTML, "b.appendChild(chartNameField(n));") < 2 {
		t.Error("chart and graph inspectors must both render the runtime name field")
	}
	if !strings.Contains(guiHTML, "markAutosaveDirty(); scheduleLiveCheck();") {
		t.Error("designer model edits do not schedule the live duplicate-name check")
	}
}

func TestRuntimeNamesUniquePass(t *testing.T) {
	dl := checkRuntimeNamesSource(t, `state doc = ""
state flow = []
view main {
	richtext name: "editor" value: doc
	table name: "users" { }
	tree name: "files" { }
	chart "Sales" name: "sales" data: flow
	graph "Flow" name: "flow" data: flow
	modal "Details" name: "details" { }
}`)
	if dl.HasErrors() {
		t.Fatalf("unique runtime names must pass: %v", dl.String())
	}
}
