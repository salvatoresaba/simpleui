# SimpleUI examples

Runnable `.sui` sources. Start with `counter.sui`, then choose by task. Build any
example with:

```bash
simpleui build examples/counter.sui --out dist
simpleui serve examples/counter.sui --watch
```

For an agent-oriented walkthrough see
[`../docs/AGENT_RECIPES.md`](../docs/AGENT_RECIPES.md).

| Example | What it shows |
|---|---|
| [`counter.sui`](counter.sui) | Smallest reactive app: state, button, re-render. |
| [`reactive.sui`](reactive.sui) | Derived values and bindings. |
| [`todo.sui`](todo.sui) | Local form + list interactions. |
| [`contacts.sui`](contacts.sui) | Local records and table-like layout. |
| [`table_data.sui`](table_data.sui) | `table.display` with arrays, objects, JSON, CSV, TSV and export. |
| [`table_sheets/`](table_sheets/README.md) | Excel-like workbook: bottom sheet tabs, declarative columns, sorting, filtering, pagination, selection and validated editing. |
| [`chart/app.sui`](chart/app.sui) | Reactive JSON/XML/CSV charts, advanced types, compatibility checks, presets, scroll and zoom. |
| [`graph/`](graph/README.md) | Dedicated Graph control for block diagrams, with layout, tags, links, grid and synchronized table. |
| [`state_machine/`](state_machine/README.md) | Legacy `chartType: "state"` shape/link diagram with events, guards, actions, self-loop and synchronized shape/transition table. |
| [`process_monitor/`](process_monitor/README.md) | Wide process monitor with runtime tag colors/text, editing/static mode, static-mode shape/tag events and a modal opened from a shape click. |
| [`gantt_calendar/`](gantt_calendar/README.md) | Gantt holidays, weekend rules, IANA timezone, critical path and editable synchronized table (see its [README](gantt_calendar/README.md)). |
| [`gantt_wbs/`](gantt_wbs/README.md) | Hierarchical WBS summary tasks, collapsed roll-up bars, external dependencies and dependency-driven auto scheduling (see its [README](gantt_wbs/README.md)). |
| [`gantt_project/`](gantt_project/README.md) | Complete project plan: WBS, dependencies, actuals, effort, resources, costs, deadlines and calendar (see its [README](gantt_project/README.md)). |
| [`gantt_construction/`](gantt_construction/README.md) | Pump-station construction programme with permits, civil, mechanical, electrical and commissioning phases, execution data and critical path (see its [README](gantt_construction/README.md)). |
| [`table_selection/`](table_selection/README.md) | Excel-style row/column/cell selection, clipboard copy/paste, selected export, delete buttons and visible numeric steppers (see its [README](table_selection/README.md)). |
| [`menu_bar.sui`](menu_bar.sui) | Configurable menu bar with submenus, links and actions. |
| [`tabs.sui`](tabs.sui) | Working tab set with panels and keyboard navigation. |
| [`pages/`](pages/README.md) | Multi-page app with `navigate()`, shared state and a full walkthrough. |
| [`modal.sui`](modal.sui) | Configurable modal dialog with cancel/save buttons and `modal.open()`. |
| [`drag_drop.sui`](drag_drop.sui) | Reorder rows and drop files/folders onto a field. |
| [`file_picker.sui`](file_picker.sui) | Native file/directory dialog with `pickFiles` and a linked field. |
| [`file_analyzer/`](file_analyzer/README.md) | Full app: select/drop files, upload to a Go backend, analyze types in a table. |
| [`desktop_picker.sui`](desktop_picker.sui) | Desktop-only native file/folder dialogs returning absolute paths. |
| [`card.sui`](card.sui) | Components and slots. |
| [`wrapper.sui`](wrapper.sui) | Component composition. |
| [`switchdemo.sui`](switchdemo.sui) | Bound controls (`checked`, input state). |
| [`monitor.sui`](monitor.sui) | Polling dashboard with `interval` and error/loading states. |
| [`chat.sui`](chat.sui) | Chat-like state and list rendering. |
| [`users/app.sui`](users/app.sui) | Remote JSON API, validation, table rendering and diagnostics. |
| [`userdata_table/app.sui`](userdata_table/app.sui) | Desktop bridge + external command + table. |
| [`spreadsheet_table/app.sui`](spreadsheet_table/app.sui) | XLSX/ODS conversion into a table (see its [README](spreadsheet_table/README.md)). |
| [`security/untrusted_text.sui`](security/untrusted_text.sui) | Escaping of untrusted text and URLs. |
| [`security/untrusted_style.sui`](security/untrusted_style.sui) | `cssSafe` for dynamic CSS. |
| [`security/untrusted_class.sui`](security/untrusted_class.sui) | Dynamic class allow-list. |
| [`security/untrusted_table.sui`](security/untrusted_table.sui) | Inert table cells and CSV formula escaping. |
| [`security/untrusted_payload.sui`](security/untrusted_payload.sui) | JSON/CSV validation and limits. |
| [`security/url_guard.sui`](security/url_guard.sui) | `http.getSafe` and SSRF allow-lists. |
| [`security/untrusted_iframe.sui`](security/untrusted_iframe.sui) | Iframe sandbox, `srcAllow` and runtime rejection. |
| [`security/desktop_files.sui`](security/desktop_files.sui) | Desktop capability checks for files/commands. |

Security examples have a dedicated guide:
[`security/README.md`](security/README.md).

> Examples are source code, not generated output. Build them into a separate
> directory (`dist/` or a temp dir); do not edit generated assets.
