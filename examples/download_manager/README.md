# Download manager

A chunked downloader with a graphical front end. The backend is a small Go
program; the front end is a single SimpleUI file. Together they show a
complete "operation" app: a long running job, split into parts, tracked one by
one, pausable, resumable, restartable part by part and verifiable.

```
examples/download_manager/
  app.sui        the SimpleUI GUI (one file, type-checked)
  server/main.go the Go backend: file server + chunked downloader
  run.ps1        launcher: builds the app and starts the backend
  run.bat        same for cmd.exe
```

The backend lives in its own Go module (`server/go.mod`), so the example does
not add a dependency to the root module.

## What it demonstrates

- **The CLI takes the URL of the file.** `dm.exe --url https://example.com/file.bin`
  is the file to download; the other flags are the configuration (part size,
  workers, throttle, allowed hosts, expected checksum). The GUI reads the
  defaults from `GET /api/config`, shows them in one line at the top and lets
  the operator change them in the settings panel.
- **Two ways to download.** The settings panel (the small ⚙ button) chooses
  between **one single file** (one request, no parts, no reassembly) and
  **split into parts**. In split mode the panel also shows the size of the
  parts and how many parts are downloaded at the same time.
- **A split decided by the size.** The file is cut into parts of a chosen
  length (`256 KiB` … `16 MiB`); the number of parts follows from the file
  size. A 24 MiB file with 1 MiB parts becomes 24 parts.
- **One Range request per part.** Every part is fetched with its own
  `Range: bytes=start-end` request, so the parts are really independent
  requests and not a single stream split in the client.
- **Two levels of progress.** A bar for the whole download and, under it, one
  bar per part with its state, its bytes and its error message. The bars
  start from the **most recent** part and the list has a limited height with a
  scroll bar, so the part that is running now is always visible.
- **The data in another tab.** The tables of the parts and of the jobs live in
  the second tab, so the live view stays readable.
- **One button for the whole lifecycle.** The lightning button next to the URL
  starts the download, becomes the pause button while it runs and the resume
  button when it is paused; its hint says what it is going to do. The cancel
  button (✕) appears next to it only while the download is paused.
- **Pause, resume and cancel.** Pausing cancels the running requests but keeps
  the bytes already on disk. Resuming asks for `bytes=start+received-end`, so a
  part that was half downloaded continues from the byte it stopped at. The
  cancel stops the job.
- **The errors are toasts.** Every failure of the app and of the backend is
  shown as an error toast (and as a line of text), and the same message is not
  repeated twice in a row.
- **A sound at the end.** When the download completes the app plays a short
  chime served by the backend (`GET /sound.wav`) with the new `sound` module.
- **A status bar.** The percentage, the size of the file, the bytes downloaded
  and the estimated time left are always visible at the bottom of the screen.
- **A failed part restarts alone.** The operator selects the failed row in the
  table and presses "Retry part N": only that part is fetched again, its bytes
  are reset and the other parts are not touched. "Retry all failed" does the
  same for every failed part at once.
- **Reassembly and verification.** When every part is complete the backend
  concatenates them in order and compares the size with the total the server
  announced; with `--expect-sha256` it also compares the SHA-256 of the result.
- **Polling with `interval`.** The GUI polls the status endpoint every 500 ms
  and repaints the progress bars.
- **A dynamic URL done safely.** The pause/resume/cancel URL contains the job
  id, so it is a dynamic URL: the app uses `http.postSafe` with a literal
  allow-list, exactly as the security rules require.
- **The backend owns the job state.** Pause, resume, cancel and retry are HTTP
  requests to the server, never local flags: the GUI can be reloaded at any
  moment without losing the download.

## How to run

```bash
# build the app and start the backend (Windows): by default the backend
# downloads its own sample file, so the example works offline
pwsh examples/download_manager/run.ps1

# or a real file, with another part size
pwsh examples/download_manager/run.ps1 -Url "https://example.com/file.bin" -PartSize "2 MiB"

# headless: build and start the backend without opening the browser
pwsh examples/download_manager/run.ps1 -NoOpen

# or by hand
simpleui build examples/download_manager/app.sui --target web --out examples/download_manager/public
cd examples/download_manager/server
go run . --port 8090 --web-dir ../public --files-dir ../files --out-dir ../downloads --url http://127.0.0.1:8090/files/sample.bin
```

Then open <http://127.0.0.1:8090/>. The backend creates a deterministic
24 MiB sample file (`files/sample.bin`) on first start, so the example works
offline and always produces the same bytes.

### The flags of the CLI

| Flag | Default | Meaning |
|---|---|---|
| `--url` | (empty) | URL of the file to download; the GUI shows it and the operator can change it |
| `--part-size` | `1 MiB` | length of one part: the file is split accordingly |
| `--workers` | `4` | parallel part downloads |
| `--throttle` | `0` | bytes/s per worker; slows the download down so the progress is visible |
| `--max-parts` | `256` | upper bound of the parts of one job |
| `--out-dir` | `../downloads` | folder with the parts and the merged file |
| `--web-dir` | (empty) | folder with the compiled SUI app |
| `--files-dir` | `../files` | folder with the files offered for download |
| `--sample-mib` | `24` | size of the generated sample file; `0` disables the generation |
| `--expect-sha256` | (empty) | expected SHA-256 of the file: the merge verifies it |
| `--allow-host` | (empty) | host that may be downloaded; the flag is repeatable |
| `--port` | `8090` | HTTP port |
| `--no-open` | – | do not open the browser |

Only `http` and `https` URLs are accepted and credentials in the URL are
refused. With `--allow-host` only the listed hosts can be downloaded:

```bash
dm.exe --url https://example.com/file.bin --allow-host example.com --allow-host cdn.example.com
```

### The port is part of the app

The GUI calls `http://127.0.0.1:8090/api/...` with **literal** URLs, and the
dynamic job URL is allowed only for the literal origin `http://127.0.0.1:8090`.
Running the backend on a different port therefore means editing `app.sui` too:
change the three URLs and the allow-list, then rebuild the app. The backend
prints the URL it listens on.

## The GUI

1. **Top bar** – the configuration in one line (mode, size of the parts, parts
   at the same time) and the small ⚙ button that opens the settings panel as
   an overlay.
2. **Target** – the URL of the file (prefilled with the `--url` of the CLI) and
   the single action button: ⚡ to start, ❚❚ to pause, ► to resume. Below it
   the hint of the current phase. The ✕ cancel button appears next to it only
   while the download is paused.
3. **Progress tab** – the bar of the whole download and, under it, a
   collapsible panel with one bar per part (the most recent one first), shown
   only when the download runs in split mode. When the download is complete a
   link appears to download the reassembled file.
4. **Data tab** – the table of the parts and the table of the jobs. Clicking a
   row of the parts table selects the part: the "Retry part N" button fetches
   that part again, "Retry all failed" requeues every failed part.
5. **Status bar** – the percentage, the size of the file, the bytes downloaded
   and the estimated time left, always at the bottom of the screen.

## HTTP API

| Endpoint | Purpose |
|---|---|
| `GET /api/config` | the configuration of the CLI (default URL, part size, workers, ...) |
| `POST /api/downloads` | start a job: `{"url":"...","partSize":"2 MiB","workers":4,"single":false}` |
| `GET /api/downloads` | all jobs, most recent first |
| `POST /api/downloads/{id}/pause` | pause: keep the bytes, stop the requests |
| `POST /api/downloads/{id}/resume` | resume: every part continues from its byte |
| `POST /api/downloads/{id}/cancel` | cancel the job |
| `POST /api/downloads/{id}/parts/{n}/retry` | download one failed part again |
| `POST /api/downloads/{id}/retry-failed` | download every failed part again |
| `GET /api/downloads/{id}/file` | the reassembled file |
| `GET /files/{name}` | a local file, with `Range` support |
| `GET /sound.wav` | the short chime played when a download completes |

`workers` is the number of parts downloaded at the same time; `single` switches
the job to the one-piece mode (one request, no parts, no reassembly). In the
one-piece mode the size of the file is unknown until the server answers, so the
part stays open and the request uses `Range: bytes=N-`.

A job has a state of its own: `running`, `paused`, `merging`, `done`,
`canceled` or `error`. A part has a state too: `pending`, `downloading`,
`done` or `error`.

The parts in the JSON response use the column names of the table, so the
backend and the GUI always agree on what a part is:

```json
{"Part": 3, "From": 2097152, "To": 3145727, "Size": 1048576,
 "Received": 1048576, "Percent": 100, "State": "done"}
```

## The complete SUI source

`app.sui`, the whole file:

```sui
import { http, json, sound } from "std"

# Download manager - the front end of the chunked downloader.
#
# The Go backend (server/main.go) takes the URL of a file, splits it into
# parts according to the chosen part size, downloads every part with its own
# Range request, tracks each part and finally reassembles them into one file.
#
# The screen has three zones:
#   - the top bar: the configuration in one row and the button that opens the
#     settings panel (the file in one piece or split into parts, the size of
#     the parts and the number of parts downloaded at the same time);
#   - the target: the URL and a single button whose icon and function follow
#     the phase of the download (start, pause, resume);
#   - the status bar at the bottom: the data of the download.
#
# The parts live in a collapsible panel of the Progress tab, shown only when
# the download runs in split mode, and the bars start from the most recent
# part. The errors are shown as toasts and the completion plays a sound.
#
# The backend listens on port 8090: the URLs below are literals, and the only
# dynamic URLs (the ones that carry the job id) go through http.postSafe with
# a literal allow-list.

state urlText = ""
state partSize = "1 MiB"
state workers = "4"
state single = false
state jobId = ""
state jobState = "idle"
state jobSingle = false
state phase = "idle"
state actionIcon = "⚡"
state actionHint = "Start the download"
state percent = 0
state totalText = "0 B"
state receivedText = "0 B"
state etaText = "--"
state verifyText = ""
state message = "Paste the URL of the file to download and press Start."
state partsView: array[any] = []
state partsTable: array[any] = []
state selectedPart = ""
state failedParts = 0
state partsSignature = ""
state lastState = ""
state lastError = ""

# The configuration of the CLI: the default URL, the part size and the default
# number of parts downloaded at the same time. The app starts with them, the
# operator can change them in the settings panel.
async fn loadConfig() -> null {
	body, err = http.get("http://127.0.0.1:8090/api/config")
	if err != null {
		notifyError("backend unreachable: " + err)
		return
	}
	cfg, jerr = json.parse(body)
	if jerr != null {
		notifyError("invalid response: " + jerr)
		return
	}
	if urlText == "" {
		urlText = str(cfg["url"])
	}
	let cfgSize = str(cfg["partSize"])
	if cfgSize != "" {
		partSize = cfgSize
	}
	let cfgWorkers = str(cfg["workers"])
	if cfgWorkers != "" {
		workers = cfgWorkers
	}
}

# The errors are shown as a toast: the operator reads them without losing the
# place in the page. The same message is not repeated twice in a row.
fn notifyError(text: string) -> null {
	if text == lastError {
		return
	}
	lastError = text
	message = text
	toast.show(text, "error", 5000)
}

fn notifyInfo(text: string) -> null {
	message = text
	toast.show(text, "info", 3000)
}

# Polling: read the jobs and refresh the views.
async fn refresh() -> null {
	body, err = http.get("http://127.0.0.1:8090/api/downloads")
	if err != null {
		notifyError("backend unreachable: " + err)
		return
	}
	data, jerr = json.parse(body)
	if jerr != null {
		notifyError("invalid response: " + jerr)
		return
	}
	let jobs = data["jobs"]
	n, terr = table.display("jobs", jobs)
	if terr != null {
		notifyError("jobs table not updated: " + terr)
	}
	if len(jobs) == 0 {
		return
	}
	updateJob(jobs[0])
}

# Refresh the view with the state of a job.
async fn updateJob(job: any) -> null {
	jobId = str(job["id"])
	jobState = str(job["state"])
	jobSingle = job["single"]
	percent = job["percent"]
	totalText = str(job["totalText"])
	receivedText = str(job["receivedText"])
	etaText = str(job["etaText"])
	verifyText = str(job["verified"])
	failedParts = job["failedParts"]
	let parts = job["parts"]
	# The bars start from the most recent part: what is happening now is at
	# the top of the list. The table keeps the order of the parts.
	partsView = reverse(parts)
	partsTable = parts
	refreshPartsTable(jobId + "|" + jobState + "|" + str(failedParts) + "|" + str(doneParts(parts)))
	syncPhase()
	let previous = lastState
	lastState = jobState
	if jobState != previous {
		if jobState == "done" {
			announceDone(job)
		}
		if jobState == "error" {
			notifyError("the download failed: " + str(job["error"]))
		}
	}
}

# The single button follows the phase of the download: it starts it, pauses
# it or resumes it, and it always says what it is going to do.
fn syncPhase() -> null {
	if jobState == "running" {
		phase = "running"
		actionIcon = "❚❚"
		actionHint = "Pause the download"
		return
	}
	if jobState == "paused" {
		phase = "paused"
		actionIcon = "►"
		actionHint = "Resume the download"
		return
	}
	phase = "idle"
	actionIcon = "⚡"
	actionHint = "Start the download"
}

# How many parts are complete: the table is refreshed when this changes, so
# the row the operator selected is not lost at every poll.
fn doneParts(parts: any) -> int {
	let n = 0
	let i = 0
	while i < len(parts) {
		if str(parts[i]["State"]) == "done" {
			n = n + 1
		}
		i = i + 1
	}
	return n
}

# The parts table is filled only when its content really changes: a table
# refresh drops the selected row.
fn refreshPartsTable(signature: string) -> null {
	if signature == partsSignature {
		return
	}
	partsSignature = signature
	n, terr = table.display("parts", partsTable)
	if terr != null {
		notifyError("parts table not updated: " + terr)
	}
}

# The download is complete: a toast and a short sound.
async fn announceDone(job: any) -> null {
	toast.show("Download complete: " + str(job["file"]), "success", 4000)
	ok, err = sound.play("http://127.0.0.1:8090/sound.wav")
	if err != null {
		message = "download complete (the browser blocked the sound)"
	}
}

# The one button: it starts, pauses or resumes the download.
async fn onAction() -> null {
	if phase == "idle" {
		startDownload()
		return
	}
	if phase == "running" {
		runAction("pause")
		return
	}
	runAction("resume")
}

# Start a download: the backend splits the file according to the part size.
async fn startDownload() -> null {
	if urlText == "" {
		notifyError("type the URL of the file to download")
		return
	}
	payload, jerr = json.stringify({"url": urlText, "partSize": partSize, "workers": workers, "single": single})
	if jerr != null {
		notifyError("invalid request: " + jerr)
		return
	}
	body, err = http.post("http://127.0.0.1:8090/api/downloads", payload)
	if err != null {
		notifyError("could not start: " + err)
		return
	}
	job, perr = json.parse(body)
	if perr != null {
		notifyError("invalid response: " + perr)
		return
	}
	selectedPart = ""
	updateJob(job)
	if single == true {
		message = "download started: " + str(job["file"]) + " - one single file"
	} else {
		message = "download started: " + str(job["file"]) + " - " + str(job["partCount"]) + " parts of " + str(job["partSizeText"]) + ", " + str(job["workers"]) + " at a time"
	}
}

# The URL contains the job id, so it is dynamic: the *Safe variant with a
# literal allow-list is the one the security rules require.
async fn runAction(act: string) -> null {
	if jobId == "" {
		notifyError("no download to manage")
		return
	}
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/" + act, "", "http://127.0.0.1:8090")
	if err != null {
		notifyError("operation " + act + " failed: " + err)
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		notifyError("invalid response: " + jerr)
		return
	}
	updateJob(job)
	notifyInfo("operation " + act + " done")
}

# Cancel: the download is interrupted. The button is there only while the
# download is paused.
async fn cancelDownload() -> null {
	runAction("cancel")
}

# The settings panel: the icon button opens it as an overlay.
fn openSettings() -> null {
	modal.open("settings")
}

fn closeSettings() -> null {
	modal.close("settings")
}

# The operator selects a row of the parts table: the part number is kept in a
# state, so the retry button always knows which part to fetch again.
fn onPartSelected() -> null {
	sel, e = table.selectedRows("parts")
	if e != null {
		notifyError("cannot read the selection: " + e)
		return
	}
	if len(sel) == 0 {
		return
	}
	rows, e2 = table.data("parts")
	if e2 != null {
		notifyError("cannot read the parts table: " + e2)
		return
	}
	let idx = sel[0]
	if idx < 0 or idx >= len(rows) {
		return
	}
	selectedPart = rows[idx][0]
	message = "Part " + selectedPart + " selected: press Retry to download it again."
}

# Retry one part: only that part is fetched again, the others are untouched.
async fn retrySelectedPart() -> null {
	if jobId == "" {
		notifyError("no download to manage")
		return
	}
	if selectedPart == "" {
		notifyError("select a row in the Data tab first")
		return
	}
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/parts/" + selectedPart + "/retry", "", "http://127.0.0.1:8090")
	if err != null {
		notifyError("retry failed: " + err)
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		notifyError("invalid response: " + jerr)
		return
	}
	updateJob(job)
	notifyInfo("part " + selectedPart + " queued again")
}

# Retry every part that failed.
async fn retryFailedParts() -> null {
	if jobId == "" {
		notifyError("no download to manage")
		return
	}
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/retry-failed", "", "http://127.0.0.1:8090")
	if err != null {
		notifyError("retry failed: " + err)
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		notifyError("invalid response: " + jerr)
		return
	}
	updateJob(job)
	notifyInfo("the failed parts are queued again")
}

async fn boot() -> null {
	interval(500, refresh)
	loadConfig()
}

# Fill the parts table again on demand.
fn refreshData() -> null {
	partsSignature = ""
	refreshPartsTable("manual")
}

view main {
	title "Download manager"
	onLoad: boot

	toast class: "toast-host" hAlign: "right" vAlign: "top" hOffset: "20px" vOffset: "20px"

	modal "Download settings" name: "settings" {
		div class: "settings" {
			row class: "setting" {
				switch "Download the file in one piece" checked: single class: "mode-switch"
				col class: "setting-text" {
					p "On: one request for the whole file, no parts and no reassembly." class: "hint"
					p "Off: the file is split into parts downloaded in parallel." class: "hint"
				}
			}
			if single == false {
				row class: "setting" {
					combobox "Part size" options: "256 KiB,512 KiB,1 MiB,2 MiB,4 MiB,8 MiB,16 MiB" value: partSize class: "setting-control"
					col class: "setting-text" {
						p "The file is split into parts of this size." class: "hint"
					}
				}
				row class: "setting" {
					number "Parts at the same time" value: workers min: "1" max: "32" step: "1" class: "setting-control"
					col class: "setting-text" {
						p "How many parts are downloaded at the same time." class: "hint"
					}
				}
			}
			if single == true {
				p "Now: the file is downloaded in one piece." class: "hint"
			} else {
				p "Now: parts of {partSize}, {workers} at a time." class: "hint"
			}
		}
		modalbutton "Close"
		modalbutton "Apply" variant: "primary" onClick: closeSettings
	}

	card class: "topbar" {
		row class: "topbar-row" {
			col class: "topbar-info" {
				if single == true {
					p "one single file - the whole file in one request" class: "config-line"
				} else {
					p "split in parts of {partSize} - {workers} at a time" class: "config-line"
				}
			}
			button "⚙" onClick: openSettings class: "btn-icon btn-icon-sm" tooltip: "Download settings"
		}
	}

	card class: "panel" {
		h3 "Target"
		row class: "field" {
			input "File URL" value: urlText placeholder: "https://example.com/file.bin" class: "url-input"
			button "{actionIcon}" onClick: onAction class: "btn-icon btn-action"
			if phase == "paused" {
				button "✕" onClick: cancelDownload class: "btn-icon btn-cancel" tooltip: "Cancel the download"
			}
		}
		p "{actionHint}" class: "action-hint"
		row class: "status-row" {
			badge "{jobState}" class: "state-badge"
			p "{message}" class: "status"
		}
	}

	tabs "Download" {
		tab "Progress" {
			card class: "panel" {
				h3 "Total progress"
				progress value: percent max: 100 class: "bar-total"
				p "{verifyText}" class: "verify"
				if jobState == "done" {
					a "Download the reassembled file" href: "http://127.0.0.1:8090/api/downloads/" + jobId + "/file" class: "dl-link"
				}
			}

			if jobSingle == false {
				accordion "Parts" {
					accordionitem "{len(partsView)} parts - the most recent one first" {
						if len(partsView) == 0 {
							p "No download yet: the bars of the parts appear here." class: "muted"
						} else {
							div class: "parts-scroll" {
								list part in partsView {
									row class: "part-row" {
										p "Part {part["Part"]}" class: "part-index"
										progress value: part["Percent"] max: 100 class: "bar-part"
										p "{part["State"]}" class: "part-state"
										p "{part["Received"]} / {part["Size"]}" class: "part-bytes"
										if part["State"] == "error" {
											p "{part["Error"]}" class: "part-error"
										}
									}
								}
							}
						}
					}
				}
			}
		}

		tab "Data" {
			card class: "panel" {
				h3 "Parts"
				row class: "controls" {
					button "Retry part {selectedPart}" onClick: retrySelectedPart class: "btn-warn"
					button "Retry all failed" onClick: retryFailedParts class: "btn-warn"
					button "Refresh data" onClick: refreshData class: "btn-ghost"
					p "Click a row to choose the part to download again." class: "muted"
				}
				table name: "parts" selection: "row" onRowSelect: onPartSelected {
					thead {
						trow {
							tcell "Part" head: true
							tcell "From" head: true
							tcell "To" head: true
							tcell "Size" head: true
							tcell "Received" head: true
							tcell "Percent" head: true
							tcell "State" head: true
						}
					}
					tbody {
						trow {
							tcell "-"
							tcell "-"
							tcell "-"
							tcell "-"
							tcell "-"
							tcell "-"
							tcell "-"
						}
					}
				}
			}

			card class: "panel" {
				h3 "Jobs"
				table name: "jobs" {
					thead {
						trow {
							tcell "Id" head: true
							tcell "File" head: true
							tcell "State" head: true
							tcell "Percent" head: true
							tcell "Total" head: true
						}
					}
					tbody {
						trow {
							tcell "-"
							tcell "-"
							tcell "-"
							tcell "-"
							tcell "-"
						}
					}
				}
			}
		}
	}

	div class: "statusbar" {
		row class: "statusbar-row" {
			col class: "sb-item" {
				p "{percent}%" class: "sb-value"
				p "progress" class: "sb-label"
			}
			col class: "sb-item" {
				p "{totalText}" class: "sb-value"
				p "size" class: "sb-label"
			}
			col class: "sb-item" {
				p "{receivedText}" class: "sb-value"
				p "downloaded" class: "sb-label"
			}
			col class: "sb-item" {
				p "{etaText}" class: "sb-value"
				p "time left" class: "sb-label"
			}
		}
	}
}

css "body \{ margin: 0; padding: 0 0 120px; background: radial-gradient(1100px 520px at 12% -8%, rgba(99, 102, 241, 0.22) 0%, rgba(15, 23, 42, 0) 60%), linear-gradient(165deg, #0b1020 0%, #131a30 55%, #0e1426 100%); color: #e6e9f5; font-family: system-ui, sans-serif; } h1 \{ font-size: 26px; margin: 0 0 6px; letter-spacing: -0.02em; } h3 \{ margin: 0 0 14px; font-size: 15px; text-transform: uppercase; letter-spacing: 0.08em; color: #a5b4fc; } p \{ line-height: 1.55; } card \{ background: rgba(255, 255, 255, 0.04); border: 1px solid rgba(148, 163, 184, 0.16); border-radius: 18px; padding: 20px 22px; margin: 16px 0; box-shadow: 0 18px 40px rgba(2, 6, 23, 0.45); } .topbar \{ background: rgba(15, 23, 42, 0.55); border: 1px solid rgba(148, 163, 184, 0.14); border-radius: 14px; padding: 10px 16px; margin: 18px 0 6px; box-shadow: none; } .topbar-row \{ align-items: center; gap: 14px; justify-content: space-between; } .topbar-info \{ flex: 1 1 auto; gap: 2px; } .config-line \{ margin: 0; font-size: 13px; color: #a8b1cc; letter-spacing: 0.02em; } .field \{ margin: 4px 0 10px; gap: 12px; align-items: center; } .url-input \{ flex: 1 1 auto; } .action-hint \{ margin: 0 0 12px; font-size: 12px; color: #7c86a8; text-transform: uppercase; letter-spacing: 0.08em; } .sui-button.btn-icon \{ background: rgba(255, 255, 255, 0.06); color: #c7d2fe; border: 1px solid rgba(129, 140, 248, 0.35); width: 42px; height: 42px; padding: 0; border-radius: 12px; font-size: 18px; line-height: 1; display: inline-flex; align-items: center; justify-content: center; transition: transform 140ms ease, box-shadow 140ms ease, color 140ms ease, background 140ms ease; } .sui-button.btn-icon:hover \{ background: rgba(99, 102, 241, 0.22); color: #e0e7ff; transform: translateY(-1px); box-shadow: 0 0 0 4px rgba(129, 140, 248, 0.16), 0 10px 22px rgba(99, 102, 241, 0.35); } .sui-button.btn-action \{ background: linear-gradient(135deg, #6366f1 0%, #22d3ee 100%); color: #061021; border: 0; font-size: 19px; font-weight: 700; box-shadow: 0 10px 24px rgba(99, 102, 241, 0.35); } .sui-button.btn-action:hover \{ color: #061021; background: linear-gradient(135deg, #818cf8 0%, #67e8f9 100%); box-shadow: 0 0 0 5px rgba(250, 204, 21, 0.22), 0 14px 30px rgba(99, 102, 241, 0.45); } .sui-button.btn-cancel \{ background: rgba(244, 63, 94, 0.14); color: #fda4af; border: 1px solid rgba(244, 63, 94, 0.38); } .sui-button.btn-cancel:hover \{ background: rgba(244, 63, 94, 0.26); color: #ffe4e6; box-shadow: 0 0 0 4px rgba(244, 63, 94, 0.18); } .sui-button.btn-icon-sm \{ width: 32px; height: 32px; font-size: 15px; border-radius: 10px; } .status-row \{ gap: 12px; align-items: center; margin-top: 4px; } .status \{ color: #a8b1cc; font-size: 13px; margin: 0; } .state-badge \{ background: rgba(148, 163, 184, 0.16); color: #cbd5e1; border: 1px solid rgba(148, 163, 184, 0.30); } .verify \{ margin: 14px 0 0; color: #6ee7b7; font-size: 13px; } .dl-link \{ display: inline-block; margin-top: 12px; color: #67e8f9; font-weight: 600; } .parts-scroll \{ max-height: 320px; overflow-y: auto; overflow-x: hidden; padding: 2px 8px 2px 2px; margin: 4px 0 10px; scrollbar-width: thin; scrollbar-color: rgba(129, 140, 248, 0.45) rgba(15, 23, 42, 0.4); } .parts-scroll::-webkit-scrollbar \{ width: 10px; } .parts-scroll::-webkit-scrollbar-track \{ background: rgba(15, 23, 42, 0.45); border-radius: 8px; } .parts-scroll::-webkit-scrollbar-thumb \{ background: rgba(129, 140, 248, 0.45); border-radius: 8px; border: 2px solid transparent; background-clip: padding-box; } .row.part-row \{ display: grid; grid-template-columns: 64px minmax(120px, 1fr) 92px 148px; align-items: center; gap: 14px; padding: 8px 10px; border-radius: 10px; } .row.part-row:hover \{ background: rgba(148, 163, 184, 0.08); } .part-index \{ font-weight: 600; color: #c7d2fe; margin: 0; font-variant-numeric: tabular-nums; } .part-state \{ font-size: 12px; color: #8b93b0; margin: 0; } .part-bytes \{ font-size: 12px; color: #8b93b0; text-align: right; margin: 0; font-variant-numeric: tabular-nums; } .part-error \{ grid-column: 1 / -1; font-size: 12px; color: #fda4af; margin: 0; } progress.bar-total \{ --sui-progress-height: 12px; --sui-progress-radius: 999px; --sui-progress-track: rgba(148, 163, 184, 0.16); --sui-progress-value: #6366f1; --sui-progress-value-end: #22d3ee; } progress.bar-part \{ --sui-progress-height: 8px; --sui-progress-radius: 999px; --sui-progress-track: rgba(148, 163, 184, 0.16); --sui-progress-value: #818cf8; --sui-progress-value-end: #38bdf8; } .settings \{ display: flex; flex-direction: column; gap: 14px; min-width: 320px; } .setting \{ gap: 16px; align-items: flex-start; padding-bottom: 12px; border-bottom: 1px solid rgba(148, 163, 184, 0.14); } .setting:last-child \{ border-bottom: 0; padding-bottom: 0; } .setting-text \{ gap: 2px; flex: 1 1 auto; } .setting-text p \{ margin: 0; font-size: 12px; color: #8b93b0; line-height: 1.45; } .setting-control \{ flex: 0 0 220px; } .mode-switch \{ flex: 0 0 auto; } .hint \{ color: #8b93b0; font-size: 13px; margin: 2px 0 10px; } .muted \{ color: #8b93b0; font-size: 13px; } .controls \{ gap: 10px; flex-wrap: wrap; margin-bottom: 14px; } .sui-button.btn-warn \{ background: rgba(251, 191, 36, 0.16); color: #fcd34d; border: 1px solid rgba(251, 191, 36, 0.42); padding: 9px 16px; border-radius: 12px; } .sui-button.btn-ghost \{ background: transparent; color: #94a3b8; border: 1px solid rgba(148, 163, 184, 0.28); padding: 9px 16px; border-radius: 12px; } .toast-host .sui-toast \{ width: 340px; } .sui-tabs \{ --sui-tabs-border: rgba(148, 163, 184, 0.16); --sui-tabs-radius: 16px; --sui-tabs-bg: rgba(255, 255, 255, 0.02); --sui-tablist-bg: rgba(15, 23, 42, 0.55); --sui-tab-padding: 10px 18px; --sui-tab-active-bg: rgba(99, 102, 241, 0.18); --sui-tab-accent: #22d3ee; --sui-tab-fg: #94a3b8; --sui-tab-active-fg: #e0e7ff; } .sui-tabpanels \{ padding: 18px 20px; } .statusbar \{ position: fixed; left: 0; right: 0; bottom: 0; z-index: 5; margin: 0; padding: 12px 22px; background: rgba(11, 16, 32, 0.94); border-top: 1px solid rgba(148, 163, 184, 0.18); backdrop-filter: blur(8px); } .statusbar-row \{ gap: 18px; flex-wrap: wrap; } .sb-item \{ flex: 1 1 120px; gap: 2px; } .sb-value \{ margin: 0; font-size: 17px; font-weight: 700; color: #e0e7ff; font-variant-numeric: tabular-nums; } .sb-label \{ margin: 0; font-size: 11px; text-transform: uppercase; letter-spacing: 0.08em; color: #7c86a8; } @media (max-width: 760px) { .row.part-row \{ grid-template-columns: 56px 1fr; row-gap: 6px; } .parts-scroll \{ max-height: 240px; } .setting \{ flex-direction: column; gap: 8px; } .setting-control \{ flex: 1 1 100%; width: 100%; } }"
```

### How to edit it

1. Change `app.sui` (the designer can open it: `simpleui gui`).
2. Type-check it: `simpleui check app.sui --json`.
3. Rebuild it: `simpleui build app.sui --target web --out public`.
4. Restart the backend (`run.ps1` does both).

## Notes

- The sample file is generated from a deterministic PRNG: the same bytes, the
  same SHA-256, on every machine.
- The backend downloads the URL the operator asked for. `--allow-host` is the
  way to restrict it to known hosts; without it every `http`/`https` URL is
  accepted, so the example is not a client to point at an internal service.
- The parts stay in `--out-dir` after the job ends. They are the evidence of
  how the file was split, and deleting them is a deliberate act.
- A job is not persisted between restarts of the backend: the example keeps
  the job state in memory and the parts on disk.
- The parts table is refilled only when its content really changes:
  `table.display` drops the selection, so a refresh at every poll would lose
  the row the operator selected for the retry.
- The `value:` of an `input` is write-only: the values coming from the CLI are
  read from the state and shown in another element of the view.
- If the server does not support `Range`, the file is downloaded as a single
  part: the split follows what the server can do, not what the operator asked.
- In the one-piece mode the total size is unknown until the server answers, so
  the part stays open and the request uses `Range: bytes=N-`: the backend
  accepts a plain `200` answer too.
- The completion sound comes from `GET /sound.wav`, a chime generated by the
  backend. The browser can refuse to play it (autoplay policy): the app shows
  the error in the status line, the download is still complete.
- The toast host is at the top right of the window: the errors of the app and
  of the backend are shown there and in the status line, never only in the
  console.

## Verified behaviour

Measured against this example (backend and GUI, not just the documentation):

| Scenario | Result |
|---|---|
| CLI configuration | the URL and the part size of the CLI appear in the GUI and the operator can change them |
| 1 MiB file, 512 KiB parts | 2 parts, each with its own Range request |
| Pause | state `paused`, the running parts keep their bytes, nothing advances |
| Resume | the parts continue from the byte they stopped at, the job completes |
| A failing part | the part goes to `error` with the message of the server, the other parts are not affected |
| "Retry part 2" | only that part is fetched again, the job reaches `done` |
| Reassembly | the size of the merged file matches the total announced by the server, the SHA-256 is shown as evidence |
| Settings panel | the ⚙ button opens the overlay; in one-piece mode the part size and the parallelism disappear |
| One-piece mode | the job has a single part and the parts panel is not shown |
| Action button | ⚡ (start) → ❚❚ (pause) → ► (resume), with the ✕ cancel button only while paused |
| Errors | the failures are shown as error toasts, the same message is not repeated |
| Completion | the toast "Download complete" is shown and the chime of `/sound.wav` is played |

The automated test in `internal/core/download_manager_test.go` drives the same
cycle headlessly: it builds the backend, splits a 1 MiB file into two 512 KiB
parts, pauses, checks that no byte advances, resumes and compares the
reassembled file with the original one. A second test uses an origin that
fails one range once and checks that the retry of that single part brings the
job to `done`.
