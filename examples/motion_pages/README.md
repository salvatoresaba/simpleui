# Shared element transition across pages

The same `motionKey` exists on the home and detail pages. When `navigate()`
switches page, the runtime captures the visible page's card, renders the new
page, then moves and resizes the card from the old position to the new one.
Navigating back plays the reverse transition.

The navigation button lives inside the card, so it travels with the card; its
label switches as soon as the new page is rendered, while the measured card and
part rectangles animate to their final state.

## Complete source

```sui
# Shared element transition across pages.
#
# The same motionKey ("hero") exists on both pages. navigate() captures the
# visible page before switching, then morphs the new hero from the old page's
# measured position and size to its final measured state.

fn openDetail() -> null {
	navigate("detail")
}

fn backHome() -> null {
	navigate("home")
}

css ".page \{ max-width: 820px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \} .page-note \{ color: #8fa0b8; font-size: 13px; \} .hero-card \{ background: #202733; border: 1px solid #3a4152; border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 8px; \} .hero-home \{ width: 300px; align-self: flex-start; \} .hero-detail \{ width: 100%; min-height: 260px; justify-content: center; \} .hero-card h2 \{ margin: 0; \} .hero-meta \{ color: #8fa0b8; font-size: 12px; margin: 0; \} .hero-card button \{ align-self: flex-start; \}"

view home {
	title "Home"
	col class: "page" {
		h1 "Projects"
		p "Open the project. The hero card is shared with the detail page and animates across." class: "page-note"
		card class: "hero-card hero-home" motion: "layout" motionKey: "hero" motionDuration: "480ms" motionEasing: "ease-out" {
			h2 "Project Alpha"
			p "Design system · 12 tasks" class: "hero-meta"
			p "This card keeps the same motionKey on both pages." class: "hero-meta"
			button "Open detail" onClick: openDetail
		}
	}
}

view detail {
	title "Detail"
	col class: "page" {
		card class: "hero-card hero-detail" motion: "layout" motionKey: "hero" motionDuration: "480ms" motionEasing: "ease-out" {
			h2 "Project Alpha"
			p "Design system · 12 tasks" class: "hero-meta"
			p "The shared hero grows to fill the detail page. Navigate back to see the reverse transition." class: "hero-meta"
			button "Back" onClick: backHome
		}
		p "Detail content is not shared; only the hero card crosses the page boundary." class: "page-note"
	}
}
```

## How it works

- `navigate(page)` calls `__SUI_MOTION.capturePage()` before switching.
- The visible page's `[data-sui-motion]` elements are snapshotted by
  `motionKey`, including the card rectangle and the rectangle of every direct
  visual part (heading, paragraph, button).
- After every page is rendered, `flushShared()` matches keys in the newly visible
  page and measures the final rectangle of the same parts.
- A content-free surface layer draws the card background, border and shadow and
  animates its measured rectangle with explicit pixel values. A separate content
  layer keeps the final layout and is clipped by that surface, so text never
  spills outside the card while it grows.
- Each part is translated from its measured start rectangle to its measured end
  rectangle with the same easing. The parts travel in straight lines; the box
  changes size independently. The browser never reflows or rewraps the text
  during the transition.
- The real new element stays hidden until the overlay reaches the final
  geometry, then the overlay is removed and the real element is revealed at the
  same rectangle, so there is no jump or cross-fade.
- Non-shared content simply appears or disappears with the page.
- Use `motionDuration`, `motionEasing` and `motionDelay` to tune the transition.

`motionKey` is required for shared elements. The same key on the same page is a
normal FLIP; the same key on different pages is a shared-element transition.

## Designer

The **Motion (FLIP / enter / exit)** inspector sets effect, key, duration,
easing and delay on the cards. Page navigation is added from the **Pages**
palette and by calling `navigate(...)` in a handler.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The `run.ps1` / `run.bat` scripts rebuild the desktop demo and launch it.