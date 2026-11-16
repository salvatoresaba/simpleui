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

// TestTableLayoutPresetEmptyE2E checks that the preset combo appears only for
// an empty table and that applying a preset normalizes residual spans.
func TestTableLayoutPresetEmptyE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping table layout preset test")
	}

	const source = `view main {
	title "Preset e2e"
	tablelayout rows: 2 columns: 2 {
		layoutcell row: 0 col: 0 colspan: 2
		layoutcell row: 1 col: 0
		layoutcell row: 1 col: 1
	}
}`
	page := injectGuiTestScript(tableLayoutPresetScript)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	var lastHTML string
	mux := guiTestMux()
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
	mux.HandleFunc("POST /__css__/setbatch", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		res := guiCssSetBatch(r.PostFormValue("css"), r.PostFormValue("class"), r.PostFormValue("decls"))
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
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
	t.Fatalf("table layout preset e2e failed:\n%s", detail)
}

const tableLayoutPresetScript = `<script>
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
  function frameDoc() {
    var f = document.getElementById('prev');
    return f && f.contentDocument ? f.contentDocument : null;
  }
  function acc(title) {
    return Array.from(document.querySelectorAll('#inspector .acc')).find(function (a) {
      var h = a.querySelector('.acchead span');
      return h && h.textContent.trim() === title;
    });
  }
  async function waitPreview() {
    for (var i = 0; i < 200; i++) {
      var doc = frameDoc();
      if (doc && doc.querySelector('table.sui-layout-table') && doc.querySelector('.sui-layout-pick-table')) return doc;
      await wait(100);
    }
    return null;
  }
  try {
    for (var i = 0; i < 200 && !document.querySelector('#nodes .node'); i++) await wait(100);
    await wait(300);
    var doc = await waitPreview();
    if (!doc) throw new Error('preview did not render');
    doc.querySelector('.sui-layout-pick-table').click();
    await wait(500);
    var section = acc('Table layout');
    if (!section) throw new Error('Table layout section missing');
    var presetRow = Array.from(section.querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'preset';
    });
    if (!presetRow) throw new Error('preset must be visible for an empty table');
    var select = presetRow.querySelector('select');
    if (!select) throw new Error('preset select missing');
    select.value = '0';
    select.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(800);
    doc = frameDoc();
    var table = doc.querySelector('table.sui-layout-table');
    var cells = Array.from(table.querySelectorAll('tbody > tr > td.sui-layout-cell'));
    if (cells.length !== 4) throw new Error('preset must leave four cells, got ' + cells.length);
    if (cells.some(function (c) { return (c.getAttribute('colspan') || '1') !== '1' || (c.getAttribute('rowspan') || '1') !== '1'; })) {
      throw new Error('preset must reset residual spans');
    }
    var rows = table.querySelectorAll('tbody > tr').length;
    var cols = table.querySelectorAll('tbody > tr:first-child > td').length;
    if (rows !== 2 || cols !== 2) throw new Error('preset grid is ' + rows + 'x' + cols);
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
