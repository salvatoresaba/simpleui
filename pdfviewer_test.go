package sui

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// pdfviewer_test.go - the PDF viewer control: compile-time validation, the
// generated markup and the desktop document endpoint that feeds it.

// minimalPdf returns a small but valid PDF document.
func minimalPdf() []byte {
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 120] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		"<< /Length 44 >>\nstream\nBT /F1 16 Tf 20 60 Td (SimpleUI test) Tj ET\nendstream",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}
	var sb strings.Builder
	sb.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = sb.Len()
		sb.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, obj))
	}
	xref := sb.Len()
	sb.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objects)+1))
	for i := 1; i <= len(objects); i++ {
		sb.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	sb.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref))
	return []byte(sb.String())
}

func mustPdfViewerProgram(t *testing.T, src string) (*Prog, DiagList) {
	t.Helper()
	prog, dl := ParseSource(src, "pdfviewer.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	return prog, Check(prog)
}

// The frame must not be sandboxed (the native viewer does not start inside a
// sandboxed frame) but it must always carry the document type and the
// referrer policy that keeps the document URL out of the Referer header.
func TestPdfViewerSecureMarkup(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	prog, dl := mustPdfViewerProgram(t, "view main {\n\tpdfviewer src: \"https://example.com/a.pdf\"\n}\n")
	if dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "type=\"application/pdf\"") {
		t.Error("pdfviewer is missing the application/pdf document type")
	}
	if !strings.Contains(js, "referrerpolicy=\"no-referrer\"") {
		t.Error("pdfviewer is missing the no-referrer referrer policy")
	}
	if strings.Contains(js, "sui-pdfviewer\" sandbox=") {
		t.Error("pdfviewer must not be sandboxed: the native PDF viewer does not start there")
	}
	if !strings.Contains(js, "sui-pdfviewer") {
		t.Error("pdfviewer is missing its base class")
	}
	if !strings.Contains(js, "function __sui_pdf_error(") {
		t.Error("pdfviewer runtime error box is missing")
	}
	if !strings.Contains(js, "__sui_iframe_src(") {
		t.Error("pdfviewer does not validate its source at runtime")
	}
	// The static export (designer preview) uses the same frame.
	html := GenerateHtmlPage(prog)
	if !strings.Contains(html, "sui-pdfviewer") || !strings.Contains(html, "application/pdf") {
		t.Errorf("static HTML misses the pdf viewer frame:\n%s", html)
	}
}

func TestPdfViewerDynamicSrcRequiresAllowList(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	src := "state u = \"https://example.com/a.pdf\"\nview main {\n\tpdfviewer src: u\n}\n"
	_, dl := mustPdfViewerProgram(t, src)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "add srcAllow") {
		t.Fatalf("dynamic src without allow-list was accepted: %v", dl.String())
	}
	src = "state u = \"https://example.com/a.pdf\"\nview main {\n\tpdfviewer src: u srcAllow: \"https://example.com\"\n}\n"
	prog, dl := mustPdfViewerProgram(t, src)
	if dl.HasErrors() {
		t.Fatalf("dynamic src with allow-list rejected: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "__sui_iframe_src(") {
		t.Error("dynamic src is not validated at runtime")
	}
}

func TestPdfViewerRejectsMalformedData(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	cases := []struct {
		src  string
		want string
	}{
		{"view main {\n\tpdfviewer\n}\n", "src is required"},
		{"view main {\n\tpdfviewer \"fallback\" src: \"https://example.com/a.pdf\"\n}\n", "cannot accept pdfviewer content"},
		{"view main {\n\tpdfviewer src: \"javascript:alert(1)\"\n}\n", "cannot accept pdfviewer src data"},
		{"view main {\n\tpdfviewer src: \"file:///etc/passwd\"\n}\n", "cannot accept pdfviewer src data"},
		{"state a = \"https://example.com\"\nview main {\n\tpdfviewer src: \"https://example.com/a.pdf\" srcAllow: a\n}\n", "cannot accept dynamic srcAllow data"},
		{"state a = \"*\"\nview main {\n\tpdfviewer src: \"https://example.com/a.pdf\" srcAllow: a\n}\n", "cannot accept dynamic srcAllow data"},
	}
	for _, c := range cases {
		prog, _ := ParseSource(c.src, "pdfviewer.sui")
		if prog == nil {
			continue
		}
		dl := Check(prog)
		if !dl.HasErrors() || !strings.Contains(dl.String(), c.want) {
			t.Errorf("expected %q, got: %v", c.want, dl.String())
		}
	}
	// A literal allow-list with a wildcard is rejected.
	prog, _ := ParseSource("view main {\n\tpdfviewer src: \"https://example.com/a.pdf\" srcAllow: \"*\"\n}\n", "pdfviewer.sui")
	if prog != nil {
		dl := Check(prog)
		if !dl.HasErrors() || !strings.Contains(dl.String(), "cannot accept pdfviewer srcAllow data") {
			t.Errorf("wildcard allow-list was accepted: %v", dl.String())
		}
	}
}

func TestPdfViewerPolicy(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	src := "state u = \"https://example.com/a.pdf\"\nview main {\n\tpdfviewer src: u\n}\n"

	// pdf.dynamicSrc = allow drops the allow-list requirement.
	cfg := DefaultSecurity()
	cfg.PDFDynamicSrc = "allow"
	SetDefaultSecurity(cfg)
	if _, dl := mustPdfViewerProgram(t, src); dl.HasErrors() {
		t.Fatalf("allow mode should compile: %v", dl.String())
	}

	// pdf.unsandboxedFrame = deny removes the control from the language.
	cfg = DefaultSecurity()
	cfg.PDFUnsandboxed = "deny"
	SetDefaultSecurity(cfg)
	_, dl := mustPdfViewerProgram(t, src)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "disabled by the security policy") {
		t.Fatalf("deny mode should reject the control: %v", dl.String())
	}
}

// The designer must model the pdf viewer like any other element, otherwise
// opening and saving a project in the designer drops the document.
func TestGuiPdfViewerAttributesRoundTrip(t *testing.T) {
	withSecurityTestGlobals(t)
	SetDefaultSecurity(DefaultSecurity())
	src := "state u = \"https://example.com/a.pdf\"\n" +
		"view main {\n\tpdfviewer src: u srcAllow: \"https://example.com\" width: \"640\" height: \"360\"\n}\n"
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen failed: %v", StringifyJson(msgs))
	}
	raw := StringifyJson(model)
	for _, want := range []string{
		`"kind":"pdfviewer"`,
		`"srcExpr":true`,
		`"srcAllow":"https://example.com"`,
		`"width":"640"`,
		`"height":"360"`,
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("designer model lost pdfviewer attribute %s: %s", want, raw)
		}
	}
	// The designer palette offers the control.
	page, err := os.ReadFile("assets/designer.html")
	if err != nil {
		t.Fatalf("read designer: %v", err)
	}
	for _, want := range []string{`"pdfviewer"`, `pdfviewer: "PDF viewer"`, `pdfviewer:1`} {
		if !strings.Contains(string(page), want) {
			t.Errorf("designer page is missing %q", want)
		}
	}
}

// Inserting a PDF viewer from the palette must produce a design that
// compiles: an empty src is a compile error, so the designer points the new
// control at the sample document it serves for the preview.
func TestGuiPdfViewerInsertsSampleDocument(t *testing.T) {
	page, err := os.ReadFile("assets/designer.html")
	if err != nil {
		t.Fatalf("read designer: %v", err)
	}
	src := string(page)
	for _, want := range []string{
		"function newPdfViewerNode()",
		`n.src = "sample.pdf"`,
		`if (kind === "pdfviewer") return newPdfViewerNode();`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("designer page is missing %q", want)
		}
	}
	// The sample is a real PDF document, not a placeholder.
	if !bytes.HasPrefix(guiSamplePDF, []byte("%PDF-")) {
		t.Error("the designer sample is not a PDF document")
	}
	if !bytes.Contains(guiSamplePDF, []byte("%%EOF")) {
		t.Error("the designer sample PDF is truncated")
	}
	// The designer serves it for the preview: the relative src resolves
	// inside /__gui__/, where the preview frame lives.
	rec := httptest.NewRecorder()
	guiSamplePDFHandler(rec, httptest.NewRequest("GET", "/__gui__/sample.pdf", nil))
	if rec.Code != 200 {
		t.Fatalf("sample status: %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("sample content type: %q", ct)
	}
	if rec.Body.Len() != len(guiSamplePDF) {
		t.Errorf("sample body: %d bytes, want %d", rec.Body.Len(), len(guiSamplePDF))
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF-")) {
		t.Error("the served sample is not a PDF document")
	}
}

func TestPdfViewerSecurityFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "simpleui.security.json")
	body := "{\"pdf\":{\"dynamicSrc\":\"allow\",\"unsandboxedFrame\":\"deny\"}}"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	cfg, _, err := LoadSecurityFile(path)
	if err != nil {
		t.Fatalf("load policy: %v", err)
	}
	if cfg.PDFDynamicSrc != "allow" || cfg.PDFUnsandboxed != "deny" {
		t.Fatalf("policy not applied: %+v", cfg)
	}
	// An unknown value is rejected instead of being ignored.
	if err := os.WriteFile(path, []byte("{\"pdf\":{\"dynamicSrc\":\"maybe\"}}"), 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	if _, _, err := LoadSecurityFile(path); err == nil {
		t.Fatal("an invalid pdf policy value was accepted")
	}
}

// The desktop bridge serves the document to the viewer frame: it needs the
// per-run document cookie, it only serves PDF files and it honours the
// allowed directories.
func TestDesktopDocumentEndpointE2E(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "report.pdf"), minimalPdf(), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a pdf"), 0o644); err != nil {
		t.Fatalf("write text: %v", err)
	}
	fixture := "import { fs } from \"std\"\n" +
		"\n" +
		"fn open() -> null {\n" +
		"\turl, err = fs.documentUrl(\"report.pdf\")\n" +
		"\t}\n" +
		"\n" +
		"view main {\n" +
		"\tpdfviewer src: \"https://example.com/a.pdf\"\n" +
		"}\n"
	srcFile := filepath.Join(dir, "demo.sui")
	if err := os.WriteFile(srcFile, []byte(fixture), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	prog, dl := LoadProgram(srcFile)
	if dl.HasErrors() {
		t.Fatalf("load diagnostics:\n%v", dl.String())
	}
	if dl = Check(prog); dl.HasErrors() {
		t.Fatalf("check diagnostics:\n%v", dl.String())
	}
	buildDir := filepath.Join(dir, "build")
	if err := RunDesktopBuild(prog, buildDir, nil, []string{dir}); err != nil {
		t.Fatalf("desktop build: %v", err)
	}
	exe := filepath.Join(buildDir, "app")
	if runtime.GOOS == "windows" {
		exe += ".exe"
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

	// The app page hands out the document cookie.
	page, err := client.Get(base + "/")
	if err != nil {
		t.Fatalf("get app page: %v", err)
	}
	_, _ = io.Copy(io.Discard, page.Body)
	page.Body.Close()
	cookie := ""
	for _, c := range page.Cookies() {
		if c.Name == "sui_doc" {
			cookie = c.Value
		}
	}
	if cookie == "" {
		t.Fatal("the app page did not set the document cookie")
	}

	docURL := base + "/__bridge__/file/document?path=" + urlEscape(filepath.Join(dir, "report.pdf"))
	get := func(target string, withCookie bool) (*http.Response, []byte) {
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		if withCookie {
			req.AddCookie(&http.Cookie{Name: "sui_doc", Value: cookie})
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("get %s: %v", target, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, body
	}

	// 1. The document is served with the cookie.
	resp, body := get(docURL, true)
	if resp.StatusCode != 200 {
		t.Fatalf("document status: %d (%s)", resp.StatusCode, string(body))
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("document content type: %q", ct)
	}
	if !strings.HasPrefix(string(body), "%PDF-") {
		t.Errorf("document is not a PDF: %q", string(body[:min(8, len(body))]))
	}
	if resp.Header.Get("X-Frame-Options") != "" || resp.Header.Get("Content-Security-Policy") != "" {
		t.Error("the document must be embeddable: it may not carry anti-framing headers")
	}
	if resp.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Error("the document must not leak its URL through the Referer header")
	}

	// 2. Without the cookie the endpoint refuses.
	resp, _ = get(docURL, false)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("document without cookie: status %d", resp.StatusCode)
	}

	// 3. A file that is not a PDF is refused.
	notPDF := base + "/__bridge__/file/document?path=" + urlEscape(filepath.Join(dir, "notes.txt"))
	resp, _ = get(notPDF, true)
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("non-pdf document: status %d", resp.StatusCode)
	}

	// 4. A path outside the allowed directories is refused.
	outside := base + "/__bridge__/file/document?path=" + urlEscape(filepath.Join(dir, "..", "outside.pdf"))
	resp, _ = get(outside, true)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("document outside the allowed dirs: status %d", resp.StatusCode)
	}

	// 5. A range request returns the requested slice.
	req, err := http.NewRequest("GET", docURL, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "sui_doc", Value: cookie})
	req.Header.Set("Range", "bytes=0-4")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("range request: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		t.Errorf("range request: status %d", resp.StatusCode)
	}
	if string(body) != "%PDF-" {
		t.Errorf("range request returned %q", string(body))
	}
}

// urlEscape percent-encodes a path for a query string.
func urlEscape(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9',
			c == '-', c == '_', c == '.', c == '~':
			sb.WriteByte(c)
		default:
			sb.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
