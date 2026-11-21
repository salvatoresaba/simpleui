package sui

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
	const source = `css ":root \{ --brand: #ff0000; \} .one \{ display: block; width: 10px; \} .two \{ display: flex; width: 20px; \}"
view main {
	card "One" class: "one"
	card "Two" class: "two"
}`
	page := injectGuiTestScript(guiCssAdvancedScript)
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
	mux.HandleFunc("POST /__css__/setselmany", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiCssSetSelectors(r.PostFormValue("css"), r.PostFormValue("selectors"), r.PostFormValue("at"), r.PostFormValue("decls"))))
	})
	go func() { _ = http.Serve(ln, mux) }()
	baseURL := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
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
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
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
    function mixedRow(prop) {
      var rows = body.querySelectorAll('.adv-row');
      for (var ri = 0; ri < rows.length; ri++) {
        var pin = rows[ri].querySelector('.adv-prop');
        var vin = rows[ri].querySelector('.adv-val');
        if (pin && (pin.value || '') === prop && vin && vin.placeholder === '(mixed)') return vin;
      }
      return null;
    }
    var displayMixed = mixedRow('display');
    var widthMixed = mixedRow('width');
    if (!displayMixed || !widthMixed) throw new Error('mixed values were not shown');
    displayMixed.value = 'grid';
    displayMixed.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(700);
    if ((root.css.match(/display: grid/g) || []).length < 2) throw new Error('mixed display was not applied to both classes');
    body = document.getElementById('ins-overlay-body');
    widthMixed = mixedRow('width');
    if (!widthMixed) throw new Error('mixed width disappeared early');
    widthMixed.value = '30px';
    widthMixed.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(700);
    if ((root.css.match(/width: 30px/g) || []).length < 2) throw new Error('mixed width was not applied to both classes');
    addDecl('background-color', '#102030');
    await wait(700);
    if (root.css.indexOf('.one') < 0 || root.css.indexOf('.two') < 0 || root.css.indexOf('#102030') < 0) throw new Error('multi-class declaration not applied');
    var addRow = document.querySelector('#ins-overlay-body .adv-add');
    var propInput = addRow.querySelector('.adv-prop');
    var valInput = addRow.querySelector('.adv-val');
    propInput.value = 'display'; propInput.dispatchEvent(new Event('input', { bubbles: true }));
    valInput.focus();
    await wait(100);
    for (var ti = 0; ti < 40 && (!tokens || !tokens.length); ti++) await wait(50);
    var valueList = document.getElementById('css-val-list');
    var hasFlex = false, hasBrand = false;
    if (valueList) for (var oi = 0; oi < valueList.options.length; oi++) {
      if (valueList.options[oi].value === 'flex') hasFlex = true;
      if (valueList.options[oi].value === 'var(--brand)') hasBrand = true;
    }
    if (!hasFlex) throw new Error('CSS value autocomplete missing display suggestions');
    if (!hasBrand) throw new Error('CSS value autocomplete missing project token');
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
    body = document.getElementById('ins-overlay-body');
    var visualBtn = null, vbs = body.querySelectorAll('button');
    for (var vi = 0; vi < vbs.length; vi++) if ((vbs[vi].textContent || '').trim() === 'Visual mode') visualBtn = vbs[vi];
    if (!visualBtn) throw new Error('Visual mode button missing');
    visualBtn.click();
    await wait(400);
    body = document.getElementById('ins-overlay-body');
    var beforeInvalid = root.css;
    var invalidAt = null, invalidInputs = body.querySelectorAll('input');
    for (var ii = 0; ii < invalidInputs.length; ii++) if ((invalidInputs[ii].placeholder || '').indexOf('@media') >= 0) invalidAt = invalidInputs[ii];
    if (!invalidAt) throw new Error('at-rule input missing for validation test');
    invalidAt.value = '@media {';
    invalidAt.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    body = document.getElementById('ins-overlay-body');
    addDecl('color', 'blue');
    await wait(300);
    if (root.css !== beforeInvalid) throw new Error('invalid at-rule should reject the whole multi-class write');
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
    if ((root.css.match(/linear-gradient\(/g) || []).length < 2) throw new Error('gradient builder was not applied atomically to both classes');
    body = document.getElementById('ins-overlay-body');
    var gradientType = body.querySelector('.adv-builder select');
    gradientType.value = 'radial';
    body.querySelector('.adv-builder .adv-btns button').click();
    await wait(800);
    if (root.css.indexOf('radial-gradient(') < 0) throw new Error('radial gradient builder did not apply');
    body = document.getElementById('ins-overlay-body');
    gradientType = body.querySelector('.adv-builder select');
    gradientType.value = 'conic';
    body.querySelector('.adv-builder .adv-btns button').click();
    await wait(800);
    if (root.css.indexOf('conic-gradient(') < 0) throw new Error('conic gradient builder did not apply');
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
    if ((root.css.match(/scale\(1\.1\)/g) || []).length < 2) throw new Error('transform builder was not applied atomically to both classes');
    body = document.getElementById('ins-overlay-body');
    var builders = body.querySelectorAll('.adv-builder');
    if (builders.length < 2) throw new Error('clip-path builder missing');
    var clipCard = builders[1];
    var clipShape = clipCard.querySelector('select');
    clipShape.value = 'polygon';
    clipCard.querySelector('.adv-btns button').click();
    await wait(800);
    if (root.css.indexOf('clip-path: polygon(') < 0) throw new Error('clip-path builder did not apply');
    closeOverlay();
    await wait(250);
    body = openBuilder('Advanced flex...', 'Flex builder');
    var flexDirection = body.querySelector('.adv-builder select');
    flexDirection.value = 'column';
    flexDirection.dispatchEvent(new Event('change', { bubbles: true }));
    body.querySelector('.adv-builder .adv-btns button').click();
    await wait(800);
    if (root.css.indexOf('flex-direction: column') < 0) throw new Error('flex builder did not apply');
    closeOverlay();
    await wait(250);
    body = openBuilder('Advanced grid...', 'Grid builder');
    var gridLabels = body.querySelectorAll('label');
    for (var gi = 0; gi < gridLabels.length; gi++) {
      if ((gridLabels[gi].textContent || '').trim() === 'columns') { gridLabels[gi].nextElementSibling.value = '1fr 2fr'; break; }
    }
    body.querySelector('.adv-builder .adv-btns button').click();
    await wait(800);
    if (root.css.indexOf('grid-template-columns: 1fr 2fr') < 0) throw new Error('grid builder did not apply');
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
    body = document.getElementById('ins-overlay-body');
    var sideSelect = body.querySelector('.adv-builder select');
    sideSelect.value = 'top';
    sideSelect.dispatchEvent(new Event('change', { bubbles: true }));
    var sideSelects = body.querySelectorAll('.adv-builder select');
    if (sideSelects.length > 1) {
      sideSelects[1].value = 'dashed';
      sideSelects[1].dispatchEvent(new Event('change', { bubbles: true }));
    }
    body.querySelector('.adv-builder .adv-btns button').click();
    await wait(800);
    if (root.css.indexOf('border-top-style: dashed') < 0) throw new Error('per-side top border did not apply');
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
