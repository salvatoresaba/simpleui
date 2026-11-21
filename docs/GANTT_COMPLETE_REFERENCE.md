# SimpleUI Gantt: Complete Reference

This is the consolidated, canonical reference for the SimpleUI `gantt` chart
type. It collects the data model, rendering rules, editing workflows, keyboard
contract, persistence, versioning, server integration contract, security
boundaries and known limitations in one place.

The executable source of truth is the implementation. If this document and the
implementation disagree, inspect these files and update this document:

- `components/chart/runtime.go` — parser, scheduler, SVG renderer, routing,
  table, editing, keyboard behavior, persistence and versioning.
- `htmlgen.go` — chart, Gantt, table, toolbar, menu and responsive CSS.
- `jsgen.go` — chart attributes and callback data attributes.
- `typechecker.go` — accepted chart attributes and callback names.
- `guihtml.go`, `gui.go` — designer inspector and source round-trip.
- `examples/gantt_construction/`, `examples/gantt_project/`,
  `examples/gantt_calendar/`, `examples/gantt_wbs/` — runnable examples.

## Overview

A Gantt chart is a time-scaled project view built from tasks with a start and a
finish. It is more than a bar chart:

- the left name column is fixed while the timeline scrolls horizontally;
- bars represent tasks, milestones are diamonds and summary tasks are WBS
  containers;
- dependencies are routed as orthogonal links with arrowheads;
- working days, weekends, holidays and time zones affect scheduling;
- a synchronized editable table mirrors every chart row;
- resources, people quantities, workload and daily costs feed task cost;
- critical path, slack, baselines of actual work, deadlines and constraints
  produce warnings and visual states;
- edits can be undone, auto-saved locally, committed as versions and sent to an
  application callback for server synchronization.

The Gantt is an SVG component with a native HTML table below or beside it. All
data values are inserted as SVG attributes, text nodes or `textContent`; task
labels and payload values never become markup.

### Lifecycle

Every redraw follows this order:

1. Parse the `data` attribute as JSON, XML or CSV.
2. On the first draw, merge a matching local draft if one exists.
3. Read the effective calendar (`weekend`, `holidays`, time zone).
4. Build the WBS tree and reject structural cycles.
5. Optionally run auto-scheduling.
6. Roll up summary dates, progress, work and cost.
7. Validate dates, constraints, deadlines, effort and dependencies.
8. Render bars, groups, WBS headers, progress, actuals and markers.
9. Precompute arrowhead boxes and route every dependency.
10. Render the synchronized table and resources panel.
11. Restore selection, update the commit button and keep the name column sticky.

### Supported targets

The component works in generated web apps and desktop builds. Persistence is
local-browser `localStorage`; server persistence is host-provided through
callbacks. The Gantt runtime performs no network request by itself.

## Quick start

```sui
state plan = "[{\"id\":\"design\",\"task\":\"Design\",\"start\":\"2027-03-01\",\"end\":\"2027-03-05\",\"progress\":40,\"group\":\"Delivery\"},{\"id\":\"build\",\"task\":\"Build\",\"start\":\"2027-03-08\",\"end\":\"2027-03-12\",\"progress\":0,\"group\":\"Delivery\",\"dependsOn\":[\"design\"]}]"

view main {
	chart "Project plan"
		name: "project_plan"
		data: plan
		chartType: "gantt"
		tableView: "both"
		editMode: "persistent"
		tableAddRow: true
		tableDeleteRow: true
		tableSelection: "row"
		criticalPath: true
		timezone: "Europe/Rome"
		weekend: "sat,sun"
}
```

The title is the chart's accessible title. `name` is the stable key used by
local storage; without it the title is used as the key.

## Data formats and fields

### JSON forms

Accepted forms include a task array:

```json
[
  { "id": "a", "task": "Analysis", "start": "2027-03-01", "end": "2027-03-05" },
  { "id": "b", "task": "Build", "start": "2027-03-08", "end": "2027-03-12", "dependsOn": ["a"] }
]
```

or an object with tasks, resources and calendar:

```json
{
  "tasks": [
    {
      "id": "design",
      "task": "Design",
      "start": "2027-03-01",
      "end": "2027-03-05",
      "progress": 40,
      "group": "Delivery",
      "parent": "",
      "resources": [{ "name": "Marta", "units": 2 }],
      "dependsOn": []
    }
  ],
  "resources": [
    {
      "name": "Marta",
      "role": "Designer",
      "maxPeople": 3,
      "dailyCost": 450,
      "notes": "Part time"
    }
  ],
  "calendar": {
    "weekend": "sat,sun",
    "holidays": ["2027-04-02", "2027-04-05..2027-04-06"]
  }
}
```

### CSV

The CSV/TSV reader accepts a header row with these canonical headers:
`task`, `start`, `end`, `duration`,
`progress`, `milestone`, `group`, `parent`, `dependsOn`, `dependencyType`,
`lag`, `constraint`, `constraintDate`, `deadline`, `actualStart`, `actualEnd`,
`remaining`, `work`, `workDone`, `resources`, `resourceUnits`,
`cost`, `color`, `status` and `id`.

### XML

XML rows are read by element/field name with the same canonical names. Keep values as
text content; the runtime escapes and validates them before rendering.

### Task fields

| Field | Meaning |
|---|---|
| `id` | Stable task identity. Dependencies, parents, summaries and audit entries use it. |
| `task` | Visible task label. |
| `start`, `end` | ISO date or ISO date-time value. Numeric positions are rejected. |
| `duration` | Length used when an end value is not supplied. |
| `progress` | Completion percentage, clamped to 0–100. |
| `milestone` | Renders a diamond; a milestone has no usable bar duration. |
| `group` | Visual phase label. It does not turn the task into a WBS summary. |
| `parent` | Parent task id. A parent with children becomes a WBS summary. |
| `dependsOn`, `dependencies` | Predecessor ids or names, comma/semicolon lists, or arrays. |
| `dependencyType` | Task-level fallback link type: `FS`, `SS`, `FF`, `SF`. |
| `lag` | Task-level fallback lag in working days. |
| `links` | Per-predecessor objects with `id`, `type`, `lag`. |
| `constraint` | `ASAP`, `SNET`, `SNLT`, `FNET`, `FNLT`, `MSO`, `MFO`. |
| `constraintDate` | Date used by the selected constraint. |
| `deadline` | Soft finish deadline; overdue tasks are highlighted and warned. |
| `actualStart`, `actualEnd` | Execution dates; rendered as a thin actual-work bar. |
| `remaining` | Remaining work or duration; derived when empty. |
| `work`, `workDone` | Effort and completed effort. When both exist, progress derives from the ratio. |
| `resource`, `resources` | One or several resource names. Array entries may be strings or `{name, units}`. |
| `resourceUnits` | Name-to-quantity map for assignments. Default quantity is `1`. |
| `cost` | Manual task cost unless every assigned resource has a daily cost. |
| `status` | Optional displayed status; the runtime also derives task state. |
| `color` | Per-task bar color. Task color is persisted locally by task id. |

Dependency entries may be strings (`"A"`, `"A[SS+2]"`) or objects
(`{"id":"A","type":"SS","lag":2}`). A missing per-link type falls back to the
task `dependencyType`; a missing per-link lag falls back to task `lag`.

### Date and time values

Date-only values are calendar dates. Date-time values retain time of day and
are interpreted using the chart `timezone` (default `UTC`). Axis ticks,
tooltips, table date rendering and calendar-day matching all use that zone.
Date-only values remain stable calendar dates regardless of the host time zone.

## Scheduling model

### Dates and working days

- `start` and `end` define the planned interval.
- `duration` is the planned length when an end is absent.
- Effective duration counts working days between start and finish using the
  active calendar.
- Dependency lags and auto-scheduling use working days.
- Moves, resizes and milestone edits snap to working dates.
- A non-working date typed into a table cell is rejected with an inline error
  and status line.
- Weekends and holidays are hidden or shown visually, but validation remains
  active even when their bands are hidden.

### Milestones

A milestone is rendered as a diamond and is treated as a zero-duration point.
Moving a milestone moves its start and end together. It can still participate in
dependencies, deadlines, constraints and critical-path calculations.

### Constraints

Supported constraint types:

- `ASAP` — as soon as possible.
- `SNET` — start no earlier than date.
- `SNLT` — start no later than date.
- `FNET` — finish no earlier than date.
- `FNLT` — finish no later than date.
- `MSO` — must start on date.
- `MFO` — must finish on date.

Warning validation compares the planned values with the constraint date. Table
selects and filters list all seven types with descriptive labels even when the
current data only uses `ASAP`.

### Deadlines

A deadline is a soft constraint. The task remains schedulable, but a planned
finish after the deadline produces a warning, a chart marker and a highlighted
table row. Deadline markers are larger than ordinary annotations, have a wide
interaction hit area and can be dragged horizontally when interactive. The table
Deadline cell updates on drop.

### Actual, remaining and effort progress

`actualStart` and `actualEnd` record execution. The chart draws actual work as a
thin top overlay on the task bar.

When `work` and `workDone` are both present, Progress is derived from
`workDone / work`. Editing Progress updates Work done; editing Work done
recomputes Progress. When they are absent, progress remains the explicit
percentage. Remaining work is derived from duration and progress when empty.

### Status, warnings and critical path

The runtime derives task states such as not started, in progress, completed,
blocked and late. `statusColors` applies status-specific styling. Warnings cover
invalid dates, non-working dates, constraint violations, deadlines, effort
mismatches and dependency/lag problems. Warnings appear in the warning banner
and on affected table rows.

`criticalPath: true` computes zero-slack tasks and critical dependency links from
the dependency network, constraints, working-day durations and the active
calendar. Critical links are thicker, not a separate bright color, so the chart
stays readable.

`showSlack` exposes slack columns and values. `showCosts` exposes cost data.
`showResources` appends resource names to chart labels and shows resource UI.

### Auto-scheduling

`autoSchedule: true` recalculates starts from link type/lag, constraints and
date values. It performs an additional roll-up/scheduling pass for WBS
hierarchies so successors of a summary use the rolled-up summary dates. Cycles
are reported as errors instead of drawing an invalid schedule.

Link scheduling semantics:

- `FS` — successor starts after predecessor finish plus lag.
- `SS` — successor starts after predecessor start plus lag.
- `FF` — successor finishes after predecessor finish plus lag.
- `SF` — successor finishes after predecessor start plus lag.

Link type affects scheduling and dash style, but not the geometry rule: every
link still exits the source's right edge and enters the target's left edge.

## WBS, groups and summaries

There are two different hierarchy concepts:

### `group`

`group` is a visual phase. Tasks with the same group value are rendered
contiguously in the chart and table, under a group header. A grouped task
remains a real task with its own bar, dependencies, progress and resource
assignments.

Group actions:

- Assign individual tasks through the table Group cell.
- Select several rows and use **Assign to group...**.
- Select tasks with Ctrl/Cmd + click in the chart, or select table rows, then
  use **Create group from selection**.
- A collapsed group shows one duration-weighted summary bar.

`group` never changes the task into a summary and never creates a WBS parent.

### `parent` / WBS

`parent` points to another task id and creates a real WBS outline. A task with
children becomes a summary container:

- expanded summary: header row only, no bar;
- collapsed summary: one summary bar replacing its children;
- child rows are indented in both chart and table;
- start, finish, progress, work and cost roll up from descendants;
- dependency links crossing the summary boundary attach to the summary bar;
- dependencies internal to a collapsed group are hidden;
- clicking the collapsed bar, name, caret or table toggle expands it again.

Summary dates are derived. Keyboard edits on a summary are rejected with an
explanation instead of silently changing derived values.

## Resources, workload and cost

### Registry

The Resources table stores:

| Field | Meaning |
|---|---|
| Name | Unique resource key used by task assignments. |
| Role | Free text classification. |
| Max people | Maximum concurrent headcount; `0` or empty means unlimited. |
| Daily cost | Cost of one person for one working day. |
| Notes | Free text notes. |

The registry is stored locally in
`sui.chart.resources.<chart-key>`. JSON data may provide an initial registry.
Any resource name typed in the Tasks table is added to the registry
automatically. Renaming a resource updates every task assignment. Deleting a
resource that is still assigned is refused. **Reset saved** clears the local
registry and rebuilds it from source data.

### Assignments and quantities

A task can have multiple resources. Each assignment has a people quantity:

- `{ "name": "Marta", "units": 2 }` inside `resources`;
- `"resourceUnits": { "Marta": 2 }`;
- `resource: "Marta, Luca"` is not accepted; use `resources`.

The Resources table cell shows removable chips with `×N` quantities. A small
plus button opens a resource-type combo and a quantity field. Confirming adds
the assignment; each chip's quantity is editable.

### Overload protection

The sum of assigned people quantities for concurrent tasks must not exceed
`maxPeople`. Interactive edits, pastes, date changes, resizes and max-people
reductions that would create a new overload are rejected and reverted with a
status message. Existing overloads show warnings rather than being silently
destroyed.

### Workload dialog

The chart menu's **Resource workload...** dialog reports, per resource:

- role;
- assigned tasks;
- workload;
- work;
- completed work;
- peak concurrent demand;
- maximum people;
- daily cost.

Workload uses `work` when present, otherwise working-day duration. A task
assigned to several resources counts in full for each resource. The current
diagnostic layer does not provide a time-phased histogram over the whole plan.

### Automatic task cost

If every resource assigned to a task has a daily cost, the task Cost cell becomes
read-only and is calculated as:

```text
sum(resource dailyCost × units) × working-day duration
```

If any assignment lacks a daily cost, the manual Cost value is retained.
Changing a resource rate, assignment quantity or working calendar recalculates
affected automatic costs. WBS summaries roll up child cost when available.

## Calendar, timezone and period bands

### Weekend and holidays

- `weekend: "sat,sun"` is the default. Weekday names or numbers `0`–`6`
  (Sunday = 0) are accepted. `weekend: "none"` disables weekend exclusion.
- `holidays` accepts exact dates (`2027-12-25`), inclusive ranges
  (`2027-12-24..2027-12-26`) and recurring dates (`*-12-25`, `12-25`).
  Recurring ranges can wrap across the year boundary.
- `timezone` accepts an IANA zone such as `Europe/Rome`; the default is `UTC`.
- Date-only values remain calendar dates; date-time values use the selected
  zone.
- `showNonWorkingDays` defaults to `true`. It only controls the visibility of
  weekend/holiday bands; it never disables scheduling or date validation.

### Calendar overlay

The chart menu's **Calendar...** dialog edits non-working weekdays and holiday
dates or closure ranges through one date-range picker: click once for one day,
click twice for a range. The calendar is stored per chart in
`sui.chart.calendar.<chart-key>`. It can also be supplied in JSON as
`{"calendar":{"weekend":"sat,sun","holidays":[...]}}`. `onCalendarChange`
receives the saved calendar JSON.

### Timeline scale and period bands

Gantt `zoom` ranges from `0.5` to `24`. `ganttScale` selects `auto`, `day`,
`week`, `month` or `year` independently of zoom.

At month scale, alternating month bands mark month boundaries. At week scale,
alternating week bands start on Monday and non-working-day bands are suppressed
even if `showNonWorkingDays` is on. At day scale, alternating day bands are
drawn; when non-working days are visible, those days use the dedicated
weekend/holiday color instead of the alternating day color. Period bands are
theme-aware and sit behind grid, bars and links.

Weekly ticks are aligned to Monday. Day ticks retain weekday names. The year
label remains visible at the top of the timeline.

## Chart attributes

The table below lists the main Gantt-facing attributes. Component-level `class`,
`style`, width and height are normal SUI component attributes.

| Attribute | Meaning |
|---|---|
| `chartType` | Must be `"gantt"`. The SUI attribute is `chartType`, not `type`, because `type` is reserved. |
| `data` | JSON, XML or CSV data string; may be a reactive state expression. |
| `format` | Optional data format hint. |
| `name` | Stable chart key for local persistence. Falls back to the title. |
| `theme` | `dark` or `light`; drives structural chart/table tokens. |
| `editMode` | `none` is view-only, `editable` allows edits for the page session and `persistent` keeps task/resource/calendar edits across reloads. |
| `overflow` | `scroll` keeps the frame fixed and scrolls content; `fit` squeezes; normal chart zoom controls apply to non-Gantt charts. |
| `zoom` | Initial Gantt timeline zoom, `0.5`–`24`, default `1`. |
| `ganttScale` | `auto`, `day`, `week`, `month`, `year`. |
| `showGrid` | Draw timeline grid lines. |
| `showValues` | Show value/progress labels where applicable. |
| `showLegend` | Show the chart legend. |
| `showProgress` | Show progress labels/overlays according to the active mode. |
| `showResources` | Append assigned resources to chart/table labels. |
| `showSlack` | Show slack data. |
| `showCosts` | Show cost data. |
| `statusColors` | Apply status-specific visual styling. |
| `criticalPath` | Highlight zero-slack tasks and critical links. |
| `autoSchedule` | Recalculate starts from links, constraints and lags. |
| `holidays` | Holiday dates, ranges and recurring dates. |
| `weekend` | Non-working weekdays or `none`. |
| `timezone` | IANA time zone; default `UTC`. |
| `showNonWorkingDays` | Show/hide weekend and holiday bands; default `true`. |
| `progressColumn` | Add a fixed percentage column; default `false`. When shown, inline percentages beside bars are hidden. |
| `linkColor` | Dependency line color; default neutral graphite. |
| `linkWidth` | Link stroke width, `0.5`–`8`, default `1`. |
| `linkGap` | Minimum clearance between vertical runs and from bars/arrows. Default `4`, range `0`–`24`. |
| `linkStyle` | `auto`, `solid`, `dashed` or `dotted`. |
| `arrowType` | `triangle`, `open` or `none`. |
| `arrowSize` | Arrowhead size, `2`–`20`; default around `5`. |
| `barBorderColor` | Optional task, summary and milestone border color. |
| `tableFilters` | Enable column filters and filter menus. |
| `tableAddRow` | Show the table **+ Row** action; default `false`. |
| `tableDeleteRow` | Show row delete controls. |
| `tableSteppers` | Show visible numeric steppers in focused numeric cells. |
| `tableContextMenu` | Enable table context menu copy/paste/delete/export. |
| `tableSelection` | `row`, `column`, `cell` or `none`; default `row`. |
| `tableView` | `chart`, `table` or `both`. |
| `onTaskChange` | Optional callback receiving the current task payload after an edit. |
| `onCalendarChange` | Optional callback receiving the saved calendar JSON. |
| `onResourceChange` | Optional callback receiving the resource registry JSON. |
| `onDataChange` | Non-Gantt chart-table edit callback. |
| `onCommit` | Optional callback receiving a committed Gantt version for host/server synchronization. |

### Dimensions and scrolling

The Gantt has no separate internal width/height attributes. Set the component
size with normal `width`/`height` or CSS. The frame stays fixed while the
timeline and table scroll. The left name column is sticky and separately
resizable by dragging its right border. The width is persisted per chart.
Horizontal scrolling masks timeline content and tick labels at the name
separator, so they cannot slide underneath the names.

The Gantt stage grows when `linkGap` needs more vertical row corridors or
horizontal lane space. The viewport scrolls instead of squeezing routes below
the configured clearance.

## Synchronized table

`tableView: "both"` shows chart and table together; `"table"` hides the chart and
shows only the accessible data table. Gantt tables have two tabs:

- **Tasks**
- **Resources**

The tab bar remains pinned while the table scrolls. The first identifying column
is sticky during horizontal scrolling.

### Task columns

Tasks expose, subject to the active mode, these columns:

- Task;
- WBS parent;
- Start;
- End;
- Deadline;
- Actual start;
- Actual end;
- Duration;
- Effective duration;
- Remaining;
- Progress;
- Work;
- Work done;
- Milestone;
- Status;
- Group;
- Resources;
- Depends on;
- Link type;
- Lag;
- Constraint;
- Constraint date;
- Slack;
- Cost.

Dates, progress, work, cost and slack use their appropriate editors and
formatters. Inline edits validate on blur/Enter and redraw the timeline. Paste,
stepper and multi-cell operations validate the same rules.

### Filtering, sorting and layout

- Each header has an Excel-style filter popup with contains search, a distinct
  value checklist and numeric/date min/max fields.
- Clicking a header sorts; Shift+click adds a multi-column sort.
- Headers can be reordered by dragging and resized by dragging the right edge
  or focusing the resize handle and pressing Left/Right/Home.
- Double-clicking a resize handle auto-fits the column.
- Widths, ordering and visibility persist in table state.
- **Reset column widths** restores automatic sizing.
- Filters, sort, column order, visibility and widths persist in
  `sui.chart.table.<chart-key>`.

### Selection and clipboard

- `tableSelection` supports row, column, cell and none.
- The row-number gutter selects/drags rows; the strip above headers selects/drags
  columns; dragging the body draws a rectangular cell selection.
- Ctrl/Cmd+click toggles; Shift+click extends; Alt+click on a header sorts.
- Selected rows, columns and cells synchronize with the chart selection.
- Context menus support copy, paste, delete selected rows and CSV/TSV/JSON/JSONL
  export of the active selection.
- Copy writes TSV; paste requests clipboard-read permission and fills selected
  cells, rows or columns.
- Status messages use an `aria-live` region.

### Undo and redo

Editable chart and Gantt tables keep a 50-step in-memory undo/redo history for
cell edits, steppers, paste, row insertion and row deletion. Use the chart/table
menus or `Ctrl+Z`, `Ctrl+Shift+Z` / `Ctrl+Y`. Focused text inputs retain native
browser text undo.

### Row insertion and deletion

`tableAddRow: true` shows **+ Row**. On Gantt it appends a task with a unique id,
a one-day range after the last task, the last group and no dependencies. The new
row is selected and focused; filters clear and the chart redraws. `tableDeleteRow`
adds delete controls.

### Virtual scrolling

Tables above 300 rows use windowed virtual scrolling. Filters, sorting and
exports still operate on the full filtered set, not only visible rows.

## Visual grammar

### Name column

The name column is a sticky left pane. It contains:

- task names, indented by WBS depth;
- group headers and WBS summary headers;
- optional progress percentage column when `progressColumn` is enabled;
- wrap up to three lines when row height permits, otherwise two;
- a horizontal boundary line and resize handle;
- group/WBS disclosure carets.

Long labels also wrap in the table Task editor up to a capped textarea height.
The left boundary masks the timeline and tick labels while scrolling.

### Task bars

- Regular task: rounded horizontal bar from start to finish.
- Milestone: diamond centered on its date.
- Summary: header-only when expanded, summary bar when collapsed.
- Status colors and task-specific colors override series styling.
- Progress is rendered as a full-height, inset, rounded overlay instead of a
  thin bottom strip. Summary progress uses the same treatment.
- Actual work is a thin top overlay.
- Critical tasks and warnings use distinct strokes/classes.
- Hover raises contrast; selection synchronizes with the name and table row.
- Focus visible uses a strong outline/stroke without changing layout.

### Group and WBS bands

Group and WBS header rows use a tinted band and bold labels. Collapsed group
summary bars use the group color and duration-weighted progress. Children are
indented in both chart and table.

### Markers

- Today line when the current date lies in range.
- Deadline line and flag, with an 18px-wide hit area for dragging.
- Weekend, holiday, closure and period bands behind the plot.
- Year label above the timeline.
- Arrowheads and link entry/exit dots.

### Themes

Dark and light themes control structural tokens. The series palette remains
author-controlled through CSS variables/`style: cssSafe(...)`. Selection tokens
use a cyan family in dark mode and a blue-tinted pale family in light mode.
Table hover backgrounds are neutral and light-mode row hover must not leave a
dark band.

Important token families:

- `--sui-chart-bg`, `--sui-chart-fg`, `--sui-chart-muted`, `--sui-chart-grid`,
  `--sui-chart-axis`;
- `--sui-chart-series-1` … `--sui-chart-series-8`;
- `--sui-chart-gantt-weekend`, `--sui-chart-gantt-holiday`,
  `--sui-chart-gantt-today`, `--sui-chart-gantt-dependency`,
  `--sui-chart-critical`, `--sui-chart-gantt-group-bg`,
  `--sui-chart-gantt-bar-border`;
- `--sui-chart-gantt-progress`, `--sui-chart-gantt-period-a/b`;
- `--sui-chart-table-selection-bg`, `--sui-chart-table-selection-border`,
  `--sui-chart-table-focus-row-bg`, `--sui-chart-table-focus-cell-bg`,
  `--sui-chart-table-focus-color`, `--sui-chart-table-hover-bg`;
- control tokens for menus, overlays, calendars, filters and pickers.

`docs/TABLE_COLORS.md` and `docs/table_colors.html` contain the detailed table
color audit and visual swatches.

## Dependency routing

### Link scheduling types

`FS`, `SS`, `FF` and `SF` control scheduling and dash style. Geometry is always
the same: exit the source's right edge, enter the target's left edge.

### Geometry rules

1. All outgoing links from one task share one source exit port at the centre of
   the source bar's right edge.
2. All incoming links to one task share one target entry port at the target
   bar's left edge.
3. A link begins with a horizontal stub of at least `4px` from the source side.
4. The arrow tip is centred exactly on the target's left edge.
5. The arrow tail is horizontal and at least `4px` long; all incoming routes meet
   the left end of the tail.
6. A vertical run never ends directly against a bar. The final approach is
   horizontal.
7. Vertical runs keep `linkGap` from task borders, arrowheads and other vertical
   runs.
8. Only the shared source/entry port coincidence is permitted.
9. A link to a lower target never rises; a link to an upper target never
   descends.
10. A collapsed group exception: when an internal child ends before the group
    summary ends, the outgoing link leaves the summary bar's bottom edge at the
    child's finish x, then enters a free lane. This initial downward stub is
    exempt from the monotonic rule.
11. The route stays inside the plot area and never runs under the fixed name
    column.
12. Routes are orthogonal. A shortest-path cost combines horizontal and
    vertical length with a tiny bend tie-breaker.
13. Existing bars, arrow boxes, horizontal runs and vertical runs are routing
    obstacles.
14. Vertical lane candidates include bar edges, target/source offsets, existing
    vertical runs and a global lane grid. Extra source and target approach lanes
    are tried when the source ends close to the target, which is common after
    horizontal zoom.
15. If a route cannot satisfy every clearance, the implementation tries the
    shortest bar-free fallback and then outer corridors. Dense networks and
    extreme zoom should still be audited visually; routing is heuristic, not a
    formal solver.

### Zoom and time-scale re-routing

Changing timeline zoom or `ganttScale` redraws the time axis and **re-routes**
tasks and links. It is not a CSS transform: bars, links, labels and the table are
recomputed. The horizontal lane space grows with the stage when needed. All
vertical-gap and bar-clearance rules remain in force after a scale change.

Dragging or resizing a task uses a throttled live redraw: links re-route,
warnings and the table update, and the undo history entry commits on drop.

### Link editing and keyboard routing

- Drag the small connector under a bar's right edge onto another bar or name to
  create an FS dependency.
- Click a dependency chip to edit its type and lag.
- Focused dependency link: Enter/Shift+F10 opens its menu; Delete removes it.
- The link connector is focusable; Enter opens the target picker.
- A dependency that would create a cycle is rejected with a status message.

## Keyboard and accessibility

### Chart-level keyboard selection

On the chart root, arrow keys move through data marks, Home/End jump, Enter/Space
select and Escape clears the active tooltip/selection. The chart root is
focusable and has an accessible label.

### Gantt task keyboard editing

A task row is a focusable group with `aria-roledescription="Gantt task"`,
accessible dates/progress and visible focus styling.

| Key | Action |
|---|---|
| `ArrowLeft` / `ArrowRight` | Move task by one working day. |
| `Shift+Arrow` | Resize only the finish by one working day. |
| `Alt+Arrow` | Resize only the start by one working day. |
| `ArrowUp` / `ArrowDown` | Move focus to the previous/next task row. |
| `Enter` | Open the task menu. |
| `Space` | Select/focus the task. |
| `Escape` | Return focus to the chart root. |

Summary rows reject direct key edits because their dates are derived. Keyboard
edits use the same overload checks, history, table synchronization, notification
and warning validation as drag/resize.

### Dependency and routing keyboard access

- The dependency connector is tabbable and opens an accessible target picker
  with Enter/Space.
- A focused dependency exposes `aria-keyshortcuts` for Enter, Shift+F10 and
  Delete.
- The link editor and target picker are modal dialogs with focus management and
  Escape handling.
- After redraw, focus is restored to the affected task, adjacent link or chart
  root instead of being lost.

### Table keyboard access

Rows, headers, cells, filters and resize handles are focusable. Header Enter
sorts, arrow keys navigate, Shift extends and Ctrl/Cmd toggles. Column resize
handles support Left/Right/Home. Status and warning updates use live regions.

### Remaining accessibility boundary

Drag and resize are fully replaceable by keyboard operations. Manual route
shape editing is not exposed as a separate lane-offset model; routing follows
the deterministic rules and is managed through dependency creation, type, lag
and deletion. Screen-reader output is designed around semantic labels, live
status messages and native table controls.

## Autosave, commit, versions and audit

### Local autosave

Every task/resource edit is auto-saved to a local draft:

```text
sui.chart.gantt.draft.<chart-key>
```

The draft contains the normalized task payload, resources, source-format hint,
timestamp and the last commit baseline. It survives reloads. When a host
callback writes the chart data back as JSON and the source string changes
format, the draft is still re-associated by identical task ids. A draft is not a
committed version and does not create audit history entries.

`onTaskChange` remains available for application state and server integration,
but it is no longer required for basic persistence.

### Commit button

The Gantt header has a **Commit** button. It shows the number of pending changes
and is disabled when nothing is pending. Commits are explicit; ordinary edits do
not create versions.

The commit dialog collects:

- author (persisted in `sui.chart.gantt.author`);
- commit message;
- a summary of pending changes.

The chart menu also exposes **Commit changes...** and **Version history...**.

### Version object

A commit stores one immutable version:

```json
{
  "id": "v...",
  "revision": 2,
  "baseRevision": 1,
  "at": 1770000000000,
  "author": "Local user",
  "message": "Update tasks",
  "chart": "n:project_plan",
  "payload": [],
  "resources": [],
  "changes": []
}
```

`payload` is the normalized task snapshot; `resources` is the resource registry
snapshot. `changes` is a field-level audit diff from the previous committed
baseline, including created/deleted tasks and resource-registry changes.

### Audit diff

Audited task fields include label, start, end, progress, milestone, group, parent,
dependencies/links, dependency type/lag, constraints/date, deadline, resources,
resource quantities, actual dates, remaining, work/work done, cost and color.
Each audit entry has task id, task label, field, before value and after value.
Summary fields may appear in the diff because derived summary dates/progress/cost
change when a child changes.

Versions are stored under:

```text
sui.chart.gantt.versions.<chart-key>
```

The local store keeps the most recent 40 versions. **Version history** lists
revision, timestamp, author, message, change count and expandable audit details.
Restoring a version loads its task and resource snapshot, updates the commit
baseline and does not create a new version. It does notify the application and
save the restored state as the current draft.

### Server synchronization contract

`onCommit` and the DOM event `sui:gantt-commit` receive the version object above.
The host can:

1. authenticate the author and authorize the project;
2. compare `baseRevision` with the server revision using `If-Match` or a
   revision predicate;
3. persist `payload`, the resource registry, `changes` and the commit message;
4. return `409 Conflict` with the server snapshot/revision if stale;
5. reload or reconcile and let the user commit again.

A host callback example:

```sui
async fn ganttCommitted(version: string) -> null {
	// Validate the payload and POST it to the application server.
	// The server owns authentication, authorization and conflict handling.
}
```

The component deliberately performs no network I/O. This keeps the chart
compatible with the SimpleUI security model and leaves transport, credentials,
server authorization, real-time broadcast and multi-user merge policy to the
host application. Without a host endpoint, versions remain local to the browser.

## Persistence keys

| Key | Contents |
|---|---|
| `sui.chart.calendar.<chart-key>` | Weekend, holidays and closures. |
| `sui.chart.resources.<chart-key>` | Resource registry. |
| `sui.chart.gantt.nameWidth.<chart-key>` | Resizable task-name column width. |
| `sui.chart.gantt.colors.<chart-key>` | Per-task colors. |
| `sui.chart.colors.<chart-key>` | Series colors. |
| `sui.chart.selected.<chart-key>` | Last selected task index. |
| `sui.chart.table.<chart-key>` | Table filters, sort, column order/visibility/widths. |
| `sui.chart.gantt.draft.<chart-key>` | Auto-saved working draft and baseline. |
| `sui.chart.gantt.versions.<chart-key>` | Committed versions and audit diffs. |
| `sui.chart.gantt.author` | Last used commit author. |

`<chart-key>` is `n:<name>` when the chart has a `name`, otherwise
`t:<title>`.

## Exports and printing

Chart menu exports:

- SVG;
- PNG;
- CSV;
- TSV;
- JSON;
- JSONL.

Table Settings exports visible/filtered rows and columns as CSV, TSV, JSON,
JSONL, XLSX or HTML. Export options control headers, hidden columns and
original/current row order. **Export selection** restricts rows and columns to
the active selection. **Print table...** prints the filtered table.
**Print / Save PDF (chart + table)...** opens the browser print dialog with the
current chart SVG and filtered table, allowing browser PDF save.

Spreadsheet exports preserve header styling and numeric alignment. Formula-like
text remains inert text; CSV for spreadsheet applications supports formula
escaping so values cannot execute as formulas.

Chart exports contain all rows; table exports use current filters/order.

## Designer workflow

In the designer, select a `chart` and set **type** to Gantt. The source attribute
is `chartType: "gantt"` because `type` is reserved by the language.

The Chart inspector is type-aware and hides irrelevant fields. Gantt settings
cover:

- data overlay;
- edit mode (none, editable or persistent);
- table view and table actions;
- critical path;
- auto schedule;
- holidays, weekend, timezone and non-working-day visibility;
- time scale, timeline zoom and progress column;
- link color/width/gap/style and arrow type/size;
- source/target callbacks including `onTaskChange`, `onCalendarChange`,
  `onResourceChange` and `onCommit`.

The data overlay validates JSON/XML/CSV before Apply. Literal JSON inside SUI
source escapes braces as `\{` and `\}`; prefer a state string for larger data.

Use **Commit** for a version checkpoint, **Version history...** for audit, and
the Resources tab for the registry. The tasks table edits normal task fields;
the Resources table edits Name, Role, Max people, Daily cost and Notes.

## Verification checklist

Always run:

```bash
go test -count=1 ./...
go vet ./...
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

Manual Gantt checks:

- data parses and type-compatibility errors are visible;
- chart and table show the same tasks and resources;
- long names wrap in chart and table;
- selection synchronizes bar, name, row and cell;
- table filters, sort, reorder, resize, auto-fit and export work;
- task move/resize and keyboard move/resize validate working days;
- non-working dates are rejected in table edits;
- dependencies can be created by drag and by keyboard target picker;
- cycles are rejected;
- link type/lag chips edit the correct link;
- vertical routing clears bars, arrows and other vertical runs by `linkGap`;
- all links leave/enter at the shared right/left ports;
- collapsed-group bottom exits use the child finish x;
- zoom and every `ganttScale` re-route tasks/links correctly;
- progress column hides inline progress percentages;
- name-column resize persists and masks tick labels while scrolling;
- non-working-day bands behave correctly at day/week/month scales;
- overload edits are rejected and existing overloads warn;
- automatic cost follows resource rates and working-day duration;
- autosave survives reload, including host JSON write-back;
- Commit is disabled with no changes and creates one version otherwise;
- audit details show field-level before/after values;
- restoring a version updates tasks/resources and focus;
- `onCommit`/`sui:gantt-commit` receive revision and baseRevision;
- `onTaskChange`/`onCalendarChange`/`onResourceChange` still fire when supplied.

Routing stress test method:

1. Open a construction-style chart with many dependencies.
2. Set timeline zoom through `0.5`, `0.75`, `1`, `1.5`, `2`, `5`, `12`, `24`.
3. Repeat with `ganttScale` `auto`, `day`, `week`, `month`, `year`.
4. Parse every dependency path and every bar rectangle.
5. Assert that overlapping vertical runs are at least `linkGap` apart.
6. Assert that vertical runs clear task borders by `linkGap`.
7. Assert no vertical run ends directly against a bar.
8. Check horizontal stubs and arrow-entry tails are at least 4px.
9. Check collapsed-group bottom exits.

## Known limitations and future work

These are deliberate boundaries or open work, not hidden behavior:

- No resource leveling or automatic allocation.
- No per-resource calendars or time-varying availability.
- Workload is an aggregate summary, not a time-phased histogram.
- No earned-value management (PV/EV/AC), actual-cost roll-up or project
  financial baseline.
- Baselines and forecast values were removed and are out of scope unless
  reintroduced as a separate model.
- No recurring-task templates, multi-project scheduling or cross-project links.
- No drag-and-drop WBS reordering; hierarchy is edited through WBS parent values
  and group actions.
- No native Microsoft Project/XLSX Gantt import.
- The Resources table does not have full parity with Tasks filters, grouping,
  export and virtual scrolling.
- The chart SVG is not row-virtualized for very large task sets; the table is
  virtualized above 300 rows.
- Link routing is heuristic. It explicitly handles the documented spacing rules
  and zoom re-routing, but extremely dense networks at minimum zoom should still
  be treated as a routing stress case. No formal constraint solver is used.
- Version history and audit are local-browser storage; a shared multi-user
  server is not bundled. Server sync requires a host endpoint and conflict
  policy.
- UI runtime strings are English; designer/runtime localization is not complete.
- Full keyboard interaction covers task move/resize, dependency creation,
  editing and deletion, but manual lane-shape override is not represented in the
  data model.

## Implementation map

Key runtime functions and responsibilities:

- `parseGanttData` — JSON/XML/CSV parsing and task normalization.
- `drawGantt` — chart geometry, bars, summaries, period bands, markers, links,
  name column and commit toolbar state.
- `ganttWbsTree`, `applyGanttWbsRollup` — parent validation and derived summary
  values.
- `autoScheduleGantt` — dependency/constraint scheduling.
- `ganttTaskWarnings` — date, effort, constraint, deadline and dependency
  warnings.
- `ganttPathOverlaps`, `ganttPathHitsBars`, `ganttPathHitsArrowheads` — routing
  clearance checks.
- `ganttKeyboardMove`, `bindGanttKeyboard` — keyboard task editing.
- `openGanttDependencyPicker`, `openGanttLinkEditor` — keyboard/picker routing
  management.
- `ganttDraftStorageKey`, `saveGanttDraft`, `loadGanttDraft` — local autosave.
- `commitGanttChanges`, `openGanttCommitDialog`, `openGanttHistory` — versions,
  audit and explicit commit.
- `renderGanttTable`, `renderGanttResourceTable` — synchronized table and
  resources.
- `bindZoom`, `ganttTimelineZoomAt` — timeline zoom and scroll-centered redraw.
- CSS in `htmlgen.go` under `.sui-chart-gantt-*`, `.sui-chart-table-*`,
  calendar/link/resource dialogs and chart controls.

Examples:

- `examples/gantt_construction/` — light theme, resources, WBS/groups, routing
  and commit/autosave.
- `examples/gantt_project/` — dark workspace plan with resources and critical
  path.
- `examples/gantt_calendar/` — time-scale, range, recurring holiday and
  calendar controls.
- `examples/gantt_wbs/` — WBS parent/child summaries, collapse, auto-schedule
  and dependency cycles.
- `examples/chart/` — interactive Gantt demonstration among other chart types.

## Glossary

- **WBS** — work breakdown structure; a parent task whose dates and values roll
  up from children.
- **Group** — a visual phase label; grouped tasks remain independent tasks.
- **Summary** — a collapsed WBS parent represented by one bar.
- **Milestone** — a zero-duration diamond marker.
- **Link type** — FS, SS, FF or SF scheduling relationship.
- **Lag** — working-day delay applied to a dependency.
- **Slack** — scheduling float; zero-slack tasks form the critical path.
- **Effective duration** — working-day count over a planned interval.
- **Link gap** — minimum clearance around vertical dependency runs.
- **Draft** — auto-saved working state; not a version.
- **Commit** — explicit immutable version plus audit entry.
- **Base revision** — server revision against which a commit was made.
- **Shared port** — the common source exit or target entry of multiple links.
- **Period band** — alternating month, week or day background strip.
