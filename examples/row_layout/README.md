# Row and column layout

A `row` lays its children out side by side (they are its **columns**) and a
`col` stacks them (they are its **rows**). A container that declares its
tracks describes the size of every cell, so the layout is part of the source
and not only of the CSS.

| Need | Use |
|---|---|
| Arrange elements in a row or in a column | `row` / `col` |
| Declare the number of cells and their size | `columns` / `columnWidths`, `rows` / `rowHeights` |
| Arrange a grid with merged cells and borders | [`tablelayout`](../table_layout/README.md) |
| Show data, sort, filter, edit and persist rows | `table` with `name:` and `table.*` |

## What it demonstrates

- a row with **three columns** of different kinds: a fixed `220px` column, a
  column of `40%` and an **automatic** column that takes the space the other
  two leave free;
- a column container with two rows: a fixed `90px` row and an automatic one;
- the size of the containers in the CSS of their class (`width: 100%`,
  `height: 180px`), like for every other element;
- a conditional block (`if` / `else`) that changes the number of columns of a
  row when a state variable changes;
- the automatic cell written as an empty value in the source and as `auto` when
  the designer regenerates it.

## Complete source

```sui
# Row and column layout: a container declares the number of its cells and the
# size of every cell. A cell without a size takes the space that the other
# cells leave free, so a fixed cell and an automatic one always fill the row.
#
# The size of the container is not an attribute: it belongs to the CSS of the
# element, like the size of every other element, and the designer sets it in
# the Size section of the inspector.
#
# The designer exposes the tracks in the Row layout / Column layout section of
# the inspector: select the row, change the number of columns or the size of a
# single cell, and drag the boundary between two cells in the preview.

state wide = true

fn toggle() -> null {
	wide = not wide
}

css ".rl-page \{ max-width: 900px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 16px; \} .rl-box \{ border: 1px solid #2b3242; border-radius: 10px; padding: 14px; \} .rl-full \{ width: 100%; \} .rl-row \{ height: 180px; \} .rl-stack \{ width: 320px; \} .rl-title \{ font-weight: 600; \} .rl-note \{ color: #8fa0b8; font-size: 13px; \}"

view main {
	title "Row and column layout"
	col class: "rl-page" {
		h1 "Row and column layout"
		p "A row lays its children out as columns, a column container stacks them as rows. The size of every cell lives in the container."

		row columns: "3" columnWidths: "220px, 40%, " class: "rl-box rl-full rl-row" {
			col class: "rl-box" {
				p "Fixed" class: "rl-title"
				p "220px" class: "rl-note"
			}
			col class: "rl-box" {
				p "Share" class: "rl-title"
				p "40% of the row" class: "rl-note"
			}
			col class: "rl-box" {
				p "Automatic" class: "rl-title"
				p "takes the space the other cells leave free" class: "rl-note"
			}
		}

		col rows: "2" rowHeights: "90px, " class: "rl-box rl-stack" {
			div {
				p "Top" class: "rl-title"
			}
			div {
				p "The second row takes the rest of the column." class: "rl-note"
			}
		}

		if wide {
			row columns: "2" columnWidths: "50%, 50%" class: "rl-box rl-full" {
				div {
					p "Two equal columns."
				}
				div {
					p "The then branch of the conditional block."
				}
			}
		} else {
			row columns: "1" columnWidths: "100%" class: "rl-box rl-full" {
				div {
					p "One single column: the else branch."
				}
			}
		}

		row columns: "1" class: "rl-box rl-full" {
			div {
				button "Toggle the layout" onClick: toggle
			}
		}
	}
}
```

## Build and run

```bash
simpleui build examples/row_layout/app.sui --out dist
simpleui serve examples/row_layout/app.sui --watch
```

## Edit it in the designer

```bash
simpleui gui examples/row_layout/app.sui
```

- Select a `row` or a `col` in the tree: the **Row layout** / **Column layout**
  section of the inspector shows the number of columns (or rows) and the size
  of every cell. The size of the container is in the **Size** section, like
  for every other element.
- Change the unit of a cell between `px`, `%` and **auto**. An automatic cell
  takes the space that the other cells leave free.
- The percentages together cannot exceed `100%`: a value that does not fit is
  refused and falls back to the largest value that fits.
- **Drag the boundary between two cells in the preview** to resize them: the
  dragged cell gets a fixed size and the cell on its right becomes automatic.
- **Alt+click** inside a cell selects the cell itself, even when it holds
  other elements.
- The right-click menu of a `row` or of a `col` offers **Add column** /
  **Add row**, which adds an empty cell.

## Limitations

- The sizes are literal values: they cannot be computed from a state variable.
- A cell without a size is automatic, but a container without tracks keeps the
  plain flex layout: the tracks only exist when the container declares them.
