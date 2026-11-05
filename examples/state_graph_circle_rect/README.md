# Circle, rectangle and triangle diagram

Minimal diagram-editor example with a circle, a rectangle, a triangle and one
orthogonal link, rendered from a single state-machine model.

## What it demonstrates

- **One model, two views.** The diagram and the **Shapes** / **Links** tables
  edit the same entities. Adding, editing, moving, resizing or deleting in one
  view is reflected immediately in the other.
- **Shape palette.** The **+ Add** menu and the table **Type** column offer
  rectangle, square, circle, stadium, ellipse, rhombus, triangle, parallelogram,
  hexagon, trapezoid, pentagon and note.
- **Link creation.** Choose **+ Add â†’ Link**, then click the first element and
  the second. Hovering an element highlights all of its contact points and the
  nearest one becomes the candidate; clicking anywhere on the element selects
  that anchor. The two chosen anchors are stored as `fromAnchor`/`toAnchor` and
  constrain the rendered link. Endpoint handles are route-aware: dragging scores
  nearby anchors by route length, bends, short segments and acute turns and keeps
  the nearest one whose route is not significantly worse; **Alt** forces the
  nearest anchor.
- **Route vertices.** On a selected link, interior vertices are draggable and
  snap to the grid. Vertices are preserved while editing (only exact duplicate
  consecutive points are merged); auto-routed bends can be grabbed and become
  manual waypoints.
  **Add waypoint here** inserts a grid-snapped draggable vertex at the click
  point; drag the handle to bend the route. **Clear manual waypoints** removes
  them.
- **Exact shape contact.** Circle, ellipse and triangle anchors are solved on the
  real boundary, not rounded to the nearest grid point. The link starts exactly
  on the shape. Triangle vertices use the angle bisector as the outgoing
  direction.
- **Routing.** Diagonal normal stubs are used only when they do not create a
  turn below 90 degrees; otherwise the router falls back to axis-aligned port
  directions while keeping the chosen anchor point. Interior horizontal/vertical
  segments are at least `stateGridGap` long and stay on grid lines. When no clean
  path exists the diagram is left untouched, the edit is reverted and a
  non-blocking notice explains the problem; quality warnings are reported in
  `data-sui-chart-route-audit` (`failures`, `hard`) without rewriting the data.
- **Self-loops.** In link mode click the same shape twice to create a link between
  two distinct attachment points of that shape (the points never coincide). The
  chosen subtype shapes it too (`straight`/`arc`/`circular`/`orthogonal`/`curved`)
  and the curve uses the real anchors, so dragging either endpoint handle moves
  just that attachment point. Bézier links expose two control-point handles; the
  other subtypes expose their own vertex/radius handles. Endpoint handles can also
  be dropped on another shape to reattach that link end.
- **Resize.** The resize handle keeps a constant screen size at every zoom level.
  Resizing keeps the top-left corner fixed; circles and squares keep their
  aspect ratio. Shapes can be reduced to one grid cell, including a circle
  diameter of one `stateGridGap`. Resizing a rotated shape auto-scrolls the
  canvas so its handle stays reachable.
- **Tags.** Shape and link tags have independent family defaults and per-element
  overrides. External tags support plain, rectangle, rounded, square, circle,
  ellipse and above-line styles, a straight or Bézier leader, an optional arrow
  and a draggable attachment point. Internal shape tags are centered and do not
  rotate with the shape; internal link tags interrupt the line and can follow the
  line direction or stay horizontal. Leaders follow the outward normal at the
  contact point: they start on the real box border (or on the underline) and end
  on the shape perimeter / link path, with the box oriented so its facing border
  is perpendicular to the leader. Automatic placement uses the midpoint of the
  highest side (topmost point for circles/ellipses; the more horizontal top side
  or top vertex for oblique polygons); links use the highest horizontal segment
  midpoint, or the midpoint with the tag to the side for vertical links. The tag
  starts two line heights from the anchor and moves farther only to avoid
  overlaps. Tags can be dragged anywhere. Straight leaders use the fewest valid
  segments: one when the tag lies on the element normal, otherwise the router
  adds perpendicular stubs and connectors while keeping every turn at least 90°.
  Bézier leaders are split into two smoothly joined curve sections. Tags can
  also be rotated manually (Auto/Manual + angle) from the tag overlay or with
  the rotation handle shown on the selected element; the box side facing the
  element is re-selected and the leader stays perpendicular. Polyline leaders
  reject bends below 90°, and Bézier control arms are limited to avoid excessive
  convexity. Arrows default to `grid gap / 4` (20 px → 5 px) and are customizable
  from the **TAGS** sub-panel in the settings **General** tab, which opens the
  family overlay with a live preview. The **Tag settings…** button opens the per-element
  overlay.
- **Editable routes.** `stateRoutesEditable: true` shows a handle on the link for
  type-aware editing: orthogonal waypoint, straight polyline, arc/curve control
  point or circular radius.
- **Selection, clipboard and alignment.** Ctrl/Cmd/Shift+click toggles
  multi-selection; left-drag on empty canvas draws a marquee, while right-drag
  pans the whole graph. Alt+drag and middle-drag also pan.
  Selected groups move and delete as one unit, preserving
  relative spacing; the final translation snaps to whole grid steps. The board
  expands when shapes are dragged beyond the viewport. Ctrl+C/V copy
  and paste, Ctrl+D or **Duplicate** duplicates; shapes duplicate without their
  links unless the links are selected. With two or more shapes selected, the
  floating toolbar aligns left/center/right, top/middle/bottom and distributes
  horizontally/vertically, with magnetic guides during drag. Clicking empty
  canvas clears the selection.
- **Pan and zoom.** Right-drag on empty canvas to pan; Alt+drag and middle-drag
  also pan. Zoom scales the grid, shapes, links, text and tags; **Fit** fits the
  width and **All** fits every component.
- **Dynamic canvas.** The board is the used graphic area plus a 20 px margin on
  every side. It grows automatically when shapes, tags or waypoints move beyond
  the border, scrollbars appear when needed, and **Fit** measures the rendered
  content so the graph is never clipped.
- **Tables.** Toolbar buttons are 28 px high; chart and tables are separated by
  10 px; tabs and headers stay fixed while only rows scroll.
- **Rotation.** Asymmetric shapes can be rotated with the rotation handle that
  appears when the shape is selected, with the context-menu ±90°/reset actions
  or with the exact `rotation` value in the properties panel. Shape, anchors and
  normals rotate together, while polygon anchors stay evenly distributed along
  each edge. Rotation handles snap to the nearest 0/90/180/270° within 6° so
  text and shapes settle easily horizontally or vertically.
- **New shape colors.** Newly inserted shapes use a soft pastel palette.
- **Chart title and colors.** Diagram settings edit the chart title and the
  general shape border color, link color and highlight color. Per-element
  border/link/highlight colors override them; the highlight also colors the
  selected table row.
- **Tag panel.** The tag overlay has a wider two-column layout, tabbed controls
  and a taller live preview with a grid background. Tabs are **Tag**, **Text**,
  **Box** and **Leader & arrow**; the panel height follows the largest tab.
  On/off options are switches, numeric values use a slider with an editable
  number, and the font is a searchable combo box. The font field shows the current
  or default font and the list contains every available family, including the
  selected one. The **Text** tab also exposes the multiline tag text, weight,
  style, decoration, alignment, letter spacing, line spacing, text transform,
  text shadow and shadow color. A **Text settings…** command in every table row
  opens this dialog directly on the **Text** tab; the shape context menu uses
  **Tag settings…**, whose **Text** tab covers the same styling. The shape
  context menu contains only **Shape properties...**, **Tag settings...** and
  **Delete**; rotation is edited in the **Position** tab of shape properties.
  Rotation controls for tags are in the **Tag** tab. Colors use a compact swatch plus a hex field, and the box background also
  has an explicit **Transparent** button. The **Box** tab contains box style,
  background, border color, border width (minimum 1 px), border style and padding.
  **No border** still keeps box style, background and padding selectable and draws
  a borderless box with the chosen background; only the border color/width/style
  controls are disabled. The **Arrow** switch enables the arrow shape/size
  controls, and text direction is enabled with manual rotation: **Always
  horizontal** keeps the text level and disables the rotation value, **Follow
  line** uses the manual angle. **Movable tags** is in the **Tag** tab: an
  internal shape tag moves only inside the shape, an internal link tag moves only
  along the link path. Internal shape tags default to a transparent background
  with no tag box and have leader/arrow controls disabled; internal link tags stay
  on the link line and interrupt it visually. Drag the tag inside the preview to
  see the leader follow it, or drag the round handle to rotate it; the rotation is
  written back to the Rotation field (snapping to 0/90/180/270° within 6°), while
  the preview position stays local. The leader uses the same routing functions as
  the diagram and the link sample is always a straight line with the same color,
  width, dash and arrow marker as the graph.
- **Settings panel.** The gear button opens a tabbed panel: **General** (chart
  title, shape border/highlight colors, theme, show-anchors switch and one row of
  GRID, SHAPES, LINKS and TAGS buttons, save-as-default/reset),
  **Versions** (pending changes, commit and history) and
  **Data** (import and JSON/CSV/SVG/PNG export). Closing a settings sub-panel
  restores General. **Export JSON** embeds shapes, links and settings, so
  re-importing it reconstructs the diagram exactly. **SHAPES** sets the defaults
  for new shapes (border style/width, dash length/gap, corner radius,
  border/highlight colors, fixed or **Random fill color for new shapes**) in two
  aligned columns with a live preview; per-element properties override them. LINK
  mirrors the per-link panel with two tabs: **Line** (subtype, line
  and highlight colors, width, dash type/length/gap, circular arc radius and
  editable routes) and **Arrows** (arrow size, start/end marker and a live
  preview). These values are the defaults inherited by every link. GRID collects
  grid visibility, grid size and grid color. Link properties add a
  **Direction** button that reverses source and target.
- **Dark theme.** The dark theme uses the main application palette (`#14171c`
  background, `#e6e6e6` text, `#1d2129` controls, `#3d4556` borders) so the chart
  and its overlays match the surrounding app.
- **Canvas and grid.** The state board is at least as large as the viewport, so
  the grid and usable canvas fill the whole view; resizing the viewport grows the
  board automatically. The graph keeps a 20 px margin. **Fit width** fits the
  content horizontally and allows vertical scrolling; **Fit all** fits both
  dimensions and centers the graph.
- **View-only mode.** The grid and the editing chrome are independent options:
  `stateGrid: false` hides the grid, and `stateHandles: false` hides anchors,
  resize/rotation grips, tag and route handles and the floating **+ Add** tool,
  and disables shape dragging, so the diagram can be published for reading only.
  Handles are enabled by default; set `stateHandles: false` for a viewer. Example:

```sui
chart "Read-only diagram" name: "viewer" data: stateGraph chartType: "state"
  stateGrid: false stateHandles: false theme: "dark"
```
- **Undo / redo.** Every diagram change is tracked: moves, resizes, rotations,
  tag drags, waypoints, table edits, additions and deletions. Use the header or
  table buttons, or `Ctrl+Z` and `Ctrl+U`.
- **Keyboard.** Arrows move the selected shape by one cell, Shift+arrows resize,
  Delete/Backspace removes it.

## Run

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
simpleui serve app.sui --watch
```

The ready-to-run desktop build is `dist/app.exe`; `dist/app.sui` is a copy of
the source.

## Complete source
```sui
state stateGraph = "\{\"states\":[\{\"id\":\"a\",\"label\":\"Circle\",\"shape\":\"circle\",\"color\":\"#16a34a\",\"w\":72,\"h\":72,\"x\":216,\"y\":192,\"tag\":\"Circle\",\"tagKind\":\"internal\",\"tagBox\":\"plain\",\"description\":\"Round shape on the grid\"\},\{\"id\":\"b\",\"label\":\"Rectangle\",\"shape\":\"rectangle\",\"color\":\"#3b82f6\",\"w\":144,\"h\":72,\"x\":480,\"y\":168,\"tag\":\"Rectangle\",\"tagKind\":\"external\",\"tagBox\":\"rounded\",\"tagLeader\":\"bezier\",\"tagArrow\":true,\"description\":\"Rectangular shape on the grid\"\},\{\"id\":\"c\",\"label\":\"Triangle\",\"shape\":\"triangle\",\"rotation\":20,\"color\":\"#f59e0b\",\"w\":120,\"h\":96,\"x\":720,\"y\":180,\"tag\":\"Triangle\",\"tagKind\":\"external\",\"tagBox\":\"circle\",\"tagLeader\":\"line\",\"tagArrow\":true,\"description\":\"Triangle with bisector anchors\"\}],\"transitions\":[\{\"from\":\"a\",\"to\":\"b\",\"event\":\"go\",\"label\":\"go\",\"tag\":\"Link\",\"tagKind\":\"external\",\"tagBox\":\"ellipse\",\"tagLeader\":\"bezier-arrow\",\"tagArrow\":true,\"routing\":\"orthogonal\",\"description\":\"Orthogonal link between the two shapes\"\}]\}"

view main {
	title "Circle, rectangle and triangle"
	col class: "state-page" {
		header class: "state-hero" {
			p "SimpleUI diagram editor" class: "state-kicker"
			h1 "Circle + rectangle + triangle" class: "state-title"
			p "One circle, one rectangle, one triangle and one orthogonal link. The diagram and the Shapes/Links tables edit the same entities: add, edit, resize or delete from either view and the other stays in sync." class: "state-subtitle"
		}

		card class: "state-card" {
			chart "Circle, rectangle and triangle" name: "circle_rect_demo" class: "state-chart" data: stateGraph chartType: "state" stateShapeTagStyle: "rectangle" stateShapeTagLink: "bezier-arrow" stateLinkTagStyle: "ellipse" stateLinkTagLink: "bezier-arrow" stateTagFontSize: 11 stateGrid: true stateGridGap: 24 stateGridColor: "#cbd5e1" stateTagsMovable: true stateRoutesEditable: true tableView: "both" tableFilters: false showGrid: false showLegend: false showValues: true interactive: true theme: "light"
		}
	}
}

css "body \{ margin: 0; background: #f3f7fb; \} .state-page \{ max-width: 1180px; margin: 0 auto; padding: 20px; display: flex; flex-direction: column; gap: 14px; \} .state-hero \{ padding: 18px 22px; border: 1px solid #cbd9e8; border-radius: 14px; background: #fff; \} .state-kicker \{ margin: 0 0 4px; color: #0369a1; font-size: 12px; font-weight: 700; letter-spacing: 1.2px; text-transform: uppercase; \} .state-title \{ margin: 0; color: #0f2740; font-size: 28px; \} .state-subtitle \{ margin: 8px 0 0; max-width: 820px; color: #526b84; font-size: 13px; line-height: 1.5; \} .state-card \{ border: 1px solid #d5e1ee; border-radius: 14px; background: #fff; padding: 10px; \} .state-chart \{ width: 100%; height: 560px; \}"
```
