# GUI editor: drag and drop

This guide explains the two independent drag-and-drop systems in SimpleUI:

1. **Designer drag and drop** - moving elements, palette items and toolbar
   sections inside the visual editor.
2. **Application drag and drop** - dragging elements, files or folders inside
   the generated application.

They share browser primitives but have different rules. Enabling one does not
enable the other.

## Quick answer

| Goal | What to set |
|---|---|
| Move components in the designer | Nothing; designer DnD is always enabled |
| Make an app element draggable | `draggable: true` and/or `dragKey` |
| Make an app element a drop target | `dropKey` plus `onDrop` |
| Reorder data after a drop | Read `drag.source()`/`drag.target()` and call `moveItem` |
| Drop files/folders onto a field | `dropFiles: state`, plus `value: state` to display paths |
| Open a file/folder dialog | `pickFiles: state`, with `multiple`, `accept`, `directory`, `pickMode` |
| Native absolute paths | Desktop build plus the separate `dialog` std module |
| Build a menu bar | Palette **Layout → Navigation → Menu bar**; edit each menu/menu item |
| Build tabbed content | Palette **Layout → Navigation → Tabs**; right-click rows to add/remove tabs |
| Build a modal dialog | Palette **Overlay → Modal**; add/remove modal buttons from the tree |

In the right-hand inspector the **Drag & drop** section uses these labels:
`draggable`, `dragKey` (drag key expression), `dropKey` (drop key expression),
`dropFiles` (drop files into state), `pickFiles` (choose files into state) and
`pickMode` (picker API mode). The `onDrop` handler is selected in **Events**.

## Editor preferences

Open **Preferences → Editor preferences.** The panel is divided into tabs so the
editor shell, the code editor, shortcuts and project files stay separate. Tabs
with many controls use a two-column grid (label and control in each column),
numeric settings use a slider plus a value box, and grouped tabs use accordions
that show one section at a time.

### Interface

Appearance and layout of the editor shell.

| Setting | Default | Effect |
|---|---|---|
| Theme | Dark | Dark, Light or High contrast editor chrome |
| Interface font | System UI | Font family for menus, panels, labels and controls |
| Custom interface font | - | CSS family used when **Custom** is selected |
| Font size | 14 | Scales the whole editor UI |
| Density | Comfortable | Comfortable or compact spacing |
| Accent color | #4c5cf7 | Focus rings, selection and primary actions |
| Panels | all visible except history | Show or hide Structure/Tokens, Inspector and Problems/History |
| Reset toolbar order | - | Restores the default toolbar layout |

### Code editor

Everything about writing SUI: typography, editing behaviour, syntax colours and
editor tooltips. The settings are grouped in accordions and a single preview at
the bottom of the tab reflects all of them. The **Interface font** in the
Interface tab and the **Code font** here are independent: changing one does not
affect the other.

#### Font & metrics

| Setting | Default | Effect |
|---|---|---|
| Code font | System monospace | Combo box with bundled stacks or a custom CSS family for the SUI code editor only |
| Custom code font | - | CSS family used when **Custom** is selected |
| Editor size | 16px | Code size, independent of the general UI zoom |
| Line height | 1.6 | Vertical spacing and line-number alignment |
| Tab size | 4 | Tab width in the overlay and textarea |
| Word wrap | off | Wrap long lines in both layers |
| Line numbers | on | Show the gutter |

#### Editing

| Setting | Default | Effect |
|---|---|---|
| Auto-close pairs | on | Insert the closing bracket or quote while typing |
| Smart indent | on | Keep indentation on Enter and indent after an opening bracket |
| Auto suggestions | on | Open the completion popup while typing; Ctrl+Space always works |

#### Syntax highlighting

| Setting | Default | Effect |
|---|---|---|
| Syntax colors | on | Colour SUI keywords, strings, numbers and comments |
| Bracket match | on | Highlight the matching bracket |
| Selected line | off | Optional background for the selected element line |

With **Syntax colors** enabled, declared functions and their calls are shown in
gold, while states, function parameters, `let` bindings and loop variables are
shown in cyan. The same colours are used in declarations, references and
interpolations such as `"Count {count}"`, so a name is easy to find and reuse.
Only the colour changes, so the overlay and textarea keep identical metrics.

The editor uses the same font, padding, line height and character metrics in the
syntax-highlight overlay and the transparent textarea, so the caret and the
visible characters stay aligned. Syntax tokens only change colour; bold/italic
styles are not applied to code tokens.

#### Editor tooltips

Tooltips that describe buttons, fields and toolbar actions of the editor. They
are part of the editor chrome and never affect the generated app.

| Setting | Default | Effect |
|---|---|---|
| Show tooltips | on | Enable the custom editor tooltip bubble |
| Delay | 450 ms | Hover delay before the bubble appears |
| Place | Auto | Top, right, bottom, left or automatic placement |
| Max width | 320px | Maximum bubble width |
| Font size | 12px | Bubble text size |
| Background / text / border | dark palette | Bubble colours |
| Radius | 6px | Bubble border radius |

#### Preview

The preview at the bottom of the tab renders sample SUI code with the current
font, size, line height, tab size, wrapping, syntax colours and selected line.
It also shows a tooltip bubble using the current editor tooltip style. It
updates while you edit the controls, before **Apply**.
### Shortcuts

The **Shortcuts** tab is a read-only reference. Use the search field to filter
by action, key or group. It lists the designer, SUI editor, menu, structure and
chart data shortcuts; shortcuts are not rebindable yet.

### Files & history

Snapshots and file locations are grouped in accordions.

| Setting | Default | Effect |
|---|---|---|
| Automatic snapshots | on | Store a project version while the designer saves |
| Interval | 60 s | Minimum time between automatic snapshots, with slider and value box |
| Versions kept | 30 | Automatic retention limit, with slider and value box |
| Snapshot now | - | Force a version immediately |
| Open project history | - | Open the restore panel |
| Clear history | - | Delete every saved version after confirmation |

The **Files** section shows the data folder, the preferences file and the
project file, with a copy button for each path. Automatic snapshot settings are
stored in the editor preferences; the version files live under the data folder.

Preferences are saved with the other designer preferences and persist across
reloads. The general **Font size** setting scales the UI, while **Editor size**
controls the code independently.
## Completion and indentation

Press **Ctrl+Space** in the SUI editor to open the completion list at the caret.
Suggestions change with the surrounding code:

| Position | Suggested items |
|---|---|
| Top level | `import`, `css`, `state`, `type`, `struct`, `fn`, `async`, `view` |
| Inside a view or element | `title`, `onLoad`, element tags and component views |
| Inside a function or block | `let`, `if`, `else`, `loop`, `while`, `return`, states, functions and local variables |
| After an element name | attributes and event names for that element |
| After an event attribute (`onClick:`) | declared functions and **+ new function…** |
| In an expression or interpolation | states, functions, local variables, local loop items, builtins and std modules |
| After `class:` | CSS classes declared in the source |
| After `state name:` or `->` | SUI types |

Typing filters the list by prefix. Arrow keys move the selection, **Enter** or
**Tab** inserts it, and **Escape** closes it. The automatic popup uses the same
context and appears after the first letters; Ctrl+Space also works with an empty
word.

When the designer regenerates source, function bodies are indented one level
inside their braces. Nested statements keep their relative indentation, so
round-tripping a project through the visual editor does not flatten function
code.

## Code editor menus and navigation

### Context menu

Right-click in the SUI editor to open grouped actions:

- **Undo / Redo**
- **Cut / Copy / Paste**
- **Select all / Select line**
- **Duplicate line, Move line up/down, Delete line, Indent, Outdent**
- **Go to definition, Find references, Rename symbol**
- **Find, Replace, Go to line, Command palette**
- **Toggle comment**

The menu is contextual: Cut and Copy are enabled only with a selection, while
navigation entries require a symbol under the cursor. Items show their shortcut,
can be reached with the arrow keys and run with Enter; Escape closes the menu.

### Line operations

| Action | Shortcut |
|---|---|
| Select line | Ctrl+L |
| Duplicate line or selection | Shift+Alt+Down |
| Move line or selection up / down | Alt+Up / Alt+Down |
| Delete line or selection | Ctrl+Shift+K |
| Indent / outdent | Tab / Shift+Tab |

### Find and replace

Press **Ctrl+F** to open the find bar and **Ctrl+H** to show the replace row.

| Action | Shortcut |
|---|---|
| Next / previous match | Enter / Shift+Enter, F3 / Shift+F3 |
| Match case | **Aa** button |
| Whole word | **W** button |
| Regular expression | **Regex** button (`.*`) |
| Find in selection | **Sel** button |
| Replace current match | **Replace** |
| Replace every match | **All** |
| Close the bar | Escape |

Matches are highlighted in the editor with the current match in a stronger
colour, and the match counter shows the current position and total. Find starts
from the current selection when it is a single-line selection.

The search options are remembered between sessions. The find field keeps a
history of recent queries. With replace open, a preview lists the first
replacements with their line numbers before **Replace** or **All** is used. In
regular-expression mode the replacement supports capture groups such as `$1`.

### Quick Fix and problems

| Action | Shortcut |
|---|---|
| Quick Fix at the cursor | Ctrl+. |
| Next problem | F8 |
| Previous problem | Shift+F8 |

Quick Fix proposes actions for the diagnostic under the cursor, including
declaring a missing state, defining a missing handler/function and making the
enclosing function `async`. Diagnostics remain clickable in the Problems panel.

### Status bar

The bottom of the code editor shows the caret line and column, the selection
size, the number of errors and warnings, the tab size and whether the source is
**Applied** or **Modified**. The problem count opens the Problems panel; the
arrow buttons move to the previous or next problem.

### Command palette

Press **Ctrl+Shift+P** to search editor commands and declared symbols.

- Type to filter by command name, shortcut, state, function or view.
- Arrow keys move the selection; Enter runs it; Escape closes.
- Symbol entries jump directly to their declaration line.
- Commands include Find, Replace, Go to line, Quick Fix, problem navigation,
  line operations, comment toggle and Apply SUI source.
### Command palette

Press **Ctrl+Shift+P** to search editor commands and declared symbols.

- Type to filter by command name, shortcut, state, function or view.
- Arrow keys move the selection; Enter runs it; Escape closes.
- Symbol entries jump directly to their declaration line.
- Commands include Find, Replace, Go to line, navigation, line operations,
  comment toggle and Apply SUI source.

## Undo and redo

The toolbar and **Edit → Undo / Redo** use one history for the whole designer,
not just the element tree. The same history is used by **Ctrl+Z** and **Ctrl+U**
for undo and by **Ctrl+Y** / **Ctrl+Shift+Z** for redo.

Recorded actions include:

- SUI text edits (a typing burst is one step until the automatic apply runs);
- element text, classes and reactive expressions;
- property changes in the inspector, including CSS rules and table options;
- inserting, deleting, duplicating, moving and pasting components;
- page changes, menu/tabs/modal structure and drag-and-drop edits;
- CSS/media-rule edits and source changes applied from the SUI editor.

Undo restores both the model and the SUI source, then updates the tree, preview
and inspector. The history panel lists the recorded steps and can jump directly
to an earlier state. When focus is in an inspector field, Ctrl+Z/Ctrl+U follow
the designer history; prompt and modal text fields keep their native undo.

## Left panel and design tokens

### Resizable panels

Drag the right edge of the left panel to resize it, exactly like the inspector.
The divider has a widened 12 px hit area and a visible line, so it is easier to
grab. The width stays between 190 and 600 px and is remembered across reloads.
Dragging uses pointer capture, so the resize remains continuous even when the
pointer crosses the preview iframe.

The **Problems/History** panel at the bottom is resized by dragging its top
edge. Its height stays between 90 px and 65% of the window, is remembered across
reloads, and a double-click on the divider restores the default 190 px.

### Design tokens

The **Tokens** tab lists the design tokens found in the project CSS. A design
token is a CSS custom property such as:

```css
:root {
  --primary: #4c5cf7;
  --radius: 10px;
  --page-gap: 16px;
}
```

They are **not** the syntax tokens used to colour SUI code. Design tokens are
reusable values for colours, spacing, radii, fonts and similar design choices.
Use them with `var(--primary)` in class rules, `style: cssSafe(...)` or theme
overrides. Changing a token updates every rule that references it.

The tab groups tokens into **Colors**, **Typography**, **Spacing**, **Radius**
and **Other**, with a search field. Each row is editable inline:

- change the name to rename the token; the declaration and every
  `var(--old)` reference in the CSS are updated;
- change the value directly, or use the colour picker for colour tokens;
- **C** copies `var(--name)`;
- **U** applies `var(--name)` to a CSS property of the selected element;
- **D** duplicates the token;
- **x** deletes it (used tokens are replaced by their value first).

The inline form at the top adds a new custom property to `:root` without browser
prompts. In the inspector, the small **T** button next to colour and length
fields saves the current value as a token and replaces the field with
`var(--name)` in one undoable step. Use tokens for values shared across the
interface, and plain values for one-off declarations.

## Project component defaults

Open **Preferences → Project preferences.** to set defaults for every instance
of a component in the project. The panel has a **Theme** combo above the
category tabs. Tabs are **All**, **Feedback**, **Layout**, **Form**,
**Navigation**, **Data**, **Overlay**, **Charts** and **General**. Inside each
tab the component buttons are arranged as a compact table grid, sized to their
text. The Theme combo applies a shared style to every control at once, including
scrollbars, instead of setting each component by hand. Choosing **Custom** leaves
the current values untouched.

Each component opens the same configuration panel used by its element settings.
The panel follows the editor preferences layout:

- categories in tabs when there are several groups;
- two-column label/control layout;
- sliders with a numeric box for numeric values;
- switches for on/off options;
- combo boxes for fonts and choices;
- a close icon in the header and Reset/Cancel/Apply actions.

Project values are stored as CSS custom properties in a managed
`/* sui-component-defaults */` block and consumed by a per-component rule block.
Applying a component from project preferences writes both blocks and refreshes
the preview. **Tooltips** in the same panel opens the existing global tooltip
settings; closing it returns to the project list.

### Project themes

Every shared style has a **Light** and a **Dark** variant. The Theme combo lists
them grouped by family, for example Sketch Light/Dark, Terminal / CRT
Light/Dark and Geometric Light/Dark. Applying a theme writes the complete set of
component defaults for every control, updates the managed CSS rule blocks,
styles the scrollbars and sets the **page starting point** (`--sui-bg`,
`--sui-fg`, base border and hover colours). A light theme therefore starts from a
light body, and a dark theme from a dark body. Base elements follow the same
variables too: the view surface, headings, links, code blocks, table headers, the
menu bar and plain buttons switch with the theme. The **Standard** family provides
**Standard Light** and **Standard Dark**, which keep the default shape of every
control. Editing one component afterwards changes the Theme combo to **Custom**,
so it is always clear whether the project still follows a single theme. A new
project starts on **Standard Dark**; choosing **Standard Light** switches both
the page and the controls to light defaults.

### Scrollbars

Scrollbars are not inserted like normal elements: they appear whenever a box
scrolls. Their appearance is configured from **Project preferences → General →
Scrollbars**. The panel controls the scrollbar size, thumb radius, Firefox width,
track, thumb, thumb hover and border. The shared themes set the scrollbar palette
consistently with the other components, and the panel preview contains a
scrollable sample.

### Chart and graph project defaults

**Project preferences → Charts** contains **Chart defaults** and **Graph
defaults**. They open the same unified chart/graph settings panel used by the
inspector, with project scope. Applying writes the appearance and interaction
settings to every chart or graph in the project and to newly inserted ones:
theme, style preset, point size, tooltip, table, Gantt, links and interaction
options. Chart data, type and X/Y bindings stay per element, and the Data tab is
hidden in this mode. The defaults are stored in a managed
`/* sui-chart-defaults ... */` comment in the project CSS.

### Element overrides

Select an element and press **Settings...** in the inspector banner. The same
panel opens with a **This element only** scope. Values are written as inline CSS
custom properties on that element, so every other instance keeps the project
defaults. **Clear overrides** removes those inline values and restores the
project defaults for that element. Closing the panel returns to the inspector.

Implemented component panels: **Progress bar**, **Badge**, **Alert**, **Spinner**,
**Card**, **Divider**, **Button**, **Input / Textarea / Select**, **Combobox**,
**Date / Time picker**, **Checkbox / Radio**, **Switch**, **Slider**, **Tabs**,
**Accordion**, **Menu**, **Tree**, **Table**, **Toast**, **Modal** and
**Scrollbars** (global). Chart, Graph and media controls keep their specialised
inspectors for now.

Each panel shows a live preview above the fields. The preview uses the current
draft values, not the saved ones, so colours, sizes and states can be checked
before Apply.

Text-bearing controls expose a **Letter spacing** field. Interactive controls
have a **Hover** tab with the colours used while the pointer is over them:
background, text, border, option, row or thumb depending on the component. This
makes hover feedback part of the same theme instead of an extra CSS edit.

### Style presets and animations

Every component panel starts with a **Style** selector. **Custom** is the raw
default, and each named preset applies a coordinated set of colours, shapes and
effects in one step. Presets are optional: they never run unless selected, and
editing any field switches the selector to **Custom** so fine-tuning stays
explicit. Every shared family exists in **Light** and **Dark** versions, for 22
shared variants in total. The Style selector lists them under **Shared styles**;
the project Theme combo selects the same variants for all components at once.

Above the preview, **Preview background** switches between **Dark** and
**Light**. The choice is remembered and makes it easy to check a component
against both a dark application shell and a light or paper-like page. Selecting a
shared style also chooses a recommended background automatically (light for
Sketch and Brutalist, dark for the others); the buttons still let you inspect
the other one.

The presets cover all twenty component panels, for example **Neon glow** and
**Barber pole** for Progress, **Dual ring**, **Gradient ring** and **Coin flip**
for Spinner, **Glass** for Card and Toast, **Pill** for Button and Input, and
**Striped**, **Borderless** and **Rounded** for Table. Several presets include a
motion effect; the Animation/Effects tab also exposes the effects directly:

- Progress: pulse, scrolling stripes, shimmer and glow.
- Spinner: classic spin, pulse and vertical flip.
- Badge: pulse, glow and float.
- Alert: fade in, slide in, shake and pulse.
- Card: fade in and float.
- Button: pulse, float and glow.
- Toast: slide down, fade, zoom, slide up and bounce.

The **Sketch** family has Light and Dark versions, shared by all twenty panels so
components placed on the same page stay visually related. The light variant uses
a warm paper surface; the dark variant uses cream ink on dark paper. Both use
hand-drawn borders, offset shadows, a handwriting font and a highlighter
selection where applicable. Every colour, font, radius and shadow remains
editable, so the family can be tuned without losing its hand-drawn character.

The **Wireframe** family also has Light and Dark versions. It removes background
fills and draws each control only with thin lines, dashed borders and soft
shadows. Accent states use a cyan line; colours, font, radius and shadow fields
remain editable, so a page can mix several Wireframe components without losing
the same outline language.

The Style selector groups these families under **Shared styles** and keeps
component-specific looks under **Component styles**. The additional shared
families, each with Light and Dark variants, are:

- **Terminal / CRT**: green phosphor text on black, monospace labels and glow.
- **Glassmorphism**: smoked translucent surfaces with blur and light edges.
- **Brutalist**: thick black outlines, flat yellow blocks and hard offset shadows.
- **Blueprint**: deep blue paper, cyan dashed technical lines and monospace type.
- **Neumorphic**: soft monochrome surfaces with paired light and dark shadows.
- **High contrast**: black, white, yellow and cyan focus for accessibility.
- **Editorial / Luxury**: deep navy, cream type, gold rules and serif type.
- **Cyberpunk**: black chrome, magenta and cyan neon lines with hard glow.
- **Geometric**: hard squares, cut corners and triangle/checker patterns.

In element settings the selector also offers **Project default**. Applying it
removes the element's component overrides, so the element inherits the project
defaults again. Named presets, by contrast, pin their values on that element.
Project presets are written in the managed `/* sui-component-defaults */` and
`/* sui-component-rules:<id> */` blocks, while element presets use inline custom
properties.

### Insert any component

The top toolbar has an **Insert component** button. It opens a searchable list of
every component tag supported by the designer, including controls that are
nested in submenus (Progress, Badge, Alert, Spinner, Combobox, Date/Time picker,
Tree, Chart, Graph and so on). Select a component to insert it at the current
position.

## Inspector and effective styles

All inspector sections start collapsed. Open a section by clicking its header;
it stays open across selection changes and preview reloads until you close it
again.

Every row in the left **Structure** tree starts with an exact type badge
(`div`, `tree`, `treeitem`, `chart`, `input`, ...). The badge is independent of
the friendly label, so the selected element type is always unambiguous. The
inspector repeats the exact type in a **Selected element** banner above the
sections.

Style fields show the value that actually reaches the selected element, not
only a declaration from its first class rule. The preview reports authored and
computed values, so the inspector also fills fields for:

- rules on compound selectors such as `table.suitbl` or `div.card`;
- inherited values, including text colour;
- inline `style` attributes;
- CSS variables and table base rules.

Authored values are preserved when available, so `width: 100%` remains `100%`;
when a value only comes from layout, the computed value (for example `400px`)
is used. Width, height, colours, typography, spacing, borders, layout and the
other style controls follow the same rule. Editing a field still writes the
appropriate rule for the selected class (or its parent container for
container-only layout properties).

The **required** property is offered for `input`, `textarea`, `select`,
`combobox`, `radio`, `number`, `date`, `time` and `datepicker`. Native controls
use the browser's `required` attribute; the combobox adds an inline required
error when no value is selected.

## Designer drag and drop

The designer's own drag and drop is always enabled. There is no setting,
property or build flag for it. The inspector's **Drag & drop** section configures
the generated application; it does not control moving components in the editor.

| Action | How |
|---|---|
| Insert a component | Drag a palette item into the preview or tree |
| Reorder a component | Drag a tree row above/below another row |
| Reparent a component | Drop a tree row inside a container row |
| Reorder toolbar sections | Drag the toolbar grip |
| Change table row/column size | Drag a table boundary grip |

The editor uses the internal element model as the source of truth. A successful
drop creates an undo snapshot, updates the tree, re-renders the preview and
regenerates the `.sui` source.

### Tree drops

Tree rows carry their element path in `dataTransfer` as text. During
`dragover` the editor highlights the potential target. On `drop` it:

1. rejects a drop into the element's own subtree;
2. removes the source node from its old parent;
3. inserts it at the target position or into the target container;
4. records an undo step.

Rows that contain other elements show a disclosure arrow. Click it to collapse
or expand the subtree without changing the selection; the state is kept while
the project is open. Leaf rows show a blank arrow space so labels stay aligned.

If a drag is ambiguous, drop on the **tree**, not the preview. The tree has
explicit parent/child rows and is the most reliable target.

### Palette drops

The palette stores a temporary kind such as `palette:button`. The preview
captures the drop, sends the target element id to the editor, and the editor
inserts the new component:

- into a container target, as a child;
- after a leaf target, as its next sibling;
- inside the selected page body when no target is selected.

The new element is selected after insertion.

### Designer limitations

- Dropping a node into its own subtree is intentionally blocked.
- Some leaves (`button`, `input`, `img`, `text`, `br`, ...) cannot contain
  children; a drop beside them is used instead.
- The preview is an iframe. If the browser does not deliver drops inside it,
  use the tree.
- Designer drag and drop is an editing feature. It does not make the generated
  application draggable.

## Application drag and drop

Application drag and drop is configured per element. There is no global
"enable drag and drop" property. A working interaction normally needs:

- a **source** attribute;
- a **target** attribute;
- an **event handler** on the target;
- declared state/functions for the operation.

### Required properties

| Property | Where | Purpose |
|---|---|---|
| `draggable: true` | Source | Marks the element as an HTML drag source |
| `dragKey: expr` | Source | Key/payload exposed as `drag.source()` |
| `dropKey: expr` | Target | Key exposed as `drag.target()` |
| `onDrop: handler` | Target | Function called when a valid drop occurs |
| `dropFiles: state` | File target | Automatically receives dropped file/folder paths |
| `pickFiles: state` | Dialog trigger | Opens a native file/folder dialog |
| `multiple: true` | Picker | Allows more than one file |
| `accept: ".csv,.txt"` | Picker | Filters selectable files |
| `pickMode: "native" \| "fs" \| "auto"` | Picker | Selects browser input, File System Access API, or automatic fallback |
| `directory: true` | Picker | Selects a folder rather than individual files |

`dropKey`/`onDrop` are for dragging elements and data inside the page.
`dropFiles` is for files and folders dragged from the operating system and does
not require `onDrop`.

All drag-and-drop parameters are grouped in the collapsible **Drag & drop**
section at the bottom of the inspector, after the style sections. Click its
header to expand or collapse it. The matching `onDrag*` and `onDrop` handlers
remain in the **Events** section.

Event attributes supported by the application runtime:

```text
onDragStart  onDragEnter  onDragOver  onDragLeave  onDrop  onDragEnd
```

Handlers are zero-argument functions. Read the drag state through the builtins:

```text
drag.source()  drag.target()  drag.data()
drag.files()   drag.folder()  drag.error()
drag.active()  drag.clear()
```

### Minimal reorder example

```sui
state items = ["Design", "Build", "Ship"]

fn reorder() -> null {
	let src = indexOf(items, drag.source())
	let dst = indexOf(items, drag.target())
	if src < 0 or dst < 0 or src == dst {
		return
	}
	items = moveItem(items, src, dst)
}

view main {
	list item in items {
		row draggable: true dragKey: item dropKey: item onDrop: reorder {
			text "{item}"
		}
	}
}
```

`dragKey: item` alone makes the row draggable. The explicit
`draggable: true` makes the intent visible in the source and in the designer.
`draggable: true` without `dragKey` sends no useful key, so the target cannot
identify the source. The target must have both `dropKey` and `onDrop`; a key
without a handler does nothing.

### Designer steps for the reorder example

1. Select the repeated row in the tree or open its source in the code view.
   The properties are in the right-hand inspector.
2. In **Drag & drop**, set:
   - `draggable` on;
   - `dragKey` to `item`;
   - `dropKey` to `item`.
3. In **Events**, choose `onDrop` and create/select `reorder`.
4. Implement `reorder` in the function/state panel.
5. Preview: drag one row onto another.

For repeated `list` rows and table cells, the designer may not expose every
per-row attribute. In that case add the attributes directly in the `.sui` source.
The runtime emits DnD callbacks inline for repeated list rows, so every row
works, not only the first one.

## Dropping files and folders

`dropFiles` is the automatic file/folder target. It requires a string state:

```sui
state attachment = ""

view main {
	input "Attachment" value: attachment dropFiles: attachment
}
```

The state receives the dropped paths. Pair it with a `value:` binding if the
control should display them. For multiple files an input uses comma-separated
paths; a textarea uses newlines. The inspector creates the string state
automatically when you enter a new name in `dropFiles` or `pickFiles`; in the
code editor, declare it yourself.

Folder drops are recursive where the browser exposes directory entries. The
selection also populates:

- `drag.files()` - metadata (`name`, `path`, `size`, `type`, `isDirectory`);
- `drag.folder()` - the common top-level folder name, when available;
- `drag.error()` - a rejection reason;
- `drag.active()` - whether a drag is in progress.

Limits and security:

- 2,000 files, 16 directory levels and 4,096 characters per path;
- names are untrusted data and are rendered with normal escaping;
- file bytes are not read by the drop handler;
- `http.postFiles` can upload the selected bytes to a backend;
- a dropped path is not permission to read the original local path.

See [Drag and drop in the API reference](API_REFERENCE.md#drag-and-drop) and the
[file analyzer example](../examples/file_analyzer/README.md).

## Menu bar control

The palette has a ready menu-bar control under **Layout → Navigation**:

- **Menu bar** inserts a `menubar` with File, Edit and Help menus and sample
  items;
- **Menu** inserts one top-level menu with two sample items;
- **Menu item** inserts one item inside the selected menu.

### Edit it from the element list

Right-click a row in the tree. The context menu is the menu editor:

| Right-clicked row | Actions |
|---|---|
| `body` | **Add menu bar** |
| `menubar` | **Add menu**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `menu` | **Add menu item**, **Add submenu**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `menuitem` | **Add item below**, **Add submenu below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| any other element | **Duplicate**, **Move up/down**, **Delete** |

**Add submenu** inserts a nested `menu` inside the selected menu. The tree shows
it as an indented menu; its own context menu offers **Add menu item** and
**Add submenu** again, so nesting is unlimited. **Add menu** is available on the
`menubar` and creates a top-level menu with two starter items.

### Configure labels and actions

1. Insert **Menu bar**. The tree shows `menubar` with `menu` and `menuitem`
   children.
2. Select a `menu` and edit **Text** to change its visible label.
3. Select a `menuitem` and edit **Text** for its label.
4. For an action, set **onClick** in **Events** to a zero-argument function.
5. For navigation, set **link** in the element inspector. Links go through the
   normal URL guard; untrusted dynamic URLs still need validation.
6. Add/delete/reorder entries with the tree context menu. Dragging **Menu** or
   **Menu item** from the palette also works.
7. Use `divider` inside a menu for a separator.
8. Style the generated parts with the classes `.sui-menubar`, `.sui-menu`,
   `.sui-menu-button`, `.sui-menu-popup` and `.sui-menuitem`.

The `menubar` **Text** field is its accessible name (`aria-label`), not a
visible heading. The `menu` **Text** is the button label and the `menuitem`
**Text** is the item label.

### Behaviour

- Clicking a menu or submenu button opens/closes its popup; clicking outside or
  pressing Escape closes the whole menu tree.
- Arrow keys move focus through an open menu; Right/Left move into and out of a
  submenu.
- A `menuitem` can have `onClick`, `link`, or both.
- `menu` nested inside another `menu` popup is a submenu; nesting is unlimited.
- Items in a closed popup are hidden in the preview. Select them in the tree;
  click the menu button in the preview to see the rendered popup.
- Labels are escaped and link URLs are sanitized by the normal runtime.
- The generated structure is semantic: `nav[role=menubar]`, menu buttons, and
  `[role=menuitem]` entries.

See [API reference: menu bar](API_REFERENCE.md#menu-bar) and the
[menu bar example](../examples/menu_bar.sui).

## Pages

Every project has one or more pages. The **page chips above the preview** are
the page switcher: click a chip to edit that page. Only the active page is
editable/visible in the designer; this is an editor view, not an app limit.

- **Insert → New page…** creates an empty page or a page template.
- The **View** inspector section edits the active page `title` and `name`.
- **Duplicate**, **Delete** and **Save as template** live in that section.
- Each page is emitted as its own SUI `view`; the first page is the generated
  app's initial page.

The generated app hides the other pages until a handler calls the built-in
`navigate("pageName")`:

```sui
fn openSettings() -> null {
	navigate("settings")
}

button "Settings" onClick: openSettings
```

`navigate` is a no-op for single-page apps and unknown page names. Use buttons,
menu items or any normal handler to expose the navigation. See
[API reference: pages](API_REFERENCE.md#pages-and-navigation) and the
[pages example](../examples/pages/README.md).

## Modal control

The **Overlay → Modal** palette item inserts a hidden dialog with a title, body
content and two preset buttons: **Cancel** and **OK** (primary). The dialog
itself uses `modal`; each footer button uses `modalbutton`.

### Configure it

1. Select `modal`; in the **Modal** inspector section edit:
   - **name** — the stable id used by `modal.open(name)` / `modal.close(name)`;
   - **Open in preview** — check the live dialog without leaving the designer.
2. Select a `modalbutton`:
   - **Text** = label;
   - **variant** = default, primary or danger;
   - **close** = close the dialog after the click (on by default);
   - **onClick** = optional action; it runs before the close.
3. Right-click tree rows:
   - `modal` → **Add modal button**, **Edit label**, **Duplicate**, **Move up/down**, **Delete**;
   - `modalbutton` → **Add button below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete**.

Body content is any direct child that is not a `modalbutton`; drop paragraphs,
inputs, tables, etc. inside the `modal`.

### Behaviour

- The dialog is hidden until `modal.open("name")` runs.
- It closes on backdrop click, `Escape`, or a default `modalbutton`.
- Focus moves into the dialog; the opener is restored when still available.
- Opening another modal closes the current one.
- Buttons with `close: false` keep the dialog open.
- The runtime restores an open dialog after a normal view re-render.
- Labels are escaped; the panel has `role=dialog` / `aria-modal=true`.

See [API reference: modal dialogs](API_REFERENCE.md#modal-dialogs) and the
[modal example](../examples/modal.sui).

## Tabs control

The **Tabs** palette item under **Layout → Navigation** inserts a working tab
set with two tabs and sample panel content. A `tabs` node is the tab set; each
direct `tab` child is one tab:

- `tabs` **Text** = accessible name (`aria-label`), not a visible heading;
- `tab` **Text** = visible tab button label;
- `tab` children = the panel content.

### Edit it from the element list

Right-click a tree row:

| Right-clicked row | Actions |
|---|---|
| `tabs` | **Add tab**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `tab` | **Add tab below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| any other element | **Duplicate**, **Move up/down**, **Delete** |

The tree labels them explicitly as `tabs 'Settings'` and `tab 'General'`. Each
new tab gets a starter paragraph; drop or insert any component into it.

### Behaviour

- The first tab is active initially.
- Clicking a tab switches its panel.
- Left/Right/Home/End move between tabs.
- The selected tab is restored after a view re-render for page-level tab sets.
- Only direct `tab` children are tabs; panel content goes inside a `tab`.
- Generated classes: `.sui-tabs`, `.sui-tablist`, `.sui-tab`,
  `.sui-tabpanels`, `.sui-tabpanel`.
- Labels are escaped and the structure uses `tablist`, `tab` and `tabpanel`
  roles.

See [API reference: tabs](API_REFERENCE.md#tabs) and the
[tabs example](../examples/tabs.sui).

## Accordion control

The **Accordion** palette item under **Layout → Navigation** inserts a working
disclosure stack with two sections:

- `accordion` **Text** = accessible name (`aria-label`);
- `accordionitem` **Text** = visible section header;
- `accordionitem` children = panel content.

Right-click a tree row:

| Right-clicked row | Actions |
|---|---|
| `accordion` | **Add accordion item**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `accordionitem` | **Add item below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |

The first section is open. Clicking a header opens it and closes the others;
clicking the open header closes it. The open section is restored after a normal
view re-render. Generated classes: `.sui-accordion`, `.sui-accordion-item`,
`.sui-accordion-head`, `.sui-accordion-panel`.

See [API reference: accordion](API_REFERENCE.md#accordion) and the
[accordion example](../examples/accordion.sui).

## Combobox control

Insert **Combobox** from **Controls**. A new control starts with an empty
options state (`options1`, `options2`, ...) and a searchable input. Its tree
label is `combobox 'label'`.

The **Combobox** inspector section contains:

| Field | Meaning |
|---|---|
| options state | string state holding the comma-separated list; leave empty for a literal list |
| Options in ... | design-time rows read from the state initial value; add or remove options here |
| value state | state name that receives the selected stored value |
| placeholder | search placeholder |
| empty text | message when no option matches |
| max option length | maximum accepted characters per option label/value; exceeds = error |
| max options | maximum number of items accepted when the list is formed |
| required | require a selected value |
| disabled | disable the control |

Adding a row updates the bound state's **initial value**. The generated source
keeps `options: stateName`, so application code can assign a different
comma-separated string at runtime and the list refreshes on the next render.
Removing every row leaves the state empty. `Label=value` separates display text
from the stored value. Use `\,` in the add field for a comma inside a label;
the state source is escaped automatically. If a row exceeds **max option
length** or the list exceeds **max options**, the candidate list is rejected
with an error and no partial options are shown. The add rows themselves are
also checked before insertion.

Search supports `*` (any sequence) and `?` (one character); `\*` searches for a
literal asterisk. Invalid runtime data shows an error below the control. When a
replacement list omits the selected value, the editor's generated state binding
clears that value automatically; a value still present is retained.

A state whose initializer is a non-literal expression cannot be edited as a
row list; change its initializer in the States panel or manage the runtime list
in code. Leaving **options state** empty switches the control back to a literal
`options: "..."` list.

The Events section exposes `onChange`; the handler runs after the selected value
has been written to the bound state.

The preview supports typing to filter, `ArrowDown`/`ArrowUp` to move, `Enter` to
select, `Escape` to close and the clear button to reset the value.

See [API reference: combobox](API_REFERENCE.md#combobox) and the
[combobox example](../examples/combobox/README.md).

## Tree control

Insert **Tree** from **Controls** (the component palette), next to the other
controls. A new tree starts with a parent item and a child item. Add **Tree item**
nodes from the same palette or duplicate existing rows.

The **Tree** inspector exposes:

| Field | Meaning |
|---|---|
| tree name | runtime target for the core `tree` import/export functions; defaults to the title |
| JSON/XML data | import/export the tree items in the design source |
| value state | string state receiving the selected item key |
| style preset | 16 tree presets from Default and Compact to Neon, Forest, High contrast and Cards; custom edits style expr |
| rename icon | pencil glyph (for example `✏️`); empty uses `✎` |
| show rename icon | hide or show the pencil for the whole tree |
| editable | enable inline rename |
| editable depth | maximum editable level; `0` = all |
| rename key state | state receiving the renamed node key |
| rename value state | state receiving the new label |

The **Tree item** inspector exposes:

| Field | Meaning |
|---|---|
| key | stable value written to the selected state |
| initially open | default expansion state |
| node style preset | Inherit / Muted / Accent / Success / Warning / Danger / Folder / Compact row / custom |
| show rename icon | inherit / show / hide |
| rename icon | per-node glyph; empty inherits the tree icon |
| editable | inherit / yes / read-only |
| Text | visible item label (Content section) |

`onChange` and `onRename` are available in **Events**. The generated source keeps
nested `treeitem` children and their keys/open/editable flags.

The tree **style preset** menu writes the component's visual CSS into the tree
element's **style expr** field (in **Appearance**). It offers **Default**,
**Compact**, **Comfortable**, **Dense**, **Bordered**, **Soft rounded**,
**Minimal**, **Neon**, **Forest**, **Amber**, **Ocean**, **High contrast**,
**Outline**, **Mono**, **Candy** and **Cards**. Editing the `cssSafe("...")`
expression directly gives full control and switches the menu to **Custom**.

Every tree item has its own **node style preset** menu (**Inherit**, **Muted**,
**Accent**, **Success**, **Warning**, **Danger**, **Folder**, **Compact row** or
**Custom**). Node presets write row-level `--sui-tree-*` variables into that
item's `style expr`, so they override only the selected node. Both levels use
the same inspector **style expr** field.

The **rename icon** field changes the pencil glyph; **show rename icon** hides
it. On a Tree item the same fields can inherit, show or hide the icon per node.
Hiding the button does not disable F2 or double-click rename.

The **tree name** field makes the tree a runtime target. The **JSON/XML data**
buttons import or export the tree items in the design source: import replaces the
selected tree's `treeitem` children, export downloads the current model.
Design-time import does not store data at runtime; use the core `tree` module
(`tree.importJson`, `tree.importXml`, `tree.importJsonFile`,
`tree.importXmlFile`, `tree.exportJson`, `tree.exportXml`,
`tree.downloadJson`, `tree.downloadXml`) for application data. The example
combines a `pickFiles` button with the file helpers and a download button. See
[`STDLIB.md`](STDLIB.md#tree-module-tree-core) for the JSON/XML schema.

The preview supports row selection, disclosure toggles, arrow-key navigation,
Home/End and inline rename (double-click, F2 or pencil; Enter/Escape). Expansion
and renamed labels are restored after a state-driven re-render. The runtime uses
`localStorage` for open state and local label overrides; labels and keys are
escaped.

See [API reference: tree](API_REFERENCE.md#tree) and the
[tree example](../examples/tree/README.md).

## Chart control

Insert **Chart** from **Controls**. A new chart starts as a bar chart with
sample CSV data, fills its container (`width: "100%"`, `height: "100%"`) and
uses `overflow: "scroll"`. New charts keep the table hidden until **table view**
is set to `table` or `both`.

The dedicated **Chart** inspector exposes:

| Field | Meaning |
|---|---|
| type | bar, horizontal bar, stacked bar, grouped bar, line, step, smooth, area, combo, pie, donut, scatter, histogram, waterfall, radar, gauge, funnel, heatmap, candlestick, treemap, box plot, bullet, dumbbell or Gantt; the source attribute is `chartType` because `type` is reserved. Block diagrams use the separate **Graph** control |
| data | button that opens the movable, resizable chart-data overlay |
| x field / y field | label and numeric columns in object/CSV data; empty auto-detects |
| point width | horizontal pixels reserved per point in scroll mode |
| point size | radius of data points in pixels (1-24, default 3) |
| bins | histogram bucket count (1-100) |
| min / max | gauge range (default 0-100) |
| interactive | enable tooltips, hover/selection and Gantt editing |
| table view | chart only, accessible data table only, or both; available for every type |
| initial zoom | starting zoom for zoom mode (0.5-4; Gantt timeline 0.5-24) |
| overflow | Scrollbars, Zoom controls or Fit to frame |
| show grid / show legend / show values | chart presentation toggles |
| preset | 14 color presets that write CSS variables into `style expr` |

The **Chart** inspector is type-aware: irrelevant fields are hidden when the
selected type cannot use them.

- **bins** only for histogram;
- **min/max** only for gauge;
- **x field** hidden for gauge and histogram;
- **y field** hidden for gauge;
- **point width** hidden for pie, donut, radar, gauge, funnel, treemap, bullet, dumbbell and Gantt;
- **point size** shown only for line, step, smooth, area, scatter, combo, radar and dumbbell;
- **table view** is available for every chart type;
- **show grid** hidden for pie, donut, radar, gauge, funnel, heatmap, treemap and bullet;
- **show legend** hidden for gauge and funnel;
- **show values** hidden for pie, donut, scatter, gauge and radar.

Changing **type** updates the visible fields immediately. If the current data is
not compatible with the selected type, a confirmation dialog warns that the
existing data will be ignored and replaced by a sample dataset suited to the new
type; cancelling restores the previous type. Compatible data is kept. If
`chartType` is an expression, all fields stay visible and compatibility is
checked at runtime.

Gantt uses the same data overlay but reads richer rows: `task`, `start`, `end`
(or `duration`), `progress`, `milestone`, `group`, `parent`, `dependsOn`,
`dependencyType`, `lag`, `constraint`, `constraintDate`, `resource`, `status`,
`color`, `cost`, `actualStart`, `actualEnd`, `remaining`, `work`, `workDone` and
`resources` (an array or a comma/semicolon list; legacy `resource` is accepted).
Dates, progress bars, dependency arrows, weekends and the today marker need no
extra inspector fields.
`parent` builds a WBS outline: an expanded summary is a header row, and only a
collapsed summary shows a summary bar; it rolls up dates, progress and cost.
Click the collapsed bar, its name, the caret or the table toggle to expand it.
`group` is different: it is a visual phase that leaves the task a real task with
its bar and links. Use the table **Group** cell for one task, or the context menu
**Assign to group...** to assign one or several selected rows without changing
their WBS parent. To turn a selection into a real WBS container, select tasks
with Ctrl/Cmd + click in the chart or select rows in the table, then choose
**Create group from selection**: a **New group** summary node is inserted above
the first selected task and becomes the parent of the selected tasks.
The Gantt header exposes **timeline zoom** controls; they change the time-axis
width without scaling labels or row text. A horizontal scrollbar appears in the
chart viewport when the zoomed timeline is wider than its frame. Ctrl/Cmd +
wheel over the timeline zooms at the cursor; a two-finger touch pinch zooms at
the gesture centre without page zoom. The chart frame keeps its configured size. The left task-name column has a fixed
width, stays left-aligned even when the timeline is narrower than the frame, and
stays fixed during horizontal scroll/zoom. Long task/resource labels wrap onto
up to three lines in that column; the table Task column wraps too. **Progress %
column** adds a fixed percentage column after the names, and deadline markers
can be dragged horizontally. Dependency arrows use the shortest
clear orthogonal lane and re-route live while a task is moved or resized. The
arrow tip rests on the target's left edge and a horizontal tail of at least 4px
connects it back to the shared entry point, so no link ends with a vertical
segment against a bar. Vertical routes never reverse direction toward the
target row, and the Gantt stage grows with `linkGap` (vertically for row
corridors, horizontally for parallel lanes) so the viewport can preserve the
clearance; warnings and the table are re-evaluated at the same time. Every Gantt edit
rechecks dependency type/lag, constraints, deadlines and effort values; routes
also avoid task bars and already-used horizontal/vertical segments. **Time scale** in Chart settings
selects tick granularity (auto, day, week, month or year). Alternating vertical
period bands follow the active scale (days, Monday-based weeks or months), with
theme-aware colors for dark and light charts; weekly ticks start on Monday and
the weekly scale never draws non-working-day bands. At day scale non-working
days keep the weekend/holiday color when they are shown. Drag the border at the
right of the task-name column to resize it; the width is saved per chart and axis tick
labels stop at that separator while the timeline is scrolled. The chart menu
can show/hide the progress column and the non-working-day bands
without changing the schedule or the table data. Showing the progress column
hides the inline percentages beside the task bars.
Collapsed groups hide their table rows too, and a group-toggle strip above the
table expands or collapses them from the table alone.
`dependsOn` accepts task ids or task names; the table shows removable name chips
and a task-name combo for adding multiple predecessors. Click a chip to edit that
link's type and lag individually. Drag the small connector handle under a bar's
right edge onto another bar or task name to create an FS dependency directly on
the chart. Dependency cycles are rejected. Critical path, auto schedule, holidays, weekend, timezone and **Show non-working days** are configured in
**Chart settings → Interactions**. The last option is on by default and only hides or shows weekend/holiday bands; date validation remains active. Holidays accept exact dates, ranges and yearly
recurrences; weekend accepts weekday names/numbers or `none`; timezone accepts an
IANA zone such as `Europe/Rome` (default `UTC`). The chart menu's **Calendar...**
overlay adds holidays and closure ranges; the calendar is stored per chart and
immediately recalculates effective duration. **table view** is available for
every chart and can show a native accessible data table, or display the table
under the chart. Gantt tables use Task, WBS parent, Start, End, Deadline, Actual
start/end, Duration, Effective duration, Remaining, Progress, Work, Work done,
Milestone, Status, Group, Resources, Depends on, Link, Lag, Constraint,
Constraint date, Slack and Cost. Resources are edited as removable chips with
suggestions from the resource registry. The table tabs switch between **Tasks**
and **Resources**: the tab bar stays visible while the table scrolls. The
Resources tab edits Name, Role, Max people, Daily cost and Notes, adds rows and
deletes unused resources. Renaming a resource updates its
task assignments; deleting a resource still in use is refused. Each resource
chip has a quantity (`×N`) for the people required by that assignment. The small
**+** button opens a resource-type combo with a quantity field; confirm to add.
Max people
limits the sum of quantities of concurrent tasks, and edits that would exceed it
are rejected; Daily cost feeds the automatic task Cost calculation. The chart menu's **Resource workload...** dialog
summarizes tasks, role, workload, work, completion, peak demand, max people and
daily cost per resource. The registry is saved in `localStorage` and can be
mirrored to the application with `onResourceChange`; **Reset saved** clears the
local copy and rebuilds it from the chart source.
Actual/remaining/work fields drive execution and effort-based progress; actual
bars appear in the chart. Duration is the calendar difference end-start;
Effective duration is the
working-day count using weekend, holidays and timezone. The Constraint combo and
its column filter always show all seven types with descriptive labels, even
when every row currently uses `ASAP`; other types use Label plus their series or
type-specific columns.

With **table view** set to **both**, every chart docks a fixed-height table at
the bottom: the chart shrinks, the table spans the full width and stays aligned
with the bottom edge of the control. Drag the separator above the table to set
its height; the value is stored per chart. The separator is keyboard focusable:
Up/Down resize by 8 px (Shift+Up/Down by 32 px), PageUp/PageDown by 10% of the
control height, Home resets to 80 px and End uses the maximum. Adding or scrolling rows moves only
the table body, while the header row and the table tools above it stay fixed.
Resizing the container keeps the table height and gives the extra space to the
chart. The Graph control follows the same rules.

**Block diagrams** use the dedicated **Graph** control (`graph "..." data: ...`),
inserted from **Controls**. It uses a generic shapes/links model instead of
numeric series. The data overlay accepts an object with
`initial`, `states` and `transitions`, a state array with nested transitions, a
CSV edge list or XML state/transition elements. Shapes have a `shape` subtype
(rectangle, square, circle, stadium, ellipse, rhombus), fill, border
style/width/color, corner radius, dimensions and tag/note.
Only explicit tags are rendered on the diagram; shape names, labels and
descriptions stay in the table and inspector. Links have a line subtype
(straight, arc, bezier, orthogonal, circular), width, dash, color, start/end
markers and an optional visible tag/note. The diagram is laid out automatically by level,
with explicit `x`/`y` coordinates available for manual placement. Legacy
documents using `chartType: "state"` (`statemachine` alias) still open and render;
the old type is shown as **Block diagram (legacy chart)** only for those charts.

The gear button in the chart header opens a tabbed **Settings** panel: **General**
(chart title, shape border/highlight colors, theme, show-anchors switch plus one
row of GRID/SHAPES/LINKS/TAGS buttons, save-as-default and reset),
**Versions** (pending changes, commit and version history) and **Data**
(import plus JSON/CSV/SVG/PNG export). Closing a GRID/SHAPES/LINKS/TAGS
sub-panel restores the General panel. The SHAPES defaults use two aligned
columns with matching row heights. The graph itself keeps only the diagram,
the shapes/links table and the add controls. **Export JSON** embeds the full
shapes, links and settings so re-importing it reconstructs the diagram exactly.
The **Edit mode** switch in the Interactions tab toggles between **Editing** and
**Static**. Static mode removes selection, drag/resize and editing handles while
keeping pan/zoom; the same value is available as `stateMode` and in the designer
chart settings. In static mode every interaction is queued by the chart; SUI
consumes the queue with `chart.pollEvents` / `chart.nextEvent` and decides which
external function to call. The queue item contains `event`, `kind`, `part`,
`id`, `label`, `tag` and pointer/modifier state. Legacy direct callbacks
are still available (`chart.on`, element data fields, chart-level `onShapeClick`
etc.) but the event list is the recommended model. Shape ids are mandatory and
unique; the designer reports missing or duplicate ids as an import error.
**Hover hints** in the Interactions tab maps to `stateHints`; turn it off to hide
all tooltips. The **HINTS...** button opens the global tooltip style panel
with two groups: switches/colors on the left and size/spacing sliders on the
right. The background swatch includes its own opacity slider, so no separate
transparent button is needed; it writes the `stateHint*` chart attributes. Each
shape/link also accepts a `hint` field, editable from the table **Hint** column
or the context menu **Hint settings...**; use `|` for a line break and `\|` for a
literal pipe.
**Element hints...** in the chart inspector lists every state and link found in
JSON, CSV or XML chart data (including a literal document held by a simple state
such as `data`) and edits their tag and hint text in single-line fields (`|`
breaks a line, `\|` a literal pipe). It writes the data back; CSV needs a header
row with `id`/`label` or `from`/`to` plus `tag`/`hint` columns, which are added
when missing. Computed expressions still use the data editor.
Chart attributes that the inspector does not expose yet (grid, routing, tag or
link styles, ...) are preserved verbatim when the document is opened and saved;
use **Chart attributes...** in the chart inspector to edit them without leaving
the designer.
With the table visible, selecting an element in the diagram focuses and
highlights its row (the **Links** tab for edges).

The grid uses `stateGridGap` as its cell size and minimum line distance, with
`stateGridColor` for its color; it is visible by default (24 px) and can be
hidden from the settings panel (GRID). Circle centers and all shape vertices snap to grid vertices;
rectangle/square/stadium boxes are grid-sized and grid-aligned, and rhombus/
ellipse centers/extremes follow the same grid. Orthogonal horizontal/vertical
segments, stubs and fallback lanes all lie on grid lines. Link anchors are named grid-line/perimeter
intersections (`V<col>T/B`, `H<row>L/R`; fallback `P:<side>`). They are shown
only when the element is selected and editing handles are enabled, and link properties display the
current anchor names and coordinates read-only. Selecting a link shows draggable
source/target handles; dropping on another anchor immediately recomputes the route
with the obstacle-avoiding router and the link subtype. Shape X/Y/width/height and
all anchors live in
the hidden row metadata and in the element properties panel, so the table keeps
the complete reconstruction data without extra columns. Hovering a shape or link
always shows its name and description, even when no tag is rendered. **Description**
and **Tag / note** in the properties panel are single-line fields; in tags write
`|` for a line break and `\|` for a literal pipe. Orthogonal links leave
and enter perpendicularly for at least `stateGridGap`, then take an A* or outer-lane
route; self-loops are normal links between two distinct attachment points of the
same shape (never the same point). Create a self-loop in link mode by clicking the
same shape twice, then drag either endpoint handle to change that attachment point
independently. The chosen subtype shapes the self-link too
(`straight`/`arc`/`circular`/`orthogonal`/`curved`), and Bézier links expose two
control-point handles. Circular links expose a radius handle that also flips the
arc when dragged across the chord, and a selected link highlights its whole stroke
including the arrowheads. Handle layers sit above shapes, so radius/vertex handles
stay reachable even on circles; while dragging, the board is frozen and the moved
element is revealed on release. Endpoint
handles can also be dropped on another shape to reattach that link end. Link
properties add **Direction** (reverse source/target) and the **LINK** panel can
draw the arrow at both ends. If the router cannot find a clean path,
the diagram is left unchanged and a non-blocking notice explains the problem.
Moving/resizing a shape snaps
continuously to the grid, persists the snapped geometry, recomputes anchors and re-anchors
links that lost their previous attachment. Zoom the diagram with Ctrl/Cmd +
mouse wheel or a two-finger pinch; the zoom controls show 50%–400%.

The table is split into **Shapes** and **Links** tabs. Shapes show `Element`,
`Tag`, `Type` and actions; Links show `Element`, `Tag`, `Type`, `From`, `To` and
actions. Element is first and contains the color picker, Tag is second, and Type
selects geometry/line variant. Clicking a graph element or its tag selects the
same item, opens the matching tab and scrolls its table row into view.
**Ctrl/Cmd/Shift+click** toggles elements in a multi-selection; **left-drag** on
empty canvas draws a selection rectangle, while **right-drag** pans the whole
graph. **Alt+drag** and middle-drag also pan. Left-clicking empty canvas clears
the selection. A selected group can be dragged as
one unit and removed together; its relative spacing is preserved and the final
translation snaps to whole grid steps on both axes. Dragging beyond the viewport
expands the board instead of clamping the shapes. **Ctrl+C**/**Ctrl+V** copy and paste,
**Ctrl+D** or **Duplicate** in the context menu duplicates; duplicating shapes
does not copy their links unless the links are selected. With two or more
shapes selected, a floating alignment toolbar provides left/center/right,
top/middle/bottom and horizontal/vertical distribution, with blue magnetic
guides during drag. Arrow keys move the selected shapes by one grid cell;
Shift+Arrow resizes a single shape (Left/Right width, Up/Down height) with the
top-left corner anchored. Delete/Backspace removes the selection. Link properties are tabbed (**Line**, **Arrows**,
**Anchors**): line/highlight colors are stacked, width uses a slider with a
numeric field, dash type and start/end markers show live graphical previews, and
anchors are read-only. The link label stays in the table; description, event,
guard, action and tag/note fields are not part of this panel. Shape properties are
tabbed too (**Shape**, **Position**): in **Shape** the border settings (borders,
style, width, corner radius, dash length/gap for dashed/dotted) are on the left,
subtype and the fill/border/highlight colors on the right, with a live shape
preview at the bottom; **Position** holds rotation, X/Y, width and height in two
top-aligned columns. The label stays in the table and text styling is in Tag
settings. The table tab bar holds one **Add row**
button at its right edge; it opens the Shape/Link
choice. The table has no own settings or undo/redo buttons: import/export live in
**Graph settings**, and the graph undo/redo buttons cover both diagram and table
edits. A compact **+** button in the chart header opens one create menu with the
six shape subtypes and **New link**, so it cannot cover diagram nodes; its tooltip
shows `+`/`N`, and `L` creates a link directly.
Selection is synchronized between graph and table with an amber selection color.
Hovering a table cell shows a tooltip with its full text; the Type combo box is
excluded because its value is already readable. `|` line breaks are displayed on
separate lines, exactly as they appear in the diagram.
The shape context menu contains **Shape properties...**, **Tag settings...** and
**Delete** only: text styling lives in **Tag settings...**, rotation is edited in
the **Position** tab, and deleting a shape removes its connected links. Undo/redo use icon buttons with accessible hints. The `…` button opens
the element properties panel (shape/link styles, X/Y, dimensions, text, tags,
markers and read-only anchors). Each Element cell starts with a color picker.
Shape and link
tag families have independent box, leader, font and size settings, configured
from the single **TAGS** sub-panel in the settings **General** tab; the overlay
also sets the tag arrow size and the between-shapes link arrow size (both default
to `stateGridGap / 4`). Tags avoid shapes, other tags and routed lines. The
settings **General** tab also exposes the GRID, SHAPES and LINKS sub-panels,
dark/light theme, **Save settings as default** for new state charts and
**Reset settings** to restore the `.sui` values. Diagram settings also
include the **Show anchors and handles** switch: turn it off for view-only
diagrams (`stateHandles: false`), which hides anchors, resize/rotation grips, tag
and route handles and disables shape dragging. The grid itself is optional from
the **GRID** panel (`stateGrid: false`). Orthogonal routes are
audited against the grid, other shapes and previously routed links; failed
candidates retry wider outer lanes and expose `data-sui-chart-route-audit` on the
chart. The table toolbar provides add/delete, undo/redo, import and JSON/CSV
export. Right-click a link for Link properties, Tag settings and manual
waypoints: **Add waypoint here** inserts a draggable grid-snapped vertex at the
click, and **Clear manual waypoints** removes them.
State edits are auto-saved and support Undo/Redo; **Commit changes…** in the
settings **Versions** tab creates versions with field-level audit,
**Version history…** inspects them and `onCommit` handles server
synchronization. See
[`examples/state_machine/`](../examples/state_machine/README.md).

Each table header has an Excel-style filter button with a popup. The popup
combines a contains search, a checklist of distinct values and min/max fields
for numeric or ISO-date columns. Click a header to sort; Shift+click adds a
multi-column sort. Drag headers to reorder columns. Drag the right edge of a
header to resize the column; double-click it to auto-fit the column to its
widest cell content. The resize handle is also keyboard accessible with
Left/Right/Home. Widths are saved with the table and **Columns → Reset column
widths** restores automatic sizing. The identifying first column is sticky during
horizontal scrolling. The **Settings** gear opens grouped
submenus: Columns, Filters and Export. Export writes visible rows as CSV, TSV,
JSON or JSONL, copies visible rows as TSV, or opens a print view. The **Add row
button** option in **Chart settings → Table** enables **+ Row** for every chart
type; Gantt rows are editable, other charts receive a zero-valued local row and
can notify `onDataChange`. The same pane controls **Delete row button**, **Number
steppers**, **Context menu** and **Selection** (`none`, `row`, `column`, `cell`).
Enabled numeric steppers sit at the right edge of numeric cells and are visible
only after a click focuses the cell, not on hover alone, and are hidden while a
selection drag is in progress. Starting a selection cancels the pending
edit and clears its focus. Row, column and cell selections share the same
selection color and a 2px `--sui-chart-table-selection-border` perimeter. In
cell mode only the cell is filled; the row remains linked to the chart without
a full-row tint. Hover focus uses
subtle neutral row/cell backgrounds (`--sui-chart-table-focus-row-bg` /
`--sui-chart-table-focus-cell-bg`); clicking or dragging removes the hover
highlight and leaves only the selection. Keyboard focus is a subtle
`--sui-chart-table-focus-color` outline, so text and grid lines remain visible,
and selected rows/cells keep the selection color while hovered. Inline edits and
steppers update the chart and table in place; generated view refreshes preserve
the chart/table DOM, so editing does not flash or rebuild the table.
Use the row-number gutter to select/drag rows, the strip above headers to
select/drag columns, or drag diagonally in the body to draw a cell rectangle.
Ctrl/Cmd+click toggles, Shift+click extends, and Alt+click sorts a header. The **Rows** menu and table
context menu copy, paste or delete the current selection; clipboard access is
requested by the browser and the result appears in the table status line.
**Export selection** exports only selected rows, columns or cells as CSV, TSV,
JSON, JSONL, XLSX or HTML. **Export options** control headers, hidden columns and
original/current row order. **Print / Save PDF (chart + table)...** prints the
chart SVG with the filtered table. Tables above 300 rows use windowed virtual
scrolling while filters and exports cover the full filtered set. Filter, sort,
order, column visibility and column widths persist in `localStorage` for that chart.
Headers sort on click or Enter; rows and headers are keyboard focusable. Filter,
sort and visible columns persist while the chart rerenders. The table caption is
shown in table-only mode and hidden when chart and table are shown together.

Right-click a chart, or use its **⚙** button, for SVG, PNG, CSV, TSV, JSON or
JSONL export, grouped in the **Export** submenu of the chart menu. Chart export
includes all rows; table Settings exports the visible rows and columns.

When **interactive** is on, Gantt Task, WBS parent, Start, End, Progress, Group and
Depends on cells are editable; non-Gantt Label/X/series cells are also editable and call
`onDataChange` (histogram bins remain read-only). With **tableAddRow** enabled, **+ Row** appends a new
task, selects it, clears active filters, redraws the timeline and calls
`onTaskChange`. Gantt and non-Gantt table edits can be undone/redone with
`Ctrl+Z`, `Ctrl+Shift+Z`/`Ctrl+Y`, or from the chart and table context menus;
inputs keep their native text undo while focused. A cell edit redraws the Gantt and calls `onTaskChange`; drag/resize on
the timeline redraws the table with the same values. Selecting a timeline bar,
its left-hand task label or its table row highlights all three representations
and scrolls the table row into view. A valid edit flashes the affected table cells and
the Gantt task for a short moment. Selection works for every chart/table pair:
a chart mark selects its row, and a row highlights all marks for that label
(including every series in multi-series charts). Selection is saved in
`localStorage` and restored after reload.

**Chart settings.** The **Chart settings** button opens the unified panel with
tabs: Appearance, Points, Tooltip, Table, Interactions and Data. It controls the
dark/light theme, **Editable colors** (color picker on/off), style preset, point size, tooltip colors/width/font/delay, table view, **Column
filters** on/off and table typography, interactive mode, **Critical path**, **Auto schedule**,
**Holidays**, **Weekend**, **Timezone**, **Link color**, **Link width** (default
`1px`), **Min link gap** (default `4px`), **Line style**, **Arrow**, **Arrow
size**, **Bar border** and **Bar border color**, and x/y fields. Apply writes attributes and
`--sui-chart-tooltip-*` / `--sui-chart-table-*` variables together. Dark and
light themes use explicit chart tokens; selection colors are shared and focus
tones adapt to the theme background. Click a series swatch in a table header or
legend to change that series color; the override is stored per chart. Chart
updates crossfade the old SVG into the new one unless reduced motion is
requested.

Charts can receive keyboard focus; use `ArrowLeft`/`ArrowRight` (or up/down) to
move through data marks, `Home`/`End` to jump, `Enter`/`Space` to select, and
`Escape` to clear the tooltip.

Tooltips, hover brightness, keyboard navigation and selection are available on
every chart. Gantt task rows are keyboard targets: Left/Right move,
Shift+Left/Right resize the finish, Alt+Left/Right resize the start, Enter opens
the task menu, and the dependency connector opens the target picker with Enter.
Focused dependency links support Enter/Shift+F10 to edit and Delete to remove.
**interactive** additionally enables Gantt drag/resize, collapsible groups and
editable Gantt table cells. Task edits are auto-saved to a local draft and are
not versioned until the header **Commit** button is pressed. **Commit changes...**
and **Version history...** in the chart menu create and inspect versions with
field-level audit details. If the chart has `onTaskChange` in **Events**, the
handler receives the updated task JSON; `onCommit` receives the committed version
for host/server synchronization with `revision` and `baseRevision`.

**Data overlay.** The **data** button replaces the old inline textarea. It opens a
floating panel over the designer with a large monospace editor. Drag the panel
header to reposition it and the bottom-right grip to resize it; the editor
remembers the last position and size. The panel has:

- **Literal data** and **State expression** modes;
- a format selector for auto, JSON, XML or CSV;
- live validation and row/character counts;
- **Load file** for `.json`, `.xml`, `.csv` or `.txt` data;
- **Insert sample** for the selected format;
- **Pretty JSON** for valid JSON;
- **Apply**, **Cancel**, backdrop click, Escape to close and Ctrl+Enter to apply.

Invalid JSON/XML is reported before Apply and blocks the update. The overlay also
checks **data/type compatibility**: single-series charts need one numeric series,
while stacked/grouped/combo need at least two; radar accepts several, heatmap
needs y+value, candlestick needs OHLC, box plot needs five statistics, bullet
needs value+target, dumbbell/Gantt need two series, pie/donut/funnel/treemap need
a positive value and histogram needs numeric values. The type selector in the overlay reports the
mismatch and disables Apply until it is resolved. Applying a literal data value
writes the escaped SUI string automatically. Expression mode stores the
state/expression and makes the chart reactive. The panel remains a
designer-only tool: the generated app keeps the normal chart runtime and its
data limits.

Chart width and height come from the standard **Size** inspector fields (or
CSS in **style expr**), not from chart-specific parameters. Use those fields to
set the visible frame. When the size is omitted, the chart fills its container
and grows with it; the table follows the container width while keeping its
fixed height. In a container without a definite height the chart keeps at least
240 px (`--sui-chart-min-height` overrides the fallback).

When `overflow` is **Scrollbars**, the chart draws at its natural width inside
the configured frame and the viewport scrolls horizontally when the points need
more space.
**Zoom controls** adds `-`, `+`, `Reset` and a percentage label; zooming in grows
the drawing and exposes scrollbars. **Fit to frame** keeps the drawing inside
the frame without controls.

The **preset** menu writes `--sui-chart-*` variables into the element's
**style expr** field under **Appearance**. Presets: **Default**, **Ocean**,
**Forest**, **Sunset**, **Neon**, **Mono**, **Pastel**, **High contrast**,
**Amber**, **Candy**, **Slate**, **Light**, **Outline** and **Minimal**. Choosing **Custom**
leaves the current expression untouched; edit it directly for series colors,
grid/axis colors, spacing, line width and scrollbar colors.

A data-expression chart is reactive: changing the bound state re-renders the
chart. Malformed data or an empty payload shows a text error inside the chart,
not a broken preview. JSON literal braces must be written `\{` and `\}` in
source; the data overlay does that automatically when applying a literal.

See [API reference: chart](API_REFERENCE.md#chart) and the
[chart example](../examples/chart/README.md).

## Graph control

Insert **Graph** from **Controls**. A new graph starts as a block diagram with
sample JSON states and transitions and fills its container (`width: "100%"`,
`height: "100%"`). It is a leaf control: it cannot contain child elements.

The graph inspector is ordered as follows:
| Field | Meaning |
|---|---|
| Graph settings | first button; opens the same runtime panel as the chart gear (General plus GRID, SHAPES, LINKS, TAGS and HINTS) |
| Styles | color preset and **Edit styles...**, which opens the Graph style editor with Title, Frame, Buttons, Table and Scrollbars tabs |
| show table | shows or hides the synchronized data table |
| edit table | allows editing table cells, colors, shape types, element settings and deleting rows; when off the table is read-only and the actions column is hidden |
| settings button | shows or hides the gear that opens Graph settings |
| zoom buttons | shows or hides zoom, reset and fit buttons |
| add element button | shows or hides the + Add shape/link button |
| undo/redo buttons | shows or hides the history buttons |
| edit data | last control; opens the JSON/XML/CSV or expression data editor |

The graph type selector is not shown: `graph` always renders the block diagram
geometry and never writes `chartType` in the source. Editing the chart type in
the data overlay is disabled for the same reason. The inspector no longer shows
initial zoom, interactive, element hints or chart attributes for Graph.

The runtime **Graph settings** panel opened from the gear (or from the inspector
Graph settings button) has no separate Style tab. The **General** tab contains a
**Theme** area with a preset combo box (Midnight, Graphite, Ocean, Forest,
Sunset, Candy, Neon, Slate, Paper, Sand, Ice, High contrast, Plum, Mono and
Custom) and an **Edit theme...** button. The button opens the theme editor panel,
split into **Control**, **Title**, **Buttons**, **Table** and **Scrollbars**
tabs. It exposes the current values of every non-diagram component: control
background, border, radius and padding; title bar background and height, title
typography, decorations and shadow; button colors for each state, border
width/radius/font; table colors, borders, typography, header and input tones,
plus hover and selected-row tones; and scrollbar size, colors and radius.
Fonts, font style, decoration and shadow use combo boxes. Presets also define
readable defaults for
axis, grid, legend and value labels, tooltips/hints and error/critical accents.
Shapes, links, tags, hints and grid style keep their own panels, which override
the theme defaults. The editor starts from the currently applied values, so
after choosing a preset its fields show that preset, not the defaults. Fields
are arranged in two columns, each with its label on the same row as the control;
color fields show the hex value, while the picker panel contains an alpha slider
and an alpha text field. Color fields are grouped in the same column and
slider/number fields in the other; combo boxes show a dropdown chevron so they
are not mistaken for text fields. Changes
apply immediately, are stored per chart, and the same control and button styles
are used by the settings panels inside the graph. **Reset theme** in the editor
(or **Reset settings**) clears them.

The designer **Graph style editor** uses the same component categories in Title,
Frame, Buttons, Table and Scrollbars tabs. Changes are previewed live on the
graph and become source values only when **Apply** is pressed.

The graph fills its container and starts with the table hidden. With **show
table** on, the table spans the full width at the bottom of the control and the
diagram shrinks, so chart and table share the available height. Drag the
separator between them to set the table height; the height is stored per chart.
The separator supports the same keyboard controls (Up/Down, Shift+Up/Down,
PageUp/PageDown, Home and End). Adding rows scrolls only the table body, while the header and the buttons above
it stay fixed. Resizing the container keeps the table height and gives the extra
space to the chart.

**edit table** enables editing in the table: the Element cell shows a color
picker and text field, Tag and Hint are editable, Type becomes a select, and the
last column exposes text settings, element properties and row deletion. With it
off, the table is read-only: cells show plain text, the actions column is hidden
and the Add row button disappears. While typing in a cell its tooltip is
suppressed. **Edit mode** in Graph settings also enables table editing. Graphs
default to static editing in the designer; existing shapes and links are edited
from the Graph settings panel, while the + Add button can still be shown with
its switch.

## Date/Time picker control

Insert **Date/Time picker** from **Controls**. A new control starts in `date`
mode with an empty value and the label `Date`.

The **Date/Time picker** inspector section contains:

| Field | Meaning |
|---|---|
| value state | string state receiving the ISO-like value |
| mode | date, time or datetime |
| min | native minimum value in the selected mode format |
| max | native maximum value in the selected mode format |
| step | optional seconds for time/datetime |
| disabled | disable input and clear |
| required | native required flag |

`onChange` is available in **Events**. The generated source keeps the mode and
range attributes; the value is bound through a hidden input and synchronized by
the runtime.

Value formats:

- date: `YYYY-MM-DD`
- time: `HH:MM` or `HH:MM:SS`
- datetime: `YYYY-MM-DDTHH:MM` or with seconds

The preview uses the browser's native picker. The clear button resets the value
and state. Invalid or out-of-range state values show an inline error.

See [API reference: date/time picker](API_REFERENCE.md#datetime-picker) and the
[datepicker example](../examples/datepicker/README.md).

## Toast control

The **Toast host** palette item under **Overlay** inserts an empty live-region
container. It is optional: without it, the runtime creates a fixed top-right
host automatically.

Use a handler to show feedback:

```sui
fn save() -> null {
	toast.show("Saved", "success", 3000)
}

fn clear() -> null {
	toast.clear()
}
```

- `toast.show(message [, kind [, duration]])`;
- kinds: `info` (default), `success`, `warning`, `error`;
- default duration: `4000` ms; `0` keeps the toast open;
- default maximum four toasts per host (`max` supports 1–8); the oldest is removed;
- users can close each toast with its X button;
- `toast.clear()` removes all visible toasts;
- messages use `textContent`, never raw HTML;
- the host uses `aria-live="polite"` and each toast uses `role="status"`.

Select the **toast host** in the tree to open its dedicated **Toast** inspector
panel. The panel contains:

| Section | Controls |
|---|---|
| Placement | horizontal `left`/`center`/`right` and vertical `top`/`center`/`bottom` |
| Offsets | horizontal offset only for left/right; vertical offset only for top/bottom |
| Behaviour | default duration, maximum visible, show icons |
| Preview | preview kind, preview message, **Show preview**, **Clear preview** |

Each axis is independent. A centered axis has no offset field because it is
centered in the viewport; an edge-anchored axis shows its own offset field
(default `16px`). A bare number is interpreted as pixels; write `%`, `em`,
`rem`, `vw` or another unit explicitly when needed. The generated source uses
`hAlign`, `vAlign`, `hOffset` and `vOffset`.

For stacked notifications, a top anchor appends each new toast below the
previous one; a bottom anchor reverses the stack so each new toast appears above
the previous one. The oldest toast is still removed first when the maximum is
reached.

The normal **Size**, **Border**, **Box model**, **Typography**, **Background**
and **Effects** inspector sections below the panel write the toast card rule.
Width and height therefore use the standard **Size** fields, and border, radius,
background and typography changes apply directly to the toast card.

The same declarations are shown in the **style expr** field for the Toast host.
That field is the editable CSS declaration list for `.sui-toast` (for example
`width: 420px; border-width: 3px;`). Change it to evaluate and update the card
style; custom properties such as `--sui-toast-icon-bg` are inherited by the
icon, message and close button.

Changes update the generated source and CSS immediately. The preview iframe
shows a live sample toast using the real app CSS, so placement, size and border
edits are visible before leaving the designer. If the project contains multiple
toast hosts, the sample appears in the selected host; at runtime `toast.show`
uses the first host in the document.

The host's tree label is `toast host`; the Text field is not used. Without a
designer host the runtime creates a fixed top-right host.

See [API reference: toast](API_REFERENCE.md#toast-notifications) and the
[toast example](../examples/toast/README.md).

## Tooltip property

Tooltip is an **element/control property**, not a standalone palette item. Select
any element (for example a button, link, input or card) and open the collapsible
**Tooltip** section in the inspector. Its `text` field sets the hover/focus
message; clear it to remove the property.

The target receives `aria-describedby` while the bubble is visible; the bubble
uses `role="tooltip"` and escaped text. Keyboard focus and `Escape` dismissal
work automatically for focusable targets.

The tooltip style of the generated app is managed in **Preferences → Project
preferences.** This panel writes `--sui-tooltip-*` into the project CSS. Editor
tooltips are separate and live in **Preferences → Editor preferences → General
editor**.

| Setting | Meaning |
|---|---|
| Place | `top`, `right`, `bottom` or `left` |
| Delay | hover delay in milliseconds |
| Background / text / border | bubble colors |
| Radius | bubble border radius |
| Max width | maximum bubble width |

The global panel writes `--sui-tooltip-*` custom properties into `:root`, so the
settings persist in the project source and apply to every tooltip. The runtime
positions the bubble in the viewport, including for void controls such as
`input`.

See [API reference: tooltip](API_REFERENCE.md#tooltip-property) and the
[tooltip example](../examples/tooltip/README.md).

## Native file and folder dialogs

Use `pickFiles` for a dialog instead of dragging:

```sui
button "Choose files..." pickFiles: attachment multiple: true
button "Choose folder..." pickFiles: attachment directory: true pickMode: "auto"
```

`pickMode` values (default `native`):

- `native` - the browser's `<input type=file>`/`webkitdirectory`;
- `fs` - the File System Access API (`showDirectoryPicker`);
- `auto` - `fs` when available, otherwise `native`.

`fs` and `auto` are folder-oriented: they use `showDirectoryPicker()` to return
a directory handle and its files. For ordinary file selection, leave
`pickMode: "native"`. Use `directory: true` when the dialog must select a
folder.

The FS API needs a Chromium-based browser, a secure context and user permission.
It can return the name of an empty folder; the native `webkitdirectory` picker
cannot. On web targets there is still no absolute local path: only names and
relative paths are available, plus the bytes through upload.

On **desktop builds**, the separate `dialog` module returns true absolute paths:

```sui
import { dialog } from "std"

async fn choose() -> null {
	path, err = dialog.pickFile()
	if err != null { status = err; return }
	if path == "" { status = "cancelled"; return }
	status = path
}
```

`dialog.pickFile`, `dialog.pickFiles` and `dialog.pickDirectory` use native OS
dialogs and are covered in [STDLIB](STDLIB.md#dialog-module-dialog-desktop).
The path is still not a capability: `fs` requires `--allow-dir` and `exec`
requires `--allow-exec`.

## Designer inspector checklist

When an application drag interaction does not work, check:

1. **Source** has `dragKey` or `draggable: true`.
2. **Target** has `dropKey`.
3. **Target** has `onDrop` naming a declared function.
4. **File target** has `dropFiles` pointing to a string state and, for display,
   a matching `value:` binding.
5. The event function exists and is not a method unavailable in the current
   view.
6. The element is not a table if you expect the generic DnD inspector; table
   internals must be edited in source.
7. `list` rows and component internals may need source editing.
8. Folder picking with `fs` requires Chromium, localhost/HTTPS and permission.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| Designer component cannot be dragged | Invalid container, drop into its own subtree, or drop in the iframe | Use the tree, drop beside/inside a valid container |
| Generated row is not draggable | No `dragKey` or `draggable: false` | Set `dragKey` and `draggable: true` |
| Drop outline appears but nothing happens | No `onDrop` handler | Create/select an `onDrop` function |
| Only the first list row reacts | Custom wiring targets one element instead of the repeated row | Put `dragKey`/`dropKey`/`onDrop` on the repeated row; the runtime inlines them per row |
| File drop does nothing | `dropFiles` state missing or not a string | Add a string state and bind `dropFiles` |
| Field shows no paths | No `value:` binding | Bind the same state with `value:` |
| Folder picker has no empty-folder name | Native `webkitdirectory` selected | Try `pickMode: "auto"`/`"fs"` on Chromium |
| FS picker unavailable | Firefox/Safari or insecure context | Use `native` or serve on localhost/HTTPS with Chromium |
| Desktop picker returns an error | Web build or headless environment | Build with `--target desktop` and run in a desktop session |

## Security notes

- Treat dropped names and paths as untrusted text.
- Do not build HTML, CSS or URLs from them without validation.
- Only upload to a backend under your control.
- The desktop picker returns paths only after explicit user selection; it does
  not grant filesystem access by itself.
- Keep `fs` behind `--allow-dir` and commands behind `--allow-exec`.
- The runtime caps file counts, depth, paths and table data.

## Related documents

- [API reference: drag and drop](API_REFERENCE.md#drag-and-drop)
- [Builtins: drag and drop](BUILTINS.md#drag-and-drop-functions)
- [Agent recipes: drag and drop](AGENT_RECIPES.md#drag-and-drop)
- [Standard library: desktop dialog](STDLIB.md#dialog-module-dialog-desktop)
- [File picker example](../examples/file_picker.sui)
- [Menu bar example](../examples/menu_bar.sui)
- [File analyzer example](../examples/file_analyzer/README.md)
- [Desktop picker example](../examples/desktop_picker.sui)
