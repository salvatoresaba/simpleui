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
| [`collections.sui`](collections.sui) | `map`, `filter`, `reduce`, `find`, `any`, `all`, `sortBy` and `unique` with declared callbacks. |
| [`syntax_ergonomics.sui`](syntax_ergonomics.sui) | Object/array destructuring, array/map spread and named function arguments. |
| [`enums.sui`](enums.sui) | Nominal enums and exhaustive `switch` with duplicate-case checks. |
| [`unions.sui`](unions.sui) | Union types, `string?` and `is` guards with branch narrowing. |
| [`literal_types.sui`](literal_types.sui) | String/number literal types, exhaustive switch and literal `is` guards. |
| [`generics.sui`](generics.sui) | Generic functions/components, rest parameters and inferred type arguments. |
| [`reactivity.sui`](reactivity.sui) | Computed values, batched updates, error boundaries and keyed lists. |
| [`intl_format.sui`](intl_format.sui) | Intl date/number/currency/percent/relative formatting and time zone. |
| [`stdlib_advanced.sui`](stdlib_advanced.sui) | Path, crypto, JSON Schema validation, storage quota and cross-tab sync. |
| [`stdlib_data.sui`](stdlib_data.sui) | Pure/in-memory std modules: bytes, binary views, file names, MIME, colors, units, print sizes, history, ZIP, CSV, read-only env, capabilities, local event buses, redacting logs and telemetry. |
| [`binary_bytes.sui`](binary_bytes.sui) | Binary documents: `http.getBytes` for a remote file, `files.readBytes` for a picked file and the format magic bytes. |
| [`pdfviewer.sui`](pdfviewer.sui) | Display a PDF document with the browser's native viewer: a remote document, a local file through the desktop bridge and a picked file. |
| [`fs_operations.sui`](fs_operations.sui) | Desktop fs.copy, fs.move/rename, fs.isDir and fs.size with capability checks. |
| [`xml_library.sui`](xml_library.sui) | XML parse/navigation/mutation, XPath, XSD subset, JSON/CSV, serialization and diff. |
| [`todo.sui`](todo.sui) | Local form + list interactions. |
| [`contacts.sui`](contacts.sui) | Local records and table-like layout. |
| [`table_data.sui`](table_data.sui) | `table.display` with arrays, objects, JSON, CSV, TSV and export. |
| [`table_sheets/`](table_sheets/README.md) | Excel-like workbook: bottom sheet tabs, declarative columns, sorting, filtering, pagination, selection, validated editing, runtime table creation and row/column structure editing. |
| [`table_layout/`](table_layout/README.md) | Structural HTML table for arranging cards and controls in columns; separate from the data `table` control. |
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
| [`motion_layout/`](motion_layout/README.md) | Keyed FLIP cards for filtered and reordered lists. |
| [`motion_pages/`](motion_pages/README.md) | Shared element transition across pages with measured start/end rectangles. |
| [`motion_graph/`](motion_graph/README.md) | Per-element state graph with click, state-variable and after-transition triggers. |
| [`motion_sequence/`](motion_sequence/README.md) | Infinite and finite loops, spring easing, emitted lifecycle events and multi-element orchestration. |
| [`motion_features/`](motion_features/README.md) | Guards, wheel/swipe/long-press gestures, viewport visibility and exit animation. |
| [`control_animations/`](control_animations/README.md) | CSS animation library on six controls, with a reduced-motion guard. |
| [`spinner_image/`](spinner_image/README.md) | CSS spinner vs animated GIF through `src`; includes the designer image/GIF workflow. |
| [`i18n/`](i18n/README.md) | Translation catalogs, plurals, locale switching, RTL and the designer Languages editor. |
| [`animation_lab/`](animation_lab/README.md) | Animatable controls only, with runtime **Style** and **Animation** selectors and new visual shapes. |
| [`prefab_controls/app.sui`](prefab_controls/app.sui) | Dropdown, popover, drawer, splitter, dropzone, timeline, empty state, skeleton, carousel, stepper and command palette. |
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
