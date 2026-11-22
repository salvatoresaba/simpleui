package sui

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestPrefabLayoutRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping drawer/splitter/dropzone test")
	}
	const source = `state show = false
state files = ""

fn openDrawer() -> null {
	show = true
}

view main {
	button "Open" onClick: openDrawer
	splitter direction: "horizontal" initial: "40%" min: "80px" {
		pane { p "Left" }
		pane { p "Right" }
	}
	drawer "Filters" open: show {
		p "Drawer body"
	}
	dropzone "Files" dropFiles: files accept: ".csv" multiple: true
	p "Files: {files}"
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + prefabLayoutScript + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}

	ln, err := ListenLocal()
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
		end := marker + 1400
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("prefab layout e2e failed:\n%s", detail)
}

const prefabLayoutScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function firstButton(label) {
    var all = document.querySelectorAll('button');
    for (var i = 0; i < all.length; i++) if (all[i].textContent.trim() === label) return all[i];
    return null;
  }
  try {
    for (var i = 0; i < 100; i++) {
      if (document.querySelector('.sui-split-bar') && document.querySelector('.sui-dropzone') && document.querySelector('[data-sui-drawer]')) break;
      await wait(100);
    }
    var bar = document.querySelector('.sui-split-bar');
    var first = bar && bar.parentNode.querySelector('.sui-split-pane');
    if (!bar || !first) throw new Error('splitter not rendered');
    var before = parseFloat(first.style.flexBasis || '0');
    bar.focus();
    bar.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true, cancelable: true }));
    await wait(100);
    var after = parseFloat(first.style.flexBasis || '0');
    if (!(after > before)) throw new Error('splitter did not resize: ' + before + ' -> ' + after);

    var openBtn = firstButton('Open');
    if (!openBtn) throw new Error('drawer open button missing');
    openBtn.click();
    await wait(250);
    var layer = document.querySelector('[data-sui-drawer]');
    if (!layer || layer.getAttribute('data-sui-drawer-open') !== '1') throw new Error('drawer did not open');
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await wait(250);
    layer = document.querySelector('[data-sui-drawer]');
    if (!layer || layer.getAttribute('data-sui-drawer-open') !== '0') throw new Error('drawer did not close on Escape');

    var dz = document.querySelector('.sui-dropzone');
    if (!dz) throw new Error('dropzone not rendered');
    var dt = new DataTransfer();
    dt.items.add(new File(['a,b'], 'rows.csv', { type: 'text/csv' }));
    dz.dispatchEvent(new DragEvent('drop', { bubbles: true, cancelable: true, dataTransfer: dt }));
    await wait(400);
    var status = document.querySelector('.sui-dropzone-status');
    var pageText = document.body.textContent || '';
    if (!status || status.textContent.indexOf('rows.csv') < 0 || pageText.indexOf('Files: rows.csv') < 0) {
      throw new Error('dropzone did not store files: ' + (status ? status.textContent : 'missing status'));
    }
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
