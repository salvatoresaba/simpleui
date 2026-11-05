# Chart example

This example is a chart gallery split into eight pages so each family can be
inspected at a comfortable size. Shared state is available on every page, and
the navigation bar moves between **Overview**, **Data**, **Bars**, **Lines**,
**Distributions**, **Parts**, **Relationships** and **Time**.

The **Data** page switches one chart between **JSON**, **XML** and **CSV** and
shows long-series scroll and zoom modes. Other pages group related chart types:
variants of bars together, line styles together, histograms with different bin
counts together, and part-to-whole charts together.

## What it demonstrates

- `chart` with `data` bound to a `str` state (`data: chartData`).
- Multi-page navigation with `navigate("page")` and shared reactive state.
- `format: "auto"` detection across JSON, CSV and XML, plus explicit
  `format: "csv"` for CSV-only series.
- `chartType`: bar, horizontal bar, stacked/grouped bar, combo, line, step,
  smooth, area, pie, donut, scatter, histogram, waterfall, radar, gauge,
  funnel, heatmap, candlestick, treemap, box plot, bullet, dumbbell and Gantt.
- Multi-series JSON/CSV/XML data plus the data/type compatibility check.
- Gantt dates, progress bars, milestones, groups, dependencies, weekend bands,
  the today marker and interactive drag/resize with `onTaskChange`.
- Interactive tooltips, hover highlights and click selection on every chart in
  the gallery; all 24 chart types set `interactive: true`.
- Tooltips on bars, line/area/step/smooth points, pie/donut slices, radar
  points, the gauge arc, funnel stages, heatmap cells and all other data marks.
- A per-chart tooltip mini panel in the Chart inspector for colors, radius,
  maximum width, font size and hover delay; it writes `--sui-chart-tooltip-*`
  variables into the chart's `style expr`.
- Gantt group header rows with task counts and triangles; collapsing a group
  removes its task rows and compacts the timeline.
- Critical-path highlighting and configurable holiday bands in the Gantt.
- Exact, range and yearly recurring holidays, configurable weekends and an IANA
  timezone for Gantt dates.
- Visible Gantt resize handles on the left/right edge of each task bar.
- Accessible data tables for every chart type through `tableView: "table"` or
  `"both"`. Gantt tables include Task, Start, End, Progress, Group and Depends
  on; other pages show Label plus series or type-specific columns.
- Optional Excel-style per-column filter popups (`tableFilters: false` disables
  them) with contains search, value checklists and numeric/date min-max ranges;
  click-to-sort, Shift+click multi-sort, draggable column reordering, sticky
  identifying column, keyboard row/header navigation, and a Settings gear with
  Columns, Filters and Export submenus (CSV/TSV/JSON/JSONL, copy visible rows,
  print, export selection). Ctrl/Cmd+click and Shift+click select multiple
  rows, columns or cells according to `tableSelection`; the Rows menu can select
  all visible, clear, copy, paste or delete them. Optional delete-row buttons,
  visible numeric steppers and a right-click context menu are enabled on the
  examples; clipboard permission is requested by the browser.
  Tables above 300 rows use windowed virtual scrolling. Filter, sort and column
  layout persist in `localStorage`; menus close when clicking outside.
- Chart context menu (right-click or **⋮**) with SVG, PNG, CSV, TSV, JSON and
  JSONL export.
- Editable table cells: Gantt Task/Start/End/Progress/Group/Depends on and
  non-Gantt Label/X/series cells when `interactive` is on; chart edits update the
  table and table edits update the chart through `onTaskChange` or
  `onDataChange`. With `tableAddRow: true`, **+ Row** appends a new
  task, clears active table filters, selects/focuses the row and updates the
  timeline. The same option is available on every chart table; non-Gantt charts
  receive a zero-valued local row and can notify `onDataChange`.
- Bidirectional selection between every chart and its table: a mark selects the
  row, a row highlights all marks for that label (including every series), plus a
  short flash on edited Gantt cells and task. Selection, filters, sorting and
  column layout persist in `localStorage`.
- Undo/redo for Gantt edits with `Ctrl+Z`, `Ctrl+Shift+Z` / `Ctrl+Y` and chart-menu
  actions.
- A unified **Chart settings** panel with Appearance, Points, Tooltip, Table,
  Interactions and Data tabs.
- Keyboard navigation over chart marks with arrow keys, Home/End, Enter/Space and
  Escape.
- `x`/`y` field selection for object and CSV data.
- Fixed frame dimensions through CSS size classes (the same rules the standard
  **Size** fields edit), plus `overflow: "scroll"`, `overflow: "zoom"` and
  `overflow: "fit"`.
- `pointWidth`, configurable `pointSize`, grid, legend and value labels. The
  Line, Scatter, Combo, Radar and Dumbbell cards use different point sizes.
- Style presets written as `--sui-chart-*` variables in `style expr`.
- Reactive data replacement without rebuilding the chart markup by hand.

Charts that need multiple series (grouped/stacked bars, combo, radar and
dumbbell) read the shared `multiData`; single-series charts reject multi-series
payloads through the same compatibility check used by the designer overlay.

## Complete source

```sui
# Chart gallery.
#
# The charts are split across pages so each family can be inspected at a
# comfortable size. Shared state stays available on every page.
#
# Navigation: Overview, Data, Bars, Lines, Distributions, Parts, Relationships,
# Time.

state chartData = "[\{\"month\":\"Jan\",\"value\":12\},\{\"month\":\"Feb\",\"value\":18\},\{\"month\":\"Mar\",\"value\":15\},\{\"month\":\"Apr\",\"value\":24\},\{\"month\":\"May\",\"value\":21\},\{\"month\":\"Jun\",\"value\":29\},\{\"month\":\"Jul\",\"value\":26\},\{\"month\":\"Aug\",\"value\":31\}]"
state formatName = "JSON"
state notice = "Switch the same chart between JSON, CSV and XML."
state longData = "Day,Value\n01,4\n02,7\n03,5\n04,9\n05,12\n06,8\n07,14\n08,11\n09,16\n10,13\n11,18\n12,15\n13,20\n14,17\n15,22\n16,19\n17,24\n18,21\n19,26\n20,23\n21,28\n22,25\n23,30\n24,27\n25,32\n26,29\n27,34\n28,31\n29,36\n30,33"
state multiData = "[\{\"month\":\"Jan\",\"sales\":12,\"cost\":5\},\{\"month\":\"Feb\",\"sales\":18,\"cost\":8\},\{\"month\":\"Mar\",\"sales\":15,\"cost\":7\},\{\"month\":\"Apr\",\"sales\":24,\"cost\":11\},\{\"month\":\"May\",\"sales\":21,\"cost\":9\},\{\"month\":\"Jun\",\"sales\":29,\"cost\":13\}]"
state histogramData = "Value\n3\n5\n5\n6\n7\n8\n8\n9\n12\n14\n15\n18"
state waterfallData = "Step,Delta\nStart,20\nSales,-5\nCosts,-3\nFinal,8"
state kpiData = "Value\n72"
state heatData = "Day,Hour,Value\nMon,9,3\nMon,10,5\nMon,11,8\nTue,9,4\nTue,10,7\nTue,11,6\nWed,9,2\nWed,10,9\nWed,11,5"
state ohlcData = "[\{\"date\":\"D1\",\"open\":10,\"high\":15,\"low\":8,\"close\":13\},\{\"date\":\"D2\",\"open\":13,\"high\":16,\"low\":11,\"close\":12\},\{\"date\":\"D3\",\"open\":12,\"high\":14,\"low\":9,\"close\":14\},\{\"date\":\"D4\",\"open\":14,\"high\":18,\"low\":13,\"close\":17\}]"
state boxData = "Name,Min,Q1,Median,Q3,Max\nA,2,4,6,8,12\nB,1,3,5,9,11\nC,3,5,7,10,14"
state bulletData = "Team,Value,Target,Max\nA,72,80,100\nB,58,70,100"
state dumbbellData = "Team,Before,After\nA,40,72\nB,55,68\nC,30,60"
state scatterData = "[\{\"x\":1,\"y\":4\},\{\"x\":2,\"y\":7\},\{\"x\":3,\"y\":5\},\{\"x\":4,\"y\":9\},\{\"x\":5,\"y\":12\},\{\"x\":6,\"y\":8\},\{\"x\":7,\"y\":14\},\{\"x\":8,\"y\":11\}]"
state ganttData = "[\{\"id\":\"design\",\"task\":\"Design\",\"start\":\"2026-11-02\",\"end\":\"2026-11-06\",\"progress\":80,\"group\":\"Progettazione\"\},\{\"id\":\"review\",\"task\":\"Design review\",\"start\":\"2026-11-06\",\"end\":\"2026-11-09\",\"progress\":60,\"group\":\"Progettazione\",\"dependsOn\":[\"design\"]\},\{\"id\":\"api\",\"task\":\"API\",\"start\":\"2026-11-09\",\"end\":\"2026-11-14\",\"progress\":55,\"group\":\"Sviluppo\",\"dependsOn\":[\"review\"]\},\{\"id\":\"ui\",\"task\":\"UI\",\"start\":\"2026-11-10\",\"end\":\"2026-11-17\",\"progress\":40,\"group\":\"Sviluppo\",\"dependsOn\":[\"review\"]\},\{\"id\":\"build\",\"task\":\"Build\",\"start\":\"2026-11-17\",\"end\":\"2026-11-20\",\"progress\":20,\"group\":\"Sviluppo\",\"dependsOn\":[\"api\",\"ui\"]\},\{\"id\":\"test\",\"task\":\"Test\",\"start\":\"2026-11-19\",\"end\":\"2026-11-24\",\"progress\":10,\"group\":\"Qualita\",\"dependsOn\":[\"build\"]\},\{\"id\":\"fix\",\"task\":\"Bug fix\",\"start\":\"2026-11-23\",\"end\":\"2026-11-26\",\"progress\":0,\"group\":\"Qualita\",\"dependsOn\":[\"test\"]\},\{\"id\":\"ship\",\"task\":\"Ship\",\"start\":\"2026-11-27\",\"end\":\"2026-11-27\",\"milestone\":true,\"group\":\"Rilascio\",\"dependsOn\":[\"fix\"]\}]"

css ".chart-page \{ display: flex; flex-direction: column; gap: 14px; width: 100%; max-width: 1500px; margin: 0 auto; padding: 24px; box-sizing: border-box; \} .chart-grid \{ display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start; \} .chart-nav \{ display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 4px; \} .chart-wide \{ width: 100%; height: 340px; \} .chart-tile \{ width: 360px; height: 260px; \} .chart-time \{ width: 100%; height: 460px; \}"

fn ganttChanged(tasks: string) -> null {
	ganttData = tasks
}

fn chartDataChanged(next: string) -> null {
	chartData = next
}

fn goOverview() -> null { navigate("main") }
fn goData() -> null { navigate("data") }
fn goBars() -> null { navigate("bars") }
fn goLines() -> null { navigate("lines") }
fn goDistributions() -> null { navigate("distributions") }
fn goParts() -> null { navigate("parts") }
fn goRelationships() -> null { navigate("relationships") }
fn goTime() -> null { navigate("time") }

fn useJson() -> null {
	chartData = "[\{\"month\":\"Jan\",\"value\":12\},\{\"month\":\"Feb\",\"value\":18\},\{\"month\":\"Mar\",\"value\":15\},\{\"month\":\"Apr\",\"value\":24\},\{\"month\":\"May\",\"value\":21\},\{\"month\":\"Jun\",\"value\":29\},\{\"month\":\"Jul\",\"value\":26\},\{\"month\":\"Aug\",\"value\":31\}]"
	formatName = "JSON"
	notice = "JSON array of objects; x and y select the fields."
}

fn useCsv() -> null {
	chartData = "Month,Value\nJan,12\nFeb,18\nMar,15\nApr,24\nMay,21\nJun,29\nJul,26\nAug,31"
	formatName = "CSV"
	notice = "CSV with a header row; auto-detection handles commas."
}

fn useXml() -> null {
	chartData = "<data><row label=\"Jan\" value=\"12\"/><row label=\"Feb\" value=\"18\"/><row label=\"Mar\" value=\"15\"/><row label=\"Apr\" value=\"24\"/><row label=\"May\" value=\"21\"/><row label=\"Jun\" value=\"29\"/><row label=\"Jul\" value=\"26\"/><row label=\"Aug\" value=\"31\"/></data>"
	formatName = "XML"
	notice = "XML rows use label/x and value/y attributes."
}

view main {
	title "Chart gallery"
	row class: "chart-page" {
		h1 "Chart gallery"
		p "Choose a page to inspect one chart family at a time. Every page has a navigation bar and a comfortable chart size." style: "margin:0;color:#8fa0b8"
		row class: "chart-nav" {
			button "Data" onClick: goData
			button "Bars" onClick: goBars
			button "Lines" onClick: goLines
			button "Distributions" onClick: goDistributions
			button "Parts" onClick: goParts
			button "Relationships" onClick: goRelationships
			button "Time" onClick: goTime
		}
		row class: "chart-grid" {
			chart "Monthly revenue" name: "revenue" class: "chart-wide" data: chartData chartType: "bar" format: "auto" x: "month" y: "value" overflow: "scroll" pointWidth: "58" showGrid: true showLegend: true showValues: true interactive: true style: cssSafe("--sui-chart-bg: #0b1f33; --sui-chart-fg: #ecfeff; --sui-chart-title-color: #a5f3fc; --sui-chart-muted: #7dd3fc; --sui-chart-grid: #164e63; --sui-chart-axis: #0e7490; --sui-chart-series-1: #22d3ee; --sui-chart-series-2: #38bdf8; --sui-chart-series-3: #2dd4bf; --sui-chart-series-4: #818cf8; --sui-chart-series-5: #67e8f9; --sui-chart-series-6: #a5b4fc; --sui-chart-series-7: #14b8a6; --sui-chart-series-8: #7dd3fc; --sui-chart-border: 1px solid #164e63; --sui-chart-radius: 10px; --sui-chart-padding: 8px;")
			chart "Revenue trend" class: "chart-tile" data: chartData chartType: "line" x: "month" y: "value" overflow: "zoom" showGrid: true showLegend: false showValues: true style: cssSafe("--sui-chart-series-1: #22c55e; --sui-chart-grid: #2c3444; --sui-chart-axis: #3a4152;") interactive: true
			chart "Share" class: "chart-tile" data: chartData chartType: "donut" overflow: "fit" showLegend: true style: cssSafe("--sui-chart-bg: #2a1033; --sui-chart-fg: #fdf2f8; --sui-chart-title-color: #f9a8d4; --sui-chart-muted: #f0abfc; --sui-chart-grid: #6b21a8; --sui-chart-axis: #86198f; --sui-chart-series-1: #ec4899; --sui-chart-series-2: #a855f7; --sui-chart-series-3: #38bdf8; --sui-chart-series-4: #f472b6; --sui-chart-series-5: #c084fc; --sui-chart-series-6: #22d3ee; --sui-chart-series-7: #fb7185; --sui-chart-series-8: #d946ef; --sui-chart-border: 1px solid #6b21a8; --sui-chart-radius: 16px; --sui-chart-padding: 10px;") interactive: true
		}
	}
}

view data {
	title "Chart data"
	row class: "chart-page" {
		h1 "Data and raw formats"
		row class: "chart-nav" {
			button "Overview" onClick: goOverview
			button "Bars" onClick: goBars
			button "Lines" onClick: goLines
			button "Distributions" onClick: goDistributions
			button "Parts" onClick: goParts
			button "Relationships" onClick: goRelationships
			button "Time" onClick: goTime
		}
		p "Data source: {formatName}. {notice}" style: "margin:0;color:#8fa0b8"
		row style: "gap:8px;flex-wrap:wrap" {
			button "JSON" onClick: useJson
			button "CSV" onClick: useCsv
			button "XML" onClick: useXml
		}
		row class: "chart-grid" {
			chart "Same data, JSON/CSV/XML" class: "chart-wide" data: chartData chartType: "bar" tableView: "both" tableAddRow: true tableDeleteRow: true tableSteppers: true tableContextMenu: true tableSelection: "cell" onDataChange: chartDataChanged x: "month" y: "value" overflow: "scroll" showGrid: true showLegend: true showValues: true interactive: true
			chart "Long series: scroll mode" class: "chart-tile" data: longData chartType: "area" format: "csv" x: "Day" y: "Value" overflow: "scroll" pointWidth: "34" showGrid: true showLegend: true interactive: true
			chart "Long series: zoom mode" class: "chart-tile" data: longData chartType: "line" format: "csv" x: "Day" y: "Value" overflow: "zoom" pointWidth: "34" showGrid: true showLegend: true interactive: true
		}
	}
}

view bars {
	title "Bar charts"
	row class: "chart-page" {
		h1 "Bars and comparisons"
		row class: "chart-nav" {
			button "Overview" onClick: goOverview
			button "Data" onClick: goData
			button "Lines" onClick: goLines
			button "Distributions" onClick: goDistributions
			button "Parts" onClick: goParts
			button "Relationships" onClick: goRelationships
			button "Time" onClick: goTime
		}
		p "The same category/value data shown as vertical, horizontal, grouped, stacked, combo and waterfall." style: "margin:0;color:#8fa0b8"
		row class: "chart-grid" {
			chart "Vertical bars" class: "chart-tile" data: chartData chartType: "bar" x: "month" y: "value" showGrid: true showLegend: true showValues: false interactive: true
			chart "Horizontal bars" class: "chart-tile" data: chartData chartType: "barhorizontal" x: "month" y: "value" showGrid: true showLegend: false showValues: true style: cssSafe("--sui-chart-series-1: #22c55e;") interactive: true
			chart "Grouped bars" class: "chart-tile" data: multiData chartType: "bargrouped" x: "month" showGrid: true showLegend: true style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #f59e0b;") interactive: true
			chart "Stacked bars" class: "chart-tile" data: multiData chartType: "barstacked" x: "month" showGrid: true showLegend: true style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #f59e0b;") interactive: true
			chart "Combo bar + line" class: "chart-tile" data: multiData chartType: "combo" x: "month" pointSize: "4" showGrid: true showLegend: true style: cssSafe("--sui-chart-series-1: #4c5cf7; --sui-chart-series-2: #f59e0b;") interactive: true
			chart "Waterfall" class: "chart-tile" data: waterfallData chartType: "waterfall" x: "Step" y: "Delta" showGrid: true showLegend: true showValues: true style: cssSafe("--sui-chart-series-2: #22c55e; --sui-chart-series-4: #ef4444;") interactive: true
		}
	}
}

view lines {
	title "Line charts"
	row class: "chart-page" {
		h1 "Lines and trends"
		row class: "chart-nav" {
			button "Overview" onClick: goOverview
			button "Data" onClick: goData
			button "Bars" onClick: goBars
			button "Distributions" onClick: goDistributions
			button "Parts" onClick: goParts
			button "Relationships" onClick: goRelationships
			button "Time" onClick: goTime
		}
		p "The same monthly series rendered with line, step, smooth, area, zoom and long-scroll modes." style: "margin:0;color:#8fa0b8"
		row class: "chart-grid" {
			chart "Line" class: "chart-tile" data: chartData chartType: "line" x: "month" y: "value" pointSize: "5" showGrid: true showLegend: true showValues: true style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-tooltip-bg: #0b1f33; --sui-chart-tooltip-fg: #ecfeff; --sui-chart-tooltip-border: #22d3ee; --sui-chart-tooltip-radius: 12px; --sui-chart-tooltip-max-width: 260px; --sui-chart-tooltip-font-size: 12px; --sui-chart-tooltip-delay: 120ms;") interactive: true
			chart "Step" class: "chart-tile" data: chartData chartType: "step" x: "month" y: "value" showGrid: true showLegend: false showValues: false style: cssSafe("--sui-chart-series-1: #a855f7;") interactive: true
			chart "Smooth" class: "chart-tile" data: chartData chartType: "smooth" x: "month" y: "value" showGrid: true showLegend: false showValues: false style: cssSafe("--sui-chart-series-1: #22d3ee;") interactive: true
			chart "Area" class: "chart-tile" data: chartData chartType: "area" x: "month" y: "value" showGrid: true showLegend: false style: cssSafe("--sui-chart-series-1: #22c55e;") interactive: true
			chart "Zoom mode" class: "chart-tile" data: chartData chartType: "line" x: "month" y: "value" overflow: "zoom" showGrid: true showLegend: false style: cssSafe("--sui-chart-series-1: #f59e0b;") interactive: true
			chart "Long series, scroll" class: "chart-tile" data: longData chartType: "area" format: "csv" x: "Day" y: "Value" overflow: "scroll" pointWidth: "30" showGrid: true showLegend: false interactive: true
		}
	}
}

view distributions {
	title "Distributions"
	row class: "chart-page" {
		h1 "Distributions and intensity"
		row class: "chart-nav" {
			button "Overview" onClick: goOverview
			button "Data" onClick: goData
			button "Bars" onClick: goBars
			button "Lines" onClick: goLines
			button "Parts" onClick: goParts
			button "Relationships" onClick: goRelationships
			button "Time" onClick: goTime
		}
		p "Histogram bin sizes, box plot, heatmap and treemap shown side by side." style: "margin:0;color:#8fa0b8"
		row class: "chart-grid" {
			chart "Histogram, 4 bins" class: "chart-tile" data: histogramData chartType: "histogram" bins: 4 showGrid: true showLegend: false showValues: true style: cssSafe("--sui-chart-series-1: #f97316;") interactive: true
			chart "Histogram, 12 bins" class: "chart-tile" data: histogramData chartType: "histogram" bins: 12 showGrid: true showLegend: false showValues: true style: cssSafe("--sui-chart-series-1: #ef4444;") interactive: true
			chart "Box plot" class: "chart-tile" data: boxData format: "csv" chartType: "boxplot" x: "Name" showGrid: true showLegend: false showValues: true style: cssSafe("--sui-chart-series-1: #38bdf8;") interactive: true
			chart "Heatmap" class: "chart-tile" data: heatData format: "csv" chartType: "heatmap" x: "Day" showLegend: false showValues: true style: cssSafe("--sui-chart-series-1: #06b6d4;") interactive: true
		}
	}
}

view parts {
	title "Parts of a whole"
	row class: "chart-page" {
		h1 "Parts, targets and stages"
		row class: "chart-nav" {
			button "Overview" onClick: goOverview
			button "Data" onClick: goData
			button "Bars" onClick: goBars
			button "Lines" onClick: goLines
			button "Distributions" onClick: goDistributions
			button "Relationships" onClick: goRelationships
			button "Time" onClick: goTime
		}
		p "Pie, donut, funnel, gauge, bullet and treemap share the same label/value data where possible." style: "margin:0;color:#8fa0b8"
		row class: "chart-grid" {
			chart "Pie" class: "chart-tile" data: chartData chartType: "pie" overflow: "fit" showLegend: true style: cssSafe("--sui-chart-series-1: #ec4899; --sui-chart-series-2: #f59e0b; --sui-chart-series-3: #22c55e; --sui-chart-series-4: #38bdf8;") interactive: true
			chart "Donut" class: "chart-tile" data: chartData chartType: "donut" overflow: "fit" showLegend: true style: cssSafe("--sui-chart-bg: #2a1033; --sui-chart-fg: #fdf2f8; --sui-chart-title-color: #f9a8d4; --sui-chart-muted: #f0abfc; --sui-chart-series-1: #ec4899; --sui-chart-series-2: #a855f7; --sui-chart-series-3: #38bdf8; --sui-chart-series-4: #f472b6;") interactive: true
			chart "Funnel" class: "chart-tile" data: chartData chartType: "funnel" x: "month" y: "value" showValues: true showLegend: false style: cssSafe("--sui-chart-series-1: #ec4899;") interactive: true
			chart "Gauge" class: "chart-tile" data: kpiData format: "csv" chartType: "gauge" min: 0 max: 100 showLegend: false style: cssSafe("--sui-chart-series-1: #22c55e;") interactive: true
			chart "Bullet" class: "chart-tile" data: bulletData format: "csv" chartType: "bullet" x: "Team" showValues: true style: cssSafe("--sui-chart-series-1: #22c55e; --sui-chart-series-4: #ef4444;") interactive: true
			chart "Treemap" class: "chart-tile" data: chartData chartType: "treemap" x: "month" y: "value" showValues: true showLegend: false style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #f59e0b; --sui-chart-series-3: #22c55e; --sui-chart-series-4: #a855f7;") interactive: true
		}
	}
}

view relationships {
	title "Relationships"
	row class: "chart-page" {
		h1 "Relationships and comparisons"
		row class: "chart-nav" {
			button "Overview" onClick: goOverview
			button "Data" onClick: goData
			button "Bars" onClick: goBars
			button "Lines" onClick: goLines
			button "Distributions" onClick: goDistributions
			button "Parts" onClick: goParts
			button "Time" onClick: goTime
		}
		p "Scatter, radar and dumbbell expose correlation and before/after comparisons." style: "margin:0;color:#8fa0b8"
		row class: "chart-grid" {
			chart "Scatter" class: "chart-tile" data: scatterData chartType: "scatter" x: "x" y: "y" pointSize: "6" showGrid: true showLegend: true showValues: true style: cssSafe("--sui-chart-series-1: #38bdf8;") interactive: true
			chart "Radar" class: "chart-tile" data: multiData chartType: "radar" x: "month" pointSize: "4" showLegend: true style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #f59e0b;") interactive: true
			chart "Dumbbell" class: "chart-tile" data: dumbbellData format: "csv" chartType: "dumbbell" x: "Team" pointSize: "5" showValues: true showGrid: true style: cssSafe("--sui-chart-series-1: #f59e0b; --sui-chart-series-2: #22c55e;") interactive: true
		}
	}
}

view time {
	title "Time planning"
	row class: "chart-page" {
		h1 "Time and planning"
		row class: "chart-nav" {
			button "Overview" onClick: goOverview
			button "Data" onClick: goData
			button "Bars" onClick: goBars
			button "Lines" onClick: goLines
			button "Distributions" onClick: goDistributions
			button "Parts" onClick: goParts
			button "Relationships" onClick: goRelationships
		}
		p "Candlestick and interactive Gantt: drag tasks, hover a bar to show the resize handles, and click a group header to collapse its task rows. The table view exposes task, start, end and progress for assistive technology." style: "margin:0;color:#8fa0b8"
		chart "Interactive Gantt" class: "chart-time" data: ganttData chartType: "gantt" tableView: "both" tableAddRow: true criticalPath: true holidays: "*-11-16, 2026-11-20..2026-11-21" weekend: "sat,sun" timezone: "Europe/Rome" showValues: true interactive: true onTaskChange: ganttChanged style: cssSafe("--sui-chart-series-1: #a855f7; --sui-chart-gantt-weekend: #6b7280; --sui-chart-gantt-holiday: #ef4444; --sui-chart-gantt-today: #f59e0b; --sui-chart-gantt-dependency: #94a3b8; --sui-chart-critical: #f59e0b;")
		row class: "chart-grid" {
			chart "Candlestick" class: "chart-tile" data: ohlcData chartType: "candlestick" x: "date" showGrid: true showLegend: true style: cssSafe("--sui-chart-series-2: #22c55e; --sui-chart-series-4: #ef4444;") interactive: true
			chart "Long series" class: "chart-tile" data: longData chartType: "area" format: "csv" x: "Day" y: "Value" overflow: "zoom" showGrid: true showLegend: false style: cssSafe("--sui-chart-series-1: #38bdf8;") interactive: true
		}
	}
}
```

## Use it in the designer

1. Insert **Chart** from **Controls**. The new chart has sample CSV data.
2. Select it and open **Chart** in the inspector.
3. Choose **type** from the full list. The inspector hides fields that do not
   apply to the selected type.
4. Click **data**. The movable, resizable overlay opens: choose **Literal data**
   or **State expression**, paste JSON/XML/CSV, choose the format, validate,
   load a file or insert a sample. The overlay also shows the selected chart type
   and blocks Apply when the data shape is incompatible.
5. Apply the overlay, then set **x field** and **y field**, or leave them on auto.
6. Set **width** and **height** in the standard **Size** inspector section;
   choose **overflow**:
   - **Scrollbars** for a fixed frame with scrolling,
   - **Zoom controls** for `-`, `+`, `Reset`,
   - **Fit to frame** for small cards.
7. For point-based charts, set **point size** (line, step, smooth, area, scatter,
   combo, radar and dumbbell). The value is the point radius in pixels.
8. Enable **interactive** for tooltips/highlights and, on Gantt, drag/resize and
   group collapse. Add `onTaskChange` from **Events** when the task changes must
   be persisted.
9. Choose a **preset** and optionally fine-tune **style expr** in Appearance.
10. Run the preview and use the page navigation; switch JSON/CSV/XML on the
   **Data** page.

Literal JSON in source needs escaped braces (`\{` and `\}`). The data overlay
emits them automatically when applying a literal.

## Build and run

```bash
go build -o simpleui.exe .
./simpleui.exe check examples/chart/app.sui --json
./simpleui.exe build examples/chart/app.sui --out examples/chart/dist
./simpleui.exe serve examples/chart/app.sui --watch
```

The prebuilt web app is in `dist/`. A desktop build can be produced with:

```bash
./simpleui.exe build examples/chart/app.sui --target desktop --out examples/chart/dist
```

## Behavior and limits

- JSON: arrays of objects, arrays of arrays, `labels`/`values`, `rows`/`columns`
  and simple object maps.
- CSV: comma, semicolon, tab and pipe delimiters; quoted cells; header detection.
- XML: `row` elements with label/value attributes or child elements; DTD/entity
  declarations are rejected.
- A chart is limited to 5,000 points and bounded text/label lengths.
- Pie and donut ignore `x`/`y` and need positive values.
- Chart data is text: labels and values are written with `textContent`; no
  payload can inject HTML or script.
- Tooltips and highlights are enabled by `interactive: true`; they affect display
  only and never rewrite the source data.
- Gantt drag/resize/group collapse need `interactive: true`. Without
  `onTaskChange`, edits are visual until the next re-render.
- Hover a Gantt bar to reveal the start/end resize handles. Dragging the bar
  moves it; dragging a handle changes only that edge.
- Collapsing a Gantt group removes its task rows and compacts the remaining
  rows. The group header stays visible with its task count.
- `tableView: "table"` renders a native table with Task, Start, End, Progress,
  Group and Depends on. `"both"` shows the timeline and table together. Table
  cells use `textContent` and the table has a caption, headers and row data.
- A chart is a visual summary. Pair it with a table when exact values matter.

See [API reference: chart](../../docs/API_REFERENCE.md#chart) and
[GUI editor: chart](../../docs/GUI_EDITOR.md#chart-control).

































