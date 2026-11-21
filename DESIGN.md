# SimpleUI Design Document

This is the original project design and rationale. The normative contracts live
in [`SPEC.md`](SPEC.md); current implementation status lives in
[`TODO.md`](TODO.md).

## 1. Vision and Core Principles

**Goal:** create a programming language for rapidly building graphical
interfaces for local desktop applications and remote web applications. The focus
is tools, dashboards and GUIs that wrap CLIs or real-time data.

1. **Zero dependencies:** users do not install Node.js, npm, Python or C++
   libraries. The compiler is a single binary.
2. **One codebase, two targets:** the same `.sui` source compiles to a desktop
   app or a static web app.
3. **Security first (static typing):** static typing with inference prevents
   runtime errors and makes security policies explicit.
4. **Data-driven development (implicit reactivity):** the programmer updates
   variables; the UI updates itself.
5. **Batteries included:** built-in HTTP, WebSocket, JSON and desktop
   filesystem/exec access.

## 2. Language Specifications

A `.sui` file mixes two logical concepts: **logic** (state, functions, modules)
and **views** (UI declaration).

- **Static with inference:** variables infer their type at declaration
  (`state counter = 0` is `int`); assigning a string later is a compile error.
- **Reactivity:** state must be declared with `state`; reassignment triggers a
  DOM re-render. State changes happen inside declared functions/handlers.
- **Async I/O:** I/O functions use `async`; the error convention is
  `result, error = function()`, with `error != null` on failure.
- **Imports:** the standard library is imported from `"std"`.

## 3. Compiler Architecture (Go)

The compiler receives `.sui` and produces the required artifacts. It is written
entirely in Go for cross-compilation and single-binary distribution.

### 3.1 Compiler phases

1. **Lexer:** tokenizes source.
2. **Parser:** validates grammar and builds the AST.
3. **Type checker:** validates types, signatures and security rules.
4. **Code generator:**
   - translates AST to vanilla JavaScript (logic);
   - translates AST to HTML/CSS (UI);
   - generates the Go wrapper/bridge for the desktop target.

### 3.2 Targets

- **Web:** output is `index.html`, `style.css`, `app.js`; `--single` can inline
  them into one HTML document.
- **Desktop:** output is one executable (e.g. `app.exe` or `app`). The compiler
  generates a Go `main.go` that embeds assets with `go:embed` and starts an
  invisible localhost server. The current implementation opens the system
  browser; a native WebView seam is documented in `SPEC.md`.

## 4. Logic and the Native Bridge

User logic is compiled to JavaScript. On desktop, JavaScript runs in the
browser context; filesystem/exec access goes through a bridge:

1. User code such as `output, err = exec.run("ls", ["-la"])` is translated by
   the compiler.
2. The generated JS sends a request to the local Go server with the per-run
   token.
3. The embedded server validates the capability allow-list, executes the real
   command via `os/exec`, and returns JSON.
4. No shell is involved; arguments are passed directly.

Bridge capabilities are granted only at build time (`--allow-exec`,
`--allow-dir`). The bridge token, loopback Host check, same-site checks and
symlink-safe directory resolution are mandatory.

## 5. Technical Solutions

### Problem 1: Implicit reactivity in vanilla JS

**Problem:** JavaScript is not reactive by default. How does the UI update when
a variable changes?

**Solution:** the Go code generator tracks `state` assignments and emits
`setState(...)` calls that re-render the affected views. The runtime binds views
to the states they read.

### Problem 2: Multiple return values in JS

**Problem:** `value, err = function()` has no direct JS equivalent.

**Solution:** async calls return a Promise resolving to an array; generated code
destructures `[value, err] = await ...`. Synchronous core functions return
arrays directly.

### Problem 3: Desktop security and permissions

**Problem:** a desktop app running local commands could be abused by web
content.

**Solution:** capability-based security: explicit command/directory
allow-lists, a per-run random token, loopback-only Host validation, same-site
checks, `X-Frame-Options`/`frame-ancestors`, and symlink-aware path checks. The
bridge never uses a shell.

## 6. Development Roadmap

The original phased roadmap (lexer/parser, static rendering, reactivity,
desktop wrapper and bridge, imports/stdlib, examples/alpha) is historical. See
[`TODO.md`](TODO.md) for the verified current status and next priorities.

## 7. Repository Layout (Recommended)

```
cmd/simpleui/          # CLI entry point
internal/lexer/
internal/parser/
internal/ast/
internal/typechecker/
internal/codegen/
internal/stdlib/
internal/desktop/      # embedded runtime and bridge
examples/              # .sui example applications
docs/                  # language/API/agent documentation
```

The actual repository keeps the Go compiler phases in the importable root
package `sui`, with the executable wrapper in `cmd/simpleui` and per-control
runtime/test packages under `components/<control>/`. `SPEC.md` is the
authoritative layout contract.
