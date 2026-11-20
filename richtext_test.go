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
		richtext "Article" name: "editor" value: doc height: "320px" theme: "dark" header: "Report \{page\}/\{pages\}" footer: "Footer \{page\}" menu: true toolbar: true contextToolbar: true statusbar: true onChange: changedValue onInput: changedValue toolbarItems: "undo,redo|styleSelect,bold,italic,underline,textColor,highlight,bulletList,orderedList|link,image,table"
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
    var menuLabels = [].map.call(root.querySelectorAll('.sui-rich-menu-btn'), function (b) { return b.textContent; });
    if (menuLabels.join('|') !== 'File|Edit|Insert|Preferences') throw new Error('menu labels: ' + menuLabels.join('|'));
    if (root.querySelector('[data-rich-menu="format"]') || root.querySelector('[data-rich-menu="tools"]') || root.querySelector('[data-rich-menu="view"]')) throw new Error('redundant Format/Tools/View menu still present');
    var prefMenuBtn = [].filter.call(root.querySelectorAll('.sui-rich-menu-btn'), function (b) { return b.textContent === 'Preferences'; })[0];
    prefMenuBtn.click();
    await wait(120);
    var prefLabels = [].map.call(root.querySelectorAll('.sui-rich-menu-panel:not([hidden]) .sui-rich-menu-item'), function (b) { return b.textContent; });
    prefMenuBtn.click();
    if (prefLabels.join('|') !== 'Document settings|Page settings|Header and footer...|Icon style ▸|Line icons|Bold line icons|Duotone icons|Soft block icons') throw new Error('Preferences items: ' + prefLabels.join('|'));
    var insertMenuBtn = [].filter.call(root.querySelectorAll('.sui-rich-menu-btn'), function (b) { return b.textContent === 'Insert'; })[0];
    insertMenuBtn.click();
    await wait(120);
    var insertLabels = [].map.call(root.querySelectorAll('.sui-rich-menu-panel:not([hidden]) .sui-rich-menu-item'), function (b) { return b.textContent; });
    insertMenuBtn.click();
    if (insertLabels.indexOf('Unlink') >= 0 || insertLabels.indexOf('Image settings...') >= 0) throw new Error('redundant insert entries: ' + insertLabels.join(','));
    if (insertLabels.indexOf('Page break') < 0) throw new Error('Insert menu lost Page break');
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
    function typeText(text) {
      var typeEditor = richRoot().querySelector('.sui-rich-editor');
      typeEditor.dispatchEvent(new InputEvent('beforeinput', { inputType: 'insertText', data: text, bubbles: true, cancelable: true }));
      document.execCommand('insertText', false, text);
    }
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

    // Color controls show the document defaults (black text, white highlight)
    // and follow an explicit inline color at the caret.
    __SUI_RICH.setToolbar('editor', 'undo,redo|styleSelect,bold,italic,underline,textColor,highlight,codeBlock,bulletList,orderedList|link,image,table');
    await wait(150);
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Plain color', marks: [] }] }] });
    await wait(200);
    var colorEditor = richRoot().querySelector('.sui-rich-editor');
    var colorText = colorEditor.querySelector('p').firstChild;
    var colorRange = document.createRange(); colorRange.setStart(colorText, 2); colorRange.collapse(true);
    var colorSel = window.getSelection(); colorSel.removeAllRanges(); colorSel.addRange(colorRange); colorEditor.focus();
    document.dispatchEvent(new Event('selectionchange'));
    await wait(120);
    var colorInputs = [].slice.call(richRoot().querySelectorAll('.sui-rich-color'));
    var textColorInput = colorInputs.filter(function (input) { return input.getAttribute('data-rich-cmd') === 'textColor'; })[0];
    var highlightInput = colorInputs.filter(function (input) { return input.getAttribute('data-rich-cmd') === 'highlight'; })[0];
    if (!textColorInput || textColorInput.value !== '#000000') throw new Error('default text color swatch: ' + (textColorInput ? textColorInput.value : 'missing'));
    if (!highlightInput || highlightInput.value !== '#ffffff') throw new Error('default highlight swatch: ' + (highlightInput ? highlightInput.value : 'missing'));
    if (Math.round(textColorInput.getBoundingClientRect().width) < 34) throw new Error('color control is too small to click');

    // Bold/italic/underline must toggle off again.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Toggle format', marks: [] }] }] });
    await wait(200);
    var toggleEditor = richRoot().querySelector('.sui-rich-editor');
    var toggleText = toggleEditor.querySelector('p').firstChild;
    var toggleRange = document.createRange(); toggleRange.setStart(toggleText, 0); toggleRange.setEnd(toggleText, 6);
    var toggleSel = window.getSelection(); toggleSel.removeAllRanges(); toggleSel.addRange(toggleRange); toggleEditor.focus();
    __SUI_RICH.command('editor', 'bold');
    await wait(200);
    if (!richRoot().querySelector('.sui-rich-editor strong, .sui-rich-editor b, .sui-rich-editor [style*="font-weight"]')) throw new Error('bold did not apply: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    __SUI_RICH.command('editor', 'bold');
    await wait(200);
    if (richRoot().querySelector('.sui-rich-editor strong, .sui-rich-editor b, .sui-rich-editor [style*="font-weight"]')) throw new Error('bold did not toggle off: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);

    // With a collapsed caret a formatting button becomes a typing mode: it
    // stays highlighted and applies to the text typed afterwards until it is
    // pressed again.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Typing ', marks: [] }] }] });
    await wait(200);
    var typingEditor = richRoot().querySelector('.sui-rich-editor');
    var typingRange = document.createRange(); typingRange.selectNodeContents(typingEditor.querySelector('p')); typingRange.collapse(false);
    var typingSel = window.getSelection(); typingSel.removeAllRanges(); typingSel.addRange(typingRange); typingEditor.focus();
    var boldButton = richRoot().querySelector('.sui-rich-toolbar [data-rich-cmd="bold"]');
    __SUI_RICH.command('editor', 'bold');
    await wait(150);
    if (!boldButton.classList.contains('on') || !document.queryCommandState('bold')) throw new Error('collapsed bold did not turn on');
    typeText('Bold');
    await wait(200);
    var typedBold = richRoot().querySelector('.sui-rich-editor strong, .sui-rich-editor b, .sui-rich-editor [style*="font-weight"]');
    if (!typedBold || typedBold.textContent.indexOf('Bold') < 0) throw new Error('collapsed bold typing mode did not apply: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    __SUI_RICH.command('editor', 'bold');
    await wait(150);
    if (boldButton.classList.contains('on')) throw new Error('collapsed bold button stayed on');
    typeText(' plain');
    await wait(250);
    var boldRuns = [].map.call(richRoot().querySelectorAll('.sui-rich-editor strong, .sui-rich-editor b, .sui-rich-editor [style*="font-weight"]'), function (node) { return node.textContent; }).join('|');
    if (boldRuns.indexOf('plain') >= 0) throw new Error('collapsed bold off still typed bold: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);

    // A color chosen with a collapsed caret stays in the swatch and colors the
    // text typed afterwards; the same applies to the background highlight.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Color ', marks: [] }] }] });
    await wait(200);
    var typingColorEditor = richRoot().querySelector('.sui-rich-editor');
    var typingColorRange = document.createRange(); typingColorRange.selectNodeContents(typingColorEditor.querySelector('p')); typingColorRange.collapse(false);
    var typingColorSel = window.getSelection(); typingColorSel.removeAllRanges(); typingColorSel.addRange(typingColorRange); typingColorEditor.focus();
    var typingTextColor = richRoot().querySelector('.sui-rich-color[data-rich-cmd="textColor"]');
    var typingHighlight = richRoot().querySelector('.sui-rich-color[data-rich-cmd="highlight"]');
    typingTextColor.value = '#cc0000'; typingTextColor.dispatchEvent(new Event('input', { bubbles: true })); typingTextColor.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(200);
    if (typingTextColor.value !== '#cc0000') throw new Error('collapsed text color swatch: ' + typingTextColor.value);
    typeText('Red');
    await wait(200);
    var redSpan = [].slice.call(richRoot().querySelectorAll('.sui-rich-editor span')).filter(function (span) { return /rgb\(204, 0, 0\)/.test(getComputedStyle(span).color); })[0];
    if (!redSpan || redSpan.textContent.indexOf('Red') < 0) throw new Error('collapsed text color was not applied to typing');
    typingHighlight.value = '#ffee00'; typingHighlight.dispatchEvent(new Event('input', { bubbles: true })); typingHighlight.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(200);
    if (typingHighlight.value !== '#ffee00') throw new Error('collapsed highlight swatch: ' + typingHighlight.value);
    typeText('Gold');
    await wait(200);
    var goldSpan = [].slice.call(richRoot().querySelectorAll('.sui-rich-editor span')).filter(function (span) { return /rgb\(255, 238, 0\)/.test(getComputedStyle(span).backgroundColor); })[0];
    if (!goldSpan || goldSpan.textContent.indexOf('Gold') < 0) throw new Error('collapsed highlight was not applied to typing');

    // Applying a color to a selection still colors the selected text only.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Selected text', marks: [] }] }] });
    await wait(200);
    var selColorEditor = richRoot().querySelector('.sui-rich-editor');
    var selColorText = selColorEditor.querySelector('p').firstChild;
    var selColorRange = document.createRange(); selColorRange.setStart(selColorText, 0); selColorRange.setEnd(selColorText, 8);
    var selColorSel = window.getSelection(); selColorSel.removeAllRanges(); selColorSel.addRange(selColorRange); selColorEditor.focus();
    var selColorInput = richRoot().querySelector('.sui-rich-color[data-rich-cmd="textColor"]');
    selColorInput.value = '#0066cc'; selColorInput.dispatchEvent(new Event('input', { bubbles: true })); selColorInput.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(250);
    var selColorSpan = [].slice.call(richRoot().querySelectorAll('.sui-rich-editor span')).filter(function (span) { return /rgb\(0, 102, 204\)/.test(getComputedStyle(span).color); })[0];
    if (!selColorSpan || selColorSpan.textContent !== 'Selected') throw new Error('selection color not applied: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    var selBgInput = richRoot().querySelector('.sui-rich-color[data-rich-cmd="highlight"]');
    selBgInput.value = '#ffe066'; selBgInput.dispatchEvent(new Event('input', { bubbles: true })); selBgInput.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(250);
    var selBgSpan = [].slice.call(richRoot().querySelectorAll('.sui-rich-editor span')).filter(function (span) { return /rgb\(255, 224, 102\)/.test(getComputedStyle(span).backgroundColor); })[0];
    if (!selBgSpan || selBgSpan.textContent !== 'Selected') throw new Error('selection highlight not applied: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Palette color', marks: [] }] }] });
    await wait(200);
    var paletteEditor = richRoot().querySelector('.sui-rich-editor');
    var paletteText = paletteEditor.querySelector('p').firstChild;
    var paletteRange = document.createRange(); paletteRange.setStart(paletteText, 0); paletteRange.setEnd(paletteText, 7);
    var paletteSel = window.getSelection(); paletteSel.removeAllRanges(); paletteSel.addRange(paletteRange); paletteEditor.focus();
    var paletteButtons = [].slice.call(richRoot().querySelectorAll('.sui-rich-color-palette-btn'));
    if (paletteButtons.length < 2) throw new Error('color palette buttons missing');
    paletteButtons[0].click();
    await wait(150);
    var palettePanel = document.querySelector('.sui-rich-color-palette');
    if (!palettePanel) throw new Error('color palette did not open');
    var paletteSwatches = palettePanel.querySelectorAll('.sui-rich-color-swatch');
    if (paletteSwatches.length < 12) throw new Error('color palette presets missing');
    var redSwatch = [].slice.call(paletteSwatches).filter(function (sw) { return sw.getAttribute('title') === '#e03131'; })[0];
    if (!redSwatch) throw new Error('red palette swatch missing');
    redSwatch.click();
    await wait(250);
    if (document.querySelector('.sui-rich-color-palette')) throw new Error('color palette did not close');
    if (![].slice.call(richRoot().querySelectorAll('.sui-rich-editor span')).some(function (span) { return /rgb\(224, 49, 49\)/.test(getComputedStyle(span).color); })) throw new Error('palette color not applied: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);
    var paletteDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (!paletteDoc.children.some(function (n) { return n.marks && n.marks.some(function (m) { return m.type === 'color' && /rgb\(224, 49, 49\)/.test(m.value); }); })) throw new Error('palette color not stored');

    // Picking a color previews it immediately; clicking outside to close the
    // picker must keep the color and commit it to the model.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Preview color', marks: [] }] }] });
    await wait(200);
    var previewEditor = richRoot().querySelector('.sui-rich-editor');
    var previewText = previewEditor.querySelector('p').firstChild;
    var previewRange = document.createRange(); previewRange.setStart(previewText, 0); previewRange.setEnd(previewText, 7);
    var previewSel = window.getSelection(); previewSel.removeAllRanges(); previewSel.addRange(previewRange); previewEditor.focus();
    var previewInput = richRoot().querySelector('.sui-rich-color[data-rich-cmd="textColor"]');
    previewInput.showPicker = function () {};
    previewInput.closest('.sui-rich-color-control').click();
    await wait(80);
    previewInput.value = '#00aa00';
    previewInput.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(120);
    if (![].slice.call(richRoot().querySelectorAll('.sui-rich-editor span')).some(function (span) { return /rgb\(0, 170, 0\)/.test(getComputedStyle(span).color); })) throw new Error('color preview not applied while picking');
    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
    await wait(350);
    var previewDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (!previewDoc.children || !previewDoc.children.some(function (n) { return n.marks && n.marks.some(function (m) { return m.type === 'color' && /rgb\(0, 170, 0\)/.test(m.value); }); })) throw new Error('color lost when closing the picker: ' + JSON.stringify(previewDoc));
    if (![].slice.call(richRoot().querySelectorAll('.sui-rich-editor span')).some(function (span) { return /rgb\(0, 170, 0\)/.test(getComputedStyle(span).color); })) throw new Error('color not visible after closing the picker');

    // Four icon sets, persisted per editor and selectable from Preferences.
    var iconRoot = richRoot();
    var iconCommands = ['iconSetLine', 'iconSetBold', 'iconSetDuotone', 'iconSetBlocks'];
    for (var is = 0; is < iconCommands.length; is++) {
      __SUI_RICH.command('editor', iconCommands[is]);
      await wait(80);
      var expectedSet = iconCommands[is].replace('iconSet', '').toLowerCase();
      if (iconRoot.getAttribute('data-sui-rich-icons') !== expectedSet) throw new Error('icon set not applied: ' + iconCommands[is]);
    }
    __SUI_RICH.command('editor', 'iconSetLine');
    await wait(80);
    var iconSvg = iconRoot.querySelector('.sui-rich-toolbar [data-rich-cmd="bulletList"] svg');
    var iconSize = iconSvg ? parseFloat(getComputedStyle(iconSvg).width) : 0;
    if (iconSize < 20) throw new Error('toolbar icons are still too small: ' + iconSize);

    // Inline code and code blocks use a readable contrast on the white page.
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'paragraph', children: [{ type: 'text', text: 'inline ', marks: [] }, { type: 'text', text: 'code', marks: [{ type: 'code' }] }] },
      { type: 'code_block', text: 'line one\nline two\nline three' }
    ] });
    await wait(250);
    var codeEditor = richRoot().querySelector('.sui-rich-editor');
    var codeEl = codeEditor.querySelector('p code');
    var preEl = codeEditor.querySelector('pre');
    var editorColor = getComputedStyle(codeEditor).color;
    if (!codeEl || getComputedStyle(codeEl).color !== editorColor) throw new Error('inline code contrast: ' + (codeEl ? getComputedStyle(codeEl).color : 'missing') + ' vs ' + editorColor);
    if (!preEl || getComputedStyle(preEl).color !== editorColor) throw new Error('code block contrast: ' + (preEl ? getComputedStyle(preEl).color : 'missing') + ' vs ' + editorColor);

    // Turning a code block off releases only the line with the caret; the
    // other lines stay in code blocks.
    var preText = preEl.querySelector('code').firstChild;
    var preRange = document.createRange(); preRange.setStart(preText, 9); preRange.collapse(true);
    var preSel = window.getSelection(); preSel.removeAllRanges(); preSel.addRange(preRange); codeEditor.focus();
    __SUI_RICH.command('editor', 'codeBlock');
    await wait(250);
    var codeBlocks = [].map.call(richRoot().querySelectorAll('.sui-rich-editor > pre, .sui-rich-editor > p'), function (n) { return n.tagName + ':' + n.textContent; });
    var codeTail = codeBlocks.slice(-3).join('|');
    if (codeTail !== 'PRE:line one|P:line two|PRE:line three') throw new Error('code block line exit: ' + codeBlocks.join('|'));
    if (richRoot().querySelector('.sui-rich-toolbar [data-rich-cmd="codeBlock"]').classList.contains('on')) throw new Error('code block button stayed on after the line left the block');

    // Word styles inside a code block are stored in the model and survive a
    // re-render; indent edits the current code line, not the whole block.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'code_block', text: 'alpha beta\nline two' }] });
    await wait(250);
    var richCodeEditor = richRoot().querySelector('.sui-rich-editor');
    var richCodeText = richCodeEditor.querySelector('pre code').firstChild;
    var richCodeRange = document.createRange(); richCodeRange.setStart(richCodeText, 0); richCodeRange.setEnd(richCodeText, 5);
    var richCodeSel = window.getSelection(); richCodeSel.removeAllRanges(); richCodeSel.addRange(richCodeRange); richCodeEditor.focus();
    __SUI_RICH.command('editor', 'bold');
    await wait(250);
    var richCodeDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (!richCodeDoc.children || !richCodeDoc.children.some(function (n) { return n.marks && n.marks.some(function (m) { return m.type === 'bold'; }); })) throw new Error('code block word style not stored: ' + JSON.stringify(richCodeDoc));
    __SUI_RICH.set('editor', JSON.parse(__SUI_RICH.get('editor')[0]));
    await wait(250);
    if (!richRoot().querySelector('.sui-rich-editor pre strong, .sui-rich-editor pre b, .sui-rich-editor pre [style*="font-weight"]')) throw new Error('code block word style lost on re-render');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'code_block', text: 'one\ntwo' }] });
    await wait(250);
    var indentCodeEditor = richRoot().querySelector('.sui-rich-editor');
    var indentCodeText = indentCodeEditor.querySelector('pre code').firstChild;
    var indentCodeRange = document.createRange(); indentCodeRange.setStart(indentCodeText, 4); indentCodeRange.collapse(true);
    var indentCodeSel = window.getSelection(); indentCodeSel.removeAllRanges(); indentCodeSel.addRange(indentCodeRange); indentCodeEditor.focus();
    __SUI_RICH.command('editor', 'indent');
    await wait(200);
    var indentedCode = richRoot().querySelector('.sui-rich-editor pre code');
    if (!indentedCode || indentedCode.textContent !== 'one\n  two') throw new Error('code line indent: ' + JSON.stringify(indentedCode ? indentedCode.textContent : null));
    __SUI_RICH.command('editor', 'outdent');
    await wait(200);
    if (richRoot().querySelector('.sui-rich-editor pre code').textContent !== 'one\ntwo') throw new Error('code line outdent: ' + JSON.stringify(richRoot().querySelector('.sui-rich-editor pre code').textContent));
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(200);
    var tabPanel = document.querySelector('.sui-rich-paragraphsettings');
    var tabSizeSelect = tabPanel ? tabPanel.querySelector('[data-rich-paragraph-tabsize]') : null;
    if (!tabSizeSelect) throw new Error('code tab size control missing');
    tabSizeSelect.value = '4';
    tabPanel.querySelector('[data-rich-paragraph-apply]').click();
    await wait(200);
    var tabCodeEditor = richRoot().querySelector('.sui-rich-editor');
    var tabCodeText = tabCodeEditor.querySelector('pre code').firstChild;
    var tabCodeRange = document.createRange(); tabCodeRange.setStart(tabCodeText, 4); tabCodeRange.collapse(true);
    var tabCodeSel = window.getSelection(); tabCodeSel.removeAllRanges(); tabCodeSel.addRange(tabCodeRange); tabCodeEditor.focus();
    __SUI_RICH.command('editor', 'indent');
    await wait(200);
    var tabbedCode = richRoot().querySelector('.sui-rich-editor pre code');
    if (!tabbedCode || tabbedCode.textContent !== 'one\n    two') throw new Error('code tab size indent: ' + JSON.stringify(tabbedCode ? tabbedCode.textContent : null));
    var tabCodeDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (String(tabCodeDoc.tabSize) !== '4') throw new Error('code tab size not stored: ' + JSON.stringify(tabCodeDoc));
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'code_block', text: 'const one = 1;\nlet two = 2;\nreturn three;' }] });
    await wait(220);
    var lineNumEditor = richRoot().querySelector('.sui-rich-editor');
    var lineNumCode = lineNumEditor.querySelector('pre code');
    var lineNumRange = document.createRange(); lineNumRange.selectNodeContents(lineNumCode); lineNumRange.collapse(true);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(lineNumRange);
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(180);
    var lineNumPanel = document.querySelector('.sui-rich-paragraphsettings');
    var lineNumCheck = lineNumPanel.querySelector('[data-rich-paragraph-linenumbers]');
    if (!lineNumCheck) throw new Error('code line numbers control missing');
    lineNumCheck.checked = true;
    lineNumPanel.querySelector('[data-rich-paragraph-apply]').click();
    await wait(220);
    var numberedPre = richRoot().querySelector('.sui-rich-editor pre[data-line-numbers]');
    if (!numberedPre) throw new Error('code line numbers not rendered');
    var numberedLines = String(numberedPre.getAttribute('data-line-numbers')).split('\n');
    if (numberedLines.join('|') !== '1|2|3') throw new Error('code line numbers wrong: ' + numberedLines.join('|'));
    var numberedDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (!numberedDoc.lineNumbers) throw new Error('code line numbers not stored');
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(180);
    var syntaxPanel = document.querySelector('.sui-rich-paragraphsettings');
    var syntaxSelect = syntaxPanel.querySelector('[data-rich-paragraph-codelang]');
    if (!syntaxSelect) throw new Error('code language control missing');
    syntaxSelect.value = 'javascript';
    syntaxPanel.querySelector('[data-rich-paragraph-apply]').click();
    await wait(250);
    var syntaxPre = richRoot().querySelector('.sui-rich-editor pre[data-code-language="javascript"]');
    if (!syntaxPre) throw new Error('code language not applied');
    if (!syntaxPre.querySelector('.sui-code-keyword')) throw new Error('syntax highlighting missing');
    var syntaxDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (syntaxDoc.codeLanguage !== 'javascript') throw new Error('code language not stored');

    // Enter inside a list creates one list item per line.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'bullet_list', children: [{ type: 'list_item', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Left right', marks: [] }] }] }] }] });
    await wait(250);
    var listEditor = richRoot().querySelector('.sui-rich-editor');
    var listText = listEditor.querySelector('li p').firstChild;
    var listRange = document.createRange(); listRange.setStart(listText, 4); listRange.collapse(true);
    var listSel = window.getSelection(); listSel.removeAllRanges(); listSel.addRange(listRange); listEditor.focus();
    listEditor.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await wait(250);
    var listItems = [].map.call(richRoot().querySelectorAll('.sui-rich-editor li'), function (li) { return li.textContent; });
    if (listItems.join('|') !== 'Left| right') throw new Error('list Enter split: ' + listItems.join('|'));
    if (richRoot().querySelectorAll('.sui-rich-editor li').length !== 2) throw new Error('list Enter did not create a second item');

    // The image dialog offers a PC file picker and keeps the URL option.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Picture', marks: [] }] }] });
    await wait(200);
    __SUI_RICH.command('editor', 'image');
    await wait(200);
    var imagePrompt = document.querySelector('.sui-rich-prompt');
    var imageFile = imagePrompt ? imagePrompt.querySelector('input[type="file"]') : null;
    if (!imageFile || !/image\//.test(imageFile.accept)) throw new Error('image file picker missing');
    if (!imagePrompt.querySelector('input[type="url"]') && !imagePrompt.textContent.match(/Image URL/)) throw new Error('image URL field missing');
    var imageCancel = [].slice.call(imagePrompt.querySelectorAll('.sui-rich-prompt-actions button')).filter(function (b) { return b.textContent === 'Cancel'; })[0];
    if (imageCancel) imageCancel.click();

    // Table settings expose fit/px/% width and the column resize keeps the
    // other columns independent.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'table', rows: [
      { type: 'table_row', cells: [{ type: 'table_cell', children: [{ type: 'text', text: 'A', marks: [] }] }, { type: 'table_cell', children: [{ type: 'text', text: 'B', marks: [] }] }, { type: 'table_cell', children: [{ type: 'text', text: 'C', marks: [] }] }] },
      { type: 'table_row', cells: [{ type: 'table_cell', children: [{ type: 'text', text: '1', marks: [] }] }, { type: 'table_cell', children: [{ type: 'text', text: '2', marks: [] }] }, { type: 'table_cell', children: [{ type: 'text', text: '3', marks: [] }] }] }
    ] }] });
    await wait(300);
    var sizeTable = richRoot().querySelector('.sui-rich-editor table');
    var sizeRange = document.createRange(); sizeRange.selectNodeContents(sizeTable.rows[0].cells[0]); sizeRange.collapse(true);
    var sizeSel = window.getSelection(); sizeSel.removeAllRanges(); sizeSel.addRange(sizeRange);
    __SUI_RICH.command('editor', 'tableProps');
    await wait(200);
    var tablePanel = document.querySelector('.sui-rich-tableprops');
    var tableMode = tablePanel ? tablePanel.querySelector('select') : null;
    if (!tableMode) throw new Error('table width mode missing');
    var tableNumbers = tablePanel.querySelectorAll('input[type="number"]');
    var tableWidthInput = tableNumbers[tableNumbers.length - 1];
    tableMode.value = 'px'; tableMode.dispatchEvent(new Event('change', { bubbles: true }));
    tableWidthInput.value = '360'; tableWidthInput.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(200);
    if ((richRoot().querySelector('.sui-rich-editor table').style.width || '') !== '360px') throw new Error('table px width not applied: ' + richRoot().querySelector('.sui-rich-editor table').style.width);
    var resizeTable = richRoot().querySelector('.sui-rich-editor table');
    var resizeCells = resizeTable.rows[0].cells;
    var resizeBefore = [].map.call(resizeCells, function (cell) { return Math.round(cell.getBoundingClientRect().width); });
    var cellB = resizeCells[1].getBoundingClientRect();
    var dragX = cellB.right - 2, dragY = cellB.top + 5;
    resizeCells[1].dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: dragX, clientY: dragY, button: 0 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: dragX + 24, clientY: dragY }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: dragX + 24, clientY: dragY }));
    await wait(250);
    var resizeCols = [].map.call(richRoot().querySelectorAll('.sui-rich-editor table colgroup col'), function (col) { return Math.round(parseFloat(col.style.width)); });
    if (resizeCols.length !== 3) throw new Error('table columns not captured for resizing: ' + resizeCols.join(','));
    if (resizeCols[0] !== resizeBefore[0] || resizeCols[2] !== resizeBefore[2]) throw new Error('table columns not independent: ' + resizeBefore.join(',') + ' -> ' + resizeCols.join(','));
    if (Math.abs((resizeCols[1] - resizeBefore[1]) - 24) > 4) throw new Error('only the dragged column should grow: ' + resizeBefore.join(',') + ' -> ' + resizeCols.join(','));
    var tablePanelClose = tablePanel.querySelector('.sui-rich-tableprops-close');
    if (tablePanelClose) tablePanelClose.click();

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
    var tableBar = richRoot().querySelector('[data-rich-table-group]');
    if (!tableBar || tableBar.hidden) throw new Error('table toolbar missing');
    var rowsBefore = richRoot().querySelectorAll('.sui-rich-editor tr').length;
    tableBar.querySelector('[data-rich-cmd="tableRowBelow"]').click();
    await wait(250);
    if (richRoot().querySelectorAll('.sui-rich-editor tr').length !== rowsBefore + 1) throw new Error('table row command: ' + richRoot().querySelector('.sui-rich-editor').innerHTML);

    var tableSelect = richRoot().querySelector('[data-rich-table-group] [data-rich-cmd="tableSelect"]');
    if (!tableSelect) throw new Error('table select command missing');
    tableSelect.click();
    await wait(100);
    if (!richRoot().querySelector('.sui-rich-editor table.sui-rich-table-selected')) throw new Error('whole table selection missing');
    richRoot().querySelector('[data-rich-table-group] [data-rich-cmd="tableAlignCenter"]').click();
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
    if (pdfText.indexOf('/FontFile2') < 0 || pdfText.indexOf('/ToUnicode') < 0) throw new Error('pdf embedded font missing');
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

    // Rich code blocks, hyphenation and table column widths reach the file
    // exports, not only the editor.
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'code_block', text: 'styled code', children: [
        { type: 'text', text: 'styled', marks: [{ type: 'bold' }, { type: 'color', value: '#123456' }] },
        { type: 'text', text: ' code', marks: [] }
      ] },
      { type: 'paragraph', hyphens: 'auto', children: [{ type: 'text', text: 'Hyphenated paragraph', marks: [] }] },
      { type: 'paragraph', children: [{ type: 'text', text: 'Plain paragraph', marks: [] }] },
      { type: 'table', colWidths: ['100px', '300px'], rows: [
        { type: 'table_row', cells: [{ type: 'table_cell', children: [{ type: 'text', text: 'A', marks: [] }] }, { type: 'table_cell', children: [{ type: 'text', text: 'B', marks: [] }] }] }
      ] }
    ] });
    await wait(120);
    var exportDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (exportDocx.indexOf('Courier New') < 0 || exportDocx.indexOf('w:color w:val="123456"') < 0) throw new Error('docx rich code block missing');
    if (exportDocx.indexOf('\u00AD') < 0) throw new Error('docx soft hyphenation missing');
    if (exportDocx.indexOf('w:spacing') >= 0) throw new Error('hyphenation changed DOCX line spacing');
    if (exportDocx.indexOf('w:gridCol') < 0) throw new Error('docx table grid missing');
    var exportOdt = toLatin(__SUI_RICH.exportBytes('editor', 'odt')[0]);
    if (exportOdt.indexOf('font-weight:bold') < 0 || exportOdt.indexOf('color:rgb(18, 52, 86)') < 0) throw new Error('odt rich code block missing');
    if (exportOdt.indexOf('\u00AD') < 0) throw new Error('odt soft hyphenation missing');
    if (exportOdt.indexOf('style:column-width') < 0) throw new Error('odt table column widths missing');
    var exportRtf = toLatin(__SUI_RICH.exportBytes('editor', 'rtf')[0]);
    if (exportRtf.indexOf('Courier New;') < 0 || exportRtf.indexOf('\\-') < 0) throw new Error('rtf code/hyphenation export missing');
    var exportPdf = __SUI_RICH.exportBytes('editor', 'pdf');
    if (exportPdf[1] !== null || exportPdf[0].length < 1000) throw new Error('pdf export with rich code block failed');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'code_block', text: 'const x = 42;\nfunction hi() {\n  return "ok";\n}', codeLanguage: 'javascript', lineNumbers: true }] });
    await wait(150);
    var syntaxDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (syntaxDocx.indexOf('7C3AED') < 0 || syntaxDocx.indexOf('888888') < 0) throw new Error('DOCX syntax colors/line numbers missing');
    var syntaxOdt = toLatin(__SUI_RICH.exportBytes('editor', 'odt')[0]);
    if (syntaxOdt.indexOf('#7c3aed') < 0 || syntaxOdt.indexOf('#888888') < 0) throw new Error('ODT syntax colors/line numbers missing');
    var syntaxRtf = toLatin(__SUI_RICH.exportBytes('editor', 'rtf')[0]);
    if (syntaxRtf.indexOf('Courier New;') < 0 || syntaxRtf.indexOf('\\cf') < 0) throw new Error('RTF syntax colors missing');
    var syntaxPdf = __SUI_RICH.exportBytes('editor', 'pdf');
    if (syntaxPdf[1] !== null || syntaxPdf[0].length < 1000) throw new Error('PDF syntax export failed');

    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'paragraph', lineHeight: '1', spaceBefore: '5pt', spaceAfter: '5pt', hyphens: 'auto', children: [{ type: 'text', text: 'Turismo paragraph', marks: [] }] },
      { type: 'paragraph', lineHeight: '13.5pt', lineRule: 'atLeast', children: [{ type: 'text', text: 'At least spacing', marks: [] }] }
    ] });
    await wait(120);
    var spacingDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (spacingDocx.indexOf('w:lineRule="auto"') < 0 || spacingDocx.indexOf('w:lineRule="atLeast"') < 0) throw new Error('DOCX line spacing rule not preserved');
    if (spacingDocx.indexOf('w:lineRule="exact"') >= 0) throw new Error('DOCX line spacing changed to exact');

    var headerPng = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9Zl9sAAAAASUVORK5CYII=';
    __SUI_RICH.set('editor', { type: 'doc', header: { type: 'doc', children: [
      { type: 'paragraph', children: [{ type: 'text', text: 'Logo ', marks: [{ type: 'bold' }] }, { type: 'image', src: headerPng, alt: 'logo', width: '16' }] }
    ] }, children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Body', marks: [] }] }] });
    await wait(150);
    var headerDocx = toLatin(__SUI_RICH.exportBytes('editor', 'docx')[0]);
    if (headerDocx.indexOf('word/_rels/header1.xml.rels') < 0 || headerDocx.indexOf('word/media/') < 0 || headerDocx.indexOf('<a:blip') < 0) throw new Error('docx header image missing: hdr=' + (headerDocx.indexOf('word/header1.xml') >= 0) + ' rels=' + (headerDocx.indexOf('header1.xml.rels') >= 0) + ' media=' + (headerDocx.indexOf('word/media/') >= 0) + ' blip=' + (headerDocx.indexOf('<a:blip') >= 0));
    var headerOdt = toLatin(__SUI_RICH.exportBytes('editor', 'odt')[0]);
    if (headerOdt.indexOf('Pictures/') < 0 || headerOdt.indexOf('draw:frame') < 0) throw new Error('odt header image missing');
    var headerPdfBytes = __SUI_RICH.exportBytes('editor', 'pdf');
    var headerPdf = toLatin(headerPdfBytes[0]);
    if (headerPdfBytes[1] !== null || headerPdf.indexOf('/Subtype /Image') < 0) throw new Error('pdf header image missing');
    var headerRtf = toLatin(__SUI_RICH.exportBytes('editor', 'rtf')[0]);
    if (headerRtf.indexOf('{\\header') < 0 || headerRtf.indexOf('{\\pict') < 0) throw new Error('rtf header image missing');

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
    if (listPdf.indexOf('/FontFile2') < 0) throw new Error('pdf list fonts missing');
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
    __SUI_RICH.set('editor', { type: 'doc', header: 'Report {page}/{pages}', footer: 'Footer {page}', children: [
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
    if ((repeatPdf.match(/\/Count \d+/g) || [])[0] === '/Count 1') throw new Error('pdf table header page break missing');
    if (repeatPdf.indexOf('/FontFile2') < 0) throw new Error('pdf table header fonts missing');
    var repeatPaginated = __SUI_RICH.command('editor', 'paginatedView');
    await wait(900);
    var repeatSheets = richRoot().querySelectorAll('.sui-rich-page-sheet');
    if (repeatSheets.length < 2) throw new Error('paginated table header sheets: ' + repeatSheets.length);
    for (var rs = 0; rs < repeatSheets.length; rs++) {
      var rsBody = repeatSheets[rs].querySelector('.sui-rich-sheet-body');
      if (rsBody && rsBody.querySelector('tr') && rsBody.querySelector('tr').innerText.indexOf('HDRCELL') < 0) throw new Error('paginated table header not repeated');
      if (rsBody) {
        var rsLast = rsBody.lastElementChild;
        if (rsLast && rsLast.getBoundingClientRect().bottom > rsBody.getBoundingClientRect().bottom + 1.5) throw new Error('table content overflows the page content area');
      }
    }
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(120);
    // Large table with a vertical merge: every model row must appear exactly
    // once, rowspan groups must stay together, and no table may cross the
    // bottom margin in paginated view or print preview.
    var spanRows = [];
    for (var sr = 0; sr < 40; sr++) {
      if (sr === 3) spanRows.push({ type: 'table_row', cells: [
        { type: 'table_cell', rowspan: 3, children: [{ type: 'text', text: 'Merged', marks: [] }] },
        { type: 'table_cell', children: [{ type: 'text', text: 'A', marks: [] }] }
      ]});
      else if (sr === 4 || sr === 5) spanRows.push({ type: 'table_row', cells: [{ type: 'table_cell', children: [{ type: 'text', text: 'B', marks: [] }] }] });
      else spanRows.push({ type: 'table_row', cells: [
        { type: 'table_cell', children: [{ type: 'text', text: 'R' + sr, marks: [] }] },
        { type: 'table_cell', children: [{ type: 'text', text: 'V' + sr, marks: [] }] }
      ]});
    }
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'table', rows: spanRows }] });
    await wait(400);
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(1400);
    var spanSheets = richRoot().querySelectorAll('.sui-rich-page-sheet');
    var spanTotal = 0;
    for (var ss = 0; ss < spanSheets.length; ss++) {
      var spanBody = spanSheets[ss].querySelector('.sui-rich-sheet-body');
      var spanTable = spanSheets[ss].querySelector('table');
      if (spanBody && spanBody.scrollHeight > spanBody.clientHeight + 1) throw new Error('paginated table exceeds the bottom margin');
      if (spanTable) spanTotal += spanTable.querySelectorAll('tr').length;
    }
    if (spanTotal !== 40) throw new Error('table rows duplicated or lost: ' + spanTotal);
    __SUI_RICH.command('editor', 'printPreview');
    await wait(1500);
    var spanPreview = document.querySelectorAll('.sui-rich-preview-pages .sui-rich-page-sheet');
    if (spanPreview.length < 2) throw new Error('span table preview pages: ' + spanPreview.length);
    for (var sp = 0; sp < spanPreview.length; sp++) {
      var spBody = spanPreview[sp].querySelector('.sui-rich-sheet-body');
      if (spBody && spBody.scrollHeight > spBody.clientHeight + 1) throw new Error('preview table exceeds the bottom margin');
    }
    var spanClose = [].slice.call(document.querySelectorAll('.sui-rich-preview button')).filter(function (b) { return b.textContent === 'Close'; })[0];
    if (spanClose) spanClose.click();
    await wait(200);
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(200);
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
    if (nestedPdf.indexOf('/FontFile2') < 0 || nestedPdf.indexOf('/Type /Page') < 0) throw new Error('nested table pdf');
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

    // Block styles: Quote/Code must always be escapable. A rendered quote
    // wraps paragraphs, so the toolbar has to treat the quote container as the
    // styled block (toggle, style select, Escape/Backspace, clear formatting
    // and Enter on the last empty line).
    __SUI_RICH.setToolbar('editor', 'styleSelect,blockquote,codeBlock,clearFormat');
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Quote me', marks: [] }] }] });
    await wait(200);
    function bsEditor() { return richRoot().querySelector('.sui-rich-editor'); }
    function bsTopTag() { var e = bsEditor(); return e.firstElementChild ? e.firstElementChild.tagName.toLowerCase() : ''; }
    function bsCaretTop() {
      var e = bsEditor(), b = e.firstElementChild;
      if (!b) return;
      var t = b.tagName.toLowerCase() === 'blockquote' ? (b.firstElementChild || b) : b;
      var r = document.createRange(); r.setStart(t.firstChild || t, 0); r.collapse(true);
      var s = window.getSelection(); s.removeAllRanges(); s.addRange(r); e.focus();
    }
    function bsCaretEnd() {
      var e = bsEditor(), b = e.firstElementChild;
      if (!b) return;
      var t = b.tagName.toLowerCase() === 'blockquote' ? (b.lastElementChild || b) : b;
      var r = document.createRange(); r.selectNodeContents(t); r.collapse(false);
      var s = window.getSelection(); s.removeAllRanges(); s.addRange(r); e.focus();
    }
    function bsSelect() { return richRoot().querySelector('select[data-rich-cmd="paragraph"]'); }
    function bsBtn(cmd) { return richRoot().querySelector('.sui-rich-toolbar .sui-rich-btn[data-rich-cmd="' + cmd + '"]'); }
    if (!bsSelect()) throw new Error('block style select missing after setToolbar');

    bsCaretTop();
    __SUI_RICH.command('editor', 'blockquote');
    await wait(200);
    if (bsTopTag() !== 'blockquote') throw new Error('quote not applied: ' + bsTopTag());
    if (bsSelect().value !== 'blockquote') throw new Error('block style select not synced: ' + bsSelect().value);
    if (!bsBtn('blockquote') || !bsBtn('blockquote').classList.contains('on')) throw new Error('quote button not pressed');
    __SUI_RICH.command('editor', 'blockquote');
    await wait(200);
    if (bsTopTag() !== 'p') throw new Error('quote toggle did not exit: ' + bsTopTag());

    bsCaretTop();
    __SUI_RICH.command('editor', 'codeBlock');
    await wait(200);
    if (bsTopTag() !== 'pre') throw new Error('code block not applied: ' + bsTopTag());
    if (!bsBtn('codeBlock') || !bsBtn('codeBlock').classList.contains('on')) throw new Error('code button not pressed');
    bsCaretTop();
    bsEditor().dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    await wait(200);
    if (bsTopTag() !== 'p') throw new Error('Escape did not leave the code block: ' + bsTopTag());

    bsCaretTop();
    __SUI_RICH.command('editor', 'blockquote');
    await wait(200);
    bsCaretTop();
    bsEditor().dispatchEvent(new KeyboardEvent('keydown', { key: 'Backspace', bubbles: true, cancelable: true }));
    await wait(200);
    if (bsTopTag() !== 'p') throw new Error('Backspace did not leave the quote: ' + bsTopTag());

    bsCaretTop();
    __SUI_RICH.command('editor', 'blockquote');
    await wait(200);
    bsCaretTop();
    var bsNormal = bsSelect();
    bsNormal.value = '';
    bsNormal.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(200);
    if (bsTopTag() !== 'p') throw new Error('Normal style did not leave the quote: ' + bsTopTag());

    bsCaretTop();
    __SUI_RICH.command('editor', 'blockquote');
    await wait(200);
    bsCaretTop();
    __SUI_RICH.command('editor', 'clearFormat');
    await wait(200);
    if (bsTopTag() !== 'p') throw new Error('clearFormat did not leave the quote: ' + bsTopTag());

    bsCaretTop();
    __SUI_RICH.command('editor', 'blockquote');
    await wait(200);
    bsCaretEnd();
    bsEditor().dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await wait(200);
    if (bsTopTag() !== 'blockquote') throw new Error('Enter left the quote too early: ' + bsTopTag());
    bsEditor().dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await wait(200);
    if (bsTopTag() !== 'blockquote' || bsEditor().children.length < 2) {
      var bsQ = bsEditor().firstElementChild;
      var bsChain = [];
      var bsNode = window.getSelection() && window.getSelection().anchorNode;
      while (bsNode && bsChain.length < 6) { bsChain.push(bsNode.nodeType === 1 ? bsNode.tagName : '#text'); bsNode = bsNode.parentNode; }
      throw new Error('Enter on the last empty line did not exit the quote; kids=' + [].map.call(bsEditor().children, function (c) { return c.tagName; }) +
        ' quoteKids=' + (bsQ ? [].map.call(bsQ.children, function (c) { return c.tagName + ':' + (c.textContent || '') + ':' + (c.innerHTML || ''); }) : []) +
        ' chain=' + bsChain);
    }

    bsCaretTop();
    __SUI_RICH.command('editor', 'codeBlock');
    await wait(200);
    if (bsTopTag() !== 'pre') throw new Error('quote to code conversion failed: ' + bsTopTag());

    // Every text feature has a control: the default toolbar exposes inline
    // code and unlink, H4 is in the block style select, and pressed/disabled
    // states follow the caret.
    __SUI_RICH.setToolbar('editor', '');
    await wait(200);
    var defaultTokens = ['bold','italic','underline','strike','superscript','subscript','code','clearFormat','copyFormat','textColor','highlight','fontFamily','fontSize','alignLeft','alignCenter','alignRight','justify','indent','outdent','paragraphSettings','bulletList','orderedList','blockquote','codeBlock','link','unlink','image','table','symbol','hr','pageBreak','pageSettings','printPreview','print','find'];
    for (var dt = 0; dt < defaultTokens.length; dt++) {
      if (!richRoot().querySelector('.sui-rich-toolbar [data-rich-cmd="' + defaultTokens[dt] + '"]')) throw new Error('default toolbar missing control: ' + defaultTokens[dt]);
    }
    if (!bsSelect() || !bsSelect().querySelector('option[value="h4"]')) throw new Error('Heading 4 missing from the block style select');
    if (!richRoot().querySelector('.sui-rich-toolbar [data-rich-zoom="in"]') || !richRoot().querySelector('.sui-rich-toolbar [data-rich-zoom="out"]')) throw new Error('zoom controls missing from the toolbar');
    if (!richRoot().querySelector('.sui-rich-toolbar .sui-rich-zoom-label')) throw new Error('zoom scale label missing');
    if (richRoot().querySelectorAll('.sui-rich-status [data-rich-zoom]').length) throw new Error('zoom controls still in the status bar');
    var reqPagesBtn = richRoot().querySelector('.sui-rich-status [data-rich-pages]');
    if (!reqPagesBtn || reqPagesBtn.getBoundingClientRect().width < 60 || getComputedStyle(reqPagesBtn).whiteSpace !== 'nowrap') throw new Error('Pages button is too small or wraps');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Code me', marks: [] }] }] });
    await wait(200);
    var codePara = bsEditor().querySelector('p');
    var codeRange = document.createRange();
    codeRange.setStart(codePara.firstChild, 0); codeRange.setEnd(codePara.firstChild, 4);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(codeRange);
    __SUI_RICH.command('editor', 'code');
    await wait(200);
    if (!bsEditor().querySelector('code')) throw new Error('inline code not applied');
    var codeNode = bsEditor().querySelector('code');
    var codeSel = document.createRange(); codeSel.selectNodeContents(codeNode);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(codeSel);
    await wait(120);
    var codeBtn = bsBtn('code');
    if (!codeBtn || !codeBtn.classList.contains('on')) throw new Error('inline code button not pressed');
    __SUI_RICH.command('editor', 'code');
    await wait(200);
    if (bsEditor().querySelector('code')) throw new Error('inline code did not toggle off');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Example', marks: [{ type: 'link', href: 'https://example.com' }] }] }] });
    await wait(200);
    var linkNode = bsEditor().querySelector('a');
    if (!linkNode) throw new Error('link markup missing');
    var linkRange = document.createRange(); linkRange.selectNodeContents(linkNode);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(linkRange);
    await wait(120);
    var unlinkBtn = bsBtn('unlink');
    if (!unlinkBtn) throw new Error('unlink button missing');
    if (unlinkBtn.disabled) throw new Error('unlink disabled inside a link');
    unlinkBtn.click();
    await wait(200);
    if (bsEditor().querySelector('a')) throw new Error('unlink did not remove the link');

    // Paragraph settings expose explicit -/+ indent buttons.
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Indent panel', marks: [] }] }] });
    await wait(300);
    var indentPara = bsEditor().querySelector('p');
    var indentRange = document.createRange(); indentRange.selectNodeContents(indentPara); indentRange.collapse(true);
    var indentSel = window.getSelection(); indentSel.removeAllRanges(); indentSel.addRange(indentRange); bsEditor().focus();
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(250);
    var paraPanel = document.querySelector('.sui-rich-paragraphsettings');
    var indentInc = paraPanel && paraPanel.querySelector('[data-rich-paragraph-indent-inc]');
    var indentDec = paraPanel && paraPanel.querySelector('[data-rich-paragraph-indent-dec]');
    if (!indentInc || !indentDec) throw new Error('paragraph indent buttons missing');
    indentInc.click(); await wait(350);
    if (bsEditor().querySelector('p').style.paddingLeft !== '24px') throw new Error('paragraph indent increase failed: ' + bsEditor().querySelector('p').style.paddingLeft);
    indentInc.click(); await wait(350);
    indentDec.click(); await wait(350);
    if (bsEditor().querySelector('p').style.paddingLeft !== '24px') throw new Error('paragraph indent decrease failed: ' + bsEditor().querySelector('p').style.paddingLeft);
    var indentDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (parseFloat(indentDoc.children[0].indent || '0') !== 24) throw new Error('paragraph indent not stored: ' + indentDoc.children[0].indent);
    if (paraPanel) paraPanel.hidden = true;

    // Page settings (page format) and Document settings (general defaults).
    __SUI_RICH.setToolbar('editor', 'pageSettings,printPreview,print,find');
    await wait(200);
    if (bsBtn('imageSettings')) throw new Error('Image settings button must not be in the toolbar');
    var pageBtn = bsBtn('pageSettings');
    if (!pageBtn) throw new Error('page settings button missing');
    pageBtn.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
    pageBtn.click();
    await wait(200);
    var pagePanel = document.querySelector('.sui-rich-pagesettings');
    if (!pagePanel || pagePanel.hidden) throw new Error('page settings panel did not open');
    pagePanel.querySelector('[data-rich-page-number]').value = '1';
    pagePanel.querySelector('[data-rich-page-size]').value = 'letter';
    pagePanel.querySelector('[data-rich-page-orientation]').value = 'landscape';
    var orientationLabels = [].map.call(pagePanel.querySelector('[data-rich-page-orientation]').options, function (o) { return o.textContent; });
    if (orientationLabels.join(',') !== 'Vertical,Horizontal') throw new Error('page orientation labels: ' + orientationLabels.join(','));
    pagePanel.querySelector('[data-rich-page-margin="top"]').value = '20';
    pagePanel.querySelector('[data-rich-page-margin="right"]').value = '25';
    pagePanel.querySelector('[data-rich-page-margin="bottom"]').value = '18';
    pagePanel.querySelector('[data-rich-page-margin="left"]').value = '22';
    pagePanel.querySelector('[data-rich-page-apply]').click();
    await wait(400);
    var pageDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var pageOne = pageDoc.pageOverrides && pageDoc.pageOverrides['1'];
    if (!pageOne || pageOne.size !== 'letter' || pageOne.orientation !== 'landscape') throw new Error('page override not stored: ' + JSON.stringify(pageDoc.pageOverrides));
    if (!pageOne.margins || pageOne.margins.top !== 20 || pageOne.margins.right !== 25 || pageOne.margins.bottom !== 18 || pageOne.margins.left !== 22) throw new Error('page override margins not stored: ' + JSON.stringify(pageOne.margins));
    __SUI_RICH.command('editor', 'undo');
    await wait(250);
    var undoneOverrides = JSON.parse(__SUI_RICH.get('editor')[0]).pageOverrides;
    if (undoneOverrides && undoneOverrides['1']) throw new Error('undo did not restore the page settings');
    __SUI_RICH.command('editor', 'redo');
    await wait(250);
    var redoneOverride = JSON.parse(__SUI_RICH.get('editor')[0]).pageOverrides;
    if (!redoneOverride || !redoneOverride['1'] || redoneOverride['1'].size !== 'letter') throw new Error('redo did not restore the page settings');

    __SUI_RICH.command('editor', 'documentSettings');
    await wait(200);
    var docPanel = document.querySelector('.sui-rich-documentsettings');
    if (!docPanel || docPanel.hidden) throw new Error('document settings panel did not open');
    docPanel.querySelector('[data-rich-doc-lineheight]').value = '1.8';
    docPanel.querySelector('[data-rich-doc-header]').value = 'H {page}/{pages}';
    docPanel.querySelector('[data-rich-doc-footer]').value = 'F {page}';
    docPanel.querySelector('[data-rich-doc-apply]').click();
    await wait(400);
    pageDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (pageDoc.page.lineHeight !== 1.8 || bsEditor().style.lineHeight !== '1.8') throw new Error('document line height not applied: ' + JSON.stringify(pageDoc.page));
    if (pageDoc.header !== 'H {page}/{pages}' || pageDoc.footer !== 'F {page}') throw new Error('header/footer not stored');

    __SUI_RICH.command('editor', 'paginatedView');
    await wait(1000);
    var pageSheet = document.querySelector('.sui-rich-page-sheet');
    if (!pageSheet) throw new Error('paginated sheet missing');
    if (pageSheet.style.width !== '1056px' || pageSheet.style.minHeight !== '816px') throw new Error('Letter landscape sheet size wrong: ' + pageSheet.style.width + ' x ' + pageSheet.style.minHeight);
    if (pageSheet.style.padding !== '76px 94px 68px 83px') throw new Error('page margins not applied to the sheet: ' + pageSheet.style.padding);
    if ((pageSheet.querySelector('.sui-rich-sheet-header') || {}).textContent !== 'H 1/1') throw new Error('sheet header text missing');
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(200);

    var pagePdf = toLatin(__SUI_RICH.exportBytes('editor', 'pdf')[0]);
    if (pagePdf.indexOf('/MediaBox [0 0 792 612]') < 0) throw new Error('PDF did not use Letter landscape');
    var pageDocxBytes = __SUI_RICH.exportBytes('editor', 'docx')[0];
    var pageDocx = toLatin(pageDocxBytes);
    if (pageDocx.indexOf('<w:pgSz w:w="15840" w:h="12240" w:orient="landscape"/>') < 0) throw new Error('DOCX page size missing');
    if (pageDocx.indexOf('<w:pgMar w:top="1134" w:right="1417" w:bottom="1020" w:left="1247"') < 0) throw new Error('DOCX margins missing');
    var pageOdt = toLatin(__SUI_RICH.exportBytes('editor', 'odt')[0]);
    if (pageOdt.indexOf('fo:page-width="27.94cm"') < 0 || pageOdt.indexOf('fo:page-height="21.59cm"') < 0) throw new Error('ODT page size missing');
    if (pageOdt.indexOf('fo:margin-top="2cm"') < 0 || pageOdt.indexOf('fo:margin-right="2.5cm"') < 0) throw new Error('ODT margins missing');
    var pageImport = await __SUI_RICH.importBytes('editor', pageDocxBytes, 'x.docx');
    await wait(300);
    var importedPage = JSON.parse(__SUI_RICH.get('editor')[0]).page;
    if (pageImport[1] !== null || !importedPage || importedPage.size !== 'letter' || importedPage.orientation !== 'landscape' || importedPage.margins.left !== 22) throw new Error('DOCX page round-trip failed: ' + JSON.stringify(importedPage));

    // Fixed control height with scrolling, toolbar zoom, paragraph panel,
    // print preview, prompt overlays, format painter and File menu order.
    __SUI_RICH.setToolbar('editor', 'zoom,paragraphSettings,printPreview,link,bold,copyFormat,clearFormat');
    await wait(200);
    var manyRows = [];
    for (var mr = 0; mr < 80; mr++) manyRows.push({ type: 'paragraph', children: [{ type: 'text', text: 'Row ' + mr, marks: [] }] });
    __SUI_RICH.set('editor', { type: 'doc', children: manyRows });
    await wait(400);
    var reqEd = bsEditor();
    if (reqEd.clientHeight < 300 || reqEd.clientHeight > 340) throw new Error('editor height is not fixed: ' + reqEd.clientHeight);
    if (reqEd.scrollHeight <= reqEd.clientHeight) throw new Error('long content does not scroll inside the editor');

    // Header/footer mini editor: tabs, text toolbar, image insertion, token
    // controls, variable help and real page numbers.
    __SUI_RICH.command('editor', 'documentSettings');
    await wait(220);
    var chromePanel = document.querySelector('.sui-rich-documentsettings');
    if (chromePanel.querySelectorAll('.sui-rich-chrome-summary-row').length !== 0) throw new Error('header/footer summaries should not be in Document settings');
    if (chromePanel.querySelector('[data-rich-chrome-help]')) throw new Error('header/footer help should not be in Document settings');
    if (!chromePanel.querySelector('[data-rich-open-chrome]')) throw new Error('single header/footer editor button missing');
    if (chromePanel.querySelector('[data-rich-edit-chrome]')) throw new Error('per-section header/footer edit buttons still present');
    __SUI_RICH.command('editor', 'headerFooter');
    await wait(220);
    var quickChrome = document.querySelector('.sui-rich-chrome-editor');
    if (!quickChrome) throw new Error('Preferences header/footer command did not open the editor');
    var quickTabs = [].slice.call(quickChrome.querySelectorAll('.sui-rich-chrome-tab')).map(function (b) { return b.textContent; });
    if (quickTabs.join('|') !== 'Header|Footer') throw new Error('single header/footer editor tabs: ' + quickTabs.join('|'));
    [].slice.call(quickChrome.querySelectorAll('.sui-rich-chrome-actions button')).filter(function (b) { return b.textContent === 'Cancel'; })[0].click();
    await wait(150);
    chromePanel.querySelector('[data-rich-open-chrome]').click();
    await wait(300);
    var chromeOverlay = document.querySelector('.sui-rich-chrome-editor');
    if (!chromeOverlay) throw new Error('header/footer mini editor did not open');
    var chromeTabs = [].slice.call(chromeOverlay.querySelectorAll('.sui-rich-chrome-tab')).map(function (b) { return b.textContent; });
    if (chromeTabs.join('|') !== 'Header|Footer') throw new Error('header/footer tabs: ' + chromeTabs.join('|'));
    var chromeTools = [].slice.call(chromeOverlay.querySelectorAll('.sui-rich-chrome-toolbar .sui-rich-btn')).map(function (b) { return b.textContent; });
    for (var ct = 0; ct < ['B','I','U','Image','{page}','{pages}','Clear','?'].length; ct++) {
      if (chromeTools.indexOf(['B','I','U','Image','{page}','{pages}','Clear','?'][ct]) < 0) throw new Error('header/footer toolbar missing ' + ['B','I','U','Image','{page}','{pages}','Clear','?'][ct]);
    }
    var chromeVars = chromeOverlay.querySelector('.sui-rich-chrome-vars');
    var chromeVarValues = chromeVars ? [].map.call(chromeVars.options, function (o) { return o.value; }) : [];
    if (chromeVarValues.indexOf('{title}') < 0 || chromeVarValues.indexOf('{date}') < 0 || chromeVarValues.indexOf('{words}') < 0) throw new Error('header/footer variable list incomplete: ' + chromeVarValues.join(','));
    var chromeHelpBtn = [].slice.call(chromeOverlay.querySelectorAll('.sui-rich-chrome-help'))[0];
    chromeHelpBtn.click();
    await wait(250);
    var chromeHelp = document.querySelector('.sui-rich-chrome-help-overlay');
    if (!chromeHelp) throw new Error('header/footer variable help did not open');
    var chromeHelpText = chromeHelp.textContent;
    if (chromeHelpText.indexOf('{title}') < 0 || chromeHelpText.indexOf('{date}') < 0 || chromeHelpText.indexOf('{words}') < 0) throw new Error('header/footer help misses variables');
    [].slice.call(chromeHelp.querySelectorAll('button')).filter(function (b) { return b.textContent === 'Close'; })[0].click();
    await wait(150);
    var chromeHeaderSurface = chromeOverlay.querySelector('[data-rich-chrome-surface="header"]');
    var chromeFooterSurface = chromeOverlay.querySelector('[data-rich-chrome-surface="footer"]');
    var chromePng = 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxMjAiIGhlaWdodD0iNDAiPjxyZWN0IHdpZHRoPSIxMjAiIGhlaWdodD0iNDAiIGZpbGw9IiMzMzY2ZGQiLz48L3N2Zz4=';
    chromeHeaderSurface.innerHTML = '<p>H [page]/[pages]</p><p><b>Line 2</b> {title}</p><p><img src="' + chromePng + '" alt="logo"></p>';
    chromeFooterSurface.innerHTML = '<p>F [page] of [pages]</p>';
    [].slice.call(chromeOverlay.querySelectorAll('.sui-rich-chrome-actions button')).filter(function (b) { return b.textContent === 'Apply'; })[0].click();
    await wait(700);
    var chromeDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (!chromeDoc.header || !chromeDoc.header.children || chromeDoc.header.children.length !== 3) throw new Error('rich header model not stored');
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(2500);
    var chromeSheets = richRoot().querySelectorAll('.sui-rich-page-sheet');
    if (chromeSheets.length < 2) throw new Error('header/footer test needs multiple pages: ' + chromeSheets.length);
    for (var cs = 0; cs < chromeSheets.length; cs++) {
      var csh = chromeSheets[cs].querySelector('.sui-rich-sheet-header');
      var csf = chromeSheets[cs].querySelector('.sui-rich-sheet-footer');
      var csBody = chromeSheets[cs].querySelector('.sui-rich-sheet-body');
      if (!csh || csh.textContent.indexOf('H ' + (cs + 1) + '/' + chromeSheets.length) < 0) throw new Error('header page token wrong: ' + (csh ? csh.textContent : 'missing'));
      if (!csf || csf.textContent !== 'F ' + (cs + 1) + ' of ' + chromeSheets.length) throw new Error('footer page token wrong: ' + (csf ? csf.textContent : 'missing'));
      if (!csh.querySelector('img')) throw new Error('header image missing on page ' + (cs + 1));
      if (!csBody || csBody.getBoundingClientRect().top < csh.getBoundingClientRect().bottom - 1) throw new Error('header overlaps the page body');
    }
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(250);
    __SUI_RICH.command('editor', 'pageSettings');
    await wait(220);
    var chromePagePanel = document.querySelector('.sui-rich-pagesettings');
    if (chromePagePanel.querySelectorAll('.sui-rich-chrome-summary-row').length !== 0) throw new Error('header/footer summaries should not be in Page settings');
    if (chromePagePanel.querySelector('[data-rich-chrome-help]')) throw new Error('header/footer help should not be in Page settings');
    if (chromePagePanel.querySelector('[data-rich-edit-chrome]')) throw new Error('page settings per-section edit buttons still present');
    chromePagePanel.querySelector('[data-rich-open-chrome]').click();
    await wait(220);
    if (!document.querySelector('.sui-rich-chrome-editor')) throw new Error('page settings did not open the mini editor');
    [].slice.call(document.querySelectorAll('.sui-rich-chrome-editor .sui-rich-chrome-actions button')).filter(function (b) { return b.textContent === 'Cancel'; })[0].click();
    var chromePageClose = chromePagePanel.querySelector('[data-rich-page-close]');
    if (chromePageClose) chromePageClose.click();
    await wait(200);

    var reqZoomLabel = richRoot().querySelector('.sui-rich-toolbar .sui-rich-zoom-label');
    richRoot().querySelector('.sui-rich-toolbar [data-rich-zoom="in"]').click();
    await wait(100);
    if (!reqZoomLabel || reqZoomLabel.textContent !== '110%') throw new Error('zoom in did not update the label');
    if (bsEditor().style.zoom !== '1.1') throw new Error('zoom did not scale the editor view: ' + bsEditor().style.zoom);
    reqZoomLabel.click();
    await wait(100);
    if (reqZoomLabel.textContent !== '100%') throw new Error('zoom reset did not return to 100%');
    if (bsEditor().style.zoom !== '1') throw new Error('zoom reset did not restore the editor scale');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Para', marks: [] }] }] });
    await wait(300);
    var reqParaP = bsEditor().querySelector('p');
    var reqCaret = document.createRange(); reqCaret.setStart(reqParaP.firstChild, 0); reqCaret.collapse(true);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(reqCaret);
    richRoot().querySelector('.sui-rich-toolbar [data-rich-cmd="paragraphSettings"]').click();
    await wait(200);
    var reqPara = document.querySelector('.sui-rich-paragraphsettings');
    if (!reqPara || reqPara.hidden) throw new Error('paragraph settings panel did not open');
    reqPara.querySelector('[data-rich-paragraph-indent]').value = '24';
    reqPara.querySelector('[data-rich-paragraph-lineheight]').value = '1.4';
    reqPara.querySelector('[data-rich-paragraph-before]').value = '6';
    reqPara.querySelector('[data-rich-paragraph-after]').value = '8';
    reqPara.querySelector('[data-rich-paragraph-align]').value = 'right';
    var reqHyphen = reqPara.querySelector('[data-rich-paragraph-hyphens]');
    if (!reqHyphen) throw new Error('paragraph hyphenation control missing');
    reqHyphen.checked = true;
    reqPara.querySelector('[data-rich-paragraph-apply]').click();
    await wait(300);
    var reqParaDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (reqParaDoc.indent !== '24px' || reqParaDoc.lineHeight !== '1.4' || reqParaDoc.spaceBefore !== '6px' || reqParaDoc.spaceAfter !== '8px' || reqParaDoc.align !== 'right' || reqParaDoc.hyphens !== 'auto') throw new Error('paragraph settings not stored: ' + JSON.stringify(reqParaDoc));
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(180);
    var justifyPanel = document.querySelector('.sui-rich-paragraphsettings');
    if (!justifyPanel.querySelector('option[value="justifyHyphen"]')) throw new Error('justify + hyphenation option missing');
    justifyPanel.querySelector('[data-rich-paragraph-align]').value = 'justifyHyphen';
    justifyPanel.querySelector('[data-rich-paragraph-hyphens]').checked = false;
    justifyPanel.querySelector('[data-rich-paragraph-apply]').click();
    await wait(220);
    var justifyDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (justifyDoc.align !== 'justify' || justifyDoc.hyphens !== 'auto') throw new Error('justify + hyphenation not applied: ' + JSON.stringify(justifyDoc));
    if (getComputedStyle(bsEditor().querySelector('p')).hyphens !== 'auto') throw new Error('hyphenation is not active on the paragraph');
    if (!bsEditor().getAttribute('lang')) throw new Error('editor hyphenation language missing');
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(180);
    var plainJustifyPanel = document.querySelector('.sui-rich-paragraphsettings');
    plainJustifyPanel.querySelector('[data-rich-paragraph-align]').value = 'justify';
    plainJustifyPanel.querySelector('[data-rich-paragraph-hyphens]').checked = false;
    plainJustifyPanel.querySelector('[data-rich-paragraph-apply]').click();
    await wait(220);
    var plainJustifyDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (plainJustifyDoc.align !== 'justify' || plainJustifyDoc.hyphens) throw new Error('plain justify not applied: ' + JSON.stringify(plainJustifyDoc));

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'straordinariamente complicatissimo', marks: [] }] }] });
    await wait(200);
    var justifyFirstPara = bsEditor().querySelector('p');
    var justifyFirstRange = document.createRange(); justifyFirstRange.selectNodeContents(justifyFirstPara); justifyFirstRange.collapse(true);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(justifyFirstRange);
    __SUI_RICH.command('editor', 'justify');
    await wait(200);
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(180);
    var justifyFirstPanel = document.querySelector('.sui-rich-paragraphsettings');
    justifyFirstPanel.querySelector('[data-rich-paragraph-hyphens]').checked = true;
    justifyFirstPanel.querySelector('[data-rich-paragraph-apply]').click();
    await wait(250);
    var justifyFirstDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (justifyFirstDoc.align !== 'justify' || justifyFirstDoc.hyphens !== 'auto') throw new Error('hyphenation not applied after justify: ' + JSON.stringify(justifyFirstDoc));
    if (bsEditor().querySelector('p').textContent.indexOf('\u00AD') < 0) throw new Error('hyphenation after justify inserted no soft hyphens');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', lineHeight: '1', spaceBefore: '5pt', spaceAfter: '5pt', children: [{ type: 'text', text: 'Spacing preserved', marks: [] }] }] });
    await wait(200);
    var spacingPara = bsEditor().querySelector('p');
    var spacingRange = document.createRange(); spacingRange.selectNodeContents(spacingPara); spacingRange.collapse(true);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(spacingRange);
    __SUI_RICH.command('editor', 'paragraphSettings');
    await wait(180);
    var spacingPanel = document.querySelector('.sui-rich-paragraphsettings');
    spacingPanel.querySelector('[data-rich-paragraph-hyphens]').checked = true;
    spacingPanel.querySelector('[data-rich-paragraph-apply]').click();
    await wait(220);
    var spacingAfterHyphen = JSON.parse(__SUI_RICH.get('editor')[0]).children[0];
    if (spacingAfterHyphen.lineHeight !== '1' || spacingAfterHyphen.spaceBefore !== '5pt' || spacingAfterHyphen.spaceAfter !== '5pt') throw new Error('hyphenation changed paragraph spacing: ' + JSON.stringify(spacingAfterHyphen));

    __SUI_RICH.set('editor', { type: 'doc', page: { lang: 'it' }, children: [{ type: 'paragraph', hyphens: 'auto', children: [{ type: 'text', text: 'sillabazione', marks: [] }] }] });
    await wait(220);
    var dictIt = bsEditor().querySelector('p').textContent.replace(/\u00AD/g, '-');
    if (dictIt !== 'sil-la-ba-zio-ne') throw new Error('Italian hyphenation dictionary: ' + dictIt);
    __SUI_RICH.set('editor', { type: 'doc', page: { lang: 'en' }, children: [{ type: 'paragraph', hyphens: 'auto', children: [{ type: 'text', text: 'hyphenation', marks: [] }] }] });
    await wait(220);
    var dictEn = bsEditor().querySelector('p').textContent.replace(/\u00AD/g, '-');
    if (dictEn !== 'hy-phen-ation') throw new Error('English hyphenation dictionary: ' + dictEn);
    var dictCases = [
      ['de', 'Silbentrennung', 'Sil-ben-tren-nung'],
      ['fr', 'syllabation', 'syl-la-ba-tion'],
      ['es', 'silabación', 'si-la-ba-ción'],
      ['pt', 'silabação', 'si-la-ba-ção'],
      ['nl', 'lettergrepen', 'let-ter-gre-pen'],
      ['pl', 'sylabizacja', 'sy-la-bi-za-cja'],
      ['sv', 'stavelser', 'sta-vel-ser']
    ];
    for (var dc = 0; dc < dictCases.length; dc++) {
      __SUI_RICH.set('editor', { type: 'doc', page: { lang: dictCases[dc][0] }, children: [{ type: 'paragraph', hyphens: 'auto', children: [{ type: 'text', text: dictCases[dc][1], marks: [] }] }] });
      await wait(220);
      var dictValue = bsEditor().querySelector('p').textContent.replace(/\u00AD/g, '-');
      if (dictValue !== dictCases[dc][2]) throw new Error(dictCases[dc][0] + ' hyphenation dictionary: ' + dictValue);
    }

    __SUI_RICH.command('editor', 'documentSettings');
    await wait(200);
    var reqDocDefaults = document.querySelector('.sui-rich-documentsettings');
    reqDocDefaults.querySelector('[data-rich-doc-size]').value = 'a4';
    reqDocDefaults.querySelector('[data-rich-doc-orientation]').value = 'portrait';
    reqDocDefaults.querySelector('[data-rich-doc-margin="top"]').value = '15';
    reqDocDefaults.querySelector('[data-rich-doc-margin="right"]').value = '17';
    reqDocDefaults.querySelector('[data-rich-doc-margin="bottom"]').value = '15';
    reqDocDefaults.querySelector('[data-rich-doc-margin="left"]').value = '17';
    reqDocDefaults.querySelector('[data-rich-doc-lang]').value = 'it';
    reqDocDefaults.querySelector('[data-rich-doc-apply]').click();
    await wait(300);
    if (bsEditor().getAttribute('lang') !== 'it') throw new Error('document language not applied to the editor');
    __SUI_RICH.command('editor', 'themeLight');
    await wait(200);
    __SUI_RICH.set('editor', { type: 'doc', header: 'H {page}/{pages}', footer: '', children: manyRows });
    await wait(400);
    richRoot().querySelector('.sui-rich-toolbar [data-rich-cmd="printPreview"]').click();
    await wait(1500);
    var reqPreviewCard = document.querySelector('.sui-rich-preview-card');
    var reqPreviewButton = document.querySelector('.sui-rich-preview-bar button');
    if (!reqPreviewCard || getComputedStyle(reqPreviewCard).backgroundColor === 'rgb(29, 33, 41)') throw new Error('print preview does not follow the light theme');
    if (!reqPreviewButton || getComputedStyle(reqPreviewButton).color === getComputedStyle(reqPreviewButton).backgroundColor) throw new Error('print preview button text is not visible');
    var reqPreviewPages = document.querySelector('.sui-rich-preview-pages');
    var reqPreviewSheets = reqPreviewPages ? reqPreviewPages.querySelectorAll('.sui-rich-page-sheet') : [];
    if (!reqPreviewPages || reqPreviewSheets.length < 2) throw new Error('print preview is not split into pages: ' + reqPreviewSheets.length);
    var reqFirstSheet = reqPreviewSheets[0];
    if (reqFirstSheet.style.width !== '1056px' || reqFirstSheet.style.minHeight !== '816px') throw new Error('preview page size wrong: ' + reqFirstSheet.style.width + ' x ' + reqFirstSheet.style.minHeight);
    if (reqFirstSheet.style.padding !== '76px 94px 68px 83px') throw new Error('preview page margins missing: ' + reqFirstSheet.style.padding);
    if (reqPreviewSheets[1] && reqPreviewSheets[1].style.width === '1056px') throw new Error('page 2 wrongly inherits the page 1 override');
    if (!/H 1\//.test((reqFirstSheet.querySelector('.sui-rich-sheet-header') || {}).textContent || '')) throw new Error('preview header missing');
    var reqPreviewClose = [].slice.call(document.querySelectorAll('.sui-rich-preview button')).filter(function (b) { return b.textContent === 'Close'; })[0];
    if (reqPreviewClose) reqPreviewClose.click();
    await wait(300);
    __SUI_RICH.command('editor', 'themeDark');
    var reqGuideCount = document.querySelectorAll('.sui-rich-guides .sui-rich-guide').length;
    if (!reqGuideCount) throw new Error('automatic page break guides are missing in the editor');
    var reqGuide = document.querySelector('.sui-rich-guide');
    var reqGuideRect = reqGuide.getBoundingClientRect();
    reqGuide.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqGuideRect.left + 10, clientY: reqGuideRect.top + 2 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: reqGuideRect.left + 10, clientY: reqGuideRect.top - 120 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: reqGuideRect.left + 10, clientY: reqGuideRect.top - 120 }));
    await wait(400);
    var reqDraggedDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (!reqDraggedDoc.children.some(function (c) { return c.type === 'page_break'; })) throw new Error('dragging a page guide did not insert a page break');
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(300);
    if (document.querySelectorAll('.sui-rich-guides .sui-rich-guide').length) throw new Error('page guides must be hidden in paginated view');
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(200);

    // A page break can be dragged earlier and deleted; every following guide
    // must be recalculated after each change.
    var reqBreakRows = [];
    for (var brp = 0; brp < 90; brp++) {
      reqBreakRows.push({ type: 'paragraph', children: [{ type: 'text', text: 'Break row ' + brp + ' with enough text to wrap a line.', marks: [] }] });
      if (brp === 5) reqBreakRows.push({ type: 'page_break' });
    }
    __SUI_RICH.set('editor', { type: 'doc', children: reqBreakRows });
    await wait(500);
    bsEditor().scrollTop = 0;
    await wait(250);
    function reqGuideTops() {
      return [].slice.call(document.querySelectorAll('.sui-rich-guides .sui-rich-guide')).map(function (g) { return parseFloat(g.style.top) || 0; });
    }
    var reqBreakBefore = reqGuideTops();
    if (!reqBreakBefore.length) throw new Error('page break test needs automatic guides');
    var reqBreakRoot = richRoot();
    if (getComputedStyle(reqBreakRoot).getPropertyValue('--sui-rich-break').trim() !== '#dc2626') throw new Error('page break colour variable is not red');
    var reqBreakLineColor = getComputedStyle(document.querySelector('.sui-rich-pagebreak')).borderTopColor;
    if (!/220,\s*38,\s*38/.test(reqBreakLineColor)) throw new Error('page break line is not red: ' + reqBreakLineColor);
    // Guides must keep a document position and follow the editor scroll.
    var reqAnchorGuide = document.querySelector('.sui-rich-guide');
    var reqAnchorCut = reqAnchorGuide.getAttribute('data-cut');
    var reqAnchorTop = parseFloat(reqAnchorGuide.style.top) || 0;
    bsEditor().scrollTop = 500;
    bsEditor().dispatchEvent(new Event('scroll'));
    await wait(220);
    reqAnchorGuide = document.querySelector('.sui-rich-guide');
    if (reqAnchorGuide.getAttribute('data-cut') !== reqAnchorCut) throw new Error('automatic guide lost its document anchor while scrolling');
    var reqAnchorTop2 = parseFloat(reqAnchorGuide.style.top) || 0;
    if (Math.abs((reqAnchorTop - reqAnchorTop2) - 500) > 5) throw new Error('automatic guide did not follow the editor scroll: top ' + reqAnchorTop + ' -> ' + reqAnchorTop2 + ', scroll ' + bsEditor().scrollTop + ', cut ' + reqAnchorCut + ', inRoot ' + richRoot().contains(reqAnchorGuide) + ', roots ' + document.querySelectorAll('[data-sui-rich]').length + ', eds ' + document.querySelectorAll('.sui-rich-editor').length + ', layers ' + document.querySelectorAll('.sui-rich-guides').length);
    bsEditor().scrollTop = 0;
    await wait(220);
    var reqBreakHandle = document.querySelector('.sui-rich-pagebreak [data-rich-pb-handle]');
    if (!reqBreakHandle) throw new Error('page break drag handle missing');
    var reqBreakHandleRect = reqBreakHandle.getBoundingClientRect();
    var reqBreakParas = [].slice.call(bsEditor().children).filter(function (c) { return c.tagName === 'P'; });
    var reqBreakTargetRect = reqBreakParas[2].getBoundingClientRect();
    reqBreakHandle.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqBreakHandleRect.left + 2, clientY: reqBreakHandleRect.top + 2 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: reqBreakHandleRect.left + 2, clientY: reqBreakTargetRect.top + 2 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: reqBreakHandleRect.left + 2, clientY: reqBreakTargetRect.top + 2 }));
    await wait(600);
    var reqBreakDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var reqBreakIndex = reqBreakDoc.children.findIndex(function (c) { return c.type === 'page_break'; });
    if (reqBreakIndex !== 2) throw new Error('page break was not moved before the target paragraph: ' + reqBreakIndex);
    var reqBreakAfterMove = reqGuideTops();
    if (!reqBreakAfterMove.length || reqBreakAfterMove[0] >= reqBreakBefore[0]) throw new Error('following guides were not recalculated after moving the break');
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(800);
    var reqBreakPages = richRoot().querySelectorAll('.sui-rich-page-sheet');
    if (reqBreakPages.length < 2) throw new Error('moved break did not repaginate: ' + reqBreakPages.length);
    for (var rbp = 0; rbp < reqBreakPages.length; rbp++) {
      var rbpBody = reqBreakPages[rbp].querySelector('.sui-rich-sheet-body');
      if (rbpBody && rbpBody.scrollHeight > rbpBody.clientHeight + 1) throw new Error('moved break leaves content past the printable area');
    }
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(300);
    var reqBreakDelete = document.querySelector('.sui-rich-pagebreak [data-rich-pb-delete]');
    if (!reqBreakDelete) throw new Error('page break delete control missing');
    reqBreakDelete.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
    await wait(500);
    var reqBreakDeleted = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (reqBreakDeleted.children.some(function (c) { return c.type === 'page_break'; })) throw new Error('page break was not deleted');
    if (document.querySelectorAll('.sui-rich-pagebreak').length) throw new Error('page break node remained after delete');
    var reqBreakAfterDelete = reqGuideTops();
    if (!reqBreakAfterDelete.length || reqBreakAfterDelete[0] >= reqBreakAfterMove[0]) throw new Error('guides were not recalculated after deleting the break');
    if (reqBreakAfterDelete.length > 1 && reqBreakAfterDelete[1] - reqBreakAfterDelete[0] < 700) throw new Error('page guides are not spaced by the full page height');

    // The whole red page-break line (label included) must be draggable, not
    // only the small handle, and a simple click must not move it.
    var reqLineRows = [];
    for (var lr2 = 0; lr2 < 40; lr2++) {
      reqLineRows.push({ type: 'paragraph', children: [{ type: 'text', text: 'Draggable line row ' + lr2 + ' with enough text for a line.', marks: [] }] });
      if (lr2 === 20) reqLineRows.push({ type: 'page_break' });
    }
    __SUI_RICH.set('editor', { type: 'doc', pageOverrides: {}, children: reqLineRows });
    await wait(500);
    var reqLineBreak = document.querySelector('.sui-rich-pagebreak');
    reqLineBreak.scrollIntoView({ block: 'center' });
    await wait(300);
    var reqLineLabel = reqLineBreak.querySelector('.sui-rich-pagebreak-label');
    var reqLineRect = reqLineLabel.getBoundingClientRect();
    var reqLineX = Math.round(reqLineRect.left + 20), reqLineY = Math.round(reqLineRect.top + reqLineRect.height / 2);
    reqLineLabel.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqLineX, clientY: reqLineY }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: reqLineX, clientY: reqLineY }));
    await wait(250);
    var reqLineClickDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (reqLineClickDoc.children.findIndex(function (c) { return c.type === 'page_break'; }) !== 21) throw new Error('a click on the break line moved it');
    reqLineLabel.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqLineX, clientY: reqLineY }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: reqLineX, clientY: reqLineY - 80 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: reqLineX, clientY: reqLineY - 80 }));
    await wait(600);
    var reqLineDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var reqLineIndex = reqLineDoc.children.findIndex(function (c) { return c.type === 'page_break'; });
    if (!(reqLineIndex < 21 && reqLineIndex > 0)) throw new Error('page break line was not dragged to an earlier position: ' + reqLineIndex);
    if (document.querySelectorAll('.sui-rich-pagebreak').length !== 1) throw new Error('dragging the break line duplicated the break');

    // Dragging a guide while the editor is scrolled must move the target in
    // document coordinates, never jump to the first block of the document.
    var reqScrollRows = [];
    for (var sgr = 0; sgr < 90; sgr++) reqScrollRows.push({ type: 'paragraph', children: [{ type: 'text', text: 'Scrolled guide row ' + sgr + ' with enough text for a line.', marks: [] }] });
    __SUI_RICH.set('editor', { type: 'doc', pageOverrides: {}, children: reqScrollRows });
    await wait(500);
    var reqScrollEditor = bsEditor();
    reqScrollEditor.focus();
    reqScrollEditor.scrollTop = 900;
    reqScrollEditor.dispatchEvent(new Event('scroll'));
    await wait(300);
    var reqScrollGuide = null;
    var reqScrollGuides = [].slice.call(document.querySelectorAll('.sui-rich-guide'));
    var reqScrollRect = reqScrollEditor.getBoundingClientRect();
    for (var sgi = 0; sgi < reqScrollGuides.length; sgi++) {
      var sgRect = reqScrollGuides[sgi].getBoundingClientRect();
      if (sgRect.top >= reqScrollRect.top + 20 && sgRect.top <= reqScrollRect.bottom - 20) { reqScrollGuide = reqScrollGuides[sgi]; break; }
    }
    if (!reqScrollGuide) {
      var reqScrollDebug = [].slice.call(reqScrollGuides).map(function (g) { var gr = g.getBoundingClientRect(); return Math.round(gr.top) + '/' + Math.round(gr.bottom) + '@' + g.getAttribute('data-cut'); }).join(', ');
      throw new Error('no visible guide while the editor is scrolled: scroll ' + reqScrollEditor.scrollTop + ', editor ' + Math.round(reqScrollRect.top) + '..' + Math.round(reqScrollRect.bottom) + ', guides ' + reqScrollGuides.length + ' [' + reqScrollDebug + ']');
    }
    var reqScrollBefore = reqScrollEditor.scrollTop;
    var reqScrollGuideRect = reqScrollGuide.getBoundingClientRect();
    reqScrollGuide.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqScrollGuideRect.left + 30, clientY: reqScrollGuideRect.top + 2 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: reqScrollGuideRect.left + 30, clientY: reqScrollGuideRect.top - 90 }));
    await wait(180);
    if (!reqScrollGuide.isConnected) throw new Error('guide was detached while dragging');
    if (bsEditor() !== reqScrollEditor) throw new Error('editor was re-rendered while dragging the guide');
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: reqScrollGuideRect.left + 30, clientY: reqScrollGuideRect.top - 90 }));
    await wait(600);
    var reqScrollDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var reqScrollBreakIndex = reqScrollDoc.children.findIndex(function (c) { return c.type === 'page_break'; });
    if (reqScrollBreakIndex <= 5) throw new Error('scrolled guide drag created a break at the start: ' + reqScrollBreakIndex);
    if (document.querySelectorAll('.sui-rich-pagebreak').length !== 1) throw new Error('scrolled guide drag created duplicate breaks');
    if (reqScrollBefore > 200 && bsEditor().scrollTop === 0) throw new Error('editor scroll was reset after the guide drag: before ' + reqScrollBefore);

    // A guide inside a very long table must split the table at a row boundary,
    // not create a break before/after the whole table (which would land at the
    // document start when the table is the first block).
    var reqTableRows = [{ type: 'table_row', cells: [
      { type: 'table_cell', header: true, children: [{ type: 'text', text: 'Name', marks: [] }] },
      { type: 'table_cell', header: true, children: [{ type: 'text', text: 'Value', marks: [] }] }
    ] }];
    for (var tr2 = 0; tr2 < 70; tr2++) reqTableRows.push({ type: 'table_row', cells: [
      { type: 'table_cell', children: [{ type: 'text', text: 'Row ' + tr2, marks: [] }] },
      { type: 'table_cell', children: [{ type: 'text', text: 'Value ' + tr2, marks: [] }] }
    ] });
    __SUI_RICH.set('editor', { type: 'doc', pageOverrides: {}, children: [{ type: 'table', header: true, rows: reqTableRows }] });
    await wait(800);
    var reqTableEditor = bsEditor();
    var reqTableGuides0 = [].slice.call(document.querySelectorAll('.sui-rich-guide'));
    if (!reqTableGuides0.length) throw new Error('long table has no automatic guides');
    var reqFirstCut = parseFloat(reqTableGuides0[0].getAttribute('data-cut')) || 0;
    reqTableEditor.scrollTop = Math.max(0, reqFirstCut - 200);
    reqTableEditor.dispatchEvent(new Event('scroll'));
    await wait(300);
    var reqTableGuides = [].slice.call(document.querySelectorAll('.sui-rich-guide'));
    var reqTableGuide = null;
    for (var tg = 0; tg < reqTableGuides.length; tg++) {
      var tgRect = reqTableGuides[tg].getBoundingClientRect();
      var teRect = reqTableEditor.getBoundingClientRect();
      if (tgRect.top > teRect.top + 20 && tgRect.top < teRect.bottom - 20) { reqTableGuide = reqTableGuides[tg]; break; }
    }
    if (!reqTableGuide) throw new Error('no guide visible inside the long table');
    var reqTableGuideRect = reqTableGuide.getBoundingClientRect();
    reqTableGuide.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqTableGuideRect.left + 30, clientY: reqTableGuideRect.top + 1 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: reqTableGuideRect.left + 30, clientY: reqTableGuideRect.top - 90 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: reqTableGuideRect.left + 30, clientY: reqTableGuideRect.top - 90 }));
    await wait(800);
    var reqTableDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    if (reqTableDoc.children.length !== 3 || reqTableDoc.children[0].type !== 'table' || reqTableDoc.children[1].type !== 'page_break' || reqTableDoc.children[2].type !== 'table') throw new Error('long table was not split around a page break: ' + JSON.stringify(reqTableDoc.children.map(function (c) { return c.type; })));
    if (document.querySelectorAll('.sui-rich-pagebreak').length !== 1) throw new Error('long table split created duplicate breaks');
    var reqBody = 0, reqLabels = {};
    for (var tb = 0; tb < reqTableDoc.children.length; tb++) {
      var tbl = reqTableDoc.children[tb];
      if (tbl.type !== 'table') continue;
      var hc2 = 0;
      for (var rh2 = 0; rh2 < tbl.rows.length; rh2++) {
        var rc2 = tbl.rows[rh2].cells || [], allH2 = rc2.length > 0;
        for (var rc3 = 0; rc3 < rc2.length; rc3++) if (!rc2[rc3].header) allH2 = false;
        if (allH2) hc2++; else break;
      }
      for (var rb2 = hc2; rb2 < tbl.rows.length; rb2++) {
        reqBody++;
        var cell2 = tbl.rows[rb2].cells[0] || {}, txt2 = (cell2.children || []).map(function (c) { return c.text || ''; }).join('');
        if (/^Row \d+$/.test(txt2)) reqLabels[txt2] = true;
      }
    }
    if (reqBody !== 70 || Object.keys(reqLabels).length !== 70) throw new Error('long table split lost or duplicated rows: body ' + reqBody + ', labels ' + Object.keys(reqLabels).length);
    var reqExportHtml = __SUI_RICH.exportFormat('editor', 'html');
    if (reqExportHtml[0] && reqExportHtml[0].indexOf('data-sui-table-group') >= 0) throw new Error('internal table group leaked into exported HTML');
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(5000);
    var reqTableSheets = richRoot().querySelectorAll('.sui-rich-page-sheet');
    var reqSeen = {};
    for (var ts2 = 0; ts2 < reqTableSheets.length; ts2++) {
      var tsBody = reqTableSheets[ts2].querySelector('.sui-rich-sheet-body');
      if (tsBody && tsBody.scrollHeight > tsBody.clientHeight + 1) throw new Error('split long table exceeds the printable area');
      if (!tsBody) continue;
      var tsRows = tsBody.querySelectorAll('table tr');
      for (var tr3 = 0; tr3 < tsRows.length; tr3++) {
        var tcell = tsRows[tr3].querySelector('td');
        var ttxt = tcell ? tcell.textContent.trim() : '';
        if (/^Row \d+$/.test(ttxt)) reqSeen[ttxt] = (reqSeen[ttxt] || 0) + 1;
      }
    }
    if (Object.keys(reqSeen).length !== 70) throw new Error('paginated split table rows: ' + Object.keys(reqSeen).length);
    for (var sk in reqSeen) if (reqSeen[sk] !== 1) throw new Error('paginated split table duplicated row ' + sk);
    var reqTablePages = reqTableSheets.length;
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(300);
    var reqTableAutoGuides = document.querySelectorAll('.sui-rich-guide').length;
    var reqTableExplicit = document.querySelectorAll('.sui-rich-pagebreak').length;
    if (reqTablePages !== reqTableAutoGuides + reqTableExplicit + 1) throw new Error('editor guides and print pages disagree: pages ' + reqTablePages + ', guides ' + reqTableAutoGuides + ', explicit ' + reqTableExplicit);
    // Moving the explicit break line into the second table fragment must split
    // that fragment too, without duplicating or losing rows.
    var reqMoveBreak = document.querySelector('.sui-rich-pagebreak');
    reqMoveBreak.scrollIntoView({ block: 'center' });
    await wait(300);
    var reqMoveLabel = reqMoveBreak.querySelector('.sui-rich-pagebreak-label');
    var reqMoveRect = reqMoveLabel.getBoundingClientRect();
    reqMoveLabel.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqMoveRect.left + 20, clientY: reqMoveRect.top + reqMoveRect.height / 2 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: reqMoveRect.left + 20, clientY: reqMoveRect.top + reqMoveRect.height / 2 + 140 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, clientX: reqMoveRect.left + 20, clientY: reqMoveRect.top + reqMoveRect.height / 2 + 140 }));
    await wait(800);
    var reqMovedDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var reqMovedBreaks = reqMovedDoc.children.filter(function (c) { return c.type === 'page_break'; });
    if (reqMovedBreaks.length !== 1 || document.querySelectorAll('.sui-rich-pagebreak').length !== 1) throw new Error('moving the break inside the table duplicated it');
    var reqMovedBody = 0, reqMovedLabels = {};
    for (var mb = 0; mb < reqMovedDoc.children.length; mb++) {
      var mbt = reqMovedDoc.children[mb];
      if (mbt.type !== 'table') continue;
      var mhc = 0;
      for (var mr = 0; mr < mbt.rows.length; mr++) {
        var mrc = mbt.rows[mr].cells || [], mall = mrc.length > 0;
        for (var mc = 0; mc < mrc.length; mc++) if (!mrc[mc].header) mall = false;
        if (mall) mhc++; else break;
      }
      for (var mrb = mhc; mrb < mbt.rows.length; mrb++) {
        reqMovedBody++;
        var mcell = mbt.rows[mrb].cells[0] || {}, mtxt = (mcell.children || []).map(function (c) { return c.text || ''; }).join('');
        if (/^Row \d+$/.test(mtxt)) reqMovedLabels[mtxt] = true;
      }
    }
    if (reqMovedBody !== 70 || Object.keys(reqMovedLabels).length !== 70) throw new Error('moving the break inside the table lost rows: body ' + reqMovedBody + ', labels ' + Object.keys(reqMovedLabels).length);
    // Deleting the break between adjacent fragments must recompose one table.
    var reqMergeDelete = document.querySelector('.sui-rich-pagebreak [data-rich-pb-delete]');
    if (!reqMergeDelete) throw new Error('merged table break delete control missing');
    reqMergeDelete.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
    await wait(800);
    var reqMergedDoc = JSON.parse(__SUI_RICH.get('editor')[0]);
    var reqMergedTables = reqMergedDoc.children.filter(function (c) { return c.type === 'table'; });
    if (reqMergedDoc.children.length !== 1 || reqMergedTables.length !== 1) throw new Error('adjacent table fragments were not merged: ' + JSON.stringify(reqMergedDoc.children.map(function (c) { return c.type; })));
    if (reqMergedTables[0]._tableGroup) throw new Error('merged table kept a fragment group');
    var reqMergedBody = 0, reqMergedLabels = {};
    var reqMergedHc = 0;
    for (var mh = 0; mh < reqMergedTables[0].rows.length; mh++) {
      var mhcells = reqMergedTables[0].rows[mh].cells || [], mhall = mhcells.length > 0;
      for (var mhc = 0; mhc < mhcells.length; mhc++) if (!mhcells[mhc].header) mhall = false;
      if (mhall) reqMergedHc++; else break;
    }
    for (var mrb2 = reqMergedHc; mrb2 < reqMergedTables[0].rows.length; mrb2++) {
      reqMergedBody++;
      var mmcell = reqMergedTables[0].rows[mrb2].cells[0] || {}, mmtxt = (mmcell.children || []).map(function (c) { return c.text || ''; }).join('');
      if (/^Row \d+$/.test(mmtxt)) reqMergedLabels[mmtxt] = true;
    }
    if (reqMergedBody !== 70 || Object.keys(reqMergedLabels).length !== 70) throw new Error('merged table lost rows: body ' + reqMergedBody + ', labels ' + Object.keys(reqMergedLabels).length);

    // A long table with a rowspan group after a preamble must start on page 1
    // and fill it; the merged group continues on page 2 instead of moving the
    // whole table and leaving the first page mostly empty.
    var reqFillRows = [{ type: 'table_row', cells: [
      { type: 'table_cell', header: true, children: [{ type: 'text', text: 'Name', marks: [] }] },
      { type: 'table_cell', header: true, children: [{ type: 'text', text: 'Value', marks: [] }] }
    ] }];
    for (var fr = 0; fr < 40; fr++) {
      var fcells = [];
      if (fr === 0) fcells.push({ type: 'table_cell', rowspan: 8, children: [{ type: 'text', text: 'Group', marks: [] }] });
      else if (fr > 0 && fr < 8) fcells.push({ type: 'table_cell', children: [{ type: 'text', text: '', marks: [] }] });
      else fcells.push({ type: 'table_cell', children: [{ type: 'text', text: 'Row ' + fr, marks: [] }] });
      fcells.push({ type: 'table_cell', children: [{ type: 'text', text: 'Value ' + fr, marks: [] }] });
      reqFillRows.push({ type: 'table_row', cells: fcells });
    }
    var reqFillKids = [];
    for (var fp = 0; fp < 18; fp++) reqFillKids.push({ type: 'paragraph', children: [{ type: 'text', text: 'Preamble ' + fp + ' before the table.', marks: [] }] });
    reqFillKids.push({ type: 'table', header: true, rows: reqFillRows });
    __SUI_RICH.set('editor', { type: 'doc', pageOverrides: {}, children: reqFillKids });
    await wait(1500);
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(6000);
    var reqFillSheets = richRoot().querySelectorAll('.sui-rich-page-sheet');
    var reqFillBody = reqFillSheets[0] ? reqFillSheets[0].querySelector('.sui-rich-sheet-body') : null;
    if (!reqFillBody || !reqFillBody.querySelector('table')) throw new Error('long table did not start on the first page');
    var reqFillMax = 0, reqFillRect = reqFillBody.getBoundingClientRect();
    for (var fc = 0; fc < reqFillBody.children.length; fc++) {
      var fcr = reqFillBody.children[fc].getBoundingClientRect();
      if (fcr.bottom > reqFillMax) reqFillMax = fcr.bottom;
    }
    var reqFillFree = Math.round(reqFillRect.bottom - reqFillMax);
    if (reqFillFree > 100) throw new Error('first page left too much free space: ' + reqFillFree);
    for (var fs = 0; fs < reqFillSheets.length; fs++) {
      var fsBody = reqFillSheets[fs].querySelector('.sui-rich-sheet-body');
      if (fsBody && fsBody.scrollHeight > fsBody.clientHeight + 1) throw new Error('rowspan table exceeds the printable area on page ' + (fs + 1));
    }
    var reqFillPages = reqFillSheets.length;
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(300);
    var reqFillGuides = document.querySelectorAll('.sui-rich-guide').length;
    if (reqFillPages !== reqFillGuides + 1) throw new Error('rowspan table guide/page mismatch: pages ' + reqFillPages + ', guides ' + reqFillGuides);

    // A paragraph taller than the printable area is split across pages, in the
    // paginated view and in print preview, and never crosses the bottom margin.
    var reqLongWords = [];
    for (var lw = 0; lw < 2000; lw++) reqLongWords.push('longword' + lw);
    __SUI_RICH.set('editor', { type: 'doc', children: [
      { type: 'paragraph', children: [{ type: 'text', text: 'Intro before a very long paragraph.', marks: [] }] },
      { type: 'paragraph', children: [{ type: 'text', text: reqLongWords.join(' '), marks: [] }] },
      { type: 'paragraph', children: [{ type: 'text', text: 'Tail after the long paragraph.', marks: [] }] }
    ] });
    await wait(500);
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(5000);
    var reqLongSheets = richRoot().querySelectorAll('.sui-rich-page-sheet');
    if (reqLongSheets.length < 4) throw new Error('oversized paragraph was not split across pages: ' + reqLongSheets.length);
    for (var ls = 0; ls < reqLongSheets.length; ls++) {
      var lsBody = reqLongSheets[ls].querySelector('.sui-rich-sheet-body');
      if (lsBody && lsBody.scrollHeight > lsBody.clientHeight + 1) throw new Error('oversized paragraph exceeds the printable area: sheet ' + ls + ' by ' + (lsBody.scrollHeight - lsBody.clientHeight) + 'px');
    }
    __SUI_RICH.command('editor', 'printPreview');
    await wait(3000);
    var reqLongPages = reqLongSheets.length;
    var reqLongPreview = document.querySelectorAll('.sui-rich-preview-pages .sui-rich-page-sheet');
    if (reqLongPreview.length < 4) throw new Error('oversized paragraph preview pages: ' + reqLongPreview.length);
    if (reqLongPreview.length !== reqLongPages) throw new Error('paginated view and preview page counts differ: ' + reqLongPages + ' vs ' + reqLongPreview.length);
    for (var lp = 0; lp < reqLongPreview.length; lp++) {
      var lpBody = reqLongPreview[lp].querySelector('.sui-rich-sheet-body');
      if (lpBody && lpBody.scrollHeight > lpBody.clientHeight + 1) throw new Error('preview oversized paragraph exceeds the printable area: page ' + lp + ' by ' + (lpBody.scrollHeight - lpBody.clientHeight) + 'px');
    }
    var reqLongClose = [].slice.call(document.querySelectorAll('.sui-rich-preview button')).filter(function (b) { return b.textContent === 'Close'; })[0];
    if (reqLongClose) reqLongClose.click();
    await wait(300);
    __SUI_RICH.command('editor', 'paginatedView');
    await wait(300);
    var reqLongAutoGuides = document.querySelectorAll('.sui-rich-guide').length;
    if (reqLongPages !== reqLongAutoGuides + 1) throw new Error('oversized paragraph guide/page mismatch: pages ' + reqLongPages + ', guides ' + reqLongAutoGuides);

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Link target', marks: [] }] }] });
    await wait(300);
    var reqLinkText = bsEditor().querySelector('p').firstChild;
    var reqLinkRange = document.createRange(); reqLinkRange.setStart(reqLinkText, 0); reqLinkRange.setEnd(reqLinkText, 4);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(reqLinkRange);
    richRoot().querySelector('.sui-rich-toolbar [data-rich-cmd="link"]').click();
    await wait(200);
    var reqPrompt = document.querySelector('.sui-rich-prompt');
    if (!reqPrompt) throw new Error('link overlay missing');
    if (reqPrompt.querySelector('.sui-rich-tableprops-title, h1, h2, h3')) throw new Error('data overlay must not have a heading');
    if (!/Enter the link URL/.test(reqPrompt.title || '')) throw new Error('link overlay hint missing');
    reqPrompt.querySelector('input').value = 'https://example.com';
    [].slice.call(reqPrompt.querySelectorAll('button')).filter(function (b) { return b.textContent === 'Insert link'; })[0].click();
    await wait(300);
    if (!bsEditor().querySelector('a[href="https://example.com"]')) throw new Error('link overlay did not insert the link');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Alpha', marks: [{ type: 'bold' }] }, { type: 'text', text: ' Beta', marks: [] }] }] });
    await wait(300);
    var reqAlpha = null, reqWalk = document.createTreeWalker(bsEditor(), NodeFilter.SHOW_TEXT, null), reqNode;
    while ((reqNode = reqWalk.nextNode())) if (reqNode.nodeValue.indexOf('Alpha') >= 0) { reqAlpha = reqNode; break; }
    var reqAlphaRange = document.createRange(); reqAlphaRange.setStart(reqAlpha, 2); reqAlphaRange.collapse(true);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(reqAlphaRange);
    __SUI_RICH.command('editor', 'clearFormat');
    await wait(300);
    if (bsEditor().querySelector('p strong, p b, p span[style*="font-weight"]')) throw new Error('clear formatting did not act on the word at the caret');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Bold', marks: [{ type: 'bold' }] }, { type: 'text', text: ' plain', marks: [] }] }] });
    await wait(300);
    var reqBoldNode = null, reqPlainNode = null, reqWalk2 = document.createTreeWalker(bsEditor(), NodeFilter.SHOW_TEXT, null), reqNode2;
    while ((reqNode2 = reqWalk2.nextNode())) {
      if (reqNode2.nodeValue.indexOf('Bold') >= 0) reqBoldNode = reqNode2;
      if (reqNode2.nodeValue.indexOf('plain') >= 0) reqPlainNode = reqNode2;
    }
    var reqBoldRange = document.createRange(); reqBoldRange.setStart(reqBoldNode, 0); reqBoldRange.setEnd(reqBoldNode, 4);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(reqBoldRange);
    __SUI_RICH.command('editor', 'copyFormat');
    await wait(150);
    if (!bsBtn('copyFormat').classList.contains('on')) throw new Error('copy format did not activate');
    var reqPlainRange = document.createRange(); reqPlainRange.setStart(reqPlainNode, reqPlainNode.nodeValue.indexOf('plain')); reqPlainRange.setEnd(reqPlainNode, reqPlainNode.nodeValue.length);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(reqPlainRange);
    bsEditor().dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));
    await wait(300);
    var reqPlainStyle = getComputedStyle(reqPlainNode.parentElement);
    if (parseInt(reqPlainStyle.fontWeight, 10) < 600) throw new Error('copy format did not paint the target word');
    if (bsBtn('copyFormat').classList.contains('on')) throw new Error('format painter stayed active after painting');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Size me', marks: [] }] }] });
    await wait(300);
    var reqSizeNode = null, reqSizeWalk = document.createTreeWalker(bsEditor(), NodeFilter.SHOW_TEXT, null), reqSizeText;
    while ((reqSizeText = reqSizeWalk.nextNode())) if (reqSizeText.nodeValue.indexOf('Size me') >= 0) { reqSizeNode = reqSizeText; break; }
    var reqSizeRange = document.createRange(); reqSizeRange.setStart(reqSizeNode, 0); reqSizeRange.setEnd(reqSizeNode, 7);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(reqSizeRange);
    __SUI_RICH.command('editor', 'fontSizeUp');
    await wait(300);
    var reqSizeSpan = bsEditor().querySelector('span[style*="font-size"]');
    if (!reqSizeSpan || parseFloat(reqSizeSpan.style.fontSize) !== 18) throw new Error('increase font size did not apply: ' + (reqSizeSpan ? reqSizeSpan.style.fontSize : 'missing'));
    var reqSizeDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0].children[0];
    if (!reqSizeDoc.marks || !reqSizeDoc.marks.some(function (m) { return m.type === 'size' && parseFloat(m.value) === 18; })) throw new Error('font size mark not stored: ' + JSON.stringify(reqSizeDoc.marks));
    var reqSizeNode2 = null, reqSizeWalk2 = document.createTreeWalker(bsEditor(), NodeFilter.SHOW_TEXT, null), reqSizeText2;
    while ((reqSizeText2 = reqSizeWalk2.nextNode())) if (reqSizeText2.nodeValue.indexOf('Size me') >= 0) { reqSizeNode2 = reqSizeText2; break; }
    var reqSizeRange2 = document.createRange(); reqSizeRange2.setStart(reqSizeNode2, 0); reqSizeRange2.setEnd(reqSizeNode2, 7);
    window.getSelection().removeAllRanges(); window.getSelection().addRange(reqSizeRange2);
    __SUI_RICH.command('editor', 'fontSizeDown');
    await wait(300);
    var reqSizeSpan2 = bsEditor().querySelector('span[style*="font-size"]');
    if (!reqSizeSpan2 || parseFloat(reqSizeSpan2.style.fontSize) !== 16) throw new Error('decrease font size did not reduce the selection: ' + (reqSizeSpan2 ? reqSizeSpan2.style.fontSize : 'missing') + ' html=' + bsEditor().innerHTML + ' model=' + JSON.stringify(JSON.parse(__SUI_RICH.get('editor')[0]).children[0].children));

    var reqPng = 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI4MCIgaGVpZ2h0PSI0MCI+PHJlY3Qgd2lkdGg9IjgwIiBoZWlnaHQ9IjQwIiBmaWxsPSIjNDQ4OCIvPjwvc3ZnPg==';
    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'image', src: reqPng, alt: 'Old', width: '80px', marks: [] }] }] });
    await wait(300);
    var reqImg = bsEditor().querySelector('img');
    if (!reqImg) throw new Error('image fixture missing');
    var reqImgRect = reqImg.getBoundingClientRect();
    reqImg.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, cancelable: true, clientX: reqImgRect.left + 5, clientY: reqImgRect.top + 5 }));
    await wait(150);
    var reqImgMenu = document.querySelector('.sui-rich-image-menu');
    if (!reqImgMenu) throw new Error('image context menu did not open');
    [].slice.call(reqImgMenu.querySelectorAll('button')).filter(function (b) { return b.textContent === 'Edit image'; })[0].click();
    await wait(250);
    var reqImgPanel = document.querySelector('.sui-rich-imagesettings');
    if (!reqImgPanel || reqImgPanel.hidden) throw new Error('image settings panel did not open');
    var reqPng2 = 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgaGVpZ2h0PSI0MCI+PHJlY3Qgd2lkdGg9IjQwIiBoZWlnaHQ9IjQwIiBmaWxsPSIjYzMzIi8+PC9zdmc+';
    reqImgPanel.querySelector('[data-rich-image-url]').value = reqPng2;
    reqImgPanel.querySelector('[data-rich-image-alt]').value = 'New alt';
    reqImgPanel.querySelector('[data-rich-image-width]').value = '120';
    reqImgPanel.querySelector('[data-rich-image-align]').value = 'center';
    reqImgPanel.querySelector('[data-rich-image-apply]').click();
    await wait(300);
    var reqImgDoc = JSON.parse(__SUI_RICH.get('editor')[0]).children[0].children[0];
    if (reqImgDoc.alt !== 'New alt' || reqImgDoc.width !== '120px' || reqImgDoc.align !== 'center') throw new Error('image settings not stored: ' + JSON.stringify(reqImgDoc));
    if (bsEditor().querySelector('img').getAttribute('alt') !== 'New alt') throw new Error('image settings not applied to the DOM');
    if (bsEditor().querySelector('img').getAttribute('src') !== reqPng2) throw new Error('image replacement URL not applied');

    __SUI_RICH.set('editor', { type: 'doc', children: [{ type: 'table', header: true, rows: [{ type: 'table_row', cells: [{ type: 'table_cell', header: true, children: [{ type: 'text', text: 'Cell', marks: [] }] }] }] }] });
    await wait(300);
    var reqCellP = bsEditor().querySelector('th, td');
    var reqCellRect = reqCellP.getBoundingClientRect();
    var reqCellDown = new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqCellRect.left + reqCellRect.width / 2, clientY: reqCellRect.top + reqCellRect.height / 2 });
    reqCellP.dispatchEvent(reqCellDown);
    if (reqCellDown.defaultPrevented) throw new Error('cell text mousedown blocks native text selection');
    if (bsEditor().querySelector('th.sui-rich-cell-selected, td.sui-rich-cell-selected')) throw new Error('text mousedown selects the whole cell');
    var reqCell = bsEditor().querySelector('th, td');
    var reqPaddingDown = new MouseEvent('mousedown', { bubbles: true, cancelable: true, clientX: reqCellRect.left + 1, clientY: reqCellRect.top + 1 });
    reqCell.dispatchEvent(reqPaddingDown);
    if (!bsEditor().querySelector('th.sui-rich-cell-selected, td.sui-rich-cell-selected')) throw new Error('cell padding does not select the cell');

    var reqFileBtn = richRoot().querySelector('.sui-rich-menu-btn[data-rich-menu="file"]');
    reqFileBtn.click();
    await wait(100);
    var reqPanel = richRoot().querySelector('.sui-rich-menu-panel');
    var reqOrder = [].map.call(reqPanel.children, function (n) { return (n.textContent || '').trim(); });
    var reqOpen = reqOrder.findIndex(function (t) { return t.indexOf('Open file') === 0; });
    var reqSave = reqOrder.findIndex(function (t) { return t.indexOf('Save as') === 0; });
    var reqPreview = reqOrder.findIndex(function (t) { return t.indexOf('Print preview') === 0; });
    var reqPrint = reqOrder.findIndex(function (t) { return t.indexOf('Print / Save') === 0; });
    if (!(reqOpen >= 0 && reqSave === reqOpen + 1 && reqPreview > reqSave && reqPrint > reqPreview)) throw new Error('File menu order: ' + reqOrder.join(' | '));
    reqFileBtn.click();
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

// TestRichTextDefaultToolbarCoversTextFeatures keeps the toolbar in sync with
// the text commands: every formatting feature must expose a control, including
// inline code and unlink, and the block select must offer Heading 4.
func TestRichTextDefaultToolbarCoversTextFeatures(t *testing.T) {
	const marker = "var defaultToolbar = '"
	i := strings.Index(suiRichTextRuntime, marker)
	if i < 0 {
		t.Fatal("defaultToolbar not found in the runtime")
	}
	rest := suiRichTextRuntime[i+len(marker):]
	end := strings.Index(rest, "'")
	if end < 0 {
		t.Fatal("defaultToolbar assignment not closed")
	}
	items := map[string]bool{}
	for _, group := range strings.Split(rest[:end], "|") {
		for _, token := range strings.Split(group, ",") {
			items[strings.TrimSpace(token)] = true
		}
	}
	for _, token := range []string{
		"undo", "redo", "copy", "cut", "paste",
		"styleSelect", "fontFamily", "fontSize", "fontSizeUp", "fontSizeDown", "bold", "italic", "underline",
		"strike", "textColor", "highlight", "superscript", "subscript", "code",
		"clearFormat", "alignLeft", "alignCenter", "alignRight", "justify",
		"indent", "outdent", "paragraphSettings",
		"bulletList", "orderedList", "blockquote", "codeBlock",
		"link", "unlink", "image", "table", "symbol", "hr", "pageBreak",
		"pageSettings", "printPreview", "print", "find", "copyFormat", "zoom",
	} {
		if !items[token] {
			t.Errorf("default toolbar is missing the %q control", token)
		}
	}
	for _, want := range []string{
		"code: '<svg",
		"unlink: '<svg",
		"copyFormat: '<svg",
		"printPreview: '<svg",
		"['portrait', 'Vertical']",
		"['landscape', 'Horizontal']",
		"['h4', 'Heading 4']",
		"indent: 'Increase paragraph indent'",
		"data-rich-paragraph-indent-inc",
		"insert: ['Link|link', 'Image|image', 'Table|table'",
		"var MENU_LABELS = { file: 'File', edit: 'Edit', insert: 'Insert', preferences: 'Preferences' };",
		"preferences: ['Document settings|documentSettings', 'Page settings|pageSettings', 'Header and footer...|headerFooter'",
		"{ label: 'Icon style', items: ['Line icons|iconSetLine'",
		"'Soft block icons|iconSetBlocks'",
	} {
		if !strings.Contains(suiRichTextRuntime, want) {
			t.Errorf("rich text runtime is missing %q", want)
		}
	}
	for _, unwanted := range []string{"Unlink|unlink", "Image settings...|imageSettings", "tools: [", "format: [", "view: ["} {
		if strings.Contains(suiRichTextRuntime, unwanted) {
			t.Errorf("rich text runtime still contains redundant menu entry %q", unwanted)
		}
	}
}
