# SimpleUI TODO List

**Status:** 0.1.0-alpha  
**Last review:** 2026-10-25  
**Legend:** `[x]` implemented and verified · `[~]` partial / documented-only · `[ ]` missing

This file tracks the remaining work. A checked item must exist in the
compiler/codegen/runtime and be covered by a test or a manual browser check. Do
not mark an item complete only because it appears in documentation.

## 0. Current snapshot

Verified on 2026-10-25 with:

```bash
go build -o simpleui.exe .
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

Known accuracy corrections from the previous TODO:

- `http.get`/`http.post` custom headers and timeout override are **not** in the
  stdlib contract yet; only URL/body plus the internal safe-fetch timeout exist.
- `http.upload`, advanced `fs.*`, `crypto.*`, WebSocket reconnect and source
  maps are **not** implemented. Unsupported builtin/stub entries were removed
  from the public surface; do not document them as available.
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

- [ ] **Enums** — `enum Status { Idle, Loading, Error }` with pattern matching
- [ ] **Union types** — `string | int | null`, `T?` shorthand
- [ ] **Type guards / narrowing** — `if v is string { ... }`
- [ ] **Generics** — generic functions and components
- [ ] **Literal types** — unions of literals (`"success" | "error"`)
- [ ] **Const** — compile-time values
- [ ] **Readonly** — immutable struct fields/variables

### 1.3 Functions & Control

- [ ] **Rest parameters** — `fn sum(...nums: array[int])`
- [ ] **Named parameters** in addition to existing optional parameters
- [ ] **Destructuring** — `let {name, age} = user`, `let [a,b] = arr`
- [ ] **Spread operator** — `[...arr1, ...arr2]`, `{...obj1, ...obj2}`

## 2. Reactivity & Runtime

### 2.1 Reactive System

- [x] **Reactive class/style binding** — `class:` and `style:` accept
      expressions and update without ids
- [x] **Conditional expression (ternary)** — with type checking on both branches
- [ ] **Computed/derived state** — `computed doubled = count * 2`
- [ ] **Effects/watchers** — `effect { ... }`
- [ ] **Batching updates** — group multiple state assignments into one re-render
- [ ] **Fine-grained reactivity** — selective DOM updates

### 2.2 Advanced UI Components

- [ ] **Error boundaries** — subtree error fallback
- [ ] **List keys** — `list item in items key: item.id`
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
- [x] **JSON** — `json.parse`, `json.stringify`

### 3.2 Missing / partial

- [ ] **HTTP custom headers + timeout override** — signatures currently accept
      only URL/body; do not document headers as available
- [ ] **HTTP multipart/form-data** — not exposed; implement `http.upload` before
      documenting it
- [ ] **Advanced WebSocket** — reconnect and heartbeat/ping-pong; current `ws`
      keeps a single connection
- [ ] **Advanced filesystem** — `fs.copy`, `fs.move/rename`, `fs.isDir`,
      `fs.size` (not exposed until implemented)
- [ ] **Crypto** — `crypto.hash`, `crypto.uuid`, `crypto.randomBytes` (not
      exposed until implemented)
- [ ] **Path utilities** — `path.join`, `dirname`, `basename`, `extname`
- [ ] **JSON Schema validation** — structured validation
- [ ] **I18n helpers** — translation and localized formatting
- [ ] **Advanced storage** — quota/sync APIs are internal/stubbed, not a stable
      language contract

> Code review note: unsupported WebSocket/FS/Crypto/Multipart/Source Maps
> stubs and the `http.upload` builtin were removed from the generated/public
> surface. They must not be reintroduced without real implementations, typed
> stdlib entries, tests and documentation.

## 4. Code Generation & Build

### 4.1 Completed

- [x] **Desktop target** — one self-contained Go executable with embedded
      assets, loopback bridge, capability allow-lists and browser launch
- [x] **Single-file web export** — `build --single`, inlining CSS/JS/runtime
- [x] **Conservative minify** — `--single --minify` removes CSS comments/
      whitespace and JS full-line comments/indentation
- [x] **Designer single-file export** — `File → Export single HTML…`
- [x] **Designer desktop build** — dialog, preflight, capabilities, log, local
      output, optional download
- [x] **CSP/frame headers** — `serve` and desktop send
      `X-Frame-Options: DENY` + `frame-ancestors 'none'`

### 4.2 To do

- [ ] **Source maps** — SUI → JS/HTML mapping for debugging
- [ ] **Full/general minifier** — current JS pass is conservative; decide
      whether to keep it or add a real minifier
- [ ] **Tree shaking** — remove unused generated runtime
- [ ] **Bundle splitting** — optional for large apps
- [ ] **`SUI_GEN_VERSION`** — version marker in generated code
- [ ] **PWA support** — manifest, service worker, installability
- [ ] **SSR/SSG** — optional static generation
- [ ] **WASM target** — long-term
- [ ] **`--csp` build option** — emit a CSP policy for the single-file build
- [ ] **Single-file asset policy** — document/validate external URLs in user CSS
      and remote data (inline build is not automatically offline)

## 5. Tooling (LSP, GUI, CLI)

### 5.1 CLI

- [x] `parse`, `fmt`, `check --json`, `build`, `build --single --minify`,
      `serve`, `gui`, `lsp`, `mcp`, `docs`
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
- [x] **Hover docs** — builtins/stdlib signatures/descriptions
- [x] **Toggle comments**, **Outline/Document Symbols**
- [x] **Table designer** — presets, selectors, grips, caption, editable/fixed,
      scroll/scale, sticky header, resizable inspector, data section
- [x] **Iframe designer** — `src`, `srcAllow`, `sandbox`, `allow`,
      `referrerpolicy`, `loading`, dimensions; runtime error UI
- [x] **Export `.sui`, `.zip`, single HTML**
- [x] **Desktop build dialog** — preflight, capabilities, log, download
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
- [x] **Manual browser checks** — single-file/minified apps, table, iframe,
      designer desktop build
- [ ] **`simpleui test`** — language-level test blocks/assertions
- [ ] **Headless runner** — deterministic DOM/browser test environment
- [ ] **Coverage** reporting
- [ ] **Fuzzing** — lexer/parser/typechecker
- [ ] **Golden/snapshot tests** for generated AST/JS/CSS
- [ ] **Benchmarks** for parse/typecheck/build
- [ ] **Cross-platform CI** — Windows/Linux/macOS matrix
- [ ] **Desktop executable e2e on CI** — run the produced binary headlessly

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
- [ ] **Permission prompts** — per-operation consent UI
- [ ] **`postMessage` helper** — origin/source validation for iframe apps
- [ ] **Desktop WebView** — native window instead of system browser
- [ ] **Signed desktop builds** — optional code signing/notarization pipeline

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
4. **Source maps + full minifier decision** — improve debuggability and make
   `--minify` behavior explicit.
5. **Native WebView** — complete the desktop product experience.
6. **`simpleui test` + cross-platform CI** — make agent-generated interfaces
   continuously verifiable.
7. **User modules/imports** — the largest language feature still missing.

**Rule for future updates:** when an item is checked, add (or point to) a test
and update the relevant docs/manifest anchors in the same change. Canonical docs
are English; if an `_IT` counterpart exists, update it too.
