# SimpleUI Documentation

> Single map for humans and AI agents. If this is your first time in the
> repository, start from [`README.md`](../README.md); if you are an agent, start
> from [`AGENTS.md`](../AGENTS.md).

SimpleUI compiles a `.sui` file into a self-contained web app or a desktop
executable. Documentation is split into **operational guides** (how to build an
interface), **reference** (syntax and APIs), and **specification** (internal
contracts). This page tells you what to read and in which order.

## Start here

| If you are... | Start with | Then |
|---|---|---|
| A person trying SimpleUI | [README](../README.md) | [Examples](../examples/README.md) and [API](API_REFERENCE.md) |
| A coding agent building an interface | [AGENTS.md](../AGENTS.md) | [Local/remote recipes](AGENT_RECIPES.md) |
| An agent using MCP | [AGENTS.md](../AGENTS.md#mcp-and-automation) | `simpleui mcp`, `sui://docs/...` resources |
| Fixing a security issue | [SECURITY.md](../SECURITY.md) | [Security examples](../examples/security/README.md) |
| Changing the compiler | [SPEC.md](../SPEC.md) | [DESIGN.md](../DESIGN.md) |

Recommended order for an agent:

1. `simpleui docs --agent` — compact context and rules.
2. `docs/manifest.json` — machine-readable index of documents and recipes.
3. `simpleui docs --read <id>` — full canonical document.
4. `simpleui check app.sui --json` — verify after every change.

## What are you building?

| Scenario | Recipe | Canonical docs | Example |
|---|---|---|---|
| Local form with state and table | [local-form-table](AGENT_RECIPES.md#local-form--reactive-table) | [API](API_REFERENCE.md), [STDLIB](STDLIB.md) | `examples/todo.sui`, `examples/table_data.sui` |
| Arrange UI regions in a structural table | [API: table layout](API_REFERENCE.md#table-layout) | [GUI editor: table layout](GUI_EDITOR.md#table-layout-control) | `examples/table_layout/README.md` |
| Build a menu bar / navigation | [API: menu bar](API_REFERENCE.md#menu-bar) | [GUI editor: menu bar](GUI_EDITOR.md#menu-bar-control) | `examples/menu_bar.sui` |
| Tabbed content | [API: tabs](API_REFERENCE.md#tabs) | [GUI editor: tabs](GUI_EDITOR.md#tabs-control) | `examples/tabs.sui` |
| Collapsible sections / FAQ | [API: accordion](API_REFERENCE.md#accordion) | [GUI editor: accordion](GUI_EDITOR.md#accordion-control) | `examples/accordion/README.md` |
| Toast feedback | [API: toast](API_REFERENCE.md#toast-notifications) | [GUI editor: toast](GUI_EDITOR.md#toast-control) | `examples/toast/README.md` |
| Hover/focus hint | [API: tooltip](API_REFERENCE.md#tooltip-property) | [GUI editor: tooltip](GUI_EDITOR.md#tooltip-property) | `examples/tooltip/README.md` |
| Searchable single choice, runtime options, wildcards | [API: combobox](API_REFERENCE.md#combobox) | [GUI editor: combobox](GUI_EDITOR.md#combobox-control) | `examples/combobox/README.md` |
| Native date/time/datetime picker | [API: date/time picker](API_REFERENCE.md#datetime-picker) | [GUI editor: date/time picker](GUI_EDITOR.md#datetime-picker-control) | `examples/datepicker/README.md` |
| Nested tree with rename, icons and styles | [API: tree](API_REFERENCE.md#tree) | [GUI editor: tree](GUI_EDITOR.md#tree-control) | `examples/tree/README.md` |
| Tree JSON/XML import and export | [STDLIB: tree](STDLIB.md#tree-module-tree-core) | [API: tree](API_REFERENCE.md#tree) | `examples/tree/README.md` |
| Charts from JSON/XML/CSV with presets, scroll and zoom | [API: chart](API_REFERENCE.md#chart) | [GUI editor: chart](GUI_EDITOR.md#chart-control) | `examples/chart/README.md` |
| Full project Gantt with WBS and actuals | [API: Gantt](API_REFERENCE.md#chart) | [GUI editor: chart](GUI_EDITOR.md#chart-control) | `examples/gantt_project/README.md` |
| Construction programme with permits and commissioning | [API: Gantt](API_REFERENCE.md#chart) | [GUI editor: chart](GUI_EDITOR.md#chart-control) | `examples/gantt_construction/README.md` |
| Graph / block diagram control | [API: Graph](API_REFERENCE.md#graph-control-block-diagrams) | [GUI editor: Graph](GUI_EDITOR.md#graph-control) | `examples/graph/README.md`, `examples/state_machine/README.md` |
| Update chart tags/colors at runtime, unique element ids, event-list polling, custom hints, table view/focus, element read/write API, static mode and shape/tag events | [Runtime guide](STATE_CHART_RUNTIME.md), [API: Graph](API_REFERENCE.md#graph-control-block-diagrams), [STDLIB: chart](STDLIB.md#chart-module-chart-core) | [GUI editor: Graph](GUI_EDITOR.md#graph-control) | `examples/process_monitor/README.md` |
| Multiple pages and navigation | [API: pages](API_REFERENCE.md#pages-and-navigation) | [GUI editor: pages](GUI_EDITOR.md#pages) | `examples/pages/README.md` |
| Modal dialogs | [API: modals](API_REFERENCE.md#modal-dialogs) | [GUI editor: modal](GUI_EDITOR.md#modal-control) | `examples/modal.sui` |
| Drag rows or drop/choose files/folders | [drag-and-drop](AGENT_RECIPES.md#drag-and-drop) | [API](API_REFERENCE.md#drag-and-drop), [BUILTINS](BUILTINS.md#drag-and-drop-functions), [GUI editor](GUI_EDITOR.md) | `examples/drag_drop.sui`, `examples/file_picker.sui` |
| Analyze selected files with a Go backend | [file-type-analyzer](AGENT_RECIPES.md#file-type-analysis-with-a-go-backend) | [STDLIB](STDLIB.md#http-module-http), [API](API_REFERENCE.md#drag-and-drop) | `examples/file_analyzer/README.md` |
| Read a remote JSON API | [remote-api-table](AGENT_RECIPES.md#remote-json-api--table) | [STDLIB](STDLIB.md#http-module-http), [SECURITY](../SECURITY.md) | `examples/users/app.sui` |
| Create/update data over HTTP | [remote-crud](AGENT_RECIPES.md#remote-crud-form) | [STDLIB](STDLIB.md#http-module-http), [SECURITY](../SECURITY.md) | `examples/users/app.sui` |
| Dashboard with periodic refresh | [polling-dashboard](AGENT_RECIPES.md#polling-dashboard) | [BUILTINS](BUILTINS.md#timer-functions), [API](API_REFERENCE.md) | `examples/monitor.sui` |
| Desktop files and commands | [desktop-files](AGENT_RECIPES.md#desktop-files-and-commands) | [SECURITY](../SECURITY.md#desktop-bridge), [STDLIB](STDLIB.md#file-system-module-fs) | `examples/security/desktop_files.sui`, `examples/desktop_picker.sui` |
| Import XLSX/ODS | [spreadsheet-data](AGENT_RECIPES.md#spreadsheet-data) | [Spreadsheet README](../examples/spreadsheet_table/README.md) | `examples/spreadsheet_table/app.sui` |
| Embed video/maps/documents | [iframe-embed](AGENT_RECIPES.md#iframe) | [API](API_REFERENCE.md#iframe), [SECURITY](../SECURITY.md#iframe-risks-in-detail) | `examples/security/untrusted_iframe.sui` |
| Review security | [security-review](AGENT_RECIPES.md#security-checklist) | [SECURITY](../SECURITY.md), [security examples](../examples/security/README.md) | `examples/security/*.sui` |

For the full list of recipes, tags and paths, use `docs/manifest.json` or
`simpleui docs --json`.

## Documentation map

### Getting started

- [README](../README.md) — installation, commands, concepts and layout.
- [AGENTS.md](../AGENTS.md) — agent workflow, rules and task routing.
- [CLAUDE.md](../CLAUDE.md) — pointer for Claude Code.
- [llms.txt](../llms.txt) — compact LLM discovery file.
- [Docs index](INDEX.md) — this page.
- [Hermes integration](../README.md#integration-with-hermes) — skills, plugins and VS Code extension.
- Italian translations: [README_IT](../README_IT.md), [AGENTS_IT](../AGENTS_IT.md),
  [SECURITY_IT](../SECURITY_IT.md), [SPEC_IT](../SPEC_IT.md),
  [DESIGN_IT](../DESIGN_IT.md), [TODO_IT](../TODO_IT.md),
  [API_REFERENCE_IT](API_REFERENCE_IT.md), [BUILTINS_IT](BUILTINS_IT.md),
  [STDLIB_IT](STDLIB_IT.md), [AGENT_RECIPES_IT](AGENT_RECIPES_IT.md),
  [GUI_EDITOR_IT](GUI_EDITOR_IT.md), [INDEX_IT](INDEX_IT.md).

### Language and API

- [API_REFERENCE](API_REFERENCE.md) — essential syntax, tables, iframes and CLI.
- [SPEC](../SPEC.md) — normative language, compiler and runtime contracts.
- [BUILTINS](BUILTINS.md) — functions that need no import.
- [STDLIB](STDLIB.md) — `http`, `ws`, `fs`, `exec`, `json`, `table` and the desktop `dialog` picker.
- [Chart table colors](TABLE_COLORS.md) — complete dark/light color audit with swatches and tokens ([visual version](table_colors.html)).
- [Gantt complete reference](GANTT_COMPLETE_REFERENCE.md) — consolidated data, scheduling, routing, keyboard, persistence, commit/version and server-sync reference.

### Data and interfaces

- [Tables and untrusted data](API_REFERENCE.md#tables-and-untrusted-data)
- [Table module](STDLIB.md#table-module-table-core)
- [Spreadsheet XLSX/ODS](../examples/spreadsheet_table/README.md)
- [Safe iframes](API_REFERENCE.md#iframe)

### Security

- [Threat model](../SECURITY.md#threat-model)
- [Language protections](../SECURITY.md#language-protections)
- [Security configuration](../SECURITY.md#security-configuration)
- [Iframe risks in detail](../SECURITY.md#iframe-risks-in-detail)
- [Residual risks](../SECURITY.md#residual-risks)
- [Security examples](../examples/security/README.md)

### Tooling and architecture

- [GUI editor: drag and drop](GUI_EDITOR.md) — designer DnD vs application DnD, inspector properties, files and pickers.
- [CLI](API_REFERENCE.md#cli)
- [Compiler design](../DESIGN.md)
- [Roadmap](../TODO.md)
- [MCP and automation](../AGENTS.md#mcp-and-automation)

### Examples

- [Examples index](../examples/README.md)
- [Security](../examples/security/README.md)
- [Spreadsheet table](../examples/spreadsheet_table/README.md)
- [Table layout](../examples/table_layout/README.md)

## Machine interfaces

SimpleUI exposes documentation in stable, automatable formats:

| Interface | Use |
|---|---|
| `docs/manifest.json` | JSON index of documents, sections, tags, invariants and recipes |
| `simpleui docs` | Print this page (also embedded in the binary) |
| `simpleui docs --json` | Full manifest |
| `simpleui docs --list` | Compact `id -> path` list |
| `simpleui docs --read api` | Read one document by id |
| `simpleui docs --search table` | Search titles, tags and summaries |
| `simpleui docs --agent` | Compact operational context for a coding agent |
| `simpleui docs --full` | All documentation in one stream |
| `simpleui mcp` | MCP tools and resources for agents (including docs) |

Client integration files: [`AGENTS.md`](../AGENTS.md),
[`CLAUDE.md`](../CLAUDE.md), [`llms.txt`](../llms.txt),
[`.github/copilot-instructions.md`](../.github/copilot-instructions.md),
[`.cursor/rules/simpleui.mdc`](../.cursor/rules/simpleui.mdc).

Agents can also read `AGENTS.md` and `docs/AGENT_RECIPES.md` directly; links
are relative to the repository root.

## Conventions

- **Source of truth**: `SPEC.md`, `SECURITY.md`, `API_REFERENCE.md`,
  `STDLIB.md`, `BUILTINS.md` and `docs/manifest.json`.
- **Task docs**: `AGENTS.md` and `docs/AGENT_RECIPES.md` explain *how* to build;
  the references explain *what* is available.
- **Do not invent APIs**: if a function is not in the references, check the
  source or compile with `simpleui check`.
- **Stable links**: prefer links to existing files and headings; the
  `docs_nav_test.go` test checks local links and manifest consistency.
- **Security first**: every example consuming external data must show the error
  path and protections (`cssSafe`, `*Safe`, `srcAllow`, `table`, desktop
  capabilities).
- **Language policy**: canonical documentation is English. Italian versions use
  the `_IT` suffix and must be updated whenever the canonical document changes.

