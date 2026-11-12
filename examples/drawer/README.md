# Drawer

A side panel whose visibility is controlled by a state. The drawer renders a
backdrop and an aside panel; the backdrop, the close button and Escape all set
the bound state to `false` (or call `onClose` when `open` is not a state name).

## Complete source

```sui
state showFilters = false

fn openDrawer() -> null {
	showFilters = true
}

fn closeDrawer() -> null {
	showFilters = false
}

css ".drawer-page \\{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \\} .drawer-note \\{ color: #8fa0b8; font-size: 13px; \\}"

view main {
	title "Drawer"
	col class: "drawer-page" {
		h1 "Drawer"
		p "A side panel controlled by a state. The backdrop, the close button and Escape close it." class: "drawer-note"
		button "Open filters" onClick: openDrawer
		p "Open: {showFilters}" class: "drawer-note"
		drawer "Filters" side: "right" width: "320px" open: showFilters {
			h3 "Filters"
			p "Choose a status and a date range."
			button "Apply" onClick: closeDrawer
		}
	}
}
```

## Attributes

| Attribute | Meaning |
|---|---|
| `side` | `right` (default) or `left`. |
| `width` | CSS width of the panel, default `320px`. |
| `open` | State or bool expression. A state binding is closed through its setter. |
| `onClose` | Function called when `open` is not a plain state name. |

The element text is the drawer heading; children become the scrollable body.

## Designer

Choose **Controls > Drawer** in the palette. Set the title from **Text** and use
the **Drawer** section for side, width, the `open` state and `onClose`. The
inspector creates the `open` state when it is still missing.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```
