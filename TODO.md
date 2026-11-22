# SimpleUI TODO List

**Status:** 0.1.0-alpha  
**Last review:** 2026-11-07
**Legend:** `[x]` implemented and verified · `[~]` partial / documented-only · `[ ]` missing

This file tracks the remaining work. A checked item must exist in the
compiler/codegen/runtime and be covered by a test or a manual browser check. Do
not mark an item complete only because it appears in documentation.

## 0. Current snapshot

Verified on 2026-10-25 with:

```bash
go build -o simpleui.exe ./cmd/simpleui
go vet ./...
go test -count=1 ./...
```

All three commands are green on Windows with Go 1.24.2.

Recently completed:

- [x] Reactive `class:`/`style:` bindings and ternary expressions.
- [x] Table data binding, designer editing, CSV/TSV/SSV/PSV/JSONL import/export,
      editable/fixed tables and formula-safe CSV export.
- [x] XLSX/ODS conversion path (`sheet2conv`/`sheet2json`) with upstream parsing.
- [x] Security configuration (`simpleui.security.json`) with strict validation.
- [x] Iframe hardening: default sandbox, `srcAllow`, runtime URL validation and a
      visible English rejection message; CSP/frame headers for `serve`/desktop.
- [x] `simpleui check --json` for CI and agents.
- [x] Embedded documentation, manifest, agent guide, recipes, `llms.txt`,
      `simpleui docs`, and MCP documentation resources/prompts.
- [x] Single-file web export (`--single`) and conservative `--minify`.
- [x] Designer desktop-build dialog with preflight, capabilities, log, local
      output path and optional token-authenticated download.
- [x] Embedded desktop WebView target (`desktop-webview`) with WebView2 /
      WKWebView / WebKitGTK, same bridge and capability model as `desktop`.
- [x] Designer target selector for browser vs embedded WebView, plus
      start-from-file (`simpleui gui app.sui`) and Windows `.sui` association.
- [x] Shared project theme system (Light/Dark families, page and base-element
      theming, tooltips, media, hover states, chart/graph theme presets, theme
      export/import and automatic contrast correction).
- [x] Theme regression coverage: computed-style contrast audit for all variants
      and golden screenshots for Standard Light/Dark.
- [x] Problems-panel filtering for generated runtime classes, keyframe steps,
      layered `:root` blocks and intentional state/pseudo overrides.
- [x] Streamlined chart inspector: type, Mode and Chart settings first, data at
      the bottom, point width in Chart settings, attributes panel removed and
      X/Y fields removed in favor of editing the data directly.
- [x] Project preferences themed controls and category cards in a two-column
      grid.
- [x] Tree control end-to-end: whole-tree selection, context-menu structural
      editing, leaf/parent/open icons, open/closed branch symbols, curated icon
      picker, `onTreeChange` and `tree.lastChange`.
- [x] Table control end-to-end: declarative columns, sorting, filtering,
      pagination, row/column/cell selection, validated editing, table events,
      `table.lastChange`, a tabbed Table settings overlay, Excel-like bottom
      sheet tabs, runtime table creation and row/column structure editing.

Known accuracy corrections from the previous TODO:

- `http.get`/`post` custom headers and timeouts, `http.upload`, WebSocket
  reconnect/heartbeat, advanced `fs.*`, `crypto.*`, path utilities and
  storage quota/sync are implemented and documented. Source maps are opt-in with `--sourcemap`.
- File imports (`import { X } from "./file.sui"`) work in `check`, `build` and
  `serve`, but the import list is not validated, all symbols are merged into one
  global namespace, and the designer still parses only the entry file. Proper
  exports/aliases/namespaces and designer multi-file support remain open.
- LSP supports basic completion plus live diagnostics, not the full editor
  feature set (hover, definition, references, symbols, rename, code actions).
- `--single --minify` is a conservative minifier, not a general JS minifier.

## 1. Core Language

### 1.1 Modules & Imports

- [ ] **User modules** — `import "path/to/file.sui"` for shared
      components/functions
- [ ] **Symbol exports** — `export view Component {...}`, `export fn {...}`
- [ ] **Import aliases** — `import { x as y } from "..."` /
      `import * as utils from "..."`
- [ ] **Relative/absolute path resolution** with import-cycle detection
- [ ] **Directory imports** (`index.sui`)

> Current state: only `import { ... } from "std"` is supported. The entry file
> can load imported `.sui` files internally, but there is no user-facing module
> syntax/export system yet.

### 1.2 Type System

- [x] **Enums and switch** - `enum Status { Idle, Loading, Error }` with nominal typing, exhaustive `switch`, duplicate-case checks and `default`
- [x] **Union types** - `string | int | null`, `T?` shorthand and union inference for mixed arrays/maps
- [x] **Type guards / narrowing** - `if v is string { ... }`, `while`, `not`, `and`/`or` branch narrowing
- [x] **Generics** — generic functions and components
- [x] **Literal types** — string/number/boolean literal types, strict assignment checks and exhaustive literal switch
- [x] **Const** — compile-time values
- [x] **Readonly** — immutable struct fields/variables

### 1.3 Functions & Control

- [x] **Rest parameters** — `fn sum(...nums: array[int])`
- [x] **Named parameters** at user-function call sites (`f(b: 2, a: 1)`)
- [x] **Destructuring** — `let {name, age} = user`, `let [a,b] = arr`
- [x] **Spread operator** — `[...arr1, ...arr2]`, `{...obj1, ...obj2}`

## 2. Reactivity & Runtime

### 2.1 Reactive System

- [x] **Reactive class/style binding** — `class:` and `style:` accept
      expressions and update without ids
- [x] **Conditional expression (ternary)** — with type checking on both branches
- [x] **Computed/derived state** — `computed doubled = count * 2`
- [ ] **Effects/watchers** — `effect { ... }`
- [x] **Batching updates** — group multiple state assignments into one re-render
- [ ] **Fine-grained reactivity** — selective DOM updates

### 2.2 Advanced UI Components

- [x] **Error boundaries** — subtree error fallback
- [x] **List keys** — `list item in items key: item.id`
- [x] **PDF viewer** — `pdfviewer` shows a document in the browser's native PDF
      viewer: a remote URL, a local file through the desktop bridge
      (`fs.documentUrl`) or a picked file. Display only: the app never reads
      the PDF bytes
- [ ] **Suspense-like loading** — async loading states
- [ ] **Context/Provide-Inject** — avoid prop drilling

## 3. Builtins/Stdlib (Improvements)

### 3.1 Completed and verified

- [x] **Time/Date** — `time.now`, `nowMs`, `format`, `parse`, `add`, `diff`,
      `toISO`, `toMs`
- [x] **URL/Encoding** — URL encode/decode/component, base64
- [x] **Cross-target storage** — get/set/remove/clear/keys
- [x] **Clipboard** — readText/writeText
- [x] **Advanced timers** — clear/cancel/debounce/throttle
- [x] **Advanced regex/string helpers** — match, matchAll, search, test,
      replaceRegex, replaceFirst, splitRegex, extract, extractAll
- [x] **Extended validation** — email/URL/empty/required/length/pattern/type
- [x] **Table core** — display/data/header/select/slice/importText/exportText/
      json/clear, with limits and inert cells
- [x] **HTTP safe variants** — `getSafe`, `getStatusSafe`, `postSafe`;
      `ws.connectSafe` with literal allow-lists, redirect blocking, timeout and
      response cap
- [x] **HTTP custom headers + timeout override** — `getWithHeaders`, `postWithHeaders`,
      `request`, `requestSafe` with configurable ms timeout and custom headers
- [x] **HTTP multipart/form-data** — `http.upload`, `http.uploadSafe`,
      `http.uploadWithHeaders`, `http.uploadWithHeadersSafe`
- [x] **Advanced WebSocket** — `ws.connectWithOptions` with exponential backoff reconnect,
      heartbeat/ping-pong, `onOpen`, `onClose`, `onError`, `status`, and `reconnect`
- [x] **JSON** — `json.parse`, `json.stringify`
- [x] **Collection callbacks** - `map`, `filter`, `reduce`, `find`,
      `findIndex`, `any`, `all`, `sortBy`, `unique` with declared functions.
- [x] **I18n catalogs and plurals** - `i18n "xx" { ... }`, `t`, locale
      switching/detection, CLDR plurals, RTL, `?locale=`, designer Languages
      editor and JSON import/export.

### 3.2 Missing / partial

- [x] **Advanced filesystem** - `fs.copy`, `fs.move/rename`, `fs.isDir`,
      `fs.size` with capability checks on both paths
- [x] **Crypto** - `crypto.hash`, `crypto.uuid`, `crypto.randomBytes`
- [x] **Path utilities** — `path.join`, `dirname`, `basename`, `extname`
- [x] **JSON Schema validation** - practical subset; `$ref` rejected
- [x] **XML library** - DOM nodes, navigation/mutation, XPath, namespaces, comments/CDATA/PI, XSD subset with `validateXsdAll`, JSON/CSV/XSLT, keyed diff with moved operations, `xml.patch`, update merge, three-way `merge3` and binary `loadBytes`/`saveBytes`
- [x] **JSON library** - JSON Pointer (RFC 6901), JSON Patch (RFC 6902), Merge Patch (RFC 7386), keyed diff with moves, `merge3`, `validateAll`, UTF-8/UTF-16 bytes and desktop `loadFile`/`saveFile`/`loadBytes`/`saveBytes`
- [x] **Rich text editor** - JSON document model, configurable toolbar/menu/context bar, status bar, templates, undo/redo, plugins, TXT/RTF/DOC/DOCX/ODT/PDF file support and JSON/HTML/Markdown/XML import/export
- [x] **I18n helpers** - translation catalogs and plural selection.
- [x] **Locale formatting** - `Intl` date/number/currency/relative/timezone
      helpers (`intl.*`).
- [x] **Advanced storage** - `storage.quota` (browser estimate) and
      `storage.sync` (local cross-tab re-render listener)

> Code review note: public stdlib functions must have a real implementation,
> typed entries, tests and documentation. Unsupported FS/Crypto operations
> remain intentionally absent; source maps are opt-in with `--sourcemap`.

## 4. Code Generation & Build

### 4.1 Completed

- [x] **Desktop target** — one self-contained Go executable with embedded
      assets, loopback bridge, capability allow-lists and browser launch
- [x] **Single-file web export** — `build --single`, inlining CSS/JS/runtime
- [x] **Conservative minify** — `--single --minify` removes CSS comments/
      whitespace and JS full-line comments/indentation
- [x] **Designer single-file export** — `File → Export single HTML…`
- [x] **Designer desktop build** - dialog, preflight, capabilities, log, local
      output, optional download, browser/WebView target selector
- [x] **Embedded WebView desktop target** - WebView2 / WKWebView / WebKitGTK
- [x] **CSP/frame headers** - `serve` and desktop send
      `X-Frame-Options: DENY` + `frame-ancestors 'none'`

### 4.2 To do

- [x] **Source maps** - opt-in `--sourcemap` for JS/CSS with external or inline maps
- [ ] **Full/general minifier** — current JS pass is conservative; decide
      whether to keep it or add a real minifier
- [ ] **Tree shaking** — remove unused generated runtime
- [ ] **Bundle splitting** — optional for large apps
- [ ] **`SUI_GEN_VERSION`** — version marker in generated code
- [ ] **PWA support** — manifest, service worker, installability
- [ ] **SSR/SSG** — optional static generation
- [ ] **WASM target** — long-term
- [ ] **`--csp` build option** — emit a CSP policy for the single-file build
- [ ] **Single-file asset policy** - document/validate external URLs in user CSS
      and remote data (inline build is not automatically offline)
- [ ] **Desktop packaging/installers** - MSI/DMG/AppImage/deb, app icon and
      bundle metadata, Windows code signing and macOS notarization
- [ ] **WebView2 runtime bootstrap/fallback** - detect or install the runtime
      and fall back to the browser target with a clear message
- [ ] **macOS/Linux file association** - double-click `.sui` outside Windows

## 5. Tooling (LSP, GUI, CLI)

### 5.1 CLI

- [x] `parse`, `fmt`, `check --json`, `build`, `build --single --minify`,
      `serve`, `gui`, `lsp`, `mcp`, `docs`
- [x] `gui [file.sui]` start-from-file and `simpleui <file.sui>` shortcut
- [x] `associate` Windows `.sui` file association with `--uninstall`/`--dry-run`
- [x] `--security` and strict security config
- [x] `--allow-dir`/`--allow-exec` for desktop builds
- [x] `docs --agent/--json/--list/--full/--read/--search`
- [ ] **`simpleui test`** — test runner for `.sui`
- [ ] **`simpleui new`/scaffolding** — project/interface templates
- [ ] **`simpleui docs --validate`** — link/manifest validation outside Go tests
- [ ] **Structured build/check output for all commands** — only `check` has JSON
      currently

### 5.2 GUI Designer

- [x] **Live inline validation** — `.cm-error`/`.cm-warn`, Problems panel
- [x] **Attribute/handler/state autocomplete**
- [x] **Go to definition** — F12
- [x] **Find references** — Shift+F12
- [x] **Rename symbol** — F2
- [x] **Bidirectional tree/preview/editor sync**
- [x] **Tree control end-to-end** — whole-tree selection, context-menu
      structural editing, leaf/parent/open icons, open/closed branch symbols,
      curated icon picker, `onTreeChange` and `tree.lastChange`
- [x] **Table control end-to-end** — declarative columns, sorting, filtering,
      pagination, row/column/cell selection, validated editing, table events,
      `table.lastChange`, a tabbed Table settings overlay, Excel-like bottom
      sheet tabs, runtime table creation and row/column structure editing
- [x] **Hover docs** — builtins/stdlib signatures/descriptions
- [x] **Toggle comments**, **Outline/Document Symbols**
- [x] **Table designer** — presets, selectors, grips, caption, editable/fixed,
      scroll/scale, sticky header, resizable inspector, data section
- [x] **Iframe designer** — `src`, `srcAllow`, `sandbox`, `allow`,
      `referrerpolicy`, `loading`, dimensions; runtime error UI
- [x] **Export `.sui`, `.zip`, single HTML**
- [x] **Desktop build dialog** - preflight, capabilities, log, download,
      browser/WebView target selector
- [x] **Project themes and contrast** - Light/Dark families, page/base elements,
      tooltips, media, hover, chart/graph presets, theme export/import
- [x] **Start from file** - `simpleui gui app.sui`, title, Save back to path
- [x] **Windows `.sui` association** - double-click launches the designer
- [ ] **Component library** — load a folder as a reusable palette
- [ ] **Theme/Design Tokens JSON** — export/import tokens
- [ ] **Multi-device preview** — side-by-side mobile/tablet/desktop
- [ ] **Visual state editor** — more complete state transitions
- [ ] **Asset manager** — full images/fonts/assets browser (font upload exists)
- [ ] **Export CSS variables** — sync SUI CSS with tokens
- [ ] **Open output folder** for desktop builds (local/server path today)

### 5.3 LSP

- [x] Basic completion + live diagnostics (parse/Check)
- [ ] Hover with type/docs
- [ ] Go to Definition
- [ ] Find References
- [ ] Document Symbols
- [ ] Rename
- [ ] Code Actions / quick fixes
- [ ] Semantic Tokens
- [ ] Multi-file workspace symbols

> The GUI has a richer editor feature set than the LSP; the LSP remains a
> minimal diagnostics/completion server.

### 5.4 Serve & Dev

- [ ] **HMR** — CSS/component updates without full reload
- [ ] **Preserve state on reload**
- [ ] **Browser error overlay** — compile/runtime errors
- [ ] **Live desktop build progress** — current dialog shows the final log

## 6. Testing & Quality

- [x] **Broad Go test suite** — parser, typechecker, JS/CSS, tables, security,
      docs links/anchors, iframe, desktop GUI preflight, single-file/minify
- [x] **Compile-time security tests** — escaping, URL, CSS, classes, iframe
- [x] **Documentation tests** — manifest coverage, embedded sync, links/anchors
- [x] **Manual browser checks** - single-file/minified apps, table, iframe,
      designer desktop build
- [x] **Theme contrast audit** - all Light/Dark variants and component text pairs
- [x] **Golden screenshots** - Standard Light/Dark palette grids; refresh with
      `SUI_UPDATE_THEME_GOLDEN=1`
- [x] **Headless desktop/server checks** - WebView target browser test and
      `--no-open` loopback check
- [ ] **`simpleui test`** — language-level test blocks/assertions
- [ ] **Headless runner** — deterministic DOM/browser test environment
- [ ] **Coverage** reporting
- [ ] **Fuzzing** — lexer/parser/typechecker
- [ ] **Golden/snapshot tests** for generated AST/JS/CSS
- [ ] **Benchmarks** for parse/typecheck/build
- [ ] **Cross-platform CI** — Windows/Linux/macOS matrix
- [ ] **Desktop executable e2e on CI** — run the produced binary headlessly
- [x] **Release workflow** — prebuilt binaries for Windows/Linux/macOS on amd64
      and arm64, one archive per target plus `SHA256SUMS`
      (`.github/workflows/release.yml`, triggered by a `v*` tag)
- [x] **MIT license** — `LICENSE` at the repository root
- [x] **Module path** — `github.com/salvatoresaba/simpleui`: the compiler is
      importable as a Go library

## 7. Security & Desktop Bridge

### 7.1 Completed

- [x] Automatic HTML escaping; no raw HTML interpolation
- [x] URL scheme filter + `*Safe` SSRF allow-lists/redirect blocking
- [x] `cssSafe` for dynamic styles
- [x] Dynamic-class allow-list
- [x] Inert table cells and CSV formula escaping option
- [x] Strict security configuration with BOM/unknown-field rejection
- [x] Iframe sandbox/allow-list/referrer defaults and runtime rejection UI
- [x] Bridge token, constant-time compare, loopback Host, same-site checks,
      symlink-safe allow-dirs, `X-Frame-Options`, `frame-ancestors`
- [x] Cross-site rejection on the desktop build endpoint
- [x] Token-authenticated desktop artifact download

### 7.2 To do

- [ ] **Bridge audit log** — structured op/path/allowed/denied
- [ ] **Exec timeouts + output cap + process kill**
- [ ] **`build --csp`** — configurable CSP for web deployments
- [ ] **Allowlist globs** — `--allow-dir-glob` with explicit warnings
- [ ] **Permission prompts** - per-operation consent UI
- [ ] **`postMessage` helper** - origin/source validation for iframe apps
- [ ] **Signed desktop builds** - optional code signing/notarization pipeline

## 8. Documentation & Ecosystem

### 8.1 Completed

- [x] **Documentation index** — `docs/INDEX.md`
- [x] **Agent guide** — `AGENTS.md`, `CLAUDE.md`, Copilot/Cursor pointers
- [x] **Machine manifest** — `docs/manifest.json`
- [x] **Recipes** — `docs/AGENT_RECIPES.md`
- [x] **`llms.txt`**
- [x] **Embedded docs + CLI** — `simpleui docs ...`
- [x] **MCP documentation resources/prompts** — `sui://docs/...`,
      `sui_docs`, build/review prompts
- [x] **Link/anchor/embedding tests**
- [x] **Examples index** — `examples/README.md`
- [x] **English canonical documentation policy** — canonical docs are English;
      Italian translations use the `_IT` suffix and are indexed;
      `TestCanonicalDocsAreEnglish` guards the canonical set
- [x] **Security/iframe/desktop docs** aligned with code

### 8.2 To do

- [ ] **Complete tutorial** — step-by-step from zero
- [ ] **Extended cookbook** — auth, pagination, error UX, deployment, packaging;
      `AGENT_RECIPES.md` covers the core patterns but not every scenario
- [ ] **Migration guide** — future breaking changes
- [ ] **Formal language reference** — grammar/BNF and exhaustive modifier table
- [ ] **Examples gallery** — screenshots/gifs and downloadable demos
- [ ] **Video tutorial** (optional)
- [ ] **Generated API docs from source** — avoid manual drift
- [ ] **MCP tools beyond docs** — `sui_fmt`, `sui_build`, AST/diagnostics
      resources; classify `sui_compile` output as assets vs. a disk build
- [ ] **`browser.preview` integration recipe** — standard review loop for agents
- [ ] **Changelog/release notes** — per version
- [ ] **Sync all `_IT` translations** — whenever a canonical doc changes, update
      its `_IT` counterpart in the same change

## 9. Next priorities

Suggested order (security and agent workflows first):

1. **Bridge exec timeouts/audit + process kill** — highest residual security gap
   in the desktop target.
2. **HTTP headers/multipart + WebSocket reconnect** — removes common remote-app
   blockers and misleading documentation.
3. **LSP advanced features** — hover, definition, references, symbols, rename;
   reuse the GUI implementations.
4. **Full minifier decision** - make `--minify` behavior explicit and decide
   whether to add a general-purpose JS minifier.
5. **Desktop packaging/signing** - installers, app icon/metadata, WebView2
   runtime bootstrap and signed/notarized builds.
6. **`simpleui test` + cross-platform CI** - make agent-generated interfaces
   continuously verifiable.
7. **User modules/imports** - the largest language feature still missing.
8. **Theme completeness** - link project themes to chart/graph nodes
   automatically and add per-element scrollbar overrides.
9. **Cross-platform file associations** - macOS/Linux double-click support.

## 10. Known product gaps

- **Project theme → chart/graph** is still manual: switching the project theme
  does not rewrite chart/graph `styleExpr` or `chartTheme`. Use the chart
  **Theme presets** or the unified chart settings panel.
- **Per-element scrollbar overrides** are not exposed; scrollbars are configured
  per project only.
- **Desktop WebView** still needs the platform web engine (WebView2, WKWebView
  or WebKitGTK); it is not a native widget toolkit.
- **No installer/signing/auto-update** yet: the desktop targets emit a bare
  executable.
- **Golden screenshots** cover Standard Light/Dark. All 24 theme variants are
  covered by the computed-style contrast audit, but not by pixel goldens.
- **File association** is implemented for Windows only.
- **Chart/graph project defaults** apply to existing nodes on import, but a
  project theme change still leaves chart styling untouched.

**Rule for future updates:** when an item is checked, add (or point to) a test
and update the relevant docs/manifest anchors in the same change. Canonical docs
are English; if an `_IT` counterpart exists, update it too.
