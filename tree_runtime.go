package main

// tree_runtime.go - browser runtime for the nested tree control.
//
// The generated markup is a role=tree with nested role=treeitem/group nodes.
// The runtime owns open state, roving focus, keyboard navigation and selection;
// selection is synchronized with a hidden value input and optional state setter
// + onChange handler. Open state is persisted per tree path in localStorage.
const suiTreeRuntime = `(function(){
  var lastDown = { tree: '', key: '', time: 0 };
  var lastSelection = { tree: '', key: '', keys: [] };
  var treeSearchTerms = Object.create(null);
  var treeUndo = Object.create(null);
  var treeLastMove = Object.create(null);
  var treeColumnOverrides = Object.create(null);
  var treeCellOverrides = Object.create(null);
  // Imported tree models survive reactive re-renders: store[name] is applied
  // to every freshly rendered tree with the same data-sui-tree-name.
  var store = Object.create(null);
  var MAX_NODES = 10000;
  var MAX_DEPTH = 64;
  var MAX_LABEL = 2000;
  var MAX_KEY = 512;
  var MAX_ICON = 64;
  var MAX_BADGE = 128;
  var MAX_META = 256;
  function treeText(v) { return v === null || v === undefined ? '' : String(v); }
  function parseCellsText(raw) {
    var out = {};
    var s = treeText(raw).trim();
    if (!s) return out;
    if (s.charCodeAt(0) === 123) {
      try {
        var obj = JSON.parse(s);
        if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
          Object.keys(obj).forEach(function(k){
            var v = obj[k];
            if (v !== null && typeof v === 'object') return;
            out[k] = treeText(v);
          });
        }
      } catch (e) {}
      return out;
    }
    var sep = s.indexOf(';') >= 0 ? ';' : '|';
    s.split(sep).forEach(function(pair){
      var i = pair.indexOf('=');
      if (i <= 0) return;
      out[pair.slice(0, i).trim()] = pair.slice(i + 1).trim();
    });
    return out;
  }
  function treeName(root) {
    return root.getAttribute('data-sui-tree-name') || root.getAttribute('data-sui-tree-key') || 'tree';
  }
  function findTrees(name) {
    var all = document.querySelectorAll('.sui-tree'), out = [];
    for (var i = 0; i < all.length; i++) {
      if (treeName(all[i]) === name) out.push(all[i]);
    }
    return out;
  }
  function treeTarget(name) {
    var ts = findTrees(name);
    return ts.length ? ts[0] : null;
  }
  function itemGroup(item) {
    for (var i = 0; i < item.children.length; i++) {
      var ch = item.children[i];
      if (ch.classList && ch.classList.contains('sui-tree-group')) return ch;
    }
    return null;
  }
  function itemRow(item) {
    for (var i = 0; i < item.children.length; i++) {
      var ch = item.children[i];
      if (ch.classList && ch.classList.contains('sui-tree-row')) return ch;
    }
    return null;
  }
  function itemToggle(item) {
    var row = itemRow(item);
    return row ? row.querySelector('.sui-tree-toggle') : null;
  }
  function itemLabel(item) {
    var label = item.querySelector('.sui-tree-label');
    return label ? label.textContent : '';
  }
  function parentItem(item, root) {
    var p = item.parentElement;
    while (p && p !== root) {
      if (p.hasAttribute && p.hasAttribute('data-sui-tree-item')) return p;
      p = p.parentElement;
    }
    return null;
  }
  function storageKey(root, path) {
    var key = root.getAttribute('data-sui-tree-key') || 'tree';
    return 'sui.tree.' + key + '.' + path;
  }
  function storedOpen(root, path) {
    try {
      var raw = localStorage.getItem(storageKey(root, path));
      if (raw === '1') return true;
      if (raw === '0') return false;
    } catch (e) {}
    return null;
  }
  function saveOpen(root, path, open) {
    try { localStorage.setItem(storageKey(root, path), open ? '1' : '0'); } catch (e) {}
  }
  function init(root) {
    if (!root) return;
    if (root.__suiTree) {
      // A reactive re-render may replace the tree's rows while the root keeps
      // its bound flag: re-init when any item is still unbound so new rows get
      // their handlers instead of silently losing click/context-menu support.
      var probe = root.querySelectorAll('[data-sui-tree-item]');
      var lacksBind = false;
      for (var pb = 0; pb < probe.length; pb++) {
        if (!probe[pb].__suiBound) { lacksBind = true; break; }
      }
      if (!lacksBind) return;
      root.__suiTree = false;
      root.__suiTreeApi = null;
    }
    var treeKey = root.getAttribute('data-sui-tree-key') || 'tree';
    var guides = root.getAttribute('data-sui-tree-guides') === '1';
    var checkboxes = root.getAttribute('data-sui-tree-checkbox') === '1';
    var counts = root.getAttribute('data-sui-tree-counts') === '1';
    var multiSelect = root.getAttribute('data-sui-tree-multi') === '1';
    var searchEnabled = root.getAttribute('data-sui-tree-search') === '1';
    var dragEnabled = root.getAttribute('data-sui-tree-drag') === '1';
    var contextMenuEnabled = root.getAttribute('data-sui-tree-context-menu') === '1';
    var customTooltips = root.getAttribute('data-sui-tree-tooltips') === '1';
    var treeColumns = (root.getAttribute('data-sui-tree-columns') || '').split(',').filter(Boolean);
    var treeColumnsEnabled = root.getAttribute('data-sui-tree-columns-enabled') !== '0';
    if (!treeColumnsEnabled) treeColumns = [];
    var treeColumnWidths = (root.getAttribute('data-sui-tree-column-widths') || '').split(',').map(function(v) { return parseInt(v, 10) || 0; });
    var treeColumnLabels = (root.getAttribute('data-sui-tree-column-labels') || '').split(',');
    var sortDir = (root.getAttribute('data-sui-tree-sort') || '').toLowerCase();
    var fontStep = parseFloat(root.getAttribute('data-sui-tree-font-step') || '0');
    if (!isFinite(fontStep) || fontStep < 0) fontStep = 0;
    var persistSelection = root.getAttribute('data-sui-tree-persist-selection') === '1';
    var treeIcon = root.getAttribute('data-sui-tree-icon') || '';
    var treeIconOpen = root.getAttribute('data-sui-tree-icon-open') || '';
    var treeIconLeaf = root.getAttribute('data-sui-tree-icon-leaf') || '';
    var treeIconParent = root.getAttribute('data-sui-tree-icon-parent') || '';
    var treeIconParentOpen = root.getAttribute('data-sui-tree-icon-parent-open') || '';
    var treeToggleClosed = root.getAttribute('data-sui-tree-toggle-closed') || '\u25b8';
    var treeToggleOpen = root.getAttribute('data-sui-tree-toggle-open') || '\u25be';
    var rowHeight = parseInt(root.getAttribute('data-sui-tree-row-height') || '', 10);
    if (!isNaN(rowHeight) && rowHeight >= 12 && rowHeight <= 120) {
      root.style.setProperty('--sui-tree-row-height', rowHeight + 'px');
    }
    var imported = store[treeName(root)];
    if (imported) {
      renderNodes(root, imported, {
        guides: guides, checkbox: checkboxes, counts: counts,
        icon: treeIcon, iconOpen: treeIconOpen, iconLeaf: treeIconLeaf,
        iconParent: treeIconParent, iconParentOpen: treeIconParentOpen,
        toggleClosed: treeToggleClosed, toggleOpen: treeToggleOpen,
        columns: treeColumns,
        columnWidths: treeColumnWidths, sort: sortDir, fontStep: fontStep
      });
    }
    root.__suiTree = true;
    var hidden = root.querySelector('.sui-tree-value');
    if (!hidden) return;
    var setter = root.getAttribute('data-sui-tree-setter') || '';
    var handler = root.getAttribute('data-sui-tree-onchange') || '';
    var checkHandler = root.getAttribute('data-sui-tree-oncheck') || '';
    var loadHandler = root.getAttribute('data-sui-tree-onload') || '';
    var moveHandler = root.getAttribute('data-sui-tree-onmove') || '';
    var contextHandler = root.getAttribute('data-sui-tree-oncontextmenu') || '';
    var treeChangeHandler = root.getAttribute('data-sui-tree-ontreechange') || '';
    var selectionSetter = root.getAttribute('data-sui-tree-selection-setter') || '';
    var loadKeySetter = root.getAttribute('data-sui-tree-load-key-setter') || '';
    var renameKeySetter = root.getAttribute('data-sui-tree-rename-key-setter') || '';
    var renameValueSetter = root.getAttribute('data-sui-tree-rename-value-setter') || '';
    var renameHandler = root.getAttribute('data-sui-tree-onrename') || '';
    var treeEditable = root.getAttribute('data-sui-tree-editable') === '1';
    var editableDepth = parseInt(root.getAttribute('data-sui-tree-editable-depth') || '0', 10);
    if (isNaN(editableDepth) || editableDepth < 0) editableDepth = 0;
    var treeRenameIcon = root.getAttribute('data-sui-tree-rename-icon') || '\u270e';
    var showTreeRenameIcon = root.getAttribute('data-sui-tree-show-rename-icon') !== '0';
    var items = Array.prototype.slice.call(root.querySelectorAll('[data-sui-tree-item]'));
    if (!items.length) return;
    var current = hidden.value || '';
    var selectionHidden = multiSelect ? root.querySelector('.sui-tree-selection') : null;
    var selectedItems = [];
    var selectionAnchor = null;
    var searchSnapshot = null;
    var contextMenuBound = false;
    var editing = null;
    var selectionTimer = null;
    var pendingItem = null;
    var checkedStore = {};
    var storedSelection = [];
    if (checkboxes) {
      try {
        var rawChecked = localStorage.getItem('sui.tree.checked.' + treeKey);
        var parsedChecked = rawChecked ? JSON.parse(rawChecked) : null;
        if (parsedChecked && typeof parsedChecked === 'object' && !Array.isArray(parsedChecked)) checkedStore = parsedChecked;
      } catch (e) { checkedStore = {}; }
    }
    if (persistSelection) {
      try {
        var rawSelection = localStorage.getItem('sui.tree.selection.' + treeKey);
        var parsedStoredSelection = rawSelection ? JSON.parse(rawSelection) : null;
        if (Array.isArray(parsedStoredSelection)) storedSelection = parsedStoredSelection;
      } catch (e) { storedSelection = []; }
    }
    if (!root.getAttribute('aria-label')) root.setAttribute('aria-label', treeKey);
    if (checkboxes || multiSelect) root.setAttribute('aria-multiselectable', 'true');

    function itemLabelEl(item) {
      return item.querySelector('.sui-tree-label');
    }
    function itemCells(item) {
      var raw = item.getAttribute('data-sui-cells') || '';
      if (!raw) return {};
      try {
        var obj = JSON.parse(raw);
        if (obj && typeof obj === 'object' && !Array.isArray(obj)) return obj;
      } catch (e) {}
      var out = {};
      var sep = raw.indexOf(';') >= 0 ? ';' : '|';
      raw.split(sep).forEach(function(pair){
        var i = pair.indexOf('=');
        if (i <= 0) return;
        out[pair.slice(0, i).trim()] = pair.slice(i + 1).trim();
      });
      return out;
    }
    function itemCellValue(item, col) {
      if (col === 'size' || col === 'date' || col === 'status') return item.getAttribute('data-sui-' + col) || '';
      var cells = itemCells(item);
      return cells[col] === undefined ? '' : String(cells[col]);
    }
    function columnLabel(col, index) {
      if (treeColumnLabels[index] && treeColumnLabels[index].trim()) return treeColumnLabels[index].trim();
      var parts = String(col).split(/[_-]+/).filter(Boolean);
      for (var i = 0; i < parts.length; i++) parts[i] = parts[i].charAt(0).toUpperCase() + parts[i].slice(1);
      return parts.length ? parts.join(' ') : col;
    }
    function rebuildHeader() {
      var old = root.querySelector('.sui-tree-grid-head');
      if (old && old.parentNode) old.parentNode.removeChild(old);
      if (!treeColumnsEnabled || !treeColumns.length) return;
      var head = document.createElement('div');
      head.className = 'sui-tree-grid-head';
      head.setAttribute('role', 'presentation');
      var name = document.createElement('span');
      name.className = 'sui-tree-grid-name';
      name.textContent = 'Name';
      head.appendChild(name);
      for (var i = 0; i < treeColumns.length; i++) {
        var col = treeColumns[i];
        var cell = document.createElement('span');
        cell.className = 'sui-tree-cell sui-tree-cell-' + col;
        cell.textContent = columnLabel(col, i);
        if (treeColumnWidths[i] > 0) cell.style.width = treeColumnWidths[i] + 'px';
        head.appendChild(cell);
      }
      var group = root.querySelector('.sui-tree-group');
      if (group && group.parentNode === root) root.insertBefore(head, group);
      else root.appendChild(head);
    }
    function rebuildRowCells(item) {
      var row = itemRow(item);
      if (!row) return;
      var old = row.querySelectorAll('.sui-tree-cell');
      for (var i = 0; i < old.length; i++) old[i].remove();
      if (!treeColumnsEnabled || !treeColumns.length) return;
      for (var c = 0; c < treeColumns.length; c++) {
        var col = treeColumns[c];
        var cell = document.createElement('span');
        cell.className = 'sui-tree-cell sui-tree-cell-' + col;
        cell.textContent = itemCellValue(item, col);
        if (treeColumnWidths[c] > 0) cell.style.width = treeColumnWidths[c] + 'px';
        row.appendChild(cell);
      }
    }
    function applyRuntimeColumns() {
      var override = treeColumnOverrides[treeName(root)];
      if (override) {
        treeColumns = override.fields.slice();
        treeColumnLabels = override.labels.slice();
        treeColumnWidths = override.widths.slice();
        treeColumnsEnabled = true;
      }
      rebuildHeader();
      for (var i = 0; i < items.length; i++) rebuildRowCells(items[i]);
    }
    function applyRuntimeCells() {
      var overrides = treeCellOverrides[treeName(root)];
      if (!overrides) return;
      for (var i = 0; i < items.length; i++) {
        var key = items[i].getAttribute('data-sui-key') || '';
        if (!overrides[key]) continue;
        var cells = itemCells(items[i]);
        Object.keys(overrides[key]).forEach(function(col){ cells[col] = overrides[key][col]; });
        items[i].setAttribute('data-sui-cells', JSON.stringify(cells));
      }
    }
    function setColumnsFor(list) {
      if (!Array.isArray(list)) throw new Error('columns must be an array');
      if (list.length > 12) throw new Error('too many columns (limit 12)');
      var fields = [], labels = [], widths = [];
      for (var i = 0; i < list.length; i++) {
        var entry = list[i];
        if (typeof entry === 'string') entry = { key: entry };
        if (!entry || typeof entry !== 'object') throw new Error('invalid column');
        var key = String(entry.key || entry.field || entry.name || '').trim().replace(/[^A-Za-z0-9_-]/g, '');
        if (!key) throw new Error('column key is required');
        var width = parseInt(entry.width, 10);
        if (!isFinite(width) || width <= 0) width = 104;
        fields.push(key);
        labels.push(entry.label === undefined ? columnLabel(key, i) : String(entry.label));
        widths.push(Math.max(20, Math.min(600, width)));
      }
      treeColumnOverrides[treeName(root)] = { fields: fields, labels: labels, widths: widths };
      treeColumns = fields; treeColumnLabels = labels; treeColumnWidths = widths; treeColumnsEnabled = true;
      applyRuntimeColumns();
      return fields.length;
    }
    function getColumnsFor() {
      var out = [];
      for (var i = 0; i < treeColumns.length; i++) {
        out.push({ key: treeColumns[i], label: columnLabel(treeColumns[i], i), width: treeColumnWidths[i] || 104 });
      }
      return out;
    }
    function setCellFor(key, column, value) {
      var item = findItemByKey(key);
      if (!item) throw new Error('tree item not found: ' + key);
      var col = String(column || '').replace(/[^A-Za-z0-9_-]/g, '');
      if (!col) throw new Error('column key is required');
      var cells = itemCells(item);
      cells[col] = String(value === null || value === undefined ? '' : value);
      item.setAttribute('data-sui-cells', JSON.stringify(cells));
      var map = treeCellOverrides[treeName(root)] || (treeCellOverrides[treeName(root)] = {});
      map[key] = map[key] || {};
      map[key][col] = cells[col];
      rebuildRowCells(item);
      notifyTreeChange();
      return true;
    }
    function labelStorageKey(root, path) {
      return 'sui.tree.label.' + treeKey + '.' + path;
    }
    function storedLabel(root, path) {
      try { return localStorage.getItem(labelStorageKey(root, path)); } catch (e) { return null; }
    }
    function saveLabel(root, path, value) {
      try { localStorage.setItem(labelStorageKey(root, path), value); } catch (e) {}
    }
    function editableFor(item) {
      var p = parentItem(item, root);
      while (p) {
        if (p.getAttribute('data-sui-editable') === '0') return false;
        p = parentItem(p, root);
      }
      var own = item.getAttribute('data-sui-editable') || 'inherit';
      if (own === '0') return false;
      if (own === '1') return true;
      if (!treeEditable) return false;
      var level = parseInt(item.getAttribute('data-sui-level') || '1', 10);
      if (editableDepth > 0 && level > editableDepth) return false;
      return true;
    }
    function cancelRename() {
      if (!editing) return;
      var old = editing;
      editing = null;
      old.label.hidden = false;
      if (old.input.parentNode) old.input.parentNode.removeChild(old.input);
    }
    function commitRename() {
      if (!editing) return;
      var old = editing;
      var value = String(old.input.value || '').trim();
      if (!value) {
        old.input.classList.add('invalid');
        old.input.setAttribute('aria-invalid', 'true');
        old.input.focus();
        return;
      }
      editing = null;
      old.label.textContent = value;
      old.item.__suiLabel = value;
      old.label.title = value;
      old.label.hidden = false;
      if (old.input.parentNode) old.input.parentNode.removeChild(old.input);
      var key = old.item.getAttribute('data-sui-key') || '';
      var path = old.item.getAttribute('data-sui-path') || '';
      saveLabel(root, path, value);
      if (renameKeySetter && typeof window[renameKeySetter] === 'function') window[renameKeySetter](key);
      if (renameValueSetter && typeof window[renameValueSetter] === 'function') window[renameValueSetter](value);
      if (renameHandler && typeof window[renameHandler] === 'function') window[renameHandler]();
      notifyTreeChange();
    }
    function startRename(item) {
      if (!item.__suiEditable || editing) return;
      if (selectionTimer) { clearTimeout(selectionTimer); selectionTimer = null; }
      if (pendingItem) { pendingItem.classList.remove('sui-tree-pending'); pendingItem = null; }
      cancelRename();
      var label = itemLabelEl(item);
      var row = itemRow(item);
      if (!label || !row) return;
      var input = document.createElement('input');
      input.type = 'text';
      input.className = 'sui-tree-rename';
      input.maxLength = 200;
      input.value = label.textContent;
      input.setAttribute('aria-label', 'Rename item');
      input.addEventListener('mousedown', function(ev) { ev.stopPropagation(); });
      input.addEventListener('click', function(ev) { ev.stopPropagation(); });
      input.addEventListener('input', function() { input.classList.remove('invalid'); input.removeAttribute('aria-invalid'); });
      input.addEventListener('keydown', function(ev) {
        ev.stopPropagation();
        if (ev.key === 'Enter') { ev.preventDefault(); commitRename(); }
        else if (ev.key === 'Escape') { ev.preventDefault(); cancelRename(); }
      });
      input.addEventListener('blur', function() { setTimeout(cancelRename, 0); });
      label.hidden = true;
      row.appendChild(input);
      editing = { item: item, label: label, input: input };
      input.focus();
      input.select();
    }
    function refreshTreeItemIcon(item, open) {
      if (!item) return;
      var row = itemRow(item);
      if (!row) return;
      var isParent = itemGroup(item) || item.getAttribute('data-sui-lazy') === '1';
      var base = item.getAttribute('data-sui-icon-base') || '';
      var baseOpen = item.getAttribute('data-sui-icon-open-base') || '';
      var icon = base;
      if (icon === '') icon = isParent ? (treeIconParent || treeIcon) : (treeIconLeaf || treeIcon);
      if (open && isParent) {
        if (baseOpen) icon = baseOpen;
        else if (treeIconParentOpen) icon = treeIconParentOpen;
        else if (treeIconOpen) icon = treeIconOpen;
      }
      var iconEl = row.querySelector('.sui-tree-icon');
      if (icon) {
        if (!iconEl) {
          iconEl = document.createElement('span');
          iconEl.className = 'sui-tree-icon';
          iconEl.setAttribute('aria-hidden', 'true');
          var anchor = itemLabelEl(item);
          if (anchor && anchor.parentNode === row) row.insertBefore(iconEl, anchor);
          else row.appendChild(iconEl);
        }
        if (iconEl.textContent !== icon) iconEl.textContent = icon;
      } else if (iconEl && iconEl.parentNode) {
        iconEl.parentNode.removeChild(iconEl);
      }
      item.setAttribute('data-sui-icon', icon || '');
    }
    function setOpen(item, open, persist) {
      var group = itemGroup(item);
      if (!group) return;
      group.hidden = !open;
      item.setAttribute('aria-expanded', open ? 'true' : 'false');
      var toggle = itemToggle(item);
      if (toggle) {
        toggle.textContent = open ? treeToggleOpen : treeToggleClosed;
        toggle.setAttribute('aria-label', (open ? 'Collapse ' : 'Expand ') + itemLabel(item));
      }
      refreshTreeItemIcon(item, open);
      item.__suiTreeOpen = open;
      if (persist) saveOpen(root, item.getAttribute('data-sui-path') || '', open);
    }
    function visible(item) {
      if (item.hidden) return false;
      var p = item.parentElement;
      while (p && p !== root) {
        if (p.hidden) return false;
        if (p.classList && p.classList.contains('sui-tree-group') && p.hidden) return false;
        p = p.parentElement;
      }
      return true;
    }
    function focusItem(item) {
      items.forEach(function(x) { x.tabIndex = -1; });
      item.tabIndex = 0;
      item.focus();
    }
    function selectionJSON() {
      var keys = [];
      for (var i = 0; i < selectedItems.length; i++) keys.push(selectedItems[i].getAttribute('data-sui-key') || '');
      return JSON.stringify(keys);
    }
    function setSelection(list, notify, sync) {
      var uniq = [];
      for (var i = 0; i < list.length; i++) {
        if (list[i] && uniq.indexOf(list[i]) < 0) uniq.push(list[i]);
      }
      selectedItems = uniq;
      if (sync !== false) {
        lastSelection.tree = treeKey;
        lastSelection.keys = [];
        for (var lk = 0; lk < uniq.length; lk++) lastSelection.keys.push(uniq[lk].getAttribute('data-sui-key') || '');
        if (persistSelection) {
          try { localStorage.setItem('sui.tree.selection.' + treeKey, JSON.stringify(lastSelection.keys)); } catch (e) {}
        }
      }
      var last = uniq.length ? uniq[uniq.length - 1] : null;
      items.forEach(function(x) {
        var on = uniq.indexOf(x) >= 0;
        x.classList.toggle('selected', on);
        x.setAttribute('aria-selected', on ? 'true' : 'false');
        x.tabIndex = -1;
      });
      if (last) last.tabIndex = 0;
      else {
        for (var f = 0; f < items.length; f++) {
          if (visible(items[f])) { items[f].tabIndex = 0; break; }
        }
      }
      current = last ? (last.getAttribute('data-sui-key') || '') : '';
      hidden.value = current;
      if (selectionHidden) selectionHidden.value = selectionJSON();
      if (sync !== false) {
        if (selectionSetter && typeof window[selectionSetter] === 'function') window[selectionSetter](selectionJSON());
      }
      if (notify) {
        if (setter && typeof window[setter] === 'function') window[setter](current);
        if (handler && typeof window[handler] === 'function') window[handler]();
      }
    }
    function select(item, notify, mods) {
      mods = mods || {};
      if (!item) return;
      if (!multiSelect) {
        setSelection([item], notify);
        selectionAnchor = item;
        lastSelection.tree = treeKey;
        lastSelection.key = item.getAttribute('data-sui-key') || '';
        return;
      }
      var vis = items.filter(visible);
      if (mods.shift && selectionAnchor) {
        var a = vis.indexOf(selectionAnchor), b = vis.indexOf(item);
        if (a >= 0 && b >= 0) {
          var start = Math.min(a, b), end = Math.max(a, b);
          setSelection(vis.slice(start, end + 1), notify);
          return;
        }
      }
      if (mods.ctrl || mods.meta) {
        var idx = selectedItems.indexOf(item);
        var next = selectedItems.slice();
        if (idx >= 0) next.splice(idx, 1);
        else next.push(item);
        setSelection(next, notify);
        selectionAnchor = item;
        lastSelection.tree = treeKey;
        lastSelection.key = item.getAttribute('data-sui-key') || '';
        return;
      }
      setSelection([item], notify);
      selectionAnchor = item;
      lastSelection.tree = treeKey;
      lastSelection.key = item.getAttribute('data-sui-key') || '';
    }
    function itemCheck(item) {
      var row = itemRow(item);
      if (!row) return null;
      for (var i = 0; i < row.children.length; i++) {
        if (row.children[i].classList && row.children[i].classList.contains('sui-tree-check')) return row.children[i];
      }
      return null;
    }
    function childItems(item) {
      var group = itemGroup(item);
      var out = [];
      if (!group) return out;
      for (var i = 0; i < group.children.length; i++) {
        var ch = group.children[i];
        if (ch.hasAttribute && ch.hasAttribute('data-sui-tree-item')) out.push(ch);
      }
      return out;
    }
    function checkStateText(v) {
      return v === true ? 'true' : v === 'mixed' ? 'mixed' : 'false';
    }
    function applyCheckState(item, state) {
      item.__suiChecked = state;
      var txt = checkStateText(state);
      item.setAttribute('aria-checked', txt);
      var box = itemCheck(item);
      if (box) box.setAttribute('aria-checked', txt);
    }
    function recomputeUp(item) {
      var kids = childItems(item);
      if (!kids.length) {
        if (item.__suiChecked === undefined) item.__suiChecked = item.getAttribute('data-sui-checked') === '1';
        applyCheckState(item, item.__suiChecked === true);
        return item.__suiChecked === true;
      }
      var all = true, none = true;
      for (var i = 0; i < kids.length; i++) {
        var v = recomputeUp(kids[i]);
        if (v !== true) all = false;
        if (v !== false) none = false;
      }
      var state = all ? true : (none ? false : 'mixed');
      applyCheckState(item, state);
      return state === true;
    }
    function setCheckedDeep(item, value) {
      applyCheckState(item, value === true);
      var kids = childItems(item);
      for (var i = 0; i < kids.length; i++) setCheckedDeep(kids[i], value);
    }
    function saveCheckedState() {
      var data = {};
      for (var i = 0; i < items.length; i++) {
        var p = items[i].getAttribute('data-sui-path') || '';
        data[p] = items[i].__suiChecked === true ? 1 : 0;
      }
      try { localStorage.setItem('sui.tree.checked.' + treeKey, JSON.stringify(data)); } catch (e) {}
    }
    function notifyCheck() {
      if (checkHandler && typeof window[checkHandler] === 'function') window[checkHandler]();
    }
    function notifyTreeChange() {
      if (treeChangeHandler && typeof window[treeChangeHandler] === 'function') window[treeChangeHandler]();
    }
    function toggleCheck(item) {
      if (!checkboxes) return;
      var next = item.__suiChecked !== true;
      setCheckedDeep(item, next);
      var parent = parentItem(item, root);
      while (parent) {
        recomputeUp(parent);
        parent = parentItem(parent, root);
      }
      for (var i = 0; i < items.length; i++) {
        if (items[i].__suiChecked === undefined) recomputeUp(items[i]);
      }
      saveCheckedState();
      notifyCheck();
      notifyTreeChange();
    }
    function initChecks() {
      if (!checkboxes) return;
      for (var i = 0; i < items.length; i++) {
        var p = items[i].getAttribute('data-sui-path') || '';
        if (Object.prototype.hasOwnProperty.call(checkedStore, p)) {
          items[i].__suiChecked = checkedStore[p] === 1 || checkedStore[p] === true || checkedStore[p] === '1';
        } else {
          items[i].__suiChecked = items[i].getAttribute('data-sui-checked') === '1';
        }
      }
      // Recompute every root so parents derive their state from children.
      for (var j = 0; j < items.length; j++) {
        if (!parentItem(items[j], root)) recomputeUp(items[j]);
      }
    }
    function findItemByKey(key) {
      for (var i = 0; i < items.length; i++) {
        if ((items[i].getAttribute('data-sui-key') || '') === String(key)) return items[i];
      }
      return null;
    }
    function freshItemByKey(key) {
      var all = document.querySelectorAll('[data-sui-tree-item]');
      var name = treeName(root);
      for (var i = 0; i < all.length; i++) {
        var owner = all[i].closest ? all[i].closest('.sui-tree') : null;
        if (owner && treeName(owner) === name && (all[i].getAttribute('data-sui-key') || '') === String(key)) return all[i];
      }
      return null;
    }
    function renderOpts() {
      return { guides: guides, checkbox: checkboxes, counts: counts, icon: treeIcon, iconOpen: treeIconOpen,
        iconLeaf: treeIconLeaf, iconParent: treeIconParent, iconParentOpen: treeIconParentOpen,
        toggleClosed: treeToggleClosed, toggleOpen: treeToggleOpen,
        columns: treeColumns, columnWidths: treeColumnWidths, sort: sortDir, fontStep: fontStep };
    }
    function ensureGroup(item) {
      var group = itemGroup(item);
      if (group) return group;
      group = document.createElement('ul');
      group.className = 'sui-tree-group';
      group.setAttribute('role', 'group');
      item.appendChild(group);
      return group;
    }
    function ensureToggle(item) {
      var row = itemRow(item);
      if (!row) return null;
      var toggle = itemToggle(item);
      if (toggle) return toggle;
      for (var i = 0; i < row.children.length; i++) {
        if (row.children[i].classList && row.children[i].classList.contains('sui-tree-spacer')) {
          toggle = document.createElement('button');
          toggle.type = 'button';
          toggle.className = 'sui-tree-toggle';
          toggle.tabIndex = -1;
          toggle.setAttribute('aria-label', 'Toggle ' + itemLabel(item));
          toggle.textContent = treeToggleClosed;
          row.replaceChild(toggle, row.children[i]);
          return toggle;
        }
      }
      return null;
    }
    function renumberItem(item, level, path) {
      item.setAttribute('data-level', String(level));
      item.setAttribute('data-path', path);
      item.setAttribute('aria-level', String(level));
      var group = itemGroup(item);
      if (!group) return;
      var n = 0;
      for (var i = 0; i < group.children.length; i++) {
        var child = group.children[i];
        if (child.hasAttribute && child.hasAttribute('data-sui-tree-item')) {
          renumberItem(child, level + 1, path + '.' + n);
          n++;
        }
      }
    }
    function renumberItems() {
      var group = rootGroup(root), n = 0;
      for (var i = 0; i < group.children.length; i++) {
        var item = group.children[i];
        if (item.hasAttribute && item.hasAttribute('data-sui-tree-item')) {
          renumberItem(item, 1, String(n));
          n++;
        }
      }
    }
    function expandAll() {
      var n = 0;
      for (var i = 0; i < items.length; i++) {
        if (itemGroup(items[i])) { setOpen(items[i], true, true); n++; }
      }
      return n;
    }
    function collapseAll() {
      var n = 0;
      for (var i = 0; i < items.length; i++) {
        if (itemGroup(items[i])) { setOpen(items[i], false, true); n++; }
      }
      return n;
    }
    function expandTo(level) {
      level = Math.max(0, parseInt(level, 10) || 0);
      var n = 0;
      for (var i = 0; i < items.length; i++) {
        if (!itemGroup(items[i])) continue;
        var itemLevel = parseInt(items[i].getAttribute('data-sui-level') || '1', 10);
        setOpen(items[i], itemLevel < level, true);
        n++;
      }
      return n;
    }
    function selectedKeys() {
      var out = [];
      for (var i = 0; i < selectedItems.length; i++) out.push(selectedItems[i].getAttribute('data-sui-key') || '');
      return out;
    }
    function checkedKeys() {
      var out = [];
      for (var i = 0; i < items.length; i++) {
        if (items[i].getAttribute('aria-checked') === 'true') out.push(items[i].getAttribute('data-sui-key') || '');
      }
      return out;
    }
    function searchLabelText(item) {
      var label = itemLabelEl(item);
      return label ? label.textContent : '';
    }
    function searchText(item) {
      return [
        searchLabelText(item),
        item.getAttribute('data-sui-badge') || '',
        item.getAttribute('data-sui-meta') || '',
        item.getAttribute('data-sui-size') || '',
        item.getAttribute('data-sui-date') || '',
        item.getAttribute('data-sui-status') || ''
      ].join(' ').toLowerCase();
    }
    function subtreeMatches(item, term) {
      if (searchText(item).indexOf(term) >= 0) return true;
      var kids = childItems(item);
      for (var i = 0; i < kids.length; i++) {
        if (subtreeMatches(kids[i], term)) return true;
      }
      return false;
    }
    function computeMatches(item, term, cache) {
      var key = item.getAttribute('data-sui-path') || '';
      if (cache[key] !== undefined) return cache[key];
      var sub = searchText(item).indexOf(term) >= 0;
      var kids = childItems(item);
      for (var i = 0; i < kids.length; i++) {
        if (computeMatches(kids[i], term, cache)) sub = true;
      }
      cache[key] = sub;
      return sub;
    }
    function highlightLabel(item, term) {
      var label = itemLabelEl(item);
      if (!label) return;
      var text = item.__suiLabel !== undefined ? item.__suiLabel : label.textContent;
      if (!term) {
        label.textContent = text;
        return;
      }
      label.textContent = '';
      var lower = text.toLowerCase();
      var pos = 0, idx = lower.indexOf(term, pos);
      while (idx >= 0) {
        if (idx > pos) label.appendChild(document.createTextNode(text.slice(pos, idx)));
        var mark = document.createElement('mark');
        mark.className = 'sui-tree-match';
        mark.textContent = text.slice(idx, idx + term.length);
        label.appendChild(mark);
        pos = idx + term.length;
        idx = lower.indexOf(term, pos);
      }
      if (pos < text.length) label.appendChild(document.createTextNode(text.slice(pos)));
    }
    function searchEmpty() {
      var el = root.querySelector('.sui-tree-empty');
      if (!el) {
        el = document.createElement('div');
        el.className = 'sui-tree-empty';
        el.setAttribute('role', 'status');
        el.textContent = 'No results';
        root.appendChild(el);
      }
      return el;
    }
    function applySearch(value) {
      if (!searchEnabled) return;
      var term = String(value || '').trim().toLowerCase();
      var searchName = treeName(root);
      if (!term) delete treeSearchTerms[searchName];
      else treeSearchTerms[searchName] = String(value || '');
      var input = root.querySelector('.sui-tree-search-input');
      var countEl = root.querySelector('.sui-tree-search-count');
      if (!term) {
        for (var i = 0; i < items.length; i++) {
          items[i].hidden = false;
          highlightLabel(items[i], '');
        }
        if (searchSnapshot) {
          for (var j = 0; j < searchSnapshot.length; j++) {
            var snap = searchSnapshot[j];
            var item = findItemByKey(snap.key);
            if (item && itemGroup(item)) setOpen(item, snap.open, false);
          }
        }
        searchSnapshot = null;
        var empty = root.querySelector('.sui-tree-empty');
        if (empty) empty.hidden = true;
        if (countEl) countEl.textContent = '';
        return;
      }
      if (!searchSnapshot) {
        searchSnapshot = [];
        for (var s = 0; s < items.length; s++) {
          searchSnapshot.push({ key: items[s].getAttribute('data-sui-key') || '', open: !!items[s].__suiTreeOpen });
        }
      }
      var matches = 0;
      var matchCache = {};
      for (var k = 0; k < items.length; k++) computeMatches(items[k], term, matchCache);
      for (var j2 = 0; j2 < items.length; j2++) {
        var it = items[j2];
        var own = searchText(it).indexOf(term) >= 0;
        var sub = matchCache[it.getAttribute('data-sui-path') || ''] === true;
        it.hidden = !sub;
        highlightLabel(it, term);
        if (own) matches++;
      }
      for (var e = 0; e < items.length; e++) {
        var node = items[e];
        if (!node.hidden && itemGroup(node) && matchCache[node.getAttribute('data-sui-path') || '']) setOpen(node, true, false);
      }
      var emptyEl = searchEmpty();
      emptyEl.hidden = matches > 0;
      if (countEl) countEl.textContent = matches > 0 ? matches + (matches === 1 ? ' match' : ' matches') : '';
      if (input && input.value !== value) input.value = value;
    }
    function focusSearch() {
      var input = root.querySelector('.sui-tree-search-input');
      if (input) { input.focus(); input.select(); }
    }
    var tooltipTimer = null;
    var tooltipEl = null;
    function ensureTooltip() {
      if (tooltipEl && tooltipEl.parentNode) return tooltipEl;
      tooltipEl = root.querySelector('.sui-tree-tooltip');
      if (!tooltipEl) {
        tooltipEl = document.createElement('div');
        tooltipEl.className = 'sui-tree-tooltip';
        tooltipEl.setAttribute('role', 'tooltip');
        tooltipEl.id = 'sui-tree-tip-' + treeKey.replace(/[^A-Za-z0-9_-]/g, '-');
        root.appendChild(tooltipEl);
      }
      return tooltipEl;
    }
    function tooltipDelay() {
      var raw = getComputedStyle(root).getPropertyValue('--sui-tree-tooltip-delay').trim();
      var n = parseFloat(raw);
      return isFinite(n) ? Math.max(0, Math.min(2000, n)) : 140;
    }
    function hideTooltip() {
      if (tooltipTimer) { clearTimeout(tooltipTimer); tooltipTimer = null; }
      if (tooltipEl) tooltipEl.classList.remove('on');
    }
    function positionTooltip(anchor) {
      if (!tooltipEl || !anchor) return;
      var placement = getComputedStyle(root).getPropertyValue('--sui-tree-tooltip-placement').trim().toLowerCase();
      if (["top", "right", "bottom", "left"].indexOf(placement) < 0) placement = "top";
      var r = anchor.getBoundingClientRect();
      var tw = tooltipEl.offsetWidth, th = tooltipEl.offsetHeight;
      var gap = 6;
      var x = r.left, y = r.bottom + gap;
      if (placement === "top") { x = r.left; y = r.top - th - gap; }
      else if (placement === "right") { x = r.right + gap; y = r.top + (r.height - th) / 2; }
      else if (placement === "bottom") { x = r.left; y = r.bottom + gap; }
      else if (placement === "left") { x = r.left - tw - gap; y = r.top + (r.height - th) / 2; }
      x = Math.max(4, Math.min(window.innerWidth - tw - 4, x));
      y = Math.max(4, Math.min(window.innerHeight - th - 4, y));
      tooltipEl.style.left = x + 'px';
      tooltipEl.style.top = y + 'px';
    }
    function showTooltip(item, anchor, immediate) {
      if (!customTooltips || !anchor) return;
      var text = item.getAttribute('data-sui-tooltip') || (anchor.textContent || '');
      if (!text) return;
      hideTooltip();
      function reveal() {
        var el = ensureTooltip();
        el.textContent = text;
        el.classList.add('on');
        positionTooltip(anchor);
      }
      if (immediate) reveal();
      else tooltipTimer = setTimeout(reveal, tooltipDelay());
    }
    function loadLazy(item) {
      if (!item || item.__suiLoading) return;
      if (itemGroup(item)) { setOpen(item, true, true); return; }
      item.__suiLoading = true;
      item.setAttribute('data-sui-loading', '1');
      item.setAttribute('aria-busy', 'true');
      item.removeAttribute('aria-invalid');
      var row = itemRow(item);
      if (row) {
        var oldError = row.querySelector('.sui-tree-error');
        if (oldError && oldError.parentNode) oldError.parentNode.removeChild(oldError);
        var oldRetry = row.querySelector('.sui-tree-retry');
        if (oldRetry && oldRetry.parentNode) oldRetry.parentNode.removeChild(oldRetry);
        var existing = row.querySelector('.sui-tree-loading');
        if (!existing) {
          var loading = document.createElement('span');
          loading.className = 'sui-tree-loading';
          loading.textContent = 'Loading\u2026';
          row.appendChild(loading);
        }
      }
      var key = item.getAttribute('data-sui-key') || '';
      if (loadKeySetter && typeof window[loadKeySetter] === 'function') window[loadKeySetter](key);
      if (loadHandler && typeof window[loadHandler] === 'function') window[loadHandler]();
      else setChildrenError(key, 'No onLoad handler');
    }
    function setChildrenError(key, message) {
      var item = findItemByKey(key);
      if (!item) return false;
      item.__suiLoading = false;
      item.removeAttribute('data-sui-loading');
      item.removeAttribute('aria-busy');
      var loading = itemRow(item) ? itemRow(item).querySelector('.sui-tree-loading') : null;
      if (loading && loading.parentNode) loading.parentNode.removeChild(loading);
      var row = itemRow(item);
      if (row) {
        var err = row.querySelector('.sui-tree-error');
        if (!err) {
          err = document.createElement('span');
          err.className = 'sui-tree-error';
          err.setAttribute('role', 'alert');
          row.appendChild(err);
        }
        err.textContent = String(message || 'Load failed');
        var retry = row.querySelector('.sui-tree-retry');
        if (!retry) {
          retry = document.createElement('button');
          retry.type = 'button';
          retry.className = 'sui-tree-retry';
          retry.textContent = 'Retry';
          retry.addEventListener('click', function(ev) {
            ev.preventDefault();
            ev.stopPropagation();
            loadLazy(item);
          });
          row.appendChild(retry);
        }
      }
      item.setAttribute('aria-invalid', 'true');
      return true;
    }
    function collectNodeKeys(nodes, set) {
      for (var i = 0; i < nodes.length; i++) {
        var key = nodes[i].key;
        if (set[key]) return false;
        set[key] = true;
        if (nodes[i].children && !collectNodeKeys(nodes[i].children, set)) return false;
      }
      return true;
    }
    function setChildren(key, text) {
      var item = findItemByKey(key);
      if (!item) return { error: 'tree item not found: ' + key };
      var nodes;
      try { nodes = parseJson(String(text || '')); } catch (e) { return { error: (e && e.message) ? e.message : String(e) }; }
      nodes = sortNodes(nodes, sortDir);
      var usedKeys = {};
      for (var u = 0; u < items.length; u++) usedKeys[items[u].getAttribute('data-sui-key') || ''] = true;
      if (!collectNodeKeys(nodes, usedKeys)) return { error: 'duplicate key in loaded children' };
      if (items.length + countNodes(nodes) > MAX_NODES) return { error: 'too many nodes (limit ' + MAX_NODES + ')' };
      var group = ensureGroup(item);
      while (group.firstChild) group.removeChild(group.firstChild);
      var opts = renderOpts();
      for (var i = 0; i < nodes.length; i++) group.appendChild(makeNode(nodes[i], key + '.' + i, parseInt(item.getAttribute('data-sui-level') || '1', 10) + 1, opts));
      item.__suiLoading = false;
      item.removeAttribute('data-sui-loading');
      item.setAttribute('data-sui-lazy', '0');
      item.removeAttribute('aria-invalid');
      item.removeAttribute('aria-busy');
      var loading = itemRow(item) ? itemRow(item).querySelector('.sui-tree-loading') : null;
      if (loading && loading.parentNode) loading.parentNode.removeChild(loading);
      ensureToggle(item);
      setOpen(item, true, true);
      renumberItems();
      syncTreeStore();
      root.__suiTree = false;
      init(root);
      notifyTreeChange();
      return { count: countNodes(nodes) };
    }
    function clearDropClasses() {
      var all = root.querySelectorAll('.sui-tree-drop-before, .sui-tree-drop-after, .sui-tree-drop-inside, .sui-tree-drop-invalid, .sui-tree-dragging');
      for (var i = 0; i < all.length; i++) {
        all[i].classList.remove('sui-tree-drop-before', 'sui-tree-drop-after', 'sui-tree-drop-inside', 'sui-tree-drop-invalid', 'sui-tree-dragging');
      }
    }
    function dropMode(row, ev) {
      var rect = row.getBoundingClientRect();
      var y = ev.clientY - rect.top;
      if (y < rect.height * 0.25) return 'before';
      if (y > rect.height * 0.75) return 'after';
      return 'inside';
    }
    function canDrop(source, target) {
      return !!source && !!target && source !== target && !source.contains(target);
    }
    function syncTreeStore() {
      var name = treeName(root);
      try {
        var text = exportJson(name)[0];
        if (text) store[name] = parseJson(text);
      } catch (e) {}
    }
    function pushUndo() {
      var name = treeName(root);
      try {
        var text = exportJson(name)[0];
        if (!text) return;
        var arr = treeUndo[name] || [];
        arr.push(text);
        while (arr.length > 20) arr.shift();
        treeUndo[name] = arr;
      } catch (e) {}
    }
    function recordLastMove(source) {
      var pg = source.parentElement;
      var parent = parentItem(source, root);
      var index = 0;
      if (pg) {
        for (var i = 0; i < pg.children.length; i++) {
          if (pg.children[i] === source) break;
          if (pg.children[i].hasAttribute && pg.children[i].hasAttribute('data-sui-tree-item')) index++;
        }
      }
      treeLastMove[treeName(root)] = {
        key: source.getAttribute('data-sui-key') || '',
        parent: parent ? (parent.getAttribute('data-sui-key') || '') : '',
        index: index
      };
    }
    function uniqueTreeKey(base) {
      var candidate = base || 'item';
      var n = 2;
      while (findItemByKey(candidate)) candidate = (base || 'item') + '-' + n++;
      return candidate;
    }
    function rekeyClone(source, clone) {
      var sourceKey = source.getAttribute('data-sui-key') || 'item';
      clone.setAttribute('data-sui-key', uniqueTreeKey(sourceKey + '-copy'));
      var sg = itemGroup(source), cg = itemGroup(clone);
      if (sg && cg) {
        for (var i = 0; i < sg.children.length; i++) {
          if (sg.children[i].hasAttribute && sg.children[i].hasAttribute('data-sui-tree-item')) {
            var cc = cg.children[i];
            if (cc && cc.hasAttribute && cc.hasAttribute('data-sui-tree-item')) rekeyClone(sg.children[i], cc);
          }
        }
      }
    }
    function duplicateItem(source) {
      if (!source) return false;
      pushUndo();
      var clone = source.cloneNode(true);
      rekeyClone(source, clone);
      clone.classList.remove('selected', 'sui-tree-dragging', 'sui-tree-pending');
      clone.setAttribute('aria-selected', 'false');
      clone.setAttribute('tabindex', '-1');
      var label = itemLabelEl(clone);
      if (label) {
        label.textContent = label.textContent + ' (copy)';
        label.title = label.textContent;
      }
      var pg = source.parentElement;
      if (!pg) return false;
      pg.insertBefore(clone, source.nextSibling);
      renumberItems();
      syncTreeStore();
      recordLastMove(clone);
      if (moveHandler && typeof window[moveHandler] === 'function') window[moveHandler]();
      root.__suiTree = false;
      init(root);
      notifyTreeChange();
      return true;
    }
    function runtimeSiblings(source) {
      var out = [];
      var pg = source ? source.parentElement : null;
      if (!pg) return out;
      for (var i = 0; i < pg.children.length; i++) {
        var ch = pg.children[i];
        if (ch.hasAttribute && ch.hasAttribute('data-sui-tree-item')) out.push(ch);
      }
      return out;
    }
    function previousSiblingItem(source) {
      var sibs = runtimeSiblings(source);
      var at = sibs.indexOf(source);
      return at > 0 ? sibs[at - 1] : null;
    }
    function nextSiblingItem(source) {
      var sibs = runtimeSiblings(source);
      var at = sibs.indexOf(source);
      return at >= 0 && at < sibs.length - 1 ? sibs[at + 1] : null;
    }
    function refreshAfterEdit(source) {
      renumberItems();
      syncTreeStore();
      recordLastMove(source);
      if (moveHandler && typeof window[moveHandler] === 'function') window[moveHandler]();
      root.__suiTree = false;
      init(root);
      notifyTreeChange();
    }
    function addRuntimeItem(source, asChild) {
      if (!source) return false;
      var pg = asChild ? ensureGroup(source) : source.parentElement;
      if (!pg) return false;
      pushUndo();
      var level = parseInt(source.getAttribute('data-sui-level') || '1', 10) + (asChild ? 1 : 0);
      var node = { key: uniqueTreeKey('item'), label: asChild ? 'New child' : 'New item', open: false };
      var li = makeNode(node, '0', level, renderOpts());
      if (asChild) {
        ensureToggle(source);
        pg.appendChild(li);
        setOpen(source, true, true);
      } else {
        pg.insertBefore(li, source.nextSibling);
      }
      refreshAfterEdit(li);
      return true;
    }
    function moveRuntimeItem(source, delta) {
      var target = delta < 0 ? previousSiblingItem(source) : nextSiblingItem(source);
      if (!source || !target) return false;
      pushUndo();
      var pg = source.parentElement;
      if (delta < 0) pg.insertBefore(source, target);
      else pg.insertBefore(source, target.nextSibling);
      refreshAfterEdit(source);
      return true;
    }
    function indentRuntimeItem(source) {
      var prev = previousSiblingItem(source);
      if (!source || !prev) return false;
      pushUndo();
      var group = ensureGroup(prev);
      ensureToggle(prev);
      group.appendChild(source);
      setOpen(prev, true, true);
      refreshAfterEdit(source);
      return true;
    }
    function outdentRuntimeItem(source) {
      var parent = parentItem(source, root);
      if (!source || !parent || !parent.parentElement) return false;
      pushUndo();
      parent.parentElement.insertBefore(source, parent.nextSibling);
      refreshAfterEdit(source);
      return true;
    }
    function deleteItem(source) {
      if (!source || !source.parentElement) return false;
      pushUndo();
      recordLastMove(source);
      source.parentElement.removeChild(source);
      renumberItems();
      syncTreeStore();
      if (moveHandler && typeof window[moveHandler] === 'function') window[moveHandler]();
      root.__suiTree = false;
      init(root);
      notifyTreeChange();
      return true;
    }
    function undoMove() {
      var name = treeName(root);
      var arr = treeUndo[name] || [];
      if (!arr.length) return { count: 0, error: 'nothing to undo' };
      try {
        var text = arr.pop();
        var nodes = parseJson(text);
        store[name] = nodes;
        root.__suiTree = false;
        init(root);
        notifyTreeChange();
        return { count: countNodes(nodes), error: null };
      } catch (e) {
        return { count: 0, error: (e && e.message) ? e.message : String(e) };
      }
    }
    var draggedItem = null;
    function moveItem(source, target, mode) {
      if (!canDrop(source, target)) return false;
      pushUndo();
      if (mode === 'inside') {
        var group = ensureGroup(target);
        ensureToggle(target);
        group.appendChild(source);
        target.setAttribute('aria-expanded', 'true');
        setOpen(target, true, true);
      } else {
        var parentGroup = target.parentElement;
        if (!parentGroup || !parentGroup.classList || !parentGroup.classList.contains('sui-tree-group')) return false;
        if (mode === 'before') parentGroup.insertBefore(source, target);
        else parentGroup.insertBefore(source, target.nextSibling);
      }
      renumberItems();
      syncTreeStore();
      recordLastMove(source);
      if (moveHandler && typeof window[moveHandler] === 'function') window[moveHandler]();
      root.__suiTree = false;
      init(root);
      notifyTreeChange();
      return true;
    }
    function closeTreeMenu() {
      var old = document.getElementById('sui-tree-context');
      if (old) old.remove();
      if (contextMenuBound) {
        document.removeEventListener('click', closeTreeMenu);
        document.removeEventListener('scroll', closeTreeMenu, true);
        contextMenuBound = false;
      }
    }
    function openTreeMenu(ev, item) {
      closeTreeMenu();
      var menuKey = item.getAttribute('data-sui-key') || '';
      function target() { return freshItemByKey(menuKey) || item; }
      function ownerApi() {
        var el = target();
        var owner = el && el.closest ? el.closest('.sui-tree') : null;
        return owner && owner.__suiTreeApi ? owner.__suiTreeApi : null;
      }
      var menu = document.createElement('div');
      menu.id = 'sui-tree-context';
      menu.className = 'sui-tree-context';
      menu.setAttribute('role', 'menu');
      menu.setAttribute('aria-label', 'Tree actions');
      function add(label, fn, disabled, act) {
        var b = document.createElement('button');
        b.type = 'button';
        b.className = 'sui-tree-context-item';
        b.setAttribute('role', 'menuitem');
        b.textContent = label;
        b.disabled = !!disabled;
        if (act) b.setAttribute('data-act', act);
        b.addEventListener('click', function(e) { e.stopPropagation(); closeTreeMenu(); fn(); });
        menu.appendChild(b);
      }
      var initial = target();
      if (initial.__suiEditable) add('Rename', function() { startRename(target()); });
      add('Add child item', function() { addRuntimeItem(target(), true); });
      add('Add item below', function() { addRuntimeItem(target(), false); });
      add('Move up', function() { moveRuntimeItem(target(), -1); }, !previousSiblingItem(initial), 'move-up');
      add('Move down', function() { moveRuntimeItem(target(), 1); }, !nextSiblingItem(initial), 'move-down');
      add('Indent', function() { indentRuntimeItem(target()); }, !previousSiblingItem(initial), 'indent');
      add('Outdent', function() { outdentRuntimeItem(target()); }, !parentItem(initial, root), 'outdent');
      if (itemGroup(initial)) {
        add(initial.__suiTreeOpen ? 'Collapse subtree' : 'Expand subtree', function() {
          var t = target();
          setOpen(t, !t.__suiTreeOpen, true);
        }, false, 'toggle-subtree');
      } else if (initial.getAttribute('data-sui-lazy') === '1') {
        add('Load children', function() { loadLazy(target()); });
      }
      add('Copy key', function() {
        var key = target().getAttribute('data-sui-key') || '';
        if (navigator.clipboard && navigator.clipboard.writeText) {
          var p = navigator.clipboard.writeText(key);
          if (p && p.catch) p.catch(function() {});
        }
      });
      add('Duplicate item', function() {
        var api = ownerApi();
        if (api && api.duplicateItem) api.duplicateItem(menuKey);
        else duplicateItem(target());
      });
      add('Delete item', function() {
        var api = ownerApi();
        if (api && api.deleteItem) api.deleteItem(menuKey);
        else deleteItem(target());
      });
      add('Expand all', function() { expandAll(); });
      add('Collapse all', function() { collapseAll(); });
      document.body.appendChild(menu);
      var x = ev.clientX, y = ev.clientY;
      var rect = menu.getBoundingClientRect();
      if (x + rect.width > window.innerWidth - 8) x = Math.max(8, window.innerWidth - rect.width - 8);
      if (y + rect.height > window.innerHeight - 8) y = Math.max(8, window.innerHeight - rect.height - 8);
      menu.style.left = x + 'px';
      menu.style.top = y + 'px';
      var buttons = menu.querySelectorAll('button');
      if (buttons.length) buttons[0].focus();
      menu.addEventListener('keydown', function(e) {
        var list = Array.prototype.slice.call(menu.querySelectorAll('button'));
        var at = list.indexOf(document.activeElement);
        if (e.key === 'Escape') { e.preventDefault(); closeTreeMenu(); focusItem(item); }
        else if (e.key === 'ArrowDown') { e.preventDefault(); if (list.length) list[Math.min(list.length - 1, at + 1)].focus(); }
        else if (e.key === 'ArrowUp') { e.preventDefault(); if (list.length) list[Math.max(0, at - 1)].focus(); }
      });
      setTimeout(function() {
        if (!menu.parentNode) return;
        // Reactive re-renders can replace the clicked node between selection
        // and menu creation; recompute the structural availability from the
        // live target so Move/Indent/Outdent never reflect a stale neighbor.
        var t = target();
        function setAct(act, disabled) {
          var b = menu.querySelector('[data-act="' + act + '"]');
          if (b) b.disabled = !!disabled;
        }
        setAct('move-up', !previousSiblingItem(t));
        setAct('move-down', !nextSiblingItem(t));
        setAct('indent', !previousSiblingItem(t));
        setAct('outdent', !parentItem(t, root));
        var tg = menu.querySelector('[data-act="toggle-subtree"]');
        if (tg) {
          if (itemGroup(t)) { tg.textContent = t.__suiTreeOpen ? 'Collapse subtree' : 'Expand subtree'; tg.style.display = ''; }
          else tg.style.display = 'none';
        }
      }, 0);
      setTimeout(function() {
        if (menu.parentNode) {
          document.addEventListener('click', closeTreeMenu);
          document.addEventListener('scroll', closeTreeMenu, true);
          contextMenuBound = true;
        }
      }, 0);
    }
    if (customTooltips && !root.__suiTooltipBound) {
      root.__suiTooltipBound = true;
      var hideTipOnScroll = function() { hideTooltip(); };
      root.addEventListener('scroll', hideTipOnScroll, true);
      window.addEventListener('scroll', hideTipOnScroll, true);
    }
    if (searchEnabled) {
      var searchInput = root.querySelector('.sui-tree-search-input');
      if (searchInput) {
        var storedSearch = treeSearchTerms[treeName(root)] || '';
        if (storedSearch && !searchInput.value) searchInput.value = storedSearch;
        searchInput.addEventListener('input', function() { applySearch(searchInput.value); });
        searchInput.addEventListener('keydown', function(ev) {
          if (ev.key === 'Escape') { ev.preventDefault(); searchInput.value = ''; applySearch(''); }
        });
      }
    }
    function renameIconFor(item) {
      var itemShow = item.getAttribute('data-sui-show-rename-icon') || 'inherit';
      var show = itemShow === '1' ? true : itemShow === '0' ? false : showTreeRenameIcon;
      var icon = item.getAttribute('data-sui-rename-icon') || treeRenameIcon;
      return show ? icon : '';
    }
    items.forEach(function(item) {
      if (item.__suiBound) return;
      item.__suiBound = true;
      var row = itemRow(item);
      if (row && item.hasAttribute('style')) {
        // Tree-item style expr belongs to this row only. Moving it off the
        // <li> prevents custom properties from leaking to nested children.
        row.setAttribute('style', item.getAttribute('style'));
        item.removeAttribute('style');
      }
      var labelEl = itemLabelEl(item);
      var stored = storedLabel(root, item.getAttribute('data-sui-path') || '');
      if (labelEl && stored !== null) labelEl.textContent = stored;
      if (labelEl) {
        item.__suiLabel = labelEl.textContent;
        var tip = item.getAttribute('data-sui-tooltip') || '';
        if (customTooltips) {
          labelEl.removeAttribute('title');
          item.setAttribute('aria-describedby', 'sui-tree-tip-' + treeKey.replace(/[^A-Za-z0-9_-]/g, '-'));
        } else {
          labelEl.title = tip || labelEl.textContent;
        }
      }
      item.__suiEditable = editableFor(item);
      item.classList.toggle('sui-tree-editable', item.__suiEditable);
      item.setAttribute('aria-readonly', item.__suiEditable ? 'false' : 'true');
      if (checkboxes) {
        var checkBox = itemCheck(item);
        if (checkBox) {
          checkBox.addEventListener('mousedown', function(ev) { ev.stopPropagation(); });
          checkBox.addEventListener('click', function(ev) {
            ev.preventDefault();
            ev.stopPropagation();
            toggleCheck(item);
          });
        }
      }
      var group = itemGroup(item);
      if (group) {
        var path = item.getAttribute('data-sui-path') || '';
        var open = storedOpen(root, path);
        if (open === null) open = item.getAttribute('data-sui-open') === '1';
        setOpen(item, open, false);
        var toggle = itemToggle(item);
        if (toggle) toggle.addEventListener('click', function(ev) {
          ev.preventDefault();
          ev.stopPropagation();
          setOpen(item, !item.__suiTreeOpen, true);
        });
      } else if (item.getAttribute('data-sui-lazy') === '1') {
        var lazyToggle = itemToggle(item);
        if (lazyToggle) lazyToggle.addEventListener('click', function(ev) {
          ev.preventDefault();
          ev.stopPropagation();
          loadLazy(item);
        });
      }
      if (row) row.addEventListener('mousedown', function(ev) {
        if (ev.target && ev.target.closest && ev.target.closest('.sui-tree-toggle')) return;
        if (ev.target && ev.target.closest && ev.target.closest('.sui-tree-rename-btn')) return;
        var now = Date.now();
        var key = item.getAttribute('data-sui-key') || '';
        if (lastDown.tree === treeKey && lastDown.key === key && now - lastDown.time < 600) {
          item.__suiDouble = true;
          lastDown.key = '';
        } else {
          lastDown.tree = treeKey;
          lastDown.key = key;
          lastDown.time = now;
        }
      });
      if (row && customTooltips) {
        row.addEventListener('mouseenter', function() { showTooltip(item, itemLabelEl(item)); });
        row.addEventListener('mouseleave', hideTooltip);
      }
      if (customTooltips) {
        item.addEventListener('focusin', function() { showTooltip(item, itemLabelEl(item), true); });
        item.addEventListener('focusout', hideTooltip);
      }
      if (row) row.addEventListener('click', function(ev) {
        if (ev.target && ev.target.closest && ev.target.closest('.sui-tree-toggle')) return;
        if (ev.target && ev.target.closest && ev.target.closest('.sui-tree-check')) return;
        var mods = { ctrl: ev.ctrlKey || ev.metaKey, shift: ev.shiftKey };
        if (selectionTimer) { clearTimeout(selectionTimer); selectionTimer = null; }
        if (pendingItem) { pendingItem.classList.remove('sui-tree-pending'); pendingItem = null; }
        if (item.__suiDouble && item.__suiEditable) {
          item.__suiDouble = false;
          ev.preventDefault();
          focusItem(item);
          setTimeout(function() { if (item.isConnected) startRename(item); }, 0);
          return;
        }
        item.__suiDouble = false;
        if (item.__suiEditable && !multiSelect) {
          // Keep the row in the DOM briefly so a double-click can start rename
          // before the reactive re-render replaces it. The pending class gives
          // immediate visual feedback for a single click.
          item.classList.add('sui-tree-pending');
          pendingItem = item;
          selectionTimer = setTimeout(function() {
            selectionTimer = null;
            if (pendingItem) pendingItem.classList.remove('sui-tree-pending');
            pendingItem = null;
            select(item, true, mods);
            focusItem(item);
          }, 500);
        } else {
          select(item, true, mods);
          focusItem(item);
        }
      });
      if (row) row.addEventListener('dblclick', function(ev) {
        if (ev.target && ev.target.closest && ev.target.closest('.sui-tree-toggle')) return;
        if (item.__suiEditable) {
          if (selectionTimer) { clearTimeout(selectionTimer); selectionTimer = null; }
          if (pendingItem) { pendingItem.classList.remove('sui-tree-pending'); pendingItem = null; }
          ev.preventDefault();
          focusItem(item);
          setTimeout(function() { if (item.isConnected) startRename(item); }, 0);
        }
      });
      if (row && (contextMenuEnabled || contextHandler)) {
        row.addEventListener('contextmenu', function(ev) {
          ev.preventDefault();
          var contextKey = item.getAttribute('data-sui-key') || '';
          select(item, true);
          focusItem(item);
          setTimeout(function() {
            var fresh = freshItemByKey(contextKey) || item;
            if (contextHandler && typeof window[contextHandler] === 'function') window[contextHandler]();
            if (contextMenuEnabled) openTreeMenu(ev, fresh);
          }, 0);
        });
      }
      if (row && dragEnabled) {
        row.draggable = true;
        row.addEventListener('dragstart', function(ev) {
          draggedItem = item;
          item.classList.add('sui-tree-dragging');
          if (ev.dataTransfer) {
            ev.dataTransfer.effectAllowed = 'move';
            try { ev.dataTransfer.setData('text/plain', item.getAttribute('data-sui-key') || ''); } catch (e) {}
          }
        });
        row.addEventListener('dragend', function() {
          draggedItem = null;
          clearDropClasses();
        });
        row.addEventListener('dragover', function(ev) {
          if (!draggedItem) return;
          ev.preventDefault();
          clearDropClasses();
          draggedItem.classList.add('sui-tree-dragging');
          if (!canDrop(draggedItem, item)) {
            if (ev.dataTransfer) ev.dataTransfer.dropEffect = 'none';
            row.parentElement.classList.add('sui-tree-drop-invalid');
            return;
          }
          var mode = dropMode(row, ev);
          if (ev.dataTransfer) ev.dataTransfer.dropEffect = 'move';
          row.parentElement.classList.add(mode === 'before' ? 'sui-tree-drop-before' : mode === 'after' ? 'sui-tree-drop-after' : 'sui-tree-drop-inside');
        });
        row.addEventListener('dragleave', function(ev) {
          if (!row.contains(ev.relatedTarget) && item.parentElement) {
            item.parentElement.classList.remove('sui-tree-drop-before', 'sui-tree-drop-after', 'sui-tree-drop-inside', 'sui-tree-drop-invalid');
          }
        });
        row.addEventListener('drop', function(ev) {
          if (!draggedItem) return;
          ev.preventDefault();
          ev.stopPropagation();
          var mode = dropMode(row, ev);
          var source = draggedItem;
          draggedItem = null;
          clearDropClasses();
          moveItem(source, item, mode);
        });
      }
      var editIcon = renameIconFor(item);
      if (row && item.__suiEditable && editIcon) {
        var editBtn = document.createElement('button');
        editBtn.type = 'button';
        editBtn.className = 'sui-tree-rename-btn';
        editBtn.tabIndex = -1;
        editBtn.textContent = editIcon;
        editBtn.setAttribute('aria-label', 'Rename ' + itemLabel(item));
        editBtn.addEventListener('mousedown', function(ev) { ev.stopPropagation(); });
        editBtn.addEventListener('click', function(ev) {
          ev.preventDefault();
          ev.stopPropagation();
          focusItem(item);
          setTimeout(function() { if (item.isConnected) startRename(item); }, 0);
        });
        var renameAnchor = labelEl || itemLabelEl(item);
        if (renameAnchor && renameAnchor.nextSibling && renameAnchor.nextSibling.classList && renameAnchor.nextSibling.classList.contains('sui-tree-badge-count')) {
          renameAnchor = renameAnchor.nextSibling;
        }
        if (renameAnchor && renameAnchor.parentNode) row.insertBefore(editBtn, renameAnchor.nextSibling);
        else row.appendChild(editBtn);
      }
      item.addEventListener('focus', function() {
        items.forEach(function(x) { x.tabIndex = x === item ? 0 : -1; });
        if (customTooltips) showTooltip(item, itemLabelEl(item), true);
      });
      item.addEventListener('blur', function() {
        if (customTooltips) hideTooltip();
      });
      item.addEventListener('keydown', function(ev) {
        var vis = items.filter(visible);
        var idx = vis.indexOf(item);
        if (searchEnabled && (ev.ctrlKey || ev.metaKey) && (ev.key === 'f' || ev.key === 'F')) {
          ev.preventDefault();
          focusSearch();
          return;
        }
        if (multiSelect && (ev.ctrlKey || ev.metaKey) && (ev.key === 'a' || ev.key === 'A')) {
          ev.preventDefault();
          setSelection(vis, true);
          selectionAnchor = vis.length ? vis[0] : null;
          return;
        }
        if (ev.key === 'F2' && item.__suiEditable) {
          ev.preventDefault();
          startRename(item);
        } else if (ev.key === 'ArrowDown') {
          ev.preventDefault();
          if (idx >= 0 && idx + 1 < vis.length) {
            focusItem(vis[idx + 1]);
            if (multiSelect && ev.shiftKey) select(vis[idx + 1], true, { shift: true });
          }
        } else if (ev.key === 'ArrowUp') {
          ev.preventDefault();
          if (idx > 0) {
            focusItem(vis[idx - 1]);
            if (multiSelect && ev.shiftKey) select(vis[idx - 1], true, { shift: true });
          }
        } else if (ev.key === 'Home') {
          ev.preventDefault();
          if (vis.length) {
            focusItem(vis[0]);
            if (multiSelect && ev.shiftKey) select(vis[0], true, { shift: true });
          }
        } else if (ev.key === 'End') {
          ev.preventDefault();
          if (vis.length) {
            var lastVis = vis[vis.length - 1];
            focusItem(lastVis);
            if (multiSelect && ev.shiftKey) select(lastVis, true, { shift: true });
          }
        } else if (ev.key === 'ArrowRight') {
          ev.preventDefault();
          var group = itemGroup(item);
          if (!group && item.getAttribute('data-sui-lazy') === '1' && !item.__suiLoading) {
            loadLazy(item);
          } else if (group && !item.__suiTreeOpen) setOpen(item, true, true);
          else if (group && item.__suiTreeOpen) {
            var first = null;
            for (var i = 0; i < group.children.length; i++) {
              var ch = group.children[i];
              if (ch.hasAttribute && ch.hasAttribute('data-sui-tree-item')) { first = ch; break; }
            }
            if (first) focusItem(first);
          }
        } else if (ev.key === 'ArrowLeft') {
          ev.preventDefault();
          var parent = parentItem(item, root);
          if (item.__suiTreeOpen && itemGroup(item)) setOpen(item, false, true);
          else if (parent) focusItem(parent);
        } else if (ev.key === 'Enter' || ev.key === ' ') {
          ev.preventDefault();
          if (ev.key === ' ' && checkboxes) toggleCheck(item);
          else if (ev.key === ' ' && multiSelect && (ev.ctrlKey || ev.metaKey)) select(item, true, { ctrl: true });
          else select(item, true);
        }
      });
    });
    initChecks();
    var initialSelection = [];
    if (multiSelect && selectionHidden && selectionHidden.value) {
      try {
        var parsedSelection = JSON.parse(selectionHidden.value);
        if (Array.isArray(parsedSelection)) {
          for (var s = 0; s < parsedSelection.length; s++) {
            for (var k = 0; k < items.length; k++) {
              if ((items[k].getAttribute('data-sui-key') || '') === String(parsedSelection[s])) {
                initialSelection.push(items[k]);
                break;
              }
            }
          }
        }
      } catch (e) {}
    }
    if (!initialSelection.length && lastSelection.tree === treeKey && lastSelection.keys.length) {
      for (var g = 0; g < lastSelection.keys.length; g++) {
        var gk = findItemByKey(lastSelection.keys[g]);
        if (gk) initialSelection.push(gk);
      }
    }
    if (!initialSelection.length && storedSelection.length) {
      for (var sp = 0; sp < storedSelection.length; sp++) {
        var spk = findItemByKey(storedSelection[sp]);
        if (spk) initialSelection.push(spk);
      }
    }
    if (!initialSelection.length && current) {
      for (var i = 0; i < items.length; i++) {
        if ((items[i].getAttribute('data-sui-key') || '') === current) {
          initialSelection.push(items[i]);
          break;
        }
      }
    }
    if (initialSelection.length) {
      setSelection(initialSelection, false, false);
      var anchorItem = lastSelection.tree === treeKey ? findItemByKey(lastSelection.key) : null;
      selectionAnchor = anchorItem || initialSelection[0];
    } else {
      // No selected value: keep the first visible row focusable without
      // selecting it or writing a synthetic value into the bound state.
      for (var i2 = 0; i2 < items.length; i2++) {
        if (visible(items[i2])) { items[i2].tabIndex = 0; break; }
      }
    }
    if (searchEnabled) {
      var searchNow = root.querySelector('.sui-tree-search-input');
      if (searchNow && searchNow.value) applySearch(searchNow.value);
    }
    applyRuntimeCells();
    applyRuntimeColumns();
    root.__suiTreeApi = {
      expandAll: expandAll,
      collapseAll: collapseAll,
      expandTo: expandTo,
      setChildren: setChildren,
      setChildrenError: setChildrenError,
      retryChildren: function(key) {
        var item = findItemByKey(key);
        if (!item) return false;
        loadLazy(item);
        return true;
      },
      duplicateItem: function(key) {
        var item = freshItemByKey(key);
        return item ? duplicateItem(item) : false;
      },
      deleteItem: function(key) {
        var item = freshItemByKey(key);
        return item ? deleteItem(item) : false;
      },
      lastMove: function() { return treeLastMove[treeName(root)] || null; },
      undoMove: undoMove,
      setColumns: function(list) { return setColumnsFor(list); },
      getColumns: function() { return getColumnsFor(); },
      moveColumn: function(from, to) {
        var list = getColumnsFor();
        from = parseInt(from, 10); to = parseInt(to, 10);
        if (!isFinite(from) || !isFinite(to) || from < 0 || from >= list.length || to < 0 || to >= list.length) return false;
        var moved = list.splice(from, 1)[0];
        list.splice(to, 0, moved);
        setColumnsFor(list);
        return true;
      },
      setCell: function(key, column, value) { return setCellFor(key, column, value); },
      selectedKeys: selectedKeys,
      checkedKeys: checkedKeys,
      notifyTreeChange: notifyTreeChange,
      search: applySearch,
      focusSearch: focusSearch
    };
  }
  // ---- JSON/XML data import and export -----------------------------------
  function cleanLabel(v) {
    var s = treeText(v).trim();
    if (s.length > MAX_LABEL) throw new Error("label is too long (limit " + MAX_LABEL + ")");
    return s;
  }
  function boolValue(v, fallback) {
    if (v === undefined || v === null || v === "") return fallback;
    if (typeof v === "boolean") return v;
    var s = treeText(v).trim().toLowerCase();
    if (s === "true" || s === "1" || s === "yes") return true;
    if (s === "false" || s === "0" || s === "no") return false;
    throw new Error("invalid boolean: " + s);
  }
  function triValue(v, name) {
    if (v === undefined || v === null || v === "") return "";
    return boolValue(v, false) ? "1" : "0";
  }
  function rawLabel(raw) {
    if (!raw || typeof raw !== "object" || Array.isArray(raw)) return treeText(raw);
    var keys = ["label", "text", "name", "title"];
    for (var i = 0; i < keys.length; i++) {
      if (raw[keys[i]] !== undefined && raw[keys[i]] !== null && raw[keys[i]] !== "") return treeText(raw[keys[i]]);
    }
    return "";
  }
  function normalizeNodes(raw, prefix, depth, ctx) {
    if (!Array.isArray(raw)) throw new Error("children must be an array");
    if (depth > MAX_DEPTH) throw new Error("nesting is too deep (limit " + MAX_DEPTH + ")");
    var out = [];
    for (var i = 0; i < raw.length; i++) {
      out.push(normalizeNode(raw[i], prefix ? prefix + "." + i : String(i), depth, ctx));
    }
    return out;
  }
  function normalizeNode(raw, path, depth, ctx) {
    if (ctx.count >= MAX_NODES) throw new Error("too many nodes (limit " + MAX_NODES + ")");
    if (depth > MAX_DEPTH) throw new Error("nesting is too deep (limit " + MAX_DEPTH + ")");
    ctx.count++;
    var obj = raw && typeof raw === "object" && !Array.isArray(raw) ? raw : null;
    var key = obj && obj.key !== undefined && obj.key !== null ? treeText(obj.key).trim() : "";
    if (!key) key = path;
    if (key.length > MAX_KEY) throw new Error("key is too long (limit " + MAX_KEY + ")");
    if (Object.prototype.hasOwnProperty.call(ctx.keys, key)) throw new Error("duplicate key: " + key);
    ctx.keys[key] = true;
    var label = cleanLabel(rawLabel(raw));
    if (!label) label = key;
    var node = { label: label, key: key, children: [] };
    if (obj) {
      var kids = obj.children !== undefined ? obj.children : obj.items;
      if (kids !== undefined && kids !== null && !Array.isArray(kids)) throw new Error("children must be an array");
      if (kids) node.children = normalizeNodes(kids, path, depth + 1, ctx);
      if (obj.open !== undefined) node.open = boolValue(obj.open, false);
      if (obj.editable !== undefined) node.editable = triValue(obj.editable, "editable");
      if (obj.renameIcon !== undefined) {
        var icon = treeText(obj.renameIcon);
        if (icon.length > MAX_ICON) throw new Error("renameIcon is too long (limit " + MAX_ICON + ")");
        node.renameIcon = icon;
      }
      if (obj.showRenameIcon !== undefined) node.showRenameIcon = triValue(obj.showRenameIcon, "showRenameIcon");
      if (obj.checked !== undefined) node.checked = boolValue(obj.checked, false);
      if (obj.icon !== undefined) {
        var iconText = treeText(obj.icon);
        if (iconText.length > MAX_ICON) throw new Error("icon is too long (limit " + MAX_ICON + ")");
        node.icon = iconText;
      }
      if (obj.iconOpen !== undefined) {
        var iconOpenText = treeText(obj.iconOpen);
        if (iconOpenText.length > MAX_ICON) throw new Error("iconOpen is too long (limit " + MAX_ICON + ")");
        node.iconOpen = iconOpenText;
      }
      if (obj.badge !== undefined) {
        var badgeText = treeText(obj.badge);
        if (badgeText.length > MAX_BADGE) throw new Error("badge is too long (limit " + MAX_BADGE + ")");
        node.badge = badgeText;
      }
      if (obj.badgeType !== undefined) {
        var badgeType = treeText(obj.badgeType).toLowerCase();
        if (badgeType !== "" && ["neutral", "info", "success", "warning", "danger"].indexOf(badgeType) < 0) {
          throw new Error("badgeType must be neutral, info, success, warning or danger");
        }
        node.badgeType = badgeType;
      }
      if (obj.meta !== undefined) {
        var metaText = treeText(obj.meta);
        if (metaText.length > MAX_META) throw new Error("meta is too long (limit " + MAX_META + ")");
        node.meta = metaText;
      }
      if (obj.lazy !== undefined) node.lazy = boolValue(obj.lazy, false);
      if (obj.size !== undefined) node.size = treeText(obj.size);
      if (obj.date !== undefined) node.date = treeText(obj.date);
      if (obj.status !== undefined) node.status = treeText(obj.status);
      if (obj.cells !== undefined) {
        if (typeof obj.cells === "string") {
          node.cells = parseCellsText(obj.cells);
        } else {
          if (!obj.cells || typeof obj.cells !== "object" || Array.isArray(obj.cells)) throw new Error("cells must be an object or key=value text");
          var cleanCells = {};
          Object.keys(obj.cells).forEach(function(k){
            var v = obj.cells[k];
            if (v !== null && typeof v === "object") return;
            cleanCells[k] = treeText(v);
          });
          node.cells = cleanCells;
        }
      }
      if (obj.tooltip !== undefined) {
        var tooltipText = treeText(obj.tooltip);
        if (tooltipText.length > MAX_LABEL) throw new Error("tooltip is too long (limit " + MAX_LABEL + ")");
        node.tooltip = tooltipText;
      }
    } else if (raw !== null && raw !== undefined && typeof raw === "object") {
      throw new Error("node must be a string or an object");
    }
    return node;
  }
  function jsonRoot(data) {
    if (Array.isArray(data)) return data;
    if (data && typeof data === "object") {
      if (Array.isArray(data.nodes)) return data.nodes;
      if (Array.isArray(data.items)) return data.items;
      if (Array.isArray(data.tree)) return data.tree;
      if (Array.isArray(data.children)) {
        if (rawLabel(data)) return [data];
        return data.children;
      }
      if (rawLabel(data)) return [data];
    }
    throw new Error("expected an array of nodes or an object with nodes/children");
  }
  function parseJson(text) {
    if (typeof text !== "string" || !text.trim()) throw new Error("empty JSON");
    if (text.length > ((typeof __SUI_SEC !== "undefined" && __SUI_SEC.maxText) || 20000000)) throw new Error("JSON is too large");
    var data;
    try { data = JSON.parse(text); } catch (e) { throw new Error("invalid JSON: " + e.message); }
    var ctx = { count: 0, keys: Object.create(null) };
    return normalizeNodes(jsonRoot(data), "", 0, ctx);
  }
  function elementKids(el) {
    var out = [];
    if (!el) return out;
    for (var i = 0; i < el.childNodes.length; i++) {
      if (el.childNodes[i].nodeType === 1) out.push(el.childNodes[i]);
    }
    return out;
  }
  function directText(el) {
    var s = "";
    for (var i = 0; i < el.childNodes.length; i++) {
      if (el.childNodes[i].nodeType === 3) s += el.childNodes[i].nodeValue || "";
    }
    return s.trim();
  }
  function xmlRaw(el) {
    var o = {};
    var label = el.getAttribute("label") || el.getAttribute("text") || el.getAttribute("name") || el.getAttribute("title") || directText(el);
    if (label) o.label = label;
    var key = el.getAttribute("key");
    if (key) o.key = key;
    var attrs = ["open", "editable", "showRenameIcon", "icon", "iconOpen", "badge", "badgeType", "meta", "checked", "tooltip", "lazy", "size", "date", "status"];
    for (var i = 0; i < attrs.length; i++) {
      var v = el.getAttribute(attrs[i]);
      if (v !== null) o[attrs[i]] = v;
    }
    var cellsAttr = el.getAttribute("cells");
    if (cellsAttr !== null) {
      try {
        var cellsObj = JSON.parse(cellsAttr);
        if (cellsObj && typeof cellsObj === "object" && !Array.isArray(cellsObj)) o.cells = cellsObj;
      } catch (e) {}
    }
    var icon = el.getAttribute("renameIcon");
    if (icon !== null) o.renameIcon = icon;
    var kids = elementKids(el);
    if (kids.length) {
      o.children = [];
      for (var j = 0; j < kids.length; j++) o.children.push(xmlRaw(kids[j]));
    }
    return o;
  }
  function parseXml(text) {
    if (typeof text !== "string" || !text.trim()) throw new Error("empty XML");
    if (text.length > ((typeof __SUI_SEC !== "undefined" && __SUI_SEC.maxText) || 20000000)) throw new Error("XML is too large");
    if (/<!DOCTYPE|<!ENTITY/i.test(text)) throw new Error("DTD and entities are not allowed");
    if (!window.DOMParser) throw new Error("XML parser is not available");
    var doc = new DOMParser().parseFromString(text, "application/xml");
    var pe = doc.getElementsByTagName("parsererror");
    if (pe.length) throw new Error("invalid XML: " + treeText(pe[0].textContent).trim());
    var root = doc.documentElement;
    if (!root) throw new Error("empty XML document");
    var raw;
    var tag = (root.localName || root.nodeName || "").toLowerCase();
    if (tag === "tree") {
      var kids = elementKids(root);
      raw = [];
      for (var i = 0; i < kids.length; i++) raw.push(xmlRaw(kids[i]));
    } else {
      raw = [xmlRaw(root)];
    }
    var ctx = { count: 0, keys: Object.create(null) };
    return normalizeNodes(raw, "", 0, ctx);
  }
  function makeNode(node, path, level, opts) {
    opts = opts || {};
    var li = document.createElement("li");
    li.className = "sui-tree-item";
    li.setAttribute("data-sui-tree-item", "1");
    li.setAttribute("data-sui-key", node.key);
    li.setAttribute("data-sui-path", path);
    li.setAttribute("data-sui-level", String(level));
    li.setAttribute("data-sui-editable", node.editable || "inherit");
    li.setAttribute("data-sui-show-rename-icon", node.showRenameIcon || "inherit");
    li.setAttribute("data-sui-rename-icon", node.renameIcon || "");
    li.setAttribute("data-sui-open", node.open ? "1" : "0");
    li.setAttribute("data-sui-icon-base", node.icon || "");
    li.setAttribute("data-sui-icon-open-base", node.iconOpen || "");
    li.setAttribute("data-sui-icon", node.icon || "");
    li.setAttribute("data-sui-badge", node.badge || "");
    li.setAttribute("data-sui-badge-type", node.badgeType || "neutral");
    li.setAttribute("data-sui-meta", node.meta || "");
    li.setAttribute("data-sui-size", node.size || "");
    li.setAttribute("data-sui-date", node.date || "");
    li.setAttribute("data-sui-status", node.status || "");
    if (node.cells) li.setAttribute("data-sui-cells", JSON.stringify(node.cells));
    li.setAttribute("data-sui-lazy", node.lazy ? "1" : "0");
    li.setAttribute("data-sui-checked", node.checked === true ? "1" : node.checked === false ? "0" : "");
    if (node.tooltip) li.setAttribute("data-sui-tooltip", node.tooltip);
    li.setAttribute("role", "treeitem");
    li.setAttribute("aria-level", String(level));
    li.setAttribute("aria-selected", "false");
    li.setAttribute("tabindex", "-1");
    var row = document.createElement("div");
    row.className = "sui-tree-row";
    if (opts.guides) {
      for (var g = 1; g < level; g++) {
        var guide = document.createElement("span");
        guide.className = "sui-tree-guide";
        guide.setAttribute("aria-hidden", "true");
        row.appendChild(guide);
      }
    }
    var runtimeParent = (node.children && node.children.length) || node.lazy;
    if (runtimeParent) {
      var toggle = document.createElement("button");
      toggle.type = "button";
      toggle.className = "sui-tree-toggle";
      toggle.tabIndex = -1;
      toggle.setAttribute("aria-label", "Toggle item");
      toggle.textContent = node.open ? (opts.toggleOpen || "\u25be") : (opts.toggleClosed || "\u25b8");
      row.appendChild(toggle);
      li.setAttribute("aria-expanded", node.open ? "true" : "false");
    } else {
      var spacer = document.createElement("span");
      spacer.className = "sui-tree-spacer";
      row.appendChild(spacer);
    }
    if (opts.checkbox) {
      var check = document.createElement("span");
      check.className = "sui-tree-check";
      check.setAttribute("role", "checkbox");
      check.setAttribute("aria-checked", node.checked === true ? "true" : "false");
      check.setAttribute("tabindex", "-1");
      check.setAttribute("aria-label", "Toggle " + node.label);
      row.appendChild(check);
      li.setAttribute("aria-checked", node.checked === true ? "true" : "false");
    }
    var icon = node.icon || (runtimeParent ? (opts.iconParent || opts.icon) : (opts.iconLeaf || opts.icon)) || "";
    if (node.open && runtimeParent) {
      if (node.iconOpen) icon = node.iconOpen;
      else if (opts.iconParentOpen) icon = opts.iconParentOpen;
      else if (opts.iconOpen) icon = opts.iconOpen;
    }
    if (icon) {
      var iconEl = document.createElement("span");
      iconEl.className = "sui-tree-icon";
      iconEl.setAttribute("aria-hidden", "true");
      iconEl.textContent = icon;
      row.appendChild(iconEl);
    }
    var label = document.createElement("span");
    label.className = "sui-tree-label";
    if (opts.fontStep > 0 && level > 1) {
      label.style.fontSize = "calc(var(--sui-tree-font, 13px) - " + ((level - 1) * opts.fontStep) + "px)";
    }
    label.textContent = node.label;
    label.title = node.tooltip || node.label;
    row.appendChild(label);
    if (opts.counts && node.children && node.children.length) {
      var countBadge = document.createElement("span");
      countBadge.className = "sui-tree-badge sui-tree-badge-count";
      countBadge.textContent = String(node.children.length);
      row.appendChild(countBadge);
    }
    if (node.badge) {
      var badge = document.createElement("span");
      badge.className = "sui-tree-badge sui-tree-badge-" + (node.badgeType || "neutral");
      badge.textContent = node.badge;
      row.appendChild(badge);
    }
    if (node.meta) {
      var meta = document.createElement("span");
      meta.className = "sui-tree-meta";
      meta.textContent = node.meta;
      row.appendChild(meta);
    }
    if (opts.columns && opts.columns.length) {
      for (var c = 0; c < opts.columns.length; c++) {
        var col = opts.columns[c];
        var cell = document.createElement("span");
        cell.className = "sui-tree-cell sui-tree-cell-" + col;
        cell.textContent = node[col] !== undefined ? node[col] : ((node.cells && node.cells[col] !== undefined) ? node.cells[col] : "");
        if (opts.columnWidths && opts.columnWidths[c] > 0) cell.style.width = opts.columnWidths[c] + "px";
        row.appendChild(cell);
      }
    }
    li.appendChild(row);
    if (node.children && node.children.length) {
      var group = document.createElement("ul");
      group.className = "sui-tree-group";
      group.setAttribute("role", "group");
      if (!node.open) group.hidden = true;
      for (var i = 0; i < node.children.length; i++) {
        group.appendChild(makeNode(node.children[i], path + "." + i, level + 1, opts));
      }
      li.appendChild(group);
    }
    return li;
  }
  function rootGroup(root) {
    for (var i = 0; i < root.children.length; i++) {
      if (root.children[i].classList && root.children[i].classList.contains("sui-tree-group")) return root.children[i];
    }
    var group = document.createElement("ul");
    group.className = "sui-tree-group";
    group.setAttribute("role", "group");
    root.appendChild(group);
    return group;
  }
  function renderNodes(root, nodes, opts) {
    opts = opts || {};
    var group = rootGroup(root);
    while (group.firstChild) group.removeChild(group.firstChild);
    var sorted = sortNodes(nodes, opts.sort);
    for (var i = 0; i < sorted.length; i++) group.appendChild(makeNode(sorted[i], String(i), 1, opts));
  }
  function clearTreeStorage(root) {
    var key = root.getAttribute("data-sui-tree-key") || "tree";
    var prefixes = ["sui.tree." + key + ".", "sui.tree.label." + key + ".", "sui.tree.checked." + key, "sui.tree.selection." + key];
    try {
      var remove = [];
      for (var i = 0; i < localStorage.length; i++) {
        var k = localStorage.key(i);
        for (var j = 0; j < prefixes.length; j++) {
          if (k && k.indexOf(prefixes[j]) === 0) { remove.push(k); break; }
        }
      }
      for (var r = 0; r < remove.length; r++) localStorage.removeItem(remove[r]);
    } catch (e) {}
  }
  function countNodes(nodes) {
    var n = 0;
    for (var i = 0; i < nodes.length; i++) n += 1 + countNodes(nodes[i].children || []);
    return n;
  }
  function sortNodes(nodes, dir) {
    if (dir !== 'asc' && dir !== 'desc') return nodes;
    var copy = nodes.slice();
    copy.sort(function(a, b) {
      var al = String(a.label || '').toLowerCase(), bl = String(b.label || '').toLowerCase();
      if (al < bl) return dir === 'asc' ? -1 : 1;
      if (al > bl) return dir === 'asc' ? 1 : -1;
      return 0;
    });
    for (var i = 0; i < copy.length; i++) {
      if (copy[i].children && copy[i].children.length) copy[i].children = sortNodes(copy[i].children, dir);
    }
    return copy;
  }
  function importedNodes(name, text, format) {
    var root = treeTarget(name);
    if (!root) return { error: "tree not found: " + name };
    var nodes;
    try { nodes = format === "xml" ? parseXml(text) : parseJson(text); } catch (e) {
      return { error: (e && e.message) ? e.message : String(e) };
    }
    clearTreeStorage(root);
    store[treeName(root)] = nodes;
    root.__suiTree = false;
    init(root);
    if (root.__suiTreeApi && root.__suiTreeApi.notifyTreeChange) root.__suiTreeApi.notifyTreeChange();
    return { count: countNodes(nodes) };
  }
  function exportNode(li) {
    var out = { label: "", key: li.getAttribute("data-sui-key") || "" };
    var label = li.querySelector(".sui-tree-label");
    out.label = label ? label.textContent : "";
    var group = itemGroup(li);
    if (group) {
      out.open = li.__suiTreeOpen !== undefined ? !!li.__suiTreeOpen : li.getAttribute("data-sui-open") === "1";
    }
    var editable = li.getAttribute("data-sui-editable");
    if (editable === "1" || editable === "0") out.editable = editable === "1";
    var show = li.getAttribute("data-sui-show-rename-icon");
    if (show === "1" || show === "0") out.showRenameIcon = show === "1";
    var icon = li.getAttribute("data-sui-rename-icon") || "";
    if (icon) out.renameIcon = icon;
    var itemIcon = li.getAttribute("data-sui-icon") || "";
    if (itemIcon) out.icon = itemIcon;
    var badge = li.getAttribute("data-sui-badge") || "";
    if (badge) {
      out.badge = badge;
      var badgeType = li.getAttribute("data-sui-badge-type") || "neutral";
      if (badgeType !== "neutral") out.badgeType = badgeType;
    }
    var meta = li.getAttribute("data-sui-meta") || "";
    if (meta) out.meta = meta;
    var size = li.getAttribute("data-sui-size") || "";
    if (size) out.size = size;
    var date = li.getAttribute("data-sui-date") || "";
    if (date) out.date = date;
    var status = li.getAttribute("data-sui-status") || "";
    if (status) out.status = status;
    var cellsRaw = li.getAttribute("data-sui-cells") || "";
    if (cellsRaw) {
      try {
        var cellsObj = JSON.parse(cellsRaw);
        if (cellsObj && typeof cellsObj === "object" && !Array.isArray(cellsObj)) out.cells = cellsObj;
      } catch (e) {}
    }
    if (li.getAttribute("data-sui-lazy") === "1") out.lazy = true;
    var tooltip = li.getAttribute("data-sui-tooltip") || "";
    if (tooltip) out.tooltip = tooltip;
    if (li.hasAttribute("aria-checked")) out.checked = li.getAttribute("aria-checked") === "true";
    if (group) {
      out.children = [];
      for (var i = 0; i < group.children.length; i++) {
        if (group.children[i].hasAttribute && group.children[i].hasAttribute("data-sui-tree-item")) {
          out.children.push(exportNode(group.children[i]));
        }
      }
    }
    return out;
  }
  function readTreeNodes(root) {
    var group = rootGroup(root), out = [];
    for (var i = 0; i < group.children.length; i++) {
      if (group.children[i].hasAttribute && group.children[i].hasAttribute("data-sui-tree-item")) {
        out.push(exportNode(group.children[i]));
      }
    }
    return out;
  }
  function xmlEscape(s) {
    return treeText(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;").replace(/'/g, "&apos;");
  }
  function xmlNodeLines(nodes, indent) {
    var out = [];
    for (var i = 0; i < nodes.length; i++) {
      var n = nodes[i];
      var attrs = " label=\"" + xmlEscape(n.label) + "\" key=\"" + xmlEscape(n.key) + "\"";
      if (n.open !== undefined) attrs += " open=\"" + (n.open ? "true" : "false") + "\"";
      if (n.editable !== undefined) attrs += " editable=\"" + (n.editable ? "true" : "false") + "\"";
      if (n.showRenameIcon !== undefined) attrs += " showRenameIcon=\"" + (n.showRenameIcon ? "true" : "false") + "\"";
      if (n.renameIcon) attrs += " renameIcon=\"" + xmlEscape(n.renameIcon) + "\"";
      if (n.icon) attrs += " icon=\"" + xmlEscape(n.icon) + "\"";
      if (n.iconOpen) attrs += " iconOpen=\"" + xmlEscape(n.iconOpen) + "\"";
      if (n.badge) {
        attrs += " badge=\"" + xmlEscape(n.badge) + "\"";
        if (n.badgeType) attrs += " badgeType=\"" + xmlEscape(n.badgeType) + "\"";
      }
      if (n.meta) attrs += " meta=\"" + xmlEscape(n.meta) + "\"";
      if (n.size) attrs += " size=\"" + xmlEscape(n.size) + "\"";
      if (n.date) attrs += " date=\"" + xmlEscape(n.date) + "\"";
      if (n.status) attrs += " status=\"" + xmlEscape(n.status) + "\"";
      if (n.cells) attrs += " cells=\"" + xmlEscape(JSON.stringify(n.cells)) + "\"";
      if (n.lazy) attrs += " lazy=\"true\"";
      if (n.tooltip) attrs += " tooltip=\"" + xmlEscape(n.tooltip) + "\"";
      if (n.checked !== undefined) attrs += " checked=\"" + (n.checked ? "true" : "false") + "\"";
      if (n.children && n.children.length) {
        out.push(indent + "<item" + attrs + ">");
        var kids = xmlNodeLines(n.children, indent + "  ");
        for (var j = 0; j < kids.length; j++) out.push(kids[j]);
        out.push(indent + "</item>");
      } else {
        out.push(indent + "<item" + attrs + "/>");
      }
    }
    return out;
  }
  function exportJson(name) {
    var root = treeTarget(name);
    if (!root) return ["", "tree not found: " + name];
    try { return [JSON.stringify(readTreeNodes(root), null, 2), null]; } catch (e) { return ["", treeText(e)]; }
  }
  function exportXml(name) {
    var root = treeTarget(name);
    if (!root) return ["", "tree not found: " + name];
    try {
      var lines = xmlNodeLines(readTreeNodes(root), "  ");
      return ["<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<tree name=\"" + xmlEscape(treeName(root)) + "\">\n" +
        (lines.length ? lines.join("\n") + "\n" : "") + "</tree>\n", null];
    } catch (e) { return ["", treeText(e)]; }
  }
  function safeFileName(v, fallback) {
    var s = treeText(v).replace(/[\\/:*?"<>|\x00-\x1f]/g, "_").trim();
    if (!s) s = fallback;
    return s.slice(0, 200);
  }
  function downloadText(name, filename, format) {
    var res = format === "xml" ? exportXml(name) : exportJson(name);
    if (res[1]) return [0, res[1]];
    var ext = format === "xml" ? ".xml" : ".json";
    var fn = safeFileName(filename, safeFileName(name, "tree") + ext);
    if (fn.toLowerCase().indexOf(ext) !== fn.length - ext.length) fn += ext;
    try {
      var blob = new Blob([res[0]], { type: format === "xml" ? "application/xml;charset=utf-8" : "application/json;charset=utf-8" });
      var url = URL.createObjectURL(blob);
      var a = document.createElement("a");
      a.href = url; a.download = fn; a.style.display = "none";
      document.body.appendChild(a); a.click();
      if (a.parentNode) a.parentNode.removeChild(a);
      setTimeout(function() { try { URL.revokeObjectURL(url); } catch (e) {} }, 0);
      return [1, null];
    } catch (e) { return [0, treeText(e)]; }
  }
  function selectedFile(path) {
    var dnd = window.__SUI_DND;
    if (!dnd || typeof dnd.files !== "function" || typeof dnd.objects !== "function") throw new Error("file picker runtime is not available");
    var files = dnd.files(), objects = dnd.objects();
    if (!files || !files.length) throw new Error("no selected file; choose a JSON/XML file first");
    var idx = 0;
    if (path) {
      idx = -1;
      for (var i = 0; i < files.length; i++) {
        if (files[i].path === path || files[i].name === path) { idx = i; break; }
      }
      if (idx < 0) throw new Error("selected file not found: " + path);
    }
    if (!objects || !objects[idx]) throw new Error("file content is unavailable; use fs.read with tree.importJson on desktop");
    return objects[idx];
  }
  function fileText(file) {
    if (file && typeof file.text === "function") return file.text();
    return new Promise(function(resolve, reject) {
      try {
        var r = new FileReader();
        r.onload = function() { resolve(String(r.result || "")); };
        r.onerror = function() { reject(new Error("cannot read the selected file")); };
        r.readAsText(file);
      } catch (e) { reject(e); }
    });
  }
  async function importFile(name, path, format) {
    try {
      var file = selectedFile(path);
      var max = (typeof __SUI_SEC !== "undefined" && __SUI_SEC.maxText) || 20000000;
      if (file.size && file.size > max) return [0, "file is too large (limit " + max + " bytes)"];
      var text = await fileText(file);
      if (text.length > max) return [0, "file is too large (limit " + max + " bytes)"];
      var res = importedNodes(name, text, format);
      return res.error ? [0, res.error] : [res.count, null];
    } catch (e) { return [0, treeText(e && e.message ? e.message : e)]; }
  }
  window.__SUI_TREE = {
    importJson: function(name, text) {
      var res = importedNodes(name, text, "json");
      return res.error ? [0, res.error] : [res.count, null];
    },
    importXml: function(name, text) {
      var res = importedNodes(name, text, "xml");
      return res.error ? [0, res.error] : [res.count, null];
    },
    importJsonFile: function(name, path) { return importFile(name, path, "json"); },
    importXmlFile: function(name, path) { return importFile(name, path, "xml"); },
    exportJson: exportJson,
    exportXml: exportXml,
    downloadJson: function(name, filename) { return downloadText(name, filename, "json"); },
    downloadXml: function(name, filename) { return downloadText(name, filename, "xml"); },
    expandAll: function(name) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      return [root.__suiTreeApi.expandAll(), null];
    },
    collapseAll: function(name) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      return [root.__suiTreeApi.collapseAll(), null];
    },
    expandTo: function(name, level) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      return [root.__suiTreeApi.expandTo(level), null];
    },
    setChildren: function(name, key, text) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      var res = root.__suiTreeApi.setChildren(key, text);
      return res.error ? [0, res.error] : [res.count, null];
    },
    setChildrenError: function(name, key, message) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      return root.__suiTreeApi.setChildrenError(key, message) ? [1, null] : [0, "tree item not found: " + key];
    },
    retryChildren: function(name, key) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      return root.__suiTreeApi.retryChildren(key) ? [1, null] : [0, "tree item not found: " + key];
    },
    lastMove: function(name) {
      var root = treeTarget(name);
      if (!root) return ["", "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return ["", "tree is empty: " + name];
      return [JSON.stringify(root.__suiTreeApi.lastMove()), null];
    },
    undoMove: function(name) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      var res = root.__suiTreeApi.undoMove();
      return res.error ? [0, res.error] : [res.count, null];
    },
    setColumns: function(name, json) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      try {
        var list = typeof json === "string" ? JSON.parse(json) : json;
        return [root.__suiTreeApi.setColumns(list), null];
      } catch (e) { return [0, (e && e.message) ? e.message : String(e)]; }
    },
    getColumns: function(name) {
      var root = treeTarget(name);
      if (!root) return ["", "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return ["", "tree is empty: " + name];
      return [JSON.stringify(root.__suiTreeApi.getColumns()), null];
    },
    moveColumn: function(name, from, to) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      return [root.__suiTreeApi.moveColumn(from, to) ? 1 : 0, null];
    },
    setCell: function(name, key, column, value) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      try { return [root.__suiTreeApi.setCell(key, column, value) ? 1 : 0, null]; }
      catch (e) { return [0, (e && e.message) ? e.message : String(e)]; }
    },
    selectedKeys: function(name) {
      var root = treeTarget(name);
      if (!root) return ["", "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return ["", "tree is empty: " + name];
      return [JSON.stringify(root.__suiTreeApi.selectedKeys()), null];
    },
    checkedKeys: function(name) {
      var root = treeTarget(name);
      if (!root) return ["", "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return ["", "tree is empty: " + name];
      return [JSON.stringify(root.__suiTreeApi.checkedKeys()), null];
    },
    search: function(name, term) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      if (!root.__suiTreeApi) init(root);
      if (!root.__suiTreeApi) return [0, "tree is empty: " + name];
      root.__suiTreeApi.search(term);
      return [1, null];
    },
    clear: function(name) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      store[treeName(root)] = [];
      clearTreeStorage(root);
      root.__suiTree = false;
      init(root);
      if (root.__suiTreeApi && root.__suiTreeApi.notifyTreeChange) root.__suiTreeApi.notifyTreeChange();
      return [0, null];
    }
  };
  function initAll() {
    var all = document.querySelectorAll('.sui-tree');
    for (var i = 0; i < all.length; i++) init(all[i]);
  }
  initAll();
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      setTimeout(function() { pending = false; initAll(); }, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
})();`
