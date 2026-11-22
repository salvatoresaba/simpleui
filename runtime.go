package sui

import (
	"strconv"
	"strings"
)

// runtime.go — the desktop runtime template.
//
// `simpleui build <file.sui> --target desktop` materializes this file as the
// mod of a tiny disposable module (go.mod "desktopapp") next to the embedded
// web assets, then runs `go build`. The result is a single self-contained
// executable `app(.exe)` that:
//
//  1. embeds the compiled HTML/CSS/JS (go:embed),
//  2. listens on an ephemeral port of 127.0.0.1 only,
//  3. injects a random per-run secret into the served app.js, and
//  4. exposes the native bridge (exec + fs) behind that secret, honouring the
//     allow-list baked in at build time (capability-based security, DESIGN 5).
//
// The window is opened in the system default browser (DESIGN 3.2/4.1); a
// native WebView seam is documented in SPEC §8. `--no-open` runs the server
// without opening anything (used by tests/headless).
//
// TokenPlaceholder is the literal the generated app.js starts with; the server
// substitutes the per-run secret where this placeholder appears.

const TokenPlaceholder = "__SUI_TOKEN__"

// DesktopMainSource returns the main.go source for a desktop build, with the
// capability allow-lists compiled in.
func DesktopMainSource(allowExec []string, allowDirs []string) string {
	src := desktopRuntimeTemplate
	src = strings.ReplaceAll(src, "%%ALLOW_EXEC%%", joinGoStrings(allowExec))
	src = strings.ReplaceAll(src, "%%ALLOW_DIRS%%", joinGoStrings(allowDirs))
	return src
}

// DesktopWebviewMainSource returns the main.go source for the embedded-WebView
// target. It reuses the desktop runtime (loopback server, token and bridge) and
// replaces the browser launch with a native WebView window.
func DesktopWebviewMainSource(allowExec []string, allowDirs []string, title string) string {
	src := DesktopMainSource(allowExec, allowDirs)
	src = strings.Replace(src, "\t\"time\"\n)", "\t\"time\"\n\n\twebview \"github.com/webview/webview_go\"\n)", 1)
	const openBlock = "\tif !noOpen {\n\t\topenBrowser(url)\n\t}\n\tselect {}\n"
	webviewBlock := "\tif noOpen {\n\t\tselect {}\n\t}\n" +
		"\tw := webview.New(false)\n" +
		"\tdefer w.Destroy()\n" +
		"\tw.SetTitle(" + strconv.Quote(title) + ")\n" +
		"\tw.SetSize(1200, 800, webview.HintNone)\n" +
		"\tw.Navigate(url)\n" +
		"\tw.Run()\n"
	src = strings.Replace(src, openBlock, webviewBlock, 1)
	return src
}

// joinGoStrings renders a []string as a Go slice literal ("" for empty).
func joinGoStrings(items []string) string {
	var parts []string
	for i := 0; i < len(items); i++ {
		s := items[i]
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		parts = append(parts, `"`+s+`"`)
	}
	return strings.Join(parts, ", ")
}

// desktopRuntimeTemplate is the body of the generated desktop main.go.
//
// IMPORTANT: this template is stored inside a raw string literal, so the
// generated source must NOT contain a backtick character. Every string in the
// template is therefore a double-quoted Go literal with explicit escaping.
// The only placeholders are %%ALLOW_EXEC%% and %%ALLOW_DIRS%%.
const desktopRuntimeTemplate = `package main

import (
	_ "embed"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//go:embed assets/index.html
var assetIndex string

//go:embed assets/style.css
var assetStyle string

//go:embed assets/app.js
var assetAppJS string

// allowExec lists the command names this app may launch through the bridge
// (capability-based security). Everything else is refused.
var allowExec = []string{%%ALLOW_EXEC%%}

// allowDirs lists the base directories the app may read/write through the
// bridge. Access outside these prefixes is refused.
var allowDirs = []string{%%ALLOW_DIRS%%}

func main() {
	noOpen := false
	port := 0
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--no-open" {
			noOpen = true
		}
		if os.Args[i] == "--port" && i+1 < len(os.Args) {
			p, err := strconv.Atoi(os.Args[i+1])
			if err == nil && p > 0 && p < 65536 {
				port = p
			}
		}
	}
	secret := newToken(32)
	// The document secret is separate from the bridge secret: it is handed to
	// the browser in a cookie scoped to the document endpoint, so it can never
	// be used for the other bridge calls (exec, file write, ...).
	docSecret := newToken(32)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// The document cookie is set when the app page is served: the PDF
		// viewer frame loads the document from the same origin and the
		// browser sends the cookie back.
		http.SetCookie(w, &http.Cookie{
			Name:     "sui_doc",
			Value:    docSecret,
			Path:     documentPath,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		writeRaw(w, "text/html; charset=utf-8", assetIndex)
	})
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		writeRaw(w, "text/css; charset=utf-8", assetStyle)
	})
	mux.HandleFunc("GET /app.js", func(w http.ResponseWriter, r *http.Request) {
		// app.js carries the bridge token. Refuse cross-site script inclusion;
		// DNS rebinding is already blocked by the Host check below.
		if !sameSiteRequest(r) {
			http.Error(w, "forbidden: cross-site script", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		js := strings.ReplaceAll(assetAppJS, "__SUI_TOKEN__", secret)
		writeRaw(w, "application/javascript; charset=utf-8", js)
	})
	mux.HandleFunc("POST /__bridge__/file/read", makeRead(secret))
	mux.HandleFunc("POST /__bridge__/file/write", makeWrite(secret))
	mux.HandleFunc("POST /__bridge__/file/readbytes", makeReadBytes(secret))
	mux.HandleFunc("POST /__bridge__/file/writebytes", makeWriteBytes(secret))
	mux.HandleFunc("POST /__bridge__/fs/stat", makeFstat(secret))
	mux.HandleFunc("POST /__bridge__/fs/list", makeList(secret))
	mux.HandleFunc("POST /__bridge__/fs/mkdir", makeMkdir(secret))
	mux.HandleFunc("POST /__bridge__/fs/remove", makeRemove(secret))
	mux.HandleFunc("POST /__bridge__/fs/copy", makeCopy(secret))
	mux.HandleFunc("POST /__bridge__/fs/move", makeMove(secret))
	mux.HandleFunc("POST /__bridge__/fs/size", makeSize(secret))
	mux.HandleFunc("POST /__bridge__/exec", makeExec(secret))
	mux.HandleFunc("POST /__bridge__/pick-file", makePickFile(secret))
	mux.HandleFunc("POST /__bridge__/pick-directory", makePickDirectory(secret))
	mux.HandleFunc("GET "+documentPath, makeDocument(docSecret))

	addr := "127.0.0.1:0"
	if port != 0 {
		addr = "127.0.0.1:" + strconv.Itoa(port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
	addr = ln.Addr().String()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bind is loopback-only; reject foreign Host headers too, which blocks
		// DNS-rebinding tricks that would otherwise read app.js (and its token).
		if !loopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if r.Method == http.MethodPost && strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
			http.Error(w, "forbidden: cross-site request", http.StatusForbidden)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.URL.Path != documentPath {
			// The document endpoint is the only response that may be framed:
			// the PDF viewer needs an embeddable frame, so the global
			// anti-framing headers are not applied to it.
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
		}
		mux.ServeHTTP(w, r)
	})
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		_ = server.Serve(ln)
	}()
	url := "http://" + addr + "/"
	fmt.Printf("SimpleUI app running at %v\n", url)
	if !noOpen {
		openBrowser(url)
	}
	select {}
}

func writeRaw(w http.ResponseWriter, ct string, body string) {
	w.Header().Set("Content-Type", ct)
	_, _ = w.Write([]byte(body))
}

// newToken returns a hex-encoded random token of n random bytes.
func newToken(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// bridgeJSON writes a uniform bridge response as JSON.
func bridgeJSON(w http.ResponseWriter, ok bool, data string, message string) {
	var sb strings.Builder
	sb.WriteString("{\"ok\":")
	if ok {
		sb.WriteString("true")
	} else {
		sb.WriteString("false")
	}
	sb.WriteString(",\"data\":")
	sb.WriteString(jsonStr(data))
	sb.WriteString(",\"message\":")
	sb.WriteString(jsonStr(message))
	sb.WriteString("}")
	writeRaw(w, "application/json; charset=utf-8", sb.String())
}

// jsonStr quotes and escapes a string for embedding in JSON.
func jsonStr(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\':
			sb.WriteString("\\\\")
		case '"':
			sb.WriteString("\\\"")
		case '\n':
			sb.WriteString("\\n")
		case '\t':
			sb.WriteString("\\t")
		case '\r':
			sb.WriteString("\\r")
		default:
			sb.WriteByte(c)
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

func postForm(r *http.Request) error {
	return r.ParseForm()
}

func secretOk(secret string, r *http.Request) bool {
	got := r.PostFormValue("token")
	if got == "" || len(got) != len(secret) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(secret)) == 1
}

func makeRead(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			bridgeJSON(w, false, "", "fs.read: "+err.Error())
			return
		}
		bridgeJSON(w, true, string(data), "")
	}
}

// makeReadBytes returns a base64 payload so arbitrary bytes (including UTF-16
// XML) survive the JSON bridge.
func makeReadBytes(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			bridgeJSON(w, false, "", "fs.readBytes: "+err.Error())
			return
		}
		bridgeJSON(w, true, base64.StdEncoding.EncodeToString(data), "")
	}
}

// documentPath is the bridge endpoint that serves a PDF document to the
// browser's native viewer. It is a GET (an iframe cannot POST) and it is
// protected by the per-run document cookie, not by the bridge token.
const documentPath = "/__bridge__/file/document"

// documentSecretOk checks the per-run document cookie. The cookie is scoped
// to the document endpoint, so it cannot be replayed on the other bridge
// endpoints, which keep using the bridge token.
func documentSecretOk(secret string, r *http.Request) bool {
	cookie, err := r.Cookie("sui_doc")
	if err != nil || len(cookie.Value) != len(secret) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(secret)) == 1
}

// makeDocument serves one PDF file to the native viewer. The frame that
// shows it is not sandboxed (the native viewer does not start inside a
// sandboxed frame), so the endpoint is deliberately narrow: it reads only
// files inside the allowed directories, it serves only files whose first
// bytes are the PDF magic "%PDF-", and it never writes anything.
func makeDocument(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !documentSecretOk(secret, r) {
			http.Error(w, "forbidden: missing or invalid document cookie", http.StatusForbidden)
			return
		}
		if strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
			http.Error(w, "forbidden: cross-site request", http.StatusForbidden)
			return
		}
		path := r.URL.Query().Get("path")
		if !pathAllowed(path) {
			http.Error(w, "forbidden: path outside the allowed directories", http.StatusForbidden)
			return
		}
		info, err := os.Stat(path)
		if err != nil {
			http.Error(w, "document: "+err.Error(), http.StatusNotFound)
			return
		}
		if info.IsDir() {
			http.Error(w, "document: path is a directory", http.StatusBadRequest)
			return
		}
		file, err := os.Open(path)
		if err != nil {
			http.Error(w, "document: "+err.Error(), http.StatusNotFound)
			return
		}
		defer file.Close()
		head := make([]byte, 5)
		n, _ := io.ReadFull(file, head)
		if n < 5 || string(head) != "%PDF-" {
			http.Error(w, "document: not a PDF file", http.StatusUnsupportedMediaType)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		size := info.Size()
		start, end, hasRange := parseByteRange(r.Header.Get("Range"), size)
		if hasRange {
			w.Header().Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+
				strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(size, 10))
			w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
			w.WriteHeader(http.StatusPartialContent)
			if _, err := file.Seek(start, io.SeekStart); err == nil {
				_, _ = io.CopyN(w, file, end-start+1)
			}
			return
		}
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		w.WriteHeader(http.StatusOK)
		if _, err := file.Seek(0, io.SeekStart); err == nil {
			_, _ = io.Copy(w, file)
		}
	}
}

// parseByteRange parses a single "bytes=start-end" range. It reports false
// when the header is absent or is not a simple range: the caller then sends
// the whole document, which every browser accepts.
func parseByteRange(header string, size int64) (int64, int64, bool) {
	if !strings.HasPrefix(header, "bytes=") || size <= 0 {
		return 0, 0, false
	}
	spec := strings.TrimSpace(strings.TrimPrefix(header, "bytes="))
	if strings.Contains(spec, ",") {
		return 0, 0, false
	}
	dash := strings.Index(spec, "-")
	if dash < 0 {
		return 0, 0, false
	}
	startText := strings.TrimSpace(spec[:dash])
	endText := strings.TrimSpace(spec[dash+1:])
	if startText == "" {
		return 0, 0, false
	}
	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false
	}
	end := size - 1
	if endText != "" {
		last, err := strconv.ParseInt(endText, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		if last < start {
			return 0, 0, false
		}
		if last < end {
			end = last
		}
	}
	return start, end, true
}

func makeWriteBytes(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		data, err := base64.StdEncoding.DecodeString(r.PostFormValue("data"))
		if err != nil {
			bridgeJSON(w, false, "", "fs.writeBytes: invalid base64 payload")
			return
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			bridgeJSON(w, false, "", "fs.writeBytes: "+err.Error())
			return
		}
		bridgeJSON(w, true, "", "")
	}
}

func makeWrite(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		data := r.PostFormValue("data")
		err := os.WriteFile(path, []byte(data), 0o644)
		if err != nil {
			bridgeJSON(w, false, "", "fs.write: "+err.Error())
			return
		}
		bridgeJSON(w, true, "", "")
	}
}

// makeFstat answers fs.exists/fs.stat: data is "file", "dir" or "missing".
func makeFstat(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		fi, serr := os.Stat(path)
		if serr != nil {
			bridgeJSON(w, true, "missing", "")
			return
		}
		if fi.IsDir() {
			bridgeJSON(w, true, "dir", "")
			return
		}
		bridgeJSON(w, true, "file", "")
	}
}

func makeMkdir(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			bridgeJSON(w, false, "", "fs.mkdir: "+err.Error())
			return
		}
		bridgeJSON(w, true, "", "")
	}
}

// makeList answers fs.list: directory entry names, one per line.
func makeList(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			bridgeJSON(w, false, "", "fs.list: "+err.Error())
			return
		}
		var names []string
		for i := 0; i < len(entries); i++ {
			names = append(names, entries[i].Name())
		}
		bridgeJSON(w, true, strings.Join(names, "\n"), "")
	}
}

func makeRemove(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		err := os.Remove(path)
		if err != nil {
			bridgeJSON(w, false, "", "fs.remove: "+err.Error())
			return
		}
		bridgeJSON(w, true, "", "")
	}
}

// makeCopy copies one file to another. Both paths must be inside the allowed
// directories; directories are refused (copying a directory needs a recursive
// policy and is intentionally not exposed).
func makeCopy(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		src := r.PostFormValue("src")
		dst := r.PostFormValue("dst")
		if !pathAllowed(src) || !pathAllowed(dst) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		if err := copyFile(src, dst); err != nil {
			bridgeJSON(w, false, "", "fs.copy: "+err.Error())
			return
		}
		bridgeJSON(w, true, "", "")
	}
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("source is a directory; use fs.move for directories")
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// makeMove renames a file or directory. Source and destination are both
// capability-checked; cross-device moves return the OS error.
func makeMove(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		src := r.PostFormValue("src")
		dst := r.PostFormValue("dst")
		if !pathAllowed(src) || !pathAllowed(dst) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		if err := os.Rename(src, dst); err != nil {
			bridgeJSON(w, false, "", "fs.move: "+err.Error())
			return
		}
		bridgeJSON(w, true, "", "")
	}
}

// makeSize answers fs.size for a file. Directories are rejected because their
// recursive size is a different, potentially expensive operation.
func makeSize(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		path := r.PostFormValue("path")
		if !pathAllowed(path) {
			bridgeJSON(w, false, "", "forbidden: path outside the allowed directories")
			return
		}
		info, err := os.Stat(path)
		if err != nil {
			bridgeJSON(w, false, "", "fs.size: "+err.Error())
			return
		}
		if info.IsDir() {
			bridgeJSON(w, false, "", "fs.size: path is a directory")
			return
		}
		bridgeJSON(w, true, strconv.FormatInt(info.Size(), 10), "")
	}
}

func makeExec(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			bridgeJSON(w, false, "", "forbidden: missing or invalid token")
			return
		}
		cmd := r.PostFormValue("cmd")
		args := r.Form["arg"]
		if !execAllowed(cmd) {
			bridgeJSON(w, false, "", "forbidden: command not in the allow-list")
			return
		}
		// Prefer the app's own directory for relative command names, so a
		// companion binary shipped next to app(.exe) is found regardless of
		// the current working directory.
		resolved := resolveExecutable(cmd)
		out, err := exec.Command(resolved, args...).CombinedOutput()
		if err != nil {
			bridgeJSON(w, false, string(out), "exec: "+err.Error())
			return
		}
		bridgeJSON(w, true, string(out), "")
	}
}

// resolveExecutable resolves a command name that contains no path separator
// against the directory of the running executable, falling back to PATH.
func resolveExecutable(name string) string {
	for i := 0; i < len(name); i++ {
		if name[i] == '/' || name[i] == '\\' {
			return name
		}
	}
	if dir, err := os.Executable(); err == nil {
		cand := filepath.Join(filepath.Dir(dir), name)
		if _, statErr := os.Stat(cand); statErr == nil {
			return cand
		}
	}
	return name
}

func pathAllowed(path string) bool {
	if path == "" || strings.IndexByte(path, 0) >= 0 {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	abs = filepath.Clean(abs)
	realAbs := resolveRealPath(abs)
	for i := 0; i < len(allowDirs); i++ {
		base, err := filepath.Abs(allowDirs[i])
		if err != nil {
			continue
		}
		base = filepath.Clean(base)
		realBase := resolveRealPath(base)
		// Check the symlink-resolved path only: a link inside an allowed
		// directory must not point outside it.
		if insidePath(realAbs, realBase) {
			return true
		}
	}
	return false
}

// resolveRealPath resolves symlinks on the longest existing prefix, so it also
// works for files that do not exist yet (fs.write / mkdir).
func resolveRealPath(p string) string {
	p = filepath.Clean(p)
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return filepath.Clean(r)
	}
	dir, name := filepath.Split(p)
	if dir == "" || filepath.Clean(dir) == p {
		return p
	}
	return filepath.Join(resolveRealPath(filepath.Clean(dir)), name)
}

func insidePath(p, base string) bool {
	if p == base {
		return true
	}
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
		base = strings.ToLower(base)
	}
	return strings.HasPrefix(p, base+string(os.PathSeparator))
}

func loopbackHost(h string) bool {
	host := h
	if hp, _, err := net.SplitHostPort(h); err == nil {
		host = hp
	}
	host = strings.Trim(host, "[]")
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}

// sameSiteRequest rejects cross-site script/API inclusion while allowing the
// app's own pages (and clients that omit Referer/Sec-Fetch-Site).
func sameSiteRequest(r *http.Request) bool {
	if strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
		return false
	}
	ref := r.Header.Get("Referer")
	if ref == "" {
		return true
	}
	u, err := url.Parse(ref)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}

func execAllowed(cmd string) bool {
	for i := 0; i < len(allowExec); i++ {
		if allowExec[i] == cmd {
			return true
		}
	}
	return false
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
`
