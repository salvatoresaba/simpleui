package sui

import (
	"strings"
	"testing"
)

func TestDocsAgentAndSearch(t *testing.T) {
	m, err := loadDocsManifest()
	if err != nil {
		t.Fatal(err)
	}
	agent := docsAgentText(m)
	for _, want := range []string{"Non-negotiable rules", "simpleui check app.sui --json", "docs/AGENT_RECIPES.md", "srcAllow"} {
		if !strings.Contains(agent, want) {
			t.Errorf("agent context missing %q", want)
		}
	}
	if !strings.Contains(docsSearchText(m, "postSafe"), "STDLIB.md") {
		t.Error("full-text search did not find postSafe in STDLIB.md")
	}
	if !strings.Contains(docsSearchText(m, "iframe"), "iframe-embed") {
		t.Error("search did not find the iframe recipe")
	}
	full, err := docsFullText(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(full, "# SimpleUI API Reference") || !strings.Contains(full, "SimpleUI coding-agent guide") {
		t.Error("full documentation bundle is incomplete")
	}
}
