package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// treeRuntimeE2EScript drives the compiled tree runtime in a real browser:
// context-menu structural edits, empty-parent cleanup, onTreeChange counting
// and tree.lastChange metadata.
const treeRuntimeE2EScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-result';
      document.body.appendChild(el);
    }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function itemByKey(key) {
    var all = document.querySelectorAll('.sui-tree-item');
    for (var i = 0; i < all.length; i++) {
      if ((all[i].getAttribute('data-sui-key') || '') === key) return all[i];
    }
    return null;
  }
  function version() {
    var ps = document.querySelectorAll('p');
    for (var i = 0; i < ps.length; i++) {
      if ((ps[i].textContent || '').indexOf('Version:') >= 0) return ps[i].textContent.trim();
    }
    return 'none';
  }
  function lastChange() {
    var pair = window.__SUI_TREE.lastChange('t');
    if (!pair || pair[1]) throw new Error('lastChange failed: ' + (pair && pair[1]));
    return JSON.parse(pair[0]);
  }
  function rightClick(li) {
    if (!li) throw new Error('item not found');
    var r = li.getBoundingClientRect();
    li.querySelector('.sui-tree-row').dispatchEvent(new MouseEvent('contextmenu', {
      bubbles: true, cancelable: true, button: 2, clientX: r.left + 8, clientY: r.top + 8
    }));
  }
  function menuClick(label) {
    var menu = document.getElementById('sui-tree-context');
    if (!menu) throw new Error('context menu not open for ' + label);
    var buttons = menu.querySelectorAll('button');
    for (var i = 0; i < buttons.length; i++) {
      if (buttons[i].textContent === label) { buttons[i].click(); return; }
    }
    throw new Error('menu item not found: ' + label);
  }
  function rowIcon(li) {
    var el = li && li.querySelector(':scope > .sui-tree-row > .sui-tree-icon');
    return el ? el.textContent : '';
  }
  function rowToggle(li) {
    return li ? li.querySelector(':scope > .sui-tree-row > .sui-tree-toggle') : null;
  }
  try {
    for (var i = 0; i < 150 && !itemByKey('a1'); i++) await wait(100);
    if (!itemByKey('a1')) throw new Error('tree did not render');
    if (version() !== 'Version: 0') throw new Error('initial version is ' + version());

    // Delete the only child: A must become a leaf (leaf icon, no toggle).
    rightClick(itemByKey('a1'));
    await wait(150);
    menuClick('Delete item');
    await wait(700);
    if (version() !== 'Version: 1') throw new Error('delete did not bump version: ' + version());
    var lc = lastChange();
    if (lc.op !== 'delete' || lc.key !== 'a1' || lc.parent !== 'a') {
      throw new Error('unexpected delete lastChange: ' + JSON.stringify(lc));
    }
    var a = itemByKey('a');
    if (rowIcon(a) !== 'L') throw new Error('parent did not become leaf: icon ' + rowIcon(a));
    if (rowToggle(a)) throw new Error('leaf still has a toggle');

    // Add a child back: A becomes a parent with the open-parent icon.
    rightClick(a);
    await wait(150);
    menuClick('Add child item');
    await wait(700);
    if (version() !== 'Version: 2') throw new Error('add did not bump version: ' + version());
    lc = lastChange();
    if (lc.op !== 'add' || lc.parent !== 'a') throw new Error('unexpected add lastChange: ' + JSON.stringify(lc));
    a = itemByKey('a');
    if (rowIcon(a) !== 'PO') throw new Error('open parent icon expected, got ' + rowIcon(a));
    if (!rowToggle(a) || rowToggle(a).textContent !== '-') throw new Error('open branch symbol missing');

    // Rename A: onTreeChange reports the rename metadata.
    var aLabel = a.querySelector(':scope > .sui-tree-row > .sui-tree-label');
    aLabel.dispatchEvent(new MouseEvent('dblclick', { bubbles: true, cancelable: true }));
    await wait(300);
    var renameInput = a.querySelector('input.sui-tree-rename');
    if (!renameInput) throw new Error('rename input did not open');
    renameInput.value = 'A renamed';
    renameInput.dispatchEvent(new Event('input', { bubbles: true }));
    renameInput.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await wait(700);
    if (version() !== 'Version: 3') throw new Error('rename did not bump version: ' + version());
    lc = lastChange();
    if (lc.op !== 'rename' || lc.key !== 'a' || lc.label !== 'A renamed') {
      throw new Error('unexpected rename lastChange: ' + JSON.stringify(lc));
    }

    // Move B up.
    rightClick(itemByKey('b'));
    await wait(150);
    menuClick('Move up');
    await wait(700);
    if (version() !== 'Version: 4') throw new Error('move did not bump version: ' + version());
    lc = lastChange();
    if (lc.op !== 'move' || lc.key !== 'b') throw new Error('unexpected move lastChange: ' + JSON.stringify(lc));

    // Import replaces the tree and reports op import.
    var imported = window.__SUI_TREE.importJson('t', JSON.stringify([{ label: 'Imported', key: 'imp' }]));
    if (!imported || imported[1]) throw new Error('import failed: ' + (imported && imported[1]));
    await wait(700);
    if (version() !== 'Version: 5') throw new Error('import did not bump version: ' + version());
    lc = lastChange();
    if (lc.op !== 'import' || lc.count !== 1) throw new Error('unexpected import lastChange: ' + JSON.stringify(lc));
    if (!itemByKey('imp')) throw new Error('imported item missing');

    // Clear reports op clear and still exposes lastChange on an empty tree.
    var cleared = window.__SUI_TREE.clear('t');
    if (!cleared || cleared[1]) throw new Error('clear failed: ' + (cleared && cleared[1]));
    await wait(700);
    if (version() !== 'Version: 6') throw new Error('clear did not bump version: ' + version());
    lc = lastChange();
    if (lc.op !== 'clear') throw new Error('unexpected clear lastChange: ' + JSON.stringify(lc));
    if (document.querySelectorAll('.sui-tree-item').length) throw new Error('tree not empty after clear');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`

func TestTreeRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping tree runtime end-to-end test")
	}

	const src = `state version = 0
fn changed() -> null {
	version = version + 1
}
view main {
	tree "T" name: "t" contextMenu: true editMode: "editable" onTreeChange: changed iconLeaf: "L" iconParent: "P" iconParentOpen: "PO" toggleClosed: "+" toggleOpen: "-" {
		treeitem "A" key: "a" open: true {
			treeitem "A1" key: "a1"
		}
		treeitem "B" key: "b"
	}
	p "Version: {version}"
}
`
	prog, dl := ParseSource(src, "tree_runtime_e2e.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	css := GenerateCss(prog)
	html := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><style>" + css +
		"</style></head><body><div id=\"app\"></div><script>" + js + "</script>" +
		treeRuntimeE2EScript + "</body></html>"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	profile := t.TempDir()
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--no-first-run",
		"--no-default-browser-check",
		"--user-data-dir=" + profile,
		"--virtual-time-budget=25000",
		"--dump-dom",
		srv.URL + "/",
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
	t.Fatalf("tree runtime e2e failed:\n%s", detail)
}
