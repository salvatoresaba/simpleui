# Gantt WBS and auto-scheduling example

This runnable example demonstrates the work-breakdown and scheduling features of
the SimpleUI Gantt:

- `parent` creates a hierarchical WBS.
- A task with children becomes a summary task: its start/end span its children,
  progress is duration-weighted and cost is summed.
- An expanded summary is a header row without a bar. Collapsing it hides the
  descendant rows in both chart and table and shows one summary bar.
- Dependencies crossing a collapsed summary stay attached to the summary bar;
  dependencies entirely inside the summary are hidden.
- `dependsOn` accepts task names as well as ids. The table shows each dependency
  as a removable chip and a combo with all available task names, so a task can
  depend on several predecessors. Click a chip to edit that link's type and lag
  independently.
- The table rejects a dependency that would create a cycle; imported cyclic data
  produces an explicit chart error.
- Dependency arrows are neutral graphite; critical links are thicker but never
  yellow. Link color, width, line style and arrow type/size are configurable in
  **Chart settings → Interactions**, and the source sets them explicitly.
  Dense rows automatically gain vertical space with a scrollable Gantt so links
  and bars stay readable.
- `autoSchedule: true` recalculates starts from dependency type, lag and
  constraint/date values. The example also exposes a button to switch it off.
- The table is synchronized and editable, including the `Parent` combo box.
- Child rows are indented to the right in both chart and table; summary/group
  names are emphasized with a header band and bold text.
- `Duration` is the calendar difference; **Effective duration** counts working
  days using the chart weekend and holidays, and can drive effort or cost.
- Open the chart menu and choose **Calendar...** to edit weekends, holidays and
  closure ranges. Changes are stored per chart and update effective duration
  immediately.

## Run it

A desktop build is included in `dist/`:

- **Windows:** double-click `examples/gantt_wbs/dist/app.exe`.

Rebuild and launch from this folder:

```powershell
./run.ps1
```

or:

```bat
run.bat
```

For a web build:

```bash
simpleui build examples/gantt_wbs/app.sui --out examples/gantt_wbs/dist
simpleui serve examples/gantt_wbs/app.sui --watch
```

## What to verify

1. **WBS outline** - Discovery, Build and Validation are summary tasks; their
   children are indented in the chart and table.
2. **Collapse** - click the triangle in a summary label. The summary bar remains
   with rolled-up progress and dates.
3. **Dependencies** - collapse Build and confirm that the dependency to
   Integration is still drawn to the Build summary bar.
4. **Scheduling** - turn auto scheduling off, edit a `Link` or `Lag` cell, then
   turn it on again. The Start/End cells change only while auto scheduling is on.
5. **Table synchronization** - edit `Parent` or `Depends on`; the timeline and
   WBS order update. A parent cycle is rejected.
6. **Summary values** - the summary row shows aggregated dates, progress and
   cost; the table is still editable for leaf tasks.

## Complete SUI source

The complete source is `app.sui` and is copied to `dist/app.sui`.

```sui
# Gantt WBS + auto-scheduling example.
#
# Demonstrates:
# - parent/child tasks rendered as a collapsible WBS outline;
# - summary bars with rolled-up start, end, progress and cost;
# - dependency-driven auto scheduling (type, lag and constraints);
# - external dependencies preserved when a summary is collapsed.
#
# The table is the editable source of truth: change Parent, Start, End,
# Depends on or other fields and the chart updates through onTaskChange.

state tasks = "[\{\"id\":\"discovery\",\"task\":\"Discovery\",\"start\":\"2027-03-01\",\"end\":\"2027-03-12\",\"progress\":0,\"group\":\"Plan\",\"parent\":\"\"\},\{\"id\":\"interviews\",\"task\":\"User interviews\",\"start\":\"2027-03-01\",\"end\":\"2027-03-04\",\"progress\":100,\"group\":\"Plan\",\"parent\":\"discovery\"\},\{\"id\":\"requirements\",\"task\":\"Requirements\",\"start\":\"2027-03-05\",\"end\":\"2027-03-12\",\"progress\":80,\"group\":\"Plan\",\"parent\":\"discovery\",\"dependsOn\":[\"User interviews\"]\},\{\"id\":\"build\",\"task\":\"Build\",\"start\":\"2027-03-15\",\"end\":\"2027-04-09\",\"progress\":0,\"group\":\"Delivery\",\"parent\":\"\"\},\{\"id\":\"api\",\"task\":\"API\",\"start\":\"2027-03-15\",\"end\":\"2027-03-26\",\"progress\":50,\"group\":\"Delivery\",\"parent\":\"build\",\"dependsOn\":[\"Requirements\"]\},\{\"id\":\"ui\",\"task\":\"User interface\",\"start\":\"2027-03-22\",\"end\":\"2027-04-02\",\"progress\":30,\"group\":\"Delivery\",\"parent\":\"build\",\"dependsOn\":[\"API\"]\},\{\"id\":\"integration\",\"task\":\"Integration\",\"start\":\"2027-04-05\",\"end\":\"2027-04-09\",\"progress\":10,\"group\":\"Delivery\",\"parent\":\"build\",\"dependsOn\":[\"User interface\"]\},\{\"id\":\"qa\",\"task\":\"Validation\",\"start\":\"2027-04-12\",\"end\":\"2027-04-23\",\"progress\":0,\"group\":\"Quality\",\"parent\":\"\"\},\{\"id\":\"testing\",\"task\":\"System testing\",\"start\":\"2027-04-12\",\"end\":\"2027-04-20\",\"progress\":0,\"group\":\"Quality\",\"parent\":\"qa\",\"dependsOn\":[\"Integration\"]\},\{\"id\":\"signoff\",\"task\":\"Sign-off\",\"start\":\"2027-04-21\",\"end\":\"2027-04-23\",\"progress\":0,\"group\":\"Quality\",\"parent\":\"qa\",\"dependsOn\":[\"System testing\"]\},\{\"id\":\"launch\",\"task\":\"Launch\",\"start\":\"2027-04-26\",\"end\":\"2027-04-26\",\"milestone\":true,\"group\":\"Release\",\"parent\":\"\",\"dependsOn\":[\"Sign-off\"]\},\{\"id\":\"training\",\"task\":\"Training\",\"start\":\"2027-04-27\",\"end\":\"2027-04-29\",\"progress\":0,\"group\":\"Release\",\"parent\":\"\",\"dependsOn\":[\"System testing\"]\}]"
state autoMode = true
state message = "Auto scheduling is on. Collapse Discovery, Build or Validation to see the summary bars."

fn tasksChanged(next: string) -> null {
	tasks = next
	message = "onTaskChange received the edited WBS data."
}

fn toggleAuto() -> null {
	autoMode = not autoMode
	if autoMode {
		message = "Auto scheduling is on."
	} else {
		message = "Auto scheduling is off; dates are still rolled up for summary rows."
	}
}

view main {
	title "Gantt WBS and auto-scheduling"
	row class: "wbs-page" {
		h1 "Gantt WBS and auto-scheduling"
		p "A parent/child work breakdown with summary roll-up, dependency scheduling and a synchronized editable table." style: "margin:0;color:#8fa0b8"
		card {
			row class: "wbs-controls" {
				button "Toggle auto scheduling" onClick: toggleAuto
				span "Auto schedule: {autoMode}" class: "wbs-status"
			}
			p "{message}" class: "wbs-status"
		}
		chart "Project WBS" name: "wbs" class: "wbs-chart" data: tasks chartType: "gantt" tableView: "both" tableFilters: true tableSelection: "row" tableContextMenu: true tableAddRow: true tableDeleteRow: true tableSteppers: true criticalPath: true autoSchedule: autoMode linkColor: "#94a3b8" linkWidth: 1 linkStyle: "solid" arrowType: "triangle" arrowSize: 5 showProgress: true showValues: true showResources: true showCosts: true interactive: true onTaskChange: tasksChanged style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #a855f7; --sui-chart-series-3: #f59e0b; --sui-chart-series-4: #22c55e; --sui-chart-grid: #334155; --sui-chart-axis: #475569; --sui-chart-gantt-dependency: #94a3b8; --sui-chart-critical: #f59e0b;") height: "620px"
		card {
			p "How to verify:" style: "margin:0 0 6px;font-weight:600;color:#e2e8f0"
			p "1. Click the triangle before Discovery, Build or Validation. The children disappear and the summary bar remains with rolled-up progress."
			p "2. Collapse a summary that has dependencies crossing its boundary: the arrow stays attached to the summary bar instead of disappearing."
			p "3. Edit a Parent cell to move a task into another summary. Use the Depends on chips and combo to add/remove predecessors; a cycle is rejected."
			p "4. Toggle auto scheduling and change a link type or lag: with auto scheduling on, Start/End are recalculated; with it off, the dates remain as entered."
			p "5. Summary rows are indented in the table and show the aggregated start, end, progress and cost."
		}
	}
}

css ".wbs-page \{ max-width: 1320px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 14px; \} .wbs-controls \{ display: flex; gap: 10px; align-items: center; flex-wrap: wrap; \} .wbs-chart \{ width: 100%; \} .wbs-status \{ color: #a5f3fc; font-size: 13px; \}"




```
