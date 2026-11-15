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

// TestRuntimeRerenderPreservesStateE2E checks that a state update re-renders
// the view without losing focus, media, select values or scroll position.
func TestRuntimeRerenderPreservesStateE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping rerender preservation test")
	}
	const source = `state name = "Ada"
state pick = "b"
state open = false

fn toggleOpen() -> null {
	open = open == true ? false : true
}

css ".tall \{ height: 2200px; \}"

view main {
	col {
		input "Name" value: name
		select "Pick" value: pick {
			option "A" value: "a"
			option "B" value: "b"
			option "C" value: "c"
		}
		img src: "pixel.png" alt: "pixel"
		div class: "tall" {
			p "Tall content"
		}
		button "Toggle at bottom" onClick: toggleOpen
		if open {
			p "Panel open"
		}
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
	page := html[:cut] + runtimePreserveScript + html[cut:]
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(page), 0o644); err != nil {
		t.Fatalf("write page: %v", err)
	}
	pixel, err := os.ReadFile(filepath.Join("examples", "all_controls_animations", "assets", "preview.png"))
	if err != nil {
		t.Fatalf("read pixel fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pixel.png"), pixel, 0o644); err != nil {
		t.Fatalf("write pixel: %v", err)
	}
	srv := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer srv.Close()
	pageURL := srv.URL + "/index.html"
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=1000,800",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=8000", "--dump-dom", pageURL,
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
		end := marker + 900
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("rerender preservation e2e failed:\n%s", detail)
}

const runtimePreserveScript = `<script>
window.addEventListener('error', function (e) {
  var el = document.getElementById('e2e-result');
  if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
  el.textContent = 'FAIL: script error: ' + ((e && e.message) || 'unknown');
});
</script><script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  try {
    for (var i = 0; i < 100 && !document.querySelector('.sui-input'); i++) await wait(50);
    var input = document.querySelector('.sui-input');
    var select = document.querySelector('.sui-select');
    var img = document.querySelector('img');
    var toggle = Array.prototype.find.call(document.querySelectorAll('button'), function (b) { return (b.textContent || '').indexOf('Toggle at bottom') >= 0; });
    if (!input || !select || !img || !toggle) throw new Error('controls missing');
    for (var w = 0; w < 60 && !(img.complete && img.naturalWidth); w++) await wait(50);
    if (!(img.complete && img.naturalWidth)) throw new Error('image fixture did not load: ' + img.outerHTML + ' complete=' + img.complete);

    input.focus();
    input.value = 'Ada Lovelace';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(150);

    var active = document.activeElement;
    if (!active || active.tagName !== 'INPUT') throw new Error('focus lost after re-render');
    if (active.selectionStart !== 'Ada Lovelace'.length) throw new Error('caret lost after re-render');
    if (!active.isConnected) throw new Error('focused input was not restored');

    var imgAfter = document.querySelector('img');
    if (imgAfter !== img) throw new Error('image node was replaced');
    if (!imgAfter.naturalWidth) throw new Error('image lost its loaded state');

    select.value = 'c';
    select.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(150);
    var selectAfter = document.querySelector('.sui-select');
    if (selectAfter.value !== 'c') throw new Error('select value not reflected after re-render: ' + selectAfter.value);

    window.scrollTo(0, document.documentElement.scrollHeight - 600);
    await wait(80);
    var scrollBefore = window.scrollY;
    toggle.scrollIntoView({ block: 'center' });
    await wait(50);
    scrollBefore = window.scrollY;
    toggle.click();
    await wait(150);
    var imgAfter2 = document.querySelector('img');
    if (imgAfter2 !== img) throw new Error('image node replaced after button update');
    if (Math.abs(window.scrollY - scrollBefore) > 4) throw new Error('scroll position changed: ' + window.scrollY + ' vs ' + scrollBefore);
    if (document.body.innerText.indexOf('Panel open') < 0) throw new Error('conditional content did not update');

    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
