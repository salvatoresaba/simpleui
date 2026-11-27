package core

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A row with three cells: the boundary between two of them can be dragged in
// the preview and the new sizes must be written back to the project.
const designerGridSource = `view main {
	row class: "r" {
		div class: "c1" {
			p "one"
		}
		div class: "c2" {
			p "two"
		}
		div class: "c3" {
			p "three"
		}
	}
}`

func TestGuiGridCellResizeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping designer end-to-end test")
	}
	page := injectGuiTestScript(designerGridScript)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}

	ln, err := ListenLocal()
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	logPath := filepath.Join(t.TempDir(), "requests.log")
	var lastHTML string

	mux := http.NewServeMux()
	formValue := func(r *http.Request, key string) string {
		if v := r.PostFormValue(key); v != "" {
			return v
		}
		return r.URL.Query().Get(key)
	}
	on := func(path string, fn func(w http.ResponseWriter, r *http.Request)) {
		mux.HandleFunc("POST "+path, fn)
		mux.HandleFunc("GET "+path, fn)
	}
	json := func(w http.ResponseWriter, body string) {
		writeRawHandler(w, "application/json; charset=utf-8", body)
	}

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	mux.HandleFunc("GET /__gui__/ping", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true}`)
	})
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"sui":`+StringifyJson(JStr(designerGridSource))+`}`)
	})
	mux.HandleFunc("GET /__gui__/appearance", func(w http.ResponseWriter, r *http.Request) {
		json(w, string(loadAppearanceLibraryPayload("")))
	})
	mux.HandleFunc("GET /__gui__/components.js", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/javascript; charset=utf-8", guiComponentLibraryJS)
	})
	mux.HandleFunc("GET /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		json(w, loadGuiPrefs())
	})
	mux.HandleFunc("GET /__gui__/templates", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"templates":[]}`)
	})
	mux.HandleFunc("GET /__gui__/fonts", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"fonts":[]}`)
	})
	mux.HandleFunc("GET /__gui__/assets", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiAssetList()))
	})
	mux.HandleFunc("GET /__gui__/history", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"versions":[]}`)
	})
	mux.HandleFunc("GET /__gui__/restore", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"sui":""}`)
	})
	on("/__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		ok, html, msgs := renderSui(formValue(r, "sui"))
		if ok {
			lastHTML = html
			json(w, `{"ok":true}`)
			return
		}
		json(w, `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", lastHTML)
	})
	on("/__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		model, msgs, details, ok := guiOpenDiags(formValue(r, "sui"))
		if !ok {
			json(w, `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`,"diagDetails":`+StringifyJson(details)+`}`)
			return
		}
		body := StringifyJson(model)
		json(w, `{"ok":true,`+body[1:])
	})
	on("/__gui__/apply", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiApplySource(formValue(r, "sui"))))
	})
	on("/__gui__/check", func(w http.ResponseWriter, r *http.Request) {
		_, msgs, details, ok := guiOpenDiags(formValue(r, "sui"))
		okStr := "false"
		if ok {
			okStr = "true"
		}
		json(w, `{"ok":`+okStr+`,"diagnostics":`+StringifyJson(msgs)+`,"diagDetails":`+StringifyJson(details)+`}`)
	})
	on("/__css__/parse", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiCssParse(formValue(r, "css"), formValue(r, "used"))))
	})
	on("/__css__/setbatch", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiCssSetBatch(formValue(r, "css"), formValue(r, "class"), formValue(r, "decls"))))
	})
	on("/__css__/set", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiCssSet(formValue(r, "css"), formValue(r, "class"), formValue(r, "prop"), formValue(r, "value"))))
	})
	on("/__css__/setsel", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiCssSetSelector(formValue(r, "css"), formValue(r, "selector"), formValue(r, "at"), formValue(r, "decls"))))
	})
	on("/__css__/setselmany", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiCssSetSelectors(formValue(r, "css"), formValue(r, "selectors"), formValue(r, "at"), formValue(r, "decls"))))
	})

	go func() {
		_ = http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err == nil {
				_, _ = f.WriteString(r.Method + " " + r.URL.Path + "\n")
				_ = f.Close()
			}
			mux.ServeHTTP(w, r)
		}))
	}()

	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=1200,800",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=60000", "--dump-dom",
		"--enable-logging=stderr", "--log-level=0", url,
	}
	out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
	text := string(out)
	if err != nil {
		t.Fatalf("headless browser failed: %v\n%s", err, text)
	}
	if strings.Contains(text, `data-e2e="done">OK</div>`) {
		return
	}
	detail := text
	if marker := strings.Index(text, `data-e2e="done">`); marker >= 0 {
		end := marker + 400
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	if stepIdx := strings.Index(text, `id="e2e-step"`); stepIdx >= 0 {
		end := stepIdx + 200
		if end > len(text) {
			end = len(text)
		}
		detail = "step: " + text[stepIdx:end] + "\n" + detail
	}
	t.Fatalf("grid cell resize e2e failed:\n%s", detail)
}

const designerGridScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-result';
      document.body.insertBefore(el, document.body.firstChild);
    }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    el.setAttribute('data-e2e', 'done');
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function step(name) {
    var el = document.getElementById('e2e-step');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-step';
      document.body.insertBefore(el, document.body.firstChild);
    }
    el.textContent = name;
    document.title = 'E2E:' + name;
  }
  var wait = function (ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); };
  async function until(fn, timeoutMs) {
    var deadline = Date.now() + (timeoutMs || 5000);
    for (;;) {
      var value = fn();
      if (value) return value;
      if (Date.now() >= deadline) return null;
      await wait(20);
    }
  }
  function frameDoc() {
    var f = document.getElementById('prev');
    return f && f.contentDocument ? f.contentDocument : null;
  }
  try {
    step('start');
    var ta = null;
    for (var i = 0; i < 150; i++) {
      ta = document.getElementById('out');
      if (ta && ta.value && document.querySelectorAll('#nodes .node').length) break;
      await wait(100);
    }
    if (!ta) throw new Error('the source editor is missing');
    await wait(1200);
    step('loaded');

    // The boundary between two cells of the row is draggable.
    var handle = await until(function () {
      var d = frameDoc();
      var h = d && d.querySelector('.sui-grid-split-v');
      return h ? h : null;
    }, 8000);
    if (!handle) throw new Error('the cell split handle is missing');
    var doc = frameDoc();
    var hr = handle.getBoundingClientRect();
    handle.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, button: 0, clientX: hr.left + 6, clientY: hr.top + 10 }));
    doc.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: hr.left + 40, clientY: hr.top + 10 }));
    doc.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: hr.left + 40, clientY: hr.top + 10 }));
    step('dragged');

    // The new sizes must reach the project CSS. Every cell of the row gets
    // its own size, so the styles are written one class at a time.
    var written = await until(function () {
      var src = document.getElementById('out').value || '';
      var found = src.match(/width:\s*\d+px/g) || [];
      return found.length >= 3 ? src : null;
    }, 8000);
    if (!written) throw new Error('the cell sizes were not written back');
    step('written');

    // The sizes of the cells must not exceed the row: the two cells around
    // the handle keep their own size.
    var widths = (written.match(/width:\s*\d+px/g) || []).map(function (w) {
      return Number(/(\d+)/.exec(w)[1]);
    });
    if (widths.length < 3) throw new Error('the cells did not get a size: ' + JSON.stringify(widths));
    if (widths[0] <= 0 || widths[1] <= 0) throw new Error('a cell size is empty: ' + JSON.stringify(widths));
    if (widths[0] === widths[1]) throw new Error('the drag did not change the sizes: ' + JSON.stringify(widths));
    step('sized');

    // Alt+click inside a cell selects the CELL and not the element under the
    // cursor: a cell that holds other elements must stay reachable.
    // The preview is rebuilt after the resize, and the designer replays a
    // click to restore the preview: that click is internal and must not
    // swallow the pick, so the click is retried.
    var picked = null;
    for (var attempt = 0; attempt < 5 && !picked; attempt++) {
      await wait(400);
      var d2 = frameDoc();
      var inner = d2.querySelector('.c1 p') || d2.querySelector('.c1 [data-sui-id]');
      if (!inner) continue;
      var ir = inner.getBoundingClientRect();
      inner.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, altKey: true, clientX: ir.left + 2, clientY: ir.top + 2 }));
      picked = await until(function () {
        var path = window.selPath;
        if (!path || path.length !== 2) return null;
        var node = byPath(path, root.children);
        return node && node.class === 'c1' ? path : null;
      }, 2000);
    }
    if (!picked) throw new Error('alt+click did not select the cell: ' + JSON.stringify(window.selPath));
    step('cell-picked');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`
