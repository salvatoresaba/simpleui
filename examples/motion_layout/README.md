# FLIP + enter/exit motion with stable keys

A filterable and reorderable card list. Cards are matched across renders by
motionKey, so the runtime can:

- FLIP cards to their new positions when the list is filtered or reordered;
- fade out a card that disappears while the others move;
- fade in a card that appears;
- keep the same card identity across different if branches.

## Complete source

```sui
# FLIP + enter/exit motion with stable keys.
#
# Cards are matched across renders by motionKey. When the filter hides a card,
# the card fades out as a fixed clone while the remaining cards FLIP to their new
# positions. Reordering flips all cards. New cards fade in.

state filter = "all"
state order = "normal"

fn showAll() -> null {
	filter = "all"
}

fn showDesign() -> null {
	filter = "design"
}

fn showDev() -> null {
	filter = "development"
}

fn toggleOrder() -> null {
	if order == "normal" {
		order = "reversed"
	} else {
		order = "normal"
	}
}

view projectCard(name: string, kind: string) {
	card class: "motion-card" motion: "layout" motionKey: name motionDuration: "420ms" motionEasing: "ease-out" {
		row class: "motion-card-head" {
			h3 "{name}"
			badge "{kind}"
		}
		p "Card {name}: this content moves with the same key across filters and reordering." class: "motion-meta"
	}
}

css ".motion-page \{ max-width: 780px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 16px; \} .motion-note \{ color: #8fa0b8; font-size: 13px; \} .motion-toolbar \{ display: flex; gap: 8px; flex-wrap: wrap; align-items: center; \} .motion-choice \{ border: 1px solid #3a4152; background: #242b38; color: #cbd5e1; \} .motion-choice:hover \{ border-color: #4c5cf7; \} .motion-choice.active \{ background: #3949d6; border-color: #7d8cff; color: #fff; box-shadow: 0 0 0 1px #7d8cff inset; \} .motion-cards \{ display: flex; flex-direction: column; gap: 12px; max-width: 520px; \} .motion-card \{ background: #202733; border: 1px solid #3a4152; border-radius: 10px; padding: 14px; \} .motion-card-head \{ display: flex; align-items: center; justify-content: space-between; gap: 10px; \} .motion-card-head h3 \{ margin: 0; \} .motion-meta \{ color: #8fa0b8; font-size: 12px; margin: 6px 0 0; \}"

view main {
	title "FLIP card layout"
	col class: "motion-page" {
		h1 "FLIP card layout"
		p "Filter or reorder the cards. Remaining cards FLIP to their new positions; hidden cards fade out; new cards fade in." class: "motion-note"

		div class: "motion-toolbar" {
			button "All" class: "motion-choice" + (filter == "all" ? " active" : "") onClick: showAll
			button "Design" class: "motion-choice" + (filter == "design" ? " active" : "") onClick: showDesign
			button "Development" class: "motion-choice" + (filter == "development" ? " active" : "") onClick: showDev
			button "Reverse order" class: "motion-choice" + (order == "reversed" ? " active" : "") onClick: toggleOrder
		}

		div class: "motion-cards" {
			if order == "normal" {
				if filter == "all" or filter == "design" {
					projectCard name: "Alpha" kind: "design"
				}
				if filter == "all" or filter == "development" {
					projectCard name: "Beta" kind: "development"
				}
				if filter == "all" or filter == "design" {
					projectCard name: "Gamma" kind: "design"
				}
				if filter == "all" or filter == "development" {
					projectCard name: "Delta" kind: "development"
				}
			} else {
				if filter == "all" or filter == "development" {
					projectCard name: "Delta" kind: "development"
				}
				if filter == "all" or filter == "design" {
					projectCard name: "Gamma" kind: "design"
				}
				if filter == "all" or filter == "development" {
					projectCard name: "Beta" kind: "development"
				}
				if filter == "all" or filter == "design" {
					projectCard name: "Alpha" kind: "design"
				}
			}
		}

		p "Shown: {filter} · Order: {order}" class: "motion-note"
	}
}

```

## Attributes

| Attribute | Meaning |
|---|---|
| motion | layout (FLIP only), ade, scale, slide or 
one. |
| motionKey | Stable identity across renders. Required for list items and components. |
| motionDuration | Transition duration, e.g. 420ms. |
| motionEasing | CSS timing function, e.g. ase-out. |
| motionDelay | Delay before the transition. |
| motionExit | alse disables the exit clone for large lists. |

layout animates position changes with a translate and lets width/height/color
changes animate separately. ade, scale and slide add the corresponding
enter/exit effect. Cards that disappear are cloned into a fixed overlay, faded
out and removed; new cards fade in.

## Why keys matter

data-sui-id is assigned per element in the source tree, so every iteration of
a list node shares the same id. motionKey supplies the per-item identity:

`sui
list project in projects {
	card motion: "layout" motionKey: project.id { ... }
}
`

In this example the component receives 
ame and uses motionKey: name.

## Designer

Select an element and open **Motion (FLIP / enter / exit)**. Set the effect,
key, duration, easing, delay and whether to animate the exit. Tables, trees,
charts and the toggle controls have their own specialized motion sections.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The un.ps1 / un.bat scripts rebuild the desktop demo and launch it.
