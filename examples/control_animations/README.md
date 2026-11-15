# Control animations

A gallery of the control animation library. Each tile runs one effect; the same
effects are available in the designer without writing CSS.

## What it shows

- **Button / jelly**: a soft squash-and-stretch loop that keeps the button
  readable while it draws attention.
- **Badge / ping**: a radiating ring, the classic live-status cue.
- **Card / halo**: a breathing coloured glow around the surface.
- **Progress / wave**: vertical bands that slide up and down inside the value.
- **Spinner / orbit**: rotation with a small lateral orbit, less mechanical than
  a plain spin.
- **Alert / breathe**: a coloured drop-shadow that pulses with the alert border.

Every tile also ships a `prefers-reduced-motion` guard: when the operating
system asks for less motion, decorative effects stop and the controls keep their
colours and layout.

## The full SUI source

```sui
css ".anim-page \{ max-width: 1040px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \} .anim-grid \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 14px; \} .anim-tile \{ background: #141a27; border: 1px solid #2a3345; border-radius: 14px; padding: 16px; display: flex; flex-direction: column; gap: 12px; min-height: 128px; \} .anim-label \{ font-size: 11px; letter-spacing: .09em; text-transform: uppercase; color: #8fa0b8; \} .anim-note \{ color: #8fa0b8; font-size: 13px; \} .anim-jelly \{ animation: demo-jelly 2.4s ease-in-out infinite; \} .anim-ping \{ animation: demo-ping 1.8s ease-out infinite; \} .anim-halo \{ animation: demo-halo 3s ease-in-out infinite; \} .sui-progress.anim-wave::-webkit-progress-value \{ animation: demo-wave 1.8s ease-in-out infinite; \} .sui-progress.anim-wave::-moz-progress-bar \{ animation: demo-wave 1.8s ease-in-out infinite; \} .sui-spinner.anim-orbit \{ animation: demo-orbit 1s linear infinite; \} .anim-breathe \{ animation: demo-breathe 2.6s ease-in-out infinite; \} @keyframes demo-jelly \{ 0%, 100% \{ scale: 1 1; \} 30% \{ scale: 1.06 .94; \} 55% \{ scale: .95 1.05; \} 80% \{ scale: 1.02 .98; \} \} @keyframes demo-ping \{ 0% \{ box-shadow: 0 0 0 0 rgba(74,222,128,.6); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(74,222,128,0); \} \} @keyframes demo-halo \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,0); \} 50% \{ box-shadow: 0 0 26px 2px rgba(76,92,247,.4); \} \} @keyframes demo-wave \{ 0%, 100% \{ background-position: 0 0; \} 50% \{ background-position: 0 100%; \} \} @keyframes demo-orbit \{ 0% \{ transform: rotate(0deg) translateX(0); \} 50% \{ transform: rotate(180deg) translateX(3px); \} 100% \{ transform: rotate(360deg) translateX(0); \} \} @keyframes demo-breathe \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 7px #f59e0b); \} \} @media (prefers-reduced-motion: reduce) \{ .anim-jelly, .anim-ping, .anim-halo, .anim-breathe, .sui-spinner.anim-orbit \{ animation: none !important; \} .sui-progress.anim-wave::-webkit-progress-value, .sui-progress.anim-wave::-moz-progress-bar \{ animation: none !important; \} \}"

state clicks = 0

fn addClick() -> null {
  clicks = clicks + 1
}

view main {
  col class: "anim-page" {
    h1 "Control animation gallery"
    p "Each tile runs one effect from the component animation library. The same effects are selectable in the designer, per control or for the whole project." class: "anim-note"

    div class: "anim-grid" {
      div class: "anim-tile" {
        span "Button / jelly" class: "anim-label"
        button "Press me" class: "anim-jelly" onClick: addClick
        p "{str(clicks)} clicks" style: "margin:0;font-size:12px;color:#8fa0b8"
      }
      div class: "anim-tile" {
        span "Badge / ping" class: "anim-label"
        row {
          badge "LIVE" class: "anim-ping"
        }
      }
      div class: "anim-tile" {
        span "Card / halo" class: "anim-label"
        card class: "anim-halo" {
          strong "Aurora card"
          p "A breathing coloured halo." style: "margin:0;font-size:12px;color:#8fa0b8"
        }
      }
      div class: "anim-tile" {
        span "Progress / wave" class: "anim-label"
        progress value: 68 max: 100 class: "anim-wave"
      }
      div class: "anim-tile" {
        span "Spinner / orbit" class: "anim-label"
        row {
          spinner class: "anim-orbit"
          span "Loading" style: "font-size:13px;color:#8fa0b8"
        }
      }
      div class: "anim-tile" {
        span "Alert / breathe" class: "anim-label"
        alert "Saved with warnings" class: "anim-breathe"
      }
    }

    p "Reduced motion: when the operating system asks for less motion, the animations stop while colours and layout stay unchanged." class: "anim-note"
  }
}
```

## Use it from the designer

1. Select a control (or open **Preferences → Project preferences** for the whole
   project).
2. Open the component panel with **Settings...**.
3. Choose a style preset for the palette, then open **Animation** (or
   **Effects** for Badge, Alert, Card, Button and Toast) and pick an effect.
4. Press **Apply**. The designer writes the colour variables and the shared
   rules, including the keyframes.

The animation effects are listed in `docs/GUI_EDITOR.md` under
**Style presets and animations**.

## Build and run

```bash
simpleui check examples/control_animations/app.sui --json
simpleui build examples/control_animations/app.sui --out examples/control_animations/dist
simpleui serve examples/control_animations/app.sui --watch
```

## Notes

- The effects are CSS-only and need no JavaScript or extra runtime support.
- Only `transform`, `scale`, `translate`, `rotate`, `filter`, `box-shadow` and
  `background-position` are animated, so colours and text stay crisp.
- The `prefers-reduced-motion` guard is part of the example; keep it when you
  copy the effects into your own project.
