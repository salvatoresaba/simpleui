package sui

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestGuiComponentLibraryE2E loads an external component library, checks that
// it is grouped in the Insert catalog and in the search dialog, and inserts a
// preset with its CSS.
func TestGuiComponentLibraryE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping component library test")
	}
	const source = `view main {
	col class: "page" {
		p "Start"
	}
}`
	const library = `{"ok":true,"source":"components/library.json","groups":[{"id":"custom","label":"Custom widgets","components":[{"id":"demo-card","label":"Demo card","kind":"card","class":"lib-demo","css":".lib-demo { padding: 10px; border-radius: 12px; background: #101820; color: #ffffff; }","children":[{"kind":"p","text":"Hello library"}]}]}]}`
	page := injectGuiTestScript(guiComponentLibraryScript)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}
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
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=20000", "--dump-dom",
		baseURL,
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
	} else if len(text) > 2000 {
		detail = text[len(text)-2000:]
	}
	t.Fatalf("component library e2e failed:\n%s", detail)
}

const guiComponentLibraryScript = `<script>
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
    await until(function () { return typeof renderInsertMenu === 'function' && typeof addLibraryComponent === 'function'; }, 12000);
    await until(function () { return typeof __prefsLoaded !== 'undefined' && __prefsLoaded && root && root.children && root.children.length > 0; }, 15000);
    COMPONENT_LIBRARY = { groups: [{ id: 'custom', label: 'Custom widgets', components: [{ id: 'demo-card', label: 'Demo card', kind: 'card', class: 'lib-demo', css: '.lib-demo { padding: 10px; border-radius: 12px; background: #101820; color: #ffffff; }', children: [{ kind: 'p', text: 'Hello library' }] }] }] };
    assert(COMPONENT_LIBRARY.groups[0].label === 'Custom widgets', 'library group label missing');
    renderInsertMenu();
    var drop = document.getElementById('drop-insert');
    var headers = Array.prototype.map.call(drop.querySelectorAll(':scope > .mi.grp'), function (x) { return x.textContent; });
    assert(headers.indexOf('Component library') >= 0, 'library header missing');
    assert(headers.indexOf('Built-in') >= 0, 'built-in header missing');
    var demo = null;
    drop.querySelectorAll('.mi-sub .mi').forEach(function (x) { if (x.textContent === 'Demo card') demo = x; });
    assert(!!demo, 'library component missing from the Insert menu');
    demo.onclick();
    await wait(700);
    function findByClass(list, cls) {
      for (var i = 0; i < list.length; i++) {
        if (list[i].class === cls) return list[i];
        if (list[i].children) { var x = findByClass(list[i].children, cls); if (x) return x; }
      }
      return null;
    }
    var n = findByClass(root.children, 'lib-demo');
    assert(!!n, 'component was not inserted: ' + JSON.stringify({ roots: root.children.map(function(x){ return x.kind + ':' + x.class + ':' + (x.children ? x.children.length : 0); }), css: root.css.indexOf('.lib-demo') >= 0, lib: COMPONENT_LIBRARY ? COMPONENT_LIBRARY.groups[0].components[0].kind : 'none' }));
    assert(n.kind === 'card' && n.class === 'lib-demo', 'preset kind/class not applied');
    assert(n.children && n.children.length === 1 && n.children[0].text === 'Hello library', 'preset child missing');
    assert(root.css.indexOf('.lib-demo') >= 0, 'preset CSS missing');
    assert(root.css.indexOf('/* sui-library:lib-demo') >= 0, 'managed CSS block missing');
    renderInsertList();
    var items = Array.prototype.map.call(document.querySelectorAll('#insert-list .insert-item'), function (x) { return x.textContent; });
    assert(items.some(function (x) { return x.indexOf('Demo card') >= 0; }), 'library component missing from search dialog');
    assert(items.some(function (x) { return x.indexOf('Button') >= 0; }), 'built-in component missing from search dialog');
    result(true, '');
  } catch (e) {
    result(false, (e && e.message) ? e.message : String(e));
  }
})();
</script>`
