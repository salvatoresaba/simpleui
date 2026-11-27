package core

import (
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// desktop_backend_test.go - the desktop target with an embedded Go backend.
//
// `simpleui build app.sui --target desktop --backend <dir>` links a Go module
// into the produced executable: the backend is mounted in front of the app
// assets and the user runs one process instead of two.

const backendFixtureSource = `package testbackend

import "net/http"

// Handler answers /api/ping and delegates every other request to the
// fallback, which serves the compiled SimpleUI app.
func Handler(fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/ping" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("pong"))
			return
		}
		fallback.ServeHTTP(w, r)
	})
}
`

// writeBackendModule creates a Go module rooted at dir and returns it.
func writeBackendModule(t *testing.T, dir string, modulePath string, source string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create backend dir: %v", err)
	}
	goMod := "module " + modulePath + "\n\ngo 1.24.2\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write backend go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "backend.go"), []byte(source), 0o644); err != nil {
		t.Fatalf("write backend source: %v", err)
	}
	return dir
}

// backendTestProg compiles a minimal app used by the backend tests.
func backendTestProg(t *testing.T) *Prog {
	t.Helper()
	src := "view main {\n\tp \"backend app\"\n}\n"
	file := filepath.Join(t.TempDir(), "app.sui")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	prog, dl := LoadProgram(file)
	if dl.HasErrors() {
		t.Fatalf("load diagnostics:\n%v", dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics:\n%v", dl.String())
	}
	return prog
}

// TestDesktopBackendBuild checks the generated module of a desktop build with
// a backend: the backend is required, replaced by its absolute directory and
// imported with an alias in the generated main.go.
func TestDesktopBackendBuild(t *testing.T) {
	// The directory has a space on purpose: the replace target must be
	// quoted, or the Go toolchain refuses the go.mod.
	backendDir := writeBackendModule(t, filepath.Join(t.TempDir(), "my backend"), "testbackend", backendFixtureSource)
	outDir := filepath.Join(t.TempDir(), "build")
	if err := RunDesktopBuildWithBackend(backendTestProg(t), outDir, nil, nil, backendDir); err != nil {
		t.Fatalf("desktop build with backend: %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(outDir, "go.mod"))
	if err != nil {
		t.Fatalf("read generated go.mod: %v", err)
	}
	goModText := string(goMod)
	if !strings.Contains(goModText, "require testbackend v0.0.0") {
		t.Errorf("generated go.mod does not require the backend:\n%v", goModText)
	}
	absBackend, err := filepath.Abs(backendDir)
	if err != nil {
		t.Fatalf("resolve backend path: %v", err)
	}
	wantReplace := "replace testbackend => \"" + filepath.ToSlash(absBackend) + "\""
	if !strings.Contains(goModText, wantReplace) {
		t.Errorf("generated go.mod does not replace the backend with its absolute path:\n%v", goModText)
	}
	target := replaceTarget(t, goModText, "testbackend")
	if !filepath.IsAbs(filepath.FromSlash(target)) {
		t.Errorf("the replace target must be an absolute path, got %q", target)
	}

	main, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	if err != nil {
		t.Fatalf("read generated main.go: %v", err)
	}
	mainText := string(main)
	if !strings.Contains(mainText, `backend "testbackend"`) {
		t.Errorf("generated main.go does not import the backend with an alias:\n%v", mainText)
	}
	if !strings.Contains(mainText, "backend.Handler(assets)") {
		t.Errorf("generated main.go does not mount the backend:\n%v", mainText)
	}

	exe := filepath.Join(outDir, "app")
	if runtime.GOOS == "windows" {
		exe = exe + ".exe"
	}
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("built executable missing: %v", err)
	}
}

// replaceTarget returns the local path a go.mod replace directive points to.
func replaceTarget(t *testing.T, goMod string, modulePath string) string {
	t.Helper()
	for _, line := range strings.Split(goMod, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "replace "+modulePath+" => ") {
			continue
		}
		target := strings.TrimSpace(strings.TrimPrefix(line, "replace "+modulePath+" => "))
		return strings.Trim(target, "\"")
	}
	t.Fatalf("no replace directive for %v in:\n%v", modulePath, goMod)
	return ""
}

// TestDesktopBackendGoMod checks the go.mod of the generated module: the
// backend is required, replaced by its absolute directory and its own
// requirements are carried over as indirect ones.
func TestDesktopBackendGoMod(t *testing.T) {
	backendDir := filepath.Join(t.TempDir(), "backend")
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatalf("create backend dir: %v", err)
	}
	goMod := "module testbackend\n\ngo 1.25.0\n\nrequire example.com/one v1.2.3\n\nrequire (\n\texample.com/two v0.4.0 // indirect\n\texample.com/three v1.0.0\n)\n\nreplace example.com/one => ../one\n"
	if err := os.WriteFile(filepath.Join(backendDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write backend go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backendDir, "backend.go"), []byte(backendFixtureSource), 0o644); err != nil {
		t.Fatalf("write backend source: %v", err)
	}
	backend, err := prepareBackend(backendDir)
	if err != nil {
		t.Fatalf("prepare backend: %v", err)
	}
	if backend == nil {
		t.Fatal("the backend was not detected")
	}
	if backend.modulePath != "testbackend" {
		t.Errorf("module path: got %q, want %q", backend.modulePath, "testbackend")
	}
	absBackend, err := filepath.Abs(backendDir)
	if err != nil {
		t.Fatalf("resolve backend path: %v", err)
	}
	generated := backend.goMod([]string{"github.com/webview/webview_go v0.0.0"})
	for _, want := range []string{
		"module desktopapp",
		"go 1.25.0",
		"require testbackend v0.0.0",
		"replace testbackend => " + filepath.ToSlash(absBackend),
		"github.com/webview/webview_go v0.0.0",
		"example.com/one v1.2.3 // indirect",
		"example.com/two v0.4.0 // indirect",
		"example.com/three v1.0.0 // indirect",
	} {
		if !strings.Contains(generated, want) {
			t.Errorf("the generated go.mod does not contain %q:\n%v", want, generated)
		}
	}
	if strings.Contains(generated, "example.com/one => ") {
		t.Errorf("the replace directives of the backend must not be copied:\n%v", generated)
	}
}

// TestDesktopBackendWebviewSource checks that the WebView target keeps its
// browser-launch patch when a backend is embedded.
func TestDesktopBackendWebviewSource(t *testing.T) {
	src := DesktopWebviewMainSourceWithBackend(nil, nil, "Demo", "testbackend")
	if _, err := parser.ParseFile(token.NewFileSet(), "main.go", src, 0); err != nil {
		t.Fatalf("the generated WebView source is not valid Go: %v", err)
	}
	if !strings.Contains(src, `backend "testbackend"`) {
		t.Errorf("the WebView source does not import the backend:\n%v", src)
	}
	if !strings.Contains(src, "backend.Handler(assets)") {
		t.Errorf("the WebView source does not mount the backend:\n%v", src)
	}
	if !strings.Contains(src, `webview "github.com/webview/webview_go"`) {
		t.Errorf("the WebView import is missing:\n%v", src)
	}
	if !strings.Contains(src, "w := webview.New(false)") {
		t.Errorf("the WebView launch block was not applied:\n%v", src)
	}
	if strings.Contains(src, "\tif !noOpen {\n\t\topenBrowser(url)\n\t}") {
		t.Errorf("the browser launch block was not replaced:\n%v", src)
	}
}

// TestDesktopBackendErrors checks the error cases of the backend contract.
func TestDesktopBackendErrors(t *testing.T) {
	prog := backendTestProg(t)

	// A directory without a go.mod is refused with a clear message.
	missing := filepath.Join(t.TempDir(), "missing")
	if err := RunDesktopBuildWithBackend(prog, filepath.Join(t.TempDir(), "build"), nil, nil, missing); err == nil {
		t.Fatal("a directory without go.mod must be refused")
	} else if !strings.Contains(err.Error(), "no go.mod in "+missing) {
		t.Errorf("unexpected error for a missing go.mod: %v", err)
	}

	// A module whose root package does not export Handler is refused.
	noHandler := filepath.Join(t.TempDir(), "nohandler")
	writeBackendModule(t, noHandler, "nohandler", "package nohandler\n\nimport \"net/http\"\n\nfunc Handler() http.Handler {\n\treturn nil\n}\n")
	if err := RunDesktopBuildWithBackend(prog, filepath.Join(t.TempDir(), "build"), nil, nil, noHandler); err == nil {
		t.Fatal("a backend without Handler must be refused")
	} else if !strings.Contains(err.Error(), "must export func Handler(fallback http.Handler) http.Handler") {
		t.Errorf("unexpected error for a backend without Handler: %v", err)
	}

	// A package main in the root of the backend is refused explicitly.
	mainPkg := filepath.Join(t.TempDir(), "mainpkg")
	writeBackendModule(t, mainPkg, "mainpkg", "package main\n\nimport \"net/http\"\n\nfunc Handler(fallback http.Handler) http.Handler {\n\treturn fallback\n}\n")
	if err := RunDesktopBuildWithBackend(prog, filepath.Join(t.TempDir(), "build"), nil, nil, mainPkg); err == nil {
		t.Fatal("a backend with a package main must be refused")
	} else if !strings.Contains(err.Error(), "package main") {
		t.Errorf("unexpected error for a package main backend: %v", err)
	}
}

// TestDesktopBackendE2E builds the desktop app with a backend, runs it and
// checks that the backend API and the app are served by the same process,
// under the same security wrapper.
func TestDesktopBackendE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping the desktop backend end-to-end test in short mode")
	}
	backendDir := writeBackendModule(t, filepath.Join(t.TempDir(), "backend"), "testbackend", backendFixtureSource)
	buildDir := filepath.Join(t.TempDir(), "build")
	if err := RunDesktopBuildWithBackend(backendTestProg(t), buildDir, nil, nil, backendDir); err != nil {
		t.Fatalf("desktop build with backend: %v", err)
	}
	exe := filepath.Join(buildDir, "app")
	if runtime.GOOS == "windows" {
		exe = exe + ".exe"
	}

	cmd := exec.Command(exe, "--no-open")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start app: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	base := readAppURL(t, stdout, &stderr)
	client := &http.Client{Timeout: 8 * time.Second}

	// The backend API is served by the app process.
	resp, err := client.Get(base + "/api/ping")
	if err != nil {
		t.Fatalf("get /api/ping: %v (stderr: %v)", err, stderr.String())
	}
	if body := readAllClose(t, resp); body != "pong" {
		t.Errorf("backend answer: got %q, want %q", body, "pong")
	}

	// The app is still served: the backend delegates the unknown paths.
	resp, err = client.Get(base + "/")
	if err != nil {
		t.Fatalf("get /: %v (stderr: %v)", err, stderr.String())
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, "<html") {
		t.Errorf("app HTML missing: %v", body)
	}

	// The security wrapper still protects the backend: a foreign Host header
	// is refused.
	req, err := http.NewRequest("GET", base+"/api/ping", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Host = "evil.example.com"
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("get /api/ping with a foreign host: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign Host must be refused, got %v", resp.Status)
	}
	resp.Body.Close()
}
