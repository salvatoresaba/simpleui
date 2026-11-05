package main

import "fmt"
import "net"
import "net/http"
import "os"
import "os/exec"
import "path/filepath"
import "runtime"
import "strconv"
import "strings"
import "time"

// serve.go - `simpleui serve --watch` development server.
//
// It builds the web assets of a .sui app into a temp directory, serves them
// same-origin, and (with --watch) recompiles on source changes, pushing an
// SSE reload event so the browser refreshes automatically.

// CmdServe parses `simpleui serve <file.sui> [--port N] [--watch] [--out DIR]`.
func CmdServe(args []string) {
	port := 0
	watch := false
	out := ""
	input := ""
	security := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			name, value, hasValue := SplitFlag(arg)
			needsValue := name == "port" || name == "out" || name == "security"
			if !hasValue && needsValue && i+1 < len(args) {
				value, hasValue = args[i+1], true
				i++
			}
			if name == "port" {
				if !hasValue {
					UsageError("--port needs a value")
				}
				p, err := strconv.Atoi(value)
				if err == nil {
					port = p
				}
			} else if name == "watch" {
				watch = true
			} else if name == "out" {
				if !hasValue {
					UsageError("--out needs a value")
				}
				out = value
			} else if name == "security" {
				if !hasValue {
					UsageError("--security needs a JSON file")
				}
				security = value
			} else {
				UsageError("unknown flag --" + name)
			}
		} else if input == "" {
			input = arg
		} else {
			UsageError("unexpected argument " + arg)
		}
	}
	if input == "" {
		UsageError("missing input file")
	}
	SetSecurityPath(security)
	if out == "" {
		tmp, err := os.MkdirTemp("", "sui-serve-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
			os.Exit(1)
		}
		out = tmp
	}

	hr := NewHotReloader()
	if err := BuildWebInto(input, out, hr, watch); err != nil {
		fmt.Fprintf(os.Stderr, "%v: build failed: %v\n", AppName, err)
		os.Exit(1)
	}

	if watch {
		go watchLoop(input, out, hr)
	}

	addr := "127.0.0.1:0"
	if port != 0 {
		addr = "127.0.0.1:" + strconv.Itoa(port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: listen: %v\n", AppName, err)
		os.Exit(1)
	}
	handler := appServer(out, watch, hr)
	url := "http://" + ln.Addr().String() + "/"
	fmt.Printf("%v serve: %v\n", AppName, url)
	if watch {
		fmt.Printf("%v watch enabled: editing the .sui file reloads the browser\n", AppName)
	}
	go func() {
		err := http.Serve(ln, handler)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: server: %v\n", AppName, err)
		}
	}()
	openBrowserCommand(url)
	select {}
}

// BuildWebInto compiles entryFile to a directory and Notifies the reloader on
// success (when notify is true). It returns an error with the diagnostics.
func BuildWebInto(entryFile string, outDir string, hr *HotReloader, notify bool) error {
	prog, dl := LoadProgram(entryFile)
	if dl.HasErrors() {
		dl.Print()
		return Fatal("the program has %v error(s)", dl.Count())
	}
	cerrs := Check(prog)
	dl.Extend(&cerrs)
	if dl.HasErrors() {
		dl.Print()
		return Fatal("the program has %v error(s)", dl.Count())
	}
	if dl.Count() > 0 {
		dl.Print() // warnings: keep building
	}
	assets := map[string]string{
		"index.html": GenerateIndexHtml(prog),
		"style.css":  GenerateCss(prog),
		"app.js":     GenerateJs(prog),
	}
	for name, body := range assets {
		path := filepath.Join(outDir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
	}
	if notify {
		hr.NotifyBuild()
	}
	return nil
}

// appServer serves the built assets; when watch is enabled it injects the
// live-reload <script> into the served index.html and exposes the SSE stream.
func appServer(assetsDir string, watch bool, hr *HotReloader) http.Handler {
	mux := http.NewServeMux()
	if watch {
		mux.HandleFunc("GET /__sui__/reload", hr.ServeSSE)
		mux.HandleFunc("GET /__sui__/reload.js", func(w http.ResponseWriter, r *http.Request) {
			writeRawHandler(w, "application/javascript; charset=utf-8", "(function(){var es=new EventSource(\"/__sui__/reload\");es.onmessage=function(){location.reload();};})();")
		})
	}
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// The app must not be framed by another site (clickjacking). This does
		// not restrict iframes embedded BY the app.
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if rel == "" {
			rel = "index.html"
		}
		data, err := os.ReadFile(filepath.Join(assetsDir, rel))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ct := "application/octet-stream"
		if strings.HasSuffix(rel, ".html") {
			ct = "text/html; charset=utf-8"
		} else if strings.HasSuffix(rel, ".css") {
			ct = "text/css; charset=utf-8"
		} else if strings.HasSuffix(rel, ".js") {
			ct = "application/javascript; charset=utf-8"
		}
		body := string(data)
		if watch && rel == "index.html" {
			body = injectReload(body)
		}
		w.Header().Set("Content-Type", ct)
		_, _ = w.Write([]byte(body))
	})
	return mux
}

func writeRawHandler(w http.ResponseWriter, ct string, body string) {
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	_, _ = w.Write([]byte(body))
}

func injectReload(html string) string {
	const script = "\n<script src=\"/__sui__/reload.js\"></script>\n</body>"
	idx := strings.LastIndex(html, "</body>")
	if idx < 0 {
		return html + "\n<script src=\"/__sui__/reload.js\"></script>"
	}
	return html[:idx] + script
}

// HotReloader fans build events out to connected SSE clients.
type HotReloader struct {
	clients map[chan struct{}]bool
}

func NewHotReloader() *HotReloader {
	return &HotReloader{clients: map[chan struct{}]bool{}}
}

func (h *HotReloader) NotifyBuild() {
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (h *HotReloader) ServeSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	fl, ok := w.(http.Flusher)
	if !ok {
		return
	}
	ch := make(chan struct{}, 4)
	h.clients[ch] = true
	defer func() {
		delete(h.clients, ch)
	}()
	// Initial "build is current" event so the page reloads on first attach.
	ch <- struct{}{}
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			_, _ = fmt.Fprintf(w, "event: reload\ndata: ok\n\n")
			fl.Flush()
		case <-time.After(20 * time.Second):
			_, _ = w.Write([]byte(": keep-alive\n\n"))
			fl.Flush()
		}
	}
}

// watchLoop polls the source file (and the whole .sui files in its directory)
// and rebuilds on change.
func watchLoop(entry string, out string, hr *HotReloader) {
	dir := filepath.Dir(entry)
	fingerprint := func() string {
		var names []string
		matches, err := filepath.Glob(filepath.Join(dir, "*.sui"))
		if err == nil {
			names = append(names, matches...)
		}
		var parts []string
		for _, n := range names {
			fi, serr := os.Stat(n)
			if serr != nil {
				continue
			}
			parts = append(parts, n+":"+fmt.Sprintf("%v:%v", fi.ModTime().UnixNano(), fi.Size()))
		}
		return strings.Join(parts, "|")
	}
	last := fingerprint()
	for {
		time.Sleep(400 * time.Millisecond)
		cur := fingerprint()
		if cur == last {
			continue
		}
		last = cur
		fmt.Printf("%v: change detected, recompiling ...\n", AppName)
		if err := BuildWebInto(entry, out, hr, true); err != nil {
			fmt.Fprintf(os.Stderr, "%v: rebuild failed: %v\n", AppName, err)
			continue
		}
		fmt.Printf("%v: rebuilt ok\n", AppName)
	}
}

func openBrowserCommand(url string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("cmd", "/c", "start", "", url).Run()
	case "darwin":
		_ = exec.Command("open", url).Run()
	default:
		_ = exec.Command("xdg-open", url).Run()
	}
}
