# Gantt calendar controls example

This is a complete, runnable demonstration of the Gantt calendar attributes
`holidays`, `weekend` and `timezone`. It also exercises the interactive table:
Task, Parent, Start, End, Progress, Group and Depends on are editable, and edits
are sent back through `onTaskChange`. The `Design` task also demonstrates
execution tracking: **Actual start/end**, **Work** and **Work done** drive
work-based Progress, **Remaining** and the actual bar. The table also exposes
the richer WBS, scheduling, resource and cost fields described below.

## Run it

A desktop build is included in `dist/`:

- **Windows:** double-click `examples/gantt_calendar/dist/app.exe`.
- Keep the `assets` folder next to `app.exe`.

Rebuild and launch from this folder:

```powershell
./run.ps1
```

or:

```bat
run.bat
```

The scripts build `app.sui` with:

```bash
simpleui build examples/gantt_calendar/app.sui --target desktop --out examples/gantt_calendar/dist
```

For a web build:

```bash
simpleui build examples/gantt_calendar/app.sui --out examples/gantt_calendar/dist
simpleui serve examples/gantt_calendar/app.sui --watch
```

## What to verify

1. **Timezone** - click `UTC`, `Europe/Rome` and `America/New_York`. Start and
   End values in the table change immediately. The kickoff task crosses the date
   boundary in New York (`2026-11-01 19:30`).
2. **Weekend** - click `Sat/Sun`, `Fri/Sat` and `None`. The shaded bands move or
   disappear.
3. **Holidays** - click `Exact`, `Range`, `Recurring`, `Recurring range`,
   `Mixed` and `None`. The red bands update. `Recurring range` uses
   `*-12-24..*-01-06`, visible on the year-end task.
4. **Table editing** - edit a Start or End value. Use local wall-clock time in
   the selected timezone. The value is converted to the stored timestamp and
   redrawn in the table, chart and timeline.
5. **Synchronization** - edit a cell or drag a task bar. The chart, table and
   status line update through `onTaskChange`.
6. **Add a row** - click **+ Row** (enabled by `tableAddRow: true`). A new task
   is appended, selected and focused; the chart, table and `onTaskChange`
   payload update. `Ctrl+Z` undoes it.
7. **Table tools** - header filters combine text search, value checklists and
   min/max ranges; Settings can copy visible rows or open a print view. The
   identifying first column stays fixed while scrolling horizontally.
8. **Working duration** - `Duration` is the calendar difference; **Effective
   duration** counts working days using the active weekend and holiday rules.
   Use it when calculating effort or cost.
9. **Table options** - the Gantt enables `tableDeleteRow`, `tableSteppers`,
   `tableContextMenu` and row selection. Each row ends with a delete button;
   numeric cells show two visible minus/plus buttons; right-click opens Copy,
   Paste, Export selection and Delete actions. The browser requests clipboard
   permission and the table status line reports the result.
10. **Multi-selection** - Ctrl/Cmd+click and Shift+click select several rows. The
    **Rows** menu can select all visible rows, clear, copy, or delete the
    selection; **Export selection** writes only the selected records. Gantt
    deletion can be undone with `Ctrl+Z`. Tables above 300 rows use windowed
    virtual scrolling.
11. **Task colors and group summaries** - each task row shows a color swatch
    next to its name. Click it to change the task color; the bar and legend
    update immediately. Collapsing a group shows one summary bar with the
    group span and duration-weighted progress; external dependencies remain
    linked to that summary bar. Collapsed groups also hide their table rows;
    the group-toggle strip above the table expands them again.

## Complete SUI source

The complete source is `app.sui` and is copied to `dist/app.sui`.

```sui
# Gantt calendar example.
#
# Demonstrates the chart-level calendar attributes:
# - holidays: exact dates, inclusive ranges, yearly recurrences and recurring ranges
# - weekend: weekday names/numbers, or "none"
# - timezone: IANA zone, visible in table dates, tooltips and axis ticks
#
# Use the buttons to switch each setting. The table under the chart is
# editable: change Task, Start, End, Actual start/end, Work, Work done,
# Progress, Group or Depends on and observe execution and remaining work
# update through onTaskChange.

state tasks = "[\{\"id\":\"kickoff\",\"task\":\"Kickoff\",\"start\":\"2026-11-02T00:30\",\"end\":\"2026-11-02T03:00\",\"progress\":100,\"group\":\"Plan\"\},\{\"id\":\"design\",\"task\":\"Design\",\"start\":\"2026-11-02T09:00\",\"end\":\"2026-11-05T17:00\",\"progress\":75,\"actualStart\":\"2026-11-03T09:00\",\"actualEnd\":\"2026-11-06T12:00\",\"work\":10,\"workDone\":7.5,\"group\":\"Build\",\"dependsOn\":[\"kickoff\"]\},\{\"id\":\"review\",\"task\":\"Review\",\"start\":\"2026-11-06T09:30\",\"end\":\"2026-11-06T12:00\",\"progress\":50,\"group\":\"Build\",\"dependsOn\":[\"design\"]\},\{\"id\":\"build\",\"task\":\"Implement\",\"start\":\"2026-11-09T08:00\",\"end\":\"2026-11-13T18:00\",\"progress\":40,\"group\":\"Build\",\"dependsOn\":[\"review\"]\},\{\"id\":\"holiday\",\"task\":\"Holiday check\",\"start\":\"2026-11-16T10:00\",\"end\":\"2026-11-16T12:00\",\"progress\":0,\"group\":\"Calendar\",\"dependsOn\":[\"build\"]\},\{\"id\":\"range\",\"task\":\"Range check\",\"start\":\"2026-11-19T09:00\",\"end\":\"2026-11-20T17:00\",\"progress\":0,\"group\":\"Calendar\",\"dependsOn\":[\"holiday\"]\},\{\"id\":\"release\",\"task\":\"Release\",\"start\":\"2026-11-27T16:00\",\"end\":\"2026-11-27T16:00\",\"milestone\":true,\"group\":\"Release\",\"dependsOn\":[\"range\"]\},\{\"id\":\"freeze\",\"task\":\"Year-end freeze\",\"start\":\"2026-12-24T08:00\",\"end\":\"2027-01-06T18:00\",\"progress\":0,\"group\":\"Calendar\",\"dependsOn\":[\"release\"]\}]"
state tz = "Europe/Rome"
state weekendDays = "sat,sun"
state holidayList = "*-11-16, 2026-11-19..2026-11-20"
state lastChange = "No task edits yet."

css ".calendar-page \{ max-width: 1280px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 14px; \} .calendar-controls \{ display: flex; flex-wrap: wrap; gap: 8px; align-items: center; \} .calendar-label \{ min-width: 86px; color: #8fa0b8; font-size: 13px; \} .calendar-chart \{ width: 100%; height: 560px; \} .calendar-status \{ color: #a5f3fc; font-size: 13px; \}"

fn setUtc() -> null {
	tz = "UTC"
	lastChange = "Timezone changed to UTC."
}

fn setRome() -> null {
	tz = "Europe/Rome"
	lastChange = "Timezone changed to Europe/Rome."
}

fn setNewYork() -> null {
	tz = "America/New_York"
	lastChange = "Timezone changed to America/New_York."
}

fn setSatSun() -> null {
	weekendDays = "sat,sun"
	lastChange = "Weekend changed to Saturday and Sunday."
}

fn setFriSat() -> null {
	weekendDays = "fri,sat"
	lastChange = "Weekend changed to Friday and Saturday."
}

fn setNoWeekend() -> null {
	weekendDays = "none"
	lastChange = "Weekend shading disabled."
}

fn setMixedHolidays() -> null {
	holidayList = "*-11-16, 2026-11-19..2026-11-20"
	lastChange = "Holidays: recurring 16 November plus the range 19-20 November."
}

fn setExactHoliday() -> null {
	holidayList = "2026-11-16"
	lastChange = "Holidays: exact date 16 November 2026."
}

fn setRangeHoliday() -> null {
	holidayList = "2026-11-19..2026-11-20"
	lastChange = "Holidays: inclusive range 19-20 November 2026."
}

fn setRecurringHoliday() -> null {
	holidayList = "*-11-16"
	lastChange = "Holidays: yearly recurrence 16 November."
}

fn setRecurringRange() -> null {
	holidayList = "*-12-24..*-01-06"
	lastChange = "Holidays: yearly recurring range 24 December-6 January, wrapping across the year boundary."
}

fn setNoHolidays() -> null {
	holidayList = ""
	lastChange = "Holiday shading cleared."
}

fn tasksChanged(next: string) -> null {
	tasks = next
	lastChange = "onTaskChange received the edited task list."
}

view main {
	title "Gantt calendar controls"
	row class: "calendar-page" {
		h1 "Gantt calendar controls"
		p "Switch holidays, weekend and timezone to verify each chart attribute. The table below the chart is accessible, editable and synchronized with the timeline." style: "margin:0;color:#8fa0b8"

		card {
			row class: "calendar-controls" {
				span "Timezone" class: "calendar-label"
				button "UTC" onClick: setUtc
				button "Europe/Rome" onClick: setRome
				button "America/New_York" onClick: setNewYork
			}
			row class: "calendar-controls" {
				span "Weekend" class: "calendar-label"
				button "Sat/Sun" onClick: setSatSun
				button "Fri/Sat" onClick: setFriSat
				button "None" onClick: setNoWeekend
			}
			row class: "calendar-controls" {
				span "Holidays" class: "calendar-label"
				button "Mixed" onClick: setMixedHolidays
				button "Exact" onClick: setExactHoliday
				button "Range" onClick: setRangeHoliday
				button "Recurring" onClick: setRecurringHoliday
				button "Recurring range" onClick: setRecurringRange
				button "None" onClick: setNoHolidays
			}
			p "Timezone: {tz} · Weekend: {weekendDays} · Holidays: {holidayList}" class: "calendar-status"
			p "{lastChange}" class: "calendar-status"
		}

		chart "Calendar controls" name: "calendar" class: "calendar-chart" data: tasks chartType: "gantt" tableView: "both" tableFilters: true tableAddRow: true tableDeleteRow: true tableSteppers: true tableContextMenu: true tableSelection: "row" criticalPath: true holidays: holidayList weekend: weekendDays timezone: tz showGrid: true showValues: true interactive: true onTaskChange: tasksChanged style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #a855f7; --sui-chart-gantt-weekend: #64748b; --sui-chart-gantt-holiday: #ef4444; --sui-chart-gantt-today: #f59e0b; --sui-chart-gantt-dependency: #94a3b8; --sui-chart-critical: #f59e0b; --sui-chart-grid: #334155; --sui-chart-axis: #475569;")

		card {
			p "How to verify:" style: "margin:0 0 6px;font-weight:600;color:#e2e8f0"
			p "1. Click UTC, Europe/Rome and America/New_York: Start/End values in the table and tooltips change with the selected zone. The kickoff crosses the date boundary in New York."
			p "2. Click Sat/Sun, Fri/Sat and None: the shaded weekend bands move or disappear."
			p "3. Click Exact, Range, Recurring, Recurring range and Mixed: the red holiday bands update, including recurring ranges that wrap across a year boundary in other datasets."
			p "4. Edit a Gantt table cell or drag a task bar: the chart and table stay synchronized and the status line confirms onTaskChange."
			p "5. Press + Row: a new task is appended, selected and focused; the chart, table and onTaskChange payload update immediately. Press Ctrl+Z to undo."
			p "6. Ctrl/Cmd+click or Shift+click selects multiple rows. Use the Rows menu to select all visible, clear, copy, paste or delete the selection, and Export selection to export only those records. Right-click the table for the same copy/paste/export/delete actions; the browser asks for clipboard permission."
			p "7. Numeric cells show visible minus/plus buttons, and each row ends with a delete button. Both behaviors are enabled by the table options in the source."
		}
	}
}
```

## Attribute reference

| Attribute | Values | Behavior |
|---|---|---|
| `holidays` | comma/space/semicolon list | Exact dates, inclusive `start..end` ranges, yearly `*-MM-DD` / `MM-DD` recurrences and recurring ranges. |
| `weekend` | `sat,sun`, `fri,sat`, `0,6`, `none` | Selects shaded non-working weekdays. `0` is Sunday and `6` is Saturday. |
| `timezone` | IANA zone, default `UTC` | Formats axis ticks, table dates and tooltips for date-time tasks. Date-only tasks stay calendar-stable. |
| `showNonWorkingDays` | `true` / `false`, default `true` | Shows or hides weekend/holiday bands without disabling non-working date validation. |
| `criticalPath` | `true` / `false` | Highlights zero-slack tasks and dependency links. |
| `tableView` | `chart`, `table`, `both` | Shows the accessible synchronized table. |
| `tableAddRow` | `true` / `false` | Shows **+ Row**. On Gantt it appends a task; on other chart types it appends a zero-valued local row and can call `onDataChange`. |
| `tableDeleteRow` | `true` / `false` | Adds a delete button at the end of every row. |
| `tableSteppers` | `true` / `false` | Replaces native number spinners with visible minus/plus buttons. |
| `tableContextMenu` | `true` / `false` | Enables right-click Copy, Export selection, Delete and selection actions. |
| `tableSelection` | `none`, `row`, `column`, `cell` | Chooses the table selection unit. |
| `interactive` | `true` / `false` | Enables Gantt drag/resize and table editing. |

The holiday matcher runs in the browser on each redraw. Invalid holiday tokens
are ignored; an invalid timezone falls back to UTC.

## Designer

Open the example in the designer and select the chart. In **Chart settings →
Interactions** you can edit Critical path, Holidays, Weekend and Timezone. The
same values are written back as chart attributes.

## Notes and limits

- Holidays are recurring yearly, not monthly or weekly.
- Weekend and holiday calendars are global for the chart, not per resource.
- Date-only values are not shifted by the timezone; date-time values are.
- Graphical dependency creation by dragging between tasks is not implemented;
  edit the **Depends on** table cell instead.
- The Gantt/table data is inert: cells use `textContent`, SVG text uses
  `textContent`, and no HTML is interpolated.















