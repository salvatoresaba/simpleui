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

func TestTreeDesignerUX(t *testing.T) {
	for _, want := range []string{
		`node.kind !== "tree" && !!(node.children`,
		`"Edit tree data..."`, `"Tree settings..."`, `openTreeDataEditor`,
		`function treeNodesFromText`, `function treeDataAddRoot`,
		`function treeDataImport`, `function treeDataExport`,
		`id="treedata"`, `id="treedata-rows"`,
		`--sui-tree-title-size`, `--sui-tree-title-position`,
		`--sui-tree-toggle-color`, `--sui-tree-border-width`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("tree designer support missing %q", want)
		}
	}
	if strings.Contains(guiHTML, `"tree", "treeitem", "chart"`) {
		t.Errorf("treeitem must not be offered as a separate insert control")
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
	for _, want := range []string{"treeControls", "treeItemControls", "initially open", "stable item key", "editable depth", "rename key state", "read-only", "TREE_STYLE_PRESETS", "style preset", "--sui-tree-panel-bg", "TREE_ITEM_STYLE_PRESETS", "node style preset", "show rename icon", "rename icon", "tree name", "tree.importJson", "treeImportData", "treeDownloadData", "Import JSON", "Export XML", `"date", "time", "tree", "chart"`, `tree: '<rect`} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer tree support missing %q", want)
		}
	}
}

func TestTreePhaseAFeatures(t *testing.T) {
	src := `state selected = ""
state checkedCount = 0
fn checked() -> null {
}
view main {
	tree "Files" name: "files" value: selected checkbox: true counts: true guides: true truncate: true rowHeight: 32 icon: "D" iconOpen: "O" onCheck: checked {
		treeitem "src" key: "src" open: true icon: "F" iconOpen: "O" badge: "12" badgeType: "info" meta: "2 KB" checked: true {
			treeitem "main.go" key: "main" icon: "f" badge: "new" badgeType: "success" meta: "1 KB" checked: true
		}
		treeitem "README.md" key: "readme" meta: "4 KB"
	}
}
`
	prog, dl := ParseSource(src, "tree_phase_a.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-tree-guides="1"`, `data-sui-tree-checkbox="1"`,
		`data-sui-tree-counts="1"`, `data-sui-tree-row-height="32"`,
		`data-sui-tree-icon="D"`, `data-sui-tree-icon-open="O"`,
		`data-sui-tree-oncheck="checked"`, `data-sui-icon="O"`,
		`data-sui-badge="12"`, `data-sui-badge-type="info"`,
		`data-sui-meta="2 KB"`, `data-sui-checked="1"`,
		`sui-tree-guide`, `sui-tree-check`, `aria-checked="true"`,
		`sui-tree-badge-info`, `class="sui-tree-meta"`, `title="src"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("phase A tree markup missing %q", want)
		}
	}
	for _, want := range []string{
		`sui-tree-check`, `aria-checked`, `sui.tree.checked.`,
		`recomputeUp`, `setCheckedDeep`, `sui-tree-guide`,
		`MAX_BADGE`, `badgeType must be neutral`,
	} {
		if !strings.Contains(suiTreeRuntime, want) {
			t.Errorf("phase A runtime missing %q", want)
		}
	}
	for _, want := range []string{
		`--sui-tree-row-height`, `--sui-tree-guide-color`,
		`--sui-tree-badge-bg`, `--sui-tree-check-on`,
		`.sui-tree-badge-info`, `.sui-tree-meta`,
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("phase A tree CSS missing %q", want)
		}
	}
}

func TestTreePhaseABadges(t *testing.T) {
	bad := "view main {\n\ttreeitem \"Bad\" badgeType: \"loud\"\n}\n"
	prog, dl := ParseSource(bad, "tree_badge.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); !cd.HasErrors() {
		t.Fatalf("invalid badgeType must be rejected")
	}
}

func TestTreePhaseADesigner(t *testing.T) {
	src := `view main {
	tree "Files" guides: false truncate: false checkbox: true counts: true rowHeight: 30 icon: "D" iconOpen: "O" onCheck: checked {
		treeitem "src" icon: "F" iconOpen: "O" badge: "2" badgeType: "warning" meta: "1 KB" checked: true
	}
}
fn checked() -> null {
}
`
	prog, dl := ParseSource(src, "tree_designer.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"guides":false`, `"truncate":false`, `"checkbox":true`, `"counts":true`,
		`"rowHeight":"30"`, `"icon":"D"`, `"iconOpen":"O"`, `"onCheck":"checked"`,
		`"badge":"2"`, `"badgeType":"warning"`, `"meta":"1 KB"`, `"checked":true`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
	out := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{
		`guides: false`, `truncate: false`, `checkbox: true`, `counts: true`,
		`rowHeight: 30`, `icon: "D"`, `iconOpen: "O"`, `onCheck: checked`,
		`badge: "2"`, `badgeType: "warning"`, `meta: "1 KB"`, `checked: true`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("designer source missing %q:\n%s", want, out)
		}
	}
	for _, want := range []string{
		`"indent guides"`, `"hierarchical checkboxes"`, `"child counts"`,
		`"default icon"`, `"open icon"`, `"badge type"`, `"metadata"`,
		`--sui-tree-guide-color`, `.td-row .td-icon`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer phase A support missing %q", want)
		}
	}
}

func TestTreeTextIsEscaped(t *testing.T) {
	src := "view main {\n\ttree \"<b>Files</b>\" {\n\t\ttreeitem \"<i>src</i>\" key: \"src\" icon: \"<img src=x onerror=alert(1)>\" badge: \"<b>12</b>\" meta: \"<i>meta</i>\"\n\t}\n}\n"
	prog, dl := ParseSource(src, "tree.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	for _, raw := range []string{"<b>Files</b>", "<i>src</i>", "<img src=x", "<b>12</b>", "<i>meta</i>"} {
		if strings.Contains(html, raw) {
			t.Errorf("tree text must be escaped: %q", raw)
		}
	}
	for _, escaped := range []string{"&lt;b&gt;Files&lt;/b&gt;", "&lt;i&gt;src&lt;/i&gt;", "&lt;img src=x", "&lt;b&gt;12&lt;/b&gt;", "&lt;i&gt;meta&lt;/i&gt;"} {
		if !strings.Contains(html, escaped) {
			t.Errorf("escaped tree text missing %q", escaped)
		}
	}
}
