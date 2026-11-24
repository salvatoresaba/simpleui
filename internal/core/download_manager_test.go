package core

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestDownloadManagerApp checks the example front end: it must type-check and
// the generated JavaScript must keep the patterns the example teaches
// (polling with interval, a parts table filled from the API, and a dynamic URL
// sent through http.postSafe with a literal allow-list).
func TestDownloadManagerApp(t *testing.T) {
	prog, dl := LoadProgram(filepath.Join("examples", "download_manager", "app.sui"))
	if dl.HasErrors() {
		t.Fatalf("load: %v", dl.String())
	}
	if dl = Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"api/downloads",
		"api/files",
		"table_display",
		"setInterval(",
		"http_postsafe",
		"http://127.0.0.1:8090",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("download manager js missing %q", want)
		}
	}
	html := GenerateIndexHtml(prog)
	if !strings.Contains(html, "<title>Download manager</title>") {
		t.Errorf("download manager html title mismatch")
	}
}

// TestDownloadManagerServerE2E runs the Go backend of the example and drives
// a real download: the file is split into parts, every part is fetched with
// its own Range request, a pause keeps the bytes, a resume continues from
// them and the reassembled file matches the original SHA-256.
func TestDownloadManagerServerE2E(t *testing.T) {
	dir := t.TempDir()
	pubDir := filepath.Join(dir, "public")
	filesDir := filepath.Join(dir, "files")
	workDir := filepath.Join(dir, "work")
	for _, d := range []string{pubDir, filesDir, workDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}

	prog, dl := LoadProgram(filepath.Join("examples", "download_manager", "app.sui"))
	if dl.HasErrors() {
		t.Fatalf("load: %v", dl.String())
	}
	for _, f := range []struct{ name, body string }{
		{"index.html", GenerateIndexHtml(prog)},
		{"style.css", GenerateCss(prog)},
		{"app.js", GenerateJs(prog)},
	} {
		if err := os.WriteFile(filepath.Join(pubDir, f.name), []byte(f.body), 0o644); err != nil {
			t.Fatalf("write asset: %v", err)
		}
	}

	exeName := "downloadmanager"
	if runtime.GOOS == "windows" {
		exeName = exeName + ".exe"
	}
	exe := filepath.Join(dir, exeName)
	ce := exec.Command("go", "build", "-o", exe, ".")
	ce.Dir = filepath.Join("examples", "download_manager", "server")
	if out, err := ce.CombinedOutput(); err != nil {
		t.Fatalf("server build: %v\n%v", err, string(out))
	}

	cmd := exec.Command(exe, "--no-open", "--port", "0", "--web-dir", pubDir, "--files-dir", filesDir,
		"--work-dir", workDir, "--sample-mib", "1", "--throttle", "131072")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout: %v", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	base := readDownloadManagerLine(t, stdout, &stderr)
	cl := &http.Client{Timeout: 20 * time.Second}

	// The catalog lists the generated sample file.
	resp, err := cl.Get(base + "/api/files")
	if err != nil {
		t.Fatalf("api files: %v", err)
	}
	files := decodeDownloadManager(t, readAllClose(t, resp))
	catalog, ok := files.([]any)
	if !ok || len(catalog) == 0 {
		t.Fatalf("catalog unexpected: %v", files)
	}
	entry, _ := catalog[0].(map[string]any)
	if entry["Name"] != "sample.bin" {
		t.Errorf("catalog first entry = %v", entry["Name"])
	}

	// Start a download: 1 MiB split into 512 KiB parts.
	create, err := cl.Post(base+"/api/downloads", "application/json",
		strings.NewReader(`{"file":"sample.bin","partSize":"512 KiB"}`))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	job := decodeDownloadManager(t, readAllClose(t, create)).(map[string]any)
	id, _ := job["id"].(string)
	if id == "" {
		t.Fatalf("job without id: %v", job)
	}
	if got := job["partCount"].(float64); got != 2 {
		t.Errorf("part count = %v, want 2", got)
	}
	parts, _ := job["parts"].([]any)
	if len(parts) != 2 {
		t.Fatalf("parts = %d, want 2", len(parts))
	}
	first, _ := parts[0].(map[string]any)
	if first["Size"].(float64) != 512*1024 {
		t.Errorf("part size = %v", first["Size"])
	}

	// Pause right away: the parts keep the bytes they already have.
	pause, err := cl.Post(base+"/api/downloads/"+id+"/pause", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatalf("pause: %v", err)
	}
	paused := decodeDownloadManager(t, readAllClose(t, pause)).(map[string]any)
	if paused["state"] != "paused" {
		t.Errorf("state after pause = %v, want 'paused'", paused["state"])
	}
	pausedBytes := paused["received"].(float64)

	// Nothing must advance while the job is paused.
	time.Sleep(700 * time.Millisecond)
	status, err := cl.Get(base + "/api/downloads")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	jobs := decodeDownloadManager(t, readAllClose(t, status)).(map[string]any)["jobs"].([]any)
	stillPaused := jobs[0].(map[string]any)
	if stillPaused["state"] != "paused" {
		t.Errorf("state = %v, want 'paused'", stillPaused["state"])
	}
	if got := stillPaused["received"].(float64); got != pausedBytes {
		t.Errorf("bytes advanced while paused: %v -> %v", pausedBytes, got)
	}

	// Resume: the job finishes and the parts are reassembled.
	resume, err := cl.Post(base+"/api/downloads/"+id+"/resume", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if s := decodeDownloadManager(t, readAllClose(t, resume)).(map[string]any)["state"]; s != "running" {
		t.Errorf("state after resume = %v, want 'in corso'", s)
	}

	deadline := time.Now().Add(60 * time.Second)
	var done map[string]any
	for time.Now().Before(deadline) {
		resp, err := cl.Get(base + "/api/downloads")
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		list := decodeDownloadManager(t, readAllClose(t, resp)).(map[string]any)["jobs"].([]any)
		cur := list[0].(map[string]any)
		if cur["state"] == "done" || cur["state"] == "error" {
			done = cur
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if done == nil {
		t.Fatalf("download did not finish")
	}
	if done["state"] != "done" {
		t.Fatalf("final state = %v (%v)", done["state"], done["error"])
	}
	if v := done["verified"].(string); !strings.Contains(v, "matches") {
		t.Errorf("verification = %q, want the SHA-256 to match the original", v)
	}
	if v := done["mergedSize"].(float64); v != 1024*1024 {
		t.Errorf("merged size = %v, want %v", v, 1024*1024)
	}

	// The reassembled file served back must be byte identical to the original.
	orig, err := os.ReadFile(filepath.Join(filesDir, "sample.bin"))
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	origSum := sha256.Sum256(orig)
	file, err := cl.Get(base + "/api/downloads/" + id + "/file")
	if err != nil {
		t.Fatalf("merged file: %v", err)
	}
	merged := readAllClose(t, file)
	if len(merged) != len(orig) {
		t.Fatalf("merged size = %d, want %d", len(merged), len(orig))
	}
	mergedSum := sha256.Sum256([]byte(merged))
	if hex.EncodeToString(origSum[:]) != hex.EncodeToString(mergedSum[:]) {
		t.Errorf("merged SHA-256 = %x, want %x", mergedSum, origSum)
	}
}

// readDownloadManagerLine waits for the backend's startup line and returns
// the base URL.
func readDownloadManagerLine(t *testing.T, stdout io.Reader, stderr *strings.Builder) string {
	lineCh := make(chan string, 1)
	go func() {
		r := bufio.NewReader(stdout)
		for {
			l, err := r.ReadString('\n')
			if err != nil && err != io.EOF {
				lineCh <- ""
				return
			}
			trimmed := strings.TrimSpace(l)
			if strings.Contains(trimmed, "running at") {
				lineCh <- trimmed
				return
			}
			if err == io.EOF {
				lineCh <- ""
				return
			}
		}
	}()
	select {
	case line := <-lineCh:
		base := strings.TrimPrefix(line, "download manager + app running at ")
		base = strings.TrimSuffix(base, "/")
		if !strings.HasPrefix(base, "http://") {
			t.Fatalf("server line: %q (stderr: %v)", line, stderr.String())
		}
		return base
	case <-time.After(30 * time.Second):
		t.Fatalf("server did not report a URL (stderr: %v)", stderr.String())
		return ""
	}
}

// decodeDownloadManager decodes a JSON payload of the example backend.
func decodeDownloadManager(t *testing.T, body string) any {
	t.Helper()
	var out any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("json %v: %v", body, err)
	}
	return out
}
