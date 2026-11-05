# Toast example

This is the complete, runnable demonstration of the toast notification control.
It covers the API, all variants, timeouts, persistent toasts, clearing, the
optional host element, accessibility and safety.

## Run it

A desktop build is included in `dist/`:

- **Windows:** double-click `examples/toast/dist/app.exe`.
- Keep the `assets` folder next to `app.exe`.

To rebuild and launch it from this folder:

```powershell
./run.ps1
```

or, from `cmd.exe`:

```bat
run.bat
```

The scripts build `../toast.sui` with:

```bash
simpleui build examples/toast.sui --target desktop --out examples/toast/dist
```

A single-file web build can be produced with:

```bash
simpleui build examples/toast.sui --single --out examples/toast/dist/toast.html
```

## Complete SUI source

The complete runnable source is `../toast.sui` and is also copied to `dist/app.sui`.

```sui
# Complete toast example.
# toast.show(message [, kind [, duration]]) adds text-only notifications.
# kind: info (default) | success | warning | error
# duration: milliseconds; 0 keeps the toast open until dismissed.
# toast.clear() removes all visible toasts.

view main {
	title "Toast - complete example"
	h1 "Toast notifications"
	p "Click the buttons to show short feedback messages. Each toast can be closed with its X button."
	p "The host below is a normal designer element. Without a toast host, the runtime creates a fixed top-right host automatically."

	row {
		button "Info (default)" onClick: showInfo
		button "Success (2s)" onClick: showSuccess
		button "Warning (6s)" onClick: showWarning
		button "Persistent error" onClick: showError
		button "Clear all" onClick: clearToasts
	}

	toast class: "demo-toast-host"

	accordion "How this example works" {
		accordionitem "API and defaults" {
			p "toast.show(text) uses the info kind and a 4000 ms timeout."
			p "toast.show(text, kind) accepts info, success, warning or error."
			p "toast.show(text, kind, duration) uses a custom duration; 0 keeps the toast open."
			p "toast.clear() removes every visible toast."
		}
		accordionitem "Host and positioning" {
			p "The toast host is optional. Drag Toast host from the Overlay palette to position notifications in a custom place."
			p "Select the host and use the Toast panel: choose horizontal and vertical anchors; offset fields appear only for edge-anchored axes. Size and Border in the inspector style the toast card."
			p "The runtime creates a fixed host if the page has no toast host."
		}
		accordionitem "Safety and accessibility" {
			p "Messages are assigned with textContent, so they cannot inject HTML."
			p "Kinds are restricted to the four supported values. Unknown kinds fall back to info."
			p "Toasts use role=status and the host uses aria-live=polite."
		}
	}
}

fn showInfo() -> null {
	toast.show("This is an info toast.")
}

fn showSuccess() -> null {
	toast.show("Saved successfully.", "success", 2000)
}

fn showWarning() -> null {
	toast.show("Please review this value.", "warning", 6000)
}

fn showError() -> null {
	toast.show("This error stays until you close it.", "error", 0)
}

fn clearToasts() -> null {
	toast.clear()
}
```

## What the example demonstrates


- `toast.show(message)` — default `info` kind and 4000 ms timeout;
- `toast.show(message, kind)` — one of `info`, `success`, `warning`, `error`;
- `toast.show(message, kind, duration)` — custom timeout in milliseconds;
- `duration: 0` — toast stays open until the user closes it;
- `toast.clear()` — dismiss every visible toast;
- automatic dismissal and manual X-button dismissal;
- at most four visible toasts per host; the oldest is removed;
- the optional `toast` host element and custom CSS positioning;
- the runtime-created fixed top-right host when no host exists;
- `aria-live="polite"`, `role="status"` and text-only messages.

## API

```sui
toast.show("Saved", "success", 3000)
toast.show("This stays open", "warning", 0)
toast.clear()
```

| Argument | Meaning |
|---|---|
| `message` | Text shown in the notification |
| `kind` | `info` (default), `success`, `warning`, `error` |
| `duration` | Milliseconds; default `4000`; `0` means no automatic close |

Unknown kinds fall back to `info`. Messages are assigned with `textContent`, so
they cannot inject HTML. The runtime keeps at most four toasts in a host.

## Host and positioning

The host is optional:

```sui
toast class: "my-toast-host"
```

Without it, the runtime creates a fixed top-right host on first use. A host
placed in a view is an empty live-region element.

The designer writes independent horizontal and vertical anchors:

```sui
toast class: "my-toast-host" hAlign: "left" vAlign: "bottom" hOffset: "24px" vOffset: "16px"
```

`hAlign` is `left`, `center` or `right`; `vAlign` is `top`, `center` or
`bottom`. An edge-anchored axis uses its `hOffset`/`vOffset` distance from that
edge. A centered axis has no offset field and ignores any offset value. The
generated CSS classes are `sui-toast-h-*` and `sui-toast-v-*`.

When several messages are visible, a top anchor adds each new toast below the
previous one, while a bottom anchor adds each new toast above the previous one.

## Designer panel and live preview

Insert **Toast host** under **Overlay**. Its tree label is `toast host`; the Text
field is not used. Add handlers to buttons with `onClick`, then call
`toast.show(...)` or `toast.clear()` in the function.

Selecting the host opens a dedicated **Toast** inspector panel:

| Section | Controls |
|---|---|
| Placement | horizontal left/center/right and vertical top/center/bottom |
| Offsets | horizontal offset only for left/right; vertical offset only for top/bottom |
| Behaviour | default duration, maximum visible, show icons |
| Preview | preview kind, preview message, Show preview, Clear preview |

The existing **Size**, **Border**, **Box model**, **Typography**,
**Background** and **Effects** sections below the panel write the toast card
rule. Width and height therefore use the normal **Size** fields, and border,
background, radius and typography fields visibly apply to the toast card.

The same declarations appear in the host **style expr** field. That field is the
editable CSS declaration list for `.sui-toast`, for example
`width: 420px; border-width: 3px;`. Editing it updates the card, and custom
properties such as `--sui-toast-icon-bg` are inherited by the icon, message and
close button.

The preview iframe shows a sample toast with the real app CSS, so placement,
size, border and style expression changes are visible while editing. The
runtime also supports the CSS variables `--sui-toast-accent`,
`--sui-toast-icon-bg`, `--sui-toast-icon-fg`, `--sui-toast-message-fg` and
`--sui-toast-close-fg` for further presets. The `toast` element has no source
children or visible label.

## Security

Messages use `textContent`, never `raw HTML`. The runtime restricts kinds to the
four known values and only toggles host/toast classes, text content and
attributes. No new file, network or execution capability is introduced.

## Related documentation

- [API reference: toast](../../docs/API_REFERENCE.md#toast-notifications)
- [GUI editor: toast control](../../docs/GUI_EDITOR.md#toast-control)
- [Italian README](../../README_IT.md)
