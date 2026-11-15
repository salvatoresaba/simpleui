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

func TestMotionGraphKeyAttr(t *testing.T) {
	graph := map[string]any{
		"initial": "a",
		"states":  []any{map[string]any{"id": "a", "name": "A", "css": "width: 200px;"}},
		"links":   []any{},
	}
	src := "view main {\n\tcard motionKey: \"source-card\" motionGraph: \"" + suiGraphLiteral(graph) + "\" {\n\t\tp \"Graph\"\n\t}\n}\n"
	prog, dl := ParseSource(src, "graph.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `data-sui-motion-key="source-card"`) {
		t.Fatal("motionKey was not emitted for a graph element without motion")
	}
	if strings.Count(js, `data-sui-motion-key="source-card"`) != 1 {
		t.Fatal("motionKey was emitted more than once")
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

const motionGraphLoopsScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function setGraph(el, cfg, key) {
    if (key) el.setAttribute('data-sui-motion-key', key);
    el.setAttribute('data-sui-motion-graph', JSON.stringify(cfg));
    el.__suiGraphJson = '';
    el.__suiGraphBound = false;
    window.__SUI_MOTION_GRAPH.init(el);
  }
  function width(el) { return Math.round(el.getBoundingClientRect().width); }
  try {
    var src = document.querySelector('.source');
    var inf = document.querySelector('.infinite');
    var fin = document.querySelector('.finite');
    var spr = document.querySelector('.spring');
    var followEnd = document.querySelector('.follower-end');
    var followStart = document.querySelector('.follower-start');
    var cancel = document.querySelector('.cancel');
    for (var i = 0; i < 120 && (!src || !inf || !fin || !spr || !followEnd || !followStart || !cancel || !window.__SUI_MOTION_GRAPH); i++) await wait(50);
    if (!src || !inf || !fin || !spr || !followEnd || !followStart || !cancel || !window.__SUI_MOTION_GRAPH) throw new Error('test elements missing');
    if (src.getAttribute('data-sui-motion-key') !== 'source-card') throw new Error('motionKey was not emitted for a graph element');
    var log = [];
    ['sui:motionstart', 'sui:motionend', 'sui:motioncancel', 'custom-end', 'custom-start'].forEach(function (name) {
      src.addEventListener(name, function (e) { log.push({ name: name, detail: e.detail }); });
    });

    var sourceCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #223344;' },
      { id: 'b', css: 'width: 200px; height: 80px; background-color: #446688;' },
      { id: 'c', css: 'width: 260px; height: 100px; background-color: #6688aa;' }
    ], links: [
      { id: 'ab', from: 'a', to: 'b', event: 'click', duration: '300ms', easing: 'linear', emit: { on: 'end', name: 'custom-end' } },
      { id: 'bc', from: 'b', to: 'c', event: 'click', duration: '300ms', easing: 'linear', emit: { on: 'start', name: 'custom-start' } },
      { id: 'ca', from: 'c', to: 'a', event: 'click', duration: '120ms', easing: 'linear' }
    ] };
    setGraph(src, sourceCfg, 'source-card');

    var endCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #332211;' },
      { id: 'b', css: 'width: 140px; height: 80px; background-color: #886644;' }
    ], links: [
      { id: 'end', from: 'a', to: 'b', event: 'custom', customEvent: 'custom-end', sourceId: 'source-card', duration: '80ms', easing: 'linear' }
    ] };
    setGraph(followEnd, endCfg, 'follower-end');
    var startCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #113322;' },
      { id: 'b', css: 'width: 140px; height: 80px; background-color: #448866;' }
    ], links: [
      { id: 'start', from: 'a', to: 'b', event: 'custom', customEvent: 'custom-start', sourceId: 'source-card', duration: '80ms', easing: 'linear' }
    ] };
    setGraph(followStart, startCfg, 'follower-start');
    var cancelCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #222222;' },
      { id: 'b', css: 'width: 150px; height: 70px; background-color: #aa4444;' }
    ], links: [
      { id: 'cancel', from: 'a', to: 'b', event: 'custom', customEvent: 'sui:motioncancel', sourceId: 'source-card', duration: '80ms', easing: 'linear' }
    ] };
    setGraph(cancel, cancelCfg, 'cancel-card');

    src.click();
    await wait(80);
    src.click();
    await wait(450);
    if (src.__suiGraphState !== 'c') throw new Error('interrupted source state: ' + src.__suiGraphState);
    if (src.__suiGraphAnim) throw new Error('interrupted source still has an animation');
    if (!log.some(function (x) { return x.name === 'sui:motioncancel'; })) throw new Error('sui:motioncancel was not emitted');
    if (!log.some(function (x) { return x.name === 'custom-start'; })) throw new Error('emit on start was not dispatched');
    if (log.some(function (x) { return x.name === 'custom-end'; })) throw new Error('interrupted transition emitted its end event');
    if (followStart.__suiGraphState !== 'b') throw new Error('custom-start follower: ' + followStart.__suiGraphState);
    if (followEnd.__suiGraphState !== 'a') throw new Error('custom-end follower fired too early: ' + followEnd.__suiGraphState);
    if (cancel.__suiGraphState !== 'b') throw new Error('sui:motioncancel listener: ' + cancel.__suiGraphState);

    src.click();
    await wait(220);
    src.click();
    await wait(450);
    if (src.__suiGraphState !== 'b') throw new Error('source did not return to b: ' + src.__suiGraphState);
    if (followEnd.__suiGraphState !== 'b') throw new Error('custom-end follower did not fire: ' + followEnd.__suiGraphState);
    if (!log.some(function (x) { return x.name === 'sui:motionend'; })) throw new Error('sui:motionend was not emitted');
    var startDetail = log.filter(function (x) { return x.name === 'sui:motionstart'; })[0];
    if (!startDetail || startDetail.detail.to !== 'b') throw new Error('motionstart detail is wrong');

    var sprCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #223344;' },
      { id: 'b', css: 'width: 200px; height: 90px; background-color: #557799;' }
    ], links: [
      { id: 'spring', from: 'a', to: 'b', event: 'click', duration: '500ms', spring: { stiffness: 220, damping: 8, mass: 1 } }
    ] };
    setGraph(spr, sprCfg, 'spring-card');
    spr.click();
    var springAnim = spr.__suiGraphAnim;
    if (!springAnim) throw new Error('spring animation is missing');
    var sprEase = springAnim.effect.getTiming().easing;
    springAnim.pause();
    springAnim.currentTime = 250;
    await wait(30);
    var springMid = width(spr);
    springAnim.currentTime = 500;
    await wait(30);
    var springEnd = width(spr);
    springAnim.play();
    await wait(650);
    if (springMid <= 205) throw new Error('spring did not overshoot: ' + springMid + ' easing=' + sprEase);
    if (springEnd !== 200) throw new Error('spring end sample: ' + springEnd);
    if (width(spr) !== 200) throw new Error('spring final width: ' + width(spr));

    var api = window.SimpleUIMotion.get(spr);
    if (!api) throw new Error('SimpleUIMotion API is missing');
    if (api.state !== 'b') throw new Error('API state: ' + api.state);
    api.to('a', { duration: '100ms', easing: 'linear' });
    await wait(250);
    if (spr.__suiGraphState !== 'a') throw new Error('API to(a): ' + spr.__suiGraphState);
    api.to('b', { duration: '400ms', easing: 'linear' });
    await wait(30);
    api.pause();
    api.seek(200);
    await wait(30);
    var pausedWidth = width(spr);
    await wait(120);
    if (Math.abs(width(spr) - pausedWidth) > 2) throw new Error('API pause did not stop the transition');
    if (!(api.progress > 0 && api.progress < 1)) throw new Error('API progress while paused: ' + api.progress);
    api.seek(399);
    api.resume();
    await wait(200);
    if (spr.__suiGraphState !== 'b') throw new Error('API resume state: ' + spr.__suiGraphState);
    if (width(spr) !== 200) throw new Error('API resume width: ' + width(spr));

    api.to('a', { duration: '400ms', easing: 'linear' });
    await wait(80);
    api.pause();
    api.seek(200);
    if (!(api.progress > 0.35)) throw new Error('API seek progress: ' + api.progress);
    api.resume();
    await wait(320);
    if (spr.__suiGraphState !== 'a') throw new Error('API seek final state: ' + spr.__suiGraphState);

    var finCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #223344;' },
      { id: 'b', css: 'width: 180px; height: 90px; background-color: #557799;' },
      { id: 'c', css: 'width: 220px; height: 100px; background-color: #7799bb;' }
    ], links: [
      { id: 'loop', from: 'a', to: 'b', event: 'click', duration: '100ms', easing: 'linear', iterations: 3, direction: 'alternate', playbackRate: 2 },
      { id: 'after', from: 'b', to: 'c', event: 'after', duration: '100ms', easing: 'linear' }
    ] };
    setGraph(fin, finCfg, 'finite-card');
    var finEnds = 0;
    fin.addEventListener('sui:motionend', function () { finEnds++; });
    fin.click();
    if (!fin.__suiGraphAnim || fin.__suiGraphAnim.playbackRate !== 2) throw new Error('finite playback rate was not applied');
    await wait(700);
    if (fin.__suiGraphState !== 'c') throw new Error('finite loop did not chain: ' + fin.__suiGraphState);
    if (width(fin) !== 220) throw new Error('finite final width: ' + width(fin));
    if (fin.__suiGraphAnim) throw new Error('finite animation was not cleaned up');
    if (finEnds !== 2) throw new Error('finite loop end count: ' + finEnds);

    var infCfg = { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 60px; background-color: #223344;' },
      { id: 'b', css: 'width: 180px; height: 90px; background-color: #557799;' }
    ], links: [
      { id: 'start', from: 'a', to: 'b', event: 'click', duration: '180ms', easing: 'linear', iterations: 'infinite', direction: 'alternate' },
      { id: 'stop', from: 'b', to: 'a', event: 'click', duration: '100ms', easing: 'linear' }
    ] };
    setGraph(inf, infCfg, 'infinite-card');
    var infEnds = 0;
    inf.addEventListener('sui:motionend', function () { infEnds++; });
    inf.click();
    await wait(80);
    if (inf.__suiGraphState !== 'b') throw new Error('infinite target state: ' + inf.__suiGraphState);
    if (!inf.__suiGraphAnim || !inf.__suiGraphInfinite) throw new Error('infinite animation is not active');
    var infTiming = inf.__suiGraphAnim.effect.getTiming();
    if (infTiming.iterations !== Infinity) throw new Error('infinite iterations were not applied');
    if (infTiming.direction !== 'alternate') throw new Error('infinite direction was not applied');
    inf.__suiGraphAnim.currentTime = 90;
    await wait(30);
    var midW = width(inf);
    inf.__suiGraphAnim.currentTime = 180;
    await wait(30);
    var endW = width(inf);
    inf.__suiGraphAnim.currentTime = 270;
    await wait(30);
    var backW = width(inf);
    if (!(midW > 105 && midW < 180)) throw new Error('infinite forward sample: ' + midW);
    if (Math.abs(endW - 180) > 2) throw new Error('infinite end sample: ' + endW);
    if (!(backW > 105 && backW < 180)) throw new Error('infinite reverse sample: ' + backW);
    await wait(250);
    if (infEnds !== 0) throw new Error('infinite animation emitted end');
    inf.click();
    await wait(200);
    if (inf.__suiGraphState !== 'a') throw new Error('infinite stop state: ' + inf.__suiGraphState);
    if (inf.__suiGraphAnim || inf.__suiGraphInfinite) throw new Error('infinite animation was not cleaned up');
    if (width(inf) !== 100) throw new Error('infinite final width: ' + width(inf));
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`

func TestMotionGraphLoopsAndEventsE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion graph test")
	}
	graph := map[string]any{
		"initial": "a",
		"states": []any{
			map[string]any{"id": "a", "name": "A", "css": "width: 100px; height: 60px; background-color: #223344;"},
			map[string]any{"id": "b", "name": "B", "css": "width: 200px; height: 80px; background-color: #446688;"},
		},
		"links": []any{},
	}
	source := "css \".box \\{ width: 100px; height: 60px; background: #223344; \\}\"\n" +
		"view main {\n" +
		"\tcard class: \"box source\" motionKey: \"source-card\" motionGraph: \"" + suiGraphLiteral(graph) + "\" { p \"Source\" }\n" +
		"\tcard class: \"box infinite\" motionKey: \"infinite-card\" { p \"Infinite\" }\n" +
		"\tcard class: \"box finite\" motionKey: \"finite-card\" { p \"Finite\" }\n" +
		"\tcard class: \"box spring\" motionKey: \"spring-card\" { p \"Spring\" }\n" +
		"\tcard class: \"box follower-end\" motionKey: \"follower-end\" { p \"Follower end\" }\n" +
		"\tcard class: \"box follower-start\" motionKey: \"follower-start\" { p \"Follower start\" }\n" +
		"\tcard class: \"box cancel\" motionKey: \"cancel-card\" { p \"Cancel\" }\n" +
		"}\n"
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + motionGraphLoopsScript + page[idx:]
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=1000,700",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=30000", "--dump-dom", url,
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
		end := marker + 2000
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("motion graph loops/events e2e failed:\n%s", detail)
}

const motionGraphGuardsGesturesExitScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  var graphSeq = 0;
  function setGraph(el, cfg, key) {
    if (key) el.setAttribute('data-sui-motion-key', key + '-' + (++graphSeq));
    el.setAttribute('data-sui-motion-graph', JSON.stringify(cfg));
    el.__suiGraphJson = '';
    window.__SUI_MOTION_GRAPH.init(el);
  }
  function baseCfg(links) {
    return { initial: 'a', states: [
      { id: 'a', css: 'width: 100px; height: 80px; background-color: #223344;' },
      { id: 'b', css: 'width: 180px; height: 100px; background-color: #557799;' }
    ], links: links };
  }
  function swipe(el, x1, y1, x2, y2) {
    el.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, cancelable: true, pointerId: 7, pointerType: 'mouse', button: 0, buttons: 1, clientX: x1, clientY: y1 }));
    document.dispatchEvent(new PointerEvent('pointerup', { bubbles: true, cancelable: true, pointerId: 7, pointerType: 'mouse', button: 0, buttons: 0, clientX: x2, clientY: y2 }));
  }
  try {
    var guard = document.querySelector('.guard-card');
    var wheel = document.querySelector('.wheel-card');
    var swipeCard = document.querySelector('.swipe-card');
    var hold = document.querySelector('.hold-card');
    var vis = document.querySelector('.vis-card');
    var toggle = document.querySelector('.toggle');
    var exitRoot = document.querySelector('.exit-root');
    var exitCard = document.querySelector('.exit-card');
    for (var i = 0; i < 120 && (!guard || !wheel || !swipeCard || !hold || !vis || !toggle || !exitRoot || !exitCard || !window.__SUI_MOTION_GRAPH); i++) await wait(50);
    if (!guard || !wheel || !swipeCard || !hold || !vis || !toggle || !exitRoot || !exitCard || !window.__SUI_MOTION_GRAPH) throw new Error('test elements missing');
    if (!window.__SUI_STATE_GET || !window.__SUI_STATE_GET.allowed) throw new Error('SUI state getter missing');

    setGraph(guard, baseCfg([
      { id: 'g1', from: 'a', to: 'b', event: 'click', guard: 'allowed && mode == "edit"', duration: '80ms', easing: 'linear' }
    ]), 'guard-card');
    guard.click(); await wait(180);
    if (guard.__suiGraphState !== 'a') throw new Error('guard should block while allowed is false');
    toggle.click(); await wait(120);
    if (!window.__SUI_STATE_GET.allowed()) throw new Error('toggle did not set allowed');
    // A state change re-renders the view, so the original nodes are detached.
    // Re-query before exercising layout-dependent gestures.
    guard = document.querySelector('.guard-card');
    wheel = document.querySelector('.wheel-card');
    swipeCard = document.querySelector('.swipe-card');
    hold = document.querySelector('.hold-card');
    vis = document.querySelector('.vis-card');
    exitRoot = document.querySelector('.exit-root');
    exitCard = document.querySelector('.exit-card');
    if (!guard || !wheel || !swipeCard || !hold || !vis || !exitRoot || !exitCard) throw new Error('test elements missing after re-render');
    guard.click(); await wait(220);
    if (guard.__suiGraphState !== 'b') throw new Error('guard did not allow the transition');
    setGraph(guard, baseCfg([
      { id: 'g2', from: 'a', to: 'b', event: 'click', guard: 'allowed &&', duration: '80ms', easing: 'linear' }
    ]), 'guard-card');
    guard.click(); await wait(180);
    if (guard.__suiGraphState !== 'a') throw new Error('invalid guard should not fire');

    setGraph(wheel, baseCfg([
      { id: 'w1', from: 'a', to: 'b', event: 'wheel', wheel: 'down', duration: '80ms', easing: 'linear' }
    ]), 'wheel-card');
    wheel.dispatchEvent(new WheelEvent('wheel', { bubbles: true, cancelable: true, deltaY: 120 }));
    await wait(220);
    if (wheel.__suiGraphState !== 'b') throw new Error('wheel down did not fire');
    setGraph(wheel, baseCfg([
      { id: 'w2', from: 'a', to: 'b', event: 'wheel', wheel: 'down', duration: '80ms', easing: 'linear' }
    ]), 'wheel-card');
    wheel.dispatchEvent(new WheelEvent('wheel', { bubbles: true, cancelable: true, deltaY: -120 }));
    await wait(180);
    if (wheel.__suiGraphState !== 'a') throw new Error('wheel up should not match down');

    setGraph(swipeCard, baseCfg([
      { id: 's1', from: 'a', to: 'b', event: 'swipe', axis: 'right', distance: 40, duration: '80ms', easing: 'linear' }
    ]), 'swipe-card');
    swipe(swipeCard, 20, 20, 120, 20);
    await wait(220);
    if (swipeCard.__suiGraphState !== 'b') throw new Error('swipe right did not fire');
    setGraph(swipeCard, baseCfg([
      { id: 's2', from: 'a', to: 'b', event: 'swipe', axis: 'right', distance: 40, duration: '80ms', easing: 'linear' }
    ]), 'swipe-card');
    swipe(swipeCard, 120, 20, 20, 20);
    await wait(180);
    if (swipeCard.__suiGraphState !== 'a') throw new Error('swipe left should not match right');

    setGraph(hold, baseCfg([
      { id: 'h1', from: 'a', to: 'b', event: 'longpress', hold: 120, duration: '80ms', easing: 'linear' }
    ]), 'hold-card');
    hold.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, pointerId: 8, pointerType: 'mouse', button: 0, buttons: 1, clientX: 20, clientY: 20 }));
    await wait(220);
    document.dispatchEvent(new PointerEvent('pointerup', { bubbles: true, pointerId: 8, pointerType: 'mouse', button: 0, buttons: 0, clientX: 20, clientY: 20 }));
    if (hold.__suiGraphState !== 'b') throw new Error('long press did not fire');
    setGraph(hold, baseCfg([
      { id: 'h2', from: 'a', to: 'b', event: 'longpress', hold: 180, duration: '80ms', easing: 'linear' }
    ]), 'hold-card');
    hold.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, pointerId: 9, pointerType: 'mouse', button: 0, buttons: 1, clientX: 20, clientY: 20 }));
    await wait(60);
    document.dispatchEvent(new PointerEvent('pointerup', { bubbles: true, pointerId: 9, pointerType: 'mouse', button: 0, buttons: 0, clientX: 20, clientY: 20 }));
    await wait(200);
    if (hold.__suiGraphState !== 'a') throw new Error('short long press should not fire');

    setGraph(vis, baseCfg([
      { id: 'v1', from: 'a', to: 'b', event: 'visible', threshold: 0.5, once: true, duration: '80ms', easing: 'linear' }
    ]), 'vis-card');
    await wait(900);
    if (vis.__suiGraphState !== 'b') throw new Error('visible trigger did not fire');

    // Deterministic threshold/once/hidden test: hide the card so the initial
    // viewport check sees ratio 0, then feed synthetic IntersectionObserver
    // entries. Headless virtual time does not deliver later observer frames.
    vis.style.display = 'none';
    setGraph(vis, baseCfg([
      { id: 'v2', from: 'a', to: 'b', event: 'visible', threshold: 0.6, once: true, duration: '60ms', easing: 'linear' },
      { id: 'v3', from: 'b', to: 'a', event: 'hidden', threshold: 0.5, duration: '60ms', easing: 'linear' }
    ]), 'vis-card');
    await wait(120);
    if (typeof vis.__suiGraphApplyVisibility !== 'function') throw new Error('visibility evaluator hook missing');
    vis.__suiGraphApplyVisibility({ isIntersecting: true, intersectionRatio: 0.4 });
    await wait(120);
    if (vis.__suiGraphState !== 'a') throw new Error('visible fired below threshold');
    vis.__suiGraphApplyVisibility({ isIntersecting: true, intersectionRatio: 0.8 });
    await wait(160);
    if (vis.__suiGraphState !== 'b') throw new Error('visible did not fire above threshold');
    vis.__suiGraphApplyVisibility({ isIntersecting: false, intersectionRatio: 0.2 });
    await wait(160);
    if (vis.__suiGraphState !== 'a') throw new Error('hidden trigger did not fire');
    vis.__suiGraphApplyVisibility({ isIntersecting: true, intersectionRatio: 0.9 });
    await wait(160);
    if (vis.__suiGraphState !== 'a') throw new Error('once visibility link fired twice');

    setGraph(exitCard, {
      initial: 'a',
      states: [
        { id: 'a', css: 'width: 100px; height: 80px; background-color: #223344;' },
        { id: 'gone', css: 'width: 160px; height: 100px; opacity: 0; background-color: #aa4444;' }
      ],
      links: [],
      exit: { state: 'gone', duration: '200ms', easing: 'linear' }
    }, 'exit-card');
    window.__SUI_MOTION_GRAPH.beforeReplace(exitRoot);
    exitRoot.innerHTML = '<p>gone</p>';
    window.__SUI_MOTION_GRAPH.afterReplace(exitRoot);
    await wait(70);
    var clone = document.body.querySelector('.sui-motion-graph-exit');
    if (!clone) throw new Error('exit clone was not created');
    var exitAnims = clone.getAnimations ? clone.getAnimations() : [];
    if (!exitAnims.length) throw new Error('exit clone has no animation');
    var frames = exitAnims[0].effect.getKeyframes();
    var lastOpacity = frames.length ? String(frames[frames.length - 1].opacity) : '';
    if (parseFloat(lastOpacity) !== 0) throw new Error('exit animation does not target transparent: ' + lastOpacity);
    // Headless virtual time does not advance animation frames after load, so
    // seek the animation explicitly before reading the computed style.
    exitAnims[0].currentTime = 100;
    await wait(20);
    var co = parseFloat(getComputedStyle(clone).opacity);
    if (!(co < 0.98)) throw new Error('exit clone did not animate opacity: ' + co);
    await wait(300);
    if (document.body.querySelector('.sui-motion-graph-exit')) throw new Error('exit clone was not removed');

    // Real SUI removal: toggling a SUI state makes __sui_replace_view drop the
    // conditional card, so the runtime hooks must create the exit clone.
    var liveToggle = document.querySelector('.exit-toggle');
    var liveCard = document.querySelector('.exit-live');
    if (!liveToggle || !liveCard) throw new Error('live exit elements missing');
    liveToggle.click();
    await wait(120);
    if (document.querySelector('.exit-live[data-sui-id]')) throw new Error('SUI did not remove the conditional card');
    var liveClone = document.body.querySelector('.sui-motion-graph-exit');
    if (!liveClone) throw new Error('SUI removal did not create an exit clone');
    await wait(350);
    if (document.body.querySelector('.sui-motion-graph-exit')) throw new Error('SUI exit clone was not removed');
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`

func TestMotionGraphGuardsGesturesExitE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion graph test")
	}
	exitGraph := `{"initial":"shown","states":[{"id":"shown","css":"width: 100px; height: 80px; background-color: #223344;"},{"id":"gone","css":"width: 100px; height: 80px; opacity: 0;"}],"links":[],"exit":{"state":"gone","duration":"200ms","easing":"linear"}}`
	exitAttr := strings.NewReplacer("{", "\\{", "}", "\\}", "\"", "\\\"").Replace(exitGraph)
	source := "css \".box \\{ width: 100px; height: 80px; background: #223344; margin: 4px; \\} .exit-root \\{ width: 300px; height: 140px; \\}\"\n" +
		"state allowed = false\n" +
		"state mode = \"edit\"\n" +
		"state showExit = true\n" +
		"fn toggleAllowed() -> null { allowed = not allowed }\n" +
		"fn toggleExit() -> null { showExit = not showExit }\n" +
		"view main {\n" +
		"\tbutton \"Toggle\" class: \"toggle\" onClick: toggleAllowed\n" +
		"\tbutton \"Exit toggle\" class: \"exit-toggle\" onClick: toggleExit\n" +
		"\tcard class: \"box guard-card\" { p \"Guard\" }\n" +
		"\tcard class: \"box wheel-card\" { p \"Wheel\" }\n" +
		"\tcard class: \"box swipe-card\" { p \"Swipe\" }\n" +
		"\tcard class: \"box hold-card\" { p \"Hold\" }\n" +
		"\tcard class: \"box vis-card\" { p \"Visible\" }\n" +
		"\tdiv class: \"exit-root\" { card class: \"box exit-card\" { p \"Exit\" } }\n" +
		"\tif showExit { card class: \"box exit-live\" motionGraph: \"" + exitAttr + "\" { p \"Exit live\" } }\n" +
		"}\n"
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + motionGraphGuardsGesturesExitScript + page[idx:]
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
		end := marker + 2000
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("motion guards/gestures/exit e2e failed:\n%s", detail)
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
