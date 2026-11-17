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

const guiI18nSource = `i18n "it" {
	"save": "Salva"
}
i18n "en" {
	"save": "Save"
}
state key = "save"
view main {
	p "{t(key)}"
}`

// TestGuiI18nEditorE2E checks the Languages modal, the preview-language
// selector and the live source update when a translation is edited.
func TestGuiI18nEditorE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping i18n designer test")
	}
	page := injectGuiTestScript(guiI18nScript)
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
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(guiI18nSource))+`}`)
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
	if !strings.Contains(text, `id="e2e-result">OK<`) {
		marker := strings.Index(text, `id="e2e-result"`)
		detail := text
		if marker >= 0 {
			end := marker + 2400
			if end > len(text) {
				end = len(text)
			}
			detail = text[marker:end]
		}
		t.Fatalf("i18n designer e2e failed:\n%s", detail)
	}
	if !strings.Contains(lastHTML, "i18n") {
		t.Fatalf("preview HTML does not contain the i18n runtime")
	}
}

const guiI18nScript = `<script>
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
  function button(label) {
    var all = Array.prototype.slice.call(document.querySelectorAll('#i18n-messages button'));
    for (var i = 0; i < all.length; i++) if (all[i].textContent === label) return all[i];
    return null;
  }
  try {
    await until(function () { return typeof __prefsLoaded !== 'undefined' && __prefsLoaded && root && root.children.length > 0; }, 15000);
    await wait(1200);
    assert(Array.isArray(root.i18n) && root.i18n.length === 2, 'model i18n catalogs missing');
    assert(root.i18n[0].locale === 'it' && root.i18n[1].locale === 'en', 'model locale order is wrong');

    openI18nEditor();
    await until(function () { return document.getElementById('i18n-editor').classList.contains('on'); }, 3000);
    var rows = document.querySelectorAll('#i18n-locales .i18n-locale');
    assert(rows.length === 2, 'locale list is missing entries: ' + rows.length);
    assert(document.querySelectorAll('#i18n-messages tbody tr').length === 1, 'message table is missing the save key');

    var lang = document.getElementById('preview-locale');
    assert(Array.prototype.some.call(lang.options, function (o) { return o.value === 'en'; }), 'preview language selector is missing en');
    lang.value = 'en';
    lang.dispatchEvent(new Event('change', { bubbles: true }));
    var english = await until(function () {
      try {
        var doc = document.getElementById('prev').contentDocument;
        return !!(doc && doc.body && doc.body.innerText.indexOf('Save') >= 0);
      } catch (e) { return false; }
    }, 8000);
    assert(english, 'preview did not switch to English');

    var textarea = document.querySelector('#i18n-messages tbody tr td:nth-child(2) textarea');
    assert(!!textarea, 'translation textarea is missing');
    textarea.focus();
    textarea.value = 'Salva!';
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(700);
    assert(String(suiSource || '').indexOf('"save": "Salva!"') >= 0, 'edited translation is missing from the source');

    var add = button('Add key.');
    assert(!!add, 'Add key button is missing');
    add.click();
    await wait(1200);
    assert(String(suiSource || '').indexOf('new.key.2') >= 0, 'new key is missing from the source');
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`
