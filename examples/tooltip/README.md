# Tooltip property

Tooltip is attached to existing elements and controls. The complete demo shows
buttons with tooltips, dynamic tooltip text, hover and keyboard focus, and the
global settings panel.

## Run it

From the repository root:

```powershell
.\simpleui.exe serve examples\tooltip.sui --watch
```

Or build a desktop demo:

```powershell
.\simpleui.exe build examples\tooltip.sui --target desktop --out examples\tooltip\dist
.\examples\tooltip\dist\app.exe
```

A ready build is in `examples/tooltip/dist/`; `dist/app.sui` is the complete
source used for that build.

## Complete source

```sui
# Tooltip property example.
# Any element can carry tooltip: "Text".
# Global placement and timing are configured from Preferences > Tooltip settings.

state count = 0

fn noop() -> null {
}

fn increment() -> null {
	count += 1
}

view main {
	title "Tooltip - complete example"
	h1 "Tooltip"
	p "Hover a control or focus it with the keyboard. Press Escape to dismiss the bubble."

	row {
		button "Save" tooltip: "Saves the current document" onClick: noop
		button "Delete" tooltip: "Delete permanently" onClick: noop
		button "Focus me" tooltip: "Keyboard focus works too" onClick: noop
		button "Count: {str(count)}" tooltip: "Clicked {str(count)} times" onClick: increment
	}

	card {
		p "The tooltip property can be attached to a button inside any layout."
		button "Card action" tooltip: "Tooltip on a button inside a card" onClick: noop
	}

	accordion "How this example works" {
		accordionitem "Tooltip property" {
			p "Select any element and open the Tooltip section in the inspector."
			p "The text field is the only per-element setting; appearance and timing are global."
		}
		accordionitem "Global settings" {
			p "Open Preferences > Tooltip settings to choose place, delay and bubble colors."
			p "The settings are written as CSS custom properties in the project."
		}
		accordionitem "Safety and accessibility" {
			p "Tooltip text is rendered as text, never as raw HTML."
			p "The target receives aria-describedby while the bubble is open."
		}
	}
}
```

## Designer workflow

1. Select an element or control in the tree or preview.
2. Open the collapsible **Tooltip** section in the inspector.
3. Enter the tooltip text. Clear the field to remove the tooltip.
4. Open **Preferences > Tooltip settings** for global place, delay, colors,
   radius and maximum width.

The global settings are written to `:root` as `--sui-tooltip-*` custom
properties; they are part of the project CSS and apply to every tooltip.

## API

```sui
button "Save" tooltip: "Saves the current document" onClick: save
```

- The property is available on every allowed element.
- The text is escaped and never treated as HTML.
- Hovering or focusing the element opens the bubble.
- Leaving/blurring hides it; `Escape` dismisses the active bubble.
- The target receives `aria-describedby`; the bubble uses `role="tooltip"`.
- Void controls such as `input` are supported because the bubble is created as a
  fixed-position portal by the runtime.
- For keyboard access, use a focusable target.

## Limitations

- A target that cannot receive focus has no keyboard activation.
- The runtime clamps the bubble inside the viewport, but an extremely narrow
  viewport can still make a large bubble difficult to read.
- Global placement is a default; very unusual layouts may need custom CSS.

See [GUI editor: tooltip property](../../docs/GUI_EDITOR.md#tooltip-property) and
[API reference: tooltip property](../../docs/API_REFERENCE.md#tooltip-property).
