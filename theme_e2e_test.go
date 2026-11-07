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

// theme_e2e_test.go drives the real designer in a headless browser and checks
// the rendered theme through computed styles and CSS variables. It covers the
// light/dark page bodies, base elements, tooltips, media, hover presets, chart
// theme presets and the custom theme export/import payload.
func TestGuiThemeVisualE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping theme visual regression test")
	}

	const source = `view main {
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

	page := strings.Replace(guiHTML, "</body>", themeE2EScript+"</body>", 1)
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
      document.body.appendChild(el);
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
    var payload = buildProjectThemePayload();
    applyImportedThemeObject(JSON.parse(JSON.stringify(payload)));
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
      exportPayload: !!(payload.page && payload.page['--sui-bg'] && payload.components && payload.tooltip && payload.charts),
      importRejectsUnknownKey: invalidRejected,
      importRejectsUnknownVariable: unknownRejected,
      importRejectsCssInjection: unsafeRejected
    };
    var failed = Object.keys(checks).filter(function (k) { return !checks[k]; });
    if (failed.length) {
      result(false, failed.join(',') + ' ' + JSON.stringify({ light: light, dark: dark, themePresets: themePresets.length, variants: STYLE_VARIANT_DEFS.length }));
      return;
    }
    result(true, '');
  } catch (e) {
    result(false, (e && (e.stack || e.message)) || String(e));
  }
})();
</script>`
