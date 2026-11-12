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

func TestDropdownPopoverRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping dropdown/popover test")
	}
	const source = `state action = "none"

fn edit() -> null {
	action = "edit"
}

view main {
	p "Action: {action}"
	dropdown "Actions" position: "right" {
		dropdownitem "Edit" onClick: edit
	}
	popover "Info" position: "bottom" {
		p "Popover content"
	}
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + dropdownPopoverScript + page[idx:]
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
		end := marker + 1200
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("dropdown/popover e2e failed:\n%s", detail)
}

const dropdownPopoverScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function status() { var p = document.querySelector('p'); return p ? p.textContent.replace(/\s+/g, ' ').trim() : ''; }
  try {
    for (var i = 0; i < 100; i++) {
      if (document.querySelector('.sui-menu-button') && document.querySelector('.sui-popover-trigger')) break;
      await wait(100);
    }
    var toggle = document.querySelector('.sui-menu-button');
    var popTrigger = document.querySelector('.sui-popover-trigger');
    var menu = toggle && toggle.closest('.sui-menu');
    var pop = popTrigger && popTrigger.closest('[data-sui-popover]');
    var panel = pop && pop.querySelector('.sui-popover-panel');
    if (!toggle || !popTrigger || !menu || !pop || !panel) throw new Error('controls not rendered');
    if (status() !== 'Action: none') throw new Error('unexpected initial state: ' + status());

    toggle.click();
    await wait(100);
    if (!menu.classList.contains('open')) throw new Error('dropdown did not open');
    var popup = menu.querySelector('.sui-menu-popup');
    var rect = popup.getBoundingClientRect();
    if (rect.width < 1 || rect.height < 1) throw new Error('dropdown popup has no size');
    if (rect.left < 0 || rect.top < 0 || rect.right > window.innerWidth || rect.bottom > window.innerHeight) {
      throw new Error('dropdown popup outside viewport: ' + JSON.stringify({ left: rect.left, top: rect.top, right: rect.right, bottom: rect.bottom, vw: window.innerWidth, vh: window.innerHeight }));
    }
    var item = menu.querySelector('.sui-menuitem');
    item.click();
    await wait(300);
    if (status() !== 'Action: edit') throw new Error('dropdown item handler did not run: ' + status());
    var menu2 = document.querySelector('.sui-menu');
    if (menu2 && menu2.classList.contains('open')) throw new Error('dropdown did not close after item');

    popTrigger = document.querySelector('.sui-popover-trigger');
    pop = popTrigger && popTrigger.closest('[data-sui-popover]');
    panel = pop && pop.querySelector('.sui-popover-panel');
    if (!popTrigger || !pop || !panel) throw new Error('popover missing after re-render');
    popTrigger.click();
    await wait(100);
    if (!pop.classList.contains('open') || panel.hidden) throw new Error('popover did not open');
    document.body.click();
    await wait(100);
    if (pop.classList.contains('open') || !panel.hidden) throw new Error('popover did not close on outside click');
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
