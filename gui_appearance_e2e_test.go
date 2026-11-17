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
    assert(!!APPEARANCE_CATALOG && APPEARANCE_KINDS.every(function (kind) {
      return APPEARANCE_CATALOG.controls.some(function (item) { return item.kind === kind; });
    }), 'external appearance catalog is missing a visible control');
    assert(APPEARANCE_KINDS.indexOf('pane') < 0, 'split pane must not be listed as a duplicate control');
    var panelGroup = PALETTE.filter(function (group) { return group.group === 'Controls'; })[0].subs.filter(function (sub) { return sub.name === 'Panels'; })[0];
    assert(panelGroup.tags.indexOf('splitter') >= 0 && panelGroup.tags.indexOf('pane') < 0, 'splitter palette should not list its child pane separately');
    var comboDef = componentDefForKind('combobox');
    var lightPalette = appearancePalette(comboDef, 'light');
    var darkPalette = appearancePalette(comboDef, 'dark');
    assert(lightPalette['--sui-combobox-hover-fg'] === '#0f172a', 'light combo selected text must be dark: ' + lightPalette['--sui-combobox-hover-fg']);
    assert(lightPalette['--sui-combobox-label'] === '#0f172a', 'light combo label must be dark');
    assert(darkPalette['--sui-combobox-hover-fg'] === '#e5e7eb', 'dark combo selected text must be light');
    var checkboxDef = componentDefForKind('checkbox');
    var checkboxProbe = { kind: 'checkbox', class: 'sui-checkbox sui-cb-shape-soft', style: '' };
    appearanceApply(checkboxProbe, checkboxDef, { structure: 'star' });
    assert(/(^|\s)sui-cb-symbol-star(\s|$)/.test(checkboxProbe.class), 'checkbox structure class was not applied: ' + checkboxProbe.class);
    assert(!/sui-cb-shape-soft/.test(checkboxProbe.class), 'previous checkbox structure class was not removed: ' + checkboxProbe.class);
    assert(appearanceCurrent(checkboxProbe, checkboxDef, 'checkbox').structure === 'star', 'checkbox structure class was not recognized');
    var checkboxPreview = appearancePreviewSource(checkboxDef, 'checkbox', { structure: 'cross', style: 'dark', animation: 'none' });
    assert(checkboxPreview.indexOf('sui-cb-symbol-cross') >= 0, 'checkbox structure class was not included in its preview');
    var checkboxCSS = appearanceControlCss(checkboxDef, 'checkbox');
    assert(checkboxCSS.indexOf('sui-cb-shape-diamond') >= 0 && checkboxCSS.indexOf('sui-cb-symbol-heart') >= 0, 'checkbox shape and symbol rules are missing');
    var radioDef = componentDefForKind('radio');
    var radioProbe = { kind: 'radio', class: 'sui-radio sui-rd-shape-soft', style: '' };
    appearanceApply(radioProbe, radioDef, { structure: 'check' });
    assert(/(^|\s)sui-rd-shape-check(\s|$)/.test(radioProbe.class), 'radio check-mark structure class was not applied: ' + radioProbe.class);
    assert(appearanceCurrent(radioProbe, radioDef, 'radio').structure === 'check', 'radio structure class was not recognized');
    assert(appearanceControlCss(radioDef, 'radio').indexOf('sui-rd-shape-diamond') >= 0, 'radio shape rules are missing');
    var spinnerDef = componentDefForKind('spinner');
    var spinnerAnims = appearanceAnimations(spinnerDef, 'spinner').map(function (item) { return item.label; });
    assert(appearanceAnimKey(spinnerDef) === '--sui-spinner-glow-animation', 'spinner animation selection must not replace its structural motion');
    assert(spinnerAnims.join(',') === 'None,Soft glow,Pulsing glow', 'spinner should expose only glow effects: ' + spinnerAnims);
    var spinnerStructures = appearanceStructures(spinnerDef, 'spinner');
    assert(spinnerStructures.some(function (item) { return item.id === 'multiple-dots' && item.attrs && item.attrs.variant === 'multiple-dots'; }), 'multiple-dot spinner structure is not mapped to its CSS variant');
    var hourglass = spinnerStructures.filter(function (item) { return item.id === 'hourglass'; })[0];
    var fillFlip = spinnerStructures.filter(function (item) { return item.id === 'fill-flip'; })[0];
    assert(!!hourglass && parseFloat(hourglass.vars['--sui-spinner-speed']) >= 5, 'hourglass must run slowly enough to show the sand cycle');
    assert(!!fillFlip && parseFloat(fillFlip.vars['--sui-spinner-speed']) >= 4, 'two-colour flip must run slowly enough to see the turn');
    var spinnerCSS = appearanceControlCss(spinnerDef, 'spinner');
    assert(spinnerCSS.indexOf("variant='multiple-dots'") >= 0 && spinnerCSS.indexOf('@keyframes sui-spinner-wave') >= 0, 'spinner variants are missing their structure rules');
    assert(spinnerCSS.indexOf("variant='hourglass']::after") >= 0 && spinnerCSS.indexOf('@keyframes sui-spinner-hourglass-sand') >= 0, 'hourglass must visibly fill its lower chamber');
    assert(spinnerCSS.indexOf("variant='fill-flip'") >= 0 && spinnerCSS.indexOf('@keyframes sui-spinner-fill-flip') >= 0, 'two-colour flip rules are missing');
    var progressDef = componentDefForKind('progress');
    var progressStructures = appearanceStructures(progressDef, 'progress').map(function (item) { return item.id; });
    assert(progressStructures.join(',') === 'thin,classic,segmented,steps,neon,square', 'progress structures should be distinct and value-driven: ' + progressStructures);
    assert(appearanceControlCss(progressDef, 'progress').indexOf('.sui-progress.sui-progress-segmented') >= 0, 'segmented progress style is missing');
    var sliderDef = componentDefForKind('slider');
    var sliderStructures = appearanceStructures(sliderDef, 'slider').map(function (item) { return item.id; });
    assert(sliderStructures.join(',') === 'thin,classic,pill,glass,square,diamond,triangle,segment,neon', 'slider structures should include distinct thumb and track shapes: ' + sliderStructures);
    assert(appearanceControlCss(sliderDef, 'slider').indexOf('.sui-slider.sui-slider-diamond::-webkit-slider-thumb') >= 0, 'diamond slider thumb style is missing');
    assert(appearanceControlCss(sliderDef, 'slider').indexOf('.sui-slider.sui-slider-triangle::-webkit-slider-thumb') >= 0, 'inverted triangle slider thumb style is missing');
    assert(appearanceControlCss(sliderDef, 'slider').indexOf('--sui-slider-thumb-height:24px') >= 0, 'vertical segment slider thumb style is missing');
    var tabsDef = componentDefForKind('tabs');
    assert(appearanceStructures(tabsDef, 'tabs').some(function (item) { return item.id === 'rail' && item.class === 'sui-tabs-rail'; }), 'vertical tab rail structure is missing');
    assert(appearanceControlCss(tabsDef, 'tabs').indexOf('.sui-tabs.sui-tabs-rail > .sui-tablist') >= 0, 'vertical tab rail CSS is missing');
    var accordionDef = componentDefForKind('accordion');
    assert(appearanceStructures(accordionDef, 'accordion').some(function (item) { return item.id === 'separated' && item.class === 'sui-accordion-separated'; }), 'separated accordion structure is missing');
    assert(appearanceControlCss(accordionDef, 'accordion').indexOf('.sui-accordion-separated .sui-accordion-item') >= 0, 'separated accordion CSS is missing');
    var toastDef = componentDefForKind('toast');
    var toastStructures = appearanceStructures(toastDef, 'toast').map(function (item) { return item.id; });
    assert(toastStructures.indexOf('glass') >= 0 && toastStructures.indexOf('compact') >= 0, 'toast structures are missing: ' + toastStructures);
    var appearanceBaseCss = APPEARANCE_CATALOG.definitions.baseCss;
    var checkboxMarkRule = appearanceBaseCss.slice(appearanceBaseCss.indexOf('.sui-checkbox::after'), appearanceBaseCss.indexOf('}', appearanceBaseCss.indexOf('.sui-checkbox::after')));
    var radioMarkRule = appearanceBaseCss.slice(appearanceBaseCss.indexOf('.sui-radio::after'), appearanceBaseCss.indexOf('}', appearanceBaseCss.indexOf('.sui-radio::after')));
    assert(checkboxMarkRule.indexOf('left:50%') >= 0 && checkboxMarkRule.indexOf('top:50%') >= 0, 'checkbox mark must be centered');
    assert(radioMarkRule.indexOf('left:50%') >= 0 && radioMarkRule.indexOf('top:50%') >= 0, 'radio mark must be centered');
    var inputDef = componentDefForKind('input');
    var inputAnims = appearanceAnimations(inputDef, 'input').map(function (a) { return a.label; });
    assert(inputAnims.join(',') === 'None,Feedback', 'text entry animations must be None/Feedback: ' + inputAnims);
    var inputHover = appearancePhaseOptions('hover', 'input').map(function (o) { return o.label; });
    assert(inputHover.indexOf('Lift') < 0 && inputHover.indexOf('Grow') < 0 && inputHover.indexOf('Tilt') < 0, 'text entry hover must not move: ' + inputHover);
    var inputSet = appearanceAnimationSetFor(inputDef, 'input', 'feedback');
    var inputMotion = [inputSet.hover, inputSet.press, inputSet.focus, inputSet.state].join(' ');
    assert(!/sui-fx-(lift|grow|tilt|press|pop|jelly|bounce|slide)/.test(inputMotion), 'text entry feedback must not move or resize: ' + inputMotion);
    assert(/sui-fx-ring/.test(inputSet.hover) && /sui-fx-ring/.test(inputSet.focus), 'text entry feedback must use the ring');
    var switchDef = componentDefForKind('switch');
    var switchAnims = appearanceAnimations(switchDef, 'switch').map(function (a) { return a.label; });
    assert(switchAnims.indexOf('Morph') >= 0 && switchAnims.indexOf('Fade') >= 0 && switchAnims.indexOf('Scale') >= 0, 'switch animations missing Morph/Fade/Scale: ' + switchAnims);
    var switchIos = appearanceStructures(switchDef, 'switch').filter(function (s) { return s.id === 'ios'; })[0];
    assert(!!switchIos && ((switchIos.attrs && switchIos.attrs.shape === 'ios') || switchIos.class === 'sui-sw-shape-ios'), 'switch iOS structure must set its shape');
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
    var sliderKind = Array.prototype.slice.call(document.querySelectorAll('#lib-kinds .lib-group')).filter(function (b) { return b.textContent.trim() === appearanceKindLabel('slider'); })[0];
    assert(!!sliderKind, 'slider is missing from the appearance navigator');
    sliderKind.click();
    await until(function () { return LIB_APP_KIND === 'slider' && document.querySelectorAll('#lib-structures .lib-app-option').length > 0; }, 4000);
    var triangleOption = Array.prototype.slice.call(document.querySelectorAll('#lib-structures .lib-app-option')).filter(function (b) { return b.getAttribute('data-id') === 'triangle'; })[0];
    assert(!!triangleOption, 'inverted triangle option is missing from the slider navigator');
    triangleOption.click();
    await until(function () { return /sui-slider-triangle/.test(frame.srcdoc); }, 4000);
    assert(LIB_APP_SEL.structure === 'triangle' && /sui-slider-triangle/.test(frame.srcdoc), 'inverted triangle selection did not update the compiled preview');
    assert(/clip-path:\s*polygon\(0 0,100% 0,50% 100%\)/.test(frame.srcdoc), 'compiled preview is missing the inverted triangle thumb CSS');
    var segmentOption = Array.prototype.slice.call(document.querySelectorAll('#lib-structures .lib-app-option')).filter(function (b) { return b.getAttribute('data-id') === 'segment'; })[0];
    assert(!!segmentOption, 'vertical segment option is missing from the slider navigator');
    segmentOption.click();
    await until(function () { return /sui-slider-segment/.test(frame.srcdoc); }, 4000);
    assert(LIB_APP_SEL.structure === 'segment' && /sui-slider-segment/.test(frame.srcdoc), 'vertical segment selection did not update the compiled preview');
    assert(/--sui-slider-thumb-height:\s*24px/.test(frame.srcdoc), 'compiled preview is missing the vertical segment thumb size');
    var tabsKind = Array.prototype.slice.call(document.querySelectorAll('#lib-kinds .lib-group')).filter(function (b) { return b.textContent.trim() === appearanceKindLabel('tabs'); })[0];
    assert(!!tabsKind, 'tabs is missing from the appearance navigator');
    tabsKind.click();
    await until(function () { return LIB_APP_KIND === 'tabs' && document.querySelectorAll('#lib-structures .lib-app-option').length > 0; }, 4000);
    var railOption = Array.prototype.slice.call(document.querySelectorAll('#lib-structures .lib-app-option')).filter(function (b) { return b.getAttribute('data-id') === 'rail'; })[0];
    assert(!!railOption, 'vertical rail option is missing from the tabs navigator');
    railOption.click();
    await until(function () { return /sui-tabs-rail/.test(frame.srcdoc); }, 4000);
    assert(/sui-tabs-rail/.test(frame.srcdoc), 'vertical tab rail did not update the compiled preview');
    var accordionKind = Array.prototype.slice.call(document.querySelectorAll('#lib-kinds .lib-group')).filter(function (b) { return b.textContent.trim() === appearanceKindLabel('accordion'); })[0];
    assert(!!accordionKind, 'accordion is missing from the appearance navigator');
    accordionKind.click();
    await until(function () { return LIB_APP_KIND === 'accordion' && document.querySelectorAll('#lib-structures .lib-app-option').length > 0; }, 4000);
    var separatedOption = Array.prototype.slice.call(document.querySelectorAll('#lib-structures .lib-app-option')).filter(function (b) { return b.getAttribute('data-id') === 'separated'; })[0];
    assert(!!separatedOption, 'separated cards option is missing from the accordion navigator');
    separatedOption.click();
    await until(function () { return /sui-accordion-separated/.test(frame.srcdoc); }, 4000);
    assert(/sui-accordion-separated/.test(frame.srcdoc), 'separated accordion did not update the compiled preview');
    var toastKind = Array.prototype.slice.call(document.querySelectorAll('#lib-kinds .lib-group')).filter(function (b) { return b.textContent.trim() === appearanceKindLabel('toast'); })[0];
    assert(!!toastKind, 'toast is missing from the appearance navigator');
    toastKind.click();
    await until(function () { return LIB_APP_KIND === 'toast' && document.querySelectorAll('#lib-structures .lib-app-option').length > 0; }, 4000);
    var glassOption = Array.prototype.slice.call(document.querySelectorAll('#lib-structures .lib-app-option')).filter(function (b) { return b.getAttribute('data-id') === 'glass'; })[0];
    assert(!!glassOption, 'glass toast option is missing from the navigator');
    glassOption.click();
    await until(function () { return /--sui-toast-backdrop:\s*blur\(14px\)/.test(frame.srcdoc); }, 4000);
    assert(/--sui-toast-backdrop:\s*blur\(14px\)/.test(frame.srcdoc), 'glass toast did not update the compiled preview');
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
