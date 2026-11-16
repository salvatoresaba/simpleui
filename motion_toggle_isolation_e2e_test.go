package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestToggleMotionIsolationE2E verifies that a re-render caused by one toggle
// does not replay the enter/scale effect on the other toggles.
func TestToggleMotionIsolationE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping toggle motion isolation test")
	}
	const source = `state a = true
state b = true

view main {
	col {
		checkbox "Alpha" checked: a animation: "scale" duration: "200ms"
		switch "Beta" checked: b animation: "scale" duration: "200ms"
		radio "Gamma" checked: true animation: "scale" duration: "200ms"
	}
}`
	ok, html, msgs := renderSui(source)
	if !ok {
		t.Fatalf("render failed: %v", StringifyJson(msgs))
	}
	cut := strings.LastIndex(html, "</body>")
	if cut < 0 {
		t.Fatal("compiled page has no closing body")
	}
	page := html[:cut] + toggleIsolationScript + html[cut:]
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(page), 0o644); err != nil {
		t.Fatalf("write page: %v", err)
	}
	srv := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=900,700",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=6000", "--dump-dom",
		srv.URL + "/index.html",
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
		end := marker + 600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("toggle isolation e2e failed:\n%s", detail)
}

const toggleIsolationScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function state(el) {
    var cs = getComputedStyle(el);
    return { transform: cs.transform, opacity: cs.opacity, style: el.getAttribute('style') || '' };
  }
  try {
    for (var i = 0; i < 80 && !document.querySelector('.sui-checkbox'); i++) await wait(50);
    var cb = document.querySelector('.sui-checkbox');
    var sw = document.querySelector('.sui-switch');
    var rd = document.querySelector('input[type=radio]');
    if (!cb || !sw || !rd) throw new Error('toggles missing');

    var before = { sw: state(sw), rd: state(rd) };
    cb.click();
    await wait(80);
    var swAfter = document.querySelector('.sui-switch');
    var rdAfter = document.querySelector('input[type=radio]');
    if (!cb.checked) { /* unchecked is expected */ }
    if (state(swAfter).style !== '' || state(swAfter).transform !== 'none') throw new Error('switch moved when the checkbox changed');
    if (state(rdAfter).style !== '' || state(rdAfter).transform !== 'none') throw new Error('radio moved when the checkbox changed');

    var cbAfter = document.querySelector('.sui-checkbox');
    var swBefore2 = state(document.querySelector('.sui-switch'));
    cbAfter.click();
    await wait(80);
    if (state(document.querySelector('.sui-switch')).transform !== swBefore2.transform) throw new Error('switch changed on a second checkbox click');

    var swLive = document.querySelector('.sui-switch');
    var cbBefore = state(document.querySelector('.sui-checkbox'));
    swLive.click();
    await wait(80);
    if (state(document.querySelector('.sui-checkbox')).transform !== cbBefore.transform) throw new Error('checkbox moved when the switch changed');

    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
