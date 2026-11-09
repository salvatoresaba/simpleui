# State chart runtime guide

This guide covers the runtime side of the `state` chart: updating tags and
colors from SUI code, switching between editing and static mode, and handling
shape/link/tag events in static mode. The complete worked example is
[`examples/process_monitor`](../examples/process_monitor/README.md).

## Overview

A state chart is identified by its `name:` attribute. Runtime code can use the
core `chart` module (no import required) to update the rendered graph, or switch
the chart to static mode and receive events when the user interacts with shapes
and tags.

The chart module is synchronous and returns `(value, error)` like the rest of
the standard library.

```sui
chart "Process monitor" name: "monitor" data: data chartType: "state"
```

## Chart identity

Runtime calls target the chart whose `name:` (or title) matches the first
argument:

```sui
ok, err = chart.setTag("monitor", "process", "RUNNING", "#f59e0b", "#111827")
```

## Element identity

Every shape must declare an explicit `id`, and every id must be unique inside
the chart. Parsing stops with an error when a shape id is missing or duplicated.
Links use the explicit `id` when present; otherwise the runtime assigns
`edge_1`, `edge_2`, ... and verifies that the result does not collide with any
shape id.

The id is the only public handle to an element:

```json
{
  "states": [
    { "id": "process", "label": "Process", "hint": "Processing the order" }
  ],
  "transitions": [
    { "id": "edge_1", "from": "process", "to": "output" }
  ]
}
```

- source data: `"id": "process"`;
- runtime API: `chart.get("monitor", "process")`;
- event payload: `event["id"]`.

Use stable ids: renaming an id breaks every runtime call. The API accepts only
the explicit id; prefixed keys, labels and endpoint pairs are not valid inputs.

## Hover hints and table view

Every shape and link accepts a `hint` field:

```json
{ "id": "process", "label": "Process", "hint": "Processing the order" }
```

- when `hint` is a non-empty string, it replaces the default tooltip;
- when it is empty or missing, the runtime shows the default tooltip;
- `hint: "none"`, `"off"` or `"-"` hides the tooltip for that element only;
- set `stateHints: false` on the chart to hide all hover hints.

The `hint` field is imported from every supported format:

- JSON: `hint` property;
- CSV: `hint` (or `tooltip`/`tip`) column;
- XML: `hint`/`tooltip` attribute or `<hint>` child on a state, and `hint`/`tooltip`
  attribute on a transition.

The hint can be edited:

- from the **Hint** column of the state table (shape and link tabs);
- from the graph context menu → **Hint settings...** (the properties dialog
  opens on the **Hint** tab).

### Multi-line tag and hint text

Tag and hint inputs are single-line. Write `|` where the text must break:

```json
{ "id": "process", "tag": "STEP 1|STEP 2", "hint": "First line|Second line" }
```

- `|` starts a new line;
- `\|` is a literal pipe and `\\` a literal backslash;
- real newlines in existing data are still accepted and normalised to `|`;
- the table cells, the element **Tooltip hint** field and **Tag text** are all
  single-line: type `|` to break a line;
- tags render at most 100 lines; the rest is merged into the last line with an
  ellipsis to keep the DOM bounded;
- older JSON exports (version 1) are migrated on import: existing pipes stay
  literal and real newlines become line breaks. This version exports `version: 2`
  and uses `|` as the break marker;
- unversioned JSON and `.sui` source use the new model directly; escape literal
  pipes as `\|` (JSON: `\\|`).

Only `tag` and `hint` interpret `|`; labels, descriptions, ids and colors keep
pipes as ordinary characters.

Escaping depends on the container:

| Container | Line break | Literal pipe |
|---|---|---|
| UI fields | `|` | `\|` |
| JSON file / API object | `|` | `\\|` (JSON escape for `\|`) |
| CSV | `|` | `\|` |
| XML | `|` | `\|` |
| Inline JSON string in `.sui` | `|` | `\\\\|` |

The extra SUI backslashes are needed because the SUI string layer and the JSON
string layer each unescape one level.

### Global hint style

The chart settings panel has a dedicated **HINTS** button (next to GRID, SHAPES,
LINKS and TAGS) that opens a standalone panel. It styles every tooltip of the
chart; per-element customisation is limited to the `hint` text and its
suppression (`"none"`). The panel lists the attributes as a two-column table and
ends with **Cancel** / **Apply**.

| Control | Effect |
|---|---|
| Show hover hints | enables or disables all tooltips |
| Hint background | tooltip fill (`#rrggbb`, `transparent`) |
| Background opacity | `0`–`1`, multiplies the background alpha |
| Hint text color | tooltip text |
| Hint border color | border color |
| Hint font | font family |
| Font size | text size in pixels |
| Border width | border width in pixels (`0` removes it) |
| Text padding | distance between text and border |
| Corner radius | container rounding |

The panel includes a live **Hint preview** that updates while you drag the sliders,
and the real tooltip uses the same CSS variables
(`--sui-chart-tooltip-bg`, `-fg`, `-border`, `-border-width`, `-padding`,
`-radius`, `-font-family`, `-font-size`, `-shadow`).

- opacity `0` or a `transparent` background removes the fill and the shadow;
- values are stored in `localStorage` per chart and survive a reload;
- **Save settings as default** (General tab) stores them as the fallback used
  by every chart that has no per-chart settings yet;
- unset values fall back to the theme defaults.

The same defaults can be written in the `.sui` source as chart attributes:
`stateHintBg`, `stateHintOpacity`, `stateHintTextColor`,
`stateHintBorderColor`, `stateHintFont`, `stateHintFontSize`,
`stateHintBorderWidth`, `stateHintPadding`, `stateHintRadius`. The designer
exposes them from **Chart settings → Interactions → HINTS...**; runtime edits
stay in `localStorage`, while **Reset settings** returns to the source values.

The General tab is also laid out as a two-column table: labels on the left,
inputs on the right, with every switch on the right. The `stateMode` option is
the **Edit mode** switch (`On` = editing, `Off` = static).

The state table is controlled by `tableView`:

| Value | Surface |
|---|---|
| `chart` | diagram only |
| `table` | table only |
| `both` | diagram plus table |

The designer exposes the same choice as **Table view**. When the table is
visible and an element is selected in the diagram, the matching table row is
highlighted, scrolled into view and focused. Selecting an edge activates the
**Links** tab automatically.

## Editing and static mode

`stateMode` has two values:

| Value | Behavior |
|---|---|
| `editing` | Normal editor: selection, drag/resize, route/tag handles, add/delete tools and editable table |
| `static` | Presentation surface: no selection, no drag/resize, no editing chrome, read-only table; pan/zoom and static events remain |

`editing` is the default. Set the mode from source or from reactive state:

```sui
state staticMode = false
state chartMode = "editing"

fn toggleMode() -> null {
	staticMode = staticMode == true ? false : true
	chartMode = staticMode ? "static" : "editing"
}

switch checked: staticMode onChange: toggleMode

chart "Monitor" name: "monitor" data: data chartType: "state" stateMode: chartMode
```

The mode can also be changed from:

- the runtime **Settings → General → Mode** selector;
- the designer chart settings, field **State mode**;
- `stateMode: "static"` directly in the source.

Static mode keeps pan and zoom. It disables editing tools, handles, selection,
route/tag editing and table editing. The gear button remains available so the
user can return to editing mode.

## Event list

The chart is an **event source**, not a dispatcher. It never needs to call a
function that changes external SUI state. Every static-mode interaction is
appended to a per-chart event queue with the event type and the element that
produced it. SUI code drains the queue and decides which external function to
call.

```sui
fn startEvents() -> null {
	interval(120, pumpChartEvents)
}

async fn pumpChartEvents() -> null {
	events, err = chart.pollEvents("monitor")
	if err != null {
		return
	}
	items, jerr = json.parse(events)
	if jerr != null {
		return
	}
	let i = 0
	while i < len(items) {
		let ev = items[i]
		if str(ev["event"]) == "click" and str(ev["part"]) == "body" {
			if str(ev["kind"]) == "shape" and str(ev["id"]) == "process" {
				processClicked(ev)
			}
		}
		i = i + 1
	}
}

fn processClicked(ev: any) -> null {
	ok, err = chart.update("monitor", "output", "\{\"text\":\"DONE\",\"tagColor\":\"#22c55e\",\"textColor\":\"#ffffff\"\}")
	modal.open("process-details")
}

view main {
	onLoad: startEvents
	chart "Monitor" name: "monitor" data: data chartType: "state"
		stateMode: "static"
}
```

### Queue functions

| Function | Result |
|---|---|
| `chart.pollEvents(name)` | all pending events as JSON and empties the queue |
| `chart.peekEvents(name)` | all pending events as JSON without emptying |
| `chart.nextEvent(name)` | the oldest event as JSON and removes it (`""` when empty) |
| `chart.clearEvents(name)` | empties the queue |
| `chart.eventCount(name)` | number of pending events |

Every event has the same payload used by the legacy callbacks:

```json
{
  "chart": "monitor",
  "kind": "shape",
  "part": "body",
  "target": "shape",
  "event": "click",
  "id": "process",
  "label": "Process",
  "tag": "RUNNING",
  "from": "",
  "to": "",
  "clientX": 420,
  "clientY": 260,
  "button": 0,
  "altKey": false,
  "ctrlKey": false,
  "metaKey": false,
  "shiftKey": false,
  "time": 1730000000000
}
```

- `kind`: `shape` or `link`;
- `part`: `body` or `tag`;
- `event`: `click`, `dblclick`, `mouseover`, `mouseout`, `contextmenu`;
- `id`: the element that produced the event;
- `label`, `tag`, `from`, `to`, pointer coordinates and modifiers.

The queue is capped at the 200 most recent events. Polling is the recommended
model because the chart only produces data; SUI owns all state changes.

### Direct callbacks removed

The legacy direct callback model (`chart.on`, element handler fields and
chart-level `onShape*`/`onLink*`/`onTag*` attributes) has been removed. Charts are
event-source only: use `chart.pollEvents` / `chart.peekEvents` / `chart.nextEvent`
and dispatch the event to your application functions.

## Runtime chart API

The state-chart functions initialize the chart data on demand, so they can be
called from `onLoad` even before the first draw completes. Each function returns
`(value, error)`.

### `chart.setTag`

```sui
ok, err = chart.setTag(name, id, text, tagColor, textColor)
```

Sets the tag text, tag background color and tag text color of one element.
Pass an empty string for a color to leave it unchanged.

```sui
ok, err = chart.setTag("monitor", "process", "RUNNING", "#f59e0b", "#111827")
```

### `chart.update`

```sui
ok, err = chart.update(name, id, patch)
```

Applies a patch to one element. The patch can be an object passed from SUI code
or a JSON object string.

```sui
ok, err = chart.update("monitor", "process", "\{\"text\":\"RUNNING\",\"tagColor\":\"#f59e0b\",\"textColor\":\"#111827\",\"color\":\"#fffbeb\"\}")
```

### `chart.updateMany`

```sui
count, err = chart.updateMany(name, updates)
```

`updates` can be a map `{ id: patch }` or an array of objects with `id` and a
patch. The chart is updated once.

```sui
n, err = chart.updateMany("monitor", "\{\"intake\":\{\"text\":\"READY\"\},\"process\":\{\"text\":\"RUNNING\",\"tagColor\":\"#f59e0b\"\}\}")
```

### `chart.get`

```sui
jsonText, err = chart.get(name, id)
```

Returns the **current** attributes of one shape or link as a JSON object. This
reflects runtime edits, not only the original source data. The result includes
`id`, `label`, `tag`, geometry and colors.

```sui
info, err = chart.get("monitor", "process")
if err == null {
	state, jerr = json.parse(info)
}
```

### `chart.elements`

```sui
jsonText, err = chart.elements(name)
```

Returns a JSON array with one summary per shape/link: `kind`, `id`, `label`,
`tag`, `hint`, `shape` (shapes), `from`/`to`/`routing` (links), `color`,
`tagTextColor`, `tagBoxColor`, and the bound `onClick`/`onTagClick` handler
names. Use it to enumerate the graph and discover the ids to update.

```sui
all, err = chart.elements("monitor")
items, jerr = json.parse(all)
```

Each item can be passed straight back to the runtime API:

```sui
id = str(item["id"])            // "process" or "edge_1"
ok, err = chart.setTag("monitor", id, "RUNNING", "#f59e0b", "#111827")
```

### `chart.find`

```sui
jsonText, err = chart.find(name, query)
```

Searches the graph and returns a JSON array with the same summaries as
`chart.elements`. An empty query returns every element. A `field:value` query
filters one field:

```sui
all, err = chart.find("monitor", "")
running, err2 = chart.find("monitor", "tag:RUNNING")
shapes, err3 = chart.find("monitor", "kind:shape")
process, err4 = chart.find("monitor", "id:process")
```

Supported query fields: `id`, `kind`, `label`, `tag`, `hint`, `shape`,
`from`, `to`, `routing`. Without a colon the query is matched against every field.

### `chart.getAttr`

```sui
value, err = chart.getAttr(name, id, field)
```

Reads one attribute as a string. `kind` is computed; other names are looked up on
the current element. Unknown or empty fields return `("", null)`.

```sui
tag, err = chart.getAttr("monitor", "process", "tag")
kind, err2 = chart.getAttr("monitor", "process", "kind")   // shape
```

### `chart.setAttr`

```sui
ok, err = chart.setAttr(name, id, field, value)
```

Writes one attribute and applies the same validation as `chart.update`.
Unsupported fields return an error.

```sui
ok, err = chart.setAttr("monitor", "process", "text", "RUNNING")
ok2, err2 = chart.setAttr("monitor", "process", "tagColor", "#f59e0b")
```

### `chart.focusRow`

```sui
ok, err = chart.focusRow(name, id)
```

Highlights, scrolls to and focuses the matching table row. It works in static
mode too, unlike graph selection. Pass `""` as `id` to clear the highlight.
Selecting a link activates the **Links** tab.

```sui
ok, err = chart.focusRow("monitor", "output")
ok2, err2 = chart.focusRow("monitor", "edge_1")
```

### `chart.pollEvents`

```sui
jsonText, err = chart.pollEvents(name)
```

Returns all pending events as a JSON array and empties the queue. This is the
recommended way to consume chart events.

### `chart.peekEvents`

```sui
jsonText, err = chart.peekEvents(name)
```

Returns the pending events without emptying the queue.

### `chart.nextEvent`

```sui
jsonText, err = chart.nextEvent(name)
```

Returns the oldest event and removes it. Returns `""` when the queue is empty.

### `chart.clearEvents`

```sui
ok, err = chart.clearEvents(name)
```

Empties the event queue.

### `chart.eventCount`

```sui
count, err = chart.eventCount(name)
```

Returns the number of pending events.

### Element identifiers

All runtime functions accept **only the explicit `id`** of the element.
Prefixes (`shape:`, `state:`, `link:`, `edge:`), labels and `from->to` are not
accepted.

| Target | Accepted id |
|---|---|
| Shape | its `id`, for example `process` |
| Link | its `id`, for example `edge_1` |

Ids are unique across shapes and links, so the id alone identifies the element.

### Typed getters and setters

For single-attribute changes use the typed helpers. Each getter returns
`(string, error)`, each setter returns `(bool, error)`.

| Attribute | Getter | Setter |
|---|---|---|
| label | `chart.getLabel(name, id)` | `chart.setLabel(name, id, value)` |
| tag text | `chart.getText(name, id)` | `chart.setText(name, id, value)` |
| hint | `chart.getHint(name, id)` | `chart.setHint(name, id, value)` |
| description | `chart.getDescription(name, id)` | `chart.setDescription(name, id, value)` |
| shape fill / link line | `chart.getColor(name, id)` | `chart.setColor(name, id, value)` |
| tag background | `chart.getTagColor(name, id)` | `chart.setTagColor(name, id, value)` |
| tag text color | `chart.getTextColor(name, id)` | `chart.setTextColor(name, id, value)` |
| shape kind | `chart.getShape(name, id)` | `chart.setShape(name, id, value)` |
| link routing | `chart.getRouting(name, id)` | `chart.setRouting(name, id, value)` |
| border color | `chart.getBorderColor(name, id)` | `chart.setBorderColor(name, id, value)` |
| border style | `chart.getBorderStyle(name, id)` | `chart.setBorderStyle(name, id, value)` |
| border width | `chart.getBorderWidth(name, id)` | `chart.setBorderWidth(name, id, value)` |
| corner radius | `chart.getCornerRadius(name, id)` | `chart.setCornerRadius(name, id, value)` |

Additional typed accessors:

| Attribute | Getter | Setter |
|---|---|---|
| width (shape `w` / link line) | `chart.getWidth` | `chart.setWidth` |
| height (shape) | `chart.getHeight` | `chart.setHeight` |
| x position | `chart.getX` | `chart.setX` |
| y position | `chart.getY` | `chart.setY` |
| rotation | `chart.getRotation` | `chart.setRotation` |
| link dash style | `chart.getDashStyle` | `chart.setDashStyle` |
| link start/end marker | `chart.getStartMarker` / `chart.getEndMarker` | `chart.setStartMarker` / `chart.setEndMarker` |
| link endpoints | `chart.getFrom` / `chart.getTo` | `chart.setFrom` / `chart.setTo` |
| link waypoints | `chart.getWaypoints` | `chart.setWaypoints` |
| tag kind/box/align/leader | `chart.getTagKind` / `getTagBox` / `getTagAlign` / `getTagLeader` | `chart.setTagKind` / `setTagBox` / `setTagAlign` / `setTagLeader` |
| text placement / tag side | `chart.getTextPlacement` / `chart.getTagSide` | `chart.setTextPlacement` / `chart.setTagSide` |
| border dash length/gap | `chart.getBorderDashLength` / `chart.getBorderDashGap` | `chart.setBorderDashLength` / `chart.setBorderDashGap` |
| snap | `chart.getSnap` | `chart.setSnap` |
| kind | `chart.getKind` | read-only |

`setWidth` is kind-aware: on a shape it writes `w`, on a link it writes the link
line width. `setHeight` is shape-only. `setWaypoints` accepts an array or a JSON
array string.

Example:

```sui
ok1, e1 = chart.setText("monitor", "output", "DONE")
ok2, e2 = chart.setTagColor("monitor", "output", "#22c55e")
ok3, e3 = chart.setTextColor("monitor", "output", "#ffffff")

label, e4 = chart.getLabel("monitor", "process")
hint, e5 = chart.getHint("monitor", "process")
kind, e6 = chart.getKind("monitor", "edge_1")
```

For attributes without a dedicated helper use the generic accessors:

```sui
value, err = chart.getAttr("monitor", "process", "tagKind")
ok, uerr = chart.setAttr("monitor", "process", "tagKind", "external")
```

`chart.update` remains available for multi-attribute or batch changes.

### Patch fields

Tag fields:

- `tag` / `text` / `tagText`
- `tagTextColor` / `textColor`
- `tagBoxColor` / `tagColor` / `backgroundColor` / `boxColor`
- `tagBorderColor`, `tagLeaderColor`
- `tagKind`, `tagAlign`, `tagBox`, `tagLeader`
- `tagArrow`, `tagArrowSize`, `tagArrowShape`
- `tagFont`, `tagFontSize`, `tagFontWeight`, `tagFontStyle`
- `tagTextDecoration`, `tagTextAlign`, `tagLetterSpacing`, `tagLineSpacing`
- `tagTextTransform`, `tagTextShadow`, `tagTextShadowColor`
- `tagPadding`, `tagRotation`, `tagOffset`

Element fields:

- `label`, `description`
- `color` (shape fill or link line)
- `borderColor`, `highlightColor`
- `borderStyle`, `borderWidth`, `cornerRadius`
- `shape` (rectangle, square, circle, stadium, ellipse, rhombus, triangle,
  parallelogram, hexagon, trapezoid, pentagon, note)
- `routing` for links (`auto`, `straight`, `arc`, `bezier`, `curved`,
  `orthogonal`, `circular`)
- `textPlacement`, `tagSide`, `snap`
- `borderDashLength`, `borderDashGap`

Shape geometry:

- `x`, `y` (center coordinates), `w`, `h`, `rotation`
- `width` / `height` are accepted as aliases for `w` / `h`

Link fields:

- `width` (line width), `dashStyle`, `dashLength`, `dashGap`, `arrowSize`
- `startMarker`, `endMarker` (`none`, `triangle`, `open`, `circle`, `square`)
- `from`, `to` (explicit shape ids; validated)
- `fromPort`, `toPort`, `fromAnchor`, `toAnchor`
- `waypoints` (array or JSON array string of `{x, y}`)
- `arcRadius`, `arcSweep`, `arcLarge`
- `bendX`, `bendY`, `bend1X`, `bend1Y`, `bend2X`, `bend2Y`

Colors accept `#rgb`, `#rrggbb`, `rgb()/rgba()`, `transparent` and `none`.
Invalid values return an error instead of being applied.

## Examples

### Buttons update tags and shape color

```sui
state data = "\{\"states\":[\{\"id\":\"intake\",\"label\":\"Intake\",\"shape\":\"stadium\",\"w\":150,\"h\":64,\"tag\":\"WAITING\"\},\{\"id\":\"process\",\"label\":\"Process\",\"shape\":\"rectangle\",\"w\":160,\"h\":76,\"tag\":\"IDLE\"\},\{\"id\":\"output\",\"label\":\"Output\",\"shape\":\"circle\",\"w\":96,\"h\":96,\"tag\":\"OUTPUT\"\}],\"transitions\":[\{\"id\":\"edge_1\",\"from\":\"intake\",\"to\":\"process\",\"tag\":\"start\"\},\{\"id\":\"edge_2\",\"from\":\"process\",\"to\":\"output\",\"tag\":\"done\"\}]\}"

fn setRunning() -> null {
	ok1, e1 = chart.update("monitor", "intake", "\{\"text\":\"READY\",\"tagColor\":\"#dbeafe\",\"textColor\":\"#1e3a8a\"\}")
	ok2, e2 = chart.update("monitor", "process", "\{\"text\":\"RUNNING\",\"tagColor\":\"#f59e0b\",\"textColor\":\"#111827\",\"color\":\"#fffbeb\"\}")
	ok3, e3 = chart.update("monitor", "output", "\{\"text\":\"WAITING\",\"tagColor\":\"#e5e7eb\",\"textColor\":\"#374151\"\}")
}

view main {
	col {
		button "Running" onClick: setRunning
		chart "Process monitor" name: "monitor" data: data chartType: "state"
	}
}
```

These updates are applied to the existing SVG tag, box and shape. The graph is
not rebuilt, so there is no flicker or viewport jump.

### Consume the event list and update another shape

The chart enqueues the click; SUI reads the queue and calls its own function:

```sui
import { json } from "std"

state data = "\{\"states\":[\{\"id\":\"process\",\"label\":\"Process\"\},\{\"id\":\"output\",\"label\":\"Output\"\}],\"transitions\":[\{\"id\":\"edge_1\",\"from\":\"process\",\"to\":\"output\"\}]\}"

fn startEvents() -> null {
	interval(120, pumpChartEvents)
}

async fn pumpChartEvents() -> null {
	events, err = chart.pollEvents("monitor")
	if err != null {
		return
	}
	items, jerr = json.parse(events)
	if jerr != null {
		return
	}
	let i = 0
	while i < len(items) {
		let ev = items[i]
		if str(ev["event"]) == "click" and str(ev["part"]) == "body" and str(ev["id"]) == "process" {
			processClicked(ev)
		}
		i = i + 1
	}
}

fn processClicked(ev: any) -> null {
	ok, uerr = chart.update("monitor", "output",
		"\{\"text\":\"PROCESS CLICKED\",\"tagColor\":\"#22c55e\",\"textColor\":\"#ffffff\"\}")
	if uerr != null {
		status = "Update failed: " + uerr
	}
}

view main {
	onLoad: startEvents
	chart "Monitor" name: "monitor" data: data chartType: "state"
		stateMode: "static"
}
```

Clicking `process` produces a queue item such as:

```json
{ "event": "click", "kind": "shape", "part": "body", "id": "process" }
```

`pumpChartEvents` associates that event with `processClicked`; the chart never
calls SUI code and never touches SUI state. Add more branches to the same loop
for other elements, tags or links.

### Identify elements and attributes

Element identifiers come from the chart data:

```json
{
  "states": [
    { "id": "intake", "label": "Intake", "shape": "stadium", "tag": "WAITING" },
    { "id": "process", "label": "Process", "shape": "rectangle", "tag": "RUNNING" },
    { "id": "output", "label": "Output", "shape": "circle", "tag": "OUTPUT" }
  ],
  "transitions": [
    { "id": "edge_1", "from": "intake", "to": "process", "tag": "start" },
    { "id": "edge_2", "from": "process", "to": "output", "tag": "done" }
  ]
}
```

- shape id: `id` (`intake`, `process`, `output`), mandatory and unique;
- shape label: `label`;
- shape geometry: `shape`, `w`, `h`, `x`, `y`, `rotation`;
- shape tag: `tag`, `tagTextColor`, `tagBoxColor`, `tagBorderColor`;
- shape color: `color`;
- shape events: queued entries read with `chart.pollEvents`;
- link id: `id` (`edge_1`);
- link tag: `tag`, colors, `routing`, `waypoints`.

From an event payload read `event["id"]`, `event["label"]`, `event["tag"]`,
`event["kind"]` (`shape`/`link`) and `event["part"]` (`body`/`tag`). The original
definitions are in the chart data state (for
example `data`). For the **current** runtime values after edits or imports use:

```sui
info, err = chart.get("monitor", "process")        // all attributes as JSON
tag, err2 = chart.getAttr("monitor", "process", "tag")
items, err3 = chart.find("monitor", "kind:shape")
all, err4 = chart.elements("monitor")
```

To inspect an element while developing, log the event payload or the parsed
shape:

```sui
async fn shapeClicked(eventJson: string) -> null {
	event, err = json.parse(eventJson)
	if err == null {
		status = "id=" + str(event["id"]) + " label=" + str(event["label"]) +
			" tag=" + str(event["tag"]) + " kind=" + str(event["kind"]) +
			" part=" + str(event["part"])
	}
}
```

### Switch editing/static

```sui
state staticMode = false
state chartMode = "editing"

fn toggleMode() -> null {
	staticMode = staticMode == true ? false : true
	chartMode = staticMode ? "static" : "editing"
}

view main {
	row {
		switch checked: staticMode onChange: toggleMode
		text "Static mode"
	}
	chart "Monitor" name: "monitor" data: data chartType: "state" stateMode: chartMode
}
```

### Shape click opens a modal

```sui
import { json } from "std"

state data = "\{\"states\":[\{\"id\":\"process\",\"label\":\"Process\"\}],\"transitions\":[]\}"
state selectedLabel = ""
state selectedTag = ""

fn startEvents() -> null {
	interval(120, pumpChartEvents)
}

async fn pumpChartEvents() -> null {
	events, err = chart.pollEvents("monitor")
	if err != null {
		return
	}
	items, jerr = json.parse(events)
	if jerr != null {
		return
	}
	let i = 0
	while i < len(items) {
		let ev = items[i]
		if str(ev["event"]) == "click" and str(ev["part"]) == "body" and str(ev["id"]) == "process" {
			selectedLabel = str(ev["label"])
			selectedTag = str(ev["tag"])
			modal.open("process-details")
		}
		i = i + 1
	}
}

view main {
	onLoad: startEvents
	chart "Monitor" name: "monitor" data: data chartType: "state"
		stateMode: "static"

	modal "Process details" name: "process-details" {
		p "Shape: {selectedLabel}"
		p "Tag: {selectedTag}"
		modalbutton "Close"
	}
}
```

The dialog is opened by `pumpChartEvents` when it finds the `process` click in
the event list. The chart runtime itself never calls `modal.open` or touches SUI
state. The same loop can branch on other ids, kinds or tag events.

### Tag click updates external state

In `pumpChartEvents`, branch on `part == "tag"`:

```sui
if str(ev["event"]) == "click" and str(ev["part"]) == "tag" {
	status = "Tag " + str(ev["tag"]) + " on " + str(ev["id"])
}
```

The queue item contains `kind`, `part`, `id`, `label` and `tag`, so the
same loop handles shape clicks, tag clicks and link clicks.

### Batch update from a feed

```sui
fn applyStatuses(payload: string) -> null {
	n, err = chart.updateMany("monitor", payload)
	if err != null {
		status = "Update error: " + err
	}
}
```

`payload` can be:

```json
{
  "intake": { "text": "READY", "tagColor": "#dbeafe", "textColor": "#1e3a8a" },
  "process": { "text": "RUNNING", "tagColor": "#f59e0b", "textColor": "#111827" },
  "output": { "text": "WAITING", "tagColor": "#e5e7eb", "textColor": "#374151" }
}
```

## Process monitor example

[`examples/process_monitor/app.sui`](../examples/process_monitor/app.sui)
combines all of the above:

- wide viewport (`width: "100%"`, `height: "620px"`);
- mandatory unique shape ids (`intake`, `process`, `output`);
- `pumpChartEvents` polling `chart.pollEvents("monitor")` every 120 ms;
- the **Process** click in the event list calls `processClicked`, which updates
  the **Output** shape and opens the **Process details** modal;
- Waiting / Running / Done buttons using `chart.update`;
- a switch bound to `stateMode`;
- tag click and hover handled from the same event list;
- custom `hint` values on shapes and links shown instead of the default tooltip;
- `tableView: "both"` and `chart.focusRow`, so a click highlights the matching
  table row even in static mode.

## Security and performance notes

- Tag text is inserted as text, never as markup.
- Colors are validated; invalid colors return an error and are not applied.
- The chart only queues plain data; it does not need to call SUI functions. SUI
  code decides which external function receives each queued event.
- Shape ids are validated for presence and uniqueness before the graph is drawn.
- Non-editable mode makes the table read-only and removes editing controls.
- Tag/color updates are applied in place to the existing SVG; structural changes
  (shape, routing, tag kind/box/font, geometry) fall back to a full redraw.
- A short draw-suppression window prevents unrelated SUI re-renders or resize
  observers from rebuilding the chart immediately after an in-place update.
- Static mode disables automatic board refit and ignores small viewport changes
  (for example the scrollbar hidden by an open modal), so shapes do not jump when
  a dialog appears or closes.
- The event queue keeps at most the 200 most recent events; poll it regularly.

## Troubleshooting

- **No event in the list:** the chart must be in `stateMode: "static"`. Editing
  mode is an editor and does not emit static events.
- **Events accumulate:** call `chart.pollEvents` regularly (for example with
  `interval`) or `chart.clearEvents`.
- **Wrong chart:** check the `name:` attribute and the first argument.
- **Element not found:** use the explicit shape/link `id` from the data. Prefixed
  keys, labels and endpoint pairs are not accepted.
- **Missing or duplicate id:** add `"id"` to every shape; ids must be unique
  inside the chart. Link ids may be omitted, but explicit ids are recommended.
- **Event handler not called:** check the branch in `pumpChartEvents`. The queue
  gives you `event`, `kind`, `part` and `id`; match those values.
- **Nothing appears after reload:** `interval` starts in `onLoad`, so the poller
  is recreated on every page load. Clear the old queue with `chart.clearEvents`
  if needed.
- **Text changes but the box looks stale:** the in-place updater rebuilds the
  text box; for complex wrapped text or font changes, use a structural patch
  (`tagFontSize`, `tagBox`, etc.) or reload the chart so a full layout pass runs.
- **Mode changes are ignored:** `stateMode` must be `editing` or `static`.
