# MCP setup for coding agents

SimpleUI ships an MCP server (`simpleui mcp`) that speaks JSON-RPC over
stdio. It gives a coding agent the compiler itself: parse and type-check
sources, compile them, render a preview, read the embedded documentation and
validate the control appearance catalog.

This page is the setup reference for a developer who wants a coding agent to
build SimpleUI interfaces. For the language itself, see
[`API_REFERENCE.md`](API_REFERENCE.md); for the agent workflow, see
[`AGENTS.md`](../AGENTS.md).

## Requirements

- Go 1.24.2 or newer, to build the compiler.
- A coding agent or editor that can start an MCP server over stdio: Claude
  Code, Cursor, VS Code with the MCP support, or any other MCP client.
- A project that owns the `.sui` files. The server is stateless: every tool
  call carries the source to work on.

## Build the compiler

The server is the compiler binary itself, so build it first:

```bash
# Windows
go build -o simpleui.exe ./cmd/simpleui

# Linux and macOS
go build -o simpleui ./cmd/simpleui

simpleui version
# simpleui 0.1.0-alpha
```

Build from `./cmd/simpleui`: the repository root is the importable `sui`
package, so `go build` there writes a Go archive instead of an executable.

## Project configuration (.mcp.json)

Create `.mcp.json` in the root of the project that owns the `.sui` files:

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

Two command variants work:

| Variant | When to use it |
|---|---|
| `"command": "simpleui"` | The binary is on `PATH`. This is the normal setup once the compiler is installed. |
| `"command": "go", "args": ["run", "./cmd/simpleui", "mcp"]` | Working inside a fresh clone of the SimpleUI repository, without installing the binary. The first launch compiles the compiler, later launches reuse the build cache. |

On Windows use `simpleui.exe`, and use an absolute path when the install
directory contains spaces (`"command": "C:/tools/simpleui/simpleui.exe"`).

The repository root of SimpleUI itself ships a ready-to-copy
[`.mcp.json`](../.mcp.json) that uses the `go run` variant, so a clone of this
repository is MCP-ready without any setup.

## User configuration

To make the server available in every project, add the same server block to
the user-scope MCP configuration of the client instead of a project file.
Claude Code can do this from the command line:

```bash
claude mcp add --scope user simpleui -- simpleui mcp
```

Check the exact flags of your client with its own help output before relying
on this form; the important part is the server block above.

## Client-specific configuration

| Client | File | Shape |
|---|---|---|
| Claude Code | `.mcp.json` in the project root | `mcpServers` |
| Cursor | `.cursor/mcp.json` | `mcpServers` |
| VS Code (MCP support) | `.vscode/mcp.json` | `servers` |
| Any MCP client | client configuration | `mcpServers` with a stdio command |

VS Code uses a different key name:

```json
{
  "servers": {
    "simpleui": {
      "type": "stdio",
      "command": "simpleui",
      "args": ["mcp"]
    }
  }
}
```

Most clients ask for approval the first time they load a project-scoped
server. That prompt is expected.

## What the server exposes

**Tools**

| Tool | Purpose |
|---|---|
| `sui_check` | Parse and type-check SUI source; returns `ok` plus diagnostics. |
| `sui_model` | Return the structured model of the root view: title, states, functions, CSS and elements. |
| `sui_compile` | Compile SUI source and return the generated `index.html`, `style.css` and `app.js`. |
| `sui_render` | Compile SUI source and return a self-contained preview page. |
| `sui_docs` | Read the embedded documentation: list ids, read one document by id or path, search, or return the full bundle. |
| `sui_appearance` | Validate the control appearance catalog (`components/appearance.json`). |

**Resources**

- `sui://docs/manifest.json` — the machine-readable documentation index.
- `sui://docs/<path>` — every document listed in the manifest.

**Prompts**

- `build_simpleui_interface` — build or modify an interface with the security
  rules and recipes applied.
- `review_simpleui_security` — review an app for escaping, CSS, URL, iframe,
  table, CSV and desktop capability risks.

The server instructions tell the agent to read `AGENTS.md` first, to check
generated source with `sui_check` before building, and to validate the
appearance catalog before reloading the designer.

## Verify the connection

1. Start the server by hand: `simpleui mcp`. It waits for JSON-RPC on stdin
   and stays silent until it receives a request, so no output is the healthy
   case.
2. In the agent, ask it to list the MCP tools, or to check a small source
   file with `sui_check`.
3. Offline check without an agent: `simpleui docs --list` prints the document
   ids, `simpleui docs --read api` prints one document.

## Agent instructions for your project

Copy this block into the `AGENTS.md` of the project that owns the `.sui`
files, so the agent follows the same workflow even without the MCP server:

```markdown
# SimpleUI app

This project builds its user interface with SimpleUI (`.sui` files).

- Read the SimpleUI documentation before writing code: `sui_docs` with id
  `agents` for the workflow and `recipes` for interface patterns, or
  `simpleui docs --agent` from the shell.
- After every change to a `.sui` file run `simpleui check <file> --json` and
  fix every diagnostic before building.
- Build with `simpleui build <file> --out dist`, or
  `simpleui build <file> --single --out app.html` for one portable file.
- Never introduce raw HTML interpolation, always wrap a dynamic `style:` in
  `cssSafe`, use the `*Safe` HTTP and WebSocket variants with a literal
  allow-list for dynamic URLs, keep the iframe `sandbox` and `srcAllow`
  explicit, and write table cells with `table.display`.
- Do not invent APIs: check the documentation or run `simpleui check`.
```

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| The client reports that the command was not found | The client starts the server with its own `PATH`. Install the binary in a directory on `PATH`, or use an absolute path in the server block. |
| The server starts but no tools appear | The client may still be waiting for approval, or the server failed to launch. Check the client MCP log and start `simpleui mcp` by hand. |
| `go run` reports that the package cannot be loaded | The client started the server from another working directory. Use the built binary instead. |
| The first launch is slow | `go run` compiles the compiler once; later launches reuse the build cache. |
| The agent invents syntax | Point it at `AGENTS.md` and the recipes, or call `sui_docs` before writing code. |

## Security notes

- The server runs the compiler over the source you pass to it. `sui_compile`
  and `sui_render` execute the full compile pipeline, so review untrusted
  `.sui` sources before handing them to an agent.
- The MCP session does not relax the language security model. `sui_check`
  still enforces `cssSafe`, the `*Safe` URL variants, iframe `srcAllow`, inert
  table data and the desktop capability rules described in
  [`SECURITY.md`](../SECURITY.md).
- Prefer `sui_check` before `sui_compile`: the check step is where the
  security diagnostics surface.
- Desktop capabilities (`--allow-dir`, `--allow-exec`) belong to the build,
  not to the MCP session. The server never grants file or command access.

## Related documents

- [`AGENTS.md`](../AGENTS.md) — canonical agent workflow and rules.
- [`AGENT_RECIPES.md`](AGENT_RECIPES.md) — interface patterns for local,
  remote and desktop applications.
- [`API_REFERENCE.md`](API_REFERENCE.md) — language and API reference.
- [`SECURITY.md`](../SECURITY.md) — threat model and protections.
- [`INDEX.md`](INDEX.md) — documentation index.
- [`llms.txt`](../llms.txt) — compact discovery file for agents.
