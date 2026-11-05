package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// desktop.go — the desktop build pipeline (DESIGN §3.2 desktop target, §4).
//
// After the front-end has produced a checked Prog, BuildDesktop materializes a
// tiny module (go.mod "desktopapp" + main.go from runtime.go's template) next
// to the embedded web assets and compiles it with the local Go toolchain into
// a single executable. The user's machine never needs Node/npm/Python to run
// the result; only the compiler build step requires a Go toolchain (see the
// note in SPEC §8).

// BuildDesktop is the CLI entry point: it runs the build and exits on failure.
func BuildDesktop(prog *Prog, outDir string, allowExec []string, allowDirs []string) {
	if err := RunDesktopBuild(prog, outDir, allowExec, allowDirs); err != nil {
		fmt.Fprintf(os.Stderr, "%v: desktop build failed: %v\n", AppName, err)
		os.Exit(1)
	}
	printDesktopSummary(outDir, allowExec, allowDirs)
}

// RunDesktopBuild writes the runtime module under outDir and compiles it into
// app(.exe). On success it returns nil; on failure it returns an error whose
// text includes the go build diagnostics. Test-friendly (never os.Exit).
func RunDesktopBuild(prog *Prog, outDir string, allowExec []string, allowDirs []string) error {
	assetsDir := filepath.Join(outDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		return err
	}
	assets := []struct{ name, body string }{
		{"index.html", GenerateIndexHtml(prog)},
		{"style.css", GenerateCss(prog)},
		{"app.js", GenerateJsFor(prog, "desktop")},
	}
	for i := 0; i < len(assets); i++ {
		path := filepath.Join(assetsDir, assets[i].name)
		if err := os.WriteFile(path, []byte(assets[i].body), 0o644); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), []byte("module desktopapp\n\ngo 1.24.2\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "main.go"), []byte(DesktopMainSource(allowExec, allowDirs)), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "dialog.go"), []byte(desktopDialogSource), 0o644); err != nil {
		return err
	}

	outName := "app"
	if runtime.GOOS == "windows" {
		outName = "app.exe"
	}
	cmd := exec.Command("go", "build", "-o", outName, ".")
	cmd.Dir = outDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed:\n%vbuild it manually with: cd %v && go build -o %v .", string(out), outDir, outName)
	}
	return nil
}

// printDesktopSummary reports the produced executable and the capability
// grants to the CLI user.
func printDesktopSummary(outDir string, allowExec []string, allowDirs []string) {
	outName := "app"
	if runtime.GOOS == "windows" {
		outName = "app.exe"
	}
	exe := filepath.Join(outDir, outName)
	fmt.Printf("built desktop app\n")
	fmt.Printf("  executable : %v\n", exe)
	fmt.Printf("  run        : %v   (opens the system browser; pass --no-open --port N for headless)\n", exe)
	if len(allowExec) == 0 {
		fmt.Printf("  security   : no commands allowed by the bridge (--allow-exec grants none)\n")
	} else {
		for i := 0; i < len(allowExec); i++ {
			fmt.Printf("  allowed cmd: %v\n", allowExec[i])
		}
	}
	if len(allowDirs) == 0 {
		fmt.Printf("  security   : no directories allowed by the bridge (--allow-dir grants none)\n")
	} else {
		for i := 0; i < len(allowDirs); i++ {
			fmt.Printf("  allowed dir: %v\n", allowDirs[i])
		}
	}
}
