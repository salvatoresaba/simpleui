package core

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// designer_interaction_test.go covers the designer interactions an author
// uses while editing a page:
//
//   - a style edit (layout, alignment) must not move the selection to another
//     element: the preview used to replay the tab restore as a real click and
//     that click picked the sheet table instead of the edited element;
//   - a drag & drop in the tree must keep the moved element selected and
//     outlined in the preview;
//   - every edit must keep the std import: the generated source IS the app,
//     so a lost import breaks the build and empties the preview;
//   - Ctrl+F must open the find bar from any panel, not only from the editor;
//   - select mode must pick an element in the preview without running the
//     action bound to it.

// TestDesignerInteractionE2E drives the real designer in a headless browser.
func TestDesignerInteractionE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping designer interaction test")
	}
	const source = `import { json } from "std"

state counter = 0

fn increment() -> null {
	counter = counter + 1
}

view main {
	card "Panel" class: "panel" {
		button "Increment" onClick: increment class: "bump"
		p "Count: {counter}" class: "counter"
	}
	tabs "Sections" {
		tab "First" {
			p "First page" class: "first"
		}
		tab "Second" {
			p "Second page" class: "second"
		}
	}
}`
	page := injectGuiTestScript(designerInteractionScript)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}

	ln, err := ListenLocal()
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
	mux.HandleFunc("GET /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{}`)
	})
	mux.HandleFunc("POST /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	mux.HandleFunc("POST /__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ok, html, msgs := renderSui(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		lastHTML = html
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	mux.HandleFunc("POST /__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		model, msgs, details, ok := guiOpenDiags(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`,"diagDetails":`+StringifyJson(details)+`}`)
			return
		}
		body := StringifyJson(model)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,`+body[1:])
	})
	mux.HandleFunc("POST /__gui__/apply", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiApplySource(r.PostFormValue("sui"))))
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", lastHTML)
	})
	mux.HandleFunc("POST /__css__/parse", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssParse(r.PostFormValue("css"), r.PostFormValue("used"))))
	})
	mux.HandleFunc("POST /__css__/setbatch", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		res := guiCssSetBatch(r.PostFormValue("css"), r.PostFormValue("class"), r.PostFormValue("decls"))
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	mux.HandleFunc("POST /__css__/setselmany", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		res := guiCssSetSelectors(r.PostFormValue("css"), r.PostFormValue("selectors"), r.PostFormValue("at"), r.PostFormValue("decls"))
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	mux.HandleFunc("GET /__gui__/fonts", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"fonts":[]}`)
	})
	mux.HandleFunc("GET /__gui__/templates", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"templates":[]}`)
	})
	mux.HandleFunc("GET /__gui__/assets", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiAssetList()))
	})
	mux.HandleFunc("GET /__gui__/history", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"versions":[]}`)
	})

	go func() { _ = http.Serve(ln, mux) }()
	baseURL := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=1200,900",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=60000", "--dump-dom", baseURL,
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
	if marker := strings.Index(text, `id="e2e-result"`); marker >= 0 {
		end := marker + 1600
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
	t.Fatalf("designer interaction e2e failed:\n%s", detail)
}

const designerInteractionScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    el.setAttribute('data-e2e', 'done');
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function step(name) {
    var el = document.getElementById('e2e-step');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-step'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = name;
  }
  var wait = function (ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); };
  function rows() { return Array.prototype.slice.call(document.querySelectorAll('#nodes .node')); }
  function rowByLabel(text) {
    return rows().find(function (r) {
      var l = r.querySelector('.node-label');
      return l && l.textContent.indexOf(text) >= 0;
    });
  }
  function previewDoc() {
    var frame = document.getElementById('prev');
    return frame ? frame.contentDocument : null;
  }
  try {
    step('start');
    var ready = false;
    for (var i = 0; i < 150; i++) {
      if (rows().length >= 3 && document.getElementById('out')) { ready = true; break; }
      await wait(100);
    }
    if (!ready) throw new Error('the designer did not load the tree');
    var ta = document.getElementById('out');
    if (ta.value.indexOf('import') < 0) throw new Error('the source has no import: ' + ta.value.slice(0, 120));
    step('loaded');

    // 1. a layout edit keeps the selection on the edited element.
    var card = rowByLabel('.panel');
    if (!card) throw new Error('the panel row is missing');
    card.click();
    await wait(600);
    var pathBefore = selPath ? selPath.join('.') : '';
    if (!pathBefore) throw new Error('the panel row did not select anything');
    var accs = Array.prototype.slice.call(document.querySelectorAll('#inspector .acc'));
    var layout = accs.find(function (a) {
      var h = a.querySelector('.acchead');
      return h && h.textContent.indexOf('Layout') >= 0;
    });
    if (!layout) throw new Error('the inspector has no Layout section');
    var disp = layout.querySelector('select');
    disp.value = 'grid';
    disp.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(2500);
    var pathAfter = selPath ? selPath.join('.') : '';
    if (pathAfter !== pathBefore) throw new Error('the selection moved after the layout edit: ' + pathBefore + ' -> ' + pathAfter);
    if (root.css.indexOf('display: grid') < 0) throw new Error('the layout edit did not reach the CSS');
    if (ta.value.indexOf('import') < 0) throw new Error('the layout edit dropped the std import');
    var doc = previewDoc();
    if (!doc || !doc.body || !doc.body.children.length) throw new Error('the preview is empty after the layout edit');
    step('layout-ok');

    // 2. a drag & drop keeps the moved element selected and outlined.
    var paraRow = rowByLabel('.counter');
    var buttonRow = rowByLabel('.bump');
    if (!paraRow || !buttonRow) throw new Error('the rows to move are missing');
    var dt = new DataTransfer();
    paraRow.dispatchEvent(new DragEvent('dragstart', { dataTransfer: dt, bubbles: true, cancelable: true }));
    buttonRow.dispatchEvent(new DragEvent('dragover', { dataTransfer: dt, bubbles: true, cancelable: true }));
    buttonRow.dispatchEvent(new DragEvent('drop', { dataTransfer: dt, bubbles: true, cancelable: true }));
    await wait(2500);
    var movedText = ta.value;
    var paraAt = movedText.indexOf('Count:');
    var buttonAt = movedText.indexOf('"Increment"');
    if (paraAt < 0 || buttonAt < 0 || paraAt > buttonAt) throw new Error('the drag & drop did not reorder the source');
    if (movedText.indexOf('import') < 0) throw new Error('the drag & drop dropped the std import');
    var movedPath = selPath ? selPath.join('.') : '';
    if (!movedPath) throw new Error('the drag & drop lost the selection');
    var movedNode = byPath(selPath, root.children);
    if (!movedNode || String(movedNode.class || '').indexOf('counter') < 0) {
      throw new Error('the selection is not the moved element: ' + movedPath);
    }
    doc = previewDoc();
    var hl = doc ? doc.querySelectorAll('.sui-hl') : [];
    if (!hl.length) throw new Error('the moved element is not outlined in the preview');
    var hlClass = hl[0].getAttribute('class') || '';
    if (hlClass.indexOf('counter') < 0) throw new Error('the preview outlines the wrong element: ' + hlClass);
    step('drag-ok');

    // 3. Ctrl+F opens the find bar from any panel.
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'f', ctrlKey: true, bubbles: true, cancelable: true }));
    await wait(500);
    if (!document.getElementById('cpane-sui').classList.contains('on')) throw new Error('Ctrl+F did not show the SUI pane');
    if (!document.getElementById('sui-findbar').classList.contains('on')) throw new Error('Ctrl+F did not open the find bar');
    var findInput = document.getElementById('sui-find-input');
    findInput.value = 'counter';
    findInput.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(500);
    var findCount = document.getElementById('sui-find-count').textContent || '';
    if (!/[1-9]/.test(findCount)) throw new Error('the find bar found no match: ' + findCount);
    step('find-ok');

    // 4. select mode picks without running the action.
    var selectBtn = document.getElementById('select-mode-btn');
    if (!selectBtn) throw new Error('the preview has no select mode button');
    selectBtn.click();
    await wait(500);
    doc = previewDoc();
    if (!doc) throw new Error('the preview is missing');
    var appButton = doc.querySelector('button');
    if (!appButton) throw new Error('the preview has no button');
    appButton.click();
    await wait(900);
    doc = previewDoc();
    var text = (doc && doc.body ? doc.body.textContent : '') || '';
    if (text.indexOf('Count: 1') >= 0) throw new Error('select mode still ran the action of the element');
    if (!document.querySelectorAll('#nodes .node.sel').length) throw new Error('the picked element is not selected in the tree');
    // back to the normal mode: the action runs again.
    selectBtn.click();
    await wait(400);
    doc = previewDoc();
    appButton = doc.querySelector('button');
    appButton.click();
    await wait(900);
    doc = previewDoc();
    text = (doc && doc.body ? doc.body.textContent : '') || '';
    if (text.indexOf('Count: 1') < 0) throw new Error('the action did not run outside select mode');
    step('select-ok');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`
