package core

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestGuiEditorToolsE2E covers the formatter, rebindable shortcuts, the image
// asset manager and the interface language switch in the real designer page.
func TestGuiEditorToolsE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping editor tools test")
	}
	const source = `view main {
	card "One" class: "one"
	card "Two" class: "two"
}`
	page := injectGuiTestScript(guiEditorToolsScript)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}
	ln, err := ListenLocal()
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	var lastHTML string
	var savedPrefs = "{}"
	mux := guiTestMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(source))+`}`)
	})
	mux.HandleFunc("GET /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", savedPrefs)
	})
	mux.HandleFunc("POST /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		savedPrefs = r.PostFormValue("prefs")
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
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false}`+`,"diagnostics":`+StringifyJson(msgs)+`}`)
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
	mux.HandleFunc("POST /__css__/setbatch", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssSetBatch(r.PostFormValue("css"), r.PostFormValue("class"), r.PostFormValue("decls"))))
	})
	mux.HandleFunc("POST /__css__/setsel", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssSetSelector(r.PostFormValue("css"), r.PostFormValue("selector"), r.PostFormValue("at"), r.PostFormValue("decls"))))
	})
	mux.HandleFunc("POST /__css__/setselmany", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssSetSelectors(r.PostFormValue("css"), r.PostFormValue("selectors"), r.PostFormValue("at"), r.PostFormValue("decls"))))
	})
	mux.HandleFunc("GET /__gui__/assets", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiAssetList()))
	})
	mux.HandleFunc("POST /__gui__/asset", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(8 << 20)
		file, hdr, err := r.FormFile("image")
		if err != nil {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"no image file"}`)
			return
		}
		defer file.Close()
		data := make([]byte, hdr.Size)
		_, _ = file.Read(data)
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiAssetUpload(data, hdr.Filename)))
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
		end := marker + 1600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("editor tools e2e failed:\n%s", detail)
}

const guiEditorToolsScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function until(fn, timeout) {
    var end = Date.now() + (timeout || 3000);
    return new Promise(function (resolve) {
      (function loop() {
        if (fn()) return resolve(true);
        if (Date.now() > end) return resolve(false);
        setTimeout(loop, 50);
      })();
    });
  }
  function key(target, opts) {
    var ev = new KeyboardEvent('keydown', Object.assign({ bubbles: true, cancelable: true }, opts));
    target.dispatchEvent(ev);
  }
  try {
    await until(function () { return typeof addEl === 'function' && root && root.children.length >= 2 && __prefsLoaded; }, 8000);

    // ---- formatter: pure function ----
    var messy = 'fn  a() {   \n      h1  "x"   \n\n\n      p "y"\n   }';
    var got = formatSuiSource(messy);
    var want = 'fn  a() {\n\th1  "x"\n\n\tp "y"\n}\n';
    if (got !== want) throw new Error('format mismatch: ' + JSON.stringify(got));
    var strSrc = 'fn b() {\n\tp "line {\nstill inside"   \n}\n';
    var strOut = formatSuiSource(strSrc);
    if (strOut.indexOf('still inside"   \n') < 0) throw new Error('formatter changed whitespace inside a string');

    // ---- formatter: apply through the document command ----
    var ta = document.getElementById('out');
    ta.value = 'view main {\n      card "z"\n}';
    formatSuiDocument();
    if (ta.value.indexOf('\tcard "z"') < 0) throw new Error('formatSuiDocument did not reformat: ' + JSON.stringify(ta.value));

    // ---- rebindable shortcuts ----
    document.activeElement.blur();
    if (!isShortcut({ key: 'z', ctrlKey: true }, 'undo')) throw new Error('default undo binding missing');
    setShortcut('undo', 'Ctrl+Alt+U');
    if (!isShortcut({ key: 'u', ctrlKey: true, altKey: true }, 'undo')) throw new Error('rebound undo not matched');
    if (isShortcut({ key: 'z', ctrlKey: true }, 'undo')) throw new Error('old undo binding still active');
    if (!prefs.shortcuts || prefs.shortcuts.undo !== 'Ctrl+Alt+U') throw new Error('shortcut was not stored in prefs');

    var before = root.children.length;
    selPath = [0]; multiSel = [];
    key(document.body, { key: 'Delete' });
    if (root.children.length !== before - 1) throw new Error('delete shortcut did not remove the element');
    key(document.body, { key: 'z', ctrlKey: true });
    if (root.children.length !== before - 1) throw new Error('old Ctrl+Z still undoes after rebinding');
    key(document.body, { key: 'u', ctrlKey: true, altKey: true });
    if (root.children.length !== before) throw new Error('rebound undo did not restore the element');
    resetShortcut('undo');

    // ---- language switch ----
    var menuFile = document.querySelector('[data-menu="file"] .mitem');
    if (!menuFile) throw new Error('file menu missing');
    var okLang = changeEditorLanguage('it');
    await wait(120);
    if (menuFile.textContent !== 'File') throw new Error('file menu title changed unexpectedly: ' + menuFile.textContent);
    var menuEdit = document.querySelector('[data-menu="edit"] .mitem');
    if (!menuEdit || menuEdit.textContent !== 'Modifica') throw new Error('italian menu missing: ' + (menuEdit && menuEdit.textContent));
    var prefTab = document.getElementById('pref-tab-editor');
    if (!prefTab || prefTab.textContent !== 'Editor di codice') throw new Error('italian preferences tab missing');
    if (document.documentElement.getAttribute('lang') !== 'it') throw new Error('html lang not updated');
    var probe = document.createElement('div');
    probe.innerHTML = '<span>Background</span><button title="Undo">x</button>';
    document.body.appendChild(probe);
    localizeRoot(probe);
    if (probe.querySelector('span').textContent !== 'Sfondo') throw new Error('dynamic node not translated: ' + probe.querySelector('span').textContent);
    if (probe.querySelector('button').getAttribute('title') !== 'Annulla') throw new Error('dynamic attribute not translated');
    probe.remove();
    changeEditorLanguage('en');
    await wait(120);
    if (menuEdit.textContent !== 'Edit') throw new Error('english menu not restored: ' + menuEdit.textContent);
    if (prefTab.textContent !== 'Code editor') throw new Error('english preferences tab not restored');

    // ---- image asset manager ----
    openAssets();
    await until(function () { return document.getElementById('asset-editor').classList.contains('on'); }, 2000);
    var png = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0]);
    var file = new File([png], 'dot.png', { type: 'image/png' });
    uploadAsset([file]);
    var cardReady = await until(function () { return !!document.querySelector('#asset-grid .asset-card'); }, 4000);
    var card = document.querySelector('#asset-grid .asset-card');
    if (!cardReady || !card) throw new Error('uploaded asset card missing');
    if ((card.textContent || '').indexOf('dot.png') < 0) throw new Error('asset name missing');
    selPath = [0]; multiSel = [];
    var bgBtn = null, buttons = card.querySelectorAll('.asset-actions button');
    for (var b = 0; b < buttons.length; b++) if ((buttons[b].textContent || '') === 'Background') bgBtn = buttons[b];
    if (!bgBtn) throw new Error('asset Background button missing');
    bgBtn.click();
    var applied = await until(function () { return root.css.indexOf('background-image: url(data:image/png;base64,') >= 0; }, 5000);
    if (!applied) throw new Error('asset background was not written to CSS');

    // ---- asset validation rejects non-images ----
    var bad = new File([new Uint8Array([0x4d, 0x5a, 0x90, 0x00])], 'evil.png', { type: 'image/png' });
    uploadAsset([bad]);
    await wait(400);
    if ((root.css.match(/data:image\/png;base64,/) || []).length > 1) throw new Error('rejected upload changed the CSS');

    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
