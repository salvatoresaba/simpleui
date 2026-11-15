# All controls and animations

A single-page reference gallery with every SimpleUI control, semantic layout
element and an **animation lab**. The lab gives every animatable control a
selector, so the effect can be changed while the app runs.

It shows both animation mechanisms:

- **CSS animation library** on progress, spinner, badge, alert, card, button,
  tabs and input focus.
- **Runtime `animation` parameter** on dropdown, popover, combobox, command
  palette, checkbox, switch and radio.

The page also covers semantic content, form containers, data controls, media,
overlays and the prefabricated controls (splitter, dropzone, timeline, empty
state, skeleton, carousel, stepper).

## Animation lab

Each lab card has a control and an **Effect** selector. Choosing an option
updates the control immediately:

| Control | Selectable effects | Mechanism |
|---|---|---|
| Progress | wave, aurora, indeterminate, none | CSS class on the element |
| Spinner | orbit, reverse, wobble, none | CSS class on the element |
| Badge | ping, pop, tilt, none | CSS class on the element |
| Alert | breathe, wobble, flash, none | CSS class on the element |
| Card | halo, drift, none | CSS class on the element |
| Button | jelly, bounce, ring, none | CSS class on the element |
| Tabs | pop, slide, glow, none | CSS class on the active tab |
| Input | focus glow, focus ring, focus lift, none | CSS class on `:focus` |
| Dropdown | slide, fade, scale, none | runtime `animation` expression |
| Popover | scale, fade, slide, none | runtime `animation` expression |
| Combobox | slide, fade, scale, none | runtime `animation` expression |
| Command palette | slide, fade, scale, none | runtime `animation` expression |
| Checkbox | scale, morph, fade, slide, none | runtime `animation` expression |
| Switch | morph, scale, fade, slide, none | runtime `animation` expression |
| Radio | slide, scale, morph, none | runtime `animation` expression |

Every CSS loop has a `prefers-reduced-motion` guard: the animation stops while
colours and layout stay unchanged.

## Re-render behaviour

Changing an effect updates a state variable and re-renders the view. The runtime
keeps the page stable while it does so:

- focus and caret stay in the field being edited;
- select values are restored from their `value` binding;
- images, video, audio and iframes keep their loaded state instead of reloading;
- running CSS animations keep their phase instead of restarting from zero;
- the window scroll position is preserved.

The example therefore lets you switch effects in the lab and keep interacting
with every other control without the page visibly jumping.

## Controls covered

- **Text and layout**: h1-h6, paragraph, strong, em, mark, small, code, pre,
  blockquote, text, span, row, col, div, section, article, header, footer, nav,
  aside, lists (`ul`, `ol`, data `list`), link, image, video, audio, iframe,
  divider, hr, br.
- **Forms**: form, fieldset, legend, label, input, textarea, select/option,
  number, color, date, time, datepicker, file, slider, checkbox, radio, switch,
  button.
- **Navigation**: tabs/tab, accordion/accordionitem, menubar/menu/menuitem,
  dropdown/dropdownitem, popover, commandbar/command.
- **Data**: table/caption/thead/tbody/tfoot/trow/tcell, tree/treeitem, chart,
  graph, combobox.
- **Overlays and feedback**: modal/modalbutton, toast, drawer, progress, badge,
  alert, spinner, tooltip.
- **Prefabricated**: splitter/pane, dropzone, timeline/timelineitem, emptystate,
  skeleton, carousel/slide, stepper/step.

## Run it

```bash
simpleui check examples/all_controls_animations/app.sui --json
simpleui serve examples/all_controls_animations/app.sui --watch
```

For a static build, keep the `assets` folder next to the generated HTML:

```bash
simpleui build examples/all_controls_animations/app.sui --out examples/all_controls_animations/dist
cp -r examples/all_controls_animations/assets examples/all_controls_animations/dist/
```

The single-file build (`--single`) inlines CSS and JavaScript but not the media
assets, so serve the output folder or copy `assets` beside it.

## Full SUI source

```sui
state items = ["Alpha", "Beta", "Gamma"]
state chartData = "[\{\"month\":\"Jan\",\"value\":12\},\{\"month\":\"Feb\",\"value\":18\},\{\"month\":\"Mar\",\"value\":15\},\{\"month\":\"Apr\",\"value\":24\},\{\"month\":\"May\",\"value\":21\}]"
state graphData = "\{\"states\":[\{\"id\":\"start\",\"label\":\"Start\",\"shape\":\"rectangle\",\"x\":30,\"y\":70,\"w\":110,\"h\":54,\"tag\":\"entry\"\},\{\"id\":\"check\",\"label\":\"Valid?\",\"shape\":\"rhombus\",\"x\":180,\"y\":60,\"w\":110,\"h\":74,\"tag\":\"decision\"\},\{\"id\":\"done\",\"label\":\"Done\",\"shape\":\"rectangle\",\"x\":330,\"y\":70,\"w\":110,\"h\":54\}],\"transitions\":[\{\"from\":\"start\",\"to\":\"check\",\"label\":\"submit\"\},\{\"from\":\"check\",\"to\":\"done\",\"label\":\"yes\"\}]\}"
state name = "Ada"
state notes = "Multi-line notes"
state pick = "b"
state qty = 3
state tint = "#4c5cf7"
state day = "2026-11-15"
state at = "09:30"
state filePath = ""
state volume = 64
state t1 = true
state t2 = false
state r1 = true
state sw1 = true
state pAnim = "demo-progress-wave"
state sAnim = "demo-spinner-orbit"
state bAnim = "demo-badge-ping"
state aAnim = "demo-alert-breathe"
state cAnim = "demo-card-halo"
state btnAnim = "demo-button-jelly"
state tabsAnim = "demo-tabs-pop"
state iAnim = "demo-input-glow"
state ddAnim = "slide"
state popAnim = "scale"
state cbAnim = "slide"
state cmdAnim = "slide"
state chkAnim = "scale"
state swAnim = "morph"
state rdAnim = "slide"
state slide = 0
state step = 1
state showPalette = false
state showFilters = false
state loading = true
state projects = 0
state action = "none"
state lastCommand = "none"
state drops = ""

fn openModal() -> null {
	modal.open("all-demo")
}

fn closeModal() -> null {
	modal.close("all-demo")
}

fn showToast() -> null {
	toast.show("Saved with the success animation.", "success", 2500)
}

fn openFilters() -> null {
	showFilters = true
}

fn closeFilters() -> null {
	showFilters = false
}

fn openPalette() -> null {
	showPalette = true
}

fn commandNew() -> null {
	lastCommand = "new"
}

fn commandSettings() -> null {
	lastCommand = "settings"
}

fn nextStep() -> null {
	if step < 2 {
		step = step + 1
	}
}

fn prevStep() -> null {
	if step > 0 {
		step = step - 1
	}
}

fn toggleLoading() -> null {
	if loading {
		loading = false
	} else {
		loading = true
	}
}

fn addProject() -> null {
	projects = projects + 1
}

fn chooseAction() -> null {
	action = "clicked"
}

fn noop() -> null {
	return null
}

css "
.all-page \{ max-width: 1180px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \}
.all-note \{ color: #8fa0b8; font-size: 13px; margin: 0; \}
.all-grid \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 14px; align-items: start; \}
.lab-grid \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 14px; align-items: start; \}
.lab-card \{ background: #141a27; border: 1px solid #2a3345; border-radius: 14px; padding: 16px; display: flex; flex-direction: column; gap: 12px; \}
.lab-preview \{ min-height: 86px; display: flex; flex-direction: column; justify-content: center; gap: 8px; padding: 12px; border: 1px dashed #2a3345; border-radius: 10px; \}
.lab-card h3 \{ margin: 0; font-size: 15px; \}
.lab-card select \{ width: 100%; \}
.all-split \{ height: 170px; \}
.all-media \{ width: 100%; height: 72px; object-fit: cover; border-radius: 8px; \}
.all-video \{ width: 100%; border-radius: 8px; \}
.all-frame \{ width: 100%; height: 140px; border: 1px solid #2a3345; border-radius: 8px; \}
.all-form \{ display: flex; flex-direction: column; gap: 8px; \}
.all-form fieldset \{ border: 1px solid #2a3345; border-radius: 8px; padding: 10px; display: flex; flex-direction: column; gap: 8px; \}
.all-form legend \{ color: #8fa0b8; font-size: 12px; padding: 0 4px; \}

/* CSS animation library selectable from the lab selectors */
.sui-progress.demo-progress-wave::-webkit-progress-value \{ animation: kf-progress-wave 1.8s ease-in-out infinite; \}
.sui-progress.demo-progress-wave::-moz-progress-bar \{ animation: kf-progress-wave 1.8s ease-in-out infinite; \}
.sui-progress.demo-progress-aurora::-webkit-progress-value \{ animation: kf-progress-aurora 7s linear infinite; \}
.sui-progress.demo-progress-aurora::-moz-progress-bar \{ animation: kf-progress-aurora 7s linear infinite; \}
.sui-progress.demo-progress-indeterminate::-webkit-progress-value \{ background: linear-gradient(90deg, transparent 12%, #4c5cf7 42%, #6366f1 58%, transparent 88%); background-size: 200% 100%; animation: kf-progress-scan 1.25s ease-in-out infinite; \}
.sui-progress.demo-progress-indeterminate::-moz-progress-bar \{ background: linear-gradient(90deg, transparent 12%, #4c5cf7 42%, #6366f1 58%, transparent 88%); background-size: 200% 100%; animation: kf-progress-scan 1.25s ease-in-out infinite; \}
.sui-spinner.demo-spinner-orbit \{ animation: kf-spinner-orbit 1s linear infinite; \}
.sui-spinner.demo-spinner-reverse \{ animation: kf-spinner-reverse 1.1s linear infinite; \}
.sui-spinner.demo-spinner-wobble \{ animation: kf-spinner-wobble 1.4s ease-in-out infinite; \}
.demo-badge-ping \{ animation: kf-badge-ping 1.8s ease-out infinite; \}
.demo-badge-pop \{ animation: kf-badge-pop .42s cubic-bezier(.2,1.4,.4,1); \}
.demo-badge-tilt \{ animation: kf-badge-tilt 2.6s ease-in-out infinite; \}
.demo-alert-breathe \{ animation: kf-alert-breathe 2.6s ease-in-out infinite; \}
.demo-alert-wobble \{ animation: kf-alert-wobble 2.4s ease-in-out infinite; \}
.demo-alert-flash \{ animation: kf-alert-flash 1.4s ease-in-out infinite; \}
.demo-card-halo \{ animation: kf-card-halo 3s ease-in-out infinite; \}
.demo-card-drift \{ animation: kf-card-drift 5s ease-in-out infinite; \}
.demo-button-jelly \{ animation: kf-button-jelly 2.4s ease-in-out infinite; \}
.demo-button-bounce \{ animation: kf-button-bounce 1.8s ease-in-out infinite; \}
.demo-button-ring \{ animation: kf-button-ring 2s ease-out infinite; \}
.demo-tabs-pop .sui-tab.active \{ animation: kf-tab-pop .32s cubic-bezier(.2,1.4,.4,1); \}
.demo-tabs-slide .sui-tab.active \{ animation: kf-tab-slide .26s ease-out; \}
.demo-tabs-glow .sui-tab.active \{ animation: kf-tab-glow 2.4s ease-in-out infinite; \}
.sui-input.demo-input-ring:focus \{ animation: kf-focus-ring .5s ease-out; \}
.sui-input.demo-input-glow:focus \{ animation: kf-focus-glow 1.8s ease-in-out infinite; \}
.sui-input.demo-input-lift:focus \{ animation: kf-focus-lift .3s ease-out; \}

@keyframes kf-progress-wave \{ 0%, 100% \{ background-position: 0 0; \} 50% \{ background-position: 0 100%; \} \}
@keyframes kf-progress-aurora \{ to \{ filter: hue-rotate(360deg); \} \}
@keyframes kf-progress-scan \{ 0% \{ background-position: 130% 0; \} 100% \{ background-position: -130% 0; \} \}
@keyframes kf-spinner-orbit \{ 0% \{ transform: rotate(0deg) translateX(0); \} 50% \{ transform: rotate(180deg) translateX(3px); \} 100% \{ transform: rotate(360deg) translateX(0); \} \}
@keyframes kf-spinner-reverse \{ to \{ transform: rotate(-360deg); \} \}
@keyframes kf-spinner-wobble \{ 0%, 100% \{ transform: rotate(-14deg) scale(.94); \} 50% \{ transform: rotate(14deg) scale(1.06); \} \}
@keyframes kf-badge-ping \{ 0% \{ box-shadow: 0 0 0 0 rgba(74,222,128,.6); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(74,222,128,0); \} \}
@keyframes kf-badge-pop \{ 0% \{ opacity: 0; scale: .6; \} 65% \{ opacity: 1; scale: 1.1; \} 100% \{ opacity: 1; scale: 1; \} \}
@keyframes kf-badge-tilt \{ 0%, 100% \{ rotate: -3deg; \} 50% \{ rotate: 3deg; \} \}
@keyframes kf-alert-breathe \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 7px #f59e0b); \} \}
@keyframes kf-alert-wobble \{ 0%, 100% \{ rotate: 0deg; \} 25% \{ rotate: -1.4deg; \} 75% \{ rotate: 1.4deg; \} \}
@keyframes kf-alert-flash \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.25); \} \}
@keyframes kf-card-halo \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,0); \} 50% \{ box-shadow: 0 0 26px 2px rgba(76,92,247,.4); \} \}
@keyframes kf-card-drift \{ 0%, 100% \{ translate: 0 0; \} 50% \{ translate: 3px -4px; \} \}
@keyframes kf-button-jelly \{ 0%, 100% \{ scale: 1 1; \} 30% \{ scale: 1.06 .94; \} 55% \{ scale: .95 1.05; \} 80% \{ scale: 1.02 .98; \} \}
@keyframes kf-button-bounce \{ 0%, 100% \{ translate: 0 0; \} 45% \{ translate: 0 -3px; \} 60% \{ translate: 0 1px; \} \}
@keyframes kf-button-ring \{ 0% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.5); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(76,92,247,0); \} \}
@keyframes kf-tab-pop \{ 0% \{ scale: .95; \} 60% \{ scale: 1.03; \} 100% \{ scale: 1; \} \}
@keyframes kf-tab-slide \{ from \{ opacity: .35; translate: -6px 0; \} to \{ opacity: 1; translate: 0 0; \} \}
@keyframes kf-tab-glow \{ 0%, 100% \{ filter: brightness(1); \} 50% \{ filter: brightness(1.22); \} \}
@keyframes kf-focus-ring \{ 0% \{ box-shadow: 0 0 0 0 rgba(76,92,247,.45); \} 100% \{ box-shadow: 0 0 0 5px rgba(76,92,247,0); \} \}
@keyframes kf-focus-glow \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 5px rgba(76,92,247,.55)); \} \}
@keyframes kf-focus-lift \{ 50% \{ translate: 0 -1px; \} \}

@media (prefers-reduced-motion: reduce) \{
  .demo-badge-ping, .demo-badge-pop, .demo-badge-tilt,
  .demo-alert-breathe, .demo-alert-wobble, .demo-alert-flash,
  .demo-card-halo, .demo-card-drift,
  .demo-button-jelly, .demo-button-bounce, .demo-button-ring,
  .sui-spinner.demo-spinner-orbit, .sui-spinner.demo-spinner-reverse, .sui-spinner.demo-spinner-wobble,
  .sui-input.demo-input-glow:focus \{ animation: none !important; \}
  .sui-progress.demo-progress-wave::-webkit-progress-value, .sui-progress.demo-progress-wave::-moz-progress-bar,
  .sui-progress.demo-progress-aurora::-webkit-progress-value, .sui-progress.demo-progress-aurora::-moz-progress-bar,
  .sui-progress.demo-progress-indeterminate::-webkit-progress-value, .sui-progress.demo-progress-indeterminate::-moz-progress-bar \{ animation: none !important; \}
  .demo-tabs-pop .sui-tab.active, .demo-tabs-slide .sui-tab.active, .demo-tabs-glow .sui-tab.active \{ animation: none !important; \}
\}
"

view main {
	title "All controls and animations"
	col class: "all-page" {
		h1 "All controls and animations"
		p "The animation lab below selects the effect of every animatable control at runtime. The rest of the page is the full control reference." class: "all-note"

		h2 "Animation lab - pick an effect"
		div class: "lab-grid" {
			card class: "lab-card" {
				h3 "Progress"
				progress value: 64 max: 100 class: pAnim
				select "Effect" value: pAnim {
					option "Wave" value: "demo-progress-wave"
					option "Aurora" value: "demo-progress-aurora"
					option "Indeterminate" value: "demo-progress-indeterminate"
					option "None" value: "demo-none-progress"
				}
			}
			card class: "lab-card" {
				h3 "Spinner"
				div class: "lab-preview" {
					spinner class: sAnim
				}
				select "Effect" value: sAnim {
					option "Orbit" value: "demo-spinner-orbit"
					option "Reverse" value: "demo-spinner-reverse"
					option "Wobble" value: "demo-spinner-wobble"
					option "None" value: "demo-none-spinner"
				}
			}
			card class: "lab-card" {
				h3 "Badge"
				div class: "lab-preview" {
					row {
						badge "LIVE" class: bAnim
					}
				}
				select "Effect" value: bAnim {
					option "Ping" value: "demo-badge-ping"
					option "Pop" value: "demo-badge-pop"
					option "Tilt" value: "demo-badge-tilt"
					option "None" value: "demo-none-badge"
				}
			}
			card class: "lab-card" {
				h3 "Alert"
				div class: "lab-preview" {
					alert "Deploy completed" class: aAnim
				}
				select "Effect" value: aAnim {
					option "Breathe" value: "demo-alert-breathe"
					option "Wobble" value: "demo-alert-wobble"
					option "Flash" value: "demo-alert-flash"
					option "None" value: "demo-none-alert"
				}
			}
			card class: "lab-card" {
				h3 "Card"
				div class: "lab-preview" {
					card class: cAnim {
						strong "Card preview"
					}
				}
				select "Effect" value: cAnim {
					option "Halo" value: "demo-card-halo"
					option "Drift" value: "demo-card-drift"
					option "None" value: "demo-none-card"
				}
			}
			card class: "lab-card" {
				h3 "Button"
				div class: "lab-preview" {
					button "Press me" class: btnAnim onClick: noop
				}
				select "Effect" value: btnAnim {
					option "Jelly" value: "demo-button-jelly"
					option "Bounce" value: "demo-button-bounce"
					option "Ring" value: "demo-button-ring"
					option "None" value: "demo-none-button"
				}
			}
			card class: "lab-card" {
				h3 "Tabs"
				tabs "Sections" class: tabsAnim {
					tab "One" {
						p "First tab" class: "all-note"
					}
					tab "Two" {
						p "Second tab" class: "all-note"
					}
				}
				select "Effect" value: tabsAnim {
					option "Pop" value: "demo-tabs-pop"
					option "Slide" value: "demo-tabs-slide"
					option "Glow" value: "demo-tabs-glow"
					option "None" value: "demo-none-tabs"
				}
			}
			card class: "lab-card" {
				h3 "Input focus"
				div class: "lab-preview" {
					input "Focus me" value: name class: iAnim
				}
				select "Effect" value: iAnim {
					option "Glow" value: "demo-input-glow"
					option "Ring" value: "demo-input-ring"
					option "Lift" value: "demo-input-lift"
					option "None" value: "demo-none-input"
				}
			}
			card class: "lab-card" {
				h3 "Dropdown (runtime)"
				div class: "lab-preview" {
					dropdown "Open menu" animation: ddAnim duration: "180ms" {
						dropdownitem "One" onClick: chooseAction
						dropdownitem "Two" onClick: chooseAction
					}
				}
				select "Effect" value: ddAnim {
					option "Slide" value: "slide"
					option "Fade" value: "fade"
					option "Scale" value: "scale"
					option "None" value: "none"
				}
			}
			card class: "lab-card" {
				h3 "Popover (runtime)"
				div class: "lab-preview" {
					popover "Open popover" animation: popAnim duration: "160ms" {
						p "Popover content." class: "all-note"
					}
				}
				select "Effect" value: popAnim {
					option "Scale" value: "scale"
					option "Fade" value: "fade"
					option "Slide" value: "slide"
					option "None" value: "none"
				}
			}
			card class: "lab-card" {
				h3 "Combobox (runtime)"
				div class: "lab-preview" {
					combobox "Country" options: "Italy,France,Germany" animation: cbAnim duration: "180ms"
				}
				select "Effect" value: cbAnim {
					option "Slide" value: "slide"
					option "Fade" value: "fade"
					option "Scale" value: "scale"
					option "None" value: "none"
				}
			}
			card class: "lab-card" {
				h3 "Command palette (runtime)"
				div class: "lab-preview" {
					button "Open commands" onClick: openPalette
				}
				select "Effect" value: cmdAnim {
					option "Slide" value: "slide"
					option "Fade" value: "fade"
					option "Scale" value: "scale"
					option "None" value: "none"
				}
				commandbar "Commands" open: showPalette hotkey: "k" animation: cmdAnim duration: "200ms" {
					command "New" onClick: commandNew
					command "Settings" onClick: commandSettings
				}
			}
			card class: "lab-card" {
				h3 "Checkbox (runtime)"
				div class: "lab-preview" {
					checkbox "Accept terms" checked: t1 animation: chkAnim duration: "280ms"
				}
				select "Effect" value: chkAnim {
					option "Scale" value: "scale"
					option "Morph" value: "morph"
					option "Fade" value: "fade"
					option "Slide" value: "slide"
					option "None" value: "none"
				}
			}
			card class: "lab-card" {
				h3 "Switch (runtime)"
				div class: "lab-preview" {
					switch "Enabled" checked: sw1 animation: swAnim duration: "300ms"
				}
				select "Effect" value: swAnim {
					option "Morph" value: "morph"
					option "Scale" value: "scale"
					option "Fade" value: "fade"
					option "Slide" value: "slide"
					option "None" value: "none"
				}
			}
			card class: "lab-card" {
				h3 "Radio (runtime)"
				div class: "lab-preview" {
					radio "Option A" checked: r1 animation: rdAnim duration: "260ms"
				}
				select "Effect" value: rdAnim {
					option "Slide" value: "slide"
					option "Scale" value: "scale"
					option "Morph" value: "morph"
					option "None" value: "none"
				}
			}
		}

		h2 "Feedback controls"
		div class: "all-grid" {
			card {
				h3 "Progress"
				progress value: 46 max: 100 class: "demo-progress-wave"
			}
			card {
				h3 "Spinner"
				row {
					spinner class: "demo-spinner-orbit"
					p "Loading" class: "all-note"
				}
			}
			card {
				h3 "Badge"
				row {
					badge "New"
					badge "Live" class: "demo-badge-ping"
				}
			}
			card {
				h3 "Alert"
				alert "Heads up" class: "demo-alert-breathe"
			}
			card {
				h3 "Toast"
				button "Show toast" onClick: showToast
			}
			card {
				h3 "Tooltip"
				badge "Hover me" tooltip: "The tooltip repositions itself and uses its own animation."
			}
		}

		h2 "Form controls"
		div class: "all-grid" {
			card {
				h3 "Text fields"
				form {
					fieldset {
						legend "Profile"
						label "Name"
						input "Name" value: name
					}
				}
				textarea "Notes" value: notes
				select "Option" value: pick {
					option "A" value: "a"
					option "B" value: "b"
					option "C" value: "c"
				}
				number "Quantity" value: qty min: 0 max: 20
				color "Tint" value: tint
			}
			card {
				h3 "Date and pickers"
				date "Day" value: day
				time "At" value: at
				datepicker "Appointment" mode: "date" value: day
				file accept: ".txt,.csv" multiple: true pickFiles: filePath
				slider "Volume" value: volume min: 0 max: 100
			}
			card {
				h3 "Toggles"
				checkbox "Checked" checked: t1
				checkbox "Unchecked" checked: t2
				radio "Selected" checked: r1
				switch "Enabled" checked: sw1
			}
			card {
				h3 "Buttons"
				row {
					button "Primary" onClick: noop
					button "Secondary" onClick: noop
				}
				row {
					button "Open modal" onClick: openModal
					button "Drawer" onClick: openFilters
				}
			}
		}

		h2 "Navigation"
		div class: "all-grid" {
			card {
				h3 "Tabs"
				tabs "Sections" {
					tab "Overview" {
						p "Overview panel" class: "all-note"
					}
					tab "Details" {
						p "Details panel" class: "all-note"
					}
					tab "More" {
						p "More panel" class: "all-note"
					}
				}
			}
			card {
				h3 "Accordion"
				accordion "Frequently asked" {
					accordionitem "First section" {
						p "Accordion content." class: "all-note"
					}
					accordionitem "Second section" {
						p "Second content." class: "all-note"
					}
				}
			}
			card {
				h3 "Menu bar"
				menubar "Main" {
					menu "File" {
						menuitem "Open"
						menuitem "Save"
					}
					menu "Help" {
						menuitem "About"
					}
				}
			}
			card {
				h3 "Dropdown"
				dropdown "Actions" animation: "slide" duration: "180ms" {
					dropdownitem "Edit" onClick: chooseAction
					dropdownitem "Delete" onClick: chooseAction
				}
				p "Last action: {action}" class: "all-note"
			}
		}

		h2 "Data controls"
		div class: "all-grid" {
			card {
				h3 "Table"
				table {
					caption "Sample data"
					thead {
						trow {
							tcell "Name"
							tcell "Value"
						}
					}
					tbody {
						trow {
							tcell "Alpha"
							tcell "12"
						}
						trow {
							tcell "Beta"
							tcell "18"
						}
					}
					tfoot {
						trow {
							tcell "Total"
							tcell "30"
						}
					}
				}
			}
			card {
				h3 "Tree"
				tree "Project" {
					treeitem "src" open: true {
						treeitem "app.sui"
						treeitem "styles.sui"
					}
					treeitem "docs" {
						treeitem "README.md"
					}
				}
			}
			card {
				h3 "Combobox"
				combobox "Country" options: "Italy,France,Germany" animation: "slide" duration: "180ms"
			}
		}

		div class: "all-grid" {
			card {
				h3 "Chart"
				chart "Monthly" data: chartData chartType: "bar" x: "month" y: "value" height: "220px"
			}
			card {
				h3 "Graph"
				graph "Flow" data: graphData height: "220px"
			}
		}

		h2 "Overlays and prefabricated controls"
		div class: "all-grid" {
			card {
				h3 "Popover"
				popover "Details" animation: "scale" duration: "160ms" {
					p "Anchored popover content." class: "all-note"
				}
			}
			card {
				h3 "Command palette"
				row {
					button "Open commands" onClick: openPalette
					p "Last: {lastCommand}" class: "all-note"
				}
				commandbar "Commands" open: showPalette hotkey: "k" animation: "slide" duration: "200ms" {
					command "New" onClick: commandNew
					command "Settings" onClick: commandSettings
				}
			}
			card {
				h3 "Splitter"
				splitter direction: "horizontal" initial: "40%" min: "80px" class: "all-split" {
					pane {
						p "Left pane" class: "all-note"
					}
					pane {
						p "Right pane" class: "all-note"
					}
				}
			}
			card {
				h3 "Dropzone"
				dropzone "Files" dropFiles: drops multiple: true hint: "Drop files or click" {
					p "Drop files here." class: "all-note"
				}
				p "State: {drops}" class: "all-note"
			}
		}

		card {
			h3 "Timeline"
			timeline {
				timelineitem "Created" date: "10 Nov" status: "done"
				timelineitem "Built" date: "11 Nov" status: "done"
				timelineitem "Review" date: "12 Nov" status: "active"
				timelineitem "Ship" date: "13 Nov" status: "pending"
			}
		}

		div class: "all-grid" {
			card {
				h3 "Empty state"
				emptystate "No projects" icon: "?" hint: "Create the first project." {
					button "Create" onClick: addProject
				}
				p "Projects: {projects}" class: "all-note"
			}
			card {
				h3 "Skeleton"
				if loading {
					skeleton lines: 2 width: "60%" height: "14px"
					skeleton lines: 3 width: "100%" height: "10px"
				} else {
					p "Content loaded." class: "all-note"
				}
				button "Toggle" onClick: toggleLoading
			}
			card {
				h3 "Carousel"
				carousel value: slide height: "140px" {
					slide label: "One" {
						p "First slide" class: "all-note"
					}
					slide label: "Two" {
						p "Second slide" class: "all-note"
					}
				}
				p "Slide {slide + 1}" class: "all-note"
			}
			card {
				h3 "Stepper"
				stepper value: step {
					step "One" description: "First"
					step "Two" description: "Second"
					step "Three" description: "Third"
				}
				row {
					button "Back" onClick: prevStep
					button "Next" onClick: nextStep
				}
			}
		}

		card {
			h3 "Text, media and lists"
			section {
				header {
					h4 "Section header"
				}
				p "Section body."
				footer {
					small "Section footer"
				}
			}
			article {
				h5 "Article heading"
				h6 "Article subtitle"
			}
			nav {
				a "Navigation link" href: "#"
			}
			aside {
				p "Aside content." class: "all-note"
			}
			p "Paragraph with inline formatting."
			row {
				strong "Bold"
				em "Italic"
				mark "Highlight"
			}
			code "const x = 1"
			pre "line one\nline two"
			blockquote "A quoted line."
			text "A plain text element."
			ul {
				li "First item"
				li "Second item"
			}
			ol {
				li "Step one"
				li "Step two"
			}
			list item in items {
				text "{item}"
			}
			img src: "assets/preview.png" alt: "Preview image" class: "all-media"
			video src: "assets/clip.mp4" class: "all-video"
			audio src: "assets/tone.wav"
			iframe src: "assets/embed.html" sandbox: "" class: "all-frame"
			hr
			row {
				span "Inline span" class: "all-note"
				br
				span "Another span" class: "all-note"
			}
		}

		modal "All controls" name: "all-demo" {
			p "Modal content." class: "all-note"
			modalbutton "Close" onClick: closeModal
		}

		drawer "Filters" side: "right" width: "320px" open: showFilters duration: "280ms" easing: "ease-out" {
			p "Drawer content." class: "all-note"
			button "Close" onClick: closeFilters
		}

		toast
	}
}
```

## Files

- `app.sui` - the gallery.
- `assets/preview.png` - small PNG used by the image demo.
- `assets/tone.wav` - short tone used by the audio demo.
- `assets/clip.mp4` - one-second colour clip used by the video demo.
- `assets/embed.html` - sandboxed page used by the iframe demo.
- `build_example.py` - regenerates `app.sui`, escaping CSS/JSON braces.

## Notes

- The CSS effects use `transform`, `scale`, `translate`, `rotate`, `filter`,
  `box-shadow` and `background-position` only.
- The runtime animation parameter also accepts `duration`, `easing` and `delay`;
  this page uses different durations to make the difference visible.
- Toggle animations run when the control changes state, not on page load.
