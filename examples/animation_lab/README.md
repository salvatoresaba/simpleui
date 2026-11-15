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
| Spinner | Ring, Tri dots, Bars, Orbit, Comet, Pulse, Square | Spin, Pulse, Flip, Reverse, Orbit, Wobble, Bounce, Ripple, Swing, None |
| Badge | Pill, Tag, Outline, Neon, Status, Brutalist, Glass | Pulse, Glow, Float, Pop, Blink, Tilt, Ping, Bounce, Jelly, Sheen, None |
| Alert | Accent bar, Filled, Outline, Glass, Brutalist, Info | Fade in, Slide in, Shake, Pulse, Zoom in, Bounce in, Slide right, Wobble, Breathe, Flash, None |
| Card | Elevated, Flat, Outline, Glass, Neon, Brutalist, Dashed | Fade in, Float, Rise, Zoom in, Tilt in, Halo, Drift, Unfold, None |
| Button | Solid, Gradient, Outline, Ghost, Neon, 3D, Pill, Brutalist | Pulse, Float, Glow, Jelly, Bounce, Tilt, Ring, Sheen, Press, None |
| Tabs | Underline, Pills, Segmented, Boxed, Brutalist, Neon | Pop, Slide, Glow, Sweep, None |
| Input | Outline, Filled, Underline, Pill, Glass, Brutalist | Focus ring, Focus glow, Focus lift, Focus shake, None |
| Checkbox | 7 shapes | Scale, Morph, Fade, Slide, None |
| Switch | 8 shapes | Scale, Morph, Fade, Slide, None |
| Radio | 5 shapes | Scale, Morph, Fade, Slide, None |
| Dropdown | Classic, Pill, Neon, Brutalist | Slide, Fade, Scale, None |
| Popover | Card, Glass, Neon, Brutalist | Slide, Fade, Scale, None |
| Combobox | Classic, Pill, Neon, Brutalist | Slide, Fade, Scale, None |
| Command palette | - | Slide, Fade, Scale, None |

Checkbox, switch and radio also expose eight colour themes: Default, Accent,
Success, Warning, Danger, Mono, Dark and Light.

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

CSS-backed controls combine two classes (surface + motion). Overlay and toggle
controls use the runtime parameters instead:

```sui
dropdown "Menu" class: ddStyle animation: ddAnim
switch "Enabled" shape: swShape theme: swTheme animation: swAnim
```

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
