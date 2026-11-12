# Command palette

A searchable command list that can stay inline or become an overlay. With `open` set, Ctrl/Cmd + `hotkey` opens it, Escape closes it, the arrow keys move and Enter invokes the active command.

## Complete source

```sui
state showPalette = false
state last = "none"

fn openPalette() -> null {
	showPalette = true
}

fn newProject() -> null {
	last = "new project"
}

fn openSettings() -> null {
	last = "settings"
}

fn help() -> null {
	last = "help"
}

css ".command-page \\{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \\} .command-note \\{ color: #8fa0b8; font-size: 13px; \\}"

view main {
	title "Command palette"
	col class: "command-page" {
		h1 "Command palette"
		p "Press Ctrl+K (or Cmd+K) to open the overlay, type to filter, use the arrow keys and Enter. Escape closes it." class: "command-note"
		button "Open palette" onClick: openPalette
		p "Last command: {last}" class: "command-note"
		commandbar "Search commands" open: showPalette hotkey: "k" {
			command "New project" shortcut: "Ctrl N" onClick: newProject
			command "Open settings" shortcut: "Ctrl ," onClick: openSettings
			command "Help" shortcut: "F1" onClick: help
		}
	}
}
```

## Designer

Choose **Overlay > Command palette** in the palette. The **Command palette** section sets the open state, hotkey and placeholder. Add `command` children with a shortcut and an `onClick` handler.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

