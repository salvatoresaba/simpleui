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
		// The backend is found at the start: the calls are relative when the
		// page comes from the backend itself, and the address of the backend
		// is the fallback when it does not.
		"apiUrl",
		"apiFallback",
		"apiRoot",
		"/api",
		"table_display",
		"setInterval(",
		"http_postsafe",
		"http://127.0.0.1:8090",
		"/retry-failed",
		// The name of the file to save: the app fills it from the URL and
		// sends it with the request that starts the download.
		"syncTargetName",
		"fileNameFromUrl",
		`"file": name`,
		// The counters of the parts.
		"setPartTotal",
		"setPartActive",
		"setPartDone",
		// The targets: one row per file. The plus button adds a row and the
		// priority of a file is changed with the priority endpoint.
		"addTarget",
		"setTargets",
		"setFiles",
		"setFilesTable",
		"onTargetPriority",
		"setFilePriority",
		"/files/",
		"/priority",
		// The buttons of a row act on the file of that row: the handler
		// receives the key of its row.
		"onTargetPlay",
		"onTargetStop",
		"onTargetCancel",
		"restartTarget",
		"fileAction",
		"targetLocked",
		"__sui_rowkey",
		// A refused request is shown, not swallowed.
		"could not start: ",
		"could not add the file: ",
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

// TestDownloadManagerFileNameE2E checks the name of the file to save: the one
// the operator chooses in the GUI is the one the backend uses, a name that is
// not a plain file name is cleaned, and a request without a name keeps the
// name that comes from the URL.
func TestDownloadManagerFileNameE2E(t *testing.T) {
	base, _, _ := startDownloadManager(t, "--sample-mib", "1", "--throttle", "0")
	cl := &http.Client{Timeout: 20 * time.Second}
	target := base + "/files/sample.bin"

	// No name in the request: the name comes from the URL.
	auto := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+target+`","partSize":"1 MiB"}`).(map[string]any)
	if got := auto["file"]; got != "sample.bin" {
		t.Errorf("file name = %v, want sample.bin (the name of the URL)", got)
	}

	// The name the operator chose: it is the name of the job.
	named := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+target+`","partSize":"1 MiB","file":"relazione finale 2026.pdf"}`).(map[string]any)
	if got := named["file"]; got != "relazione_finale_2026.pdf" {
		t.Errorf("file name = %v, want relazione_finale_2026.pdf", got)
	}

	// A name that tries to escape the folder of the job: the separators are
	// cleaned, so the file stays inside the job.
	traversal := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+target+`","partSize":"1 MiB","file":"../../etc/passwd"}`).(map[string]any)
	name, _ := traversal["file"].(string)
	if name != ".._.._etc_passwd" {
		t.Errorf("file name = %q, want the path separators to be cleaned", name)
	}
	if strings.ContainsAny(name, `/\`) {
		t.Errorf("file name %q still has a path separator", name)
	}
}

// TestDownloadManagerFilePriorityE2E checks the priority of a file: the
// endpoint changes it, the file keeps the one it was added with when the
// value is unknown, and every file of the download has its own priority.
func TestDownloadManagerFilePriorityE2E(t *testing.T) {
	base, _, _ := startDownloadManager(t, "--sample-mib", "1", "--throttle", "0")
	cl := &http.Client{Timeout: 20 * time.Second}
	target := base + "/files/sample.bin"

	job := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+target+`","partSize":"1 MiB","priority":"low"}`).(map[string]any)
	id, _ := job["id"].(string)
	if id == "" {
		t.Fatalf("job without id: %v", job)
	}
	files, _ := job["files"].([]any)
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}
	first, _ := files[0].(map[string]any)
	if first["priority"] != "low" {
		t.Errorf("the priority of the file = %v, want low", first["priority"])
	}
	fileID, _ := first["id"].(string)
	if fileID == "" {
		t.Fatalf("the file has no id: %v", first)
	}

	// The priority of the file is changed with its own endpoint.
	changed := postDownloadManager(t, cl, base+"/api/downloads/"+id+"/files/"+fileID+"/priority",
		`{"priority":"high"}`).(map[string]any)
	changedFiles, _ := changed["files"].([]any)
	cf, _ := changedFiles[0].(map[string]any)
	if cf["priority"] != "high" {
		t.Errorf("the priority of the file = %v, want high", cf["priority"])
	}

	// A priority the backend does not know becomes the normal one.
	unknown := postDownloadManager(t, cl, base+"/api/downloads/"+id+"/files/"+fileID+"/priority",
		`{"priority":"urgent"}`).(map[string]any)
	unknownFiles, _ := unknown["files"].([]any)
	uf, _ := unknownFiles[0].(map[string]any)
	if uf["priority"] != "normal" {
		t.Errorf("the priority of the file = %v, want normal", uf["priority"])
	}

	// A file that is not part of the download is refused.
	resp, err := cl.Post(base+"/api/downloads/"+id+"/files/file-99/priority", "application/json", strings.NewReader(`{"priority":"high"}`))
	if err != nil {
		t.Fatalf("post the priority: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for a file that is not in the download", resp.StatusCode)
	}
}

// TestDownloadManagerMergedFileNameE2E checks that the reassembled file is
// served with the name the operator chose.
func TestDownloadManagerMergedFileNameE2E(t *testing.T) {
	base, _, _ := startDownloadManager(t, "--sample-mib", "1", "--throttle", "0")
	cl := &http.Client{Timeout: 20 * time.Second}
	job := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+base+`/files/sample.bin","partSize":"1 MiB","file":"relazione finale.pdf"}`).(map[string]any)
	id, _ := job["id"].(string)
	if id == "" {
		t.Fatalf("job without id: %v", job)
	}
	done := waitDownloadManager(t, cl, base, id)
	if done["state"] != "done" {
		t.Fatalf("final state = %v (%v)", done["state"], done["error"])
	}
	resp, err := cl.Get(base + "/api/downloads/" + id + "/file")
	if err != nil {
		t.Fatalf("get the merged file: %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Content-Disposition"); got != `attachment; filename="relazione_finale.pdf"` {
		t.Errorf("Content-Disposition = %q, want the name the operator chose", got)
	}
}

// TestDownloadManagerServerE2E runs the Go backend of the example and drives
// a real download from a URL: the file is split into parts, every part is
// fetched with its own Range request, a pause keeps the bytes, a resume
// continues from them and the reassembled file matches the original.
func TestDownloadManagerServerE2E(t *testing.T) {
	base, filesDir, _ := startDownloadManager(t, "--sample-mib", "1", "--throttle", "131072")
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

// TestDownloadManagerSharedQueueE2E checks the queue the files of a download
// share: the workers of the job are split between the files, the file with
// the highest priority goes first, and the part files are deleted when the
// file is reassembled.
func TestDownloadManagerSharedQueueE2E(t *testing.T) {
	origin, reqs := gatedOrigin(t)
	defer origin.Close()

	base, _, workDir := startDownloadManager(t, "--sample-mib", "0", "--throttle", "0", "--workers", "2")
	cl := &http.Client{Timeout: 30 * time.Second}

	// The first file: four parts, two of them are picked right away.
	job := postDownloadManager(t, cl, base+"/api/downloads",
		`{"url":"`+origin.URL+`/a.bin","partSize":"512 KiB"}`).(map[string]any)
	id, _ := job["id"].(string)
	if id == "" {
		t.Fatalf("job without id: %v", job)
	}
	first := waitGatedRequest(t, reqs)
	if first.file != "a.bin" {
		t.Fatalf("the first request is for %q, want a.bin", first.file)
	}
	second := waitGatedRequest(t, reqs)
	if second.file != "a.bin" {
		t.Fatalf("the second request is for %q, want a.bin", second.file)
	}

	// The second file joins the download with a high priority. The workers
	// are busy with the first file, which still has parts in the queue: the
	// next free worker must take a part of the file with the high priority.
	added := postDownloadManager(t, cl, base+"/api/downloads/"+id+"/files",
		`{"url":"`+origin.URL+`/b.bin","priority":"high"}`).(map[string]any)
	files, _ := added["files"].([]any)
	if len(files) != 2 {
		t.Fatalf("the download has %d files, want 2", len(files))
	}
	bf, _ := files[1].(map[string]any)
	if bf["priority"] != "high" {
		t.Errorf("the priority of the second file = %v, want high", bf["priority"])
	}
	if bf["file"] != "b.bin" {
		t.Errorf("the name of the second file = %v, want b.bin", bf["file"])
	}

	// The workers are held by the two requests of the first file: nothing
	// else is asked for until one of them is released.
	select {
	case extra := <-reqs:
		t.Fatalf("a request was made for %q while the workers were busy", extra.file)
	case <-time.After(300 * time.Millisecond):
	}
	first.release <- struct{}{}
	next := waitGatedRequest(t, reqs)
	if next.file != "b.bin" {
		t.Errorf("the next request is for %q, want b.bin: the high priority file goes first", next.file)
	}

	// The pool of the job is shared: the workers never exceed the number the
	// job was given.
	if origin.maxConcurrent() > 2 {
		t.Errorf("%d requests at the same time, want at most 2 (the workers of the job)", origin.maxConcurrent())
	}

	// Let everything finish.
	origin.releaseAll()
	done := waitDownloadManager(t, cl, base, id)
	if done["state"] != "done" {
		t.Fatalf("final state = %v (%v)", done["state"], done["error"])
	}
	final, _ := done["files"].([]any)
	if len(final) != 2 {
		t.Fatalf("the download has %d files, want 2", len(final))
	}
	for _, f := range final {
		fm := f.(map[string]any)
		if fm["state"] != "done" {
			t.Errorf("the file %v is %v, want done", fm["file"], fm["state"])
		}
		if fm["merged"] == "" {
			t.Errorf("the file %v has no merged file", fm["file"])
		}
		// Every file of the download is served on its own endpoint, with the
		// name it was given.
		fileID, _ := fm["id"].(string)
		resp, err := cl.Get(base + "/api/downloads/" + id + "/files/" + fileID + "/file")
		if err != nil {
			t.Fatalf("get the merged file of %v: %v", fm["file"], err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("the merged file of %v answers with %d, want 200", fm["file"], resp.StatusCode)
		}
		if float64(len(body)) != fm["total"].(float64) {
			t.Errorf("the merged file of %v is %d bytes, want %v", fm["file"], len(body), fm["total"])
		}
		if got := resp.Header.Get("Content-Disposition"); !strings.Contains(got, fm["file"].(string)) {
			t.Errorf("Content-Disposition = %q, want the name of the file", got)
		}
	}

	// The part files are deleted when the file is reassembled: the merged
	// file is the result of the download.
	parts := 0
	entries, err := os.ReadDir(filepath.Join(workDir, id))
	if err != nil {
		t.Fatalf("read the job folder: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "part-") {
			parts++
		}
	}
	if parts != 0 {
		t.Errorf("%d part files are still on disk, want 0: the merge deletes them", parts)
	}
}

// gatedRequest is a request the origin holds until the test lets it go.
type gatedRequest struct {
	file    string
	release chan struct{}
}

// gatedOrigin serves two files with Range support and holds every request:
// the order of the requests shows the order the scheduler picked the parts in.
func gatedOrigin(t *testing.T) (*gatedServer, chan gatedRequest) {
	t.Helper()
	sizes := map[string]int{"/a.bin": 2 * 1024 * 1024, "/b.bin": 1024 * 1024}
	reqs := make(chan gatedRequest, 64)
	g := &gatedServer{sizes: sizes}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size := g.sizes[r.URL.Path]
		if size == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		// The probe of the backend asks for the first byte only: it is not a
		// part, so it is not held and it does not count as a worker.
		probe := r.Header.Get("Range") == "bytes=0-0"
		if !probe && !g.isReleased() {
			rel := make(chan struct{})
			g.enter()
			g.hold(rel)
			reqs <- gatedRequest{file: filepath.Base(r.URL.Path), release: rel}
			select {
			case <-rel:
			case <-r.Context().Done():
				g.leave()
				return
			}
			g.leave()
		}
		start, end := 0, size-1
		if spec := strings.TrimPrefix(r.Header.Get("Range"), "bytes="); spec != r.Header.Get("Range") {
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
		if end >= size {
			end = size - 1
		}
		payload := make([]byte, end-start+1)
		for i := range payload {
			payload[i] = byte(i % 251)
		}
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(payload)
	}))
	g.srv = srv
	g.URL = srv.URL
	return g, reqs
}

// gatedServer counts the requests that are running at the same time.
type gatedServer struct {
	srv      *httptest.Server
	URL      string
	sizes    map[string]int
	mu       sync.Mutex
	live     int
	peak     int
	held     []chan struct{}
	released bool
}

func (g *gatedServer) enter() {
	g.mu.Lock()
	g.live++
	if g.live > g.peak {
		g.peak = g.live
	}
	g.mu.Unlock()
}

func (g *gatedServer) leave() {
	g.mu.Lock()
	g.live--
	g.mu.Unlock()
}

func (g *gatedServer) maxConcurrent() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.peak
}

// hold keeps the channel of a held request so that releaseAll can let it go.
func (g *gatedServer) hold(rel chan struct{}) {
	g.mu.Lock()
	g.held = append(g.held, rel)
	g.mu.Unlock()
}

// isReleased says whether the test already let the requests go: from that
// moment on the origin answers without holding anything.
func (g *gatedServer) isReleased() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.released
}

// releaseAll lets the held requests go and opens the gate: the requests that
// come later are not held any more.
func (g *gatedServer) releaseAll() {
	g.mu.Lock()
	held := g.held
	g.held = nil
	g.released = true
	g.mu.Unlock()
	for _, r := range held {
		close(r)
	}
}

func (g *gatedServer) Close() { g.srv.Close() }

// waitGatedRequest waits for the next request the origin holds.
func waitGatedRequest(t *testing.T, reqs chan gatedRequest) gatedRequest {
	t.Helper()
	select {
	case r := <-reqs:
		return r
	case <-time.After(20 * time.Second):
		t.Fatalf("the origin did not receive a request")
		return gatedRequest{}
	}
}

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

	base, _, _ := startDownloadManager(t, "--sample-mib", "0", "--throttle", "0")
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
// returns its base URL, the folder with the local files and the folder with
// the downloads.
func startDownloadManager(t *testing.T, extra ...string) (string, string, string) {
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

	return readDownloadManagerLine(t, stdout, &stderr), filesDir, workDir
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
