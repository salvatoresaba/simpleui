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

// TestGuiLibraryNavigatorE2E opens the Controls library navigator, filters by
// control type and search text, checks the interactive preview frame, inserts a
// component and closes the navigator.
func TestGuiLibraryNavigatorE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping library navigator test")
	}
	const source = `view main {
	col class: "page" {
		p "Start"
	}
}`
	const library = `{"ok":true,"source":"components/library.json","groups":[{"id":"cards","label":"Cards","components":[{"id":"metric","label":"Metric card","kind":"card","default":true,"class":"lib-metric","css":".lib-metric { padding: 12px; border-radius: 12px; background: #111827; color: #ffffff; }","children":[{"kind":"p","text":"Revenue"}]}]},{"id":"inputs","label":"Inputs","components":[{"id":"search","label":"Search bar","kind":"row","default":true,"class":"lib-search","css":".lib-search { display: flex; gap: 8px; }"}]}]}`
	page := injectGuiTestScript(guiLibraryNavigatorScript)
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
		end := marker + 2000
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("library navigator e2e failed:\n%s", detail)
}

const guiLibraryNavigatorScript = `<script>
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
  function findByClass(list, cls) {
    for (var i = 0; i < list.length; i++) {
      if (list[i].class === cls) return list[i];
      if (list[i].children) { var x = findByClass(list[i].children, cls); if (x) return x; }
    }
    return null;
  }
  try {
    await until(function () { return typeof __prefsLoaded !== 'undefined' && __prefsLoaded && root && root.children.length > 0; }, 15000);
    openLibraryNavigator();
    await until(function () { return document.getElementById('lib-nav').classList.contains('on'); }, 4000);
    await until(function () { return document.querySelectorAll('#lib-list .lib-item').length === 2; }, 4000);
    var groups = Array.prototype.map.call(document.querySelectorAll('#lib-groups .lib-group'), function (b) { return b.textContent; });
    assert(groups.join('|') === 'All (2)|Cards (1)|Inputs (1)', 'unexpected groups: ' + groups.join('|'));
    var search = document.getElementById('lib-search');
    search.value = 'search';
    search.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(300);
    var items = document.querySelectorAll('#lib-list .lib-item');
    assert(items.length === 1 && items[0].textContent.indexOf('Search bar') >= 0, 'search filter failed');
    items[0].click();
    await wait(500);
    assert(document.getElementById('lib-preview').srcdoc.indexOf('Preview') >= 0, 'preview frame was not filled');
    assert(document.getElementById('lib-insert').disabled === false, 'insert button is disabled');
    document.getElementById('lib-insert').click();
    await wait(900);
    var n = findByClass(root.children, 'lib-search');
    assert(!!n, 'inserted library component missing');
    assert(n.class === 'lib-search', 'inserted class missing');
    assert(root.css.indexOf('.lib-search') >= 0, 'inserted css missing');
    closeLibraryNavigator();
    assert(!document.getElementById('lib-nav').classList.contains('on'), 'navigator did not close');
    result(true, '');
  } catch (e) {
    result(false, (e && e.message) ? e.message : String(e));
  }
})();
</script>`
