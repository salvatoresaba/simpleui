// download_manager — the Go backend of the SimpleUI "Download manager" example.
//
// It does three things on one origin:
//
//	GET  /                        the compiled SimpleUI app (assets from --web-dir)
//	GET  /files/{name}            a file to download, with HTTP Range support
//	GET  /api/files               the catalog of downloadable files
//	POST /api/downloads            start a chunked download
//	GET  /api/downloads            status of the download jobs
//	POST /api/downloads/{id}/{act} pause | resume | cancel
//	GET  /api/downloads/{id}/file  the reassembled file
//
// The point of the example is the download manager itself: the file is split
// into parts according to its size, every part is fetched with its own Range
// request, the parts are tracked one by one, a pause keeps the bytes already
// downloaded, and the parts are finally concatenated and verified against the
// SHA-256 of the original file.
//
// Run it after building the app:
//
//	simpleui build app.sui --target web --out public
//	cd server && go run . --port 8090 --web-dir ../public
//
// The example ships run.ps1 / run.bat which do both steps and open the
// browser.

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
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
}

// Job is one download: the file, how it was split, and how far it got.
type Job struct {
	ID       string `json:"id"`
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
	Merged   string `json:"merged"`
	MergedSz int64  `json:"mergedSize"`
	SHA      string `json:"sha"`
	Verify   string `json:"verified"`
	Error    string `json:"error"`
	Created  string `json:"created"`

	parts []*Part
	dir   string
	src   string

	baseURLValue string
	clientValue  *http.Client

	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	workers  int
	expected string
	merging  bool
	seq      int
	throttle int64
}

// Manager keeps the jobs of the process.
type Manager struct {
	mu       sync.Mutex
	jobs     map[string]*Job
	order    []string
	seq      int
	filesDir string
	workDir  string
	baseURL  string
	workers  int
	throttle int64
	client   *http.Client
}

func main() {
	port := 8090
	webDir := filepath.Join("..", "public")
	filesDir := filepath.Join("..", "files")
	workDir := filepath.Join("..", "downloads")
	sampleMiB := 24
	workers := 4
	throttle := int64(0)
	noOpen := false

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--port":
			if i+1 < len(os.Args) {
				if p, err := strconv.Atoi(os.Args[i+1]); err == nil && p >= 0 && p < 65536 {
					port = p
				}
			}
			i++
		case "--web-dir":
			if i+1 < len(os.Args) {
				webDir = os.Args[i+1]
			}
			i++
		case "--files-dir":
			if i+1 < len(os.Args) {
				filesDir = os.Args[i+1]
			}
			i++
		case "--work-dir":
			if i+1 < len(os.Args) {
				workDir = os.Args[i+1]
			}
			i++
		case "--sample-mib":
			if i+1 < len(os.Args) {
				if v, err := strconv.Atoi(os.Args[i+1]); err == nil && v > 0 {
					sampleMiB = v
				}
			}
			i++
		case "--workers":
			if i+1 < len(os.Args) {
				if v, err := strconv.Atoi(os.Args[i+1]); err == nil && v > 0 && v <= 32 {
					workers = v
				}
			}
			i++
		case "--throttle":
			if i+1 < len(os.Args) {
				if v, err := strconv.ParseInt(os.Args[i+1], 10, 64); err == nil && v >= 0 {
					throttle = v
				}
			}
			i++
		case "--no-open":
			noOpen = true
		}
	}

	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "files dir:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "work dir:", err)
		os.Exit(1)
	}
	sample := filepath.Join(filesDir, "sample.bin")
	if _, err := os.Stat(sample); os.IsNotExist(err) {
		if err := generateSample(sample, int64(sampleMiB)); err != nil {
			fmt.Fprintln(os.Stderr, "sample file:", err)
			os.Exit(1)
		}
		fmt.Printf("server: created the sample file %s (%d MiB)\n", sample, sampleMiB)
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	// Port 0 asks the OS for a free port: the base URL is then built from
	// the address the listener actually got.
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
	baseURL = "http://" + ln.Addr().String()

	mgr := &Manager{
		jobs:     map[string]*Job{},
		filesDir: filesDir,
		workDir:  workDir,
		baseURL:  baseURL,
		workers:  workers,
		throttle: throttle,
		client:   &http.Client{Timeout: 60 * time.Second},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/files", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, mgr.catalog())
	})
	mux.HandleFunc("POST /api/downloads", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleCreate(w, r)
	})
	mux.HandleFunc("GET /api/downloads", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"jobs": mgr.snapshot()})
	})
	mux.HandleFunc("POST /api/downloads/{id}/{action}", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleAction(w, r)
	})
	mux.HandleFunc("GET /api/downloads/{id}/file", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleMergedFile(w, r)
	})
	mux.HandleFunc("GET /files/{name}", func(w http.ResponseWriter, r *http.Request) {
		mgr.handleFile(w, r)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		file := filepath.Join(webDir, filepath.FromSlash(strings.TrimPrefix(path, "/")))
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

	url := baseURL + "/"
	fmt.Println("======================================================")
	fmt.Printf("download manager + app running at %v\n", url)
	fmt.Printf("  files dir : %v\n", absPath(filesDir))
	fmt.Printf("  work dir  : %v\n", absPath(workDir))
	fmt.Println("  the app calls /api/downloads, /api/files and /api/downloads/{id}/...")
	fmt.Println("======================================================")
	if !noOpen {
		openBrowser(url)
	}
	select {}
}

// ---------------------------------------------------------------- catalog ---

type catalogEntry struct {
	Name string `json:"Name"`
	Size int64  `json:"Size"`
	Text string `json:"SizeText"`
}

func (m *Manager) catalog() []catalogEntry {
	entries, err := os.ReadDir(m.filesDir)
	if err != nil {
		return []catalogEntry{}
	}
	out := []catalogEntry{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, catalogEntry{
			Name: e.Name(),
			Size: info.Size(),
			Text: humanSize(info.Size()),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ------------------------------------------------------------------ files ---

// handleFile serves a file with Range support: the download manager asks for
// one part at a time with "bytes=start-end" and gets a 206 back.
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
	File     string `json:"file"`
	PartSize string `json:"partSize"`
	Workers  int    `json:"workers"`
	Throttle int64  `json:"throttle"`
}

func (m *Manager) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}
	name := filepath.Base(strings.TrimSpace(req.File))
	if name == "" || name == "." || name == ".." {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid file name"})
		return
	}
	src := filepath.Join(m.filesDir, name)
	info, err := os.Stat(src)
	if err != nil || info.IsDir() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found: " + name})
		return
	}
	partSize, err := parseSize(req.PartSize)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	total := info.Size()
	partCount := int((total + partSize - 1) / partSize)
	if partCount < 1 {
		partCount = 1
	}
	if partCount > 512 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("too many parts (%d): use a bigger part size", partCount)})
		return
	}
	workers := req.Workers
	if workers <= 0 {
		workers = m.workers
	}
	if workers > partCount {
		workers = partCount
	}
	throttle := req.Throttle
	if throttle <= 0 {
		throttle = m.throttle
	}

	m.mu.Lock()
	m.seq++
	id := fmt.Sprintf("job-%d-%d", time.Now().Unix(), m.seq)
	m.mu.Unlock()

	job := &Job{
		ID:       id,
		File:     name,
		State:    "running",
		PartSize: partSize,
		Total:    total,
		PartCnt:  partCount,
		PartTx:   humanSize(partSize),
		Created:  time.Now().Format("15:04:05"),
		dir:      filepath.Join(m.workDir, id),
		src:      src,
		workers:  workers,

		baseURLValue: m.baseURL,
		clientValue:  m.client,
		throttle:     throttle,
	}
	for i := 0; i < partCount; i++ {
		start := int64(i) * partSize
		end := start + partSize - 1
		if end > total-1 {
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
	job.expected = fileSHA256(src)
	fmt.Printf("server: new job %s  file=%s size=%s parts=%d x %s workers=%d\n",
		id, name, humanSize(total), partCount, job.PartTx, workers)

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
	m.mu.Lock()
	job := m.jobs[id]
	m.mu.Unlock()
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

func (m *Manager) handleMergedFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	m.mu.Lock()
	job := m.jobs[id]
	m.mu.Unlock()
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
		m.mu.Lock()
		job := m.jobs[ids[i]]
		m.mu.Unlock()
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
	ctx, cancel := context.WithCancel(context.Background())
	j.ctx, j.cancel = ctx, cancel
	j.mu.Unlock()
	for i := 0; i < j.workers; i++ {
		j.wg.Add(1)
		go j.worker(ctx)
	}
	go j.watcher(ctx)
}

func (j *Job) worker(ctx context.Context) {
	defer j.wg.Done()
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

// progressWriter updates the counters while the bytes flow, and logs the
// part progress so the server console tells the same story as the GUI.
type progressWriter struct {
	job   *Job
	part  *Part
	w     io.Writer
	last  time.Time
	shout bool
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
		now := time.Now()
		if !p.shout && (now.Sub(p.last) > 500*time.Millisecond || p.part.Received == p.part.Size) {
			p.shout = true
			p.last = now
		}
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

	url := fmt.Sprintf("%s/files/%s", j.baseURL(), j.File)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		j.setPartState(p, "error", err.Error())
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, p.End))
	resp, err := j.client().Do(req)
	if err != nil {
		if ctx.Err() != nil {
			j.setPartState(p, "pending", "")
			return
		}
		j.setPartState(p, "error", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		j.setPartState(p, "error", fmt.Sprintf("HTTP %d from the file server", resp.StatusCode))
		return
	}

	path := j.partPath(p.Index)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		j.setPartState(p, "error", err.Error())
		return
	}
	defer f.Close()

	w := &progressWriter{job: j, part: p, w: f, last: time.Now()}
	if _, err := io.Copy(w, resp.Body); err != nil {
		if ctx.Err() != nil {
			// Paused: the bytes already written stay on disk.
			j.setPartState(p, "pending", "")
			return
		}
		j.setPartState(p, "error", err.Error())
		return
	}

	j.mu.Lock()
	complete := p.Received >= p.Size
	j.mu.Unlock()
	if !complete {
		j.setPartState(p, "error", fmt.Sprintf("received %d bytes out of %d", p.Received, p.Size))
		return
	}
	j.setPartState(p, "done", "")
	fmt.Printf("server: job %s part %d/%d ok (%s)\n", j.ID, p.Index, j.PartCnt, humanSize(p.Size))
	j.maybeMerge()
}

// maybeMerge concatenates the parts when they are all complete and checks
// the result against the SHA-256 of the original file.
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
		verify := "SHA-256 differs from the original"
		if sum == j.expected {
			verify = "SHA-256 matches the original"
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
}

// snapshot is the JSON view of the job.
func (j *Job) snapshot() map[string]any {
	j.mu.Lock()
	defer j.mu.Unlock()
	parts := []partView{}
	for _, p := range j.parts {
		parts = append(parts, partView{
			Part:     p.Index,
			From:     p.Start,
			To:       p.End,
			Size:     p.Size,
			Received: p.Received,
			Percent:  p.Percent,
			State:    p.State,
			Error:    p.Error,
		})
	}
	return map[string]any{
		"id":           j.ID,
		"file":         j.File,
		"state":        j.State,
		"partSize":     j.PartSize,
		"partSizeText": j.PartTx,
		"partCount":    j.PartCnt,
		"throttle":     j.throttle,
		"total":        j.Total,
		"totalText":    humanSize(j.Total),
		"received":     j.Received,
		"receivedText": humanSize(j.Received),
		"percent":      j.Percent,
		"speed":        j.Speed,
		"speedText":    j.SpeedTx,
		"eta":          j.ETA,
		"etaText":      j.ETATx,
		"merged":       j.Merged,
		"mergedSize":   j.MergedSz,
		"sha":          j.SHA,
		"verified":     j.Verify,
		"error":        j.Error,
		"created":      j.Created,
		"parts":        parts,
	}
}

// ----------------------------------------------------------------- paths ---

func (j *Job) baseURL() string { return j.baseURLValue }

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
