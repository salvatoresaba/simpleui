package core

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestDownloadManagerApp checks the example front end: it must type-check and
// the generated JavaScript must keep the patterns the example teaches
// (polling with interval, a parts table filled from the API, a list of parts
// with one progress bar each, the tabs that keep the tables apart, and a
// dynamic URL sent through http.postSafe with a literal allow-list).
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
		"api/config",
		"table_display",
		"setInterval(",
		"http_postsafe",
		"http://127.0.0.1:8090",
		"/retry-failed",
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
// a real download from a URL: the file is split into parts, every part is
// fetched with its own Range request, a pause keeps the bytes, a resume
// continues from them and the reassembled file matches the original.
func TestDownloadManagerServerE2E(t *testing.T) {
	base, filesDir := startDownloadManager(t, "--sample-mib", "1", "--throttle", "131072")
	cl := &http.Client{Timeout: 20 * time.Second}

	// The CLI configuration is what the app starts with.
	cfg := decodeDownloadManager(t, getDownloadManager(t, cl, base+"/api/config")).(map[string]any)
	if cfg["partSize"] != "1 MiB" {
		t.Errorf("default part size = %v, want 1 MiB", cfg["partSize"])
	}

	// Start a download of a URL: the file is split into two 512 KiB parts.
	target := base + "/files/sample.bin"
	job := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+target+`","partSize":"512 KiB"}`).(map[string]any)
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
	paused := postDownloadManager(t, cl, base+"/api/downloads/"+id+"/pause", "").(map[string]any)
	if paused["state"] != "paused" {
		t.Errorf("state after pause = %v, want 'paused'", paused["state"])
	}
	pausedBytes := paused["received"].(float64)

	// Nothing must advance while the job is paused.
	time.Sleep(700 * time.Millisecond)
	status := decodeDownloadManager(t, getDownloadManager(t, cl, base+"/api/downloads")).(map[string]any)
	jobs := status["jobs"].([]any)
	stillPaused := jobs[0].(map[string]any)
	if stillPaused["state"] != "paused" {
		t.Errorf("state = %v, want 'paused'", stillPaused["state"])
	}
	if got := stillPaused["received"].(float64); got != pausedBytes {
		t.Errorf("bytes advanced while paused: %v -> %v", pausedBytes, got)
	}

	// Resume: the job finishes and the parts are reassembled.
	resumed := postDownloadManager(t, cl, base+"/api/downloads/"+id+"/resume", "").(map[string]any)
	if s := resumed["state"]; s != "running" {
		t.Errorf("state after resume = %v, want 'running'", s)
	}

	done := waitDownloadManager(t, cl, base, id)
	if done["state"] != "done" {
		t.Fatalf("final state = %v (%v)", done["state"], done["error"])
	}
	if v := done["verified"].(string); !strings.Contains(v, "matches") {
		t.Errorf("verification = %q, want the size to match the server", v)
	}
	if v := done["mergedSize"].(float64); v != 1024*1024 {
		t.Errorf("merged size = %v, want %v", v, 1024*1024)
	}

	// The reassembled file served back must be byte identical to the original.
	orig, err := os.ReadFile(filepath.Join(filesDir, "sample.bin"))
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	want := sha256.Sum256(orig)
	merged := getDownloadManager(t, cl, base+"/api/downloads/"+id+"/file")
	if len(merged) != len(orig) {
		t.Fatalf("merged size = %d, want %d", len(merged), len(orig))
	}
	got := sha256.Sum256([]byte(merged))
	if hex.EncodeToString(want[:]) != hex.EncodeToString(got[:]) {
		t.Errorf("merged SHA-256 = %x, want %x", got, want)
	}
}

// TestDownloadManagerRetryPartE2E drives the retry of a single failed part:
// the origin refuses one range once, the part is marked as failed, the job
// waits for the operator, and the retry of that part alone completes the
// download.
func TestDownloadManagerRetryPartE2E(t *testing.T) {
	payload := make([]byte, 1024*1024)
	for i := range payload {
		payload[i] = byte(i * 7)
	}
	const failedRange = "bytes=524288-1048575"
	origin, failures := flakyOrigin(payload, failedRange)
	defer origin.Close()

	base, _ := startDownloadManager(t, "--sample-mib", "0", "--throttle", "0")
	cl := &http.Client{Timeout: 20 * time.Second}

	job := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+origin.URL+`/file.bin","partSize":"512 KiB"}`).(map[string]any)
	id, _ := job["id"].(string)
	if id == "" {
		t.Fatalf("job without id: %v", job)
	}

	// The second part fails: the job stops and waits for the operator.
	var failed map[string]any
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		list := decodeDownloadManager(t, getDownloadManager(t, cl, base+"/api/downloads")).(map[string]any)["jobs"].([]any)
		cur := list[0].(map[string]any)
		if cur["state"] == "error" {
			failed = cur
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if failed == nil {
		t.Fatalf("the job did not report the failed part")
	}
	if got := failed["failedParts"].(float64); got != 1 {
		t.Errorf("failed parts = %v, want 1", got)
	}
	parts, _ := failed["parts"].([]any)
	errored := 0
	for _, p := range parts {
		pm := p.(map[string]any)
		if pm["State"] == "error" {
			errored++
			if pm["Part"].(float64) != 2 {
				t.Errorf("the wrong part failed: %v", pm["Part"])
			}
			if pm["Error"] == "" {
				t.Errorf("the failed part has no error message")
			}
		}
	}
	if errored != 1 {
		t.Fatalf("errored parts = %d, want 1", errored)
	}

	// Retry that part alone: the origin answers this time and the job ends.
	retried := postDownloadManager(t, cl, base+"/api/downloads/"+id+"/parts/2/retry", "").(map[string]any)
	if retried["state"] != "running" {
		t.Errorf("state after the retry = %v, want 'running'", retried["state"])
	}
	if *failures != 1 {
		t.Errorf("the origin failed %d times, want 1", *failures)
	}

	done := waitDownloadManager(t, cl, base, id)
	if done["state"] != "done" {
		t.Fatalf("final state = %v (%v)", done["state"], done["error"])
	}
	merged := getDownloadManager(t, cl, base+"/api/downloads/"+id+"/file")
	if len(merged) != len(payload) {
		t.Fatalf("merged size = %d, want %d", len(merged), len(payload))
	}
	want := sha256.Sum256(payload)
	got := sha256.Sum256([]byte(merged))
	if hex.EncodeToString(want[:]) != hex.EncodeToString(got[:]) {
		t.Errorf("merged SHA-256 = %x, want %x", got, want)
	}
}

// flakyOrigin serves a file with Range support and refuses one range once:
// the download of that part fails, the retry succeeds.
func flakyOrigin(payload []byte, failRange string) (*httptest.Server, *int) {
	var mu sync.Mutex
	failures := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		mu.Lock()
		if rangeHeader == failRange && failures == 0 {
			failures++
			mu.Unlock()
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		mu.Unlock()

		start, end := 0, len(payload)-1
		if spec := strings.TrimPrefix(rangeHeader, "bytes="); spec != rangeHeader {
			parts := strings.SplitN(spec, "-", 2)
			if len(parts) == 2 {
				if s, err := strconv.Atoi(parts[0]); err == nil {
					start = s
				}
				if e, err := strconv.Atoi(parts[1]); err == nil && e < end {
					end = e
				}
			}
		}
		if start < 0 {
			start = 0
		}
		if end >= len(payload) {
			end = len(payload) - 1
		}
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Type", "application/octet-stream")
		if rangeHeader != "" {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(payload)))
			w.WriteHeader(http.StatusPartialContent)
		}
		_, _ = w.Write(payload[start : end+1])
	}))
	return srv, &failures
}

// startDownloadManager builds and starts the backend of the example and
// returns its base URL and the folder with the local files.
func startDownloadManager(t *testing.T, extra ...string) (string, string) {
	t.Helper()
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

	args := []string{"--no-open", "--port", "0", "--web-dir", pubDir, "--files-dir", filesDir, "--out-dir", workDir}
	args = append(args, extra...)
	cmd := exec.Command(exe, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout: %v", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	return readDownloadManagerLine(t, stdout, &stderr), filesDir
}

// waitDownloadManager waits for a job to settle and returns its last state.
func waitDownloadManager(t *testing.T, cl *http.Client, base, id string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		list := decodeDownloadManager(t, getDownloadManager(t, cl, base+"/api/downloads")).(map[string]any)["jobs"].([]any)
		for _, j := range list {
			cur := j.(map[string]any)
			if cur["id"] != id {
				continue
			}
			if cur["state"] == "done" || cur["state"] == "error" || cur["state"] == "canceled" {
				return cur
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("the job did not finish")
	return nil
}

func getDownloadManager(t *testing.T, cl *http.Client, url string) string {
	t.Helper()
	resp, err := cl.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	return readAllClose(t, resp)
}

func postDownloadManager(t *testing.T, cl *http.Client, url, body string) any {
	t.Helper()
	resp, err := cl.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		payload := readAllClose(t, resp)
		t.Fatalf("post %s: HTTP %d: %s", url, resp.StatusCode, payload)
	}
	return decodeDownloadManager(t, readAllClose(t, resp))
}

// readDownloadManagerLine waits for the backend's startup line and returns
// the base URL. The rest of the output keeps being drained: a backend that
// blocks on a full stdout pipe would stop answering the API.
func readDownloadManagerLine(t *testing.T, stdout io.Reader, stderr *strings.Builder) string {
	lineCh := make(chan string, 1)
	go func() {
		r := bufio.NewReader(stdout)
		for {
			l, err := r.ReadString('\n')
			trimmed := strings.TrimSpace(l)
			if strings.Contains(trimmed, "running at") {
				select {
				case lineCh <- trimmed:
				default:
				}
			}
			if err != nil {
				if err != io.EOF {
					select {
					case lineCh <- "":
					default:
					}
				}
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
