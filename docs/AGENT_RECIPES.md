# Agent recipes: local and remote interfaces

Operational patterns for coding agents. These are **recipes**, not a separate
API reference: canonical signatures live in
[`API_REFERENCE.md`](API_REFERENCE.md), [`STDLIB.md`](STDLIB.md) and
[`BUILTINS.md`](BUILTINS.md). Check every example with `simpleui check`.

## Workflow

1. Read [`AGENTS.md`](../AGENTS.md) and [`docs/manifest.json`](manifest.json).
2. Choose the closest recipe below and the closest file under `examples/`.
3. Write the smallest working `.sui` app, with explicit loading/error states.
4. Run `simpleui check app.sui --json`; fix all errors.
5. Build or serve, then verify the rendered states with a browser.
6. Re-run `go test -count=1 ./...` if compiler/runtime files changed.

Do not generate code from memory. If a function or attribute is not documented,
search the docs (`simpleui docs --search <term>`) or inspect the compiler.

## Common skeleton

A robust remote-data screen has three states: loading, error and data (including
empty). Keep them separate from the payload.

```sui
import { http, json } from "std"

state status = "loading"
state errorMsg = ""
state items = []

async fn refresh() -> null {
	status = "loading"
	errorMsg = ""
	body, err = http.get("https://api.example.com/items")
	if err != null {
		errorMsg = err
		status = "error"
		return
	}
	data, perr = json.parse(body)
	if perr != null {
		errorMsg = perr
		status = "error"
		return
	}
	n, terr = table.display("items", data)
	if terr != null {
		errorMsg = terr
		status = "error"
		return
	}
	status = "ok"
}

view main {
	title "Items"
	onLoad: refresh
	h1 "Items"
	if status == "error" {
		alert "{errorMsg}"
		button "Retry" onClick: refresh
	} else {
		table name: "items" {
			thead {
				trow {
					tcell "ID" head: true
					tcell "Name" head: true
				}
			}
			tbody {
				trow {
					tcell ""
					tcell ""
				}
			}
		}
	}
}
```

Notes:

- `table.display` replaces the body and preserves the table's style; keep a
  static model row so the table exists before data arrives.
- `json.parse` may return an object, array or scalar. Validate its shape before
  indexing or looping.
- Show the actual error string; otherwise the user cannot distinguish a network
  failure from malformed JSON.

## Drag and drop

Use a stable key per item. For a list of strings the string itself is a stable
key; for records use `item.id`.

```sui
state items = ["Design", "Build", "Ship"]
state attachment = ""

fn reorder() -> null {
	let src = indexOf(items, drag.source())
	let dst = indexOf(items, drag.target())
	if src < 0 or dst < 0 or src == dst {
		return
	}
	items = moveItem(items, src, dst)
}

view main {
	list item in items {
		row draggable: true dragKey: item dropKey: item onDrop: reorder {
			text "{item}"
		}
	}
	input "Attachment" value: attachment dropFiles: attachment
	button "Choose files…" pickFiles: attachment accept: ".csv,.txt" multiple: true
}
```

- `dragKey` + `dropKey` identify a list row as source and target; `onDrop`
  performs the actual reorder.
- `dropFiles: state` is the automatic file/folder target; pair it with
  `value: state` on an input or textarea to display the paths.
- `directory: true` on a `file` control or picker button opens a folder picker.
  The browser returns all files recursively and `drag.folder()` exposes the
  common folder name. For a subset, use the multi-file picker or app-side
  filtering.
- `pickFiles: state` opens the native dialog from a `file` control, `button` or
  other clickable element; `accept`, `multiple` and `directory` configure it.
- Use `drag.files()` in `onDrop`/after a picker selection for immediate
  metadata. Directory contents are collected asynchronously for `dropFiles`.
- Browsers do not expose absolute paths for security. Show the name/relative
  path returned by the browser; never assume it is a filesystem path.
- Treat names as untrusted text. Validate any URL built from a dropped name and
  keep desktop filesystem access behind `--allow-dir`.

See [`API_REFERENCE.md`](API_REFERENCE.md#drag-and-drop),
[`BUILTINS.md`](BUILTINS.md#drag-and-drop-functions) and the
[GUI editor guide](GUI_EDITOR.md) for the inspector steps and the difference
between designer and application drag and drop. The runnable example is
[`examples/drag_drop.sui`](../examples/drag_drop.sui) and
[`examples/file_picker.sui`](../examples/file_picker.sui).

### File type analysis with a Go backend

Use `http.postFiles` to send the current `pickFiles`/`dropFiles` selection as
multipart data, then render the backend's `{columns, rows}` response with
`table.display`:

```sui
import { http, json } from "std"

state status = "idle"
state errorMsg = ""
state files = ""

async fn analyze() -> null {
	status = "loading"
	errorMsg = ""
	body, err = http.postFiles("/api/analyze")
	if err != null {
		status = "error"
		errorMsg = err
		return
	}
	data, perr = json.parse(body)
	if perr != null {
		status = "error"
		errorMsg = perr
		return
	}
	status = "ok"
	n, terr = table.display("analysis", data)
	if terr != null {
		status = "error"
		errorMsg = terr
	}
}

view main {
	button "Choose files..." pickFiles: files multiple: true
	button "Analyze" onClick: analyze
	input "Selected" value: files dropFiles: files
	if status == "ok" {
		table name: "analysis" {
			thead { trow { tcell "File" head: true tcell "MIME" head: true } }
			tbody { trow { tcell "" tcell "" } }
		}
	}
}
```

The full companion server, upload limits, multipart analysis and table output
are in [`examples/file_analyzer/README.md`](../examples/file_analyzer/README.md).

## Local application

Local apps use the web runtime and do not need network capabilities. Use states
for form fields, in-memory collections and status. Optional persistence uses
`storage.*` (localStorage on web, memory fallback on desktop).

### Local form + reactive table

```sui
state name = ""
state items = []
state msg = ""

fn add() -> null {
	let n = trim(name)
	if n == "" {
		msg = "Enter a name"
		return
	}
	items = append(items, n)
	name = ""
	msg = ""
	nrows, err = table.display("items", items)
	if err != null { msg = err } else { msg = str(nrows) + " rows" }
}

view main {
	title "Local items"
	h1 "Local items"
	row {
		input value: name placeholder: "Name"
		button "Add" onClick: add
	}
	if msg != "" { p "{msg}" }
	table name: "items" {
		thead { trow { tcell "Name" head: true } }
		tbody { trow { tcell "" } }
	}
}
```

Use `let` for local values, `state` for values that must trigger a re-render.
For derived values, compute them in the view (`"{len(items)}"`) instead of
duplicating state.

## Remote API

Remote apps must decide between a **literal** URL and a **dynamic** URL:

```sui
# Fixed destination: literal URL is allowed.
body, err = http.get("https://api.example.com/items")

# Dynamic destination: *Safe + literal allow-list is required.
body, err = http.getSafe(endpoint, "https://api.example.com")
```

The allow-list is matched against scheme/host/port and an optional path prefix;
credentials and redirects are rejected. Do not build an allow-list from input.

### Remote JSON API + table

See [`examples/users/app.sui`](../examples/users/app.sui) for a complete
diagnostic-style example. Minimal shape:

```sui
import { http, json } from "std"

state status = "loading"
state errText = ""
state rows: array[array[string]] = []

async fn loadUsers() -> null {
	status = "loading"
	body, herr = http.get("/api/users")
	if herr != null {
		errText = herr
		status = "error"
		return
	}
	data, jerr = json.parse(body)
	if jerr != null {
		errText = jerr
		status = "error"
		return
	}
	n, terr = table.display("users", data)
	if terr != null {
		errText = terr
		status = "error"
		return
	}
	status = "ok"
}

view main {
	title "Users"
	onLoad: loadUsers
	h1 "Users"
	if status == "error" {
		alert "{errText}"
		button "Retry" onClick: loadUsers
	} else {
		table name: "users" {
			thead {
				trow {
					tcell "ID" head: true
					tcell "Name" head: true
					tcell "Email" head: true
				}
			}
			tbody {
				trow {
					tcell ""
					tcell ""
					tcell ""
				}
			}
		}
	}
}
```

An empty result is a valid `ok` state. Add an explicit empty message if the
table would otherwise look broken.

## CRUD form

### Remote CRUD form

Validate locally, send with `postSafe`, then reload or update the local model.
Never concatenate JSON strings.

```sui
import { http, json } from "std"

state name = ""
state status = ""
state errText = ""

async fn create() -> null {
	let n = trim(name)
	if n == "" {
		errText = "Name is required"
		return
	}
	payload, jerr = json.stringify({"name": n})
	body, herr = http.postSafe("https://api.example.com/items", payload, "https://api.example.com")
	if herr != null {
		errText = herr
		status = "error"
		return
	}
	name = ""
	status = "created"
}

view main {
	title "Create item"
	h1 "Create item"
	input value: name placeholder: "Name"
	button "Create" onClick: create
	if status == "error" { alert "{errText}" } else { p "{status}" }
}
```

For update/delete, check the HTTP status with `http.getStatusSafe` when the API
does not return a body. Treat every non-success status as an error.

## Polling dashboard

```sui
import { http, json } from "std"

state status = "loading"
state rows = []
state busy = false

async fn refresh() -> null {
	if busy { return }
	busy = true
	body, err = http.get("/api/metrics")
	if err != null {
		status = "error"
	} else {
		data, perr = json.parse(body)
		if perr != null {
			status = "error"
		} else {
			rows = data
			status = "ok"
		}
	}
	busy = false
}

fn boot() -> null {
	interval(2000, refresh)
}

view main {
	title "Metrics"
	onLoad: boot
	h1 "Metrics"
	p "Status: {status}"
	if status == "ok" {
		list m in rows { p "{m}" }
	} else {
		p "Unavailable"
	}
	button "Refresh" onClick: refresh
}
```

`interval` takes a function name, not a lambda. Keep polling intervals modest and
guard against overlapping requests. For real WebSockets use `ws.connectSafe`
with a literal `ws://`/`wss://` allow-list.

## Long job with a Go backend

A long running operation (a download, an import, a batch) belongs in a backend:
the SUI app is the control panel, not the engine. The backend owns the job
state; the app polls it and shows what happens.

```sui
import { http, json } from "std"

state jobId = ""
state jobState = "no download"
state percent = 0
state message = "Ready."

# Polling: the job id makes the status URL stable, but the per-part table is
# always refilled from the payload.
async fn refresh() -> null {
	if jobId == "" { return }
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
	if len(jobs) == 0 { return }
	let job = jobs[0]
	jobId = str(job["id"])
	jobState = str(job["state"])
	percent = job["percent"]
	n, terr = table.display("parts", job["parts"])
	if terr != null {
		message = "parts table not updated: " + terr
	}
}

# The action URL contains the job id, so it is dynamic: use the *Safe variant
# with a literal allow-list.
async fn pauseDownload() -> null {
	if jobId == "" { return }
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/pause", "", "http://127.0.0.1:8090")
	if err != null {
		message = "pause failed: " + err
		return
	}
	refresh()
}

fn boot() -> null {
	interval(500, refresh)
}
```

- Keep the job identifier in a state and build the action URLs from it; a
  dynamic URL requires `http.postSafe` with a **literal** allow-list.
- Show the per-unit progress (the parts, the steps, the items), not only a
  percentage: it is what tells the operator that the job is really moving.
- Count the units as well: how many there are, how many are running right now
  and how many are done. The numbers tell in one glance whether the job is
  moving forward or stuck.
- Let the operator choose the name of the file to save: fill the box from the
  URL, keep the name they type, and send it with the request. The backend
  cleans it (a file name is never a path) and uses it for the file it writes.
- Let a download hold more than one file. The files share the queue of the
  units and the workers of the job, so with six workers and two files the
  units are taken three by three. Give every file its own priority (low,
  normal, high) and take the units of the files with the highest priority
  first: the scheduler walks the files of the same priority round-robin.
- Delete the units once the file is reassembled: the parts of the disk are
  scratch space, the merged file is the result.
- The backend owns the truth. A pause, a resume or a cancel is a request to
  the backend, never a local flag: only the backend knows which units are
  already done.
- Report the backend errors as they are. A failed part is a real failure, not
  an empty row.
- Let a failed unit restart alone. The endpoint takes the unit number and
  requeues only that unit; the operator selects it in the table, so the table
  keeps the selection of the operator (a table refresh drops it).
- Keep the live view and the data apart: bars and counters for the progress,
  the tables in another tab, so the screen stays readable.
- One button for the whole download in the header, and **one row per file**
  below it: the row holds the URL, the name of the file, the priority as an
  icon and the buttons of that file (play/pause, stop/start again, cancel).
  The handlers declared inside a `list` are wired for **every** row and
  receive the key of the row they belong to, so every row acts on itself and
  never only the first one.
- Lock the URL of a file while it downloads. The stop button turns the URL
  back into an editable box: the operator writes the new URL and starts the
  file again, and the backend downloads it from the new URL. A file that is
  complete can be stopped too, so its URL can be changed in the same way.
- Show the errors as toasts (`toast.show(message, "error", 5000)`) and not only
  as a line of text: a failed request must be visible even when the operator
  is looking at the progress bars. Deduplicate the repeated message.
- Play a sound when the job ends with `sound.play(url)`: the URL is a literal,
  a dynamic one needs `sound.playSafe` with a literal allow-list. The browser
  may refuse to play it (autoplay policy): show the error, do not fail the job.
- Keep the settings in a collapsible overlay (`modal`): the mode, the size of
  the parts and the parallelism belong to the configuration, not to the main
  view.

See [`examples/download_manager/`](../examples/download_manager/README.md) for
the complete example: a Go backend that downloads the URL given on the command
line, splits the file into parts (or downloads it in one piece), fetches each
part with its own `Range` request, keeps the bytes of a paused part, restarts
a failed part on its own and reassembles the file with a size check.

## Desktop files and commands

Desktop capabilities are granted only at build time:

```bash
simpleui build app.sui --target desktop --out dist \
  --allow-dir C:/data --allow-exec helper.exe
```

The designer exposes the same flow through **File → Build desktop
executable…**. It preflights Go 1.24.2+, the source, the output directory and
the capabilities, shows the build log, and offers an optional executable
download.

```sui
import { fs, exec, json } from "std"

state path = "C:/data/input.json"
state status = ""

async fn read() -> null {
	body, err = fs.read(path)
	if err != null { status = err; return }
	data, jerr = json.parse(body)
	if jerr != null { status = jerr; return }
	n, terr = table.display("data", data)
	if terr != null { status = terr } else { status = str(n) + " rows" }
}

async fn runHelper() -> null {
	out, err = exec.run("helper.exe", [path])
	if err != null { status = err; return }
	status = out
}
```

The desktop build also exposes native pickers:

```sui
import { dialog } from "std"

async fn choose() -> null {
	path, err = dialog.pickFile()
	if err != null { status = err; return }
	if path == "" { status = "cancelled"; return }
	status = path
}
```

`dialog.pickFile()`, `dialog.pickFiles()` and `dialog.pickDirectory()` return
absolute paths on the desktop target; web builds return a clear error. The path
is not a capability: `fs` still needs `--allow-dir` and `exec` still needs
`--allow-exec`. The picker is user-driven; cancellation returns an empty value.
See [`examples/desktop_picker.sui`](../examples/desktop_picker.sui).

Validate user-supplied paths/names before calling `fs`/`exec`; the bridge
allow-list is a second layer, not a replacement for app validation. Always
provide a web fallback or a clear "desktop only" message.

## JSON documents

Use JSON Pointer to address one value, JSON Patch to change a copy and
`json.diff`/`json.merge3` for synchronisation workflows. JSON values are native
maps/arrays, so no DOM is involved.

```sui
import { json } from "std"

state status = ""

async fn update() -> null {
	let doc = {"items": [{"id": 1, "done": false}, {"id": 2, "done": true}]}
	first, perr = json.pointer(doc, "/items/0/id")
	if perr != null { status = perr; return }

	patched, err = json.patch(doc, [
		{"op": "replace", "path": "/items/0/done", "value": true},
		{"op": "add", "path": "/items/-", "value": {"id": 3, "done": false}},
		{"op": "test", "path": "/items/1/done", "value": true}
	])
	if err != null { status = err; return }

	errors, verr = json.validateAll(patched, {"type": "object", "required": ["items"]})
	if verr != null { status = verr; return }
	text, terr = json.stringify(patched, {"indent": 2, "sortKeys": true})
	status = str(first) + " / " + str(len(errors)) + " errors\n" + text
}

async fn syncFiles() -> null {
	local, lerr = json.loadFile("C:/data/local.json")
	if lerr != null { status = lerr; return }
	remote, rerr = json.loadFile("C:/data/remote.json")
	if rerr != null { status = rerr; return }
	conflicts, cerr = json.merge3(local, local, remote, {"key": "id"})
	if cerr != null { status = cerr; return }
	ok, serr = json.saveFile("C:/data/local.json", local, {"indent": 2})
	status = serr == null ? str(len(conflicts)) + " conflicts" : serr
}
```

`json.diff(a, b, {"key": "id"})` emits standard RFC 6902 operations (`add`,
`remove`, `replace`, `move`) and the result can be passed directly to
`json.patch`. `json.merge` is RFC 7386 Merge Patch; `json.merge3` mutates
`ours` and returns conflict strings. `json.loadFile`/`saveFile` and
`loadBytes`/`saveBytes` are desktop-only and require `--allow-dir`.
See [`examples/json_library.sui`](../examples/json_library.sui).

## Binary documents

PDF, images and archives must never travel as text: `http.get` decodes the body
as UTF-8 and corrupts the bytes. Three functions keep a binary payload intact,
one per source.

```sui
import { http, files, bytes } from "std"

state picked = ""
state size = 0
state magic = ""

async fn loadRemote() -> null {
	data, err = http.getBytes("https://example.com/report.pdf")
	if err != null { magic = err; return }
	size = len(data)
	head, herr = bytes.slice(data, 0, 4)
	if herr != null { magic = herr; return }
	hex, xerr = bytes.toHex(head)
	if xerr != null { magic = xerr; return }
	magic = hex
}

async fn loadPicked() -> null {
	data, err = files.readBytes(picked)
	if err != null { magic = err; return }
	size = len(data)
}

view main {
	row {
		button "Remote" onClick: loadRemote
		file accept: ".pdf" pickFiles: picked
	}
	p "bytes: {size} magic: {magic}"
}
```

- Remote: `http.getBytes(url)`, or `http.getBytesSafe(url, allow)` for a dynamic
  URL. Both reject a non-2xx status and a body over 10 MB.
- Folder the app may read (desktop): `fs.readBytes(path)`; the build-time
  `--allow-dir` capability still applies.
- File the user picked or dropped (web): `files.readBytes(name)`, where `name`
  matches the selected file name or relative path. It never reads a path.
- Validate the magic bytes before parsing, and treat every byte as untrusted
  data.

Runnable example: [`examples/binary_bytes.sui`](../examples/binary_bytes.sui).

## Spreadsheet data

SimpleUI does not parse binary XLSX/ODS in the browser. Convert upstream with
`sheet2json`:

```bash
sheet2json data.xlsx
sheet2json data.ods -sheet 2 -format csv
sheet2json data.xlsx -list
```

Then feed JSON/CSV into the table:

```sui
import { http } from "std"

state status = "loading"

async fn load() -> null {
	body, err = http.getSafe("/api/sheet?file=data.xlsx&format=json", "/api/")
	if err != null { status = err; return }
	n, terr = table.display("data", body)
	if terr != null { status = terr } else { status = str(n) + " rows" }
}
```

See [`examples/spreadsheet_table/README.md`](../examples/spreadsheet_table/README.md)
for the local converter service and `exec.run` variant.

## Iframe

```sui
# Static trusted endpoint. sandbox/referrerpolicy are added automatically.
iframe src: "https://example.com/embed"

# Dynamic endpoint: literal allow-list required.
iframe src: videoUrl srcAllow: "https://example.com" sandbox: "allow-scripts"
```

- Default `sandbox=""` blocks scripts, forms, popups and same-origin access.
  Add only the tokens the embedded page actually needs.
- `allow-scripts` + `allow-same-origin` is an error by default; do not weaken the
  policy unless the content is fully trusted.
- Malformed or non-allow-listed dynamic URLs are not loaded; the runtime shows
  `Cannot accept the data: ...`.
- Never use `srcdoc`; fallback markup is rejected.

See [`examples/security/untrusted_iframe.sui`](../examples/security/untrusted_iframe.sui).

## PDF viewer

```sui
import { fs } from "std"

state docUrl = ""

fn open() -> null {
	url, err = fs.documentUrl("docs/report.pdf")
	if err == null {
		docUrl = url
	}
}

view main {
	button "Open" onClick: open

	# A remote document served by your own backend.
	pdfviewer src: "https://example.com/manual.pdf"

	# A local file (desktop): the URL comes from the local bridge.
	pdfviewer src: docUrl srcAllow: "/" width: 900 height: 620
}
```

- Use `pdfviewer`, not `iframe`: the browser's native PDF viewer does not start
  inside a sandboxed frame, so the viewer frame is never sandboxed.
- The source is the security boundary: a dynamic `src` needs a literal
  `srcAllow` allow-list; a literal URL needs none.
- On desktop, grant the folder at build time (`--allow-dir`) and get the URL
  with `fs.documentUrl(path)`; the bridge serves only real PDF files.
- The control only displays the document: never read its text, pages or
  annotations, and never feed the bytes to another tool.

See [`examples/pdfviewer.sui`](../examples/pdfviewer.sui).

## Security checklist

Before finishing, verify:

- [ ] Every untrusted value is rendered as data, not markup.
- [ ] Dynamic `style:` uses `cssSafe`; dynamic classes do not carry security
      meaning.
- [ ] Every dynamic HTTP/WS URL uses `*Safe` with a literal allow-list.
- [ ] Every iframe has a trusted static `src` or a literal `srcAllow`; sandbox
      tokens are minimal.
- [ ] Every `pdfviewer` has a trusted static `src` or a literal `srcAllow`: its
      frame is never sandboxed.
- [ ] Table/CSV/JSON/bridge errors are checked and displayed.
- [ ] CSV exported for Excel/LibreOffice includes `;formula=escape`.
- [ ] Desktop capabilities are minimal and user input is validated on top.
- [ ] No secrets in URLs, logs or view text.
- [ ] `simpleui check` is clean; the build was actually run.

## Verification

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --single --out app.html
simpleui build app.sui --single --minify --out app.html
simpleui serve app.sui --watch --sourcemap
```

If the browser tool is available, verify:

1. loading state on first render;
2. success state with data;
3. empty result;
4. malformed response and network error;
5. a rejected/allow-listed iframe URL;
6. form validation and a successful action.

For compiler changes, also run:

```bash
go build -o simpleui.exe ./cmd/simpleui
go vet ./...
go test -count=1 ./...
```


