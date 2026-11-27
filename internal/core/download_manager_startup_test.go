package core

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The example must show the URL of the file to download and add a row even
// when the backend does not answer: the operator has a field to type the URL
// in any case, and the "+" button always adds a row.
func TestDownloadManagerTargetRowWithoutTheBackend(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("examples", "download_manager", "app.sui"))
	if err != nil {
		t.Fatalf("read the example: %v", err)
	}
	// The backend of the example is not running during the test: the port
	// points to a closed one, so the app has to work without an answer.
	code := strings.ReplaceAll(string(src), "127.0.0.1:8090", "127.0.0.1:9")
	runHeadlessFeature(t, code, featureScriptHead+`
    await wait(700);
    var rows = document.querySelectorAll('.target-row');
    if (rows.length < 1) throw new Error('the row of the target is missing');
    var url = rows[0].querySelector('.target-url');
    if (!url || url.tagName !== 'INPUT') throw new Error('the field of the URL is missing');
    var before = rows.length;
    var add = document.querySelector('.btn-add');
    if (!add) throw new Error('the add button is missing');
    add.click();
    await wait(600);
    var after = document.querySelectorAll('.target-row').length;
    if (after !== before + 1) throw new Error('the add button did not add a row: ' + before + ' -> ' + after);
`+featureScriptTail)
}

// The app finds the backend that served the page: the calls are relative, so
// the download works on every port and the app does not need to know the
// address of the backend.
func TestDownloadManagerFindsTheBackendThatServedIt(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("examples", "download_manager", "app.sui"))
	if err != nil {
		t.Fatalf("read the example: %v", err)
	}
	browser := FindHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found")
	}
	index, css, js, diags, ok := guiBuildFiles(string(src))
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	script := featureScriptHead + `
    await wait(900);
    var rows = document.querySelectorAll('.target-row');
    if (rows.length < 1) throw new Error('the row of the target is missing');
    var url = rows[0].querySelector('.target-url');
    if (!url) throw new Error('the field of the URL is missing');
    if (url.value !== 'http://example.com/from-the-config.bin') throw new Error('the backend was not found: ' + url.value);
    var name = rows[0].querySelector('.target-name');
    if (!name) throw new Error('the field of the name is missing');
    if (name.value !== 'from-the-config.bin') throw new Error('the name of the file was not filled: ' + name.value);
` + featureScriptTail
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + script + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}

	// The page and the API are served by the same server: the app has to
	// find the backend without knowing its address.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"http://example.com/from-the-config.bin","partSize":"7 MiB","workers":9}`))
	})
	mux.HandleFunc("GET /api/downloads", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jobs":[]}`))
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	go func() { _ = http.Serve(ln, mux) }()
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=900,700",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=30000", "--dump-dom",
		"http://" + ln.Addr().String() + "/",
	}
	out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
	text := string(out)
	if err != nil {
		t.Fatalf("headless browser failed: %v\n%s", err, text)
	}
	if strings.Contains(text, `id="e2e-result">OK<`) {
		return
	}
	t.Fatalf("the app did not find the backend:\n%s", text)
}
