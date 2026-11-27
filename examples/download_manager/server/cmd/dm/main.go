// cmd/dm — the standalone download manager of the SimpleUI example.
//
// The program serves the compiled app and the download manager together: the
// manager is the library of the module (downloadmanager), the app comes from
// the disk (--web-dir) or from the executable itself, when the assets were
// embedded at build time.
//
//	go run ./cmd/dm --port 8090 --web-dir ../public
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"

	"downloadmanager"
)

// embeddedWeb holds the compiled SimpleUI app: the release build carries the
// app inside the executable. The placeholder README.txt keeps `go build`
// working before the first frontend build.
//
//go:embed all:web
var embeddedWeb embed.FS

func main() {
	cfg := downloadmanager.ParseConfig(os.Args[1:])
	// The frontend comes from the disk (--web-dir) or from the executable
	// itself, when the app was embedded at build time.
	var webFS fs.FS
	if cfg.WebDir == "" {
		sub, err := fs.Sub(embeddedWeb, "web")
		if err != nil {
			fmt.Fprintln(os.Stderr, "embedded web assets:", err)
			os.Exit(1)
		}
		webFS = sub
	}
	downloadmanager.Serve(cfg, frontendHandler(cfg.WebDir, webFS))
}

// frontendHandler serves the compiled app: from the disk when a directory is
// given (the development launcher) and from the executable itself when the
// app was embedded at build time (the release build).
func frontendHandler(webDir string, webFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		// The path is cleaned before it is used, so it can never escape the
		// directory of the app.
		rel := strings.TrimPrefix(pathpkg.Clean("/"+path), "/")
		var data []byte
		var err error
		if webFS != nil {
			data, err = fs.ReadFile(webFS, rel)
		} else {
			data, err = os.ReadFile(filepath.Join(webDir, filepath.FromSlash(rel)))
		}
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentTypeOf(path))
		_, _ = w.Write(data)
	}
}

// contentTypeOf is the type of an asset of the app.
func contentTypeOf(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	}
	return "application/octet-stream"
}
