package main

// richtext_runtime.go - the rich text editor runtime (Core, Rendering, UI,
// Plugins and Serialization). The runtime is emitted with every web/desktop
// app; the tag is inert outside `richtext`.
//
// Security: the editor never inserts imported HTML directly. HTML/XML/Markdown
// are parsed into the JSON document model and rebuilt with DOM APIs; URLs are
// restricted to http(s)/mailto/tel/#, and images to http(s)/blob/data:image.
// Scripts, styles, event attributes and unknown elements are discarded.
const suiRichTextRuntime = `/* SimpleUI rich text editor runtime (Core + Rendering + UI + Plugins + Serialization). */
window.__SUI_RICH = (function () {
  'use strict';
  var instances = Object.create(null);
  var commands = Object.create(null);
  var plugins = Object.create(null);
  var pluginOrder = [];
  var MAX_HISTORY = 120;
  var MAX_NODES = 10000;
  var DEFAULT_DOC = { type: 'doc', children: [{ type: 'paragraph', children: [] }] };
  var defaultToolbar = 'undo,redo,copy,cut,paste|styleSelect,fontFamily,fontSize,bold,italic,underline,strike,textColor,highlight,superscript,subscript,clearFormat|alignLeft,alignCenter,alignRight,justify,indent,outdent,lineHeight,spaceBefore,spaceAfter,bulletList,orderedList,blockquote,codeBlock|link,image,table,symbol,hr';
  var ICONS = {
    undo: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M9 7H5V3M5 7c3-4 10-4 13 1s1 11-4 13"/></svg>',
    redo: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 7h4V3m0 4c-3-4-10-4-13 1s-1 11 4 13"/></svg>',
    copy: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="8" y="8" width="11" height="12" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h2"/></svg>',
    cut: '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="6" cy="18" r="2"/><circle cx="18" cy="18" r="2"/><path d="M8 16 18 4M16 16 6 4"/></svg>',
    paste: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M9 5h6v3H9zM8 5H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2"/><path d="M8 12h8M8 16h6"/></svg>',
    bold: '<b aria-hidden="true">B</b>',
    italic: '<i aria-hidden="true">I</i>',
    underline: '<u aria-hidden="true">U</u>',
    strike: '<s aria-hidden="true">S</s>',
    superscript: '<span aria-hidden="true">x<sup>2</sup></span>',
    subscript: '<span aria-hidden="true">x<sub>2</sub></span>',
    clearFormat: '<span aria-hidden="true">Tx</span>',
    alignLeft: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M4 10h11M4 14h16M4 18h9"/></svg>',
    alignCenter: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M7 10h10M4 14h16M8 18h8"/></svg>',
    alignRight: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M9 10h11M4 14h16M11 18h9"/></svg>',
    justify: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M4 10h16M4 14h16M4 18h16"/></svg>',
    indent: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M10 10h10M10 14h10M4 18h16M4 10l3 2-3 2"/></svg>',
    outdent: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M10 10h10M10 14h10M4 18h16M7 10l-3 2 3 2"/></svg>',
    bulletList: '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="5" cy="7" r="1.4"/><circle cx="5" cy="12" r="1.4"/><circle cx="5" cy="17" r="1.4"/><path d="M9 7h11M9 12h11M9 17h11"/></svg>',
    orderedList: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h2v2H4zM4 11h2v2H4zM4 16h2v2H4zM9 7h11M9 12h11M9 17h11"/></svg>',
    link: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M10 13a5 5 0 0 0 7 0l2-2a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-2 2a5 5 0 0 0 7 7l1-1"/></svg>',
    unlink: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 4l16 16M10 13a5 5 0 0 0 7 0M14 11a5 5 0 0 0-7 0"/></svg>',
    image: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="8" cy="9" r="1.6"/><path d="M4 17l5-5 4 4 3-3 4 4"/></svg>',
    table: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M3 9h18M9 4v16M15 4v16"/></svg>',
    tableRowAbove: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="11" width="18" height="9" rx="1"/><path d="M3 15h18M9 11v9M15 11v9M12 3v5M9 6l3-3 3 3"/></svg>',
    tableRowBelow: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="9" rx="1"/><path d="M3 9h18M9 4v9M15 4v9M12 16v5M9 18l3 3 3-3"/></svg>',
    tableRowDelete: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M3 9h18M9 4v16M15 4v16M7 13h10"/></svg>',
    tableColLeft: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="11" y="4" width="10" height="16" rx="1"/><path d="M16 4v16M3 12h6M6 9l-3 3 3 3"/></svg>',
    tableColRight: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="10" height="16" rx="1"/><path d="M8 4v16M15 12h6M18 9l3 3-3 3"/></svg>',
    tableColDelete: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M3 9h18M9 4v16M15 4v16M7 13h10"/></svg>',
    tableHeader: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M3 9h18M9 4v16M15 4v16M3 5h18v4H3z"/></svg>',
    tableMerge: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M12 4v16M8 9l4-4 4 4M8 15l4 4 4-4"/></svg>',
    tableSplit: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M12 4v16M8 8l-4 4 4 4M16 8l4 4-4 4"/></svg>',
    tableAlignLeft: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M7 9h7M7 13h10M7 17h5"/></svg>',
    tableAlignCenter: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M8 9h8M6 13h12M9 17h6"/></svg>',
    tableAlignRight: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1"/><path d="M10 9h7M7 13h10M12 17h5"/></svg>',
    tableDelete: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16M9 7V5h6v2M7 7l1 12h8l1-12M10 11v5M14 11v5"/></svg>',
    tableDock: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 5h16v12H4zM8 9h8M8 13h5M12 17v4M9 20l3 3 3-3"/></svg>',
    tableSelect: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="1" stroke-dasharray="3 2"/><path d="M7 8l3 3-3 3M17 8l-3 3 3 3"/></svg>',
    tableDrag: '<span aria-hidden="true">\u28ff</span>',
    hr: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 12h16"/></svg>',
    blockquote: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 7h5v5H6c0 2 2 3 4 3M14 7h5v5h-4c0 2 2 3 4 3"/></svg>',
    codeBlock: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 7 3 12l5 5M16 7l5 5-5 5M13 5l-2 14"/></svg>',
    textColor: '<span aria-hidden="true" style="font-weight:700">A</span>',
    highlight: '<span aria-hidden="true" style="background:#ffe066;color:#111;padding:0 2px">A</span>',
    paragraph: '<span aria-hidden="true">P</span>'
  };
  var MENUS = {
    file: ['New document|newDoc', 'Open file...|openFile',
      { label: 'Import', items: ['JSON|importJSON', 'HTML|importHTML', 'Markdown|importMarkdown', 'XML|importXML'] },
      { label: 'Export', items: ['JSON|exportJSON', 'HTML|exportHTML', 'Markdown|exportMarkdown', 'XML|exportXML'] },
      { label: 'Save as', items: ['Plain text (.txt)|saveTXT', 'RTF (.rtf)|saveRTF', 'Word (.doc, RTF)|saveDOC', 'Word (.docx)|saveDOCX', 'OpenDocument (.odt)|saveODT', 'PDF (.pdf)|savePDF', 'HTML (.html)|saveHTML', 'Markdown (.md)|saveMarkdown', 'JSON (.json)|saveJSON', 'XML (.xml)|saveXML'] }],
    edit: ['Undo|undo', 'Redo|redo', 'Copy|copy', 'Cut|cut', 'Paste|paste', 'Select all|selectAll'],
    insert: ['Link|link', 'Image|image', 'Table|table', 'Horizontal rule|hr'],
    format: ['Paragraph|paragraph', 'Heading 1|h1', 'Heading 2|h2', 'Heading 3|h3', 'Bold|bold', 'Italic|italic', 'Underline|underline', 'Strikethrough|strike', 'Code|code', 'Clear formatting|clearFormat'],
    view: ['Toggle status bar|toggleStatus', 'Toggle context toolbar|toggleContext', 'Article template|templateArticle', 'Letter template|templateLetter', 'Report template|templateReport', 'Compact template|templateCompact'],
    tools: ['Clear formatting|clearFormat', 'Plugin information|pluginInfo']
  };
  var MENU_LABELS = { file: 'File', edit: 'Edit', insert: 'Insert', format: 'Format', view: 'View', tools: 'Tools' };

  function hasOwn(o, k) { return Object.prototype.hasOwnProperty.call(o, k); }
  function attr(root, name) { return root.getAttribute(name) || ''; }
  function boolAttr(root, name, fallback) {
    var v = attr(root, name);
    if (v === '') return fallback;
    return v !== '0' && v !== 'false';
  }
  function cssLength(value, fallback) {
    var v = String(value == null ? '' : value).trim();
    if (/^\d{1,4}(?:\.\d+)?(?:px|rem|em|%|vh|vw|pt)?$/.test(v)) return v;
    return fallback || '';
  }
  function safeFont(value) {
    var v = String(value == null ? '' : value).trim();
    if (/^[a-zA-Z0-9 ,'.\-"]{1,120}$/.test(v)) return v;
    return '';
  }
  function cssFontSize(value) {
    var v = String(value == null ? '' : value).trim();
    if (/^(?:xx-small|x-small|small|medium|large|x-large|xx-large|smaller|larger|\d+(?:\.\d+)?(?:px|pt|em|rem|%)?)$/.test(v)) return v;
    return '';
  }
  function cssColor(value) {
    var v = String(value == null ? '' : value).trim();
    if (/^#[0-9a-fA-F]{3,8}$/.test(v)) return v;
    if (/^(?:rgb|hsl)a?\([0-9.,%\s]+\)$/.test(v)) return v;
    if (/^[a-zA-Z]{3,20}$/.test(v)) return v;
    return '';
  }
  function safeUrl(value, image) {
    var v = String(value == null ? '' : value).trim();
    if (!v) return '';
    if (image) {
      if (/^data:image\/(?:png|jpe?g|gif|webp|svg\+xml);base64,/i.test(v)) return v;
      if (/^(?:https?:|blob:)/i.test(v)) return v;
      return '';
    }
    if (/^(?:https?:|mailto:|tel:)/i.test(v)) return v;
    if (v.charAt(0) === '#' || v.charAt(0) === '/' || /^\.\.?\//.test(v)) return v;
    return '';
  }
  function el(tag, cls, text) {
    var node = document.createElement(tag);
    if (cls) node.className = cls;
    if (text != null) node.textContent = String(text);
    return node;
  }
  function parseJSON(text, fallback) {
    try {
      var value = JSON.parse(String(text || ''));
      return value && typeof value === 'object' ? value : fallback;
    } catch (e) { return fallback; }
  }
  function emptyDoc() { return { type: 'doc', children: [{ type: 'paragraph', children: [] }] }; }
  function cloneDoc(doc) { try { return JSON.parse(JSON.stringify(doc)); } catch (e) { return emptyDoc(); } }
  function blockLabel(type, node) {
    if (type === 'heading') return 'Heading ' + String(node.level || 1);
    if (type === 'list_item') return 'List item';
    if (type === 'bullet_list') return 'Bullet list';
    if (type === 'ordered_list') return 'Numbered list';
    if (type === 'code_block') return 'Code block';
    if (type === 'table') return 'Table';
    if (type === 'image') return 'Image';
    if (type === 'hr') return 'Horizontal rule';
    if (type === 'blockquote') return 'Quote';
    return 'Paragraph';
  }

  function inlineFrom(node, marks, out, depth) {
    if (!node || out.length > MAX_NODES || depth > 32) return;
    if (node.nodeType === 3) {
      if (node.nodeValue) out.push({ type: 'text', text: node.nodeValue, marks: marks.slice() });
      return;
    }
    if (node.nodeType !== 1) return;
    var tag = node.tagName.toLowerCase();
    if (tag === 'script' || tag === 'style' || tag === 'template') return;
    if (tag === 'br') {
      var parentTag = node.parentNode && node.parentNode.nodeType === 1 ? node.parentNode.tagName.toLowerCase() : '';
      if (node.parentNode && node.parentNode.childNodes.length === 1 && /^(p|h[1-6]|li)$/.test(parentTag)) return;
      out.push({ type: 'break', marks: [] });
      return;
    }
    if (tag === 'img') {
      out.push({ type: 'image', src: safeUrl(node.getAttribute('src'), true), alt: node.getAttribute('alt') || '', title: node.getAttribute('title') || '', width: cssLength(node.style.width, ''), align: node.style.textAlign || '' });
      return;
    }
    var next = marks.slice();
    var mark = null;
    if (tag === 'strong' || tag === 'b') mark = 'bold';
    else if (tag === 'em' || tag === 'i') mark = 'italic';
    else if (tag === 'u') mark = 'underline';
    else if (tag === 's' || tag === 'strike' || tag === 'del') mark = 'strike';
    else if (tag === 'code') mark = 'code';
    else if (tag === 'sup') mark = 'sup';
    else if (tag === 'sub') mark = 'sub';
    if (mark) next.push({ type: mark });
    var style = node.style || {};
    if (style.fontWeight && (style.fontWeight === 'bold' || style.fontWeight === 'bolder' || Number(style.fontWeight) >= 600)) next.push({ type: 'bold' });
    if (style.fontStyle && style.fontStyle.indexOf('italic') >= 0) next.push({ type: 'italic' });
    var decoration = String(style.textDecoration || style.textDecorationLine || '');
    if (decoration.indexOf('underline') >= 0) next.push({ type: 'underline' });
    if (decoration.indexOf('line-through') >= 0) next.push({ type: 'strike' });
    if (style.verticalAlign === 'super') next.push({ type: 'sup' });
    if (style.verticalAlign === 'sub') next.push({ type: 'sub' });
    var color = cssColor(style.color); if (color) next.push({ type: 'color', value: color });
    var bg = cssColor(style.backgroundColor); if (bg && bg !== 'transparent') next.push({ type: 'background', value: bg });
    var family = safeFont(style.fontFamily); if (family) next.push({ type: 'font', value: family });
    var size = String(style.fontSize || ''); if (size && /^(?:xx-small|x-small|small|medium|large|x-large|xx-large|smaller|larger|\d+(?:\.\d+)?(?:px|pt|em|rem|%)?)$/.test(size)) next.push({ type: 'size', value: size });
    if (tag === 'font') {
      var fc = cssColor(node.getAttribute('color')); if (fc) next.push({ type: 'color', value: fc });
      var ff = node.getAttribute('face'); if (ff) next.push({ type: 'font', value: ff });
    }
    if (tag === 'a') {
      var href = safeUrl(node.getAttribute('href'), false);
      if (href) next.push({ type: 'link', href: href, target: node.getAttribute('target') || '', title: node.getAttribute('title') || '' });
    }
    for (var i = 0; i < node.childNodes.length; i++) inlineFrom(node.childNodes[i], next, out, depth + 1);
  }

  function blockStyle(node) {
    var style = node.style || {};
    return {
      align: style.textAlign || '',
      indent: cssLength(style.paddingLeft || style.marginLeft, ''),
      lineHeight: cssLength(style.lineHeight, ''),
      spaceBefore: cssLength(style.marginTop, ''),
      spaceAfter: cssLength(style.marginBottom, '')
    };
  }

  function blocksFrom(parent, depth) {
    var out = [], inline = [], i;
    if (!parent || depth > 32) return out;
    function flush() {
      if (!inline.length) return;
      var clean = inline.slice();
      inline.length = 0;
      if (clean.length === 1 && clean[0].type === 'text' && !String(clean[0].text || '').trim()) return;
      out.push({ type: 'paragraph', children: clean });
    }
    for (i = 0; i < parent.childNodes.length; i++) {
      var node = parent.childNodes[i];
      if (node.nodeType === 3) {
        if (String(node.nodeValue || '').trim() || inline.length) inline.push({ type: 'text', text: node.nodeValue || '', marks: [] });
        continue;
      }
      if (node.nodeType !== 1) continue;
      var tag = node.tagName.toLowerCase();
      if (tag === 'script' || tag === 'style' || tag === 'template') continue;
      if (tag === 'p' || tag === 'div' || tag === 'section' || tag === 'article' || tag === 'main' || tag === 'header' || tag === 'footer') {
        flush();
        var p = { type: 'paragraph', children: [] };
        var st = blockStyle(node);
        for (var sk in st) if (hasOwn(st, sk) && st[sk]) p[sk] = st[sk];
        for (var c = 0; c < node.childNodes.length; c++) inlineFrom(node.childNodes[c], [], p.children, depth + 1);
        out.push(p);
        continue;
      }
      if (/^h[1-6]$/.test(tag)) {
        flush();
        var h = { type: 'heading', level: Number(tag.charAt(1)), children: [] };
        var hs = blockStyle(node);
        for (var hk in hs) if (hasOwn(hs, hk) && hs[hk]) h[hk] = hs[hk];
        for (var hc = 0; hc < node.childNodes.length; hc++) inlineFrom(node.childNodes[hc], [], h.children, depth + 1);
        out.push(h);
        continue;
      }
      if (tag === 'ul' || tag === 'ol') {
        flush();
        var list = { type: tag === 'ul' ? 'bullet_list' : 'ordered_list', children: [] };
        for (var li = 0; li < node.childNodes.length; li++) {
          if (node.childNodes[li].nodeType === 1 && node.childNodes[li].tagName.toLowerCase() === 'li') {
            var item = { type: 'list_item', children: [] };
            var itemBlocks = blocksFrom(node.childNodes[li], depth + 1);
            if (!itemBlocks.length) {
              var itemInline = [];
              for (var ic = 0; ic < node.childNodes[li].childNodes.length; ic++) inlineFrom(node.childNodes[li].childNodes[ic], [], itemInline, depth + 1);
              itemBlocks = [{ type: 'paragraph', children: itemInline }];
            }
            item.children = itemBlocks;
            list.children.push(item);
          }
        }
        out.push(list);
        continue;
      }
      if (tag === 'blockquote') {
        flush();
        out.push({ type: 'blockquote', children: blocksFrom(node, depth + 1) });
        continue;
      }
      if (tag === 'pre') {
        flush();
        out.push({ type: 'code_block', text: node.textContent || '' });
        continue;
      }
      if (tag === 'hr') { flush(); out.push({ type: 'hr' }); continue; }
      if (tag === 'img') {
        flush();
        out.push({ type: 'image', src: safeUrl(node.getAttribute('src'), true), alt: node.getAttribute('alt') || '', title: node.getAttribute('title') || '', width: cssLength(node.style.width, ''), align: node.style.textAlign || '' });
        continue;
      }
      if (tag === 'table') {
        flush();
        var tableAlign = node.style.marginLeft === 'auto' && node.style.marginRight === 'auto' ? 'center' : node.style.marginLeft === 'auto' ? 'right' : node.style.marginRight === 'auto' ? 'left' : '';
        var table = { type: 'table', rows: [], header: false, border: cssColor(node.style.borderColor || '') || '', width: cssLength(node.style.width, ''), height: cssLength(node.style.height, ''), align: tableAlign };
        for (var r = 0; r < node.rows.length; r++) {
          var row = { type: 'table_row', height: cssLength(node.rows[r].style.height, ''), cells: [] };
          for (var ci = 0; ci < node.rows[r].cells.length; ci++) {
            var cell = node.rows[r].cells[ci];
            var cellInline = [];
            for (var cc = 0; cc < cell.childNodes.length; cc++) inlineFrom(cell.childNodes[cc], [], cellInline, depth + 1);
            row.cells.push({ type: 'table_cell', header: cell.tagName.toLowerCase() === 'th', colspan: Number(cell.getAttribute('colspan') || 1), rowspan: Number(cell.getAttribute('rowspan') || 1), align: cell.style.textAlign || '', width: cssLength(cell.style.width, ''), children: cellInline });
          }
          table.rows.push(row);
        }
        if (table.rows.length && table.rows[0].cells.length && table.rows[0].cells[0].header) table.header = true;
        out.push(table);
        continue;
      }
      if (tag === 'figure') {
        flush();
        var img = node.querySelector('img');
        var cap = node.querySelector('figcaption');
        out.push({ type: 'image', src: img ? safeUrl(img.getAttribute('src'), true) : '', alt: img ? (img.getAttribute('alt') || '') : '', caption: cap ? (cap.textContent || '') : '', width: img ? cssLength(img.style.width, '') : '', align: node.style.textAlign || '' });
        continue;
      }
      inlineFrom(node, [], inline, depth + 1);
    }
    flush();
    return out;
  }

  function docFromDom(root) {
    var children = blocksFrom(root, 0);
    if (!children.length) children = [{ type: 'paragraph', children: [] }];
    return { type: 'doc', children: children };
  }

  function styleForBlock(element, node) {
    if (!element || !node) return;
    if (node.align) element.style.textAlign = node.align;
    if (node.indent) element.style.paddingLeft = node.indent;
    if (node.lineHeight) element.style.lineHeight = node.lineHeight;
    if (node.spaceBefore) element.style.marginTop = node.spaceBefore;
    if (node.spaceAfter) element.style.marginBottom = node.spaceAfter;
  }

  function applyMarks(text, marks) {
    var node = document.createTextNode(text);
    for (var i = marks.length - 1; i >= 0; i--) {
      var mark = marks[i];
      var wrapper;
      if (mark.type === 'link') {
        var href = safeUrl(mark.href, false);
        if (!href) continue;
        wrapper = el('a');
        wrapper.setAttribute('href', href);
        if (mark.target === '_blank') { wrapper.setAttribute('target', '_blank'); wrapper.setAttribute('rel', 'noopener noreferrer'); }
        if (mark.title) wrapper.setAttribute('title', String(mark.title));
      } else {
        var tags = { bold: 'strong', italic: 'em', underline: 'u', strike: 's', code: 'code', sup: 'sup', sub: 'sub' };
        var tag = tags[mark.type];
        if (tag) wrapper = el(tag);
        else if (mark.type === 'color') { wrapper = el('span'); wrapper.style.color = cssColor(mark.value) || ''; }
        else if (mark.type === 'background') { wrapper = el('span'); wrapper.style.backgroundColor = cssColor(mark.value) || ''; }
        else if (mark.type === 'font') { wrapper = el('span'); wrapper.style.fontFamily = safeFont(mark.value); }
        else if (mark.type === 'size') { wrapper = el('span'); wrapper.style.fontSize = cssFontSize(mark.value); }
        if (!wrapper) continue;
      }
      wrapper.appendChild(node);
      node = wrapper;
    }
    return node;
  }

  function inlineDom(children, parent) {
    for (var i = 0; i < children.length; i++) {
      var node = children[i];
      if (!node || typeof node !== 'object') continue;
      if (node.type === 'text') parent.appendChild(applyMarks(String(node.text == null ? '' : node.text), node.marks || []));
      else if (node.type === 'break') parent.appendChild(el('br'));
      else if (node.type === 'image') {
        var src = safeUrl(node.src, true);
        if (!src) continue;
        var img = el('img');
        img.setAttribute('src', src);
        if (node.alt) img.setAttribute('alt', String(node.alt));
        if (node.title) img.setAttribute('title', String(node.title));
        if (node.width) img.style.width = cssLength(node.width, '');
        if (node.align) img.style.display = 'block', img.style.marginLeft = node.align === 'right' ? 'auto' : node.align === 'center' ? 'auto' : '0';
        parent.appendChild(img);
      }
    }
  }

  function blockDom(node, depth) {
    if (!node || typeof node !== 'object' || depth > 32) return null;
    var type = String(node.type || 'paragraph'), box = null;
    if (type === 'paragraph') box = el('p');
    else if (type === 'heading') box = el('h' + Math.max(1, Math.min(6, Number(node.level) || 1)));
    else if (type === 'bullet_list' || type === 'ordered_list') {
      box = el(type === 'bullet_list' ? 'ul' : 'ol');
      var items = Array.isArray(node.children) ? node.children : [];
      for (var i = 0; i < items.length; i++) {
        var li = el('li');
        var itemBlocks = Array.isArray(items[i].children) ? items[i].children : [];
        for (var j = 0; j < itemBlocks.length; j++) {
          var child = blockDom(itemBlocks[j], depth + 1);
          if (child) li.appendChild(child);
        }
        if (!li.childNodes.length) li.appendChild(el('p'));
        box.appendChild(li);
      }
      return box;
    } else if (type === 'blockquote') {
      box = el('blockquote');
      var q = Array.isArray(node.children) ? node.children : [];
      for (var qi = 0; qi < q.length; qi++) { var qb = blockDom(q[qi], depth + 1); if (qb) box.appendChild(qb); }
      return box;
    } else if (type === 'code_block') {
      box = el('pre'); box.appendChild(el('code', '', String(node.text == null ? '' : node.text))); return box;
    } else if (type === 'hr') {
      return el('hr');
    } else if (type === 'image') {
      var figure = el('figure');
      var src = safeUrl(node.src, true);
      if (src) {
        var image = el('img'); image.setAttribute('src', src);
        if (node.alt) image.setAttribute('alt', String(node.alt));
        if (node.title) image.setAttribute('title', String(node.title));
        if (node.width) image.style.width = cssLength(node.width, '');
        figure.appendChild(image);
      }
      if (node.caption) figure.appendChild(el('figcaption', '', String(node.caption)));
      if (node.align) figure.style.textAlign = node.align;
      return figure;
    } else if (type === 'table') {
      box = el('table');
      if (node.border) box.style.borderColor = cssColor(node.border);
      if (node.width) box.style.width = cssLength(node.width, '');
      if (node.height) box.style.height = cssLength(node.height, '');
      if (node.align === 'center') { box.style.marginLeft = 'auto'; box.style.marginRight = 'auto'; }
      else if (node.align === 'right') { box.style.marginLeft = 'auto'; box.style.marginRight = '0'; }
      else if (node.align === 'left') { box.style.marginLeft = '0'; box.style.marginRight = 'auto'; }
      if (node.width || node.height) box.style.tableLayout = 'fixed';
      var rows = Array.isArray(node.rows) ? node.rows : [];
      var tbody = el('tbody');
      for (var ri = 0; ri < rows.length; ri++) {
        var tr = el('tr');
        if (rows[ri].height) tr.style.height = cssLength(rows[ri].height, '');
        var cells = Array.isArray(rows[ri].cells) ? rows[ri].cells : [];
        for (var ci = 0; ci < cells.length; ci++) {
          var cell = el(cells[ci].header ? 'th' : 'td');
          if (cells[ci].align) cell.style.textAlign = cells[ci].align;
          if (cells[ci].width) cell.style.width = cssLength(cells[ci].width, '');
          if (Number(cells[ci].colspan) > 1) cell.setAttribute('colspan', String(Number(cells[ci].colspan)));
          if (Number(cells[ci].rowspan) > 1) cell.setAttribute('rowspan', String(Number(cells[ci].rowspan)));
          inlineDom(Array.isArray(cells[ci].children) ? cells[ci].children : [], cell);
          if (!cell.childNodes.length) cell.appendChild(el('p'));
          tr.appendChild(cell);
        }
        tbody.appendChild(tr);
      }
      box.appendChild(tbody);
      return box;
    } else {
      // Plugin-defined block renderer.
      var plugin = plugins[String(node.plugin || '')];
      if (plugin && typeof plugin.render === 'function') {
        try { var rendered = plugin.render(node); if (rendered && rendered.nodeType) return rendered; } catch (e) {}
      }
      box = el('p');
      inlineDom(Array.isArray(node.children) ? node.children : [], box);
    }
    if (box) styleForBlock(box, node);
    if (type === 'paragraph' || type === 'heading') {
      inlineDom(Array.isArray(node.children) ? node.children : [], box);
      if (!box.childNodes.length) box.appendChild(el('br'));
    }
    return box;
  }

  function renderDoc(doc) {
    var frag = document.createDocumentFragment();
    var children = doc && Array.isArray(doc.children) ? doc.children : [];
    if (!children.length) children = [{ type: 'paragraph', children: [] }];
    for (var i = 0; i < children.length && i < MAX_NODES; i++) {
      var b = blockDom(children[i], 0);
      if (b) frag.appendChild(b);
    }
    if (!frag.childNodes.length) frag.appendChild(el('p'));
    return frag;
  }

  function textOf(node) {
    if (!node) return '';
    if (node.type === 'text' || node.type === 'code_block') return String(node.text || '');
    if (node.type === 'break') return '\n';
    if (node.type === 'image') return node.alt || node.caption || '';
    var out = '';
    var kids = node.children || [];
    for (var i = 0; i < kids.length; i++) out += textOf(kids[i]);
    return out;
  }
  function docStats(doc) {
    var text = '', words = 0, paragraphs = 0;
    function walk(node) {
      if (!node) return;
      if (node.type === 'text') { text += node.text || ''; return; }
      if (node.type === 'break') { text += '\n'; return; }
      if (node.type === 'paragraph' || node.type === 'heading' || node.type === 'blockquote' || node.type === 'code_block' || node.type === 'table_cell') {
        if (node.type !== 'table_cell') paragraphs++;
      }
      if (node.type === 'code_block') { text += node.text || ''; return; }
      var kids = node.children || [];
      for (var i = 0; i < kids.length; i++) walk(kids[i]);
      if (node.type === 'table') {
        for (var r = 0; r < (node.rows || []).length; r++) for (var c = 0; c < (node.rows[r].cells || []).length; c++) walk(node.rows[r].cells[c]);
      }
      if (node.type === 'paragraph' || node.type === 'heading' || node.type === 'blockquote') text += '\n';
    }
    walk(doc);
    var parts = text.trim().split(/\s+/);
    words = parts.length && parts[0] ? parts.length : 0;
    return { words: words, chars: text.replace(/\n/g, '').length, paragraphs: paragraphs };
  }

  function currentBlock(inst) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return null;
    var node = sel.anchorNode;
    while (node && node !== inst.editor) {
      if (node.nodeType === 1) {
        var tag = node.tagName.toLowerCase();
        if (/^(p|div|h[1-6]|li|blockquote|pre|figure|table|th|td|ul|ol)$/.test(tag)) return node;
      }
      node = node.parentNode;
    }
    return inst.editor.firstElementChild;
  }

  function topBlock(inst, node) {
    var block = null;
    while (node && node !== inst.editor) {
      if (node.nodeType === 1) {
        var tag = node.tagName.toLowerCase();
        if (/^(p|div|h[1-6]|li|blockquote|pre|figure|table)$/.test(tag)) block = node;
        if (node.parentNode === inst.editor) return node;
      }
      node = node.parentNode;
    }
    return block;
  }
  function textOffsetWithin(root, container, offset) {
    try {
      var range = document.createRange();
      range.selectNodeContents(root);
      range.setEnd(container, offset);
      return range.toString().length;
    } catch (e) { return 0; }
  }
  function selectionInfo(inst) {
    var sel = window.getSelection();
    var info = { collapsed: true, text: '', start: 0, end: 0, block: '', inside: false };
    if (!sel || !sel.rangeCount) return info;
    var range = sel.getRangeAt(0);
    info.collapsed = sel.isCollapsed;
    info.text = sel.toString();
    info.inside = inst.editor.contains(range.commonAncestorContainer);
    var pre = document.createRange();
    pre.selectNodeContents(inst.editor);
    pre.setEnd(range.startContainer, range.startOffset);
    info.start = pre.toString().length;
    pre.setEnd(range.endContainer, range.endOffset);
    info.end = pre.toString().length;
    var block = currentBlock(inst);
    if (block) info.block = blockLabel(block.tagName.toLowerCase() === 'h1' ? 'heading' : 'paragraph', {});
    var top = topBlock(inst, range.startContainer);
    if (top && top.parentNode === inst.editor) {
      var children = Array.prototype.slice.call(inst.editor.children);
      info.blockIndex = children.indexOf(top);
      info.blockOffset = textOffsetWithin(top, range.startContainer, range.startOffset);
    }
    return info;
  }

  function saveSelection(inst) {
    var info = selectionInfo(inst);
    info.active = document.activeElement === inst.editor;
    if (info.active) inst.savedSelection = info;
  }
  function restoreSelection(inst) {
    var info = inst.savedSelection;
    if (!info || !info.active) return;
    inst.savedSelection = null;
    if (info.blockIndex != null && info.blockIndex >= 0) {
      var blocks = inst.editor.children;
      var block = blocks[info.blockIndex];
      if (block) {
        var blockWalker = document.createTreeWalker(block, NodeFilter.SHOW_TEXT, null);
        var blockPos = 0, blockNode = null, blockOffset = 0, bn;
        while ((bn = blockWalker.nextNode())) {
          var blen = (bn.nodeValue || '').length;
          if (info.blockOffset <= blockPos + blen) { blockNode = bn; blockOffset = Math.max(0, info.blockOffset - blockPos); break; }
          blockPos += blen;
        }
        try {
          var blockRange = document.createRange();
          if (blockNode) blockRange.setStart(blockNode, blockOffset);
          else { blockRange.selectNodeContents(block); blockRange.collapse(true); }
          blockRange.collapse(true);
          var blockSel = window.getSelection();
          blockSel.removeAllRanges(); blockSel.addRange(blockRange);
          inst.editor.focus();
        } catch (e) {}
        return;
      }
    }
    var walker = document.createTreeWalker(inst.editor, NodeFilter.SHOW_TEXT, null);
    var pos = 0, startNode = null, startOffset = 0, endNode = null, endOffset = 0, node;
    while ((node = walker.nextNode())) {
      var len = (node.nodeValue || '').length;
      if (!startNode && info.start <= pos + len) { startNode = node; startOffset = Math.max(0, info.start - pos); }
      if (!endNode && info.end <= pos + len) { endNode = node; endOffset = Math.max(0, info.end - pos); break; }
      pos += len;
    }
    if (!startNode) {
      // Empty blocks contain only <br>; place the caret in the last block.
      var blocks = inst.editor.querySelectorAll('p,h1,h2,h3,h4,blockquote,pre,li');
      var target = blocks.length ? blocks[blocks.length - 1] : inst.editor;
      try {
        var emptyRange = document.createRange();
        emptyRange.selectNodeContents(target); emptyRange.collapse(true);
        var emptySel = window.getSelection();
        emptySel.removeAllRanges(); emptySel.addRange(emptyRange);
        inst.editor.focus();
      } catch (e) {}
      return;
    }
    try {
      var range = document.createRange();
      range.setStart(startNode, startOffset);
      if (endNode) range.setEnd(endNode, endOffset); else range.collapse(true);
      var sel = window.getSelection();
      sel.removeAllRanges(); sel.addRange(range);
      inst.editor.focus();
    } catch (e) {}
  }
  function notify(inst, kind) {
    saveSelection(inst);
    var input = inst.root.querySelector('.sui-rich-value');
    if (input) input.value = JSON.stringify(inst.doc);
    var setter = attr(inst.root, 'data-sui-rich-setter');
    var handler = kind === 'input' ? attr(inst.root, 'data-sui-rich-oninput') : attr(inst.root, 'data-sui-rich-onchange');
    // Live typing only updates the hidden value and onInput; the state setter
    // runs on change/blur/commands so a bound state does not replace the
    // contenteditable surface and its caret on every keystroke.
    if (setter && kind !== 'input' && typeof window[setter] === 'function') window[setter](JSON.stringify(inst.doc));
    if (handler && typeof window[handler] === 'function') window[handler](JSON.stringify(inst.doc));
  }

  function sync(inst) {
    var next = docFromDom(inst.editor);
    var changed = JSON.stringify(next) !== JSON.stringify(inst.doc);
    if (changed) { inst.doc = next; inst.dirty = true; }
    return changed;
  }
  function remember(inst) {
    var text = JSON.stringify(inst.doc);
    if (inst.history[inst.index] === text) return;
    inst.history = inst.history.slice(0, inst.index + 1);
    inst.history.push(text);
    if (inst.history.length > MAX_HISTORY) inst.history.shift();
    inst.index = inst.history.length - 1;
  }
  function commit(inst, kind) {
    sync(inst);
    remember(inst);
    updateStats(inst);
    notify(inst, kind || 'change');
  }
  function restore(inst, text) {
    inst.doc = parseJSON(text, emptyDoc());
    inst.editor.innerHTML = '';
    inst.editor.appendChild(renderDoc(inst.doc));
    updateStats(inst);
  }
  function undo(inst) {
    sync(inst);
    if (inst.index <= 0) return false;
    inst.index--;
    restore(inst, inst.history[inst.index]);
    notify(inst, 'change');
    return true;
  }
  function redo(inst) {
    sync(inst);
    if (inst.index >= inst.history.length - 1) return false;
    inst.index++;
    restore(inst, inst.history[inst.index]);
    notify(inst, 'change');
    return true;
  }

  function insertNode(inst, node) {
    var sel = window.getSelection();
    inst.editor.focus();
    if (!sel || !sel.rangeCount) { inst.editor.appendChild(node); return; }
    var range = sel.getRangeAt(0);
    range.deleteContents();
    range.insertNode(node);
    range.setStartAfter(node);
    range.collapse(true);
    sel.removeAllRanges();
    sel.addRange(range);
  }
  function insertHTMLNode(inst, tag, attrs) {
    var node = el(tag);
    for (var k in attrs || {}) if (hasOwn(attrs, k)) node.setAttribute(k, String(attrs[k]));
    insertNode(inst, node);
    return node;
  }
  function wrapSelection(inst, tag) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount || sel.isCollapsed) return;
    var range = sel.getRangeAt(0);
    var wrapper = el(tag);
    try { wrapper.appendChild(range.extractContents()); range.insertNode(wrapper); }
    catch (e) { var content = range.extractContents(); wrapper.appendChild(content); range.insertNode(wrapper); }
    sel.removeAllRanges();
    var r = document.createRange(); r.selectNodeContents(wrapper); sel.addRange(r);
  }
  function unwrapTag(inst, tag) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return false;
    var node = sel.anchorNode, found = null;
    while (node && node !== inst.editor) {
      if (node.nodeType === 1 && node.tagName.toLowerCase() === tag) { found = node; break; }
      node = node.parentNode;
    }
    if (!found) return false;
    var parent = found.parentNode;
    while (found.firstChild) parent.insertBefore(found.firstChild, found);
    parent.removeChild(found);
    return true;
  }
  function applyBlock(inst, tag) {
    var block = currentBlock(inst);
    if (!block) return;
    var next = el(tag);
    for (var i = 0; i < block.attributes.length; i++) if (block.attributes[i].name === 'style') next.setAttribute('style', block.getAttribute('style'));
    while (block.firstChild) next.appendChild(block.firstChild);
    block.parentNode.replaceChild(next, block);
    var sel = window.getSelection();
    if (sel) { var r = document.createRange(); r.selectNodeContents(next); r.collapse(false); sel.removeAllRanges(); sel.addRange(r); }
  }
  function buildTable(rows, cols) {
    var table = el('table');
    var tbody = el('tbody');
    rows = Math.max(1, Math.min(30, Number(rows) || 3));
    cols = Math.max(1, Math.min(20, Number(cols) || 3));
    for (var r = 0; r < rows; r++) {
      var tr = el('tr');
      for (var c = 0; c < cols; c++) {
        var cell = el(r === 0 ? 'th' : 'td');
        var cellP = el('p');
        cellP.appendChild(el('br'));
        cell.appendChild(cellP);
        tr.appendChild(cell);
      }
      tbody.appendChild(tr);
    }
    table.appendChild(tbody);
    return table;
  }
  function insertTable(inst, rows, cols) {
    var table = buildTable(rows, cols);
    insertBlock(inst, table);
    var first = table.querySelector('td,th');
    if (first) { var range = document.createRange(); range.selectNodeContents(first); range.collapse(true); var sel = window.getSelection(); sel.removeAllRanges(); sel.addRange(range); }
  }
  function closestIn(inst, tag) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return null;
    var node = sel.anchorNode;
    while (node && node !== inst.editor) {
      if (node.nodeType === 1 && node.tagName.toLowerCase() === tag) return node;
      node = node.parentNode;
    }
    return null;
  }
  function clearCellSelection(inst) {
    if (inst.selectedCells) {
      for (var i = 0; i < inst.selectedCells.length; i++) if (inst.selectedCells[i] && inst.selectedCells[i].classList) inst.selectedCells[i].classList.remove('sui-rich-cell-selected');
    }
    if (inst.selectedTable) inst.selectedTable.classList.remove('sui-rich-table-selected');
    inst.selectedCells = [];
    inst.selectedTable = null;
    inst.cellAnchor = null;
  }
  function selectWholeTable(inst, table) {
    clearCellSelection(inst);
    inst.selectedTable = table;
    inst.selectedCells = Array.prototype.slice.call(table.querySelectorAll('td,th'));
    table.classList.add('sui-rich-table-selected');
    updateCellSelection(inst);
  }
  function updateCellSelection(inst) {
    var cells = inst.editor.querySelectorAll('td,th');
    for (var i = 0; i < cells.length; i++) cells[i].classList.remove('sui-rich-cell-selected');
    for (var j = 0; j < (inst.selectedCells || []).length; j++) if (inst.selectedCells[j] && inst.selectedCells[j].classList) inst.selectedCells[j].classList.add('sui-rich-cell-selected');
  }
  function selectCellRange(inst, table, a, b) {
    if (!a || !b) return;
    var minRow = Math.min(a.parentNode.rowIndex, b.parentNode.rowIndex);
    var maxRow = Math.max(a.parentNode.rowIndex, b.parentNode.rowIndex);
    var minCol = Math.min(a.cellIndex, b.cellIndex);
    var maxCol = Math.max(a.cellIndex, b.cellIndex);
    var out = [];
    for (var r = 0; r < table.rows.length; r++) {
      if (table.rows[r].rowIndex < minRow || table.rows[r].rowIndex > maxRow) continue;
      for (var c = 0; c < table.rows[r].cells.length; c++) {
        var cell = table.rows[r].cells[c];
        if (cell.cellIndex >= minCol && cell.cellIndex <= maxCol) out.push(cell);
      }
    }
    inst.selectedCells = out;
    inst.cellAnchor = a;
    updateCellSelection(inst);
  }
  function cellMatrix(inst, table) {
    var cells = selectedCells(inst, table);
    if (!cells.length) return null;
    var minRow = Infinity, maxRow = -1, minCol = Infinity, maxCol = -1;
    for (var i = 0; i < cells.length; i++) {
      var r = cells[i].parentNode.rowIndex, c = cells[i].cellIndex;
      minRow = Math.min(minRow, r); maxRow = Math.max(maxRow, r);
      minCol = Math.min(minCol, c); maxCol = Math.max(maxCol, c);
    }
    var matrix = [];
    for (var rr = minRow; rr <= maxRow; rr++) {
      var row = [];
      for (var cc = minCol; cc <= maxCol; cc++) {
        var value = '';
        for (var k = 0; k < cells.length; k++) {
          if (cells[k].parentNode.rowIndex === rr && cells[k].cellIndex === cc) { value = cells[k].innerText || cells[k].textContent || ''; break; }
        }
        row.push(String(value).replace(/\r/g, ''));
      }
      matrix.push(row);
    }
    return matrix;
  }
  function matrixText(matrix) {
    return matrix.map(function (row) { return row.join('\t'); }).join('\n');
  }
  function setCellText(cell, text) {
    while (cell.firstChild) cell.removeChild(cell.firstChild);
    var p = el('p');
    var lines = String(text == null ? '' : text).replace(/\r/g, '').split('\n');
    for (var i = 0; i < lines.length; i++) {
      if (i > 0) p.appendChild(el('br'));
      if (lines[i]) p.appendChild(document.createTextNode(lines[i]));
    }
    if (!p.firstChild) p.appendChild(el('br'));
    cell.appendChild(p);
  }
  function copyCellSelection(inst, cut) {
    var table = closestIn(inst, 'table') || inst.selectedTable;
    if (!table) return '';
    var matrix = cellMatrix(inst, table);
    if (!matrix) return '';
    inst.cellClipboard = matrix;
    if (cut) {
      var cells = selectedCells(inst, table);
      for (var i = 0; i < cells.length; i++) setCellText(cells[i], '');
    }
    return matrixText(matrix);
  }
  function ensureTableSize(inst, table, startCell, rows, cols) {
    var startRow = startCell.parentNode.rowIndex;
    while (table.rows.length < startRow + rows) {
      var clone = table.rows[table.rows.length - 1].cloneNode(true);
      for (var c = 0; c < clone.cells.length; c++) setCellText(clone.cells[c], '');
      table.rows[table.rows.length - 1].parentNode.appendChild(clone);
    }
    for (var r = startRow; r < startRow + rows; r++) {
      while (table.rows[r].cells.length < startCell.cellIndex + cols) {
        var cell = el(table.rows[r].cells.length ? table.rows[r].cells[0].tagName.toLowerCase() : 'td');
        cell.appendChild(el('p'));
        table.rows[r].appendChild(cell);
      }
    }
  }
  function pasteCellMatrix(inst, text) {
    var table = closestIn(inst, 'table') || inst.selectedTable;
    if (!table) return false;
    var selected = closestIn(inst, 'td,th');
    var start = (selected && table.contains(selected)) ? selected : ((inst.selectedCells && inst.selectedCells[0] && inst.selectedCells[0].isConnected) ? inst.selectedCells[0] : table.querySelector('td,th'));
    if (!start) start = table.rows[0].cells[0];
    var matrix = inst.cellClipboard;
    if (!matrix) {
      if (typeof text !== 'string' || (!text.length)) return false;
      matrix = text.replace(/\r/g, '').split('\n').map(function (line) { return line.split('\t'); });
    }
    if (!matrix.length || !matrix[0].length) return false;
    ensureTableSize(inst, table, start, matrix.length, matrix[0].length);
    var startRow = start.parentNode.rowIndex, startCol = start.cellIndex;
    var pasted = [];
    for (var r = 0; r < matrix.length; r++) {
      for (var c = 0; c < matrix[r].length; c++) {
        var cell = table.rows[startRow + r] && table.rows[startRow + r].cells[startCol + c];
        if (!cell) continue;
        setCellText(cell, matrix[r][c]);
        pasted.push(cell);
      }
    }
    inst.selectedTable = null;
    inst.selectedCells = pasted;
    inst.cellAnchor = pasted[0] || null;
    updateCellSelection(inst);
    return true;
  }
  function selectedCells(inst, table) {
    if (inst.selectedCells && inst.selectedCells.length) {
      var selected = [];
      for (var si = 0; si < inst.selectedCells.length; si++) {
        var sc = inst.selectedCells[si];
        if (sc && sc.isConnected && sc.closest('table') === table) selected.push(sc);
      }
      if (selected.length) return selected;
    }
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return [];
    var range = sel.getRangeAt(0);
    var out = [];
    var cells = table.querySelectorAll('td,th');
    for (var i = 0; i < cells.length; i++) {
      var hit = false;
      try { hit = range.intersectsNode(cells[i]); } catch (e) { hit = true; }
      if (hit) out.push(cells[i]);
    }
    return out;
  }
  function forceCellCaret(inst, cell) {
    if (!cell) return;
    var caretBlock = cell.querySelector('p') || cell;
    if (!caretBlock.firstChild) caretBlock.appendChild(el('br'));
    var caretRange = document.createRange();
    var textNode = null;
    for (var i = 0; i < caretBlock.childNodes.length; i++) if (caretBlock.childNodes[i].nodeType === 3) { textNode = caretBlock.childNodes[i]; break; }
    if (textNode) { caretRange.setStart(textNode, textNode.nodeValue.length); caretRange.collapse(true); }
    else { caretRange.setStart(caretBlock, 0); caretRange.collapse(true); }
    var caretSel = window.getSelection();
    caretSel.removeAllRanges(); caretSel.addRange(caretRange);
    inst.editor.focus();
  }
  function tableResizeHit(inst, e) {
    var cell = e.target && e.target.closest ? e.target.closest('td,th') : null;
    if (!cell) return null;
    var table = cell.closest('table');
    if (!table) return null;
    var tr = table.getBoundingClientRect();
    if (e.clientX >= tr.right - 12 && e.clientY >= tr.bottom - 12) {
      return { kind: 'table', table: table, startX: e.clientX, startY: e.clientY, width: tr.width, height: tr.height };
    }
    var cr = cell.getBoundingClientRect();
    if (e.clientX >= cr.right - 6) {
      return { kind: 'col', table: table, index: cell.cellIndex, startX: e.clientX, width: cr.width };
    }
    if (e.clientY >= cr.bottom - 6) {
      return { kind: 'row', table: table, row: cell.parentNode, startY: e.clientY, height: cr.height };
    }
    return null;
  }
  function startTableResize(inst, e, hit) {
    e.preventDefault();
    e.stopPropagation();
    var table = hit.table;
    if (hit.kind === 'col') {
      table.style.tableLayout = 'fixed';
      table.style.width = table.getBoundingClientRect().width + 'px';
    }
    function move(ev) {
      if (hit.kind === 'table') {
        table.style.width = Math.max(120, hit.width + ev.clientX - hit.startX) + 'px';
        table.style.height = Math.max(60, hit.height + ev.clientY - hit.startY) + 'px';
      } else if (hit.kind === 'col') {
        var width = Math.max(40, hit.width + ev.clientX - hit.startX);
        for (var r = 0; r < table.rows.length; r++) {
          for (var c = 0; c < table.rows[r].cells.length; c++) {
            if (table.rows[r].cells[c].cellIndex === hit.index) table.rows[r].cells[c].style.width = width + 'px';
          }
        }
      } else if (hit.kind === 'row') {
        hit.row.style.height = Math.max(24, hit.height + ev.clientY - hit.startY) + 'px';
      }
    }
    function up() {
      document.removeEventListener('mousemove', move);
      document.removeEventListener('mouseup', up);
      inst.editor.style.cursor = '';
      commit(inst, 'change');
    }
    document.addEventListener('mousemove', move);
    document.addEventListener('mouseup', up);
  }
  function bindTableResize(inst) {
    inst.editor.addEventListener('mousemove', function (e) {
      var hit = tableResizeHit(inst, e);
      inst.editor.style.cursor = hit ? (hit.kind === 'row' ? 'row-resize' : hit.kind === 'table' ? 'nwse-resize' : 'col-resize') : '';
    });
    inst.editor.addEventListener('mouseleave', function () { inst.editor.style.cursor = ''; });
  }
  function bindCellSelection(inst) {
    inst.editor.addEventListener('mousedown', function (e) {
      var resizeHit = tableResizeHit(inst, e);
      if (resizeHit) { startTableResize(inst, e, resizeHit); return; }
      var cell = e.target && e.target.closest ? e.target.closest('td,th') : null;
      if (!cell || !inst.editor.contains(cell)) { clearCellSelection(inst); return; }
      var table = cell.closest('table');
      if (inst.selectedTable) { inst.selectedTable.classList.remove('sui-rich-table-selected'); inst.selectedTable = null; }
      inst.cellDragging = true;
      if ((e.shiftKey || e.ctrlKey || e.metaKey) && inst.cellAnchor && inst.cellAnchor.closest('table') === table) {
        selectCellRange(inst, table, inst.cellAnchor, cell);
      } else {
        inst.selectedCells = [cell];
        inst.cellAnchor = cell;
        updateCellSelection(inst);
      }
      // Clicking the empty padding of a cell can make the browser put the
      // caret in the neighbouring cell. Force it into the clicked cell.
      if (e.target === cell || !String(cell.textContent || '').trim()) {
        e.preventDefault();
        forceCellCaret(inst, cell);
      }
    });
    inst.editor.addEventListener('mouseup', function (e) {
      if (!inst.cellAnchor || !inst.selectedCells || inst.selectedCells.length !== 1) return;
      var cell = inst.cellAnchor;
      if (!inst.editor.contains(cell)) return;
      if (e.target === cell || !String(cell.textContent || '').trim()) {
        e.preventDefault();
        forceCellCaret(inst, cell);
      }
    });
    inst.editor.addEventListener('click', function (e) {
      var cell = e.target && e.target.closest ? e.target.closest('td,th') : null;
      if (!cell || !inst.editor.contains(cell)) return;
      // Chromium can finalize the caret on click after mousedown; repeat the
      // fix for empty cells and clicks on cell padding.
      if (e.target === cell || !String(cell.textContent || '').trim()) {
        e.preventDefault();
        forceCellCaret(inst, inst.cellAnchor && inst.editor.contains(inst.cellAnchor) ? inst.cellAnchor : cell);
      }
    });
    inst.editor.addEventListener('mouseover', function (e) {
      if (!inst.cellDragging || !e.target || !e.target.closest) return;
      var cell = e.target.closest('td,th');
      if (!cell || !inst.cellAnchor) return;
      var table = cell.closest('table');
      if (table && table === inst.cellAnchor.closest('table')) selectCellRange(inst, table, inst.cellAnchor, cell);
    });
    document.addEventListener('mouseup', function () { inst.cellDragging = false; });
  }
  function tableCommand(inst, kind) {
    var table = closestIn(inst, 'table');
    if (!table) return;
    var row = closestIn(inst, 'tr');
    var cells = selectedCells(inst, table);
    if (!row && cells.length) row = cells[0].parentNode;
    if (!row && table.rows.length) row = table.rows[0];
    if (kind === 'addRowAbove' && row) {
      var above = row.cloneNode(true);
      for (var ai = 0; ai < above.cells.length; ai++) above.cells[ai].innerHTML = '<p><br></p>';
      row.parentNode.insertBefore(above, row);
    } else if (kind === 'addRowBelow' && row) {
      var below = row.cloneNode(true);
      for (var bi = 0; bi < below.cells.length; bi++) below.cells[bi].innerHTML = '<p><br></p>';
      row.parentNode.insertBefore(below, row.nextSibling);
    } else if (kind === 'deleteRow' && row && table.rows.length > 1) row.parentNode.removeChild(row);
    else if (kind === 'addColumnLeft' && row) {
      var li = row.cellIndex;
      for (var lr = 0; lr < table.rows.length; lr++) {
        var lc = table.rows[lr].insertCell(li);
        lc.innerHTML = '<p><br></p>';
      }
    } else if (kind === 'addColumnRight' && row) {
      var ri = row.cellIndex + 1;
      for (var rr = 0; rr < table.rows.length; rr++) {
        var rc = table.rows[rr].insertCell(ri);
        rc.innerHTML = '<p><br></p>';
      }
    } else if (kind === 'deleteColumn' && row && row.cells.length > 1) {
      var col = row.cellIndex;
      for (var dr = 0; dr < table.rows.length; dr++) if (table.rows[dr].cells[col]) table.rows[dr].deleteCell(col);
    } else if (kind === 'toggleHeader' && table.rows.length) {
      var first = table.rows[0];
      var makeTh = first.cells.length && first.cells[0].tagName.toLowerCase() !== 'th';
      for (var hc = 0; hc < first.cells.length; hc++) {
        var replacement = el(makeTh ? 'th' : 'td');
        replacement.innerHTML = first.cells[hc].innerHTML;
        first.replaceChild(replacement, first.cells[hc]);
      }
    } else if (kind === 'deleteTable') {
      table.parentNode.removeChild(table);
    } else if (kind === 'dock') {
      dockTableBar(inst);
    } else if (kind === 'merge' && cells.length > 1) {
      var minRow = Infinity, maxRow = -1, minCol = Infinity, maxCol = -1;
      for (var mr = 0; mr < cells.length; mr++) {
        var cr = cells[mr].parentNode.rowIndex, cc = cells[mr].cellIndex;
        minRow = Math.min(minRow, cr); maxRow = Math.max(maxRow, cr);
        minCol = Math.min(minCol, cc); maxCol = Math.max(maxCol, cc);
      }
      var firstCell = cells[0];
      var content = '<p>';
      for (var mi = 0; mi < cells.length; mi++) content += (cells[mi].textContent || '') + (mi < cells.length - 1 ? ' ' : '');
      content += '</p>';
      firstCell.innerHTML = content;
      firstCell.setAttribute('colspan', String(maxCol - minCol + 1));
      firstCell.setAttribute('rowspan', String(maxRow - minRow + 1));
      for (var md = 1; md < cells.length; md++) if (cells[md].parentNode) cells[md].parentNode.removeChild(cells[md]);
    } else if (kind === 'split' && cells.length) {
      var cell = cells[0];
      var span = Number(cell.getAttribute('colspan') || 1);
      if (span > 1) {
        cell.setAttribute('colspan', '1');
        for (var si = 1; si < span; si++) {
          var split = el('td');
          split.innerHTML = '<p><br></p>';
          cell.parentNode.insertBefore(split, cell.nextSibling);
        }
      }
    } else if (kind === 'selectTable') {
      selectWholeTable(inst, table);
    } else if (kind === 'alignLeft' || kind === 'alignCenter' || kind === 'alignRight') {
      var align = kind === 'alignLeft' ? 'left' : kind === 'alignCenter' ? 'center' : 'right';
      if (inst.selectedTable === table) {
        if (align === 'center') { table.style.marginLeft = 'auto'; table.style.marginRight = 'auto'; }
        else if (align === 'right') { table.style.marginLeft = 'auto'; table.style.marginRight = '0'; }
        else { table.style.marginLeft = '0'; table.style.marginRight = 'auto'; }
      } else {
        for (var ci = 0; ci < cells.length; ci++) cells[ci].style.textAlign = align;
      }
    }
  }

  function applyBlockProp(inst, prop, value) {
    var block = currentBlock(inst);
    if (block) block.style[prop] = value;
  }
  function changeIndent(inst, delta) {
    var block = currentBlock(inst);
    if (!block) return;
    var current = parseFloat(block.style.paddingLeft || block.style.marginLeft || '0') || 0;
    current = Math.max(0, Math.min(240, current + delta));
    block.style.paddingLeft = current + 'px';
  }
  function toggleList(inst, type) {
    var block = currentBlock(inst);
    if (!block) return;
    var list = block.closest('ul,ol');
    if (list && list.parentNode === inst.editor) {
      var parent = list.parentNode;
      var items = Array.prototype.slice.call(list.children);
      for (var i = 0; i < items.length; i++) {
        var p = el('p');
        while (items[i].firstChild) p.appendChild(items[i].firstChild);
        parent.insertBefore(p, list);
      }
      parent.removeChild(list);
      return;
    }
    if (block.tagName.toLowerCase() === 'li') {
      var parentList = block.parentNode;
      var p2 = el('p');
      while (block.firstChild) p2.appendChild(block.firstChild);
      parentList.parentNode.replaceChild(p2, parentList);
      return;
    }
    var listEl = el(type);
    var li = el('li');
    while (block.firstChild) li.appendChild(block.firstChild);
    listEl.appendChild(li);
    block.parentNode.replaceChild(listEl, block);
    var range = document.createRange(); range.selectNodeContents(li); range.collapse(false);
    var sel = window.getSelection(); sel.removeAllRanges(); sel.addRange(range);
  }
  function insertBlock(inst, node) {
    var block = currentBlock(inst);
    if (block && block.tagName.toLowerCase() === 'li') {
      var list = block.closest('ul,ol');
      if (list && list.parentNode) { list.parentNode.insertBefore(node, list.nextSibling); return; }
    }
    if (block && block.parentNode) block.parentNode.insertBefore(node, block.nextSibling);
    else inst.editor.appendChild(node);
  }
  function splitBlock(inst) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return;
    var range = sel.getRangeAt(0);
    var block = currentBlock(inst);
    if (!block || !inst.editor.contains(block)) return;
    var tag = block.tagName.toLowerCase();
    if (tag === 'pre' || tag === 'table' || tag === 'td' || tag === 'th') {
      document.execCommand('insertLineBreak');
      return;
    }
    var nextTag = tag === 'li' ? 'li' : 'p';
    var after = range.cloneRange();
    after.setStart(range.endContainer, range.endOffset);
    after.setEnd(block, block.childNodes.length);
    var fragment = after.extractContents();
    var next = el(nextTag);
    next.appendChild(fragment);
    if (!next.textContent && !next.querySelector('img,table,hr,br')) next.appendChild(el('br'));
    block.parentNode.insertBefore(next, block.nextSibling);
    inst.editor.focus();
    var nextRange = document.createRange();
    nextRange.selectNodeContents(next);
    nextRange.collapse(true);
    sel.removeAllRanges(); sel.addRange(nextRange);
  }
  function clearFormatting(inst) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return;
    var range = sel.getRangeAt(0);
    var nodes = inst.editor.querySelectorAll('span,font,strong,b,em,i,u,s,strike,code,sup,sub,a');
    for (var i = nodes.length - 1; i >= 0; i--) {
      var node = nodes[i];
      var intersects = false;
      try { intersects = range.intersectsNode(node); } catch (e) { intersects = true; }
      if (!intersects || !node.parentNode) continue;
      if (node.tagName.toLowerCase() === 'a') {
        while (node.firstChild) node.parentNode.insertBefore(node.firstChild, node);
        node.parentNode.removeChild(node);
      } else {
        while (node.firstChild) node.parentNode.insertBefore(node.firstChild, node);
        node.parentNode.removeChild(node);
      }
    }
    try { document.execCommand('removeFormat'); } catch (e) {}
  }
  function runBuiltin(inst, cmd, value) {
    var editor = inst.editor;
    var selection = window.getSelection();
    if ((!selection || !selection.rangeCount || !editor.contains(selection.anchorNode)) && inst.savedSelection) {
      restoreSelection(inst);
      selection = window.getSelection();
    }
    var savedRange = selection && selection.rangeCount ? selection.getRangeAt(0).cloneRange() : null;
    if (document.activeElement !== editor) {
      editor.focus();
      if (savedRange && savedRange.startContainer && editor.contains(savedRange.startContainer)) {
        try { selection.removeAllRanges(); selection.addRange(savedRange); } catch (e) {}
      }
    }
    try { document.execCommand('styleWithCSS', false, true); } catch (e) {}
    if (cmd === 'undo') return undo(inst);
    if (cmd === 'redo') return redo(inst);
    if (cmd === 'bold') document.execCommand('bold');
    else if (cmd === 'italic') document.execCommand('italic');
    else if (cmd === 'underline') document.execCommand('underline');
    else if (cmd === 'strike') document.execCommand('strikeThrough');
    else if (cmd === 'superscript') document.execCommand('superscript');
    else if (cmd === 'subscript') document.execCommand('subscript');
    else if (cmd === 'code') { if (!unwrapTag(inst, 'code')) wrapSelection(inst, 'code'); }
    else if (cmd === 'clearFormat') clearFormatting(inst);
    else if (cmd === 'paragraph') {
      var blockTag = String(value || '');
      applyBlock(inst, /^(h[1-4]|blockquote|pre)$/.test(blockTag) ? blockTag : 'p');
    }
    else if (/^h[1-4]$/.test(cmd)) applyBlock(inst, cmd);
    else if (cmd === 'blockquote') applyBlock(inst, 'blockquote');
    else if (cmd === 'codeBlock') applyBlock(inst, 'pre');
    else if (cmd === 'alignLeft') applyBlockProp(inst, 'textAlign', 'left');
    else if (cmd === 'alignCenter') applyBlockProp(inst, 'textAlign', 'center');
    else if (cmd === 'alignRight') applyBlockProp(inst, 'textAlign', 'right');
    else if (cmd === 'justify') applyBlockProp(inst, 'textAlign', 'justify');
    else if (cmd === 'indent') changeIndent(inst, 24);
    else if (cmd === 'outdent') changeIndent(inst, -24);
    else if (cmd === 'bulletList') toggleList(inst, 'ul');
    else if (cmd === 'orderedList') toggleList(inst, 'ol');
    else if (cmd === 'hr') insertBlock(inst, el('hr'));
    else if (cmd === 'link') {
      var url = window.prompt('Link URL', 'https://');
      if (url && safeUrl(url, false)) {
        var sel = window.getSelection();
        if (sel && !sel.isCollapsed) document.execCommand('createLink', false, safeUrl(url, false));
        else { var a = el('a', '', url); a.setAttribute('href', safeUrl(url, false)); insertNode(inst, a); }
      }
    } else if (cmd === 'unlink') document.execCommand('unlink');
    else if (cmd === 'image') {
      var src = window.prompt('Image URL', 'https://');
      if (src && safeUrl(src, true)) {
        var alt = window.prompt('Alternative text', '') || '';
        insertHTMLNode(inst, 'img', { src: safeUrl(src, true), alt: alt });
      }
    } else if (cmd === 'table') { var tableNode = buildTable(3, 3); insertBlock(inst, tableNode); }
    else if (cmd === 'symbol') {
      var sym = window.prompt('Insert symbol', '\u00a9');
      if (sym != null) document.execCommand('insertText', false, sym);
    }
    else if (cmd === 'tableAddRow') tableCommand(inst, 'addRowBelow');
    else if (cmd === 'tableDeleteRow') tableCommand(inst, 'deleteRow');
    else if (cmd === 'tableAddColumn') tableCommand(inst, 'addColumnRight');
    else if (cmd === 'tableDeleteColumn') tableCommand(inst, 'deleteColumn');
    else if (cmd === 'tableRowAbove') tableCommand(inst, 'addRowAbove');
    else if (cmd === 'tableRowBelow') tableCommand(inst, 'addRowBelow');
    else if (cmd === 'tableRowDelete') tableCommand(inst, 'deleteRow');
    else if (cmd === 'tableColLeft') tableCommand(inst, 'addColumnLeft');
    else if (cmd === 'tableColRight') tableCommand(inst, 'addColumnRight');
    else if (cmd === 'tableColDelete') tableCommand(inst, 'deleteColumn');
    else if (cmd === 'tableHeader') tableCommand(inst, 'toggleHeader');
    else if (cmd === 'tableDelete') tableCommand(inst, 'deleteTable');
    else if (cmd === 'tableMerge') tableCommand(inst, 'merge');
    else if (cmd === 'tableSplit') tableCommand(inst, 'split');
    else if (cmd === 'tableAlignLeft') tableCommand(inst, 'alignLeft');
    else if (cmd === 'tableAlignCenter') tableCommand(inst, 'alignCenter');
    else if (cmd === 'tableAlignRight') tableCommand(inst, 'alignRight');
    else if (cmd === 'tableDock') tableCommand(inst, 'dock');
    else if (cmd === 'tableSelect') tableCommand(inst, 'selectTable');
    else if (cmd === 'fontFamily') document.execCommand('fontName', false, String(value || ''));
    else if (cmd === 'fontSize') document.execCommand('fontSize', false, String(value || '3'));
    else if (cmd === 'textColor') document.execCommand('foreColor', false, cssColor(value) || '#000000');
    else if (cmd === 'highlight') document.execCommand('hiliteColor', false, cssColor(value) || '#ffff00');
    else if (cmd === 'lineHeight' || cmd === 'spaceBefore' || cmd === 'spaceAfter' || cmd === 'indentBy') {
      var block = currentBlock(inst);
      if (block) {
        if (cmd === 'lineHeight') block.style.lineHeight = cssLength(value, '');
        else if (cmd === 'spaceBefore') block.style.marginTop = cssLength(value, '');
        else if (cmd === 'spaceAfter') block.style.marginBottom = cssLength(value, '');
        else block.style.paddingLeft = cssLength(value, '');
      }
    } else if (cmd === 'copy') {
      if (inst.selectedCells && inst.selectedCells.length > 1) { copyCellSelection(inst, false); }
      document.execCommand('copy');
    }
    else if (cmd === 'cut') {
      if (inst.selectedCells && inst.selectedCells.length > 1) { copyCellSelection(inst, true); commit(inst); }
      else document.execCommand('cut');
    }
    else if (cmd === 'paste') {
      var pasteTable = closestIn(inst, 'table') || inst.selectedTable;
      if (pasteTable && inst.cellClipboard) { pasteCellMatrix(inst, ''); commit(inst); return true; }
      if (navigator.clipboard && navigator.clipboard.readText) navigator.clipboard.readText().then(function (text) {
        inst.editor.focus();
        if (pasteTable && (text.indexOf('\t') >= 0 || text.indexOf('\n') >= 0)) { if (pasteCellMatrix(inst, text)) { commit(inst); return; } }
        document.execCommand('insertText', false, text); commit(inst);
      });
      else document.execCommand('paste');
    } else if (cmd === 'selectAll') { var range = document.createRange(); range.selectNodeContents(editor); var s = window.getSelection(); s.removeAllRanges(); s.addRange(range); }
    else if (cmd === 'newDoc') { inst.doc = emptyDoc(); inst.editor.innerHTML = ''; inst.editor.appendChild(renderDoc(inst.doc)); }
    else if (cmd === 'openFile') { openFile(inst.name); return true; }
    else if (/^save(TXT|RTF|DOC|DOCX|ODT|PDF|HTML|Markdown|JSON|XML)$/.test(cmd)) {
      var saveFormat = cmd.slice(4).toLowerCase() === 'markdown' ? 'markdown' : cmd.slice(4).toLowerCase();
      var saveResult = saveAs(inst.name, saveFormat);
      if (saveResult[1] !== null) window.alert(saveResult[1]);
      return true;
    }
    else if (cmd === 'toggleStatus') { var status = inst.root.querySelector('.sui-rich-status'); if (status) status.hidden = !status.hidden; }
    else if (cmd === 'toggleContext') inst.config.context = !inst.config.context;
    else if (cmd === 'templateArticle') return setTemplate(inst.name, 'article')[0];
    else if (cmd === 'templateLetter') return setTemplate(inst.name, 'letter')[0];
    else if (cmd === 'templateReport') return setTemplate(inst.name, 'report')[0];
    else if (cmd === 'templateCompact') return setTemplate(inst.name, 'compact')[0];
    else if (cmd === 'pluginInfo') window.alert('Plugins: ' + (pluginOrder.length ? pluginOrder.join(', ') : 'none'));
    else if (cmd === 'importJSON' || cmd === 'importHTML' || cmd === 'importMarkdown' || cmd === 'importXML') {
      var fmt = cmd.slice(6).toLowerCase();
      var text = window.prompt('Paste ' + fmt + ' content', '');
      if (text != null) importFormat(nameOf(inst), text, fmt);
    } else if (/^export(JSON|HTML|Markdown|XML)$/.test(cmd)) {
      var format = cmd.slice(6).toLowerCase();
      var result = exportFormat(nameOf(inst), format);
      if (result[1] === null) { if (navigator.clipboard && navigator.clipboard.writeText) navigator.clipboard.writeText(result[0]); window.alert('Exported to clipboard'); }
      else window.alert(result[1]);
    } else return false;
    return true;
  }

  function nameOf(inst) { return inst.name; }

  function command(name, cmd, value) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    try {
      var handled = runBuiltin(inst, cmd, value);
      if (!handled) {
        var pluginCmd = commands[cmd];
        if (!pluginCmd) return [false, 'unknown richtext command: ' + cmd];
        var r = typeof pluginCmd.run === 'function' ? pluginCmd.run(inst, value) : runBuiltin(inst, pluginCmd.command, value);
        if (r === false) return [false, 'command failed: ' + cmd];
      }
      if (cmd !== 'tableSelect' && cmd !== 'tableDock' && cmd !== 'copy') commit(inst, 'change');
      return [true, null];
    } catch (e) { return [false, String(e && e.message ? e.message : e)]; }
  }

  function format(name, fmt, value) { return command(name, fmt, value); }

  function getValue(name) {
    var inst = instances[name];
    if (!inst) return [null, 'richtext instance not found: ' + name];
    sync(inst);
    return [JSON.stringify(inst.doc), null];
  }
  function setValue(name, value) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var doc = typeof value === 'string' ? parseJSON(value, null) : value;
    if (!doc || typeof doc !== 'object') return [false, 'richtext document must be JSON'];
    inst.doc = doc;
    inst.editor.innerHTML = '';
    inst.editor.appendChild(renderDoc(inst.doc));
    remember(inst);
    updateStats(inst);
    notify(inst, 'change');
    return [true, null];
  }

  function insertTextValue(name, text) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    inst.editor.focus();
    try { document.execCommand('insertText', false, String(text == null ? '' : text)); }
    catch (e) { insertNode(inst, document.createTextNode(String(text == null ? '' : text))); }
    commit(inst);
    return [true, null];
  }
  function insertImageValue(name, src, alt) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var url = safeUrl(src, true);
    if (!url) return [false, 'invalid image URL'];
    var img = el('img'); img.setAttribute('src', url); img.setAttribute('alt', String(alt == null ? '' : alt));
    insertNode(inst, img); commit(inst);
    return [true, null];
  }
  function insertLinkValue(name, href, text) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var url = safeUrl(href, false);
    if (!url) return [false, 'invalid link URL'];
    var sel = window.getSelection();
    if (sel && !sel.isCollapsed && inst.editor.contains(sel.anchorNode)) document.execCommand('createLink', false, url);
    else {
      var a = el('a', '', String(text == null || text === '' ? url : text));
      a.setAttribute('href', url); insertNode(inst, a);
    }
    commit(inst); return [true, null];
  }
  function insertTableValue(name, rows, cols) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    insertTable(inst, rows, cols); commit(inst); return [true, null];
  }

  function exportHTML(inst) {
    var div = el('div');
    var kids = inst.doc.children || [];
    for (var i = 0; i < kids.length; i++) { var b = blockDom(kids[i], 0); if (b) div.appendChild(b); }
    return div.innerHTML;
  }
  function exportMarkdown(inst) {
    var out = [];
    function inlineText(nodes) {
      var s = '';
      for (var i = 0; i < (nodes || []).length; i++) {
        var n = nodes[i];
        if (n.type === 'break') { s += '  \n'; continue; }
        if (n.type === 'image') { s += '![' + (n.alt || '') + '](' + (n.src || '') + ')'; continue; }
        if (n.type !== 'text') continue;
        var t = String(n.text || '').replace(/([\\\x60*_{}\[\]()#+\-.!>])/g, '\\$1');
        var marks = n.marks || [];
        for (var m = 0; m < marks.length; m++) {
          var mark = marks[m];
          if (mark.type === 'bold') t = '**' + t + '**';
          else if (mark.type === 'italic') t = '*' + t + '*';
          else if (mark.type === 'strike') t = '~~' + t + '~~';
          else if (mark.type === 'code') t = String.fromCharCode(96) + t + String.fromCharCode(96);
          else if (mark.type === 'link') t = '[' + t + '](' + (mark.href || '') + ')';
        }
        s += t;
      }
      return s;
    }
    function block(n, indent) {
      var pad = indent || '';
      if (!n) return;
      if (n.type === 'paragraph') out.push(pad + inlineText(n.children));
      else if (n.type === 'heading') out.push(pad + new Array(Math.max(1, Math.min(6, n.level || 1)) + 1).join('#') + ' ' + inlineText(n.children));
      else if (n.type === 'blockquote') { out.push(pad + '> ' + inlineText(n.children)); }
      else if (n.type === 'code_block') out.push(pad + '\x60\x60\x60\n' + String(n.text || '') + '\n\x60\x60\x60');
      else if (n.type === 'hr') out.push(pad + '---');
      else if (n.type === 'image') out.push(pad + '![' + (n.alt || '') + '](' + (n.src || '') + ')');
      else if (n.type === 'bullet_list' || n.type === 'ordered_list') {
        var items = n.children || [];
        for (var i = 0; i < items.length; i++) {
          var prefix = n.type === 'ordered_list' ? (i + 1) + '. ' : '- ';
          var inner = items[i].children || [];
          out.push(pad + prefix + (inner[0] ? inlineText(inner[0].children) : ''));
          for (var j = 1; j < inner.length; j++) block(inner[j], pad + '  ');
        }
      } else if (n.type === 'table') {
        var rows = n.rows || [];
        for (var r = 0; r < rows.length; r++) {
          var cells = [];
          for (var c = 0; c < (rows[r].cells || []).length; c++) cells.push(inlineText(rows[r].cells[c].children));
          out.push(pad + '| ' + cells.join(' | ') + ' |');
          if (r === 0) out.push(pad + '|' + cells.map(function () { return ' --- |'; }).join(''));
        }
      }
    }
    for (var i = 0; i < (inst.doc.children || []).length; i++) block(inst.doc.children[i], '');
    return out.join('\n\n') + '\n';
  }
  function xmlEscape(s) { return String(s == null ? '' : s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;'); }
  function exportXML(inst) {
    var out = ['<?xml version="1.0" encoding="UTF-8"?><document>'];
    function inline(nodes) {
      var s = '';
      for (var i = 0; i < (nodes || []).length; i++) {
        var n = nodes[i];
        if (n.type === 'text') {
          var marks = (n.marks || []).map(function (m) { return m.type === 'link' ? 'link:' + m.href : m.type; }).join(',');
          s += '<text' + (marks ? ' marks="' + xmlEscape(marks) + '"' : '') + '>' + xmlEscape(n.text) + '</text>';
        } else if (n.type === 'break') s += '<break/>';
        else if (n.type === 'image') s += '<image src="' + xmlEscape(n.src || '') + '" alt="' + xmlEscape(n.alt || '') + '"/>';
      }
      return s;
    }
    function block(n) {
      if (!n) return;
      if (n.type === 'paragraph') out.push('<paragraph>' + inline(n.children) + '</paragraph>');
      else if (n.type === 'heading') out.push('<heading level="' + (Number(n.level) || 1) + '">' + inline(n.children) + '</heading>');
      else if (n.type === 'blockquote') out.push('<blockquote>' + (n.children || []).map(block).join('') + '</blockquote>');
      else if (n.type === 'code_block') out.push('<code_block><![CDATA[' + String(n.text || '').replace(/\]\]>/g, ']]]]><![CDATA[>') + ']]></code_block>');
      else if (n.type === 'hr') out.push('<hr/>');
      else if (n.type === 'image') out.push('<image src="' + xmlEscape(n.src || '') + '" alt="' + xmlEscape(n.alt || '') + '"/>');
      else if (n.type === 'bullet_list' || n.type === 'ordered_list') out.push('<' + (n.type === 'bullet_list' ? 'bullet_list' : 'ordered_list') + '>' + (n.children || []).map(block).join('') + '</' + (n.type === 'bullet_list' ? 'bullet_list' : 'ordered_list') + '>');
      else if (n.type === 'list_item') out.push('<list_item>' + (n.children || []).map(block).join('') + '</list_item>');
      else if (n.type === 'table') {
        out.push('<table>');
        for (var r = 0; r < (n.rows || []).length; r++) {
          out.push('<table_row>');
          for (var c = 0; c < (n.rows[r].cells || []).length; c++) {
            var cell = n.rows[r].cells[c];
            out.push('<' + (cell.header ? 'header_cell' : 'table_cell') + '>' + inline(cell.children) + '</' + (cell.header ? 'header_cell' : 'table_cell') + '>');
          }
          out.push('</table_row>');
        }
        out.push('</table>');
      }
    }
    for (var i = 0; i < (inst.doc.children || []).length; i++) block(inst.doc.children[i]);
    out.push('</document>');
    return out.join('');
  }

  function importHTML(text) {
    var parser = new DOMParser();
    var dom = parser.parseFromString(String(text || ''), 'text/html');
    return docFromDom(dom.body);
  }
  function xmlInlineNodes(parent) {
    var out = [];
    for (var i = 0; i < parent.childNodes.length; i++) {
      var n = parent.childNodes[i];
      if (n.nodeType === 3) { out.push(markText(n.nodeValue || '', [])); continue; }
      if (n.nodeType !== 1) continue;
      var tag = n.localName || n.nodeName;
      if (tag === 'text') {
        var marks = (n.getAttribute('marks') || '').split(',').filter(Boolean).map(function (m) {
          if (m.indexOf('link:') === 0) return { type: 'link', href: safeUrl(m.slice(5), false) };
          return { type: m };
        });
        out.push(markText(n.textContent || '', marks));
      } else if (tag === 'break') out.push({ type: 'break' });
      else if (tag === 'image') out.push({ type: 'image', src: safeUrl(n.getAttribute('src'), true), alt: n.getAttribute('alt') || '' });
      else out.push(markText(n.textContent || '', []));
    }
    return out;
  }
  function xmlBlocks(parent) {
    var out = [];
    for (var i = 0; i < parent.childNodes.length; i++) {
      var n = parent.childNodes[i];
      if (n.nodeType !== 1) continue;
      var tag = n.localName || n.nodeName;
      if (tag === 'paragraph') out.push({ type: 'paragraph', children: xmlInlineNodes(n) });
      else if (tag === 'heading') out.push({ type: 'heading', level: Number(n.getAttribute('level') || 1), children: xmlInlineNodes(n) });
      else if (tag === 'code_block') out.push({ type: 'code_block', text: n.textContent || '' });
      else if (tag === 'hr') out.push({ type: 'hr' });
      else if (tag === 'image') out.push({ type: 'image', src: safeUrl(n.getAttribute('src'), true), alt: n.getAttribute('alt') || '' });
      else if (tag === 'blockquote') out.push({ type: 'blockquote', children: xmlBlocks(n) });
      else if (tag === 'bullet_list' || tag === 'ordered_list') {
        var list = { type: tag, children: [] };
        for (var c = 0; c < n.childNodes.length; c++) if (n.childNodes[c].nodeType === 1 && n.childNodes[c].localName === 'list_item') list.children.push({ type: 'list_item', children: xmlBlocks(n.childNodes[c]) });
        out.push(list);
      } else if (tag === 'table') {
        var table = { type: 'table', rows: [] };
        for (var r = 0; r < n.childNodes.length; r++) {
          if (n.childNodes[r].nodeType !== 1 || n.childNodes[r].localName !== 'table_row') continue;
          var row = { type: 'table_row', cells: [] };
          for (var ci = 0; ci < n.childNodes[r].childNodes.length; ci++) {
            var cell = n.childNodes[r].childNodes[ci];
            if (cell.nodeType !== 1 || (cell.localName !== 'table_cell' && cell.localName !== 'header_cell')) continue;
            row.cells.push({ type: 'table_cell', header: cell.localName === 'header_cell', colspan: 1, rowspan: 1, children: xmlInlineNodes(cell) });
          }
          table.rows.push(row);
        }
        out.push(table);
      } else out.push({ type: 'paragraph', children: xmlInlineNodes(n) });
    }
    return out;
  }
  function importXML(text) {
    var src = String(text || '');
    if (/<!\s*(DOCTYPE|ENTITY)\b/i.test(src)) throw new Error('DOCTYPE and ENTITY declarations are not allowed');
    var parser = new DOMParser();
    var dom = parser.parseFromString(src, 'application/xml');
    var err = dom.getElementsByTagName('parsererror');
    if (err && err.length) throw new Error(err[0].textContent || 'invalid XML');
    var root = dom.documentElement;
    if (!root) throw new Error('XML has no root element');
    var blocks = (root.localName || root.nodeName) === 'document' ? xmlBlocks(root) : docFromDom(root).children;
    return { type: 'doc', children: blocks.length ? blocks : [{ type: 'paragraph', children: [] }] };
  }
  function markText(text, marks) { return { type: 'text', text: text, marks: marks || [] }; }
  function parseInlineMarkdown(line) {
    var out = [], re = /(\*\*[^*]+\*\*|__[^_]+__|\*[^*]+\*|_[^_]+_|~~[^~]+~~|\x60[^\x60]+\x60|\[[^\]]+\]\([^)]+\)|!\[[^\]]*\]\([^)]+\))/g, last = 0, m;
    while ((m = re.exec(line))) {
      if (m.index > last) out.push(markText(line.slice(last, m.index), []));
      var token = m[0];
      if (/^\*\*/.test(token)) out.push(markText(token.slice(2, -2), [{ type: 'bold' }]));
      else if (/^__/.test(token)) out.push(markText(token.slice(2, -2), [{ type: 'bold' }]));
      else if (/^\*/.test(token)) out.push(markText(token.slice(1, -1), [{ type: 'italic' }]));
      else if (/^_/.test(token)) out.push(markText(token.slice(1, -1), [{ type: 'italic' }]));
      else if (/^~~/.test(token)) out.push(markText(token.slice(2, -2), [{ type: 'strike' }]));
      else if (/^\x60/.test(token)) out.push(markText(token.slice(1, -1), [{ type: 'code' }]));
      else if (/^!\[/.test(token)) {
        var im = /^!\[([^\]]*)\]\(([^)]+)\)$/.exec(token);
        if (im) out.push({ type: 'image', src: safeUrl(im[2], true), alt: im[1] });
      } else {
        var lm = /^\[([^\]]+)\]\(([^)]+)\)$/.exec(token);
        if (lm) out.push(markText(lm[1], [{ type: 'link', href: safeUrl(lm[2], false) }]));
      }
      last = m.index + token.length;
    }
    if (last < line.length) out.push(markText(line.slice(last), []));
    return out.length ? out : [markText('', [])];
  }
  function importMarkdown(text) {
    var lines = String(text || '').replace(/\r\n/g, '\n').split('\n'), blocks = [], paragraph = [], list = null, inCode = false, code = [];
    function flushParagraph() { if (paragraph.length) { blocks.push({ type: 'paragraph', children: parseInlineMarkdown(paragraph.join(' ')) }); paragraph = []; } }
    function flushList() { if (list) { blocks.push(list); list = null; } }
    for (var i = 0; i < lines.length; i++) {
      var line = lines[i];
      if (/^\x60\x60\x60/.test(line)) {
        flushParagraph(); flushList();
        if (inCode) { blocks.push({ type: 'code_block', text: code.join('\n') }); code = []; inCode = false; }
        else inCode = true;
        continue;
      }
      if (inCode) { code.push(line); continue; }
      if (!line.trim()) { flushParagraph(); flushList(); continue; }
      var h = /^(#{1,6})\s+(.*)$/.exec(line);
      if (h) { flushParagraph(); flushList(); blocks.push({ type: 'heading', level: h[1].length, children: parseInlineMarkdown(h[2]) }); continue; }
      if (/^---+$/.test(line.trim())) { flushParagraph(); flushList(); blocks.push({ type: 'hr' }); continue; }
      var ul = /^\s*[-*+]\s+(.*)$/.exec(line), ol = /^\s*\d+[.)]\s+(.*)$/.exec(line);
      if (ul || ol) {
        flushParagraph();
        var listType = ul ? 'bullet_list' : 'ordered_list';
        if (!list || list.type !== listType) { flushList(); list = { type: listType, children: [] }; }
        list.children.push({ type: 'list_item', children: [{ type: 'paragraph', children: parseInlineMarkdown((ul || ol)[1]) }] });
        continue;
      }
      var bq = /^>\s?(.*)$/.exec(line);
      if (bq) { flushParagraph(); flushList(); blocks.push({ type: 'blockquote', children: [{ type: 'paragraph', children: parseInlineMarkdown(bq[1]) }] }); continue; }
      if (/^\|/.test(line) && i + 1 < lines.length && /^\|?\s*:?-+/.test(lines[i + 1])) {
        flushParagraph(); flushList();
        var headers = line.split('|').slice(1, -1);
        var rows = [];
        i += 2;
        var cells = headers.map(function (h) { return { type: 'table_cell', header: true, colspan: 1, rowspan: 1, children: parseInlineMarkdown(h.trim()) }; });
        rows.push({ type: 'table_row', cells: cells });
        while (i < lines.length && /^\|/.test(lines[i])) {
          var parts = lines[i].split('|').slice(1, -1);
          if (parts.length) rows.push({ type: 'table_row', cells: parts.map(function (c) { return { type: 'table_cell', header: false, colspan: 1, rowspan: 1, children: parseInlineMarkdown(c.trim()) }; }) });
          i++;
        }
        i--;
        blocks.push({ type: 'table', header: true, rows: rows });
        continue;
      }
      paragraph.push(line);
    }
    flushParagraph(); flushList();
    if (inCode && code.length) blocks.push({ type: 'code_block', text: code.join('\n') });
    return { type: 'doc', children: blocks.length ? blocks : [{ type: 'paragraph', children: [] }] };
  }

  function exportFormat(name, format) {
    var inst = instances[name];
    if (!inst) return [null, 'richtext instance not found: ' + name];
    sync(inst);
    var fmt = String(format || 'json').toLowerCase();
    try {
      if (fmt === 'json') return [JSON.stringify(inst.doc), null];
      if (fmt === 'html') return [exportHTML(inst), null];
      if (fmt === 'markdown' || fmt === 'md') return [exportMarkdown(inst), null];
      if (fmt === 'xml') return [exportXML(inst), null];
      return [null, 'unsupported richtext format: ' + format];
    } catch (e) { return [null, String(e && e.message ? e.message : e)]; }
  }
  function importFormatTo(name, text, format) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var fmt = String(format || 'json').toLowerCase();
    try {
      var doc;
      if (fmt === 'json') doc = parseJSON(text, null);
      else if (fmt === 'html') doc = importHTML(text);
      else if (fmt === 'markdown' || fmt === 'md') doc = importMarkdown(text);
      else if (fmt === 'xml') doc = importXML(text);
      else return [false, 'unsupported richtext format: ' + format];
      if (!doc || typeof doc !== 'object') return [false, 'invalid richtext document'];
      inst.doc = doc;
      inst.editor.innerHTML = '';
      inst.editor.appendChild(renderDoc(inst.doc));
      remember(inst);
      updateStats(inst);
      notify(inst, 'change');
      return [true, null];
    } catch (e) { return [false, String(e && e.message ? e.message : e)]; }
  }

  function registerPlugin(name, plugin) {
    if (typeof plugin === 'string') plugin = parseJSON(plugin, null);
    if (!name || !plugin || typeof plugin !== 'object') return [false, 'invalid plugin'];
    plugins[String(name)] = plugin;
    if (pluginOrder.indexOf(String(name)) < 0) pluginOrder.push(String(name));
    var cmds = plugin.commands || {};
    for (var key in cmds) if (hasOwn(cmds, key)) {
      var spec = cmds[key];
      commands[key] = typeof spec === 'function' ? { run: spec } : { title: spec.title || key, icon: spec.icon || '', command: spec.command || key, run: spec.run };
    }
    for (var n in instances) if (hasOwn(instances, n)) buildToolbar(instances[n]);
    return [true, null];
  }
  function pluginNames() { return pluginOrder.slice(); }

  function updateStats(inst) {
    var stats = docStats(inst.doc);
    var status = inst.root && inst.root.querySelector('.sui-rich-stats');
    if (status) status.textContent = stats.words + ' words \u00b7 ' + stats.chars + ' chars \u00b7 ' + stats.paragraphs + ' paragraphs';
    var pos = inst.root && inst.root.querySelector('.sui-rich-position');
    if (pos) {
      var info = selectionInfo(inst);
      pos.textContent = 'Cursor ' + info.start + (info.end !== info.start ? '\u2013' + info.end : '') + (info.block ? ' \u00b7 ' + info.block : '');
    }
    var saved = inst.root && inst.root.querySelector('.sui-rich-saved');
    if (saved) saved.textContent = inst.dirty ? 'Modified' : 'Ready';
    updateTableBar(inst);
    updateProperties(inst);
  }
  function updateProperties(inst) {
    var panel = inst.root && inst.root.querySelector('.sui-rich-props');
    if (!panel || panel.hidden) return;
    var block = currentBlock(inst);
    var label = block ? blockLabel(block.tagName.toLowerCase() === 'h1' ? 'heading' : 'paragraph', { level: block.tagName.toLowerCase() === 'h1' ? 1 : 2 }) : 'Document';
    panel.innerHTML = '';
    var title = el('div', 'sui-rich-props-title', label);
    panel.appendChild(title);
    var align = el('select', 'sui-rich-prop');
    align.setAttribute('data-rich-prop', 'align');
    [['', 'Align'], ['left', 'Left'], ['center', 'Center'], ['right', 'Right'], ['justify', 'Justify']].forEach(function (o) { var op = el('option', '', o[1]); op.value = o[0]; align.appendChild(op); });
    panel.appendChild(align);
    var lh = el('input', 'sui-rich-prop'); lh.setAttribute('data-rich-prop', 'lineHeight'); lh.placeholder = 'Line height (e.g. 1.6)'; panel.appendChild(lh);
    var link = closestIn(inst, 'a');
    if (link) { var href = el('input', 'sui-rich-prop'); href.setAttribute('data-rich-prop', 'href'); href.value = link.getAttribute('href') || ''; panel.appendChild(href); }
    var img = closestIn(inst, 'img');
    if (img) { var alt = el('input', 'sui-rich-prop'); alt.setAttribute('data-rich-prop', 'alt'); alt.value = img.getAttribute('alt') || ''; panel.appendChild(alt); }
    var table = closestIn(inst, 'table');
    if (table) {
      [['tableAddRow', 'Add row'], ['tableDeleteRow', 'Delete row'], ['tableAddColumn', 'Add column'], ['tableDeleteColumn', 'Delete column']].forEach(function (item) {
        var b = el('button', 'sui-rich-prop-btn', item[1]); b.type = 'button'; b.setAttribute('data-rich-cmd', item[0]); panel.appendChild(b);
      });
    }
  }
  function bindProperties(inst) {
    var panel = inst.root.querySelector('.sui-rich-props');
    if (!panel) return;
    panel.addEventListener('change', function (e) {
      var field = e.target && e.target.getAttribute ? e.target.getAttribute('data-rich-prop') : '';
      if (!field) return;
      var value = e.target.value;
      if (field === 'align') { var b = currentBlock(inst); if (b) b.style.textAlign = value; }
      else if (field === 'lineHeight') { var b2 = currentBlock(inst); if (b2) b2.style.lineHeight = cssLength(value, ''); }
      else if (field === 'href') { var a = closestIn(inst, 'a'); if (a && safeUrl(value, false)) a.setAttribute('href', safeUrl(value, false)); }
      else if (field === 'alt') { var im = closestIn(inst, 'img'); if (im) im.setAttribute('alt', value); }
      commit(inst);
    });
    panel.addEventListener('click', function (e) {
      var cmd = e.target && e.target.getAttribute ? e.target.getAttribute('data-rich-cmd') : '';
      if (cmd) { command(inst.name, cmd); }
    });
  }

  function button(inst, cmd, label, title) {
    var b = el('button', 'sui-rich-btn');
    b.type = 'button';
    b.setAttribute('data-rich-cmd', cmd);
    b.setAttribute('title', title || label);
    b.setAttribute('aria-label', title || label);
    if (ICONS[cmd]) b.innerHTML = ICONS[cmd]; else b.textContent = label;
    return b;
  }
  function select(inst, cmd, options, label) {
    var s = el('select', 'sui-rich-select');
    s.setAttribute('data-rich-cmd', cmd);
    s.setAttribute('aria-label', label);
    for (var i = 0; i < options.length; i++) { var o = el('option', '', options[i][1]); o.value = options[i][0]; s.appendChild(o); }
    return s;
  }
  function buildToolbar(inst) {
    var bar = inst.root.querySelector('.sui-rich-toolbar');
    if (!bar) return;
    bar.innerHTML = '';
    var items = inst.config.toolbar || defaultToolbar;
    var groups = String(items).split('|');
    for (var g = 0; g < groups.length; g++) {
      if (!groups[g].trim()) continue;
      var group = el('div', 'sui-rich-group');
      var tokens = groups[g].split(',');
      for (var t = 0; t < tokens.length; t++) {
        var token = tokens[t].trim();
        if (!token) continue;
        if (token === 'styleSelect') group.appendChild(select(inst, 'paragraph', [['', 'Normal'], ['h1', 'Heading 1'], ['h2', 'Heading 2'], ['h3', 'Heading 3'], ['blockquote', 'Quote'], ['pre', 'Code block']], 'Block style'));
        else if (token === 'fontFamily') group.appendChild(select(inst, 'fontFamily', [['', 'Font'], ['Arial, sans-serif', 'Arial'], ['Georgia, serif', 'Georgia'], ['Times New Roman, serif', 'Times'], ['Courier New, monospace', 'Courier'], ['system-ui, sans-serif', 'System']], 'Font family'));
        else if (token === 'fontSize') group.appendChild(select(inst, 'fontSize', [['', 'Size'], ['1', 'Small'], ['3', 'Normal'], ['5', 'Large'], ['7', 'Huge']], 'Font size'));
        else if (token === 'lineHeight') group.appendChild(select(inst, 'lineHeight', [['', 'Line height'], ['1.2', '1.2'], ['1.5', '1.5'], ['1.8', '1.8'], ['2', '2.0']], 'Line height'));
        else if (token === 'spaceBefore') group.appendChild(select(inst, 'spaceBefore', [['', 'Space before'], ['0', '0'], ['0.5em', '0.5em'], ['1em', '1em'], ['1.5em', '1.5em']], 'Space before'));
        else if (token === 'spaceAfter') group.appendChild(select(inst, 'spaceAfter', [['', 'Space after'], ['0', '0'], ['0.5em', '0.5em'], ['1em', '1em'], ['1.5em', '1.5em']], 'Space after'));
        else if (token === 'textColor' || token === 'highlight') {
          var color = el('input', 'sui-rich-color');
          color.type = 'color';
          color.setAttribute('data-rich-cmd', token);
          color.value = token === 'textColor' ? '#e6e6ef' : '#ffe066';
          color.title = token === 'textColor' ? 'Text color' : 'Highlight color';
          group.appendChild(color);
        } else group.appendChild(button(inst, token, token, token));
      }
      bar.appendChild(group);
    }
    var registered = [];
    for (var p = 0; p < pluginOrder.length; p++) {
      var pl = plugins[pluginOrder[p]];
      if (pl && Array.isArray(pl.toolbar)) for (var ti = 0; ti < pl.toolbar.length; ti++) registered.push(pl.toolbar[ti]);
    }
    if (registered.length) {
      var pg = el('div', 'sui-rich-group');
      for (var r = 0; r < registered.length; r++) pg.appendChild(button(inst, registered[r].id, registered[r].label || registered[r].id, registered[r].title || registered[r].label || registered[r].id));
      bar.appendChild(pg);
    }
  }
  function buildMenu(inst) {
    var bar = inst.root.querySelector('.sui-rich-menu');
    if (!bar) return;
    bar.innerHTML = '';
    Object.keys(MENUS).forEach(function (key) {
      var wrap = el('div', 'sui-rich-menu-wrap');
      var b = el('button', 'sui-rich-menu-btn', MENU_LABELS[key]);
      b.type = 'button'; b.setAttribute('data-rich-menu', key); b.setAttribute('aria-haspopup', 'true');
      var panel = el('div', 'sui-rich-menu-panel');
      panel.hidden = true;
      MENUS[key].forEach(function (spec) {
        if (spec && typeof spec === 'object' && Array.isArray(spec.items)) {
          var subwrap = el('div', 'sui-rich-menu-subwrap');
          var parent = el('button', 'sui-rich-menu-item sui-rich-menu-parent', spec.label + ' \u25b8');
          parent.type = 'button';
          parent.setAttribute('data-rich-submenu', spec.label);
          var sub = el('div', 'sui-rich-menu-panel sui-rich-submenu');
          sub.hidden = true;
          spec.items.forEach(function (subspec) {
            var sp = subspec.split('|');
            var subitem = el('button', 'sui-rich-menu-item', sp[0]);
            subitem.type = 'button'; subitem.setAttribute('data-rich-cmd', sp[1]);
            sub.appendChild(subitem);
          });
          subwrap.appendChild(parent); subwrap.appendChild(sub);
          panel.appendChild(subwrap);
          return;
        }
        var parts = spec.split('|');
        var item = el('button', 'sui-rich-menu-item', parts[0]);
        item.type = 'button'; item.setAttribute('data-rich-cmd', parts[1]);
        panel.appendChild(item);
      });
      wrap.appendChild(b); wrap.appendChild(panel); bar.appendChild(wrap);
    });
  }
  function buildTableBar(inst) {
    var bar = inst.root.querySelector('.sui-rich-tablebar');
    if (!bar) return;
    bar.innerHTML = '';
    var handle = el('span', 'sui-rich-tablebar-handle', '\u28ff');
    handle.setAttribute('data-rich-table-drag', '1');
    handle.setAttribute('title', 'Drag to reposition; double-click to dock');
    bar.appendChild(handle);
    var groups = [
      [['tableSelect', 'T', 'Select whole table']],
      [['tableRowAbove', '+R', 'Add row above'], ['tableRowBelow', '+R\u2193', 'Add row below'], ['tableRowDelete', '-R', 'Delete row']],
      [['tableColLeft', '+C', 'Add column left'], ['tableColRight', '+C\u2192', 'Add column right'], ['tableColDelete', '-C', 'Delete column']],
      [['tableHeader', 'H', 'Toggle header row'], ['tableMerge', 'M', 'Merge selected cells'], ['tableSplit', 'S', 'Split cell']],
      [['tableAlignLeft', 'L', 'Align cell left'], ['tableAlignCenter', 'C', 'Align cell center'], ['tableAlignRight', 'R', 'Align cell right'], ['tableDelete', '\u00d7', 'Delete table'], ['tableDock', 'D', 'Dock table toolbar']]
    ];
    for (var gi = 0; gi < groups.length; gi++) {
      var group = el('div', 'sui-rich-group');
      for (var bi = 0; bi < groups[gi].length; bi++) {
        var spec = groups[gi][bi];
        group.appendChild(button(inst, spec[0], spec[1], spec[2]));
      }
      bar.appendChild(group);
    }
  }
  function updateTableBar(inst) {
    var bar = inst.root && inst.root.querySelector('.sui-rich-tablebar');
    if (!bar) return;
    var table = closestIn(inst, 'table');
    if (!table && inst.selectedTable && inst.selectedTable.isConnected) table = inst.selectedTable;
    if (!table) { bar.hidden = true; try { clearCellSelection(inst); } catch (e) {} return; }
    bar.hidden = false;
  }
  function dockTableBar(inst) {
    var bar = inst.root && inst.root.querySelector('.sui-rich-tablebar');
    if (!bar) return;
    bar.classList.remove('floating');
    bar.style.left = '';
    bar.style.top = '';
    try { localStorage.removeItem('sui-rich-tablebar-' + inst.name); } catch (e) {}
  }
  function restoreTableBarPosition(inst) {
    var bar = inst.root && inst.root.querySelector('.sui-rich-tablebar');
    if (!bar) return;
    try {
      var saved = localStorage.getItem('sui-rich-tablebar-' + inst.name);
      if (saved) {
        var pos = parseJSON(saved, null);
        if (pos && isFinite(pos.x) && isFinite(pos.y)) {
          bar.classList.add('floating');
          bar.style.left = Math.max(0, Number(pos.x)) + 'px';
          bar.style.top = Math.max(0, Number(pos.y)) + 'px';
        }
      }
    } catch (e) {}
  }
  function bindTableBarDrag(inst) {
    var bar = inst.root && inst.root.querySelector('.sui-rich-tablebar');
    if (!bar) return;
    var handle = bar.querySelector('[data-rich-table-drag]');
    if (!handle || handle.__suiDragBound) return;
    handle.__suiDragBound = true;
    handle.addEventListener('dblclick', function (e) { e.preventDefault(); dockTableBar(inst); });
    handle.addEventListener('mousedown', function (e) {
      e.preventDefault(); e.stopPropagation();
      var rect = bar.getBoundingClientRect();
      var offsetX = e.clientX - rect.left, offsetY = e.clientY - rect.top;
      if (!bar.classList.contains('floating')) { bar.style.left = rect.left + 'px'; bar.style.top = rect.top + 'px'; }
      bar.classList.add('floating');
      function move(ev) {
        bar.style.left = Math.max(0, Math.min(window.innerWidth - 80, ev.clientX - offsetX)) + 'px';
        bar.style.top = Math.max(0, Math.min(window.innerHeight - 40, ev.clientY - offsetY)) + 'px';
      }
      function up() {
        document.removeEventListener('mousemove', move);
        document.removeEventListener('mouseup', up);
        try { localStorage.setItem('sui-rich-tablebar-' + inst.name, JSON.stringify({ x: parseFloat(bar.style.left) || 0, y: parseFloat(bar.style.top) || 0 })); } catch (err) {}
      }
      document.addEventListener('mousemove', move);
      document.addEventListener('mouseup', up);
    });
  }
  function closeMenus(inst) {
    var panels = inst.root.querySelectorAll('.sui-rich-menu-panel');
    for (var i = 0; i < panels.length; i++) panels[i].hidden = true;
  }
  function buildContext(inst) {
    var bar = inst.root.querySelector('.sui-rich-context');
    if (!bar) return;
    bar.innerHTML = '';
    ['bold', 'italic', 'underline', 'strike', 'link', 'highlight', 'clearFormat'].forEach(function (cmd) { bar.appendChild(button(inst, cmd, cmd, cmd)); });
  }
  function buildStatus(inst) {
    var status = inst.root.querySelector('.sui-rich-status');
    if (!status) return;
    status.innerHTML = '';
    var stats = el('span', 'sui-rich-stats'); status.appendChild(stats);
    var pos = el('span', 'sui-rich-position'); status.appendChild(pos);
    var saved = el('span', 'sui-rich-saved', 'Ready'); status.appendChild(saved);
    var zoomOut = el('button', 'sui-rich-zoom-btn', '\u2212'); zoomOut.type = 'button'; zoomOut.setAttribute('data-rich-zoom', 'out');
    var zoomLabel = el('span', 'sui-rich-zoom-label', '100%');
    var zoomIn = el('button', 'sui-rich-zoom-btn', '+'); zoomIn.type = 'button'; zoomIn.setAttribute('data-rich-zoom', 'in');
    status.appendChild(zoomOut); status.appendChild(zoomLabel); status.appendChild(zoomIn);
  }

  function bindRoot(inst) {
    var root = inst.root;
    if (root.__suiRichBound) return;
    root.__suiRichBound = true;
    root.addEventListener('mousedown', function (e) {
      if (!e.target.closest) return;
      var chrome = e.target.closest('.sui-rich-toolbar,.sui-rich-menu,.sui-rich-context,.sui-rich-tablebar,.sui-rich-status');
      if (!chrome) return;
      // Preserve the text selection before a select/color input takes focus.
      saveSelection(inst);
      if (!e.target.closest('select,input')) e.preventDefault();
    });
    root.addEventListener('click', function (e) {
      var target = e.target;
      if (!target || !target.closest) return;
      var submenuBtn = target.closest('[data-rich-submenu]');
      if (submenuBtn) {
        var subPanel = submenuBtn.parentNode.querySelector('.sui-rich-submenu');
        var wasSubOpen = subPanel && !subPanel.hidden;
        root.querySelectorAll('.sui-rich-submenu').forEach(function (p) { p.hidden = true; });
        if (subPanel && !wasSubOpen) subPanel.hidden = false;
        return;
      }
      var menuBtn = target.closest('[data-rich-menu]');
      if (menuBtn) {
        var panel = menuBtn.parentNode.querySelector('.sui-rich-menu-panel');
        var wasOpen = panel && !panel.hidden;
        root.querySelectorAll('.sui-rich-menu-panel').forEach(function (p) { p.hidden = true; });
        if (panel && !wasOpen) panel.hidden = false;
        return;
      }
      var zoom = target.closest('[data-rich-zoom]');
      if (zoom) {
        inst.zoom = Math.max(0.6, Math.min(2, (inst.zoom || 1) + (zoom.getAttribute('data-rich-zoom') === 'in' ? 0.1 : -0.1)));
        inst.editor.style.fontSize = (inst.baseFontSize * inst.zoom) + 'px';
        var label = root.querySelector('.sui-rich-zoom-label'); if (label) label.textContent = Math.round(inst.zoom * 100) + '%';
        return;
      }
      var cmdEl = target.closest('[data-rich-cmd]');
      if (!cmdEl) { closeMenus(inst); return; }
      var cmd = cmdEl.getAttribute('data-rich-cmd');
      var value = cmdEl.value || '';
      if (cmdEl.tagName === 'SELECT' || cmdEl.tagName === 'INPUT') return;
      if (cmdEl.closest('.sui-rich-menu-panel')) closeMenus(inst);
      command(inst.name, cmd, value);
    });
    root.addEventListener('change', function (e) {
      var target = e.target;
      if (!target || !target.getAttribute) return;
      var cmd = target.getAttribute('data-rich-cmd');
      if (!cmd) return;
      if (target.tagName === 'SELECT' || target.tagName === 'INPUT') command(inst.name, cmd, target.value);
    });
    root.addEventListener('input', function (e) {
      if (!inst.editor.contains(e.target)) return;
      clearTimeout(inst.inputTimer);
      inst.inputTimer = setTimeout(function () { commit(inst, 'input'); }, 250);
    });
    root.addEventListener('keydown', function (e) {
      if (!inst.editor.contains(e.target)) return;
      if (e.key === 'Enter' && !e.shiftKey && !e.ctrlKey && !e.metaKey) {
        e.preventDefault();
        splitBlock(inst);
        commit(inst, 'input');
        return;
      }
      if (e.key === 'Enter' && e.shiftKey) {
        e.preventDefault();
        try { document.execCommand('insertLineBreak'); } catch (err) { insertNode(inst, el('br')); }
        commit(inst, 'input');
        return;
      }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'z') { e.preventDefault(); if (e.shiftKey) redo(inst); else undo(inst); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'y') { e.preventDefault(); redo(inst); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') { e.preventDefault(); command(inst.name, 'bold'); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'i') { e.preventDefault(); command(inst.name, 'italic'); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'u') { e.preventDefault(); command(inst.name, 'underline'); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') { e.preventDefault(); commit(inst, 'change'); }
    });
    bindCellSelection(inst); bindTableResize(inst);
    if (!inst.menuOutsideBound) {
      inst.menuOutsideBound = true;
      function closeMenusOutside(ev) {
        if (!inst.root || !inst.root.contains(ev.target) || !ev.target.closest('.sui-rich-menu-wrap')) closeMenus(inst);
      }
      document.addEventListener('mousedown', closeMenusOutside);
      document.addEventListener('click', closeMenusOutside);
      document.addEventListener('keydown', function (ev) { if (ev.key === 'Escape') closeMenus(inst); });
    }
    inst.editor.addEventListener('blur', function () { setTimeout(function () { commit(inst, 'change'); }, 0); });
    inst.editor.addEventListener('copy', function (e) {
      if (!inst.selectedCells || inst.selectedCells.length < 2) return;
      var text = copyCellSelection(inst, false);
      if (!text) return;
      e.preventDefault();
      if (e.clipboardData) { e.clipboardData.setData('text/plain', text); }
    });
    inst.editor.addEventListener('cut', function (e) {
      if (!inst.selectedCells || inst.selectedCells.length < 2) return;
      var text = copyCellSelection(inst, true);
      if (!text) return;
      e.preventDefault();
      if (e.clipboardData) { e.clipboardData.setData('text/plain', text); }
      commit(inst, 'change');
    });
    inst.editor.addEventListener('paste', function (e) {
      var data = e.clipboardData || window.clipboardData;
      var textData = data && data.getData ? data.getData('text/plain') : '';
      var table = closestIn(inst, 'table') || inst.selectedTable;
      if (table && (inst.cellClipboard || (textData && (textData.indexOf('\t') >= 0 || textData.indexOf('\n') >= 0)))) {
        e.preventDefault();
        if (pasteCellMatrix(inst, textData)) { commit(inst, 'change'); return; }
      }
      var html = data && data.getData ? data.getData('text/html') : '';
      if (html) {
        e.preventDefault();
        var pasted = importHTML(html);
        var fragment = renderDoc(pasted);
        var sel = window.getSelection();
        if (sel && sel.rangeCount) {
          var range = sel.getRangeAt(0);
          range.deleteContents();
          range.insertNode(fragment);
          range.collapse(false);
          sel.removeAllRanges(); sel.addRange(range);
        } else inst.editor.appendChild(fragment);
        commit(inst, 'input');
      } else setTimeout(function () { commit(inst, 'input'); }, 0);
    });
    document.addEventListener('selectionchange', function () { if (document.activeElement === inst.editor || inst.root.contains(document.activeElement)) updateStats(inst); });
  }

  function init(root) {
    var name = attr(root, 'data-sui-rich-name') || 'richtext';
    var input = root.querySelector('.sui-rich-value');
    var editor = root.querySelector('.sui-rich-editor');
    if (!editor) return;
    var inst = instances[name];
    if (!inst) {
      var initial = input ? parseJSON(input.value, null) : null;
      inst = {
        name: name, root: root, editor: editor, config: {
          context: boolAttr(root, 'data-sui-rich-context', true),
          toolbar: attr(root, 'data-sui-rich-toolbar-items') || defaultToolbar,
          properties: boolAttr(root, 'data-sui-rich-properties', false)
        },
        doc: initial || emptyDoc(),
        history: [], index: -1, dirty: false, zoom: 1,
        baseFontSize: parseFloat(attr(root, 'data-sui-rich-font-size') || '16') || 16
      };
      inst.history.push(JSON.stringify(inst.doc));
      inst.index = 0;
      instances[name] = inst;
    } else {
      inst.root = root; inst.editor = editor;
    }
    if (!inst.editor.__suiRichRendered || inst.editor.__suiRichDoc !== JSON.stringify(inst.doc)) {
      inst.editor.innerHTML = '';
      inst.editor.appendChild(renderDoc(inst.doc));
      inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
    }
    inst.editor.__suiRichRendered = true;
    if (input) input.value = JSON.stringify(inst.doc);
    var height = attr(root, 'data-sui-rich-height');
    if (height) inst.editor.style.minHeight = cssLength(height, '320px');
    var initialTemplate = attr(root, 'data-sui-rich-template');
    if (initialTemplate) setTemplate(name, initialTemplate);
    var readOnly = boolAttr(root, 'data-sui-rich-readonly', false);
    inst.editor.setAttribute('contenteditable', readOnly ? 'false' : 'true');
    if (attr(root, 'data-sui-rich-placeholder')) inst.editor.setAttribute('data-placeholder', attr(root, 'data-sui-rich-placeholder'));
    buildToolbar(inst); buildMenu(inst); buildContext(inst); buildTableBar(inst); restoreTableBarPosition(inst); bindTableBarDrag(inst); buildStatus(inst);
    var toolbarEl = root.querySelector('.sui-rich-toolbar');
    if (toolbarEl && !boolAttr(root, 'data-sui-rich-toolbar', true)) toolbarEl.hidden = true;
    var menuEl = root.querySelector('.sui-rich-menu');
    if (menuEl && !boolAttr(root, 'data-sui-rich-menu', true)) menuEl.hidden = true;
    var statusEl = root.querySelector('.sui-rich-status');
    if (statusEl && !boolAttr(root, 'data-sui-rich-statusbar', true)) statusEl.hidden = true;
    bindRoot(inst); bindProperties(inst);
    updateStats(inst);
    restoreSelection(inst);
    var ready = attr(root, 'data-sui-rich-onready');
    if (ready && typeof window[ready] === 'function') window[ready](name, JSON.stringify(inst.doc));
  }

  function afterRender() {
    var roots = document.querySelectorAll('[data-sui-rich]');
    for (var i = 0; i < roots.length; i++) init(roots[i]);
  }
  function destroy(name) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    if (inst.root && inst.root.parentNode) inst.root.__suiRich = null;
    delete instances[name];
    return [true, null];
  }
  function status(name) {
    var inst = instances[name];
    if (!inst) return [null, 'richtext instance not found: ' + name];
    sync(inst);
    var stats = docStats(inst.doc), info = selectionInfo(inst);
    return [JSON.stringify({ words: stats.words, chars: stats.chars, paragraphs: stats.paragraphs, cursor: info.start, selection: info.end, block: info.block, modified: inst.dirty, zoom: inst.zoom }), null];
  }
  function selection(name) {
    var inst = instances[name];
    if (!inst) return [null, 'richtext instance not found: ' + name];
    return [JSON.stringify(selectionInfo(inst)), null];
  }
  function setTemplate(name, template) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var allowed = ['article', 'letter', 'report', 'compact'];
    var next = String(template || '').toLowerCase();
    if (allowed.indexOf(next) < 0) return [false, 'unknown template: ' + template];
    inst.template = next;
    inst.root.setAttribute('data-sui-rich-template', next);
    return [true, null];
  }
  function setToolbar(name, items) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    inst.config.toolbar = String(items || '');
    buildToolbar(inst);
    return [true, null];
  }

  document.addEventListener('selectionchange', function () {
    for (var name in instances) if (hasOwn(instances, name)) {
      var inst = instances[name];
      if (inst.config.context && inst.root.contains(document.activeElement)) {
        var info = selectionInfo(inst);
        var bar = inst.root.querySelector('.sui-rich-context');
        if (bar && info.inside && !info.collapsed) {
          var sel = window.getSelection();
          var rect = sel.getRangeAt(0).getBoundingClientRect();
          bar.hidden = false;
          bar.style.left = Math.max(8, rect.left) + 'px';
          bar.style.top = Math.max(8, rect.top - 44) + 'px';
        } else if (bar) bar.hidden = true;
        var selectionHandler = attr(inst.root, 'data-sui-rich-onselection');
        if (selectionHandler && typeof window[selectionHandler] === 'function') window[selectionHandler](JSON.stringify(info));
        updateProperties(inst);
      }
    }
  });

  // -- Serialization: plain text, RTF, DOCX, ODT and PDF ------------------
  function bytesFromText(text) { return Array.prototype.slice.call(new TextEncoder().encode(String(text))); }
  function utf8Bytes(text) { return new TextEncoder().encode(String(text)); }
  function concatBytes(parts) {
    var total = 0, i;
    for (i = 0; i < parts.length; i++) total += parts[i].length;
    var out = new Uint8Array(total), at = 0;
    for (i = 0; i < parts.length; i++) { out.set(parts[i], at); at += parts[i].length; }
    return out;
  }
  function u16(value) { return [value & 255, (value >>> 8) & 255]; }
  function u32(value) { return [value & 255, (value >>> 8) & 255, (value >>> 16) & 255, (value >>> 24) & 255]; }
  var crcTable = null;
  function crc32(bytes) {
    if (!crcTable) {
      crcTable = [];
      for (var n = 0; n < 256; n++) {
        var c = n;
        for (var k = 0; k < 8; k++) c = (c & 1) ? (0xEDB88320 ^ (c >>> 1)) : (c >>> 1);
        crcTable[n] = c >>> 0;
      }
    }
    var crc = 0xFFFFFFFF;
    for (var i = 0; i < bytes.length; i++) crc = crcTable[(crc ^ bytes[i]) & 255] ^ (crc >>> 8);
    return (crc ^ 0xFFFFFFFF) >>> 0;
  }
  function zipStore(entries) {
    var parts = [], central = [], offset = 0, i;
    for (i = 0; i < entries.length; i++) {
      var entry = entries[i];
      var data = typeof entry.data === 'string' ? utf8Bytes(entry.data) : entry.data;
      var name = utf8Bytes(entry.name);
      var crc = crc32(data);
      var local = [].concat(u32(0x04034b50), u16(20), u16(0x0800), u16(0), u16(0), u16(0),
        u32(crc), u32(data.length), u32(data.length), u16(name.length), u16(0));
      parts.push(new Uint8Array(local)); parts.push(name); parts.push(data);
      central.push({ name: name, crc: crc, size: data.length, offset: offset });
      offset += local.length + name.length + data.length;
    }
    var centralStart = offset;
    for (i = 0; i < central.length; i++) {
      var c = central[i];
      var cd = [].concat(u32(0x02014b50), u16(20), u16(20), u16(0x0800), u16(0), u16(0), u16(0),
        u32(c.crc), u32(c.size), u32(c.size), u16(c.name.length), u16(0), u16(0),
        u16(0), u16(0), u32(0), u32(c.offset));
      parts.push(new Uint8Array(cd)); parts.push(c.name);
      offset += cd.length + c.name.length;
    }
    parts.push(new Uint8Array([].concat(u32(0x06054b50), u16(0), u16(0), u16(central.length), u16(central.length),
      u32(offset - centralStart), u32(centralStart), u16(0))));
    return concatBytes(parts);
  }
  function readU16(bytes, at) { return bytes[at] | (bytes[at + 1] << 8); }
  function readU32(bytes, at) { return (bytes[at] | (bytes[at + 1] << 8) | (bytes[at + 2] << 16) | (bytes[at + 3] << 24)) >>> 0; }
  async function unzip(bytes) {
    var data = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
    var eocd = -1;
    for (var i = data.length - 22; i >= Math.max(0, data.length - 65557); i--) if (readU32(data, i) === 0x06054b50) { eocd = i; break; }
    if (eocd < 0) throw new Error('invalid ZIP file');
    var count = readU16(data, eocd + 10), at = readU32(data, eocd + 16), files = {};
    for (var n = 0; n < count; n++) {
      if (readU32(data, at) !== 0x02014b50) throw new Error('invalid ZIP directory');
      var method = readU16(data, at + 10), compSize = readU32(data, at + 20);
      var nameLen = readU16(data, at + 28), extraLen = readU16(data, at + 30), commentLen = readU16(data, at + 32);
      var localOffset = readU32(data, at + 42);
      var name = new TextDecoder().decode(data.slice(at + 46, at + 46 + nameLen));
      var localName = readU16(data, localOffset + 26), localExtra = readU16(data, localOffset + 28);
      var dataAt = localOffset + 30 + localName + localExtra;
      var chunk = data.slice(dataAt, dataAt + compSize);
      if (method === 0) files[name] = chunk;
      else if (method === 8) {
        if (typeof DecompressionStream === 'undefined') throw new Error('ZIP deflate is not supported in this browser');
        var stream = new Blob([chunk]).stream().pipeThrough(new DecompressionStream('deflate-raw'));
        files[name] = new Uint8Array(await new Response(stream).arrayBuffer());
      }
      at += 46 + nameLen + extraLen + commentLen;
    }
    return files;
  }
  function xmlEscapeText(text) { return String(text == null ? '' : text).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;'); }
  function plainTextFromDoc(doc) {
    var out = [];
    function inline(nodes) {
      var s = '';
      for (var i = 0; i < (nodes || []).length; i++) {
        var n = nodes[i];
        if (n.type === 'text') s += n.text || '';
        else if (n.type === 'break') s += '\n';
        else if (n.type === 'image') s += n.alt || n.caption || '';
      }
      return s;
    }
    function block(n) {
      if (!n) return;
      if (n.type === 'paragraph' || n.type === 'heading') out.push(inline(n.children));
      else if (n.type === 'blockquote') { for (var q = 0; q < (n.children || []).length; q++) block(n.children[q]); }
      else if (n.type === 'code_block') out.push(String(n.text || ''));
      else if (n.type === 'hr') out.push('---');
      else if (n.type === 'image') out.push(n.alt || n.caption || '');
      else if (n.type === 'bullet_list' || n.type === 'ordered_list') {
        for (var i = 0; i < (n.children || []).length; i++) {
          var item = n.children[i];
          var text = item.children && item.children[0] ? inline(item.children[0].children) : '';
          out.push((n.type === 'ordered_list' ? (i + 1) + '. ' : '- ') + text);
        }
      } else if (n.type === 'table') {
        for (var r = 0; r < (n.rows || []).length; r++) {
          var cells = [];
          for (var c = 0; c < (n.rows[r].cells || []).length; c++) cells.push(inline(n.rows[r].cells[c].children));
          out.push(cells.join('\t'));
        }
      }
    }
    for (var i = 0; i < (doc.children || []).length; i++) block(doc.children[i]);
    return out.join('\n\n');
  }
  function importPlain(text) {
    var lines = String(text == null ? '' : text).replace(/\r\n/g, '\n').split('\n');
    var children = [];
    for (var i = 0; i < lines.length; i++) {
      if (lines[i] === '' && i === lines.length - 1) continue;
      children.push({ type: 'paragraph', children: lines[i] ? [{ type: 'text', text: lines[i], marks: [] }] : [] });
    }
    return { type: 'doc', children: children.length ? children : [{ type: 'paragraph', children: [] }] };
  }
  function rtfInlineFromText(text) { return text ? [{ type: 'text', text: text, marks: [] }] : []; }
  function decodeCP1252(code) {
    if (code >= 0x80 && code <= 0x9F) {
      var special = { 0x80: '\u20ac', 0x82: '\u201a', 0x83: '\u0192', 0x84: '\u201e', 0x85: '\u2026', 0x86: '\u2020', 0x87: '\u2021', 0x88: '\u02c6', 0x89: '\u2030', 0x8A: '\u0160', 0x8B: '\u2039', 0x8C: '\u0152', 0x8E: '\u017d', 0x91: '\u2018', 0x92: '\u2019', 0x93: '\u201c', 0x94: '\u201d', 0x95: '\u2022', 0x96: '\u2013', 0x97: '\u2014', 0x98: '\u02dc', 0x99: '\u2122', 0x9A: '\u0161', 0x9B: '\u203a', 0x9C: '\u0153', 0x9E: '\u017e', 0x9F: '\u0178' };
      return special[code] || '';
    }
    return String.fromCharCode(code);
  }
  function importRTF(text) {
    var source = String(text || '');
    var colorTable = [];
    var colorGroup = /\\colortbl([^}]*)/i.exec(source);
    if (colorGroup) {
      var colorRe = /\\red(\d+)\\green(\d+)\\blue(\d+)/g, cm;
      while ((cm = colorRe.exec(colorGroup[1]))) colorTable.push('#' + [cm[1], cm[2], cm[3]].map(function (v) { return ('0' + Math.max(0, Math.min(255, Number(v))).toString(16)).slice(-2); }).join(''));
    }
    var state = { bold: false, italic: false, underline: false, strike: false, sup: false, sub: false, color: '', highlight: '', size: '' };
    var stack = [];
    var i = 0, depth = 0, skipDepth = -1, ucSkip = 1;
    var current = [], blocks = [], buffer = '', align = '';
    function marksNow() {
      var marks = [];
      if (state.bold) marks.push({ type: 'bold' });
      if (state.italic) marks.push({ type: 'italic' });
      if (state.underline) marks.push({ type: 'underline' });
      if (state.strike) marks.push({ type: 'strike' });
      if (state.sup) marks.push({ type: 'sup' });
      if (state.sub) marks.push({ type: 'sub' });
      if (state.color) marks.push({ type: 'color', value: state.color });
      if (state.highlight) marks.push({ type: 'background', value: state.highlight });
      if (state.size) marks.push({ type: 'size', value: state.size });
      return marks;
    }
    function flushText() {
      if (buffer) {
        current.push({ type: 'text', text: buffer, marks: marksNow() });
        buffer = '';
      }
    }
    function flushBlock() {
      flushText();
      if (current.length) {
        blocks.push({ type: 'paragraph', align: align, children: current });
        current = [];
      }
      align = '';
    }
    function addText(value) { if (skipDepth < 0) buffer += value; }
    function addBreak() { flushText(); if (skipDepth < 0) current.push({ type: 'break', marks: [] }); }
    while (i < source.length) {
      var ch = source.charAt(i);
      if (ch === '{') {
        depth++;
        stack.push({ bold: state.bold, italic: state.italic, underline: state.underline, strike: state.strike, sup: state.sup, sub: state.sub, color: state.color, highlight: state.highlight, size: state.size });
        i++;
        continue;
      }
      if (ch === '}') {
        flushText();
        if (skipDepth === depth) skipDepth = -1;
        depth--;
        if (stack.length) state = stack.pop();
        i++;
        continue;
      }
      if (ch === '\\') {
        i++;
        if (source.charAt(i) === '*') {
          skipDepth = depth;
          i++;
          continue;
        }
        if (source.charAt(i) === "'") {
          var hex = source.slice(i + 1, i + 3);
          if (/^[0-9a-fA-F]{2}$/.test(hex)) addText(decodeCP1252(parseInt(hex, 16)));
          i += 3;
          continue;
        }
        if (source.charAt(i) === '\\' || source.charAt(i) === '{' || source.charAt(i) === '}') {
          addText(source.charAt(i)); i++; continue;
        }
        if (source.charAt(i) === '~') { addText('\u00a0'); i++; continue; }
        if (source.charAt(i) === '_') { addText('\u00ad'); i++; continue; }
        if (source.charAt(i) === '\n' || source.charAt(i) === '\r') { i++; continue; }
        var start = i;
        while (i < source.length && /[a-zA-Z]/.test(source.charAt(i))) i++;
        var word = source.slice(start, i);
        var numText = '';
        if (source.charAt(i) === '-' || /[0-9]/.test(source.charAt(i))) {
          if (source.charAt(i) === '-') i++;
          while (i < source.length && /[0-9]/.test(source.charAt(i))) { numText += source.charAt(i); i++; }
          if (source.charAt(i - 1) === '-') numText = String(-Number(numText || '0'));
        }
        var hasNum = numText !== '';
        var num = hasNum ? Number(numText) : 0;
        if (source.charAt(i) === ' ') i++;
        if (word === 'fonttbl' || word === 'colortbl' || word === 'stylesheet' || word === 'info' || word === 'pict' || word === 'header' || word === 'footer' || word === 'object' || word === 'themedata' || word === 'datastore' || word === 'latentstyles') {
          skipDepth = depth;
          continue;
        }
        if (skipDepth >= 0) continue;
        if (word === 'par' || word === 'sect' || word === 'page') flushBlock();
        else if (word === 'line') addBreak();
        else if (word === 'tab') addText('\t');
        else if (word === 'b') { flushText(); state.bold = !hasNum || num !== 0; }
        else if (word === 'i') { flushText(); state.italic = !hasNum || num !== 0; }
        else if (word === 'ul') { flushText(); state.underline = !hasNum || num !== 0; }
        else if (word === 'strike') { flushText(); state.strike = !hasNum || num !== 0; }
        else if (word === 'super') { flushText(); state.sup = !hasNum || num !== 0; state.sub = false; }
        else if (word === 'sub') { flushText(); state.sub = !hasNum || num !== 0; state.sup = false; }
        else if (word === 'nosupersub') { flushText(); state.sup = false; state.sub = false; }
        else if (word === 'plain') { flushText(); state.bold = state.italic = state.underline = state.strike = state.sup = state.sub = false; state.color = state.highlight = state.size = ''; }
        else if (word === 'pard') { flushBlock(); }
        else if (word === 'qc' || word === 'center') align = 'center';
        else if (word === 'qr' || word === 'right') align = 'right';
        else if (word === 'ql' || word === 'left' || word === 'qj') align = word === 'qj' ? 'justify' : (word === 'ql' ? 'left' : '');
        else if (word === 'fs' && hasNum) { flushText(); state.size = Math.max(6, Math.round(num / 2)) + 'pt'; }
        else if (word === 'cf' && hasNum) { flushText(); state.color = colorTable[num] || ''; }
        else if (word === 'highlight' && hasNum) { flushText(); state.highlight = colorTable[num] || ''; }
        else if (word === 'u' && hasNum) {
          var code = num < 0 ? num + 65536 : num;
          addText(String.fromCharCode(code));
          var skip = ucSkip;
          while (skip > 0 && i < source.length && source.charAt(i) !== '\\' && source.charAt(i) !== '}' && source.charAt(i) !== '{') { i++; skip--; }
        }
        else if (word === 'uc' && hasNum) ucSkip = Math.max(0, num);
        else if (word === 'bullet') addText('\u2022 ');
        else if (word === 'emdash') addText('\u2014');
        else if (word === 'endash') addText('\u2013');
        continue;
      }
      if (ch === '\n' || ch === '\r') { i++; continue; }
      addText(ch);
      i++;
    }
    flushBlock();
    blocks = blocks.filter(function (b) { return b.children && b.children.length && !(b.children.length === 1 && b.children[0].type === 'text' && !b.children[0].text.trim()); });
    if (!blocks.length) return importPlain(source);
    return { type: 'doc', children: blocks };
  }
  function rtfMarks(marks) {
    var open = '', close = '';
    for (var i = 0; i < (marks || []).length; i++) {
      var t = marks[i].type;
      if (t === 'bold') { open += '\\b '; close = '\\b0 ' + close; }
      else if (t === 'italic') { open += '\\i '; close = '\\i0 ' + close; }
      else if (t === 'underline') { open += '\\ul '; close = '\\ul0 ' + close; }
      else if (t === 'strike') { open += '\\strike '; close = '\\strike0 ' + close; }
      else if (t === 'sup') { open += '\\super '; close = '\\nosupersub ' + close; }
      else if (t === 'sub') { open += '\\sub '; close = '\\nosupersub ' + close; }
    }
    return { open: open, close: close };
  }
  function exportRTF(inst) {
    var out = ['{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Helvetica;}}\\viewkind4'];
    function inline(nodes) {
      var s = '';
      for (var i = 0; i < (nodes || []).length; i++) {
        var n = nodes[i];
        if (n.type === 'break') { s += '\\line '; continue; }
        if (n.type !== 'text') continue;
        var m = rtfMarks(n.marks);
        var value = String(n.text || '').replace(/\\/g, '\\\\').replace(/[{}]/g, '\\$&').replace(/\n/g, '\\line ');
        value = value.replace(/[^\x00-\x7F]/g, function (ch) { var code = ch.charCodeAt(0); return code > 32767 ? '\\u' + (code - 65536) + '?' : '\\u' + code + '?'; });
        s += m.open + value + m.close;
      }
      return s;
    }
    function block(n) {
      if (!n) return;
      if (n.type === 'heading') { out.push('\\pard\\sa120\\b\\fs' + (28 + (4 - Math.min(4, Number(n.level) || 1)) * 4) + ' ' + inline(n.children) + '\\b0\\fs24\\par'); }
      else if (n.type === 'paragraph') {
        out.push('\\pard\\sa120 ' + (n.align === 'center' ? '\\qc ' : n.align === 'right' ? '\\qr ' : '') + inline(n.children) + '\\par');
      } else if (n.type === 'blockquote') { for (var k = 0; k < (n.children || []).length; k++) block(n.children[k]); }
      else if (n.type === 'code_block') { out.push('\\pard\\f0\\fs20\\cf1 ' + String(n.text || '').replace(/\n/g, '\\line ') + '\\cf0\\fs24\\par'); }
      else if (n.type === 'hr') out.push('\\pard\\brdrb\\brdrs\\brdrw10\\par');
      else if (n.type === 'image') out.push('\\pard[' + (n.alt || 'image') + ']\\par');
      else if (n.type === 'bullet_list' || n.type === 'ordered_list') {
        for (var i = 0; i < (n.children || []).length; i++) {
          var item = n.children[i], text = item.children && item.children[0] ? inline(item.children[0].children) : '';
          out.push('\\pard\\li360 ' + (n.type === 'ordered_list' ? (i + 1) + '. ' : '\\u8226? ') + text + '\\par');
        }
      } else if (n.type === 'table') {
        for (var r = 0; r < (n.rows || []).length; r++) {
          var cells = [];
          for (var c = 0; c < (n.rows[r].cells || []).length; c++) cells.push(inline(n.rows[r].cells[c].children));
          out.push('\\pard\\trowd\\trgaph108\\cellx2000\\cellx4000\\cellx6000 ' + cells.join('\\cell ') + '\\cell\\row\\pard');
        }
      }
    }
    for (var i = 0; i < (inst.doc.children || []).length; i++) block(inst.doc.children[i]);
    out.push('}');
    return out.join('\n');
  }
  function docxRun(text, marks) {
    var props = '';
    for (var i = 0; i < (marks || []).length; i++) {
      var t = marks[i].type;
      if (t === 'bold') props += '<w:b/>';
      else if (t === 'italic') props += '<w:i/>';
      else if (t === 'underline') props += '<w:u w:val="single"/>';
      else if (t === 'strike') props += '<w:strike/>';
      else if (t === 'sup') props += '<w:vertAlign w:val="superscript"/>';
      else if (t === 'sub') props += '<w:vertAlign w:val="subscript"/>';
      else if (t === 'link') props += '<w:color w:val="4C5CF7"/><w:u w:val="single"/>';
      else if (t === 'color' && /^#[0-9a-f]{6}$/i.test(marks[i].value || '')) props += '<w:color w:val="' + marks[i].value.slice(1) + '"/>';
      else if (t === 'size' && marks[i].value) props += '<w:sz w:val="' + Math.max(8, Math.min(96, parseInt(marks[i].value, 10) * 2 || 24)) + '"/>';
    }
    return '<w:r><w:rPr>' + props + '</w:rPr><w:t xml:space="preserve">' + xmlEscapeText(text) + '</w:t></w:r>';
  }
  function docxInline(nodes) {
    var out = '';
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (n.type === 'text') out += docxRun(n.text || '', n.marks);
      else if (n.type === 'break') out += '<w:r><w:br/></w:r>';
      else if (n.type === 'image') out += docxRun('[' + (n.alt || 'image') + ']', []);
    }
    return out;
  }
  function docxBlock(n) {
    if (!n) return '';
    if (n.type === 'paragraph' || n.type === 'heading') {
      var pPr = '<w:pPr>';
      if (n.align) pPr += '<w:jc w:val="' + (n.align === 'justify' ? 'both' : n.align) + '"/>';
      if (n.type === 'heading') pPr += '<w:spacing w:before="240" w:after="120"/>';
      pPr += '</w:pPr>';
      return '<w:p>' + pPr + docxInline(n.children) + '</w:p>';
    }
    if (n.type === 'blockquote') return (n.children || []).map(docxBlock).join('');
    if (n.type === 'code_block') return '<w:p><w:pPr><w:shd w:fill="F2F2F2"/></w:pPr>' + docxRun(String(n.text || ''), [{ type: 'code' }]) + '</w:p>';
    if (n.type === 'hr') return '<w:p><w:pPr><w:pBdr><w:bottom w:val="single" w:sz="6" w:space="1" w:color="999999"/></w:pBdr></w:pPr></w:p>';
    if (n.type === 'image') return '<w:p>' + docxRun('[' + (n.alt || 'image') + ']', []) + '</w:p>';
    if (n.type === 'bullet_list' || n.type === 'ordered_list') {
      var out = '';
      for (var i = 0; i < (n.children || []).length; i++) {
        var item = n.children[i], text = item.children && item.children[0] ? docxInline(item.children[0].children) : '';
        out += '<w:p><w:pPr><w:ind w:left="360"/></w:pPr>' + docxRun((n.type === 'ordered_list' ? (i + 1) + '. ' : '\u2022 '), []) + text + '</w:p>';
      }
      return out;
    }
    if (n.type === 'table') {
      var rows = '';
      for (var r = 0; r < (n.rows || []).length; r++) {
        var cells = '';
        for (var c = 0; c < (n.rows[r].cells || []).length; c++) {
          cells += '<w:tc><w:tcPr>' + (Number(n.rows[r].cells[c].colspan) > 1 ? '<w:gridSpan w:val="' + Number(n.rows[r].cells[c].colspan) + '"/>' : '') + '</w:tcPr><w:p>' + docxInline(n.rows[r].cells[c].children) + '</w:p></w:tc>';
        }
        rows += '<w:tr>' + cells + '</w:tr>';
      }
      return '<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders><w:top w:val="single" w:sz="4" w:color="999999"/><w:left w:val="single" w:sz="4" w:color="999999"/><w:bottom w:val="single" w:sz="4" w:color="999999"/><w:right w:val="single" w:sz="4" w:color="999999"/><w:insideH w:val="single" w:sz="4" w:color="999999"/><w:insideV w:val="single" w:sz="4" w:color="999999"/></w:tblBorders></w:tblPr>' + rows + '</w:tbl><w:p/>';
    }
    return '';
  }
  function exportDOCX(inst) {
    var body = '';
    for (var i = 0; i < (inst.doc.children || []).length; i++) body += docxBlock(inst.doc.children[i]);
    var documentXml = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>' + body + '<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134"/></w:sectPr></w:body></w:document>';
    var contentTypes = '<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>';
    var rels = '<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>';
    return zipStore([
      { name: '[Content_Types].xml', data: contentTypes },
      { name: '_rels/.rels', data: rels },
      { name: 'word/document.xml', data: documentXml }
    ]);
  }
  function odtInline(nodes) {
    var out = '';
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (n.type === 'break') { out += '<text:line-break/>'; continue; }
      if (n.type !== 'text') continue;
      var marks = n.marks || [], attrs = [];
      for (var m = 0; m < marks.length; m++) {
        var t = marks[m].type;
        if (t === 'bold') attrs.push('font-weight:bold');
        else if (t === 'italic') attrs.push('font-style:italic');
        else if (t === 'underline') attrs.push('text-decoration:underline');
        else if (t === 'strike') attrs.push('text-decoration:line-through');
        else if (t === 'sup') attrs.push('vertical-align:super');
        else if (t === 'sub') attrs.push('vertical-align:sub');
        else if (t === 'color' && marks[m].value) attrs.push('color:' + marks[m].value);
        else if (t === 'background' && marks[m].value) attrs.push('background-color:' + marks[m].value);
      }
      var text = xmlEscapeText(n.text || '');
      out += attrs.length ? '<text:span text:style-name="T' + (attrs.length % 7) + '" style="' + attrs.join(';') + '">' + text + '</text:span>' : text;
    }
    return out;
  }
  function odtBlock(n) {
    if (!n) return '';
    if (n.type === 'heading') return '<text:h text:style-name="Heading_20_' + Math.min(6, Number(n.level) || 1) + '" text:outline-level="' + (Number(n.level) || 1) + '">' + odtInline(n.children) + '</text:h>';
    if (n.type === 'paragraph') {
      var style = n.align ? ' style="text-align:' + n.align + '"' : '';
      return '<text:p' + style + '>' + odtInline(n.children) + '</text:p>';
    }
    if (n.type === 'blockquote') return '<text:section>' + (n.children || []).map(odtBlock).join('') + '</text:section>';
    if (n.type === 'code_block') return '<text:p><text:span style="font-family:monospace">' + xmlEscapeText(n.text || '') + '</text:span></text:p>';
    if (n.type === 'hr') return '<text:p>---</text:p>';
    if (n.type === 'image') return '<text:p>[' + xmlEscapeText(n.alt || n.caption || 'image') + ']</text:p>';
    if (n.type === 'bullet_list' || n.type === 'ordered_list') {
      var out = '<text:list>';
      for (var i = 0; i < (n.children || []).length; i++) {
        var item = n.children[i], text = item.children && item.children[0] ? odtInline(item.children[0].children) : '';
        out += '<text:list-item><text:p>' + (n.type === 'ordered_list' ? (i + 1) + '. ' : '\u2022 ') + text + '</text:p></text:list-item>';
      }
      return out + '</text:list>';
    }
    if (n.type === 'table') {
      var out = '<table:table>';
      for (var r = 0; r < (n.rows || []).length; r++) {
        out += '<table:table-row>';
        for (var c = 0; c < (n.rows[r].cells || []).length; c++) {
          var cell = n.rows[r].cells[c];
          out += '<table:table-cell office:value-type="string"' + (Number(cell.colspan) > 1 ? ' table:number-columns-spanned="' + Number(cell.colspan) + '"' : '') + '><text:p>' + odtInline(cell.children) + '</text:p></table:table-cell>';
        }
        out += '</table:table-row>';
      }
      return out + '</table:table>';
    }
    return '';
  }
  function exportODT(inst) {
    var body = '';
    for (var i = 0; i < (inst.doc.children || []).length; i++) body += odtBlock(inst.doc.children[i]);
    var content = '<?xml version="1.0" encoding="UTF-8"?><office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" office:version="1.2"><office:body><office:text>' + body + '</office:text></office:body></office:document-content>';
    var styles = '<?xml version="1.0" encoding="UTF-8"?><office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" office:version="1.2"/>';
    var manifest = '<?xml version="1.0" encoding="UTF-8"?><manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.text"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/><manifest:file-entry manifest:full-path="styles.xml" manifest:media-type="text/xml"/></manifest:manifest>';
    return zipStore([
      { name: 'mimetype', data: utf8Bytes('application/vnd.oasis.opendocument.text') },
      { name: 'content.xml', data: content },
      { name: 'styles.xml', data: styles },
      { name: 'META-INF/manifest.xml', data: manifest }
    ]);
  }
  function importDOCX(files) {
    var xml = new TextDecoder().decode(files['word/document.xml'] || new Uint8Array());
    if (!xml) throw new Error('DOCX has no word/document.xml');
    var dom = new DOMParser().parseFromString(xml, 'application/xml');
    var body = dom.getElementsByTagNameNS('*', 'body')[0];
    if (!body) throw new Error('DOCX has no body');
    function inline(node) {
      var out = [];
      var runs = node.getElementsByTagNameNS('*', 'r');
      for (var i = 0; i < runs.length; i++) {
        var marks = [], props = runs[i].getElementsByTagNameNS('*', 'rPr')[0];
        if (props) {
          if (props.getElementsByTagNameNS('*', 'b').length) marks.push({ type: 'bold' });
          if (props.getElementsByTagNameNS('*', 'i').length) marks.push({ type: 'italic' });
          if (props.getElementsByTagNameNS('*', 'u').length) marks.push({ type: 'underline' });
          if (props.getElementsByTagNameNS('*', 'strike').length) marks.push({ type: 'strike' });
          var vert = props.getElementsByTagNameNS('*', 'vertAlign')[0];
          if (vert) marks.push({ type: vert.getAttribute('w:val') === 'subscript' ? 'sub' : 'sup' });
        }
        var texts = runs[i].getElementsByTagNameNS('*', 't');
        var text = '';
        for (var t = 0; t < texts.length; t++) text += texts[t].textContent;
        if (text) out.push({ type: 'text', text: text, marks: marks });
      }
      return out;
    }
    function blocks(parent) {
      var out = [];
      for (var i = 0; i < parent.childNodes.length; i++) {
        var n = parent.childNodes[i];
        if (n.nodeType !== 1) continue;
        var tag = n.localName;
        if (tag === 'p') {
          var pPr = n.getElementsByTagNameNS('*', 'pPr')[0];
          var heading = 0, align = '';
          if (pPr) {
            var style = pPr.getElementsByTagNameNS('*', 'pStyle')[0];
            if (style) { var m = /^Heading(\d)$/i.exec(style.getAttribute('w:val') || style.getAttribute('val') || ''); if (m) heading = Number(m[1]); }
            var jc = pPr.getElementsByTagNameNS('*', 'jc')[0];
            if (jc) align = jc.getAttribute('w:val') || '';
          }
          out.push({ type: heading ? 'heading' : 'paragraph', level: heading || undefined, align: align === 'both' ? 'justify' : align, children: inline(n) });
        } else if (tag === 'tbl') {
          var rows = [], trs = n.getElementsByTagNameNS('*', 'tr');
          for (var r = 0; r < trs.length; r++) {
            var row = { type: 'table_row', cells: [] };
            for (var c = 0; c < trs[r].childNodes.length; c++) if (trs[r].childNodes[c].nodeType === 1 && trs[r].childNodes[c].localName === 'tc') {
              row.cells.push({ type: 'table_cell', header: false, colspan: 1, rowspan: 1, children: inline(trs[r].childNodes[c]) });
            }
            rows.push(row);
          }
          out.push({ type: 'table', rows: rows });
        }
      }
      return out;
    }
    return { type: 'doc', children: blocks(body) };
  }
  function importODT(files) {
    var xml = new TextDecoder().decode(files['content.xml'] || new Uint8Array());
    if (!xml) throw new Error('ODT has no content.xml');
    var dom = new DOMParser().parseFromString(xml, 'application/xml');
    var body = dom.getElementsByTagNameNS('*', 'text')[0];
    if (!body) throw new Error('ODT has no text body');
    function inline(node) {
      var out = [];
      for (var i = 0; i < node.childNodes.length; i++) {
        var n = node.childNodes[i];
        if (n.nodeType === 3 && n.nodeValue) out.push({ type: 'text', text: n.nodeValue, marks: [] });
        else if (n.nodeType === 1 && n.localName === 'span') {
          var child = inline(n);
          var style = n.getAttribute('style') || '';
          var marks = [];
          if (/font-weight\s*:\s*bold/.test(style)) marks.push({ type: 'bold' });
          if (/font-style\s*:\s*italic/.test(style)) marks.push({ type: 'italic' });
          if (/underline/.test(style)) marks.push({ type: 'underline' });
          if (/line-through/.test(style)) marks.push({ type: 'strike' });
          if (/vertical-align\s*:\s*super/.test(style)) marks.push({ type: 'sup' });
          if (/vertical-align\s*:\s*sub/.test(style)) marks.push({ type: 'sub' });
          for (var c = 0; c < child.length; c++) child[c].marks = (child[c].marks || []).concat(marks);
          out = out.concat(child);
        } else if (n.nodeType === 1) out = out.concat(inline(n));
      }
      return out;
    }
    function blocks(parent) {
      var out = [];
      for (var i = 0; i < parent.childNodes.length; i++) {
        var n = parent.childNodes[i];
        if (n.nodeType !== 1) continue;
        var tag = n.localName;
        if (tag === 'h') out.push({ type: 'heading', level: Number(n.getAttribute('outline-level') || 1), children: inline(n) });
        else if (tag === 'p') out.push({ type: 'paragraph', children: inline(n) });
        else if (tag === 'list') {
          var list = { type: 'bullet_list', children: [] };
          for (var c = 0; c < n.childNodes.length; c++) if (n.childNodes[c].nodeType === 1 && n.childNodes[c].localName === 'list-item') list.children.push({ type: 'list_item', children: blocks(n.childNodes[c]) });
          out.push(list);
        } else if (tag === 'table') {
          var table = { type: 'table', rows: [] };
          var trs = n.getElementsByTagNameNS('*', 'table-row');
          for (var r = 0; r < trs.length; r++) {
            var row = { type: 'table_row', cells: [] };
            for (var cc = 0; cc < trs[r].childNodes.length; cc++) if (trs[r].childNodes[cc].nodeType === 1 && trs[r].childNodes[cc].localName === 'table-cell') row.cells.push({ type: 'table_cell', header: false, colspan: 1, rowspan: 1, children: inline(trs[r].childNodes[cc]) });
            table.rows.push(row);
          }
          out.push(table);
        }
      }
      return out;
    }
    return { type: 'doc', children: blocks(body) };
  }
  function pdfEscape(text) { return String(text == null ? '' : text).replace(/\\/g, '\\\\').replace(/\(/g, '\\(').replace(/\)/g, '\\)'); }
  function pdfBytes(text) { var out = []; for (var i = 0; i < text.length; i++) { var c = text.charCodeAt(i); if (c > 255) c = 63; out.push(c); } return out; }
  function wrapPdf(text, maxChars) {
    var words = String(text == null ? '' : text).split(/\s+/), lines = [], line = '';
    for (var i = 0; i < words.length; i++) {
      var next = line ? line + ' ' + words[i] : words[i];
      if (next.length > maxChars && line) { lines.push(line); line = words[i]; }
      else line = next;
    }
    if (line) lines.push(line);
    return lines.length ? lines : [''];
  }
  function exportPDF(inst) {
    var pages = [], content = [], y = 785;
    function newPage() { if (content.length) pages.push(content.join('\n')); content = []; y = 785; }
    function line(text, size, font, indent) {
      var lines = wrapPdf(text, Math.max(20, Math.floor((500 - (indent || 0)) / (size * 0.52))));
      for (var i = 0; i < lines.length; i++) {
        if (y < 56) newPage();
        content.push('BT /' + (font || 'F1') + ' ' + size + ' Tf ' + (56 + (indent || 0)) + ' ' + y + ' Td (' + pdfEscape(lines[i]) + ') Tj ET');
        y -= size * 1.35;
      }
    }
    function inlineText(nodes) {
      var s = '';
      for (var i = 0; i < (nodes || []).length; i++) if (nodes[i].type === 'text') s += nodes[i].text || '';
      return s;
    }
    function block(n) {
      if (!n) return;
      if (n.type === 'heading') line(inlineText(n.children), 18 - Math.min(4, Number(n.level) || 1) * 2, 'F2');
      else if (n.type === 'paragraph') line(inlineText(n.children), 11, 'F1');
      else if (n.type === 'blockquote') line(inlineText(n.children), 11, 'F3', 18);
      else if (n.type === 'code_block') line(String(n.text || ''), 9, 'F1');
      else if (n.type === 'hr') { if (y < 56) newPage(); content.push('0.7 w 56 ' + y + ' m 539 ' + y + ' l S'); y -= 14; }
      else if (n.type === 'image') line('[' + (n.alt || 'image') + ']', 10, 'F3');
      else if (n.type === 'bullet_list' || n.type === 'ordered_list') {
        for (var i = 0; i < (n.children || []).length; i++) {
          var item = n.children[i], text = item.children && item.children[0] ? inlineText(item.children[0].children) : '';
          line((n.type === 'ordered_list' ? (i + 1) + '. ' : '\u2022 ') + text, 11, 'F1', 14);
        }
      } else if (n.type === 'table') {
        for (var r = 0; r < (n.rows || []).length; r++) {
          var cells = [];
          for (var c = 0; c < (n.rows[r].cells || []).length; c++) cells.push(inlineText(n.rows[r].cells[c].children));
          line(cells.join(' | '), 9, 'F1');
        }
      }
    }
    for (var i = 0; i < (inst.doc.children || []).length; i++) block(inst.doc.children[i]);
    newPage();
    var objects = [], pageIds = [];
    var fontIds = { F1: 3, F2: 4, F3: 5 };
    var objectCount = 2 + 3 + pages.length * 2;
    for (var p = 0; p < pages.length; p++) pageIds.push(7 + p * 2);
    objects[1] = '<< /Type /Catalog /Pages 2 0 R >>';
    objects[2] = '<< /Type /Pages /Kids [' + pageIds.map(function (id) { return id + ' 0 R'; }).join(' ') + '] /Count ' + pages.length + ' >>';
    objects[3] = '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>';
    objects[4] = '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>';
    objects[5] = '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Oblique >>';
    for (p = 0; p < pages.length; p++) {
      var contentId = 8 + p * 2;
      var stream = pages[p];
      objects[7 + p * 2] = '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 3 0 R /F2 4 0 R /F3 5 0 R >> >> /Contents ' + contentId + ' 0 R >>';
      objects[contentId] = '<< /Length ' + pdfBytes(stream).length + ' >>\nstream\n' + stream + '\nendstream';
    }
    var out = '%PDF-1.4\n', offsets = [];
    for (var id = 1; id < objects.length; id++) {
      if (!objects[id]) continue;
      offsets[id] = pdfBytes(out).length;
      out += id + ' 0 obj\n' + objects[id] + '\nendobj\n';
    }
    var xref = pdfBytes(out).length;
    out += 'xref\n0 ' + objects.length + '\n0000000000 65535 f \n';
    for (id = 1; id < objects.length; id++) out += ('0000000000' + (offsets[id] || 0)).slice(-10) + ' 00000 n \n';
    out += 'trailer << /Size ' + objects.length + ' /Root 1 0 R >>\nstartxref\n' + xref + '\n%%EOF';
    return new Uint8Array(pdfBytes(out));
  }
  function formatFromName(filename) {
    var m = /\.([a-z0-9]+)$/i.exec(String(filename || ''));
    return m ? m[1].toLowerCase() : '';
  }
  function exportBytes(name, format) {
    var inst = instances[name];
    if (!inst) return [null, 'richtext instance not found: ' + name];
    sync(inst);
    var fmt = String(format || 'txt').toLowerCase();
    try {
      if (fmt === 'txt' || fmt === 'text') return [bytesFromText(plainTextFromDoc(inst.doc)), null];
      if (fmt === 'rtf' || fmt === 'doc') return [bytesFromText(exportRTF(inst)), null];
      if (fmt === 'docx') return [Array.prototype.slice.call(exportDOCX(inst)), null];
      if (fmt === 'odt') return [Array.prototype.slice.call(exportODT(inst)), null];
      if (fmt === 'pdf') return [Array.prototype.slice.call(exportPDF(inst)), null];
      if (fmt === 'html') return [bytesFromText(exportHTML(inst)), null];
      if (fmt === 'markdown' || fmt === 'md') return [bytesFromText(exportMarkdown(inst)), null];
      if (fmt === 'xml') return [bytesFromText(exportXML(inst)), null];
      if (fmt === 'json') return [bytesFromText(JSON.stringify(inst.doc)), null];
      return [null, 'unsupported richtext format: ' + format];
    } catch (e) { return [null, String(e && e.message ? e.message : e)]; }
  }
  async function importBytes(name, bytes, filename) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var fmt = formatFromName(filename || '');
    if (!fmt) fmt = 'txt';
    try {
      var data = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
      var doc;
      if (fmt === 'txt' || fmt === 'text') {
        doc = importPlain(new TextDecoder().decode(data));
      } else if (fmt === 'rtf') {
        doc = importRTF(new TextDecoder().decode(data));
      } else if (fmt === 'doc') {
        if (data.length > 4 && data[0] === 0xD0 && data[1] === 0xCF && data[2] === 0x11 && data[3] === 0xE0) return [false, 'legacy binary .doc is not supported; use .docx or .rtf'];
        var docText = new TextDecoder().decode(data);
        if (/^\s*\{\\rtf/i.test(docText)) doc = importRTF(docText);
        else if (/^\s*</.test(docText)) doc = importHTML(docText);
        else doc = importPlain(docText);
      } else if (fmt === 'docx') {
        doc = importDOCX(await unzip(data));
      } else if (fmt === 'odt') {
        doc = importODT(await unzip(data));
      } else if (fmt === 'html' || fmt === 'htm') {
        doc = importHTML(new TextDecoder().decode(data));
      } else if (fmt === 'md' || fmt === 'markdown') {
        doc = importMarkdown(new TextDecoder().decode(data));
      } else if (fmt === 'xml') {
        doc = importXML(new TextDecoder().decode(data));
      } else if (fmt === 'json') {
        doc = parseJSON(new TextDecoder().decode(data), null);
      } else return [false, 'unsupported richtext format: ' + fmt];
      if (!doc || typeof doc !== 'object') return [false, 'invalid richtext document'];
      setValue(name, doc);
      return [true, null];
    } catch (e) { return [false, String(e && e.message ? e.message : e)]; }
  }
  function openFile(name) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var input = document.createElement('input');
    input.type = 'file';
    input.accept = '.txt,.text,.rtf,.doc,.docx,.odt,.html,.htm,.md,.markdown,.json,.xml';
    input.style.display = 'none';
    input.addEventListener('change', async function () {
      if (!input.files || !input.files[0]) return;
      var file = input.files[0];
      try {
        var buffer = await file.arrayBuffer();
        var res = await importBytes(name, new Uint8Array(buffer), file.name);
        if (res[1] !== null) window.alert(res[1]);
      } catch (e) { window.alert(String(e && e.message ? e.message : e)); }
      input.remove();
    });
    document.body.appendChild(input);
    input.click();
    return [true, null];
  }
  function saveAs(name, format) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var fmt = String(format || 'txt').toLowerCase();
    var result = exportBytes(name, fmt);
    if (result[1] !== null) return [false, result[1]];
    var mime = fmt === 'pdf' ? 'application/pdf' : fmt === 'docx' ? 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' : fmt === 'odt' ? 'application/vnd.oasis.opendocument.text' : fmt === 'rtf' || fmt === 'doc' ? 'application/rtf' : fmt === 'html' ? 'text/html' : fmt === 'md' || fmt === 'markdown' ? 'text/markdown' : fmt === 'json' ? 'application/json' : fmt === 'xml' ? 'application/xml' : 'text/plain';
    var ext = fmt === 'doc' ? 'doc' : fmt;
    var blob = new Blob([new Uint8Array(result[0])], { type: mime });
    var url = URL.createObjectURL(blob);
    var a = document.createElement('a');
    a.href = url; a.download = 'document.' + ext;
    document.body.appendChild(a); a.click();
    setTimeout(function () { URL.revokeObjectURL(url); a.remove(); }, 1000);
    return [true, null];
  }

  return {
    afterRender: afterRender, get: getValue, set: setValue, command: command, format: format,
    insertText: insertTextValue, insertImage: insertImageValue, insertLink: insertLinkValue, insertTable: insertTableValue,
    undo: function (n) { var i = instances[n]; return i && undo(i) ? [true, null] : [false, 'nothing to undo']; },
    redo: function (n) { var i = instances[n]; return i && redo(i) ? [true, null] : [false, 'nothing to redo']; },
    focus: function (n) { var i = instances[n]; if (!i) return [false, 'richtext instance not found: ' + n]; i.editor.focus(); return [true, null]; },
    exportFormat: exportFormat, importFormat: importFormatTo, exportBytes: exportBytes, importBytes: importBytes, openFile: openFile, saveAs: saveAs, registerPlugin: registerPlugin, plugins: function () { return [pluginNames(), null]; }, setTemplate: setTemplate,
    status: status, selection: selection, setToolbar: setToolbar, destroy: destroy
  };
})();
`

// suiRichTextCss styles the editor, toolbar, menus, contextual toolbar and
// properties/status panels. Colors are CSS variables so hosts can theme it.
const suiRichTextCss = `
/* Rich text editor ------------------------------------------------------- */
.sui-rich {
  --sui-rich-bg: var(--sui-panel, #181c24);
  --sui-rich-fg: var(--sui-fg, #e6e6ef);
  --sui-rich-border: var(--sui-border, #2c3444);
  --sui-rich-muted: var(--sui-muted, #9aa7b8);
  --sui-rich-accent: var(--sui-accent, #6f7cff);
  --sui-rich-hover: var(--sui-hover, #242b38);
  --sui-rich-toolbar-bg: var(--sui-rich-bg);
  --sui-rich-editor-bg: var(--sui-rich-bg);
  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0;
  border: 1px solid var(--sui-rich-border);
  border-radius: var(--sui-radius, 10px);
  background: var(--sui-rich-bg);
  color: var(--sui-rich-fg);
  font-family: system-ui, sans-serif;
  overflow: visible;
}
.sui-rich[data-sui-rich-theme="light"] {
  --sui-rich-bg: #ffffff;
  --sui-rich-fg: #1c2330;
  --sui-rich-border: #d6dce8;
  --sui-rich-muted: #65718a;
  --sui-rich-hover: #eef2fa;
  --sui-rich-toolbar-bg: #f7f9fd;
  --sui-rich-editor-bg: #ffffff;
  color-scheme: light;
}
.sui-rich-title { padding: 10px 12px 0; font-weight: 600; color: var(--sui-rich-fg); }
.sui-rich-menu { display: flex; flex-wrap: wrap; gap: 2px; padding: 4px 8px; border-bottom: 1px solid var(--sui-rich-border); background: var(--sui-rich-toolbar-bg); position: relative; z-index: 3; }
.sui-rich-menu-wrap { position: relative; }
.sui-rich-menu-btn { border: 0; background: transparent; color: var(--sui-rich-fg); padding: 5px 9px; border-radius: 6px; cursor: pointer; font: inherit; font-size: 12px; }
.sui-rich-menu-btn:hover, .sui-rich-menu-btn:focus-visible { background: var(--sui-rich-hover); outline: none; }
.sui-rich-menu-panel { position: absolute; left: 0; top: calc(100% + 4px); min-width: 190px; padding: 5px; background: var(--sui-rich-bg); border: 1px solid var(--sui-rich-border); border-radius: 8px; box-shadow: 0 12px 30px rgba(0,0,0,.28); z-index: 20; }
.sui-rich-menu-item { display: block; width: 100%; text-align: left; border: 0; background: transparent; color: var(--sui-rich-fg); padding: 7px 9px; border-radius: 6px; cursor: pointer; font: inherit; font-size: 12px; }
.sui-rich-menu-item:hover, .sui-rich-menu-item:focus-visible { background: var(--sui-rich-hover); outline: none; }
.sui-rich-menu-subwrap { position: relative; }
.sui-rich-menu-parent { display: flex; justify-content: space-between; align-items: center; }
.sui-rich-submenu { left: calc(100% + 4px); top: -5px; min-width: 150px; }
.sui-rich-toolbar { display: flex; flex-wrap: wrap; gap: 5px; padding: 6px 8px; border-bottom: 1px solid var(--sui-rich-border); background: var(--sui-rich-toolbar-bg); }
.sui-rich-group { display: flex; align-items: center; gap: 2px; padding-right: 5px; border-right: 1px solid var(--sui-rich-border); }
.sui-rich-group:last-child { border-right: 0; padding-right: 0; }
.sui-rich-btn { display: inline-flex; align-items: center; justify-content: center; min-width: 28px; height: 28px; padding: 0 6px; border: 1px solid transparent; border-radius: 6px; background: transparent; color: var(--sui-rich-fg); cursor: pointer; font: inherit; font-size: 12px; line-height: 1; }
.sui-rich-btn svg { width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.sui-rich-btn:hover, .sui-rich-btn:focus-visible { background: var(--sui-rich-hover); border-color: var(--sui-rich-border); outline: none; }
.sui-rich-btn:active { transform: translateY(1px); }
.sui-rich-select, .sui-rich-color, .sui-rich-prop { height: 28px; border: 1px solid var(--sui-rich-border); border-radius: 6px; background: var(--sui-rich-bg); color: var(--sui-rich-fg); font: inherit; font-size: 12px; padding: 0 5px; }
.sui-rich-color { width: 34px; padding: 2px; cursor: pointer; }
.sui-rich-body { display: flex; min-height: 220px; background: var(--sui-rich-editor-bg); }
.sui-rich-editor { flex: 1 1 auto; min-width: 0; min-height: 220px; padding: 22px 26px; outline: none; overflow: auto; font-size: 16px; line-height: 1.65; color: var(--sui-rich-fg); overflow-wrap: anywhere; }
.sui-rich-editor:empty::before, .sui-rich-editor[data-placeholder]:empty::before { content: attr(data-placeholder); color: var(--sui-rich-muted); pointer-events: none; }
.sui-rich-editor p, .sui-rich-editor h1, .sui-rich-editor h2, .sui-rich-editor h3, .sui-rich-editor h4, .sui-rich-editor blockquote, .sui-rich-editor pre, .sui-rich-editor figure, .sui-rich-editor table { margin: 0 0 .8em; }
.sui-rich-editor h1 { font-size: 1.9em; line-height: 1.25; }
.sui-rich-editor h2 { font-size: 1.55em; line-height: 1.3; }
.sui-rich-editor h3 { font-size: 1.28em; line-height: 1.35; }
.sui-rich-editor h4 { font-size: 1.1em; }
.sui-rich-editor blockquote { border-left: 3px solid var(--sui-rich-accent); padding: .2em 0 .2em 14px; color: var(--sui-rich-muted); }
.sui-rich-editor pre { background: rgba(0,0,0,.22); border: 1px solid var(--sui-rich-border); border-radius: 8px; padding: 12px; overflow: auto; font-family: ui-monospace, Consolas, monospace; font-size: 13px; }
.sui-rich-editor code { font-family: ui-monospace, Consolas, monospace; font-size: .92em; background: rgba(127,127,127,.16); border-radius: 4px; padding: .1em .3em; }
.sui-rich-editor hr { border: 0; border-top: 1px solid var(--sui-rich-border); margin: 1.2em 0; }
.sui-rich-editor img { max-width: 100%; height: auto; border-radius: 6px; }
.sui-rich-editor figure { text-align: center; }
.sui-rich-editor figcaption { color: var(--sui-rich-muted); font-size: .85em; margin-top: 4px; }
.sui-rich-editor table { width: auto; max-width: 100%; border-collapse: collapse; }
.sui-rich-editor th, .sui-rich-editor td { border: 1px solid var(--sui-rich-border); padding: 6px 8px; vertical-align: top; min-width: 40px; }
.sui-rich-editor th { background: var(--sui-rich-hover); text-align: left; }
.sui-rich-editor a { color: var(--sui-rich-accent); text-decoration: underline; }
.sui-rich-editor ul, .sui-rich-editor ol { padding-left: 1.6em; }
.sui-rich-props { width: 210px; border-left: 1px solid var(--sui-rich-border); padding: 12px; display: flex; flex-direction: column; gap: 8px; background: var(--sui-rich-toolbar-bg); }
.sui-rich-props-title { font-size: 12px; font-weight: 700; color: var(--sui-rich-muted); text-transform: uppercase; letter-spacing: .05em; }
.sui-rich-prop { width: 100%; }
.sui-rich-prop-btn { border: 1px solid var(--sui-rich-border); border-radius: 6px; background: transparent; color: var(--sui-rich-fg); padding: 5px 8px; font: inherit; font-size: 12px; text-align: left; cursor: pointer; }
.sui-rich-prop-btn:hover { background: var(--sui-rich-hover); }
.sui-rich-status { display: flex; align-items: center; flex-wrap: wrap; gap: 10px; padding: 5px 10px; border-top: 1px solid var(--sui-rich-border); background: var(--sui-rich-toolbar-bg); color: var(--sui-rich-muted); font-size: 11px; }
.sui-rich-stats { margin-right: auto; }
.sui-rich-saved { color: var(--sui-rich-muted); }
.sui-rich-zoom-btn { width: 22px; height: 22px; border: 1px solid var(--sui-rich-border); border-radius: 5px; background: transparent; color: var(--sui-rich-fg); cursor: pointer; line-height: 1; }
.sui-rich-context { position: fixed; z-index: 900; display: flex; gap: 2px; padding: 4px; background: var(--sui-rich-bg); border: 1px solid var(--sui-rich-border); border-radius: 8px; box-shadow: 0 12px 30px rgba(0,0,0,.3); }
.sui-rich-context[hidden] { display: none; }
.sui-rich-tablebar { display: flex; flex-wrap: wrap; gap: 2px; padding: 4px 8px; background: var(--sui-rich-toolbar-bg); border-bottom: 1px solid var(--sui-rich-border); position: relative; z-index: 4; }
.sui-rich-tablebar[hidden] { display: none; }
.sui-rich-tablebar.floating { position: fixed; left: 24px; top: 90px; z-index: 950; border: 1px solid var(--sui-rich-border); border-radius: 8px; box-shadow: 0 12px 30px rgba(0,0,0,.3); }
.sui-rich-tablebar .sui-rich-group { border-right: 1px solid var(--sui-rich-border); padding-right: 3px; }
.sui-rich-tablebar .sui-rich-group:last-child { border-right: 0; padding-right: 0; }
.sui-rich-tablebar-handle { display: inline-flex; align-items: center; justify-content: center; width: 18px; color: var(--sui-rich-muted); cursor: grab; user-select: none; font-size: 12px; }
.sui-rich-tablebar-handle:active { cursor: grabbing; }
.sui-rich-editor td.sui-rich-cell-selected, .sui-rich-editor th.sui-rich-cell-selected { outline: 2px solid var(--sui-rich-accent); outline-offset: -2px; background: color-mix(in srgb, var(--sui-rich-accent) 16%, transparent); }
.sui-rich-editor table.sui-rich-table-selected { outline: 2px dashed var(--sui-rich-accent); outline-offset: 3px; }
.sui-rich[data-sui-rich-theme="light"] .sui-rich-editor pre { background: #f3f5fa; }
.sui-rich[data-sui-rich-theme="light"] .sui-rich-editor code { background: #eef1f7; }
@media (max-width: 720px) {
  .sui-rich-body { flex-direction: column; }
  .sui-rich-props { width: 100%; border-left: 0; border-top: 1px solid var(--sui-rich-border); }
  .sui-rich-editor { padding: 16px; }
}
.sui-rich[data-sui-rich-template="article"] .sui-rich-editor { max-width: 780px; margin: 0 auto; width: 100%; font-family: Georgia, 'Times New Roman', serif; font-size: 17px; line-height: 1.75; }
.sui-rich[data-sui-rich-template="letter"] .sui-rich-editor { max-width: 680px; margin: 0 auto; width: 100%; font-family: Georgia, 'Times New Roman', serif; font-size: 16px; line-height: 1.8; }
.sui-rich[data-sui-rich-template="report"] .sui-rich-editor { max-width: 960px; margin: 0 auto; width: 100%; font-family: system-ui, sans-serif; font-size: 15px; line-height: 1.6; }
.sui-rich[data-sui-rich-template="compact"] .sui-rich-editor { padding: 12px 14px; font-size: 14px; line-height: 1.5; }
@media (prefers-reduced-motion: reduce) {
  .sui-rich-btn, .sui-rich-menu-btn, .sui-rich-menu-item, .sui-rich-prop-btn { transition: none !important; }
}
`

// progUsesRichText reports whether the program references the editor tag or
// the richtext API, so the runtime is emitted only when needed.
func progUsesRichText(p *Prog) bool {
	if p == nil {
		return false
	}
	var exprUses func(*Expr) bool
	exprUses = func(e *Expr) bool {
		if e == nil {
			return false
		}
		if e.kind == "member" && e.obj != nil && e.obj.kind == "ident" && e.obj.name == "richtext" {
			return true
		}
		if exprUses(e.obj) || exprUses(e.target) || exprUses(e.operand) || exprUses(e.l) || exprUses(e.r) || exprUses(e.cond) || exprUses(e.idx) || exprUses(e.lo) || exprUses(e.hi) {
			return true
		}
		for _, a := range e.args {
			if exprUses(a) {
				return true
			}
		}
		for _, a := range e.elements {
			if exprUses(a) {
				return true
			}
		}
		for _, v := range e.vals {
			if exprUses(v) {
				return true
			}
		}
		return false
	}
	var stmtsUse func([]*Stmt) bool
	stmtsUse = func(stmts []*Stmt) bool {
		for _, st := range stmts {
			if st == nil {
				continue
			}
			if exprUses(st.init) || exprUses(st.rhs) || exprUses(st.lvalue) || exprUses(st.cond) || exprUses(st.iterable) || exprUses(st.value) {
				return true
			}
			for _, e := range st.values {
				if exprUses(e) {
					return true
				}
			}
			if stmtsUse(st.then_) || stmtsUse(st.else_) || stmtsUse(st.body) || stmtsUse(st.default_) {
				return true
			}
			for _, c := range st.cases {
				if exprUses(c.expr) || stmtsUse(c.body) {
					return true
				}
			}
		}
		return false
	}
	var nodesUse func([]*ViewNode) bool
	nodesUse = func(nodes []*ViewNode) bool {
		for _, n := range nodes {
			if n == nil {
				continue
			}
			if n.kind == "element" && n.elem != nil {
				if n.elem.tag == "richtext" {
					return true
				}
				for _, a := range n.elem.attrs {
					if exprUses(a.expr) {
						return true
					}
				}
				if nodesUse(childPtrs(n.elem.children)) {
					return true
				}
			}
			if exprUses(n.cond) || exprUses(n.iterable) || exprUses(n.key) {
				return true
			}
			if nodesUse(n.then_) || nodesUse(n.else_) || nodesUse(n.body) {
				return true
			}
		}
		return false
	}
	for _, e := range p.consts {
		if exprUses(e.init) {
			return true
		}
	}
	for _, e := range p.computed {
		if exprUses(e.init) {
			return true
		}
	}
	for _, e := range p.states {
		if exprUses(e.init) {
			return true
		}
	}
	for _, f := range p.funcs {
		if stmtsUse(f.body) {
			return true
		}
		for _, param := range f.params {
			if exprUses(param.dflt) {
				return true
			}
		}
	}
	for _, v := range p.views {
		if nodesUse(childPtrs(v.body)) {
			return true
		}
		for _, f := range v.methods {
			if stmtsUse(f.body) {
				return true
			}
		}
	}
	return false
}
