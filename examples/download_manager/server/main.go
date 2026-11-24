// download_manager — the Go backend of the SimpleUI "Download manager" example.
//
// It serves the compiled SimpleUI app and drives chunked downloads:
//
//	GET  /                            the compiled SimpleUI app (assets from --web-dir)
//	GET  /files/{name}                a local file, for offline demos
//	GET  /api/config                  the CLI configuration (default URL, part size, ...)
//	GET  /api/downloads               the jobs, most recent first
//	POST /api/downloads               start a job: {"url": "...", "partSize": "1 MiB"}
//	POST /api/downloads/{id}/{act}    pause | resume | cancel
//	POST /api/downloads/{id}/parts/{n}/retry    retry one failed part
//	POST /api/downloads/{id}/retry-failed       retry every failed part
//	GET  /api/downloads/{id}/file     the reassembled file
//
// The point of the example is the download manager itself: the file is split
// into parts according to its size, every part is fetched with its own Range
// request, the parts are tracked one by one, a pause keeps the bytes already
// downloaded, a failed part can be retried alone and the parts are finally
// concatenated and checked against the size the server announced.
//
// The CLI takes the URL to download and the other configuration parameters:
//
//	dm.exe --url https://example.com/file.bin --part-size 1 MiB --workers 4
//
// Run it after building the app:
//
//	simpleui build app.sui --target web --out public
//	cd server && go run . --port 8090 --web-dir ../public --url https://example.com/file.bin

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Part is one slice of the file. The JSON keys are the column headers the
// SUI table shows, so they are human readable on purpose.
type Part struct {
	Index    int    `json:"Part"`
	Start    int64  `json:"From"`
	End      int64  `json:"To"`
	Size     int64  `json:"Size"`
	Received int64  `json:"Received"`
	Percent  int    `json:"Percent"`
	State    string `json:"State"`
	Error    string `json:"Error,omitempty"`
	Attempts int    `json:"Attempts"`
}

// Job is one download: the target URL, how it was split and how far it got.
type Job struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	File     string `json:"file"`
	State    string `json:"state"`
	PartSize int64  `json:"partSize"`
	Total    int64  `json:"total"`
	Received int64  `json:"received"`
	Percent  int    `json:"percent"`
	Speed    int64  `json:"speed"`
	SpeedTx  string `json:"speedText"`
	ETA      int64  `json:"eta"`
	ETATx    string `json:"etaText"`
	PartCnt  int    `json:"partCount"`
	PartTx   string `json:"partSizeText"`
	Range    bool   `json:"rangeSupported"`
	Merged   string `json:"merged"`
	MergedSz int64  `json:"mergedSize"`
	SHA      string `json:"sha"`
	Verify   string `json:"verified"`
	Error    string `json:"error"`
	Created  string `json:"created"`
	Single   bool   `json:"single"`

	parts []*Part
	dir   string

	clientValue *http.Client

	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	workers   int
	active    int
	throttle  int64
	merging   bool
	expectSHA string
}

// Manager keeps the jobs of the process.
type Manager struct {
	mu       sync.Mutex
	jobs     map[string]*Job
	order    []string
	seq      int
	outDir   string
	filesDir string
	workers  int
	throttle int64
	maxParts int
	expect   string
	hosts    []string
	client   *http.Client
}

func main() {
	opts := options{
		port:      8090,
		partSize:  "1 MiB",
		workers:   4,
		maxParts:  256,
		filesDir:  filepath.Join("..", "files"),
		outDir:    filepath.Join("..", "downloads"),
		sampleMiB: 24,
	}
	opts.parse(os.Args[1:])

	if opts.filesDir != "" {
		if err := os.MkdirAll(opts.filesDir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "files dir:", err)
			os.Exit(1)
		}
		sample := filepath.Join(opts.filesDir, "sample.bin")
		if _, err := os.Stat(sample); os.IsNotExist(err) && opts.sampleMiB > 0 {
			if err := generateSample(sample, int64(opts.sampleMiB)); err != nil {
				fmt.Fprintln(os.Stderr, "sample file:", err)
				os.Exit(1)
			}
			fmt.Printf("server: created the sample file %s (%d MiB)\n", sample, opts.sampleMiB)
		}
	}
	if err := os.MkdirAll(opts.outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "out dir:", err)
		os.Exit(1)
	}

	// Port 0 asks the OS for a free port: the base URL is then built from
	// the address the listener actually got.
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
	baseURL := "http://" + ln.Addr().String()

	mgr := &Manager{
		jobs:     map[string]*Job{},
		outDir:   opts.outDir,
		filesDir: opts.filesDir,
		workers:  opts.workers,
		throttle: opts.throttle,
		maxParts: opts.maxParts,
		expect:   strings.ToLower(strings.TrimSpace(opts.expectSHA)),
		hosts:    opts.allowHosts,
		client:   &http.Client{Timeout: 120 * time.Second},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"url":        opts.url,
			"partSize":   opts.partSize,
			"workers":    opts.workers,
			"maxParts":   opts.maxParts,
			"throttle":   opts.throttle,
			"outDir":     absPath(opts.outDir),
			"allowHosts": opts.allowHosts,
			"expectSHA":  opts.expectSHA,
		})
	})
	mux.HandleFunc("GET /api/downloads", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"jobs": mgr.snapshot()})
	})
	mux.HandleFunc("POST /api/downloads", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleCreate(w, r)
	})
	mux.HandleFunc("POST /api/downloads/{id}/{action}", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleAction(w, r)
	})
	mux.HandleFunc("POST /api/downloads/{id}/parts/{index}/retry", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleRetryPart(w, r)
	})
	mux.HandleFunc("POST /api/downloads/{id}/retry-failed", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleRetryFailed(w, r)
	})
	mux.HandleFunc("GET /api/downloads/{id}/file", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleMergedFile(w, r)
	})
	// The sound the app plays when a download is complete.
	mux.HandleFunc("GET /sound.wav", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Content-Length", strconv.Itoa(len(soundBytes)))
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(soundBytes)
	})
	if opts.filesDir != "" {
		mux.HandleFunc("GET /files/{name}", func(w http.ResponseWriter, r *http.Request) {
			mgr.handleFile(w, r)
		})
	}
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		file := filepath.Join(opts.webDir, filepath.FromSlash(strings.TrimPrefix(path, "/")))
		data, err := os.ReadFile(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ct := "application/octet-stream"
		switch {
		case strings.HasSuffix(path, ".html"):
			ct = "text/html; charset=utf-8"
		case strings.HasSuffix(path, ".css"):
			ct = "text/css; charset=utf-8"
		case strings.HasSuffix(path, ".js"):
			ct = "application/javascript; charset=utf-8"
		}
		w.Header().Set("Content-Type", ct)
		_, _ = w.Write(data)
	})

	srv := &http.Server{Handler: cors(logRequests(mux))}
	go func() { _ = srv.Serve(ln) }()

	fmt.Println("======================================================")
	fmt.Printf("download manager + app running at %v/\n", baseURL)
	if opts.url != "" {
		fmt.Printf("  target url   : %v\n", opts.url)
	} else {
		fmt.Println("  target url   : (none, type one in the app)")
	}
	fmt.Printf("  part size    : %v\n", opts.partSize)
	fmt.Printf("  workers      : %v\n", opts.workers)
	fmt.Printf("  out dir      : %v\n", absPath(opts.outDir))
	if len(opts.allowHosts) > 0 {
		fmt.Printf("  allowed hosts: %v\n", strings.Join(opts.allowHosts, ", "))
	}
	if opts.filesDir != "" {
		fmt.Printf("  local files  : %v (try http://%v/files/sample.bin)\n", absPath(opts.filesDir), ln.Addr().String())
	}
	fmt.Println("  the app calls /api/downloads, /api/config and /api/downloads/{id}/...")
	fmt.Println("======================================================")
	if !opts.noOpen {
		openBrowser(baseURL + "/")
	}
	select {}
}

// ------------------------------------------------------------------ flags ---

type options struct {
	url        string
	partSize   string
	workers    int
	throttle   int64
	maxParts   int
	outDir     string
	webDir     string
	filesDir   string
	expectSHA  string
	allowHosts []string
	port       int
	sampleMiB  int
	noOpen     bool
}

func (o *options) parse(args []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url":
			if i+1 < len(args) {
				o.url = strings.TrimSpace(args[i+1])
			}
			i++
		case "--part-size":
			if i+1 < len(args) {
				o.partSize = args[i+1]
			}
			i++
		case "--workers":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil && v > 0 && v <= 32 {
					o.workers = v
				}
			}
			i++
		case "--throttle":
			if i+1 < len(args) {
				if v, err := strconv.ParseInt(args[i+1], 10, 64); err == nil && v >= 0 {
					o.throttle = v
				}
			}
			i++
		case "--max-parts":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil && v > 0 {
					o.maxParts = v
				}
			}
			i++
		case "--out-dir":
			if i+1 < len(args) {
				o.outDir = args[i+1]
			}
			i++
		case "--web-dir":
			if i+1 < len(args) {
				o.webDir = args[i+1]
			}
			i++
		case "--files-dir":
			if i+1 < len(args) {
				o.filesDir = args[i+1]
			}
			i++
		case "--sample-mib":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil && v >= 0 {
					o.sampleMiB = v
				}
			}
			i++
		case "--expect-sha256":
			if i+1 < len(args) {
				o.expectSHA = args[i+1]
			}
			i++
		case "--allow-host":
			if i+1 < len(args) {
				o.allowHosts = append(o.allowHosts, strings.ToLower(strings.TrimSpace(args[i+1])))
			}
			i++
		case "--port":
			if i+1 < len(args) {
				if p, err := strconv.Atoi(args[i+1]); err == nil && p >= 0 && p < 65536 {
					o.port = p
				}
			}
			i++
		case "--no-open":
			o.noOpen = true
		}
	}
}

// ------------------------------------------------------------------- urls ---

// validateTarget checks the URL the operator asked to download. Only http and
// https are accepted, credentials in the URL are refused and, when the CLI
// lists the allowed hosts, the host must be one of them.
func validateTarget(raw string, allowed []string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("missing url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q: only http and https are allowed", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid url: missing host")
	}
	if u.User != nil {
		return nil, fmt.Errorf("invalid url: credentials are not allowed")
	}
	if len(allowed) > 0 {
		host := strings.ToLower(u.Hostname())
		ok := false
		for _, a := range allowed {
			if a != "" && (a == host || strings.HasSuffix(host, "."+a)) {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("host %q is not in the allowed list", host)
		}
	}
	return u, nil
}

// fileNameFromURL returns a safe file name for the reassembled file.
func fileNameFromURL(u *url.URL) string {
	base := filepath.Base(u.Path)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == "/" {
		base = "download.bin"
	}
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	name := b.String()
	if name == "" || name == "." || name == ".." {
		name = "download.bin"
	}
	return name
}

// probe asks the server for the first byte. A 206 answers with the total
// size and tells that the server honours the Range requests; a 200 means the
// server ignored the range and the file cannot be split.
func probe(ctx context.Context, u *url.URL, client *http.Client) (int64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusPartialContent:
		total := parseContentRangeTotal(resp.Header.Get("Content-Range"))
		if total <= 0 {
			return 0, false, fmt.Errorf("the server did not announce the file size")
		}
		return total, true, nil
	case http.StatusOK:
		if resp.ContentLength > 0 {
			return resp.ContentLength, false, nil
		}
		return 0, false, nil
	default:
		return 0, false, fmt.Errorf("unexpected HTTP %d from %v", resp.StatusCode, u.Host)
	}
}

// parseContentRangeTotal reads the total size from "bytes 0-0/12345".
func parseContentRangeTotal(header string) int64 {
	idx := strings.LastIndex(header, "/")
	if idx < 0 {
		return 0
	}
	total := strings.TrimSpace(header[idx+1:])
	if total == "*" {
		return 0
	}
	n, err := strconv.ParseInt(total, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// ------------------------------------------------------------------ files ---

// handleFile serves a local file with Range support: the download manager can
// point at http://127.0.0.1:8090/files/sample.bin to try the example offline.
func (m *Manager) handleFile(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.PathValue("name"))
	path := filepath.Join(m.filesDir, name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")

	start, end, partial := parseRange(r.Header.Get("Range"), info.Size())
	if !partial {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		if r.Method == http.MethodHead {
			return
		}
		_, _ = io.Copy(w, f)
		return
	}
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = f.Seek(start, io.SeekStart)
	_, _ = io.CopyN(w, f, end-start+1)
}

// parseRange understands a single "bytes=start-end" range.
func parseRange(header string, size int64) (int64, int64, bool) {
	if !strings.HasPrefix(header, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(header, "bytes=")
	spec = strings.TrimSpace(spec)
	if strings.Contains(spec, ",") {
		return 0, 0, false
	}
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	start, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false
	}
	end := size - 1
	if s := strings.TrimSpace(parts[1]); s != "" {
		e, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		if e < end {
			end = e
		}
	}
	if end < start {
		return 0, 0, false
	}
	return start, end, true
}

// ------------------------------------------------------------------- jobs ---

type createRequest struct {
	URL      string `json:"url"`
	PartSize string `json:"partSize"`
	// Workers accepts a number or a string: the app sends the value of a
	// number control, that is a string.
	Workers any  `json:"workers"`
	Single  bool `json:"single"`
}

// workersFrom reads the number of parts downloaded at the same time. The
// value can be a JSON number or a string.
func workersFrom(v any, def int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(n)); err == nil && i > 0 {
			return i
		}
	}
	return def
}

func (m *Manager) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing url"})
		return
	}
	u, err := validateTarget(rawURL, m.hosts)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	partSize, err := parseSize(req.PartSize)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	total, rangeOK, err := probe(ctx, u, m.client)
	cancel()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "cannot reach the url: " + err.Error()})
		return
	}

	partCount := 1
	switch {
	case req.Single:
		// One single part: the whole file in one request, no split at all.
		partCount = 1
		if total > 0 {
			partSize = total
		} else {
			// The server does not announce the size: the part is open ended.
			partSize = 0
		}
	case rangeOK && total > 0:
		partCount = int((total + partSize - 1) / partSize)
		if partCount < 1 {
			partCount = 1
		}
		if partCount > m.maxParts {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("too many parts (%d): use a bigger part size (max %d)", partCount, m.maxParts),
			})
			return
		}
	default:
		// The server does not honour the ranges: the file is one single part.
		partSize = total
	}
	if partSize <= 0 && !req.Single {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "the file size is unknown"})
		return
	}

	workers := workersFrom(req.Workers, m.workers)
	if workers > partCount {
		workers = partCount
	}

	m.mu.Lock()
	m.seq++
	id := fmt.Sprintf("job-%d-%d", time.Now().Unix(), m.seq)
	m.mu.Unlock()

	job := &Job{
		ID:       id,
		URL:      u.String(),
		File:     fileNameFromURL(u),
		State:    "running",
		PartSize: partSize,
		Total:    total,
		PartCnt:  partCount,
		PartTx:   humanSize(partSize),
		Range:    rangeOK,
		Created:  time.Now().Format("15:04:05"),
		dir:      filepath.Join(m.outDir, id),
		workers:  workers,
		Single:   req.Single,

		clientValue: m.client,
		throttle:    m.throttle,
		expectSHA:   m.expect,
	}
	for i := 0; i < partCount; i++ {
		start := int64(i) * partSize
		end := start + partSize - 1
		if rangeOK && total > 0 && end > total-1 {
			end = total - 1
		}
		job.parts = append(job.parts, &Part{
			Index: i + 1,
			Start: start,
			End:   end,
			Size:  end - start + 1,
			State: "pending",
		})
	}
	if err := os.MkdirAll(job.dir, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	fmt.Printf("server: new job %s  url=%s size=%s parts=%d x %s workers=%d\n",
		id, job.URL, humanSize(total), partCount, job.PartTx, workers)

	m.mu.Lock()
	m.jobs[id] = job
	m.order = append(m.order, id)
	m.mu.Unlock()

	job.start()
	writeJSON(w, http.StatusOK, job.snapshot())
}

func (m *Manager) handleAction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	action := r.PathValue("action")
	job := m.job(id)
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown download: " + id})
		return
	}
	switch action {
	case "pause":
		job.pause()
		fmt.Printf("server: job %s paused\n", id)
	case "resume":
		job.resume()
		fmt.Printf("server: job %s resumed\n", id)
	case "cancel":
		job.cancelJob()
		fmt.Printf("server: job %s canceled\n", id)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown action: " + action})
		return
	}
	writeJSON(w, http.StatusOK, job.snapshot())
}

// handleRetryPart restarts a single part: the bytes of that part are dropped
// and it is queued again, the other parts are not touched.
func (m *Manager) handleRetryPart(w http.ResponseWriter, r *http.Request) {
	job := m.job(r.PathValue("id"))
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown download: " + r.PathValue("id")})
		return
	}
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil || index < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid part number: " + r.PathValue("index")})
		return
	}
	if err := job.retryPart(index); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	fmt.Printf("server: job %s part %d queued again\n", job.ID, index)
	writeJSON(w, http.StatusOK, job.snapshot())
}

// handleRetryFailed restarts every part that failed.
func (m *Manager) handleRetryFailed(w http.ResponseWriter, r *http.Request) {
	job := m.job(r.PathValue("id"))
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown download: " + r.PathValue("id")})
		return
	}
	n, err := job.retryFailed()
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	fmt.Printf("server: job %s queued %d failed parts again\n", job.ID, n)
	writeJSON(w, http.StatusOK, job.snapshot())
}

func (m *Manager) job(id string) *Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.jobs[id]
}

func (m *Manager) handleMergedFile(w http.ResponseWriter, r *http.Request) {
	job := m.job(r.PathValue("id"))
	if job == nil {
		http.Error(w, "unknown download", http.StatusNotFound)
		return
	}
	job.mu.Lock()
	merged, state := job.Merged, job.State
	job.mu.Unlock()
	if merged == "" || state != "done" {
		http.Error(w, "the download is not complete yet", http.StatusConflict)
		return
	}
	f, err := os.Open(merged)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	info, _ := f.Stat()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(merged)+"\"")
	_, _ = io.Copy(w, f)
}

// snapshot returns a copy of the jobs, most recent first.
func (m *Manager) snapshot() []map[string]any {
	m.mu.Lock()
	ids := append([]string(nil), m.order...)
	m.mu.Unlock()
	out := []map[string]any{}
	for i := len(ids) - 1; i >= 0; i-- {
		job := m.job(ids[i])
		if job == nil {
			continue
		}
		out = append(out, job.snapshot())
	}
	return out
}

// ------------------------------------------------------------- job engine ---

func (j *Job) start() {
	j.mu.Lock()
	if j.cancel != nil {
		// Stop the previous generation of workers, if any.
		j.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	j.ctx, j.cancel = ctx, cancel
	j.active = 0
	j.mu.Unlock()
	for i := 0; i < j.workers; i++ {
		j.wg.Add(1)
		j.mu.Lock()
		j.active++
		j.mu.Unlock()
		go j.worker(ctx)
	}
	go j.watcher(ctx)
}

func (j *Job) worker(ctx context.Context) {
	defer j.wg.Done()
	defer func() {
		j.mu.Lock()
		j.active--
		j.mu.Unlock()
	}()
	for {
		if ctx.Err() != nil {
			return
		}
		p := j.nextPending()
		if p == nil {
			return
		}
		j.downloadPart(ctx, p)
	}
}

// nextPending hands the next free part to a worker.
func (j *Job) nextPending() *Part {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.State != "running" {
		return nil
	}
	for _, p := range j.parts {
		if p.State == "pending" {
			p.State = "downloading"
			return p
		}
	}
	return nil
}

// progressWriter updates the counters while the bytes flow.
type progressWriter struct {
	job  *Job
	part *Part
	w    io.Writer
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	if n > 0 {
		p.job.mu.Lock()
		p.part.Received += int64(n)
		p.job.Received += int64(n)
		p.part.Percent = percent(p.part.Received, p.part.Size)
		p.job.Percent = percent(p.job.Received, p.job.Total)
		throttle := p.job.throttle
		p.job.mu.Unlock()
		if throttle > 0 {
			// Slow the stream down so the progress is visible and a pause
			// lands in the middle of a part.
			time.Sleep(time.Duration(float64(n) / float64(throttle) * float64(time.Second)))
		}
	}
	return n, err
}

// downloadPart fetches one part with its own Range request. The bytes
// already on disk are kept: on resume the request starts where it stopped.
func (j *Job) downloadPart(ctx context.Context, p *Part) {
	j.mu.Lock()
	start := p.Start + p.Received
	state := j.State
	j.mu.Unlock()
	if state != "running" {
		return
	}
	if start > p.End {
		j.setPartState(p, "done", "")
		j.maybeMerge()
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, j.URL, nil)
	if err != nil {
		j.setPartState(p, "error", err.Error())
		j.settle()
		return
	}
	if p.End >= 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, p.End))
	} else {
		// The size of the file is unknown: ask for everything from the byte
		// the part stopped at.
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
	}
	resp, err := j.client().Do(req)
	if err != nil {
		if ctx.Err() != nil {
			j.setPartState(p, "pending", "")
			return
		}
		j.setPartState(p, "error", err.Error())
		j.settle()
		return
	}
	defer resp.Body.Close()
	// A part that covers the whole file also accepts a plain 200 answer: some
	// servers ignore the Range header when the range is the whole file.
	wholeFile := start == 0 && (p.End < 0 || p.End >= j.Total-1)
	if resp.StatusCode != http.StatusPartialContent && !(resp.StatusCode == http.StatusOK && wholeFile) {
		j.setPartState(p, "error", fmt.Sprintf("HTTP %d from the file server", resp.StatusCode))
		j.settle()
		return
	}

	path := j.partPath(p.Index)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		j.setPartState(p, "error", err.Error())
		j.settle()
		return
	}
	defer f.Close()

	w := &progressWriter{job: j, part: p, w: f}
	if _, err := io.Copy(w, resp.Body); err != nil {
		if ctx.Err() != nil {
			// Paused: the bytes already written stay on disk.
			j.setPartState(p, "pending", "")
			return
		}
		j.setPartState(p, "error", err.Error())
		j.settle()
		return
	}

	j.mu.Lock()
	complete := p.Size <= 0 || p.Received >= p.Size
	j.mu.Unlock()
	if !complete {
		j.setPartState(p, "error", fmt.Sprintf("received %d bytes out of %d", p.Received, p.Size))
		j.settle()
		return
	}
	j.setPartState(p, "done", "")
	fmt.Printf("server: job %s part %d/%d ok (%s)\n", j.ID, p.Index, j.PartCnt, humanSize(p.Size))
	j.settle()
	j.maybeMerge()
}

// settle closes the job when no part is running any more and some of them
// failed: the operator can then retry the failed parts one by one.
func (j *Job) settle() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.State != "running" {
		return
	}
	failed := 0
	pending := 0
	for _, p := range j.parts {
		switch p.State {
		case "error":
			failed++
		case "pending", "downloading":
			pending++
		}
	}
	if pending == 0 && failed > 0 {
		j.State = "error"
		j.Error = fmt.Sprintf("%d part(s) failed: retry them from the app", failed)
	}
}

// retryPart queues one part again: its bytes are dropped and it is fetched
// from the start.
func (j *Job) retryPart(index int) error {
	j.mu.Lock()
	if j.State == "canceled" || j.State == "done" {
		state := j.State
		j.mu.Unlock()
		return fmt.Errorf("the download is %s: it cannot be retried", state)
	}
	var part *Part
	for _, p := range j.parts {
		if p.Index == index {
			part = p
			break
		}
	}
	if part == nil {
		j.mu.Unlock()
		return fmt.Errorf("unknown part %d", index)
	}
	part.State = "pending"
	part.Received = 0
	part.Percent = 0
	part.Error = ""
	part.Attempts++
	j.mu.Unlock()

	// The bytes of the failed part are dropped: the part is fetched again
	// from its first byte.
	if err := os.Remove(j.partPath(index)); err != nil && !os.IsNotExist(err) {
		return err
	}
	j.mu.Lock()
	j.State = "running"
	j.Error = ""
	needsWorkers := j.active == 0
	j.mu.Unlock()

	if needsWorkers {
		// The job had stopped on a failed part: the workers are gone, so
		// they are started again for the retried part.
		j.start()
	}
	return nil
}

// retryFailed queues every failed part again.
func (j *Job) retryFailed() (int, error) {
	j.mu.Lock()
	var indexes []int
	for _, p := range j.parts {
		if p.State == "error" {
			indexes = append(indexes, p.Index)
		}
	}
	j.mu.Unlock()
	if len(indexes) == 0 {
		return 0, fmt.Errorf("there is no failed part to retry")
	}
	for _, idx := range indexes {
		if err := j.retryPart(idx); err != nil {
			return 0, err
		}
	}
	return len(indexes), nil
}

// maybeMerge concatenates the parts when they are all complete and checks
// the result against the size the server announced.
func (j *Job) maybeMerge() {
	j.mu.Lock()
	if j.merging || j.State != "running" {
		j.mu.Unlock()
		return
	}
	for _, p := range j.parts {
		if p.State != "done" {
			j.mu.Unlock()
			return
		}
	}
	j.merging = true
	j.State = "merging"
	j.mu.Unlock()

	go func() {
		merged := filepath.Join(j.dir, "merged-"+j.File)
		out, err := os.Create(merged)
		if err != nil {
			j.failMerge(err.Error())
			return
		}
		for _, p := range j.parts {
			in, err := os.Open(j.partPath(p.Index))
			if err != nil {
				out.Close()
				j.failMerge(err.Error())
				return
			}
			if _, err := io.Copy(out, in); err != nil {
				in.Close()
				out.Close()
				j.failMerge(err.Error())
				return
			}
			in.Close()
		}
		if err := out.Close(); err != nil {
			j.failMerge(err.Error())
			return
		}
		size := int64(0)
		if info, err := os.Stat(merged); err == nil {
			size = info.Size()
		}
		sum := fileSHA256(merged)
		verify := "size matches the server"
		if j.Total > 0 && size != j.Total {
			verify = fmt.Sprintf("size mismatch: %d bytes instead of %d", size, j.Total)
		}
		if j.expectSHA != "" {
			if sum == j.expectSHA {
				verify = "SHA-256 matches the expected value"
			} else {
				verify = "SHA-256 does not match the expected value"
			}
		}
		j.mu.Lock()
		j.Merged = merged
		j.MergedSz = size
		j.SHA = sum
		j.Verify = verify
		j.State = "done"
		j.merging = false
		j.mu.Unlock()
		fmt.Printf("server: job %s merged %s (%s) %s\n", j.ID, filepath.Base(merged), humanSize(size), verify)
	}()
}

func (j *Job) failMerge(msg string) {
	j.mu.Lock()
	j.State = "error"
	j.Error = msg
	j.merging = false
	j.mu.Unlock()
	fmt.Printf("server: job %s merge failed: %s\n", j.ID, msg)
}

func (j *Job) setPartState(p *Part, state, msg string) {
	j.mu.Lock()
	p.State = state
	if state == "done" && p.Size <= 0 {
		// The size of the file was unknown: the bar is full when the part is.
		p.Percent = 100
	}
	if msg != "" {
		p.Error = msg
	}
	j.mu.Unlock()
}

func (j *Job) pause() {
	j.mu.Lock()
	if j.State != "running" {
		j.mu.Unlock()
		return
	}
	j.State = "paused"
	cancel := j.cancel
	j.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	j.wg.Wait()
	j.mu.Lock()
	for _, p := range j.parts {
		if p.State == "downloading" {
			p.State = "pending"
		}
	}
	j.mu.Unlock()
}

func (j *Job) resume() {
	j.mu.Lock()
	if j.State != "paused" {
		j.mu.Unlock()
		return
	}
	j.State = "running"
	j.mu.Unlock()
	j.start()
}

func (j *Job) cancelJob() {
	j.mu.Lock()
	if j.State == "done" {
		j.mu.Unlock()
		return
	}
	j.State = "canceled"
	cancel := j.cancel
	j.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	j.wg.Wait()
	j.mu.Lock()
	for _, p := range j.parts {
		if p.State == "downloading" {
			p.State = "pending"
		}
	}
	j.mu.Unlock()
}

// watcher keeps the speed and the ETA up to date.
func (j *Job) watcher(ctx context.Context) {
	t := time.NewTicker(400 * time.Millisecond)
	defer t.Stop()
	j.mu.Lock()
	lastBytes, lastTime := j.Received, time.Now()
	j.mu.Unlock()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			j.mu.Lock()
			cur := j.Received
			dt := now.Sub(lastTime).Seconds()
			if dt > 0 {
				j.Speed = int64(float64(cur-lastBytes) / dt)
			}
			lastBytes, lastTime = cur, now
			if j.Speed > 0 && j.Total > cur {
				j.ETA = int64(float64(j.Total-cur) / float64(j.Speed))
			} else {
				j.ETA = 0
			}
			j.SpeedTx = humanSize(j.Speed) + "/s"
			j.ETATx = humanDuration(j.ETA)
			j.mu.Unlock()
		}
	}
}

// partView is the JSON view of a part. It is a struct, not a map, because
// Go sorts map keys alphabetically: with a struct the table shows the columns
// in the order the operator expects.
type partView struct {
	Part     int    `json:"Part"`
	From     int64  `json:"From"`
	To       int64  `json:"To"`
	Size     int64  `json:"Size"`
	Received int64  `json:"Received"`
	Percent  int    `json:"Percent"`
	State    string `json:"State"`
	Error    string `json:"Error,omitempty"`
	Attempts int    `json:"Attempts"`
}

// snapshot is the JSON view of the job.
func (j *Job) snapshot() map[string]any {
	j.mu.Lock()
	defer j.mu.Unlock()
	parts := []partView{}
	failed := 0
	for _, p := range j.parts {
		if p.State == "error" {
			failed++
		}
		parts = append(parts, partView{
			Part:     p.Index,
			From:     p.Start,
			To:       p.End,
			Size:     p.Size,
			Received: p.Received,
			Percent:  p.Percent,
			State:    p.State,
			Error:    p.Error,
			Attempts: p.Attempts,
		})
	}
	return map[string]any{
		"id":             j.ID,
		"url":            j.URL,
		"file":           j.File,
		"state":          j.State,
		"partSize":       j.PartSize,
		"partSizeText":   j.PartTx,
		"partCount":      j.PartCnt,
		"workers":        j.workers,
		"single":         j.Single,
		"throttle":       j.throttle,
		"total":          j.Total,
		"totalText":      humanSize(j.Total),
		"received":       j.Received,
		"receivedText":   humanSize(j.Received),
		"percent":        j.Percent,
		"speed":          j.Speed,
		"speedText":      j.SpeedTx,
		"eta":            j.ETA,
		"etaText":        j.ETATx,
		"merged":         j.Merged,
		"mergedSize":     j.MergedSz,
		"sha":            j.SHA,
		"verified":       j.Verify,
		"error":          j.Error,
		"created":        j.Created,
		"rangeSupported": j.Range,
		"failedParts":    failed,
		"parts":          parts,
	}
}

// ----------------------------------------------------------------- paths ---

func (j *Job) client() *http.Client { return j.clientValue }

func (j *Job) partPath(index int) string {
	return filepath.Join(j.dir, fmt.Sprintf("part-%04d.bin", index))
}

// ---------------------------------------------------------------- helpers ---

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// cors lets the app call the API even when it is served from another origin
// (for example the desktop build, which serves its own assets).
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		if r.URL.Path == "/api/downloads" || strings.HasPrefix(r.URL.Path, "/files/") {
			fmt.Printf("server: %v %v -> %v\n", r.Method, r.URL.Path, rec.status)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func percent(done, total int64) int {
	if total <= 0 {
		return 0
	}
	p := int(float64(done) / float64(total) * 100)
	if p > 100 {
		p = 100
	}
	return p
}

func humanSize(n int64) string {
	if n <= 0 {
		return "0 B"
	}
	units := []string{"B", "KiB", "MiB", "GiB"}
	v := float64(n)
	i := 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d %s", n, units[i])
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}

func humanDuration(seconds int64) string {
	if seconds <= 0 {
		return "--"
	}
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// parseSize accepts "1 MiB", "512KiB", "2MB" or a plain byte count.
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, fmt.Errorf("missing part size")
	}
	mult := int64(1)
	for _, u := range []struct {
		suffix string
		mult   int64
	}{
		{"gib", 1 << 30}, {"mib", 1 << 20}, {"kib", 1 << 10},
		{"gb", 1 << 30}, {"mb", 1 << 20}, {"kb", 1 << 10}, {"b", 1},
	} {
		if strings.HasSuffix(s, u.suffix) {
			mult = u.mult
			s = strings.TrimSpace(strings.TrimSuffix(s, u.suffix))
			break
		}
	}
	if s == "" {
		return 0, fmt.Errorf("invalid part size")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid part size: %q", s)
	}
	if n <= 0 {
		return 0, fmt.Errorf("the part size must be positive")
	}
	return n * mult, nil
}

func fileSHA256(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// generateSample writes a deterministic file so the example works offline
// and always produces the same checksum.
func generateSample(path string, mib int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	const chunk = 1 << 20
	buf := make([]byte, chunk)
	x := uint64(0x9E3779B97F4A7C15)
	for i := int64(0); i < mib; i++ {
		for j := 0; j < len(buf); j += 8 {
			x ^= x << 13
			x ^= x >> 7
			x ^= x << 17
			binaryPut(buf[j:], x)
		}
		if _, err := f.Write(buf); err != nil {
			return err
		}
	}
	return nil
}

func binaryPut(b []byte, v uint64) {
	for i := 0; i < 8 && i < len(b); i++ {
		b[i] = byte(v >> (8 * i))
	}
}

// soundBytes is the chime the app plays when a download is complete: it is
// built once at startup and served as a WAV file.
var soundBytes = buildChime()

// buildChime renders a short two-tone chime as a 16 bit PCM WAV file: the
// app plays it with the sound module when the download is complete.
func buildChime() []byte {
	const rate = 44100
	notes := []struct {
		freq   float64
		length float64
	}{
		{783.99, 0.17},  // G5
		{1046.50, 0.33}, // C6
	}
	total := 0.0
	for _, n := range notes {
		total += n.length
	}
	samples := make([]int16, int(total*rate))
	pos := 0
	for _, n := range notes {
		count := int(n.length * rate)
		for i := 0; i < count && pos < len(samples); i++ {
			env := math.Exp(-4.2 * float64(i) / float64(count))
			v := math.Sin(2*math.Pi*n.freq*float64(i)/rate) * env * 0.32
			samples[pos] = int16(v * 32767)
			pos++
		}
	}
	data := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(s))
	}

	var buf bytes.Buffer
	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(36+len(data)))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // mono
	_ = binary.Write(&buf, binary.LittleEndian, uint32(rate))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(rate*2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

func absPath(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("cmd", "/c", "start", "", url).Run()
	case "darwin":
		_ = exec.Command("open", url).Run()
	default:
		_ = exec.Command("xdg-open", url).Run()
	}
}
