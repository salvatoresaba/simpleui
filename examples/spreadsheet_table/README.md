# Spreadsheet table (XLSX / ODS)

SimpleUI does not parse binary spreadsheet formats itself: conversion happens
**upstream**. `sheet2json` (Go, standard library only) opens `.xlsx` and `.ods`
and converts them to the text formats already supported by `table.display`:

- `json` — `{"sheet":"Data","columns":[...],"rows":[[...]]}` (default)
- `objects` — `[{"col":"value", ...}]`
- `jsonl` — one JSON object/array per line
- `csv` — RFC 4180 text with a configurable delimiter

## Web demo

1. put one or more `.xlsx` / `.ods` files in `data/`
2. run `run.ps1` (or `run.bat`)
3. open <http://127.0.0.1:8099/>

The script compiles `app.sui` into `web/`, starts the service and opens the
browser. The app calls `http.getSafe(url, "/api/")` and passes the response to
`table.display("data", body)`; the table preserves the style defined in
`app.sui`.

## CLI

```bash
# JSON with a header (ready for table.display)
sheet2json data.xlsx

# second sheet as CSV
sheet2json data.ods -sheet 2 -format csv

# list sheets
sheet2json data.xlsx -list

# objects keyed by the header
sheet2json data.xlsx -format objects

# no header: all rows are data
sheet2json data.xlsx -header 0
```

Flags: `-sheet <name|index>`, `-header 0|1`, `-format json|objects|csv|jsonl`,
`-sep ,|;|\||tab`, `-safe` (CSV: neutralize formulas), `-out file`.

## HTTP service

```bash
sheet2json -serve -port 8099 -dir ./data -web-dir ./web
```

- `GET /api/files` — list `.xlsx`/`.ods` files in the configured directory
- `GET /api/sheet?file=data.xlsx&sheet=1&format=json&header=1&sep=,&safe=1`
- only files in the configured directory; no path traversal

`json` is the format consumed directly by `table.display`. For the web target,
the app must be served from the same origin as the service (as `run.ps1` does)
or the service must expose CORS.

## Desktop alternative (`exec.run`)

When the app is a desktop build and the converter sits next to the executable:

```sui
import { exec } from "std"

async fn load() -> null {
    out, err = exec.run("sheet2json.exe", [path])
    if err != null { status = err; return }
    n, e = table.display("data", out)
    if e != null { status = e } else { status = str(n) + " rows" }
}
```

Build with `--allow-exec sheet2json.exe` (see `examples/userdata_table` for the
complete pattern).

## Security

The spreadsheet is untrusted input:

- no disk extraction (only the necessary ZIP parts are read);
- 64 MiB per XML part, 200,000 rows, 10,000 columns, 2,000,000 cells;
- `encoding/xml` does not resolve external entities: no XXE;
- server binds `127.0.0.1` only; simple filenames, only `.xlsx`/`.ods`;
- JSON output is safe; for CSV opened in Excel use `-safe`/`safe=1`
  (formula injection);
- formulas are read from cached values, never executed.

Details: [`SECURITY.md`](../../SECURITY.md).

## What is read

- All sheets (selected by name or 1-based index)
- Text (including rich text), numbers, booleans, errors, dates/times
- Excel dates converted through styles (`14`, `22`, custom formats)
- ODS repeated cells/rows (`number-rows-repeated`,
  `number-columns-repeated`) and covered cells
- ODS multi-paragraph text, tabs and spaces (`text:s`)

Known limitations:

- `.xls` (BIFF) is not supported: convert to `.xlsx`/`.ods`/CSV
- formulas return the value cached by the program that saved the file
- exotic date formats may remain serial numbers
- images, charts and formatting are not transferred; only values
- very large files are read entirely into memory
