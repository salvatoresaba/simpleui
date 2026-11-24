# SimpleUI Stdlib Documentation

SimpleUI's standard library is importable from `"std"`. Calls compile to
`__SUI_STD.<jsKey>(...)` and are implemented by the runtime. Every
**non-fire-and-forget** function returns **two values**: `value, err`
(`err = "null"` on success). Functions marked *fire-and-forget* return `null`
and do not require a full binding.

Available modules: `http`, `ws`, `sound`, `fs`, `exec`, `json`, `table`, `tree`, `chart`, `dialog` (desktop) and the pure/in-memory data modules documented in [STDLIB_DATA](STDLIB_DATA.md): `bytes`, `bin`, `files`, `mime`, `color`, `units`, `print`, `history`, `zip`, `archive`, `csv`, `env`, `capabilities`, `events`, `bus`, `log`, `telemetry`.

## Import

```sui
import { http, json } from "std"
import { fs } from "std"
```

## HTTP Module (`http`)

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `http.get(url)` | string | (string, string) | true | HTTP GET; returns the body as string |
| `http.getStatus(url)` | string | (int, string) | true | HTTP GET; returns the status code |
| `http.getBytes(url)` | string | (array[int], string) | true | HTTP GET; returns the body as raw bytes (binary-safe) |
| `http.getBytesSafe(url, allow)` | string, string | (array[int], string) | true | GET as raw bytes with SSRF allow-list |
| `http.post(url, body)` | string, string | (string, string) | true | HTTP POST with a string body (JSON/text) |
| `http.postFiles(url)` | string | (string, string) | true | POST the current selected/dropped files as `multipart/form-data` field `files` |
| `http.postFilesSafe(url, allow)` | string, string | (string, string) | true | File upload with SSRF allow-list |
| `http.getSafe(url, allow)` | string, string | (string, string) | true | GET with SSRF allow-list (exact scheme/host, no redirects) |
| `http.getStatusSafe(url, allow)` | string, string | (int, string) | true | Status code with SSRF allow-list |
| `http.postSafe(url, body, allow)` | string, string, string | (string, string) | true | POST with SSRF allow-list |
| `http.getWithHeaders(url, headers)` | string, any | (string, string) | true | HTTP GET with custom headers map/object |
| `http.getWithHeadersSafe(url, headers, allow)` | string, any, string | (string, string) | true | GET with headers and SSRF allow-list |
| `http.postWithHeaders(url, body, headers)` | string, string, any | (string, string) | true | HTTP POST with custom headers map/object |
| `http.postWithHeadersSafe(url, body, headers, allow)` | string, string, any, string | (string, string) | true | POST with headers and SSRF allow-list |
| `http.request(method, url, headers, body, timeoutMs)` | string, string, any, string, int | (string, string) | true | Custom HTTP method request with custom headers, body and timeout override in ms |
| `http.requestSafe(method, url, headers, body, timeoutMs, allow)` | string, string, any, string, int, string | (string, string) | true | Custom HTTP request with SSRF allow-list |
| `http.upload(url, fieldName, filename, content)` | string, string, string, string | (string, string) | true | Upload content as `multipart/form-data` |
| `http.uploadSafe(url, fieldName, filename, content, allow)` | string, string, string, string, string | (string, string) | true | Upload content with SSRF allow-list |
| `http.uploadWithHeaders(url, fieldName, filename, content, headers)` | string, string, string, string, any | (string, string) | true | Upload content as `multipart/form-data` with custom headers |
| `http.uploadWithHeadersSafe(url, fieldName, filename, content, headers, allow)` | string, string, string, string, any, string | (string, string) | true | Upload content with custom headers and SSRF allow-list |

**SSRF:** `http.get/getStatus/getBytes/post/postFiles/getWithHeaders/postWithHeaders/request/upload/uploadWithHeaders` accept only **literal** URLs. A dynamic URL
must use the `*Safe` variant with a **literal** allow-list (the compiler rejects
a variable). Entries are separated by commas/spaces:

- `https://api.example.com` (all paths) or `https://api.example.com/v1`;
- `/api/` (same-origin path prefix);
- WebSocket: `ws://host:port` / `wss://host:port`.

The guard rejects credentials, non-http(s)/ws(s) schemes, hosts/ports outside
the list and redirects; timeout 30 s, response max 10 MB. The requirement to use
`*Safe` for dynamic URLs is a compile-time policy configurable in
`simpleui.security.json` (default `safe`).

`http.postFiles` sends the **current selection** from `pickFiles:`/`dropFiles:`
as multipart data under the field name `files`. It also sends a parallel
`paths` field with the browser-visible relative paths, so a backend can display
folder/subfolder context. It uploads the actual file bytes; the backend must
enforce size/count limits and validate file names. Use `postFilesSafe` with a
literal allow-list when the destination is dynamic.

**Binary responses:** `http.get` and `http.post` decode the body as text, which
corrupts binary payloads such as PDF. Use `http.getBytes` (and
`http.getBytesSafe`) to receive the body as raw bytes (`array[int]`). The byte
variants return an error for a non-2xx status and reject responses over 10 MB;
the `*Safe` variant applies the same SSRF guard, timeout and size limit as the
text variants.

## WebSocket Module (`ws`)

Real WebSockets integrated with the runtime. Requires `import { ws } from "std"`.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `ws.connect(url)` | string | (any, string) | true | Open a WebSocket; returns the runtime instance |
| `ws.connectSafe(url, allow)` | string, string | (any, string) | true | WebSocket with a `ws://`/`wss://` allow-list |
| `ws.connectWithOptions(url, options)` | string, any | (any, string) | true | Open WebSocket with options map (`reconnect`, `delay`, `maxDelay`, `heartbeat`, `ping`) |
| `ws.connectWithOptionsSafe(url, options, allow)` | string, any, string | (any, string) | true | WebSocket with options and SSRF allow-list |
| `ws.send(ws, data)` | any, string | (null, string) | true | Send a message (fire-and-forget) |
| `ws.onMessage(ws, fn)` | any, string \| fn | (null, string) | true | Register a message handler (fire-and-forget) |
| `ws.onOpen(ws, fn)` | any, string \| fn | (null, string) | true | Register an open/connect handler (fire-and-forget) |
| `ws.onClose(ws, fn)` | any, string \| fn | (null, string) | true | Register a close/disconnect handler (fire-and-forget) |
| `ws.onError(ws, fn)` | any, string \| fn | (null, string) | true | Register an error handler (fire-and-forget) |
| `ws.status(ws)` | any | (string, string) | false | Get connection status: `"open"`, `"connecting"`, `"closing"`, `"closed"` |
| `ws.reconnect(ws)` | any | (null, string) | true | Force an immediate reconnect (fire-and-forget) |
| `ws.close(ws)` | any | (null, string) | true | Close the connection explicitly and cancel timers (fire-and-forget) |

Notes:

- `ws.send`, `ws.onMessage`, `ws.onOpen`, `ws.onClose`, `ws.onError`, `ws.reconnect` and `ws.close` are fire-and-forget (`ret0 = "null"`).
- `ws.onMessage` can be state-driven: update a state and trigger a re-render.
- Reconnect and heartbeat are supported via `ws.connectWithOptions`: set `reconnect: "true"`, `reconnectDelay`, `maxReconnectDelay`, `heartbeat` (ms) and `ping` (message). Manual reconnect is available via `ws.reconnect(conn)`. Status is inspectable via `ws.status(conn)`.

## Sound Module (`sound`)

Sound effects played by the browser. Requires `import { sound } from "std"`.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `sound.play(url)` | string | (bool, string) | true | Play the audio file at a **literal** URL; `true` when the browser accepted the playback |
| `sound.playSafe(url, allow)` | string, string | (bool, string) | true | Play the audio file at a dynamic URL, guarded by a **literal** allow-list |

Notes:

- `sound.play` accepts only a literal URL: a dynamic one is a compile-time
  error and the `*Safe` variant is required, exactly like `http.getSafe`.
- The URL guard is the same one of the HTTP module: only `http`/`https`, no
  credentials in the URL, no redirects, and the allow-list of `sound.playSafe`
  must be a literal string.
- The function resolves when the browser has **started** the playback, not
  when the sound ends. A browser that blocks the autoplay policy (a sound
  played without a user gesture) returns `(false, error)`: show the error,
  do not treat it as a crash.
- The runtime creates a detached `Audio` element for every call: the app does
  not need an `audio` control in the view.

```sui
import { sound } from "std"

async fn onDownloadDone() -> null {
	ok, err = sound.play("https://example.com/done.wav")
	if err != null {
		notifyError("the browser blocked the sound: " + err)
	}
}
```

## File System Module (`fs`)

Filesystem access is available through the **desktop bridge** with capability
security (`--allow-dir`). Some operations may be unavailable in a web context.
All functions are async, except `fs.documentUrl`, which only builds a URL.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `fs.read(path)` | string | (string, string) | true | Read a file as string |
| `fs.write(path, data)` | string, string | (null, string) | true | Write a file (fire-and-forget) |
| `fs.exists(path)` | string | (bool, string) | true | Whether the path exists |
| `fs.stat(path)` | string | (string, string) | true | File/dir info (descriptive string) |
| `fs.mkdir(path)` | string | (null, string) | true | Create a directory (`mkdir -p` style; fire-and-forget) |
| `fs.remove(path)` | string | (null, string) | true | Remove a file/directory (fire-and-forget) |
| `fs.list(path)` | string | (array[string], string) | true | List directory entries |
| `fs.copy(src, dst)` | string, string | (null, string) | true | Copy one file; both paths must be allowed |
| `fs.move(src, dst)` | string, string | (null, string) | true | Move/rename a file or directory; both paths must be allowed |
| `fs.rename(src, dst)` | string, string | (null, string) | true | Alias of `fs.move` |
| `fs.isDir(path)` | string | (bool, string) | true | Whether the path is an existing directory |
| `fs.size(path)` | string | (int, string) | true | Size in bytes of a file (directories are rejected) |
| `fs.readBytes(path)` | string | (array[int], string) | true | Read a file as raw bytes |
| `fs.writeBytes(path, bytes)` | string, array[int] | (null, string) | true | Write raw bytes (base64 over the bridge) |
| `fs.documentUrl(path)` | string | (string, string) | false | URL of the local document endpoint for a PDF file (desktop only) |

Both source and destination are checked against the `--allow-dir` capability
list. `fs.copy` copies regular files only; create the destination directory
first and use `fs.move` for directories. On the web target these calls return
an error.

`fs.documentUrl(path)` returns the URL that the `pdfviewer` control loads: the
desktop bridge serves the file from the allowed directories and only as a PDF
document. The URL carries no secret; the endpoint is protected by a per-run
cookie. Use it with a literal allow-list:

```simpleui
url, err = fs.documentUrl("docs/report.pdf")
if err == null {
	docUrl = url
}
```

```simpleui
pdfviewer src: docUrl srcAllow: "/"
```

## Exec Module (`exec`)

Command execution through the **desktop bridge** with an allow-list
(`--allow-exec`).

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `exec.run(cmd, args)` | string, array[string] | (string, string) | true | Run a command. `cmd` is resolved next to the executable too. Returns combined stdout/stderr |

**Example (see `examples/userdata_table/`):**

```sui
import { exec, json } from "std"
async fn load() {
  out, err = exec.run("userdata.exe", [])
  if err != null { ... return }
  data, perr = json.parse(out)
  ...
}
```

## JSON Module (`json`)

```sui
import { json } from "std"
```

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `json.parse(s)` | string | (any, string) | true | Parse JSON text; rejects a leading BOM and large input |
| `json.parseBytes(bytes, [encoding])` | array[int], string? | (any, string) | false | Parse UTF-8/UTF-16LE/UTF-16BE bytes; BOM detected automatically |
| `json.stringify(value, [options])` | any, map? | (string, string) | true | Serialize JSON. Options: `indent` (bool/int/string), `sortKeys`, `newline` |
| `json.toBytes(value, encoding, [options])` | any, string, map? | (array[int], string) | false | Serialize to UTF-8/UTF-16LE/UTF-16BE bytes |
| `json.validate(value, schema)` | any, any | (bool, string) | false | Validate with a JSON Schema subset; returns the first error path |
| `json.validateAll(value, schema)` | any, any | (array[string], string) | false | Validate and collect up to 100 errors |
| `json.pointer(value, pointer)` | any, string | (any, string) | false | Read a JSON Pointer (RFC 6901) |
| `json.setPointer(value, pointer, next)` | any, string, any | (any, string) | false | Return a copy with the pointer set or added |
| `json.removePointer(value, pointer)` | any, string | (any, string) | false | Return a copy with the pointer removed |
| `json.clone(value)` | any | (any, string) | false | Deep copy; rejects cycles and excessive nesting |
| `json.equal(a, b)` | any, any | (bool, string) | false | Deep JSON equality (array order matters) |
| `json.diff(a, b, [options])` | any, any, map? | (array[any], string) | false | Produce RFC 6902 operations. Options: `key`, `detectMoves` |
| `json.patch(value, operations)` | any, any | (any, string) | false | Apply RFC 6902 operations to a copy; aborts on the first error |
| `json.merge(target, patch)` | any, any | (any, string) | false | Apply an RFC 7386 JSON Merge Patch |
| `json.merge3(base, ours, theirs, [options])` | any, any, any, map? | (array[string], string) | false | Three-way merge into `ours`; returns conflict descriptions |
| `json.loadFile(path, [encoding])` | string, string? | (any, string) | true | Desktop bridge load under `--allow-dir` |
| `json.loadBytes(path, [encoding])` | string, string? | (any, string) | true | Desktop binary load with BOM/encoding detection |
| `json.saveFile(path, value, [options])` | string, any, map? | (null, string) | true | Desktop bridge save; UTF-8 text |
| `json.saveBytes(path, value, encoding, [options])` | string, any, string, map? | (null, string) | true | Desktop binary save; UTF-8/UTF-16LE/UTF-16BE |

### JSON Pointer (RFC 6901)

Pointers address exactly one value. `""` is the whole document, `/store/books/1`
selects an array element, and `/a~1b` addresses the object key `a/b` (`~1` is
`/`, `~0` is `~`). Missing paths return `(null, error)`; a stored `null` value
returns `(null, null)` and is therefore distinguishable from a missing path.
`setPointer` replaces an existing value, creates a missing object member, or
appends an array element; it never mutates its input. `removePointer` rejects
the document root.

### JSON Patch (RFC 6902)

`json.patch` supports `add`, `remove`, `replace`, `move`, `copy` and `test`.
Operations are applied to a deep clone, in order, and the input is never
partially modified: any error aborts and returns `(null, message)`. `test`
enables optimistic concurrency checks.

```sui
patched, err = json.patch(doc, [
  {"op": "replace", "path": "/store/open", "value": false},
  {"op": "add", "path": "/store/books/-", "value": {"id": 3, "title": "C"}},
  {"op": "test", "path": "/store/open", "value": false}
])
```

### Diff and three-way merge

`json.diff` emits directly applicable RFC 6902 operations. Without `key` it
compares arrays positionally. With `{"key": "id"}` it aligns array items by that
member, emits `move` for reordering and `add`/`remove` for membership changes.
Each operation also carries an optional `key` field for inspection; standard
clients ignore it. Structural operations are emitted before nested value
changes so their paths stay valid. `detectMoves: false` replaces a reordered
array wholesale instead of emitting moves.

`json.merge` implements RFC 7386: objects merge recursively, `null` deletes a
member, and arrays/scalars replace the target value. `json.merge3` computes
`base -> theirs`, applies non-conflicting changes to `ours` in place and
returns conflict descriptions such as `replace /items/0/value`. Conflicting
nodes keep the `ours` value. The root of `ours` must be an object or array.

### Validation

`json.validate` returns the first violation; `json.validateAll` returns up to
100 messages. The schema may be a map or a JSON string. Supported keywords:
`type`, `enum`, `const`, `properties`, `required`, `additionalProperties`,
`items` (schema or tuple), `additionalItems`, `minItems`/`maxItems`,
`uniqueItems`, `minLength`/`maxLength`, `pattern`, `format` (`email`, `uri`,
`url`, `uuid`, `date`, `date-time`), numeric bounds, `multipleOf`,
`minProperties`/`maxProperties`, `allOf`/`anyOf`/`oneOf`/`not` and
`if`/`then`/`else`. `$ref` is rejected. The schema is trusted developer code;
data is never executed.

### Bytes, encoding and file helpers

`json.parseBytes` and `json.toBytes` support UTF-8, UTF-16LE and UTF-16BE,
detect or write a BOM, and reject UTF-32 with an explicit error. JSON text is
limited to 20 million characters/bytes, nesting to 128 levels, patches to
10,000 operations and diffs to 10,000 changes. Desktop file helpers use the
same allow-list as `fs` and `xml`; on the web they return an explicit error.
The object key `__proto__` is rejected by patch, diff and clone to avoid
prototype pollution.

## XML Module (`xml`)

`import { xml } from "std"`. The module uses the browser XML DOM
(`DOMParser`, `XMLSerializer`, XPath and `XSLTProcessor`) on web and desktop.
All functions follow the std convention: bind both values with `a, b = ...`.
Documents and nodes are typed `xmlDocument` / `xmlNode`; a document is
assignable to a node.

```sui
import { xml } from "std"

doc, err = xml.parse("<people><person id=\"1\"><name>Ada</name></person></people>")
root, rerr = xml.root(doc)
nameEl, nerr = xml.find(root, "name")
name, textErr = xml.text(nameEl)
out = name
```

### A. Loading and syntax

| Function | Params | Returns | Async | Description |
|---|---|---|---|---|
| `xml.parse(text)` | string | (any, string) | false | Parse a string; rejects DOCTYPE/ENTITY |
| `xml.parseBytes(bytes, [encoding])` | array[int], string | (any, string) | false | Decode bytes using BOM, the XML declaration encoding, or an explicit encoding |
| `xml.loadFile(path)` | string | (any, string) | true | Desktop bridge file load under `--allow-dir` |
| `xml.loadBytes(path, [encoding])` | string, string? | (any, string) | true | Desktop binary file load; honours BOM/declaration/encoding |
| `xml.loadUrlSafe(url, allow)` | string, string | (any, string) | true | Fetch XML through the SSRF guard; `allow` must be a literal origin/path list |
| `xml.validate(text)` | string | (bool, string) | false | Syntax check; the error includes line/column when the browser provides them |

Limits: 5,000,000 characters and 200,000 elements. `DOCTYPE` and `ENTITY`
declarations are rejected before parsing. True streaming and incremental
loading are not available; the complete document is materialized.

### B. Navigation

| Function | Description |
|---|---|
| `xml.root(doc)` | Document element |
| `xml.parent(node)` | Parent node (the document for the root) |
| `xml.children(node)` | Element children |
| `xml.childNodes(node)` | Every child node (text, comments, CDATA, PI, elements) |
| `xml.childCount(node)` | Number of element children |
| `xml.next(node)` / `xml.previous(node)` | Next/previous element sibling |
| `xml.find(node, name)` | First descendant element with that tag |
| `xml.findNs(node, localName, namespaceURI)` | Namespace-aware descendant search |
| `xml.findAll(node, name)` | All descendant elements with that tag |
| `xml.descendants(node)` | All descendant elements |
| `xml.has(node, name)` | Whether a descendant exists |

### C. Reading data

| Function | Description |
|---|---|
| `xml.type(node)` | `document`, `element`, `text`, `comment`, `cdata`, `pi`, ... |
| `xml.name(node)` / `xml.localName(node)` | Qualified/local name |
| `xml.namespace(node)` / `xml.prefix(node)` | Namespace URI and prefix |
| `xml.text(node)` | **Direct** Text/CDATA children only |
| `xml.content(node)` | Aggregated `textContent` of the subtree |
| `xml.attr(node, name)` / `xml.attrNs(node, local, ns)` | Attribute value or `null` |
| `xml.attrs(node)` | Map of all attributes |
| `xml.path(node)` | Absolute path such as `/people/person[1]/name` |

`text` and `content` are intentionally different: use `text` for the value
directly inside an element and `content` when descendants must be included.

### D. Mutation

Element and tree operations:

| Function | Description |
|---|---|
| `xml.createDocument([rootName])` | New empty XML document |
| `xml.createElement(doc, name)` / `xml.createElementNs(doc, ns, qname)` | New element |
| `xml.createText(doc, text)` / `xml.createComment` / `xml.createCdata` / `xml.createPi` | Other node types |
| `xml.parseFragment(doc, text)` | Parse nodes without a single root (returns a fragment) |
| `xml.setInnerXml(node, text)` | Replace children with parsed XML fragment content |
| `xml.replaceChildren(node, nodes)` | Replace children with an array of nodes |
| `xml.append(parent, node)` / `xml.prepend` | Add a child at the end/start |
| `xml.insertBefore(parent, node, ref)` / `xml.insertAfter` | Insert relative to a child |
| `xml.remove(node)` / `xml.replace(old, new)` / `xml.clear(node)` | Remove, replace or empty |
| `xml.clone(node, [deep])` | Copy a node |
| `xml.rename(node, name)` | Rename an element (returns the new node) |
| `xml.setText(node, text)` | Replace children with one text node |
| `xml.move(node, parent)` | Move a node under another parent |
| `xml.reorder(parent, indexes)` | Reorder element children using a permutation of indexes |

Attribute operations: `xml.setAttr`, `xml.setAttrNs`, `xml.removeAttr`,
`xml.renameAttr`, `xml.hasAttr`. Nodes from another document are imported
automatically when inserted.

### E. Search and XPath

- `xml.find`, `xml.findNs`, `xml.findAll`, `xml.descendants`, `xml.findAttr(node, name, [value])` (omitted or `null` matches any value)
- `xml.findText(node, term, [regex])` — direct text search, optional RegExp
- `xml.xpath(node, expression, [namespaces])` — XPath 1.0; namespaces is a map of
  prefix → URI. Returns an array for node sets, or a string/number/bool for
  scalar expressions.
- `xml.xpathOne(node, expression, [namespaces])` — first node or scalar value.

```sui
names, xerr = xml.xpath(doc, "//person[@id='2']/name", {})
```

### F. Validation

| Function | Description |
|---|---|
| `xml.validate(text)` | XML syntax |
| `xml.validateXsd(doc, xsdText)` | XSD subset; returns `(false, firstError)` for the first violation |
| `xml.validateXsdAll(doc, xsdText)` | XSD subset; returns `(errors, null)` with up to 100 messages |
| `xml.validateRules(doc, rules)` | Custom rules by element name: required children/attributes, min/max counts, field types |

DTD validation is not supported: DOCTYPE/ENTITY declarations are rejected for
security. XSD support is a practical subset, not the complete XSD 1.0/1.1
specification. It covers global/local elements with `ref`, `sequence`/`all`/
`choice`, `minOccurs`/`maxOccurs`, attributes, attribute groups and global
attributes, `simpleType` restrictions with enumeration, length/pattern and
numeric facets, `complexType` with `simpleContent`/`complexContent` extension
and restriction, namespaces via `targetNamespace`, and content models inherited
through `complexContent`. Not implemented: `$ref` to schema documents or
elements, substitution groups, `key`/`keyref`/`unique`, wildcards beyond
`anyAttribute`, `xs:anyType` derivation, and schema imports/includes.
Namespace checks are available through `xml.namespace`, `xml.findNs` and XSD
`targetNamespace`.

### G. Export and saving

| Function | Description |
|---|---|
| `xml.stringify(node, [options])` | Serialize to string. Options: `indent` (bool/int), `declaration`, `encoding`, `sortAttributes` |
| `xml.toBytes(node, encoding, [options])` | Serialize to UTF-8/UTF-16LE/UTF-16BE bytes |
| `xml.saveFile(path, node, [options])` | Desktop bridge save under `--allow-dir`; UTF-8 only |
| `xml.saveBytes(path, node, encoding, [options])` | Desktop binary save; UTF-8/UTF-16LE/UTF-16BE |

Comments, CDATA, processing instructions and namespaces are preserved by the
DOM serializer. `indent: true` produces readable output; omit it for compact
XML. Incremental/dirty-region saving is not implemented: the document is
serialized again.

### H. Transformation and comparison

| Function | Description |
|---|---|
| `xml.toJson(node, [options])` | Conventional JSON by default; `{"lossless": true}` returns a typed tree with document/element/text/cdata/comment/pi nodes |
| `xml.fromJson(value, [rootName])` | Build an XML document from the same convention |
| `xml.toCsv(node, rowName, [columns])` | CSV from repeated elements; columns may be `@attribute` or child names |
| `xml.xslt(node, stylesheet)` | Apply an XSLT stylesheet; browser `XSLTProcessor` required |
| `xml.merge(target, source, [options])` | Default appends cloned source children. `{"mode": "update", "key": "id"}` updates matching children and appends new ones; returns the changed node count |
| `xml.merge3(base, ours, theirs, [options])` | Three-way merge into `ours`; applies non-conflicting changes from `theirs` and returns an array of conflict descriptions |
| `xml.diff(a, b, [options])` | Structural diff with applicable changes. Options: `key` (ID attribute), `detectMoves` (default true), `includeXml` (default true), `maxXml` |
| `xml.patch(target, changes)` | Apply a diff produced by `xml.diff`; resolves every node before mutating, so removals/moves do not invalidate paths |

`xml.diff` changes carry `op`, `path`, and enough data to apply them: `xml` for
added/replaced/removed nodes, `name`/`old`/`new` for attributes and text,
`parent`/`index` for structural changes, and `from`/`to`/`fromIndex` for moves.
With `{"key": "id"}` children are aligned by that attribute and reordering
produces `moved` operations; without a key the diff stays positional. Paths
use key predicates such as `/root/item[@id="2"]`, so the same diff can be
applied to a clone. `xml.patch` supports `attribute`, `text`, `added`,
`removed`, `replaced` and `moved` operations and reports unresolved paths
instead of silently skipping them. `xml.merge3` resolves non-conflicting
attribute/text/child changes and leaves conflicting nodes as they are in
`ours`; compare the returned conflict strings before saving.

XML → HTML and arbitrary XML → XML transformations are done with XSLT. JSON
conversion is deliberately conventional and may lose mixed-content order;
use the DOM functions when order matters. Moves are detected within one parent;
cross-parent moves are reported as a removal plus an addition.

Security: XML data is never inserted into HTML; the serializer only emits XML
text. XPath expressions, XSD schemas and XSLT stylesheets are developer code.
DTD/ENTITY declarations are rejected, remote loads require a literal allow
list, and XML parsing is bounded by the limits above.

## Dialog Module (`dialog`, desktop)

Native OS file and folder selection. Import it with `import { dialog } from "std"`; on web targets the calls return an error. The dialog is user-driven and the paths can be passed to `exec.run`/`fs` (which still enforce their capability allow-lists).

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `dialog.pickFile()` | - | (string, string) | true | Native file dialog; empty string on cancel |
| `dialog.pickFiles()` | - | (array[string], string) | true | Native multi-file dialog |
| `dialog.pickDirectory()` | - | (string, string) | true | Native folder dialog; empty string on cancel |

Desktop implementation:

- Windows: PowerShell + `System.Windows.Forms` dialogs;
- macOS: `osascript` + `choose file` / `choose folder`;
- Linux: `zenity`, falling back to `kdialog`.
## Table Module (`table`, core)

`table` is a **core** module: always available and no import required. It
targets the table with the matching `name:` attribute. Functions are synchronous
(DOM-only) and return `(value, error)`.

| Function | Params | Return | Description |
|---|---|---|---|
| `table.display(name, data)` | string, any | (int, string) | Fill from JSON, CSV/TSV/SSV/PSV, JSONL, arrays or objects |
| `table.data(name)` | string | (array[array[string]], string) | Visible rows, including edits |
| `table.header(name)` | string | (array[string], string) | Header labels |
| `table.select(name, cols)` | string, array[int] | (array[array[string]], string) | Selected columns only |
| `table.slice(name, from, count)` | string, int, int | (array[array[string]], string) | Row block (`count <= 0` = to the end) |
| `table.sort(name, column, dir)` | string, string, string | (int, string) | Sort by column key/index; `asc`, `desc` or empty |
| `table.filter(name, column, value)` | string, string, string | (int, string) | Filter one column by text |
| `table.filterText(name, text)` | string, string | (int, string) | Global text filter |
| `table.page(name, index)` | string, int | (int, string) | Zero-based page |
| `table.setPageSize(name, size)` | string, int | (int, string) | Rows per page (`0` = all) |
| `table.selectedRows(name)` | string | (array[int], string) | Selected row indexes |
| `table.selectedCells(name)` | string | (string, string) | JSON `[{row,column}]` |
| `table.selectedColumns(name)` | string | (string, string) | JSON array of column indexes |
| `table.lastChange(name)` | string | (string, string) | JSON `{op,row,column,old,new,...}` of the last change |
| `table.importText(name, text, spec)` | string, string, string | (int, string) | Explicit parse: `csv`, `tsv`, `ssv`, `psv`, `json`, `jsonl`, `auto` |
| `table.exportText(name, spec)` | string, string | (string, string) | Serialize to CSV/TSV/SSV/PSV/JSON/JSONL |
| `table.json(name)` | string | (string, string) | JSON (objects when a header exists) |
| `table.clear(name)` | string | (int, string) | Clear the body |
| `table.clearPersisted(name)` | string | (int, string) | Discard saved persistent data/layout and restore the design-time table |

`importText`/`exportText` spec: format plus options separated by `;`:

- `header=0|1`
- `sep=comma|semicolon|tab|pipe|<char>`
- `quote=<char>`
- `trim=1`
- `formula=escape` (CSV export: neutralizes `= + - @` with `'`)

```sui
state out = ""

fn load() -> null {
	n, e = table.display("users", "ID,Name\n1,Alice\n2,Bob")
	if e != null { out = e } else { out = str(n) + " rows" }
}

fn saveCSV() -> null {
	csv, e = table.exportText("users", "csv;formula=escape")
	out = csv
}
```

Security:

- cells are written with `textContent`: HTML in data is never interpreted;
- limits: 100,000 rows, 10,000 columns, 20 MB of text per call;
- `formula=escape` prevents formula injection when CSV is opened in
  Excel/LibreOffice (opt-in).

Full details: [`SECURITY.md`](../SECURITY.md).

## Chart Module (`chart`, core)

`chart` is a **core** module: always available and no import required. It
targets the rendered state chart with the matching `name:` attribute (or its
title). Functions are synchronous and return `(value, error)`. They update the
tag text and colors of shapes and links without rebuilding the whole chart, which
makes the chart usable as a process/event monitor. Static interactions are also
available as an event queue (`chart.pollEvents`, `chart.nextEvent`, ...) so the
chart never needs to call SUI state-changing functions directly.

| Function | Params | Return | Description |
|---|---|---|---|
| `chart.setTag(name, id, text, color, textColor)` | string, string, string, string, string | (bool, string) | Set tag text, tag background color and tag text color |
| `chart.update(name, id, patch)` | string, string, any | (bool, string) | Apply a patch object or JSON object to one element |
| `chart.updateMany(name, updates)` | string, any | (int, string) | Apply a map `{ id: patch }` or array of patches |
| `chart.get(name, id)` | string, string | (string, string) | Current attributes of one shape/link as JSON |
| `chart.elements(name)` | string | (string, string) | Current shape/link summaries as a JSON array |
| `chart.find(name, query)` | string, string | (string, string) | Search elements by `field:value` or any field |
| `chart.getAttr(name, id, field)` | string, string, string | (string, string) | Read one attribute as a string |
| `chart.setAttr(name, id, field, value)` | string, string, string, any | (bool, string) | Write one validated attribute |
| `chart.focusRow(name, id)` | string, string | (bool, string) | Highlight, scroll to and focus a table row (works in non-editable mode too) |
| `chart.pollEvents(name)` | string | (string, string) | All pending events as JSON and empties the queue |
| `chart.peekEvents(name)` | string | (string, string) | Pending events as JSON without emptying |
| `chart.nextEvent(name)` | string | (string, string) | Oldest pending event as JSON and removes it |
| `chart.clearEvents(name)` | string | (bool, string) | Empty the event queue |
| `chart.clearPersisted(name)` | string | (int, string) | Discard saved persistent chart/graph state and restore the design-time source |
| `chart.eventCount(name)` | string | (int, string) | Number of pending events |

Element ids:

- shape `id` (`"a"`), mandatory and unique inside the chart;
- link `id` (`"edge_1"`); links may omit it and receive `edge_1`, `edge_2`, ... with a uniqueness check.

All runtime functions accept only the explicit `id`. Prefixed keys (`shape:`,
`state:`, `link:`, `edge:`), labels and `from->to` are rejected.

Events report the explicit `id`. For single-attribute changes use the typed
helpers (`chart.setText`, `chart.setTagColor`, `chart.setHint`,
`chart.setColor`, ...); `chart.getAttr`/`chart.setAttr` cover the remaining fields and `chart.update`
handles multi-attribute patches. Additional typed accessors: `getWidth`/`setWidth`
(shape `w` or link line), `getHeight`/`setHeight` (shape), `getX`/`setX`,
`getY`/`setY`, `getRotation`/`setRotation`, `getDashStyle`/`setDashStyle`,
`getStartMarker`/`setStartMarker`, `getEndMarker`/`setEndMarker`,
`getFrom`/`setFrom`, `getTo`/`setTo`, `getWaypoints`/`setWaypoints`,
`getTagKind`/`setTagKind`, `getTagBox`/`setTagBox`, `getTagAlign`/`setTagAlign`,
`getTagLeader`/`setTagLeader`, `getTextPlacement`/`setTextPlacement`,
`getTagSide`/`setTagSide`, `getBorderDashLength`/`setBorderDashLength`,
`getBorderDashGap`/`setBorderDashGap`, `getSnap`/`setSnap`.

Read the current runtime values (after edits/imports) with:

```sui
shapeJson, err = chart.get("monitor", "process")
allJson, err2 = chart.elements("monitor")
```

`chart.get` returns `id`, `label`, `tag`, `hint`, geometry, colors and handler
fields. `chart.elements` returns a JSON array with `kind`, `id`, `label`, `tag`,
`hint`, geometry/routing, colors and `onClick`/`onTagClick` names.

The chart is an event source. Every static interaction is queued; SUI drains the
queue and calls its own external functions. This is the recommended model:

```sui
events, err = chart.pollEvents("monitor")
```

The queue payload includes `event`, `kind`, `part`, `id`, `label`, `tag`,
`from`, `to`, pointer coordinates and modifiers. Use `chart.pollEvents` (drain),
`chart.peekEvents` (read without draining), `chart.nextEvent` (one event),
`chart.clearEvents` and `chart.eventCount`.

Charts are event-source only. Interactions are consumed from the queue with
`chart.pollEvents` / `chart.peekEvents` / `chart.nextEvent`; the SUI code decides
which application function to call. Direct function callbacks (`chart.on`,
element handler fields and chart-level `onShape*`/`onLink*`/`onTag*`) are not
available.


Patch fields for tags: `tag`/`text`, `tagTextColor`/`textColor`,
`tagBoxColor`/`tagColor`/`backgroundColor`, `tagBorderColor`,
`tagLeaderColor`, `tagKind`, `tagAlign`, `tagBox`, `tagLeader`, `tagArrow`,
`tagArrowSize`, `tagArrowShape`, `tagFont`, `tagFontSize`, `tagFontWeight`,
`tagFontStyle`, `tagTextDecoration`, `tagTextAlign`, `tagLetterSpacing`,
`tagLineSpacing`, `tagTextTransform`, `tagTextShadow`,
`tagTextShadowColor`, `tagPadding`, `tagRotation`, `tagOffset`.
Element fields accepted by `update`: `label`, `description`, `hint`, `color`,
`borderColor`, `highlightColor`, `borderStyle`, `borderWidth`, `cornerRadius`,
`shape`, `textPlacement`, `tagSide`, `snap`, `borderDashLength`,
`borderDashGap`, shape geometry `x`/`y`/`w`/`h`/`rotation`, and for links
`routing`, `width`, `dashStyle`, `dashLength`, `dashGap`, `arrowSize`,
`startMarker`, `endMarker`, `from`, `to`, `fromPort`, `toPort`, `fromAnchor`,
`toAnchor`, `waypoints`, `arcRadius`, `arcSweep`, `arcLarge` and `bend*`
controls. In `setTag`/`tagColor`, `color` means the tag background; in `update`,
`color` means the shape fill or link line color.

Colors accept `#rgb`, `#rrggbb`, `rgb()/rgba()` and `transparent`/`none`.

```sui
fn updateMonitor() -> null {
	ok, e = chart.setTag("monitor", "a", "Running", "#22c55e", "#ffffff")
	ok2, e2 = chart.update("monitor", "edge_1", "\{\"text\":\"go\",\"tagColor\":\"#0ea5e9\",\"textColor\":\"#ffffff\"\}")
	n, e3 = chart.updateMany("monitor", "\{\"b\":\{\"tag\":\"Waiting\"\},\"c\":\{\"tag\":\"Done\"\}\}")
}

interval(1000, updateMonitor)
```

## Tree Module (`tree`, core)

`tree` is a **core** module: always available and no import required. It targets
the tree with the matching `name:` attribute (`name: "files"`). Functions return
`(value, error)`.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `tree.importJson(name, text)` | string, string | (int, string) | no | Replace nodes from JSON; returns node count |
| `tree.importXml(name, text)` | string, string | (int, string) | no | Replace nodes from XML; returns node count |
| `tree.importJsonFile(name, path)` | string, string | (int, string) | yes | Read a selected browser file and import JSON |
| `tree.importXmlFile(name, path)` | string, string | (int, string) | yes | Read a selected browser file and import XML |
| `tree.exportJson(name)` | string | (string, string) | no | Export the current nodes as JSON text |
| `tree.exportXml(name)` | string | (string, string) | no | Export the current nodes as XML text |
| `tree.downloadJson(name, filename)` | string, string | (int, string) | no | Download current nodes as a JSON file (`1` = success) |
| `tree.downloadXml(name, filename)` | string, string | (int, string) | no | Download current nodes as an XML file (`1` = success) |
| `tree.clear(name)` | string | (int, string) | no | Remove all nodes |
| `tree.expandAll(name)` | string | (int, string) | no | Expand every parent; returns the number of parents |
| `tree.collapseAll(name)` | string | (int, string) | no | Collapse every parent |
| `tree.expandTo(name, level)` | string, int | (int, string) | no | Expand parents above `level`; collapse deeper ones |
| `tree.setChildren(name, key, json)` | string, string, string | (int, string) | no | Add/replace children of an item from JSON; returns node count |
| `tree.setChildrenError(name, key, message)` | string, string, string | (int, string) | no | Show a lazy-load error on an item |
| `tree.retryChildren(name, key)` | string, string | (int, string) | no | Retry a lazy load through `onLoad` |
| `tree.lastMove(name)` | string | (string, string) | no | JSON `{key, parent, index}` of the last reorder |
| `tree.lastChange(name)` | string | (string, string) | no | JSON `{op, key, parent, index, ...}` of the last data change |
| `tree.undoMove(name)` | string | (int, string) | no | Restore the previous tree from the in-memory undo stack |
| `tree.selectedKeys(name)` | string | (string, string) | no | JSON array of selected keys |
| `tree.checkedKeys(name)` | string | (string, string) | no | JSON array of checked keys |
| `tree.search(name, term)` | string, string | (int, string) | no | Apply or clear the inline filter |
| `tree.setColumns(name, json)` | string, string | (int, string) | no | Create/replace columns from a JSON array |
| `tree.getColumns(name)` | string | (string, string) | no | Current columns as a JSON array |
| `tree.moveColumn(name, from, to)` | string, int, int | (int, string) | no | Reorder a column |
| `tree.setCell(name, key, column, value)` | string, string, string, string | (int, string) | no | Set one cell value |

### JSON format

The root can be an array, or an object with `nodes`, `items`, `tree` or
`children`. Each node supports:

```json
[
  {
    "label": "src",
    "key": "src",
    "open": true,
    "editable": true,
    "renameIcon": "A",
    "showRenameIcon": false,
    "icon": "📁",
    "iconOpen": "📂",
    "badge": "12",
    "badgeType": "info",
    "meta": "2 KB",
    "checked": true,
    "tooltip": "Source directory",
    "lazy": false,
    "size": "2 KB",
    "date": "today",
    "status": "ok",
    "cells": { "owner": "Anna" },
    "children": [
      { "label": "main.go", "key": "main" }
    ]
  }
]
```

- `label` is preferred; `text`, `name` and `title` are accepted aliases.
- `key` is optional; a generated path is used when it is missing. Keys must be
  unique.
- `open`, `editable`, `showRenameIcon` and `checked` accept booleans or
  `true`/`false`/`1`/`0` strings.
- `icon`, `iconOpen`, `badge`, `badgeType`, `meta`, `tooltip`, `size`, `date`
  and `status` are inert text. `badgeType` is `neutral`, `info`, `success`,
  `warning` or `danger`.
- `lazy: true` marks an item whose children are loaded later with
  `tree.setChildren`.
- `cells` is an object of custom column values (`{ "owner": "Anna" }`); the
  SUI attribute also accepts `key=value` text.
- Strings and numbers are accepted as leaf labels.

### XML format

```xml
<?xml version="1.0" encoding="UTF-8"?>
<tree name="files">
  <item label="src" key="src" open="true" icon="📁" iconOpen="📂" badge="12" badgeType="info" meta="2 KB" checked="true">
    <item label="main.go" key="main"/>
    <item label="README.md" key="readme" renameIcon="📝"/>
  </item>
  <item label="generated" key="generated" editable="false"/>
</tree>
```

`<tree>` is the container; any element is accepted as a node, so `<item>`,
`<node>` or application-specific tags work. The element text is used as the
label when no `label`/`text`/`name`/`title` attribute is present. XML attributes
`icon`, `iconOpen`, `badge`, `badgeType`, `meta`, `tooltip`, `lazy`, `size`,
`date`, `status` and `checked` are preserved.

### Lazy loading

An item marked `lazy: true` has no children and shows a disclosure arrow.
Expanding it writes its key to the state bound by `loadKey` and calls `onLoad`.
The handler can append children later:

```sui
state loadKey = ""

fn loaded() -> null {
	// start an async request using loadKey
}

async fn finishLoad() -> null {
	n, e = tree.setChildren("files", loadKey, "[{\"label\":\"child\"}]")
	if e != null {
		tree.setChildrenError("files", loadKey, e)
	}
}
```

Use `tree.selectedKeys`/`tree.checkedKeys` to read selection and checkbox state;
both return JSON arrays.

### File helpers

`importJsonFile`/`importXmlFile` read the file selected by `pickFiles:` or a drop
(`drag.files()`), matching `path` or `name`; pass `""` to use the first selected
file. On desktop, a native `dialog.pickFile()` returns only a path: read it with
`fs.read(path)` and pass the text to `tree.importJson`/`tree.importXml` instead.

```sui
state importPath = ""
state note = ""

async fn importJSON() -> null {
	n, e = tree.importJsonFile("files", importPath)
	if e != null { note = e } else { note = "imported " + str(n) + " nodes" }
}
```

`downloadJson`/`downloadXml` create a browser download; the filename is
sanitized and the extension is added when missing.

### Limits and safety

- limits: 10,000 nodes, nesting depth 64, 2,000 characters per label and
  tooltip, 512 per key, 64 per icon and 128/256 per badge/metadata; input text
  uses the global `maxText` limit (default 20 MB);
- labels, keys, icons, badges and metadata are written with
  `textContent`/attributes, never HTML;
- XML DTD and entity declarations are rejected before parsing;
- imported data replaces only the node structure: the tree title, style,
  `name` and runtime bindings are preserved;
- local open/rename overrides for that tree are cleared on import;
- imported models survive reactive view re-renders;
- export reads the visible labels and open state, so inline renames are
  reflected.

## Error Handling Pattern

Every two-value function follows this pattern:

```sui
async fn fetchData() {
  body, err = http.get("https://api.example.com/data")
  if err != null {
    // handle error
    return
  }
  data, perr = json.parse(body)
  if perr != null {
    // handle parse error
    return
  }
  // use data
}
```

## Security Notes (Desktop)

- **Capability-based**: explicit permissions via CLI (`--allow-exec`,
  `--allow-dir`)
- **Path traversal protection**: normalization, allow-list and symlink
  resolution (a link inside an allowed directory cannot escape)
- **Per-run secret**: bridge calls use a random token compared in constant time
- **Localhost only**: binds `127.0.0.1`, requires a loopback Host header
  (anti-DNS-rebinding), rejects cross-site requests
- **Command resolution**: commands without a path separator are resolved next
  to the executable; no shell, so arguments are not reinterpreted
- **app.js**: `Cache-Control: no-store`, `X-Frame-Options: DENY`

## Not Yet Available

- filesystem operations not listed above (recursive copy, watch, glob);
- source maps.
