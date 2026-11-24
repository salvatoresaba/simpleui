# Download manager

A chunked downloader with a graphical front end. The backend is a small Go
program; the front end is a single SimpleUI file. Together they show a
complete "operation" app: a long running job, split into parts, tracked one by
one, pausable, resumable and verifiable.

```
examples/download_manager/
  app.sui        the SimpleUI GUI (one file, type-checked)
  server/main.go the Go backend: file server + download manager
  run.ps1        launcher: builds the app and starts the backend
  run.bat        same for cmd.exe
```

The backend lives in its own Go module (`server/go.mod`), so the example does
not add a dependency to the root module.

## What it demonstrates

- **A split decided by the size.** The file is cut into parts of a chosen
  length (`256 KiB` … `16 MiB`); the number of parts follows from the file
  size. A 24 MiB file with 1 MiB parts becomes 24 parts.
- **One Range request per part.** Every part is fetched with its own
  `Range: bytes=start-end` request, so the parts are really independent
  requests and not a single stream split in the client.
- **A table of parts.** Every part keeps its own state (`pending`,
  `downloading`, `done`, `error`) and its own byte counter. The GUI shows the
  whole table, not just a percentage.
- **Pause and resume.** Pausing cancels the running requests but keeps the
  bytes already on disk. Resuming asks for `bytes=start+received-end`, so a
  part that was half downloaded continues from the byte it stopped at.
- **Reassembly and verification.** When every part is complete the backend
  concatenates them in order and compares the SHA-256 of the result with the
  SHA-256 of the original file.
- **Polling with `interval`.** The GUI polls the status endpoint every 500 ms
  and repaints the progress bar and the parts table.
- **A dynamic URL done safely.** The pause/resume/cancel URL contains the job
  id, so it is a dynamic URL: the app uses `http.postSafe` with a literal
  allow-list, exactly as the security rules require.
- **The backend owns the job state.** Pause, resume and cancel are HTTP
  requests to the server, never local flags: the GUI can be reloaded at any
  moment without losing the download.

## How to run

```bash
# build the app and start the backend (Windows)
pwsh examples/download_manager/run.ps1

# or by hand
simpleui build examples/download_manager/app.sui --target web --out examples/download_manager/public
cd examples/download_manager/server
go run . --port 8090 --web-dir ../public --files-dir ../files --work-dir ../downloads
```

Then open <http://127.0.0.1:8090/>. The backend creates a deterministic
24 MiB sample file (`files/sample.bin`) on first start, so the example works
offline and always produces the same checksum.

Useful flags:

| Flag | Default | Meaning |
|---|---|---|
| `--port` | `8090` | HTTP port |
| `--web-dir` | `../public` | folder with the compiled SUI app |
| `--files-dir` | `../files` | folder with the files offered for download |
| `--work-dir` | `../downloads` | folder with the parts and the merged file |
| `--sample-mib` | `24` | size of the generated sample file |
| `--workers` | `4` | parallel part downloads |
| `--throttle` | `0` | bytes/s per worker; slows the download down so the progress is visible |
| `--no-open` | – | do not open the browser |

Put any file in `--files-dir` and it shows up in the app: the catalog comes
from the server, not from a hardcoded list.

### The port is part of the app

The GUI calls `http://127.0.0.1:8090/api/...` with **literal** URLs, and the
dynamic job URL is allowed only for the literal origin `http://127.0.0.1:8090`.
Running the backend on a different port therefore means editing `app.sui` too:
change the three URLs and the allow-list, then rebuild the app. The backend
prints the URL it listens on.

## The GUI

1. **Pick what to download** – the file (a combobox filled at runtime from
   `GET /api/files`) and the part size.
2. **Progress** – the progress bar, the state, the bytes downloaded, the
   speed, the remaining time and the SHA-256 verdict. When the download is
   complete a link appears to download the reassembled file.
3. **Parts of the download** – one row per part: number, byte range, size,
   bytes received, percentage and state.

## HTTP API

| Endpoint | Purpose |
|---|---|
| `GET /api/files` | catalog of downloadable files |
| `POST /api/downloads` | start a job: `{"file":"sample.bin","partSize":"2 MiB"}` |
| `GET /api/downloads` | all jobs, most recent first |
| `POST /api/downloads/{id}/pause` | pause: keep the bytes, stop the requests |
| `POST /api/downloads/{id}/resume` | resume: every part continues from its byte |
| `POST /api/downloads/{id}/cancel` | cancel the job |
| `GET /api/downloads/{id}/file` | the reassembled file |
| `GET /files/{name}` | the file to download, with `Range` support |

A job has a state of its own: `running`, `paused`, `merging`, `done`,
`canceled` or `error`.

The parts in the JSON response use the column names of the table, so the
backend and the GUI always agree on what a part is:

```json
{"Part": 3, "From": 2097152, "To": 3145727, "Size": 1048576,
 "Received": 1048576, "Percent": 100, "State": "done"}
```

The catalog uses the same convention: `{"Name": "sample.bin", "Size": 25165824,
"SizeText": "24.0 MiB"}`.

## The complete SUI source

`app.sui`, the whole file:

```sui
import { http, json } from "std"

# Download manager - the front end of the chunked downloader.
#
# The Go backend (server/main.go) splits the file into parts according to its
# size, downloads every part with its own Range request, tracks each part and
# finally reassembles them into one file verified against the SHA-256 of the
# original.
#
# The GUI follows the progress, shows the table of the parts and lets the
# operator pause, resume and cancel the download.
#
# The backend listens on port 8090: the URLs below are literals, and the only
# dynamic URL (the one that carries the job id) goes through http.postSafe
# with a literal allow-list.

state fileOptions = ""
state selectedFile = ""
state partSize = "1 MiB"
state jobId = ""
state jobState = "no download"
state percent = 0
state progressText = "waiting"
state speedText = "--"
state etaText = "--"
state verifyText = ""
state message = "Pick a file and press Start download."

fn onFileChange() -> null {
	message = "Selected file: " + selectedFile
}

fn onPartSizeChange() -> null {
	message = "Every part will be " + partSize
}

# Catalog of the downloadable files: fills the combobox and the table.
async fn loadCatalog() -> null {
	body, err = http.get("http://127.0.0.1:8090/api/files")
	if err != null {
		message = "backend unreachable: " + err
		return
	}
	items, jerr = json.parse(body)
	if jerr != null {
		message = "invalid response: " + jerr
		return
	}
	let names = ""
	let i = 0
	while i < len(items) {
		let name = str(items[i]["Name"])
		if names == "" {
			names = name
			# The default file is selected in the same render as the options:
			# a value set earlier would not be in the list yet and the control
			# would drop it.
			selectedFile = name
		} else {
			names = names + "," + name
		}
		i = i + 1
	}
	fileOptions = names
	n, terr = table.display("files", items)
	if terr != null {
		message = "files table not updated: " + terr
	}
	loadStatus()
}

# State of the downloads: the most recent job and its parts.
async fn loadStatus() -> null {
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
	if len(jobs) == 0 {
		return
	}
	updateJob(jobs[0])
}

# Polling: re-read the state of the current download.
async fn refresh() -> null {
	if jobId == "" {
		return
	}
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
	n, terr = table.display("parts", job["parts"])
	if terr != null {
		message = "parts table not updated: " + terr
	}
}

# Start a download: the backend picks the parts according to the size.
async fn startDownload() -> null {
	if selectedFile == "" {
		message = "pick a file first"
		return
	}
	payload, jerr = json.stringify({"file": selectedFile, "partSize": partSize})
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
	jobId = str(job["id"])
	message = "download started: " + str(job["file"]) + " - " + str(job["partCount"]) + " parts of " + str(job["partSizeText"])
	updateJob(job)
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

async fn boot() -> null {
	interval(500, refresh)
	loadCatalog()
}

view main {
	title "Download manager"
	onLoad: boot

	h1 "Download manager"
	p "The backend splits the file into parts, downloads every one with its own Range request, tracks them and finally reassembles the file checking its SHA-256."

	card {
		h3 "1. Pick what to download"
		row {
			combobox "File" options: fileOptions value: selectedFile onChange: onFileChange placeholder: "Pick a file..." empty: "No file available" maxOptionLength: 64 maxOptions: 32
			combobox "Part size" options: "256 KiB,512 KiB,1 MiB,2 MiB,4 MiB,8 MiB,16 MiB" value: partSize onChange: onPartSizeChange
		}
		row {
			button "Start download" onClick: startDownload
			button "Pause" onClick: pauseDownload
			button "Resume" onClick: resumeDownload
			button "Cancel" onClick: cancelDownload
		}
	}

	card {
		h3 "2. Progress"
		progress value: percent max: 100
		p "State: {jobState} - {progressText}"
		p "Speed: {speedText} - time left: {etaText}"
		p "{verifyText}"
		if jobState == "done" {
			a "Download the reassembled file" href: "http://127.0.0.1:8090/api/downloads/" + jobId + "/file"
		}
	}

	h3 "3. Parts of the download"
	table name: "parts" {
		thead {
			trow {
				tcell "Part" head: true
				tcell "From" head: true
				tcell "To" head: true
				tcell "Size" head: true
				tcell "Received" head: true
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
			}
		}
	}

	h3 "Available files"
	table name: "files" {
		thead {
			trow {
				tcell "Name" head: true
				tcell "Size" head: true
			}
		}
		tbody {
			trow {
				tcell "-"
				tcell "-"
			}
		}
	}

	p "{message}" class: "muted"
}

css "body \{ margin: 0; background: #f4f6fb; color: #0f172a; font-family: system-ui, sans-serif; \} .muted \{ color: #64748b; font-size: 13px; \} h3 \{ margin: 18px 0 8px; \} card \{ display: block; background: #ffffff; border: 1px solid #e2e8f0; border-radius: 12px; padding: 16px; margin: 12px 0; \} row \{ gap: 12px; flex-wrap: wrap; margin: 8px 0; \} table \{ width: 100%; margin-bottom: 12px; \}"
```

### How to edit it

1. Change `app.sui` (the designer can open it: `simpleui gui`).
2. Type-check it: `simpleui check app.sui --json`.
3. Rebuild it: `simpleui build app.sui --target web --out public`.
4. Restart the backend (`run.ps1` does both).

## Notes

- The sample file is generated from a deterministic PRNG: the same bytes, the
  same SHA-256, on every machine.
- The backend only serves files from `--files-dir` and never follows a URL
  provided by the client: there is no SSRF surface in the example.
- The parts stay in `--work-dir` after the job ends. They are the evidence of
  how the file was split, and deleting them is a deliberate act.
- A job is not persisted between restarts of the backend: the example keeps
  the job state in memory and the parts on disk.
- The catalog of the files and the parts table use the JSON keys of the
  backend as column headers, so the table always shows the columns in the
  order the backend defines them.

## Verified behaviour

Measured against this example (backend and GUI, not just the documentation):

| Scenario | Result |
|---|---|
| 24 MiB file, 1 MiB parts | 24 parts, each with its own Range request |
| Pause at 10% | state `paused`, the running parts keep their bytes, nothing advances (2.5 MiB before and after 3 s) |
| Resume | the parts continue from the byte they stopped at, the job completes |
| Reassembly | 25 165 824 bytes, SHA-256 identical to the original file |
| GUI | progress bar, parts table, pause/resume and the download link all work |

The automated test in `internal/core/download_manager_test.go` drives the
same cycle headlessly: it builds the backend, splits a 1 MiB file into two
512 KiB parts, pauses, checks that no byte advances, resumes and compares the
SHA-256 of the reassembled file with the original one.
