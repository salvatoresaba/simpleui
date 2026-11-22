package apitools

import (
	"path/filepath"
	"testing"
)

// repoRoot is the SimpleUI repository root, two levels above this package.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

// TestPublicApiMatchesGolden compares the exported API of the root package
// with the snapshot taken before the implementation moved into internal/core.
//
// The snapshot is the contract of the public module
// github.com/salvatoresaba/simpleui: every function, type, constant and
// variable it lists must still be reachable from the root package, with the
// same signature. Methods are verified at compile time by the generated
// assertions in api_methods_test.go.
func TestPublicApiMatchesGolden(t *testing.T) {
	golden, err := ReadGolden(filepath.Join("testdata", "api_surface.txt"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	live, err := ExtractDir(repoRoot(t))
	if err != nil {
		t.Fatalf("extract live surface: %v", err)
	}
	diffs := Compare(live, golden)
	for _, diff := range diffs {
		t.Errorf("public API drift: %s", diff)
	}
}

// TestGoldenIsNotEmpty guards against a snapshot that would silently accept
// any implementation.
func TestGoldenIsNotEmpty(t *testing.T) {
	golden, err := ReadGolden(filepath.Join("testdata", "api_surface.txt"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if len(golden["functions"]) < 50 {
		t.Fatalf("golden snapshot looks truncated: %d functions", len(golden["functions"]))
	}
	if len(golden["types"]) < 20 {
		t.Fatalf("golden snapshot looks truncated: %d types", len(golden["types"]))
	}
	if len(golden["methods"]) < 10 {
		t.Fatalf("golden snapshot looks truncated: %d methods", len(golden["methods"]))
	}
}
