# Table layout

`tablelayout` is a **structural** HTML table. It arranges existing interface
elements in an explicit row/column grid and is intentionally separate from the
data `table` control:

| Need | Use |
|---|---|
| Arrange cards, text and controls in rows/columns | `tablelayout` |
| Merge cells, set column/row sizes, gaps and borders | `tablelayout` settings |
| Show data, sort, filter, edit and persist rows | `table` with `name:` and `table.*` |

## What it demonstrates

- an explicit `3 x 2` grid with a merged full-width header (`colspan: 2`);
- column widths in `%` and row heights in `auto`/`px`;
- separate horizontal and vertical gaps;
- dashed cell borders with custom width and color;
- per-cell padding and alignment on the result card;
- responsive stacking below 640px and a semantic table with `aria-label`;
- cards, text, inputs and buttons inside cells;
- a `col` wrapper that keeps an input and a button in the same cell;
- reactive state changes through normal `input` and `button` bindings.

## Complete source

```sui
# Table layout is a structural container, not the data table control.
# It renders a real <table> and places each direct child in its own cell.
# Use `tablelayout` to arrange interface regions in rows/columns; use `table`
# when you need a data grid with table.display()/table.importText().

state name = ""
state clicks = 0

css ".layout-page \{ max-width: 980px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 14px; \} .layout-demo \{ width: 100%; \} .layout-stack \{ display: flex; flex-direction: column; gap: 10px; padding: 4px; \} .layout-note \{ color: #8fa0b8; font-size: 13px; \} .layout-cell-pad \{ padding: 10px; \}"

fn greet() -> null {
	if trim(name) == "" {
		name = "there"
	}
	clicks = clicks + 1
}

view main {
	title "Table layout"
	col class: "layout-page" {
		h1 "Table layout"
		p "A structural HTML table that hosts cards, text and controls. It is not a data table."
		tablelayout rows: 3 columns: 2 width: "100%" columnWidths: "55%, 45%" rowHeights: "auto, auto, 90px" gapX: "18px" gapY: "14px" borderStyle: "dashed" borderWidth: "1px" borderColor: "#475569" responsive: true semantic: true label: "Table layout demo" class: "layout-demo" {
			layoutcell row: 0 col: 0 colspan: 2 {
				card "Structural table" {
					h2 "Cells, not data rows"
					p "Rows, columns, merged cells and spacing are configured in the Table layout inspector."
				}
			}
			layoutcell row: 1 col: 0 {
				card "Host any control" {
					h2 "Composable"
					p "Buttons, inputs and other layout containers keep their normal behavior."
				}
			}
			layoutcell row: 1 col: 1 {
				col class: "layout-stack" {
					input "Your name" value: name placeholder: "Type a name"
					button "Say hello" onClick: greet
				}
			}
			layoutcell row: 2 col: 0 class: "layout-cell-pad" {
				card "Result" {
					h2 "Interactive"
					p "Hello {name}."
					p "Button clicks: {clicks}"
				}
			}
			layoutcell row: 2 col: 1 {
				p "Wrap several elements in a col to keep them inside a single cell." class: "layout-note"
			}
		}
	}
}

```

## Use it

```sui
tablelayout rows: 2 columns: 3
  width: "100%"
  columnWidths: "200px, 40%, auto"
  rowHeights: "80px, auto"
  gapX: "16px" gapY: "12px"
  borderStyle: "solid" borderWidth: "1px" borderColor: "#475569"
  responsive: true semantic: true label: "Project dashboard" {
  card "A" colspan: 2
  card "B"
  card "C" rowspan: 2
  card "D"
}
```

- `rows` accepts `1`-`50`, `columns` accepts `1`-`12`.
- If `rows` is omitted it is derived from the children. Explicit rows are a
  minimum; extra children expand the grid.
- Each direct child fills the next free cell, row by row.
- `colspan` / `rowspan` merge a cell right/down. Spans are clamped to the
  remaining grid.
- `columnWidths` / `rowHeights` use one size per column/row: `px`, `%`, `em`,
  `rem`, `vw`, `vh`, `pt` or `auto`.
- `gapX` / `gapY` set column/row spacing.
- `borderStyle` (`none`, `solid`, `dashed`, `dotted`, `double`), `borderWidth`
  and `borderColor` style every cell.
- Per-cell `cellPadding`, `cellBackground`, `cellBorderStyle`, `cellBorderWidth`
  and `cellBorderColor` override the cell. Align content by styling the child
  element.
- `cellMinHeight` keeps empty cells visible and droppable (default `72px`).
- `responsive: true` stacks cells on screens up to 640px.
- `semantic: true` uses `role="table"`; `label` sets `aria-label`.
- Wrap multiple elements in `col`, `div` or `card` to share a cell.
- Designer-generated layouts use explicit `layoutcell` wrappers with `row`,
  `col`, `rowspan` and `colspan`; direct children remain supported.
- `tablelayout` has no `name`, `editMode`, sorting, filtering or persisted data.

## Designer workflow

1. Open this folder's `app.sui` with `simpleui gui app.sui`.
2. Choose **Layout > Table layout** from the palette.
3. Select the layout table and use the **Table layout** inspector, including
   **Split vertical** / **Split horizontal** and **Merge selected**:
   - preset (shown only while the table is empty);
   - rows and columns;
   - **Structure table settings** for per-column widths and per-row heights;
     editing a size beyond the declared rows/columns raises the counters;
   - separate column/row gaps;
   - empty-cell height;
   - responsive stacking;
   - semantic role and aria label.
   The table width/height are in the generic **Size** section and the table
   border in the generic **Border** section. **Structure table settings** opens
   a panel with a scrollable column list and row list, so large grids stay
   usable. Reducing rows/columns when cells contain elements asks whether to
   remove that content or move it into the nearest remaining cell; Cancel keeps
   the grid unchanged.
4. Drag cards or controls inside it. An empty cell can be clicked directly, so
   you can see its row/column; the next palette or Insert element goes inside
   the selected cell. **Alt+click** an occupied cell or use its always-visible
   bottom-left triangle to select the cell itself; clicking a child selects the
   child.
5. Drag the thin split handles in the preview to resize columns/rows; the sizes
   are shown and editable in **Structure table settings**. Click a split handle
   without dragging to select it; Ctrl+click adds more splits and **Delete
   split** merges the adjacent cells. Ctrl+click cells to add them to a
   multi-cell split or merge. If a split would make a column narrower than 40px
   or a row shorter than 24px, a dialog asks whether to continue. Select a cell and use
   **Layout cell** to set `column span` / `row span`; padding, background,
   border and effects are edited in the generic inspector sections, while
   inner elements use their own **Alignment** section. **Reset to 1 x 1**
   splits the cell again. Growing a span merges the fully covered cells and
   keeps their content; a partial overlap is rejected. Shrinking a span keeps
   the content in the cell and recreates the released area as empty cells.
   With a cell selected, Position, Size, Responsive, Layout, Grid and margin are
   hidden; padding, background, border, effects and Drag & drop stay available.
   With several cells selected, the generic CSS sections apply to all of them.
   The table frame is edited in **Border**; the global cell border attributes
   remain source-level.
6. Cells are outlined in the designer preview, including empty and merged cells,
   so the grid is visible even when application borders are set to none.
   Responsive stacking is disabled in the preview; the compiled app still
   stacks below 640px.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The ready-to-run `dist/` folder contains the desktop executable, the generated
web assets and a copy of this complete source as `dist/app.sui`.
