# Animation lab

A focused gallery of the controls whose graphics can be animated. Each card has
independent **Style** and **Animation** selectors that apply at runtime; the
checkbox, switch and radio cards also expose **Shape** and **Theme**.

The page intentionally contains no tables, trees, charts or media: only
animatable controls.

## Controls, styles and animations

| Control | Styles | Animations |
|---|---|---|
| Progress | Classic, Thin line, Segments, Stripes, Square, Neon, Dashed | Pulse, Stripes, Shimmer, Glow, Indeterminate, Aurora, Breathe, Ticker, Spark, Wave, None |
| Spinner | Ring, Tri dots, Bars, Orbit, Comet, Pulse, Square, **Hourglass**, **Drop fill** | Spin, Pulse, Flip, Reverse, Orbit, Wobble, Bounce, Ripple, Swing, None |
| Badge | Pill, Tag, Outline, Neon, Status, Brutalist, Glass | Pulse, Glow, Float, Pop, Blink, Tilt, Ping, Bounce, Jelly, Sheen, **Dot to pill**, **Border first**, None |
| Alert | Accent bar, Filled, Outline, Glass, Brutalist, Info | Fade in, Slide in, Shake, Pulse, Zoom in, Bounce in, Slide right, Wobble, Breathe, Flash, **Wipe**, **Border first**, None |
| Card | Elevated, Flat, Outline, Glass, Neon, Brutalist, Dashed | Fade in, Float, Rise, Zoom in, Tilt in, Halo, Drift, Unfold, **Dot to card**, **Border first**, **Wipe**, **Flip**, **Squash**, None |
| Button | Solid, Gradient, Outline, Ghost, Neon, 3D, Pill, Brutalist | Pulse, Float, Glow, Jelly, Bounce, Tilt, Ring, Sheen, Press, **Border first**, **Wipe**, **Pop**, None |
| Tabs | Underline, Pills, Segmented, Boxed, Brutalist, Neon | Pop, Slide, Glow, Sweep, None |
| Input | Outline, Filled, Underline, Pill, Glass, Brutalist | Focus ring, Focus glow, Focus lift, Focus shake, None |
| Checkbox | **Classic, Round, Tile, LED, Brutalist, Neon, Sweep fill** (plus 5 colour themes) | Auto, Scale, Morph, Fade, Slide, None |
| Switch | 8 shapes | Auto, Scale, Morph, Fade, Slide, None |
| Radio | 5 shapes | Auto, Scale, Morph, Fade, Slide, None |
| Dropdown | Classic, Pill, Neon, Brutalist, Glass, Outline; **item highlight**: accent bar, glow, aura, gradient, border, dot, check mark, **bar + aura**, **bar + colour cloud**, **bar + underline**, **bar + glow**, none | Slide, Fade, Scale, None |
| Popover | Card, Glass, Neon, Brutalist | Slide, Fade, Scale, None |
| Combobox | Classic, Pill, Neon, Brutalist | Slide, Fade, Scale, None |
| Command palette | - | Slide, Fade, Scale, None |

Checkbox, switch and radio also expose eight colour themes: Default, Accent,
Success, Warning, Danger, Mono, Dark and Light.

## Checkbox structures

The checkbox is not a recoloured square: the seven styles change size, radius,
border, surface and the mark itself.

| Style | Structure |
|---|---|
| Classic | 20 px rounded square with a drawn tick |
| Round | 22 px circle with a ring when checked |
| Tile | 30 px tile with a gradient fill and a large tick |
| LED | dark square with a glowing dot instead of a tick |
| Brutalist | hard black border, yellow fill, offset shadow |
| Neon | hollow ring with a glowing cyan tick |
| Sweep fill | the accent colour rises from the bottom to fill the box |

Five themes (Accent, Success, Warning, Danger, Mono) recolour the style through
a CSS variable, so structure and colour stay independent.

## Spinner graphics

New shapes beyond the ring family:

- **Hourglass**: a clipped hourglass silhouette; the sand drains, the glass
  flips and the sand fills the other half.
- **Drop fill**: a drop falls into a beaker and the beaker fills with a gradient.

## Dropdown styles and item highlights

A dropdown style applies to **both** the trigger button and the popup panel:
Classic, Pill, Neon, Brutalist, Glass and Outline each define the button surface,
the panel surface, the item text colour and the open/hover state, so the two
parts always belong to the same graphic language.

The item highlight is chosen separately and also supports combinations:

| Highlight | Effect |
|---|---|
| Accent bar | a vertical accent bar grows on hover/focus |
| Glow | border and outer glow |
| Aura | pulsing radial light |
| Gradient | a gradient sweeps across the item |
| Border | an inset outline appears |
| Dot | a dot pops in |
| Check mark | a check scales in on the right |
| **Bar + aura** | accent bar and pulsing light together |
| **Bar + colour cloud** | accent bar over a radial colour cloud |
| **Bar + underline** | the accent stretches into an underline under the item |
| **Bar + glow** | accent bar with a glowing surface |

They react to hover and keyboard focus, so mouse and keyboard selection look the
same.

## Shape transitions

The new **Dot to card** transition turns a small circle into a bar and then into
the full rectangle; **Border first** draws the outline before the surface fades
in; **Wipe** reveals the element from one side; **Flip** and **Squash** deform
the whole surface. They are available on Card, Button, Badge and Alert.

The radio card contains **two options** in the same `name` group, so selecting
one shows the actual choice action and the transition on the newly checked
control.

## Readability and hit areas

- Text inside a styled surface inherits that surface colour; the brutalist
  yellow card, tabs and popover use dark text instead of the muted grey.
- Toggle rows are full width and 42 px tall, so clicking the label text or the
  empty part of the row toggles the control.
- Buttons, overlay triggers and selects have a 34-42 px hit area and
  `touch-action: manipulation`, which removes the touch tap delay.
- The Button card shows a live "Clicks" counter, so the press feedback is
  immediate instead of silent.
- Checkbox, switch and radio apply the new checked state on the next frame
  (180 ms tween when an animation is selected); there is no fixed 700 ms
  freeze between the click and the state change.

## Runtime selection

Every selector is bound to a state variable:

```sui
state name = ""
state pStyle = "lab-p-classic"
state pAnim = "lab-p-anim-wave"
state sShape = "lab-s-ring"
state sAnim = "lab-s-anim-spin"
state bStyle = "lab-b-pill"
state bAnim = "lab-b-anim-ping"
state aStyle = "lab-a-bar"
state aAnim = "lab-a-anim-breathe"
state cStyle = "lab-c-elevated"
state cAnim = "lab-c-anim-halo"
state btnStyle = "lab-btn-gradient"
state btnAnim = "lab-btn-anim-jelly"
state tStyle = "lab-t-underline"
state tAnim = "lab-t-anim-pop"
state iStyle = "lab-i-outline"
state iAnim = "lab-i-anim-glow"
state chkOn = true
state chkStyle = "lab-check-tile"
state chkTheme = "lab-cb-theme-accent"
state chkAnim = "scale"
state swOn = true
state swShape = "pill"
state swTheme = "accent"
state swAnim = "scale"
state rdSel = "a"
state rdShape = "default"
state rdTheme = "accent"
state rdAnim = "slide"
state ddStyle = "lab-dd-classic"
state ddItem = "lab-items-bar"
state ddAnim = "slide"
state popStyle = "lab-pop-card"
state popAnim = "scale"
state cbStyle = "lab-cb-classic"
state cbAnim = "slide"
state cmdAnim = "slide"
state showCommands = false
state motionSpeed = "1"
state motionEase = "default"
state motionDelay = "0"
state motionMsg = ""

fn noop() -> null {
	return null
}


fn copyMotion() -> null {
  let easeCss = ""
  if motionEase != "default" {
    easeCss = "; --lab-motion-easing: " + motionEase
  }
  let delayCss = ""
  if motionDelay != "0" and motionDelay != "0.0" {
    delayCss = "; --lab-motion-delay: " + motionDelay + "s"
  }
  let cssText = ":root \{ --lab-motion-speed: " + motionSpeed + easeCss + delayCss + "; \}"
  let ok = clipboard.writeText(cssText)
  if ok {
    motionMsg = "CSS copied: " + cssText
  } else {
    motionMsg = "Copy failed."
  }
}

fn openCommands() -> null {
	showCommands = true
}

fn pickRadioA() -> null {
	rdSel = "a"
}

fn pickRadioB() -> null {
	rdSel = "b"
}

state presses = 0

fn press() -> null {
	presses = presses + 1
}

css "/* Animation lab: visual styles and motion for the animatable controls. */
.lab-page \{ max-width: 1200px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \}
.lab-head \{ margin: 0; \}
.lab-note \{ color: #8fa0b8; font-size: 13px; margin: 0; \}
.lab-grid \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(340px, 1fr)); gap: 16px; align-items: start; \}
.lab-card \{ background: #121826; border: 1px solid #263141; border-radius: 16px; padding: 16px; display: flex; flex-direction: column; gap: 12px; \}
.lab-card h3 \{ margin: 0; font-size: 15px; \}
.lab-preview \{ min-height: 104px; display: flex; align-items: center; justify-content: center; padding: 14px; border-radius: 12px; background: radial-gradient(circle at 50% 0%, #1a2436, #0d131f); border: 1px solid #1f2937; overflow: hidden; \}
.lab-fields \{ display: grid; grid-template-columns: 1fr 1fr; gap: 8px; \}
.lab-fields .lab-wide \{ grid-column: 1 / -1; \}
.lab-fields select \{ width: 100%; \}
.lab-preview .lab-note \{ color: inherit; opacity: .78; \}
.lab-preview .sui-checkbox-wrap, .lab-preview .sui-switch-wrap, .lab-preview .sui-radio-wrap \{ width: 100%; box-sizing: border-box; justify-content: center; min-height: 42px; padding: 8px 14px; border-radius: 10px; cursor: pointer; touch-action: manipulation; \}
.lab-preview .sui-checkbox-wrap:hover, .lab-preview .sui-switch-wrap:hover, .lab-preview .sui-radio-wrap:hover \{ background: rgba(148,163,184,.12); \}
.lab-preview .sui-button, .lab-preview .sui-menu-button, .lab-preview .sui-popover-trigger \{ min-height: 38px; padding: 8px 16px; touch-action: manipulation; \}
.lab-fields select \{ min-height: 34px; touch-action: manipulation; \}
.lab-preview .sui-button:active, .lab-preview .sui-menu-button:active, .lab-preview .sui-popover-trigger:active \{ scale: .97; \}
.lab-stack \{ display: flex; flex-direction: column; align-items: center; gap: 8px; \}
.lab-kbd \{ font-family: ui-monospace, Consolas, monospace; font-size: 11px; color: #c7d2fe; background: #1e2637; border: 1px solid #334155; border-radius: 4px; padding: 1px 5px; \}

/* ---------- Progress: styles ---------- */
.sui-progress.lab-p-classic \{ height: 10px; border-radius: 999px; background: #263141; \}
.sui-progress.lab-p-classic::-webkit-progress-bar \{ background: #263141; \}
.sui-progress.lab-p-classic::-webkit-progress-value \{ background: linear-gradient(90deg, #4c5cf7, #8b5cf6); border-radius: 999px; \}
.sui-progress.lab-p-thin \{ height: 4px; border-radius: 999px; background: #1f2937; \}
.sui-progress.lab-p-thin::-webkit-progress-bar \{ background: #1f2937; \}
.sui-progress.lab-p-thin::-webkit-progress-value \{ background: #22d3ee; border-radius: 999px; \}
.sui-progress.lab-p-segments \{ height: 14px; border-radius: 4px; background: #0b1220; \}
.sui-progress.lab-p-segments::-webkit-progress-bar \{ background: #0b1220; \}
.sui-progress.lab-p-segments::-webkit-progress-value \{ background: repeating-linear-gradient(90deg, #22d3ee 0 8px, transparent 8px 14px); border-radius: 3px; \}
.sui-progress.lab-p-stripes \{ height: 16px; border-radius: 4px; background: #1f2937; \}
.sui-progress.lab-p-stripes::-webkit-progress-bar \{ background: #1f2937; \}
.sui-progress.lab-p-stripes::-webkit-progress-value \{ background: repeating-linear-gradient(45deg, #f59e0b 0 10px, #ef4444 10px 20px); \}
.sui-progress.lab-p-square \{ height: 16px; border-radius: 0; background: #1f2937; border: 2px solid #0b1220; box-shadow: 5px 5px 0 rgba(0,0,0,.35); \}
.sui-progress.lab-p-square::-webkit-progress-bar \{ background: #1f2937; \}
.sui-progress.lab-p-square::-webkit-progress-value \{ background: #4c5cf7; border-radius: 0; \}
.sui-progress.lab-p-neon \{ height: 12px; border-radius: 999px; background: #0b1220; box-shadow: inset 0 0 6px #22d3ee; \}
.sui-progress.lab-p-neon::-webkit-progress-bar \{ background: #0b1220; \}
.sui-progress.lab-p-neon::-webkit-progress-value \{ background: linear-gradient(90deg, #22d3ee, #a855f7); box-shadow: 0 0 12px #22d3ee; \}
.sui-progress.lab-p-dashed \{ height: 14px; border-radius: 0; background: transparent; border: 2px dashed #64748b; padding: 2px; \}
.sui-progress.lab-p-dashed::-webkit-progress-bar \{ background: transparent; \}
.sui-progress.lab-p-dashed::-webkit-progress-value \{ background: linear-gradient(90deg, #94a3b8, #cbd5e1); \}

/* ---------- Progress: animations ---------- */
.sui-progress.lab-p-anim-pulse::-webkit-progress-value \{ animation: lab-progress-pulse calc(1.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-stripes::-webkit-progress-value \{ background: repeating-linear-gradient(45deg, #4c5cf7 0 10px, #8b5cf6 10px 20px); background-size: 28px 100%; animation: lab-progress-stripes calc(.9s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, linear) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-shimmer::-webkit-progress-value \{ background: linear-gradient(90deg, transparent 20%, rgba(255,255,255,.6) 50%, transparent 80%), linear-gradient(90deg, #4c5cf7, #6366f1); background-size: 220% 100%; animation: lab-progress-shimmer calc(2.2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, linear) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-glow::-webkit-progress-value \{ animation: lab-progress-glow calc(1.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-indeterminate::-webkit-progress-value \{ background: linear-gradient(90deg, transparent 12%, #4c5cf7 42%, #6366f1 58%, transparent 88%); background-size: 200% 100%; animation: lab-progress-scan calc(1.25s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-aurora::-webkit-progress-value \{ animation: lab-progress-aurora calc(7s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, linear) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-breathe::-webkit-progress-value \{ animation: lab-progress-breathe calc(2.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-ticker::-webkit-progress-value \{ background: repeating-linear-gradient(90deg, #22d3ee 0 6px, transparent 6px 12px); animation: lab-progress-ticker calc(1s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, steps(10)) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-spark::-webkit-progress-value \{ animation: lab-progress-spark calc(1.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, steps(6)) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-wave::-webkit-progress-value \{ background: linear-gradient(180deg, #4c5cf7 0 34%, #6366f1 34% 62%, #4c5cf7 62% 100%); background-size: 100% 240%; animation: lab-progress-wave calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-progress.lab-p-anim-none \{ animation: none; \}
@keyframes lab-progress-pulse \{ 0%, 100% \{ opacity: 1; \} 50% \{ opacity: .62; \} \}
@keyframes lab-progress-stripes \{ to \{ background-position: 28px 0; \} \}
@keyframes lab-progress-shimmer \{ from \{ background-position: 220% 0; \} to \{ background-position: -220% 0; \} \}
@keyframes lab-progress-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.5); \} \}
@keyframes lab-progress-scan \{ 0% \{ background-position: 130% 0; \} 100% \{ background-position: -130% 0; \} \}
@keyframes lab-progress-aurora \{ to \{ filter: hue-rotate(360deg); \} \}
@keyframes lab-progress-breathe \{ 0%, 100% \{ transform: scaleY(.8); filter: brightness(.92); \} 50% \{ transform: scaleY(1); filter: brightness(1.22); \} \}
@keyframes lab-progress-ticker \{ to \{ background-position: 24px 0; \} \}
@keyframes lab-progress-spark \{ 0%, 100% \{ opacity: 1; \} 40% \{ opacity: .5; \} 55% \{ opacity: .85; \} \}
@keyframes lab-progress-wave \{ 0%, 100% \{ background-position: 0 0; \} 50% \{ background-position: 0 100%; \} \}

/* ---------- Spinner: shapes ---------- */
.sui-spinner.lab-s-ring \{ border-color: rgba(76,92,247,.22); border-top-color: #4c5cf7; border-bottom-color: transparent; \}
.sui-spinner.lab-s-dots \{ width: 30px; height: 30px; border: 0; border-radius: 50%; background: radial-gradient(circle at 50% 8%, #4c5cf7 3.5px, transparent 4px), radial-gradient(circle at 88% 78%, #6366f1 3.5px, transparent 4px), radial-gradient(circle at 12% 78%, #22d3ee 3.5px, transparent 4px); \}
.sui-spinner.lab-s-bars \{ width: 28px; height: 24px; border: 0; border-radius: 0; background: linear-gradient(#4c5cf7, #4c5cf7) 0 100%/5px 50% no-repeat, linear-gradient(#6366f1, #6366f1) 50% 100%/5px 80% no-repeat, linear-gradient(#22d3ee, #22d3ee) 100% 100%/5px 40% no-repeat; \}
.sui-spinner.lab-s-orbit \{ border: 2px solid rgba(76,92,247,.25); border-top-color: transparent; border-bottom-color: transparent; background: transparent; position: relative; \}
.sui-spinner.lab-s-orbit::before \{ content: \"\"; position: absolute; top: -3px; left: 50%; width: 8px; height: 8px; margin-left: -4px; border-radius: 50%; background: #4c5cf7; box-shadow: 0 0 8px #4c5cf7; \}
.sui-spinner.lab-s-comet \{ border: 0; background: conic-gradient(from 0deg, transparent 0 55%, #22d3ee 78%, #4c5cf7 100%); -webkit-mask: radial-gradient(farthest-side, transparent calc(100% - 5px), #000 calc(100% - 4px)); mask: radial-gradient(farthest-side, transparent calc(100% - 5px), #000 calc(100% - 4px)); \}
.sui-spinner.lab-s-pulse \{ border: 0; background: #4c5cf7; \}
.sui-spinner.lab-s-square \{ border-radius: 6px; border-color: rgba(76,92,247,.22); border-top-color: #4c5cf7; \}
/* ---------- Spinner: animations ---------- */
.sui-spinner.lab-s-anim-spin \{ animation: lab-spin calc(1s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, linear) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-pulse \{ animation: lab-spin-pulse calc(1.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-flip \{ animation: lab-spin-flip calc(1.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-reverse \{ animation: lab-spin-reverse calc(1.1s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, linear) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-orbit \{ animation: lab-spin-orbit calc(1.2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, linear) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-wobble \{ animation: lab-spin-wobble calc(1.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-bounce \{ animation: lab-spin-bounce calc(.9s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-ripple \{ animation: lab-spin-ripple calc(1.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-swing \{ animation: lab-spin-swing calc(1.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-anim-none \{ animation: none; \}
@keyframes lab-spin \{ to \{ transform: rotate(360deg); \} \}
@keyframes lab-spin-reverse \{ to \{ transform: rotate(-360deg); \} \}
@keyframes lab-spin-pulse \{ 0%, 100% \{ transform: scale(.68); opacity: .55; \} 50% \{ transform: scale(1); opacity: 1; \} \}
@keyframes lab-spin-flip \{ to \{ transform: rotateY(360deg); \} \}
@keyframes lab-spin-orbit \{ 0% \{ transform: rotate(0deg) translateX(0); \} 50% \{ transform: rotate(180deg) translateX(3px); \} 100% \{ transform: rotate(360deg) translateX(0); \} \}
@keyframes lab-spin-wobble \{ 0%, 100% \{ transform: rotate(-14deg) scale(.94); \} 50% \{ transform: rotate(14deg) scale(1.06); \} \}
@keyframes lab-spin-bounce \{ 0%, 100% \{ transform: translateY(-16%) scale(.94); \} 55% \{ transform: translateY(16%) scale(1.04); \} \}
@keyframes lab-spin-ripple \{ 0% \{ transform: scale(.66); opacity: .45; \} 70% \{ transform: scale(1.04); opacity: 1; \} 100% \{ transform: scale(.66); opacity: .45; \} \}
@keyframes lab-spin-swing \{ 0%, 100% \{ transform: rotate(-38deg); \} 50% \{ transform: rotate(38deg); \} \}

/* ---------- Badge: styles ---------- */
.sui-badge.lab-b-pill \{ background: #252b45; color: #a5b4fc; border: 1px solid #3b4470; border-radius: 999px; padding: 3px 10px; \}
.sui-badge.lab-b-tag \{ background: #f59e0b; color: #1a1206; border-radius: 4px; text-transform: uppercase; letter-spacing: .08em; font-weight: 700; padding: 3px 7px; \}
.sui-badge.lab-b-outline \{ background: transparent; color: #93c5fd; border: 1px solid #3b82f6; border-radius: 999px; padding: 3px 10px; \}
.sui-badge.lab-b-neon \{ background: #07131a; color: #22d3ee; border: 1px solid #22d3ee; border-radius: 6px; box-shadow: 0 0 12px rgba(34,211,238,.5); padding: 3px 10px; \}
.sui-badge.lab-b-status \{ background: #052e1b; color: #86efac; border-radius: 999px; padding: 4px 10px 4px 22px; position: relative; \}
.sui-badge.lab-b-status::before \{ content: \"\"; position: absolute; left: 8px; top: 50%; width: 7px; height: 7px; margin-top: -3.5px; border-radius: 50%; background: #22c55e; box-shadow: 0 0 6px #22c55e; \}
.sui-badge.lab-b-brutalist \{ background: #facc15; color: #111; border: 2px solid #111; border-radius: 0; box-shadow: 3px 3px 0 #111; padding: 3px 9px; font-weight: 700; \}
.sui-badge.lab-b-glass \{ background: rgba(148,163,184,.18); color: #e2e8f0; border: 1px solid rgba(255,255,255,.25); border-radius: 999px; backdrop-filter: blur(8px); padding: 3px 10px; \}
/* ---------- Badge: animations ---------- */
.lab-b-anim-pulse \{ animation: lab-badge-pulse calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-glow \{ animation: lab-badge-glow calc(2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-float \{ animation: lab-badge-float calc(2.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-pop \{ animation: lab-badge-pop calc(.42s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.4,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-b-anim-blink \{ animation: lab-badge-blink calc(1.2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, steps(1,end)) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-tilt \{ animation: lab-badge-tilt calc(2.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-ping \{ animation: lab-badge-ping calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-bounce \{ animation: lab-badge-bounce calc(1.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-jelly \{ animation: lab-badge-jelly calc(2.2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-sheen \{ animation: lab-badge-sheen calc(2.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-b-anim-dot \{ animation: lab-badge-dot calc(.55s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.3,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-b-anim-border \{ animation: lab-badge-border calc(.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-b-anim-none \{ animation: none; \}
@keyframes lab-badge-pulse \{ 0%, 100% \{ transform: scale(1); \} 50% \{ transform: scale(1.06); \} \}
@keyframes lab-badge-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.35); \} \}
@keyframes lab-badge-float \{ 0%, 100% \{ transform: translateY(0); \} 50% \{ transform: translateY(-3px); \} \}
@keyframes lab-badge-pop \{ 0% \{ opacity: 0; scale: .6; \} 65% \{ opacity: 1; scale: 1.1; \} 100% \{ opacity: 1; scale: 1; \} \}
@keyframes lab-badge-blink \{ 0%, 49% \{ opacity: 1; \} 50%, 100% \{ opacity: .35; \} \}
@keyframes lab-badge-tilt \{ 0%, 100% \{ rotate: -3deg; \} 50% \{ rotate: 3deg; \} \}
@keyframes lab-badge-ping \{ 0% \{ box-shadow: 0 0 0 0 rgba(74,222,128,.6); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(74,222,128,0); \} \}
@keyframes lab-badge-bounce \{ 0%, 100% \{ translate: 0 0; \} 40% \{ translate: 0 -4px; \} 60% \{ translate: 0 1px; \} \}
@keyframes lab-badge-jelly \{ 0%, 100% \{ scale: 1 1; \} 30% \{ scale: 1.12 .88; \} 55% \{ scale: .92 1.08; \} 80% \{ scale: 1.03 .97; \} \}
@keyframes lab-badge-sheen \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.45) saturate(1.15); \} \}
@keyframes lab-badge-dot \{ 0% \{ opacity: 0; scale: .2; border-radius: 50%; \} 60% \{ opacity: 1; scale: 1.15 .78; border-radius: 999px; \} 100% \{ scale: 1 1; \} \}
@keyframes lab-badge-border \{ 0% \{ background-color: transparent; box-shadow: inset 0 0 0 2px currentColor; \} 50% \{ background-color: transparent; box-shadow: inset 0 0 0 2px currentColor; \} 100% \{ box-shadow: none; \} \}

/* ---------- Alert: styles ---------- */
.sui-alert.lab-a-bar \{ background: #1b2230; color: #dbeafe; border: 0; border-left: 5px solid #3b82f6; border-radius: 4px; \}
.sui-alert.lab-a-filled \{ background: #7f1d1d; color: #fee2e2; border: 0; border-radius: 10px; \}
.sui-alert.lab-a-outline \{ background: transparent; color: #93c5fd; border: 1px dashed #3b82f6; border-radius: 10px; \}
.sui-alert.lab-a-glass \{ background: rgba(148,163,184,.16); color: #e2e8f0; border: 1px solid rgba(255,255,255,.22); border-radius: 12px; backdrop-filter: blur(10px); \}
.sui-alert.lab-a-brutalist \{ background: #facc15; color: #111; border: 2px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; \}
.sui-alert.lab-a-info \{ background: #0c2340; color: #bae6fd; border: 1px solid #0ea5e9; border-radius: 10px; \}
/* ---------- Alert: animations ---------- */
.lab-a-anim-fade \{ animation: lab-alert-in calc(.24s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-a-anim-slide \{ animation: lab-alert-slide calc(.26s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-a-anim-shake \{ animation: lab-alert-shake calc(.5s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s); \}
.lab-a-anim-pulse \{ animation: lab-alert-pulse calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-a-anim-zoom \{ animation: lab-alert-zoom calc(.3s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.2,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-a-anim-bounce-in \{ animation: lab-alert-bounce-in calc(.46s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.3,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-a-anim-slide-right \{ animation: lab-alert-slide-right calc(.28s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-a-anim-wobble \{ animation: lab-alert-wobble calc(2.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-a-anim-breathe \{ animation: lab-alert-breathe calc(2.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-a-anim-flash \{ animation: lab-alert-flash calc(1.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-a-anim-wipe \{ animation: lab-alert-wipe calc(.55s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-a-anim-border \{ animation: lab-alert-border calc(.7s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-a-anim-none \{ animation: none; \}
@keyframes lab-alert-in \{ from \{ opacity: 0; \} to \{ opacity: 1; \} \}
@keyframes lab-alert-slide \{ from \{ opacity: 0; transform: translateY(-6px); \} to \{ opacity: 1; transform: translateY(0); \} \}
@keyframes lab-alert-shake \{ 0%, 100% \{ transform: translateX(0); \} 20% \{ transform: translateX(-4px); \} 40% \{ transform: translateX(4px); \} 60% \{ transform: translateX(-2px); \} 80% \{ transform: translateX(2px); \} \}
@keyframes lab-alert-pulse \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(127,58,61,0); \} 50% \{ box-shadow: 0 0 0 5px rgba(127,58,61,.32); \} \}
@keyframes lab-alert-zoom \{ from \{ opacity: 0; scale: .92; \} to \{ opacity: 1; scale: 1; \} \}
@keyframes lab-alert-bounce-in \{ 0% \{ opacity: 0; translate: 0 -10px; scale: .97; \} 60% \{ opacity: 1; translate: 0 2px; scale: 1.005; \} 100% \{ opacity: 1; translate: 0 0; scale: 1; \} \}
@keyframes lab-alert-slide-right \{ from \{ opacity: 0; translate: 14px 0; \} to \{ opacity: 1; translate: 0 0; \} \}
@keyframes lab-alert-wobble \{ 0%, 100% \{ rotate: 0deg; \} 25% \{ rotate: -1.4deg; \} 75% \{ rotate: 1.4deg; \} \}
@keyframes lab-alert-breathe \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 7px #f59e0b); \} \}
@keyframes lab-alert-flash \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.25); \} \}
@keyframes lab-alert-wipe \{ from \{ clip-path: inset(0 0 0 100%); \} to \{ clip-path: inset(0 0 0 0); \} \}
@keyframes lab-alert-border \{ 0% \{ background-color: transparent; box-shadow: inset 0 0 0 2px currentColor; \} 50% \{ background-color: transparent; box-shadow: inset 0 0 0 2px currentColor; \} 100% \{ box-shadow: none; \} \}

/* ---------- Card: styles ---------- */
.sui-card.lab-c-elevated \{ background: #161d2b; border: 1px solid #2a3446; border-radius: 14px; box-shadow: 0 14px 34px rgba(0,0,0,.4); \}
.sui-card.lab-c-flat \{ background: #161d2b; border: 0; border-radius: 4px; box-shadow: none; \}
.sui-card.lab-c-outline \{ background: transparent; border: 1px solid #3b82f6; border-radius: 14px; box-shadow: none; \}
.sui-card.lab-c-glass \{ background: rgba(148,163,184,.14); border: 1px solid rgba(255,255,255,.22); border-radius: 16px; backdrop-filter: blur(12px); \}
.sui-card.lab-c-neon \{ background: #0b1220; border: 1px solid #22d3ee; border-radius: 14px; box-shadow: 0 0 22px rgba(34,211,238,.35); \}
.sui-card.lab-c-brutalist \{ background: #fde68a; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 6px 6px 0 #111; \}
.sui-card.lab-c-dashed \{ background: #0f172a; border: 2px dashed #475569; border-radius: 10px; box-shadow: none; \}
/* ---------- Card: animations ---------- */
.lab-c-anim-fade \{ animation: lab-card-in calc(.28s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-c-anim-float \{ animation: lab-card-float calc(3.2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-c-anim-rise \{ animation: lab-card-rise calc(.38s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1,.3,1)) var(--lab-motion-delay, 0s); \}
.lab-c-anim-zoom \{ animation: lab-card-zoom calc(.34s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.2,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-c-anim-tilt \{ animation: lab-card-tilt calc(.45s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1,.3,1)) var(--lab-motion-delay, 0s); \}
.lab-c-anim-halo \{ animation: lab-card-halo calc(3s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-c-anim-drift \{ animation: lab-card-drift calc(5s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-c-anim-unfold \{ animation: lab-card-unfold calc(.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1,.3,1)) var(--lab-motion-delay, 0s); \}
.lab-c-anim-dot \{ animation: lab-card-dot calc(.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.2,.3,1)) var(--lab-motion-delay, 0s); \}
.lab-c-anim-border \{ animation: lab-card-border calc(.7s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-c-anim-wipe \{ animation: lab-card-wipe calc(.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-c-anim-flip \{ animation: lab-card-flip calc(.7s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1,.3,1)) var(--lab-motion-delay, 0s); \}
.lab-c-anim-squash \{ animation: lab-card-squash calc(.5s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.3,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-c-anim-none \{ animation: none; \}
@keyframes lab-card-in \{ from \{ opacity: 0; transform: translateY(6px); \} to \{ opacity: 1; transform: translateY(0); \} \}
@keyframes lab-card-float \{ 0%, 100% \{ transform: translateY(0); \} 50% \{ transform: translateY(-5px); \} \}
@keyframes lab-card-rise \{ from \{ opacity: 0; translate: 0 14px; \} to \{ opacity: 1; translate: 0 0; \} \}
@keyframes lab-card-zoom \{ from \{ opacity: 0; scale: .94; \} to \{ opacity: 1; scale: 1; \} \}
@keyframes lab-card-tilt \{ from \{ opacity: 0; transform: perspective(600px) rotateX(8deg) translateY(10px); \} to \{ opacity: 1; transform: perspective(600px) rotateX(0) translateY(0); \} \}
@keyframes lab-card-halo \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,0); \} 50% \{ box-shadow: 0 0 26px 2px rgba(76,92,247,.4); \} \}
@keyframes lab-card-drift \{ 0%, 100% \{ translate: 0 0; \} 50% \{ translate: 3px -4px; \} \}
@keyframes lab-card-unfold \{ from \{ opacity: 0; transform: scaleY(.72); transform-origin: top center; \} to \{ opacity: 1; transform: scaleY(1); \} \}
@keyframes lab-card-dot \{ 0% \{ opacity: 0; scale: .22; border-radius: 50%; \} 55% \{ opacity: 1; scale: 1.02 .55; border-radius: 999px; \} 100% \{ opacity: 1; scale: 1; \} \}
@keyframes lab-card-border \{ 0% \{ background-color: transparent; border-color: #60a5fa; box-shadow: inset 0 0 0 2px #60a5fa; \} 48% \{ background-color: transparent; border-color: #60a5fa; box-shadow: inset 0 0 0 2px #60a5fa; \} 100% \{ box-shadow: none; \} \}
@keyframes lab-card-wipe \{ from \{ clip-path: inset(0 100% 0 0); \} to \{ clip-path: inset(0 0 0 0); \} \}
@keyframes lab-card-flip \{ from \{ opacity: .2; transform: perspective(900px) rotateY(-75deg); \} to \{ opacity: 1; transform: perspective(900px) rotateY(0); \} \}
@keyframes lab-card-squash \{ 0% \{ opacity: .3; scale: 1 .35; \} 60% \{ scale: 1 1.06; \} 100% \{ opacity: 1; scale: 1 1; \} \}

/* ---------- Button: styles ---------- */
.sui-button.lab-btn-solid \{ background: #2d3340; color: #e6e6e6; border: 1px solid #3d4556; border-radius: 8px; \}
.sui-button.lab-btn-gradient \{ background: linear-gradient(135deg, #4c5cf7, #a855f7); color: #fff; border: 0; border-radius: 10px; \}
.sui-button.lab-btn-outline \{ background: transparent; color: #93c5fd; border: 1px solid #3b82f6; border-radius: 8px; \}
.sui-button.lab-btn-ghost \{ background: transparent; color: #cbd5e1; border: 0; border-radius: 8px; \}
.sui-button.lab-btn-ghost:hover \{ background: #1f2937; \}
.sui-button.lab-btn-neon \{ background: #07131a; color: #22d3ee; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 14px rgba(34,211,238,.4); \}
.sui-button.lab-btn-3d \{ background: #4c5cf7; color: #fff; border: 0; border-radius: 10px; box-shadow: 0 5px 0 #1e2a78; transform: translateY(-1px); \}
.sui-button.lab-btn-3d:active \{ transform: translateY(4px); box-shadow: 0 1px 0 #1e2a78; \}
.sui-button.lab-btn-pill \{ background: #22c55e; color: #052e16; border: 0; border-radius: 999px; font-weight: 600; \}
.sui-button.lab-btn-brutalist \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 5px 5px 0 #111; font-weight: 700; \}
/* ---------- Button: animations ---------- */
.lab-btn-anim-pulse \{ animation: lab-button-pulse calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-float \{ animation: lab-button-float calc(2.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-glow \{ animation: lab-button-glow calc(2.2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-jelly \{ animation: lab-button-jelly calc(2.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-bounce \{ animation: lab-button-bounce calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-tilt \{ animation: lab-button-tilt calc(3s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-ring \{ animation: lab-button-ring calc(2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-sheen \{ animation: lab-button-sheen calc(2.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-press \{ animation: lab-button-press calc(2s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-btn-anim-border \{ animation: lab-button-border calc(.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-btn-anim-wipe \{ animation: lab-button-wipe calc(.55s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-btn-anim-pop \{ animation: lab-button-pop calc(.45s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.5,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-btn-anim-none \{ animation: none; \}
@keyframes lab-button-pulse \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.45); \} 50% \{ box-shadow: 0 0 0 6px rgba(76,92,247,0); \} \}
@keyframes lab-button-float \{ 0%, 100% \{ transform: translateY(0); \} 50% \{ transform: translateY(-3px); \} \}
@keyframes lab-button-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.25); \} \}
@keyframes lab-button-jelly \{ 0%, 100% \{ scale: 1 1; \} 30% \{ scale: 1.06 .94; \} 55% \{ scale: .95 1.05; \} 80% \{ scale: 1.02 .98; \} \}
@keyframes lab-button-bounce \{ 0%, 100% \{ translate: 0 0; \} 45% \{ translate: 0 -3px; \} 60% \{ translate: 0 1px; \} \}
@keyframes lab-button-tilt \{ 0%, 100% \{ rotate: -1.5deg; \} 50% \{ rotate: 1.5deg; \} \}
@keyframes lab-button-ring \{ 0% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.5); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(76,92,247,0); \} \}
@keyframes lab-button-sheen \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.3) saturate(1.1); \} \}
@keyframes lab-button-press \{ 0%, 100% \{ scale: 1; \} 50% \{ scale: .97; \} \}
@keyframes lab-button-border \{ 0% \{ background-color: transparent; box-shadow: inset 0 0 0 2px #60a5fa; \} 50% \{ background-color: transparent; box-shadow: inset 0 0 0 2px #60a5fa; \} 100% \{ box-shadow: none; \} \}
@keyframes lab-button-wipe \{ from \{ clip-path: inset(0 100% 0 0); \} to \{ clip-path: inset(0 0 0 0); \} \}
@keyframes lab-button-pop \{ 0% \{ opacity: 0; scale: .3; border-radius: 50%; \} 60% \{ opacity: 1; scale: 1.08; \} 100% \{ scale: 1; \} \}

/* ---------- Tabs: styles ---------- */
.sui-tabs.lab-t-underline \{ background: transparent; border: 0; \}
.sui-tabs.lab-t-underline .sui-tab \{ background: transparent; color: #94a3b8; border-radius: 0; \}
.sui-tabs.lab-t-underline .sui-tab.active \{ background: transparent; color: #e2e8f0; box-shadow: inset 0 -2px 0 #4c5cf7; \}
.sui-tabs.lab-t-pills \{ background: #0f172a; border: 1px solid #1e293b; border-radius: 999px; padding: 4px; \}
.sui-tabs.lab-t-pills .sui-tab \{ border-radius: 999px; padding: 6px 14px; color: #94a3b8; \}
.sui-tabs.lab-t-pills .sui-tab.active \{ background: #4c5cf7; color: #fff; box-shadow: none; \}
.sui-tabs.lab-t-segmented \{ background: #0b1220; border: 1px solid #1e293b; border-radius: 8px; padding: 3px; \}
.sui-tabs.lab-t-segmented .sui-tab \{ border-radius: 6px; color: #94a3b8; \}
.sui-tabs.lab-t-segmented .sui-tab.active \{ background: #1e293b; color: #e2e8f0; box-shadow: none; \}
.sui-tabs.lab-t-boxed \{ background: #111827; border: 1px solid #334155; border-radius: 8px; \}
.sui-tabs.lab-t-boxed .sui-tab \{ border-radius: 1px; color: #94a3b8; \}
.sui-tabs.lab-t-boxed .sui-tab.active \{ background: #1e293b; color: #fff; box-shadow: none; \}
.sui-tabs.lab-t-brutalist \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; \}
.sui-tabs.lab-t-brutalist .sui-tab \{ border-radius: 0; color: #111; font-weight: 700; \}
.sui-tabs.lab-t-brutalist .sui-tab.active \{ background: #111; color: #facc15; box-shadow: none; \}
.sui-tabs.lab-t-neon \{ background: #07131a; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 18px rgba(34,211,238,.25); \}
.sui-tabs.lab-t-neon .sui-tab \{ border-radius: 6px; color: #67e8f9; \}
.sui-tabs.lab-t-neon .sui-tab.active \{ background: #0e7490; color: #ecfeff; box-shadow: none; \}
/* ---------- Tabs: animations ---------- */
.lab-t-anim-pop .sui-tab.active \{ animation: lab-tab-pop calc(.32s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, cubic-bezier(.2,1.4,.4,1)) var(--lab-motion-delay, 0s); \}
.lab-t-anim-slide .sui-tab.active \{ animation: lab-tab-slide calc(.26s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-t-anim-glow .sui-tab.active \{ animation: lab-tab-glow calc(2.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.lab-t-anim-sweep .sui-tab.active \{ animation: lab-tab-sweep calc(.45s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.lab-t-anim-none .sui-tab.active \{ animation: none; \}
@keyframes lab-tab-pop \{ 0% \{ scale: .95; \} 60% \{ scale: 1.03; \} 100% \{ scale: 1; \} \}
@keyframes lab-tab-slide \{ from \{ opacity: .35; translate: -6px 0; \} to \{ opacity: 1; translate: 0 0; \} \}
@keyframes lab-tab-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.22); \} \}
@keyframes lab-tab-sweep \{ from \{ background-size: 0% 100%; \} to \{ background-size: 100% 100%; \} \}

/* ---------- Input: styles ---------- */
.sui-input.lab-i-outline \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 8px; \}
.sui-input.lab-i-filled \{ background: #1e293b; color: #e2e8f0; border: 1px solid transparent; border-radius: 8px; \}
.sui-input.lab-i-underline \{ background: transparent; color: #e2e8f0; border: 0; border-bottom: 2px solid #475569; border-radius: 0; \}
.sui-input.lab-i-pill \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 999px; padding-left: 16px; padding-right: 16px; \}
.sui-input.lab-i-glass \{ background: rgba(148,163,184,.14); color: #e2e8f0; border: 1px solid rgba(255,255,255,.24); border-radius: 10px; backdrop-filter: blur(8px); \}
.sui-input.lab-i-brutalist \{ background: #fff; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; \}
/* ---------- Input: focus animations ---------- */
.sui-input.lab-i-anim-ring:focus \{ animation: lab-focus-ring calc(.5s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.sui-input.lab-i-anim-glow:focus \{ animation: lab-focus-glow calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-input.lab-i-anim-lift:focus \{ animation: lab-focus-lift calc(.3s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-out) var(--lab-motion-delay, 0s); \}
.sui-input.lab-i-anim-shake:focus \{ animation: lab-focus-shake calc(.4s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s); \}
.sui-input.lab-i-anim-none \{ animation: none; \}
@keyframes lab-focus-ring \{ 0% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.45); \} 100% \{ box-shadow: 0 0 0 5px rgba(76,92,247,0); \} \}
@keyframes lab-focus-glow \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 5px rgba(76,92,247,.55)); \} \}
@keyframes lab-focus-lift \{ 50% \{ translate: 0 -1px; \} \}
@keyframes lab-focus-shake \{ 0%, 100% \{ translate: 0 0; \} 25% \{ translate: -2px 0; \} 75% \{ translate: 2px 0; \} \}

/* ---------- Dropdown styles: trigger and panel stay coherent ---------- */
.lab-dd-classic .sui-menu-button \{ background: #2d3340; color: #e6e6e6; border: 1px solid #3d4556; border-radius: 8px; \}
.lab-dd-classic .sui-menu-popup \{ background: #1d2129; color: #e6e6e6; border: 1px solid #3d4556; border-radius: 8px; box-shadow: 0 18px 40px rgba(0,0,0,.45); \}
.lab-dd-classic .sui-menu-button:hover, .lab-dd-classic .sui-menu-button:focus-visible, .lab-dd-classic.sui-menu.open > .sui-menu-button \{ background: #3a4356; color: #fff; \}
.lab-dd-classic .sui-menuitem:hover, .lab-dd-classic .sui-menuitem:focus \{ background: #2a3143; color: #fff; \}

.lab-dd-pill .sui-menu-button \{ background: #4c5cf7; color: #fff; border: 0; border-radius: 999px; padding: 7px 16px; \}
.lab-dd-pill .sui-menu-popup \{ background: #111827; color: #e2e8f0; border: 1px solid #334155; border-radius: 16px; padding: 6px; box-shadow: 0 18px 40px rgba(0,0,0,.45); \}
.lab-dd-pill .sui-menuitem \{ border-radius: 999px; color: #e2e8f0; \}
.lab-dd-pill .sui-menu-button:hover, .lab-dd-pill .sui-menu-button:focus-visible, .lab-dd-pill.sui-menu.open > .sui-menu-button \{ background: #6366f1; color: #fff; \}
.lab-dd-pill .sui-menuitem:hover, .lab-dd-pill .sui-menuitem:focus \{ background: #1e293b; color: #fff; \}

.lab-dd-neon .sui-menu-button \{ background: #07131a; color: #22d3ee; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 12px rgba(34,211,238,.4); \}
.lab-dd-neon .sui-menu-popup \{ background: #07131a; color: #a5f3fc; border: 1px solid #22d3ee; border-radius: 10px; box-shadow: 0 0 22px rgba(34,211,238,.35); \}
.lab-dd-neon .sui-menuitem \{ color: #a5f3fc; \}
.lab-dd-neon .sui-menu-button:hover, .lab-dd-neon .sui-menu-button:focus-visible, .lab-dd-neon.sui-menu.open > .sui-menu-button \{ background: #0b1f2a; color: #67e8f9; box-shadow: 0 0 18px rgba(34,211,238,.6); \}
.lab-dd-neon .sui-menuitem:hover, .lab-dd-neon .sui-menuitem:focus \{ background: #0e7490; color: #ecfeff; \}

.lab-dd-brutalist .sui-menu-button \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; font-weight: 700; \}
.lab-dd-brutalist .sui-menu-popup \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 6px 6px 0 #111; \}
.lab-dd-brutalist .sui-menuitem \{ color: #111; border-radius: 0; \}
.lab-dd-brutalist .sui-menu-button:hover, .lab-dd-brutalist .sui-menu-button:focus-visible, .lab-dd-brutalist.sui-menu.open > .sui-menu-button \{ background: #fde047; color: #111; \}
.lab-dd-brutalist .sui-menuitem:hover, .lab-dd-brutalist .sui-menuitem:focus \{ background: #111; color: #facc15; \}

.lab-dd-glass .sui-menu-button \{ background: rgba(148,163,184,.16); color: #e2e8f0; border: 1px solid rgba(255,255,255,.24); border-radius: 10px; backdrop-filter: blur(10px); \}
.lab-dd-glass .sui-menu-popup \{ background: rgba(15,23,42,.78); color: #e2e8f0; border: 1px solid rgba(255,255,255,.24); border-radius: 14px; backdrop-filter: blur(16px); box-shadow: 0 18px 40px rgba(0,0,0,.45); \}
.lab-dd-glass .sui-menuitem \{ color: #e2e8f0; \}
.lab-dd-glass .sui-menu-button:hover, .lab-dd-glass .sui-menu-button:focus-visible, .lab-dd-glass.sui-menu.open > .sui-menu-button \{ background: rgba(148,163,184,.28); color: #fff; \}
.lab-dd-glass .sui-menuitem:hover, .lab-dd-glass .sui-menuitem:focus \{ background: rgba(255,255,255,.12); color: #fff; \}

.lab-dd-outline .sui-menu-button \{ background: transparent; color: #93c5fd; border: 1px solid #3b82f6; border-radius: 8px; \}
.lab-dd-outline .sui-menu-popup \{ background: #0f172a; color: #dbeafe; border: 1px solid #3b82f6; border-radius: 10px; box-shadow: 0 18px 40px rgba(0,0,0,.4); \}
.lab-dd-outline .sui-menuitem \{ color: #dbeafe; \}
.lab-dd-outline .sui-menu-button:hover, .lab-dd-outline .sui-menu-button:focus-visible, .lab-dd-outline.sui-menu.open > .sui-menu-button \{ background: rgba(59,130,246,.18); color: #dbeafe; \}
.lab-dd-outline .sui-menuitem:hover, .lab-dd-outline .sui-menuitem:focus \{ background: #1e293b; color: #fff; \}
/* ---------- Popover styles ---------- */
.lab-pop-card .sui-popover-panel \{ background: #161d2b; color: #e2e8f0; border: 1px solid #2a3446; border-radius: 12px; box-shadow: 0 16px 36px rgba(0,0,0,.45); \}
.lab-pop-glass .sui-popover-panel \{ background: rgba(148,163,184,.16); color: #e2e8f0; border: 1px solid rgba(255,255,255,.24); border-radius: 14px; backdrop-filter: blur(12px); \}
.lab-pop-neon .sui-popover-panel \{ background: #07131a; color: #a5f3fc; border: 1px solid #22d3ee; border-radius: 12px; box-shadow: 0 0 22px rgba(34,211,238,.4); \}
.lab-pop-brutalist .sui-popover-panel \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 6px 6px 0 #111; \}
/* ---------- Combobox styles ---------- */
.lab-cb-classic .sui-combobox-input \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 8px; \}
.lab-cb-pill .sui-combobox-input \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 999px; padding-left: 16px; padding-right: 16px; \}
.lab-cb-neon .sui-combobox-input \{ background: #07131a; color: #a5f3fc; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 12px rgba(34,211,238,.35); \}
.lab-cb-brutalist .sui-combobox-input \{ background: #fff; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; \}

/* ---------- Reduced motion ---------- */
@media (prefers-reduced-motion: reduce) \{
  .lab-b-anim-pulse, .lab-b-anim-glow, .lab-b-anim-float, .lab-b-anim-blink, .lab-b-anim-tilt, .lab-b-anim-ping, .lab-b-anim-bounce, .lab-b-anim-jelly, .lab-b-anim-sheen, .lab-b-anim-dot, .lab-b-anim-border,
  .lab-a-anim-pulse, .lab-a-anim-wobble, .lab-a-anim-breathe, .lab-a-anim-flash, .lab-a-anim-wipe, .lab-a-anim-border,
  .lab-c-anim-float, .lab-c-anim-halo, .lab-c-anim-drift, .lab-c-anim-dot, .lab-c-anim-border, .lab-c-anim-wipe, .lab-c-anim-flip, .lab-c-anim-squash,
  .lab-btn-anim-pulse, .lab-btn-anim-float, .lab-btn-anim-glow, .lab-btn-anim-jelly, .lab-btn-anim-bounce, .lab-btn-anim-tilt, .lab-btn-anim-ring, .lab-btn-anim-sheen, .lab-btn-anim-press, .lab-btn-anim-border, .lab-btn-anim-wipe, .lab-btn-anim-pop,
  .lab-t-anim-glow .sui-tab.active,
  .sui-spinner.lab-s-anim-spin, .sui-spinner.lab-s-anim-pulse, .sui-spinner.lab-s-anim-flip, .sui-spinner.lab-s-anim-reverse, .sui-spinner.lab-s-anim-orbit, .sui-spinner.lab-s-anim-wobble, .sui-spinner.lab-s-anim-bounce, .sui-spinner.lab-s-anim-ripple, .sui-spinner.lab-s-anim-swing,
  .sui-input.lab-i-anim-glow:focus \{ animation: none !important; \}
  .sui-progress.lab-p-anim-pulse::-webkit-progress-value, .sui-progress.lab-p-anim-stripes::-webkit-progress-value, .sui-progress.lab-p-anim-shimmer::-webkit-progress-value, .sui-progress.lab-p-anim-glow::-webkit-progress-value, .sui-progress.lab-p-anim-indeterminate::-webkit-progress-value, .sui-progress.lab-p-anim-aurora::-webkit-progress-value, .sui-progress.lab-p-anim-breathe::-webkit-progress-value, .sui-progress.lab-p-anim-ticker::-webkit-progress-value, .sui-progress.lab-p-anim-spark::-webkit-progress-value, .sui-progress.lab-p-anim-wave::-webkit-progress-value \{ animation: none !important; \}
\}


/* ---------- Checkbox: genuinely different structures ---------- */
.sui-checkbox.lab-check-classic \{ width: 20px; height: 20px; border-radius: 5px; border: 1.5px solid #475569; background: #0f172a; \}
.sui-checkbox.lab-check-classic:checked \{ background: var(--lab-cb-color, #4c5cf7); border-color: var(--lab-cb-color, #4c5cf7); \}
.sui-checkbox.lab-check-classic::after \{ width: 5px; height: 10px; border: solid #fff; border-width: 0 2px 2px 0; background: none; border-radius: 0; transform: rotate(45deg); left: 6px; top: 2px; opacity: 0; \}
.sui-checkbox.lab-check-classic:checked::after \{ opacity: 1; \}

.sui-checkbox.lab-check-round \{ width: 22px; height: 22px; border-radius: 50%; border: 2px solid #64748b; background: transparent; \}
.sui-checkbox.lab-check-round:checked \{ background: var(--lab-cb-color, #4c5cf7); border-color: var(--lab-cb-color, #4c5cf7); box-shadow: 0 0 0 4px rgba(76,92,247,.18); \}
.sui-checkbox.lab-check-round::after \{ width: 5px; height: 10px; border: solid #fff; border-width: 0 2px 2px 0; background: none; border-radius: 0; transform: rotate(45deg); left: 7px; top: 3px; opacity: 0; \}
.sui-checkbox.lab-check-round:checked::after \{ opacity: 1; \}

.sui-checkbox.lab-check-tile \{ width: 30px; height: 30px; border-radius: 9px; border: 0; background: #1e293b; \}
.sui-checkbox.lab-check-tile:checked \{ background: linear-gradient(135deg, var(--lab-cb-color, #4c5cf7), #a855f7); box-shadow: 0 6px 16px rgba(0,0,0,.35); \}
.sui-checkbox.lab-check-tile::after \{ width: 7px; height: 14px; border: solid #fff; border-width: 0 3px 3px 0; background: none; border-radius: 0; transform: rotate(45deg); left: 10px; top: 5px; opacity: 0; \}
.sui-checkbox.lab-check-tile:checked::after \{ opacity: 1; \}

.sui-checkbox.lab-check-led \{ width: 22px; height: 22px; border-radius: 6px; border: 1px solid #334155; background: #0b1220; \}
.sui-checkbox.lab-check-led:checked \{ border-color: var(--lab-cb-color, #22d3ee); background: #0b1220; \}
.sui-checkbox.lab-check-led::after \{ width: 8px; height: 8px; border: 0; border-radius: 50%; background: transparent; transform: none; left: 50%; top: 50%; margin: -4px 0 0 -4px; opacity: 0; \}
.sui-checkbox.lab-check-led:checked::after \{ opacity: 1; background: var(--lab-cb-color, #22d3ee); box-shadow: 0 0 10px var(--lab-cb-color, #22d3ee); \}

.sui-checkbox.lab-check-brutalist \{ width: 26px; height: 26px; border-radius: 0; border: 3px solid #111; background: #fff; box-shadow: 3px 3px 0 #111; \}
.sui-checkbox.lab-check-brutalist:checked \{ background: #facc15; border-color: #111; \}
.sui-checkbox.lab-check-brutalist::after \{ width: 7px; height: 13px; border: solid #111; border-width: 0 4px 4px 0; background: none; border-radius: 0; transform: rotate(45deg); left: 8px; top: 2px; opacity: 0; \}
.sui-checkbox.lab-check-brutalist:checked::after \{ opacity: 1; \}

.sui-checkbox.lab-check-neon \{ width: 22px; height: 22px; border-radius: 8px; border: 2px solid var(--lab-cb-color, #22d3ee); background: transparent; \}
.sui-checkbox.lab-check-neon:checked \{ background: rgba(34,211,238,.14); border-color: var(--lab-cb-color, #22d3ee); box-shadow: inset 0 0 10px var(--lab-cb-color, #22d3ee), 0 0 12px rgba(34,211,238,.45); \}
.sui-checkbox.lab-check-neon::after \{ width: 5px; height: 10px; border: solid var(--lab-cb-color, #22d3ee); border-width: 0 2px 2px 0; background: none; border-radius: 0; transform: rotate(45deg); left: 7px; top: 3px; opacity: 0; \}
.sui-checkbox.lab-check-neon:checked::after \{ opacity: 1; \}

.sui-checkbox.lab-check-sweep \{ width: 24px; height: 24px; border-radius: 6px; border: 1px solid #475569; background-color: #0f172a; background-image: linear-gradient(var(--lab-cb-color, #4c5cf7), var(--lab-cb-color, #4c5cf7)); background-repeat: no-repeat; background-position: bottom; background-size: 100% 0%; transition: background-size .28s ease-out, border-color .2s ease; \}
.sui-checkbox.lab-check-sweep:checked \{ border-color: var(--lab-cb-color, #4c5cf7); background-size: 100% 100%; \}
.sui-checkbox.lab-check-sweep::after \{ width: 5px; height: 10px; border: solid #fff; border-width: 0 2px 2px 0; background: none; border-radius: 0; transform: rotate(45deg); left: 8px; top: 4px; opacity: 0; \}
.sui-checkbox.lab-check-sweep:checked::after \{ opacity: 1; \}

.lab-cb-theme-accent \{ --lab-cb-color: #4c5cf7; \}
.lab-cb-theme-success \{ --lab-cb-color: #22c55e; \}
.lab-cb-theme-warning \{ --lab-cb-color: #f59e0b; \}
.lab-cb-theme-danger \{ --lab-cb-color: #ef4444; \}
.lab-cb-theme-mono \{ --lab-cb-color: #94a3b8; \}

/* ---------- Spinner: hourglass and drop ---------- */
.sui-spinner.lab-s-hourglass \{ width: 26px; height: 34px; border: 0; border-radius: 0; background: #1e293b; clip-path: polygon(0 0, 100% 0, 50% 50%, 100% 100%, 0 100%, 50% 50%); position: relative; animation: lab-hourglass-flip calc(2.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-hourglass::after \{ content: \"\"; position: absolute; left: 0; right: 0; top: 0; height: 50%; background: #4c5cf7; animation: lab-hourglass-sand calc(2.6s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-drop \{ width: 26px; height: 36px; border: 2px solid #334155; border-radius: 4px 4px 12px 12px; background-color: #0b1220; background-image: linear-gradient(#22d3ee, #4c5cf7); background-repeat: no-repeat; background-position: bottom; background-size: 100% 0%; position: relative; overflow: hidden; animation: lab-drop-fill calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
.sui-spinner.lab-s-drop::before \{ content: \"\"; position: absolute; top: -10px; left: 50%; width: 8px; height: 8px; margin-left: -4px; background: #22d3ee; border-radius: 50% 50% 50% 0; transform: rotate(45deg); animation: lab-drop-fall calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in) var(--lab-motion-delay, 0s) infinite; \}
@keyframes lab-hourglass-flip \{ 0%, 42% \{ transform: rotate(0deg); \} 55%, 100% \{ transform: rotate(180deg); \} \}
@keyframes lab-hourglass-sand \{ 0% \{ height: 50%; \} 42% \{ height: 0%; \} 55% \{ height: 0%; \} 95% \{ height: 50%; \} 100% \{ height: 50%; \} \}
@keyframes lab-drop-fill \{ 0% \{ background-size: 100% 0%; \} 65% \{ background-size: 100% 72%; \} 100% \{ background-size: 100% 100%; \} \}
@keyframes lab-drop-fall \{ 0% \{ top: -10px; opacity: 0; \} 15% \{ opacity: 1; \} 62% \{ top: 10px; opacity: 0; \} 100% \{ top: 10px; opacity: 0; \} \}

/* ---------- Dropdown: item highlight styles ---------- */
.lab-items-bar .sui-menuitem \{ position: relative; padding-left: 16px; \}
.lab-items-bar .sui-menuitem::before \{ content: \"\"; position: absolute; left: 5px; top: 50%; width: 3px; height: 0; border-radius: 2px; background: var(--sui-accent, #4c5cf7); transform: translateY(-50%); transition: height .18s ease; \}
.lab-items-bar .sui-menuitem:hover::before, .lab-items-bar .sui-menuitem:focus::before \{ height: 62%; \}

.lab-items-glow .sui-menuitem:hover, .lab-items-glow .sui-menuitem:focus \{ background: #1e2637; box-shadow: 0 0 0 1px #4c5cf7, 0 0 14px rgba(76,92,247,.55); \}

.lab-items-aura .sui-menuitem:hover, .lab-items-aura .sui-menuitem:focus \{ background: radial-gradient(circle at 18% 50%, rgba(76,92,247,.5), transparent 70%); animation: lab-item-aura calc(1.8s * var(--lab-motion-speed, 1)) var(--lab-motion-easing, ease-in-out) var(--lab-motion-delay, 0s) infinite; \}
@keyframes lab-item-aura \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,0); \} 50% \{ box-shadow: 0 0 16px 2px rgba(76,92,247,.45); \} \}

.lab-items-gradient .sui-menuitem \{ background-image: linear-gradient(90deg, #4c5cf7, #a855f7); background-size: 0% 100%; background-repeat: no-repeat; background-position: left; transition: background-size .25s ease; \}
.lab-items-gradient .sui-menuitem:hover, .lab-items-gradient .sui-menuitem:focus \{ background-size: 100% 100%; color: #fff; \}

.lab-items-border .sui-menuitem \{ border: 1px solid transparent; \}
.lab-items-border .sui-menuitem:hover, .lab-items-border .sui-menuitem:focus \{ border-color: #4c5cf7; box-shadow: inset 0 0 0 1px rgba(76,92,247,.45); \}

.lab-items-dot .sui-menuitem \{ position: relative; padding-left: 18px; \}
.lab-items-dot .sui-menuitem::before \{ content: \"\"; position: absolute; left: 6px; top: 50%; width: 6px; height: 6px; margin-top: -3px; border-radius: 50%; background: #4c5cf7; opacity: 0; transform: scale(.4); transition: opacity .15s ease, transform .15s ease; \}
.lab-items-dot .sui-menuitem:hover::before, .lab-items-dot .sui-menuitem:focus::before \{ opacity: 1; transform: scale(1); \}

.lab-items-check .sui-menuitem \{ position: relative; padding-right: 22px; \}
.lab-items-check .sui-menuitem::after \{ content: \"\\2713\"; position: absolute; right: 8px; top: 50%; transform: translateY(-50%) scale(.5); opacity: 0; color: #4c5cf7; transition: opacity .15s ease, transform .15s ease; \}
.lab-items-check .sui-menuitem:hover::after, .lab-items-check .sui-menuitem:focus::after \{ opacity: 1; transform: translateY(-50%) scale(1); \}

.lab-items-none .sui-menuitem \{ \}

/* ---------- Item highlight combinations ---------- */
.lab-items-cloud .sui-menuitem:hover, .lab-items-cloud .sui-menuitem:focus \{ background: radial-gradient(ellipse at 12% 50%, rgba(76,92,247,.62), transparent 68%); \}
.lab-items-underline .sui-menuitem \{ position: relative; \}
.lab-items-underline .sui-menuitem::after \{ content: \"\"; position: absolute; left: 8px; right: 8px; bottom: 4px; height: 2px; border-radius: 2px; background: var(--sui-accent, #4c5cf7); transform: scaleX(0); transform-origin: left; transition: transform .22s ease; \}
.lab-items-underline .sui-menuitem:hover::after, .lab-items-underline .sui-menuitem:focus::after \{ transform: scaleX(1); \}
.lab-items-bar.lab-items-underline .sui-menuitem::before \{ height: 0; \}
.lab-items-bar.lab-items-underline .sui-menuitem:hover::before, .lab-items-bar.lab-items-underline .sui-menuitem:focus::before \{ height: 0; \}

.lab-motion-card \{ grid-column: 1 / -1; \}
.lab-motion-card .lab-fields \{ grid-template-columns: repeat(3, minmax(130px, 1fr)); \}
.lab-motion-card button \{ min-height: 34px; touch-action: manipulation; \}
"

view main {
  title "Animation lab"
  div class: "lab-page" style: cssSafe("--lab-motion-speed: " + motionSpeed + (motionEase == "default" ? "" : "; --lab-motion-easing: " + motionEase) + (motionDelay == "0" or motionDelay == "0.0" ? "" : "; --lab-motion-delay: " + motionDelay + "s")) {
    h1 "Animation lab" class: "lab-head"
    p "Only controls whose graphics can be animated. Every style and every animation is chosen live; the page does not jump thanks to the re-render preservation in the runtime." class: "lab-note"

    div class: "lab-grid" {
      card class: "lab-card lab-motion-card" {
        h3 "Motion editor"
        p "Tune every animation live: speed multiplier, easing and delay. New keyframes are added at build time in lab.css." class: "lab-note"
        div class: "lab-fields" {
          number "Speed" value: motionSpeed min: 0.3 max: 2.5 step: 0.1
          select "Easing" value: motionEase {
            option "Default" value: "default"
            option "Linear" value: "linear"
            option "Ease" value: "ease"
            option "Ease in" value: "ease-in"
            option "Ease out" value: "ease-out"
            option "Ease in out" value: "ease-in-out"
          }
          number "Delay" value: motionDelay min: 0 max: 1 step: 0.05
          row class: "lab-wide" {
            button "Copy CSS" onClick: copyMotion
          }
        }
        p "{motionMsg}" class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Progress"
        div class: "lab-preview" {
          progress value: 62 max: 100 class: pStyle + " " + pAnim
        }
        div class: "lab-fields" {
          select "Style" value: pStyle {
          					option "Classic" value: "lab-p-classic"
          					option "Thin line" value: "lab-p-thin"
          					option "Segments" value: "lab-p-segments"
          					option "Stripes" value: "lab-p-stripes"
          					option "Square" value: "lab-p-square"
          					option "Neon" value: "lab-p-neon"
          					option "Dashed" value: "lab-p-dashed"
          }
          select "Animation" value: pAnim {
          					option "Pulse" value: "lab-p-anim-pulse"
          					option "Stripes" value: "lab-p-anim-stripes"
          					option "Shimmer" value: "lab-p-anim-shimmer"
          					option "Glow" value: "lab-p-anim-glow"
          					option "Indeterminate" value: "lab-p-anim-indeterminate"
          					option "Aurora" value: "lab-p-anim-aurora"
          					option "Breathe" value: "lab-p-anim-breathe"
          					option "Ticker" value: "lab-p-anim-ticker"
          					option "Spark" value: "lab-p-anim-spark"
          					option "Wave" value: "lab-p-anim-wave"
          					option "None" value: "lab-p-anim-none"
          }
        }
        p "7 shapes x 11 animations, chosen live." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Spinner"
        div class: "lab-preview" {
          spinner class: sShape + " " + sAnim
        }
        div class: "lab-fields" {
          select "Shape" value: sShape {
          					option "Ring" value: "lab-s-ring"
          					option "Tri dots" value: "lab-s-dots"
          					option "Bars" value: "lab-s-bars"
          					option "Orbit" value: "lab-s-orbit"
          					option "Comet" value: "lab-s-comet"
          					option "Pulse" value: "lab-s-pulse"
          					option "Square" value: "lab-s-square"
          					option "Hourglass" value: "lab-s-hourglass"
          					option "Drop fill" value: "lab-s-drop"
          }
          select "Animation" value: sAnim {
          					option "Spin" value: "lab-s-anim-spin"
          					option "Pulse" value: "lab-s-anim-pulse"
          					option "Flip" value: "lab-s-anim-flip"
          					option "Reverse" value: "lab-s-anim-reverse"
          					option "Orbit" value: "lab-s-anim-orbit"
          					option "Wobble" value: "lab-s-anim-wobble"
          					option "Bounce" value: "lab-s-anim-bounce"
          					option "Ripple" value: "lab-s-anim-ripple"
          					option "Swing" value: "lab-s-anim-swing"
          					option "None" value: "lab-s-anim-none"
          }
        }
        p "Nine geometries, ten motions." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Badge"
        div class: "lab-preview" {
          row {
            badge "LIVE" class: bStyle + " " + bAnim
          }
        }
        div class: "lab-fields" {
          select "Style" value: bStyle {
          					option "Pill" value: "lab-b-pill"
          					option "Tag" value: "lab-b-tag"
          					option "Outline" value: "lab-b-outline"
          					option "Neon" value: "lab-b-neon"
          					option "Status" value: "lab-b-status"
          					option "Brutalist" value: "lab-b-brutalist"
          					option "Glass" value: "lab-b-glass"
          }
          select "Animation" value: bAnim {
          					option "Pulse" value: "lab-b-anim-pulse"
          					option "Glow" value: "lab-b-anim-glow"
          					option "Float" value: "lab-b-anim-float"
          					option "Pop" value: "lab-b-anim-pop"
          					option "Blink" value: "lab-b-anim-blink"
          					option "Tilt" value: "lab-b-anim-tilt"
          					option "Ping" value: "lab-b-anim-ping"
          					option "Bounce" value: "lab-b-anim-bounce"
          					option "Jelly" value: "lab-b-anim-jelly"
          					option "Sheen" value: "lab-b-anim-sheen"
          					option "Dot to pill" value: "lab-b-anim-dot"
          					option "Border first" value: "lab-b-anim-border"
          					option "None" value: "lab-b-anim-none"
          }
        }
        p "Tag, status dot, neon, brutalist and more." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Alert"
        div class: "lab-preview" {
          alert "Deploy completed" class: aStyle + " " + aAnim
        }
        div class: "lab-fields" {
          select "Style" value: aStyle {
          					option "Accent bar" value: "lab-a-bar"
          					option "Filled" value: "lab-a-filled"
          					option "Outline" value: "lab-a-outline"
          					option "Glass" value: "lab-a-glass"
          					option "Brutalist" value: "lab-a-brutalist"
          					option "Info" value: "lab-a-info"
          }
          select "Animation" value: aAnim {
          					option "Fade in" value: "lab-a-anim-fade"
          					option "Slide in" value: "lab-a-anim-slide"
          					option "Shake" value: "lab-a-anim-shake"
          					option "Pulse" value: "lab-a-anim-pulse"
          					option "Zoom in" value: "lab-a-anim-zoom"
          					option "Bounce in" value: "lab-a-anim-bounce-in"
          					option "Slide right" value: "lab-a-anim-slide-right"
          					option "Wobble" value: "lab-a-anim-wobble"
          					option "Breathe" value: "lab-a-anim-breathe"
          					option "Flash" value: "lab-a-anim-flash"
          					option "Wipe" value: "lab-a-anim-wipe"
          					option "Border first" value: "lab-a-anim-border"
          					option "None" value: "lab-a-anim-none"
          }
        }
        p "Entrances, loops and attention effects." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Card"
        div class: "lab-preview" {
          card class: cStyle + " " + cAnim {
            strong "Aurora card"
            p "Surface and motion are independent." class: "lab-note"
          }
        }
        div class: "lab-fields" {
          select "Style" value: cStyle {
          					option "Elevated" value: "lab-c-elevated"
          					option "Flat" value: "lab-c-flat"
          					option "Outline" value: "lab-c-outline"
          					option "Glass" value: "lab-c-glass"
          					option "Neon" value: "lab-c-neon"
          					option "Brutalist" value: "lab-c-brutalist"
          					option "Dashed" value: "lab-c-dashed"
          }
          select "Animation" value: cAnim {
          					option "Fade in" value: "lab-c-anim-fade"
          					option "Float" value: "lab-c-anim-float"
          					option "Rise" value: "lab-c-anim-rise"
          					option "Zoom in" value: "lab-c-anim-zoom"
          					option "Tilt in" value: "lab-c-anim-tilt"
          					option "Halo" value: "lab-c-anim-halo"
          					option "Drift" value: "lab-c-anim-drift"
          					option "Unfold" value: "lab-c-anim-unfold"
          					option "Dot to card" value: "lab-c-anim-dot"
          					option "Border first" value: "lab-c-anim-border"
          					option "Wipe" value: "lab-c-anim-wipe"
          					option "Flip" value: "lab-c-anim-flip"
          					option "Squash" value: "lab-c-anim-squash"
          					option "None" value: "lab-c-anim-none"
          }
        }
        p "Seven surfaces, fourteen motions." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Button"
        div class: "lab-preview" {
          col class: "lab-stack" {
            button "Press me" class: btnStyle + " " + btnAnim onClick: press
            text "Clicks: {presses}" class: "lab-note"
          }
        }
        div class: "lab-fields" {
          select "Style" value: btnStyle {
          					option "Solid" value: "lab-btn-solid"
          					option "Gradient" value: "lab-btn-gradient"
          					option "Outline" value: "lab-btn-outline"
          					option "Ghost" value: "lab-btn-ghost"
          					option "Neon" value: "lab-btn-neon"
          					option "3D" value: "lab-btn-3d"
          					option "Pill" value: "lab-btn-pill"
          					option "Brutalist" value: "lab-btn-brutalist"
          }
          select "Animation" value: btnAnim {
          					option "Pulse" value: "lab-btn-anim-pulse"
          					option "Float" value: "lab-btn-anim-float"
          					option "Glow" value: "lab-btn-anim-glow"
          					option "Jelly" value: "lab-btn-anim-jelly"
          					option "Bounce" value: "lab-btn-anim-bounce"
          					option "Tilt" value: "lab-btn-anim-tilt"
          					option "Ring" value: "lab-btn-anim-ring"
          					option "Sheen" value: "lab-btn-anim-sheen"
          					option "Press" value: "lab-btn-anim-press"
          					option "Border first" value: "lab-btn-anim-border"
          					option "Wipe" value: "lab-btn-anim-wipe"
          					option "Pop" value: "lab-btn-anim-pop"
          					option "None" value: "lab-btn-anim-none"
          }
        }
        p "Solid, gradient, 3D, ghost, brutalist..." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Tabs"
        div class: "lab-preview" {
          tabs "Sections" class: tStyle + " " + tAnim {
            tab "One" {
              p "First tab" class: "lab-note"
            }
            tab "Two" {
              p "Second tab" class: "lab-note"
            }
          }
        }
        div class: "lab-fields" {
          select "Style" value: tStyle {
          					option "Underline" value: "lab-t-underline"
          					option "Pills" value: "lab-t-pills"
          					option "Segmented" value: "lab-t-segmented"
          					option "Boxed" value: "lab-t-boxed"
          					option "Brutalist" value: "lab-t-brutalist"
          					option "Neon" value: "lab-t-neon"
          }
          select "Animation" value: tAnim {
          					option "Pop" value: "lab-t-anim-pop"
          					option "Slide" value: "lab-t-anim-slide"
          					option "Glow" value: "lab-t-anim-glow"
          					option "Sweep" value: "lab-t-anim-sweep"
          					option "None" value: "lab-t-anim-none"
          }
        }
        p "The effect runs on the active tab." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Input"
        div class: "lab-preview" {
          input "Focus me" value: name class: iStyle + " " + iAnim
        }
        div class: "lab-fields" {
          select "Style" value: iStyle {
          					option "Outline" value: "lab-i-outline"
          					option "Filled" value: "lab-i-filled"
          					option "Underline" value: "lab-i-underline"
          					option "Pill" value: "lab-i-pill"
          					option "Glass" value: "lab-i-glass"
          					option "Brutalist" value: "lab-i-brutalist"
          }
          select "Animation" value: iAnim {
          					option "Focus ring" value: "lab-i-anim-ring"
          					option "Focus glow" value: "lab-i-anim-glow"
          					option "Focus lift" value: "lab-i-anim-lift"
          					option "Focus shake" value: "lab-i-anim-shake"
          					option "None" value: "lab-i-anim-none"
          }
        }
        p "Click the field to see the focus effect." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Checkbox"
        div class: "lab-preview" {
          checkbox "Accept terms" checked: chkOn class: chkStyle + " " + chkTheme animation: chkAnim duration: "180ms" easing: "ease-out"
        }
        div class: "lab-fields" {
          select "Style" value: chkStyle class: "lab-wide" {
          					option "Classic" value: "lab-check-classic"
          					option "Round" value: "lab-check-round"
          					option "Tile" value: "lab-check-tile"
          					option "LED" value: "lab-check-led"
          					option "Brutalist" value: "lab-check-brutalist"
          					option "Neon" value: "lab-check-neon"
          					option "Sweep fill" value: "lab-check-sweep"
          }
          select "Theme" value: chkTheme {
          					option "Accent" value: "lab-cb-theme-accent"
          					option "Success" value: "lab-cb-theme-success"
          					option "Warning" value: "lab-cb-theme-warning"
          					option "Danger" value: "lab-cb-theme-danger"
          					option "Mono" value: "lab-cb-theme-mono"
          }
          select "Animation" value: chkAnim {
          					option "Auto" value: "auto"
          					option "Scale" value: "scale"
          					option "Morph" value: "morph"
          					option "Fade" value: "fade"
          					option "Slide" value: "slide"
          					option "None" value: "none"
          }
        }
        p "Seven structures: round, tile, LED, brutalist, neon, sweep fill." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Switch"
        div class: "lab-preview" {
          switch "Enabled" checked: swOn shape: swShape theme: swTheme animation: swAnim duration: "180ms" easing: "ease-out"
        }
        div class: "lab-fields" {
          select "Shape" value: swShape {
          					option "Default" value: "default"
          					option "Rounded" value: "rounded"
          					option "Square" value: "square"
          					option "Pill" value: "pill"
          					option "iOS" value: "ios"
          					option "Material" value: "material"
          					option "Minimal" value: "minimal"
          					option "Soft" value: "soft"
          }
          select "Theme" value: swTheme {
          					option "Default" value: "default"
          					option "Accent" value: "accent"
          					option "Success" value: "success"
          					option "Warning" value: "warning"
          					option "Danger" value: "danger"
          					option "Mono" value: "mono"
          					option "Dark" value: "dark"
          					option "Light" value: "light"
          }
          select "Animation" value: swAnim class: "lab-wide" {
          					option "Auto" value: "auto"
          					option "Scale" value: "scale"
          					option "Morph" value: "morph"
          					option "Fade" value: "fade"
          					option "Slide" value: "slide"
          					option "None" value: "none"
          }
        }
        p "Eight shapes and eight themes at runtime." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Radio"
        div class: "lab-preview" {
          col {
            radio "Option A" name: "lab-radio" checked: rdSel == "a" shape: rdShape theme: rdTheme animation: rdAnim onChange: pickRadioA
            radio "Option B" name: "lab-radio" checked: rdSel == "b" shape: rdShape theme: rdTheme animation: rdAnim onChange: pickRadioB
          }
        }
        div class: "lab-fields" {
          select "Shape" value: rdShape {
          					option "Default" value: "default"
          					option "Compact" value: "compact"
          					option "Large" value: "large"
          					option "Soft" value: "soft"
          					option "Minimal" value: "minimal"
          }
          select "Theme" value: rdTheme {
          					option "Default" value: "default"
          					option "Accent" value: "accent"
          					option "Success" value: "success"
          					option "Warning" value: "warning"
          					option "Danger" value: "danger"
          					option "Mono" value: "mono"
          					option "Dark" value: "dark"
          					option "Light" value: "light"
          }
          select "Animation" value: rdAnim class: "lab-wide" {
          					option "Auto" value: "auto"
          					option "Scale" value: "scale"
          					option "Morph" value: "morph"
          					option "Fade" value: "fade"
          					option "Slide" value: "slide"
          					option "None" value: "none"
          }
        }
        p "Two options in one group, so the choice is visible." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Dropdown"
        div class: "lab-preview" {
          dropdown "Open menu" class: ddStyle + " " + ddItem animation: ddAnim duration: "200ms" {
            dropdownitem "Edit"
            dropdownitem "Duplicate"
            dropdownitem "Delete"
          }
        }
        div class: "lab-fields" {
          select "Style" value: ddStyle {
          					option "Classic" value: "lab-dd-classic"
          					option "Pill" value: "lab-dd-pill"
          					option "Neon" value: "lab-dd-neon"
          					option "Brutalist" value: "lab-dd-brutalist"
          					option "Glass" value: "lab-dd-glass"
          					option "Outline" value: "lab-dd-outline"
          }
          select "Animation" value: ddAnim {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
          select "Item highlight" value: ddItem class: "lab-wide" {
          					option "Accent bar" value: "lab-items-bar"
          					option "Glow" value: "lab-items-glow"
          					option "Aura" value: "lab-items-aura"
          					option "Gradient" value: "lab-items-gradient"
          					option "Border" value: "lab-items-border"
          					option "Dot" value: "lab-items-dot"
          					option "Check mark" value: "lab-items-check"
          					option "Bar + aura" value: "lab-items-bar lab-items-aura"
          					option "Bar + colour cloud" value: "lab-items-bar lab-items-cloud"
          					option "Bar + underline" value: "lab-items-bar lab-items-underline"
          					option "Bar + glow" value: "lab-items-bar lab-items-glow"
          					option "None" value: "lab-items-none"
          }
        }
        p "Hover an item to see the selection highlight." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Popover"
        div class: "lab-preview" {
          popover "Open popover" class: popStyle animation: popAnim duration: "180ms" {
            p "Anchored panel." class: "lab-note"
          }
        }
        div class: "lab-fields" {
          select "Style" value: popStyle {
          					option "Card" value: "lab-pop-card"
          					option "Glass" value: "lab-pop-glass"
          					option "Neon" value: "lab-pop-neon"
          					option "Brutalist" value: "lab-pop-brutalist"
          }
          select "Animation" value: popAnim {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
        }
      }

      card class: "lab-card" {
        h3 "Combobox"
        div class: "lab-preview" {
          combobox "Country" class: cbStyle options: "Italy,France,Germany" animation: cbAnim duration: "200ms"
        }
        div class: "lab-fields" {
          select "Style" value: cbStyle {
          					option "Classic" value: "lab-cb-classic"
          					option "Pill" value: "lab-cb-pill"
          					option "Neon" value: "lab-cb-neon"
          					option "Brutalist" value: "lab-cb-brutalist"
          }
          select "Animation" value: cbAnim {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
        }
      }

      card class: "lab-card" {
        h3 "Command palette"
        div class: "lab-preview" {
          button "Open commands" onClick: openCommands
          commandbar "Commands" open: showCommands hotkey: "k" animation: cmdAnim duration: "200ms" {
            command "New" onClick: noop
            command "Settings" onClick: noop
          }
        }
        div class: "lab-fields" {
          select "Animation" value: cmdAnim class: "lab-wide" {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
        }
        p "Press Ctrl+K or use the button." class: "lab-note"
      }
    }
  }
}
```

CSS-backed controls combine two classes (surface + motion). Overlay and toggle
controls use the runtime parameters instead:

```sui
dropdown "Menu" class: ddStyle animation: ddAnim
switch "Enabled" shape: swShape theme: swTheme animation: swAnim
```

## Motion editor

The first card tunes the animation library live:

- **Speed** multiplies every animation duration (`0.3` to `2.5`);
- **Easing** overrides the timing function (Default keeps each effect's own);
- **Delay** shifts the start of every animation;
- **Copy CSS** copies a snippet with the generated custom properties.

The animation declarations use
`calc(<duration> * var(--lab-motion-speed, 1))` plus
`var(--lab-motion-easing, ...)` and `var(--lab-motion-delay, 0s)`, so the values
apply instantly to every control on the page.

New keyframes are a **build-time** feature: add `@keyframes` in `lab.css` (or
in the project CSS) and rebuild; the motion is then selectable like the others.
Arbitrary CSS injection at runtime is intentionally blocked by the class
allow-list.

To scope the tuning to one control, paste the snippet with a selector instead of
`:root`, for example:

```css
.my-card {
  --lab-motion-speed: .7;
  --lab-motion-easing: ease-out;
}
```

To draw a control and its keyframes rather than retune the existing library, use
the designer: **Tools → Shape & motion…**. It writes managed blocks
(`/* sui-motion:... */`) into `root.css` and can either insert a new control or
animate the current selection. The generated keyframes are ordinary project CSS,
so this lab can keep using its build-time library.

## Re-render safety

Because these selectors update state, the view re-renders when you choose an
option. The runtime keeps the page stable: focus and caret are preserved, media
and iframes keep their loaded state, running CSS animations keep their phase,
select values are restored, and the scroll position does not jump. A control
whose value binding is not read anywhere else (for example the input field in
this page) does not re-render at all while you type.

## Reduced motion

Every looping or entrance effect ships a `prefers-reduced-motion` guard that
stops the animation while keeping the selected colours and layout.

## Full SUI source

```sui
state name = ""
state pStyle = "lab-p-classic"
state pAnim = "lab-p-anim-wave"
state sShape = "lab-s-ring"
state sAnim = "lab-s-anim-spin"
state bStyle = "lab-b-pill"
state bAnim = "lab-b-anim-ping"
state aStyle = "lab-a-bar"
state aAnim = "lab-a-anim-breathe"
state cStyle = "lab-c-elevated"
state cAnim = "lab-c-anim-halo"
state btnStyle = "lab-btn-gradient"
state btnAnim = "lab-btn-anim-jelly"
state tStyle = "lab-t-underline"
state tAnim = "lab-t-anim-pop"
state iStyle = "lab-i-outline"
state iAnim = "lab-i-anim-glow"
state chkOn = true
state chkShape = "default"
state chkTheme = "accent"
state chkAnim = "scale"
state swOn = true
state swShape = "pill"
state swTheme = "accent"
state swAnim = "morph"
state rdOn = true
state rdShape = "default"
state rdTheme = "accent"
state rdAnim = "slide"
state ddStyle = "lab-dd-classic"
state ddAnim = "slide"
state popStyle = "lab-pop-card"
state popAnim = "scale"
state cbStyle = "lab-cb-classic"
state cbAnim = "slide"
state cmdAnim = "slide"
state showCommands = false

fn noop() -> null {
	return null
}

fn openCommands() -> null {
	showCommands = true
}

css "/* Animation lab: visual styles and motion for the animatable controls. */
.lab-page \{ max-width: 1200px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \}
.lab-head \{ margin: 0; \}
.lab-note \{ color: #8fa0b8; font-size: 13px; margin: 0; \}
.lab-grid \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(340px, 1fr)); gap: 16px; align-items: start; \}
.lab-card \{ background: #121826; border: 1px solid #263141; border-radius: 16px; padding: 16px; display: flex; flex-direction: column; gap: 12px; \}
.lab-card h3 \{ margin: 0; font-size: 15px; \}
.lab-preview \{ min-height: 104px; display: flex; align-items: center; justify-content: center; padding: 14px; border-radius: 12px; background: radial-gradient(circle at 50% 0%, #1a2436, #0d131f); border: 1px solid #1f2937; overflow: hidden; \}
.lab-fields \{ display: grid; grid-template-columns: 1fr 1fr; gap: 8px; \}
.lab-fields .lab-wide \{ grid-column: 1 / -1; \}
.lab-fields select \{ width: 100%; \}
.lab-kbd \{ font-family: ui-monospace, Consolas, monospace; font-size: 11px; color: #c7d2fe; background: #1e2637; border: 1px solid #334155; border-radius: 4px; padding: 1px 5px; \}

/* ---------- Progress: styles ---------- */
.sui-progress.lab-p-classic \{ height: 10px; border-radius: 999px; background: #263141; \}
.sui-progress.lab-p-classic::-webkit-progress-bar \{ background: #263141; \}
.sui-progress.lab-p-classic::-webkit-progress-value \{ background: linear-gradient(90deg, #4c5cf7, #8b5cf6); border-radius: 999px; \}
.sui-progress.lab-p-thin \{ height: 4px; border-radius: 999px; background: #1f2937; \}
.sui-progress.lab-p-thin::-webkit-progress-bar \{ background: #1f2937; \}
.sui-progress.lab-p-thin::-webkit-progress-value \{ background: #22d3ee; border-radius: 999px; \}
.sui-progress.lab-p-segments \{ height: 14px; border-radius: 4px; background: #0b1220; \}
.sui-progress.lab-p-segments::-webkit-progress-bar \{ background: #0b1220; \}
.sui-progress.lab-p-segments::-webkit-progress-value \{ background: repeating-linear-gradient(90deg, #22d3ee 0 8px, transparent 8px 14px); border-radius: 3px; \}
.sui-progress.lab-p-stripes \{ height: 16px; border-radius: 4px; background: #1f2937; \}
.sui-progress.lab-p-stripes::-webkit-progress-bar \{ background: #1f2937; \}
.sui-progress.lab-p-stripes::-webkit-progress-value \{ background: repeating-linear-gradient(45deg, #f59e0b 0 10px, #ef4444 10px 20px); \}
.sui-progress.lab-p-square \{ height: 16px; border-radius: 0; background: #1f2937; border: 2px solid #0b1220; box-shadow: 5px 5px 0 rgba(0,0,0,.35); \}
.sui-progress.lab-p-square::-webkit-progress-bar \{ background: #1f2937; \}
.sui-progress.lab-p-square::-webkit-progress-value \{ background: #4c5cf7; border-radius: 0; \}
.sui-progress.lab-p-neon \{ height: 12px; border-radius: 999px; background: #0b1220; box-shadow: inset 0 0 6px #22d3ee; \}
.sui-progress.lab-p-neon::-webkit-progress-bar \{ background: #0b1220; \}
.sui-progress.lab-p-neon::-webkit-progress-value \{ background: linear-gradient(90deg, #22d3ee, #a855f7); box-shadow: 0 0 12px #22d3ee; \}
.sui-progress.lab-p-dashed \{ height: 14px; border-radius: 0; background: transparent; border: 2px dashed #64748b; padding: 2px; \}
.sui-progress.lab-p-dashed::-webkit-progress-bar \{ background: transparent; \}
.sui-progress.lab-p-dashed::-webkit-progress-value \{ background: linear-gradient(90deg, #94a3b8, #cbd5e1); \}

/* ---------- Progress: animations ---------- */
.sui-progress.lab-p-anim-pulse::-webkit-progress-value \{ animation: lab-progress-pulse 1.4s ease-in-out infinite; \}
.sui-progress.lab-p-anim-stripes::-webkit-progress-value \{ background: repeating-linear-gradient(45deg, #4c5cf7 0 10px, #8b5cf6 10px 20px); background-size: 28px 100%; animation: lab-progress-stripes .9s linear infinite; \}
.sui-progress.lab-p-anim-shimmer::-webkit-progress-value \{ background: linear-gradient(90deg, transparent 20%, rgba(255,255,255,.6) 50%, transparent 80%), linear-gradient(90deg, #4c5cf7, #6366f1); background-size: 220% 100%; animation: lab-progress-shimmer 2.2s linear infinite; \}
.sui-progress.lab-p-anim-glow::-webkit-progress-value \{ animation: lab-progress-glow 1.6s ease-in-out infinite; \}
.sui-progress.lab-p-anim-indeterminate::-webkit-progress-value \{ background: linear-gradient(90deg, transparent 12%, #4c5cf7 42%, #6366f1 58%, transparent 88%); background-size: 200% 100%; animation: lab-progress-scan 1.25s ease-in-out infinite; \}
.sui-progress.lab-p-anim-aurora::-webkit-progress-value \{ animation: lab-progress-aurora 7s linear infinite; \}
.sui-progress.lab-p-anim-breathe::-webkit-progress-value \{ animation: lab-progress-breathe 2.4s ease-in-out infinite; \}
.sui-progress.lab-p-anim-ticker::-webkit-progress-value \{ background: repeating-linear-gradient(90deg, #22d3ee 0 6px, transparent 6px 12px); animation: lab-progress-ticker 1s steps(10) infinite; \}
.sui-progress.lab-p-anim-spark::-webkit-progress-value \{ animation: lab-progress-spark 1.6s steps(6) infinite; \}
.sui-progress.lab-p-anim-wave::-webkit-progress-value \{ background: linear-gradient(180deg, #4c5cf7 0 34%, #6366f1 34% 62%, #4c5cf7 62% 100%); background-size: 100% 240%; animation: lab-progress-wave 1.8s ease-in-out infinite; \}
.sui-progress.lab-p-anim-none \{ animation: none; \}
@keyframes lab-progress-pulse \{ 0%, 100% \{ opacity: 1; \} 50% \{ opacity: .62; \} \}
@keyframes lab-progress-stripes \{ to \{ background-position: 28px 0; \} \}
@keyframes lab-progress-shimmer \{ from \{ background-position: 220% 0; \} to \{ background-position: -220% 0; \} \}
@keyframes lab-progress-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.5); \} \}
@keyframes lab-progress-scan \{ 0% \{ background-position: 130% 0; \} 100% \{ background-position: -130% 0; \} \}
@keyframes lab-progress-aurora \{ to \{ filter: hue-rotate(360deg); \} \}
@keyframes lab-progress-breathe \{ 0%, 100% \{ transform: scaleY(.8); filter: brightness(.92); \} 50% \{ transform: scaleY(1); filter: brightness(1.22); \} \}
@keyframes lab-progress-ticker \{ to \{ background-position: 24px 0; \} \}
@keyframes lab-progress-spark \{ 0%, 100% \{ opacity: 1; \} 40% \{ opacity: .5; \} 55% \{ opacity: .85; \} \}
@keyframes lab-progress-wave \{ 0%, 100% \{ background-position: 0 0; \} 50% \{ background-position: 0 100%; \} \}

/* ---------- Spinner: shapes ---------- */
.sui-spinner.lab-s-ring \{ border-color: rgba(76,92,247,.22); border-top-color: #4c5cf7; border-bottom-color: transparent; \}
.sui-spinner.lab-s-dots \{ width: 30px; height: 30px; border: 0; border-radius: 50%; background: radial-gradient(circle at 50% 8%, #4c5cf7 3.5px, transparent 4px), radial-gradient(circle at 88% 78%, #6366f1 3.5px, transparent 4px), radial-gradient(circle at 12% 78%, #22d3ee 3.5px, transparent 4px); \}
.sui-spinner.lab-s-bars \{ width: 28px; height: 24px; border: 0; border-radius: 0; background: linear-gradient(#4c5cf7, #4c5cf7) 0 100%/5px 50% no-repeat, linear-gradient(#6366f1, #6366f1) 50% 100%/5px 80% no-repeat, linear-gradient(#22d3ee, #22d3ee) 100% 100%/5px 40% no-repeat; \}
.sui-spinner.lab-s-orbit \{ border: 2px solid rgba(76,92,247,.25); border-top-color: transparent; border-bottom-color: transparent; background: transparent; position: relative; \}
.sui-spinner.lab-s-orbit::before \{ content: \"\"; position: absolute; top: -3px; left: 50%; width: 8px; height: 8px; margin-left: -4px; border-radius: 50%; background: #4c5cf7; box-shadow: 0 0 8px #4c5cf7; \}
.sui-spinner.lab-s-comet \{ border: 0; background: conic-gradient(from 0deg, transparent 0 55%, #22d3ee 78%, #4c5cf7 100%); -webkit-mask: radial-gradient(farthest-side, transparent calc(100% - 5px), #000 calc(100% - 4px)); mask: radial-gradient(farthest-side, transparent calc(100% - 5px), #000 calc(100% - 4px)); \}
.sui-spinner.lab-s-pulse \{ border: 0; background: #4c5cf7; \}
.sui-spinner.lab-s-square \{ border-radius: 6px; border-color: rgba(76,92,247,.22); border-top-color: #4c5cf7; \}
/* ---------- Spinner: animations ---------- */
.sui-spinner.lab-s-anim-spin \{ animation: lab-spin 1s linear infinite; \}
.sui-spinner.lab-s-anim-pulse \{ animation: lab-spin-pulse 1.4s ease-in-out infinite; \}
.sui-spinner.lab-s-anim-flip \{ animation: lab-spin-flip 1.4s ease-in-out infinite; \}
.sui-spinner.lab-s-anim-reverse \{ animation: lab-spin-reverse 1.1s linear infinite; \}
.sui-spinner.lab-s-anim-orbit \{ animation: lab-spin-orbit 1.2s linear infinite; \}
.sui-spinner.lab-s-anim-wobble \{ animation: lab-spin-wobble 1.4s ease-in-out infinite; \}
.sui-spinner.lab-s-anim-bounce \{ animation: lab-spin-bounce .9s ease-in-out infinite; \}
.sui-spinner.lab-s-anim-ripple \{ animation: lab-spin-ripple 1.4s ease-out infinite; \}
.sui-spinner.lab-s-anim-swing \{ animation: lab-spin-swing 1.6s ease-in-out infinite; \}
.sui-spinner.lab-s-anim-none \{ animation: none; \}
@keyframes lab-spin \{ to \{ transform: rotate(360deg); \} \}
@keyframes lab-spin-reverse \{ to \{ transform: rotate(-360deg); \} \}
@keyframes lab-spin-pulse \{ 0%, 100% \{ transform: scale(.68); opacity: .55; \} 50% \{ transform: scale(1); opacity: 1; \} \}
@keyframes lab-spin-flip \{ to \{ transform: rotateY(360deg); \} \}
@keyframes lab-spin-orbit \{ 0% \{ transform: rotate(0deg) translateX(0); \} 50% \{ transform: rotate(180deg) translateX(3px); \} 100% \{ transform: rotate(360deg) translateX(0); \} \}
@keyframes lab-spin-wobble \{ 0%, 100% \{ transform: rotate(-14deg) scale(.94); \} 50% \{ transform: rotate(14deg) scale(1.06); \} \}
@keyframes lab-spin-bounce \{ 0%, 100% \{ transform: translateY(-16%) scale(.94); \} 55% \{ transform: translateY(16%) scale(1.04); \} \}
@keyframes lab-spin-ripple \{ 0% \{ transform: scale(.66); opacity: .45; \} 70% \{ transform: scale(1.04); opacity: 1; \} 100% \{ transform: scale(.66); opacity: .45; \} \}
@keyframes lab-spin-swing \{ 0%, 100% \{ transform: rotate(-38deg); \} 50% \{ transform: rotate(38deg); \} \}

/* ---------- Badge: styles ---------- */
.sui-badge.lab-b-pill \{ background: #252b45; color: #a5b4fc; border: 1px solid #3b4470; border-radius: 999px; padding: 3px 10px; \}
.sui-badge.lab-b-tag \{ background: #f59e0b; color: #1a1206; border-radius: 4px; text-transform: uppercase; letter-spacing: .08em; font-weight: 700; padding: 3px 7px; \}
.sui-badge.lab-b-outline \{ background: transparent; color: #93c5fd; border: 1px solid #3b82f6; border-radius: 999px; padding: 3px 10px; \}
.sui-badge.lab-b-neon \{ background: #07131a; color: #22d3ee; border: 1px solid #22d3ee; border-radius: 6px; box-shadow: 0 0 12px rgba(34,211,238,.5); padding: 3px 10px; \}
.sui-badge.lab-b-status \{ background: #052e1b; color: #86efac; border-radius: 999px; padding: 4px 10px 4px 22px; position: relative; \}
.sui-badge.lab-b-status::before \{ content: \"\"; position: absolute; left: 8px; top: 50%; width: 7px; height: 7px; margin-top: -3.5px; border-radius: 50%; background: #22c55e; box-shadow: 0 0 6px #22c55e; \}
.sui-badge.lab-b-brutalist \{ background: #facc15; color: #111; border: 2px solid #111; border-radius: 0; box-shadow: 3px 3px 0 #111; padding: 3px 9px; font-weight: 700; \}
.sui-badge.lab-b-glass \{ background: rgba(148,163,184,.18); color: #e2e8f0; border: 1px solid rgba(255,255,255,.25); border-radius: 999px; backdrop-filter: blur(8px); padding: 3px 10px; \}
/* ---------- Badge: animations ---------- */
.lab-b-anim-pulse \{ animation: lab-badge-pulse 1.8s ease-in-out infinite; \}
.lab-b-anim-glow \{ animation: lab-badge-glow 2s ease-in-out infinite; \}
.lab-b-anim-float \{ animation: lab-badge-float 2.4s ease-in-out infinite; \}
.lab-b-anim-pop \{ animation: lab-badge-pop .42s cubic-bezier(.2,1.4,.4,1); \}
.lab-b-anim-blink \{ animation: lab-badge-blink 1.2s steps(1,end) infinite; \}
.lab-b-anim-tilt \{ animation: lab-badge-tilt 2.6s ease-in-out infinite; \}
.lab-b-anim-ping \{ animation: lab-badge-ping 1.8s ease-out infinite; \}
.lab-b-anim-bounce \{ animation: lab-badge-bounce 1.6s ease-in-out infinite; \}
.lab-b-anim-jelly \{ animation: lab-badge-jelly 2.2s ease-in-out infinite; \}
.lab-b-anim-sheen \{ animation: lab-badge-sheen 2.4s ease-in-out infinite; \}
.lab-b-anim-none \{ animation: none; \}
@keyframes lab-badge-pulse \{ 0%, 100% \{ transform: scale(1); \} 50% \{ transform: scale(1.06); \} \}
@keyframes lab-badge-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.35); \} \}
@keyframes lab-badge-float \{ 0%, 100% \{ transform: translateY(0); \} 50% \{ transform: translateY(-3px); \} \}
@keyframes lab-badge-pop \{ 0% \{ opacity: 0; scale: .6; \} 65% \{ opacity: 1; scale: 1.1; \} 100% \{ opacity: 1; scale: 1; \} \}
@keyframes lab-badge-blink \{ 0%, 49% \{ opacity: 1; \} 50%, 100% \{ opacity: .35; \} \}
@keyframes lab-badge-tilt \{ 0%, 100% \{ rotate: -3deg; \} 50% \{ rotate: 3deg; \} \}
@keyframes lab-badge-ping \{ 0% \{ box-shadow: 0 0 0 0 rgba(74,222,128,.6); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(74,222,128,0); \} \}
@keyframes lab-badge-bounce \{ 0%, 100% \{ translate: 0 0; \} 40% \{ translate: 0 -4px; \} 60% \{ translate: 0 1px; \} \}
@keyframes lab-badge-jelly \{ 0%, 100% \{ scale: 1 1; \} 30% \{ scale: 1.12 .88; \} 55% \{ scale: .92 1.08; \} 80% \{ scale: 1.03 .97; \} \}
@keyframes lab-badge-sheen \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.45) saturate(1.15); \} \}

/* ---------- Alert: styles ---------- */
.sui-alert.lab-a-bar \{ background: #1b2230; color: #dbeafe; border: 0; border-left: 5px solid #3b82f6; border-radius: 4px; \}
.sui-alert.lab-a-filled \{ background: #7f1d1d; color: #fee2e2; border: 0; border-radius: 10px; \}
.sui-alert.lab-a-outline \{ background: transparent; color: #93c5fd; border: 1px dashed #3b82f6; border-radius: 10px; \}
.sui-alert.lab-a-glass \{ background: rgba(148,163,184,.16); color: #e2e8f0; border: 1px solid rgba(255,255,255,.22); border-radius: 12px; backdrop-filter: blur(10px); \}
.sui-alert.lab-a-brutalist \{ background: #facc15; color: #111; border: 2px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; \}
.sui-alert.lab-a-info \{ background: #0c2340; color: #bae6fd; border: 1px solid #0ea5e9; border-radius: 10px; \}
/* ---------- Alert: animations ---------- */
.lab-a-anim-fade \{ animation: lab-alert-in .24s ease-out; \}
.lab-a-anim-slide \{ animation: lab-alert-slide .26s ease-out; \}
.lab-a-anim-shake \{ animation: lab-alert-shake .5s ease-in-out; \}
.lab-a-anim-pulse \{ animation: lab-alert-pulse 1.8s ease-in-out infinite; \}
.lab-a-anim-zoom \{ animation: lab-alert-zoom .3s cubic-bezier(.2,1.2,.4,1); \}
.lab-a-anim-bounce-in \{ animation: lab-alert-bounce-in .46s cubic-bezier(.2,1.3,.4,1); \}
.lab-a-anim-slide-right \{ animation: lab-alert-slide-right .28s ease-out; \}
.lab-a-anim-wobble \{ animation: lab-alert-wobble 2.4s ease-in-out infinite; \}
.lab-a-anim-breathe \{ animation: lab-alert-breathe 2.6s ease-in-out infinite; \}
.lab-a-anim-flash \{ animation: lab-alert-flash 1.4s ease-in-out infinite; \}
.lab-a-anim-none \{ animation: none; \}
@keyframes lab-alert-in \{ from \{ opacity: 0; \} to \{ opacity: 1; \} \}
@keyframes lab-alert-slide \{ from \{ opacity: 0; transform: translateY(-6px); \} to \{ opacity: 1; transform: translateY(0); \} \}
@keyframes lab-alert-shake \{ 0%, 100% \{ transform: translateX(0); \} 20% \{ transform: translateX(-4px); \} 40% \{ transform: translateX(4px); \} 60% \{ transform: translateX(-2px); \} 80% \{ transform: translateX(2px); \} \}
@keyframes lab-alert-pulse \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(127,58,61,0); \} 50% \{ box-shadow: 0 0 0 5px rgba(127,58,61,.32); \} \}
@keyframes lab-alert-zoom \{ from \{ opacity: 0; scale: .92; \} to \{ opacity: 1; scale: 1; \} \}
@keyframes lab-alert-bounce-in \{ 0% \{ opacity: 0; translate: 0 -10px; scale: .97; \} 60% \{ opacity: 1; translate: 0 2px; scale: 1.005; \} 100% \{ opacity: 1; translate: 0 0; scale: 1; \} \}
@keyframes lab-alert-slide-right \{ from \{ opacity: 0; translate: 14px 0; \} to \{ opacity: 1; translate: 0 0; \} \}
@keyframes lab-alert-wobble \{ 0%, 100% \{ rotate: 0deg; \} 25% \{ rotate: -1.4deg; \} 75% \{ rotate: 1.4deg; \} \}
@keyframes lab-alert-breathe \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 7px #f59e0b); \} \}
@keyframes lab-alert-flash \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.25); \} \}

/* ---------- Card: styles ---------- */
.sui-card.lab-c-elevated \{ background: #161d2b; border: 1px solid #2a3446; border-radius: 14px; box-shadow: 0 14px 34px rgba(0,0,0,.4); \}
.sui-card.lab-c-flat \{ background: #161d2b; border: 0; border-radius: 4px; box-shadow: none; \}
.sui-card.lab-c-outline \{ background: transparent; border: 1px solid #3b82f6; border-radius: 14px; box-shadow: none; \}
.sui-card.lab-c-glass \{ background: rgba(148,163,184,.14); border: 1px solid rgba(255,255,255,.22); border-radius: 16px; backdrop-filter: blur(12px); \}
.sui-card.lab-c-neon \{ background: #0b1220; border: 1px solid #22d3ee; border-radius: 14px; box-shadow: 0 0 22px rgba(34,211,238,.35); \}
.sui-card.lab-c-brutalist \{ background: #fde68a; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 6px 6px 0 #111; \}
.sui-card.lab-c-dashed \{ background: #0f172a; border: 2px dashed #475569; border-radius: 10px; box-shadow: none; \}
/* ---------- Card: animations ---------- */
.lab-c-anim-fade \{ animation: lab-card-in .28s ease-out; \}
.lab-c-anim-float \{ animation: lab-card-float 3.2s ease-in-out infinite; \}
.lab-c-anim-rise \{ animation: lab-card-rise .38s cubic-bezier(.2,1,.3,1); \}
.lab-c-anim-zoom \{ animation: lab-card-zoom .34s cubic-bezier(.2,1.2,.4,1); \}
.lab-c-anim-tilt \{ animation: lab-card-tilt .45s cubic-bezier(.2,1,.3,1); \}
.lab-c-anim-halo \{ animation: lab-card-halo 3s ease-in-out infinite; \}
.lab-c-anim-drift \{ animation: lab-card-drift 5s ease-in-out infinite; \}
.lab-c-anim-unfold \{ animation: lab-card-unfold .4s cubic-bezier(.2,1,.3,1); \}
.lab-c-anim-none \{ animation: none; \}
@keyframes lab-card-in \{ from \{ opacity: 0; transform: translateY(6px); \} to \{ opacity: 1; transform: translateY(0); \} \}
@keyframes lab-card-float \{ 0%, 100% \{ transform: translateY(0); \} 50% \{ transform: translateY(-5px); \} \}
@keyframes lab-card-rise \{ from \{ opacity: 0; translate: 0 14px; \} to \{ opacity: 1; translate: 0 0; \} \}
@keyframes lab-card-zoom \{ from \{ opacity: 0; scale: .94; \} to \{ opacity: 1; scale: 1; \} \}
@keyframes lab-card-tilt \{ from \{ opacity: 0; transform: perspective(600px) rotateX(8deg) translateY(10px); \} to \{ opacity: 1; transform: perspective(600px) rotateX(0) translateY(0); \} \}
@keyframes lab-card-halo \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,0); \} 50% \{ box-shadow: 0 0 26px 2px rgba(76,92,247,.4); \} \}
@keyframes lab-card-drift \{ 0%, 100% \{ translate: 0 0; \} 50% \{ translate: 3px -4px; \} \}
@keyframes lab-card-unfold \{ from \{ opacity: 0; transform: scaleY(.72); transform-origin: top center; \} to \{ opacity: 1; transform: scaleY(1); \} \}

/* ---------- Button: styles ---------- */
.sui-button.lab-btn-solid \{ background: #2d3340; color: #e6e6e6; border: 1px solid #3d4556; border-radius: 8px; \}
.sui-button.lab-btn-gradient \{ background: linear-gradient(135deg, #4c5cf7, #a855f7); color: #fff; border: 0; border-radius: 10px; \}
.sui-button.lab-btn-outline \{ background: transparent; color: #93c5fd; border: 1px solid #3b82f6; border-radius: 8px; \}
.sui-button.lab-btn-ghost \{ background: transparent; color: #cbd5e1; border: 0; border-radius: 8px; \}
.sui-button.lab-btn-ghost:hover \{ background: #1f2937; \}
.sui-button.lab-btn-neon \{ background: #07131a; color: #22d3ee; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 14px rgba(34,211,238,.4); \}
.sui-button.lab-btn-3d \{ background: #4c5cf7; color: #fff; border: 0; border-radius: 10px; box-shadow: 0 5px 0 #1e2a78; transform: translateY(-1px); \}
.sui-button.lab-btn-3d:active \{ transform: translateY(4px); box-shadow: 0 1px 0 #1e2a78; \}
.sui-button.lab-btn-pill \{ background: #22c55e; color: #052e16; border: 0; border-radius: 999px; font-weight: 600; \}
.sui-button.lab-btn-brutalist \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 5px 5px 0 #111; font-weight: 700; \}
/* ---------- Button: animations ---------- */
.lab-btn-anim-pulse \{ animation: lab-button-pulse 1.8s ease-in-out infinite; \}
.lab-btn-anim-float \{ animation: lab-button-float 2.6s ease-in-out infinite; \}
.lab-btn-anim-glow \{ animation: lab-button-glow 2.2s ease-in-out infinite; \}
.lab-btn-anim-jelly \{ animation: lab-button-jelly 2.4s ease-in-out infinite; \}
.lab-btn-anim-bounce \{ animation: lab-button-bounce 1.8s ease-in-out infinite; \}
.lab-btn-anim-tilt \{ animation: lab-button-tilt 3s ease-in-out infinite; \}
.lab-btn-anim-ring \{ animation: lab-button-ring 2s ease-out infinite; \}
.lab-btn-anim-sheen \{ animation: lab-button-sheen 2.6s ease-in-out infinite; \}
.lab-btn-anim-press \{ animation: lab-button-press 2s ease-in-out infinite; \}
.lab-btn-anim-none \{ animation: none; \}
@keyframes lab-button-pulse \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.45); \} 50% \{ box-shadow: 0 0 0 6px rgba(76,92,247,0); \} \}
@keyframes lab-button-float \{ 0%, 100% \{ transform: translateY(0); \} 50% \{ transform: translateY(-3px); \} \}
@keyframes lab-button-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.25); \} \}
@keyframes lab-button-jelly \{ 0%, 100% \{ scale: 1 1; \} 30% \{ scale: 1.06 .94; \} 55% \{ scale: .95 1.05; \} 80% \{ scale: 1.02 .98; \} \}
@keyframes lab-button-bounce \{ 0%, 100% \{ translate: 0 0; \} 45% \{ translate: 0 -3px; \} 60% \{ translate: 0 1px; \} \}
@keyframes lab-button-tilt \{ 0%, 100% \{ rotate: -1.5deg; \} 50% \{ rotate: 1.5deg; \} \}
@keyframes lab-button-ring \{ 0% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.5); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(76,92,247,0); \} \}
@keyframes lab-button-sheen \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.3) saturate(1.1); \} \}
@keyframes lab-button-press \{ 0%, 100% \{ scale: 1; \} 50% \{ scale: .97; \} \}

/* ---------- Tabs: styles ---------- */
.sui-tabs.lab-t-underline \{ background: transparent; border: 0; \}
.sui-tabs.lab-t-underline .sui-tab \{ background: transparent; color: #94a3b8; border-radius: 0; \}
.sui-tabs.lab-t-underline .sui-tab.active \{ background: transparent; color: #e2e8f0; box-shadow: inset 0 -2px 0 #4c5cf7; \}
.sui-tabs.lab-t-pills \{ background: #0f172a; border: 1px solid #1e293b; border-radius: 999px; padding: 4px; \}
.sui-tabs.lab-t-pills .sui-tab \{ border-radius: 999px; padding: 6px 14px; color: #94a3b8; \}
.sui-tabs.lab-t-pills .sui-tab.active \{ background: #4c5cf7; color: #fff; box-shadow: none; \}
.sui-tabs.lab-t-segmented \{ background: #0b1220; border: 1px solid #1e293b; border-radius: 8px; padding: 3px; \}
.sui-tabs.lab-t-segmented .sui-tab \{ border-radius: 6px; color: #94a3b8; \}
.sui-tabs.lab-t-segmented .sui-tab.active \{ background: #1e293b; color: #e2e8f0; box-shadow: none; \}
.sui-tabs.lab-t-boxed \{ background: #111827; border: 1px solid #334155; border-radius: 8px; \}
.sui-tabs.lab-t-boxed .sui-tab \{ border-radius: 1px; color: #94a3b8; \}
.sui-tabs.lab-t-boxed .sui-tab.active \{ background: #1e293b; color: #fff; box-shadow: none; \}
.sui-tabs.lab-t-brutalist \{ background: #facc15; border: 3px solid #111; border-radius: 0; \}
.sui-tabs.lab-t-brutalist .sui-tab \{ border-radius: 0; color: #111; font-weight: 700; \}
.sui-tabs.lab-t-brutalist .sui-tab.active \{ background: #111; color: #facc15; box-shadow: none; \}
.sui-tabs.lab-t-neon \{ background: #07131a; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 18px rgba(34,211,238,.25); \}
.sui-tabs.lab-t-neon .sui-tab \{ border-radius: 6px; color: #67e8f9; \}
.sui-tabs.lab-t-neon .sui-tab.active \{ background: #0e7490; color: #ecfeff; box-shadow: none; \}
/* ---------- Tabs: animations ---------- */
.lab-t-anim-pop .sui-tab.active \{ animation: lab-tab-pop .32s cubic-bezier(.2,1.4,.4,1); \}
.lab-t-anim-slide .sui-tab.active \{ animation: lab-tab-slide .26s ease-out; \}
.lab-t-anim-glow .sui-tab.active \{ animation: lab-tab-glow 2.4s ease-in-out infinite; \}
.lab-t-anim-sweep .sui-tab.active \{ animation: lab-tab-sweep .45s ease-out; \}
.lab-t-anim-none .sui-tab.active \{ animation: none; \}
@keyframes lab-tab-pop \{ 0% \{ scale: .95; \} 60% \{ scale: 1.03; \} 100% \{ scale: 1; \} \}
@keyframes lab-tab-slide \{ from \{ opacity: .35; translate: -6px 0; \} to \{ opacity: 1; translate: 0 0; \} \}
@keyframes lab-tab-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.22); \} \}
@keyframes lab-tab-sweep \{ from \{ background-size: 0% 100%; \} to \{ background-size: 100% 100%; \} \}

/* ---------- Input: styles ---------- */
.sui-input.lab-i-outline \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 8px; \}
.sui-input.lab-i-filled \{ background: #1e293b; color: #e2e8f0; border: 1px solid transparent; border-radius: 8px; \}
.sui-input.lab-i-underline \{ background: transparent; color: #e2e8f0; border: 0; border-bottom: 2px solid #475569; border-radius: 0; \}
.sui-input.lab-i-pill \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 999px; padding-left: 16px; padding-right: 16px; \}
.sui-input.lab-i-glass \{ background: rgba(148,163,184,.14); color: #e2e8f0; border: 1px solid rgba(255,255,255,.24); border-radius: 10px; backdrop-filter: blur(8px); \}
.sui-input.lab-i-brutalist \{ background: #fff; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; \}
/* ---------- Input: focus animations ---------- */
.sui-input.lab-i-anim-ring:focus \{ animation: lab-focus-ring .5s ease-out; \}
.sui-input.lab-i-anim-glow:focus \{ animation: lab-focus-glow 1.8s ease-in-out infinite; \}
.sui-input.lab-i-anim-lift:focus \{ animation: lab-focus-lift .3s ease-out; \}
.sui-input.lab-i-anim-shake:focus \{ animation: lab-focus-shake .4s ease-in-out; \}
.sui-input.lab-i-anim-none \{ animation: none; \}
@keyframes lab-focus-ring \{ 0% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.45); \} 100% \{ box-shadow: 0 0 0 5px rgba(76,92,247,0); \} \}
@keyframes lab-focus-glow \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 5px rgba(76,92,247,.55)); \} \}
@keyframes lab-focus-lift \{ 50% \{ translate: 0 -1px; \} \}
@keyframes lab-focus-shake \{ 0%, 100% \{ translate: 0 0; \} 25% \{ translate: -2px 0; \} 75% \{ translate: 2px 0; \} \}

/* ---------- Dropdown styles ---------- */
.lab-dd-classic .sui-menu-button \{ background: #2d3340; color: #e6e6e6; border: 1px solid #3d4556; border-radius: 8px; \}
.lab-dd-pill .sui-menu-button \{ background: #4c5cf7; color: #fff; border: 0; border-radius: 999px; padding: 7px 16px; \}
.lab-dd-neon .sui-menu-button \{ background: #07131a; color: #22d3ee; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 12px rgba(34,211,238,.4); \}
.lab-dd-brutalist .sui-menu-button \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; font-weight: 700; \}
/* ---------- Popover styles ---------- */
.lab-pop-card .sui-popover-panel \{ background: #161d2b; color: #e2e8f0; border: 1px solid #2a3446; border-radius: 12px; box-shadow: 0 16px 36px rgba(0,0,0,.45); \}
.lab-pop-glass .sui-popover-panel \{ background: rgba(148,163,184,.16); color: #e2e8f0; border: 1px solid rgba(255,255,255,.24); border-radius: 14px; backdrop-filter: blur(12px); \}
.lab-pop-neon .sui-popover-panel \{ background: #07131a; color: #a5f3fc; border: 1px solid #22d3ee; border-radius: 12px; box-shadow: 0 0 22px rgba(34,211,238,.4); \}
.lab-pop-brutalist .sui-popover-panel \{ background: #facc15; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 6px 6px 0 #111; \}
/* ---------- Combobox styles ---------- */
.lab-cb-classic .sui-combobox-input \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 8px; \}
.lab-cb-pill .sui-combobox-input \{ background: #0f172a; color: #e2e8f0; border: 1px solid #3d4556; border-radius: 999px; padding-left: 16px; padding-right: 16px; \}
.lab-cb-neon .sui-combobox-input \{ background: #07131a; color: #a5f3fc; border: 1px solid #22d3ee; border-radius: 8px; box-shadow: 0 0 12px rgba(34,211,238,.35); \}
.lab-cb-brutalist .sui-combobox-input \{ background: #fff; color: #111; border: 3px solid #111; border-radius: 0; box-shadow: 4px 4px 0 #111; \}

/* ---------- Reduced motion ---------- */
@media (prefers-reduced-motion: reduce) \{
  .lab-b-anim-pulse, .lab-b-anim-glow, .lab-b-anim-float, .lab-b-anim-blink, .lab-b-anim-tilt, .lab-b-anim-ping, .lab-b-anim-bounce, .lab-b-anim-jelly, .lab-b-anim-sheen,
  .lab-a-anim-pulse, .lab-a-anim-wobble, .lab-a-anim-breathe, .lab-a-anim-flash,
  .lab-c-anim-float, .lab-c-anim-halo, .lab-c-anim-drift,
  .lab-btn-anim-pulse, .lab-btn-anim-float, .lab-btn-anim-glow, .lab-btn-anim-jelly, .lab-btn-anim-bounce, .lab-btn-anim-tilt, .lab-btn-anim-ring, .lab-btn-anim-sheen, .lab-btn-anim-press,
  .lab-t-anim-glow .sui-tab.active,
  .sui-spinner.lab-s-anim-spin, .sui-spinner.lab-s-anim-pulse, .sui-spinner.lab-s-anim-flip, .sui-spinner.lab-s-anim-reverse, .sui-spinner.lab-s-anim-orbit, .sui-spinner.lab-s-anim-wobble, .sui-spinner.lab-s-anim-bounce, .sui-spinner.lab-s-anim-ripple, .sui-spinner.lab-s-anim-swing,
  .sui-input.lab-i-anim-glow:focus \{ animation: none !important; \}
  .sui-progress.lab-p-anim-pulse::-webkit-progress-value, .sui-progress.lab-p-anim-stripes::-webkit-progress-value, .sui-progress.lab-p-anim-shimmer::-webkit-progress-value, .sui-progress.lab-p-anim-glow::-webkit-progress-value, .sui-progress.lab-p-anim-indeterminate::-webkit-progress-value, .sui-progress.lab-p-anim-aurora::-webkit-progress-value, .sui-progress.lab-p-anim-breathe::-webkit-progress-value, .sui-progress.lab-p-anim-ticker::-webkit-progress-value, .sui-progress.lab-p-anim-spark::-webkit-progress-value, .sui-progress.lab-p-anim-wave::-webkit-progress-value \{ animation: none !important; \}
\}
"

view main {
  title "Animation lab"
  col class: "lab-page" {
    h1 "Animation lab" class: "lab-head"
    p "Only controls whose graphics can be animated. Every style and every animation is chosen live; the page does not jump thanks to the re-render preservation in the runtime." class: "lab-note"

    div class: "lab-grid" {
      card class: "lab-card" {
        h3 "Progress"
        div class: "lab-preview" {
          progress value: 62 max: 100 class: pStyle + " " + pAnim
        }
        div class: "lab-fields" {
          select "Style" value: pStyle {
          					option "Classic" value: "lab-p-classic"
          					option "Thin line" value: "lab-p-thin"
          					option "Segments" value: "lab-p-segments"
          					option "Stripes" value: "lab-p-stripes"
          					option "Square" value: "lab-p-square"
          					option "Neon" value: "lab-p-neon"
          					option "Dashed" value: "lab-p-dashed"
          }
          select "Animation" value: pAnim {
          					option "Pulse" value: "lab-p-anim-pulse"
          					option "Stripes" value: "lab-p-anim-stripes"
          					option "Shimmer" value: "lab-p-anim-shimmer"
          					option "Glow" value: "lab-p-anim-glow"
          					option "Indeterminate" value: "lab-p-anim-indeterminate"
          					option "Aurora" value: "lab-p-anim-aurora"
          					option "Breathe" value: "lab-p-anim-breathe"
          					option "Ticker" value: "lab-p-anim-ticker"
          					option "Spark" value: "lab-p-anim-spark"
          					option "Wave" value: "lab-p-anim-wave"
          					option "None" value: "lab-p-anim-none"
          }
        }
        p "7 shapes x 11 animations, chosen live." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Spinner"
        div class: "lab-preview" {
          spinner class: sShape + " " + sAnim
        }
        div class: "lab-fields" {
          select "Shape" value: sShape {
          					option "Ring" value: "lab-s-ring"
          					option "Tri dots" value: "lab-s-dots"
          					option "Bars" value: "lab-s-bars"
          					option "Orbit" value: "lab-s-orbit"
          					option "Comet" value: "lab-s-comet"
          					option "Pulse" value: "lab-s-pulse"
          					option "Square" value: "lab-s-square"
          }
          select "Animation" value: sAnim {
          					option "Spin" value: "lab-s-anim-spin"
          					option "Pulse" value: "lab-s-anim-pulse"
          					option "Flip" value: "lab-s-anim-flip"
          					option "Reverse" value: "lab-s-anim-reverse"
          					option "Orbit" value: "lab-s-anim-orbit"
          					option "Wobble" value: "lab-s-anim-wobble"
          					option "Bounce" value: "lab-s-anim-bounce"
          					option "Ripple" value: "lab-s-anim-ripple"
          					option "Swing" value: "lab-s-anim-swing"
          					option "None" value: "lab-s-anim-none"
          }
        }
        p "Seven geometries, ten motions." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Badge"
        div class: "lab-preview" {
          row {
            badge "LIVE" class: bStyle + " " + bAnim
          }
        }
        div class: "lab-fields" {
          select "Style" value: bStyle {
          					option "Pill" value: "lab-b-pill"
          					option "Tag" value: "lab-b-tag"
          					option "Outline" value: "lab-b-outline"
          					option "Neon" value: "lab-b-neon"
          					option "Status" value: "lab-b-status"
          					option "Brutalist" value: "lab-b-brutalist"
          					option "Glass" value: "lab-b-glass"
          }
          select "Animation" value: bAnim {
          					option "Pulse" value: "lab-b-anim-pulse"
          					option "Glow" value: "lab-b-anim-glow"
          					option "Float" value: "lab-b-anim-float"
          					option "Pop" value: "lab-b-anim-pop"
          					option "Blink" value: "lab-b-anim-blink"
          					option "Tilt" value: "lab-b-anim-tilt"
          					option "Ping" value: "lab-b-anim-ping"
          					option "Bounce" value: "lab-b-anim-bounce"
          					option "Jelly" value: "lab-b-anim-jelly"
          					option "Sheen" value: "lab-b-anim-sheen"
          					option "None" value: "lab-b-anim-none"
          }
        }
        p "Tag, status dot, neon, brutalist and more." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Alert"
        div class: "lab-preview" {
          alert "Deploy completed" class: aStyle + " " + aAnim
        }
        div class: "lab-fields" {
          select "Style" value: aStyle {
          					option "Accent bar" value: "lab-a-bar"
          					option "Filled" value: "lab-a-filled"
          					option "Outline" value: "lab-a-outline"
          					option "Glass" value: "lab-a-glass"
          					option "Brutalist" value: "lab-a-brutalist"
          					option "Info" value: "lab-a-info"
          }
          select "Animation" value: aAnim {
          					option "Fade in" value: "lab-a-anim-fade"
          					option "Slide in" value: "lab-a-anim-slide"
          					option "Shake" value: "lab-a-anim-shake"
          					option "Pulse" value: "lab-a-anim-pulse"
          					option "Zoom in" value: "lab-a-anim-zoom"
          					option "Bounce in" value: "lab-a-anim-bounce-in"
          					option "Slide right" value: "lab-a-anim-slide-right"
          					option "Wobble" value: "lab-a-anim-wobble"
          					option "Breathe" value: "lab-a-anim-breathe"
          					option "Flash" value: "lab-a-anim-flash"
          					option "None" value: "lab-a-anim-none"
          }
        }
        p "Entrances, loops and attention effects." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Card"
        div class: "lab-preview" {
          card class: cStyle + " " + cAnim {
            strong "Aurora card"
            p "Surface and motion are independent." class: "lab-note"
          }
        }
        div class: "lab-fields" {
          select "Style" value: cStyle {
          					option "Elevated" value: "lab-c-elevated"
          					option "Flat" value: "lab-c-flat"
          					option "Outline" value: "lab-c-outline"
          					option "Glass" value: "lab-c-glass"
          					option "Neon" value: "lab-c-neon"
          					option "Brutalist" value: "lab-c-brutalist"
          					option "Dashed" value: "lab-c-dashed"
          }
          select "Animation" value: cAnim {
          					option "Fade in" value: "lab-c-anim-fade"
          					option "Float" value: "lab-c-anim-float"
          					option "Rise" value: "lab-c-anim-rise"
          					option "Zoom in" value: "lab-c-anim-zoom"
          					option "Tilt in" value: "lab-c-anim-tilt"
          					option "Halo" value: "lab-c-anim-halo"
          					option "Drift" value: "lab-c-anim-drift"
          					option "Unfold" value: "lab-c-anim-unfold"
          					option "None" value: "lab-c-anim-none"
          }
        }
        p "Seven surfaces, eight motions." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Button"
        div class: "lab-preview" {
          button "Press me" class: btnStyle + " " + btnAnim onClick: noop
        }
        div class: "lab-fields" {
          select "Style" value: btnStyle {
          					option "Solid" value: "lab-btn-solid"
          					option "Gradient" value: "lab-btn-gradient"
          					option "Outline" value: "lab-btn-outline"
          					option "Ghost" value: "lab-btn-ghost"
          					option "Neon" value: "lab-btn-neon"
          					option "3D" value: "lab-btn-3d"
          					option "Pill" value: "lab-btn-pill"
          					option "Brutalist" value: "lab-btn-brutalist"
          }
          select "Animation" value: btnAnim {
          					option "Pulse" value: "lab-btn-anim-pulse"
          					option "Float" value: "lab-btn-anim-float"
          					option "Glow" value: "lab-btn-anim-glow"
          					option "Jelly" value: "lab-btn-anim-jelly"
          					option "Bounce" value: "lab-btn-anim-bounce"
          					option "Tilt" value: "lab-btn-anim-tilt"
          					option "Ring" value: "lab-btn-anim-ring"
          					option "Sheen" value: "lab-btn-anim-sheen"
          					option "Press" value: "lab-btn-anim-press"
          					option "None" value: "lab-btn-anim-none"
          }
        }
        p "Solid, gradient, 3D, ghost, brutalist..." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Tabs"
        div class: "lab-preview" {
          tabs "Sections" class: tStyle + " " + tAnim {
            tab "One" {
              p "First tab" class: "lab-note"
            }
            tab "Two" {
              p "Second tab" class: "lab-note"
            }
          }
        }
        div class: "lab-fields" {
          select "Style" value: tStyle {
          					option "Underline" value: "lab-t-underline"
          					option "Pills" value: "lab-t-pills"
          					option "Segmented" value: "lab-t-segmented"
          					option "Boxed" value: "lab-t-boxed"
          					option "Brutalist" value: "lab-t-brutalist"
          					option "Neon" value: "lab-t-neon"
          }
          select "Animation" value: tAnim {
          					option "Pop" value: "lab-t-anim-pop"
          					option "Slide" value: "lab-t-anim-slide"
          					option "Glow" value: "lab-t-anim-glow"
          					option "Sweep" value: "lab-t-anim-sweep"
          					option "None" value: "lab-t-anim-none"
          }
        }
        p "The effect runs on the active tab." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Input"
        div class: "lab-preview" {
          input "Focus me" value: name class: iStyle + " " + iAnim
        }
        div class: "lab-fields" {
          select "Style" value: iStyle {
          					option "Outline" value: "lab-i-outline"
          					option "Filled" value: "lab-i-filled"
          					option "Underline" value: "lab-i-underline"
          					option "Pill" value: "lab-i-pill"
          					option "Glass" value: "lab-i-glass"
          					option "Brutalist" value: "lab-i-brutalist"
          }
          select "Animation" value: iAnim {
          					option "Focus ring" value: "lab-i-anim-ring"
          					option "Focus glow" value: "lab-i-anim-glow"
          					option "Focus lift" value: "lab-i-anim-lift"
          					option "Focus shake" value: "lab-i-anim-shake"
          					option "None" value: "lab-i-anim-none"
          }
        }
        p "Click the field to see the focus effect." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Checkbox"
        div class: "lab-preview" {
          checkbox "Accept terms" checked: chkOn shape: chkShape theme: chkTheme animation: chkAnim duration: "300ms" easing: "ease-out"
        }
        div class: "lab-fields" {
          select "Shape" value: chkShape {
          					option "Default" value: "default"
          					option "Rounded" value: "rounded"
          					option "Square" value: "square"
          					option "Circle" value: "circle"
          					option "iOS" value: "ios"
          					option "Minimal" value: "minimal"
          					option "Soft" value: "soft"
          }
          select "Theme" value: chkTheme {
          					option "Default" value: "default"
          					option "Accent" value: "accent"
          					option "Success" value: "success"
          					option "Warning" value: "warning"
          					option "Danger" value: "danger"
          					option "Mono" value: "mono"
          					option "Dark" value: "dark"
          					option "Light" value: "light"
          }
          select "Animation" value: chkAnim class: "lab-wide" {
          					option "Scale" value: "scale"
          					option "Morph" value: "morph"
          					option "Fade" value: "fade"
          					option "Slide" value: "slide"
          					option "None" value: "none"
          }
        }
        p "Shape, colour theme and transition are separate." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Switch"
        div class: "lab-preview" {
          switch "Enabled" checked: swOn shape: swShape theme: swTheme animation: swAnim duration: "300ms" easing: "ease-out"
        }
        div class: "lab-fields" {
          select "Shape" value: swShape {
          					option "Default" value: "default"
          					option "Rounded" value: "rounded"
          					option "Square" value: "square"
          					option "Pill" value: "pill"
          					option "iOS" value: "ios"
          					option "Material" value: "material"
          					option "Minimal" value: "minimal"
          					option "Soft" value: "soft"
          }
          select "Theme" value: swTheme {
          					option "Default" value: "default"
          					option "Accent" value: "accent"
          					option "Success" value: "success"
          					option "Warning" value: "warning"
          					option "Danger" value: "danger"
          					option "Mono" value: "mono"
          					option "Dark" value: "dark"
          					option "Light" value: "light"
          }
          select "Animation" value: swAnim class: "lab-wide" {
          					option "Scale" value: "scale"
          					option "Morph" value: "morph"
          					option "Fade" value: "fade"
          					option "Slide" value: "slide"
          					option "None" value: "none"
          }
        }
        p "Nine shapes and eight themes at runtime." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Radio"
        div class: "lab-preview" {
          radio "Selected" checked: rdOn shape: rdShape theme: rdTheme animation: rdAnim duration: "300ms" easing: "ease-out"
        }
        div class: "lab-fields" {
          select "Shape" value: rdShape {
          					option "Default" value: "default"
          					option "Compact" value: "compact"
          					option "Large" value: "large"
          					option "Soft" value: "soft"
          					option "Minimal" value: "minimal"
          }
          select "Theme" value: rdTheme {
          					option "Default" value: "default"
          					option "Accent" value: "accent"
          					option "Success" value: "success"
          					option "Warning" value: "warning"
          					option "Danger" value: "danger"
          					option "Mono" value: "mono"
          					option "Dark" value: "dark"
          					option "Light" value: "light"
          }
          select "Animation" value: rdAnim class: "lab-wide" {
          					option "Scale" value: "scale"
          					option "Morph" value: "morph"
          					option "Fade" value: "fade"
          					option "Slide" value: "slide"
          					option "None" value: "none"
          }
        }
        p "Independent radio; group them with name:." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Dropdown"
        div class: "lab-preview" {
          dropdown "Open menu" class: ddStyle animation: ddAnim duration: "200ms" {
            dropdownitem "Edit"
            dropdownitem "Duplicate"
            dropdownitem "Delete"
          }
        }
        div class: "lab-fields" {
          select "Style" value: ddStyle {
          					option "Classic" value: "lab-dd-classic"
          					option "Pill" value: "lab-dd-pill"
          					option "Neon" value: "lab-dd-neon"
          					option "Brutalist" value: "lab-dd-brutalist"
          }
          select "Animation" value: ddAnim {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
        }
        p "The animation plays when the panel opens." class: "lab-note"
      }

      card class: "lab-card" {
        h3 "Popover"
        div class: "lab-preview" {
          popover "Open popover" class: popStyle animation: popAnim duration: "180ms" {
            p "Anchored panel." class: "lab-note"
          }
        }
        div class: "lab-fields" {
          select "Style" value: popStyle {
          					option "Card" value: "lab-pop-card"
          					option "Glass" value: "lab-pop-glass"
          					option "Neon" value: "lab-pop-neon"
          					option "Brutalist" value: "lab-pop-brutalist"
          }
          select "Animation" value: popAnim {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
        }
      }

      card class: "lab-card" {
        h3 "Combobox"
        div class: "lab-preview" {
          combobox "Country" class: cbStyle options: "Italy,France,Germany" animation: cbAnim duration: "200ms"
        }
        div class: "lab-fields" {
          select "Style" value: cbStyle {
          					option "Classic" value: "lab-cb-classic"
          					option "Pill" value: "lab-cb-pill"
          					option "Neon" value: "lab-cb-neon"
          					option "Brutalist" value: "lab-cb-brutalist"
          }
          select "Animation" value: cbAnim {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
        }
      }

      card class: "lab-card" {
        h3 "Command palette"
        div class: "lab-preview" {
          button "Open commands" onClick: openCommands
          commandbar "Commands" open: showCommands hotkey: "k" animation: cmdAnim duration: "200ms" {
            command "New" onClick: noop
            command "Settings" onClick: noop
          }
        }
        div class: "lab-fields" {
          select "Animation" value: cmdAnim class: "lab-wide" {
          					option "Slide" value: "slide"
          					option "Fade" value: "fade"
          					option "Scale" value: "scale"
          					option "None" value: "none"
          }
        }
        p "Press Ctrl+K or use the button." class: "lab-note"
      }
    }
  }
}
```

## Run it

```bash
simpleui check examples/animation_lab/app.sui --json
simpleui serve examples/animation_lab/app.sui --watch
```

`lab.css` is the readable stylesheet; `build_example.py` inlines it into
`app.sui` (escaping braces and quotes). Edit the CSS file and rerun the script
to regenerate the app and this README source block.
