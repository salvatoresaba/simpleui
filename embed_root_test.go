package sui

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
)

// embedRootManifest mirrors the parts of docs/manifest.json the embedded
// documentation tree must cover.
type embedRootManifest struct {
	Docs []struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	} `json:"docs"`
	Entrypoints map[string]string `json:"entrypoints"`
}

// TestEmbeddedDocsCoverManifest verifies that every document referenced by
// the manifest is really inside the binary.
//
// The embed lives in this package because go:embed cannot reach the
// repository root from internal/core, so this is the place where the
// completeness of the embedded tree is guaranteed.
func TestEmbeddedDocsCoverManifest(t *testing.T) {
	data, err := fs.ReadFile(docsFS, "docs/manifest.json")
	if err != nil {
		t.Fatalf("read embedded manifest: %v", err)
	}
	var m embedRootManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse embedded manifest: %v", err)
	}
	if len(m.Docs) == 0 {
		t.Fatal("embedded manifest lists no documents")
	}
	for _, d := range m.Docs {
		if d.Path == "" {
			t.Errorf("document %q has no path", d.ID)
			continue
		}
		if _, err := fs.ReadFile(docsFS, d.Path); err != nil {
			t.Errorf("document %q (%s) is not embedded: %v", d.ID, d.Path, err)
		}
	}
	for name, p := range m.Entrypoints {
		if _, err := fs.ReadFile(docsFS, p); err != nil {
			t.Errorf("entrypoint %s (%s) is not embedded: %v", name, p, err)
		}
	}
}

// TestEmbeddedRepoFiles verifies that the repository files the compiler needs
// offline are part of the binary.
func TestEmbeddedRepoFiles(t *testing.T) {
	files := []string{
		"AGENTS.md",
		"CLAUDE.md",
		"DESIGN.md",
		"README.md",
		"SECURITY.md",
		"SPEC.md",
		"TODO.md",
		"llms.txt",
		".mcp.json",
		".github/copilot-instructions.md",
		".cursor/rules/simpleui.mdc",
		"examples/README.md",
		"examples/security/README.md",
	}
	for _, f := range files {
		data, err := fs.ReadFile(docsFS, f)
		if err != nil {
			t.Errorf("%s is not embedded: %v", f, err)
			continue
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			t.Errorf("%s is embedded but empty", f)
		}
	}
}
