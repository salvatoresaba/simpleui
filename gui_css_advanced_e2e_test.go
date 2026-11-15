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

// TestGuiCssAdvancedE2E drives the advanced CSS overlay with a multi-selection:
// mixed values, text mode, pseudo-element context, at-rule context and value
// autocomplete. It also checks that the visual builder buttons are present.
func TestGuiCssAdvancedE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping advanced CSS designer test")
	}
	const source = `view main {
	card "One" class: "one"
	card "Two" class: "two"
}`
	page := strings.Replace(guiHTML, "</body>", guiCssAdvancedScript+"</body>", 1)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	var lastHTML string
	mux := http.NewServeMux()
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
	mux.HandleFunc("POST /__css__/parse", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssParse(r.PostFormValue("css"), r.PostFormValue("used"))))
	})
	mux.HandleFunc("POST /__css__/setbatch", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssSetBatch(r.PostFormValue("css"), r.PostFormValue("class"), r.PostFormValue("decls"))))
	})
	mux.HandleFunc("POST /__css__/setsel", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssSetSelector(r.PostFormValue("css"), r.PostFormValue("selector"), r.PostFormValue("at"), r.PostFormValue("decls"))))
	})
	go func() { _ = http.Serve(ln, mux) }()
	baseURL := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=1000,800",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=20000", "--dump-dom", baseURL,
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
	t.Fatalf("advanced CSS designer e2e failed:\n%s", detail)
}

const guiCssAdvancedScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); }
  function nodeByText(text) {
    var list = document.querySelectorAll('#nodes .node');
    for (var i = 0; i < list.length; i++) if ((list[i].textContent || '').indexOf(text) >= 0) return list[i];
    return null;
  }
  var errs = [];
  window.addEventListener('error', function (e) { errs.push((e && e.message) || 'error'); });
  function buttonInInspector(text) {
    var list = document.querySelectorAll('#inspector button');
    for (var i = 0; i < list.length; i++) if ((list[i].textContent || '').trim() === text) return list[i];
    return null;
  }
  function inspectorRow(label) {
    var rows = document.querySelectorAll('#inspector .rowline');
    for (var i = 0; i < rows.length; i++) if ((rows[i].textContent || '').indexOf(label) >= 0) return rows[i];
    return null;
  }
  function addDecl(prop, value) {
    var add = document.querySelector('#ins-overlay-body .adv-add');
    if (!add) throw new Error('advanced CSS add row missing');
    var pin = add.querySelector('.adv-prop');
    var vin = add.querySelector('.adv-val');
    pin.value = prop; vin.value = value;
    add.querySelector('button').click();
  }
  try {
    for (var i = 0; i < 120 && !nodeByText('Two'); i++) await wait(100);
    var one = nodeByText('One'), two = nodeByText('Two');
    if (!one || !two) throw new Error('tree nodes missing');
    one.click();
    await wait(250);
    two.dispatchEvent(new MouseEvent('click', { bubbles: true, ctrlKey: true }));
    await wait(350);
    var advRow = inspectorRow('CSS declarations');
    var edit = advRow ? advRow.querySelector('button') : null;
    if (!edit) throw new Error('CSS declarations Edit button missing');
    edit.click();
    await wait(400);
    var body = document.getElementById('ins-overlay-body');
    if (!body) throw new Error('CSS overlay missing: ' + errs.join(' | ') + ' disabled=' + edit.disabled + ' multi=' + multiSel.length + ' sel=' + (selPath ? selPath.join(',') : 'none') + ' edit=' + (edit.textContent || '') + ' row=' + (advRow.textContent || '').slice(0, 120));
    if (body.textContent.indexOf('Editing 2 selected elements') < 0) throw new Error('multi-selection hint missing');
    addDecl('background-color', '#102030');
    await wait(700);
    if (root.css.indexOf('.one') < 0 || root.css.indexOf('.two') < 0 || root.css.indexOf('#102030') < 0) throw new Error('multi-class declaration not applied');
    var addRow = document.querySelector('#ins-overlay-body .adv-add');
    var propInput = addRow.querySelector('.adv-prop');
    var valInput = addRow.querySelector('.adv-val');
    propInput.value = 'display'; propInput.dispatchEvent(new Event('input', { bubbles: true }));
    valInput.focus();
    await wait(100);
    var valueList = document.getElementById('css-val-list');
    var hasFlex = false;
    if (valueList) for (var oi = 0; oi < valueList.options.length; oi++) if (valueList.options[oi].value === 'flex') hasFlex = true;
    if (!hasFlex) throw new Error('CSS value autocomplete missing display suggestions');
    propInput.value = ''; valInput.value = '';
    var pseudo = body.querySelector('select');
    if (!pseudo) throw new Error('pseudo context select missing');
    pseudo.value = '::before'; pseudo.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    body = document.getElementById('ins-overlay-body');
    var content = null, labels = body.querySelectorAll('label');
    for (var c = 0; c < labels.length; c++) if ((labels[c].textContent || '').trim() === 'content') content = labels[c].nextElementSibling;
    if (!content) throw new Error('pseudo content field missing');
    content.value = '"x"';
    var setContent = null, bs = body.querySelectorAll('button');
    for (var b = 0; b < bs.length; b++) if ((bs[b].textContent || '').trim() === 'Set content') setContent = bs[b];
    if (!setContent) throw new Error('Set content button missing');
    setContent.click();
    await wait(700);
    if (root.css.indexOf('.one::before') < 0 || root.css.indexOf('.two::before') < 0) throw new Error('pseudo-element rules missing');
    var at = null, inputs = body.querySelectorAll('input');
    for (var a = 0; a < inputs.length; a++) if ((inputs[a].placeholder || '').indexOf('@media') >= 0) at = inputs[a];
    if (!at) throw new Error('at-rule input missing');
    at.value = '@media (max-width: 767px)'; at.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    addDecl('color', 'red');
    await wait(700);
    if (root.css.indexOf('@media (max-width: 767px)') < 0 || root.css.indexOf('color: red') < 0) throw new Error('at-rule declaration missing');
    body = document.getElementById('ins-overlay-body');
    var textBtn = null, buttons = body.querySelectorAll('button');
    for (var t = 0; t < buttons.length; t++) if ((buttons[t].textContent || '').trim() === 'Text mode') textBtn = buttons[t];
    if (!textBtn) throw new Error('Text mode button missing');
    textBtn.click();
    await wait(400);
    var ta = document.getElementById('adv-text-mode');
    if (!ta) throw new Error('Advanced CSS text mode missing');
    ta.value = 'background-color: #abcdef;';
    var applyText = null, tbs = document.querySelectorAll('#ins-overlay-body button');
    for (var x = 0; x < tbs.length; x++) if ((tbs[x].textContent || '').trim() === 'Apply text') applyText = tbs[x];
    if (!applyText) throw new Error('Apply text button missing');
    applyText.click();
    await wait(800);
    if (root.css.indexOf('#abcdef') < 0) throw new Error('text mode declaration not applied');
    function closeOverlay() { var c = document.querySelector('#ins-overlay-h button'); if (c) c.click(); }
    function openBuilder(label, titlePart) {
      var b = buttonInInspector(label);
      if (!b) throw new Error(label + ' button missing');
      b.click();
      var title = document.querySelector('#ins-overlay-body .adv-builder-title');
      if (!title || (title.textContent || '').indexOf(titlePart) < 0) throw new Error(titlePart + ' builder missing');
      return document.getElementById('ins-overlay-body');
    }
    closeOverlay();
    await wait(250);
    body = openBuilder('Advanced background...', 'Gradient builder');
    await wait(300);
    var gradientApply = body.querySelector('.adv-builder .adv-btns button');
    if (!gradientApply) throw new Error('gradient Apply missing');
    gradientApply.click();
    await wait(800);
    if (root.css.indexOf('linear-gradient(') < 0) throw new Error('gradient builder did not apply');
    closeOverlay();
    await wait(250);
    body = openBuilder('Transform and clip path...', 'Transform builder');
    var labels = body.querySelectorAll('label');
    for (var li = 0; li < labels.length; li++) {
      if ((labels[li].textContent || '').trim() === 'scale') { labels[li].nextElementSibling.value = '1.1'; break; }
    }
    var transformApply = body.querySelector('.adv-builder .adv-btns button');
    if (!transformApply) throw new Error('transform Apply missing');
    transformApply.click();
    await wait(800);
    if (root.css.indexOf('scale(1.1)') < 0) throw new Error('transform builder did not apply');
    closeOverlay();
    await wait(250);
    body = openBuilder('Per-side borders...', 'Per-side border builder');
    var borderLabels = body.querySelectorAll('label');
    for (var bi = 0; bi < borderLabels.length; bi++) {
      if ((borderLabels[bi].textContent || '').trim() === 'style') { borderLabels[bi].nextElementSibling.value = 'solid'; }
      if ((borderLabels[bi].textContent || '').trim() === 'width') { borderLabels[bi].nextElementSibling.value = '2px'; }
    }
    var borderApply = body.querySelector('.adv-builder .adv-btns button');
    if (!borderApply) throw new Error('border Apply missing');
    borderApply.click();
    await wait(800);
    if (root.css.indexOf('border-style: solid') < 0 && root.css.indexOf('border-top-style') < 0) throw new Error('per-side border builder did not apply');
    closeOverlay();
    await wait(250);
    body = openBuilder('Border image...', 'Border image builder');
    var borderInputs = body.querySelectorAll('input');
    for (var ci = 0; ci < borderInputs.length; ci++) {
      if ((borderInputs[ci].placeholder || '').indexOf('url(') >= 0) borderInputs[ci].value = 'url(/x.png)';
    }
    var imageApply = body.querySelector('.adv-builder .adv-btns button');
    if (!imageApply) throw new Error('border image Apply missing');
    imageApply.click();
    await wait(800);
    if (root.css.indexOf('border-image-source: url(/x.png)') < 0) throw new Error('border image builder did not apply');
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
