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
		richtext "Article" name: "editor" value: doc height: "320px" theme: "dark" menu: true toolbar: true contextToolbar: true statusbar: true onChange: changedValue onInput: changedValue toolbarItems: "undo,redo|styleSelect,bold,italic,underline,bulletList,orderedList|link,image,table"
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
