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

- an explicit `rows` x `columns` grid;
- a merged full-width header with `colspan: 2`;
- column widths in `%` and row heights in `auto`/`px`;
- separate horizontal and vertical gaps;
- dashed cell borders with custom width and color;
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

css ".layout-page \{ max-width: 980px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 14px; \} .layout-demo \{ width: 100%; \} .layout-stack \{ display: flex; flex-direction: column; gap: 10px; padding: 4px; \} .layout-note \{ color: #8fa0b8; font-size: 13px; \}"

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
		tablelayout rows: 3 columns: 2 width: "100%" columnWidths: "55%, 45%" rowHeights: "auto, auto, 90px" gapX: "18px" gapY: "14px" borderStyle: "dashed" borderWidth: "1px" borderColor: "#475569" class: "layout-demo" {
			card "Structural table" colspan: 2 {
				h2 "Cells, not data rows"
				p "Rows, columns, merged cells and spacing are configured in the Table layout inspector."
			}
			card "Host any control" {
				h2 "Composable"
				p "Buttons, inputs and other layout containers keep their normal behavior."
			}
			col class: "layout-stack" {
				input "Your name" value: name placeholder: "Type a name"
				button "Say hello" onClick: greet
			}
			card "Result" {
				h2 "Interactive"
				p "Hello {name}."
				p "Button clicks: {clicks}"
			}
			p "Wrap several elements in a col to keep them inside a single cell." class: "layout-note"
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
  borderStyle: "solid" borderWidth: "1px" borderColor: "#475569" {
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
- `columnWidths` / `rowHeights` use comma-separated `px`, `%`, `em`, `rem`,
  `vw`, `vh`, `pt` or `auto`.
- `gapX` / `gapY` set column/row spacing.
- `borderStyle` (`none`, `solid`, `dashed`, `dotted`, `double`), `borderWidth`
  and `borderColor` style every cell.
- `cellMinHeight` keeps empty cells visible and droppable (default `72px`).
- Wrap multiple elements in `col`, `div` or `card` to share a cell.
- `tablelayout` has no `name`, `editMode`, sorting, filtering or persisted data.

## Designer workflow

1. Open this folder's `app.sui` with `simpleui gui app.sui`.
2. Choose **Layout > Table layout** from the palette.
3. Select the layout table and set **rows**, **columns**, table/column/row sizes,
   gaps, borders and empty-cell height in the **Table layout** inspector.
4. Drag cards or controls inside it. Each direct child fills the next cell.
5. Select a direct child and use **Layout cell** to set `column span` /
   `row span`; **Reset to 1 x 1** splits the cell again.
6. Cells are outlined in the designer preview, including empty and merged cells,
   so the grid is visible even when application borders are set to none.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The ready-to-run `dist/` folder contains the desktop executable, the generated
web assets and a copy of this complete source as `dist/app.sui`.
