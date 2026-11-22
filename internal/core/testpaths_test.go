package core

import (
	"os"
	"path/filepath"
	"testing"
)

// pkgDir is the directory of this package (internal/core), captured before
// TestMain moves the working directory to the repository root. Tests that
// read the sources of this package use it to stay independent of the working
// directory.
var pkgDir string

// TestMain restores the working directory the tests of this package had
// before the implementation moved from the repository root to internal/core.
//
// The tests address repository files (examples/, components/, Test/testdata/)
// with paths relative to the repository root, and they must keep doing so
// after the move. It also injects the documentation file system: the embed
// lives in the root package, which is the only place go:embed can reach the
// repository files, so the tests read the same tree from disk.
func TestMain(m *testing.M) {
	dir, err := filepath.Abs(".")
	if err != nil {
		panic("resolve package directory: " + err.Error())
	}
	root := filepath.Join(dir, "..", "..")
	if err := os.Chdir(root); err != nil {
		panic("enter repository root: " + err.Error())
	}
	pkgDir = dir
	docsFS = os.DirFS(root)
	os.Exit(m.Run())
}
