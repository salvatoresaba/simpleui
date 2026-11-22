package core

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestGuiComponentFallbackE2E loads a project whose library reference points to
// a component that no longer exists. The editor must warn and use the default
// component for that kind, updating the class and the managed CSS while keeping
// the element content.
func TestGuiComponentFallbackE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping component fallback test")
	}
	const source = `css "/* sui-library:old-card component=missing-card kind=card */\n.old-card \{ background: #ff0000; \}\n/* /sui-library:old-card */"

view main {
	card "Old" class: "old-card" {
		p "Keep me"
	}
}`
	const library = `{"ok":true,"source":"components/library.json","groups":[{"id":"custom","label":"Custom","components":[{"id":"lib-new","label":"New default","kind":"card","default":true,"class":"lib-new","css":".lib-new { padding: 12px; border-radius: 12px; background: #0ea5e9; color: #ffffff; }"}]}]}`
	page := injectGuiTestScript(guiComponentFallbackScript)
	ln, err := ListenLocal()
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
		writeRawHandler(w, "application/json; charset=utf-8", library)
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=1200,900",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=20000", "--dump-dom", baseURL,
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
		end := marker + 1800
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("component fallback e2e failed:\n%s", detail)
}

const guiComponentFallbackScript = `<script>
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
  try {
    await until(function () { return typeof __prefsLoaded !== 'undefined' && __prefsLoaded && root && root.children.length > 0; }, 15000);
    await until(function () { return typeof LIBRARY_REF_WARNINGS !== 'undefined' && LIBRARY_REF_WARNINGS.length > 0; }, 12000);
    await wait(900);
    var n = root.children[0];
    assert(!!n, 'root card missing');
    assert(n.class === 'lib-new', 'fallback class not applied: ' + n.class);
    assert(n.children && n.children.length === 1 && n.children[0].text === 'Keep me', 'content was not preserved');
    assert(root.css.indexOf('.lib-new') >= 0, 'fallback css missing');
    assert(root.css.indexOf('.old-card') < 0, 'old library css was not replaced');
    assert(root.css.indexOf('component=lib-new') >= 0, 'managed marker was not rewritten');
    assert(LIBRARY_REF_WARNINGS.join(' ').indexOf('missing-card') >= 0, 'missing component was not reported');
    result(true, '');
  } catch (e) {
    result(false, (e && e.message) ? e.message : String(e));
  }
})();
</script>`
