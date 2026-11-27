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
//	POST /api/downloads/{id}/files/{fid}/{act}    pause | resume | stop | restart | cancel
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

	// file is the file the part belongs to: the parts of the files of a job
	// share the queue of the job.
	file *File
}

// The priorities of a file. The lower the rank, the sooner the workers pick
// the parts of the file.
const (
	prioHigh   = "high"
	prioNormal = "normal"
	prioLow    = "low"
)

// priorityRank turns a priority into its rank: high first, then normal, then
// low.
func priorityRank(p string) int {
	switch p {
	case prioHigh:
		return 0
	case prioLow:
		return 2
	default:
		return 1
	}
}

// normalizePriority accepts what the app sends and falls back to the normal
// priority for anything unknown.
func normalizePriority(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case prioHigh:
		return prioHigh
	case prioLow:
		return prioLow
	default:
		return prioNormal
	}
}

// File is one file of a download: its own URL, name, size, priority and
// parts. The files of a job share the queue of the parts and the workers of
// the job: the priority of a file decides how soon its parts are picked.
type File struct {
	ID       string
	URL      string
	Name     string
	Priority string
	Total    int64
	Received int64
	Percent  int
	PartSize int64
	PartCnt  int
	PartTx   string
	Range    bool
	State    string
	Error    string
	Merged   string
	MergedSz int64
	SHA      string
	Verify   string

	parts []*Part
	// pending counts the parts of the file that are still waiting: the
	// scheduler uses it to know if the file has work left.
	pending int
	// paused is true when the operator paused this file alone: its parts
	// stay in the queue but the scheduler does not pick them.
	paused bool
}

// Job is one download: the files to fetch, the workers that fetch their
// parts and how far they got.
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

	// files are the files of the download, in the order they were added.
	files []*File
	// parts are all the parts of all the files, in the same order.
	parts []*Part
	// rr is the cursor of the scheduler: the files at the same priority take
	// turns, so the workers are shared fairly between them.
	rr int
	// seq numbers the files of the job.
	seq int

	dir string

	clientValue *http.Client

	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	workers   int
	active    int
	throttle  int64
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
	// Another file in a download that is already running: the files share the
	// queue of the parts.
	mux.HandleFunc("POST /api/downloads/{id}/files", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleAddFile(w, r)
	})
	// The priority of a file: the parts of the files with the highest
	// priority are picked first.
	mux.HandleFunc("POST /api/downloads/{id}/files/{file}/priority", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleSetPriority(w, r)
	})
	// One file of a download: pause, resume, stop, restart or cancel. The
	// actions of a file do not touch the other files of the download.
	mux.HandleFunc("POST /api/downloads/{id}/files/{file}/{action}", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleFileAction(w, r)
	})
	mux.HandleFunc("GET /api/downloads/{id}/file", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleMergedFile(w, r)
	})
	mux.HandleFunc("GET /api/downloads/{id}/files/{file}/file", func(w http.ResponseWriter, r *http.Request) {
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
	return cleanFileName(base)
}

// cleanFileName keeps only the characters a plain file name can have: letters,
// digits, dot, dash and underscore. Everything else (spaces, path separators,
// quotes) becomes an underscore, so a name can never point outside the folder
// of the job.
func cleanFileName(base string) string {
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
	// A name that is too long would not fit in the path of the job.
	if len(name) > 120 {
		name = name[:120]
	}
	return name
}

// chosenFileName returns the name the operator chose in the GUI. The name is
// cleaned like the one taken from the URL, so it is always a plain file name
// inside the folder of the job.
func chosenFileName(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		return ""
	}
	return cleanFileName(name)
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
	// File is the name the operator chose for the file to save. It is
	// optional: when it is empty the name comes from the URL.
	File string `json:"file"`
	// Priority decides how soon the parts of the file are picked: high,
	// normal or low. The files of a job share the queue of the parts.
	Priority string `json:"priority"`
}

// fileRequest adds a file to a download that is already running.
type fileRequest struct {
	URL      string `json:"url"`
	File     string `json:"file"`
	Priority string `json:"priority"`
}

// priorityRequest changes the priority of a file of a download.
type priorityRequest struct {
	Priority string `json:"priority"`
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

// planFile probes the URL and works out how the file is split: it returns the
// file with its parts ready to be queued. The part size is in bytes.
func (m *Manager) planFile(job *Job, rawURL, chosenName string, partSize int64, single bool, priority string) (*File, error) {
	u, err := validateTarget(rawURL, m.hosts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	total, rangeOK, err := probe(ctx, u, m.client)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("cannot reach the url: %w", err)
	}

	partCount := 1
	switch {
	case single:
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
			return nil, fmt.Errorf("too many parts (%d): use a bigger part size (max %d)", partCount, m.maxParts)
		}
	default:
		// The server does not honour the ranges: the file is one single part.
		partSize = total
	}
	if partSize <= 0 && !single {
		return nil, fmt.Errorf("the file size is unknown")
	}

	// The name of the file to save: the one the operator chose in the GUI or,
	// when there is none, the one that comes from the URL.
	name := fileNameFromURL(u)
	if chosen := chosenFileName(chosenName); chosen != "" {
		name = chosen
	}

	f := &File{
		URL:      u.String(),
		Name:     name,
		Priority: normalizePriority(priority),
		Total:    total,
		PartSize: partSize,
		PartCnt:  partCount,
		PartTx:   humanSize(partSize),
		Range:    rangeOK,
		State:    "pending",
	}
	for i := 0; i < partCount; i++ {
		start := int64(i) * partSize
		end := start + partSize - 1
		if rangeOK && total > 0 && end > total-1 {
			end = total - 1
		}
		p := &Part{
			Start: start,
			End:   end,
			Size:  end - start + 1,
			State: "pending",
			file:  f,
		}
		f.parts = append(f.parts, p)
		f.pending++
	}
	return f, nil
}

// addFile puts a file in the download and updates the totals of the job. The
// parts are numbered inside the whole download: the number is unique, so the
// part files and the retry endpoint stay simple.
func (j *Job) addFile(f *File) {
	j.seq++
	f.ID = fmt.Sprintf("file-%d", j.seq)
	for _, p := range f.parts {
		p.Index = len(j.parts) + 1
		j.parts = append(j.parts, p)
	}
	j.files = append(j.files, f)
	j.Total += f.Total
	j.PartCnt += f.PartCnt
	j.PartSize = f.PartSize
	j.PartTx = f.PartTx
	j.Range = f.Range
	if j.URL == "" {
		j.URL = f.URL
		j.File = f.Name
	}
	j.Percent = percent(j.Received, j.Total)
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

	m.mu.Lock()
	m.seq++
	id := fmt.Sprintf("job-%d-%d", time.Now().Unix(), m.seq)
	m.mu.Unlock()

	job := &Job{
		ID:      id,
		State:   "running",
		Created: time.Now().Format("15:04:05"),
		dir:     filepath.Join(m.outDir, id),
		workers: workersFrom(req.Workers, m.workers),
		Single:  req.Single,

		clientValue: m.client,
		throttle:    m.throttle,
		expectSHA:   m.expect,
	}
	partSize, err := parseSize(req.PartSize)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	f, err := m.planFile(job, rawURL, req.File, partSize, req.Single, req.Priority)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	job.addFile(f)
	if err := os.MkdirAll(job.dir, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	fmt.Printf("server: new job %s  url=%s size=%s parts=%d x %s workers=%d priority=%s\n",
		id, f.URL, humanSize(f.Total), f.PartCnt, f.PartTx, job.workers, f.Priority)

	m.mu.Lock()
	m.jobs[id] = job
	m.order = append(m.order, id)
	m.mu.Unlock()

	job.start()
	writeJSON(w, http.StatusOK, job.snapshot())
}

// handleAddFile puts another file in a download that is already running: the
// file joins the queue of the parts and the workers of the job.
func (m *Manager) handleAddFile(w http.ResponseWriter, r *http.Request) {
	job := m.job(r.PathValue("id"))
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown download"})
		return
	}
	if job.State == "canceled" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "the download is canceled: it cannot take new files"})
		return
	}
	var req fileRequest
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

	job.mu.Lock()
	partSize := job.PartSize
	single := job.Single
	job.mu.Unlock()

	f, err := m.planFile(job, rawURL, req.File, partSize, single, req.Priority)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	job.mu.Lock()
	job.addFile(f)
	job.State = "running"
	job.Error = ""
	job.mu.Unlock()
	fmt.Printf("server: job %s takes the file %s (%s, priority %s)\n", job.ID, f.Name, f.PartTx, f.Priority)

	// The workers of the job may have finished their queue: they start again
	// to pick the parts of the new file.
	job.ensureWorkers()
	writeJSON(w, http.StatusOK, job.snapshot())
}

// handleSetPriority changes the priority of a file of a download: the
// scheduler picks the parts of the files with the highest priority first.
func (m *Manager) handleSetPriority(w http.ResponseWriter, r *http.Request) {
	job := m.job(r.PathValue("id"))
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown download"})
		return
	}
	var req priorityRequest
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}
	priority := normalizePriority(req.Priority)
	job.mu.Lock()
	found := false
	for _, f := range job.files {
		if f.ID == r.PathValue("file") {
			f.Priority = priority
			found = true
			break
		}
	}
	job.mu.Unlock()
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown file in the download"})
		return
	}
	fmt.Printf("server: job %s file %s priority %s\n", job.ID, r.PathValue("file"), priority)
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

// handleFileAction acts on one file of a download: pause, resume, stop,
// restart or cancel. The other files of the download keep going.
//
//   - pause: the file stops, the bytes already on disk are kept;
//   - resume: the file goes back into the queue;
//   - stop: the file stops and its URL can be changed again;
//   - restart: the file is downloaded again from its first byte, with the URL
//     of the request when the operator changed it;
//   - cancel: the file leaves the download and its bytes are deleted.
func (m *Manager) handleFileAction(w http.ResponseWriter, r *http.Request) {
	job := m.job(r.PathValue("id"))
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown download: " + r.PathValue("id")})
		return
	}
	action := r.PathValue("action")
	f := job.fileByID(r.PathValue("file"))
	if f == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown file in the download"})
		return
	}

	var req fileRequest
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if len(strings.TrimSpace(string(body))) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
			return
		}
	}

	switch action {
	case "pause":
		job.pauseFile(f)
	case "resume":
		job.resumeFile(f)
	case "stop":
		job.cancelFile(f)
	case "restart":
		rawURL := strings.TrimSpace(req.URL)
		if rawURL != "" && rawURL != f.URL {
			// The operator changed the URL of a file that had stopped: the
			// file is planned again and downloaded from the new URL.
			job.mu.Lock()
			partSize := job.PartSize
			single := job.Single
			priority := f.Priority
			job.mu.Unlock()
			fresh, err := m.planFile(job, rawURL, req.File, partSize, single, priority)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			fresh.ID = f.ID
			job.mu.Lock()
			oldParts := make([]*Part, 0, len(f.parts))
			for _, p := range f.parts {
				oldParts = append(oldParts, p)
			}
			oldMerged := f.Merged
			job.mu.Unlock()
			for _, p := range oldParts {
				os.Remove(job.partPath(p.Index))
			}
			if oldMerged != "" {
				os.Remove(oldMerged)
			}
			job.replaceFile(f, fresh)
			job.mu.Lock()
			job.State = "running"
			job.Error = ""
			job.mu.Unlock()
			job.ensureWorkers()
		} else if err := job.restartFile(f); err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
	case "cancel":
		if err := job.removeFile(f); err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown action: " + action})
		return
	}
	fmt.Printf("server: job %s file %s %s\n", job.ID, r.PathValue("file"), action)
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

// handleMergedFile serves the reassembled file of a download. Without a file
// id it serves the first file that is complete.
func (m *Manager) handleMergedFile(w http.ResponseWriter, r *http.Request) {
	job := m.job(r.PathValue("id"))
	if job == nil {
		http.Error(w, "unknown download", http.StatusNotFound)
		return
	}
	want := r.PathValue("file")
	job.mu.Lock()
	var merged, state, name string
	for _, f := range job.files {
		if want != "" && f.ID != want {
			continue
		}
		if f.Merged == "" {
			continue
		}
		merged, state, name = f.Merged, f.State, f.Name
		break
	}
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
	// The file is downloaded with the name the operator chose: on disk the
	// merged file keeps the "merged-" prefix, so it cannot clash with a part.
	if name == "" {
		name = filepath.Base(merged)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
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

// ensureWorkers starts the workers of the job when they are not running: a
// file added to a job whose queue was empty needs them again.
func (j *Job) ensureWorkers() {
	j.mu.Lock()
	if j.State != "running" || j.active > 0 {
		j.mu.Unlock()
		return
	}
	pending := 0
	for _, f := range j.files {
		pending += f.pending
	}
	j.mu.Unlock()
	if pending == 0 {
		return
	}
	j.start()
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

// nextPending hands the next free part to a worker. The parts of the files
// with the highest priority go first; inside a priority the files take turns,
// so the workers are shared fairly between them: with six workers and two
// files of the same priority three parts of each file are downloaded at the
// same time.
func (j *Job) nextPending() *Part {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.State != "running" {
		return nil
	}
	best := ""
	for _, f := range j.files {
		if f.pending == 0 || f.paused || f.State == "canceled" {
			continue
		}
		if best == "" || priorityRank(f.Priority) < priorityRank(best) {
			best = f.Priority
		}
	}
	if best == "" {
		return nil
	}
	for i := 0; i < len(j.files); i++ {
		idx := (j.rr + i) % len(j.files)
		f := j.files[idx]
		if f.Priority != best || f.pending == 0 || f.paused || f.State == "canceled" {
			continue
		}
		for _, p := range f.parts {
			if p.State == "pending" {
				p.State = "downloading"
				f.pending--
				if f.State == "pending" {
					f.State = "running"
				}
				// The next worker starts with the following file: the files of
				// the same priority take turns.
				j.rr = (idx + 1) % len(j.files)
				return p
			}
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
		if p.part.file != nil {
			p.part.file.Received += int64(n)
			p.part.file.Percent = percent(p.part.file.Received, p.part.file.Total)
		}
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
	url := j.URL
	total := j.Total
	if p.file != nil {
		url = p.file.URL
		total = p.file.Total
	}
	j.mu.Unlock()
	if state != "running" {
		return
	}
	if start > p.End {
		j.setPartState(p, "done", "")
		j.maybeMerge()
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	wholeFile := start == 0 && (p.End < 0 || p.End >= total-1)
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

	w := &progressWriter{job: j, part: p, w: f}
	_, copyErr := io.Copy(w, resp.Body)
	// The bytes are on disk: the handle is closed, so the file of the part can
	// be deleted right away when its file has been canceled.
	f.Close()
	if copyErr != nil {
		if ctx.Err() != nil {
			// Paused: the bytes already written stay on disk.
			j.setPartState(p, "pending", "")
			return
		}
		j.setPartState(p, "error", copyErr.Error())
		j.settle()
		return
	}

	j.mu.Lock()
	complete := p.Size <= 0 || p.Received >= p.Size
	// The file may have been paused or canceled while the part was running:
	// the part goes back to the queue of a paused file and is dropped from a
	// canceled one.
	dropped := false
	if p.file != nil {
		if p.file.State == "canceled" {
			dropped = true
		} else if p.file.paused && !complete {
			j.mu.Unlock()
			j.setPartState(p, "pending", "")
			return
		}
	}
	j.mu.Unlock()
	if dropped {
		j.setPartState(p, "canceled", "")
		// The file of the part was canceled while it was downloading: its
		// bytes are not wanted any more and the handle is closed now.
		if err := os.Remove(j.partPath(p.Index)); err != nil && !os.IsNotExist(err) {
			fmt.Printf("server: job %s part %d not removed: %v\n", j.ID, p.Index, err)
		}
		return
	}
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
	for _, f := range j.files {
		fileFailed := 0
		filePending := 0
		for _, p := range f.parts {
			switch p.State {
			case "error":
				fileFailed++
			case "pending", "downloading":
				filePending++
			}
		}
		failed += fileFailed
		pending += filePending
		if filePending == 0 && fileFailed > 0 && f.State != "done" {
			f.State = "error"
			f.Error = fmt.Sprintf("%d part(s) failed", fileFailed)
		}
	}
	if pending == 0 && failed > 0 {
		j.State = "error"
		j.Error = fmt.Sprintf("%d part(s) failed: retry them from the app", failed)
		return
	}
	if pending == 0 && failed == 0 {
		// Nothing is running any more: the job follows the files that are
		// left. A file paused on its own pauses the job when it is the only
		// one with work left.
		done, canceled, paused, other := 0, 0, 0, 0
		for _, f := range j.files {
			switch f.State {
			case "done":
				done++
			case "canceled":
				canceled++
			case "paused":
				paused++
			default:
				other++
			}
		}
		total := len(j.files)
		switch {
		case total > 0 && done == total:
			j.State = "done"
		case total > 0 && other == 0 && canceled > 0 && paused == 0:
			j.State = "canceled"
		case total > 0 && other == 0 && paused > 0 && canceled == 0:
			j.State = "paused"
		}
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
	if part.file != nil {
		// The part goes back to the queue of its file.
		part.file.pending++
		if part.file.State == "error" {
			part.file.State = "pending"
			part.file.Error = ""
		}
	}
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

// maybeMerge merges the files whose parts are all complete.
func (j *Job) maybeMerge() {
	j.mu.Lock()
	var ready []*File
	for _, f := range j.files {
		if f.Merged != "" || f.State == "merging" {
			continue
		}
		all := true
		for _, p := range f.parts {
			if p.State != "done" {
				all = false
				break
			}
		}
		if all && len(f.parts) > 0 {
			f.State = "merging"
			ready = append(ready, f)
		}
	}
	j.mu.Unlock()
	for _, f := range ready {
		j.mergeFile(f)
	}
}

// mergeFile concatenates the parts of one file and checks the result against
// the size the server announced. The part files are deleted: the merged file
// is the result of the download, the parts were only the pieces of it.
func (j *Job) mergeFile(f *File) {
	merged := filepath.Join(j.dir, "merged-"+f.ID+"-"+f.Name)
	go func() {
		out, err := os.Create(merged)
		if err != nil {
			j.failMerge(f, err.Error())
			return
		}
		for _, p := range f.parts {
			in, err := os.Open(j.partPath(p.Index))
			if err != nil {
				out.Close()
				j.failMerge(f, err.Error())
				return
			}
			if _, err := io.Copy(out, in); err != nil {
				in.Close()
				out.Close()
				j.failMerge(f, err.Error())
				return
			}
			in.Close()
		}
		if err := out.Close(); err != nil {
			j.failMerge(f, err.Error())
			return
		}
		size := int64(0)
		if info, err := os.Stat(merged); err == nil {
			size = info.Size()
		}
		sum := fileSHA256(merged)
		verify := "size matches the server"
		if f.Total > 0 && size != f.Total {
			verify = fmt.Sprintf("size mismatch: %d bytes instead of %d", size, f.Total)
		}
		if j.expectSHA != "" {
			if sum == j.expectSHA {
				verify = "SHA-256 matches the expected value"
			} else {
				verify = "SHA-256 does not match the expected value"
			}
		}
		// The parts are not needed any more: the merged file is the result of
		// the download.
		for _, p := range f.parts {
			if err := os.Remove(j.partPath(p.Index)); err != nil && !os.IsNotExist(err) {
				fmt.Printf("server: job %s part %d not removed: %v\n", j.ID, p.Index, err)
			}
		}
		j.mu.Lock()
		f.Merged = merged
		f.MergedSz = size
		f.SHA = sum
		f.Verify = verify
		f.Error = ""
		// The file may have been stopped while its last part was running: it
		// stays stopped, so the operator can still change its URL.
		if f.State != "stopped" {
			f.State = "done"
		}
		j.Merged, j.MergedSz, j.SHA, j.Verify = merged, size, sum, verify
		j.mu.Unlock()
		fmt.Printf("server: job %s merged %s (%s) %s\n", j.ID, filepath.Base(merged), humanSize(size), verify)
		j.settle()
	}()
}

func (j *Job) failMerge(f *File, msg string) {
	j.mu.Lock()
	if f != nil {
		f.State = "error"
		f.Error = msg
	}
	j.State = "error"
	j.Error = msg
	j.mu.Unlock()
	fmt.Printf("server: job %s merge failed: %s\n", j.ID, msg)
}

func (j *Job) setPartState(p *Part, state, msg string) {
	j.mu.Lock()
	old := p.State
	p.State = state
	if p.file != nil {
		// The scheduler counts the parts that are still waiting: the counter
		// follows the state of the part.
		if old == "pending" && state != "pending" {
			p.file.pending--
		}
		if state == "pending" && old != "pending" {
			p.file.pending++
		}
	}
	if state == "done" && p.Size <= 0 {
		// The size of the file was unknown: the bar is full when the part is.
		p.Percent = 100
	}
	if msg != "" {
		p.Error = msg
	}
	j.mu.Unlock()
}

// resetRunning sends the running parts back to the queue: the bytes already
// on disk are kept, the download continues from them.
func (j *Job) resetRunning() {
	j.mu.Lock()
	defer j.mu.Unlock()
	for _, p := range j.parts {
		if p.State == "downloading" {
			p.State = "pending"
			if p.file != nil {
				p.file.pending++
				if p.file.State == "running" {
					p.file.State = "pending"
				}
			}
		}
	}
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
	j.resetRunning()
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

// cancelJob stops the whole download: every file of the job is canceled.
func (j *Job) cancelJob() {
	j.mu.Lock()
	if j.State == "done" {
		j.mu.Unlock()
		return
	}
	j.State = "canceled"
	for _, f := range j.files {
		if f.State != "done" {
			f.State = "canceled"
		}
	}
	cancel := j.cancel
	j.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	j.wg.Wait()
	j.resetRunning()
}

// ---------------------------------------------------------------- per file --

// fileByID returns the file of the download with the given id.
func (j *Job) fileByID(id string) *File {
	j.mu.Lock()
	defer j.mu.Unlock()
	for _, f := range j.files {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// pauseFile pauses one file of the download: the other files keep going. The
// part that is downloading when the pause arrives finishes first: the pause
// takes effect on the next part.
func (j *Job) pauseFile(f *File) {
	j.mu.Lock()
	if f.State == "done" || f.State == "canceled" {
		j.mu.Unlock()
		return
	}
	f.paused = true
	f.State = "paused"
	j.mu.Unlock()
}

// resumeFile takes a paused file back into the queue.
func (j *Job) resumeFile(f *File) {
	j.mu.Lock()
	if f.State == "done" || f.State == "canceled" {
		j.mu.Unlock()
		return
	}
	f.paused = false
	if f.State == "paused" {
		f.State = "pending"
	}
	j.State = "running"
	j.Error = ""
	j.mu.Unlock()
	j.ensureWorkers()
}

// cancelFile stops one file of the download: the file stays in the list, but
// it is not downloaded any more. The bytes already on disk are kept, so the
// operator can start it again where it stopped. The URL of a stopped file can
// be changed: the file is planned again from the new URL. A complete file can
// be stopped too: its URL becomes editable again and the operator can start
// the download once more.
func (j *Job) cancelFile(f *File) {
	j.mu.Lock()
	if f.State == "canceled" {
		j.mu.Unlock()
		return
	}
	f.paused = true
	f.State = "stopped"
	// A download that was complete is not complete any more: the operator can
	// start the file again from its first byte.
	if j.State == "done" {
		j.State = "paused"
	}
	j.mu.Unlock()
}

// removeFile takes a file out of the download: its parts leave the queue and
// the bytes already on disk are deleted.
func (j *Job) removeFile(f *File) error {
	j.mu.Lock()
	kept := make([]*File, 0, len(j.files))
	var old *File
	for _, x := range j.files {
		if x == f {
			old = x
			continue
		}
		kept = append(kept, x)
	}
	if old == nil {
		j.mu.Unlock()
		return fmt.Errorf("unknown file in the download")
	}
	j.files = kept
	parts := make([]*Part, 0, len(j.parts))
	for _, p := range j.parts {
		if p.file == old {
			continue
		}
		parts = append(parts, p)
	}
	j.parts = parts
	j.Received -= old.Received
	if j.Received < 0 {
		j.Received = 0
	}
	j.Total -= old.Total
	if j.Total < 0 {
		j.Total = 0
	}
	j.PartCnt -= old.PartCnt
	old.State = "canceled"
	j.Percent = percent(j.Received, j.Total)
	indexes := make([]int, 0, len(old.parts))
	merged := old.Merged
	for _, p := range old.parts {
		indexes = append(indexes, p.Index)
	}
	j.mu.Unlock()

	// The bytes of the removed file are deleted: the operator does not want
	// them on disk any more. A part that is downloading keeps its file open:
	// on Windows the file cannot be deleted yet, so the part removes it when
	// it ends.
	for _, index := range indexes {
		if err := os.Remove(j.partPath(index)); err != nil && !os.IsNotExist(err) {
			fmt.Printf("server: job %s part %d not removed yet: %v\n", j.ID, index, err)
		}
	}
	if merged != "" {
		os.Remove(merged)
	}
	return nil
}

// restartFile drops the bytes of a file and queues it again from its first
// byte. The file becomes editable again: the operator can change the URL and
// the name before it starts.
func (j *Job) restartFile(f *File) error {
	j.mu.Lock()
	if f.State == "canceled" {
		j.mu.Unlock()
		return fmt.Errorf("the file has been canceled: it cannot be restarted")
	}
	oldMerged := f.Merged
	oldReceived := f.Received
	for _, p := range f.parts {
		p.State = "pending"
		p.Received = 0
		p.Percent = 0
		p.Error = ""
		p.Attempts++
	}
	f.pending = len(f.parts)
	f.Received = 0
	f.Percent = 0
	f.Error = ""
	f.paused = false
	f.State = "pending"
	f.Merged = ""
	f.MergedSz = 0
	f.SHA = ""
	f.Verify = ""
	j.Received -= oldReceived
	if j.Received < 0 {
		j.Received = 0
	}
	j.Percent = percent(j.Received, j.Total)
	j.State = "running"
	j.Error = ""
	j.mu.Unlock()

	// The bytes of the old attempt are dropped: the file is fetched again
	// from its first byte.
	for _, p := range f.parts {
		if err := os.Remove(j.partPath(p.Index)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	// The merged file of the old attempt goes away too: the link of the files
	// table must not point to the bytes of a download that is not there any
	// more.
	if oldMerged != "" {
		os.Remove(oldMerged)
	}
	j.ensureWorkers()
	return nil
}

// replaceFile swaps the parts of a file with a new plan: the operator changed
// the URL of a file that had not started yet. The old parts leave the queue
// and the new ones take their place.
func (j *Job) replaceFile(old *File, fresh *File) {
	j.mu.Lock()
	kept := make([]*Part, 0, len(j.parts))
	for _, p := range j.parts {
		if p.file == old {
			continue
		}
		kept = append(kept, p)
	}
	j.parts = kept
	j.Received -= old.Received
	if j.Received < 0 {
		j.Received = 0
	}
	j.Total -= old.Total
	if j.Total < 0 {
		j.Total = 0
	}
	j.PartCnt -= old.PartCnt
	fresh.pending = len(fresh.parts)
	for _, p := range fresh.parts {
		p.Index = len(j.parts) + 1
		j.parts = append(j.parts, p)
	}
	for i, f := range j.files {
		if f == old {
			j.files[i] = fresh
			break
		}
	}
	j.Total += fresh.Total
	j.PartCnt += fresh.PartCnt
	j.Percent = percent(j.Received, j.Total)
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

// fileView is the JSON view of a file of a download.
type fileView struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	File     string `json:"file"`
	Priority string `json:"priority"`
	State    string `json:"state"`
	Total    int64  `json:"total"`
	TotalTx  string `json:"totalText"`
	Received int64  `json:"received"`
	RecvTx   string `json:"receivedText"`
	Percent  int    `json:"percent"`
	PartCnt  int    `json:"partCount"`
	PartTx   string `json:"partSizeText"`
	Merged   string `json:"merged"`
	MergedSz int64  `json:"mergedSize"`
	SHA      string `json:"sha"`
	Verify   string `json:"verified"`
	Error    string `json:"error"`
	Paused   bool   `json:"paused"`
}

// partView is the JSON view of a part. It is a struct, not a map, because
// Go sorts map keys alphabetically: with a struct the table shows the columns
// in the order the operator expects.
type partView struct {
	File     string `json:"File"`
	Priority string `json:"Priority"`
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
	files := []fileView{}
	failed := 0
	for _, p := range j.parts {
		if p.State == "error" {
			failed++
		}
		name := ""
		priority := ""
		if p.file != nil {
			name = p.file.Name
			priority = p.file.Priority
		}
		parts = append(parts, partView{
			File:     name,
			Priority: priority,
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
	for _, f := range j.files {
		files = append(files, fileView{
			ID:       f.ID,
			URL:      f.URL,
			File:     f.Name,
			Priority: f.Priority,
			State:    f.State,
			Total:    f.Total,
			TotalTx:  humanSize(f.Total),
			Received: f.Received,
			RecvTx:   humanSize(f.Received),
			Percent:  f.Percent,
			PartCnt:  f.PartCnt,
			PartTx:   f.PartTx,
			Merged:   f.Merged,
			MergedSz: f.MergedSz,
			SHA:      f.SHA,
			Verify:   f.Verify,
			Error:    f.Error,
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
		"files":          files,
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
