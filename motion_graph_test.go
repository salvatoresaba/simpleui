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

const motionGraphEdgeScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function setGraph(el, cfg) {
    el.setAttribute('data-sui-motion-graph', JSON.stringify(cfg));
    el.__suiGraphJson = '';
    el.__suiGraphBound = false;
    window.__SUI_MOTION_GRAPH.init(el);
  }
  try {
    var main = document.querySelector('.main-card');
    var ext = document.querySelector('.ext-card');
    var other = document.querySelector('.other-card');
    var extstate = document.querySelector('.extstate-card');
    var keys = document.querySelector('.keys-card');
    var src = document.querySelector('.src-btn');
    for (var i = 0; i < 100 && (!main || !ext || !other || !extstate || !keys || !src || !window.__SUI_MOTION_GRAPH); i++) await wait(50);
    if (!main || !ext || !other || !extstate || !keys || !src || !window.__SUI_MOTION_GRAPH) throw new Error('test elements missing');
    var srcId = src.getAttribute('data-sui-id');
    var otherId = other.getAttribute('data-sui-id');
    if (!srcId || !otherId) throw new Error('missing data-sui-id');

    var extCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 200px; height: 100px; background-color: #223344;' },
      { id: 'b', css: 'width: 300px; height: 160px; background-color: #446688;' }
    ], links: [ { id: 'xe', from: 'a', to: 'b', event: 'click', sourceId: srcId, duration: '120ms', easing: 'linear' } ] };
    setGraph(ext, extCfg);
    src.click();
    await wait(300);
    if (ext.__suiGraphState !== 'b') throw new Error('external event did not fire: ' + ext.__suiGraphState);

    var otherCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #553311;' },
      { id: 'b', css: 'width: 140px; height: 80px; background-color: #aa6633;' }
    ], links: [ { id: 'o1', from: 'a', to: 'b', event: 'click', duration: '80ms', easing: 'linear' } ] };
    setGraph(other, otherCfg);
    var esCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 200px; height: 100px; background-color: #223344;' },
      { id: 'b', css: 'width: 280px; height: 140px; background-color: #884422;' }
    ], links: [ { id: 'xs', from: 'a', to: 'b', event: 'elementState', sourceId: otherId, whenState: 'b', duration: '120ms', easing: 'linear' } ] };
    setGraph(extstate, esCfg);
    other.click();
    await wait(400);
    if (extstate.__suiGraphState !== 'b') throw new Error('external state did not fire: ' + extstate.__suiGraphState);

    var mainCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 200px; height: 120px; background-color: #223344;' },
      { id: 'b', css: 'width: 360px; height: 200px; background-color: #446688;' },
      { id: 'c', css: 'width: 300px; height: 150px; transform: translateX(30px); filter: blur(2px); box-shadow: 0 0 12px rgba(0, 0, 0, 0.5);' }
    ], links: [
      { id: 'm1', from: 'a', to: 'b', event: 'click', duration: '320ms', easing: 'linear' },
      { id: 'm2', from: 'b', to: 'c', event: 'click', duration: '320ms', easing: 'linear' }
    ] };
    setGraph(main, mainCfg);
    main.click();
    await wait(100);
    main.click();
    await wait(700);
    if (main.__suiGraphState !== 'c') throw new Error('interrupted state not applied: ' + main.__suiGraphState);
    var leftover = main.style.getPropertyValue('background-color');
    if (leftover) throw new Error('interrupted transition left background-color: ' + leftover + ' props=' + JSON.stringify(main.__suiGraphProps) + ' anim=' + !!main.__suiGraphAnim + ' style=' + main.getAttribute('style'));
    if (Math.round(main.getBoundingClientRect().width) !== 300) throw new Error('interrupted width: ' + main.getBoundingClientRect().width);
    var cs = getComputedStyle(main);
    var matrix = null;
    try { matrix = new DOMMatrix(cs.transform); } catch (e) {}
    if (!matrix || Math.round(matrix.m41) !== 30) throw new Error('transform not applied: ' + cs.transform);
    if (cs.filter.indexOf('blur(2px)') < 0) throw new Error('filter not applied: ' + cs.filter);
    if (cs.boxShadow.indexOf('12px') < 0) throw new Error('box shadow not applied: ' + cs.boxShadow);

    var keyCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 180px; height: 90px; background-color: #224455;' },
      { id: 'b', css: 'width: 220px; height: 120px; background-color: #557799;' }
    ], links: [
      { id: 'k1', from: 'a', to: 'b', event: 'keydown', key: 'Enter', ctrl: true, duration: '80ms', easing: 'linear' },
      { id: 'c1', from: 'b', to: 'a', event: 'custom', customEvent: 'flip', duration: '80ms', easing: 'linear' }
    ] };
    setGraph(keys, keyCfg);
    keys.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', ctrlKey: false, bubbles: true }));
    await wait(150);
    if (keys.__suiGraphState !== 'a') throw new Error('key filter should not fire without Ctrl: ' + keys.__suiGraphState);
    keys.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', ctrlKey: true, bubbles: true }));
    await wait(200);
    if (keys.__suiGraphState !== 'b') throw new Error('key filter did not fire with Ctrl: ' + keys.__suiGraphState);
    keys.dispatchEvent(new Event('flip', { bubbles: true }));
    await wait(200);
    if (keys.__suiGraphState !== 'a') throw new Error('custom event did not fire: ' + keys.__suiGraphState);
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`

func TestMotionGraphEdgeCasesE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion graph test")
	}
	source := "css \".box \\{ width: 200px; height: 120px; background: #223344; \\}\"\n" +
		"view main {\n" +
		"\tbutton \"Source\" class: \"src-btn\"\n" +
		"\tcard class: \"box main-card\" { p \"Main\" }\n" +
		"\tcard class: \"box ext-card\" { p \"Ext\" }\n" +
		"\tcard class: \"box other-card\" { p \"Other\" }\n" +
		"\tcard class: \"box extstate-card\" { p \"ExtState\" }\n" +
		"\tcard class: \"box keys-card\" { p \"Keys\" }\n" +
		"}\n"
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + motionGraphEdgeScript + page[idx:]
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
		end := marker + 1600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("motion graph edge case e2e failed:\n%s", detail)
}

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
