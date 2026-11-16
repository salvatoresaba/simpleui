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

// TestGuiShapeComposerE2E drives the shape & motion composer in a real browser:
// the Tools entry opens it, SVG import creates nodes, booleans and fillets
// change the geometry, and inserting writes masked CSS plus the sui-cad model
// into the project.
func TestGuiShapeComposerE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping shape composer test")
	}
	const source = `view main {
	button "Go" class: "go"
}`
	page := strings.Replace(guiHTML, "</body>", guiShapeComposerScript+"</body>", 1)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	var lastHTML string
	mux := http.NewServeMux()
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
	t.Fatalf("shape composer e2e failed:\n%s", detail)
}

const guiShapeComposerScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
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
  try {
    await until(function () { return typeof openMotionComposer === 'function' && __prefsLoaded; }, 15000);
    assert(!!document.querySelector('#topbar .menu[data-menu="tools"]'), 'Tools menu missing');
    assert(!document.querySelector('#drop-insert .mi[onclick*="openMotionComposer"]'), 'shape editor still in Insert');
    openMotionComposer();
    await until(function () { return document.getElementById('motion-composer').classList.contains('on'); }, 4000);
    assert(__mcPaths.length === 1, 'default box missing');
    assert(document.querySelectorAll('#mc-handles .mc-node').length === 8, 'default box nodes missing');

    mcImportSvgText("<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><rect x='5' y='5' width='40' height='30' rx='6'/><circle cx='70' cy='60' r='20'/></svg>");
    assert(__mcPaths.length === 2, 'SVG import did not create two paths');
    assert(__mcPaths[0].nodes.length === 8, 'rounded rect import lost its arcs');

    __mcPaths = [
      { closed: true, nodes: [mcNode(10,10), mcNode(70,10), mcNode(70,70), mcNode(10,70)] },
      { closed: true, nodes: [mcNode(50,30), mcNode(110,30), mcNode(110,90), mcNode(50,90)] }
    ];
    __mcSelPath = 0; __mcSelNode = -1; __mcAutoBox = false;
    mcBoolean('union');
    assert(__mcPaths.length === 1, 'union did not replace the two paths');
    assert(__mcPaths[0].nodes.length >= 6, 'union lost its outline');

    __mcPaths = [{ closed: true, nodes: [mcNode(10,10), mcNode(110,10), mcNode(110,80), mcNode(10,80)] }];
    __mcSelPath = 0; __mcSelNode = 1;
    mcFillet();
    assert(__mcPaths[0].nodes.length === 5, 'fillet did not add the arc nodes');
    assert(__mcPaths[0].nodes[1].ox !== 0 || __mcPaths[0].nodes[2].ix !== 0, 'fillet has no handles');

    mcSetTool('select');
    __mcSelNode = 1;
    mcCutNode();
    assert(__mcPaths[0].closed === false, 'cut did not open the path');
    mcUndoCad();
    assert(__mcPaths[0].closed === true, 'shape undo did not restore the path');

    var name = document.getElementById('mc-name');
    name.value = 'cad-e2e';
    name.dispatchEvent(new Event('input', { bubbles: true }));
    document.getElementById('mc-bg').value = '#0ea5e9';
    document.getElementById('mc-stroke').value = '#ffffff';
    document.getElementById('mc-sw').value = '3';
    mcInsert();
    await wait(600);
    assert(root.css.indexOf('/* sui-motion-shape:cad-e2e */') >= 0, 'shape block missing from root.css');
    assert(root.css.indexOf('sui-cad:') >= 0, 'shape model metadata missing');
    assert(root.css.indexOf('mask-image') >= 0, 'masked silhouette missing');
    assert(root.css.indexOf('@keyframes sui-motion-cad-e2e') >= 0, 'motion keyframes missing');
    var el = document.querySelector('#prev').contentDocument.querySelector('.cad-e2e');
    assert(!!el, 'inserted control missing from preview');
    assert(getComputedStyle(el).maskImage.indexOf('data:image/svg+xml') >= 0, 'preview mask missing');
    assert(getComputedStyle(el).animationName === 'sui-motion-cad-e2e', 'preview animation missing');

    var goIdx = -1;
    for (var gi = 0; gi < root.children.length; gi++) {
      if (/(^| )go( |$)/.test(root.children[gi].class || '')) { goIdx = gi; break; }
    }
    assert(goIdx >= 0, 'existing go button not found');
    selPath = [goIdx]; multiSel = [];
    renderAll();
    name.value = 'cad-reshaped';
    name.dispatchEvent(new Event('input', { bubbles: true }));
    mcApplyShape();
    await wait(500);
    assert(/(^| )cad-reshaped( |$)/.test(root.children[goIdx].class || ''), 'apply shape did not add the class');
    assert(root.css.indexOf('/* sui-motion-shape:cad-reshaped */') >= 0, 'apply shape block missing');
    var reshaped = document.querySelector('#prev').contentDocument.querySelector('.cad-reshaped');
    assert(!!reshaped, 'reshaped control missing from preview');
    assert(getComputedStyle(reshaped).maskImage.indexOf('data:image/svg+xml') >= 0, 'reshaped mask missing');
    result(true, '');
  } catch (e) {
    result(false, (e && e.message) ? e.message : String(e));
  }
})();
</script>`
