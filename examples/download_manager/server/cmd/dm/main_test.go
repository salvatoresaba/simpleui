package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The release build carries the app inside the executable: the frontend is
// served from the embedded files, with no directory next to it.
func TestTheEmbeddedFrontendServesTheApp(t *testing.T) {
	sub, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		t.Fatalf("embedded web assets: %v", err)
	}
	rec := httptest.NewRecorder()
	frontendHandler("", sub).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("the app was not served: %v", rec.Code)
	}
	page := rec.Body.String()
	if !strings.Contains(page, "<title>Download manager</title>") {
		t.Fatalf("the page is not the app: %v", page[:min(len(page), 120)])
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("wrong content type: %v", ct)
	}
}

// The script of the app is served too: the page alone cannot work without it.
func TestTheEmbeddedFrontendServesTheScript(t *testing.T) {
	sub, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		t.Fatalf("embedded web assets: %v", err)
	}
	rec := httptest.NewRecorder()
	frontendHandler("", sub).ServeHTTP(rec, httptest.NewRequest("GET", "/app.js", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("the script was not served: %v", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "__sui_render") {
		t.Fatal("the script is not the compiled app")
	}
}

// The frontend comes from the disk when a directory is given: the
// development launcher serves public/ from the disk.
func TestTheFrontendIsServedFromTheDisk(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<title>From disk</title>"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	rec := httptest.NewRecorder()
	frontendHandler(dir, nil).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("the app was not served: %v", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "From disk") {
		t.Fatalf("the page is not the one of the directory: %v", rec.Body.String())
	}
}

// A path cannot escape the directory of the app.
func TestTheFrontendRefusesToLeaveItsDirectory(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(filepath.Dir(dir), "secret.txt")
	if err := os.WriteFile(secret, []byte("secret"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(secret) })
	rec := httptest.NewRecorder()
	frontendHandler(dir, nil).ServeHTTP(rec, httptest.NewRequest("GET", "/../secret.txt", nil))
	if rec.Code == http.StatusOK {
		t.Fatal("the path left the directory of the app")
	}
}
