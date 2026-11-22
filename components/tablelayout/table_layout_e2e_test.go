package tablelayout_test

import . "github.com/salvatoresaba/simpleui"

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const tableLayoutE2ESource = `view main {
	title "Layout e2e"
	tablelayout rows: 1 columns: 2 width: "120px" rowHeights: "200px, 200px" class: "lt" {
		layoutcell row: 0 col: 0 {
			p "Hello"
		}
		layoutcell row: 0 col: 1 {
			card "B"
		}
		layoutcell row: 1 col: 0 {
			card "C"
		}
		layoutcell row: 1 col: 1 {
			card "D"
		}
	}
}`

// TestGuiTableLayoutVisualE2E drives the real designer in a headless browser:
// cell inspector filtering, alignment inside a cell, span merge/reset, split
// drag and the grid reduction dialog.
func TestGuiTableLayoutVisualE2E(t *testing.T) {
	browser := FindHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping table layout visual test")
	}

	page := InjectGuiTestScript(tableLayoutE2EScript)
	if page == GuiHTML {
		t.Fatal("designer page has no closing body")
	}

	var lastHTML string
	mux := GUITestMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		WriteRawHandler(w, "text/html; charset=utf-8", page)
	})
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		WriteRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(tableLayoutE2ESource))+`}`)
	})
	mux.HandleFunc("POST /__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ok, html, msgs := RenderSui(r.PostFormValue("sui"))
		if ok {
			lastHTML = html
			WriteRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
			return
		}
		WriteRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
	})
	mux.HandleFunc("POST /__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		model, msgs, ok := GuiOpen(r.PostFormValue("sui"))
		if !ok {
			WriteRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		body := StringifyJson(model)
		WriteRawHandler(w, "application/json; charset=utf-8", `{"ok":true,`+body[1:])
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		WriteRawHandler(w, "text/html; charset=utf-8", lastHTML)
	})
	mux.HandleFunc("POST /__css__/setbatch", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		res := GuiCssSetBatch(r.PostFormValue("css"), r.PostFormValue("class"), r.PostFormValue("decls"))
		WriteRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	mux.HandleFunc("POST /__css__/setselmany", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		res := GuiCssSetSelectors(r.PostFormValue("css"), r.PostFormValue("selectors"), r.PostFormValue("at"), r.PostFormValue("decls"))
		WriteRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	server := NewLocalServer(mux)
	defer server.Close()

	url := server.URL + "/"
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--no-first-run",
		"--no-default-browser-check",
		"--user-data-dir=" + t.TempDir(),
		"--virtual-time-budget=45000",
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
		end := marker + 1600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("table layout visual e2e failed:\n%s", detail)
}

const tableLayoutE2EScript = `<script>
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
  function until(fn, timeoutMs) {
    return (async function () {
      var deadline = Date.now() + (timeoutMs || 5000);
      for (;;) {
        var value = fn();
        if (value) return value;
        if (Date.now() >= deadline) return null;
        await wait(20);
      }
    })();
  }
  function frameDoc() {
    var f = document.getElementById('prev');
    return f && f.contentDocument ? f.contentDocument : null;
  }
  async function waitPreview() {
    for (var i = 0; i < 200; i++) {
      var doc = frameDoc();
      if (doc && doc.querySelector('table.sui-layout-table') && doc.querySelector('.sui-layout-pick-cell')) return doc;
      await wait(100);
    }
    return null;
  }
  function acc(title) {
    return Array.from(document.querySelectorAll('#inspector .acc')).find(function (a) {
      var h = a.querySelector('.acchead span');
      return h && h.textContent.trim() === title;
    });
  }
  function secLabels(title) {
    var a = acc(title);
    return a ? Array.from(a.querySelectorAll('.rowline label')).map(function (l) { return l.textContent; }) : [];
  }
  function heads() {
    return Array.from(document.querySelectorAll('#inspector .acchead span:first-child')).map(function (h) { return h.textContent.trim(); });
  }
  function selectionMatches(row, col, add) {
    var cells = window.selLayoutCells;
    if (!cells) return false;
    return add ? cells.length >= 2 : cells.length === 1 && cells[0].row === row && cells[0].col === col;
  }
  async function selectCell(row, col, add) {
    for (var attempt = 0; attempt < 5; attempt++) {
      // Re-acquire the pick immediately before clicking: under load the
      // preview can be re-rendered between two asynchronous waits, leaving a
      // stale node whose click never reaches the designer.
      var pick = await until(function () {
        var d = frameDoc();
        if (!d) return null;
        return Array.from(d.querySelectorAll('.sui-layout-pick-cell')).find(function (p) {
          var c = p.parentElement;
          return c && c.getAttribute('data-sui-layout-row') === String(row) && c.getAttribute('data-sui-layout-col') === String(col);
        }) || null;
      }, 8000);
      if (!pick) throw new Error('cell pick ' + row + ',' + col + ' not found');
      if (add) pick.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, ctrlKey: true }));
      else pick.click();
      var selected = await until(function () { return selectionMatches(row, col, add) ? true : null; }, 8000);
      if (!selected && selectionMatches(row, col, add)) selected = true;
      if (!selected) {
        await wait(100);
        continue;
      }
      if (await until(function () { return acc('Layout cell') ? true : null; }, 8000)) return;
      throw new Error('cell selection updated but the inspector did not render');
    }
    throw new Error('cell selection did not update (cells ' + JSON.stringify(window.selLayoutCells) + ')');
  }
  async function selectTable() {
    for (var attempt = 0; attempt < 5; attempt++) {
      var pick = await until(function () {
        var d = frameDoc();
        return d && d.querySelector('.sui-layout-pick-table') || null;
      }, 8000);
      if (!pick) throw new Error('table pick not found');
      pick.click();
      if (await until(function () {
        return window.selLayoutCells && window.selLayoutCells.length === 0 ? true : null;
      }, 8000)) return;
      await wait(100);
    }
    throw new Error('table selection did not clear the cell selection');
  }
  function cell(row, col) {
    return frameDoc().querySelector('.sui-layout-cell[data-sui-layout-row="' + row + '"][data-sui-layout-col="' + col + '"]');
  }
  try {
    for (var i = 0; i < 200 && !document.querySelector('#nodes .node'); i++) await wait(100);
    await wait(400);
    if (!(await waitPreview())) throw new Error('preview did not render');

    // Opening Structure syncs the declared counters with the real grid.
    await selectTable();
    var structBtn = await until(function () {
      var section = acc('Table layout');
      if (!section) return null;
      return Array.from(section.querySelectorAll('button')).find(function (b) {
        return b.textContent.trim() === 'Structure table settings';
      }) || null;
    }, 8000);
    if (!structBtn) throw new Error('structure button missing');
    structBtn.click();
    await wait(700);
    var structRows = Array.from(acc('Table layout').querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'rows';
    });
    if (!structRows || structRows.querySelector('input').value !== '2') throw new Error('structure open did not sync rows counter');
    var structOverlay = document.getElementById('layoutstruct');
    if (!structOverlay.classList.contains('settings-floating')) throw new Error('structure panel must float');
    if (window.getComputedStyle(structOverlay).pointerEvents !== 'none') throw new Error('floating overlay must not block the editor');
    var structBox = document.getElementById('layoutstructbox');
    if (!structBox.querySelector('.settings-grip')) throw new Error('structure grip missing');
    var structHead = structBox.querySelector('.settings-drag-head');
    var structLeft = structBox.getBoundingClientRect().left;
    structHead.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, cancelable: true, pointerId: 1, pointerType: 'mouse', button: 0, clientX: 100, clientY: 100 }));
    structHead.dispatchEvent(new PointerEvent('pointermove', { bubbles: true, cancelable: true, pointerId: 1, pointerType: 'mouse', clientX: 130, clientY: 130 }));
    structHead.dispatchEvent(new PointerEvent('pointerup', { bubbles: true, cancelable: true, pointerId: 1, pointerType: 'mouse', clientX: 130, clientY: 130 }));
    await wait(150);
    if (structBox.getBoundingClientRect().left < structLeft + 10) throw new Error('structure panel did not move');
    var structClose = Array.from(document.querySelectorAll('#layoutstruct button')).find(function (b) { return b.textContent.trim() === 'Close'; });
    if (structClose) structClose.click();
    await wait(300);
    structBtn.click();
    await wait(500);
    if (structBox.style.left !== '' || structBox.style.top !== '') throw new Error('structure panel must restart centered');
    structClose = Array.from(document.querySelectorAll('#layoutstruct button')).find(function (b) { return b.textContent.trim() === 'Close'; });
    if (structClose) structClose.click();
    await wait(300);

    await selectCell(0, 0);
    var h = heads();
    if (h.indexOf('Position') >= 0) throw new Error('Position must be hidden for a layout cell');
    if (h.indexOf('Size') >= 0) throw new Error('Size must be hidden for a layout cell');
    if (h.indexOf('Responsive') >= 0) throw new Error('Responsive must be hidden for a layout cell');
    if (h.indexOf('Alignment') >= 0) throw new Error('Alignment must be hidden for a layout cell');
    if (h.indexOf('Layout') >= 0) throw new Error('Layout must be hidden for a layout cell');
    if (h.indexOf('Grid') >= 0) throw new Error('Grid must be hidden for a layout cell');
    if (h.indexOf('Bindings') >= 0) throw new Error('Bindings must be hidden for a layout cell');
    if (h.indexOf('Tooltip') < 0) throw new Error('Tooltip must stay available for a layout cell');
    if (h.indexOf('Events') < 0) throw new Error('Events must stay available for a layout cell');
    if (h.indexOf('Drag & drop') < 0) throw new Error('Drag & drop must stay available for a layout cell');
    var box = secLabels('Box model');
    if (box.indexOf('margin') >= 0) throw new Error('margin must be hidden for a layout cell');
    if (box.indexOf('padding') < 0) throw new Error('padding must stay available for a layout cell');
    if (heads().indexOf('Table layout') < 0 || heads().indexOf('Layout cell') < 0) throw new Error('table and cell sections must stay visible');
    var preset = secLabels('Table layout');
    if (preset.indexOf('preset') >= 0) throw new Error('preset must be hidden when the table has content');

    // Select the text inside the cell and center it in the 200px row.
    var doc = frameDoc();
    var p = doc.querySelector('table.sui-layout-table p');
    var pb = p.getBoundingClientRect();
    var target = doc.elementFromPoint(pb.left + 12, pb.top + pb.height / 2);
    var o = { bubbles: true, cancelable: true, button: 0, clientX: pb.left + 12, clientY: pb.top + pb.height / 2 };
    target.dispatchEvent(new MouseEvent('mousedown', o));
    target.dispatchEvent(new MouseEvent('mouseup', o));
    target.dispatchEvent(new MouseEvent('click', o));
    await wait(450);
    if (!(await until(function () { return !!acc('Alignment'); }, 5000))) {
      throw new Error('Alignment must be available for content inside a cell');
    }
    var align = acc('Alignment');
    var center = Array.from(align.querySelectorAll('button')).find(function (b) { return b.title === 'Center'; });
    if (!center) throw new Error('Center alignment button not found');
    center.click();
    if (!(await until(function () {
      var d = frameDoc();
      var para = d && d.querySelector('table.sui-layout-table p');
      return para && d.defaultView.getComputedStyle(para).textAlign === 'center' ? true : null;
    }, 8000))) throw new Error('text-align was not applied');
    doc = frameDoc();
    p = doc.querySelector('table.sui-layout-table p');
    var pc = p.parentElement;
    var pr = p.getBoundingClientRect(), cr = pc.getBoundingClientRect();
    if (doc.defaultView.getComputedStyle(pc).display !== 'table-cell') throw new Error('layout cell must stay a table-cell after alignment');
    if (Math.abs((pr.top - cr.top) - (cr.bottom - pr.bottom)) > 24) throw new Error('text is not vertically centered');

    // Span merge and reset on the first cell.
    await selectCell(0, 0);
    if (!(await until(function () { return secLabels('Layout cell').indexOf('column span') >= 0; }, 8000))) {
      throw new Error('column span field missing');
    }
    var spanRow = Array.from(acc('Layout cell').querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'column span';
    });
    var si = spanRow.querySelector('input');
    si.value = '2';
    si.dispatchEvent(new Event('change', { bubbles: true }));
    if (!(await until(function () {
      var d = frameDoc();
      var firstCell = d && d.querySelector('.sui-layout-cell[data-sui-layout-row="0"][data-sui-layout-col="0"]');
      return firstCell && firstCell.getAttribute('colspan') === '2' && d.querySelectorAll('table.sui-layout-table td').length === 3 ? true : null;
    }, 8000))) throw new Error('span merge did not update the preview');
    await selectCell(0, 0);
    var reset = Array.from(acc('Layout cell').querySelectorAll('button')).find(function (b) { return b.textContent.trim() === 'Reset to 1 x 1'; });
    reset.click();
    if (!(await until(function () {
      var d = frameDoc();
      return d && d.querySelectorAll('table.sui-layout-table td').length === 4 ? true : null;
    }, 8000))) throw new Error('span reset must recreate the released cell');

    // Multi-cell generic styles must apply to every selected cell.
    await selectCell(0, 0);
    await selectCell(0, 1, true);
    var padRow = Array.from(acc('Box model').querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'padding';
    });
    if (!padRow) throw new Error('padding field missing on a multi-cell selection');
    var pi = padRow.querySelector('input');
    pi.value = '12px';
    pi.dispatchEvent(new Event('change', { bubbles: true }));
    if (!(await until(function () {
      var c0 = cell(0, 0), c1 = cell(0, 1);
      return c0 && c1 && /padding:\s*12px/.test(c0.getAttribute('style') || '') && /padding:\s*12px/.test(c1.getAttribute('style') || '') ? true : null;
    }, 8000))) throw new Error('padding was not applied to every selected cell');
    doc = frameDoc();
    var mc0 = cell(0, 0), mc1 = cell(0, 1);
    if (!mc0 || !mc1) throw new Error('cells missing after multi-selection edit');
    // Non-mapped generic CSS properties must reach every selected cell class too.
    var radiusRow = Array.from(acc('Border').querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'radius';
    });
    if (!radiusRow) throw new Error('border radius field missing');
    var radiusInputs = radiusRow.querySelectorAll('input');
    var ri = radiusInputs[radiusInputs.length - 1];
    ri.value = '6px';
    ri.dispatchEvent(new Event('change', { bubbles: true }));
    if (!(await until(function () {
      var d = frameDoc();
      var c0 = cell(0, 0), c1 = cell(0, 1);
      return d && c0 && c1 && d.defaultView.getComputedStyle(c0).borderTopLeftRadius === '6px' && d.defaultView.getComputedStyle(c1).borderTopLeftRadius === '6px' ? true : null;
    }, 8000))) throw new Error('radius was not applied to every selected cell');
    doc = frameDoc();
    mc0 = cell(0, 0); mc1 = cell(0, 1);

    // Split drag: the column widths must be written back.
    // The span reset above re-rendered the preview, so the split handle is
    // not in the frame yet: wait for it instead of reading it once.
    var handle = await until(function () {
      var d = frameDoc();
      var h = d && d.querySelector('.sui-layout-split-v');
      return h ? h : null;
    }, 8000);
    if (!handle) throw new Error('split handle not found');
    doc = frameDoc();
    var hr = handle.getBoundingClientRect();
    handle.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, button: 0, clientX: hr.left + 6, clientY: hr.top + 10 }));
    doc.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: hr.left + 34, clientY: hr.top + 10 }));
    doc.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: hr.left + 34, clientY: hr.top + 10 }));
    if (!(await until(function () {
      var d = frameDoc();
      var colEl = d && d.querySelector('table.sui-layout-table colgroup col');
      return colEl && colEl.getAttribute('style') ? true : null;
    }, 8000))) throw new Error('split drag did not write column widths');
    doc = frameDoc();

    // Reduce rows with content: Move must keep the elements.
    await selectTable();
    var tableRows = secLabels('Table layout');
    if (tableRows.indexOf('preset') >= 0) throw new Error('preset must not appear for a non-empty table');
    var rowField = Array.from(acc('Table layout').querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'rows';
    });
    var ri = rowField.querySelector('input');
    ri.value = '1';
    ri.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(350);
    var dlg = document.getElementById('layoutreducedlg');
    if (!dlg || !dlg.classList.contains('on')) throw new Error('grid reduction dialog did not open');
    var move = document.getElementById('layoutreduce-move');
    if (!move) throw new Error('move content button missing');
    move.click();
    if (!(await until(function () {
      var d = frameDoc();
      return d && d.querySelectorAll('table.sui-layout-table tbody tr').length === 1 && d.querySelectorAll('table.sui-layout-table .sui-card').length === 3 ? true : null;
    }, 8000))) throw new Error('grid reduction did not apply and keep all card content');
    doc = frameDoc();

    // A split that would go below the minimum asks first; Cancel keeps the grid.
    await selectCell(0, 0);
    var beforeSplit = frameDoc().querySelectorAll('table.sui-layout-table td').length;
    var splitBtn = Array.from(acc('Table layout').querySelectorAll('button')).find(function (b) { return b.textContent.trim() === 'Vertical'; });
    if (!splitBtn) throw new Error('split button missing');
    splitBtn.click();
    await wait(350);
    var warning = document.getElementById('confdlg');
    if (!warning || !warning.classList.contains('on')) {
      var wdoc2 = frameDoc();
      var selCell = wdoc2.querySelector('.sui-layout-cell[data-sui-layout-row="0"][data-sui-layout-col="0"]');
      throw new Error('small split warning did not open (selected width ' + (selCell ? selCell.getBoundingClientRect().width : -1) + ')');
    }
    document.getElementById('conf-cancel').click();
    await wait(400);
    if (frameDoc().querySelectorAll('table.sui-layout-table td').length !== beforeSplit) throw new Error('cancelled split changed the grid');

    // The Border section on a layout table styles the table frame.
    await selectTable();
    var border = acc('Border');
    if (!border) throw new Error('Border section missing on the table');
    var bStyleRow = Array.from(border.querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'style';
    });
    var bStyle = bStyleRow && bStyleRow.querySelector('select');
    bStyle.value = 'solid';
    bStyle.dispatchEvent(new Event('change', { bubbles: true }));
    if (!(await until(function () {
      var d = frameDoc();
      var t = d && d.querySelector('table.sui-layout-table');
      return t && d.defaultView.getComputedStyle(t).borderTopStyle === 'solid' ? true : null;
    }, 8000))) throw new Error('table frame border style not applied');
    border = acc('Border');
    var bWidthRow = Array.from(border.querySelectorAll('.rowline')).find(function (r) {
      var l = r.querySelector('label'); return l && l.textContent === 'width';
    });
    var bWidth = bWidthRow && bWidthRow.querySelector('input');
    bWidth.value = '3px';
    bWidth.dispatchEvent(new Event('change', { bubbles: true }));
    if (!(await until(function () {
      var d = frameDoc();
      var t = d && d.querySelector('table.sui-layout-table');
      return t && parseFloat(d.defaultView.getComputedStyle(t).borderTopWidth) > 0 ? true : null;
    }, 8000))) throw new Error('table frame border width not applied');
    border = acc('Border');
    var bHex = border && border.querySelector('.color-hex');
    bHex.value = '#ff0000';
    bHex.dispatchEvent(new Event('change', { bubbles: true }));
    if (!(await until(function () {
      var d = frameDoc();
      var t = d && d.querySelector('table.sui-layout-table');
      return t && d.defaultView.getComputedStyle(t).borderTopColor === 'rgb(255, 0, 0)' ? true : null;
    }, 8000))) throw new Error('table frame border color not applied');
    doc = frameDoc();
    var frameTable = doc.querySelector('table.sui-layout-table');
    var fcs = doc.defaultView.getComputedStyle(frameTable);
    if (fcs.borderTopStyle !== 'solid') throw new Error('table frame border style not applied');
    if (parseFloat(fcs.borderTopWidth) <= 0) throw new Error('table frame border width not applied');

    // Panels with an internal preview stay modal.
    openPrefs();
    await wait(300);
    var pref = document.getElementById('pref');
    if (pref.classList.contains('settings-floating')) throw new Error('Preferences must stay modal');
    if (window.getComputedStyle(pref).pointerEvents !== 'auto') throw new Error('modal backdrop must block the editor');
    if (!document.getElementById('prefbox').querySelector('.settings-grip')) throw new Error('Preferences grip missing');
    closePrefs();

    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
