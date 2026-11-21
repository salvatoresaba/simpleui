package sui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// desktop_test.go - the desktop target: JS generation variants and the
// end-to-end Bridge test against a real compiled desktop executable.

func TestJsgenDesktopVariant(t *testing.T) {
	src := "import { exec, fs } from \"std\"\n" +
		"\n" +
		"async fn run() -> null {\n" +
		"\tout, err1 = exec.run(\"dir\", [])\n" +
		"\tdata, err2 = fs.read(\"C:\\\\x\")\n" +
		"}\n"
	prog, dl := ParseSource(src, "d.sui")
	if dl.Count() != 0 {
		t.Fatalf("parse diagnostics: %v", dl.String())
	}
	desk := GenerateJsFor(prog, "desktop")
	for _, want := range []string{
		"const __SUI_TOKEN = \"__SUI_TOKEN__\";",
		"/__bridge__/exec",
		"/__bridge__/file/read",
		"fd.set(\"cmd\", cmd)",
		"for (const a of args) fd.append",
	} {
		if !strings.Contains(desk, want) {
			t.Errorf("desktop jsgen missing %q", want)
		}
	}
	if strings.Contains(desk, "disabled on the web target") {
		t.Errorf("desktop jsgen must not use the web-disabled message")
	}
	web := GenerateJs(prog)
	if !strings.Contains(web, "exec.run is disabled on the web target") {
		t.Errorf("web jsgen should keep the disabled exec stub")
	}
	if strings.Contains(web, "/__bridge__/") {
		t.Errorf("web jsgen must not reference the desktop bridge")
	}
}

func TestDesktopMainSourceAllowLists(t *testing.T) {
	allow := []string{"ls", "git status"}
	src := DesktopMainSource(allow, []string{`C:\data`})
	if !strings.Contains(src, `var allowExec = []string{"ls", "git status"}`) {
		t.Errorf("allowExec literal wrong:\n%v", src[:600])
	}
	if !strings.Contains(src, `var allowDirs = []string{"C:\\data"}`) {
		t.Errorf("allowDirs literal wrong")
	}
	if strings.Contains(src, "%%ALLOW") {
		t.Errorf("placeholder left unexpanded")
	}
}

func TestDesktopWebviewMainSource(t *testing.T) {
	src := DesktopWebviewMainSource([]string{"git"}, []string{`C:\data`}, `My "App"`)
	for _, want := range []string{
		`webview "github.com/webview/webview_go"`,
		`if noOpen {`,
		`w := webview.New(false)`,
		`w.SetTitle("My \"App\"")`,
		`w.SetSize(1200, 800, webview.HintNone)`,
		`w.Navigate(url)`,
		`w.Run()`,
		`var allowExec = []string{"git"}`,
		`var allowDirs = []string{"C:\\data"}`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("webview main source missing %q", want)
		}
	}
	if strings.Contains(src, "if !noOpen") || strings.Contains(src, "%%ALLOW") {
		t.Errorf("webview main source still contains browser launch or placeholders")
	}
}

func TestDesktopWebviewBuild(t *testing.T) {
	if os.Getenv("SUI_TEST_WEBVIEW") != "1" {
		t.Skip("set SUI_TEST_WEBVIEW=1 to compile the embedded-WebView target")
	}
	prog, dl := ParseSource("view main {\n\th1 \"Hello WebView\"\n}\n", "app.sui")
	if dl.Count() != 0 {
		t.Fatalf("parse diagnostics: %v", dl.String())
	}
	dir := t.TempDir()
	if err := RunDesktopWebviewBuild(prog, dir, nil, nil); err != nil {
		t.Fatalf("webview build failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, desktopOutName())); err != nil {
		t.Fatalf("webview executable missing: %v", err)
	}
}
