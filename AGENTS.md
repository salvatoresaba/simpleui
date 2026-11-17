# SimpleUI coding-agent guide

SimpleUI (SUI) is a small reactive UI language: a `.sui` file is compiled into a
self-contained web app (HTML/CSS/JS) or a desktop executable. This file is the
canonical entry point for coding agents. Read it before generating or editing
an interface.

## Start here

1. Run `simpleui docs --agent` for a compact context bundle.
2. Read `docs/manifest.json` (or `simpleui docs --json`) for the complete,
   machine-readable map of documents, tags, invariants and recipes.
3. Pick a recipe from `docs/AGENT_RECIPES.md` and open its linked canonical
   documents. Do not invent syntax from memory.
4. Read the closest example under `examples/`.
5. Write or edit the `.sui` file, then run `simpleui check app.sui --json`.
6. Build with `simpleui build app.sui --out dist` and serve with
   `simpleui serve app.sui --watch` when a browser preview helps.

If the `simpleui` binary is not built yet:

```bash
go build -o simpleui.exe .     # Windows
go build -o simpleui .         # Linux/macOS
```

## Non-negotiable rules

These rules are part of the language's security model. Do not work around them.

- **Commit every tracked change.** After finishing any edit to tracked files,
  run `git add -A` and create a descriptive commit before ending the turn. Do
  not leave tracked changes uncommitted. If a task is incomplete or failing,
  commit with a `wip:` prefix (on a branch when possible) so the work is
  recoverable. Generated outputs are ignored; never force-add them. Verify with
  `git status --short` and report the commit hash.
- **No raw HTML.** Interpolated text and attributes are escaped automatically.
  Raw HTML interpolation is intentionally unavailable. Data is data, never
  markup.
- **Validate untrusted input.** API bodies, files, dropped/picked file and
  folder names, CSV/JSON, URLs and iframe data are untrusted. Check every
  `(value, error)` result and show the error.
- **Dynamic CSS requires `cssSafe`.** `style: cssSafe(expr)` is required for a
  dynamic style; static styles are trusted. Do not add a sanitizer bypass.
- **Dynamic classes are filtered** to the app's class allow-list. Use static
  classes for layout.
- **Dynamic HTTP/WebSocket URLs require `*Safe`.** Use
  `http.getSafe/postSafe/getStatusSafe` or `ws.connectSafe` with a **literal**
  allow-list. A literal URL may use `http.get`/`ws.connect`.
- **Iframes are sandboxed by default.** Missing `sandbox` becomes `sandbox=""`;
  missing `referrerpolicy` becomes `no-referrer`. A dynamic `src` requires a
  literal `srcAllow` allow-list and is validated at runtime. Never use
  `srcdoc` or fallback markup.
- **Table data is inert.** Cells are written with `textContent`. Use
  `table.display`/`table.importText`, check the error, and preserve the table's
  styling. Use `;formula=escape` when exporting CSV for Excel/LibreOffice.
- **Desktop capabilities are explicit.** File/command access exists only in
  desktop builds with `--allow-dir` / `--allow-exec`. Validate user names in
  addition to the bridge checks.
- **Security configuration is strict.** `simpleui.security.json` is validated;
  malformed or unknown data blocks the build. Defaults are secure.
- **Do not edit generated output.** `dist/`, preview HTML and temporary files
  are results, not sources.
- **Do not put backticks inside `guihtml.go`.** It is a Go raw string containing
  JavaScript; adding a backtick breaks the Go string.
- **Do not invent APIs.** Confirm in `docs/API_REFERENCE.md`,
  `docs/STDLIB.md`, `docs/BUILTINS.md` or run `simpleui check`.
- **Documentation language.** Canonical documentation, code comments and UI
  strings are English. If a document has an Italian `_IT` counterpart, update
  both in the same change.

## Component delivery checklist

Whenever you add a new component, tag or control, deliver it as a complete
feature, not only as runtime code:

1. Add the compiler/type-checker support, generated HTML/CSS/JS, designer
   palette entry and inspector fields.
2. Add at least one runnable example that demonstrates **all** supported
   behavior: content, state, keyboard/accessibility, designer workflow and
   limitations. Prefer `examples/<component>/app.sui` with comments and an
   `examples/<component>/README.md`; a small `examples/<component>.sui` teaser
   may also be kept.
3. Include the **complete SUI source** in the example README (in a `sui` code
   block), not only a short snippet. Explain what is demonstrated, how to use
   the component, how to edit it in the designer and how to rebuild it.
4. When practical, include a ready-to-run build (for example
   `examples/<component>/dist/app.exe`) plus `run.ps1` / `run.bat` and a copy
   of the complete source as `dist/app.sui`, so a beginner can launch and
   inspect the demo without running compiler commands.
5. Update English and Italian docs, the manifest, the index and tests in the
   same change. Run `simpleui check`, the relevant build and `go test ./...`.

## Repository map

| Path | Purpose |
|---|---|
| `lexer.go`, `parser.go` | Tokens, AST, source loading |
| `ast.go`, `typechecker.go` | AST shapes and semantic/security checks |
| `jsgen.go`, `htmlgen.go` | JS/CSS generation and runtime prelude |
| `dnd.go` | Browser drag-and-drop runtime (component keys and file/folder drops) |
| `runtime.go`, `desktop_dialog.go` | Desktop bridge, served runtime and native dialogs |
| `builtins.go`, `stdlib.go` | Builtins and std modules |
| `main.go`, `serve.go`, `docs_cli.go` | CLI, dev server, docs command |
| `gui.go`, `guihtml.go` | Visual designer (live preview + `.sui` export) |
| `mcp.go`, `lsp.go` | MCP server and language server |
| `examples/` | Runnable, compiling examples |
| `docs/` | Navigation, recipes, API, stdlib and builtins |
| `SECURITY.md` | Threat model, protections and residual risks |
| `SPEC.md` | Normative language/compiler/runtime contracts |

## Build and verify

```bash
# Always after a language/compiler change
go build -o simpleui.exe .
go vet ./...
go test -count=1 ./...

# Always after editing a .sui file
./simpleui.exe check app.sui --json
./simpleui.exe build app.sui --out dist
./simpleui.exe build app.sui --single --out app.html
./simpleui.exe build app.sui --single --minify --out app.html
./simpleui.exe serve app.sui --watch

# Always after editing the control appearance catalog
./simpleui.exe appearance check components/appearance.json --json
```

`check --json` returns `ok` plus structured diagnostics; use it in an edit loop
instead of parsing human output. `build` fails on errors and prints warnings.
`serve --watch` rebuilds and reloads the browser. `appearance check` uses the
same validator as the designer; text-entry controls only allow the `None`
animation and `feedback.static` must not contain motion effects. The catalog is
self-contained: `definitions.baseCss`, `phaseCss` and `keyframes` plus each
control's `markup`, `classes`, `css` and `keyframes`; a class or keyframe with
no definition is rejected.

For desktop capabilities:

```bash
./simpleui.exe build app.sui --target desktop --out dist \
  --allow-dir C:/data --allow-exec helper.exe

# Embedded native window (WebView2 / WKWebView / WebKitGTK)
./simpleui.exe build app.sui --target desktop-webview --out dist \
  --allow-dir C:/data --allow-exec helper.exe

# Windows: double-click .sui files in the editor (HKCU, reversible)
./simpleui.exe associate
./simpleui.exe associate --uninstall
```

## Task routing

| Need | Read | Example |
|---|---|---|
| Local form, state, validation, table | `docs/AGENT_RECIPES.md#local-form--reactive-table`, `docs/API_REFERENCE.md` | `examples/todo.sui`, `examples/table_data.sui` |
| Animated image / GIF spinner | `docs/API_REFERENCE.md#spinner`, `docs/GUI_EDITOR.md#spinner-imagegif-mode` | `examples/spinner_image/README.md` |
| Translations, plurals, locale switching | `docs/API_REFERENCE.md#internationalization-i18n`, `docs/GUI_EDITOR.md#languages` | `examples/i18n/README.md` |
| Filter/transform/reduce lists with callbacks | `docs/BUILTINS.md#collection-functions` | `examples/collections.sui` |
| Destructuring, spread and named arguments | `docs/API_REFERENCE.md#destructuring-spread-and-named-arguments` | `examples/syntax_ergonomics.sui` |
| External component library (add/regroup palette presets) | `docs/COMPONENT_LIBRARY.md` | `components/library.json` |
| Control appearance (structure/style/animation, themes, navigator) | `docs/APPEARANCE_LIBRARY.md` | a `button` in the designer |
| Structural table layout for arranging UI regions | `docs/API_REFERENCE.md#table-layout`, `docs/GUI_EDITOR.md#table-layout-control` | `examples/table_layout/README.md` |
| Excel-like workbook with multiple tables in bottom sheet tabs | `docs/API_REFERENCE.md#excel-like-sheet-tabs`, `docs/GUI_EDITOR.md#excel-like-sheet-tabs` | `examples/table_sheets/README.md` |
| Menu bar / navigation | `docs/API_REFERENCE.md#menu-bar`, `docs/GUI_EDITOR.md#menu-bar-control` | `examples/menu_bar.sui` |
| Accordion / collapsible sections | `docs/API_REFERENCE.md#accordion`, `docs/GUI_EDITOR.md#accordion-control` | `examples/accordion/README.md`, `examples/accordion/dist/app.exe` |
| Toast notifications / feedback | `docs/API_REFERENCE.md#toast-notifications`, `docs/GUI_EDITOR.md#toast-control` | `examples/toast/README.md`, `examples/toast/dist/app.exe` |
| Date/time picker | `docs/API_REFERENCE.md#datetime-picker`, `docs/GUI_EDITOR.md#datetime-picker-control` | `examples/datepicker/README.md`, `examples/datepicker/dist/app.exe` |
| Tree / nested navigation | `docs/API_REFERENCE.md#tree`, `docs/GUI_EDITOR.md#tree-control` | `examples/tree/README.md`, `examples/tree/dist/app.exe` |
| Tree JSON/XML import/export | `docs/STDLIB.md#tree-module-tree-core`, `docs/API_REFERENCE.md#tree` | `examples/tree/README.md` |
| Graph / block diagram | `docs/API_REFERENCE.md#graph-control-block-diagrams`, `docs/GUI_EDITOR.md#graph-control` | `examples/graph/README.md` |
| Chart from JSON/XML/CSV with presets, scroll/zoom | `docs/API_REFERENCE.md#chart`, `docs/GUI_EDITOR.md#chart-control` | `examples/chart/README.md`, `examples/chart/dist/app.exe` |
| Runtime chart tags/colors, unique element ids, event-list polling, custom hints, table view/focus, element read/write API, static mode and shape/tag events | `docs/STATE_CHART_RUNTIME.md`, `docs/STDLIB.md#chart-module-chart-core` | `examples/process_monitor/README.md` |
| Multi-page navigation | `docs/API_REFERENCE.md#pages-and-navigation`, `docs/GUI_EDITOR.md#pages` | `examples/pages/README.md` |
| Modal dialogs | `docs/API_REFERENCE.md#modal-dialogs`, `docs/GUI_EDITOR.md#modal-control` | `examples/modal.sui` |
| Motion loops, spring easing, graph events, guards, gestures, visibility and exit | `docs/API_REFERENCE.md#motion-graph-states-and-triggers`, `docs/GUI_EDITOR.md#animation-panel` | `examples/motion_features/README.md`, `examples/motion_sequence/README.md` |
| Drag/drop or native file picker | `docs/GUI_EDITOR.md`, `docs/AGENT_RECIPES.md#drag-and-drop`, `docs/API_REFERENCE.md#drag-and-drop` | `examples/drag_drop.sui`, `examples/file_picker.sui` |
| Remote JSON API | `docs/AGENT_RECIPES.md#remote-json-api--table`, `docs/STDLIB.md#http-module-http` | `examples/users/app.sui` |
| HTTP headers, multipart upload & advanced WebSocket | `docs/STDLIB.md#http-module-http`, `docs/STDLIB.md#websocket-module-ws` | `examples/networking_suite/README.md` |
| POST/update remote data | `docs/AGENT_RECIPES.md#remote-crud-form`, `SECURITY.md` | `examples/users/app.sui` |
| Polling dashboard | `docs/AGENT_RECIPES.md#polling-dashboard` | `examples/monitor.sui` |
| Desktop files/commands | `docs/AGENT_RECIPES.md#desktop-files-and-commands`, `SECURITY.md#desktop-bridge` | `examples/security/desktop_files.sui` |
| Native file/folder dialogs | `docs/STDLIB.md#dialog-module-dialog-desktop` | `examples/desktop_picker.sui` |
| XLSX/ODS conversion | `examples/spreadsheet_table/README.md` | `examples/spreadsheet_table/app.sui` |
| Third-party embed | `docs/AGENT_RECIPES.md#iframe`, `SECURITY.md#iframe-risks-in-detail` | `examples/security/untrusted_iframe.sui` |
| Security review | `SECURITY.md`, `examples/security/README.md` | `examples/security/*.sui` |

Full index: `docs/INDEX.md`. Machine manifest: `docs/manifest.json`.

## MCP and automation

Run `simpleui mcp` for a stdio MCP server. It exposes:

- tools: `sui_check`, `sui_model`, `sui_compile`, `sui_render`, `sui_docs`, `sui_appearance`;
- resources: `sui://docs/<path>` for every manifest document;
- prompts: `build_simpleui_interface` and `review_simpleui_security`.

Typical client configuration:

```json
{
  "mcpServers": {
    "simpleui": {
      "command": "simpleui",
      "args": ["mcp"]
    }
  }
}
```

Offline documentation commands (work without MCP):

```bash
simpleui docs --agent          # compact context
simpleui docs --json           # full manifest
simpleui docs --list           # id and path
simpleui docs --read api       # one document
simpleui docs --search table   # title/tag/summary search
simpleui docs --full           # all embedded docs
```

The docs are embedded in the binary at build time, so these commands work from
any directory. The source of truth remains the Markdown/JSON in this repository;
`docs_nav_test.go` checks links, anchors and manifest consistency.

## Editing guidance

- Match the existing style in `.sui` files and the compiler. Keep security
  checks and their diagnostics explicit.
- Prefer small, composable views and components; derive display text in views
  and keep handlers small.
- For remote data, model explicit `loading`, `ok`, `empty` and `error` states.
- For tables, keep a static model in the view (headers/cells) and populate it
  with `table.display`; do not rebuild the table markup for every payload.
- For local persistence, use `storage.*`; do not assume `fs` on web.
- For desktop, gate UI actions on capabilities; show a clear message when a
  bridge call is denied.
- When changing docs, update `docs/manifest.json`, `docs/INDEX.md` and any
  relevant anchors together; run the docs tests.

## Output contract

When you finish a task, report:

1. the `.sui` files (and, only if necessary, compiler files) changed;
2. `simpleui check` result and the build/serve command used;
3. any security decisions (allow-lists, sandbox tokens, desktop capabilities);
4. known residual risks or deployment headers still required;
5. the git commit created for the change (hash and message).

Do not claim a UI works if `simpleui check` fails or if the build was not run.


