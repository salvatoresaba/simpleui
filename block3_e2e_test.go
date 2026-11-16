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

func TestBlock3RuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping Block 3 runtime test")
	}
	const source = `state slide = 0
state step = 1
state showPalette = false
state last = "none"

fn newProject() -> null {
	last = "new"
}

fn settings() -> null {
	last = "settings"
}

view main {
	carousel value: slide {
		slide { p "A" }
		slide { p "B" }
		slide { p "C" }
	}
	stepper value: step {
		step "One"
		step "Two"
		step "Three"
	}
	commandbar "Search" open: showPalette hotkey: "k" {
		command "New project" shortcut: "Ctrl N" onClick: newProject
		command "Open settings" shortcut: "Ctrl ," onClick: settings
	}
	p "slide {slide} step {step} last {last}"
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + block3Script + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := guiTestMux()
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
		"--no-first-run", "--no-default-browser-check", "--window-size=900,700",
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
	t.Fatalf("Block 3 e2e failed:\n%s", detail)
}

const block3Script = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function pageText() { return (document.body.textContent || '').replace(/\s+/g, ' '); }
  function key(target, k, extra) {
    var ev = new KeyboardEvent('keydown', Object.assign({ key: k, bubbles: true, cancelable: true }, extra || {}));
    target.dispatchEvent(ev);
  }
  try {
    for (var i = 0; i < 100; i++) {
      if (document.querySelector('.sui-carousel') && document.querySelector('.sui-stepper') && document.querySelector('.sui-commandbar')) break;
      await wait(100);
    }
    if (!document.querySelector('.sui-carousel')) throw new Error('carousel not rendered');
    var next = document.querySelector('.sui-carousel-next');
    next.click();
    await wait(300);
    var car = document.querySelector('.sui-carousel');
    if (car.getAttribute('data-sui-carousel-index') !== '1') throw new Error('carousel index did not change: ' + car.getAttribute('data-sui-carousel-index'));
    if (pageText().indexOf('slide 1') < 0) throw new Error('carousel state not bound: ' + pageText());

    var steps = document.querySelectorAll('[data-sui-step]');
    if (steps.length !== 3) throw new Error('stepper steps: ' + steps.length);
    steps[2].click();
    await wait(300);
    if (pageText().indexOf('step 2') < 0) throw new Error('stepper state not bound: ' + pageText());

    key(document, 'k', { ctrlKey: true });
    await wait(350);
    var bar = document.querySelector('[data-sui-commandbar-overlay]');
    if (!bar || bar.getAttribute('data-sui-commandbar-open') !== '1') throw new Error('command palette did not open with Ctrl+K');
    var input = bar.querySelector('.sui-commandbar-input');
    if (!input) throw new Error('command input missing');
    input.value = 'settings';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(150);
    var visible = bar.querySelectorAll('.sui-command:not([hidden])');
    if (visible.length !== 1) throw new Error('command filter returned ' + visible.length + ' rows');
    key(input, 'Enter');
    await wait(350);
    if (pageText().indexOf('last settings') < 0) throw new Error('command action did not run: ' + pageText());
    bar = document.querySelector('[data-sui-commandbar-overlay]');
    if (bar.getAttribute('data-sui-commandbar-open') !== '0') throw new Error('palette did not close after command');

    key(document, 'k', { ctrlKey: true });
    await wait(300);
    bar = document.querySelector('[data-sui-commandbar-overlay]');
    if (bar.getAttribute('data-sui-commandbar-open') !== '1') throw new Error('palette did not reopen');
    key(document, 'Escape');
    await wait(300);
    bar = document.querySelector('[data-sui-commandbar-overlay]');
    if (bar.getAttribute('data-sui-commandbar-open') !== '0') throw new Error('palette did not close on Escape');
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
