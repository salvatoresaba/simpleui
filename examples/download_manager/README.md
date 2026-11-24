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
  defaults from `GET /api/config`, shows them and lets the operator change the
  URL and the part size before starting.
- **A split decided by the size.** The file is cut into parts of a chosen
  length (`256 KiB` … `16 MiB`); the number of parts follows from the file
  size. A 24 MiB file with 1 MiB parts becomes 24 parts.
- **One Range request per part.** Every part is fetched with its own
  `Range: bytes=start-end` request, so the parts are really independent
  requests and not a single stream split in the client.
- **Two levels of progress.** A bar for the whole download and, under it, one
  bar per part with its state, its bytes and its error message.
- **The data in another tab.** The tables of the parts and of the jobs live in
  the second tab, so the live view stays readable.
- **Pause, resume and cancel.** Pausing cancels the running requests but keeps
  the bytes already on disk. Resuming asks for `bytes=start+received-end`, so a
  part that was half downloaded continues from the byte it stopped at. The
  cancel stops the job.
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

1. **Target** – the URL of the file (prefilled with the `--url` of the CLI),
   the part size and the Start / Pause / Resume / Cancel buttons.
2. **Progress tab** – the bar of the whole download with the bytes, the speed
   and the remaining time; under it, one bar per part with its state, its
   bytes and its error message. When the download is complete a link appears
   to download the reassembled file.
3. **Data tab** – the table of the parts and the table of the jobs. Clicking a
   row of the parts table selects the part: the "Retry part N" button fetches
   that part again, "Retry all failed" requeues every failed part.

## HTTP API

| Endpoint | Purpose |
|---|---|
| `GET /api/config` | the configuration of the CLI (default URL, part size, ...) |
| `POST /api/downloads` | start a job: `{"url":"...","partSize":"2 MiB"}` |
| `GET /api/downloads` | all jobs, most recent first |
| `POST /api/downloads/{id}/pause` | pause: keep the bytes, stop the requests |
| `POST /api/downloads/{id}/resume` | resume: every part continues from its byte |
| `POST /api/downloads/{id}/cancel` | cancel the job |
| `POST /api/downloads/{id}/parts/{n}/retry` | download one failed part again |
| `POST /api/downloads/{id}/retry-failed` | download every failed part again |
| `GET /api/downloads/{id}/file` | the reassembled file |
| `GET /files/{name}` | a local file, with `Range` support |

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
import { http, json } from "std"

# Download manager - the front end of the chunked downloader.
#
# The Go backend (server/main.go) takes the URL of a file, splits it into
# parts according to the chosen part size, downloads every part with its own
# Range request, tracks each part and finally reassembles them into one file.
#
# The app follows the progress, shows a bar for the whole download and one bar
# for every part, and lets the operator pause, resume, cancel or retry a
# single failed part. The tables live in a second tab.
#
# The backend listens on port 8090: the URLs below are literals, and the only
# dynamic URLs (the ones that carry the job id) go through http.postSafe with
# a literal allow-list.

state urlText = ""
state partSize = "1 MiB"
state jobId = ""
state jobState = "idle"
state percent = 0
state progressText = "waiting"
state speedText = "--"
state etaText = "--"
state verifyText = ""
state message = "Paste the URL of the file to download and press Start."
state partsView: array[any] = []
state selectedPart = ""
state failedParts = 0
state partsSignature = ""

# The configuration of the CLI: the default URL and the default part size. The
# app starts with them, the operator can change them.
async fn loadConfig() -> null {
	body, err = http.get("http://127.0.0.1:8090/api/config")
	if err != null {
		message = "backend unreachable: " + err
		return
	}
	cfg, jerr = json.parse(body)
	if jerr != null {
		message = "invalid response: " + jerr
		return
	}
	if urlText == "" {
		urlText = str(cfg["url"])
	}
	# The part size of the CLI is the default: the operator can change it
	# afterwards, the configuration is read only once.
	let cfgSize = str(cfg["partSize"])
	if cfgSize != "" {
		partSize = cfgSize
	}
}

# Polling: read the jobs and refresh the views.
async fn refresh() -> null {
	body, err = http.get("http://127.0.0.1:8090/api/downloads")
	if err != null {
		message = "backend unreachable: " + err
		return
	}
	data, jerr = json.parse(body)
	if jerr != null {
		message = "invalid response: " + jerr
		return
	}
	let jobs = data["jobs"]
	n, terr = table.display("jobs", jobs)
	if terr != null {
		message = "jobs table not updated: " + terr
	}
	if len(jobs) == 0 {
		return
	}
	updateJob(jobs[0])
}

# Refresh the view with the state of a job.
fn updateJob(job: any) -> null {
	jobId = str(job["id"])
	jobState = str(job["state"])
	percent = job["percent"]
	progressText = str(job["percent"]) + "% - " + str(job["receivedText"]) + " of " + str(job["totalText"])
	speedText = str(job["speedText"])
	etaText = str(job["etaText"])
	verifyText = str(job["verified"])
	failedParts = job["failedParts"]
	partsView = job["parts"]
	refreshPartsTable(jobId + "|" + jobState + "|" + str(failedParts) + "|" + str(doneParts(job["parts"])))
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
	n, terr = table.display("parts", partsView)
	if terr != null {
		message = "parts table not updated: " + terr
	}
}

# Start a download: the backend splits the file according to the part size.
async fn startDownload() -> null {
	if urlText == "" {
		message = "type the URL of the file to download"
		return
	}
	payload, jerr = json.stringify({"url": urlText, "partSize": partSize})
	if jerr != null {
		message = "invalid request: " + jerr
		return
	}
	body, err = http.post("http://127.0.0.1:8090/api/downloads", payload)
	if err != null {
		message = "could not start: " + err
		return
	}
	job, perr = json.parse(body)
	if perr != null {
		message = "invalid response: " + perr
		return
	}
	selectedPart = ""
	updateJob(job)
	message = "download started: " + str(job["file"]) + " - " + str(job["partCount"]) + " parts of " + str(job["partSizeText"])
}

# Pause: the parts already downloaded stay, the download restarts from there.
async fn pauseDownload() -> null {
	runAction("pause")
}

# Resume: the backend continues every part from the byte it stopped at.
async fn resumeDownload() -> null {
	runAction("resume")
}

# Cancel: the download is interrupted.
async fn cancelDownload() -> null {
	runAction("cancel")
}

# The URL contains the job id, so it is dynamic: the *Safe variant with a
# literal allow-list is the one the security rules require.
async fn runAction(act: string) -> null {
	if jobId == "" {
		message = "no download to manage"
		return
	}
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/" + act, "", "http://127.0.0.1:8090")
	if err != null {
		message = "operation " + act + " failed: " + err
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		message = "invalid response: " + jerr
		return
	}
	updateJob(job)
	message = "operation " + act + " done"
}

# The operator selects a row of the parts table: the part number is kept in a
# state, so the retry button always knows which part to fetch again.
fn onPartSelected() -> null {
	sel, e = table.selectedRows("parts")
	if e != null {
		message = "cannot read the selection: " + e
		return
	}
	if len(sel) == 0 {
		return
	}
	rows, e2 = table.data("parts")
	if e2 != null {
		message = "cannot read the parts table: " + e2
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
		message = "no download to manage"
		return
	}
	if selectedPart == "" {
		message = "select a row in the Data tab first"
		return
	}
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/parts/" + selectedPart + "/retry", "", "http://127.0.0.1:8090")
	if err != null {
		message = "retry failed: " + err
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		message = "invalid response: " + jerr
		return
	}
	updateJob(job)
	message = "part " + selectedPart + " queued again"
}

# Retry every part that failed.
async fn retryFailedParts() -> null {
	if jobId == "" {
		message = "no download to manage"
		return
	}
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/retry-failed", "", "http://127.0.0.1:8090")
	if err != null {
		message = "retry failed: " + err
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		message = "invalid response: " + jerr
		return
	}
	updateJob(job)
	message = "the failed parts are queued again"
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

	card class: "hero" {
		row class: "hero-row" {
			col class: "hero-text" {
				h1 "Download manager"
				p "The backend splits the file into parts, downloads every one with its own Range request and reassembles them at the end."
			}
			badge "chunked" class: "hero-badge"
		}
	}

	card class: "panel" {
		h3 "1 · Target"
		row class: "field" {
			input "File URL" value: urlText placeholder: "https://example.com/file.bin" class: "url-input"
		}
		p "Target: {urlText}" class: "hint"
		row class: "controls" {
			combobox "Part size" options: "256 KiB,512 KiB,1 MiB,2 MiB,4 MiB,8 MiB,16 MiB" value: partSize class: "size-select"
			button "Start download" onClick: startDownload class: "btn-primary"
			button "Pause" onClick: pauseDownload class: "btn-soft"
			button "Resume" onClick: resumeDownload class: "btn-soft"
			button "Cancel" onClick: cancelDownload class: "btn-ghost"
		}
		p "Every part will be {partSize} - the backend splits the file accordingly." class: "hint"
		row class: "status-row" {
			badge "{jobState}" class: "state-badge"
			p "{message}" class: "status"
		}
	}

	tabs "Download" {
		tab "Progress" {
			card class: "panel" {
				h3 "2 · Total progress"
				progress value: percent max: 100 class: "bar-total"
				row class: "stats" {
					col class: "stat" {
						p "{percent}%" class: "stat-value"
						p "downloaded" class: "stat-label"
					}
					col class: "stat" {
						p "{progressText}" class: "stat-value-sm"
						p "bytes" class: "stat-label"
					}
					col class: "stat" {
						p "{speedText}" class: "stat-value-sm"
						p "speed" class: "stat-label"
					}
					col class: "stat" {
						p "{etaText}" class: "stat-value-sm"
						p "time left" class: "stat-label"
					}
				}
				p "{verifyText}" class: "verify"
				if jobState == "done" {
					a "Download the reassembled file" href: "http://127.0.0.1:8090/api/downloads/" + jobId + "/file" class: "dl-link"
				}
			}

			card class: "panel" {
				h3 "3 · Parts"
				if len(partsView) == 0 {
					p "No download yet: the bars of the parts appear here." class: "muted"
				} else {
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
}

css "body \{ margin: 0; padding: 0 0 40px; background: radial-gradient(1100px 520px at 12% -8%, rgba(99, 102, 241, 0.22) 0%, rgba(15, 23, 42, 0) 60%), linear-gradient(165deg, #0b1020 0%, #131a30 55%, #0e1426 100%); color: #e6e9f5; font-family: system-ui, sans-serif; \} h1 \{ font-size: 26px; margin: 0 0 6px; letter-spacing: -0.02em; \} h3 \{ margin: 0 0 14px; font-size: 15px; text-transform: uppercase; letter-spacing: 0.08em; color: #a5b4fc; \} p \{ line-height: 1.55; \} card \{ background: rgba(255, 255, 255, 0.04); border: 1px solid rgba(148, 163, 184, 0.16); border-radius: 18px; padding: 20px 22px; margin: 16px 0; box-shadow: 0 18px 40px rgba(2, 6, 23, 0.45); \} .hero \{ background: linear-gradient(135deg, rgba(99, 102, 241, 0.20) 0%, rgba(34, 211, 238, 0.10) 100%); border: 1px solid rgba(129, 140, 248, 0.32); border-radius: 20px; padding: 22px 24px; margin: 20px 0 4px; \} .hero-row \{ align-items: flex-start; gap: 16px; } .hero-text \{ flex: 1 1 auto; gap: 6px; } .hero-text p \{ margin: 0; color: #a8b1cc; max-width: 640px; } .hero-badge \{ background: rgba(34, 211, 238, 0.16); color: #67e8f9; border: 1px solid rgba(34, 211, 238, 0.35); font-weight: 600; letter-spacing: 0.04em; } .field \{ margin: 4px 0 16px; } .url-input \{ flex: 1 1 100%; } .controls \{ gap: 10px; flex-wrap: wrap; margin-bottom: 14px; } .size-select \{ flex: 0 0 190px; } .sui-button.btn-primary \{ background: linear-gradient(135deg, #6366f1 0%, #22d3ee 100%); color: #061021; border: 0; font-weight: 700; padding: 10px 20px; border-radius: 12px; box-shadow: 0 10px 24px rgba(99, 102, 241, 0.35); } .sui-button.btn-soft \{ background: rgba(99, 102, 241, 0.16); color: #c7d2fe; border: 1px solid rgba(129, 140, 248, 0.38); padding: 9px 16px; border-radius: 12px; } .sui-button.btn-warn \{ background: rgba(251, 191, 36, 0.16); color: #fcd34d; border: 1px solid rgba(251, 191, 36, 0.42); padding: 9px 16px; border-radius: 12px; } .sui-button.btn-ghost \{ background: transparent; color: #94a3b8; border: 1px solid rgba(148, 163, 184, 0.28); padding: 9px 16px; border-radius: 12px; } .sui-button.btn-primary:hover \{ filter: brightness(1.08); } .sui-button.btn-soft:hover \{ background: rgba(99, 102, 241, 0.26); } .status-row \{ gap: 12px; align-items: center; border-top: 1px solid rgba(148, 163, 184, 0.14); padding-top: 14px; } .state-badge \{ background: rgba(148, 163, 184, 0.16); color: #cbd5f5; border: 1px solid rgba(148, 163, 184, 0.28); font-weight: 600; text-transform: uppercase; font-size: 11px; letter-spacing: 0.06em; } .status \{ margin: 0; color: #94a3b8; font-size: 13px; } .muted \{ color: #7d879f; font-size: 12px; margin: 0; } .hint \{ color: #7d879f; font-size: 12px; margin: 0 0 14px; word-break: break-all; } .stats \{ display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px; margin-top: 16px; } .stat \{ gap: 2px; padding: 12px 14px; border-radius: 12px; background: rgba(148, 163, 184, 0.08); border: 1px solid rgba(148, 163, 184, 0.12); } .stat-value \{ font-size: 22px; font-weight: 700; color: #f1f5f9; margin: 0; } .stat-value-sm \{ font-size: 15px; font-weight: 600; color: #e2e8f0; margin: 0; word-break: break-word; } .stat-label \{ font-size: 11px; text-transform: uppercase; letter-spacing: 0.07em; color: #7d879f; margin: 0; } .sui-progress.bar-total \{ --sui-progress-height: 14px; --sui-progress-radius: 999px; --sui-progress-track: rgba(148, 163, 184, 0.18); --sui-progress-value: #6366f1; --sui-progress-value-end: #22d3ee; box-shadow: 0 6px 18px rgba(99, 102, 241, 0.25); } .sui-progress.bar-part \{ --sui-progress-height: 8px; --sui-progress-radius: 999px; --sui-progress-track: rgba(148, 163, 184, 0.16); --sui-progress-value: #818cf8; --sui-progress-value-end: #38bdf8; } .row.part-row \{ display: grid; grid-template-columns: 64px minmax(120px, 1fr) 92px 148px; align-items: center; gap: 14px; padding: 8px 10px; border-radius: 10px; } .row.part-row:hover \{ background: rgba(148, 163, 184, 0.08); } .part-index \{ font-weight: 600; color: #c7d2fe; margin: 0; font-variant-numeric: tabular-nums; } .part-state \{ font-size: 12px; color: #8b93b0; margin: 0; } .part-bytes \{ font-size: 12px; color: #8b93b0; text-align: right; margin: 0; font-variant-numeric: tabular-nums; } .part-error \{ grid-column: 1 / -1; font-size: 12px; color: #fda4af; margin: 0; } .verify \{ margin: 14px 0 0; color: #6ee7b7; font-size: 13px; } .dl-link \{ display: inline-block; margin-top: 12px; color: #67e8f9; font-weight: 600; } .sui-tabs \{ --sui-tabs-border: rgba(148, 163, 184, 0.16); --sui-tabs-radius: 16px; --sui-tabs-bg: rgba(255, 255, 255, 0.02); --sui-tablist-bg: rgba(15, 23, 42, 0.55); --sui-tab-padding: 10px 18px; --sui-tab-active-bg: rgba(99, 102, 241, 0.18); --sui-tab-accent: #22d3ee; --sui-tab-fg: #94a3b8; --sui-tab-active-fg: #e0e7ff; } .sui-tabpanels \{ padding: 18px 20px; } @media (max-width: 760px) \{ .row.part-row \{ grid-template-columns: 56px 1fr; row-gap: 6px; } .part-bytes \{ text-align: left; } \}"
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

The automated test in `internal/core/download_manager_test.go` drives the same
cycle headlessly: it builds the backend, splits a 1 MiB file into two 512 KiB
parts, pauses, checks that no byte advances, resumes and compares the
reassembled file with the original one. A second test uses an origin that
fails one range once and checks that the retry of that single part brings the
job to `done`.
