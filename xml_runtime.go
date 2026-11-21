package sui

// xml_runtime.go - the browser/desktop implementation of the `xml` std module.
//
// The generated app always uses the browser XML DOM (DOMParser, XMLSerializer,
// XPathResult, XSLTProcessor). This keeps namespaces, comments, CDATA and
// processing instructions intact and gives access to XPath without shipping a
// second parser. The runtime is emitted only when `xml` is imported.
//
// Security: documents are parsed after rejecting DOCTYPE/ENTITY declarations,
// with a character and node-count limit. Data is never inserted into HTML; the
// serializer only emits XML text. See SECURITY.md.
const xmlRuntimeSource = `var __SUI_XML = (function () {
  function err(message) { return [null, String(message == null ? 'XML error' : message).slice(0, 500)]; }
  function ok(value) { return [value, null]; }
  function nodeOf(value) {
    return value && typeof value === 'object' && typeof value.nodeType === 'number' ? value : null;
  }
  function ownerOf(node) {
    if (!node) return null;
    if (node.nodeType === 9) return node;
    return node.ownerDocument || null;
  }
  function elementOf(value) {
    var n = nodeOf(value);
    if (!n) return null;
    if (n.nodeType === 1) return n;
    if (n.nodeType === 9) return n.documentElement;
    return null;
  }
  function nodeTypeName(n) {
    switch (n.nodeType) {
      case 1: return 'element';
      case 3: return 'text';
      case 4: return 'cdata';
      case 7: return 'pi';
      case 8: return 'comment';
      case 9: return 'document';
      case 11: return 'fragment';
      default: return 'node';
    }
  }
  function parserMessage(raw) {
    var text = String(raw || '').replace(/\s+/g, ' ').trim();
    var line = null, col = null;
    var m = /line\s+(\d+)(?:[^0-9]+column\s+(\d+))?/i.exec(text);
    if (m) { line = Number(m[1]); col = m[2] ? Number(m[2]) : null; }
    var prefix = 'XML syntax error';
    if (line !== null) prefix += ' at line ' + line + (col !== null ? ', column ' + col : '');
    if (!text) return prefix;
    if (text.length > 240) text = text.slice(0, 240) + '...';
    return prefix + ': ' + text;
  }
  function parseText(text, limit) {
    if (typeof text !== 'string') return err('xml.parse expects a string');
    var max = Number(limit) > 0 ? Number(limit) : 5000000;
    if (text.length > max) return err('xml document is too large (max ' + max + ' characters)');
    if (/<!\s*(DOCTYPE|ENTITY)\b/i.test(text)) return err('DOCTYPE and ENTITY declarations are not allowed');
    if (typeof DOMParser === 'undefined') return err('XML parsing is not available in this environment');
    var parser = new DOMParser();
    var doc;
    try { doc = parser.parseFromString(text, 'application/xml'); }
    catch (e) { return err('XML parser failed: ' + (e && e.message ? e.message : String(e))); }
    var errors = doc.getElementsByTagName ? doc.getElementsByTagName('parsererror') : null;
    if (errors && errors.length > 0) return err(parserMessage(errors[0].textContent));
    if (doc.documentElement && doc.documentElement.nodeName === 'parsererror') return err(parserMessage(doc.documentElement.textContent));
    if (!doc.documentElement) return err('XML document has no root element');
    var all = doc.getElementsByTagName('*');
    if (all && all.length > 200000) return err('XML document has too many elements (max 200,000)');
    return ok(doc);
  }
  function decodeBytes(bytes, encoding) {
    if (!bytes || typeof bytes.length !== 'number') return err('xml.parseBytes expects a byte array');
    var data = new Uint8Array(bytes.length);
    for (var i = 0; i < bytes.length; i++) {
      var v = Number(bytes[i]);
      if (!isFinite(v) || v < 0 || v > 255) return err('xml.parseBytes contains an invalid byte');
      data[i] = v;
    }
    var label = encoding ? String(encoding) : '';
    if (!label) {
      if (data.length >= 2 && data[0] === 0xFE && data[1] === 0xFF) label = 'utf-16be';
      else if (data.length >= 2 && data[0] === 0xFF && data[1] === 0xFE) label = 'utf-16le';
      else {
        var head = '';
        var headLen = Math.min(data.length, 512);
        for (var hi = 0; hi < headLen; hi++) head += String.fromCharCode(data[hi]);
        var encMatch = /encoding\s*=\s*["']([^"']+)["']/i.exec(head);
        label = encMatch ? encMatch[1] : 'utf-8';
      }
    }
    try {
      return ok(new TextDecoder(label, { fatal: false }).decode(data));
    } catch (e) {
      return err('unsupported XML encoding: ' + label);
    }
  }
  function parseBytes(bytes, encoding) {
    var decoded = decodeBytes(bytes, encoding);
    if (decoded[1] !== null) return decoded;
    return parseText(decoded[0]);
  }
  function root(doc) { return ok(elementOf(doc)); }
  function parent(node) {
    var n = nodeOf(node);
    return n && n.parentNode ? ok(n.parentNode) : ok(null);
  }
  function children(node) {
    var n = nodeOf(node);
    var out = [];
    if (n) for (var i = 0; i < n.childNodes.length; i++) if (n.childNodes[i].nodeType === 1) out.push(n.childNodes[i]);
    return ok(out);
  }
  function childNodes(node) {
    var n = nodeOf(node);
    var out = [];
    if (n) for (var i = 0; i < n.childNodes.length; i++) out.push(n.childNodes[i]);
    return ok(out);
  }
  function childCount(node) {
    var n = nodeOf(node);
    if (!n) return ok(0);
    var count = 0;
    for (var i = 0; i < n.childNodes.length; i++) if (n.childNodes[i].nodeType === 1) count++;
    return ok(count);
  }
  function sibling(node, next) {
    var n = nodeOf(node);
    if (!n) return ok(null);
    var s = next ? n.nextSibling : n.previousSibling;
    while (s && s.nodeType !== 1) s = next ? s.nextSibling : s.previousSibling;
    return ok(s || null);
  }
  function descendants(node) {
    var n = nodeOf(node);
    if (!n) return ok([]);
    var list = n.getElementsByTagName ? n.getElementsByTagName('*') : [];
    return ok(Array.prototype.slice.call(list));
  }
  function findAll(node, name) {
    var n = nodeOf(node);
    if (!n || typeof name !== 'string') return ok([]);
    var list = n.getElementsByTagName(name);
    return ok(Array.prototype.slice.call(list));
  }
  function find(node, name) {
    var list = findAll(node, name);
    if (list[1] !== null) return list;
    return ok(list[0].length ? list[0][0] : null);
  }
  function findNs(node, localName, namespaceURI) {
    var n = nodeOf(node);
    var out = [], target = String(localName || ''), ns = namespaceURI == null ? null : String(namespaceURI);
    if (!n) return ok(out);
    var all = n.getElementsByTagName ? n.getElementsByTagName('*') : [];
    for (var i = 0; i < all.length; i++) {
      if (all[i].localName === target && (ns === null || all[i].namespaceURI === ns)) out.push(all[i]);
    }
    return ok(out);
  }
  function has(node, name) {
    var found = find(node, name);
    return ok(found[1] === null && found[0] !== null);
  }
  function name(node) {
    var n = nodeOf(node);
    if (!n) return ok('');
    if (n.nodeType === 7) return ok(n.target || n.nodeName || '');
    return ok(n.nodeName || '');
  }
  function localName(node) { var n = nodeOf(node); return ok(n && n.localName ? n.localName : (n ? n.nodeName : '')); }
  function namespace(node) { var n = nodeOf(node); return ok(n && n.namespaceURI ? n.namespaceURI : ''); }
  function prefix(node) { var n = nodeOf(node); return ok(n && n.prefix ? n.prefix : ''); }
  function directText(n) {
    var out = '';
    if (!n) return out;
    for (var i = 0; i < n.childNodes.length; i++) {
      var c = n.childNodes[i];
      if (c.nodeType === 3 || c.nodeType === 4) out += c.nodeValue || '';
    }
    return out;
  }
  function text(node) { var n = nodeOf(node); return ok(directText(n)); }
  function content(node) { var n = nodeOf(node); return ok(n && n.textContent != null ? n.textContent : ''); }
  function attr(node, name) {
    var n = elementOf(node);
    if (!n || typeof name !== 'string' || !n.hasAttribute(name)) return ok(null);
    return ok(n.getAttribute(name));
  }
  function attrNs(node, localName, namespaceURI) {
    var n = elementOf(node);
    if (!n) return ok(null);
    var v = n.getAttributeNS(namespaceURI == null ? null : String(namespaceURI), String(localName || ''));
    return ok(v === null ? null : v);
  }
  function attrs(node) {
    var n = elementOf(node), out = {};
    if (!n || !n.attributes) return ok(out);
    for (var i = 0; i < n.attributes.length; i++) out[n.attributes[i].name] = n.attributes[i].value;
    return ok(out);
  }
  function path(node) {
    var n = nodeOf(node);
    if (!n) return ok('');
    if (n.nodeType === 9) return ok('/');
    var parts = [];
    while (n && n.nodeType !== 9) {
      if (n.nodeType === 1) {
        var same = 0, index = 0, p = n.parentNode ? n.parentNode.firstChild : null;
        while (p) {
          if (p.nodeType === 1 && p.nodeName === n.nodeName) { same++; if (p === n) index = same; }
          p = p.nextSibling;
        }
        parts.unshift(n.nodeName + (same > 1 ? '[' + index + ']' : ''));
      } else if (n.nodeType === 3 || n.nodeType === 4) {
        var textIndex = 0, q = n.parentNode ? n.parentNode.firstChild : null;
        while (q) { if (q.nodeType === n.nodeType) { if (q === n) break; textIndex++; } q = q.nextSibling; }
        parts.unshift('text()[' + (textIndex + 1) + ']');
      } else if (n.nodeType === 8) parts.unshift('comment()');
      else if (n.nodeType === 7) parts.unshift('processing-instruction(' + (n.target || '') + ')');
      n = n.parentNode;
    }
    return ok('/' + parts.join('/'));
  }
  function createDocument(rootName) {
    if (typeof document === 'undefined' || !document.implementation) return err('XML documents are not available in this environment');
    var doc;
    try { doc = document.implementation.createDocument(null, null, null); }
    catch (e) { return err('cannot create XML document: ' + (e && e.message ? e.message : String(e))); }
    if (rootName != null && String(rootName) !== '') {
      var rootEl = doc.createElement(String(rootName));
      doc.appendChild(rootEl);
    }
    return ok(doc);
  }
  function importFor(parentNode, child) {
    var parentDoc = ownerOf(parentNode);
    if (!parentDoc) throw new Error('parent node has no document');
    if (child.nodeType === 9) child = child.documentElement;
    if (!child) throw new Error('child document is empty');
    if (child.ownerDocument !== parentDoc) child = parentDoc.importNode(child, true);
    return child;
  }
  function append(parentNode, child) {
    var p = nodeOf(parentNode), c = nodeOf(child);
    if (!p || !c) return err('xml.append expects two nodes');
    try { p.appendChild(importFor(p, c)); return [true, null]; }
    catch (e) { return err('xml.append: ' + (e && e.message ? e.message : String(e))); }
  }
  function prepend(parentNode, child) {
    var p = nodeOf(parentNode), c = nodeOf(child);
    if (!p || !c) return err('xml.prepend expects two nodes');
    try { p.insertBefore(importFor(p, c), p.firstChild); return [true, null]; }
    catch (e) { return err('xml.prepend: ' + (e && e.message ? e.message : String(e))); }
  }
  function insertBefore(parentNode, child, reference) {
    var p = nodeOf(parentNode), c = nodeOf(child), r = nodeOf(reference);
    if (!p || !c) return err('xml.insertBefore expects a parent and a child');
    try { p.insertBefore(importFor(p, c), r || null); return [true, null]; }
    catch (e) { return err('xml.insertBefore: ' + (e && e.message ? e.message : String(e))); }
  }
  function insertAfter(parentNode, child, reference) {
    var p = nodeOf(parentNode), c = nodeOf(child), r = nodeOf(reference);
    if (!p || !c) return err('xml.insertAfter expects a parent and a child');
    try { p.insertBefore(importFor(p, c), r && r.nextSibling ? r.nextSibling : null); return [true, null]; }
    catch (e) { return err('xml.insertAfter: ' + (e && e.message ? e.message : String(e))); }
  }
  function remove(node) {
    var n = nodeOf(node);
    if (!n || !n.parentNode) return err('xml.remove: node has no parent');
    try { n.parentNode.removeChild(n); return [true, null]; }
    catch (e) { return err('xml.remove: ' + (e && e.message ? e.message : String(e))); }
  }
  function replace(oldNode, newNode) {
    var oldN = nodeOf(oldNode), newN = nodeOf(newNode);
    if (!oldN || !newN || !oldN.parentNode) return err('xml.replace: invalid node or missing parent');
    try { oldN.parentNode.replaceChild(importFor(oldN.parentNode, newN), oldN); return [true, null]; }
    catch (e) { return err('xml.replace: ' + (e && e.message ? e.message : String(e))); }
  }
  function clear(node) {
    var n = nodeOf(node);
    if (!n) return err('xml.clear expects a node');
    try { while (n.firstChild) n.removeChild(n.firstChild); return [true, null]; }
    catch (e) { return err('xml.clear: ' + (e && e.message ? e.message : String(e))); }
  }
  function clone(node, deep) {
    var n = nodeOf(node);
    if (!n) return err('xml.clone expects a node');
    try { return ok(n.cloneNode(deep !== false)); }
    catch (e) { return err('xml.clone: ' + (e && e.message ? e.message : String(e))); }
  }
  function rename(node, newName) {
    var n = nodeOf(node), name = String(newName == null ? '' : newName);
    if (!n || n.nodeType !== 1) return err('xml.rename expects an element');
    if (!name) return err('xml.rename requires a non-empty name');
    try {
      var doc = ownerOf(n);
      var el = n.namespaceURI ? doc.createElementNS(n.namespaceURI, name) : doc.createElement(name);
      for (var i = 0; i < n.attributes.length; i++) {
        var a = n.attributes[i];
        if (a.namespaceURI) el.setAttributeNS(a.namespaceURI, a.name, a.value);
        else el.setAttribute(a.name, a.value);
      }
      while (n.firstChild) el.appendChild(n.firstChild);
      if (n.parentNode) n.parentNode.replaceChild(el, n);
      return ok(el);
    } catch (e) { return err('xml.rename: ' + (e && e.message ? e.message : String(e))); }
  }
  function setText(node, value) {
    var n = nodeOf(node);
    if (!n) return err('xml.setText expects a node');
    try {
      while (n.firstChild) n.removeChild(n.firstChild);
      n.appendChild(ownerOf(n).createTextNode(String(value == null ? '' : value)));
      return [true, null];
    } catch (e) { return err('xml.setText: ' + (e && e.message ? e.message : String(e))); }
  }
  function move(node, newParent) {
    var n = nodeOf(node), p = nodeOf(newParent);
    if (!n || !p) return err('xml.move expects two nodes');
    try { p.appendChild(importFor(p, n)); return [true, null]; }
    catch (e) { return err('xml.move: ' + (e && e.message ? e.message : String(e))); }
  }
  function reorder(parentNode, indexes) {
    var p = nodeOf(parentNode);
    if (!p || !indexes || typeof indexes.length !== 'number') return err('xml.reorder expects a parent and an index array');
    var elements = [];
    for (var i = 0; i < p.childNodes.length; i++) if (p.childNodes[i].nodeType === 1) elements.push(p.childNodes[i]);
    if (indexes.length !== elements.length) return err('xml.reorder must list every child exactly once');
    var seen = Object.create(null);
    var ordered = [];
    for (var j = 0; j < indexes.length; j++) {
      var idx = Number(indexes[j]);
      if (!isFinite(idx) || idx < 0 || idx >= elements.length || seen[idx]) return err('xml.reorder contains an invalid or duplicate index');
      seen[idx] = true; ordered.push(elements[idx]);
    }
    try { for (var k = 0; k < ordered.length; k++) p.appendChild(ordered[k]); return [true, null]; }
    catch (e) { return err('xml.reorder: ' + (e && e.message ? e.message : String(e))); }
  }
  function setAttr(node, name, value) {
    var n = elementOf(node);
    if (!n) return err('xml.setAttr expects an element');
    try { n.setAttribute(String(name), String(value == null ? '' : value)); return [true, null]; }
    catch (e) { return err('xml.setAttr: ' + (e && e.message ? e.message : String(e))); }
  }
  function setAttrNs(node, namespaceURI, qualifiedName, value) {
    var n = elementOf(node);
    if (!n) return err('xml.setAttrNs expects an element');
    try { n.setAttributeNS(namespaceURI == null ? null : String(namespaceURI), String(qualifiedName), String(value == null ? '' : value)); return [true, null]; }
    catch (e) { return err('xml.setAttrNs: ' + (e && e.message ? e.message : String(e))); }
  }
  function removeAttr(node, name) {
    var n = elementOf(node);
    if (!n) return err('xml.removeAttr expects an element');
    try { n.removeAttribute(String(name)); return [true, null]; }
    catch (e) { return err('xml.removeAttr: ' + (e && e.message ? e.message : String(e))); }
  }
  function renameAttr(node, oldName, newName) {
    var n = elementOf(node);
    if (!n || !n.hasAttribute(String(oldName))) return err('xml.renameAttr: attribute not found');
    try {
      var value = n.getAttribute(String(oldName));
      n.removeAttribute(String(oldName));
      n.setAttribute(String(newName), value);
      return [true, null];
    } catch (e) { return err('xml.renameAttr: ' + (e && e.message ? e.message : String(e))); }
  }
  function hasAttr(node, name) { var n = elementOf(node); return ok(!!n && n.hasAttribute(String(name))); }
  function createElement(doc, name) {
    var d = ownerOf(doc) || (doc && doc.nodeType === 9 ? doc : null);
    if (!d) return err('xml.createElement expects a document');
    try { return ok(d.createElement(String(name))); }
    catch (e) { return err('xml.createElement: ' + (e && e.message ? e.message : String(e))); }
  }
  function createElementNs(doc, namespaceURI, qualifiedName) {
    var d = ownerOf(doc) || (doc && doc.nodeType === 9 ? doc : null);
    if (!d) return err('xml.createElementNs expects a document');
    try { return ok(d.createElementNS(namespaceURI == null ? null : String(namespaceURI), String(qualifiedName))); }
    catch (e) { return err('xml.createElementNs: ' + (e && e.message ? e.message : String(e))); }
  }
  function createText(doc, value) {
    var d = ownerOf(doc) || (doc && doc.nodeType === 9 ? doc : null);
    if (!d) return err('xml.createText expects a document');
    try { return ok(d.createTextNode(String(value == null ? '' : value))); }
    catch (e) { return err('xml.createText: ' + (e && e.message ? e.message : String(e))); }
  }
  function createComment(doc, value) {
    var d = ownerOf(doc) || (doc && doc.nodeType === 9 ? doc : null);
    if (!d) return err('xml.createComment expects a document');
    try { return ok(d.createComment(String(value == null ? '' : value))); }
    catch (e) { return err('xml.createComment: ' + (e && e.message ? e.message : String(e))); }
  }
  function createCdata(doc, value) {
    var d = ownerOf(doc) || (doc && doc.nodeType === 9 ? doc : null);
    if (!d) return err('xml.createCdata expects a document');
    try { return ok(d.createCDATASection(String(value == null ? '' : value))); }
    catch (e) { return err('xml.createCdata: ' + (e && e.message ? e.message : String(e))); }
  }
  function createPi(doc, target, data) {
    var d = ownerOf(doc) || (doc && doc.nodeType === 9 ? doc : null);
    if (!d) return err('xml.createPi expects a document');
    try { return ok(d.createProcessingInstruction(String(target), String(data == null ? '' : data))); }
    catch (e) { return err('xml.createPi: ' + (e && e.message ? e.message : String(e))); }
  }
  function escapeText(value) {
    return String(value == null ? '' : value).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }
  function escapeAttr(value) {
    return String(value == null ? '' : value).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }
  function serializeAttributes(el, sortAttributes) {
    var list = [];
    for (var i = 0; i < el.attributes.length; i++) list.push(el.attributes[i]);
    if (sortAttributes) list.sort(function (a, b) { return a.name < b.name ? -1 : (a.name > b.name ? 1 : 0); });
    var out = '';
    for (var j = 0; j < list.length; j++) out += ' ' + list[j].name + '="' + escapeAttr(list[j].value) + '"';
    return out;
  }
  function serializeNode(n, level, opts) {
    var indent = opts.indentWidth;
    var pad = indent > 0 ? new Array(level + 1).join(new Array(indent + 1).join(' ')) : '';
    if (n.nodeType === 3) return escapeText(n.nodeValue);
    if (n.nodeType === 4) return pad + '<![CDATA[' + String(n.nodeValue || '').replace(/]]>/g, ']]]]><![CDATA[>') + ']]>';
    if (n.nodeType === 8) return pad + '<!--' + String(n.nodeValue || '') + '-->';
    if (n.nodeType === 7) return pad + '<?' + (n.target || '') + (n.data ? ' ' + n.data : '') + '?>';
    if (n.nodeType !== 1) return '';
    var open = '<' + n.nodeName + serializeAttributes(n, opts.sortAttributes);
    var kids = [];
    for (var i = 0; i < n.childNodes.length; i++) kids.push(n.childNodes[i]);
    if (kids.length === 0) return pad + open + '/>';
    var hasElement = false, hasNonWhitespaceText = false;
    for (var h = 0; h < kids.length; h++) {
      if (kids[h].nodeType === 1) hasElement = true;
      if (kids[h].nodeType === 3 && String(kids[h].nodeValue || '').trim() !== '') hasNonWhitespaceText = true;
    }
    var pretty = indent > 0 && hasElement && !hasNonWhitespaceText;
    if (!pretty) {
      var inline = '';
      for (var k = 0; k < kids.length; k++) inline += serializeNode(kids[k], 0, opts);
      return pad + open + '>' + inline + '</' + n.nodeName + '>';
    }
    var body = '';
    for (var j = 0; j < kids.length; j++) {
      if (kids[j].nodeType === 3 && String(kids[j].nodeValue || '').trim() === '') continue;
      body += '\n' + serializeNode(kids[j], level + 1, opts);
    }
    return pad + open + '>' + body + '\n' + pad + '</' + n.nodeName + '>';
  }
  function stringify(node, options) {
    var n = nodeOf(node);
    if (!n) return err('xml.stringify expects a node');
    var opts = options && typeof options === 'object' ? options : {};
    var indent = Number(opts.indent);
    if (opts.indent === true) indent = 2;
    if (!(indent > 0)) indent = 0;
    var enc = String(opts.encoding || 'UTF-8');
    var declaration = opts.declaration !== false;
    var sortAttributes = !!opts.sortAttributes;
    var config = { indentWidth: indent, sortAttributes: sortAttributes };
    try {
      var body = '';
      if (n.nodeType === 9) {
        for (var i = 0; i < n.childNodes.length; i++) {
          if (i > 0) body += indent > 0 ? '\n' : '';
          body += serializeNode(n.childNodes[i], 0, config);
        }
        if (declaration) body = '<?xml version="1.0" encoding="' + enc + '"?>' + (indent > 0 ? '\n' : '') + body;
      } else {
        body = serializeNode(n, 0, config);
        if (declaration && n.nodeType === 1 && opts.declaration === true) body = '<?xml version="1.0" encoding="' + enc + '"?>' + (indent > 0 ? '\n' : '') + body;
      }
      return ok(body);
    } catch (e) { return err('xml.stringify: ' + (e && e.message ? e.message : String(e))); }
  }
  function toBytes(node, encoding, options) {
    var text = stringify(node, options);
    if (text[1] !== null) return text;
    var enc = String(encoding || 'utf-8').toLowerCase();
    var str = text[0];
    try {
      if (enc === 'utf-8' || enc === 'utf8') return ok(Array.prototype.slice.call(new TextEncoder().encode(str)));
      if (enc === 'utf-16' || enc === 'utf-16le' || enc === 'utf16le') {
        var le = [0xFF, 0xFE];
        for (var i = 0; i < str.length; i++) { var c = str.charCodeAt(i); le.push(c & 255, c >> 8); }
        return ok(le);
      }
      if (enc === 'utf-16be' || enc === 'utf16be') {
        var be = [0xFE, 0xFF];
        for (var j = 0; j < str.length; j++) { var d = str.charCodeAt(j); be.push(d >> 8, d & 255); }
        return ok(be);
      }
      return err('unsupported output encoding: ' + encoding);
    } catch (e) { return err('xml.toBytes: ' + (e && e.message ? e.message : String(e))); }
  }
  function toJson(node, options) {
    var n = nodeOf(node);
    if (!n) return err('xml.toJson expects a node');
    var opts = options && typeof options === 'object' ? options : {};
    var withAttributes = opts.attributes !== false;
    var withText = opts.text !== false;
    var trim = opts.trim !== false;
    function lossless(n) {
      if (!n) return null;
      if (n.nodeType === 9) {
        var children = [];
        for (var i = 0; i < n.childNodes.length; i++) children.push(lossless(n.childNodes[i]));
        return { type: 'document', children: children };
      }
      if (n.nodeType === 1) {
        var node = { type: 'element', name: n.nodeName, localName: n.localName || n.nodeName, namespace: n.namespaceURI || '', prefix: n.prefix || '', attributes: {}, children: [] };
        for (var a = 0; a < n.attributes.length; a++) node.attributes[n.attributes[a].name] = n.attributes[a].value;
        for (var c = 0; c < n.childNodes.length; c++) node.children.push(lossless(n.childNodes[c]));
        return node;
      }
      if (n.nodeType === 3) return { type: 'text', text: n.nodeValue || '' };
      if (n.nodeType === 4) return { type: 'cdata', text: n.nodeValue || '' };
      if (n.nodeType === 8) return { type: 'comment', text: n.nodeValue || '' };
      if (n.nodeType === 7) return { type: 'pi', target: n.target || '', data: n.data || '' };
      return { type: nodeTypeName(n), text: n.nodeValue || '' };
    }
    function elementValue(el) {
      var obj = {}, has = false;
      if (withAttributes && el.attributes) {
        for (var i = 0; i < el.attributes.length; i++) { obj['@' + el.attributes[i].name] = el.attributes[i].value; has = true; }
      }
      var direct = directText(el);
      if (withText && (trim ? direct.trim() !== '' : direct !== '')) { obj['#text'] = trim ? direct.trim() : direct; has = true; }
      for (var j = 0; j < el.childNodes.length; j++) {
        var ch = el.childNodes[j];
        if (ch.nodeType !== 1) continue;
        var key = ch.nodeName, value = elementValue(ch);
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
          if (!Array.isArray(obj[key])) obj[key] = [obj[key]];
          obj[key].push(value);
        } else obj[key] = value;
        has = true;
      }
      if (!has) return trim ? direct.trim() : direct;
      return obj;
    }
    try {
      if (opts.lossless) return ok(lossless(n));
      if (n.nodeType === 9) {
        var out = {};
        for (var k = 0; k < n.childNodes.length; k++) if (n.childNodes[k].nodeType === 1) out[n.childNodes[k].nodeName] = elementValue(n.childNodes[k]);
        return ok(out);
      }
      if (n.nodeType !== 1) return ok(directText(n));
      return ok(elementValue(n));
    } catch (e) { return err('xml.toJson: ' + (e && e.message ? e.message : String(e))); }
  }
  function fromJson(value, rootName) {
    var name = String(rootName == null || rootName === '' ? 'root' : rootName);
    var docRes = createDocument();
    if (docRes[1] !== null) return docRes;
    var doc = docRes[0];
    function buildTree(node) {
      if (!node || typeof node !== 'object') return null;
      if (node.type === 'element') {
        var el = node.namespace ? doc.createElementNS(String(node.namespace), String(node.name)) : doc.createElement(String(node.name));
        var attrs = node.attributes || {};
        var aks = Object.keys(attrs);
        for (var ai = 0; ai < aks.length; ai++) {
          var an = aks[ai];
          if (an === 'xmlns' || an.indexOf('xmlns:') === 0) {
            try { el.setAttributeNS('http://www.w3.org/2000/xmlns/', an, String(attrs[an])); } catch (e) {}
            continue;
          }
          try { el.setAttribute(an, String(attrs[an])); } catch (e) {}
        }
        var kids = node.children || [];
        for (var ki = 0; ki < kids.length; ki++) {
          var child = buildTree(kids[ki]);
          if (child) el.appendChild(child);
        }
        return el;
      }
      if (node.type === 'text') return doc.createTextNode(String(node.text || ''));
      if (node.type === 'cdata') return doc.createCDATASection(String(node.text || ''));
      if (node.type === 'comment') return doc.createComment(String(node.text || ''));
      if (node.type === 'pi') return doc.createProcessingInstruction(String(node.target || 'pi'), String(node.data || ''));
      return null;
    }
    function build(name, value) {
      var el = doc.createElement(name);
      if (value === null || value === undefined) return el;
      if (Array.isArray(value)) {
        for (var i = 0; i < value.length; i++) el.appendChild(build(name, value[i]));
        return el;
      }
      if (typeof value !== 'object') { el.appendChild(doc.createTextNode(String(value))); return el; }
      var keys = Object.keys(value);
      for (var k = 0; k < keys.length; k++) {
        var key = keys[k], v = value[key];
        if (key.charAt(0) === '@') { el.setAttribute(key.slice(1), String(v == null ? '' : v)); continue; }
        if (key === '#text') { el.appendChild(doc.createTextNode(String(v == null ? '' : v))); continue; }
        if (Array.isArray(v)) {
          for (var a = 0; a < v.length; a++) el.appendChild(build(key, v[a]));
        } else {
          el.appendChild(build(key, v));
        }
      }
      return el;
    }
    try {
      var rootEl;
      if (value && typeof value === 'object' && value.type === 'document') {
        var dkids = value.children || [];
        for (var di = 0; di < dkids.length; di++) {
          var dchild = buildTree(dkids[di]);
          if (dchild) doc.appendChild(dchild);
        }
        return ok(doc);
      }
      if (value && typeof value === 'object' && value.type === 'element') {
        rootEl = buildTree(value);
      } else if (Array.isArray(value)) {
        rootEl = doc.createElement(name);
        for (var ri = 0; ri < value.length; ri++) rootEl.appendChild(build('item', value[ri]));
      } else {
        rootEl = build(name, value);
      }
      doc.appendChild(rootEl);
      return ok(doc);
    } catch (e) { return err('xml.fromJson: ' + (e && e.message ? e.message : String(e))); }
  }
  function toCsv(node, rowName, columns) {
    var all = findAll(node, rowName);
    if (all[1] !== null) return all;
    var rows = all[0], cols = [];
    if (Array.isArray(columns)) cols = columns.map(function (c) { return String(c); });
    else if (typeof columns === 'string' && columns.trim() !== '') cols = columns.split(',').map(function (c) { return c.trim(); });
    if (!cols.length) {
      var seen = Object.create(null);
      for (var i = 0; i < rows.length; i++) {
        var attrs = attrsOf(rows[i]);
        for (var a in attrs) if (!seen['@' + a]) { seen['@' + a] = true; cols.push('@' + a); }
        for (var j = 0; j < rows[i].childNodes.length; j++) {
          var ch = rows[i].childNodes[j];
          if (ch.nodeType === 1 && !seen[ch.nodeName]) { seen[ch.nodeName] = true; cols.push(ch.nodeName); }
        }
      }
    }
    function cell(row, col) {
      if (col.charAt(0) === '@') return row.getAttribute(col.slice(1)) || '';
      var child = find(row, col);
      return child[1] === null && child[0] ? directText(child[0]) : '';
    }
    function csv(v) { var s = String(v == null ? '' : v); return '"' + s.replace(/"/g, '""') + '"'; }
    var lines = [cols.map(csv).join(',')];
    for (var r = 0; r < rows.length; r++) lines.push(cols.map(function (c) { return csv(cell(rows[r], c)); }).join(','));
    return ok(lines.join('\n'));
  }
  function attrsOf(el) { var o = {}; if (el && el.attributes) for (var i = 0; i < el.attributes.length; i++) o[el.attributes[i].name] = el.attributes[i].value; return o; }
  function xpath(node, expression, namespaces) {
    var n = nodeOf(node), doc = ownerOf(n);
    if (!n || !doc) return err('xml.xpath expects a node');
    if (typeof XPathResult === 'undefined' || !doc.evaluate) return err('XPath is not available in this environment');
    var ns = namespaces && typeof namespaces === 'object' ? namespaces : {};
    var resolver = function (prefix) {
      if (prefix === '' || prefix === null || prefix === undefined) return ns[''] || null;
      return Object.prototype.hasOwnProperty.call(ns, prefix) ? ns[prefix] : null;
    };
    try {
      var result = doc.evaluate(String(expression), n, resolver, XPathResult.ANY_TYPE, null);
      var rt = result.resultType;
      var snapshotTypes = [XPathResult.UNORDERED_NODE_SNAPSHOT_TYPE, XPathResult.ORDERED_NODE_SNAPSHOT_TYPE, XPathResult.NODESET];
      var iteratorTypes = [XPathResult.UNORDERED_NODE_ITERATOR_TYPE, XPathResult.ORDERED_NODE_ITERATOR_TYPE];
      if (snapshotTypes.indexOf(rt) >= 0) {
        var out = [];
        for (var i = 0; i < result.snapshotLength; i++) out.push(result.snapshotItem(i));
        return ok(out);
      }
      if (iteratorTypes.indexOf(rt) >= 0) {
        var iter = [];
        var nextNode = result.iterateNext();
        while (nextNode) { iter.push(nextNode); nextNode = result.iterateNext(); }
        return ok(iter);
      }
      if (result.resultType === XPathResult.STRING_TYPE) return ok(result.stringValue);
      if (result.resultType === XPathResult.NUMBER_TYPE) return ok(result.numberValue);
      if (result.resultType === XPathResult.BOOLEAN_TYPE) return ok(result.booleanValue);
      return ok(null);
    } catch (e) { return err('xml.xpath: ' + (e && e.message ? e.message : String(e))); }
  }
  function xpathOne(node, expression, namespaces) {
    var r = xpath(node, expression, namespaces);
    if (r[1] !== null) return r;
    if (Array.isArray(r[0])) return ok(r[0].length ? r[0][0] : null);
    return r;
  }
  function findAttr(node, name, value) {
    var n = nodeOf(node), out = [], attrName = String(name || '');
    if (!n) return ok(out);
    var all = n.getElementsByTagName ? n.getElementsByTagName('*') : [];
    for (var i = 0; i < all.length; i++) {
      if (!all[i].hasAttribute(attrName)) continue;
      if (value === undefined || value === null || all[i].getAttribute(attrName) === String(value)) out.push(all[i]);
    }
    return ok(out);
  }
  function findText(node, term, regex) {
    var n = nodeOf(node), out = [], needle = String(term == null ? '' : term);
    if (!n) return ok(out);
    var candidates = [n].concat(Array.prototype.slice.call(n.getElementsByTagName ? n.getElementsByTagName('*') : []));
    var re = null;
    if (regex) { try { re = new RegExp(needle); } catch (e) { return err('xml.findText: invalid regex'); } }
    for (var i = 0; i < candidates.length; i++) {
      var v = directText(candidates[i]);
      if (re ? re.test(v) : v.indexOf(needle) >= 0) out.push(candidates[i]);
    }
    return ok(out);
  }
  function xmlChildNamed(el, local) {
    if (!el) return null;
    for (var i = 0; i < el.childNodes.length; i++) if (el.childNodes[i].nodeType === 1 && el.childNodes[i].localName === local) return el.childNodes[i];
    return null;
  }
  function xmlChildrenNamed(el, local) {
    var out = [];
    if (!el) return out;
    for (var i = 0; i < el.childNodes.length; i++) if (el.childNodes[i].nodeType === 1 && el.childNodes[i].localName === local) out.push(el.childNodes[i]);
    return out;
  }
  function xmlAttrValue(el, name) { return el && el.hasAttribute(name) ? el.getAttribute(name) : null; }
  function xmlLocalRef(name) { var v = String(name || ''); var i = v.indexOf(':'); return i >= 0 ? v.slice(i + 1) : v; }
  function xmlTypeOk(text, typeName) {
    var t = xmlLocalRef(typeName);
    var v = String(text == null ? '' : text);
    if (t === 'string' || t === 'normalizedString' || t === 'token' || t === 'anyURI' || t === '') return true;
    if (t === 'int' || t === 'integer' || t === 'long' || t === 'short' || t === 'byte') return /^-?[0-9]+$/.test(v);
    if (t === 'decimal' || t === 'double' || t === 'float') return v !== '' && isFinite(Number(v));
    if (t === 'boolean') return v === 'true' || v === 'false' || v === '1' || v === '0';
    if (t === 'date') return /^[0-9]{4}-[0-9]{2}-[0-9]{2}$/.test(v) && !isNaN(Date.parse(v));
    if (t === 'dateTime') return !isNaN(Date.parse(v));
    if (t === 'time') return /^[0-9]{2}:[0-9]{2}:[0-9]{2}/.test(v);
    return true;
  }
  function validateXsdCore(doc, xsdText) {
  var errors = [];
  function fail(path, message) {
    if (errors.length < 100) errors.push(path + ': ' + message);
  }
  var rootRes = root(doc);
  var instanceRoot = rootRes[0];
  if (!instanceRoot) return { errors: errors, error: 'XSD: document has no root element' };
  var parsed = parseText(String(xsdText));
  if (parsed[1] !== null) return { errors: errors, error: 'XSD: ' + parsed[1] };
  var schema = parsed[0].documentElement;
  if (!schema || schema.localName !== 'schema') return { errors: errors, error: 'XSD: root element must be xs:schema' };
  var targetNS = xmlAttrValue(schema, 'targetNamespace');
  if (targetNS && instanceRoot.namespaceURI !== targetNS) fail('/' + instanceRoot.nodeName, 'root namespace does not match targetNamespace');

  var globals = Object.create(null), complexTypes = Object.create(null), simpleTypes = Object.create(null);
  var attributeGroups = Object.create(null), globalAttributes = Object.create(null);
  var allSchemaChildren = [];
  for (var si = 0; si < schema.childNodes.length; si++) if (schema.childNodes[si].nodeType === 1) allSchemaChildren.push(schema.childNodes[si]);
  for (var gi = 0; gi < allSchemaChildren.length; gi++) {
    var g = allSchemaChildren[gi];
    var gname = xmlAttrValue(g, 'name');
    if (g.localName === 'element' && gname) globals[gname] = g;
    else if (g.localName === 'complexType' && gname) complexTypes[gname] = g;
    else if (g.localName === 'simpleType' && gname) simpleTypes[gname] = g;
    else if (g.localName === 'attributeGroup' && gname) attributeGroups[gname] = g;
    else if (g.localName === 'attribute' && gname) globalAttributes[gname] = g;
  }

  function resolveElement(decl) {
    if (!decl) return null;
    var ref = xmlAttrValue(decl, 'ref');
    if (ref && globals[xmlLocalRef(ref)]) return globals[xmlLocalRef(ref)];
    return decl;
  }
  function resolveAttribute(decl) {
    if (!decl) return null;
    var ref = xmlAttrValue(decl, 'ref');
    if (ref && globalAttributes[xmlLocalRef(ref)]) {
      var target = globalAttributes[xmlLocalRef(ref)];
      if (!xmlAttrValue(decl, 'type')) decl = target;
    }
    return decl;
  }
  function typeDeclaration(name, inlineComplex, inlineSimple) {
    if (inlineComplex) return inlineComplex;
    if (inlineSimple) return inlineSimple;
    if (!name) return null;
    var local = xmlLocalRef(name);
    return complexTypes[local] || simpleTypes[local] || null;
  }
  function typeOfElement(decl) {
    var typeName = xmlAttrValue(decl, 'type');
    var inlineComplex = xmlChildNamed(decl, 'complexType');
    var inlineSimple = xmlChildNamed(decl, 'simpleType');
    return { name: typeName, decl: typeDeclaration(typeName, inlineComplex, inlineSimple) };
  }
  function checkSimpleType(value, typeName, typeDecl, path) {
    var text = String(value == null ? '' : value);
    if (typeName && !typeDecl) {
      if (!xmlTypeOk(text, typeName)) fail(path, 'value "' + text + '" is not ' + xmlLocalRef(typeName));
      return;
    }
    if (!typeDecl) return;
    if (typeDecl.localName === 'simpleType') {
      var restriction = xmlChildNamed(typeDecl, 'restriction');
      if (restriction) {
        function facetValue(name) {
          var facet = xmlChildNamed(restriction, name);
          return facet ? xmlAttrValue(facet, 'value') : null;
        }
        var base = xmlAttrValue(restriction, 'base');
        if (base) checkSimpleType(text, base, simpleTypes[xmlLocalRef(base)] || null, path);
        var enums = xmlChildrenNamed(restriction, 'enumeration');
        if (enums.length) {
          var found = false;
          for (var e = 0; e < enums.length; e++) if (xmlAttrValue(enums[e], 'value') === text) found = true;
          if (!found) fail(path, 'value "' + text + '" is not in the allowed enumeration');
        }
        var length = facetValue('length');
        if (length !== null && text.length !== Number(length)) fail(path, 'value must have exactly ' + length + ' characters');
        var minLen = facetValue('minLength');
        if (minLen !== null && text.length < Number(minLen)) fail(path, 'value is shorter than minLength ' + minLen);
        var maxLen = facetValue('maxLength');
        if (maxLen !== null && text.length > Number(maxLen)) fail(path, 'value is longer than maxLength ' + maxLen);
        var pattern = facetValue('pattern');
        if (pattern) {
          try { if (!(new RegExp(pattern)).test(text)) fail(path, 'value does not match the pattern'); }
          catch (err) { fail(path, 'invalid pattern "' + pattern + '"'); }
        }
        if (text !== '' && isFinite(Number(text))) {
          var num = Number(text);
          var minInc = facetValue('minInclusive');
          if (minInc !== null && num < Number(minInc)) fail(path, 'value must be >= ' + minInc);
          var maxInc = facetValue('maxInclusive');
          if (maxInc !== null && num > Number(maxInc)) fail(path, 'value must be <= ' + maxInc);
          var minExc = facetValue('minExclusive');
          if (minExc !== null && num <= Number(minExc)) fail(path, 'value must be > ' + minExc);
          var maxExc = facetValue('maxExclusive');
          if (maxExc !== null && num >= Number(maxExc)) fail(path, 'value must be < ' + maxExc);
          var multiple = facetValue('multipleOf');
          if (multiple !== null && Number(multiple) > 0) {
            var q = num / Number(multiple);
            if (Math.abs(q - Math.round(q)) > 1e-9) fail(path, 'value must be a multiple of ' + multiple);
          }
          var totalDigits = facetValue('totalDigits');
          if (totalDigits !== null && text.replace(/[^0-9]/g, '').length > Number(totalDigits)) fail(path, 'value has too many digits');
          var fractionDigits = facetValue('fractionDigits');
          if (fractionDigits !== null && text.split('.')[1] && text.split('.')[1].length > Number(fractionDigits)) fail(path, 'value has too many fraction digits');
        }
        var list = xmlChildNamed(typeDecl, 'list');
        if (list) {
          var itemType = xmlAttrValue(list, 'itemType');
          var values = text.split(/\s+/);
          for (var li = 0; li < values.length; li++) if (values[li] !== '' && itemType) checkSimpleType(values[li], itemType, simpleTypes[xmlLocalRef(itemType)] || null, path);
        }
        return;
      }
      var union = xmlChildNamed(typeDecl, 'union');
      if (union) return; // union member validation is not implemented
    }
  }
  function collectAttributes(typeDecl, seen) {
    var out = [];
    if (!typeDecl) return out;
    if (seen.has(typeDecl)) return out;
    seen.add(typeDecl);
    var direct = xmlChildrenNamed(typeDecl, 'attribute');
    for (var i = 0; i < direct.length; i++) out.push(direct[i]);
    var groups = xmlChildrenNamed(typeDecl, 'attributeGroup');
    for (var g = 0; g < groups.length; g++) {
      var ref = xmlAttrValue(groups[g], 'ref');
      if (ref && attributeGroups[xmlLocalRef(ref)]) {
        var nested = collectAttributes(attributeGroups[xmlLocalRef(ref)], seen);
        for (var n = 0; n < nested.length; n++) out.push(nested[n]);
      }
    }
    var simpleContent = xmlChildNamed(typeDecl, 'simpleContent');
    var complexContent = xmlChildNamed(typeDecl, 'complexContent');
    var content = simpleContent || complexContent;
    if (content) {
      var ext = xmlChildNamed(content, 'extension') || xmlChildNamed(content, 'restriction');
      if (ext) {
        var base = xmlAttrValue(ext, 'base');
        var baseType = base ? typeDeclaration(base, null, null) : null;
        if (baseType) {
          var inherited = collectAttributes(baseType, seen);
          for (var b = 0; b < inherited.length; b++) out.push(inherited[b]);
        }
        var own = xmlChildrenNamed(ext, 'attribute');
        for (var o = 0; o < own.length; o++) out.push(own[o]);
        var extGroups = xmlChildrenNamed(ext, 'attributeGroup');
        for (var eg = 0; eg < extGroups.length; eg++) {
          var egRef = xmlAttrValue(extGroups[eg], 'ref');
          if (egRef && attributeGroups[xmlLocalRef(egRef)]) {
            var egAttrs = collectAttributes(attributeGroups[xmlLocalRef(egRef)], seen);
            for (var ega = 0; ega < egAttrs.length; ega++) out.push(egAttrs[ega]);
          }
        }
      }
    }
    return out;
  }
  function collectGroups(typeDecl, seen) {
    var out = [];
    if (!typeDecl || seen.has(typeDecl)) return out;
    seen.add(typeDecl);
    var direct = ['sequence', 'all', 'choice'];
    for (var i = 0; i < direct.length; i++) {
      var group = xmlChildNamed(typeDecl, direct[i]);
      if (group) out.push(group);
    }
    var complexContent = xmlChildNamed(typeDecl, 'complexContent');
    if (complexContent) {
      var ext = xmlChildNamed(complexContent, 'extension') || xmlChildNamed(complexContent, 'restriction');
      if (ext) {
        var base = xmlAttrValue(ext, 'base');
        var baseType = base ? typeDeclaration(base, null, null) : null;
        if (baseType) {
          var inherited = collectGroups(baseType, seen);
          for (var b = 0; b < inherited.length; b++) out.push(inherited[b]);
        }
        var own = xmlChildNamed(ext, 'sequence') || xmlChildNamed(ext, 'all') || xmlChildNamed(ext, 'choice');
        if (own && out.indexOf(own) < 0) out.push(own);
      }
    }
    return out;
  }
  function checkAttributes(el, typeDecl, path) {
    var declarations = collectAttributes(typeDecl, new Set());
    var declared = Object.create(null);
    for (var i = 0; i < declarations.length; i++) {
      var a = resolveAttribute(declarations[i]);
      var name = xmlAttrValue(a, 'name');
      if (!name) name = xmlLocalRef(xmlAttrValue(a, 'ref'));
      if (!name) continue;
      declared[name] = true;
      var value = el.hasAttribute(name) ? el.getAttribute(name) : null;
      if (value === null) {
        if (xmlAttrValue(a, 'use') === 'required') fail(path, 'missing required attribute "' + name + '"');
        continue;
      }
      var type = xmlAttrValue(a, 'type');
      if (type) checkSimpleType(value, type, simpleTypes[xmlLocalRef(type)] || null, path + '/@' + name);
    }
    var any = xmlChildNamed(typeDecl, 'anyAttribute') !== null;
    if (!any) {
      for (var k = 0; k < el.attributes.length; k++) {
        var an = el.attributes[k].name;
        if (an === 'xmlns' || an.indexOf('xmlns:') === 0) continue;
        if (!declared[an]) fail(path, 'unexpected attribute "' + an + '"');
      }
    }
  }
  function elementName(part) {
    var decl = resolveElement(part);
    return xmlAttrValue(decl, 'name') || xmlLocalRef(xmlAttrValue(decl, 'ref'));
  }
  function checkGroup(el, group, path) {
    var parts = xmlChildrenNamed(group, 'element');
    var children = [];
    for (var c = 0; c < el.childNodes.length; c++) if (el.childNodes[c].nodeType === 1) children.push(el.childNodes[c]);
    if (!parts.length) return;
    if (group.localName === 'all') {
      var used = Object.create(null);
      for (var p = 0; p < parts.length; p++) {
        var part = parts[p], name = elementName(part);
        var min = xmlAttrValue(part, 'minOccurs');
        var max = xmlAttrValue(part, 'maxOccurs');
        var minCount = min === null ? 1 : Number(min);
        var maxCount = max === null ? 1 : (max === 'unbounded' ? Infinity : Number(max));
        var count = 0;
        for (var ci = 0; ci < children.length; ci++) {
          if (children[ci].localName !== name) continue;
          count++;
          used[ci] = true;
          if (count <= maxCount) checkElement(children[ci], part, path + '/' + children[ci].nodeName);
        }
        if (count < minCount) fail(path, 'expected at least ' + minCount + ' element "' + name + '"');
        if (count > maxCount) fail(path, 'too many elements "' + name + '" (max ' + max + ')');
      }
      for (var ua = 0; ua < children.length; ua++) if (!used[ua]) fail(path, 'unexpected element "' + children[ua].nodeName + '"');
    } else if (group.localName === 'choice') {
      var matched = false;
      var usedChoice = Object.create(null);
      for (var ch = 0; ch < children.length; ch++) {
        var foundPart = null;
        for (var cp = 0; cp < parts.length; cp++) if (elementName(parts[cp]) === children[ch].localName) { foundPart = parts[cp]; break; }
        if (!foundPart) continue;
        matched = true;
        usedChoice[ch] = true;
        checkElement(children[ch], foundPart, path + '/' + children[ch].nodeName);
      }
      if (!matched && xmlAttrValue(group, 'minOccurs') !== '0') fail(path, 'expected one element from the choice');
      for (var uc = 0; uc < children.length; uc++) if (!usedChoice[uc]) fail(path, 'unexpected element "' + children[uc].nodeName + '"');
    } else {
      var cursor = 0;
      var usedSeq = Object.create(null);
      for (var sp = 0; sp < parts.length; sp++) {
        var spart = parts[sp], sname = elementName(spart);
        var smin = xmlAttrValue(spart, 'minOccurs');
        var smax = xmlAttrValue(spart, 'maxOccurs');
        var need = smin === null ? 1 : Number(smin);
        var maxAllowed = smax === null ? 1 : (smax === 'unbounded' ? Infinity : Number(smax));
        var got = 0;
        while (cursor < children.length && got < maxAllowed && children[cursor].localName === sname) {
          usedSeq[cursor] = true;
          checkElement(children[cursor], spart, path + '/' + children[cursor].nodeName);
          got++;
          cursor++;
        }
        if (got < need) fail(path, 'expected at least ' + need + ' element "' + sname + '"');
      }
      for (var us = 0; us < children.length; us++) if (!usedSeq[us]) fail(path, 'unexpected element "' + children[us].nodeName + '"');
    }
  }
  function checkElement(el, decl, path) {
    decl = resolveElement(decl);
    if (!decl) return;
    var info = typeOfElement(decl);
    var typeName = info.name;
    var typeDecl = info.decl;
    if (typeName && simpleTypes[xmlLocalRef(typeName)]) {
      checkSimpleType(directText(el), typeName, simpleTypes[xmlLocalRef(typeName)], path);
      return;
    }
    if (typeDecl && typeDecl.localName === 'simpleType') {
      checkSimpleType(directText(el), '', typeDecl, path);
      return;
    }
    if (typeName && !typeDecl) {
      checkSimpleType(directText(el), typeName, null, path);
      return;
    }
    if (!typeDecl || typeDecl.localName !== 'complexType') return;
    checkAttributes(el, typeDecl, path);
    var simpleContent = xmlChildNamed(typeDecl, 'simpleContent');
    if (simpleContent) {
      var ext = xmlChildNamed(simpleContent, 'extension') || xmlChildNamed(simpleContent, 'restriction');
      if (ext) {
        var base = xmlAttrValue(ext, 'base');
        if (base) checkSimpleType(directText(el), base, simpleTypes[xmlLocalRef(base)] || null, path);
      }
      for (var sc = 0; sc < el.childNodes.length; sc++) if (el.childNodes[sc].nodeType === 1) fail(path, 'simpleContent cannot contain element "' + el.childNodes[sc].nodeName + '"');
      return;
    }
    var groups = collectGroups(typeDecl, new Set());
    for (var g = 0; g < groups.length; g++) checkGroup(el, groups[g], path);
  }

  var rootDecl = globals[instanceRoot.localName];
  if (!rootDecl) return { errors: errors, error: 'XSD: no global element declaration for "' + instanceRoot.localName + '"' };
  checkElement(instanceRoot, rootDecl, '/' + instanceRoot.nodeName);
  return { errors: errors, error: null };
}
function validateXsd(doc, xsdText) {
  var result = validateXsdCore(doc, xsdText);
  if (result.error) return [false, result.error];
  return result.errors.length === 0 ? ok(true) : [false, result.errors[0]];
}
function validateXsdAll(doc, xsdText) {
  var result = validateXsdCore(doc, xsdText);
  if (result.error) return [[], result.error];
  return [result.errors, null];
}
function validateRules(doc, rules) {
    var rootRes = root(doc), instanceRoot = rootRes[0];
    if (!instanceRoot) return [false, 'rules: document has no root element'];
    if (!rules || typeof rules !== 'object' || Array.isArray(rules)) return [false, 'rules must be an object keyed by element name'];
    var names = Object.keys(rules);
    for (var i = 0; i < names.length; i++) {
      var name = names[i], rule = rules[name] || {};
      var matches = [];
      if (instanceRoot.localName === name) matches.push(instanceRoot);
      var all = instanceRoot.getElementsByTagName('*');
      for (var a = 0; a < all.length; a++) if (all[a].localName === name) matches.push(all[a]);
      if (rule.required === true && matches.length === 0) return [false, 'rules: required element "' + name + '" is missing'];
      if (rule.min !== undefined && matches.length < Number(rule.min)) return [false, 'rules: expected at least ' + rule.min + ' element "' + name + '"'];
      if (rule.max !== undefined && matches.length > Number(rule.max)) return [false, 'rules: expected at most ' + rule.max + ' element "' + name + '"'];
      for (var m = 0; m < matches.length; m++) {
        var el = matches[m], path = '/' + el.nodeName;
        if (Array.isArray(rule.required)) {
          for (var r = 0; r < rule.required.length; r++) if (!xmlChildNamed(el, rule.required[r])) return [false, path + ': missing required child "' + rule.required[r] + '"'];
        }
        if (Array.isArray(rule.attributes)) {
          for (var at = 0; at < rule.attributes.length; at++) if (!el.hasAttribute(rule.attributes[at])) return [false, path + ': missing required attribute "' + rule.attributes[at] + '"'];
        }
        if (rule.types && typeof rule.types === 'object') {
          var keys = Object.keys(rule.types);
          for (var k = 0; k < keys.length; k++) {
            var field = keys[k], type = rule.types[k], value;
            if (field.charAt(0) === '@') value = el.hasAttribute(field.slice(1)) ? el.getAttribute(field.slice(1)) : null;
            else { var child = xmlChildNamed(el, field); value = child ? directText(child).trim() : null; }
            if (value !== null && !xmlTypeOk(value, type)) return [false, path + ': "' + field + '" is not ' + type];
          }
        }
      }
    }
    return ok(true);
  }
  function parseFragment(doc, text) {
    var d = ownerOf(doc) || (doc && doc.nodeType === 9 ? doc : null);
    if (!d) return err('xml.parseFragment expects a document');
    if (typeof text !== 'string') return err('xml.parseFragment expects a string');
    var wrapper = parseText("<sui:fragment xmlns:sui=\"urn:sui-fragment\">" + text + "</sui:fragment>");
    if (wrapper[1] !== null) return wrapper;
    try {
      var fragment = d.createDocumentFragment();
      var rootEl = wrapper[0].documentElement;
      for (var i = 0; i < rootEl.childNodes.length; i++) fragment.appendChild(d.importNode(rootEl.childNodes[i].cloneNode(true), true));
      return ok(fragment);
    } catch (e) { return err('xml.parseFragment: ' + (e && e.message ? e.message : String(e))); }
  }
  function setInnerXml(node, text) {
    var n = nodeOf(node);
    if (!n) return err('xml.setInnerXml expects a node');
    var frag = parseFragment(ownerOf(n), text);
    if (frag[1] !== null) return frag;
    try {
      while (n.firstChild) n.removeChild(n.firstChild);
      n.appendChild(frag[0]);
      return [true, null];
    } catch (e) { return err('xml.setInnerXml: ' + (e && e.message ? e.message : String(e))); }
  }
  function replaceChildren(node, nodes) {
    var n = nodeOf(node);
    if (!n || !nodes || typeof nodes.length !== 'number') return err('xml.replaceChildren expects a node and an array');
    var prepared = [];
    try {
      for (var i = 0; i < nodes.length; i++) {
        var child = nodeOf(nodes[i]);
        if (!child) return err('xml.replaceChildren contains a non-node value');
        prepared.push(importFor(n, child));
      }
      while (n.firstChild) n.removeChild(n.firstChild);
      for (var j = 0; j < prepared.length; j++) n.appendChild(prepared[j]);
      return [true, null];
    } catch (e) { return err('xml.replaceChildren: ' + (e && e.message ? e.message : String(e))); }
  }
  function xslt(node, stylesheet) {
    var n = nodeOf(node);
    if (!n) return err('xml.xslt expects a node');
    if (typeof XSLTProcessor === 'undefined') return err('XSLT is not available in this environment');
    var parsed = parseText(String(stylesheet));
    if (parsed[1] !== null) return parsed;
    try {
      var proc = new XSLTProcessor();
      proc.importStylesheet(parsed[0]);
      var result = n.nodeType === 9 ? proc.transformToDocument(n) : proc.transformToFragment(n, ownerOf(n));
      if (!result) return err('xml.xslt produced no result');
      return ok(new XMLSerializer().serializeToString(result));
    } catch (e) { return err('xml.xslt: ' + (e && e.message ? e.message : String(e))); }
  }
  function diffOptions(options) {
    var opts = options && typeof options === 'object' ? options : {};
    var key = opts.key != null ? opts.key : (opts.match != null ? opts.match : '');
    return {
      key: typeof key === 'string' ? key : '',
      detectMoves: opts.detectMoves !== false,
      includeXml: opts.includeXml !== false,
      maxXml: Number(opts.maxXml) > 0 ? Number(opts.maxXml) : 20000
    };
  }
  function meaningfulChildren(node) {
    var out = [];
    if (!node) return out;
    for (var i = 0; i < node.childNodes.length; i++) {
      var c = node.childNodes[i];
      if (c.nodeType === 3 && String(c.nodeValue || '').trim() === '') continue;
      if (c.nodeType === 1 || c.nodeType === 3 || c.nodeType === 4 || c.nodeType === 7 || c.nodeType === 8) out.push(c);
    }
    return out;
  }
  function elementIndex(node) {
    var index = 0, p = node && node.parentNode ? node.parentNode.firstChild : null;
    while (p) {
      if (p.nodeType === 1) {
        if (p === node) return index;
        index++;
      }
      p = p.nextSibling;
    }
    return -1;
  }
  function sameNameCount(node) {
    var count = 0, p = node.parentNode ? node.parentNode.firstChild : null;
    while (p) {
      if (p.nodeType === 1 && p.nodeName === node.nodeName) count++;
      p = p.nextSibling;
    }
    return count;
  }
  function sameNameIndex(node) {
    var index = 0, p = node.parentNode ? node.parentNode.firstChild : null;
    while (p) {
      if (p.nodeType === 1 && p.nodeName === node.nodeName) {
        index++;
        if (p === node) return index;
      }
      p = p.nextSibling;
    }
    return 1;
  }
  function textNodeIndex(node) {
    var index = 0, p = node.parentNode ? node.parentNode.firstChild : null;
    while (p) {
      if (p.nodeType === node.nodeType) {
        index++;
        if (p === node) return index;
      }
      p = p.nextSibling;
    }
    return 1;
  }
  function diffSegment(node, keyName) {
    var key = keyName ? node.getAttribute(keyName) : null;
    if (key !== null && key !== '') return node.nodeName + '[@' + keyName + '="' + String(key).replace(/"/g, '&quot;') + '"]';
    var total = sameNameCount(node);
    return node.nodeName + (total > 1 ? '[' + sameNameIndex(node) + ']' : '');
  }
  function diffPath(rootNode, node, keyName) {
    if (!node) return '';
    if (node.nodeType === 9) return '/';
    var parts = [], n = node, guard = 0;
    while (n && n.nodeType !== 9 && guard++ < 1000) {
      if (n.nodeType === 1) parts.unshift(diffSegment(n, keyName));
      else if (n.nodeType === 3 || n.nodeType === 4) parts.unshift('text()[' + textNodeIndex(n) + ']');
      else if (n.nodeType === 8) parts.unshift('comment()[' + textNodeIndex(n) + ']');
      else if (n.nodeType === 7) parts.unshift('processing-instruction(' + (n.target || '') + ')');
      if (n === rootNode) break;
      n = n.parentNode;
    }
    return '/' + parts.join('/');
  }
  function diffSerialize(node, opts) {
    if (!opts.includeXml || !node) return undefined;
    try {
      var target = node.nodeType === 9 ? node.documentElement : node;
      if (!target) return undefined;
      var text = new XMLSerializer().serializeToString(target);
      if (text.length > opts.maxXml) text = text.slice(0, opts.maxXml);
      return text;
    } catch (e) { return undefined; }
  }
  function diffValue(node) {
    return node && node.nodeType === 1 ? String(directText(node)) : '';
  }
  function diff(a, b, options) {
    var x = nodeOf(a), y = nodeOf(b);
    if (!x || !y) return err('xml.diff expects two nodes');
    var opts = diffOptions(options);
    var changes = [];
    function keyOf(node) {
      if (!opts.key || !node || node.nodeType !== 1) return null;
      var value = node.getAttribute(opts.key);
      return value === null || value === '' ? null : value;
    }
    function push(change) { changes.push(change); }
    function describeNode(item, node) {
      item.kind = nodeTypeName(node);
      if (node.nodeType === 1) {
        var k = keyOf(node);
        if (k !== null) item.key = k;
      } else {
        item.value = node.nodeValue == null ? '' : String(node.nodeValue);
        if (node.nodeType === 7) item.target = node.target || '';
      }
    }
    function removedChange(node, parentPath, index) {
      var item = { op: 'removed', path: diffPath(x, node, opts.key), parent: parentPath, index: index, name: node.nodeName };
      describeNode(item, node);
      var xml = diffSerialize(node, opts);
      if (xml !== undefined) item.xml = xml;
      push(item);
    }
    function addedChange(node, parentPath, index) {
      var item = { op: 'added', path: diffPath(y, node, opts.key), parent: parentPath, index: index, name: node.nodeName };
      describeNode(item, node);
      var xml = diffSerialize(node, opts);
      if (xml !== undefined) item.xml = xml;
      push(item);
    }
    function compare(n1, n2, parentPath, index) {
      if (!n1 && !n2) return;
      if (!n1) { addedChange(n2, parentPath, index); return; }
      if (!n2) { removedChange(n1, parentPath, index); return; }
      if (n1.nodeType !== n2.nodeType || n1.nodeName !== n2.nodeName) {
        var item = { op: 'replaced', path: diffPath(x, n1, opts.key), parent: parentPath, index: index, name: n2.nodeName, oldName: n1.nodeName };
        describeNode(item, n2);
        var xml = diffSerialize(n2, opts);
        if (xml !== undefined) item.xml = xml;
        push(item);
        return;
      }
      if (n1.nodeType !== 1 && n1.nodeType !== 9) {
        var oldText = n1.nodeValue == null ? '' : String(n1.nodeValue);
        var newText = n2.nodeValue == null ? '' : String(n2.nodeValue);
        if (oldText !== newText) push({ op: 'text', path: diffPath(x, n1, opts.key), old: oldText, new: newText });
        return;
      }
      if (n1.nodeType === 1) {
        var pa = attrsOf(n1), pb = attrsOf(n2), names = {}, k;
        for (k in pa) names[k] = true;
        for (k in pb) names[k] = true;
        for (k in names) {
          var av = pa[k] === undefined ? null : pa[k];
          var bv = pb[k] === undefined ? null : pb[k];
          if (av !== bv) push({ op: 'attribute', path: diffPath(x, n1, opts.key), name: k, old: av, new: bv });
        }
      }
      compareChildren(n1, n2);
    }
    function compareChildren(n1, n2) {
      var kids1 = meaningfulChildren(n1), kids2 = meaningfulChildren(n2);
      var parentPath = diffPath(y, n2, opts.key);
      if (!opts.key) {
        var max = Math.max(kids1.length, kids2.length);
        for (var i = 0; i < max; i++) compare(kids1[i], kids2[i], parentPath, i);
        return;
      }
      var elements1 = [], elements2 = [], others1 = [], others2 = [];
      for (var a = 0; a < kids1.length; a++) { if (kids1[a].nodeType === 1) elements1.push(kids1[a]); else others1.push(kids1[a]); }
      for (var b2 = 0; b2 < kids2.length; b2++) { if (kids2[b2].nodeType === 1) elements2.push(kids2[b2]); else others2.push(kids2[b2]); }
      var byKey = Object.create(null), unkeyed2 = [], e;
      for (e = 0; e < elements2.length; e++) {
        var k2 = keyOf(elements2[e]);
        if (k2 !== null) {
          if (!byKey[k2]) byKey[k2] = [];
          byKey[k2].push(elements2[e]);
        } else unkeyed2.push(elements2[e]);
      }
      var pairs = [], unkeyed1 = [], used2 = [];
      for (e = 0; e < elements1.length; e++) {
        var k1 = keyOf(elements1[e]);
        if (k1 !== null && byKey[k1] && byKey[k1].length) {
          var match = byKey[k1].shift();
          pairs.push({ src: elements1[e], dst: match, key: k1 });
          used2.push(match);
        } else unkeyed1.push(elements1[e]);
      }
      var umax = Math.max(unkeyed1.length, unkeyed2.length);
      for (var u = 0; u < umax; u++) {
        if (unkeyed2[u]) used2.push(unkeyed2[u]);
        pairs.push({ src: unkeyed1[u], dst: unkeyed2[u], key: null });
      }
      if (opts.detectMoves) {
        var matched = [];
        for (var m = 0; m < pairs.length; m++) if (pairs[m].key !== null && pairs[m].src && pairs[m].dst) matched.push(pairs[m]);
        var srcOrder = [], dstOrder = [];
        for (var s = 0; s < matched.length; s++) srcOrder.push(matched[s].key);
        var sorted = matched.slice().sort(function (p, q) { return elementIndex(p.dst) - elementIndex(q.dst); });
        for (var d = 0; d < sorted.length; d++) dstOrder.push(sorted[d].key);
        if (srcOrder.join('\u0000') !== dstOrder.join('\u0000')) {
          var srcPos = Object.create(null), dstPos = Object.create(null), q;
          for (q = 0; q < srcOrder.length; q++) srcPos[srcOrder[q]] = q;
          for (q = 0; q < dstOrder.length; q++) dstPos[dstOrder[q]] = q;
          for (var mo = 0; mo < matched.length; mo++) {
            var pair = matched[mo];
            if (srcPos[pair.key] === dstPos[pair.key]) continue;
            push({
              op: 'moved',
              path: diffPath(y, pair.dst, opts.key),
              from: diffPath(x, pair.src, opts.key),
              parent: parentPath,
              name: pair.dst.nodeName,
              key: pair.key,
              fromIndex: elementIndex(pair.src),
              to: elementIndex(pair.dst)
            });
          }
        }
      }
      var usedSrc = [];
      for (var p = 0; p < pairs.length; p++) {
        if (pairs[p].src) usedSrc.push(pairs[p].src);
        if (pairs[p].src && pairs[p].dst) {
          var targetIndex = pairs[p].dst.parentNode ? Array.prototype.indexOf.call(pairs[p].dst.parentNode.childNodes, pairs[p].dst) : p;
          compare(pairs[p].src, pairs[p].dst, parentPath, targetIndex);
        } else if (pairs[p].src) {
          removedChange(pairs[p].src, parentPath, pairs[p].src.parentNode ? Array.prototype.indexOf.call(pairs[p].src.parentNode.childNodes, pairs[p].src) : p);
        } else if (pairs[p].dst) {
          addedChange(pairs[p].dst, parentPath, pairs[p].dst.parentNode ? Array.prototype.indexOf.call(pairs[p].dst.parentNode.childNodes, pairs[p].dst) : p);
        }
      }
      for (var extra = 0; extra < elements2.length; extra++) {
        if (used2.indexOf(elements2[extra]) < 0) addedChange(elements2[extra], parentPath, elementIndex(elements2[extra]));
      }
      var omax = Math.max(others1.length, others2.length);
      for (var o = 0; o < omax; o++) {
        var oi = others1[o] && others1[o].parentNode ? Array.prototype.indexOf.call(others1[o].parentNode.childNodes, others1[o]) : o;
        compare(others1[o], others2[o], parentPath, oi);
      }
    }
    try {
      compare(x, y, x.nodeType === 9 ? '/' : diffPath(x, x, opts.key), 0);
      return ok(changes);
    } catch (e) {
      return err('xml.diff: ' + (e && e.message ? e.message : String(e)));
    }
  }
  function parsePath(path) {
    var segments = [], text = String(path == null ? '' : path), i = 0;
    while (i < text.length) {
      if (text.charAt(i) !== '/') { i++; continue; }
      i++;
      var start = i;
      while (i < text.length && text.charAt(i) !== '/' && text.charAt(i) !== '[') i++;
      var name = text.slice(start, i), pred = null;
      if (text.charAt(i) === '[') {
        var close = text.indexOf(']', i);
        if (close < 0) break;
        pred = text.slice(i + 1, close);
        i = close + 1;
      }
      if (name) segments.push({ name: name, pred: pred });
    }
    return segments;
  }
  function pathSegmentMatches(node, name) {
    if (name === 'text()') return node.nodeType === 3 || node.nodeType === 4;
    if (name === 'comment()') return node.nodeType === 8;
    if (name.indexOf('processing-instruction(') === 0) return node.nodeType === 7 && ('processing-instruction(' + (node.target || '') + ')') === name;
    if (node.nodeType !== 1) return false;
    return node.nodeName === name || node.localName === name;
  }
  function lookupPath(rootNode, path) {
    var start = nodeOf(rootNode), segments = parsePath(path);
    if (!start) return null;
    var current = start, first = 0;
    if (current.nodeType === 1 && segments.length && pathSegmentMatches(current, segments[0].name) && !segments[0].pred) first = 1;
    for (var i = first; i < segments.length; i++) {
      var segment = segments[i], pred = segment.pred, keyName = null, keyValue = null, nth = 0;
      if (pred) {
        var keyMatch = /^@([^=]+)="([\s\S]*)"$/.exec(pred);
        if (keyMatch) {
          keyName = keyMatch[1];
          keyValue = keyMatch[2].replace(/&quot;/g, '"');
        } else if (/^\d+$/.test(pred)) nth = Number(pred);
      }
      var found = null, count = 0;
      if (current.nodeType !== 1 && current.nodeType !== 9 && current.nodeType !== 11) return null;
      for (var c = 0; c < current.childNodes.length; c++) {
        var child = current.childNodes[c];
        if (!pathSegmentMatches(child, segment.name)) continue;
        if (keyName !== null) {
          if (child.nodeType === 1 && child.getAttribute(keyName) === keyValue) { found = child; break; }
          continue;
        }
        count++;
        if (!nth || count === nth) { found = child; break; }
      }
      if (!found) return null;
      current = found;
    }
    return current;
  }
  function insertElementAt(parentNode, node, index) {
    var before = null, count = 0;
    for (var i = 0; i < parentNode.childNodes.length; i++) {
      var child = parentNode.childNodes[i];
      if (child.nodeType !== 1) continue;
      if (count === index) { before = child; break; }
      count++;
    }
    parentNode.insertBefore(node, before);
  }
  function insertNodeAt(parentNode, node, index) {
    var before = null, count = 0;
    for (var i = 0; i < parentNode.childNodes.length; i++) {
      var child = parentNode.childNodes[i];
      if (count === index) { before = child; break; }
      count++;
    }
    parentNode.insertBefore(node, before);
  }
  function depthOf(node) {
    var depth = 0, n = node;
    while (n && n.parentNode) { depth++; n = n.parentNode; }
    return depth;
  }
  function applyNodeChange(node, change) {
    if (change.op === 'attribute') {
      if (change.new === null || change.new === undefined) node.removeAttribute(String(change.name));
      else node.setAttribute(String(change.name), String(change.new));
      return;
    }
    if (change.op === 'text') {
      node.nodeValue = String(change.new == null ? '' : change.new);
      return;
    }
    if (change.op === 'replaced') {
      if (change.kind && change.kind !== 'element') {
        if (change.kind === 'text' || change.kind === 'cdata' || change.kind === 'comment' || change.kind === 'pi') {
          node.nodeValue = String(change.value == null ? '' : change.value);
          return;
        }
        throw new Error('replaced change has no element XML');
      }
      var owner = ownerOf(node);
      var fragment = parseFragment(owner, String(change.xml || ''));
      if (fragment[1] !== null) throw new Error(fragment[1]);
      var replacement = null;
      for (var i = 0; i < fragment[0].childNodes.length; i++) {
        if (fragment[0].childNodes[i].nodeType === 1) { replacement = fragment[0].childNodes[i]; break; }
      }
      if (!replacement) throw new Error('replaced change has no element XML');
      node.parentNode.replaceChild(replacement, node);
      return;
    }
    throw new Error('unsupported change op "' + change.op + '"');
  }
  function patch(target, changes) {
    var root = nodeOf(target);
    if (!root) return err('xml.patch expects a node');
    if (!changes || typeof changes.length !== 'number') return err('xml.patch expects an array of changes');
    var nodeOps = [], removals = [], moves = [], additions = [], issues = [];
    function issue(message) { if (issues.length < 20) issues.push(message); }
    for (var i = 0; i < changes.length; i++) {
      var change = changes[i];
      if (!change || typeof change !== 'object' || typeof change.op !== 'string') { issue('change ' + (i + 1) + ' is not an object with an op'); continue; }
      try {
        if (change.op === 'attribute' || change.op === 'text' || change.op === 'replaced') {
          var node = lookupPath(root, change.path);
          if (!node) { issue('missing node for ' + change.op + ' at ' + change.path); continue; }
          nodeOps.push({ change: change, node: node });
        } else if (change.op === 'removed') {
          var gone = lookupPath(root, change.from || change.path);
          if (gone) removals.push({ change: change, node: gone });
          else issue('missing node for removed at ' + (change.from || change.path));
        } else if (change.op === 'moved') {
          var moved = lookupPath(root, change.from || change.path);
          var moveParent = lookupPath(root, change.parent);
          if (moved && moveParent) moves.push({ change: change, node: moved, parent: moveParent });
          else issue('missing node for moved at ' + (change.from || change.path));
        } else if (change.op === 'added') {
          var parent = lookupPath(root, change.parent);
          if (parent) additions.push({ change: change, parent: parent });
          else issue('missing parent for added at ' + change.parent);
        } else issue('unknown op "' + change.op + '"');
      } catch (e) {
        issue('invalid change ' + (i + 1) + ': ' + (e && e.message ? e.message : String(e)));
      }
    }
    try {
      for (var n = 0; n < nodeOps.length; n++) applyNodeChange(nodeOps[n].node, nodeOps[n].change);
      removals.sort(function (p, q) { return depthOf(q.node) - depthOf(p.node); });
      for (var r = 0; r < removals.length; r++) if (removals[r].node.parentNode) removals[r].node.parentNode.removeChild(removals[r].node);
      var groups = [];
      for (var m = 0; m < moves.length; m++) {
        var group = null;
        for (var g = 0; g < groups.length; g++) if (groups[g].parent === moves[m].parent) { group = groups[g]; break; }
        if (!group) { group = { parent: moves[m].parent, items: [] }; groups.push(group); }
        group.items.push(moves[m]);
      }
      for (var gi = 0; gi < groups.length; gi++) {
        var items = groups[gi].items;
        items.sort(function (p, q) { return (Number(p.change.to) || 0) - (Number(q.change.to) || 0); });
        for (var mi = 0; mi < items.length; mi++) if (items[mi].node.parentNode) items[mi].node.parentNode.removeChild(items[mi].node);
        for (var mj = 0; mj < items.length; mj++) insertElementAt(groups[gi].parent, items[mj].node, Number(items[mj].change.to) || 0);
      }
      var addGroups = [];
      for (var a = 0; a < additions.length; a++) {
        var addGroup = null;
        for (var ag = 0; ag < addGroups.length; ag++) if (addGroups[ag].parent === additions[a].parent) { addGroup = addGroups[ag]; break; }
        if (!addGroup) { addGroup = { parent: additions[a].parent, items: [] }; addGroups.push(addGroup); }
        addGroup.items.push(additions[a]);
      }
      for (var ai = 0; ai < addGroups.length; ai++) {
        var addItems = addGroups[ai].items;
        addItems.sort(function (p, q) { return (Number(p.change.index) || 0) - (Number(q.change.index) || 0); });
        for (var aj = 0; aj < addItems.length; aj++) {
          var addChange = addItems[aj].change;
          var addParent = addGroups[ai].parent;
          var addDoc = ownerOf(addParent);
          var addedNode = null;
          if (addChange.kind && addChange.kind !== 'element') {
            if (addChange.kind === 'text') addedNode = addDoc.createTextNode(String(addChange.value == null ? '' : addChange.value));
            else if (addChange.kind === 'cdata') addedNode = addDoc.createCDATASection(String(addChange.value == null ? '' : addChange.value));
            else if (addChange.kind === 'comment') addedNode = addDoc.createComment(String(addChange.value == null ? '' : addChange.value));
            else if (addChange.kind === 'pi') addedNode = addDoc.createProcessingInstruction(String(addChange.target || ''), String(addChange.value || ''));
          } else if (addChange.xml) {
            var parsed = parseFragment(addDoc, String(addChange.xml));
            if (parsed[1] !== null) throw new Error(parsed[1]);
            for (var ac = 0; ac < parsed[0].childNodes.length; ac++) {
              if (parsed[0].childNodes[ac].nodeType === 1) { addedNode = parsed[0].childNodes[ac]; break; }
            }
          } else if (addChange.name) addedNode = addDoc.createElement(String(addChange.name));
          if (!addedNode) throw new Error('added change has no node XML');
          if (addedNode.nodeType === 1) insertElementAt(addParent, addedNode, Number(addChange.index) || 0);
          else insertNodeAt(addParent, addedNode, Number(addChange.index) || 0);
        }
      }
      return issues.length ? [false, issues.join('; ').slice(0, 500)] : [true, null];
    } catch (e) {
      return err('xml.patch: ' + (e && e.message ? e.message : String(e)));
    }
  }
  function merge(target, source, options) {
    var t = nodeOf(target), s = nodeOf(source);
    if (!t || !s) return err('xml.merge expects two nodes');
    var opts = options && typeof options === 'object' ? options : {};
    var mode = opts.mode ? String(opts.mode) : 'append';
    var src = s.nodeType === 9 ? s.documentElement : s;
    if (!src) return err('xml.merge source is empty');
    var count = 0;
    try {
      if (mode !== 'update') {
        var kids = Array.prototype.slice.call(src.childNodes);
        for (var i = 0; i < kids.length; i++) { t.appendChild(importFor(t, kids[i].cloneNode(true))); count++; }
        return ok(count);
      }
      var keyName = opts.key != null ? String(opts.key) : '';
      var targetEl = elementOf(t);
      if (!targetEl) return err('xml.merge target has no element');
      function keyOf(node) {
        if (!keyName || !node || node.nodeType !== 1) return null;
        var value = node.getAttribute(keyName);
        return value === null || value === '' ? null : value;
      }
      function mergeElement(dst, srcEl) {
        for (var a = 0; a < srcEl.attributes.length; a++) {
          var attr = srcEl.attributes[a];
          if (attr.name === 'xmlns' || attr.name.indexOf('xmlns:') === 0) continue;
          if (dst.getAttribute(attr.name) !== attr.value) { dst.setAttribute(attr.name, attr.value); count++; }
        }
        var srcKids = [];
        for (var c = 0; c < srcEl.childNodes.length; c++) if (srcEl.childNodes[c].nodeType === 1) srcKids.push(srcEl.childNodes[c]);
        if (!srcKids.length) {
          var value = String(directText(srcEl));
          if (String(directText(dst)) !== value) {
            while (dst.firstChild) dst.removeChild(dst.firstChild);
            if (value !== '') dst.appendChild(ownerOf(dst).createTextNode(value));
            count++;
          }
          return;
        }
        var existing = Object.create(null), unkeyed = [], d;
        for (d = 0; d < dst.childNodes.length; d++) {
          var child = dst.childNodes[d];
          if (child.nodeType !== 1) continue;
          var childKey = keyOf(child);
          if (childKey !== null) (existing[childKey] || (existing[childKey] = [])).push(child);
          else unkeyed.push(child);
        }
        for (var sk = 0; sk < srcKids.length; sk++) {
          var sourceChild = srcKids[sk], sourceKey = keyOf(sourceChild);
          var match = null, picked = -1;
          if (sourceKey !== null && existing[sourceKey] && existing[sourceKey].length) match = existing[sourceKey].shift();
          else {
            for (var u = 0; u < unkeyed.length; u++) {
              if (unkeyed[u].nodeName === sourceChild.nodeName) { match = unkeyed[u]; picked = u; break; }
            }
            if (picked >= 0) unkeyed.splice(picked, 1);
          }
          if (match) mergeElement(match, sourceChild);
          else { dst.appendChild(importFor(dst, sourceChild.cloneNode(true))); count++; }
        }
      }
      mergeElement(targetEl, src);
      return ok(count);
    } catch (e) { return err('xml.merge: ' + (e && e.message ? e.message : String(e))); }
  }
  function merge3(base, ours, theirs, options) {
    var b = nodeOf(base), o = nodeOf(ours), t = nodeOf(theirs);
    if (!b || !o || !t) return err('xml.merge3 expects three nodes');
    var opts = diffOptions(options);
    var changes = diff(b, t, options);
    if (changes[1] !== null) return [[], changes[1]];
    var conflicts = [], safe = [];
    function describe(change) { return change.op + ' at ' + (change.path || change.from || '/'); }
    function conflictChange(change) {
      if (change.op === 'attribute') {
        var mine = lookupPath(o, change.path);
        if (!mine || mine.nodeType !== 1) return false;
        var mineValue = mine.hasAttribute(change.name) ? mine.getAttribute(change.name) : null;
        if (mineValue === change.new) return false;
        if (mineValue === change.old) return false;
        return true;
      }
      if (change.op === 'text') {
        var mineText = lookupPath(o, change.path);
        if (!mineText) return false;
        var textValue = mineText.nodeValue == null ? '' : String(mineText.nodeValue);
        if (textValue === change.new || textValue === change.old) return false;
        return true;
      }
      if (change.op === 'added' || change.op === 'replaced') {
        var mineNode = lookupPath(o, change.path);
        if (!mineNode) return false;
        if (change.xml === undefined) return false;
        try { return new XMLSerializer().serializeToString(mineNode) !== change.xml; } catch (e) { return true; }
      }
      if (change.op === 'removed') {
        var still = lookupPath(o, change.path);
        if (!still) return false;
        var original = lookupPath(b, change.path);
        if (!original) return false;
        try { return new XMLSerializer().serializeToString(still) !== new XMLSerializer().serializeToString(original); } catch (e) { return true; }
      }
      if (change.op === 'moved') {
        var mineMoved = lookupPath(o, change.from);
        if (!mineMoved) return true;
        var mineIndex = elementIndex(mineMoved);
        if (mineIndex === change.to || mineIndex === change.fromIndex) return false;
        return true;
      }
      return false;
    }
    for (var i = 0; i < changes[0].length; i++) {
      var change = changes[0][i];
      if (conflictChange(change)) conflicts.push(describe(change));
      else safe.push(change);
    }
    var applied = patch(o, safe);
    if (applied[1] !== null) return [[], applied[1]];
    return ok(conflicts);
  }
  return {
    parse: parseText,
    parseBytes: parseBytes,
    validate: function (text) { var r = parseText(text); return r[1] === null ? ok(true) : [false, r[1]]; },
    root: root, parent: parent, children: children, childNodes: childNodes, childCount: childCount,
    next: function (n) { return sibling(n, true); }, previous: function (n) { return sibling(n, false); },
    find: find, findNs: findNs, findAll: findAll, descendants: descendants, has: has,
    type: function (n) { var v = nodeOf(n); return ok(v ? nodeTypeName(v) : ''); },
    name: name, localName: localName, namespace: namespace, prefix: prefix,
    text: text, content: content, attr: attr, attrNs: attrNs, attrs: attrs, path: path,
    createDocument: createDocument, createElement: createElement, createElementNs: createElementNs,
    createText: createText, createComment: createComment, createCdata: createCdata, createPi: createPi,
    parseFragment: parseFragment, setInnerXml: setInnerXml, replaceChildren: replaceChildren,
    append: append, prepend: prepend, insertBefore: insertBefore, insertAfter: insertAfter,
    remove: remove, replace: replace, clear: clear, clone: clone, rename: rename, setText: setText,
    move: move, reorder: reorder,
    setAttr: setAttr, setAttrNs: setAttrNs, removeAttr: removeAttr, renameAttr: renameAttr, hasAttr: hasAttr,
    xpath: xpath, xpathOne: xpathOne, findAttr: findAttr, findText: findText,
    stringify: stringify, toBytes: toBytes, toJson: toJson, fromJson: fromJson, toCsv: toCsv,
    xslt: xslt, merge: merge, merge3: merge3, diff: diff, patch: patch,
    validateXsd: validateXsd, validateXsdAll: validateXsdAll, validateRules: validateRules
  };
})();
`
