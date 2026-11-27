# Download manager

A chunked downloader with a graphical front end. The backend is a small Go
program; the front end is a single SimpleUI file. Together they show a
complete "operation" app: a long running job, split into parts, tracked one by
one, pausable, resumable, restartable part by part and verifiable.

```
examples/download_manager/
  app.sui           the SimpleUI GUI (one file, type-checked)
  server/           the Go backend: the library of the download manager
  server/cmd/dm/    the standalone program (the library and the app together)
  run.ps1           launcher: builds the app and starts the backend
  run.bat           same for cmd.exe
  build.ps1         release build: one executable with the app and the backend
  build.bat         same for cmd.exe
```

The backend lives in its own Go module (`server/go.mod`), so the example does
not add a dependency to the root module. The root package of that module is a
**library** (`downloadmanager`): it exports `Handler(fallback http.Handler) http.Handler`,
the contract the compiler uses to mount a backend in front of a compiled app.
The program of the example is `server/cmd/dm`: it starts the library on its own
port and serves the app, from the disk or from the executable itself.

## What it demonstrates

- **One executable, app and downloader inside.** `build.ps1` asks the compiler
  for a desktop build with the backend of the example:
  `simpleui build app.sui --target desktop --backend server`. The compiler
  mounts the Go backend in front of the compiled app, so the release build is a
  single file that opens the app and downloads with it: no other file and no
  other process are needed.
- **http, https and ftp.** The URL of a target can be an http, an https or an
  ftp address. The ftp client is written here, without any dependency: it asks
  the server for the size of the file (`SIZE`), starts a transfer in the middle
  of it (`REST`) and reads the bytes (`PASV` + `RETR`). A server that does not
  know the restart command gets the whole file in one single part. The login of
  the ftp server travels in the URL (`ftp://user:password@host/file.bin`) and
  the password never shows in the app: the URL that is displayed is masked.
- **A URL pasted without a scheme works.** A URL copied from a page often
  keeps the protocol-relative form (`//host/path`) or loses the scheme
  completely (`host/path`): both are read as an **https** URL, the safe default
  for a download. A URL that names another scheme (say `mailto:`) is still
  refused, with a message that says which schemes are allowed.
- **The app finds the backend by itself.** When the app starts it asks the
  server that served the page for its configuration: if that server is the
  download manager (the usual case) the calls stay on the same origin, so the
  app works on every port. If the page comes from somewhere else — the desktop
  build, which serves the app alone — the app falls back to the address of the
  backend. When the backend does not answer at all, the toast says how to
  start it.
- **The CLI takes the URL of the file.** `dm.exe --url https://example.com/file.bin`
  is the file to download; the other flags are the configuration (part size,
  workers, throttle, allowed hosts, expected checksum). The GUI reads the
  defaults from `GET /api/config` and shows them in the settings panel, where
  the operator can change them.
- **The name of the file to save.** Every target row has a box with the name of
  the file to save, next to the URL. The program fills it from the URL (the last
  segment of the path, cleaned) and the operator can change it before the
  download starts; once it is changed the program stops filling it. The name
  travels with the request and the backend uses it for the file it writes and
  for the name the browser saves it with.
- **Two ways to download.** The settings panel (the small ⚙ button of the
  header) chooses between **one single file** (one request, no parts, no
  reassembly) and **split into parts**. In split mode the panel also shows the
  size of the parts and how many parts are downloaded at the same time.
- **A big file still downloads.** The parts of a file are 1 MiB long by default, so
  a file of 300 MiB would need more parts than the CLI allows (256). The
  manager does not refuse the file: the parts grow (2 MiB, 4 MiB, ...) until
  the file fits, and the title of the parts panel shows the length really
  used.
- **A server that does not announce the size is fine.** Some servers answer
  without a `Content-Length` and ignore the `Range` header. The file is then
  downloaded in one open ended part, that goes on until the server closes the
  connection: there is no percentage and no time left (the status bar shows
  `--`), but the bytes that have arrived are counted.
- **One row for every file.** The **Targets** card holds one row for every file
  that is going to be downloaded, that is downloading or that has been
  downloaded. A row is the URL, the name of the file, the priority as an icon
  and the buttons of the file (play/pause, stop/start again, cancel), all on
  one line; under it a thin bar with the bytes. The **+** button of the card
  adds a row: every file that is added gets its own row.
- **The first row is always there.** The row where the operator types the URL
  is created when the app starts, before the backend answers: the field is
  there even when the backend is not reachable yet, and the **+** button always
  adds a row. The list of the rows is never inside a branch of a conditional
  block, so a row can never disappear with a branch.
- **The URL is locked while the file downloads.** The URL of a row can be
  changed only before the download starts or after the file has been stopped.
  The stop button (■) turns the URL back into an editable box: the operator
  writes the new URL and presses play, and the file is downloaded again from
  the new URL. A file that is complete can be stopped too, so its URL can be
  changed in the same way.
- **More files in the same download.** The files of the rows **share the queue
  of the parts and the workers of the job**: with six workers and two files of
  the same priority the parts are taken three by three, and a file added later
  joins the queue of the parts that are still waiting.
- **A priority for every file.** Every file has its own priority — **low**,
  **normal** or **high** — shown as an icon (↓, →, ↑) in its row: a click on
  the icon moves the priority from low to normal to high and back to low. The
  scheduler always takes the parts of the files with the highest priority first
  and then goes round-robin between the files with the same priority.
- **The buttons act on one file only.** The play button of a row pauses and
  resumes that file, the stop button stops it (or starts it again) and the
  cancel button takes it out of the download and deletes its bytes. The other
  files of the download go on.
- **The parts of the file are deleted after the merge.** When a file is
  reassembled the backend verifies the result and then removes the `part-*.bin`
  files from the disk: only the reassembled file stays in the folder of the
  download.
- **A split decided by the size.** The file is cut into parts of a chosen
  length (`256 KiB` … `16 MiB`); the number of parts follows from the file
  size. A 24 MiB file with 1 MiB parts becomes 24 parts.
- **One Range request per part.** Every part is fetched with its own
  `Range: bytes=start-end` request, so the parts are really independent
  requests and not a single stream split in the client.
- **Two levels of progress.** A bar for the whole download and, inside the same
  panel, a collapsible section with one bar per part. Only the parts whose
  download has started get a bar: a part that is still waiting shows nothing.
  The bars start from the **most recent** part and the list has a limited
  height with a scroll bar, so the part that is running now is always visible.
- **The counters of the parts.** The section of the parts opens with three
  numbers: how many parts the job has, how many are downloading right now and
  how many are complete. They follow the state of the job at every poll.
- **The data in another tab.** The tables of the parts and of the jobs live in
  the second tab, so the live view stays readable.
- **One button for the whole download.** The small button of the header pauses
  the whole download and resumes it; its hint says what it is going to do.
- **Pause, resume and cancel.** Pausing cancels the running requests but keeps
  the bytes already on disk. Resuming asks for `bytes=start+received-end`, so a
  part that was half downloaded continues from the byte it stopped at. The
  cancel stops the job.
- **The errors are toasts.** Every failure of the app and of the backend is
  shown as an error toast (and as a line of text), and the same message is not
  repeated twice in a row.
- **A sound at the end.** When the download completes the app plays a short
  chime served by the backend (`GET /sound.wav`) with the new `sound` module.
- **A status bar.** The state of the download and its message are in the
  status bar at the bottom of the screen, in a row above the data of the
  download: percentage, size of the file, bytes downloaded and estimated time
  left.
- **A failed part restarts alone.** The operator selects the failed row in the
  table and presses "Retry part N": only that part is fetched again, its bytes
  are reset and the other parts are not touched. "Retry all failed" does the
  same for every failed part at once.
- **Reassembly and verification.** When every part is complete the backend
  concatenates them in order and compares the size with the total the server
  announced; with `--expect-sha256` it also compares the SHA-256 of the result.
  The parts are then deleted: the folder of the download keeps only the
  reassembled files.
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
# the release build: one executable with the app and the downloader inside
pwsh examples/download_manager/build.ps1
.\examples\download_manager\bin\download_manager.exe

# build the app and start the backend (Windows): by default the backend
# downloads its own sample file, so the example works offline
pwsh examples/download_manager/run.ps1

# or a real file, with another part size
pwsh examples/download_manager/run.ps1 -Url "https://example.com/file.bin" -PartSize "2 MiB"

# a file of an ftp server: the login travels in the URL
pwsh examples/download_manager/run.ps1 -Url "ftp://user:password@ftp.example.com/pub/file.bin"

# headless: build and start the backend without opening the browser
pwsh examples/download_manager/run.ps1 -NoOpen

# or by hand
simpleui build examples/download_manager/app.sui --target web --out examples/download_manager/public
cd examples/download_manager/server
go run ./cmd/dm --port 8090 --web-dir ../public --files-dir ../files --out-dir ../downloads --url http://127.0.0.1:8090/files/sample.bin
```

Then open <http://127.0.0.1:8090/>. The backend creates a deterministic
24 MiB sample file (`files/sample.bin`) on first start, so the example works
offline and always produces the same bytes.

**The example needs the backend.** The GUI talks to the download manager, so
the two have to meet: the app looks for it at the start and asks the server
that served the page first, then the address of the backend (`127.0.0.1:8090`).
The **release build** (`build.ps1`) is the simple way: one executable carries
the app and the backend, so the download works as soon as it is started. The
**development launcher** (`run.ps1`) does the same, but it serves the app from
the disk (`public/`), so the app can be rebuilt while it runs. The **desktop**
target without a backend (`simpleui build app.sui --target desktop`) compiles
the same GUI into an executable that serves it on its own port: it finds the
backend when one is running on 8090, and when there is none the toast says how
to start it.

### The flags of the CLI

| Flag | Default | Meaning |
|---|---|---|
| `--url` | (empty) | URL of the file to download; the GUI shows it and the operator can change it. Without it the manager downloads its own sample file, at the port it listens on |
| `--part-size` | `1 MiB` | length of one part: the file is split accordingly |
| `--workers` | `4` | parallel part downloads |
| `--throttle` | `0` | bytes/s per worker; slows the download down so the progress is visible |
| `--max-parts` | `256` | upper bound of the parts of one job: when the file does not fit in them, the parts grow instead of refusing the download |
| `--out-dir` | `../downloads` | folder with the parts and the merged file |
| `--web-dir` | (empty) | folder with the compiled SUI app; without it the app embedded in the executable is served |
| `--files-dir` | `../files` | folder with the files offered for download |
| `--sample-mib` | `24` | size of the generated sample file; `0` disables the generation |
| `--expect-sha256` | (empty) | expected SHA-256 of the file: the merge verifies it |
| `--allow-host` | (empty) | host that may be downloaded; the flag is repeatable |
| `--port` | `8090` | HTTP port |
| `--no-open` | – | do not open the browser |

Only `http`, `https` and `ftp` URLs are accepted. A URL written without a
scheme (`//host/path` or `host/path`) is read as an **https** URL. The
credentials of an http URL are refused; an ftp URL carries the login of the
server (`ftp://user:password@host/file.bin`) and the password is masked in the
URL that the app displays. With `--allow-host` only the listed hosts can be
downloaded:

```bash
dm.exe --url https://example.com/file.bin --allow-host example.com --allow-host cdn.example.com
```

### The port does not matter

The GUI asks the server that served the page for the configuration, so the
calls of the API stay on that origin: the app works on every port and the
backend does not have to be on 8090. When the page comes from another host
(the desktop build), the app falls back to `http://127.0.0.1:8090`, and the
dynamic job URL is allowed for the API of the backend only — the allow-list is
the literal `/api http://127.0.0.1:8090/api`, exactly as the security rules
require. The backend prints the URL it listens on.

## The GUI

1. **Header** – the title of the application, the button that pauses or resumes
   the whole download and the small ⚙ button that opens the settings panel as
   an overlay.
2. **Targets card** – one row for every file of the download, all on one line:
   the URL (an editable box, or the text of the URL when the file is running
   and the box is locked), the name of the file to save, the icon of the
   priority (↓, →, ↑ — a click moves the priority) and the buttons of the file
   (▶ / ❚❚ play and pause, ■ stop or ↻ start again, ✕ cancel). Under the row a
   thin bar and the bytes that have arrived. The **+** button of the card adds
   a row for another file.
3. **Progress tab** – the card of the total progress: the bar of the whole
   download and, inside it, the collapsible section with the counters of the
   parts (how many there are, how many are downloading, how many are complete)
   and one bar per part (the most recent one first, only the parts whose
   download has started). When the download is complete a link appears to
   download the reassembled file.
4. **Files tab** – the files of the download: the table with the id, the name,
   the priority, the state, the number of parts and the size of every file,
   and under it one row per file with the link to its reassembled file. The
   priority of a file is changed from its row in the **Targets** card.
5. **Data tab** – the table of the parts (with the file and its priority) and
   the table of the jobs. Clicking a row of the parts table selects the part:
   the "Retry part N" button fetches that part again, "Retry all failed"
   requeues every failed part.
6. **Status bar** – the state of the download and its message in the first
   row, the percentage, the size of the file, the bytes downloaded and the
   estimated time left in the row below.

## HTTP API

| Endpoint | Purpose |
|---|---|
| `GET /api/config` | the configuration of the CLI (default URL, part size, workers, ...) |
| `POST /api/downloads` | start a job: `{"url":"...","partSize":"2 MiB","workers":4,"single":false,"file":"report.pdf","priority":"normal"}` |
| `GET /api/downloads` | all jobs, most recent first |
| `POST /api/downloads/{id}/files` | add a file to the job: `{"url":"...","file":"report.pdf","priority":"high"}` |
| `POST /api/downloads/{id}/files/{file}/priority` | change the priority of a file: `{"priority":"low"}` |
| `POST /api/downloads/{id}/files/{file}/pause` | pause one file: the bytes are kept, the other files go on |
| `POST /api/downloads/{id}/files/{file}/resume` | take one paused file back into the queue |
| `POST /api/downloads/{id}/files/{file}/stop` | stop one file: it stays in the list and its URL can be changed again |
| `POST /api/downloads/{id}/files/{file}/restart` | download one file again from its first byte, with `{"url":"..."}` when the URL changed |
| `POST /api/downloads/{id}/files/{file}/cancel` | take one file out of the job and delete its bytes |
| `POST /api/downloads/{id}/pause` | pause: keep the bytes, stop the requests |
| `POST /api/downloads/{id}/resume` | resume: every part continues from its byte |
| `POST /api/downloads/{id}/cancel` | cancel the job |
| `POST /api/downloads/{id}/parts/{n}/retry` | download one failed part again |
| `POST /api/downloads/{id}/retry-failed` | download every failed part again |
| `GET /api/downloads/{id}/file` | the reassembled file (the first one that is complete) |
| `GET /api/downloads/{id}/files/{file}/file` | the reassembled file of one file of the job |
| `GET /files/{name}` | a local file, with `Range` support |
| `GET /sound.wav` | the short chime played when a download completes |

`workers` is the number of parts downloaded at the same time; `single` switches
the job to the one-piece mode (one request, no parts, no reassembly). In the
one-piece mode the size of the file is unknown until the server answers, so the
part stays open and the request uses `Range: bytes=N-`.

`file` is the name of the file to save, the one the operator chose in the GUI.
It is optional: without it the name comes from the URL. The backend cleans it
like the name of the URL (only letters, digits, dot, dash and underscore, the
rest becomes an underscore), so a name can never point outside the folder of
the job. The cleaned name is the name of the reassembled file and the name the
browser saves it with.

A job has a state of its own: `running`, `paused`, `merging`, `done`,
`canceled` or `error`. A part has a state too: `pending`, `downloading`,
`done` or `error`. A file has its own state: `pending`, `running`, `paused`
(paused on its own), `stopped` (stopped, the URL can be changed again),
`done`, `error` or `canceled`.

A job holds one or more **files**. Every file has its own id (`file-1`,
`file-2`, ...), its own name, its own priority (`low`, `normal` or `high`) and
its own parts; the parts of the whole job are numbered together, so the part
files of the job are `part-0001.bin`, `part-0002.bin`, ... The scheduler
always takes a part of a file with the highest priority and walks the files
of the same priority round-robin: with six workers and two files of the same
priority the parts are taken three by three, and a file with the `high`
priority jumps in front of the files with the `normal` one.

The parts in the JSON response use the column names of the table, so the
backend and the GUI always agree on what a part is:

```json
{"File": "report.pdf", "Priority": "high", "Part": 3, "From": 2097152,
 "To": 3145727, "Size": 1048576, "Received": 1048576, "Percent": 100,
 "State": "done"}
```

## The complete SUI source

`app.sui`, the whole file:

```sui
import { http, json, sound } from "std"

# Download manager - the front end of the chunked downloader.
#
# The Go backend (the library in server/) takes the URL of a file, splits it
# into parts according to the chosen part size, downloads every part with its
# own Range request, tracks each part and finally reassembles them into one
# file.
#
# A download can hold more than one file. The screen shows the targets: every
# file that is going to be downloaded, or that is being downloaded, has a row
# with its URL, the name of the file to save, its priority as an icon and the
# buttons that act on it (play, pause, stop, start again and cancel). The files
# share the queue of the parts and the workers of the download, so with six
# workers and two files the parts are taken three by three. Every file has its
# own priority (low, normal or high): the scheduler takes the parts of the
# files with the highest priority first and then goes round-robin between the
# files with the same priority.
#
# The URL of a target cannot be changed once its download has started: the
# stop button frees it and the button that starts it again takes the new URL.
#
# The files of the download live in the Progress tab, with their progress and
# the link to the reassembled file; the parts live in a collapsible panel of
# the same tab, shown only when the download runs in split mode, and the bars
# start from the most recent part. The errors are shown as toasts and the
# completion plays a sound.
#
# The backend is found when the app starts: when the page is served by the
# backend itself (the same origin) the calls are relative, so the app works on
# every port. When the page is served by another host (the desktop build) the
# backend has its own address. The calls with a dynamic URL (the ones that
# carry the job id) go through http.postSafe with the API of the backend as
# allow-list.

state partSize = "1 MiB"
state workers = "4"
state single = false
state jobId = ""
state jobState = "idle"
state jobSingle = false
state phase = "idle"
state percent = 0
state totalText = "0 B"
state receivedText = "0 B"
state etaText = "--"
state verifyText = ""
state jobPartSize = ""
state sizeUnknown = false
state message = ""
state partsView: array[any] = []
state partsTable: array[any] = []
state selectedPart = ""
state failedParts = 0
state partsSignature = ""
state lastState = ""
state lastError = ""
# The targets: every file that is going to be downloaded, or that is being
# downloaded, has a row. The id of the row is its key: the handlers of the
# row receive it, so every row acts on itself.
state targets: array[any] = []
state targetSeq = 0
# The counters of the parts: how many there are, how many are downloading and
# how many are done.
state partTotal = 0
state partActive = 0
state partDone = 0
# The files of the download: the table shows them.
state files: array[any] = []
state filesTable: any = []
state filesSignature = ""
# The base of the API: it is empty when the page comes from the backend
# itself (the same origin) and it is the address of the backend when the page
# comes from another host (the desktop build).
state apiRoot = ""
# The address of the backend used when the page does not come from it.
state apiFallback = "http://127.0.0.1:8090"
# The configuration of the backend, when it has answered.
state backendConfig: any = null
# The jobs of the backend, when they have been read.
state jobsList: any = null

# The URL of an API call: the base of the backend plus the path of the call.
fn apiUrl(path: string) -> string {
	return apiRoot + "/api" + path
}

# The URL of a file of the backend that is not part of the API (the sound of
# the end of a download).
fn backendUrl(path: string) -> string {
	return apiRoot + path
}

# The answer of the backend: it is the configuration of the download manager
# and not the page of another program (the desktop build answers every path
# with its own page).
fn isConfig(body: any) -> bool {
	let text = str(body)
	if text == "" {
		return false
	}
	return contains(text, "partSize") and contains(text, "workers")
}

# The message the operator reads when the backend does not answer: it says how
# to start it.
fn backendDown() -> string {
	return "the download manager is not running: start it with run.ps1"
}

# The configuration of the CLI: the default URL, the part size and the default
# number of parts downloaded at the same time. The app starts with them, the
# operator can change them in the settings panel.
async fn loadConfig() -> null {
	# The first row is there before the backend answers: the operator can type
	# the URL of the file to download even when the backend is not there yet.
	ensureFirstTarget()
	# The backend is looked for in the two possible places: the page can be
	# served by the backend itself (the same origin) or by another host (the
	# desktop build), where the backend has its own address.
	backendConfig = null
	body, err = http.getSafe(apiUrl("/config"), "/api http://127.0.0.1:8090/api")
	if err == null {
		parsed, jerr = json.parse(body)
		if jerr == null and isConfig(body) {
			backendConfig = parsed
		}
	}
	if backendConfig == null {
		apiRoot = apiFallback
		body2, err2 = http.getSafe(apiUrl("/config"), "/api http://127.0.0.1:8090/api")
		if err2 == null {
			parsed2, jerr2 = json.parse(body2)
			if jerr2 == null and isConfig(body2) {
				backendConfig = parsed2
			}
		}
	}
	if backendConfig == null {
		apiRoot = ""
		notifyError(backendDown())
		return
	}
	let cfg = backendConfig
	# The URL of the command line fills the first row, but only when the row is
	# still empty: a row that already holds a file keeps its own URL. The rows
	# can be gone (the operator can cancel them), so the list is checked.
	if len(targets) > 0 {
		let first = targetById(str(targets[0]["id"]))
		if str(first["fileId"]) == ""  and  str(first["url"]) == "" {
			setTarget(str(first["id"]), "url", str(cfg["url"]))
			syncTargetName(str(first["id"]))
		}
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

# The title of the parts: how many there are and how long they are. The
# backend grows the length of the parts when the file does not fit in the
# parts the CLI allows, so the title shows the length really used.
fn partsTitle() -> string {
	let text = str(len(partsView)) + " parts started - the most recent one first"
	if jobPartSize != "" {
		text = text + " (" + jobPartSize + " each)"
	}
	return text
}

# The name of the file to save, taken from the URL: the last segment of the
# path, without the query string, cleaned like the backend does.
fn fileNameFromUrl(u: string) -> string {
	let clean = trim(u)
	if clean == "" {
		return "download.bin"
	}
	let noHash = split(clean, "#")
	let noQuery = split(noHash[0], "?")
	let segments = split(noQuery[0], "/")
	let last = trim(segments[len(segments) - 1])
	if last == "" {
		return "download.bin"
	}
	if contains(last, ":") {
		return "download.bin"
	}
	let safe = replaceRegex(last, "[^A-Za-z0-9._-]", "_")
	if safe == "" {
		return "download.bin"
	}
	if safe == "." {
		return "download.bin"
	}
	if safe == ".." {
		return "download.bin"
	}
	return safe
}

# ------------------------------------------------------------------ targets ---
#
# A target is one file that is going to be downloaded, or that is being
# downloaded. Every target has a row in the list and the handlers of the row
# receive the id of the row, so each row acts on itself.

# A row with the default values: the code that reads a field of a row does not
# have to check every time that the row is there.
fn emptyTarget(id: string) -> any {
	return {"id": id, "fileId": "", "url": "", "name": "", "named": "no",
		"priority": "normal", "state": "new", "percent": 0, "received": "0 B", "total": "0 B", "merged": ""}
}

# A new target: it waits for the URL of the operator.
fn newTarget() -> any {
	targetSeq = targetSeq + 1
	return emptyTarget("t" + str(targetSeq))
}

# The row of a target, found by its id.
fn targetById(id: string) -> any {
	let i = 0
	while i < len(targets) {
		if str(targets[i]["id"]) == id {
			return targets[i]
		}
		i = i + 1
	}
	return emptyTarget(id)
}

# The row of a target, found by the id the backend gave to its file.
fn targetByFileId(fileId: string) -> any {
	let i = 0
	while i < len(targets) {
		if str(targets[i]["fileId"]) == fileId {
			return targets[i]
		}
		i = i + 1
	}
	return emptyTarget("")
}

# Change a field of a row: the array is rebuilt with the new row.
fn setTarget(id: string, field: string, value: any) -> null {
	let out = []
	let i = 0
	while i < len(targets) {
		let r = targets[i]
		if str(r["id"]) == id {
			r[field] = value
		}
		out = append(out, r)
		i = i + 1
	}
	targets = out
}

# Take a row away: the target is not wanted any more.
fn removeTarget(id: string) -> null {
	let out = []
	let i = 0
	while i < len(targets) {
		if str(targets[i]["id"]) != id {
			out = append(out, targets[i])
		}
		i = i + 1
	}
	targets = out
}

# Add a row: the operator types the URL in it and presses play.
fn addTarget() -> null {
	targets = append(targets, newTarget())
}

# The first row of the targets: it is always there, so the operator has a URL
# to type in even when the backend does not answer.
fn ensureFirstTarget() -> null {
	if len(targets) > 0 {
		return
	}
	addTarget()
}

# The name of the file follows the URL, but only until the operator changes
# it: after that the name they chose is kept and the program stops filling it.
fn syncTargetName(id: string) -> null {
	let r = targetById(id)
	if str(r["named"]) == "yes" {
		return
	}
	let u = trim(str(r["url"]))
	if u == "" {
		return
	}
	let wanted = fileNameFromUrl(u)
	if wanted == ""  or  wanted == str(r["name"]) {
		return
	}
	setTarget(id, "name", wanted)
}

# The URL of a target: the name of the file follows it.
fn onTargetUrl(ev: any, id: string) -> null {
	setTarget(id, "url", str(ev["target"]["value"]))
	syncTargetName(id)
}

# The name of the file to save: the operator took it over.
fn onTargetName(ev: any, id: string) -> null {
	setTarget(id, "name", str(ev["target"]["value"]))
	setTarget(id, "named", "yes")
}

# The priority of a target: the arrows show the order the parts of the file
# are taken in. The button moves the priority from low to normal to high and
# back to low.
fn prioIcon(p: string) -> string {
	if p == "high" {
		return "↑"
	}
	if p == "low" {
		return "↓"
	}
	return "→"
}

fn prioLabel(p: string) -> string {
	if p == "high" {
		return "High priority"
	}
	if p == "low" {
		return "Low priority"
	}
	return "Normal priority"
}

fn nextPriority(p: string) -> string {
	if p == "low" {
		return "normal"
	}
	if p == "normal" {
		return "high"
	}
	return "low"
}

async fn onTargetPriority(ev: any, id: string) -> null {
	let r = targetById(id)
	let want = nextPriority(str(r["priority"]))
	setTarget(id, "priority", want)
	if str(r["fileId"]) != "" {
		setFilePriority(str(r["fileId"]), want)
	}
}

# The icon of the play button: it follows the state of the target.
fn playIcon(t: any) -> string {
	let s = str(t["state"])
	if s == "running" {
		return "❚❚"
	}
	if s == "done" {
		return "✓"
	}
	if s == "error" {
		return "!"
	}
	return "▶"
}

fn playHint(t: any) -> string {
	let s = str(t["state"])
	if s == "running" {
		return "Pause the download of this file"
	}
	if s == "paused" {
		return "Resume the download of this file"
	}
	if s == "stopped" {
		return "Start the download again with the current URL"
	}
	if s == "done" {
		return "This file is complete"
	}
	return "Start the download of this file"
}

# The icon of the stop button: it stops the file or starts it again.
fn stopIcon(t: any) -> string {
	let s = str(t["state"])
	if s == "running"  or  s == "paused"  or  s == "done" {
		return "■"
	}
	return "↻"
}

fn stopHint(t: any) -> string {
	let s = str(t["state"])
	if s == "running"  or  s == "paused" {
		return "Stop this file: the URL can be changed again"
	}
	if s == "done" {
		return "Close this file: the URL can be changed again"
	}
	return "Start the download again with the current URL"
}

# Play or pause the target of the row.
async fn onTargetPlay(ev: any, id: string) -> null {
	let r = targetById(id)
	let s = str(r["state"])
	if str(r["fileId"]) == "" {
		startTarget(id)
		return
	}
	if s == "running" {
		fileAction(str(r["fileId"]), "pause")
		return
	}
	if s == "paused" {
		fileAction(str(r["fileId"]), "resume")
		return
	}
	if s == "stopped" {
		restartTarget(id)
		return
	}
}

# Stop the target of the row, or start it again from its first byte. A file
# that is running, paused or complete is stopped: the URL becomes editable
# again and the operator can start the download once more.
async fn onTargetStop(ev: any, id: string) -> null {
	let r = targetById(id)
	let s = str(r["state"])
	if str(r["fileId"]) == "" {
		return
	}
	if s == "running"  or  s == "paused"  or  s == "done" {
		fileAction(str(r["fileId"]), "stop")
		return
	}
	restartTarget(id)
}

# Cancel: the file leaves the download and its row goes away.
async fn onTargetCancel(ev: any, id: string) -> null {
	let r = targetById(id)
	let fileId = str(r["fileId"])
	if fileId == "" {
		removeTarget(id)
		return
	}
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/files/" + fileId + "/cancel"), "", "/api http://127.0.0.1:8090/api")
	if err != null {
		notifyError("the cancel failed: " + err)
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		notifyError("invalid response: " + jerr)
		return
	}
	if has(job, "error") and str(job["error"]) != "" {
		notifyError("the cancel failed: " + str(job["error"]))
		return
	}
	# The file is not in the download any more: the row goes away with it.
	removeTarget(id)
	updateJob(job)
}

# The id of a row that is still empty: the row of the first file of the
# download takes its place, so the operator does not find a useless empty row
# after a reload.
fn freeTargetId() -> string {
	let i = 0
	while i < len(targets) {
		let r = targets[i]
		if str(r["fileId"]) == ""  and  str(r["url"]) == "" {
			return str(r["id"])
		}
		i = i + 1
	}
	return ""
}

# The rows follow the files of the download: the state and the progress of
# every row come from the backend. A file that has no row yet (the app was
# started again) gets one.
fn syncTargets(jobFiles: any) -> null {
	let i = 0
	while i < len(jobFiles) {
		let f = jobFiles[i]
		let fileId = str(f["id"])
		let row = targetByFileId(fileId)
		let id = str(row["id"])
		if id == "" {
			let free = freeTargetId()
			if free != "" {
				id = free
			} else {
				let fresh = newTarget()
				targets = append(targets, fresh)
				id = str(fresh["id"])
			}
			setTarget(id, "fileId", fileId)
			setTarget(id, "url", str(f["url"]))
			setTarget(id, "name", str(f["file"]))
			setTarget(id, "named", "yes")
			setTarget(id, "priority", str(f["priority"]))
		}
		setTarget(id, "state", str(f["state"]))
		setTarget(id, "percent", f["percent"])
		setTarget(id, "received", str(f["receivedText"]))
		setTarget(id, "total", str(f["totalText"]))
		setTarget(id, "merged", str(f["merged"]))
		i = i + 1
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
	# The answer can come from the same origin or from the address of the
	# backend: when the page is served by another host the backend has its own
	# address, and it can also start after the page.
	jobsList = null
	body, err = http.getSafe(apiUrl("/downloads"), "/api http://127.0.0.1:8090/api")
	if err == null {
		parsed, jerr = json.parse(body)
		if jerr == null {
			jobsList = parsed["jobs"]
		}
	}
	if jobsList == null and apiRoot == "" {
		apiRoot = apiFallback
		body2, err2 = http.getSafe(apiUrl("/downloads"), "/api http://127.0.0.1:8090/api")
		if err2 == null {
			parsed2, jerr2 = json.parse(body2)
			if jerr2 == null {
				jobsList = parsed2["jobs"]
			}
		}
		if jobsList == null {
			apiRoot = ""
		}
	}
	if jobsList == null {
		notifyError(backendDown())
		return
	}
	let jobs = jobsList
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
	# The length of the parts the backend really uses: it grows the parts when
	# the file does not fit in the parts the CLI allows.
	jobPartSize = str(job["partSizeText"])
	# The server does not always announce the size of the file: without it there
	# is no percentage and no time left, only the bytes that have arrived.
	sizeUnknown = str(job["total"]) == "0"
	if sizeUnknown and jobState == "running" {
		message = "the server does not announce the size: the download goes on until it ends"
	}
	let parts = job["parts"]
	# Only the parts that have started get a bar: a pending part would show
	# an empty bar. The bars start from the most recent part: what is
	# happening now is at the top of the list. The table keeps the order of
	# the parts.
	partsView = reverse(filter(parts, isStarted))
	partsTable = parts
	# The counters of the parts: how many there are, how many are downloading
	# and how many are done.
	partTotal = len(parts)
	partActive = countParts(parts, "downloading")
	partDone = countParts(parts, "done")
	# The files of the download: the table shows them and the rows of the
	# targets follow them.
	let jobFiles = job["files"]
	files = jobFiles
	filesTable = fileTable(jobFiles)
	refreshPartsTable(jobId + "|" + jobState + "|" + str(failedParts) + "|" + str(partDone))
	refreshFilesTable(jobId + "|" + filesStamp(jobFiles))
	syncTargets(jobFiles)
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

# One row of the files table: the id comes first because the priority of a
# file is changed with it.
fn fileRow(f: any) -> array[string] {
	return [str(f["id"]), str(f["file"]), str(f["priority"]), str(f["state"]), str(f["partCount"]), str(f["totalText"])]
}

# The rows of the files table. The rows are arrays and not objects: the table
# then keeps the columns of the design.
fn fileRows(jobFiles: any) -> array[any] {
	return map(jobFiles, fileRow)
}

# The payload of the files table: the columns come with the rows, so the table
# shows the columns of the design and not the keys of the JSON.
fn fileTable(jobFiles: any) -> any {
	return {"columns": ["Id", "File", "Priority", "State", "Parts", "Size"], "rows": fileRows(jobFiles)}
}

# The stamp of the files: it changes when a file is added, when its priority
# changes or when it reaches another state. The progress of a file is not part
# of it, so the table keeps the row the operator selected at every poll.
fn filesStamp(jobFiles: any) -> string {
	let stamp = ""
	let i = 0
	while i < len(jobFiles) {
		let f = jobFiles[i]
		stamp = stamp + str(f["id"]) + ":" + str(f["priority"]) + ":" + str(f["state"]) + ";"
		i = i + 1
	}
	return stamp
}

# The one button of the header follows the phase of the download: it pauses
# it or resumes it, and it always says what it is going to do.
fn syncPhase() -> null {
	if jobState == "running" {
		phase = "running"
		return
	}
	if jobState == "paused" {
		phase = "paused"
		return
	}
	phase = "idle"
}

fn phaseIcon() -> string {
	if phase == "running" {
		return "❚❚"
	}
	return "▶"
}

fn phaseHint() -> string {
	if phase == "running" {
		return "Pause every download"
	}
	if phase == "paused" {
		return "Resume every download"
	}
	return "Pause or resume every download"
}

# The URL of a target can be changed before its download starts and after it
# has been stopped: a running file keeps the URL it started with.
fn targetLocked(t: any) -> bool {
	let s = str(t["state"])
	if s == "new" {
		return false
	}
	if s == "stopped" {
		return false
	}
	return true
}

# A part shows its bar only when its download has started: a part that is
# still waiting has nothing to show. A part that was interrupted keeps the
# bytes already on disk, so it keeps its bar too.
fn isStarted(part: any) -> bool {
	if str(part["State"]) == "pending" {
		return str(part["Received"]) != "0"
	}
	return true
}

# How many parts are in a given state: the table is refreshed when this
# changes, so the row the operator selected is not lost at every poll.
fn countParts(parts: any, want: string) -> int {
	let n = 0
	let i = 0
	while i < len(parts) {
		if str(parts[i]["State"]) == want {
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

# The files table is filled only when its content really changes: a table
# refresh drops the selected row.
fn refreshFilesTable(signature: string) -> null {
	if signature == filesSignature {
		return
	}
	filesSignature = signature
	n, terr = table.display("files", filesTable)
	if terr != null {
		notifyError("files table not updated: " + terr)
	}
}

# The download is complete: a toast and a short sound.
async fn announceDone(job: any) -> null {
	toast.show("Download complete: " + str(len(job["files"])) + " files", "success", 4000)
	ok, err = sound.playSafe(backendUrl("/sound.wav"), "/sound.wav http://127.0.0.1:8090/sound.wav")
	if err != null {
		message = "download complete (the browser blocked the sound)"
	}
}

# The button of the header: it pauses the whole download or resumes it.
async fn onAction() -> null {
	if jobId == "" {
		return
	}
	if phase == "running" {
		runAction("pause")
		return
	}
	if phase == "paused" {
		runAction("resume")
	}
}

# Start the target of a row: the file joins the download. The first file of a
# download starts it, the other ones join the queue of its parts.
async fn startTarget(id: string) -> null {
	let r = targetById(id)
	let u = trim(str(r["url"]))
	if u == "" {
		notifyError("type the URL of the file to download")
		return
	}
	let name = str(r["name"])
	let prio = str(r["priority"])
	if jobId == "" {
		# There is no download yet: the target becomes its first file.
		payload, jerr = json.stringify({"url": u, "partSize": partSize, "workers": workers, "single": single, "file": name, "priority": prio})
		if jerr != null {
			notifyError("invalid request: " + jerr)
			return
		}
		body, err = http.postSafe(apiUrl("/downloads"), payload, "/api http://127.0.0.1:8090/api")
		if err != null {
			notifyError("could not start: " + err)
			return
		}
		job, perr = json.parse(body)
		if perr != null {
			notifyError("invalid response: " + perr)
			return
		}
		# A refused request (a bad URL, too many parts, a name the backend does
		# not accept) answers with an error and no job: the operator reads why.
		if has(job, "error") and str(job["error"]) != "" {
			notifyError("could not start: " + str(job["error"]))
			return
		}
		bindTargetFile(id, job)
		selectedPart = ""
		updateJob(job)
		return
	}
	payload, jerr = json.stringify({"url": u, "file": name, "priority": prio})
	if jerr != null {
		notifyError("invalid request: " + jerr)
		return
	}
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/files"), payload, "/api http://127.0.0.1:8090/api")
	if err != null {
		notifyError("could not add the file: " + err)
		return
	}
	job, perr = json.parse(body)
	if perr != null {
		notifyError("invalid response: " + perr)
		return
	}
	if has(job, "error") and str(job["error"]) != "" {
		notifyError("could not add the file: " + str(job["error"]))
		return
	}
	bindTargetFile(id, job)
	updateJob(job)
}

# The row takes the file the backend created for it: the last file of the
# download is the one that was just added.
fn bindTargetFile(id: string, job: any) -> null {
	let jobFiles = job["files"]
	if len(jobFiles) == 0 {
		return
	}
	let f = jobFiles[len(jobFiles) - 1]
	setTarget(id, "fileId", str(f["id"]))
	setTarget(id, "state", str(f["state"]))
}

# Start a stopped target again: the file is downloaded again from its first
# byte, with the URL the operator has now.
async fn restartTarget(id: string) -> null {
	let r = targetById(id)
	let fileId = str(r["fileId"])
	if fileId == "" {
		return
	}
	payload, jerr = json.stringify({"url": trim(str(r["url"])), "file": str(r["name"])})
	if jerr != null {
		notifyError("invalid request: " + jerr)
		return
	}
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/files/" + fileId + "/restart"), payload, "/api http://127.0.0.1:8090/api")
	if err != null {
		notifyError("could not start the file again: " + err)
		return
	}
	job, perr = json.parse(body)
	if perr != null {
		notifyError("invalid response: " + perr)
		return
	}
	if has(job, "error") and str(job["error"]) != "" {
		notifyError("could not start the file again: " + str(job["error"]))
		return
	}
	updateJob(job)
}

# The action of one file of the download: pause, resume, stop, restart or
# cancel. The other files of the download keep going.
async fn fileAction(fileId: string, act: string) -> null {
	if jobId == "" {
		notifyError("no download to manage")
		return
	}
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/files/" + fileId + "/" + act), "", "/api http://127.0.0.1:8090/api")
	if err != null {
		notifyError("the " + act + " failed: " + err)
		return
	}
	job, jerr = json.parse(body)
	if jerr != null {
		notifyError("invalid response: " + jerr)
		return
	}
	if has(job, "error") and str(job["error"]) != "" {
		notifyError("the " + act + " failed: " + str(job["error"]))
		return
	}
	updateJob(job)
}

# The priority of one file of the download: the scheduler takes the parts of
# the files with the highest priority first.
async fn setFilePriority(fileId: string, want: string) -> null {
	if jobId == "" {
		return
	}
	payload, jerr = json.stringify({"priority": want})
	if jerr != null {
		notifyError("invalid request: " + jerr)
		return
	}
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/files/" + fileId + "/priority"), payload, "/api http://127.0.0.1:8090/api")
	if err != null {
		notifyError("could not change the priority: " + err)
		return
	}
	job, perr = json.parse(body)
	if perr != null {
		notifyError("invalid response: " + perr)
		return
	}
	if has(job, "error") and str(job["error"]) != "" {
		notifyError("could not change the priority: " + str(job["error"]))
		return
	}
	updateJob(job)
}

# The URL contains the job id, so it is dynamic: the *Safe variant with a
# literal allow-list is the one the security rules require.
async fn runAction(act: string) -> null {
	if jobId == "" {
		notifyError("no download to manage")
		return
	}
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/" + act), "", "/api http://127.0.0.1:8090/api")
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
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/parts/" + selectedPart + "/retry"), "", "/api http://127.0.0.1:8090/api")
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
	body, err = http.postSafe(apiUrl("/downloads/" + jobId + "/retry-failed"), "", "/api http://127.0.0.1:8090/api")
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
	# The first row of the targets is there from the start: the operator has a
	# URL to type in even if the backend never answers.
	ensureFirstTarget()
	interval(500, refresh)
	loadConfig()
}

# Fill the parts table again on demand.
fn refreshData() -> null {
	partsSignature = ""
	refreshPartsTable("manual")
	filesSignature = ""
	refreshFilesTable("manual")
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
			h1 "Download manager"
			row class: "topbar-actions" {
				button "{phaseIcon()}" onClick: onAction class: "btn-icon btn-icon-sm" tooltip: "{phaseHint()}"
				button "⚙" onClick: openSettings class: "btn-icon btn-icon-sm" tooltip: "Download settings"
			}
		}
	}

	card class: "panel panel-targets" {
		row class: "targets-head" {
			h3 "Targets"
			# The plus button adds a row: every file that is going to be
			# downloaded, or that is being downloaded, has one.
			button "+" onClick: addTarget class: "btn-icon btn-add btn-icon-sm" tooltip: "Add a target"
		}
		# The list of the targets is always rendered: a target that does not
		# exist yet simply has no row. The hint tells the operator what to do
		# when there is no row at all. The list never lives inside a branch of
		# a conditional block: the rows would disappear with the branch.
		if len(targets) == 0 {
			p "Add a target: paste the URL (http, https or ftp) and press play." class: "muted"
		}
		list t in targets key: t["id"] {
			row class: "target-row" {
				# The URL of a target cannot be changed once its download
				# has started: the stop button frees it.
				if targetLocked(t) {
					p "{t["url"]}" class: "target-url target-locked" tooltip: "The URL cannot be changed while the file is downloading"
				} else {
					input "URL" value: t["url"] placeholder: "https://example.com/file.bin" onInput: onTargetUrl class: "target-url"
				}
				input "Save as" value: t["name"] placeholder: "download.bin" onInput: onTargetName class: "target-name"
				button "{prioIcon(t["priority"])}" onClick: onTargetPriority class: "btn-icon btn-prio" tooltip: "{prioLabel(t["priority"])}"
				button "{playIcon(t)}" onClick: onTargetPlay class: "btn-icon btn-play" tooltip: "{playHint(t)}"
				button "{stopIcon(t)}" onClick: onTargetStop class: "btn-icon btn-stop" tooltip: "{stopHint(t)}"
				button "✕" onClick: onTargetCancel class: "btn-icon btn-cancel" tooltip: "Cancel this target"
			}
			row class: "target-foot" {
				progress value: t["percent"] max: 100 class: "bar-target"
				p "{t["received"]} / {t["total"]}" class: "target-bytes"
			}
		}
	}

	tabs "Download" {
		tab "Progress" {
			card class: "panel" {
				h3 "Total progress"
				progress value: percent max: 100 class: "bar-total"
				p "{verifyText}" class: "verify"
				if jobState == "done" {
					a "Download the reassembled file" href: apiUrl("/downloads/" + jobId + "/file") class: "dl-link"
				}
				# The parts are a child of the total progress: the bars of the
				# parts belong to the progress of the download.
				if jobSingle == false {
					accordion "Parts" {
						accordionitem "{partsTitle()}" {
							# The counters of the parts: how many there are, how many
							# are downloading and how many are done.
							row class: "part-counts" {
								col class: "part-count" {
									p "{partTotal}" class: "count-value"
									p "parts" class: "count-label"
								}
								col class: "part-count" {
									p "{partActive}" class: "count-value count-active"
									p "in download" class: "count-label"
								}
								col class: "part-count" {
									p "{partDone}" class: "count-value count-done"
									p "completed" class: "count-label"
								}
							}
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
		}

		tab "Files" {
			card class: "panel" {
				h3 "Files"
				if len(files) == 0 {
					p "No download yet: the files of the download appear here." class: "muted"
				} else {
					table name: "files" {
						thead {
							trow {
								tcell "Id" head: true
								tcell "File" head: true
								tcell "Priority" head: true
								tcell "State" head: true
								tcell "Parts" head: true
								tcell "Size" head: true
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
					# The link to the file that has been reassembled.
					div class: "files-list" {
						list f in files {
							row class: "file-row" {
								p "{f["file"]}" class: "file-name"
								if f["merged"] != "" {
									a "Download" href: apiUrl("/downloads/" + jobId + "/files/" + str(f["id"]) + "/file") class: "dl-file"
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
							tcell "File" head: true
							tcell "Priority" head: true
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
		row class: "statusbar-status" {
			badge "{jobState}" class: "state-badge"
			p "{message}" class: "status"
		}
		row class: "statusbar-row" {
			col class: "sb-item" {
				if sizeUnknown {
					p "--" class: "sb-value"
				} else {
					p "{percent}%" class: "sb-value"
				}
				p "progress" class: "sb-label"
			}
			col class: "sb-item" {
				if sizeUnknown {
					p "--" class: "sb-value"
				} else {
					p "{totalText}" class: "sb-value"
				}
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

css "body \{ margin: 0; padding: 0 0 150px; background: radial-gradient(1100px 520px at 12% -8%, rgba(99, 102, 241, 0.22) 0%, rgba(15, 23, 42, 0) 60%), linear-gradient(165deg, #0b1020 0%, #131a30 55%, #0e1426 100%); color: #e6e9f5; font-family: system-ui, sans-serif; } h1 \{ font-size: 26px; margin: 0 0 6px; letter-spacing: -0.02em; } h3 \{ margin: 0 0 14px; font-size: 15px; text-transform: uppercase; letter-spacing: 0.08em; color: #a5b4fc; } p \{ line-height: 1.55; } card \{ background: rgba(255, 255, 255, 0.04); border: 1px solid rgba(148, 163, 184, 0.16); border-radius: 18px; padding: 20px 22px; margin: 16px 0; box-shadow: 0 18px 40px rgba(2, 6, 23, 0.45); } .topbar \{ background: rgba(15, 23, 42, 0.55); border: 1px solid rgba(148, 163, 184, 0.14); border-radius: 14px; padding: 10px 16px; margin: 18px 0 6px; box-shadow: none; } .topbar-row \{ align-items: center; gap: 14px; justify-content: space-between; } .topbar-actions \{ gap: 8px; align-items: center; } .targets-head \{ display: flex; align-items: center; justify-content: space-between; gap: 12px; } .targets-head h3 \{ margin: 0; } .panel-targets \{ padding: 14px 18px; } .target-row \{ display: flex; align-items: center; gap: 8px; padding: 6px 0; border-top: 1px solid rgba(148, 163, 184, 0.10); } .target-url \{ flex: 1 1 auto; min-width: 0; } .target-url.target-locked \{ overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; color: #8b93b0; padding: 6px 2px; } .target-name \{ flex: 0 1 190px; min-width: 0; } .target-foot \{ display: flex; align-items: center; gap: 10px; padding-bottom: 8px; } .target-foot progress.sui-progress \{ flex: 1 1 auto; --sui-progress-height: 4px; --sui-progress-radius: 999px; } .target-bytes \{ margin: 0; font-size: 11px; color: #7c86a8; font-variant-numeric: tabular-nums; white-space: nowrap; } .sui-button.btn-prio \{ background: rgba(255, 255, 255, 0.05); color: #a5b4fc; border: 1px solid rgba(129, 140, 248, 0.30); width: 32px; height: 32px; padding: 0; border-radius: 9px; font-size: 14px; } .sui-button.btn-prio:hover \{ background: rgba(99, 102, 241, 0.22); color: #e0e7ff; } .sui-button.btn-play \{ background: linear-gradient(135deg, #6366f1, #22d3ee); color: #061021; border: 0; width: 34px; height: 34px; padding: 0; border-radius: 10px; font-size: 12px; } .sui-button.btn-play:hover \{ box-shadow: 0 6px 16px rgba(99, 102, 241, 0.45); } .sui-button.btn-stop \{ background: rgba(255, 255, 255, 0.06); color: #c7d2fe; border: 1px solid rgba(148, 163, 184, 0.30); width: 34px; height: 34px; padding: 0; border-radius: 10px; font-size: 12px; } .sui-button.btn-stop:hover \{ background: rgba(148, 163, 184, 0.20); color: #e0e7ff; } .sui-button.btn-icon \{ background: rgba(255, 255, 255, 0.06); color: #c7d2fe; border: 1px solid rgba(129, 140, 248, 0.35); width: 42px; height: 42px; padding: 0; border-radius: 12px; font-size: 18px; line-height: 1; display: inline-flex; align-items: center; justify-content: center; transition: transform 140ms ease, box-shadow 140ms ease, color 140ms ease, background 140ms ease; } .sui-button.btn-icon:hover \{ background: rgba(99, 102, 241, 0.22); color: #e0e7ff; transform: translateY(-1px); box-shadow: 0 0 0 4px rgba(129, 140, 248, 0.16), 0 10px 22px rgba(99, 102, 241, 0.35); } .sui-button.btn-action \{ background: linear-gradient(135deg, #6366f1 0%, #22d3ee 100%); color: #061021; border: 0; font-size: 19px; font-weight: 700; box-shadow: 0 10px 24px rgba(99, 102, 241, 0.35); } .sui-button.btn-action:hover \{ color: #061021; background: linear-gradient(135deg, #818cf8 0%, #67e8f9 100%); box-shadow: 0 0 0 5px rgba(250, 204, 21, 0.22), 0 14px 30px rgba(99, 102, 241, 0.45); } .sui-button.btn-add \{ background: rgba(34, 211, 238, 0.16); color: #67e8f9; border: 1px solid rgba(34, 211, 238, 0.42); font-size: 22px; font-weight: 700; } .sui-button.btn-add:hover \{ background: rgba(34, 211, 238, 0.28); color: #cffafe; box-shadow: 0 0 0 4px rgba(34, 211, 238, 0.18); } .sui-button.btn-cancel \{ background: rgba(244, 63, 94, 0.14); color: #fda4af; border: 1px solid rgba(244, 63, 94, 0.38); } .sui-button.btn-cancel:hover \{ background: rgba(244, 63, 94, 0.26); color: #ffe4e6; box-shadow: 0 0 0 4px rgba(244, 63, 94, 0.18); } .sui-button.btn-icon-sm \{ width: 32px; height: 32px; font-size: 15px; border-radius: 10px; } .status \{ color: #a8b1cc; font-size: 13px; margin: 0; } .state-badge \{ background: rgba(148, 163, 184, 0.16); color: #cbd5e1; border: 1px solid rgba(148, 163, 184, 0.30); } .verify \{ margin: 14px 0 0; color: #6ee7b7; font-size: 13px; } .dl-link \{ display: inline-block; margin-top: 12px; color: #67e8f9; font-weight: 600; } .dl-file \{ color: #67e8f9; font-weight: 600; font-size: 12px; white-space: nowrap; } .parts-scroll \{ max-height: 320px; overflow-y: auto; overflow-x: hidden; padding: 2px 8px 2px 2px; margin: 4px 0 10px; scrollbar-width: thin; scrollbar-color: rgba(129, 140, 248, 0.45) rgba(15, 23, 42, 0.4); } .parts-scroll::-webkit-scrollbar \{ width: 10px; } .parts-scroll::-webkit-scrollbar-track \{ background: rgba(15, 23, 42, 0.45); border-radius: 8px; } .parts-scroll::-webkit-scrollbar-thumb \{ background: rgba(129, 140, 248, 0.45); border-radius: 8px; border: 2px solid transparent; background-clip: padding-box; } .row.part-row \{ display: grid; grid-template-columns: 64px minmax(120px, 1fr) 92px 148px; align-items: center; gap: 14px; padding: 8px 10px; border-radius: 10px; } .row.part-row:hover \{ background: rgba(148, 163, 184, 0.08); } .part-index \{ font-weight: 600; color: #c7d2fe; margin: 0; font-variant-numeric: tabular-nums; } .part-state \{ font-size: 12px; color: #8b93b0; margin: 0; } .part-bytes \{ font-size: 12px; color: #8b93b0; text-align: right; margin: 0; font-variant-numeric: tabular-nums; } .part-error \{ grid-column: 1 / -1; font-size: 12px; color: #fda4af; margin: 0; } progress.bar-total \{ --sui-progress-height: 12px; --sui-progress-radius: 999px; --sui-progress-track: rgba(148, 163, 184, 0.16); --sui-progress-value: #6366f1; --sui-progress-value-end: #22d3ee; } progress.bar-part \{ --sui-progress-height: 8px; --sui-progress-radius: 999px; --sui-progress-track: rgba(148, 163, 184, 0.16); --sui-progress-value: #818cf8; --sui-progress-value-end: #38bdf8; } progress.settings \{ display: flex; flex-direction: column; gap: 14px; min-width: 320px; } .setting \{ gap: 16px; align-items: flex-start; padding-bottom: 12px; border-bottom: 1px solid rgba(148, 163, 184, 0.14); } .setting:last-child \{ border-bottom: 0; padding-bottom: 0; } .setting-text \{ gap: 2px; flex: 1 1 auto; } .setting-text p \{ margin: 0; font-size: 12px; color: #8b93b0; line-height: 1.45; } .setting-control \{ flex: 0 0 220px; } .mode-switch \{ flex: 0 0 auto; } .hint \{ color: #8b93b0; font-size: 13px; margin: 2px 0 10px; } .muted \{ color: #8b93b0; font-size: 13px; } .controls \{ gap: 10px; flex-wrap: wrap; margin-bottom: 14px; } .sui-button.btn-warn \{ background: rgba(251, 191, 36, 0.16); color: #fcd34d; border: 1px solid rgba(251, 191, 36, 0.42); padding: 9px 16px; border-radius: 12px; } .sui-button.btn-ghost \{ background: transparent; color: #94a3b8; border: 1px solid rgba(148, 163, 184, 0.28); padding: 9px 16px; border-radius: 12px; } .toast-host .sui-toast \{ width: 340px; } .sui-tabs \{ --sui-tabs-border: rgba(148, 163, 184, 0.16); --sui-tabs-radius: 16px; --sui-tabs-bg: rgba(255, 255, 255, 0.02); --sui-tablist-bg: rgba(15, 23, 42, 0.55); --sui-tab-padding: 10px 18px; --sui-tab-active-bg: rgba(99, 102, 241, 0.18); --sui-tab-accent: #22d3ee; --sui-tab-fg: #94a3b8; --sui-tab-active-fg: #e0e7ff; } .sui-tabpanels \{ padding: 18px 20px; } .statusbar \{ position: fixed; left: 0; right: 0; bottom: 0; z-index: 5; margin: 0; padding: 12px 22px; background: rgba(11, 16, 32, 0.94); border-top: 1px solid rgba(148, 163, 184, 0.18); backdrop-filter: blur(8px); } .statusbar-status \{ gap: 12px; align-items: center; margin-bottom: 10px; } .statusbar-row \{ gap: 18px; flex-wrap: wrap; } .sb-item \{ flex: 1 1 120px; gap: 2px; } .sb-value \{ margin: 0; font-size: 17px; font-weight: 700; color: #e0e7ff; font-variant-numeric: tabular-nums; } .sb-label \{ margin: 0; font-size: 11px; text-transform: uppercase; letter-spacing: 0.08em; color: #7c86a8; } .field-file \{ margin: 0 0 8px; } .part-counts \{ gap: 16px; margin: 0 0 14px; flex-wrap: wrap; } .part-count \{ flex: 0 0 auto; gap: 2px; min-width: 76px; } .count-value \{ margin: 0; font-size: 20px; font-weight: 700; color: #e0e7ff; font-variant-numeric: tabular-nums; } .count-label \{ margin: 0; font-size: 11px; text-transform: uppercase; letter-spacing: 0.08em; color: #7c86a8; } .count-active \{ color: #38bdf8; } .count-done \{ color: #34d399; } .files-list \{ margin: 12px 0 6px; } .row.file-row \{ display: flex; align-items: center; justify-content: space-between; gap: 14px; padding: 8px 10px; border-radius: 10px; } .row.file-row:hover \{ background: rgba(148, 163, 184, 0.08); } .file-name \{ margin: 0; font-weight: 600; color: #e0e7ff; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; } .file-bytes \{ margin: 0; font-size: 12px; color: #8b93b0; text-align: right; font-variant-numeric: tabular-nums; } @media (max-width: 760px) { .row.part-row \{ grid-template-columns: 56px 1fr; row-gap: 6px; } .row.file-row \{ grid-template-columns: 1fr 74px; row-gap: 6px; } .parts-scroll \{ max-height: 240px; } .setting \{ flex-direction: column; gap: 8px; } .setting-control \{ flex: 1 1 100%; width: 100%; } }"
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
- A job is not persisted between restarts of the backend: the example keeps
  the job state in memory and the parts on disk.
- The parts of a file are deleted when the file is reassembled: the folder of
  the job keeps only the `merged-*` files. The merged file of a file is named
  `merged-<file id>-<name>`, so two files with the same name do not collide;
  the name the browser saves it with is the one the operator chose.
- The files of a download share the queue of the parts and the workers of the
  job: the download is one job with many files, not one job per file. A pause,
  a resume or a cancel acts on all of them. The scheduler takes the parts of
  the files with the highest priority first and then walks the files of the
  same priority round-robin.
- The parts table is refilled only when its content really changes:
  `table.display` drops the selection, so a refresh at every poll would lose
  the row the operator selected for the retry. The same goes for the table of
  the files: its stamp carries the id, the priority and the state of every
  file, not the progress, so the selected row survives the poll.
- The `value:` of an `input` is write-only: the values coming from the CLI are
  read from the state and shown in another element of the view.
- The name of the file to save follows the URL with a short timer
  (`interval(200, syncFileName)`): the view is re-rendered only when the name
  really changes. The box keeps the name the operator chose: the program stops
  filling it as soon as it sees a name that is not the one it filled.
- If the server does not support `Range`, the file is downloaded as a single
  part: the split follows what the server can do, not what the operator asked.
- In the one-piece mode the total size is unknown until the server answers, so
  the part stays open and the request uses `Range: bytes=N-`: the backend
  accepts a plain `200` answer too.
- The completion sound comes from `GET /sound.wav`, a chime generated by the
  backend. The browser can refuse to play it (autoplay policy): the app shows
  the error in the status line, the download is still complete.
- The toast host is at the top right of the window: the errors of the app and
  of the backend are shown there and in the status bar, never only in the
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
| Action button | the header button pauses and resumes the whole download; the button of a row acts on its own file only |
| Target rows | one row per file: the URL, the name, the priority icon and the buttons of that file, all on one line |
| The first row | the URL field is there when the app starts, even when the backend does not answer |
| The + button | it adds a row for another file: every file that is added has its own row |
| Locked URL | the URL of a row is an editable box only before the download starts and after the file has been stopped |
| Stop | the file becomes `stopped`, the URL becomes editable again and the operator can start the file with another URL |
| Priority of a file | the icon of the row cycles from low to normal to high and changes the priority of that file only |
| Name of the file to save | the box next to the URL is filled from the URL (`sample.bin`), a name typed by the operator is kept and the download uses it |
| Parts | the section of the parts is inside the card of the total progress |
| Counters of the parts | the section of the parts opens with the total number of parts, the ones in download and the ones completed |
| Shared queue | the workers of the job are shared between the files: with two files and the same priority the parts are taken three by three |
| High priority | the next part the scheduler takes is the one of the file with the `high` priority, even if the other files were waiting |
| Files tab | the table shows the id, the name, the priority, the state, the parts and the size of every file; under it one bar per file |
| Merged files | every file has its own endpoint and its own link; the file is served with the name the operator chose |
| Parts deleted | after the merge the folder of the job holds only the `merged-*` files: no `part-*.bin` is left on the disk |
| Bars | only the parts whose download has started get a bar, the waiting ones show nothing |
| Status bar | the state and the message are in the first row, the data of the download in the row below |
| Errors | the failures are shown as error toasts, the same message is not repeated |
| Completion | the toast "Download complete" is shown and the chime of `/sound.wav` is played |

The automated test in `internal/core/download_manager_test.go` drives the same
cycle headlessly: it builds the backend, splits a 1 MiB file into two 512 KiB
parts, pauses, checks that no byte advances, resumes and compares the
reassembled file with the original one. A second test uses an origin that
fails one range once and checks that the retry of that single part brings the
job to `done`. Two more tests cover the name of the file to save: the name the
operator chose is the name of the job, a name that tries to escape the folder
of the job is cleaned, and the reassembled file is served with that name.

Two more tests cover the download with many files: one checks that the
priority of a file is changed with its own endpoint (and that an unknown
priority becomes the normal one), the other uses an origin that holds every
request: it checks that the workers of the job are shared between the files,
that the file with the high priority is the one the scheduler takes first, and
that the folder of the job holds no `part-*.bin` file once the files are
reassembled.
