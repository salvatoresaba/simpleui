// fileanalyzer - demo companion server for the SimpleUI "File analyzer" example.
//
// The SUI app runs in the browser and uploads the current selection with
// http.postFiles("/api/analyze"). This server receives the multipart files,
// detects their type from a bounded prefix and returns columns/rows for
// table.display.
//
// It serves the compiled web app from --web-dir or, when omitted, from the
// assets embedded into the executable. App and API share an origin, so no CORS
// is needed in the normal run/build scripts.
//
// Development run:
//
//	cd examples/file_analyzer
//	../../simpleui.exe build app.sui --out public
//	cd server
//	go run . --port 8097 --web-dir ../public
//
// Release build (embeds server/web):
//
//	cd examples/file_analyzer
//	./build.ps1
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// embeddedWeb is populated by build.ps1/build.bat. The placeholder README
// keeps `go build` working before the first frontend build.
//
//go:embed all:web
var embeddedWeb embed.FS

const (
	maxBody     = 64 << 20 // 64 MiB total upload
	maxFiles    = 200
	maxPrefix   = 512
	workerCount = 8
)

var columns = []string{"File", "Path", "Size", "MIME", "Category", "Detected by", "Detail"}

type apiResponse struct {
	Error   string     `json:"error,omitempty"`
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}

type fileResult struct {
	name     string
	path     string
	size     int64
	mime     string
	category string
	source   string
	detail   string
	rowErr   string
}

func main() {
	port := 8097
	webDir := ""
	noOpen := false
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--port":
			if i+1 < len(os.Args) {
				if p, err := strconv.Atoi(os.Args[i+1]); err == nil && p > 0 && p < 65536 {
					port = p
				}
				i++
			}
		case "--web-dir":
			if i+1 < len(os.Args) {
				webDir = os.Args[i+1]
				i++
			}
		case "--no-open":
			noOpen = true
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/analyze", handleAnalyze)
	mux.HandleFunc("/api/analyze-folder", handleAnalyzeFolder)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	var root http.Handler
	if webDir != "" {
		root = http.FileServer(http.Dir(webDir))
	} else {
		sub, err := fs.Sub(embeddedWeb, "web")
		if err != nil {
			fmt.Fprintln(os.Stderr, "embedded web assets:", err)
			os.Exit(1)
		}
		root = http.FileServer(http.FS(sub))
	}
	mux.Handle("/", root)

	addr := "127.0.0.1:" + strconv.Itoa(port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
	srv := &http.Server{Handler: logRequests(mux)}
	go func() { _ = srv.Serve(ln) }()

	url := "http://" + ln.Addr().String() + "/"
	fmt.Println("======================================================")
	fmt.Printf("File analyzer running at %v\n", url)
	fmt.Println("The browser uploads selected files to POST /api/analyze.")
	fmt.Println("Only requested files are read; size/count limits are enforced.")
	fmt.Println("======================================================")
	if !noOpen {
		openBrowser(url)
	}
	select {}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		fmt.Printf("server: %v %v -> %v\n", r.Method, r.URL.Path, rec.status)
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

func writeJSON(w http.ResponseWriter, status int, v apiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "POST required", Columns: columns})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := r.ParseMultipartForm(maxBody); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "Invalid multipart upload: " + err.Error(), Columns: columns})
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	files := r.MultipartForm.File["files"]
	paths := r.MultipartForm.Value["paths"]
	if len(files) == 0 {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "No files selected", Columns: columns})
		return
	}
	if len(files) > maxFiles {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "Too many files (limit " + strconv.Itoa(maxFiles) + ")", Columns: columns})
		return
	}

	results := make([]fileResult, len(files))
	sem := make(chan struct{}, workerCount)
	var wg sync.WaitGroup
	for i, h := range files {
		wg.Add(1)
		display := ""
		if i < len(paths) {
			display = paths[i]
		}
		go func(i int, h *multipart.FileHeader, displayPathValue string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = analyzeFile(h, displayPathValue)
		}(i, h, display)
	}
	wg.Wait()

	rows := make([][]string, 0, len(results))
	for _, res := range results {
		rows = append(rows, []string{
			res.name,
			res.path,
			humanSize(res.size),
			res.mime,
			res.category,
			res.source,
			res.detail,
		})
	}
	writeJSON(w, http.StatusOK, apiResponse{Columns: columns, Rows: rows})
}

func displayPath(p, fallback string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		p = fallback
	}
	if len(p) > 4096 {
		p = p[:4096]
	}
	p = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, p)
	if p == "" {
		return fallback
	}
	return p
}

func handleAnalyzeFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "POST required", Columns: columns})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Folder string `json:"folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "Invalid JSON: " + err.Error(), Columns: columns})
		return
	}
	name := displayPath(req.Folder, "")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "Folder name is required", Columns: columns})
		return
	}
	rows := [][]string{{
		name,
		name,
		"-",
		"inode/directory",
		"Directory",
		"folder",
		"folder selected without files",
	}}
	writeJSON(w, http.StatusOK, apiResponse{Columns: columns, Rows: rows})
}

func analyzeFile(h *multipart.FileHeader, display string) fileResult {
	name := filepath.Base(h.Filename)
	if name == "." || name == "" {
		name = "unnamed"
	}
	ext := strings.ToLower(filepath.Ext(name))
	res := fileResult{
		name:     name,
		path:     displayPath(display, name),
		size:     h.Size,
		mime:     "application/octet-stream",
		category: "Other",
		source:   "unknown",
		detail:   "no recognizable content",
	}
	f, err := h.Open()
	if err != nil {
		res.rowErr = "cannot open uploaded file"
		res.detail = res.rowErr
		return res
	}
	defer f.Close()

	prefix := make([]byte, maxPrefix)
	n, readErr := io.ReadFull(f, prefix)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		res.rowErr = "cannot read uploaded file"
		res.detail = res.rowErr
		return res
	}
	detected := ""
	if n > 0 {
		detected = http.DetectContentType(prefix[:n])
	}
	byExt := mime.TypeByExtension(ext)

	res.mime = byExt
	res.source = "extension"
	if detected != "" && detected != "application/octet-stream" {
		res.mime = detected
		res.source = "content"
		if byExt != "" && !sameMIME(detected, byExt) {
			res.detail = "content " + shortMIME(detected) + "; extension " + shortMIME(byExt)
		} else {
			res.detail = "content signature " + shortMIME(detected)
		}
	} else if byExt != "" {
		res.detail = "extension " + ext
	} else {
		res.mime = "application/octet-stream"
		res.source = "unknown"
		res.detail = "unknown extension and generic content"
	}
	if ext != "" && !strings.Contains(res.detail, "extension") {
		res.detail += "; ext " + ext
	}
	res.category = classify(ext, res.mime)
	if res.size < 0 {
		res.size = 0
	}
	return res
}

func sameMIME(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(strings.Split(a, ";")[0]))
	b = strings.ToLower(strings.TrimSpace(strings.Split(b, ";")[0]))
	return a == b
}

func shortMIME(v string) string {
	if i := strings.IndexByte(v, ';'); i >= 0 {
		return v[:i]
	}
	return v
}

func classify(ext, mimeType string) string {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".tif", ".tiff", ".svg", ".heic", ".avif":
		return "Image"
	case ".pdf", ".doc", ".docx", ".odt", ".rtf", ".txt", ".md", ".tex":
		return "Document"
	case ".xls", ".xlsx", ".ods", ".csv", ".tsv":
		return "Spreadsheet"
	case ".ppt", ".pptx", ".odp":
		return "Presentation"
	case ".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac":
		return "Audio"
	case ".mp4", ".mkv", ".mov", ".avi", ".webm", ".m4v":
		return "Video"
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz":
		return "Archive"
	case ".go", ".js", ".mjs", ".ts", ".tsx", ".jsx", ".py", ".java", ".c", ".h", ".cpp", ".hpp", ".cs", ".rs", ".rb", ".php", ".sh", ".ps1", ".sql", ".sui":
		return "Code"
	case ".json", ".jsonl", ".xml", ".yaml", ".yml", ".toml", ".ini":
		return "Data"
	case ".exe", ".dll", ".so", ".dylib", ".msi", ".app", ".bin":
		return "Executable"
	case ".woff", ".woff2", ".ttf", ".otf":
		return "Font"
	}
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "Image"
	case strings.HasPrefix(mimeType, "audio/"):
		return "Audio"
	case strings.HasPrefix(mimeType, "video/"):
		return "Video"
	case mimeType == "application/pdf":
		return "Document"
	case strings.Contains(mimeType, "zip"), strings.Contains(mimeType, "compressed"), strings.Contains(mimeType, "tar"):
		return "Archive"
	case strings.HasPrefix(mimeType, "text/"):
		return "Text"
	}
	return "Other"
}

func humanSize(n int64) string {
	if n < 1024 {
		return strconv.FormatInt(n, 10) + " B"
	}
	units := []string{"KB", "MB", "GB", "TB"}
	f := float64(n)
	for _, u := range units {
		f /= 1024
		if f < 1024 || u == "TB" {
			return fmt.Sprintf("%.1f %s", f, u)
		}
	}
	return strconv.FormatInt(n, 10) + " B"
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
