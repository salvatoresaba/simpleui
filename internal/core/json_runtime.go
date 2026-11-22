package core

import "strings"

// json_runtime.go - the browser/desktop implementation of the `json` module.
//
// JSON values are native SUI maps/arrays. This runtime adds JSON Pointer
// (RFC 6901), JSON Patch (RFC 6902), JSON Merge Patch (RFC 7386), a
// multi-error JSON Schema check, byte decoding and formatting options. It is
// emitted only when the `json` module is imported.
//
// Security: pointers are data, not code. Patch operations are applied to a
// deep clone and abort on the first error, so the input is never partially
// modified. Object keys named "__proto__" are rejected to avoid prototype
// pollution. File access still goes through the desktop bridge allow-list.
const jsonRuntimeSource = `var __SUI_JSON = (function () {
  var MAX_BYTES = 20000000;
  var MAX_DEPTH = 128;
  var MAX_ERRORS = 100;
  function fail(message) { return [null, String(message == null ? 'JSON error' : message).slice(0, 500)]; }
  function ok(value) { return [value, null]; }
  function hasOwn(obj, key) { return Object.prototype.hasOwnProperty.call(obj, key); }
  function isObject(value) { return value !== null && typeof value === 'object' && !Array.isArray(value); }
  function setKey(obj, key, value) {
    if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
    Object.defineProperty(obj, key, { value: value, writable: true, enumerable: true, configurable: true });
  }
  function cloneValue(value, depth, seen) {
    if (value === null || typeof value !== 'object') return value;
    if (depth > MAX_DEPTH) throw new Error('JSON value is nested too deeply (max ' + MAX_DEPTH + ')');
    seen = seen || [];
    if (seen.indexOf(value) >= 0) throw new Error('JSON value contains a cycle');
    seen.push(value);
    var out, i, keys;
    if (Array.isArray(value)) {
      out = [];
      for (i = 0; i < value.length; i++) out.push(cloneValue(value[i], depth + 1, seen));
    } else {
      out = {};
      keys = Object.keys(value);
      for (i = 0; i < keys.length; i++) setKey(out, keys[i], cloneValue(value[keys[i]], depth + 1, seen));
    }
    seen.pop();
    return out;
  }
  function clone(value) {
    try { return ok(cloneValue(value, 0)); }
    catch (e) { return fail('json.clone: ' + (e && e.message ? e.message : String(e))); }
  }
  function equalValue(a, b, depth) {
    if (a === b) return true;
    if (a === null || b === null || typeof a !== 'object' || typeof b !== 'object') return false;
    if (depth > MAX_DEPTH) return false;
    if (Array.isArray(a) !== Array.isArray(b)) return false;
    var i, keys;
    if (Array.isArray(a)) {
      if (a.length !== b.length) return false;
      for (i = 0; i < a.length; i++) if (!equalValue(a[i], b[i], depth + 1)) return false;
      return true;
    }
    keys = Object.keys(a);
    if (keys.length !== Object.keys(b).length) return false;
    for (i = 0; i < keys.length; i++) {
      if (!hasOwn(b, keys[i])) return false;
      if (!equalValue(a[keys[i]], b[keys[i]], depth + 1)) return false;
    }
    return true;
  }
  function equal(a, b) {
    try { return ok(equalValue(a, b, 0)); }
    catch (e) { return fail('json.equal: ' + (e && e.message ? e.message : String(e))); }
  }
  function parsePointer(pointer) {
    if (typeof pointer !== 'string') throw new Error('JSON Pointer must be a string');
    if (pointer === '') return [];
    if (pointer.charAt(0) !== '/') throw new Error('JSON Pointer must be empty or start with "/"');
    var raw = pointer.slice(1).split('/'), out = [];
    for (var i = 0; i < raw.length; i++) {
      var token = raw[i];
      if (/~(?![01])/.test(token)) throw new Error('invalid JSON Pointer escape in "' + token + '"');
      out.push(token.replace(/~1/g, '/').replace(/~0/g, '~'));
    }
    return out;
  }
  function pointerString(tokens) {
    if (!tokens || tokens.length === 0) return '';
    var parts = [];
    for (var i = 0; i < tokens.length; i++) {
      parts.push(String(tokens[i]).replace(/~/g, '~0').replace(/\//g, '~1'));
    }
    return '/' + parts.join('/');
  }
  function arrayIndex(token, length, allowEnd) {
    if (token === '-') {
      if (allowEnd) return length;
      throw new Error('"-" is only valid as the last token of an add operation');
    }
    if (!/^(0|[1-9][0-9]*)$/.test(token)) throw new Error('invalid array index "' + token + '"');
    var index = Number(token);
    if (index > length || (!allowEnd && index === length)) throw new Error('array index ' + index + ' is out of range');
    return index;
  }
  function getTokens(root, tokens, pointerText) {
    var current = root;
    for (var i = 0; i < tokens.length; i++) {
      var token = tokens[i];
      if (Array.isArray(current)) current = current[arrayIndex(token, current.length, false)];
      else if (isObject(current)) {
        if (!hasOwn(current, token)) throw new Error('path not found at "' + pointerString(tokens.slice(0, i + 1)) + '"');
        current = current[token];
      } else throw new Error('cannot traverse "' + pointerText + '" through a scalar');
    }
    return current;
  }
  function getPointer(root, pointer) {
    try {
      var tokens = parsePointer(pointer);
      return ok(getTokens(root, tokens, pointer));
    } catch (e) { return fail('json.pointer: ' + (e && e.message ? e.message : String(e))); }
  }
  function addTokens(root, tokens, value) {
    if (tokens.length === 0) return value;
    var parent = getTokens(root, tokens.slice(0, -1), pointerString(tokens));
    var last = tokens[tokens.length - 1];
    if (Array.isArray(parent)) {
      var index = arrayIndex(last, parent.length, true);
      parent.splice(index, 0, value);
    } else if (isObject(parent)) {
      setKey(parent, last, value);
    } else throw new Error('cannot add a value under a scalar');
    return root;
  }
  function replaceTokens(root, tokens, value) {
    if (tokens.length === 0) return value;
    var parent = getTokens(root, tokens.slice(0, -1), pointerString(tokens));
    var last = tokens[tokens.length - 1];
    if (Array.isArray(parent)) {
      var index = arrayIndex(last, parent.length, false);
      parent[index] = value;
    } else if (isObject(parent)) {
      if (!hasOwn(parent, last)) throw new Error('path not found at "' + pointerString(tokens) + '"');
      setKey(parent, last, value);
    } else throw new Error('cannot replace a value under a scalar');
    return root;
  }
  function removeTokens(root, tokens) {
    if (tokens.length === 0) throw new Error('cannot remove the document root');
    var parent = getTokens(root, tokens.slice(0, -1), pointerString(tokens));
    var last = tokens[tokens.length - 1];
    if (Array.isArray(parent)) {
      var index = arrayIndex(last, parent.length, false);
      parent.splice(index, 1);
    } else if (isObject(parent)) {
      if (!hasOwn(parent, last)) throw new Error('path not found at "' + pointerString(tokens) + '"');
      delete parent[last];
    } else throw new Error('cannot remove a value under a scalar');
    return root;
  }
  function pointer(value, path) { return getPointer(value, path); }
  function setPointer(value, path, newValue) {
    try {
      var copy = cloneValue(value, 0);
      var tokens = parsePointer(path);
      var exists = true;
      try { getTokens(copy, tokens, path); } catch (e) { exists = false; }
      var next = cloneValue(newValue, 0);
      copy = exists ? replaceTokens(copy, tokens, next) : addTokens(copy, tokens, next);
      return ok(copy);
    } catch (e) { return fail('json.setPointer: ' + (e && e.message ? e.message : String(e))); }
  }
  function removePointer(value, path) {
    try {
      var copy = cloneValue(value, 0);
      removeTokens(copy, parsePointer(path));
      return ok(copy);
    } catch (e) { return fail('json.removePointer: ' + (e && e.message ? e.message : String(e))); }
  }
  function requireValue(op, index) {
    if (!hasOwn(op, 'value')) throw new Error('operation ' + (index + 1) + ' requires "value"');
    return cloneValue(op.value, 0);
  }
  function patch(value, operations) {
    try {
      if (typeof operations === 'string') {
        var parsed = parse(operations);
        if (parsed[1] !== null) return [null, 'json.patch: ' + parsed[1]];
        operations = parsed[0];
      }
      if (!Array.isArray(operations)) throw new Error('operations must be an array');
      if (operations.length > 10000) throw new Error('too many operations (max 10000)');
      var root = cloneValue(value, 0);
      for (var i = 0; i < operations.length; i++) {
        var op = operations[i];
        if (!op || typeof op !== 'object' || Array.isArray(op)) throw new Error('operation ' + (i + 1) + ' must be an object');
        var kind = op.op;
        if (typeof kind !== 'string') throw new Error('operation ' + (i + 1) + ' requires an "op" string');
        if (typeof op.path !== 'string') throw new Error('operation ' + (i + 1) + ' requires a "path" string');
        var tokens = parsePointer(op.path);
        if (kind === 'add') root = addTokens(root, tokens, requireValue(op, i));
        else if (kind === 'replace') root = replaceTokens(root, tokens, requireValue(op, i));
        else if (kind === 'remove') root = removeTokens(root, tokens);
        else if (kind === 'test') {
          var actual = getTokens(root, tokens, op.path);
          if (!equalValue(actual, op.value, 0)) throw new Error('test failed at "' + op.path + '"');
        } else if (kind === 'copy') {
          if (typeof op.from !== 'string') throw new Error('operation ' + (i + 1) + ' requires a "from" string');
          var copied = cloneValue(getTokens(root, parsePointer(op.from), op.from), 0);
          root = addTokens(root, tokens, copied);
        } else if (kind === 'move') {
          if (typeof op.from !== 'string') throw new Error('operation ' + (i + 1) + ' requires a "from" string');
          if (op.from === op.path) continue;
          var fromTokens = parsePointer(op.from);
          if (fromTokens.length === 0) throw new Error('cannot move the document root');
          if (op.path.indexOf(op.from + '/') === 0) throw new Error('cannot move a value into its own child');
          var moved = getTokens(root, fromTokens, op.from);
          root = removeTokens(root, fromTokens);
          root = addTokens(root, tokens, moved);
        } else throw new Error('operation ' + (i + 1) + ' has unsupported op "' + kind + '"');
      }
      return ok(root);
    } catch (e) { return fail('json.patch: ' + (e && e.message ? e.message : String(e))); }
  }
  function keyMeta(item, keyName) {
    if (!keyName || item === null || typeof item !== 'object' || Array.isArray(item)) return undefined;
    if (!hasOwn(item, keyName)) return undefined;
    var value = item[keyName];
    if (value === null || typeof value === 'object') return undefined;
    return value;
  }
  function itemKey(item, keyName) {
    var value = keyMeta(item, keyName);
    return value === undefined ? null : typeof value + ':' + String(value);
  }
  function diff(a, b, options) {
    var opts = options && typeof options === 'object' ? options : {};
    var keyName = typeof opts.key === 'string' ? opts.key : '';
    var detectMoves = opts.detectMoves !== false;
    var changes = [];
    function push(change) {
      if (changes.length >= 10000) throw new Error('diff is too large (more than 10000 operations)');
      changes.push(change);
    }
    function at(tokens) { return pointerString(tokens); }
    function addOp(tokens, value, keyValue) {
      var change = { op: 'add', path: at(tokens), value: cloneValue(value, 0) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function removeOp(tokens, keyValue) {
      var change = { op: 'remove', path: at(tokens) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function replaceOp(tokens, value, keyValue) {
      var change = { op: 'replace', path: at(tokens), value: cloneValue(value, 0) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function moveOp(fromTokens, toTokens, keyValue) {
      var change = { op: 'move', from: at(fromTokens), path: at(toTokens) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function compare(x, y, tokens) {
      if (equalValue(x, y, 0)) return;
      var xo = x !== null && typeof x === 'object';
      var yo = y !== null && typeof y === 'object';
      if (!xo || !yo || Array.isArray(x) !== Array.isArray(y)) { replaceOp(tokens, y); return; }
      if (Array.isArray(x)) compareArrays(x, y, tokens);
      else compareObjects(x, y, tokens);
    }
    function compareObjects(x, y, tokens) {
      var keysX = Object.keys(x), keysY = Object.keys(y), i, key;
      for (i = 0; i < keysX.length; i++) {
        key = keysX[i];
        if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
        if (!hasOwn(y, key)) removeOp(tokens.concat([key]));
      }
      for (i = 0; i < keysY.length; i++) {
        key = keysY[i];
        if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
        if (!hasOwn(x, key)) addOp(tokens.concat([key]), y[key]);
        else compare(x[key], y[key], tokens.concat([key]));
      }
    }
    function compareArrays(x, y, tokens) {
      if (!keyName) {
        var common = Math.min(x.length, y.length), i;
        for (i = 0; i < common; i++) compare(x[i], y[i], tokens.concat([String(i)]));
        if (x.length > y.length) {
          for (i = x.length - 1; i >= y.length; i--) removeOp(tokens.concat([String(i)]));
        } else if (y.length > x.length) {
          for (i = common; i < y.length; i++) addOp(tokens.concat([String(i)]), y[i]);
        }
        return;
      }
      var queues = Object.create(null), j, k;
      for (j = 0; j < y.length; j++) {
        k = itemKey(y[j], keyName);
        if (k !== null) (queues[k] || (queues[k] = [])).push(j);
      }
      var entries = [], used = [];
      for (j = 0; j < x.length; j++) {
        k = itemKey(x[j], keyName);
        var match = -1;
        if (k !== null && queues[k] && queues[k].length > 0) match = queues[k].shift();
        entries.push({ a: x[j], b: match >= 0 ? y[match] : null, bIndex: match, key: k });
        if (match >= 0) used[match] = true;
      }
      for (var r = entries.length - 1; r >= 0; r--) {
        if (!entries[r].b) {
          removeOp(tokens.concat([String(r)]), keyMeta(entries[r].a, keyName));
          entries.splice(r, 1);
        }
      }
      var reordered = false;
      for (var t = 0; t < y.length; t++) {
        if (used[t]) {
          var current = -1;
          for (var e = 0; e < entries.length; e++) if (entries[e].bIndex === t) { current = e; break; }
          if (current !== t) {
            reordered = true;
            if (!detectMoves) { replaceOp(tokens, y); return; }
            moveOp(tokens.concat([String(current)]), tokens.concat([String(t)]), keyMeta(y[t], keyName));
            var moved = entries.splice(current, 1)[0];
            entries.splice(t, 0, moved);
          }
        } else {
          addOp(tokens.concat([String(t)]), y[t], keyMeta(y[t], keyName));
          entries.splice(t, 0, { a: null, b: y[t], bIndex: t, key: itemKey(y[t], keyName) });
        }
      }
      for (var n = 0; n < entries.length; n++) {
        if (entries[n].a !== null && entries[n].b !== null) compare(entries[n].a, entries[n].b, tokens.concat([String(n)]));
      }
    }
    try {
      compare(a, b, []);
      return ok(changes);
    } catch (e) { return fail('json.diff: ' + (e && e.message ? e.message : String(e))); }
  }
  function merge(target, patchValue) {
    try {
      function apply(current, patchDoc) {
        if (!isObject(patchDoc)) return cloneValue(patchDoc, 0);
        var base = isObject(current) ? cloneValue(current, 0) : {};
        var keys = Object.keys(patchDoc);
        for (var i = 0; i < keys.length; i++) {
          var key = keys[i];
          if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
          if (patchDoc[key] === null) delete base[key];
          else setKey(base, key, apply(base[key], patchDoc[key]));
        }
        return base;
      }
      return ok(apply(target, patchValue));
    } catch (e) { return fail('json.merge: ' + (e && e.message ? e.message : String(e))); }
  }
  function tryGet(root, path) {
    try { return { found: true, value: getTokens(root, parsePointer(path), path) }; }
    catch (e) { return { found: false, value: null }; }
  }
  function replaceInPlace(target, merged) {
    if (Array.isArray(target) && Array.isArray(merged)) {
      target.length = 0;
      for (var i = 0; i < merged.length; i++) target.push(merged[i]);
      return;
    }
    if (isObject(target) && isObject(merged)) {
      var keys = Object.keys(target), j;
      for (j = 0; j < keys.length; j++) delete target[keys[j]];
      keys = Object.keys(merged);
      for (j = 0; j < keys.length; j++) setKey(target, keys[j], merged[keys[j]]);
      return;
    }
    throw new Error('requires object or array roots that keep their type');
  }
  function merge3(base, ours, theirs, options) {
    var changes = diff(base, theirs, options);
    if (changes[1] !== null) return [[], changes[1]];
    var conflicts = [], safe = [];
    for (var i = 0; i < changes[0].length; i++) {
      var op = changes[0][i];
      var mine = tryGet(ours, op.path);
      var before = tryGet(base, op.path);
      var conflict = false;
      if (op.op === 'replace') {
        if (!mine.found) conflict = before.found;
        else if (equalValue(mine.value, op.value, 0)) continue;
        else if (before.found && equalValue(mine.value, before.value, 0)) conflict = false;
        else conflict = true;
      } else if (op.op === 'remove') {
        if (!mine.found) continue;
        if (before.found && equalValue(mine.value, before.value, 0)) conflict = false;
        else conflict = true;
      } else if (op.op === 'add') {
        if (mine.found) {
          if (equalValue(mine.value, op.value, 0)) continue;
          conflict = true;
        }
      } else if (op.op === 'move') {
        var mineFrom = tryGet(ours, op.from);
        var baseFrom = tryGet(base, op.from);
        if (!mineFrom.found) conflict = true;
        else if (baseFrom.found && !equalValue(mineFrom.value, baseFrom.value, 0)) conflict = true;
      } else if (op.op === 'test') {
        if (mine.found && equalValue(mine.value, op.value, 0)) continue;
        conflict = true;
      }
      if (conflict) conflicts.push(op.op + ' at ' + op.path);
      else safe.push(op);
    }
    var applied = patch(ours, safe);
    if (applied[1] !== null) return [[], applied[1]];
    try {
      if (!equalValue(applied[0], ours, 0)) replaceInPlace(ours, applied[0]);
    } catch (e) { return [[], 'json.merge3: ' + (e && e.message ? e.message : String(e))]; }
    return ok(conflicts);
  }
  function parse(text) {
    try {
      if (typeof text !== 'string') throw new Error('json.parse expects a string');
      if (text.length > MAX_BYTES) throw new Error('JSON text is too large (max ' + MAX_BYTES + ' characters)');
      if (text.charCodeAt(0) === 0xFEFF) text = text.slice(1);
      var value = JSON.parse(text);
      return ok(value);
    } catch (e) { return fail('json.parse: ' + (e && e.message ? e.message : String(e))); }
  }
  function decodeBytes(bytes, encoding) {
    if (!bytes || typeof bytes.length !== 'number') throw new Error('expected a byte array');
    if (bytes.length > MAX_BYTES) throw new Error('JSON file is too large (max ' + MAX_BYTES + ' bytes)');
    var data = new Uint8Array(bytes.length);
    for (var i = 0; i < bytes.length; i++) {
      var value = Number(bytes[i]);
      if (!isFinite(value) || value < 0 || value > 255) throw new Error('invalid byte value at index ' + i);
      data[i] = value;
    }
    if (data.length >= 4 && data[0] === 0x00 && data[1] === 0x00 && data[2] === 0xFE && data[3] === 0xFF) throw new Error('UTF-32BE JSON is not supported');
    if (data.length >= 4 && data[0] === 0xFF && data[1] === 0xFE && data[2] === 0x00 && data[3] === 0x00) throw new Error('UTF-32LE JSON is not supported');
    var label = encoding ? String(encoding).toLowerCase().replace(/_/g, '-') : '';
    if (!label) {
      if (data.length >= 3 && data[0] === 0xEF && data[1] === 0xBB && data[2] === 0xBF) label = 'utf-8';
      else if (data.length >= 2 && data[0] === 0xFF && data[1] === 0xFE) label = 'utf-16le';
      else if (data.length >= 2 && data[0] === 0xFE && data[1] === 0xFF) label = 'utf-16be';
      else label = 'utf-8';
    }
    if (label === 'utf8') label = 'utf-8';
    if (label === 'utf16le') label = 'utf-16le';
    if (label === 'utf16be') label = 'utf-16be';
    if (label !== 'utf-8' && label !== 'utf-16le' && label !== 'utf-16be') throw new Error('unsupported JSON encoding "' + label + '"');
    if (label === 'utf-16le' || label === 'utf-16be') {
      var little = label === 'utf-16le';
      var offset = 0;
      if (data.length >= 2 && ((little && data[0] === 0xFF && data[1] === 0xFE) || (!little && data[0] === 0xFE && data[1] === 0xFF))) offset = 2;
      var text = '';
      for (var j = offset; j + 1 < data.length; j += 2) {
        var code = little ? data[j] | (data[j + 1] << 8) : (data[j] << 8) | data[j + 1];
        text += String.fromCharCode(code);
      }
      return text;
    }
    if (typeof TextDecoder === 'undefined') throw new Error('TextDecoder is not available');
    return new TextDecoder('utf-8').decode(data).replace(/^\uFEFF/, '');
  }
  function parseBytes(bytes, encoding) {
    try {
      var text = decodeBytes(bytes, encoding);
      return parse(text);
    } catch (e) { return fail('json.parseBytes: ' + (e && e.message ? e.message : String(e))); }
  }
  function encodeBytes(text, encoding) {
    var label = encoding ? String(encoding).toLowerCase().replace(/_/g, '-') : 'utf-8';
    if (label === 'utf8') label = 'utf-8';
    if (label === 'utf16le') label = 'utf-16le';
    if (label === 'utf16be') label = 'utf-16be';
    if (label === 'utf-8') {
      if (typeof TextEncoder === 'undefined') throw new Error('TextEncoder is not available');
      return Array.prototype.slice.call(new TextEncoder().encode(text));
    }
    if (label !== 'utf-16le' && label !== 'utf-16be') throw new Error('unsupported JSON encoding "' + label + '"');
    var little = label === 'utf-16le', out = [0xFF, 0xFE];
    if (!little) out = [0xFE, 0xFF];
    for (var i = 0; i < text.length; i++) {
      var code = text.charCodeAt(i);
      if (little) { out.push(code & 255, (code >> 8) & 255); }
      else { out.push((code >> 8) & 255, code & 255); }
    }
    return out;
  }
  function toBytes(value, encoding, options) {
    var result = stringify(value, options);
    if (result[1] !== null) return result;
    try { return ok(encodeBytes(result[0], encoding || 'utf-8')); }
    catch (e) { return fail('json.toBytes: ' + (e && e.message ? e.message : String(e))); }
  }
  function stringify(value, options) {
    try {
      var opts = options && typeof options === 'object' ? options : {};
      var indent = opts.indent;
      if (indent === true) indent = 2;
      if (typeof indent === 'number') indent = Math.max(0, Math.min(10, Math.floor(indent)));
      else if (typeof indent === 'string') indent = indent.slice(0, 10);
      else indent = 0;
      var valueToWrite = value;
      if (opts.sortKeys) {
        function sort(v) {
          if (Array.isArray(v)) return v.map(sort);
          if (!isObject(v)) return v;
          var out = {}, keys = Object.keys(v).sort();
          for (var i = 0; i < keys.length; i++) setKey(out, keys[i], sort(v[keys[i]]));
          return out;
        }
        valueToWrite = sort(value);
      }
      var text = JSON.stringify(valueToWrite, null, indent || undefined);
      if (typeof text !== 'string') throw new Error('value is not JSON serializable');
      if (opts.newline === true) text += '\n';
      return ok(text);
    } catch (e) { return fail('json.stringify: ' + (e && e.message ? e.message : String(e))); }
  }
  function validateAll(value, schema) {
    var errors = [];
    function add(path, message) { if (errors.length < MAX_ERRORS) errors.push(path + ': ' + message); }
    try { if (typeof schema === 'string') schema = JSON.parse(schema); }
    catch (e) { return [[], '$: schema is not valid JSON']; }
    function check(v, s, path, depth) {
      if (errors.length >= MAX_ERRORS) return;
      if (depth > 64) { add(path, 'schema nesting is too deep'); return; }
      if (s === true) return;
      if (s === false) { add(path, 'rejected by false schema'); return; }
      if (s === null || typeof s !== 'object' || Array.isArray(s)) { add(path, 'schema must be an object or boolean'); return; }
      if (s.$ref !== undefined) { add(path, '$ref is not supported'); return; }
      var i, key;
      if (s.type !== undefined) {
        var types = Array.isArray(s.type) ? s.type : [s.type], typeOk = false;
        for (i = 0; i < types.length; i++) if (__sui_schema_type_ok(v, types[i])) { typeOk = true; break; }
        if (!typeOk) add(path, 'expected ' + types.join(' or ') + ', got ' + __sui_schema_kind(v));
      }
      if (s.enum !== undefined) {
        if (!Array.isArray(s.enum)) add(path, 'enum must be an array');
        else {
          var found = false;
          for (i = 0; i < s.enum.length; i++) if (__sui_schema_equal(v, s.enum[i])) { found = true; break; }
          if (!found) add(path, 'value is not in enum');
        }
      }
      if (hasOwn(s, 'const') && !__sui_schema_equal(v, s.const)) add(path, 'value does not match const');
      if (typeof v === 'number') {
        if (typeof s.minimum === 'number' && v < s.minimum) add(path, 'must be >= ' + s.minimum);
        if (typeof s.maximum === 'number' && v > s.maximum) add(path, 'must be <= ' + s.maximum);
        if (typeof s.exclusiveMinimum === 'number' && v <= s.exclusiveMinimum) add(path, 'must be > ' + s.exclusiveMinimum);
        if (typeof s.exclusiveMaximum === 'number' && v >= s.exclusiveMaximum) add(path, 'must be < ' + s.exclusiveMaximum);
        if (typeof s.multipleOf === 'number' && s.multipleOf > 0) {
          var q = v / s.multipleOf;
          if (Math.abs(q - Math.round(q)) > 1e-9) add(path, 'must be a multiple of ' + s.multipleOf);
        }
      }
      if (typeof v === 'string') {
        if (typeof s.minLength === 'number' && v.length < s.minLength) add(path, 'must have at least ' + s.minLength + ' characters');
        if (typeof s.maxLength === 'number' && v.length > s.maxLength) add(path, 'must have at most ' + s.maxLength + ' characters');
        if (typeof s.pattern === 'string') {
          if (s.pattern.length > 512) add(path, 'pattern is too long');
          else { try { if (!(new RegExp(s.pattern)).test(v)) add(path, 'does not match the pattern'); } catch (e) { add(path, 'invalid pattern'); } }
        }
        if (typeof s.format === 'string') {
          if (s.format === 'email' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) add(path, 'must be an email');
          if ((s.format === 'uri' || s.format === 'url') && !__sui_valid_url(v)) add(path, 'must be a URI');
          if (s.format === 'uuid' && !/^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$/.test(v)) add(path, 'must be a UUID');
          if (s.format === 'date' && !/^\d{4}-\d{2}-\d{2}$/.test(v)) add(path, 'must be a date');
          if (s.format === 'date-time' && isNaN(Date.parse(v))) add(path, 'must be a date-time');
        }
      }
      if (Array.isArray(v)) {
        if (typeof s.minItems === 'number' && v.length < s.minItems) add(path, 'must have at least ' + s.minItems + ' items');
        if (typeof s.maxItems === 'number' && v.length > s.maxItems) add(path, 'must have at most ' + s.maxItems + ' items');
        if (s.uniqueItems === true) {
          var seen = {}, duplicates = 0;
          for (i = 0; i < v.length; i++) {
            var ukey = JSON.stringify(v[i]);
            if (seen[ukey]) { duplicates++; break; }
            seen[ukey] = true;
          }
          if (duplicates) add(path, 'items must be unique');
        }
        if (Array.isArray(s.items)) {
          for (i = 0; i < s.items.length && i < v.length; i++) check(v[i], s.items[i], path + '[' + i + ']', depth + 1);
          if (v.length > s.items.length) {
            if (s.additionalItems === false) add(path, 'additional items are not allowed');
            else if (s.additionalItems && typeof s.additionalItems === 'object') {
              for (i = s.items.length; i < v.length; i++) check(v[i], s.additionalItems, path + '[' + i + ']', depth + 1);
            }
          }
        } else if (s.items !== undefined) {
          for (i = 0; i < v.length; i++) check(v[i], s.items, path + '[' + i + ']', depth + 1);
        }
      } else if (isObject(v)) {
        var keys = Object.keys(v);
        if (typeof s.minProperties === 'number' && keys.length < s.minProperties) add(path, 'must have at least ' + s.minProperties + ' properties');
        if (typeof s.maxProperties === 'number' && keys.length > s.maxProperties) add(path, 'must have at most ' + s.maxProperties + ' properties');
        if (Array.isArray(s.required)) {
          for (i = 0; i < s.required.length; i++) if (!hasOwn(v, s.required[i])) add(path, 'missing required property "' + s.required[i] + '"');
        }
        var props = s.properties && typeof s.properties === 'object' ? s.properties : {};
        for (i = 0; i < keys.length; i++) {
          key = keys[i];
          if (props[key] !== undefined) check(v[key], props[key], path + '.' + key, depth + 1);
          else if (s.additionalProperties === false) add(path, 'unexpected property "' + key + '"');
          else if (s.additionalProperties && typeof s.additionalProperties === 'object') check(v[key], s.additionalProperties, path + '.' + key, depth + 1);
        }
      }
      if (Array.isArray(s.allOf)) for (i = 0; i < s.allOf.length; i++) check(v, s.allOf[i], path, depth + 1);
      if (Array.isArray(s.anyOf)) {
        var anyOk = false;
        for (i = 0; i < s.anyOf.length; i++) if (__sui_schema_check(v, s.anyOf[i], path, depth + 1) === null) { anyOk = true; break; }
        if (!anyOk) add(path, 'does not match anyOf');
      }
      if (Array.isArray(s.oneOf)) {
        var oneCount = 0;
        for (i = 0; i < s.oneOf.length; i++) if (__sui_schema_check(v, s.oneOf[i], path, depth + 1) === null) oneCount++;
        if (oneCount !== 1) add(path, 'must match exactly one oneOf schema');
      }
      if (s.not !== undefined && __sui_schema_check(v, s.not, path, depth + 1) === null) add(path, 'must not match the not schema');
      if (s.if !== undefined) {
        var condition = __sui_schema_check(v, s.if, path, depth + 1);
        if (condition === null && s.then !== undefined) check(v, s.then, path, depth + 1);
        else if (condition !== null && s.else !== undefined) check(v, s.else, path, depth + 1);
      }
    }
    check(value, schema, '$', 0);
    return ok(errors);
  }
  return {
    parse: parse,
    parseBytes: parseBytes,
    decodeBytes: decodeBytes,
    encodeBytes: encodeBytes,
    stringify: stringify,
    toBytes: toBytes,
    validateAll: validateAll,
    pointer: pointer,
    setPointer: setPointer,
    removePointer: removePointer,
    clone: clone,
    equal: equal,
    patch: patch,
    diff: diff,
    merge: merge,
    merge3: merge3
  };
})();
`

// jsonModuleImported reports whether the program imports the `json` std module.
func jsonModuleImported(p *Prog) bool {
	if p == nil {
		return false
	}
	for i := 0; i < len(p.imports); i++ {
		im := p.imports[i]
		if im.kind != "std" {
			continue
		}
		for j := 0; j < len(im.names); j++ {
			if strings.TrimSpace(im.names[j]) == "json" {
				return true
			}
		}
	}
	return false
}
