# Excel-like table workbook

A runnable demo with two named tables in **bottom sheet tabs**, declarative
columns, sorting, filtering, pagination, selection and validated editing.

## What it demonstrates

- `tabs position: "bottom" variant: "sheets"` — Excel-like tab strip under the
  panels.
- Two independent tables with their own `name`, data, sort/filter/page state and
  events.
- Declarative `columns`, `columnLabels`, `columnWidths`, `columnAlign`,
  `columnTypes`, `columnFormats` and `columnOptions`.
- `sortable` header sorting, `filterable` text filter and `pageSize`.
- Cell text modes selectable in the designer: **truncate** (default; stable
  column width, ellipsis and full text in the hover overlay) and **wrap**
  (stable column width, text wraps and the row grows).
- The **gear at the table's top-right** opens the runtime settings panel:
  header sorting and multi-column sort levels. Each header also has a funnel
  button that opens a searchable checklist of column values (Select All,
  Clear Filter, Apply). Toggle filters on, pick values, then add several sort
  levels in the panel to reproduce Excel's filter and multi-sort behavior.
- Row selection with `selection: "row"` and `multiSelect: true`.
- Editable order cells with number/currency validation.
- `editMode: "persistent"` on the Orders sheet: user edits are stored in
  `localStorage` and survive reload/restart. The Users sheet keeps the legacy
  read-only/selection behavior; switch it to `editMode: "editable"` to make it
  session-editable or `"persistent"` to keep its data across restarts.
- `onTableChange`, `onCellChange` and `table.lastChange`.
- CSV import through `table.importText`.
- Runtime **+** tab creation and row/column structure editing
  (`showAddTable` / `structureEditable`).
- The large tabbed **Table settings...** overlay in the designer.
- One workbook control: it fills its host with a 640 × 300 px minimum, the
  active sheet area follows the control size, the header stays sticky and wide
  columns scroll horizontally just above the sheet tabs.
- In the designer the elements panel shows a single **table** node for the whole
  workbook; clicking it outlines the outer control and shows the resize handles,
  while clicking a sheet tab only switches sheets (no extent box or handles).
- A single-sheet workbook hides its sheet tab; the designer always keeps the
  **+**, while a compiled app shows it only for an editable workbook. With two or
  more sheets the tabs appear. Double-click a tab or use its context menu to
  rename the workbook; the small **×** on the tab (or **Delete workbook** in its
  context menu) removes the workbook. In the compiled app the runtime adds the
  **×** to editable-sheet tabs and keeps the sheet removed for the page session,
  including across reactive re-renders.
- The small triangle in the first cell selects all cells; row/column selection
  strips straddle the cell border and live in the control padding.
- The row/column selection strips live in the control padding (default 10 px,
  adjustable in **Table settings → Scroll & scale → control padding**), so they
  never overlap cells or the control resize handles. The top-left corner selects
  all cells (all rows and columns together).

## Run

```bash
# from the repository root
./simpleui.exe check examples/table_sheets/app.sui --json
./simpleui.exe serve examples/table_sheets/app.sui --port 8080
```

Open `http://127.0.0.1:8080`.

Or build a single HTML file:

```bash
./simpleui.exe build examples/table_sheets/app.sui --single --out examples/table_sheets/app.html
```

## Source

```sui
# Excel-like workbook with two tables in bottom sheet tabs.
# Each table keeps its own name, data, sort/filter/page state and events.

state msg = "Use Reload demo data to fill both sheets."
state version = 0

fn changed() -> null {
	version = version + 1
	info, err = table.lastChange("orders")
	if err == null {
		msg = info
	}
}

fn loadDemo() -> null {
	let users = "id,name,status,amount\n1,Alice,active,1200.5\n2,Bob,paused,300\n3,Cara,active,99"
	n1, e1 = table.importText("users", users, "csv;header=1")
	let orders = "name,qty,price\nKeyboard,2,79.9\nMouse,4,24.5\nMonitor,1,249"
	n2, e2 = table.importText("orders", orders, "csv;header=1")
	if e1 != null {
		msg = e1
	} else if e2 != null {
		msg = e2
	} else {
		msg = "Loaded " + str(n1) + " users and " + str(n2) + " orders."
	}
}

view main {
	onLoad: loadDemo
	h2 "Workbook"
	p "Bottom sheet tabs, one table per sheet, declarative columns and validated editing."

	tabs "Workbook" position: "bottom" variant: "sheets" {
		tab "Users" {
			table name: "users" columns: "id,name,status,amount" columnLabels: "ID,Name,Status,Amount" columnWidths: "60,180,110,120" columnAlign: "right,left,center,right" columnTypes: "number,text,select,number" columnFormats: "text,text,text,currency:EUR" columnOptions: "status=active|paused" sortable: true filterable: true pageSize: 5 selection: "row" multiSelect: true showAddTable: true structureEditable: true onTableChange: changed {
			}
		}
		tab "Orders" {
			# Persistent edit mode: user edits to this sheet survive reload and restart.
			table name: "orders" columns: "name,qty,price" columnLabels: "Product,Qty,Price" columnWidths: "220,80,110" columnAlign: "left,right,right" columnTypes: "text,number,number" columnFormats: "text,number,currency:EUR" sortable: true filterable: true editMode: "persistent" onCellChange: changed onTableChange: changed {
			}
		}
	}

	row {
		button "Reload demo data" onClick: loadDemo
		p "Table revision: {version}"
	}
	p "{msg}"
}
```

## Runtime API used

```sui
table.importText("users", csv, "csv;header=1")
table.sort("users", "name", "asc")
table.filterText("users", "active")
table.page("users", 1)
table.setPageSize("users", 20)
table.selectedRows("users")
table.lastChange("orders")
```

See [`API_REFERENCE.md`](../../docs/API_REFERENCE.md#tables-and-untrusted-data)
and [`STDLIB.md`](../../docs/STDLIB.md#table-module-table-core).
