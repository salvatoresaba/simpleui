# File analyzer (SimpleUI + Go backend)

A complete example that selects files with the native dialog or drag and drop,
uploads them to a local Go service, analyzes their type in the backend, and
renders the results with `table.display`.

## Build executables

```powershell
cd examples/file_analyzer
.\build.ps1
# or: .\build.bat
```

This embeds the compiled SUI frontend into the Go server and produces:

```text
bin/file_analyzer.exe                Windows amd64
bin/file_analyzer_linux_amd64        Linux amd64
bin/file_analyzer_darwin_arm64       macOS Apple Silicon
```

Run the standalone executable (no separate web directory needed):

```powershell
.\bin\file_analyzer.exe --port 8097
```

`run.ps1` remains the development launcher (it serves `public/` from disk);
`build.ps1` is the release build with embedded assets.

## Run

```powershell
cd examples/file_analyzer
.\run.ps1
```

or:

```bat
cd examples\file_analyzer
run.bat
```

The script builds `app.sui` into `public/`, starts the Go server on
<http://127.0.0.1:8097/> and opens the browser.

Manual equivalent:

```bash
cd examples/file_analyzer
../../simpleui.exe build app.sui --out public
cd server
go run . --port 8097 --web-dir ../public
```

## How it works

1. The user chooses files with **Choose files...** (`pickFiles`) or drops them
   on the input (`dropFiles`). Multiple files and dropped directories are
   supported. **Choose folder...** uses `pickFiles ... directory: true`: the
   browser returns every file inside the folder recursively and
   `drag.folder()` shows the common folder name.
2. **Analyze** calls `http.postFiles("/api/analyze")`. The runtime uploads the
   actual selected file bytes as `multipart/form-data` with field name `files`.
3. The Go server analyzes each upload with a worker pool and returns a
   `{"columns": [...], "rows": [...]}` payload.
4. `table.display("analysis", data)` renders the rows while preserving the
   table design.

## Detection

The backend reads only the first 512 bytes of each upload and combines:

- `http.DetectContentType` content signatures (PDF, PNG, ZIP, text, ...);
- the file extension and `mime.TypeByExtension`;
- a category mapping (Image, Document, Spreadsheet, Presentation, Audio, Video,
  Archive, Code, Data, Executable, Font, Text, Other).

For ZIP-based Office files the content signature is `application/zip`; the
extension refines the category to Document/Spreadsheet/Presentation. The table
shows both the MIME type and the detection source.

## Displayed path

The table shows `name` and `path`. The path comes from the browser selection
metadata appended by `http.postFiles`:

- folder selections/drops expose a relative path such as `photos/2026/a.jpg`;
- individual files usually expose only their name;
- some desktop hosts expose an absolute `File.path`; browsers do not.

The path is display-only: the backend never opens it and never trusts it for
filesystem access.

## Folder mode

The **Choose folder...** button uses `pickMode: "auto"`:

- on Chromium secure contexts it uses the File System Access API
  (`showDirectoryPicker`), so it can retain the name of an empty folder;
- elsewhere it falls back to the native `webkitdirectory` picker.

Both forms return all files inside the folder recursively and show the folder
name through `drag.folder()`. The File System Access API is read-only and asks
for user permission.

A folder picker cannot choose only some files inside one folder. For a subset,
use **Choose files...** with multi-select. An empty folder chosen by the native
fallback returns no files and therefore no name; use the FS mode on Chromium when
the empty-folder name is required. In FS mode the app sends the folder name to
`/api/analyze-folder` and shows a `Directory` row.

## Limits and security

- localhost-only server (`127.0.0.1`);
- 64 MiB total request body, 200 files, 8 analysis workers;
- only the uploaded bytes are inspected; no path from the browser is trusted or
  read from the server filesystem;
- filenames are reduced to `filepath.Base` and rendered as text;
- content is never executed or decompressed and is not retained after the
  request (large multipart uploads may use Go's temporary parser storage, which
  `ParseMultipartForm`/`RemoveAll` cleans up);
- the browser runtime limits a selection to 2,000 files and 4,096 characters per
  path.

`http.postFiles` sends real bytes. Keep the endpoint under the application's
control and enforce the same limits on any production backend.

## API

`POST /api/analyze` with multipart field `files` (and parallel `paths`):

```json
{
  "columns": ["File", "Path", "Size", "MIME", "Category", "Detected by", "Detail"],
  "rows": [
    ["report.pdf", "inbox/report.pdf", "12.3 KB", "application/pdf", "Document", "content", "content signature application/pdf; ext .pdf"]
  ]
}
```

A folder selected through `pickMode: "fs"` can be empty. In that case the
frontend calls `POST /api/analyze-folder` with:

```json
{"folder": "EmptyFolder"}
```

and receives a single `Directory` row. This is how the backend can receive a
folder name even when there are no file bytes to upload.

On validation errors the response is JSON with an `error` field; the SUI app
shows it in an alert.

## Files

- `app.sui` - SimpleUI frontend;
- `server/main.go` - Go upload analyzer and static server (embedded web assets);
- `run.ps1` / `run.bat` - development launcher;
- `build.ps1` / `build.bat` - release executables;
- `public/` - disk web build used by the development launcher;
- `server/web/` - web build embedded in release executables;
- `bin/` - built executables.
