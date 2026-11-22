package sui

import "strings"

// stdlib_extra_runtime.go - browser/desktop implementation of the pure and
// in-memory std modules registered in stdlib_extra.go.
//
// The prelude is emitted only when one of those modules is imported. It has no
// network, filesystem or process-environment access. All methods return the
// standard [value, error] pair. Data is always treated as data: ZIP entry names
// are validated against traversal/symlinks, CSV exports escape spreadsheet
// formulas, logs redact common secret names and env exposes a fixed read-only
// set of runtime facts.

const stdlibExtraRuntimeRaw = `var __SUI_STD_EXTRA = (function () {
  var TARGET = "__SUI_TARGET__";
  var MAX_BYTES = 10000000;
  var MAX_ZIP_TOTAL = 50000000;
  var MAX_ENTRIES = 1000;
  var MAX_NAME_BYTES = 512;

  function ok(value) { return [value, null]; }
  function okNull() { return [null, null]; }
  function fail(message) { return [null, String(message == null ? 'error' : message).slice(0, 500)]; }
  function failFalse(message) { return [false, String(message == null ? 'error' : message).slice(0, 500)]; }
  function failZero(message) { return [0, String(message == null ? 'error' : message).slice(0, 500)]; }
  function failEmpty(message) { return ['', String(message == null ? 'error' : message).slice(0, 500)]; }
  function msg(e) { return e && e.message ? e.message : String(e); }
  function isObj(v) { return v !== null && typeof v === 'object' && !Array.isArray(v); }
  function setKey(o, k, v) {
    k = String(k);
    if (k === '__proto__') throw new Error('unsafe object key "__proto__"');
    Object.defineProperty(o, k, { value: v, writable: true, enumerable: true, configurable: true });
  }
  function cloneValue(v, depth, seen) {
    if (v === null || typeof v !== 'object') return v;
    depth = depth || 0;
    if (depth > 64) throw new Error('value is nested too deeply');
    seen = seen || [];
    if (seen.indexOf(v) >= 0) throw new Error('value contains a cycle');
    seen.push(v);
    var out, keys, i;
    if (Array.isArray(v)) {
      if (v.length > 10000) throw new Error('array is too large');
      out = [];
      for (i = 0; i < v.length; i++) out.push(cloneValue(v[i], depth + 1, seen));
    } else {
      keys = Object.keys(v);
      if (keys.length > 1000) throw new Error('object is too large');
      out = {};
      for (i = 0; i < keys.length; i++) setKey(out, keys[i], cloneValue(v[keys[i]], depth + 1, seen));
    }
    seen.pop();
    return out;
  }
  function safeClone(v) { try { return cloneValue(v, 0, []); } catch (e) { return v; } }
  function number(v, label) {
    var n = Number(v);
    if (!isFinite(n)) throw new Error(label + ' must be a finite number');
    return n;
  }
  function integer(v, label) { return Math.trunc(number(v, label)); }
  function clamp(n, lo, hi) { return Math.min(hi, Math.max(lo, n)); }
  function opt(v) { return isObj(v) ? v : {}; }
  function optString(o, k, d) { return o[k] == null ? d : String(o[k]); }
  function optNumber(o, k, d) { return o[k] == null ? d : number(o[k], k); }
  function optBool(o, k, d) {
    if (o[k] == null) return d;
    if (typeof o[k] === 'string') return o[k] === 'true' || o[k] === '1';
    return !!o[k];
  }
  function optList(value) {
    if (Array.isArray(value)) return value.map(function (x) { return String(x).trim().toLowerCase(); }).filter(Boolean);
    if (typeof value === 'string') return value.split(/[,\s]+/).map(function (x) { return x.trim().toLowerCase(); }).filter(Boolean);
    return [];
  }
  function bytesOf(value, label) {
    var i, n;
    if (value instanceof Uint8Array) {
      if (value.length > MAX_BYTES) throw new Error(label + ' is too large');
      return value;
    }
    if (typeof ArrayBuffer !== 'undefined' && value instanceof ArrayBuffer) {
      if (value.byteLength > MAX_BYTES) throw new Error(label + ' is too large');
      return new Uint8Array(value);
    }
    if (!Array.isArray(value)) throw new Error(label + ' must be an array of bytes');
    if (value.length > MAX_BYTES) throw new Error(label + ' is too large');
    var out = new Uint8Array(value.length);
    for (i = 0; i < value.length; i++) {
      n = number(value[i], label + '[' + i + ']');
      if (!Number.isInteger(n) || n < 0 || n > 255) throw new Error(label + '[' + i + '] is not a byte');
      out[i] = n;
    }
    return out;
  }
  function outBytes(u8) { return Array.prototype.slice.call(u8); }
  function concatBytes(chunks, total) {
    var out = new Uint8Array(total == null ? chunks.reduce(function (a, c) { return a + c.length; }, 0) : total);
    var pos = 0;
    for (var i = 0; i < chunks.length; i++) { out.set(chunks[i], pos); pos += chunks[i].length; }
    return out;
  }
  function utf8Encode(text) { return new TextEncoder().encode(String(text)); }
  function utf8Decode(bytes) { return new TextDecoder('utf-8', { fatal: true }).decode(bytes); }
  function utf16Encode(text, little) {
    var s = String(text), out = new Uint8Array(s.length * 2);
    for (var i = 0; i < s.length; i++) {
      var c = s.charCodeAt(i);
      out[i * 2] = little ? (c & 255) : (c >> 8);
      out[i * 2 + 1] = little ? (c >> 8) : (c & 255);
    }
    return out;
  }
  function utf16Decode(bytes, little) {
    var chars = [];
    for (var i = 0; i + 1 < bytes.length; i += 2) {
      chars.push(String.fromCharCode(little ? (bytes[i] | (bytes[i + 1] << 8)) : ((bytes[i] << 8) | bytes[i + 1])));
    }
    return chars.join('');
  }
  function toBase64(u8) {
    var s = '';
    for (var i = 0; i < u8.length; i += 8192) s += String.fromCharCode.apply(null, u8.subarray(i, i + 8192));
    return btoa(s);
  }
  function fromBase64(text) {
    var s = atob(String(text).replace(/\s+/g, '')), out = new Uint8Array(s.length);
    for (var i = 0; i < s.length; i++) out[i] = s.charCodeAt(i);
    return out;
  }
  function toHex(u8) {
    var s = '';
    for (var i = 0; i < u8.length; i++) s += ('0' + u8[i].toString(16)).slice(-2);
    return s;
  }
  function fromHex(text) {
    var s = String(text).replace(/\s+/g, '');
    if (s.length % 2 !== 0 || !/^[0-9a-fA-F]*$/.test(s)) throw new Error('invalid hex string');
    var out = new Uint8Array(s.length / 2);
    for (var i = 0; i < out.length; i++) out[i] = parseInt(s.substr(i * 2, 2), 16);
    return out;
  }
  function byteCompare(a, b) {
    if (a.length !== b.length) return a.length - b.length;
    for (var i = 0; i < a.length; i++) if (a[i] !== b[i]) return a[i] - b[i];
    return 0;
  }

  var bytes = {
    fromString: function (text, encoding) {
      try {
        var enc = String(encoding || 'utf-8').toLowerCase();
        if (enc === 'utf-8' || enc === 'utf8') return ok(outBytes(utf8Encode(text)));
        if (enc === 'utf-16le' || enc === 'utf-16') return ok(outBytes(utf16Encode(text, true)));
        if (enc === 'utf-16be') return ok(outBytes(utf16Encode(text, false)));
        return fail('bytes.fromString: unsupported encoding ' + enc);
      } catch (e) { return fail('bytes.fromString: ' + msg(e)); }
    },
    toString: function (value, encoding) {
      try {
        var u = bytesOf(value, 'bytes.toString'), enc = String(encoding || 'utf-8').toLowerCase();
        if (enc === 'utf-8' || enc === 'utf8') return ok(utf8Decode(u));
        if (enc === 'utf-16le' || enc === 'utf-16') return ok(utf16Decode(u, true));
        if (enc === 'utf-16be') return ok(utf16Decode(u, false));
        return fail('bytes.toString: unsupported encoding ' + enc);
      } catch (e) { return fail('bytes.toString: ' + msg(e)); }
    },
    fromBase64: function (text) { try { return ok(outBytes(fromBase64(text))); } catch (e) { return fail('bytes.fromBase64: ' + msg(e)); } },
    toBase64: function (value) { try { return ok(toBase64(bytesOf(value, 'bytes.toBase64'))); } catch (e) { return fail('bytes.toBase64: ' + msg(e)); } },
    fromHex: function (text) { try { return ok(outBytes(fromHex(text))); } catch (e) { return fail('bytes.fromHex: ' + msg(e)); } },
    toHex: function (value) { try { return ok(toHex(bytesOf(value, 'bytes.toHex'))); } catch (e) { return fail('bytes.toHex: ' + msg(e)); } },
    concat: function (a, b) {
      try {
        var x = bytesOf(a, 'bytes.concat first argument'), y = bytesOf(b, 'bytes.concat second argument');
        if (x.length + y.length > MAX_BYTES) throw new Error('result is too large');
        return ok(outBytes(concatBytes([x, y])));
      } catch (e) { return fail('bytes.concat: ' + msg(e)); }
    },
    slice: function (value, start, end) {
      try {
        var u = bytesOf(value, 'bytes.slice'), s = integer(start, 'start'), e2 = end == null ? u.length : integer(end, 'end');
        if (s < 0 || e2 < s || e2 > u.length) throw new Error('slice range out of bounds');
        return ok(outBytes(u.subarray(s, e2)));
      } catch (e) { return fail('bytes.slice: ' + msg(e)); }
    },
    indexOf: function (value, needle, start) {
      try {
        var u = bytesOf(value, 'bytes.indexOf'), n = bytesOf(needle, 'needle'), from = start == null ? 0 : integer(start, 'start');
        if (from < 0 || from > u.length) throw new Error('start out of bounds');
        if (n.length === 0) return ok(from);
        for (var i = from; i + n.length <= u.length; i++) {
          var found = true;
          for (var j = 0; j < n.length; j++) if (u[i + j] !== n[j]) { found = false; break; }
          if (found) return ok(i);
        }
        return ok(-1);
      } catch (e) { return failZero('bytes.indexOf: ' + msg(e)); }
    },
    equal: function (a, b) { try { return ok(byteCompare(bytesOf(a, 'a'), bytesOf(b, 'b')) === 0); } catch (e) { return failFalse('bytes.equal: ' + msg(e)); } },
    length: function (value) { try { return ok(bytesOf(value, 'bytes.length').length); } catch (e) { return failZero('bytes.length: ' + msg(e)); } },
    fill: function (length, value) {
      try {
        var n = integer(length, 'length'), v = integer(value, 'byte');
        if (n < 0 || n > MAX_BYTES) throw new Error('length out of bounds');
        if (v < 0 || v > 255) throw new Error('byte out of bounds');
        return ok(outBytes(new Uint8Array(n).fill(v)));
      } catch (e) { return fail('bytes.fill: ' + msg(e)); }
    },
    reverse: function (value) {
      try {
        var out = new Uint8Array(bytesOf(value, 'bytes.reverse'));
        out.reverse();
        return ok(outBytes(out));
      } catch (e) { return fail('bytes.reverse: ' + msg(e)); }
    },
    xor: function (a, b) {
      try {
        var x = bytesOf(a, 'a'), y = bytesOf(b, 'b');
        if (x.length !== y.length) throw new Error('byte arrays must have the same length');
        var out = new Uint8Array(x.length);
        for (var i = 0; i < x.length; i++) out[i] = x[i] ^ y[i];
        return ok(outBytes(out));
      } catch (e) { return fail('bytes.xor: ' + msg(e)); }
    }
  };

  function dataView(value, label) {
    var u = bytesOf(value, label);
    return [new DataView(u.buffer, u.byteOffset, u.byteLength), u];
  }
  function checkOffset(u, offset, size) {
    var o = integer(offset, 'offset');
    if (o < 0 || o + size > u.length) throw new Error('offset out of bounds');
    return o;
  }
  function littleOf(v) { return v === undefined ? true : !!v; }
  function writeArray(u8) { return ok(outBytes(u8)); }
  function packU16(v) { var a = new Uint8Array(2); new DataView(a.buffer).setUint16(0, v & 65535, true); return a; }
  function packU32(v) { var a = new Uint8Array(4); new DataView(a.buffer).setUint32(0, v >>> 0, true); return a; }

  var bin = {
    u8: function (value, offset) {
      try { var p = dataView(value, 'bin.u8'), o = checkOffset(p[1], offset, 1); return ok(p[0].getUint8(o)); }
      catch (e) { return failZero('bin.u8: ' + msg(e)); }
    },
    u16: function (value, offset, little) {
      try { var p = dataView(value, 'bin.u16'), o = checkOffset(p[1], offset, 2); return ok(p[0].getUint16(o, littleOf(little))); }
      catch (e) { return failZero('bin.u16: ' + msg(e)); }
    },
    u32: function (value, offset, little) {
      try { var p = dataView(value, 'bin.u32'), o = checkOffset(p[1], offset, 4); return ok(p[0].getUint32(o, littleOf(little))); }
      catch (e) { return failZero('bin.u32: ' + msg(e)); }
    },
    i8: function (value, offset) {
      try { var p = dataView(value, 'bin.i8'), o = checkOffset(p[1], offset, 1); return ok(p[0].getInt8(o)); }
      catch (e) { return failZero('bin.i8: ' + msg(e)); }
    },
    i16: function (value, offset, little) {
      try { var p = dataView(value, 'bin.i16'), o = checkOffset(p[1], offset, 2); return ok(p[0].getInt16(o, littleOf(little))); }
      catch (e) { return failZero('bin.i16: ' + msg(e)); }
    },
    i32: function (value, offset, little) {
      try { var p = dataView(value, 'bin.i32'), o = checkOffset(p[1], offset, 4); return ok(p[0].getInt32(o, littleOf(little))); }
      catch (e) { return failZero('bin.i32: ' + msg(e)); }
    },
    f32: function (value, offset, little) {
      try { var p = dataView(value, 'bin.f32'), o = checkOffset(p[1], offset, 4); return ok(p[0].getFloat32(o, littleOf(little))); }
      catch (e) { return failZero('bin.f32: ' + msg(e)); }
    },
    f64: function (value, offset, little) {
      try { var p = dataView(value, 'bin.f64'), o = checkOffset(p[1], offset, 8); return ok(p[0].getFloat64(o, littleOf(little))); }
      catch (e) { return failZero('bin.f64: ' + msg(e)); }
    },
    writeU8: function (value) {
      try { var a = new Uint8Array(1); a[0] = integer(value, 'value') & 255; return writeArray(a); }
      catch (e) { return fail('bin.writeU8: ' + msg(e)); }
    },
    writeU16: function (value, little) {
      try { var a = new Uint8Array(2); new DataView(a.buffer).setUint16(0, integer(value, 'value'), littleOf(little)); return writeArray(a); }
      catch (e) { return fail('bin.writeU16: ' + msg(e)); }
    },
    writeU32: function (value, little) {
      try { var a = new Uint8Array(4); new DataView(a.buffer).setUint32(0, integer(value, 'value'), littleOf(little)); return writeArray(a); }
      catch (e) { return fail('bin.writeU32: ' + msg(e)); }
    },
    writeI8: function (value) {
      try { var a = new Uint8Array(1); new DataView(a.buffer).setInt8(0, integer(value, 'value')); return writeArray(a); }
      catch (e) { return fail('bin.writeI8: ' + msg(e)); }
    },
    writeI16: function (value, little) {
      try { var a = new Uint8Array(2); new DataView(a.buffer).setInt16(0, integer(value, 'value'), littleOf(little)); return writeArray(a); }
      catch (e) { return fail('bin.writeI16: ' + msg(e)); }
    },
    writeI32: function (value, little) {
      try { var a = new Uint8Array(4); new DataView(a.buffer).setInt32(0, integer(value, 'value'), littleOf(little)); return writeArray(a); }
      catch (e) { return fail('bin.writeI32: ' + msg(e)); }
    },
    writeF32: function (value, little) {
      try { var a = new Uint8Array(4); new DataView(a.buffer).setFloat32(0, number(value, 'value'), littleOf(little)); return writeArray(a); }
      catch (e) { return fail('bin.writeF32: ' + msg(e)); }
    },
    writeF64: function (value, little) {
      try { var a = new Uint8Array(8); new DataView(a.buffer).setFloat64(0, number(value, 'value'), littleOf(little)); return writeArray(a); }
      catch (e) { return fail('bin.writeF64: ' + msg(e)); }
    }
  };

  function cleanPath(value) { return String(value == null ? '' : value).replace(/\\/g, '/'); }
  function baseName(value) { var p = cleanPath(value), parts = p.split('/'); return parts[parts.length - 1] || ''; }
  function extOf(value) {
    var n = baseName(value), i = n.lastIndexOf('.');
    return i > 0 ? n.slice(i).toLowerCase() : '';
  }
  function validateFileName(name, options) {
    var o = opt(options), n = String(name == null ? '' : name);
    var maxName = clamp(integer(o.maxName == null ? 200 : o.maxName, 'maxName'), 1, 1000);
    if (!n) throw new Error('file name is empty');
    if (n.length > maxName) throw new Error('file name is too long');
    if (/[\u0000-\u001f\u007f]/.test(n)) throw new Error('file name contains control characters');
    if (n.indexOf('/') >= 0 || n.indexOf('\\') >= 0) throw new Error('file name contains a path separator');
    if (n.indexOf('..') >= 0) throw new Error('file name contains a traversal segment');
    if (!optBool(o, 'allowHidden', false) && n.charAt(0) === '.') throw new Error('hidden file names are not allowed');
    var allowed = optList(o.allowedExt || o.allow);
    if (allowed.length) {
      var ext = extOf(n).replace(/^\./, '');
      if (allowed.indexOf(ext) < 0) throw new Error('file extension is not allowed');
    }
    var size = o.size == null ? -1 : number(o.size, 'size');
    var maxSize = o.maxSize == null ? -1 : number(o.maxSize, 'maxSize');
    if (size >= 0 && maxSize >= 0 && size > maxSize) throw new Error('file is too large');
    return true;
  }

  var files = {
    name: function (path) { try { return ok(baseName(path)); } catch (e) { return failEmpty('files.name: ' + msg(e)); } },
    stem: function (path) {
      try { var n = baseName(path), e = extOf(n); return ok(e ? n.slice(0, n.length - e.length) : n); }
      catch (e) { return failEmpty('files.stem: ' + msg(e)); }
    },
    ext: function (path) { try { return ok(extOf(path)); } catch (e) { return failEmpty('files.ext: ' + msg(e)); } },
    split: function (path) {
      try {
        var n = baseName(path), e = extOf(n), out = {};
        setKey(out, 'name', n);
        setKey(out, 'stem', e ? n.slice(0, n.length - e.length) : n);
        setKey(out, 'ext', e);
        return ok(out);
      } catch (e) { return fail('files.split: ' + msg(e)); }
    },
    sanitize: function (name, maxLen) {
      try {
        var n = integer(maxLen == null ? 120 : maxLen, 'maxLen');
        n = clamp(n, 1, 1000);
        var s = String(name == null ? '' : name)
          .replace(/[\u0000-\u001f\u007f<>:"|?*\\\/]/g, '_')
          .replace(/\.\.+/g, '.')
          .replace(/^\.+/, '')
          .replace(/\s+/g, ' ')
          .trim()
          .slice(0, n);
        return ok(s || 'file');
      } catch (e) { return failEmpty('files.sanitize: ' + msg(e)); }
    },
    validate: function (name, options) {
      try { return ok(validateFileName(name, options)); }
      catch (e) { return failFalse('files.validate: ' + msg(e)); }
    },
    isHidden: function (name) { try { return ok(baseName(name).charAt(0) === '.'); } catch (e) { return failFalse('files.isHidden: ' + msg(e)); } },
    extAllowed: function (name, allow) {
      try {
        var list = optList(allow);
        if (!list.length) return ok(true);
        return ok(list.indexOf(extOf(name).replace(/^\./, '')) >= 0);
      } catch (e) { return failFalse('files.extAllowed: ' + msg(e)); }
    },
    readBytes: function (name) {
      // The only data-module function that touches bytes: it reads the file
      // the user selected or dropped in this app, never a path and never the
      // network. The desktop target reads files through fs.readBytes.
      if (TARGET === 'desktop') {
        return Promise.resolve([null, 'files.readBytes is not available on the desktop target: use fs.readBytes(path)']);
      }
      if (typeof __sui_dnd_read_bytes !== 'function') {
        return Promise.resolve([null, 'files.readBytes needs a selection: add a file control with pickFiles or drop a file']);
      }
      return Promise.resolve(__sui_dnd_read_bytes(name)).then(function (r) {
        return r;
      }, function (e) {
        return [null, String(e && e.message ? e.message : e)];
      });
    },
    documentUrl: function (name) {
      // Builds a blob: URL for a file the user selected, so the PDF viewer
      // can show it without a backend. The bytes are checked (PDF magic)
      // and the type is forced: the native viewer must never receive a
      // document the browser could execute. The desktop target reads files
      // through the bridge (fs.documentUrl).
      if (TARGET === 'desktop') {
        return Promise.resolve([null, 'files.documentUrl is not available on the desktop target: use fs.documentUrl(path)']);
      }
      if (typeof __sui_dnd_read_bytes !== 'function') {
        return Promise.resolve([null, 'files.documentUrl needs a selection: add a file control with pickFiles or drop a file']);
      }
      return Promise.resolve(__sui_dnd_read_bytes(name)).then(function (r) {
        var bytes = r[0];
        var err = r[1];
        if (err !== null) return [null, err];
        if (!bytes || bytes.length < 5) return [null, 'the file is not a PDF document'];
        var head = '';
        for (var i = 0; i < 5; i++) head += String.fromCharCode(bytes[i]);
        if (head !== '%PDF-') return [null, 'the file is not a PDF document'];
        try {
          var blob = new Blob([new Uint8Array(bytes)], { type: 'application/pdf' });
          return [URL.createObjectURL(blob), null];
        } catch (e) {
          return [null, String(e && e.message ? e.message : e)];
        }
      }, function (e) {
        return [null, String(e && e.message ? e.message : e)];
      });
    }
  };

  var mimeByExt = {
    txt: 'text/plain', text: 'text/plain', md: 'text/markdown', csv: 'text/csv',
    json: 'application/json', xml: 'application/xml', html: 'text/html', htm: 'text/html',
    css: 'text/css', js: 'text/javascript', pdf: 'application/pdf',
    png: 'image/png', jpg: 'image/jpeg', jpeg: 'image/jpeg', gif: 'image/gif',
    webp: 'image/webp', svg: 'image/svg+xml', mp3: 'audio/mpeg', mp4: 'video/mp4',
    zip: 'application/zip', gz: 'application/gzip', docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    odt: 'application/vnd.oasis.opendocument.text', ods: 'application/vnd.oasis.opendocument.spreadsheet'
  };
  var extByMime = {
    'text/plain': '.txt', 'text/markdown': '.md', 'text/csv': '.csv',
    'application/json': '.json', 'application/xml': '.xml', 'text/html': '.html',
    'text/css': '.css', 'text/javascript': '.js', 'application/pdf': '.pdf',
    'image/png': '.png', 'image/jpeg': '.jpg', 'image/gif': '.gif',
    'image/webp': '.webp', 'image/svg+xml': '.svg', 'audio/mpeg': '.mp3',
    'video/mp4': '.mp4', 'application/zip': '.zip', 'application/gzip': '.gz'
  };
  function mimeAllowed(mime, allow) {
    var list = optList(allow), m = String(mime || '').toLowerCase();
    if (!list.length) return true;
    for (var i = 0; i < list.length; i++) {
      var a = list[i];
      if (a === '*/*' || a === m || (a.slice(-2) === '/*' && m.indexOf(a.slice(0, -1)) === 0)) return true;
    }
    return false;
  }

  var mime = {
    fromName: function (name) {
      try { return ok(mimeByExt[extOf(name).replace(/^\./, '')] || 'application/octet-stream'); }
      catch (e) { return failEmpty('mime.fromName: ' + msg(e)); }
    },
    fromBytes: function (value) {
      try {
        var u = bytesOf(value, 'mime.fromBytes');
        if (u.length >= 8 && u[0] === 137 && u[1] === 80 && u[2] === 78 && u[3] === 71) return ok('image/png');
        if (u.length >= 3 && u[0] === 255 && u[1] === 216 && u[2] === 255) return ok('image/jpeg');
        if (u.length >= 3 && u[0] === 71 && u[1] === 73 && u[2] === 70) return ok('image/gif');
        if (u.length >= 4 && u[0] === 37 && u[1] === 80 && u[2] === 68 && u[3] === 70) return ok('application/pdf');
        if (u.length >= 4 && u[0] === 80 && u[1] === 75 && u[2] === 3 && u[3] === 4) return ok('application/zip');
        if (u.length >= 3 && u[0] === 239 && u[1] === 187 && u[2] === 191) return ok('text/plain');
        var head = utf8Decode(u.subarray(0, Math.min(u.length, 64))).replace(/^\uFEFF/, '').trim();
        if (head.indexOf('<?xml') === 0) return ok('application/xml');
        if (head.charAt(0) === '{' || head.charAt(0) === '[') return ok('application/json');
        return ok('application/octet-stream');
      } catch (e) { return failEmpty('mime.fromBytes: ' + msg(e)); }
    },
    extension: function (mimeType) { try { return ok(extByMime[String(mimeType || '').toLowerCase()] || ''); } catch (e) { return failEmpty('mime.extension: ' + msg(e)); } },
    isAllowed: function (mimeType, allow) { try { return ok(mimeAllowed(mimeType, allow)); } catch (e) { return failFalse('mime.isAllowed: ' + msg(e)); } },
    validate: function (name, mimeType, size, options) {
      try {
        var o = opt(options);
        var merged = {
          maxName: o.maxName, allowedExt: o.allowedExt, allowHidden: o.allowHidden,
          size: size, maxSize: o.maxSize
        };
        validateFileName(name, merged);
        if (o.allow && !mimeAllowed(mimeType, o.allow)) throw new Error('MIME type is not allowed');
        return ok(true);
      } catch (e) { return failFalse('mime.validate: ' + msg(e)); }
    }
  };

  var namedColors = {
    transparent: [0, 0, 0, 0], black: [0, 0, 0, 1], white: [255, 255, 255, 1],
    red: [255, 0, 0, 1], green: [0, 128, 0, 1], blue: [0, 0, 255, 1],
    gray: [128, 128, 128, 1], grey: [128, 128, 128, 1], yellow: [255, 255, 0, 1],
    cyan: [0, 255, 255, 1], magenta: [255, 0, 255, 1], orange: [255, 165, 0, 1],
    purple: [128, 0, 128, 1], silver: [192, 192, 192, 1], navy: [0, 0, 128, 1]
  };
  function channel(tok) {
    var s = String(tok).trim(), n;
    if (s.charAt(s.length - 1) === '%') { n = parseFloat(s.slice(0, -1)) * 2.55; }
    else n = parseFloat(s);
    if (!isFinite(n)) throw new Error('invalid color channel');
    return clamp(Math.round(n), 0, 255);
  }
  function alphaOf(tok) {
    var s = tok == null ? '1' : String(tok).trim(), n;
    if (s.charAt(s.length - 1) === '%') n = parseFloat(s.slice(0, -1)) / 100;
    else n = parseFloat(s);
    if (!isFinite(n)) throw new Error('invalid alpha');
    return clamp(n, 0, 1);
  }
  function hslToRgb(h, s, l) {
    h = ((number(h, 'hue') % 360) + 360) % 360;
    s = clamp(number(s, 'saturation'), 0, 100) / 100;
    l = clamp(number(l, 'lightness'), 0, 100) / 100;
    var c = (1 - Math.abs(2 * l - 1)) * s;
    var x = c * (1 - Math.abs(((h / 60) % 2) - 1));
    var m = l - c / 2, r = 0, g = 0, b = 0;
    if (h < 60) { r = c; g = x; }
    else if (h < 120) { r = x; g = c; }
    else if (h < 180) { g = c; b = x; }
    else if (h < 240) { g = x; b = c; }
    else if (h < 300) { r = x; b = c; }
    else { r = c; b = x; }
    return [Math.round((r + m) * 255), Math.round((g + m) * 255), Math.round((b + m) * 255)];
  }
  function rgbToHsl(r, g, b) {
    r /= 255; g /= 255; b /= 255;
    var max = Math.max(r, g, b), min = Math.min(r, g, b), d = max - min;
    var h = 0, l = (max + min) / 2, s = d === 0 ? 0 : d / (1 - Math.abs(2 * l - 1));
    if (d !== 0) {
      if (max === r) h = 60 * (((g - b) / d) % 6);
      else if (max === g) h = 60 * ((b - r) / d + 2);
      else h = 60 * ((r - g) / d + 4);
    }
    if (h < 0) h += 360;
    return [h, s * 100, l * 100];
  }
  function parseColor(value) {
    if (isObj(value)) {
      return [channel(value.r), channel(value.g), channel(value.b), alphaOf(value.a == null ? 1 : value.a)];
    }
    var s = String(value == null ? '' : value).trim().toLowerCase();
    if (namedColors[s]) return namedColors[s].slice();
    var m = /^#([0-9a-f]{3,4}|[0-9a-f]{6}|[0-9a-f]{8})$/.exec(s);
    if (m) {
      var h = m[1];
      if (h.length === 3 || h.length === 4) {
        var r = parseInt(h.charAt(0) + h.charAt(0), 16), g = parseInt(h.charAt(1) + h.charAt(1), 16), b = parseInt(h.charAt(2) + h.charAt(2), 16);
        var a = h.length === 4 ? parseInt(h.charAt(3) + h.charAt(3), 16) / 255 : 1;
        return [r, g, b, a];
      }
      return [parseInt(h.slice(0, 2), 16), parseInt(h.slice(2, 4), 16), parseInt(h.slice(4, 6), 16), h.length === 8 ? parseInt(h.slice(6, 8), 16) / 255 : 1];
    }
    m = /^rgba?\(([^)]+)\)$/.exec(s);
    if (m) {
      var parts = m[1].split(',');
      if (parts.length !== 3 && parts.length !== 4) throw new Error('invalid rgb color');
      return [channel(parts[0]), channel(parts[1]), channel(parts[2]), parts.length === 4 ? alphaOf(parts[3]) : 1];
    }
    m = /^hsla?\(([^)]+)\)$/.exec(s);
    if (m) {
      var hp = m[1].split(',');
      if (hp.length !== 3 && hp.length !== 4) throw new Error('invalid hsl color');
      var rgb = hslToRgb(parseFloat(hp[0]), parseFloat(hp[1]), parseFloat(hp[2]));
      rgb.push(hp.length === 4 ? alphaOf(hp[3]) : 1);
      return rgb;
    }
    throw new Error('unsupported color value');
  }
  function colorObject(c) {
    var out = {};
    setKey(out, 'r', c[0]); setKey(out, 'g', c[1]); setKey(out, 'b', c[2]); setKey(out, 'a', c[3]);
    return out;
  }
  function hex2(n) { return ('0' + Math.round(n).toString(16)).slice(-2); }
  function formatColor(value, format) {
    var c = parseColor(value), f = String(format || 'hex').toLowerCase(), hsl;
    if (f === 'hex') {
      var s = '#' + hex2(c[0]) + hex2(c[1]) + hex2(c[2]);
      if (c[3] < 1) s += hex2(c[3] * 255);
      return s;
    }
    if (f === 'rgb') return 'rgb(' + c[0] + ', ' + c[1] + ', ' + c[2] + ')';
    if (f === 'rgba') return 'rgba(' + c[0] + ', ' + c[1] + ', ' + c[2] + ', ' + Number(c[3].toFixed(3)) + ')';
    if (f === 'hsl' || f === 'hsla') {
      hsl = rgbToHsl(c[0], c[1], c[2]);
      var head = Math.round(hsl[0]) + ', ' + Math.round(hsl[1]) + '%, ' + Math.round(hsl[2]) + '%';
      return f === 'hsl' ? 'hsl(' + head + ')' : 'hsla(' + head + ', ' + Number(c[3].toFixed(3)) + ')';
    }
    throw new Error('unsupported color format');
  }
  function luminance(value) {
    var c = parseColor(value);
    function lin(v) { v /= 255; return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4); }
    return 0.2126 * lin(c[0]) + 0.7152 * lin(c[1]) + 0.0722 * lin(c[2]);
  }

  var color = {
    parse: function (value) { try { return ok(colorObject(parseColor(value))); } catch (e) { return fail('color.parse: ' + msg(e)); } },
    format: function (value, format) { try { return ok(formatColor(value, format)); } catch (e) { return failEmpty('color.format: ' + msg(e)); } },
    rgb: function (r, g, b, a) {
      try { return ok(colorObject([channel(r), channel(g), channel(b), alphaOf(a == null ? 1 : a)])); }
      catch (e) { return fail('color.rgb: ' + msg(e)); }
    },
    hsl: function (h, s, l, a) {
      try { var c = hslToRgb(h, s, l); c.push(alphaOf(a == null ? 1 : a)); return ok(colorObject(c)); }
      catch (e) { return fail('color.hsl: ' + msg(e)); }
    },
    mix: function (a, b, t) {
      try {
        var x = parseColor(a), y = parseColor(b), p = clamp(number(t, 't'), 0, 1);
        return ok(colorObject([
          Math.round(x[0] + (y[0] - x[0]) * p), Math.round(x[1] + (y[1] - x[1]) * p),
          Math.round(x[2] + (y[2] - x[2]) * p), x[3] + (y[3] - x[3]) * p
        ]));
      } catch (e) { return fail('color.mix: ' + msg(e)); }
    },
    withAlpha: function (value, alpha) {
      try { var c = parseColor(value); c[3] = alphaOf(alpha); return ok(colorObject(c)); }
      catch (e) { return fail('color.withAlpha: ' + msg(e)); }
    },
    luminance: function (value) { try { return ok(Number(luminance(value).toFixed(6))); } catch (e) { return failZero('color.luminance: ' + msg(e)); } },
    contrast: function (a, b) {
      try {
        var x = luminance(a), y = luminance(b);
        return ok(Number(((Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05)).toFixed(6)));
      } catch (e) { return failZero('color.contrast: ' + msg(e)); }
    },
    isDark: function (value) { try { return ok(luminance(value) < 0.5); } catch (e) { return failFalse('color.isDark: ' + msg(e)); } }
  };

  var unitFactors = { px: 1, mm: 96 / 25.4, cm: 96 / 2.54, in: 96, pt: 96 / 72, pc: 16, q: 96 / 101.6 };
  function factorOf(unit) {
    var u = String(unit == null ? '' : unit).trim().toLowerCase();
    var f = unitFactors[u];
    if (!f) throw new Error('unsupported unit ' + (u || '(empty)'));
    return f;
  }
  var units = {
    parse: function (value) {
      try {
        var m = /^\s*(-?(?:\d+\.?\d*|\.\d+))\s*([a-z%]*)\s*$/i.exec(String(value == null ? '' : value));
        if (!m) throw new Error('invalid length');
        var out = {};
        setKey(out, 'value', parseFloat(m[1]));
        setKey(out, 'unit', m[2].toLowerCase());
        return ok(out);
      } catch (e) { return fail('units.parse: ' + msg(e)); }
    },
    convert: function (value, from, to) {
      try { return ok(number(value, 'value') * factorOf(from) / factorOf(to)); }
      catch (e) { return failZero('units.convert: ' + msg(e)); }
    },
    toBase: function (value, unit) { try { return ok(number(value, 'value') * factorOf(unit)); } catch (e) { return failZero('units.toBase: ' + msg(e)); } },
    fromBase: function (value, unit) { try { return ok(number(value, 'value') / factorOf(unit)); } catch (e) { return failZero('units.fromBase: ' + msg(e)); } },
    format: function (value, unit, precision) {
      try {
        var p = clamp(integer(precision == null ? 2 : precision, 'precision'), 0, 6);
        var s = number(value, 'value').toFixed(p);
        if (s.indexOf('.') >= 0) s = s.replace(/0+$/, '').replace(/\.$/, '');
        if (s === '-0') s = '0';
        return ok(s + String(unit == null ? '' : unit));
      } catch (e) { return failEmpty('units.format: ' + msg(e)); }
    }
  };

  var paperSizes = {
    a4: [210, 297], a3: [297, 420], a5: [148, 210],
    letter: [216, 279.4], legal: [216, 356], tabloid: [279.4, 432]
  };
  var print = {
    pxToMm: function (px) { try { return ok(number(px, 'px') / unitFactors.mm); } catch (e) { return failZero('print.pxToMm: ' + msg(e)); } },
    mmToPx: function (mm) { try { return ok(number(mm, 'mm') * unitFactors.mm); } catch (e) { return failZero('print.mmToPx: ' + msg(e)); } },
    ptToMm: function (pt) { try { return ok(number(pt, 'pt') * unitFactors.pt / unitFactors.mm); } catch (e) { return failZero('print.ptToMm: ' + msg(e)); } },
    paper: function (name, orientation) {
      try {
        var key = String(name == null ? '' : name).toLowerCase().replace(/\s+/g, '');
        var size = paperSizes[key];
        if (!size) throw new Error('unknown paper size');
        var o = String(orientation || 'portrait').toLowerCase();
        if (o !== 'portrait' && o !== 'landscape') throw new Error('orientation must be portrait or landscape');
        var w = size[0], h = size[1];
        if (o === 'landscape') { var t = w; w = h; h = t; }
        var out = {};
        setKey(out, 'width', w); setKey(out, 'height', h); setKey(out, 'unit', 'mm'); setKey(out, 'orientation', o);
        return ok(out);
      } catch (e) { return fail('print.paper: ' + msg(e)); }
    },
    fitBox: function (width, height, maxW, maxH) {
      try {
        var w = number(width, 'width'), h = number(height, 'height'), mw = number(maxW, 'maxW'), mh = number(maxH, 'maxH');
        if (w <= 0 || h <= 0 || mw <= 0 || mh <= 0) throw new Error('dimensions must be positive');
        var scale = Math.min(mw / w, mh / h, 1), out = {};
        setKey(out, 'width', Number((w * scale).toFixed(3)));
        setKey(out, 'height', Number((h * scale).toFixed(3)));
        setKey(out, 'scale', Number(scale.toFixed(6)));
        return ok(out);
      } catch (e) { return fail('print.fitBox: ' + msg(e)); }
    }
  };

  var histories = {};
  function historyFor(name, create) {
    var n = String(name == null ? '' : name);
    if (!/^[A-Za-z0-9_.:-]{1,128}$/.test(n) || n === '__proto__') throw new Error('invalid history name');
    var h = histories[n];
    if (!h && create) { h = { items: [], index: -1, limit: 50 }; histories[n] = h; }
    if (!h) throw new Error('history not found: ' + n);
    return h;
  }
  var history = {
    push: function (name, value, limit) {
      try {
        var h = historyFor(name, true), lim = clamp(integer(limit == null ? 50 : limit, 'limit'), 1, 1000);
        h.limit = lim;
        h.items = h.items.slice(0, h.index + 1);
        h.items.push(cloneValue(value, 0, []));
        if (h.items.length > lim) h.items.shift();
        h.index = h.items.length - 1;
        return ok(h.items.length);
      } catch (e) { return failZero('history.push: ' + msg(e)); }
    },
    undo: function (name) {
      try { var h = historyFor(name, false); if (h.index <= 0) throw new Error('nothing to undo'); h.index--; return ok(cloneValue(h.items[h.index], 0, [])); }
      catch (e) { return fail('history.undo: ' + msg(e)); }
    },
    redo: function (name) {
      try { var h = historyFor(name, false); if (h.index >= h.items.length - 1) throw new Error('nothing to redo'); h.index++; return ok(cloneValue(h.items[h.index], 0, [])); }
      catch (e) { return fail('history.redo: ' + msg(e)); }
    },
    current: function (name) {
      try { var h = historyFor(name, false); if (h.index < 0) throw new Error('history is empty'); return ok(cloneValue(h.items[h.index], 0, [])); }
      catch (e) { return fail('history.current: ' + msg(e)); }
    },
    canUndo: function (name) { try { var h = historyFor(name, false); return ok(h.index > 0); } catch (e) { return failFalse('history.canUndo: ' + msg(e)); } },
    canRedo: function (name) { try { var h = historyFor(name, false); return ok(h.index < h.items.length - 1); } catch (e) { return failFalse('history.canRedo: ' + msg(e)); } },
    clear: function (name) { try { var h = historyFor(name, true); h.items = []; h.index = -1; return okNull(); } catch (e) { return fail('history.clear: ' + msg(e)); } },
    list: function () { try { return ok(Object.keys(histories).sort()); } catch (e) { return fail('history.list: ' + msg(e)); } },
    size: function (name) { try { return ok(historyFor(name, false).items.length); } catch (e) { return failZero('history.size: ' + msg(e)); } },
    jump: function (name, index) {
      try {
        var h = historyFor(name, false), i = integer(index, 'index');
        if (i < 0 || i >= h.items.length) throw new Error('history index out of bounds');
        h.index = i;
        return ok(cloneValue(h.items[i], 0, []));
      } catch (e) { return fail('history.jump: ' + msg(e)); }
    }
  };

  var crcTable = (function () {
    var table = [], n, k, c;
    for (n = 0; n < 256; n++) {
      c = n;
      for (k = 0; k < 8; k++) c = (c & 1) ? (0xEDB88320 ^ (c >>> 1)) : (c >>> 1);
      table[n] = c >>> 0;
    }
    return table;
  })();
  function crc32(u8) {
    var c = 0xFFFFFFFF;
    for (var i = 0; i < u8.length; i++) c = crcTable[(c ^ u8[i]) & 255] ^ (c >>> 8);
    return (c ^ 0xFFFFFFFF) >>> 0;
  }
  function validateZipName(name, bytes) {
    var n = String(name == null ? '' : name);
    if (!n) throw new Error('entry name is empty');
    if (!bytes || bytes.length > MAX_NAME_BYTES) throw new Error('entry name is too long');
    if (n.indexOf('\\') >= 0 || n.charAt(0) === '/') throw new Error('entry name contains an absolute or Windows path');
    if (/^[A-Za-z]:/.test(n)) throw new Error('entry name contains a drive letter');
    if (/\u0000/.test(n)) throw new Error('entry name contains NUL');
    if (n.charAt(n.length - 1) === '/') throw new Error('directory entries are not supported');
    var parts = n.split('/');
    for (var i = 0; i < parts.length; i++) {
      if (!parts[i] || parts[i] === '.' || parts[i] === '..') throw new Error('entry name contains an unsafe path segment');
    }
    return n;
  }
  function zipWriter() { return { parts: [], entries: [], size: 0, closed: false, marker: true }; }
  function addZipEntry(w, name, data) {
    if (!w || !w.marker || w.closed) throw new Error('archive is closed or invalid');
    if (w.entries.length >= MAX_ENTRIES) throw new Error('too many archive entries');
    if (data.length > MAX_BYTES) throw new Error('archive entry is too large');
    if (w.size + data.length > MAX_ZIP_TOTAL) throw new Error('archive is too large');
    var nameBytes = utf8Encode(name);
    validateZipName(name, nameBytes);
    var offset = w.size, crc = crc32(data);
    var header = concatBytes([
      packU32(0x04034b50), packU16(20), packU16(0x0800), packU16(0), packU16(0), packU16(0),
      packU32(crc), packU32(data.length), packU32(data.length), packU16(nameBytes.length), packU16(0), nameBytes
    ]);
    w.parts.push(header, data);
    w.size += header.length + data.length;
    w.entries.push({ name: name, nameBytes: nameBytes, data: data, crc: crc, size: data.length, offset: offset });
    return w.entries.length;
  }
  function finishZip(w) {
    if (!w || !w.marker || w.closed) throw new Error('archive is closed or invalid');
    var central = [], centralSize = 0, i, h;
    for (i = 0; i < w.entries.length; i++) {
      var e = w.entries[i];
      h = concatBytes([
        packU32(0x02014b50), packU16(20), packU16(20), packU16(0x0800), packU16(0), packU16(0), packU16(0),
        packU32(e.crc), packU32(e.size), packU32(e.size), packU16(e.nameBytes.length), packU16(0), packU16(0),
        packU16(0), packU16(0), packU32(0), packU32(e.offset), e.nameBytes
      ]);
      central.push(h);
      centralSize += h.length;
    }
    var cdOffset = w.size;
    var eocd = concatBytes([
      packU32(0x06054b50), packU16(0), packU16(0), packU16(w.entries.length), packU16(w.entries.length),
      packU32(centralSize), packU32(cdOffset), packU16(0)
    ]);
    w.closed = true;
    return concatBytes(w.parts.concat(central, [eocd]));
  }
  function openZip(value) {
    var u8 = bytesOf(value, 'archive bytes');
    if (u8.length < 22) throw new Error('archive is too short');
    var view = new DataView(u8.buffer, u8.byteOffset, u8.byteLength);
    var min = Math.max(0, u8.length - 65557), i, eocd = -1;
    for (i = u8.length - 22; i >= min; i--) {
      if (view.getUint32(i, true) === 0x06054b50) { eocd = i; break; }
    }
    if (eocd < 0) throw new Error('end of central directory not found');
    var count = view.getUint16(eocd + 10, true), cdSize = view.getUint32(eocd + 12, true), cdOffset = view.getUint32(eocd + 16, true);
    if (count > MAX_ENTRIES) throw new Error('too many archive entries');
    if (cdOffset + cdSize > u8.length) throw new Error('central directory is out of bounds');
    var entries = [], byName = {}, pos = cdOffset, n;
    for (n = 0; n < count; n++) {
      if (pos + 46 > u8.length || view.getUint32(pos, true) !== 0x02014b50) throw new Error('invalid central directory entry');
      var flags = view.getUint16(pos + 8, true), method = view.getUint16(pos + 10, true);
      var crc = view.getUint32(pos + 16, true), size = view.getUint32(pos + 24, true);
      var nameLen = view.getUint16(pos + 28, true), extraLen = view.getUint16(pos + 30, true), commentLen = view.getUint16(pos + 32, true);
      var attrs = view.getUint32(pos + 38, true), localOffset = view.getUint32(pos + 42, true);
      if ((flags & 1) !== 0) throw new Error('encrypted archive entries are not supported');
      if (method !== 0) throw new Error('only stored (uncompressed) archive entries are supported');
      if (((attrs >>> 16) & 0xF000) === 0xA000) throw new Error('symbolic links are not allowed');
      if (size > MAX_BYTES || view.getUint32(pos + 20, true) !== size) throw new Error('archive entry is too large or inconsistent');
      if (pos + 46 + nameLen + extraLen + commentLen > u8.length) throw new Error('central directory entry is truncated');
      var nameBytes = u8.subarray(pos + 46, pos + 46 + nameLen);
      var name = validateZipName(utf8Decode(nameBytes), nameBytes);
      if (localOffset + 30 > u8.length || view.getUint32(localOffset, true) !== 0x04034b50) throw new Error('invalid local file header');
      var localNameLen = view.getUint16(localOffset + 26, true), localExtraLen = view.getUint16(localOffset + 28, true);
      var localName = u8.subarray(localOffset + 30, localOffset + 30 + localNameLen);
      if (localNameLen !== nameBytes.length || byteCompare(localName, nameBytes) !== 0) throw new Error('local and central names differ');
      var dataOffset = localOffset + 30 + localNameLen + localExtraLen;
      if (dataOffset + size > u8.length || dataOffset + size > cdOffset) throw new Error('archive entry is out of bounds');
      var entry = { name: name, crc: crc, size: size, offset: dataOffset, method: 0 };
      entries.push(entry);
      setKey(byName, name, entry);
      pos += 46 + nameLen + extraLen + commentLen;
    }
    return { marker: true, entries: entries, byName: byName, bytes: u8 };
  }
  function zipRead(archive, name) {
    if (!archive || !archive.marker) throw new Error('archive is invalid');
    var key = String(name == null ? '' : name), entry = archive.byName[key];
    if (!entry) throw new Error('archive entry not found: ' + key);
    var data = archive.bytes.subarray(entry.offset, entry.offset + entry.size);
    if (crc32(data) !== entry.crc) throw new Error('archive entry failed its CRC check');
    return data;
  }
  function zipInfo(entries) {
    return entries.map(function (e) {
      var out = {};
      setKey(out, 'name', e.name); setKey(out, 'size', e.size); setKey(out, 'crc', e.crc); setKey(out, 'method', e.method == null ? 0 : e.method);
      return out;
    });
  }
  function requireZipFormat(format) {
    var f = format == null || format === '' ? 'zip' : String(format).toLowerCase();
    if (f !== 'zip') throw new Error('unsupported archive format: ' + f);
  }

  var zip = {
    create: function () { try { return ok(zipWriter()); } catch (e) { return fail('zip.create: ' + msg(e)); } },
    add: function (archive, name, value) {
      try { return ok(addZipEntry(archive, String(name == null ? '' : name), bytesOf(value, 'entry bytes'))); }
      catch (e) { return failZero('zip.add: ' + msg(e)); }
    },
    addText: function (archive, name, text) {
      try { return ok(addZipEntry(archive, String(name == null ? '' : name), utf8Encode(text))); }
      catch (e) { return failZero('zip.addText: ' + msg(e)); }
    },
    finish: function (archive) { try { return ok(outBytes(finishZip(archive))); } catch (e) { return fail('zip.finish: ' + msg(e)); } },
    open: function (value) { try { return ok(openZip(value)); } catch (e) { return fail('zip.open: ' + msg(e)); } },
    list: function (archive) {
      try { if (!archive || !archive.marker) throw new Error('archive is invalid'); return ok(zipInfo(archive.entries)); }
      catch (e) { return fail('zip.list: ' + msg(e)); }
    },
    read: function (archive, name) { try { return ok(outBytes(zipRead(archive, name))); } catch (e) { return fail('zip.read: ' + msg(e)); } },
    readText: function (archive, name) { try { return ok(utf8Decode(zipRead(archive, name))); } catch (e) { return failEmpty('zip.readText: ' + msg(e)); } },
    has: function (archive, name) {
      try { if (!archive || !archive.marker) throw new Error('archive is invalid'); return ok(Object.prototype.hasOwnProperty.call(archive.byName, String(name))); }
      catch (e) { return failFalse('zip.has: ' + msg(e)); }
    },
    count: function (archive) {
      try { if (!archive || !archive.marker) throw new Error('archive is invalid'); return ok(archive.entries.length); }
      catch (e) { return failZero('zip.count: ' + msg(e)); }
    }
  };

  function packFiles(files, format, textMode) {
    requireZipFormat(format);
    if (!isObj(files)) throw new Error('files must be a map');
    var w = zipWriter(), keys = Object.keys(files).sort(), i;
    for (i = 0; i < keys.length; i++) {
      var value = files[keys[i]];
      addZipEntry(w, keys[i], textMode ? utf8Encode(value) : bytesOf(value, 'file ' + keys[i]));
    }
    return finishZip(w);
  }
  function unpackFiles(value, format, textMode) {
    requireZipFormat(format);
    var a = openZip(value), out = {};
    for (var i = 0; i < a.entries.length; i++) {
      var e = a.entries[i], data = zipRead(a, e.name);
      setKey(out, e.name, textMode ? utf8Decode(data) : outBytes(data));
    }
    return out;
  }
  var archive = {
    pack: function (files, format) { try { return ok(packFiles(files, format, false)); } catch (e) { return fail('archive.pack: ' + msg(e)); } },
    unpack: function (value, format) { try { return ok(unpackFiles(value, format, false)); } catch (e) { return fail('archive.unpack: ' + msg(e)); } },
    packText: function (files, format) { try { return ok(packFiles(files, format, true)); } catch (e) { return fail('archive.packText: ' + msg(e)); } },
    unpackText: function (value, format) { try { return ok(unpackFiles(value, format, true)); } catch (e) { return fail('archive.unpackText: ' + msg(e)); } },
    info: function (value) { try { return ok(zipInfo(openZip(value).entries)); } catch (e) { return fail('archive.info: ' + msg(e)); } }
  };

  function csvOptions(value) {
    var o = opt(value), delimiter = o.delimiter == null ? ',' : String(o.delimiter);
    if (delimiter.length !== 1) throw new Error('delimiter must be one character');
    return {
      delimiter: delimiter,
      quote: o.quote == null ? '"' : String(o.quote),
      newline: o.newline == null ? '\n' : String(o.newline),
      columns: Array.isArray(o.columns) ? o.columns.map(String) : null,
      header: optBool(o, 'header', true)
    };
  }
  function parseCsv(text, options) {
    var o = csvOptions(options), s = String(text == null ? '' : text);
    var rows = [], row = [], field = '', inQuotes = false, i = 0, ch, next;
    if (s.charCodeAt(0) === 0xFEFF) s = s.slice(1);
    while (i < s.length) {
      ch = s.charAt(i); next = s.charAt(i + 1);
      if (inQuotes) {
        if (ch === o.quote) {
          if (next === o.quote) { field += o.quote; i++; }
          else inQuotes = false;
        } else field += ch;
      } else if (ch === o.quote && field === '') {
        inQuotes = true;
      } else if (ch === o.delimiter) {
        row.push(field); field = '';
      } else if (ch === '\r' || ch === '\n') {
        if (ch === '\r' && next === '\n') i++;
        row.push(field); rows.push(row); row = []; field = '';
      } else field += ch;
      i++;
    }
    if (inQuotes) throw new Error('unterminated quoted CSV field');
    if (field !== '' || row.length) { row.push(field); rows.push(row); }
    if (rows.length > 10000) throw new Error('too many CSV rows');
    return rows.filter(function (r) { return !(r.length === 1 && r[0] === ''); });
  }
  function formulaSafe(value) {
    var s = value == null ? '' : String(value);
    return /^[=+\-@\t\r]/.test(s) ? "'" + s : s;
  }
  function csvCell(value, o) {
    var s = formulaSafe(value);
    if (s.indexOf(o.delimiter) >= 0 || s.indexOf(o.quote) >= 0 || s.indexOf('\n') >= 0 || s.indexOf('\r') >= 0) {
      return o.quote + s.split(o.quote).join(o.quote + o.quote) + o.quote;
    }
    return s;
  }
  function csvRow(row, o) {
    if (!Array.isArray(row)) throw new Error('each CSV row must be an array');
    if (row.length > 1000) throw new Error('too many CSV columns');
    return row.map(function (v) { return csvCell(v, o); }).join(o.delimiter);
  }
  var csv = {
    parse: function (text, options) { try { return ok(parseCsv(text, options)); } catch (e) { return fail('csv.parse: ' + msg(e)); } },
    parseObjects: function (text, options) {
      try {
        var rows = parseCsv(text, options);
        if (!rows.length) return ok([]);
        var headers = rows[0], seen = {}, i, j;
        for (i = 0; i < headers.length; i++) {
          if (!headers[i] || seen[headers[i]]) throw new Error('duplicate or empty CSV header');
          seen[headers[i]] = true;
        }
        var out = [];
        for (i = 1; i < rows.length; i++) {
          if (rows[i].length !== headers.length) throw new Error('CSV row has a different column count');
          var obj = {};
          for (j = 0; j < headers.length; j++) setKey(obj, headers[j], rows[i][j]);
          out.push(obj);
        }
        return ok(out);
      } catch (e) { return fail('csv.parseObjects: ' + msg(e)); }
    },
    stringify: function (rows, options) {
      try {
        var o = csvOptions(options);
        if (!Array.isArray(rows)) throw new Error('rows must be an array');
        if (rows.length > 10000) throw new Error('too many CSV rows');
        var lines = [];
        if (o.columns && o.header) lines.push(csvRow(o.columns, o));
        for (var i = 0; i < rows.length; i++) lines.push(csvRow(rows[i], o));
        return ok(lines.join(o.newline));
      } catch (e) { return failEmpty('csv.stringify: ' + msg(e)); }
    },
    stringifyObjects: function (rows, options) {
      try {
        var o = csvOptions(options);
        if (!Array.isArray(rows)) throw new Error('rows must be an array');
        var columns = o.columns || (rows.length ? Object.keys(rows[0]) : []);
        var lines = [];
        if (o.header) lines.push(csvRow(columns, o));
        for (var i = 0; i < rows.length; i++) {
          if (!isObj(rows[i])) throw new Error('each row must be a map');
          var values = columns.map(function (c) { return rows[i][c] == null ? '' : rows[i][c]; });
          lines.push(csvRow(values, o));
        }
        return ok(lines.join(o.newline));
      } catch (e) { return failEmpty('csv.stringifyObjects: ' + msg(e)); }
    },
    validate: function (text, options) {
      try {
        var rows = parseCsv(text, options);
        if (!rows.length) return ok(true);
        var width = rows[0].length;
        for (var i = 1; i < rows.length; i++) if (rows[i].length !== width) throw new Error('CSV row has a different column count');
        return ok(true);
      } catch (e) { return failFalse('csv.validate: ' + msg(e)); }
    }
  };

  function envValues() {
    var nav = typeof navigator !== 'undefined' ? navigator : {};
    var ud = nav.userAgentData || {};
    var platform = ud.platform || nav.platform || 'unknown';
    var language = nav.language || 'en';
    var timezone = 'UTC';
    try { if (typeof Intl !== 'undefined') timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'; } catch (e) {}
    var out = {};
    setKey(out, 'SUI_TARGET', TARGET);
    setKey(out, 'SUI_PLATFORM', String(platform));
    setKey(out, 'SUI_LANGUAGE', String(language));
    setKey(out, 'SUI_TIMEZONE', String(timezone));
    setKey(out, 'SUI_USER_AGENT', String(nav.userAgent || ''));
    setKey(out, 'SUI_DESKTOP', TARGET === 'desktop' || TARGET === 'desktop-webview' ? 'true' : 'false');
    return out;
  }
  function capabilityMap() {
    var nav = typeof navigator !== 'undefined' ? navigator : {};
    var desktop = TARGET === 'desktop' || TARGET === 'desktop-webview';
    var out = {};
    setKey(out, 'storage', typeof localStorage !== 'undefined');
    setKey(out, 'clipboard', !!(nav.clipboard && nav.clipboard.writeText));
    setKey(out, 'crypto', typeof crypto !== 'undefined' && !!crypto.subtle);
    setKey(out, 'fs', desktop);
    setKey(out, 'exec', desktop);
    setKey(out, 'dialog', desktop);
    setKey(out, 'notifications', typeof Notification !== 'undefined');
    setKey(out, 'intl', typeof Intl !== 'undefined');
    setKey(out, 'typedArrays', typeof Uint8Array !== 'undefined');
    setKey(out, 'textEncoder', typeof TextEncoder !== 'undefined');
    return out;
  }
  var env = {
    target: function () { try { return ok(TARGET); } catch (e) { return failEmpty('env.target: ' + msg(e)); } },
    platform: function () { try { return ok(envValues().SUI_PLATFORM); } catch (e) { return failEmpty('env.platform: ' + msg(e)); } },
    get: function (name) {
      try {
        var key = String(name == null ? '' : name).toUpperCase(), values = envValues();
        if (!Object.prototype.hasOwnProperty.call(values, key)) throw new Error('environment key is not exposed');
        return ok(values[key]);
      } catch (e) { return failEmpty('env.get: ' + msg(e)); }
    },
    has: function (name) {
      try { return ok(Object.prototype.hasOwnProperty.call(envValues(), String(name == null ? '' : name).toUpperCase())); }
      catch (e) { return failFalse('env.has: ' + msg(e)); }
    },
    all: function () { try { return ok(envValues()); } catch (e) { return fail('env.all: ' + msg(e)); } },
    language: function () { try { return ok(envValues().SUI_LANGUAGE); } catch (e) { return failEmpty('env.language: ' + msg(e)); } },
    timezone: function () { try { return ok(envValues().SUI_TIMEZONE); } catch (e) { return failEmpty('env.timezone: ' + msg(e)); } }
  };
  var capabilities = {
    has: function (name) {
      try { return ok(Object.prototype.hasOwnProperty.call(capabilityMap(), String(name == null ? '' : name).toLowerCase())); }
      catch (e) { return failFalse('capabilities.has: ' + msg(e)); }
    },
    check: function (name) {
      try {
        var key = String(name == null ? '' : name).toLowerCase(), caps = capabilityMap();
        if (!caps[key]) throw new Error('capability is unavailable: ' + key);
        return ok(true);
      } catch (e) { return failFalse('capabilities.check: ' + msg(e)); }
    },
    list: function () { try { return ok(Object.keys(capabilityMap()).sort()); } catch (e) { return fail('capabilities.list: ' + msg(e)); } },
    target: function () { try { return ok(TARGET); } catch (e) { return failEmpty('capabilities.target: ' + msg(e)); } }
  };

  var eventBus = {};
  function eventFor(name, create) {
    var n = String(name == null ? '' : name);
    if (!n || n.length > 128) throw new Error('invalid event name');
    var b = eventBus[n];
    if (!b && create) { b = { listeners: [] }; eventBus[n] = b; }
    if (!b) throw new Error('event not found: ' + n);
    return b;
  }
  function eventAdd(name, fn, once) {
    var b = eventFor(name, true);
    if (typeof fn !== 'function') throw new Error('listener must be a function');
    var entry = { fn: fn, once: !!once };
    b.listeners.push(entry);
    return b.listeners.length;
  }
  function eventRemove(name, fn) {
    var b = eventFor(name, false);
    if (typeof fn !== 'function') throw new Error('listener must be a function');
    var before = b.listeners.length;
    b.listeners = b.listeners.filter(function (e) { return e.fn !== fn && e.source !== fn; });
    return before - b.listeners.length;
  }
  function eventRemoveEntry(name, entry) {
    var b = eventFor(name, false), i = b.listeners.indexOf(entry);
    if (i < 0) return 0;
    b.listeners.splice(i, 1);
    return 1;
  }
  function eventEmit(name, payload) {
    var b = eventFor(name, true), listeners = b.listeners.slice(), data = safeClone(payload), count = 0, err = null;
    for (var i = 0; i < listeners.length; i++) {
      var entry = listeners[i];
      if (entry.once) eventRemoveEntry(name, entry);
      count++;
      try { entry.fn(data); }
      catch (e) { if (!err) err = msg(e); }
    }
    return [count, err];
  }
  var events = {
    on: function (name, fn) { try { return ok(eventAdd(name, fn, false)); } catch (e) { return failZero('events.on: ' + msg(e)); } },
    once: function (name, fn) { try { return ok(eventAdd(name, fn, true)); } catch (e) { return failZero('events.once: ' + msg(e)); } },
    off: function (name, fn) { try { return ok(eventRemove(name, fn)); } catch (e) { return failZero('events.off: ' + msg(e)); } },
    emit: function (name, payload) { try { var r = eventEmit(name, payload); return [r[0], r[1]]; } catch (e) { return failZero('events.emit: ' + msg(e)); } },
    clear: function (name) { try { var b = eventFor(name, true); b.listeners = []; return okNull(); } catch (e) { return fail('events.clear: ' + msg(e)); } },
    count: function (name) { try { return ok(eventFor(name, false).listeners.length); } catch (e) { return failZero('events.count: ' + msg(e)); } },
    names: function () { try { return ok(Object.keys(eventBus).sort()); } catch (e) { return fail('events.names: ' + msg(e)); } }
  };
  var bus = {
    on: events.on, once: events.once, off: events.off, emit: events.emit,
    clear: events.clear, count: events.count, names: events.names
  };

  var logState = { level: 'info', entries: [] };
  var logLevels = { debug: 10, info: 20, warn: 30, error: 40 };
  var secretKey = /^(password|passwd|pwd|token|secret|api[_-]?key|authorization|cookie|session|creditcard|cardnumber|cvv|pin)$/i;
  function redactText(text) {
    var s = String(text == null ? '' : text);
    s = s.replace(/(Bearer\s+)[A-Za-z0-9._~+\/=-]+/gi, '$1[REDACTED]');
    s = s.replace(/((?:password|passwd|pwd|token|secret|api[_-]?key|authorization|cookie|session)\s*[=:]\s*)([^\s&,;]+)/gi, '$1[REDACTED]');
    return s.slice(0, 2000);
  }
  function redactValue(value, depth, seen) {
    if (value === null || typeof value !== 'object') return typeof value === 'string' ? redactText(value) : value;
    depth = depth || 0;
    if (depth > 32) return '[TRUNCATED]';
    seen = seen || [];
    if (seen.indexOf(value) >= 0) return '[CYCLE]';
    seen.push(value);
    var out, keys, i;
    if (Array.isArray(value)) {
      if (value.length > 1000) return '[TRUNCATED]';
      out = [];
      for (i = 0; i < value.length; i++) out.push(redactValue(value[i], depth + 1, seen));
    } else {
      keys = Object.keys(value);
      if (keys.length > 500) return '[TRUNCATED]';
      out = {};
      for (i = 0; i < keys.length; i++) {
        if (secretKey.test(keys[i])) setKey(out, keys[i], '[REDACTED]');
        else setKey(out, keys[i], redactValue(value[keys[i]], depth + 1, seen));
      }
    }
    seen.pop();
    return out;
  }
  function addLog(level, message, context) {
    if (logLevels[level] < logLevels[logState.level]) return okNull();
    var entry = {};
    setKey(entry, 'level', level);
    setKey(entry, 'time', new Date().toISOString());
    setKey(entry, 'message', redactText(message));
    setKey(entry, 'context', context == null ? null : redactValue(context, 0, []));
    logState.entries.push(entry);
    if (logState.entries.length > 500) logState.entries.shift();
    return okNull();
  }
  var log = {
    debug: function (message, context) { try { return addLog('debug', message, context); } catch (e) { return fail('log.debug: ' + msg(e)); } },
    info: function (message, context) { try { return addLog('info', message, context); } catch (e) { return fail('log.info: ' + msg(e)); } },
    warn: function (message, context) { try { return addLog('warn', message, context); } catch (e) { return fail('log.warn: ' + msg(e)); } },
    error: function (message, context) { try { return addLog('error', message, context); } catch (e) { return fail('log.error: ' + msg(e)); } },
    setLevel: function (level) {
      try {
        var l = String(level || '').toLowerCase();
        if (!logLevels[l]) throw new Error('unknown log level');
        logState.level = l;
        return okNull();
      } catch (e) { return fail('log.setLevel: ' + msg(e)); }
    },
    level: function () { try { return ok(logState.level); } catch (e) { return failEmpty('log.level: ' + msg(e)); } },
    clear: function () { try { logState.entries = []; return okNull(); } catch (e) { return fail('log.clear: ' + msg(e)); } },
    entries: function (limit) {
      try {
        var n = clamp(integer(limit == null ? 100 : limit, 'limit'), 0, 500);
        return ok(cloneValue(logState.entries.slice(-n), 0, []));
      } catch (e) { return fail('log.entries: ' + msg(e)); }
    },
    export: function () { try { return ok(JSON.stringify(logState.entries, null, 2)); } catch (e) { return failEmpty('log.export: ' + msg(e)); } },
    redact: function (text) { try { return ok(redactText(text)); } catch (e) { return failEmpty('log.redact: ' + msg(e)); } },
    redactObject: function (value) { try { return ok(redactValue(value, 0, [])); } catch (e) { return fail('log.redactObject: ' + msg(e)); } }
  };

  var telemetryState = {};
  function metricName(name) {
    var n = String(name == null ? '' : name);
    if (!/^[A-Za-z0-9_.:-]{1,128}$/.test(n) || n === '__proto__') throw new Error('invalid metric name');
    return n;
  }
  function metricFor(name, type) {
    var n = metricName(name), m = telemetryState[n];
    if (!m || m.type !== type) { m = { type: type, value: 0, updated: new Date().toISOString() }; telemetryState[n] = m; }
    return m;
  }
  function metricObject(name, m) {
    var out = {};
    setKey(out, 'name', name);
    setKey(out, 'type', m.type);
    setKey(out, 'value', Number(m.value.toFixed(6)));
    setKey(out, 'updated', m.updated);
    return out;
  }
  var telemetry = {
    counter: function (name, delta) {
      try {
        var m = metricFor(name, 'counter'), d = delta == null ? 1 : number(delta, 'delta');
        m.value += d; m.updated = new Date().toISOString();
        return ok(m.value);
      } catch (e) { return failZero('telemetry.counter: ' + msg(e)); }
    },
    gauge: function (name, value) {
      try {
        var m = metricFor(name, 'gauge');
        m.value = number(value, 'value'); m.updated = new Date().toISOString();
        return ok(m.value);
      } catch (e) { return failZero('telemetry.gauge: ' + msg(e)); }
    },
    timing: function (name, ms) {
      try {
        var m = metricFor(name, 'timing');
        m.value = number(ms, 'ms'); m.updated = new Date().toISOString();
        return ok(m.value);
      } catch (e) { return failZero('telemetry.timing: ' + msg(e)); }
    },
    snapshot: function () {
      try {
        var out = {}, keys = Object.keys(telemetryState).sort();
        for (var i = 0; i < keys.length; i++) setKey(out, keys[i], metricObject(keys[i], telemetryState[keys[i]]));
        return ok(out);
      } catch (e) { return fail('telemetry.snapshot: ' + msg(e)); }
    },
    reset: function () { try { telemetryState = {}; return okNull(); } catch (e) { return fail('telemetry.reset: ' + msg(e)); } },
    names: function () { try { return ok(Object.keys(telemetryState).sort()); } catch (e) { return fail('telemetry.names: ' + msg(e)); } }
  };

  return {
    bytes: bytes, bin: bin, files: files, mime: mime, color: color, units: units,
    print: print, history: history, zip: zip, archive: archive, csv: csv,
    env: env, capabilities: capabilities, events: events, bus: bus, log: log,
    telemetry: telemetry
  };
})();`

// stdlibExtraRuntimeSource returns the shared prelude with the compile target
// embedded as a literal. The target is compiler-controlled, never user input.
func stdlibExtraRuntimeSource(target string) string {
	safe := strings.ReplaceAll(target, `\`, `\\`)
	safe = strings.ReplaceAll(safe, `"`, `\"`)
	safe = strings.ReplaceAll(safe, "\n", "")
	safe = strings.ReplaceAll(safe, "\r", "")
	return strings.ReplaceAll(stdlibExtraRuntimeRaw, "__SUI_TARGET__", safe)
}
