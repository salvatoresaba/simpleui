package main

import (
	"bufio"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// serve_test.go - `simpleui serve` dev server: asset building, the live-
// reload injection and the SSE stream.

func TestServeBuildAssets(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "app.sui")
	src := "state x = 1\nview main {\n\th1 \"{x}\"\n}\n"
	if err := os.WriteFile(entry, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	hr := NewHotReloader()
	if err := BuildWebInto(entry, dir, hr, true); err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, f := range []string{"index.html", "style.css", "app.js"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing asset %v", f)
		}
	}
}

func TestServeHotReloadEndpoints(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "app.sui")
	src := "view main {\n\tp \"hi\"\n}\n"
	if err := os.WriteFile(entry, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	hr := NewHotReloader()
	if err := BuildWebInto(entry, dir, hr, true); err != nil {
		t.Fatalf("build: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		_ = http.Serve(ln, appServer(dir, true, hr))
	}()
	base := "http://" + ln.Addr().String()

	// index.html carries the injected live-reload script.
	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatalf("get /: %v", err)
	}
	if resp.Header.Get("X-Frame-Options") != "DENY" ||
		!strings.Contains(resp.Header.Get("Content-Security-Policy"), "frame-ancestors 'none'") ||
		resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing clickjacking headers: %v", resp.Header)
	}
	html := readAllClose(t, resp)
	if !strings.Contains(html, `id="app"`) || !strings.Contains(html, "/__sui__/reload.js") {
		t.Errorf("index missing #app or reload script:\n%v", html[:300])
	}

	// reload.js exists and uses EventSource.
	js, err := http.Get(base + "/__sui__/reload.js")
	if err != nil {
		t.Fatalf("get reload.js: %v", err)
	}
	if !strings.Contains(readAllClose(t, js), "EventSource") {
		t.Errorf("reload.js must use EventSource")
	}

	// SSE stream pushes an initial reload event.
	sse, err := http.Get(base + "/__sui__/reload")
	if err != nil {
		t.Fatalf("get reload stream: %v", err)
	}
	defer sse.Body.Close()
	rd := bufio.NewReader(sse.Body)
	lineCh := make(chan string, 1)
	go func() {
		l, _ := rd.ReadString('\n')
		lineCh <- l
	}()
	select {
	case line := <-lineCh:
		if !strings.Contains(line, "reload") {
			t.Errorf("SSE initial line = %q, want a reload event", line)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("no SSE reload event received")
	}
}
