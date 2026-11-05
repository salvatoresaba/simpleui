package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadProgramKeepsCss guards the CLI/build path: LoadProgram must merge
// `css "..."` declarations from the entry file (and imports), not only from
// ParseSource (which the designer uses).
func TestLoadProgramKeepsCss(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.sui")
	src := "css \".active \\{ color: red; \\}\"\n" +
		"view main {\n" +
		"\tdiv \"x\" class: \"active\"\n" +
		"}\n"
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, dl := LoadProgram(file)
	if dl.HasErrors() {
		t.Fatalf("load: %v", dl.String())
	}
	if len(prog.css) != 1 || !strings.Contains(prog.css[0], ".active") {
		t.Fatalf("prog.css = %#v", prog.css)
	}
	if css := GenerateCss(prog); !strings.Contains(css, ".active { color: red; }") {
		t.Errorf("generated CSS is missing the user rule:\n%s", css)
	}
}
