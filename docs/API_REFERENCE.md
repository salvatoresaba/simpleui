# SimpleUI API Reference

Complete guide to the SimpleUI language, builtins, stdlib and CLI.

## Summary

- [SUI language](#sui-language-essential-syntax)
- [Drag and drop](#drag-and-drop)
- [Menu bar](#menu-bar)
- [Tabs](#tabs)
- [Pages and navigation](#pages-and-navigation)
- [Modal dialogs](#modal-dialogs)
- [Tables and untrusted data](#tables-and-untrusted-data)
- [Security](#security)
- [Iframe](#iframe)
- [Builtins](BUILTINS.md)
- [Stdlib](STDLIB.md)
- [CLI](#cli)

## SUI language (essential syntax)

### Types

- `int`, `float`, `string`, `bool`, `null`, `any`
- `array[T]`, `map[string]T`
- Structured types: `type Name struct { field: Type }`

### Declarations

```sui
state count = 0
let name = "Alice"
type User struct { id: int, name: string }
view App { ... }
fn inc() { count++ }
async fn load() { ... }
```

### Control flow

```sui
if count > 0 { ... } else { ... }
loop i in 0..5 { ... }
while count < 10 { ... }
loop k, v in m { ... }  # map iteration
break
continue
```

### Components and views

```sui
view Card(title: string) {
  children  # slot
}

view App {
  Card title: "Demo" {
    text "Content"
  }
}
```

### Assignments and operators

- Compound: `+=`, `-=`, `*=`, `/=`
- Post/pre: `x++`, `x--`
- Logical: `and`, `or`, `not`
- Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
- Arithmetic: `+`, `-`, `*`, `/`, `%`

### Form controls

`required: true` is available on `input`, `textarea`, `select`, `combobox`,
`radio`, `number`, `date`, `time` and `datepicker`. Native controls use the
browser's `required` attribute; the combobox also reports an inline required
error when no value is selected.

## Drag and drop

The runtime supports moving elements inside the page and accepting files or
directories from the operating system. Handlers are still zero-argument
functions; read the current drag with the `drag.*` builtins. For the designer
inspector, the difference between designer and application DnD, and the
list/table limitations, see [GUI editor: drag and drop](GUI_EDITOR.md).

```sui
state items = ["Design", "Build", "Ship"]
state attachment = ""

fn reorder() -> null {
	let src = indexOf(items, drag.source())
	let dst = indexOf(items, drag.target())
	if src < 0 or dst < 0 or src == dst {
		return
	}
	items = moveItem(items, src, dst)
}

view main {
	list item in items {
		row draggable: true dragKey: item dropKey: item onDrop: reorder {
			text "{item}"
		}
	}
	input "Attachment" value: attachment dropFiles: attachment
}
```

### Attributes

| Attribute | Meaning |
|---|---|
| `draggable: true` | Marks the element as a drag source. `dragKey` alone also makes it draggable. |
| `dragKey: expr` | Payload key for the source; exposed as `drag.source()`. |
| `dropKey: expr` | Payload key for the target; exposed as `drag.target()`. |
| `dropFiles: state` | Automatic file/directory target. The named **string state** receives the dropped paths; an `input`/`textarea` with a matching `value:` binding displays them. |
| `pickFiles: state` | Native **file dialog** target. The named string state receives the selected paths. On a `file` control the native input is used; clicking a `button`, container or field opens a hidden picker. |
| `pickMode: "native" \| "fs" \| "auto"` | Picker implementation. `native` uses `<input type=file>`/`webkitdirectory`; `fs` uses the File System Access API folder picker (`showDirectoryPicker`); `auto` uses `fs` when available and falls back to native. For ordinary file selection, use `native`. |
| `directory: true` | On `file` uses the directory picker; on a `pickFiles` trigger it uses a hidden directory picker (`webkitdirectory`). |
| `onDragStart`, `onDragEnter`, `onDragOver`, `onDragLeave`, `onDrop`, `onDragEnd` | Event handlers; use `drag.*` to inspect the payload. |

`draggable`, `dragKey`, `dropKey`, `dropFiles` work on containers, rows, list
items and controls. `dropFiles` is not restricted to inputs: it can fill a
state used by any visible element. Always set `dragKey` on a source; a bare
`draggable: true` sends no useful key to `onDrop`.

### Builtins

| Function | Return | Meaning |
|---|---|---|
| `drag.source()` | string | Source key (the `dragKey`). |
| `drag.target()` | string | Current/drop-target key (the `dropKey`). |
| `drag.data()` | string | `text/plain` payload (or the source key for internal drags). |
| `drag.files()` | array[map] | File metadata: `name`, `path`, `size`, `type`, `isDirectory`. |
| `drag.folder()` | string | Common top-level folder of the current selection, or `""` for individual files / mixed roots. |
| `drag.error()` | string | Last file/directory rejection reason. |
| `drag.active()` | bool | Whether a drag is active. |
| `drag.clear()` | null | Clears the runtime drag state. |
| `moveItem(arr, from, to)` | array[T] | Returns a new array with one item moved. |

### File and directory drops

- The browser exposes **names and relative paths**, not absolute local paths.
  A folder drop includes paths such as `photos/2026/a.jpg`; a single file
  dropped from the desktop usually exposes only its name. Desktop hosts that
  provide `File.path` can expose the absolute path; do not depend on it.
- Directory traversal is recursive and asynchronous. `dropFiles` updates the
  target state when collection finishes. `drag.files()` provides the immediate
  metadata captured on `drop`; use `dropFiles` when you need every file in a
  dropped directory. Traversal relies on `webkitGetAsEntry` (Chromium/Edge);
  browsers without it fall back to the files that the browser exposes.
- Limits: 2,000 files, nesting depth 16 and 4,096 characters per path. Invalid
  or oversized input is rejected with a visible
  `Cannot accept the files: ...` message.
- On a textarea, paths are separated by newlines; on an input, by `, `.
- Only metadata is read. File bytes are not read, uploaded or written by this
  runtime.

Treat file names as untrusted text: render them through state (escaped by the
view) or `textContent`; never use them as HTML, CSS or an unvalidated URL. A
dropped name is not permission to read a path: desktop `fs` still requires the
build-time `--allow-dir` capability.

### File System Access API folder picker

On Chromium-based browsers and secure contexts (localhost/HTTPS), use
`pickMode: "fs"` for a folder picker backed by `showDirectoryPicker()`:

```sui
button "Choose folder..." pickFiles: attachment directory: true pickMode: "fs"
```

`pickMode: "auto"` uses the File System Access API when available and falls
back to the native `webkitdirectory` picker elsewhere. Unlike the native
picker, the FS API returns a directory handle, so an empty folder can still
yield its name. The runtime traverses it read-only, builds the same relative
paths/metadata, and fills `drag.folder()` explicitly.

Constraints:

- the user must grant permission in a browser gesture;
- Chromium/Edge only (not Firefox/Safari in current versions);
- secure context required;
- files are traversed read-only; no write/delete operations are exposed;
- reusing a previous handle/permission after reload is not implemented;
- `pickMode: "fs"` shows an explicit error when the API is unavailable, while
  `auto` falls back to the native picker.

### Folder selection

`pickFiles: state directory: true` opens the browser's folder picker. The
browser returns the files inside the folder recursively; `drag.folder()`
returns the common top-level folder name when the selected paths share one.

The native picker selects a **whole folder**, not an arbitrary subset of files
inside it. To analyze only some files, use the multi-file picker
(`multiple: true` without `directory`) or select the folder and filter the
selection in app logic. An empty folder chosen by the native dialog yields no
file entries and no folder name; use `pickMode: "fs"` when the empty folder
name is required. A dropped empty folder may yield a synthetic directory entry
in Chromium (`drag.files()` has `isDirectory: true`), but it has no file bytes
to upload.

Cross-browser folder access is intentionally limited to this file-based model.
`pickMode: "fs"` exposes a read-only directory handle, but the runtime still
selects all files in it. A directory-tree UI with per-file checkboxes would
need to be built on top of that handle and is not part of the current runtime.

### File picker dialog

Use the native dialog when dragging is not practical:

```sui
state attachment = ""

view main {
	button "Choose files…" pickFiles: attachment accept: ".csv,.txt" multiple: true
	input "Selected" value: attachment
}
```

A `file` control can show the native picker directly and store the result:

```sui
file accept: ".csv,.txt" multiple: true pickFiles: attachment
input "Selected" value: attachment
```

- `pickFiles` works on `button`, `file`, containers and fields. On a non-`file`
  element the runtime creates a hidden file input and opens it on click.
- `accept`, `multiple` and `directory` configure the dialog. On `file` they are
  native attributes; on other picker triggers they are passed to the hidden
  input.
- The selected paths use the same metadata rules and limits as drops.
  `drag.files()` exposes the selection; `drag.error()` exposes a rejection.
- A text field with `pickFiles` itself opens the dialog on click; use this only
  for a read-only/path-style field. Prefer a separate `button` when the field
  must remain editable.
- Browsers expose names and relative paths, not absolute paths. A native `file`
  control displays the selected names, but a state-driven re-render can clear
  that native display; keep a linked field (`value: state`) as the stable view
  of the selection.

## Menu bar

`menubar`, `menu` and `menuitem` build a horizontal navigation bar with
clickable dropdown menus. The labels, links and handlers are normal SUI
values, so there is no raw HTML:

```sui
menubar "Main menu" {
	menu "File" {
		menuitem "New" onClick: newFile
		menuitem "Open" onClick: openFile
		divider
		menuitem "Project site" link: "https://example.com"
	}
	menu "Edit" {
		menuitem "Undo" onClick: undo
		menuitem "Redo" onClick: redo
		menu "More" {
			menuitem "Preferences" onClick: preferences
		}
	}
}
```

| Element | Text | Useful attributes | Children |
|---|---|---|---|
| `menubar "Accessible name"` | Accessible name (`aria-label`) | `class`, `style` | `menu` |
| `menu "Label"` | Visible menu/submenu button | `class`, `style` | `menuitem`, `menu`, `divider` |
| `menuitem "Label"` | Visible item label | `class`, `style`, `link`, `onClick` | no |

The bar is a real `nav` with `role="menubar"`. A menu is a real button with a
hidden popup; clicking the button toggles it, clicking outside or pressing
Escape closes it, and Arrow keys move through open items. A `menuitem` with
`link` is wrapped in a safe hyperlink (`javascript:` and other active schemes
are rejected by the normal URL guard). A `menuitem` with `onClick` calls the
named zero-argument function like a button.

A `menu` nested inside another `menu` popup is a submenu; submenus can nest
further. Use `divider` for a separator inside a menu. All labels and links are
escaped; treat dynamic labels as untrusted text.

In the visual designer, open **Layout → Navigation** and insert **Menu bar**.
It starts with File/Edit/Help entries. Right-click a row in the element tree
for the menu editor: **Add menu**, **Add menu item**, **Add submenu**,
**Duplicate**, **Move up/down**, **Delete**. You can also drag **Menu** or
**Menu item** into the bar/menu and edit **Text**, **link** and **onClick** in
the inspector. See [GUI editor: menu bar](GUI_EDITOR.md#menu-bar-control).

## Tabs

`tabs` creates a tab bar; each direct `tab` child supplies the button label
(its text) and its panel content (its children):

```sui
tabs "Settings" {
	tab "General" {
		p "General settings"
		input "Name" value: name
	}
	tab "Advanced" {
		p "Advanced settings"
	}
	tab "About" {
		p "SimpleUI tabs example."
	}
}
```

| Element | Text | Children |
|---|---|---|
| `tabs "Accessible name"` | Accessible name (`aria-label`) | direct `tab` elements |
| `tab "Button label"` | Visible tab button | panel content |

The first tab is active initially. Clicking a tab or using Left/Right/Home/End
switches panels; the selected tab is restored after a normal view re-render.
Only direct `tab` children become tabs. The generated markup is a `role=tablist`
with `role=tab` buttons and `role=tabpanel` panels; labels are escaped normally.

In the designer, insert **Tabs** under **Layout → Navigation**. It starts with
two tabs. Right-click a `tabs` row for **Add tab**, or a `tab` row for **Add tab
below**, **Edit label**, **Duplicate**, **Move up/down** and **Delete**. Each tab
is a container: drop the panel content inside it. See
[GUI editor: tabs](GUI_EDITOR.md#tabs-control).

## Accordion

`accordion` creates stacked disclosure sections. Each direct `accordionitem`
child supplies the section label (its text) and its panel content (its
children):

```sui
accordion "FAQ" {
	accordionitem "What is SimpleUI?" {
		p "A small reactive UI language."
	}
	accordionitem "Can I use my own CSS?" {
		p "Yes, with a css block or the designer."
	}
}
```

| Element | Text | Children |
|---|---|---|
| `accordion "Accessible name"` | Accessible name (`aria-label`) | direct `accordionitem` elements |
| `accordionitem "Section label"` | Visible disclosure button | panel content |

The first section is open initially. Clicking a section header opens it and
closes the others; clicking the open header closes it. The open section is
restored after a normal view re-render. Headers are native buttons with
`aria-expanded`; labels and content are escaped normally.

In the designer, insert **Accordion** under **Layout → Navigation**. It starts
with two items. Right-click an `accordion` row for **Add accordion item**, or an
`accordionitem` row for **Add item below**, **Edit label**, **Duplicate**, **Move
up/down** and **Delete**. Each item is a container for its panel content.

## Toast notifications

Toasts are temporary feedback messages shown from code:

```sui
toast class: "my-toast-host"

fn save() -> null {
	toast.show("Saved", "success", 3000)
}

fn clear() -> null {
	toast.clear()
}
```

`toast.show(message [, kind [, duration]])`:

| Argument | Meaning |
|---|---|
| `message` | Text shown in the notification; it is always treated as text |
| `kind` | `info` (default), `success`, `warning` or `error` |
| `duration` | Milliseconds before automatic dismissal; default `4000`; `0` keeps it open |

`toast.clear()` removes all visible notifications. Users can close each toast
with its X button. By default at most four toasts are kept in a host; the `max`
attribute supports 1–8 and the oldest is removed when the limit is exceeded.

The optional `toast` element is a host:

```sui
toast class: "my-toast-host"
```

Without a host the runtime creates a fixed top-right host automatically. A
designer-placed host uses independent horizontal and vertical anchors:

```sui
toast hAlign: "left" vAlign: "bottom" hOffset: "24px" vOffset: "16px"
```

`hAlign` is `left`, `center` or `right`; `vAlign` is `top`, `center` or
`bottom`. An edge-anchored axis uses its `hOffset`/`vOffset` distance from that
edge. A centered axis ignores its offset and needs none. The host also accepts
`duration`, `max` and `icons` for runtime defaults. A top-anchored host stacks
new messages below older ones; a bottom-anchored host stacks new messages above
older ones. It uses `aria-live="polite"`, each toast uses `role="status"`, and
message content is assigned through `textContent` (never raw HTML).

Toast visuals are normal CSS on the internal parts:
`.sui-toast` (card), `.sui-toast-icon`, `.sui-toast-message` and
`.sui-toast-close`. A host class scopes them, for example
`.my-host .sui-toast { width: 420px; }`. The runtime also reads
`--sui-toast-accent`, `--sui-toast-icon-bg`, `--sui-toast-icon-fg`,
`--sui-toast-message-fg` and `--sui-toast-close-fg`.

In the designer, insert **Toast host** under **Overlay**, or add `toast` from
the palette. Selecting it opens the dedicated **Toast** panel with the two axis
selectors, conditional offset fields, runtime defaults and a live sample
preview. The normal **Size**, **Border**, **Box model**, **Typography**,
**Background** and **Effects** sections below the panel update the card rule;
the same declarations appear in the host **style expr** field, where the raw CSS
for `.sui-toast` can be evaluated and edited. Use a handler such as
`onClick: save` to call `toast.show(...)`. The **Text** field is not used for
this host. With multiple hosts, the designer previews the selected one while
`toast.show` targets the first host in the document. See
[GUI editor: toast](GUI_EDITOR.md#toast-control).

## Tooltip property

Any element or control can carry a `tooltip` attribute. The text is shown on
hover or keyboard focus; the target does not need a wrapper control.

```sui
button "Save" tooltip: "Saves the current document" onClick: save
button "Delete" tooltip: "Delete permanently" onClick: remove
```

- `tooltip` is available on every allowed element and control;
- the target should be focusable for keyboard access (for example a button,
  link or input);
- the tooltip is text-only and escaped, never raw HTML;
- leaving/blurring hides it; `Escape` dismisses the open bubble;
- the target receives `aria-describedby` while the bubble is visible;
- the bubble has `role="tooltip"` and is positioned in the viewport, so it also
  works for void controls such as `input`;
- with a tooltip open, the bubble is moved inside the viewport when possible.

### Global tooltip settings

The designer has **Preferences > Tooltip settings** for global defaults:
place (`top`, `right`, `bottom`, `left`), delay, background, text color, border
color, radius and maximum width. These are written into the project CSS as
`:root` custom properties, so they apply to every element with a tooltip:

```css
:root {
  --sui-tooltip-place: top;
  --sui-tooltip-delay: 140ms;
  --sui-tooltip-bg: #1d2129;
  --sui-tooltip-fg: #e6e6ef;
  --sui-tooltip-border: #3a4152;
  --sui-tooltip-radius: 6px;
  --sui-tooltip-max-width: 260px;
}
```

## Combobox

A searchable single-choice control. The element text is the label, `options` is
a comma-separated list, and `value` binds the selected stored value.

```sui
state country = ""
state countryOptions = ""

combobox "Country"
	options: countryOptions
	value: country
	onChange: countryChanged
	placeholder: "Search a country..."
	empty: "No results"
	maxOptionLength: 24
	maxOptions: 8
```

- `options`: comma-separated labels. It can be a literal such as
  `"Italy,France,Germany"` or a string state such as `countryOptions`. Use
  `Label=value` when the stored value differs from the displayed label.
- A dynamic `options` state is evaluated on every view render: assign it a new
  comma-separated string from a handler and the list refreshes automatically.
  This is the supported path for runtime or remote-loaded lists.
- Use a backslash to escape a comma inside an option label. At runtime the
  option data contains `Washington\, D.C.`; in SUI source the backslash is
  written twice: `"Washington\\, D.C."`.
- `maxOptionLength`: maximum accepted characters for each option label or stored
  value when the list is loaded. If any single item exceeds it, the whole list is
  rejected and an error is shown; it is not silently truncated. `0` or no value
  means no application limit (the runtime hard cap is 10,000 characters).
- `maxOptions`: maximum number of items accepted when the list is formed. If the
  loaded list is larger, the whole list is rejected and an error is shown. `0` or
  no value means no application limit; the runtime hard cap is 2,000 items.
- `value`: state binding; selecting an option updates it automatically.
- `onChange`: optional zero-argument function called after the value is updated.
  It is not called when a runtime list replacement automatically clears a value
  that is no longer present.
- `required`: marks the control as required. The search field gets
  `aria-required`; leaving it without a selected value reports an inline
  required error on focus loss.
- `placeholder`: search/display placeholder; default `Search...`.
- `empty`: text shown when filtering has no results; default `No results`.
- `disabled`: disables the input and toggle.
- Search is case-insensitive. Without wildcards it matches a substring. `*`
  matches any sequence, `?` matches one character, and `\` escapes a wildcard
  for a literal search (for example `\*`).
- Invalid runtime option data displays an error below the control and does not
  show a partially parsed list. Empty labels/values, control characters,
  dangling escapes, options over `maxOptionLength` and lists over `maxOptions`
  are rejected as a whole.
- If a valid replacement list no longer contains the selected stored value, the
  value and its bound state are cleared automatically. If the value is still
  present, it is kept while only the list changes.
- Keyboard: `ArrowDown`/`ArrowUp`, `Enter`, `Escape`, `Tab`.
- The input uses `role="combobox"`, the popup `role="listbox"`, and rows
  `role="option"` with `aria-selected`.
- Labels and empty text use text nodes, never raw HTML.

```sui
state countryOptions = ""

fn loadCountries() -> null {
	# Runtime value: Italy,France,Germany,Spain=es,Washington\, D.C.
	countryOptions = "Italy,France,Germany,Spain=es,Washington\\, D.C."
}

fn clearCountries() -> null {
	countryOptions = ""
}
```

The selected value is stored in a hidden input and the visible input is the
search field. If a runtime replacement removes the selected value, the value
and its bound state are cleared automatically; if it is still present, only the
list changes. For remote data, fetch and validate it first, then assign the
comma-separated string to the state.

### Runtime update security

- Treat the options string as untrusted data when it comes from an API, file or
  storage. The control renders labels with `textContent`, and generated dynamic
  attributes are HTML-escaped, so it cannot inject markup.
- The parser rejects malformed escapes, empty labels/values, control characters,
  oversized options (over 10,000 characters), lists over 500,000 characters and
  more than 2,000 options. The error is generic and never echoes the payload.
- Wildcard search uses an iterative glob matcher, not user-supplied regular
  expressions, avoiding regex injection and catastrophic backtracking. Queries
  are capped at 256 characters.
- `maxOptionLength` and `maxOptions` are validation limits: the complete
  candidate list is checked before it is published. They are not a substitute
  for response-size and schema limits on remote data.
- Automatic clearing of a removed selected value is a state mutation. It calls
  the generated setter with `""` and does **not** call `onChange`. Avoid
  publishing transient empty or partial lists during loading; stage the list
  behind a loading state, then publish the complete list.

## Tree

A nested, keyboard-accessible tree. `tree` supplies the visible title and
selected value binding; each `treeitem` supplies a label, a stable `key`, and an
optional initial `open` state.

```sui
state selected = "main"
state renameKey = ""
state renameValue = ""

tree "Project files"
	name: "files"
	value: selected
	onChange: selectedChanged
	editable: true
	editableDepth: 2
	renameKey: renameKey
	renameValue: renameValue
	onRename: renamed {
	treeitem "src" key: "src" open: true {
		treeitem "main.go" key: "main"
		treeitem "components" key: "components" {
			treeitem "tree.go" key: "tree"
		}
	}
	treeitem "README.md" key: "readme"
	treeitem "generated" key: "generated" editable: false
}
```

- `name` is the runtime target used by the core `tree` module. It defaults to
  the visible tree title; set a unique name for data import/export.
- `value` binds the selected item key to a string state.
- `onChange` runs after the selected state has been updated.
- `treeitem key`: stable value stored in the state. If omitted, the item path is
  used.
- `treeitem open: true`: initial expansion state for a parent.
- `editable: true` enables inline rename; `editableDepth: 2` allows levels 1-2.
  `0` means all levels.
- `treeitem editable: true|false` overrides the tree default. `false` locks the
  node and its descendants. No attribute means inherit.
- `renameIcon: "✏️"` changes the pencil glyph; `showRenameIcon: false` hides it.
  Both are text, never HTML. `treeitem renameIcon`/`showRenameIcon` override the
  tree value; omit them to inherit. Hiding the button leaves F2 and
  double-click rename available.
- Rename: double-click, F2 or the pencil button; Enter commits, Escape cancels; empty labels are
  rejected. The key never changes.
- `renameKey`/`renameValue` receive the node key and new label before
  `onRename` runs.
- Clicking a row selects it; clicking the disclosure arrow only expands or
  collapses it.
- Expansion is independent from selection and is persisted per tree path in
  `localStorage`. Renamed labels are also stored locally per tree/path.
- Keyboard: Up/Down move through visible items; Right expands or enters a
  parent; Left collapses or returns to the parent; Enter/Space selects; F2
  renames an editable item; Home/End jump.
- ARIA: `role="tree"`, `role="treeitem"`, `role="group"`, `aria-level`,
  `aria-expanded`, `aria-selected` and `aria-readonly`. Only one item is in the
  tab order.
- Visual styling is carried by `style expr` as CSS custom properties on the tree
  root. Tree presets: **Default**, **Compact**, **Comfortable**, **Dense**,
  **Bordered**, **Soft rounded**, **Minimal**, **Neon**, **Forest**, **Amber**,
  **Ocean**, **High contrast**, **Outline**, **Mono**, **Candy** and **Cards**;
  selecting one rewrites `style expr`.
- Each `treeitem` also exposes node presets: **Inherit tree style**, **Muted**,
  **Accent**, **Success**, **Warning**, **Danger**, **Folder** and **Compact row**.
  A node preset writes row-level variables into that item's `style expr` and
  overrides only that node; **Custom** edits the expression directly.
- The preset variables include `--sui-tree-gap`, `--sui-tree-row-padding`,
  `--sui-tree-row-radius`, `--sui-tree-row-bg`, `--sui-tree-hover-bg`,
  `--sui-tree-selected-bg`, `--sui-tree-selected-fg`, `--sui-tree-title-color`,
  `--sui-tree-title-size`, `--sui-tree-panel-bg`,
  `--sui-tree-panel-border`, `--sui-tree-panel-radius`,
  `--sui-tree-panel-padding` and `--sui-tree-indent`.
- Edit the generated `style: cssSafe("...")` expression directly for custom
  values. Dynamic style stays under the normal `cssSafe` policy. Because
  `cssSafe` accepts comma-free declarations, use 8-digit hex colors such as
  `#4c5cf761` instead of `rgba(...)` inside `style expr`.
- Data: `tree.importJson`/`tree.importXml` replace nodes from text;
  `tree.importJsonFile`/`tree.importXmlFile` read a selected file;
  `tree.exportJson`/`tree.exportXml` return text; `tree.downloadJson`/
  `tree.downloadXml` download a file; `tree.clear` removes all nodes. See
  [`STDLIB.md`](STDLIB.md#tree-module-tree-core) for schemas and limits.
- Imported data is applied to the tree with the matching `name`; it survives
  re-renders and replaces only the node structure (title/style/bindings stay).
  Labels, keys and icons are written as text, never HTML. XML DTD/entities are
  rejected.
- Labels, keys and item text are escaped text; no raw HTML.

Keys should be unique within a tree. The selected key is plain application data;
validate it before using it to access server resources. `editable` is a UI hint,
not an authorization boundary. Rename labels are stored in `localStorage`; use
`renameKey`/`renameValue` and `onRename` to persist them in the application or
server model.

## Chart

`chart` renders JSON, XML or CSV data as an SVG bar, line, area, pie, donut or
scatter chart. The optional element text is the visible title. Data is text,
never markup, and every SVG label is written with `textContent`.

```sui
state sales = "[\{\"month\":\"Jan\",\"value\":12\},\{\"month\":\"Feb\",\"value\":18\}]"

chart "Monthly sales"
	name: "sales"
	data: sales
	chartType: "bar"
	format: "auto"
	x: "month"
	y: "value"
	overflow: "scroll"
	pointWidth: 48
	showGrid: true
	showLegend: true
	showValues: false
```

`width` and `height` are optional. When omitted, empty, `auto` or a
percentage, the chart fills its container and follows it when the container
grows; the table keeps its fixed pixel height while the graphic area grows in
both directions. An explicit pixel size keeps a fixed frame. In a container
without a definite height the chart falls back to at least 240 px
(`--sui-chart-min-height` overrides the fallback), so it never collapses.

`theme: "dark"` (default) or `theme: "light"` selects the built-in token set.
The light theme defines explicit `--sui-chart-*` values for background, text,
grid, controls, tooltip, table and scrollbars. Dialog overlays, the chart menu,
context menus, table settings/filter popups, the resource picker and the workload
dialog copy the active theme tokens, so they stay readable in both dark and light
charts. Focus and selection tokens
(`--sui-chart-table-selection-bg`, `--sui-chart-table-selection-border`,
`--sui-chart-table-focus-row-bg`, `--sui-chart-table-focus-cell-bg` and
`--sui-chart-table-focus-color`) can be overridden independently: selection is
shared while focus row/cell tones have theme-aware light defaults. Clicking a
series color swatch in a table header or in the legend opens a color picker; the
filled legend swatch shows the exact color used by that series, and the choice
updates the corresponding `--sui-chart-series-N` token immediately, is
stored per chart in `localStorage`, and does not rebuild the table or redraw the
SVG. `colorPicker: false` keeps the swatches visible but disables editing
(default `true`). For block diagrams, `settingsButton: false` hides the header
gear, `showZoom: false` hides the zoom controls, `showAdd: false` hides the
**+ Add** button and `showHistory: false` hides undo/redo. A chart with more than one legend color always shows the
legend, even when `showLegend` is false, so each color remains identifiable.
Gantt task rows show a color swatch next to the task name; changing it updates the
bar and table. Gantt legend entries represent groups or individual tasks when
their colors differ. A task may also provide an initial `color` field in JSON,
XML or CSV data; user choices are stored separately per task id. Chart updates crossfade the previous SVG into the new one unless
`prefers-reduced-motion` is active; reactive view refreshes preserve the chart
and table DOM and redraw only charts whose chart attributes changed.

`type` is a reserved SUI word, so the chart geometry attribute is
**`chartType`**. Its values are:

| `chartType` | Geometry | Series |
|---|---|---|
| `bar` | Vertical bars with a zero baseline (default). | one |
| `barhorizontal` | Horizontal bars; ideal for long labels. | one |
| `barstacked` | Bars stacked by series. | two or more |
| `bargrouped` | Series side by side in each category. | two or more |
| `line` | Connected line and points. | one |
| `step` | Step line with horizontal/vertical segments. | one |
| `smooth` | Smooth cubic line. | one |
| `area` | Line with a translucent fill down to zero. | one |
| `combo` | First series as bars, remaining series as lines; all share one value scale. | two or more |
| `pie` | Slices from label/value pairs. | one |
| `donut` | Pie with a central hole. | one |
| `scatter` | Numeric x/y points; labels may be numeric. | one |
| `histogram` | Distribution of numeric values in `bins` buckets. | one |
| `waterfall` | Cumulative increases/decreases. | one |
| `radar` | Polygon comparison across categories. | one or more |
| `gauge` | KPI arc with `min`/`max`; uses the first value. | one |
| `funnel` | Conversion stages with proportional widths. | one |
| `heatmap` | Intensity over x categories and numeric y rows. | exactly two: y and value |
| `candlestick` | OHLC candles. | exactly four: open, high, low, close |
| `treemap` | Flat area-proportional treemap. | one positive |
| `boxplot` | Box and whisker per category. | exactly five: min, q1, median, q3, max |
| `bullet` | KPI value against a target and optional max. | two or three: value, target, max |
| `dumbbell` | Before/after dots connected per category. | exactly two |
| `gantt` | Horizontal start/end tasks. | exactly two: start, end |
| `state` | Legacy block-diagram geometry (`statemachine` alias); prefer the dedicated `graph` control. | states/transitions object or edge list |

### Data

`data` is a literal string or a reactive expression (usually a `str` state).
`format` is `auto` (default), `json`, `xml` or `csv`. Auto-detection uses the
first non-space character: `{`/`[` is JSON, `<` is XML, anything else is CSV.

JSON accepts an array of objects, an array of arrays (with an optional first
header row), `{ "labels": [...], "values": [...] }`, `{ "rows": [...],
"columns": [...] }`, an array of numbers (useful for histograms) or a simple
object map. `x` selects the label column. `y` selects one numeric column; when
omitted, every other numeric column becomes a separate series. A comma-separated
`y` such as `"sales,cost"` selects multiple series.

```sui
chart "Trend" data: "Month,Value\nJan,12\nFeb,18\nMar,15" format: "csv" x: "Month" y: "Value"
chart "Rows" data: "<data><row label=\"Jan\" value=\"12\"/><row label=\"Feb\" value=\"18\"/></data>" format: "xml"
chart "Stacked" data: "Month,Sales,Cost\nJan,12,5\nFeb,18,8" chartType: "barstacked" x: "Month"
chart "Smooth" data: "Month,Value\nJan,12\nFeb,18\nMar,15" chartType: "smooth" x: "Month" y: "Value"
chart "Histogram" data: "[3,5,5,6,7,8,8,9,12,14,15,18]" chartType: "histogram" bins: 6
chart "Waterfall" data: "Step,Delta\nStart,20\nSales,-5\nCosts,-3\nFinal,8" chartType: "waterfall" x: "Step" y: "Delta"
```

CSV supports comma, semicolon, tab and pipe delimiters plus quoted cells. A
header row is detected from text cells that sit above numeric cells. XML accepts
`<row label="..." value="..."/>`, `<row><label>...</label><value>...</value></row>`,
`x`/`y` attributes and nested elements. For multiple series, XML rows can carry
additional numeric attributes (`<row label="Jan" sales="12" cost="8"/>`) or
numeric child elements. Pie/donut charts use labels and positive values; XML and
JSON DTD/entity declarations are rejected.

Heatmap rows use the label column as x, the first numeric series as the numeric
row (y) and the second numeric series as the cell value. Candlestick rows need
the four numeric columns in this order: `open`, `high`, `low`, `close`.

Gantt data can use numeric positions or ISO dates and these fields:
`id`, `task` (or `label`/`name`), `start`, `end`, `duration`, `progress`,
`milestone`, `group`, `parent`, `dependsOn`/`dependencies`, `dependencyType`,
`lag`, `constraint`, `constraintDate`, `deadline`, `resource`/`resources`, `status`, `color`, `cost`,
`actualStart`, `actualEnd`, `remaining`, `work`, `workDone` and `resourceUnits`.
`resources` accepts an array or a comma/semicolon separated list; array entries
may be strings or `{ "name": "...", "units": N }` objects. `resourceUnits` is an
optional name-to-quantity map. The legacy `resource` string is split into the
same list, with quantity `1` by default. Date axes choose hour, day, week, month or
year ticks automatically; weekends and holidays are shaded and a `Today` line
appears when the current date is in range. `progress` draws an inner completion
bar, `milestone` renders a diamond and `dependsOn` draws dependency arrows.
Dependency lines use a neutral graphite color; critical links are thicker but
not yellow. The runtime tests each route against all task bars and falls back to
outer lanes when a segment would cross a bar. With
`overflow: "scroll"`, dense Gantt layouts grow vertically and scroll so rows keep
enough spacing for readable bars and connections.
Gantt charts also expose **timeline zoom** controls (`−`, `Timeline N%`, `+`,
`Reset`) in the chart header. Zoom changes the time axis width only; task labels,
rows and text keep their size. When the resulting timeline is wider than the
chart frame, the dedicated viewport shows a horizontal scrollbar. `zoom` sets the
initial timeline scale (0.5–24); `ganttScale` (`auto`, `day`, `week`, `month`,
`year`) selects the tick granularity independently. Weekly ticks and week bands
start on Monday, and the weekly scale does not render non-working-day bands. At
high zoom the auto scale shows day ticks with weekday names and a year label.
The chart frame keeps its
assigned size. The left task-name column has a fixed width and stays left-aligned
even when the scaled timeline is narrower than the frame; it remains sticky while
the timeline scrolls or zooms. Long task and resource labels wrap onto up to
three lines inside the fixed name column; the synchronized Task column also
wraps long text instead of clipping it. Multiple resources are shown joined by
commas in the chart/tooltip and as separate chips in the table. `progressColumn: true` adds a fixed
percentage column between the task names and the timeline and hides the inline
percentage labels beside the task bars. Drag the border at the right of the
task-name column to resize it; the width is persisted per chart. When the
timeline scrolls horizontally, axis tick labels stop at that separator instead
of sliding under the task names. Deadline markers are larger and draggable
horizontally; the table Deadline field updates on drop.
Dependency routing evaluates several orthogonal lanes, rejects paths that cross
task bars, avoids overlapping existing horizontal/vertical link segments, and
chooses the shortest remaining route measured as horizontal + vertical length.
Vertical runs keep the configured `linkGap` from task borders and from other
vertical link runs; only the shared source/entry port is allowed to coincide.
All links leave a task from its right edge and enter the target from its left
edge, regardless of SS/SF/FF semantics. A side exit starts at the centre of the
side with a horizontal stub of at least 4px; that stub may be shared by several
outgoing links. One exception applies to a collapsed
group: when an internal child ends before the group summary end, the outgoing
link leaves the **bottom edge of the summary bar at the child's finish x**, never
from the group end. All links leaving one task share the
same exit port, and all links entering one task share the same entry port; that
port coincidence is the only allowed overlap. The arrow tip is centred on the target's left edge; a horizontal tail
of at least 4px starts at the back of the arrowhead, and all incoming routes
meet at the left end of that tail. Routing is clamped to the plot area, so
dependency lines never run underneath the fixed task-name column. Lanes may
travel across the plot area, but every vertical run keeps `linkGap` from task
borders and from other vertical runs, and the arrow-tail entry still lies on the
target's left edge.
Vertical runs are monotonic in the target row direction: a link to a lower task
never rises, and a link to an upper task never descends. The initial downward
stub of a collapsed-group bottom exit is exempt from that rule. When the requested
`linkGap` needs more room, the Gantt stage grows vertically and, when parallel
vertical lanes need space, horizontally; the viewport scrolls, preserving the
gap instead of squeezing the route. Horizontal runs are
kept close to the source and target rather than immediately
using the far left/right edge of the chart. Moving or resizing a task triggers a throttled live redraw: links are
re-routed, scheduling warnings are re-evaluated, and the table stays
synchronized; the history entry is committed on drop.
Every Gantt redraw validates task dates, actual dates, remaining/work values,
constraints, deadlines and each dependency's type/lag. Violations appear in the
warning banner and on the affected table rows, so every edit is checked
immediately. Dragging or resizing snaps a start, finish or milestone to a
working date; writing a non-working date in the table is rejected with an
inline error and a table status message. With the pointer over the
timeline, **Ctrl/Cmd + wheel** (or a trackpad pinch, which browsers report as
Ctrl + wheel) also zooms the time axis. On touch screens, a **two-finger pinch**
does the same; it does not trigger the browser page zoom.
`parent` points to another task id: tasks with children become WBS summary tasks.
This changes the parent into a summary container. To keep every task as a real
task with its own bar and links, use `group` instead: it is only a visual phase
label and does not alter the WBS parent. Tasks with the same `group` share a
group header; group membership can also be assigned to several selected rows
from the table context menu using **Assign to group...**. Select tasks with
Ctrl/Cmd + click in the chart, or select table rows, then choose **Create group
from selection**. This inserts a new summary node named **New group** just above
the first selected task and sets that node as the WBS parent of every selected
task; their bars and links are preserved.
whose start, end, progress and cost roll up from their descendants. Children are
indented to the right of their parent in both chart and table; summary/group
labels use a highlighted header band and bold text. An expanded summary is a
header row without a bar; collapsing it replaces the children with one summary
bar and external dependencies stay attached to it. Click the collapsed bar,
its name, the disclosure caret or the table toggle to expand it again.
`dependsOn` may contain task
ids or task names, and a field can list several predecessors. Each entry can be
a string (`"A"`, `"A[SS+2]"`) or an object
(`{"id":"A","type":"SS","lag":2}`), allowing a different link type and lag per
predecessor; task-level `dependencyType`/`lag` remain the fallback. The runtime
resolves names to ids, rejects dependency cycles when editing, and shows an
explicit cycle error for imported cyclic data. Older two-column numeric
`start`/`end` data remains supported.

Gantt calendar attributes are chart-level:

- `weekend: "sat,sun"` (default) selects non-working weekdays by name or number
  (`0` = Sunday through `6` = Saturday); `weekend: "none"` disables weekend
  bands.
- `timezone: "Europe/Rome"` uses an IANA time zone for axis ticks, table dates,
  tooltips and calendar-day matching when the data contains date-time values.
  Date-only values remain stable calendar dates. The default is `UTC`.
- `showNonWorkingDays`: default `true`. Shows or hides weekend/holiday bands in
  the timeline without changing scheduling, effective duration or non-working
  date validation. At week scale these bands are omitted regardless of the
  setting; the alternating week bands remain visible.
- `holidays` accepts exact dates (`2026-12-25`), inclusive ranges
  (`2026-12-24..2026-12-26`) and yearly recurring dates (`*-12-25` or `12-25`).
  Recurring ranges such as `*-12-24..*-01-06` wrap across the year boundary.

The Gantt chart menu includes **Calendar...**: a dedicated overlay lets the user
choose non-working weekdays and add holiday dates or closure ranges through one
date-range picker (click once for a single day, twice for a range). The calendar
is stored per chart in `localStorage` (`sui.chart.calendar.<chart-key>`), can also
be supplied in JSON data as `{"tasks": [...], "calendar": {"weekend":
"sat,sun", "holidays": ["2026-12-25", "2026-12-24..2026-12-26"]}}`, and updates
weekend/holiday bands plus **Effective duration** immediately. An optional
`onCalendarChange` handler receives the saved calendar JSON. The same chart menu
toggles **Show/Hide progress column** and **Show/Hide non-working days** for the
session.

Alternating vertical period bands follow the active timeline scale: months at
month scale (`ganttScale: "month"` or auto), Monday-based weeks at week scale,
and calendar days at day scale. The bands use theme-aware colors in both dark
and light mode and sit behind weekend/holiday bands and bars. At day scale, a
non-working day uses the weekend/holiday color instead of the alternating day
color whenever non-working days are shown.

Gantt link appearance is configured with chart-level attributes:

- `linkColor`: CSS color for dependency lines; default neutral `#94a3b8`.
- `linkWidth`: line width in pixels (`0.5`-`8`), default `1`; critical links
  use a slightly thicker value.
- `linkGap`: minimum clear distance in pixels between dependency segments and
  other links, task/milestone borders or arrowheads; vertical runs always keep
  this clearance from bars and from other vertical link segments. Default `4`,
  range `0`-`24`.
- `linkStyle`: `auto` (link-type dash pattern), `solid`, `dashed` or `dotted`.
- `arrowType`: `triangle`, `open` or `none`.
- `arrowSize`: arrowhead size in pixels (`2`-`20`).
- `progressColumn`: default `false`; adds a fixed percentage column between the
  task-name column and the timeline. When visible, it replaces the inline
  percentage labels next to task bars.
- `barBorderColor`: optional CSS color for Gantt task, summary and milestone
  borders; empty/omitted keeps the theme border.

A complete runnable demo is in
[`examples/gantt_calendar/`](../examples/gantt_calendar/README.md); it switches
each attribute at runtime and shows the resulting table dates and bands.
[`examples/gantt_wbs/`](../examples/gantt_wbs/README.md) demonstrates parent/child
summary tasks, collapsed summary bars, auto scheduling and dependency cycles.

### Graph control (block diagrams)

The dedicated `graph` control renders a block/state diagram. It always uses the
shapes/links geometry, so no `chartType` attribute is written:

```sui
state flow = "\{\"states\":[\{\"id\":\"input\",\"label\":\"Input\"\},\{\"id\":\"process\",\"label\":\"Process\"\},\{\"id\":\"output\",\"label\":\"Output\"\}],\"transitions\":[\{\"from\":\"input\",\"to\":\"process\"\},\{\"from\":\"process\",\"to\":\"output\"\}]\}"

view main {
  graph "Pipeline" name: "pipeline" data: flow
}
```

`graph` accepts the same data, layout, grid, tag/link styling, table and runtime
attributes as the block-diagram chart geometry. The legacy
`chart "..." chartType: "state"` form (alias `statemachine`) keeps working for
existing documents, but the designer no longer lists it in the chart type menu.
Changing a chart to a type whose data shape is incompatible shows a confirmation
dialog and, if accepted, replaces the data with a sample suited to the new type.

The primary JSON form is:

```json
{
  "states": [
    { "id": "idle", "label": "Idle", "shape": "circle", "w": 48, "h": 48, "description": "Waiting" },
    { "id": "choice", "label": "Payment?", "shape": "rhombus", "w": 96, "h": 96 },
    { "id": "done", "label": "Done", "shape": "circle", "w": 48, "h": 48 }
  ],
  "transitions": [
    { "id": "edge_1", "from": "idle", "to": "choice", "event": "submit" },
    { "id": "edge_2", "from": "choice", "to": "done", "event": "capture", "guard": "payment.ok", "action": "reserveStock()" }
  ]
}
```

Every state must declare an explicit `id`; it must be unique inside the chart.
Missing or duplicate ids stop parsing with a clear error. Link ids are optional
and default to `edge_1`, `edge_2`, ...; the runtime still checks that they do not
collide with shape ids. Runtime functions accept only the explicit `id`.

State fields are `id`, `label`/`name`, optional `description`, `hint`, `color`,
`shape`, border/text-tag settings and normalized/pixel `x`/`y`/`w`/`h`. A
non-empty `hint` replaces the default hover tooltip; `"none"`, `"off"` or `"-"`
hides it for that element. In `hint` and `tag`, `|` starts a new line and `\|` is
a literal pipe (`\\` a literal backslash); the editors are single-line. Real
newlines are accepted as legacy input and normalised to `|`. The chart attribute
`stateHints: false` hides all hover hints. The **Hint** column in the state table
and the context menu **Hint settings...** edit the same field. The chart settings
panel has a standalone **HINTS** panel that styles every tooltip globally
(enable/disable, background and opacity, text/border color, font and size,
border width, padding, radius) with a live preview. Source defaults use
`stateHintBg`, `stateHintOpacity`, `stateHintTextColor`, `stateHintBorderColor`,
`stateHintFont`, `stateHintFontSize`, `stateHintBorderWidth`, `stateHintPadding`
and `stateHintRadius`. `tableView` selects `chart`,
`table` or `both`; when the table is visible, selecting an element in the diagram
focuses and highlights its row. Legacy `type` values are accepted as metadata
but do not change geometry or add role markers; use `shape` for the visible
geometry.

Shape fields:

- `shape`: `rectangle`, `square`, `circle`, `stadium`, `ellipse`, `rhombus`,
  `triangle`, `parallelogram`, `hexagon`, `trapezoid`, `pentagon` or `note`;
- `borderStyle`: `solid`, `dashed` or `dotted`;
- `borderWidth`, `borderColor` (the border is always a single stroke);
- `cornerRadius` rounds rectangle corners and the vertices of polygon shapes
  (rhombus, triangle, parallelogram, hexagon, trapezoid, pentagon, note);
- `rotation` in degrees for shapes whose orientation matters (triangle,
  parallelogram, trapezoid, pentagon, note, rectangle, ellipse and stadium).
  Anchors and their normals rotate with the shape. Selecting a shape shows a
  rotation handle, and the **Position** properties tab exposes the exact angle;
- `textPlacement`: `inside`, `outside` or `tag`;
- `tag` for an additional note/comment tag and `tagSide` for its side.

Transition fields are `id`, `from`, `to`, `label`, `event`, `guard`, `action`,
`description`, `color`, `routing`/`linkType` (`straight`, `arc`, `bezier`,
`orthogonal`, `circular`), `width`, `dashStyle`, `startMarker`, `endMarker`,
`fromPort`, `toPort`, optional `waypoints` and a `tag` note. Missing transition
ids are generated and preserved in the local payload.

JSON also accepts an array of states with nested `transitions`, or an array of
edge objects with `from`/`to`. CSV accepts an edge list with
`from,to,label,event,guard,action` headers, or state rows with columns such as
`id,label,type,x,y,description`. XML accepts `<state>` nodes with optional nested
`<transition>` children, or sibling `<transition from="..." to="...">` elements.
Unknown transition endpoints create inferred states from their ids. JSON export
preserves the complete shape/link/tag model; CSV exports the link fields including `hint`.

Rendering:

- `shape` selects the visible geometry; when omitted it defaults to a rectangle.
  `stadium` is a rectangle with two
  semicircular ends; `triangle` is an upward isosceles triangle with anchors on
  its three edges and bisector directions at its vertices. `parallelogram`
  (input/output), `hexagon` (preparation), `trapezoid` (manual operation),
  `pentagon` (off-page connector) and `note` (folded top-right corner) are
  polygon shapes. Their anchors are distributed uniformly along each edge
  (vertices plus evenly spaced interior points) and rotate together with the
  shape, avoiding uneven clusters on slanted sides;
- fill, border style/width/color and corner radius are shape properties; the
  border is always a single stroke. There are no initial/final/choice role
  markers or special shapes;
- shape labels, descriptions and names are not rendered on the diagram; only
  the explicit `tag`/note is drawn, with its leader line/arrow. The element name
  remains in the table, inspector, tooltip and accessibility label;
- the grid uses `stateGridGap` as its cell size and as the minimum line-to-line
  distance; `stateGridColor` sets its color. It is optional: hide it with
  `stateGrid: false` or from the single **GRID** panel (visibility, size and
  color); without the grid the canvas stays uniform. Default gap is `24` px.
  Anchors and editing handles are optional too: `stateHandles: false` (or the
  **Diagram settings** switch) hides anchor dots, resize/rotation grips, tag and
  route handles and the floating **+ Add** tool, and shapes cannot be dragged, so
  the diagram is view-only. Handles are enabled by default; set
  `stateHandles: false` to hide them;
- the diagram board is at least as large as the visible viewport, so the grid and
  the usable canvas fill the whole view. The grid is painted on the viewport and
  follows scrolling and panning, so it stays continuous even when the view moves
  beyond the board. Enlarging the viewport automatically
  enlarges the board; the board also grows when shapes, tags or waypoints move
  beyond the current border. The drawn content keeps a 20 px margin. **Fit width**
  fits the rendered content horizontally and keeps vertical scrolling, while
  **Fit all** fits both dimensions and centers the graph; neither clips the
  content;
- circle centers and all shape vertices are snapped to grid vertices; rectangle/
  square/stadium boxes are grid-sized and grid-aligned, rhombus/ellipse centers
  and extremes are grid-aligned, and resizing keeps the same rule;
- every interior horizontal/vertical segment of an orthogonal link lies on a
  grid line and is at least `stateGridGap` long; the first/last attachment stub
  may start on an exact (possibly off-grid) curved boundary;
- link anchors are computed on the actual shape boundary. For circles, ellipses,
  rhombus edges and stadium caps one coordinate comes from the grid and the
  other is solved on the curve, so the link starts exactly on the circumference
  instead of using a rounded approximation. Anchors are named deterministically
  (`V<col>T/B`, `H<row>L/R`; fallback `P:<side>`) and are shown only when the
  element is selected and editing handles are enabled;
- non-self transitions start and end at one of those anchors. Orthogonal routes
  leave the source perpendicularly and enter the destination perpendicularly for
  at least `stateGridGap` pixels before following the A* corridor. When the
  anchor lies on a curved or oblique surface (circle, ellipse, stadium cap or
  rhombus edge), that first/last stub follows the true surface normal: it may be
  diagonal, is not forced onto the orthogonal lattice, and is exempt from the
  orthogonal-segment rule. The rest of the polyline remains orthogonal;
- link geometry: `straight`, single `arc`, `bezier`, `orthogonal` polyline or
  `circular` (a circular arc); width, dash and independent start/end markers
  (`none`, `triangle`, `open`, `circle`, `square`) are per-link;
- the endpoint handles can be dropped on **any** shape, not only on the shape the
  link started from: dragging a link start/end onto another shape reattaches that
  end. A self-loop always keeps two **distinct** attachment points on the same
  shape, even when it is created from the table;
- link properties include **Direction**: choosing the reversed direction swaps
  source/target, their anchors and ports, reverses manual waypoints and mirrors
  the tag position, so the arrowhead follows the new target;
- the **LINK** panel mirrors the per-link panel with two tabs: **Line** (subtype,
  line/highlight colors, width slider, dash type/length/gap, circular arc radius
  and editable routes) and **Arrows** (arrow size, start/end marker and a live
  arrow preview). These are the defaults inherited by every link; a per-link
  **Start marker**/**End marker** can still override them (**default** inherits
  the chart setting);
- link names, events, guards and actions are not rendered on the diagram; only
  the explicit link `tag`/note is visible. They remain editable in the table and
  inspector and available in the tooltip/accessibility label;
- link properties include **Direction** (reverse source/target) and inherit the
  chart's **LINK** configuration: a new link stores no style field. Opening its
  properties shows the effective defaults; changing a field stores an override
  for that link only (choose **default**/**auto** to inherit it again), and
  **Reset style to defaults** clears all style overrides;
- tags have an independent family configuration for shapes and links
  (`stateShapeTag*` and `stateLinkTag*`). Each element can override the defaults
  with `tagKind`, `tagBox`, `tagAlign`, `tagLeader`, `tagArrow`, `tagArrowSize`,
  `tagFont` and `tagFontSize`, plus an attachment position (`tagAnchor` for
  shapes, `tagAnchorT` for links):
  - **external** tags sit outside the element and can be joined by a straight or
    Bézier leader (`tagLeader`: `line`, `bezier`) with an optional arrow
    (`tagArrow`). Box styles are `plain` (no border), `rectangle`, `rounded`,
    `square`, `circle`, `ellipse` and `line` (text above a line);
  - **internal** shape tags are centered inside the shape, do not rotate with
    the shape and stay selectable/editable;
  - **internal** link tags interrupt the line with a background mask, support the
    same box styles, and either follow the local line direction or stay
    horizontal (`tagAlign`: `line`, `horizontal`);
  - the orange attachment dot shown on a selected element moves the point where
    the leader/line touches the element. Shape contacts are stored as a
    normalized boundary point (`tagAnchor: {rx, ry}`) and survive resizing and
    rotation; link contacts are stored as a fraction along the path
    (`tagAnchorT`);
  - the leader is always aligned with the outward normal at the contact point:
    it starts on the real border of the chosen box (rectangle side,
    circle/ellipse circumference) or on the underline for the `line` style, and
    ends exactly on the shape perimeter or on the link path. The box is oriented
    so its facing border is perpendicular to the leader (on slanted shapes the
    tag content rotates with the normal; circular boxes keep the text upright
    because any radial leader is perpendicular). The arrow tip sits on that
    contact point, and shape anchors are taken from the real rotated boundary,
    so a rotated triangle/polygon receives the tip exactly on its edge. The tag
    can be dragged anywhere. A straight leader uses the minimum number of
    segments: one when the tag lies on the element normal (also the default
    placement for a tag that has not been positioned yet), otherwise three
    (perpendicular exit from the element, one connector, perpendicular arrival at
    the box). A Bézier leader is split into two smoothly joined cubic sections
    whose end tangents are perpendicular to the box and to the element;
  - `tagRotation` rotates the whole tag (box and text). When it is `null`/unset
    the tag follows the automatic orientation; a manual angle selects the box
    side facing the element and re-routes the leader so it remains perpendicular
    at both contacts. The family defaults are `stateShapeTagRotation` and
    `stateLinkTagRotation`; the per-element and family TAG overlays expose **Tag
    rotation** (Auto/Manual) and **Rotation (deg)**. Selecting an element also
    shows a rotation handle on its tag: dragging rotates the tag around its
    centre, switches it to manual and `Shift` snaps to 15°;
  - automatic external tag attachment: shapes use the midpoint of the highest
    side; a circle/ellipse uses its topmost point; an oblique polygon uses the
    midpoint of the more horizontal of the two top sides, or the top vertex when
    their slopes match. Links use the midpoint of the highest horizontal segment;
    a vertical link uses its midpoint with the tag to the side. The text starts
    two line heights away along the outward normal and moves farther only when it
    would overlap another shape, link or tag. The automatic anchor is chosen once
    and cached (normalized for shapes, a path fraction for links), so later
    redraws keep the same relative point and the board does not oscillate;
  - tag leader polylines never contain an interior angle below 90°: candidates
    with acute bends are rejected and the router adds an orthogonal detour until
    every turn is at least 90°. Bézier leaders keep their control arms within
    35% of the chord and are split into two smoothly joined sections, so they
    cannot develop a kink sharper than 90°;
  - arrows are proportional to the grid gap by default: length and base are
    `stateGridGap / 4`, so a 20 px gap gives a 5 px arrow. `arrowSize` overrides
    the between-shapes link arrows, `stateShapeTagArrowSize` /
    `stateLinkTagArrowSize` override the tag families and each element can set
    `tagArrowSize`. All of them are editable in the TAG panel;
  - the **TAGS** sub-panel in the settings **General** tab opens a single overlay
    for the shape and link families (default type, box, text direction, leader, arrow size, font
    and movable tags) with a live preview. The **Tag settings…** button in
    shape/link properties (also in the context menu) opens the per-element
    overlay; the old tag submenus were removed so tag configuration lives in one
    place. Controls are organized into tabs (**Tag**, **Text**, **Box** and
    **Leader & arrow**), so the panel height follows the largest tab: on/off
    options are switches, numeric values use a slider paired with an editable
    number field, and the font is a searchable combo box. The font field shows the
    current or default font and the list always contains every available family,
    including the selected one. Applying the per-element tag overlay stores only
    the fields that changed, so untouched properties keep inheriting the family
    defaults. The **Text** tab also exposes the multiline tag
    text, font size, weight, style, decoration, alignment, letter spacing, line
    spacing, text transform, text shadow and shadow color. A **Text settings…**
    command in every table row and in every element context menu opens this dialog
    directly on the **Text** tab. Rotation
    controls live in the **Tag** tab. Colors use a compact swatch plus a hex edit
    field, and the box background also offers an explicit **Transparent** button.
    The **Box** tab contains the box style, background, border color, border width
    (minimum 1 px), border style and the padding between text and border.
    **No border** still keeps the box style, background and padding selectable and
    renders a borderless box with the chosen background; only the border color,
    width and style controls are disabled. The **Arrow** switch enables or
    disables the arrow shape and size controls. **Movable tags** lives in the
    **Tag** tab: with an internal shape tag the tag can only move inside the shape,
    and with an internal link tag only along the link path. In the preview you can
    drag the tag and use the round handle to rotate it: rotation snaps to
    0/90/180/270° within 6° and is written back to the rotation field, while the
    dragged position is local to the preview. **Text direction** is enabled only
    with manual rotation: **Always horizontal** keeps the text level and disables
    the rotation value, while **Follow line** uses the manual angle. An internal
    shape tag defaults to a transparent background with no tag box and has its
    leader/arrow controls disabled; an internal link tag sits on the link line,
    interrupts it visually and is never above or below it. The link sample is
    always a straight line and uses the same color, width, dash and arrow marker
    as the live graph. The preview leader is built with the same polyline/Bézier
    functions, arrows and start/end normals as the diagram;
- hovering a shape or link always shows its name, description and optional tag,
  even when the diagram tag is absent. Tag labels reuse the same tooltip;
- **Description** and **Tag / note** in the properties panel are single-line
  fields; in tags write `|` for a line break and `\|` for a literal pipe;
- self-loops are normal links between two **distinct** attachment points of the
  same shape (never the same point). They can be created from the canvas (in link
  mode click the same shape twice) and from the table. The selected link subtype
  applies here too: `straight`, `arc`, `circular` (a major circular arc loop
  around the shape, resized with its radius handle), `orthogonal` (a rectangular
  loop around the shape) or `curved` (a Bézier leaving/entering along the anchor
  normals). The two ends are the real anchors and can be moved independently with
  the endpoint handles. No per-self-loop side/radius/gap configuration is
  required;
- if the router cannot produce a clean path (a link would cross a shape, an
  empty/acute/off-grid route), the diagram is left unchanged, the change is
  reverted and a non-blocking notice reports the problem. Soft quality warnings
  (short segments, clearance from other links) stay in
  `data-sui-chart-route-audit` (`failures`, `hard`) and never rewrite anchors,
  waypoints or routes automatically;
- backward links curve around the shape;
- automatic layout assigns levels from the first node and does not inflate
  levels for cycles or self-loops;
- explicit `x`/`y` state positions override the automatic layout; values from
  `0` to `1` are normalized to the chart area;
- the table is split into **Shapes** and **Links** tabs. Shapes use `Element`,
  `Tag`, `Type` and actions; links use `Element`, `Tag`, `Type`, `From`, `To` and
  actions. Element is first and contains the color picker, Tag is second, and
  Type selects the geometry/line variant. The table tab bar holds a single
  **Add row** button aligned to the right; it opens the Shape/Link choice. The
  table has no settings or undo/redo buttons: imports and exports live in the
  Graph settings panel, and the graph undo/redo buttons cover both diagram and
  table edits.
  Event/guard/action fields remain available in the link properties panel.
  Hovering a table cell shows a tooltip with its full text; the Type combo box is
  excluded. `|` line breaks are rendered on separate lines, as in the diagram.
  The diagram and the tables are two views of the same model: adding, editing,
  moving, resizing, retagging or deleting in either view updates the shared
  element and is reflected immediately in the other view. Deleting a shape also
  removes its connected links.
  Selecting a graph element automatically opens its tab and scrolls the matching
  table row into view. The chart viewport and the table area are separated by a
  10 px gap. The table toolbar and the Shapes/Links tabs stay fixed; only the
  data rows scroll, with the column headers kept sticky. Toolbar, tab and zoom
  buttons share the same 28 px height. The compact **+** button in the chart header opens one
  create menu with the seven shapes and **New link**, so it never covers diagram
  nodes. Its tooltip and `aria-keyshortcuts` document `+`/`N`; `L` creates a link
  directly. Undo/redo are icon buttons in the chart header and table toolbar,
  with `Ctrl+Z` / `Ctrl+U` shortcuts; they track moves, resizes, rotations,
  tag drags, waypoint edits, table edits, additions and deletions. Hidden row metadata keeps the complete element,
  geometry and anchor data in the table without adding visible columns. The `…`
  button opens the element properties panel;
- creating a link is a two-click gesture: the first click selects the source
  contact point, the second selects the target. Hovering an element exposes all
  of its contact points and highlights the nearest candidate; clicking anywhere
  on the element selects that nearest anchor. The two chosen anchors are stored
  as `fromAnchor`/`toAnchor` and constrain the rendered link, including when the
  router has to fall back to axis-aligned ports;
- link endpoint handles are route-aware: while dragging, the nearest contact
  points are evaluated and the handle keeps the nearest one whose route is not
  significantly longer or more complex (length, bends, short segments and acute
  turns are scored). Hold **Alt** to force the exact nearest anchor;
- manual route vertices are draggable on a selected link: each vertex snaps to
  the grid and is preserved while editing (only exact duplicate consecutive
  points are merged). Auto-routed bends can be grabbed too; the first drag turns
  the current bends into manual waypoints. **Add waypoint here** inserts a grid-snapped
  draggable vertex at the clicked point; drag its handle to bend the route, and use
  **Clear manual waypoints** to remove them. Every link subtype
  is editable on the canvas: polylines expose one handle per vertex, a `curved`
  Bézier exposes **two independent control-point handles** (stored as
  `bend1X`/`bend1Y` and `bend2X`/`bend2Y`), an `arc` exposes its single control
  point (`bendX`/`bendY`), a `circular` link exposes its radius (`arcRadius`)
  and its endpoints, and the start/end handles can always be moved to another
  anchor or another shape;
- each Element cell starts with a color swatch/color picker; changing it updates
  the chart fill or edge stroke immediately and is saved with the payload;
- the properties panel is tabbed. Shape properties use **Shape** (border settings
  on the left: style, width, corner radius and, for dashed/dotted borders,
  dash length/gap; subtype and the fill/border/highlight color selectors in the
  right column; a live shape preview at the bottom) and **Position** (rotation,
  X/Y, width and height in two top-aligned columns). The label is edited from the
  table and text styling from Tag settings; the old Text section, State role field
  and double-border option were removed. Link properties are grouped into tabs: **Line** uses two
  columns (both color selectors stacked in the same column; subtype, width slider
  with numeric field, dash type, dash length/gap and a direction button that
  reverses the link when clicked) and **Arrows** (arrow size plus start/end
  marker selectors, each with a graphical example) and **Anchors** (source/target
  anchor names and coordinates, read-only). The link label stays in the table;
  the old description, event, guard, action and tag/note fields were removed;
- link style is inherited from the chart **LINK** configuration; editing a field
  in a single link's properties stores an override for that link only, and
  **Reset style to defaults** clears those overrides;
- when a shape moves/resizes (or `stateGridGap` changes), anchors are recomputed and
  links whose old anchor disappeared are re-anchored automatically before save;
- `stateRouting` selects the default link variant (`curved`, `arc`, `circular`,
  `orthogonal` or `straight`); orthogonal links always use the obstacle-avoiding
  shortest-path router (with automatic outer-lane retries), so there is no
  algorithm choice to make;
- `stateGridGap` is both the grid cell size and the minimum clearance for
  orthogonal state routes, non-attached node borders and parallel line runs
  (default `24`, range `0`-`64`);
- every orthogonal route is audited after generation: segments must be on grid
  lines and keep `stateGridGap` from other shapes and previously routed links.
  The curved/oblique normal stub and its short connector are the only exempt
  points; manual waypoints are authoritative even when the obstacle search has
  no corridor. Consecutive path segments never form an interior angle below
  90 degrees: if a diagonal normal stub would create an acute turn, the router
  falls back to the closest axis-aligned port; a manual waypoint that would
  require a U-turn is reached through a rectangular dogleg. Every horizontal or
  vertical segment is at least `stateGridGap` long; a candidate with shorter
  segments is rebuilt with axis-aligned ports. State diagrams draw no axes,
  ticks or scale labels along the chart border; only element tags are rendered.
  A failed candidate
  is retried with progressively wider outer lanes; the result is exposed as
  `__suiStateRouteAudit` and in `data-sui-chart-route-audit` (`gap`, `checked`,
  `failures`, `minNode`, `minLink`);
- diagram settings are persisted per chart key in localStorage
  (`sui.chart.state.settings.<chart-key>`). **Apply** saves the current chart,
  **Save as default** also stores a global fallback for new state charts, and
  **Reset settings** restores the source `.sui` values. The same defaults can be
  set directly with chart attributes. The panel also edits the **chart title**
  (`data-sui-chart-title`) and the general **shape border color**, **link color**
  and **highlight color**; per-element `borderColor`, `color` and
  `highlightColor` override them, and the highlight also styles the selected
  table row;
- newly inserted shapes use a soft pastel palette. Rotation handles snap to the
  nearest 0/90/180/270° within 6°, which makes it easy to settle tags, text and
  shapes horizontally or vertically;
- **Movable tags** (`stateTagsMovable`, also in the chart menu and Diagram
  settings) makes every rendered tag/note draggable, including link tags. The
  offset is applied to the whole tag group (text, background and leader) and is
  saved as `tagOffset` in the state payload, so it survives reload, undo/redo
  and commits. When disabled, tags stay at their automatic position;
- **Editable routes** (`stateRoutesEditable`, also in the chart menu and Diagram
  settings) shows the editing handles on each link. Dragging edits the route
  according to the link subtype: `orthogonal` creates or moves grid-snapped
  waypoints and rebuilds the route, `straight` creates a straight polyline
  through the waypoints, a `curved` Bézier exposes **two independent control
  handles** (`bend1X`/`bend1Y`, `bend2X`/`bend2Y`), `arc` exposes its control
  point (`bendX`/`bendY`), `circular` exposes its radius (`arcRadius`) and its
  endpoints; dragging the radius handle across the chord flips the arc to the
  other side (`arcSweep`) and beyond the semicircle it becomes a major arc
  (`arcLarge`) so the loop keeps growing instead of clamping. Self-links use a
  major arc by default; the radius handle follows the pointer and sits above the
  shapes so it stays reachable even on circular nodes.
  The selected link highlights the whole stroke, including its arrowheads. While
  a shape or a handle is dragged the board stays fixed (no coordinates jump) and
  the moved element is scrolled back into view on release. When disabled, routes
  stay automatic;
- `stateShapeTagStyle`/`stateLinkTagStyle` select the tag box independently:
  `plain` (no border), `rectangle`, `rounded`, `square`, `circle`, `ellipse`,
  `line` (text above a centered line), `line-start` or `line-end` (the leader
  attaches to the left or right end of the line under the text).
  `stateShapeTagLink`/`stateLinkTagLink` select `none`, `line`, `bezier`, `arrow`
  or `bezier-arrow`. `stateShapeTagKind`/`stateLinkTagKind` set the family default
  (`external`/`internal`) and `stateShapeTagAlign`/`stateLinkTagAlign` select
  `line` or `horizontal`. `stateShapeTagArrowSize`/`stateLinkTagArrowSize` set the
  family arrow size (0 or unset = `stateGridGap / 4`), as does `arrowSize` for
  diagram links; `dashLength`/`dashGap` set a link's dash pattern, while the
  chart defaults are `data-sui-chart-dash-length`/`dash-gap` and
  `data-sui-chart-state-start-marker`. `stateShapeTagArrowShape`/`stateLinkTagArrowShape` select
  `triangle`, `open`, `circle` or `square`, and
  `stateShapeTagTextColor`/`stateLinkTagTextColor`,
  `stateShapeTagBoxColor`/`stateLinkTagBoxColor` and
  `stateShapeTagLeaderColor`/`stateLinkTagLeaderColor` color the text, box and
  leader independently (the general `stateTagTextColor`, `stateTagBoxColor`,
  `stateTagLeaderColor` and `stateTagArrowShape` are fallbacks).
  `stateShapeTagBorderColor`/`stateLinkTagBorderColor`,
  `stateShapeTagBorderWidth`/`stateLinkTagBorderWidth`,
  `stateShapeTagBorderStyle`/`stateLinkTagBorderStyle` and
  `stateShapeTagPadding`/`stateLinkTagPadding` style the tag box border and the
  padding between text and border (`solid`, `dashed`, `dotted`). The tag text
  typography is configurable per family with
  `stateShapeTagFont`/`stateLinkTagFont` plus `…FontSize`,
  `…FontWeight`, `…FontStyle`, `…TextDecoration`, `…TextAlign`,
  `…LetterSpacing`, `…LineSpacing`, `…TextTransform`, `…TextShadow` and
  `…TextShadowColor` (general `stateTagFont*` variants are fallbacks). The link
  itself
  uses `linkStyle` (`auto`/`solid`/`dashed`/`dotted`) and `linkWidth` when set
  from source attributes or link properties. Per-element overrides in the payload
  (`tagKind`, `tagBox`, `tagAlign`, `tagLeader`, `tagArrow`, `tagArrowShape`,
  `tagArrowSize`, `tagFont`, `tagFontSize`, `tagFontWeight`, `tagFontStyle`,
  `tagTextDecoration`, `tagTextAlign`, `tagLetterSpacing`, `tagLineSpacing`,
  `tagTextTransform`, `tagTextShadow`, `tagTextShadowColor`, `tagTextColor`, `tagBoxColor`,
  `tagBorderColor`, `tagBorderWidth`, `tagBorderStyle`, `tagPadding`,
  `tagLeaderColor`, `tagAnchor`, `tagAnchorT`) take precedence;
  `stateShapeTagFont`/`stateLinkTagFont` and their
  `…FontSize` variants customize each family, with
  `stateTagFont`/`stateTagFontSize` as fallbacks;
- at runtime, `chart.setTag(chart, id, text, color, textColor)` updates one tag,
  `chart.update(chart, id, patch)` applies a field patch and
  `chart.updateMany(chart, updates)` updates several elements with one redraw.
  Element ids are the explicit shape/link `id` only; prefixes, labels and
  `from->to` are rejected. Typed helpers (`chart.setText`, `chart.setTagColor`,
  `chart.setHint`, `chart.setColor`, ...) cover single attributes; patch aliases
  include `text`, `textColor`, `tagColor`/`backgroundColor`. See
  [`STDLIB.md`](STDLIB.md#chart-module-chart-core) for the full field list;
- shapes are measured from their text and cannot be smaller than the label,
  including after editing or resizing;
- nodes can be moved and resized; both operations snap continuously to the grid,
  and the snapped X/Y/width/height are persisted. Resizing keeps the top-left
  corner fixed; circle and square keep their aspect ratio, and the resize handle
  keeps a constant screen size across zoom levels. Shapes can be reduced to one
  grid cell; a circle's diameter can go down to one `stateGridGap` (the label
  text no longer forces a larger minimum, because names are not drawn on the
  shape). Resizing a rotated shape keeps its handle reachable: the canvas
  auto-scrolls so the handle stays inside the viewport. Selecting a link shows
  source/target endpoint handles. Dragging them snaps to the nearest grid anchor on the corresponding
  shape and immediately recomputes the path with the obstacle-avoiding router
  and the link subtype. Anchor names remain read-only in link properties;
- state diagrams use a virtual canvas: initially it matches the visible
  viewport and the scrollbars stay hidden. Moving or inserting a shape beyond
  the visible border extends the canvas and reveals scrollbars sized to the new
  content; dragging near an edge also auto-scrolls. The canvas keeps a reserve
  for moving shapes left/up, so the border is not hit immediately;
- state diagrams zoom with **Ctrl/Cmd + mouse wheel** or a two-finger pinch;
  zoom ranges from 20% to 400% and keeps the pointer/pinch center stable. Zoom
  scales the whole drawing, including grid spacing, shape/link geometry, text
  and tags. The zoom controls show the current percentage; **Fit** fits the
  diagram width and **All** fits every graph component (shapes, links and tags)
  into the visible viewport. Dragging empty canvas with the primary button pans
  the whole diagram together (nodes, links and tags); starting the drag on a
  shape, link, tag or handle keeps the normal editing behavior;
- each link has its own `routing` override and optional `waypoints`;
- the orthogonal router uses a grid/A* search around every shape, the chart
  border and previously routed lines, so crossings are not accepted as a normal
  solution;
- tags are placed in free space away from shapes, other tags and routed lines;
  box/pill tags paint an opaque background. With movable tags enabled, dragging
  moves the tag text/background while its leader line and arrow stay attached to
  the shape or link anchor;
- selection stays synchronized between chart objects, shape/link tag labels and
  table rows. Clicking either the element or its tag selects the same element,
  switches to the matching tab and scrolls the row into the visible table area.
  **Ctrl/Cmd/Shift+click** adds or removes elements from the selection;
  **left-drag** on empty canvas draws a selection rectangle over the shapes,
  while **right-drag** pans the whole graph. **Alt+drag** and middle-drag also
  pan. Left-clicking empty canvas clears the whole selection. A selected group can be dragged as one unit and deleted together;
  its internal spacing never changes and the final translation is rounded to a
  whole `stateGridGap` step on both axes. The board grows when shapes are
  dragged beyond the viewport instead of clamping them; the selected elements
  use the `--sui-chart-state-selection` color (amber by default) in both the
  graph and the table row;
- **Ctrl+C** copies the selection, **Ctrl+V** pastes it and **Ctrl+D** or the
  **Duplicate** context-menu command duplicates it. Copying/duplicating shapes
  does not copy their links unless the links are selected too. With two or more
  shapes selected, the alignment toolbar appears for left/center/right,
  top/middle/bottom and horizontal/vertical distribution; blue guide lines
  appear while dragging near another shape's edge or center;
- with a shape selected, **Arrow** keys move the selected shapes by one grid
  cell and **Shift+Arrow** resizes a single selected shape (Left/Right width,
  Up/Down height) while keeping the top-left corner anchored.
  **Delete**/**Backspace** deletes the selected shapes and links; deleting a
  shape also deletes its connected links;
- `stateMode: "editing"` (default) keeps the editor chrome. `stateMode: "static"`
  turns the graph into a presentation surface: selection, drag/resize, route/tag
  editing and creation tools are disabled, while pan/zoom stay available. The
  mode can be changed from source/reactive state, from the designer chart
  settings, or from the runtime Settings **General** tab;
- `tableEdit: true` enables editing in the Shapes/Links table even when
  `stateMode: "static"`: color, text, shape type, element settings and row
  deletion. Without it, in static mode, the table is read-only: cells show plain
  text, the actions column and the Add row button are hidden, and the cell
  tooltip is suppressed while typing;
- static-mode interactions are queued by the chart and consumed with
  `chart.pollEvents` / `chart.nextEvent`. Each event carries `event`, `kind`,
  `part`, `id`, `label`, `tag`, `from`, `to`, pointer coordinates and
  modifiers; SUI code decides which external function to call. Use
  `chart.clearEvents` and `chart.eventCount` to manage the queue.
  `chart.focusRow` highlights, scrolls to and focuses a table row (also in
  static mode). Legacy direct callbacks are still available:
  `chart.on(chart, element, event, handler)`, per-element data fields
  (`onClick`, `onDoubleClick`, `onMouseOver`, `onMouseOut`, `onContextMenu`, the
  `onShape*`/`onLink*` aliases and `onTag*`) and chart-level attributes
  (`onShapeClick`, ...). These call SUI functions from inside the chart runtime
  and are resolved only against `window.__SUI_USER_FUNCS`, so imported data
  cannot call arbitrary browser globals.

Right-click a transition for Link properties, Tag settings, waypoint commands,
Duplicate and Delete. Link subtype/routing is selected in the **Line** tab of
Link properties; **Add waypoint here** inserts a draggable vertex at the click
position and **Clear manual waypoints** removes manual vertices.

State nodes, shape tags and transition labels are keyboard-focusable and expose
tooltip text. Enter/Space on a tag selects its element; Shift+F10 or the context
menu key opens the element menu. `tableView: "both"` shows the state table; `"table"` shows the table
without the diagram. State diagrams do not use numeric series, so the designer's
numeric-series compatibility check treats them specially. The gear button in the
chart header opens a tabbed **Settings** panel: **General** holds the chart
title, shape border and highlight colors, dark/light theme and the **Show anchors
and handles** switch, plus one row of **GRID**, **SHAPES**, **LINKS** and **TAGS**
buttons, save-as-default and reset; **Versions** holds pending changes, commit and version
history; **Data** holds state-data import and JSON/CSV/SVG/PNG export. The graph
itself keeps only the diagram, the shapes/links table and the add controls.
Closing a GRID/SHAPES/LINKS/TAGS sub-panel restores the General settings panel.
SHAPES sets the defaults for newly
inserted shapes (border style/width, dash length/gap, corner radius, border and
highlight colors, plus a **Random fill color** switch or a fixed fill color) with a
live preview; its controls sit in two aligned columns with matching row heights.
Per-element properties override the defaults. LINK
collects appearance, routing, editable routes and arrows, GRID collects grid
visibility, size and color. **Export JSON** embeds the complete shapes, links and
settings, so re-importing that file rebuilds the diagram with the same geometry,
tags, routes, styles and chart settings. The table toolbar
exposes add/delete, undo/redo, import and JSON/CSV export.

State edits are auto-saved locally under `sui.chart.state.<chart-key>` and use
the same 50-step undo/redo model as Gantt edits: task/state moves, resizes,
field edits, additions and deletions are all reversible. The state table also
has Undo/Redo controls. **Commit changes…** in the settings **Versions** tab groups
pending changes into one immutable version and audit entry, stored under
`sui.chart.state.versions.<chart-key>`; **Version history…** lives in the same
tab. Restoring a version loads its state snapshot without creating another version.

The `onCommit` callback and the `sui:state-commit` DOM event receive:

```json
{
  "id": "sv...",
  "revision": 2,
  "baseRevision": 1,
  "at": 0,
  "author": "...",
  "message": "...",
  "chart": "...",
  "kind": "state",
  "payload": { "initial": "...", "states": [], "transitions": [] },
  "changes": []
}
```

As with Gantt, the component performs no network I/O; the host may POST the
version with `If-Match`/`baseRevision` and handle `409` conflicts.

### Data/type compatibility

The chart validates the parsed data against `chartType` at runtime, and the data
overlay shows the same check before Apply:

- single-series types accept exactly one numeric series;
- `barstacked`, `bargrouped` and `combo` require at least two numeric series;
- `radar` accepts one or more series and needs at least three categories;
- `heatmap` needs exactly two numeric series: y row and cell value;
- `candlestick` needs exactly four numeric series in order: open, high, low,
  close;
- `treemap` needs one positive numeric series;
- `boxplot` needs five series: min, q1, median, q3, max;
- `bullet` needs two or three series: value, target, optional max;
- `dumbbell` and `gantt` need exactly two numeric series;
- `state` needs at least one state or transition and uses its own state/transition schema;
- `gauge` uses the first value and optional `min`/`max` (default 0-100);
- `pie`/`donut`/`funnel`/`treemap` require at least one positive value;
- `histogram` requires numeric values and accepts `bins` (1-100);
- `waterfall` accepts positive and negative deltas; all-positive data is allowed
  but only shows increases.

If a source has several numeric columns, choose a multi-series type or set `y`
to one field. The error message names the chosen chart type and the mismatch.
The data overlay disables Apply until the data and type are compatible; the
runtime refuses to draw an incompatible combination.

Literal JSON in SUI source must escape its braces as `\{` and `\}`; the data
overlay writes those escapes automatically. Prefer a `state` string when the
payload is longer than a few rows.

### Size, scrolling and zoom

The chart has no separate width/height parameters: use the normal component
size (`width` and `height` in the inspector **Size** section, or CSS in
`style expr`). The generated chart fills that frame. The default size is
640 x 320 px.

- `overflow: "scroll"` (default) grows the drawing to fit the points and keeps
  the frame fixed. Scrollbars appear inside the chart viewport when the drawing
  is larger.
- `overflow: "zoom"` keeps the drawing at frame size initially and shows
  `-`, `+`, `Reset` controls plus the current percentage. The range is 50%-400%
  in 25% steps; zooming in grows the drawing and exposes scrollbars.
- `overflow: "fit"` squeezes the drawing into the frame without controls. Use it
  for small dashboards and pie/donut cards; axis labels may become dense.
- `pointWidth` reserves horizontal space per point in scroll mode (default 48).
- `pointSize` is the radius of data points in pixels (default 3, range 1-24).
  It applies to line, step, smooth, area, scatter, combo, radar and dumbbell;
  other chart types ignore it.
- `zoom` is the initial zoom for zoom mode (default 1).

`showGrid`, `showLegend` and `showValues` are booleans. `pointWidth`,
`pointSize` and `zoom` accept reactive expressions as well as literals.

`tableView` is available for every chart: `"chart"` (default), `"table"` or
`"both"`. `"gantt"` remains an alias for `"chart"` on Gantt charts. The table is
a native accessible table. Gantt uses Task, WBS parent, Start, End, Duration,
Progress, Milestone, Status, Group, Resource, Depends on, Link, Lag, Constraint,
Constraint date, Slack and Cost; other types expose Label
plus their series, or type-specific columns
such as Open/High/Low/Close, Min/Q1/Median/Q3/Max and heatmap X/Y/Value.
`"table"` hides the chart; `"both"` shows chart and table together. In `"both"`
mode the table is docked to the bottom edge with a fixed height shared by every
chart type: the chart shrinks, a draggable separator above the table sets its
height (stored per chart), only the table body scrolls while the header and the
table tools stay fixed, and resizing the container keeps the table height while
the chart takes the extra space. The separator is keyboard focusable: Up/Down
resize by 8 px (Shift+Up/Down by 32 px), PageUp/PageDown by 10% of the chart
height, Home resets to 80 px and End uses the maximum height. Each header
cell has a filter button that opens an Excel-style popup filter, so columns are filtered independently. The popup
combines a contains search, a value checklist (up to 40 distinct values) and,
for numeric or ISO-date columns, minimum/maximum range fields. Click a header
(not its filter) to sort ascending/descending; Shift+click adds a multi-column
sort. Drag headers to reorder columns; drag the right edge of a header to resize
that column (48-900 px), or focus the resize handle and use Left/Right.
Double-click the handle to auto-fit the column to its widest cell content.
Widths persist with the table state and **Reset column widths** restores
automatic sizing. Rows and headers are keyboard focusable and the identifying first column
is sticky while the table scrolls horizontally.
The table **Settings** menu has Columns, Filters and Export submenus; Export
writes visible rows as CSV, TSV, JSON or JSONL, copies the visible rows as TSV,
or opens a print view. `tableFilters: false` removes column filter buttons and
the Filters submenu. Filter, sort, column order, visibility and widths persist in
`localStorage` for that chart. The caption is shown in table-only mode and
hidden when chart and table are both shown.

`tableAddRow: true` (default `false`) shows the **+ Row** button for every chart
table. On a Gantt it appends a task with a unique id, a one-day range after the
last task, the last group and no dependencies. On other chart types it appends a
zero-valued row to the local chart data, clears active filters, redraws the
chart/table and calls `onDataChange` when configured. Derived tables such as
histograms recompute their bins from the appended value. Without
`onDataChange`, the inserted non-Gantt row remains local until the chart data is
re-rendered from its source.

`tableSelection` chooses the body selection unit: `"row"` (default), `"column"`,
`"cell"` or `"none"`. A narrow row-number gutter selects and drags rows; a thin
strip above each header selects and drags columns; dragging in the body draws a
rectangular cell range. Ctrl/Cmd+click toggles and Shift+click extends from the
anchor. Alt+click on a header sorts instead of selecting.
`tableDeleteRow: true` adds a visible delete button at the end of each row.
`tableSteppers: true` hides the native number spinners and places two adjacent
minus/plus buttons at the right edge of numeric cells; they become visible only
after a click focuses the cell (or via keyboard focus), not on hover alone, and
are hidden while a row, column or cell selection drag is active. Without it numeric cells
have no spinner controls. Row, column and cell selections all use the same
`--sui-chart-table-selection-bg` color, with a 2px
`--sui-chart-table-selection-border` perimeter around the selected block. In
cell mode only the selected cell is filled; the row remains linked to the chart
selection without a full-row tint.
Hover focus uses subtle neutral row/cell tones via
`--sui-chart-table-focus-row-bg` and `--sui-chart-table-focus-cell-bg`;
clicking or dragging removes the hover focus and leaves only selection. Keyboard
focus is a 1px `--sui-chart-table-focus-color` outline (default `#64748b`,
with a lighter-tint variant for the light theme). Selected rows/cells keep
their selection color while hovered. Starting a selection or clicking another
cell cancels the pending edit and removes its focus/highlight.
`tableContextMenu: true` enables a right-click menu on the table with copy,
paste, selected-row deletion and CSV/TSV/JSON/JSONL export of the current
selection. Copy writes TSV to the clipboard; Paste requests clipboard-read
permission and fills the selected cells, rows or columns. Editable Gantt and
chart tables validate pasted values and notify `onTaskChange` / `onDataChange`.

The application requests clipboard write/read permission from the browser when
copying or pasting and shows the result in an `aria-live` table status line; if
the API is unavailable it falls back to a hidden textarea for copy.

Rows, columns and cells support multi-selection. The table **Rows** menu can
select all visible cells in the current mode, clear the selection, copy or paste
the selection as TSV for Excel/Word, and delete selected rows. **Export selection**
writes CSV, TSV, JSON, JSONL, XLSX or HTML for rows, columns or cells.

The table settings **Export options** control header rows, hidden columns and
original-versus-current row order. **Export visible rows** applies the current
table order and filtering; **Export selection** restricts rows and columns to
the active selection. XLSX and HTML preserve header styling and numeric
alignment; formula-like values stay inert text in CSV/TSV/XLSX, so they cannot
execute as spreadsheet formulas. **Print table...** prints the filtered table;
**Print / Save PDF (chart + table)...** opens the browser print dialog with the
current chart SVG and the filtered table, from which the browser can save a PDF.

All editable chart tables and Gantt tables keep a 50-step in-memory undo/redo
history for cell edits, numeric stepper changes, paste, row insertion and row
deletion. Use **Undo**/**Redo** in the chart header, table toolbar, the Gantt
chart menu or the table context menu, or press `Ctrl+Z` and `Ctrl+U` (`Ctrl+Shift+Z` and
`Ctrl+Y` are also accepted). While an input is focused, the browser keeps its
native text undo instead. Cell edits and steppers update the
chart and the edited table in place; when `onDataChange` updates reactive state
and the surrounding view refreshes, the chart/table DOM is preserved so the
table does not flash or rebuild. Histogram bins are derived and
therefore read-only and not individually deletable. Tables above 300 rows use
windowed virtual scrolling while filters and exports still cover all filtered
rows, not only the visible window.

With `interactive: true`, Gantt Task, WBS parent, Start, End, Progress, Group and
Depends on cells are editable, and non-Gantt Label/X/series cells are editable in place
(histogram bins remain read-only). Non-Gantt edits update the local chart data
and call `onDataChange`. The **+ Row** table button appends a new task with a unique
id, a default one-day range after the last task, the last group and no
dependencies. The new row is selected and focused for immediate editing; adding
a row clears active table filters, redraws the timeline and calls
`onTaskChange`. A valid edit redraws the timeline and calls `onTaskChange`, while
dragging or resizing the timeline updates the table. Selection is synchronized for every
chart/table pair: selecting a Gantt bar or its left-hand task label selects the
matching row, and selecting a table row — including its row-number gutter —
highlights both the bar and the label.
For other chart types, selecting a mark selects its row and selecting a row
highlights all marks for that label. In multi-series charts a row highlights all
series marks for that label. Gantt edits also briefly flash the affected cells and task.
Selection is stored in `localStorage` and restored after reload; table filters,
sort, column order and visibility use their own persisted table state.

Gantt tables expose Task, WBS parent, Start, End, Deadline, Actual start/end,
Duration, Effective duration, Remaining, Progress, Work, Work done, Milestone,
Status, Group, Resources, Depends on, Link type, Lag, Constraint, Constraint date,
Slack and Cost. The Resources cell shows removable chips and a `×N` people
quantity per assignment. A small **+** button opens a picker with the resource
types from the Resources tab and a quantity field; confirming adds the
assignment. For Gantt charts the chart menu also exposes
**Resource workload...**, a per-resource summary of assigned tasks and effort.
Workload uses each task `work` value when present, otherwise its working-day
duration; a task assigned to several resources counts in full for each one.
The table has two tabs: **Tasks** and **Resources**. The tab bar stays pinned at
the top while the table scrolls. The Resources tab stores a registry with Name,
Role, Max people, Daily cost and Notes. Resource names are the keys used
by task assignments: renaming a resource updates every task, while deleting a
resource that is still assigned is refused. Any resource typed in the Tasks tab
is added to the registry automatically. A Gantt JSON object may also provide an
initial registry:

```json
{
  "tasks": [ { "task": "Design", "start": "2027-03-01", "end": "2027-03-05", "resources": [ { "name": "Marta", "units": 2 } ] } ],
  "resources": [ { "name": "Marta", "role": "Design", "maxPeople": 3, "dailyCost": 450, "notes": "Part time" } ]
}
```

Each resource assignment has a quantity: use `{ "name": "...", "units": N }`
inside `resources`, or a separate `resourceUnits` object such as
`{ "Marta": 2 }`. `units` defaults to `1` and is edited directly in the resource
chip of the Tasks table. `maxPeople` is the maximum headcount available for that
resource: the sum of `units` of all concurrent tasks assigned to it must not
exceed the limit. Interactive date, resource, quantity and max-people edits,
pastes and similar changes that would create a new overload are rejected and
reverted; pre-existing overloads still show warnings. `0` or an empty
`maxPeople` means unlimited. `dailyCost` is the cost of one person for one day.
When every resource assigned to a task has a daily cost, the task **Cost** cell
becomes read-only and is calculated as
`sum(resource dailyCost × units) × working-day duration`; otherwise the manual
Cost value is kept. Changing a resource rate, an assignment quantity or the
working calendar recalculates the affected task costs.

**Deadline** is a soft constraint: the task is still scheduled, but a planned
finish after the deadline is highlighted in the chart and table.
Duration is the calendar difference `end - start`; **Effective duration** is a
read-only count of working days in the same period using `weekend`, `holidays`
and `timezone`. It is therefore useful for effort and cost calculations (for
numeric-position data it shows `-`). **Actual start/end** record execution.
**Remaining** is editable: when empty it is derived from duration and progress in
working days. **Work** and **Work done** support effort-based progress: when both
are present, the Progress column is derived from `workDone / work`; editing
Progress updates Work done, while editing Work done recomputes Progress.
The chart draws actual work as a thin top bar. Milestone, Status, WBS parent, Link
type and Constraint use constrained
controls. The constraint combo and its
column filter list all seven supported types (`ASAP`, `SNET`, `SNLT`, `FNET`,
`FNLT`, `MSO`, `MFO`) with explanatory labels, even when the current data only
uses `ASAP`. **Depends on** lists task
names as removable chips and offers a task-name combo for adding one or more
predecessors; clicking a chip opens a link editor for its type and lag, while
selecting a dependency that would create a cycle is rejected with a table status
message. Each interactive bar also has a small connector handle under its right
edge: drag it onto another bar or task name to create an FS dependency, then use
the chip editor to change type and lag. `parent` creates the WBS outline; summary rows are
indented and their dates, progress and cost are derived from descendants.
Collapsing a summary also hides its descendant rows in the table, so the table
and timeline show the same tasks. Tasks are rendered as contiguous blocks under
their `group` in both the chart and table, even when the underlying task order is
different; a new row inherits the last task's group and is shown inside that
group. Collapsed `group` values behave the same way;
the table shows a group-toggle strip so groups can be expanded or collapsed
without leaving the table. `autoSchedule: true` recalculates starts from each
link's type/lag and constraint/date values. In date mode, dependency shifts and
lags use working days from the active calendar (weekends, holidays and closures);
task durations are interpreted as working days and end dates are recalculated
accordingly. Effective duration shows the working-day equivalent. Cycles are reported instead of drawing an
invalid schedule.

Right-click the chart, or use its **⋮** button, to export SVG, PNG, CSV, TSV,
JSON or JSONL, grouped in the **Export** submenu of the chart menu. Chart data
exports contain all rows; table Settings exports the
currently visible rows and columns.

Hover tooltips, hover highlights and selection are available on every chart.
`criticalPath: true` highlights zero-slack Gantt tasks and their dependency
links. The calculation uses each link's type/lag, task constraints, working-day
durations and the active weekend/holiday calendar. `holidays: "*-11-16, 2026-11-20..2026-11-21"` shades exact, range and
yearly recurring non-working days. `weekend` and `timezone` configure the
calendar used by shading, ticks and table dates.
`interactive: true` additionally enables Gantt task drag/resize, collapsible
groups and editable Gantt table cells. Gantt edits can be undone/redone with
`Ctrl+Z`, `Ctrl+Shift+Z` / `Ctrl+Y`, or from the chart menu. Hover a task bar to reveal
its start/end resize handles; dragging the body moves the task, dragging an edge
changes only that date. The whole task row is also a keyboard target:
`ArrowLeft`/`ArrowRight` moves by one working day, `Shift+Arrow` resizes the
finish, `Alt+Arrow` resizes the start, and `Enter` opens the task menu. The
dependency connector is focusable: `Enter` opens the target picker; a focused
dependency supports `Enter`/`Shift+F10` to edit and `Delete` to remove. Group
header rows show a triangle and task count; click
the header to collapse the group. A collapsed group keeps one summary bar whose
start/end span all children, whose progress is duration-weighted, and whose
color is the group color. Dependencies crossing the group boundary remain
connected to the summary bar; dependencies internal to the group are hidden.
Tooltips are available on
every chart type and data mark: bars, line/area/step/smooth points, pie/donut
slices, radar points, the gauge arc, funnel stages, heatmap cells, candlestick
bodies, treemap cells, box plots, bullet bars, dumbbell points and Gantt tasks.
A Gantt can call a handler when a task changes:

```sui
chart "Plan"
	data: ganttData
	chartType: "gantt"
	interactive: true
	onTaskChange: ganttChanged
```

The handler receives a JSON string with the current tasks (`id`, `task`,
`start`, `end`, `progress`, `milestone`, `group`, `dependsOn`). A handler is
optional: every edit is automatically saved to a local draft under
`sui.chart.gantt.draft.<chart-key>`, including after a reload. The draft is not
a committed version.

The Gantt header has a **Commit** button. It groups the changes made since the
last commit into one immutable local version and an audit entry, rather than
creating history after every small edit. The button is disabled when there is
nothing to commit; the chart menu also exposes **Commit changes...** and
**Version history...**. Versions and their field-level audit details are stored
under `sui.chart.gantt.versions.<chart-key>`; restoring a version loads its task
and resource snapshot without creating another version. Use `onCommit` to send
the committed version to an application or server:

```sui
chart "Plan"
	data: ganttData
	chartType: "gantt"
	interactive: true
	onCommit: ganttCommitted
```

The callback and the `sui:gantt-commit` DOM event receive
`{ id, revision, baseRevision, at, author, message, chart, payload, resources,
changes }`. `baseRevision` supports optimistic multi-user concurrency: the host
can POST the payload to its endpoint with `If-Match`/`baseRevision`, and reload
or reconcile on a `409` conflict. The control performs no network I/O itself;
authentication, authorization and the shared server are supplied by the host.
`onResourceChange` remains available for applications that also mirror the
resource registry into their own model.

The resource registry is persisted in `localStorage` under
`sui.chart.resources.<chart-key>`. Add `onResourceChange` to save it in the host
application as well:

```sui
chart "Plan"
	data: ganttData
	chartType: "gantt"
	interactive: true
	onResourceChange: resourcesChanged
```

The handler receives a JSON array of `{ name, role, maxPeople, dailyCost, notes }`.
The **Reset saved** button in the Resources tab clears the local registry and
rebuilds it from the chart source and current task names.

Non-Gantt row insertion can notify the application with `onDataChange`:

```sui
chart "Revenue"
	data: chartData
	chartType: "bar"
	tableView: "both"
	tableAddRow: true
	onDataChange: chartDataChanged
```

The handler receives a JSON array of objects containing `label`, `x` and one
property per series. Assign it back to the source state to make the new row
survive later re-renders.

### Colors and style presets

The root supports these CSS custom properties through `style expr`:

`--sui-chart-bg`, `--sui-chart-fg`, `--sui-chart-title-color`,
`--sui-chart-muted`, `--sui-chart-grid`, `--sui-chart-axis`,
`--sui-chart-series-1` through `--sui-chart-series-8`, `--sui-chart-border`,
`--sui-chart-radius`, `--sui-chart-padding`, `--sui-chart-title-size`,
`--sui-chart-label-size`, `--sui-chart-value-size`, `--sui-chart-line-width`,
`--sui-chart-bar-opacity`, `--sui-chart-area-opacity`, `--sui-chart-slice-stroke`,
`--sui-chart-scroll-thumb`, `--sui-chart-scroll-track`,
`--sui-chart-control-bg`, `--sui-chart-control-fg`,
`--sui-chart-control-border`, `--sui-chart-control-muted`,
`--sui-chart-control-hover-bg`, `--sui-chart-control-hover-fg`,
`--sui-chart-control-hover-border`, `--sui-chart-control-radius`,
`--sui-chart-legend-color`, `--sui-chart-legend-bg`,
`--sui-chart-legend-size`, `--sui-chart-tooltip-bg`,
`--sui-chart-tooltip-fg`, `--sui-chart-tooltip-border`,
`--sui-chart-tooltip-radius`, `--sui-chart-tooltip-max-width`,
`--sui-chart-tooltip-font-size`, `--sui-chart-tooltip-delay`,
`--sui-chart-table-font-size`, `--sui-chart-table-cell-padding`,
`--sui-chart-table-head-bg`, `--sui-chart-table-stripe-opacity`,
`--sui-chart-table-caption-display` and `--sui-chart-table-header-position`.

The designer exposes 14 presets: **Default**, **Ocean**, **Forest**, **Sunset**,
**Neon**, **Mono**, **Pastel**, **High contrast**, **Amber**, **Candy**,
**Slate**, **Light**, **Outline**, **Minimal** and **Custom**. Selecting a preset rewrites
`style expr` with CSS variables; edit that expression for custom colors. The
zoom buttons, legend, axis labels and scrollbars use `--sui-chart-control-*`,
`--sui-chart-legend-*` and scrollbar variables with fallbacks to the preset
foreground/axis/muted colors, so they stay legible under every preset. Use
8-digit hex colors instead of `rgba(...)` inside `cssSafe` style expressions.

### Accessibility and security

The chart root is a labelled group and the SVG has `role="img"` plus an
`aria-label` from the title. Charts are visual summaries: provide the same
numbers nearby (for example in a table) when the values must be read precisely.

`data` is untrusted text. Parsing has a 5,000-point limit and a text-size limit;
unknown formats and malformed rows produce a visible message instead of partial
markup. SVG nodes are created with `createElementNS` and labels with
`textContent`, so a payload cannot inject HTML or script. XML DTDs and entity
declarations are rejected. See [`SECURITY.md`](../SECURITY.md#3c-chart-data-is-inert).

## Date/Time picker

`datepicker` wraps the browser's native date, time or datetime-local input. It
supports a visible label, state binding, clear action, min/max range and
required/disabled states.

```sui
state when = ""

datepicker "Appointment"
	mode: "datetime"
	value: when
	min: "2026-01-01T08:00"
	max: "2026-12-31T20:00"
	step: 60
	onChange: appointmentChanged
	required: true
	disabled: false
```

- `mode`: `"date"` (default), `"time"` or `"datetime"`. The generated input uses
  `type=date`, `type=time` or `type=datetime-local`.
- `value`: binds a string state. The stored format is ISO-like:
  - date: `YYYY-MM-DD`
  - time: `HH:MM` (or `HH:MM:SS`)
  - datetime: `YYYY-MM-DDTHH:MM` (or with seconds)
- `min` / `max`: optional native range limits in the same format as the mode.
- `step`: optional step in seconds for time/datetime (passed to the native
  input).
- `onChange`: optional zero-argument function called after the state setter.
- `required` and `disabled`: passed to the native input. Clear is disabled when
  the control is disabled.
- The visible clear button sets the value and state to `""` and invokes
  `onChange`.
- An invalid or out-of-range state value displays an inline error with
  `role="alert"` and sets `aria-invalid`; the native picker still provides its
  own validation UI.
- The label is associated with the input and the clear button has an accessible
  name.

The browser owns the calendar/clock popup and keyboard interaction. The
datepicker does not add network, file or command capabilities. Validate
out-of-range or malformed values before assigning remote data to the bound
state.

## Pages and navigation

Each page is a non-invoked root `view`. The generated app shows the first page
initially; the other pages stay in the DOM but are hidden. Use the built-in
`navigate(pageName)` from a handler to show another page:

```sui
fn openSettings() -> null {
	navigate("settings")
}

fn backHome() -> null {
	navigate("main")
}

view main {
	title "Home"
	h1 "Home"
	button "Settings" onClick: openSettings
}

view settings {
	title "Settings"
	h1 "Settings"
	button "Back" onClick: backHome
}
```

`navigate` takes a page-name string, returns `null`, needs no import, and is a
no-op in a single-page app or when the name is unknown. Page state is global, so
values set on one page remain available on another.

In the **designer**, the page chips above the preview are the page switcher:
click a chip to edit that page; **Insert → New page…** adds pages. The
**View** inspector section edits the active page title/name and offers
duplicate/delete/save-as-template actions. Page navigation inside the generated
app is separate: add buttons or menu items whose handlers call `navigate`.

See the [pages example](../examples/pages/README.md).

## Modal dialogs

A `modal` is a hidden dialog with a title, body content and a footer of
`modalbutton` children. A `name` identifies it for the runtime:

```sui
modal "Edit name" name: "edit-name" {
	p "Update the name."
	input "Name" value: name
	modalbutton "Cancel"
	modalbutton "Save" variant: "primary" onClick: saveName
}
```

Open and close it from handlers:

```sui
fn openName() -> null {
	modal.open("edit-name")
}

fn closeName() -> null {
	modal.close("edit-name")
}

button "Edit name" onClick: openName
```

| Element | Text | Attributes | Children |
|---|---|---|---|
| `modal "Title"` | Dialog title/accessibility label | `name`, `class`, `style` | body content + `modalbutton` |
| `modalbutton "Label"` | Button label | `variant`, `close`, `onClick` | no |

`modal.open(name)` and `modal.close(name)` are built-ins, need no import, and
return `null`. Unknown names are ignored. Opening another modal closes the
current one. The dialog closes on backdrop click, `Escape`, or a default modal
button. Focus moves into the dialog; the opener is focused again when it is still
available after a re-render. The runtime compensates for the hidden page
scrollbar and restores focus with `preventScroll`, so opening or closing a dialog
does not shift or scroll the underlying page (including a static chart).

`modalbutton` closes the dialog by default after its `onClick` handler runs. Set
`close: false` when the action must keep the dialog open. `variant` is
`primary`, `danger`, or empty.

In the designer use **Overlay → Modal**. It starts with a title, body and
Cancel/OK buttons. Right-click the `modal` row for **Add modal button**, or a
`modalbutton` row for **Add button below**, **Edit label**, **Duplicate**,
**Move up/down** and **Delete**. The **Modal** inspector section edits the
stable `name`; select a button to edit its variant and close behaviour. See the
[modal example](../examples/modal.sui).

## Tables and untrusted data

A table with `name:` is a runtime target:

```sui
table name: "users" {
  thead { trow { tcell "ID" head: true ... } }
  tbody { trow { tcell "" ... } }   # visual model
}
```

```sui
fn load() -> null {
  n, e = table.display("users", data)     # JSON, CSV/TSV, arrays, objects
  if e != null { out = e }
}

fn save() -> null {
  rows, e = table.data("users")           # visible rows, including edits
  part, e2 = table.select("users", [0, 2])
  body, e3 = table.json("users")
  csv, e4 = table.exportText("users", "csv;formula=escape")
}
```

- `table.display` sniffs JSON, JSONL, CSV/TSV/SSV/PSV, arrays and objects.
- Cells are written with `textContent`: data cannot inject HTML.
- Limits: 100,000 rows, 10,000 columns, 20 MB of text.
- `table.importText`/`table.exportText` support `csv`, `tsv`, `ssv`, `psv`,
  `json`, `jsonl` and options `header`, `sep`, `quote`, `trim`, `formula`.
- Details: [`STDLIB.md`](STDLIB.md#table-module-table-core).

## Security

Interpolated values and attributes are **escaped automatically**;
`href`/`src` are filtered by scheme (`javascript:`/`data:` are blocked) and
static active schemes are rejected at compile time. Dynamic `style:` requires
`cssSafe(...)`; dynamic classes are filtered against the app class allow-list.
Dynamic URLs require `http.getSafe`/`postSafe`/`getStatusSafe` (or
`ws.connectSafe`) with a literal allow-list: no SSRF, redirects or credentials.
Do not concatenate untrusted data into `style` or unvalidated URLs.

Iframe attributes are validated: malformed data is rejected at compile time;
dynamic `src` requires `srcAllow` (a literal origin/path allow-list) and is
validated at runtime. A rejected URL is not loaded and renders
`Cannot accept the data: ...`.

Policies (style/class/URL, table limits, CSV defaults, iframe) can be changed
with `simpleui.security.json` or `--security <file>`; defaults are strict.
Full model, residual risks and safe patterns: [`SECURITY.md`](../SECURITY.md).

### Iframe

```simpleui
iframe src: "https://example.com/embed"
   sandbox: "allow-scripts allow-forms"
   allow: "camera; microphone"
   referrerpolicy: "no-referrer"
   loading: "lazy"
   width: 640
   height: 360

iframe src: videoUrl srcAllow: "https://example.com"
```

- `src`: `http(s)` or relative; a dynamic URL requires `srcAllow` (default).
- `srcAllow`: literal origin/path allow-list, separated by spaces or commas.
  It is not emitted into the HTML.
- `sandbox`: literal tokens; when absent the compiler adds `sandbox=""`.
- `referrerpolicy`: literal value; when absent the compiler adds
  `no-referrer`.
- `allow`, `loading`, `width`, `height`: literals, validated.
- A rejected dynamic URL is not loaded; the page shows
  `Cannot accept the data: ...`.

## CLI

```bash
simpleui version
simpleui help
simpleui check examples/counter.sui --json
simpleui parse examples/counter.sui
simpleui fmt examples/counter.sui --write
simpleui docs --agent
simpleui docs --search table
simpleui build examples/counter.sui --target web --out dist
simpleui build examples/counter.sui --single --out app.html
simpleui build examples/counter.sui --single --minify --out app.html
simpleui build examples/counter.sui --target desktop --out dist \
  --allow-exec cmd --allow-dir "C:/logs"
simpleui serve examples/counter.sui --watch
simpleui gui
simpleui lsp
simpleui mcp
```

## Targets

- **web**: emits `index.html`, `style.css`, `app.js`; `--single` emits one HTML
  with CSS/JS/runtime inline. `--out app.html` names the file; `--out dist`
  produces `dist/app.html`.
- **web `--minify`**: requires `--single`; removes comments and indentation
  conservatively.
- **single HTML and CSP**: CSS/JS are inline; a strict Content-Security-Policy
  must allow `'unsafe-inline'` (or use hashes/nonces). Multi-file builds do not
  have this constraint.
- **desktop**: emits one self-contained executable with a native bridge
  (`exec`/`fs`), a localhost micro-server and browser launch. Supports
  `--no-open`, `--port N`. Not combinable with `--single`.
