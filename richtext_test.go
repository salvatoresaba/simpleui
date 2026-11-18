package main

import (
	"strings"
	"testing"
)

func TestRichTextWeb(t *testing.T) {
	src := `import { json } from "std"
state doc = ""
state out = ""
state status = "ready"

fn changed() -> null {
	status = "changed " + str(len(doc))
}

async fn run() -> null {
	let sample = {"type": "doc", "children": [
		{"type": "heading", "level": 1, "children": [{"type": "text", "text": "Heading", "marks": []}]},
		{"type": "paragraph", "children": [{"type": "text", "text": "Hello world", "marks": []}]}
	]}
	okSet, errSet = richtext.set("editor", sample)
	jsonDoc, errGet = richtext.get("editor")
	okBold, errBold = richtext.command("editor", "bold")
	okHeading, errHeading = richtext.command("editor", "h2")
	okList, errList = richtext.command("editor", "bulletList")
	okImage, errImage = richtext.insertImage("editor", "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxMCIgaGVpZ2h0PSIxMCI+PC9zdmc+", "dot")
	okLink, errLink = richtext.insertLink("editor", "https://example.com", "Example")
	okTable, errTable = richtext.insertTable("editor", 2, 2)
	md, errMd = richtext.exportFormat("editor", "markdown")
	html, errHtml = richtext.exportFormat("editor", "html")
	xmlText, errXml = richtext.exportFormat("editor", "xml")
	jsonText, errJson = richtext.exportFormat("editor", "json")
	okImport, errImport = richtext.importFormat("editor", "# Imported\n\n**bold** text\n\n1. one\n2. two", "markdown")
	afterImport, errAfter = richtext.get("editor")
	okHtmlImport, errHtmlImport = richtext.importFormat("editor", "<h2>From HTML</h2><p>safe <strong>text</strong><script>bad()</script></p>", "html")
	afterHtml, errAfterHtml = richtext.get("editor")
	okXmlImport, errXmlImport = richtext.importFormat("editor", "<document><paragraph><text marks=\"bold\">From XML</text></paragraph></document>", "xml")
	okUndo, errUndo = richtext.undo("editor")
	okRedo, errRedo = richtext.redo("editor")
	statusText, errStatus = richtext.status("editor")
	selectionText, errSelection = richtext.selection("editor")
	okPlugin, errPlugin = richtext.registerPlugin("demo", {"commands": {"shout": {"title": "Shout", "command": "bold"}}, "toolbar": [{"id": "shout", "label": "Shout"}]})
	pluginList, errPlugins = richtext.plugins()
	okPluginCmd, errPluginCmd = richtext.command("editor", "shout")
	okTemplate, errTemplate = richtext.setTemplate("editor", "letter")
	okToolbar, errToolbar = richtext.setToolbar("editor", "undo,redo|bold,italic,underline|link,image,table")
	out = "SET=" + str(okSet) + "|" + str(contains(jsonDoc, "Heading")) + "|" + str(errSet == null) + "|" + str(errGet == null)
	out = out + "\nCMDS=" + str(okBold) + "|" + str(okHeading) + "|" + str(okList) + "|" + str(okImage) + "|" + str(okLink) + "|" + str(okTable)
	out = out + "\nEXPORT=" + str(len(md)) + "|" + str(len(html)) + "|" + str(len(xmlText)) + "|" + str(len(jsonText)) + "|" + str(errMd == null) + "|" + str(errHtml == null) + "|" + str(errXml == null) + "|" + str(errJson == null)
	out = out + "\nIMPORT=" + str(okImport) + "|" + str(contains(afterImport, "Imported")) + "|" + str(okHtmlImport) + "|" + str(contains(afterHtml, "From HTML")) + "|" + str(contains(afterHtml, "bad")) + "|" + str(okXmlImport) + "|" + str(errAfter == null) + "|" + str(errAfterHtml == null)
	out = out + "\nHISTORY=" + str(okUndo) + "|" + str(okRedo) + "|" + str(errUndo == null) + "|" + str(errRedo == null)
	out = out + "\nSTATUS=" + str(contains(statusText, "words")) + "|" + str(contains(selectionText, "collapsed")) + "|" + str(errStatus == null) + "|" + str(errSelection == null)
	out = out + "\nPLUGIN=" + str(okPlugin) + "|" + str(len(pluginList)) + "|" + str(okPluginCmd) + "|" + str(errPlugin == null) + "|" + str(errPluginCmd == null)
	out = out + "\nUI=" + str(okTemplate) + "|" + str(okToolbar) + "|" + str(errTemplate == null) + "|" + str(errToolbar == null)
}

async fn changedValue() -> null {
	status = "changed"
}

view main {
	col {
		richtext "Article" name: "editor" value: doc height: "320px" theme: "dark" header: "Report \{page\}/\{pages\}" footer: "Footer \{page\}" menu: true toolbar: true contextToolbar: true statusbar: true onChange: changedValue onInput: changedValue toolbarItems: "undo,redo|styleSelect,bold,italic,underline,bulletList,orderedList|link,image,table"
		button "Run" onClick: run
		p "{status}"
		pre "{out}"
	}
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(400);
    button('Run').click();
    for (var wi = 0; wi < 200 && document.body.innerText.indexOf('UI=') < 0; wi++) await wait(50);
    var text = document.body.innerText;
    if (text.indexOf('SET=true|true|true|true') < 0) throw new Error('set/get: ' + text);
    if (text.indexOf('CMDS=true|true|true|true|true|true') < 0) throw new Error('commands: ' + text);
    if (text.indexOf('EXPORT=') < 0 || text.indexOf('IMPORT=true|true|true|true|false|true|true|true') < 0) throw new Error('import/export: ' + text);
    if (text.indexOf('HISTORY=true|true|true|true') < 0) throw new Error('history: ' + text);
    if (text.indexOf('STATUS=true|true|true|true') < 0) throw new Error('status: ' + text);
    if (text.indexOf('PLUGIN=true|1|true|true|true') < 0) throw new Error('plugin: ' + text);
    if (text.indexOf('UI=true|true|true|true') < 0) throw new Error('ui config: ' + text);

    var root = document.querySelector('[data-sui-rich]');
    var editor = root.querySelector('.sui-rich-editor');
    var toolbarButtons = root.querySelectorAll('.sui-rich-toolbar [data-rich-cmd]').length;
    if (toolbarButtons < 8) throw new Error('toolbar not built: ' + toolbarButtons);
    if (!root.querySelector('.sui-rich-menu') || !root.querySelector('.sui-rich-status')) throw new Error('menu/status missing');
    editor.querySelector('p').textContent = 'Toolbar text';
    var textNode = editor.querySelector('p').firstChild;
    var range = document.createRange(); range.setStart(textNode, 0); range.setEnd(textNode, 7);
    var sel = window.getSelection(); sel.removeAllRanges(); sel.addRange(range);
    __SUI_RICH.command('editor', 'italic');
    await wait(200);
    if (editor.querySelector('em, i, [style*="font-style"]') === null) throw new Error('toolbar formatting failed: ' + editor.innerHTML);
    var before = __SUI_RICH.exportFormat('editor', 'json')[0];
    if (before.indexOf('Toolbar') < 0) throw new Error('JSON model missing text: ' + before);

    function richRoot() { return document.querySelector('[data-sui-rich]'); }
    var themeToggle = richRoot().querySelector('.sui-rich-toolbar [data-rich-theme-toggle]');
    if (!themeToggle) throw new Error('theme button missing from toolbar');
    if (richRoot().querySelector('.sui-rich-status [data-rich-theme]')) throw new Error('theme picker still in status bar');
    if (richRoot().querySelectorAll('.sui-rich-tool-group').length < 2) throw new Error('toolbar categories missing');
    if (richRoot().querySelectorAll('[data-rich-tool-handle]').length < 2) throw new Error('toolbar handles missing');
var themeBefore = richRoot().getAttribute('data-sui-rich-theme');
    themeToggle.click();
    await wait(80);
    if (richRoot().getAttribute('data-sui-rich-theme') === themeBefore) throw new Error('theme button did not switch');
    if (getComputedStyle(richRoot().querySelector('.sui-rich-body')).backgroundColor !== 'rgb(255, 255, 255)') throw new Error('page background changed with the interface theme');
    themeToggle.click();
    await wait(80);
    var firstGroup = richRoot().querySelector('.sui-rich-tool-group');
    var firstHandle = firstGroup.querySelector('[data-rich-tool-handle]');
    var hrect = firstHandle.getBoundingClientRect();
    firstHandle.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, clientX: hrect.left + 3, clientY: hrect.top + 3 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: 640, clientY: 620 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: 640, clientY: 620 }));
    await wait(60);
    if (!firstGroup.classList.contains('floating')) throw new Error('toolbar did not float');
    var dockBar = richRoot().querySelector('.sui-rich-toolbar');
    var dockRect = dockBar.getBoundingClientRect();
    var fh = firstGroup.querySelector('[data-rich-tool-handle]');
    var fr = fh.getBoundingClientRect();
    fh.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, clientX: fr.left + 3, clientY: fr.top + 3 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: dockRect.right - 12, clientY: dockRect.top + 10 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: dockRect.right - 12, clientY: dockRect.top + 10 }));
    await wait(60);
    if (firstGroup.classList.contains('floating')) throw new Error('toolbar did not dock on the bar');
    if (firstGroup.parentNode !== dockBar) throw new Error('toolbar left the dock');

    var richEditor = richRoot().querySelector('.sui-rich-editor');
    richEditor.innerHTML = '<p>One</p>';
    richEditor.focus();
    var one = richEditor.querySelector('p');
    var oneRange = document.createRange(); oneRange.selectNodeContents(one); oneRange.collapse(false);
    var oneSel = window.getSelection(); oneSel.removeAllRanges(); oneSel.addRange(oneRange);
    richEditor.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await wait(250);
    document.execCommand('insertText', false, 'Two');
    await wait(250);
    if (richRoot().querySelectorAll('.sui-rich-editor > p').length !== 2) throw new Error('Enter split: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    var firstBlock = richRoot().querySelector('.sui-rich-editor > p');
    var blockRange = document.createRange(); blockRange.selectNodeContents(firstBlock);
    var blockSel = window.getSelection(); blockSel.removeAllRanges(); blockSel.addRange(blockRange);
    __SUI_RICH.command('editor', 'paragraph', 'h2');
    await wait(250);
    if (!richRoot().querySelector('.sui-rich-editor > h2')) throw new Error('heading command: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    var alignedHeading = richRoot().querySelector('.sui-rich-editor > h2');
    var alignedRange = document.createRange(); alignedRange.selectNodeContents(alignedHeading);
    var alignedSel = window.getSelection(); alignedSel.removeAllRanges(); alignedSel.addRange(alignedRange);
    __SUI_RICH.command('editor', 'alignCenter');
    await wait(250);
    if ((richRoot().querySelector('.sui-rich-editor > h2').style.textAlign || '') !== 'center') throw new Error('align command: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    var heading = richRoot().querySelector('.sui-rich-editor > h2');
    var headingRange = document.createRange(); headingRange.selectNodeContents(heading);
    var headingSel = window.getSelection(); headingSel.removeAllRanges(); headingSel.addRange(headingRange);
    __SUI_RICH.command('editor', 'bulletList');
    await wait(250);
    if (!richRoot().querySelector('.sui-rich-editor ul li')) throw new Error('list command: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    if (richRoot().querySelector('.sui-rich-props')) throw new Error('properties panel should not exist');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Table', marks: [] }] }] });
    await wait(200);
    __SUI_RICH.command('editor', 'table');
    await wait(250);
    var tableCell = richRoot().querySelector('.sui-rich-editor td, .sui-rich-editor th');
    var tableRange = document.createRange(); tableRange.selectNodeContents(tableCell);
    var tableSel = window.getSelection(); tableSel.removeAllRanges(); tableSel.addRange(tableRange);
    await wait(200);
    var menuButton = richRoot().querySelector('.sui-rich-menu-btn');
    var menuPanel = menuButton.parentNode.querySelector('.sui-rich-menu-panel');
    menuButton.click();
    if (menuPanel.hidden) throw new Error('menu did not open');
    document.body.click();
    if (!menuPanel.hidden) throw new Error('menu did not close outside');
    var tableBar = richRoot().querySelector('.sui-rich-tablebar');
    if (!tableBar || tableBar.hidden) throw new Error('table toolbar missing');
    var rowsBefore = richRoot().querySelectorAll('.sui-rich-editor tr').length;
    tableBar.querySelector('[data-rich-cmd="tableRowBelow"]').click();
    await wait(250);
    if (richRoot().querySelectorAll('.sui-rich-editor tr').length !== rowsBefore + 1) throw new Error('table row command: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);

    var tableSelect = richRoot().querySelector('.sui-rich-tablebar [data-rich-cmd="tableSelect"]');
    if (!tableSelect) throw new Error('table select command missing');
    tableSelect.click();
    await wait(100);
    if (!richRoot().querySelector('.sui-rich-editor table.sui-rich-table-selected')) throw new Error('whole table selection missing');
    richRoot().querySelector('.sui-rich-tablebar [data-rich-cmd="tableAlignCenter"]').click();
    await wait(250);
    if ((richRoot().querySelector('.sui-rich-editor table').style.marginLeft || '') !== 'auto') throw new Error('table alignment: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Format test', marks: [{ type: 'bold' }] }] }] });
    await wait(100);
    var docx = __SUI_RICH.exportBytes('editor', 'docx');
    if (docx[1] !== null || docx[0][0] !== 80 || docx[0][1] !== 75) throw new Error('docx export: ' + docx[1]);
    var odt = __SUI_RICH.exportBytes('editor', 'odt');
    if (odt[1] !== null || odt[0][0] !== 80 || odt[0][1] !== 75) throw new Error('odt export: ' + odt[1]);
    var pdf = __SUI_RICH.exportBytes('editor', 'pdf');
    if (pdf[1] !== null || pdf[0][0] !== 37 || pdf[0][1] !== 80) throw new Error('pdf export: ' + pdf[1]);
    var rtf = __SUI_RICH.exportBytes('editor', 'rtf');
    if (rtf[1] !== null || rtf[0][0] !== 123) throw new Error('rtf export: ' + rtf[1]);
    var docxImport = await __SUI_RICH.importBytes('editor', docx[0], 'x.docx');
    if (docxImport[1] !== null || __SUI_RICH.exportFormat('editor', 'markdown')[0].indexOf('Format test') < 0) throw new Error('docx import: ' + docxImport[1]);
    var odtImport = await __SUI_RICH.importBytes('editor', odt[0], 'x.odt');
    if (odtImport[1] !== null || __SUI_RICH.exportFormat('editor', 'markdown')[0].indexOf('Format test') < 0) throw new Error('odt import: ' + odtImport[1]);
    var rtfImport = await __SUI_RICH.importBytes('editor', rtf[0], 'x.rtf');
    if (rtfImport[1] !== null || __SUI_RICH.exportFormat('editor', 'markdown')[0].indexOf('Format test') < 0) throw new Error('rtf import: ' + rtfImport[1]);

    var png = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9Zl1sAAAAASUVORK5CYII=';
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'image', src: png, alt: 'dot', width: '20px' }] }] });
    await wait(100);
    for (var formatIndex = 0; formatIndex < 3; formatIndex++) {
      var fmt = ['docx', 'odt', 'rtf'][formatIndex];
      var exported = __SUI_RICH.exportBytes('editor', fmt);
      var imported = await __SUI_RICH.importBytes('editor', exported[0], 'image.' + fmt);
      if (imported[1] !== null || richRoot().querySelectorAll('.sui-rich-editor img').length < 1) throw new Error('image roundtrip ' + fmt + ': ' + imported[1]);
    }

    var rtfTable = '{\\rtf1\\ansi{\\fonttbl{\\f0 Liberation Serif;}{\\f1 Arial;}}{\\header\\pard{\\pict\\picwgoal1500\\pichgoal1500\\pngblip 89504e470d0a1a0a0000000d4948445200000001000000010804000000b51c0c020000000b4944415478da63fcff1f0002eb01f58f59975b0000000049454e44ae426082}\\par}' +
      '\\trowd\\trqc\\cellx1000\\cellx3000\\cellx6000\\pard\\intbl{\\f0 A\\u-3928\\\'3f}\\cell B\\cell C\\cell\\row' +
      '\\trowd\\cellx1000\\cellx6000\\pard\\intbl{\\b D}\\cell E\\cell\\row\\pard\\par}';
    var tableImport = await __SUI_RICH.importBytes('editor', new TextEncoder().encode(rtfTable), 'table.rtf');
    if (tableImport[1] !== null) throw new Error('rtf table import: ' + tableImport[1]);
    var tableDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var tables = tableDoc.children.filter(function (c) { return c.type === 'table'; });
    if (tables.length !== 1) throw new Error('rtf table count: ' + tableDoc.children.length);
    var tableRows = tables[0].rows;
    if (tableRows.length !== 2) throw new Error('rtf table rows: ' + tableRows.length);
    if (tableRows[0].cells.length !== 3 || tableRows[1].cells.length !== 2) throw new Error('rtf table cells: ' + tableRows[0].cells.length + ',' + tableRows[1].cells.length);
    if (tableRows[0].cells[1].children[0].text !== 'B') throw new Error('rtf cell text: ' + JSON.stringify(tableRows[0].cells[1].children));
    if (tableRows[1].cells[1].colspan !== 2) throw new Error('rtf cell colspan: ' + tableRows[1].cells[1].colspan);
    if (!tableRows[1].cells[0].children[0].marks.some(function (m) { return m.type === 'bold'; })) throw new Error('rtf cell marks: ' + JSON.stringify(tableRows[1].cells[0].children));
    if (tableDoc.children[0].type !== 'image' || tableDoc.children[0].width !== '100px') throw new Error('rtf header image: ' + JSON.stringify(tableDoc.children[0]));
var firstCellText = tableRows[0].cells[0].children.map(function (n) { return n.text || ''; }).join('');
    if (firstCellText.indexOf('\u2610') < 0) throw new Error('rtf checkbox glyph: ' + firstCellText);
    if (!tableRows[0].cells[0].children.some(function (n) { return (n.marks || []).some(function (m) { return m.type === 'font' && /serif/i.test(m.value); }); })) throw new Error('rtf font mark: ' + JSON.stringify(tableRows[0].cells[0].children));
    if (richRoot().querySelectorAll('.sui-rich-editor table tr').length !== 2) throw new Error('rtf table dom');
    if (richRoot().querySelectorAll('.sui-rich-editor img').length !== 1) throw new Error('rtf header image dom: ' + richRoot().querySelectorAll('.sui-rich-editor img').length);
    var pdfTable = __SUI_RICH.exportBytes('editor', 'pdf');
    if (pdfTable[1] !== null || pdfTable[0][0] !== 37) throw new Error('pdf export: ' + pdfTable[1]);
    var pdfText = '';
    for (var pb = 0; pb < pdfTable[0].length; pb += 8192) {
      var pdfChunk = [];
      for (var pq = pb; pq < Math.min(pdfTable[0].length, pb + 8192); pq++) pdfChunk.push(pdfTable[0][pq]);
      pdfText += String.fromCharCode.apply(null, pdfChunk);
    }
    if (pdfText.indexOf('/Subtype /Image') < 0) throw new Error('pdf image missing');
    if (pdfText.indexOf('/BaseFont /Times-Roman') < 0) throw new Error('pdf serif font missing');
    if (pdfText.indexOf('/Type /Page') < 0) throw new Error('pdf page missing');
    var cellBackground = getComputedStyle(richRoot().querySelector('.sui-rich-editor td')).backgroundColor;
    if (cellBackground !== 'rgba(0, 0, 0, 0)' && cellBackground !== 'transparent') throw new Error('rtf cell background: ' + cellBackground);
    var pictureBackground = getComputedStyle(richRoot().querySelector('.sui-rich-editor img')).backgroundColor;
    if (pictureBackground !== 'rgba(0, 0, 0, 0)' && pictureBackground !== 'transparent') throw new Error('rtf picture background: ' + pictureBackground);

    var tableRtfOut = __SUI_RICH.exportBytes('editor', 'rtf');
    if (tableRtfOut[1] !== null || tableRtfOut[0][0] !== 123) throw new Error('rtf table export: ' + tableRtfOut[1]);
    var tableBack = await __SUI_RICH.importBytes('editor', tableRtfOut[0], 'back.rtf');
    if (tableBack[1] !== null) throw new Error('rtf table re-import: ' + tableBack[1]);
    var backDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var backTables = backDoc.children.filter(function (c) { return c.type === 'table'; });
    if (backTables.length !== 1 || backTables[0].rows.length !== 2) throw new Error('rtf table re-import rows');
    if (richRoot().querySelector('.sui-rich-editor').innerText.indexOf('D') < 0) throw new Error('rtf table re-import text');
    var toLatin = function (bytes) {
      var out = '';
      for (var bi = 0; bi < bytes.length; bi += 8192) {
        var chunk = [];
        for (var bj = bi; bj < Math.min(bytes.length, bi + 8192); bj++) chunk.push(bytes[bj]);
        out += String.fromCharCode.apply(null, chunk);
      }
      return out;
    };
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'paragraph', align: 'center', spaceBefore: '6pt', indent: '24px', children: [
        { type: 'text', text: 'Formatting', marks: [{ type: 'bold' }, { type: 'color', value: 'rgb(31, 73, 125)' }, { type: 'size', value: '13pt' }, { type: 'font', value: 'Georgia, serif' }, { type: 'background', value: '#FFFF00' }] }
      ] },
      { type: 'page_break' },
      { type: 'paragraph', children: [{ type: 'text', text: 'After break', marks: [] }] }
    ] });
    await wait(100);
    var formatJson = __SUI_RICH.get('editor')[0];
    if (formatJson.indexOf('"page_break"') < 0) throw new Error('page break not in the document model');
    if (richRoot().querySelectorAll('.sui-rich-editor .sui-rich-pagebreak').length !== 1) throw new Error('page break not rendered');
    var formatDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (formatDocx.indexOf('<w:br w:type="page"/>') < 0) throw new Error('docx page break missing');
    if (formatDocx.indexOf('<w:jc w:val="center"/>') < 0) throw new Error('docx alignment missing');
    if (formatDocx.indexOf('<w:spacing w:before="120"') < 0) throw new Error('docx spacing missing');
    if (formatDocx.indexOf('<w:ind w:left="360"/>') < 0) throw new Error('docx indent missing');
    if (formatDocx.indexOf('<w:rFonts w:ascii="Georgia"') < 0) throw new Error('docx font missing');
    if (formatDocx.indexOf('<w:color w:val="1F497D"/>') < 0) throw new Error('docx color missing');
    if (formatDocx.indexOf('<w:sz w:val="26"/>') < 0) throw new Error('docx size missing');
    if (formatDocx.indexOf('w:fill="FFFF00"') < 0) throw new Error('docx highlight missing');
    var formatRtf = toLatin(__SUI_RICH.exportBytes('editor', 'rtf')[0]);
    if (formatRtf.indexOf('\\page') < 0) throw new Error('rtf page break missing');
    if (formatRtf.indexOf('\\qc') < 0 || formatRtf.indexOf('\\sb120') < 0 || formatRtf.indexOf('\\li360') < 0) throw new Error('rtf paragraph formatting missing');
    if (formatRtf.indexOf('\\fs26') < 0 || formatRtf.indexOf('Georgia') < 0) throw new Error('rtf run formatting missing');
    if (formatRtf.indexOf('\\colortbl') < 0 || formatRtf.indexOf('\\highlight') < 0) throw new Error('rtf colors missing');
    var formatPdf = toLatin(__SUI_RICH.exportBytes('editor', 'pdf')[0]);
    if (formatPdf.indexOf('/Count 2') < 0) throw new Error('pdf page break missing');
    var chromeHeader = richRoot().querySelector('.sui-rich-pageheader');
    if (!chromeHeader || chromeHeader.textContent.indexOf('Report') < 0) throw new Error('editor header strip missing');
    if (formatPdf.indexOf('Report 1/2') < 0) throw new Error('pdf header missing');
    if (formatPdf.indexOf('Footer 2') < 0) throw new Error('pdf footer page number missing');
    if (formatDocx.indexOf('word/header1.xml') < 0 || formatDocx.indexOf('Report ') < 0) throw new Error('docx header missing');
    if (formatDocx.indexOf('w:fldSimple w:instr=" PAGE "') < 0) throw new Error('docx header page field missing');
    if (formatDocx.indexOf('word/footer1.xml') < 0) throw new Error('docx footer missing');
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'bullet_list', children: [
        { type: 'list_item', children: [
          { type: 'paragraph', children: [{ type: 'text', text: 'One', marks: [] }] },
          { type: 'ordered_list', children: [
            { type: 'list_item', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Nested', marks: [] }] }] }
          ] }
        ] },
        { type: 'list_item', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Two', marks: [] }] }] }
      ] }
    ] });
    await wait(120);
    var listDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (listDocx.indexOf('word/numbering.xml') < 0) throw new Error('docx numbering part missing');
    if (listDocx.indexOf('<w:numPr>') < 0) throw new Error('docx list numbering missing');
    if (listDocx.indexOf('<w:ilvl w:val="1"/>') < 0) throw new Error('docx nested list level missing');
    var listOdt = toLatin(__SUI_RICH.exportBytes('editor', 'odt')[0]);
    if (listOdt.indexOf('text:list-style') < 0 || listOdt.indexOf('text:style-name="LBullets"') < 0) throw new Error('odt list style missing');
    var listPdf = toLatin(__SUI_RICH.exportBytes('editor', 'pdf')[0]);
    if (listPdf.indexOf('One') < 0 || listPdf.indexOf('Nested') < 0) throw new Error('pdf list text missing');
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'paragraph', children: [{ type: 'text', text: 'Page one', marks: [] }] },
      { type: 'page_break' },
      { type: 'paragraph', children: [{ type: 'text', text: 'Page two', marks: [] }] }
    ] });
    await wait(100);
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(700);
    if (!richRoot().classList.contains('sui-rich-paginated')) throw new Error('paginated view not enabled');
    var sheets = richRoot().querySelectorAll('.sui-rich-page-sheet');
    if (sheets.length < 2) throw new Error('paginated view sheets: ' + sheets.length);
    if (sheets[0].innerText.indexOf('Page one') < 0 || sheets[1].innerText.indexOf('Page two') < 0) throw new Error('paginated view content');
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(120);
    if (richRoot().classList.contains('sui-rich-paginated')) throw new Error('paginated view not disabled');
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'paragraph', children: [{ type: 'text', text: 'RT', marks: [{ type: 'bold' }, { type: 'color', value: '#1F497D' }, { type: 'size', value: '14pt' }, { type: 'font', value: 'Georgia' }] }] },
      { type: 'bullet_list', children: [{ type: 'list_item', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Item', marks: [] }] }] }] }
    ] });
    await wait(100);
    var rtDocxImport = await __SUI_RICH.importBytes('editor', __SUI_RICH.exportBytes('editor', 'docx')[0], 'rt.docx');
    if (rtDocxImport[1] !== null) throw new Error('docx roundtrip import: ' + rtDocxImport[1]);
    var rt1 = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (!rt1.header || rt1.header.indexOf('Report') < 0) throw new Error('docx header roundtrip');
    if (!rt1.children.some(function (c) { return c.type === 'bullet_list'; })) throw new Error('docx list roundtrip');
    var rtMarks = (rt1.children[0].children[0].marks || []).map(function (m) { return m.type; });
    if (rtMarks.indexOf('color') < 0 || rtMarks.indexOf('font') < 0 || rtMarks.indexOf('size') < 0) throw new Error('docx marks roundtrip: ' + rtMarks.join(','));
    var rtOdtImport = await __SUI_RICH.importBytes('editor', __SUI_RICH.exportBytes('editor', 'odt')[0], 'rt.odt');
    if (rtOdtImport[1] !== null) throw new Error('odt roundtrip import: ' + rtOdtImport[1]);
    var rt2 = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (!rt2.header || rt2.header.indexOf('Report') < 0) throw new Error('odt header roundtrip');
    if (!rt2.children.some(function (c) { return c.type === 'bullet_list'; })) throw new Error('odt list roundtrip');
    var editorEl = richRoot().querySelector('.sui-rich-editor');
    if (editorEl.getAttribute('role') !== 'textbox' || editorEl.getAttribute('spellcheck') !== 'true') throw new Error('editor accessibility attributes');
    if (richRoot().querySelector('.sui-rich-toolbar').getAttribute('role') !== 'toolbar') throw new Error('toolbar role');
    var headerRows = [{ type: 'table_row', cells: [
      { type: 'table_cell', header: true, children: [{ type: 'text', text: 'HDRCELL', marks: [] }] },
      { type: 'table_cell', header: true, children: [{ type: 'text', text: 'HDRCELL', marks: [] }] }
    ] }];
    for (var hrow = 1; hrow <= 50; hrow++) headerRows.push({ type: 'table_row', height: '40px', cells: [
      { type: 'table_cell', children: [{ type: 'text', text: 'r' + hrow, marks: [] }] },
      { type: 'table_cell', children: [{ type: 'text', text: 'v' + hrow, marks: [] }] }
    ] });
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'table', header: true, rows: headerRows }] });
    await wait(120);
    var repeatPdf = toLatin(__SUI_RICH.exportBytes('editor', 'pdf')[0]);
    if ((repeatPdf.match(/HDRCELL/g) || []).length < 4) throw new Error('pdf table header not repeated');
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'table', rows: [
        { type: 'table_row', cells: [
          { type: 'table_cell', children: [{ type: 'text', text: 'A', marks: [] }] },
          { type: 'table_cell', children: [{ type: 'text', text: 'B', marks: [] }] }
        ] }
      ] }
    ] });
    await wait(120);
    var cellEl = richRoot().querySelector('.sui-rich-editor td');
    var cellEditor = richRoot().querySelector('.sui-rich-editor');
    cellEditor.focus();
    var cellRange = document.createRange();
    cellRange.selectNodeContents(cellEl);
    cellRange.collapse(true);
    var cellSel = window.getSelection();
    cellSel.removeAllRanges();
    cellSel.addRange(cellRange);
    await wait(100);
    var propsButton = richRoot().querySelector('[data-rich-cmd="tableProps"]');
    if (!propsButton) throw new Error('table properties button missing');
    var propsCmd = __SUI_RICH.command('editor', 'tableProps');
    await wait(150);
    var propsInputs = document.querySelectorAll('.sui-rich-tableprops input');
    if (propsInputs.length < 3) throw new Error('table properties panel missing: ' + JSON.stringify(propsCmd));
    propsInputs[0].value = '10'; propsInputs[0].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(100);
    propsInputs[1].value = '2'; propsInputs[1].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(100);
    propsInputs[2].value = '#1f497d'; propsInputs[2].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(120);
    var propsDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var propsTable = propsDoc.children.filter(function (c) { return c.type === 'table'; })[0];
    if (!propsTable || propsTable.padding !== '10px' || propsTable.borderWidth !== '2px' || !propsTable.border) throw new Error('table properties model: ' + JSON.stringify(propsTable));
    var propsDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (propsDocx.indexOf('<w:tblCellMar>') < 0) throw new Error('docx cell margins missing');
    if (propsDocx.indexOf('1F497D') < 0) throw new Error('docx table border color missing');
    if (propsDocx.indexOf('w:sz="16"') < 0) throw new Error('docx table border width missing');
    var sideBoxes = document.querySelectorAll('.sui-rich-tableprops-sides input');
    if (!sideBoxes.length) throw new Error('table border side toggles missing');
    sideBoxes[0].checked = false;
    sideBoxes[0].dispatchEvent(new Event('change', { bubbles: true }));
    await wait(120);
    var sideDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var sideTable = sideDoc.children.filter(function (c) { return c.type === 'table'; })[0];
    if (!sideTable.borders || sideTable.borders.indexOf('top') >= 0 || sideTable.borders.indexOf('left') < 0) throw new Error('table border sides model: ' + JSON.stringify(sideTable.borders));
    var sideDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (sideDocx.indexOf('<w:top w:val="nil"') < 0) throw new Error('docx top border not disabled');
    if (sideDocx.indexOf('<w:left w:val="single"') < 0) throw new Error('docx left border missing');
    var sideBack = await __SUI_RICH.importBytes('editor', __SUI_RICH.exportBytes('editor', 'docx')[0], 'side.docx');
    if (sideBack[1] !== null) throw new Error('table props docx roundtrip: ' + sideBack[1]);
    await wait(100);
    var sideBackDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var sideBackTable = sideBackDoc.children.filter(function (c) { return c.type === 'table'; })[0];
    if (!sideBackTable || sideBackTable.padding !== '10px' || sideBackTable.borderWidth !== '2px' || !sideBackTable.border) throw new Error('table props roundtrip: ' + JSON.stringify(sideBackTable && { padding: sideBackTable.padding, borderWidth: sideBackTable.borderWidth, border: sideBackTable.border }));
    if (sideBackTable.borders && sideBackTable.borders.indexOf('top') >= 0) throw new Error('table side roundtrip');
    var htmlTable = '<table style="border:1px solid #ff0000"><tr><td style="border:1px solid #ff0000;padding:8px;background-color:#eeeeee">X</td><td style="border:1px solid #ff0000;padding:8px">Y</td></tr></table>';
    var htmlImport = __SUI_RICH.importFormat('editor', htmlTable, 'html');
    if (htmlImport[1] !== null) throw new Error('html table import: ' + htmlImport[1]);
    await wait(100);
    var htmlDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var htmlTableModel = htmlDoc.children.filter(function (c) { return c.type === 'table'; })[0];
    if (!htmlTableModel || !htmlTableModel.padding || !htmlTableModel.borderWidth || !htmlTableModel.border) throw new Error('html table styles: ' + JSON.stringify(htmlTableModel && { padding: htmlTableModel.padding, borderWidth: htmlTableModel.borderWidth, border: htmlTableModel.border }));
    if (!htmlTableModel.rows[0].cells[0].background) throw new Error('html table cell background');
    var nestedTable = { type: 'table', rows: [
      { type: 'table_row', cells: [
        { type: 'table_cell', children: [{ type: 'text', text: 'Inner1', marks: [] }] },
        { type: 'table_cell', children: [{ type: 'text', text: 'Inner2', marks: [] }] }
      ] }
    ] };
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'table', rows: [
        { type: 'table_row', cells: [
          { type: 'table_cell', blocks: [
            { type: 'paragraph', children: [{ type: 'text', text: 'Outer', marks: [] }] },
            nestedTable
          ] },
          { type: 'table_cell', children: [{ type: 'text', text: 'B', marks: [] }] }
        ] }
      ] }
    ] });
    await wait(150);
    if (!richRoot().querySelector('.sui-rich-editor td table')) throw new Error('nested table not rendered');
    var nestedModel = JSON.parse(__SUI_RICH.get('editor')[0]);
    var nestedCell = nestedModel.children[0].rows[0].cells[0];
    if (!nestedCell.blocks || !nestedCell.blocks.some(function (b) { return b.type === 'table'; })) throw new Error('nested table model');
    var nestedDocxBytes = __SUI_RICH.exportBytes('editor', 'docx')[0];
    var nestedDocx = toLatin(nestedDocxBytes);
    if ((nestedDocx.match(/<w:tbl>/g) || []).length < 2) throw new Error('nested table docx');
    var nestedBack = await __SUI_RICH.importBytes('editor', nestedDocxBytes, 'nested.docx');
    await wait(120);
    var nestedBackCell = JSON.parse(__SUI_RICH.get('editor')[0]).children[0].rows[0].cells[0];
    if (nestedBack[1] !== null || !nestedBackCell.blocks || !nestedBackCell.blocks.some(function (b) { return b.type === 'table'; })) throw new Error('nested table docx roundtrip');
    var nestedOdt = toLatin(__SUI_RICH.exportBytes('editor', 'odt')[0]);
    if ((nestedOdt.match(/<table:table>/g) || []).length < 2) throw new Error('nested table odt');
    var nestedPdf = toLatin(__SUI_RICH.exportBytes('editor', 'pdf')[0]);
    if (nestedPdf.indexOf('Inner1') < 0) throw new Error('nested table pdf');
    var nestedHtml = __SUI_RICH.importFormat('editor', '<table><tr><td>Outer<div><table><tr><td>Nested</td></tr></table></div></td></tr></table>', 'html');
    await wait(120);
    var nestedHtmlCell = JSON.parse(__SUI_RICH.get('editor')[0]).children[0].rows[0].cells[0];
    if (nestedHtml[1] !== null || !nestedHtmlCell.blocks || !nestedHtmlCell.blocks.some(function (b) { return b.type === 'table'; })) throw new Error('nested table html import');
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'table', rows: [
        { type: 'table_row', cells: [
          { type: 'table_cell', rowspan: 2, children: [{ type: 'text', text: 'Span', marks: [] }] },
          { type: 'table_cell', children: [{ type: 'text', text: 'B1', marks: [] }] }
        ] },
        { type: 'table_row', cells: [
          { type: 'table_cell', children: [{ type: 'text', text: 'C1', marks: [] }] }
        ] }
      ] }
    ] });
    await wait(120);
    var spanCell = richRoot().querySelector('.sui-rich-editor td[rowspan]');
    if (!spanCell) throw new Error('rowspan cell not rendered');
    var spanEditor = richRoot().querySelector('.sui-rich-editor');
    spanEditor.focus();
    var spanRange = document.createRange();
    spanRange.selectNodeContents(spanCell);
    spanRange.collapse(true);
    var spanSel = window.getSelection();
    spanSel.removeAllRanges();
    spanSel.addRange(spanRange);
    await wait(80);
    var splitResult = __SUI_RICH.command('editor', 'tableSplit');
    await wait(150);
    var splitDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var splitTable = splitDoc.children.filter(function (c) { return c.type === 'table'; })[0];
    if (!splitTable || Number(splitTable.rows[0].cells[0].rowspan) > 1) throw new Error('table split did not clear rowspan: ' + JSON.stringify(splitResult));
    if (splitTable.rows[1].cells.length < 2) throw new Error('table split did not add the row cell');
`+featureScriptTail)
}

func TestRichTextTagErrors(t *testing.T) {
	bad := []struct{ name, src, want string }{
		{"unknown attr", `state v = ""
fn run() -> null {}
view main { richtext "x" name: "e" value: v unsupported: true }`, "unsupported"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "richtext.sui")
			if d.HasErrors() {
				t.Fatalf("parse: %v", d.String())
			}
			d = Check(p)
			if !d.HasErrors() || !strings.Contains(d.String(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, d.String())
			}
		})
	}
}
