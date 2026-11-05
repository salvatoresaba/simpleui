# Security examples

Runnable examples accompanying [`SECURITY.md`](../../SECURITY.md). Every file is
self-contained and compiles with `simpleui build`.

| Use case | Example | What it shows |
|---|---|---|
| Hostile text from API/WS/file | `untrusted_text.sui` | automatic text/attribute escaping; `javascript:` URLs blocked |
| Untrusted dynamic CSS | `untrusted_style.sui` | compiler requires `cssSafe(...)`; property/value sanitizer |
| Dynamic classes | `untrusted_class.sui` | only classes declared by the app survive |
| Hostile cells and CSV injection | `untrusted_table.sui` | `table.display` uses `textContent`; raw vs `;formula=escape` CSV |
| Arbitrary user JSON/CSV | `untrusted_payload.sui` | parse/validation errors, size limits, data displayed as text |
| Untrusted URLs (SSRF) | `url_guard.sui` | `getSafe`/`getStatusSafe` with a literal allow-list, no redirects |
| Desktop files/commands | `desktop_files.sui` | name validation + `--allow-dir` / `--allow-exec` capabilities |
| Untrusted iframes | `untrusted_iframe.sui` | default `sandbox`/`referrerpolicy`, `srcAllow`, runtime validation and visible error |

## Try it

```bash
# text, table and payload (web)
../../simpleui.exe serve untrusted_text.sui
../../simpleui.exe serve untrusted_style.sui
../../simpleui.exe serve untrusted_class.sui
../../simpleui.exe serve untrusted_table.sui
../../simpleui.exe serve untrusted_payload.sui
../../simpleui.exe serve untrusted_iframe.sui

# requires network; edit the allow-list in the example
../../simpleui.exe serve url_guard.sui

# desktop only: the bridge applies the permissions granted at build time
../../simpleui.exe build desktop_files.sui --target desktop --out dist \
  --allow-dir C:/data --allow-exec sheet2json.exe
```

## Security configuration

`simpleui.security.json` next to the `.sui` changes **compile-time** policies;
without it, strict defaults apply. To use another file:

```bash
../../simpleui.exe build untrusted_style.sui --security relaxed.security.json
```

`simpleui.security.json` (here) contains the defaults;
`relaxed.security.json` shows a permissive example. Values, precedence and
**non-configurable** protections are documented in
[`SECURITY.md`](../../SECURITY.md#security-configuration).

## Practical rules

- Untrusted data is **always data**: never concatenate it into HTML, URLs or CSS.
- `{...}` interpolation is escaped: injecting markup from views is neither
  needed nor possible.
- `href`/`src` pass through a scheme filter; for untrusted destinations still use
  an allow-list.
- When in doubt, `table.display` is the safest way to show rows: cells are text,
  never HTML.
- When exporting CSV for Excel/LibreOffice, use
  `table.exportText("table", "csv;formula=escape")`.
- On desktop, grant only the directories and commands actually needed; the
  bridge enforces the allow-list, but the app must still validate names.
- For an iframe, use explicit `sandbox` tokens only when needed; dynamic `src`
  requires `srcAllow`, and a rejected URL shows
  `Cannot accept the data: ...`.
