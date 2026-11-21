package sui

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDesktopWebviewDialogTarget(t *testing.T) {
	for _, want := range []string{
		`id="desk-target"`,
		`value="desktop-webview"`,
		`Desktop WebView (native window)`,
		`fd.set("target"`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("desktop build dialog missing %q", want)
		}
	}
}

func TestDesktopBuildHandlerRejectsUnknownTarget(t *testing.T) {
	req := httptest.NewRequest("POST", "/__gui__/build-desktop", strings.NewReader("sui=view%20main%20%7B%7D&target=unknown&out=dist"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	guiDesktopBuildHandler(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "unknown desktop target") {
		t.Fatalf("unknown target response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestGoVersionHelpers(t *testing.T) {
	if got := parseGoVersionOutput("go version go1.24.2 windows/amd64"); got != "1.24.2" {
		t.Fatalf("parseGoVersionOutput = %q", got)
	}
	if !goVersionAtLeast("1.24.2", 1, 24, 2) || goVersionAtLeast("1.24.1", 1, 24, 2) || goVersionAtLeast("1.23.9", 1, 24, 2) {
		t.Fatal("goVersionAtLeast comparison is wrong")
	}
	if got := splitDesktopList(" C:/a\r\n\r\nC:/b \n"); len(got) != 2 || got[0] != "C:/a" || got[1] != "C:/b" {
		t.Fatalf("splitDesktopList = %#v", got)
	}
}

func TestDesktopPreflightValidation(t *testing.T) {
	if _, _, _, err := desktopGoInfo(); err != nil {
		t.Skip("Go toolchain unavailable: " + err.Error())
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "build")
	src := "view main {\n\th1 \"hi\"\n}\n"
	res, err := desktopPreflight(src, out, nil, nil)
	if err != nil {
		t.Fatalf("valid preflight: %v", err)
	}
	if res.GoVersion == "" || res.OutDir != out || res.Prog == nil {
		t.Fatalf("incomplete preflight result: %#v", res)
	}

	if _, err := desktopPreflight("view main {\n\th9 \"bad\"\n}\n", out, nil, nil); err == nil {
		t.Fatal("invalid source passed preflight")
	}

	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := desktopPreflight(src, file, nil, nil); err == nil || !strings.Contains(err.Error(), "output path is a file") {
		t.Fatalf("output file was accepted: %v", err)
	}
	if _, err := desktopPreflight(src, out, []string{filepath.Join(dir, "missing")}, nil); err == nil {
		t.Fatal("missing allowed directory was accepted")
	}
}

func TestDesktopDownloadToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.exe")
	if err := os.WriteFile(path, []byte("bin"), 0o644); err != nil {
		t.Fatal(err)
	}
	token := registerDesktopDownload(path)
	if got, ok := lookupDesktopDownload(token); !ok || got != path {
		t.Fatalf("lookupDesktopDownload = %q, %v", got, ok)
	}
	if _, ok := lookupDesktopDownload("wrong"); ok {
		t.Fatal("invalid token was accepted")
	}
}
