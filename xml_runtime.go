package main

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
      else label = 'utf-8';
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
      if (Array.isArray(value)) {
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
  function validateXsd(doc, xsdText) {
    var rootRes = root(doc);
    var instanceRoot = rootRes[0];
    if (!instanceRoot) return [false, 'XSD: document has no root element'];
    var parsed = parseText(String(xsdText));
    if (parsed[1] !== null) return [false, 'XSD: ' + parsed[1]];
    var schema = parsed[0].documentElement;
    if (!schema || schema.localName !== 'schema') return [false, 'XSD: root element must be xs:schema'];
    var targetNS = xmlAttrValue(schema, 'targetNamespace');
    if (targetNS && instanceRoot.namespaceURI !== targetNS) return [false, 'XSD: root namespace does not match targetNamespace'];
    var globals = Object.create(null), complexTypes = Object.create(null), simpleTypes = Object.create(null);
    var schemaChildren = [];
    for (var si = 0; si < schema.childNodes.length; si++) if (schema.childNodes[si].nodeType === 1) schemaChildren.push(schema.childNodes[si]);
    for (var gi = 0; gi < schemaChildren.length; gi++) {
      var g = schemaChildren[gi];
      if (g.localName === 'element' && xmlAttrValue(g, 'name')) globals[xmlAttrValue(g, 'name')] = g;
      else if (g.localName === 'complexType' && xmlAttrValue(g, 'name')) complexTypes[xmlAttrValue(g, 'name')] = g;
      else if (g.localName === 'simpleType' && xmlAttrValue(g, 'name')) simpleTypes[xmlAttrValue(g, 'name')] = g;
    }
    var rootDecl = globals[instanceRoot.localName];
    if (!rootDecl) return [false, 'XSD: no global element declaration for "' + instanceRoot.localName + '"'];
    function simpleError(value, decl, path) {
      var typeName = xmlAttrValue(decl, 'type');
      var inline = xmlChildNamed(decl, 'simpleType');
      var typeDecl = inline;
      if (typeName && simpleTypes[xmlLocalRef(typeName)]) typeDecl = simpleTypes[xmlLocalRef(typeName)];
      if (!inline && !typeName) return null;
      var source = typeName ? xmlLocalRef(typeName) : '';
      if (typeDecl && typeDecl.localName === 'simpleType') {
        var restriction = xmlChildNamed(typeDecl, 'restriction');
        if (restriction) {
          var base = xmlAttrValue(restriction, 'base');
          if (base && !xmlTypeOk(value, base)) return path + ': value "' + value + '" is not ' + xmlLocalRef(base);
          var enums = xmlChildrenNamed(restriction, 'enumeration');
          if (enums.length) {
            var found = false;
            for (var e = 0; e < enums.length; e++) if (xmlAttrValue(enums[e], 'value') === value) found = true;
            if (!found) return path + ': value "' + value + '" is not in the allowed enumeration';
          }
          var minLen = xmlAttrValue(restriction, 'minLength');
          if (minLen !== null && value.length < Number(minLen)) return path + ': value is shorter than minLength ' + minLen;
          var maxLen = xmlAttrValue(restriction, 'maxLength');
          if (maxLen !== null && value.length > Number(maxLen)) return path + ': value is longer than maxLength ' + maxLen;
          var pattern = xmlAttrValue(restriction, 'pattern');
          if (pattern) { try { if (!(new RegExp(pattern)).test(value)) return path + ': value does not match the pattern'; } catch (err) { return path + ': invalid pattern'; } }
          return null;
        }
      }
      if (source && !xmlTypeOk(value, source)) return path + ': value "' + value + '" is not ' + source;
      return null;
    }
    function validateAttributes(el, decl, path) {
      var attrs = xmlChildrenNamed(decl, 'attribute');
      var any = xmlChildNamed(decl, 'anyAttribute') !== null;
      var declared = Object.create(null);
      for (var i = 0; i < attrs.length; i++) {
        var a = attrs[i], name = xmlAttrValue(a, 'name');
        if (!name) {
          var ref = xmlAttrValue(a, 'ref');
          name = ref ? xmlLocalRef(ref) : '';
        }
        if (!name) continue;
        declared[name] = true;
        var value = el.hasAttribute(name) ? el.getAttribute(name) : null;
        if (value === null) {
          if (xmlAttrValue(a, 'use') === 'required') return path + ': missing required attribute "' + name + '"';
          continue;
        }
        var type = xmlAttrValue(a, 'type');
        if (type && !xmlTypeOk(value, type)) return path + ': attribute "' + name + '" is not ' + xmlLocalRef(type);
      }
      if (!any) {
        for (var k = 0; k < el.attributes.length; k++) {
          var an = el.attributes[k].name;
          if (an === 'xmlns' || an.indexOf('xmlns:') === 0) continue;
          if (!declared[an]) return path + ': unexpected attribute "' + an + '"';
        }
      }
      return null;
    }
    function declarationFor(decl) {
      var typeName = xmlAttrValue(decl, 'type');
      if (typeName) {
        var local = xmlLocalRef(typeName);
        if (complexTypes[local]) return complexTypes[local];
        if (simpleTypes[local]) return simpleTypes[local];
      }
      var inlineComplex = xmlChildNamed(decl, 'complexType');
      if (inlineComplex) return inlineComplex;
      var inlineSimple = xmlChildNamed(decl, 'simpleType');
      if (inlineSimple) return inlineSimple;
      return null;
    }
    function validateElement(el, decl, path) {
      if (!decl) return null;
      var typeDecl = declarationFor(decl);
      var typeName = xmlAttrValue(decl, 'type');
      if (typeDecl && typeDecl.localName === 'simpleType') return simpleError(directText(el), decl, path);
      if (typeName && !typeDecl && !xmlTypeOk(directText(el), typeName)) return path + ': value "' + directText(el) + '" is not ' + xmlLocalRef(typeName);
      if (!typeDecl || typeDecl.localName !== 'complexType') return null;
      var attrError = validateAttributes(el, typeDecl, path);
      if (attrError) return attrError;
      var group = xmlChildNamed(typeDecl, 'sequence') || xmlChildNamed(typeDecl, 'all') || xmlChildNamed(typeDecl, 'choice');
      if (!group) return null;
      var parts = xmlChildrenNamed(group, 'element');
      if (!parts.length) return null;
      var children = [];
      for (var c = 0; c < el.childNodes.length; c++) if (el.childNodes[c].nodeType === 1) children.push(el.childNodes[c]);
      var used = {};
      function nameOf(part) {
        var n = xmlAttrValue(part, 'name');
        if (n) return n;
        var r = xmlAttrValue(part, 'ref');
        return r ? xmlLocalRef(r) : '';
      }
      function partDecl(part) {
        var r = xmlAttrValue(part, 'ref');
        if (r && globals[xmlLocalRef(r)]) return globals[xmlLocalRef(r)];
        return part;
      }
      if (group.localName === 'all') {
        for (var p = 0; p < parts.length; p++) {
          var name = nameOf(parts[p]);
          var min = xmlAttrValue(parts[p], 'minOccurs');
          var matched = 0;
          for (var ci = 0; ci < children.length; ci++) if (children[ci].localName === name) { matched++; used[ci] = true; var e = validateElement(children[ci], partDecl(parts[p]), path + '/' + children[ci].nodeName); if (e) return e; }
          if (min !== null && matched < Number(min)) return path + ': expected at least ' + min + ' element "' + name + '"';
        }
      } else if (group.localName === 'choice') {
        var choices = 0;
        for (var q = 0; q < parts.length; q++) {
          var qname = nameOf(parts[q]);
          for (var qi = 0; qi < children.length; qi++) {
            if (children[qi].localName === qname && !used[qi]) {
              used[qi] = true; choices++;
              var ce = validateElement(children[qi], partDecl(parts[q]), path + '/' + children[qi].nodeName);
              if (ce) return ce;
            }
          }
        }
        if (choices === 0 && parts.length) return path + ': expected one element from the choice';
      } else {
        var cursor = 0;
        for (var pi = 0; pi < parts.length; pi++) {
          var part = parts[pi], pname = nameOf(part);
          var minOccurs = xmlAttrValue(part, 'minOccurs');
          var maxOccurs = xmlAttrValue(part, 'maxOccurs');
          var need = minOccurs === null ? 1 : Number(minOccurs);
          var max = maxOccurs === null ? 1 : (maxOccurs === 'unbounded' ? Infinity : Number(maxOccurs));
          var count = 0;
          while (cursor < children.length && count < max && children[cursor].localName === pname) {
            used[cursor] = true; count++;
            var pe = validateElement(children[cursor], partDecl(part), path + '/' + children[cursor].nodeName);
            if (pe) return pe;
            cursor++;
          }
          if (count < need) return path + ': expected at least ' + need + ' element "' + pname + '"';
        }
      }
      for (var ui = 0; ui < children.length; ui++) if (!used[ui]) return path + ': unexpected element "' + children[ui].nodeName + '"';
      return null;
    }
    var xsdError = validateElement(instanceRoot, rootDecl, '/' + instanceRoot.nodeName);
    return xsdError === null ? ok(true) : [false, xsdError];
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
  function merge(target, source) {
    var t = nodeOf(target), s = nodeOf(source);
    if (!t || !s) return err('xml.merge expects two nodes');
    var src = s.nodeType === 9 ? s.documentElement : s;
    if (!src) return err('xml.merge source is empty');
    var count = 0;
    try {
      var kids = Array.prototype.slice.call(src.childNodes);
      for (var i = 0; i < kids.length; i++) { t.appendChild(importFor(t, kids[i].cloneNode(true))); count++; }
      return ok(count);
    } catch (e) { return err('xml.merge: ' + (e && e.message ? e.message : String(e))); }
  }
  function diff(a, b) {
    var x = nodeOf(a), y = nodeOf(b);
    if (!x || !y) return err('xml.diff expects two nodes');
    var changes = [];
    function add(op, path, extra) {
      var item = { op: op, path: path };
      if (extra) for (var k in extra) item[k] = extra[k];
      changes.push(item);
    }
    function compare(n1, n2, path) {
      if (!n1 && !n2) return;
      if (!n1) { add('added', path, { name: n2.nodeName }); return; }
      if (!n2) { add('removed', path, { name: n1.nodeName }); return; }
      if (n1.nodeType !== n2.nodeType || n1.nodeName !== n2.nodeName) {
        add('replaced', path, { old: n1.nodeName, new: n2.nodeName }); return;
      }
      if (n1.nodeType === 3 || n1.nodeType === 4 || n1.nodeType === 8 || n1.nodeType === 7) {
        if (String(n1.nodeValue || '') !== String(n2.nodeValue || '')) add('text', path, { old: n1.nodeValue || '', new: n2.nodeValue || '' });
        return;
      }
      if (n1.nodeType === 1) {
        var aa = attrsOf(n1), bb = attrsOf(n2), keys = {};
        for (var ka in aa) keys[ka] = true;
        for (var kb in bb) keys[kb] = true;
        for (var k in keys) {
          if (aa[k] !== bb[k]) add('attribute', path, { name: k, old: aa[k] === undefined ? null : aa[k], new: bb[k] === undefined ? null : bb[k] });
        }
      }
      var c1 = [], c2 = [];
      for (var i = 0; i < n1.childNodes.length; i++) if (n1.childNodes[i].nodeType !== 3 || String(n1.childNodes[i].nodeValue || '').trim() !== '') c1.push(n1.childNodes[i]);
      for (var j = 0; j < n2.childNodes.length; j++) if (n2.childNodes[j].nodeType !== 3 || String(n2.childNodes[j].nodeValue || '').trim() !== '') c2.push(n2.childNodes[j]);
      var max = Math.max(c1.length, c2.length);
      for (var c = 0; c < max; c++) compare(c1[c], c2[c], path + '/' + ((c1[c] && c1[c].nodeName) || (c2[c] && c2[c].nodeName) || (c + 1)));
    }
    try { compare(x, y, x.nodeType === 9 ? '' : '/' + x.nodeName); return ok(changes); }
    catch (e) { return err('xml.diff: ' + (e && e.message ? e.message : String(e))); }
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
    append: append, prepend: prepend, insertBefore: insertBefore, insertAfter: insertAfter,
    remove: remove, replace: replace, clear: clear, clone: clone, rename: rename, setText: setText,
    move: move, reorder: reorder,
    setAttr: setAttr, setAttrNs: setAttrNs, removeAttr: removeAttr, renameAttr: renameAttr, hasAttr: hasAttr,
    xpath: xpath, xpathOne: xpathOne, findAttr: findAttr, findText: findText,
    stringify: stringify, toBytes: toBytes, toJson: toJson, fromJson: fromJson, toCsv: toCsv,
    xslt: xslt, merge: merge, diff: diff,
    validateXsd: validateXsd, validateRules: validateRules
  };
})();
`
