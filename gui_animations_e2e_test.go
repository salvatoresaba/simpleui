package sui

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestGuiControlAnimationsE2E checks the control animation library: every
// option points at a keyframe that ships with the component, the new controls
// expose an animation field, and applying an option actually writes the
// keyframes and the custom property to the project CSS.
func TestGuiControlAnimationsE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping control animation test")
	}
	const source = `view main {
	card "Preview" class: "demo" {
		button "Go" class: "go"
		switch "Enabled"
		progress
	}
}`
	page := injectGuiTestScript(guiAnimationsScript)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=1100,860",
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
		end := marker + 1400
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("control animation e2e failed:\n%s", detail)
}

const guiAnimationsScript = `<script>
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
        setTimeout(loop, 40);
      })();
    });
  }
  function defById(id) {
    for (var i = 0; i < COMPONENT_DEFS.length; i++) if (COMPONENT_DEFS[i].id === id) return COMPONENT_DEFS[i];
    return null;
  }
  function animationField(def) {
    for (var c = 0; c < def.categories.length; c++) {
      var fields = def.categories[c].fields || [];
      for (var f = 0; f < fields.length; f++) {
        if (fields[f].type === 'select' && /-animation$/.test(fields[f].key)) return fields[f];
      }
    }
    return null;
  }
  try {
    await until(function () { return typeof COMPONENT_DEFS !== 'undefined' && COMPONENT_DEFS.length > 10 && __prefsLoaded; }, 15000);
    await until(function () { return typeof root !== 'undefined' && root && root.children && root.children.length > 0; }, 15000);

    // 1. Every option must reference a keyframe shipped by the component.
    var animated = 0, options = 0;
    COMPONENT_DEFS.forEach(function (def) {
      var field = animationField(def);
      if (!field) return;
      animated++;
      if (!/prefers-reduced-motion/.test(def.css)) throw new Error(def.id + ' has no reduced-motion guard');
      (field.opts || []).forEach(function (opt) {
        var value = String(opt[1] || '');
        if (!value || value === 'none' || value === 'sui-spin') return;
        options++;
        var name = value.split(/\s+/)[0];
        if (name.indexOf('sui-') !== 0) throw new Error(def.id + ' option "' + opt[0] + '" is not a sui keyframe');
        if (def.css.indexOf('@keyframes ' + name + ' ') < 0) {
          throw new Error(def.id + ' option "' + opt[0] + '" references missing keyframes ' + name);
        }
      });
    });
    if (animated < 11) throw new Error('expected at least 11 animated controls, found ' + animated);
    if (options < 45) throw new Error('expected at least 45 animation options, found ' + options);
    ['switch', 'check', 'form', 'tabs'].forEach(function (id) {
      var def = defById(id);
      if (!def) throw new Error(id + ' component missing');
      if (!animationField(def)) throw new Error(id + ' has no animation field');
    });

    // 2. Applying a button animation writes keyframes and the custom property.
    var button = defById('button');
    var field = animationField(button);
    var jelly = null;
    field.opts.forEach(function (o) { if (o[0] === 'Jelly') jelly = o[1]; });
    if (!jelly) throw new Error('button Jelly option missing');
    openComponentSettings(button, null);
    __compSettings.values['--sui-button-animation'] = jelly;
    applyComponentSettings();
    await wait(300);
    if (root.css.indexOf('@keyframes sui-button-jelly') < 0) throw new Error('button keyframes not written to CSS');
    if (root.css.indexOf('--sui-button-animation: ' + jelly) < 0) throw new Error('button animation variable not written');
    if (root.css.indexOf('@media (prefers-reduced-motion: reduce)') < 0) throw new Error('reduced-motion guard not written');

    // 3. Progress animations that swap the fill still land together.
    var progress = defById('progress');
    var pfield = animationField(progress);
    var wave = null;
    pfield.opts.forEach(function (o) { if (o[0] === 'Wave') wave = o; });
    if (!wave) throw new Error('progress Wave option missing');
    openComponentSettings(progress, null);
    __compSettings.values['--sui-progress-animation'] = wave[1];
    if (wave[2]) Object.keys(wave[2]).forEach(function (k) { __compSettings.values[k] = wave[2][k]; });
    applyComponentSettings();
    await wait(300);
    if (root.css.indexOf('@keyframes sui-progress-wave') < 0) throw new Error('progress wave keyframes missing');
    if (root.css.indexOf('--sui-progress-fill: ' + wave[2]['--sui-progress-fill']) < 0) throw new Error('progress wave fill not applied');

    // 4. Switch/check/tab animations are wired to a state selector.
    var switchDef = defById('switch');
    if (switchDef.css.indexOf('animation: var(--sui-switch-animation, none)') < 0) throw new Error('switch animation not wired');
    var checkDef = defById('check');
    if (checkDef.css.indexOf('animation: var(--sui-check-animation, none)') < 0) throw new Error('check animation not wired');
    var tabsDef = defById('tabs');
    if (tabsDef.css.indexOf('animation: var(--sui-tab-animation, none)') < 0) throw new Error('tab animation not wired');
    var formDef = defById('form');
    if (formDef.css.indexOf('animation: var(--sui-field-animation, none)') < 0) throw new Error('field focus animation not wired');

    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
