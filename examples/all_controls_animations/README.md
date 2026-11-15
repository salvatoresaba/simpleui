# All controls and animations

A single-page reference gallery with every SimpleUI control and semantic
layout element. It shows both animation mechanisms:

- **CSS animation library** on feedback controls: progress wave, spinner orbit,
  badge ping, alert breathe, card halo, button jelly and tab pop.
- **Runtime `animation` parameter** on overlays and toggles: dropdowns with
  `fade`, `scale` and `slide`; popover with `scale`; command palette with
  `slide`; checkbox, radio and switch with their own state transitions.

The page also covers semantic content, form containers, data controls, media,
overlays and the prefabricated controls (splitter, dropzone, timeline, empty
state, skeleton, carousel, stepper).

## Animation map

| Control | Mechanism | Effect |
|---|---|---|
| Progress | CSS | `demo-wave`: vertical bands slide inside the value |
| Spinner | CSS | `demo-orbit`: rotation with a lateral orbit |
| Badge | CSS | `demo-ping`: radiating status ring |
| Alert | CSS | `demo-breathe`: coloured glow pulse |
| Card | CSS | `demo-halo`: breathing coloured halo |
| Button | CSS | `demo-jelly`: squash and stretch loop |
| Tabs | CSS | `demo-tab-pop`: the active tab pops when selected |
| Dropdown | runtime `animation` | three instances: `fade`, `scale`, `slide` |
| Popover / command palette | runtime `animation` | `scale` and `slide` |
| Checkbox / radio / switch | runtime `animation` | `scale`, `fade`, `slide`, `morph` |

Every CSS loop has a `prefers-reduced-motion` guard: the animation stops while
colours and layout stay unchanged.

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
state on = true
state checked = true
state radio = true
state morphOn = true
state slide = 0
state step = 1
state showPalette = false
state showFilters = false
state loading = true
state projects = 0
state action = "none"
state lastCommand = "none"
state drops = ""
state items = ["Alpha", "Beta", "Gamma"]

fn countClick() -> null {
	notes = "Clicked"
}

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

fn chooseAction(a: string) -> null {
	action = a
}

fn inspectFiles() -> null {
	drops = "files received"
}

css "
.all-page \{ max-width: 1180px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 18px; \}
.all-note \{ color: #8fa0b8; font-size: 13px; margin: 0; \}
.all-grid \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 14px; align-items: start; \}
.all-card \{ background: #141a27; border: 1px solid #2a3345; border-radius: 14px; padding: 16px; display: flex; flex-direction: column; gap: 10px; \}
.all-card h3 \{ margin: 0; font-size: 15px; \}
.all-kbd \{ font-family: ui-monospace, Consolas, monospace; font-size: 11px; color: #c7d2fe; background: #1e2637; border: 1px solid #334155; border-radius: 4px; padding: 1px 5px; \}
.demo-jelly \{ animation: demo-jelly 2.4s ease-in-out infinite; \}
.demo-ping \{ animation: demo-ping 1.8s ease-out infinite; \}
.demo-halo \{ animation: demo-halo 3s ease-in-out infinite; \}
.sui-progress.demo-wave::-webkit-progress-value \{ animation: demo-wave 1.8s ease-in-out infinite; \}
.sui-progress.demo-wave::-moz-progress-bar \{ animation: demo-wave 1.8s ease-in-out infinite; \}
.demo-orbit \{ animation: demo-orbit 1s linear infinite; \}
.demo-breathe \{ animation: demo-breathe 2.6s ease-in-out infinite; \}
.demo-tabs .sui-tab.active \{ animation: demo-tab-pop .32s cubic-bezier(.2,1.4,.4,1); \}
.all-split \{ height: 170px; \}
.all-media \{ width: 100%; height: 72px; object-fit: cover; border-radius: 8px; \}
.all-video \{ width: 100%; border-radius: 8px; \}
.all-frame \{ width: 100%; height: 140px; border: 1px solid #2a3345; border-radius: 8px; \}
.all-form \{ display: flex; flex-direction: column; gap: 8px; \}
.all-form fieldset \{ border: 1px solid #2a3345; border-radius: 8px; padding: 10px; display: flex; flex-direction: column; gap: 8px; \}
.all-form legend \{ color: #8fa0b8; font-size: 12px; padding: 0 4px; \}
@keyframes demo-jelly \{ 0%, 100% \{ scale: 1 1; \} 30% \{ scale: 1.06 .94; \} 55% \{ scale: .95 1.05; \} 80% \{ scale: 1.02 .98; \} \}
@keyframes demo-ping \{ 0% \{ box-shadow: 0 0 0 0 rgba(74,222,128,.6); \} 70%, 100% \{ box-shadow: 0 0 0 8px rgba(74,222,128,0); \} \}
@keyframes demo-halo \{ 0%, 100% \{ box-shadow: 0 0 0 0 rgba(76,92,247,0); \} 50% \{ box-shadow: 0 0 26px 2px rgba(76,92,247,.4); \} \}
@keyframes demo-wave \{ 0%, 100% \{ background-position: 0 0; \} 50% \{ background-position: 0 100%; \} \}
@keyframes demo-orbit \{ 0% \{ transform: rotate(0deg) translateX(0); \} 50% \{ transform: rotate(180deg) translateX(3px); \} 100% \{ transform: rotate(360deg) translateX(0); \} \}
@keyframes demo-breathe \{ 0%, 100% \{ filter: drop-shadow(0 0 0 transparent); \} 50% \{ filter: drop-shadow(0 0 7px #f59e0b); \} \}
@keyframes demo-tab-pop \{ 0% \{ scale: .95; \} 60% \{ scale: 1.03; \} 100% \{ scale: 1; \} \}
@media (prefers-reduced-motion: reduce) \{ .demo-jelly, .demo-ping, .demo-halo, .demo-breathe, .demo-orbit \{ animation: none !important; \} .sui-progress.demo-wave::-webkit-progress-value, .sui-progress.demo-wave::-moz-progress-bar \{ animation: none !important; \} .demo-tabs .sui-tab.active \{ animation: none !important; \} \}
"

view main {
	title "All controls and animations"
	col class: "all-page" {
		h1 "All controls and animations"
		p "A single page with every SimpleUI control. Feedback cards use the CSS animation library; dropdowns, popovers and toggles use the runtime animation parameter." class: "all-note"

		row {
			badge "Tooltip" tooltip: "The tooltip repositions itself and uses its own animation."
			badge "LIVE" class: "demo-ping"
			badge "New"
			spinner class: "demo-orbit"
		}

		div class: "all-grid" {
			card class: "demo-halo" {
				h3 "Progress - wave"
				progress value: 64 max: 100 class: "demo-wave"
			}
			card {
				h3 "Spinner - orbit"
				row {
					spinner class: "demo-orbit"
					p "Loading" class: "all-note"
				}
			}
			card {
				h3 "Alert - breathe"
				alert "Deploy completed" class: "demo-breathe"
			}
			card {
				h3 "Button - jelly"
				button "Press me" class: "demo-jelly" onClick: countClick
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
				input "Name" value: name
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
				h3 "Toggles with runtime animation"
				checkbox "Scale" checked: on animation: "scale" duration: "280ms" easing: "ease-out"
				checkbox "Fade" checked: checked animation: "fade" duration: "220ms"
				radio "Slide" checked: radio animation: "slide" duration: "240ms"
				switch "Morph" checked: morphOn animation: "morph" duration: "320ms" easing: "ease-in-out"
				p "The checkbox, radio and switch animate their own state change." class: "all-note"
			}
			card {
				h3 "Buttons"
				row {
					button "Primary" onClick: countClick
					button "Secondary" onClick: countClick
				}
				row {
					button "Open modal" onClick: openModal
					button "Toast" onClick: showToast
					button "Drawer" onClick: openFilters
				}
			}
		}

		h2 "Navigation"
		div class: "all-grid" {
			card {
				h3 "Tabs"
				tabs "Sections" class: "demo-tabs" {
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
				h3 "Dropdown - three animations"
				row {
					dropdown "Fade" animation: "fade" duration: "180ms" {
						dropdownitem "Fade item" onClick: chooseAction
					}
					dropdown "Scale" animation: "scale" duration: "200ms" {
						dropdownitem "Scale item" onClick: chooseAction
					}
					dropdown "Slide" animation: "slide" duration: "220ms" {
						dropdownitem "Slide item" onClick: chooseAction
					}
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
				h3 "Combobox - slide"
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
				h3 "Popover - scale"
				popover "Details" animation: "scale" duration: "160ms" {
					p "Anchored popover content." class: "all-note"
				}
			}
			card {
				h3 "Command palette - slide"
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
			a "SimpleUI documentation" href: "https://example.com/docs"
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
