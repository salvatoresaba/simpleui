package core

import "testing"

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
	runDesignerE2E(t, designerGridSource, designerGridScript)
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
    }, 20000);
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
    }, 20000);
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
