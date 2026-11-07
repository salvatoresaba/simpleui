# Tree

A complete nested tree demo with selection state, independent expansion,
keyboard navigation, ARIA tree semantics and persisted open state.

## Run it

From the repository root:

```powershell
.\simpleui.exe serve examples\tree\app.sui --watch
```

Or build a desktop demo:

```powershell
.\simpleui.exe build examples\tree\app.sui --target desktop --out examples\tree\dist
.\examples\tree\dist\app.exe
```

A ready build is in `examples/tree/dist/`; `dist/app.sui` is the complete source
used for that build.

## Complete source

```sui
# Nested tree example.
# tree value: selected key state
# treeitem key: stable item identifier
# treeitem open: initial expansion state
# open/closed state is persisted per tree path in localStorage.

state selected = "main"
state renameKey = ""
state renameValue = ""
state importPath = ""
state notice = "Choose a JSON/XML file, then import it. Downloads use the current tree."
state exportText = ""

fn selectedChanged() -> null {
}

fn renamed() -> null {
}

async fn importJsonFile() -> null {
	if importPath == "" {
		notice = "Choose a JSON/XML file first."
		return
	}
	n, err = tree.importJsonFile("files", importPath)
	if err != null {
		notice = "JSON import failed: " + err
	} else {
		notice = "Imported " + str(n) + " nodes from JSON."
	}
}

async fn importXmlFile() -> null {
	if importPath == "" {
		notice = "Choose a JSON/XML file first."
		return
	}
	n, err = tree.importXmlFile("files", importPath)
	if err != null {
		notice = "XML import failed: " + err
	} else {
		notice = "Imported " + str(n) + " nodes from XML."
	}
}

fn downloadJson() -> null {
	body, err = tree.exportJson("files")
	if err != null {
		notice = "JSON export failed: " + err
		return
	}
	exportText = body
	n, derr = tree.downloadJson("files", "tree.json")
	if derr != null {
		notice = "JSON download failed: " + derr
	} else {
		notice = "JSON exported and downloaded."
	}
}

fn downloadXml() -> null {
	body, err = tree.exportXml("files")
	if err != null {
		notice = "XML export failed: " + err
		return
	}
	exportText = body
	n, derr = tree.downloadXml("files", "tree.xml")
	if derr != null {
		notice = "XML download failed: " + derr
	} else {
		notice = "XML exported and downloaded."
	}
}

fn clearTree() -> null {
	n, err = tree.clear("files")
	if err != null {
		notice = "Clear failed: " + err
	} else {
		exportText = ""
		notice = "Tree cleared. Import a file to load data."
	}
}

view main {
	title "Tree"
	h1 "Nested tree"
	p "Click a row to select it. Use the disclosure buttons or arrow keys to expand and collapse."

	tree "Project files" name: "files" value: selected onChange: selectedChanged editable: true editableDepth: 2 renameKey: renameKey renameValue: renameValue renameIcon: "✏️" style: cssSafe("--sui-tree-gap: 2px; --sui-tree-row-padding: 4px 8px; --sui-tree-row-radius: 9px; --sui-tree-hover-bg: #4c5cf724; --sui-tree-selected-bg: #4c5cf761; --sui-tree-selected-fg: #ffffff; --sui-tree-title-color: #dbe4ff; --sui-tree-panel-bg: #1c2230; --sui-tree-panel-border: 1px solid #303b55; --sui-tree-panel-radius: 12px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;") onRename: renamed {
		treeitem "src" key: "src" open: true style: cssSafe("--sui-tree-row-bg: #1a253b; --sui-tree-row-radius: 7px; --sui-tree-selected-bg: #2a3a5f; --sui-tree-selected-fg: #dbeafe;") {
			treeitem "main.go" key: "main" showRenameIcon: false
			treeitem "components" key: "components" open: true {
				treeitem "tree.go" key: "tree"
				treeitem "runtime.go" key: "runtime"
			}
		}
		treeitem "docs" key: "docs" {
			treeitem "README.md" key: "readme" renameIcon: "📝"
		}
		treeitem "go.mod" key: "gomod"
		treeitem "generated" key: "generated" editable: false
	}

	p "Selected key: {selected}"
	p "Last rename: {renameKey} -> {renameValue}"

	card {
		p "The selected key is stored in state selected. The tree restores expansion state after a re-render."
		p "Editable levels 1-2: double-click a label, press F2 or use the pencil, type, then Enter. Escape cancels. generated is explicitly read-only."
		p "The tree uses a custom ✏️ icon. main.go hides the pencil with showRenameIcon: false but still supports F2 and double-click. src uses a per-node folder style preset."
		p "Keyboard: Up/Down move, Right expands or enters, Left collapses or returns to the parent, Enter/Space selects, Home/End jump."
		p "ARIA: role=tree, treeitem, group, aria-expanded, aria-level and aria-selected."
	}

	card {
		p "Import/export: choose a .json or .xml file, import it into the tree, or download the current tree data."
		row {
			button "Choose JSON/XML..." pickFiles: importPath accept: ".json,.xml"
			button "Import JSON" onClick: importJsonFile
			button "Import XML" onClick: importXmlFile
		}
		row {
			button "Download JSON" onClick: downloadJson
			button "Download XML" onClick: downloadXml
			button "Clear tree" onClick: clearTree
		}
		p "Selected file: {importPath}"
		p "{notice}"
		pre "{exportText}"
	}

	accordion "How this example works" {
		accordionitem "Selection and expansion" {
			p "Clicking the label selects the item; clicking the arrow only expands or collapses it."
			p "Selection is independent from expansion and persists through state-driven re-renders."
		}
		accordionitem "Keys" {
			p "Each treeitem has a stable key used as the value of the selected state. If no key is set, the item path is used."
		}
		accordionitem "Accessibility" {
			p "Only one treeitem is in the tab order at a time; arrow keys move focus."
			p "Expanded parents expose aria-expanded; the selected item exposes aria-selected=true."
		}
	}
}
```

## Designer

Insert **Tree** from **Controls** (the component palette). It starts with a
parent item and a child. **Tree item** is not offered in the palette or Insert
menu: items are created and edited from **Edit tree data** or the tree context
menu. The structure list shows only the `tree` row, not its internal items.

The **Tree** inspector exposes only two actions:

- **Tree settings...** - style selector, General, Bindings, Columns, Colors,
  Title, Shape, Icons and Tooltip;
- **Edit tree data...** - labels, keys, nesting, open/editable flags and
  JSON/XML import/export.

The **General** tab holds the visible **Title**, the **Runtime name** (target
for `tree.*`, not the title), default/open node icons, and the behavior switches.
**Bindings** holds optional state bindings and suggests declared project states.
**Columns** enables and configures an arbitrary column list (add, rename,
resize, remove, reorder) plus child sorting. **Extra** has two sections: the edit
button (icon, background, text, font, border, radius, padding, hover) and child
counts (background, text, font, size, weight, padding, border, radius, minimum
width). **Shape** includes **Font step**, which reduces the font size for each
level below the root. The **Tooltip** tab enables styled tooltips. The **Style**
selector includes the shared styles and the 16 tree themes; **Apply** writes the
chosen style to the tree preview. There is no separate Hover tab: hover colors
are in **Colors**.

The right-click menu on the tree row offers **Edit tree data**, **Tree
settings**, **Add root item**, **Import JSON/XML** and **Export JSON/XML**.

Selecting an item in the live preview opens the per-item **Tree item** panel:

- `key` - stable string written to the selected state;
- `initially open` - default expansion state;
- `editable` - inherit / yes / read-only;
- `icon` / `open icon` - per-node glyphs;
- `badge` / `badge type` - status pill and its colour family;
- `metadata` - right-aligned text such as a size or date;
- `checked` - initial hierarchical checkbox state;
- `lazy children` - load children on first expansion;
- `size` / `date` / `status` - tree-grid cells;
- label text via **Content**.

`onChange`, `onRename`, `onCheck`, `onMove`, `onLoad` and `onContextMenu` are
available in **Events**.

```sui
tree "Files" name: "files" guides: true truncate: true checkbox: true counts: true rowHeight: 30 icon: "📁" iconOpen: "📂" multiSelect: true selection: selectedKeys search: true drag: true contextMenu: true columns: "size,date,status" loadKey: loadKey onCheck: checked onMove: moved onLoad: loaded {
	treeitem "src" open: true icon: "📁" iconOpen: "📂" badge: "12" badgeType: "info" meta: "2 KB" size: "2 KB" date: "today" status: "ok" checked: true {
		treeitem "main.go" icon: "📄" meta: "1 KB"
	}
	treeitem "remote" lazy: true
}
```

`treeitem badge` and `meta` render as inert text. A parent checkbox becomes
mixed when only some children are checked; clicking it updates the whole
subtree. `onCheck` runs after the click and `tree.exportJson(name)` includes the
resulting `checked` flags.

`multiSelect` enables Ctrl/Command and Shift selection; `selection` receives a
JSON array of keys. `search` adds a filter box, `drag` enables reordering with
cycle checks, and `contextMenu` adds a built-in right-click menu with rename,
duplicate, delete, expand/collapse, copy key and expand/collapse all. `columns`
renders a tree-grid from each item's `size`, `date` and `status`;
`columnLabels`/`columnWidths` customize the header and widths, and `sort` orders
children by label. `persistSelection` restores selected keys after reload.
`virtual` enables browser offscreen rendering (`content-visibility`) for large
trees. Lazy items call `onLoad`, write their key to `loadKey`, and receive
children through `tree.setChildren(name, key, json)`; errors offer a Retry
button and `tree.retryChildren`. `tree.expandAll`, `tree.collapseAll`,
`tree.expandTo`, `tree.selectedKeys`, `tree.checkedKeys`, `tree.lastMove` and
`tree.undoMove` are available from application code. Runtime columns can be
created, renamed, resized and reordered with `tree.setColumns(name, json)`,
inspected with `tree.getColumns(name)`, reordered with
`tree.moveColumn(name, from, to)` and updated cell-by-cell with
`tree.setCell(name, key, column, value)`.

## Behavior

- Clicking a row selects it; clicking the disclosure arrow only toggles it.
- Ctrl/Command click toggles a row; Shift click selects a range.
- Shift+Up/Down and Shift+Home/End extend a multi-selection.
- Ctrl/Command+F focuses the search box; Escape clears it.
- Search expands matching branches and highlights every occurrence.
- Right-click selects the item and opens the context menu when enabled.
- Dragging moves nodes between before/after/inside targets; invalid cycles are
  rejected.
- Selection is stored in the bound state and announced with `aria-selected`.
- Expansion is independent from selection and persisted per tree path in
  `localStorage`.
- Up/Down move through visible items.
- Right expands a parent or moves into its first child.
- Left collapses a parent or moves to its parent.
- Enter/Space selects the focused item.
- Home/End move to the first or last visible item.
- Only one item is in the tab order at a time.
- `role="tree"`, `role="treeitem"`, `role="group"`, `aria-level`,
  `aria-expanded` and `aria-selected` describe the structure.

## Inline rename

`editable: true` enables rename; `editableDepth: 2` limits it to levels 1-2.

- Double-click a label, press `F2`, or click the pencil button on an editable
  item.
- `Enter` commits; `Escape` cancels.
- Empty labels are rejected.
- The generated `key` never changes.
- `renameKey` receives the node key and `renameValue` the new label.
- `onRename` runs after both setters.
- The label is stored in `localStorage` keyed by tree title and item path, so it
  survives re-renders and reloads.
- A `treeitem editable: false` explicitly locks that node and its descendants.
- A `treeitem editable: true` can re-enable a node unless an ancestor is locked.
- Nodes without an override follow `tree editable` and `editableDepth`.

## Styling

The Tree inspector has a **style preset** menu:

- Default
- Compact
- Comfortable
- Dense
- Bordered
- Soft rounded
- Minimal
- Neon
- Forest
- Amber
- Ocean
- High contrast
- Outline
- Mono
- Candy
- Cards
- Custom / edit style expr

A preset writes CSS custom properties (`--sui-tree-*`) into the tree element's
**style expr** field. **Tree settings** exposes the same colors, typography and
title options as structured fields (Colors, Title, Shape); the title can be
left/center/right aligned and placed above or below the items. Each Tree item
also has a **node style preset** menu
(Inherit, Muted, Accent, Success, Warning, Danger, Folder, Compact row or
Custom); node presets write row variables into that item's own `style expr`.
The example uses the **Soft rounded** tree preset and the **Folder** node preset
directly in its source:

```sui
# Nested tree example.
# tree value: selected key state
# treeitem key: stable item identifier
# treeitem open: initial expansion state
# open/closed state is persisted per tree path in localStorage.

state selected = "main"
state renameKey = ""
state renameValue = ""
state importPath = ""
state notice = "Choose a JSON/XML file, then import it. Downloads use the current tree."
state exportText = ""

fn selectedChanged() -> null {
}

fn renamed() -> null {
}

async fn importJsonFile() -> null {
	if importPath == "" {
		notice = "Choose a JSON/XML file first."
		return
	}
	n, err = tree.importJsonFile("files", importPath)
	if err != null {
		notice = "JSON import failed: " + err
	} else {
		notice = "Imported " + str(n) + " nodes from JSON."
	}
}

async fn importXmlFile() -> null {
	if importPath == "" {
		notice = "Choose a JSON/XML file first."
		return
	}
	n, err = tree.importXmlFile("files", importPath)
	if err != null {
		notice = "XML import failed: " + err
	} else {
		notice = "Imported " + str(n) + " nodes from XML."
	}
}

fn downloadJson() -> null {
	body, err = tree.exportJson("files")
	if err != null {
		notice = "JSON export failed: " + err
		return
	}
	exportText = body
	n, derr = tree.downloadJson("files", "tree.json")
	if derr != null {
		notice = "JSON download failed: " + derr
	} else {
		notice = "JSON exported and downloaded."
	}
}

fn downloadXml() -> null {
	body, err = tree.exportXml("files")
	if err != null {
		notice = "XML export failed: " + err
		return
	}
	exportText = body
	n, derr = tree.downloadXml("files", "tree.xml")
	if derr != null {
		notice = "XML download failed: " + derr
	} else {
		notice = "XML exported and downloaded."
	}
}

fn clearTree() -> null {
	n, err = tree.clear("files")
	if err != null {
		notice = "Clear failed: " + err
	} else {
		exportText = ""
		notice = "Tree cleared. Import a file to load data."
	}
}

view main {
	title "Tree"
	h1 "Nested tree"
	p "Click a row to select it. Use the disclosure buttons or arrow keys to expand and collapse."

	tree "Project files" name: "files" value: selected onChange: selectedChanged editable: true editableDepth: 2 renameKey: renameKey renameValue: renameValue renameIcon: "✏️" style: cssSafe("--sui-tree-gap: 2px; --sui-tree-row-padding: 4px 8px; --sui-tree-row-radius: 9px; --sui-tree-hover-bg: #4c5cf724; --sui-tree-selected-bg: #4c5cf761; --sui-tree-selected-fg: #ffffff; --sui-tree-title-color: #dbe4ff; --sui-tree-panel-bg: #1c2230; --sui-tree-panel-border: 1px solid #303b55; --sui-tree-panel-radius: 12px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;") onRename: renamed {
		treeitem "src" key: "src" open: true style: cssSafe("--sui-tree-row-bg: #1a253b; --sui-tree-row-radius: 7px; --sui-tree-selected-bg: #2a3a5f; --sui-tree-selected-fg: #dbeafe;") {
			treeitem "main.go" key: "main" showRenameIcon: false
			treeitem "components" key: "components" open: true {
				treeitem "tree.go" key: "tree"
				treeitem "runtime.go" key: "runtime"
			}
		}
		treeitem "docs" key: "docs" {
			treeitem "README.md" key: "readme" renameIcon: "📝"
		}
		treeitem "go.mod" key: "gomod"
		treeitem "generated" key: "generated" editable: false
	}

	p "Selected key: {selected}"
	p "Last rename: {renameKey} -> {renameValue}"

	card {
		p "The selected key is stored in state selected. The tree restores expansion state after a re-render."
		p "Editable levels 1-2: double-click a label, press F2 or use the pencil, type, then Enter. Escape cancels. generated is explicitly read-only."
		p "The tree uses a custom ✏️ icon. main.go hides the pencil with showRenameIcon: false but still supports F2 and double-click. src uses a per-node folder style preset."
		p "Keyboard: Up/Down move, Right expands or enters, Left collapses or returns to the parent, Enter/Space selects, Home/End jump."
		p "ARIA: role=tree, treeitem, group, aria-expanded, aria-level and aria-selected."
	}

	card {
		p "Import/export: choose a .json or .xml file, import it into the tree, or download the current tree data."
		row {
			button "Choose JSON/XML..." pickFiles: importPath accept: ".json,.xml"
			button "Import JSON" onClick: importJsonFile
			button "Import XML" onClick: importXmlFile
		}
		row {
			button "Download JSON" onClick: downloadJson
			button "Download XML" onClick: downloadXml
			button "Clear tree" onClick: clearTree
		}
		p "Selected file: {importPath}"
		p "{notice}"
		pre "{exportText}"
	}

	accordion "How this example works" {
		accordionitem "Selection and expansion" {
			p "Clicking the label selects the item; clicking the arrow only expands or collapses it."
			p "Selection is independent from expansion and persists through state-driven re-renders."
		}
		accordionitem "Keys" {
			p "Each treeitem has a stable key used as the value of the selected state. If no key is set, the item path is used."
		}
		accordionitem "Accessibility" {
			p "Only one treeitem is in the tab order at a time; arrow keys move focus."
			p "Expanded parents expose aria-expanded; the selected item exposes aria-selected=true."
		}
	}
}
```

Edit the `style expr` field in Appearance for custom CSS. The CSS variables
control row spacing/radius/colors, hover and selected colors, title color/size,
panel background/border/radius/padding and indentation. Dynamic style remains
subject to `cssSafe`: use 8-digit hex colors such as `#4c5cf761` instead of
`rgba(...)`, because comma-containing declarations are dropped.

The **rename icon** field changes the pencil glyph. **show rename icon** hides
the whole tree's pencil; a Tree item can inherit, show or hide it per node and
override the glyph. The example uses `✏️` for the tree, `📝` for `README.md`
and `showRenameIcon: false` on `main.go`. Hidden pencils do not disable F2 or
double-click rename.

## JSON/XML import and export

The Tree inspector also has **Import JSON**, **Import XML**, **Export JSON** and
**Export XML** buttons. They replace/download the design-time `treeitem`
children in the source, which is useful for seeding a demo or a static tree.

For application data, the example declares `name: "files"` on the tree and uses
the core `tree` module:

- **Choose JSON/XML...** uses `pickFiles: importPath` to select a file.
- **Import JSON** calls `tree.importJsonFile("files", importPath)`.
- **Import XML** calls `tree.importXmlFile("files", importPath)`.
- **Download JSON/XML** calls export, then `tree.downloadJson`/`downloadXml`.
- **Clear tree** calls `tree.clear("files")`.

The app also shows the exported JSON/XML in a `<pre>` block. Imported models
survive state-driven re-renders, and only the node structure is replaced: the
tree title, name, styling, selection/rename bindings and accessibility runtime
stay in place.

The same API accepts text directly:

```sui
# Nested tree example.
# tree value: selected key state
# treeitem key: stable item identifier
# treeitem open: initial expansion state
# open/closed state is persisted per tree path in localStorage.

state selected = "main"
state renameKey = ""
state renameValue = ""
state importPath = ""
state notice = "Choose a JSON/XML file, then import it. Downloads use the current tree."
state exportText = ""

fn selectedChanged() -> null {
}

fn renamed() -> null {
}

async fn importJsonFile() -> null {
	if importPath == "" {
		notice = "Choose a JSON/XML file first."
		return
	}
	n, err = tree.importJsonFile("files", importPath)
	if err != null {
		notice = "JSON import failed: " + err
	} else {
		notice = "Imported " + str(n) + " nodes from JSON."
	}
}

async fn importXmlFile() -> null {
	if importPath == "" {
		notice = "Choose a JSON/XML file first."
		return
	}
	n, err = tree.importXmlFile("files", importPath)
	if err != null {
		notice = "XML import failed: " + err
	} else {
		notice = "Imported " + str(n) + " nodes from XML."
	}
}

fn downloadJson() -> null {
	body, err = tree.exportJson("files")
	if err != null {
		notice = "JSON export failed: " + err
		return
	}
	exportText = body
	n, derr = tree.downloadJson("files", "tree.json")
	if derr != null {
		notice = "JSON download failed: " + derr
	} else {
		notice = "JSON exported and downloaded."
	}
}

fn downloadXml() -> null {
	body, err = tree.exportXml("files")
	if err != null {
		notice = "XML export failed: " + err
		return
	}
	exportText = body
	n, derr = tree.downloadXml("files", "tree.xml")
	if derr != null {
		notice = "XML download failed: " + derr
	} else {
		notice = "XML exported and downloaded."
	}
}

fn clearTree() -> null {
	n, err = tree.clear("files")
	if err != null {
		notice = "Clear failed: " + err
	} else {
		exportText = ""
		notice = "Tree cleared. Import a file to load data."
	}
}

view main {
	title "Tree"
	h1 "Nested tree"
	p "Click a row to select it. Use the disclosure buttons or arrow keys to expand and collapse."

	tree "Project files" name: "files" value: selected onChange: selectedChanged editable: true editableDepth: 2 renameKey: renameKey renameValue: renameValue renameIcon: "✏️" style: cssSafe("--sui-tree-gap: 2px; --sui-tree-row-padding: 4px 8px; --sui-tree-row-radius: 9px; --sui-tree-hover-bg: #4c5cf724; --sui-tree-selected-bg: #4c5cf761; --sui-tree-selected-fg: #ffffff; --sui-tree-title-color: #dbe4ff; --sui-tree-panel-bg: #1c2230; --sui-tree-panel-border: 1px solid #303b55; --sui-tree-panel-radius: 12px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;") onRename: renamed {
		treeitem "src" key: "src" open: true style: cssSafe("--sui-tree-row-bg: #1a253b; --sui-tree-row-radius: 7px; --sui-tree-selected-bg: #2a3a5f; --sui-tree-selected-fg: #dbeafe;") {
			treeitem "main.go" key: "main" showRenameIcon: false
			treeitem "components" key: "components" open: true {
				treeitem "tree.go" key: "tree"
				treeitem "runtime.go" key: "runtime"
			}
		}
		treeitem "docs" key: "docs" {
			treeitem "README.md" key: "readme" renameIcon: "📝"
		}
		treeitem "go.mod" key: "gomod"
		treeitem "generated" key: "generated" editable: false
	}

	p "Selected key: {selected}"
	p "Last rename: {renameKey} -> {renameValue}"

	card {
		p "The selected key is stored in state selected. The tree restores expansion state after a re-render."
		p "Editable levels 1-2: double-click a label, press F2 or use the pencil, type, then Enter. Escape cancels. generated is explicitly read-only."
		p "The tree uses a custom ✏️ icon. main.go hides the pencil with showRenameIcon: false but still supports F2 and double-click. src uses a per-node folder style preset."
		p "Keyboard: Up/Down move, Right expands or enters, Left collapses or returns to the parent, Enter/Space selects, Home/End jump."
		p "ARIA: role=tree, treeitem, group, aria-expanded, aria-level and aria-selected."
	}

	card {
		p "Import/export: choose a .json or .xml file, import it into the tree, or download the current tree data."
		row {
			button "Choose JSON/XML..." pickFiles: importPath accept: ".json,.xml"
			button "Import JSON" onClick: importJsonFile
			button "Import XML" onClick: importXmlFile
		}
		row {
			button "Download JSON" onClick: downloadJson
			button "Download XML" onClick: downloadXml
			button "Clear tree" onClick: clearTree
		}
		p "Selected file: {importPath}"
		p "{notice}"
		pre "{exportText}"
	}

	accordion "How this example works" {
		accordionitem "Selection and expansion" {
			p "Clicking the label selects the item; clicking the arrow only expands or collapses it."
			p "Selection is independent from expansion and persists through state-driven re-renders."
		}
		accordionitem "Keys" {
			p "Each treeitem has a stable key used as the value of the selected state. If no key is set, the item path is used."
		}
		accordionitem "Accessibility" {
			p "Only one treeitem is in the tab order at a time; arrow keys move focus."
			p "Expanded parents expose aria-expanded; the selected item exposes aria-selected=true."
		}
	}
}
```

JSON node schema:

```json
[
  {
    "label": "src",
    "key": "src",
    "open": true,
    "children": [
      { "label": "main.go", "key": "main" }
    ]
  }
]
```

XML equivalent:

```xml
<tree name="files">
  <item label="src" key="src" open="true">
    <item label="main.go" key="main"/>
  </item>
</tree>
```

Limits and security:

- 10,000 nodes, depth 64, labels up to 2,000 characters, keys up to 512;
- input uses the global `maxText` limit (20 MB by default);
- labels, keys and icons are text, never HTML;
- XML DTD/entity declarations are rejected;
- local open/rename overrides for the tree are cleared by an import;
- export reads the visible labels and open state, so inline renames are included.

On desktop, `dialog.pickFile()` returns a path rather than a browser file object:
use `fs.read(path)` and pass the text to `tree.importJson`/`tree.importXml`.

## Security

Tree labels, keys and item text are generated as escaped text/attributes; no raw
HTML is used. Rename input uses text values, never HTML. The control adds no
network, file or command capability. Keys are application data and should be
unique within a tree. Open state and local rename labels are stored in
`localStorage` and are not a security boundary. Renaming is a UI permission:
enforce the same permission on the server before persisting changes.
