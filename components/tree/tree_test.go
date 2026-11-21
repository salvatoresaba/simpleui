package tree_test

import . "example.com/simpleui"

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
	tree "Files" name: "files" value: selected onChange: changed editMode: "editable" editDepth: 2 renameKey: renameKey renameValue: renameValue renameIcon: "E" showRenameIcon: false onRename: renamed style: cssSafe("--sui-tree-gap: 1px; --sui-tree-selected-bg: #3949d6;") {
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
		`data-sui-tree-edit-mode="editable"`,
		`data-sui-tree-edit-depth="2"`,
		`data-sui-tree-rename-key-setter="setRenameKey"`,
		`data-sui-tree-rename-value-setter="setRenameValue"`,
		`data-sui-tree-name="files"`,
		`data-sui-tree-rename-icon="E"`,
		`data-sui-tree-show-rename-icon="0"`,
		`data-sui-show-rename-icon="0"`,
		`data-sui-rename-icon="R"`,
		`--sui-tree-row-bg: #1a253b`,
		`data-sui-tree-onrename="renamed"`,
		`data-sui-rename="inherit"`,
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
		`data-sui-tree-edit-mode="editable"`,
		`data-sui-tree-edit-depth="2"`,
		`data-sui-tree-name="files"`,
		`data-sui-tree-rename-icon="E"`,
		`data-sui-tree-show-rename-icon="0"`,
		`data-sui-show-rename-icon="0"`,
		`role="tree"`,
		`data-sui-key="src"`,
		`data-sui-rename="inherit"`,
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
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("tree designer support missing %q", want)
		}
	}
	if strings.Contains(GuiHTML, `"tree", "treeitem", "chart"`) {
		t.Errorf("treeitem must not be offered as a separate insert control")
	}
}

func TestTreeDataAPI(t *testing.T) {
	for _, full := range []string{
		"tree.importJson", "tree.importXml", "tree.importJsonFile", "tree.importXmlFile",
		"tree.exportJson", "tree.exportXml", "tree.downloadJson", "tree.downloadXml", "tree.clear",
		"tree.expandAll", "tree.collapseAll", "tree.expandTo", "tree.setChildren", "tree.setChildrenError",
		"tree.retryChildren", "tree.lastMove", "tree.undoMove",
		"tree.setColumns", "tree.getColumns", "tree.moveColumn", "tree.setCell",
		"tree.selectedKeys", "tree.checkedKeys", "tree.search",
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
		if !strings.Contains(TreeRuntime, want) {
			t.Errorf("tree runtime missing %q", want)
		}
	}
}

func TestTreeDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(treeTestSource, "tree.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	model := StringifyJson(GuiModelJSON(prog))
	for _, want := range []string{
		`"kind":"tree"`,
		`"kind":"treeitem"`,
		`"name":"files"`,
		`"key":"src"`,
		`"open":true`,
		`"value":"selected"`,
		`"editMode":"editable"`,
		`"editDepth":"2"`,
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
	src := ViewNodesSource(prog.Views()[0].Body(), 1)
	for _, want := range []string{
		`tree "Files"`,
		`name: "files"`,
		`treeitem "src"`,
		`key: "src"`,
		`open: true`,
		`value: selected`,
		`onChange: changed`,
		`editMode: "editable"`,
		`editDepth: 2`,
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
	for _, want := range []string{"treeControls", "Edit depth", "Rename key state", "read-only", "TREE_STYLE_PRESETS", "style preset", "--sui-tree-panel-bg", "TREE_ITEM_STYLE_PRESETS", "Runtime name", "Tooltip", "treeImportData", "treeDownloadData", "Import JSON", "Export XML", `"date", "time"`, `"tree"`, `"chart"`, `tree: '<rect`} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer tree support missing %q", want)
		}
	}
}

func TestTreeWholeSelectionAndContextEditing(t *testing.T) {
	// The designer must select the whole tree, never an internal item, and the
	// inspector must not invoke the per-item panel.
	if strings.Contains(GuiHTML, "box.appendChild(treeItemControls(n))") {
		t.Fatalf("tree item inspector must not be invoked")
	}
	for _, want := range []string{
		"function treeOwnerPath", "function normalizeTreePaths", "function previewTreeContext",
		"function findPreviewTreeItem", "function treeInsertItem", "function treeDeleteItem",
		"function treeDuplicateItem", "function treeMoveItem", "function treeIndentItem",
		"function treeOutdentItem", "Add child item", "Add item below",
		"Edit item label...", "Indent (under previous)", "Outdent",
		"sui-tree-ctx", "ownsEvents", "CHART_EVENT_FIELDS", "renderChartEventControls",
		`{ name: "Events", fields: [`, `type: "event"`,
	} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer whole-tree support missing %q", want)
		}
	}
	for _, want := range []string{
		"addRuntimeItem", "moveRuntimeItem", "indentRuntimeItem", "outdentRuntimeItem",
		"Add child item", "Add item below", "Outdent",
	} {
		if !strings.Contains(TreeRuntime, want) {
			t.Errorf("runtime tree structural menu missing %q", want)
		}
	}
	if !strings.Contains(GuiPreviewPicker, ".sui-tree-item") || !strings.Contains(GuiPreviewPicker, "sui-tree-ctx") {
		t.Fatalf("preview picker must forward right-clicks on tree items")
	}
}

func TestTreeDefaultIconsAndBranchSymbols(t *testing.T) {
	src := `view main {
	tree "Files" iconLeaf: "L" iconParent: "P" iconParentOpen: "PO" toggleClosed: "+" toggleOpen: "-" {
		treeitem "folder" open: true {
			treeitem "child"
		}
		treeitem "file"
	}
}
`
	prog, dl := ParseSource(src, "tree_icons.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-tree-icon-leaf="L"`, `data-sui-tree-icon-parent="P"`,
		`data-sui-tree-icon-parent-open="PO"`, `data-sui-tree-toggle-closed="+"`,
		`data-sui-tree-toggle-open="-"`, `data-sui-icon="PO"`, `data-sui-icon="L"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("default icon/branch markup missing %q", want)
		}
	}
	model := StringifyJson(GuiModelJSON(prog))
	for _, want := range []string{`"iconLeaf":"L"`, `"iconParent":"P"`, `"iconParentOpen":"PO"`, `"toggleClosed":"+"`, `"toggleOpen":"-"`} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
	out := ViewNodesSource(prog.Views()[0].Body(), 1)
	for _, want := range []string{`iconLeaf: "L"`, `iconParent: "P"`, `iconParentOpen: "PO"`, `toggleClosed: "+"`, `toggleOpen: "-"`} {
		if !strings.Contains(out, want) {
			t.Errorf("designer source missing %q:\n%s", want, out)
		}
	}
	for _, want := range []string{
		`treeIconLeaf`, `treeIconParent`, `treeIconParentOpen`,
		`treeToggleClosed`, `treeToggleOpen`, `data-sui-tree-icon-leaf`,
		`iconLeaf: treeIconLeaf`, `toggleClosed: treeToggleClosed`,
	} {
		if !strings.Contains(TreeRuntime, want) {
			t.Errorf("tree runtime defaults missing %q", want)
		}
	}
	for _, want := range []string{
		`name: "Leaf"`, `name: "Folders"`, `name: "Structure"`,
		`name: "Branches"`, `name: "Editing"`, `iconSetForField`,
	} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("icon picker missing %q", want)
		}
	}
	for _, unwanted := range []string{`name: "People"`, `name: "Symbols"`, `name: "Media"`} {
		if strings.Contains(GuiHTML, unwanted) {
			t.Errorf("icon picker still exposes unrelated category %q", unwanted)
		}
	}
}

func TestTreeOnTreeChangeEvent(t *testing.T) {
	src := `state version = 0
fn changed() -> null {
	version = version + 1
}
fn snapshot() -> null {
	text, err = tree.lastChange("files")
	if err != null {
		text = ""
	}
}
view main {
	tree "Files" name: "files" contextMenu: true onTreeChange: changed {
		treeitem "src"
	}
}
`
	prog, dl := ParseSource(src, "tree_change.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `data-sui-tree-ontreechange="changed"`) {
		t.Errorf("generated tree markup missing onTreeChange handler")
	}
	if !strings.Contains(js, `__SUI_TREE.lastChange`) {
		t.Errorf("generated JS missing tree.lastChange mapping")
	}
	model := StringifyJson(GuiModelJSON(prog))
	if !strings.Contains(model, `"onTreeChange":"changed"`) {
		t.Errorf("designer model missing onTreeChange: %s", model)
	}
	out := ViewNodesSource(prog.Views()[0].Body(), 1)
	if !strings.Contains(out, `onTreeChange: changed`) {
		t.Errorf("designer source missing onTreeChange:\n%s", out)
	}
	for _, want := range []string{
		`data-sui-tree-ontreechange`, `treeChangeHandler`, `notifyTreeChange`,
		`notifyTreeChange(`, `treeLastChange`, `changeInfo`, `lastChange`,
		`changeInfo('delete'`, `changeInfo('cell'`, `refreshAfterEdit(li, 'add'`,
		`op: 'import'`, `op: 'clear'`,
	} {
		if !strings.Contains(TreeRuntime, want) {
			t.Errorf("tree runtime onTreeChange support missing %q", want)
		}
	}
	if !strings.Contains(GuiHTML, `onTreeChange`) || !strings.Contains(GuiHTML, `ownsContent`) {
		t.Errorf("designer tree events/content handling missing")
	}
}

func TestTreePhaseAFeatures(t *testing.T) {
	src := `state selected = ""
state checkedCount = 0
fn checked() -> null {
}
view main {
	tree "Files" name: "files" value: selected checkbox: true counts: true guides: true truncate: true rowHeight: 32 iconLeaf: "D" iconParent: "D" iconParentOpen: "O" onCheck: checked {
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
		`data-sui-tree-icon-leaf="D"`, `data-sui-tree-icon-parent="D"`,
		`data-sui-tree-icon-parent-open="O"`,
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
		if !strings.Contains(TreeRuntime, want) {
			t.Errorf("phase A runtime missing %q", want)
		}
	}
	for _, want := range []string{
		`--sui-tree-row-height`, `--sui-tree-guide-color`,
		`--sui-tree-badge-bg`, `--sui-tree-check-on`,
		`.sui-tree-badge-info`, `.sui-tree-meta`,
	} {
		if !strings.Contains(SuiCssBase, want) {
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
	tree "Files" guides: false truncate: false checkbox: true counts: true rowHeight: 30 iconLeaf: "D" iconParent: "D" iconParentOpen: "O" onCheck: checked {
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
	model := StringifyJson(GuiModelJSON(prog))
	for _, want := range []string{
		`"guides":false`, `"truncate":false`, `"checkbox":true`, `"counts":true`,
		`"rowHeight":"30"`, `"iconLeaf":"D"`, `"iconParent":"D"`, `"iconParentOpen":"O"`, `"icon":"F"`, `"iconOpen":"O"`, `"onCheck":"checked"`,
		`"badge":"2"`, `"badgeType":"warning"`, `"meta":"1 KB"`, `"checked":true`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
	out := ViewNodesSource(prog.Views()[0].Body(), 1)
	for _, want := range []string{
		`guides: false`, `truncate: false`, `checkbox: true`, `counts: true`,
		`rowHeight: 30`, `iconLeaf: "D"`, `iconParent: "D"`, `iconParentOpen: "O"`, `icon: "F"`, `iconOpen: "O"`, `onCheck: checked`,
		`badge: "2"`, `badgeType: "warning"`, `meta: "1 KB"`, `checked: true`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("designer source missing %q:\n%s", want, out)
		}
	}
	for _, want := range []string{
		`"Indent guides"`, `"Hierarchical checkboxes"`, `"Child counts"`,
		`"Item icon (no children)"`, `"Parent icon (closed)"`, `"Parent icon (open)"`,
		`"Closed branch symbol"`, `"Open branch symbol"`, `"metadata"`,
		`--sui-tree-guide-color`, `.td-row .td-icon`,
	} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer phase A support missing %q", want)
		}
	}
}

func TestTreePhaseBCDFeatures(t *testing.T) {
	src := `state selected = ""
state selectedKeys = "[]"
state loadKey = ""
fn moved() -> null {
}
fn loaded() -> null {
}
fn menu() -> null {
}
view main {
	tree "Files" name: "files" value: selected selection: selectedKeys multiSelect: true search: true drag: true contextMenu: true columns: "size,date,status" columnLabels: "Size,Modified,State" columnWidths: "120,90,80" sort: "asc" persistSelection: true loadKey: loadKey onMove: moved onLoad: loaded onContextMenu: menu virtual: true tooltips: true fontStep: 1 {
		treeitem "remote" key: "remote" lazy: true size: "2 KB" date: "today" status: "ok"
	}
}
`
	prog, dl := ParseSource(src, "tree_bcd.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-tree-multi="1"`, `data-sui-tree-search="1"`,
		`data-sui-tree-drag="1"`, `data-sui-tree-context-menu="1"`,
		`data-sui-tree-virtual="1"`,
		`data-sui-tree-tooltips="1"`, `data-sui-tree-font-step="1"`,
		`data-sui-tree-columns="size,date,status"`,
		`data-sui-tree-column-labels="Size,Modified,State"`,
		`data-sui-tree-column-widths="120,90,80"`,
		`data-sui-tree-sort="asc"`, `data-sui-tree-persist-selection="1"`,
		`sui-tree-grid-head`, `>Size</span>`, `style="width:120px"`,
		`data-sui-tree-selection-setter="setSelectedKeys"`,
		`data-sui-tree-load-key-setter="setLoadKey"`,
		`data-sui-tree-onmove="moved"`, `data-sui-tree-onload="loaded"`,
		`data-sui-tree-oncontextmenu="menu"`,
		`sui-tree-search-input`, `sui-tree-grid-head`,
		`sui-tree-cell-size`, `sui-tree-cell-date`, `sui-tree-cell-status`,
		`data-sui-size="2 KB"`, `data-sui-date="today"`, `data-sui-status="ok"`,
		`data-sui-lazy="1"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("phase B/C/D tree markup missing %q", want)
		}
	}
	for _, want := range []string{
		`applySearch`, `subtreeMatches`, `sui-tree-match`,
		`canDrop`, `syncTreeStore`, `sui-tree-drop-inside`,
		`setChildren`, `setChildrenError`, `loadLazy`, `sui-tree-loading`,
		`selectedKeys`, `checkedKeys`, `expandAll`, `collapseAll`, `expandTo`,
		`sui-tree-context`, `aria-multiselectable`, `data-sui-tree-selection-setter`,
		`sortNodes`, `pushUndo`, `undoMove`, `lastMove`, `retryChildren`, `sui-tree-retry`,
		`sui-tree-tooltip`, `showTooltip`, `aria-describedby`,
	} {
		if !strings.Contains(TreeRuntime, want) {
			t.Errorf("phase B/C/D runtime missing %q", want)
		}
	}
	for _, want := range []string{
		`.sui-tree-search`, `.sui-tree-grid-head`, `.sui-tree-cell`,
		`.sui-tree-context`, `.sui-tree-drop-before`, `.sui-tree-drop-after`,
		`.sui-tree-drop-inside`, `.sui-tree-loading`, `.sui-tree-match`,
		`content-visibility: auto`, `.sui-tree-retry`, `.sui-tree-tooltip`,
	} {
		if !strings.Contains(SuiCssBase, want) {
			t.Errorf("phase B/C/D CSS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`sui-tree-search-input`, `sui-tree-grid-head`, `sui-tree-cell-size`,
		`>Size</span>`, `style="width:120px"`,
		`data-sui-size="2 KB"`, `data-sui-lazy="1"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static phase B/C/D markup missing %q", want)
		}
	}
}

func TestTreePhaseBCDesigner(t *testing.T) {
	src := `state selectedKeys = "[]"
state loadKey = ""
fn moved() -> null {
}
view main {
	tree "Files" multiSelect: true selection: selectedKeys search: true drag: true contextMenu: true columns: "size,date,status" columnLabels: "Size,Modified,State" columnWidths: "120,90,80" sort: "asc" persistSelection: true loadKey: loadKey onMove: moved {
		treeitem "remote" lazy: true size: "2 KB" date: "today" status: "ok"
	}
}
`
	prog, dl := ParseSource(src, "tree_bcd_designer.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	model := StringifyJson(GuiModelJSON(prog))
	for _, want := range []string{
		`"multiSelect":true`, `"selection":"selectedKeys"`, `"search":true`,
		`"drag":true`, `"contextMenu":true`, `"columns":"size,date,status"`,
		`"columnLabels":"Size,Modified,State"`, `"columnWidths":"120,90,80"`,
		`"sort":"asc"`, `"persistSelection":true`,
		`"loadKey":"loadKey"`, `"onMove":"moved"`, `"lazy":true`,
		`"size":"2 KB"`, `"date":"today"`, `"status":"ok"`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer phase B/C model missing %q: %s", want, model)
		}
	}
	out := ViewNodesSource(prog.Views()[0].Body(), 1)
	for _, want := range []string{
		`multiSelect: true`, `selection: selectedKeys`, `search: true`,
		`drag: true`, `contextMenu: true`, `columns: "size,date,status"`,
		`columnLabels: "Size,Modified,State"`, `columnWidths: "120,90,80"`,
		`sort: "asc"`, `persistSelection: true`,
		`loadKey: loadKey`, `onMove: moved`, `lazy: true`,
		`size: "2 KB"`, `date: "today"`, `status: "ok"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("designer phase B/C source missing %q:\n%s", want, out)
		}
	}
	for _, want := range []string{
		`"Multi-select"`, `"Search box"`, `"Drag and drop"`, `"Context menu"`,
		`"Show columns"`, `"Sort children"`, `"Columns"`,
		`"Persist selection"`, `.td-row .td-size`,
		`name: "General"`, `name: "Extra"`, `name: "Tooltip"`, `"Runtime name"`,
	} {
		if !strings.Contains(GuiHTML, want) {
			t.Errorf("designer phase B/C support missing %q", want)
		}
	}
}

func TestTreeColumnsDisabled(t *testing.T) {
	src := "view main {\n\ttree \"Files\" columns: \"size,date,status\" columnsEnabled: false {\n\t\ttreeitem \"a\" size: \"1 KB\" date: \"today\" status: \"ok\"\n\t}\n}\n"
	prog, dl := ParseSource(src, "tree_columns.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	if strings.Contains(html, "sui-tree-grid-head") {
		t.Error("disabled columns must not render the grid header")
	}
	if strings.Contains(html, "sui-tree-cell-size") {
		t.Error("disabled columns must not render grid cells")
	}
	if !strings.Contains(html, `data-sui-tree-columns-enabled="0"`) {
		t.Error("disabled columns must expose the disabled flag")
	}
}

func TestTreeCustomColumnsAndFontStep(t *testing.T) {
	src := `view main {
	tree "Files" columns: "owner,size" columnLabels: "Owner,Size" columnWidths: "90,70" fontStep: 1 {
		treeitem "src" size: "2 KB" cells: "owner=Anna" {
			treeitem "main.go" size: "1 KB" cells: "owner=Luca"
		}
	}
}
`
	prog, dl := ParseSource(src, "tree_custom_columns.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`Owner`, `Size`, `Anna`, `Luca`, `2 KB`, `1 KB`,
		`data-sui-cells="owner=Anna"`,
		`font-size:calc(var(--sui-tree-font, 13px) - 1px)`,
		`data-sui-tree-font-step="1"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("custom columns/font step missing %q", want)
		}
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-tree-font-step="1"`, `data-sui-cells="owner=Anna"`,
		`data-sui-tree-columns="owner,size"`, `data-sui-tree-column-labels="Owner,Size"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("dynamic custom columns missing %q", want)
		}
	}
	for _, want := range []string{
		`setColumnsFor`, `moveColumn`, `setCellFor`,
		`treeColumnOverrides`, `treeCellOverrides`, `itemCells`,
	} {
		if !strings.Contains(TreeRuntime, want) {
			t.Errorf("runtime columns support missing %q", want)
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
