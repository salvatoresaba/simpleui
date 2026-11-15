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

func TestOverlayRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping overlay runtime test")
	}
	const source = `css ":root \\{ --sui-tooltip-place: auto; --sui-tooltip-animation: slide; --sui-tooltip-duration: 150ms; \\}"

view main {
	div style: "height: 260px" {
	}
	combobox "Paese" options: "Italia,Francia,Germania" placement: "auto" animation: "slide" duration: "180ms" {
	}
	button "Info" tooltip: "Suggerimento utile"
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + overlayRuntimeScript + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	go func() { _ = http.Serve(ln, mux) }()
	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=800,600",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=20000", "--dump-dom", url,
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
		end := marker + 1600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("overlay runtime e2e failed:\n%s", detail)
}

const overlayRuntimeScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function inside(r) { return r.left >= 0 && r.top >= 0 && r.right <= window.innerWidth && r.bottom <= window.innerHeight; }
  try {
    for (var i = 0; i < 100; i++) {
      if (document.querySelector('.sui-combobox') && document.querySelector('[data-sui-tooltip]')) break;
      await wait(100);
    }
    var combo = document.querySelector('.sui-combobox');
    if (!combo) throw new Error('combobox missing');
    var toggle = combo.querySelector('.sui-combobox-toggle') || combo.querySelector('.sui-combobox-input');
    toggle.click();
    await wait(250);
    var popup = combo.querySelector('.sui-combobox-popup');
    if (!popup || popup.hidden) throw new Error('combobox popup did not open');
    var cr = combo.getBoundingClientRect();
    var pr = popup.getBoundingClientRect();
    if (!inside(pr)) throw new Error('combobox popup outside viewport: ' + JSON.stringify({ left: pr.left, top: pr.top, right: pr.right, bottom: pr.bottom }));
    if (Math.abs(pr.width - cr.width) > 2) throw new Error('combobox popup width mismatch: ' + pr.width + ' vs ' + cr.width);
    if (!popup.hasAttribute('data-sui-anim')) throw new Error('combobox animation missing');
    var inp = combo.querySelector('.sui-combobox-input');
    inp.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    await wait(100);
    if (!popup.hidden) throw new Error('combobox popup did not close on Escape');

    var tip = document.querySelector('[data-sui-tooltip]');
    if (!tip) throw new Error('tooltip target missing');
    tip.focus();
    await wait(300);
    var bubble = document.querySelector('.sui-tooltip-bubble');
    if (!bubble || bubble.hidden) throw new Error('tooltip did not show');
    var br = bubble.getBoundingClientRect();
    if (!inside(br)) throw new Error('tooltip outside viewport: ' + JSON.stringify({ left: br.left, top: br.top, right: br.right, bottom: br.bottom }));
    if (!bubble.hasAttribute('data-sui-anim')) throw new Error('tooltip animation missing');
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    await wait(100);
    if (!bubble.hidden) throw new Error('tooltip did not hide on Escape');
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
