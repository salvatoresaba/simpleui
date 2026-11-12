# Shared element transition across pages

The same motionKey exists on the home and detail pages. When 
avigate()
switches page, the runtime captures the visible page's card, renders the new
page, then FLIPs the new card from the old position and size to the new one.
Navigating back plays the reverse transition.

## Complete source

```sui
# Shared element transition across pages.
#
# The same motionKey ("hero") exists on both pages. navigate() captures the
# visible page before switching, then FLIPs the new page's hero from the old
# page's position and size.

fn openDetail() -> null {
	navigate("detail")
}

fn backHome() -> null {
	navigate("home")
}

css ".page \{ max-width: 820px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \} .page-note \{ color: #8fa0b8; font-size: 13px; \} .hero-card \{ background: #202733; border: 1px solid #3a4152; border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 8px; \} .hero-home \{ width: 300px; align-self: flex-start; \} .hero-detail \{ width: 100%; min-height: 260px; justify-content: center; \} .hero-card h2 \{ margin: 0; \} .hero-meta \{ color: #8fa0b8; font-size: 12px; margin: 0; \} .page-actions \{ display: flex; gap: 8px; flex-wrap: wrap; \}"

view home {
	title "Home"
	col class: "page" {
		h1 "Projects"
		p "Open the project. The hero card is shared with the detail page and animates across." class: "page-note"
		card class: "hero-card hero-home" motion: "layout" motionKey: "hero" motionDuration: "480ms" motionEasing: "ease-out" {
			h2 "Project Alpha"
			p "Design system · 12 tasks" class: "hero-meta"
			p "This card keeps the same motionKey on both pages." class: "hero-meta"
		}
		div class: "page-actions" {
			button "Open detail" onClick: openDetail
		}
	}
}

view detail {
	title "Detail"
	col class: "page" {
		div class: "page-actions" {
			button "Back" onClick: backHome
		}
		card class: "hero-card hero-detail" motion: "layout" motionKey: "hero" motionDuration: "480ms" motionEasing: "ease-out" {
			h2 "Project Alpha"
			p "Design system · 12 tasks" class: "hero-meta"
			p "The shared hero grows to fill the detail page. Navigate back to see the reverse transition." class: "hero-meta"
		}
		p "Detail content is not shared; only the hero card crosses the page boundary." class: "page-note"
	}
}
```

## How it works

- 
avigate(page) calls __SUI_MOTION.capturePage() before switching.
- The visible page's [data-sui-motion] elements are snapshotted by
  motionKey.
- After every page is rendered, lushShared() matches keys in the newly visible
  page and animates them from the old rectangle to the new one, including
  width/height.
- Non-shared content simply appears or disappears with the page.
- The shared element keeps motion: "layout"; use the usual motionDuration,
  motionEasing and motionDelay to tune it.

motionKey is required for shared elements. The same key on the same page is a
normal FLIP; the same key on different pages is a shared-element transition.

## Designer

The **Motion (FLIP / enter / exit)** inspector sets effect, key, duration,
easing and delay on the cards. Page navigation is added from the **Pages**
palette and by calling 
avigate(...) in a handler.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The un.ps1 / un.bat scripts rebuild the desktop demo and launch it.
