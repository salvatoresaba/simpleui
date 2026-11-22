package tablelayout_test

import . "github.com/salvatoresaba/simpleui"

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestTableLayoutCellEventsE2E checks that a cell onClick runs on empty areas
// but does not steal clicks from a child component, which keeps its own handler.
func TestTableLayoutCellEventsE2E(t *testing.T) {
	browser := FindHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping table layout events test")
	}

	const source = `state cellClicks = 0
state childClicks = 0

fn bumpCell() -> null {
	cellClicks = cellClicks + 1
}

fn bumpChild() -> null {
	childClicks = childClicks + 1
}

view main {
	p "cell {cellClicks} child {childClicks}"
	tablelayout rows: 1 columns: 1 rowHeights: "200px" {
		layoutcell row: 0 col: 0 onClick: bumpCell tooltip: "Cell tip" {
			button "Child" onClick: bumpChild
		}
	}
}`
	index, css, js, diags, ok := GUIBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + tableLayoutEventsScript + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := GUITestMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		WriteRawHandler(w, "text/html; charset=utf-8", page)
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
		"--window-size=800,600",
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
	t.Fatalf("table layout events e2e failed:\n%s", detail)
}

const tableLayoutEventsScript = `<script>
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
  function wait(ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); }
  function status() {
    var p = document.querySelector('p');
    return p ? p.textContent.replace(/\s+/g, ' ').trim() : '';
  }
  try {
    for (var i = 0; i < 100; i++) {
      if (document.querySelector('table.sui-layout-table') && document.querySelector('table.sui-layout-table button')) break;
      await wait(100);
    }
    var cell = document.querySelector('table.sui-layout-table .sui-layout-cell');
    var button = cell && cell.querySelector('button');
    if (!cell || !button) throw new Error('cell or child button not rendered');
    if (status() !== 'cell 0 child 0') throw new Error('unexpected initial state: ' + status());

    cell.dispatchEvent(new PointerEvent('pointerover', { bubbles: true }));
    await wait(350);
    var tip = document.querySelector('.sui-tooltip-bubble');
    if (!tip || tip.hidden || tip.textContent !== 'Cell tip') throw new Error('cell tooltip did not open');
    cell.dispatchEvent(new PointerEvent('pointerout', { bubbles: true }));

    cell.click();
    await wait(350);
    if (status() !== 'cell 1 child 0') throw new Error('empty cell click did not run the cell handler: ' + status());

    button.click();
    await wait(350);
    if (status() !== 'cell 1 child 1') throw new Error('child click did not stay on the child handler: ' + status());

    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
