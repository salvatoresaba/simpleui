# Table layout

`tablelayout` is a **structural** HTML table. It places existing interface
elements in cells and is intentionally separate from the data `table` control:

| Need | Use |
|---|---|
| Arrange cards, text and controls in rows/columns | `tablelayout` |
| Show data, sort, filter, edit and persist rows | `table` with `name:` and `table.*` |

## What it demonstrates

- a real `<table>`/`<tbody>`/`<tr>`/`<td>` structure;
- a fixed `columns` count with children flowing left to right;
- cards, text, inputs and buttons inside separate cells;
- a `col` wrapper that keeps an input and a button in the same cell;
- reactive state changes through normal `input` and `button` bindings;
- `--sui-layout-gap` for spacing between cells.

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
		tablelayout columns: 2 class: "layout-demo" style: "--sui-layout-gap:18px" {
			card "Structural table" {
				h2 "Cells, not data rows"
				p "Each direct child of tablelayout fills one cell, left to right."
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
		}
		p "Wrap several elements in a col to keep them inside a single cell." class: "layout-note"
	}
}
```

## Use it

```sui
tablelayout columns: 3 {
  card "Left"
  card "Center"
  card "Right"
}
```

- `columns` accepts `1` to `6` (default `2`).
- Each direct child becomes one cell.
- Wrap multiple elements in `col`, `div` or `card` to keep them together.
- Set `--sui-layout-gap` with a static `style` or a project CSS class.
- `tablelayout` has no `name`, `editMode`, sorting, filtering or persisted data.

## Designer workflow

1. Open this folder's `app.sui` with `simpleui gui app.sui`.
2. Choose **Layout → Table layout** from the palette.
3. Drop cards or controls inside it. Each direct child fills one cell.
4. Select the layout table and use its **Table layout** inspector section to
   change the column count or append an empty cell container.
5. Use `col` wrappers if several controls must share one cell.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The ready-to-run `dist/` folder contains the desktop executable, the generated
web assets and a copy of this complete source as `dist/app.sui`.
