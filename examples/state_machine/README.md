# State machine chart

This example demonstrates the legacy `chartType: "state"` geometry: a generic
shapes/links diagram with events, guards, actions, a self-loop and an accessible
synchronized data table. New documents should use the dedicated **Graph** control
(`graph "..." data: ...`) documented in
[`examples/graph`](../graph/README.md); this example intentionally keeps the old
syntax to verify backward compatibility.

## What it demonstrates

- JSON object input with `states` and `transitions`.
- Explicit `shape` geometry for circles, stadiums, rhombuses and rectangles.
- No initial/final/choice role markers: every node is a regular shape.
- Forward, backward and circular self-loop transitions.
- Shape subtypes: rectangle, square, circle, stadium, ellipse and rhombus.
- Visible optional grid with configurable cell size and color.
- Circle centers and all shape vertices snapped to grid vertices.
- Orthogonal horizontal/vertical segments constrained to grid lines.
- Named grid/perimeter anchors, orthogonal links and perpendicular port stubs.
- Synchronized amber selection in graph/table and context-menu deletion; deleting a shape removes connected links.
- Shape border/fill/text/tag styling; link width/dash/markers/tag styling.
- Separate shape-tag and link-tag families with configurable style, leader and font.
- Automatic circular self-loop; A* or outer-lane routing.
- Hidden row metadata for complete reconstruction and read-only anchors in link properties.
- Only explicit shape/link tags are drawn on the diagram; event, guard and action
  stay in the table and inspector.
- A synchronized editable state/transition table with the element name first and
  a color picker for each state/transition.
- Theme-aware SVG nodes, edges and tag labels; transition-tag leader lines/arrows.

## Run it

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
simpleui serve app.sui --watch
```

A ready-to-run desktop build is included in `dist/`.

## Complete source

```sui
state orderState = "\{\"states\":[\{\"id\":\"idle\",\"label\":\"Idle\",\"shape\":\"circle\",\"w\":48,\"h\":48,\"description\":\"Waiting for an order\"\},\{\"id\":\"validating\",\"label\":\"Validating\",\"description\":\"Check stock and payment\",\"shape\":\"stadium\",\"borderStyle\":\"solid\",\"borderWidth\":1.5,\"cornerRadius\":18,\"tag\":\"Stock and payment check\"\},\{\"id\":\"paymentChoice\",\"label\":\"Payment?\",\"shape\":\"rhombus\",\"w\":96,\"h\":96\},\{\"id\":\"paid\",\"label\":\"Paid\",\"description\":\"Payment captured\"\},\{\"id\":\"cancelled\",\"label\":\"Cancelled\",\"shape\":\"circle\",\"w\":48,\"h\":48\},\{\"id\":\"shipped\",\"label\":\"Shipped\",\"shape\":\"circle\",\"w\":48,\"h\":48\}],\"transitions\":[\{\"from\":\"idle\",\"to\":\"validating\",\"event\":\"submit\",\"label\":\"submit\"\},\{\"from\":\"validating\",\"to\":\"paymentChoice\",\"event\":\"valid\",\"label\":\"valid\"\},\{\"from\":\"validating\",\"to\":\"cancelled\",\"event\":\"invalid\",\"label\":\"invalid\"\},\{\"from\":\"paymentChoice\",\"to\":\"paid\",\"event\":\"capture\",\"guard\":\"payment.ok\",\"action\":\"reserveStock()\"\},\{\"from\":\"paymentChoice\",\"to\":\"cancelled\",\"event\":\"decline\"\},\{\"from\":\"paid\",\"to\":\"shipped\",\"event\":\"dispatch\",\"guard\":\"stock.ready\",\"linkType\":\"bezier\",\"width\":2,\"dashStyle\":\"solid\"\},\{\"from\":\"validating\",\"to\":\"validating\",\"event\":\"retry\",\"label\":\"retry\"\}]\}"

view main {
	title "Order state machine"
	col class: "state-page" {
		header class: "state-hero" {
			p "SimpleUI state machine" class: "state-kicker"
			h1 "Order lifecycle" class: "state-title"
			p "States, transitions, events, guards and actions rendered from one JSON state machine. Use the table to inspect the same data." class: "state-subtitle"
		}

		card class: "state-card" {
			chart "Order state machine" name: "order_state" class: "state-chart" data: orderState chartType: "state" stateTagLink: "line" stateShapeTagStyle: "pill" stateShapeTagLink: "arrow" stateLinkTagStyle: "box" stateLinkTagLink: "line" stateTagFontSize: 11 stateGrid: true stateGridGap: 24 stateGridColor: "#cbd5e1" tableView: "both" tableFilters: false showGrid: false showLegend: false showValues: true interactive: true theme: "light" style: cssSafe("--sui-chart-series-1: #0284c7; --sui-chart-series-2: #16a34a; --sui-chart-series-3: #d97706; --sui-chart-series-4: #9333ea; --sui-chart-gantt-dependency: #64748b; --sui-chart-fg: #102a43; --sui-chart-muted: #64748b; --sui-chart-bg: #ffffff;")
		}

		row class: "state-help" {
			p "Shapes are explicit: circle, rectangle, stadium, rhombus and the other palette geometries." class: "state-note"
			p "Self-loops, tags, events, guards and actions are rendered from the same data." class: "state-note"
			p "Keyboard: Tab focuses states and transition labels; the chart root supports arrow navigation through marks." class: "state-note"
		}
	}
}

css "body \{ margin: 0; background: #f3f7fb; \} .state-page \{ max-width: 1180px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 16px; background: #f3f7fb; min-height: 100vh; \} .state-hero \{ padding: 22px 24px; border: 1px solid #cbd9e8; border-radius: 16px; background: linear-gradient(135deg, #ffffff 0%, #eaf4ff 100%); box-shadow: 0 8px 24px rgba(15,54,90,.08); \} .state-kicker \{ margin: 0 0 6px; color: #0369a1; font-size: 12px; font-weight: 750; letter-spacing: 1.4px; text-transform: uppercase; \} .state-title \{ margin: 0; color: #0f2740; font-size: 30px; line-height: 1.1; \} .state-subtitle \{ margin: 8px 0 0; max-width: 760px; color: #526b84; font-size: 13px; line-height: 1.5; \} .state-card \{ border: 1px solid #d5e1ee; border-radius: 14px; background: #ffffff; padding: 12px; box-shadow: 0 3px 12px rgba(15,54,90,.05); \} .state-chart \{ width: 100%; height: 620px; \} .state-help \{ display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px 20px; \} .state-note \{ margin: 0; color: #526b84; font-size: 12px; line-height: 1.55; \} @media (max-width: 760px) \{ .state-help \{ grid-template-columns: 1fr; \} .state-title \{ font-size: 24px; \} .state-chart \{ height: 540px; \} \}"
```

## Data shape

```json
{
  "states": [
    { "id": "idle", "label": "Idle", "shape": "circle", "w": 48, "h": 48, "description": "Waiting" },
    { "id": "choice", "label": "Payment?", "shape": "rhombus", "w": 96, "h": 96 },
    { "id": "done", "label": "Done", "shape": "circle", "w": 48, "h": 48 }
  ],
  "transitions": [
    { "from": "idle", "to": "choice", "event": "submit" },
    { "from": "choice", "to": "done", "event": "capture", "guard": "payment.ok", "action": "reserveStock()" }
  ]
}
```

CSV input accepts `from,to,label,event,guard,action` rows (an edge list). XML
accepts `<state>` elements with optional nested `<transition>` elements, or
sibling `<transition>` elements with `from` and `to` attributes.

## Rendering notes

The layout is automatic and layered from the transition levels, starting with
the first node. Cycles and self-loops do not inflate the layout. Explicit `x` and `y` values on state objects can be
used for manual positions; normalized values between `0` and `1` are scaled to
the chart area.

Every node is rendered from its explicit `shape` field; there are no
initial/final/choice role markers. Transition arrows leave the source boundary
and enter the target boundary. The diagram renders only the explicit shape/link
tags; names, labels, events, guards and actions remain in the table, inspector,
tooltip and accessibility text.

The table is editable when the chart is rendered interactively. It has separate
**Shapes** (`Element`, `Tag`, `Type`, actions) and **Links** (`Element`, `Tag`,
`Type`, `From`, `To`, actions) tabs; Element is first and starts with the color
picker, Tag is second. One **Add row** button opens the Shape/Link choice and Type selects the
variant. Selecting an element in the graph opens its tab and scrolls the matching
row into view. A compact **+** button in the chart header opens one create menu
with the six shapes and **New link**; it cannot cover diagram nodes. The tooltip
documents `+`/`N`, and `L` creates a link directly.
Event/guard/action are edited in link properties.
Undo/redo are icon buttons with hints. X/Y/width/height and anchor data are kept
in hidden row metadata and in the element properties panel. Show the grid from
the settings **GRID** sub-panel to see grid/perimeter anchors; link properties display the current
source/target anchor names and coordinates read-only. Moving or resizing a shape
snaps continuously to the grid, and the snapped X/Y/width/height are persisted.
Hovering a shape or link always shows its name, description and optional tag,
even when no tag is drawn; tag labels use the same tooltip. Description and
Tag / note use multiline textareas and preserve explicit line breaks. Clicking a
shape/link or its tag selects the same item and scrolls its table row into view.
Arrow keys move the selected shape by one grid cell; Shift+Arrow resizes it
(Left/Right width, Up/Down height) with the top-left corner anchored;
Delete/Backspace deletes the selected shape or link. The `…` button opens properties for a shape
or link. Shape properties include subtype, fill, border style/width/color,
single/double border, corner radius, dimensions, tag/note and tag side. Link properties include subtype, color, width, dash, start/end marker,
source/target grip or port, label and tag. Each Element cell starts with a color
picker.

The example sets a stadium shape with an external note tag and a bezier link to
demonstrate categories, subtypes, style and tag families. The chart menu selects
connection style, link/shape tag style and
leader, tag font/size, diagram settings and theme. Diagram settings are saved per
chart key; **Save as default** stores a global fallback and **Reset settings**
restores the `.sui` values. Shape and link tags avoid shapes, other tags and
routed lines. Self-loops are normal links between two distinct attachment points
of the same shape (never the same point): create one in link mode by clicking the
same shape twice, then drag either endpoint handle to change that attachment point
independently. The selected subtype shapes the self-link as well, and Bézier links
expose two control-point handles. Endpoint handles can also be dropped on another
shape to reattach that link end. Link properties include a **Direction** control to
reverse source and target, and the LINK panel can draw the arrow at both ends. When
the router cannot find a clean path the diagram is left unchanged and a notice
explains the issue.
Anchors and
handles are shown only when the element is selected and editing handles are enabled. Selecting a link shows
draggable source/target endpoints; dragging snaps to the nearest anchor and
reroutes with the link subtype. Ctrl/Cmd + mouse wheel and two-finger
pinch zoom the diagram (50%-400%) around the pointer/pinch center. Orthogonal
routes always use the obstacle-avoiding A* router with `stateGridGap` and retry
on wider outer lanes when no corridor fits. Every final route is audited
for grid alignment, clearance from other shapes and clearance from previously
routed links. The summary is available
in `data-sui-chart-route-audit` on the chart. Edits are auto-saved, support
Undo/Redo and can be grouped into versions with the header Commit button;
`onCommit`/`sui:state-commit` expose committed versions for host/server
synchronization.

## Related documentation

- [Chart API](../../docs/API_REFERENCE.md#chart)
- [GUI editor chart control](../../docs/GUI_EDITOR.md#chart-control)
- [Complete Gantt reference](../../docs/GANTT_COMPLETE_REFERENCE.md)
