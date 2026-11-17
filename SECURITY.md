# SimpleUI Security

SimpleUI treats application data as untrusted and makes the dangerous paths
explicit. This document is the threat model, the available protections, the
security configuration, and the remaining risks.

## Threat model

| Source | Examples | Boundary |
|---|---|---|
| User text | forms, query params, storage | HTML escaping; data is never markup |
| Remote service | JSON/CSV bodies, WebSocket messages | validated as data; URL allow-lists for dynamic destinations |
| Files | CSV, JSON, XLSX/ODS, local files, dropped files/folders | parsing limits; metadata-only drops; bridge capabilities on desktop |
| URLs | `href`, `src`, `link`, iframe sources | scheme filter; `*Safe` variants for dynamic URLs |
| Desktop | commands, filesystem paths | capability allow-lists, loopback bridge, per-run token |

The goal is not to make an arbitrary remote page safe. It is to ensure that
data cannot become executable markup/CSS/URLs in the app, and that desktop
access is explicit and minimal.

## Language protections

### 1. Automatic HTML escaping

Interpolated values in text and attributes become `__sui_esc(...)`. Raw HTML
interpolation is intentionally unavailable. This applies to static text with
`{expressions}` as well.

### 2. URL filtering

`href`/`src`/`link` allow relative URLs and `http(s)`, `mailto`, `tel`.
`javascript:`, `vbscript:`, `data:` and other active schemes are neutralized at
runtime and rejected at compile time when static. One exception: `data:image/...`
is accepted in the `src` of `img` and `spinner`, where browsers render the value
as an image and never execute scripts; it stays rejected in `href`, `link` and
non-image media. Destination allow-lists remain the app's responsibility
(SSRF/open redirects).

### 3. Dynamic CSS requires `cssSafe(...)`

A non-static `style:` must be a `cssSafe(...)` call, including interpolated
style templates. `cssSafe` removes `url()`, `image-set()`, `paint()`, `attr()`,
`var()`, `expression()`, CSS escapes, positioning/overlay properties and
vendor-prefixed properties. Safe CSS custom properties (`--name`) are allowed,
but their values pass the same filters. Static developer-authored styles are
trusted, with one exception: author CSS cannot load external resources.
`css "..."`, static `style:` values and motion-graph state CSS reject
`@import` and `url(...)` values with network/file schemes (`http:`, `https:`,
`file:`, `ftp:`, `ws:`, `wss:` and protocol-relative `//`). Relative assets,
fragments and `data:` URIs remain allowed.

### 3b. Tree data import is inert

`tree.importJson`/`tree.importXml` treat payloads as node data. Labels, keys
and icons are written with `textContent` and attributes, never HTML. XML DTD and
entity declarations are rejected, parsing errors leave the current tree intact,
and imports are bounded by node/depth/label/key/text limits. Import replaces
only the node structure; application bindings and styling stay in place.

### 3c. Chart data is inert

`chart` treats its `data` attribute as untrusted text. JSON, XML and CSV are
parsed into numeric points; SVG nodes are created with `createElementNS` and
every label/value is written with `textContent`. No payload string reaches
`innerHTML`. XML DTD/entity declarations are rejected, malformed data shows a
short visible error, and parsing is bounded by a 5,000-point limit, a text-size
limit and a label-length limit. Chart data never loads a URL, script or stylesheet.

### 4. Dynamic classes are allow-listed

The compiler collects classes from app CSS and static class tokens/literals.
`class:` expressions use `__sui_class(...)`, which drops tokens outside that
set. Static classes are unchanged.

### 5. Compile-time checks

Static active URLs are errors by default; dynamic styles must use `cssSafe`;
dynamic URLs must use `*Safe` with a literal allow-list; unknown iframe
attributes/data are rejected; duplicate attributes are rejected.

### 6. Inert table data

`table.display`/`importText` write cells with `textContent`. JSON, CSV, TSV,
JSONL and converted XLSX/ODS cannot inject markup. Limits: 100,000 rows, 10,000
columns, 20 MB of text per call.

### 7. CSV export and formula injection

CSV follows RFC 4180 quoting. `formula=escape` neutralizes leading `= + - @`
TAB/CR to prevent spreadsheet formula execution when opened by Excel or
LibreOffice. It is opt-in; use it for untrusted exports.

### 8. No dynamic evaluation

Generated apps contain no `eval`/`new Function`; event handlers must be declared
functions; JSON uses `JSON.parse`.

### 8b. XML documents

- The XML parser rejects `DOCTYPE`/`ENTITY` declarations before `DOMParser`,
  preventing external/internal entity expansion and XXE. Documents are capped
  at 5,000,000 characters and 200,000 elements.
- `xml.loadUrlSafe` uses the same SSRF guard and literal allow-list as
  `http.getSafe`; the response size limit applies.
- XPath expressions, XSD schemas and XSLT stylesheets are trusted developer
  code. `xml.validateXsd` implements a practical subset (no `$ref`, schema
  imports or substitution groups); DTD validation is unsupported.
- The serializer emits XML text only. XML data is never interpolated into HTML,
  and JSON/CSV conversion treats values as data.
- `xml.xslt` depends on the browser `XSLTProcessor`; browsers that do not
  provide it return an explicit error.

### 9. Dynamic URLs and SSRF

`http.get/getStatus/post/postFiles` and `ws.connect` accept only literal URLs. Dynamic
destinations use `getSafe/getStatusSafe/postSafe/connectSafe` with a **literal**
allow-list. The runtime matches scheme/host/port (or a same-origin path prefix),
rejects credentials and active schemes, refuses redirects, applies a 30 s
timeout and a 10 MB response cap.

`http.postFiles` uploads the actual selected/dropped files as multipart data to
a literal URL; `http.postFilesSafe` adds the same literal allow-list for dynamic
destinations. The frontend cannot make a destination safe by itself: the
receiving service must enforce request size, file count, filename and content
limits. Do not upload to an endpoint that is not under the application's
control.

### 10. Iframes

- Default `sandbox=""` (no scripts/forms/popups/same-origin).
- Default `referrerpolicy="no-referrer"`.
- Dynamic `src` requires a literal `srcAllow` allow-list and is validated at
  runtime.
- Malformed/non-allow-listed URLs are not loaded; the runtime shows
  `Cannot accept the data: ...` in an alert.
- `allow-scripts` + `allow-same-origin` is an error by default.
- `srcdoc`, fallback markup and duplicate attributes are rejected.
- `serve` and desktop send `X-Frame-Options: DENY` and
  `Content-Security-Policy: frame-ancestors 'none'`.

### 11. Drag-and-drop metadata

Dropped or picked files are **metadata**, not content: the runtime reads
`name`, `path`, `size`, `type` and directory structure, but never the bytes.
The `dropFiles:` and `pickFiles:` targets require a string state; values are
rendered through the normal escaped state path. Directory traversal is bounded
(2,000 files, depth 16, 4,096 characters per path) and errors are shown in a
visible `Cannot accept the files: ...` box. `pickFiles` creates a hidden file
input for non-`file` triggers; it is removed after selection or cancellation.
`pickMode: "fs"` uses the File System Access API only after a user gesture and
permission grant, read-only, on Chromium secure contexts; handles are not
persisted across reloads. `auto` falls back to the native picker when the API is
unavailable.

Browsers do not expose absolute local paths. A name can be untrusted text and is
not evidence that a desktop path is safe to read; keep `fs` behind
`--allow-dir`. Validate any URL built from a dropped name.

### 12. Desktop bridge

- **Token**: random per-run 32-byte secret, constant-time comparison; `app.js`
  is not cacheable.
- **Loopback Host required** (`127.0.0.1`, `localhost`, `::1`): blocks DNS
  rebinding.
- **Same-site checks** and `X-Frame-Options: DENY`.
- **Exact command allow-list** (`--allow-exec`); commands run via
  `exec.Command` (no shell), so arguments are not interpreted by `cmd`/`sh`.
- **Directory allow-list** (`--allow-dir`) with symlink resolution: a link
  inside an allowed directory cannot escape. `fs.copy` and `fs.move` check both
  source and destination; `fs.copy` copies regular files only (no recursive
  copy), and `fs.size` rejects directories.
- **Crypto**: `crypto.hash` uses Web Crypto `subtle.digest` with SHA-256/384/512;
  `crypto.uuid`/`crypto.randomBytes` use the platform CSPRNG. There is no
  `Math.random` fallback. `crypto.status()` reports `ready`, `random-only` or
  `unavailable`; crypto calls throw a descriptive error instead of returning
  empty values, and `crypto.hash` requires a secure context (HTTPS or
  localhost). Hashes are not a password KDF.
- **JSON Schema validation**: `json.validate` treats the schema as trusted
  developer code and the data as untrusted. It supports a practical subset;
  `$ref` is rejected. `pattern` is developer-supplied RegExp, so avoid
  catastrophic patterns on huge untrusted strings. Validation is not a
  sanitizer: keep escaping data rendered in views.
- **Storage quota/sync**: `storage.quota` is advisory browser data;
  `storage.sync` listens for local cross-tab storage events and re-renders
  views. It is not a cloud sync or conflict-resolution service, and stored
  values remain untrusted data.
- **No bridge in web builds**: `serve` exposes static files only.
- The token is visible to a local process reading `app.js`; the bridge does not
  defend against a malicious local process, only against web content.

## Security configuration

Defaults are strict. A project can change compile-time policies with
`simpleui.security.json` (next to the `.sui` entry) or `--security <file>`:

```bash
simpleui build app.sui --security security.json
simpleui serve app.sui --security security.json
simpleui gui --security security.json       # otherwise ./simpleui.security.json
# lsp/mcp also use ./simpleui.security.json
```

Precedence: `--security` → `simpleui.security.json` next to the entry →
defaults. Malformed JSON, unknown fields or invalid values **fail the build**
with an explicit error. No configuration is silently accepted if it is invalid.

```json
{
  "style": { "dynamic": "safe" },
  "class": { "dynamic": "filter" },
  "urls": {
    "dynamic": "safe",
    "staticActiveSchemes": "error"
  },
  "table": {
    "maxRows": 100000,
    "maxCols": 10000,
    "maxText": 20000000
  },
  "csv": { "formulaEscapeByDefault": false },
  "iframe": {
    "sandbox": "default",
    "referrerPolicy": "no-referrer",
    "scriptsSameOrigin": "error",
    "dynamicSrc": "allowlist"
  }
}
```

| Field | Values | Default | Effect |
|---|---|---|---|
| `style.dynamic` | `safe`, `warn`, `allow` | `safe` | `safe`: dynamic `style:` requires `cssSafe`; `warn`: warning but compiles (CSS not sanitized); `allow`: no check |
| `class.dynamic` | `filter`, `allow` | `filter` | `filter`: `__sui_class` allow-list; `allow`: no filter (escaping only) |
| `urls.dynamic` | `safe`, `warn`, `allow` | `safe` | `safe`: dynamic URL requires `*Safe` + literal allow-list; `warn`/`allow`: permits `http.get(dynamicUrl)` |
| `urls.staticActiveSchemes` | `error`, `warn`, `allow` | `error` | Static active schemes in `href`/`src`/`link` or literal `http.get` |
| `table.maxRows/maxCols/maxText` | positive integers | 100000 / 10000 / 20000000 | Runtime table limits |
| `csv.formulaEscapeByDefault` | `true`/`false` | `false` | Default for `;formula=escape` in `table.exportText` |
| `iframe.sandbox` | `default`, `require`, `allow` | `default` | `default`: add `sandbox=""` if missing; `require`: error if missing; `allow`: add nothing |
| `iframe.referrerPolicy` | valid policy, `unset` | `no-referrer` | Add `referrerpolicy` when missing; `unset`: add nothing |
| `iframe.scriptsSameOrigin` | `error`, `warn`, `allow` | `error` | `allow-scripts allow-same-origin` combination |
| `iframe.dynamicSrc` | `allowlist`, `sandbox`, `allow` | `allowlist` | `allowlist`: dynamic `src` requires `srcAllow`; others do not |

**Not configurable** (security invariants):

- HTML escaping of text and attributes;
- the runtime URL-scheme filter for `href`/`src`/`link` (even with
  `staticActiveSchemes: "allow"`, a `javascript:` remains neutralized at runtime);
- runtime iframe URL validation and its visible error message;
- table cells written with `textContent`;
- bridge token, loopback Host, `X-Frame-Options`, symlink-safe allow-dirs;
- CSV quoting.

`warn` is intended for migration: the build continues and prints a warning.
`allow` disables the check, not the underlying runtime protection.

## Desktop bridge

Desktop apps are served on `127.0.0.1` and can use `fs`/`exec` only through the
protected bridge described above. `--allow-exec` is an exact command
allow-list; `--allow-dir` bounds filesystem access and resolves symlinks. The
bridge is absent in web builds.

The desktop `dialog.pickFile`/`pickFiles`/`pickDirectory` endpoints open the
native OS dialog (PowerShell/Windows Forms, `osascript`, or `zenity`/`kdialog`).
They are protected by the same per-run bridge token and require an explicit user
selection; the returned paths are not permission to read or execute. `fs` and
`exec` still enforce `--allow-dir` and `--allow-exec`. Cancellation returns an
empty value, not a path.

## Spreadsheet converter

`sheet2json`/`sheetconv` treat the spreadsheet as untrusted:

- no disk extraction: only known ZIP parts are read;
- 64 MiB per XML part and row/column/cell limits;
- `encoding/xml` does not resolve external entities (no XXE) and does not follow
  DTDs;
- server binds `127.0.0.1`, accepts only simple filenames (no path traversal,
  drives or separators) and only `.xlsx`/`.ods`;
- default output is JSON, which does not execute formulas; for CSV use
  `-safe`/`safe=1` when it will be opened by a spreadsheet;
- formulas are read from cached values, never executed.

## Iframe risks in detail

An `<iframe>` opens a separate browsing context (usually another origin) inside
the app. Risks include:

1. **Phishing/UI redressing** — content imitates app UI and captures input.
2. **App clickjacking** — an external site frames the app; headers/CSP above
   protect `serve`/desktop, static hosting needs equivalent headers.
3. **Top navigation** — a non-sandboxed frame can navigate the top window.
4. **Popups/downloads** — hostile content can open windows or downloads.
5. **Cookies/storage/same-origin** — `allow-same-origin` on same-origin content
   can expose app DOM/storage; `allow-scripts` + `allow-same-origin` may defeat
   the sandbox.
6. **Permissions** — camera, microphone, geolocation, clipboard, payments.
7. **`postMessage`** — untrusted messages need `origin`/`source` validation.
8. **Referrer leakage** — URL/query/tokens can leak; default `no-referrer`.
9. **SSRF** — an https iframe can target localhost/internal services; use
   `srcAllow` for dynamic sources.
10. **Spoofing** — the user cannot see the embedded origin; do not frame
    untrusted content.
11. **Malvertising/browser exploits** — sandbox/allow-lists reduce surface, not
    risk.
12. **DoS** — heavy pages consume CPU/memory; `loading="lazy"` helps only
    partially.

Implemented mitigations:

- `src`, `srcAllow`, `sandbox`, `allow`, `referrerpolicy`, `loading`, `width`,
  `height`;
- `sandbox=""` and `no-referrer` defaults;
- compile-time validation of URLs, sandbox tokens, permissions policy,
  referrer policy, loading and dimensions;
- dynamic `src` runtime validation with a visible rejection alert;
- literal `srcAllow` allow-list for dynamic sources;
- `allow-scripts` + `allow-same-origin` error by default;
- `srcdoc`/fallback content rejected.

Deployment still needs:

- CSP `frame-src`/`frame-ancestors` when assets are hosted outside
  `simpleui serve`;
- a `postMessage` helper for app-specific protocols;
- minimal sandbox tokens and trusted endpoints only.

## Residual risks

- **CSS**: static `style:` is trusted developer code; dynamic `cssSafe` is
  conservative and may drop legitimate properties. It is not a full CSS editor.
- **Classes**: filtering removes unknown classes, but a known app class can
  still change appearance (visual spoofing).
- **SSRF**: `*Safe` blocks hosts/redirects/schemes, but an allow-listed hostname
  can resolve to a private IP (DNS rebinding) or be compromised; CORS limits
  reading, not sending. Use narrow allow-lists and trusted endpoints.
- **Iframes**: sandbox is empty by default and dynamic URLs are validated, but
  the embedded page remains another origin and may be hostile. `srcAllow`
  checks the initial URL, not redirects inside the frame or DNS rebinding. If
  `allow-scripts`, `allow-same-origin`, `allow-popups` or permissions are
  enabled, risks return; `postMessage` and `frame-src` remain app/deployment
  responsibilities.
- **ReDoS**: regexes built from untrusted input can block the thread; limit
  length/patterns.
- **Prototype pollution**: do not merge untrusted JSON into shared objects;
  `merge` creates a new object but is not a sanitizer.
- **Generic DoS**: `json.parse`, loops and regex have no timeout; table limits
  cover only tabular payloads.
- **CSV**: quoting prevents structural injection; `formula=escape` is opt-in.
- **Symlink TOCTOU**: the bridge resolves symlinks before access, but a local
  process racing the filesystem remains a residual risk.
- **Desktop**: a malicious local process can read `app.js` and the bridge token;
  the bridge protects against web content, not the local user/OS.

## Examples

Runnable examples live in [`examples/security/`](examples/security/README.md):
untrusted text, CSS, classes, tables, payloads, URLs, iframes and desktop files.

## Reporting

Please report security issues privately to the maintainers before opening a
public issue. Include a minimal `.sui` reproduction and the exact command used.

