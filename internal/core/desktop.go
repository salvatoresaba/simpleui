package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// desktop.go - the desktop build pipeline (DESIGN §3.2 desktop target, §4).
//
// After the front-end has produced a checked Prog, BuildDesktop materializes a
// tiny module (go.mod "desktopapp" + main.go from runtime.go's template) next
// to the embedded web assets and compiles it with the local Go toolchain into
// a single executable. The user's machine never needs Node/npm/Python to run
// the result; only the compiler build step requires a Go toolchain (see the
// note in SPEC §8).
//
// Two desktop renderers are available:
//   - desktop: opens the system default browser on a loopback URL;
//   - desktop-webview: opens an embedded native WebView window (WebView2 on
//     Windows, WKWebView on macOS, WebKitGTK on Linux).

const webviewGoModule = "github.com/webview/webview_go"
const webviewGoVersion = "v0.0.0-20240831120633-6173450d4dd6"

// BuildDesktop is the CLI entry point: it runs the build and exits on failure.
func BuildDesktop(prog *Prog, outDir string, allowExec []string, allowDirs []string) {
	if err := RunDesktopBuild(prog, outDir, allowExec, allowDirs); err != nil {
		fmt.Fprintf(os.Stderr, "%v: desktop build failed: %v\n", AppName, err)
		os.Exit(1)
	}
	printDesktopSummary(outDir, allowExec, allowDirs)
}

// BuildDesktopWebview is the CLI entry point for the embedded-WebView target.
func BuildDesktopWebview(prog *Prog, outDir string, allowExec []string, allowDirs []string) {
	if err := RunDesktopWebviewBuild(prog, outDir, allowExec, allowDirs); err != nil {
		fmt.Fprintf(os.Stderr, "%v: desktop-webview build failed: %v\n", AppName, err)
		os.Exit(1)
	}
	printDesktopWebviewSummary(outDir, allowExec, allowDirs)
}

// RunDesktopBuild writes the runtime module under outDir and compiles it into
// app(.exe). On success it returns nil; on failure it returns an error whose
// text includes the go build diagnostics. Test-friendly (never os.Exit).
func RunDesktopBuild(prog *Prog, outDir string, allowExec []string, allowDirs []string) error {
	if err := writeDesktopAssets(prog, outDir); err != nil {
		return err
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
	return runDesktopGoBuild(outDir, false)
}

// RunDesktopWebviewBuild writes and compiles the embedded-WebView variant.
// The resulting executable no longer launches a separate browser.
func RunDesktopWebviewBuild(prog *Prog, outDir string, allowExec []string, allowDirs []string) error {
	if err := writeDesktopAssets(prog, outDir); err != nil {
		return err
	}
	goMod := "module desktopapp\n\ngo 1.24.2\n\nrequire " + webviewGoModule + " " + webviewGoVersion + "\n"
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return err
	}
	title := PageTitle(prog)
	if strings.TrimSpace(title) == "" {
		title = "SimpleUI App"
	}
	if err := os.WriteFile(filepath.Join(outDir, "main.go"), []byte(DesktopWebviewMainSource(allowExec, allowDirs, title)), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "dialog.go"), []byte(desktopDialogSource), 0o644); err != nil {
		return err
	}
	if err := runGoCommand(outDir, "mod", "tidy"); err != nil {
		return fmt.Errorf("%v\nhint: the WebView target needs Go modules access (first build downloads %v); on Linux install gtk+-3.0 and webkit2gtk-4.0 development packages", err, webviewGoModule)
	}
	return runDesktopGoBuild(outDir, true)
}

// writeDesktopAssets emits the generated application assets shared by both
// desktop renderers.
func writeDesktopAssets(prog *Prog, outDir string) error {
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
	return nil
}

func desktopOutName() string {
	if runtime.GOOS == "windows" {
		return "app.exe"
	}
	return "app"
}

// runDesktopGoBuild compiles the generated module. WebView builds need CGO;
// on Windows the executable is linked as a GUI app instead of a console app.
func runDesktopGoBuild(outDir string, webview bool) error {
	outName := desktopOutName()
	args := []string{"build", "-o", outName}
	if webview && runtime.GOOS == "windows" {
		args = append(args, "-ldflags=-H=windowsgui")
	}
	args = append(args, ".")
	cmd := exec.Command("go", args...)
	cmd.Dir = outDir
	if webview {
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		manual := "cd " + outDir + " && go build -o " + outName + " ."
		if webview {
			manual = "cd " + outDir + " && go mod tidy && CGO_ENABLED=1 go build -o " + outName + " ."
		}
		return fmt.Errorf("go build failed:\n%vbuild it manually with: %v", string(out), manual)
	}
	return nil
}

func runGoCommand(dir string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go %v failed:\n%v", strings.Join(args, " "), string(out))
	}
	return nil
}

// printDesktopSummary reports the produced executable and the capability
// grants to the CLI user.
func printDesktopSummary(outDir string, allowExec []string, allowDirs []string) {
	printDesktopCapabilities(filepath.Join(outDir, desktopOutName()), "opens the system browser; pass --no-open --port N for headless", allowExec, allowDirs)
}

// printDesktopWebviewSummary reports the embedded-WebView executable.
func printDesktopWebviewSummary(outDir string, allowExec []string, allowDirs []string) {
	printDesktopCapabilities(filepath.Join(outDir, desktopOutName()), "opens a native WebView window; pass --no-open --port N for headless", allowExec, allowDirs)
}

func printDesktopCapabilities(exe, run string, allowExec []string, allowDirs []string) {
	fmt.Printf("built desktop app\n")
	fmt.Printf("  executable : %v\n", exe)
	fmt.Printf("  run        : %v   (%v)\n", exe, run)
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
