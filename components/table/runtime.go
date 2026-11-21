package table

// Runtime is the table runtime: it binds tabular data to the table
// named with `name:` (table.display / table.data / ...), keeps the designer's
// row and cell styling, re-applies displayed data after every render, and
// wires the editable/fixed behaviors (contenteditable, truncation overlay,
// draggable row/column boundaries).
const Runtime = `const __SUI_TBL = (function(){
  // ---- data binding --------------------------------------------------------
  // store[name] = { header: [] | null, rows: [[...]] }. It survives view
  // re-renders: afterRender() re-applies it to the freshly rendered table.
  var store = Object.create(null);
  var changes = Object.create(null);
  var tableClip = Object.create(null);
  var tableBooks = Object.create(null);
  var tpls = new WeakMap(); // table element -> { heads, bodies } row templates
  // Resource limits: the table is a DOM widget, so huge payloads are refused
  // instead of freezing the page. Applications should paginate.
  var MAX_ROWS = (__SUI_SEC && __SUI_SEC.maxRows) || 100000;
  var MAX_COLS = (__SUI_SEC && __SUI_SEC.maxCols) || 10000;
  var MAX_TEXT = (__SUI_SEC && __SUI_SEC.maxText) || 20000000;
  // Split so the generated component JS never contains the literal attribute
  // name; component-only apps rely on that (inline handlers, no query ids).
  var TBL_ID_ATTR = 'data-' + 'sui-id';

  function splitList(v){ return String(v || '').split(',').map(function(s){ return s.trim(); }).filter(function(s){ return s !== ''; }); }
  function intList(v){ return splitList(v).map(function(s){ return parseInt(s, 10) || 0; }); }
  function parseColumnOptions(v){
    var out = {};
    var parts = String(v || '').split(';');
    for (var i = 0; i < parts.length; i++) {
      var eq = parts[i].indexOf('=');
      if (eq <= 0) continue;
      var k = parts[i].slice(0, eq).trim();
      var vals = parts[i].slice(eq + 1).split('|').map(function(s){ return s.trim(); });
      if (k) out[k] = vals;
    }
    return out;
  }
  function cfgOf(tbl){
    function attr(k){ return tbl.getAttribute('data-sui-' + k) || ''; }
    var mode = attr('edit-mode');
    if (mode !== 'editable' && mode !== 'persistent') mode = 'none';
    var editable = mode !== 'none';
    var cfg = {
      columns: splitList(attr('columns')),
      labels: splitList(attr('column-labels')),
      widths: intList(attr('column-widths')),
      align: splitList(attr('column-align')),
      types: splitList(attr('column-types')),
      formats: splitList(attr('column-formats')),
      options: parseColumnOptions(attr('column-options')),
      sortable: attr('sortable') === '1',
      filterable: attr('filterable') === '1',
      pageSize: parseInt(attr('page-size'), 10) || 0,
      selection: attr('selection') || 'none',
      multi: attr('multi') === '1',
      editMode: mode,
      editable: editable,
      fixed: attr('fixed') === '1',
      showAddTable: editable && attr('show-add-table') === '1',
      structureEditable: editable && attr('structure-editable') === '1'
    };
    return cfg;
  }
  function tableName(tbl){ return tbl.getAttribute('name') || ''; }
  function firstTable(name){ var ts = findTables(name); return ts.length ? ts[0] : null; }
  function tableValue(tbl, key){
    var a = tbl.getAttribute('data-sui-table-' + key) || '';
    return a;
  }
  function fireTableEvent(name, key){
    var tbl = firstTable(name);
    if (!tbl) return;
    var fn = tableValue(tbl, key);
    if (fn && typeof window[fn] === 'function') window[fn]();
  }
  // ---- persistent edit mode ------------------------------------------------
  // Persistent tables store their header and rows in localStorage, keyed by
  // table name and page. Editable (session) tables keep the in-memory store;
  // they return to the source data on reload.
  var tableBootAt = Date.now();
  var tableUserTouched = false;
  function markTableUserTouched(){ tableUserTouched = true; }
  document.addEventListener('pointerdown', markTableUserTouched, true);
  document.addEventListener('keydown', markTableUserTouched, true);
  function tablePersistKey(name){
    var path = (window.location && window.location.pathname) || '';
    return 'sui.table.persist.' + path + '.' + name;
  }
  function captureColumnWidths(tbl){
    var out = [];
    if (!tbl) return out;
    var row = tbl.querySelector('thead tr') || (tbl.tBodies && tbl.tBodies[0] ? tbl.tBodies[0].rows[0] : null);
    if (!row) return out;
    for (var i = 0; i < row.cells.length; i++) {
      var w = parseFloat(row.cells[i].style.width || row.cells[i].style.minWidth || '');
      out.push(isFinite(w) ? Math.round(w) : 0);
    }
    return out;
  }
  function captureRowHeights(tbl){
    var out = [];
    if (!tbl) return out;
    var rows = bodyRows(tbl);
    for (var i = 0; i < rows.length; i++) {
      var h = parseFloat(rows[i].style.height || '');
      out.push(isFinite(h) ? Math.round(h) : 0);
    }
    return out;
  }
  function persistTableState(name, st, cfg){
    if (!cfg || cfg.editMode !== 'persistent') return;
    if (!window.localStorage) return;
    var tbl = firstTable(name);
    try {
      window.localStorage.setItem(tablePersistKey(name), JSON.stringify({
        version: 2,
        header: st && st.header ? st.header : [],
        rows: st && st.allRows ? st.allRows : [],
        textHeader: !!(st && st.textHeader),
        columnWidths: captureColumnWidths(tbl),
        rowHeights: captureRowHeights(tbl),
        cellText: tbl ? (tbl.getAttribute('data-sui-cell-text') || '') : '',
        sortKeys: st ? sortKeysOf(st) : [],
        filter: st ? (st.filter || '') : '',
        colFilters: st ? (st.colFilters || {}) : {},
        colValueFilters: st ? (st.colValueFilters || {}) : {},
        page: st ? (st.page || 0) : 0
      }));
    } catch (e) {}
  }
  function persistTable(name){
    var tbl = firstTable(name);
    if (!tbl) return;
    persistTableState(name, store[name], cfgOf(tbl));
  }
  function loadPersistedTable(name){
    try {
      var raw = window.localStorage && window.localStorage.getItem(tablePersistKey(name));
      if (!raw) return null;
      var p = JSON.parse(raw);
      if (!p || !Array.isArray(p.rows)) return null;
      return p;
    } catch (e) { return null; }
  }
  function applyPersistedTable(st, p){
    st.header = Array.isArray(p.header) ? p.header.map(function(v){ return String(v); }) : [];
    st.allRows = p.rows.map(function(r){ return Array.isArray(r) ? r.slice() : []; });
    st.rows = st.allRows;
    st.textHeader = !!p.textHeader;
    st.__fromDom = false;
    st.__colStruct = false;
    st.persistWidths = Array.isArray(p.columnWidths) ? p.columnWidths.slice() : [];
    st.persistRowHeights = Array.isArray(p.rowHeights) ? p.rowHeights.slice() : [];
    st.persistCellText = p.cellText || '';
    setSortKeys(st, Array.isArray(p.sortKeys) ? p.sortKeys : []);
    st.filter = p.filter || '';
    st.colFilters = p.colFilters && typeof p.colFilters === 'object' ? p.colFilters : {};
    st.colValueFilters = p.colValueFilters && typeof p.colValueFilters === 'object' ? p.colValueFilters : {};
    st.page = p.page || 0;
    clearSelection(st);
  }
  function applyPersistedTableAttrs(tbl, st){
    if (!tbl || !st) return;
    if (st.persistWidths && st.persistWidths.length) {
      tbl.setAttribute('data-sui-column-widths', st.persistWidths.join(','));
    }
    if (st.persistCellText === 'truncate' || st.persistCellText === 'wrap') {
      tbl.setAttribute('data-sui-cell-text', st.persistCellText);
      if (st.persistCellText === 'truncate') tbl.setAttribute('data-sui-fixed', '1');
    }
  }
  function captureTableSource(tbl, st){
    if (!st || st.__sourceRows) return;
    st.__sourceHeader = (st.header || []).slice();
    st.__sourceRows = (st.allRows || []).map(function(r){ return r.slice(); });
    st.__sourceWidthsAttr = tbl ? (tbl.getAttribute('data-sui-column-widths') || '') : '';
    st.__sourceCellText = tbl ? (tbl.getAttribute('data-sui-cell-text') || '') : '';
    if (!st.__sourceWidths) st.__sourceWidths = captureColumnWidths(tbl);
  }
  function restoreTableSourceAttrs(tbl, st){
    if (!tbl || !st) return;
    if (st.__sourceWidthsAttr) tbl.setAttribute('data-sui-column-widths', st.__sourceWidthsAttr);
    else tbl.removeAttribute('data-sui-column-widths');
    if (st.__sourceCellText) {
      tbl.setAttribute('data-sui-cell-text', st.__sourceCellText);
      if (st.__sourceCellText === 'truncate') tbl.setAttribute('data-sui-fixed', '1');
      else tbl.removeAttribute('data-sui-fixed');
    } else {
      tbl.removeAttribute('data-sui-cell-text');
      tbl.removeAttribute('data-sui-fixed');
    }
  }
  // During the initial boot a script-driven table.display/importText must not
  // overwrite saved persistent data: the user's saved rows win on reload.
  function persistentBootBlocks(name){
    var tbl = firstTable(name);
    if (!tbl) return false;
    var st = store[name];
    if (!st || !st.__persistLoaded) return false;
    if (cfgOf(tbl).editMode !== 'persistent') return false;
    return !tableUserTouched && (Date.now() - tableBootAt) < 2000;
  }
  function recordChange(name, rec){
    rec.time = Date.now();
    changes[name] = rec;
    if (rec && /^(cell|display|import|clear|row-|column-|sort|filter|page|resize)/.test(String(rec.op || ''))) persistTable(name);
  }
  function clearPersisted(name){
    if (!name) return [0, 'table.clearPersisted: name is required'];
    var removed = 0;
    try {
      if (window.localStorage && window.localStorage.getItem(tablePersistKey(name)) !== null) {
        window.localStorage.removeItem(tablePersistKey(name));
        removed = 1;
      }
    } catch (e) {}
    var st = store[name];
    var tbl = firstTable(name);
    if (st) {
      if (st.__sourceRows) {
        st.header = (st.__sourceHeader || []).slice();
        st.allRows = (st.__sourceRows || []).map(function(r){ return r.slice(); });
        st.rows = st.allRows;
        st.textHeader = false;
        st.persistWidths = (st.__sourceWidths || []).slice();
        st.persistRowHeights = [];
        st.persistCellText = st.__sourceCellText || '';
        st.filter = '';
        st.colFilters = {};
        st.colValueFilters = {};
        st.colFilter = null;
        setSortKeys(st, []);
        st.page = 0;
        clearSelection(st);
        restoreTableSourceAttrs(tbl, st);
        renderName(name);
      }
      delete st.__persistLoaded;
      delete st.persistWidths;
      delete st.persistRowHeights;
      delete st.persistCellText;
    }
    return [removed, null];
  }
  // Workbooks in persistent mode also keep the runtime-created sheet list.
  function tableBookPersistKey(box){
    var path = (window.location && window.location.pathname) || '';
    return 'sui.table.book.' + path + '.' + tableBookKey(box);
  }
  function workbookHasPersistentSheet(box){
    var panels = box ? box.querySelectorAll('.sui-tabpanels .sui-tabpanel') : [];
    for (var i = 0; i < panels.length; i++) {
      var t = panels[i].querySelector('table');
      if (t && cfgOf(t).editMode === 'persistent') return true;
    }
    return false;
  }
  function workbookPersistentMode(box){
    return workbookHasPersistentSheet(box) ? 'persistent' : '';
  }
  function saveTableBookDefs(box){
    if (!box || !window.localStorage) return;
    try {
      var key = tableBookPersistKey(box);
      if (workbookHasPersistentSheet(box)) {
        window.localStorage.setItem(key, JSON.stringify(tableBooks[tableBookKey(box)] || []));
      } else if (window.localStorage.getItem(key)) {
        window.localStorage.removeItem(key);
      }
    } catch (e) {}
  }
  function loadTableBookDefs(box){
    if (!box || !workbookHasPersistentSheet(box)) return [];
    try {
      var raw = window.localStorage && window.localStorage.getItem(tableBookPersistKey(box));
      var defs = raw ? JSON.parse(raw) : [];
      return Array.isArray(defs) ? defs : [];
    } catch (e) { return []; }
  }
  function tableStore(name){
    if (!store[name]) store[name] = { header: [], allRows: [], rows: [], page: 0,
      sortCol: null, sortDir: '', sortKeys: [], filter: '', colFilter: null,
      colFilters: {}, colValueFilters: {}, headerEdits: {}, filtersOn: null, sortOn: null,
      selectedRows: {}, selectedCells: {}, selectedCols: {} };
    return store[name];
  }
  function sortKeysOf(st){
    if (st.sortKeys && st.sortKeys.length) return st.sortKeys;
    if (st.sortCol !== null && st.sortCol !== undefined && st.sortDir) return [{ col: st.sortCol, dir: st.sortDir }];
    return [];
  }
  function setSortKeys(st, keys){
    st.sortKeys = keys || [];
    if (st.sortKeys.length) { st.sortCol = st.sortKeys[0].col; st.sortDir = st.sortKeys[0].dir; }
    else { st.sortCol = null; st.sortDir = ''; }
  }
  function colFiltersOf(st){ return st.colFilters || (st.colFilters = {}); }
  function colIndex(cfg, key){
    var s = String(key === undefined || key === null ? '' : key).trim();
    if (s === '') return -1;
    if (/^\d+$/.test(s)) return parseInt(s, 10);
    for (var i = 0; i < cfg.columns.length; i++) if (cfg.columns[i] === s) return i;
    for (var j = 0; j < cfg.labels.length; j++) if (cfg.labels[j] === s) return j;
    return -1;
  }
  function colOpts(cfg, col){
    if (!cfg.options) return null;
    return cfg.options[col] || cfg.options[String(col)] || null;
  }
  function validValue(v, type, opts){
    var s = String(v === undefined || v === null ? '' : v).trim();
    if (s === '') return '';
    type = String(type || 'text').toLowerCase();
    if (type === 'number') return isFinite(Number(s.replace(',', '.'))) ? '' : 'number';
    if (type === 'date') return isNaN(Date.parse(s)) ? 'date' : '';
    if (type === 'boolean') {
      var low = s.toLowerCase();
      return (low === 'true' || low === 'false' || low === 'yes' || low === 'no' || low === '1' || low === '0') ? '' : 'boolean';
    }
    if (type === 'select' && opts && opts.length) return opts.indexOf(s) >= 0 ? '' : 'select';
    return '';
  }
  function formatValue(v, fmt){
    var s = String(v === undefined || v === null ? '' : v);
    if (!fmt || fmt === 'text' || s === '') return s;
    var n = Number(s.replace(',', '.'));
    if (fmt === 'number') return isFinite(n) ? n.toLocaleString() : s;
    if (fmt === 'percent') return isFinite(n) ? n.toLocaleString(undefined, { style: 'percent', maximumFractionDigits: 2 }) : s;
    if (fmt.indexOf('currency') === 0) {
      var cur = fmt.split(':')[1] || 'USD';
      return isFinite(n) ? n.toLocaleString(undefined, { style: 'currency', currency: cur }) : s;
    }
    if (fmt === 'date') {
      var d = new Date(s);
      return isNaN(d.getTime()) ? s : d.toLocaleDateString();
    }
    return s;
  }
  function clearSelection(st){
    st.selectedRows = {}; st.selectedCells = {}; st.selectedCols = {};
  }
  function rowText(row){ return (row || []).join(' ').toLowerCase(); }
  function applyView(name){
    var st = tableStore(name);
    var tbl = firstTable(name);
    var cfg = tbl ? cfgOf(tbl) : { pageSize: 0, columns: [], labels: [] };
    st.filtersOn = (st.filtersOn === null || st.filtersOn === undefined) ? !!cfg.filterable : !!st.filtersOn;
    st.sortOn = (st.sortOn === null || st.sortOn === undefined) ? !!cfg.sortable : !!st.sortOn;
    var rows = st.allRows.slice();
    if (cfg.columns.length && !st.header.length) st.header = cfg.labels.length ? cfg.labels.slice() : cfg.columns.slice();
    if (st.filter) {
      var term = String(st.filter).toLowerCase();
      rows = rows.filter(function(r){ return rowText(r).indexOf(term) >= 0; });
    }
    if (st.colFilter) {
      var ci = colIndex(cfg, st.colFilter.column);
      var cv = String(st.colFilter.value || '').toLowerCase();
      if (ci >= 0 && cv) rows = rows.filter(function(r){ return String(r[ci] === undefined ? '' : r[ci]).toLowerCase().indexOf(cv) >= 0; });
    }
    // Per-column filters (the Excel-like header filter row).
    var cf = colFiltersOf(st);
    for (var cfk in cf) {
      if (!Object.prototype.hasOwnProperty.call(cf, cfk)) continue;
      var cfIdx = parseInt(cfk, 10);
      var cfv = String(cf[cfk] === undefined || cf[cfk] === null ? '' : cf[cfk]).toLowerCase();
      if (cfIdx >= 0 && cfv) {
        rows = rows.filter(function(r){ return String(r[cfIdx] === undefined ? '' : r[cfIdx]).toLowerCase().indexOf(cfv) >= 0; });
      }
    }
    // Excel-like value filters from the header filter popup: exact membership.
    var cvf = st.colValueFilters || {};
    for (var cvk in cvf) {
      if (!Object.prototype.hasOwnProperty.call(cvf, cvk)) continue;
      var cvIdx = parseInt(cvk, 10);
      var selected = cvf[cvk] || [];
      if (cvIdx >= 0) {
        rows = rows.filter(function(r){
          return selected.indexOf(String(r[cvIdx] === undefined ? '' : r[cvIdx])) >= 0;
        });
      }
    }
    var keys = sortKeysOf(st);
    if (keys.length) {
      rows.sort(function(a, b){
        for (var k = 0; k < keys.length; k++) {
          var key = keys[k];
          if (!key || !key.dir) continue;
          var sd = key.dir === 'desc' ? -1 : 1;
          var av = a[key.col] === undefined ? '' : a[key.col], bv = b[key.col] === undefined ? '' : b[key.col];
          var an = Number(av), bn = Number(bv);
          var cmp = 0;
          if (isFinite(an) && isFinite(bn) && String(av).trim() !== '' && String(bv).trim() !== '') cmp = an - bn;
          else {
            var as = String(av).toLowerCase(), bs = String(bv).toLowerCase();
            cmp = as < bs ? -1 : as > bs ? 1 : 0;
          }
          if (cmp) return cmp * sd;
        }
        return 0;
      });
    }
    var pageSize = (st.pageSize !== undefined && st.pageSize !== null) ? st.pageSize : cfg.pageSize;
    pageSize = pageSize > 0 ? pageSize : 0;
    var pageCount = pageSize > 0 ? Math.max(1, Math.ceil(rows.length / pageSize)) : 1;
    if (st.page >= pageCount) st.page = pageCount - 1;
    if (st.page < 0) st.page = 0;
    st.pageCount = pageCount;
    st.rows = pageSize > 0 ? rows.slice(st.page * pageSize, st.page * pageSize + pageSize) : rows;
    return { st: st, cfg: cfg };
  }

  function isObj(v){ return v !== null && typeof v === 'object' && !Array.isArray(v); }
  function cellStr(v){
    if (v === null || v === undefined) return '';
    if (typeof v === 'object') { try { return JSON.stringify(v); } catch(e){ return String(v); } }
    return String(v);
  }
  // normValue normalizes already-parsed data: arrays of arrays, arrays of
  // objects, {columns, rows} payloads, key/value objects and scalars.
  function normValue(data){
    if (data === null || data === undefined) return { header: null, rows: [] };
    if (Array.isArray(data)) {
      if (!data.length) return { header: null, rows: [] };
      if (data.length > MAX_ROWS) return { error: 'table.display: too many rows (limit ' + MAX_ROWS + ')' };
      var rows = [], header = null, i, j;
      if (Array.isArray(data[0])) {
        for (i = 0; i < data.length; i++) {
          var src = Array.isArray(data[i]) ? data[i] : [data[i]], rr = [];
          if (src.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
          for (j = 0; j < src.length; j++) rr.push(cellStr(src[j]));
          rows.push(rr);
        }
      } else if (isObj(data[0])) {
        header = [];
        for (i = 0; i < data.length; i++) {
          if (!isObj(data[i])) continue;
          for (var k in data[i]) {
            if (Object.prototype.hasOwnProperty.call(data[i], k) && header.indexOf(k) < 0) header.push(k);
          }
          if (header.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
        }
        for (i = 0; i < data.length; i++) {
          var o = isObj(data[i]) ? data[i] : {}, r2 = [];
          for (j = 0; j < header.length; j++) r2.push(cellStr(o[header[j]]));
          rows.push(r2);
        }
      } else {
        for (i = 0; i < data.length; i++) rows.push([cellStr(data[i])]);
      }
      return { header: header, rows: rows };
    }
    if (isObj(data)) {
      if (Array.isArray(data.rows)) {
        var sub = normValue(data.rows);
        if (sub.error) return sub;
        var h = null;
        if (Array.isArray(data.columns)) h = data.columns.map(cellStr);
        else if (Array.isArray(data.headers)) h = data.headers.map(cellStr);
        else h = sub.header;
        if (h && h.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
        return { header: h, rows: sub.rows };
      }
      var ks = Object.keys(data), kv = [];
      if (ks.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
      for (var x = 0; x < ks.length; x++) kv.push([ks[x], cellStr(data[ks[x]])]);
      return { header: ['key', 'value'], rows: kv };
    }
    return { header: null, rows: [[cellStr(data)]] };
  }

  // ---- delimited text (CSV, TSV, semicolon/pipe separated) -----------------
  function countSep(line, sep){
    var n = 0, q = false, i, ch;
    for (i = 0; i < line.length; i++) {
      ch = line.charAt(i);
      if (ch === '"') q = !q;
      else if (!q && ch === sep) n++;
    }
    return n;
  }
  function detectSep(text){
    var line = String(text).split(/\r?\n/)[0] || '';
    var cands = [',', ';', '\t', '|'], best = '', bestN = 0, i, n;
    for (i = 0; i < cands.length; i++) {
      n = countSep(line, cands[i]);
      if (n > bestN) { bestN = n; best = cands[i]; }
    }
    return best || ',';
  }
  function delimFor(fmt){
    if (fmt === 'tsv') return '\t';
    if (fmt === 'ssv') return ';';
    if (fmt === 'psv') return '|';
    return ',';
  }
  // RFC4180-style parser: quoted fields, "" escapes, embedded delimiters and
  // newlines inside quotes, CRLF, UTF-8 BOM.
  function parseDelimited(text, sep, quote, trim){
    var rows = [], row = [], field = '', inQ = false, i, ch;
    text = String(text).replace(/^\uFEFF/, '');
    for (i = 0; i < text.length; i++) {
      ch = text.charAt(i);
      if (inQ) {
        if (ch === quote) {
          if (text.charAt(i + 1) === quote) { field += quote; i++; }
          else inQ = false;
        } else field += ch;
      } else if (quote && ch === quote && field === '') {
        inQ = true;
      } else if (ch === sep) {
        row.push(trim ? field.trim() : field); field = '';
        if (row.length > MAX_COLS) return null;
      } else if (ch === '\n') {
        row.push(trim ? field.trim() : field); rows.push(row); row = []; field = '';
        if (rows.length > MAX_ROWS) return null;
      } else if (ch === '\r') {
        // CRLF: the following \n ends the row
      } else {
        field += ch;
      }
    }
    if (field !== '' || row.length) {
      row.push(trim ? field.trim() : field);
      if (row.length > MAX_COLS) return null;
      rows.push(row);
      if (rows.length > MAX_ROWS) return null;
    }
    if (rows.length && rows[rows.length - 1].length === 1 && rows[rows.length - 1][0] === '') rows.pop();
    return rows;
  }
  function normDelimited(text, header, sep, quote, trim){
    var s = sep || detectSep(text);
    var rows = parseDelimited(text, s, quote === undefined ? '"' : quote, !!trim);
    if (!rows) return { error: 'table.importText: too many rows or columns (limit ' + MAX_ROWS + 'x' + MAX_COLS + ')' };
    if (!rows.length) return { header: null, rows: [] };
    var useHeader = header === null || header === undefined ? rows.length > 1 : !!header;
    var h = useHeader ? rows.shift() : null;
    return { header: h, rows: rows, textHeader: useHeader, sep: s };
  }
  function tryJsonLines(text){
    var lines = String(text).split(/\r?\n/), vals = [], i, line;
    for (i = 0; i < lines.length; i++) {
      line = lines[i].trim();
      if (!line) continue;
      try { vals.push(JSON.parse(line)); } catch(e) { return null; }
      if (vals.length > MAX_ROWS) return { error: 'table.display: too many rows (limit ' + MAX_ROWS + ')' };
    }
    if (vals.length < 2) return null;
    return normValue(vals);
  }
  // startsJSON checks the first character without embedding raw delimiters in
  // the runtime (keeps the generated source's bracket count balanced).
  function startsJSON(s){
    var c = String(s).charCodeAt(0);
    return c === 91 || c === 123; // opening bracket or brace
  }
  // spec syntax: "csv", "tsv", "ssv" (;), "psv" (|), "json", "jsonl", "auto",
  // plus ;header=0|1 ;sep=comma|semicolon|tab|pipe|<char> ;quote=<char> ;trim=1
  function parseSpec(spec, defFmt){
    var s = String(spec === undefined || spec === null ? '' : spec).trim().toLowerCase();
    var bits = s ? s.split(';') : [];
    var fmt = bits.length ? bits.shift().trim() : (defFmt || 'auto');
    var o = { format: fmt || 'auto', header: null, sep: '', quote: '"', trim: false,
      formula: (typeof __SUI_SEC !== 'undefined' && __SUI_SEC.csvFormulaEscape) ? 'escape' : 'raw' };
    for (var i = 0; i < bits.length; i++) {
      var b = bits[i].trim();
      if (!b) continue;
      var eq = b.indexOf('=');
      var k = (eq < 0 ? b : b.slice(0, eq)).trim();
      var v = eq < 0 ? '' : b.slice(eq + 1).trim();
      if (k === 'header') o.header = !(v === '0' || v === 'false' || v === 'no' || v === 'off');
      else if (k === 'sep') {
        if (v === 'tab' || v === '\\t') o.sep = '\t';
        else if (v === 'comma') o.sep = ',';
        else if (v === 'semicolon') o.sep = ';';
        else if (v === 'pipe') o.sep = '|';
        else o.sep = v;
      }
      else if (k === 'quote') o.quote = v;
      else if (k === 'trim') o.trim = (v === '1' || v === 'true' || v === 'yes' || v === 'on');
      else if (k === 'formula' || k === 'safe') {
        o.formula = (v === 'escape' || v === 'safe' || v === '1' || v === 'true' || v === 'yes') ? 'escape' : 'raw';
      }
    }
    return o;
  }
  // normText parses an explicit text payload according to spec.
  function normText(text, spec){
    var o = parseSpec(spec, 'auto');
    var s = String(text === undefined || text === null ? '' : text).trim();
    if (!s) return { header: null, rows: [] };
    if (s.length > MAX_TEXT) return { error: 'table.importText: text too large (limit ' + MAX_TEXT + ' characters)' };
    if (o.format === 'json') {
      try { return normValue(JSON.parse(s)); }
      catch(e) { return { error: 'table.importText: invalid JSON (' + e.message + ')' }; }
    }
    if (o.format === 'jsonl' || o.format === 'ndjson') {
      var jl = tryJsonLines(s);
      return jl || { error: 'table.importText: invalid JSON Lines' };
    }
    if (o.format === 'auto') {
      if (startsJSON(s)) {
        var jerr = null;
        try { return normValue(JSON.parse(s)); } catch(e) { jerr = e; }
        var auto = tryJsonLines(s);
        if (auto) return auto;
        return { error: 'table.importText: invalid JSON (' + (jerr ? jerr.message : 'parse error') + ')' };
      }
    }
    var sep = o.sep || (o.format === 'auto' ? detectSep(s) : delimFor(o.format));
    var useHeader = o.header === null ? true : !!o.header;
    return normDelimited(s, useHeader, sep, o.quote, o.trim);
  }
  // norm is the table.display() entry point: it sniffs JSON, JSON Lines and
  // delimited text, and treats the first row of a multi-row delimited payload
  // as the header.
  function norm(data){
    if (typeof data !== 'string') return normValue(data);
    var s = data.trim();
    if (!s) return { header: null, rows: [] };
    if (s.length > MAX_TEXT) return { error: 'table.display: text too large (limit ' + MAX_TEXT + ' characters)' };
    if (startsJSON(s)) {
      var err = null;
      try { return normValue(JSON.parse(s)); } catch(e) { err = e; }
      var jl = tryJsonLines(s);
      if (jl) return jl;
      return { error: 'table.display: invalid JSON (' + (err ? err.message : 'parse error') + ')' };
    }
    return normDelimited(s, null, detectSep(s), '"', false);
  }

  function findTables(name){
    var out = [], ts = document.querySelectorAll('table');
    for (var i = 0; i < ts.length; i++) { if (ts[i].getAttribute('name') === name) out.push(ts[i]); }
    return out;
  }
  function cellsOf(tr){
    var out = [], ch = tr.children;
    for (var i = 0; i < ch.length; i++) { if (ch[i].tagName === 'TD' || ch[i].tagName === 'TH') out.push(ch[i]); }
    return out;
  }
  function tbodyOf(tbl){
    if (tbl.tBodies && tbl.tBodies.length) return tbl.tBodies[0];
    var tb = tbl.querySelector('tbody');
    if (!tb) { tb = document.createElement('tbody'); tbl.appendChild(tb); }
    return tb;
  }
  function bodyRows(tbl){
    var tb = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
    if (tb) return Array.prototype.slice.call(tb.rows);
    var out = [], trs = tbl.querySelectorAll('tr');
    for (var i = 0; i < trs.length; i++) { if (!trs[i].parentNode || trs[i].parentNode.tagName !== 'THEAD') out.push(trs[i]); }
    return out;
  }
  function templates(tbl){
    var t = tpls.get(tbl);
    if (t) return t;
    var bodies = [], heads = [], tb = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
    if (tb) for (var i = 0; i < tb.rows.length; i++) bodies.push(tb.rows[i].cloneNode(true));
    var hr = tbl.querySelector('thead tr');
    if (hr) heads.push(hr.cloneNode(true));
    if (!bodies.length) bodies.push(hr ? hr.cloneNode(true) : document.createElement('tr'));
    // Keep the designed ids so a runtime re-render (sort/filter/settings) does
    // not break the designer's preview picking.
    var rowIds = [], headCells = [];
    if (hr) {
      var hcs = cellsOf(hr);
      for (var hc = 0; hc < hcs.length; hc++) headCells.push(hcs[hc].getAttribute(TBL_ID_ATTR));
    }
    for (var b = 0; b < bodies.length; b++) {
      var bcs = cellsOf(bodies[b]), ids = [];
      for (var bi = 0; bi < bcs.length; bi++) ids.push(bcs[bi].getAttribute(TBL_ID_ATTR));
      rowIds.push({ row: bodies[b].getAttribute(TBL_ID_ATTR), cells: ids });
    }
    t = { heads: heads, bodies: bodies, headCells: headCells, rowIds: rowIds };
    tpls.set(tbl, t);
    return t;
  }
  function toTd(cell){
    if (cell.tagName !== 'TH') return cell;
    var td = document.createElement('td');
    if (cell.className) td.className = cell.className;
    var st = cell.getAttribute('style');
    if (st) td.setAttribute('style', st);
    return td;
  }
  function prepRow(tr, n, want){
    var cells = cellsOf(tr);
    while (cells.length > n) { tr.removeChild(cells.pop()); }
    while (cells.length < n) {
      var c = cells.length ? cells[cells.length - 1].cloneNode(false) : document.createElement(want === 'td' ? 'td' : 'th');
      tr.appendChild(c); cells.push(c);
    }
    for (var i = 0; i < cells.length; i++) {
      cells[i].removeAttribute('colspan');
      cells[i].removeAttribute('rowspan');
      cells[i].removeAttribute('data-' + 'sui-id');
      cells[i].removeAttribute('contenteditable');
      if (want === 'td') {
        var td = toTd(cells[i]);
        if (td !== cells[i]) tr.replaceChild(td, cells[i]);
        cells[i] = td;
      }
    }
    return cells;
  }
  function applyTo(tbl, ds){
    if (tableFilterPop && tableFilterPop.name === tableName(tbl)) closeTableFilterPop();
    var t = templates(tbl);
    var n = 0, i, j;
    var hr = tbl.querySelector('thead tr');
    if (!hr && ds.header && ds.header.length) {
      var thead = tbl.querySelector('thead');
      if (!thead) { thead = document.createElement('thead'); tbl.insertBefore(thead, tbl.firstChild); }
      hr = document.createElement('tr');
      thead.appendChild(hr);
    }
    var rows = ds.rows;
    // Delimited text always carries its first row as the header. When the
    // designed table has no <thead>, keep that row visible as the first body
    // row instead of dropping it.
    if (ds.textHeader && ds.header && !hr) rows = [ds.header].concat(rows);
    if (ds.header) n = ds.header.length;
    for (i = 0; i < rows.length; i++) n = Math.max(n, rows[i].length);
    if (!ds.header) {
      // No header information in the data: keep the designed column count and
      // labels (arrays/JSON arrays carry no column names).
      if (hr) n = Math.max(n, cellsOf(hr).length);
      if (t.bodies.length && t.bodies[0].children) n = Math.max(n, cellsOf(t.bodies[0]).length);
    }
    if (!n) {
      var tb0 = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
      if (tb0) while (tb0.firstChild) tb0.removeChild(tb0.firstChild);
      return;
    }
    if (hr) {
      var hc = prepRow(hr, n, 'th');
      if (ds.header) {
        for (j = 0; j < hc.length; j++) hc[j].textContent = j < ds.header.length ? ds.header[j] : '';
      } else if (t.heads.length) {
        // restore the labels as designed (a previous object payload may have
        // replaced them with key names)
        var orig = cellsOf(t.heads[0]);
        for (j = 0; j < hc.length; j++) hc[j].textContent = j < orig.length ? orig[j].textContent : '';
      }
      for (j = 0; j < hc.length; j++) {
        if (t.headCells && t.headCells[j]) hc[j].setAttribute(TBL_ID_ATTR, t.headCells[j]);
      }
    }
    var tb = tbodyOf(tbl);
    while (tb.firstChild) tb.removeChild(tb.firstChild);
    for (i = 0; i < rows.length; i++) {
      var origIdx = i;
      if (ds.allRows) {
        var found = ds.allRows.indexOf(rows[i]);
        if (found >= 0) origIdx = found;
      }
      var tr = t.bodies[i % t.bodies.length].cloneNode(true);
      tr.removeAttribute(TBL_ID_ATTR);
      var cs = prepRow(tr, n, 'td');
      for (j = 0; j < cs.length; j++) cs[j].textContent = rows[i][j] !== undefined ? rows[i][j] : '';
      var ids = t.rowIds && t.rowIds[origIdx];
      if (ids) {
        if (ids.row) tr.setAttribute(TBL_ID_ATTR, ids.row);
        for (j = 0; j < cs.length; j++) {
          if (ids.cells[j]) cs[j].setAttribute(TBL_ID_ATTR, ids.cells[j]);
        }
      }
      tb.appendChild(tr);
    }
  }
  function selectionChanged(name, rec){
    recordChange(name, rec);
    renderName(name);
    fireTableEvent(name, 'onrowselect');
  }
  function selectRow(name, row, multi){
    var st = tableStore(name);
    if (!multi) clearSelection(st);
    if (st.selectedRows[row]) delete st.selectedRows[row]; else st.selectedRows[row] = true;
    selectionChanged(name, { op: 'select-row', row: row });
  }
  function selectCell(name, row, col, multi){
    var st = tableStore(name);
    if (!multi) clearSelection(st);
    var key = row + ':' + col;
    if (st.selectedCells[key]) delete st.selectedCells[key]; else st.selectedCells[key] = true;
    selectionChanged(name, { op: 'select-cell', row: row, column: col });
  }
  function selectColumn(name, col, multi){
    var st = tableStore(name);
    if (!multi) clearSelection(st);
    if (st.selectedCols[col]) delete st.selectedCols[col]; else st.selectedCols[col] = true;
    selectionChanged(name, { op: 'select-column', column: col });
  }
  function structRowList(st, row){
    var keys = [];
    for (var k in st.selectedRows) if (st.selectedRows[k]) keys.push(parseInt(k, 10));
    keys.sort(function(a,b){ return a-b; });
    if (!keys.length && row >= 0) keys = [row];
    return keys;
  }
  function structColList(st, col){
    var keys = [];
    for (var k in st.selectedCols) if (st.selectedCols[k]) keys.push(parseInt(k, 10));
    keys.sort(function(a,b){ return a-b; });
    if (!keys.length && col >= 0) keys = [col];
    return keys;
  }
  function structApply(name, st, rec){
    st.page = 0;
    clearSelection(st);
    recordChange(name, rec);
    renderName(name);
    fireTableEvent(name, 'ontablechange');
  }
  function structInsertRows(name, at, rows){
    var st = tableStore(name);
    var w = st.header.length;
    for (var i = 0; i < rows.length; i++) {
      var r = rows[i].slice();
      while (r.length < w) r.push('');
      st.allRows.splice(at + i, 0, r);
    }
    structApply(name, st, { op: 'row-insert', row: at, count: rows.length });
  }
  function structDeleteRows(name, indexes){
    var st = tableStore(name);
    indexes = indexes.slice().sort(function(a,b){ return b-a; });
    for (var i = 0; i < indexes.length; i++) st.allRows.splice(indexes[i], 1);
    structApply(name, st, { op: 'row-delete', row: indexes.length ? indexes[indexes.length-1] : 0, count: indexes.length });
  }
  function structInsertColumns(name, at, labels, values){
    var st = tableStore(name);
    for (var i = 0; i < labels.length; i++) st.header.splice(at + i, 0, labels[i]);
    for (var r = 0; r < st.allRows.length; r++) {
      for (var c = 0; c < labels.length; c++) st.allRows[r].splice(at + c, 0, values[r] && values[r][c] !== undefined ? values[r][c] : '');
    }
    st.__colStruct = true;
    structApply(name, st, { op: 'column-insert', column: at, count: labels.length });
  }
  function structDeleteColumns(name, indexes){
    var st = tableStore(name);
    indexes = indexes.slice().sort(function(a,b){ return b-a; });
    for (var i = 0; i < indexes.length; i++) {
      var col = indexes[i];
      if (st.header.length) st.header.splice(col, 1);
      for (var r = 0; r < st.allRows.length; r++) st.allRows[r].splice(col, 1);
    }
    st.__colStruct = true;
    structApply(name, st, { op: 'column-delete', column: indexes.length ? indexes[indexes.length-1] : 0, count: indexes.length });
  }
  function openTableStructMenu(ev, tbl, name, row, col, cfg){
    var old = document.getElementById('sui-table-struct');
    if (old) old.remove();
    var st = tableStore(name);
    var menu = document.createElement('div');
    menu.id = 'sui-table-struct';
    menu.className = 'sui-table-struct-menu';
    menu.setAttribute('role', 'menu');
    function add(label, fn, disabled) {
      var b = document.createElement('button');
      b.type = 'button'; b.className = 'sui-table-struct-item'; b.textContent = label;
      b.disabled = !!disabled;
      b.addEventListener('click', function(e){ e.stopPropagation(); menu.remove(); fn(); });
      menu.appendChild(b);
    }
    function sep(){ var d = document.createElement('div'); d.className = 'sui-table-struct-sep'; menu.appendChild(d); }
    var rows = structRowList(st, row), cols = structColList(st, col);
    if (row >= 0) {
      add('Insert row above', function(){ structInsertRows(name, Math.min.apply(null, rows), [[]]); });
      add('Insert row below', function(){ structInsertRows(name, Math.max.apply(null, rows) + 1, [[]]); });
      add('Delete row' + (rows.length > 1 ? 's' : ''), function(){ structDeleteRows(name, rows); });
      add('Copy row' + (rows.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'rows', rows: rows.map(function(i){ return (st.allRows[i] || []).slice(); }) };
      });
      add('Cut row' + (rows.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'rows', rows: rows.map(function(i){ return (st.allRows[i] || []).slice(); }) };
        structDeleteRows(name, rows);
      });
      if (tableClip[name] && tableClip[name].kind === 'rows') add('Paste rows', function(){ structInsertRows(name, row, tableClip[name].rows); });
      sep();
    }
    if (col >= 0) {
      add('Insert column left', function(){ structInsertColumns(name, Math.min.apply(null, cols), ['Column'], []); });
      add('Insert column right', function(){ structInsertColumns(name, Math.max.apply(null, cols) + 1, ['Column'], []); });
      add('Delete column' + (cols.length > 1 ? 's' : ''), function(){ structDeleteColumns(name, cols); });
      add('Copy column' + (cols.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'cols', labels: cols.map(function(c){ return st.header[c] || ''; }),
          values: st.allRows.map(function(r){ return cols.map(function(c){ return r[c] === undefined ? '' : r[c]; }); }) };
      });
      add('Cut column' + (cols.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'cols', labels: cols.map(function(c){ return st.header[c] || ''; }),
          values: st.allRows.map(function(r){ return cols.map(function(c){ return r[c] === undefined ? '' : r[c]; }); }) };
        structDeleteColumns(name, cols);
      });
      if (tableClip[name] && tableClip[name].kind === 'cols') add('Paste columns', function(){ structInsertColumns(name, col, tableClip[name].labels, tableClip[name].values); });
      sep();
    }
    add('Add row at end', function(){ structInsertRows(name, st.allRows.length, [[]]); });
    add('Add column at end', function(){ structInsertColumns(name, st.header.length || (st.allRows[0] ? st.allRows[0].length : 0), ['Column'], []); });
    var x = ev.clientX, y = ev.clientY;
    document.body.appendChild(menu);
    var rect = menu.getBoundingClientRect();
    if (x + rect.width > window.innerWidth - 8) x = Math.max(8, window.innerWidth - rect.width - 8);
    if (y + rect.height > window.innerHeight - 8) y = Math.max(8, window.innerHeight - rect.height - 8);
    menu.style.left = x + 'px'; menu.style.top = y + 'px';
    var close = function(){ menu.remove(); document.removeEventListener('click', close); };
    setTimeout(function(){ document.addEventListener('click', close); }, 0);
  }
  function tableBookKey(box){
    return box.getAttribute('data-sui-tabs-key') || ('tabs:' + (box.getAttribute('aria-label') || '') + ':' + Array.prototype.indexOf.call(document.querySelectorAll('.sui-tabs'), box));
  }
  function addRuntimeTableDom(box, name, label, templatePanel){
    var list = box.querySelector('.sui-tablist');
    var panels = box.querySelector('.sui-tabpanels');
    if (!list || !panels || !templatePanel) return null;
    if (list.querySelector('[data-sui-table-sheet="' + name + '"]')) return null;
    var clone = templatePanel.cloneNode(true);
    clone.classList.remove('active');
    clone.hidden = true;
    clone.setAttribute('data-sui-table-sheet', name);
    Array.prototype.forEach.call(clone.querySelectorAll('[' + 'data-' + 'sui-id' + ']'), function(el){ el.removeAttribute('data-' + 'sui-id'); });
    Array.prototype.forEach.call(clone.querySelectorAll('.sui-table-filter, .sui-table-pager, .sui-table-add, .sui-table-topbar, .sui-table-tools'), function(el){ if (el.parentNode) el.parentNode.removeChild(el); });
      var newTable = clone.querySelector('table');
    if (newTable) {
      newTable.setAttribute('name', name);
      // A new sheet gets its own empty title row, not the source caption.
      var clonedCaption = newTable.querySelector('caption');
      if (clonedCaption && clonedCaption.parentNode) clonedCaption.parentNode.removeChild(clonedCaption);
      var tb = newTable.tBodies && newTable.tBodies.length ? newTable.tBodies[0] : null;
      // Keep a real body row as the template so the new sheet starts with two
      // styled empty rows (not header clones, which could look unstyled/white).
      var bodyTemplate = tb && tb.rows.length ? tb.rows[0].cloneNode(true) : null;
      if (tb) while (tb.firstChild) tb.removeChild(tb.firstChild);
      if (tb) {
        for (var rIdx = 0; rIdx < 2; rIdx++) {
          var tr = bodyTemplate ? bodyTemplate.cloneNode(true) : document.createElement('tr');
          if (!bodyTemplate) {
            var colCount = newTable.querySelectorAll('thead th').length;
            for (var cIdx = 0; cIdx < colCount; cIdx++) tr.appendChild(document.createElement('td'));
          }
          Array.prototype.forEach.call(tr.querySelectorAll('[' + TBL_ID_ATTR + ']'), function(el){ el.removeAttribute(TBL_ID_ATTR); });
          tr.removeAttribute(TBL_ID_ATTR);
          Array.prototype.forEach.call(tr.querySelectorAll('td,th'), function(el){
            el.removeAttribute(TBL_ID_ATTR);
            el.removeAttribute('data-sui-raw');
            el.textContent = '';
          });
          tb.appendChild(tr);
        }
      }
      // A new sheet keeps its own generic header instead of inheriting the
      // source table's labels, types and formats: a blank page must accept any
      // text, even when it was cloned from a read-only or numeric table.
      var headCells = newTable.querySelectorAll('thead tr:first-child th, thead tr:first-child td');
      // On a re-render the template panel may not be decorated yet, so derive
      // the column count from the declared columns when there are no th cells.
      var colCount = headCells.length;
      if (!colCount) {
        var colAttr = newTable.getAttribute('data-sui-columns') || '';
        if (colAttr) colCount = colAttr.split(',').filter(function(x){ return x.trim() !== ''; }).length;
      }
      // A runtime sheet may have gained columns before a re-render; keep the
      // stored width so the header and the data stay aligned.
      var existingStore = store[name];
      if (existingStore && existingStore.header && existingStore.header.length > colCount) colCount = existingStore.header.length;
      if (colCount) {
        var headLabels = [], headKeys = [];
        for (var hi = 0; hi < colCount; hi++) {
          var headLabel = hi < 26 ? String.fromCharCode(65 + hi) : ('Col ' + (hi + 1));
          headLabels.push(headLabel);
          headKeys.push(String.fromCharCode(97 + (hi % 26)) + (hi >= 26 ? String(hi + 1) : ''));
          if (headCells[hi]) {
            headCells[hi].removeAttribute('data-sui-raw');
            headCells[hi].textContent = headLabel;
          }
        }
        newTable.setAttribute('data-sui-column-labels', headLabels.join(','));
        newTable.setAttribute('data-sui-columns', headKeys.join(','));
      }
      newTable.removeAttribute('data-sui-column-types');
      newTable.removeAttribute('data-sui-column-formats');
      newTable.removeAttribute('data-sui-column-options');
      // A runtime sheet is for manual data entry, so it is always editable.
      newTable.setAttribute('data-sui-editable', '1');
    }
    panels.appendChild(clone);
    var tab = document.createElement('button');
    tab.type = 'button'; tab.className = 'sui-tab';
    tab.setAttribute('role', 'tab');
    tab.setAttribute('aria-selected', 'false');
    tab.setAttribute('data-sui-table-sheet', name);
    tab.tabIndex = -1;
    tab.textContent = label;
    var addTab = list.querySelector('.sui-tab-add');
    if (addTab) list.insertBefore(tab, addTab); else list.appendChild(tab);
    return { tab: tab, panel: clone, table: newTable };
  }
  function restoreTableBooks(){
    var boxes = document.querySelectorAll('.sui-tabs');
    for (var b = 0; b < boxes.length; b++) {
      var box = boxes[b], key = tableBookKey(box), defs = tableBooks[key] || [];
      if (!defs.length) {
        var savedDefs = loadTableBookDefs(box);
        if (savedDefs.length) { tableBooks[key] = savedDefs; defs = savedDefs; }
      }
      if (!defs.length) continue;
      var panels = box.querySelector('.sui-tabpanels');
      var template = panels ? panels.querySelector('.sui-tabpanel') : null;
      if (!template) continue;
      for (var i = 0; i < defs.length; i++) {
        var made = addRuntimeTableDom(box, defs[i].name, defs[i].label, template);
        var bookMode = workbookPersistentMode(box);
        if (made && made.table && bookMode) made.table.setAttribute('data-sui-edit-mode', bookMode);
        if (made && store[defs[i].name]) {
          try { applyTo(made.table, store[defs[i].name]); } catch (e) {}
        }
      }
    }
  }
  function ensureAddTableButton(tbl, name, st, cfg){
    if (!cfg.showAddTable) return;
    var box = tbl.closest ? tbl.closest('.sui-tabs') : null;
    if (!box) return;
    var list = box.querySelector('.sui-tablist');
    if (!list) return;
    var btn = list.querySelector('.sui-tab-add');
    if (!btn) {
      btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'sui-tab-add';
      btn.setAttribute('data-sui-tab-add', '1');
      btn.textContent = '+';
      btn.title = 'Add table';
      btn.setAttribute('aria-label', 'Add table');
      btn.addEventListener('click', function(ev){
        ev.preventDefault();
        ev.stopPropagation();
        var panels = box.querySelector('.sui-tabpanels');
        var panel = panels ? (panels.querySelector('.sui-tabpanel.active') || panels.querySelector('.sui-tabpanel:not([hidden])')) : null;
        if (!panel) panel = tbl.closest('.sui-tabpanel');
        if (!panel) return;
        var activeTable = panel.querySelector('table');
        var activeCfg = activeTable ? cfgOf(activeTable) : cfg;
        var n = 1, newName = 'table-' + n;
        while (findTables(newName).length) { n++; newName = 'table-' + n; }
        var label = 'Table ' + (box.querySelectorAll('.sui-tabpanel').length + 1);
        var made = addRuntimeTableDom(box, newName, label, panel);
        if (!made) return;
        if (made.table) {
          // A new sheet inherits the workbook edit mode: persistent sheets
          // keep their data after a reload, editable ones stay per session.
          made.table.setAttribute('data-sui-edit-mode', activeCfg.editMode);
          made.table.setAttribute('data-sui-editable', '1');
          made.table.setAttribute('data-sui-show-add-table', '1');
        }
        if (activeCfg.structureEditable && made.table) {
          made.table.setAttribute('data-sui-editable', '1');
          made.table.setAttribute('data-sui-structure-editable', '1');
        }
        var key = tableBookKey(box);
        // A runtime sheet name may be recycled after a deletion.
        if (deletedWorkbookSheets[key]) delete deletedWorkbookSheets[key]['n:' + newName];
        (tableBooks[key] || (tableBooks[key] = [])).push({ name: newName, label: label });
        saveTableBookDefs(box);
        made.tab.click();
        init();
      });
      list.appendChild(btn);
    } else if (list.lastElementChild !== btn) {
      list.appendChild(btn);
    }
  }
  function applyHeaderSort(name, tbl, idx, cell, addLevel){
    var st = tableStore(name);
    var cfg = cfgOf(tbl);
    if (cfg.selection === 'column') {
      // The caller already handled plain column selection clicks.
    }
    if (!st.sortOn) return;
    var keys = sortKeysOf(st).map(function(k){ return { col: k.col, dir: k.dir }; });
    var pos = -1;
    for (var q = 0; q < keys.length; q++) if (keys[q].col === idx) { pos = q; break; }
    if (addLevel) {
      if (pos >= 0) {
        if (keys[pos].dir === 'asc') keys[pos].dir = 'desc';
        else keys.splice(pos, 1);
      } else keys.push({ col: idx, dir: 'asc' });
    } else if (keys.length === 1 && pos === 0) {
      keys = keys[0].dir === 'asc' ? [{ col: idx, dir: 'desc' }] : [];
    } else {
      keys = [{ col: idx, dir: 'asc' }];
    }
    setSortKeys(st, keys);
    st.page = 0;
    clearSelection(st);
    var lbl = cell.querySelector('.sui-table-hlabel');
    recordChange(name, { op: 'sort', column: idx, dir: keys.length ? keys[keys.length - 1].dir : '',
      keys: keys.map(function(k){ return { column: k.col, dir: k.dir }; }),
      label: lbl ? lbl.textContent : cell.textContent });
    renderName(name);
    fireTableEvent(name, 'onsort');
    fireTableEvent(name, 'ontablechange');
  }
  function decorateTable(tbl, st, cfg){
    var name = tableName(tbl);
    // Cell input is internal to the table. Stop it at the table capture phase
    // so application-level onInput handlers on the table or its ancestors do
    // not re-render the view (and blur the edited cell) on every letter.
    if (tbl && !tbl.__suiInputGuard) {
      tbl.__suiInputGuard = true;
      tbl.addEventListener('input', function(ev){ ev.stopPropagation(); }, true);
    }
    if (st.filtersOn === null || st.filtersOn === undefined) st.filtersOn = !!cfg.filterable;
    if (st.sortOn === null || st.sortOn === undefined) st.sortOn = !!cfg.sortable;
    var headerRow = tbl.querySelector('thead tr');
    var hrCells = headerRow ? cellsOf(headerRow) : [];
    for (var h = 0; h < hrCells.length; h++) {
      var hc = hrCells[h];
      hc.setAttribute('data-col', String(h));
      var hwrap = hc.querySelector('.sui-table-hwrap');
      if (!hwrap) {
        hwrap = document.createElement('div');
        hwrap.className = 'sui-table-hwrap';
        while (hc.firstChild) hwrap.appendChild(hc.firstChild);
        hc.appendChild(hwrap);
      }
      var hlabel = hwrap.querySelector('.sui-table-hlabel');
      if (!hlabel) {
        hlabel = document.createElement('span');
        hlabel.className = 'sui-table-hlabel';
        var moveNodes = [];
        for (var mi = 0; mi < hwrap.childNodes.length; mi++) {
          var mn = hwrap.childNodes[mi];
          if (mn.classList && mn.classList.contains('sui-table-hbtns')) continue;
          moveNodes.push(mn);
        }
        for (var mj = 0; mj < moveNodes.length; mj++) hlabel.appendChild(moveNodes[mj]);
        hwrap.insertBefore(hlabel, hwrap.firstChild);
      }
      var editedLabel = st.headerEdits && Object.prototype.hasOwnProperty.call(st.headerEdits, h);
      var targetLabel = editedLabel ? st.headerEdits[h] : (cfg.labels[h] !== undefined && cfg.labels[h] !== '' ? cfg.labels[h] : (st.header && st.header[h] !== undefined ? st.header[h] : ""));
      if (targetLabel === undefined || targetLabel === null) targetLabel = "";
      if (!st.__colStruct && hc !== editCell && hlabel.textContent !== targetLabel) hlabel.textContent = targetLabel;
      if (targetLabel && hlabel.title !== targetLabel) hlabel.title = targetLabel;
      else if (!targetLabel && hlabel.title) hlabel.removeAttribute('title');
      if (cfg.widths[h] > 0) { hc.style.width = cfg.widths[h] + 'px'; hc.style.minWidth = cfg.widths[h] + 'px'; }
      hlabel.style.textAlign = cfg.align[h] || '';
      hc.classList.toggle('sui-table-sortable', !!st.sortOn);
      if (st.sortOn) hc.setAttribute('tabindex', '0'); else hc.removeAttribute('tabindex');
      if (!hc.__suiSortBound) {
        hc.__suiSortBound = true;
        hc.addEventListener('click', function(ev){
          if (ev.target && ev.target.closest && ev.target.closest('.sui-table-hbtns')) return;
          var cell = ev.currentTarget;
          var idx = parseInt(cell.getAttribute('data-col'), 10) || 0;
          var st2 = tableStore(name);
          var cfg2 = cfgOf(tbl);
          // Editable headers are text fields: a plain click places the caret.
          // Sorting stays on the explicit header sort button so it cannot
          // re-render the table (and blur the header) while editing.
          if (cfg2.editable) return;
          if (cfg2.selection === 'column' && !ev.altKey) {
            selectColumn(name, idx, ev.shiftKey || ev.ctrlKey || ev.metaKey || cfg2.multi);
            return;
          }
          if (!st2.sortOn) return;
          applyHeaderSort(name, tbl, idx, cell, ev.ctrlKey || ev.metaKey || ev.shiftKey);
        });
      }
      var sortKeys = sortKeysOf(st);
      var sortDir = '', sortRank = -1;
      for (var sk = 0; sk < sortKeys.length; sk++) {
        if (sortKeys[sk].col === h) { sortDir = sortKeys[sk].dir; sortRank = sk; break; }
      }
      hc.setAttribute('aria-sort', sortDir === 'asc' ? 'ascending' : sortDir === 'desc' ? 'descending' : 'none');
      hc.classList.toggle('sui-table-sort-asc', sortDir === 'asc');
      hc.classList.toggle('sui-table-sort-desc', sortDir === 'desc');
      if (sortRank >= 0 && sortKeys.length > 1) hc.setAttribute('data-sort-rank', String(sortRank + 1));
      else hc.removeAttribute('data-sort-rank');
      ensureHeaderTools(hc, hwrap, hlabel, tbl, name, h, st, cfg, sortDir);
      if (cfg.structureEditable && !hc.__suiStructBound) {
        hc.__suiStructBound = true;
        hc.addEventListener('contextmenu', function(ev){
          ev.preventDefault();
          var col = parseInt(ev.currentTarget.getAttribute('data-col'), 10) || 0;
          openTableStructMenu(ev, tbl, name, -1, col, cfg);
        });
      }
    }
    // Keep cfgOf/header()/column pickers aligned with runtime header edits.
    var hasHeaderEdits = st.headerEdits && Object.keys(st.headerEdits).length > 0;
    if (hrCells.length && (tbl.hasAttribute('data-sui-column-labels') || hasHeaderEdits)) {
      var runtimeLabels = [];
      for (var rl = 0; rl < hrCells.length; rl++) {
        var rlLabel = hrCells[rl].querySelector('.sui-table-hlabel');
        runtimeLabels.push(rlLabel ? (rlLabel.textContent || '') : (hrCells[rl].textContent || '').trim());
      }
      var runtimeLabelsStr = runtimeLabels.join(',');
      if ((tbl.getAttribute('data-sui-column-labels') || '') !== runtimeLabelsStr) {
        tbl.setAttribute('data-sui-column-labels', runtimeLabelsStr);
      }
    }
    var rows = bodyRows(tbl);
    for (var i = 0; i < rows.length; i++) {
      var tr = rows[i];
      tr.setAttribute('data-row', String(i));
      var rowSelected = !!st.selectedRows[i];
      tr.classList.toggle('sui-table-row-selected', rowSelected);
      tr.setAttribute('aria-selected', rowSelected ? 'true' : 'false');
      var cs = cellsOf(tr);
      for (var j = 0; j < cs.length; j++) {
        var c = cs[j];
        c.setAttribute('data-col', String(j));
        if (cfg.widths[j] > 0) { c.style.width = cfg.widths[j] + 'px'; c.style.minWidth = cfg.widths[j] + 'px'; }
        if (cfg.align[j]) c.style.textAlign = cfg.align[j];
        var raw = st.rows[i] && st.rows[i][j] !== undefined ? st.rows[i][j] : (c.getAttribute('data-sui-raw') || c.textContent || '');
        if (st.rows[i]) {
          c.setAttribute('data-sui-raw', raw);
          if (c !== editCell) {
            var fv = formatValue(raw, cfg.formats[j]);
            if (c.textContent !== fv) c.textContent = fv;
          }
        } else if (cfg.formats[j]) {
          c.setAttribute('data-sui-raw', raw);
          if (c !== editCell) {
            var fv2 = formatValue(raw, cfg.formats[j]);
            if (c.textContent !== fv2) c.textContent = fv2;
          }
        }
        var cellKey = i + ':' + j;
        var cellSel = !!st.selectedCells[cellKey];
        c.classList.toggle('sui-table-cell-selected', cellSel);
        c.classList.toggle('sui-table-col-selected', !!st.selectedCols[j]);
        c.setAttribute('aria-selected', (rowSelected || cellSel || !!st.selectedCols[j]) ? 'true' : 'false');
        if (!c.__suiClickBound) {
          c.__suiClickBound = true;
          c.addEventListener('click', function(ev){
            var cell = ev.currentTarget;
            var row = parseInt(cell.parentNode.getAttribute('data-row'), 10) || 0;
            var col = parseInt(cell.getAttribute('data-col'), 10) || 0;
            recordChange(name, { op: 'click-cell', row: row, column: col });
            fireTableEvent(name, 'oncellclick');
            fireTableEvent(name, 'onrowclick');
          });
        }
        if (cfg.structureEditable && !c.__suiStructBound) {
          c.__suiStructBound = true;
          c.addEventListener('contextmenu', function(ev){
            ev.preventDefault();
            var cell = ev.currentTarget;
            var row = parseInt(cell.parentNode.getAttribute('data-row'), 10) || 0;
            var col = parseInt(cell.getAttribute('data-col'), 10) || 0;
            openTableStructMenu(ev, tbl, name, row, col, cfg);
          });
        }
        if (cfg.selection !== 'none' && !cfg.editable && !c.__suiSelectBound) {
          c.__suiSelectBound = true;
          c.addEventListener('click', function(ev){
            var cell = ev.currentTarget;
            var cfg2 = cfgOf(tbl);
            var row = parseInt(cell.parentNode.getAttribute('data-row'), 10) || 0;
            var col = parseInt(cell.getAttribute('data-col'), 10) || 0;
            if (cfg2.selection === 'row') selectRow(name, row, cfg2.multi && (ev.ctrlKey || ev.metaKey || ev.shiftKey));
            else if (cfg2.selection === 'cell') selectCell(name, row, col, cfg2.multi && (ev.ctrlKey || ev.metaKey || ev.shiftKey));
            else if (cfg2.selection === 'column') selectColumn(name, col, cfg2.multi && (ev.ctrlKey || ev.metaKey || ev.shiftKey));
          });
        }
      }
    }
    if (st.persistRowHeights && st.persistRowHeights.length) {
      var heightRows = bodyRows(tbl);
      for (var hr2 = 0; hr2 < heightRows.length && hr2 < st.persistRowHeights.length; hr2++) {
        var ph = st.persistRowHeights[hr2];
        if (!(ph > 0)) continue;
        heightRows[hr2].style.height = ph + 'px';
        for (var hc2 = 0; hc2 < heightRows[hr2].children.length; hc2++) heightRows[hr2].children[hc2].style.height = ph + 'px';
      }
    }
    updateWorkbookTableWidth(tbl);
    ensureTableTools(tbl, name);
  }
  function ensureFilterInput(tbl, name, st, cfg){
    if (!cfg.filterable) return;
    var host = tbl.previousElementSibling;
    var toolsHost = (host && host.classList && host.classList.contains('sui-table-tools')) ? host : null;
    var prev = toolsHost ? toolsHost.previousElementSibling : host;
    var inp = prev && prev.classList && prev.classList.contains('sui-table-filter') &&
      prev.getAttribute('data-sui-table-filter-for') === name ? prev : null;
    if (!inp) {
      inp = document.createElement('input');
      inp.type = 'text';
      inp.className = 'sui-table-filter';
      inp.setAttribute('data-sui-table-filter-for', name);
      inp.setAttribute('placeholder', 'Filter table\u2026');
      inp.setAttribute('aria-label', 'Filter table ' + (name || ''));
      inp.value = st.filter || '';
      inp.addEventListener('input', function(){
        var st2 = tableStore(name);
        st2.filter = inp.value;
        st2.page = 0;
        clearSelection(st2);
        renderName(name);
      });
      if (tbl.parentNode) tbl.parentNode.insertBefore(inp, toolsHost || tbl);
    } else if (document.activeElement !== inp && inp.value !== (st.filter || '')) {
      inp.value = st.filter || '';
    }
  }
  // ---- per-column filters and the runtime settings panel ------------------
  var tableCfg = null;
  function cfgEl(tag, cls, text){
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined) e.textContent = text;
    return e;
  }
  function columnChoices(tbl, cfg){
    var out = [], hr = tbl.querySelector('thead tr'), cells = hr ? cellsOf(hr) : [];
    var n = Math.max(cells.length, cfg.columns.length, cfg.labels.length);
    for (var i = 0; i < n; i++) {
      var lbl = '';
      if (cells[i]) {
        var span = cells[i].querySelector('.sui-table-hlabel');
        lbl = span ? span.textContent : (cells[i].getAttribute('data-sui-raw') || cells[i].textContent);
      }
      if (!lbl) lbl = cfg.labels[i] || cfg.columns[i] || ('Column ' + (i + 1));
      out.push({ index: i, label: String(lbl).trim() || ('Column ' + (i + 1)) });
    }
    return out;
  }
  function hasValueFilter(st, col){
    return Object.prototype.hasOwnProperty.call(st.colValueFilters || {}, String(col));
  }
  function columnValuePairs(tbl, name, col, cfg){
    var st = tableStore(name), seen = {}, out = [];
    function add(v){
      var raw = String(v === undefined || v === null ? '' : v);
      if (Object.prototype.hasOwnProperty.call(seen, raw)) return;
      seen[raw] = 1;
      var label = raw === '' ? '(Blanks)' : formatValue(raw, cfg.formats[col]);
      out.push({ raw: raw, label: label === '' ? '(Blanks)' : label });
    }
    var rows = st.allRows && st.allRows.length ? st.allRows : null;
    if (rows) {
      for (var i = 0; i < rows.length; i++) add(rows[i][col]);
    } else {
      var tb = tbl.tBodies && tbl.tBodies[0];
      if (tb) {
        for (var r = 0; r < tb.rows.length; r++) {
          var c = tb.rows[r].cells[col];
          if (!c) continue;
          var rawAttr = c.getAttribute('data-sui-raw');
          add(rawAttr !== null ? rawAttr : c.textContent);
        }
      }
    }
    out.sort(function(a, b){
      if (a.raw === '') return 1;
      if (b.raw === '') return -1;
      return a.label.localeCompare(b.label, undefined, { numeric: true, sensitivity: 'base' });
    });
    return out;
  }
  var tableFilterPop = null;
  function closeTableFilterPop(){
    if (!tableFilterPop) return;
    if (tableFilterPop.el && tableFilterPop.el.parentNode) tableFilterPop.el.parentNode.removeChild(tableFilterPop.el);
    if (tableFilterPop.btn) tableFilterPop.btn.setAttribute('aria-expanded', 'false');
    document.removeEventListener('click', tableFilterPop.outside, true);
    document.removeEventListener('keydown', tableFilterPop.esc, true);
    window.removeEventListener('resize', tableFilterPop.resize);
    window.removeEventListener('scroll', tableFilterPop.resize, true);
    tableFilterPop = null;
  }
  function openTableFilterPop(btn, tbl, name, col){
    closeTableFilterPop();
    var st = tableStore(name), cfg = cfgOf(tbl);
    if (!st.colValueFilters) st.colValueFilters = {};
    var pairs = columnValuePairs(tbl, name, col, cfg);
    var allValues = pairs.map(function(p){ return p.raw; });
    var active = hasValueFilter(st, col);
    var selected = active ? (st.colValueFilters[String(col)] || []).slice() : null;
    var el = cfgEl('div', 'sui-table-fpop');
    el.setAttribute('role', 'dialog');
    el.setAttribute('aria-label', 'Filter column');
    el.appendChild(cfgEl('div', 'sui-table-fpop-title', 'Filter'));
    var search = document.createElement('input');
    search.type = 'search';
    search.className = 'sui-table-fpop-search';
    search.setAttribute('placeholder', 'Search');
    search.setAttribute('aria-label', 'Search filter values');
    el.appendChild(search);
    var list = cfgEl('div', 'sui-table-fpop-list');
    el.appendChild(list);
    var allRow = cfgEl('div', 'sui-table-fpop-item');
    var allChk = document.createElement('input');
    allChk.type = 'checkbox';
    allChk.setAttribute('aria-label', 'Select all');
    allRow.appendChild(allChk);
    allRow.appendChild(cfgEl('span', 'sui-table-fpop-lbl', '(Select All)'));
    list.appendChild(allRow);
    var rows = [];
    pairs.forEach(function(pair){
      var row = cfgEl('div', 'sui-table-fpop-item');
      var chk = document.createElement('input');
      chk.type = 'checkbox';
      chk.setAttribute('aria-label', 'Filter value ' + pair.label);
      chk.checked = !active || selected.indexOf(pair.raw) >= 0;
      var lbl = cfgEl('span', 'sui-table-fpop-lbl', pair.label);
      lbl.title = pair.label;
      row.appendChild(chk);
      row.appendChild(lbl);
      row.__pair = pair;
      row.__chk = chk;
      row.addEventListener('click', function(ev){
        if (ev.target !== chk) chk.checked = !chk.checked;
        updateAll();
      });
      chk.addEventListener('change', updateAll);
      list.appendChild(row);
      rows.push(row);
    });
    if (!pairs.length) list.appendChild(cfgEl('div', 'sui-table-fpop-empty', '(No values)'));
    function visibleRows(){
      return rows.filter(function(row){ return row.style.display !== 'none'; });
    }
    function updateAll(){
      var vis = visibleRows(), checked = 0;
      vis.forEach(function(row){ if (row.__chk.checked) checked++; });
      allChk.checked = vis.length > 0 && checked === vis.length;
      allChk.indeterminate = checked > 0 && checked < vis.length;
    }
    allChk.addEventListener('change', function(){
      var target = allChk.checked;
      visibleRows().forEach(function(row){ row.__chk.checked = target; });
      updateAll();
    });
    search.addEventListener('input', function(){
      var q = search.value.toLowerCase();
      rows.forEach(function(row){
        var match = !q || row.__pair.label.toLowerCase().indexOf(q) >= 0;
        row.style.display = match ? '' : 'none';
      });
      updateAll();
    });
    var actions = cfgEl('div', 'sui-table-fpop-actions');
    var clear = cfgEl('button', 'sui-table-fpop-btn', 'Clear Filter');
    clear.type = 'button';
    clear.addEventListener('click', function(){
      var st2 = tableStore(name);
      delete st2.colValueFilters[String(col)];
      st2.page = 0;
      clearSelection(st2);
      recordChange(name, { op: 'filter', column: col });
      closeTableFilterPop();
      renderName(name);
    });
    var apply = cfgEl('button', 'sui-table-fpop-btn', 'Apply');
    apply.type = 'button';
    apply.addEventListener('click', function(){
      var st2 = tableStore(name), chosen = [];
      rows.forEach(function(row){ if (row.__chk.checked) chosen.push(row.__pair.raw); });
      if (chosen.length === allValues.length) delete st2.colValueFilters[String(col)];
      else st2.colValueFilters[String(col)] = chosen;
      st2.page = 0;
      clearSelection(st2);
      recordChange(name, { op: 'filter', column: col });
      closeTableFilterPop();
      renderName(name);
    });
    actions.appendChild(clear);
    actions.appendChild(apply);
    el.appendChild(actions);
    document.body.appendChild(el);
    var outside = function(ev){
      if (!tableFilterPop) return;
      if (el.contains(ev.target) || btn.contains(ev.target)) return;
      closeTableFilterPop();
    };
    var esc = function(ev){ if (ev.key === 'Escape') closeTableFilterPop(); };
    var resize = function(){ closeTableFilterPop(); };
    tableFilterPop = { el: el, btn: btn, name: name, col: col, outside: outside, esc: esc, resize: resize };
    btn.setAttribute('aria-expanded', 'true');
    updateAll();
    var r = btn.getBoundingClientRect();
    el.style.left = Math.max(8, Math.min(window.innerWidth - el.offsetWidth - 8, r.right - el.offsetWidth)) + 'px';
    var top = r.bottom + 4;
    if (top + el.offsetHeight > window.innerHeight - 8) top = Math.max(8, r.top - el.offsetHeight - 4);
    el.style.top = top + 'px';
    setTimeout(function(){ document.addEventListener('click', outside, true); }, 0);
    document.addEventListener('keydown', esc, true);
    window.addEventListener('resize', resize);
    window.addEventListener('scroll', resize, true);
    search.focus();
  }
  function ensureHeaderTools(hc, hwrap, hlabel, tbl, name, col, st, cfg, sortDir){
    // The header cell is contenteditable for renaming; its sort/filter controls
    // must stay outside the editable text so a caret or Ctrl+A cannot eat them.
    var hbtns = hwrap.querySelector('.sui-table-hbtns');
    if (hbtns) hbtns.setAttribute('contenteditable', 'false');
    function ensureHbtns(){
      if (!hbtns) {
        hbtns = document.createElement('span');
        hbtns.className = 'sui-table-hbtns';
        hbtns.setAttribute('contenteditable', 'false');
        hwrap.appendChild(hbtns);
      }
      return hbtns;
    }
    var fbtn = hbtns ? hbtns.querySelector('.sui-table-filterbtn') : null;
    if (st.filtersOn) {
      if (!fbtn) {
        ensureHbtns();
        fbtn = document.createElement('button');
        fbtn.type = 'button';
        fbtn.className = 'sui-table-filterbtn';
        fbtn.addEventListener('click', function(ev){
          ev.preventDefault();
          ev.stopPropagation();
          var c = parseInt(hc.getAttribute('data-col'), 10) || 0;
          if (tableFilterPop && tableFilterPop.btn === fbtn) { closeTableFilterPop(); return; }
          openTableFilterPop(fbtn, tbl, name, c);
        });
        hbtns.appendChild(fbtn);
      }
      var active = hasValueFilter(st, col);
      fbtn.classList.toggle('active', active);
      fbtn.title = active ? 'Filter (active)' : 'Filter';
      fbtn.setAttribute('aria-label', 'Filter column ' + (col + 1));
      fbtn.setAttribute('aria-expanded', (tableFilterPop && tableFilterPop.btn === fbtn) ? 'true' : 'false');
    } else if (fbtn && fbtn.parentNode) {
      if (tableFilterPop && tableFilterPop.btn === fbtn) closeTableFilterPop();
      fbtn.parentNode.removeChild(fbtn);
    }
    var sbtn = hbtns ? hbtns.querySelector('.sui-table-sortbtn') : null;
    if (st.sortOn) {
      if (!sbtn) {
        ensureHbtns();
        sbtn = document.createElement('button');
        sbtn.type = 'button';
        sbtn.className = 'sui-table-sortbtn';
        sbtn.addEventListener('click', function(ev){
          ev.preventDefault();
          ev.stopPropagation();
          var c = parseInt(hc.getAttribute('data-col'), 10) || 0;
          applyHeaderSort(name, tbl, c, hc, ev.ctrlKey || ev.metaKey || ev.shiftKey);
        });
        hbtns.appendChild(sbtn);
      }
      sbtn.textContent = '';
      sbtn.classList.toggle('asc', sortDir === 'asc');
      sbtn.classList.toggle('desc', sortDir === 'desc');
      sbtn.classList.toggle('none', sortDir !== 'asc' && sortDir !== 'desc');
      sbtn.title = sortDir === 'asc' ? 'Sorted ascending' : sortDir === 'desc' ? 'Sorted descending' : 'Sort column';
      sbtn.setAttribute('aria-label', (sortDir === 'asc' ? 'Sort column descending' : sortDir === 'desc' ? 'Clear sort' : 'Sort column ascending') + ' (' + (col + 1) + ')');
    } else if (sbtn && sbtn.parentNode) {
      sbtn.parentNode.removeChild(sbtn);
    }
    var rank = hc.getAttribute('data-sort-rank') || '';
    var rankEl = hbtns ? hbtns.querySelector('.sui-table-sort-rank') : null;
    if (rank && st.sortOn) {
      ensureHbtns();
      if (!rankEl) {
        rankEl = document.createElement('span');
        rankEl.className = 'sui-table-sort-rank';
        hbtns.appendChild(rankEl);
      }
      if (rankEl.textContent !== rank) rankEl.textContent = rank;
      rankEl.title = 'Sort level ' + rank;
    } else if (rankEl && rankEl.parentNode) {
      rankEl.parentNode.removeChild(rankEl);
    }
    if (hbtns && !hbtns.children.length && hbtns.parentNode) hbtns.parentNode.removeChild(hbtns);
    hc.classList.toggle('sui-table-filter-active', hasValueFilter(st, col));
  }
  function positionTableCfg(){
    if (!tableCfg) return;
    var btn = tableCfg.btn, el = tableCfg.el;
    if (!btn || !el) return;
    var r = btn.getBoundingClientRect();
    el.style.left = Math.max(8, Math.min(window.innerWidth - el.offsetWidth - 8, r.right - el.offsetWidth)) + 'px';
    var top = r.bottom + 6;
    if (top + el.offsetHeight > window.innerHeight - 8) top = Math.max(8, r.top - el.offsetHeight - 6);
    el.style.top = top + 'px';
  }
  function closeTableCfg(){
    if (!tableCfg) return;
    if (tableCfg.el && tableCfg.el.parentNode) tableCfg.el.parentNode.removeChild(tableCfg.el);
    if (tableCfg.btn) tableCfg.btn.setAttribute('aria-expanded', 'false');
    document.removeEventListener('click', tableCfg.outside, true);
    document.removeEventListener('keydown', tableCfg.esc, true);
    window.removeEventListener('resize', tableCfg.resize);
    window.removeEventListener('scroll', tableCfg.resize, true);
    tableCfg = null;
  }
  function sortKeysCopy(st){
    return sortKeysOf(st).map(function(k){ return { col: k.col, dir: k.dir }; });
  }
  function renderTableCfg(){
    if (!tableCfg) return;
    var name = tableCfg.name, tbl = firstTable(name);
    if (!tbl) { closeTableCfg(); return; }
    var st = tableStore(name), cfg = cfgOf(tbl), box = tableCfg.el;
    box.textContent = '';
    var head = cfgEl('div', 'sui-table-cfg-head');
    head.appendChild(cfgEl('span', 'sui-table-cfg-title', 'Table settings'));
    var close = cfgEl('button', 'sui-table-cfg-close', '\u00d7');
    close.type = 'button';
    close.setAttribute('aria-label', 'Close table settings');
    close.addEventListener('click', closeTableCfg);
    head.appendChild(close);
    box.appendChild(head);

    var frow = cfgEl('div', 'sui-table-cfg-row');
    var fchk = document.createElement('input');
    fchk.type = 'checkbox';
    fchk.checked = !!st.filtersOn;
    fchk.setAttribute('aria-label', 'Filter on column headers');
    fchk.addEventListener('change', function(){
      st.filtersOn = fchk.checked;
      if (!st.filtersOn) {
        st.colFilters = {};
        st.colValueFilters = {};
        st.colFilter = null;
        closeTableFilterPop();
      }
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    frow.appendChild(fchk);
    frow.appendChild(cfgEl('span', '', 'Filter on column headers'));
    box.appendChild(frow);
    box.appendChild(cfgEl('div', 'sui-table-cfg-empty', 'Use the funnel button in each header to pick column values.'));

    var srow = cfgEl('div', 'sui-table-cfg-row');
    var schk = document.createElement('input');
    schk.type = 'checkbox';
    schk.checked = !!st.sortOn;
    schk.setAttribute('aria-label', 'Sort by clicking headers');
    schk.addEventListener('change', function(){
      st.sortOn = schk.checked;
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    srow.appendChild(schk);
    srow.appendChild(cfgEl('span', '', 'Sort by clicking headers'));
    box.appendChild(srow);

    var sect = cfgEl('div', 'sui-table-cfg-sect');
    sect.appendChild(cfgEl('div', 'sui-table-cfg-sec', 'Sort levels'));
    var keys = sortKeysCopy(st);
    var choices = columnChoices(tbl, cfg);
    if (!keys.length) sect.appendChild(cfgEl('div', 'sui-table-cfg-empty', 'No sort levels. Add one, or click a header (Ctrl/Shift+click adds a level).'));
    keys.forEach(function(key, i){
      var row = cfgEl('div', 'sui-table-cfg-level');
      var colSel = document.createElement('select');
      colSel.setAttribute('aria-label', 'Sort column ' + (i + 1));
      choices.forEach(function(ch){
        var o = document.createElement('option');
        o.value = String(ch.index);
        o.textContent = ch.label;
        colSel.appendChild(o);
      });
      colSel.value = String(key.col);
      colSel.addEventListener('change', function(){
        key.col = parseInt(colSel.value, 10) || 0;
        setSortKeys(st, keys);
        st.page = 0;
        clearSelection(st);
        renderName(name);
        renderTableCfg();
      });
      var dirSel = document.createElement('select');
      dirSel.setAttribute('aria-label', 'Sort direction ' + (i + 1));
      [['asc', 'Ascending'], ['desc', 'Descending']].forEach(function(d){
        var o = document.createElement('option');
        o.value = d[0];
        o.textContent = d[1];
        dirSel.appendChild(o);
      });
      dirSel.value = key.dir === 'desc' ? 'desc' : 'asc';
      dirSel.addEventListener('change', function(){
        key.dir = dirSel.value === 'desc' ? 'desc' : 'asc';
        setSortKeys(st, keys);
        st.page = 0;
        clearSelection(st);
        renderName(name);
        renderTableCfg();
      });
      var rm = cfgEl('button', 'sui-table-cfg-btn', '\u00d7');
      rm.type = 'button';
      rm.title = 'Remove sort level';
      rm.setAttribute('aria-label', 'Remove sort level ' + (i + 1));
      rm.addEventListener('click', function(){
        keys.splice(i, 1);
        setSortKeys(st, keys);
        st.page = 0;
        clearSelection(st);
        renderName(name);
        renderTableCfg();
      });
      row.appendChild(colSel);
      row.appendChild(dirSel);
      row.appendChild(rm);
      sect.appendChild(row);
    });
    var add = cfgEl('button', 'sui-table-cfg-btn', '+ Add level');
    add.type = 'button';
    add.addEventListener('click', function(){
      var k2 = sortKeysCopy(st);
      var used = {};
      k2.forEach(function(k){ used[k.col] = 1; });
      var pick = -1;
      for (var c = 0; c < choices.length; c++) {
        if (!used[choices[c].index]) { pick = choices[c].index; break; }
      }
      if (pick < 0 && choices.length) pick = choices[0].index;
      if (pick < 0) return;
      k2.push({ col: pick, dir: 'asc' });
      setSortKeys(st, k2);
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    sect.appendChild(add);
    box.appendChild(sect);

    var acts = cfgEl('div', 'sui-table-cfg-actions');
    var clearF = cfgEl('button', 'sui-table-cfg-btn', 'Clear filters');
    clearF.type = 'button';
    clearF.addEventListener('click', function(){
      st.colFilters = {};
      st.colValueFilters = {};
      st.colFilter = null;
      st.filter = '';
      st.page = 0;
      clearSelection(st);
      closeTableFilterPop();
      renderName(name);
      renderTableCfg();
    });
    var clearS = cfgEl('button', 'sui-table-cfg-btn', 'Clear sort');
    clearS.type = 'button';
    clearS.addEventListener('click', function(){
      setSortKeys(st, []);
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    var done = cfgEl('button', 'sui-table-cfg-btn', 'Done');
    done.type = 'button';
    done.addEventListener('click', closeTableCfg);
    acts.appendChild(clearF);
    acts.appendChild(clearS);
    if (cfg.editMode === 'persistent') {
      var reset = cfgEl('button', 'sui-table-cfg-btn', 'Reset saved data');
      reset.type = 'button';
      reset.addEventListener('click', function(){
        if (!window.confirm('Discard saved table data and restore the design-time table?')) return;
        clearPersisted(name);
        closeTableCfg();
      });
      acts.appendChild(reset);
    }
    acts.appendChild(done);
    box.appendChild(acts);
    positionTableCfg();
  }
  function openTableCfg(tbl, name, btn){
    closeTableCfg();
    var el = cfgEl('div', 'sui-table-cfg');
    el.setAttribute('role', 'dialog');
    el.setAttribute('aria-label', 'Table settings');
    document.body.appendChild(el);
    var outside = function(ev){
      if (!tableCfg) return;
      if (el.contains(ev.target) || (btn && btn.contains(ev.target))) return;
      closeTableCfg();
    };
    var esc = function(ev){ if (ev.key === 'Escape') closeTableCfg(); };
    var resize = function(){ positionTableCfg(); };
    tableCfg = { el: el, name: name, btn: btn, outside: outside, esc: esc, resize: resize };
    btn.setAttribute('aria-expanded', 'true');
    renderTableCfg();
    setTimeout(function(){ document.addEventListener('click', outside, true); }, 0);
    document.addEventListener('keydown', esc, true);
    window.addEventListener('resize', resize);
    window.addEventListener('scroll', resize, true);
  }
  function ensureTableTools(tbl, name){
    var wrap = tbl.closest ? tbl.closest('.tscroll') : null;
    var anchor = wrap || tbl;
    var parent = anchor.parentNode;
    if (!parent) return;
    var bar = null, kids = parent.children;
    for (var i = 0; i < kids.length; i++) {
      if (kids[i].classList && kids[i].classList.contains('sui-table-topbar') &&
          kids[i].getAttribute('data-sui-table-topbar-for') === name) { bar = kids[i]; break; }
    }
    if (!bar) {
      bar = document.createElement('div');
      bar.className = 'sui-table-topbar';
      bar.setAttribute('data-sui-table-topbar-for', name);
      parent.insertBefore(bar, anchor);
    } else if (bar.nextElementSibling !== anchor) {
      parent.insertBefore(bar, anchor);
    }
    // The caption becomes the title row above the table, on the same line as
    // the settings gear. The original <caption> is hidden to avoid duplicating
    // the title inside the table.
    var cap = tbl.querySelector('caption');
    var title = bar.querySelector('.sui-table-title');
    if (cap && (cap.textContent || '').trim()) {
      if (!title) {
        title = document.createElement('div');
        title.className = 'sui-table-title';
        bar.insertBefore(title, bar.firstChild);
      }
      var capText = cap.textContent;
      if (title.textContent !== capText) title.textContent = capText;
      if (cap.getAttribute('data-sui-caption-style') === null) {
        cap.setAttribute('data-sui-caption-style', cap.getAttribute('style') || '');
      }
      var capStyle = cap.getAttribute('data-sui-caption-style') || '';
      if ((title.getAttribute('style') || '') !== capStyle) {
        if (capStyle) title.setAttribute('style', capStyle); else title.removeAttribute('style');
        title.style.display = '';
      }
      if (cap.style.display !== 'none') cap.style.display = 'none';
    } else {
      if (title && title.parentNode) title.parentNode.removeChild(title);
      if (cap && cap.style.display === 'none') cap.style.display = '';
    }
    var host = bar.querySelector('.sui-table-tools');
    if (!host) {
      host = document.createElement('div');
      host.className = 'sui-table-tools';
      bar.appendChild(host);
    }
    var btn = host.querySelector('.sui-table-settings');
    if (!btn) {
      btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'sui-table-settings';
      btn.textContent = '\u2699';
      btn.title = 'Table settings';
      btn.setAttribute('aria-label', 'Table settings');
      btn.setAttribute('aria-expanded', 'false');
      host.appendChild(btn);
      btn.addEventListener('click', function(ev){
        ev.preventDefault();
        ev.stopPropagation();
        if (tableCfg && tableCfg.name === name && tableCfg.btn === btn) { closeTableCfg(); return; }
        openTableCfg(tbl, name, btn);
      });
    }
    var w = (wrap || tbl).getBoundingClientRect().width;
    if (w > 0 && Math.abs(bar.offsetWidth - w) > 1) bar.style.width = Math.round(w) + 'px';
  }
  function ensurePager(tbl, name, st, cfg){
    var next = tbl.nextElementSibling;
    var pager = next && next.classList && next.classList.contains('sui-table-pager') &&
      next.getAttribute('data-sui-table-pager-for') === name ? next : null;
    var pages = st.pageCount || 1;
    if (!cfg.pageSize || pages <= 1) {
      if (pager) pager.style.display = 'none';
      return;
    }
    if (!pager) {
      pager = document.createElement('div');
      pager.className = 'sui-table-pager';
      pager.setAttribute('data-sui-table-pager-for', name);
      if (tbl.parentNode) tbl.parentNode.insertBefore(pager, tbl.nextSibling);
    }
    pager.style.display = '';
    pager.textContent = '';
    function btn(label, delta, disabled) {
      var b = document.createElement('button');
      b.type = 'button';
      b.className = 'sui-table-page-btn';
      b.textContent = label;
      b.disabled = !!disabled;
      b.addEventListener('click', function(){
        var st2 = tableStore(name);
        var p = (st2.page || 0) + delta;
        if (p < 0 || p >= (st2.pageCount || 1)) return;
        st2.page = p;
        clearSelection(st2);
        renderName(name);
      });
      pager.appendChild(b);
    }
    btn('\u2039 Prev', -1, (st.page || 0) <= 0);
    var info = document.createElement('span');
    info.className = 'sui-table-page-info';
    info.textContent = 'Page ' + ((st.page || 0) + 1) + ' / ' + pages;
    pager.appendChild(info);
    btn('Next \u203a', 1, (st.page || 0) >= pages - 1);
  }
  function renderName(name){
    var ts = findTables(name);
    if (!ts.length) return;
    var view = applyView(name), st = view.st, cfg = view.cfg;
    for (var i = 0; i < ts.length; i++) {
      applyTo(ts[i], st);
      decorateTable(ts[i], st, cfg);
      ensureFilterInput(ts[i], name, st, cfg);
      ensurePager(ts[i], name, st, cfg);
      ensureAddTableButton(ts[i], name, st, cfg);
    }
    init();
  }
  function storeApply(name, ds){
    var st = tableStore(name);
    st.header = ds.header ? ds.header.map(function(v){ return String(v); }) : [];
    st.allRows = ds.rows ? ds.rows.map(function(r){ return r.slice(); }) : [];
    st.textHeader = ds.textHeader;
    st.sep = ds.sep;
    st.page = 0;
    st.__fromDom = false;
    st.__colStruct = false;
    clearSelection(st);
    store[name] = st;
    renderName(name);
    return { n: st.allRows.length };
  }
  function display(name, data){
    if (persistentBootBlocks(name)) {
      var kept = store[name] ? store[name].allRows.length : 0;
      return [kept, null];
    }
    var ds = norm(data);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    if (r.error) return [0, r.error];
    recordChange(name, { op: 'display', count: r.n });
    fireTableEvent(name, 'ontablechange');
    return [r.n, null];
  }
  // importText parses text explicitly: see parseSpec for the spec syntax.
  function importText(name, text, spec){
    if (persistentBootBlocks(name)) {
      var keptImport = store[name] ? store[name].allRows.length : 0;
      return [keptImport, null];
    }
    var ds = normText(text, spec);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    if (r.error) return [0, r.error];
    recordChange(name, { op: 'import', count: r.n });
    fireTableEvent(name, 'ontablechange');
    return [r.n, null];
  }
  function data(name){
    var st = store[name];
    if (st) return [st.rows.map(function(r){ return r.slice(); }), null];
    var ts = findTables(name);
    if (!ts.length) return [[], null];
    var out = [], rows = bodyRows(ts[0]);
    for (var i = 0; i < rows.length; i++) {
      var cs = cellsOf(rows[i]), r = [];
      for (var j = 0; j < cs.length; j++) r.push(cs[j].hasAttribute('data-sui-raw') ? (cs[j].getAttribute('data-sui-raw') || '') : (cs[j].textContent || '').trim());
      out.push(r);
    }
    return [out, null];
  }
  function header(name){
    var st = store[name];
    if (st && st.header && st.header.length) return [st.header.slice(), null];
    var ts = findTables(name), out = [];
    if (!ts.length) return [out, null];
    var hr = ts[0].querySelector('thead tr');
    if (hr) {
      var cs = cellsOf(hr);
      for (var i = 0; i < cs.length; i++) out.push((cs[i].textContent || '').trim());
    }
    return [out, null];
  }
  function json(name){
    var h = header(name)[0], rows = data(name)[0], out = [], i, j;
    if (h && h.length) {
      for (i = 0; i < rows.length; i++) {
        var o = Object.create(null);
        for (j = 0; j < rows[i].length; j++) o[h[j] || ('col' + (j + 1))] = rows[i][j];
        out.push(o);
      }
      return [JSON.stringify(out), null];
    }
    return [JSON.stringify(rows), null];
  }
  function select(name, cols){
    if (!Array.isArray(cols)) return [[], 'table.select: columns must be an array of indexes'];
    var rows = data(name)[0], out = [];
    for (var i = 0; i < rows.length; i++) {
      var r = [];
      for (var j = 0; j < cols.length; j++) {
        var c = Number(cols[j]);
        r.push(c >= 0 && c < rows[i].length ? rows[i][c] : '');
      }
      out.push(r);
    }
    return [out, null];
  }
  function clear(name){
    var clearTbl = firstTable(name);
    var clearCfg = clearTbl ? cfgOf(clearTbl) : null;
    delete store[name];
    if (tableCfg && tableCfg.name === name) closeTableCfg();
    if (tableFilterPop && tableFilterPop.name === name) closeTableFilterPop();
    var ts = findTables(name);
    for (var i = 0; i < ts.length; i++) {
      var tb = ts[i].tBodies && ts[i].tBodies.length ? ts[i].tBodies[0] : null;
      if (tb) while (tb.firstChild) tb.removeChild(tb.firstChild);
      var wrap = ts[i].closest ? ts[i].closest('.tscroll') : null;
      var anchor = wrap || ts[i];
      if (anchor.parentNode) {
        var bar = null, kids = anchor.parentNode.children;
        for (var b = 0; b < kids.length; b++) {
          if (kids[b].classList && kids[b].classList.contains('sui-table-topbar') && kids[b].getAttribute('data-sui-table-topbar-for') === name) { bar = kids[b]; break; }
        }
        if (bar) {
          var cap = ts[i].querySelector('caption');
          if (cap) cap.style.display = '';
          bar.parentNode.removeChild(bar);
        }
      }
      var prev = ts[i].previousElementSibling;
      if (prev && prev.classList && prev.classList.contains('sui-table-tools') && prev.getAttribute('data-sui-table-tools-for') === name) {
        var fprev = prev.previousElementSibling;
        prev.parentNode.removeChild(prev);
        prev = fprev;
      }
      if (prev && prev.classList && prev.classList.contains('sui-table-filter') && prev.getAttribute('data-sui-table-filter-for') === name) prev.parentNode.removeChild(prev);
      var next = ts[i].nextElementSibling;
      if (next && next.classList && next.classList.contains('sui-table-pager') && next.getAttribute('data-sui-table-pager-for') === name) next.parentNode.removeChild(next);
    }
    persistTableState(name, { header: [], allRows: [], textHeader: false }, clearCfg);
    recordChange(name, { op: 'clear' });
    fireTableEvent(name, 'ontablechange');
    return [0, null];
  }
  function sort(name, column, dir){
    var st = tableStore(name);
    var tbl = firstTable(name), cfg = tbl ? cfgOf(tbl) : { columns: [], labels: [] };
    var idx = colIndex(cfg, column);
    if (idx < 0) return [0, 'table.sort: column not found'];
    var d = String(dir || '').toLowerCase();
    if (d !== 'asc' && d !== 'desc' && d !== '') return [0, 'table.sort: direction must be asc or desc'];
    setSortKeys(st, d ? [{ col: idx, dir: d }] : []);
    st.page = 0;
    clearSelection(st);
    recordChange(name, { op: 'sort', column: idx, dir: d });
    renderName(name);
    fireTableEvent(name, 'onsort');
    return [1, null];
  }
  function filter(name, column, value){
    var st = tableStore(name);
    var tbl = firstTable(name), cfg = tbl ? cfgOf(tbl) : { columns: [], labels: [] };
    var idx = colIndex(cfg, column);
    if (idx < 0) return [0, 'table.filter: column not found'];
    var fv = String(value === undefined || value === null ? '' : value);
    st.colFilters = {};
    if (fv) st.colFilters[idx] = fv;
    st.colFilter = { column: idx, value: fv };
    st.page = 0;
    clearSelection(st);
    recordChange(name, { op: 'filter', column: idx, value: fv });
    renderName(name);
    return [1, null];
  }
  function filterText(name, text){
    var st = tableStore(name);
    st.filter = String(text === undefined || text === null ? '' : text);
    st.page = 0;
    clearSelection(st);
    recordChange(name, { op: 'filter', column: '', value: st.filter });
    renderName(name);
    return [1, null];
  }
  function page(name, index){
    var st = tableStore(name);
    var p = parseInt(index, 10);
    if (!isFinite(p) || p < 0) return [0, 'table.page: invalid page'];
    st.page = p;
    clearSelection(st);
    renderName(name);
    return [1, null];
  }
  function setPageSize(name, size){
    var st = tableStore(name);
    var n = parseInt(size, 10);
    if (!isFinite(n) || n < 0) return [0, 'table.setPageSize: invalid size'];
    st.pageSize = n;
    st.page = 0;
    renderName(name);
    return [1, null];
  }
  function selectedRows(name){ var st = tableStore(name), out = []; for (var k in st.selectedRows) if (st.selectedRows[k]) out.push(parseInt(k, 10)); return [out, null]; }
  function selectedCells(name){ var st = tableStore(name), out = []; for (var k in st.selectedCells) if (st.selectedCells[k]) { var p = k.split(':'); out.push({ row: parseInt(p[0], 10), column: parseInt(p[1], 10) }); } return [JSON.stringify(out), null]; }
  function selectedColumns(name){ var st = tableStore(name), out = []; for (var k in st.selectedCols) if (st.selectedCols[k]) out.push(parseInt(k, 10)); return [JSON.stringify(out), null]; }
  function lastChange(name){ return [JSON.stringify(changes[name] || null), null]; }
  function applyAll(){
    for (var k in store) {
      try { renderName(k); } catch(e) {}
    }
  }
  function slice(name, from, count){
    var rows = data(name)[0];
    var a = Math.max(0, Number(from) || 0);
    var n = Number(count) || 0;
    return [n > 0 ? rows.slice(a, a + n) : rows.slice(a), null];
  }
  // exportText serializes the visible table as CSV/TSV/SSV/PSV/JSON/JSONL.
  function exportText(name, spec){
    var o = parseSpec(spec, 'csv');
    if (o.format === 'json') return json(name);
    var h = header(name)[0], rows = data(name)[0], i, j, ob;
    if (o.format === 'jsonl' || o.format === 'ndjson') {
      var objs = [];
      if (h.length) {
        for (i = 0; i < rows.length; i++) {
          ob = Object.create(null);
          for (j = 0; j < rows[i].length; j++) ob[h[j] || ('col' + (j + 1))] = rows[i][j];
          objs.push(ob);
        }
      } else {
        objs = rows;
      }
      var txt = objs.map(function(v){ return JSON.stringify(v); }).join('\n');
      return [txt ? txt + '\n' : '', null];
    }
    var sep = o.sep || delimFor(o.format);
    var q = o.quote === undefined ? '"' : o.quote;
    function esc(v){
      v = v === null || v === undefined ? '' : String(v);
      // CSV/Excel formula injection: opt-in prefix for cells that a
      // spreadsheet would evaluate as a formula.
      if (o.formula === 'escape' && /^[=+\-@\t\r]/.test(v)) v = "'" + v;
      if (q && (v.indexOf(sep) >= 0 || v.indexOf(q) >= 0 || v.indexOf('\n') >= 0 || v.indexOf('\r') >= 0)) {
        return q + v.split(q).join(q + q) + q;
      }
      return v;
    }
    var lines = [];
    var useHeader = o.header === null ? true : !!o.header;
    if (useHeader && h.length) lines.push(h.map(esc).join(sep));
    for (i = 0; i < rows.length; i++) lines.push(rows[i].map(esc).join(sep));
    return [lines.join('\n'), null];
  }
  function afterRender(){
    restoreTableBooks();
    applyAll();
    init();
    // Restore the selected tab synchronously, before the browser can paint the
    // regenerated view with the first tab active.
    if (window.__SUI_TABS && window.__SUI_TABS.restore) window.__SUI_TABS.restore();
  }

  // ---- editable / fixed behaviors -----------------------------------------
  function ensureDeclaredHeader(tbl, cfg){
    if (tbl.querySelector('thead tr')) return;
    var labels = cfg.labels.length ? cfg.labels : cfg.columns;
    if (!labels.length) return;
    var thead = tbl.querySelector('thead');
    if (!thead) { thead = document.createElement('thead'); tbl.insertBefore(thead, tbl.firstChild); }
    var tr = document.createElement('tr');
    for (var i = 0; i < labels.length; i++) {
      var th = document.createElement('th');
      th.textContent = labels[i];
      if (cfg.widths[i] > 0) { th.style.width = cfg.widths[i] + 'px'; th.style.minWidth = cfg.widths[i] + 'px'; }
      if (cfg.align[i]) th.style.textAlign = cfg.align[i];
      tr.appendChild(th);
    }
    thead.appendChild(tr);
  }
  function syncStoreFromDom(tbl, name){
    var st = tableStore(name);
    if (st.__fromDom || st.allRows.length) return;
    st.__fromDom = true;
    st.header = [];
    st.allRows = [];
    var hr = tbl.querySelector('thead tr');
    if (hr) cellsOf(hr).forEach(function(c){ st.header.push(c.getAttribute('data-sui-raw') || c.textContent || ''); });
    var rows = bodyRows(tbl);
    for (var i = 0; i < rows.length; i++) {
      var cs = cellsOf(rows[i]), r = [];
      for (var j = 0; j < cs.length; j++) r.push(cs[j].getAttribute('data-sui-raw') || cs[j].textContent || '');
      st.allRows.push(r);
    }
    st.rows = st.allRows;
    store[name] = st;
  }
  // Runtime workbook sheet deletion is kept for the page session. Source sheets
  // removed at runtime stay hidden across view re-renders; runtime sheets are
  // also dropped from tableBooks so restoreTableBooks cannot recreate them.
  var deletedWorkbookSheets = {};
  function workbookTabList(box){
    var list = box ? box.querySelector('.sui-tablist') : null;
    return list ? Array.prototype.filter.call(list.children, function(el){
      return el.classList && el.classList.contains('sui-tab');
    }) : [];
  }
  function workbookPanelList(box){
    var panels = box ? box.querySelector('.sui-tabpanels') : null;
    return panels ? Array.prototype.filter.call(panels.children, function(el){
      return el.classList && el.classList.contains('sui-tabpanel');
    }) : [];
  }
  function workbookSheetLabel(tab){
    return (tab.getAttribute('data-sui-sheet-label') || tab.textContent.replace(/\u00d7\s*$/,'').trim());
  }
  function workbookSheetKey(table, label){
    var name = table ? (table.getAttribute('name') || '') : '';
    return name ? ('n:' + name) : ('l:' + label);
  }
  function ensureWorkbookSheetLabels(box){
    var tabs = workbookTabList(box);
    for (var i = 0; i < tabs.length; i++) {
      if (!tabs[i].hasAttribute('data-sui-sheet-label')) {
        tabs[i].setAttribute('data-sui-sheet-label', workbookSheetLabel(tabs[i]));
      }
    }
  }
  function applyWorkbookDeletions(box){
    var dkey = tableBookKey(box);
    var marks = deletedWorkbookSheets[dkey];
    if (!marks) return;
    var tabs = workbookTabList(box), panels = workbookPanelList(box);
    var activeRemoved = false;
    for (var i = tabs.length - 1; i >= 0; i--) {
      var panel = panels[i];
      var table = panel ? panel.querySelector('table') : null;
      var key = workbookSheetKey(table, workbookSheetLabel(tabs[i]));
      if (!marks[key]) continue;
      if (tabs[i].classList.contains('active')) activeRemoved = true;
      if (tabs[i].parentNode) tabs[i].parentNode.removeChild(tabs[i]);
      if (panel && panel.parentNode) panel.parentNode.removeChild(panel);
    }
    if (activeRemoved) {
      var rest = workbookTabList(box);
      if (rest.length) { try { rest[0].click(); } catch (e) {} }
    }
  }
  function deleteWorkbookSheet(box, index){
    var tabs = workbookTabList(box);
    if (tabs.length <= 1) return;
    var tab = tabs[index];
    if (!tab) return;
    ensureWorkbookSheetLabels(box);
    var panels = workbookPanelList(box);
    var panel = panels[index];
    var table = panel ? panel.querySelector('table') : null;
    var key = workbookSheetKey(table, workbookSheetLabel(tab));
    var dkey = tableBookKey(box);
    var marks = deletedWorkbookSheets[dkey] || (deletedWorkbookSheets[dkey] = {});
    marks[key] = true;
    var name = table ? (table.getAttribute('name') || '') : '';
    if (name && tableBooks[dkey]) {
      tableBooks[dkey] = tableBooks[dkey].filter(function(def){ return def.name !== name; });
    }
    var wasActive = tab.classList.contains('active');
    if (tab.parentNode) tab.parentNode.removeChild(tab);
    if (panel && panel.parentNode) panel.parentNode.removeChild(panel);
    if (name) {
      try { if (window.localStorage) window.localStorage.removeItem(tablePersistKey(name)); } catch (e) {}
    }
    saveTableBookDefs(box);
    // Update the visible tab count, editable + state and the x buttons.
    normalizeWorkbookRuntime();
    var rest = workbookTabList(box);
    if (rest.length) {
      var active = -1;
      for (var i = 0; i < rest.length; i++) if (rest[i].classList.contains('active')) { active = i; break; }
      if (active < 0) active = wasActive ? Math.min(index, rest.length - 1) : 0;
      // Re-click the active tab so the tabs runtime stores the shifted index.
      try { rest[active].click(); } catch (e) {}
    }
  }
  function ensureWorkbookDeleteButtons(box){
    var designer = !!(document.body && document.body.classList && document.body.classList.contains('sui-designer-preview'));
    var tabs = workbookTabList(box), panels = workbookPanelList(box);
    ensureWorkbookSheetLabels(box);
    for (var i = 0; i < tabs.length; i++) {
      var panel = panels[i];
      var table = panel ? panel.querySelector('table') : null;
      var editable = !!(table && table.getAttribute('data-sui-editable') === '1');
      var x = tabs[i].querySelector('.sui-tab-x');
      if (!editable && !designer) {
        if (x && x.parentNode) x.parentNode.removeChild(x);
        continue;
      }
      if (x) continue;
      x = document.createElement('span');
      x.className = 'sui-tab-x';
      x.textContent = '\u00d7';
      x.title = 'Delete workbook';
      x.setAttribute('aria-label', 'Delete workbook');
      x.contentEditable = 'false';
      x.addEventListener('click', function(ev){
        ev.preventDefault();
        ev.stopPropagation();
        var tab = this.parentNode;
        var book = tab && tab.closest ? tab.closest('.sui-tabs') : null;
        if (!book) return;
        var list = workbookTabList(book);
        var idx = Array.prototype.indexOf.call(list, tab);
        if (idx >= 0) deleteWorkbookSheet(book, idx);
      });
      tabs[i].appendChild(x);
    }
  }
  // Runtime tab context menu: Delete workbook on an editable sheet. The
  // designer picker intercepts the contextmenu event first in its own preview.
  document.addEventListener('contextmenu', function(ev){
    var tab = ev.target && ev.target.closest ? ev.target.closest('.sui-tab') : null;
    if (!tab) return;
    var box = tab.closest('.sui-tabs');
    if (!box) return;
    var tabs = workbookTabList(box);
    if (tabs.length <= 1) return;
    var idx = Array.prototype.indexOf.call(tabs, tab);
    var panels = workbookPanelList(box);
    var table = panels[idx] ? panels[idx].querySelector('table') : null;
    if (!table || table.getAttribute('data-sui-editable') !== '1') return;
    ev.preventDefault();
    ev.stopPropagation();
    var old = document.getElementById('sui-tab-menu-runtime');
    if (old) old.remove();
    var menu = document.createElement('div');
    menu.id = 'sui-tab-menu-runtime';
    menu.className = 'sui-table-struct-menu';
    menu.setAttribute('role', 'menu');
    var item = document.createElement('button');
    item.type = 'button';
    item.className = 'sui-table-struct-item';
    item.textContent = 'Delete workbook';
    item.addEventListener('click', function(e){
      e.stopPropagation();
      menu.remove();
      deleteWorkbookSheet(box, idx);
    });
    menu.appendChild(item);
    document.body.appendChild(menu);
    var x = ev.clientX, y = ev.clientY;
    var rect = menu.getBoundingClientRect();
    if (x + rect.width > window.innerWidth - 8) x = Math.max(8, window.innerWidth - rect.width - 8);
    if (y + rect.height > window.innerHeight - 8) y = Math.max(8, window.innerHeight - rect.height - 8);
    menu.style.left = x + 'px';
    menu.style.top = y + 'px';
    var close = function(){ menu.remove(); document.removeEventListener('click', close); };
    setTimeout(function(){ document.addEventListener('click', close); }, 0);
  });
  // A tabs control whose panels all hold a table is a workbook: every sheet is
  // a data-entry surface, shares one visible table size and offers the + tab.
  // This also upgrades sources built before the designer normalized workbooks.
  function normalizeWorkbookRuntime(){
    var boxes = document.querySelectorAll('.sui-tabs');
    for (var b = 0; b < boxes.length; b++) {
      var panels = boxes[b].querySelectorAll('.sui-tabpanels .sui-tabpanel');
      if (!panels.length) continue;
      applyWorkbookDeletions(boxes[b]);
      panels = boxes[b].querySelectorAll('.sui-tabpanels .sui-tabpanel');
      if (!panels.length) continue;
      var tables = [], all = true;
      for (var p = 0; p < panels.length; p++) {
        var t = panels[p].querySelector('table');
        if (!t) { all = false; break; }
        tables.push(t);
      }
      if (!all || !tables.length) continue;
      boxes[b].setAttribute('data-sui-workbook', '1');
      // A single workbook has no sheet tabs; the + tab remains available to
      // create the second sheet. Two or more sheets show their tabs.
      var tabBtns = boxes[b].querySelectorAll('.sui-tablist .sui-tab');
      boxes[b].setAttribute('data-sui-tabcount', String(tabBtns.length));
      var anyEditable = false;
      for (var tj = 0; tj < tables.length; tj++) {
        var t2 = tables[tj];
        var sheetMode = t2.getAttribute('data-sui-edit-mode') || 'none';
        if (sheetMode === 'none') {
          // Non-editable sheet: keep it read-only even inside a workbook that is
          // otherwise editable.
          t2.removeAttribute('data-sui-editable');
          t2.removeAttribute('data-sui-show-add-table');
        } else {
          t2.setAttribute('data-sui-editable', '1');
          // Keep the + in the DOM for workbooks: CSS hides it when no sheet is
          // editable (the designer preview shows it).
          if (t2.getAttribute('name')) t2.setAttribute('data-sui-show-add-table', '1');
        }
        if (t2.getAttribute('data-sui-editable') === '1') anyEditable = true;
        // Stable cell sizing: truncate is the default, wrap is explicit.
        var cellText = t2.getAttribute('data-sui-cell-text') || '';
        if (!cellText) cellText = 'truncate';
        t2.setAttribute('data-sui-cell-text', cellText);
        if (cellText === 'truncate') t2.setAttribute('data-sui-fixed', '1');
        var w2 = t2.parentElement;
        if (w2 && w2.classList && w2.classList.contains('tscroll')) {
          // The sheet area follows the control size; per-sheet fixed sizes made
          // the sheets diverge and broke the shared workbook layout.
          w2.style.width = '';
          w2.style.height = '';
          w2.style.maxWidth = '';
          w2.style.maxHeight = '';
        }
      }
      // The runtime + is shown only for an editable workbook; the designer
      // preview overrides this with its body class and always offers it.
      boxes[b].setAttribute('data-sui-workbook-editable', anyEditable ? '1' : '0');
      ensureWorkbookDeleteButtons(boxes[b]);
    }
  }
  // The workbook table width is driven by the declared/default column widths,
  // never by cell text. When the sum exceeds the viewport the table overflows
  // and the sheet area shows the horizontal scrollbar.
  function updateWorkbookTableWidth(tbl){
    if (!tbl || !tbl.closest) return;
    var box = tbl.closest('.sui-tabs');
    if (!box || box.getAttribute('data-sui-workbook') !== '1') return;
    var row = (tbl.tHead && tbl.tHead.rows.length) ? tbl.tHead.rows[0] :
      ((tbl.tBodies && tbl.tBodies.length && tbl.tBodies[0].rows.length) ? tbl.tBodies[0].rows[0] : null);
    if (!row) return;
    var sum = 0;
    for (var i = 0; i < row.cells.length; i++) {
      var c = row.cells[i];
      var sw = (c.style && (c.style.width || c.style.minWidth)) || '';
      var w = parseFloat(sw);
      if (!isFinite(w) || sw.indexOf('%') >= 0) w = 160;
      sum += w * (c.colSpan || 1);
    }
    if (sum > 0) tbl.style.minWidth = Math.round(sum) + 'px';
  }
  function init(){
    normalizeWorkbookRuntime();
    var edTables = document.querySelectorAll('table[data-sui-editable]');
    for (var et = 0; et < edTables.length; et++) {
      var etMode = cfgOf(edTables[et]).editMode;
      var etCells = edTables[et].querySelectorAll('td,th');
      for (var ec = 0; ec < etCells.length; ec++) {
        if (etMode === 'none') {
          if (etCells[ec].getAttribute('contenteditable') === 'true') etCells[ec].removeAttribute('contenteditable');
        } else if (etCells[ec].getAttribute('contenteditable') !== 'true') {
          etCells[ec].setAttribute('contenteditable', 'true');
          etCells[ec].spellcheck = false;
        }
      }
    }
    var fix = document.querySelectorAll('table[data-sui-fixed] td, table[data-sui-fixed] th');
    for (var j=0;j<fix.length;j++){
      var d=fix[j];
      if (!d.hasAttribute('tabindex')) d.tabIndex=0;
      var isEd = d.getAttribute('contenteditable') === 'true';
      if (!isEd) { var txt = d.textContent; if (d.getAttribute('data-full') !== txt) d.setAttribute('data-full', txt); }
    }
    restoreTableBooks();
    // Restored runtime sheets change the tab count: update the workbook tab
    // visibility before decorating the restored tables.
    normalizeWorkbookRuntime();
    var named = document.querySelectorAll('table[name]');
    for (var t = 0; t < named.length; t++) {
      var tbl = named[t], name = tableName(tbl);
      if (!name) continue;
      var cfg = cfgOf(tbl);
      ensureDeclaredHeader(tbl, cfg);
      syncStoreFromDom(tbl, name);
      var st = store[name];
      captureTableSource(tbl, st);
      if (cfg.editMode === 'persistent' && !st.__persistLoaded) {
        var savedTable = loadPersistedTable(name);
        if (savedTable) applyPersistedTable(st, savedTable);
        st.__persistLoaded = true;
      }
      applyPersistedTableAttrs(tbl, st);
      var view = applyView(name);
      cfg = view.cfg;
      st = view.st;
      decorateTable(tbl, st, cfg);
      ensureFilterInput(tbl, name, st, cfg);
      ensurePager(tbl, name, st, cfg);
      ensureAddTableButton(tbl, name, st, cfg);
    }
  }
  var pop=null, popCell=null;
  function closePop(){ if(pop){ if(pop.parentNode) pop.parentNode.removeChild(pop); pop=null; popCell=null; } }
  function openPop(c, txt){
    if (popCell === c) return;
    closePop();
    pop=document.createElement('div'); pop.className='sui-cell-pop'; pop.textContent=txt;
    var cs=getComputedStyle(c);
    var bg=cs.getPropertyValue('--pop-bg').trim(); if(bg) pop.style.background=bg;
    var fg=cs.getPropertyValue('--pop-color').trim(); if(fg) pop.style.color=fg;
    var fs=cs.getPropertyValue('--pop-size').trim(); if(fs) pop.style.fontSize=fs;
    var mw=cs.getPropertyValue('--pop-maxw').trim(); if(mw) pop.style.maxWidth=mw;
    var br=cs.getPropertyValue('--pop-radius').trim(); if(br) pop.style.borderRadius=br;
    document.body.appendChild(pop);
    var r=c.getBoundingClientRect();
    var pr=pop.getBoundingClientRect();
    var left=Math.max(8, Math.min(window.innerWidth - pr.width - 8, r.left));
    var top=r.top - pr.height - 8;
    if (top < 8) top = Math.min(window.innerHeight - pr.height - 8, r.bottom + 6);
    pop.style.left=left+'px'; pop.style.top=top+'px';
    popCell=c;
  }
  document.addEventListener('mouseover', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-fixed] td, table[data-sui-fixed] th') : null;
    if(!c){ return; }
    if(document.activeElement === c) return; // being edited: text is already visible
    if(c.scrollWidth <= c.clientWidth + 1 && c.scrollHeight <= c.clientHeight + 1) { closePop(); return; }
    openPop(c, c.textContent);
  }, true);
  document.addEventListener('mouseout', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-fixed] td, table[data-sui-fixed] th') : null;
    if(c) closePop();
  }, true);
  document.addEventListener('keydown', function(ev){ if(ev.key==='Escape') closePop(); });

  // ---- draggable row/column boundaries (editable tables only) --------------
  var drag=null;
  function columnCells(cell){
    var tr=cell.parentNode, idx=Array.prototype.indexOf.call(tr.children, cell);
    var rows=cell.closest('table').querySelectorAll('tr'), out=[];
    for(var i=0;i<rows.length;i++){ var c=rows[i].children[idx]; if(c) out.push(c); }
    return out;
  }
  function edge(el, ev){
    var r=el.getBoundingClientRect();
    if ((r.right - ev.clientX) <= 7 && ev.clientX >= r.left + 10) return 'col';
    if ((r.bottom - ev.clientY) <= 7 && ev.clientY >= r.top + 8) return 'row';
    return '';
  }
  document.addEventListener('mousemove', function(ev){
    if(drag) return;
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-editable] td, table[data-sui-editable] th') : null;
    if(!c) return;
    var e=edge(c, ev);
    c.style.cursor = e === 'col' ? 'col-resize' : e === 'row' ? 'row-resize' : '';
  }, true);
  document.addEventListener('mousedown', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-editable] td, table[data-sui-editable] th') : null;
    if(!c) return;
    var e=edge(c, ev);
    if(!e) return;
    ev.preventDefault(); ev.stopPropagation();
    var r=c.getBoundingClientRect();
    drag = { kind:e, cell:c, start: e==='col' ? ev.clientX : ev.clientY, size: e==='col' ? r.width : r.height };
    document.body.style.cursor = e==='col' ? 'col-resize' : 'row-resize';
  }, true);
  document.addEventListener('mousemove', function(ev){
    if(!drag) return;
    ev.preventDefault();
    if(drag.kind==='col'){
      var w=Math.max(24, drag.size + (ev.clientX - drag.start));
      var cells=columnCells(drag.cell);
      for(var i=0;i<cells.length;i++){ cells[i].style.width=w+'px'; cells[i].style.minWidth=w+'px'; }
      updateWorkbookTableWidth(drag.cell.closest('table'));
    } else {
      var h=Math.max(18, drag.size + (ev.clientY - drag.start));
      var tr=drag.cell.parentNode;
      tr.style.height=h+'px';
      for(var j=0;j<tr.children.length;j++){ tr.children[j].style.height=h+'px'; }
    }
  }, true);
  document.addEventListener('mouseup', function(){
    if (!drag) return;
    var dragTable = drag.cell && drag.cell.closest ? drag.cell.closest('table') : null;
    var dragName = dragTable ? (dragTable.getAttribute('name') || '') : '';
    drag = null;
    document.body.style.cursor = '';
    if (dragName) {
      recordChange(dragName, { op: 'resize' });
      persistTable(dragName);
    }
  }, true);

  // ---- keep edited cells in the store so re-renders do not lose them -------
  // Header cells also contain sort/filter controls: the editable text is the
  // .sui-table-hlabel span, never the whole th (which includes the controls).
  function editTextOf(c){
    var lbl = c && c.querySelector ? c.querySelector('.sui-table-hlabel') : null;
    return lbl ? (lbl.textContent || '') : (c ? (c.textContent || '') : '');
  }
  function setCellEditText(c, txt){
    var lbl = c && c.querySelector ? c.querySelector('.sui-table-hlabel') : null;
    if (lbl) lbl.textContent = txt;
    else if (c) c.textContent = txt;
  }
  function focusInside(c){
    return !!(c && c.contains && document.activeElement && c.contains(document.activeElement));
  }
  // Keep the declarative label attribute in sync after a header edit so tools
  // that read data-sui-column-labels see the new text before the next render.
  function syncHeaderLabelsAttr(tbl){
    if (!tbl || !tbl.hasAttribute('data-sui-column-labels')) return;
    var cells = tbl.querySelectorAll('thead tr:first-child th, thead tr:first-child td');
    var labels = [];
    for (var i = 0; i < cells.length; i++) {
      labels.push(editTextOf(cells[i]));
    }
    var str = labels.join(',');
    if (tbl.getAttribute('data-sui-column-labels') !== str) tbl.setAttribute('data-sui-column-labels', str);
  }
  document.addEventListener('input', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    if (!c) return;
    var tbl = c.closest('table');
    if (!tbl) return;
    var name = tbl.getAttribute('name');
    if (!name || !store[name]) return;
    if (editCell !== c) {
      editCell = c;
      var committed = c.getAttribute('data-sui-raw');
      editOld = committed === null ? editTextOf(c) : committed;
    }
    var st = store[name];
    var tr = c.parentNode, col = Array.prototype.indexOf.call(tr.children, c);
    if (col < 0) return;
    var text = editTextOf(c);
    // Commit data-sui-raw on blur. Updating it on every keystroke would
    // spam the MutationObserver and make the table look like it is redrawing.
    if (tr.parentNode && tr.parentNode.tagName === 'THEAD') {
      if (!st.header) st.header = [];
      if (!st.headerEdits) st.headerEdits = {};
      st.header[col] = text;
      st.headerEdits[col] = text;
    } else {
      var tb = tr.parentNode, r = Array.prototype.indexOf.call(tb.rows, tr);
      if (r >= 0) {
        if (!st.allRows[r]) st.allRows[r] = [];
        st.allRows[r][col] = text;
        if (st.rows[r]) st.rows[r][col] = text;
      }
    }
    // The insertion phase ends after a pause even if the cell stays focused.
    clearEditTimer();
    editTimer = setTimeout(function(){
      editTimer = null;
      if (!c.isConnected) { if (editCell === c) { editCell = null; editOld = ''; } return; }
      commitCellEdit(c, tbl, name, cfgOf(tbl), false);
      if (editCell === c && (!c.isConnected || !focusInside(c))) { editCell = null; editOld = ''; }
    }, 900);
  }, true);
  var editCell = null, editOld = '', editTimer = null;
  function clearEditTimer(){ if (editTimer) { clearTimeout(editTimer); editTimer = null; } }
  // commitCellEdit stores the current cell text and fires the change events.
  // finalize=true (blur) also validates, restores invalid values and formats
  // the cell. finalize=false (idle timeout) commits valid text without moving
  // the caret or re-formatting the focused cell.
  function commitCellEdit(c, tbl, name, cfg, finalize){
    if (!c || !c.isConnected) {
      if (editCell === c) { editCell = null; editOld = ''; }
      return false;
    }
    var tr = c.parentNode, col = Array.prototype.indexOf.call(tr.children, c);
    if (col < 0) return false;
    var raw = (editTextOf(c) || '').trim();
    var oldVal = editOld === null || editOld === undefined ? raw : String(editOld);
    var rowIdx = tr.parentNode && tr.parentNode.tagName === 'THEAD' ? -1 : Array.prototype.indexOf.call(tr.parentNode.rows, tr);
    var type = cfg.types[col] || 'text';
    var err = rowIdx < 0 ? '' : validValue(raw, type, colOpts(cfg, col));
    var st = store[name];
    if (err) {
      if (!finalize) return false;
      setCellEditText(c, oldVal);
      c.classList.add('sui-table-invalid');
      c.setAttribute('aria-invalid', 'true');
      c.title = 'Invalid ' + err + ' value';
      if (st) {
        if (tr.parentNode && tr.parentNode.tagName === 'THEAD') {
          if (st.header) st.header[col] = oldVal;
        } else {
          var r0 = Array.prototype.indexOf.call(tr.parentNode.rows, tr);
          if (r0 >= 0) { if (!st.allRows[r0]) st.allRows[r0] = []; st.allRows[r0][col] = oldVal; if (st.rows[r0]) st.rows[r0][col] = oldVal; }
        }
      }
      return false;
    }
    c.classList.remove('sui-table-invalid');
    c.removeAttribute('aria-invalid');
    c.removeAttribute('title');
    if (raw !== oldVal) {
      if (st) {
        if (rowIdx < 0) {
          if (!st.header) st.header = [];
          if (!st.headerEdits) st.headerEdits = {};
          st.header[col] = raw;
          st.headerEdits[col] = raw;
        } else {
          if (!st.allRows[rowIdx]) st.allRows[rowIdx] = [];
          st.allRows[rowIdx][col] = raw;
          if (st.rows[rowIdx]) st.rows[rowIdx][col] = raw;
        }
      }
      c.setAttribute('data-sui-raw', raw);
      if (finalize && rowIdx < 0) {
        var labelEl = c.querySelector('.sui-table-hlabel');
        if (labelEl) {
          if (labelEl.textContent !== raw) labelEl.textContent = raw;
        } else if (c.textContent !== raw) {
          c.textContent = raw;
        }
      }
      if (finalize) {
        recordChange(name, { op: 'cell', row: rowIdx, column: col, old: oldVal, new: raw, value: raw });
        fireTableEvent(name, 'oncellchange');
        if (rowIdx >= 0) fireTableEvent(name, 'onrowchange');
        fireTableEvent(name, 'ontablechange');
        editOld = raw;
      }
    }
    if (finalize && rowIdx >= 0) {
      var fv = formatValue(raw, cfg.formats[col]);
      if (c.textContent !== fv) c.textContent = fv;
    }
    if (rowIdx < 0) syncHeaderLabelsAttr(tbl);
    return true;
  }
  document.addEventListener('focusin', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    var tbl = c && c.closest ? c.closest('table[data-sui-editable]') : null;
    if (!tbl) return;
    clearEditTimer();
    editCell = c;
    editOld = c.getAttribute('data-sui-raw');
    if (editOld === null) editOld = editTextOf(c);
    else if ((editTextOf(c) || '') !== editOld) setCellEditText(c, editOld);
    c.classList.add('sui-table-editing');
    c.classList.remove('sui-table-invalid');
    c.removeAttribute('aria-invalid');
    c.removeAttribute('title');
  }, true);
  document.addEventListener('focusout', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    if (!c || c !== editCell) return;
    clearEditTimer();
    var tbl = c.closest('table[data-sui-editable]');
    var name = tbl ? tbl.getAttribute('name') : '';
    if (!tbl || !name) { c.classList.remove('sui-table-editing'); editCell = null; editOld = ''; return; }
    commitCellEdit(c, tbl, name, cfgOf(tbl), true);
    c.classList.remove('sui-table-editing');
    editCell = null; editOld = '';
  }, true);

  if(window.MutationObserver){
    var pending=false;
    new MutationObserver(function(){
      if (editCell && !editCell.isConnected) { editCell = null; editOld = ''; }
      var ae = document.activeElement;
      if (editCell || (ae && ae.closest && ae.closest('td[contenteditable="true"], th[contenteditable="true"]'))) return;
      if(pending) return; pending=true; setTimeout(function(){
        pending=false;
        var a2 = document.activeElement;
        if (!editCell && !(a2 && a2.closest && a2.closest('td[contenteditable="true"], th[contenteditable="true"]'))) init();
      }, 60);
    }).observe(document.body,{childList:true,subtree:true,characterData:true});
  }
  init();
  return {
    display: display, data: data, header: header, json: json,
    select: select, slice: slice, clear: clear, importText: importText,
    exportText: exportText, applyAll: applyAll, afterRender: afterRender,
    sort: sort, filter: filter, filterText: filterText, page: page, setPageSize: setPageSize,
    selectedRows: selectedRows, selectedCells: selectedCells, selectedColumns: selectedColumns,
    lastChange: lastChange, clearPersisted: clearPersisted
  };
})();
window.__SUI_TBL = __SUI_TBL;`
