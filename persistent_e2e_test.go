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

const persistentE2EScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function tableText() {
    var t = document.querySelector('table[name="p"]');
    return t ? (t.textContent || '').replace(/\s+/g, ' ') : '';
  }
  function persistKeys() {
    var out = [];
    for (var i = 0; i < localStorage.length; i++) {
      var k = localStorage.key(i);
      if (k && k.indexOf('sui.table.persist.') === 0) out.push(k);
    }
    return out;
  }
  try {
    for (var i = 0; i < 150 && !document.querySelector('table[name="p"]'); i++) await wait(100);
    if (!document.querySelector('table[name="p"]')) throw new Error('table did not render');
    var phase = sessionStorage.getItem('sui-persist-e2e') || 'edit';
    if (phase === 'edit') {
      // Past the boot guard, with a user interaction, as a real edit would be.
      await wait(2200);
      document.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true }));
      var r = window.__SUI_TBL.display('p', JSON.stringify([['savedA', '2'], ['savedB', '1']]));
      if (!r || r[1]) throw new Error('display failed: ' + (r && r[1]));
      await wait(300);
      var th = document.querySelector('table[name="p"] thead th');
      th.style.width = '123px';
      th.style.minWidth = '123px';
      var s = window.__SUI_TBL.sort('p', 'a', 'desc');
      if (!s || s[1]) throw new Error('sort failed: ' + (s && s[1]));
      await wait(600);
      if (!persistKeys().length) throw new Error('persistent table was not saved');
      if (tableText().indexOf('saved') < 0) throw new Error('saved row is not visible');
      sessionStorage.setItem('sui-persist-e2e', 'reload');
      location.reload();
      return;
    }
    if (phase === 'reload') {
      if (tableText().indexOf('saved') < 0) throw new Error('persisted row was not restored: ' + tableText());
      var restoredTh = document.querySelector('table[name="p"] thead th');
      if (restoredTh.style.width !== '123px') throw new Error('persisted column width was not restored: ' + restoredTh.style.width);
      var firstCell = document.querySelector('table[name="p"] tbody tr td');
      if (!firstCell || firstCell.textContent.trim() !== 'savedB') throw new Error('persisted sort was not restored: ' + (firstCell && firstCell.textContent));
      var c = window.__SUI_TBL.clearPersisted('p');
      if (!c || c[1]) throw new Error('clearPersisted failed: ' + (c && c[1]));
      await wait(300);
      if (tableText().indexOf('src') < 0) throw new Error('source row was not restored: ' + tableText());
      if (persistKeys().length) throw new Error('persistent key still present after reset');
      sessionStorage.removeItem('sui-persist-e2e');
      result(true, '');
      return;
    }
    throw new Error('unexpected phase: ' + phase);
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`

func TestPersistentTableReloadE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping persistent reload e2e")
	}
	src := `view main {
	table name: "p" columns: "a,b" editMode: "persistent" sortable: true {
		tbody {
			trow {
				tcell "src"
				tcell "1"
			}
		}
	}
}`
	prog, dl := ParseSource(src, "persistent_e2e.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><style>" + GenerateCss(prog) +
		"</style></head><body><div id=\"app\"></div><script>" + GenerateJs(prog) + "</script>" +
		persistentE2EScript + "</body></html>"

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=12000", "--dump-dom",
		"http://" + ln.Addr().String() + "/",
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
		end := marker + 500
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("persistent reload e2e failed:\n%s", detail)
}
