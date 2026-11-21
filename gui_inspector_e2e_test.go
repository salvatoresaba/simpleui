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

// TestGuiInspectorGuidanceE2E checks the inspector makes a selection
// actionable: primary sections are open, the CSS class is visible, and a class
// carrying an infinite animation is surfaced with a one-click way to stop it.
// It also guards that the removed shape composer and Tools menu stay gone.
func TestGuiInspectorGuidanceE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping inspector guidance test")
	}
	const source = `css ".pulse-demo { padding: 12px; animation: demo-pulse 1.2s linear infinite; } @keyframes demo-pulse { 0%, 100% { opacity: 1; } 50% { opacity: .5; } }"

view main {
	card "Animated card" class: "pulse-demo" {
		p "This card pulses."
	}
}`
	page := injectGuiTestScript(guiInspectorScript)
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
		end := marker + 1600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("inspector guidance e2e failed:\n%s", detail)
}

const guiInspectorScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function until(fn, timeout) {
    var end = Date.now() + (timeout || 4000);
    return new Promise(function (resolve) {
      (function loop() {
        if (fn()) return resolve(true);
        if (Date.now() > end) return resolve(false);
        setTimeout(loop, 40);
      })();
    });
  }
  function assert(ok, message) { if (!ok) throw new Error(message); }
  function section(title) {
    var list = document.querySelectorAll('#inspector .ins-sec, #inspector .acc');
    for (var i = 0; i < list.length; i++) {
      if (list[i].title === title) return list[i];
    }
    return null;
  }
  function buttonIn(root, label) {
    var list = root.querySelectorAll('button');
    for (var i = 0; i < list.length; i++) if (list[i].textContent.trim() === label) return list[i];
    return null;
  }
  try {
    await until(function () { return typeof renderInspector === 'function' && __prefsLoaded && root && root.children.length; }, 15000);
    await until(function () { return typeof findRule === 'function' && !!findRule('pulse-demo'); }, 6000);
    assert(!document.querySelector('#topbar .menu[data-menu="tools"]'), 'Tools menu is still present');
    assert(!document.getElementById('motion-composer'), 'shape composer is still present');
    selPath = [0]; multiSel = [];
    renderTree(); renderInspector();
    await until(function () { return !!document.getElementById('f_class'); }, 4000);
    assert(!!document.querySelector('.ins-toolbar'), 'inspector toolbar missing');
    var app = section('Appearance');
    assert(app && app.classList.contains('open'), 'Appearance is closed by default');
    assert(!!document.getElementById('f_class'), 'class field missing from Appearance');
    var anim = section('Animation');
    assert(!!anim, 'Animation section missing for an animated class');
    assert(anim.textContent.indexOf('pulse-demo') >= 0, 'animated class name missing');
    assert(anim.textContent.indexOf('infinite') >= 0, 'infinite flag missing from the animation summary');
    assert(!!buttonIn(anim, 'Remove class'), 'Remove class button missing');
    assert(!!buttonIn(anim, 'Stop in CSS'), 'Stop in CSS button missing');
    buttonIn(anim, 'Remove class').click();
    await until(function () { return String((root.children[0].class || '')).indexOf('pulse-demo') < 0; }, 4000);
    assert(String((root.children[0].class || '')).indexOf('pulse-demo') < 0, 'Remove class did not remove the class');
    result(true, '');
  } catch (e) {
    result(false, (e && e.message) ? e.message : String(e));
  }
})();
</script>`
