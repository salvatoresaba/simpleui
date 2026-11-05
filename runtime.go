package main

import "strings"

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
	"encoding/hex"
	"fmt"
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
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("POST /__bridge__/fs/stat", makeFstat(secret))
	mux.HandleFunc("POST /__bridge__/fs/list", makeList(secret))
	mux.HandleFunc("POST /__bridge__/fs/mkdir", makeMkdir(secret))
	mux.HandleFunc("POST /__bridge__/fs/remove", makeRemove(secret))
	mux.HandleFunc("POST /__bridge__/exec", makeExec(secret))
	mux.HandleFunc("POST /__bridge__/pick-file", makePickFile(secret))
	mux.HandleFunc("POST /__bridge__/pick-directory", makePickDirectory(secret))

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
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
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
