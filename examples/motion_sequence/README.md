# Motion sequences: loops, springs and graph events

## What it demonstrates

This example uses the `motionGraph` attribute as a small coordinated scene. It
covers the features added after the base state graph:

- **Spring easing** on the scene trigger and on a follower.
- **Infinite alternate loop** on the pulse card, stopped by a second click.
- **Finite alternate loop** on the bounce card (three iterations), followed by
  an `after` link that settles it.
- **Emitted custom events** (`scene-open` and `scene-close`) from the scene
  trigger.
- **Built-in lifecycle events**: `sui:motionstart`, `sui:motionend` and
  `sui:motioncancel` drive the spark and cancel-light cards.
- **Stable cross-element sources**: each card has a `motionKey`; the followers
  use it as their `sourceId`.
- **Element control API**: `window.SimpleUIMotion` and `element.__suiMotion`.

## Complete SUI source

```sui
# Motion sequences: loops, spring easing and graph-to-graph events.
#
# Open the designer, select a card, then Motion > Open animation panel.
# Each card has a motionKey so other graphs can use it as an event source.

css ".page \{ max-width: 1080px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \} .grid \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); gap: 16px; align-items: start; \} .card \{ box-sizing: border-box; border: 1px solid #3a4152; border-radius: 14px; padding: 16px; color: #e6e6ef; display: flex; flex-direction: column; gap: 8px; overflow: hidden; justify-self: start; \} .card h2 \{ margin: 0; font-size: 16px; \} .note \{ color: #8fa0b8; font-size: 12px; line-height: 1.45; \} .kbd \{ color: #c7d2fe; font-weight: 600; \}"

view main {
	title "Motion sequences"
	col class: "page" {
		h1 "Motion sequences"
		p "Click the scene trigger to open a spring transition; the event fans out to other cards. The pulse loops until a second click, the bounce plays three iterations and settles, and the spark reacts to the transition start." class: "note"
		div class: "grid" {
			card class: "card" motionKey: "scene-trigger" motionGraph: "\{\"initial\":\"idle\",\"states\":[\{\"id\":\"idle\",\"name\":\"Idle\",\"css\":\"width: 280px; height: 150px; background-color: #243247; transform: translateY(0px); box-shadow: 0 10px 24px rgba(0,0,0,0.28);\"\},\{\"id\":\"open\",\"name\":\"Open\",\"css\":\"width: 340px; height: 180px; background-color: #3e6fa8; transform: translateY(-8px); box-shadow: 0 20px 44px rgba(0,0,0,0.45);\"\}],\"links\":[\{\"id\":\"to-open\",\"from\":\"idle\",\"to\":\"open\",\"event\":\"click\",\"duration\":\"560ms\",\"spring\":\{\"stiffness\":180,\"damping\":14,\"mass\":1\},\"emit\":\{\"on\":\"end\",\"name\":\"scene-open\"\}\},\{\"id\":\"to-idle\",\"from\":\"open\",\"to\":\"idle\",\"event\":\"click\",\"duration\":\"320ms\",\"easing\":\"ease-in\",\"emit\":\{\"on\":\"end\",\"name\":\"scene-close\"\}\}]\}" {
				h2 "Scene trigger"
				p "Click: spring to open and emit scene-open. Click again to close." class: "note"
			}
			card class: "card" motionKey: "follower-a" motionGraph: "\{\"initial\":\"waiting\",\"states\":[\{\"id\":\"waiting\",\"name\":\"Waiting\",\"css\":\"width: 280px; height: 150px; background-color: #2b2b3a; transform: translateY(0px) scale(1); opacity: 0.65;\"\},\{\"id\":\"shown\",\"name\":\"Shown\",\"css\":\"width: 280px; height: 150px; background-color: #6750a4; transform: translateY(-4px) scale(1.02); opacity: 1;\"\}],\"links\":[\{\"id\":\"show\",\"from\":\"waiting\",\"to\":\"shown\",\"event\":\"custom\",\"customEvent\":\"scene-open\",\"sourceId\":\"scene-trigger\",\"duration\":\"520ms\",\"spring\":\{\"stiffness\":140,\"damping\":12,\"mass\":1\}\},\{\"id\":\"hide\",\"from\":\"shown\",\"to\":\"waiting\",\"event\":\"custom\",\"customEvent\":\"scene-close\",\"sourceId\":\"scene-trigger\",\"duration\":\"300ms\",\"easing\":\"ease-in\"\}]\}" {
				h2 "Follower A"
				p "Listens to the emitted scene-open custom event." class: "note"
			}
			card class: "card" motionKey: "follower-b" motionGraph: "\{\"initial\":\"hidden\",\"states\":[\{\"id\":\"hidden\",\"name\":\"Hidden\",\"css\":\"width: 280px; height: 150px; background-color: #2a2f3a; transform: translateX(0px); opacity: 0.4;\"\},\{\"id\":\"visible\",\"name\":\"Visible\",\"css\":\"width: 280px; height: 150px; background-color: #2f7d6b; transform: translateX(18px); opacity: 1;\"\}],\"links\":[\{\"id\":\"show\",\"from\":\"hidden\",\"to\":\"visible\",\"event\":\"custom\",\"customEvent\":\"sui:motionend\",\"sourceId\":\"scene-trigger\",\"duration\":\"480ms\",\"easing\":\"cubic-bezier(0.34, 1.56, 0.64, 1)\",\"delay\":\"80ms\"\},\{\"id\":\"hide\",\"from\":\"visible\",\"to\":\"hidden\",\"event\":\"custom\",\"customEvent\":\"scene-close\",\"sourceId\":\"scene-trigger\",\"duration\":\"260ms\",\"easing\":\"ease-in\"\}]\}" {
				h2 "Follower B"
				p "Listens to the built-in sui:motionend event with an extra delay." class: "note"
			}
			card class: "card" motionKey: "pulse-card" motionGraph: "\{\"initial\":\"rest\",\"states\":[\{\"id\":\"rest\",\"name\":\"Rest\",\"css\":\"width: 280px; height: 150px; background-color: #30343f; transform: scale(1); box-shadow: 0 0 0 rgba(124,77,255,0);\"\},\{\"id\":\"pulse\",\"name\":\"Pulse\",\"css\":\"width: 280px; height: 150px; background-color: #5b3fa8; transform: scale(1.06); box-shadow: 0 0 28px rgba(124,77,255,0.7);\"\}],\"links\":[\{\"id\":\"start\",\"from\":\"rest\",\"to\":\"pulse\",\"event\":\"click\",\"duration\":\"760ms\",\"easing\":\"ease-in-out\",\"iterations\":\"infinite\",\"direction\":\"alternate\"\},\{\"id\":\"stop\",\"from\":\"pulse\",\"to\":\"rest\",\"event\":\"click\",\"duration\":\"260ms\",\"easing\":\"ease-out\"\}]\}" {
				h2 "Pulse"
				p "Click: infinite alternate loop. Click again to stop and cancel it." class: "note"
			}
			card class: "card" motionKey: "bounce-card" motionGraph: "\{\"initial\":\"ready\",\"states\":[\{\"id\":\"ready\",\"name\":\"Ready\",\"css\":\"width: 280px; height: 150px; background-color: #2d3a2f; transform: translateY(0px);\"\},\{\"id\":\"up\",\"name\":\"Up\",\"css\":\"width: 280px; height: 150px; background-color: #4d7a56; transform: translateY(-16px);\"\},\{\"id\":\"done\",\"name\":\"Done\",\"css\":\"width: 280px; height: 150px; background-color: #6a9f74; transform: translateY(0px);\"\}],\"links\":[\{\"id\":\"bounce\",\"from\":\"ready\",\"to\":\"up\",\"event\":\"click\",\"duration\":\"180ms\",\"easing\":\"ease-out\",\"iterations\":3,\"direction\":\"alternate\"\},\{\"id\":\"settle\",\"from\":\"up\",\"to\":\"done\",\"event\":\"after\",\"duration\":\"220ms\",\"easing\":\"ease-out\"\},\{\"id\":\"reset\",\"from\":\"done\",\"to\":\"ready\",\"event\":\"click\",\"duration\":\"220ms\",\"easing\":\"ease-in\"\}]\}" {
				h2 "Bounce"
				p "Three alternate iterations, then an after link settles the card." class: "note"
			}
			card class: "card" motionKey: "spark-card" motionGraph: "\{\"initial\":\"dim\",\"states\":[\{\"id\":\"dim\",\"name\":\"Dim\",\"css\":\"width: 280px; height: 150px; background-color: #2b2b3a; opacity: 0.45;\"\},\{\"id\":\"lit\",\"name\":\"Lit\",\"css\":\"width: 280px; height: 150px; background-color: #ffb347; opacity: 1;\"\}],\"links\":[\{\"id\":\"light\",\"from\":\"dim\",\"to\":\"lit\",\"event\":\"custom\",\"customEvent\":\"sui:motionstart\",\"sourceId\":\"scene-trigger\",\"duration\":\"160ms\",\"easing\":\"ease-out\"\},\{\"id\":\"fade\",\"from\":\"lit\",\"to\":\"dim\",\"event\":\"timer\",\"timer\":\"650ms\",\"duration\":\"260ms\",\"easing\":\"ease-in\"\}]\}" {
				h2 "Start spark"
				p "Lights on sui:motionstart from the scene trigger, then fades on a timer." class: "note"
			}
			card class: "card" motionKey: "cancel-light" motionGraph: "\{\"initial\":\"off\",\"states\":[\{\"id\":\"off\",\"name\":\"Off\",\"css\":\"width: 280px; height: 150px; background-color: #2b2b3a; opacity: 0.45;\"\},\{\"id\":\"flash\",\"name\":\"Flash\",\"css\":\"width: 280px; height: 150px; background-color: #d95f5f; opacity: 1;\"\}],\"links\":[\{\"id\":\"flash\",\"from\":\"off\",\"to\":\"flash\",\"event\":\"custom\",\"customEvent\":\"sui:motioncancel\",\"sourceId\":\"pulse-card\",\"duration\":\"180ms\",\"easing\":\"ease-out\"\},\{\"id\":\"clear\",\"from\":\"flash\",\"to\":\"off\",\"event\":\"timer\",\"timer\":\"450ms\",\"duration\":\"220ms\",\"easing\":\"ease-in\"\}]\}" {
				h2 "Cancel light"
				p "Flashes when the pulse loop is interrupted (sui:motioncancel)." class: "note"
			}
		}
	}
}
```

## How it works

### Spring

The scene trigger opens with:

```json
{ "id": "to-open", "from": "idle", "to": "open", "event": "click",
  "duration": "560ms", "spring": { "stiffness": 180, "damping": 14, "mass": 1 },
  "emit": { "on": "end", "name": "scene-open" } }
```

`spring` overrides `easing`. The runtime samples the damped spring into a CSS
`linear()` curve; if the browser cannot use it, it falls back to an overshoot
`cubic-bezier`. `duration` controls the spring speed, while `stiffness`,
`damping` and `mass` shape the curve. The follower cards use different spring
values so the same click produces a layered motion.

### Loops

The pulse card starts an infinite loop:

```json
{ "id": "start", "from": "rest", "to": "pulse", "event": "click",
  "duration": "760ms", "easing": "ease-in-out",
  "iterations": "infinite", "direction": "alternate" }
```

An infinite link never finishes: it does not emit `sui:motionend` and cannot
start an `after` link. A second click on the pulse card uses the `pulse → rest`
link; the runtime samples the current computed values, cancels the loop and emits
`sui:motioncancel` before starting the next transition.

The bounce card uses a finite loop:

```json
{ "id": "bounce", "from": "ready", "to": "up", "event": "click",
  "duration": "180ms", "easing": "ease-out",
  "iterations": 3, "direction": "alternate" }
```

Three alternate iterations end on the target side. When the loop really finishes,
the `up → done` `after` link runs. `after` starts from the finish callback, so
it never races the final values. With an even alternate count the last visual
frame is the start side; the runtime still settles on the target state, which
causes a final snap. Use an odd count when the loop must end on the target.

### Events and orchestration

The runtime dispatches these DOM events on the animated element:

| Event | When | Detail |
|---|---|---|
| `sui:motionstate` | State entry | `{from, to, state, link}` |
| `sui:motionstart` | Transition start | `{from, to, state, link}` |
| `sui:motionend` | Transition finish | `{from, to, state, link}` |
| `sui:motioncancel` | Interrupted transition | `{from, to, state, link}` |

A link can also emit its own event with `emit`:

```json
"emit": { "on": "end", "name": "scene-open" }
```

`on` accepts `start`, `end` or `cancel`; a string value defaults to `end`.
Events bubble and are composed. Another graph listens with a normal `custom`
trigger:

```json
{ "from": "waiting", "to": "shown", "event": "custom",
  "customEvent": "scene-open", "sourceId": "scene-trigger" }
```

`sourceId` matches either `data-sui-id` or `data-sui-motion-key`. A literal
`motionKey` is emitted as `data-sui-motion-key`, so it is the stable choice for
hand-written or designer-authored cross-element links. Each card in this example
uses a readable key such as `scene-trigger` or `pulse-card`.

### Element API

In the compiled page, each graph element exposes:

```js
var api = window.SimpleUIMotion.get('.spring-card');
api.to('b', { duration: '400ms', easing: 'ease-out' });
api.pause();
api.resume();
api.seek(200);        // milliseconds
api.progress;         // 0..1 in the current iteration for infinite loops
api.state;
api.paused;
```

`to(state, opts)` accepts `duration`, `easing`, `delay`, `iterations`,
`direction`, `playbackRate`, `spring` and `emit`. The same object is available
as `element.__suiMotion`.

## Designer workflow

1. Open the example:
   `simpleui gui examples/motion_sequence/app.sui`.
2. Select a card, open **Motion > Open animation panel**, then double-click the
   state or arrow you want to edit.
3. Double-click an arrow to edit its trigger, duration, easing, delay,
   iterations, direction, playback rate, spring values and emitted event. The
   timing and spring fields use a compact two-column grid; numeric fields are
   spin boxes with explicit units (`ms`, `×`, `N/m`, `N·s/m`, `kg`). **OK** closes
   the overlay after the live edits.
4. Use **+** on a state to create a link, **Create state** for an unlinked state,
   and **Duplicate state** to copy CSS and parameters. The normal element
   **Duplicate** copies the whole graph.
5. For a cross-element trigger, choose **event source**. Each option shows the
   element name, its `motionKey` or `data-sui-id`, and its tree path; a **Source
   code** line repeats the exact identifier. Prefer a `motionKey` for a stable
   readable name.
6. Use **Preview** to play a state or link with its delay, spring, loop direction
   and playback rate. **Stop preview** cancels a running preview; infinite
   previews also stop when another state, link or panel is opened. The panel
   supports pan, zoom and Fit.

## Build and run

```bash
simpleui check examples/motion_sequence/app.sui --json
simpleui build examples/motion_sequence/app.sui --out examples/motion_sequence/dist
simpleui build examples/motion_sequence/app.sui --single --out example.html
simpleui serve examples/motion_sequence/app.sui --watch
```

The `run.ps1` / `run.bat` scripts rebuild the desktop demo and launch it. A
ready-to-run build is in `dist/`, with the complete source as `dist/app.sui`.

## Limits and critical points

- Infinite links never emit `sui:motionend` and never trigger `after` links;
  use an explicit click, timer, variable or external event to leave the loop.
- A finite `alternate` loop settles on the target state. An even count ends the
  visual cycle on the start side and then snaps to the target; use an odd count
  when that matters.
- A `sui:motionstart`, `sui:motionend`, `sui:motioncancel` or custom `emit`
  listener can start another transition. The runtime checks the transition token
  and aborts the interrupted setup, so the newest transition wins.
- `after` links run one at a time per source state. Build longer sequences by
  chaining states; use timers or events when steps must overlap.
- The spring is a visual approximation sampled into `linear()`; very short
  durations or extreme damping values may look flatter than a physical spring.
- `prefers-reduced-motion` jumps to the target state and ignores loops.
- External event triggers are DOM events, not SUI function calls. Use the API
  when application code needs to drive the graph directly.
