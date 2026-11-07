package main

// tree_runtime.go - browser runtime for the nested tree control.
//
// The generated markup is a role=tree with nested role=treeitem/group nodes.
// The runtime owns open state, roving focus, keyboard navigation and selection;
// selection is synchronized with a hidden value input and optional state setter
// + onChange handler. Open state is persisted per tree path in localStorage.
const suiTreeRuntime = `(function(){
  var lastDown = { tree: '', key: '', time: 0 };
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
    if (!root || root.__suiTree) return;
    var treeKey = root.getAttribute('data-sui-tree-key') || 'tree';
    var guides = root.getAttribute('data-sui-tree-guides') === '1';
    var checkboxes = root.getAttribute('data-sui-tree-checkbox') === '1';
    var counts = root.getAttribute('data-sui-tree-counts') === '1';
    var treeIcon = root.getAttribute('data-sui-tree-icon') || '';
    var treeIconOpen = root.getAttribute('data-sui-tree-icon-open') || '';
    var rowHeight = parseInt(root.getAttribute('data-sui-tree-row-height') || '', 10);
    if (!isNaN(rowHeight) && rowHeight >= 12 && rowHeight <= 120) {
      root.style.setProperty('--sui-tree-row-height', rowHeight + 'px');
    }
    var imported = store[treeName(root)];
    if (imported) {
      renderNodes(root, imported, {
        guides: guides, checkbox: checkboxes, counts: counts,
        icon: treeIcon, iconOpen: treeIconOpen
      });
    }
    root.__suiTree = true;
    var hidden = root.querySelector('.sui-tree-value');
    if (!hidden) return;
    var setter = root.getAttribute('data-sui-tree-setter') || '';
    var handler = root.getAttribute('data-sui-tree-onchange') || '';
    var checkHandler = root.getAttribute('data-sui-tree-oncheck') || '';
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
    var editing = null;
    var selectionTimer = null;
    var pendingItem = null;
    var checkedStore = {};
    if (checkboxes) {
      try {
        var rawChecked = localStorage.getItem('sui.tree.checked.' + treeKey);
        var parsedChecked = rawChecked ? JSON.parse(rawChecked) : null;
        if (parsedChecked && typeof parsedChecked === 'object' && !Array.isArray(parsedChecked)) checkedStore = parsedChecked;
      } catch (e) { checkedStore = {}; }
    }
    if (!root.getAttribute('aria-label')) root.setAttribute('aria-label', treeKey);
    if (checkboxes) root.setAttribute('aria-multiselectable', 'true');

    function itemLabelEl(item) {
      return item.querySelector('.sui-tree-label');
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
      old.label.hidden = false;
      if (old.input.parentNode) old.input.parentNode.removeChild(old.input);
      var key = old.item.getAttribute('data-sui-key') || '';
      var path = old.item.getAttribute('data-sui-path') || '';
      saveLabel(root, path, value);
      if (renameKeySetter && typeof window[renameKeySetter] === 'function') window[renameKeySetter](key);
      if (renameValueSetter && typeof window[renameValueSetter] === 'function') window[renameValueSetter](value);
      if (renameHandler && typeof window[renameHandler] === 'function') window[renameHandler]();
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
    function setOpen(item, open, persist) {
      var group = itemGroup(item);
      if (!group) return;
      group.hidden = !open;
      item.setAttribute('aria-expanded', open ? 'true' : 'false');
      var toggle = itemToggle(item);
      if (toggle) {
        toggle.textContent = open ? '\u25be' : '\u25b8';
        toggle.setAttribute('aria-label', (open ? 'Collapse ' : 'Expand ') + itemLabel(item));
      }
      item.__suiTreeOpen = open;
      if (persist) saveOpen(root, item.getAttribute('data-sui-path') || '', open);
    }
    function visible(item) {
      var p = item.parentElement;
      while (p && p !== root) {
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
    function select(item, notify) {
      current = item.getAttribute('data-sui-key') || '';
      hidden.value = current;
      items.forEach(function(x) {
        var on = x === item;
        x.classList.toggle('selected', on);
        x.setAttribute('aria-selected', on ? 'true' : 'false');
        x.tabIndex = on ? 0 : -1;
      });
      if (notify) {
        if (setter && typeof window[setter] === 'function') window[setter](current);
        if (handler && typeof window[handler] === 'function') window[handler]();
      }
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
    function renameIconFor(item) {
      var itemShow = item.getAttribute('data-sui-show-rename-icon') || 'inherit';
      var show = itemShow === '1' ? true : itemShow === '0' ? false : showTreeRenameIcon;
      var icon = item.getAttribute('data-sui-rename-icon') || treeRenameIcon;
      return show ? icon : '';
    }
    items.forEach(function(item) {
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
        var tip = item.getAttribute('data-sui-tooltip') || '';
        labelEl.title = tip || labelEl.textContent;
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
      if (row) row.addEventListener('click', function(ev) {
        if (ev.target && ev.target.closest && ev.target.closest('.sui-tree-toggle')) return;
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
        if (item.__suiEditable) {
          // Keep the row in the DOM briefly so a double-click can start rename
          // before the reactive re-render replaces it. The pending class gives
          // immediate visual feedback for a single click.
          item.classList.add('sui-tree-pending');
          pendingItem = item;
          selectionTimer = setTimeout(function() {
            selectionTimer = null;
            if (pendingItem) pendingItem.classList.remove('sui-tree-pending');
            pendingItem = null;
            select(item, true);
            focusItem(item);
          }, 500);
        } else {
          select(item, true);
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
        row.appendChild(editBtn);
      }
      item.addEventListener('focus', function() {
        items.forEach(function(x) { x.tabIndex = x === item ? 0 : -1; });
      });
      item.addEventListener('keydown', function(ev) {
        var vis = items.filter(visible);
        var idx = vis.indexOf(item);
        if (ev.key === 'F2' && item.__suiEditable) {
          ev.preventDefault();
          startRename(item);
        } else if (ev.key === 'ArrowDown') {
          ev.preventDefault();
          if (idx >= 0 && idx + 1 < vis.length) focusItem(vis[idx + 1]);
        } else if (ev.key === 'ArrowUp') {
          ev.preventDefault();
          if (idx > 0) focusItem(vis[idx - 1]);
        } else if (ev.key === 'Home') {
          ev.preventDefault();
          if (vis.length) focusItem(vis[0]);
        } else if (ev.key === 'End') {
          ev.preventDefault();
          if (vis.length) focusItem(vis[vis.length - 1]);
        } else if (ev.key === 'ArrowRight') {
          ev.preventDefault();
          var group = itemGroup(item);
          if (group && !item.__suiTreeOpen) setOpen(item, true, true);
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
          else select(item, true);
        }
      });
    });
    initChecks();
    var selected = null;
    items.forEach(function(item) {
      if ((item.getAttribute('data-sui-key') || '') === current) selected = item;
    });
    if (selected) {
      items.forEach(function(item) {
        var on = item === selected;
        item.classList.toggle('selected', on);
        item.setAttribute('aria-selected', on ? 'true' : 'false');
        item.tabIndex = on ? 0 : -1;
      });
    } else {
      // No selected value: keep the first visible row focusable without
      // selecting it or writing a synthetic value into the bound state.
      for (var i = 0; i < items.length; i++) {
        if (visible(items[i])) { items[i].tabIndex = 0; break; }
      }
    }
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
    var attrs = ["open", "editable", "showRenameIcon", "icon", "iconOpen", "badge", "badgeType", "meta", "checked", "tooltip"];
    for (var i = 0; i < attrs.length; i++) {
      var v = el.getAttribute(attrs[i]);
      if (v !== null) o[attrs[i]] = v;
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
    li.setAttribute("data-sui-icon", node.icon || "");
    li.setAttribute("data-sui-badge", node.badge || "");
    li.setAttribute("data-sui-badge-type", node.badgeType || "neutral");
    li.setAttribute("data-sui-meta", node.meta || "");
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
    if (node.children && node.children.length) {
      var toggle = document.createElement("button");
      toggle.type = "button";
      toggle.className = "sui-tree-toggle";
      toggle.tabIndex = -1;
      toggle.setAttribute("aria-label", "Toggle item");
      toggle.textContent = node.open ? "\u25be" : "\u25b8";
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
    var icon = node.icon || opts.icon || "";
    if (node.open && node.iconOpen) icon = node.iconOpen;
    else if (node.open && opts.iconOpen) icon = opts.iconOpen;
    if (icon) {
      var iconEl = document.createElement("span");
      iconEl.className = "sui-tree-icon";
      iconEl.setAttribute("aria-hidden", "true");
      iconEl.textContent = icon;
      row.appendChild(iconEl);
    }
    var label = document.createElement("span");
    label.className = "sui-tree-label";
    label.textContent = node.label;
    label.title = node.tooltip || node.label;
    row.appendChild(label);
    if (node.badge) {
      var badge = document.createElement("span");
      badge.className = "sui-tree-badge sui-tree-badge-" + (node.badgeType || "neutral");
      badge.textContent = node.badge;
      row.appendChild(badge);
    }
    if (opts.counts && node.children && node.children.length) {
      var countBadge = document.createElement("span");
      countBadge.className = "sui-tree-badge sui-tree-badge-count";
      countBadge.textContent = String(node.children.length);
      row.appendChild(countBadge);
    }
    if (node.meta) {
      var meta = document.createElement("span");
      meta.className = "sui-tree-meta";
      meta.textContent = node.meta;
      row.appendChild(meta);
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
    var group = rootGroup(root);
    while (group.firstChild) group.removeChild(group.firstChild);
    for (var i = 0; i < nodes.length; i++) group.appendChild(makeNode(nodes[i], String(i), 1, opts || {}));
  }
  function clearTreeStorage(root) {
    var key = root.getAttribute("data-sui-tree-key") || "tree";
    var prefixes = ["sui.tree." + key + ".", "sui.tree.label." + key + ".", "sui.tree.checked." + key];
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
    clear: function(name) {
      var root = treeTarget(name);
      if (!root) return [0, "tree not found: " + name];
      store[treeName(root)] = [];
      clearTreeStorage(root);
      root.__suiTree = false;
      init(root);
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
