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
  var tableGroupSeq = 0;
  var MAX_NODES = 10000;
  var DEFAULT_DOC = { type: 'doc', children: [{ type: 'paragraph', children: [] }] };
  var defaultToolbar = 'undo,redo,copy,cut,paste|styleSelect,fontFamily,fontSize,fontSizeUp,fontSizeDown,bold,italic,underline,strike,textColor,highlight,superscript,subscript,code,clearFormat,copyFormat|alignLeft,alignCenter,alignRight,justify,indent,outdent,paragraphSettings,bulletList,orderedList,blockquote,codeBlock|link,unlink,image,table,symbol,hr,pageBreak|pageSettings,printPreview,print,find|zoom';
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
    clearFormat: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 15l7-7 6 6-7 7H6zM12 6l3-3 6 6-3 3M4 21h16"/></svg>',
    copyFormat: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M9 4h6v4H9zM7 8h10v4l-2 2v6H9v-6l-2-2z"/><path d="M12 20v2"/></svg>',
    printPreview: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 12s3.5-6 9-6 9 6 9 6-3.5 6-9 6-9-6-9-6z"/><circle cx="12" cy="12" r="2.4"/></svg>',
    paragraphSettings: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 6h9M5 10h7M5 14h9M5 18h5"/><path d="M17 13l3 3-3 3M20 16h-5"/></svg>',
    alignLeft: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M4 10h11M4 14h16M4 18h9"/></svg>',
    alignCenter: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M7 10h10M4 14h16M8 18h8"/></svg>',
    alignRight: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M9 10h11M4 14h16M11 18h9"/></svg>',
    justify: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M4 10h16M4 14h16M4 18h16"/></svg>',
    indent: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M10 5h11M10 10h11M10 15h11M10 20h11M3 8v8M3 12h5M6 9l3 3-3 3"/></svg>',
    outdent: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M10 5h11M10 10h11M10 15h11M10 20h11M8 8v8M8 12H3M5 9l-3 3 3 3"/></svg>',
    bulletList: '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="4.5" cy="6.5" r="1.9"/><circle cx="4.5" cy="12" r="1.9"/><circle cx="4.5" cy="17.5" r="1.9"/><path d="M9 6.5h12M9 12h12M9 17.5h12"/></svg>',
    orderedList: '<svg viewBox="0 0 24 24" aria-hidden="true"><text x="2.2" y="8.2" font-size="7.5" font-family="system-ui, sans-serif" style="fill:currentColor;stroke:none">1</text><text x="2.2" y="18.2" font-size="7.5" font-family="system-ui, sans-serif" style="fill:currentColor;stroke:none">2</text><path d="M9.5 6.5h11.5M9.5 16.5h11.5"/></svg>',
    link: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M10 13a5 5 0 0 0 7 0l2-2a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-2 2a5 5 0 0 0 7 7l1-1"/></svg>',
    unlink: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 4l16 16M10 13a5 5 0 0 0 7 0M14 11a5 5 0 0 0-7 0"/></svg>',
    image: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="8" cy="9" r="1.6"/><path d="M4 17l5-5 4 4 3-3 4 4"/></svg>',
    imageSettings: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="8" cy="9" r="1.4"/><path d="M4 17l5-5 4 4 3-3 4 4"/><path d="M17 14l2 2-2 2M15 16h4"/></svg>',
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
    pageBreak: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M7 3h7l3 3v15H7z"/><path d="M3 13h18" stroke-dasharray="2 2"/></svg>',
    blockquote: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 7h5v5H6c0 2 2 3 4 3M14 7h5v5h-4c0 2 2 3 4 3"/></svg>',
    codeBlock: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 7 3 12l5 5M16 7l5 5-5 5M13 5l-2 14"/></svg>',
    code: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M9 8 5 12l4 4M15 8l4 4-4 4"/></svg>',
    fontSizeUp: '<span aria-hidden="true">A<sup>+</sup></span>',
    fontSizeDown: '<span aria-hidden="true">A<sup>\u2212</sup></span>',
    textColor: '<span class="sui-rich-color-glyph" aria-hidden="true">A</span>',
    highlight: '<span class="sui-rich-color-glyph sui-rich-color-glyph-highlight" aria-hidden="true">A</span>',
    paragraph: '<span aria-hidden="true">P</span>',
    pageSettings: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 3h9l4 4v14H6z"/><path d="M9 8h5M9 12h7M9 16h4"/></svg>',
    documentSettings: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 3h9l4 4v14H6z"/><path d="M9 8h7M9 12h7M9 16h4"/><circle cx="17" cy="17" r="2.2"/></svg>',
    print: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M7 8V4h10v4M7 17H4V9h16v8h-3"/><rect x="7" y="13" width="10" height="7"/></svg>',
    find: '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="10.5" cy="10.5" r="6"/><path d="M15 15l5 5"/></svg>'
  };
  var ICON_SETS = [
    ['line', 'Line icons'],
    ['bold', 'Bold line icons'],
    ['duotone', 'Duotone icons'],
    ['blocks', 'Soft block icons']
  ];
  var TOOL_LABELS = {
    undo: 'Undo', redo: 'Redo', copy: 'Copy', cut: 'Cut', paste: 'Paste',
    bold: 'Bold', italic: 'Italic', underline: 'Underline', strike: 'Strikethrough',
    superscript: 'Superscript', subscript: 'Subscript', fontSizeUp: 'Increase font size', fontSizeDown: 'Decrease font size', code: 'Inline code',
    codeBlock: 'Code block', clearFormat: 'Clear formatting', copyFormat: 'Copy formatting',
    alignLeft: 'Align left', alignCenter: 'Align center', alignRight: 'Align right', justify: 'Justify',
    indent: 'Increase paragraph indent', outdent: 'Decrease paragraph indent',
    bulletList: 'Bulleted list', orderedList: 'Numbered list',
    blockquote: 'Quote', link: 'Insert link', unlink: 'Remove link',
    image: 'Insert image', imageSettings: 'Image settings', table: 'Insert table', symbol: 'Insert symbol',
    hr: 'Horizontal rule', pageBreak: 'Page break',
    documentSettings: 'Document settings', pageSettings: 'Page settings', printPreview: 'Print preview',
    print: 'Print / Save as PDF', find: 'Find and replace',
    paragraphSettings: 'Paragraph settings',
    zoomIn: 'Zoom in', zoomOut: 'Zoom out', zoomReset: 'Reset zoom to 100%'
  };
  var MENUS = {
    file: ['New document|newDoc', 'Open file...|openFile',
      { label: 'Save as', items: ['Plain text (.txt)|saveTXT', 'RTF (.rtf)|saveRTF', 'Word (.doc, RTF)|saveDOC', 'Word (.docx)|saveDOCX', 'OpenDocument (.odt)|saveODT', 'PDF (.pdf)|savePDF', 'HTML (.html)|saveHTML', 'Markdown (.md)|saveMarkdown', 'JSON (.json)|saveJSON', 'XML (.xml)|saveXML'] },
      'Print preview|printPreview', 'Print / Save as PDF|print',
      { label: 'Import', items: ['JSON|importJSON', 'HTML|importHTML', 'Markdown|importMarkdown', 'XML|importXML'] },
      { label: 'Export', items: ['JSON|exportJSON', 'HTML|exportHTML', 'Markdown|exportMarkdown', 'XML|exportXML'] }],
    edit: ['Undo|undo', 'Redo|redo', 'Copy|copy', 'Cut|cut', 'Paste|paste', 'Select all|selectAll', 'Find and replace|find'],
    insert: ['Link|link', 'Image|image', 'Table|table', 'Horizontal rule|hr', 'Page break|pageBreak'],
    preferences: ['Document settings|documentSettings', 'Page settings|pageSettings', 'Header and footer...|headerFooter',
      { label: 'Icon style', items: ['Line icons|iconSetLine', 'Bold line icons|iconSetBold', 'Duotone icons|iconSetDuotone', 'Soft block icons|iconSetBlocks'] }]
  };
  var MENU_LABELS = { file: 'File', edit: 'Edit', insert: 'Insert', preferences: 'Preferences' };

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
      if (node.nodeValue) out.push({ type: 'text', text: String(node.nodeValue).replace(/\u00AD/g, ''), marks: marks.slice() });
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
      spaceAfter: cssLength(style.marginBottom, ''),
      hyphens: style.hyphens || style.webkitHyphens || '',
      tabSize: style.tabSize || '',
      lineRule: node.getAttribute && node.getAttribute('data-line-rule') ? node.getAttribute('data-line-rule') : ''
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
      if (tag === 'div' && (node.getAttribute('data-sui-page-break') !== null || node.classList.contains('sui-rich-pagebreak'))) { flush(); out.push({ type: 'page_break' }); continue; }
      if ((tag === 'div' || tag === 'section' || tag === 'article' || tag === 'main' || tag === 'header' || tag === 'footer') && node.querySelector && node.querySelector('p,div,table,ul,ol,pre,blockquote,h1,h2,h3,h4,h5,h6')) {
        flush();
        var innerBlocks = blocksFrom(node, depth + 1);
        for (var ib = 0; ib < innerBlocks.length; ib++) out.push(innerBlocks[ib]);
        continue;
      }
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
        var codeBlock = { type: 'code_block', text: node.textContent || '' };
        if (node.getAttribute && node.getAttribute('data-line-numbers') !== null) codeBlock.lineNumbers = true;
        if (node.getAttribute && node.getAttribute('data-code-language')) codeBlock.codeLanguage = codeLanguage(node.getAttribute('data-code-language'));
        var preStyle = blockStyle(node);
        for (var pk in preStyle) if (hasOwn(preStyle, pk) && preStyle[pk]) codeBlock[pk] = preStyle[pk];
        var codeHost = null;
        for (var chn = 0; chn < node.children.length; chn++) {
          if (node.children[chn].tagName && node.children[chn].tagName.toLowerCase() === 'code') { codeHost = node.children[chn]; break; }
        }
        if (!codeHost) codeHost = node;
        var preInline = [];
        for (var pnc = 0; pnc < codeHost.childNodes.length; pnc++) inlineFrom(codeHost.childNodes[pnc], [], preInline, depth + 1);
        if (preInline.length) codeBlock.children = preInline;
        out.push(codeBlock);
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
        var table = { type: 'table', rows: [], header: false, border: cssColor(node.style.borderColor || '') || '', borderWidth: cssLength(node.getAttribute('data-table-border-width') || '', ''), padding: cssLength(node.getAttribute('data-table-padding') || '', ''), width: cssLength(node.style.width, ''), height: cssLength(node.style.height, ''), align: tableAlign };
        var tableColEls = node.querySelectorAll('colgroup > col');
        if (tableColEls.length) {
          table.colWidths = [];
          for (var tcl = 0; tcl < tableColEls.length; tcl++) table.colWidths.push(cssLength(tableColEls[tcl].style.width, ''));
        }
        var tableGroup = node.getAttribute('data-sui-table-group');
        if (tableGroup) table._tableGroup = tableGroup;
        if (node.getAttribute('data-table-borders') !== null) table.borders = tableBorderFlags(node);
        function cellPaddingOf(cellNode) {
          var shorthand = String(cellNode.style.padding || '').trim();
          if (shorthand) {
            var parts = shorthand.split(/\s+/);
            var chosen = parts.length === 1 ? parts[0] : (parts[3] || parts[1] || parts[0]);
            var value = cssLength(chosen, '');
            if (value) return value;
          }
          var left = cssLength(cellNode.style.paddingLeft || cellNode.style.paddingRight || '', '');
          return left;
        }
        for (var r = 0; r < node.rows.length; r++) {
          var row = { type: 'table_row', height: cssLength(node.rows[r].style.height, ''), cells: [] };
          for (var ci = 0; ci < node.rows[r].cells.length; ci++) {
            var cell = node.rows[r].cells[ci];
            var cellInline = [];
            for (var cc = 0; cc < cell.childNodes.length; cc++) inlineFrom(cell.childNodes[cc], [], cellInline, depth + 1);
            var nestedTable = false;
            for (var nc = 0; nc < cell.childNodes.length; nc++) {
              var cn = cell.childNodes[nc];
              if (cn.nodeType === 1 && (cn.tagName.toLowerCase() === 'table' || (cn.querySelector && cn.querySelector('table')))) { nestedTable = true; break; }
            }
            var cellBackground = cssColor(cell.style.backgroundColor || '') || cssColor(cell.getAttribute('bgcolor') || '') || '';
            var cellData = { type: 'table_cell', header: cell.tagName.toLowerCase() === 'th', colspan: Number(cell.getAttribute('colspan') || 1), rowspan: Number(cell.getAttribute('rowspan') || 1), align: cell.style.textAlign || '', width: cssLength(cell.style.width, ''), background: cellBackground, children: nestedTable ? [] : cellInline };
            if (nestedTable) { cellData.blocks = blocksFrom(cell, depth + 1); if (!cellData.blocks.length) cellData.blocks = [{ type: 'paragraph', children: cellInline }]; }
            if (!table.padding) { var cellPadValue = cellPaddingOf(cell); if (cellPadValue) table.padding = cellPadValue; }
            if (!table.borderWidth) { var cellBorderWidth = cssLength(cell.style.borderTopWidth || cell.style.borderWidth || '', ''); if (cellBorderWidth && parseFloat(cellBorderWidth) > 0) { table.borderWidth = cellBorderWidth; table.border = table.border || cssColor(cell.style.borderTopColor || cell.style.borderColor || '') || ''; } }
            row.cells.push(cellData);
          }
          table.rows.push(row);
        }
        if (!table.border && node.style.borderTopColor) table.border = cssColor(node.style.borderTopColor) || '';
        if (!table.borderWidth && node.style.borderTopWidth) table.borderWidth = cssLength(node.style.borderTopWidth, '');
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
    if (node.hyphens) {
      element.style.hyphens = node.hyphens;
      element.style.webkitHyphens = node.hyphens;
    }
    if (node.tabSize) element.style.tabSize = node.tabSize;
    if (node.lineRule) element.setAttribute('data-line-rule', String(node.lineRule));
    else if (element.removeAttribute) element.removeAttribute('data-line-rule');
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
        if (node.align) {
          img.style.display = 'block';
          img.style.textAlign = node.align;
          img.style.marginLeft = (node.align === 'right' || node.align === 'center') ? 'auto' : '0';
          img.style.marginRight = (node.align === 'left' || node.align === 'center') ? 'auto' : '0';
        }
        parent.appendChild(img);
      }
    }
  }

  function tableBorderFlags(table) {
    var attr = table.getAttribute && table.getAttribute('data-table-borders');
    if (attr === null) return ['top', 'right', 'bottom', 'left', 'insideH', 'insideV'];
    return attr.split(',').filter(function (v) { return v; });
  }
  function tableGridColumns(table) {
    var rows = table.rows, max = 0;
    for (var r = 0; r < rows.length; r++) {
      var total = 0;
      for (var c = 0; c < rows[r].cells.length; c++) total += Math.max(1, rows[r].cells[c].colSpan || 1);
      if (total > max) max = total;
    }
    return max;
  }
  function applyTableBorderSides(table, flags, width, color) {
    if (!table || !table.rows) return;
    var totalCols = tableGridColumns(table) || 1;
    function on(name) { return flags.indexOf(name) >= 0; }
    for (var r = 0; r < table.rows.length; r++) {
      var row = table.rows[r], col = 0;
      for (var c = 0; c < row.cells.length; c++) {
        var cell = row.cells[c];
        var span = Math.max(1, cell.colSpan || 1), rowspan = Math.max(1, cell.rowSpan || 1);
        var edges = {
          Top: r === 0 ? on('top') : on('insideH'),
          Bottom: (r + rowspan >= table.rows.length) ? on('bottom') : on('insideH'),
          Left: col === 0 ? on('left') : on('insideV'),
          Right: (col + span >= totalCols) ? on('right') : on('insideV')
        };
        for (var side in edges) {
          if (!Object.prototype.hasOwnProperty.call(edges, side)) continue;
          if (edges[side]) {
            cell.style['border' + side + 'Width'] = width ? width + 'px' : (width === '' ? '' : '');
            cell.style['border' + side + 'Style'] = 'solid';
            cell.style['border' + side + 'Color'] = color || '';
          } else {
            cell.style['border' + side + 'Width'] = '0px';
            cell.style['border' + side + 'Style'] = 'none';
            cell.style['border' + side + 'Color'] = '';
          }
        }
        col += span;
      }
    }
  }
  // TeX hyphenation patterns: hyph-it 4.9 and hyph-en-us (LPPL/MIT).
  var HYPHEN_PATTERNS = {"it":".a3p2n\n.anti1\n.anti3m2n\n.bio1\n.ca4p3s\n.circu2m1\n.contro1\n.di2s3cine\n.e2x1eu\n.fran2k3\n.free3\n.li3p2sa\n.narco1\n.opto1\n.orto3p2\n.para1\n.ph2l\n.ph2r\n.poli3p2\n.pre1\n.p2s\n.re1i2scr\n.sha2re3\n.tran2s3c\n.tran2s3d\n.tran2s3l\n.tran2s3n\n.tran2s3p\n.tran2s3r\n.tran2s3t\n.su2b3lu\n.su2b3r\n.wa2g3n\n.wel2t1\n2'2\n2’2\na1ia\na1ie\na1io\na1iu\na1uo\na1ya\n2at.\ne1iu\ne2w\no1ia\no1ie\no1io\no1iu\n1b\n2bb\n2bc\n2bd\n2bf\n2bm\n2bn\n2bp\n2bs\n2bt\n2bv\nb2l\nb2r\n2b.\n2b'\n2b’\n1c\n2cb\n2cc\n2cd\n2cf\n2ck\n2cm\n2cn\n2cq\n2cs\n2ct\n2cz\n2chh\nc2h\n2ch.\n2ch'.\n2ch’.\n2ch''.\n2ch’’.\n2chb\nch2r\n2chn\nc2l\nc2r\n2c.\n2c'\n2c’\n.c2\n1d\n2db\n2dd\n2dg\n2dl\n2dm\n2dn\n2dp\nd2r\n2ds\n2dt\n2dv\n2dw\n2d.\n2d'\n2d’\n.d2\n1f\n2fb\n2fg\n2ff\n2fn\nf2l\nf2r\n2fs\n2ft\n2f.\n2f'\n2f’\n1g\n2gb\n2gd\n2gf\n2gg\ng2h\ng2l\n2gm\ng2n\n2gp\ng2r\n2gs\n2gt\n2gv\n2gw\n2gz\n2gh2t\n2g.\n2g'\n2g’\n.h2\n1h\n2hb\n2hd\n2hh\nhi3p2n\nh2l\n2hm\n2hn\n2hr\n2hv\n2h.\n2h'\n2h’\n.j2\n1j\n2j.\n2j'\n2j’\n.k2\n1k\n2kg\n2kf\nk2h\n2kk\nk2l\n2km\nk2r\n2ks\n2kt\n2k.\n2k'\n2k’\n1l\n2lb\n2lc\n2ld\n2l3f2\n2lg\nl2h\nl2j\n2lk\n2ll\n2lm\n2ln\n2lp\n2lq\n2lr\n2ls\n2lt\n2lv\n2lw\n2lz\n2l.\n2l'.\n2l’.\n2l''\n2l’’\n1m\n2mb\n2mc\n2mf\n2ml\n2mm\n2mn\n2mp\n2mq\n2mr\n2ms\n2mt\n2mv\n2mw\n2m.\n2m'\n2m’\n1n\n2nb\n2nc\n2nd\n2nf\n2ng\n2nk\n2nl\n2nm\n2nn\n2np\n2nq\n2nr\n2ns\nn2s3fer\n2nt\n2nv\n2nz\nn2g3n\n2nheit\n2n.\n2n'\n2n’\n1p\n2pd\np2h\np2l\n2pn\n3p2ne\n2pp\np2r\n2ps\n3p2sic\n2pt\n2pz\n2p.\n2p'\n2p’\n1q\n2qq\n2q.\n2q'\n2q’\n1r\n2rb\n2rc\n2rd\n2rf\nr2h\n2rg\n2rk\n2rl\n2rm\n2rn\n2rp\n2rq\n2rr\n2rs\n2rt\nr2t2s3\n2rv\n2rx\n2rw\n2rz\n2r.\n2r'\n2r’\n1s2\n2shm\n2sh.\n2sh'\n2sh’\n2s3s\ns4s3m\n2s3p2n\n2stb\n2stc\n2std\n2stf\n2stg\n2stm\n2stn\n2stp\n2sts\n2stt\n2stv\n2sz\n4s.\n4s'.\n4s’.\n4s''\n4s’’\n.t2\n1t\n2tb\n2tc\n2td\n2tf\n2tg\nt2h\n2th.\nt2l\n2tm\n2tn\n2tp\nt2r\nt2s\n3t2sch\n2tt\nt2t3s\n2tv\n2tw\nt2z\n2tzk\ntz2s\n2t.\n2t'.\n2t’.\n2t''\n2t’’\n1v\n2vc\nv2l\nv2r\n2vv\n2v.\n2v'.\n2v’.\n2v''\n2v’’\n1w\nw2h\nwa2r\n2w1y\n2w.\n2w'\n2w’\n1x\n2xb\n2xc\n2xf\n2xh\n2xm\n2xp\n2xt\n2xw\n2x.\n2x'\n2x’\ny1ou\ny1i\n1z\n2zb\n2zd\n2zl\n2zn\n2zp\n2zt\n2zs\n2zv\n2zz\n2z.\n2z'.\n2z’.\n2z''\n2z’’\n.z2","en":".ach4\n.ad4der\n.af1t\n.al3t\n.am5at\n.an5c\n.ang4\n.ani5m\n.ant4\n.an3te\n.anti5s\n.ar5s\n.ar4tie\n.ar4ty\n.as3c\n.as1p\n.as1s\n.aster5\n.atom5\n.au1d\n.av4i\n.awn4\n.ba4g\n.ba5na\n.bas4e\n.ber4\n.be5ra\n.be3sm\n.be5sto\n.bri2\n.but4ti\n.cam4pe\n.can5c\n.capa5b\n.car5ol\n.ca4t\n.ce4la\n.ch4\n.chill5i\n.ci2\n.cit5r\n.co3e\n.co4r\n.cor5ner\n.de4moi\n.de3o\n.de3ra\n.de3ri\n.des4c\n.dictio5\n.do4t\n.du4c\n.dumb5\n.earth5\n.eas3i\n.eb4\n.eer4\n.eg2\n.el5d\n.el3em\n.enam3\n.en3g\n.en3s\n.eq5ui5t\n.er4ri\n.es3\n.eu3\n.eye5\n.fes3\n.for5mer\n.ga2\n.ge2\n.gen3t4\n.ge5og\n.gi5a\n.gi4b\n.go4r\n.hand5i\n.han5k\n.he2\n.hero5i\n.hes3\n.het3\n.hi3b\n.hi3er\n.hon5ey\n.hon3o\n.hov5\n.id4l\n.idol3\n.im3m\n.im5pin\n.in1\n.in3ci\n.ine2\n.in2k\n.in3s\n.ir5r\n.is4i\n.ju3r\n.la4cy\n.la4m\n.lat5er\n.lath5\n.le2\n.leg5e\n.len4\n.lep5\n.lev1\n.li4g\n.lig5a\n.li2n\n.li3o\n.li4t\n.mag5a5\n.mal5o\n.man5a\n.mar5ti\n.me2\n.mer3c\n.me5ter\n.mis1\n.mist5i\n.mon3e\n.mo3ro\n.mu5ta\n.muta5b\n.ni4c\n.od2\n.odd5\n.of5te\n.or5ato\n.or3c\n.or1d\n.or3t\n.os3\n.os4tl\n.oth3\n.out3\n.ped5al\n.pe5te\n.pe5tit\n.pi4e\n.pio5n\n.pi2t\n.pre3m\n.ra4c\n.ran4t\n.ratio5na\n.ree2\n.re5mit\n.res2\n.re5stat\n.ri4g\n.rit5u\n.ro4q\n.ros5t\n.row5d\n.ru4d\n.sci3e\n.self5\n.sell5\n.se2n\n.se5rie\n.sh2\n.si2\n.sing4\n.st4\n.sta5bl\n.sy2\n.ta4\n.te4\n.ten5an\n.th2\n.ti2\n.til4\n.tim5o5\n.ting4\n.tin5k\n.ton4a\n.to4p\n.top5i\n.tou5s\n.trib5ut\n.un1a\n.un3ce\n.under5\n.un1e\n.un5k\n.un5o\n.un3u\n.up3\n.ure3\n.us5a\n.ven4de\n.ve5ra\n.wil5i\n.ye4\n4ab.\na5bal\na5ban\nabe2\nab5erd\nabi5a\nab5it5ab\nab5lat\nab5o5liz\n4abr\nab5rog\nab3ul\na4car\nac5ard\nac5aro\na5ceou\nac1er\na5chet\n4a2ci\na3cie\nac1in\na3cio\nac5rob\nact5if\nac3ul\nac4um\na2d\nad4din\nad5er.\n2adi\na3dia\nad3ica\nadi4er\na3dio\na3dit\na5diu\nad4le\nad3ow\nad5ran\nad4su\n4adu\na3duc\nad5um\nae4r\naeri4e\na2f\naff4\na4gab\naga4n\nag5ell\nage4o\n4ageu\nag1i\n4ag4l\nag1n\na2go\n3agog\nag3oni\na5guer\nag5ul\na4gy\na3ha\na3he\nah4l\na3ho\nai2\na5ia\na3ic.\nai5ly\na4i4n\nain5in\nain5o\nait5en\na1j\nak1en\nal5ab\nal3ad\na4lar\n4aldi\n2ale\nal3end\na4lenti\na5le5o\nal1i\nal4ia.\nali4e\nal5lev\n4allic\n4alm\na5log.\na4ly.\n4alys\n5a5lyst\n5alyt\n3alyz\n4ama\nam5ab\nam3ag\nama5ra\nam5asc\na4matis\na4m5ato\nam5era\nam3ic\nam5if\nam5ily\nam1in\nami4no\na2mo\na5mon\namor5i\namp5en\na2n\nan3age\n3analy\na3nar\nan3arc\nanar4i\na3nati\n4and\nande4s\nan3dis\nan1dl\nan4dow\na5nee\na3nen\nan5est.\na3neu\n2ang\nang5ie\nan1gl\na4n1ic\na3nies\nan3i3f\nan4ime\na5nimi\na5nine\nan3io\na3nip\nan3ish\nan3it\na3niu\nan4kli\n5anniz\nano4\nan5ot\nanoth5\nan2sa\nan4sco\nan4sn\nan2sp\nans3po\nan4st\nan4sur\nantal4\nan4tie\n4anto\nan2tr\nan4tw\nan3ua\nan3ul\na5nur\n4ao\napar4\nap5at\nap5ero\na3pher\n4aphi\na4pilla\nap5illar\nap3in\nap3ita\na3pitu\na2pl\napoc5\nap5ola\napor5i\napos3t\naps5es\na3pu\naque5\n2a2r\nar3act\na5rade\nar5adis\nar3al\na5ramete\naran4g\nara3p\nar4at\na5ratio\nar5ativ\na5rau\nar5av4\naraw4\narbal4\nar4chan\nar5dine\nar4dr\nar5eas\na3ree\nar3ent\na5ress\nar4fi\nar4fl\nar1i\nar5ial\nar3ian\na3riet\nar4im\nar5inat\nar3io\nar2iz\nar2mi\nar5o5d\na5roni\na3roo\nar2p\nar3q\narre4\nar4sa\nar2sh\n4as.\nas4ab\nas3ant\nashi4\na5sia.\na3sib\na3sic\n5a5si4t\nask3i\nas4l\na4soc\nas5ph\nas4sh\nas3ten\nas1tr\nasur5a\na2ta\nat3abl\nat5ac\nat3alo\nat5ap\nate5c\nat5ech\nat3ego\nat3en.\nat3era\nater5n\na5terna\nat3est\nat5ev\n4ath\nath5em\na5then\nat4ho\nath5om\n4ati.\na5tia\nat5i5b\nat1ic\nat3if\nation5ar\nat3itu\na4tog\na2tom\nat5omiz\na4top\na4tos\na1tr\nat5rop\nat4sk\nat4tag\nat5te\nat4th\na2tu\nat5ua\nat5ue\nat3ul\nat3ura\na2ty\nau4b\naugh3\nau3gu\nau4l2\naun5d\nau3r\nau5sib\naut5en\nau1th\na2va\nav3ag\na5van\nave4no\nav3era\nav5ern\nav5ery\nav1i\navi4er\nav3ig\nav5oc\na1vor\n3away\naw3i\naw4ly\naws4\nax4ic\nax4id\nay5al\naye4\nays4\nazi4er\nazz5i\n5ba.\nbad5ger\nba4ge\nbal1a\nban5dag\nban4e\nban3i\nbarbi5\nbari4a\nbas4si\n1bat\nba4z\n2b1b\nb2be\nb3ber\nbbi4na\n4b1d\n4be.\nbeak4\nbeat3\n4be2d\nbe3da\nbe3de\nbe3di\nbe3gi\nbe5gu\n1bel\nbe1li\nbe3lo\n4be5m\nbe5nig\nbe5nu\n4bes4\nbe3sp\nbe5str\n3bet\nbet5iz\nbe5tr\nbe3tw\nbe3w\nbe5yo\n2bf\n4b3h\nbi2b\nbi4d\n3bie\nbi5en\nbi4er\n2b3if\n1bil\nbi3liz\nbina5r4\nbin4d\nbi5net\nbi3ogr\nbi5ou\nbi2t\n3bi3tio\nbi3tr\n3bit5ua\nb5itz\nb1j\nbk4\nb2l2\nblath5\nb4le.\nblen4\n5blesp\nb3lis\nb4lo\nblun4t\n4b1m\n4b3n\nbne5g\n3bod\nbod3i\nbo4e\nbol3ic\nbom4bi\nbon4a\nbon5at\n3boo\n5bor.\n4b1ora\nbor5d\n5bore\n5bori\n5bos4\nb5ota\nboth5\nbo4to\nbound3\n4bp\n4brit\nbroth3\n2b5s2\nbsor4\n2bt\nbt4l\nb4to\nb3tr\nbuf4fer\nbu4ga\nbu3li\nbumi4\nbu4n\nbunt4i\nbu3re\nbus5ie\nbuss4e\n5bust\n4buta\n3butio\nb5uto\nb1v\n4b5w\n5by.\nbys4\n1ca\ncab3in\nca1bl\ncach4\nca5den\n4cag4\n2c5ah\nca3lat\ncal4la\ncall5in\n4calo\ncan5d\ncan4e\ncan4ic\ncan5is\ncan3iz\ncan4ty\ncany4\nca5per\ncar5om\ncast5er\ncas5tig\n4casy\nca4th\n4cativ\ncav5al\nc3c\nccha5\ncci4a\nccompa5\nccon4\nccou3t\n2ce.\n4ced.\n4ceden\n3cei\n5cel.\n3cell\n1cen\n3cenc\n2cen4e\n4ceni\n3cent\n3cep\nce5ram\n4cesa\n3cessi\nces5si5b\nces5t\ncet4\nc5e4ta\ncew4\n2ch\n4ch.\n4ch3ab\n5chanic\nch5a5nis\nche2\ncheap3\n4ched\nche5lo\n3chemi\nch5ene\nch3er.\nch3ers\n4ch1in\n5chine.\nch5iness\n5chini\n5chio\n3chit\nchi2z\n3cho2\nch4ti\n1ci\n3cia\nci2a5b\ncia5r\nci5c\n4cier\n5cific.\n4cii\nci4la\n3cili\n2cim\n2cin\nc4ina\n3cinat\ncin3em\nc1ing\nc5ing.\n5cino\ncion4\n4cipe\nci3ph\n4cipic\n4cista\n4cisti\n2c1it\ncit3iz\n5ciz\nck1\nck3i\n1c4l4\n4clar\nc5laratio\n5clare\ncle4m\n4clic\nclim4\ncly4\nc5n\n1co\nco5ag\ncoe2\n2cog\nco4gr\ncoi4\nco3inc\ncol5i\n5colo\ncol3or\ncom5er\ncon4a\nc4one\ncon3g\ncon5t\nco3pa\ncop3ic\nco4pl\n4corb\ncoro3n\ncos4e\ncov1\ncove4\ncow5a\ncoz5e\nco5zi\nc1q\ncras5t\n5crat.\n5cratic\ncre3at\n5cred\n4c3reta\ncre4v\ncri2\ncri5f\nc4rin\ncris4\n5criti\ncro4pl\ncrop5o\ncros4e\ncru4d\n4c3s2\n2c1t\ncta4b\nct5ang\nc5tant\nc2te\nc3ter\nc4ticu\nctim3i\nctu4r\nc4tw\ncud5\nc4uf\nc4ui\ncu5ity\n5culi\ncul4tis\n3cultu\ncu2ma\nc3ume\ncu4mi\n3cun\ncu3pi\ncu5py\ncur5a4b\ncu5ria\n1cus\ncuss4i\n3c4ut\ncu4tie\n4c5utiv\n4cutr\n1cy\ncze4\n1d2a\n5da.\n2d3a4b\ndach4\n4daf\n2dag\nda2m2\ndan3g\ndard5\ndark5\n4dary\n3dat\n4dativ\n4dato\n5dav4\ndav5e\n5day\nd1b\nd5c\nd1d4\n2de.\ndeaf5\ndeb5it\nde4bon\ndecan4\nde4cil\nde5com\n2d1ed\n4dee.\nde5if\ndeli4e\ndel5i5q\nde5lo\nd4em\n5dem.\n3demic\ndem5ic.\nde5mil\nde4mons\ndemor5\n1den\nde4nar\nde3no\ndenti5f\nde3nu\nde1p\nde3pa\ndepi4\nde2pu\nd3eq\nd4erh\n5derm\ndern5iz\nder5s\ndes2\nd2es.\nde1sc\nde2s5o\ndes3ti\nde3str\nde4su\nde1t\nde2to\nde1v\ndev3il\n4dey\n4d1f\nd4ga\nd3ge4t\ndg1i\nd2gy\nd1h2\n5di.\n1d4i3a\ndia5b\ndi4cam\nd4ice\n3dict\n3did\n5di3en\nd1if\ndi3ge\ndi4lato\nd1in\n1dina\n3dine.\n5dini\ndi5niz\n1dio\ndio5g\ndi4pl\ndir2\ndi1re\ndirt5i\ndis1\n5disi\nd4is3t\nd2iti\n1di1v\nd1j\nd5k2\n4d5la\n3dle.\n3dled\n3dles.\n4dless\n2d3lo\n4d5lu\n2dly\nd1m\n4d1n4\n1do\n3do.\ndo5de\n5doe\n2d5of\nd4og\ndo4la\ndoli4\ndo5lor\ndom5iz\ndo3nat\ndoni4\ndoo3d\ndop4p\nd4or\n3dos\n4d5out\ndo4v\n3dox\nd1p\n1dr\ndrag5on\n4drai\ndre4\ndrea5r\n5dren\ndri4b\ndril4\ndro4p\n4drow\n5drupli\n4dry\n2d1s2\nds4p\nd4sw\nd4sy\nd2th\n1du\nd1u1a\ndu2c\nd1uca\nduc5er\n4duct.\n4ducts\ndu5el\ndu4g\nd3ule\ndum4be\ndu4n\n4dup\ndu4pe\nd1v\nd1w\nd2y\n5dyn\ndy4se\ndys5p\ne1a4b\ne3act\nead1\nead5ie\nea4ge\nea5ger\nea4l\neal5er\neal3ou\neam3er\ne5and\near3a\near4c\near5es\near4ic\near4il\near5k\near2t\neart3e\nea5sp\ne3ass\neast3\nea2t\neat5en\neath3i\ne5atif\ne4a3tu\nea2v\neav3en\neav5i\neav5o\n2e1b\ne4bel.\ne4bels\ne4ben\ne4bit\ne3br\ne4cad\necan5c\necca5\ne1ce\nec5essa\nec2i\ne4cib\nec5ificat\nec5ifie\nec5ify\nec3im\neci4t\ne5cite\ne4clam\ne4clus\ne2col\ne4comm\ne4compe\ne4conc\ne2cor\nec3ora\neco5ro\ne1cr\ne4crem\nec4tan\nec4te\ne1cu\ne4cul\nec3ula\n2e2da\n4ed3d\ne4d1er\nede4s\n4edi\ne3dia\ned3ib\ned3ica\ned3im\ned1it\nedi5z\n4edo\ne4dol\nedon2\ne4dri\ne4dul\ned5ulo\nee2c\need3i\nee2f\neel3i\nee4ly\nee2m\nee4na\nee4p1\nee2s4\neest4\nee4ty\ne5ex\ne1f\ne4f3ere\n1eff\ne4fic\n5efici\nefil4\ne3fine\nef5i5nite\n3efit\nefor5es\ne4fuse.\n4egal\neger4\neg5ib\neg4ic\neg5ing\ne5git5\neg5n\ne4go.\ne4gos\neg1ul\ne5gur\n5egy\ne1h4\neher4\nei2\ne5ic\nei5d\neig2\nei5gl\ne3imb\ne3inf\ne1ing\ne5inst\neir4d\neit3e\nei3th\ne5ity\ne1j\ne4jud\nej5udi\neki4n\nek4la\ne1la\ne4la.\ne4lac\nelan4d\nel5ativ\ne4law\nelaxa4\ne3lea\nel5ebra\n5elec\ne4led\nel3ega\ne5len\ne4l1er\ne1les\nel2f\nel2i\ne3libe\ne4l5ic.\nel3ica\ne3lier\nel5igib\ne5lim\ne4l3ing\ne3lio\ne2lis\nel5ish\ne3liv3\n4ella\nel4lab\nello4\ne5loc\nel5og\nel3op.\nel2sh\nel4ta\ne5lud\nel5ug\ne4mac\ne4mag\ne5man\nem5ana\nem5b\ne1me\ne2mel\ne4met\nem3ica\nemi4e\nem5igra\nem1in2\nem5ine\nem3i3ni\ne4mis\nem5ish\ne5miss\nem3iz\n5emniz\nemo4g\nemoni5o\nem3pi\ne4mul\nem5ula\nemu3n\ne3my\nen5amo\ne4nant\nench4er\nen3dic\ne5nea\ne5nee\nen3em\nen5ero\nen5esi\nen5est\nen3etr\ne3new\nen5ics\ne5nie\ne5nil\ne3nio\nen3ish\nen3it\ne5niu\n5eniz\n4enn\n4eno\neno4g\ne4nos\nen3ov\nen4sw\nent5age\n4enthes\nen3ua\nen5uf\ne3ny.\n4en3z\ne5of\neo2g\ne4oi4\ne3ol\neop3ar\ne1or\neo3re\neo5rol\neos4\ne4ot\neo4to\ne5out\ne5ow\ne2pa\ne3pai\nep5anc\ne5pel\ne3pent\nep5etitio\nephe4\ne4pli\ne1po\ne4prec\nep5reca\ne4pred\nep3reh\ne3pro\ne4prob\nep4sh\nep5ti5b\ne4put\nep5uta\ne1q\nequi3l\ne4q3ui3s\ner1a\nera4b\n4erand\ner3ar\n4erati.\n2erb\ner4bl\ner3ch\ner4che\n2ere.\ne3real\nere5co\nere3in\ner5el.\ner3emo\ner5ena\ner5ence\n4erene\ner3ent\nere4q\ner5ess\ner3est\neret4\ner1h\ner1i\ne1ria4\n5erick\ne3rien\neri4er\ner3ine\ne1rio\n4erit\ner4iu\neri4v\ne4riva\ner3m4\ner4nis\n4ernit\n5erniz\ner3no\n2ero\ner5ob\ne5roc\nero4r\ner1ou\ner1s\ner3set\nert3er\n4ertl\ner3tw\n4eru\neru4t\n5erwau\ne1s4a\ne4sage.\ne4sages\nes2c\ne2sca\nes5can\ne3scr\nes5cu\ne1s2e\ne2sec\nes5ecr\nes5enc\ne4sert.\ne4serts\ne4serva\n4esh\ne3sha\nesh5en\ne1si\ne2sic\ne2sid\nes5iden\nes5igna\ne2s5im\nes4i4n\nesis4te\nesi4u\ne5skin\nes4mi\ne2sol\nes3olu\ne2son\nes5ona\ne1sp\nes3per\nes5pira\nes4pre\n2ess\nes4si4b\nestan4\nes3tig\nes5tim\n4es2to\ne3ston\n2estr\ne5stro\nestruc5\ne2sur\nes5urr\nes4w\neta4b\neten4d\ne3teo\nethod3\net1ic\ne5tide\netin4\neti4no\ne5tir\ne5titio\net5itiv\n4etn\net5ona\ne3tra\ne3tre\net3ric\net5rif\net3rog\net5ros\net3ua\net5ym\net5z\n4eu\ne5un\ne3up\neu3ro\neus4\neute4\neuti5l\neu5tr\neva2p5\ne2vas\nev5ast\ne5vea\nev3ell\nevel3o\ne5veng\neven4i\nev1er\ne5verb\ne1vi\nev3id\nevi4l\ne4vin\nevi4v\ne5voc\ne5vu\ne1wa\ne4wag\ne5wee\ne3wh\newil5\new3ing\ne3wit\n1exp\n5eyc\n5eye.\neys4\n1fa\nfa3bl\nfab3r\nfa4ce\n4fag\nfain4\nfall5e\n4fa4ma\nfam5is\n5far\nfar5th\nfa3ta\nfa3the\n4fato\nfault5\n4f5b\n4fd\n4fe.\nfeas4\nfeath3\nfe4b\n4feca\n5fect\n2fed\nfe3li\nfe4mo\nfen2d\nfend5e\nfer1\n5ferr\nfev4\n4f1f\nf4fes\nf4fie\nf5fin.\nf2f5is\nf4fly\nf2fy\n4fh\n1fi\nfi3a\n2f3ic.\n4f3ical\nf3ican\n4ficate\nf3icen\nfi3cer\nfic4i\n5ficia\n5ficie\n4fics\nfi3cu\nfi5del\nfight5\nfil5i\nfill5in\n4fily\n2fin\n5fina\nfin2d5\nfi2ne\nf1in3g\nfin4n\nfis4ti\nf4l2\nf5less\nflin4\nflo3re\nf2ly5\n4fm\n4fn\n1fo\n5fon\nfon4de\nfon4t\nfo2r\nfo5rat\nfor5ay\nfore5t\nfor4i\nfort5a\nfos5\n4f5p\nfra4t\nf5rea\nfres5c\nfri2\nfril4\nfrol5\n2f3s\n2ft\nf4to\nf2ty\n3fu\nfu5el\n4fug\nfu4min\nfu5ne\nfu3ri\nfusi4\nfus4s\n4futa\n1fy\n1ga\ngaf4\n5gal.\n3gali\nga3lo\n2gam\nga5met\ng5amo\ngan5is\nga3niz\ngani5za\n4gano\ngar5n4\ngass4\ngath3\n4gativ\n4gaz\ng3b\ngd4\n2ge.\n2ged\ngeez4\ngel4in\nge5lis\nge5liz\n4gely\n1gen\nge4nat\nge5niz\n4geno\n4geny\n1geo\nge3om\ng4ery\n5gesi\ngeth5\n4geto\nge4ty\nge4v\n4g1g2\ng2ge\ng3ger\ngglu5\nggo4\ngh3in\ngh5out\ngh4to\n5gi.\n1gi4a\ngia5r\ng1ic\n5gicia\ng4ico\ngien5\n5gies.\ngil4\ng3imen\n3g4in.\ngin5ge\n5g4ins\n5gio\n3gir\ngir4l\ng3isl\ngi4u\n5giv\n3giz\ngl2\ngla4\nglad5i\n5glas\n1gle\ngli4b\ng3lig\n3glo\nglo3r\ng1m\ng4my\ngn4a\ng4na.\ngnet4t\ng1ni\ng2nin\ng4nio\ng1no\ng4non\n1go\n3go.\ngob5\n5goe\n3g4o4g\ngo3is\ngon2\n4g3o3na\ngondo5\ngo3ni\n5goo\ngo5riz\ngor5ou\n5gos.\ngov1\ng3p\n1gr\n4grada\ng4rai\ngran2\n5graph.\ng5rapher\n5graphic\n4graphy\n4gray\ngre4n\n4gress.\n4grit\ng4ro\ngruf4\ngs2\ng5ste\ngth3\ngu4a\n3guard\n2gue\n5gui5t\n3gun\n3gus\n4gu4t\ng3w\n1gy\n2g5y3n\ngy5ra\nh3ab4l\nhach4\nhae4m\nhae4t\nh5agu\nha3la\nhala3m\nha4m\nhan4ci\nhan4cy\n5hand.\nhan4g\nhang5er\nhang5o\nh5a5niz\nhan4k\nhan4te\nhap3l\nhap5t\nha3ran\nha5ras\nhar2d\nhard3e\nhar4le\nharp5en\nhar5ter\nhas5s\nhaun4\n5haz\nhaz3a\nh1b\n1head\n3hear\nhe4can\nh5ecat\nh4ed\nhe5do5\nhe3l4i\nhel4lis\nhel4ly\nh5elo\nhem4p\nhe2n\nhena4\nhen5at\nheo5r\nhep5\nh4era\nhera3p\nher4ba\nhere5a\nh3ern\nh5erou\nh3ery\nh1es\nhe2s5p\nhe4t\nhet4ed\nheu4\nh1f\nh1h\nhi5an\nhi4co\nhigh5\nh4il2\nhimer4\nh4ina\nhion4e\nhi4p\nhir4l\nhi3ro\nhir4p\nhir4r\nhis3el\nhis4s\nhith5er\nhi2v\n4hk\n4h1l4\nhlan4\nh2lo\nhlo3ri\n4h1m\nhmet4\n2h1n\nh5odiz\nh5ods\nho4g\nhoge4\nhol5ar\n3hol4e\nho4ma\nhome3\nhon4a\nho5ny\n3hood\nhoon4\nhor5at\nho5ris\nhort3e\nho5ru\nhos4e\nho5sen\nhos1p\n1hous\nhouse3\nhov5el\n4h5p\n4hr4\nhree5\nhro5niz\nhro3po\n4h1s2\nh4sh\nh4tar\nht1en\nht5es\nh4ty\nhu4g\nhu4min\nhun5ke\nhun4t\nhus3t4\nhu4t\nh1w\nh4wart\nhy3pe\nhy3ph\nhy2s\n2i1a\ni2al\niam4\niam5ete\ni2an\n4ianc\nian3i\n4ian4t\nia5pe\niass4\ni4ativ\nia4tric\ni4atu\nibe4\nib3era\nib5ert\nib5ia\nib3in\nib5it.\nib5ite\ni1bl\nib3li\ni5bo\ni1br\ni2b5ri\ni5bun\n4icam\n5icap\n4icar\ni4car.\ni4cara\nicas5\ni4cay\niccu4\n4iceo\n4ich\n2ici\ni5cid\nic5ina\ni2cip\nic3ipa\ni4cly\ni2c5oc\n4i1cr\n5icra\ni4cry\nic4te\nictu2\nic4t3ua\nic3ula\nic4um\nic5uo\ni3cur\n2id\ni4dai\nid5anc\nid5d\nide3al\nide4s\ni2di\nid5ian\nidi4ar\ni5die\nid3io\nidi5ou\nid1it\nid5iu\ni3dle\ni4dom\nid3ow\ni4dr\ni2du\nid5uo\n2ie4\nied4e\n5ie5ga\nield3\nien5a4\nien4e\ni5enn\ni3enti\ni1er.\ni3esc\ni1est\ni3et\n4if.\nif5ero\niff5en\nif4fr\n4ific.\ni3fie\ni3fl\n4ift\n2ig\niga5b\nig3era\night3i\n4igi\ni3gib\nig3il\nig3in\nig3it\ni4g4l\ni2go\nig3or\nig5ot\ni5gre\nigu5i\nig1ur\ni3h\n4i5i4\ni3j\n4ik\ni1la\nil3a4b\ni4lade\ni2l5am\nila5ra\ni3leg\nil1er\nilev4\nil5f\nil1i\nil3ia\nil2ib\nil3io\nil4ist\n2ilit\nil2iz\nill5ab\n4iln\nil3oq\nil4ty\nil5ur\nil3v\ni4mag\nim3age\nima5ry\nimenta5r\n4imet\nim1i\nim5ida\nimi5le\ni5mini\n4imit\nim4ni\ni3mon\ni2mu\nim3ula\n2in.\ni4n3au\n4inav\nincel4\nin3cer\n4ind\nin5dling\n2ine\ni3nee\niner4ar\ni5ness\n4inga\n4inge\nin5gen\n4ingi\nin5gling\n4ingo\n4ingu\n2ini\ni5ni.\ni4nia\nin3io\nin1is\ni5nite.\n5initio\nin3ity\n4ink\n4inl\n2inn\n2i1no\ni4no4c\nino4s\ni4not\n2ins\nin3se\ninsur5a\n2int.\n2in4th\nin1u\ni5nus\n4iny\n2io\n4io.\nioge4\nio2gr\ni1ol\nio4m\nion3at\nion4ery\nion3i\nio5ph\nior3i\ni4os\nio5th\ni5oti\nio4to\ni4our\n2ip\nipe4\niphras4\nip3i\nip4ic\nip4re4\nip3ul\ni3qua\niq5uef\niq3uid\niq3ui3t\n4ir\ni1ra\nira4b\ni4rac\nird5e\nire4de\ni4ref\ni4rel4\ni4res\nir5gi\nir1i\niri5de\nir4is\niri3tu\n5i5r2iz\nir4min\niro4g\n5iron.\nir5ul\n2is.\nis5ag\nis3ar\nisas5\n2is1c\nis3ch\n4ise\nis3er\n3isf\nis5han\nis3hon\nish5op\nis3ib\nisi4d\ni5sis\nis5itiv\n4is4k\nislan4\n4isms\ni2so\niso5mer\nis1p\nis2pi\nis4py\n4is1s\nis4sal\nissen4\nis4ses\nis4ta.\nis1te\nis1ti\nist4ly\n4istral\ni2su\nis5us\n4ita.\nita4bi\ni4tag\n4ita5m\ni3tan\ni3tat\n2ite\nit3era\ni5teri\nit4es\n2ith\ni1ti\n4itia\n4i2tic\nit3ica\n5i5tick\nit3ig\nit5ill\ni2tim\n2itio\n4itis\ni4tism\ni2t5o5m\n4iton\ni4tram\nit5ry\n4itt\nit3uat\ni5tud\nit3ul\n4itz.\ni1u\n2iv\niv3ell\niv3en.\ni4v3er.\ni4vers.\niv5il.\niv5io\niv1it\ni5vore\niv3o3ro\ni4v3ot\n4i5w\nix4o\n4iy\n4izar\nizi4\n5izont\n5ja\njac4q\nja4p\n1je\njer5s\n4jestie\n4jesty\njew3\njo4p\n5judg\n3ka.\nk3ab\nk5ag\nkais4\nkal4\nk1b\nk2ed\n1kee\nke4g\nke5li\nk3en4d\nk1er\nkes4\nk3est.\nke4ty\nk3f\nkh4\nk1i\n5ki.\n5k2ic\nk4ill\nkilo5\nk4im\nk4in.\nkin4de\nk5iness\nkin4g\nki4p\nkis4\nk5ish\nkk4\nk1l\n4kley\n4kly\nk1m\nk5nes\n1k2no\nko5r\nkosh4\nk3ou\nkro5n\n4k1s2\nk4sc\nks4l\nk4sy\nk5t\nk1w\nlab3ic\nl4abo\nlaci4\nl4ade\nla3dy\nlag4n\nlam3o\n3land\nlan4dl\nlan5et\nlan4te\nlar4g\nlar3i\nlas4e\nla5tan\n4lateli\n4lativ\n4lav\nla4v4a\n2l1b\nlbin4\n4l1c2\nlce4\nl3ci\n2ld\nl2de\nld4ere\nld4eri\nldi4\nld5is\nl3dr\nl4dri\nle2a\nle4bi\nleft5\n5leg.\n5legg\nle4mat\nlem5atic\n4len.\n3lenc\n5lene.\n1lent\nle3ph\nle4pr\nlera5b\nler4e\n3lerg\n3l4eri\nl4ero\nles2\nle5sco\n5lesq\n3less\n5less.\nl3eva\nlev4er.\nlev4era\nlev4ers\n3ley\n4leye\n2lf\nl5fr\n4l1g4\nl5ga\nlgar3\nl4ges\nlgo3\n2l3h\nli4ag\nli2am\nliar5iz\nli4as\nli4ato\nli5bi\n5licio\nli4cor\n4lics\n4lict.\nl4icu\nl3icy\nl3ida\nlid5er\n3lidi\nlif3er\nl4iff\nli4fl\n5ligate\n3ligh\nli4gra\n3lik\n4l4i4l\nlim4bl\nlim3i\nli4mo\nl4im4p\nl4ina\n1l4ine\nlin3ea\nlin3i\nlink5er\nli5og\n4l4iq\nlis4p\nl1it\nl2it.\n5litica\nl5i5tics\nliv3er\nl1iz\n4lj\nlka3\nl3kal\nlka4t\nl1l\nl4law\nl2le\nl5lea\nl3lec\nl3leg\nl3lel\nl3le4n\nl3le4t\nll2i\nl2lin4\nl5lina\nll4o\nlloqui5\nll5out\nl5low\n2lm\nl5met\nlm3ing\nl4mod\nlmon4\n2l1n2\n3lo.\nlob5al\nlo4ci\n4lof\n3logic\nl5ogo\n3logu\nlom3er\n5long\nlon4i\nl3o3niz\nlood5\n5lope.\nlop3i\nl3opm\nlora4\nlo4rato\nlo5rie\nlor5ou\n5los.\nlos5et\n5losophiz\n5losophy\nlos4t\nlo4ta\nloun5d\n2lout\n4lov\n2lp\nlpa5b\nl3pha\nl5phi\nlp5ing\nl3pit\nl4pl\nl5pr\n4l1r\n2l1s2\nl4sc\nl2se\nl4sie\n4lt\nlt5ag\nltane5\nl1te\nlten4\nltera4\nlth3i\nl5ties.\nltis4\nl1tr\nltu2\nltur3a\nlu5a\nlu3br\nluch4\nlu3ci\nlu3en\nluf4\nlu5id\nlu4ma\n5lumi\nl5umn.\n5lumnia\nlu3o\nluo3r\n4lup\nluss4\nlus3te\n1lut\nl5ven\nl5vet4\n2l1w\n1ly\n4lya\n4lyb\nly5me\nly3no\n2lys4\nl5yse\n1ma\n2mab\nma2ca\nma5chine\nma4cl\nmag5in\n5magn\n2mah\nmaid5\n4mald\nma3lig\nma5lin\nmal4li\nmal4ty\n5mania\nman5is\nman3iz\n4map\nma5rine.\nma5riz\nmar4ly\nmar3v\nma5sce\nmas4e\nmas1t\n5mate\nmath3\nma3tis\n4matiza\n4m1b\nmba4t5\nm5bil\nm4b3ing\nmbi4v\n4m5c\n4me.\n2med\n4med.\n5media\nme3die\nm5e5dy\nme2g\nmel5on\nmel4t\nme2m\nmem1o3\n1men\nmen4a\nmen5ac\nmen4de\n4mene\nmen4i\nmens4\nmensu5\n3ment\nmen4te\nme5on\nm5ersa\n2mes\n3mesti\nme4ta\nmet3al\nme1te\nme5thi\nm4etr\n5metric\nme5trie\nme3try\nme4v\n4m1f\n2mh\n5mi.\nmi3a\nmid4a\nmid4g\nmig4\n3milia\nm5i5lie\nm4ill\nmin4a\n3mind\nm5inee\nm4ingl\nmin5gli\nm5ingly\nmin4t\nm4inu\nmiot4\nm2is\nmis4er.\nmis5l\nmis4ti\nm5istry\n4mith\nm2iz\n4mk\n4m1l\nm1m\nmma5ry\n4m1n\nmn4a\nm4nin\nmn4o\n1mo\n4mocr\n5mocratiz\nmo2d1\nmo4go\nmois2\nmoi5se\n4mok\nmo5lest\nmo3me\nmon5et\nmon5ge\nmoni3a\nmon4ism\nmon4ist\nmo3niz\nmonol4\nmo3ny.\nmo2r\n4mora.\nmos2\nmo5sey\nmo3sp\nmoth3\nm5ouf\n3mous\nmo2v\n4m1p\nmpara5\nmpa5rab\nmpar5i\nm3pet\nmphas4\nm2pi\nmpi4a\nmp5ies\nm4p1in\nm5pir\nmp5is\nmpo3ri\nmpos5ite\nm4pous\nmpov5\nmp4tr\nm2py\n4m3r\n4m1s2\nm4sh\nm5si\n4mt\n1mu\nmula5r4\n5mult\nmulti3\n3mum\nmun2\n4mup\nmu4u\n4mw\n1na\n2n1a2b\nn4abu\n4nac.\nna4ca\nn5act\nnag5er.\nnak4\nna4li\nna5lia\n4nalt\nna5mit\nn2an\nnanci4\nnan4it\nnank4\nnar3c\n4nare\nnar3i\nnar4l\nn5arm\nn4as\nnas4c\nnas5ti\nn2at\nna3tal\nnato5miz\nn2au\nnau3se\n3naut\nnav4e\n4n1b4\nncar5\nn4ces.\nn3cha\nn5cheo\nn5chil\nn3chis\nnc1in\nnc4it\nncour5a\nn1cr\nn1cu\nn4dai\nn5dan\nn1de\nnd5est.\nndi4b\nn5d2if\nn1dit\nn3diz\nn5duc\nndu4r\nnd2we\n2ne.\nn3ear\nne2b\nneb3u\nne2c\n5neck\n2ned\nne4gat\nneg5ativ\n5nege\nne4la\nnel5iz\nne5mi\nne4mo\n1nen\n4nene\n3neo\nne4po\nne2q\nn1er\nnera5b\nn4erar\nn2ere\nn4er5i\nner4r\n1nes\n2nes.\n4nesp\n2nest\n4nesw\n3netic\nne4v\nn5eve\nne4w\nn3f\nn4gab\nn3gel\nnge4n4e\nn5gere\nn3geri\nng5ha\nn3gib\nng1in\nn5git\nn4gla\nngov4\nng5sh\nn1gu\nn4gum\nn2gy\n4n1h4\nnha4\nnhab3\nnhe4\n3n4ia\nni3an\nni4ap\nni3ba\nni4bl\nni4d\nni5di\nni4er\nni2fi\nni5ficat\nn5igr\nnik4\nn1im\nni3miz\nn1in\n5nine.\nnin4g\nni4o\n5nis.\nnis4ta\nn2it\nn4ith\n3nitio\nn3itor\nni3tr\nn1j\n4nk2\nn5kero\nn3ket\nnk3in\nn1kl\n4n1l\nn5m\nnme4\nnmet4\n4n1n2\nnne4\nnni3al\nnni4v\nnob4l\nno3ble\nn5ocl\n4n3o2d\n3noe\n4nog\nnoge4\nnois5i\nno5l4i\n5nologis\n3nomic\nn5o5miz\nno4mo\nno3my\nno4n\nnon4ag\nnon5i\nn5oniz\n4nop\n5nop5o5li\nnor5ab\nno4rary\n4nosc\nnos4e\nnos5t\nno5ta\n1nou\n3noun\nnov3el3\nnowl3\nn1p4\nnpi4\nnpre4c\nn1q\nn1r\nnru4\n2n1s2\nns5ab\nnsati4\nns4c\nn2se\nn4s3es\nnsid1\nnsig4\nn2sl\nns3m\nn4soc\nns4pe\nn5spi\nnsta5bl\nn1t\nnta4b\nnter3s\nnt2i\nn5tib\nnti4er\nnti2f\nn3tine\nn4t3ing\nnti4p\nntrol5li\nnt4s\nntu3me\nnu1a\nnu4d\nnu5en\nnuf4fe\nn3uin\n3nu3it\nn4um\nnu1me\nn5umi\n3nu4n\nn3uo\nnu3tr\nn1v2\nn1w4\nnym4\nnyp4\n4nz\nn3za\n4oa\noad3\no5a5les\noard3\noas4e\noast5e\noat5i\nob3a3b\no5bar\nobe4l\no1bi\no2bin\nob5ing\no3br\nob3ul\no1ce\noch4\no3chet\nocif3\no4cil\no4clam\no4cod\noc3rac\noc5ratiz\nocre3\n5ocrit\noctor5a\noc3ula\no5cure\nod5ded\nod3ic\nodi3o\no2do4\nodor3\nod5uct.\nod5ucts\no4el\no5eng\no3er\noe4ta\no3ev\no2fi\nof5ite\nofit4t\no2g5a5r\nog5ativ\no4gato\no1ge\no5gene\no5geo\no4ger\no3gie\n1o1gis\nog3it\no4gl\no5g2ly\n3ogniz\no4gro\nogu5i\n1ogy\n2ogyn\no1h2\nohab5\noi2\noic3es\noi3der\noiff4\noig4\noi5let\no3ing\noint5er\no5ism\noi5son\noist5en\noi3ter\no5j\n2ok\no3ken\nok5ie\no1la\no4lan\nolass4\nol2d\nold1e\nol3er\no3lesc\no3let\nol4fi\nol2i\no3lia\no3lice\nol5id.\no3li4f\no5lil\nol3ing\no5lio\no5lis.\nol3ish\no5lite\no5litio\no5liv\nolli4e\nol5ogiz\nolo4r\nol5pl\nol2t\nol3ub\nol3ume\nol3un\no5lus\nol2v\no2ly\nom5ah\noma5l\nom5atiz\nom2be\nom4bl\no2me\nom3ena\nom5erse\no4met\nom5etry\no3mia\nom3ic.\nom3ica\no5mid\nom1in\no5mini\n5ommend\nomo4ge\no4mon\nom3pi\nompro5\no2n\non1a\non4ac\no3nan\non1c\n3oncil\n2ond\non5do\no3nen\non5est\non4gu\non1ic\no3nio\non1is\no5niu\non3key\non4odi\non3omy\non3s\nonspi4\nonspir5a\nonsu4\nonten4\non3t4i\nontif5\non5um\nonva5\noo2\nood5e\nood5i\noo4k\noop3i\no3ord\noost5\no2pa\nope5d\nop1er\n3opera\n4operag\n2oph\no5phan\no5pher\nop3ing\no3pit\no5pon\no4posi\no1pr\nop1u\nopy5\no1q\no1ra\no5ra.\no4r3ag\nor5aliz\nor5ange\nore5a\no5real\nor3ei\nore5sh\nor5est.\norew4\nor4gu\n4o5ria\nor3ica\no5ril\nor1in\no1rio\nor3ity\no3riu\nor2mi\norn2e\no5rof\nor3oug\nor5pe\n3orrh\nor4se\nors5en\norst4\nor3thi\nor3thy\nor4ty\no5rum\no1ry\nos3al\nos2c\nos4ce\no3scop\n4oscopi\no5scr\nos4i4e\nos5itiv\nos3ito\nos3ity\nosi4u\nos4l\no2so\nos4pa\nos4po\nos2ta\no5stati\nos5til\nos5tit\no4tan\notele4g\not3er.\not5ers\no4tes\n4oth\noth5esi\noth3i4\not3ic.\not5ica\no3tice\no3tif\no3tis\noto5s\nou2\nou3bl\nouch5i\nou5et\nou4l\nounc5er\noun2d\nou5v\nov4en\nover4ne\nover3s\nov4ert\no3vis\noviti4\no5v4ol\now3der\now3el\now5est\now1i\nown5i\no4wo\noy1a\n1pa\npa4ca\npa4ce\npac4t\np4ad\n5pagan\np3agat\np4ai\npain4\np4al\npan4a\npan3el\npan4ty\npa3ny\npa1p\npa4pu\npara5bl\npar5age\npar5di\n3pare\npar5el\np4a4ri\npar4is\npa2te\npa5ter\n5pathic\npa5thy\npa4tric\npav4\n3pay\n4p1b\npd4\n4pe.\n3pe4a\npear4l\npe2c\n2p2ed\n3pede\n3pedi\npedia4\nped4ic\np4ee\npee4d\npek4\npe4la\npeli4e\npe4nan\np4enc\npen4th\npe5on\np4era.\npera5bl\np4erag\np4eri\nperi5st\nper4mal\nperme5\np4ern\nper3o\nper3ti\npe5ru\nper1v\npe2t\npe5ten\npe5tiz\n4pf\n4pg\n4ph.\nphar5i\nphe3no\nph4er\nph4es.\nph1ic\n5phie\nph5ing\n5phisti\n3phiz\nph2l\n3phob\n3phone\n5phoni\npho4r\n4phs\nph3t\n5phu\n1phy\npi3a\npian4\npi4cie\npi4cy\np4id\np5ida\npi3de\n5pidi\n3piec\npi3en\npi4grap\npi3lo\npi2n\np4in.\npind4\np4ino\n3pi1o\npion4\np3ith\npi5tha\npi2tu\n2p3k2\n1p2l2\n3plan\nplas5t\npli3a\npli5er\n4plig\npli4n\nploi4\nplu4m\nplum4b\n4p1m\n2p3n\npo4c\n5pod.\npo5em\npo3et5\n5po4g\npoin2\n5point\npoly5t\npo4ni\npo4p\n1p4or\npo4ry\n1pos\npos1s\np4ot\npo4ta\n5poun\n4p1p\nppa5ra\np2pe\np4ped\np5pel\np3pen\np3per\np3pet\nppo5site\npr2\npray4e\n5preci\npre5co\npre3em\npref5ac\npre4la\npre3r\np3rese\n3press\npre5ten\npre3v\n5pri4e\nprin4t3\npri4s\npris3o\np3roca\nprof5it\npro3l\npros3e\npro1t\n2p1s2\np2se\nps4h\np4sib\n2p1t\npt5a4b\np2te\np2th\npti3m\nptu4r\np4tw\npub3\npue4\npuf4\npul3c\npu4m\npu2n\npur4r\n5pus\npu2t\n5pute\nput3er\npu3tr\nput4ted\nput4tin\np3w\nqu2\nqua5v\n2que.\n3quer\n3quet\n2rab\nra3bi\nrach4e\nr5acl\nraf5fi\nraf4t\nr2ai\nra4lo\nram3et\nr2ami\nrane5o\nran4ge\nr4ani\nra5no\nrap3er\n3raphy\nrar5c\nrare4\nrar5ef\n4raril\nr2as\nration4\nrau4t\nra5vai\nrav3el\nra5zie\nr1b\nr4bab\nr4bag\nrbi2\nrbi4f\nr2bin\nr5bine\nrb5ing.\nrb4o\nr1c\nr2ce\nrcen4\nr3cha\nrch4er\nr4ci4b\nrc4it\nrcum3\nr4dal\nrd2i\nrdi4a\nrdi4er\nrdin4\nrd3ing\n2re.\nre1al\nre3an\nre5arr\n5reav\nre4aw\nr5ebrat\nrec5oll\nrec5ompe\nre4cre\n2r2ed\nre1de\nre3dis\nred5it\nre4fac\nre2fe\nre5fer.\nre3fi\nre4fy\nreg3is\nre5it\nre1li\nre5lu\nr4en4ta\nren4te\nre1o\nre5pin\nre4posi\nre1pu\nr1er4\nr4eri\nrero4\nre5ru\nr4es.\nre4spi\nress5ib\nres2t\nre5stal\nre3str\nre4ter\nre4ti4z\nre3tri\nreu2\nre5uti\nrev2\nre4val\nrev3el\nr5ev5er.\nre5vers\nre5vert\nre5vil\nrev5olu\nre4wh\nr1f\nrfu4\nr4fy\nrg2\nrg3er\nr3get\nr3gic\nrgi4n\nrg3ing\nr5gis\nr5git\nr1gl\nrgo4n\nr3gu\nrh4\n4rh.\n4rhal\nri3a\nria4b\nri4ag\nr4ib\nrib3a\nric5as\nr4ice\n4rici\n5ricid\nri4cie\nr4ico\nrid5er\nri3enc\nri3ent\nri1er\nri5et\nrig5an\n5rigi\nril3iz\n5riman\nrim5i\n3rimo\nrim4pe\nr2ina\n5rina.\nrin4d\nrin4e\nrin4g\nri1o\n5riph\nriph5e\nri2pl\nrip5lic\nr4iq\nr2is\nr4is.\nris4c\nr3ish\nris4p\nri3ta3b\nr5ited.\nrit5er.\nrit5ers\nrit3ic\nri2tu\nrit5ur\nriv5el\nriv3et\nriv3i\nr3j\nr3ket\nrk4le\nrk4lin\nr1l\nrle4\nr2led\nr4lig\nr4lis\nrl5ish\nr3lo4\nr1m\nrma5c\nr2me\nr3men\nrm5ers\nrm3ing\nr4ming.\nr4mio\nr3mit\nr4my\nr4nar\nr3nel\nr4ner\nr5net\nr3ney\nr5nic\nr1nis4\nr3nit\nr3niv\nrno4\nr4nou\nr3nu\nrob3l\nr2oc\nro3cr\nro4e\nro1fe\nro5fil\nrok2\nro5ker\n5role.\nrom5ete\nrom4i\nrom4p\nron4al\nron4e\nro5n4is\nron4ta\n1room\n5root\nro3pel\nrop3ic\nror3i\nro5ro\nros5per\nros4s\nro4the\nro4ty\nro4va\nrov5el\nrox5\nr1p\nr4pea\nr5pent\nrp5er.\nr3pet\nrp4h4\nrp3ing\nr3po\nr1r4\nrre4c\nrre4f\nr4reo\nrre4st\nrri4o\nrri4v\nrron4\nrros4\nrrys4\n4rs2\nr1sa\nrsa5ti\nrs4c\nr2se\nr3sec\nrse4cr\nrs5er.\nrs3es\nrse5v2\nr1sh\nr5sha\nr1si\nr4si4b\nrson3\nr1sp\nr5sw\nrtach4\nr4tag\nr3teb\nrten4d\nrte5o\nr1ti\nrt5ib\nrti4d\nr4tier\nr3tig\nrtil3i\nrtil4l\nr4tily\nr4tist\nr4tiv\nr3tri\nrtroph4\nrt4sh\nru3a\nru3e4l\nru3en\nru4gl\nru3in\nrum3pl\nru2n\nrunk5\nrun4ty\nr5usc\nruti5n\nrv4e\nrvel4i\nr3ven\nrv5er.\nr5vest\nr3vey\nr3vic\nrvi4v\nr3vo\nr1w\nry4c\n5rynge\nry3t\nsa2\n2s1ab\n5sack\nsac3ri\ns3act\n5sai\nsalar4\nsal4m\nsa5lo\nsal4t\n3sanc\nsan4de\ns1ap\nsa5ta\n5sa3tio\nsat3u\nsau4\nsa5vor\n5saw\n4s5b\nscan4t5\nsca4p\nscav5\ns4ced\n4scei\ns4ces\nsch2\ns4cho\n3s4cie\n5scin4d\nscle5\ns4cli\nscof4\n4scopy\nscour5a\ns1cu\n4s5d\n4se.\nse4a\nseas4\nsea5w\nse2c3o\n3sect\n4s4ed\nse4d4e\ns5edl\nse2g\nseg3r\n5sei\nse1le\n5self\n5selv\n4seme\nse4mol\nsen5at\n4senc\nsen4d\ns5ened\nsen5g\ns5enin\n4sentd\n4sentl\nsep3a3\n4s1er.\ns4erl\nser4o\n4servo\ns1e4s\nse5sh\nses5t\n5se5um\n5sev\nsev3en\nsew4i\n5sex\n4s3f\n2s3g\ns2h\n2sh.\nsh1er\n5shev\nsh1in\nsh3io\n3ship\nshiv5\nsho4\nsh5old\nshon3\nshor4\nshort5\n4shw\nsi1b\ns5icc\n3side.\n5sides\n5sidi\nsi5diz\n4signa\nsil4e\n4sily\n2s1in\ns2ina\n5sine.\ns3ing\n1sio\n5sion\nsion5a\nsi2r\nsir5a\n1sis\n3sitio\n5siu\n1siv\n5siz\nsk2\n4ske\ns3ket\nsk5ine\nsk5ing\ns1l2\ns3lat\ns2le\nslith5\n2s1m\ns3ma\nsmall3\nsman3\nsmel4\ns5men\n5smith\nsmol5d4\ns1n4\n1so\nso4ce\nsoft3\nso4lab\nsol3d2\nso3lic\n5solv\n3som\n3s4on.\nsona4\nson4g\ns4op\n5sophic\ns5ophiz\ns5ophy\nsor5c\nsor5d\n4sov\nso5vi\n2spa\n5spai\nspa4n\nspen4d\n2s5peo\n2sper\ns2phe\n3spher\nspho5\nspil4\nsp5ing\n4spio\ns4ply\ns4pon\nspor4\n4spot\nsqual4l\ns1r\n2ss\ns1sa\nssas3\ns2s5c\ns3sel\ns5seng\ns4ses.\ns5set\ns1si\ns4sie\nssi4er\nss5ily\ns4sl\nss4li\ns4sn\nsspend4\nss2t\nssur5a\nss5w\n2st.\ns2tag\ns2tal\nstam4i\n5stand\ns4ta4p\n5stat.\ns4ted\nstern5i\ns5tero\nste2w\nstew5a\ns3the\nst2i\ns4ti.\ns5tia\ns1tic\n5stick\ns4tie\ns3tif\nst3ing\n5stir\ns1tle\n5stock\nstom3a\n5stone\ns4top\n3store\nst4r\ns4trad\n5stratu\ns4tray\ns4trid\n4stry\n4st3w\ns2ty\n1su\nsu1al\nsu4b3\nsu2g3\nsu5is\nsuit3\ns4ul\nsu2m\nsum3i\nsu2n\nsu2r\n4sv\nsw2\n4swo\ns4y\n4syc\n3syl\nsyn5o\nsy5rin\n1ta\n3ta.\n2tab\nta5bles\n5taboliz\n4taci\nta5do\n4taf4\ntai5lo\nta2l\nta5la\ntal5en\ntal3i\n4talk\ntal4lis\nta5log\nta5mo\ntan4de\ntanta3\nta5per\nta5pl\ntar4a\n4tarc\n4tare\nta3riz\ntas4e\nta5sy\n4tatic\nta4tur\ntaun4\ntav4\n2taw\ntax4is\n2t1b\n4tc\nt4ch\ntch5et\n4t1d\n4te.\ntead4i\n4teat\ntece4\n5tect\n2t1ed\nte5di\n1tee\nteg4\nte5ger\nte5gi\n3tel.\nteli4\n5tels\nte2ma2\ntem3at\n3tenan\n3tenc\n3tend\n4tenes\n1tent\nten4tag\n1teo\nte4p\nte5pe\nter3c\n5ter3d\n1teri\nter5ies\nter3is\nteri5za\n5ternit\nter5v\n4tes.\n4tess\nt3ess.\nteth5e\n3teu\n3tex\n4tey\n2t1f\n4t1g\n2th.\nthan4\nth2e\n4thea\nth3eas\nthe5at\nthe3is\n3thet\nth5ic.\nth5ica\n4thil\n5think\n4thl\nth5ode\n5thodic\n4thoo\nthor5it\ntho5riz\n2ths\n1tia\nti4ab\nti4ato\n2ti2b\n4tick\nt4ico\nt4ic1u\n5tidi\n3tien\ntif2\nti5fy\n2tig\n5tigu\ntill5in\n1tim\n4timp\ntim5ul\n2t1in\nt2ina\n3tine.\n3tini\n1tio\nti5oc\ntion5ee\n5tiq\nti3sa\n3tise\ntis4m\nti5so\ntis4p\n5tistica\nti3tl\nti4u\n1tiv\ntiv4a\n1tiz\nti3za\nti3zen\n2tl\nt5la\ntlan4\n3tle.\n3tled\n3tles.\nt5let.\nt5lo\n4t1m\ntme4\n2t1n2\n1to\nto3b\nto5crat\n4todo\n2tof\nto2gr\nto5ic\nto2ma\ntom4b\nto3my\nton4ali\nto3nat\n4tono\n4tony\nto2ra\nto3rie\ntor5iz\ntos2\n5tour\n4tout\nto3war\n4t1p\n1tra\ntra3b\ntra5ch\ntraci4\ntrac4it\ntrac4te\ntras4\ntra5ven\ntrav5es5\ntre5f\ntre4m\ntrem5i\n5tria\ntri5ces\n5tricia\n4trics\n2trim\ntri4v\ntro5mi\ntron5i\n4trony\ntro5phe\ntro3sp\ntro3v\ntru5i\ntrus4\n4t1s2\nt4sc\ntsh4\nt4sw\n4t3t2\nt4tes\nt5to\nttu4\n1tu\ntu1a\ntu3ar\ntu4bi\ntud2\n4tue\n4tuf4\n5tu3i\n3tum\ntu4nis\n2t3up.\n3ture\n5turi\ntur3is\ntur5o\ntu5ry\n3tus\n4tv\ntw4\n4t1wa\ntwis4\n4two\n1ty\n4tya\n2tyl\ntype3\nty5ph\n4tz\ntz4e\n4uab\nuac4\nua5na\nuan4i\nuar5ant\nuar2d\nuar3i\nuar3t\nu1at\nuav4\nub4e\nu4bel\nu3ber\nu4bero\nu1b4i\nu4b5ing\nu3ble.\nu3ca\nuci4b\nuc4it\nucle3\nu3cr\nu3cu\nu4cy\nud5d\nud3er\nud5est\nudev4\nu1dic\nud3ied\nud3ies\nud5is\nu5dit\nu4don\nud4si\nu4du\nu4ene\nuens4\nuen4te\nuer4il\n3ufa\nu3fl\nugh3en\nug5in\n2ui2\nuil5iz\nui4n\nu1ing\nuir4m\nuita4\nuiv3\nuiv4er.\nu5j\n4uk\nu1la\nula5b\nu5lati\nulch4\n5ulche\nul3der\nul4e\nu1len\nul4gi\nul2i\nu5lia\nul3ing\nul5ish\nul4lar\nul4li4b\nul4lis\n4ul3m\nu1l4o\n4uls\nuls5es\nul1ti\nultra3\n4ultu\nu3lu\nul5ul\nul5v\num5ab\num4bi\num4bly\nu1mi\nu4m3ing\numor5o\num2p\nunat4\nu2ne\nun4er\nu1ni\nun4im\nu2nin\nun5ish\nuni3v\nun3s4\nun4sw\nunt3ab\nun4ter.\nun4tes\nunu4\nun5y\nun5z\nu4ors\nu5os\nu1ou\nu1pe\nuper5s\nu5pia\nup3ing\nu3pl\nup3p\nupport5\nupt5ib\nuptu4\nu1ra\n4ura.\nu4rag\nu4ras\nur4be\nurc4\nur1d\nure5at\nur4fer\nur4fr\nu3rif\nuri4fic\nur1in\nu3rio\nu1rit\nur3iz\nur2l\nurl5ing.\nur4no\nuros4\nur4pe\nur4pi\nurs5er\nur5tes\nur3the\nurti4\nur4tie\nu3ru\n2us\nu5sad\nu5san\nus4ap\nusc2\nus3ci\nuse5a\nu5sia\nu3sic\nus4lin\nus1p\nus5sl\nus5tere\nus1tr\nu2su\nusur4\nuta4b\nu3tat\n4ute.\n4utel\n4uten\nuten4i\n4u1t2i\nuti5liz\nu3tine\nut3ing\nution5a\nu4tis\n5u5tiz\nu4t1l\nut5of\nuto5g\nuto5matic\nu5ton\nu4tou\nuts4\nu3u\nuu4m\nu1v2\nuxu3\nuz4e\n1va\n5va.\n2v1a4b\nvac5il\nvac3u\nvag4\nva4ge\nva5lie\nval5o\nval1u\nva5mo\nva5niz\nva5pi\nvar5ied\n3vat\n4ve.\n4ved\nveg3\nv3el.\nvel3li\nve4lo\nv4ely\nven3om\nv5enue\nv4erd\n5vere.\nv4erel\nv3eren\nver5enc\nv4eres\nver3ie\nvermi4n\n3verse\nver3th\nv4e2s\n4ves.\nves4te\nve4te\nvet3er\nve4ty\nvi5ali\n5vian\n5vide.\n5vided\n4v3iden\n5vides\n5vidi\nv3if\nvi5gn\nvik4\n2vil\n5vilit\nv3i3liz\nv1in\n4vi4na\nv2inc\nvin5d\n4ving\nvio3l\nv3io4r\nvi1ou\nvi4p\nvi5ro\nvis3it\nvi3so\nvi3su\n4viti\nvit3r\n4vity\n3viv\n5vo.\nvoi4\n3vok\nvo4la\nv5ole\n5volt\n3volv\nvom5i\nvor5ab\nvori4\nvo4ry\nvo4ta\n4votee\n4vv4\nv4y\nw5abl\n2wac\nwa5ger\nwag5o\nwait5\nw5al.\nwam4\nwar4t\nwas4t\nwa1te\nwa5ver\nw1b\nwea5rie\nweath3\nwed4n\nweet3\nwee5v\nwel4l\nw1er\nwest3\nw3ev\nwhi4\nwi2\nwil2\nwill5in\nwin4de\nwin4g\nwir4\n3wise\nwith3\nwiz5\nw4k\nwl4es\nwl3in\nw4no\n1wo2\nwom1\nwo5ven\nw5p\nwra4\nwri4\nwrita4\nw3sh\nws4l\nws4pe\nw5s4t\n4wt\nwy4\nx1a\nxac5e\nx4ago\nxam3\nx4ap\nxas5\nx3c2\nx1e\nxe4cuto\nx2ed\nxer4i\nxe5ro\nx1h\nxhi2\nxhil5\nxhu4\nx3i\nxi5a\nxi5c\nxi5di\nx4ime\nxi5miz\nx3o\nx4ob\nx3p\nxpan4d\nxpecto5\nxpe3d\nx1t2\nx3ti\nx1u\nxu3a\nxx4\ny5ac\n3yar4\ny5at\ny1b\ny1c\ny2ce\nyc5er\ny3ch\nych4e\nycom4\nycot4\ny1d\ny5ee\ny1er\ny4erf\nyes4\nye4t\ny5gi\n4y3h\ny1i\ny3la\nylla5bl\ny3lo\ny5lu\nymbol5\nyme4\nympa3\nyn3chr\nyn5d\nyn5g\nyn5ic\n5ynx\ny1o4\nyo5d\ny4o5g\nyom4\nyo5net\ny4ons\ny4os\ny4ped\nyper5\nyp3i\ny3po\ny4poc\nyp2ta\ny5pu\nyra5m\nyr5ia\ny3ro\nyr4r\nys4c\ny3s2e\nys3ica\nys3io\n3ysis\ny4so\nyss4\nys1t\nys3ta\nysur4\ny3thin\nyt3ic\ny1w\nza1\nz5a2b\nzar2\n4zb\n2ze\nze4n\nze4p\nz1er\nze3ro\nzet4\n2z1i\nz4il\nz4is\n5zl\n4zm\n1zo\nzo4m\nzo5ol\nzte4\n4z1z2\nz4zy\n.con5gr\n.de5riva\n.dri5v4\n.eth1y6l1\n.eu4ler\n.ev2\n.ever5si5b\n.ga4s1om1\n.ge4ome\n.ge5ot1\n.he3mo1\n.he3p6a\n.he3roe\n.in5u2t\n.kil2n3i\n.ko6r1te1\n.le6ices\n.me4ga1l\n.met4ala\n.mim5i2c1\n.mi1s4ers\n.ne6o3f\n.noe1th\n.non1e2m\n.poly1s\n.post1am\n.pre1am\n.rav5en1o\n.semi5\n.sem4ic\n.semid6\n.semip4\n.semir4\n.sem6is4\n.semiv4\n.sph6in1\n.spin1o\n.ta5pes1tr\n.te3legr\n.to6pog\n.to2q\n.un3at5t\n.un5err5\n.vi2c3ar\n.we2b1l\n.re1e4c\na5bolic\na2cabl\naf6fish\nam1en3ta5b\nanal6ys\nano5a2c\nans5gr\nans3v\nanti1d\nan3ti1n2\nanti1re\na4pe5able\nar3che5t\nar2range\nas5ymptot\nath3er1o1s\nat6tes.\naugh4tl\nau5li5f\nav3iou\nback2er.\nba6r1onie\nba1thy\nbbi4t\nbe2vie\nbi5d2if\nbil2lab\nbio5m\nbi1orb\nbio1rh\nb1i3tive\nblan2d1\nblin2d1\nblon2d2\nbor1no5\nbo2t1u1l\nbrus4q\nbus6i2er\nbus6i2es\nbuss4ing\nbut2ed.\nbut4ted\ncad5e1m\ncat1a1s2\n4chs.\nchs3hu\nchie5vo\ncig3a3r\ncin2q\ncle4ar\nco6ph1o3n\ncous2ti\ncri3tie\ncroc1o1d\ncro5e2co\nc2tro3me6c\n1cu2r1ance\n2d3alone\ndata1b\ndd5a5b\nd2d5ib\nde4als.\nde5clar1\nde2c5lina\nde3fin3iti\nde2mos\ndes3ic\nde2tic\ndic1aid\ndif5fra\n3di1methy\ndi2ren\ndi2rer\n2d1lead\n2d1li2e\n3do5word\ndren1a5l\ndrif2t1a\nd1ri3pleg5\ndrom3e5d\nd3tab\ndu2al.\ndu1op1o1l\nea4n3ies\ne3chas\nedg1l\ned1uling\neli2t1is\ne1loa\nen1dix\neo3grap\n1e6p3i3neph1\ne2r3i4an.\ne3spac6i\neth1y6l1ene\n5eu2clid1\nfeb1rua\nfermi1o\n3fich\nfit5ted.\nfla1g6el\nflow2er.\n3fluor\ngen2cy.\nge3o1d\nght1we\ng1lead\nget2ic.\n4g1lish\n5glo5bin\n1g2nac\ngnet1ism\ngno5mo\ng2n1or.\ng2noresp\n2g1o4n3i1za\ngraph5er.\ngriev1\ng1utan\nhair1s\nha2p3ar5r\nhatch1\nhex2a3\nhite3sid\nh3i5pel1a4\nhnau3z\nho6r1ic.\nh2t1eou\nhypo1tha\nid4ios\nifac1et\nign4it\nignit1er\ni4jk\nim3ped3a\ninfra1s2\ni5nitely.\nirre6v3oc\ni1tesima\nith5i2l\nitin5er5ar\njanu3a\njapan1e2s\nje1re1m\n1ke6ling\n1ki5netic\n1kovian\nk3sha\nla4c3i5e\nlai6n3ess\nlar5ce1n\nl3chai\nl3chil6d1\nlead6er.\nlea4s1a\n1lec3ta6b\nle3g6en2dre\n1le1noid\nlith1o5g\nll1fl\nl2l3ish\nl5mo3nell\nlo1bot1o1\nlo2ges.\nload4ed.\nload6er.\nl3tea\nlth5i2ly\nlue1p\n1lunk3er\n1lum5bia.\n3lyg1a1mi\nly5styr\nma1la1p\nm2an.\nman3u1sc\nmar1gin1\nmedi2c\nmed3i3cin\nmedio6c1\nme3gran3\nm2en.\n3mi3da5b\n3milita\nmil2l1ag\nmil5li5li\nmi6n3is.\nmi1n2ut1er\nmi1n2ut1est\nm3ma1b\n5maph1ro1\n5moc1ra1t\nmo5e2las\nmol1e5c\nmon4ey1l\nmono3ch\nmo4no1en\nmoro6n5is\nmono1s6\nmoth4et2\nm1ou3sin\nm5shack2\nmu2dro\nmul2ti5u\nn3ar4chs.\nn3ch2es1t\nne3back\n2ne1ski\nn1dieck\nnd3thr\nnfi6n3ites\n4n5i4an.\nnge5nes\nng1ho\nng1spr\nnk3rup\nn5less\n5noc3er1os\nnom1a6l\nnom5e1no\nn1o1mist\nnon1eq\nnon1i4so\n5nop1oly.\nno1vemb\nns5ceiv\nns4moo\nntre1p\nobli2g1\no3chas\nodel3li\nodit1ic\noerst2\noke1st\no3les3ter\noli3gop1o1\no1lo3n4om\no3mecha6\nonom1ic\no3norma\no3no2t1o3n\no3nou\nop1ism.\nor4tho3ni4t\north1ri\nor5tively\no4s3pher\no5test1er\no5tes3tor\noth3e1o1s\nou3ba3do\no6v3i4an.\noxi6d1ic\npal6mat\nparag6ra4\npar4a1le\nparam4\npara3me\npee2v1\nphi2l3ant\nphi5lat1e3l\npi2c1a3d\npli2c1ab\npli5nar\npoin3ca\n1pole.\npoly1e\npo3lyph1ono\n1prema3c\npre1neu\npres2pli\npro2cess\nproc3i3ty.\npro2g1e\n3pseu2d\npseu3d6o3d2\npseu3d6o3f2\npto3mat4\np5trol3\npu5bes5c\nquain2t1e\nqu6a3si3\nquasir6\nquasis6\nquin5tes5s\nqui3v4ar\nr1abolic\n3rab1o1loi\nra3chu\nr3a3dig\nradi1o6g\nr2amen\n3ra4m5e1triz\nra3mou\nra5n2has\nra1or\nr3bin1ge\nre2c3i1pr\nrec5t6ang\nre4t1ribu\nr3ial.\nriv1o1l\n6rk.\nrk1ho\nr1krau\n6rks.\nr5le5qu\nro1bot1\nro5e2las\nro5epide1\nro3mesh\nro1tron\nr3pau5li\nrse1rad1i\nr1thou\nr1treu\nr1veil\nrz1sc\nsales3c\nsales5w\n5sa3par5il\nsca6p1er\nsca2t1ol\ns4chitz\nschro1ding1\n1sci2utt\nscrap4er.\nscy4th1\nsem1a1ph\nse3mes1t\nse1mi6t5ic\nsep3temb\nshoe1st\nsid2ed.\nside5st\nside5sw\nsi5resid\nsky1sc\n3slova1kia\n3s2og1a1my\nso2lute\n3s2pace\n1s2pacin\nspe3cio\nspher1o\nspi2c1il\nspokes5w\nsports3c\nsports3w\ns3qui3to\ns2s1a3chu1\nss3hat\ns2s3i4an.\ns5sign5a3b\n1s2tamp\ns2t1ant5shi\nstar3tli\nsta1ti\nst5b\n1stor1ab\nstrat1a1g\nstrib5ut\nst5scr\nstu1pi4d1\nstyl1is\nsu2per1e6\n1sync\n1syth3i2\nswimm6\n5tab1o1lism\nta3gon.\ntalk1a5\nt1a1min\nt6ap6ath\n5tar2rh\ntch1c\ntch3i1er\nt1cr\nteach4er.\ntele2g\ntele1r6o\n3ter1gei\nter2ic.\nt3ess2es\ntha4l1am\ntho3don\nth1o5gen1i\ntho1k2er\nthy4l1an\nthy3sc\n2t3i4an.\nti2n3o1m\nt1li2er\ntolo2gy\ntot3ic\ntrai3tor1\ntra1vers\ntravers3a3b\ntreach1e\ntr4ial.\n3tro1le1um\ntrof4ic.\ntro3fit\ntro1p2is\n3trop1o5les\n3trop1o5lis\nt1ro1pol3it\ntsch3ie\nttrib1ut1\nturn3ar\nt1wh\nty2p5al\nua3drati\nuad1ratu\nu5do3ny\nuea1m\nu2r1al.\nuri4al.\nus2er.\nv1ativ\nv1oir5du1\nva6guer\nvaude3v\n1verely.\nv1er1eig\nves1tite\nvi1vip3a3r\nvoice1p\nwaste3w6a2\nwave1g4\nw3c\nweek1n\nwide5sp\nwo4k1en\nwrap3aro\nwrit6er.\nx1q\nxquis3\ny5che3d\nym5e5try\ny1stro\nyes5ter1y\nz3ian.\nz3o1phr\nz2z3w"};
  var hyphenationLanguage = '';
  var HYPHEN_INDEX = {};
  var HYPHEN_POS_CACHE = {};
  function hyphenCode(lang) {
    var code = String(lang || hyphenationLanguage || '').toLowerCase().split(/[-_]/)[0];
    if (code === 'it' || code === 'en') return code;
    return '';
  }
  function hyphenPatternIndex(code) {
    if (HYPHEN_INDEX[code]) return HYPHEN_INDEX[code];
    var text = HYPHEN_PATTERNS[code] || '', index = {};
    var lines = text.split('\n');
    for (var l = 0; l < lines.length; l++) {
      var raw = lines[l].trim();
      if (!raw) continue;
      var key = '', weights = [0];
      for (var c = 0; c < raw.length; c++) {
        var ch = raw.charAt(c);
        if (ch >= '0' && ch <= '9') weights[weights.length - 1] = Number(ch);
        else { key += ch; weights.push(0); }
      }
      if (!key) continue;
      var first = key.charAt(0);
      if (!index[first]) index[first] = [];
      index[first].push({ key: key, weights: weights });
    }
    HYPHEN_INDEX[code] = index;
    return index;
  }
  function patternHyphenPoints(clean, code) {
    var lower = clean.toLowerCase();
    var cacheKey = code + ':' + lower;
    if (hasOwn(HYPHEN_POS_CACHE, cacheKey)) return HYPHEN_POS_CACHE[cacheKey];
    var index = hyphenPatternIndex(code), padded = '.' + lower + '.';
    var scores = new Array(padded.length + 2).fill(0);
    var lookup = {};
    for (var c = 0; c < padded.length; c++) lookup[padded.charAt(c)] = 1;
    for (var ch in lookup) {
      if (!hasOwn(lookup, ch)) continue;
      var list = index[ch] || [];
      for (var p = 0; p < list.length; p++) {
        var pattern = list[p], pos = padded.indexOf(pattern.key);
        while (pos >= 0) {
          for (var w = 0; w < pattern.weights.length; w++) {
            var at = pos + w;
            if (pattern.weights[w] > scores[at]) scores[at] = pattern.weights[w];
          }
          pos = padded.indexOf(pattern.key, pos + 1);
        }
      }
    }
    var points = [];
    for (var i = 2; i < padded.length - 1; i++) {
      if (scores[i] % 2 === 1) points.push(i - 1);
    }
    HYPHEN_POS_CACHE[cacheKey] = points;
    return points;
  }
  function heuristicHyphenPoints(lower) {
    var vowels = /[aeiouyàèéìòù]/;
    var points = {};
    for (var i = 2; i < lower.length - 2; i++) {
      var prev = lower.charAt(i - 1), cur = lower.charAt(i), next = lower.charAt(i + 1);
      if (vowels.test(prev) && !vowels.test(cur) && vowels.test(next)) points[i] = 1;
      if (cur === prev && !vowels.test(cur)) points[i] = 1;
    }
    var suffixes = ['tissimo', 'tissima', 'zione', 'zioni', 'mente', 'sione', 'sioni', 'logia', 'grafia', 'ità', 'tion', 'sion', 'ment', 'ness', 'able', 'ible', 'ally', 'ing'];
    for (var s = 0; s < suffixes.length; s++) {
      var pos = lower.lastIndexOf(suffixes[s]);
      if (pos >= 3 && pos < lower.length - 3) points[pos] = 1;
    }
    var keys = [];
    for (var k in points) if (hasOwn(points, k)) keys.push(Number(k));
    if (!keys.length) {
      var middle = Math.floor(lower.length / 2);
      if (middle >= 3 && middle < lower.length - 2) keys.push(middle);
    }
    return keys;
  }
  function hyphenateWord(word, lang) {
    var clean = String(word || '').replace(/\u00AD/g, '');
    if (clean.length < 6 || !/^[A-Za-zÀ-ÖØ-öø-ÿ]+$/.test(clean)) return clean;
    var code = hyphenCode(lang);
    var positions = code ? patternHyphenPoints(clean, code) : [];
    if (!positions.length) positions = heuristicHyphenPoints(clean.toLowerCase());
    positions = positions.filter(function (pos) { return pos > 1 && pos < clean.length - 1; })
      .sort(function (a, b) { return a - b; })
      .filter(function (pos, idx, all) { return idx === 0 || pos !== all[idx - 1]; });
    var out = '', last = 0;
    for (var p = 0; p < positions.length; p++) {
      out += clean.slice(last, positions[p]) + '\u00AD';
      last = positions[p];
    }
    return out + clean.slice(last);
  }
  function hyphenateText(text, lang) {
    return String(text == null ? '' : text).replace(/[A-Za-zÀ-ÖØ-öø-ÿ]{6,}/g, function (word) { return hyphenateWord(word, lang); });
  }
  function removeSoftHyphens(root) {
    if (!root) return;
    var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, null), nodes = [], node;
    while ((node = walker.nextNode())) nodes.push(node);
    for (var i = 0; i < nodes.length; i++) {
      if (nodes[i].nodeValue && nodes[i].nodeValue.indexOf('\u00AD') >= 0) nodes[i].nodeValue = nodes[i].nodeValue.replace(/\u00AD/g, '');
    }
  }
  function softHyphenateElement(root) {
    if (!root) return;
    var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, null), nodes = [], node;
    while ((node = walker.nextNode())) nodes.push(node);
    for (var i = 0; i < nodes.length; i++) {
      var parent = nodes[i].parentElement;
      if (parent && parent.closest && parent.closest('code,pre,script,style')) continue;
      if (!nodes[i].nodeValue || nodes[i].nodeValue.indexOf('\u00AD') >= 0) continue;
      nodes[i].nodeValue = hyphenateText(nodes[i].nodeValue);
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
      box = el('pre');
      var codeBox = el('code');
      if (Array.isArray(node.children) && node.children.length) inlineDom(node.children, codeBox);
      else codeBox.appendChild(document.createTextNode(String(node.text == null ? '' : node.text)));
      box.appendChild(codeBox);
      styleForBlock(box, node);
      if (node.codeLanguage && node.codeLanguage !== 'plain') {
        box.setAttribute('data-code-language', node.codeLanguage);
        highlightCodeBox(codeBox, node.codeLanguage);
      }
      if (node.lineNumbers) box.setAttribute('data-line-numbers', codeLineNumbers(codeBox));
      return box;
    } else if (type === 'hr') {
      return el('hr');
    } else if (type === 'page_break') {
      box = pageBreakNode();
      box.style.breakAfter = 'page';
      return box;
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
      if (node.borderWidth) box.setAttribute('data-table-border-width', cssLength(node.borderWidth, ''));
      if (node.padding) box.setAttribute('data-table-padding', cssLength(node.padding, ''));
      if (node.borders) box.setAttribute('data-table-borders', node.borders.join(','));
      if (node._tableGroup) box.setAttribute('data-sui-table-group', String(node._tableGroup));
      if (node.width) { box.style.width = cssLength(node.width, ''); box.style.maxWidth = 'none'; }
      if (node.height) box.style.height = cssLength(node.height, '');
      if (node.align === 'center') { box.style.marginLeft = 'auto'; box.style.marginRight = 'auto'; }
      else if (node.align === 'right') { box.style.marginLeft = 'auto'; box.style.marginRight = '0'; }
      else if (node.align === 'left') { box.style.marginLeft = '0'; box.style.marginRight = 'auto'; }
      if (node.width || node.height) box.style.tableLayout = 'fixed';
      if (Array.isArray(node.colWidths) && node.colWidths.length) {
        var colgroup = el('colgroup');
        for (var cw = 0; cw < node.colWidths.length; cw++) {
          var colNode = el('col');
          if (node.colWidths[cw]) colNode.style.width = cssLength(node.colWidths[cw], '');
          colgroup.appendChild(colNode);
        }
        box.appendChild(colgroup);
      }
      var rows = Array.isArray(node.rows) ? node.rows : [];
      var tbody = el('tbody');
      for (var ri = 0; ri < rows.length; ri++) {
        var tr = el('tr');
        if (rows[ri].height) tr.style.height = cssLength(rows[ri].height, '');
        var cells = Array.isArray(rows[ri].cells) ? rows[ri].cells : [];
        for (var ci = 0; ci < cells.length; ci++) {
          var cell = el(cells[ci].header ? 'th' : 'td');
          if (cells[ci].align) cell.style.textAlign = cells[ci].align;
          if (cells[ci].background) cell.style.backgroundColor = cssColor(cells[ci].background) || '';
          if (node.padding) cell.style.padding = cssLength(node.padding, '');
          if (node.borderWidth) cell.style.borderWidth = cssLength(node.borderWidth, '');
          if (node.border) cell.style.borderColor = cssColor(node.border) || '';
          if (cells[ci].width) cell.style.width = cssLength(cells[ci].width, '');
          if (Number(cells[ci].colspan) > 1) cell.setAttribute('colspan', String(Number(cells[ci].colspan)));
          if (Number(cells[ci].rowspan) > 1) cell.setAttribute('rowspan', String(Number(cells[ci].rowspan)));
          if (Array.isArray(cells[ci].blocks) && cells[ci].blocks.length) {
            var cellBlocks = el('div', 'sui-rich-cell-blocks');
            for (var cb = 0; cb < cells[ci].blocks.length; cb++) {
              var childBlock = blockDom(cells[ci].blocks[cb], depth + 1);
              if (childBlock) cellBlocks.appendChild(childBlock);
            }
            cell.appendChild(cellBlocks);
          } else inlineDom(Array.isArray(cells[ci].children) ? cells[ci].children : [], cell);
          if (!cell.childNodes.length) cell.appendChild(el('p'));
          tr.appendChild(cell);
        }
        tbody.appendChild(tr);
      }
      box.appendChild(tbody);
      if (node.borders || node.borderWidth || node.border) applyTableBorderSides(box, node.borders || ['top', 'right', 'bottom', 'left', 'insideH', 'insideV'], cssLength(node.borderWidth || '', ''), cssColor(node.border || '') || '');
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
    if (box && node.hyphens === 'auto' && type !== 'code_block') softHyphenateElement(box);
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
    var fallback = inst.editor ? inst.editor.firstElementChild : null;
    if (!sel || !sel.rangeCount) return fallback;
    var node = sel.anchorNode;
    // A reactive re-render replaces the editor: the old selection may still
    // point into the detached view. Never treat that wrapper as a block.
    if (!node || !inst.editor || !inst.editor.contains(node)) return fallback;
    while (node && node !== inst.editor) {
      if (node.nodeType === 1) {
        var tag = node.tagName.toLowerCase();
        if (/^(p|div|h[1-6]|li|blockquote|pre|figure|table|th|td|ul|ol)$/.test(tag)) return node;
      }
      node = node.parentNode;
    }
    return fallback;
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
      var leaf = currentBlock(inst);
      info.leafIndex = -1;
      info.leafOffset = info.blockOffset;
      if (leaf && leaf !== top && top.contains(leaf)) {
        var leaves = top.querySelectorAll('p,h1,h2,h3,h4,pre,li');
        info.leafIndex = Array.prototype.indexOf.call(leaves, leaf);
        info.leafOffset = textOffsetWithin(leaf, range.startContainer, range.startOffset);
      }
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
    if (info.start === info.end && info.blockIndex != null && info.blockIndex >= 0) {
      var blocks = inst.editor.children;
      var block = blocks[info.blockIndex];
      if (block) {
        var host = block;
        var hostOffset = info.blockOffset;
        if (info.leafIndex != null && info.leafIndex >= 0) {
          var hostLeaves = block.querySelectorAll('p,h1,h2,h3,h4,pre,li');
          if (hostLeaves[info.leafIndex]) { host = hostLeaves[info.leafIndex]; hostOffset = info.leafOffset; }
        }
        var blockWalker = document.createTreeWalker(host, NodeFilter.SHOW_TEXT, null);
        var blockPos = 0, blockNode = null, blockOffset = 0, bn;
        while ((bn = blockWalker.nextNode())) {
          var blen = (bn.nodeValue || '').length;
          if (hostOffset <= blockPos + blen) { blockNode = bn; blockOffset = Math.max(0, hostOffset - blockPos); break; }
          blockPos += blen;
        }
        try {
          var blockRange = document.createRange();
          if (blockNode) blockRange.setStart(blockNode, blockOffset);
          else {
            var emptyChild = host.lastElementChild;
            blockRange.selectNodeContents(emptyChild || host);
            blockRange.collapse(true);
          }
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
    if (inst.doc && inst.doc.header !== undefined) next.header = inst.doc.header;
    if (inst.doc && inst.doc.footer !== undefined) next.footer = inst.doc.footer;
    if (inst.doc && inst.doc.page !== undefined) next.page = inst.doc.page;
    if (inst.doc && inst.doc.pageOverrides !== undefined) next.pageOverrides = inst.doc.pageOverrides;
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
    inst.scrollGuardUntil = Date.now() + 300;
    var changed = sync(inst);
    if (changed) inst.pageBreakSources = [];
    if (inst.editor) {
      inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
      inst.scrollTop = inst.editor.scrollTop;
      inst.keepScrollTop = inst.scrollTop;
    }
    remember(inst);
    updateStats(inst);
    refreshCodeLineNumbers(inst);
    updatePageGuides(inst);
    // A collapsed formatting toggle or color choice changes only the native
    // typing state, not the document. Skipping the state setter keeps that
    // typing state (and the caret) alive instead of re-rendering the editor.
    if (changed || kind === 'input') {
      if (!inst.suppressNotify || kind === 'input') notify(inst, kind || 'change');
    }
    inst.suppressNotify = false;
    if (inst.paginated) schedulePaginate(inst); else schedulePageGuides(inst);
  }
  function restore(inst, text) {
    inst.scrollGuardUntil = Date.now() + 300;
    var keepScroll = inst.keepScrollTop || inst.scrollTop || (inst.editor ? inst.editor.scrollTop : 0);
    inst.doc = parseJSON(text, emptyDoc());
    inst.pageBreakSources = [];
    refreshChromeValues(inst);
    hyphenationLanguage = documentLanguage(inst);
    inst.editor.innerHTML = '';
    inst.editor.appendChild(renderDoc(inst.doc));
    inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
    applyPageLineHeight(inst);
    applyEditorScroll(inst, keepScroll);
    refreshPageChrome(inst);
    updateStats(inst);
    updatePageGuides(inst);
    if (inst.paginated) schedulePaginate(inst); else schedulePageGuides(inst);
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
  // quoteAncestor returns the blockquote that contains block. A rendered quote
  // wraps one paragraph per quote line, so the inner paragraph must not be
  // treated as the styled block.
  function quoteAncestor(block, editor) {
    var node = block;
    while (node && node !== editor) {
      if (node.nodeType === 1 && node.tagName.toLowerCase() === 'blockquote') return node;
      node = node.parentNode;
    }
    return null;
  }
  // blockKind returns the block style under the caret: 'blockquote', 'pre',
  // 'h1'..'h4' or the element tag for ordinary blocks.
  function blockKind(inst) {
    var block = currentBlock(inst);
    if (!block) return '';
    if (quoteAncestor(block, inst.editor)) return 'blockquote';
    return block.tagName ? block.tagName.toLowerCase() : '';
  }
  // unwrapQuote moves the quote children before the quote and removes it,
  // keeping the caret in the same paragraph.
  function unwrapQuote(inst, quote) {
    var parent = quote.parentNode;
    if (!parent) return;
    var active = currentBlock(inst);
    var nodes = Array.prototype.slice.call(quote.childNodes);
    var inline = [];
    var first = null;
    function flushInline() {
      if (!inline.length) return;
      var para = el('p');
      parent.insertBefore(para, quote);
      for (var i = 0; i < inline.length; i++) para.appendChild(inline[i]);
      inline = [];
      if (!first) first = para;
    }
    for (var n = 0; n < nodes.length; n++) {
      var node = nodes[n];
      var isBlock = node.nodeType === 1 && /^(p|h[1-6]|pre|ul|ol|table|figure|div)$/.test(node.tagName.toLowerCase());
      if (isBlock) {
        flushInline();
        parent.insertBefore(node, quote);
        if (!first) first = node;
      } else if (node.nodeType === 1 || (node.nodeValue && node.nodeValue.length)) {
        inline.push(node);
      }
    }
    flushInline();
    if (quote.parentNode === parent) parent.removeChild(quote);
    var caretBlock = (active && active !== quote && parent.contains(active)) ? active : first;
    if (caretBlock) {
      var range = document.createRange();
      range.selectNodeContents(caretBlock);
      range.collapse(false);
      var sel = window.getSelection();
      if (sel) { sel.removeAllRanges(); sel.addRange(range); }
    }
    syncBlockControls(inst);
  }
  // blockTextLines turns a block into logical lines, treating <br> and block
  // boundaries as line breaks. It keeps the line breaks when a code block is
  // converted back to normal paragraphs.
  function blockTextLines(node) {
    var text = '';
    function walk(n) {
      if (n.nodeType === 3) { text += String(n.nodeValue || ''); return; }
      if (n.nodeType !== 1) return;
      var tag = n.tagName.toLowerCase();
      if (tag === 'br') { text += '\n'; return; }
      if (tag === 'div' || tag === 'p') {
        if (text && text.charAt(text.length - 1) !== '\n') text += '\n';
        for (var c = n.firstChild; c; c = c.nextSibling) walk(c);
        if (text && text.charAt(text.length - 1) !== '\n') text += '\n';
        return;
      }
      for (var ch = n.firstChild; ch; ch = ch.nextSibling) walk(ch);
    }
    walk(node);
    var normalized = String(text || '').replace(/\r\n?/g, '\n').replace(/\u00a0/g, ' ');
    if (normalized.charAt(normalized.length - 1) === '\n') normalized = normalized.slice(0, -1);
    return normalized.split('\n');
  }
  function codeLineNumbers(root) {
    var lines = blockTextLines(root);
    var out = [];
    for (var i = 0; i < lines.length; i++) out.push(String(i + 1));
    return out.join('\n');
  }
  function refreshCodeLineNumbers(inst) {
    if (!inst || !inst.editor) return;
    var pres = inst.editor.querySelectorAll('pre[data-line-numbers]');
    for (var i = 0; i < pres.length; i++) {
      var code = pres[i].querySelector('code') || pres[i];
      pres[i].setAttribute('data-line-numbers', codeLineNumbers(code));
    }
  }
  var CODE_LANGUAGES = [
    ['plain', 'Plain text'], ['javascript', 'JavaScript'], ['typescript', 'TypeScript'], ['html', 'HTML'],
    ['css', 'CSS'], ['json', 'JSON'], ['xml', 'XML'], ['markdown', 'Markdown'], ['sql', 'SQL'],
    ['python', 'Python'], ['go', 'Go'], ['java', 'Java'], ['csharp', 'C#'], ['bash', 'Bash']
  ];
  var CODE_KEYWORDS = {
    javascript: 'const let var function return if else for while do switch case break continue class new import from export default async await try catch finally throw true false null undefined this typeof instanceof in of delete void yield static get set',
    typescript: 'const let var function return if else for while do switch case break continue class interface type enum extends implements new import from export default async await try catch finally throw true false null undefined this typeof instanceof public private protected readonly as keyof namespace declare',
    go: 'package import func type struct interface var const return if else for range switch case default break continue go defer chan map select fallthrough true false nil',
    python: 'def class import from return if elif else for while in is not and or None True False try except finally with as lambda yield async await pass global nonlocal assert raise del',
    java: 'public private protected class interface extends implements new return if else for while do switch case break continue try catch finally throw throws import package static final void int long double float boolean char String true false null this super',
    csharp: 'public private protected internal class interface struct enum new return if else for foreach while do switch case break continue try catch finally throw using namespace static readonly const void string int long double float bool true false null this base var async await',
    sql: 'select from where insert into update delete join left right inner outer on group by order having limit offset as and or not null is in exists create table alter drop primary key foreign references distinct union all case when then else end',
    json: 'true false null',
    css: '', html: '', xml: '', markdown: '',
    bash: 'if then else elif fi for do done while case esac function echo export local read cd exit return in'
  };
  var CODE_TOKEN_SELECTOR = '.sui-code-keyword,.sui-code-string,.sui-code-comment,.sui-code-number,.sui-code-tag';
  function codeLanguage(value) {
    var lang = String(value || '').toLowerCase();
    for (var i = 0; i < CODE_LANGUAGES.length; i++) if (CODE_LANGUAGES[i][0] === lang) return lang;
    return 'plain';
  }
  var CODE_TOKEN_COLORS = { keyword: '#7c3aed', string: '#0a7a3d', comment: '#6b7280', number: '#b45309', tag: '#1d4ed8', linenum: '#888888' };
  function codeExportText(node) {
    var text = String(node && node.text != null ? node.text : '');
    if (!text && node && Array.isArray(node.children)) {
      for (var i = 0; i < node.children.length; i++) {
        var n = node.children[i];
        if (!n) continue;
        if (n.type === 'text') text += String(n.text || '');
        else if (n.type === 'break') text += '\n';
      }
    }
    return text;
  }
  function codeExportSegments(node) {
    var text = codeExportText(node);
    var lang = codeLanguage(node && node.codeLanguage);
    var lines = text.split('\n');
    var numberWidth = String(lines.length).length;
    var out = [];
    for (var i = 0; i < lines.length; i++) {
      if (node && node.lineNumbers) {
        var num = String(i + 1);
        while (num.length < numberWidth) num = ' ' + num;
        out.push({ text: num + '  ', cls: 'linenum' });
      }
      if (lang === 'plain') out.push({ text: lines[i], cls: '' });
      else out = out.concat(codeTokenize(lines[i], lang));
      if (i < lines.length - 1) out.push({ text: '\n', cls: '' });
    }
    return out;
  }
  function codeSegmentMarks(seg) {
    var marks = [{ type: 'font', value: 'Courier New' }, { type: 'size', value: '9' }];
    var color = CODE_TOKEN_COLORS[seg.cls];
    if (color) marks.push({ type: 'color', value: color });
    if (seg.cls === 'keyword') marks.push({ type: 'bold' });
    return marks;
  }
  function codeTokenize(text, lang) {
    var tokens = [], source = String(text || '');
    if (!source) return tokens;
    var kw = CODE_KEYWORDS[lang] || '';
    var parts = [];
    if (lang === 'html' || lang === 'xml') parts.push('<\\/?[A-Za-z][^>]*>');
    parts.push('\\/\\*[\\s\\S]*?\\*\\/');
    if (lang === 'sql') parts.push('--[^\\n]*');
    else parts.push('\\/\\/[^\\n]*');
    if (lang === 'python' || lang === 'bash') parts.push('#[^\\n]*');
    parts.push('<!--[\\s\\S]*?-->');
    parts.push('"(?:\\\\.|[^"\\\\])*"');
    parts.push("'(?:\\\\.|[^'\\\\])*'");
    var bt = String.fromCharCode(96);
    parts.push(bt + '(?:\\\\.|[^' + bt + '\\\\])*' + bt);
    parts.push('\\b\\d+(?:\\.\\d+)?\\b');
    var words = kw.split(/\s+/).filter(function (w) { return w; }).sort(function (a, b) { return b.length - a.length; });
    if (words.length) parts.push('\\b(?:' + words.join('|') + ')\\b');
    var re;
    try { re = new RegExp('(' + parts.join('|') + ')', lang === 'sql' ? 'gi' : 'g'); } catch (e) { re = /$^/g; }
    var last = 0, match;
    while ((match = re.exec(source))) {
      if (match.index > last) tokens.push({ text: source.slice(last, match.index), cls: '' });
      var token = match[0], cls = '';
      if (/^\/\*/.test(token) || /^\/\//.test(token) || /^#/.test(token) || /^--/.test(token) || /^<!--/.test(token)) cls = 'comment';
      else if (/^["']/.test(token) || token.charCodeAt(0) === 96) cls = 'string';
      else if (/^\d/.test(token)) cls = 'number';
      else if ((lang === 'html' || lang === 'xml') && /^</.test(token)) cls = 'tag';
      else cls = 'keyword';
      tokens.push({ text: token, cls: cls });
      last = match.index + token.length;
      if (match[0] === '') re.lastIndex++;
    }
    if (last < source.length) tokens.push({ text: source.slice(last), cls: '' });
    return tokens;
  }
  function clearCodeHighlight(root) {
    if (!root) return;
    var nodes = root.querySelectorAll(CODE_TOKEN_SELECTOR);
    for (var i = nodes.length - 1; i >= 0; i--) {
      var node = nodes[i];
      if (!node.parentNode) continue;
      while (node.firstChild) node.parentNode.insertBefore(node.firstChild, node);
      node.parentNode.removeChild(node);
    }
  }
  function highlightCodeBox(root, lang) {
    if (!root || !lang || lang === 'plain') return;
    var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, null), nodes = [], node;
    while ((node = walker.nextNode())) nodes.push(node);
    for (var i = 0; i < nodes.length; i++) {
      var parent = nodes[i].parentElement;
      if (parent && parent.closest && parent.closest(CODE_TOKEN_SELECTOR)) continue;
      var tokens = codeTokenize(nodes[i].nodeValue, lang);
      if (tokens.length <= 1 && (!tokens[0] || !tokens[0].cls)) continue;
      var frag = document.createDocumentFragment();
      for (var t = 0; t < tokens.length; t++) {
        if (!tokens[t].text) continue;
        if (!tokens[t].cls) frag.appendChild(document.createTextNode(tokens[t].text));
        else {
          var span = el('span', 'sui-code-' + tokens[t].cls);
          span.textContent = tokens[t].text;
          frag.appendChild(span);
        }
      }
      if (nodes[i].parentNode) nodes[i].parentNode.replaceChild(frag, nodes[i]);
    }
  }
  function blockLine(name, text) {
    var node = el(name);
    if (text) node.appendChild(document.createTextNode(text));
    else node.appendChild(el('br'));
    return node;
  }
  function preFromLines(lines) {
    var pre = el('pre');
    var code = el('code');
    code.textContent = lines.join('\n');
    pre.appendChild(code);
    return pre;
  }
  function preOffsetOf(pre, container, offset) {
    var total = 0, found = null;
    function walk(node) {
      if (found !== null) return;
      if (node.nodeType === 3) {
        if (node === container) { total += Math.min(offset, String(node.nodeValue || '').length); found = total; return; }
        total += String(node.nodeValue || '').length;
        return;
      }
      if (node.nodeType !== 1) return;
      var tag = node.tagName.toLowerCase();
      if (tag === 'br') {
        if (node === container) { found = total; return; }
        total += 1;
        return;
      }
      if (node === container) {
        for (var c = 0; c < node.childNodes.length && c < offset; c++) walk(node.childNodes[c]);
        if (found === null) found = total;
        return;
      }
      for (var i = 0; i < node.childNodes.length; i++) { walk(node.childNodes[i]); if (found !== null) return; }
    }
    walk(pre);
    return found === null ? total : found;
  }
  function preCaretOffset(pre, range) {
    return range ? preOffsetOf(pre, range.startContainer, range.startOffset) : 0;
  }
  function textPositionAt(root, target) {
    var remaining = target, result = null;
    function walk(node) {
      if (result || !node) return;
      if (node.nodeType === 3) {
        var len = String(node.nodeValue || '').length;
        if (remaining <= len) { result = { node: node, offset: remaining }; return; }
        remaining -= len;
        return;
      }
      if (node.nodeType !== 1) return;
      if (node.tagName.toLowerCase() === 'br') {
        if (remaining === 0) {
          var parent = node.parentNode, index = 0;
          while (index < parent.childNodes.length && parent.childNodes[index] !== node) index++;
          result = { node: parent, offset: index };
          return;
        }
        remaining -= 1;
        return;
      }
      for (var i = 0; i < node.childNodes.length; i++) { walk(node.childNodes[i]); if (result) return; }
    }
    walk(root);
    return result;
  }
  function codeLineInfo(pre) {
    var lines = blockTextLines(pre);
    var starts = [], pos = 0;
    for (var i = 0; i < lines.length; i++) { starts.push(pos); pos += lines[i].length + 1; }
    var sel = window.getSelection();
    var range = sel && sel.rangeCount ? sel.getRangeAt(0) : null;
    var start = 0, end = 0;
    if (range && pre.contains(range.commonAncestorContainer)) {
      start = preOffsetOf(pre, range.startContainer, range.startOffset);
      end = preOffsetOf(pre, range.endContainer, range.endOffset);
      if (range.collapsed) end = start;
    } else {
      start = end = preCaretOffset(pre, range);
    }
    function indexAt(offset) {
      for (var i = 0; i < lines.length; i++) {
        if (offset <= starts[i] + lines[i].length) return i;
      }
      return Math.max(0, lines.length - 1);
    }
    return { lines: lines, starts: starts, startLine: indexAt(start), endLine: indexAt(end), startOffset: start, endOffset: end };
  }
  // indentCodeLines indents the selected code lines with two spaces. It edits
  // text nodes in place so bold/color spans inside the code block survive.
  function indentCodeLines(inst, pre, delta) {
    var info = codeLineInfo(pre);
    var first = info.startLine, last = info.endLine;
    if (last < first) { var swap = first; first = last; last = swap; }
    var caretLine = info.startLine;
    var step = Math.max(1, Math.min(8, parseInt(pre.style.tabSize, 10) || 2));
    var caretColumn = Math.max(0, info.startOffset - info.starts[caretLine]);
    for (var line = last; line >= first; line--) {
      var lineStart = info.starts[line];
      if (delta > 0) {
        var pos = textPositionAt(pre, lineStart);
        if (!pos) continue;
        var range = document.createRange();
        range.setStart(pos.node, pos.offset);
        range.collapse(true);
        range.insertNode(document.createTextNode(new Array(step + 1).join(' ')));
      } else {
        var leading = /^ */.exec(info.lines[line])[0].length;
        var remove = Math.min(step, leading);
        if (!remove) continue;
        var from = textPositionAt(pre, lineStart);
        var to = textPositionAt(pre, lineStart + remove);
        if (!from || !to) continue;
        var cut = document.createRange();
        cut.setStart(from.node, from.offset);
        cut.setEnd(to.node, to.offset);
        cut.deleteContents();
      }
    }
    var newInfo = codeLineInfo(pre);
    var targetLine = Math.min(caretLine, newInfo.lines.length - 1);
    var targetColumn = Math.max(0, caretColumn + (delta > 0 ? step : -Math.min(step, caretColumn)));
    var target = textPositionAt(pre, newInfo.starts[targetLine] + targetColumn);
    if (target) {
      var caret = document.createRange();
      caret.setStart(target.node, target.offset);
      caret.collapse(true);
      var sel = window.getSelection();
      if (sel) { sel.removeAllRanges(); sel.addRange(caret); }
    }
    if (inst.editor) inst.editor.focus();
    saveSelection(inst);
  }
  function toggleCodeBlockLine(inst, pre, tag) {
    var sel = window.getSelection();
    var range = sel && sel.rangeCount ? sel.getRangeAt(0) : null;
    var lines = blockTextLines(pre);
    var offset = preCaretOffset(pre, range);
    var index = 0, start = 0;
    for (var l = 0; l < lines.length; l++) {
      var end = start + lines[l].length;
      if (offset <= end) { index = l; break; }
      start = end + 1;
      if (l === lines.length - 1) index = l;
    }
    var before = lines.slice(0, index);
    var after = lines.slice(index + 1);
    var current = lines[index] == null ? '' : lines[index];
    var offsetInLine = Math.max(0, Math.min(current.length, offset - start));
    var parent = pre.parentNode;
    var style = pre.getAttribute('style') || '';
    function hasContent(list) {
      for (var i = 0; i < list.length; i++) if (String(list[i]).length) return true;
      return false;
    }
    if (before.length && hasContent(before)) {
      var beforePre = preFromLines(before);
      if (style) beforePre.setAttribute('style', style);
      parent.insertBefore(beforePre, pre);
    }
    var caretHost = blockLine(tag === 'p' ? 'p' : tag, current);
    parent.insertBefore(caretHost, pre);
    if (after.length && hasContent(after)) {
      var afterPre = preFromLines(after);
      if (style) afterPre.setAttribute('style', style);
      parent.insertBefore(afterPre, pre);
    }
    parent.removeChild(pre);
    var caret = document.createRange();
    var firstText = null;
    for (var t = 0; t < caretHost.childNodes.length; t++) if (caretHost.childNodes[t].nodeType === 3) { firstText = caretHost.childNodes[t]; break; }
    if (firstText) caret.setStart(firstText, offsetInLine);
    else { caret.selectNodeContents(caretHost); }
    caret.collapse(true);
    if (sel) { sel.removeAllRanges(); sel.addRange(caret); }
    if (inst.editor) inst.editor.focus();
  }
  function applyBlock(inst, tag) {
    var block = currentBlock(inst);
    if (!block) return;
    var quote = quoteAncestor(block, inst.editor);
    if (quote) {
      if (/^(blockquote)$/.test(tag)) return;
      unwrapQuote(inst, quote);
      if (tag === 'p') return;
      block = currentBlock(inst);
      if (!block) return;
    }
    var blockTag = block.tagName ? block.tagName.toLowerCase() : '';
    if (tag === 'pre' && blockTag === 'pre') return;
    // Turning a code block off releases only the line with the caret; the
    // other lines stay inside code blocks instead of all becoming paragraphs.
    if (blockTag === 'pre' && tag !== 'pre') {
      toggleCodeBlockLine(inst, block, tag);
      syncBlockControls(inst);
      return;
    }
    if (tag === 'pre') {
      var codeLines = blockTextLines(block);
      var pre = el('pre');
      var code = el('code');
      code.textContent = codeLines.join('\n');
      pre.appendChild(code);
      for (var pa = 0; pa < block.attributes.length; pa++) if (block.attributes[pa].name === 'style') pre.setAttribute('style', block.getAttribute('style'));
      block.parentNode.replaceChild(pre, block);
      var preInputSel = window.getSelection();
      var preInputRange = document.createRange();
      preInputRange.selectNodeContents(code);
      preInputRange.collapse(false);
      if (preInputSel) { preInputSel.removeAllRanges(); preInputSel.addRange(preInputRange); }
      syncBlockControls(inst);
      return;
    }
    var next = el(tag);
    for (var i = 0; i < block.attributes.length; i++) if (block.attributes[i].name === 'style') next.setAttribute('style', block.getAttribute('style'));
    while (block.firstChild) next.appendChild(block.firstChild);
    var caretHost = next;
    if (tag === 'blockquote') {
      var hasBlockChild = false;
      for (var qc = next.firstChild; qc; qc = qc.nextSibling) {
        if (qc.nodeType === 1 && /^(p|h[1-6]|pre|ul|ol|table|figure|div)$/.test(qc.tagName.toLowerCase())) { hasBlockChild = true; break; }
      }
      if (!hasBlockChild) {
        var para = el('p');
        while (next.firstChild) para.appendChild(next.firstChild);
        next.appendChild(para);
      }
      caretHost = next.lastElementChild || next;
    }
    block.parentNode.replaceChild(next, block);
    var sel = window.getSelection();
    if (sel) { var r = document.createRange(); r.selectNodeContents(caretHost); r.collapse(false); sel.removeAllRanges(); sel.addRange(r); }
    syncBlockControls(inst);
  }
  // blockStyleValue maps a block kind to the value of the "Block style"
  // select. Normal paragraphs use the empty value.
  function blockStyleValue(kind) {
    if (/^h[1-6]$/.test(kind)) return kind;
    if (kind === 'blockquote') return 'blockquote';
    if (kind === 'pre') return 'pre';
    return '';
  }
  // toggleBlock applies a block style, or returns to a paragraph when the
  // caret is already inside it. This is what lets the Quote/Code buttons work
  // as toggles instead of trapping the caret in the format.
  function toggleBlock(inst, tag) {
    applyBlock(inst, blockKind(inst) === tag ? 'p' : tag);
  }
  function rgbHex(value) {
    var text = String(value || '').trim();
    var match = /^rgba?\(([^)]+)\)$/i.exec(text);
    if (match) {
      var parts = match[1].split(',');
      if (parts.length >= 4 && Number(parts[3]) === 0) return '';
      var r = Math.max(0, Math.min(255, Math.round(Number(parts[0]) || 0)));
      var g = Math.max(0, Math.min(255, Math.round(Number(parts[1]) || 0)));
      var b = Math.max(0, Math.min(255, Math.round(Number(parts[2]) || 0)));
      return '#' + [r, g, b].map(function (n) { return ('0' + n.toString(16)).slice(-2); }).join('');
    }
    var hex = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(text);
    if (!hex) return '';
    var body = hex[1];
    if (body.length === 3) body = body.charAt(0) + body.charAt(0) + body.charAt(1) + body.charAt(1) + body.charAt(2) + body.charAt(2);
    return '#' + body.toLowerCase();
  }
  // inlineAncestorValue reads only an explicit inline mark (not the paper
  // color or a table cell background), so the color swatches can show the
  // document default as black text and white highlight.
  function inlineAncestorValue(inst, prop, attrName) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount || !inst.editor || !inst.editor.contains(sel.anchorNode)) return '';
    var node = sel.anchorNode;
    if (node && node.nodeType !== 1) node = node.parentElement;
    while (node && node !== inst.editor) {
      if (node.nodeType === 1) {
        if (node.tagName && node.tagName.toLowerCase() === 'mark' && prop === 'backgroundColor') return '#ffff00';
        var inline = node.style && node.style[prop];
        if (inline) return inline;
        if (attrName && node.getAttribute) {
          var attrValue = node.getAttribute(attrName);
          if (attrValue) return attrValue;
        }
      }
      node = node.parentElement;
    }
    return '';
  }
  // syncBlockControls reflects the caret block in the toolbar: the style
  // select shows Quote/Code/Heading, and the matching buttons get their
  // pressed state. Without this the select stayed on Normal and selecting
  // Normal again did not fire a change event, leaving no way out.
  function syncBlockControls(inst) {
    if (!inst || !inst.root) return;
    var value = blockStyleValue(blockKind(inst));
    var selects = inst.root.querySelectorAll('select[data-rich-cmd="paragraph"]');
    for (var i = 0; i < selects.length; i++) {
      if (selects[i].value !== value) selects[i].value = value;
      selects[i].setAttribute('data-rich-block', value || 'p');
    }
    var sel = window.getSelection();
    var inside = !!(sel && sel.rangeCount && inst.editor && inst.editor.contains(sel.anchorNode));
    var inlineCmds = { bold: 'bold', italic: 'italic', underline: 'underline', strike: 'strikeThrough', superscript: 'superscript', subscript: 'subscript' };
    var inCode = inside ? !!closestIn(inst, 'code') : false;
    var inLink = inside ? !!closestIn(inst, 'a') : false;
    var buttons = inst.root.querySelectorAll('.sui-rich-btn[data-rich-cmd]');
    for (var b = 0; b < buttons.length; b++) {
      var cmd = buttons[b].getAttribute('data-rich-cmd');
      var on = false, disabled = false;
      if (cmd === 'blockquote') on = value === 'blockquote';
      else if (cmd === 'codeBlock') on = value === 'pre';
      else if (/^h[1-4]$/.test(cmd)) on = cmd === value;
      else if (inlineCmds[cmd]) {
        if (inside) {
          var forcedTyping = !!(inst.typingMarks && inst.typingMarks[cmd] !== undefined && sel && sel.isCollapsed);
          if (forcedTyping) on = !!inst.typingMarks[cmd];
          else { try { on = document.queryCommandState(inlineCmds[cmd]); } catch (e) { on = false; } }
        }
      }
      else if (cmd === 'code') {
        if (inside && sel && sel.isCollapsed && inst.typingMarks && inst.typingMarks.code !== undefined) on = !!inst.typingMarks.code;
        else on = inCode;
      }
      else if (cmd === 'link') on = inLink;
      else if (cmd === 'copyFormat') on = !!inst.painter;
      else if (cmd === 'unlink') disabled = !inLink;
      else if (cmd === 'imageSettings') disabled = !inst.editor.querySelector('img');
      else continue;
      buttons[b].classList.toggle('on', on);
      buttons[b].setAttribute('aria-pressed', on ? 'true' : 'false');
      if (cmd === 'unlink' || cmd === 'imageSettings') {
        buttons[b].disabled = disabled;
        buttons[b].setAttribute('aria-disabled', disabled ? 'true' : 'false');
      }
    }
    var colorInputs = inst.root.querySelectorAll('.sui-rich-color[data-rich-cmd]');
    var textColor = inside ? rgbHex(inlineAncestorValue(inst, 'color', 'color')) : '';
    var highlight = inside ? rgbHex(inlineAncestorValue(inst, 'backgroundColor', 'bgcolor')) : '';
    var hasTextSelection = !!(inside && sel && sel.rangeCount && !sel.isCollapsed);
    if (!hasTextSelection) {
      // A color chosen with a collapsed caret is a typing color: keep showing
      // it until the user picks another one, even while a color input has focus.
      if (inst.typingColor) textColor = inst.typingColor;
      if (inst.typingHighlight) highlight = inst.typingHighlight;
    }
    for (var ci = 0; ci < colorInputs.length; ci++) {
      var colorCmd = colorInputs[ci].getAttribute('data-rich-cmd');
      var nextColor = colorCmd === 'textColor' ? (textColor || '#000000') : (highlight || '#ffffff');
      if (colorInputs[ci].value !== nextColor) colorInputs[ci].value = nextColor;
      var colorControl = colorInputs[ci].closest('.sui-rich-color-control');
      if (colorControl) colorControl.style.setProperty('--sui-rich-color-current', nextColor);
    }
  }
  function buildTable(rows, cols) {
    var table = el('table');
    table.style.width = '100%';
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
  function tableColumnWidths(table) {
    var widths = [];
    var firstRow = null;
    for (var r = 0; r < table.rows.length; r++) {
      if (table.rows[r].cells.length) { firstRow = table.rows[r]; break; }
    }
    if (!firstRow) return widths;
    var colgroup = table.querySelector('colgroup');
    var cols = colgroup ? colgroup.querySelectorAll('col') : [];
    var tableWidth = table.getBoundingClientRect().width || 0;
    for (var c = 0; c < cols.length; c++) {
      var rawWidth = String(cols[c].style.width || '');
      widths.push(/%$/.test(rawWidth) ? tableWidth * (parseFloat(rawWidth) || 0) / 100 : (parseFloat(rawWidth) || 0));
    }
    var measured = [];
    for (var i = 0; i < firstRow.cells.length; i++) {
      var cell = firstRow.cells[i];
      var span = Math.max(1, Number(cell.colSpan) || 1);
      var w = cell.getBoundingClientRect().width / span;
      for (var s = 0; s < span; s++) measured.push(w);
    }
    for (var m = 0; m < measured.length; m++) if (!widths[m]) widths[m] = measured[m];
    for (var f = 0; f < widths.length; f++) if (!widths[f]) widths[f] = Math.max(40, table.getBoundingClientRect().width / widths.length);
    return widths;
  }
  function applyTableColumnWidths(table, widths) {
    if (!widths || !widths.length) return;
    var colgroup = table.querySelector('colgroup');
    if (!colgroup) {
      colgroup = el('colgroup');
      table.insertBefore(colgroup, table.firstChild);
    }
    colgroup.innerHTML = '';
    var total = 0;
    for (var i = 0; i < widths.length; i++) {
      var width = Math.max(32, Math.round(widths[i]));
      var col = el('col');
      col.style.width = width + 'px';
      colgroup.appendChild(col);
      total += width;
    }
    table.style.tableLayout = 'fixed';
    table.style.width = total + 'px';
    table.style.maxWidth = 'none';
  }
  function tableGridMap(table) {
    var map = [];
    for (var r = 0; r < table.rows.length; r++) {
      map[r] = map[r] || [];
      var col = 0;
      for (var c = 0; c < table.rows[r].cells.length; c++) {
        while (map[r][col]) col++;
        var cell = table.rows[r].cells[c];
        var rowspan = Math.max(1, Number(cell.rowSpan) || 1);
        var colspan = Math.max(1, Number(cell.colSpan) || 1);
        for (var dr = 0; dr < rowspan; dr++) {
          map[r + dr] = map[r + dr] || [];
          for (var dc = 0; dc < colspan; dc++) map[r + dr][col + dc] = cell;
        }
        col += colspan;
      }
    }
    return map;
  }
  function tableResizeHit(inst, e, pad) {
    var cell = e.target && e.target.closest ? e.target.closest('td,th') : null;
    if (!cell) return null;
    var edgePad = pad || 7;
    var table = cell.closest('table');
    if (!table) return null;
    var tr = table.getBoundingClientRect();
    if (e.clientX >= tr.right - 12 && e.clientY >= tr.bottom - 12) {
      return { kind: 'table', table: table, startX: e.clientX, startY: e.clientY, width: tr.width, height: tr.height };
    }
    var cr = cell.getBoundingClientRect();
    if (e.clientX >= cr.right - edgePad) {
      var map = tableGridMap(table);
      var rowMap = map[cell.parentNode.rowIndex] || [];
      var last = -1;
      for (var c = 0; c < rowMap.length; c++) if (rowMap[c] === cell) last = c;
      if (last < 0) last = cell.cellIndex;
      return { kind: 'col', table: table, index: last, startX: e.clientX, width: cr.width };
    }
    if (e.clientY >= cr.bottom - 6) {
      return { kind: 'row', table: table, row: cell.parentNode, startY: e.clientY, height: cr.height };
    }
    return null;
  }
  function applyTableResize(hit, widths, dx, dy) {
    if (hit.kind === 'table') {
      hit.table.style.width = Math.max(120, hit.width + dx) + 'px';
      hit.table.style.maxWidth = 'none';
      hit.table.style.height = Math.max(60, hit.height + dy) + 'px';
    } else if (hit.kind === 'col') {
      var nextWidths = widths.slice();
      nextWidths[hit.index] = Math.max(40, widths[hit.index] + dx);
      applyTableColumnWidths(hit.table, nextWidths);
    } else if (hit.kind === 'row') {
      hit.row.style.height = Math.max(24, hit.height + dy) + 'px';
    }
  }
  function startTableResize(inst, e, hit) {
    e.preventDefault();
    e.stopPropagation();
    var widths = hit.kind === 'col' ? tableColumnWidths(hit.table) : null;
    function move(ev) { applyTableResize(hit, widths, ev.clientX - hit.startX, ev.clientY - hit.startY); }
    function up() {
      document.removeEventListener('mousemove', move);
      document.removeEventListener('mouseup', up);
      inst.editor.style.cursor = '';
      commit(inst, 'change');
    }
    document.addEventListener('mousemove', move);
    document.addEventListener('mouseup', up);
  }
  function startTouchTableResize(inst, hit, touch) {
    var widths = hit.kind === 'col' ? tableColumnWidths(hit.table) : null;
    function move(ev) {
      if (ev.touches && ev.touches.length) {
        ev.preventDefault();
        var t = ev.touches[0];
        applyTableResize(hit, widths, t.clientX - hit.startX, t.clientY - hit.startY);
      }
    }
    function up() {
      document.removeEventListener('touchmove', move);
      document.removeEventListener('touchend', up);
      document.removeEventListener('touchcancel', up);
      commit(inst, 'change');
    }
    document.addEventListener('touchmove', move, { passive: false });
    document.addEventListener('touchend', up);
    document.addEventListener('touchcancel', up);
  }
  function bindTableResize(inst) {
    inst.editor.addEventListener('mousemove', function (e) {
      var hit = tableResizeHit(inst, e);
      inst.editor.style.cursor = hit ? (hit.kind === 'row' ? 'row-resize' : hit.kind === 'table' ? 'nwse-resize' : 'col-resize') : '';
    });
    inst.editor.addEventListener('mouseleave', function () { inst.editor.style.cursor = ''; });
    inst.editor.addEventListener('touchstart', function (e) {
      if (!e.touches || !e.touches.length) return;
      var t = e.touches[0];
      var hit = tableResizeHit(inst, { target: e.target, clientX: t.clientX, clientY: t.clientY }, 14);
      if (!hit) return;
      e.preventDefault();
      e.stopPropagation();
      startTouchTableResize(inst, hit, t);
    }, { passive: false });
    inst.editor.addEventListener('mousedown', function (e) {
      var del = e.target && e.target.closest ? e.target.closest('[data-rich-pb-delete]') : null;
      if (del) {
        e.preventDefault(); e.stopPropagation();
        var pb = del.closest('.sui-rich-pagebreak');
        if (pb && pb.parentNode) pb.parentNode.removeChild(pb);
        sync(inst);
        if (mergeAdjacentTableFragments(inst)) {
          inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
          remember(inst);
          restore(inst, JSON.stringify(inst.doc));
          notify(inst, 'change');
        } else {
          commit(inst, 'change');
        }
        schedulePageGuides(inst);
        return;
      }
      var breakNode = e.target && e.target.closest ? e.target.closest('.sui-rich-pagebreak') : null;
      if (breakNode) {
        e.preventDefault(); e.stopPropagation();
        startPageBreakDrag(inst, breakNode, e);
        return;
      }
    });
    inst.editor.addEventListener('click', function (e) {
      var img = e.target && e.target.tagName === 'IMG' ? e.target : null;
      if (img) { selectImage(inst, img); return; }
      if (inst.lastImage) {
        inst.lastImage.classList.remove('sui-rich-image-selected');
        inst.lastImage = null;
        updateStats(inst);
      }
    });
    inst.editor.addEventListener('dblclick', function (e) {
      var img = e.target && e.target.tagName === 'IMG' ? e.target : null;
      if (img) { e.preventDefault(); openImageSettings(inst, img); }
    });
    inst.editor.addEventListener('contextmenu', function (e) {
      var img = e.target && e.target.tagName === 'IMG' ? e.target : null;
      if (!img) return;
      e.preventDefault();
      selectImage(inst, img);
      var images = inst.editor.querySelectorAll('img');
      inst.imageIndex = Array.prototype.indexOf.call(images, img);
      openImageMenu(inst, img, e.clientX, e.clientY);
    });
  }
  function cellTextHit(cell, e) {
    if (!cell || !e) return false;
    if (!String(cell.textContent || '').trim()) return false;
    var range = document.createRange();
    range.selectNodeContents(cell);
    var rects = range.getClientRects();
    for (var i = 0; i < rects.length; i++) {
      var r = rects[i];
      if (e.clientX >= r.left - 2 && e.clientX <= r.right + 2 && e.clientY >= r.top - 2 && e.clientY <= r.bottom + 2) return true;
    }
    return false;
  }
  function bindCellSelection(inst) {
    inst.editor.addEventListener('mousedown', function (e) {
      var resizeHit = tableResizeHit(inst, e);
      if (resizeHit) { startTableResize(inst, e, resizeHit); return; }
      var cell = e.target && e.target.closest ? e.target.closest('td,th') : null;
      if (!cell || !inst.editor.contains(cell)) { clearCellSelection(inst); return; }
      var table = cell.closest('table');
      if (inst.selectedTable) { inst.selectedTable.classList.remove('sui-rich-table-selected'); inst.selectedTable = null; }
      var wantsRange = !!(e.shiftKey || e.ctrlKey || e.metaKey);
      var cellPadding = !cellTextHit(cell, e);
      // A normal drag inside the cell text must stay a native text selection,
      // so formatting can be applied to part of the cell. Cell ranges are
      // selected with Shift/Ctrl, or by starting on the cell padding/empty cell.
      if (wantsRange) {
        e.preventDefault();
        inst.cellDragging = true;
        if (inst.cellAnchor && inst.cellAnchor.closest('table') === table) selectCellRange(inst, table, inst.cellAnchor, cell);
        else { inst.selectedCells = [cell]; inst.cellAnchor = cell; updateCellSelection(inst); }
        return;
      }
      if (cellPadding) {
        e.preventDefault();
        inst.cellDragging = true;
        inst.selectedCells = [cell];
        inst.cellAnchor = cell;
        updateCellSelection(inst);
        forceCellCaret(inst, cell);
        return;
      }
      clearCellSelection(inst);
      inst.cellAnchor = cell;
      inst.cellDragging = false;
    });
    inst.editor.addEventListener('mouseup', function (e) {
      if (!inst.cellAnchor || !inst.selectedCells || inst.selectedCells.length !== 1) return;
      var cell = inst.cellAnchor;
      if (!inst.editor.contains(cell)) return;
      if (!cellTextHit(cell, e)) {
        e.preventDefault();
        forceCellCaret(inst, cell);
      }
    });
    inst.editor.addEventListener('click', function (e) {
      var cell = e.target && e.target.closest ? e.target.closest('td,th') : null;
      if (!cell || !inst.editor.contains(cell)) return;
      // Chromium can finalize the caret on click after mousedown; repeat the
      // fix for empty cells and clicks on cell padding.
      if (!cellTextHit(cell, e)) {
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
      var vspan = Number(cell.getAttribute('rowspan') || 1);
      if (span > 1 || vspan > 1) {
        var splitRow = cell.parentNode.rowIndex;
        var splitCol = cell.cellIndex;
        cell.setAttribute('colspan', '1');
        cell.setAttribute('rowspan', '1');
        for (var sri = 0; sri < vspan; sri++) {
          var splitTr = table.rows[splitRow + sri];
          if (!splitTr) break;
          for (var sci = 0; sci < span; sci++) {
            if (sri === 0 && sci === 0) continue;
            var at = Math.min(splitCol + sci, splitTr.cells.length);
            var splitCell = splitTr.insertCell(at);
            splitCell.innerHTML = '<p><br></p>';
          }
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
  function indentListItem(li) {
    var prev = li.previousElementSibling;
    while (prev && prev.tagName.toLowerCase() !== 'li') prev = prev.previousElementSibling;
    if (!prev) return false;
    var list = li.parentNode;
    var nested = null;
    for (var i = 0; i < prev.children.length; i++) {
      var child = prev.children[i];
      if (child.tagName && /^(UL|OL)$/.test(child.tagName)) { nested = child; break; }
    }
    if (!nested) { nested = el(list.tagName.toLowerCase()); prev.appendChild(nested); }
    nested.appendChild(li);
    return true;
  }
  function outdentListItem(li) {
    var list = li.parentNode;
    var parentLi = list.parentNode;
    if (!parentLi || parentLi.tagName.toLowerCase() !== 'li') return false;
    var outer = parentLi.parentNode;
    outer.insertBefore(li, parentLi.nextSibling);
    if (!list.children.length) list.parentNode.removeChild(list);
    return true;
  }
  function changeIndent(inst, delta) {
    var block = currentBlock(inst);
    if (block && block.tagName && block.tagName.toLowerCase() === 'li') {
      if (delta > 0) indentListItem(block); else outdentListItem(block);
      return;
    }
    if (!block) return;
    if (block.tagName && block.tagName.toLowerCase() === 'pre') {
      indentCodeLines(inst, block, delta);
      return;
    }
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
  function pageBreakNode() {
    var node = el('div', 'sui-rich-pagebreak');
    node.setAttribute('contenteditable', 'false');
    node.setAttribute('data-sui-page-break', '1');
    node.title = 'Drag to move this page break';
    var handle = el('span', 'sui-rich-pagebreak-handle', '\u28ff');
    handle.setAttribute('data-rich-pb-handle', '1');
    handle.setAttribute('aria-label', 'Move page break');
    var label = el('span', 'sui-rich-pagebreak-label', 'Page break');
    var del = el('button', 'sui-rich-pagebreak-delete', '\u00d7');
    del.type = 'button';
    del.setAttribute('data-rich-pb-delete', '1');
    del.setAttribute('aria-label', 'Delete page break');
    del.title = 'Delete page break';
    node.appendChild(handle); node.appendChild(label); node.appendChild(del);
    return node;
  }
  function insertPageBreak(inst) {
    var node = pageBreakNode();
    insertBlock(inst, node);
    var next = el('p');
    next.appendChild(el('br'));
    node.parentNode.insertBefore(next, node.nextSibling);
    inst.editor.focus();
    var range = document.createRange();
    range.selectNodeContents(next);
    range.collapse(true);
    var sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
  }
  function materializeBlobImages(inst, scope) {
    var images = scope && scope.querySelectorAll ? scope.querySelectorAll('img[src^="blob:"]') : [];
    for (var i = 0; i < images.length; i++) {
      (function (img) {
        var src = img.getAttribute('src') || '';
        if (!src || !window.fetch) return;
        fetch(src).then(function (response) { return response.blob(); }).then(function (blob) {
          return new Promise(function (resolve, reject) {
            var reader = new FileReader();
            reader.onload = function () { resolve(String(reader.result || '')); };
            reader.onerror = reject;
            reader.readAsDataURL(blob);
          });
        }).then(function (dataUrl) {
          if (!dataUrl || !img.parentNode) return;
          img.setAttribute('src', dataUrl);
          commit(inst, 'change');
        }).catch(function () {});
      })(images[i]);
    }
  }
  function insertDocumentAtCaret(inst, doc) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount || !inst.editor.contains(sel.anchorNode)) {
      var blocks = doc && Array.isArray(doc.children) ? doc.children : [];
      for (var bi = 0; bi < blocks.length; bi++) { var bb = blockDom(blocks[bi], 0); if (bb) inst.editor.appendChild(bb); }
      return;
    }
    var range = sel.getRangeAt(0);
    range.deleteContents();
    var blocks2 = doc && Array.isArray(doc.children) ? doc.children : [];
    if (blocks2.length === 1 && (blocks2[0].type === 'paragraph' || blocks2[0].type === 'heading') && !(blocks2[0].align || blocks2[0].indent || blocks2[0].spaceBefore || blocks2[0].spaceAfter || blocks2[0].lineHeight)) {
      insertInlineAtCaret(inst, blocks2[0].children, range);
      return;
    }
    var block = currentBlock(inst);
    if (block && block.tagName.toLowerCase() === 'td') {
      var cell = block;
      for (var ci = 0; ci < blocks2.length; ci++) { var cb = blockDom(blocks2[ci], 0); if (cb) insertIntoCell(cell, cb); }
      return;
    }
    var after = block && block.parentNode === inst.editor ? block : null;
    var reference = after ? after.nextSibling : null;
    var last = null;
    for (var i = 0; i < blocks2.length; i++) {
      var node = blockDom(blocks2[i], 0);
      if (!node) continue;
      inst.editor.insertBefore(node, reference);
      last = node;
    }
    if (last) {
      var caret = document.createRange();
      caret.selectNodeContents(last);
      caret.collapse(false);
      sel.removeAllRanges();
      sel.addRange(caret);
    }
  }
  function insertIntoCell(cell, node) {
    if (node.tagName && /^(P|DIV|TABLE|UL|OL|PRE|BLOCKQUOTE)$/.test(node.tagName)) {
      var p = cell.querySelector('p');
      if (p) p.parentNode.insertBefore(node, p.nextSibling);
      else cell.appendChild(node);
    } else cell.appendChild(node);
  }
  function insertInlineAtCaret(inst, nodes, range) {
    var sel = window.getSelection();
    var frag = document.createDocumentFragment();
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (n.type === 'text') frag.appendChild(applyMarks(n.text || '', n.marks || []));
      else if (n.type === 'break') frag.appendChild(el('br'));
      else if (n.type === 'image') {
        var img = el('img');
        var src = safeUrl(n.src, true);
        if (src) { img.setAttribute('src', src); if (n.alt) img.setAttribute('alt', String(n.alt)); if (n.width) img.style.width = cssLength(n.width, ''); }
        else continue;
        frag.appendChild(img);
      }
    }
    var lastNode = frag.lastChild;
    range.insertNode(frag);
    if (lastNode) {
      var after = document.createRange();
      after.setStartAfter(lastNode);
      after.collapse(true);
      if (sel) { sel.removeAllRanges(); sel.addRange(after); }
    }
  }
  function insertBlock(inst, node) {
    var block = currentBlock(inst);
    if (block && block.tagName.toLowerCase() === 'li') {
      var list = block.closest('ul,ol');
      if (list && list.parentNode) { list.parentNode.insertBefore(node, list.nextSibling); return; }
    }
    if (block && block.tagName && block.tagName.toLowerCase() === 'td') {
      insertIntoCell(block, node);
      return;
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
    var quote = quoteAncestor(block, inst.editor);
    if (quote) {
      var emptyLine = !block.textContent.replace(/\u200b/g, '').trim() && !block.querySelector('img,table,hr');
      if (emptyLine && block === quote.lastElementChild) {
        // Enter on the empty last line of a quote exits the quote.
        var exit = el('p');
        exit.appendChild(el('br'));
        quote.parentNode.insertBefore(exit, quote.nextSibling);
        quote.removeChild(block);
        if (!quote.textContent.trim() && !quote.querySelector('img,table,hr,ul,ol,table')) quote.parentNode.removeChild(quote);
        inst.editor.focus();
        var exitRange = document.createRange();
        exitRange.selectNodeContents(exit);
        exitRange.collapse(true);
        sel.removeAllRanges();
        sel.addRange(exitRange);
        return;
      }
    }
    if (tag === 'pre' || tag === 'table' || tag === 'td' || tag === 'th') {
      document.execCommand('insertLineBreak');
      return;
    }
    var listItem = block.closest ? block.closest('li') : null;
    if (listItem && listItem !== block && inst.editor.contains(listItem)) {
      var emptyItem = !block.textContent.replace(/\u200b/g, '').trim() && !block.querySelector('img,table,hr');
      if (emptyItem && !listItem.querySelector('ul,ol')) {
        // Enter on an empty list item leaves the list at that point.
        var listEl = listItem.parentNode;
        var exit = el('p');
        exit.appendChild(el('br'));
        if (listEl.children.length === 1) {
          listEl.parentNode.replaceChild(exit, listEl);
        } else {
          var rest = listEl.cloneNode(false);
          var restItem = listItem.nextSibling;
          while (restItem) { var restNext = restItem.nextSibling; rest.appendChild(restItem); restItem = restNext; }
          listEl.removeChild(listItem);
          listEl.parentNode.insertBefore(exit, listEl.nextSibling);
          if (rest.children.length) listEl.parentNode.insertBefore(rest, exit.nextSibling);
        }
        inst.editor.focus();
        var exitRange = document.createRange();
        exitRange.selectNodeContents(exit);
        exitRange.collapse(true);
        sel.removeAllRanges();
        sel.addRange(exitRange);
        return;
      }
      var afterItem = range.cloneRange();
      afterItem.setStart(range.endContainer, range.endOffset);
      afterItem.setEnd(block, block.childNodes.length);
      var itemFragment = afterItem.extractContents();
      var itemNext = el(tag === 'li' ? 'p' : tag);
      itemNext.appendChild(itemFragment);
      if (!itemNext.textContent && !itemNext.querySelector('img,table,hr,br')) itemNext.appendChild(el('br'));
      block.parentNode.insertBefore(itemNext, block.nextSibling);
      var newItem = el('li');
      newItem.appendChild(itemNext);
      var follower = itemNext.nextSibling;
      while (follower) { var followerNext = follower.nextSibling; newItem.appendChild(follower); follower = followerNext; }
      listItem.parentNode.insertBefore(newItem, listItem.nextSibling);
      inst.editor.focus();
      var itemRange = document.createRange();
      itemRange.selectNodeContents(itemNext);
      itemRange.collapse(true);
      sel.removeAllRanges();
      sel.addRange(itemRange);
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
  // selectionOrWordRange returns the current selection, or the word under a
  // collapsed caret. Clear formatting and copy formatting use it so they act
  // on the selection first, then on the word at the caret.
  function selectionOrWordRange(inst) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return null;
    var range = sel.getRangeAt(0);
    if (!range.collapsed) return range.cloneRange();
    var node = range.startContainer, offset = range.startOffset;
    if (node.nodeType !== 3) {
      var before = node.childNodes[offset - 1], after = node.childNodes[offset];
      if (after && after.nodeType === 3) { node = after; offset = 0; }
      else if (before && before.nodeType === 3) { node = before; offset = (before.nodeValue || '').length; }
      else return range.cloneRange();
    }
    var text = String(node.nodeValue || '');
    var start = Math.min(offset, text.length), end = start;
    while (start > 0 && /[^\s]/.test(text.charAt(start - 1))) start--;
    while (end < text.length && /[^\s]/.test(text.charAt(end))) end++;
    if (start === end) return range.cloneRange();
    var word = document.createRange();
    word.setStart(node, start); word.setEnd(node, end);
    return word;
  }
  function selectRange(range) {
    var sel = window.getSelection();
    if (!sel || !range) return;
    sel.removeAllRanges(); sel.addRange(range);
  }
  function clearInlineRange(inst, range, keepLinks) {
    if (!range || range.collapsed) return;
    selectRange(range);
    var nodes = inst.editor.querySelectorAll('span,font,strong,b,em,i,u,s,strike,code,sup,sub,a');
    for (var i = nodes.length - 1; i >= 0; i--) {
      var node = nodes[i];
      if (keepLinks && node.tagName.toLowerCase() === 'a') continue;
      var intersects = false;
      try { intersects = range.intersectsNode(node); } catch (e) { intersects = true; }
      if (!intersects || !node.parentNode) continue;
      while (node.firstChild) node.parentNode.insertBefore(node.firstChild, node);
      node.parentNode.removeChild(node);
    }
    try { document.execCommand('removeFormat'); } catch (e) {}
  }
  function firstTextInRange(range) {
    if (!range) return null;
    if (range.startContainer.nodeType === 3) return range.startContainer;
    var root = range.commonAncestorContainer;
    if (root.nodeType === 3) return root;
    var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, null), node;
    while ((node = walker.nextNode())) {
      var hit = false;
      try { hit = range.intersectsNode(node); } catch (e) { hit = true; }
      if (hit) return node;
    }
    return null;
  }
  function changeFontSize(inst, delta) {
    var range = selectionOrWordRange(inst);
    if (!range || range.collapsed) return;
    var textStart = firstTextInRange(range);
    var start = textStart ? textStart.parentElement : (range.startContainer.nodeType === 1 ? range.startContainer : range.startContainer.parentElement);
    var current = parseFloat(getComputedStyle(start).fontSize) || (inst.baseFontSize || 16);
    var next = Math.max(8, Math.min(96, Math.round((current + delta) * 100) / 100));
    // If the whole selected text already sits in a size wrapper, update that
    // wrapper in place instead of nesting a second size on top of it.
    if (start && start.tagName === 'SPAN' && start.style.fontSize && range.toString().length >= String(start.textContent || '').length) {
      start.style.fontSize = next + 'px';
      inst.editor.focus();
      saveSelection(inst);
      commit(inst, 'change');
      return;
    }
    var fragment = range.extractContents();
    var spans = fragment.querySelectorAll('span');
    for (var i = 0; i < spans.length; i++) if (spans[i].style.fontSize) spans[i].style.fontSize = '';
    var fonts = fragment.querySelectorAll('font[size]');
    for (var f = 0; f < fonts.length; f++) fonts[f].removeAttribute('size');
    var walker = document.createTreeWalker(fragment, NodeFilter.SHOW_TEXT, null);
    var nodes = [], node;
    while ((node = walker.nextNode())) if (node.nodeValue) nodes.push(node);
    for (var n = 0; n < nodes.length; n++) {
      var textNode = nodes[n];
      if (!textNode.parentNode) continue;
      var span = el('span');
      span.style.fontSize = next + 'px';
      textNode.parentNode.insertBefore(span, textNode);
      span.appendChild(textNode);
    }
    if (!fragment.firstChild) return;
    var first = fragment.firstChild, last = fragment.lastChild;
    range.insertNode(fragment);
    // Extract+insert can leave an empty wrapper with the previous size.
    var empties = inst.editor.querySelectorAll('span,font');
    for (var e = empties.length - 1; e >= 0; e--) {
      var empty = empties[e];
      if (!empty.textContent && !empty.querySelector('img,br,table')) { empty.parentNode.removeChild(empty); continue; }
      if (empty.tagName.toLowerCase() === 'span' && empty.getAttribute('style') !== null && !empty.getAttribute('style').trim()) {
        empty.removeAttribute('style');
        if (!empty.attributes.length) {
          while (empty.firstChild) empty.parentNode.insertBefore(empty.firstChild, empty);
          empty.parentNode.removeChild(empty);
        }
      }
    }
    var selected = document.createRange();
    selected.setStartBefore(first); selected.setEndAfter(last);
    selectRange(selected);
    inst.editor.focus();
    saveSelection(inst);
    commit(inst, 'change');
  }
  function clearFormatting(inst) {
    var range = selectionOrWordRange(inst);
    if (!range || range.collapsed) return;
    clearInlineRange(inst, range, false);
    var mode = blockKind(inst);
    if (mode === 'blockquote' || mode === 'pre') applyBlock(inst, 'p');
  }
  function unwrapInlineNode(node) {
    if (!node || !node.parentNode) return;
    while (node.firstChild) node.parentNode.insertBefore(node.firstChild, node);
    node.parentNode.removeChild(node);
  }
  function stripInlineMark(inst, cmd, range) {
    if (!range || range.collapsed) return;
    var tags = {
      bold: 'b,strong', italic: 'i,em', underline: 'u', strike: 's,strike,del',
      superscript: 'sup', subscript: 'sub'
    };
    var props = {
      bold: ['fontWeight'], italic: ['fontStyle'], underline: ['textDecoration', 'textDecorationLine'],
      strike: ['textDecoration', 'textDecorationLine'], superscript: ['verticalAlign'], subscript: ['verticalAlign']
    };
    var selector = tags[cmd];
    if (selector) {
      var nodes = inst.editor.querySelectorAll(selector);
      for (var i = nodes.length - 1; i >= 0; i--) {
        var hit = false;
        try { hit = range.intersectsNode(nodes[i]); } catch (e) { hit = false; }
        if (hit) unwrapInlineNode(nodes[i]);
      }
    }
    var styled = inst.editor.querySelectorAll('[style]');
    var remove = props[cmd] || [];
    for (var s = 0; s < styled.length; s++) {
      var elNode = styled[s];
      var intersects = false;
      try { intersects = range.intersectsNode(elNode); } catch (e) { intersects = false; }
      if (!intersects || !elNode.style) continue;
      if ((cmd === 'underline' || cmd === 'strike') && (elNode.style.textDecoration || elNode.style.textDecorationLine)) {
        var wanted = cmd === 'underline' ? 'underline' : 'line-through';
        var parts = String(elNode.style.textDecorationLine || elNode.style.textDecoration || '').split(/\s+/).filter(function (part) { return part && part !== wanted; });
        if (parts.length) elNode.style.textDecoration = parts.join(' ');
        else { elNode.style.removeProperty('text-decoration'); elNode.style.removeProperty('text-decoration-line'); }
      } else {
        for (var p = 0; p < remove.length; p++) elNode.style.removeProperty(remove[p].replace(/[A-Z]/g, function (m) { return '-' + m.toLowerCase(); }));
      }
      if (!elNode.getAttribute('style')) elNode.removeAttribute('style');
      if (elNode.tagName.toLowerCase() === 'span' && !elNode.attributes.length) unwrapInlineNode(elNode);
    }
  }
  // toggleInlineMark uses the native toggle first (it handles collapsed
  // carets and mixed selections) and strips the mark explicitly if the
  // browser left the "on" state in place, so a second click always turns
  // bold/italic/underline/strike off.
  function applyRangeInlineStyle(inst, prop, value) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount || sel.isCollapsed || !inst.editor) return false;
    var range = sel.getRangeAt(0);
    if (!inst.editor.contains(range.commonAncestorContainer)) return false;
    var probe = range.cloneContents();
    if (probe.querySelector && probe.querySelector('p,div,table,pre,ul,ol,h1,h2,h3,h4,blockquote')) return false;
    var span = el('span');
    span.style[prop] = value;
    try {
      span.appendChild(range.extractContents());
      range.insertNode(span);
    } catch (e) { return false; }
    var selected = document.createRange();
    selected.selectNodeContents(span);
    sel.removeAllRanges();
    sel.addRange(selected);
    saveSelection(inst);
    return true;
  }
  function beginColorSelection(inst, input, cmd) {
    var sel = window.getSelection();
    var range = null;
    if (sel && sel.rangeCount && inst.editor && inst.editor.contains(sel.anchorNode)) {
      range = sel.getRangeAt(0).cloneRange();
    } else if (inst.savedSelection) {
      restoreSelection(inst);
      sel = window.getSelection();
      if (sel && sel.rangeCount && inst.editor.contains(sel.anchorNode)) range = sel.getRangeAt(0).cloneRange();
    }
    inst.colorSelection = { cmd: cmd, range: range, input: input };
    inst.colorPreview = null;
    inst.colorCommittedValue = input ? input.value : '';
  }
  function previewColor(inst, cmd, value) {
    var info = inst.colorSelection;
    if (!info || info.cmd !== cmd) return;
    var prop = cmd === 'textColor' ? 'color' : 'backgroundColor';
    var range = info.range;
    if (!range) return;
    if (range.collapsed) {
      if (cmd === 'textColor') inst.typingColor = value;
      else inst.typingHighlight = value;
      return;
    }
    if (inst.colorPreview && inst.colorPreview.cmd === cmd && inst.colorPreview.span && inst.colorPreview.span.parentNode) {
      inst.colorPreview.span.style[prop] = value;
      inst.colorPreview.value = value;
      return;
    }
    var probe = range.cloneContents();
    if (probe.querySelector && probe.querySelector('p,div,table,pre,ul,ol,h1,h2,h3,h4,blockquote')) return;
    var span = el('span');
    span.style[prop] = value;
    try {
      span.appendChild(range.extractContents());
      range.insertNode(span);
    } catch (e) { return; }
    var selected = document.createRange();
    selected.selectNodeContents(span);
    var sel = window.getSelection();
    if (sel) { sel.removeAllRanges(); sel.addRange(selected); }
    info.range = selected.cloneRange();
    inst.colorPreview = { cmd: cmd, span: span, value: value };
    saveSelection(inst);
  }
  function finalizeColor(inst) {
    var info = inst.colorSelection;
    if (!info) return;
    var cmd = info.cmd;
    var input = info.input;
    var value = input ? input.value : '';
    // runBuiltin uses the locked range and any preview span before the state
    // is cleared, so the color is kept even if focus moved outside the editor.
    command(inst.name, cmd, value);
    inst.colorSelection = null;
    inst.colorPreview = null;
    inst.colorCommittedValue = value;
  }
  var COLOR_PRESETS = ['#000000', '#ffffff', '#e03131', '#f08c00', '#ffd43b', '#2f9e44', '#0ca678', '#1971c2', '#5f3dc4', '#868e96', '#495057', '#f783ac'];
  function rememberRecentColor(inst, color) {
    var hex = rgbHex(color || '') || '';
    if (!/^#[0-9a-f]{6}$/i.test(hex)) return;
    var key = 'sui-rich-colors-' + inst.name;
    var list = [];
    try { list = parseJSON(localStorage.getItem(key) || '[]', []) || []; } catch (e) { list = []; }
    list = [hex].concat(list.filter(function (item) { return item !== hex; })).slice(0, 12);
    try { localStorage.setItem(key, JSON.stringify(list)); } catch (e) {}
  }
  function closeColorPalette() {
    var panel = document.querySelector('.sui-rich-color-palette');
    if (panel && panel.parentNode) panel.parentNode.removeChild(panel);
  }
  function beginPaletteSelection(inst, input, cmd, anchor) {
    var sel = window.getSelection();
    var range = null;
    if (sel && sel.rangeCount && inst.editor && inst.editor.contains(sel.anchorNode)) range = sel.getRangeAt(0).cloneRange();
    else if (inst.savedSelection) {
      restoreSelection(inst);
      sel = window.getSelection();
      if (sel && sel.rangeCount && inst.editor.contains(sel.anchorNode)) range = sel.getRangeAt(0).cloneRange();
    }
    closeColorPalette();
    var panel = el('div', 'sui-rich-color-palette');
    panel.setAttribute('role', 'dialog');
    panel.setAttribute('aria-label', cmd === 'textColor' ? 'Text color palette' : 'Highlight color palette');
    applyChromeTheme(inst, panel);
    var title = el('div', 'sui-rich-color-palette-title', cmd === 'textColor' ? 'Text color' : 'Highlight color');
    panel.appendChild(title);
    var recent = [];
    try { recent = parseJSON(localStorage.getItem('sui-rich-colors-' + inst.name) || '[]', []) || []; } catch (e) { recent = []; }
    function grid(colors) {
      var box = el('div', 'sui-rich-color-grid');
      colors.forEach(function (color) {
        var sw = el('button', 'sui-rich-color-swatch');
        sw.type = 'button';
        sw.style.backgroundColor = color;
        sw.title = color;
        sw.setAttribute('aria-label', color);
        sw.addEventListener('click', function (ev) {
          ev.preventDefault();
          var target = { cmd: cmd, range: range, input: input };
          inst.colorSelection = target;
          input.value = color;
          var wrap = input.closest('.sui-rich-color-control');
          if (wrap) wrap.style.setProperty('--sui-rich-color-current', color);
          command(inst.name, cmd, color);
          inst.colorSelection = null;
          inst.colorPreview = null;
          inst.colorCommittedValue = color;
          rememberRecentColor(inst, color);
          closeColorPalette();
        });
        box.appendChild(sw);
      });
      return box;
    }
    panel.appendChild(grid(COLOR_PRESETS));
    if (recent.length) {
      panel.appendChild(el('div', 'sui-rich-color-palette-title', 'Recent'));
      panel.appendChild(grid(recent));
    }
    var custom = el('button', 'sui-rich-btn sui-rich-color-custom', 'Custom...');
    custom.type = 'button';
    custom.addEventListener('click', function (ev) {
      ev.preventDefault();
      closeColorPalette();
      beginColorSelection(inst, input, cmd);
      saveSelection(inst);
      try {
        if (typeof input.showPicker === 'function') input.showPicker();
        else input.click();
      } catch (e) { try { input.click(); } catch (x) {} }
    });
    panel.appendChild(custom);
    document.body.appendChild(panel);
    var rect = anchor.getBoundingClientRect();
    var width = 196;
    var left = Math.max(8, Math.min(window.innerWidth - width - 8, rect.left));
    var top = rect.bottom + 6;
    if (top + panel.offsetHeight > window.innerHeight) top = Math.max(8, rect.top - panel.offsetHeight - 6);
    panel.style.left = left + 'px';
    panel.style.top = top + 'px';
    if (!window.__suiRichPaletteOutside) {
      window.__suiRichPaletteOutside = true;
      document.addEventListener('mousedown', function (ev) {
        var p = document.querySelector('.sui-rich-color-palette');
        if (!p) return;
        if (p.contains(ev.target) || (ev.target.closest && ev.target.closest('.sui-rich-color-palette-btn'))) return;
        closeColorPalette();
      });
    }
  }
  function rememberTypingColor(inst, cmd, value) {
    if (cmd === 'textColor') inst.typingColor = value;
    else if (cmd === 'highlight') inst.typingHighlight = value;
    rememberRecentColor(inst, value);
  }
  function applyTypingState(inst) {
    var native = { bold: 'bold', italic: 'italic', underline: 'underline', strike: 'strikeThrough', superscript: 'superscript', subscript: 'subscript' };
    try { document.execCommand('styleWithCSS', false, true); } catch (e) {}
    for (var cmd in native) {
      if (!hasOwn(native, cmd) || !inst.typingMarks || inst.typingMarks[cmd] === undefined) continue;
      var state = false;
      try { state = document.queryCommandState(native[cmd]); } catch (e) { state = false; }
      if (state !== !!inst.typingMarks[cmd]) {
        try { document.execCommand(native[cmd], false, null); } catch (e) {}
      }
    }
    if (inst.typingColor) { try { document.execCommand('foreColor', false, inst.typingColor); } catch (e) {} }
    if (inst.typingHighlight) {
      try { document.execCommand('hiliteColor', false, inst.typingHighlight); } catch (e) {}
      try { document.execCommand('backColor', false, inst.typingHighlight); } catch (e) {}
    }
  }
  function markAncestor(inst, cmd) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount || !inst.editor || !inst.editor.contains(sel.anchorNode)) return null;
    var node = sel.anchorNode;
    if (node && node.nodeType !== 1) node = node.parentElement;
    var tags = {
      bold: ['b', 'strong'], italic: ['i', 'em'], underline: ['u'], strike: ['s', 'strike', 'del'],
      superscript: ['sup'], subscript: ['sub'], code: ['code']
    }[cmd] || [];
    while (node && node !== inst.editor) {
      if (node.nodeType === 1) {
        var tag = node.tagName ? node.tagName.toLowerCase() : '';
        if (tags.indexOf(tag) >= 0) return node;
        var style = node.style || {};
        if (cmd === 'bold' && style.fontWeight && (style.fontWeight === 'bold' || parseInt(style.fontWeight, 10) >= 600)) return node;
        if (cmd === 'italic' && style.fontStyle === 'italic') return node;
        if ((cmd === 'underline' || cmd === 'strike') && (style.textDecoration || style.textDecorationLine)) {
          var wanted = cmd === 'underline' ? 'underline' : 'line-through';
          if (String(style.textDecorationLine || style.textDecoration).indexOf(wanted) >= 0) return node;
        }
        if ((cmd === 'superscript' || cmd === 'subscript') && (style.verticalAlign === 'super' || style.verticalAlign === 'sub')) return node;
      }
      node = node.parentElement;
    }
    return null;
  }
  function splitMarkAtCaret(inst, mark, range) {
    if (!mark || !mark.parentNode || !range) return;
    var after = document.createRange();
    after.setStart(range.startContainer, range.startOffset);
    after.setEnd(mark, mark.childNodes.length);
    var fragment = after.extractContents();
    var clone = null;
    if (fragment.textContent || fragment.querySelector('img,br,table')) {
      clone = mark.cloneNode(false);
      clone.appendChild(fragment);
      mark.parentNode.insertBefore(clone, mark.nextSibling);
    }
    var isEmpty = !mark.textContent && !mark.querySelector('img,br,table');
    var caret = document.createRange();
    if (isEmpty && clone) caret.setStartBefore(clone);
    else if (mark.parentNode) caret.setStartAfter(mark);
    else if (clone) caret.setStartBefore(clone);
    caret.collapse(true);
    if (isEmpty && mark.parentNode) mark.parentNode.removeChild(mark);
    var sel = window.getSelection();
    if (sel) { sel.removeAllRanges(); sel.addRange(caret); }
  }
  function toggleInlineCode(inst) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return;
    var range = sel.getRangeAt(0);
    if (!sel.isCollapsed) {
      if (!unwrapTag(inst, 'code')) wrapSelection(inst, 'code');
      if (!inst.typingMarks) inst.typingMarks = {};
      inst.typingMarks.code = !!closestIn(inst, 'code');
      return;
    }
    var inCode = !!closestIn(inst, 'code');
    var desired = inst.typingMarks && inst.typingMarks.code !== undefined ? !inst.typingMarks.code : !inCode;
    if (!desired) {
      var mark = markAncestor(inst, 'code');
      if (mark) {
        splitMarkAtCaret(inst, mark, range);
        inst.suppressNotify = true;
      }
    }
    if (!inst.typingMarks) inst.typingMarks = {};
    inst.typingMarks.code = desired;
    if (inst.editor) inst.editor.focus();
    saveSelection(inst);
  }
  function insertTypedCode(inst, text) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return;
    var range = sel.getRangeAt(0);
    range.deleteContents();
    var code = el('code');
    code.appendChild(document.createTextNode(String(text || '')));
    range.insertNode(code);
    var caret = document.createRange();
    caret.selectNodeContents(code);
    caret.collapse(false);
    sel.removeAllRanges();
    sel.addRange(caret);
    saveSelection(inst);
    commit(inst, 'input');
  }
  function toggleInlineMark(inst, cmd) {
    var native = {
      bold: 'bold', italic: 'italic', underline: 'underline',
      strike: 'strikeThrough', superscript: 'superscript', subscript: 'subscript'
    }[cmd];
    if (!native) return;
    var on = false;
    try { on = document.queryCommandState(native); } catch (e) { on = false; }
    var sel = window.getSelection();
    var collapsedMark = (on && sel && sel.rangeCount && sel.isCollapsed) ? markAncestor(inst, cmd) : null;
    try { document.execCommand(native, false, null); } catch (e) {}
    var still = false;
    try { still = document.queryCommandState(native); } catch (e) { still = false; }
    var split = false;
    if (on && collapsedMark && collapsedMark.parentNode && sel && sel.rangeCount && sel.isCollapsed) {
      splitMarkAtCaret(inst, collapsedMark, sel.getRangeAt(0));
      split = true;
      inst.suppressNotify = true;
      // The caret may still sit next to the closed mark; force the native
      // typing state off so the next typed characters are plain.
      try {
        if (document.queryCommandState(native)) document.execCommand(native, false, null);
      } catch (e) {}
    }
    if (on && still && !split) {
      if (sel && sel.rangeCount) stripInlineMark(inst, cmd, sel.getRangeAt(0));
    }
    var finalState = false;
    try { finalState = document.queryCommandState(native); } catch (e) { finalState = false; }
    if (!inst.typingMarks) inst.typingMarks = {};
    inst.typingMarks[cmd] = finalState;
    if (inst.editor) inst.editor.focus();
    saveSelection(inst);
  }
  function captureFormat(inst, range) {
    var node = range.startContainer;
    if (node.nodeType === 3) node = node.parentElement;
    if (!node || !inst.editor.contains(node)) return null;
    var cs = getComputedStyle(node);
    var decoration = String(cs.textDecorationLine || cs.textDecoration || '');
    var background = String(cs.backgroundColor || '');
    return {
      bold: parseInt(cs.fontWeight, 10) >= 600,
      italic: cs.fontStyle === 'italic',
      underline: decoration.indexOf('underline') >= 0,
      strike: decoration.indexOf('line-through') >= 0,
      code: !!node.closest('code'),
      sup: !!node.closest('sup'),
      sub: !!node.closest('sub'),
      fontFamily: cs.fontFamily || '',
      fontSize: cs.fontSize || '',
      color: cs.color || '',
      background: background && background !== 'rgba(0, 0, 0, 0)' && background !== 'transparent' ? background : ''
    };
  }
  function cancelFormatPainter(inst) {
    inst.painter = null;
    if (inst.root) inst.root.removeAttribute('data-rich-painter');
    updateStats(inst);
  }
  function copyFormat(inst) {
    if (inst.painter) { cancelFormatPainter(inst); return; }
    var range = selectionOrWordRange(inst);
    if (!range || range.collapsed) return;
    var format = captureFormat(inst, range);
    if (!format) return;
    inst.painter = format;
    if (inst.root) inst.root.setAttribute('data-rich-painter', '1');
    updateStats(inst);
  }
  function applyFormatPainter(inst) {
    if (!inst.painter) return;
    var format = inst.painter;
    var range = selectionOrWordRange(inst);
    if (!range || range.collapsed || !inst.editor.contains(range.startContainer)) {
      cancelFormatPainter(inst);
      return;
    }
    inst.painter = null;
    if (inst.root) inst.root.removeAttribute('data-rich-painter');
    clearInlineRange(inst, range, true);
    var span = el('span');
    span.style.fontWeight = format.bold ? '700' : '400';
    span.style.fontStyle = format.italic ? 'italic' : 'normal';
    span.style.textDecoration = (format.underline || format.strike)
      ? ((format.underline ? 'underline ' : '') + (format.strike ? 'line-through' : '')).trim()
      : 'none';
    if (format.fontFamily) span.style.fontFamily = format.fontFamily;
    if (format.fontSize) span.style.fontSize = format.fontSize;
    if (format.color) span.style.color = format.color;
    if (format.background) span.style.backgroundColor = format.background;
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount) return;
    var target = sel.getRangeAt(0);
    if (target.collapsed) return;
    try { span.appendChild(target.extractContents()); target.insertNode(span); }
    catch (e) { return; }
    if (format.code) {
      var code = el('code');
      span.parentNode.insertBefore(code, span);
      code.appendChild(span);
    }
    if (format.sup || format.sub) {
      var wrapper = el(format.sup ? 'sup' : 'sub');
      var parent = span.parentNode;
      parent.insertBefore(wrapper, span);
      wrapper.appendChild(span);
    }
    var reselect = document.createRange();
    reselect.selectNodeContents(span);
    selectRange(reselect);
    commit(inst, 'change');
  }
  function runBuiltin(inst, cmd, value) {
    var liveEditor = inst.root && inst.root.querySelector('.sui-rich-editor');
    if (liveEditor) inst.editor = liveEditor;
    var editor = inst.editor;
    var selection = window.getSelection();
    if ((cmd === 'textColor' || cmd === 'highlight') && inst.colorSelection && inst.colorSelection.cmd === cmd) {
      var lockedColorRange = inst.colorSelection.range;
      if (lockedColorRange && lockedColorRange.startContainer && editor.contains(lockedColorRange.startContainer)) {
        selection = window.getSelection();
        if (selection) {
          try { selection.removeAllRanges(); selection.addRange(lockedColorRange); } catch (e) {}
        }
      }
    }
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
    if (cmd === 'bold') toggleInlineMark(inst, 'bold');
    else if (cmd === 'italic') toggleInlineMark(inst, 'italic');
    else if (cmd === 'underline') toggleInlineMark(inst, 'underline');
    else if (cmd === 'strike') toggleInlineMark(inst, 'strike');
    else if (cmd === 'superscript') toggleInlineMark(inst, 'superscript');
    else if (cmd === 'subscript') toggleInlineMark(inst, 'subscript');
    else if (cmd === 'code') { toggleInlineCode(inst); }
    else if (cmd === 'clearFormat') clearFormatting(inst);
    else if (cmd === 'copyFormat') copyFormat(inst);
    else if (cmd === 'fontSizeUp') changeFontSize(inst, 2);
    else if (cmd === 'fontSizeDown') changeFontSize(inst, -2);
    else if (cmd === 'paragraph') {
      var blockTag = String(value || '');
      applyBlock(inst, /^(h[1-4]|blockquote|pre)$/.test(blockTag) ? blockTag : 'p');
    }
    else if (/^h[1-4]$/.test(cmd)) applyBlock(inst, cmd);
    else if (cmd === 'blockquote') toggleBlock(inst, 'blockquote');
    else if (cmd === 'codeBlock') toggleBlock(inst, 'pre');
    else if (cmd === 'alignLeft') applyBlockProp(inst, 'textAlign', 'left');
    else if (cmd === 'alignCenter') applyBlockProp(inst, 'textAlign', 'center');
    else if (cmd === 'alignRight') applyBlockProp(inst, 'textAlign', 'right');
    else if (cmd === 'justify') applyBlockProp(inst, 'textAlign', 'justify');
    else if (cmd === 'indent') changeIndent(inst, 24);
    else if (cmd === 'outdent') changeIndent(inst, -24);
    else if (cmd === 'bulletList') toggleList(inst, 'ul');
    else if (cmd === 'orderedList') toggleList(inst, 'ol');
    else if (cmd === 'hr') insertBlock(inst, el('hr'));
    else if (cmd === 'pageBreak') insertPageBreak(inst);
    else if (cmd === 'link') {
      var linkSel = window.getSelection();
      var linkHasSelection = !!(linkSel && linkSel.rangeCount && !linkSel.isCollapsed && editor.contains(linkSel.anchorNode));
      var linkRange = promptRange(inst);
      richPrompt(inst, {
        instructions: linkHasSelection
          ? 'Enter the link URL. The selected text will become the link.'
          : 'Enter the link URL and the text to display. The link is inserted at the caret.',
        fields: linkHasSelection
          ? [{ key: 'url', label: 'URL', value: 'https://', placeholder: 'https://example.com', autocomplete: 'url' }]
          : [{ key: 'url', label: 'URL', value: 'https://', placeholder: 'https://example.com', autocomplete: 'url' }, { key: 'text', label: 'Text', value: '', placeholder: 'Link text' }],
        submitLabel: 'Insert link',
        onSubmit: function (v) {
          var url = safeUrl(v.url, false);
          if (!url) return 'Use an http(s), mailto, tel or relative URL.';
          restorePromptRange(inst, linkRange);
          if (linkHasSelection) {
            document.execCommand('createLink', false, url);
          } else {
            var anchor = el('a', '', String(v.text || v.url || url));
            anchor.setAttribute('href', url);
            insertNode(inst, anchor);
          }
          commit(inst, 'change');
          return null;
        }
      });
    } else if (cmd === 'unlink') document.execCommand('unlink');
    else if (cmd === 'image') {
      var imageRange = promptRange(inst);
      richPrompt(inst, {
        instructions: 'Choose an image from this PC or enter an image URL, then add a short alternative text. The image is inserted at the caret.',
        fields: [
          { key: 'file', label: 'Image from this PC', type: 'file', accept: 'image/*' },
          { key: 'src', label: 'Image URL', value: 'https://', placeholder: 'https://example.com/image.png', autocomplete: 'url' },
          { key: 'alt', label: 'Alternative text', value: '', placeholder: 'Describe the image' }
        ],
        submitLabel: 'Insert image',
        onSubmit: function (v) {
          if (v.file) {
            if (!/^image\//i.test(String(v.file.type || ''))) return 'Choose an image file (PNG, JPEG, GIF, WebP or SVG).';
            if (v.file.size > 8 * 1024 * 1024) return 'The image must be smaller than 8 MB.';
            var reader = new FileReader();
            reader.onload = function () {
              var data = String(reader.result || '');
              if (!/^data:image\//i.test(data)) return;
              restorePromptRange(inst, imageRange);
              insertHTMLNode(inst, 'img', { src: data, alt: v.alt || v.file.name || '' });
              materializeBlobImages(inst, inst.editor);
              commit(inst, 'change');
            };
            reader.readAsDataURL(v.file);
            return null;
          }
          var src = safeUrl(v.src, true);
          if (!src) return 'Choose a file or use an http(s) or data:image URL.';
          restorePromptRange(inst, imageRange);
          insertHTMLNode(inst, 'img', { src: src, alt: v.alt || '' });
          commit(inst, 'change');
          return null;
        }
      });
    } else if (cmd === 'table') { var tableNode = buildTable(3, 3); insertBlock(inst, tableNode); }
    else if (cmd === 'symbol') {
      var symbolRange = promptRange(inst);
      richPrompt(inst, {
        instructions: 'Type the symbol to insert at the caret.',
        fields: [{ key: 'symbol', label: 'Symbol', value: '\u00a9', maxLength: 8, placeholder: '\u00a9' }],
        submitLabel: 'Insert symbol',
        onSubmit: function (v) {
          if (!v.symbol) return 'Type at least one character.';
          restorePromptRange(inst, symbolRange);
          document.execCommand('insertText', false, v.symbol);
          commit(inst, 'change');
          return null;
        }
      });
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
    else if (cmd === 'tableProps') openTableProps(inst);
    else if (cmd === 'tableSelect') tableCommand(inst, 'selectTable');
    else if (cmd === 'fontFamily') document.execCommand('fontName', false, String(value || ''));
    else if (cmd === 'fontSize') document.execCommand('fontSize', false, String(value || '3'));
    else if (cmd === 'textColor') {
      var textColorValue = cssColor(value) || '#000000';
      var textPreview = inst.colorPreview;
      if (textPreview && textPreview.cmd === 'textColor' && textPreview.span && textPreview.span.parentNode) {
        textPreview.span.style.color = textColorValue;
      } else if (!applyRangeInlineStyle(inst, 'color', textColorValue)) {
        try { document.execCommand('foreColor', false, textColorValue); } catch (e) {}
      }
      rememberTypingColor(inst, 'textColor', textColorValue);
    }
    else if (cmd === 'highlight') {
      var highlightValue = cssColor(value) || '#ffffff';
      var bgPreview = inst.colorPreview;
      if (bgPreview && bgPreview.cmd === 'highlight' && bgPreview.span && bgPreview.span.parentNode) {
        bgPreview.span.style.backgroundColor = highlightValue;
      } else if (!applyRangeInlineStyle(inst, 'backgroundColor', highlightValue)) {
        try { document.execCommand('hiliteColor', false, highlightValue); } catch (e) {}
        try { document.execCommand('backColor', false, highlightValue); } catch (e) {}
      }
      rememberTypingColor(inst, 'highlight', highlightValue);
    }
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
    else if (cmd === 'paginatedView') { setPaginated(inst); }
    else if (cmd === 'documentSettings') { openDocumentSettings(inst); }
    else if (cmd === 'headerFooter') { openChromeEditor(inst, 'header'); }
    else if (cmd === 'pageSettings') { openPageSettings(inst); }
    else if (cmd === 'paragraphSettings') { openParagraphSettings(inst); }
    else if (cmd === 'imageSettings') { openImageSettings(inst, imageTarget(inst)); }
    else if (cmd === 'zoomIn') { setZoom(inst, (inst.zoom || 1) + 0.1); }
    else if (cmd === 'zoomOut') { setZoom(inst, (inst.zoom || 1) - 0.1); }
    else if (cmd === 'zoomReset') { setZoom(inst, 1); }
    else if (cmd === 'iconSet') { setIconSet(inst, value); }
    else if (cmd === 'iconSetLine') { setIconSet(inst, 'line'); }
    else if (cmd === 'iconSetBold') { setIconSet(inst, 'bold'); }
    else if (cmd === 'iconSetDuotone') { setIconSet(inst, 'duotone'); }
    else if (cmd === 'iconSetBlocks') { setIconSet(inst, 'blocks'); }
    else if (cmd === 'printPreview') { openPrintPreview(inst); }
    else if (cmd === 'print') { printDocument(inst); }
    else if (cmd === 'find') { openFindBar(inst); if (value) findInDocument(inst, value); }
    else if (cmd === 'findNext') nextMatch(inst, 1);
    else if (cmd === 'findPrev') nextMatch(inst, -1);
    else if (cmd === 'findClose') closeFindBar(inst);
    else if (cmd === 'replace') replaceMatch(inst, value);
    else if (cmd === 'replaceAll') replaceAllMatches(inst, value);
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
    else if (cmd === 'themeDark') return setTheme(inst.name, 'dark')[0];
    else if (cmd === 'themeLight') return setTheme(inst.name, 'light')[0];
    else if (cmd === 'themeAuto') return setTheme(inst.name, 'auto')[0];
    else if (cmd === 'pluginInfo') window.alert('Plugins: ' + (pluginOrder.length ? pluginOrder.join(', ') : 'none'));
    else if (cmd === 'importJSON' || cmd === 'importHTML' || cmd === 'importMarkdown' || cmd === 'importXML') {
      var fmt = cmd.slice(6).toLowerCase();
      richPrompt(inst, {
        instructions: 'Paste the ' + fmt + ' content to import. It replaces the current document.',
        fields: [{ key: 'content', label: fmt.toUpperCase() + ' content', value: '', multiline: true, rows: 10, placeholder: 'Paste here...' }],
        submitLabel: 'Import',
        onSubmit: function (v) {
          var result = importFormat(nameOf(inst), v.content, fmt);
          if (result && result[1] !== null) return result[1];
          return null;
        }
      });
    } else if (/^export(JSON|HTML|Markdown|XML)$/.test(cmd)) {
      var format = cmd.slice(6).toLowerCase();
      var result = exportFormat(nameOf(inst), format);
      if (result[1] === null) { if (navigator.clipboard && navigator.clipboard.writeText) navigator.clipboard.writeText(result[0]); window.alert('Exported to clipboard'); }
      else window.alert(result[1]);
    } else return false;
    return true;
  }

  function nameOf(inst) { return inst.name; }

  function resolveInstance(name) {
    var inst = instances[name];
    if (!inst) return null;
    var roots = document.querySelectorAll('[data-sui-rich-name]');
    for (var i = 0; i < roots.length; i++) {
      if (roots[i].getAttribute('data-sui-rich-name') !== name || !roots[i].isConnected) continue;
      if (roots[i] !== inst.root) {
        inst.root = roots[i];
        var live = roots[i].querySelector('.sui-rich-editor');
        if (live) inst.editor = live;
      }
      break;
    }
    var liveEditor = inst.root && inst.root.querySelector('.sui-rich-editor');
    if (liveEditor) inst.editor = liveEditor;
    return inst;
  }
  function command(name, cmd, value) {
    var inst = resolveInstance(name);
    if (!inst) return [false, 'richtext instance not found: ' + name];
    try {
      var handled = runBuiltin(inst, cmd, value);
      if (!handled) {
        var pluginCmd = commands[cmd];
        if (!pluginCmd) return [false, 'unknown richtext command: ' + cmd];
        var r = typeof pluginCmd.run === 'function' ? pluginCmd.run(inst, value) : runBuiltin(inst, pluginCmd.command, value);
        if (r === false) return [false, 'command failed: ' + cmd];
      }
      if (!/^(tableSelect|tableDock|copy|tableProps|find|findNext|findPrev|findClose|print|printPreview|paginatedView|pageSettings|documentSettings|headerFooter|paragraphSettings|imageSettings|copyFormat|zoomIn|zoomOut|zoomReset|iconSet|iconSetLine|iconSetBold|iconSetDuotone|iconSetBlocks)$/.test(cmd)) commit(inst, 'change');
      return [true, null];
    } catch (e) { return [false, String(e && e.message ? e.message : e)]; }
  }

  function format(name, fmt, value) { return command(name, fmt, value); }

  function getValue(name) {
    var inst = resolveInstance(name);
    if (!inst) return [null, 'richtext instance not found: ' + name];
    sync(inst);
    return [JSON.stringify(inst.doc), null];
  }
  function refreshPageChrome(inst) {
    var head = inst.root.querySelector('.sui-rich-pageheader');
    var foot = inst.root.querySelector('.sui-rich-pagefooter');
    var hasHead = chromeHasContent(inst.headerValue);
    var hasFoot = chromeHasContent(inst.footerValue);
    if (head && !hasHead) { head.parentNode.removeChild(head); head = null; }
    if (foot && !hasFoot) { foot.parentNode.removeChild(foot); foot = null; }
    if (head) renderChromeInto(head, inst, inst.headerValue, 1, 1);
    if (foot) renderChromeInto(foot, inst, inst.footerValue, 1, 1);
    if (!head && hasHead) buildPageChrome(inst);
    if (!foot && hasFoot) buildPageChrome(inst);
  }
  function resolvedHeader(inst, doc) {
    if (inst.headerOverride) return doc && doc.header != null ? doc.header : (inst.headerValue || '');
    return inst.headerAttr || (doc && doc.header != null ? doc.header : '');
  }
  function resolvedFooter(inst, doc) {
    if (inst.footerOverride) return doc && doc.footer != null ? doc.footer : (inst.footerValue || '');
    return inst.footerAttr || (doc && doc.footer != null ? doc.footer : '');
  }
  function chromeDoc(value) {
    if (value && typeof value === 'object') {
      if (Array.isArray(value.children)) return value;
      if (Array.isArray(value.blocks)) return { type: 'doc', children: value.blocks };
      if (value.type) return { type: 'doc', children: [value] };
    }
    var text = String(value == null ? '' : value);
    return { type: 'doc', children: text ? [{ type: 'paragraph', children: [{ type: 'text', text: text, marks: [] }] }] : [] };
  }
  function chromeInlineText(nodes) {
    var out = '';
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (!n) continue;
      if (n.type === 'text') out += String(n.text || '');
      else if (n.type === 'image') out += String(n.alt || '');
      else if (n.type === 'break') out += '\n';
      else if (Array.isArray(n.children)) out += chromeInlineText(n.children);
    }
    return out;
  }
  function chromePlainText(value) {
    if (!value || typeof value !== 'object') return String(value == null ? '' : value);
    var blocks = value.children || value.blocks || [];
    var out = [];
    for (var i = 0; i < blocks.length; i++) {
      var block = blocks[i];
      if (!block) continue;
      if (block.type === 'paragraph' || block.type === 'heading') out.push(chromeInlineText(block.children));
      else if (block.type === 'image') out.push(String(block.alt || ''));
      else if (block.type === 'code_block') out.push(String(block.text || ''));
      else if (block.type === 'page_break') out.push('');
      else if (Array.isArray(block.children)) out.push(chromePlainText({ type: 'doc', children: block.children }));
    }
    return out.join('\n');
  }
  function chromeHasImage(value) {
    if (!value || typeof value !== 'object') return false;
    if (value.type === 'image') return true;
    var groups = [value.children, value.blocks, value.rows];
    for (var g = 0; g < groups.length; g++) {
      var list = groups[g] || [];
      for (var i = 0; i < list.length; i++) {
        var item = list[i];
        if (chromeHasImage(item)) return true;
        if (item && Array.isArray(item.cells)) {
          for (var c = 0; c < item.cells.length; c++) if (chromeHasImage(item.cells[c])) return true;
        }
      }
    }
    return false;
  }
  function chromeHasContent(value) {
    return chromePlainText(value).trim() !== '' || chromeHasImage(value);
  }
  function chromeRender(inst, value, page, total) {
    var box = el('div', 'sui-rich-chrome-content');
    box.appendChild(renderDoc(chromeDoc(value)));
    var walker = document.createTreeWalker(box, NodeFilter.SHOW_TEXT, null);
    var textNodes = [], node;
    while ((node = walker.nextNode())) textNodes.push(node);
    for (var i = 0; i < textNodes.length; i++) textNodes[i].nodeValue = applyPageTokens(textNodes[i].nodeValue, page, total, true, inst);
    var imgs = box.querySelectorAll('img');
    for (var im = 0; im < imgs.length; im++) {
      if (imgs[im].getAttribute('alt')) imgs[im].setAttribute('alt', applyPageTokens(imgs[im].getAttribute('alt'), page, total, true, inst));
      if (imgs[im].getAttribute('title')) imgs[im].setAttribute('title', applyPageTokens(imgs[im].getAttribute('title'), page, total, true, inst));
    }
    return box;
  }
  function renderChromeInto(node, inst, value, page, total) {
    if (!node) return;
    node.innerHTML = '';
    node.appendChild(chromeRender(inst, value, page, total));
    var imgs = node.querySelectorAll('img');
    for (var i = 0; i < imgs.length; i++) {
      if (!imgs[i].complete) {
        (function (img) {
          img.addEventListener('load', function () {
            if (!inst) return;
            refreshChromeHeights(inst);
            if (inst.paginated) schedulePaginate(inst); else schedulePageGuides(inst);
          }, { once: true });
        })(imgs[i]);
      }
    }
  }
  function measureChromeHeight(inst, value, width) {
    if (!chromeHasContent(value)) return 0;
    var probe = el('div', 'sui-rich-chrome-probe');
    probe.style.cssText = 'position:absolute;left:-12000px;top:0;width:' + width + 'px;visibility:hidden;font-size:11px;line-height:1.25;';
    probe.appendChild(chromeRender(inst, value, 1, 1));
    document.body.appendChild(probe);
    var h = probe.getBoundingClientRect().height;
    if (probe.parentNode) probe.parentNode.removeChild(probe);
    return Math.ceil(h);
  }
  function refreshChromeHeights(inst) {
    if (!inst) return;
    var page = pageConfig(inst, 1);
    var size = pageSizeMm(page);
    var width = Math.round(mmToPx(size.w) - mmToPx(page.margins.left) - mmToPx(page.margins.right));
    inst.headerH = chromeHasContent(inst.headerValue) ? Math.max(14, measureChromeHeight(inst, inst.headerValue, width)) : 0;
    inst.footerH = chromeHasContent(inst.footerValue) ? Math.max(14, measureChromeHeight(inst, inst.footerValue, width)) : 0;
  }
  function refreshChromeValues(inst) {
    inst.headerValue = resolvedHeader(inst, inst.doc);
    inst.footerValue = resolvedFooter(inst, inst.doc);
    inst.headerText = chromePlainText(inst.headerValue);
    inst.footerText = chromePlainText(inst.footerValue);
    refreshChromeHeights(inst);
    syncChromePanel(inst);
  }
  function setValue(name, value) {
    var inst = resolveInstance(name);
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var doc = typeof value === 'string' ? parseJSON(value, null) : value;
    if (!doc || typeof doc !== 'object') return [false, 'richtext document must be JSON'];
    // Replacing/importing content keeps the document-level settings unless the
    // incoming document explicitly provides them.
    if (doc.page === undefined && inst.doc && inst.doc.page !== undefined) doc.page = inst.doc.page;
    if (doc.pageOverrides === undefined && inst.doc && inst.doc.pageOverrides !== undefined) doc.pageOverrides = inst.doc.pageOverrides;
    if (doc.header == null && inst.doc && inst.doc.header != null) doc.header = inst.doc.header;
    if (doc.footer == null && inst.doc && inst.doc.footer != null) doc.footer = inst.doc.footer;
    if (doc.header != null) inst.headerOverride = true;
    if (doc.footer != null) inst.footerOverride = true;
    inst.scrollGuardUntil = Date.now() + 300;
    var keepScroll = inst.keepScrollTop || inst.scrollTop || (inst.editor ? inst.editor.scrollTop : 0);
    inst.doc = doc;
    inst.pageBreakSources = [];
    refreshChromeValues(inst);
    hyphenationLanguage = documentLanguage(inst);
    refreshPageChrome(inst);
    inst.editor.innerHTML = '';
    inst.editor.appendChild(renderDoc(inst.doc));
    inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
    applyEditorScroll(inst, keepScroll);
    remember(inst);
    updateStats(inst);
    updatePageGuides(inst);
    if (inst.paginated) schedulePaginate(inst); else schedulePageGuides(inst);
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
    hyphenationLanguage = documentLanguage(inst);
    var div = el('div');
    var kids = inst.doc.children || [];
    for (var i = 0; i < kids.length; i++) { var b = blockDom(kids[i], 0); if (b) div.appendChild(b); }
    var tagged = div.querySelectorAll('[data-sui-table-group]');
    for (var g = 0; g < tagged.length; g++) tagged[g].removeAttribute('data-sui-table-group');
    return div.innerHTML;
  }
  function blocksTextMarkdown(blocks, inlineFn) {
    var parts = [];
    for (var i = 0; i < (blocks || []).length; i++) {
      var block = blocks[i];
      if (block.type === 'paragraph' || block.type === 'heading') parts.push(inlineFn(block.children));
      else if (block.type === 'code_block') parts.push(String(block.text || ''));
      else if (block.type === 'table') {
        var rows = [];
        for (var r = 0; r < (block.rows || []).length; r++) {
          var cells = [];
          for (var c = 0; c < (block.rows[r].cells || []).length; c++) {
            var cell = block.rows[r].cells[c];
            cells.push(Array.isArray(cell.blocks) && cell.blocks.length ? blocksTextMarkdown(cell.blocks, inlineFn) : inlineFn(cell.children));
          }
          rows.push(cells.join(' | '));
        }
        parts.push(rows.join(' / '));
      } else if (block.type === 'image') parts.push('[' + (block.alt || 'image') + ']');
    }
    return parts.join(' ');
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
      else if (n.type === 'page_break') out.push(pad + '<!-- page break -->');
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
          for (var c = 0; c < (rows[r].cells || []).length; c++) {
            var mdCell = rows[r].cells[c];
            cells.push(Array.isArray(mdCell.blocks) && mdCell.blocks.length ? blocksTextMarkdown(mdCell.blocks, inlineText) : inlineText(mdCell.children));
          }
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
      else if (n.type === 'page_break') out.push('<page_break/>');
      else if (n.type === 'image') out.push('<image src="' + xmlEscape(n.src || '') + '" alt="' + xmlEscape(n.alt || '') + '"/>');
      else if (n.type === 'bullet_list' || n.type === 'ordered_list') out.push('<' + (n.type === 'bullet_list' ? 'bullet_list' : 'ordered_list') + '>' + (n.children || []).map(block).join('') + '</' + (n.type === 'bullet_list' ? 'bullet_list' : 'ordered_list') + '>');
      else if (n.type === 'list_item') out.push('<list_item>' + (n.children || []).map(block).join('') + '</list_item>');
      else if (n.type === 'table') {
        out.push('<table>');
        for (var r = 0; r < (n.rows || []).length; r++) {
          out.push('<table_row>');
          for (var c = 0; c < (n.rows[r].cells || []).length; c++) {
            var cell = n.rows[r].cells[c];
            var xmlCellBody = (Array.isArray(cell.blocks) && cell.blocks.length) ? cell.blocks.map(block).join('') : inline(cell.children);
            out.push('<' + (cell.header ? 'header_cell' : 'table_cell') + '>' + xmlCellBody + '</' + (cell.header ? 'header_cell' : 'table_cell') + '>');
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
      else if (tag === 'page_break' || tag === 'page-break') out.push({ type: 'page_break' });
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
            var xmlNested = false;
            for (var xnc = 0; xnc < cell.childNodes.length; xnc++) {
              var xncNode = cell.childNodes[xnc];
              if (xncNode.nodeType === 1 && /^(table|paragraph|heading|bullet_list|ordered_list|code_block|image)$/.test(xncNode.localName)) { xmlNested = true; break; }
            }
            var xmlCell = { type: 'table_cell', header: cell.localName === 'header_cell', colspan: 1, rowspan: 1, children: xmlNested ? [] : xmlInlineNodes(cell) };
            if (xmlNested) { xmlCell.blocks = xmlBlocks(cell); if (!xmlCell.blocks.length) xmlCell.blocks = [{ type: 'paragraph', children: [] }]; }
            row.cells.push(xmlCell);
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
    if (saved) saved.textContent = inst.painter ? 'Format painter: select text' : (inst.dirty ? 'Modified' : 'Ready');
    updateTableBar(inst);
    updateProperties(inst);
    syncBlockControls(inst);
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
  function toolbarGroupLabel(tokens) {
    var list = String(tokens || '').split(',');
    function has(re) { for (var i = 0; i < list.length; i++) if (re.test(list[i].trim())) return true; return false; }
    if (has(/^(undo|redo|copy|cut|paste)$/)) return 'History and clipboard';
    if (has(/^(styleSelect|fontFamily|fontSize|fontSizeUp|fontSizeDown|bold|italic|underline|strike|textColor|highlight|superscript|subscript|code|clearFormat|copyFormat)$/)) return 'Text formatting';
    if (has(/^(align|indent|outdent|lineHeight|spaceBefore|spaceAfter|paragraphSettings|bulletList|orderedList|blockquote|codeBlock|justify)/)) return 'Paragraph';
    if (has(/^(link|unlink|image|imageSettings|table|symbol|hr|pageBreak)$/)) return 'Insert';
    if (has(/^(documentSettings|pageSettings|print|printPreview|find)$/)) return 'Document';
    if (has(/^zoom$/)) return 'Zoom';
    return 'Tools';
  }
  function buildToolbar(inst) {
    var bar = inst.root.querySelector('.sui-rich-toolbar');
    if (!bar) return;
    bar.setAttribute('role', 'toolbar');
    bar.setAttribute('aria-label', 'Formatting toolbar');
    bar.innerHTML = '';
    var items = inst.config.toolbar || defaultToolbar;
    var groups = String(items).split('|');
    for (var g = 0; g < groups.length; g++) {
      if (!groups[g].trim()) continue;
      var group = el('div', 'sui-rich-group sui-rich-tool-group');
      group.setAttribute('data-rich-cat', String(g));
      group.setAttribute('role', 'group');
      group.setAttribute('aria-label', toolbarGroupLabel(groups[g]));
      var handle = el('span', 'sui-rich-tool-handle', '\u28ff');
      handle.title = 'Drag to move this toolbar; drop it on the top bar to dock it';
      handle.setAttribute('data-rich-tool-handle', '1');
      group.appendChild(handle);
      var tokens = groups[g].split(',');
      for (var t = 0; t < tokens.length; t++) {
        var token = tokens[t].trim();
        if (!token) continue;
        if (token === 'styleSelect') group.appendChild(select(inst, 'paragraph', [['', 'Normal'], ['h1', 'Heading 1'], ['h2', 'Heading 2'], ['h3', 'Heading 3'], ['h4', 'Heading 4'], ['blockquote', 'Quote'], ['pre', 'Code block']], 'Block style'));
        else if (token === 'fontFamily') group.appendChild(select(inst, 'fontFamily', [['', 'Font'], ['Arial, sans-serif', 'Arial'], ['Georgia, serif', 'Georgia'], ['Times New Roman, serif', 'Times'], ['Courier New, monospace', 'Courier'], ['system-ui, sans-serif', 'System']], 'Font family'));
        else if (token === 'fontSize') group.appendChild(select(inst, 'fontSize', [['', 'Size'], ['1', 'Small'], ['3', 'Normal'], ['5', 'Large'], ['7', 'Huge']], 'Font size'));
        else if (token === 'lineHeight') group.appendChild(select(inst, 'lineHeight', [['', 'Line height'], ['1.2', '1.2'], ['1.5', '1.5'], ['1.8', '1.8'], ['2', '2.0']], 'Line height'));
        else if (token === 'spaceBefore') group.appendChild(select(inst, 'spaceBefore', [['', 'Space before'], ['0', '0'], ['0.5em', '0.5em'], ['1em', '1em'], ['1.5em', '1.5em']], 'Space before'));
        else if (token === 'spaceAfter') group.appendChild(select(inst, 'spaceAfter', [['', 'Space after'], ['0', '0'], ['0.5em', '0.5em'], ['1em', '1em'], ['1.5em', '1.5em']], 'Space after'));
        else if (token === 'textColor' || token === 'highlight') {
          var colorWrap = el('button', 'sui-rich-color-control');
          colorWrap.type = 'button';
          colorWrap.title = token === 'textColor' ? 'Text color' : 'Highlight color';
          colorWrap.setAttribute('aria-label', colorWrap.title);
          var colorIcon = el('span', 'sui-rich-color-icon');
          colorIcon.innerHTML = ICONS[token] || '';
          var color = el('input', 'sui-rich-color');
          color.type = 'color';
          color.setAttribute('data-rich-cmd', token);
          color.value = token === 'textColor' ? '#000000' : '#ffffff';
          color.title = colorWrap.title;
          color.setAttribute('aria-label', colorWrap.title);
          colorWrap.style.setProperty('--sui-rich-color-current', color.value);
          (function (wrap, input, cmdToken) {
            input.addEventListener('input', function () {
              wrap.style.setProperty('--sui-rich-color-current', input.value);
              previewColor(inst, cmdToken, input.value);
            });
            wrap.addEventListener('click', function (ev) {
              ev.preventDefault();
              if (inst.colorSelection) finalizeColor(inst);
              beginColorSelection(inst, input, cmdToken);
              saveSelection(inst);
              try {
                if (typeof input.showPicker === 'function') input.showPicker();
                else input.click();
              } catch (e) {
                try { input.click(); } catch (x) {}
              }
            });
          })(colorWrap, color, token);
          colorWrap.appendChild(colorIcon);
          colorWrap.appendChild(color);
          group.appendChild(colorWrap);
          var paletteBtn = el('button', 'sui-rich-color-palette-btn', '\u25be');
          paletteBtn.type = 'button';
          paletteBtn.title = 'Choose from a color palette';
          paletteBtn.setAttribute('aria-label', paletteBtn.title);
          (function (paletteInput, cmdToken, anchor) {
            anchor.addEventListener('click', function (ev) {
              ev.preventDefault();
              beginPaletteSelection(inst, paletteInput, cmdToken, anchor);
            });
          })(color, token, paletteBtn);
          group.appendChild(paletteBtn);
        } else if (token === 'zoom') {
          var zoomOutBtn = button(inst, 'zoomOut', '\u2212', TOOL_LABELS.zoomOut);
          zoomOutBtn.setAttribute('data-rich-zoom', 'out');
          var zoomResetBtn = el('button', 'sui-rich-zoom-btn sui-rich-zoom-label', Math.round((inst.zoom || 1) * 100) + '%');
          zoomResetBtn.type = 'button';
          zoomResetBtn.title = TOOL_LABELS.zoomReset;
          zoomResetBtn.setAttribute('aria-label', TOOL_LABELS.zoomReset);
          zoomResetBtn.setAttribute('data-rich-zoom', 'reset');
          var zoomInBtn = button(inst, 'zoomIn', '+', TOOL_LABELS.zoomIn);
          zoomInBtn.setAttribute('data-rich-zoom', 'in');
          group.appendChild(zoomOutBtn); group.appendChild(zoomResetBtn); group.appendChild(zoomInBtn);
        } else group.appendChild(button(inst, token, TOOL_LABELS[token] || token, TOOL_LABELS[token] || token));
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
    var themeGroup = el('div', 'sui-rich-group sui-rich-theme-group');
    var themeBtn = el('button', 'sui-rich-btn sui-rich-theme-btn');
    themeBtn.type = 'button';
    themeBtn.setAttribute('data-rich-theme-toggle', '1');
    themeBtn.setAttribute('aria-label', 'Toggle interface theme');
    themeGroup.appendChild(themeBtn);
    bar.appendChild(themeGroup);
    updateThemeButton(inst);
    bindToolbarDrag(inst);
    restoreToolbarPositions(inst);
    buildTableBar(inst);
  }
  function saveToolbarPositions(inst) {
    var store = {};
    var groups = inst.root.querySelectorAll('.sui-rich-tool-group');
    for (var i = 0; i < groups.length; i++) {
      var cat = groups[i].getAttribute('data-rich-cat');
      if (groups[i].classList.contains('floating')) store[cat] = { x: parseFloat(groups[i].style.left) || 0, y: parseFloat(groups[i].style.top) || 0 };
    }
    try { localStorage.setItem('sui-rich-toolbars-' + inst.name, JSON.stringify(store)); } catch (e) {}
  }
  function restoreToolbarPositions(inst) {
    var store = null;
    try { store = parseJSON(localStorage.getItem('sui-rich-toolbars-' + inst.name) || '', null); } catch (e) { store = null; }
    if (!store) return;
    var groups = inst.root.querySelectorAll('.sui-rich-tool-group');
    for (var i = 0; i < groups.length; i++) {
      var cat = groups[i].getAttribute('data-rich-cat');
      var pos = store[cat];
      if (!pos || !isFinite(pos.x) || !isFinite(pos.y)) continue;
      groups[i].classList.add('floating');
      groups[i].style.left = Math.max(0, Number(pos.x)) + 'px';
      groups[i].style.top = Math.max(0, Number(pos.y)) + 'px';
    }
  }
  function dockToolbar(inst, group, dropX, dropY) {
    var dock = inst.root.querySelector('.sui-rich-toolbar');
    if (dock) {
      var list = [];
      var all = dock.querySelectorAll('.sui-rich-tool-group');
      for (var a = 0; a < all.length; a++) if (all[a] !== group) list.push(all[a]);
      var rowGroups = list;
      if (list.length && typeof dropY === 'number') {
        var sameRow = [];
        for (var r = 0; r < list.length; r++) {
          var rr = list[r].getBoundingClientRect();
          if (Math.abs((rr.top + rr.height / 2) - dropY) <= Math.max(18, rr.height / 2 + 6)) sameRow.push(list[r]);
        }
        if (sameRow.length) rowGroups = sameRow;
      }
      var next = null;
      if (typeof dropX === 'number') {
        for (var i = 0; i < rowGroups.length; i++) {
          var rect = rowGroups[i].getBoundingClientRect();
          if (dropX < rect.left + rect.width / 2) { next = rowGroups[i]; break; }
        }
        if (!next && rowGroups.length) next = rowGroups[rowGroups.length - 1].nextSibling;
      }
      if (next && next !== group) dock.insertBefore(group, next);
      else if (!next) dock.appendChild(group);
    }
    group.classList.remove('floating');
    group.style.left = '';
    group.style.top = '';
    group.style.width = '';
    saveToolbarPositions(inst);
  }
  function bindToolbarDrag(inst) {
    var groups = inst.root.querySelectorAll('.sui-rich-tool-group');
    for (var i = 0; i < groups.length; i++) {
      (function (group) {
        var handle = group.querySelector('[data-rich-tool-handle]');
        if (handle && !handle.__suiBound) {
          handle.__suiBound = true;
          handle.addEventListener('mousedown', function (e) {
            e.preventDefault();
            e.stopPropagation();
            var rect = group.getBoundingClientRect();
            if (!group.classList.contains('floating')) {
              group.style.width = Math.max(60, rect.width) + 'px';
              group.classList.add('floating');
              group.style.left = rect.left + 'px';
              group.style.top = rect.top + 'px';
            }
            var offsetX = e.clientX - rect.left, offsetY = e.clientY - rect.top;
            function move(ev) {
              group.style.left = Math.max(0, Math.min(window.innerWidth - 60, ev.clientX - offsetX)) + 'px';
              group.style.top = Math.max(0, Math.min(window.innerHeight - 30, ev.clientY - offsetY)) + 'px';
            }
            function up(ev) {
              document.removeEventListener('mousemove', move);
              document.removeEventListener('mouseup', up);
              var dockBar = inst.root.querySelector('.sui-rich-toolbar');
              var dockRect = dockBar ? dockBar.getBoundingClientRect() : null;
              var dropX = ev && typeof ev.clientX === 'number' ? ev.clientX : -1;
              var dropY = ev && typeof ev.clientY === 'number' ? ev.clientY : -1;
              if (dockRect && dropX >= dockRect.left && dropX <= dockRect.right && dropY >= dockRect.top - 10 && dropY <= dockRect.bottom + 10) {
                dockToolbar(inst, group, dropX, dropY);
              } else {
                saveToolbarPositions(inst);
              }
            }
            document.addEventListener('mousemove', move);
            document.addEventListener('mouseup', up);
          });
          handle.addEventListener('dblclick', function (e) {
            e.preventDefault();
            dockToolbar(inst, group);
          });
        }
        // Docking happens by dropping the floating toolbar on the top bar.
      })(groups[i]);
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
  function panelTitle(text, hint) {
    var title = el('div', 'sui-rich-tableprops-title', text);
    if (hint) {
      var mark = el('span', 'sui-rich-hint', '?');
      mark.title = hint;
      mark.setAttribute('aria-label', hint);
      title.appendChild(mark);
    }
    return title;
  }
  function tablePropsPanel(inst) {
    var panel = document.querySelector('.sui-rich-tableprops[data-rich-props="' + inst.name + '"]');
    if (!panel) {
      panel = el('div', 'sui-rich-tableprops');
      panel.setAttribute('data-rich-props', inst.name);
      panel.hidden = true;
      document.body.appendChild(panel);
    }
    return panel;
  }
  function clearTableColumnWidths(table) {
    if (!table) return;
    var colgroup = table.querySelector('colgroup');
    if (colgroup && colgroup.parentNode) colgroup.parentNode.removeChild(colgroup);
    var cells = table.querySelectorAll('td,th');
    for (var i = 0; i < cells.length; i++) cells[i].style.width = '';
  }
  function applyTableLook(inst, table, opts) {
    resolveInstance(inst.name);
    if (!table || !table.isConnected) return;
    var cells = table.querySelectorAll('td,th');
    var borderFlags = opts.borders || tableBorderFlags(table);
    if (opts.borders) table.setAttribute('data-table-borders', opts.borders.join(','));
    var borderWidth = opts.borderWidth !== undefined ? opts.borderWidth : (parseFloat(table.getAttribute('data-table-border-width')) || '');
    var borderColor = opts.borderColor || table.style.borderColor || '';
    if (opts.padding !== undefined) {
      if (opts.padding !== '') {
        table.setAttribute('data-table-padding', opts.padding + 'px');
        for (var pc = 0; pc < cells.length; pc++) cells[pc].style.padding = opts.padding + 'px';
      } else {
        table.removeAttribute('data-table-padding');
        for (var pc2 = 0; pc2 < cells.length; pc2++) cells[pc2].style.padding = '';
      }
    }
    if (opts.borderWidth !== undefined) {
      if (opts.borderWidth !== '') {
        table.setAttribute('data-table-border-width', opts.borderWidth + 'px');
        for (var bc = 0; bc < cells.length; bc++) cells[bc].style.borderWidth = opts.borderWidth + 'px';
      } else {
        table.removeAttribute('data-table-border-width');
        for (var bc2 = 0; bc2 < cells.length; bc2++) cells[bc2].style.borderWidth = '';
      }
    }
    if (opts.borderColor) {
      table.style.borderColor = opts.borderColor;
      for (var cc = 0; cc < cells.length; cc++) cells[cc].style.borderColor = opts.borderColor;
    }
    applyTableBorderSides(table, borderFlags, borderWidth, borderColor);
    commit(inst, 'change');
  }
  function openTableProps(inst) {
    var liveRoot = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"]');
    if (liveRoot && liveRoot !== inst.root) {
      inst.root = liveRoot;
      var liveEd = liveRoot.querySelector('.sui-rich-editor');
      if (liveEd) inst.editor = liveEd;
    }
    var table = closestIn(inst, 'table') || (inst.editor && inst.editor.querySelector('table'));
    if (!table) return;
    var tableIndex = 0;
    var allTables = inst.editor.querySelectorAll('table');
    for (var ti = 0; ti < allTables.length; ti++) if (allTables[ti] === table) { tableIndex = ti; break; }
    function currentTable() {
      resolveInstance(inst.name);
      var list = inst.editor.querySelectorAll('table');
      return list[tableIndex] || table;
    }
    var panel = tablePropsPanel(inst);
    var rootStyle = getComputedStyle(inst.root);
    ['--sui-rich-bg', '--sui-rich-fg', '--sui-rich-border', '--sui-rich-muted', '--sui-rich-hover', '--sui-rich-accent'].forEach(function (v) {
      var value = rootStyle.getPropertyValue(v);
      if (value) panel.style.setProperty(v, value.trim());
    });
    panel.innerHTML = '';
    panel.hidden = false;
    if (!window.__suiTablePropsOutside) {
      window.__suiTablePropsOutside = true;
      document.addEventListener('mousedown', function (ev) {
        if (ev.target.closest && (ev.target.closest('.sui-rich-tableprops') || ev.target.closest('[data-rich-cmd="tableProps"]'))) return;
        var panels = document.querySelectorAll('.sui-rich-tableprops');
        for (var pi = 0; pi < panels.length; pi++) panels[pi].hidden = true;
      });
    }
    var title = el('div', 'sui-rich-tableprops-title', 'Table settings');
    panel.appendChild(title);
    function property(label, input) {
      var row = el('div', 'sui-rich-tableprops-row');
      var lb = el('label', '', label);
      row.appendChild(lb);
      row.appendChild(input);
      panel.appendChild(row);
      return input;
    }
    var pad = el('input', 'sui-rich-prop');
    pad.type = 'number'; pad.min = '0'; pad.max = '60'; pad.step = '1';
    pad.value = String(parseFloat(table.getAttribute('data-table-padding')) || '');
    pad.placeholder = 'default';
    pad.onchange = function () { applyTableLook(inst, currentTable(), { padding: String(pad.value).replace(/[^0-9.]/g, '') }); };
    property('Cell padding (px)', pad);
    var borderW = el('input', 'sui-rich-prop');
    borderW.type = 'number'; borderW.min = '0'; borderW.max = '12'; borderW.step = '1';
    borderW.value = String(parseFloat(table.getAttribute('data-table-border-width')) || '');
    borderW.placeholder = 'default';
    borderW.onchange = function () { applyTableLook(inst, currentTable(), { borderWidth: String(borderW.value).replace(/[^0-9.]/g, '') }); };
    property('Border width (px)', borderW);
    var color = el('input', 'sui-rich-prop');
    color.type = 'color';
    var currentColor = table.style.borderColor || '#9aa7b8';
    color.value = /^#[0-9a-f]{6}$/i.test(currentColor) ? currentColor : '#9aa7b8';
    color.onchange = function () { applyTableLook(inst, currentTable(), { borderColor: color.value }); };
    property('Border color', color);
    var initialWidth = String(table.style.width || '');
    var widthMode = (!initialWidth || initialWidth === '100%') ? 'fit' : (/%$/.test(initialWidth) ? 'percent' : 'px');
    var widthSelect = el('select', 'sui-rich-prop');
    [['fit', 'Fit to page'], ['px', 'Pixels'], ['percent', 'Percent']].forEach(function (option) {
      var opt = el('option', '', option[1]);
      opt.value = option[0];
      widthSelect.appendChild(opt);
    });
    widthSelect.value = widthMode;
    var widthValue = el('input', 'sui-rich-prop');
    widthValue.type = 'number'; widthValue.min = '10'; widthValue.max = '2000'; widthValue.step = '1';
    widthValue.value = String(Math.round(parseFloat(initialWidth) || (widthMode === 'percent' ? 100 : (table.getBoundingClientRect().width || 100))));
    widthValue.disabled = widthMode === 'fit';
    function applyWidthSetting() {
      var current = currentTable();
      if (!current) return;
      if (widthSelect.value === 'fit') {
        clearTableColumnWidths(current);
        current.style.width = '100%';
        current.style.maxWidth = '';
        current.style.tableLayout = 'auto';
        commit(inst, 'change');
        return;
      }
      var value = Math.max(10, Math.min(2000, Number(widthValue.value) || 0));
      if (!value) return;
      clearTableColumnWidths(current);
      current.style.width = value + (widthSelect.value === 'percent' ? '%' : 'px');
      current.style.maxWidth = 'none';
      current.style.tableLayout = 'auto';
      commit(inst, 'change');
    }
    widthSelect.onchange = function () {
      widthValue.disabled = widthSelect.value === 'fit';
      if (widthSelect.value === 'fit') applyWidthSetting();
    };
    widthValue.onchange = applyWidthSetting;
    property('Table width', widthSelect);
    property('Width value', widthValue);
    panel.appendChild(el('div', 'sui-rich-tableprops-note', 'Fit to page uses the text width inside the page margins. Drag a column edge to resize one column without changing the others.'));
    var flags = tableBorderFlags(table);
    var sidesBox = el('div', 'sui-rich-tableprops-sides');
    [['top', 'Top'], ['right', 'Right'], ['bottom', 'Bottom'], ['left', 'Left'], ['insideH', 'Inside H'], ['insideV', 'Inside V']].forEach(function (pair) {
      var wrap = el('label', 'sui-rich-tableprops-side');
      var cb = el('input');
      cb.type = 'checkbox';
      cb.checked = flags.indexOf(pair[0]) >= 0;
      cb.onchange = function () {
        var next = flags.filter(function (v) { return v !== pair[0]; });
        if (cb.checked) next.push(pair[0]);
        flags = next;
        var current = currentTable();
        if (!current) return;
        applyTableLook(inst, current, { borders: next });
      };
      wrap.appendChild(cb);
      wrap.appendChild(el('span', '', pair[1]));
      sidesBox.appendChild(wrap);
    });
    panel.appendChild(sidesBox);
    var close = el('button', 'sui-rich-btn sui-rich-tableprops-close', 'Close');
    close.type = 'button';
    close.onclick = function () { panel.hidden = true; };
    panel.appendChild(close);
    var bar = inst.root.querySelector('[data-rich-table-group]');
    var rect = (bar && !bar.hidden ? bar : table).getBoundingClientRect();
    panel.style.left = Math.max(8, Math.min(window.innerWidth - 286, rect.left)) + 'px';
    panel.style.top = Math.min(window.innerHeight - 220, rect.bottom + 6) + 'px';
  }
  function buildTableBar(inst) {
    var toolbar = inst.root && inst.root.querySelector('.sui-rich-toolbar');
    if (!toolbar) return;
    var legacy = inst.root.querySelector('.sui-rich-tablebar');
    if (legacy) { legacy.hidden = true; legacy.innerHTML = ''; }
    var old = toolbar.querySelector('[data-rich-table-group]');
    if (old && old.parentNode) old.parentNode.removeChild(old);
    var group = el('div', 'sui-rich-group sui-rich-tool-group');
    group.setAttribute('data-rich-cat', 'table');
    group.setAttribute('data-rich-table-group', '1');
    group.setAttribute('role', 'group');
    group.setAttribute('aria-label', 'Table actions');
    group.hidden = true;
    var handle = el('span', 'sui-rich-tool-handle', '\u28ff');
    handle.title = 'Drag to move this toolbar; drop it on the top bar to dock it';
    handle.setAttribute('data-rich-tool-handle', '1');
    group.appendChild(handle);
    var specs = [
      ['tableSelect', 'T', 'Select whole table'],
      ['tableRowAbove', '+R', 'Add row above'], ['tableRowBelow', '+R\u2193', 'Add row below'], ['tableRowDelete', '-R', 'Delete row'],
      ['tableColLeft', '+C', 'Add column left'], ['tableColRight', '+C\u2192', 'Add column right'], ['tableColDelete', '-C', 'Delete column'],
      ['tableHeader', 'H', 'Toggle header row'], ['tableMerge', 'M', 'Merge selected cells'], ['tableSplit', 'S', 'Split cell'],
      ['tableAlignLeft', 'L', 'Align cell left'], ['tableAlignCenter', 'C', 'Align cell center'], ['tableAlignRight', 'R', 'Align cell right'],
      ['tableProps', '\u2699', 'Table settings'], ['tableDelete', '\u00d7', 'Delete table']
    ];
    for (var i = 0; i < specs.length; i++) group.appendChild(button(inst, specs[i][0], specs[i][1], specs[i][2]));
    var theme = toolbar.querySelector('.sui-rich-theme-group');
    toolbar.insertBefore(group, theme || null);
    bindToolbarDrag(inst);
    restoreToolbarPositions(inst);
    updateTableBar(inst);
  }
  function updateTableBar(inst) {
    var group = inst.root && inst.root.querySelector('[data-rich-table-group]');
    if (!group) return;
    var table = closestIn(inst, 'table');
    if (!table && inst.selectedTable && inst.selectedTable.isConnected) table = inst.selectedTable;
    if (!table) { group.hidden = true; try { clearCellSelection(inst); } catch (e) {} return; }
    group.hidden = false;
  }
  function dockTableBar(inst) {
    var group = inst.root && inst.root.querySelector('[data-rich-table-group]');
    if (!group) return;
    dockToolbar(inst, group);
    var theme = inst.root.querySelector('.sui-rich-theme-group');
    if (theme && theme.parentNode === group.parentNode) group.parentNode.insertBefore(group, theme);
  }
  function closeMenus(inst) {
    var panels = inst.root.querySelectorAll('.sui-rich-menu-panel');
    for (var i = 0; i < panels.length; i++) panels[i].hidden = true;
  }
  function buildContext(inst) {
    var bar = inst.root.querySelector('.sui-rich-context');
    if (!bar) return;
    bar.innerHTML = '';
    ['bold', 'italic', 'underline', 'strike', 'link', 'highlight', 'indent', 'outdent', 'clearFormat'].forEach(function (cmd) { bar.appendChild(button(inst, cmd, cmd, cmd)); });
  }
  function applyTheme(inst) {
    if (!inst || !inst.root) return;
    var choice = inst.theme || 'dark';
    var resolved = choice;
    if (choice === 'auto') {
      try { resolved = window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'; }
      catch (e) { resolved = 'dark'; }
    }
    inst.root.setAttribute('data-sui-rich-theme', resolved);
    var pickers = inst.root.querySelectorAll('[data-rich-theme]');
    for (var i = 0; i < pickers.length; i++) pickers[i].value = choice;
    var toggle = inst.root.querySelector('[data-rich-theme-toggle]');
    if (toggle) {
      var next = resolved === 'light' ? 'dark' : 'light';
      toggle.textContent = resolved === 'light' ? '\u263e' : '\u2600';
      toggle.title = 'Switch to ' + next + ' mode';
      toggle.setAttribute('aria-label', toggle.title);
      toggle.setAttribute('data-rich-next-theme', next);
    }
  }
  function updateThemeButton(inst) {
    if (!inst || !inst.root) return;
    applyTheme(inst);
  }
  function validIconSet(value) {
    value = String(value || '').toLowerCase();
    for (var i = 0; i < ICON_SETS.length; i++) if (ICON_SETS[i][0] === value) return value;
    return 'line';
  }
  function applyIconSet(inst) {
    if (!inst || !inst.root) return;
    var set = validIconSet(inst.iconSet);
    inst.iconSet = set;
    inst.root.setAttribute('data-sui-rich-icons', set);
    try { localStorage.setItem('sui-rich-icons-' + inst.name, set); } catch (e) {}
  }
  function setIconSet(inst, value) {
    inst.iconSet = validIconSet(value);
    applyIconSet(inst);
    return inst.iconSet;
  }
  function setZoom(inst, value) {
    var next = Math.max(0.6, Math.min(2, Number(value) || 1));
    next = Math.round(next * 100) / 100;
    inst.zoom = next;
    // Zoom is a view-only scale: CSS zoom magnifies characters, images and
    // tables without touching the document font sizes or the JSON model.
    if (inst.editor) {
      if ('zoom' in inst.editor.style) { inst.editor.style.zoom = String(next); inst.editor.style.fontSize = ''; }
      else inst.editor.style.fontSize = ((inst.baseFontSize || 16) * next) + 'px';
    }
    var pages = inst.root && inst.root.querySelector('.sui-rich-pages');
    if (pages) pages.style.zoom = String(next);
    var label = inst.root && inst.root.querySelector('.sui-rich-zoom-label');
    if (label) label.textContent = Math.round(next * 100) + '%';
    updatePageGuides(inst);
    schedulePageGuides(inst);
  }
  function setTheme(name, theme) {
    var inst = instances[name];
    if (!inst) return [false, 'richtext instance not found: ' + name];
    var next = String(theme || '').toLowerCase();
    if (next !== 'dark' && next !== 'light' && next !== 'auto') return [false, 'unknown theme: ' + theme];
    inst.theme = next;
    try { localStorage.setItem('sui-rich-theme-' + name, next); } catch (e) {}
    applyTheme(inst);
    return [true, null];
  }
  function chromeDateText(kind) {
    var now = new Date();
    try {
      if (kind === 'date') return new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'short', day: '2-digit' }).format(now);
      if (kind === 'time') return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit' }).format(now);
      return new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(now);
    } catch (e) {
      var iso = now.toISOString();
      if (kind === 'date') return iso.slice(0, 10);
      if (kind === 'time') return iso.slice(11, 16);
      return iso.slice(0, 16).replace('T', ' ');
    }
  }
  function applyPageTokens(text, page, total, known, inst) {
    var value = String(text || '');
    var vars = {
      page: known ? String(page) : '__SUI_PAGE__',
      pages: known ? String(total) : '__SUI_PAGES__',
      title: inst && inst.title ? String(inst.title) : '',
      name: inst && inst.name ? String(inst.name) : '',
      date: chromeDateText('date'),
      time: chromeDateText('time'),
      datetime: chromeDateText('datetime'),
      words: '',
      chars: ''
    };
    if (inst && inst.doc) {
      try { var stats = docStats(inst.doc); vars.words = String(stats.words); vars.chars = String(stats.chars); } catch (e) {}
    }
    for (var key in vars) {
      if (!hasOwn(vars, key)) continue;
      value = value.replace(new RegExp('\\{' + key + '\\}', 'gi'), vars[key]);
      value = value.replace(new RegExp('\\[' + key + '\\]', 'gi'), vars[key]);
    }
    return value;
  }
  function resolveChromeVariables(text, inst) {
    var value = String(text || '');
    var vars = {
      title: inst && inst.title ? String(inst.title) : '',
      name: inst && inst.name ? String(inst.name) : '',
      date: chromeDateText('date'),
      time: chromeDateText('time'),
      datetime: chromeDateText('datetime'),
      words: '',
      chars: ''
    };
    if (inst && inst.doc) {
      try { var stats = docStats(inst.doc); vars.words = String(stats.words); vars.chars = String(stats.chars); } catch (e) {}
    }
    for (var key in vars) {
      if (!hasOwn(vars, key)) continue;
      value = value.replace(new RegExp('\\{' + key + '\\}', 'gi'), vars[key]);
      value = value.replace(new RegExp('\\[' + key + '\\]', 'gi'), vars[key]);
    }
    return value;
  }
  function sheetChromeText(text, page, total, known, inst) {
    return applyPageTokens(text, page, total, known, inst);
  }
  function defaultPageConfig() {
    return { size: 'a4', orientation: 'portrait', lineHeight: 1.65, lang: '', margins: { top: 15, right: 17, bottom: 15, left: 17 } };
  }
  function clampPageMargin(value, fallback) {
    var n = Number(value);
    if (!isFinite(n)) return fallback;
    return Math.max(0, Math.min(50, Math.round(n * 10) / 10));
  }
  function documentPageConfig(inst) {
    var page = defaultPageConfig();
    var raw = inst && inst.doc && inst.doc.page;
    if (raw && typeof raw === 'object') {
      if (raw.size === 'a4' || raw.size === 'letter') page.size = raw.size;
      if (raw.orientation === 'landscape' || raw.orientation === 'portrait') page.orientation = raw.orientation;
      var lh = Number(raw.lineHeight);
      if (isFinite(lh) && lh >= 0.8 && lh <= 3) page.lineHeight = Math.round(lh * 100) / 100;
      if (typeof raw.lang === 'string' && /^[A-Za-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$/.test(raw.lang.trim())) page.lang = raw.lang.trim();
      var m = raw.margins || {};
      page.margins = {
        top: clampPageMargin(m.top, page.margins.top),
        right: clampPageMargin(m.right, page.margins.right),
        bottom: clampPageMargin(m.bottom, page.margins.bottom),
        left: clampPageMargin(m.left, page.margins.left)
      };
    }
    return page;
  }
  function pageConfig(inst, pageNumber) {
    var page = documentPageConfig(inst);
    if (!pageNumber || !inst || !inst.doc || !inst.doc.pageOverrides) return page;
    var ov = inst.doc.pageOverrides[String(pageNumber)];
    if (!ov || typeof ov !== 'object') return page;
    if (ov.size === 'a4' || ov.size === 'letter') page.size = ov.size;
    if (ov.orientation === 'landscape' || ov.orientation === 'portrait') page.orientation = ov.orientation;
    var lh = Number(ov.lineHeight);
    if (isFinite(lh) && lh >= 0.8 && lh <= 3) page.lineHeight = Math.round(lh * 100) / 100;
    var m = ov.margins || {};
    page.margins = {
      top: clampPageMargin(m.top, page.margins.top),
      right: clampPageMargin(m.right, page.margins.right),
      bottom: clampPageMargin(m.bottom, page.margins.bottom),
      left: clampPageMargin(m.left, page.margins.left)
    };
    return page;
  }
  function documentLanguage(inst) {
    var page = inst ? pageConfig(inst, 1) : null;
    if (page && page.lang) return page.lang;
    return navigator.language || document.documentElement.lang || 'en';
  }
  function pageSizeMm(page) {
    var w = page.size === 'letter' ? 215.9 : 210;
    var h = page.size === 'letter' ? 279.4 : 297;
    return page.orientation === 'landscape' ? { w: h, h: w } : { w: w, h: h };
  }
  function mmToPx(mm) { return Number(mm) * 96 / 25.4; }
  function mmToPt(mm) { return Number(mm) * 72 / 25.4; }
  function mmToTwips(mm) { return Math.round(Number(mm) * 1440 / 25.4); }
  function mmToCm(mm) { return (Number(mm) / 10).toFixed(2).replace(/0+$/, '').replace(/\.$/, '') + 'cm'; }
  function breakInsertionPoint(inst, contentY, excludeNode) {
    if (!inst || !inst.editor) return null;
    var ed = inst.editor;
    var edRect = ed.getBoundingClientRect();
    var scroll = ed.scrollTop;
    var children = Array.prototype.slice.call(ed.children).filter(function (child) {
      return child !== excludeNode && !(child.classList && child.classList.contains('sui-rich-pagebreak'));
    });
    if (!children.length) return null;
    var best = null, bestDist = Infinity, after = false;
    for (var i = 0; i < children.length; i++) {
      var r = children[i].getBoundingClientRect();
      var top = r.top - edRect.top + scroll, bottom = top + r.height;
      var dTop = Math.abs(top - contentY), dBottom = Math.abs(bottom - contentY);
      var dist = Math.min(dTop, dBottom);
      if (dist < bestDist) { bestDist = dist; best = children[i]; after = dBottom < dTop; }
    }
    return best ? { node: best, after: after } : null;
  }
  function movePageBreakTo(inst, node, contentY) {
    var ed = inst.editor;
    var point = breakInsertionPoint(inst, contentY, node);
    if (!point) { ed.appendChild(node); return; }
    var ref = point.after ? point.node.nextSibling : point.node;
    if (ref === node) return;
    ed.insertBefore(node, ref);
  }
  function guideContentHeight(inst) {
    var ed = inst.editor;
    return Math.max(ed.scrollHeight, ed.clientHeight);
  }
  function tableRowSplitAt(inst, contentY) {
    if (!inst || !inst.editor) return null;
    var ed = inst.editor;
    var edRect = ed.getBoundingClientRect(), scroll = ed.scrollTop;
    var children = ed.children;
    for (var ti = 0; ti < children.length; ti++) {
      var table = children[ti];
      if (!table.tagName || table.tagName.toLowerCase() !== 'table') continue;
      var tr = table.getBoundingClientRect();
      var ttop = tr.top - edRect.top + scroll, tbottom = ttop + tr.height;
      if (contentY <= ttop + 1 || contentY >= tbottom - 1) continue;
      var rows = Array.prototype.slice.call(table.querySelectorAll('tr')).filter(function (row) {
        return row.closest('table') === table;
      });
      if (rows.length < 2) return null;
      var headerCount = 0;
      for (var hi = 0; hi < rows.length; hi++) {
        var cells = rows[hi].cells, allH = !!(cells && cells.length);
        for (var ci = 0; ci < cells.length; ci++) if (cells[ci].tagName.toLowerCase() !== 'th') allH = false;
        if (allH) headerCount++; else break;
      }
      var body = rows.slice(headerCount);
      // Keep rows joined by rowspan in the same fragment.
      var groups = [], gi = 0;
      while (gi < body.length) {
        var groupEnd = gi, changed = true;
        while (changed) {
          changed = false;
          for (var gr = gi; gr <= groupEnd && gr < body.length; gr++) {
            var gcells = body[gr].cells || [];
            for (var gc = 0; gc < gcells.length; gc++) {
              var span = Number(gcells[gc].getAttribute('rowspan') || 1);
              if (span > 1 && gr + span - 1 > groupEnd) {
                groupEnd = Math.min(body.length - 1, gr + span - 1);
                changed = true;
              }
            }
          }
        }
        groups.push({ start: gi, end: groupEnd });
        gi = groupEnd + 1;
      }
      var minIdx = headerCount > 0 ? headerCount + 1 : 1;
      var maxIdx = rows.length - 1;
      var best = null, bestDist = Infinity;
      for (var g = 0; g < groups.length; g++) {
        var startIdx = headerCount + groups[g].start;
        var endIdx = headerCount + groups[g].end;
        var candidates = [
          { index: startIdx, y: rows[startIdx].getBoundingClientRect().top - edRect.top + scroll },
          { index: endIdx + 1, y: rows[endIdx].getBoundingClientRect().bottom - edRect.top + scroll }
        ];
        for (var cd = 0; cd < candidates.length; cd++) {
          var cand = candidates[cd];
          if (cand.index < minIdx || cand.index > maxIdx) continue;
          var dist = Math.abs(cand.y - contentY);
          if (dist < bestDist) { bestDist = dist; best = cand; }
        }
      }
      if (best) return { table: table, splitIndex: best.index, y: best.y, headerCount: headerCount, rows: rows.length };
      return null;
    }
    return null;
  }
  function tableHeaderCount(block) {
    var rows = (block && block.rows) || [], count = 0;
    for (var i = 0; i < rows.length; i++) {
      var cells = rows[i].cells || [], allH = cells.length > 0;
      for (var c = 0; c < cells.length; c++) if (!cells[c].header) allH = false;
      if (allH) count++; else break;
    }
    return count;
  }
  function tableColumns(block) {
    var rows = (block && block.rows) || [], max = 0;
    for (var r = 0; r < rows.length; r++) {
      var cells = rows[r].cells || [], n = 0;
      for (var c = 0; c < cells.length; c++) n += Math.max(1, Number(cells[c].colspan) || 1);
      if (n > max) max = n;
    }
    return max;
  }
  function tableMergeCompatible(left, right) {
    if (!left || !right || left.type !== 'table' || right.type !== 'table') return false;
    if (!left._tableGroup || left._tableGroup !== right._tableGroup) return false;
    var leftCols = tableColumns(left), rightCols = tableColumns(right);
    if (!leftCols || leftCols !== rightCols) return false;
    if (tableHeaderCount(left) !== tableHeaderCount(right)) return false;
    var keys = ['border', 'borderWidth', 'padding', 'width', 'height', 'align', 'header'];
    for (var k = 0; k < keys.length; k++) {
      if (String(left[keys[k]] || '') !== String(right[keys[k]] || '')) return false;
    }
    if (JSON.stringify(left.borders || []) !== JSON.stringify(right.borders || [])) return false;
    return true;
  }
  function mergeAdjacentTableFragments(inst) {
    var blocks = (inst && inst.doc && inst.doc.children) || [];
    var changed = false;
    for (var i = 1; i < blocks.length; ) {
      var left = blocks[i - 1], right = blocks[i];
      if (tableMergeCompatible(left, right)) {
        var hc = tableHeaderCount(right);
        left.rows = (left.rows || []).concat((right.rows || []).slice(hc));
        if (!left.caption && right.caption) left.caption = right.caption;
        blocks.splice(i, 1);
        changed = true;
      } else {
        i++;
      }
    }
    if (changed) {
      var counts = {};
      for (var b = 0; b < blocks.length; b++) {
        var g = blocks[b] && blocks[b].type === 'table' ? blocks[b]._tableGroup : '';
        if (g) counts[g] = (counts[g] || 0) + 1;
      }
      for (var t = 0; t < blocks.length; t++) {
        if (blocks[t] && blocks[t].type === 'table' && blocks[t]._tableGroup && counts[blocks[t]._tableGroup] === 1) delete blocks[t]._tableGroup;
      }
    }
    return changed;
  }
  function splitTableBlock(inst, tableNode, splitIndex, removeBreakNode) {
    if (!inst || !inst.editor) return false;
    sync(inst);
    var idx = Array.prototype.indexOf.call(inst.editor.children, tableNode);
    if (idx < 0) return false;
    var blocks = inst.doc.children || [];
    var block = blocks[idx];
    if (!block || block.type !== 'table' || !block.rows) return false;
    var rows = block.rows;
    if (!(splitIndex > 0 && splitIndex < rows.length)) return false;
    var hc = tableHeaderCount(block);
    if (hc > 0 && splitIndex <= hc) splitIndex = Math.min(rows.length - 1, hc + 1);
    if (hc > 0 && splitIndex >= rows.length) splitIndex = rows.length - 1;
    if (splitIndex <= 0 || splitIndex >= rows.length) return false;
    if (rows.length < splitIndex + 1) return false;
    var group = block._tableGroup || ('tbl-' + (++tableGroupSeq) + '-' + Date.now().toString(36));
    var first = JSON.parse(JSON.stringify(block));
    var second = JSON.parse(JSON.stringify(block));
    first._tableGroup = group;
    second._tableGroup = group;
    first.rows = rows.slice(0, splitIndex);
    second.rows = rows.slice(splitIndex);
    if (hc > 0 && splitIndex > hc) second.rows = rows.slice(0, hc).concat(rows.slice(splitIndex));
    if (second.caption) second.caption = '';
    var insertAt = idx;
    blocks.splice(insertAt, 1, first, { type: 'page_break' }, second);
    if (removeBreakNode && removeBreakNode.parentNode) {
      var breakIdx = Array.prototype.indexOf.call(inst.editor.children, removeBreakNode);
      if (breakIdx >= 0) {
        // Remove the old break from the model; splitTableBlock inserted the new
        // break between the two table fragments.
        var modelBreak = breakIdx > insertAt ? breakIdx + 2 : breakIdx;
        if (blocks[modelBreak] && blocks[modelBreak].type === 'page_break') blocks.splice(modelBreak, 1);
      }
    }
    mergeAdjacentTableFragments(inst);
    remember(inst);
    restore(inst, JSON.stringify(inst.doc));
    notify(inst, 'change');
    return true;
  }
  function snapGuideContentY(inst, contentY, excludeNode) {
    if (!inst || !inst.editor) return contentY;
    var tableSplit = tableRowSplitAt(inst, contentY);
    if (tableSplit) return tableSplit.y;
    var ed = inst.editor;
    var edRect = ed.getBoundingClientRect(), scroll = ed.scrollTop;
    var children = ed.children, best = contentY, bestDist = Infinity;
    for (var i = 0; i < children.length; i++) {
      var child = children[i];
      if (child === excludeNode) continue;
      if (child.classList && child.classList.contains('sui-rich-pagebreak')) continue;
      var r = child.getBoundingClientRect();
      var top = r.top - edRect.top + scroll, bottom = top + r.height;
      var dTop = Math.abs(top - contentY), dBottom = Math.abs(bottom - contentY);
      if (dTop < bestDist) { bestDist = dTop; best = top; }
      if (dBottom < bestDist) { bestDist = dBottom; best = bottom; }
    }
    return Math.max(0, Math.min(guideContentHeight(inst), best));
  }
  function breakInsertionBlocked(ed, point) {
    var ref = point ? (point.after ? point.node.nextSibling : point.node) : ed.lastChild;
    if (ref && ref.classList && ref.classList.contains('sui-rich-pagebreak')) return true;
    if (ref && ref.previousSibling && ref.previousSibling.classList && ref.previousSibling.classList.contains('sui-rich-pagebreak')) return true;
    if (!ref && ed.lastChild && ed.lastChild.classList && ed.lastChild.classList.contains('sui-rich-pagebreak')) return true;
    return false;
  }
  function insertPageBreakAt(inst, contentY) {
    if (!inst || !inst.editor) return;
    var tableSplit = tableRowSplitAt(inst, contentY);
    if (tableSplit) { splitTableBlock(inst, tableSplit.table, tableSplit.splitIndex, null); return; }
    var ed = inst.editor;
    var point = breakInsertionPoint(inst, contentY, null);
    // A break already sitting at this boundary must not be duplicated.
    if (breakInsertionBlocked(ed, point)) { schedulePageGuides(inst); return; }
    var br = pageBreakNode();
    movePageBreakTo(inst, br, contentY);
    commit(inst, 'change');
    schedulePageGuides(inst);
  }
  function startPageDrag(inst, line, ev, breakNode) {
    ev.preventDefault();
    ev.stopPropagation();
    // Capture any pending DOM edit before the drag, without notifying the host
    // (a host re-render here would recreate the editor and the dragged line).
    sync(inst);
    if (inst.editor) inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
    var ed = inst.editor;
    var isBreak = !!breakNode;
    var layer = isBreak ? null : line.parentNode;
    var hostRect = layer ? layer.parentNode.getBoundingClientRect() : null;
    var edRect = ed.getBoundingClientRect();
    var startY = ev.clientY, startScroll = ed.scrollTop, startEditorTop = edRect.top;
    var startContent;
    if (isBreak) {
      startContent = breakNode.getBoundingClientRect().top - edRect.top + startScroll;
    } else {
      // The guide stores its document position in data-cut. Do not rebuild it
      // from the current viewport top: a focus/scroll change can make the
      // visual style stale and would map the drag to the start of the document.
      var lineCut = parseFloat(line.getAttribute('data-cut'));
      startContent = isFinite(lineCut) ? lineCut : (parseFloat(line.style.top) || 0) - (edRect.top - hostRect.top) + startScroll;
    }
    var targetContent = startContent;
    var autoDir = 0, autoFrame = null, lastY = ev.clientY, moved = false;
    inst.guideDragging = true;
    line.classList.add('dragging');
    line.__suiDragging = true;
    if (isBreak) { line.style.position = 'relative'; line.style.zIndex = '6'; }
    function render() {
      if (isBreak) {
        // The page break itself follows the pointer: no tiny handle and no
        // ghost line, the red line is the object being moved.
        line.style.transform = 'translateY(' + (targetContent - startContent) + 'px)';
        return;
      }
      var currentRect = ed.getBoundingClientRect();
      line.style.top = ((currentRect.top - hostRect.top) + targetContent - ed.scrollTop) + 'px';
      line.setAttribute('data-cut', String(targetContent));
    }
    function refresh(clientY) {
      // Include both the editor scroll and any page/ancestor scroll: the
      // target must stay on the content under the pointer, not on a viewport
      // pixel that can drift when the page moves.
      var editorShift = ed.getBoundingClientRect().top - startEditorTop;
      targetContent = startContent + (clientY - startY) + (ed.scrollTop - startScroll) - editorShift;
      targetContent = Math.max(0, Math.min(guideContentHeight(inst), targetContent));
      targetContent = snapGuideContentY(inst, targetContent, breakNode);
      if (!isBreak) {
        // An automatic break is already the latest possible cut. Moving it
        // down would only add a redundant break and a mostly blank page, so
        // automatic guides can be dragged earlier (up), never later (down).
        if (targetContent >= startContent - 2) {
          targetContent = startContent;
          moved = false;
        } else {
          moved = true;
        }
      } else if (Math.abs(clientY - startY) > 2 || Math.abs(ed.scrollTop - startScroll) > 2) {
        moved = true;
      }
      render();
    }
    function autoTick() {
      autoFrame = null;
      if (!autoDir) return;
      var before = ed.scrollTop;
      ed.scrollTop = Math.max(0, Math.min(ed.scrollHeight - ed.clientHeight, before + autoDir * 14));
      if (ed.scrollTop !== before) {
        refresh(lastY);
        positionPageGuides(inst);
      }
      autoFrame = window.setTimeout(autoTick, 16);
    }
    function move(moveEv) {
      lastY = moveEv.clientY;
      var currentRect = ed.getBoundingClientRect();
      if (moveEv.clientY < currentRect.top + 28) autoDir = -1;
      else if (moveEv.clientY > currentRect.bottom - 28) autoDir = 1;
      else autoDir = 0;
      refresh(moveEv.clientY);
      if (autoDir && !autoFrame) autoFrame = window.setTimeout(autoTick, 16);
    }
    function resetVisual() {
      line.classList.remove('dragging');
      line.__suiDragging = false;
      if (isBreak) { line.style.transform = ''; line.style.zIndex = ''; line.style.position = ''; }
    }
    function up(upEv) {
      document.removeEventListener('mousemove', move);
      document.removeEventListener('mouseup', up);
      if (autoFrame) { window.clearTimeout(autoFrame); autoFrame = null; }
      resetVisual();
      inst.guideDragging = false;
      if (!moved) { positionPageGuides(inst); return; }
      targetContent = snapGuideContentY(inst, targetContent, breakNode);
      if (isBreak) {
        var targetSplit = tableRowSplitAt(inst, targetContent);
        if (targetSplit && splitTableBlock(inst, targetSplit.table, targetSplit.splitIndex, breakNode)) return;
        movePageBreakTo(inst, breakNode, targetContent);
        sync(inst);
        if (mergeAdjacentTableFragments(inst)) {
          inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
          remember(inst);
          restore(inst, JSON.stringify(inst.doc));
          notify(inst, 'change');
        } else {
          commit(inst, 'change');
        }
        schedulePageGuides(inst);
      } else {
        insertPageBreakAt(inst, targetContent);
      }
    }
    render();
    document.addEventListener('mousemove', move);
    document.addEventListener('mouseup', up);
  }
  function startPageBreakDrag(inst, node, ev) {
    startPageDrag(inst, node, ev, node);
  }
  function startPageGuideDrag(inst, line, ev) {
    startPageDrag(inst, line, ev, null);
  }
  function pageGuideLayer(inst) {
    if (!inst || !inst.editor || !inst.editor.parentNode) return null;
    var host = inst.editor.parentNode;
    var layer = host.querySelector(':scope > .sui-rich-guides');
    if (!layer) {
      layer = el('div', 'sui-rich-guides');
      layer.setAttribute('aria-hidden', 'true');
      host.appendChild(layer);
      if (getComputedStyle(host).position === 'static') host.style.position = 'relative';
    }
    if (!layer.__suiGuideBound) {
      layer.__suiGuideBound = true;
      layer.addEventListener('mousedown', function (ev) {
        var line = ev.target && ev.target.closest ? ev.target.closest('.sui-rich-guide') : null;
        if (line) startPageGuideDrag(inst, line, ev);
      });
    }
    return layer;
  }
  function schedulePageGuides(inst) {
    if (inst.guideTimer) return;
    inst.guideTimer = setTimeout(function () {
      inst.guideTimer = null;
      if (!inst.guideDragging) refreshPageBreaks(inst);
    }, 40);
  }
  async function refreshPageBreaks(inst) {
    if (!inst || !inst.editor) return;
    if (inst.breakBusy) { inst.breakPending = true; return; }
    inst.breakBusy = true;
    do {
      inst.breakPending = false;
      try { await paginate(inst, true); } catch (e) {}
    } while (inst.breakPending);
    inst.breakBusy = false;
    updatePageGuides(inst);
  }
  function positionPageGuides(inst) {
    if (!inst || !inst.editor) return;
    var host = inst.editor.parentNode;
    var layer = host && host.querySelector(':scope > .sui-rich-guides');
    if (!layer) return;
    var ed = inst.editor;
    var edRect = ed.getBoundingClientRect(), hostRect = layer.parentNode.getBoundingClientRect();
    var offset = edRect.top - hostRect.top, scroll = ed.scrollTop;
    for (var i = 0; i < layer.children.length; i++) {
      var line = layer.children[i];
      if (line.__suiDragging) continue;
      var cut = parseFloat(line.getAttribute('data-cut'));
      if (!isFinite(cut)) continue;
      line.style.top = (offset + cut - scroll) + 'px';
    }
  }
  function pageBreakSourceY(inst, source) {
    if (!inst || !inst.editor || !source) return null;
    var ed = inst.editor;
    var node = ed.children[source.block];
    if (!node) return null;
    if (node.classList && node.classList.contains('sui-rich-pagebreak')) return null;
    var edRect = ed.getBoundingClientRect(), scroll = ed.scrollTop;
    var tag = node.tagName ? node.tagName.toLowerCase() : '';
    if (source.row != null && tag === 'table') {
      var rows = Array.prototype.slice.call(node.querySelectorAll('tr')).filter(function (row) {
        return row.closest('table') === node;
      });
      var row = rows[source.row];
      if (row) return row.getBoundingClientRect().top - edRect.top + scroll;
    }
    if (source.offset != null) {
      if (source.kind === 'list') {
        var li = node.children[source.offset];
        if (li) return li.getBoundingClientRect().top - edRect.top + scroll;
      } else {
        var walker = document.createTreeWalker(node, NodeFilter.SHOW_TEXT, null);
        var pos = 0, textNode;
        while ((textNode = walker.nextNode())) {
          var len = (textNode.nodeValue || '').length;
          if (source.offset <= pos + len) {
            var range = document.createRange();
            range.setStart(textNode, Math.max(0, Math.min(len, source.offset - pos)));
            range.collapse(true);
            var rects = range.getClientRects();
            var rect = rects.length ? rects[0] : range.getBoundingClientRect();
            if (rect && (rect.height || rect.top)) return rect.top - edRect.top + scroll;
            break;
          }
          pos += len;
        }
      }
    }
    return node.getBoundingClientRect().top - edRect.top + scroll;
  }
  function updatePageGuides(inst) {
    if (!inst || !inst.editor) return;
    if (inst.guideDragging) { positionPageGuides(inst); return; }
    var layer = pageGuideLayer(inst);
    if (!layer) return;
    if (inst.paginated || (inst.root && inst.root.classList.contains('sui-rich-paginated'))) { layer.innerHTML = ''; return; }
    var ed = inst.editor;
    var m = pageMetrics(inst, 1);
    if (!(m.contentH > 40)) { layer.innerHTML = ''; return; }
    var edRect = ed.getBoundingClientRect(), host = layer.parentNode, hostRect = host.getBoundingClientRect();
    var scroll = ed.scrollTop;
    var sources = inst.pageBreakSources || [];
    var cuts = [];
    for (var i = 0; i < sources.length; i++) {
      if (!sources[i] || sources[i].explicit) continue;
      var y = pageBreakSourceY(inst, sources[i]);
      if (y == null) continue;
      if (!cuts.length || y > cuts[cuts.length - 1] + 2) cuts.push(y);
    }
    layer.innerHTML = '';
    var offset = edRect.top - hostRect.top;
    for (var c = 0; c < cuts.length; c++) {
      var top = offset + cuts[c] - scroll;
      var line = el('div', 'sui-rich-guide');
      line.style.top = top + 'px';
      line.setAttribute('data-cut', String(cuts[c]));
      line.title = 'Automatic page break: drag up to move it earlier (it cannot be removed while the content needs another page)';
      line.setAttribute('data-sui-guide', '1');
      line.appendChild(el('span', 'sui-rich-guide-label', 'Auto break  \u28ff  drag up'));
      layer.appendChild(line);
    }
    positionPageGuides(inst);
  }
  function pageMetrics(inst, pageNumber) {
    var page = pageConfig(inst, pageNumber || 1);
    var size = pageSizeMm(page);
    var width = Math.round(mmToPx(size.w)), height = Math.round(mmToPx(size.h));
    var padTop = Math.round(mmToPx(page.margins.top)), padRight = Math.round(mmToPx(page.margins.right));
    var padBottom = Math.round(mmToPx(page.margins.bottom)), padLeft = Math.round(mmToPx(page.margins.left));
    var headerH = inst.headerH || 0, footerH = inst.footerH || 0;
    var topPad = Math.max(padTop, headerH ? headerH + 10 : 0);
    var bottomPad = Math.max(padBottom, footerH ? footerH + 10 : 0);
    return {
      width: width, height: height,
      padTop: topPad, padRight: padRight, padBottom: bottomPad, padLeft: padLeft,
      headerH: headerH, footerH: footerH,
      lineHeight: page.lineHeight, pageNumber: pageNumber || 1,
      contentW: width - padLeft - padRight,
      contentH: height - topPad - bottomPad
    };
  }
  function setPaginated(inst, on) {
    var rootFlag = inst.root && inst.root.getAttribute('data-sui-rich-paginated') === '1';
    var next = on === undefined ? !(inst.paginated || rootFlag) : !!on;
    inst.paginated = next;
    inst.root.setAttribute('data-sui-rich-paginated', next ? '1' : '0');
    inst.root.classList.toggle('sui-rich-paginated', inst.paginated);
    var btn = inst.root.querySelector('[data-rich-pages]');
    if (btn) { btn.classList.toggle('on', inst.paginated); btn.setAttribute('aria-pressed', inst.paginated ? 'true' : 'false'); btn.textContent = inst.paginated ? 'Edit' : 'Pages'; }
    var wrap = inst.root.querySelector('.sui-rich-pages');
    if (!wrap) {
      wrap = el('div', 'sui-rich-pages');
      inst.root.querySelector('.sui-rich-body').appendChild(wrap);
    }
    wrap.hidden = !inst.paginated;
    updatePageGuides(inst);
    if (inst.paginated) schedulePaginate(inst);
  }
  function schedulePaginate(inst) {
    if (!inst.paginated) return;
    clearTimeout(inst.pageTimer);
    inst.pageTimer = setTimeout(function () { paginate(inst).catch(function (e) { try { console.error('PAGINATE', e && e.stack); inst.root.setAttribute('data-page-error', String(e && e.message)); } catch (x) {} }); }, 140);
  }
  function sheetNode(inst, m) {
    var sheet = el('div', 'sui-rich-page-sheet');
    sheet.setAttribute('lang', documentLanguage(inst));
    sheet.style.width = m.width + 'px';
    sheet.style.minHeight = m.height + 'px';
    sheet.style.padding = m.padTop + 'px ' + m.padRight + 'px ' + m.padBottom + 'px ' + m.padLeft + 'px';
    var head = null, foot = null;
    if (chromeHasContent(inst.headerValue)) {
      head = el('div', 'sui-rich-sheet-header');
      head.style.left = m.padLeft + 'px';
      head.style.right = m.padRight + 'px';
      head.style.top = Math.max(6, m.padTop - m.headerH - 5) + 'px';
      sheet.appendChild(head);
    }
    var body = el('div', 'sui-rich-editor sui-rich-sheet-body');
    body.style.height = m.contentH + 'px';
    body.style.lineHeight = String(m.lineHeight);
    sheet.appendChild(body);
    if (chromeHasContent(inst.footerValue)) {
      foot = el('div', 'sui-rich-sheet-footer');
      foot.style.left = m.padLeft + 'px';
      foot.style.right = m.padRight + 'px';
      foot.style.bottom = Math.max(6, m.padBottom - m.footerH - 5) + 'px';
      sheet.appendChild(foot);
    }
    var number = el('div', 'sui-rich-sheet-number', '1');
    sheet.appendChild(number);
    return { sheet: sheet, body: body, head: head, foot: foot, number: number, used: 0, metrics: m };
  }
  function tableShell(table) {
    var shell = table.cloneNode(false);
    // Keep column definitions and captions; losing colgroup collapses the
    // column widths when a table is split across pages.
    var colgroup = table.querySelector('colgroup');
    if (colgroup) shell.appendChild(colgroup.cloneNode(true));
    var caption = table.querySelector('caption');
    if (caption) shell.appendChild(caption.cloneNode(true));
    var tbody = el('tbody');
    shell.appendChild(tbody);
    return { shell: shell, tbody: tbody };
  }
  async function paginate(inst, recordOnly) {
    inst = resolveInstance(inst.name) || inst;
    if (!recordOnly) {
      if (!inst.paginated && inst.root.getAttribute('data-sui-rich-paginated') === '1') inst.paginated = true;
      if (!inst.paginated) return;
      inst.root.classList.add('sui-rich-paginated');
      inst.root.setAttribute('data-sui-rich-paginated', '1');
    }
    if (recordOnly) {
      sync(inst);
      if (inst.editor) inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
    }
    var wrap;
    if (recordOnly) {
      wrap = el('div', 'sui-rich-pages');
      wrap.style.cssText = 'position:absolute;left:-12000px;top:0;visibility:hidden;pointer-events:none;';
      inst.root.querySelector('.sui-rich-body').appendChild(wrap);
    } else {
      wrap = inst.root.querySelector('.sui-rich-pages');
      if (!wrap) {
        wrap = el('div', 'sui-rich-pages');
        inst.root.querySelector('.sui-rich-body').appendChild(wrap);
      }
      wrap.hidden = false;
    }
    var breakSources = [];
    var m = pageMetrics(inst, 1);
    wrap.innerHTML = '';
    var measure = el('div', 'sui-rich-editor sui-rich-sheet-body sui-rich-measure');
    measure.setAttribute('lang', documentLanguage(inst));
    measure.style.cssText = 'position:absolute;left:-12000px;top:0;width:' + m.contentW + 'px;height:auto;overflow:visible;padding:0;visibility:hidden;line-height:' + m.lineHeight + ';';
    // .sui-rich-sheet-body forces width:auto !important, so the measurement
    // copy would otherwise shrink-to-fit and report a single-line height.
    measure.style.setProperty('width', m.contentW + 'px', 'important');
    inst.root.querySelector('.sui-rich-body').appendChild(measure);
    var blocks = (inst.doc && inst.doc.children) || [];
    var items = [];
    for (var i = 0; i < blocks.length; i++) {
      if (blocks[i].type === 'page_break') { items.push({ block: blocks[i], node: null, index: i }); continue; }
      var node = blockDom(blocks[i], 0);
      if (!node) continue;
      measure.appendChild(node);
      items.push({ block: blocks[i], node: node, index: i });
    }
    var imgs = Array.prototype.slice.call(measure.querySelectorAll('img'));
    await Promise.all(imgs.map(function (img) {
      if (img.complete) return Promise.resolve();
      return new Promise(function (resolve) { img.onload = resolve; img.onerror = resolve; setTimeout(resolve, 800); });
    }));
    function blockHeight(node) {
      var rect = node.getBoundingClientRect();
      var cs = getComputedStyle(node);
      var border = 0;
      if (cs.boxSizing !== 'border-box') {
        border = (parseFloat(cs.borderTopWidth) || 0) + (parseFloat(cs.borderBottomWidth) || 0);
      }
      return rect.height + (parseFloat(cs.marginTop) || 0) + (parseFloat(cs.marginBottom) || 0) + border;
    }
    function measureBlock(node) {
      // Each page can have a different content width and line height (page
      // overrides), so the measurement copy must follow the current page.
      var wantWidth = current.metrics.contentW + 'px';
      if (measure.style.getPropertyValue('width') !== wantWidth) measure.style.setProperty('width', wantWidth, 'important');
      measure.style.lineHeight = String(current.metrics.lineHeight);
      var detached = node.parentNode !== measure;
      if (detached) measure.appendChild(node);
      var h = blockHeight(node);
      if (detached) measure.removeChild(node);
      return h;
    }
    function takeFromMeasure(node) {
      if (node && node.parentNode === measure) measure.removeChild(node);
    }
    function appendBlock(node, height) {
      takeFromMeasure(node);
      current.body.appendChild(node);
      current.used += (height === undefined ? blockHeight(node) : height);
    }
    // Build a copy of a block containing only the text between two offsets.
    // Inline marks survive because the whole subtree is cloned and every leaf
    // text node is trimmed independently.
    function cloneTextSlice(node, from, to) {
      var clone = node.cloneNode(true);
      var walker = document.createTreeWalker(clone, NodeFilter.SHOW_TEXT, null);
      var leaves = [], textNode, pos = 0;
      while ((textNode = walker.nextNode())) {
        var value = textNode.nodeValue || '';
        leaves.push({ node: textNode, start: pos, end: pos + value.length });
        pos += value.length;
      }
      if (!pos || from >= pos) return null;
      to = Math.min(to, pos);
      for (var li = 0; li < leaves.length; li++) {
        var leaf = leaves[li], val = leaf.node.nodeValue || '';
        var a = Math.max(0, from - leaf.start);
        var b = Math.min(val.length, to - leaf.start);
        leaf.node.nodeValue = b <= a ? '' : val.slice(a, b);
      }
      if (from > 0) {
        var w2 = document.createTreeWalker(clone, NodeFilter.SHOW_TEXT, null);
        while ((textNode = w2.nextNode())) {
          if ((textNode.nodeValue || '').length) {
            textNode.nodeValue = textNode.nodeValue.replace(/^\s+/, '');
            break;
          }
        }
      }
      return clone;
    }
    function splitTextBlock(node, maxH) {
      var full = node.textContent || '';
      var total = full.length;
      if (total < 2) return null;
      var best = null, bestLen = 0;
      var lo = 1, hi = total, guard = 0;
      while (lo <= hi && guard++ < 40) {
        var mid = (lo + hi) >> 1;
        var slice = cloneTextSlice(node, 0, mid);
        if (!slice) { hi = mid - 1; continue; }
        if (measureBlock(slice) <= maxH) { best = slice; bestLen = mid; lo = mid + 1; }
        else hi = mid - 1;
      }
      if (!best || bestLen <= 0 || bestLen >= total) return null;
      // Prefer a word boundary when it is close to the measured cut.
      var space = full.lastIndexOf(' ', bestLen);
      if (space > bestLen - 24 && space > bestLen / 2) {
        var adjusted = cloneTextSlice(node, 0, space + 1);
        if (adjusted && measureBlock(adjusted) <= maxH) { best = adjusted; bestLen = space + 1; }
      }
      var rest = cloneTextSlice(node, bestLen, total);
      if (!rest || !(rest.textContent || '').replace(/\s+/g, '')) return null;
      return { head: best, rest: rest, offset: bestLen };
    }
    function splitListBlock(node, maxH) {
      var items = Array.prototype.slice.call(node.children).filter(function (child) {
        return child.tagName && child.tagName.toLowerCase() === 'li';
      });
      if (items.length < 2) return null;
      var best = 0, bestClone = null;
      for (var k = 1; k <= items.length; k++) {
        var clone = node.cloneNode(false);
        for (var i = 0; i < k; i++) clone.appendChild(items[i].cloneNode(true));
        if (measureBlock(clone) <= maxH) { best = k; bestClone = clone; }
        else break;
      }
      if (!best || best >= items.length) return null;
      var rest = node.cloneNode(false);
      for (var j = best; j < items.length; j++) rest.appendChild(items[j].cloneNode(true));
      return { head: bestClone, rest: rest, offset: best, kind: 'list' };
    }
    function splitBlockToFit(node, maxH) {
      var tag = (node.tagName || '').toLowerCase();
      if (tag === 'ul' || tag === 'ol') return splitListBlock(node, maxH);
      if (tag === 'p' || tag === 'h1' || tag === 'h2' || tag === 'h3' || tag === 'h4' || tag === 'blockquote' || tag === 'pre') return splitTextBlock(node, maxH);
      return null;
    }
    function blockCanSplit(type) {
      return type === 'paragraph' || type === 'pre' || type === 'code_block' ||
        type === 'blockquote' || type === 'bullet_list' || type === 'ordered_list';
    }
    function placeTextBlock(node, itemIndex, blockType) {
      var limit = function () { return current.metrics.contentH - 12; };
      var h = measureBlock(node);
      var splittable = blockCanSplit(blockType);
      var baseOffset = 0, baseKind = '';
      if (current.used > 0 && h <= limit() && current.used + h > limit() && !splittable) {
        fresh({ block: itemIndex, offset: baseOffset, kind: baseKind }); current.used = 0;
      }
      if (current.used + h <= limit()) { appendBlock(node, h); return; }
      if (h <= limit() && !splittable) {
        fresh({ block: itemIndex, offset: baseOffset, kind: baseKind });
        current.used = 0;
        appendBlock(node, h);
        return;
      }
      // A block that does not fit the remaining space is split so the page is
      // filled; headings and images stay whole. Each part is measured against
      // the real page so it never enters the margins.
      var remaining = node, guard = 0;
      while (remaining && guard++ < 200) {
        var maxH = limit() - current.used;
        var rh = measureBlock(remaining);
        if (rh <= maxH) { appendBlock(remaining, rh); return; }
        var parts = splitBlockToFit(remaining, maxH);
        if (!parts) {
          if (current.used > 0) {
            fresh({ block: itemIndex, offset: baseOffset, kind: baseKind });
            current.used = 0;
            continue;
          }
          appendBlock(remaining, rh);
          return;
        }
        var headH = measureBlock(parts.head);
        appendBlock(parts.head, headH);
        remaining = parts.rest;
        var nextOffset = baseOffset + (parts.offset || 0);
        var nextKind = parts.kind || '';
        if (!(remaining.textContent || '').replace(/\s+/g, '')) { takeFromMeasure(remaining); return; }
        fresh({ block: itemIndex, offset: nextOffset, kind: nextKind });
        current.used = 0;
        baseOffset = nextOffset;
        baseKind = nextKind;
      }
    }
    var sheets = [];
    var current = sheetNode(inst, m);
    sheets.push(current); wrap.appendChild(current.sheet);
    function fresh(source) {
      var nextNo = sheets.length + 1;
      current = sheetNode(inst, pageMetrics(inst, nextNo));
      sheets.push(current); wrap.appendChild(current.sheet); current.used = 0;
      if (source) breakSources.push(source);
    }
    function placeTable(node, height, itemIndex) {
      var rows = Array.prototype.slice.call(node.querySelectorAll('tr'));
      var tcs = getComputedStyle(node);
      var tableEdge = (parseFloat(tcs.borderTopWidth) || 0) + (parseFloat(tcs.borderBottomWidth) || 0);
      var tableMarginTop = parseFloat(tcs.marginTop) || 0, tableMarginBottom = parseFloat(tcs.marginBottom) || 0;
      if (rows.length <= 1) { fresh({ block: itemIndex }); current.body.appendChild(node); current.used += height; return; }
      var headerCount = 0;
      for (var hr = 0; hr < rows.length; hr++) {
        var cellsH = rows[hr].cells;
        var allH = !!(cellsH && cellsH.length);
        for (var hc = 0; hc < cellsH.length; hc++) if (cellsH[hc].tagName.toLowerCase() !== 'th') allH = false;
        if (allH) headerCount++; else break;
      }
      var bodyRows = rows.slice(headerCount);
      var groups = [], gi = 0;
      while (gi < bodyRows.length) {
        var groupEnd = gi, changed = true;
        while (changed) {
          changed = false;
          for (var gr = gi; gr <= groupEnd && gr < bodyRows.length; gr++) {
            var gcells = bodyRows[gr].cells || [];
            for (var gc = 0; gc < gcells.length; gc++) {
              var span = Number(gcells[gc].getAttribute('rowspan') || 1);
              if (span > 1 && gr + span - 1 > groupEnd) {
                groupEnd = Math.min(bodyRows.length - 1, gr + span - 1);
                changed = true;
              }
            }
          }
        }
        groups.push({ start: gi, end: groupEnd });
        gi = groupEnd + 1;
      }
      current.used += tableMarginTop;
      var shell = null;
      function newPart() {
        shell = tableShell(node);
        current.body.appendChild(shell.shell);
        for (var hh = 0; hh < headerCount; hh++) shell.tbody.appendChild(rows[hh].cloneNode(true));
      }
      function partHeight() { return shell.shell.getBoundingClientRect().height; }
      function bodyCount() { return shell.tbody.children.length - headerCount; }
      function removeRows(n) { for (var rr = 0; rr < n; rr++) if (shell.tbody.lastChild) shell.tbody.removeChild(shell.tbody.lastChild); }
      function finishPart() {
        if (!shell) return false;
        if (!bodyCount()) { if (shell.shell.parentNode) shell.shell.parentNode.removeChild(shell.shell); shell = null; return false; }
        var h = partHeight();
        current.used += h + tableEdge;
        shell = null;
        return true;
      }
      var pending = [];
      function insertPending(tr) {
        if (!pending.length) return;
        pending.sort(function (a, b) { return a.col - b.col; });
        for (var p = 0; p < pending.length; p++) {
          var cell = pending[p].node.cloneNode(true);
          var ref = tr.cells[pending[p].col] || null;
          tr.insertBefore(cell, ref);
        }
        pending = [];
      }
      function appendRow(index) {
        var tr = bodyRows[index].cloneNode(true);
        if (shell.tbody.children.length === headerCount) insertPending(tr);
        shell.tbody.appendChild(tr);
        return tr;
      }
      function adjustPartRowspans() {
        var body = Array.prototype.slice.call(shell.tbody.children).slice(headerCount);
        var total = body.length, next = [];
        for (var ri = 0; ri < total; ri++) {
          var cells = body[ri].cells || [];
          for (var ci = 0; ci < cells.length; ci++) {
            var cell = cells[ci];
            var span = Number(cell.getAttribute('rowspan') || 1);
            if (span > 1 && ri + span > total) {
              var remaining = span - (total - ri);
              cell.setAttribute('rowspan', String(total - ri));
              var clone = cell.cloneNode(false);
              clone.setAttribute('rowspan', String(remaining));
              next.push({ col: ci, node: clone });
            }
          }
        }
        return next;
      }
      newPart();
      var limitForPage = function () { return current.metrics.contentH - 16 - tableEdge; };
      for (var g = 0; g < groups.length; g++) {
        var group = groups[g];
        var sr = group.start;
        var guard = 0;
        while (sr <= group.end && guard++ < 200) {
          var added = 0;
          for (; sr <= group.end; sr++) {
            var tr = appendRow(sr);
            var h = partHeight();
            if (current.used + h > limitForPage()) {
              shell.tbody.removeChild(tr);
              // A row taller than a whole page cannot fit anywhere; place it on
              // an empty page so the loop can make progress (clipped by print).
              if (added === 0 && bodyCount() === 0) {
                shell.tbody.appendChild(tr);
                added = 1;
                sr++;
              }
              break;
            }
            added++;
          }
          if (added === 0) {
            finishPart();
            fresh({ block: itemIndex, row: headerCount + sr });
            current.used = 0;
            newPart();
            continue;
          }
          var continuations = adjustPartRowspans();
          if (sr <= group.end) {
            // The part is full inside a rowspan group: keep the remaining
            // merged rows in the next part and continue the merge there.
            pending = continuations;
            finishPart();
            fresh({ block: itemIndex, row: headerCount + sr });
            current.used = 0;
            newPart();
          } else {
            if (continuations.length) pending = continuations;
            break;
          }
        }
      }
      finishPart();
      current.used += tableMarginBottom;
    }
    for (var it = 0; it < items.length; it++) {
      var item = items[it];
      if (!item.node) { if (current.used > 0) fresh({ block: it + 1, explicit: true }); continue; }
      var h = measureBlock(item.node);
      if (item.block.type === 'table') {
        if (h > current.metrics.contentH - 12 || (current.used > 0 && current.used + h > current.metrics.contentH - 12)) {
          placeTable(item.node, h, item.index);
          continue;
        }
        appendBlock(item.node, h);
        continue;
      }
      placeTextBlock(item.node, item.index, item.block.type);
    }
    if (measure.parentNode) measure.parentNode.removeChild(measure);
    inst.pageBreakSources = breakSources;
    if (recordOnly) {
      if (wrap.parentNode) wrap.parentNode.removeChild(wrap);
      return;
    }
    var total = sheets.length;
    for (var s2 = 0; s2 < sheets.length; s2++) {
      var sheet = sheets[s2];
      sheet.number.textContent = String(s2 + 1);
      if (sheet.head) renderChromeInto(sheet.head, inst, inst.headerValue, s2 + 1, total);
      if (sheet.foot) renderChromeInto(sheet.foot, inst, inst.footerValue, s2 + 1, total);
      sheet.body.style.height = sheet.metrics.contentH + 'px';
    }
    var bodyBox = inst.root.querySelector('.sui-rich-body');
    var available = Math.max(320, bodyBox ? bodyBox.clientWidth - 36 : 760);
    wrap.style.zoom = Math.min(1, available / m.width);
  }

  function buildStatus(inst) {
    var status = inst.root.querySelector('.sui-rich-status');
    if (!status) return;
    status.setAttribute('role', 'status');
    status.innerHTML = '';
    var stats = el('span', 'sui-rich-stats'); status.appendChild(stats);
    var pos = el('span', 'sui-rich-position'); status.appendChild(pos);
    var saved = el('span', 'sui-rich-saved', 'Ready'); status.appendChild(saved);
    var pagesBtn = el('button', 'sui-rich-zoom-btn sui-rich-pages-btn', 'Pages');
    pagesBtn.type = 'button';
    pagesBtn.setAttribute('data-rich-pages', '1');
    pagesBtn.title = 'Paginated preview (A4 pages)';
    pagesBtn.setAttribute('aria-pressed', 'false');
    status.appendChild(pagesBtn);
    // Zoom lives in the toolbar (see the "zoom" token); the status bar keeps
    // only the paginated-view toggle. 100% reset is the label button itself.
  }

  function findTextNodes(inst, query) {
    var matches = [];
    if (!query) return matches;
    var walker = document.createTreeWalker(inst.editor, NodeFilter.SHOW_TEXT, {
      acceptNode: function (node) {
        var parent = node.parentNode;
        if (!parent) return NodeFilter.FILTER_REJECT;
        if (parent.closest && parent.closest('[contenteditable="false"], .sui-rich-pagebreak, .sui-rich-findbar')) return NodeFilter.FILTER_REJECT;
        return String(node.nodeValue || '').length ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_REJECT;
      }
    });
    var needle = String(query).toLowerCase(), node;
    while ((node = walker.nextNode())) {
      var text = String(node.nodeValue || '');
      var normalized = '', indexMap = [];
      for (var ch = 0; ch < text.length; ch++) {
        if (text.charAt(ch) === '\u00AD') continue;
        normalized += text.charAt(ch);
        indexMap.push(ch);
      }
      var hay = normalized.toLowerCase();
      var from = 0, at;
      while ((at = hay.indexOf(needle, from)) >= 0) {
        var start = indexMap[at];
        var end = indexMap[at + needle.length - 1] + 1;
        matches.push({ node: node, start: start, end: end });
        from = at + needle.length;
      }
    }
    return matches;
  }
  function findRange(match) {
    var range = document.createRange();
    range.setStart(match.node, match.start);
    range.setEnd(match.node, match.end);
    return range;
  }
  function findCountText(inst) {
    if (!inst.find) return '0/0';
    return (inst.find.index >= 0 ? inst.find.index + 1 : 0) + '/' + inst.find.matches.length;
  }
  function updateFindBar(inst) {
    var bar = inst.root.querySelector('.sui-rich-findbar');
    if (!bar) return;
    var input = bar.querySelector('[data-rich-find="query"]');
    var repl = bar.querySelector('[data-rich-find="replacement"]');
    if (input && input.value !== (inst.find ? inst.find.query : '')) input.value = inst.find ? inst.find.query : '';
    var count = bar.querySelector('.sui-rich-find-count');
    if (count) count.textContent = findCountText(inst);
    var scope = inst.find && inst.find.matches.length ? (inst.find.index + 1) + ' of ' + inst.find.matches.length : 'no matches';
    if (repl) repl.setAttribute('aria-label', 'Replace with');
    bar.setAttribute('data-find-status', scope);
  }
  function paintFind(inst) {
    try {
      if (!window.Highlight || !window.CSS || !CSS.highlights) return;
      var normal = new Highlight(), current = new Highlight();
      for (var i = 0; i < inst.find.matches.length; i++) {
        var range = findRange(inst.find.matches[i]);
        if (i === inst.find.index) current.add(range); else normal.add(range);
      }
      CSS.highlights.set('sui-rich-find', normal);
      CSS.highlights.set('sui-rich-find-current', current);
    } catch (e) {}
  }
  function focusFindMatch(inst) {
    if (!inst.find || inst.find.index < 0) return;
    var match = inst.find.matches[inst.find.index];
    if (!match) return;
    var range = findRange(match);
    var sel = window.getSelection();
    if (sel) { sel.removeAllRanges(); sel.addRange(range); }
    var element = match.node.parentNode;
    if (element && element.scrollIntoView) {
      var editor = inst.editor;
      var rect = element.getBoundingClientRect();
      var box = editor.getBoundingClientRect();
      if (rect.top < box.top || rect.bottom > box.bottom) {
        try { element.scrollIntoView({ block: 'center', behavior: 'auto' }); } catch (e) { editor.scrollTop += rect.top - box.top - 20; }
      }
    }
  }
  function openFindBar(inst, query) {
    var bar = inst.root.querySelector('.sui-rich-findbar');
    if (!bar) return;
    bar.hidden = false;
    var input = bar.querySelector('[data-rich-find="query"]');
    if (input) { if (query) input.value = query; input.focus(); input.select(); }
  }
  function closeFindBar(inst) {
    var bar = inst.root.querySelector('.sui-rich-findbar');
    if (bar) bar.hidden = true;
    try { if (window.CSS && CSS.highlights) { CSS.highlights.delete('sui-rich-find'); CSS.highlights.delete('sui-rich-find-current'); } } catch (e) {}
    if (inst.find) { inst.find.matches = []; inst.find.index = -1; }
  }
  function findInDocument(inst, query) {
    inst.find = { query: String(query || ''), matches: [], index: -1 };
    inst.find.matches = findTextNodes(inst, inst.find.query);
    inst.find.index = inst.find.matches.length ? 0 : -1;
    paintFind(inst);
    updateFindBar(inst);
    focusFindMatch(inst);
    return inst.find.matches.length;
  }
  function nextMatch(inst, direction) {
    if (!inst.find || !inst.find.matches.length) return;
    var total = inst.find.matches.length;
    inst.find.index = (inst.find.index + (direction < 0 ? total - 1 : 1)) % total;
    paintFind(inst);
    updateFindBar(inst);
    focusFindMatch(inst);
  }
  function replaceMatch(inst, replacement) {
    if (!inst.find || inst.find.index < 0 || !inst.find.matches.length) return false;
    var match = inst.find.matches[inst.find.index];
    var text = replacement == null ? '' : String(replacement);
    var value = String(match.node.nodeValue || '');
    var nextValue = value.slice(0, match.start) + text + value.slice(match.end);
    match.node.nodeValue = nextValue;
    var caretNode = match.node;
    var caretOffset = match.start + text.length;
    var query = inst.find.query;
    var keep = inst.find.index;
    inst.find.matches = findTextNodes(inst, query);
    if (!inst.find.matches.length) inst.find.index = -1;
    else inst.find.index = Math.min(keep, inst.find.matches.length - 1);
    paintFind(inst);
    updateFindBar(inst);
    var sel = window.getSelection();
    if (sel && caretNode.parentNode) {
      var range = document.createRange();
      range.setStart(caretNode, Math.min(caretOffset, caretNode.nodeValue.length));
      range.collapse(true);
      sel.removeAllRanges();
      sel.addRange(range);
    }
    commit(inst, 'change');
    return true;
  }
  function replaceAllMatches(inst, replacement) {
    if (!inst.find || !inst.find.query) return 0;
    var matches = findTextNodes(inst, inst.find.query);
    var text = replacement == null ? '' : String(replacement);
    var touched = {};
    for (var i = matches.length - 1; i >= 0; i--) {
      var match = matches[i];
      var value = String(match.node.nodeValue || '');
      match.node.nodeValue = value.slice(0, match.start) + text + value.slice(match.end);
      touched[match.node] = true;
    }
    inst.find.matches = findTextNodes(inst, inst.find.query);
    inst.find.index = inst.find.matches.length ? 0 : -1;
    paintFind(inst);
    updateFindBar(inst);
    commit(inst, 'change');
    return matches.length;
  }
  function buildPageChrome(inst) {
    if (!chromeHasContent(inst.headerValue) && !chromeHasContent(inst.footerValue)) return;
    var body = inst.root.querySelector('.sui-rich-body');
    if (!body || body.querySelector('.sui-rich-page')) return;
    var page = el('div', 'sui-rich-page');
    body.insertBefore(page, inst.editor);
    if (chromeHasContent(inst.headerValue)) {
      var head = el('div', 'sui-rich-pageheader');
      head.setAttribute('contenteditable', 'false');
      page.appendChild(head);
      renderChromeInto(head, inst, inst.headerValue, 1, 1);
    }
    page.appendChild(inst.editor);
    if (chromeHasContent(inst.footerValue)) {
      var foot = el('div', 'sui-rich-pagefooter');
      foot.setAttribute('contenteditable', 'false');
      page.appendChild(foot);
      renderChromeInto(foot, inst, inst.footerValue, 1, 1);
    }
  }
  function pageMarginInput(parent, side, label, prefix) {
    prefix = prefix || 'page';
    var row = el('label');
    row.appendChild(el('span', '', label));
    var input = el('input');
    input.type = 'number';
    input.min = '0';
    input.max = '50';
    input.step = '1';
    input.setAttribute('data-rich-' + prefix + '-margin', side);
    input.setAttribute('aria-label', label + ' margin (mm)');
    input.title = label + ' margin in millimetres (0-50)';
    row.appendChild(input);
    parent.appendChild(row);
    return input;
  }
  function currentPageNumber(inst) {
    if (!inst || !inst.editor) return 1;
    var block = currentBlock(inst);
    if (!block) return 1;
    var ed = inst.editor;
    var r = block.getBoundingClientRect(), er = ed.getBoundingClientRect();
    var y = r.top - er.top + ed.scrollTop;
    var cs = getComputedStyle(ed);
    var start = parseFloat(cs.paddingTop) || 0, used = 0, page = 1;
    var m = pageMetrics(inst, page);
    var children = ed.children;
    for (var i = 0; i < children.length; i++) {
      var child = children[i];
      var cr = child.getBoundingClientRect();
      var ccs = getComputedStyle(child);
      var h = cr.height + (parseFloat(ccs.marginTop) || 0) + (parseFloat(ccs.marginBottom) || 0);
      if (child.classList && child.classList.contains('sui-rich-pagebreak')) {
        var bottom = cr.top - er.top + ed.scrollTop + cr.height;
        if (y < bottom) return page;
        page++;
        start = bottom;
        used = 0;
        m = pageMetrics(inst, page);
        continue;
      }
      if (y < start + used + h) return page;
      used += h;
      if (used > m.contentH) {
        start += m.contentH;
        used -= m.contentH;
        page++;
        m = pageMetrics(inst, page);
      }
    }
    return page;
  }
  function pageSettingsPanel(inst) {
    var panel = document.querySelector('.sui-rich-pagesettings[data-rich-pagesettings="' + inst.name + '"]');
    if (panel) return panel;
    panel = el('div', 'sui-rich-pagesettings');
    panel.setAttribute('data-rich-pagesettings', inst.name);
    panel.hidden = true;
    panel.appendChild(panelTitle('Page settings', 'Overrides Document settings for this page. Reset restores the document defaults. Margins accept 0-50 mm.'));

    var pageRow = el('label');
    pageRow.appendChild(el('span', '', 'Page'));
    var pageNumber = el('input');
    pageNumber.type = 'number';
    pageNumber.min = '1';
    pageNumber.max = '999';
    pageNumber.step = '1';
    pageNumber.setAttribute('data-rich-page-number', '1');
    pageNumber.setAttribute('aria-label', 'Page number');
    pageNumber.title = 'Page number for this override';
    pageRow.appendChild(pageNumber);
    panel.appendChild(pageRow);

    var sizeRow = el('label');
    sizeRow.appendChild(el('span', '', 'Size'));
    var sizeSelect = el('select');
    sizeSelect.setAttribute('data-rich-page-size', '1');
    sizeSelect.setAttribute('aria-label', 'Page size');
    [['a4', 'A4'], ['letter', 'Letter']].forEach(function (o) { var op = el('option', '', o[1]); op.value = o[0]; sizeSelect.appendChild(op); });
    sizeRow.appendChild(sizeSelect);
    panel.appendChild(sizeRow);

    var orientRow = el('label');
    orientRow.appendChild(el('span', '', 'Orientation'));
    var orientSelect = el('select');
    orientSelect.setAttribute('data-rich-page-orientation', '1');
    orientSelect.setAttribute('aria-label', 'Page orientation');
    [['portrait', 'Vertical'], ['landscape', 'Horizontal']].forEach(function (o) { var op = el('option', '', o[1]); op.value = o[0]; orientSelect.appendChild(op); });
    orientRow.appendChild(orientSelect);
    panel.appendChild(orientRow);

    panel.appendChild(el('div', 'sui-rich-tableprops-title', 'Margins (mm)'));
    var grid = el('div', 'sui-rich-pagesettings-grid');
    panel.appendChild(grid);
    pageMarginInput(grid, 'top', 'Top');
    pageMarginInput(grid, 'right', 'Right');
    pageMarginInput(grid, 'bottom', 'Bottom');
    pageMarginInput(grid, 'left', 'Left');

    chromeOpenButton(inst, panel);

    var actions = el('div', 'sui-rich-pagesettings-actions');
    var apply = el('button', 'sui-rich-btn', 'Apply');
    apply.type = 'button';
    apply.setAttribute('data-rich-page-apply', '1');
    var reset = el('button', 'sui-rich-btn', 'Reset');
    reset.type = 'button';
    reset.setAttribute('data-rich-page-reset', '1');
    var close = el('button', 'sui-rich-btn', 'Close');
    close.type = 'button';
    close.setAttribute('data-rich-page-close', '1');
    actions.appendChild(apply); actions.appendChild(reset); actions.appendChild(close);
    panel.appendChild(actions);

    panel.addEventListener('click', function (ev) {
      var target = ev.target;
      if (!target || !target.closest) return;
      if (target.closest('[data-rich-edit-chrome]')) { openChromeEditor(inst, target.closest('[data-rich-edit-chrome]').getAttribute('data-rich-edit-chrome')); return; }
      if (target.closest('[data-rich-open-chrome]')) { openChromeEditor(inst, 'header'); return; }
      if (target.closest('[data-rich-chrome-help]')) { openChromeHelp(inst); return; }
      if (target.closest('[data-rich-page-apply]')) { applyPageSettings(inst); return; }
      if (target.closest('[data-rich-page-reset]')) { resetPageOverride(inst, panel); return; }
      if (target.closest('[data-rich-page-close]')) { panel.hidden = true; return; }
    });
    panel.addEventListener('change', function (ev) {
      if (!ev.target || ev.target.getAttribute('data-rich-page-number') === null) return;
      var n = Math.max(1, Math.min(999, parseInt(ev.target.value, 10) || 1));
      ev.target.value = String(n);
      fillPageSettings(inst, panel, pageConfig(inst, n), n);
    });
    document.body.appendChild(panel);
    if (!window.__suiPageSettingsBound) {
      window.__suiPageSettingsBound = true;
      document.addEventListener('mousedown', function (ev) {
        var target = ev.target;
        if (target && target.closest && (target.closest('.sui-rich-pagesettings') || target.closest('[data-rich-cmd="pageSettings"]'))) return;
        var panels = document.querySelectorAll('.sui-rich-pagesettings');
        for (var i = 0; i < panels.length; i++) panels[i].hidden = true;
      });
    }
    return panel;
  }
  function fillPageSettings(inst, panel, page, pageNumber) {
    if (pageNumber) {
      var pageInput = panel.querySelector('[data-rich-page-number]');
      if (pageInput) pageInput.value = String(pageNumber);
    }
    var size = panel.querySelector('[data-rich-page-size]');
    var orient = panel.querySelector('[data-rich-page-orientation]');
    if (size) size.value = page.size;
    if (orient) orient.value = page.orientation;
    ['top', 'right', 'bottom', 'left'].forEach(function (side) {
      var input = panel.querySelector('[data-rich-page-margin="' + side + '"]');
      if (input) input.value = String(page.margins[side]);
    });
  }
  function applyPageLineHeight(inst) {
    var lh = String(pageConfig(inst).lineHeight);
    if (inst.editor) inst.editor.style.lineHeight = lh;
    var nodes = inst.root ? inst.root.querySelectorAll('.sui-rich-sheet-body, .sui-rich-measure') : [];
    for (var i = 0; i < nodes.length; i++) nodes[i].style.lineHeight = lh;
  }
  function applyPageSettings(inst) {
    inst = resolveInstance(inst.name) || inst;
    var panel = pageSettingsPanel(inst);
    function value(sel) { var node = panel.querySelector(sel); return node ? String(node.value || '') : ''; }
    function margin(side) {
      var node = panel.querySelector('[data-rich-page-margin="' + side + '"]');
      return node ? clampPageMargin(node.value, defaultPageConfig().margins[side]) : defaultPageConfig().margins[side];
    }
    var n = Math.max(1, Math.min(999, parseInt(value('[data-rich-page-number]'), 10) || 1));
    var defaults = documentPageConfig(inst);
    var override = {};
    var size = value('[data-rich-page-size]') === 'letter' ? 'letter' : 'a4';
    var orientation = value('[data-rich-page-orientation]') === 'landscape' ? 'landscape' : 'portrait';
    if (size !== defaults.size) override.size = size;
    if (orientation !== defaults.orientation) override.orientation = orientation;
    var margins = {};
    ['top', 'right', 'bottom', 'left'].forEach(function (side) {
      var v = margin(side);
      if (v !== defaults.margins[side]) margins[side] = v;
    });
    if (Object.keys(margins).length) override.margins = margins;
    var overrides = inst.doc.pageOverrides || {};
    if (Object.keys(override).length) overrides[String(n)] = override;
    else delete overrides[String(n)];
    inst.doc.pageOverrides = overrides;
    updatePageGuides(inst);
    if (inst.paginated) schedulePaginate(inst);
    commit(inst, 'change');
    panel.hidden = true;
  }
  function resetPageOverride(inst, panel) {
    inst = resolveInstance(inst.name) || inst;
    var input = panel.querySelector('[data-rich-page-number]');
    var n = Math.max(1, Math.min(999, parseInt(input ? input.value : '1', 10) || 1));
    var overrides = inst.doc.pageOverrides || {};
    delete overrides[String(n)];
    inst.doc.pageOverrides = overrides;
    fillPageSettings(inst, panel, documentPageConfig(inst), n);
    updatePageGuides(inst);
    if (inst.paginated) schedulePaginate(inst);
    commit(inst, 'change');
  }
  function openPageSettings(inst) {
    var liveRoot = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"]');
    if (liveRoot) {
      inst.root = liveRoot;
      var liveEditor = liveRoot.querySelector('.sui-rich-editor');
      if (liveEditor) inst.editor = liveEditor;
    }
    var panel = pageSettingsPanel(inst);
    // The panel lives on <body>, outside the .sui-rich scope: copy the theme
    // custom properties so it matches the editor and stays opaque.
    var rootStyle = getComputedStyle(inst.root);
    ['--sui-rich-bg', '--sui-rich-fg', '--sui-rich-border', '--sui-rich-muted', '--sui-rich-hover', '--sui-rich-accent', '--sui-rich-shadow'].forEach(function (name) {
      var value = rootStyle.getPropertyValue(name);
      if (value) panel.style.setProperty(name, value);
    });
    var pageNo = currentPageNumber(inst);
    fillPageSettings(inst, panel, pageConfig(inst, pageNo), pageNo);
    panel.hidden = false;
    var trigger = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"] .sui-rich-toolbar [data-rich-cmd="pageSettings"]');
    var rect = trigger ? trigger.getBoundingClientRect() : inst.root.getBoundingClientRect();
    var panelW = 300, panelH = 400;
    var left = rect && rect.width ? rect.left : 16;
    var top = rect && rect.height ? rect.bottom + 6 : 16;
    left = Math.max(8, Math.min(window.innerWidth - panelW, left));
    if (top + panelH > window.innerHeight) top = Math.max(8, window.innerHeight - panelH);
    panel.style.left = left + 'px';
    panel.style.top = top + 'px';
    var first = panel.querySelector('[data-rich-page-size]');
    if (first) first.focus();
  }
  function paragraphSettingsPanel(inst) {
    var panel = document.querySelector('.sui-rich-paragraphsettings[data-rich-paragraphsettings="' + inst.name + '"]');
    if (panel) return panel;
    panel = el('div', 'sui-rich-paragraphsettings');
    panel.setAttribute('data-rich-paragraphsettings', inst.name);
    panel.hidden = true;
    panel.appendChild(panelTitle('Paragraph settings', 'Applies to the selection, or to the paragraph at the caret.'));

    var alignRow = el('label');
    alignRow.appendChild(el('span', '', 'Alignment'));
    var align = el('select');
    align.setAttribute('data-rich-paragraph-align', '1');
    align.setAttribute('aria-label', 'Paragraph alignment');
    [['', 'Default'], ['left', 'Left'], ['center', 'Center'], ['right', 'Right'], ['justify', 'Justify'], ['justifyHyphen', 'Justify + hyphenation']].forEach(function (o) { var op = el('option', '', o[1]); op.value = o[0]; align.appendChild(op); });
    alignRow.appendChild(align);
    panel.appendChild(alignRow);

    function numberRow(label, key, min, max, step, unit) {
      var row = el('label');
      row.appendChild(el('span', '', label));
      var input = el('input');
      input.type = 'number';
      input.min = String(min);
      input.max = String(max);
      input.step = String(step);
      input.setAttribute('data-rich-paragraph-' + key, '1');
      input.setAttribute('aria-label', label + (unit ? ' (' + unit + ')' : ''));
      input.title = label + (unit ? ' in ' + unit : '');
      row.appendChild(input);
      panel.appendChild(row);
      return input;
    }
    var indentInput = numberRow('Indent left', 'indent', 0, 400, 1, 'px');
    var indentActions = el('div', 'sui-rich-indent-actions');
    var indentLess = el('button', 'sui-rich-btn', '\u2212 Indent'); indentLess.type = 'button'; indentLess.setAttribute('data-rich-paragraph-indent-dec', '1');
    var indentMore = el('button', 'sui-rich-btn', '+ Indent'); indentMore.type = 'button'; indentMore.setAttribute('data-rich-paragraph-indent-inc', '1');
    indentActions.appendChild(indentLess); indentActions.appendChild(indentMore);
    panel.appendChild(indentActions);
    numberRow('Line height', 'lineheight', 0.8, 3, 0.05, '');
    numberRow('Space before', 'before', 0, 200, 1, 'px');
    numberRow('Space after', 'after', 0, 200, 1, 'px');
    var tabRow = el('label');
    tabRow.appendChild(el('span', '', 'Code tab size'));
    var tabSelect = el('select');
    tabSelect.setAttribute('data-rich-paragraph-tabsize', '1');
    tabSelect.setAttribute('aria-label', 'Code indentation width');
    [['2', '2 spaces'], ['4', '4 spaces'], ['8', '8 spaces']].forEach(function (o) {
      var op = el('option', '', o[1]); op.value = o[0]; tabSelect.appendChild(op);
    });
    tabRow.appendChild(tabSelect);
    panel.appendChild(tabRow);
    var lineNumRow = el('label', 'sui-rich-paragraph-check');
    lineNumRow.appendChild(el('span', '', 'Line numbers (code)'));
    var lineNum = el('input');
    lineNum.type = 'checkbox';
    lineNum.setAttribute('data-rich-paragraph-linenumbers', '1');
    lineNum.setAttribute('aria-label', 'Show line numbers in code blocks');
    lineNum.title = 'Show a line-number gutter in code blocks.';
    lineNumRow.appendChild(lineNum);
    panel.appendChild(lineNumRow);
    var codeLangRow = el('label');
    codeLangRow.appendChild(el('span', '', 'Code language'));
    var codeLangSelect = el('select');
    codeLangSelect.setAttribute('data-rich-paragraph-codelang', '1');
    codeLangSelect.setAttribute('aria-label', 'Code syntax highlighting language');
    CODE_LANGUAGES.forEach(function (pair) {
      var op = el('option', '', pair[1]); op.value = pair[0]; codeLangSelect.appendChild(op);
    });
    codeLangRow.appendChild(codeLangSelect);
    panel.appendChild(codeLangRow);
    var hyphenRow = el('label', 'sui-rich-paragraph-check');
    hyphenRow.appendChild(el('span', '', 'Hyphenation'));
    var hyphen = el('input');
    hyphen.type = 'checkbox';
    hyphen.setAttribute('data-rich-paragraph-hyphens', '1');
    hyphen.setAttribute('aria-label', 'Enable automatic hyphenation');
    hyphen.title = 'Break long words with hyphens at the line end.';
    hyphenRow.appendChild(hyphen);
    panel.appendChild(hyphenRow);
    align.addEventListener('change', function () {
      if (align.value === 'justify') hyphen.checked = false;
      else if (align.value === 'justifyHyphen') hyphen.checked = true;
    });

    var actions = el('div', 'sui-rich-pagesettings-actions');
    var apply = el('button', 'sui-rich-btn', 'Apply'); apply.type = 'button'; apply.setAttribute('data-rich-paragraph-apply', '1');
    var reset = el('button', 'sui-rich-btn', 'Reset'); reset.type = 'button'; reset.setAttribute('data-rich-paragraph-reset', '1');
    var close = el('button', 'sui-rich-btn', 'Close'); close.type = 'button'; close.setAttribute('data-rich-paragraph-close', '1');
    actions.appendChild(apply); actions.appendChild(reset); actions.appendChild(close);
    panel.appendChild(actions);

    panel.addEventListener('click', function (ev) {
      var target = ev.target;
      if (!target || !target.closest) return;
      if (target.closest('[data-rich-paragraph-apply]')) { applyParagraphSettings(inst); return; }
      if (target.closest('[data-rich-paragraph-indent-dec]')) { changeParagraphIndent(inst, -24); return; }
      if (target.closest('[data-rich-paragraph-indent-inc]')) { changeParagraphIndent(inst, 24); return; }
      if (target.closest('[data-rich-paragraph-reset]')) { fillParagraphSettings(inst, panel, null); return; }
      if (target.closest('[data-rich-paragraph-close]')) { panel.hidden = true; return; }
    });
    document.body.appendChild(panel);
    if (!window.__suiParagraphSettingsBound) {
      window.__suiParagraphSettingsBound = true;
      document.addEventListener('mousedown', function (ev) {
        var target = ev.target;
        if (target && target.closest && (target.closest('.sui-rich-paragraphsettings') || target.closest('[data-rich-cmd="paragraphSettings"]'))) return;
        var panels = document.querySelectorAll('.sui-rich-paragraphsettings');
        for (var i = 0; i < panels.length; i++) panels[i].hidden = true;
      });
    }
    return panel;
  }
  function paragraphTargets(inst) {
    var out = [];
    var sel = window.getSelection();
    if (sel && sel.rangeCount && !sel.isCollapsed && inst.editor.contains(sel.anchorNode)) {
      var range = sel.getRangeAt(0);
      var nodes = inst.editor.querySelectorAll('p,h1,h2,h3,h4,li,pre');
      for (var i = 0; i < nodes.length; i++) {
        var hit = false;
        try { hit = range.intersectsNode(nodes[i]); } catch (e) { hit = false; }
        if (hit) out.push(nodes[i]);
      }
    }
    if (!out.length) {
      var block = currentBlock(inst);
      if (block) out.push(block);
    }
    return out;
  }
  function fillParagraphSettings(inst, panel, values) {
    var block = currentBlock(inst);
    var source = values || {};
    if (!values && block) {
      source = {
        align: block.style.textAlign || '',
        indent: parseFloat(block.style.paddingLeft) || 0,
        lineheight: parseFloat(block.style.lineHeight) || '',
        before: parseFloat(block.style.marginTop) || 0,
        after: parseFloat(block.style.marginBottom) || 0,
        hyphens: block.style.hyphens || block.style.webkitHyphens || '',
        tabSize: block.style.tabSize || '2',
        lineNumbers: block.getAttribute && block.getAttribute('data-line-numbers') !== null,
        codeLanguage: (block.getAttribute && block.getAttribute('data-code-language')) || 'plain'
      };
    }
    var align = panel.querySelector('[data-rich-paragraph-align]');
    if (align) align.value = (source.align === 'justify' && source.hyphens === 'auto') ? 'justifyHyphen' : (source.align || '');
    var hyphenInput = panel.querySelector('[data-rich-paragraph-hyphens]');
    if (hyphenInput) hyphenInput.checked = source.hyphens === 'auto';
    var tabInput = panel.querySelector('[data-rich-paragraph-tabsize]');
    if (tabInput) tabInput.value = source.tabSize || '2';
    var lineNumInput = panel.querySelector('[data-rich-paragraph-linenumbers]');
    if (lineNumInput) lineNumInput.checked = !!source.lineNumbers;
    var codeLangInput = panel.querySelector('[data-rich-paragraph-codelang]');
    if (codeLangInput) codeLangInput.value = source.codeLanguage || 'plain';
    [['indent', 'indent'], ['lineheight', 'lineheight'], ['before', 'before'], ['after', 'after']].forEach(function (pair) {
      var input = panel.querySelector('[data-rich-paragraph-' + pair[1] + ']');
      if (input) input.value = source[pair[0]] == null ? '' : String(source[pair[0]]);
    });
  }
  function changeParagraphIndent(inst, delta) {
    var blocks = paragraphTargets(inst);
    for (var i = 0; i < blocks.length; i++) {
      var block = blocks[i];
      if (!block || !block.style) continue;
      if (block.tagName && block.tagName.toLowerCase() === 'pre') {
        indentCodeLines(inst, block, delta);
        continue;
      }
      var current = parseFloat(block.style.paddingLeft || block.style.marginLeft || '0') || 0;
      current = Math.max(0, Math.min(240, current + delta));
      block.style.paddingLeft = current + 'px';
    }
    if (!blocks.length) return;
    commit(inst, 'change');
    fillParagraphSettings(inst, paragraphSettingsPanel(inst), null);
  }
  function applyParagraphSettings(inst) {
    inst = resolveInstance(inst.name) || inst;
    var panel = paragraphSettingsPanel(inst);
    function value(sel) { var node = panel.querySelector(sel); return node ? String(node.value || '') : ''; }
    var alignValue = value('[data-rich-paragraph-align]');
    var justifyHyphen = alignValue === 'justifyHyphen';
    var align = justifyHyphen ? 'justify' : alignValue;
    var indent = parseFloat(value('[data-rich-paragraph-indent]'));
    var lineheight = parseFloat(value('[data-rich-paragraph-lineheight]'));
    var before = parseFloat(value('[data-rich-paragraph-before]'));
    var after = parseFloat(value('[data-rich-paragraph-after]'));
    var hyphensInput = panel.querySelector('[data-rich-paragraph-hyphens]');
    var hyphens = !!(hyphensInput && hyphensInput.checked) || justifyHyphen;
    var tabSizeInput = panel.querySelector('[data-rich-paragraph-tabsize]');
    var tabSize = tabSizeInput ? String(tabSizeInput.value || '2') : '2';
    var lineNumInput = panel.querySelector('[data-rich-paragraph-linenumbers]');
    var lineNumbers = !!(lineNumInput && lineNumInput.checked);
    var codeLangInput = panel.querySelector('[data-rich-paragraph-codelang]');
    var codeLang = codeLanguage(codeLangInput ? codeLangInput.value : 'plain');
    var blocks = paragraphTargets(inst);
    for (var i = 0; i < blocks.length; i++) {
      var block = blocks[i];
      var currentAlign = block.style.textAlign || '';
      if (align !== currentAlign) block.style.textAlign = align;
      var desiredIndent = isFinite(indent) && indent > 0 ? indent : 0;
      var currentIndent = parseFloat(block.style.paddingLeft) || 0;
      if (Math.abs(desiredIndent - currentIndent) > 0.01) block.style.paddingLeft = desiredIndent > 0 ? desiredIndent + 'px' : '';
      var desiredLineHeight = isFinite(lineheight) && lineheight > 0 ? lineheight : 0;
      var currentLineHeight = parseFloat(block.style.lineHeight) || 0;
      if (Math.abs(desiredLineHeight - currentLineHeight) > 0.01) {
        block.style.lineHeight = desiredLineHeight > 0 ? String(desiredLineHeight) : '';
        if (block.removeAttribute) block.removeAttribute('data-line-rule');
      }
      var desiredBefore = isFinite(before) && before > 0 ? before : 0;
      var currentBefore = parseFloat(block.style.marginTop) || 0;
      if (Math.abs(desiredBefore - currentBefore) > 0.01) block.style.marginTop = desiredBefore > 0 ? desiredBefore + 'px' : '';
      var desiredAfter = isFinite(after) && after > 0 ? after : 0;
      var currentAfter = parseFloat(block.style.marginBottom) || 0;
      if (Math.abs(desiredAfter - currentAfter) > 0.01) block.style.marginBottom = desiredAfter > 0 ? desiredAfter + 'px' : '';
      if (block.tagName && block.tagName.toLowerCase() === 'pre') {
        block.style.tabSize = tabSize;
        if (lineNumbers) block.setAttribute('data-line-numbers', codeLineNumbers(block.querySelector('code') || block));
        else block.removeAttribute('data-line-numbers');
        var codeBox = block.querySelector('code') || block;
        clearCodeHighlight(codeBox);
        if (codeLang === 'plain') block.removeAttribute('data-code-language');
        else {
          block.setAttribute('data-code-language', codeLang);
          highlightCodeBox(codeBox, codeLang);
        }
      }
      var isPre = block.tagName && block.tagName.toLowerCase() === 'pre';
      var hasHyphens = block.style.hyphens === 'auto' || block.style.webkitHyphens === 'auto';
      if (!isPre && hyphens !== hasHyphens) {
        if (hyphens) {
          block.style.hyphens = 'auto';
          block.style.webkitHyphens = 'auto';
          softHyphenateElement(block);
        } else {
          block.style.hyphens = '';
          block.style.webkitHyphens = '';
        }
      }
      if (!isPre && !hyphens) removeSoftHyphens(block);
    }
    commit(inst, 'change');
    panel.hidden = true;
  }
  function openParagraphSettings(inst) {
    var liveRoot = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"]');
    if (liveRoot) {
      inst.root = liveRoot;
      var liveEditor = liveRoot.querySelector('.sui-rich-editor');
      if (liveEditor) inst.editor = liveEditor;
    }
    var panel = paragraphSettingsPanel(inst);
    var rootStyle = getComputedStyle(inst.root);
    ['--sui-rich-bg', '--sui-rich-fg', '--sui-rich-border', '--sui-rich-muted', '--sui-rich-hover', '--sui-rich-accent', '--sui-rich-shadow'].forEach(function (name) {
      var value = rootStyle.getPropertyValue(name);
      if (value) panel.style.setProperty(name, value);
    });
    fillParagraphSettings(inst, panel, null);
    panel.hidden = false;
    var trigger = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"] .sui-rich-toolbar [data-rich-cmd="paragraphSettings"]');
    var rect = trigger ? trigger.getBoundingClientRect() : inst.root.getBoundingClientRect();
    var panelW = 292, panelH = 360;
    var left = rect && rect.width ? rect.left : 16;
    var top = rect && rect.height ? rect.bottom + 6 : 16;
    left = Math.max(8, Math.min(window.innerWidth - panelW, left));
    if (top + panelH > window.innerHeight) top = Math.max(8, window.innerHeight - panelH);
    panel.style.left = left + 'px';
    panel.style.top = top + 'px';
    var first = panel.querySelector('[data-rich-paragraph-align]');
    if (first) first.focus();
  }
  function imageTarget(inst) {
    if (inst.lastImage && inst.lastImage.isConnected && inst.editor && inst.editor.contains(inst.lastImage)) return inst.lastImage;
    var sel = window.getSelection();
    if (sel && sel.rangeCount) {
      var node = sel.anchorNode;
      var element = node && node.nodeType === 1 ? node : (node && node.parentElement);
      var img = element && element.closest ? element.closest('img') : null;
      if (img && inst.editor.contains(img)) return img;
    }
    var images = inst.editor ? inst.editor.querySelectorAll('img') : [];
    return images.length ? images[0] : null;
  }
  function selectImage(inst, img) {
    if (inst.lastImage && inst.lastImage !== img) inst.lastImage.classList.remove('sui-rich-image-selected');
    inst.lastImage = img;
    if (img) img.classList.add('sui-rich-image-selected');
    updateStats(inst);
  }
  function closeImageMenu() {
    var menu = document.querySelector('.sui-rich-image-menu');
    if (menu && menu.parentNode) menu.parentNode.removeChild(menu);
  }
  function openImageMenu(inst, img, x, y) {
    closeImageMenu();
    var menu = el('div', 'sui-rich-image-menu');
    menu.setAttribute('role', 'menu');
    [['Edit image', 'edit'], ['Replace image...', 'replace'], ['Delete image', 'delete']].forEach(function (item) {
      var button = el('button', 'sui-rich-menu-item', item[0]);
      button.type = 'button';
      button.setAttribute('role', 'menuitem');
      button.setAttribute('data-rich-image-action', item[1]);
      menu.appendChild(button);
    });
    document.body.appendChild(menu);
    menu.style.left = Math.max(8, Math.min(window.innerWidth - 190, x)) + 'px';
    menu.style.top = Math.max(8, Math.min(window.innerHeight - 120, y)) + 'px';
    menu.addEventListener('click', function (ev) {
      var target = ev.target;
      if (!target || !target.closest) return;
      var action = target.getAttribute('data-rich-image-action');
      closeImageMenu();
      if (action === 'edit') openImageSettings(inst, img);
      else if (action === 'replace') {
        openImageSettings(inst, img);
        var panel = imageSettingsPanel(inst);
        var file = panel.querySelector('[data-rich-image-file]');
        if (file) file.click();
      } else if (action === 'delete') {
        var images = inst.editor.querySelectorAll('img');
        inst.imageIndex = Array.prototype.indexOf.call(images, img);
        removeImage(inst);
      }
    });
    var close = function (ev) {
      if (ev && ev.target && ev.target.closest && ev.target.closest('.sui-rich-image-menu')) return;
      closeImageMenu();
      document.removeEventListener('mousedown', close, true);
      document.removeEventListener('keydown', onKey);
    };
    function onKey(ev) { if (ev.key === 'Escape') { ev.preventDefault(); close(); } }
    setTimeout(function () {
      document.addEventListener('mousedown', close, true);
      document.addEventListener('keydown', onKey);
    }, 0);
  }
  function imageSettingsPanel(inst) {
    var panel = document.querySelector('.sui-rich-imagesettings[data-rich-imagesettings="' + inst.name + '"]');
    if (panel) return panel;
    panel = el('div', 'sui-rich-imagesettings');
    panel.setAttribute('data-rich-imagesettings', inst.name);
    panel.hidden = true;
    panel.appendChild(panelTitle('Image settings', 'Edit the URL or choose a file to replace the image. Width is in pixels; leave it empty for automatic.'));

    var urlRow = el('label');
    urlRow.appendChild(el('span', '', 'URL'));
    var url = el('input');
    url.type = 'text';
    url.setAttribute('data-rich-image-url', '1');
    url.setAttribute('aria-label', 'Image URL');
    url.title = 'Image URL or data URL';
    url.placeholder = 'https://example.com/image.png';
    urlRow.appendChild(url);
    panel.appendChild(urlRow);

    var fileRow = el('label');
    fileRow.appendChild(el('span', '', 'Replace'));
    var file = el('input');
    file.type = 'file';
    file.accept = 'image/*';
    file.setAttribute('data-rich-image-file', '1');
    file.setAttribute('aria-label', 'Replace image file');
    fileRow.appendChild(file);
    panel.appendChild(fileRow);

    var altRow = el('label');
    altRow.appendChild(el('span', '', 'Alt text'));
    var alt = el('input');
    alt.type = 'text';
    alt.setAttribute('data-rich-image-alt', '1');
    alt.setAttribute('aria-label', 'Alternative text');
    alt.placeholder = 'Describe the image';
    altRow.appendChild(alt);
    panel.appendChild(altRow);

    var widthRow = el('label');
    widthRow.appendChild(el('span', '', 'Width'));
    var width = el('input');
    width.type = 'number';
    width.min = '0';
    width.step = '1';
    width.setAttribute('data-rich-image-width', '1');
    width.setAttribute('aria-label', 'Image width in pixels; empty for automatic');
    width.title = 'Width in pixels; leave empty for automatic';
    widthRow.appendChild(width);
    panel.appendChild(widthRow);

    var alignRow = el('label');
    alignRow.appendChild(el('span', '', 'Alignment'));
    var align = el('select');
    align.setAttribute('data-rich-image-align', '1');
    align.setAttribute('aria-label', 'Image alignment');
    [['', 'Default'], ['left', 'Left'], ['center', 'Center'], ['right', 'Right']].forEach(function (o) { var op = el('option', '', o[1]); op.value = o[0]; align.appendChild(op); });
    alignRow.appendChild(align);
    panel.appendChild(alignRow);
    var error = el('p', 'sui-rich-imagesettings-error');
    error.hidden = true;
    panel.appendChild(error);

    var actions = el('div', 'sui-rich-pagesettings-actions');
    var apply = el('button', 'sui-rich-btn', 'Apply'); apply.type = 'button'; apply.setAttribute('data-rich-image-apply', '1');
    var remove = el('button', 'sui-rich-btn', 'Remove'); remove.type = 'button'; remove.setAttribute('data-rich-image-remove', '1');
    var close = el('button', 'sui-rich-btn', 'Close'); close.type = 'button'; close.setAttribute('data-rich-image-close', '1');
    actions.appendChild(apply); actions.appendChild(remove); actions.appendChild(close);
    panel.appendChild(actions);

    file.addEventListener('change', function () {
      var picked = file.files && file.files[0];
      if (!picked) return;
      var reader = new FileReader();
      reader.onload = function () {
        var value = String(reader.result || '');
        if (!/^data:image\//i.test(value)) { error.textContent = 'Choose an image file.'; error.hidden = false; return; }
        url.value = value;
        error.hidden = true;
      };
      reader.readAsDataURL(picked);
    });
    panel.addEventListener('click', function (ev) {
      var target = ev.target;
      if (!target || !target.closest) return;
      if (target.closest('[data-rich-image-apply]')) { applyImageSettings(inst); return; }
      if (target.closest('[data-rich-image-remove]')) { removeImage(inst); return; }
      if (target.closest('[data-rich-image-close]')) { panel.hidden = true; return; }
    });
    document.body.appendChild(panel);
    if (!window.__suiImageSettingsBound) {
      window.__suiImageSettingsBound = true;
      document.addEventListener('mousedown', function (ev) {
        var target = ev.target;
        if (target && target.closest && (target.closest('.sui-rich-imagesettings') || target.closest('[data-rich-cmd="imageSettings"]'))) return;
        var panels = document.querySelectorAll('.sui-rich-imagesettings');
        for (var i = 0; i < panels.length; i++) panels[i].hidden = true;
      });
    }
    return panel;
  }
  function openImageSettings(inst, img) {
    inst = resolveInstance(inst.name) || inst;
    var liveRoot = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"]');
    if (liveRoot) {
      inst.root = liveRoot;
      var liveEditor = liveRoot.querySelector('.sui-rich-editor');
      if (liveEditor) inst.editor = liveEditor;
    }
    img = img || imageTarget(inst);
    if (!img) return;
    selectImage(inst, img);
    var images = inst.editor.querySelectorAll('img');
    inst.imageIndex = Array.prototype.indexOf.call(images, img);
    var panel = imageSettingsPanel(inst);
    var rootStyle = getComputedStyle(inst.root);
    ['--sui-rich-bg', '--sui-rich-fg', '--sui-rich-border', '--sui-rich-muted', '--sui-rich-hover', '--sui-rich-accent', '--sui-rich-shadow'].forEach(function (name) {
      var value = rootStyle.getPropertyValue(name);
      if (value) panel.style.setProperty(name, value);
    });
    var url = panel.querySelector('[data-rich-image-url]');
    var alt = panel.querySelector('[data-rich-image-alt]');
    var width = panel.querySelector('[data-rich-image-width]');
    var align = panel.querySelector('[data-rich-image-align]');
    var error = panel.querySelector('.sui-rich-imagesettings-error');
    if (url) url.value = img.getAttribute('src') || '';
    if (alt) alt.value = img.getAttribute('alt') || '';
    if (width) width.value = parseFloat(img.style.width) ? String(Math.round(parseFloat(img.style.width))) : '';
    if (align) align.value = img.style.textAlign || '';
    if (error) error.hidden = true;
    panel.hidden = false;
    var rect = img.getBoundingClientRect();
    var panelW = 300, panelH = 390;
    var left = rect && rect.width ? rect.left : 16;
    var top = rect && rect.height ? rect.bottom + 6 : 16;
    left = Math.max(8, Math.min(window.innerWidth - panelW, left));
    if (top + panelH > window.innerHeight) top = Math.max(8, window.innerHeight - panelH);
    panel.style.left = left + 'px';
    panel.style.top = top + 'px';
    if (url) url.focus();
  }
  function applyImageSettings(inst) {
    inst = resolveInstance(inst.name) || inst;
    var panel = imageSettingsPanel(inst);
    var images = inst.editor.querySelectorAll('img');
    var img = images[inst.imageIndex] || inst.lastImage;
    if (!img) return;
    var url = panel.querySelector('[data-rich-image-url]').value;
    var src = safeUrl(url, true);
    var error = panel.querySelector('.sui-rich-imagesettings-error');
    if (!src) { error.textContent = 'Use an http(s) or data:image URL.'; error.hidden = false; return; }
    error.hidden = true;
    img.setAttribute('src', src);
    img.setAttribute('alt', panel.querySelector('[data-rich-image-alt]').value || '');
    var width = parseFloat(panel.querySelector('[data-rich-image-width]').value);
    img.style.width = isFinite(width) && width > 0 ? Math.round(width) + 'px' : '';
    var align = panel.querySelector('[data-rich-image-align]').value;
    img.style.textAlign = align || '';
    img.style.display = align ? 'block' : '';
    img.style.marginLeft = align === 'right' || align === 'center' ? 'auto' : (align === 'left' ? '0' : '');
    img.style.marginRight = align === 'left' || align === 'center' ? 'auto' : (align === 'right' ? '0' : '');
    commit(inst, 'change');
    panel.hidden = true;
  }
  function removeImage(inst) {
    inst = resolveInstance(inst.name) || inst;
    var panel = imageSettingsPanel(inst);
    var images = inst.editor.querySelectorAll('img');
    var img = images[inst.imageIndex] || inst.lastImage;
    if (img && img.parentNode) img.parentNode.removeChild(img);
    inst.lastImage = null;
    commit(inst, 'change');
    panel.hidden = true;
  }
  function chromeSummaryText(value) {
    var text = String(chromePlainText(value) || '').replace(/\s+/g, ' ').trim();
    if (text) return text.length > 80 ? text.slice(0, 77) + '...' : text;
    return chromeHasImage(value) ? 'Image' : 'Empty';
  }
  function syncChromePanel(inst) {
    if (!inst) return;
    var inputs = document.querySelectorAll('[data-rich-doc-header]');
    for (var h = 0; h < inputs.length; h++) inputs[h].value = inst.headerText || '';
    inputs = document.querySelectorAll('[data-rich-doc-footer]');
    for (var f = 0; f < inputs.length; f++) inputs[f].value = inst.footerText || '';
    var summaries = document.querySelectorAll('[data-rich-chrome-summary]');
    for (var i = 0; i < summaries.length; i++) {
      var which = summaries[i].getAttribute('data-rich-chrome-summary');
      summaries[i].textContent = chromeSummaryText(which === 'footer' ? inst.footerValue : inst.headerValue);
    }
  }
  function chromeHiddenInputs(inst, panel) {
    ['header', 'footer'].forEach(function (which) {
      var input = el('input');
      input.type = 'hidden';
      input.setAttribute('data-rich-doc-' + which, '1');
      input.value = which === 'header' ? (inst.headerText || '') : (inst.footerText || '');
      panel.appendChild(input);
    });
  }
  function chromeOpenButton(inst, panel) {
    var open = el('button', 'sui-rich-btn sui-rich-chrome-open', 'Header and footer...');
    open.type = 'button';
    open.setAttribute('data-rich-open-chrome', '1');
    open.title = 'Open the header and footer editor with Header and Footer tabs';
    panel.appendChild(open);
    return open;
  }
  function applyChromeTheme(inst, node) {
    if (!inst || !inst.root || !node) return;
    var rootStyle = getComputedStyle(inst.root);
    ['--sui-rich-bg', '--sui-rich-fg', '--sui-rich-border', '--sui-rich-muted', '--sui-rich-hover', '--sui-rich-accent', '--sui-rich-shadow'].forEach(function (name) {
      var value = rootStyle.getPropertyValue(name);
      if (value) node.style.setProperty(name, value);
    });
  }
  function closeChromeEditor() {
    var overlay = document.querySelector('.sui-rich-chrome-editor');
    if (overlay && overlay.parentNode) overlay.parentNode.removeChild(overlay);
  }
  function openChromeHelp(inst) {
    var existing = document.querySelector('.sui-rich-chrome-help-overlay');
    if (existing && existing.parentNode) existing.parentNode.removeChild(existing);
    var overlay = el('div', 'sui-rich-chrome-help-overlay');
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-modal', 'true');
    overlay.setAttribute('aria-label', 'Header and footer variables');
    var card = el('div', 'sui-rich-chrome-help-card');
    applyChromeTheme(inst, card);
    card.appendChild(el('div', 'sui-rich-tableprops-title', 'Header and footer variables'));
    var rows = [
      ['{page}', 'Current page number'],
      ['{pages}', 'Total number of pages'],
      ['{title}', 'Document title'],
      ['{name}', 'Editor name'],
      ['{date}', 'Current date'],
      ['{time}', 'Current time'],
      ['{datetime}', 'Current date and time'],
      ['{words}', 'Word count'],
      ['{chars}', 'Character count']
    ];
    for (var i = 0; i < rows.length; i++) {
      var row = el('div', 'sui-rich-chrome-help-row');
      row.appendChild(el('span', 'sui-rich-chrome-help-code', rows[i][0]));
      row.appendChild(el('span', '', rows[i][1]));
      card.appendChild(row);
    }
    card.appendChild(el('div', 'sui-rich-chrome-help-note', 'Square brackets are accepted too: [page], [pages], ... Tokens are case-insensitive and can be mixed with text and images.'));
    var close = el('button', 'sui-rich-btn', 'Close');
    close.type = 'button';
    card.appendChild(close);
    overlay.appendChild(card);
    document.body.appendChild(overlay);
    function cleanup() {
      document.removeEventListener('keydown', onKey);
      if (overlay.parentNode) overlay.parentNode.removeChild(overlay);
    }
    function onKey(ev) { if (ev.key === 'Escape') { ev.preventDefault(); cleanup(); } }
    close.addEventListener('click', cleanup);
    overlay.addEventListener('mousedown', function (ev) { if (ev.target === overlay) cleanup(); });
    document.addEventListener('keydown', onKey);
  }
  function openChromeEditor(inst, initialWhich) {
    closeChromeEditor();
    var overlay = el('div', 'sui-rich-chrome-editor');
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-modal', 'true');
    overlay.setAttribute('aria-label', 'Edit header and footer');
    var card = el('div', 'sui-rich-chrome-editor-card');
    applyChromeTheme(inst, card);
    var head = el('div', 'sui-rich-chrome-editor-head');
    head.appendChild(el('strong', '', 'Header and footer'));
    var help = el('button', 'sui-rich-btn sui-rich-chrome-help', '?');
    help.type = 'button';
    help.title = 'Header/footer variables help';
    help.setAttribute('aria-label', help.title);
    var close = el('button', 'sui-rich-btn', 'Close');
    close.type = 'button';
    head.appendChild(help); head.appendChild(close);
    card.appendChild(head);
    var tabs = el('div', 'sui-rich-chrome-tabs');
    var headerTab = el('button', 'sui-rich-chrome-tab', 'Header');
    headerTab.type = 'button';
    var footerTab = el('button', 'sui-rich-chrome-tab', 'Footer');
    footerTab.type = 'button';
    tabs.appendChild(headerTab); tabs.appendChild(footerTab);
    card.appendChild(tabs);
    var toolbar = el('div', 'sui-rich-chrome-toolbar');
    card.appendChild(toolbar);
    var surfaces = el('div', 'sui-rich-chrome-surfaces');
    card.appendChild(surfaces);
    var actions = el('div', 'sui-rich-chrome-actions');
    var apply = el('button', 'sui-rich-btn', 'Apply');
    apply.type = 'button';
    var cancel = el('button', 'sui-rich-btn', 'Cancel');
    cancel.type = 'button';
    actions.appendChild(apply); actions.appendChild(cancel);
    card.appendChild(actions);
    overlay.appendChild(card);

    function makeSurface(value, which) {
      var surface = el('div', 'sui-rich-chrome-surface');
      surface.setAttribute('contenteditable', 'true');
      surface.setAttribute('role', 'textbox');
      surface.setAttribute('aria-multiline', 'true');
      surface.setAttribute('aria-label', which === 'header' ? 'Header content' : 'Footer content');
      surface.setAttribute('data-rich-chrome-surface', which);
      surface.appendChild(renderDoc(chromeDoc(value)));
      return surface;
    }
    var headerSurface = makeSurface(inst.headerValue, 'header');
    var footerSurface = makeSurface(inst.footerValue, 'footer');
    surfaces.appendChild(headerSurface); surfaces.appendChild(footerSurface);
    var activeWhich = initialWhich === 'footer' ? 'footer' : 'header';
    var lastRange = null;

    function active() { return activeWhich === 'footer' ? footerSurface : headerSurface; }
    function focusSurface(surface) {
      surface.focus();
      try {
        var range = document.createRange();
        range.selectNodeContents(surface);
        range.collapse(false);
        var sel = window.getSelection();
        sel.removeAllRanges();
        sel.addRange(range);
      } catch (e) {}
    }
    function setTab(which) {
      activeWhich = which;
      headerTab.classList.toggle('on', which === 'header');
      footerTab.classList.toggle('on', which === 'footer');
      headerSurface.hidden = which !== 'header';
      footerSurface.hidden = which !== 'footer';
      lastRange = null;
      focusSurface(active());
    }
    function saveRange() {
      var sel = window.getSelection();
      var surface = active();
      if (sel && sel.rangeCount && surface.contains(sel.anchorNode)) lastRange = sel.getRangeAt(0).cloneRange();
      else lastRange = null;
    }
    function restoreRange() {
      if (!lastRange) return false;
      var surface = active();
      if (!surface.contains(lastRange.startContainer)) { lastRange = null; return false; }
      surface.focus();
      var sel = window.getSelection();
      sel.removeAllRanges();
      sel.addRange(lastRange);
      return true;
    }
    function exec(cmd) {
      active().focus();
      try {
        document.execCommand('styleWithCSS', false, false);
        document.execCommand(cmd, false, null);
      } catch (e) {}
    }
    function insertToken(token) {
      var surface = active();
      if (!restoreRange()) surface.focus();
      try {
        document.execCommand('insertText', false, token);
      } catch (e) {
        var sel = window.getSelection();
        if (sel && sel.rangeCount && surface.contains(sel.anchorNode)) {
          var range = sel.getRangeAt(0);
          range.deleteContents();
          range.insertNode(document.createTextNode(token));
          range.collapse(false);
          sel.removeAllRanges();
          sel.addRange(range);
        } else {
          surface.appendChild(document.createTextNode(token));
        }
      }
      lastRange = null;
    }
    function tool(text, title, action) {
      var button = el('button', 'sui-rich-btn', text);
      button.type = 'button';
      button.title = title;
      button.setAttribute('aria-label', title);
      button.addEventListener('mousedown', function (ev) { ev.preventDefault(); });
      button.addEventListener('click', function (ev) { ev.preventDefault(); action(); });
      toolbar.appendChild(button);
      return button;
    }
    tool('B', 'Bold', function () { exec('bold'); });
    tool('I', 'Italic', function () { exec('italic'); });
    tool('U', 'Underline', function () { exec('underline'); });
    tool('L', 'Align left', function () { exec('justifyLeft'); });
    tool('C', 'Align center', function () { exec('justifyCenter'); });
    tool('R', 'Align right', function () { exec('justifyRight'); });
    var imageInput = el('input');
    imageInput.type = 'file';
    imageInput.accept = 'image/*';
    imageInput.style.display = 'none';
    toolbar.appendChild(imageInput);
    var imageBtn = tool('Image', 'Insert image', function () {
      saveRange();
      imageInput.click();
    });
    imageInput.addEventListener('change', function () {
      var picked = imageInput.files && imageInput.files[0];
      if (!picked) return;
      var reader = new FileReader();
      reader.onload = function () {
        var surface = active();
        var img = el('img');
        img.setAttribute('src', String(reader.result || ''));
        img.setAttribute('alt', '');
        img.style.maxWidth = '100%';
        img.style.maxHeight = '100px';
        if (restoreRange()) {
          var sel = window.getSelection();
          var range = sel.getRangeAt(0);
          range.deleteContents();
          range.insertNode(img);
          range.setStartAfter(img);
          range.collapse(true);
          sel.removeAllRanges();
          sel.addRange(range);
        } else {
          var p = el('p');
          p.appendChild(img);
          surface.appendChild(p);
          focusSurface(surface);
        }
        lastRange = null;
        imageInput.value = '';
      };
      reader.readAsDataURL(picked);
    });
    tool('{page}', 'Insert current page number', function () { insertToken('{page}'); });
    tool('{pages}', 'Insert total page count', function () { insertToken('{pages}'); });
    var vars = el('select', 'sui-rich-chrome-vars');
    vars.title = 'Insert variable';
    vars.setAttribute('aria-label', 'Insert variable');
    var placeholder = el('option', '', 'Variables...');
    placeholder.value = '';
    vars.appendChild(placeholder);
    [
      ['{title}', 'Title'],
      ['{name}', 'Name'],
      ['{date}', 'Date'],
      ['{time}', 'Time'],
      ['{datetime}', 'Date and time'],
      ['{words}', 'Words'],
      ['{chars}', 'Characters']
    ].forEach(function (pair) {
      var option = el('option', '', pair[1]);
      option.value = pair[0];
      vars.appendChild(option);
    });
    vars.addEventListener('mousedown', saveRange);
    vars.addEventListener('change', function () {
      if (vars.value) insertToken(vars.value);
      vars.value = '';
    });
    toolbar.appendChild(vars);
    tool('Clear', 'Clear active section', function () {
      var surface = active();
      surface.innerHTML = '';
      var p = el('p');
      p.appendChild(el('br'));
      surface.appendChild(p);
      focusSurface(surface);
      lastRange = null;
    });
    tool('?', 'Header/footer variables help', function () { openChromeHelp(inst); });

    headerTab.addEventListener('click', function () { setTab('header'); });
    footerTab.addEventListener('click', function () { setTab('footer'); });
    help.addEventListener('click', function () { openChromeHelp(inst); });
    apply.addEventListener('click', function () {
      var headerModel = docFromDom(headerSurface);
      var footerModel = docFromDom(footerSurface);
      inst.doc.header = chromeHasContent(headerModel) ? headerModel : '';
      inst.doc.footer = chromeHasContent(footerModel) ? footerModel : '';
      inst.headerOverride = true;
      inst.footerOverride = true;
      refreshChromeValues(inst);
      syncChromePanel(inst);
      refreshPageChrome(inst);
      updatePageGuides(inst);
      if (inst.paginated) schedulePaginate(inst); else schedulePageGuides(inst);
      commit(inst, 'change');
      closeChromeEditor();
    });
    cancel.addEventListener('click', closeChromeEditor);
    close.addEventListener('click', closeChromeEditor);
    overlay.addEventListener('mousedown', function (ev) { if (ev.target === overlay) closeChromeEditor(); });
    setTab(activeWhich);
    document.body.appendChild(overlay);
  }
  function documentSettingsPanel(inst) {
    var panel = document.querySelector('.sui-rich-documentsettings[data-rich-documentsettings="' + inst.name + '"]');
    if (panel) return panel;
    panel = el('div', 'sui-rich-documentsettings');
    panel.setAttribute('data-rich-documentsettings', inst.name);
    panel.hidden = true;
    panel.appendChild(panelTitle('Document settings', 'General defaults for every page; Page settings can override size, orientation and margins for a single page. Use {page} and {pages} in header/footer.'));

    var sizeRow = el('label');
    sizeRow.appendChild(el('span', '', 'Size'));
    var sizeSelect = el('select');
    sizeSelect.setAttribute('data-rich-doc-size', '1');
    sizeSelect.setAttribute('aria-label', 'Default page size');
    [['a4', 'A4'], ['letter', 'Letter']].forEach(function (o) { var op = el('option', '', o[1]); op.value = o[0]; sizeSelect.appendChild(op); });
    sizeRow.appendChild(sizeSelect);
    panel.appendChild(sizeRow);

    var orientRow = el('label');
    orientRow.appendChild(el('span', '', 'Orientation'));
    var orientSelect = el('select');
    orientSelect.setAttribute('data-rich-doc-orientation', '1');
    orientSelect.setAttribute('aria-label', 'Default page orientation');
    [['portrait', 'Vertical'], ['landscape', 'Horizontal']].forEach(function (o) { var op = el('option', '', o[1]); op.value = o[0]; orientSelect.appendChild(op); });
    orientRow.appendChild(orientSelect);
    panel.appendChild(orientRow);

    panel.appendChild(el('div', 'sui-rich-tableprops-title', 'Margins (mm)'));
    var grid = el('div', 'sui-rich-pagesettings-grid');
    panel.appendChild(grid);
    pageMarginInput(grid, 'top', 'Top', 'doc');
    pageMarginInput(grid, 'right', 'Right', 'doc');
    pageMarginInput(grid, 'bottom', 'Bottom', 'doc');
    pageMarginInput(grid, 'left', 'Left', 'doc');

    var lineRow = el('label');
    lineRow.appendChild(el('span', '', 'Line height'));
    var lineInput = el('input');
    lineInput.type = 'number';
    lineInput.min = '0.8';
    lineInput.max = '3';
    lineInput.step = '0.05';
    lineInput.setAttribute('data-rich-doc-lineheight', '1');
    lineInput.setAttribute('aria-label', 'Default document line height');
    lineInput.title = 'Between 0.8 and 3';
    lineRow.appendChild(lineInput);
    panel.appendChild(lineRow);

    var langRow = el('label');
    langRow.appendChild(el('span', '', 'Language'));
    var langSelect = el('select');
    langSelect.setAttribute('data-rich-doc-lang', '1');
    langSelect.setAttribute('aria-label', 'Document language');
    langSelect.title = 'Used for spell checking and automatic hyphenation.';
    [['', 'Automatic'], ['it', 'Italiano'], ['en', 'English'], ['de', 'Deutsch'], ['fr', 'Français'], ['es', 'Español'], ['pt', 'Português'], ['nl', 'Nederlands'], ['pl', 'Polski'], ['sv', 'Svenska']].forEach(function (o) {
      var op = el('option', '', o[1]); op.value = o[0]; langSelect.appendChild(op);
    });
    langRow.appendChild(langSelect);
    panel.appendChild(langRow);

    chromeHiddenInputs(inst, panel);
    chromeOpenButton(inst, panel);

    var actions = el('div', 'sui-rich-pagesettings-actions');
    var apply = el('button', 'sui-rich-btn', 'Apply'); apply.type = 'button'; apply.setAttribute('data-rich-doc-apply', '1');
    var reset = el('button', 'sui-rich-btn', 'Reset'); reset.type = 'button'; reset.setAttribute('data-rich-doc-reset', '1');
    var close = el('button', 'sui-rich-btn', 'Close'); close.type = 'button'; close.setAttribute('data-rich-doc-close', '1');
    actions.appendChild(apply); actions.appendChild(reset); actions.appendChild(close);
    panel.appendChild(actions);

    panel.addEventListener('click', function (ev) {
      var target = ev.target;
      if (!target || !target.closest) return;
      if (target.closest('[data-rich-edit-chrome]')) { openChromeEditor(inst, target.closest('[data-rich-edit-chrome]').getAttribute('data-rich-edit-chrome')); return; }
      if (target.closest('[data-rich-open-chrome]')) { openChromeEditor(inst, 'header'); return; }
      if (target.closest('[data-rich-chrome-help]')) { openChromeHelp(inst); return; }
      if (target.closest('[data-rich-doc-apply]')) { applyDocumentSettings(inst); return; }
      if (target.closest('[data-rich-doc-reset]')) { fillDocumentSettings(inst, panel, defaultPageConfig()); return; }
      if (target.closest('[data-rich-doc-close]')) { panel.hidden = true; return; }
    });
    document.body.appendChild(panel);
    if (!window.__suiDocumentSettingsBound) {
      window.__suiDocumentSettingsBound = true;
      document.addEventListener('mousedown', function (ev) {
        var target = ev.target;
        if (target && target.closest && (target.closest('.sui-rich-documentsettings') || target.closest('[data-rich-cmd="documentSettings"]'))) return;
        var panels = document.querySelectorAll('.sui-rich-documentsettings');
        for (var i = 0; i < panels.length; i++) panels[i].hidden = true;
      });
    }
    return panel;
  }
  function fillDocumentSettings(inst, panel, page) {
    var size = panel.querySelector('[data-rich-doc-size]');
    var orient = panel.querySelector('[data-rich-doc-orientation]');
    if (size) size.value = page.size;
    if (orient) orient.value = page.orientation;
    ['top', 'right', 'bottom', 'left'].forEach(function (side) {
      var input = panel.querySelector('[data-rich-doc-margin="' + side + '"]');
      if (input) input.value = String(page.margins[side]);
    });
    var line = panel.querySelector('[data-rich-doc-lineheight]');
    if (line) line.value = String(page.lineHeight);
    var lang = panel.querySelector('[data-rich-doc-lang]');
    if (lang) lang.value = page.lang || '';
    var header = panel.querySelector('[data-rich-doc-header]');
    var footer = panel.querySelector('[data-rich-doc-footer]');
    if (header) { header.value = inst.headerText || ''; try { header.dispatchEvent(new Event('input')); } catch (e) {} }
    if (footer) { footer.value = inst.footerText || ''; try { footer.dispatchEvent(new Event('input')); } catch (e) {} }
    syncChromePanel(inst);
  }
  function applyDocumentSettings(inst) {
    inst = resolveInstance(inst.name) || inst;
    var panel = documentSettingsPanel(inst);
    function value(sel) { var node = panel.querySelector(sel); return node ? String(node.value || '') : ''; }
    var lh = parseFloat(value('[data-rich-doc-lineheight]'));
    if (!isFinite(lh)) lh = defaultPageConfig().lineHeight;
    function margin(side) {
      var node = panel.querySelector('[data-rich-doc-margin="' + side + '"]');
      return node ? clampPageMargin(node.value, defaultPageConfig().margins[side]) : defaultPageConfig().margins[side];
    }
    var current = documentPageConfig(inst);
    current.size = value('[data-rich-doc-size]') === 'letter' ? 'letter' : 'a4';
    current.orientation = value('[data-rich-doc-orientation]') === 'landscape' ? 'landscape' : 'portrait';
    current.margins = { top: margin('top'), right: margin('right'), bottom: margin('bottom'), left: margin('left') };
    current.lineHeight = Math.max(0.8, Math.min(3, Math.round(lh * 100) / 100));
    var langValue = value('[data-rich-doc-lang]').trim();
    current.lang = /^[A-Za-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$/.test(langValue) ? langValue : '';
    var header = value('[data-rich-doc-header]').slice(0, 500).trim();
    var footer = value('[data-rich-doc-footer]').slice(0, 500).trim();
    inst.doc.page = current;
    if (header !== inst.headerText) inst.doc.header = header;
    if (footer !== inst.footerText) inst.doc.footer = footer;
    inst.headerOverride = true;
    inst.footerOverride = true;
    if (inst.editor) inst.editor.setAttribute('lang', documentLanguage(inst));
    hyphenationLanguage = documentLanguage(inst);
    refreshChromeValues(inst);
    applyPageLineHeight(inst);
    refreshPageChrome(inst);
    updatePageGuides(inst);
    if (inst.paginated) schedulePaginate(inst);
    commit(inst, 'change');
    panel.hidden = true;
  }
  function openDocumentSettings(inst) {
    var liveRoot = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"]');
    if (liveRoot) {
      inst.root = liveRoot;
      var liveEditor = liveRoot.querySelector('.sui-rich-editor');
      if (liveEditor) inst.editor = liveEditor;
    }
    var panel = documentSettingsPanel(inst);
    var rootStyle = getComputedStyle(inst.root);
    ['--sui-rich-bg', '--sui-rich-fg', '--sui-rich-border', '--sui-rich-muted', '--sui-rich-hover', '--sui-rich-accent', '--sui-rich-shadow'].forEach(function (name) {
      var value = rootStyle.getPropertyValue(name);
      if (value) panel.style.setProperty(name, value);
    });
    fillDocumentSettings(inst, panel, documentPageConfig(inst));
    panel.hidden = false;
    var trigger = document.querySelector('[data-sui-rich][data-sui-rich-name="' + inst.name + '"] .sui-rich-toolbar [data-rich-cmd="documentSettings"]');
    var rect = trigger ? trigger.getBoundingClientRect() : inst.root.getBoundingClientRect();
    var panelW = 300, panelH = 470;
    var left = rect && rect.width ? rect.left : 16;
    var top = rect && rect.height ? rect.bottom + 6 : 16;
    left = Math.max(8, Math.min(window.innerWidth - panelW, left));
    if (top + panelH > window.innerHeight) top = Math.max(8, window.innerHeight - panelH);
    panel.style.left = left + 'px';
    panel.style.top = top + 'px';
    var first = panel.querySelector('[data-rich-doc-lineheight]');
    if (first) first.focus();
  }
  function escapePrintText(value) {
    return String(value == null ? '' : value).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }
  function printHTML(inst) {
    var body = exportHTML(inst);
    var page = pageConfig(inst, 1);
    var header = inst.headerText ? '<div class="sui-print-header">' + escapePrintText(inst.headerText.replace(/\{pages?\}|\[pages?\]/gi, '')) + '</div>' : '';
    var footer = inst.footerText ? '<div class="sui-print-footer">' + escapePrintText(inst.footerText.replace(/\{pages?\}|\[pages?\]/gi, '')) + '</div>' : '';
    var pageSize = (page.size === 'letter' ? 'Letter' : 'A4') + (page.orientation === 'landscape' ? ' landscape' : '');
    return '<!doctype html><html><head><meta charset="utf-8"><title>' + escapePrintText(inst.name) + '</title><style>' +
      '@page { size: ' + pageSize + '; margin: ' + page.margins.top + 'mm ' + page.margins.right + 'mm ' + page.margins.bottom + 'mm ' + page.margins.left + 'mm; }' +
      'body { margin: 0; font-family: Georgia, "Times New Roman", serif; color: #111; line-height: ' + page.lineHeight + '; }' +
      '.sui-print-header { position: fixed; top: 0; left: 0; right: 0; text-align: center; font-size: 10pt; color: #555; border-bottom: 1px solid #ccc; padding-bottom: 4px; }' +
      '.sui-print-footer { position: fixed; bottom: 0; left: 0; right: 0; text-align: center; font-size: 10pt; color: #555; border-top: 1px solid #ccc; padding-top: 4px; }' +
      'main { padding: 0; }' +
      'main img { max-width: 100%; height: auto; }' +
      'main table { border-collapse: collapse; width: 100%; }' +
      'main td, main th { border: 1px solid #999; padding: 5px 7px; vertical-align: top; }' +
      'main .sui-rich-pagebreak { break-after: page; page-break-after: always; border: 0; height: 0; margin: 0; font-size: 0; color: transparent; }' +
      '</style></head><body>' + header + '<main>' + body + '</main>' + footer + '</body></html>';
  }
  function printPreviewHTML(inst) {
    var page = pageConfig(inst, 1);
    var size = pageSizeMm(page);
    var body = exportHTML(inst);
    var head = inst.headerText ? '<div class="sui-preview-header">' + escapePrintText(sheetChromeText(inst.headerText, 1, 1, true)) + '</div>' : '';
    var foot = inst.footerText ? '<div class="sui-preview-footer">' + escapePrintText(sheetChromeText(inst.footerText, 1, 1, true)) + '</div>' : '';
    var pageSize = (page.size === 'letter' ? 'Letter' : 'A4') + (page.orientation === 'landscape' ? ' landscape' : '');
    return '<!doctype html><html><head><meta charset="utf-8"><title>' + escapePrintText(inst.name) + '</title><style>' +
      '@page { size: ' + pageSize + '; margin: ' + page.margins.top + 'mm ' + page.margins.right + 'mm ' + page.margins.bottom + 'mm ' + page.margins.left + 'mm; }' +
      'html, body { margin: 0; padding: 0; }' +
      'body { background: #e6e9ef; padding: 24px; font-family: Georgia, "Times New Roman", serif; color: #111; line-height: ' + page.lineHeight + '; }' +
      '.sui-preview-page { box-sizing: border-box; width: ' + size.w + 'mm; min-height: ' + size.h + 'mm; margin: 0 auto; padding: ' + page.margins.top + 'mm ' + page.margins.right + 'mm ' + page.margins.bottom + 'mm ' + page.margins.left + 'mm; background: #fff; box-shadow: 0 10px 28px rgba(0,0,0,.28); }' +
      '.sui-preview-header { text-align: center; font-size: 10pt; color: #555; border-bottom: 1px solid #ccc; padding-bottom: 4px; margin-bottom: 8px; }' +
      '.sui-preview-footer { text-align: center; font-size: 10pt; color: #555; border-top: 1px solid #ccc; padding-top: 4px; margin-top: 8px; }' +
      'main img { max-width: 100%; height: auto; }' +
      'main table { border-collapse: collapse; width: 100%; }' +
      'main td, main th { border: 1px solid #999; padding: 5px 7px; vertical-align: top; }' +
      'main .sui-rich-pagebreak { break-after: page; page-break-after: always; border: 0; height: 0; margin: 0; font-size: 0; color: transparent; }' +
      '@media print { body { background: #fff; padding: 0; } .sui-preview-page { width: auto; min-height: 0; margin: 0; padding: 0; box-shadow: none; } }' +
      '</style></head><body><div class="sui-preview-page">' + head + '<main>' + body + '</main>' + foot + '</div></body></html>';
  }
  async function openPrintPreview(inst, autoPrint) {
    inst = resolveInstance(inst.name) || inst;
    var existing = document.querySelector('.sui-rich-preview');
    if (existing && existing.parentNode) existing.parentNode.removeChild(existing);
    var page = pageConfig(inst), dims = pageSizeMm(page);
    var pageSize = (page.size === 'letter' ? 'Letter' : 'A4') + (page.orientation === 'landscape' ? ' landscape' : '');
    var pageStyle = document.getElementById('sui-preview-print-style');
    if (!pageStyle) { pageStyle = document.createElement('style'); pageStyle.id = 'sui-preview-print-style'; document.head.appendChild(pageStyle); }
    pageStyle.textContent = '@page { size: ' + pageSize + '; margin: 0; }';
    var overlay = el('div', 'sui-rich-preview');
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-modal', 'true');
    overlay.setAttribute('aria-label', 'Print preview');
    var card = el('div', 'sui-rich-preview-card');
    var bar = el('div', 'sui-rich-preview-bar');
    bar.appendChild(el('strong', '', 'Print preview'));
    var printBtn = el('button', 'sui-rich-btn', 'Print');
    printBtn.type = 'button';
    var closeBtn = el('button', 'sui-rich-btn', 'Close');
    closeBtn.type = 'button';
    bar.appendChild(printBtn); bar.appendChild(closeBtn);
    var pages = el('div', 'sui-rich-preview-pages');
    pages.appendChild(el('div', 'sui-rich-preview-loading', 'Preparing preview...'));
    card.appendChild(bar); card.appendChild(pages); overlay.appendChild(card);
    var rootStyle = getComputedStyle(inst.root);
    for (var vi = 0; vi < rootStyle.length; vi++) {
      var varName = rootStyle[vi];
      if (varName && varName.indexOf('--sui-rich') === 0) {
        var varValue = rootStyle.getPropertyValue(varName);
        if (varValue) overlay.style.setProperty(varName, varValue);
      }
    }
    document.body.appendChild(overlay);
    function close() {
      if (overlay.parentNode) overlay.parentNode.removeChild(overlay);
      document.removeEventListener('keydown', onKey);
      document.body.classList.remove('sui-rich-printing');
      if (pageStyle && pageStyle.parentNode) pageStyle.parentNode.removeChild(pageStyle);
    }
    function onKey(ev) { if (ev.key === 'Escape') { ev.preventDefault(); close(); } }
    document.addEventListener('keydown', onKey);
    closeBtn.addEventListener('click', close);
    printBtn.addEventListener('click', function () {
      document.body.classList.add('sui-rich-printing');
      var done = function () { document.body.classList.remove('sui-rich-printing'); window.removeEventListener('afterprint', done); };
      window.addEventListener('afterprint', done);
      try { window.focus(); window.print(); } catch (e) {}
      setTimeout(done, 1500);
    });
    overlay.addEventListener('mousedown', function (ev) { if (ev.target === overlay) close(); });
    printBtn.focus();
    // Reuse the editor pagination so the preview is split exactly like the
    // paginated view (headers/footers, repeated table headers, page numbers).
    var wasPaginated = !!inst.paginated || (inst.root && inst.root.getAttribute('data-sui-rich-paginated') === '1');
    try {
      sync(inst);
      if (!wasPaginated) setPaginated(inst, true);
      clearTimeout(inst.pageTimer);
      await paginate(inst);
      if (!overlay.parentNode) return;
      var wrap = inst.root.querySelector('.sui-rich-pages');
      pages.innerHTML = wrap ? wrap.innerHTML : '';
      if (!pages.children.length) pages.appendChild(el('div', 'sui-rich-preview-loading', 'No content to preview.'));
      if (autoPrint) {
        var autoClose = function () {
          window.removeEventListener('afterprint', autoClose);
          setTimeout(close, 250);
        };
        window.addEventListener('afterprint', autoClose);
        setTimeout(function () { try { printBtn.click(); } catch (e) {} }, 150);
      }
    } catch (e) {
      pages.innerHTML = '';
      pages.appendChild(el('div', 'sui-rich-preview-loading', 'Preview error: ' + String(e && e.message ? e.message : e)));
    } finally {
      if (!wasPaginated && inst.root && inst.root.getAttribute('data-sui-rich-paginated') === '1') setPaginated(inst, false);
    }
  }
  function promptRange(inst) {
    var sel = window.getSelection();
    if (!sel || !sel.rangeCount || !inst.editor || !inst.editor.contains(sel.anchorNode)) return null;
    return sel.getRangeAt(0).cloneRange();
  }
  function restorePromptRange(inst, range) {
    if (!inst.editor) return;
    inst.editor.focus();
    if (!range || !inst.editor.contains(range.startContainer)) return;
    var sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
  }
  function closeRichPrompt() {
    var existing = document.querySelector('.sui-rich-prompt');
    if (existing && existing.parentNode) existing.parentNode.removeChild(existing);
  }
  function richPrompt(inst, opts) {
    closeRichPrompt();
    var overlay = el('div', 'sui-rich-prompt');
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-modal', 'true');
    var card = el('div', 'sui-rich-prompt-card');
    if (inst && inst.root) {
      var rootStyle = getComputedStyle(inst.root);
      ['--sui-rich-bg', '--sui-rich-fg', '--sui-rich-border', '--sui-rich-muted', '--sui-rich-hover', '--sui-rich-accent'].forEach(function (name) {
        var value = rootStyle.getPropertyValue(name);
        if (value) card.style.setProperty(name, value);
      });
    }
    card.title = opts.instructions || '';
    overlay.title = opts.instructions || '';
    overlay.setAttribute('aria-label', opts.instructions || 'Dialog');
    var inputs = {};
    var first = null;
    (opts.fields || []).forEach(function (field) {
      var label = el('label', 'sui-rich-prompt-field');
      label.appendChild(el('span', '', field.label));
      var input = field.multiline ? document.createElement('textarea') : document.createElement('input');
      if (field.multiline) input.rows = field.rows || 8;
      else input.type = field.type || 'text';
      if (field.type === 'file' && field.accept) input.accept = field.accept;
      if (field.type !== 'file') input.value = field.value == null ? '' : String(field.value);
      if (field.placeholder) input.placeholder = field.placeholder;
      if (field.maxLength) input.maxLength = field.maxLength;
      if (field.autocomplete) input.autocomplete = field.autocomplete;
      label.appendChild(input);
      card.appendChild(label);
      inputs[field.key] = input;
      if (!first) first = input;
    });
    var error = el('p', 'sui-rich-prompt-error');
    error.hidden = true;
    card.appendChild(error);
    var actions = el('div', 'sui-rich-prompt-actions');
    var cancel = el('button', 'sui-rich-btn', 'Cancel');
    cancel.type = 'button';
    var submitBtn = el('button', 'sui-rich-btn', opts.submitLabel || 'OK');
    submitBtn.type = 'button';
    actions.appendChild(cancel); actions.appendChild(submitBtn);
    card.appendChild(actions);
    overlay.appendChild(card);
    document.body.appendChild(overlay);

    function close() {
      if (overlay.parentNode) overlay.parentNode.removeChild(overlay);
      document.removeEventListener('keydown', onKey);
      if (inst && inst.editor) inst.editor.focus();
    }
    function submit() {
      var values = {};
      for (var key in inputs) if (Object.prototype.hasOwnProperty.call(inputs, key)) {
        values[key] = inputs[key].type === 'file' ? (inputs[key].files && inputs[key].files[0]) : inputs[key].value;
      }
      var message = null;
      try { message = opts.onSubmit ? opts.onSubmit(values) : null; }
      catch (e) { message = String(e && e.message ? e.message : e); }
      if (message) { error.textContent = String(message); error.hidden = false; return; }
      close();
    }
    function onKey(ev) { if (ev.key === 'Escape') { ev.preventDefault(); close(); } }
    document.addEventListener('keydown', onKey);
    cancel.addEventListener('click', close);
    submitBtn.addEventListener('click', submit);
    overlay.addEventListener('mousedown', function (ev) { if (ev.target === overlay) close(); });
    card.addEventListener('keydown', function (ev) {
      if (ev.key === 'Enter' && !ev.shiftKey && ev.target && ev.target.tagName !== 'TEXTAREA') { ev.preventDefault(); submit(); }
    });
    if (first) { first.focus(); if (first.select) first.select(); }
  }
  function printDocument(inst) {
    // Print the same paginated sheets used by the preview, so headers and
    // footers keep their real page numbers instead of being stripped.
    openPrintPreview(inst, true);
  }

  function buildFindBar(inst) {
    var body = inst.root.querySelector('.sui-rich-body');
    if (!body || body.querySelector('.sui-rich-findbar')) return;
    var bar = el('div', 'sui-rich-findbar');
    bar.hidden = true;
    var query = el('input', 'sui-rich-find-input');
    query.type = 'search'; query.placeholder = 'Find'; query.setAttribute('data-rich-find', 'query');
    query.setAttribute('aria-label', 'Find text');
    var repl = el('input', 'sui-rich-find-input');
    repl.type = 'text'; repl.placeholder = 'Replace'; repl.setAttribute('data-rich-find', 'replacement');
    function action(name, label, title) {
      var b = el('button', 'sui-rich-find-btn', label);
      b.type = 'button'; b.title = title; b.setAttribute('data-rich-find-action', name);
      b.setAttribute('aria-label', title);
      return b;
    }
    var prev = action('prev', '\u2191', 'Previous match');
    var next = action('next', '\u2193', 'Next match');
    var one = action('replace', 'Replace', 'Replace current match');
    var all = action('replaceAll', 'All', 'Replace all matches');
    var close = action('close', '\u00d7', 'Close find bar');
    var count = el('span', 'sui-rich-find-count', '0/0');
    bar.appendChild(query); bar.appendChild(count); bar.appendChild(prev); bar.appendChild(next); bar.appendChild(repl); bar.appendChild(one); bar.appendChild(all); bar.appendChild(close);
    body.appendChild(bar);
    query.addEventListener('input', function () { findInDocument(inst, query.value); });
    query.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') { e.preventDefault(); if (inst.find && inst.find.query === query.value && inst.find.matches.length) nextMatch(inst, e.shiftKey ? -1 : 1); else findInDocument(inst, query.value); }
      else if (e.key === 'Escape') { e.preventDefault(); closeFindBar(inst); }
    });
    repl.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') { e.preventDefault(); replaceMatch(inst, repl.value); }
      else if (e.key === 'Escape') { e.preventDefault(); closeFindBar(inst); }
    });
    bar.addEventListener('click', function (e) {
      var target = e.target && e.target.closest ? e.target.closest('[data-rich-find-action]') : null;
      if (!target) return;
      var act = target.getAttribute('data-rich-find-action');
      if (act === 'prev') nextMatch(inst, -1);
      else if (act === 'next') nextMatch(inst, 1);
      else if (act === 'replace') replaceMatch(inst, repl.value);
      else if (act === 'replaceAll') replaceAllMatches(inst, repl.value);
      else if (act === 'close') closeFindBar(inst);
    });
  }

  function bindEditorSurface(inst) {
    var editor = inst.editor;
    if (!editor || editor.__suiRichSurfaceBound) return;
    editor.__suiRichSurfaceBound = true;
    editor.addEventListener('scroll', function () {
      // A replaced editor can report scrollTop 0 while it is detached; ignore
      // it so it cannot overwrite the scroll position kept for the live editor.
      if (inst.editor !== editor || !editor.isConnected) return;
      // A fresh editor reports scrollTop 0 before the stored position is
      // restored; do not let that transient value erase it.
      if (editor.scrollTop === 0 && (inst.keepScrollTop || inst.scrollTop) > 0 && Date.now() < (inst.scrollGuardUntil || 0)) return;
      inst.scrollTop = editor.scrollTop;
      inst.keepScrollTop = inst.scrollTop;
      positionPageGuides(inst);
    });
    editor.addEventListener('blur', function () {
      setTimeout(function () {
        // Clicking a guide/page-break line blurs the editor. Do not commit and
        // re-render the component in the middle of the drag: it would detach
        // the line being dragged and reset its position.
        if (inst.guideDragging) return;
        commit(inst, 'change');
      }, 0);
    });
    editor.addEventListener('copy', function (e) {
      if (!inst.selectedCells || inst.selectedCells.length < 2) return;
      var text = copyCellSelection(inst, false);
      if (!text) return;
      e.preventDefault();
      if (e.clipboardData) { e.clipboardData.setData('text/plain', text); }
    });
    editor.addEventListener('cut', function (e) {
      if (!inst.selectedCells || inst.selectedCells.length < 2) return;
      var text = copyCellSelection(inst, true);
      if (!text) return;
      e.preventDefault();
      if (e.clipboardData) { e.clipboardData.setData('text/plain', text); }
      commit(inst, 'change');
    });
    editor.addEventListener('paste', function (e) {
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
        insertDocumentAtCaret(inst, importHTML(html));
        materializeBlobImages(inst, inst.editor);
        commit(inst, 'input');
      } else setTimeout(function () { commit(inst, 'input'); }, 0);
    });
    bindCellSelection(inst); bindTableResize(inst);
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
      var themeToggle = target.closest('[data-rich-theme-toggle]');
      if (themeToggle) {
        var resolvedTheme = inst.theme === 'auto' ? (inst.root.getAttribute('data-sui-rich-theme') || 'dark') : (inst.theme || 'dark');
        setTheme(inst.name, resolvedTheme === 'light' ? 'dark' : 'light');
        return;
      }
      var pagesToggle = target.closest('[data-rich-pages]');
      if (pagesToggle) { setPaginated(inst); return; }
      var zoom = target.closest('[data-rich-zoom]');
      if (zoom) {
        var zoomAction = zoom.getAttribute('data-rich-zoom');
        setZoom(inst, zoomAction === 'reset' ? 1 : (inst.zoom || 1) + (zoomAction === 'in' ? 0.1 : -0.1));
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
      if (target.getAttribute('data-rich-theme') !== null) { setTheme(inst.name, target.value); return; }
      var cmd = target.getAttribute('data-rich-cmd');
      if (!cmd) return;
      if (target.tagName === 'SELECT' || target.tagName === 'INPUT') {
        if ((cmd === 'textColor' || cmd === 'highlight') && !inst.colorSelection && inst.colorCommittedValue === target.value) return;
        var changeValue = target.value;
        command(inst.name, cmd, changeValue);
        if (cmd === 'textColor' || cmd === 'highlight') {
          inst.colorSelection = null;
          inst.colorPreview = null;
          inst.colorCommittedValue = changeValue;
        }
      }
    });
    root.addEventListener('mouseup', function (e) {
      if (!inst.painter || !e.target || !inst.editor.contains(e.target)) return;
      setTimeout(function () { if (inst.painter) applyFormatPainter(inst); }, 0);
    });
    root.addEventListener('keyup', function (e) {
      if (!inst.painter || !inst.editor.contains(e.target)) return;
      var sel = window.getSelection();
      if (sel && sel.rangeCount && !sel.isCollapsed) applyFormatPainter(inst);
    });
    root.addEventListener('input', function (e) {
      if (!inst.editor.contains(e.target)) return;
      clearTimeout(inst.inputTimer);
      inst.inputTimer = setTimeout(function () { saveSelection(inst); commit(inst, 'input'); }, 250);
    });
    root.addEventListener('beforeinput', function (e) {
      if (!inst.editor || !inst.editor.contains(e.target)) return;
      if (e.inputType !== 'insertText' && e.inputType !== 'insertReplacementText') return;
      var inputSel = window.getSelection();
      if (!inputSel || !inputSel.rangeCount || !inputSel.isCollapsed) return;
      applyTypingState(inst);
      if (inst.typingMarks && inst.typingMarks.code && e.data && !closestIn(inst, 'code')) {
        e.preventDefault();
        insertTypedCode(inst, e.data);
      }
    });
    root.addEventListener('keydown', function (e) {
      if (!inst.editor.contains(e.target)) return;
      // Escape leaves a quote/code block; Backspace at its very start does
      // the same, matching common word processors.
      if (e.key === 'Escape' && inst.painter) {
        e.preventDefault();
        cancelFormatPainter(inst);
        return;
      }
      if (e.key === 'Escape' || e.key === 'Backspace') {
        var escapeMode = blockKind(inst);
        if (escapeMode === 'blockquote' || escapeMode === 'pre') {
          var escapeBlock = currentBlock(inst);
          var escapeQuote = quoteAncestor(escapeBlock, inst.editor);
          var atStart = e.key === 'Escape';
          if (!atStart) {
            var escapeSelection = window.getSelection();
            if (escapeSelection && escapeSelection.isCollapsed && escapeSelection.rangeCount) {
              var escapeRange = escapeSelection.getRangeAt(0);
              var atOffsetZero = textOffsetWithin(escapeBlock, escapeRange.startContainer, escapeRange.startOffset) === 0;
              atStart = atOffsetZero && (escapeMode === 'pre' || !escapeQuote || escapeBlock === escapeQuote.firstElementChild);
            }
          }
          if (atStart) {
            e.preventDefault();
            applyBlock(inst, 'p');
            saveSelection(inst);
            commit(inst, 'change');
            return;
          }
        }
      }
      if (e.key === 'Enter' && !e.shiftKey && !e.ctrlKey && !e.metaKey) {
        e.preventDefault();
        splitBlock(inst);
        saveSelection(inst);
        commit(inst, 'input');
        return;
      }
      if (e.key === 'Enter' && e.shiftKey) {
        e.preventDefault();
        try { document.execCommand('insertLineBreak'); } catch (err) { insertNode(inst, el('br')); }
        saveSelection(inst);
        commit(inst, 'input');
        return;
      }
      if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) { e.preventDefault(); command(inst.name, 'pageBreak'); return; }
      if (e.key === 'Tab') {
        var tabBlock = currentBlock(inst);
        if (tabBlock && tabBlock.tagName && tabBlock.tagName.toLowerCase() === 'li') {
          e.preventDefault();
          if (e.shiftKey) outdentListItem(tabBlock); else indentListItem(tabBlock);
          commit(inst, 'change');
          return;
        }
      }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') { e.preventDefault(); command(inst.name, 'find'); return; }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'p') { e.preventDefault(); command(inst.name, 'print'); return; }
      if (e.key === 'F3') { e.preventDefault(); var fbar = inst.root.querySelector('.sui-rich-findbar'); if (!fbar || fbar.hidden) command(inst.name, 'find'); else command(inst.name, e.shiftKey ? 'findPrev' : 'findNext'); return; }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'z') { e.preventDefault(); if (e.shiftKey) redo(inst); else undo(inst); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'y') { e.preventDefault(); redo(inst); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') { e.preventDefault(); command(inst.name, 'bold'); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'i') { e.preventDefault(); command(inst.name, 'italic'); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'u') { e.preventDefault(); command(inst.name, 'underline'); }
      else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') { e.preventDefault(); commit(inst, 'change'); }
    });
    if (!inst.menuOutsideBound) {
      inst.menuOutsideBound = true;
      function closeMenusOutside(ev) {
        if (!inst.root || !inst.root.contains(ev.target) || !ev.target.closest('.sui-rich-menu-wrap')) closeMenus(inst);
      }
      document.addEventListener('mousedown', closeMenusOutside);
      document.addEventListener('click', closeMenusOutside);
      document.addEventListener('keydown', function (ev) { if (ev.key === 'Escape') closeMenus(inst); });
    }
    document.addEventListener('selectionchange', function () { if (document.activeElement === inst.editor || inst.root.contains(document.activeElement)) updateStats(inst); });
    if (!window.__suiRichColorOutsideBound) {
      window.__suiRichColorOutsideBound = true;
      document.addEventListener('mousedown', function (ev) {
        for (var colorName in instances) {
          if (!hasOwn(instances, colorName)) continue;
          var colorInst = instances[colorName];
          if (!colorInst || !colorInst.colorSelection) continue;
          var colorInput = colorInst.colorSelection.input;
          var colorWrap = colorInput && colorInput.closest ? colorInput.closest('.sui-rich-color-control') : null;
          if (colorWrap && (colorWrap === ev.target || colorWrap.contains(ev.target))) continue;
          (function (pendingInst, pendingSelection) {
            setTimeout(function () {
              if (pendingInst.colorSelection === pendingSelection) finalizeColor(pendingInst);
            }, 0);
          })(colorInst, colorInst.colorSelection);
        }
      }, true);
    }
    if (!window.__suiRichResizeBound) {
      window.__suiRichResizeBound = true;
      window.addEventListener('resize', function () {
        for (var rn in instances) if (instances[rn]) {
          if (instances[rn].paginated) schedulePaginate(instances[rn]);
          updatePageGuides(instances[rn]);
          schedulePageGuides(instances[rn]);
        }
      });
    }
    if (!window.__suiRichScrollBound) {
      window.__suiRichScrollBound = true;
      window.addEventListener('scroll', function () {
        for (var sn in instances) if (instances[sn]) positionPageGuides(instances[sn]);
      }, true);
    }
  }

  function applyEditorScroll(inst, value) {
    if (!inst || !inst.editor || !(value > 0)) return;
    if (inst.scrollRestoreTimer) { clearTimeout(inst.scrollRestoreTimer); inst.scrollRestoreTimer = null; }
    inst.editor.scrollTop = value;
    var applied = inst.editor.scrollTop;
    if (applied === 0 && value > 0 && inst.editor.scrollHeight > inst.editor.clientHeight) {
      // The editor may not have its final height yet in the same frame; retry
      // once the layout is settled instead of losing the stored position.
      inst.scrollRestoreTimer = setTimeout(function () {
        inst.scrollRestoreTimer = null;
        if (!inst.editor) return;
        inst.editor.scrollTop = value;
        inst.scrollTop = inst.keepScrollTop = inst.editor.scrollTop;
        positionPageGuides(inst);
      }, 0);
      return;
    }
    inst.scrollTop = inst.keepScrollTop = applied;
    positionPageGuides(inst);
  }
  function init(root) {
    var name = attr(root, 'data-sui-rich-name') || 'richtext';
    var earlyInst = instances[name];
    if (earlyInst) earlyInst.scrollGuardUntil = Date.now() + 300;
    var keepScroll = earlyInst ? (earlyInst.keepScrollTop || earlyInst.scrollTop || 0) : 0;
    var input = root.querySelector('.sui-rich-value');
    var editor = root.querySelector('.sui-rich-editor');
    if (!editor) return;
    if (!editor.getAttribute('role')) editor.setAttribute('role', 'textbox');
    if (!editor.getAttribute('aria-multiline')) editor.setAttribute('aria-multiline', 'true');
    if (!editor.getAttribute('aria-label')) editor.setAttribute('aria-label', attr(root, 'data-sui-rich-title') || 'Rich text editor');
    if (boolAttr(root, 'data-sui-rich-readonly', false)) editor.setAttribute('aria-readonly', 'true');
    else editor.setAttribute('spellcheck', 'true');
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
        headerAttr: attr(root, 'data-sui-rich-header'),
        footerAttr: attr(root, 'data-sui-rich-footer'),
        headerText: attr(root, 'data-sui-rich-header') || ((initial && initial.header) ? String(initial.header) : ''),
        footerText: attr(root, 'data-sui-rich-footer') || ((initial && initial.footer) ? String(initial.footer) : ''),
        headerOverride: false, footerOverride: false,
        headerValue: '', footerValue: '', headerH: 0, footerH: 0, title: '',
        history: [], index: -1, dirty: false, zoom: 1, scrollTop: 0, keepScrollTop: 0,
        typingMarks: {}, typingColor: '', typingHighlight: '',
        baseFontSize: parseFloat(attr(root, 'data-sui-rich-font-size') || '16') || 16
      };
      inst.history.push(JSON.stringify(inst.doc));
      inst.index = 0;
      instances[name] = inst;
    } else {
      inst.root = root; inst.editor = editor;
    }
    if (!inst.typingMarks) inst.typingMarks = {};
    editor.setAttribute('lang', documentLanguage(inst));
    hyphenationLanguage = documentLanguage(inst);
    var chromeTitle = root.querySelector('.sui-rich-title');
    inst.title = chromeTitle ? String(chromeTitle.textContent || '').trim() : (attr(root, 'data-sui-rich-title') || name);
    refreshChromeValues(inst);
    if (!inst.editor.__suiRichRendered || inst.editor.__suiRichDoc !== JSON.stringify(inst.doc)) {
      inst.editor.innerHTML = '';
      inst.editor.appendChild(renderDoc(inst.doc));
      inst.editor.__suiRichDoc = JSON.stringify(inst.doc);
    }
    inst.editor.__suiRichRendered = true;
    if (input) input.value = JSON.stringify(inst.doc);
    var editorHeight = cssLength(attr(root, 'data-sui-rich-height'), '320px');
    root.style.setProperty('--sui-rich-content-height', editorHeight);
    inst.editor.style.height = editorHeight;
    inst.editor.style.minHeight = '';
    applyPageLineHeight(inst);
    var initialTemplate = attr(root, 'data-sui-rich-template');
    if (initialTemplate) setTemplate(name, initialTemplate);
    var readOnly = boolAttr(root, 'data-sui-rich-readonly', false);
    inst.editor.setAttribute('contenteditable', readOnly ? 'false' : 'true');
    if (attr(root, 'data-sui-rich-placeholder')) inst.editor.setAttribute('data-placeholder', attr(root, 'data-sui-rich-placeholder'));
    buildToolbar(inst); buildMenu(inst); buildContext(inst); buildStatus(inst); buildFindBar(inst); buildPageChrome(inst);
    if (attr(root, 'data-sui-rich-paginated') === '1') setPaginated(inst, true);
    var savedTheme = '';
    try { savedTheme = localStorage.getItem('sui-rich-theme-' + name) || ''; } catch (e) {}
    var initialTheme = savedTheme || attr(root, 'data-sui-rich-theme') || 'dark';
    if (initialTheme !== 'dark' && initialTheme !== 'light' && initialTheme !== 'auto') initialTheme = 'dark';
    inst.theme = initialTheme;
    applyTheme(inst);
    if (!inst.themeMediaBound && window.matchMedia) {
      inst.themeMediaBound = true;
      try {
        var themeMedia = window.matchMedia('(prefers-color-scheme: light)');
        if (themeMedia.addEventListener) themeMedia.addEventListener('change', function () { if (inst.theme === 'auto') applyTheme(inst); });
      } catch (e) {}
    }
    var savedIconSet = '';
    try { savedIconSet = localStorage.getItem('sui-rich-icons-' + name) || ''; } catch (e) {}
    inst.iconSet = validIconSet(savedIconSet || attr(root, 'data-sui-rich-icons') || 'line');
    applyIconSet(inst);
    var toolbarEl = root.querySelector('.sui-rich-toolbar');
    if (toolbarEl && !boolAttr(root, 'data-sui-rich-toolbar', true)) toolbarEl.hidden = true;
    var menuEl = root.querySelector('.sui-rich-menu');
    if (menuEl && !boolAttr(root, 'data-sui-rich-menu', true)) menuEl.hidden = true;
    var statusEl = root.querySelector('.sui-rich-status');
    if (statusEl && !boolAttr(root, 'data-sui-rich-statusbar', true)) statusEl.hidden = true;
    bindEditorSurface(inst); bindRoot(inst); bindProperties(inst);
    setZoom(inst, inst.zoom || 1);
    updateStats(inst);
    updatePageGuides(inst);
    schedulePageGuides(inst);
    restoreSelection(inst);
    // Restoring the caret can scroll the editor; put the stored position back
    // last so a re-render never jumps the document to the caret.
    applyEditorScroll(inst, keepScroll);
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
  function base64FromBytes(bytes) {
    var binary = '', chunk = 8192;
    for (var i = 0; i < bytes.length; i += chunk) binary += String.fromCharCode.apply(null, bytes.slice(i, i + chunk));
    return btoa(binary);
  }
  function mimeFromName(name) {
    var ext = String(name || '').toLowerCase().split('.').pop();
    if (ext === 'png') return 'image/png';
    if (ext === 'jpg' || ext === 'jpeg') return 'image/jpeg';
    if (ext === 'gif') return 'image/gif';
    if (ext === 'webp') return 'image/webp';
    if (ext === 'svg') return 'image/svg+xml';
    return '';
  }
  function dataUrlFromBytes(bytes, name) {
    var mime = mimeFromName(name);
    if (!mime) return '';
    return 'data:' + mime + ';base64,' + base64FromBytes(bytes);
  }
  function bytesFromDataUrl(src) {
    var match = /^data:(image\/(?:png|jpe?g|gif|webp|svg\+xml));base64,([\s\S]*)$/i.exec(String(src || ''));
    if (!match) return null;
    try {
      var binary = atob(match[2]), bytes = new Uint8Array(binary.length);
      for (var i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
      return { bytes: bytes, mime: match[1], ext: match[1] === 'image/jpeg' ? 'jpg' : match[1] === 'image/svg+xml' ? 'svg' : match[1].split('/')[1] };
    } catch (e) { return null; }
  }
  function extensionForMime(mime) {
    if (mime === 'image/png') return 'png';
    if (mime === 'image/jpeg') return 'jpg';
    if (mime === 'image/gif') return 'gif';
    if (mime === 'image/webp') return 'webp';
    if (mime === 'image/svg+xml') return 'svg';
    return 'png';
  }
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
      else if (n.type === 'page_break') out.push('');
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
    var colorTable = [''];
    var colorGroup = /\\colortbl([^}]*)/i.exec(source);
    if (colorGroup) {
      var colorRe = /\\red(\d+)\\green(\d+)\\blue(\d+)/g, cm;
      while ((cm = colorRe.exec(colorGroup[1]))) colorTable.push('#' + [cm[1], cm[2], cm[3]].map(function (v) { return ('0' + Math.max(0, Math.min(255, Number(v))).toString(16)).slice(-2); }).join(''));
    }
    var fontTable = [];
    var fontGroup = /\\fonttbl(?![a-zA-Z])/i.exec(source);
    if (fontGroup) {
      var fontOpen = source.lastIndexOf('{', fontGroup.index);
      if (fontOpen >= 0) {
        var fontBlock = source.slice(fontOpen, groupSpanFrom(source, fontOpen));
        var fontRe = /\\f(\d+)((?:[^{}]|\{[^{}]*\})*);/g, fontMatch;
        while ((fontMatch = fontRe.exec(fontBlock))) {
          var fontName = fontMatch[2].replace(/\{\\\*\\falt[^}]*\}/gi, '').replace(/\{\\\*[^}]*\}/g, '').replace(/\\[a-zA-Z]+-?\d*\s?/g, '').replace(/[;{}]/g, '').trim();
          fontTable[Number(fontMatch[1])] = fontName;
        }
      }
    }
    var SYMBOL_PUA = { 0xF020: '\u2610', 0xF0A8: '\u2610', 0xF0FE: '\u2611', 0xF0FD: '\u2612', 0xF0FC: '\u2713', 0xF0B7: '\u2022' };
    function mapSymbols(value) {
      return String(value).replace(/[\uF020-\uF0FF]/g, function (ch) {
        return SYMBOL_PUA[ch.charCodeAt(0)] || ch;
      });
    }
    function fontStack(name) {
      var n = String(name || '').trim();
      var lower = n.toLowerCase();
      var list;
      if (!n || /symbol|wingdings|webdings|dingbat|mtextra/.test(lower)) return '';
      if (/courier|mono|consolas|lucida console/.test(lower)) list = [n, 'Consolas', 'monospace'];
      else if (/sans|arial|helvetica|verdana|tahoma|segoe|calibri|aptos|futura|gill/.test(lower)) list = [n, 'Arial', 'Helvetica', 'sans-serif'];
      else if (/times|serif|georgia|garamond|cambria|book|roman|palatino/.test(lower)) list = [n, 'Georgia', 'Times New Roman', 'serif'];
      else list = [n, 'sans-serif'];
      var seen = {};
      return list.filter(function (f) {
        var key = f.toLowerCase();
        if (seen[key]) return false;
        seen[key] = true;
        return true;
      }).join(', ');
    }
    var state = { bold: false, italic: false, underline: false, strike: false, sup: false, sub: false, color: '', highlight: '', size: '', font: '' };
    var stack = [];
    var i = 0, depth = 0, skipDepth = -1, ucSkip = 1;
    var current = [], blocks = [], buffer = '', align = '';
    var paraIndent = '', paraSpaceBefore = '', paraSpaceAfter = '', paraLineHeight = '', paraStyle = 0;
    var tableState = null, rowCells = null, cellInline = null;
    var cellXs = [], cellProps = [], pendingVmerge = '', pendingCellBg = '', rowHeight = '', cellAlign = '';
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
      if (state.font) marks.push({ type: 'font', value: state.font });
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
        var heading = paraStyle === 28 ? 1 : 0;
        var p = { type: heading ? 'heading' : 'paragraph', children: current };
        if (heading) p.level = heading;
        if (align) p.align = align;
        if (paraIndent) p.indent = paraIndent;
        if (paraSpaceBefore) p.spaceBefore = paraSpaceBefore;
        if (paraSpaceAfter) p.spaceAfter = paraSpaceAfter;
        if (paraLineHeight) p.lineHeight = paraLineHeight;
        if (tableState && rowCells && cellInline) {
          if (align) cellAlign = align;
          if (cellInline.length) cellInline.push({ type: 'break', marks: [] });
          for (var ci = 0; ci < current.length; ci++) cellInline.push(current[ci]);
        } else {
          if (tableState) tableState = null;
          blocks.push(p);
        }
        current = [];
      }
      align = ''; paraIndent = ''; paraSpaceBefore = ''; paraSpaceAfter = ''; paraLineHeight = ''; paraStyle = 0;
    }
    function addText(value) { if (skipDepth < 0) buffer += mapSymbols(value); }
    function addBreak() { flushText(); if (skipDepth < 0) current.push({ type: 'break', marks: [] }); }
    function readPictureGroup() {
      var start = i, level = 0;
      while (i < source.length) {
        if (source.charAt(i) === '{') level++;
        else if (source.charAt(i) === '}') {
          if (level === 0) { var text = source.slice(start, i); i++; return text; }
          level--;
        }
        i++;
      }
      return source.slice(start);
    }
    function pictureNode(group) {
      var mime = /\\pngblip/.test(group) ? 'image/png' : /\\jpegblip/.test(group) ? 'image/jpeg' : '';
      if (!mime) return null;
      var blip = /\\pngblip|\\jpegblip/i.exec(group);
      var data = blip ? group.slice(blip.index + blip[0].length) : group;
      var cleaned = data.replace(/[^0-9a-fA-F]/g, '');
      if (cleaned.length < 4) return null;
      if (cleaned.length % 2) cleaned = cleaned.slice(0, cleaned.length - 1);
      var bytes = new Uint8Array(cleaned.length / 2);
      for (var n = 0; n < bytes.length; n++) bytes[n] = parseInt(cleaned.slice(n * 2, n * 2 + 2), 16);
      var wm = /\\picwgoal(\d+)/.exec(group) || /\\picw(\d+)/.exec(group), hm = /\\pichgoal(\d+)/.exec(group) || /\\pich(\d+)/.exec(group);
      var sx = /\\picscalex(\d+)/.exec(group);
      var width = wm ? Number(wm[1]) / 15 : 0, height = hm ? Number(hm[1]) / 15 : 0;
      if (sx && Number(sx[1]) > 0 && Number(sx[1]) < 100) { width = width * Number(sx[1]) / 100; if (height > 0) height = height * Number(sx[1]) / 100; }
      var maxWidth = 680;
      if (width > maxWidth) { if (height > 0) height = height * maxWidth / width; width = maxWidth; }
      return { type: 'image', src: dataUrlFromBytes(bytes, mime === 'image/jpeg' ? 'image.jpg' : 'image.png'), alt: '', width: width > 0 ? Math.round(width) + 'px' : '', height: height > 0 ? Math.round(height) + 'px' : '' };
    }
    function groupSpanFrom(text, start) {
      var level = 0, j = start;
      while (j < text.length) {
        var c = text.charAt(j);
        if (c === '\\') { j += 2; continue; }
        if (c === '{') level++;
        else if (c === '}') { level--; if (level === 0) return j + 1; }
        j++;
      }
      return text.length;
    }
    function destinationPictures(dest) {
      var out = [];
      var re = new RegExp('\\\\' + dest + '(?![a-zA-Z])', 'i');
      var m = re.exec(source);
      if (!m) return out;
      var open = source.lastIndexOf('{', m.index);
      if (open < 0) return out;
      var body = source.slice(open, groupSpanFrom(source, open));
      var at = 0;
      while (at < body.length) {
        var p = body.indexOf('\\pict', at);
        if (p < 0) break;
        var gs = body.lastIndexOf('{', p);
        if (gs < 0) break;
        var ge = groupSpanFrom(body, gs);
        var node = pictureNode(body.slice(gs, ge));
        if (node) out.push(node);
        at = ge;
      }
      return out;
    }
    function finalizeTable(table) {
      var rows = table.rows || [];
      if (!rows.length) return;
      var edges = [];
      for (var r = 0; r < rows.length; r++) {
        var bounds = rows[r]._bounds || [];
        for (var b = 0; b < bounds.length; b++) if (edges.indexOf(bounds[b]) < 0) edges.push(bounds[b]);
      }
      if (!edges.length) {
        for (var r0 = 0; r0 < rows.length; r0++) {
          delete rows[r0]._bounds;
          (rows[r0].cells || []).forEach(function (cell) { delete cell._vmerge; });
        }
        return;
      }
      edges.sort(function (a, b) { return a - b; });
      var starts = [0].concat(edges);
      var total = edges[edges.length - 1];
      function gridIndex(x) {
        var best = 0, dist = Infinity;
        for (var s = 0; s < starts.length; s++) {
          var d = Math.abs(starts[s] - x);
          if (d < dist) { dist = d; best = s; }
        }
        return best;
      }
      var layout = [];
      for (var r = 0; r < rows.length; r++) {
        var bounds = rows[r]._bounds || [];
        var cells = rows[r].cells || [];
        var placed = [];
        for (var c = 0; c < cells.length; c++) {
          var x1 = bounds[c] != null ? bounds[c] : (c === cells.length - 1 ? total : (bounds[c - 1] || 0));
          var x0 = c === 0 ? 0 : (bounds[c - 1] != null ? bounds[c - 1] : 0);
          var si = gridIndex(x0), ei = gridIndex(x1);
          if (ei <= si) ei = si + 1;
          placed.push({ cell: cells[c], start: si, end: ei, x0: x0, x1: x1 });
        }
        for (var p = placed.length - 1; p >= 0; p--) {
          if (placed[p].cell._vmerge !== 'continue') continue;
          var target = null;
          for (var pr = r - 1; pr >= 0 && !target; pr--) {
            var above = layout[pr] || [];
            for (var q = 0; q < above.length; q++) {
              if (above[q].start === placed[p].start) { target = above[q]; break; }
            }
          }
          if (target) target.cell.rowspan = (Number(target.cell.rowspan) || 1) + 1;
          placed.splice(p, 1);
        }
        layout[r] = placed;
        var kept = [];
        for (var k = 0; k < placed.length; k++) {
          var info = placed[k];
          var cell = info.cell;
          cell.colspan = Math.max(1, info.end - info.start);
          cell.width = Math.round(((info.x1 - info.x0) / total) * 1000) / 10 + '%';
          delete cell._vmerge;
          kept.push(cell);
        }
        rows[r].cells = kept;
        delete rows[r]._bounds;
      }
    }
    while (i < source.length) {
      var ch = source.charAt(i);
      if (ch === '{') {
        depth++;
        stack.push({ bold: state.bold, italic: state.italic, underline: state.underline, strike: state.strike, sup: state.sup, sub: state.sub, color: state.color, highlight: state.highlight, size: state.size, font: state.font });
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
          if (skipDepth >= 0) { i++; continue; }
          var starAt = i + 1;
          while (starAt < source.length && /\s/.test(source.charAt(starAt))) starAt++;
          var starWord = '';
          while (starAt < source.length && /[a-zA-Z]/.test(source.charAt(starAt))) { starWord += source.charAt(starAt); starAt++; }
          if (starWord === 'shppict' || starWord === 'nonshppict') { i++; continue; }
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
          var negativeNum = source.charAt(i) === '-';
          if (negativeNum) i++;
          while (i < source.length && /[0-9]/.test(source.charAt(i))) { numText += source.charAt(i); i++; }
          numText = String((negativeNum ? -1 : 1) * Number(numText || '0'));
        }
        var hasNum = numText !== '';
        var num = hasNum ? Number(numText) : 0;
        if (source.charAt(i) === ' ') i++;
        if (word === 'pict') {
          if (skipDepth >= 0) continue;
          flushText();
          var picture = pictureNode(readPictureGroup());
          if (depth > 0) { depth--; if (stack.length) state = stack.pop(); }
          if (picture) current.push(picture);
          continue;
        }
        if (word === 'fonttbl' || word === 'colortbl' || word === 'stylesheet' || word === 'info' || word === 'header' || word === 'footer' || word === 'object' || word === 'themedata' || word === 'datastore' || word === 'latentstyles') {
          if (skipDepth < 0) skipDepth = depth;
          continue;
        }
        if (skipDepth >= 0) continue;
        if (word === 'trowd') {
          if (!tableState) { tableState = { type: 'table', rows: [], align: '' }; blocks.push(tableState); }
          rowCells = []; cellInline = []; current = [];
          cellXs = []; cellProps = []; pendingVmerge = ''; pendingCellBg = ''; rowHeight = ''; cellAlign = '';
          align = ''; paraIndent = ''; paraSpaceBefore = ''; paraSpaceAfter = ''; paraLineHeight = ''; paraStyle = 0;
          continue;
        }
        if (word === 'cellx' && hasNum) {
          cellXs.push(num);
          cellProps.push({ vmerge: pendingVmerge, background: pendingCellBg });
          pendingVmerge = '';
          pendingCellBg = '';
          continue;
        }
        if (word === 'clvmgf') { pendingVmerge = 'start'; continue; }
        if (word === 'clvmrg') { pendingVmerge = 'continue'; continue; }
        if (word === 'clcbpat' && hasNum) { pendingCellBg = colorTable[num] || ''; continue; }
        if (word === 'trrh' && hasNum) { rowHeight = Math.round(Math.abs(num) / 15) + 'px'; continue; }
        if (word === 'trqc') { if (tableState) tableState.align = 'center'; continue; }
        if (word === 'trql') { if (tableState) tableState.align = 'left'; continue; }
        if (word === 'trqr') { if (tableState) tableState.align = 'right'; continue; }
        if (word === 'trgaph' || word === 'trleft' || word === 'trpaddb' || word === 'trpaddr' || word === 'trpaddl' || word === 'trpaddt') continue;
        if (word === 'cell') {
          flushBlock();
          rowCells = rowCells || [];
          var cellIndex = rowCells.length;
          var x0 = cellIndex > 0 ? (cellXs[cellIndex - 1] || 0) : 0;
          var x1 = cellXs[cellIndex] != null ? cellXs[cellIndex] : x0;
          var cellWidth = x1 > x0 ? Math.round((x1 - x0) / 15) + 'px' : '';
          var props = cellProps[cellIndex] || {};
          rowCells.push({ type: 'table_cell', header: false, colspan: 1, rowspan: 1, align: cellAlign, width: cellWidth, background: props.background || '', children: cellInline || [], _vmerge: props.vmerge || '' });
          cellInline = []; current = []; align = ''; cellAlign = '';
          continue;
        }
        if (word === 'row') {
          flushBlock();
          if (tableState && rowCells && rowCells.length) {
            tableState.rows.push({ type: 'table_row', height: rowHeight, cells: rowCells, _bounds: cellXs.slice() });
          }
          rowCells = null; cellInline = null; align = ''; cellXs = []; cellProps = []; pendingVmerge = ''; rowHeight = '';
          continue;
        }
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
        else if (word === 'plain') { flushText(); state.bold = state.italic = state.underline = state.strike = state.sup = state.sub = false; state.color = state.highlight = state.size = ''; state.font = ''; }
        else if (word === 'pard') { flushBlock(); }
        else if (word === 'qc' || word === 'center') align = 'center';
        else if (word === 'qr' || word === 'right') align = 'right';
        else if (word === 'ql' || word === 'left' || word === 'qj') align = word === 'qj' ? 'justify' : (word === 'ql' ? 'left' : '');
        else if (word === 'fs' && hasNum) { flushText(); state.size = Math.max(6, Math.round(num / 2)) + 'pt'; }
        else if (word === 'f' && hasNum) { flushText(); state.font = fontStack(fontTable[num] || ''); }
        else if (word === 'li' && hasNum) { flushText(); paraIndent = Math.round(num / 15) + 'px'; }
        else if (word === 'lin' && hasNum) { flushText(); paraIndent = Math.round(num / 15) + 'px'; }
        else if (word === 'sb' && hasNum) { flushText(); paraSpaceBefore = Math.round(num / 20) + 'pt'; }
        else if (word === 'sa' && hasNum) { flushText(); paraSpaceAfter = Math.round(num / 20) + 'pt'; }
        else if (word === 'sl' && hasNum) { flushText(); paraLineHeight = Math.round(num / 20) + 'pt'; }
        else if (word === 's' && hasNum) { paraStyle = num; }
        else if (word === 'outlinelevel' && hasNum) { paraStyle = 28; }
        else if (word === 'cf' && hasNum) { flushText(); state.color = colorTable[num] || ''; }
        else if (word === 'highlight' && hasNum) { flushText(); state.highlight = colorTable[num] || ''; }
        else if (word === 'u' && hasNum) {
          var code = num < 0 ? num + 65536 : num;
          addText(String.fromCharCode(code));
          var skip = ucSkip;
          while (skip > 0 && i < source.length) {
            var ucChar = source.charAt(i);
            if (ucChar === '{' || ucChar === '}') break;
            if (ucChar === '\\') {
              if (source.charAt(i + 1) === "'") { i += 4; skip--; continue; }
              i++;
              while (i < source.length && /[a-zA-Z]/.test(source.charAt(i))) i++;
              if (source.charAt(i) === '-' || /[0-9]/.test(source.charAt(i))) {
                if (source.charAt(i) === '-') i++;
                while (i < source.length && /[0-9]/.test(source.charAt(i))) i++;
              }
              if (source.charAt(i) === ' ') i++;
              skip--;
              continue;
            }
            i++; skip--;
          }
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
    for (var tb = 0; tb < blocks.length; tb++) if (blocks[tb] && blocks[tb].type === 'table') finalizeTable(blocks[tb]);
    blocks = blocks.filter(function (b) { return b.type === 'table' || (b.children && b.children.length && !(b.children.length === 1 && b.children[0].type === 'text' && !b.children[0].text.trim())); });
    var headerPics = destinationPictures('header'), footerPics = destinationPictures('footer');
    if (headerPics.length || footerPics.length) {
      var existing = {};
      function notePictures(node) {
        if (!node || typeof node !== 'object') return;
        if (node.type === 'image' && node.src) existing[node.src] = true;
        if (Array.isArray(node.children)) node.children.forEach(notePictures);
        if (Array.isArray(node.rows)) node.rows.forEach(function (row) {
          if (!row || !Array.isArray(row.cells)) return;
          row.cells.forEach(function (cell) { if (cell && Array.isArray(cell.children)) cell.children.forEach(notePictures); });
        });
      }
      blocks.forEach(notePictures);
      var withPictures = [];
      headerPics.forEach(function (pic) { if (!existing[pic.src]) { pic.align = 'center'; withPictures.push(pic); existing[pic.src] = true; } });
      blocks.forEach(function (b) { withPictures.push(b); });
      footerPics.forEach(function (pic) { if (!existing[pic.src]) { pic.align = 'center'; withPictures.push(pic); existing[pic.src] = true; } });
      blocks = withPictures;
    }
    if (!blocks.length) return importPlain(source);
    return { type: 'doc', children: blocks };
  }
  function hexColor(value) {
    var v = String(value == null ? '' : value).trim();
    var m = /^#([0-9a-f]{6})$/i.exec(v);
    if (m) return m[1].toUpperCase();
    m = /^#([0-9a-f]{3})$/i.exec(v);
    if (m) return (m[1].charAt(0) + m[1].charAt(0) + m[1].charAt(1) + m[1].charAt(1) + m[1].charAt(2) + m[1].charAt(2)).toUpperCase();
    var rgb = /^rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/i.exec(v);
    if (rgb) return [1, 2, 3].map(function (i) { return ('0' + Math.max(0, Math.min(255, Number(rgb[i]))).toString(16)).slice(-2); }).join('').toUpperCase();
    return '';
  }
  function lineHeightInfo(value) {
    var raw = String(value == null ? '' : value).trim();
    var num = parseFloat(raw);
    if (!isFinite(num) || num <= 0) return null;
    if (!/[a-z%]/i.test(raw)) return { multiple: num };
    if (/%/.test(raw)) return { multiple: num / 100 };
    if (/px/i.test(raw)) return { pt: num * 0.75 };
    return { pt: num };
  }
  function cssPt(value) {
    var num = parseFloat(value);
    if (!isFinite(num)) return 0;
    return /px/i.test(String(value)) ? num * 0.75 : num;
  }
  function firstFontFamily(value) {
    var first = String(value || '').split(',')[0].replace(/["']/g, '').trim();
    return /^[A-Za-z0-9 ._-]{1,60}$/.test(first) ? first : '';
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
      else if (t === 'color' && marks[i].value) {
        var ci = rtfColorIndex(marks[i].value);
        if (ci) { open += '\\cf' + ci + ' '; close = '\\cf0 ' + close; }
      } else if (t === 'background' && marks[i].value) {
        var hi = rtfColorIndex(marks[i].value);
        if (hi) { open += '\\highlight' + hi + ' '; close = '\\highlight0 ' + close; }
      } else if (t === 'size' && marks[i].value) {
        var fs = Math.round(cssPt(marks[i].value) * 2);
        if (fs > 0) { open += '\\fs' + fs + ' '; close = '\\fs24 ' + close; }
      } else if (t === 'font' && marks[i].value) {
        var fi = rtfFontIndex(marks[i].value);
        if (fi != null) { open += '\\f' + fi + ' '; close = '\\f0 ' + close; }
      }
    }
    return { open: open, close: close };
  }
  function rtfPicture(src, width, height) {
    var info = bytesFromDataUrl(src);
    if (!info) return '';
    var hex = '';
    for (var i = 0; i < info.bytes.length; i++) hex += ('0' + info.bytes[i].toString(16)).slice(-2);
    var w = Math.round((Number(width) || 160) * 15), h = Math.round((Number(height) || 120) * 15);
    return '{\\pict\\' + (info.mime === 'image/jpeg' ? 'jpegblip' : 'pngblip') + '\\picw' + w + '\\pich' + h + '\\picwgoal' + w + '\\pichgoal' + h + ' ' + hex + '}';
  }
  var rtfFonts = ['Helvetica'], rtfFontMap = {}, rtfColors = [], rtfColorMap = {};
  function rtfFontIndex(value) {
    var name = firstFontFamily(value) || 'Helvetica';
    var key = name.toLowerCase();
    if (rtfFontMap[key] == null) { rtfFontMap[key] = rtfFonts.length; rtfFonts.push(name); }
    return rtfFontMap[key];
  }
  function rtfColorIndex(value) {
    var hex = hexColor(value);
    if (!hex) return 0;
    if (rtfColorMap[hex] == null) { rtfColorMap[hex] = rtfColors.length + 1; rtfColors.push(hex); }
    return rtfColorMap[hex];
  }
  function rtfNoteDoc(node) {
    if (!node || typeof node !== 'object') return;
    if (node.type === 'text') {
      var marks = node.marks || [];
      for (var i = 0; i < marks.length; i++) {
        if (marks[i].type === 'font' && marks[i].value) rtfFontIndex(marks[i].value);
        else if ((marks[i].type === 'color' || marks[i].type === 'background') && marks[i].value) rtfColorIndex(marks[i].value);
      }
    }
    if (node.type === 'table') {
      for (var r = 0; r < (node.rows || []).length; r++) {
        for (var c = 0; c < (node.rows[r].cells || []).length; c++) {
          var cell = node.rows[r].cells[c];
          if (cell.background) rtfColorIndex(cell.background);
          rtfNoteDoc({ type: 'inline', children: cell.children });
        }
      }
      return;
    }
    var kids = node.children || [];
    for (var k = 0; k < kids.length; k++) rtfNoteDoc(kids[k]);
  }
  function rtfHexToRgb(hex) { return [parseInt(hex.slice(0, 2), 16), parseInt(hex.slice(2, 4), 16), parseInt(hex.slice(4, 6), 16)]; }
  function blocksTextRtf(blocks, inlineFn) {
    var parts = [];
    for (var i = 0; i < (blocks || []).length; i++) {
      var block = blocks[i];
      if (block.type === 'paragraph' || block.type === 'heading') parts.push(inlineFn(block.children));
      else if (block.type === 'code_block') parts.push(String(block.text || ''));
      else if (block.type === 'table') {
        var rows = [];
        for (var r = 0; r < (block.rows || []).length; r++) {
          var cells = [];
          for (var c = 0; c < (block.rows[r].cells || []).length; c++) cells.push(cellTextRtf(block.rows[r].cells[c], inlineFn));
          rows.push(cells.join(' / '));
        }
        parts.push(rows.join(' ; '));
      } else if (block.type === 'bullet_list' || block.type === 'ordered_list') {
        var items = [];
        for (var li = 0; li < (block.children || []).length; li++) items.push(blocksTextRtf(block.children[li].children || [], inlineFn));
        parts.push(items.join(' / '));
      } else if (block.type === 'image') parts.push('[image]');
    }
    return parts.join(' / ');
  }
  function cellTextRtf(cell, inlineFn) {
    if (Array.isArray(cell.blocks) && cell.blocks.length) return blocksTextRtf(cell.blocks, inlineFn);
    return inlineFn(cell.children);
  }
  function exportRTF(inst) {
    hyphenationLanguage = documentLanguage(inst);
    rtfFonts = ['Helvetica', 'Courier New']; rtfFontMap = { helvetica: 0, 'courier new': 1 }; rtfColors = []; rtfColorMap = {};
    Object.keys(CODE_TOKEN_COLORS).forEach(function (color) { rtfColorIndex(color); });
    for (var ci = 0; ci < (inst.doc.children || []).length; ci++) rtfNoteDoc(inst.doc.children[ci]);
    if (inst.headerValue) rtfNoteDoc(inst.headerValue);
    if (inst.footerValue) rtfNoteDoc(inst.footerValue);
    var fontDefs = '';
    for (var f = 0; f < rtfFonts.length; f++) fontDefs += '{\\f' + f + ' ' + rtfFonts[f] + ';}';
    var header = '{\\rtf1\\ansi\\deff0\\uc1{\\fonttbl' + fontDefs + '}';
    var rtfPage = pageConfig(inst, 1), rtfDims = pageSizeMm(rtfPage);
    header += '\\paperw' + mmToTwips(rtfDims.w) + '\\paperh' + mmToTwips(rtfDims.h) +
      '\\margl' + mmToTwips(rtfPage.margins.left) + '\\margr' + mmToTwips(rtfPage.margins.right) +
      '\\margt' + mmToTwips(rtfPage.margins.top) + '\\margb' + mmToTwips(rtfPage.margins.bottom);
    if (rtfColors.length) {
      header += '{\\colortbl;';
      for (var c = 0; c < rtfColors.length; c++) {
        var rgb = rtfHexToRgb(rtfColors[c]);
        header += '\\red' + rgb[0] + '\\green' + rgb[1] + '\\blue' + rgb[2] + ';';
      }
      header += '}';
    }
    var out = [header + '\\viewkind4'];
    function inline(nodes, baseMarks, hyphenate) {
      var s = '';
      for (var i = 0; i < (nodes || []).length; i++) {
        var n = nodes[i];
        if (n.type === 'break') { s += '\\line '; continue; }
        if (n.type === 'image') { s += rtfPicture(n.src, n.width, n.height); continue; }
        if (n.type !== 'text') continue;
        var m = rtfMarks((baseMarks || []).concat(n.marks || []));
        var value = String(hyphenate ? hyphenateText(n.text) : (n.text || ''));
        value = value.replace(/\\/g, '\\\\').replace(/[{}]/g, '\\$&').replace(/\n/g, '\\line ');
        if (hyphenate) value = value.replace(/\u00AD/g, '\\-');
        value = value.replace(/[^\x00-\x7F]/g, function (ch) { var code = ch.charCodeAt(0); return code > 32767 ? '\\u' + (code - 65536) + '?' : '\\u' + code + '?'; });
        s += m.open + value + m.close;
      }
      return s;
    }
    function rtfCodeSegments(node) {
      var segs = codeExportSegments(node), out = '';
      for (var i = 0; i < segs.length; i++) {
        if (!segs[i].text) continue;
        out += inline([{ type: 'text', text: segs[i].text, marks: codeSegmentMarks(segs[i]) }], null, false);
      }
      return out;
    }
    function chromeInline(value) {
      var blocks = chromeDoc(value).children || [], s = '';
      for (var b = 0; b < blocks.length; b++) {
        var block = blocks[b];
        if (!block) continue;
        if (block.type === 'paragraph' || block.type === 'heading') {
          if (s) s += '\\line ';
          for (var i = 0; i < (block.children || []).length; i++) {
            var n = block.children[i];
            if (!n) continue;
            if (n.type === 'break') { s += '\\line '; continue; }
            if (n.type === 'image') { s += rtfPicture(n.src, n.width, n.height); continue; }
            if (n.type !== 'text') continue;
            var valueText = applyPageTokens(n.text || '', 1, 1, false, inst);
            var parts = String(valueText).split(/(__SUI_PAGE__|__SUI_PAGES__)/);
            for (var p = 0; p < parts.length; p++) {
              var part = parts[p];
              if (!part) continue;
              if (part === '__SUI_PAGE__') s += '{\\field{\\*\\fldinst PAGE}{\\fldrslt 1}}';
              else if (part === '__SUI_PAGES__') s += '{\\field{\\*\\fldinst NUMPAGES}{\\fldrslt 1}}';
              else s += inline([{ type: 'text', text: part, marks: n.marks || [] }]);
            }
          }
        } else if (block.type === 'image') {
          s += rtfPicture(block.src, block.width, block.height);
        }
      }
      return s;
    }
    var rtfHeaderBody = chromeInline(inst.headerValue || inst.headerText || '');
    var rtfFooterBody = chromeInline(inst.footerValue || inst.footerText || '');
    if (rtfHeaderBody) out.splice(1, 0, '{\\header \\pard\\qc ' + rtfHeaderBody + '\\par}');
    if (rtfFooterBody) out.splice(1 + (rtfHeaderBody ? 1 : 0), 0, '{\\footer \\pard\\qc ' + rtfFooterBody + '\\par}');
    function paragraphProps(n) {
      var p = '\\pard';
      if (n.align === 'center') p += '\\qc';
      else if (n.align === 'right') p += '\\qr';
      else if (n.align === 'justify') p += '\\qj';
      var indent = parseFloat(n.indent);
      if (isFinite(indent) && indent > 0) p += '\\li' + Math.round(indent * 15);
      var before = cssPt(n.spaceBefore), after = cssPt(n.spaceAfter), lineInfo = lineHeightInfo(n.lineHeight);
      if (before) p += '\\sb' + Math.round(before * 20);
      if (after) p += '\\sa' + Math.round(after * 20);
      if (lineInfo && lineInfo.multiple) p += '\\sl' + Math.round(lineInfo.multiple * 240) + '\\slmult1';
      else if (lineInfo && lineInfo.pt) p += '\\sl' + Math.round(lineInfo.pt * 20) + '\\slmult0';
      return p;
    }
    function block(n) {
      if (!n) return;
      if (n.type === 'heading') {
        var hsize = Math.max(12, 20 - ((Math.max(1, Math.min(6, Number(n.level) || 1)) - 1) * 2));
        out.push(paragraphProps(n) + '\\b\\fs' + Math.round(hsize * 2) + ' ' + inline(n.children, null, n.hyphens === 'auto') + '\\b0\\fs24\\par');
      } else if (n.type === 'paragraph') {
        out.push(paragraphProps(n) + ' ' + inline(n.children, null, n.hyphens === 'auto') + '\\par');
      } else if (n.type === 'blockquote') { for (var k = 0; k < (n.children || []).length; k++) block(n.children[k]); }
      else if (n.type === 'code_block') {
        var rtfCodeBody = (n.codeLanguage && n.codeLanguage !== 'plain') || n.lineNumbers ? rtfCodeSegments(n) : (function () {
          var nodes = (Array.isArray(n.children) && n.children.length) ? n.children : [{ type: 'text', text: String(n.text || ''), marks: [] }];
          return inline(nodes, [{ type: 'font', value: 'Courier New' }, { type: 'size', value: '9' }]);
        })();
        out.push(paragraphProps(n) + '\\f0\\fs20 ' + rtfCodeBody + '\\fs24\\par');
      }
      else if (n.type === 'hr') out.push('\\pard\\brdrb\\brdrs\\brdrw10\\par');
      else if (n.type === 'page_break') out.push('\\page');
      else if (n.type === 'image') out.push('\\pard' + (rtfPicture(n.src, n.width, n.height) || ('[' + (n.alt || 'image') + ']')) + '\\par');
      else if (n.type === 'bullet_list' || n.type === 'ordered_list') {
        for (var i = 0; i < (n.children || []).length; i++) {
          var item = n.children[i], firstBlock = item.children && item.children[0];
          var text = firstBlock ? inline(firstBlock.children, null, firstBlock.hyphens === 'auto') : '';
          out.push('\\pard\\li360 ' + (n.type === 'ordered_list' ? (i + 1) + '. ' : '\\u8226? ') + text + '\\par');
        }
      } else if (n.type === 'table') {
        var rtfColTwips = [];
        if (Array.isArray(n.colWidths) && n.colWidths.length) {
          var rtwTotal = 0, rtwVals = [];
          for (var rwi = 0; rwi < n.colWidths.length; rwi++) {
            var rwv = parseFloat(n.colWidths[rwi]) || 0;
            rtwVals.push(rwv); rtwTotal += rwv;
          }
          if (rtwTotal > 0) for (var rwj = 0; rwj < rtwVals.length; rwj++) rtfColTwips.push(Math.round((rtwVals[rwj] / rtwTotal) * 9638));
        }
        for (var r = 0; r < (n.rows || []).length; r++) {
          var defs = '', cells = '', x = 0, rtfCol = 0;
          var row = n.rows[r];
          for (var cc = 0; cc < (row.cells || []).length; cc++) {
            var cell = row.cells[cc];
            var rtfSpan = Math.max(1, Number(cell.colspan) || 1);
            if (rtfColTwips.length) {
              var rtfCellW = 0;
              for (var rws = 0; rws < rtfSpan; rws++) rtfCellW += rtfColTwips[rtfCol + rws] || 0;
              x += rtfCellW || Math.round(9638 / Math.max(1, row.cells.length));
              rtfCol += rtfSpan;
            } else {
              var pctMatch = /^([\d.]+)%$/.exec(String(cell.width || ''));
              x += pctMatch ? Math.round((parseFloat(pctMatch[1]) / 100) * 9638) : Math.round(9638 / Math.max(1, row.cells.length));
            }
            var ci2 = cell.background ? rtfColorIndex(cell.background) : 0;
            defs += '\\clbrdrt\\brdrs\\brdrw10\\clbrdrl\\brdrs\\brdrw10\\clbrdrb\\brdrs\\brdrw10\\clbrdrr\\brdrs\\brdrw10';
            if (ci2) defs += '\\clcbpat' + ci2;
            defs += '\\cellx' + x;
            var cellProps = '\\pard\\intbl';
            if (cell.align === 'center') cellProps += '\\qc';
            else if (cell.align === 'right') cellProps += '\\qr';
            cells += cellProps + ' ' + cellTextRtf(cell, inline) + '\\cell ';
          }
          out.push('\\trowd\\trgaph108' + defs + ' ' + cells + '\\row');
        }
        out.push('\\pard');
      }
    }
    for (var i = 0; i < (inst.doc.children || []).length; i++) block(inst.doc.children[i]);
    out.push('}');
    return out.join('\n');
  }
  var docxImages = Object.create(null);
  function docxImage(src, width, height) {
    var info = docxImages[src];
    if (!info) return '';
    var w = Math.max(32, Number(width) || 160), h = Math.max(24, Number(height) || 120);
    var cx = Math.round(w * 9525), cy = Math.round(h * 9525);
    return '<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="' + cx + '" cy="' + cy + '"/><wp:docPr id="' + info.id + '" name="Picture ' + info.id + '"/><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:nvPicPr><pic:cNvPr id="' + info.id + '" name="' + info.name + '"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="' + info.rId + '"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="' + cx + '" cy="' + cy + '"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>';
  }
  function docxRun(text, marks, extra) {
    var props = '', font = '', size = 0, color = '', fill = '';
    for (var i = 0; i < (marks || []).length; i++) {
      var t = marks[i].type;
      if (t === 'bold') props += '<w:b/><w:bCs/>';
      else if (t === 'italic') props += '<w:i/><w:iCs/>';
      else if (t === 'underline') props += '<w:u w:val="single"/>';
      else if (t === 'strike') props += '<w:strike/>';
      else if (t === 'sup') props += '<w:vertAlign w:val="superscript"/>';
      else if (t === 'sub') props += '<w:vertAlign w:val="subscript"/>';
      else if (t === 'link') { color = '4C5CF7'; props += '<w:u w:val="single"/>'; }
      else if (t === 'color' && marks[i].value) color = hexColor(marks[i].value) || color;
      else if (t === 'background' && marks[i].value) fill = hexColor(marks[i].value) || '';
      else if (t === 'size' && marks[i].value) size = cssPt(marks[i].value);
      else if (t === 'font' && marks[i].value) font = firstFontFamily(marks[i].value);
    }
    if (font) props += '<w:rFonts w:ascii="' + xmlEscapeText(font) + '" w:hAnsi="' + xmlEscapeText(font) + '" w:cs="' + xmlEscapeText(font) + '"/>';
    if (size > 0) { var half = Math.max(8, Math.min(192, Math.round(size * 2))); props += '<w:sz w:val="' + half + '"/><w:szCs w:val="' + half + '"/>'; }
    if (color) props += '<w:color w:val="' + color + '"/>';
    if (fill) props += '<w:shd w:val="clear" w:color="auto" w:fill="' + fill + '"/>';
    if (docxLanguage) props += '<w:lang w:val="' + xmlEscapeText(docxLanguage) + '"/>';
    return '<w:r><w:rPr>' + props + (extra || '') + '</w:rPr><w:t xml:space="preserve">' + xmlEscapeText(text) + '</w:t></w:r>';
  }
  function docxInline(nodes, extra, hyphenate) {
    var out = '';
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (n.type === 'text') out += docxRun(hyphenate ? hyphenateText(n.text) : n.text, n.marks, extra);
      else if (n.type === 'break') out += '<w:r><w:br/></w:r>';
      else if (n.type === 'image') out += docxImage(n.src, n.width, n.height) || docxRun('[' + (n.alt || 'image') + ']', [], extra);
    }
    return out;
  }
  function docxCodeSegments(node) {
    var segs = codeExportSegments(node), out = '';
    for (var i = 0; i < segs.length; i++) {
      if (!segs[i].text) continue;
      var parts = String(segs[i].text).split('\n');
      var marks = codeSegmentMarks(segs[i]);
      for (var p = 0; p < parts.length; p++) {
        if (p) out += '<w:r><w:br/></w:r>';
        if (parts[p]) out += docxRun(parts[p], marks);
      }
    }
    return out;
  }
  function docxCodeInline(nodes, fallbackText) {
    var baseMarks = [{ type: 'font', value: 'Courier New' }, { type: 'size', value: '9' }];
    var list = (nodes && nodes.length) ? nodes : [{ type: 'text', text: String(fallbackText || ''), marks: [] }];
    var out = '';
    for (var i = 0; i < list.length; i++) {
      var n = list[i];
      if (n.type === 'break') { out += '<w:r><w:br/></w:r>'; continue; }
      if (n.type === 'image') { out += docxImage(n.src, n.width, n.height) || docxRun('[' + (n.alt || 'image') + ']', baseMarks); continue; }
      if (n.type !== 'text') continue;
      var parts = String(n.text == null ? '' : n.text).split('\n');
      for (var p = 0; p < parts.length; p++) {
        if (p) out += '<w:r><w:br/></w:r>';
        if (parts[p]) out += docxRun(parts[p], (n.marks || []).concat(baseMarks));
      }
    }
    return out;
  }
  function docxParagraphProps(n, extra) {
    var pPr = '';
    if (n.align) pPr += '<w:jc w:val="' + (n.align === 'justify' ? 'both' : String(n.align)) + '"/>';
    var before = cssPt(n.spaceBefore), after = cssPt(n.spaceAfter);
    var lineInfo = lineHeightInfo(n.lineHeight);
    var spacing = '';
    if (before > 0) spacing += ' w:before="' + Math.round(before * 20) + '"';
    if (after > 0) spacing += ' w:after="' + Math.round(after * 20) + '"';
    if (lineInfo && lineInfo.multiple) spacing += ' w:line="' + Math.round(lineInfo.multiple * 240) + '" w:lineRule="auto"';
    else if (lineInfo && lineInfo.pt) spacing += ' w:line="' + Math.round(lineInfo.pt * 20) + '" w:lineRule="' + (n.lineRule === 'atLeast' ? 'atLeast' : 'exact') + '"';
    if (spacing) pPr += '<w:spacing' + spacing + '/>';
    var indent = parseFloat(n.indent);
    if (isFinite(indent) && indent > 0) pPr += '<w:ind w:left="' + Math.round(indent * 15) + '"/>';
    return '<w:pPr>' + (extra || '') + pPr + '</w:pPr>';
  }
  function docxCellProps(cell, continuation) {
    var props = '';
    var pct = /^([\d.]+)%$/.exec(String(cell && cell.width || ''));
    if (pct) props += '<w:tcW w:w="' + Math.max(1, Math.round(parseFloat(pct[1]) * 50)) + '" w:type="pct"/>';
    if (continuation) { props += '<w:vMerge/>'; return '<w:tc><w:tcPr>' + props + '</w:tcPr><w:p/></w:tc>'; }
    if (Number(cell.colspan) > 1) props += '<w:gridSpan w:val="' + Number(cell.colspan) + '"/>';
    if (Number(cell.rowspan) > 1) props += '<w:vMerge w:val="restart"/>';
    var fill = hexColor(cell.background);
    if (fill) props += '<w:shd w:val="clear" w:color="auto" w:fill="' + fill + '"/>';
    var jc = cell.align ? '<w:jc w:val="' + (cell.align === 'justify' ? 'both' : String(cell.align)) + '"/>' : '';
    var cellBody = '';
    if (Array.isArray(cell.blocks) && cell.blocks.length) {
      for (var cbx = 0; cbx < cell.blocks.length; cbx++) cellBody += docxBlock(cell.blocks[cbx], 1);
      if (cell.blocks[cell.blocks.length - 1].type === 'table') cellBody += '<w:p/>';
    }
    if (!cellBody) cellBody = '<w:p><w:pPr>' + jc + '</w:pPr>' + docxInline(cell.children) + '</w:p>';
    return '<w:tc><w:tcPr>' + props + '</w:tcPr>' + cellBody + '</w:tc>';
  }
  function docxChromeInline(nodes, inst) {
    var out = '';
    var fieldRun = '<w:r><w:rPr><w:sz w:val="18"/><w:color w:val="666666"/></w:rPr><w:t>1</w:t></w:r>';
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (!n) continue;
      if (n.type === 'break') { out += '<w:r><w:br/></w:r>'; continue; }
      if (n.type === 'image') { out += docxImage(n.src, n.width, n.height) || ''; continue; }
      if (n.type !== 'text') continue;
      var value = applyPageTokens(n.text || '', 1, 1, false, inst);
      var parts = String(value).split(/(__SUI_PAGE__|__SUI_PAGES__)/);
      for (var p = 0; p < parts.length; p++) {
        var part = parts[p];
        if (!part) continue;
        if (part === '__SUI_PAGE__') out += '<w:fldSimple w:instr=" PAGE ">' + fieldRun + '</w:fldSimple>';
        else if (part === '__SUI_PAGES__') out += '<w:fldSimple w:instr=" NUMPAGES ">' + fieldRun + '</w:fldSimple>';
        else out += docxRun(part, n.marks || []);
      }
    }
    return out;
  }
  function docxChromeBody(value, inst) {
    var doc = chromeDoc(value);
    var blocks = doc.children || [];
    var out = '';
    for (var i = 0; i < blocks.length; i++) {
      var block = blocks[i];
      if (!block) continue;
      if (block.type === 'paragraph' || block.type === 'heading') {
        var border = i === blocks.length - 1 ? '<w:pBdr><w:bottom w:val="single" w:sz="4" w:space="1" w:color="CCCCCC"/></w:pBdr>' : '';
        out += '<w:p><w:pPr><w:jc w:val="center"/>' + border + '</w:pPr>' + docxChromeInline(block.children, inst) + '</w:p>';
      } else if (block.type === 'image') {
        out += '<w:p><w:pPr><w:jc w:val="center"/></w:pPr>' + (docxImage(block.src, block.width, block.height) || '') + '</w:p>';
      }
    }
    return out;
  }
  function docxChromeRuns(text, inst) {
    var format = '<w:rPr><w:sz w:val="18"/><w:color w:val="666666"/></w:rPr>';
    var parts = String(resolveChromeVariables(text, inst)).split(/(\{page\}|\{pages\}|\[page\]|\[pages\])/i), runs = '';
    for (var i = 0; i < parts.length; i++) {
      var part = parts[i];
      if (/^(\{page\}|\[page\])$/i.test(part)) runs += '<w:fldSimple w:instr=" PAGE "><w:r>' + format + '<w:t>1</w:t></w:r></w:fldSimple>';
      else if (/^(\{pages\}|\[pages\])$/i.test(part)) runs += '<w:fldSimple w:instr=" NUMPAGES "><w:r>' + format + '<w:t>1</w:t></w:r></w:fldSimple>';
      else if (part) runs += '<w:r>' + format + '<w:t xml:space="preserve">' + xmlEscapeText(part) + '</w:t></w:r>';
    }
    return runs;
  }
  function docxList(node, level) {
    var numId = node.type === 'ordered_list' ? 2 : 1;
    var out = '';
    for (var i = 0; i < (node.children || []).length; i++) {
      var item = node.children[i];
      var blocks = item.children || [];
      for (var b = 0; b < blocks.length; b++) {
        var inner = blocks[b];
        if (inner.type === 'paragraph') {
          var extra = '<w:numPr><w:ilvl w:val="' + Math.min(8, level) + '"/><w:numId w:val="' + numId + '"/></w:numPr>';
          out += '<w:p>' + docxParagraphProps(inner, extra) + docxInline(inner.children, '', inner.hyphens === 'auto') + '</w:p>';
        } else if (inner.type === 'bullet_list' || inner.type === 'ordered_list') {
          out += docxList(inner, level + 1);
        } else out += docxBlock(inner);
      }
    }
    return out;
  }
  function docxNumberingXml() {
    function levels(fmt, pattern, bullet) {
      var xml = '';
      for (var l = 0; l < 9; l++) {
        var text = bullet ? (l % 3 === 0 ? '\u2022' : l % 3 === 1 ? '\u25e6' : '\u25aa') : pattern.replace('%1', '%' + (l + 1)) ;
        xml += '<w:lvl w:ilvl="' + l + '"><w:start w:val="1"/><w:numFmt w:val="' + (bullet ? 'bullet' : fmt) + '"/><w:lvlText w:val="' + text + '"/><w:lvlJc w:val="left"/><w:pPr><w:ind w:left="' + (720 + l * 720) + '" w:hanging="360"/></w:pPr>' + (bullet ? '<w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/></w:rPr>' : '') + '</w:lvl>';
      }
      return xml;
    }
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">' +
      '<w:abstractNum w:abstractNumId="0">' + levels('bullet', '', true) + '</w:abstractNum>' +
      '<w:abstractNum w:abstractNumId="1">' + levels('decimal', '%1.', false) + '</w:abstractNum>' +
      '<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>' +
      '<w:num w:numId="2"><w:abstractNumId w:val="1"/></w:num>' +
      '</w:numbering>';
  }
  function docxBlock(n, depth) {
    if (!n) return '';
    depth = depth || 0;
    if (n.type === 'paragraph') return '<w:p>' + docxParagraphProps(n) + docxInline(n.children, '', n.hyphens === 'auto') + '</w:p>';
    if (n.type === 'heading') {
      var level = Math.max(1, Math.min(6, Number(n.level) || 1));
      var hsize = Math.max(12, 20 - (level - 1) * 2);
      return '<w:p>' + docxParagraphProps(n, '<w:keepNext/>') + docxInline(n.children, '<w:b/><w:sz w:val="' + Math.round(hsize * 2) + '"/>', n.hyphens === 'auto') + '</w:p>';
    }
    if (n.type === 'blockquote') return (n.children || []).map(docxBlock).join('');
    if (n.type === 'code_block') {
      var codeExtra = '<w:shd w:val="clear" w:color="auto" w:fill="F2F2F2"/><w:suppressAutoHyphens/>';
      var codeBody = (n.codeLanguage && n.codeLanguage !== 'plain') || n.lineNumbers ? docxCodeSegments(n) : docxCodeInline(n.children, n.text);
      return '<w:p>' + docxParagraphProps(n, codeExtra) + codeBody + '</w:p>';
    }
    if (n.type === 'hr') return '<w:p><w:pPr><w:pBdr><w:bottom w:val="single" w:sz="6" w:space="1" w:color="999999"/></w:pBdr></w:pPr></w:p>';
    if (n.type === 'page_break') return '<w:p><w:r><w:br w:type="page"/></w:r></w:p>';
    if (n.type === 'image') return '<w:p><w:pPr><w:jc w:val="center"/></w:pPr>' + (docxImage(n.src, n.width, n.height) || docxRun('[' + (n.alt || 'image') + ']', [])) + '</w:p>';
    if (n.type === 'bullet_list' || n.type === 'ordered_list') return docxList(n, 0);
    if (n.type === 'table') {
      var merges = {}, rows = '';
      var maxCols = 0;
      for (var mc = 0; mc < (n.rows || []).length; mc++) {
        var msum = 0;
        for (var mm = 0; mm < (n.rows[mc].cells || []).length; mm++) msum += Math.max(1, Number(n.rows[mc].cells[mm].colspan) || 1);
        if (msum > maxCols) maxCols = msum;
      }
      var gridWeights = new Array(maxCols).fill(0);
      var colWeights = [];
      if (Array.isArray(n.colWidths) && n.colWidths.length) {
        var cwValues = [], cwTotal = 0;
        for (var cwi = 0; cwi < n.colWidths.length; cwi++) {
          var cwNum = parseFloat(n.colWidths[cwi]);
          if (isFinite(cwNum) && cwNum > 0) { cwValues.push(cwNum); cwTotal += cwNum; }
          else cwValues.push(0);
        }
        if (cwTotal > 0) for (var cwj = 0; cwj < cwValues.length && cwj < maxCols; cwj++) colWeights[cwj] = cwValues[cwj] / cwTotal;
      }
      for (var cwk = 0; cwk < maxCols; cwk++) if (colWeights[cwk]) gridWeights[cwk] = colWeights[cwk];
      for (var gw = 0; gw < (n.rows || []).length; gw++) {
        var gcol = 0;
        for (var gc = 0; gc < (n.rows[gw].cells || []).length; gc++) {
          var gcell = n.rows[gw].cells[gc];
          var gspan = Math.max(1, Number(gcell.colspan) || 1);
          var gpct = /^([\d.]+)%$/.exec(String(gcell.width || ''));
          if (gpct) {
            var gshare = (parseFloat(gpct[1]) / 100) / gspan;
            for (var gs = 0; gs < gspan && gcol + gs < maxCols; gs++) if (!colWeights[gcol + gs] && gshare > gridWeights[gcol + gs]) gridWeights[gcol + gs] = gshare;
          }
          gcol += gspan;
        }
      }
      var gsum = 0, gmissing = 0;
      for (var gi = 0; gi < maxCols; gi++) { if (gridWeights[gi] > 0) gsum += gridWeights[gi]; else gmissing++; }
      var gfill = gmissing ? Math.max(0, (1 - gsum) / gmissing) : 0;
      var gridXml = '';
      for (var gj = 0; gj < maxCols; gj++) {
        var gwk = gridWeights[gj] || gfill || 1 / maxCols;
        gridXml += '<w:gridCol w:w="' + Math.max(120, Math.round(gwk * 9638)) + '"/>';
      }
      for (var r = 0; r < (n.rows || []).length; r++) {
        var row = n.rows[r], cells = '', col = 0, guard = 0;
        while (merges[col] && merges[col].left > 0 && guard++ < 128) {
          var m = merges[col];
          cells += docxCellProps({ width: m.width }, true);
          m.left--;
          col += m.span || 1;
        }
        cells += '';
        for (var c = 0; c < (row.cells || []).length; c++) {
          var cell = row.cells[c];
          cells += docxCellProps(cell, false);
          var span = Math.max(1, Number(cell.colspan) || 1);
          if (Number(cell.rowspan) > 1) merges[col] = { left: Number(cell.rowspan) - 1, span: span, width: cell.width };
          col += span;
        }
        var trPrInner = '';
        var allHeader = row.cells && row.cells.length > 0 && row.cells.every(function (rc) { return rc.header; });
        if (allHeader || (r === 0 && n.header)) trPrInner += '<w:tblHeader/>';
        var rowH = parseFloat(row.height);
        if (isFinite(rowH) && rowH > 0) trPrInner += '<w:trHeight w:val="' + Math.round(rowH * 15) + '" w:hRule="atLeast"/>';
        rows += '<w:tr>' + (trPrInner ? '<w:trPr>' + trPrInner + '</w:trPr>' : '') + cells + '</w:tr>';
      }
      var docxBorderColor = hexColor(n.border) || '999999';
      var docxBorderWidth = n.borderWidth ? Math.max(0, Math.round(parseFloat(n.borderWidth) * 8)) : 4;
      var docxBorderFlags = n.borders || ['top', 'right', 'bottom', 'left', 'insideH', 'insideV'];
      var docxBorder = function (side, inside) {
        var enabled = docxBorderFlags.indexOf(side) >= 0 && docxBorderWidth > 0;
        return '<w:' + side + ' w:val="' + (enabled ? 'single' : 'nil') + '"' + (enabled ? ' w:sz="' + docxBorderWidth + '" w:space="0" w:color="' + docxBorderColor + '"' : '') + '/>';
      };
      var padTwips = n.padding ? Math.max(0, Math.round(parseFloat(n.padding) * 15)) : 0;
      var cellMar = '<w:tblCellMar><w:top w:w="' + padTwips + '" w:type="dxa"/><w:left w:w="' + padTwips + '" w:type="dxa"/><w:bottom w:w="' + padTwips + '" w:type="dxa"/><w:right w:w="' + padTwips + '" w:type="dxa"/></w:tblCellMar>';
      var nestedTable = depth > 0;
      return '<w:tbl><w:tblPr>' + (nestedTable ? '<w:tblW w:w="0" w:type="auto"/><w:tblLayout w:type="autofit"/>' : '<w:tblW w:w="9638" w:type="dxa"/><w:tblLayout w:type="fixed"/>') + '<w:tblBorders>' + docxBorder('top') + docxBorder('left') + docxBorder('bottom') + docxBorder('right') + docxBorder('insideH') + docxBorder('insideV') + '</w:tblBorders>' + cellMar + '</w:tblPr>' + (nestedTable ? '' : '<w:tblGrid>' + gridXml + '</w:tblGrid>') + rows + '</w:tbl><w:p/>';
    }
    return '';
  }
  var docxAutoHyphenation = false;
  var docxLanguage = '';
  function docHasAutoHyphenation(blocks) {
    for (var i = 0; i < (blocks || []).length; i++) {
      var n = blocks[i];
      if (!n) continue;
      if (n.hyphens === 'auto') return true;
      if (n.children && docHasAutoHyphenation(n.children)) return true;
      if (n.type === 'table') {
        for (var r = 0; r < (n.rows || []).length; r++) {
          for (var c = 0; c < (n.rows[r].cells || []).length; c++) {
            var cell = n.rows[r].cells[c];
            if (docHasAutoHyphenation(cell.blocks || cell.children || [])) return true;
          }
        }
      }
    }
    return false;
  }
  function exportDOCX(inst) {
    hyphenationLanguage = documentLanguage(inst);
    docxImages = Object.create(null);
    docxAutoHyphenation = docHasAutoHyphenation(inst.doc.children || []);
    docxLanguage = String(documentLanguage(inst) || 'en');
    var media = [], relsXml = '', extSet = {}, imageId = 0, hasLists = false;
    function collect(node) {
      if (!node) return;
      if (node.type === 'image' && node.src && !docxImages[node.src]) {
        var info = bytesFromDataUrl(node.src);
        if (info) {
          imageId++;
          var name = 'image' + imageId + '.' + info.ext;
          var rId = 'rId' + (100 + imageId);
          docxImages[node.src] = { id: imageId, name: name, rId: rId };
          media.push({ name: 'word/media/' + name, data: info.bytes });
          relsXml += '<Relationship Id="' + rId + '" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/' + name + '"/>';
          extSet[info.ext] = info.mime;
        }
      }
      if (node.type === 'bullet_list' || node.type === 'ordered_list') hasLists = true;
      var kids = node.children || [];
      for (var i = 0; i < kids.length; i++) collect(kids[i]);
      if (node.type === 'table') for (var r = 0; r < (node.rows || []).length; r++) for (var c = 0; c < (node.rows[r].cells || []).length; c++) collect(node.rows[r].cells[c]);
    }
    for (var ci = 0; ci < (inst.doc.children || []).length; ci++) collect(inst.doc.children[ci]);
    if (inst.headerValue) collect(inst.headerValue);
    if (inst.footerValue) collect(inst.footerValue);
    var body = '';
    for (var i = 0; i < (inst.doc.children || []).length; i++) body += docxBlock(inst.doc.children[i]);
    var headerBody = docxChromeBody(inst.headerValue || inst.headerText || '', inst);
    var footerBody = docxChromeBody(inst.footerValue || inst.footerText || '', inst);
    var chromeNS = ' xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"';
    var headerXml = headerBody ? '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"' + chromeNS + '>' + headerBody + '</w:hdr>' : '';
    var footerXml = footerBody ? '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:ftr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"' + chromeNS + '>' + footerBody + '</w:ftr>' : '';
    var docxPage = pageConfig(inst, 1), docxDims = pageSizeMm(docxPage);
    var docxPgSz = '<w:pgSz w:w="' + mmToTwips(docxDims.w) + '" w:h="' + mmToTwips(docxDims.h) + '"' + (docxPage.orientation === 'landscape' ? ' w:orient="landscape"' : '') + '/>';
    var docxPgMar = '<w:pgMar w:top="' + mmToTwips(docxPage.margins.top) + '" w:right="' + mmToTwips(docxPage.margins.right) + '" w:bottom="' + mmToTwips(docxPage.margins.bottom) + '" w:left="' + mmToTwips(docxPage.margins.left) + '" w:header="720" w:footer="720" w:gutter="0"/>';
    var sectRefs = '';
    if (headerXml) { relsXml += '<Relationship Id="rIdHeader1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>'; sectRefs += '<w:headerReference w:type="default" r:id="rIdHeader1"/>'; }
    if (footerXml) { relsXml += '<Relationship Id="rIdFooter1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>'; sectRefs += '<w:footerReference w:type="default" r:id="rIdFooter1"/>'; }
    var documentXml = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><w:body>' + body + '<w:sectPr>' + sectRefs + docxPgSz + docxPgMar + '</w:sectPr></w:body></w:document>';
    var defaults = '';
    for (var ext in extSet) if (Object.prototype.hasOwnProperty.call(extSet, ext)) defaults += '<Default Extension="' + ext + '" ContentType="' + extSet[ext] + '"/>';
    var contentTypes = '<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/>' + defaults + '<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>' + (headerXml ? '<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>' : '') + (footerXml ? '<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>' : '') + '</Types>';
    var settingsXml = '';
    var rootRels = '<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>';
    if (hasLists) {
      relsXml += '<Relationship Id="rIdNumbering" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>';
      contentTypes = contentTypes.replace('</Types>', '<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/></Types>');
    }
    var docRels = '<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' + relsXml + '</Relationships>';
    var chromeImageRels = '';
    for (var chromeKey in docxImages) {
      if (!hasOwn(docxImages, chromeKey)) continue;
      var chromeInfo = docxImages[chromeKey];
      chromeImageRels += '<Relationship Id="' + chromeInfo.rId + '" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/' + chromeInfo.name + '"/>';
    }
    var headerRelsXml = (headerXml && chromeImageRels) ? '<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' + chromeImageRels + '</Relationships>' : '';
    var footerRelsXml = (footerXml && chromeImageRels) ? '<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' + chromeImageRels + '</Relationships>' : '';
    var entries = [
      { name: '[Content_Types].xml', data: contentTypes },
      { name: '_rels/.rels', data: rootRels },
      { name: 'word/document.xml', data: documentXml },
      { name: 'word/_rels/document.xml.rels', data: docRels }
    ];
    if (hasLists) entries.push({ name: 'word/numbering.xml', data: docxNumberingXml() });
    if (settingsXml) entries.push({ name: 'word/settings.xml', data: settingsXml });
    if (headerXml) entries.push({ name: 'word/header1.xml', data: headerXml });
    if (footerXml) entries.push({ name: 'word/footer1.xml', data: footerXml });
    if (headerRelsXml) entries.push({ name: 'word/_rels/header1.xml.rels', data: headerRelsXml });
    if (footerRelsXml) entries.push({ name: 'word/_rels/footer1.xml.rels', data: footerRelsXml });
    for (var m = 0; m < media.length; m++) entries.push(media[m]);
    return zipStore(entries);
  }
  var odtImages = Object.create(null);
  var odtHyphenStyles = Object.create(null);
  var odtHyphenSeq = 0;
  var odtLanguage = '';
  var odtTableStyles = Object.create(null);
  var odtTableSeq = 0;
  var odtContentWcm = 17;
  function odtTableColumns(n) {
    var cols = Array.isArray(n.colWidths) ? n.colWidths : [];
    if (!cols.length) return '';
    var mx = 0;
    for (var i = 0; i < cols.length; i++) mx += parseFloat(cols[i]) || 0;
    if (!mx) return '';
    var out = '';
    for (var c = 0; c < cols.length; c++) {
      var num = parseFloat(cols[c]) || 0;
      var width = /%$/.test(String(cols[c])) ? (num / 100) * odtContentWcm : (num / 96) * 2.54;
      if (!(width > 0)) width = odtContentWcm / cols.length;
      var key = 'Col' + (++odtTableSeq);
      odtTableStyles[key] = '<style:style style:name="' + key + '" style:family="table-column"><style:table-column-properties style:column-width="' + (Math.round(width * 100) / 100) + 'cm"/></style:style>';
      out += '<table:table-column table:style-name="' + key + '"/>';
    }
    return out;
  }
  function odtHyphenStyle(n, parentStyle) {
    if (!n || n.hyphens !== 'auto') return '';
    var key = 'Hyphen' + (++odtHyphenSeq);
    var parts = String(odtLanguage || 'en').split('-');
    odtHyphenStyles[key] = '<style:style style:name="' + key + '" style:family="paragraph"' + (parentStyle ? ' style:parent-style-name="' + parentStyle + '"' : '') + '><style:paragraph-properties fo:hyphenate="true" fo:language="' + xmlEscapeText(parts[0] || 'en') + '"' + (parts[1] ? ' fo:country="' + xmlEscapeText(parts[1]) + '"' : '') + '/></style:style>';
    return ' text:style-name="' + key + '"';
  }
  function odtImage(src, width, height) {
    var info = odtImages[src];
    if (!info) return '';
    var w = Math.max(32, Number(width) || 160), h = Math.max(24, Number(height) || 120);
    return '<draw:frame draw:name="' + info.name + '" text:anchor-type="as-char" svg:width="' + w + 'px" svg:height="' + h + 'px"><draw:image xlink:href="Pictures/' + info.name + '" xlink:type="simple" xlink:show="embed" xlink:actuate="onLoad"/></draw:frame>';
  }
  function odtInline(nodes, baseStyle, hyphenate) {
    var out = '';
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (n.type === 'break') { out += '<text:line-break/>'; continue; }
      if (n.type === 'image') { out += odtImage(n.src, n.width, n.height); continue; }
      if (n.type !== 'text') continue;
      var marks = n.marks || [], attrs = [];
      if (baseStyle) attrs.push(baseStyle);
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
        else if (t === 'size' && marks[m].value) attrs.push('font-size:' + marks[m].value);
        else if (t === 'font' && marks[m].value) attrs.push('font-family:' + marks[m].value);
      }
      var text = xmlEscapeText(hyphenate ? hyphenateText(n.text) : (n.text || ''));
      if (text.indexOf('\n') >= 0) text = text.split('\n').join('<text:line-break/>');
      out += attrs.length ? '<text:span text:style-name="T' + (attrs.length % 7) + '" style="' + attrs.join(';') + '">' + text + '</text:span>' : text;
    }
    return out;
  }
  function odtList(node) {
    var styleName = node.type === 'ordered_list' ? 'LNumbers' : 'LBullets';
    var out = '<text:list text:style-name="' + styleName + '">';
    for (var i = 0; i < (node.children || []).length; i++) {
      var item = node.children[i], blocks = item.children || [];
      out += '<text:list-item>';
      for (var b = 0; b < blocks.length; b++) {
        var inner = blocks[b];
        if (inner.type === 'paragraph') out += '<text:p>' + odtInline(inner.children, '', inner.hyphens === 'auto') + '</text:p>';
        else if (inner.type === 'bullet_list' || inner.type === 'ordered_list') out += odtList(inner);
        else out += odtBlock(inner);
      }
      out += '</text:list-item>';
    }
    return out + '</text:list>';
  }
  function odtListStyles() {
    var bullets = '', numbers = '';
    var chars = ['\u2022', '\u25e6', '\u25aa'];
    for (var l = 1; l <= 9; l++) {
      bullets += '<text:list-level-style-bullet text:level="' + l + '" text:bullet-char="' + chars[(l - 1) % 3] + '"><style:list-level-properties text:space-before="' + (0.5 + (l - 1) * 0.5) + 'cm" text:min-label-width="0.5cm"/></text:list-level-style-bullet>';
      numbers += '<text:list-level-style-number text:level="' + l + '" style:num-format="1" style:num-suffix="."><style:list-level-properties text:space-before="' + (0.5 + (l - 1) * 0.5) + 'cm" text:min-label-width="0.6cm"/></text:list-level-style-number>';
    }
    return '<text:list-style style:name="LBullets">' + bullets + '</text:list-style><text:list-style style:name="LNumbers">' + numbers + '</text:list-style>';
  }
  function odtChromeInline(nodes, inst) {
    var out = '';
    for (var i = 0; i < (nodes || []).length; i++) {
      var n = nodes[i];
      if (!n) continue;
      if (n.type === 'break') { out += '<text:line-break/>'; continue; }
      if (n.type === 'image') { out += odtImage(n.src, n.width, n.height); continue; }
      if (n.type !== 'text') continue;
      var value = applyPageTokens(n.text || '', 1, 1, false, inst);
      var parts = String(value).split(/(__SUI_PAGE__|__SUI_PAGES__)/);
      for (var p = 0; p < parts.length; p++) {
        var part = parts[p];
        if (!part) continue;
        if (part === '__SUI_PAGE__') out += '<text:page-number text:select-page="current">1</text:page-number>';
        else if (part === '__SUI_PAGES__') out += '<text:page-count>1</text:page-count>';
        else out += odtInline([{ type: 'text', text: part, marks: n.marks || [] }]);
      }
    }
    return out;
  }
  function odtChromeBody(value, inst, styleName) {
    var doc = chromeDoc(value);
    var blocks = doc.children || [];
    var out = '';
    for (var i = 0; i < blocks.length; i++) {
      var block = blocks[i];
      if (!block) continue;
      if (block.type === 'paragraph' || block.type === 'heading') {
        out += '<text:p text:style-name="' + styleName + '">' + odtChromeInline(block.children, inst) + '</text:p>';
      } else if (block.type === 'image') {
        out += '<text:p text:style-name="' + styleName + '">' + (odtImage(block.src, block.width, block.height) || '') + '</text:p>';
      }
    }
    return out;
  }
  function odtCodeSegments(node) {
    var segs = codeExportSegments(node), out = '';
    for (var i = 0; i < segs.length; i++) {
      if (!segs[i].text) continue;
      out += odtInline([{ type: 'text', text: segs[i].text, marks: codeSegmentMarks(segs[i]) }]);
    }
    return out;
  }
  function odtBlock(n) {
    if (!n) return '';
    if (n.type === 'heading') {
      var odtHLevel = Math.min(6, Number(n.level) || 1);
      var odtHName = 'Heading_20_' + odtHLevel;
      return '<text:h text:style-name="' + odtHName + '" text:outline-level="' + (Number(n.level) || 1) + '">' + odtInline(n.children, '', n.hyphens === 'auto') + '</text:h>';
    }
    if (n.type === 'paragraph') {
      var css = [];
      if (n.align) css.push('text-align:' + n.align);
      var lh = lineHeightInfo(n.lineHeight);
      if (lh) css.push('line-height:' + (lh.multiple ? Math.round(lh.multiple * 100) + '%' : lh.pt + 'pt'));
      var style = css.length ? ' style="' + css.join(';') + '"' : '';
      return '<text:p' + style + '>' + odtInline(n.children, '', n.hyphens === 'auto') + '</text:p>';
    }
    if (n.type === 'blockquote') return '<text:section>' + (n.children || []).map(odtBlock).join('') + '</text:section>';
    if (n.type === 'code_block') {
      var odtCodeBody = (n.codeLanguage && n.codeLanguage !== 'plain') || n.lineNumbers ? odtCodeSegments(n) : odtInline((Array.isArray(n.children) && n.children.length) ? n.children : [{ type: 'text', text: String(n.text || ''), marks: [] }], 'font-family:monospace;font-size:9pt');
      return '<text:p>' + odtCodeBody + '</text:p>';
    }
    if (n.type === 'hr') return '<text:p>---</text:p>';
    if (n.type === 'page_break') return '<text:p text:style-name="PageBreak"/>';
    if (n.type === 'image') return '<text:p>' + (odtImage(n.src, n.width, n.height) || ('[' + xmlEscapeText(n.alt || n.caption || 'image') + ']')) + '</text:p>';
    if (n.type === 'bullet_list' || n.type === 'ordered_list') return odtList(n);
    if (n.type === 'table') {
      var out = '<table:table>' + odtTableColumns(n);
      for (var r = 0; r < (n.rows || []).length; r++) {
        out += '<table:table-row>';
        for (var c = 0; c < (n.rows[r].cells || []).length; c++) {
          var cell = n.rows[r].cells[c];
          var cellStyle = cell.background ? ' style="background-color:' + cell.background + '"' : '';
          var odtCellBody = '';
          if (Array.isArray(cell.blocks) && cell.blocks.length) {
            for (var ocb = 0; ocb < cell.blocks.length; ocb++) odtCellBody += odtBlock(cell.blocks[ocb]);
          }
          if (!odtCellBody) odtCellBody = '<text:p>' + odtInline(cell.children) + '</text:p>';
          out += '<table:table-cell office:value-type="string"' + cellStyle + (Number(cell.colspan) > 1 ? ' table:number-columns-spanned="' + Number(cell.colspan) + '"' : '') + '>' + odtCellBody + '</table:table-cell>';
        }
        out += '</table:table-row>';
      }
      return out + '</table:table>';
    }
    return '';
  }
  function exportODT(inst) {
    hyphenationLanguage = documentLanguage(inst);
    odtImages = Object.create(null);
    odtHyphenStyles = Object.create(null);
    odtHyphenSeq = 0;
    odtLanguage = String(documentLanguage(inst) || 'en');
    odtTableStyles = Object.create(null);
    odtTableSeq = 0;
    var odtSetupPage = pageConfig(inst, 1), odtSetupDims = pageSizeMm(odtSetupPage);
    odtContentWcm = mmToCm(odtSetupDims.w - odtSetupPage.margins.left - odtSetupPage.margins.right);
    var media = [], manifestMedia = '', imageId = 0;
    function collect(node) {
      if (!node) return;
      if (node.type === 'image' && node.src && !odtImages[node.src]) {
        var info = bytesFromDataUrl(node.src);
        if (info) {
          imageId++;
          var name = 'image' + imageId + '.' + info.ext;
          odtImages[node.src] = { name: name };
          media.push({ name: 'Pictures/' + name, data: info.bytes });
          manifestMedia += '<manifest:file-entry manifest:full-path="Pictures/' + name + '" manifest:media-type="' + info.mime + '"/>';
        }
      }
      var kids = node.children || [];
      for (var i = 0; i < kids.length; i++) collect(kids[i]);
      if (node.type === 'table') for (var r = 0; r < (node.rows || []).length; r++) for (var c = 0; c < (node.rows[r].cells || []).length; c++) collect(node.rows[r].cells[c]);
    }
    for (var ci = 0; ci < (inst.doc.children || []).length; ci++) collect(inst.doc.children[ci]);
    if (inst.headerValue) collect(inst.headerValue);
    if (inst.footerValue) collect(inst.footerValue);
    var body = '';
    for (var i = 0; i < (inst.doc.children || []).length; i++) body += odtBlock(inst.doc.children[i]);
    var odtAuto = '';
    for (var ahs in odtHyphenStyles) if (Object.prototype.hasOwnProperty.call(odtHyphenStyles, ahs)) odtAuto += odtHyphenStyles[ahs];
    for (var ats in odtTableStyles) if (Object.prototype.hasOwnProperty.call(odtTableStyles, ats)) odtAuto += odtTableStyles[ats];
    var odtContent = '<?xml version="1.0" encoding="UTF-8"?><office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:svg="urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0" office:version="1.2">' + (odtAuto ? '<office:automatic-styles>' + odtAuto + '</office:automatic-styles>' : '') + '<office:body><office:text>' + body + '</office:text></office:body></office:document-content>';
    function odtChrome(text) {
      var parts = String(resolveChromeVariables(text, inst)).split(/(\{page\}|\{pages\}|\[page\]|\[pages\])/i), out = '';
      for (var p = 0; p < parts.length; p++) {
        if (/^(\{page\}|\[page\])$/i.test(parts[p])) out += '<text:page-number text:select-page="current">1</text:page-number>';
        else if (/^(\{pages\}|\[pages\])$/i.test(parts[p])) out += '<text:page-count>1</text:page-count>';
        else if (parts[p]) out += xmlEscapeText(parts[p]);
      }
      return out;
    }
    var odtPage = pageConfig(inst, 1), odtDims = pageSizeMm(odtPage);
    var odtPageProps = '<style:page-layout-properties fo:page-width="' + mmToCm(odtDims.w) + '" fo:page-height="' + mmToCm(odtDims.h) + '" fo:margin-top="' + mmToCm(odtPage.margins.top) + '" fo:margin-bottom="' + mmToCm(odtPage.margins.bottom) + '" fo:margin-left="' + mmToCm(odtPage.margins.left) + '" fo:margin-right="' + mmToCm(odtPage.margins.right) + '"/>';
    var styles;
    if (inst.headerText || inst.footerText) {
      var odtHeaderBody = odtChromeBody(inst.headerValue || inst.headerText || '', inst, 'Header');
      var odtFooterBody = odtChromeBody(inst.footerValue || inst.footerText || '', inst, 'Footer');
      var odtHeader = odtHeaderBody ? '<style:header>' + odtHeaderBody + '</style:header>' : '';
      var odtFooter = odtFooterBody ? '<style:footer>' + odtFooterBody + '</style:footer>' : '';
      styles = '<?xml version="1.0" encoding="UTF-8"?><office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" office:version="1.2">' +
        '<office:styles><style:style style:name="PageBreak" style:family="paragraph"><style:paragraph-properties fo:break-before="page"/></style:style><style:style style:name="Header" style:family="paragraph"><style:paragraph-properties fo:text-align="center"/><style:text-properties fo:font-size="9pt" fo:color="#666666"/></style:style>' +
        '<style:style style:name="Footer" style:family="paragraph"><style:paragraph-properties fo:text-align="center"/><style:text-properties fo:font-size="9pt" fo:color="#666666"/></style:style>' + odtListStyles() + '</office:styles>' +
        '<office:automatic-styles><style:page-layout style:name="pm1">' + odtPageProps +
        '<style:header-style><style:header-footer-properties fo:min-height="0.8cm" fo:margin-bottom="0.3cm"/></style:header-style>' +
        '<style:footer-style><style:header-footer-properties fo:min-height="0.8cm" fo:margin-top="0.3cm"/></style:footer-style>' +
        '</style:page-layout></office:automatic-styles>' +
        '<office:master-styles><style:master-page style:name="Standard" style:page-layout-name="pm1">' + odtHeader + odtFooter + '</style:master-page></office:master-styles>' +
        '</office:document-styles>';
    } else {
      styles = '<?xml version="1.0" encoding="UTF-8"?><office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" office:version="1.2"><office:styles><style:style style:name="PageBreak" style:family="paragraph"><style:paragraph-properties fo:break-before="page"/></style:style>' + odtListStyles() + '</office:styles><office:automatic-styles><style:page-layout style:name="pm1">' + odtPageProps + '</style:page-layout></office:automatic-styles><office:master-styles><style:master-page style:name="Standard" style:page-layout-name="pm1"/></office:master-styles></office:document-styles>';
    }
    styles = styles.replace('<office:document-styles ', '<office:document-styles xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0" xmlns:svg="urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0" xmlns:xlink="http://www.w3.org/1999/xlink" ');
    var manifest = '<?xml version="1.0" encoding="UTF-8"?><manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.text"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/><manifest:file-entry manifest:full-path="styles.xml" manifest:media-type="text/xml"/>' + manifestMedia + '</manifest:manifest>';
    var entries = [
      { name: 'mimetype', data: utf8Bytes('application/vnd.oasis.opendocument.text') },
      { name: 'content.xml', data: odtContent },
      { name: 'styles.xml', data: styles },
      { name: 'META-INF/manifest.xml', data: manifest }
    ];
    for (var m = 0; m < media.length; m++) entries.push(media[m]);
    return zipStore(entries);
  }
  function importDOCX(files) {
    var xml = new TextDecoder().decode(files['word/document.xml'] || new Uint8Array());
    if (!xml) throw new Error('DOCX has no word/document.xml');
    var dom = new DOMParser().parseFromString(xml, 'application/xml');
    var body = dom.getElementsByTagNameNS('*', 'body')[0];
    if (!body) throw new Error('DOCX has no body');
    var relMap = Object.create(null);
    var relXml = files['word/_rels/document.xml.rels'] ? new TextDecoder().decode(files['word/_rels/document.xml.rels']) : '';
    if (relXml) {
      var relDom = new DOMParser().parseFromString(relXml, 'application/xml');
      var rels = relDom.getElementsByTagNameNS('*', 'Relationship');
      for (var ri = 0; ri < rels.length; ri++) {
        var target = rels[ri].getAttribute('Target') || '';
        relMap[rels[ri].getAttribute('Id')] = target.indexOf('media/') === 0 ? 'word/' + target : target;
      }
    }
    var mediaMap = Object.create(null);
    for (var fname in files) if (Object.prototype.hasOwnProperty.call(files, fname) && fname.indexOf('word/media/') === 0) {
      mediaMap[fname] = dataUrlFromBytes(files[fname], fname);
    }
    var numberingFmt = Object.create(null);
    var numberingFile = files['word/numbering.xml'] ? new TextDecoder().decode(files['word/numbering.xml']) : '';
    if (numberingFile) {
      try {
        var numDom = new DOMParser().parseFromString(numberingFile, 'application/xml');
        var numToAbs = Object.create(null), absFmt = Object.create(null);
        var nums = numDom.getElementsByTagNameNS('*', 'num');
        for (var ni = 0; ni < nums.length; ni++) {
          var aid = nums[ni].getElementsByTagNameNS('*', 'abstractNumId')[0];
          if (aid) numToAbs[nums[ni].getAttribute('w:numId') || nums[ni].getAttribute('numId')] = aid.getAttribute('w:val') || aid.getAttribute('val');
        }
        var abs = numDom.getElementsByTagNameNS('*', 'abstractNum');
        for (var ai = 0; ai < abs.length; ai++) {
          var aid2 = abs[ai].getAttribute('w:abstractNumId') || abs[ai].getAttribute('abstractNumId');
          var levels = abs[ai].getElementsByTagNameNS('*', 'lvl');
          var fmts = {};
          for (var li = 0; li < levels.length; li++) {
            var ilvl = levels[li].getAttribute('w:ilvl') || levels[li].getAttribute('ilvl') || '0';
            var fmtNode = levels[li].getElementsByTagNameNS('*', 'numFmt')[0];
            var fmt = fmtNode ? (fmtNode.getAttribute('w:val') || fmtNode.getAttribute('val')) : 'bullet';
            fmts[ilvl] = fmt;
          }
          absFmt[aid2] = fmts;
        }
        for (var nid in numToAbs) numberingFmt[nid] = absFmt[numToAbs[nid]] || {};
      } catch (e) {}
    }
    function relTarget(id) {
      var target = relMap[id] || '';
      if (target && target.indexOf('http') !== 0 && target.indexOf('word/') !== 0 && target.indexOf('media/') !== 0) target = 'word/' + target;
      return target;
    }
    function textFrom(part) {
      var bytes = files[part];
      if (!bytes) return '';
      var partDom = new DOMParser().parseFromString(new TextDecoder().decode(bytes), 'application/xml');
      var out = '';
      function walk(node) {
        if (!node) return;
        if (node.nodeType === 3) { out += node.nodeValue || ''; return; }
        if (node.nodeType !== 1) return;
        var local = node.localName;
        if (local === 'fldSimple') {
          var instr = node.getAttribute('w:instr') || node.getAttribute('instr') || '';
          if (/NUMPAGES/i.test(instr)) out += '{pages}';
          else if (/PAGE/i.test(instr)) out += '{page}';
          else for (var c = 0; c < node.childNodes.length; c++) walk(node.childNodes[c]);
          return;
        }
        if (local === 'instrText') return;
        if (local === 't') { out += node.textContent || ''; return; }
        for (var i = 0; i < node.childNodes.length; i++) walk(node.childNodes[i]);
      }
      walk(partDom.documentElement);
      return out.replace(/\u00AD/g, '').replace(/\s+/g, ' ').trim();
    }
    var headerText = '', footerText = '', pageCfg = null;
    var sect = body.getElementsByTagNameNS('*', 'sectPr')[0];
    if (sect) {
      var refs = sect.getElementsByTagNameNS('*', 'headerReference');
      if (refs.length) headerText = textFrom(relTarget(refs[0].getAttribute('r:id') || refs[0].getAttributeNS('http://schemas.openxmlformats.org/officeDocument/2006/relationships', 'id')));
      var frefs = sect.getElementsByTagNameNS('*', 'footerReference');
      if (frefs.length) footerText = textFrom(relTarget(frefs[0].getAttribute('r:id') || frefs[0].getAttributeNS('http://schemas.openxmlformats.org/officeDocument/2006/relationships', 'id')));
      var pgSz = sect.getElementsByTagNameNS('*', 'pgSz')[0];
      var pgMar = sect.getElementsByTagNameNS('*', 'pgMar')[0];
      if (pgSz || pgMar) {
        function twipsToMm(value, fallback) {
          if (value == null || value === '') return fallback;
          var n = Number(value);
          if (!isFinite(n) || n < 0) return fallback;
          return Math.round(n * 25.4 / 1440 * 10) / 10;
        }
        var wmm = pgSz ? twipsToMm(pgSz.getAttribute('w:w'), 210) : 210;
        var hmm = pgSz ? twipsToMm(pgSz.getAttribute('w:h'), 297) : 297;
        var shortSide = Math.min(wmm, hmm), longSide = Math.max(wmm, hmm);
        var pageSize = (Math.abs(shortSide - 215.9) <= 8 && Math.abs(longSide - 279.4) <= 12) ? 'letter' : 'a4';
        pageCfg = {
          size: pageSize,
          orientation: wmm > hmm ? 'landscape' : 'portrait',
          margins: {
            top: twipsToMm(pgMar ? pgMar.getAttribute('w:top') : null, 15),
            right: twipsToMm(pgMar ? pgMar.getAttribute('w:right') : null, 17),
            bottom: twipsToMm(pgMar ? pgMar.getAttribute('w:bottom') : null, 15),
            left: twipsToMm(pgMar ? pgMar.getAttribute('w:left') : null, 17)
          }
        };
      }
    }
    function imageFromRun(run) {
      var blips = run.getElementsByTagNameNS('*', 'blip');
      for (var b = 0; b < blips.length; b++) {
        var rid = blips[b].getAttribute('r:embed') || blips[b].getAttributeNS('http://schemas.openxmlformats.org/officeDocument/2006/relationships', 'embed');
        var media = rid ? mediaMap[relMap[rid]] : '';
        if (media) {
          var extent = run.getElementsByTagNameNS('*', 'extent')[0];
          var width = extent ? Math.round(Number(extent.getAttribute('cx') || 0) / 9525) : 0;
          return { type: 'image', src: media, alt: '', width: width ? width + 'px' : '', marks: [] };
        }
      }
      return null;
    }
    function markColor(value) {
      var v = String(value || '').replace(/^#/, '');
      if (!/^[0-9a-f]{6}$/i.test(v) || /^000000$/i.test(v)) return '';
      return '#' + v.toUpperCase();
    }
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
          if (vert) marks.push({ type: (vert.getAttribute('w:val') || '') === 'subscript' ? 'sub' : 'sup' });
          var colorNode = props.getElementsByTagNameNS('*', 'color')[0];
          if (colorNode) { var col = markColor(colorNode.getAttribute('w:val')); if (col) marks.push({ type: 'color', value: col }); }
          var shdNode = props.getElementsByTagNameNS('*', 'shd')[0];
          if (shdNode) { var fill = markColor(shdNode.getAttribute('w:fill')); if (fill) marks.push({ type: 'background', value: fill }); }
          var hl = props.getElementsByTagNameNS('*', 'highlight')[0];
          if (hl) {
            var hlMap = { yellow: '#FFFF00', green: '#00FF00', cyan: '#00FFFF', magenta: '#FF00FF', red: '#FF0000', blue: '#0000FF', darkBlue: '#000080', darkRed: '#800000', gray: '#808080', lightGray: '#C0C0C0' };
            var hlColor = hlMap[hl.getAttribute('w:val') || ''] || '';
            if (hlColor) marks.push({ type: 'background', value: hlColor });
          }
          var sz = props.getElementsByTagNameNS('*', 'sz')[0];
          if (sz) { var pt = Number(sz.getAttribute('w:val') || 0) / 2; if (pt >= 6 && pt <= 96) marks.push({ type: 'size', value: Math.round(pt * 10) / 10 + 'pt' }); }
          var fonts = props.getElementsByTagNameNS('*', 'rFonts')[0];
          if (fonts) {
            var family = fonts.getAttribute('w:ascii') || fonts.getAttribute('w:hAnsi') || '';
            if (/^[A-Za-z0-9 ._-]{1,60}$/.test(family)) marks.push({ type: 'font', value: family });
          }
        }
        var link = null;
        var ancestor = runs[i].parentNode;
        while (ancestor && ancestor !== node) {
          if (ancestor.localName === 'hyperlink') {
            var hid = ancestor.getAttribute('r:id') || ancestor.getAttributeNS('http://schemas.openxmlformats.org/officeDocument/2006/relationships', 'id');
            var href = relTarget(hid);
            if (/^https?:|^mailto:|^tel:/i.test(href)) link = { type: 'link', href: href };
            break;
          }
          ancestor = ancestor.parentNode;
        }
        if (link) marks.push(link);
        var image = imageFromRun(runs[i]);
        if (image) out.push(image);
        var texts = runs[i].getElementsByTagNameNS('*', 't');
        var text = '';
        for (var t = 0; t < texts.length; t++) text += texts[t].textContent;
        if (text) out.push({ type: 'text', text: text, marks: marks });
      }
      return out;
    }
    function paragraphProps(pPr) {
      var props = {};
      if (!pPr) return props;
      var jc = pPr.getElementsByTagNameNS('*', 'jc')[0];
      if (jc) { var val = jc.getAttribute('w:val') || ''; if (val) props.align = val === 'both' ? 'justify' : val; }
      var spacing = pPr.getElementsByTagNameNS('*', 'spacing')[0];
      if (spacing) {
        var before = Number(spacing.getAttribute('w:before') || 0), after = Number(spacing.getAttribute('w:after') || 0), line = Number(spacing.getAttribute('w:line') || 0);
        var rule = spacing.getAttribute('w:lineRule') || 'auto';
        if (before) props.spaceBefore = Math.round(before / 20) + 'pt';
        if (after) props.spaceAfter = Math.round(after / 20) + 'pt';
        if (line) {
          if (rule === 'auto') props.lineHeight = String(Math.round((line / 240) * 100) / 100);
          else {
            props.lineHeight = Math.round(line / 20) + 'pt';
            props.lineRule = rule;
          }
        }
      }
      var ind = pPr.getElementsByTagNameNS('*', 'ind')[0];
      if (ind) {
        var left = Number(ind.getAttribute('w:left') || ind.getAttribute('w:start') || 0);
        if (left) props.indent = Math.round(left / 15) + 'px';
      }
      return props;
    }
    function numberingOf(pPr) {
      if (!pPr) return null;
      var numPr = pPr.getElementsByTagNameNS('*', 'numPr')[0];
      if (!numPr) return null;
      var ilvlNode = numPr.getElementsByTagNameNS('*', 'ilvl')[0];
      var numIdNode = numPr.getElementsByTagNameNS('*', 'numId')[0];
      var numId = numIdNode ? (numIdNode.getAttribute('w:val') || '') : '';
      var level = ilvlNode ? Number(ilvlNode.getAttribute('w:val') || 0) : 0;
      var fmts = numberingFmt[numId] || {};
      var fmt = fmts[String(level)] || fmts['0'] || 'bullet';
      return { level: level, ordered: fmt !== 'bullet' };
    }
    function cellFrom(tc) {
      var tcPr = tc.getElementsByTagNameNS('*', 'tcPr')[0];
      var nested = false;
      for (var tcn = 0; tcn < tc.childNodes.length; tcn++) if (tc.childNodes[tcn].nodeType === 1 && tc.childNodes[tcn].localName === 'tbl') { nested = true; break; }
      var cell = { type: 'table_cell', header: false, colspan: 1, rowspan: 1, children: nested ? [] : inline(tc) };
      if (nested) cell.blocks = blocks(tc);
      if (tcPr) {
        var span = tcPr.getElementsByTagNameNS('*', 'gridSpan')[0];
        if (span) cell.colspan = Math.max(1, Number(span.getAttribute('w:val') || 1));
        var shd = tcPr.getElementsByTagNameNS('*', 'shd')[0];
        if (shd) { var fill = markColor(shd.getAttribute('w:fill')); if (fill) cell.background = fill; }
        var jc = tcPr.getElementsByTagNameNS('*', 'jc')[0];
        if (jc) { var val = jc.getAttribute('w:val') || ''; if (val) cell.align = val === 'both' ? 'justify' : val; }
        var vMerge = tcPr.getElementsByTagNameNS('*', 'vMerge')[0];
        if (vMerge) {
          var vval = vMerge.getAttribute('w:val') || 'continue';
          cell._vmerge = vval === 'restart' ? 'start' : 'continue';
        }
      }
      return cell;
    }
    function finalizeTable(rows) {
      var layout = [], edges = [];
      for (var r = 0; r < rows.length; r++) {
        var col = 0, placed = [];
        for (var c = 0; c < rows[r].cells.length; c++) {
          var cell = rows[r].cells[c];
          var start = col, span = Math.max(1, Number(cell.colspan) || 1);
          placed.push({ cell: cell, start: start });
          if (cell._vmerge !== 'continue') edges.push(start);
          col += span;
        }
        layout.push(placed);
      }
      var keep = [];
      for (var r2 = 0; r2 < rows.length; r2++) {
        var kept = [];
        var placed2 = layout[r2];
        for (var p2 = 0; p2 < placed2.length; p2++) {
          var info = placed2[p2], cell2 = info.cell;
          if (cell2._vmerge === 'continue') {
            var target = null;
            for (var pr = r2 - 1; pr >= 0 && !target; pr--) {
              for (var q = 0; q < layout[pr].length; q++) if (layout[pr][q].start === info.start) { target = layout[pr][q].cell; break; }
            }
            if (target) target.rowspan = (Number(target.rowspan) || 1) + 1;
            continue;
          }
          delete cell2._vmerge;
          kept.push(cell2);
        }
        keep.push(kept);
      }
      for (var r3 = 0; r3 < rows.length; r3++) rows[r3].cells = keep[r3];
    }
    function blocks(parent) {
      var out = [], stack = [];
      function closeLists(toLevel) {
        while (stack.length && stack[stack.length - 1].level > toLevel) stack.pop();
        if (!stack.length) stack = [];
      }
      for (var i = 0; i < parent.childNodes.length; i++) {
        var n = parent.childNodes[i];
        if (n.nodeType !== 1) continue;
        var tag = n.localName;
        if (tag === 'p') {
          var pPr = n.getElementsByTagNameNS('*', 'pPr')[0];
          var heading = 0;
          if (pPr) {
            var style = pPr.getElementsByTagNameNS('*', 'pStyle')[0];
            if (style) { var m = /^Heading(\d)$/i.exec(style.getAttribute('w:val') || ''); if (m) heading = Number(m[1]); }
          }
          var para = { type: heading ? 'heading' : 'paragraph', children: inline(n) };
          if (heading) para.level = heading;
          var props = paragraphProps(pPr);
          for (var pk in props) if (Object.prototype.hasOwnProperty.call(props, pk)) para[pk] = props[pk];
          var num = numberingOf(pPr);
          if (num) {
            closeLists(num.level);
            var type = num.ordered ? 'ordered_list' : 'bullet_list';
            if (!stack.length || stack[stack.length - 1].level < num.level) {
              var list = { type: type, children: [] };
              if (stack.length) {
                var parentList = stack[stack.length - 1].list;
                var lastItem = parentList.children[parentList.children.length - 1];
                if (!lastItem) { lastItem = { type: 'list_item', children: [] }; parentList.children.push(lastItem); }
                lastItem.children.push(list);
              } else out.push(list);
              stack.push({ level: num.level, list: list });
            }
            var current = stack[stack.length - 1].list;
            current.children.push({ type: 'list_item', children: [para] });
          } else {
            stack = [];
            out.push(para);
          }
        } else if (tag === 'tbl') {
          stack = [];
          var rows = [], trs = n.getElementsByTagNameNS('*', 'tr');
          for (var r = 0; r < trs.length; r++) {
            var row = { type: 'table_row', cells: [] };
            var trPr = trs[r].getElementsByTagNameNS('*', 'trPr')[0];
            if (trPr) {
              var th = trPr.getElementsByTagNameNS('*', 'trHeight')[0];
              if (th) { var hv = Number(th.getAttribute('w:val') || 0); if (hv) row.height = Math.round(hv / 15) + 'px'; }
            }
            var headerRow = !!(trPr && trPr.getElementsByTagNameNS('*', 'tblHeader').length);
            for (var c = 0; c < trs[r].childNodes.length; c++) if (trs[r].childNodes[c].nodeType === 1 && trs[r].childNodes[c].localName === 'tc') {
              var importedCell = cellFrom(trs[r].childNodes[c]);
              if (headerRow) importedCell.header = true;
              row.cells.push(importedCell);
            }
            rows.push(row);
          }
          finalizeTable(rows);
          var tableNode = { type: 'table', rows: rows };
          var tblPrEl = n.getElementsByTagNameNS('*', 'tblPr')[0];
          if (tblPrEl) {
            var cellMar = tblPrEl.getElementsByTagNameNS('*', 'tblCellMar')[0];
            if (cellMar) {
              var marLeft = cellMar.getElementsByTagNameNS('*', 'left')[0];
              var marTop = cellMar.getElementsByTagNameNS('*', 'top')[0];
              var marValue = marLeft ? Number(marLeft.getAttribute('w:w') || 0) : (marTop ? Number(marTop.getAttribute('w:w') || 0) : 0);
              if (marValue > 0) tableNode.padding = Math.round(marValue / 15 * 10) / 10 + 'px';
            }
            var tblBorders = tblPrEl.getElementsByTagNameNS('*', 'tblBorders')[0];
            if (tblBorders) {
              var sideNames = ['top', 'right', 'bottom', 'left', 'insideH', 'insideV'];
              var sides = [];
              for (var sb = 0; sb < sideNames.length; sb++) {
                var sideEl = tblBorders.getElementsByTagNameNS('*', sideNames[sb])[0];
                if (!sideEl) continue;
                var sideVal = sideEl.getAttribute('w:val') || '';
                var bsz = Number(sideEl.getAttribute('w:sz') || 0);
                var bcolor = sideEl.getAttribute('w:color') || '';
                if (sideVal !== 'nil') {
                  sides.push(sideNames[sb]);
                  if (!tableNode.borderWidth && bsz > 0) tableNode.borderWidth = Math.max(1, Math.round(bsz / 8)) + 'px';
                  if (!tableNode.border && bcolor && bcolor !== 'auto') tableNode.border = '#' + bcolor.replace(/^#/, '');
                }
              }
              if (sides.length && sides.length < 6) tableNode.borders = sides;
            }
          }
          if (rows.length && rows[0].cells.length && rows[0].cells.every(function (rc) { return rc.header; })) tableNode.header = true;
          out.push(tableNode);
        }
      }
      stack = [];
      return out;
    }
    var doc = { type: 'doc', children: blocks(body) };
    if (headerText) doc.header = headerText;
    if (footerText) doc.footer = footerText;
    if (pageCfg) doc.page = pageCfg;
    return doc;
  }
  function importODT(files) {
    var xml = new TextDecoder().decode(files['content.xml'] || new Uint8Array());
    if (!xml) throw new Error('ODT has no content.xml');
    var dom = new DOMParser().parseFromString(xml, 'application/xml');
    var body = dom.getElementsByTagNameNS('*', 'text')[0];
    if (!body) throw new Error('ODT has no text body');
    var mediaMap = Object.create(null);
    for (var fname in files) if (Object.prototype.hasOwnProperty.call(files, fname) && fname.indexOf('Pictures/') === 0) {
      mediaMap[fname] = dataUrlFromBytes(files[fname], fname);
    }
    var orderedLists = Object.create(null), headerText = '', footerText = '', pageCfg = null;
    var stylesXml = files['styles.xml'] ? new TextDecoder().decode(files['styles.xml']) : '';
    if (stylesXml) {
      try {
        var stylesDom = new DOMParser().parseFromString(stylesXml, 'application/xml');
        var listStyles = stylesDom.getElementsByTagNameNS('*', 'list-style');
        for (var ls = 0; ls < listStyles.length; ls++) {
          var name = listStyles[ls].getAttribute('style:name') || listStyles[ls].getAttribute('name') || '';
          if (!name) continue;
          orderedLists[name] = listStyles[ls].getElementsByTagNameNS('*', 'list-level-style-number').length > 0;
        }
        function masterText(kind) {
          var master = stylesDom.getElementsByTagNameNS('*', 'master-page')[0];
          if (!master) return '';
          var part = master.getElementsByTagNameNS('*', kind)[0];
          if (!part) return '';
          var out = '';
          function walk(node) {
            if (!node) return;
            if (node.nodeType === 3) { out += node.nodeValue || ''; return; }
            if (node.nodeType !== 1) return;
            if (node.localName === 'page-number') { out += '{page}'; return; }
            if (node.localName === 'page-count') { out += '{pages}'; return; }
            if (node.localName === 's') { out += ' '; return; }
            if (node.localName === 'tab') { out += ' '; return; }
            for (var c = 0; c < node.childNodes.length; c++) walk(node.childNodes[c]);
          }
          walk(part);
          return out.replace(/\s+/g, ' ').trim();
        }
        headerText = masterText('header');
        footerText = masterText('footer');
        var pageProps = stylesDom.getElementsByTagNameNS('*', 'page-layout-properties')[0];
        if (pageProps) {
          function odtLengthMm(value, fallback) {
            var m = /^\s*([\d.]+)\s*(mm|cm|in|pt|pc)?\s*$/i.exec(String(value || ''));
            if (!m) return fallback;
            var n = parseFloat(m[1]);
            if (!isFinite(n) || n < 0) return fallback;
            var unit = (m[2] || 'mm').toLowerCase();
            if (unit === 'cm') return n * 10;
            if (unit === 'in') return n * 25.4;
            if (unit === 'pt') return n * 25.4 / 72;
            if (unit === 'pc') return n * 25.4 / 6;
            return n;
          }
          var odtW = odtLengthMm(pageProps.getAttribute('fo:page-width'), 210);
          var odtH = odtLengthMm(pageProps.getAttribute('fo:page-height'), 297);
          var odtShort = Math.min(odtW, odtH), odtLong = Math.max(odtW, odtH);
          pageCfg = {
            size: (Math.abs(odtShort - 215.9) <= 8 && Math.abs(odtLong - 279.4) <= 12) ? 'letter' : 'a4',
            orientation: odtW > odtH ? 'landscape' : 'portrait',
            margins: {
              top: odtLengthMm(pageProps.getAttribute('fo:margin-top'), 15),
              right: odtLengthMm(pageProps.getAttribute('fo:margin-right'), 17),
              bottom: odtLengthMm(pageProps.getAttribute('fo:margin-bottom'), 15),
              left: odtLengthMm(pageProps.getAttribute('fo:margin-left'), 17)
            }
          };
        }
      } catch (e) {}
    }
    function styleMarks(style) {
      var marks = [];
      if (/font-weight\s*:\s*(bold|[6-9]00)/i.test(style)) marks.push({ type: 'bold' });
      if (/font-style\s*:\s*italic/i.test(style)) marks.push({ type: 'italic' });
      if (/text-decoration[^;]*underline/i.test(style)) marks.push({ type: 'underline' });
      if (/line-through/i.test(style)) marks.push({ type: 'strike' });
      if (/vertical-align\s*:\s*super/i.test(style)) marks.push({ type: 'sup' });
      if (/vertical-align\s*:\s*sub/i.test(style)) marks.push({ type: 'sub' });
      var color = /(?:^|;)\s*color\s*:\s*([^;]+)/i.exec(style);
      if (color && /^#|rgb/i.test(color[1].trim())) marks.push({ type: 'color', value: color[1].trim() });
      var bg = /background(?:-color)?\s*:\s*([^;]+)/i.exec(style);
      if (bg && /^#|rgb/i.test(bg[1].trim())) marks.push({ type: 'background', value: bg[1].trim() });
      var size = /font-size\s*:\s*([\d.]+)(pt|px)/i.exec(style);
      if (size) marks.push({ type: 'size', value: size[1] + (size[2].toLowerCase() === 'px' ? 'px' : 'pt') });
      var family = /font-family\s*:\s*([^;]+)/i.exec(style);
      if (family) marks.push({ type: 'font', value: family[1].trim() });
      return marks;
    }
    function inline(node) {
      var out = [];
      for (var i = 0; i < node.childNodes.length; i++) {
        var n = node.childNodes[i];
        if (n.nodeType === 3 && n.nodeValue) out.push({ type: 'text', text: n.nodeValue, marks: [] });
        else if (n.nodeType === 1 && n.localName === 'frame') {
          var image = n.getElementsByTagNameNS('*', 'image')[0];
          var href = image ? (image.getAttribute('xlink:href') || image.getAttributeNS('http://www.w3.org/1999/xlink', 'href')) : '';
          var src = href ? mediaMap[href] : '';
          if (src) out.push({ type: 'image', src: src, alt: '', width: n.getAttribute('svg:width') || '', height: n.getAttribute('svg:height') || '' });
        }
        else if (n.nodeType === 1 && n.localName === 'span') {
          var child = inline(n);
          var marks = styleMarks(n.getAttribute('style') || '');
          for (var c = 0; c < child.length; c++) child[c].marks = (child[c].marks || []).concat(marks);
          out = out.concat(child);
        } else if (n.nodeType === 1 && n.localName === 'line-break') {
          out.push({ type: 'break' });
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
        else if (tag === 'p') {
          var align = /text-align\s*:\s*(left|center|right|justify)/i.exec(n.getAttribute('style') || '');
          var p = { type: 'paragraph', children: inline(n) };
          if (align) p.align = align[1].toLowerCase();
          out.push(p);
        }
        else if (tag === 'list') {
          var styleName = n.getAttribute('text:style-name') || n.getAttribute('style-name') || '';
          var list = { type: orderedLists[styleName] ? 'ordered_list' : 'bullet_list', children: [] };
          for (var c = 0; c < n.childNodes.length; c++) if (n.childNodes[c].nodeType === 1 && n.childNodes[c].localName === 'list-item') list.children.push({ type: 'list_item', children: blocks(n.childNodes[c]) });
          out.push(list);
        } else if (tag === 'table') {
          var table = { type: 'table', rows: [] };
          var trs = n.getElementsByTagNameNS('*', 'table-row');
          for (var r = 0; r < trs.length; r++) {
            var row = { type: 'table_row', cells: [] };
            for (var cc = 0; cc < trs[r].childNodes.length; cc++) if (trs[r].childNodes[cc].nodeType === 1 && trs[r].childNodes[cc].localName === 'table-cell') {
              var tc = trs[r].childNodes[cc];
              var spanAttr = tc.getAttribute('table:number-columns-spanned') || tc.getAttribute('number-columns-spanned') || '1';
              var nestedOdt = false;
              for (var otcn = 0; otcn < tc.childNodes.length; otcn++) if (tc.childNodes[otcn].nodeType === 1 && tc.childNodes[otcn].localName === 'table') { nestedOdt = true; break; }
              var cell = { type: 'table_cell', header: false, colspan: Math.max(1, Number(spanAttr) || 1), rowspan: 1, children: nestedOdt ? [] : inline(tc) };
              if (nestedOdt) cell.blocks = blocks(tc);
              var bg = /background-color\s*:\s*(#[0-9a-f]{6})/i.exec(tc.getAttribute('style') || '');
              if (bg) cell.background = bg[1];
              row.cells.push(cell);
            }
            table.rows.push(row);
          }
          out.push(table);
        }
      }
      return out;
    }
    var doc = { type: 'doc', children: blocks(body) };
    if (headerText) doc.header = headerText;
    if (footerText) doc.footer = footerText;
    if (pageCfg) doc.page = pageCfg;
    return doc;
  }
  function pdfEscape(text) { return String(text == null ? '' : text).replace(/\\/g, '\\\\').replace(/\(/g, '\\(').replace(/\)/g, '\\)'); }
  function pdfBytes(text) { var out = []; for (var i = 0; i < text.length; i++) { var c = text.charCodeAt(i); out.push(c & 255); } return out; }
  function pdfNumber(n) { var v = Number(n); if (!isFinite(v)) v = 0; return Math.round(v * 100) / 100; }
  function pdfColor(value) {
    var v = String(value == null ? '' : value).trim();
    var hex = /^#([0-9a-f]{6})$/i.exec(v);
    if (hex) return [parseInt(hex[1].slice(0, 2), 16) / 255, parseInt(hex[1].slice(2, 4), 16) / 255, parseInt(hex[1].slice(4, 6), 16) / 255];
    var rgb = /^rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/i.exec(v);
    if (rgb) return [Math.min(255, Number(rgb[1])) / 255, Math.min(255, Number(rgb[2])) / 255, Math.min(255, Number(rgb[3])) / 255];
    return null;
  }
  function pdfFill(value) {
    var c = pdfColor(value) || [0, 0, 0];
    return pdfNumber(c[0]) + ' ' + pdfNumber(c[1]) + ' ' + pdfNumber(c[2]) + ' rg';
  }
  function pdfStroke(value) {
    var c = pdfColor(value) || [0, 0, 0];
    return pdfNumber(c[0]) + ' ' + pdfNumber(c[1]) + ' ' + pdfNumber(c[2]) + ' RG';
  }
  var PDF_WINANSI = { 0x20AC: 128, 0x201A: 130, 0x0192: 131, 0x201E: 132, 0x2026: 133, 0x2020: 134, 0x2021: 135, 0x02C6: 136, 0x2030: 137, 0x0160: 138, 0x2039: 139, 0x0152: 140, 0x017D: 142, 0x2018: 145, 0x2019: 146, 0x201C: 147, 0x201D: 148, 0x2022: 149, 0x2013: 150, 0x2014: 151, 0x02DC: 152, 0x2122: 153, 0x0161: 154, 0x203A: 155, 0x0153: 156, 0x017E: 158, 0x0178: 159, 0x00A0: 32, 0x2006: 32, 0x2009: 32, 0x200A: 32, 0x2007: 32, 0x2002: 32, 0x2003: 32 };
  function pdfEncode(text) {
    var src = String(text == null ? '' : text), out = '';
    for (var i = 0; i < src.length; i++) {
      var code = src.charCodeAt(i);
      if (code === 9) { out += ' '; continue; }
      if (code < 32) { out += ' '; continue; }
      if (code < 127) { out += src.charAt(i); continue; }
      var mapped = PDF_WINANSI[code];
      if (mapped == null && code >= 160 && code <= 255) mapped = code;
      out += String.fromCharCode(mapped == null ? 63 : mapped);
    }
    return out;
  }
  function inflateZlib(data) {
    if (!data || data.length < 2) return null;
    if ((data[0] & 15) !== 8) return null;
    var pos = 2;
    if (data[1] & 32) pos += 4;
    var bitPos = 0;
    function bits(n) {
      var v = 0;
      for (var i = 0; i < n; i++) {
        var index = pos + (bitPos >> 3);
        if (index >= data.length) throw new Error('inflate eof');
        v |= ((data[index] >> (bitPos & 7)) & 1) << i;
        bitPos++;
      }
      return v;
    }
    function alignByte() { bitPos = (bitPos + 7) & ~7; }
    function build(lengths) {
      var counts = new Array(16).fill(0), offs = new Array(16).fill(0), symbols = [];
      for (var i = 0; i < lengths.length; i++) counts[lengths[i]]++;
      counts[0] = 0;
      for (var k = 1; k < 16; k++) offs[k] = offs[k - 1] + counts[k - 1];
      for (var j = 0; j < lengths.length; j++) if (lengths[j]) symbols[offs[lengths[j]]++] = j;
      return { counts: counts, symbols: symbols };
    }
    function decode(h) {
      var code = 0, first = 0, index = 0;
      for (var len = 1; len < 16; len++) {
        code |= bits(1);
        var count = h.counts[len];
        if (code - first < count) return h.symbols[index + (code - first)];
        index += count; first += count; first <<= 1; code <<= 1;
      }
      return -1;
    }
    var lenBase = [3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258];
    var lenExtra = [0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0];
    var distBase = [1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577];
    var distExtra = [0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13];
    var out = [];
    function block(fixed) {
      var lit, dist;
      if (fixed) {
        var ll = [], dl = [];
        for (var i = 0; i < 288; i++) ll.push(i < 144 ? 8 : i < 256 ? 9 : i < 280 ? 7 : 8);
        for (var d = 0; d < 30; d++) dl.push(5);
        lit = build(ll); dist = build(dl);
      } else {
        var hlit = bits(5) + 257, hdist = bits(5) + 1, hclen = bits(4) + 4;
        var order = [16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15];
        var cl = new Array(19).fill(0);
        for (var c = 0; c < hclen; c++) cl[order[c]] = bits(3);
        var clh = build(cl), lens = [];
        while (lens.length < hlit + hdist) {
          var sym = decode(clh);
          if (sym < 16) lens.push(sym);
          else if (sym === 16) { var rep = 3 + bits(2), prev = lens[lens.length - 1] || 0; for (var r1 = 0; r1 < rep; r1++) lens.push(prev); }
          else if (sym === 17) { var rep2 = 3 + bits(3); for (var r2 = 0; r2 < rep2; r2++) lens.push(0); }
          else if (sym === 18) { var rep3 = 11 + bits(7); for (var r3 = 0; r3 < rep3; r3++) lens.push(0); }
          else return;
        }
        lit = build(lens.slice(0, hlit));
        dist = build(lens.slice(hlit, hlit + hdist));
      }
      while (true) {
        var symbol = decode(lit);
        if (symbol < 0) return;
        if (symbol < 256) out.push(symbol);
        else if (symbol === 256) return;
        else {
          var li = symbol - 257;
          if (li >= lenBase.length) return;
          var length = lenBase[li] + bits(lenExtra[li]);
          var ds = decode(dist);
          if (ds < 0 || ds >= distBase.length) return;
          var distance = distBase[ds] + bits(distExtra[ds]);
          var start = out.length - distance;
          if (start < 0) return;
          for (var k = 0; k < length; k++) out.push(out[start + k]);
        }
      }
    }
    var last = 0;
    try {
      while (!last) {
        last = bits(1);
        var type = bits(2);
        if (type === 0) {
          alignByte();
          var at = pos + (bitPos >> 3);
          if (at + 4 > data.length) break;
          var len = data[at] | (data[at + 1] << 8);
          at += 4;
          for (var i = 0; i < len; i++) out.push(data[at + i]);
          pos = at + len; bitPos = 0;
        } else if (type === 1) block(true);
        else if (type === 2) block(false);
        else break;
      }
    } catch (e) { return null; }
    return new Uint8Array(out);
  }
  function pngImage(bytes) {
    var pos = 8, width = 0, height = 0, depth = 0, colorType = 0, interlace = 0, plte = null, trns = null, idat = [];
    while (pos + 8 <= bytes.length) {
      var len = ((bytes[pos] << 24) | (bytes[pos + 1] << 16) | (bytes[pos + 2] << 8) | bytes[pos + 3]) >>> 0;
      var type = String.fromCharCode(bytes[pos + 4], bytes[pos + 5], bytes[pos + 6], bytes[pos + 7]);
      var data = bytes.subarray(pos + 8, pos + 8 + len);
      if (type === 'IHDR') { width = (data[0] << 24 | data[1] << 16 | data[2] << 8 | data[3]) >>> 0; height = (data[4] << 24 | data[5] << 16 | data[6] << 8 | data[7]) >>> 0; depth = data[8]; colorType = data[9]; interlace = data[12]; }
      else if (type === 'PLTE') plte = data;
      else if (type === 'tRNS') trns = data;
      else if (type === 'IDAT') idat.push(data);
      else if (type === 'IEND') break;
      pos += 12 + len;
    }
    if (!width || !height || depth !== 8 || interlace) return null;
    var channels = colorType === 0 ? 1 : colorType === 2 ? 3 : colorType === 3 ? 1 : colorType === 4 ? 2 : colorType === 6 ? 4 : 0;
    if (!channels) return null;
    var total = 0;
    for (var i = 0; i < idat.length; i++) total += idat[i].length;
    var packed = new Uint8Array(total), at = 0;
    for (var j = 0; j < idat.length; j++) { packed.set(idat[j], at); at += idat[j].length; }
    var raw = inflateZlib(packed);
    if (!raw) return null;
    var stride = width * channels;
    if (raw.length < (stride + 1) * height) return null;
    var rgb = new Uint8Array(width * height * 3);
    var prev = new Uint8Array(stride);
    var off = 0;
    for (var y = 0; y < height; y++) {
      var filter = raw[off++];
      var cur = new Uint8Array(stride);
      for (var x = 0; x < stride; x++) {
        var a = x >= channels ? cur[x - channels] : 0;
        var b = prev[x];
        var c = x >= channels ? prev[x - channels] : 0;
        var value = raw[off + x];
        if (filter === 1) value = (value + a) & 255;
        else if (filter === 2) value = (value + b) & 255;
        else if (filter === 3) value = (value + ((a + b) >> 1)) & 255;
        else if (filter === 4) {
          var pp = a + b - c, pa = Math.abs(pp - a), pb = Math.abs(pp - b), pc = Math.abs(pp - c);
          value = (value + (pa <= pb && pa <= pc ? a : pb <= pc ? b : c)) & 255;
        }
        cur[x] = value;
      }
      off += stride;
      for (var px = 0; px < width; px++) {
        var o = (y * width + px) * 3;
        if (colorType === 3) {
          var idx = cur[px];
          if (plte && idx * 3 + 2 < plte.length) { rgb[o] = plte[idx * 3]; rgb[o + 1] = plte[idx * 3 + 1]; rgb[o + 2] = plte[idx * 3 + 2]; }
          if (trns && idx < trns.length && trns[idx] === 0) { rgb[o] = 255; rgb[o + 1] = 255; rgb[o + 2] = 255; }
        } else if (colorType === 0 || colorType === 4) {
          var gray = cur[px * channels], alpha = colorType === 4 ? cur[px * 2 + 1] : 255;
          var g = Math.round((gray * alpha + 255 * (255 - alpha)) / 255);
          rgb[o] = rgb[o + 1] = rgb[o + 2] = g;
        } else if (colorType === 2) {
          rgb[o] = cur[px * 3]; rgb[o + 1] = cur[px * 3 + 1]; rgb[o + 2] = cur[px * 3 + 2];
        } else {
          var al = cur[px * 4 + 3];
          rgb[o] = Math.round((cur[px * 4] * al + 255 * (255 - al)) / 255);
          rgb[o + 1] = Math.round((cur[px * 4 + 1] * al + 255 * (255 - al)) / 255);
          rgb[o + 2] = Math.round((cur[px * 4 + 2] * al + 255 * (255 - al)) / 255);
        }
      }
      prev = cur;
    }
    return { width: width, height: height, data: rgb };
  }
  function jpegInfo(bytes) {
    var i = 2;
    while (i + 9 < bytes.length) {
      if (bytes[i] !== 0xFF) { i++; continue; }
      var marker = bytes[i + 1];
      if (marker >= 0xC0 && marker <= 0xCF && marker !== 0xC4 && marker !== 0xC8 && marker !== 0xCC) {
        return { width: (bytes[i + 7] << 8) | bytes[i + 8], height: (bytes[i + 5] << 8) | bytes[i + 6], components: bytes[i + 9] };
      }
      if (marker === 0xD8 || marker === 0xD9 || (marker >= 0xD0 && marker <= 0xD7)) { i += 2; continue; }
      var len = (bytes[i + 2] << 8) | bytes[i + 3];
      if (len < 2) break;
      i += 2 + len;
    }
    return null;
  }
  function binaryString(bytes) {
    var out = '';
    for (var i = 0; i < bytes.length; i += 8192) out += String.fromCharCode.apply(null, bytes.subarray(i, i + 8192));
    return out;
  }
  var PDF_FONTS = ['Helvetica', 'Helvetica-Bold', 'Helvetica-Oblique', 'Helvetica-BoldOblique', 'Times-Roman', 'Times-Bold', 'Times-Italic', 'Times-BoldItalic', 'Courier', 'Courier-Bold'];
  var PDF_FONT_KEYS = {
    helvetica: { normal: 1, bold: 2, italic: 3, bolditalic: 4 },
    times: { normal: 5, bold: 6, italic: 7, bolditalic: 8 },
    courier: { normal: 9, bold: 10, italic: 9, bolditalic: 10 }
  };
  var PDF_EMB_KEYS = ['sans-regular', 'sans-bold', 'sans-italic', 'sans-bolditalic', 'serif-regular', 'serif-bold', 'serif-italic', 'serif-bolditalic'];
  var PDF_EMB_NAMES = {};
  for (var pe = 0; pe < PDF_EMB_KEYS.length; pe++) PDF_EMB_NAMES[PDF_EMB_KEYS[pe]] = 'E' + pe;
  var PDF_EMB_CACHE = {};
  var pdfEmbUse = {};
  function pdfFontSlot(family, bold, italic) { return bold && italic ? 'bolditalic' : bold ? 'bold' : italic ? 'italic' : 'regular'; }
  function parsePdfFont(base64, key) {
    if (PDF_EMB_CACHE[key] !== undefined) return PDF_EMB_CACHE[key];
    try {
      var binary = atob(base64), bytes = new Uint8Array(binary.length);
      for (var i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
      function u16(o) { return (bytes[o] << 8) | bytes[o + 1]; }
      function i16(o) { var v = u16(o); return v >= 32768 ? v - 65536 : v; }
      function u32(o) { return ((bytes[o] << 24) | (bytes[o + 1] << 16) | (bytes[o + 2] << 8) | bytes[o + 3]) >>> 0; }
      function tag(o) { return String.fromCharCode(bytes[o], bytes[o + 1], bytes[o + 2], bytes[o + 3]); }
      var tables = {}, numTables = u16(4);
      for (var t = 0; t < numTables; t++) {
        var rec = 12 + t * 16;
        tables[tag(rec)] = { offset: u32(rec + 8), length: u32(rec + 12) };
      }
      if (!tables.head || !tables.cmap || !tables.hhea || !tables.hmtx) { PDF_EMB_CACHE[key] = null; return null; }
      var unitsPerEm = u16(tables.head.offset + 18) || 1000;
      var xMin = i16(tables.head.offset + 36), yMin = i16(tables.head.offset + 38);
      var xMax = i16(tables.head.offset + 40), yMax = i16(tables.head.offset + 42);
      var ascent = i16(tables.hhea.offset + 4), descent = i16(tables.hhea.offset + 6);
      var numHMetrics = u16(tables.hhea.offset + 34);
      var numGlyphs = tables.maxp ? u16(tables.maxp.offset + 4) : 0;
      var widths = [];
      for (var h = 0; h < numHMetrics; h++) widths.push(u16(tables.hmtx.offset + h * 4));
      if (numGlyphs > numHMetrics) {
        var lastWidth = widths[widths.length - 1] || 0;
        for (var g = numHMetrics; g < numGlyphs; g++) widths.push(lastWidth);
      }
      var cmapOff = tables.cmap.offset, numSub = u16(cmapOff + 2), best = 0, bestScore = -1;
      for (var s2 = 0; s2 < numSub; s2++) {
        var sub = cmapOff + 4 + s2 * 8;
        var pid = u16(sub), eid = u16(sub + 2), off = u32(sub + 4);
        var fmt = u16(cmapOff + off);
        var score = 0;
        if (fmt === 12) score = 3;
        else if (fmt === 4) score = 2;
        else continue;
        if (pid === 3 && eid === 10) score += 2;
        if (pid === 3 && eid === 1) score += 1;
        if (pid === 0) score += 1;
        if (score > bestScore) { bestScore = score; best = cmapOff + off; }
      }
      var format = best ? u16(best) : 0;
      var segCount = 0, endCodes = null, startCodes = null, idDelta = null, idRangeOffset = null, glyphIdArrayOff = 0, groups = null;
      if (format === 4) {
        segCount = u16(best + 6) / 2;
        endCodes = []; startCodes = []; idDelta = []; idRangeOffset = [];
        var endOff = best + 14, startOff = endOff + segCount * 2 + 2;
        var deltaOff = startOff + segCount * 2, roOff = deltaOff + segCount * 2;
        for (var q = 0; q < segCount; q++) {
          endCodes.push(u16(endOff + q * 2));
          startCodes.push(u16(startOff + q * 2));
          idDelta.push(i16(deltaOff + q * 2));
          idRangeOffset.push(u16(roOff + q * 2));
        }
        glyphIdArrayOff = roOff + segCount * 2;
      } else if (format === 12) {
        var nGroups = u32(best + 12), gOff = best + 16;
        groups = [];
        for (var gr = 0; gr < nGroups; gr++) groups.push({ start: u32(gOff + gr * 12), end: u32(gOff + gr * 12 + 4), gid: u32(gOff + gr * 12 + 8) });
      } else { PDF_EMB_CACHE[key] = null; return null; }
      function gidFor(code) {
        if (groups) {
          for (var gg = 0; gg < groups.length; gg++) {
            var grp = groups[gg];
            if (code >= grp.start && code <= grp.end) return grp.gid + (code - grp.start);
          }
          return 0;
        }
        for (var si = 0; si < segCount; si++) {
          if (code > endCodes[si]) continue;
          if (code < startCodes[si]) return 0;
          if (idRangeOffset[si] === 0) return (code + idDelta[si]) & 0xFFFF;
          var idx = glyphIdArrayOff + si * 2 + idRangeOffset[si] + (code - startCodes[si]) * 2;
          var gid = u16(idx);
          return gid ? (gid + idDelta[si]) & 0xFFFF : 0;
        }
        return 0;
      }
      var font = { key: key, bytes: bytes, unitsPerEm: unitsPerEm, xMin: xMin, yMin: yMin, xMax: xMax, yMax: yMax, ascent: ascent, descent: descent, widths: widths, gidFor: gidFor };
      PDF_EMB_CACHE[key] = font;
      return font;
    } catch (e) { PDF_EMB_CACHE[key] = null; return null; }
  }
  function pdfFontParts(family, bold, italic) {
    var slot = pdfFontSlot(family, bold, italic);
    var embFamily = family === 'times' ? 'serif' : family === 'helvetica' ? 'sans' : '';
    var key = embFamily ? embFamily + '-' + slot : '';
    if (key && window.SUI_PDF_FONTS && window.SUI_PDF_FONTS[embFamily] && window.SUI_PDF_FONTS[embFamily][slot]) {
      var emb = parsePdfFont(window.SUI_PDF_FONTS[embFamily][slot], key);
      if (emb) {
        var std = PDF_FONT_KEYS[family][slot === 'bold' ? 'bold' : slot === 'italic' ? 'italic' : slot === 'bolditalic' ? 'bolditalic' : 'normal'];
        return { name: PDF_EMB_NAMES[key], key: key, family: family, bold: bold, italic: italic, emb: emb, stdName: 'F' + std };
      }
    }
    var id = PDF_FONT_KEYS[family][slot === 'bold' ? 'bold' : slot === 'italic' ? 'italic' : slot === 'bolditalic' ? 'bolditalic' : 'normal'];
    return { id: id, name: 'F' + id, family: family, bold: bold, italic: italic, emb: null };
  }
  function registerEmbText(font, text) {
    if (!font || !font.emb || !font.key) return;
    var use = pdfEmbUse[font.key] || (pdfEmbUse[font.key] = {});
    for (var i = 0; i < text.length; i++) {
      var gid = font.emb.gidFor(text.charCodeAt(i));
      if (gid) use[gid] = text.charCodeAt(i);
    }
  }
  function pdfGids(text, emb) {
    var hex = '';
    for (var i = 0; i < text.length; i++) {
      var gid = emb.gidFor(text.charCodeAt(i)) || 0;
      hex += ('0000' + gid.toString(16)).slice(-4);
    }
    return hex;
  }
  function pdfHex4(value) { return ('0000' + (Number(value) || 0).toString(16)).slice(-4); }
  function pdfToUnicode(name, used) {
    var gids = Object.keys(used), body = '';
    for (var i = 0; i < gids.length; i += 100) {
      var chunk = gids.slice(i, i + 100);
      body += chunk.length + ' beginbfchar\n';
      for (var j = 0; j < chunk.length; j++) body += '<' + pdfHex4(chunk[j]) + '> <' + pdfHex4(used[chunk[j]]) + '>\n';
      body += 'endbfchar\n';
    }
    return '/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n/CMapName /' + name + ' def\n/CMapType 2 def\n1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n' + body + 'endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend';
  }
  function pdfFont(marks) {
    var bold = false, italic = false, family = 'helvetica';
    for (var i = 0; i < (marks || []).length; i++) {
      var mark = marks[i];
      if (mark.type === 'bold') bold = true;
      else if (mark.type === 'italic') italic = true;
      else if (mark.type === 'font') {
        var value = String(mark.value || '').toLowerCase();
        if (/courier|monospace|consolas/.test(value)) family = 'courier';
        else if (/times|georgia|palatino|garamond|cambria|serif/.test(value) && !/sans-?serif/.test(value)) family = 'times';
        else family = 'helvetica';
      }
    }
    return pdfFontParts(family, bold, italic);
  }
  function pdfSize(marks) {
    var size = 11;
    for (var i = 0; i < (marks || []).length; i++) {
      var mark = marks[i];
      if (mark.type === 'size') {
        var v = String(mark.value || '');
        var num = parseFloat(v);
        if (isFinite(num) && num > 0) size = /px/i.test(v) ? num * 0.75 : num;
      }
    }
    return size;
  }
  var PDF_WIDTHS = {
    helvetica: [278,278,355,556,556,889,667,191,333,333,389,584,278,333,278,278,556,556,556,556,556,556,556,556,556,556,278,278,584,584,584,556,1015,667,667,722,722,667,611,778,722,278,500,667,556,833,722,778,667,778,722,667,611,722,667,944,667,667,611,278,278,278,469,556,333,556,556,500,556,556,278,556,556,222,222,500,222,833,556,556,556,556,333,500,278,556,500,722,500,500,500,334,260,334,584],
    times: [250,333,408,500,500,833,778,180,333,333,500,564,250,333,250,278,500,500,500,500,500,500,500,500,500,500,278,278,564,564,564,444,921,722,667,667,722,611,556,722,722,333,389,722,611,889,722,722,556,722,667,556,611,722,722,944,722,722,611,333,278,333,469,500,333,444,500,444,500,444,333,500,500,278,278,500,278,778,500,500,500,500,333,389,278,500,500,722,500,500,444,480,200,480,541]
  };
  var PDF_ACCENTS = '\u00e0\u00e1\u00e2\u00e3\u00e4\u00e5\u00e7\u00e8\u00e9\u00ea\u00eb\u00ec\u00ed\u00ee\u00ef\u00f1\u00f2\u00f3\u00f4\u00f5\u00f6\u00f9\u00fa\u00fb\u00fc\u00fd\u00ff\u00c0\u00c1\u00c2\u00c3\u00c4\u00c5\u00c7\u00c8\u00c9\u00ca\u00cb\u00cc\u00cd\u00ce\u00cf\u00d1\u00d2\u00d3\u00d4\u00d5\u00d6\u00d9\u00da\u00db\u00dc\u00dd';
  var PDF_BASE_LETTERS = 'aaaaaaceeeeiiiinooooouuuuyyAAAAAACEEEEIIIINOOOOOUUUUY';
  function pdfCharWidth(ch, font) {
    if (ch === '\u00AD') return 0;
    if (font.emb) {
      var egid = font.emb.gidFor(ch.charCodeAt(0));
      if (!egid) return 0.5;
      return (font.emb.widths[egid] || 0) / font.emb.unitsPerEm;
    }
    if (font.family === 'courier') return 0.6;
    var code = ch.charCodeAt(0);
    var table = PDF_WIDTHS[font.family] || PDF_WIDTHS.helvetica;
    var width = 0;
    if (code >= 32 && code <= 126) width = table[code - 32];
    else {
      var accentAt = PDF_ACCENTS.indexOf(ch);
      if (accentAt >= 0) width = table[PDF_BASE_LETTERS.charCodeAt(accentAt) - 32];
      else if (code === 0x20AC) width = 556;
      else if (code === 0x2022) width = 350;
      else if (code === 0x2013 || code === 0x2014) width = 500;
      else if (code === 0x2026) width = 1000;
      else width = 500;
    }
    return (width / 1000) * (font.bold ? 1.03 : 1);
  }
  function pdfTextWidth(text, font, size) {
    var w = 0;
    for (var i = 0; i < text.length; i++) w += pdfCharWidth(text.charAt(i), font);
    return w * size;
  }
  function exportPDF(inst) {
    hyphenationLanguage = documentLanguage(inst);
    var liveEditor = inst.root && inst.root.querySelector('.sui-rich-editor');
    if (liveEditor) inst.editor = liveEditor;
    var pageCfg = pageConfig(inst, 1), pageDims = pageSizeMm(pageCfg);
    var PAGE_W = Math.round(mmToPt(pageDims.w)), PAGE_H = Math.round(mmToPt(pageDims.h));
    var MARGIN = mmToPt(pageCfg.margins.left), MARGIN_R = mmToPt(pageCfg.margins.right);
    var MARGIN_T = mmToPt(pageCfg.margins.top), MARGIN_B = mmToPt(pageCfg.margins.bottom);
    var CONTENT_W = PAGE_W - MARGIN - MARGIN_R;
    var headerText = inst.headerText || '', footerText = inst.footerText || '';
    var headerValue = inst.headerValue != null ? inst.headerValue : headerText;
    var footerValue = inst.footerValue != null ? inst.footerValue : footerText;
    var hasHeader = chromeHasContent(headerValue);
    var hasFooter = chromeHasContent(footerValue);
    var HEADER_H = hasHeader ? Math.max(30, (inst.headerH || 0) * 0.75) : 0;
    var FOOTER_H = hasFooter ? Math.max(30, (inst.footerH || 0) * 0.75) : 0;
    var headerSpace = hasHeader ? Math.max(MARGIN_T, HEADER_H + 12) : MARGIN_T;
    var footerSpace = hasFooter ? Math.max(MARGIN_B, FOOTER_H + 12) : MARGIN_B;
    var START_Y = headerSpace, BOTTOM_LIMIT = PAGE_H - footerSpace;
    pdfEmbUse = {};
    var pages = [], ops = [], y = START_Y, pageSeq = 0;
    var images = [], imageCache = {};
    var doc = inst.doc || { children: [] };
    var domTables = inst.editor ? inst.editor.querySelectorAll('table') : [];
    var tableSeq = 0;
    function chromeRuns(nodes) {
      var mapped = [];
      for (var i = 0; i < (nodes || []).length; i++) {
        var n = nodes[i];
        if (n && n.type === 'text') mapped.push({ type: 'text', text: applyPageTokens(n.text || '', 1, 1, false, inst), marks: n.marks || [] });
        else if (n) mapped.push(n);
      }
      return runsOf(mapped, [{ type: 'size', value: '9' }]);
    }
    function drawChromeRuns(runs, top, maxWidth, maxHeight) {
      var total = 0;
      for (var i = 0; i < runs.length; i++) {
        var run = runs[i];
        if (run.type === 'run') total += pdfTextWidth(run.text.replace(/__SUI_(PAGE|PAGES)__/g, '00'), run.font, run.fontSize);
        else if (run.type === 'image') {
          var sizedW = imageSize(run.node, maxWidth);
          if (sizedW) {
            var fit = Math.min(1, (maxHeight - 2) / sizedW.height);
            total += sizedW.width * fit + 4;
          }
        }
      }
      var x = MARGIN + Math.max(0, (CONTENT_W - Math.min(total, maxWidth)) / 2);
      for (var r = 0; r < runs.length; r++) {
        var item = runs[r];
        if (item.type === 'image') {
          var sized = imageSize(item.node, maxWidth);
          if (sized) {
            var scale = Math.min(1, (maxHeight - 2) / sized.height);
            var w = sized.width * scale, h = sized.height * scale;
            drawImage(sized.entry, x, top, w, h);
            x += w + 4;
          }
          continue;
        }
        if (item.type !== 'run') continue;
        var width = pdfTextWidth(item.text.replace(/__SUI_(PAGE|PAGES)__/g, '00'), item.font, item.fontSize);
        ops.push('BT ' + pdfFill(item.color || '#666666') + ' /' + item.font.name + ' ' + item.fontSize + ' Tf ' + pdfNumber(x) + ' ' + pdfNumber(ty(top + item.fontSize * 0.8)) + ' Td (' + pdfEscape(pdfEncode(item.text)) + ') Tj ET');
        x += width;
      }
    }
    function drawChromeValue(value, top, areaH, borderTop) {
      var blocks = chromeDoc(value).children || [];
      var lineTop = top;
      for (var b = 0; b < blocks.length; b++) {
        var block = blocks[b];
        if (!block) continue;
        if (block.type === 'paragraph' || block.type === 'heading') {
          var runs = chromeRuns(block.children);
          if (lineTop + 11 > top + areaH) break;
          drawChromeRuns(runs, lineTop, CONTENT_W, Math.min(18, areaH));
          lineTop += 11;
        } else if (block.type === 'image') {
          var sized = imageSize(block, CONTENT_W);
          if (sized && lineTop + sized.height <= top + areaH) {
            drawImage(sized.entry, MARGIN + Math.max(0, (CONTENT_W - sized.width) / 2), lineTop, sized.width, sized.height);
            lineTop += sized.height + 2;
          }
        }
      }
      ops.push('q ' + pdfStroke('#cccccc') + ' 0.5 w ' + pdfNumber(MARGIN) + ' ' + pdfNumber(ty(borderTop)) + ' m ' + pdfNumber(PAGE_W - MARGIN_R) + ' ' + pdfNumber(ty(borderTop)) + ' l S Q');
    }
    function drawChrome() {
      if (hasHeader) drawChromeValue(headerValue, 6, Math.max(12, headerSpace - 10), headerSpace - 3);
      if (hasFooter) drawChromeValue(footerValue, PAGE_H - footerSpace + 4, Math.max(12, footerSpace - 10), PAGE_H - footerSpace + 3);
    }
    function newPage() { if (ops.length) pages.push(ops.join('\n')); ops = []; y = START_Y; pageSeq++; drawChrome(); }
    function ensure(h) { if (y + h > BOTTOM_LIMIT) newPage(); }
    function ty(top) { return PAGE_H - top; }
    function imageFor(node) {
      var src = String(node.src || '');
      if (imageCache[src] !== undefined) return imageCache[src];
      var match = /^data:image\/(png|jpe?g);base64,([\s\S]+)$/i.exec(src);
      var entry = null;
      if (match) {
        try {
          var binary = atob(match[2]), bytes = new Uint8Array(binary.length);
          for (var i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
          if (/jpe?g/i.test(match[1])) {
            var info = jpegInfo(bytes);
            if (info && info.components >= 1 && info.components <= 4) entry = { kind: 'jpeg', width: info.width, height: info.height, bytes: bytes, stream: binaryString(bytes) };
          } else {
            var png = pngImage(bytes);
            if (png) entry = { kind: 'raw', width: png.width, height: png.height, bytes: png.data, stream: binaryString(png.data) };
          }
        } catch (e) { entry = null; }
      }
      if (entry) { entry.name = 'Im' + images.length; images.push(entry); }
      imageCache[src] = entry;
      return entry;
    }
    function drawImage(entry, x, top, width, height) {
      ops.push('q ' + pdfNumber(width) + ' 0 0 ' + pdfNumber(height) + ' ' + pdfNumber(x) + ' ' + pdfNumber(ty(top + height)) + ' cm /' + entry.name + ' Do Q');
    }
    function imageSize(node, maxWidth) {
      var entry = imageFor(node);
      var width = parseFloat(node.width) || 0, height = parseFloat(node.height) || 0;
      if (!entry) return null;
      if (!width && !height) { width = entry.width; height = entry.height; }
      else if (!width) width = height * entry.width / entry.height;
      else if (!height) height = width * entry.height / entry.width;
      var scale = Math.min(1, maxWidth / width);
      return { entry: entry, width: width * scale, height: height * scale };
    }
    function runsOf(nodes, baseMarks, hyphenate) {
      var runs = [];
      for (var i = 0; i < (nodes || []).length; i++) {
        var node = nodes[i];
        if (!node) continue;
        if (node.type === 'text') {
          var text = hyphenate ? hyphenateText(node.text) : String(node.text || '');
          var marks = (baseMarks || []).concat(node.marks || []);
          var font = pdfFont(marks), size = pdfSize(marks), color = null, background = null, underline = false, strike = false, sup = false, sub = false;
          for (var m = 0; m < marks.length; m++) {
            var mark = marks[m];
            if (mark.type === 'color') color = mark.value;
            else if (mark.type === 'background') background = mark.value;
            else if (mark.type === 'underline') underline = true;
            else if (mark.type === 'strike') strike = true;
            else if (mark.type === 'sup') sup = true;
            else if (mark.type === 'sub') sub = true;
          }
          var current = null;
          for (var c = 0; c < text.length; c++) {
            var ch = text.charAt(c), base = { font: font, size: size, color: color, background: background, underline: underline, strike: strike, sup: sup, sub: sub };
            if (ch === '\n') { runs.push({ type: 'break' }); current = null; continue; }
            if (ch === '\u2610' || ch === '\u2611' || ch === '\u2612') {
              runs.push({ type: 'checkbox', checked: ch === '\u2611', crossed: ch === '\u2612', size: size, width: size * 0.95 });
              current = null;
              continue;
            }
            if (!current) { current = Object.assign({ type: 'run', text: '' }, base); runs.push(current); }
            current.text += ch;
          }
        } else if (node.type === 'break') {
          runs.push({ type: 'break' });
        } else if (node.type === 'image') {
          runs.push({ type: 'image', node: node });
        }
      }
      for (var r = 0; r < runs.length; r++) {
        var run = runs[r];
        if (run.type === 'run') {
          if (run.font.emb) {
            var mapped = true;
            for (var mc = 0; mc < run.text.length; mc++) if (!run.font.emb.gidFor(run.text.charCodeAt(mc))) { mapped = false; break; }
            if (!mapped) run.font = { id: 0, name: run.font.stdName, family: run.font.family, bold: run.font.bold, italic: run.font.italic, emb: null };
          }
          if (run.font.emb) registerEmbText(run.font, run.text);
          run.width = pdfTextWidth(run.text, run.font, run.size);
          run.fontSize = run.size * (run.sup || run.sub ? 0.72 : 1);
        }
      }
      return runs;
    }
    function codeRunsFromNode(node) {
      var segs = codeExportSegments(node), runs = [];
      for (var i = 0; i < segs.length; i++) {
        if (!segs[i].text) continue;
        var parts = String(segs[i].text).split('\n');
        var marks = codeSegmentMarks(segs[i]);
        for (var p = 0; p < parts.length; p++) {
          if (p) runs.push({ type: 'break' });
          if (parts[p]) runs = runs.concat(runsOf([{ type: 'text', text: parts[p], marks: marks }], null, false));
        }
      }
      return runs;
    }
    function wrapRuns(runs, maxWidth, preserveSpace) {
      var lines = [], line = [], width = 0;
      var maxSize = 0;
      function flush() { lines.push({ runs: line, width: width, size: maxSize || 11 }); line = []; width = 0; maxSize = 0; }
      function add(part, run, isSpace) {
        if (isSpace && !line.length && !preserveSpace) return;
        line.push({ run: run, text: part, width: pdfTextWidth(part, run.font, run.fontSize) });
        width += line[line.length - 1].width;
        if (run.fontSize > maxSize) maxSize = run.fontSize;
      }
      function addCharSplit(word, run) {
        while (word) {
          var take = '';
          while (word.length) {
            var candidate = take + word.charAt(0);
            if (pdfTextWidth(candidate, run.font, run.fontSize) > maxWidth - width && take) break;
            take = candidate; word = word.slice(1);
          }
          add(take, run, false);
          if (word) flush();
        }
      }
      function addHyphenated(part, run) {
        var segs = part.split('\u00AD').filter(function (s) { return s.length > 0; });
        if (segs.length < 2) { add(part.replace(/\u00AD/g, ''), run, false); return; }
        var acc = '';
        for (var s = 0; s < segs.length; s++) {
          var seg = segs[s];
          var test = acc + seg;
          var testWidth = pdfTextWidth(test, run.font, run.fontSize);
          if (width + testWidth <= maxWidth || !acc) {
            if (!acc && line.length && width + testWidth > maxWidth) flush();
            acc = test;
            continue;
          }
          var hyphenWidth = pdfTextWidth('-', run.font, run.fontSize);
          var accWidth = pdfTextWidth(acc, run.font, run.fontSize);
          add(width + accWidth + hyphenWidth <= maxWidth + 0.5 ? acc + '-' : acc, run, false);
          flush();
          acc = seg;
        }
        if (acc) {
          if (pdfTextWidth(acc, run.font, run.fontSize) > maxWidth && !line.length) addCharSplit(acc, run);
          else add(acc, run, false);
        }
      }
      for (var i = 0; i < runs.length; i++) {
        var run = runs[i];
        if (run.type === 'break') { flush(); continue; }
        if (run.type === 'checkbox') {
          if (width + run.width > maxWidth && line.length) flush();
          line.push({ checkbox: run, width: run.width });
          width += run.width;
          if (run.size > maxSize) maxSize = run.size;
          continue;
        }
        if (run.type === 'image') {
          if (line.length) flush();
          var sized = imageSize(run.node, maxWidth);
          if (sized) line.push({ image: sized, width: sized.width, height: sized.height });
          if (sized) width += sized.width;
          flush();
          continue;
        }
        var parts = run.text.split(/(\s+)/);
        for (var p = 0; p < parts.length; p++) {
          var part = parts[p];
          if (!part) continue;
          var isSpace = /^\s+$/.test(part);
          var w = pdfTextWidth(part, run.font, run.fontSize);
          if (width + w > maxWidth && line.length && !isSpace) flush();
          if (isSpace && !preserveSpace && width + w > maxWidth) { flush(); continue; }
          if (!isSpace && part.indexOf('\u00AD') >= 0) { addHyphenated(part, run); continue; }
          if (w > maxWidth && !isSpace) {
            addCharSplit(part, run);
            continue;
          }
          add(part, run, isSpace);
        }
      }
      flush();
      if (!lines.length) lines.push({ runs: [], width: 0, size: 11 });
      return lines;
    }
    function lineHeightFor(size, node) {
      var info = lineHeightInfo(node && node.lineHeight);
      if (info && info.pt) return Math.max(size * 1.18, info.pt);
      var factor = info && info.multiple ? info.multiple : (parseFloat(pageConfig(inst).lineHeight) || 1.35);
      return Math.max(size * 1.18, size * factor);
    }
    function drawLines(lines, x0, maxWidth, align, node, allowBreak) {
      for (var i = 0; i < lines.length; i++) {
        var line = lines[i];
        var size = line.size || 11;
        var height = lineHeightFor(size, node);
        if (allowBreak !== false) ensure(height);
        var alignValue = String((node && node.align) || '').toLowerCase();
        var startX = x0;
        if (alignValue === 'center') startX = x0 + Math.max(0, (maxWidth - line.width) / 2);
        else if (alignValue === 'right') startX = x0 + Math.max(0, maxWidth - line.width);
        var baseline = y + size * 0.82;
        var segments = [], cursor = startX;
        for (var r = 0; r < line.runs.length; r++) {
          segments.push({ piece: line.runs[r], x: cursor });
          cursor += line.runs[r].width;
        }
        for (var b = 0; b < segments.length; b++) {
          var seg = segments[b];
          if (seg.piece.run && seg.piece.run.background) ops.push('q ' + pdfFill(seg.piece.run.background) + ' ' + pdfNumber(seg.x) + ' ' + pdfNumber(ty(y + height - 1)) + ' ' + pdfNumber(seg.piece.width) + ' ' + pdfNumber(height - 2) + ' re f Q');
        }
        var inText = false, lastFont = '', lastSize = 0, lastFill = '';
        function openText(atX, atBaseline) {
          if (inText) return;
          ops.push('BT 1 0 0 1 ' + pdfNumber(atX) + ' ' + pdfNumber(ty(atBaseline)) + ' Tm');
          inText = true;
        }
        function closeText() { if (inText) { ops.push('ET'); inText = false; lastFont = ''; lastSize = 0; lastFill = ''; } }
        for (var s = 0; s < segments.length; s++) {
          var seg2 = segments[s], piece2 = seg2.piece;
          if (piece2.checkbox) {
            closeText();
            var cb = piece2.checkbox, box = cb.size * 0.72, boxY = y + (height - box) / 2;
            ops.push('q ' + pdfStroke('#333333') + ' 0.7 w ' + pdfNumber(seg2.x) + ' ' + pdfNumber(ty(boxY + box)) + ' ' + pdfNumber(box) + ' ' + pdfNumber(box) + ' re S Q');
            if (cb.checked || cb.crossed) ops.push('q ' + pdfStroke('#333333') + ' 0.9 w ' + pdfNumber(seg2.x + box * 0.18) + ' ' + pdfNumber(ty(boxY + box * 0.55)) + ' m ' + pdfNumber(seg2.x + box * 0.42) + ' ' + pdfNumber(ty(boxY + box * 0.82)) + ' l ' + pdfNumber(seg2.x + box * 0.85) + ' ' + pdfNumber(ty(boxY + box * 0.16)) + ' l S Q');
            continue;
          }
          if (piece2.image) {
            closeText();
            var img = piece2.image;
            ops.push('q ' + pdfNumber(img.width) + ' 0 0 ' + pdfNumber(img.height) + ' ' + pdfNumber(seg2.x) + ' ' + pdfNumber(ty(y + img.height)) + ' cm /' + img.entry.name + ' Do Q');
            continue;
          }
          var run2 = piece2.run, fontSize2 = run2.fontSize;
          var drawBaseline = baseline + (run2.sup ? -fontSize2 * 0.35 : run2.sub ? fontSize2 * 0.22 : 0);
          openText(seg2.x, drawBaseline);
          var fill2 = run2.color || '#000000';
          if (fill2 !== lastFill) { ops.push(pdfFill(fill2)); lastFill = fill2; }
          if (run2.font.name !== lastFont || fontSize2 !== lastSize) { ops.push('/' + run2.font.name + ' ' + pdfNumber(fontSize2) + ' Tf'); lastFont = run2.font.name; lastSize = fontSize2; }
          if (run2.font.emb) { registerEmbText(run2.font, piece2.text); ops.push('<' + pdfGids(piece2.text, run2.font.emb) + '> Tj'); }
          else ops.push('(' + pdfEscape(pdfEncode(piece2.text)) + ') Tj');
        }
        closeText();
        for (var u = 0; u < segments.length; u++) {
          var segU = segments[u], runU = segU.piece.run;
          if (!runU) continue;
          var baselineU = baseline + (runU.sup ? -runU.fontSize * 0.35 : runU.sub ? runU.fontSize * 0.22 : 0);
          if (runU.underline) ops.push('q ' + pdfStroke(runU.color || '#000000') + ' 0.5 w ' + pdfNumber(segU.x) + ' ' + pdfNumber(ty(baselineU + 1.5)) + ' m ' + pdfNumber(segU.x + segU.piece.width) + ' ' + pdfNumber(ty(baselineU + 1.5)) + ' l S Q');
          if (runU.strike) ops.push('q ' + pdfStroke(runU.color || '#000000') + ' 0.5 w ' + pdfNumber(segU.x) + ' ' + pdfNumber(ty(baselineU - runU.fontSize * 0.3)) + ' m ' + pdfNumber(segU.x + segU.piece.width) + ' ' + pdfNumber(ty(baselineU - runU.fontSize * 0.3)) + ' l S Q');
        }
        y += height;
      }
    }
    function spacing(node, key) {
      var v = parseFloat(node && node[key]);
      return isFinite(v) && v > 0 ? v : 0;
    }
    function paragraph(node) {
      var before = spacing(node, 'spaceBefore'), after = spacing(node, 'spaceAfter');
      if (before) y += before;
      var indent = Math.max(0, (parseFloat(node.indent) || 0) * 0.75);
      var runs = runsOf(node.children, null, node.hyphens === 'auto');
      var lines = wrapRuns(runs, Math.max(40, CONTENT_W - indent));
      drawLines(lines, MARGIN + indent, CONTENT_W - indent, node.align, node);
      if (after) y += after;
      else y += 2;
    }
    function heading(node) {
      var level = Math.max(1, Math.min(6, Number(node.level) || 1));
      var size = Math.max(12, 20 - (level - 1) * 2);
      var runs = runsOf(node.children, null, node.hyphens === 'auto');
      for (var i = 0; i < runs.length; i++) if (runs[i].type === 'run') {
        if (!runs[i].font.bold) {
          var family = runs[i].font.family;
          runs[i].font = pdfFontParts(family, true, runs[i].font.italic);
        }
        runs[i].size = size; runs[i].fontSize = size;
        runs[i].width = pdfTextWidth(runs[i].text, runs[i].font, size);
      }
      var lines = wrapRuns(runs, CONTENT_W);
      y += 4;
      drawLines(lines, MARGIN, CONTENT_W, node.align, node);
      y += 4;
    }
    function cellTextOf(cell) {
      if (Array.isArray(cell.blocks) && cell.blocks.length) {
        var parts = [];
        for (var b = 0; b < cell.blocks.length; b++) parts.push(blockTextOf(cell.blocks[b]));
        return parts.join(' / ');
      }
      var runs = runsOf(cell.children), text = '';
      for (var r = 0; r < runs.length; r++) if (runs[r].type === 'run') text += runs[r].text;
      return text;
    }
    function blockTextOf(block) {
      if (!block) return '';
      if (block.type === 'paragraph' || block.type === 'heading') return cellTextOf({ children: block.children });
      if (block.type === 'table') {
        var rows = [];
        for (var r = 0; r < (block.rows || []).length; r++) {
          var cells = [];
          for (var c = 0; c < (block.rows[r].cells || []).length; c++) cells.push(cellTextOf(block.rows[r].cells[c]));
          rows.push(cells.join(' | '));
        }
        return rows.join(' / ');
      }
      if (block.type === 'code_block') return String(block.text || '');
      if (block.type === 'image') return block.alt || '';
      if (block.type === 'bullet_list' || block.type === 'ordered_list') {
        var items = [];
        for (var li = 0; li < (block.children || []).length; li++) items.push(blockTextOf({ type: 'paragraph', children: (block.children[li].children && block.children[li].children[0] ? block.children[li].children[0].children : []) }));
        return items.join(' / ');
      }
      return '';
    }
    function runsOfBlocks(blocks) {
      var out = [];
      for (var i = 0; i < blocks.length; i++) {
        var block = blocks[i];
        if (block.type === 'paragraph' || block.type === 'heading') {
          out = out.concat(runsOf(block.children));
          out.push({ type: 'break' });
        } else if (block.type === 'table') {
          out.push({ type: 'run', text: blockTextOf(block), font: pdfFont([]), size: 11, fontSize: 11, width: 0, color: null, background: null, underline: false, strike: false, sup: false, sub: false });
          out.push({ type: 'break' });
        } else if (block.type === 'code_block') {
          out.push({ type: 'run', text: String(block.text || ''), font: pdfFont([]), size: 10, fontSize: 10, width: 0, color: null, background: null, underline: false, strike: false, sup: false, sub: false });
          out.push({ type: 'break' });
        } else if (block.type === 'bullet_list' || block.type === 'ordered_list') {
          out.push({ type: 'run', text: blockTextOf(block), font: pdfFont([]), size: 11, fontSize: 11, width: 0, color: null, background: null, underline: false, strike: false, sup: false, sub: false });
          out.push({ type: 'break' });
        } else if (block.type === 'image') {
          out.push({ type: 'image', node: block });
        }
      }
      for (var k = 0; k < out.length; k++) if (out[k].type === 'run') out[k].width = pdfTextWidth(out[k].text, out[k].font, out[k].fontSize);
      return out;
    }
    function table(node) {
      var rows = node.rows || [];
      if (!rows.length) return;
      var dom = domTables[tableSeq++];
      var tableW = 0, tableLeft = 0;
      if (dom) { var tableRect = dom.getBoundingClientRect(); tableW = tableRect.width; tableLeft = tableRect.left; }
      var cellPad = node.padding ? Math.max(0, Math.min(30, parseFloat(node.padding) * 0.75)) : 2.5;
      var borderW = node.borderWidth ? Math.max(0, Math.min(4, parseFloat(node.borderWidth) * 0.75)) : 0.5;
      var borderColor = node.border || '#9aa7b8';
      var rowData = [];
      for (var rr = 0; rr < rows.length; rr++) {
        var row = rows[rr], domRow = dom ? dom.rows[rr] : null;
        var cells = [], fallbackX = 0;
        var colStart = 0;
        for (var cc = 0; cc < (row.cells || []).length; cc++) {
          var cell = row.cells[cc];
          var pct = /^([\d.]+)%$/.exec(String(cell.width || ''));
          var width = pct ? (parseFloat(pct[1]) / 100) * CONTENT_W : CONTENT_W / Math.max(1, (row.cells || []).length);
          var x = MARGIN + fallbackX;
          if (domRow && domRow.cells[cc] && tableW > 0) {
            var cellRect = domRow.cells[cc].getBoundingClientRect();
            if (cellRect.width > 0) {
              x = MARGIN + ((cellRect.left - tableLeft) / tableW) * CONTENT_W;
              width = (cellRect.width / tableW) * CONTENT_W;
            }
          }
          fallbackX += width;
          var runs = (Array.isArray(cell.blocks) && cell.blocks.length) ? runsOfBlocks(cell.blocks) : runsOf(cell.children);
          var lines = wrapRuns(runs, Math.max(12, width - cellPad * 2));
          var maxSize = 11;
          for (var lI = 0; lI < lines.length; lI++) if (lines[lI].size > maxSize) maxSize = lines[lI].size;
          var h = lines.length * lineHeightFor(maxSize, null) + cellPad * 2;
          var spanCount = Math.max(1, Number(cell.colspan) || 1);
          cells.push({ cell: cell, x: x, width: width, lines: lines, height: h, rowspan: Math.max(1, Number(cell.rowspan) || 1), start: colStart, span: spanCount });
          colStart += spanCount;
        }
        var rowHeight = 0;
        for (var hI = 0; hI < cells.length; hI++) if (cells[hI].rowspan === 1) rowHeight = Math.max(rowHeight, cells[hI].height);
        if (!rowHeight) rowHeight = 14;
        if (row.height) rowHeight = Math.max(rowHeight, (parseFloat(row.height) || 0) * 0.75);
        rowData.push({ cells: cells, height: rowHeight });
      }
      var borderFlags = node.borders || ['top', 'right', 'bottom', 'left', 'insideH', 'insideV'];
      var totalCols = 0;
      for (var tc = 0; tc < rowData.length; tc++) for (var tcc = 0; tcc < rowData[tc].cells.length; tcc++) totalCols = Math.max(totalCols, rowData[tc].cells[tcc].start + rowData[tc].cells[tcc].span);
      for (var fr = 0; fr < rowData.length; fr++) {
        for (var fc = 0; fc < rowData[fr].cells.length; fc++) {
          var spanCell = rowData[fr].cells[fc];
          if (spanCell.rowspan <= 1) continue;
          var total = 0, last = fr;
          for (var sr = 0; sr < spanCell.rowspan && fr + sr < rowData.length; sr++) { total += rowData[fr + sr].height; last = fr + sr; }
          if (total < spanCell.height) rowData[last].height += spanCell.height - total;
        }
      }
      function drawRowCells(data, rowTop, rowIndex) {
        for (var cI = 0; cI < data.cells.length; cI++) {
          var info = data.cells[cI];
          var cellHeight = data.height;
          if (info.rowspan > 1) {
            cellHeight = 0;
            for (var rs = 0; rs < info.rowspan && rowIndex + rs < rowData.length; rs++) cellHeight += rowData[rowIndex + rs].height;
          }
          var cellFill = info.cell.background || (info.cell.header ? '#e8ecf3' : '');
          if (cellFill) ops.push('q ' + pdfFill(cellFill) + ' ' + pdfNumber(info.x) + ' ' + pdfNumber(ty(rowTop + cellHeight)) + ' ' + pdfNumber(info.width) + ' ' + pdfNumber(cellHeight) + ' re f Q');
          var onSide = function (name) { return borderFlags.indexOf(name) >= 0; };
          var topOn = rowIndex === 0 ? onSide('top') : onSide('insideH');
          var bottomOn = (rowIndex + info.rowspan >= rowData.length) ? onSide('bottom') : onSide('insideH');
          var leftOn = info.start === 0 ? onSide('left') : onSide('insideV');
          var rightOn = (info.start + info.span >= totalCols) ? onSide('right') : onSide('insideV');
          function pdfLine(x1, y1, x2, y2) {
            ops.push('q ' + pdfStroke(borderColor) + ' ' + pdfNumber(borderW) + ' w ' + pdfNumber(x1) + ' ' + pdfNumber(ty(y1)) + ' m ' + pdfNumber(x2) + ' ' + pdfNumber(ty(y2)) + ' l S Q');
          }
          if (topOn) pdfLine(info.x, rowTop, info.x + info.width, rowTop);
          if (bottomOn) pdfLine(info.x, rowTop + cellHeight, info.x + info.width, rowTop + cellHeight);
          if (leftOn) pdfLine(info.x, rowTop, info.x, rowTop + cellHeight);
          if (rightOn) pdfLine(info.x + info.width, rowTop, info.x + info.width, rowTop + cellHeight);
          var savedY = y;
          y = rowTop + cellPad;
          drawLines(info.lines, info.x + cellPad, Math.max(10, info.width - cellPad * 2), String(info.cell.align || '').toLowerCase(), null, false);
          y = savedY;
        }
      }
      var headerRows = [];
      for (var hr = 0; hr < rowData.length; hr++) {
        var hrow = rowData[hr];
        var allHeader = hrow.cells.length > 0;
        for (var hc = 0; hc < hrow.cells.length; hc++) if (!hrow.cells[hc].cell.header) allHeader = false;
        if (allHeader || (hr === 0 && node.header)) headerRows.push(hr);
        else break;
      }
      for (var rowIndex = 0; rowIndex < rowData.length; rowIndex++) {
        var data = rowData[rowIndex];
        var beforePage = pageSeq;
        ensure(data.height);
        if (pageSeq !== beforePage && headerRows.length) {
          for (var hi = 0; hi < headerRows.length; hi++) {
            var hd = rowData[headerRows[hi]];
            if (headerRows[hi] >= rowIndex) continue;
            drawRowCells(hd, y, headerRows[hi]);
            y += hd.height;
          }
        }
        var rowTop = y;
        drawRowCells(data, rowTop, rowIndex);
        y = rowTop + data.height;
      }
      y += 3;
    }
    function pdfList(node, level) {
      var number = 0;
      for (var i = 0; i < (node.children || []).length; i++) {
        var item = node.children[i];
        var blocks = item.children || [];
        for (var b = 0; b < blocks.length; b++) {
          var inner = blocks[b];
          if (inner.type === 'paragraph') {
            number++;
            var prefix = node.type === 'ordered_list' ? (number + '. ') : '\u2022 ';
            var head = { type: 'run', text: prefix, font: pdfFont([]), size: 11, fontSize: 11, width: pdfTextWidth(prefix, pdfFont([]), 11) };
            var indent = level * 18;
            var lines = wrapRuns([head].concat(runsOf(inner.children, null, inner.hyphens === 'auto')), Math.max(40, CONTENT_W - indent - 18));
            drawLines(lines, MARGIN + indent + 18, Math.max(40, CONTENT_W - indent - 18), inner.align, null);
          } else if (inner.type === 'bullet_list' || inner.type === 'ordered_list') {
            pdfList(inner, level + 1);
          } else block(inner);
        }
      }
    }
    function block(node) {
      if (!node) return;
      if (node.type === 'heading') heading(node);
      else if (node.type === 'paragraph') paragraph(node);
      else if (node.type === 'blockquote') paragraph(node);
      else if (node.type === 'code_block') {
        var codeNodes = (Array.isArray(node.children) && node.children.length) ? node.children : [{ type: 'text', text: String(node.text || ''), marks: [] }];
        var codeRuns = (node.codeLanguage && node.codeLanguage !== 'plain') || node.lineNumbers ? codeRunsFromNode(node) : runsOf(codeNodes, [{ type: 'font', value: 'Courier New' }, { type: 'size', value: '9' }]);
        var codeLines = wrapRuns(codeRuns, CONTENT_W - 12, true);
        drawLines(codeLines, MARGIN + 6, CONTENT_W - 12, node.align, node);
      } else if (node.type === 'hr') {
        ensure(10);
        ops.push('q ' + pdfStroke('#9aa7b8') + ' 0.7 w ' + pdfNumber(MARGIN) + ' ' + pdfNumber(ty(y)) + ' m ' + pdfNumber(PAGE_W - MARGIN_R) + ' ' + pdfNumber(ty(y)) + ' l S Q');
        y += 10;
      } else if (node.type === 'page_break') {
        newPage();
      } else if (node.type === 'image') {
        var sized = imageSize(node, CONTENT_W);
        if (!sized) { ensure(12); y += 12; return; }
        ensure(sized.height + 6);
        var ix = MARGIN;
        if (node.align === 'center') ix = MARGIN + Math.max(0, (CONTENT_W - sized.width) / 2);
        else if (node.align === 'right') ix = MARGIN + Math.max(0, CONTENT_W - sized.width);
        drawImage(sized.entry, ix, y, sized.width, sized.height);
        y += sized.height + 6;
      } else if (node.type === 'bullet_list' || node.type === 'ordered_list') {
        pdfList(node, 0);
      } else if (node.type === 'table') table(node);
    }
    drawChrome();
    for (var i = 0; i < (doc.children || []).length; i++) block(doc.children[i]);
    newPage();
    if (!pages.length) pages.push('');
    var objects = [];
    objects[1] = '<< /Type /Catalog /Pages 2 0 R >>';
    for (var f = 0; f < PDF_FONTS.length; f++) objects[3 + f] = '<< /Type /Font /Subtype /Type1 /BaseFont /' + PDF_FONTS[f] + ' /Encoding /WinAnsiEncoding >>';
    var nextId = 3 + PDF_FONTS.length;
    var imageIds = [];
    for (var im = 0; im < images.length; im++) imageIds.push(nextId++);
    var embKeys = [];
    for (var ek in pdfEmbUse) if (Object.prototype.hasOwnProperty.call(pdfEmbUse, ek) && PDF_EMB_NAMES[ek] && PDF_EMB_CACHE[ek]) embKeys.push(ek);
    var embIds = {};
    for (var ei = 0; ei < embKeys.length; ei++) embIds[embKeys[ei]] = { type0: nextId++, cid: nextId++, desc: nextId++, file: nextId++, tu: nextId++ };
    var pageIds = [], contentIds = [];
    for (var p = 0; p < pages.length; p++) { pageIds.push(nextId++); contentIds.push(nextId++); }
    for (var ii = 0; ii < images.length; ii++) {
      var entry = images[ii];
      objects[imageIds[ii]] = '<< /Type /XObject /Subtype /Image /Width ' + entry.width + ' /Height ' + entry.height + ' /ColorSpace /DeviceRGB /BitsPerComponent 8 ' + (entry.kind === 'jpeg' ? '/Filter /DCTDecode ' : '') + '/Length ' + entry.bytes.length + ' >>\nstream\n' + entry.stream + '\nendstream';
    }
    for (var eb = 0; eb < embKeys.length; eb++) {
      var ebKey = embKeys[eb], ebFont = PDF_EMB_CACHE[ebKey], ebId = embIds[ebKey];
      var ebScale = 1000 / ebFont.unitsPerEm;
      var ebBase = 'SUI' + ebKey.split('-').map(function (part) { return part.charAt(0).toUpperCase() + part.slice(1); }).join('-');
      var ebUsed = pdfEmbUse[ebKey];
      var ebGids = Object.keys(ebUsed);
      var ebW = ebGids.map(function (gid) { return gid + ' [' + Math.round((ebFont.widths[Number(gid)] || 0) * ebScale) + ']'; }).join(' ');
      var ebBBox = [ebFont.xMin, ebFont.yMin, ebFont.xMax, ebFont.yMax].map(function (v) { return Math.round(v * ebScale); }).join(' ');
      var ebAsc = Math.round(ebFont.ascent * ebScale), ebDesc = Math.round(ebFont.descent * ebScale), ebCap = Math.round(ebFont.ascent * ebScale * 0.7);
      var ebTu = pdfToUnicode(ebBase, ebUsed);
      objects[ebId.file] = '<< /Length ' + ebFont.bytes.length + ' /Length1 ' + ebFont.bytes.length + ' >>' + String.fromCharCode(10) + 'stream' + String.fromCharCode(10) + binaryString(ebFont.bytes) + String.fromCharCode(10) + 'endstream';
      objects[ebId.desc] = '<< /Type /FontDescriptor /FontName /' + ebBase + ' /Flags 32 /FontBBox [' + ebBBox + '] /ItalicAngle 0 /Ascent ' + ebAsc + ' /Descent ' + ebDesc + ' /CapHeight ' + ebCap + ' /StemV 80 /FontFile2 ' + ebId.file + ' 0 R >>';
      objects[ebId.cid] = '<< /Type /Font /Subtype /CIDFontType2 /BaseFont /' + ebBase + ' /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor ' + ebId.desc + ' 0 R /DW 1000 /W [' + ebW + '] /CIDToGIDMap /Identity >>';
      objects[ebId.type0] = '<< /Type /Font /Subtype /Type0 /BaseFont /' + ebBase + ' /Encoding /Identity-H /DescendantFonts [' + ebId.cid + ' 0 R] /ToUnicode ' + ebId.tu + ' 0 R >>';
      objects[ebId.tu] = '<< /Length ' + pdfBytes(ebTu).length + ' >>' + String.fromCharCode(10) + 'stream' + String.fromCharCode(10) + ebTu + String.fromCharCode(10) + 'endstream';
    }
    var fontRes = '/Font << ' + PDF_FONTS.map(function (_, index) { return '/F' + (index + 1) + ' ' + (3 + index) + ' 0 R'; }).join(' ') + embKeys.map(function (key) { return ' /' + PDF_EMB_NAMES[key] + ' ' + embIds[key].type0 + ' 0 R'; }).join('') + ' >>';
    var xobjRes = images.length ? ' /XObject << ' + images.map(function (_, index) { return '/Im' + index + ' ' + imageIds[index] + ' 0 R'; }).join(' ') + ' >>' : '';
    objects[2] = '<< /Type /Pages /Kids [' + pageIds.map(function (id) { return id + ' 0 R'; }).join(' ') + '] /Count ' + pages.length + ' >>';
    for (var pg = 0; pg < pages.length; pg++) {
      objects[pageIds[pg]] = '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ' + PAGE_W + ' ' + PAGE_H + '] /Resources << ' + fontRes + xobjRes + ' >> /Contents ' + contentIds[pg] + ' 0 R >>';
      var streamPg = pages[pg].split('__SUI_PAGE__').join(String(pg + 1)).split('__SUI_PAGES__').join(String(pages.length));
      objects[contentIds[pg]] = '<< /Length ' + pdfBytes(streamPg).length + ' >>\nstream\n' + streamPg + '\nendstream';
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
    exportFormat: exportFormat, importFormat: importFormatTo, exportBytes: exportBytes, importBytes: importBytes, openFile: openFile, saveAs: saveAs, setTheme: setTheme, registerPlugin: registerPlugin, plugins: function () { return [pluginNames(), null]; }, setTemplate: setTemplate,
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
  --sui-rich-paper: #ffffff;
  --sui-rich-paper-fg: #1c2330;
  --sui-rich-paper-border: #c9d3e0;
  --sui-rich-paper-muted: #5a6577;
  --sui-rich-break: #dc2626;
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
  color-scheme: light;
}
.sui-rich[data-sui-rich-content="dark"] {
  --sui-rich-paper: #14171c;
  --sui-rich-paper-fg: #e6e6ef;
  --sui-rich-paper-border: #2c3444;
  --sui-rich-paper-muted: #9aa7b8;
  --sui-rich-break: #f87171;
}
.sui-rich-title { padding: 10px 12px 0; font-weight: 600; color: var(--sui-rich-fg); }
.sui-rich-menu { display: flex; flex-wrap: wrap; gap: 2px; padding: 4px 8px; border-bottom: 1px solid var(--sui-rich-border); background: var(--sui-rich-toolbar-bg); position: relative; z-index: 30; }
.sui-rich-menu-wrap { position: relative; }
.sui-rich-menu-btn { border: 0; background: transparent; color: var(--sui-rich-fg); padding: 5px 9px; border-radius: 6px; cursor: pointer; font: inherit; font-size: 12px; }
.sui-rich-menu-btn:hover, .sui-rich-menu-btn:focus-visible { background: var(--sui-rich-hover); outline: none; }
.sui-rich-menu-panel { position: absolute; left: 0; top: calc(100% + 4px); min-width: 190px; padding: 5px; background: var(--sui-rich-bg); border: 1px solid var(--sui-rich-border); border-radius: 8px; box-shadow: 0 12px 30px rgba(0,0,0,.28); z-index: 20; }
.sui-rich-menu-item { display: block; width: 100%; text-align: left; border: 0; background: transparent; color: var(--sui-rich-fg); padding: 7px 9px; border-radius: 6px; cursor: pointer; font: inherit; font-size: 12px; }
.sui-rich-menu-item:hover, .sui-rich-menu-item:focus-visible { background: var(--sui-rich-hover); outline: none; }
.sui-rich-menu-subwrap { position: relative; }
.sui-rich-menu-parent { display: flex; justify-content: space-between; align-items: center; }
.sui-rich-submenu { left: calc(100% + 4px); top: -5px; min-width: 150px; }
.sui-rich-toolbar { display: flex; flex-wrap: wrap; gap: 5px; padding: 6px 8px; border-bottom: 1px solid var(--sui-rich-border); background: var(--sui-rich-toolbar-bg); align-items: center; }
.sui-rich-tool-group { display: flex; align-items: center; gap: 2px; padding: 2px 4px; border: 1px solid var(--sui-rich-border); border-radius: 9px; background: color-mix(in srgb, var(--sui-rich-fg) 4%, transparent); }
.sui-rich-tool-group[hidden] { display: none; }
.sui-rich-tool-handle { cursor: grab; color: var(--sui-rich-muted); font-size: 11px; padding: 0 2px; user-select: none; line-height: 1; }
.sui-rich-tool-handle:active { cursor: grabbing; }
.sui-rich-pages { flex: 1 1 auto; min-width: 0; height: var(--sui-rich-content-height, 320px); padding: 18px 18px 26px; overflow: auto; background: #e6e9ef; }
.sui-rich-pages[hidden] { display: none; }
.sui-rich-paginated .sui-rich-body > .sui-rich-page,
.sui-rich-paginated .sui-rich-body > .sui-rich-editor { display: none; }
.sui-rich-paginated .sui-rich-body > .sui-rich-measure { display: block; }
.sui-rich-paginated .sui-rich-pages { display: block; }
.sui-rich-page-sheet { position: relative; margin: 0 auto 22px; background: #ffffff; color: #1c2330; box-shadow: 0 10px 28px rgba(0,0,0,.28); box-sizing: border-box; overflow: hidden; }
.sui-rich-sheet-body { padding: 0 !important; height: auto; min-height: 0; overflow: visible !important; max-width: none !important; width: auto !important; flex: none; }
.sui-rich-sheet-body img { max-height: 100%; }
.sui-rich-sheet-header { position: absolute; left: 64px; right: 64px; top: 26px; font-size: 11px; color: #65718a; text-align: center; border-bottom: 1px solid #d6dce8; padding-bottom: 3px; }
.sui-rich-sheet-footer { position: absolute; left: 64px; right: 64px; bottom: 34px; font-size: 11px; color: #65718a; text-align: center; border-top: 1px solid #d6dce8; padding-top: 3px; }
.sui-rich-sheet-number { position: absolute; right: 18px; bottom: 8px; font-size: 10px; color: #8a94a6; }
.sui-rich-zoom-btn.sui-rich-pages-btn { width: auto; min-width: 72px; height: 28px; padding: 0 12px; font-size: 12px; white-space: nowrap; flex: 0 0 auto; }
.sui-rich-pages-btn.on { background: var(--sui-rich-accent); color: #fff; border-color: var(--sui-rich-accent); }
.sui-rich-page { flex: 1 1 auto; min-width: 0; display: flex; flex-direction: column; background: var(--sui-rich-paper); }
.sui-rich-pageheader { padding: 9px 26px 5px; border-bottom: 1px solid var(--sui-rich-paper-border); font-size: 11px; color: var(--sui-rich-paper-muted); text-align: center; }
.sui-rich-pagefooter { padding: 5px 26px 9px; border-top: 1px solid var(--sui-rich-paper-border); font-size: 11px; color: var(--sui-rich-paper-muted); text-align: center; }
.sui-rich-findbar { position: absolute; top: 8px; right: 14px; z-index: 25; display: flex; align-items: center; gap: 4px; padding: 5px 6px; background: var(--sui-rich-bg); border: 1px solid var(--sui-rich-border); border-radius: 9px; box-shadow: 0 10px 24px rgba(0,0,0,.25); color: var(--sui-rich-fg); }
.sui-rich-findbar[hidden] { display: none; }
.sui-rich-find-input { height: 26px; width: 140px; padding: 0 7px; border: 1px solid var(--sui-rich-border); border-radius: 6px; background: var(--sui-rich-bg); color: var(--sui-rich-fg); font: inherit; font-size: 12px; }
.sui-rich-find-count { font-size: 11px; color: var(--sui-rich-muted); min-width: 36px; text-align: center; }
.sui-rich-find-btn { height: 26px; min-width: 24px; padding: 0 6px; border: 1px solid var(--sui-rich-border); border-radius: 6px; background: transparent; color: var(--sui-rich-fg); cursor: pointer; font: inherit; font-size: 12px; }
.sui-rich-find-btn:hover { background: var(--sui-rich-hover); }
::highlight(sui-rich-find) { background: rgba(255, 214, 0, .35); }
::highlight(sui-rich-find-current) { background: rgba(255, 149, 0, .75); color: #111; }
.sui-rich-tool-group.floating { position: fixed; z-index: 940; padding: 5px; flex-wrap: wrap; max-width: calc(100vw - 16px); background: var(--sui-rich-bg); border-color: var(--sui-rich-border); box-shadow: 0 12px 30px rgba(0,0,0,.3); }
.sui-rich-tool-group.floating .sui-rich-tool-dock { display: inline-flex; }
.sui-rich-group { display: flex; align-items: center; gap: 2px; padding-right: 5px; border-right: 1px solid var(--sui-rich-border); }
.sui-rich-group:last-child { border-right: 0; padding-right: 0; }
.sui-rich-theme-group { margin-left: auto; }
.sui-rich-theme-select { height: 28px; }
.sui-rich-btn { display: inline-flex; align-items: center; justify-content: center; min-width: 36px; height: 34px; padding: 0 8px; border: 1px solid transparent; border-radius: 7px; background: transparent; color: var(--sui-rich-fg); cursor: pointer; font: inherit; font-size: 14px; line-height: 1; }
.sui-rich-btn svg { width: 22px; height: 22px; fill: none; stroke: currentColor; stroke-width: 1.9; stroke-linecap: round; stroke-linejoin: round; }
.sui-rich-btn b, .sui-rich-btn strong { font-size: 18px; }
.sui-rich-btn i, .sui-rich-btn em { font-size: 18px; }
.sui-rich-btn u, .sui-rich-btn s { font-size: 17px; }
.sui-rich[data-sui-rich-icons="bold"] .sui-rich-btn svg { width: 24px; height: 24px; stroke-width: 2.6; }
.sui-rich[data-sui-rich-icons="bold"] .sui-rich-btn b, .sui-rich[data-sui-rich-icons="bold"] .sui-rich-btn strong { font-size: 20px; font-weight: 800; }
.sui-rich[data-sui-rich-icons="duotone"] .sui-rich-btn svg { width: 23px; height: 23px; fill: currentColor; fill-opacity: .22; stroke-width: 1.7; }
.sui-rich[data-sui-rich-icons="blocks"] .sui-rich-btn { background: color-mix(in srgb, var(--sui-rich-fg) 7%, transparent); border-color: color-mix(in srgb, var(--sui-rich-border) 70%, transparent); border-radius: 9px; }
.sui-rich[data-sui-rich-icons="blocks"] .sui-rich-btn:hover, .sui-rich[data-sui-rich-icons="blocks"] .sui-rich-btn:focus-visible { background: var(--sui-rich-hover); }
.sui-rich[data-sui-rich-icons="blocks"] .sui-rich-btn.on { background: color-mix(in srgb, var(--sui-rich-accent) 18%, var(--sui-rich-bg)); }
.sui-rich-btn:hover, .sui-rich-btn:focus-visible { background: var(--sui-rich-hover); border-color: var(--sui-rich-border); outline: none; }
.sui-rich-btn:active { transform: translateY(1px); }
.sui-rich-btn.on { background: var(--sui-rich-hover); color: var(--sui-rich-accent); border-color: var(--sui-rich-accent); }
.sui-rich-btn[disabled] { opacity: .45; cursor: default; }
.sui-rich[data-rich-painter="1"] .sui-rich-editor { cursor: copy; }
.sui-rich-guides { position: absolute; inset: 0; overflow: hidden; pointer-events: none; z-index: 4; }
.sui-rich-guide { position: absolute; left: 0; right: 0; height: 2px; margin-top: -1px; border-top: 2px dashed color-mix(in srgb, var(--sui-rich-break, #dc2626) 72%, transparent); opacity: .9; pointer-events: auto; cursor: grab; }
.sui-rich-guide::before { content: ''; position: absolute; left: 0; right: 0; top: -30px; height: 44px; }
.sui-rich-guide::after { content: ''; position: absolute; left: 50%; top: -6px; width: 10px; height: 10px; margin-left: -5px; border-radius: 50%; background: var(--sui-rich-break, #dc2626); opacity: 0; transition: opacity .12s ease; }
.sui-rich-guide-label { position: absolute; left: 10px; top: -20px; padding: 2px 7px; border-radius: 4px; background: var(--sui-rich-break, #dc2626); color: #fff; font: 600 10px/1.2 system-ui, sans-serif; letter-spacing: .06em; text-transform: uppercase; white-space: nowrap; opacity: 0; pointer-events: auto; cursor: grab; transition: opacity .12s ease; }
.sui-rich-guide:hover, .sui-rich-guide.dragging { border-top-style: solid; border-top-color: var(--sui-rich-break, #dc2626); opacity: 1; }
.sui-rich-guide:hover::after, .sui-rich-guide.dragging::after, .sui-rich-guide:hover .sui-rich-guide-label, .sui-rich-guide.dragging .sui-rich-guide-label { opacity: 1; }
.sui-rich-guide.dragging { cursor: grabbing; }
.sui-rich-guide.dragging .sui-rich-guide-label { cursor: grabbing; }
.sui-rich-paginated .sui-rich-guides { display: none; }
.sui-rich-imagesettings-error { margin: 0; color: #fca5a5; font-size: 12px; }
.sui-rich-image-menu { position: fixed; z-index: 6200; width: 182px; padding: 4px; border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 8px; background: var(--sui-rich-bg, #1d2129); box-shadow: 0 12px 30px rgba(0,0,0,.4); }
.sui-rich-image-menu .sui-rich-menu-item { width: 100%; text-align: left; }
.sui-rich-editor img.sui-rich-image-selected { outline: 2px solid var(--sui-rich-accent); outline-offset: 2px; }
.sui-rich-select, .sui-rich-color, .sui-rich-prop { height: 28px; border: 1px solid var(--sui-rich-border); border-radius: 6px; background: var(--sui-rich-bg); color: var(--sui-rich-fg); font: inherit; font-size: 12px; padding: 0 5px; }
.sui-rich-color-control { position: relative; display: inline-flex; align-items: center; justify-content: center; gap: 4px; width: 48px; height: 34px; padding: 0 5px; border: 1px solid transparent; border-radius: 7px; cursor: pointer; background: transparent; color: var(--sui-rich-fg); font: inherit; overflow: hidden; }
.sui-rich-color-control:hover, .sui-rich-color-control:focus-visible { background: var(--sui-rich-hover); border-color: var(--sui-rich-border); outline: none; }
.sui-rich-color-icon { position: relative; display: inline-flex; align-items: center; justify-content: center; width: 22px; height: 22px; font-size: 15px; line-height: 1; pointer-events: none; }
.sui-rich-color-icon::after { content: ''; position: absolute; left: 1px; right: 1px; bottom: 0; height: 4px; border-radius: 2px; background: var(--sui-rich-color-current, #000); }
.sui-rich-color-glyph { font-weight: 700; font-size: 18px; line-height: 1; }
.sui-rich-color-glyph-highlight { padding: 0 3px; border-radius: 3px; background: var(--sui-rich-color-current, #fff); color: #111; box-shadow: inset 0 0 0 1px rgba(0,0,0,.22); }
.sui-rich-color { position: absolute; inset: 0; width: 100%; height: 100%; min-width: 0; padding: 0; border: 0; opacity: 0; pointer-events: none; cursor: pointer; z-index: 2; }
.sui-rich-color::-webkit-color-swatch-wrapper { padding: 0; }
.sui-rich-color::-webkit-color-swatch { border: 0; }
.sui-rich-color::-moz-color-swatch { border: 0; }
.sui-rich-color-palette-btn { width: 18px; min-width: 18px; height: 34px; padding: 0; border: 0; border-radius: 5px; background: transparent; color: var(--sui-rich-muted); cursor: pointer; font: inherit; font-size: 11px; line-height: 1; }
.sui-rich-color-palette-btn:hover, .sui-rich-color-palette-btn:focus-visible { background: var(--sui-rich-hover); color: var(--sui-rich-fg); outline: none; }
.sui-rich-color-palette { position: fixed; z-index: 7300; width: 196px; padding: 8px; border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 9px; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); box-shadow: 0 12px 30px rgba(0,0,0,.4); }
.sui-rich-color-palette-title { margin: 2px 0 5px; color: var(--sui-rich-muted, #9aa7b8); font-size: 10.5px; font-weight: 600; text-transform: uppercase; letter-spacing: .04em; }
.sui-rich-color-grid { display: grid; grid-template-columns: repeat(6, 24px); gap: 5px; }
.sui-rich-color-swatch { width: 24px; height: 24px; padding: 0; border: 1px solid rgba(0,0,0,.25); border-radius: 5px; cursor: pointer; }
.sui-rich-color-swatch:hover, .sui-rich-color-swatch:focus-visible { outline: 2px solid var(--sui-rich-accent, #6f7cff); outline-offset: 1px; }
.sui-rich-color-custom { width: 100%; margin-top: 7px; justify-content: center; border: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-body { display: flex; min-height: 220px; background: var(--sui-rich-paper); position: relative; }
.sui-rich-editor { flex: 1 1 auto; min-width: 0; height: var(--sui-rich-content-height, 320px); min-height: 0; padding: 22px 26px; outline: none; overflow: auto; font-size: 16px; line-height: 1.65; color: var(--sui-rich-paper-fg); overflow-wrap: break-word; word-break: normal; font-family: system-ui, 'Segoe UI Symbol', 'Noto Sans Symbols 2', sans-serif; }
.sui-rich-editor:empty::before, .sui-rich-editor[data-placeholder]:empty::before { content: attr(data-placeholder); color: var(--sui-rich-muted); pointer-events: none; }
.sui-rich-editor p, .sui-rich-editor h1, .sui-rich-editor h2, .sui-rich-editor h3, .sui-rich-editor h4, .sui-rich-editor blockquote, .sui-rich-editor pre, .sui-rich-editor figure, .sui-rich-editor table { margin: 0 0 .8em; }
.sui-rich-editor h1 { font-size: 1.9em; line-height: 1.25; }
.sui-rich-editor h2 { font-size: 1.55em; line-height: 1.3; }
.sui-rich-editor h3 { font-size: 1.28em; line-height: 1.35; }
.sui-rich-editor h4 { font-size: 1.1em; }
.sui-rich-editor blockquote { border-left: 3px solid var(--sui-rich-accent); padding: .2em 0 .2em 14px; color: var(--sui-rich-paper-muted); }
.sui-rich-editor pre { background: color-mix(in srgb, var(--sui-rich-paper-fg) 7%, transparent); color: var(--sui-rich-paper-fg); border: 1px solid var(--sui-rich-paper-border); border-radius: 8px; padding: 12px; overflow: auto; font-family: ui-monospace, Consolas, monospace; font-size: 13px; line-height: 1.5; white-space: pre-wrap; tab-size: 4; }
.sui-rich-editor code { font-family: ui-monospace, Consolas, monospace; font-size: .92em; color: var(--sui-rich-paper-fg); background: rgba(127,127,127,.16); border-radius: 4px; padding: .1em .3em; }
.sui-rich-editor pre code { display: block; padding: 0; border-radius: 0; background: transparent; color: inherit; font-size: 1em; white-space: inherit; }
.sui-rich-editor pre[data-line-numbers] { position: relative; padding-left: 52px; white-space: pre; overflow: auto; }
.sui-rich-editor pre[data-line-numbers] code { white-space: pre; }
.sui-rich-editor pre[data-line-numbers]::before { content: attr(data-line-numbers); position: absolute; left: 0; top: 12px; width: 36px; padding-right: 8px; border-right: 1px solid var(--sui-rich-paper-border); color: var(--sui-rich-paper-muted); text-align: right; white-space: pre; pointer-events: none; font: inherit; font-size: 12px; line-height: 1.5; }
.sui-rich-editor .sui-code-keyword { color: #7c3aed; font-weight: 600; }
.sui-rich-editor .sui-code-string { color: #0a7a3d; }
.sui-rich-editor .sui-code-comment { color: #6b7280; font-style: italic; }
.sui-rich-editor .sui-code-number { color: #b45309; }
.sui-rich-editor .sui-code-tag { color: #1d4ed8; }
.sui-rich[data-sui-rich-content="dark"] .sui-rich-editor .sui-code-keyword { color: #c4b5fd; }
.sui-rich[data-sui-rich-content="dark"] .sui-rich-editor .sui-code-string { color: #86efac; }
.sui-rich[data-sui-rich-content="dark"] .sui-rich-editor .sui-code-comment { color: #94a3b8; }
.sui-rich[data-sui-rich-content="dark"] .sui-rich-editor .sui-code-number { color: #fbbf24; }
.sui-rich[data-sui-rich-content="dark"] .sui-rich-editor .sui-code-tag { color: #93c5fd; }
.sui-rich-editor hr { border: 0; border-top: 1px solid var(--sui-rich-paper-border); margin: 1.2em 0; }
.sui-rich-editor img { max-width: 100%; height: auto; border-radius: 6px; background: transparent; border: 0; box-shadow: none; padding: 0; object-fit: contain; }
.sui-rich-editor figure { text-align: center; }
.sui-rich-editor figcaption { color: var(--sui-rich-paper-muted); font-size: .85em; margin-top: 4px; }
.sui-rich-editor table { width: auto; max-width: 100%; border-collapse: collapse; }
.sui-rich-editor .sui-rich-cell-blocks > *:last-child { margin-bottom: 0; }
.sui-rich-editor .sui-rich-cell-blocks table { margin: 0; }
.sui-rich-editor th, .sui-rich-editor td { border: 1px solid var(--sui-rich-paper-border); padding: 6px 8px; vertical-align: top; min-width: 40px; }
.sui-rich-editor .sui-rich-pagebreak { display: flex; align-items: center; gap: 6px; margin: 16px 0; padding-top: 5px; border-top: 2px dashed var(--sui-rich-break, #dc2626); color: var(--sui-rich-break, #dc2626); font-size: 10px; letter-spacing: .08em; text-transform: uppercase; user-select: none; cursor: grab; }
.sui-rich-editor .sui-rich-pagebreak:active { cursor: grabbing; }
.sui-rich-editor .sui-rich-pagebreak.dragging { opacity: .92; cursor: grabbing; background: color-mix(in srgb, var(--sui-rich-break, #dc2626) 7%, transparent); border-radius: 4px; box-shadow: 0 4px 14px color-mix(in srgb, var(--sui-rich-break, #dc2626) 22%, transparent); }
.sui-rich-pagebreak-handle { cursor: grab; font-size: 12px; padding: 0 2px; }
.sui-rich-pagebreak-handle:active { cursor: grabbing; }
.sui-rich-pagebreak-label { flex: 1 1 auto; }
.sui-rich-pagebreak-delete { border: 1px solid color-mix(in srgb, var(--sui-rich-break, #dc2626) 45%, transparent); border-radius: 4px; background: transparent; color: var(--sui-rich-break, #dc2626); font: inherit; font-size: 12px; line-height: 1; padding: 1px 5px; cursor: pointer; }
.sui-rich-pagebreak-delete:hover { color: #fff; background: var(--sui-rich-break, #dc2626); border-color: var(--sui-rich-break, #dc2626); }
.sui-rich-editor tbody tr:hover, .sui-rich-editor thead tr:hover { background: transparent; }
.sui-rich-editor th { background: color-mix(in srgb, var(--sui-rich-paper-fg) 8%, transparent); text-align: left; }
.sui-rich-editor a { color: var(--sui-rich-accent); text-decoration: underline; }
.sui-rich-editor ul, .sui-rich-editor ol { padding-left: 1.6em; }
.sui-rich-editor li { margin: 0; }
.sui-rich-editor li > p { margin: 0 0 .15em; }
.sui-rich-editor li > p:last-child { margin-bottom: 0; }
.sui-rich-editor li::marker { color: var(--sui-rich-paper-fg); font-weight: 600; }
.sui-rich-props { width: 210px; border-left: 1px solid var(--sui-rich-border); padding: 12px; display: flex; flex-direction: column; gap: 8px; background: var(--sui-rich-toolbar-bg); }
.sui-rich-props-title { font-size: 12px; font-weight: 700; color: var(--sui-rich-muted); text-transform: uppercase; letter-spacing: .05em; }
.sui-rich-prop { width: 100%; }
.sui-rich-prop-btn { border: 1px solid var(--sui-rich-border); border-radius: 6px; background: transparent; color: var(--sui-rich-fg); padding: 5px 8px; font: inherit; font-size: 12px; text-align: left; cursor: pointer; }
.sui-rich-prop-btn:hover { background: var(--sui-rich-hover); }
.sui-rich-status { display: flex; align-items: center; flex-wrap: wrap; gap: 10px; padding: 5px 10px; border-top: 1px solid var(--sui-rich-border); background: var(--sui-rich-toolbar-bg); color: var(--sui-rich-muted); font-size: 11px; }
.sui-rich-stats { margin-right: auto; }
.sui-rich-saved { color: var(--sui-rich-muted); }
.sui-rich-zoom-btn { width: 22px; height: 22px; border: 1px solid var(--sui-rich-border); border-radius: 5px; background: transparent; color: var(--sui-rich-fg); cursor: pointer; line-height: 1; white-space: nowrap; }
.sui-rich-zoom-label { width: auto; min-width: 48px; padding: 0 6px; font-size: 11px; font-variant-numeric: tabular-nums; }
.sui-rich-context { position: fixed; z-index: 900; display: flex; gap: 2px; padding: 4px; background: var(--sui-rich-bg); border: 1px solid var(--sui-rich-border); border-radius: 8px; box-shadow: 0 12px 30px rgba(0,0,0,.3); }
.sui-rich-context[hidden] { display: none; }
.sui-rich-tableprops { position: fixed; z-index: 965; width: 262px; padding: 10px 12px; background: var(--sui-rich-bg); border: 1px solid var(--sui-rich-border); border-radius: 10px; box-shadow: 0 14px 34px rgba(0,0,0,.35); color: var(--sui-rich-fg); font-size: 12px; }
.sui-rich-prompt { position: fixed; inset: 0; z-index: 6200; display: flex; align-items: center; justify-content: center; padding: 24px; background: rgba(0, 0, 0, .55); }
.sui-rich-prompt-card { width: min(440px, 92vw); max-height: 90vh; overflow: auto; padding: 16px; border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 12px; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); box-shadow: 0 18px 50px rgba(0,0,0,.5); }
.sui-rich-prompt-field { display: flex; flex-direction: column; gap: 4px; margin-bottom: 10px; font-size: 12px; color: var(--sui-rich-muted, #8fa0b8); }
.sui-rich-prompt-field input, .sui-rich-prompt-field textarea { width: 100%; border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 6px; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); font: inherit; font-size: 13px; padding: 7px 8px; }
.sui-rich-prompt-field textarea { min-height: 150px; resize: vertical; font-family: ui-monospace, Consolas, monospace; }
.sui-rich-prompt-error { margin: 0 0 8px; color: #fca5a5; font-size: 12px; }
.sui-rich-prompt-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 14px; }
.sui-rich-prompt-actions .sui-rich-btn { border: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-preview { position: fixed; inset: 0; z-index: 6000; display: flex; align-items: center; justify-content: center; padding: 24px; background: rgba(0, 0, 0, .55); }
.sui-rich-preview-card { width: min(920px, 94vw); height: min(90vh, 900px); display: flex; flex-direction: column; overflow: hidden; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 12px; box-shadow: 0 18px 50px rgba(0,0,0,.5); }
.sui-rich-preview-bar { display: flex; align-items: center; gap: 8px; padding: 8px 10px; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); font-size: 13px; border-bottom: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-preview-bar strong { margin-right: auto; font-weight: 600; }
.sui-rich-preview-bar .sui-rich-btn { background: transparent; color: var(--sui-rich-fg, #e6e6ef); border: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-preview-bar .sui-rich-btn:hover, .sui-rich-preview-bar .sui-rich-btn:focus-visible { background: var(--sui-rich-hover, #2d3340); color: var(--sui-rich-fg, #e6e6ef); border-color: var(--sui-rich-accent, #4c5cf7); outline: none; }
.sui-rich-preview-pages { flex: 1 1 auto; min-height: 0; overflow: auto; padding: 18px; background: #e6e9ef; }
.sui-rich-preview-pages .sui-rich-page-sheet { margin: 0 auto 18px; }
.sui-rich-preview-loading { color: #333; font-size: 13px; padding: 18px; text-align: center; }
@media print {
  body.sui-rich-printing > *:not(.sui-rich-preview) { display: none !important; }
  body.sui-rich-printing .sui-rich-preview { position: static; inset: auto; display: block; padding: 0; background: #fff; }
  body.sui-rich-printing .sui-rich-preview-card { width: auto; height: auto; display: block; border: 0; border-radius: 0; box-shadow: none; background: #fff; }
  body.sui-rich-printing .sui-rich-preview-bar { display: none; }
  body.sui-rich-printing .sui-rich-preview-pages { overflow: visible; height: auto; padding: 0; background: #fff; }
  body.sui-rich-printing .sui-rich-preview-pages .sui-rich-page-sheet { margin: 0; box-shadow: none; break-after: page; page-break-after: always; }
  body.sui-rich-printing .sui-rich-preview-pages .sui-rich-page-sheet:last-child { break-after: auto; page-break-after: auto; }
  .sui-rich-guides { display: none !important; }
}
.sui-rich-pagesettings, .sui-rich-documentsettings, .sui-rich-paragraphsettings, .sui-rich-imagesettings { position: fixed; z-index: 5000; width: 276px; padding: 10px 12px; background: var(--sui-rich-bg); border: 1px solid var(--sui-rich-border); border-radius: 10px; box-shadow: var(--sui-rich-shadow, 0 14px 34px rgba(0,0,0,.35)); color: var(--sui-rich-fg); font-size: 12px; display: flex; flex-direction: column; gap: 8px; }
.sui-rich-pagesettings[hidden], .sui-rich-documentsettings[hidden], .sui-rich-paragraphsettings[hidden], .sui-rich-imagesettings[hidden] { display: none; }
.sui-rich-pagesettings label, .sui-rich-documentsettings label, .sui-rich-paragraphsettings label, .sui-rich-imagesettings label { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.sui-rich-pagesettings select, .sui-rich-pagesettings input, .sui-rich-documentsettings select, .sui-rich-documentsettings input, .sui-rich-paragraphsettings select, .sui-rich-paragraphsettings input, .sui-rich-imagesettings select, .sui-rich-imagesettings input { height: 26px; border: 1px solid var(--sui-rich-border); border-radius: 6px; background: var(--sui-rich-bg); color: var(--sui-rich-fg); font: inherit; font-size: 12px; padding: 0 6px; min-width: 0; }
.sui-rich-pagesettings input[type="number"], .sui-rich-paragraphsettings input[type="number"] { width: 72px; }
.sui-rich-pagesettings label > span, .sui-rich-documentsettings label > span, .sui-rich-paragraphsettings label > span, .sui-rich-imagesettings label > span { flex: none; }
.sui-rich-pagesettings label > input[type="text"], .sui-rich-documentsettings label > input[type="text"], .sui-rich-paragraphsettings label > input[type="text"], .sui-rich-imagesettings label > input[type="text"] { flex: 1 1 auto; width: 100%; min-width: 0; }
.sui-rich-pagesettings-grid, .sui-rich-paragraphsettings .sui-rich-pagesettings-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 6px 10px; }
.sui-rich-hint { display: inline-flex; align-items: center; justify-content: center; width: 14px; height: 14px; margin-left: 6px; border: 1px solid var(--sui-rich-border); border-radius: 50%; color: var(--sui-rich-muted); font-size: 10px; font-weight: 700; line-height: 1; text-transform: none; cursor: help; vertical-align: middle; }
.sui-rich-hint:hover { color: var(--sui-rich-fg); border-color: var(--sui-rich-accent); }
.sui-rich-pagesettings-actions { display: flex; gap: 6px; }
.sui-rich-indent-actions { display: flex; gap: 6px; margin-top: -2px; }
.sui-rich-indent-actions .sui-rich-btn { flex: 1; justify-content: center; }
.sui-rich-paragraph-check input[type="checkbox"] { width: 16px; height: 16px; accent-color: var(--sui-rich-accent); }
.sui-rich-chrome { display: flex; flex-direction: column; gap: 4px; }
.sui-rich-chrome-tools { display: flex; flex-wrap: wrap; gap: 4px; }
.sui-rich-chrome-tools .sui-rich-btn { font-size: 11px; padding: 3px 6px; }
.sui-rich-token { font-family: ui-monospace, Consolas, monospace; font-size: 11px; }
.sui-rich-chrome-preview { color: var(--sui-rich-muted); font-size: 11px; min-height: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sui-rich-chrome-summary-row { display: flex; flex-wrap: wrap; align-items: center; gap: 5px; }
.sui-rich-chrome-summary-label { min-width: 54px; color: var(--sui-rich-muted); font-size: 11px; }
.sui-rich-chrome-summary { flex: 1 1 100px; color: var(--sui-rich-muted); font-size: 11px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sui-rich-chrome-summary-row .sui-rich-btn { font-size: 11px; padding: 3px 7px; }
.sui-rich-chrome-open { width: 100%; margin: 7px 0 2px; justify-content: center; border: 1px solid var(--sui-rich-border); background: var(--sui-rich-hover); }
.sui-rich-chrome-content p, .sui-rich-chrome-content h1, .sui-rich-chrome-content h2, .sui-rich-chrome-content h3, .sui-rich-chrome-content h4 { margin: 0 0 .2em; }
.sui-rich-chrome-content > :last-child { margin-bottom: 0; }
.sui-rich-chrome-content img { max-width: 100%; max-height: 64px; object-fit: contain; }
.sui-rich-chrome-editor { position: fixed; inset: 0; z-index: 7200; display: flex; align-items: center; justify-content: center; padding: 24px; background: rgba(0,0,0,.55); }
.sui-rich-chrome-editor-card { width: min(820px, 96vw); max-height: min(88vh, 860px); display: flex; flex-direction: column; overflow: hidden; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 12px; box-shadow: 0 18px 50px rgba(0,0,0,.5); }
.sui-rich-chrome-editor-head { display: flex; align-items: center; gap: 8px; padding: 10px 12px; border-bottom: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-chrome-editor-head strong { margin-right: auto; font-weight: 600; }
.sui-rich-chrome-tabs { display: flex; gap: 4px; padding: 8px 12px 0; }
.sui-rich-chrome-tab { border: 1px solid var(--sui-rich-border, #3d4556); background: transparent; color: var(--sui-rich-fg, #e6e6ef); border-radius: 6px 6px 0 0; padding: 5px 12px; cursor: pointer; font: inherit; font-size: 12px; }
.sui-rich-chrome-tab.on { background: var(--sui-rich-hover, #242b38); }
.sui-rich-chrome-toolbar { display: flex; flex-wrap: wrap; gap: 4px; align-items: center; padding: 8px 12px; border-bottom: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-chrome-toolbar .sui-rich-btn { min-width: 28px; padding: 4px 7px; border: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-chrome-vars { height: 28px; min-width: 110px; border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 6px; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); font: inherit; font-size: 11px; }
.sui-rich-chrome-surfaces { flex: 1; overflow: auto; padding: 12px; background: color-mix(in srgb, var(--sui-rich-fg, #e6e6ef) 5%, transparent); }
.sui-rich-chrome-surface { min-height: 180px; padding: 14px 16px; background: var(--sui-rich-paper, #ffffff); color: var(--sui-rich-paper-fg, #1c2330); border: 1px solid var(--sui-rich-paper-border, #c9d3e0); border-radius: 8px; outline: none; overflow: auto; font-family: Georgia, 'Times New Roman', serif; font-size: 13px; line-height: 1.45; }
.sui-rich-chrome-surface[hidden] { display: none; }
.sui-rich-chrome-surface img { max-width: 100%; max-height: 100px; object-fit: contain; }
.sui-rich-chrome-surface p { margin: 0 0 .35em; }
.sui-rich-chrome-actions { display: flex; justify-content: flex-end; gap: 6px; padding: 10px 12px; border-top: 1px solid var(--sui-rich-border, #3d4556); }
.sui-rich-chrome-help-overlay { position: fixed; inset: 0; z-index: 7300; display: flex; align-items: center; justify-content: center; padding: 24px; background: rgba(0,0,0,.5); }
.sui-rich-chrome-help-card { width: min(560px, 92vw); max-height: 80vh; overflow: auto; padding: 16px; background: var(--sui-rich-bg, #1d2129); color: var(--sui-rich-fg, #e6e6ef); border: 1px solid var(--sui-rich-border, #3d4556); border-radius: 10px; box-shadow: 0 18px 50px rgba(0,0,0,.5); }
.sui-rich-chrome-help-row { display: flex; gap: 10px; padding: 4px 0; border-bottom: 1px solid var(--sui-rich-border, #3d4556); font-size: 12px; }
.sui-rich-chrome-help-code { min-width: 100px; font-family: ui-monospace, Consolas, monospace; color: var(--sui-rich-accent, #6f7cff); }
.sui-rich-chrome-help-note { margin: 10px 0; color: var(--sui-rich-muted, #9aa7b8); font-size: 11px; line-height: 1.4; }
.sui-rich-pagesettings-actions .sui-rich-btn { flex: 1; border: 1px solid var(--sui-rich-border); }
.sui-rich-tableprops[hidden] { display: none; }
.sui-rich-tableprops-title { font-weight: 600; margin-bottom: 6px; }
.sui-rich-tableprops-row { display: flex; align-items: center; gap: 8px; margin: 5px 0; }
.sui-rich-tableprops-row label { flex: 1; color: var(--sui-rich-muted); }
.sui-rich-tableprops-row input, .sui-rich-tableprops-row select { width: 96px; height: 26px; border: 1px solid var(--sui-rich-border); border-radius: 6px; background: var(--sui-rich-bg); color: var(--sui-rich-fg); font: inherit; padding: 0 6px; }
.sui-rich-tableprops-row input[type="color"] { padding: 1px; }
.sui-rich-tableprops-note { margin: 6px 0 2px; color: var(--sui-rich-muted); font-size: 10.5px; line-height: 1.35; }
.sui-rich-tableprops-close { margin-top: 6px; width: 100%; border: 1px solid var(--sui-rich-border); border-radius: 6px; background: transparent; color: var(--sui-rich-fg); padding: 5px; cursor: pointer; }
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
.sui-rich[data-sui-rich-template="article"] .sui-rich-editor { max-width: 780px; margin: 0 auto; width: 100%; font-family: Georgia, 'Times New Roman', 'Segoe UI Symbol', serif; font-size: 17px; line-height: 1.75; }
.sui-rich[data-sui-rich-template="letter"] .sui-rich-editor { max-width: 680px; margin: 0 auto; width: 100%; font-family: Georgia, 'Times New Roman', 'Segoe UI Symbol', serif; font-size: 16px; line-height: 1.8; }
.sui-rich[data-sui-rich-template="report"] .sui-rich-editor { max-width: 960px; margin: 0 auto; width: 100%; font-family: system-ui, 'Segoe UI Symbol', sans-serif; font-size: 15px; line-height: 1.6; }
.sui-rich[data-sui-rich-template="compact"] .sui-rich-editor { padding: 12px 14px; font-size: 14px; line-height: 1.5; }
@media (prefers-reduced-motion: reduce) {
  .sui-rich-btn, .sui-rich-menu-btn, .sui-rich-menu-item, .sui-rich-prop-btn { transition: none !important; }
}
`

// progUsesRichText reports whether the program references the editor tag or
// the richtext API, so the runtime is emitted only when needed.
func progUsesRichText(p *Prog) bool {
	return scanProgFeatures(p).needsRichText()
}
