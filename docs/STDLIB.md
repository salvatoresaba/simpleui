# SimpleUI Stdlib Documentation

SimpleUI's standard library is importable from `"std"`. Calls compile to
`__SUI_STD.<jsKey>(...)` and are implemented by the runtime. Every
**non-fire-and-forget** function returns **two values**: `value, err`
(`err = "null"` on success). Functions marked *fire-and-forget* return `null`
and do not require a full binding.

Available modules: `http`, `ws`, `fs`, `exec`, `json`, `table`, `tree`, `chart`, `dialog` (desktop).

## Import

```sui
import { http, json } from "std"
import { fs } from "std"
```

## HTTP Module (`http`)

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `http.get(url)` | string | (string, string) | true | HTTP GET; returns the body as string |
| `http.getStatus(url)` | string | (int, string) | true | HTTP GET; returns the status code |
| `http.post(url, body)` | string, string | (string, string) | true | HTTP POST with a string body (JSON/text) |
| `http.postFiles(url)` | string | (string, string) | true | POST the current selected/dropped files as `multipart/form-data` field `files` |
| `http.postFilesSafe(url, allow)` | string, string | (string, string) | true | File upload with SSRF allow-list |
| `http.getSafe(url, allow)` | string, string | (string, string) | true | GET with SSRF allow-list (exact scheme/host, no redirects) |
| `http.getStatusSafe(url, allow)` | string, string | (int, string) | true | Status code with SSRF allow-list |
| `http.postSafe(url, body, allow)` | string, string, string | (string, string) | true | POST with SSRF allow-list |

**SSRF:** `http.get/getStatus/post/postFiles` accept only **literal** URLs. A dynamic URL
must use the `*Safe` variant with a **literal** allow-list (the compiler rejects
a variable). Entries are separated by commas/spaces:

- `https://api.example.com` (all paths) or `https://api.example.com/v1`;
- `/api/` (same-origin path prefix);
- WebSocket: `ws://host:port` / `wss://host:port`.

The guard rejects credentials, non-http(s)/ws(s) schemes, hosts/ports outside
the list and redirects; timeout 30 s, response max 10 MB. The requirement to use
`*Safe` for dynamic URLs is a compile-time policy configurable in
`simpleui.security.json` (default `safe`).

`http.postFiles` sends the **current selection** from `pickFiles:`/`dropFiles:`
as multipart data under the field name `files`. It also sends a parallel
`paths` field with the browser-visible relative paths, so a backend can display
folder/subfolder context. It uploads the actual file bytes; the backend must
enforce size/count limits and validate file names. Use `postFilesSafe` with a
literal allow-list when the destination is dynamic.

## WebSocket Module (`ws`)

Real WebSockets integrated with the runtime. Requires `import { ws } from "std"`.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `ws.connect(url)` | string | (any, string) | true | Open a WebSocket; returns the runtime instance |
| `ws.connectSafe(url, allow)` | string, string | (any, string) | true | WebSocket with a `ws://`/`wss://` allow-list |
| `ws.send(ws, data)` | any, string | (null, string) | true | Send a message (fire-and-forget) |
| `ws.onMessage(ws, fn)` | any, string \| fn | (null, string) | true | Register a message handler (fire-and-forget) |
| `ws.close(ws)` | any | (null, string) | true | Close the connection (fire-and-forget) |

Notes:

- `ws.send`, `ws.onMessage` and `ws.close` are fire-and-forget (`ret0 = "null"`).
- `ws.onMessage` can be state-driven: update a state and trigger a re-render.
- Reconnect/heartbeat are **not** implemented; see [`TODO.md`](../TODO.md).

## File System Module (`fs`)

Filesystem access is available through the **desktop bridge** with capability
security (`--allow-dir`). Some operations may be unavailable in a web context.
All functions are async.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `fs.read(path)` | string | (string, string) | true | Read a file as string |
| `fs.write(path, data)` | string, string | (null, string) | true | Write a file (fire-and-forget) |
| `fs.exists(path)` | string | (bool, string) | true | Whether the path exists |
| `fs.stat(path)` | string | (string, string) | true | File/dir info (descriptive string) |
| `fs.mkdir(path)` | string | (null, string) | true | Create a directory (`mkdir -p` style; fire-and-forget) |
| `fs.remove(path)` | string | (null, string) | true | Remove a file/directory (fire-and-forget) |
| `fs.list(path)` | string | (array[string], string) | true | List directory entries |

`fs.copy`, `fs.move/rename`, `fs.isDir` and `fs.size` are **not exposed** yet.

## Exec Module (`exec`)

Command execution through the **desktop bridge** with an allow-list
(`--allow-exec`).

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `exec.run(cmd, args)` | string, array[string] | (string, string) | true | Run a command. `cmd` is resolved next to the executable too. Returns combined stdout/stderr |

**Example (see `examples/userdata_table/`):**

```sui
import { exec, json } from "std"
async fn load() {
  out, err = exec.run("userdata.exe", [])
  if err != null { ... return }
  data, perr = json.parse(out)
  ...
}
```

## JSON Module (`json`)

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `json.parse(s)` | string | (any, string) | true | Parse JSON |
| `json.stringify(v)` | any | (string, null) | true | Serialize JSON |

## Dialog Module (`dialog`, desktop)

Native OS file and folder selection. Import it with `import { dialog } from "std"`; on web targets the calls return an error. The dialog is user-driven and the paths can be passed to `exec.run`/`fs` (which still enforce their capability allow-lists).

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `dialog.pickFile()` | - | (string, string) | true | Native file dialog; empty string on cancel |
| `dialog.pickFiles()` | - | (array[string], string) | true | Native multi-file dialog |
| `dialog.pickDirectory()` | - | (string, string) | true | Native folder dialog; empty string on cancel |

Desktop implementation:

- Windows: PowerShell + `System.Windows.Forms` dialogs;
- macOS: `osascript` + `choose file` / `choose folder`;
- Linux: `zenity`, falling back to `kdialog`.
## Table Module (`table`, core)

`table` is a **core** module: always available and no import required. It
targets the table with the matching `name:` attribute. Functions are synchronous
(DOM-only) and return `(value, error)`.

| Function | Params | Return | Description |
|---|---|---|---|
| `table.display(name, data)` | string, any | (int, string) | Fill from JSON, CSV/TSV/SSV/PSV, JSONL, arrays or objects |
| `table.data(name)` | string | (array[array[string]], string) | Visible rows, including edits |
| `table.header(name)` | string | (array[string], string) | Header labels |
| `table.select(name, cols)` | string, array[int] | (array[array[string]], string) | Selected columns only |
| `table.slice(name, from, count)` | string, int, int | (array[array[string]], string) | Row block (`count <= 0` = to the end) |
| `table.importText(name, text, spec)` | string, string, string | (int, string) | Explicit parse: `csv`, `tsv`, `ssv`, `psv`, `json`, `jsonl`, `auto` |
| `table.exportText(name, spec)` | string, string | (string, string) | Serialize to CSV/TSV/SSV/PSV/JSON/JSONL |
| `table.json(name)` | string | (string, string) | JSON (objects when a header exists) |
| `table.clear(name)` | string | (int, string) | Clear the body |

`importText`/`exportText` spec: format plus options separated by `;`:

- `header=0|1`
- `sep=comma|semicolon|tab|pipe|<char>`
- `quote=<char>`
- `trim=1`
- `formula=escape` (CSV export: neutralizes `= + - @` with `'`)

```sui
state out = ""

fn load() -> null {
	n, e = table.display("users", "ID,Name\n1,Alice\n2,Bob")
	if e != null { out = e } else { out = str(n) + " rows" }
}

fn saveCSV() -> null {
	csv, e = table.exportText("users", "csv;formula=escape")
	out = csv
}
```

Security:

- cells are written with `textContent`: HTML in data is never interpreted;
- limits: 100,000 rows, 10,000 columns, 20 MB of text per call;
- `formula=escape` prevents formula injection when CSV is opened in
  Excel/LibreOffice (opt-in).

Full details: [`SECURITY.md`](../SECURITY.md).

## Chart Module (`chart`, core)

`chart` is a **core** module: always available and no import required. It
targets the rendered state chart with the matching `name:` attribute (or its
title). Functions are synchronous and return `(value, error)`. They update the
tag text and colors of shapes and links without rebuilding the whole chart, which
makes the chart usable as a process/event monitor. Static interactions are also
available as an event queue (`chart.pollEvents`, `chart.nextEvent`, ...) so the
chart never needs to call SUI state-changing functions directly.

| Function | Params | Return | Description |
|---|---|---|---|
| `chart.setTag(name, id, text, color, textColor)` | string, string, string, string, string | (bool, string) | Set tag text, tag background color and tag text color |
| `chart.update(name, id, patch)` | string, string, any | (bool, string) | Apply a patch object or JSON object to one element |
| `chart.updateMany(name, updates)` | string, any | (int, string) | Apply a map `{ id: patch }` or array of patches |
| `chart.get(name, id)` | string, string | (string, string) | Current attributes of one shape/link as JSON |
| `chart.elements(name)` | string | (string, string) | Current shape/link summaries as a JSON array |
| `chart.find(name, query)` | string, string | (string, string) | Search elements by `field:value` or any field |
| `chart.getAttr(name, id, field)` | string, string, string | (string, string) | Read one attribute as a string |
| `chart.setAttr(name, id, field, value)` | string, string, string, any | (bool, string) | Write one validated attribute |
| `chart.on(name, id, event, handler)` | string, string, string, string | (bool, string) | Bind a SUI function to one element event |
| `chart.off(name, id, event)` | string, string, string | (bool, string) | Remove one or all element bindings |
| `chart.bindings(name)` | string | (string, string) | Active `chart.on` bindings as JSON |
| `chart.focusRow(name, id)` | string, string | (bool, string) | Highlight, scroll to and focus a table row (works in static too) |
| `chart.pollEvents(name)` | string | (string, string) | All pending events as JSON and empties the queue |
| `chart.peekEvents(name)` | string | (string, string) | Pending events as JSON without emptying |
| `chart.nextEvent(name)` | string | (string, string) | Oldest pending event as JSON and removes it |
| `chart.clearEvents(name)` | string | (bool, string) | Empty the event queue |
| `chart.eventCount(name)` | string | (int, string) | Number of pending events |

Element ids:

- shape `id` (`"a"`), mandatory and unique inside the chart;
- link `id` (`"edge_1"`); links may omit it and receive `edge_1`, `edge_2`, ... with a uniqueness check.

All runtime functions accept only the explicit `id`. Prefixed keys (`shape:`,
`state:`, `link:`, `edge:`), labels and `from->to` are rejected.

Events report the explicit `id`. For single-attribute changes use the typed
helpers (`chart.setText`, `chart.setTagColor`, `chart.setHint`,
`chart.setColor`, ...); `chart.getAttr`/`chart.setAttr` cover the remaining fields and `chart.update`
handles multi-attribute patches. Additional typed accessors: `getWidth`/`setWidth`
(shape `w` or link line), `getHeight`/`setHeight` (shape), `getX`/`setX`,
`getY`/`setY`, `getRotation`/`setRotation`, `getDashStyle`/`setDashStyle`,
`getStartMarker`/`setStartMarker`, `getEndMarker`/`setEndMarker`,
`getFrom`/`setFrom`, `getTo`/`setTo`, `getWaypoints`/`setWaypoints`,
`getTagKind`/`setTagKind`, `getTagBox`/`setTagBox`, `getTagAlign`/`setTagAlign`,
`getTagLeader`/`setTagLeader`, `getTextPlacement`/`setTextPlacement`,
`getTagSide`/`setTagSide`, `getBorderDashLength`/`setBorderDashLength`,
`getBorderDashGap`/`setBorderDashGap`, `getSnap`/`setSnap`.

Read the current runtime values (after edits/imports) with:

```sui
shapeJson, err = chart.get("monitor", "process")
allJson, err2 = chart.elements("monitor")
```

`chart.get` returns `id`, `label`, `tag`, `hint`, geometry, colors and handler
fields. `chart.elements` returns a JSON array with `kind`, `id`, `label`, `tag`,
`hint`, geometry/routing, colors and `onClick`/`onTagClick` names.

The chart is an event source. Every static interaction is queued; SUI drains the
queue and calls its own external functions. This is the recommended model:

```sui
events, err = chart.pollEvents("monitor")
```

The queue payload includes `event`, `kind`, `part`, `id`, `label`, `tag`,
`from`, `to`, pointer coordinates and modifiers. Use `chart.pollEvents` (drain),
`chart.peekEvents` (read without draining), `chart.nextEvent` (one event),
`chart.clearEvents` and `chart.eventCount`.

Legacy direct callbacks are still available with `chart.on` (or element data
fields / chart-level attributes) but they call a SUI function from inside the
chart runtime. Prefer the event list for new code:

```sui
ok, err = chart.on("monitor", "process", "click", "processClicked")
```

Event names: `click`, `dblclick`, `mouseover`, `mouseout`, `contextmenu`, and
`tagClick`, `tagDoubleClick`, `tagMouseOver`, `tagMouseOut`, `tagContextMenu`.
Use `chart.off` to remove bindings and `chart.bindings` to inspect them.

Element data can also carry declarative handlers as a fallback:

```json
{ "id": "process", "label": "Process", "onClick": "processClicked" }
```

Supported fields: `onClick`, `onDoubleClick`, `onMouseOver`, `onMouseOut`,
`onContextMenu`, the `onShape*`/`onLink*` aliases, and `onTagClick`,
`onTagDoubleClick`, `onTagMouseOver`, `onTagMouseOut`, `onTagContextMenu`.
Resolution order: `chart.on` binding, element data field, chart-level attribute.
Handlers must be SUI functions.


Patch fields for tags: `tag`/`text`, `tagTextColor`/`textColor`,
`tagBoxColor`/`tagColor`/`backgroundColor`, `tagBorderColor`,
`tagLeaderColor`, `tagKind`, `tagAlign`, `tagBox`, `tagLeader`, `tagArrow`,
`tagArrowSize`, `tagArrowShape`, `tagFont`, `tagFontSize`, `tagFontWeight`,
`tagFontStyle`, `tagTextDecoration`, `tagTextAlign`, `tagLetterSpacing`,
`tagLineSpacing`, `tagTextTransform`, `tagTextShadow`,
`tagTextShadowColor`, `tagPadding`, `tagRotation`, `tagOffset`.
Element fields accepted by `update`: `label`, `description`, `hint`, `color`,
`borderColor`, `highlightColor`, `borderStyle`, `borderWidth`, `cornerRadius`,
`shape`, `textPlacement`, `tagSide`, `snap`, `borderDashLength`,
`borderDashGap`, shape geometry `x`/`y`/`w`/`h`/`rotation`, and for links
`routing`, `width`, `dashStyle`, `dashLength`, `dashGap`, `arrowSize`,
`startMarker`, `endMarker`, `from`, `to`, `fromPort`, `toPort`, `fromAnchor`,
`toAnchor`, `waypoints`, `arcRadius`, `arcSweep`, `arcLarge` and `bend*`
controls. In `setTag`/`tagColor`, `color` means the tag background; in `update`,
`color` means the shape fill or link line color.

Colors accept `#rgb`, `#rrggbb`, `rgb()/rgba()` and `transparent`/`none`.

```sui
fn updateMonitor() -> null {
	ok, e = chart.setTag("monitor", "a", "Running", "#22c55e", "#ffffff")
	ok2, e2 = chart.update("monitor", "edge_1", "\{\"text\":\"go\",\"tagColor\":\"#0ea5e9\",\"textColor\":\"#ffffff\"\}")
	n, e3 = chart.updateMany("monitor", "\{\"b\":\{\"tag\":\"Waiting\"\},\"c\":\{\"tag\":\"Done\"\}\}")
}

interval(1000, updateMonitor)
```

## Tree Module (`tree`, core)

`tree` is a **core** module: always available and no import required. It targets
the tree with the matching `name:` attribute (`name: "files"`). Functions return
`(value, error)`.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `tree.importJson(name, text)` | string, string | (int, string) | no | Replace nodes from JSON; returns node count |
| `tree.importXml(name, text)` | string, string | (int, string) | no | Replace nodes from XML; returns node count |
| `tree.importJsonFile(name, path)` | string, string | (int, string) | yes | Read a selected browser file and import JSON |
| `tree.importXmlFile(name, path)` | string, string | (int, string) | yes | Read a selected browser file and import XML |
| `tree.exportJson(name)` | string | (string, string) | no | Export the current nodes as JSON text |
| `tree.exportXml(name)` | string | (string, string) | no | Export the current nodes as XML text |
| `tree.downloadJson(name, filename)` | string, string | (int, string) | no | Download current nodes as a JSON file (`1` = success) |
| `tree.downloadXml(name, filename)` | string, string | (int, string) | no | Download current nodes as an XML file (`1` = success) |
| `tree.clear(name)` | string | (int, string) | no | Remove all nodes |
| `tree.expandAll(name)` | string | (int, string) | no | Expand every parent; returns the number of parents |
| `tree.collapseAll(name)` | string | (int, string) | no | Collapse every parent |
| `tree.expandTo(name, level)` | string, int | (int, string) | no | Expand parents above `level`; collapse deeper ones |
| `tree.setChildren(name, key, json)` | string, string, string | (int, string) | no | Add/replace children of an item from JSON; returns node count |
| `tree.setChildrenError(name, key, message)` | string, string, string | (int, string) | no | Show a lazy-load error on an item |
| `tree.retryChildren(name, key)` | string, string | (int, string) | no | Retry a lazy load through `onLoad` |
| `tree.lastMove(name)` | string | (string, string) | no | JSON `{key, parent, index}` of the last reorder |
| `tree.undoMove(name)` | string | (int, string) | no | Restore the previous tree from the in-memory undo stack |
| `tree.selectedKeys(name)` | string | (string, string) | no | JSON array of selected keys |
| `tree.checkedKeys(name)` | string | (string, string) | no | JSON array of checked keys |
| `tree.search(name, term)` | string, string | (int, string) | no | Apply or clear the inline filter |
| `tree.setColumns(name, json)` | string, string | (int, string) | no | Create/replace columns from a JSON array |
| `tree.getColumns(name)` | string | (string, string) | no | Current columns as a JSON array |
| `tree.moveColumn(name, from, to)` | string, int, int | (int, string) | no | Reorder a column |
| `tree.setCell(name, key, column, value)` | string, string, string, string | (int, string) | no | Set one cell value |

### JSON format

The root can be an array, or an object with `nodes`, `items`, `tree` or
`children`. Each node supports:

```json
[
  {
    "label": "src",
    "key": "src",
    "open": true,
    "editable": true,
    "renameIcon": "A",
    "showRenameIcon": false,
    "icon": "📁",
    "iconOpen": "📂",
    "badge": "12",
    "badgeType": "info",
    "meta": "2 KB",
    "checked": true,
    "tooltip": "Source directory",
    "lazy": false,
    "size": "2 KB",
    "date": "today",
    "status": "ok",
    "cells": { "owner": "Anna" },
    "children": [
      { "label": "main.go", "key": "main" }
    ]
  }
]
```

- `label` is preferred; `text`, `name` and `title` are accepted aliases.
- `key` is optional; a generated path is used when it is missing. Keys must be
  unique.
- `open`, `editable`, `showRenameIcon` and `checked` accept booleans or
  `true`/`false`/`1`/`0` strings.
- `icon`, `iconOpen`, `badge`, `badgeType`, `meta`, `tooltip`, `size`, `date`
  and `status` are inert text. `badgeType` is `neutral`, `info`, `success`,
  `warning` or `danger`.
- `lazy: true` marks an item whose children are loaded later with
  `tree.setChildren`.
- `cells` is an object of custom column values (`{ "owner": "Anna" }`); the
  SUI attribute also accepts `key=value` text.
- Strings and numbers are accepted as leaf labels.

### XML format

```xml
<?xml version="1.0" encoding="UTF-8"?>
<tree name="files">
  <item label="src" key="src" open="true" icon="📁" iconOpen="📂" badge="12" badgeType="info" meta="2 KB" checked="true">
    <item label="main.go" key="main"/>
    <item label="README.md" key="readme" renameIcon="📝"/>
  </item>
  <item label="generated" key="generated" editable="false"/>
</tree>
```

`<tree>` is the container; any element is accepted as a node, so `<item>`,
`<node>` or application-specific tags work. The element text is used as the
label when no `label`/`text`/`name`/`title` attribute is present. XML attributes
`icon`, `iconOpen`, `badge`, `badgeType`, `meta`, `tooltip`, `lazy`, `size`,
`date`, `status` and `checked` are preserved.

### Lazy loading

An item marked `lazy: true` has no children and shows a disclosure arrow.
Expanding it writes its key to the state bound by `loadKey` and calls `onLoad`.
The handler can append children later:

```sui
state loadKey = ""

fn loaded() -> null {
	// start an async request using loadKey
}

async fn finishLoad() -> null {
	n, e = tree.setChildren("files", loadKey, "[{\"label\":\"child\"}]")
	if e != null {
		tree.setChildrenError("files", loadKey, e)
	}
}
```

Use `tree.selectedKeys`/`tree.checkedKeys` to read selection and checkbox state;
both return JSON arrays.

### File helpers

`importJsonFile`/`importXmlFile` read the file selected by `pickFiles:` or a drop
(`drag.files()`), matching `path` or `name`; pass `""` to use the first selected
file. On desktop, a native `dialog.pickFile()` returns only a path: read it with
`fs.read(path)` and pass the text to `tree.importJson`/`tree.importXml` instead.

```sui
state importPath = ""
state note = ""

async fn importJSON() -> null {
	n, e = tree.importJsonFile("files", importPath)
	if e != null { note = e } else { note = "imported " + str(n) + " nodes" }
}
```

`downloadJson`/`downloadXml` create a browser download; the filename is
sanitized and the extension is added when missing.

### Limits and safety

- limits: 10,000 nodes, nesting depth 64, 2,000 characters per label and
  tooltip, 512 per key, 64 per icon and 128/256 per badge/metadata; input text
  uses the global `maxText` limit (default 20 MB);
- labels, keys, icons, badges and metadata are written with
  `textContent`/attributes, never HTML;
- XML DTD and entity declarations are rejected before parsing;
- imported data replaces only the node structure: the tree title, style,
  `name` and runtime bindings are preserved;
- local open/rename overrides for that tree are cleared on import;
- imported models survive reactive view re-renders;
- export reads the visible labels and open state, so inline renames are
  reflected.

## Error Handling Pattern

Every two-value function follows this pattern:

```sui
async fn fetchData() {
  body, err = http.get("https://api.example.com/data")
  if err != null {
    // handle error
    return
  }
  data, perr = json.parse(body)
  if perr != null {
    // handle parse error
    return
  }
  // use data
}
```

## Security Notes (Desktop)

- **Capability-based**: explicit permissions via CLI (`--allow-exec`,
  `--allow-dir`)
- **Path traversal protection**: normalization, allow-list and symlink
  resolution (a link inside an allowed directory cannot escape)
- **Per-run secret**: bridge calls use a random token compared in constant time
- **Localhost only**: binds `127.0.0.1`, requires a loopback Host header
  (anti-DNS-rebinding), rejects cross-site requests
- **Command resolution**: commands without a path separator are resolved next
  to the executable; no shell, so arguments are not reinterpreted
- **app.js**: `Cache-Control: no-store`, `X-Frame-Options: DENY`

## Not Yet Available

- HTTP custom headers/timeout override and multipart upload;
- WebSocket reconnect/heartbeat;
- advanced `fs` operations, crypto and path utilities;
- source maps.
