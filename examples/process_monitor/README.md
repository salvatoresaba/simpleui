# Process monitor example

A runtime state/shape chart used as a process monitor. The chart starts in
**static** mode and only **queues events**; SUI drains the queue with
`chart.pollEvents` every 120 ms and calls its own functions. The queue item
identifies the event type and the element that produced it. Clicking the Process
shape updates the **Output** shape and opens a modal dialog. Buttons update tag
text/colors through the core `chart` module; the switch returns the graph to
editing mode. Every shape has a mandatory unique `id`.

## What it demonstrates

- Mandatory unique shape ids; runtime calls use the explicit `id` only.
- The chart as an event source: no per-element function binding.

  ```sui
  events, err = chart.pollEvents("monitor")
  ```

- `pumpChartEvents` maps queued events to external functions
  (`processClicked`, `shapeClicked`, `tagClicked`).
- `chart.peekEvents`, `chart.nextEvent`, `chart.clearEvents`,
  `chart.eventCount`.
- Reading and writing element attributes with `chart.getAttr`,
  `chart.setAttr` and `chart.update`.
- Calling `chart.update` from outside the chart to change another shape's tag
  and color in place.
- `chart.update` / `chart.setTag` updating shape and link tags at runtime
  without rebuilding the whole graph.
- Editing/static mode switched at runtime with `stateMode`.
- In static mode the diagram does not refit or shift when a modal opens/closes;
  element positions stay fixed.
- Custom `hint` values on shapes and links replace the default tooltip; the
  **Hint** column and the context menu **Hint settings...** edit them.
- Tag and hint fields are single-line: `|` starts a new line, `\|` is a literal
  pipe (real newlines in old data are normalised to `|`).
- Chart settings → **HINTS** styles every tooltip globally: background color and
  opacity (including fully transparent), text/border color, font family and
  size, border width, text padding and corner radius, with a live preview.
- `tableView: "both"`: selecting a shape focuses and highlights its table row;
  the event pump also calls `chart.focusRow`, so static clicks highlight the row.
- A wider chart viewport with buttons that update colors/text in place.
- A shape click opening a modal dialog.

## Build and run

```bash
simpleui check examples/process_monitor/app.sui --json
simpleui build examples/process_monitor/app.sui --out examples/process_monitor/dist
simpleui serve examples/process_monitor/app.sui --watch
```

## Complete source

```sui
import { json } from "std"

state data = "\{\"states\":[\{\"id\":\"intake\",\"label\":\"Intake\",\"shape\":\"stadium\",\"w\":150,\"h\":64,\"tag\":\"WAITING\",\"tagBox\":\"rounded\",\"tagBoxColor\":\"#e2e8f0\",\"tagTextColor\":\"#0f172a\",\"hint\":\"Waiting for a new item\"\},\{\"id\":\"process\",\"label\":\"Process\",\"shape\":\"rectangle\",\"w\":160,\"h\":76,\"tag\":\"IDLE\",\"tagBox\":\"rounded\",\"tagBoxColor\":\"#fef3c7\",\"tagTextColor\":\"#78350f\",\"hint\":\"Processing the order; click|to mark the output.\"\},\{\"id\":\"output\",\"label\":\"Output\",\"shape\":\"circle\",\"w\":96,\"h\":96,\"tag\":\"OUTPUT\",\"tagBox\":\"rounded\",\"tagBoxColor\":\"#dcfce7\",\"tagTextColor\":\"#14532d\",\"hint\":\"Completed item\"\}],\"transitions\":[\{\"id\":\"edge_1\",\"from\":\"intake\",\"to\":\"process\",\"tag\":\"start\",\"hint\":\"Start processing\"\},\{\"id\":\"edge_2\",\"from\":\"process\",\"to\":\"output\",\"tag\":\"done\",\"hint\":\"Processing finished\"\}]\}"

state staticMode = true
state chartMode = "static"
state status = "Static mode: click a shape or tag. Events are read from the chart event list."
state selectedShape = ""
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
		let eventName = str(ev["event"])
		let eventPart = str(ev["part"])
		let eventKind = str(ev["kind"])
		if eventName == "click" {
			okFocus, focusErr = chart.focusRow("monitor", str(ev["id"]))
			if eventPart == "tag" {
				tagClicked(ev)
			} else if eventKind == "shape" {
				if str(ev["id"]) == "process" {
					processClicked(ev)
				} else {
					shapeClicked(ev)
				}
			}
		} else if eventName == "mouseover" and eventKind == "shape" {
			status = "Hover " + str(ev["id"])
		}
		i = i + 1
	}
}

fn setWaiting() -> null {
	ok1, e1 = chart.update("monitor", "intake", "\{\"text\":\"WAITING\",\"tagColor\":\"#e2e8f0\",\"textColor\":\"#0f172a\"\}")
	ok2, e2 = chart.update("monitor", "process", "\{\"text\":\"IDLE\",\"tagColor\":\"#fef3c7\",\"textColor\":\"#78350f\",\"color\":\"#f8fafc\"\}")
	ok3, e3 = chart.update("monitor", "output", "\{\"text\":\"OUTPUT\",\"tagColor\":\"#dcfce7\",\"textColor\":\"#14532d\"\}")
}

fn setRunning() -> null {
	ok1, e1 = chart.update("monitor", "intake", "\{\"text\":\"READY\",\"tagColor\":\"#dbeafe\",\"textColor\":\"#1e3a8a\"\}")
	ok2, e2 = chart.update("monitor", "process", "\{\"text\":\"RUNNING\",\"tagColor\":\"#f59e0b\",\"textColor\":\"#111827\",\"color\":\"#fffbeb\"\}")
	ok3, e3 = chart.update("monitor", "output", "\{\"text\":\"WAITING\",\"tagColor\":\"#e5e7eb\",\"textColor\":\"#374151\"\}")
}

fn setDone() -> null {
	ok1, e1 = chart.update("monitor", "intake", "\{\"text\":\"DONE\",\"tagColor\":\"#d1fae5\",\"textColor\":\"#065f46\"\}")
	ok2, e2 = chart.update("monitor", "process", "\{\"text\":\"DONE\",\"tagColor\":\"#d1fae5\",\"textColor\":\"#065f46\"\}")
	ok3, e3 = chart.update("monitor", "output", "\{\"text\":\"COMPLETE\",\"tagColor\":\"#22c55e\",\"textColor\":\"#ffffff\",\"color\":\"#dcfce7\"\}")
}

fn toggleMode() -> null {
	staticMode = staticMode == true ? false : true
	chartMode = staticMode ? "static" : "editing"
	status = staticMode ? "Static mode: click shapes or tags." : "Editing mode."
}

fn shapeClicked(ev: any) -> null {
	selectedShape = str(ev["label"])
	selectedTag = str(ev["tag"])
	status = "Shape " + str(ev["id"]) + " · " + str(ev["event"])
	modal.open("process-details")
}

fn processClicked(ev: any) -> null {
	selectedShape = str(ev["label"])
	selectedTag = str(ev["tag"])

	ok1, e1 = chart.setText("monitor", "output", "PROCESS CLICKED")
	ok2, e2 = chart.setTagColor("monitor", "output", "#22c55e")
	ok3, e3 = chart.setTextColor("monitor", "output", "#ffffff")
	ok4, e4 = chart.setColor("monitor", "output", "#dcfce7")

	if e1 != null {
		status = "Update failed: " + e1
	} else if e2 != null {
		status = "Update failed: " + e2
	} else if e3 != null {
		status = "Update failed: " + e3
	} else if e4 != null {
		status = "Update failed: " + e4
	} else {
		status = "Event list: process click updated the output shape"
	}

	modal.open("process-details")
}

fn tagClicked(ev: any) -> null {
	selectedShape = str(ev["label"])
	selectedTag = str(ev["tag"])
	status = "Tag " + str(ev["tag"]) + " on " + str(ev["id"]) + " · " + str(ev["event"])
	modal.open("process-details")
}

view main {
	title "Process monitor"
	onLoad: startEvents
	col class: "monitor" {
		header {
			h1 "Process monitor"
			p "The chart only queues events (type + element). SUI reads them with chart.pollEvents and calls its own functions; the Process handler updates another shape."
		}
		row {
			button "Waiting" onClick: setWaiting
			button "Running" onClick: setRunning
			button "Done" onClick: setDone
			switch checked: staticMode onChange: toggleMode
			text "Static mode"
		}
		p "Status: {status}"
		chart "Process monitor" name: "monitor" data: data chartType: "state" width: "100%" height: "620px" stateMode: chartMode stateGrid: true stateGridGap: 24 stateGridColor: "#cbd5e1" stateShapeTagStyle: "rounded" stateShapeTagKind: "external" stateShapeTagLink: "line" stateLinkTagStyle: "rounded" stateLinkTagKind: "external" stateLinkTagLink: "line" tableView: "both" interactive: true theme: "dark" style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #22c55e; --sui-chart-series-3: #f59e0b; --sui-chart-bg: #0f172a;")
	}

	modal "Process details" name: "process-details" {
		p "Shape: {selectedShape}"
		p "Tag: {selectedTag}"
		p "Opened from the chart event list by pumpChartEvents → processClicked."
		modalbutton "Close"
	}
}

css "body \{ margin: 0; background: #020617; color: #e2e8f0; font-family: system-ui, sans-serif; \} .monitor \{ max-width: 1400px; width: 100%; box-sizing: border-box; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 16px; \} .monitor h1 \{ margin: 0 0 6px; \} .monitor p \{ margin: 0; color: #94a3b8; \}"
```

## How the event list works

Shape ids are mandatory and unique inside the chart. A duplicate or missing
`id` stops parsing with an explicit error. Links get an explicit `id` if
provided; otherwise they receive `edge_1`, `edge_2`, ... and the runtime verifies
that the result does not collide with a shape id.

The chart does not call application functions. It appends every static
interaction to a queue:

```json
{
  "event": "click",
  "kind": "shape",
  "part": "body",
  "target": "shape",
  "id": "process",
  "label": "Process",
  "tag": "IDLE",
  "clientX": 420,
  "clientY": 260,
  "button": 0,
  "time": 1730000000000
}
```

SUI drains the queue and owns the association between events and external
functions:

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
			} else if str(ev["kind"]) == "shape" {
				shapeClicked(ev)
			}
		} else if str(ev["event"]) == "click" and str(ev["part"]) == "tag" {
			tagClicked(ev)
		}
		i = i + 1
	}
}
```

Queue functions:

| Function | Result |
|---|---|
| `chart.pollEvents(name)` | all pending events as JSON and empties the queue |
| `chart.peekEvents(name)` | pending events without emptying |
| `chart.nextEvent(name)` | oldest event and removes it |
| `chart.clearEvents(name)` | empties the queue |
| `chart.eventCount(name)` | number of pending events |

The queue keeps at most the 200 most recent events. Poll regularly.

Legacy direct callbacks (`chart.on`, data `onClick` fields, chart-level
`onShapeClick`) are still supported, but they call SUI functions from inside the
chart runtime. The event list is the recommended model: the chart only produces
data and never touches SUI state.

## Reading and writing elements

The chart exposes synchronous functions; check the error tuple every time.

```sui
// Search: all elements, or by id/label/tag/kind/etc.
all, err = chart.find("monitor", "")
running, err2 = chart.find("monitor", "tag:RUNNING")
shapes, err3 = chart.find("monitor", "kind:shape")

// Read all attributes of one element (JSON) or one attribute (string).
info, err4 = chart.get("monitor", "process")
tag, err5 = chart.getAttr("monitor", "process", "tag")

// Write one attribute or several attributes.
ok, err6 = chart.setAttr("monitor", "process", "text", "RUNNING")
ok2, err7 = chart.setAttr("monitor", "process", "tagColor", "#f59e0b")
ok3, err8 = chart.update("monitor", "output", "\{\"text\":\"DONE\",\"tagColor\":\"#22c55e\"\}")

// Read the event queue.
events, err9 = chart.pollEvents("monitor")
peek, err10 = chart.peekEvents("monitor")
one, err11 = chart.nextEvent("monitor")
count, err12 = chart.eventCount("monitor")
ok4, err13 = chart.clearEvents("monitor")
```

Accepted element references: only the explicit `id`, for example `process` or
`edge_1`. Prefixes, labels and `from->to` are rejected. `chart.find` accepts an
optional `field:value` query (`id:`, `kind:`, `label:`, `tag:`, `hint:`,
`shape:`, `from:`, `to:`, `routing:`); without a colon it searches every field.

`chart.setAttr` and `chart.update` accept the same patch fields. `chart.get`
returns the full current element object, including `id`, `label`, `tag`,
geometry and colors. The typed helpers (`chart.setText`, `chart.setTagColor`,
`chart.setHint`, `chart.setColor`, ...) are preferred for single attributes. The
event queue is the recommended way to react to chart interactions; legacy
`chart.on` bindings remain available.









