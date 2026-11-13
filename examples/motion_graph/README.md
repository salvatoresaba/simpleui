# Motion graph: states, triggers and chained transitions

This example uses the `motionGraph` attribute. One element owns a small state
machine: each state is a CSS snapshot, each link carries a trigger and timing.
The runtime interpolates the measured start and end values.

- **Card 1 — click graph.** The first card toggles between `closed` and `open`
  on click. Its width, height, left offset and background animate.
- **Card 2 — state trigger.** The second card follows the `open` state created
  by the switch. The link uses `event: "state"` and `when: "open"` (or `!open`),
  so changing the switch drives the transition.
- **Chaining.** Links with `event: "after"` continue when the previous animation
  ends. The runtime keeps a token per element, so a new transition cancels the
  pending chain.

## Complete source

```sui
# Motion graph example.
#
# Card 1: click toggles between two captured states.
# Card 2: a state variable (open) drives the transition.
#
# Open the designer, select a card, open Motion > Open animation panel.
state open = false

css ".page \{ max-width: 900px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 20px; \} .demo-card \{ border: 1px solid #3a4152; border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 8px; color: #e6e6ef; overflow: hidden; \} .demo-card h2 \{ margin: 0; \} .demo-note \{ color: #8fa0b8; font-size: 13px; \} .demo-row \{ display: flex; gap: 16px; align-items: flex-start; flex-wrap: wrap; \}"

view main {
	title "Motion graph"
	col class: "page" {
		h1 "Animation graph"
		p "The first card toggles on click. The second card follows the Open switch through a state trigger." class: "demo-note"
		card class: "demo-card" motionGraph: "\{\"initial\":\"closed\",\"states\":[\{\"id\":\"closed\",\"name\":\"Chiuso\",\"css\":\"width: 320px; height: 160px; margin-left: 0px; background-color: #202733;\"\},\{\"id\":\"open\",\"name\":\"Aperto\",\"css\":\"width: 520px; height: 240px; margin-left: 60px; background-color: #2a3a55;\"\}],\"links\":[\{\"id\":\"l1\",\"from\":\"closed\",\"to\":\"open\",\"event\":\"click\",\"duration\":\"360ms\",\"easing\":\"ease-out\",\"delay\":\"0ms\"\},\{\"id\":\"l2\",\"from\":\"open\",\"to\":\"closed\",\"event\":\"click\",\"duration\":\"300ms\",\"easing\":\"ease-in\",\"delay\":\"0ms\"\}]\}" {
			h2 "Click card"
			p "Click anywhere on this card. It moves and grows, then returns on the next click." class: "demo-note"
		}
		div class: "demo-row" {
			card class: "demo-card" motionGraph: "\{\"initial\":\"closed\",\"states\":[\{\"id\":\"closed\",\"name\":\"Chiuso\",\"css\":\"width: 320px; height: 160px; background-color: #202733;\"\},\{\"id\":\"open\",\"name\":\"Aperto\",\"css\":\"width: 520px; height: 220px; background-color: #3b2f55;\"\}],\"links\":[\{\"id\":\"l1\",\"from\":\"closed\",\"to\":\"open\",\"event\":\"state\",\"when\":\"open\",\"duration\":\"360ms\",\"easing\":\"ease-out\",\"delay\":\"0ms\"\},\{\"id\":\"l2\",\"from\":\"open\",\"to\":\"closed\",\"event\":\"state\",\"when\":\"!open\",\"duration\":\"300ms\",\"easing\":\"ease-in\",\"delay\":\"0ms\"\}]\}" {
				h2 "State-driven card"
				p "This card follows the Open switch using a state trigger." class: "demo-note"
			}
			switch "Open" checked: open
		}
	}
}
```

## How it works

- `motionGraph` is a literal JSON object with `initial`, `states` and `links`.
- `states[].css` stores declarations such as `width`, `height`, `margin-left`,
  colors and opacity. These are explicit start/end values.
- `links[].event` is a DOM event (`click`, `pointerenter`, `focus`, ...),
  `state` for a boolean SUI state, or `after` for a transition chain.
- `links[].duration`, `easing` and `delay` describe the interpolation.
- Multiple outgoing links are matched in order; the first one for the current
  state and trigger wins.
- The runtime honors `prefers-reduced-motion` by jumping to the target state.
- A re-render keeps the current graph state for the same element key.

`motionGraph` is per element. For shared elements across pages use `motionKey`
and the FLIP engine described in `examples/motion_pages`.

## Designer workflow

1. Open `simpleui gui examples/motion_graph/app.sui`.
2. Select a card in the tree or preview.
3. Inspector → **Motion (animation panel)** → **Open animation panel**.
   The panel is compact and resizable from its bottom-right handle; the canvas
   centers and scales the diagram. Opening the panel selects the element's
   current or initial state. Panels and overlays stay inside the viewport.
4. **Layout & shared** in the toolbar opens a draggable overlay for
   FLIP/key/duration/easing/delay/exit for lists and page transitions.
5. State blocks grow taller when outgoing triggers are added, and can be
   dragged to custom positions. Each block has **+** to add a connected state
   and **×** to delete it. **+** immediately creates a positioned state block
   and a link, then opens the source state overlay to configure the trigger.
   **Aggiungi trigger** in the overlay is for existing states.
6. Triggers belong to the source state. Outgoing links are numbered at their
   starting point; hovering the number shows the trigger hint. Links always
   leave the right side and enter the target's left centre; backward links are
   routed around the boxes. Click and other DOM events can be used once per
   source state; variable and timer triggers can be repeated. A variable
   trigger lists the available SUI states and asks for the activating value; a
   timer trigger starts after the chosen duration.
7. Double-click a state to edit its overlay (name, CSS and the list of active
   triggers with destinations); use **Aggiungi trigger** for more. Double-click
   an arrow for trigger, duration, easing and delay.
8. Clicking a state applies it in the preview and locks the element. Inspector
   changes are written into the focused state and appear immediately.
9. While editing, the graph collapses to a bar with **Salva stato** and
   **Chiudi**. Save validates and returns to the updated panel; close restores
   the snapshot and discards changes. Saving rejects an empty trigger, a missing
   variable, duplicate state parameters or a repeated single-use event.
10. Press **Anteprima** to play the selected state or link.

States/links and Layout & shared live in the same panel; there are no duplicate
motion fields in the inspector. **Remove states** clears the graph but keeps
Layout & shared.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The `run.ps1` / `run.bat` scripts rebuild the desktop demo and launch it. A
ready-to-run build is in `dist/`, with the complete source as `dist/app.sui`.

## Limits

- The graph is per element; multi-element choreography needs one graph per
  element or an `after`/shared state trigger.
- State snapshots use the current viewport values. For responsive layouts,
  capture or review each breakpoint.
- A state trigger supports a boolean state name, optionally negated; arbitrary
  expressions are not evaluated.
