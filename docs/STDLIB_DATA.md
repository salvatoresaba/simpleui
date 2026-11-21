# SimpleUI Data Stdlib Documentation

This document covers the pure and in-memory std modules. They never access the
network, the filesystem, the process environment or the DOM. They are available
on both web and desktop targets and follow the normal std convention:
`value, err = module.function(...)` with `err = null` on success. All functions
in these modules are synchronous.

The modules are `bytes`, `bin`, `files`, `mime`, `color`, `units`, `print`,
`history`, `zip`, `archive`, `csv`, `env`, `capabilities`, `events`, `bus`, `log`
and `telemetry`. For I/O modules (`http`, `ws`, `fs`, `exec`, `dialog`) see
[STDLIB](STDLIB.md).

## Import

```sui
import { bytes, csv, zip, color } from "std"
```

Only imported modules are emitted in the generated runtime. Importing a single
module from this document does not pull in the others.

## Bytes Module (`bytes`)

Byte arrays are `array[int]` values containing 0-255. Inputs are validated; the
maximum buffer is 10,000,000 bytes.

| Function | Params | Return | Description |
|---|---|---|---|
| `bytes.fromString(text, [encoding])` | string, string? | (array[int], string) | Encode text as `utf-8` (default), `utf-16le` or `utf-16be` |
| `bytes.toString(bytes, [encoding])` | any, string? | (string, string) | Decode bytes; invalid UTF-8 returns an error |
| `bytes.fromBase64(text)` | string | (array[int], string) | Decode base64 |
| `bytes.toBase64(bytes)` | any | (string, string) | Encode base64 in chunks (no call-stack limit) |
| `bytes.fromHex(text)` | string | (array[int], string) | Decode hexadecimal, ignoring whitespace |
| `bytes.toHex(bytes)` | any | (string, string) | Encode lowercase hexadecimal |
| `bytes.concat(a, b)` | any, any | (array[int], string) | Concatenate two buffers |
| `bytes.slice(bytes, start, [end])` | any, int, int? | (array[int], string) | Copy a byte range |
| `bytes.indexOf(bytes, needle, [start])` | any, any, int? | (int, string) | First index of a byte sequence, or `-1` |
| `bytes.equal(a, b)` | any, any | (bool, string) | Byte-wise equality |
| `bytes.length(bytes)` | any | (int, string) | Buffer length |
| `bytes.fill(length, byte)` | int, int | (array[int], string) | Create a filled buffer |
| `bytes.reverse(bytes)` | any | (array[int], string) | Copy and reverse a buffer |
| `bytes.xor(a, b)` | any, any | (array[int], string) | Byte-wise XOR; lengths must match |

## Binary Module (`bin`)

`bin` reads and writes integers and IEEE-754 floats with `DataView`. The
`little` argument defaults to `true` for multi-byte values. Offsets are checked.

| Function | Params | Return | Description |
|---|---|---|---|
| `bin.u8(bytes, offset)` | any, int | (int, string) | Read an unsigned byte |
| `bin.u16(bytes, offset, [little])` | any, int, bool? | (int, string) | Read `uint16` |
| `bin.u32(bytes, offset, [little])` | any, int, bool? | (int, string) | Read `uint32` |
| `bin.i8(bytes, offset)` | any, int | (int, string) | Read `int8` |
| `bin.i16(bytes, offset, [little])` | any, int, bool? | (int, string) | Read `int16` |
| `bin.i32(bytes, offset, [little])` | any, int, bool? | (int, string) | Read `int32` |
| `bin.f32(bytes, offset, [little])` | any, int, bool? | (float, string) | Read `float32` |
| `bin.f64(bytes, offset, [little])` | any, int, bool? | (float, string) | Read `float64` |
| `bin.writeU8(value)` | int | (array[int], string) | Write an unsigned byte |
| `bin.writeU16(value, [little])` | int, bool? | (array[int], string) | Write `uint16` |
| `bin.writeU32(value, [little])` | int, bool? | (array[int], string) | Write `uint32` |
| `bin.writeI8(value)` | int | (array[int], string) | Write `int8` |
| `bin.writeI16(value, [little])` | int, bool? | (array[int], string) | Write `int16` |
| `bin.writeI32(value, [little])` | int, bool? | (array[int], string) | Write `int32` |
| `bin.writeF32(value, [little])` | float, bool? | (array[int], string) | Write `float32` |
| `bin.writeF64(value, [little])` | float, bool? | (array[int], string) | Write `float64` |

## Files Module (`files`)

These helpers validate names and do **not** read or write files. Use the
desktop `fs` module for actual file I/O. File names are treated as untrusted.

| Function | Params | Return | Description |
|---|---|---|---|
| `files.name(path)` | string | (string, string) | Base name; both `/` and `\` are accepted |
| `files.stem(path)` | string | (string, string) | Base name without the last extension |
| `files.ext(path)` | string | (string, string) | Lowercase extension including `.` |
| `files.split(path)` | string | (map[string]string, string) | Map with `name`, `stem` and `ext` |
| `files.sanitize(name, [maxLen])` | string, int? | (string, string) | Remove separators, controls and leading dots; default limit 120 |
| `files.validate(name, [options])` | string, any? | (bool, string) | Validate name, hidden flag, extension and size |
| `files.isHidden(name)` | string | (bool, string) | Whether the base name starts with `.` |
| `files.extAllowed(name, allow)` | string, string | (bool, string) | Check an extension allow-list |

`files.validate` options are `maxName` (default 200), `allowHidden` (default
false), `allowedExt` or `allow` (comma/space separated list or array), `size`
and `maxSize`. Traversal, path separators, drive letters, control characters and
hidden names are rejected by default.

## MIME Module (`mime`)

| Function | Params | Return | Description |
|---|---|---|---|
| `mime.fromName(name)` | string | (string, string) | MIME type from the extension; falls back to `application/octet-stream` |
| `mime.fromBytes(bytes)` | any | (string, string) | Detect common image, PDF, ZIP, XML and JSON signatures |
| `mime.extension(mimeType)` | string | (string, string) | Suggested extension including `.`, or an empty string |
| `mime.isAllowed(mimeType, allow)` | string, string | (bool, string) | Exact, wildcard (`image/*`) or `*/*` allow-list |
| `mime.validate(name, mimeType, size, [options])` | string, string, int, any? | (bool, string) | Combine file-name, size, extension and MIME checks |

`mime.validate` reuses the `files.validate` options and adds `allow` for MIME
types. Both helpers return `(false, error)` on rejection.

## Color Module (`color`)

Colors are maps with numeric `r`, `g`, `b` (0-255) and `a` (0-1). Parsing
supports `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa`, `rgb()`, `rgba()`, `hsl()`,
`hsla()` and a small set of named colors. Unsupported values fail with an
error; no CSS is injected or evaluated.

| Function | Params | Return | Description |
|---|---|---|---|
| `color.parse(value)` | string | (any, string) | Parse to a color map |
| `color.format(color, [format])` | any, string? | (string, string) | Format as `hex` (default), `rgb`, `rgba`, `hsl` or `hsla` |
| `color.rgb(r, g, b, [a])` | int, int, int, float? | (any, string) | Build a color map |
| `color.hsl(h, s, l, [a])` | float, float, float, float? | (any, string) | Convert HSL to a color map |
| `color.mix(a, b, t)` | any, any, float | (any, string) | Interpolate with `t` clamped to 0-1 |
| `color.withAlpha(color, alpha)` | any, float | (any, string) | Replace the alpha channel |
| `color.luminance(color)` | any | (float, string) | WCAG relative luminance |
| `color.contrast(a, b)` | any, any | (float, string) | WCAG contrast ratio |
| `color.isDark(color)` | any | (bool, string) | Whether luminance is below 0.5 |

## Units Module (`units`)

Absolute CSS units are supported: `px`, `mm`, `cm`, `in`, `pt`, `pc` and `q`.
The pixel base is 96 dpi. Relative units such as `em`, `rem` and `%` are
rejected by conversion helpers because their base depends on the document.

| Function | Params | Return | Description |
|---|---|---|---|
| `units.parse(value)` | string | (any, string) | Split a length into `value` and `unit` |
| `units.convert(value, from, to)` | float, string, string | (float, string) | Convert between absolute units |
| `units.toBase(value, unit)` | float, string | (float, string) | Convert to CSS pixels |
| `units.fromBase(value, unit)` | float, string | (float, string) | Convert from CSS pixels |
| `units.format(value, unit, [precision])` | float, string, int? | (string, string) | Fixed-precision text with trailing zeros removed |

## Print Module (`print`)

| Function | Params | Return | Description |
|---|---|---|---|
| `print.pxToMm(px)` | float | (float, string) | Pixels to millimetres at 96 dpi |
| `print.mmToPx(mm)` | float | (float, string) | Millimetres to pixels at 96 dpi |
| `print.ptToMm(pt)` | float | (float, string) | Points to millimetres |
| `print.paper(name, [orientation])` | string, string? | (map[string]any, string) | `a3`, `a4`, `a5`, `letter`, `legal`, `tabloid`; `portrait` default |
| `print.fitBox(width, height, maxW, maxH)` | float, float, float, float | (map[string]any, string) | Scale a box down while preserving its aspect ratio |

`print.paper` returns `width`, `height`, `unit` (`"mm"`) and `orientation`.
`print.fitBox` returns `width`, `height` and `scale`.

## History Module (`history`)

Histories are named, in-memory stacks. Values are deep-copied on push and on
read, so later mutations of the caller's data cannot change a saved state.
History names allow letters, digits, `_`, `.`, `:` and `-` (max 128).

| Function | Params | Return | Description |
|---|---|---|---|
| `history.push(name, value, [limit])` | string, any, int? | (int, string) | Save a snapshot; default limit 50, max 1000 |
| `history.undo(name)` | string | (any, string) | Move back one step and return the snapshot |
| `history.redo(name)` | string | (any, string) | Move forward one step and return the snapshot |
| `history.current(name)` | string | (any, string) | Return the current snapshot without moving |
| `history.canUndo(name)` | string | (bool, string) | Whether an undo step exists |
| `history.canRedo(name)` | string | (bool, string) | Whether a redo step exists |
| `history.clear(name)` | string | (null, string) | Remove all snapshots |
| `history.list()` | none | (array[string], string) | Sorted history names |
| `history.size(name)` | string | (int, string) | Number of snapshots |
| `history.jump(name, index)` | string, int | (any, string) | Move to a snapshot by index |

Cyclic values and values nested deeper than 64 levels are rejected.

## ZIP Module (`zip`)

ZIP archives are created in memory and use the **stored** method only. This
keeps the runtime small and removes decompression bombs. Up to 1,000 entries,
10 MB per entry and 50 MB per archive are accepted.

| Function | Params | Return | Description |
|---|---|---|---|
| `zip.create()` | none | (any, string) | Create a new writer handle |
| `zip.add(writer, name, bytes)` | any, string, any | (int, string) | Add a binary entry; returns the entry count |
| `zip.addText(writer, name, text)` | any, string, string | (int, string) | Add a UTF-8 text entry |
| `zip.finish(writer)` | any | (array[int], string) | Emit the archive bytes and close the writer |
| `zip.open(bytes)` | any | (any, string) | Parse and validate an archive |
| `zip.list(archive)` | any | (array[any], string) | Entry maps: `name`, `size`, `crc`, `method` |
| `zip.read(archive, name)` | any, string | (array[int], string) | Read entry bytes after CRC verification |
| `zip.readText(archive, name)` | any, string | (string, string) | Decode an entry as UTF-8 |
| `zip.has(archive, name)` | any, string | (bool, string) | Whether an entry exists |
| `zip.count(archive)` | any | (int, string) | Number of entries |

Entry names must be relative UTF-8 paths. Absolute paths, Windows drive
letters, backslashes, `.`/`..` segments, directory entries, NUL bytes,
encrypted entries, compressed methods and symbolic links are rejected. CRCs are
verified when reading.

## Archive Module (`archive`)

`archive` is a map-oriented wrapper over the same ZIP writer and reader.

| Function | Params | Return | Description |
|---|---|---|---|
| `archive.pack(files, [format])` | any, string? | (array[int], string) | Map of name to byte arrays into ZIP bytes |
| `archive.unpack(bytes, [format])` | any, string? | (map[string]any, string) | ZIP bytes into name to byte arrays |
| `archive.packText(files, [format])` | any, string? | (array[int], string) | Map of name to text into ZIP bytes |
| `archive.unpackText(bytes, [format])` | any, string? | (map[string]string, string) | ZIP bytes into name to text |
| `archive.info(bytes)` | any | (array[any], string) | Parse and list entry metadata |

Only `format = "zip"` is supported; any other value returns an error. Object
keys named `__proto__` are rejected when building maps.

## CSV Module (`csv`)

CSV parsing follows RFC 4180 style quoting: doubled quotes escape a quote, and
quoted fields may contain delimiters or newlines. Up to 10,000 rows and 1,000
columns are accepted. Options are `delimiter` (one character), `quote`,
`newline`, `columns` and `header` (default true for exports).

| Function | Params | Return | Description |
|---|---|---|---|
| `csv.parse(text, [options])` | string, any? | (array[any], string) | Rows of string arrays; a leading BOM is ignored |
| `csv.parseObjects(text, [options])` | string, any? | (array[any], string) | First row becomes object keys; duplicates are rejected |
| `csv.stringify(rows, [options])` | any, any? | (string, string) | Serialize arrays of rows |
| `csv.stringifyObjects(rows, [options])` | any, any? | (string, string) | Serialize maps using `columns` or the first row's keys |
| `csv.validate(text, [options])` | string, any? | (bool, string) | Parse and require a consistent column count |

Exports always escape spreadsheet formulas: a cell starting with `=`, `+`,
`-`, `@`, tab or carriage return is prefixed with a single quote. This is not
configurable, because CSV is often opened by spreadsheet applications. Data
imported into a table must still go through `table.display`.

## Environment Module (`env`)

`env` exposes only the following read-only facts; it never reads process
variables and never writes to them:

| Key | Meaning |
|---|---|
| `SUI_TARGET` | Compile target, for example `web` or `desktop` |
| `SUI_PLATFORM` | Browser-reported platform, or `unknown` |
| `SUI_LANGUAGE` | Browser language, or `en` |
| `SUI_TIMEZONE` | IANA time zone, or `UTC` |
| `SUI_USER_AGENT` | Browser user-agent string, or empty |
| `SUI_DESKTOP` | `"true"` for desktop targets |

| Function | Params | Return | Description |
|---|---|---|---|
| `env.target()` | none | (string, string) | Same as `SUI_TARGET` |
| `env.platform()` | none | (string, string) | Same as `SUI_PLATFORM` |
| `env.get(name)` | string | (string, string) | Read one allow-listed key; unknown keys return an error |
| `env.has(name)` | string | (bool, string) | Whether the key is exposed |
| `env.all()` | none | (map[string]string, string) | Copy of all exposed keys |
| `env.language()` | none | (string, string) | Same as `SUI_LANGUAGE` |
| `env.timezone()` | none | (string, string) | Same as `SUI_TIMEZONE` |

Names are case-insensitive for `get` and `has`. Values are computed from the
current runtime and browser; they are informational and must not be used for
authorization decisions.

## Capabilities Module (`capabilities`)

| Function | Params | Return | Description |
|---|---|---|---|
| `capabilities.has(name)` | string | (bool, string) | Whether the capability is known in this build |
| `capabilities.check(name)` | string | (bool, string) | Like `has`, but returns an error when unavailable |
| `capabilities.list()` | none | (array[string], string) | Sorted capability names |
| `capabilities.target()` | none | (string, string) | Same as `env.target()` |

Known names: `storage`, `clipboard`, `crypto`, `fs`, `exec`, `dialog`,
`notifications`, `intl`, `typedArrays` and `textEncoder`. `fs`, `exec` and
`dialog` are true only for desktop targets; they still require the build-time
`--allow-dir` / `--allow-exec` grants to work.

## Events Module (`events`)

The bus is local to the running app. Listeners are JavaScript callback values
passed by name, for example `events.on("data", onData)`. Payloads are deep-copied
before delivery where possible. Errors thrown by a listener are returned by
`emit` after all listeners have run.

| Function | Params | Return | Description |
|---|---|---|---|
| `events.on(name, fn)` | string, any | (int, string) | Register a listener; returns the listener count |
| `events.once(name, fn)` | string, any | (int, string) | Register a listener removed after its first call |
| `events.off(name, fn)` | string, any | (int, string) | Remove matching listeners; returns the number removed |
| `events.emit(name, payload)` | string, any | (int, string) | Notify listeners; returns the number called |
| `events.clear(name)` | string | (null, string) | Remove all listeners for the event |
| `events.count(name)` | string | (int, string) | Listener count |
| `events.names()` | none | (array[string], string) | Sorted event names with listeners |

## Event Bus Module (`bus`)

`bus` is an alias of `events` with the same signatures and behavior. It exists
so applications can choose the terminology that fits their design:

`bus.on`, `bus.once`, `bus.off`, `bus.emit`, `bus.clear`, `bus.count` and
`bus.names`.

## Log Module (`log`)

Logs stay in memory: a 500-entry ring buffer, 2,000-character messages and no
network transport. Secret-looking keys and common token patterns are redacted
before an entry is stored. The level threshold is `debug` < `info` < `warn` <
`error`; the default is `info`.

| Function | Params | Return | Description |
|---|---|---|---|
| `log.debug(message, [context])` | string, any? | (null, string) | Add a debug entry |
| `log.info(message, [context])` | string, any? | (null, string) | Add an info entry |
| `log.warn(message, [context])` | string, any? | (null, string) | Add a warning entry |
| `log.error(message, [context])` | string, any? | (null, string) | Add an error entry |
| `log.setLevel(level)` | string | (null, string) | Set the threshold |
| `log.level()` | none | (string, string) | Current threshold |
| `log.clear()` | none | (null, string) | Clear stored entries |
| `log.entries([limit])` | int? | (array[any], string) | Last entries, up to 500; default 100 |
| `log.export()` | none | (string, string) | JSON text of the stored entries |
| `log.redact(text)` | string | (string, string) | Redact secret patterns in text |
| `log.redactObject(value)` | any | (any, string) | Deep copy with secret keys redacted |

Redacted key names include `password`, `token`, `secret`, `apiKey`, `api_key`,
`authorization`, `cookie`, `session`, `creditCard`, `cardNumber`, `cvv` and
`pin`. `Bearer` tokens are also replaced.

## Telemetry Module (`telemetry`)

Telemetry is a local metric registry. It does not export or send data.

| Function | Params | Return | Description |
|---|---|---|---|
| `telemetry.counter(name, [delta])` | string, float? | (float, string) | Increase a counter; default delta 1 |
| `telemetry.gauge(name, value)` | string, float | (float, string) | Set a gauge value |
| `telemetry.timing(name, ms)` | string, float | (float, string) | Record a duration in milliseconds |
| `telemetry.snapshot()` | none | (map[string]any, string) | Copy of all metrics |
| `telemetry.reset()` | none | (null, string) | Remove all metrics |
| `telemetry.names()` | none | (array[string], string) | Sorted metric names |

Metric names allow letters, digits, `_`, `.`, `:` and `-` (max 128). Each
snapshot entry contains `name`, `type`, `value` and `updated`.

## Security notes

- No raw HTML or CSS injection is introduced. These modules return data only.
- `files` and `mime` validate untrusted names but do not grant filesystem
  access; desktop capabilities still apply to `fs`, `exec` and `dialog`.
- ZIP reading accepts only stored entries, rejects unsafe names and symlinks,
  enforces size limits and checks CRCs.
- CSV formula escaping is always enabled on export.
- `env` is an allow-listed read-only view of runtime facts, not a secret store.
- Logs redact common secrets, but applications must still avoid logging
  sensitive data under unusual key names.

## Example

See `examples/stdlib_data.sui` for a runnable app that exercises every module in
this document, including ZIP round-tripping, CSV formula escaping, the event
bus and in-memory logging. Validate and build it with:

```bash
./simpleui.exe check examples/stdlib_data.sui --json
./simpleui.exe build examples/stdlib_data.sui --out dist
```
