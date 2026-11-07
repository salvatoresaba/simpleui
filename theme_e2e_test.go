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
    var nameInput = document.getElementById('projprefs-theme-name');
    if (nameInput) nameInput.value = 'QA theme';
    var payload = buildProjectThemePayload();
    applyImportedThemeObject(JSON.parse(JSON.stringify(payload)));
    var importedName = normalizeThemePayload(payload).name;
    var fuzzyKey = chartPresetKey({ styleExpr: 'cssSafe(' + JSON.stringify(chartPresetCss('theme-standard-light') + '; --sui-chart-extra-check: 1') + ')' });
    newSui();
    await wait(80);
    addEl('chart');
    await wait(200);
    var chartNode = byPath(selPath, root.children);
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
      result(false, failed.join(',') + ' ' + JSON.stringify({ light: light, dark: dark, themePresets: themePresets.length, variants: STYLE_VARIANT_DEFS.length }));
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
		page := strings.Replace(guiHTML, "</body>", script+"</body>", 1)
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
	mux := http.NewServeMux()
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

	goldenDir := filepath.Join("testdata", "theme_golden")
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
      '--sui-media-bg', '--sui-divider-hover-color'
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
