package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func suiGraphLiteral(v any) string {
	b, _ := json.Marshal(v)
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, `{`, `\{`, `}`, `\}`)
	return r.Replace(string(b))
}

func TestMotionGraphAttrs(t *testing.T) {
	graph := map[string]any{
		"initial": "a",
		"states": []any{
			map[string]any{"id": "a", "name": "A", "css": "width: 200px;"},
			map[string]any{"id": "b", "name": "B", "css": "width: 360px;"},
		},
		"links": []any{
			map[string]any{"id": "l1", "from": "a", "to": "b", "event": "click", "duration": "320ms"},
		},
	}
	src := "view main {\n\tcard motionGraph: \"" + suiGraphLiteral(graph) + "\" {\n\t\tp \"Graph\"\n\t}\n}\n"
	prog, dl := ParseSource(src, "graph.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "data-sui-motion-graph=") {
		t.Fatal("motionGraph attribute was not emitted")
	}
	if !strings.Contains(js, "__SUI_MOTION_GRAPH") {
		t.Fatal("motion graph runtime was not emitted")
	}
}

func TestMotionGraphKeepsLayoutMotion(t *testing.T) {
	graph := map[string]any{
		"initial": "a",
		"states":  []any{map[string]any{"id": "a", "name": "A", "css": "width: 200px;"}},
		"links":   []any{},
	}
	src := "view main {\n\tcard motion: \"fade\" motionGraph: \"" + suiGraphLiteral(graph) + "\" {\n\t\tp \"Graph\"\n\t}\n}\n"
	prog, dl := ParseSource(src, "graph.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "data-sui-motion-graph=") {
		t.Fatal("motionGraph attribute was not emitted")
	}
	if !strings.Contains(js, "data-sui-motion-effect=") {
		t.Fatal("layout/shared motion must remain available alongside a graph")
	}
}

const motionGraphScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  try {
    var el = null;
    for (var i = 0; i < 100; i++) {
      el = document.querySelector('[data-sui-motion-graph]');
      if (el) break;
      await wait(50);
    }
    if (!el) throw new Error('graph element missing');
    var before = Math.round(el.getBoundingClientRect().width);
    if (before !== 200) throw new Error('unexpected initial width ' + before);
    el.click();
    await wait(140);
    var mid = Math.round(el.getBoundingClientRect().width);
    if (!(mid > 210 && mid < 350)) throw new Error('width did not tween: ' + mid);
    await wait(420);
    var end = Math.round(el.getBoundingClientRect().width);
    if (end !== 360) throw new Error('final width ' + end);
    if (el.__suiGraphState !== 'b') throw new Error('graph state ' + el.__suiGraphState);
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`

func TestMotionGraphE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion graph test")
	}
	graph := map[string]any{
		"initial": "a",
		"states": []any{
			map[string]any{"id": "a", "name": "A", "css": "width: 200px; height: 120px; background-color: #223344;"},
			map[string]any{"id": "b", "name": "B", "css": "width: 360px; height: 200px; margin-left: 80px; background-color: #446688;"},
		},
		"links": []any{
			map[string]any{"id": "l1", "from": "a", "to": "b", "event": "click", "duration": "320ms", "easing": "linear"},
		},
	}
	source := "css \".box \\{ width: 200px; height: 120px; background: #223344; \\}\"\n" +
		"view main {\n\tcard class: \"box\" motion: \"layout\" motionKey: \"graph\" motionGraph: \"" + suiGraphLiteral(graph) + "\" {\n\t\tp \"Click\"\n\t}\n}\n"
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + motionGraphScript + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	go func() { _ = http.Serve(ln, mux) }()
	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=1000,700",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=20000", "--dump-dom", url,
	}
	out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
	text := string(out)
	if err != nil {
		t.Fatalf("headless browser failed: %v\n%s", err, text)
	}
	if strings.Contains(text, `id="e2e-result">OK<`) {
		return
	}
	marker := strings.Index(text, `id="e2e-result"`)
	detail := text
	if marker >= 0 {
		end := marker + 1200
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("motion graph e2e failed:\n%s", detail)
}
