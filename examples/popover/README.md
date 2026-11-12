# Popover

An anchored panel with rich content. Click the trigger to toggle it; click
outside or press Escape to close it.

## Complete source

```sui
state clicked = 0

fn bump() -> null {
	clicked = clicked + 1
}

css ".popover-page \{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \} .popover-note \{ color: #8fa0b8; font-size: 13px; \}"

view main {
	title "Popover"
	col class: "popover-page" {
		h1 "Popover"
		p "An anchored panel with rich content. Click outside or press Esc to close." class: "popover-note"
		popover "Details" position: "bottom" {
			h3 "Anchored panel"
			p "The panel can host text, cards, inputs and buttons."
			p "Clicks: {clicked}" class: "popover-note"
		}
		button "Increment counter" onClick: bump
	}
}
```

## Designer

Choose **Controls > Popover** in the palette. The element text is the trigger
label; child elements become the panel content. Set the side in the **Popover**
section.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```
