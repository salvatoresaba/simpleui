# SimpleUI Builtins Documentation

This page lists the built-in functions available in SimpleUI (no import required).

## String Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `len(s)` | string \| array | int | Length of a string or array |
| `str(v)` | any | string | Convert a value to string |
| `upper(s)` | string | string | Convert to uppercase |
| `lower(s)` | string | string | Convert to lowercase |
| `trim(s)` | string | string | Remove leading/trailing whitespace |
| `cssSafe(s)` | string | string | Sanitize dynamic CSS for `style:` (blocks `url()`, positioning, vendor properties) |
| `trimLeft(s)` | string | string | Remove leading whitespace |
| `trimRight(s)` | string | string | Remove trailing whitespace |
| `contains(s, sub)` | string, string | bool | Whether the string contains the substring |
| `startsWith(s, sub)` | string, string | bool | Whether the string starts with the substring |
| `endsWith(s, sub)` | string, string | bool | Whether the string ends with the substring |
| `indexOf(s, sub)` | string, string (or array, any) | int | Index of the first occurrence (-1 if absent) |
| `lastIndexOf(s, sub)` | string, string | int | Index of the last occurrence (-1 if absent) |
| `slice(s, start, end)` | string, int, int | string | Extract a substring range |
| `substring(s, start, len)` | string, int, int | string | Extract `len` characters starting at `start` |
| `repeat(s, count)` | string, int | string | Repeat a string N times |
| `padLeft(s, len, pad)` | string, int, string | string | Left padding |
| `padRight(s, len, pad)` | string, int, string | string | Right padding |
| `chars(s)` | string | array[string] | Split a string into characters |
| `split(s, sep)` | string, string | array[string] | Split on a separator |
| `splitRegex(s, pattern, [limit])` | string, string, int? | array[string] | Split with a regex (supports an internal `g` flag) |
| `join(arr, sep)` | array[string], string | string | Join an array with a separator |
| `replace(s, old, new)` | string, string, string | string | Replace all occurrences |
| `replaceFirst(s, old, new)` | string, string, string | string | Replace the first occurrence |
| `replaceRegex(s, pattern, repl, [flags])` | string, string, string, string? | string | Regex replacement (`i,m,s,u,y,g`) |
| `match(s, pattern, [flags])` | string, string, string? | array[string] | First match groups |
| `matchAll(s, pattern, [flags])` | string, string, string? | array[any] | All matches |
| `search(s, pattern)` | string, string | int | Index of the first regex match (-1 if absent) |
| `test(s, pattern)` | string, string | bool | Test a regex |
| `extract(s, pattern)` | string, string | string | First capture group (or the full match) |
| `extractAll(s, pattern)` | string, string | array[string] | All capture groups |
| `fmt(n, decimals)` | number, int | string | Format a number with decimals (`toFixed`) |

## Collection Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `append(arr, v)` | array[T], T | array[T] | Append an item (returns a new array) |
| `concat(a, b)` | array, array | array | Concatenate two arrays |
| `remove(arr, idx)` | array, int | array | Remove the item at an index (new array) |
| `includes(arr, v)` | array, any | bool | Whether the array contains the value |
| `sort(arr)` | array | array | Sort a copy |
| `reverse(arr)` | array | array | Reverse a copy |
| `moveItem(arr, from, to)` | array[T], int, int | array[T] | Move one item to another index (new array) |
| `map(arr, fn)` | array[T], fn(T, [int]) | array[U] | Transform every item; the callback may take `(item)` or `(item, index)` |
| `filter(arr, fn)` | array[T], fn(T, [int]) -> bool | array[T] | Keep items whose callback returns true |
| `reduce(arr, init, fn)` | array[T], U, fn(U, T, [int]) -> U | U | Fold the array into one value |
| `find(arr, fn)` | array[T], fn(T, [int]) -> bool | any | First matching item, or `null` |
| `findIndex(arr, fn)` | array[T], fn(T, [int]) -> bool | int | Index of the first match, or `-1` |
| `any(arr, fn)` | array[T], fn(T, [int]) -> bool | bool | Whether at least one item matches |
| `all(arr, fn)` | array[T], fn(T, [int]) -> bool | bool | Whether every item matches |
| `sortBy(arr, fn)` | array[T], fn(T, [int]) -> any | array[T] | Sorted copy ordered by the callback key |
| `unique(arr)` | array[T] | array[T] | Remove duplicates by value/reference identity |

## Map/Object Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `keys(m)` | map | array[string] | Map keys |
| `values(m)` | map | array[T] | Map values |
| `merge(m1, m2)` | map, map | map | Merge two maps |
| `has(m, k)` | map, string | bool | Whether the map contains a key |
| `removeKey(m, k)` | map, string | map | Remove a key (new map) |

## Math Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `int(v)` | number | int | Truncate to integer (`Math.trunc`) |
| `float(v)` | number | float | Convert to number (`Number`) |
| `min(a, b)` | number, number | number | Minimum |
| `max(a, b)` | number, number | number | Maximum |
| `abs(n)` | number | number | Absolute value |
| `round(n)` | number | int | Round |
| `floor(n)` | number | int | Floor |
| `ceil(n)` | number | int | Ceil |
| `pow(a, b)` | number, number | float | Power (`a^b`) |
| `sqrt(n)` | number | float | Square root |
| `random()` | - | float | Random number in `[0,1)` |
| `randint(min, max)` | int, int | int | Random integer in `[min,max]` |

## Timer Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `after(ms, fn)` | number, fn() | null | Run a declared function after a delay |
| `interval(ms, fn)` | number, fn() | null | Run a declared function periodically |
| `timers.cancel(id)` | number | null | Clear a timeout/interval |
| `timers.debounce(fn, ms)` | fn(), number | int | Create a debounced wrapper; returns the timer id |
| `timers.throttle(fn, ms)` | fn(), number | int | Create a throttled wrapper; returns the timer id |

## Time/Date Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `time.now()` | - | any (Date) | Current date/time |
| `time.nowMs()` | - | int | Current timestamp in milliseconds (`Date.now`) |
| `time.format(dt, [fmt])` | any, string? | string | Format with a custom pattern (`YYYY-MM-DD HH:mm:ss`) or `iso` |
| `time.parse(s, [fmt])` | string, string? | (any, any) | Parse with a custom pattern (`YYYY-MM-DD HH:mm:ss`) |
| `time.add(dt, amt, [unit])` | any, number, string? | any (Date) | Add a duration; units: ms, s/sec, m/min, h/hour, d/day, w/week, mo/month, y/year (default ms) |
| `time.diff(a, b, [unit])` | any, any, string? | int | Difference `a-b`; units: ms, s/sec, m/min, h/hour, d/day (default ms) |
| `time.toISO(dt)` | any | string | Convert to ISO 8601 |
| `time.toMs(dt)` | any | int | Convert to a millisecond timestamp |

## URL/Encoding Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `url.encode(s)` | string | string | `encodeURI` |
| `url.decode(s)` | string | (string, any) | `decodeURI`; returns `(value, err/null)` |
| `url.encodeComponent(s)` | string | string | `encodeURIComponent` |
| `url.decodeComponent(s)` | string | (string, any) | `decodeURIComponent`; returns `(value, err/null)` |
| `encoding.base64Encode(s)` | string | string | Base64 encode (UTF-8 safe) |
| `encoding.base64Decode(s)` | string | (string, any) | Base64 decode; returns `(value, err/null)` |

## Storage Functions (Cross-target)

Web uses `localStorage`; desktop falls back to in-memory storage on failure.
All functions return `(value, err)` or `(ok, err)`.

| Function | Params | Return | Description |
|---|---|---|---|
| `storage.get(k, [def])` | string, any? | (any, any) | Read a value (JSON parsed when possible); missing -> def/null |
| `storage.set(k, v)` | string, any | (bool, any) | Store a value (JSON for non-strings) |
| `storage.remove(k)` | string | (bool, any) | Remove a key |
| `storage.clear()` | - | (bool, any) | Clear storage |
| `storage.keys()` | - | (array[string], any) | List keys |

## Clipboard Functions

Web requires permissions/HTTPS (Clipboard API); desktop uses the bridge when
available.

| Function | Params | Return | Description |
|---|---|---|---|
| `clipboard.readText()` | - | (string, any) | Read clipboard text |
| `clipboard.writeText(t)` | string | (bool, any) | Write clipboard text |

## Validation Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `validation.isEmail(s)` | string | bool | Validate email format (regex) |
| `validation.isUrl(s)` | string | bool | Validate a URL (`new URL`) |
| `validation.isEmpty(v)` | any | bool | Empty: null/undefined, empty/trimmed string, empty array/object |
| `validation.required(v)` | any | bool | Required: `!isEmpty` |
| `validation.minLen(v, n)` | any, number | bool | Minimum length (string/array/object keys) |
| `validation.maxLen(v, n)` | any, number | bool | Maximum length |
| `validation.length(v, a, [b])` | any, number, number? | bool | Exact length (2 args) or range `[a,b]` (3 args) |
| `validation.pattern(s, regex)` | string, string | bool | Regex test (`new RegExp`) |
| `validation.isNumber(v)` | any | bool | `typeof number` and not NaN |
| `validation.isBool(v)` | any | bool | `typeof boolean` |
| `validation.isString(v)` | any | bool | `typeof string` |
| `validation.isArray(v)` | any | bool | `Array.isArray` |
| `validation.isObject(v)` | any | bool | Object, not null/array |

## Drag and Drop Functions

Drag handlers are zero-argument functions. These accessors read the payload
captured by the browser runtime; see
[`API_REFERENCE.md`](API_REFERENCE.md#drag-and-drop) for the attributes and the
[GUI editor guide](GUI_EDITOR.md) for inspector configuration.

| Function | Params | Return | Description |
|---|---|---|---|
| `drag.source()` | - | string | Source key from `dragKey` |
| `drag.target()` | - | string | Current/drop-target key from `dropKey` |
| `drag.data()` | - | string | `text/plain` drag payload |
| `drag.files()` | - | array[map] | File metadata (`name`, `path`, `size`, `type`, `isDirectory`) |
| `drag.folder()` | - | string | Common top-level folder of the selection, or `""` |
| `drag.error()` | - | string | Last rejection reason |
| `drag.active()` | - | bool | Whether a drag is active |
| `drag.clear()` | - | null | Clear the runtime drag state |

`drag.files()` contains immediate top-level metadata. Recursive directory
collection is asynchronous and fills a `dropFiles:` state when it finishes.
A `pickFiles:` selection from the native dialog also populates `drag.files()`.

## Navigation

| Function | Params | Return | Description |
|---|---|---|---|
| `navigate(name)` | string | null | Show another page (non-invoked root view) |
| `modal.open(name)` | string | null | Open the modal with that `name` |
| `modal.close(name)` | string | null | Close the modal with that `name` |

## Internationalization Functions

`i18n "locale" { ... }` declares a catalog; `t` and the `i18n.*` helpers need no
import. The first declared catalog is the fallback. See
[API reference: i18n](API_REFERENCE.md#internationalization-i18n).

| Function | Params | Return | Description |
|---|---|---|---|
| `t(key, [vars])` | string, map? | string | Translate a key; `vars` supplies `{placeholder}` values and `count` for plurals |
| `i18n.t(key, [vars])` | string, map? | string | Alias of `t` |
| `i18n.locale()` | - | string | Active locale code |
| `i18n.setLocale(code)` | string | bool | Switch language and persist the choice; `"auto"` re-detects |
| `i18n.detect()` | - | string | Best loaded match for the browser languages |
| `i18n.available()` | - | array[string] | Loaded locale codes |
| `i18n.dir()` | - | string | `"ltr"` or `"rtl"` |
| `i18n.has(key)` | string | bool | Whether the key exists in the active or fallback catalog |
| `i18n.missing()` | - | array[string] | Keys present in the fallback but missing from the active locale |

`navigate` needs no import. It is a no-op in single-page apps and for unknown
page names. Use it from a normal zero-argument handler:

```sui
fn openSettings() -> null {
	navigate("settings")
}
```

## Assertions and Locale Formatting

| Function | Params | Return | Description |
|---|---|---|---|
| `assert(cond, message)` | bool, string | null | Throw an `Error` when `cond` is false; useful with `errorboundary` |
| `intl.date(value, [style], [locale], [timeZone])` | string \| int, string?, string?, string? | string | Format a date; `style` is `short`, `medium`, `long`, `full` or `iso` |
| `intl.time(value, [style], [locale], [timeZone])` | string \| int, string?, string?, string? | string | Format a time |
| `intl.datetime(value, [style], [locale], [timeZone])` | string \| int, string?, string?, string? | string | Format a date and time |
| `intl.number(value, [decimals], [locale])` | number, int?, string? | string | Grouped decimal formatting |
| `intl.currency(value, currency, [locale])` | number, string, string? | string | Currency formatting with an ISO 4217 code |
| `intl.percent(value, [decimals], [locale])` | number, int?, string? | string | Percent from a fraction |
| `intl.relative(offset, unit, [locale], [numeric])` | int, string, string?, string? | string | Relative time such as `2 days ago` |
| `intl.timezone()` | - | string | Resolved IANA time zone |

Omitted locales use the active i18n locale, then the browser language. See
[API reference: Intl](API_REFERENCE.md#locale-formatting-intl).

## Security

`cssSafe` is required for **dynamic** `style:` values:

```sui
div "x" style: cssSafe(dynamicCss)   # ok
div "x" style: dynamicCss            # compile-time error
```

The sanitizer is conservative: it removes `url()`, `paint()`, `expression()`,
CSS escapes, positioning/overlay properties and vendor prefixes. Static styles
written in the view are trusted code. See [`SECURITY.md`](../SECURITY.md).
