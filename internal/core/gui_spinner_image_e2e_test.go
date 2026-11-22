package core

import (
	"context"
	"encoding/base64"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestGuiSpinnerImageE2E checks the spinner inspector image mode: the checkbox
// reveals the source picker, disables structure/style/animation, applies an
// uploaded asset and renders the animated image in the live preview.
func TestGuiSpinnerImageE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping spinner image test")
	}
	gif, err := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
	if err != nil {
		t.Fatal(err)
	}
	res := guiAssetUpload(gif, "spinner.gif")
	if !asBool(get(res, "ok")) {
		t.Fatalf("asset upload failed: %s", StringifyJson(res))
	}
	assetURL := asStr(get(get(res, "asset"), "url"))

	const source = `view main {
	spinner "Loading"
}`
	page := injectGuiTestScript(guiSpinnerImageScript)
	ln, err := ListenLocal()
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	var lastHTML, lastImageHTML string
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
			if strings.Contains(r.PostFormValue("sui"), "src:") {
				lastImageHTML = html
			}
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
			end := marker + 2200
			if end > len(text) {
				end = len(text)
			}
			detail = text[marker:end]
		}
		t.Fatalf("spinner image e2e failed:\n%s", detail)
	}
	if !strings.Contains(lastImageHTML, "sui-spinner-image") || !strings.Contains(lastImageHTML, assetURL) {
		t.Fatalf("preview HTML does not contain the image spinner source")
	}
}

const guiSpinnerImageScript = `<script>
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
  function spinnerSection() {
    var secs = Array.prototype.slice.call(document.querySelectorAll('#inspector .acc'));
    for (var i = 0; i < secs.length; i++) if (secs[i].title === 'Spinner') return secs[i];
    return null;
  }
  try {
    await until(function () { return typeof __prefsLoaded !== 'undefined' && __prefsLoaded && root && root.children.length > 0; }, 15000);
    await wait(1200);
    selPath = [0]; renderTree(); renderInspector();
    await until(function () { return !!spinnerSection(); }, 4000);
    var sec = spinnerSection();
    var toggle = sec.querySelector('#f_spinner_image');
    var source = sec.querySelector('#f_spinner_image_src');
    var upload = sec.querySelector('#f_spinner_image_upload');
    assert(!!toggle && !!source && !!upload, 'spinner image controls are missing');
    assert(toggle.checked === false, 'image mode must start disabled');
    var fields = Array.prototype.slice.call(sec.querySelectorAll('select[data-appearance-field]'));
    var phases = Array.prototype.slice.call(sec.querySelectorAll('select[data-phase]'));
    assert(fields.length === 4 && phases.length === 4, 'appearance selectors are missing');
    assert(fields.every(function (s) { return !s.disabled; }), 'appearance selectors must start enabled');

    toggle.checked = true;
    toggle.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(300);
    assert(fields.every(function (s) { return s.disabled; }), 'structure/style/animation must be disabled in image mode');
    assert(phases.every(function (s) { return s.disabled; }), 'phase selectors must be disabled in image mode');
    assert(source.parentNode.style.display !== 'none', 'source picker must be visible in image mode');

    var ready = await until(function () {
      return Array.prototype.some.call(source.options, function (o) { return o.textContent.indexOf('spinner.gif') >= 0; });
    }, 5000);
    assert(ready, 'uploaded asset is missing from the source picker');
    var assetURL = '';
    for (var i = 0; i < source.options.length; i++) {
      if (source.options[i].textContent.indexOf('spinner.gif') >= 0) { assetURL = source.options[i].value; break; }
    }
    source.value = assetURL;
    source.dispatchEvent(new Event('change', { bubbles: true }));

    var applied = await until(function () {
      var n = root.children[0];
      return !!(n && String(n.src || '').indexOf('data:image/gif;base64,') === 0);
    }, 6000);
    assert(applied, 'spinner src was not applied');
    assert(String(suiSource || '').indexOf('src: "data:image/gif;base64,') >= 0, 'generated source is missing the image src');
    assert(String(suiSource || '').indexOf('variant:') < 0, 'image mode should not add a CSS variant');

    var frame = document.getElementById('prev');
    var readyPreview = await until(function () {
      try {
        var doc = frame && frame.contentDocument;
        return !!(doc && doc.querySelector('.sui-spinner-img'));
      } catch (e) { return false; }
    }, 8000);
    assert(readyPreview, 'preview did not render the image spinner');
    var doc = frame.contentDocument;
    var img = doc.querySelector('.sui-spinner-img');
    var wrap = doc.querySelector('.sui-spinner-image');
    assert(String(img.getAttribute('src') || '').indexOf('data:image/gif;base64,') === 0, 'preview image source is wrong');
    assert(img.getAttribute('alt') === 'Loading', 'preview image alt text is missing');
    var wrapStyle = frame.contentWindow.getComputedStyle(wrap);
    assert(wrapStyle.animationName === 'none', 'image spinner must not run the CSS animation');
    assert(wrapStyle.borderTopWidth === '0px', 'image spinner must not draw the CSS ring');

    selPath = [0]; renderTree(); renderInspector();
    await until(function () { return !!spinnerSection(); }, 4000);
    sec = spinnerSection();
    toggle = sec.querySelector('#f_spinner_image');
    assert(toggle.checked === true, 'image mode must stay checked after re-render');
    fields = Array.prototype.slice.call(sec.querySelectorAll('select[data-appearance-field]'));
    assert(fields.every(function (s) { return s.disabled; }), 'appearance selectors must stay disabled after re-render');

    toggle.checked = false;
    toggle.dispatchEvent(new Event('change', { bubbles: true }));
    var cleared = await until(function () {
      var n = root.children[0];
      return !n.src;
    }, 6000);
    assert(cleared, 'disabling image mode did not clear the source');
    assert(String(suiSource || '').indexOf('src:') < 0, 'disabling image mode left src in the generated source');
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`
