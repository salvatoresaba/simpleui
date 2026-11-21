package main

import (
	"bytes"
	"context"
	"image/png"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const themeE2ESource = `view main {
	menubar {
		menu "File" {
			menuitem "Open"
		}
	}
	h1 "Theme check"
	p "Body"
	button "Save"
	progress value: 62 max: 100
	divider
	spinner
	img src: "logo.png"
}`

// theme_e2e_test.go drives the real designer in a headless browser and checks
// the rendered theme through computed styles and CSS variables. It covers the
// light/dark page bodies, base elements, tooltips, media, hover presets, chart
// theme presets and the custom theme export/import payload.
func TestGuiThemeVisualE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping theme visual regression test")
	}

	const source = themeE2ESource

	page := injectGuiTestScript(themeE2EScript)
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
	go func() { _ = http.Serve(ln, mux) }()

	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	profile := t.TempDir()
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--no-first-run",
		"--no-default-browser-check",
		"--user-data-dir=" + profile,
		"--virtual-time-budget=30000",
		"--dump-dom",
		url,
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
		end := marker + 1200
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("theme visual e2e failed:\n%s", detail)
}

const themeE2EScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-result';
      document.body.insertBefore(el, document.body.firstChild);
    }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  var wait = function (ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); };
  function lum(color) {
    var c = String(color || '').trim();
    var hex = c.match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);
    if (hex) {
      var h = hex[1];
      if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
      return 0.2126 * parseInt(h.slice(0, 2), 16) + 0.7152 * parseInt(h.slice(2, 4), 16) + 0.0722 * parseInt(h.slice(4, 6), 16);
    }
    var m = c.match(/[0-9.]+/g);
    if (!m || m.length < 3) return -1;
    return 0.2126 * parseFloat(m[0]) + 0.7152 * parseFloat(m[1]) + 0.0722 * parseFloat(m[2]);
  }
  function frame() {
    var f = document.getElementById('prev');
    return f && f.contentDocument ? f.contentDocument : null;
  }
  function docStyle(doc, selector, prop) {
    var el = doc && doc.querySelector(selector);
    if (!el) return '';
    return doc.defaultView.getComputedStyle(el)[prop] || '';
  }
  function rootVar(doc, name) {
    if (!doc) return '';
    return String(doc.defaultView.getComputedStyle(doc.documentElement).getPropertyValue(name) || '').trim();
  }
  async function waitPreview() {
    for (var i = 0; i < 160; i++) {
      var doc = frame();
      if (doc && doc.querySelector('.sui-view') && doc.querySelector('h1')) return doc;
      await wait(100);
    }
    return null;
  }
  async function setTheme(id) {
    openProjectPreferences();
    await wait(120);
    var sel = document.getElementById('projprefs-theme');
    if (!sel) throw new Error('theme selector not found');
    sel.value = id;
    sel.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(900);
    closeProjectPreferences();
    await wait(200);
    return waitPreview();
  }
  try {
    for (var i = 0; i < 160 && !document.querySelector('#nodes .node'); i++) await wait(100);
    await wait(300);

    var lightDoc = await setTheme('standard-light');
    if (!lightDoc) throw new Error('light preview did not render');
    var light = {
      bodyBg: docStyle(lightDoc, 'body', 'backgroundColor'),
      bodyFg: docStyle(lightDoc, 'body', 'color'),
      viewBg: docStyle(lightDoc, '.sui-view', 'backgroundColor'),
      menuBg: docStyle(lightDoc, '.sui-menubar', 'backgroundColor'),
      h1Fg: docStyle(lightDoc, 'h1', 'color'),
      pFg: docStyle(lightDoc, 'p', 'color'),
      imgBg: docStyle(lightDoc, 'img', 'backgroundColor'),
      imgRadius: docStyle(lightDoc, 'img', 'borderRadius'),
      tooltipBg: rootVar(lightDoc, '--sui-tooltip-bg'),
      tooltipFg: rootVar(lightDoc, '--sui-tooltip-fg'),
      mediaBg: rootVar(lightDoc, '--sui-media-bg'),
      progressHover: rootVar(lightDoc, '--sui-progress-hover-track'),
      spinnerHover: rootVar(lightDoc, '--sui-spinner-hover-track'),
      dividerHover: rootVar(lightDoc, '--sui-divider-hover-color')
    };

    var darkDoc = await setTheme('standard-dark');
    if (!darkDoc) throw new Error('dark preview did not render');
    var dark = {
      bodyBg: docStyle(darkDoc, 'body', 'backgroundColor'),
      bodyFg: docStyle(darkDoc, 'body', 'color'),
      viewBg: docStyle(darkDoc, '.sui-view', 'backgroundColor'),
      menuBg: docStyle(darkDoc, '.sui-menubar', 'backgroundColor'),
      h1Fg: docStyle(darkDoc, 'h1', 'color'),
      tooltipBg: rootVar(darkDoc, '--sui-tooltip-bg'),
      tooltipFg: rootVar(darkDoc, '--sui-tooltip-fg'),
      mediaBg: rootVar(darkDoc, '--sui-media-bg'),
      imgBg: docStyle(darkDoc, 'img', 'backgroundColor')
    };

    var themePresets = CHART_STYLE_PRESETS.filter(function (p) { return p.group === 'theme'; });
    var roundtrip = STYLE_VARIANT_DEFS.every(function (v) {
      var key = 'theme-' + v.id;
      return chartPresetKey({ styleExpr: chartPresetStyleExpr(key) }) === key;
    });
    function rgb(c) {
      c = String(c || '').trim().toLowerCase();
      if (!c || c === 'transparent' || c === 'none' || c === 'inherit' || c === 'currentcolor') return null;
      if (c[0] === '#') {
        var h = c.slice(1);
        if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
        if (!/^[0-9a-f]{6}$/.test(h)) return null;
        return [parseInt(h.slice(0, 2), 16), parseInt(h.slice(2, 4), 16), parseInt(h.slice(4, 6), 16), 1];
      }
      var m = c.match(/rgba?\(([^)]+)\)/);
      if (m) {
        var p = m[1].split(',').map(function (x) { return parseFloat(x.trim()); });
        if (p.length >= 3) return [p[0], p[1], p[2], p.length > 3 ? p[3] : 1];
      }
      return null;
    }
    function blendRgb(fg, bg) {
      var a = fg[3] === undefined ? 1 : fg[3];
      if (a >= 1) return fg;
      return [fg[0] * a + bg[0] * (1 - a), fg[1] * a + bg[1] * (1 - a), fg[2] * a + bg[2] * (1 - a), 1];
    }
    function luminanceRgb(c) {
      function f(v) { v /= 255; return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4); }
      return 0.2126 * f(c[0]) + 0.7152 * f(c[1]) + 0.0722 * f(c[2]);
    }
    function contrastRatio(a, b) {
      var x = luminanceRgb(a), y = luminanceRgb(b);
      if (y > x) { var t = x; x = y; y = t; }
      return (x + 0.05) / (y + 0.05);
    }
    var contrastFailures = [];
    STYLE_VARIANT_DEFS.forEach(function (v) {
      var page = v.page || (v.vars && v.vars.page) || {};
      var pageBg = rgb(page['--sui-bg']) || [20, 23, 28, 1];
      var panel = rgb(page['--sui-panel']) || pageBg;
      function vars(id) { return (v.vars && v.vars[id]) || {}; }
      function checkPair(label, fg, bg, min) {
        var f = rgb(fg), b = rgb(bg) || panel;
        if (!f || !b) return;
        var solidBg = blendRgb(b, pageBg);
        var solidFg = blendRgb(f, solidBg);
        var r = contrastRatio(solidFg, solidBg);
        if (r + 0.002 < min) contrastFailures.push(v.id + ':' + label + ' ' + r.toFixed(2) + '<' + min);
      }
      function width(map, key, def) { var w = parseFloat(map[key]); return isFinite(w) ? w : (def === undefined ? 1 : def); }
      checkPair('page text', page['--sui-fg'], page['--sui-bg'], 4.5);
      checkPair('panel text', page['--sui-fg'], page['--sui-panel'], 4.5);
      checkPair('panel2 text', page['--sui-fg'], page['--sui-panel2'], 4.5);
      checkPair('muted page', page['--sui-muted'], page['--sui-bg'], 3);
      checkPair('muted panel', page['--sui-muted'], page['--sui-panel'], 3);
      checkPair('link page', page['--sui-link'], page['--sui-bg'], 4.5);
      checkPair('link panel', page['--sui-link'], page['--sui-panel'], 4.5);
      var b = vars('button');
      checkPair('button', b['--sui-button-fg'], b['--sui-button-bg'], 4.5);
      checkPair('button hover', b['--sui-button-hover-fg'] || b['--sui-button-fg'], b['--sui-button-hover'] || b['--sui-button-bg'], 4.5);
      var f = vars('form');
      checkPair('field', f['--sui-field-fg'], f['--sui-field-bg'], 4.5);
      if (width(f, '--sui-field-border-width', 1) > 0) checkPair('field border', f['--sui-field-border'], f['--sui-field-bg'], 1.8);
      checkPair('field focus', f['--sui-field-focus'], f['--sui-field-bg'], 2.5);
      var bd = vars('badge');
      checkPair('badge', bd['--sui-badge-fg'], bd['--sui-badge-bg'], 4.5);
      checkPair('badge hover', bd['--sui-badge-hover-fg'] || bd['--sui-badge-fg'], bd['--sui-badge-hover-bg'] || bd['--sui-badge-bg'], 4.5);
      var al = vars('alert');
      checkPair('alert', al['--sui-alert-fg'], al['--sui-alert-bg'], 4.5);
      checkPair('alert hover', al['--sui-alert-hover-fg'] || al['--sui-alert-fg'], al['--sui-alert-hover-bg'] || al['--sui-alert-bg'], 4.5);
      checkPair('card', vars('card')['--sui-card-fg'], vars('card')['--sui-card-bg'], 4.5);
      var ck = vars('check');
      checkPair('check mark', ck['--sui-check-mark'], ck['--sui-check-checked'], 3);
      if (width(ck, '--sui-check-border-width', 1) > 0) checkPair('check border', ck['--sui-check-border'], ck['--sui-check-bg'] || page['--sui-panel'], 2.5);
      var t = vars('tabs');
      checkPair('tab', t['--sui-tab-fg'], t['--sui-tablist-bg'], 3);
      checkPair('active tab', t['--sui-tab-active-fg'], t['--sui-tab-active-bg'] || t['--sui-tabs-bg'], 4.5);
      checkPair('tab hover', t['--sui-tab-hover-fg'] || t['--sui-tab-fg'], t['--sui-tab-hover-bg'] || t['--sui-tablist-bg'], 4.5);
      var ac = vars('accordion');
      checkPair('accordion', ac['--sui-accordion-fg'], ac['--sui-accordion-head-bg'] || ac['--sui-accordion-bg'], 4.5);
      checkPair('accordion hover', ac['--sui-accordion-hover-fg'] || ac['--sui-accordion-fg'], ac['--sui-accordion-hover-bg'] || ac['--sui-accordion-head-bg'] || ac['--sui-accordion-bg'], 4.5);
      var cb = vars('combobox');
      checkPair('combobox', cb['--sui-combobox-fg'], cb['--sui-combobox-bg'], 4.5);
      checkPair('combobox label', cb['--sui-combobox-label'], cb['--sui-combobox-bg'], 3);
      var dp = vars('datepicker');
      checkPair('datepicker', dp['--sui-datepicker-fg'], dp['--sui-datepicker-bg'], 4.5);
      checkPair('datepicker label', dp['--sui-datepicker-label'], dp['--sui-datepicker-bg'], 3);
      var tr = vars('tree');
      checkPair('tree', tr['--sui-tree-fg'], page['--sui-panel'], 4.5);
      checkPair('tree selected', tr['--sui-tree-selected-fg'], tr['--sui-tree-selected-bg'] || page['--sui-panel'], 4.5);
      var tb = vars('table');
      checkPair('table', tb['--sui-table-fg'], tb['--sui-table-bg'], 4.5);
      checkPair('table head', tb['--sui-table-head-fg'], tb['--sui-table-head-bg'], 4.5);
      checkPair('table row', tb['--sui-table-fg'], tb['--sui-table-row-bg'], 4.5);
      checkPair('table striped', tb['--sui-table-fg'], tb['--sui-table-striped-bg'], 4.5);
      var to = vars('toast');
      checkPair('toast', to['--sui-toast-fg'], to['--sui-toast-bg'], 4.5);
      checkPair('toast hover', to['--sui-toast-hover-fg'] || to['--sui-toast-fg'], to['--sui-toast-hover-bg'] || to['--sui-toast-bg'], 4.5);
      var mo = vars('modal');
      checkPair('modal', mo['--sui-modal-fg'], mo['--sui-modal-bg'], 4.5);
      checkPair('modal button', mo['--sui-modal-button-fg'], mo['--sui-modal-button-bg'], 4.5);
      checkPair('modal button hover', mo['--sui-modal-button-hover-fg'] || mo['--sui-modal-button-fg'], mo['--sui-modal-button-hover'] || mo['--sui-modal-button-bg'], 4.5);
      var me = vars('menu');
      checkPair('menu', me['--sui-menu-fg'], me['--sui-menu-button-bg'] || page['--sui-panel'], 4.5);
      checkPair('menu popup', me['--sui-menu-fg'], me['--sui-menu-popup-bg'] || page['--sui-panel'], 4.5);
      checkPair('menu hover', me['--sui-menu-hover-fg'] || me['--sui-menu-fg'], me['--sui-menu-hover'] || me['--sui-menu-button-bg'] || page['--sui-panel'], 4.5);
    });
    var nameInput = document.getElementById('projprefs-theme-name');
    if (nameInput) nameInput.value = 'QA theme';
    var payload = buildProjectThemePayload();
    applyImportedThemeObject(JSON.parse(JSON.stringify(payload)));
    var importedName = normalizeThemePayload(payload).name;
    var fuzzyKey = chartPresetKey({ styleExpr: 'cssSafe(' + JSON.stringify(chartPresetCss('theme-standard-light') + '; --sui-chart-extra-check: 1') + ')' });
    newSui();
    await wait(80);
    applyProjectTheme('standard-light');
    await wait(300);
    addEl('chart');
    await wait(200);
    var chartNode = byPath(selPath, root.children);
    chartNode.chartTheme = 'dark';
    chartNode.styleExpr = chartPresetStyleExpr('default');
    setOut();
    autoPreview();
    await wait(900);
    var chartDoc = frame();
    var chartCtrlBg = docStyle(chartDoc, '.sui-chart-menu-btn', 'backgroundColor');
    var chartCtrlFg = docStyle(chartDoc, '.sui-chart-menu-btn', 'color');
    var chartBgVar = chartDoc && chartDoc.querySelector('.sui-chart') ? chartDoc.defaultView.getComputedStyle(chartDoc.querySelector('.sui-chart')).getPropertyValue('--sui-chart-bg').trim() : '';
    var chartControlContrast = !!(chartCtrlBg && chartCtrlFg && contrastRatio(rgb(chartCtrlFg), blendRgb(rgb(chartCtrlBg), rgb('#14171c'))) >= 4.5);
    var chartDarkOnLight = chartControlContrast && chartBgVar === '#14171c';
    var chartPayload = JSON.parse(JSON.stringify(payload));
    chartPayload.charts = { chart: { chartTheme: 'light', styleExpr: chartPresetStyleExpr('theme-standard-light') }, graph: null };
    applyImportedThemeObject(chartPayload);
    var chartReapplied = !!chartNode && chartNode.chartTheme === 'light' && chartNode.styleExpr === chartPayload.charts.chart.styleExpr;
    var invalidRejected = false;
    try {
      normalizeThemePayload({ format: 'simpleui-theme', version: 1, page: { 'color': 'red' } });
    } catch (e) { invalidRejected = true; }
    var unsafeRejected = false;
    try {
      normalizeThemePayload({ format: 'simpleui-theme', version: 1, page: { '--sui-bg': 'red; } body { display:none' } });
    } catch (e) { unsafeRejected = true; }
    var unknownRejected = false;
    try {
      normalizeThemePayload({ format: 'simpleui-theme', version: 1, page: { '--sui-not-a-real-key': 'red' } });
    } catch (e) { unknownRejected = true; }

    var checks = {
      lightBody: lum(light.bodyBg) > 180,
      lightText: lum(light.bodyFg) < 90,
      lightView: lum(light.viewBg) > 200,
      lightMenu: lum(light.menuBg) > 200,
      lightH1: lum(light.h1Fg) < 90,
      lightParagraph: lum(light.pFg) < 90,
      lightMediaBg: lum(light.imgBg) > 200,
      lightMediaVars: light.mediaBg !== '' && lum(light.mediaBg) > 200,
      lightTooltip: light.tooltipBg !== '' && lum(light.tooltipBg) > 200 && lum(light.tooltipFg) < 90,
      lightHoverVars: light.progressHover !== '' && light.spinnerHover !== '' && light.dividerHover !== '',
      darkBody: lum(dark.bodyBg) < 60,
      darkText: lum(dark.bodyFg) > 160,
      darkView: lum(dark.viewBg) < 90,
      darkMenu: lum(dark.menuBg) < 90,
      darkH1: lum(dark.h1Fg) > 160,
      darkMedia: dark.imgBg !== '' && lum(dark.imgBg) < 90,
      darkTooltip: dark.tooltipBg !== '' && lum(dark.tooltipBg) < 90 && lum(dark.tooltipFg) > 160,
      mediaRadius: light.imgRadius !== '',
      themePresetCount: themePresets.length === STYLE_VARIANT_DEFS.length && themePresets.length >= 24,
      themePresetRoundtrip: roundtrip,
      contrastAudit: contrastFailures.length === 0,
      chartDarkOnLight: chartDarkOnLight,
      themePresetFuzzy: fuzzyKey === 'theme-standard-light',
      importReappliesCharts: chartReapplied,
      themeName: importedName === 'QA theme',
      exportPayload: !!(payload.page && payload.page['--sui-bg'] && payload.components && payload.tooltip && payload.charts),
      importRejectsUnknownKey: invalidRejected,
      importRejectsUnknownVariable: unknownRejected,
      importRejectsCssInjection: unsafeRejected
    };
    var failed = Object.keys(checks).filter(function (k) { return !checks[k]; });
    if (failed.length) {
      result(false, failed.join(',') + ' ' + JSON.stringify({ light: light, dark: dark, themePresets: themePresets.length, variants: STYLE_VARIANT_DEFS.length, contrast: contrastFailures.slice(0, 30) }));
      return;
    }
    result(true, '');
  } catch (e) {
    result(false, (e && (e.stack || e.message)) || String(e));
  }
})();
</script>`

// TestGuiThemeScreenshotGolden captures a deterministic palette grid for the
// Standard Light and Dark themes and compares it with checked-in golden PNGs.
// The grid uses solid theme colours only (no text or animation), so the result
// is stable across machines. Set SUI_UPDATE_THEME_GOLDEN=1 to regenerate.
func TestGuiThemeScreenshotGolden(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping theme screenshot test")
	}

	themes := []string{"standard-light", "standard-dark"}
	pages := map[string]string{}
	for _, id := range themes {
		script := strings.ReplaceAll(themeShotScriptTemplate, "__THEME__", id)
		page := injectGuiTestScript(script)
		if page == guiHTML {
			t.Fatal("designer page has no closing body")
		}
		pages[id] = page
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	var lastHTML string
	mux := guiTestMux()
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(themeE2ESource))+`}`)
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
	for _, id := range themes {
		theme := id
		mux.HandleFunc("GET /"+theme, func(w http.ResponseWriter, r *http.Request) {
			writeRawHandler(w, "text/html; charset=utf-8", pages[theme])
		})
	}
	go func() { _ = http.Serve(ln, mux) }()

	baseURL := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	goldenDir := filepath.Join("Test", "testdata", "theme_golden")
	update := os.Getenv("SUI_UPDATE_THEME_GOLDEN") == "1"
	if update {
		if err := os.MkdirAll(goldenDir, 0o755); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
	}

	for _, id := range themes {
		actual := filepath.Join(t.TempDir(), id+".png")
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		args := []string{
			"--headless=new",
			"--disable-gpu",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--no-first-run",
			"--no-default-browser-check",
			"--user-data-dir=" + t.TempDir(),
			"--force-device-scale-factor=1",
			"--force-color-profile=srgb",
			"--hide-scrollbars",
			"--window-size=480,320",
			"--virtual-time-budget=20000",
			"--screenshot=" + actual,
			baseURL + id,
		}
		out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("%s: headless screenshot failed: %v\n%s", id, err, string(out))
		}
		got, err := os.ReadFile(actual)
		if err != nil {
			t.Fatalf("%s: screenshot was not written: %v\n%s", id, err, string(out))
		}
		golden := filepath.Join(goldenDir, id+".png")
		if update {
			if err := os.WriteFile(golden, got, 0o644); err != nil {
				t.Fatalf("%s: write golden: %v", id, err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("%s: golden %s missing; run with SUI_UPDATE_THEME_GOLDEN=1", id, golden)
		}
		compareThemePNG(t, id, golden, want, got)
	}
}

// compareThemePNG allows tiny antialiasing/color-profile differences while
// still catching real theme regressions.
func compareThemePNG(t *testing.T, name, goldenPath string, want, got []byte) {
	t.Helper()
	wantImg, err := png.Decode(bytes.NewReader(want))
	if err != nil {
		t.Fatalf("%s: decode golden: %v", name, err)
	}
	gotImg, err := png.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("%s: decode screenshot: %v", name, err)
	}
	wb, gb := wantImg.Bounds(), gotImg.Bounds()
	if wb != gb {
		t.Fatalf("%s: screenshot size %v, golden size %v", name, gb, wb)
	}
	const channelTolerance = 4
	diff := 0
	for y := wb.Min.Y; y < wb.Max.Y; y++ {
		for x := wb.Min.X; x < wb.Max.X; x++ {
			r1, g1, b1, a1 := wantImg.At(x, y).RGBA()
			r2, g2, b2, a2 := gotImg.At(x, y).RGBA()
			d := absInt(int(r1>>8) - int(r2>>8))
			if v := absInt(int(g1>>8) - int(g2>>8)); v > d {
				d = v
			}
			if v := absInt(int(b1>>8) - int(b2>>8)); v > d {
				d = v
			}
			if v := absInt(int(a1>>8) - int(a2>>8)); v > d {
				d = v
			}
			if d > channelTolerance {
				diff++
			}
		}
	}
	limit := (wb.Dx() * wb.Dy()) / 200 // 0.5% of pixels
	if diff > limit {
		t.Fatalf("%s: %d pixels differ from %s (limit %d)", name, diff, goldenPath, limit)
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

const themeShotScriptTemplate = `<script>
(async function () {
  var wait = function (ms) { return new Promise(function (r) { setTimeout(r, ms); }); };
  function previewDoc() {
    var f = document.getElementById('prev');
    return f && f.contentDocument ? f.contentDocument : null;
  }
  async function waitPreview() {
    for (var i = 0; i < 160; i++) {
      var doc = previewDoc();
      if (doc && doc.querySelector('.sui-view')) return doc;
      await wait(100);
    }
    return null;
  }
  try {
    for (var i = 0; i < 160 && !document.getElementById('projprefs-list'); i++) await wait(100);
    await wait(300);
    openProjectPreferences();
    await wait(120);
    var sel = document.getElementById('projprefs-theme');
    sel.value = '__THEME__';
    sel.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(900);
    closeProjectPreferences();
    var doc = await waitPreview();
    if (!doc) throw new Error('preview missing');
    var cs = doc.defaultView.getComputedStyle(doc.documentElement);
    var names = [
      '--sui-bg', '--sui-fg', '--sui-panel', '--sui-panel2', '--sui-border', '--sui-accent',
      '--sui-hover', '--sui-table-hover-bg', '--sui-tooltip-bg', '--sui-tooltip-fg',
      '--sui-media-bg', '--sui-link'
    ];
    var values = names.map(function (n) { return String(cs.getPropertyValue(n) || '').trim(); });
    document.documentElement.style.cssText = 'margin:0;padding:0;background:#000;';
    document.body.innerHTML = '';
    document.body.style.cssText = 'margin:0;padding:0;overflow:hidden;background:#000;';
    var grid = document.createElement('div');
    grid.style.cssText = 'position:fixed;inset:0;display:grid;grid-template-columns:repeat(6,1fr);grid-template-rows:repeat(2,1fr);';
    values.forEach(function (v) {
      var cell = document.createElement('div');
      cell.style.cssText = 'background:' + (v || 'transparent') + ';';
      grid.appendChild(cell);
    });
    document.body.appendChild(grid);
    document.title = 'THEME_SHOT_READY';
  } catch (e) {
    document.title = 'THEME_SHOT_FAIL';
    document.body.textContent = '';
  }
})();
</script>`
