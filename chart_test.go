package main

import (
	"os"
	"strings"
	"testing"
)

const chartTestSource = `state chartData = "[\{\"month\":\"Jan\",\"value\":12\}]"
view main {
	chart "Revenue" name: "revenue" data: chartData chartType: "donut" theme: "light" format: "auto" x: "month" y: "value" width: "520px" height: "300px" overflow: "zoom" zoom: 1.5 pointWidth: 56 pointSize: 5 tableView: "both" tableFilters: false showGrid: false showLegend: true showValues: true style: cssSafe("--sui-chart-series-1: #00e5ff;")
}
`

const chartLiteralSource = `view main {
	chart "Escaped <b>title</b>" data: "Month,Value\nJan,12\nFeb,18" chartType: "bar" format: "csv"
}
`

func TestChartDesignerDefaultsFillContainer(t *testing.T) {
	start := strings.Index(guiHTML, "function newChartNode(")
	if start < 0 {
		t.Fatal("newChartNode not found in designer HTML")
	}
	end := strings.Index(guiHTML[start:], "function newGraphNode(")
	if end < 0 {
		t.Fatal("newGraphNode boundary not found after newChartNode")
	}
	block := guiHTML[start : start+end]
	for _, want := range []string{
		`n.chartWidth = "100%"`,
		`n.chartHeight = "100%"`,
		`n.chartTableView = "chart"`,
	} {
		if !strings.Contains(block, want) {
			t.Errorf("new chart node missing %q", want)
		}
	}
}

func TestChartDefaultSizeFillsContainer(t *testing.T) {
	if !strings.Contains(suiChartRuntime, `var w = attr(root, "width") || "100%"`) ||
		!strings.Contains(suiChartRuntime, `var h = attr(root, "height") || "100%"`) {
		t.Error("chart runtime does not default to a container-filling size")
	}
	if !strings.Contains(suiCssBase, "width: var(--sui-chart-width, 100%)") ||
		!strings.Contains(suiCssBase, "height: var(--sui-chart-height, 100%)") {
		t.Error("chart CSS does not fall back to a container-filling size")
	}
	if !strings.Contains(suiCssBase, ".sui-chart-table-both > .sui-chart-table") ||
		!strings.Contains(suiCssBase, "width: 100%") {
		t.Error("chart table does not follow the control width")
	}
	if !strings.Contains(suiCssBase, "min-height: var(--sui-chart-min-height, 240px)") ||
		!strings.Contains(suiCssBase, ".sui-chart.sui-chart-fixed-height { min-height: 0; }") {
		t.Error("chart CSS is missing the auto-height minimum or the fixed-height override")
	}
	if !strings.Contains(suiChartRuntime, `classList.toggle("sui-chart-fixed-height", !autoHeight)`) {
		t.Error("chart runtime does not skip the minimum for explicit pixel heights")
	}
}

func TestChartSplitterKeyboardResize(t *testing.T) {
	for _, want := range []string{
		`splitter.addEventListener("keydown"`,
		`if (key === "ArrowUp") next = current + step;`,
		`else if (key === "ArrowDown") next = current - step;`,
		`else if (key === "PageUp")`,
		`else if (key === "PageDown")`,
		`else if (key === "Home") next = 80;`,
		`else if (key === "End") next = chartTableHeightMax(root);`,
		`function syncChartSplitterAria(root)`,
		`splitter.setAttribute("aria-valuenow", String(now));`,
		`splitter.setAttribute("aria-valuetext", now + " pixels");`,
		`function setChartTableHeight(root, height)`,
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("chart runtime missing splitter keyboard support %q", want)
		}
	}
}

func TestChartRender(t *testing.T) {
	prog, dl := ParseSource(chartTestSource, "chart.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`class="sui-chart"`,
		`data-sui-chart="1"`,
		`data-sui-chart-type="donut"`,
		`data-sui-chart-theme="light"`,
		`data-sui-chart-color-picker="1"`,
		`data-sui-chart-format="auto"`,
		`data-sui-chart-x="month"`,
		`data-sui-chart-y="value"`,
		`data-sui-chart-width="520px"`,
		`data-sui-chart-height="300px"`,
		`data-sui-chart-overflow="zoom"`,
		`data-sui-chart-zoom="1.5"`,
		`data-sui-chart-point-width="56"`,
		`data-sui-chart-point-size="5"`,
		`data-sui-chart-table-view="both"`,
		`data-sui-chart-table-filters="0"`,
		`data-sui-chart-grid="0"`,
		`data-sui-chart-legend="1"`,
		`data-sui-chart-values="1"`,
		`data-sui-chart-data="${__sui_esc(chartData)}"`,
		`class="sui-chart-viewport"`,
		`class="sui-chart-splitter"`,
		`class="sui-chart-stage"`,
		`sui-chart-menu-btn`,
		`data-zoom="in"`,
		`data-zoom="reset"`,
		`--sui-chart-series-1: #00e5ff`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`class="sui-chart"`,
		`data-sui-chart="1"`,
		`data-sui-chart-type="donut"`,
		`data-sui-chart-theme="light"`,
		`data-sui-chart-color-picker="1"`,
		`data-sui-chart-overflow="zoom"`,
		`data-sui-chart-zoom="1.5"`,
		`data-sui-chart-point-size="5"`,
		`data-sui-chart-table-view="both"`,
		`data-sui-chart-table-filters="0"`,
		`class="sui-chart-viewport"`,
		`class="sui-chart-splitter"`,
		`class="sui-chart-table"`,
		`class="sui-chart-error"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static HTML missing %q", want)
		}
	}
	css := GenerateCss(prog)
	for _, want := range []string{
		".sui-chart", ".sui-chart-viewport", ".sui-chart-stage",
		".sui-chart-grid", ".sui-chart-series-1", "var(--sui-chart-series-1",
		"var(--sui-chart-width", ".sui-chart-zoom",
		"--sui-chart-control-bg", "--sui-chart-control-hover-fg",
		"--sui-chart-legend-color", "--sui-chart-scroll-thumb",
		"--sui-chart-tooltip-max-width", "--sui-chart-tooltip-font-size",
		"--sui-chart-tooltip-radius", ".sui-chart-table", "sui-chart-table-only",
		"sui-chart-table-both", "--sui-chart-table-font-size",
		"--sui-chart-table-cell-padding", "--sui-chart-table-stripe-opacity",
		"sui-chart-table-row-selected", "sui-chart-table-cell-flash",
		"[data-sui-chart-theme=\"light\"]", "color-scheme: light",
		"--sui-chart-table-selection-bg", "--sui-chart-table-focus-row-bg", "sui-chart-ghost",
		"sui-chart-table-dragging .sui-chart-table-step-btn",
		"td:focus-within .sui-chart-table-step-btn",
		"sui-chart-table-series-color", "sui-chart-legend-colorizable",
		".sui-chart-swatch.sui-chart-series-1", "sui-chart-gantt-task-color", "sui-chart-gantt-label",
		"sui-chart-gantt-selected", "sui-chart-gantt-label-selected", "sui-chart-gantt-flash",
		"sui-chart-table-tabs", "sui-chart-table-tab", "sui-chart-resource-table", "sui-chart-resource-delete",
		"sui-chart-resource-picker-dialog", "sui-chart-resource-picker-select", "sui-chart-gantt-month-band",
		"sui-chart-gantt-week-band", "sui-chart-gantt-day-band", "sui-chart-gantt-name-resize", "sui-chart-gantt-commit", "sui-chart-gantt-history",
		"sui-chart-state-shape", "sui-chart-state-edge-line", "sui-chart-state-label",
		"sui-chart-state-table", "sui-chart-state-col-resize", "sui-chart-state-edge-handle", "sui-chart-state-row-selected", "sui-chart-state-tag-bg",
		"sui-chart-state-color", "sui-chart-state-tag-link", "sui-chart-state-tag-arrow",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("generated CSS missing %q", want)
		}
	}
	if strings.Contains(css, "td:hover .sui-chart-table-step-btn") {
		t.Error("generated CSS must not reveal steppers on hover")
	}
}

func TestChartLiteralStaticAndEscaping(t *testing.T) {
	prog, dl := ParseSource(chartLiteralSource, "chart.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	html := GenerateHtmlPage(prog)
	for _, want := range []string{
		`data-sui-chart-data="Month,Value`,
		`data-sui-chart-type="bar"`,
		`data-sui-chart-format="csv"`,
		`data-sui-chart-table-view="chart"`,
		`Escaped &lt;b&gt;title&lt;/b&gt;`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static chart HTML missing %q", want)
		}
	}
	if strings.Contains(html, "Escaped <b>title</b>") {
		t.Error("chart title must be escaped")
	}
	if strings.Contains(html, "data-sui-chart-width=") || strings.Contains(html, "data-sui-chart-height=") {
		t.Error("chart without explicit width/height must use the standard component size")
	}
}

func TestChartRuntimeDataParsers(t *testing.T) {
	for _, want := range []string{
		"parseJson", "parseXml", "parseCsv", "csvRows", "DOMParser",
		"pointsFromObjects", "pointsFromArrays", "data-sui-chart-",
		"[data-sui-chart]", "drawPie", "drawXY", "drawHorizontalBar",
		"drawMultiBars", "drawCombo", "drawHistogram", "drawWaterfall",
		"parseGanttData", "ganttDate", "gantt-progress", "gantt-milestone",
		"gantt-dependency", "gantt-weekend", "bindGanttDrag", "bindChartInteractions",
		"data-sui-chart-gantt-resize", "sui-chart-gantt-handle", "sui-chart-gantt-group-hit",
		"if (!to || from === to) continue;", "renderGanttTable", "sui-chart-table-only",
		"Depends on", "sui-chart-table-both", "commitGanttTableEdit",
		"notifyGanttTasks", "__suiGanttLocal", "sui-chart-table-input", "renderChartTable",
		"decorateTable", "applyTableState", "tableDownload", "sui-chart-table-tools",
		"sui-chart-table-cols", "aria-sort",
		"sui-chart-table-col-filter", "sui-chart-table-menu-action",
		"sui-chart-table-settings-btn", "sui-chart-table-submenu",
		"sui-chart-table-filter-btn", "filtersEnabled", "table-filters",
		"state.sorts", "state.order", "localStorage", "draggable", "data-col",
		"pushGanttHistory", "undoGantt", "redoGantt", "selectionStorageKey",
		"chartTableHistory", "pushChartTableHistory", "undoChartTable", "redoChartTable",
		"applyChartTableSnapshot", "tableUndo", "tableRedo", "keepTable", "pendingUntil", "redrawRoot",
		"cancelTableEdit", "__suiCommittedValue", "__suiTableCancelEdit",
		"openSeriesColorPicker", "restoreSeriesColors", "sui-chart-table-series-color",
		"sui-chart-legend-colorizable", "sui.chart.colors.", "colorPickerEnabled",
		"ganttLegendItems", "ganttTaskColor", "ganttColorStorageKey", "sui-chart-gantt-task-color",
		"legendData.length > 1",
		"sui-chart-ghost", "prefers-reduced-motion", "__suiChartReady",
		"ganttCriticalPath", "ganttWbsTree", "applyGanttWbsRollup", "ganttParentWouldCycle",
		"GANTT_CONSTRAINT_OPTIONS", "ASAP - As soon as possible", "data-filter-options", "effectiveDuration", "ganttWorkingDuration",
		"sui-chart-table-col-resize", "autoFitColumn", "dblclick", "Reset column widths", "state.widths", "tableLayout",
		"ganttVisibleRowCount", "minRowHeight", "sui-chart-gantt-arrow-open", "ganttTaskWarnings", "sui-chart-gantt-deadline", "deadline",
		"ganttEffectiveProgress", "ganttRemainingDuration", "sui-chart-gantt-actual",
		"actualStart", "actualEnd", "workDone",
		"__suiGanttTimelineZoom", "Zoom timeline in", "Reset timeline zoom", "ganttTimelineZoomAt", "touchstart", "ctrlKey",
		"updateGanttStickyLabels", "sui-chart-gantt-sticky", "ganttPathCost", "ganttPathOverlaps", "scheduleLiveDraw", "GANTT_LABEL_WIDTH", "stage.style.minWidth", "allArrowBoxes",
		"routeInset", "plotStart", "entersFromLeft", "arrowBaseX", "entryX",
		"laneClear", "altBranchY", "bar.id === ignoreA && s === 1", "milestoneLike ? Math.max(bh, 14)",
		"monotonicInRowDirection", "ganttLinkGap", "verticalDir", "minRowHeight", "show-non-working-days", "progress-column",
		"bindGanttDeadlineDrag", "data-sui-chart-gantt-deadline", "sui-chart-gantt-progress-column",
		"not satisfied", "missing predecessor", "linkGap",
		"openGanttCalendar", "sui-chart-calendar-overlay", "sui-chart-datepicker", "sui-chart-datepicker-grid",
		"addSubmenu", "sui-chart-menu-submenu", "ganttEffectiveCalendar", "oncalendarchange",
		"ganttLinkSpecs", "ganttShiftWorkingDays", "openGanttLinkEditor", "ganttNormalizeDependencies", "ganttDependencyCycle", "ganttDependencyWouldCycle", "ganttDependencyControl",
		"ganttResourceNames", "ganttResourceControl", "ganttTaskResources", "sui-chart-gantt-resource-chip", "ganttResourceUniverse",
		"ganttWorkloadRows", "openGanttResourceWorkload", "Resource workload...", "sui-chart-workload-table",
		"ganttResourceRegistry", "ganttResourceEntry", "ganttEnsureResourceEntries", "renderGanttResourceTable",
		"appendGanttTableTabs", "commitGanttResourceField", "ganttReplaceTaskResource", "sui-chart-table-tab",
		"ganttResourceCostFor", "ganttResourceOverloads", "ganttOverloadIncrease", "ganttOverloadMapIncrease",
		"maxPeople", "dailyCost", "Max people", "Daily cost", "ganttTaskResourceUnits", "ganttResourceUnitMap",
		"commitGanttResourceUnits", "sui-chart-gantt-resource-units", "resourceUnits", "ganttResourcePeak",
		"openGanttResourcePicker", "sui-chart-resource-picker-dialog", "sui-chart-resource-picker-select",
		"loadStoredGanttResources", "saveGanttResources", "onResourceChange", "data-sui-chart-onresourcechange", "Reset saved",
		"bindGanttLinkDrag", "sui-chart-gantt-link-handle", "sui-chart-gantt-link-preview", "sui-chart-gantt-link-target",
		"openGanttGroupDialog", "Assign to group...", "sui-chart-group-dialog",
		"createGanttGroup", "Create group from selection", "openGanttTaskMenu", "toggleGanttTaskSelection",
		"Hide progress column", "Show progress column", "Hide non-working days", "Show non-working days", "bandKind",
		"ganttNameColumnWidth", "bindGanttNameResize", "sui-chart-gantt-name-resize", "ganttProgressFill", "showNonWorkingBands",
		"bindGanttKeyboard", "ganttKeyboardMove", "ganttDraftStorageKey", "commitGanttChanges", "openGanttHistory", "onCommit", "ganttPendingChanges", "sui:gantt-commit",
		"parseStateData", "drawStateDiagram", "renderStateTable", "sui-chart-state",
		"sui-chart-state-table-tabs", "data-sui-chart-state-tab-button", "Shapes (", "Links (",
		"stateRevealSelectedRow", "stateSyncNodeGridGeometry", "stateKeyboardGeometry",
		"stateKeyboardDelete", "stateKeyboardStep", "Select shape ", "Select link ",
		"sui-chart-state-cell-input", "sui-chart-tag-field-wide", "Description: ",
		"[\"Element\", \"Tag\", \"Hint\", \"Type\", \"\"]",
		"[\"Element\", \"Tag\", \"Hint\", \"Type\", \"From\", \"To\", \"\"]",
		"hint: normalizeLineText(stateFieldText(obj, [\"hint\", \"tooltip\", \"tip\"]))",
		"if (n.hint) out.hint = n.hint;",
		"stateHintsEnabled",
		"data-sui-chart-state-hints",
		"Show hover hints",
		"Hint settings...",
		"openStateHintSettings", "HINTS settings", "Edit mode",
		"applyStateHintSettings", "stateChartHintColor",
		"data-sui-chart-state-hint-bg", "data-sui-chart-state-hint-opacity",
		"data-sui-chart-state-hint-font", "data-sui-chart-state-hint-border-width",
		"sui-chart-hint-preview-box", "Hint preview", "Hint background opacity", "stateTagFieldColorAlpha", "sui-chart-color-alpha-field", "Corner radius",
		"sui-chart-setting-row", "sui-chart-setting-control", "sui-chart-hint-groups", "sui-chart-hint-group", "Hints and colors", "Size and spacing",
		"sui-chart-state-create-btn", "sui-chart-state-create-menu",
		"openStateCreateMenu", "Add graphic element", "Choose shape type", "aria-keyshortcuts",
		"stateSettingsStorageKey", "stateSettingsDefaultStorageKey", "applyStateSettings",
		"saveStateSettings", "resetStateSettings", "Save as default", "Reset settings",
		"stateRouteClearance", "stateBuildOuterLaneRoute", "stateRouteStats", "__suiStateRouteAudit",
		"data-sui-chart-route-audit", "node clearance", "link clearance", "off-grid segment",
		"stateTagLink", "stateColorControl", "stateNodeDisplayColor", "openStateTagFamilySettings", "TAG", "edge_", "stateSelfLoopGeometry", "statePortPoint",
		"ganttDependencyIdsFromText", "ganttPathHitsBars", "localPoints", "globalPoints",
		"data-sui-chart-gantt-wbs-toggle", "autoScheduleGantt", "sui-chart-gantt-dep-chip",
		"ganttWeekendSet", "ganttHolidayMatcher", "ganttDateParts",
		"ganttZonedTime", "ganttWallClockValue", "ganttPayloadDate", "startTimed", "endTimed",
		"addGanttTask", "ganttNewTaskId", "sui-chart-table-add-row",
		"addChartRow", "chartDataPayload", "notifyChartData", "data-sui-chart-ondatachange", "table-add-row",
		"tableMultiSet", "toggleTableRow", "selectTableRange", "selectAllVisibleTableRows",
		"deleteSelectedTableRows", "deleteChartRows", "deleteGanttRows", "commitChartTableEdit",
		"applyVirtualRows", "sui-chart-table-spacer", "sui-chart-table-row-multi",
		"tableSelectionMode", "tableColSet", "tableCellSet", "tableSelectedData",
		"appendNumberStepper", "sui-chart-table-step-btn", "sui-chart-table-row-delete",
		"bindTableContextMenu", "sui-chart-table-context", "sui-chart-table-col-multi",
		"setTableStatus", "tableClipboardWrite", "tableClipboardRead", "tablePasteSelection",
		"parseTableClipboard", "tablePasteTarget", "pasteGanttMatrix", "pasteChartMatrix",
		"tableHtmlDocument", "tableXlsxBytes", "xlsxZip", "printDocument",
		"Export options", "Print / Save PDF", "Export selection XLSX",
		"sui-chart-table-status", "Paste selection",
		"Export selection", "Copy selection", "Delete selected rows",
		"tableColumnKind", "tableFilterMatch", "tableFilterDate", "tableVisibleData",
		"tableCopyVisible", "tablePrint", "table-filter-range", "table-filter-values",
		"sui-chart-gantt-critical", "sui-chart-gantt-holiday",
		"sui-chart-table-submenu-trigger",
		"bindChartMenu", "chartExportSvg", "chartExportPng", "chartExportData",
		"selectGanttTask", "applyGanttSelection", "applyGanttHighlight",
		"__SUI_CHART_GANTT_UI", "ganttUiKey", "setGanttHighlight",
		"sui-chart-table-row-selected", "sui-chart-table-cell-flash",
		"sui-chart-gantt-flash", "sui-chart-gantt-selected",
		"selectChartMark", "applyChartSelection", "data-sui-chart-row",
		"pointerdown", "pointermove", "sui-chart-gantt-dragging",
		"chartHit", "sui-chart-tooltip", "sui-chart-point-hit", "data-sui-chart-ontaskchange",
		"updateStateGridBackground", "bindStateGrid", "linear-gradient(to right", "backgroundPosition", "gridClone.style.display",
		"bindChartKeyboard", "ArrowRight", "root.setAttribute(\"tabindex\", \"0\")",
		"validateChartData", "barstacked", "bargrouped", "linePath",
		"MAX_POINTS", "MutationObserver", "aria-label", "textContent",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("chart runtime missing %q", want)
		}
	}
	if !strings.Contains(suiChartRuntime, `var labelText = edge.tag || "";`) {
		t.Error("state links must render only the tag text")
	}
	if strings.Contains(suiChartRuntime, "edge.tag || edge.label") || strings.Contains(suiChartRuntime, "stateDrawTag(shapeTagLayer, node.label") {
		t.Error("state diagrams must not render element names in the graph")
	}
	if strings.Contains(suiChartRuntime, "innerHTML") {
		t.Error("chart runtime must not use innerHTML")
	}
	if strings.Contains(suiChartRuntime, "insertAdjacentHTML") {
		t.Error("chart runtime must not use insertAdjacentHTML")
	}
	for _, banned := range []string{"forecast", "Forecast", "Finish variance", "ganttFinishVariance"} {
		if strings.Contains(suiChartRuntime, banned) {
			t.Errorf("chart runtime must not contain forecast support: %q", banned)
		}
	}
}

func TestChartHintTooltipCSS(t *testing.T) {
	for _, want := range []string{
		"var(--sui-chart-tooltip-bg",
		"var(--sui-chart-tooltip-fg",
		"var(--sui-chart-tooltip-border-width, 1px)",
		"var(--sui-chart-tooltip-border,",
		"var(--sui-chart-tooltip-padding",
		"var(--sui-chart-tooltip-radius",
		"var(--sui-chart-tooltip-font-family",
		"var(--sui-chart-tooltip-font-size",
		"var(--sui-chart-tooltip-shadow",
		".sui-chart-hint-preview-box",
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("chart tooltip CSS missing %q", want)
		}
	}
}

func TestChartReactiveFlag(t *testing.T) {
	src := `state autoMode = true
view main {
	chart "C" autoSchedule: autoMode
}`
	prog, dl := ParseSource(src, "chart.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `${((autoMode) ? "1" : "0")}`) {
		t.Fatalf("reactive autoSchedule flag was not emitted dynamically")
	}
}

func TestChartValidation(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"type", `view main { chart "C" chartType: "unknown-type" }`, "chartType must be one of"},
		{"format", `view main { chart "C" format: "yaml" }`, "format must be one of"},
		{"theme", `view main { chart "C" theme: "sepia" }`, "theme must be one of"},
		{"overflow", `view main { chart "C" overflow: "crop" }`, "overflow must be one of"},
		{"grid", `view main { chart "C" showGrid: 1 }`, "showGrid must be a bool"},
		{"colorPicker", `view main { chart "C" colorPicker: 1 }`, "colorPicker must be a bool"},
		{"autoSchedule", `view main { chart "C" autoSchedule: 1 }`, "autoSchedule must be a bool"},
		{"linkStyle", `view main { chart "C" linkStyle: "zigzag" }`, "linkStyle must be one of"},
		{"arrowType", `view main { chart "C" arrowType: "diamond" }`, "arrowType must be one of"},
		{"linkWidth", `view main { chart "C" linkWidth: "wide" }`, "linkWidth must be a number"},
		{"zoom", `view main { chart "C" zoom: "big" }`, "zoom must be a number"},
		{"children", `view main { chart "C" { p "nested" } }`, "chart cannot contain child elements"},
	} {
		prog, dl := ParseSource(tc.src, "chart.sui")
		if dl.HasErrors() {
			t.Fatalf("%s parse: %s", tc.name, dl.String())
		}
		cd := Check(prog)
		if !cd.HasErrors() || !strings.Contains(cd.String(), tc.want) {
			t.Errorf("%s: expected %q, got %s", tc.name, tc.want, cd.String())
		}
	}
}

func TestChartLinkAppearanceDesigner(t *testing.T) {
	src := `view main {
	chart "C" linkColor: "#ff0000" linkWidth: 2.5 linkStyle: "dashed" arrowType: "open" arrowSize: 7
}`
	prog, dl := ParseSource(src, "chart.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{`"chartLinkColor":"#ff0000"`, `"chartLinkWidth":"2.5"`, `"chartLinkStyle":"dashed"`, `"chartArrowType":"open"`, `"chartArrowSize":"7"`} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q", want)
		}
	}
	out := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{`linkColor: "#ff0000"`, `linkWidth: 2.5`, `linkStyle: "dashed"`, `arrowType: "open"`, `arrowSize: 7`} {
		if !strings.Contains(out, want) {
			t.Errorf("designer export missing %q: %s", want, out)
		}
	}
}

func TestChartDesignerRoundTrip(t *testing.T) {
	prog, dl := ParseSource(chartTestSource, "chart.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	model := StringifyJson(guiModelJson(prog))
	for _, want := range []string{
		`"kind":"chart"`,
		`"name":"revenue"`,
		`"chartData":"chartData"`,
		`"chartDataExpr":true`,
		`"chartType":"donut"`,
		`"chartTheme":"light"`,
		`"chartColorPicker":true`,
		`"chartFormat":"auto"`,
		`"chartX":"month"`,
		`"chartY":"value"`,
		`"chartWidth":"520px"`,
		`"chartHeight":"300px"`,
		`"chartOverflow":"zoom"`,
		`"chartZoom":"1.5"`,
		`"chartPointWidth":"56"`,
		`"chartPointSize":"5"`,
		`"chartTableView":"both"`,
		`"chartTableFilters":false`,
		`"chartGrid":false`,
		`"chartLegend":true`,
		`"chartLinkColor":"#94a3b8"`,
		`"chartLinkStyle":"auto"`,
		`"chartArrowType":"triangle"`,
		`"chartValues":true`,
		`--sui-chart-series-1: #00e5ff`,
	} {
		if !strings.Contains(model, want) {
			t.Errorf("designer model missing %q: %s", want, model)
		}
	}
	src := viewNodesSource(prog.views[0].body, 1)
	for _, want := range []string{
		`chart "Revenue"`,
		`name: "revenue"`,
		`data: chartData`,
		`chartType: "donut"`,
		`theme: "light"`,
		`overflow: "zoom"`,
		`zoom: 1.5`,
		`pointWidth: 56`,
		`pointSize: 5`,
		`tableView: "both"`,
		`tableFilters: false`,
		`showGrid: false`,
		`showValues: true`,
		`style: cssSafe`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("designer source missing %q:\n%s", want, src)
		}
	}
	for _, want := range []string{
		"chartControls", "newChartNode", "CHART_STYLE_PRESETS", "Style presets",
		"chart-data-btn", "chartDataButtonText", "openChartDataEditor", "cs-theme", "chartTheme", "cs-colorpicker", "chartColorPicker",
		"chartDataValidateText", "chartDataCompatibility", "chartDataProfile",
		"chartDataApply", "chartDataSaveGeom",
		"Mode", "show grid", "Chart settings...", "Edit hints...",
		"chartPresetStyleExpr", "chart: '<path", `"date", "time", "tree", "treeitem", "chart"`,
		"node-type", "sel-type", "--sui-chart-control-bg",
		"--sui-chart-control-hover-fg",
		"CHART_TYPE_OPTIONS", "chartDataCompatibility", "chartDataProfile",
		"cd-type", "barstacked", "bargrouped", "barhorizontal", "histogram", "waterfall",
		"chartBins", "chartPointSize", "Point size", "cs-pointwidth",
		"applyChartTypeVisibility", "TYPE_HINTS", "setRowVisible",
		"Fields below follow", "radar", "gauge", "funnel", "heatmap", "candlestick",
		"chartInteractive", "onTaskChange", "onDataChange", "interactive",
		"cttcfg", "ctt-preview", "openChartTooltipCfg", "applyChartTooltipCfg",
		"resetChartTooltipCfg", "--sui-chart-tooltip-max-width",
		"--sui-chart-tooltip-font-size", "--sui-chart-tooltip-delay",
		"chartTableView", "Mode", "Timeline + table",
		"ctablecfg", "ctable-preview", "openChartTableCfg", "applyChartTableCfg",
		"resetChartTableCfg", "--sui-chart-table-font-size",
		"--sui-chart-table-stripe-opacity", "--sui-chart-table-header-position",
		"chintscfg", "chintscfgbox", "chints-preview", "openChartHintsCfg", "applyChartHintsCfg",
		"resetChartHintsCfg", "chartStateHintBg", "chartStateHintOpacity", "stateHintBg", "stateHintRadius",
		"cs-state-edit", "Edit mode",
		"csettings", "Chart settings...", "openChartSettings", "applyChartSettings",
		"showChartSettingsTab", "cset-tab", "cs-preset", "cs-tableview", "cs-tb-filters", "cs-tb-addrow",
		"cs-pointwidth", "chartPointWidth", "openChartElementHintsFromSettings",
		"chartTableFilters", "chartTableAddRow", "chartTableDeleteRow", "chartTableSteppers",
		"chartTableContextMenu", "chartTableSelection", "cs-tb-deleterow", "cs-tb-steppers",
		"cs-tb-contextmenu", "cs-tb-selection", "cs-critical", "cs-holidays", "cs-weekend", "cs-timezone", "cs-gantt-autoschedule", "chartAutoSchedule",
		"cs-gantt-linkcolor", "cs-gantt-linkwidth", "cs-gantt-linkstyle", "cs-gantt-arrowtype", "cs-gantt-arrowsize", "cs-gantt-barborder", "cs-gantt-barborder-enable",
		"chartLinkColor", "chartLinkWidth", "chartLinkStyle", "chartArrowType", "chartArrowSize", "chartBarBorderColor", "onCalendarChange",
		"chartCriticalPath", "chartHolidays", "chartWeekend", "chartTimezone",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer chart support missing %q", want)
		}
	}
}

func TestChartDataEditorOverlay(t *testing.T) {
	for _, want := range []string{
		`id="chart-data-layer"`, `id="cd-panel"`, `id="cd-resize"`,
		`id="cd-text"`, `id="cd-expr"`, `id="cd-format"`,
		`id="cd-load"`, `id="cd-sample"`, `id="cd-pretty"`,
		`id="cd-apply"`, `id="cd-cancel"`, `id="cd-fileinput"`,
		"#chart-data-layer", ".cd-panel", ".cd-resize", ".cd-head",
		"Drag the header", "Ctrl+Enter", "Load file", "Insert sample",
		"State expression", "Pretty JSON",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("chart data overlay missing %q", want)
		}
	}
	if strings.Contains(guiHTML, `id="cd-panel" class="cd-panel" style=`) {
		t.Error("chart data panel should be positioned by the editor, not hard-coded")
	}
}

func TestChartAdvancedTypesCompile(t *testing.T) {
	src := `state single = "Month,Value\nJan,12\nFeb,18"
state multi = "[\{\"month\":\"Jan\",\"sales\":12,\"cost\":5\},\{\"month\":\"Feb\",\"sales\":18,\"cost\":8\}]"
state samples = "[3,5,5,6,7,8,9]"
state falls = "Step,Delta\nStart,20\nSales,-5\nFinal,8"
state heat = "Day,Hour,Value\nMon,9,3\nMon,10,5"
state ohlc = "[\{\"date\":\"D1\",\"open\":10,\"high\":15,\"low\":8,\"close\":13\}]"
state box = "Name,Min,Q1,Median,Q3,Max\nA,2,4,6,8,12"
state bullet = "Team,Value,Target,Max\nA,72,80,100"
state dumbbell = "Team,Before,After\nA,40,72"
state gantt = "Task,Start,End\nDesign,0,5\nBuild,5,14"
view main {
	chart "H" data: single format: "csv" chartType: "barhorizontal" x: "Month" y: "Value"
	chart "S" data: multi chartType: "barstacked" x: "month"
	chart "G" data: multi chartType: "bargrouped" x: "month"
	chart "C" data: multi chartType: "combo" x: "month"
	chart "Step" data: single format: "csv" chartType: "step" x: "Month" y: "Value"
	chart "Smooth" data: single format: "csv" chartType: "smooth" x: "Month" y: "Value"
	chart "Hist" data: samples chartType: "histogram" bins: 6
	chart "W" data: falls format: "csv" chartType: "waterfall" x: "Step" y: "Delta"
	chart "Radar" data: multi chartType: "radar" x: "month"
	chart "Gauge" data: samples chartType: "gauge" min: 0 max: 20
	chart "Funnel" data: single format: "csv" chartType: "funnel" x: "Month" y: "Value"
	chart "Heat" data: heat format: "csv" chartType: "heatmap" x: "Day"
	chart "Candle" data: ohlc chartType: "candlestick" x: "date"
	chart "Tree" data: single format: "csv" chartType: "treemap" x: "Month" y: "Value"
	chart "Box" data: box format: "csv" chartType: "boxplot" x: "Name"
	chart "Bullet" data: bullet format: "csv" chartType: "bullet" x: "Team"
	chart "Dumb" data: dumbbell format: "csv" chartType: "dumbbell" x: "Team"
	chart "Gantt" data: gantt format: "csv" chartType: "gantt" x: "Task" tableView: "both" tableAddRow: true tableDeleteRow: true tableSteppers: true tableContextMenu: true tableSelection: "cell" criticalPath: true holidays: "*-11-16, 2026-11-20..2026-11-21" weekend: "fri,sat" timezone: "Europe/Rome" interactive: true
}`
	prog, dl := ParseSource(src, "chart_types.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-chart-type="barhorizontal"`,
		`data-sui-chart-type="barstacked"`,
		`data-sui-chart-type="bargrouped"`,
		`data-sui-chart-type="combo"`,
		`data-sui-chart-type="step"`,
		`data-sui-chart-type="smooth"`,
		`data-sui-chart-type="histogram"`,
		`data-sui-chart-type="waterfall"`,
		`data-sui-chart-type="radar"`,
		`data-sui-chart-type="gauge"`,
		`data-sui-chart-type="funnel"`,
		`data-sui-chart-type="heatmap"`,
		`data-sui-chart-type="candlestick"`,
		`data-sui-chart-type="treemap"`,
		`data-sui-chart-type="boxplot"`,
		`data-sui-chart-type="bullet"`,
		`data-sui-chart-type="dumbbell"`,
		`data-sui-chart-type="gantt"`,
		`data-sui-chart-bins="6"`,
		`data-sui-chart-min="0"`,
		`data-sui-chart-max="20"`,
		`data-sui-chart-interactive="1"`,
		`data-sui-chart-table-view="both"`,
		`data-sui-chart-table-add-row="1"`,
		`data-sui-chart-table-delete-row="1"`,
		`data-sui-chart-table-steppers="1"`,
		`data-sui-chart-table-context-menu="1"`,
		`data-sui-chart-table-selection="cell"`,
		`data-sui-chart-critical-path="1"`,
		`data-sui-chart-holidays="*-11-16, 2026-11-20..2026-11-21"`,
		`data-sui-chart-weekend="fri,sat"`,
		`data-sui-chart-timezone="Europe/Rome"`,
		`data-sui-chart-link-color="#94a3b8"`,
		`data-sui-chart-link-width="1"`,
		`data-sui-chart-bar-border-color=""`,
		`data-sui-chart-link-style="auto"`,
		`data-sui-chart-arrow-type="triangle"`,
		`data-sui-chart-arrow-size=""`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("advanced chart output missing %q", want)
		}
	}
}

func TestGanttCalendarExample(t *testing.T) {
	raw, err := os.ReadFile("examples/gantt_calendar/app.sui")
	if err != nil {
		t.Fatalf("read gantt calendar example: %v", err)
	}
	prog, dl := ParseSource(string(raw), "examples/gantt_calendar/app.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-chart-holidays="${__sui_esc(holidayList)}"`,
		`data-sui-chart-weekend="${__sui_esc(weekendDays)}"`,
		`data-sui-chart-timezone="${__sui_esc(tz)}"`,
		`data-sui-chart-table-add-row="1"`,
		"setRecurringRange", "setNoWeekend", "setNewYork", "tasksChanged",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("gantt calendar example missing %q", want)
		}
	}
}

func TestGanttWbsExample(t *testing.T) {
	raw, err := os.ReadFile("examples/gantt_wbs/app.sui")
	if err != nil {
		t.Fatalf("read gantt WBS example: %v", err)
	}
	prog, dl := ParseSource(string(raw), "examples/gantt_wbs/app.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-chart-auto-schedule="${((autoMode) ? "1" : "0")}"`,
		`data-sui-chart-table-view="both"`,
		"parent", "User interviews", "Sign-off", "toggleAuto",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("gantt WBS example missing %q", want)
		}
	}
}

func TestStateMachineChartType(t *testing.T) {
	src := `view main {
	chart "State" data: "From,To,Label\nIdle,Running,start" chartType: "state" stateTagsMovable: true stateRoutesEditable: true tableView: "both"
}`
	prog, dl := ParseSource(src, "state.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{`data-sui-chart-type="state"`, `data-sui-chart-table-view="both"`, `data-zoom="fit"`, `data-zoom="fitall"`, `data-sui-chart-state-tags-movable="1"`, `data-sui-chart-state-routes-editable="1"`, `data-sui-chart-state-handles="1"`} {
		if !strings.Contains(js, want) {
			t.Errorf("state chart missing %q", want)
		}
	}
	if strings.Contains(js, "state-route-algorithm") {
		t.Error("state chart must not expose a routing algorithm option")
	}
}

func TestStateDiagramViewOnlyHandles(t *testing.T) {
	src := `view main {
	chart "State" data: "From,To,Label\nIdle,Running,start" chartType: "state" stateGrid: false stateHandles: false tableView: "both"
}`
	prog, dl := ParseSource(src, "state.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, `data-sui-chart-state-handles="0"`) {
		t.Error("stateHandles: false must emit a disabled handles attribute")
	}
	if !strings.Contains(js, `data-sui-chart-state-grid="0"`) {
		t.Error("stateGrid: false must emit a disabled grid attribute")
	}
}

func TestStateDiagramMovableTagsAndEditableRoutes(t *testing.T) {
	for _, want := range []string{
		"function stateAnchorHasDiagonalStub(info)",
		"function stateAnchorRouteDirection(info)",
		"stateAnchorHasDiagonalStub(sourceInfo)",
		"out.tagOffset = { x: Number(n.tagOffset.x), y: Number(n.tagOffset.y) }",
		"edgeRef.tagOffset = { x: tagDrag.x, y: tagDrag.y }",
		"state-tags-movable",
		"state-routes-editable",
		"function stateHandlesEnabled(root)",
		"stateHandlesEnabled(stateRoot) && flag(stateRoot, \"state-tags-movable\")",
		"stateSelfLoopGeometry(selfFromInfo, selfToInfo, stateLinkGap)",
		"function stateEdgeRoutingFailed(root, edgeIndex, edgeRef)",
		"Routing failed: the link endpoint was restored.",
		"Routing failed: the waypoint was restored.",
		"Editable routes",
		"sui-chart-settings-actions-row",
		"addSubAction(\"GRID\"",
		"addSubAction(\"LINKS\"",
		"openStateLinkSettings",
		"openStateGridSettings",
		"LINK settings",
		"GRID settings",
		"sui-chart-link-settings-dialog",
		"sui-chart-grid-settings-dialog",
		"routeKind === \"straight\"",
		"Manual waypoints are authoritative",
		"function stateSvgChartSize(svg)",
		"function stateFitAll(root)",
		"function stateFitWidth(root)",
		"__suiStateResizeObserver",
		"function stateDiagramBounds(root)",
		"position.setOffset = function(dx, dy)",
		"stateSvgChartSize(svg).w / Math.max(1, box.width)",
		"function statePathAcuteTurnIndex(points)",
		"function statePathHasShortSegment(points, minLength)",
		"function stateRemoveBacktracks(points, step, keepPoints)",
		"function stateStubEndpoint(port, direction, stubLength, step)",
		"function bindStatePan(root, svg)",
		"function stateResetPan(root)",
		"__suiStatePanX",
		"sui-chart-state-panning",
		"sui-chart-state-link-mode",
		"sui-chart-state-anchor-active",
		"fromPortPoint",
		"sui-chart-state-table-scroll",
		"sui-chart-state-triangle",
		"parallelogram",
		"trapezoid",
		"pentagon",
		"Rotation (deg)",
		"updateStateHistoryButtons",
		"__suiStateUndoBtn",
		"key === \"u\"",
		"data-sui-chart-state-rotate",
		"add(\"Delete\", function() {",
		"add(\"Duplicate\", function() { stateDuplicateSelection(root); });",
		"add(\"Tag settings...\"",
		"stateWorldPad",
		"__suiStateWorldPadX",
		"__suiStateRenderedLocalBounds",
		"stateDrawInternalTag",
		"statePathAnchorPoint",
		"sui-chart-state-tag-internal",
		"sui-chart-state-tag-anchor",
		"tagAnchorT",
		"tagArrowSize",
		"tagRotation",
		"stateShapeTagRotation",
		"Tag rotation",
		"stateShapeTopSideAnchor",
		"stateDefaultLinkTagAnchor",
		"statePolylineMinAngle",
		"__suiAutoTagAnchor",
		"__suiAutoTagT",
		"stateAnchorCandidateScore",
		"stateSimplifyWaypoints",
		"stateDedupeWaypoints",
		"Waypoint added. Drag its handle to shape the route.",
		"sui-chart-state-waypoint",
		"STATE_SOFT_COLORS",
		"stateSnapRotation",
		"stateShapeBorderColor",
		"stateSelectionColor",
		"highlightColor",
		"sui-chart-state-tag-rotate",
		"stateArrowDefault",
		"stateShapeTagArrowSize",
		"Tag settings",
		"sui-chart-state-tag-preview",
		"stateTagPreviewController",
		"stateTagPreviewCaption",
		"stateTagFieldSwitch",
		"stateTagFieldNumber",
		"stateTagFieldFont",
		"STATE_TAG_FONTS",
		"sui-chart-switch-track",
		"sui-chart-number-range",
		"sui-chart-tag-section",
		"sui-chart-tag-tabs",
		"sui-chart-tag-pane",
		"sui-chart-tag-tab", "stateLineText", "stateLineParts", "MAX_TAG_LINES", "stateLegacyLineText",
		`"tag", "hint"`,
		"sui-chart-tag-col-left",
		"sui-chart-tag-col-right",
		"line-start",
		"line-end",
		"Add waypoint here",
		"Text shadow",
		"Shadow color",
		"stateTagFieldText",
		"Tag text",
		"tagTextShadowColor",
		"stateTagTextShadowColor",
		"tagTextShadow",
		"stateTagTextShadow",
		"stateTagFontWeight",
		"stateTagFontStyle",
		"stateTagTextDecoration",
		"stateTagTextAlign",
		"stateTagLetterSpacing",
		"stateTagLineSpacing",
		"stateTagTextTransform",
		"stateTagTextColor",
		"stateTagBoxColor",
		"stateTagLeaderColor",
		"tagArrowShape",
		"stateTagFieldColor",
		"stateTagFieldSetDisabled",
		"stateClampTagOffsetInShape",
		"sui-chart-tag-mask",
		"linkMaskId",
		"sui-chart-color-text",
		"sui-chart-color-picker",
		"sui-chart-color-picker-alpha-text",
		"Box padding",
		"tagPadding",
		"stateTagPadding",
		"stateTagSectionSetDisabled",
		"tagBorderColor",
		"linkStyle",
		"linkWidth",
		"Drag the tag inside the shape",
		"stateTagLeaderPolyline",
		"stateTagLeaderBezierPath",
		"sui-chart-state-tag-preview-hint",
		"Drag the tag to move it",
		"if (!onGrid(points[0])) { skipGridPoint[0] = true; skipGridPoint[1] = true; }",
		"short segment ",
		"acute turn at ",
		"stateAxisEdgeCopy(edge, auditSourceInfo, auditTargetInfo)",
		"type !== \"state\" && type !== \"gantt\"",
		"stateSyncViewportSize",
		"root.__suiStateSizeWatch",
		"viewportW / zoom",
		"viewportH / zoom",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("state diagram runtime missing %q", want)
		}
	}
}

func TestStateSelfLoopCreationAndRoutingStability(t *testing.T) {
	if strings.Contains(suiChartRuntime, "Self-loop not allowed in quick creation") {
		t.Error("self-loops must be creatable from the canvas")
	}
	for _, want := range []string{
		"var selfLoop = fromId === toId",
		"stateEdgeRoutingFailed(stateRoot, edgeIndex, edgeRef)",
		"stateRouteNotice",
		"stateRoot.__suiRouteSignal",
		"A self-link always connects two distinct attachment points",
		"A self-link never uses the same attachment point twice",
		"function stateSelfLoopGeometry(fromInfo, toInfo, gap)",
		"stateSelfLoopGeometry(selfFromInfo, selfToInfo, stateLinkGap)",
		"function stateSelfOrthogonalPath(A, B, n1, n2, waypoints, geom, gap)",
		"function statePolylinePath(points)",
		"var selfKind = edgeRouting === \"bezier\" ? \"curved\" : edgeRouting",
		"makeControlHandle(curve1X, curve1Y, 1, \"bend1X\", \"bend1Y\")",
		"makeControlHandle(curve2X, curve2Y, 2, \"bend2X\", \"bend2Y\")",
		"\"bend1X\", \"bend1Y\"",
		"\"bend2X\", \"bend2Y\"",
		"function stateBoardFreeze(root)",
		"function stateBoardUnfreeze(root)",
		"function stateRevealGraphElement(root, selector)",
		"__suiStateFrozenExtent",
		"sui-chart-state-handles",
		"A plain click on empty canvas clears every selection",
		"function stateClearSelection(root)",
		"if (root.__suiStateSelected) stateClearSelection(root);",
		"group(\"Shape\")",
		"group(\"Position\")",
		"Border dash length",
		"syncBorderDash",
		"function stateShapePreview(getValues)",
		"shapePreview.render()",
		"function stateRoundedPolygonPath(vertices, radius)",
		"function statePolygonShapeElement(vertices, className, cornerRadius)",
		"function openStateShapeSettings(root)",
		"function openStateSettingsPanel(root)",
		"openStateSettingsPanel(root);",
		"function stateSelectedKeys(root)",
		"function stateSetSelection(root, keys, primary)",
		"function stateNudgeSelection(root, dx, dy)",
		"function stateCopySelection(root)",
		"function statePasteClipboard(root)",
		"function stateDuplicateSelection(root)",
		"function stateAlignSelection(root, mode)",
		"function stateAlignToolbarIcon(mode)",
		"function stateMagneticSnap(root, primaryNode, moveNodes, dx, dy, baseGeom)",
		"function stateEditingMode(root)",
		"if (!stateEditingMode(stateRoot)) return;",
		"function stateChartFireEvent(root, kind, item, part, eventName, ev)",
		"function stateChartHandlerAllowed(name)",
		"function stateChartElementHandler(root, item, kind, part, eventName)",
		"function stateChartBindStaticEvents(root, kind, item, element, part)",
		"element.__suiStaticSuppressClick = Date.now();",
		"root.__suiChartEvents",
		"queue.push(payload)",
		"payload.time = Date.now()",
		"function stateChartSetTag(name, id, text, boxColor, textColor)",
		"function stateChartUpdate(name, id, patch)",
		"function stateChartUpdateMany(name, updates)",
		"function stateChartGet(name, id)",
		"function stateChartElements(name)",
		"function stateChartFind(name, query)",
		"function stateChartGetAttr(name, id, field)",
		"function stateChartSetAttr(name, id, field, value)",
		"function stateChartGetLabel(name, id)",
		"function stateChartSetLabel(name, id, value)",
		"function stateChartGetText(name, id)",
		"function stateChartSetText(name, id, value)",
		"function stateChartGetHint(name, id)",
		"function stateChartSetHint(name, id, value)",
		"function stateChartGetKind(name, id)",
		"function stateChartOn(name, id, eventName, handler)",
		"function stateChartOff(name, id, eventName)",
		"function stateChartBindings(name)",
		"function stateChartPollEvents(name)",
		"function stateChartPeekEvents(name)",
		"function stateChartNextEvent(name)",
		"function stateChartClearEvents(name)",
		"function stateChartEventCount(name)",
		"function stateChartApplyVisual(root, data, targets)",
		"STATE_CHART_VISUAL_FIELDS",
		"__suiChartSuppressDraw",
		"setTag: stateChartSetTag",
		"__suiStateRefitSuppressUntil",
		"sameLocalBounds(measuredLocal, previousPreviousLocal)",
		"previousPadX >= minPadX - 0.01",
		"previousPadY >= minPadY - 0.01",
		"snappedDx = Math.round(worldDx / moveStep) * moveStep",
		"snappedDy = Math.round(worldDy / moveStep) * moveStep",
		"moveItem.node.x = moveItem.baseX + magnetic.dx",
		"var marqueeIntent = stateEditingMode(root) && (ev.button === 0 || ev.button === undefined);",
		"var panIntent = ev.button === 1 || ev.button === 2 || ev.altKey || !marqueeIntent;",
		"addSubAction(\"GRID\"",
		"addSubAction(\"SHAPES\"",
		"addSubAction(\"LINKS\"",
		"addSubAction(\"TAGS\"",
		"sui-chart-settings-actions-row",
		"sui-chart-shape-settings-row",
		"sui-chart-shape-settings-placeholder",
		"payload.settings = stateSettingsSnapshot(root);",
		"data.stateSettings = importedSettings;",
		"stateSettingsApply(root, next.stateSettings);",
		"Chart title",
		"Show anchors and handles",
		"Commit changes…",
		"Version history…",
		"Import state data…",
		"function stateShapeDefaults(root)",
		"shape-random-color",
		"add(\"Fill color\"",
		"function stateLinkDefaults(root)",
		"Link style inherits the chart's LINK configuration",
		"Reset style to defaults",
		"chosen === stateLinkDefaults(root).routing",
		"delete item.endMarker",
		"var dashPreview = function()",
		"function stateLinkArrowPreview(getValues)",
		"var linkArrowPreview = stateLinkArrowPreview",
		"data-sui-chart-dash-length",
		"state-start-marker",
		"syncDashEnabled",
		"stateTagFieldNumber(initialWidth, 0.5, 12, 0.5)",
		"group(\"Anchors\")",
		"linkDirectionReversed",
		"Dash length",
		"Dash gap",
		"Arrow size (0 = default)",
		"initialTagValues",
		"values.rotationMode === \"manual\" ? values.rotation : null",
		"var nodeAtPoint = function(chartX, chartY, clientX, clientY)",
		"edgeRef.from = activeNode.id",
		"edgeRef.to = activeNode.id",
		"A self-loop must attach at two distinct points",
		"Start marker",
		"state-start-arrow",
		"item.from = item.to; item.to = swapValue;",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("self-loop/routing stability runtime missing %q", want)
		}
	}
}

func TestStateDiagramHasNoRoleMarkers(t *testing.T) {
	for _, unwanted := range []string{
		"sui-chart-state-initial-marker",
		"sui-chart-state-initial-line",
		"sui-chart-state-final-ring",
		"sui-chart-state-final-core",
		"sui-chart-state-choice",
		`aria-roledescription": "State`,
	} {
		if strings.Contains(suiChartRuntime, unwanted) {
			t.Errorf("state chart runtime still contains role marker %q", unwanted)
		}
		if strings.Contains(suiCssBase, unwanted) {
			t.Errorf("state chart CSS still contains role marker %q", unwanted)
		}
	}
}

func TestStateLinkSelectionHighlightAndCircularEditing(t *testing.T) {
	for _, want := range []string{
		".sui-chart-state-edge.sui-chart-state-selected .sui-chart-state-arrow",
		".sui-chart-state-edge.sui-chart-state-selected .sui-chart-state-arrow-open",
		".sui-chart-link-preview-field",
		".sui-chart-link-preview-line",
		".sui-chart-direction-button",
		".sui-chart-direction-field",
		".sui-chart-shape-preview-box",
		".sui-chart-shape-settings-dialog",
		".sui-chart-state-settings-dialog",
		".sui-chart-settings-action",
		".sui-chart-state-align-toolbar",
		".sui-chart-state-align-icon",
		".sui-chart-state-align-icon-block",
		".sui-chart-state-align-icon-guide",
		".sui-chart-state-marquee",
		".sui-chart-state-align-guide",
		"align-items: start",
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("state link selection CSS missing %q", want)
		}
	}
	for _, want := range []string{
		"edgeRef.arcSweep = sweep",
		"out.arcSweep = t.arcSweep",
		"out.arcLarge = t.arcLarge",
		"lastTransition.arcSweep = arcSweepValue >= 0.5 ? 1 : 0",
		"lastTransition.arcLarge = arcLargeValue >= 0.5 ? 1 : 0",
		"var useLargeArc = isSelfLoop || pointerDistance > halfChord + 0.5",
		"edgeRef.arcLarge = useLargeArc ? 1 : 0",
		"var projection = (pointerPoint.x - midX) * normalX + (pointerPoint.y - midY) * normalY",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("circular link runtime missing %q", want)
		}
	}
}

func TestStateGraphAndTableShareOneModel(t *testing.T) {
	for _, want := range []string{
		"function refreshStateTable(root, data)",
		"refreshStateTable(stateRoot, data);",
		"data.stateNodes.splice(index, 1)",
		"data.stateTransitions = (data.stateTransitions || []).filter",
		"function stateAddNode(root, shape, x, y)",
		"function stateAddTransition(root)",
		"root.__suiStateTableTab = \"links\"",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("state graph/table sync runtime missing %q", want)
		}
	}
}

func TestChartZoomAndScrollRuntime(t *testing.T) {
	for _, want := range []string{
		"overflow === \"scroll\"", "90 + series.labels.length * pointWidth",
		"attr(root, \"point-size\")", "pointSize = Math.max(1, Math.min(24, pointSize))",
		"--sui-chart-tooltip-delay", "tooltipDelay",
		"isGantt ? Math.min(24, z * 1.5) : Math.min(4, z + 0.25)",
		"isGantt ? Math.max(0.5, z / 1.5) : Math.max(isState ? 0.05 : 0.5, z - 0.25)",
		"pxPerDay", "sui-chart-gantt-year", "dayNames",
		"overflow !== \"zoom\"", "baseW * zoom", "sui-chart-zoom-label",
		"action === \"fit\"",
	} {
		if !strings.Contains(suiChartRuntime, want) {
			t.Errorf("chart scroll/zoom runtime missing %q", want)
		}
	}
}
