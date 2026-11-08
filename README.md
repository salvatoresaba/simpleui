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
- **Multi-target**: web apps, browser-launched desktop executables and native-window desktop executables with an embedded WebView
- **No Node/npm**: generated web code needs only a modern browser; desktop executables need the platform WebView runtime
- **Secure for desktop**: native bridge with capability-based permission system and native file/folder dialogs

## Requirements

- [Go 1.24.2+](https://go.dev/dl/) to compile the compiler and desktop builds
- A modern browser to run web apps
- `desktop-webview` also needs CGO, a C toolchain and the platform WebView runtime (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux)
- Windows/Linux/macOS for desktop builds

## Compiling the compiler

Clone or download the project and build the binary:

```bash
# Clone the repository (if from git)
git clone <repo-url>
cd GO_PROJECT_LANG

# Build the compiler
go build -o simpleui

# Or on Windows
go build -o simpleui.exe
```

Verify the installation:

```bash
./simpleui version
# Output: simpleui 0.1.0-alpha
```

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

## Integration with Hermes
- Skill: `simpleui` (`skills/devops/simpleui/`)
- Plugin: `simpleui-preview` (`desktop-plugins/simpleui-preview/plugin.js`)
- Skill: `opencode` (`skills/autonomous-ai-agents/opencode/`)
- Plugin: `opencode-preview` (`desktop-plugins/opencode-preview/plugin.js`)
- VS Code extension: `vscode-hermes-extension/` (generated `.vsix`)
- Build: `go build -o simpleui.exe .`
- Check: `./simpleui.exe check app.sui --json`

## Project structure

- `parser.go` / `lexer.go`: lexical analysis and parsing of `.sui`
- `typechecker.go`: static type checking with inference
- `htmlgen.go` / `jsgen.go` / `cssmodel.go`: code generation for web
- `dnd.go`: browser drag-and-drop runtime (component keys and file/folder drops)
- `desktop.go` / `runtime.go`: desktop target and capability-checked bridge
- `gui.go` / `guihtml.go`: visual designer, preview, exports and desktop build dialog
- `security.go` / `iframe.go`: security policy and iframe validation

## Documentation

- [`docs/INDEX.md`](docs/INDEX.md) — human and agent navigation index
- [`AGENTS.md`](AGENTS.md) — canonical coding-agent workflow and rules
- [`docs/AGENT_RECIPES.md`](docs/AGENT_RECIPES.md) — local, remote, desktop and iframe recipes
- [`docs/GUI_EDITOR.md`](docs/GUI_EDITOR.md) — designer drag and drop vs application drag and drop, inspector properties and pickers
- [`docs/manifest.json`](docs/manifest.json) — machine-readable documentation map
- [`llms.txt`](llms.txt) — compact LLM discovery file
- [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) — language and API reference
- [`SECURITY.md`](SECURITY.md) — threat model, iframe/SSRF, bridge and residual risks
- [`TODO.md`](TODO.md) — verified implementation status and roadmap

Canonical documentation is English. Italian translations use the `_IT` suffix;
when a canonical document changes, update its `_IT` counterpart in the same
change.

## Implementation status

The verified status is tracked in [`TODO.md`](TODO.md). WebSocket reconnect, advanced filesystem operations, crypto, multipart upload and source maps are **not** complete yet and are not exposed as stable APIs. Clipboard, storage, timers, regex, validation, tables and the core HTTP/WebSocket safe guards are implemented.
