package main

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestGuiAppearanceLibraryE2E checks the per-control appearance section
// (structure/style/animation/theme), its live effect on the selected control,
// the appearance navigator preview and saving a favourite theme.
func TestGuiAppearanceLibraryE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping appearance library test")
	}
	const source = `view main {
	button "Go" class: "go"
}`
	page := injectGuiTestScript(guiAppearanceScript)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	var lastHTML string
	mux := guiTestMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	mux.HandleFunc("GET /__gui__/components", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"source":"","groups":[]}`)
	})
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(source))+`}`)
	})
	mux.HandleFunc("GET /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{}`)
	})
	mux.HandleFunc("POST /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	mux.HandleFunc("POST /__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ok, html, msgs := renderSui(r.PostFormValue("sui"))
		if ok {
			lastHTML = html
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
	})
	mux.HandleFunc("POST /__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		model, msgs, ok := guiOpen(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		body := StringifyJson(model)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,`+body[1:])
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", lastHTML)
	})
	mux.HandleFunc("POST /__css__/parse", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssParse(r.PostFormValue("css"), r.PostFormValue("used"))))
	})
	mux.HandleFunc("GET /__gui__/fonts", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"fonts":[]}`)
	})
	go func() { _ = http.Serve(ln, mux) }()
	baseURL := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=1200,900",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=25000", "--dump-dom", baseURL,
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
		end := marker + 2200
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("appearance library e2e failed:\n%s", detail)
}

const guiAppearanceScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function until(fn, timeout) {
    var end = Date.now() + (timeout || 5000);
    return new Promise(function (resolve) {
      (function loop() {
        if (fn()) return resolve(true);
        if (Date.now() > end) return resolve(false);
        setTimeout(loop, 40);
      })();
    });
  }
  function assert(ok, message) { if (!ok) throw new Error(message); }
  function appSection() {
    return Array.prototype.slice.call(document.querySelectorAll('#inspector .acc')).filter(function (s) { return s.title === 'Control appearance'; })[0] || null;
  }
  function selects() {
    var sec = appSection();
    return sec ? Array.prototype.slice.call(sec.querySelectorAll('select')) : [];
  }
  try {
    await until(function () { return typeof __prefsLoaded !== 'undefined' && __prefsLoaded && root && root.children.length > 0; }, 15000);
    selPath = [0]; multiSel = [];
    renderTree(); renderInspector();
    await until(function () { return !!appSection(); }, 4000);
    var s = selects();
    assert(s.length === 4, 'expected structure, style, animation and theme selects');
    var structs = Array.prototype.slice.call(s[0].options).map(function (o) { return o.textContent; });
    assert(structs.indexOf('Circular') >= 0, 'button structures missing');
    s[0].value = 'circle'; s[0].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(500);
    var btn = byPath(selPath, root.children);
    assert(String(btn.style || '').indexOf('width:56px') >= 0 || String(btn.style || '').indexOf('width: 56px') >= 0, 'circular structure not applied');
    s = selects(); s[1].value = 'coral'; s[1].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(500);
    btn = byPath(selPath, root.children);
    assert(String(btn.style || '').indexOf('#f43f5e') >= 0, 'coral style not applied');
    s = selects();
    var pulse = Array.prototype.slice.call(s[2].options).filter(function (o) { return o.textContent === 'Pulse'; })[0];
    assert(!!pulse, 'pulse animation missing');
    s[2].value = pulse.value; s[2].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(500);
    btn = byPath(selPath, root.children);
    assert(/sui-button-pulse/.test(String(btn.style || '')), 'pulse animation not applied');
    // The saved theme must appear in both the inspector and the navigator.
    var oldAsk = window.askPrompt;
    window.askPrompt = function (title, label, initial, cb) { cb('Test theme'); };
    var def = componentDefForKind('button');
    appearanceSaveFavorite(def, 'button', btn, appearanceCurrent(btn, def, 'button'));
    window.askPrompt = oldAsk;
    await wait(400);
    assert(appearanceFavorites().length === 1, 'theme was not saved');
    s = selects();
    var themeLabels = Array.prototype.slice.call(s[3].options).map(function (o) { return o.textContent; });
    assert(themeLabels.join(' ').indexOf('Test theme') >= 0, 'saved theme missing from the inspector');
    openLibraryNavigator('appearance', 'button');
    await until(function () { return document.querySelectorAll('#lib-kinds .lib-group').length > 0; }, 4000);
    var kinds = Array.prototype.slice.call(document.querySelectorAll('#lib-kinds .lib-group')).map(function (b) { return b.textContent; });
    assert(kinds[0] === 'Button', 'button should be first in the appearance kinds');
    var opts = Array.prototype.slice.call(document.querySelectorAll('#lib-structures .lib-app-option')).map(function (b) { return b.textContent; });
    assert(opts.join(' ').indexOf('Circular') >= 0, 'navigator structures missing');
    await wait(900);
    var frame = document.getElementById('lib-app-preview');
    assert(frame.srcdoc.indexOf('Button') >= 0, 'appearance preview was not compiled');
    try { localStorage.removeItem('sui.appearance.themes.v1'); } catch (e) {}
    result(true, '');
  } catch (e) {
    try { localStorage.removeItem('sui.appearance.themes.v1'); } catch (e2) {}
    result(false, (e && e.message) ? e.message : String(e));
  }
})();
</script>`
