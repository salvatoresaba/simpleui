package main

import (
	"strings"
	"testing"
)

const treeTestSource = `state selected = ""
	state renameKey = ""
	state renameValue = ""
fn changed() -> null {
}
fn renamed() -> null {
}
view main {
	tree "Files" name: "files" value: selected onChange: changed editable: true editableDepth: 2 renameKey: renameKey renameValue: renameValue renameIcon: "E" showRenameIcon: false onRename: renamed style: cssSafe("--sui-tree-gap: 1px; --sui-tree-selected-bg: #3949d6;") {
		treeitem "src" key: "src" open: true style: cssSafe("--sui-tree-row-bg: #1a253b;") {
			treeitem "main.go" key: "main" showRenameIcon: false
			treeitem "util.go" key: "util"
		}
		treeitem "README.md" key: "readme" renameIcon: "R"
	}
}
`

func TestTreeRender(t *testing.T) {
	prog, dl := ParseSource(treeTestSource, "tree.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-tree"`,
		`data-sui-tree="1"`,
		`role="tree"`,
		`data-sui-tree-setter="setSelected"`,
		`data-sui-tree-onchange="changed"`,
		`data-sui-tree-editable="1"`,
		`data-sui-tree-editable-depth="2"`,
		`data-sui-tree-rename-key-setter="setRenameKey"`,
		`data-sui-tree-rename-value-setter="setRenameValue"`,
		`data-sui-tree-name="files"`,
		`data-sui-tree-rename-icon="E"`,
		`data-sui-tree-show-rename-icon="0"`,
		`data-sui-show-rename-icon="0"`,
		`data-sui-rename-icon="R"`,
		`--sui-tree-row-bg: #1a253b`,
		`data-sui-tree-onrename="renamed"`,
		`data-sui-editable="inherit"`,
		`data-sui-key="src"`,
		`data-sui-key="main"`,
		`aria-level="2"`,
		`data-sui-open="1"`,
		`sui-tree-toggle`,
		`sui-tree-value`,
		`--sui-tree-gap: 1px`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-tree"`,
		`data-sui-tree="1"`,
		`data-sui-tree-editable="1"`,
		`data-sui-tree-editable-depth="2"`,
		`data-sui-tree-name="files"`,
		`data-sui-tree-rename-icon="E"`,
		`data-sui-tree-show-rename-icon="0"`,
		`data-sui-show-rename-icon="0"`,
		`role="tree"`,
		`data-sui-key="src"`,
		`data-sui-editable="inherit"`,
		`role="treeitem"`,
		`data-sui-open="1"`,
		`Files`,
		`README.md`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
	css := GenerateCss(prog)
	for _, want := range []string{".sui-tree", ".sui-tree-row", ".sui-tree-item.selected", ".sui-tree-toggle", "var(--sui-tree-panel-bg", "var(--sui-tree-selected-bg"} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}
}

func TestTreeDataAPI(t *testing.T) {
	for _, full := range []string{
		"tree.importJson", "tree.importXml", "tree.importJsonFile", "tree.importXmlFile",
		"tree.exportJson", "tree.exportXml", "tree.downloadJson", "tree.downloadXml", "tree.clear",
	} {
		if _, ok := StdLookup(full); !ok {
			t.Errorf("missing core tree intrinsic %s", full)
		}
	}
	if !IsCoreModule("tree") {
		t.Fatal("tree must be a core module")
	}
	for _, want := range []string{
		"__SUI_TREE", "importJsonFile", "exportXml", "downloadText",
		"data-sui-tree-name", "DOMParser", "parseXml", "clearTreeStorage",
	} {
		if !strings.Contains(suiTreeRuntime, want) {
			t.Errorf("tree runtime missing %q", want)
		}
	}
}

func TestTreeDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(treeTestSource, "tree.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"kind":"tree"`,
		`"kind":"treeitem"`,
		`"name":"files"`,
		`"key":"src"`,
		`"open":true`,
		`"value":"selected"`,
		`"editable":true`,
		`"editableDepth":"2"`,
		`"renameKey":"renameKey"`,
		`"renameValue":"renameValue"`,
		`"onRename":"renamed"`,
		`"renameIcon":"E"`,
		`"renameIcon":"R"`,
		`"showRenameIcon":false`,
		`--sui-tree-row-bg: #1a253b`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
	src := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{
		`tree "Files"`,
		`name: "files"`,
		`treeitem "src"`,
		`key: "src"`,
		`open: true`,
		`value: selected`,
		`onChange: changed`,
		`editable: true`,
		`editableDepth: 2`,
		`renameKey: renameKey`,
		`renameValue: renameValue`,
		`renameIcon: "E"`,
		`renameIcon: "R"`,
		`showRenameIcon: false`,
		`onRename: renamed`,
		`style: cssSafe`, `--sui-tree-gap: 1px`,
		`--sui-tree-row-bg: #1a253b`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{"treeControls", "treeItemControls", "initially open", "stable item key", "editable depth", "rename key state", "read-only", "TREE_STYLE_PRESETS", "style preset", "--sui-tree-panel-bg", "TREE_ITEM_STYLE_PRESETS", "node style preset", "show rename icon", "rename icon", "tree name", "tree.importJson", "treeImportData", "treeDownloadData", "Import JSON", "Export XML", `"date", "time", "tree", "treeitem"`, `tree: '<rect`} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer tree support missing %q", want)
		}
	}
}

func TestTreeTextIsEscaped(t *testing.T) {
	src := "view main {\n\ttree \"<b>Files</b>\" {\n\t\ttreeitem \"<i>src</i>\" key: \"src\"\n\t}\n}\n"
	prog, dl := ParseSource(src, "tree.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	if strings.Contains(html, "<b>Files</b>") || strings.Contains(html, "<i>src</i>") {
		t.Errorf("tree text must be escaped")
	}
	if !strings.Contains(html, "&lt;b&gt;Files&lt;/b&gt;") || !strings.Contains(html, "&lt;i&gt;src&lt;/i&gt;") {
		t.Errorf("escaped tree text missing")
	}
}
