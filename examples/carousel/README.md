# Carousel

A slide track with previous/next buttons, dots and keyboard support. The active slide is bound to an integer state through `value`.

## Complete source

```sui
state slide = 0

css ".carousel-page \{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \} .carousel-note \{ color: #8fa0b8; font-size: 13px; \} .carousel-slide-title \{ margin: 0 0 6px; \}"

view main {
	title "Carousel"
	col class: "carousel-page" {
		h1 "Carousel"
		p "Arrow buttons, dots, Left/Right keys and an optional autoplay. The active slide is bound to a state." class: "carousel-note"
		carousel value: slide height: "180px" {
			slide label: "Welcome" {
				h2 "Welcome" class: "carousel-slide-title"
				p "The first slide introduces the product."
			}
			slide label: "Layout" {
				h2 "Layout" class: "carousel-slide-title"
				p "Build rows, columns, cards and split panes."
			}
			slide label: "Controls" {
				h2 "Controls" class: "carousel-slide-title"
				p "Trees, charts, dropdowns, drawers and more."
			}
		}
		p "Slide {slide + 1} of 3" class: "carousel-note"
	}
}
```

## Designer

Choose **Content > Carousel** in the palette. The **Carousel** section sets the bound state, height, autoplay, interval, loop, controls and indicators. Add `slide` children from the tree.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

