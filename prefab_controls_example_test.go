package sui

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const prefabExamplePath = "examples/prefab_controls/app.sui"

func TestPrefabControlsExample(t *testing.T) {
	data, err := os.ReadFile(prefabExamplePath)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	prog, dl := ParseSource(string(data), prefabExamplePath)
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	html := GenerateHtmlPage(prog)

	jsMarkers := map[string]string{
		"dropdown":        `data-sui-dropdown`,
		"popover":         `data-sui-popover`,
		"drawer":          `data-sui-drawer`,
		"splitter":        `data-sui-splitter`,
		"dropzone":        `sui-dropzone`,
		"timeline":        `sui-timeline`,
		"empty state":     `sui-empty`,
		"skeleton":        `sui-skeleton`,
		"carousel":        `sui-carousel`,
		"stepper":         `sui-stepper`,
		"command palette": `data-sui-commandbar`,
	}
	for name, marker := range jsMarkers {
		if !strings.Contains(js, marker) {
			t.Errorf("%s marker %q missing from generated JS", name, marker)
		}
	}
	htmlMarkers := map[string]string{
		"dropdown":        `sui-menu`,
		"popover":         `sui-popover`,
		"drawer":          `data-sui-drawer`,
		"splitter":        `sui-splitter`,
		"dropzone":        `sui-dropzone`,
		"timeline":        `sui-timeline`,
		"carousel":        `sui-carousel`,
		"stepper":         `sui-stepper`,
		"command palette": `sui-commandbar`,
	}
	for name, marker := range htmlMarkers {
		if !strings.Contains(html, marker) {
			t.Errorf("%s marker %q missing from static HTML", name, marker)
		}
	}
	for _, state := range []string{`setAction`, `setShowFilters`, `setSlide`, `setStep`, `setShowPalette`, `setFiles`, `setFolder`} {
		if !strings.Contains(js, state) {
			t.Errorf("state setter %q missing", state)
		}
	}
}

func TestConditionalBranchWiringGuard(t *testing.T) {
	src := `state show = true

fn a() -> null {
}

fn b() -> null {
}

fn c() -> null {
}

view main {
	if show {
		button "A" onClick: a
	} else {
		button "B" onClick: b
	}
	button "C" onClick: c
}`
	prog, dl := ParseSource(src, "conditional.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	// Only one branch exists after a render. Each wiring block must be guarded
	// so the absent branch cannot throw and skip the handlers that follow.
	for _, want := range []string{
		"if (e0) {",
		"if (e1) {",
		"if (e2) {",
		"e0.onclick = a",
		"e1.onclick = b",
		"e2.onclick = c",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("conditional wiring missing %q", want)
		}
	}
}

func TestPrefabControlsExampleE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping prefab controls example test")
	}
	data, err := os.ReadFile(prefabExamplePath)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	index, css, js, diags, ok := guiBuildFiles(string(data))
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + prefabExampleScript + page[idx:]
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
		"--no-first-run", "--no-default-browser-check", "--window-size=1100,900",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=30000", "--dump-dom", url,
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
		end := marker + 1800
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("prefab controls example e2e failed:\n%s", detail)
}

const prefabExampleScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function text() { return ((document.body.innerText || document.body.textContent) || '').replace(/\s+/g, ' '); }
  function button(label) {
    var all = document.querySelectorAll('button');
    for (var i = 0; i < all.length; i++) if (all[i].textContent.trim() === label) return all[i];
    return null;
  }
  function key(target, k, extra) {
    target.dispatchEvent(new KeyboardEvent('keydown', Object.assign({ key: k, bubbles: true, cancelable: true }, extra || {})));
  }
  try {
    for (var i = 0; i < 120; i++) {
      if (document.querySelector('[data-sui-dropdown]') && document.querySelector('.sui-commandbar') && document.querySelector('.sui-dropzone')) break;
      await wait(100);
    }

    // Dropdown
    var dd = document.querySelector('[data-sui-dropdown]');
    if (!dd) throw new Error('dropdown missing');
    dd.querySelector('.sui-menu-button').click();
    await wait(120);
    var item = dd.querySelector('.sui-menuitem');
    if (!item) throw new Error('dropdown item missing');
    item.click();
    await wait(300);
    if (text().indexOf('Ultima azione: modifica') < 0) throw new Error('dropdown action not applied: ' + text());

    // Popover
    var pop = document.querySelector('[data-sui-popover]');
    if (!pop) throw new Error('popover missing');
    pop.querySelector('.sui-popover-trigger').click();
    await wait(150);
    if (!pop.classList.contains('open')) throw new Error('popover did not open');
    var popBtn = pop.querySelector('.sui-popover-panel button');
    if (!popBtn) throw new Error('popover internal button missing');
    popBtn.click();
    await wait(300);
    var pop2 = document.querySelector('[data-sui-popover]');
    pop2.querySelector('.sui-popover-trigger').click();
    await wait(150);
    if (text().indexOf('Click interni: 1') < 0) throw new Error('popover content action not applied');
    key(document, 'Escape');
    await wait(150);
    if (document.querySelector('[data-sui-popover]').classList.contains('open')) throw new Error('popover did not close on Escape');

    // Drawer
    var openDrawer = button('Apri filtri');
    if (!openDrawer) throw new Error('drawer open button missing');
    openDrawer.click();
    await wait(300);
    var drawer = document.querySelector('[data-sui-drawer]');
    if (drawer.getAttribute('data-sui-drawer-open') !== '1') throw new Error('drawer did not open');
    key(document, 'Escape');
    await wait(300);
    drawer = document.querySelector('[data-sui-drawer]');
    if (drawer.getAttribute('data-sui-drawer-open') !== '0') throw new Error('drawer did not close on Escape');

    // Splitter
    var bar = document.querySelector('.sui-split-bar');
    var pane = bar && bar.parentNode.querySelector('.sui-split-pane');
    if (!bar || !pane) throw new Error('splitter missing');
    var before = parseFloat(pane.style.flexBasis || '0');
    bar.focus();
    key(bar, 'ArrowRight');
    await wait(120);
    var after = parseFloat(pane.style.flexBasis || '0');
    if (!(after > before)) throw new Error('splitter did not resize: ' + before + ' -> ' + after);

    // Dropzone
    var dz = document.querySelector('.sui-dropzone');
    if (!dz) throw new Error('dropzone missing');
    var dt = new DataTransfer();
    dt.items.add(new File(['a,b'], 'rows.csv', { type: 'text/csv' }));
    dz.dispatchEvent(new DragEvent('drop', { bubbles: true, cancelable: true, dataTransfer: dt }));
    await wait(450);
    var status = document.querySelector('.sui-dropzone-status');
    if (!status || status.textContent.indexOf('rows.csv') < 0) throw new Error('dropzone did not store the file');

    // Timeline
    if (document.querySelectorAll('.sui-timeline-item').length !== 4) throw new Error('timeline item count mismatch');
    if (!document.querySelector('.sui-tl-done') || !document.querySelector('.sui-tl-active')) throw new Error('timeline status classes missing');

    // Empty state
    var add = button('Crea progetto');
    if (!add) throw new Error('empty state action missing');
    add.click();
    await wait(300);
    if (text().indexOf('Progetti creati: 1') < 0) throw new Error('empty state action not applied: ' + text());
    if (document.querySelector('.sui-empty')) throw new Error('empty state did not disappear');

    // Skeleton
    var toggle = button('Toggle caricamento');
    if (!toggle) throw new Error('skeleton toggle missing');
    toggle.click();
    await wait(300);
    if (document.querySelector('.sui-skeleton')) throw new Error('skeleton still visible after toggle');
    if (text().indexOf('Contenuto caricato') < 0) throw new Error('loaded content missing');

    // Carousel
    var next = document.querySelector('.sui-carousel-next');
    if (!next) throw new Error('carousel next missing');
    next.click();
    await wait(350);
    var car = document.querySelector('.sui-carousel');
    if (car.getAttribute('data-sui-carousel-index') !== '1') throw new Error('carousel index not updated');
    if (text().indexOf('Slide 2 di 3') < 0) throw new Error('carousel label not updated');

    // Stepper
    var steps = document.querySelectorAll('[data-sui-step]');
    if (steps.length !== 3) throw new Error('stepper step count mismatch');
    steps[2].click();
    await wait(350);
    if (text().indexOf('Passo 3') < 0) throw new Error('stepper index not updated');

    // Command palette
    key(document, 'k', { ctrlKey: true });
    await wait(400);
    var cmdbar = document.querySelector('[data-sui-commandbar-overlay]');
    if (!cmdbar || cmdbar.getAttribute('data-sui-commandbar-open') !== '1') throw new Error('command palette did not open');
    var input = cmdbar.querySelector('.sui-commandbar-input');
    input.value = 'impostazioni';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(150);
    var visible = cmdbar.querySelectorAll('.sui-command:not([hidden])');
    if (visible.length !== 1) throw new Error('command filter returned ' + visible.length + ' rows');
    key(input, 'Enter');
    await wait(400);
    if (text().indexOf('Ultimo comando: impostazioni') < 0) throw new Error('command action not applied');
    cmdbar = document.querySelector('[data-sui-commandbar-overlay]');
    if (cmdbar.getAttribute('data-sui-commandbar-open') !== '0') throw new Error('command palette did not close');

    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
