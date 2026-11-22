package core

// dnd.go - browser drag-and-drop runtime for generated SimpleUI apps.
//
// The language has no event object and no per-element closures, so DnD is
// exposed through a small runtime state:
//
//	drag: true
//	dragKey: item.id
//	dropKey: item.id
//	onDrop: reorder
//	input value: path dropFiles: path
//
// `drag.source()` / `drag.target()` expose the keys captured by the runtime;
// `dropFiles` is an automatic field target that accepts files and directories
// from the operating system. Files are metadata only (name/path/size/type):
// their bytes are never read by this runtime.
const suiDndRuntime = `const __SUI_DND = (function(){
  var state = { from: '', to: '', data: '', files: [], objects: [], folder: '', error: '', active: false, source: null };
  var MAX_FILES = 2000;
  var MAX_DEPTH = 16;
  var MAX_PATH = 4096;
  var lastTarget = null;
  var dropSeq = 0;
  var bound = false;

  function text(v) {
    return v === null || v === undefined ? '' : String(v);
  }
  function cleanPath(v) {
    var p = text(v);
    while (p.charAt(0) === '/') p = p.slice(1);
    return p;
  }
  function closest(el, selector) {
    var n = el && el.nodeType === 3 ? el.parentNode : el;
    while (n && n.nodeType === 1) {
      if (n.matches && n.matches(selector)) return n;
      n = n.parentNode;
    }
    return null;
  }
  function dropZone(el) {
    return closest(el, '[data-sui-dropzone], [data-sui-dropfiles], [data-sui-drop-key]');
  }
  function isFileDrag(dt) {
    if (!dt) return false;
    if (dt.files && dt.files.length) return true;
    if (!dt.types) return false;
    for (var i = 0; i < dt.types.length; i++) {
      if (dt.types[i] === 'Files') return true;
    }
    return false;
  }
  function fileMeta(file, entryPath) {
    if (!file) return null;
    var p = text(file.path || file.webkitRelativePath || '');
    if (!p) p = cleanPath(entryPath || file.name || '');
    return {
      name: text(file.name),
      path: p,
      size: file.size || 0,
      type: text(file.type),
      isDirectory: false,
      lastModified: file.lastModified || 0
    };
  }
  function filesFromList(list) {
    var out = [];
    if (!list) return out;
    for (var i = 0; i < list.length && out.length < MAX_FILES; i++) {
      var m = fileMeta(list[i], '');
      if (m && m.path.length <= MAX_PATH) out.push(m);
    }
    return out;
  }
  function immediateFiles(dt) {
    var out = filesFromList(dt && dt.files);
    if (!out.length && dt && dt.items) {
      for (var i = 0; i < dt.items.length; i++) {
        var item = dt.items[i];
        if (!item || item.kind !== 'file' || !item.webkitGetAsEntry) continue;
        var entry = null;
        try { entry = item.webkitGetAsEntry(); } catch (e) { entry = null; }
        if (entry && entry.isDirectory) {
          out.push({
            name: text(entry.name),
            path: cleanPath(entry.fullPath),
            size: 0,
            type: '',
            isDirectory: true,
            lastModified: 0
          });
        }
      }
    }
    return out;
  }
  function readAll(reader) {
    return new Promise(function(resolve, reject) {
      var all = [];
      function step() {
        reader.readEntries(function(batch) {
          if (!batch || !batch.length) { resolve(all); return; }
          all = all.concat(Array.prototype.slice.call(batch));
          step();
        }, function() { reject('cannot list a directory'); });
      }
      step();
    });
  }
  function walkEntry(entry, out, depth, objects) {
    if (out.length >= MAX_FILES) return Promise.reject('too many files (limit ' + MAX_FILES + ')');
    if (depth > MAX_DEPTH) return Promise.reject('directory nesting is too deep (limit ' + MAX_DEPTH + ')');
    if (entry && entry.isFile) {
      return new Promise(function(resolve, reject) {
        entry.file(function(file) {
          var m = fileMeta(file, entry.fullPath);
          if (!m || !m.path) { resolve(); return; }
          if (m.path.length > MAX_PATH) { reject('path is too long (limit ' + MAX_PATH + ')'); return; }
          out.push(m);
          if (objects) objects.push(file);
          resolve();
        }, function() { reject('cannot read ' + text(entry.name)); });
      });
    }
    if (entry && entry.isDirectory) {
      return readAll(entry.createReader()).then(function(entries) {
        var chain = Promise.resolve();
        entries.forEach(function(child) {
          chain = chain.then(function() { return walkEntry(child, out, depth + 1, objects); });
        });
        return chain;
      });
    }
    return Promise.resolve();
  }
  function collectFiles(dt, objects) {
    if (!dt) return Promise.resolve([]);
    var entries = [];
    if (dt.items && dt.items.length) {
      for (var i = 0; i < dt.items.length; i++) {
        var item = dt.items[i];
        if (!item || item.kind !== 'file' || !item.webkitGetAsEntry) continue;
        var entry = null;
        try { entry = item.webkitGetAsEntry(); } catch (e) { entry = null; }
        if (entry) entries.push(entry);
      }
    }
    if (!entries.length) {
      if (objects && dt.files) {
        for (var j = 0; j < dt.files.length; j++) objects.push(dt.files[j]);
      }
      return Promise.resolve(filesFromList(dt.files));
    }
    var out = [];
    var chain = Promise.resolve();
    entries.forEach(function(entry) {
      chain = chain.then(function() { return walkEntry(entry, out, 0, objects); });
    });
    return chain.then(function() {
      if (objects && !objects.length && dt.files) {
        for (var i = 0; i < dt.files.length; i++) objects.push(dt.files[i]);
      }
      if (!out.length) return immediateFiles(dt);
      return out;
    });
  }
  function setTarget(el) {
    if (lastTarget && lastTarget !== el) lastTarget.classList.remove('sui-drop-active');
    lastTarget = el;
    if (el) {
      el.classList.add('sui-drop-active');
      state.to = text(el.getAttribute('data-sui-drop-key'));
    } else {
      state.to = '';
    }
  }
  function clearTarget() {
    if (lastTarget) lastTarget.classList.remove('sui-drop-active');
    lastTarget = null;
  }
  function showError(el, msg) {
    state.error = text(msg);
    if (el && el.parentNode) {
      var old = el.__suiDropError;
      if (old && old.parentNode) old.parentNode.removeChild(old);
      var box = document.createElement('div');
      box.className = 'sui-drop-error';
      box.setAttribute('role', 'alert');
      box.textContent = state.error;
      el.parentNode.insertBefore(box, el.nextSibling);
      el.__suiDropError = box;
    }
  }
  function applyDropFiles(el, setter, list) {
    var paths = [];
    for (var i = 0; i < list.length; i++) {
      var p = text(list[i] && (list[i].path || list[i].name));
      if (p) paths.push(p);
    }
    var sep = el && el.tagName === 'TEXTAREA' ? '\n' : ', ';
    var value = paths.join(sep);
    if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA')) {
      try { el.value = value; } catch (e) {}
    }
    var fn = window[setter];
    if (typeof fn !== 'function') {
      showError(el, 'Cannot accept the files: drop target state is not available');
      return;
    }
    fn(value);
  }
  function clearError(el) {
    if (el && el.__suiDropError && el.__suiDropError.parentNode) {
      el.__suiDropError.parentNode.removeChild(el.__suiDropError);
    }
    if (el) el.__suiDropError = null;
  }
  function pickedFiles(input) {
    var files = input && input.files ? input.files : [];
    if (files.length > MAX_FILES) {
      return { files: [], error: 'too many files (limit ' + MAX_FILES + ')' };
    }
    var out = [];
    for (var i = 0; i < files.length; i++) {
      var m = fileMeta(files[i], files[i].webkitRelativePath || files[i].name);
      if (!m || !m.path) continue;
      if (m.path.length > MAX_PATH) {
        return { files: [], error: 'path is too long (limit ' + MAX_PATH + ')' };
      }
      out.push(m);
    }
    return { files: out, error: '' };
  }
  function handlePicked(input, trigger, setter) {
    var res = pickedFiles(input);
    state.folder = '';
    state.error = '';
    clearError(trigger);
    if (res.error) {
      state.files = [];
      state.objects = [];
      showError(trigger, 'Cannot accept the files: ' + res.error);
      return;
    }
    state.files = res.files;
    state.objects = input && input.files ? Array.prototype.slice.call(input.files) : [];
    if (!res.files.length) {
      state.objects = [];
      showError(trigger, 'Cannot accept the files: no files were selected');
      return;
    }
    applyDropFiles(trigger, setter, res.files);
  }
  function openPicker(trigger) {
    var inp = document.createElement('input');
    inp.type = 'file';
    inp.style.cssText = 'position:fixed;left:-9999px;top:0;width:1px;height:1px;opacity:0';
    var accept = trigger.getAttribute('data-sui-pick-accept');
    if (accept) inp.accept = accept;
    if (trigger.getAttribute('data-sui-pick-multiple') === '1') inp.multiple = true;
    if (trigger.getAttribute('data-sui-pick-directory') === '1') {
      inp.webkitdirectory = true;
      inp.setAttribute('webkitdirectory', '');
    }
    var setter = trigger.getAttribute('data-sui-pickfiles');
    if (!setter) return;
    function cleanup() {
      if (inp.parentNode) inp.parentNode.removeChild(inp);
    }
    inp.addEventListener('change', function() {
      handlePicked(inp, trigger, setter);
      cleanup();
    });
    inp.addEventListener('cancel', cleanup);
    document.body.appendChild(inp);
    try {
      inp.click();
    } catch (e) {
      cleanup();
      showError(trigger, 'Cannot accept the files: file picker is not available');
    }
    setTimeout(cleanup, 120000);
  }
  async function collectDirectoryHandle(handle) {
    var files = [];
    var objects = [];
    async function walk(dir, prefix, depth) {
      if (depth > MAX_DEPTH) throw new Error('directory nesting is too deep (limit ' + MAX_DEPTH + ')');
      for await (const entry of dir.values()) {
        if (entry.kind === 'file') {
          if (files.length >= MAX_FILES) throw new Error('too many files (limit ' + MAX_FILES + ')');
          var file = await entry.getFile();
          var p = prefix + entry.name;
          var m = fileMeta(file, p);
          if (!m || !m.path) continue;
          if (m.path.length > MAX_PATH) throw new Error('path is too long (limit ' + MAX_PATH + ')');
          files.push(m);
          objects.push(file);
        } else if (entry.kind === 'directory') {
          await walk(entry, prefix + entry.name + '/', depth + 1);
        }
      }
    }
    await walk(handle, handle.name + '/', 0);
    return { files: files, objects: objects };
  }
  function applyDirectoryHandle(handle, res, trigger, setter) {
    state.folder = text(handle.name);
    state.files = res.files;
    state.objects = res.objects;
    state.error = '';
    clearError(trigger);
    if (!res.files.length) {
      // Empty folder: keep the folder name usable even though there are no
      // file objects to upload.
      var fn = window[setter];
      if (typeof fn === 'function') fn(state.folder);
      return;
    }
    applyDropFiles(trigger, setter, res.files);
  }
  async function openFSPicker(trigger) {
    if (typeof window.showDirectoryPicker !== 'function') {
      showError(trigger, 'Cannot accept the folder: File System Access API is not available in this browser');
      return;
    }
    var setter = trigger.getAttribute('data-sui-pickfiles');
    if (!setter) return;
    try {
      var handle = await window.showDirectoryPicker({ mode: 'read' });
      var res = await collectDirectoryHandle(handle);
      applyDirectoryHandle(handle, res, trigger, setter);
    } catch (e) {
      if (e && e.name === 'AbortError') return;
      showError(trigger, 'Cannot accept the folder: ' + text(e && e.message ? e.message : e));
    }
  }
  function onPickClick(e) {
    var trigger = closest(e.target, '[data-sui-pickfiles]');
    if (!trigger) return;
    var mode = text(trigger.getAttribute('data-sui-pick-mode')).toLowerCase();
    if (mode === 'fs' || (mode === 'auto' && typeof window.showDirectoryPicker === 'function')) {
      e.preventDefault();
      openFSPicker(trigger);
      return;
    }
    // The native <input type=file> opens its own dialog and is handled by the
    // change listener below.
    if (trigger.tagName === 'INPUT' && trigger.type === 'file') return;
    e.preventDefault();
    openPicker(trigger);
  }
  function onPickChange(e) {
    var input = e.target;
    if (!input || input.tagName !== 'INPUT' || input.type !== 'file') return;
    var setter = input.getAttribute('data-sui-pickfiles');
    if (!setter) return;
    handlePicked(input, input, setter);
  }
  function onDragStart(e) {
    dropSeq++;
    var el = closest(e.target, '[data-sui-drag-key]');
    state.from = el ? text(el.getAttribute('data-sui-drag-key')) : '';
    state.to = '';
    state.data = state.from;
    state.error = '';
    state.active = true;
    state.source = el;
    if (el) el.classList.add('sui-dragging');
    if (e.dataTransfer) {
      if (state.from) e.dataTransfer.setData('text/plain', state.from);
      e.dataTransfer.effectAllowed = 'move';
    }
  }
  function onDragOver(e) {
    var el = dropZone(e.target);
    if (el) {
      e.preventDefault();
      if (e.dataTransfer) {
        e.dataTransfer.dropEffect = isFileDrag(e.dataTransfer) ? 'copy' : 'move';
      }
      setTarget(el);
    } else {
      // A file dragged from the desktop must never navigate away from the app.
      if (isFileDrag(e.dataTransfer)) e.preventDefault();
      setTarget(null);
    }
  }
  function onDragLeave(e) {
    var el = dropZone(e.target);
    if (el && el === lastTarget) {
      if (!e.relatedTarget || !el.contains(e.relatedTarget)) setTarget(null);
    }
  }
  function onDrop(e) {
    var el = dropZone(e.target);
    var dt = e.dataTransfer;
    if (!el) {
      if (isFileDrag(dt)) e.preventDefault();
      dropSeq++;
      return;
    }
    e.preventDefault();
    setTarget(el);
    state.error = '';
    clearError(el);
    if (dt) {
      var plain = '';
      try { plain = text(dt.getData('text/plain')); } catch (err) { plain = ''; }
      if (plain) state.data = plain;
      if (!state.from && plain) state.from = plain;
      if (isFileDrag(dt)) {
        state.files = immediateFiles(dt);
        state.objects = [];
        state.folder = '';
      }
    }
    if (isFileDrag(dt)) state.active = true;
    var setter = el.getAttribute('data-sui-dropfiles');
    if (dt && isFileDrag(dt)) {
      var seq = ++dropSeq;
      collectFiles(dt, state.objects).then(function(list) {
        if (seq !== dropSeq) return;
        state.files = list;
        if (!list.length) {
          if (setter) showError(el, 'Cannot accept the files: no files were found');
          return;
        }
        if (setter) applyDropFiles(el, setter, list);
      }).catch(function(err) {
        if (seq !== dropSeq) return;
        state.objects = [];
        showError(el, 'Cannot accept the files: ' + text(err || 'unreadable data'));
      });
    } else {
      dropSeq++;
    }
    setTimeout(function() {
      clearTarget();
      state.active = false;
    }, 0);
  }
  function onDragEnd(e) {
    if (state.source) state.source.classList.remove('sui-dragging');
    clearTarget();
    setTimeout(function() {
      state.from = '';
      state.data = '';
      state.active = false;
      state.source = null;
      // Keep state.files/state.objects: they are the last selection and are
      // consumed by http.postFiles() after the drag/drop gesture has ended.
    }, 0);
  }
  function init() {
    if (bound || typeof document === 'undefined') return;
    bound = true;
    document.addEventListener('dragstart', onDragStart, true);
    document.addEventListener('dragenter', onDragOver, true);
    document.addEventListener('dragover', onDragOver, true);
    document.addEventListener('dragleave', onDragLeave, true);
    document.addEventListener('drop', onDrop, true);
    document.addEventListener('dragend', onDragEnd, true);
    document.addEventListener('click', onPickClick, true);
    document.addEventListener('change', onPickChange, true);
  }
  function afterRender() {
    clearTarget();
  }
  function folderName() {
    if (state.folder) return state.folder;
    // A directory dropped without files is represented by a synthetic entry.
    if (state.files.length === 1 && state.files[0] && state.files[0].isDirectory) {
      return text(state.files[0].name || state.files[0].path);
    }
    var root = '';
    for (var i = 0; i < state.files.length; i++) {
      var p = text(state.files[i] && state.files[i].path);
      var slash = p.indexOf('/');
      if (slash <= 0) return '';
      var seg = p.slice(0, slash);
      if (!root) root = seg;
      else if (root !== seg) return '';
    }
    return root;
  }
  function clear() {
    dropSeq++;
    if (state.source) state.source.classList.remove('sui-dragging');
    clearTarget();
    state.from = '';
    state.to = '';
    state.data = '';
    state.files = [];
    state.objects = [];
    state.folder = '';
    state.error = '';
    state.active = false;
    state.source = null;
  }
  init();
  return {
    afterRender: afterRender,
    from: function() { return state.from; },
    to: function() { return state.to; },
    data: function() { return state.data; },
    files: function() { return state.files; },
    objects: function() { return state.objects; },
    folder: function() { return folderName(); },
    error: function() { return state.error; },
    active: function() { return state.active; },
    clear: clear
  };
})();
window.__SUI_DND = __SUI_DND;

function __sui_drag_source() { return __SUI_DND.from(); }
function __sui_drag_target() { return __SUI_DND.to(); }
function __sui_drag_data() { return __SUI_DND.data(); }
function __sui_drag_files() { return __SUI_DND.files(); }
function __sui_drag_folder() { return __SUI_DND.folder(); }
function __sui_drag_error() { return __SUI_DND.error(); }
function __sui_drag_active() { return __SUI_DND.active(); }
function __sui_drag_clear() { return __SUI_DND.clear(); }
// __sui_dnd_read_bytes reads the bytes of a file in the current selection.
// The browser only exposes the files the user picked or dropped, so this never
// touches a path: the name matches the selected file name or relative path.
function __sui_dnd_read_bytes(name) {
  var want = String(name === null || name === undefined ? '' : name).trim();
  var files = __SUI_DND.objects() || [];
  var meta = __SUI_DND.files() || [];
  if (!files.length) return Promise.resolve([null, 'no file is selected']);
  var idx = -1;
  for (var i = 0; i < files.length; i++) {
    var m = meta[i] || {};
    var f = files[i] || {};
    var p = String(m.path || m.name || '');
    var base = p.slice(p.lastIndexOf('/') + 1);
    var n = String(f.name || '');
    if (want && (p === want || base === want || n === want)) { idx = i; break; }
  }
  if (idx < 0) {
    if (!want && files.length === 1) idx = 0;
    else return Promise.resolve([null, 'the file "' + want + '" is not part of the current selection']);
  }
  var file = files[idx];
  if (!file || typeof file.arrayBuffer !== 'function') {
    return Promise.resolve([null, 'the browser cannot read the bytes of this file']);
  }
  return file.arrayBuffer().then(function (buf) {
    var v = new Uint8Array(buf);
    if (v.length > 10000000) return [null, 'the file is too large (limit 10000000 bytes)'];
    var out = new Array(v.length);
    for (var i = 0; i < v.length; i++) out[i] = v[i];
    return [out, null];
  }).catch(function (e) {
    return [null, String(e && e.message ? e.message : e)];
  });
}
function __sui_files_form() {
  var files = __SUI_DND.objects();
  if (!files || !files.length) return null;
  var meta = __SUI_DND.files();
  var fd = new FormData();
  for (var i = 0; i < files.length; i++) {
    var f = files[i];
    fd.append('files', f, f.name || 'file');
    // Parallel field: browsers do not expose absolute paths, but the relative
    // path (folder/sub/file) is useful to the backend for display/grouping.
    var p = '';
    if (meta && meta[i]) p = meta[i].path || meta[i].name || '';
    fd.append('paths', p || f.name || '');
  }
  return fd;
}
function __sui_move_item(arr, from, to) {
  var src = Array.isArray(arr) ? arr.slice() : [];
  var a = parseInt(from, 10);
  var b = parseInt(to, 10);
  if (!isFinite(a) || !isFinite(b) || a < 0 || a >= src.length) return src;
  if (b < 0) b = 0;
  if (b >= src.length) b = src.length - 1;
  var item = src.splice(a, 1)[0];
  src.splice(b, 0, item);
  return src;
}`
