# SimpleUI (SUI) — Technical Specification (v0.1)

This is the living engineering contract for the SimpleUI compiler, written in
Go. It fixes the source-language grammar, the compiler's phase interface, and
the *generated code* so every contributor builds against the same contracts.

Version: **0.1.0-alpha** - Targets: **web**, **desktop** (system browser) and
**desktop-webview** (embedded native WebView).

---

## 1. Repository layout (adapted from the Design Document)

The Design Document proposes `cmd/`, `pkg/`, `runtime/` subdirectories. **The Go
1.24 toolchain resolves packages by directory and its module mode only
resolves std + required modules for import paths; local intra-module
subpackages are not resolvable in a module build.** We therefore use a **flat
single-package layout**: every `.go` file lives in the module root and declares
`package main`. Logical ownership is preserved by file names. All functions
that cross files are **Capitalized** (Go lower-case names are file-scoped).

```
/ (module root, package main)
├── go.mod               module example.com/simpleui
├── main.go              CLI entry (completed at integration time)
├── version.go           AppName / AppVersion / AppTagline
├── errors.go            Diag, DiagList
├── ast.go               AST node types (shared contract)
├── stdlib.go            standard library signatures
├── lexer.go             tokenizer            (front-end)
├── parser.go            recursive-descent     (front-end)
├── astjson.go           AST -> JSON dump      (front-end)
├── typechecker.go       inference + checks    (front-end)
├── jsgen.go             -> app.js             (codegen)
├── htmlgen.go           -> index.html + style.css (codegen)
dnd.go               browser drag-and-drop runtime
├── *_test.go            unit tests per module (Go `go test` files)
├── testdata/            .sui fixtures (NO .go files here)
└── examples/            sample .sui apps (NO .go files here)
```

Build/test commands (run from the module root):

```
go build         # produces ./simpleui.exe
go test          # runs every *_test.go TestXxx
```

---

## 2. The SUI language

### 2.1 Lexical

- Comments: `#` to end of line.
- Identifiers: `[A-Za-z_][A-Za-z0-9_]*`.
- Integers: digits only (`0`, `42`). Floats: `digits.digits` (`3.14`). No
  signs in literals: `-x` is unary minus.
- Strings: double-quoted `"..."` with escapes `\" \\ \n \t` and interpolation
  `{ expr }`. Use `\{` / `\}` for literal braces.
- Keywords: `state fn async if else loop in return break continue let import
  from view title true false null and or not`.
- Operators/punct: `( ) { } [ ] = == != < <= > >= + - * / % -> . , : ..`.
- Newlines are significant (statement boundaries). Blocks ignore them.
- Multi-char operators match greedily (`->`, `..`, `==`, `!=`, `<=`, `>=`).

### 2.2 Program structure (top level only)

```
Program  := { ImportDecl | TypeDecl | StateDecl | CssDecl | FuncDecl | ViewDecl }

ImportDecl := 'import' '{' ident { ',' ident } '}' 'from' String
TypeDecl   := 'type' ident 'struct' '{' Field { ',' Field } '}'
Field      := ident ':' TypeName
StateDecl  := 'state' ident [ ':' TypeName ] '=' Expr
CssDecl    := 'css' String
FuncDecl   := [ 'async' ] 'fn' ident '(' [ Params ] ')' [ '->' ( TypeName | '(' TypeName ',' TypeName ')' ) ] Block
Params     := Param { ',' Param }
Param      := ident ':' TypeName [ '=' Expr ]
ViewDecl   := 'view' ident '{' ViewBody '}'
```

- `TypeName` (string) is one of: `int float string bool array[T] map[string]V
  any null`. For `state`, the type may be inferred from the initializer.
- Function **parameters and the return type must be annotated** in v0.1
  (inference applies to `let`/`state` bindings, not to function headers).
- `css "…"` declares a **raw CSS snippet** appended verbatim to the generated
  `style.css` (it overrides/augments the default look — useful for custom
  classes such as `.card { … }`). The string follows the usual SUI escaping
  (`\n`, `\{`, `\}` …), so CSS braces are written escaped in the source. Any
  number of `css` declarations is allowed.

### 2.2a User-defined struct types

```
type Contact struct { id: string, name: string, email: string }

state contacts: array[Contact] = [
    Contact{ id: "1", name: "Alice", email: "alice@example.dev" },
    Contact{ id: "2", name: "Bob",   email: "bob@example.dev" }
]
```

- A declared struct name is a valid `TypeName` anywhere (state/param/return/
  array/map value). Constructors are `Contact{ field: expr, ... }` with ident
  keys (fields may be on separate lines); every declared field must be present
  exactly once and match its declared type.
- Field access `c.name` works in expressions and view interpolation
  (`p "{c.name}"`). Unknown type/field names and missing fields are compile
  errors with "did you mean" hints.
- Structs compare by field (struct-vs-struct `==` is rejected; compare fields
  instead in v0.5). Fields ARE mutable via `u.name = ...` (see data mutation
  below) and `type`/`struct` are reserved keywords.

### 2.3 Statements (inside functions)

```
Block     := '{' { Stmt } '}'
Stmt      := LetDecl | Assign | IfStmt | LoopStmt | WhileStmt | ReturnStmt
           | BreakStmt | ContinueStmt | ExprStmt
LetDecl   := 'let' ident '=' Expr
Assign    := ( LValue | Ident { ',' Ident } ) ( '=' | CompoundOp ) Expr
Postfix   := ident '++' | ident '--'
LValue    := Ident { '.' ident | '[' Expr ']' }
IfStmt    := 'if' Expr Block { 'else' 'if' Expr Block } [ 'else' Block ]
LoopStmt  := 'loop' ( ident | ident ',' ident ) 'in' Expr [ '..' Expr ] Block
WhileStmt := 'while' Expr Block
ReturnStmt:= 'return' [ Expr ]
```

- **Optional parameters**: `fn greet(name: string, excited: bool = true)`.
  Defaults must be trailing (a required parameter after an optional one is an
  error) and are type-checked at the declaration; call sites may omit them and
  the generator appends the default expression. Missing *required* arguments
  are still errors.
- `assign` with **two** targets must be either a std intrinsic call on the RHS
  (tuple return, e.g. `body, err = http.get("https://api.example.com/data")` or
  `body, err = http.getSafe(url, "https://api.example.com")`) or a **user function that
  declares multiple return values**: `fn divmod(a: int, b: int) -> (int, int)`
  with `return a / b, a % b`; call sites bind them as `q, r = divmod(7, 3)`.
  Tuple targets must be fresh local names (assign into states afterwards).
  Calling a multi-return fn in expression position is an error.
- Multi-return type list: `'->' '(' TypeName ',' TypeName ')'`. Counts and
  types of `return` values are checked against the declaration. JS sync fns
  return the tuple array directly (`const r = f()`); async fns are awaited.
- Compound assignment (`+= -= *= /= %=`) is sugar: `x += e` compiles exactly
  like `x = x + e` (single target only). Postfix `x++` / `x--` are sugar for
  `x += 1` / `x -= 1` on an identifier.
- **Data mutation** (v0.5): struct fields and array/map elements can be
  assigned directly — `u.name = "x"`, `arr[i] = v`, `m[k] += 1` (compound
  included). When the root is a `state`, the compiled JS mutates in place and
  re-syncs through the setter, so the UI re-renders exactly like a whole-state
  assignment. Field/member targets are type-checked (unknown field, bad index
  type, mismatched value) and must live inside a function.
- `loop i in 0..10 { }` iterates an int range; `loop x in items { }` iterates
  an array; `loop k, v in m { }` iterates a **map** (k: string key, v: the
  value type). `while cond { }` runs until the (bool) condition is false.
  `break`/`continue` are allowed inside `loop` and `while`. The same two-name
  form works in views: `list k, v in m { ... }`. Riding a map with one name
  (or an array/map/range with two) is a compile error.

### 2.4a Built-in functions (no import needed)

`Builtins` (builtins.go) — the batteries that need no `import`:

| builtin | signature | notes |
|---------|-----------|-------|
| len     | `len(string|array[T]) -> int` | JS `.length` semantics |
| append  | `append(array[T], T) -> array[T]` | returns a new array (`[...a, x]`) |
| upper / lower / trim | `(string) -> string` | |
| contains | `(string, string) -> bool` | |
| split   | `(string, string) -> array[string]` | separator string |
| join    | `(array[string], string) -> string` | |
| substring | `(string, int, int) -> string` | start index + count |
| str     | `(any) -> string` | |
| int     | `(number) -> int` | truncates (`Math.trunc`) |
| float   | `(number) -> float` | |
| keys    | `(map[string]V) -> array[string]` | |
| values  | `(map[string]V) -> array[V]` | |
| merge   | `(map, map) -> map` | second map wins on duplicate keys (`Object.assign`) |
| has     | `(map[string]V, string) -> bool` | |
| removeKey | `(map[string]V, string) -> map` | copy without the key |

**String built-ins (v0.6)**

| builtin | signature | notes |
|---------|-----------|-------|
| replace | `(string, string, string) -> string` | all occurrences (`replaceAll`) |
| startsWith / endsWith | `(string, string) -> bool` | |
| indexOf | `(string, string) or (array[T], T) -> int` | first index, JS semantics (-1 when absent) |
| lastIndexOf | `(string, string) -> int` | |
| slice | `(string, int, int) -> string` | start/end indexes (`"abc".slice(1,3)` = "bc") |
| repeat | `(string, int) -> string` | |
| padLeft / padRight | `(string, int, string) -> string` | pads with the fill string |
| trimLeft / trimRight | `(string) -> string` | |
| chars | `(string) -> array[string]` | code points |
| fmt | `(number, int) -> string` | `toFixed` decimals |

**Collection built-ins (v0.6)**

| builtin | signature | notes |
|---------|-----------|-------|
| concat | `(array[T], array[T]) -> array[T]` | new array |
| remove | `(array[T], int) -> array[T]` | copy without the index |
| includes | `(array[T], T) -> bool` | |
| sort / reverse | `(array[T]) -> array[T]` | new array (JS sort semantics) |
| moveItem | `(array[T], int, int) -> array[T]` | new array with one item moved |
| removeKey | `(map[string]V, string) -> map` | copy without the key |

**Drag-and-drop builtins**

| builtin | signature | notes |
|---------|-----------|-------|
| drag.source / drag.target | `() -> string` | source/drop keys |
| drag.data | `() -> string` | `text/plain` payload |
| drag.files | `() -> array[map]` | metadata only |
| drag.folder | `() -> string` | common top-level folder, or empty |
| drag.error / drag.active | `() -> string` / `() -> bool` | rejection reason / active flag |
| drag.clear | `() -> null` | clear runtime state |

**Timers (v0.7)**

| builtin | signature | notes |
|---------|-----------|-------|
| after | `(int, fn) -> null` | runs the function once after `ms` milliseconds |
| interval | `(int, fn) -> null` | runs the function every `ms` milliseconds (returns `null`) |

The fn argument must be the **name of a declared function**; the generator
emits `setTimeout(() => name(), ms)` / `setInterval`. Perfect for
auto-refreshing dashboards:

```
state ticks = 0
fn bump() -> null { ticks += 1 }
fn boot() -> null { interval(500, bump) }
view main { onLoad: boot  p "ticks: {ticks}" }
```

**Math (v0.7)**

| builtin | signature |
|---------|-----------|
| min / max | `(number, number) -> number` (int when both int) |
| abs | `(number) -> number` (float for float) |
| round / floor / ceil | `(number) -> int` |
| pow | `(number, number) -> float` |
| sqrt | `(number) -> float` |
| random | `() -> float` in `[0,1)` |
| randint | `(int, int) -> int` in `[lo, hi]` inclusive |

A user `fn` with the same name shadows a built-in. Types and arity are
validated by the typechecker; the JS generator translates each call into the
native expression from the `Builtins` table.

### 2.4 Expressions (precedence, low to high)

```
or   -> and  -> cmp(== != < <= > >=) -> add(+ -) -> mul(* / %) -> unary(not -)
postfix: Primary { '.' ident |  '[' Expr ']' | '(' Args ')' }
Primary  := Int | Float | String | true | false | null | ident
          | '(' Expr ')' | '[' [ Expr { ',' Expr } ] ']'
          | '{' [ String ':' Expr { ',' String ':' Expr } ] '}'
```

- `http.get("https://api.example.com/data")` is a member expression `http.get`
  applied with one argument; dynamic destinations use `http.getSafe(url, allow)`.
- Array literal `[1,2,3]` infers `array[int]`; map literal `{"a": 1}` infers
  `map[string]int`. Mixed-element arrays are an error (except all-`null`).

### 2.5 Views

```
ViewBody := { 'title' String | 'onLoad' ':' ident | Element | If | List }
Element  := Tag [ String ] { Attr } [ '{' ViewBody '}' ]
Attr     := ident ':' ( String | Expr )
If       := 'if' Expr '{' ViewBody '}' [ 'else' '{' ViewBody '}' ]
List     := 'list' ident 'in' Expr '{' ViewBody '}'
Tag      := h1 | h2 | h3 | p | text | button | input | img | row | col | spacer | a
          | table | thead | trow | tcell | tabs | tab | modal | modalbutton
          | tablelayout | menubar | menu | menuitem
```

`onLoad: fn` (anywhere in a view, at most once) makes the compiled app call
`fn()` once right after the first render, so async data loaders can populate
the UI immediately on startup.

Elements and their supported attributes (v0.1):

| tag      | text | attrs                                    | children |
|----------|------|------------------------------------------|----------|
| h1..h3,p | yes  | class                                    | no       |
| text     | yes  | class                                    | no       |
| button   | yes  | class, onClick (fn)                       | no       |
| input    | yes* | class, placeholder, value (state), onChange (fn) | no |
| img      | no   | class, src (string/expr), alt             | no       |
| a        | yes  | class, href (string/expr)                 | no       |
| row, col | no   | class                                    | yes      |
| spacer   | no   | class                                    | no       |
| table    | no   | class, name, editable, fixed             | yes      |
| tablelayout | no | class, style, rows, columns, width, height, columnWidths, rowHeights, gapX, gapY, borderStyle, borderWidth, borderColor, cellMinHeight | yes      |
| thead    | no   | class                                    | yes      |
| trow     | no   | class                                    | yes      |
| tcell    | yes  | class, head (bool)                       | no       |
| tabs     | yes  | class, style                             | tab      |
| tab      | yes  | class, style                             | yes      |
| modal    | yes  | class, style, name                       | body + modalbutton |
| modalbutton | yes | class, style, variant, close, onClick | no |
| menubar  | yes  | class, style                             | menu     |
| menu     | yes  | class, style                             | menuitem, menu, divider |
| menuitem | yes  | class, style, link, onClick (fn)          | no       |

\* `input` text is treated as its `placeholder`.

A `tabs` renders a `role=tablist` and panels from its direct `tab` children.
Each `tab` text is the button label and its children are the panel content.
The first tab is active initially; the runtime switches tabs by click and
Left/Right/Home/End and restores the active tab after a page-level re-render.
Only direct `tab` children are tabs; labels are escaped like other view text.

Each non-invoked root `view` is a page. Multi-page programs keep every page
in the DOM but only the first is visible initially. The no-import builtin
`navigate(name)` switches to the named page (`__sui_page` at runtime); it is a
no-op in single-page programs and for unknown names. Page state is shared by
all pages.

A `modal` is hidden initially and identified by its `name`; `modal.open(name)`
and `modal.close(name)` show/hide it. Its direct `modalbutton` children form the
footer; other children are body content. Buttons close the modal by default
(`close: false` keeps it open), run `onClick` first, and support the `primary`
or `danger` variant. Backdrop click and Escape close the active modal; focus
returns to the opener when it is still available.

A `menubar` renders a `nav[role=menubar]`; its text is the accessible name
(`aria-label`), not a visible label. Each `menu` renders a button plus a hidden
popup; the button toggles it at runtime and Escape/outside click closes it.
Each `menuitem` is a `role=menuitem` entry. `link` uses the normal URL guard;
`onClick` names a zero-argument function. A `menu` nested in another `menu`
popup is a submenu and may itself contain `menu` children; nesting is unlimited.
Labels are escaped like other view text.

The table family renders real HTML table structure: `table` > (`thead` >
`trow` > `tcell head: true`) / dynamic rows via `list`. `tcell head: true`
emits `<th>`, plain cells `<td>`. Dynamic body rows are produced with a
`list` inside `table`:

```
table class: "users" {
    thead { trow { tcell "ID" head: true ... } }
    list r in rows { trow { tcell "{r[0]}" ... } }
}
```

Interpolation in view text refers to **state variables** (and is how views get
their reactive bindings). In expressions, `onClick`/`onChange` must name a
declared function; `value`/`src`/`href` may name a state or any expression.

#### Drag and drop

Every element accepts `draggable: bool`, `dragKey: expr`, `dropKey: expr`,
`dropFiles: state`, `pickFiles: state`, `pickMode: "native"|"fs"|"auto"`,
`directory: bool`, and the event handlers `onDragStart`, `onDragEnter`,
`onDragOver`, `onDragLeave`, `onDrop`, `onDragEnd`. A `dragKey` makes an element draggable even when `draggable` is
absent. `dropFiles` and `pickFiles` require a declared string state; the runtime
calls its setter with dropped/selected paths (comma-separated on an input,
newline on a textarea) and writes the same text into the control when
applicable.

`pickFiles` uses the native `<input type=file>` on a `file` control and creates
a hidden one on another trigger (for example a `button`) when clicked.
`accept`, `multiple` and `directory` configure either form. `pickMode: "fs"`
uses `showDirectoryPicker()` instead (Chromium secure contexts, read-only,
user-granted permission) and traverses the directory handle; `pickMode: "auto"`
uses it when available and otherwise falls back to native. FS mode can retain
the name of an empty folder; native `webkitdirectory` cannot. A picker selection
uses the same metadata path and populates `drag.files()`; errors use
`drag.error()` and the same limits as drops. With `directory: true` the native
folder picker returns all files recursively; `drag.folder()` exposes the common
top-level folder name when paths share one. Selecting an arbitrary subset inside
one folder requires the multi-file picker or app-side filtering. An empty folder
chosen by the dialog has no file metadata/name; a dropped empty folder may be
represented by a synthetic directory entry in Chromium but has no bytes.

Handlers remain zero-argument functions. The runtime exposes the captured
payload through `drag.source()`, `drag.target()`, `drag.data()`,
`drag.files()` (array of maps with `name`, `path`, `size`, `type`,
`isDirectory`), `drag.error()`, `drag.active()` and `drag.clear()`.
`moveItem(array[T], int, int) -> array[T]` returns a moved copy.

Dropped files are metadata only: bytes are not read. Directory traversal uses
`webkitGetAsEntry`, is asynchronous, and is capped at 2,000 files, depth 16 and
4,096 characters per path. Browsers do not expose absolute local paths; a
dropped name is untrusted text and does not grant filesystem access. DnD
handlers are emitted inline on root elements so repeated `list` rows each
receive a handler; component bodies use their existing inline dispatch.

#### Runtime table data (`table` core module)

Give a table a `name:` and it becomes a runtime handle that can be filled from
any tabular payload while keeping the look configured in the view:

```
table name: "users" {
    thead { trow { tcell "ID" head: true ... } }
    tbody { trow { tcell "" ... } }   # style template for generated rows
}
```

`table.display(name, data)` accepts JSON text (a string, parsed automatically),
an array of arrays, an array of objects (first-seen keys become the header;
`{"columns": [...], "rows": [...]}` is also accepted), a key/value object, or a
scalar. Generated rows and cells clone the view's row/cell classes and inline
styles; table classes, caption, presets, editable/fixed flags and the scroll
wrapper are untouched. Object data sets the header labels and the column count;
array data keeps the designed header and columns. Displayed data survives view
re-renders (state changes), and cells edited in `editable` mode keep their value.

Strings are sniffed: JSON first (`[`/`{`), then JSON Lines (one JSON value per
line), then delimited text. Delimited text auto-detects `,`, `;`, tab or `|`,
understands RFC 4180 quoting (embedded delimiters/newlines, `""` escapes,
CRLF, BOM) and uses the first row as the header. When the designed table has no
`<thead>`, that first row stays visible as the first body row. A one-row payload
is treated as data, not as a header.

Read back the visible table with:

| call | returns |
|------|---------|
| `table.data(name)` | body rows as `array[array[string]]` (includes edits) |
| `table.header(name)` | header labels as `array[string]` |
| `table.select(name, cols)` | body rows restricted to the `array[int]` columns |
| `table.slice(name, from, count)` | a block of body rows (`count <= 0` = to the end) |
| `table.importText(name, text, spec)` | explicit CSV/TSV/SSV/PSV/JSON/JSONL parse |
| `table.exportText(name, spec)` | CSV/TSV/SSV/PSV/JSON/JSONL text |
| `table.json(name)` | JSON string: array of objects keyed by the header (arrays when no header) |
| `table.clear(name)` | empties the body rows |

The `spec` argument is `format` optionally followed by `;header=0|1`,
`;sep=comma|semicolon|tab|pipe|<char>`, `;quote=<char>`, `;trim=1` and
`;formula=escape` (CSV only: prefix cells starting with `= + - @ TAB CR` with
`'` to neutralise spreadsheet formulas; opt-in). Formats: `csv`, `tsv`, `ssv`
(`;`), `psv` (`|`), `json`, `jsonl`, `auto`.
Examples: `"csv"`, `"csv;header=0"`, `"ssv"`, `"tsv;trim=1"`, `"jsonl"`,
`"csv;formula=escape"`.
`exportText` includes the header line unless `header=0`; CSV fields are quoted
when they contain the delimiter, a quote or a newline.

Runtime limits: at most 100000 rows, 10000 columns and 20 MB of text per call;
larger payloads return an error. Cells are always written with `textContent`,
never parsed as HTML.

### Security semantics of views

Interpolated text and attributes are **HTML-escaped automatically**; there is
no raw-HTML interpolation. `href`, `src` and `link` additionally pass through a
scheme filter: relative URLs, `http(s)`, `mailto` and `tel` are allowed, while
`javascript:`, `vbscript:`, `data:` and other active schemes become `#`.
`style`/`class` values are escaped but remain developer-controlled presentation:
do not bind them to untrusted data. See `SECURITY.md` for the full model and
`examples/security/` for runnable examples.

Every call returns `(value, error)`, e.g. `rows, err = table.data("users")`.
`table` is a **core module**: it is always available and needs no `import`.
These helpers are synchronous because they only touch the rendered DOM, so they
can be called from ordinary (non-async) functions.

```
state status = "ready"

fn load() -> null {
    n, err = table.display("users", [["1", "Alice"], ["2", "Bob"]])
    if err != null { status = err } else { status = str(n) + " rows" }
}

fn export() -> null {
    part, err = table.select("users", [0])
    body, err2 = table.json("users")
    status = body
}
```

#### Reactive attributes (class, style, and any expression attribute)

An attribute value is a full expression. This is what lets a function change how
an element *looks* without needing to name the element: the function mutates a
`state`, and the attribute re-renders from that state.

- `class:` accepts a static string, a template with `{state}`, **or** an
  expression. An expression is *appended* to the element's built-in class, so
  built-in styling is preserved:
  ```
  card "Hello" class: active ? "on" : ""     # -> class="sui-card on" or "sui-card "
  ```
- `style:` accepts a string template or an expression producing CSS declarations:
  ```
  p "…" style: "font-size:" + str(size) + "px"
  ```
- Other attributes (`value`, `src`, `href`, `disabled`, …) already accept
  expressions and react to state the same way.

Boolean attributes are special: `checked` on `checkbox`/`switch`/`radio` is a
*presence* attribute, so the compiler emits it only when the binding is truthy
(`checked="false"` would still render as ON).

Label-bearing controls render as **wrapper `<label>` widgets**: a
`checkbox`/`switch`/`radio` with text (or `label:`) emits

```
<label data-sui-id="i" class="sui-checkbox-wrap <userclass>">
    <input type="checkbox" class="sui-checkbox <userclass>" ...>
    <span>label text</span>
</label>
```

(the wrapper class is `sui-switch-wrap` for `switch` and `sui-radio-wrap` for
`radio`). The element's badge `data-sui-id` and any user class apply to the
wrapper **and** the inner control, so a click on the label text selects the
element in the design tool and a user rule (background, padding, font) styles
the whole widget; the built-in `sui-*` base classes stay on the inner control
so skin states (`:checked`) keep working. The runtime wires handlers on the
wrapper — the inner control's DOM events bubble to it — which means an element
with both a state binding and an explicit `onChange` toggles exactly once.

Binding a control's value to state also wires the matching DOM event *unless the
element already declares that handler*. If you write your own `onChange` (or
`onInput`), it **owns** the event: the compiler will not add an automatic state
setter, so your function is free to update whatever state it wants (including a
counter of a different type). With no explicit handler, the state is updated
automatically from the control's value.

A value control **must stay bound to state**: because views re-render by
rebuilding their HTML, a control whose value does not reference any declared
state (a static literal like `checked: true` or `value: 5` does not count) and
that declares no event handler is **inert** — nothing can ever read or set its
value, and every re-render resets it. The checker flags this as a

```
warning: input control is not bound to any state and has no handler;
         nothing reads its value and re-renders reset it
    = bind it with value: <state> to make it live
```

covering every control family (`switch`/`checkbox`/`radio` with `checked:`,
`input`/`textarea`/`select`/`number`/`slider`/`color`/`date`/`time` with
`value:`, and `file` which only makes sense with a handler). `button` is
checked the same way: a button without an event handler is inert (a click does
nothing) and warns "button has no handler, so nothing reacts to it" with hint
"add onClick: <fn> to react to the click". A control that declares a handler
but has no state binding is flagged too ("has a handler but no checked:/value:
binding, so it will not reflect state"). Warnings never block the build.

```
state on = false
fn toggle() -> null { on = on == true ? false : true }

view main {
    switch "" checked: on onChange: toggle   # handler owns the state; toggles on click
    p "on: {str(on)}"
}
```

```
state on = false
view main {
    switch "" checked: on                    # no handler: state follows the control
}
```

The language has a **conditional (ternary) expression** `cond ? a : b` for this
kind of inline choice; both branches must have the same type:

```
badge "ON" class: active ? "sui-badge on" : "sui-badge off"
```

A complete example — a function changes colour and size of an element through
state, with no element id:

```
state active = false
state bg     = "#1b1f27"
state size   = 16

fn toggle() -> null {
    if active == false {
        active = true
        bg = "#2563eb"
        size = 28
    } else {
        active = false
        bg = "#1b1f27"
        size = 16
    }
}

view App {
    col {
        card "Cambia colore" class: active ? "on" : "" style: "background:" + bg
        p "Dimensione: {size}" style: "font-size:" + str(size) + "px"
        button "Toggle" onClick: toggle
    }
}
```

### 2.5a Components (parameterized views)

A view may declare parameters; any other view can then invoke it as a
component:

```
view Card(title: string, note: string) {
    h2 "{title}"          # parameters are locals in the view body
    p "{note}"
}

view main {
    Card title: "Hello" note: "a reusable view"
    list c in cards { Card title: c ... }
}
```

- Grammar: `view Name ( ident ':' TypeName, ... ) { ViewBody }`. The parameter
  types follow `TypeName`, so structs and arrays work as parameters.
- Invocation is an ident element whose tag names a declared view:
  `Card title: "x" note: y`. Values may be string literals (interpolated) or
  any expression, checked against the parameter type. Every parameter must be
  provided exactly once; unknowns/missing/wrong types are compile errors.
- Rendering: each component becomes a tiny chunk function
  `__html_<Name>({ p1, p2 }) { return <html>; }`, invoked inline from the
  parent's template. Components have no root renderer and never touch ``#app``.
- Events inside components use **inline handlers** on global functions
  (`onclick="inc()"`, `oninput="setX(this.value)"`), so repeated instances
  never collide on ids.
- **Handler parameters**: a parameter typed `fn` accepts the name of a
  declared function; the component wires it inline at runtime:

  ```
  view Switch(label: string, action: fn) {
      button "{label}" onClick: action
  }
  view main {
      Switch label: "Go"   action: turnOn
      Switch label: "Stop" action: loudOff
  }
  ```

  The argument must be a declared function; the generated attribute becomes
  `onclick="turnOn()"` / `onclick="loudOff()"` per instance.
- **Children / slots**: a component can render the *caller's block* via the
  `children` marker (at most once, no text/attrs):

  ```
  view Panel(heading: string) {
      h2 "{heading}"
      children                      # ← slot
  }
  view main {
      Panel heading: "Box" {
          p "inside {clicks}"       # caller content
          button "Go" onClick: go   # caller button → root wiring
      }
  }
  ```

  The caller's block is generated in the caller's scope: its states are
  reactive, its interactive elements use the root `data-sui-id` wiring, and
  (with a slot) the component chunk signature gains a `children` parameter
  (`__html_Panel({ heading, children })`) rendered by the marker. A `children`
  marker outside a component is an error.
- **Per-instance state & view-local methods** (v0.12): a component may own
  `state` and `fn` declarations whose scope is *that instance*:

  ```
  view Counter(start: int) {
      state count = start        # per-instance, initialized from params
      fn inc() -> null {         # view-local method (instance scope)
          count += 1
      }
      button "{count}" onClick: inc
  }
  view main {
      Counter start: 0
      Counter start: 10
  }
  ```

  Each invocation site gets a stable key into a runtime store
  (`__SUI_inst`); the chunk is `__html_Counter({ __inst: inst, __key, start })`,
  instances initialize their own copy, and inline handlers dispatch through
  `__sui_vm(key, "method")` so a click on instance 0 never touches instance 1.
  Method names are unique across the program and callable **only inside their
  component view** (a top-level call is an error).
- **Reactivity crosses boundaries**: state variables referenced inside a
  component are bound on the root view that invokes it (component parameters
  are not states and are never bound). `data-sui-id` remains a root-only
  mechanism.
- Circular component references (`A` invoking `B` invoking `A`) are compile
  errors.

### 2.6 Semantics (typechecker rules)

1. `state`/`let` infer their type from the initializer; reassignments must be
   type-compatible. `state` may only be **declared at top level**.
2. **Scope rule**: state may be *mutated* only inside functions
   (event-callbacks included). A top-level `counter = ...` is an error.
3. Functions may only be called with the declared parameter types; return type
   is checked. Unknown identifiers are errors. Duplicate top-level names are
   errors (duplicates across a merged import too).
4. Operators: arithmetic on `int`/`int` -> `int`, `float`/`float` -> `float`;
   mixing `int`+`float` is an error in v0.1. `+` on strings concatenates.
   Comparisons return `bool`; `and`/`or`/`not` need `bool`; `if` conditions and
   view-list evaluables need `bool` / `array` / `range`.
5. `array[T]` indexing -> `T`; `map[string]V` indexing -> `V`; loop variable
   types follow the iterable.
6. std intrinsics (see §3) are typed via `stdlib.go` (`StdLookup`). A std call
   is allowed **only** as the RHS of a two-target assignment, and only inside
   `async` functions or event-callback functions of `async`-capable... note:
   std intrinsics may be called from `async fn` only. Calling a user `async fn`
   from a non-async fn is an error.
7. `return` in a function with a declared return type must produce a value of
   that type; bare `return` only in void/`null`-returning functions.
8. Views: unknown element tag, unsupported attribute, or an event handler /
   state reference that does not exist are errors. A view's title, if any, is
   string-only.

---

## 3. Standard library (contract — see `stdlib.go`)

Imported with `import { http, ws, fs, exec, json, dialog } from "std"` (only names
actually used should be imported; the compiler is lenient but the checker
validates they exist in `StdModules`).

Every intrinsic returns **exactly two values**: `(value, errorString)`.
`null` error means success. Intrinsics whose primary value is `null`
(`fs.write/mkdir/remove`, `ws.send/onMessage/close`, `json.stringify`) may
also be called fire-and-forget as plain expression statements (no tuple
binding).

| full name | params | ret0 | ret1 | jsKey |
|-----------|--------|------|------|-------|
| http.get  | string | string | string | http_get |
| http.getStatus | string | int | string | http_getstatus |
| http.post | string, string | string | string | http_post |
| http.postFiles | string | string | string | http_postfiles |
| http.postFilesSafe | string, string | string | string | http_postfilessafe |
| http.getSafe | string, string | string | string | http_getsafe |
| http.getStatusSafe | string, string | int | string | http_getstatussafe |
| http.postSafe | string, string, string | string | string | http_postsafe |
| ws.connect| string | any | string | ws_connect |
| ws.connectSafe | string, string | any | string | ws_connectsafe |
| ws.send   | any, string | null | string | ws_send |
| ws.onMessage | any, string | null | string | ws_onmessage | handler is a declared fn receiving the message |
| ws.close  | any | null | string | ws_close |
| fs.read   | string | string | string | fs_read |
| fs.write  | string, string | null | string | fs_write |
| fs.exists | string | bool | string | fs_exists |
| fs.stat   | string | string | string | fs_stat | "file"/"dir"/"missing" |
| fs.list   | string | array[string] | string | fs_list | directory entries |
| fs.mkdir  | string | null | string | fs_mkdir | recursive |
| fs.remove | string | null | string | fs_remove | file only |
| exec.run  | string, array[string] | string | string | exec_run |
| dialog.pickFile | - | string | string | dialog_pickfile |
| dialog.pickFiles | - | array[string] | string | dialog_pickfiles |
| dialog.pickDirectory | - | string | string | dialog_pickdirectory |
| json.parse| string | any | string | json_parse |
| json.stringify | any | string | null | json_stringify |

`http.postFiles` uploads selected/dropped file bytes as multipart and includes a
parallel `paths` field with browser-visible relative paths for display/grouping.

SSRF policy: `http.get`, `http.getStatus`, `http.post`, `http.postFiles` and `ws.connect` accept
only **literal** URLs. A dynamic URL must use the corresponding `*Safe`
function with a **literal allow-list** as the last parameter (see §9). The
allow-list accepts absolute origins (`https://api.example.com`, optionally with
a path prefix), same-origin path prefixes (`/api/`), and ws(s) origins for
`connectSafe`; entries are separated by commas/whitespace. The guard rejects
credentials, non-http(s)/ws(s) schemes, hosts outside the list, and redirects.

### 3.1 Core `table` module (no import required)

`table` is always available (it is also listed in `StdModules`, so importing it
is allowed but unnecessary). Its calls are synchronous DOM helpers, not I/O.

| full name | params | ret0 | ret1 |
|-----------|--------|------|------|
| table.display | string, any | int | string |
| table.data | string | array[array[string]] | string |
| table.header | string | array[string] | string |
| table.select | string, array[int] | array[array[string]] | string |
| table.slice | string, int, int | array[array[string]] | string |
| table.importText | string, string, string | int | string |
| table.exportText | string, string | string | string |
| table.json | string | string | string |
| table.clear | string | int | string |

The `string` name matches the table's `name:` attribute. See §2.5 for payload
shapes and styling behavior.

Web behavior: `http.*` and `ws.*` use native `fetch`/`WebSocket`;
`fs.*` targets `/api/fs/read` & `/api/fs/write` on the serving backend;
`exec.run` returns `[null, "exec.run is disabled on the web target"]`.

---

## 4. Phase interfaces (drop-in contracts)

All cross-file functions must be **Capitalized**. Error streams are
`DiagList` from `errors.go` (never `panic` for user errors — collect and
print). `Span`/`Diag` positions are 1-based.

### lexer.go — Agent A

```go
type Token struct {
    kind   string // see §2.1 token set
    text   string // literal text (numbers; raw inner text for strings)
    line   int
    col    int
    offset int
}
func Lex(source string, file string) ([]Token, DiagList)
```

Token kinds use the strings `ident int float str state fn async if else loop in
return break continue let import from view title true false null and or not
( ) { } [ ] = == != < <= > >= + - * / % -> . , : .. newline eof err`.
Unknown characters yield an `err` token plus a diagnostic and lexing continues.

### parser.go — Agent A

```go
func ParseSource(source string, file string) (*Prog, DiagList)
func LoadProgram(entryFile string) (*Prog, DiagList)
```

- `ParseSource` parses a single file into a `Prog`.
- `LoadProgram` reads `entryFile`, parses it, then resolves every
  `import { names } from "./rel.sui"`: parses the target file and merges its
  states/functions/views, erroring on duplicate symbols or import cycles.
  `import ... from "std"` only records the declaration.
- String interpolation is parsed in the parser: the lexer returns the raw
  inner text; `parseStringParts` decodes escapes and re-lexes/parses `{...}`
  sub-expressions. Unescaped newline inside a string is an error.

### astjson.go — Agent A

```go
func ProgToJson(p *Prog) string
func AstJsonFromSource(source string, file string) (string, DiagList)
```
Pretty-printed (2-space indent), keys in the same order as the struct fields,
`kind` first. View declarations are emitted with a synthetic leading
`"kind": "view"` field (so every tagged node is self-describing, matching the
Expr/Stmt/ViewNode convention). Needed by `simpleui parse`.

### typechecker.go — Agent B

```go
func Check(prog *Prog) DiagList
```
Implements §2.6 and validates views with the symbol table (event handlers and
state references). Also resolves which state variables each view reads, which
is exposed for codegen:

```go
func ViewStateRefs(v *ViewDecl) []string
```

### jsgen.go — Agent C

```go
func GenerateJs(p *Prog) string
func ExprToJs(e *Expr) string
func StatementsToJs(stmts []*Stmt) string
func EscapeJsStr(s string) string
func EscapeJsTemplate(s string) string
```

Emit order inside `app.js`: prelude, state declarations + setters, user
functions, `__SUI_STD` object (only imported modules), view renderers +
bindings, boot `__sui_render()`. See §5 for the exact runtime contract.

### htmlgen.go — Agent D

```go
func PageTitle(p *Prog) string
func GenerateIndexHtml(p *Prog) string
func GenerateCss(p *Prog) string
```

`PageTitle` returns the first view's literal title text, or "" (no escaping).
`GenerateIndexHtml` embeds an escaped title. `GenerateCss` walks every view and
emits base styles + the `class:` names it finds. See §5.

---

## 5. Generated-web-app contract (the runtime surface)

All asset filenames are fixed: `index.html`, `style.css`, `app.js`.

`index.html` (htmlgen):

```html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>…</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
<div id="app"></div>
<script src="app.js"></script>
</body>
</html>
```

`app.js` precise contract (jsgen). Everything below with a `__sui_` prefix or
`__SUI_` object is part of the runtime and MUST be emitted verbatim when
applicable.

**Prelude** (always emitted):

```js
"use strict";

const __SUI_views = Object.create(null);
const __SUI_bind  = Object.create(null);

function __sui_bindView(stateVar, viewName) {
  (__SUI_bind[stateVar] ||= []).push(viewName);
}
function __sui_react(stateVar) {
  for (const v of (__SUI_bind[stateVar] || [])) __SUI_views[v]();
}
function __sui_render() {
  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();
}
```

**State + reactivity**: `state counter = 0` compiles to

```js
let counter = 0;
function setCounter(v) { counter = v; __sui_react("counter"); }
```

`setCounter` capitalizes the first ASCII letter. Every SUI assignment to a
state inside a function becomes a `setX(...)` call. Assignments to local
`let`s stay plain `x = ...;`. A two-target assignment emits a uniquely-named
temporary (first call in a function uses `__sui_r`, subsequent ones
`__sui_r2`, `__sui_r3`, ... so two std calls in one function never redeclare
a `const`):

```js
const __sui_r = await __SUI_STD.http_get(url);
const body = __sui_r[0];
const err  = __sui_r[1];
```

Reserved names: the runtime globals (`__sui_*`, `__SUI_*`), the generated
setters `set<Name>` per state, and the `render_<view>` renderer functions must
not be reused by user declarations — the typechecker rejects them.

**std module object** — only imported modules appear; entry names from
`stdlib.go` jsKeys; bodies as specified in §3 (`fetch`, `WebSocket`,
`/api/fs/...`, disabled exec). `ws_connect` builds on `class __SUI_WS`
(send/close/open/onMessage/onClose/onError) which is part of the prelude when
`ws` is imported.

**User functions**: `async fn f(a int) -> int { ... }` -> `async function f(a)
{ ... }`; sync fn -> `function f(a) { ... }`. `if/else`, loops, `return` map
1:1. `loop i in a..b` -> `for (let i = <a>; i < <b>; i++)`; `loop x in arr` ->
`for (const x of <arr>)`.

**View renderer**: each view `v` emits one function plus registration:

```js
function render_v() {
  const el = document.getElementById("app");
  el.innerHTML = `<div class="sui-view" data-sui-view="v">`
    + <body JS-html>
    + `</div>`;
  const e0 = document.querySelector('[data-sui-id="0"]');
  e0.onclick = myHandler;
}
__SUI_views["v"] = render_v;
__sui_bindView("counter", "v");   // once per state the view reads
```

- Elements are emitted as real HTML; text content is interpolated inside the
  template literal with `${...}` for `{...}` spans. `data-sui-id="N"` is
  attached to every interactive element (`button`, `input`, `a`) so wiring can
  find it; ids are unique per view, starting at 0.
- `if` in a view: `+ ((${cond}) ? `tpl` : `tpl2`) +`.
- `list i in expr`: `+ (() => { let s = ""; for (const i of (${expr})) { s +=
  `tpl`; } return s; })() +`.
- Buttons/links/inputs expose no `data-sui-id` unless they carry an event
  handler or (for input) a bound value.

Boot statement (last line): `__sui_render();`.

`style.css` base (htmlgen): system font stack, dark palette, `.sui-view` as a
max-width card, `.row`/`.col` flex helpers, `.spacer`, plus base styles for
h1..h3/p/button/input/a/code. Every `class:"foo"` literal found in the views
also gets a pass-through rule `.foo { /* generated */ }` so class names are
stylable by the user's own injected CSS later.

---

## 6. CLI contract (main.go — integration milestone)

```
simpleui version
simpleui help
simpleui parse <file.sui>
simpleui check <file.sui> [--json] [--security <file>]
simpleui fmt <file.sui> [--write]
simpleui docs [--index|--json|--list|--agent|--full|--read <id>|--search <q>]
simpleui serve <file.sui> [--port <n>] [--watch]
simpleui gui [--port <n>]
simpleui lsp
simpleui mcp
simpleui build <file.sui> [--target web|desktop|desktop-webview] [--out <dir>]
simpleui build <file.sui> --single [--minify] [--out <app.html|dir>]
```

- `check` parses and typechecks without emitting assets. `--json` returns
  `{ok, file, diagnostics}` with line/column/kind/message/hint and exits 1 on
  errors; `--security <file>` selects the policy explicitly. This is the
  intended edit loop for CI, editors and coding agents.
- `docs` serves the embedded documentation (also usable outside the repo):
  default index, `--json` manifest, `--list`, `--agent` compact context,
  `--full` concatenated docs, `--read <id>` and `--search <q>`. Documents are
  embedded with `//go:embed`; `docs_nav_test.go` checks links, anchors,
  manifest coverage and that the embedded copy matches disk.
- `mcp` exposes `sui_check`, `sui_model`, `sui_compile`, `sui_render` and
  `sui_docs` tools, `sui://docs/<path>` resources, and the prompts
  `build_simpleui_interface` / `review_simpleui_security`.
- `build --single` emits one self-contained HTML document: CSS, JS and runtime
  inline, no `style.css`/`app.js` requests. `--out app.html` names the file;
  `--out <dir>` writes `<dir>/app.html`; the default is `dist/app.html`.
  `--minify` is conservative (comments and indentation; no semantic rewrites)
  and requires `--single`. Desktop is already a single executable and rejects
  `--single`.
- `fmt` is a **token-level canonical formatter**: 2-space indent per block,
  one statement per line, normalized operator/punctuation spacing (both sides
  of binary operators, unary `-`/`+` tight, `a, b`, `a: int`, `a..b`), at most
  one blank line between top-level declarations, and **comments preserved**
  (standalone comments are re-emitted on their own line at the current
  indent). It only touches layout — semantics never change. Invalid sources
  are rejected. `--write` updates the file in place (default prints to
  stdout).
- `lsp` speaks JSON-RPC 2.0 over stdio (Content-Length framing): on
  `initialize` it reports `textDocumentSync: 1` + a basic completion provider;
  on `didOpen`/`didChange` it publishes **live diagnostics** (parse + Check)
  as `textDocument/publishDiagnostics`; `textDocument/completion` returns
  keywords, built-ins and (via an AST analysis) the program's own names;
  `shutdown`/`exit` close cleanly. This is what gives SUI "errors while you
  type" in any LSP-capable editor.
- `gui` is the **visual design tool** with a dockable IDE+designer layout:
  top bar (responsive breakpoints, export, status), left panel (Structure
  tree + Design Tokens), center workspace (Preview | SUI | CSS tabs), right
  inspector (visual styler), collapsible bottom panel (Problems | History).
  The **live preview fills the remaining space** and is a sandboxed iframe.
  - **structured CSS engine** (`cssmodel.go`): CSS is treated as an AST, not
    text — a tolerant parser (text → AST, with per-rule error recovery and
    diagnostics), a pretty-printing generator (AST → text), a specificity
    calculator, and an analyzer reporting **unused classes, overridden
    declarations (with the winning selector + why) and duplicate selectors**.
    Endpoints: `/__css__/parse` (format + AST + tokens + problems) and
    `/__css__/set` (edit one declaration on a class, creating the rule if
    needed) — the single source of truth is the CSS text; visual edits go
    through `/__css__/set`, text edits go through `/__css__/parse`, so code
    and UI never drift (programmatic `.value` writes don't fire `oninput`,
    preventing sync loops);
  - **two-way binding + visual styler**: selecting an element (tree or by
    clicking it in the preview) shows state buttons (base/:hover/:focus/
    :active), a specificity badge and accordion controls (layout/flex, box
    model, typography, background, size) that read/write the AST. The
    inspector is **dynamic per capability**: a field/section exists only for
    element kinds that can actually carry it — textual kinds get the Text
    field, media kinds get `src` (never a text field), only `a` gets `link`,
    value bindings follow the per-kind binding table, and only kinds with an
    explicit event table get handler fields (rules/spacers/fire-free elements
    show no Events section); `Layout`/`Grid` accordions appear only for
    containers and lexical elements, `Typography` only for kinds that paint or
    inherit text, and `Alignment` hides itself for the page body, table
    structure and pure layout elements (hr/br/spacer/divider);
    **multi-selection intersects**: with 2+ elements the inspector shows only
    the fields/events/bindings EVERY selected element supports, reading their
    shared value (an empty field means the values differ) and applying edits to
    all of them. Element
    attribute fields (text/onClick/value/src/…) are shown **only when the
    selected tag supports them** (e.g. `textarea` gets text+placeholder
    +value, `img` gets only src, a `div` only text, a `button` text+onClick).
    Controls are fully customizable: an element styled without a class gets an
    auto-generated one that is **unique per element** and never collides with
    the built-in skin names (`sui-*` are reserved: the renderer prepends them
    to native controls, so reusing them would leak the rule into every
    same-kind control); `checkbox`/`switch`/`radio` carry the user class on
    their wrapper `<label>` **and** the inner control (see §2.5), so the rule
    reaches the whole widget; base-skin pseudo-state rules (`:checked`,
    `:hover`, `:focus`, `:disabled` on the bare tags) are emitted at
    `:where()`-reduced specificity so a user class always wins;
  - **cross-browser separation**: friendly controls (e.g. an "Opacity"
    0–100 slider, a "Roundness" slider) map to real CSS via a compat table
    that emits vendor-prefixed rules where needed (`border-radius` →
    `-webkit-`/`-moz-`/plain, opacity, transform, transition, user-select …);
    the endpoint `/__css__/setbatch` applies several declarations at once;
  - **design tokens**: `--name` custom properties are listed as swatches;
    **problems** panel: unused classes are also outlined in the tree, and
    clicking a problem jumps to the CSS line; **history** panel for
    undo/redo time-travel;
  - **selection-driven properties**: the element-attribute fields appear only
    while an element is selected — in the tree **or by clicking the widget
    directly inside the preview** (root-view elements carry a unique
    DFS-order `data-sui-id`; a picker script forwards clicks to the editor);
  - a **live preview**: the browser POSTs the generated `.sui` source to
    `/__gui__/render`, the compiler parses + checks + emits the app
    in-process, and the page renders it in a sandboxed iframe;
  - **export**: the tool generates a compilable `.sui` (escaping strings and
    auto-adding `fn` stubs for `onClick` handlers), a web build (`.zip`), or a
    single self-contained `app.html` with CSS/JS/runtime/fonts inline;
  - **desktop build dialog**: `File → Build desktop executable…` preflights Go
    1.24.2+, source, output directory and capabilities before invoking
    `RunDesktopBuild`; shows the build log and the server-local executable path,
    with an optional authenticated download of the produced binary;
  - **recursive element tree**: `row`/`col` containers nest arbitrarily deep,
    rendered as an indented tree in the left panel;
  - **drag & drop**: any node can be dragged onto another to reparent it
    (including into/out of containers) or reordered in place;
  - **undo/redo**: every mutation is pushed on a history stack (Ctrl+Z /
    Ctrl+Shift+Z);
  - **pages**: a project owns one or more **pages** (SUI views); a new project
    starts with a single `main` page. There is no dedicated pages panel: pages
    appear as **tabs inside the preview pane** (click activates, double-click
    renames), `New page…` lives in the **Insert menu** (and the icon dock) —
    organized as
    **cascading submenus** (top level: Pagina / Layout / Text / Controls /
    Tables / Media; hover a group to open its flyout, nested sub-groups such
    as Navigation / Lists / Feedback open one level deeper) — and
    opens the new-page dialog offering an **empty page** or a **page template**
    — built-ins (dashboard, form, CLI wrapper) plus user-saved page templates
    from the shared library. With a page tab active the **inspector** shows the
    page attributes (title, page name) plus **Duplicate / Delete / Save as
    template** actions. Every page is emitted as its own `view` and validated
    against the project environment; the active page drives the tree,
    inspector and preview (`__sui_page(name)` switches the visible page
    box in the generated app, which keeps one box per root view under `#app`
    with only the active one visible);
  - **blocks**: the bottom-left library section (`Save selection as block`)
    stores one or more selected elements as a reusable block; blocks persist
    in `%APPDATA%\simpleui\templates.json` (shared across projects, beside
    page templates, tagged with `kind: block|page`) and are inserted via
    `insertTemplate`, validated against the **live project environment**
    (`/__gui__/template/open`) so snippets binding project states/functions
    insert correctly; failures open a styled dialog with the diagnostics;
  - **editable view model**: the view `title`, global `state`s and helper
    `fn`s are first-class fields of the model (add/remove/rename, initial
    values, parameter lists and return types, re-parsed when opening a
    `.sui`); elements that bind `value:`/`onClick:` to a name auto-declare the
    matching `state`/`fn`;
  - **open a `.sui`**: the file is POSTed to `/__gui__/open`, the compiler
    parses + checks it in-process, and the first view's element tree plus the
    top-level `state`/`fn`/`title` are loaded into the designer (interpolated
    text keeps its literal parts; the export re-emits `state`s above the view,
    `fn`s below, matching the canonical file layout).

  The design tool visualises every non-invoked view as a page and invoked
  views as raw components; `if`/`loop`/`list` blocks are
  not visualised in the tree.

- Flags are parsed manually (the `flag` package stops at the first positional):
  both `--name value` and `--name=value` are accepted.
- `parse`: prints the AST JSON to stdout; exit 0. On diagnostics: print to
  stderr, exit 1.
- `build`: default target `web`, default out `./dist`. Creates the output
  directory if missing, writes the three assets, prints a summary. Missing
  input file: error + exit 1.
- `--target desktop` builds a single executable that serves the app on
  localhost and opens the system browser. `--target desktop-webview` builds the
  same executable with an embedded native WebView window (WebView2 on Windows,
  WKWebView on macOS, WebKitGTK on Linux). Both accept `--allow-exec` and
  `--allow-dir`, and support `--no-open` / `--port N` at runtime.
- Every unexpected panic is caught and reported as `internal error` (exit 2),
  never a stack trace.
- `serve`: builds the web assets into a temp dir, serves them same-origin
  (127.0.0.1) and opens the browser. With `--watch` it polls the source
  (entry + sibling `.sui` files), recompiles on change and pushes an SSE
  `reload` event; the served `index.html` gets an injected
  `/__sui__/reload.js` client, so the browser reloads automatically. Port
  defaults to a random free one (`--port` to fix it).

### Diagnostics contract (what the user sees)

All phases funnel errors into `DiagList` (errors.go). Each diagnostic carries a
position (`file:line:col`, 1-based), a category (`lexical`/`syntax`/`type`/`io`/
`codegen`), a short message, and an optional one-line hint explaining **what to
do**. The CLI renders:

```
type error: cannot assign a value of type string to "counter" of type int
   --> examples/bad.sui:4:2
    |
  4 |     counter = "test"
    |     ^
    = the state was declared here: examples/bad.sui:1
1 error. fix the problem above and recompile.
```

- `RegisterSource(file, content)` is called by `ParseSource`, so every
  diagnostic printed to a registered file shows the offending source line and a
  caret; otherwise a classic `file:line:col: message` line is used.
- Categories make multi-phase cascades readable (a lexer failure produces a
  `lexical error` and the parser then reports `invalid character (...)`).
- High-value hints: "did you mean ...?" (edit-distance ≤ 2) for unknown
  variables/functions, the state's declaration site on type mismatches,
  guidance for std calls outside `async`, reserved-name explanations, file-read
  failures.
- `DiagList.String()` renders the full report without touching a stream
  (tests/embeddings); `Print()` writes it to stderr.

---

## 7. Go idioms the team must use (verified on this toolchain)

- Package file header: `package main`.
- Struct: `type Foo struct { a int, b string }`; init `Foo{a: 1, b: "x"}`;
  pointer `*Foo` with `&x`; slices `[]Foo` with `x = append(x, v)`, `len(x)`,
  `x[i]`.
- No `enum`/`union` keywords: tag with a `kind string`.
- Strings: `len(s)`, `s[i]` (byte), `s[:i]`, `s[i+1:]`, `strings.Split`,
  `strings.Join`, `strings.HasPrefix/HasSuffix/IndexAny/Trim`. Build with
  `var sb strings.Builder; sb.WriteString(".."); s:= sb.String()`.
- Files: `data, err := os.ReadFile(path)` -> `([]byte, error)`;
  `os.WriteFile(path, data, os.Perm(0o644))`. Test `err != nil`, print with
  `%v` if needed. `os.Exit(code)`.
- Import order + only what is used (`go vet` FAILS on unused imports).
- Cross-file functions: Capitalized. Within-file helpers: lower-case.
- Tests: files `*_test.go`, `package main`, `import "testing"`,
  `func TestXxx(t *testing.T) { if ... { t.Errorf("...") } }`. Run `go test`.
- Iterate only your own files; never edit another agent's file.

---

## 8. Roadmap status

| Phase (Design Document) | status |
|---|---|
| 1 lexer + parser               | this iteration |
| 2 codegen HTML/CSS             | this iteration |
| 3 codegen JS + reactivity      | this iteration |
| 4 desktop wrapper + bridge     | this iteration (browser desktop and embedded WebView targets) |
| 5 modules + stdlib             | minimal stdlib + import merging this iteration |
| 6 examples + E2E               | examples/ + go tests this iteration (web + desktop) |

### 9. Desktop target contract

`simpleui build app.sui --target desktop|desktop-webview [--allow-exec <cmd>] [--allow-dir <dir>]`
writes a disposable module (go.mod `desktopapp`) into `--out`:

```
out/
├── go.mod          module desktopapp
├── main.go         runtime source (from runtime.go template)
└── assets/         index.html, style.css, app.js (desktop variant)
```

Then it runs `go build` with the local Go toolchain, producing a single
self-contained executable `app` (`app.exe` on Windows).

Runtime behaviour of the compiled executable:
- Embeds the three assets with `go:embed`; listens on `127.0.0.1` only on an
  ephemeral port (or `--port N`). Prints its URL to stdout.
- Generates a per-run 64-char hex secret and injects it into the served
  `app.js` wherever `__SUI_TOKEN__` appears; bridge calls must echo it back.
- Serves `GET /`, `/style.css`, `/app.js` same-origin.
- Exposes POST `__bridge__` endpoints (form-encoded):
  - `POST /__bridge__/exec`  fields `token, cmd, arg...` — runs `cmd` only if
    it is in the compiled `--allow-exec` list (exact match), via `os/exec`,
    returning `{"ok":bool,"data":"output","message":"err"}`.
  - `POST /__bridge__/file/read` `token, path` - reads a file only when its
    symlink-resolved path stays inside one of the compiled `--allow-dir` bases.
  - `POST /__bridge__/file/write` `token, path, data` - same path rule.
  - `POST /__bridge__/fs/stat`, `/fs/list`, `/fs/mkdir`, `/fs/remove` - same path rule.
  - `POST /__bridge__/pick-file` `token, multiple` - opens a native file dialog and returns selected absolute paths.
  - `POST /__bridge__/pick-directory` `token` - opens a native folder dialog and returns its absolute path.
- The bridge token is random per launch and compared in constant time; the
  server rejects non-loopback `Host` headers, cross-site POSTs and cross-site
  loads of `app.js` (which carries the token), and sends `X-Frame-Options:
  DENY`.
- Capability grants are **baked into the binary at build time** (no runtime
  prompt): a build without `--allow-exec` disables exec entirely; without
  `--allow-dir` disables fs entirely (capability-based security, DESIGN §5.3).
- `desktop` opens the system default browser via `open`/`xdg-open`/`cmd start`;
  `desktop-webview` opens an embedded native WebView window (WebView2,
  WKWebView, WebKitGTK) with the same server and bridge. `--no-open` runs
  headless (used by tests). The server and capability model are identical for
  both targets; only the window host changes.
- The *compiler build step* requires a Go toolchain on the build machine, but
  the produced executable has zero runtimes/dependencies for end users.

JS mapping (`GenerateJsFor(p, "desktop")`): `fs.*`, `exec.run` and
`dialog.pickFile/pickFiles/pickDirectory` become `fetch("/__bridge__/…")` posts
carrying `__SUI_TOKEN`; `http.*`, `ws.*` and `json.*` keep their web
implementations. On the `web` target `exec.run` and `dialog.pick*` stay disabled
and `fs.*` keeps targeting the user backend (`/api/fs/...`).

Open follow-ups: signed/notarized desktop builds, optional WebView2 runtime
bootstrap and packaging/installers; a bridge manifest file (`simpleui.json`)
replacing flag-only grants; command-argument validation beyond the name
allow-list; platform-specific picker filters.

---

## 9. Security model (untrusted data)

The SUI source is trusted; every **value** it renders from an external source
is not. The compiler and runtime enforce:

1. **Automatic HTML escaping** of view text and attribute values (static text
   included). Interspersed expressions become `${__sui_esc(expr)}`; there is
   no raw-HTML interpolation.
2. **URL scheme filtering** for `href`/`src`/`link`: relative URLs and
   `http(s)`, `mailto`, `tel` are allowed; `javascript:`, `vbscript:`, `data:`
   and other active schemes become `#`. A *static* active scheme is also a
   compile-time error. Destination allow-lists remain the app's responsibility
   (SSRF/open redirect).
3. **Dynamic CSS requires `cssSafe(...)`**: the compiler rejects a non-static
   `style:` that is not a `cssSafe(...)` call (including interpolated style
   templates). `cssSafe` drops `url()`, `image-set()`, `paint()`, `attr()`,
   `var()`, `expression()`, CSS escapes, positioning/overlay properties and
   vendor/custom properties. Static developer-authored styles are trusted.
4. **Dynamic classes are allow-listed**: the compiler collects classes from the
   app's CSS and from static class tokens/literals in class expressions
   (`"btn {state}"`, `cond ? "on" : "off"`). `class:` expressions are emitted as
   `__sui_class(expr)`, which drops tokens not in that set. Static classes are
   unchanged.
5. **SSRF guard for dynamic URLs**: `http.get/getStatus/post/postFiles` and `ws.connect`
   accept only literal URLs; a dynamic URL requires `getSafe/getStatusSafe/
   postSafe/postFilesSafe/connectSafe` with a **literal allow-list**. The runtime matches
   scheme/host/port (or a same-origin path prefix), rejects credentials and
   active schemes, refuses redirects, and applies a 30 s timeout / 10 MB cap.
   `http.get`-style literals remain for fixed endpoints.
6. **Table data is inert**: `table.display`/`importText` write cells with
   `textContent`; JSON/CSV/TSV/JSONL and converted XLSX/ODS can never inject
   markup. Limits: 100k rows, 10k columns, 20 MB of text.
7. **CSV export** quotes fields (RFC 4180) and offers `;formula=escape` to
   neutralise spreadsheet formulas (`= + - @ TAB CR`) when the CSV will be
   opened by Excel/LibreOffice.
8. **No dynamic evaluation**: generated apps contain no `eval`/`new Function`;
   event handlers must be declared functions; `json.parse` uses `JSON.parse`.
9. **Desktop capabilities**: random token, constant-time comparison,
   loopback-only Host, same-site checks, `X-Frame-Options: DENY` plus
   `frame-ancestors 'none'` (also on `serve`), exact command allow-list
   (`exec.Command`, no shell), allow-dirs resolved through symlinks.
10. **Spreadsheet conversion** (`examples/spreadsheet_table`): no disk
    extraction, 64 MiB per XML part, row/column/cell limits, no external XML
    entities, localhost-only HTTP service, file-name sanitisation.
11. **Iframes** (`src`, `srcAllow`, `sandbox`, `allow`, `referrerpolicy`,
    `loading`, `width`, `height`): compile-time validation rejects malformed
    data; a missing sandbox becomes `sandbox=""` and a missing referrer policy
    becomes `no-referrer`. Dynamic `src` requires a literal `srcAllow` by
    default and is validated at runtime (http(s), no credentials/control
    characters, origin/path allow-list); rejected data is not loaded and shows
    a `Cannot accept the data: ...` alert. `allow-scripts` +
    `allow-same-origin` is an error by default; `srcdoc`/fallback markup is
    rejected.

Residual risks (known-class visual spoofing, conservative CSS sanitizer, DNS
rebinding / compromised allow-listed hosts, ReDoS, iframe origins, local
processes, CSV formula injection when `formula=escape` is not used) and safe
patterns are documented in `SECURITY.md`; runnable examples live in
`examples/security/`.

**Security policy file.** The compile-time policies (dynamic style/class,
dynamic URLs, static active schemes), table/CSV defaults and iframe sandbox/
allow-list defaults are configurable with `simpleui.security.json` next to the
entry `.sui` or `--security <file>`. Defaults are strict; a malformed file, an
unknown field or an invalid value is a build error. HTTP escaping, the runtime
URL-scheme filter, iframe runtime URL validation, table `textContent`, bridge
protections and CSV quoting are not configurable. See
`SECURITY.md#security-configuration`.


