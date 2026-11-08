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

const tableRuntimeE2EScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function table(name) { return document.querySelector('table[name="' + name + '"]'); }
  function rows(name) { return Array.prototype.slice.call(table(name).querySelectorAll('tbody tr')); }
  function texts(name) { return rows(name).map(function (tr) { return Array.prototype.slice.call(tr.children).map(function (c) { return c.textContent; }); }); }
  function header(name) { return Array.prototype.slice.call(table(name).querySelectorAll('thead th')).map(function (c) { return c.textContent; }); }
  function version() { var m = (document.body.textContent || '').match(/Version: (\d+)/); return m ? m[1] : 'none'; }
  function cellVersion() { var m = (document.body.textContent || '').match(/Cell version: (\d+)/); return m ? m[1] : 'none'; }
  function lastChange(name) { var p = window.__SUI_TBL.lastChange(name); if (!p || p[1]) throw new Error('lastChange failed'); return JSON.parse(p[0]); }
  try {
    for (var i = 0; i < 150 && !table('users'); i++) await wait(100);
    if (!table('users')) throw new Error('tables did not render');
    var d = window.__SUI_TBL.display('users', JSON.stringify([
      { id: 1, name: 'Alice', status: 'active', amount: 1200.5 },
      { id: 2, name: 'Bob', status: 'paused', amount: 300 },
      { id: 3, name: 'Cara', status: 'active', amount: 99 }
    ]));
    if (!d || d[1]) throw new Error('display failed: ' + (d && d[1]));
    await wait(400);
    if (header('users').join(',') !== 'ID,Name,Status,Amount') throw new Error('declarative labels missing: ' + header('users').join(','));
    if (rows('users').length !== 2) throw new Error('page size 2 not applied: ' + rows('users').length);
    if (!document.querySelector('.sui-table-filter[data-sui-table-filter-for="users"]')) throw new Error('filter input missing');
    if (!document.querySelector('.sui-table-pager[data-sui-table-pager-for="users"]')) throw new Error('pager missing');
    if (version() !== '1') throw new Error('display did not fire onTableChange: ' + version());
    if (texts('users')[0][3].indexOf('€') < 0) throw new Error('currency formatting missing: ' + texts('users')[0][3]);

    if (!window.__SUI_TBL.sort('users', 'status', 'asc')[0]) throw new Error('sort failed');
    await wait(300);
    if (texts('users')[0][2] !== 'active' || texts('users')[1][2] !== 'active') throw new Error('sort order wrong: ' + JSON.stringify(texts('users')));
    if (lastChange('users').op !== 'sort') throw new Error('sort lastChange missing');
    if (version() !== '1') throw new Error('sort must not fire onTableChange');

    if (!window.__SUI_TBL.filterText('users', 'active')[0]) throw new Error('filter failed');
    await wait(300);
    if (rows('users').length !== 2) throw new Error('filter rows wrong: ' + rows('users').length);
    window.__SUI_TBL.filterText('users', '');
    await wait(300);
    if (!window.__SUI_TBL.page('users', 1)[0]) throw new Error('page failed');
    await wait(300);
    if (rows('users').length !== 1) throw new Error('page 2 rows wrong: ' + rows('users').length);
    window.__SUI_TBL.page('users', 0);
    await wait(300);

    rows('users')[1].querySelectorAll('td')[1].click();
    await wait(300);
    var sel = window.__SUI_TBL.selectedRows('users')[0];
    if (!sel || sel.length !== 1 || sel[0] !== 1) throw new Error('row selection wrong: ' + JSON.stringify(sel));

    var ed = table('editable');
    var nameCell = ed.querySelectorAll('tbody tr td')[0];
    nameCell.dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    nameCell.textContent = 'Alicia';
    nameCell.dispatchEvent(new Event('input', { bubbles: true }));
    nameCell.dispatchEvent(new FocusEvent('focusout', { bubbles: true, relatedTarget: null }));
    await wait(300);
    if (cellVersion() !== '1') throw new Error('onCellChange did not fire: ' + cellVersion());
    var lc = lastChange('editable');
    if (lc.op !== 'cell' || lc.old !== 'Alice' || lc.new !== 'Alicia') throw new Error('cell lastChange wrong: ' + JSON.stringify(lc));

    var amountCell = table('editable').querySelectorAll('tbody tr td')[1];
    amountCell.dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    amountCell.textContent = 'abc';
    amountCell.dispatchEvent(new Event('input', { bubbles: true }));
    amountCell.dispatchEvent(new FocusEvent('focusout', { bubbles: true, relatedTarget: null }));
    await wait(250);
    if (amountCell.getAttribute('aria-invalid') !== 'true') throw new Error('invalid number not marked');
    if (window.__SUI_TBL.data('editable')[0][0][1] !== '10') throw new Error('invalid value was committed: ' + JSON.stringify(window.__SUI_TBL.data('editable')[0]));
    if (cellVersion() !== '1') throw new Error('invalid edit must not fire onCellChange');

    // Runtime settings panel: per-column header filters and multi-column sort.
    var gear = document.querySelector('.sui-table-tools .sui-table-settings');
    if (!gear) throw new Error('table settings gear missing');
    gear.click();
    await wait(300);
    var cfgPanel = document.querySelector('.sui-table-cfg');
    if (!cfgPanel) throw new Error('table settings panel missing');
    var toggles = cfgPanel.querySelectorAll('input[type=checkbox]');
    if (toggles.length !== 2) throw new Error('table settings toggles missing');
    if (!toggles[0].checked) throw new Error('filter toggle must follow filterable');
    if (!table('users').querySelector('.sui-table-filterbtn')) throw new Error('header filter button missing');
    if (!table('users').querySelector('.sui-table-sortbtn')) throw new Error('header sort button missing');
    cfgPanel.querySelector('.sui-table-cfg-close').click();
    await wait(200);
    if (document.querySelector('.sui-table-cfg')) throw new Error('settings panel did not close');

    if (!window.__SUI_TBL.filter('users', 'status', 'paused')[0]) throw new Error('table.filter failed');
    await wait(300);
    if (rows('users').length !== 1) throw new Error('column filter rows wrong: ' + rows('users').length);
    if (rows('users')[0].children[2].textContent !== 'paused') throw new Error('column filter selected wrong rows');
    window.__SUI_TBL.filter('users', 'status', '');
    await wait(250);

    // Excel-like value filter popup: search, Select All, value checkboxes,
    // Clear Filter and Apply.
    var statusTh = Array.prototype.slice.call(table('users').querySelectorAll('thead th')).filter(function(th){
      var l = th.querySelector('.sui-table-hlabel');
      return (l ? l.textContent : th.textContent) === 'Status';
    })[0];
    if (!statusTh) throw new Error('status header missing');
    var fbtn = statusTh.querySelector('.sui-table-filterbtn');
    if (!fbtn) throw new Error('status filter button missing');
    fbtn.click();
    await wait(300);
    var fpop = document.querySelector('.sui-table-fpop');
    if (!fpop) throw new Error('filter popup missing');
    if (fpop.querySelectorAll('.sui-table-fpop-item').length < 3) throw new Error('filter values missing');
    var fsearch = fpop.querySelector('.sui-table-fpop-search');
    if (!fsearch) throw new Error('filter search missing');
    fsearch.value = 'pau';
    fsearch.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(150);
    var visibleValues = Array.prototype.slice.call(fpop.querySelectorAll('.sui-table-fpop-item')).filter(function(r){ return r.style.display !== 'none'; });
    if (visibleValues.length !== 2) throw new Error('filter search did not narrow: ' + visibleValues.length);
    fsearch.value = '';
    fsearch.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(150);
    var fitems = Array.prototype.slice.call(fpop.querySelectorAll('.sui-table-fpop-item'));
    var activeRow = fitems.filter(function(r){ var l = r.querySelector('.sui-table-fpop-lbl'); return l && l.textContent === 'active'; })[0];
    if (!activeRow) throw new Error('active filter value missing');
    var activeChk = activeRow.querySelector('input');
    activeChk.checked = false;
    activeChk.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(100);
    var applyBtn = Array.prototype.slice.call(fpop.querySelectorAll('button')).filter(function(b){ return b.textContent === 'Apply'; })[0];
    if (!applyBtn) throw new Error('filter apply missing');
    applyBtn.click();
    await wait(400);
    if (document.querySelector('.sui-table-fpop')) throw new Error('filter popup did not close');
    if (rows('users').length !== 1) throw new Error('value filter rows wrong: ' + rows('users').length);
    if (rows('users')[0].children[2].textContent !== 'paused') throw new Error('value filter selected wrong rows');
    fbtn = table('users').querySelector('thead th:nth-child(3) .sui-table-filterbtn');
    fbtn.click();
    await wait(250);
    fpop = document.querySelector('.sui-table-fpop');
    var clearFilter = fpop ? Array.prototype.slice.call(fpop.querySelectorAll('button')).filter(function(b){ return b.textContent === 'Clear Filter'; })[0] : null;
    if (!clearFilter) throw new Error('filter clear missing');
    clearFilter.click();
    await wait(400);
    if (rows('users').length !== 2) throw new Error('value filter clear failed: ' + rows('users').length);

    gear = document.querySelector('.sui-table-tools .sui-table-settings');
    gear.click();
    await wait(250);
    cfgPanel = document.querySelector('.sui-table-cfg');
    var addLevel = cfgPanel ? Array.prototype.slice.call(cfgPanel.querySelectorAll('button')).filter(function(b){ return b.textContent === '+ Add level'; })[0] : null;
    if (!addLevel) throw new Error('sort level action missing');
    var levelsBefore = cfgPanel.querySelectorAll('.sui-table-cfg-level').length;
    addLevel.click();
    await wait(450);
    cfgPanel = document.querySelector('.sui-table-cfg');
    var levelRows = cfgPanel ? cfgPanel.querySelectorAll('.sui-table-cfg-level') : [];
    if (levelRows.length !== levelsBefore + 1) throw new Error('sort level not added: ' + (cfgPanel ? levelRows.length + ' vs ' + levelsBefore : 'panel closed'));
    var lastSelects = levelRows[levelRows.length - 1].querySelectorAll('select');
    if (lastSelects.length < 2) throw new Error('sort level selects missing');
    var amountIdx = header('users').indexOf('Amount');
    if (amountIdx < 0) throw new Error('amount column label missing: ' + header('users').join(','));
    lastSelects[0].value = String(amountIdx);
    lastSelects[0].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(450);
    // Existing level is Status asc; the new one is Amount asc: active rows first
    // (Cara 99, Alice 1200.5), then the paused row.
    if (rows('users').length < 2) throw new Error('sorted page too short');
    if (rows('users')[0].children[1].textContent !== 'Cara' || rows('users')[1].children[1].textContent !== 'Alice') {
      throw new Error('multi-sort levels not applied: ' + texts('users').map(function(r){ return r[1]; }).join(','));
    }
    cfgPanel = document.querySelector('.sui-table-cfg');
    var clearSort = cfgPanel ? Array.prototype.slice.call(cfgPanel.querySelectorAll('button')).filter(function(b){ return b.textContent === 'Clear sort'; })[0] : null;
    if (!clearSort) throw new Error('clear sort action missing');
    clearSort.click();
    await wait(450);
    cfgPanel = document.querySelector('.sui-table-cfg');
    if (cfgPanel) cfgPanel.querySelector('.sui-table-cfg-close').click();
    await wait(200);

    // Runtime table creation and structure editing.
    var tabsBefore = document.querySelectorAll('.sui-tablist .sui-tab').length;
    var addBtn = document.querySelector('.sui-tab-add');
    if (!addBtn) throw new Error('add table button missing');
    addBtn.click();
    await wait(500);
    if (document.querySelectorAll('.sui-tablist .sui-tab').length !== tabsBefore + 1) throw new Error('runtime table not created');
    var createdTab = document.querySelectorAll('.sui-tablist .sui-tab')[tabsBefore];
    createdTab.click();
    await wait(300);
    var createdPanel = document.querySelectorAll('.sui-tabpanel')[tabsBefore];
    var createdTable = createdPanel ? createdPanel.querySelector('table') : null;
    if (!createdTable) throw new Error('runtime sheet table missing');
    var createdHead = Array.prototype.slice.call(createdTable.querySelectorAll('thead th')).map(function(th){
      var l = th.querySelector('.sui-table-hlabel');
      return (l ? l.textContent : th.textContent).trim();
    });
    if (createdHead.join(',') !== 'A,B,C,D') throw new Error('runtime sheet inherited header: ' + createdHead.join(','));
    if (createdTable.querySelectorAll('tbody tr').length !== 2) throw new Error('runtime sheet must start with two empty rows');
    document.querySelectorAll('.sui-tablist .sui-tab')[0].click();
    await wait(250);
    var ucell = table('users').querySelector('tbody td');
    var ur = ucell.getBoundingClientRect();
    ucell.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, cancelable: true, button: 2, clientX: ur.left + 8, clientY: ur.top + 8 }));
    await wait(200);
    var sm = document.getElementById('sui-table-struct');
    if (!sm) throw new Error('structure context menu missing');
    var insRow = Array.prototype.slice.call(sm.querySelectorAll('button')).filter(function(b){ return b.textContent === 'Insert row above'; })[0];
    if (!insRow) throw new Error('insert row action missing');
    insRow.click();
    await wait(500);
    if (version() !== '2') throw new Error('row insert did not fire onTableChange: ' + version());
    if (document.querySelectorAll('.sui-tablist .sui-tab').length !== tabsBefore + 1) throw new Error('runtime table lost after re-render');
    document.querySelectorAll('.sui-tablist .sui-tab')[0].click();
    await wait(250);
    var th = table('users').querySelector('thead th');
    var tr = th.getBoundingClientRect();
    th.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, cancelable: true, button: 2, clientX: tr.left + 8, clientY: tr.top + 8 }));
    await wait(200);
    var sm2 = document.getElementById('sui-table-struct');
    var insCol = sm2 ? Array.prototype.slice.call(sm2.querySelectorAll('button')).filter(function(b){ return b.textContent === 'Insert column right'; })[0] : null;
    if (!insCol) throw new Error('insert column action missing');
    insCol.click();
    await wait(500);
    if (table('users').querySelectorAll('thead th').length !== 5) throw new Error('column insert failed');
    if (version() !== '3') throw new Error('column insert did not fire onTableChange: ' + version());

    // The selected runtime sheet must stay active synchronously after the
    // reactive re-render triggered by a cell edit (no first-sheet flash).
    var runtimeSheetTab = document.querySelectorAll('.sui-tablist .sui-tab')[tabsBefore];
    runtimeSheetTab.click();
    await wait(300);
    var runtimePanel = document.querySelectorAll('.sui-tabpanel')[tabsBefore];
    var runtimeCell = runtimePanel.querySelectorAll('tbody td')[1];
    runtimeCell.dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    runtimeCell.textContent = 'persisted';
    runtimeCell.dispatchEvent(new Event('input', { bubbles: true }));
    runtimeCell.dispatchEvent(new FocusEvent('focusout', { bubbles: true, relatedTarget: null }));
    var syncTabs = document.querySelectorAll('.sui-tablist .sui-tab');
    if (!syncTabs[tabsBefore] || !syncTabs[tabsBefore].classList.contains('active')) {
      throw new Error('selected sheet is not restored synchronously after re-render');
    }
    await wait(400);
    var settledTabs = document.querySelectorAll('.sui-tablist .sui-tab');
    if (!settledTabs[tabsBefore] || !settledTabs[tabsBefore].classList.contains('active')) {
      throw new Error('selected sheet lost after re-render');
    }
    if (document.querySelectorAll('.sui-tabpanel')[tabsBefore].querySelectorAll('tbody td')[1].textContent !== 'persisted') {
      var dbgNames = Array.prototype.slice.call(document.querySelectorAll('table[name]')).map(function(t){ return t.getAttribute('name'); });
      var dbgData = window.__SUI_TBL.data('table-1');
      throw new Error('runtime sheet cell edit lost after re-render: names=' + dbgNames.join(',') + ' tabs=' + Array.prototype.slice.call(document.querySelectorAll('.sui-tablist .sui-tab')).map(function(t){ return t.textContent; }).join(',') + ' data=' + JSON.stringify(dbgData));
    }

    // Runtime header edits survive blur and reactive re-render.
    var headCell = table('editable').querySelector('thead th');
    headCell.dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    headCell.textContent = 'RenamedHeader';
    headCell.dispatchEvent(new Event('input', { bubbles: true }));
    headCell.dispatchEvent(new FocusEvent('focusout', { bubbles: true, relatedTarget: null }));
    await wait(600);
    var headAfter = table('editable').querySelector('thead th');
    if (!headAfter || headAfter.textContent.indexOf('RenamedHeader') < 0) throw new Error('runtime header edit lost: ' + (headAfter ? headAfter.textContent : 'missing'));
    if ((table('editable').getAttribute('data-sui-column-labels') || '').indexOf('RenamedHeader') < 0) throw new Error('runtime header edit not persisted');
    var editedHeader = window.__SUI_TBL.header('editable')[0];
    if (!editedHeader || editedHeader[0] !== 'RenamedHeader') throw new Error('header API lost runtime edit');
    // Headers are labels, not typed numbers: the second editable column is a
    // number column and must still accept a textual header label.
    var head2 = table('editable').querySelectorAll('thead th')[1];
    head2.dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    head2.textContent = 'AmountRenamed';
    head2.dispatchEvent(new Event('input', { bubbles: true }));
    head2.dispatchEvent(new FocusEvent('focusout', { bubbles: true, relatedTarget: null }));
    await wait(600);
    var head2After = table('editable').querySelectorAll('thead th')[1];
    if (!head2After || head2After.textContent.indexOf('AmountRenamed') < 0) throw new Error('numeric-column header edit rejected');
    var editedHeader2 = window.__SUI_TBL.header('editable')[0];
    if (!editedHeader2 || editedHeader2[1] !== 'AmountRenamed') throw new Error('numeric-column header API lost runtime edit');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`

func TestTableRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping table runtime end-to-end test")
	}
	const src = `state version = 0
state cellVersion = 0
fn tableChanged() -> null {
	version = version + 1
}
fn cellChanged() -> null {
	cellVersion = cellVersion + 1
}
fn sorted() -> null {
}
view main {
	tabs "Book" position: "bottom" variant: "sheets" {
		tab "Users" {
			table name: "users" columns: "id,name,status,amount" columnLabels: "ID,Name,Status,Amount" columnWidths: "60,160,100,110" columnAlign: "right,left,center,right" columnTypes: "number,text,select,number" columnFormats: "text,text,text,currency:EUR" columnOptions: "status=active|paused" sortable: true filterable: true pageSize: 2 selection: "row" multiSelect: true showAddTable: true structureEditable: true onTableChange: tableChanged onSort: sorted {
				thead {
					trow {
						tcell "ID" head: true
						tcell "Name" head: true
						tcell "Status" head: true
						tcell "Amount" head: true
					}
				}
				tbody {
					trow {
						tcell ""
						tcell ""
						tcell ""
						tcell ""
					}
				}
			}
		}
	}
	table name: "editable" editable: true columns: "name,amount" columnLabels: "Name,Amount" columnTypes: "text,number" columnFormats: "text,currency:EUR" onCellChange: cellChanged {
		thead {
			trow {
				tcell "Name" head: true
				tcell "Amount" head: true
			}
		}
		tbody {
			trow {
				tcell "Alice"
				tcell "10"
			}
		}
	}
	p "Version: {version}"
	p "Cell version: {cellVersion}"
}
`
	prog, dl := ParseSource(src, "table_runtime_e2e.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><style>" + GenerateCss(prog) +
		"</style></head><body><div id=\"app\"></div><script>" + GenerateJs(prog) + "</script>" +
		tableRuntimeE2EScript + "</body></html>"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=25000", "--dump-dom",
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
	t.Fatalf("table runtime e2e failed:\n%s", detail)
}
