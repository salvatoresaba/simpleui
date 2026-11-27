package core

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// A row that declares its columns and the size of every cell: the designer
// must read them from the model, show them in the Layout section of the
// inspector and refuse a set of widths that does not fit the row.
const designerRowLayoutSource = `view main {
	row columns: "2" columnWidths: "60%, 40%" class: "r" {
		div class: "c1" {
			p "one"
		}
		div class: "c2" {
			p "two"
		}
	}
}`

// A row that declares its columns and its cells: dragging the boundary
// between two cells must write the new size in the container.
const designerRowResizeSource = `view main {
	row columns: "3" class: "r" {
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

func TestGuiRowLayoutInspectorE2E(t *testing.T) {
	runDesignerE2E(t, designerRowLayoutSource, designerRowLayoutScript)
}

func TestGuiRowResizeWritesTheContainerE2E(t *testing.T) {
	runDesignerE2E(t, designerRowResizeSource, designerRowResizeScript)
}

// runDesignerE2E starts the designer page against a mock backend that serves
// the given source and runs the script in a headless browser. The test fails
// when the script does not write the OK marker.
func runDesignerE2E(t *testing.T, source string, script string) {
	t.Helper()
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping designer end-to-end test")
	}
	page := injectGuiTestScript(script)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}

	ln, err := ListenLocal()
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	mux := http.NewServeMux()
	var lastHTML string
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
		json(w, `{"ok":true,"sui":`+StringifyJson(JStr(source))+`}`)
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
		_ = http.Serve(ln, http.HandlerFunc(mux.ServeHTTP))
	}()

	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=1200,800",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=120000", "--dump-dom",
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
	t.Fatalf("designer e2e failed:\n%s", detail)
}

const designerRowLayoutScript = `<script>
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
  function layoutSection() {
    var sections = document.querySelectorAll('#inspector .acc');
    for (var i = 0; i < sections.length; i++) {
      var head = sections[i].querySelector('.acchead');
      if (head && head.textContent.indexOf('Row layout') === 0) return sections[i];
    }
    return null;
  }
  // The inspector is rebuilt after every change: the fields are read again
  // before each edit.
  function layoutFields() {
    var section = layoutSection();
    if (!section) return null;
    return {
      section: section,
      numbers: section.querySelectorAll('input[type=number]'),
      units: section.querySelectorAll('select')
    };
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

    // The declared columns and the size of the cells survive the round trip
    // through the model.
    var src = ta.value || '';
    if (src.indexOf('columns: "2"') < 0) throw new Error('the column count was lost: ' + src);
    if (src.indexOf('columnWidths: "60%, 40%"') < 0) throw new Error('the column widths were lost: ' + src);
    step('round-trip');

    // Select the row and read the Layout section of the inspector.
    selPath = [0];
    multiSel = [];
    selBody = false;
    renderTree();
    renderInspector();
    await wait(200);
    var fields = layoutFields();
    if (!fields) throw new Error('the Layout section is missing');
    var numbers = fields.numbers;
    var units = fields.units;
    if (numbers.length < 3) throw new Error('the Layout section has no size fields: ' + numbers.length);
    if (numbers[0].value !== '2') throw new Error('the column count is wrong: ' + numbers[0].value);
    if (numbers[1].value !== '60') throw new Error('the width of the first column is wrong: ' + numbers[1].value);
    if (numbers[2].value !== '40') throw new Error('the width of the second column is wrong: ' + numbers[2].value);
    if (units.length < 2) throw new Error('the Layout section has no unit fields');
    if (units[0].value !== '%') throw new Error('the unit of the first column is wrong: ' + units[0].value);
    step('inspector');

    // The second column cannot be wider than the space the first one leaves
    // free: 50% is refused and the value falls back to 40%.
    numbers[2].value = '50';
    numbers[2].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    fields = layoutFields();
    if (!fields || fields.numbers[2].value !== '40') {
      throw new Error('the width was not clamped: ' + (fields ? fields.numbers[2].value : 'no section'));
    }
    var after = document.getElementById('out').value || '';
    if (after.indexOf('columnWidths: "60%, 40%"') < 0) throw new Error('the clamped width was not written: ' + after);
    step('clamped');

    // A column without a size takes the space the other columns leave free.
    fields = layoutFields();
    fields.numbers[1].value = '';
    fields.units[0].value = 'auto';
    fields.units[0].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    var auto = document.getElementById('out').value || '';
    if (auto.indexOf('columnWidths: "auto, 40%"') < 0) throw new Error('the automatic column was not written: ' + auto);
    step('automatic');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`

const designerRowResizeScript = `<script>
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
    }, 20000);
    if (!handle) throw new Error('the cell split handle is missing');
    var doc = frameDoc();
    var hr = handle.getBoundingClientRect();
    handle.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, button: 0, clientX: hr.left + 6, clientY: hr.top + 10 }));
    doc.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: hr.left + 40, clientY: hr.top + 10 }));
    doc.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: hr.left + 40, clientY: hr.top + 10 }));
    step('dragged');

    // The container declares the size of its cells: the dragged cell gets a
    // fixed size and the cell on its right takes the space left free.
    var written = await until(function () {
      var src = document.getElementById('out').value || '';
      return src.indexOf('columnWidths:') >= 0 ? src : null;
    }, 20000);
    if (!written) throw new Error('the cell sizes were not written back');
    var match = /columnWidths: "([^"]*)"/.exec(written);
    if (!match) throw new Error('the column widths are missing: ' + written);
    var sizes = match[1].split(',').map(function (s) { return s.trim(); });
    if (sizes.length !== 3) throw new Error('the size list does not describe every cell: ' + match[1]);
    if (!/^\d+px$/.test(sizes[0])) throw new Error('the dragged cell has no fixed size: ' + match[1]);
    if (sizes[1] !== 'auto') throw new Error('the cell on the right is not automatic: ' + match[1]);
    step('written');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`
