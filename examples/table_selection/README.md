# Table selection example

This example demonstrates Excel-style selection, editing and clipboard
operations on a 10×10 chart table. The chart and table are shown together, so
you can verify row, column and rectangular cell selection directly.

## Run it

```powershell
./run.ps1
```

or:

```bat
run.bat
```

The desktop build is in `dist/` and uses `assets` next to `app.exe`.

For the web version:

```bash
simpleui serve examples/table_selection/app.sui --watch
```

## What it demonstrates

- **Rows** selection: click the row number in the narrow left gutter or drag
  down the gutter to add rows.
- **Columns** selection: click the thin strip above a header or drag across
  those strips to add columns; Alt+click sorts the clicked column.
- **Column resizing**: drag the right edge of any header, double-click it to fit
  the widest cell content, or focus the resize handle and use Left/Right.
  Widths persist in `localStorage`; Columns → Reset column widths restores
  automatic sizing.
- **Cells** selection: click a cell and drag diagonally to draw a rectangular
  range; Shift+click extends from the previous cell.
- **None** selection mode.
- **Copy selection** to the clipboard as TSV, ready for Excel, Word or
  LibreOffice.
- **Paste selection** back into editable cells, rows or columns.
- **Export selection** as CSV, TSV, JSON, JSONL, XLSX or HTML.
- XLSX and HTML preserve styled headers and numeric alignment; formula-like values stay inert text.
- Export options for header rows, hidden columns and original/current row order.
- Print/Save PDF with the chart SVG and filtered table together.
- **Delete selected rows** from the Rows menu or the context menu.
- A delete button at the end of every row.
- Minus/plus buttons at the right edge of numeric cells, shown only after a click focuses the cell and hidden while dragging a selection.
- Starting a new selection cancels the pending cell edit and clears its focus.
- One shared selection color for rows, columns and cells, plus a 2px selection perimeter.
- Hover-only focus: the hovered row and cell get subtle neutral tones; clicking or dragging switches to selection and does not leave a focus highlight.
- Selected rows and cells keep the selection color; hover does not override them.
- Undo/redo for cell edits, steppers, paste, row insertion and row deletion with `Ctrl+Z`, `Ctrl+Shift+Z` / `Ctrl+Y`, the chart menu or the table context menu.
- In-place updates: editing a cell does not rebuild the table or flash the page, even when `onDataChange` updates the reactive state.
- Dark/Light themes with the same selection color and focus tones adapted to the background.
- Click a series color swatch in a table header or legend to change that series color; the override is stored per chart. Use `colorPicker: false` to keep swatches visible but disable editing.
- Row insertion with **+ Row**.
- Column filters, sorting, virtual scrolling and table settings.

## Complete SUI source used by the example

```sui
# Excel-style table selection example.
#
# Switch between row, column and cell selection. Drag to create a range,
# Ctrl/Cmd+click to add or remove, Shift+click to extend, right-click for the
# clipboard/export menu, and use the optional numeric steppers and delete
# buttons at the end of each row. Ctrl+Z / Ctrl+Shift+Z (or Ctrl+Y) undo and
# redo table edits, including steppers, paste, row insertion and deletion.

state data = "Name,Plan,Design,Build,Test,Review,Deploy,Support,Docs,Release\nA,12,7,3,5,4,2,6,8,1\nB,18,9,4,7,6,3,8,10,2\nC,15,8,5,6,5,2,7,9,3\nD,24,12,6,9,8,4,11,13,5\nE,21,10,4,8,7,3,9,12,4\nF,29,14,7,11,9,5,13,15,6\nG,26,13,6,10,8,4,12,14,5\nH,31,16,8,12,10,6,14,17,7\nI,35,18,9,14,12,7,16,19,8\nJ,40,21,10,16,13,8,18,22,9"
state selectionMode = "cell"
state themeMode = "dark"
state lastChange = "Drag diagonally across cells, use the row numbers on the left for rows, or the strip above headers for columns."

css ".selection-page \{ max-width: 1400px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 14px; \} .selection-controls \{ display: flex; flex-wrap: wrap; gap: 8px; align-items: center; \} .selection-label \{ min-width: 80px; color: #8fa0b8; font-size: 13px; \} .selection-status \{ color: #a5f3fc; font-size: 13px; \} .selection-table \{ width: 100%; height: 780px; \}"

fn setRows() -> null {
	selectionMode = "row"
	lastChange = "Row mode: click a row, drag across rows, or use Ctrl/Cmd/Shift."
}

fn setColumns() -> null {
	selectionMode = "column"
	lastChange = "Column mode: click or drag a header. Alt+click sorts the column."
}

fn setCells() -> null {
	selectionMode = "cell"
	lastChange = "Cell mode: click and drag to select a rectangular range."
}

fn setNone() -> null {
	selectionMode = "none"
	lastChange = "Selection disabled for the table."
}

fn setDarkTheme() -> null {
	themeMode = "dark"
	lastChange = "Dark chart theme."
}

fn setLightTheme() -> null {
	themeMode = "light"
	lastChange = "Light chart theme with the selection color preserved and focus adapted."
}

fn dataChanged(next: string) -> null {
	data = next
	lastChange = "Table data updated through onDataChange."
}

view main {
	title "Table selection"
	row class: "selection-page" {
		h1 "Excel-style table selection"
		p "10 rows × 10 columns. Select rows from the left gutter, columns from the strip above the headers, or drag diagonally across cells. Copy and paste from the context menu or the Rows menu." style: "margin:0;color:#8fa0b8"

		card {
			row class: "selection-controls" {
				span "Selection" class: "selection-label"
				button "Rows" onClick: setRows
				button "Columns" onClick: setColumns
				button "Cells" onClick: setCells
				button "None" onClick: setNone
				button "Dark" onClick: setDarkTheme
				button "Light" onClick: setLightTheme
			}
			p "Mode: {selectionMode}" class: "selection-status"
			p "{lastChange}" class: "selection-status"
		}

		chart "10x10 selection matrix" name: "selection_table" class: "selection-table" data: data chartType: "barstacked" theme: themeMode tableView: "both" showGrid: true showLegend: false tableFilters: true tableSelection: selectionMode tableAddRow: true tableDeleteRow: true tableSteppers: true tableContextMenu: true interactive: true onDataChange: dataChanged style: cssSafe("--sui-chart-series-1: #38bdf8; --sui-chart-series-2: #a855f7;")

		card {
			p "How to try it:" style: "margin:0 0 6px;font-weight:600;color:#e2e8f0"
			p "Rows: click the row number on the left or drag down the row numbers to add rows."
			p "Columns: click the narrow strip above a header or drag across the strips; Alt+click sorts."
			p "Cells: click a cell and drag diagonally to draw a rectangular selection. Shift+click extends."
			p "Right-click for Copy, Paste, Export selection CSV/TSV/JSON/JSONL/XLSX/HTML, delete or selection actions. Use table settings for export options and Print / Save PDF."
			p "The + Row button appends a row; numeric cells use the visible −/+ buttons; every row has a delete button."
		}
	}
}
```

The four buttons update the reactive `selectionMode` state, so the selection
behavior changes immediately without editing the source.

## Selection gestures

| Gesture | Result |
|---|---|
| Click row number | Select the whole row. |
| Drag row numbers | Add contiguous rows to the selection. |
| Click top column strip | Select the whole column. |
| Drag top column strips | Add contiguous columns to the selection. |
| Click and drag in cells | Draw a rectangular cell selection. |
| Shift+click | Extend from the previous anchor. |
| Ctrl/Cmd+click | Add or remove from the selection. |
| Ctrl/Cmd+Space | Toggle the focused row in row mode. |
| Delete key | Delete selected rows when the table is editable. |
| Right-click | Open Copy, Paste, Export selection, Delete and selection actions. |

The row gutter and column strip work in every selection mode, so rows, columns
and cell blocks can be selected directly. Table filters remain available from the
filter buttons in each header.

## Clipboard permissions

Clipboard access is requested by the browser when Copy or Paste is used. The
result is reported in the table status line. Copy falls back to a hidden
textarea if the Clipboard API is unavailable. Paste requires browser read
permission; if it is denied, the status line explains how to retry.

The complete source is shown above, is saved as `app.sui`, and is copied to
`dist/app.sui`.



