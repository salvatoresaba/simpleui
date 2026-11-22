package sui

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// pdfviewer_blob_test.go - the PDF viewer with a document the user picked in
// the browser.
//
// On the web target the picked file becomes a blob: URL: the native viewer
// loads it, so no backend is needed. The blob is only accepted when the
// project allows it, only for the app's own origin and only after the same
// allow-list check that every other dynamic source goes through.

const pdfViewerBlobSource = `import { files } from "std"

state docUrl = ""

async fn openPicked() -> null {
	url, err = files.documentUrl("report.pdf")
	if err == null {
		docUrl = url
	}
}

view main {
	pdfviewer src: docUrl srcAllow: "/" width: "900" height: "620"
}
`

// The policy decides whether the runtime accepts a blob: source.
func TestPdfViewerBlobSrcPolicy(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	prog, dl := mustPdfViewerProgram(t, pdfViewerBlobSource)
	if dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}

	// Default: blob sources are allowed, so the guard is called with the
	// blob flag on.
	js := GenerateJs(prog)
	if !strings.Contains(js, `__sui_pdf_src(docUrl, "/", true, true)`) {
		t.Errorf("the pdfviewer does not accept blob sources by default:\n%s", pdfGuardCall(js))
	}

	// pdf.blobSrc = deny turns the blob branch off.
	cfg := DefaultSecurity()
	cfg.PDFBlobSrc = "deny"
	SetDefaultSecurity(cfg)
	prog, dl = mustPdfViewerProgram(t, pdfViewerBlobSource)
	if dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js = GenerateJs(prog)
	if !strings.Contains(js, `__sui_pdf_src(docUrl, "/", true, false)`) {
		t.Errorf("pdf.blobSrc=deny did not disable the blob branch:\n%s", pdfGuardCall(js))
	}
}

// pdfGuardCall returns the guard call of the generated pdfviewer, so a failed
// assertion shows what the compiler actually emitted.
func pdfGuardCall(js string) string {
	idx := strings.Index(js, "__sui_pdf_src(")
	if idx < 0 {
		return "no guard call found"
	}
	end := idx + 120
	if end > len(js) {
		end = len(js)
	}
	return js[idx:end]
}

// The runtime guard: a blob of the app's own origin is accepted, a blob of
// another origin and a blob when the policy denies it are refused, and the
// http(s) sources keep working exactly as before.
func TestPdfViewerBlobSrcRuntime(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not available")
	}
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	prog, dl := mustPdfViewerProgram(t, pdfViewerBlobSource)
	if dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)

	harness := "global.location = { origin: 'http://localhost:8204' };\n" +
		extractJsFunc(js, "__sui_iframe_src") + "\n" +
		extractJsFunc(js, "__sui_pdf_src") + "\n" +
		`var cases = [
			['blob-same-origin', 'blob:http://localhost:8204/abc', '/', true, true],
			['blob-other-origin', 'blob:http://evil.example/abc', '/', true, true],
			['blob-policy-off', 'blob:http://localhost:8204/abc', '/', true, false],
			['blob-not-in-allow-list', 'blob:http://localhost:8204/abc', '/docs', true, true],
			['https-allowed', 'https://example.com/a.pdf', 'https://example.com', true, true],
			['https-not-allowed', 'https://evil.example/a.pdf', 'https://example.com', true, true],
			['data-url', 'data:application/pdf;base64,AAAA', '/', true, true]
		];
		var out = [];
		for (var i = 0; i < cases.length; i++) {
			var c = cases[i];
			out.push([c[0], __sui_pdf_src(c[1], c[2], c[3], c[4])]);
		}
		console.log(JSON.stringify(out));
`
	dir := t.TempDir()
	script := filepath.Join(dir, "harness.js")
	if err := os.WriteFile(script, []byte(harness), 0o644); err != nil {
		t.Fatalf("write harness: %v", err)
	}
	cmd := exec.Command(nodePath, script)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("node harness: %v", err)
	}
	var results [][]any
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatalf("harness output %q: %v", string(out), err)
	}
	if len(results) != 7 {
		t.Fatalf("unexpected harness results: %s", string(out))
	}
	// [name, [url, reason]]
	urlOf := func(i int) string {
		pair, _ := results[i][1].([]any)
		s, _ := pair[0].(string)
		return s
	}
	reasonOf := func(i int) string {
		pair, _ := results[i][1].([]any)
		s, _ := pair[1].(string)
		return s
	}
	if url := urlOf(0); url != "blob:http://localhost:8204/abc" {
		t.Errorf("a blob of the app's own origin was refused: %v", results[0][1])
	}
	if reason := reasonOf(1); reason != "URL not allowed" {
		t.Errorf("a blob of another origin was accepted: %v", results[1][1])
	}
	if reason := reasonOf(2); reason != "blob URLs are not allowed" {
		t.Errorf("a blob was accepted with the policy off: %v", results[2][1])
	}
	if reason := reasonOf(3); reason != "URL not allowed" {
		t.Errorf("a blob outside the allow-list was accepted: %v", results[3][1])
	}
	if url := urlOf(4); url != "https://example.com/a.pdf" {
		t.Errorf("an allowed https source was refused: %v", results[4][1])
	}
	if reason := reasonOf(5); reason != "URL not allowed" {
		t.Errorf("an https source outside the allow-list was accepted: %v", results[5][1])
	}
	if reason := reasonOf(6); reason != "scheme not allowed" {
		t.Errorf("a data: source was accepted: %v", results[6][1])
	}
}

// extractJsFunc cuts a function out of the generated JavaScript.
func extractJsFunc(js, name string) string {
	start := strings.Index(js, "function "+name+"(")
	if start < 0 {
		return ""
	}
	end := strings.Index(js[start:], "\n}\n")
	if end < 0 {
		return ""
	}
	return js[start : start+end+3]
}

// files.documentUrl is the web helper that builds the blob: URL of a file the
// user picked: the bytes are checked and the blob type is forced.
func TestFilesDocumentUrlStdlib(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	prog, dl := ParseSource(pdfViewerBlobSource, "pdfviewer_blob.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl = Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, frag := range []string{
		"files_documenturl:",
		"await __SUI_STD.files_documenturl(",
		"__SUI_STD_EXTRA.files.documentUrl",
		"%PDF-",
		"application/pdf",
		"URL.createObjectURL(blob)",
	} {
		if !strings.Contains(js, frag) {
			t.Errorf("generated app.js missing %q", frag)
		}
	}
}

// The helper is web-only: on desktop the file system is the source of truth.
func TestFilesDocumentUrlDesktopHint(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	prog, dl := ParseSource(pdfViewerBlobSource, "pdfviewer_blob.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	js := GenerateJsFor(prog, "desktop")
	if !strings.Contains(js, "files.documentUrl is not available on the desktop target") {
		t.Error("the desktop build does not point to fs.documentUrl(path)")
	}
	if !strings.Contains(js, "use fs.documentUrl(path)") {
		t.Error("the desktop build does not point to fs.documentUrl(path)")
	}
}
