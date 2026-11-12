# Split panes

A two-pane layout with a draggable divider. Each direct `pane` child becomes a
flex pane; the first pane starts at `initial` and both respect `min`. Dragging
the bar or focusing it and using the arrow keys resizes the panes. `Home` and
`End` jump to the minimum and maximum.

## Complete source

```sui
css ".split-page \\{ max-width: 900px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \\} .split-note \\{ color: #8fa0b8; font-size: 13px; \\} .split-panel \\{ height: 320px; \\}"

view main {
	title "Split panes"
	col class: "split-page" {
		h1 "Split panes"
		p "Two panes with a draggable divider. Focus the divider and use the arrow keys, Home or End." class: "split-note"
		splitter direction: "horizontal" initial: "38%" min: "120px" class: "split-panel" {
			pane {
				h3 "List"
				p "The first pane keeps its initial size."
				p "Drag the blue bar to resize."
			}
			pane {
				h3 "Detail"
				p "The second pane fills the remaining space."
				p "Both panes scroll independently."
			}
		}
		splitter direction: "vertical" initial: "50%" min: "60px" class: "split-panel" {
			pane {
				h3 "Top"
				p "Vertical splitter inside the same page."
			}
			pane {
				h3 "Bottom"
				p "Arrow Up/Down resize the panes."
			}
		}
	}
}
```

## Attributes

| Attribute | Meaning |
|---|---|
| `direction` | `horizontal` (default) or `vertical`. |
| `initial` | Initial size of the first pane, default `50%`. |
| `min` | Minimum size of either pane, default `120px`. |

A splitter requires exactly two `pane` children; `simpleui check` reports any
other count.

## Designer

Choose **Controls > Split panes** in the palette. The **Splitter** section sets
direction, initial size and minimum. Add or remove `pane` rows from the element
tree.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```
