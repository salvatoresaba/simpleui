# All prefabricated controls

A single gallery app that uses every Block 2 and Block 3 control together:
dropdown, popover, drawer, splitter, dropzone, timeline, empty state, skeleton,
carousel, stepper and command palette.

The states and handlers are wired so the controls can be exercised interactively:
the dropdown records the last action, the drawer opens from a button, the
dropzones fill string states, the carousel and stepper bind integer states, the
empty state and skeleton react to counters/booleans, and the command palette
opens with Ctrl+K or a button.

## Complete source

```sui
# All prefabricated controls in one app.
#
# This example is the reference gallery for the Block 2 and Block 3 controls:
# dropdown, popover, drawer, splitter, dropzone, timeline, empty state,
# skeleton, carousel, stepper and command palette.
#
# Every control is wired to a state or a handler so the runtime behavior can be
# verified end to end.

state action = "none"
state popoverClicks = 0
state showFilters = false
state files = ""
state folder = ""
state slide = 0
state step = 1
state showPalette = false
state lastCommand = "none"
state loading = true
state projects = 0

fn editAction() -> null {
	action = "modifica"
}

fn duplicateAction() -> null {
	action = "duplica"
}

fn deleteAction() -> null {
	action = "elimina"
}

fn bumpPopover() -> null {
	popoverClicks = popoverClicks + 1
}

fn openDrawer() -> null {
	showFilters = true
}

fn closeDrawer() -> null {
	showFilters = false
}

fn addProject() -> null {
	projects = projects + 1
}

fn toggleLoading() -> null {
	if loading {
		loading = false
	} else {
		loading = true
	}
}

fn previousStep() -> null {
	if step > 0 {
		step = step - 1
	}
}

fn nextStep() -> null {
	if step < 2 {
		step = step + 1
	}
}

fn openPalette() -> null {
	showPalette = true
}

fn newProjectCommand() -> null {
	lastCommand = "nuovo progetto"
}

fn settingsCommand() -> null {
	lastCommand = "impostazioni"
}

fn helpCommand() -> null {
	lastCommand = "aiuto"
}

css ".prefab-page \\{ max-width: 980px; margin: 0 auto; padding: 26px; display: flex; flex-direction: column; gap: 16px; \\} .prefab-grid \\{ display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 16px; align-items: start; \\} .prefab-note \\{ color: #8fa0b8; font-size: 13px; \\} .prefab-split \\{ height: 220px; \\} .prefab-carousel \\{ min-height: 170px; \\} .prefab-slide-title \\{ margin: 0 0 6px; \\}"

view main {
	title "Prefabricated controls"
	col class: "prefab-page" {
		h1 "Prefabricated controls"
		p "Dropdown, popover, drawer, splitter, dropzone, timeline, empty state, skeleton, carousel, stepper and command palette together." class: "prefab-note"

		div class: "prefab-grid" {
			card {
				h3 "Dropdown"
				dropdown "Azioni" position: "right" {
					dropdownitem "Modifica" onClick: editAction
					dropdownitem "Duplica" onClick: duplicateAction
					dropdownitem separator: true
					dropdownitem "Elimina" onClick: deleteAction
				}
				p "Ultima azione: {action}" class: "prefab-note"
			}

			card {
				h3 "Popover"
				popover "Dettagli" position: "bottom" {
					h3 "Pannello ancorato"
					p "Il popover può contenere testi, card e pulsanti."
					p "Click interni: {popoverClicks}" class: "prefab-note"
					button "Azione interna" onClick: bumpPopover
				}
				p "Click fuori o Esc per chiudere." class: "prefab-note"
			}

			card {
				h3 "Drawer"
				button "Apri filtri" onClick: openDrawer
				p "Aperto: {showFilters}" class: "prefab-note"
				drawer "Filtri" side: "right" width: "320px" open: showFilters {
					h3 "Filtri"
					p "Scegli stato e intervallo di date."
					button "Applica" onClick: closeDrawer
				}
			}

			card {
				h3 "Split panes"
				splitter direction: "horizontal" initial: "40%" min: "100px" class: "prefab-split" {
					pane {
						h3 "Lista"
						p "Trascina la barra o usa le frecce."
					}
					pane {
						h3 "Dettaglio"
						p "Il secondo pannello occupa lo spazio restante."
					}
				}
			}

			card {
				h3 "Dropzone file"
				dropzone "File CSV" dropFiles: files accept: ".csv,text/csv" multiple: true hint: "Trascina i file o clicca" {
					p "I percorsi finiscono nello stato files." class: "prefab-note"
				}
				p "File: {files}" class: "prefab-note"
			}

			card {
				h3 "Dropzone cartella"
				dropzone "Cartella" dropFiles: folder directory: true pickMode: "auto" hint: "Trascina una cartella o clicca" {
					p "La cartella finisce nello stato folder." class: "prefab-note"
				}
				p "Cartella: {folder}" class: "prefab-note"
			}
		}

		card {
			h3 "Timeline"
			timeline {
				timelineitem "Progetto creato" date: "10 nov, 09:12" status: "done"
				timelineitem "Build completata" date: "10 nov, 09:30" status: "done" {
					p "Gli artefatti sono pronti per la revisione."
				}
				timelineitem "In revisione" date: "11 nov, 14:02" status: "active" {
					p "In attesa di approvazione."
				}
				timelineitem "Deploy" date: "12 nov, 08:00" status: "pending"
			}
		}

		div class: "prefab-grid" {
			card {
				h3 "Empty state"
				if projects == 0 {
					emptystate "Nessun progetto" icon: "▢" hint: "Crea il primo progetto per iniziare." {
						button "Crea progetto" onClick: addProject
					}
				} else {
					p "Progetti creati: {projects}" class: "prefab-note"
					button "Aggiungi progetto" onClick: addProject
				}
			}

			card {
				h3 "Skeleton"
				if loading {
					skeleton lines: 2 width: "45%" height: "16px"
					skeleton lines: 4 width: "100%" height: "10px"
					row {
						skeleton circle: true height: "48px"
						skeleton lines: 3 width: "70%" height: "10px"
					}
				} else {
					p "Contenuto caricato." class: "prefab-note"
				}
				button "Toggle caricamento" onClick: toggleLoading
			}
		}

		card {
			h3 "Carousel"
			carousel value: slide height: "170px" class: "prefab-carousel" {
				slide label: "Uno" {
					h3 "Slide 1" class: "prefab-slide-title"
					p "Frecce, dots e tasti sinistra/destra cambiano slide."
				}
				slide label: "Due" {
					h3 "Slide 2" class: "prefab-slide-title"
					p "La slide attiva è legata allo stato slide."
				}
				slide label: "Tre" {
					h3 "Slide 3" class: "prefab-slide-title"
					p "Il carousel può anche avanzare da solo con autoplay."
				}
			}
			p "Slide {slide + 1} di 3" class: "prefab-note"
		}

		card {
			h3 "Stepper"
			stepper value: step {
				step "Dettagli" description: "Nome e proprietario"
				step "Pagamento" description: "Metodo"
				step "Conferma" description: "Rivedi e invia"
			}
			row {
				button "Indietro" onClick: previousStep
				button "Avanti" onClick: nextStep
			}
			p "Passo {step + 1}" class: "prefab-note"
		}

		card {
			h3 "Command palette"
			p "Premi Ctrl+K o clicca il pulsante. Filtra, usa le frecce e Invio; Esc chiude." class: "prefab-note"
			row {
				button "Apri comandi" onClick: openPalette
				p "Ultimo comando: {lastCommand}" class: "prefab-note"
			}
			commandbar "Cerca comandi" open: showPalette hotkey: "k" {
				command "Nuovo progetto" shortcut: "Ctrl N" onClick: newProjectCommand
				command "Impostazioni" shortcut: "Ctrl ," onClick: settingsCommand
				command "Aiuto" shortcut: "F1" onClick: helpCommand
			}
		}
	}
}
```

## What it demonstrates

| Control | State / handler | Try it |
|---|---|---|
| dropdown | ction, ditAction, duplicateAction, deleteAction | Open **Azioni**, choose an item and read the last action. |
| popover | popoverClicks, umpPopover | Open **Dettagli**, click the internal button, close with Esc or an outside click. |
| drawer | showFilters, openDrawer, closeDrawer | Open **Apri filtri**; close with the backdrop, **Applica** or Esc. |
| splitter | none | Drag the bar or focus it and use the arrow keys, Home/End. |
| dropzone file | iles | Drop CSV files or click the zone; paths appear below. |
| dropzone folder | older, pickMode: "auto" | Drop a folder or use the picker. |
| 	imeline | none | Read the four dated events and their status colors. |
| mptystate | projects, ddProject | Click **Crea progetto**; the empty state becomes a counter. |
| skeleton | loading, 	oggleLoading | Toggle the loading placeholder off and on. |
| carousel | slide | Use the arrows, dots, or Left/Right keys; the counter follows. |
| stepper | step, previousStep, 
extStep | Click a step or use **Indietro**/**Avanti**. |
| commandbar | showPalette, lastCommand, three handlers | Press Ctrl+K or **Apri comandi**, filter, press Enter, close with Esc. |

All dynamic values are escaped by the runtime; file names and paths are data,
never markup. The dropzones reuse the validated drag-and-drop/picker runtime.

## Designer

Every control is available from the palette:

- Block 2 under **Controls**: Dropdown menu, Popover, Drawer, Split panes, File dropzone.
- Block 3 under **Content**: Timeline, Empty state, Skeleton loader, Carousel, Stepper.
- Command palette under **Overlay**.

Open the example in the designer to inspect the dedicated sections and the
standard style/binding panels, then edit the source or the element tree.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

The un.ps1 / un.bat scripts rebuild the desktop demo and launch it.
