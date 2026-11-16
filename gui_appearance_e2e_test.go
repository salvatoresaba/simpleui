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
	dropdown "Menu" {
		dropdownitem "One"
	}
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
  var APP_TITLE = 'Button';
  function appSection() {
    return Array.prototype.slice.call(document.querySelectorAll('#inspector .acc')).filter(function (s) { return s.title === APP_TITLE; })[0] || null;
  }
  function selects() {
    var sec = appSection();
    return sec ? Array.prototype.slice.call(sec.querySelectorAll('select')) : [];
  }
  try {
    await until(function () { return typeof __prefsLoaded !== 'undefined' && __prefsLoaded && root && root.children.length > 0; }, 15000);
    await wait(1200);
    await until(function () { return typeof APPEARANCE_CATALOG !== 'undefined' && !!APPEARANCE_CATALOG; }, 4000);
    assert(!!APPEARANCE_CATALOG && APPEARANCE_CATALOG.controls.length === APPEARANCE_KINDS.length, 'external appearance catalog not loaded');
    function pathOfKind(kind) {
      function walk(list, base) {
        for (var i = 0; i < list.length; i++) {
          if (list[i].kind === kind) return base.concat([i]);
          if (list[i].children) { var r = walk(list[i].children, base.concat([i])); if (r) return r; }
        }
        return null;
      }
      return walk(root.children, []);
    }
    var buttonPath = pathOfKind('button');
    assert(!!buttonPath, 'button node missing');
    selPath = buttonPath; multiSel = [];
    renderTree(); renderInspector();
    await until(function () { return !!appSection(); }, 4000);
    var s = selects();
    assert(s.length === 8, 'expected theme, structure, style, animation and four phase selects; got ' + s.length + ' accs=' + Array.prototype.map.call(document.querySelectorAll('#inspector .acc'), function(a){ return a.title; }).join(','));
    assert(appSection().title === 'Button', 'button section must be named Button');
    var phaseLabels = s.slice(4).map(function (sel) { return sel.getAttribute('data-phase') || ''; });
    assert(phaseLabels.join(',') === 'hover,press,focus,state', 'phase selects missing: ' + phaseLabels.join(','));
    var structs = Array.prototype.slice.call(s[1].options).map(function (o) { return o.textContent; });
    assert(structs.indexOf('Circular') >= 0, 'button structures missing');
    var btnStyleLabels = Array.prototype.slice.call(s[2].options).map(function (o) { return o.textContent; });
    assert(btnStyleLabels.indexOf('Dark') >= 0 && btnStyleLabels.indexOf('Light') >= 0, 'button styles must include Dark and Light');
    s[1].value = 'circle'; s[1].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(500);
    var btn = byPath(selPath, root.children);
    assert(String(btn.style || '').indexOf('width:56px') >= 0 || String(btn.style || '').indexOf('width: 56px') >= 0, 'circular structure not applied');
    s = selects(); s[2].value = 'coral'; s[2].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(500);
    btn = byPath(selPath, root.children);
    assert(String(btn.style || '').indexOf('#f43f5e') >= 0, 'coral style not applied');
    s = selects();
    var pulse = Array.prototype.slice.call(s[3].options).filter(function (o) { return o.textContent === 'Pulse'; })[0];
    assert(!!pulse, 'pulse animation missing');
    s[3].value = pulse.value; s[3].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(500);
    btn = byPath(selPath, root.children);
    var btnStyle = String(btn.style || '');
    assert(/sui-button-pulse/.test(btnStyle), 'pulse animation not applied');
    assert(/--sui-hover-animation:/.test(btnStyle), 'hover feedback missing');
    assert(/--sui-active-animation:/.test(btnStyle), 'click feedback missing');
    assert(/--sui-focus-animation:/.test(btnStyle), 'focus feedback missing');
    assert(/--sui-state-animation:/.test(btnStyle), 'state feedback missing');
    assert(/(^|\s)sui-anim-hover(\s|$)/.test(String(btn.class || '')), 'hover phase class missing');
    assert(/(^|\s)sui-anim-press(\s|$)/.test(String(btn.class || '')), 'press phase class missing');
    // The Click menu overrides only the press phase.
    s = selects();
    var pressOpt = Array.prototype.slice.call(s[5].options).filter(function (o) { return o.textContent === 'Press'; })[0];
    assert(!!pressOpt, 'press phase option missing');
    s[5].value = pressOpt.value; s[5].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(500);
    btn = byPath(selPath, root.children);
    assert(String(btn.style || '').indexOf('sui-fx-press') >= 0, 'click phase not applied');
    assert(root.css.indexOf('sui-appearance-phase-rules') >= 0, 'phase CSS block missing');
    // The saved theme must appear in both the inspector and the navigator.
    var oldAsk = window.askPrompt;
    window.askPrompt = function (title, label, initial, cb) { cb('Test theme'); };
    var def = componentDefForKind('button');
    appearanceSaveFavorite(def, 'button', btn, appearanceCurrent(btn, def, 'button'));
    window.askPrompt = oldAsk;
    await wait(400);
    assert(appearanceFavorites().length === 1, 'theme was not saved');
    s = selects();
    var themeLabels = Array.prototype.slice.call(s[0].options).map(function (o) { return o.textContent; });
    assert(themeLabels.join(' ').indexOf('Test theme') >= 0, 'saved theme missing from the inspector');
    openLibraryNavigator('appearance', 'button');
    await until(function () { return document.querySelectorAll('#lib-kinds .lib-group').length > 0; }, 4000);
    var kinds = Array.prototype.slice.call(document.querySelectorAll('#lib-kinds .lib-group')).map(function (b) { return b.textContent; });
    assert(kinds[0] === 'Button', 'button should be first in the appearance kinds');
    var opts = Array.prototype.slice.call(document.querySelectorAll('#lib-structures .lib-app-option')).map(function (b) { return b.textContent; });
    assert(opts.join(' ').indexOf('Circular') >= 0, 'navigator structures missing');
    var animNames = Array.prototype.slice.call(document.querySelectorAll('#lib-animations .lib-app-option')).map(function (b) { return b.textContent; });
    assert(animNames.indexOf('Feedback') >= 0, 'Feedback animation set missing');
    await wait(900);
    var frame = document.getElementById('lib-app-preview');
    assert(frame.srcdoc.indexOf('Button') >= 0, 'appearance preview was not compiled');
    // Keyboard: moving inside a list selects the item and refreshes the preview.
    var firstStruct = document.querySelector('#lib-structures .lib-app-option');
    assert(!!firstStruct, 'structure options missing');
    firstStruct.focus();
    firstStruct.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await wait(800);
    var focused = document.activeElement;
    assert(focused && focused !== firstStruct && focused.getAttribute('data-col') === 'structure', 'ArrowDown did not move focus in the structure list');
    var keyboardId = focused.getAttribute('data-id');
    assert(LIB_APP_SEL.structure === keyboardId, 'ArrowDown did not select the focused structure');
    assert(frame.srcdoc.indexOf('--sui-button-padding: 8px 20px') >= 0, 'ArrowDown did not update the preview');
    focused = document.activeElement;
    focused.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await wait(500);
    assert(LIB_APP_SEL.structure === keyboardId, 'Enter did not select the focused structure');
    // The preview must consume the control animation variable, not another slot.
    var glow = Array.prototype.slice.call(document.querySelectorAll('#lib-animations .lib-app-option')).filter(function (b) { return b.textContent === 'Glow'; })[0];
    assert(!!glow, 'Glow animation missing');
    glow.click();
    await wait(900);
    assert(LIB_APP_SEL.animation === 'glow', 'Glow was not selected');
    assert(/--sui-button-animation:\s*sui-button-glow/.test(frame.srcdoc), 'preview does not apply the button animation variable');
    assert(!/--sui-button-shadow:\s*sui-button-glow/.test(frame.srcdoc), 'animation must not overwrite the shadow variable');
    assert(/--sui-hover-animation:/.test(frame.srcdoc), 'preview hover feedback missing');
    assert(/--sui-active-animation:/.test(frame.srcdoc), 'preview click feedback missing');
    assert(/sui-anim-hover/.test(frame.srcdoc), 'preview hover class missing');
    assert(/sui-anim-press/.test(frame.srcdoc), 'preview press class missing');
    assert(/sui-demo-hover/.test(frame.srcdoc), 'preview phase demo missing');
    closeLibraryNavigator();
    // Every non-complex control gets the section; list is a generic catalog.
    var genericPath = pathOfKind('dropdown');
    assert(!!genericPath, 'dropdown node missing');
    APP_TITLE = appearanceKindLabel('dropdown');
    selPath = genericPath; multiSel = []; renderTree(); renderInspector();
    await until(function () { return !!appSection(); }, 4000);
    var gs = selects();
    assert(gs.length === 8, 'generic control appearance section missing phase selects');
    var gStructs = Array.prototype.slice.call(gs[1].options).map(function (o) { return o.textContent; });
    assert(gStructs.indexOf('Rounded') >= 0, 'generic structures missing');
    var gStyles = Array.prototype.slice.call(gs[2].options).map(function (o) { return o.textContent; });
    assert(gStyles.indexOf('Dark') >= 0 && gStyles.indexOf('Light') >= 0, 'generic styles must include Dark and Light');
    gs[1].value = 'rounded'; gs[1].dispatchEvent(new Event('change', { bubbles: true })); await wait(500);
    gs = selects(); gs[2].value = 'accent'; gs[2].dispatchEvent(new Event('change', { bubbles: true })); await wait(500);
    gs = selects();
    var gPulse = Array.prototype.slice.call(gs[3].options).filter(function (o) { return o.textContent === 'Pulse'; })[0];
    assert(!!gPulse, 'generic pulse animation missing');
    gs[3].value = gPulse.value; gs[3].dispatchEvent(new Event('change', { bubbles: true })); await wait(600);
    var genericNode = byPath(genericPath, root.children);
    var genericStyle = String(genericNode.style || '');
    assert(/border-radius:\s*12px/.test(genericStyle), 'generic structure not applied');
    assert(genericStyle.indexOf('#4c5cf7') >= 0, 'generic style not applied');
    assert(/sui-appearance-pulse/.test(genericStyle), 'generic animation not applied');
    assert(/--sui-hover-animation:/.test(genericStyle), 'generic hover feedback missing');
    assert(/--sui-active-animation:/.test(genericStyle), 'generic click feedback missing');
    assert(root.css.indexOf('sui-appearance-keyframes') >= 0, 'generic keyframes missing');
    assert(!appearanceKindSupported('table') && !appearanceKindSupported('graph') && !appearanceKindSupported('chart'), 'complex controls must not expose appearance');
    // Insert toolbar: Controls has subcategories and Advanced holds table/graph/chart.
    renderInsertMenu();
    var drop = document.getElementById('drop-insert');
    var controlsRow = null;
    drop.querySelectorAll(':scope > .mi.has-sub').forEach(function (row) {
      var label = row.querySelector('span') ? row.querySelector('span').textContent : '';
      if (label === 'Controls') controlsRow = row;
    });
    assert(!!controlsRow, 'Controls group missing');
    var subLabels = Array.prototype.map.call(controlsRow.querySelector('.mi-sub').querySelectorAll(':scope > .mi.has-sub > span:first-child'), function (s) { return s.textContent; });
    assert(subLabels.indexOf('Inputs') >= 0 && subLabels.indexOf('Advanced') >= 0, 'Controls subcategories missing');
    var advancedRow = null;
    controlsRow.querySelector('.mi-sub').querySelectorAll(':scope > .mi.has-sub').forEach(function (row) {
      var label = row.querySelector('span') ? row.querySelector('span').textContent : '';
      if (label === 'Advanced') advancedRow = row;
    });
    assert(!!advancedRow, 'Advanced submenu missing');
    var advancedText = advancedRow.querySelector('.mi-sub').textContent;
    assert(advancedText.indexOf('Table') >= 0 && advancedText.indexOf('Graph') >= 0 && advancedText.indexOf('Chart') >= 0, 'Advanced must contain table, graph and chart');
    try { localStorage.removeItem('sui.appearance.themes.v1'); } catch (e) {}
    result(true, '');
  } catch (e) {
    try { localStorage.removeItem('sui.appearance.themes.v1'); } catch (e2) {}
    result(false, (e && e.message) ? e.message : String(e));
  }
})();
</script>`
