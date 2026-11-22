# SimpleUI (SUI)

SimpleUI is a reactive DSL language for building user interfaces, compiled into standalone web and desktop applications written in Go.

## What is SimpleUI

SimpleUI lets you write UI in a `.sui` file using a simple, declarative syntax. The compiler automatically translates your code into **vanilla HTML, CSS and JavaScript**, with no external dependencies.

Key features:
- **Reactive**: `state` assignments automatically update the interface
- **Drag and drop / file picker**: reorder elements, drag files/folders, or use the native file dialog
- **Menu bar**: insert a configurable navigation menu with submenus, links and actions
- **Tabs**: insert working tab sets with clickable panels and keyboard navigation
- **Editor preferences**: choose code font, size, line height, tabs, wrapping and line-number options
- **Context-aware completion**: press Ctrl+Space for declarations, elements, attributes, handlers and expressions at the caret
- **Formatted function bodies**: generated functions keep their nested indentation after designer edits
- **Symbol highlighting**: function names and variables use distinct colors in declarations, references and interpolations
- **Unified undo/redo**: text, properties, elements and designer actions share one history with buttons, Ctrl+Z and Ctrl+U
- **Effective inspector styles**: width, height, colors and other fields reflect authored, inherited, inline and computed values
- **Collapsible element tree**: container rows expand and collapse with a disclosure arrow
- **Accordion control**: nested disclosure sections with keyboard-accessible headers and restored open state
- **Tree control**: whole-tree selection, context-menu structural editing, leaf/parent/open icons, open/closed branch symbols, curated icon picker, tree-grid columns and `onTreeChange` with `tree.lastChange`
- **Table control**: declarative columns, sorting, filtering, pagination, row/column/cell selection, editable validated cells, table events, a tabbed Table settings overlay, multiple tables in Excel-like bottom sheet tabs, runtime table creation and row/column structure editing
- **Chart control**: SVG bar, horizontal/stacked/grouped bar, line/step/smooth, area, combo, pie/donut, scatter, histogram, waterfall, radar, gauge, funnel, heatmap, candlestick, treemap, box plot, bullet, dumbbell, Gantt and state-machine charts from JSON/XML/CSV with compatibility checks, scroll/zoom, presets, a resizable data editor, synchronized accessible tables, optional row insertion/editing/deletion, row/column/cell selection, visible numeric steppers, context-menu copy/export, selected-region export and virtual scrolling, critical path, holidays, configurable weekends and IANA timezone
- **Exact type badges**: every structure row shows its precise element type (`div`, `treeitem`, `chart`, ...)
- **Toast notifications**: safe text feedback with info/success/warning/error variants, timeouts and an optional host
- **Date/Time picker**: native date, time and datetime controls with state binding, min/max, clear and validation
- **Combobox**: searchable option list with runtime updates, escaping and wildcard search
- **Pages**: multiple views with `navigate("page")` and editor page chips
- **Modal dialogs**: configurable dialogs with preset primary/danger buttons and `modal.open()` / `modal.close()`
- **Components**: create reusable components with parameters, slots, and per-instance state
- **Typed**: static type checking with inference for `state` and `let`
- **Expressive syntax**: destructuring, spread and named function arguments
- **Nominal enums**: exhaustive `switch`, duplicate-case checks and a checked enum type
- **Union types**: `string | int`, `string?` and narrowing `is` guards
- **Literal types**: `"idle" | "loading"` with exhaustive switch
- **Generics and rest parameters**: `fn first<T>(items: array[T])`, `...values: array[int]`
- **Const, computed and readonly**: immutable values, derived state, immutable fields
- **Reactive batching, error boundaries and list keys**
- **Intl formatting**: dates, numbers, currencies, percentages, relative time and time zones
- **Advanced stdlib**: fs copy/move/size, crypto hash/UUID/random bytes, path helpers, JSON Schema validation, storage quota and cross-tab sync
- **JSON library**: JSON Pointer (RFC 6901), JSON Patch (RFC 6902), Merge Patch (RFC 7386), keyed diff with moves, three-way merge, multi-error schema validation, UTF-8/UTF-16 bytes and desktop file helpers
- **Rich text editor**: JSON document model, rich text/HTML/Markdown authoring profiles with editable HTML/Markdown source, configurable toolbar/menu/context bar, status bar, undo/redo, TXT/RTF/DOC/DOCX/ODT/PDF file support, JSON/HTML/Markdown/XML import/export, page background colors and plugin commands
- **XML library**: DOM navigation/mutation, XPath, namespaces, comments/CDATA/PI, XSD subset with full error collection, JSON/CSV conversion, XSLT, keyed diff with move detection, patch application, update/three-way merge and binary (UTF-8/UTF-16) file I/O
- **Multi-target**: web apps, browser-launched desktop executables and native-window desktop executables with an embedded WebView
- **No Node/npm**: generated web code needs only a modern browser; desktop executables need the platform WebView runtime
- **Secure for desktop**: native bridge with capability-based permission system and native file/folder dialogs

## After cloning the repository

Everything below runs from the repository root and needs **only Go**: the
compiler is a single binary and the apps it produces are plain HTML/CSS/JS.
The commands use `simpleui.exe` on Windows and `simpleui` on Linux/macOS.

### Prerequisites

- [Go 1.24.2+](https://go.dev/dl/) — the only requirement, `go version` must work
- A modern browser, to run the compiled apps and the visual designer
- Windows/Linux/macOS for desktop builds
- Node.js 20+ only to build the VS Code extension in `vscode-hermes-extension/`
  (`scripts/verify.ps1` skips that step when npm is missing)
- `desktop-webview` builds also need CGO, a C toolchain and the platform WebView
  runtime (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux)

### 1. Get the code and build the compiler

```bash
git clone https://github.com/salvatoresaba/simpleui.git
cd simpleui

# Windows
go build -o simpleui.exe ./cmd/simpleui

# Linux/macOS
go build -o simpleui ./cmd/simpleui
```

Build from `./cmd/simpleui`: the repository root is the importable `sui`
package, so `go build -o simpleui` there writes a Go archive, not a runnable
executable.

Check that it works:

```bash
./simpleui.exe version     # -> simpleui 0.1.0-alpha
./simpleui.exe help        # -> every available command
```

The module path is `github.com/salvatoresaba/simpleui`: the compiler can also
be imported as a Go library from another module.

### 2. Open the visual designer

```bash
# Windows: double-click GUI_editor.bat (it compiles the compiler if missing)
GUI_editor.bat

# Linux/macOS
chmod +x GUI_editor.sh
./GUI_editor.sh

# Or start it directly, with or without a file
./simpleui.exe gui
./simpleui.exe gui examples/counter.sui
```

The designer opens in your browser at the local URL printed in the terminal;
**File → Save** writes back to the same `.sui` file. On Windows you can also
register the editor for `.sui` files: `./simpleui.exe associate`
(`--dry-run` to preview the registry commands, `--uninstall` to undo).

### 3. Compile and run your first app

```bash
# Parse + typecheck: ok plus structured diagnostics
./simpleui.exe check examples/counter.sui --json

# Static web app in dist/ (index.html, style.css, app.js)
./simpleui.exe build examples/counter.sui --out dist

# Local server that rebuilds and reloads on every save
./simpleui.exe serve examples/counter.sui --watch

# Desktop executable (opens the system browser)
./simpleui.exe build examples/counter.sui --target desktop --out dist
```

`dist/` is a plain static app: serve it with any web server, or open
`index.html` directly. Desktop builds get no ambient permissions —
`--allow-dir` and `--allow-exec` are the only capabilities you grant, and they
are always explicit.

### 4. Run the whole verification suite

```bash
pwsh scripts/verify.ps1
```

Format, build, vet, the complete test suite (including the headless-browser
E2E tests), every example, the appearance catalog, the embedded documentation
and the VS Code extension.

### 5. Where to go next

| I want to | Open |
|---|---|
| Learn the language | [`docs/INDEX.md`](docs/INDEX.md), then [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) |
| See working apps | [`examples/README.md`](examples/README.md) |
| Copy ready-made patterns | [`docs/AGENT_RECIPES.md`](docs/AGENT_RECIPES.md) |
| Understand the contracts | [`SPEC.md`](SPEC.md), [`DESIGN.md`](DESIGN.md) |
| Check the security model | [`SECURITY.md`](SECURITY.md) |
| Work on the compiler itself | [`AGENTS.md`](AGENTS.md), [`docs/manifest.json`](docs/manifest.json) |

## Releases

Prebuilt binaries are published as GitHub Releases: every tag `v*` builds the
compiler for Windows, Linux and macOS (`amd64` and `arm64`). Each release
contains one archive per target plus a `SHA256SUMS` file.

**No release has been published yet**, so the download links below only start
working once the first tag is pushed. Until then, build the binary from source
with the commands in [After cloning the repository](#after-cloning-the-repository).

```bash
# Linux/macOS
curl -LO https://github.com/salvatoresaba/simpleui/releases/latest/download/simpleui_linux_amd64.zip
unzip simpleui_linux_amd64.zip && ./simpleui version

# Windows (PowerShell)
Invoke-WebRequest -OutFile simpleui_windows_amd64.zip https://github.com/salvatoresaba/simpleui/releases/latest/download/simpleui_windows_amd64.zip
Expand-Archive simpleui_windows_amd64.zip
.\simpleui.exe version
```

Verify the download before running it:

```bash
sha256sum --check SHA256SUMS
```

The archives are built by `.github/workflows/release.yml` from the same source
as the CI workflow; the release is not signed, so check the checksums.

## Quick start

### 1. Create an example app

Create a `counter.sui` file:

```sui
state count = 0

fn increment() {
  count = count + 1
}

fn decrement() {
  count = count - 1
}

view App {
  col {
    h1 "Counter"
    text "Value: " + str(count)
    row {
      button "Decrement" onClick: decrement
      button "Increment" onClick: increment
    }
  }
}
```

### 2. Build for web

```bash
./simpleui build counter.sui --target web --out dist
```

This produces:
- `dist/index.html`
- `dist/style.css`
- `dist/app.js`

Open `dist/index.html` in your browser to see the app.

To export everything into a **single HTML file** (inline CSS, JS, runtime, and designer fonts):

```bash
./simpleui build counter.sui --single --out app.html
./simpleui build counter.sui --single --minify --out app.html
```

With `--out dist` the file becomes `dist/app.html`. `--minify` is conservative: it removes comments and indentation without changing semantics.

To debug the generated code, enable source maps explicitly:

```bash
./simpleui build counter.sui --out dist --sourcemap
./simpleui serve counter.sui --watch --sourcemap
./simpleui build counter.sui --single --sourcemap --out app.html
```

Multi-file builds write `app.js.map` and `style.css.map`; single-file builds inline both maps as `data:` URLs. Browsers map JavaScript and CSS, not HTML: the SUI-to-HTML view path is covered by the JavaScript map because `app.html` is rendered by `app.js`. Source maps embed `sourcesContent`, so keep them private and use the flag only for development. `--sourcemap` cannot be combined with `--minify`.

### 3. Build for desktop

Browser renderer:

```bash
./simpleui build counter.sui --target desktop --out dist
./dist/app.exe  # Windows
# or ./dist/app on Linux/macOS
```

Opens the app in the default browser; use `--no-open` for headless mode.

Embedded native WebView (no separate browser window):

```bash
./simpleui build counter.sui --target desktop-webview --out dist
./dist/app.exe  # Windows
```

This produces a native window using WebView2 on Windows, WKWebView on macOS
and WebKitGTK on Linux. The same localhost bridge and `--allow-exec` /
`--allow-dir` grants apply. Runtime options are `--no-open` and `--port N`.

The first build downloads `github.com/webview/webview_go`; the WebView target
needs CGO and a C toolchain. On macOS install the Xcode Command Line Tools. On
Linux install `libgtk-3-dev` and `libwebkit2gtk-4.0-dev` (or the equivalents
for your distribution). On Windows WebView2 Runtime is normally already present
on Windows 10/11.

### 4. Visual designer

Open an empty designer, or start directly from an existing `.sui` file:

```bash
./simpleui gui
./simpleui gui examples/counter.sui
```

With a file argument the editor opens that file immediately and **Save** writes
back to the same path. Use `--port N` for a fixed editor port and
`--security <file>` for an explicit policy.

On Windows and Linux/macOS the repository also ships one-click launchers:
`GUI_editor.bat` and `GUI_editor.sh`. They start the editor from the
repository root and compile the compiler automatically if the binary is
missing, so they work on a fresh checkout:

```bash
./GUI_editor.bat                    # Windows
./GUI_editor.sh examples/counter.sui # Linux/macOS, optional file argument
```

On Windows you can associate `.sui` files with the editor once:

```bash
./simpleui associate
./simpleui associate --dry-run     # show the registry commands only
./simpleui associate --uninstall
```

The association is written under `HKEY_CURRENT_USER`, so it needs no
administrator rights and is reversible. Double-clicking a `.sui` file then opens
the designer on that file. If Windows already had a different default app for
the extension, choose SimpleUI once from **Open with** or reset it in Windows
Settings.

### 5. Development mode

Use the dev server with live reload:

```bash
./simpleui serve counter.sui --watch
```

Open the URL shown in the terminal to see the app with automatic updates on every `.sui` file change.

### 6. Additional tools

```bash
# Check parse + typecheck (--json for agents/CI)
./simpleui check counter.sui --json

# Show AST
./simpleui parse counter.sui

# Format file
./simpleui fmt counter.sui

# Show help
./simpleui --help
```

## Integration with coding agents

The compiler ships everything an agent needs to build SimpleUI interfaces.
None of these assets require a plugin marketplace.

| Asset | Purpose |
|---|---|
| MCP server (`simpleui mcp`) | Tools, documentation resources and prompts over stdio. See [`docs/MCP_SETUP.md`](docs/MCP_SETUP.md). |
| [`.mcp.json`](.mcp.json) | Ready project-scoped MCP configuration. |
| [`AGENTS.md`](AGENTS.md) | Canonical agent workflow, security rules and task routing. |
| [`CLAUDE.md`](CLAUDE.md) | Pointer for Claude Code. |
| [`.cursor/rules/simpleui.mdc`](.cursor/rules/simpleui.mdc) | Cursor rule. |
| [`.github/copilot-instructions.md`](.github/copilot-instructions.md) | GitHub Copilot instructions. |
| [`llms.txt`](llms.txt) | Compact discovery file for LLM clients. |
| [`vscode-hermes-extension/`](vscode-hermes-extension) | VS Code extension (Hermes): check and build `.sui` files, packaged as `simpleui-vscode-0.1.0.vsix`. |

- Build: `go build -o simpleui.exe ./cmd/simpleui`
- Check: `./simpleui.exe check app.sui --json`

The Hermes skill and preview plugin folders (`skills/`, `desktop-plugins/`)
are not part of this repository: the MCP server and the agent instruction
files above give an agent the same capabilities.

## Project structure

- `cmd/simpleui/main.go`: executable entry point; the module root is the
  importable `sui` package
- `facade.go`: generated public API of `sui` (type aliases and wrappers over
  `internal/core`)
- `components/<control>/`: per-control browser runtime (`Runtime`) and tests
- `internal/core/`: compiler, designer, stdlib and CLI implementation
- `internal/core/assets/`: embedded designer page, PDF font data and web runtime assets
- `internal/apitools/`: public API snapshot, parity test and facade generator
- `scripts/verify.ps1`: single verification command (format, build, vet, tests, checks)
- `internal/core/parser.go` / `internal/core/lexer.go`: lexical analysis and parsing of `.sui`
- `internal/core/typechecker.go`: static type checking with inference
- `internal/core/htmlgen.go` / `internal/core/jsgen.go` / `internal/core/cssmodel.go`: code generation for web
- `internal/core/dnd.go`: browser drag-and-drop runtime (component keys and file/folder drops)
- `internal/core/desktop.go` / `internal/core/runtime.go`: desktop target and capability-checked bridge
- `internal/core/gui.go` / `internal/core/guihtml.go`: visual designer, preview, exports and desktop build dialog
- `internal/core/security.go` / `internal/core/iframe.go` / `internal/core/pdfviewer.go`: security policy, iframe and PDF viewer validation

## Documentation

- [`docs/INDEX.md`](docs/INDEX.md) — human and agent navigation index
- [`AGENTS.md`](AGENTS.md) — canonical coding-agent workflow and rules
- [`docs/AGENT_RECIPES.md`](docs/AGENT_RECIPES.md) — local, remote, desktop and iframe recipes
- [`docs/GUI_EDITOR.md`](docs/GUI_EDITOR.md) — designer drag and drop vs application drag and drop, inspector properties and pickers
- [`docs/manifest.json`](docs/manifest.json) — machine-readable documentation map
- [`docs/MCP_SETUP.md`](docs/MCP_SETUP.md) — register the MCP server with a coding agent (ready `.mcp.json` blocks)
- [`llms.txt`](llms.txt) — compact LLM discovery file
- [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) — language and API reference
- [`SECURITY.md`](SECURITY.md) — threat model, iframe/SSRF, bridge and residual risks
- [`TODO.md`](TODO.md) — verified implementation status and roadmap

Canonical documentation is English. Italian translations use the `_IT` suffix;
when a canonical document changes, update its `_IT` counterpart in the same
change.

## Implementation status

The verified status is tracked in [`TODO.md`](TODO.md). WebSocket reconnect, advanced filesystem operations, crypto and multipart upload are **not** complete yet and are not exposed as stable APIs. Clipboard, storage, timers, regex, validation, tables, the core HTTP/WebSocket safe guards and opt-in JS/CSS source maps are implemented.

## License

MIT — see [`LICENSE`](LICENSE). The project is at **0.1.0-alpha**: use it with
awareness in development and experimental environments.
