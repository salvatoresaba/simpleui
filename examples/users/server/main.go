// usersdemo — demo companion server for the SimpleUI "Users table" example.
//
// It serves two things on one origin:
//
//	GET /            the compiled SimpleUI web app (assets from --web-dir)
//	GET /style.css   ditto
//	GET /app.js      ditto
//	GET /api/users   a JSON data feed the SUI app renders as a table
//
// Because the app and the API live on the same origin, `http.get("/api/users")`
// needs no CORS configuration. Run it after building the app:
//
//	simpleui build steps:
//	   cd <repo>/examples/users
//	   <repo>/simpleui.exe build app.sui --out public --target web
//	then:
//	   cd server && go run . --port 8080 --web-dir ../public
//
// The example ships run.ps1 / run.bat which do both steps and open the
// browser.

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	port := 0
	webDir := filepath.Join("..", "public")
	noOpen := false
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--port" && i+1 < len(os.Args) {
			if p, err := strconv.Atoi(os.Args[i+1]); err == nil && p >= 0 && p < 65536 {
				port = p
			}
		}
		if os.Args[i] == "--web-dir" && i+1 < len(os.Args) {
			webDir = os.Args[i+1]
		}
		if os.Args[i] == "--no-open" {
			noOpen = true
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/users", handleUsers)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		file := filepath.Join(webDir, strings.TrimPrefix(path, "/"))
		data, err := os.ReadFile(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ct := "application/octet-stream"
		if strings.HasSuffix(path, ".html") {
			ct = "text/html; charset=utf-8"
		} else if strings.HasSuffix(path, ".css") {
			ct = "text/css; charset=utf-8"
		} else if strings.HasSuffix(path, ".js") {
			ct = "application/javascript; charset=utf-8"
		}
		w.Header().Set("Content-Type", ct)
		_, _ = w.Write(data)
	})

	addr := "127.0.0.1:0"
	if port != 0 {
		addr = "127.0.0.1:" + strconv.Itoa(port)
	}
	srv := &http.Server{Handler: logRequests(mux)}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
	go func() {
		_ = srv.Serve(ln)
	}()
	url := "http://" + ln.Addr().String() + "/"
	fmt.Println("======================================================")
	fmt.Printf("users API + app running at %v\n", url)
	fmt.Println("Open that address in your browser. Watch these lines:")
	fmt.Println("  - a request from the client appears here as 'GET /api/users'")
	fmt.Println("  - if you never see it, the client is NOT reaching this server")
	fmt.Println("======================================================")
	if !noOpen {
		openBrowser(url)
	}
	select {}

}

// logRequests prints one clear line per request so the user can follow the
// client/server conversation in the server's console window.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		msg := fmt.Sprintf("server: %v %v -> %v", r.Method, r.URL.Path, rec.status)
		if rec.status == 404 {
			msg += "  (not found: is the app served from this same server?)"
		}
		fmt.Println(msg)
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

// handleUsers returns a small JSON array of "rows", one per user. The SUI
// table example expects arrays of [id, name, email].
func handleUsers(w http.ResponseWriter, r *http.Request) {
	// Permit other origins (e.g. the app opened straight from disk or from a
	// separate dev server); same-origin usage is unaffected.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body := `[` +
		`["1","Alice","alice@example.dev"],` +
		`["2","Bob","bob@example.dev"],` +
		`["3","Carol","carol@example.dev"],` +
		`["4","Dan","dan@example.dev"]` +
		`]`
	_, _ = w.Write([]byte(body))
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
