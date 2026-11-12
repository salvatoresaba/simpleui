# Empty state

A centered placeholder with an icon, a title, a hint and optional actions. It is usually shown when a collection is empty.

## Complete source

```sui
state count = 0

fn addProject() -> null {
	count = count + 1
}

css ".empty-page \{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \} .empty-note \{ color: #8fa0b8; font-size: 13px; \}"

view main {
	title "Empty state"
	col class: "empty-page" {
		h1 "Empty state"
		p "A centered placeholder with an icon, a title, a hint and an action." class: "empty-note"
		if count == 0 {
			emptystate "No projects" icon: "▢" hint: "Create the first project to get started." {
				button "New project" onClick: addProject
			}
		} else {
			alert "Workspace ready"
			p "Projects: {count}" class: "empty-note"
			button "Add another project" onClick: addProject
		}
	}
}
```

## Designer

Choose **Content > Empty state** in the palette. Set the icon, hint and compact mode in the **Empty state** section; child elements become the body/actions.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

