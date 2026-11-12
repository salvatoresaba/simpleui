# Dropdown menu

A compact button with an anchored action menu. It reuses the menu runtime, so
outside clicks, Escape, keyboard focus and item actions behave like the menu bar.

## Complete source

```sui
state action = "none"

fn editAction() -> null {
	action = "edit"
}

fn duplicateAction() -> null {
	action = "duplicate"
}

fn deleteAction() -> null {
	action = "delete"
}

css ".dropdown-page \{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \} .dropdown-note \{ color: #8fa0b8; font-size: 13px; \}"

view main {
	title "Dropdown menu"
	col class: "dropdown-page" {
		h1 "Dropdown menu"
		p "A compact button with an anchored action menu." class: "dropdown-note"
		dropdown "Actions" position: "right" {
			dropdownitem "Edit" onClick: editAction
			dropdownitem "Duplicate" onClick: duplicateAction
			dropdownitem separator: true
			dropdownitem "Delete" onClick: deleteAction
		}
		p "Last action: {action}" class: "dropdown-note"
	}
}
```

## Designer

Choose **Controls > Dropdown menu** in the palette. Set the button label from the
text field, choose **left** or **right** in the **Dropdown** section, then add
`dropdownitem` rows from the element tree. Each item can be disabled or turned
into a separator.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```
