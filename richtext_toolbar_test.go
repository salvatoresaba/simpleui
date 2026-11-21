package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRichTextToolbarGroupsFromRuntime(t *testing.T) {
	groups := richTextToolbarGroups()
	if len(groups) < 6 {
		t.Fatalf("expected the default toolbar groups, got %d", len(groups))
	}
	ids := map[string]bool{}
	for _, group := range groups {
		if group.Name == "" {
			t.Fatal("toolbar group has no label")
		}
		if len(group.Items) == 0 {
			t.Fatalf("toolbar group %q has no items", group.Name)
		}
		for _, item := range group.Items {
			if item.ID == "" || item.Label == "" {
				t.Fatalf("toolbar item has an empty id or label: %+v", item)
			}
			ids[item.ID] = true
		}
	}
	for _, id := range []string{"undo", "styleSelect", "bold", "paragraphSettings", "link", "pageSettings", "zoom"} {
		if !ids[id] {
			t.Errorf("default toolbar item %q not exposed to the designer", id)
		}
	}
	for _, id := range []string{"lineHeight", "spaceBefore", "spaceAfter", "imageSettings", "documentSettings", "sourceMode"} {
		if !ids[id] {
			t.Errorf("extra toolbar item %q not exposed to the designer", id)
		}
	}
}

func TestRichTextToolbarDesignerInjection(t *testing.T) {
	raw := richTextToolbarJSON()
	var groups []richTextToolbarGroup
	if err := json.Unmarshal([]byte(raw), &groups); err != nil {
		t.Fatalf("toolbar JSON is invalid: %v (%s)", err, raw)
	}
	if len(groups) == 0 {
		t.Fatal("toolbar JSON has no groups")
	}
	if !strings.Contains(guiPage, "window.SUI_RICH_TOOLBAR_GROUPS="+raw+";") {
		t.Fatal("designer page does not inject the toolbar groups")
	}
	if !strings.Contains(guiPage, `"find","sourceMode"`) {
		t.Fatal("designer toolbar picker does not allow sourceMode in markup mode")
	}
	if !strings.Contains(suiRichTextRuntime, "sourceMode: 'Source code'") {
		t.Fatal("runtime does not label the sourceMode toolbar button")
	}
	if !strings.Contains(suiRichTextRuntime, "function noFormatToolbar(inst)") {
		t.Fatal("runtime does not implement toolbarItems: none")
	}
	if !strings.Contains(suiRichTextRuntime, "if (allowedRegistered.length && !noToolbar)") {
		t.Fatal("toolbarItems: none still shows plugin toolbar buttons")
	}
	for _, marker := range []string{
		"if (noFormatToolbar(inst)) return;",
		"if (noFormatToolbar(inst) || !commands.length) { bar.hidden = true; return; }",
	} {
		if !strings.Contains(suiRichTextRuntime, marker) {
			t.Errorf("runtime missing none guard: %s", marker)
		}
	}
}

func TestRichTextThemeResolvedBeforeToolbarBuild(t *testing.T) {
	runtime := suiRichTextRuntime
	theme := strings.Index(runtime, "var explicitTheme =")
	toolbar := strings.Index(runtime, "buildToolbar(inst); buildMenu(inst)")
	if theme < 0 || toolbar < 0 || theme > toolbar {
		t.Fatal("initial theme must be resolved before buildToolbar overwrites the theme attribute")
	}
}

func richtextTestElement(t *testing.T, src string) (*Prog, *Element) {
	t.Helper()
	prog, dl := ParseSource(src, "richtext_toolbar_test.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	var find func([]ViewNode) *Element
	find = func(nodes []ViewNode) *Element {
		for i := range nodes {
			n := &nodes[i]
			if n.kind == "element" && n.elem != nil {
				if n.elem.tag == "richtext" {
					return n.elem
				}
				if found := find(n.elem.children); found != nil {
					return found
				}
			}
			for _, branch := range [][]*ViewNode{n.body, n.then_, n.else_} {
				for j := range branch {
					if found := find([]ViewNode{*branch[j]}); found != nil {
						return found
					}
				}
			}
		}
		return nil
	}
	for i := range prog.views {
		if found := find(prog.views[i].body); found != nil {
			return prog, found
		}
	}
	t.Fatal("richtext element not found")
	return nil, nil
}

func TestRichTextThemeAndDesignerRoundTrip(t *testing.T) {
	prog, elem := richtextTestElement(t, `state doc = ""
view main {
	richtext "Article" name: "editor" value: doc theme: "light" contentTheme: "dark" toolbarItems: "undo,bold|link"
}`)
	js := GenerateJs(prog)
	for _, marker := range []string{
		`data-sui-rich-theme="light" data-sui-rich-theme-explicit="1" data-sui-rich-content="light"`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS missing %s", marker)
		}
	}
	if strings.Contains(js, `data-sui-rich-content="dark"`) {
		t.Error("generated JS still emits a dark document page")
	}
	model := guiElementFrom(elem, map[string]bool{})
	for key, want := range map[string]string{
		"theme":        "light",
		"contentTheme": "dark",
		"toolbarItems": "undo,bold|link",
	} {
		if got := model.obj[key].s; got != want {
			t.Errorf("designer model %s = %q, want %q", key, got, want)
		}
	}

	prog, _ = richtextTestElement(t, `state doc = ""
view main {
	richtext "Article" name: "editor" value: doc contentTheme: "dark"
}`)
	js = GenerateJs(prog)
	if !strings.Contains(js, `data-sui-rich-theme="dark"`) {
		t.Error("default theme was not emitted")
	}
	if strings.Contains(js, `data-sui-rich-theme="dark" data-sui-rich-theme-explicit="1"`) {
		t.Error("theme flag emitted without a theme attribute")
	}
	if strings.Contains(js, `data-sui-rich-content="dark"`) {
		t.Error("contentTheme still changes the generated paper")
	}
}

func TestRichTextModeDesignerRoundTrip(t *testing.T) {
	prog, elem := richtextTestElement(t, `state doc = ""
view main {
	richtext name: "page" mode: "markup" value: doc
}`)
	js := GenerateJs(prog)
	if !strings.Contains(js, `data-sui-rich-mode="markup"`) {
		t.Error("generated JS lost the markup mode")
	}
	model := guiElementFrom(elem, map[string]bool{})
	if got := model.obj["mode"].s; got != "markup" {
		t.Errorf("designer model mode = %q, want markup", got)
	}
}

func TestRichTextInspectorIsMinimal(t *testing.T) {
	start := strings.Index(guiHTML, "function richtextControls(n) {")
	end := strings.Index(guiHTML, "function inputControls(nodes) {")
	if start < 0 || end <= start {
		t.Fatal("richtextControls function not found in the designer")
	}
	src := guiHTML[start:end]
	for _, marker := range []string{
		`rowInput("header"`,
		`rowInput("footer"`,
		`"Status bar"`,
		`"Selection toolbar"`,
		`className = "hint"`,
	} {
		if strings.Contains(src, marker) {
			t.Errorf("richtext Inspector still renders %s", marker)
		}
	}
	for _, marker := range []string{
		`"Editor chrome theme"`,
		`toolbarBtn.title = "Choose the toolbar buttons"`,
		`hyphBtn.title = "Choose the embedded hyphenation dictionaries"`,
		`roSwitch.title = "Show the document without editing controls"`,
		`mkSwitch(n.readonly === undefined ? false : !!n.readonly`,
		`nameInput.value = n.name || ""`,
		`nameInput.title = "Runtime name used by richtext.*"`,
		`upd("name", nameInput.value.trim())`,
		`[["richtext", "Rich text editor"], ["html", "HTML editor"], ["markup", "Markup editor"]]`,
	} {
		if !strings.Contains(src, marker) {
			t.Errorf("richtext Inspector missing %s", marker)
		}
	}
}

func TestRichTextEditorModesE2E(t *testing.T) {
	src := `state htmlDoc = ""
state mdDoc = ""
state viewerDoc = ""
view main {
	richtext name: "html" mode: "html" value: htmlDoc height: "240px" theme: "dark" header: "HTML header" footer: "HTML footer"
	richtext name: "markup" mode: "markup" value: mdDoc height: "240px" theme: "dark" header: "MD header" footer: "MD footer" toolbarItems: "bold,link,documentSettings,sourceMode"
	richtext name: "viewer" mode: "html" readonly: true value: viewerDoc height: "160px" theme: "dark"
}`
	script := featureScriptHead + `
    function root(name) { return document.querySelector('[data-sui-rich-name="' + name + '"]'); }
    function cmd(name, token) { return root(name).querySelector('[data-rich-cmd="' + token + '"]'); }
    await wait(350);
    __SUI_RICH.set('html', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'HTML body', marks: [] }] }] });
    __SUI_RICH.set('markup', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Markup body', marks: [] }] }] });
    __SUI_RICH.set('viewer', { type: 'doc', children: [{ type: 'paragraph', children: [{ type: 'text', text: 'Viewer body', marks: [] }] }] });
    await wait(450);
    if (root('html').getAttribute('data-sui-rich-mode') !== 'html') throw new Error('HTML mode attribute missing');
    if (root('markup').getAttribute('data-sui-rich-mode') !== 'markup') throw new Error('markup mode attribute missing');
    if (cmd('html', 'pageBreak')) throw new Error('HTML profile still offers page break');
    if (!cmd('html', 'bold') || !cmd('html', 'documentSettings') || !cmd('html', 'pageSettings')) throw new Error('HTML profile is missing basic tools');
    if (cmd('markup', 'textColor') || cmd('markup', 'alignLeft') || cmd('markup', 'pageBreak')) throw new Error('markup profile still offers rich-text tools');
    if (!cmd('markup', 'bold') || !cmd('markup', 'link') || !cmd('markup', 'documentSettings')) throw new Error('markup profile is missing basic tools');
    if (!cmd('markup', 'sourceMode')) throw new Error('markup toolbar does not expose sourceMode');
    if (cmd('markup', 'sourceMode').closest('[role="group"]').getAttribute('aria-label') !== 'Document') throw new Error('sourceMode is not grouped under Document');
    if (!root('html').querySelector('[data-rich-cmd="saveHTML"]') || root('html').querySelector('[data-rich-cmd="saveDOCX"]') || root('html').querySelector('[data-rich-cmd="headerFooter"]')) throw new Error('HTML file menu is not restricted');
    if (!root('markup').querySelector('[data-rich-cmd="saveMarkdown"]') || root('markup').querySelector('[data-rich-cmd="saveHTML"]') || root('markup').querySelector('[data-rich-cmd="headerFooter"]')) throw new Error('markup file menu is not restricted');
    var htmlSourceBtn = root('html').querySelector('[data-rich-source-toggle]');
    if (!htmlSourceBtn) throw new Error('HTML source toggle missing');
    htmlSourceBtn.click();
    await wait(120);
    var htmlSource = root('html').querySelector('.sui-rich-source');
    if (!htmlSource || htmlSource.hidden) throw new Error('HTML source panel did not open');
    var htmlArea = htmlSource.querySelector('.sui-rich-source-code');
    if (htmlArea.value.indexOf('HTML body') < 0) throw new Error('HTML source did not reflect the document');
    var gutter = htmlSource.querySelector('.sui-rich-source-gutter');
    var mirror = htmlSource.querySelector('.sui-rich-source-mirror');
    if (!gutter || !mirror || gutter.textContent.trim() !== '1') throw new Error('source gutter is missing or wrong');
    if (mirror.textContent !== htmlArea.value) throw new Error('source highlight mirror does not match the editor');
    htmlArea.value = 'alpha\nbeta';
    htmlArea.dispatchEvent(new Event('input', { bubbles: true }));
    if (gutter.textContent !== '1\n2' || mirror.textContent !== 'alpha\nbeta') throw new Error('source gutter/mirror did not follow edits');
    htmlArea.focus();
    htmlArea.setSelectionRange(0, 11);
    var tabKey = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true });
    htmlArea.dispatchEvent(tabKey);
    if (!tabKey.defaultPrevented || htmlArea.value !== '  alpha\n  beta') throw new Error('Tab did not indent source lines: ' + htmlArea.value);
    var shiftTabKey = new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true });
    htmlArea.dispatchEvent(shiftTabKey);
    if (!shiftTabKey.defaultPrevented || htmlArea.value !== 'alpha\nbeta') throw new Error('Shift+Tab did not outdent source lines: ' + htmlArea.value);
    var unsafeSource = '<script>window.__SUI_SCRIPT_RAN=1<\/script><p>safe</p>';
    htmlArea.value = unsafeSource;
    htmlArea.dispatchEvent(new Event('input', { bubbles: true }));
    if (htmlSource.querySelector('script') || window.__SUI_SCRIPT_RAN) throw new Error('source mirror executed raw HTML');
    if (mirror.textContent !== unsafeSource) throw new Error('source mirror did not escape raw HTML');
    htmlArea.value = '<h2>From source</h2><p>External <b>HTML</b></p>';
    htmlArea.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(650);
    if (root('html').querySelector('.sui-rich-editor').textContent.indexOf('From source') < 0) throw new Error('HTML source was not applied');
    root('html').querySelector('[data-rich-source-toggle]').click();
    await wait(120);
    if (!htmlSource.hidden) throw new Error('HTML source panel did not close');
    cmd('markup', 'sourceMode').click();
    await wait(120);
    var mdSource = root('markup').querySelector('.sui-rich-source');
    if (!mdSource || mdSource.hidden || mdSource.querySelector('.sui-rich-source-label').textContent !== 'Markdown source') throw new Error('sourceMode toolbar button did not open the Markdown source editor');
    var mdArea = mdSource.querySelector('.sui-rich-source-code');
    mdArea.value = '# From markdown\n\n**external** code';
    mdArea.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(650);
    if (root('markup').querySelector('.sui-rich-editor').textContent.indexOf('From markdown') < 0) throw new Error('Markdown source was not applied');
    root('markup').querySelector('[data-rich-source-toggle]').click();
    await wait(120);
    if (root('viewer').querySelector('[data-rich-source-toggle]')) throw new Error('read-only viewer still offers source editing');

    // Incompatible clipboard content is converted to the profile source format.
    function pasteInto(name, kind, payload) {
      var editor = root(name).querySelector('.sui-rich-editor');
      editor.focus();
      var ev = new Event('paste', { bubbles: true, cancelable: true });
      Object.defineProperty(ev, 'clipboardData', { value: { getData: function (type) {
        if (kind === 'html' && type === 'text/html') return payload;
        if (kind === 'rtf' && type === 'text/rtf') return payload;
        if (type === 'text/plain') return kind === 'plain' ? payload : '';
        return '';
      } } });
      editor.dispatchEvent(ev);
      return ev.defaultPrevented;
    }
    if (!pasteInto('html', 'plain', '# Pasted heading\n\n**Pasted bold**')) throw new Error('HTML editor did not handle plain paste');
    await wait(260);
    var htmlAfterPaste = __SUI_RICH.get('html')[0];
    if (htmlAfterPaste.indexOf('heading') < 0 || htmlAfterPaste.indexOf('Pasted heading') < 0) throw new Error('Markdown was not converted to HTML: ' + htmlAfterPaste);
    var htmlStatus = __SUI_RICH.status('html')[0];
    if (htmlStatus.indexOf('"detected":"markdown"') < 0 || htmlStatus.indexOf('"target":"html"') < 0) throw new Error('HTML paste import was not recorded: ' + htmlStatus);
    var htmlNotice = root('html').querySelector('.sui-rich-notice');
    if (!htmlNotice || htmlNotice.hidden || htmlNotice.textContent.indexOf('converted to HTML') < 0) throw new Error('HTML conversion notice missing');

    if (!pasteInto('html', 'rtf', '{\\rtf1\\ansi Rich \\b bold\\b0}')) throw new Error('HTML editor did not handle RTF paste');
    await wait(260);
    if (root('html').querySelector('.sui-rich-editor').innerText.indexOf('Rich bold') < 0) throw new Error('RTF paste was not imported');
    var rtfNotice = root('html').querySelector('.sui-rich-notice');
    if (!rtfNotice || rtfNotice.getAttribute('data-notice-kind') !== 'warning' || rtfNotice.textContent.indexOf('best-effort') < 0) throw new Error('RTF lossy warning missing');

    if (!pasteInto('markup', 'html', '<h2>Pasted HTML</h2><p><strong>Strong</strong></p>')) throw new Error('Markup editor did not handle HTML paste');
    await wait(260);
    var markupAfterPaste = __SUI_RICH.get('markup')[0];
    if (markupAfterPaste.indexOf('Pasted HTML') < 0 || markupAfterPaste.indexOf('Strong') < 0) throw new Error('HTML was not converted to Markdown: ' + markupAfterPaste);
    var markupStatus = __SUI_RICH.status('markup')[0];
    if (markupStatus.indexOf('"detected":"html"') < 0 || markupStatus.indexOf('"target":"markdown"') < 0) throw new Error('Markup paste import was not recorded: ' + markupStatus);
    var markupNotice = root('markup').querySelector('.sui-rich-notice');
    if (!markupNotice || markupNotice.hidden || markupNotice.getAttribute('data-notice-kind') !== 'warning' || markupNotice.textContent.indexOf('simplified to Markdown') < 0) throw new Error('Markdown lossy warning missing');

    __SUI_RICH.command('html', 'documentSettings');
    await wait(150);
    var docPanel = document.querySelector('.sui-rich-documentsettings[data-rich-documentsettings="html"]');
    if (!docPanel || !docPanel.querySelector('[data-rich-doc-background]')) throw new Error('document background setting missing');
    if (docPanel.querySelector('[data-rich-open-chrome]')) throw new Error('HTML settings still show header/footer');
    __SUI_RICH.command('markup', 'pageSettings');
    await wait(150);
    var pagePanel = document.querySelector('.sui-rich-pagesettings[data-rich-pagesettings="markup"]');
    if (!pagePanel || !pagePanel.querySelector('[data-rich-page-background]')) throw new Error('page background setting missing');
    if (pagePanel.querySelector('[data-rich-open-chrome]')) throw new Error('markup settings still show header/footer');
    docPanel.querySelector('[data-rich-doc-background]').value = '#112233';
    docPanel.querySelector('[data-rich-doc-apply]').click();
    await wait(200);
    var htmlBody = root('html').querySelector('.sui-rich-body');
    if (getComputedStyle(htmlBody).backgroundColor !== 'rgb(17, 34, 51)') throw new Error('document background was not applied: ' + getComputedStyle(htmlBody).backgroundColor);
    if (getComputedStyle(root('html').querySelector('.sui-rich-editor')).color !== 'rgb(245, 247, 250)') throw new Error('page text did not adapt to a dark background');
    if (__SUI_RICH.get('html')[0].indexOf('#112233') < 0) throw new Error('document background was not stored in the document JSON');
    __SUI_RICH.command('html', 'paginatedView');
    await wait(650);
    if (root('html').querySelector('.sui-rich-sheet-header, .sui-rich-sheet-footer, .sui-rich-pageheader, .sui-rich-pagefooter')) throw new Error('HTML mode still renders header/footer');
    __SUI_RICH.command('html', 'pageSettings');
    await wait(120);
    var htmlPagePanel = document.querySelector('.sui-rich-pagesettings[data-rich-pagesettings="html"]');
    htmlPagePanel.querySelector('[data-rich-page-number]').value = '1';
    htmlPagePanel.querySelector('[data-rich-page-background]').value = '#fff3cd';
    htmlPagePanel.querySelector('[data-rich-page-apply]').click();
    await wait(500);
    var firstSheet = root('html').querySelector('.sui-rich-page-sheet[data-rich-page="1"]');
    if (!firstSheet || getComputedStyle(firstSheet).backgroundColor !== 'rgb(255, 243, 205)') throw new Error('page background override was not applied');
    if (__SUI_RICH.get('html')[0].indexOf('#fff3cd') < 0) throw new Error('page background override was not stored in the document JSON');
    if (__SUI_RICH.saveAs('html', 'docx')[0] !== false) throw new Error('HTML mode saved a DOCX file');
    if (__SUI_RICH.saveAs('markup', 'html')[0] !== false) throw new Error('markup mode saved an HTML file');
    if (__SUI_RICH.exportFormat('html', 'json')[0] !== null) throw new Error('HTML mode exported JSON through exportFormat');
    if (__SUI_RICH.exportBytes('html', 'docx')[0] !== null) throw new Error('HTML mode exported DOCX bytes');
    if (__SUI_RICH.importFormat('markup', '# x', 'html')[0] !== false) throw new Error('markup mode imported HTML');
    __SUI_RICH.registerPlugin('modetest', { commands: { shout: { title: 'Shout', command: 'bold' } }, toolbar: [{ id: 'shout', label: 'Shout' }] });
    if (!root('html').querySelector('[data-rich-cmd="shout"]')) throw new Error('HTML mode hid an allowed plugin button');
    if (root('markup').querySelector('[data-rich-cmd="shout"]')) throw new Error('markup mode still shows a plugin button');
    if (__SUI_RICH.command('markup', 'shout')[0] !== false) throw new Error('markup mode ran a plugin command');
    var markupEditor = root('markup').querySelector('.sui-rich-editor');
    markupEditor.focus();
    var printKey = new KeyboardEvent('keydown', { key: 'p', ctrlKey: true, bubbles: true, cancelable: true });
    markupEditor.dispatchEvent(printKey);
    if (!printKey.defaultPrevented) throw new Error('markup mode did not swallow the print shortcut');
    var viewer = root('viewer');
    if (viewer.querySelector('.sui-rich-editor').getAttribute('contenteditable') !== 'false') throw new Error('read-only HTML is still editable');
    if (!viewer.querySelector('.sui-rich-toolbar').hidden || !viewer.querySelector('.sui-rich-menu').hidden) throw new Error('read-only HTML still shows editing chrome');
    if (!viewer.querySelector('.sui-rich-editor').innerText || viewer.querySelector('.sui-rich-editor').innerText.indexOf('Viewer body') < 0) throw new Error('read-only HTML is not a viewer');
` + featureScriptTail
	runHeadlessFeature(t, src, script)
}

func TestRichTextReadOnlyAndStatusLayout(t *testing.T) {
	for _, marker := range []string{
		"function isReadOnly(inst)",
		"function applyReadOnlyChrome(inst)",
		"applyReadOnlyChrome(inst);",
		"if (isReadOnly(inst)) return;",
		"inst.config.context && !isReadOnly(inst)",
		"root.style.setProperty('--sui-rich-content-height'",
		"if (contextBar && !boolAttr(inst.root, 'data-sui-rich-context', true)) contextBar.hidden = true;",
	} {
		if !strings.Contains(suiRichTextRuntime, marker) {
			t.Errorf("runtime missing read-only marker: %s", marker)
		}
	}
	if strings.Contains(suiRichTextRuntime, "['.sui-rich-context', 'data-sui-rich-context']") {
		t.Error("editable editors force the selection toolbar visible at init")
	}
	for _, marker := range []string{
		".sui-rich-body { display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0;",
		".sui-rich-status { display: flex; flex: none; align-items: center;",
		".sui-rich-status[hidden] { display: none; }",
		".sui-rich-menu[hidden] { display: none; }",
		".sui-rich-toolbar[hidden] { display: none; }",
	} {
		if !strings.Contains(suiRichTextCss, marker) {
			t.Errorf("editor CSS missing status layout marker: %s", marker)
		}
	}
	if strings.Contains(suiRichTextRuntime, "inst.editor.style.height = editorHeight;") {
		t.Error("editor still has a fixed inline height that overrides the content-height variable")
	}
}

func TestRichTextTitleIsMetadataOnly(t *testing.T) {
	js := generateJsFromInline(t, `
view main {
	richtext "Article" name: "editor"
}
`)
	if strings.Contains(js, `class="sui-rich-title"`) {
		t.Error("richtext still renders a visible title above the menu")
	}
	if !strings.Contains(js, `data-sui-rich-title="Article"`) {
		t.Error("richtext dropped the title metadata used by {title} tokens")
	}
}

func TestRichTextLoadContentWiring(t *testing.T) {
	js := generateJsFromInline(t, `
state doc = ""

async fn load() -> null {
	ok, err = richtext.loadContent("editor", "# Loaded from code", "markdown", "report.md")
}

view main {
	richtext name: "editor" value: doc
	button "Load" onClick: load
}
`)
	for _, marker := range []string{
		"richtext_loadcontent",
		"__SUI_RICH.loadContent(name, content, fmt, filename)",
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS missing loadContent marker: %s", marker)
		}
	}
}

func TestRichTextModeAttributeAndValidation(t *testing.T) {
	js := generateJsFromInline(t, `
state doc = ""
view main {
	richtext name: "page" mode: "html" value: doc
}`)
	if !strings.Contains(js, `data-sui-rich-mode="html"`) {
		t.Error("generated HTML missing the richtext editor mode")
	}

	prog, pdl := ParseSource(`view main {
	richtext name: "page" mode: "nonsense"
}`, "mode.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	if dl := Check(prog); !dl.HasErrors() || !strings.Contains(dl.String(), "richtext mode must be one of") {
		t.Fatalf("invalid mode was not rejected: %v", dl.String())
	}
}

func TestRichTextModeProfilesAndPageBackgroundMarkers(t *testing.T) {
	for _, marker := range []string{
		"var htmlToolbar =",
		"var markupToolbar =",
		"function toolbarForMode(inst, configured)",
		"function menusForMode(mode)",
		"function modeFileAccept(mode)",
		"function modeFileAllowed(mode, filename)",
		"function sniffRichTextFormat(",
		"function convertPastedDoc(",
		"function reportPastedImport(",
		"sui-rich-notice",
		"data-notice-kind",
		"sourceMode: 'Source code'",
		"function updateSourceGutter(",
		"function updateSourceMirror(",
		"function indentSourceSelection(",
		"sui-rich-source-gutter",
		"sui-rich-source-mirror",
		"function chromeEnabled(inst)",
		"function pageInkFor(background)",
		"function applyPageBackground(inst)",
		"data-rich-doc-background",
		"data-rich-page-background",
		"var sheetBg = cssColor(pageConfig(inst, pageNo || 1).background) || '#ffffff';",
		"sheet.style.backgroundColor = sheetBg;",
		`<w:background w:color="`,
		`fo:background-color="#`,
		`\\*\\background`,
		`pdfFill(pageBackground)`,
		"function exportPageSignature(",
		"function pageFromMm(",
		"function pageBgHex(",
		"function pageOverrideFromConfig(",
		"function rtfScanSections(",
		"function odtLayoutConfig(props)",
		"function sectionConfig(sect)",
		`<w:br w:type="page"/>`,
		`fo:break-before="page"`,
		"function exportPageSections(",
		"function rtfBackgroundShape(",
		"function docxSectionProps(",
		"function odtBreakStyle(",
		`<w:type w:val="nextPage"/>`,
		`<v:fill color="#`,
		`style:master-page-name="`,
		`\\sect`,
	} {
		if !strings.Contains(suiRichTextRuntime, marker) {
			t.Errorf("runtime missing editor-mode/background marker: %s", marker)
		}
	}
	if !strings.Contains(suiRichTextCss, ".sui-rich-page-sheet") {
		t.Error("page sheet styles missing")
	}
	if !strings.Contains(suiRichTextCss, ".sui-rich-source-editor") || !strings.Contains(suiRichTextCss, ".sui-rich-source-mirror .sui-code-tag") {
		t.Error("source editor gutter/highlight styles missing")
	}
}

func TestRichTextLoadContentAcceptsExistingFiles(t *testing.T) {
	for _, marker := range []string{
		"function looksLikeDocumentPath(value)",
		"async function readDesktopFile(path)",
		"/__bridge__/file/readbytes",
		"reading a local file path is available only in desktop builds",
		"return openFile(name)",
	} {
		if !strings.Contains(suiRichTextRuntime, marker) {
			t.Errorf("runtime missing file-loading marker: %s", marker)
		}
	}
}

func TestRichTextExampleDropsLegacyToggles(t *testing.T) {
	raw, err := os.ReadFile("examples/richtext_editor/app.sui")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"menu: true", "toolbar: true", "contextToolbar: true", "statusbar: true"} {
		if strings.Contains(string(raw), marker) {
			t.Errorf("example still declares %s", marker)
		}
	}
}
