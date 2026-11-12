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

// TestTableLayoutRuntimeResponsiveE2E renders the compiled app (not the
// designer preview) at a narrow viewport and checks that responsive stacking is
// active and the cells are really stacked.
func TestTableLayoutRuntimeResponsiveE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping table layout responsive test")
	}

	const source = `view main {
	title "Responsive e2e"
	tablelayout rows: 1 columns: 2 responsive: true {
		layoutcell row: 0 col: 0 {
			card "A"
		}
		layoutcell row: 0 col: 1 {
			card "B"
		}
	}
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	html := CombineSingleHTML(index, css, js, false)
	page := html
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + tableLayoutResponsiveScript + page[idx:]
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
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--no-first-run",
		"--no-default-browser-check",
		"--window-size=420,800",
		"--user-data-dir=" + t.TempDir(),
		"--virtual-time-budget=20000",
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
	t.Fatalf("responsive e2e failed:\n%s", detail)
}

const tableLayoutResponsiveScript = `<script>
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
  function wait(ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); }
  try {
    var table = null;
    for (var i = 0; i < 100; i++) {
      table = document.querySelector('table.sui-layout-table');
      if (table) break;
      await wait(100);
    }
    if (!table) throw new Error('layout table did not render');
    if (window.innerWidth > 640) throw new Error('viewport is not narrow: ' + window.innerWidth);
    if (!table.classList.contains('sui-layout-responsive')) throw new Error('responsive class missing');
    var cells = Array.from(table.querySelectorAll('tbody > tr > .sui-layout-cell'));
    if (cells.length < 2) throw new Error('expected at least two cells');
    var displays = cells.map(function (c) { return window.getComputedStyle(c).display; });
    if (displays.some(function (d) { return d !== 'block'; })) throw new Error('cells are not stacked: ' + displays.join(','));
    var rects = cells.map(function (c) { return c.getBoundingClientRect(); });
    if (rects[1].top < rects[0].bottom - 2) throw new Error('second cell is not below the first');
    if (Math.abs(rects[0].width - rects[1].width) > 4) throw new Error('stacked cells have different widths');
    if (rects[0].width < window.innerWidth * 0.6) throw new Error('stacked cell is too narrow');
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
