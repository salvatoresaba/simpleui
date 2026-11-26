# GUI editor: drag and drop

This guide explains the two independent drag-and-drop systems in SimpleUI:

1. **Designer drag and drop** - moving elements, palette items and toolbar
   sections inside the visual editor.
2. **Application drag and drop** - dragging elements, files or folders inside
   the generated application.

They share browser primitives but have different rules. Enabling one does not
enable the other.

Start the designer with an existing file with `simpleui gui path/app.sui`: the
editor opens that source immediately and **Save** writes back to the same path.
Run `simpleui gui` with no file to start from the last project or a blank one.
On Windows, `simpleui associate` registers `.sui` files with the editor under
`HKEY_CURRENT_USER`; `--uninstall` reverses it and `--dry-run` shows the registry
commands.

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

All settings panels are draggable from the grip in their header and restart
centered each time they open. Panels with an internal preview keep the dark
modal backdrop (Preferences, Component settings, Hints, Insert, Icon picker, File
browser, New page, Desktop build). Operational panels without a preview float
without a backdrop so their changes stay visible in the editor (Project
preferences, Tooltip, Table settings, Chart settings, Graph style, Structure
table, Tree data, Media queries, Project history and Chart data).

### Interface

Appearance and layout of the editor shell.

| Setting | Default | Effect |
|---|---|---|
| Theme | Dark | Dark, Light or High contrast editor chrome |
| Language | English | Editor interface language: English, Italiano, Espa\u00f1ol, Fran\u00e7ais or Deutsch |
| Interface font | System UI | Font family for menus, panels, labels and controls |
| Custom interface font | - | CSS family used when **Custom** is selected |
| Font size | 14 | Scales the whole editor UI |
| Density | Comfortable | Comfortable or compact spacing |
| Accent color | #4c5cf7 | Focus rings, selection and primary actions |
| Panels | all visible except history | Show or hide Structure/Tokens, Inspector and Problems/History |
| Toolbar layout | side by side | Both top bars share one row when space allows; drag a grip to reorder |
| Reset toolbar order | - | Restores the default toolbar layout |

### Interface language

**Appearance → Language** switches the editor interface between English,
Italiano, Español, Français and Deutsch. The choice is saved with the other
editor preferences and applied immediately to the chrome: menus, toolbar
tooltips, panel tabs, the inspector section titles, preferences and shortcut
labels.

Translation is dictionary-driven and falls back to English for any string that
has no translation yet, so a partially translated language never shows blanks.
The setting changes only the editor: the SUI source, the generated preview and
the application itself are not affected.

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

The **Shortcuts** tab lists every binding with its key combination. Factory
combinations can be changed: click a combination, press the new keys and the
change is stored in the editor preferences. **Esc** cancels the capture,
**Backspace** clears a custom binding and the arrow button next to a row
restores its default. **Reset all shortcuts** clears every custom binding.

Combinations must include **Ctrl**, **Alt** or **Meta**, or be a function key,
so typing in the editor is never hijacked. A combination already used by
another action is rejected with a status message. The toolbar tooltips and the
command palette show the current combinations.

Actions marked **fixed** (completion navigation, line operations, symbol
navigation, menu and chart keys) stay built into editing and cannot be
rebound. Shortcuts apply to the designer and to the SUI editor: for example,
rebinding **Undo** changes both.

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

### Formatting

**Shift+Alt+F**, or **Format document** in the command palette, normalizes the
SUI source without changing any token:

- indentation is rebuilt from brace depth with tabs;
- trailing spaces are removed and blank runs collapse to one line;
- a closing brace is aligned with its opening statement;
- the file ends with exactly one newline.

Quoted strings are never touched, including multi-line strings: lines that start
inside a literal, or open a literal that continues on the next line, keep their
whitespace because it is part of the value. Formatting is a normal undoable
edit and triggers the usual auto-apply.

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
The shortcut works from every panel of the editor: when the preview is the
active tab the editor switches to the **SUI** tab and opens the bar there.
The **Find** button in the SUI toolbar opens the same bar, so the search is
always reachable without a keyboard.

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
The SUI source can be edited by hand and applied: an edit that does not compile
is **underlined in the file** at the offending line, not only reported in the
status bar, and the same underlines appear while typing (live check).
CSS analysis filters generated noise: runtime classes (`sui-*`, `suitbl` and
state classes), keyframe steps, layered `:root` variable blocks, and
hover/focus/checked or pseudo-element rules are not reported as problems.
Author classes that are genuinely unused are still listed.
The analysis is **context-aware**: the same selector in another at-rule
context (a media query, a print sheet) is a conditional refinement, not a
duplicate or an override, and the structural pseudo-classes (`:last-child`,
`:nth-child()`, `:not()`, `:empty`) refine their base rule on purpose.
A class used inside a conditional or repeated block (`if`, `list`) counts as
used even though the designer keeps those blocks as source.

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

### Autosave

The editor writes a recovery copy to `localStorage` every 2 seconds. The
topbar shows **Autosaved HH:MM:SS**, **Unsaved changes**, **Saving…** or
**Autosave failed**. Autosave never overwrites the project file implicitly:
**Save** / **Save As** remain explicit. When a recovery copy differs from the
current editor content, the topbar shows **Restore autosave**; restoring
replaces the editor content and reloads the model after confirmation.

### Theme toggle

The toolbar theme button cycles the editor between dark and light. The choice is
saved in the editor preferences; the Preferences → Interface → Theme select
stays in sync. High-contrast remains available from the same preferences select.

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

### Element tree labels

Each tree row keeps the type chip and adds a readable name plus a code chip. The
name comes from the element text or heading, its placeholder, its `name`
attribute or its first class. The code chip shows the literal `motionKey` when
present, otherwise the tree path (for example `0.2.3`). Hovering the row shows
the full name, key and path together.

Use the `motionKey` or the path to distinguish repeated cards, rows and controls.
The same code is available in the trigger **event source** picker.

Clicking a tree row selects the element **and outlines it in the preview**, so
the two sides of the designer always show the same element. The outline follows
the selection in the other direction too: picking an element in the preview
selects its tree row. Either way the **inspector on the right** shows the
selected element, and the selected row is brought into view in the tree.

A **dynamic block** (the body of an `if` or of a `list`) has no rows of its
own: the tree shows it as a single **if block** / **list block** row and
picking one of its elements in the preview selects the element that contains
the block.

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
of a component in the project. The panel has a **Theme** combo and a **Name**
field above the category tabs; both follow the editor theme. Tabs are **All**,
**Feedback**, **Layout**, **Form**, **Navigation**, **Data**, **Overlay**,
**Charts** and **General**. Components are grouped by category into cards: the
**All** tab lays the cards out in a two-column table grid, while a single
category tab uses one full-width card. Each component remains a normal editor
button showing only the control name and opens its settings subpanel. The Theme
combo applies a shared style to every control at once, including scrollbars,
instead of setting each component by hand. Choosing **Custom** leaves the current
values untouched.

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
settings; closing it returns to the project list. The tooltip panel follows the
project theme, so a light theme gets a light bubble and a dark theme a dark one,
and it also exposes a **Letter spacing** field.

### Project themes

Every shared style has a **Light** and a **Dark** variant. The Theme combo lists
them grouped by family, for example Sketch Light/Dark, Terminal / CRT
Light/Dark and Geometric Light/Dark. Applying a theme writes the complete set of
component defaults for every control, updates the managed CSS rule blocks,
styles the scrollbars and sets the **page starting point** (`--sui-bg`,
`--sui-fg`, base border and hover colours). A light theme therefore starts from a
light body, and a dark theme from a dark body. Base elements follow the same
variables too: the view surface, headings, links, code blocks, table headers, the
menu bar and plain buttons switch with the theme. Tooltips, images, video, audio
and iframes use the same page palette, and the chart/graph theme presets use the
same family names. The **Standard** family provides
**Standard Light** and **Standard Dark**, which keep the default shape of every
control. Editing one component afterwards changes the Theme combo to **Custom**,
so it is always clear whether the project still follows a single theme. A new
project starts on **Standard Dark**; choosing **Standard Light** switches both
the page and the controls to light defaults.

Shared themes are also checked for readability: page text, panel text, muted
text, links, buttons, badges, alerts, cards, tabs, accordions, form labels,
trees, tables, toasts, modals and menus meet contrast thresholds against their
own surfaces. Links use the dedicated `--sui-link` token, which is derived from
the accent colour and darkened or lightened when needed, so an accent that works
as a fill does not become unreadable as link text. Active tabs and derived hover
colours are adjusted in the same way, and component borders stay visible on
light glass and standard surfaces.

### Export and import a theme

**Export theme** in the project preferences downloads a `simpleui-theme.json`
file containing the page palette, every component default, tooltip settings and
the chart/graph project defaults. The **Name** field sets the theme name and the
downloaded file name. **Import theme** loads that file back after a strict
validation: unknown fields, invalid custom-property names and CSS values that
could add declarations or rules are rejected. The import replaces the current
project theme, reapplies the imported chart/graph defaults to the existing chart
and graph elements, restores the theme name and refreshes the preview; undo
returns to the previous values.

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
options. Chart data and type stay per element, and the Data tab is
hidden in this mode. The defaults are stored in a managed
`/* sui-chart-defaults ... */` comment in the project CSS.

### Element overrides

Select an element and press **Settings...** in the inspector banner. The same
panel opens with a **This element only** scope. Values are written as inline CSS
custom properties on that element, so every other instance keeps the project
defaults. **Clear overrides** removes those inline values and restores the
project defaults for that element. Closing the panel returns to the inspector.

Implemented component panels: **Progress bar**, **Badge**, **Alert**, **Spinner**,
**Card**, **Divider**, **Images & media**, **Button**, **Input / Textarea /
Select**, **Combobox**, **Date / Time picker**, **Checkbox / Radio**, **Switch**,
**Slider**, **Tabs**, **Accordion**, **Menu**, **Tree**, **Table**, **Toast**,
**Modal** and **Scrollbars** (global). Chart and Graph keep their specialised
inspectors.

Each panel shows a live preview above the fields. The preview uses the current
draft values, not the saved ones, so colours, sizes and states can be checked
before Apply.

Text-bearing controls expose a **Letter spacing** field. Interactive controls
have a **Hover** tab with the colours used while the pointer is over them:
background, text, border, option, row or thumb depending on the component. This
makes hover feedback part of the same theme instead of an extra CSS edit.
**Progress bar**, **Spinner** and **Divider** also expose a **Hover** tab, so
non-interactive controls can still react to the pointer.

The **Images & media** panel themes the image, video, audio, iframe and PDF
viewer elements: background, border, radius, shadow, object fit and padding. It
participates in project themes and in per-element overrides like the other
component panels.

### Style presets and animations

Every component panel starts with a **Style** selector. **Custom** is the raw
default, and each named preset applies a coordinated set of colours, shapes and
effects in one step. Presets are optional: they never run unless selected, and
editing any field switches the selector to **Custom** so fine-tuning stays
explicit. Every shared family exists in **Light** and **Dark** versions, for 24
shared variants in total. The Style selector lists them under **Shared styles**;
the project Theme combo selects the same variants for all components at once.

Above the preview, **Preview background** switches between **Dark** and
**Light**. The choice is remembered and makes it easy to check a component
against both a dark application shell and a light or paper-like page. Selecting a
shared style also chooses a recommended background automatically (light for
Sketch and Brutalist, dark for the others); the buttons still let you inspect
the other one.

The presets cover all twenty-one component panels, for example **Neon glow** and
**Barber pole** for Progress, **Dual ring**, **Gradient ring** and **Coin flip**
for Spinner, **Glass** for Card and Toast, **Pill** for Button and Input, and
**Striped**, **Borderless** and **Rounded** for Table. Several presets include a
motion effect; the Animation/Effects tab also exposes the effects directly.

The animation library is organised by what the control does, so each effect
reacts to the control's own mechanism instead of being a generic loop:

| Control | Effects |
|---|---|
| Progress | none, pulse, stripes, shimmer, glow, **indeterminate** (scanning band), **aurora** (hue drift), **breathe** (vertical breathing), **ticker** (stepped blocks), **spark** (irregular flicker), **wave** (vertical bands) |
| Spinner | classic spin, pulse, flip, **reverse**, **orbit**, **wobble**, **bounce**, **ripple**, **swing** |
| Badge | none, pulse, glow, float, **pop**, **blink**, **tilt**, **ping** (radiating ring), **bounce**, **jelly**, **sheen** |
| Alert | none, fade in, slide in, shake, pulse, **zoom in**, **bounce in**, **slide right**, **wobble**, **breathe** (coloured glow), **flash** |
| Card | none, fade in, float, **rise**, **zoom in**, **tilt in** (3D), **halo**, **drift**, **unfold** |
| Button | none, pulse, float, glow, **jelly**, **bounce**, **tilt**, **ring**, **sheen**, **press** |
| Toast | slide down, fade, zoom, slide up, bounce, **flip**, **swing**, **pop**, **slide left**, **squash** |
| Switch | none, **pop**, **bounce**, **pulse ring**, **jelly** (runs when the switch turns on) |
| Checkbox / Radio | none, **pop**, **bounce**, **jelly**, **ring** (runs when checked) |
| Input / Textarea / Select | none, **focus ring**, **focus glow**, **focus lift** |
| Tabs | none, **pop**, **slide**, **glow** (runs on the active tab) |

State-driven effects run when the control changes state (checked, focused,
active tab) rather than looping forever. Every looping or entrance effect also
ships a `prefers-reduced-motion` guard: when the operating system asks for
reduced motion, decorative animations stop or the spinner slows down, while the
control keeps its colours and layout.

Spinner presets now include real shape changes: **Tri dots**, **Equalizer** and
**Comet** in addition to the ring variants.

### Spinner image/GIF mode

An animated GIF can replace the CSS ring. Select a spinner, open its **Spinner**
inspector section and tick **Image / GIF**. The **Theme**, **Structure**,
**Style**, **Animation** and phase selectors are disabled, and two source
controls appear:

- **Source** lists the images already uploaded to the project; choosing one
  applies it immediately.
- **Upload image** accepts GIF, PNG, JPEG, WebP and AVIF files up to 4 MB,
  stores the image as a data URI and applies it.

The generated element keeps the `spinner` tag and gains a literal or reactive
`src:`. The preview shows the animation and the generated source stays
self-contained because the image is embedded as `data:image/...;base64,...`.
Untick **Image / GIF** to return to the CSS spinner; the previous structure,
style and animation are kept. The media manager's **Src** button also works on a
selected spinner. See `examples/spinner_image/` for a runnable comparison.

Presets that showcase the new effects: **Aurora** and **Scanner** for Progress,
**Live** for Badge, **Arcade** for Button and **Aurora halo** for Card. The
**Style** selector still writes palette and shape together, so choosing a preset
and then changing the animation keeps the rest of the look.

The runnable **animation lab** in `examples/animation_lab` shows only the
animatable controls, each with runtime **Style** and **Animation** selectors, so
the effect can be changed while the app runs. Its README lists every style and
every animation.

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

The inspector opens the primary sections by default: **Content**, **Appearance**,
**Layout**, **Size**, **Typography**, **View** and, when it applies, **Animation**.
Other sections start collapsed. Use **Expand all** / **Collapse all** above the
fields; opening or closing a section stays remembered across selection changes
and preview reloads until you change it again.

Each section header has a small grip on the left. Drag the grip to reorder the
inspector sections; the order is saved in the editor's local storage and is
restored on reload. Sections that do not apply to the current element disappear
temporarily and keep their stored position when they return. Reordering only
changes the editor UI, not the generated application. The **View** section
(page title, page name, onLoad and page actions) is a normal reorderable section.

When an edit rebuilds the inspector, the open section, the scroll position and
the focused field are restored. You can continue editing the same parameter
without searching for the section again.

Every row in the left **Structure** tree starts with an exact type badge
(`div`, `tree`, `treeitem`, `chart`, `input`, ...). The badge is independent of
the friendly label, so the selected element type is always unambiguous. The
inspector repeats the exact type in a **Selected element** banner above the
sections.

An **Animation** section appears when the selected element's class carries an
`animation` declaration. It shows the class and the animation value, calls out
an `infinite` loop, and offers **Remove class** (stops it on this element) and
**Stop in CSS** (sets `animation: none` on the class, for every element using
it). Motion-graph elements also open their **Motion (animation panel)** section.

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

### Shadow, filter and blend effects

The **Effects** section uses compact cards instead of a single crowded row:

- **Box shadow**: presets (`none`, `soft`, `raised`, `crisp`, `inset`), X/Y/
  blur/spread spin boxes in `px`, an alpha spin box in `%`, an inset toggle, the
  project colour picker and a raw **CSS** field. `rgba()` alpha is preserved.
- **Filter**: blur (`px`), brightness/contrast/saturation/grayscale/sepia/
  invert (`%`) and hue (`°`) with presets and a raw CSS field. Unknown functions
  such as `drop-shadow()` or `url()` are kept in the raw value and preserved
  when a known control changes.
- **Backdrop filter**: glass, frost and tint presets with blur, saturation,
  brightness and contrast controls; the editor writes both the standard and
  `-webkit-` property.
- **Blend mode**: `mix-blend-mode`, from `normal` to `luminosity`.

The **Typography** section uses the same card for `text-shadow` (without spread
and inset). The raw CSS field is the escape hatch for multiple shadow layers or
advanced filter functions; a preset or a numeric control replaces a multi-layer
shadow with a single layer.

### Advanced style overlays

Less common properties are grouped behind buttons in the section they belong
to. Each button opens a draggable overlay; the overlay stays open while values
change and stays in sync with the CSS editor.

- **Background** → **Advanced background...**: a **Gradient builder** for
  linear/radial/conic gradients with three colour stops, a live preview and
  Apply/Clear, plus image, size, position, repeat, attachment, clip, origin and
  blend.
- **Layout** → **Advanced flex...**: a **Flex builder** for direction, wrap,
  justify, align, content and gap with a live preview, plus flex-wrap, order and
  grow/shrink/basis.
- **Grid** → **Advanced grid...**: a **Grid builder** for template columns/rows,
  auto flow, justify/align and gap with a live preview, plus the raw grid fields.
- **Typography** → **Advanced text...**: text-overflow, overflow-wrap,
  word-break, writing-mode and direction.
- **Size** → **Overflow and sizing...**: overflow, box-sizing, aspect-ratio,
  min-width/min-height and object-fit.
- **Effects** → **Transform and clip path...**: **Transform builder**
  (translate/scale/rotate/skew), **Clip-path builder** (circle, ellipse, inset,
  polygon), transform-origin and isolation.
- **Border** → **Per-side borders...**: one side or all sides with style, width
  and colour; **Border image...**: source, slice, width, outset and repeat;
  **Outline...**: outline style, width, colour and offset.
- **Interaction**: cursor, pointer-events, user-select, visibility, float,
  clear and scroll-behavior. These target the selected class rule. Cursor,
  pointer-events, user-select and visibility are inherited, so on a container
  they also affect descendants; use **Appearance → style expr** for one element
  only.

The overlay closes with its **×** button or **Escape**; selecting another
element closes it as well.

### Advanced CSS

Open **Appearance → CSS declarations → Edit...**. The overlay lists the
declarations of the selected class rule. Every row has a read-only property, an
editable value and a delete button. The add row accepts any valid CSS property,
including custom properties (`--accent: #4c5cf7`), with property and value
suggestions and validation. Values autocomplete common keywords and function
prefixes (`display`, `position`, `flex-*`, `grid-*`, `border-*`,
`linear-gradient(`, `clip-path`, `content`, `var(` and so on), plus every
project design token and custom property as `var(--name)`. Examples:

```css
overflow: hidden;
aspect-ratio: 16 / 9;
transform: rotate(2deg);
--accent: #4c5cf7;
```

With several elements selected, the overlay edits every selected class. Fields
with different current values show **(mixed)** and a tooltip listing the values;
entering a value applies it to all selected classes. Multi-class writes are sent
as one transaction, so an invalid selector or at-rule rejects the whole update
instead of leaving some classes changed. **Text mode** switches to a
declaration textarea (`property: value;` lines) with an **Apply text** button;
mixed values are shown as comments and are not overwritten unless typed.

The **Rule context** card adds pseudo-elements (`::before`, `::after`, `:hover`,
`:focus`, `:focus-visible`, `:active`, `:disabled`) and an at-rule field
(`@media`, `@supports`, …). For `::before`/`::after`, **Set content** writes the
`content` declaration. The context applies to every selected class. This is a
second view of the same CSS model used by the CSS editor: edits update `root.css`,
the generated source and the CSS editor immediately. The overlay edits class,
pseudo-element and at-rule rules, not inline effective values; use **style
expr** for reactive inline declarations. External CSS resources (`@import` and
network/file `url(...)`) are rejected by the compiler; use data URIs or local
relative assets. Vendor-prefixed properties are allowed; editing a property that
has a known compatibility entry updates its prefixed variants too.

Visual builders (gradient, transform, clip-path, flex, grid, per-side borders
and border-image) write their declarations to every selected element in one
multi-class transaction too, so a multi-selection edit cannot leave half the
classes updated.

## Control appearance

Every control except **table**, **graph** and **chart** shows a section named
after the control (**Button**, **Card**, **Dropdown menu**, ...) in the
Inspector. It contains four main drop-downs in this order: **Theme**,
**Structure**, **Style**, **Animation**, plus **Hover**, **Click**, **Focus** and
**State** selects under Animation for the interaction feedback. **Dark** and
**Light** are always available as the first two styles.

Each animation is a coherent set: the idle effect plus a visible hover reaction,
a click/press reaction, a focus ring and a checked/selected/open state reaction.
**None** disables every phase. The navigator preview plays the phases once on
load and then stays interactive, so the click feedback is visible without
leaving the editor. Open menus and popups keep their offset from the trigger;
labels wrap or ellipsize inside the control shape instead of spilling out.
Text-entry controls (**input**, **textarea**, **select**, **combobox**,
**date/time**, **number**, **color**, **file**) only offer **None** and
**Feedback**; their feedback never moves or resizes the field while typing.

| Selector | What it changes |
|---|---|
| Theme | a named structure + style + animation combination |
| Structure | shape and drawing for the kind (for a button: rectangular, pill, circular, square, wide) |
| Style | colour palette (border, text, fill, shadow); Dark and Light are always first |
| Animation | idle effect plus hover, click, focus and state feedback |

**Browse…** opens the Controls library navigator filtered to that control;
**Save theme…** stores the current combination and custom declarations in the
editor's favourite themes, which then appear in the **Theme** selector. Values
apply to the selected element only; **Settings…** continues to manage the project
default for every control of that kind. Text, font, border, events and alignment
stay in their own inspector sections.

**Insert → Controls library…** opens the navigator with an **Appearance** tab
(every control, three side-by-side lists for structures, styles and animations,
interactive preview, themes) and a **Components** tab for insertable presets.
The appearance lists are keyboard-navigable: Up/Down moves and selects the reached
option (the preview updates immediately), Left/Right switches list, Home/End jumps
to the first/last option, Enter re-applies the focused option. The catalog itself is `components/appearance.json` (format
`simpleui-appearance-library`): it lists structures, styles and animations for
every control, requires Dark and Light styles, and is validated by the server.
A missing or rejected file leaves the built-in catalog active. Reload it with
**Preferences  Reload component and appearance libraries**.

The Insert toolbar groups controls
under **Inputs**, **Toggles**, **Actions**, **Panels**, **Data**, **Feedback** and
**Advanced**; Advanced holds table, graph and chart. Favourite themes are stored in the
editor's local storage (`sui.appearance.themes.v1`), not in the `.sui` file.

See [Control appearance library](APPEARANCE_LIBRARY.md) for the catalog, the
button example and the workflow.

## Image assets

**Insert → Image assets…** opens the asset manager. It uploads an image and
keeps it in memory as a data URL, so the project stays self-contained and the
compiler's external-resource rule keeps holding. PNG, JPEG, GIF, WebP and AVIF
are accepted up to 4 MB; the server validates the file signature, not just the
extension, and SVG is rejected because it can carry active content.

Every uploaded asset appears as a card with a thumbnail, name and size:

| Action | Effect |
|---|---|
| Background | Writes `background-image: url(data:...)` on the selected element |
| Src | Sets `src` on the selected `img`, `video`, `audio` or `iframe` element |
| Copy | Copies the embedded data URL to the clipboard |

Uploads are session-scoped: they are not written to disk, and the data URL is
embedded directly in the SUI source when you apply it. Use a local relative path
instead of an asset when you want the file to live next to the project.

## Languages

**Preferences → Languages...** opens the translation editor for the
generated app. It is separate from **Editor preferences → Interface
language**, which translates the designer itself.

- The left column lists the locales; the **first** locale is the fallback. Click
  a locale to edit its keys.
- **Add language.** asks for a BCP 47 code (`it`, `en`, `pt-BR`, ...) and
  creates an empty catalog.
- **Add key.** adds a message. The **Plural** checkbox turns it into a CLDR
  plural block; **Add plural form...** adds `zero`, `two`, `few` or `many`.
  `one` and `other` are created by default and `other` is required.
- **Import JSON** and **Export JSON** use a plain object:
  `{ "it": { "save": "Salva", "items": { "one": "...", "other": "..." } } }`.
- The preview toolbar **Lang** selector reloads the preview with
  `?locale=xx` without saving the choice. In the app, `i18n.setLocale(code)`
  persists the choice in `localStorage` (`sui.locale`).
- Generated source gains `i18n "xx" { ... }` blocks. `t(key, vars)` and the
  `i18n.*` helpers are documented in the
  [API reference](API_REFERENCE.md#internationalization-i18n).
- Use `?locale=xx` on a normal page load to override the stored/browser
  language without changing it.

See `examples/i18n/` for a runnable Italian/English/Arabic demo with plurals,
RTL and the complete source.

## Component library

**Insert** is extended by `components/library.json` next to the project (or in
the working directory). The GUI reads it at startup, validates the whole file
and shows the accepted groups before the built-in ones. External tools can edit
the file and use **Preferences → Reload component library** without restarting
the editor.

A rejected file is reported as `Library rejected: ...` in the Insert menu and in
the status bar; the built-in catalog remains available. A library can add
presets over the standard element kinds, group them by control type, and rename
or re-group a built-in kind with `"override": true`. Every group must contain at
least one component.

**Insert → Controls library…** opens the navigator: filter by control type,
search, read the description, inspect the **interactive preview**, and insert the
selected component. The **Default** badge marks the fallback version of a kind.

If a project references a library component that is no longer in the file, the
editor shows a warning and rewrites that element to the default version of its
kind (explicit `default`, otherwise the first component of the kind), keeping the
element content. If the kind has no component at all, the warning says so and the
current styles are kept.

See [External component library](COMPONENT_LIBRARY.md) for the JSON schema, the
supported kinds, the CSS and attribute allow-lists, the size limits and the
compile check performed on every component.

## Preview select mode

Clicking an element in the preview selects it **and** runs the action bound to
it: a button with an `onClick` opens its modal, a link navigates, a form
control changes the application state. When the goal is only to inspect or
restyle that element, the **Select** button in the preview toolbar switches the
preview to select mode.

| Mode | Click in the preview |
|---|---|
| Normal | The element is selected and its action runs |
| Select (toolbar button) | The element is selected, its action does **not** run |

In select mode the click never reaches the application: `onClick` handlers,
link navigation, form submits and the native behaviour of form controls are
all suppressed, while the designer still picks the element, updates the tree
row, the inspector and the outline in the preview. The mode is a designer
setting: it is remembered between sessions and it is re-sent to the preview
every time the preview is rebuilt.

Use it to restyle a button without opening the modal it triggers, to inspect a
control that would otherwise submit a form, or to pick an element inside a
modal without dismissing it.

## Designer drag and drop

The designer's own drag and drop is always enabled. There is no setting,
property or build flag for it. The inspector's **Drag & drop** section configures
the generated application; it does not control moving components in the editor.

| Action | How |
|---|---|
| Insert a component | Drag a palette item into the preview or tree |
| Reorder a component | Drag a tree row above/below another row |
| Reparent a component | Drop a tree row inside a container row |
| Reorder toolbar sections | Drag the toolbar grip; the two bars sit side by side when space allows |
| Change table row/column size | Drag a table boundary grip |

The editor uses the internal element model as the source of truth. A successful
drop creates an undo snapshot, updates the tree, re-renders the preview and
regenerates the `.sui` source. The moved element keeps the selection: its tree
row stays selected, it stays outlined in the preview and the inspector keeps
showing it, so the relation between the tree, the preview and the source
survives the move.

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

### PDF viewer sample

Inserting a **PDF viewer** from the Media group creates a control that points
at `sample.pdf`, a small document embedded in the designer and served at
`/__gui__/sample.pdf`. The preview therefore shows a real PDF instead of an
error: an empty `src` is a compile error, so a control without a source would
break the whole page.

The relative `src` also works in the exported app: it looks for `sample.pdf`
next to `index.html`. Replace the `src` (Content section of the inspector) with
your own document URL, a state, or the URL returned by `files.documentUrl()`
before shipping. A desktop build embeds only the app assets, so there the
document must come from a real URL or from the desktop bridge.

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

## Dropdown and popover controls

**Controls > Dropdown menu** inserts a button with an anchored list of
`dropdownitem` children. Set the label from **Text**, the start/end alignment and
the preferred `placement` in the **Dropdown** section; add items from the tree.
Items support `onClick`, `disabled` and `separator`. The menu flips and clamps
itself inside the viewport.

**Controls > Popover** inserts a trigger with rich child content in an anchored
panel. Set the trigger label from **Text** and the preferred side in the
**Popover** section. It closes on outside clicks and Escape. Auto placement keeps
the panel inside the window.

**Controls > Drawer** inserts a side panel. Set the heading from **Text**; the
**Drawer** section controls the side (`left`/`right`), width, the `open` state,
an optional `onClose` function and the slide `duration`/`easing`. The state field
creates the state when it is missing.

### Animation settings

Dropdown, popover, combobox and command palette inspectors expose an
**animation** effect (`fade`, `scale`, `slide`, `none`), `duration`, `easing` and
`delay`. Drawer and carousel expose `duration`/`easing`; skeleton exposes the
shimmer `duration`. The same fields are available on the source attributes and
are previewed live in the designer.

### Toggle structure, theme and motion

Selecting a **checkbox**, **switch** or **radio** opens **structure & motion**.
The **structure** selector picks a shape preset, **theme** picks a color preset
and **animation** picks `auto`, `morph`, `fade`, `scale`, `slide` or `none`.
`duration`, `easing`, `delay` and `stagger` tune the transition. Shape and theme
are CSS-variable presets; the motion runtime interpolates every part between the
two states, so the switch thumb, track, check mark and borders morph instead of
jumping. These attributes can be bound to states and changed at runtime.

### Layout motion (FLIP)

Select any card, list item, panel or container and open **Motion (FLIP / enter /
exit)**. Set the effect (`layout`, `fade`, `scale`, `slide` or off), a stable
**key**, duration, easing, delay and whether to animate the exit. The runtime
matches elements by key across renders, FLIPs them to their new positions and
fades missing/new elements. Repeated list items and components must set
`motionKey` (for example `item.id` or the component parameter); `data-sui-id` is
not unique inside a list. Tables, trees, charts and toggles have their own
specialized motion sections.

On multi-page apps, `navigate(page)` captures the visible page before switching
and flushes shared keys after the new page renders. The same `motionKey` on two
pages produces a shared-element transition, including size and position. The
runtime measures the start and end rectangles of the card and its direct visual
parts, grows a content-free surface layer and translates each part in a straight
line to its measured final position; the content is clipped to the growing box
instead of being reflowed.

### Animation panel

Select an element and open **Motion > Open animation panel**. The panel contains
only the graph; parameters open in draggable overlay panels:

- **Layout & shared** in the toolbar opens the FLIP/list/page settings overlay.
  **Exit animation** opens the exit overlay: choose the destination state plus
  duration, easing and delay. **Spring easing** adds stiffness, damping and mass;
  spring overrides easing and duration controls its speed. **Preview exit** plays
  the transition to the exit state without removing the element; **Stop preview**
  cancels it. The runtime animates a fixed ghost clone when SUI removes the
  element.
- The panel is compact and has a drag handle on its bottom-right corner. The
  graph canvas centers and fits the diagram; the toolbar has **&#8722;**, **+** and
  **Fit** controls with a zoom percentage. Drag empty canvas space to pan and use
  the wheel to zoom around the pointer. The panel and overlays stay fully inside
  the viewport and can be dragged or resized.

- State blocks are intentionally small but grow taller as outgoing triggers are
  added, so their connection points stay readable. Drag a block to move it; the
  position is kept in the graph. Each block has **+** to add a link, a pencil
  icon to edit its appearance, and **×** to delete the state.
- **Create state** in the animation panel creates a new state block without a
  link. Move the mouse over the graph and click to place the block where there is
  enough free space, then use **Confirm position**. The new state is selected and
  can be edited with the pencil icon.
- Right-click a state for **Parameters**, **Duplicate state** and **Delete
  state**; the state overlay also has **Duplicate state**. Duplicating a state
  copies its CSS and parameters; adjust them before saving. The normal element
  **Duplicate** deep-clones the element, including `motion`, `motionKey`,
  duration/easing/delay and `motionGraph`.
- The **+** icon on a state creates a link from that state to an existing
  destination. The trigger chooser opens with **Cancel** and **Confirm**; the
  destination is preselected to the most recently created state. Confirm creates
  the link, closes the panel and draws it with the completed color. To return to
  the previous state, press **+** on the new state and choose the old state. The
  source point is always on the right side and the target point at the centre of
  the left side. An older link without a trigger can still be completed by
  double-clicking the source state or its red trigger row.
- A trigger belongs to the state that starts the transition. Every outgoing
  link is numbered at its starting point; the number is the trigger index.
  Hovering the number shows the trigger hint (for example `click`). Several
  links can leave the same state and point to different targets.
- A trigger can listen on this element or on another element: the chooser has
  **event source**. For an external source choose a DOM event or **external
  element state**, then one of the source element's motion-graph states. External
  state triggers fire when that source graph enters the selected state.
- State changes start immediately. An event that arrives during a tween is
  evaluated in the target state and starts a new tween from the current computed
  values (latest event wins); events are not queued. A new transition cancels the
  previous `after` link and timers, and interrupted tweens cannot leave stale
  intermediate properties. Timers start on entry; `after` links wait for the
  transition to finish.
- Links always start at the right side of the source and end at the centre of
  the target's left side, horizontally. Routing is orthogonal with a **15 px**
  minimum distance between boxes, between links, and between links and boxes.
  The first and last segment are perpendicular to the box edge and at least
  15 px long. Ports are assigned in vertical target order, so a fan-out nests
  without crossings; every change re-routes all links and tries several orders
  before choosing a route. Segments never cross a box, use right-angle corners
  and prefer the shortest path with the fewest segments. Links that go back to
  an earlier state are routed around the boxes. The route is redrawn live while
  a state box is dragged. If no crossing-free layout exists, links may cross:
  the later link is drawn over the earlier one with a small background gap. When
  a small automatic state move can restore a clean route, the editor moves that
  state and re-routes every link; placed positions are pinned so the layout
  stays stable. State blocks and links use theme-aware
  colors and stay readable in light and dark themes.
- Click and the other single-use DOM events can be used once per source state;
  variable and timer triggers can be repeated. The chooser disables already-used
  single events and validation rejects duplicates. For a reversible hover use
  `pointerenter` + `pointerleave` (or `mouseenter` + `mouseleave`);
  `mouseover`/`mouseout` bubble, so they also fire when the pointer crosses child
  elements and can repeat more than once.
- The chooser also exposes `keydown`, `keyup`, `input`, `submit`, `pointerdown`,
  `pointerup`, **wheel**, **swipe**, **long press**, **enters viewport**,
  **leaves viewport** and **custom** events. Keyboard and custom links can carry a
  key, mouse-button and Ctrl/Alt/Shift/Meta filters; when several links match the
  same state and event, the most specific filter wins. Custom events use a DOM
  event name entered in the chooser.
- Wheel links can filter `down` or `up`. Swipe links choose an axis and a minimum
  distance (default 40 px); long-press links choose a hold time (default 500 ms).
  Visibility links choose a ratio between 0 and 1 (`0.5` for enters, `0.2` for
  leaves) and an optional **fire once** flag. Gestures and viewport triggers are
  repeatable, so they are not in the single-use list. Gesture graphs are
  keyboard accessible: Arrow keys trigger swipe, PageUp/PageDown trigger wheel
  and holding Space/Enter triggers long press; the element gets a focus outline
  and `aria-keyshortcuts`.
- The trigger overlay has an optional **guard** field. It accepts SUI state
  names, `true`/`false`/`null`, numbers, strings, `!`, `&&`, `||`, comparisons and
  parentheses (for example `open && mode == "edit"`). The editor validates the
  expression before saving and the runtime never uses `eval`; an invalid guard
  blocks the transition.
- Transitions use the browser animation engine with computed start and end
  values, so `transform`, filters, box shadows, clip paths, gradients, colors,
  lengths and custom properties animate where the browser supports them;
  non-animatable values change discretely. Final values are applied at the end.
- A **variable** trigger lists the available SUI states and asks for the value
  that activates it (`true`/`false`, text or number). A **timer** trigger asks
  for a duration and starts the transition automatically after that time in the
  source state.
- An **after transition** trigger starts automatically when the transition into
  the source state finishes. Chain several `after` links (A → B → C) to build
  multi-step animations. Each transition exposes `duration`, an **easing** combo
  box with the standard CSS timing functions and common `cubic-bezier` presets,
  and `delay`.
- The link editor also exposes **iterations** (a number or `infinite`),
  **direction** (`normal`, `reverse`, `alternate`, `alternate-reverse`),
  **playback rate**, a **spring easing** toggle with stiffness/damping/mass, and
  an **emit event** name with **emit when** (`start`, `end`, `cancel`). Spring
  overrides the easing curve; duration controls its speed. Alternate loops settle
  on the target state, so use an odd iteration count to end on the target side.
- A link can emit a custom event; the runtime also dispatches `sui:motionstate`,
  `sui:motionstart`, `sui:motionend` and `sui:motioncancel`. Another graph can
  listen to these with a **custom** trigger; choose the source element and use its
  `motionKey` for a stable reference.
- The trigger overlay arranges the timing and spring fields in a compact
  two-column grid. Duration, delay and timer are spin boxes in milliseconds;
  playback rate shows `×`; spring stiffness/damping/mass show `N/m`, `N·s/m` and
  `kg`. **OK** closes the overlay after the live edits are committed.
- The event source picker labels each element with its name, its `motionKey` or
  `data-sui-id`, and its tree path. A **Source code** line repeats the exact
  identifier to paste into the source. Set a layout key on the source for a
  stable readable code; text-only nodes are omitted from the picker.
- **Preview** plays the selected link with its delay, spring, iterations,
  direction and playback rate. **Stop preview** cancels a running preview; an
  infinite preview is also stopped when another state, link or panel is opened.
- Infinite loops run until another event cancels them and never fire `after`
  links. The next transition starts from the current computed values.
- Double-click a state to open its overlay: name, CSS and the list of active
  triggers with their destination. There are no incoming-trigger fields. When
  the state has an untriggered outgoing link, the trigger chooser opens directly:
  **Cancel** returns to the state overlay without changing the link,
  **Confirm** applies the trigger, closes the overlay, restores the graph and
  draws the link with the normal completed color. **Add trigger** uses the
  same chooser for a new link. Double-click an arrow to edit its trigger,
  duration, easing and delay; an arrow without a trigger opens the chooser for
  that link. Overlays are draggable and stay above the graph; opening them does
  not minimize the animation panel.
- A single click selects the state and previews its parameters without leaving
  the panel.
- The pencil icon on a state block enters graphic editing: the animation panel
  is hidden, the element is isolated and centered in the preview, and only the
  edit bar appears with **Save state**, **Cancel** and **Close**. Save captures
  the preview style, associates it with the state and reopens the panel; Cancel
  discards the changes and reopens the panel; Close discards them, closes the
  animation panel and returns to the normal editor.
- Saving rejects an empty trigger, a variable trigger without a variable, two
  states with identical parameters, or the same single-use event twice from one
  state. The error is shown in the editor notification and the state stays open
  for correction.
- Opening the panel selects the element's current or initial state. Clicking in
  the preview minimizes the panel; the **–** button, drag, resize and close
  controls remain available. A floating **Animation panel** button restores a
  closed panel.
- **Set initial** marks the state applied on load. **Preview** plays the
  selected state or link in the preview.
- **Remove states** deletes the state graph but keeps Layout & shared.

The editor writes a literal `motionGraph` JSON attribute plus the classic motion
attributes from Layout & shared. The runtime keeps the current state on the
element and uses the browser animation engine with computed start/end values; it
honors reduced motion. The panel is per element; Layout & shared still uses
`motionKey` for shared pages and lists.

**Controls > Split panes** inserts two `pane` children and a draggable divider.
Set `direction`, `initial` and `min` in the **Splitter** section. A splitter must
keep exactly two panes.

**Controls > File dropzone** inserts a drop area that also opens the picker. The
**Dropzone** section sets the target state (`dropFiles`), an optional separate
`pickFiles` state, `accept`, `multiple`, `directory`, `pickMode` and the hint.

## Content and feedback controls

The **Content** palette group provides the **Rich text editor**, timeline,
empty state, skeleton, carousel and stepper; the command palette is under
**Overlay**. The editor inspector exposes the runtime **name**, **Mode**
(*Rich text editor*, *HTML editor* or *Markup editor*), theme (`dark`/`light`),
**Toolbar items...**, **Hyphenation languages...** and a single **Read only**
switch. HTML and Markup profiles restrict the toolbar, menus, file types and
settings; they remove header/footer and use Document/Page settings without it.
Header, footer and
the toolbar/menu/context/status toggles are not offered here; those attributes
remain available in hand-written sources. Controls carry short tooltips instead
of explanatory text. Height and width are edited in the generic **Size**
section: the height sizes the content area while the status bar stays pinned to
the bottom border. The element text is not rendered above the menu; it only
feeds the `{title}` token and the fallback runtime name. **Toolbar items...**
opens a searchable multi-select overlay
built from the runtime toolbar commands; **Default**, **All items** and **None**
are available as presets, and `None` leaves only the theme button. **Source code**
(`sourceMode`) is an opt-in item under **Document**, off in every preset: it opens
the HTML/Markdown source editor in the HTML and Markup profiles. With
**Read only** enabled no menu or toolbar is shown and the content cannot be
changed; the control becomes a real viewer for rich text, HTML or Markdown.
The page background is white by default and can be changed in Document settings
and Page settings; dark backgrounds automatically use readable text. Page
settings are keyed by number: explicit page breaks carry those backgrounds,
sizes, margins and header/footer strings into PDF, DOCX, ODT and RTF exports.
Per-page header/footer text is set in the JSON `pageOverrides[n].header` /
`.footer`, not in the settings panels; imports read it from every section or
master. In HTML and Markup modes the status bar also has a **Source** toggle: it opens a source
editor where you can type, paste, apply or revert HTML/Markdown code, then return
to the visual editor. Read-only controls never expose source editing.
**Hyphenation languages...** opens
a searchable multi-select overlay that writes `hyphenLangs` and embeds only the
chosen TeX dictionaries in `app.js`; the **All languages**, **European** and
**None** presets make it easy to trade hyphenation coverage for a much smaller
bundle. Binding `value` to a state stores the JSON document; use the
`richtext.*` API or the toolbar at runtime.

- **Timeline** inserts three `timelineitem` events. Each item exposes **date**
  and **status** (pending, active, done, error).
- **Empty state** inserts an icon, title, hint and body. The **Empty state**
  section sets the icon, hint and compact mode.
- **Skeleton loader** inserts three animated lines. Set `lines`, `width`,
  `height`, `gap`, `circle` and `animated`.
- **Carousel** inserts slides with previous/next buttons and dots. Bind `value`
  to an int state; the **Carousel** section sets height, autoplay, interval,
  loop, controls and indicators.
- **Stepper** inserts three steps. Bind `value` to an int state; each step
  exposes a description and an optional icon.
- **Command palette** inserts an overlay bound to an `open` state with a `k`
  hotkey. Each `command` exposes a shortcut and its `onClick` handler. Without
  `open`, the same control stays inline.

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
| popup placement | preferred side (`auto`, `bottom`, `top`, `left`, `right`) |
| animation | popup open effect (`fade`, `scale`, `slide`, `none`) |
| duration / easing / delay | open effect timing |

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

Insert **Tree** from **Controls**. A new tree starts with a parent item and a
child item. **Tree item** is not offered in the Insert menu or palette: items
are created and edited from **Tree data** or the tree context menu. The element
list shows only the **tree** row, not its internal items.

The **Tree** inspector exposes only:

| Field | Meaning |
|---|---|
| settings | opens **Tree settings** |
| data | opens **Edit tree data** |

**Tree settings** opens the component panel with these tabs:

- **General** - `Title` (the visible tree title), `Runtime name` (the target used
  by `tree.*`; not the visible title) and behavior switches: indent guides,
  truncate labels, hierarchical checkboxes, multi-select, search box, drag and
  drop, context menu, virtualize offscreen rows, persist selection and edit mode.
- **Bindings** - optional state bindings: value state, selection state (shown for
  multi-select), load key state, edit depth and rename key/value. State fields suggest declared project states from a datalist; leave
  them empty when the app does not use the corresponding behavior.
- **Columns** - `Show columns` and a column editor where rows can be added,
  renamed, resized, removed and reordered. Column keys are free-form; `size`,
  `date` and `status` have convenience item attributes, other keys use the item
  `cells` values. Columns stay hidden until **Show columns** is on. Child sorting
  is in the same tab.
- **Colors** - text, title, row, hover, selected, toggle, border, guide, badge,
  metadata and checkbox colors.
- **Title** - title size, weight, font, spacing, alignment, top/bottom position
  and padding. These values apply to the **Title** field in General.
- **Shape** - indent, row padding/height, radius, border, toggle size, font
  size/weight, font step, letter spacing and font family, plus the curated node
  icons and branch symbols: **Item icon (no children)**, **Parent icon
  (closed)**, **Parent icon (open)**, **Closed branch symbol** and **Open branch
  symbol**. **Font step** reduces the font size for each level below the root;
  default `0`. Icon fields open a categorized overlay with only the relevant
  families: leaf icons for files/documents, folder/structure icons for parents,
  disclosure arrows for branches and editing tools for the rename button; an
  **Add from file…** action still accepts `.txt`, `.json` or `.csv` lists.
- **Extra** - two sections: **Edit button** (visibility, icon, background, text
  color, font, size, border, radius, padding and hover colors) and **Child
  counts** (visibility, background, text color, font, size, weight, padding,
  radius, border and minimum width). Both render next to the node name.
- **Tooltip** - custom tooltip switch plus background, text, border, radius,
  padding, max width, font size, delay and placement.

The **Style** selector at the top of Tree settings contains the shared styles
plus the 16 tree themes (Default, Compact, Comfortable, Dense, Bordered, Soft
rounded, Minimal, Neon, Forest, Amber, Ocean, High contrast, Outline, Mono,
Candy, Cards). Selecting a style previews it; **Apply** writes it to the tree and
updates the editor preview. Tree settings have no separate **Hover** tab: hover
colors are in **Colors**. Named presets are available in the same panel.

**Edit tree data** opens a dedicated editor with one row per item. The row shows
the label, an optional stable key, icon, badge, metadata, size/date/status grid
cells, open/check/edit state and compact actions: add child, add sibling, move
up, move down and delete. The toolbar adds root items and imports/exports JSON or
XML. **Apply** writes the items into the tree; Cancel leaves the model unchanged.
The same actions are available from the tree context menu (**Edit tree data**,
**Tree settings**, **Add root item**, **Import JSON/XML**, **Export JSON/XML**).

Selecting a node in the live preview always selects the **whole tree**: a
tree item is never selected on its own and the inspector never opens a
per-item panel. To edit the tree's structure, right-click a node in the
preview. The context menu offers **Add child item**, **Add item below**,
**Edit item label...**, **Duplicate**, **Move up**, **Move down**,
**Indent (under previous)**, **Outdent** and **Delete**. Right-clicking empty
tree space (or the tree itself in the element list) offers **Add root item**,
**Edit tree data**, **Tree settings** and the JSON/XML import/export actions.

Item-only fields (key, initially open, node style preset, rename icon,
icon/open icon, badge, metadata, checked state, lazy children, size/date/status
and tree-grid cells) remain part of the generated `treeitem` source and are
edited from **Edit tree data** or directly in the SUI. At runtime the same
structure is available from the tree context menu when **Context menu** is on:
add child/item, rename, duplicate, move, indent/outdent, delete, expand/collapse
and copy key.

`onChange`, `onRename`, `onCheck`, `onMove`, `onLoad`, `onContextMenu` and
`onTreeChange` are
available in **Tree settings > Events** (the generic inspector Events panel is
intentionally hidden for trees). `onTreeChange` runs whenever the tree data
changes: add/delete, move, indent/outdent, import/clear, rename, checkbox, lazy
children or grid cells. Use `tree.lastChange(name)` to read the operation, key,
parent and index. The generated source keeps nested `treeitem` children and their keys/open/rename flags.

The tree **style preset** menu writes the component's visual CSS into the tree
element's **style expr** field (in **Appearance**). It offers **Default**,
**Compact**, **Comfortable**, **Dense**, **Bordered**, **Soft rounded**,
**Minimal**, **Neon**, **Forest**, **Amber**, **Ocean**, **High contrast**,
**Outline**, **Mono**, **Candy** and **Cards**. Editing the `cssSafe("...")`
expression directly gives full control and switches the menu to **Custom**.

A tree item can still carry its own **node style preset** (`Inherit`, `Muted`,
`Accent`, `Success`, `Warning`, `Danger`, `Folder`, `Compact row` or `Custom`)
through `styleExpr` in the SUI source. Node presets write row-level
`--sui-tree-*` variables into that item's style expression, so they override
only that node.

The tree **rename icon** field changes the pencil glyph; **show rename icon**
hides it. A tree item can inherit, show or hide the icon per node through the
generated `renameIcon` / `showRenameIcon` attributes. Hiding the button does not
disable F2 or double-click rename.

The **tree name** field makes the tree a runtime target. The **Edit tree data**
panel imports or exports the tree items in the design source: import replaces
the selected tree's `treeitem` children, export downloads the current model.
Design-time import does not store data at runtime; use the core `tree` module
(`tree.importJson`, `tree.importXml`, `tree.importJsonFile`,
`tree.importXmlFile`, `tree.exportJson`, `tree.exportXml`,
`tree.downloadJson`, `tree.downloadXml`) for application data. The example
combines a `pickFiles` button with the file helpers and a download button. See
[`STDLIB.md`](STDLIB.md#tree-module-tree-core) for the JSON/XML schema.

The preview supports row selection, multi-selection, disclosure toggles, guide
lines, icons, badges, metadata, tree-grid columns, search/filter, hierarchical
tri-state checkboxes, drag and drop, a right-click menu, arrow-key navigation,
Home/End and inline rename (double-click, F2 or pencil; Enter/Escape). Expansion,
checked state and renamed labels are restored after a state-driven re-render;
search terms and selection anchors are retained for the page session. The
runtime uses `localStorage` for open/checked state and local label overrides;
labels, keys, icons, badges, metadata and grid cells are escaped and rendered as
text.

See [API reference: tree](API_REFERENCE.md#tree) and the
[tree example](../examples/tree/README.md).

## Table layout control

**Layout > Table layout** inserts a structural HTML table, distinct from the
data **Table** control. It has no data name, sorting, filtering, editing or
persistence: it only arranges other elements in an explicit row/column grid.

The **Table layout** inspector section controls:

- **preset** (2 x 2, sidebar + content, header + 2 columns, 3 x 2 dashboard);
- **rows** (1-50) and **columns** (1-12);
- **Structure table settings**, which opens a panel with one input per column
  width and per row height; the two lists scroll independently so large grids
  stay usable, and an empty field means `auto`. Opening the panel aligns the
  rows/columns counters with the actual grid; editing a size beyond the declared
  rows/columns raises the matching counter;
- **column gap** and **row gap**;
- **empty cell height** so empty cells stay visible and droppable;
- **stack on small screens** (responsive stacking below 640px);
- **semantic table** (`role="table"` instead of `role="presentation"`) and an
  **aria label**.

The table **width** / **height** are edited in the generic **Size** section, and
the table frame in the generic **Border** section. Select a cell to edit its own
border in the same **Border** section. The global per-cell border
(`borderStyle`/`borderWidth`/`borderColor`) remains source-level.

A new layout table starts as a single empty cell. Use **Split vertical** /
**Split horizontal** in the inspector to halve the selected cells; Ctrl+click
adds cells to the selection and **Merge selected** joins cells only when the
first two share a full side; that side sets the merge direction and every
additional cell must continue the same direction. Changing **rows** appends/removes complete rows at the bottom (one cell per
column); changing **columns** appends/removes cells on the right. When a
reduction would remove cells that contain elements, a dialog asks whether to
**remove** that content or **move** it into the nearest remaining cell; Cancel
keeps the grid unchanged. The **preset** list is shown only while the table is
empty. Drag the thin
vertical/horizontal split handles in the preview to resize columns or rows; the
matching sizes are shown and editable in **Structure table settings**. Click a
split handle without dragging to select the split; Ctrl+click adds more splits
and **Delete split** merges the adjacent cells while keeping their content.
If a split would make a column narrower than 40px or a row shorter than 24px, a
dialog asks whether to continue.
Solid split segments separate two cells; dashed segments are only a
prolongation across a merged cell. Clicking elsewhere clears the cell or split
selection; Ctrl+click is the only way to add a second cell.

Each direct child fills the next free cell, row by row. Click an empty cell in
the preview to select it and see its row/column; the selected cell is the
insertion container for the next palette/Insert element. **Alt+click** an
occupied cell, click its always-visible bottom-left triangle, or click the
content row in the element tree to work with its content; clicking a child selects that child
and clears the cell selection. Clicking an edit box clears any cell selection;
selecting a cell or another element blurs the edit box.
The element tree follows the selection: the table row when the table is
selected, the cell row when a cell is selected (Ctrl+click keeps the other
selected cells highlighted), and the content row for an element inside a cell.
The green triangle just outside the table's top-left corner selects the whole
table; clicking inside the first cell selects that cell or its content.
Select a cell to open the **Layout cell** section and set `column span` /
`row span` to merge it with the cells on its right or below; **Reset to 1 x 1**
splits it again. Increasing a span merges the fully covered cells and appends
their content; a partial overlap is rejected with a message. Reducing a span
keeps the content in the cell and recreates the released area as empty cells. Padding, background, border, effects and size are edited in the
generic inspector sections, which target the cell directly. Alignment belongs
to the elements inside the cell: select them and use their **Alignment** section.
The cell stays a real table cell, so vertical alignment uses `vertical-align`
and horizontal alignment uses auto margins; **Clear** restores the top-left
position. With a cell selected, the inspector hides **Position**, **Size**,
**Responsive**, **Layout**, **Grid** and **Bindings**, and **Box model** shows
only **padding** because margins do not apply to table cells. **Tooltip**,
**Events** and **Drag & drop** work on the `<td>`: they fire on empty cell areas,
while an event on a child component takes precedence. With several cells
selected, padding, background, border and the other generic CSS properties are
applied to every selected cell.

Arrow keys move selection between empty cells. In the designer preview responsive stacking is temporarily disabled so the grid
and its split handles stay visible; the compiled app still stacks below 640px.
Aligned cells become flex-column when stacked, so their horizontal/vertical
alignment still applies; other cells stay block. Cells are outlined, including empty and merged cells, so
the grid is visible even without application borders. Wrap several elements in a
`col`, `div` or `card` to keep them in the same cell. Use the data **Table**
control when you need `table.display`, `table.importText`, sorting, filtering or
persistent edits.

See [API reference: table layout](API_REFERENCE.md#table-layout) and the
[table layout example](../examples/table_layout/README.md).

## Table control

Insert **Table** from **Controls**: it creates a workbook with one sheet and a
**+** tab by default. The elements panel shows a single **table** node for the
whole workbook; the tabs, scroll wrapper and cells are internal. Clicking that
node selects the visible sheet and shows the resize handles around the table
control. The inspector starts with the **Structure** buttons (`+ Col`, `− Col`,
`+ Row`, `− Row`). Cell, row and column selection happens directly in the
preview: click a cell, use the outer gutters for rows and columns, or the
top-left corner to select the whole table.

The inspector keeps the **Structure** buttons and adds **+ Table** / **− Table**
for multiple tables. The **Table settings...** button opens a narrower overlay
with every parameter group in tabs; controls inside each tab are laid out in two
columns with the same compact sizing as the rest of the inspector:

**Appearance**, **Borders**, **Header**, **Caption**, **Behavior**, **Data**,
**Columns**, **Data view**, **Events** and **Tooltip**. Every field in the overlay updates the
preview immediately; the overlay stays open while editing.

The **− Table** button and the sheet selector appear only when more than one
table exists. **+ Table** wraps a lone table in a bottom sheet-tab workbook (or
adds a new sheet to the existing workbook) and selects the new table. A new sheet
copies the source table's appearance and behavior (sortable, filterable, events),
keeps its own generic header (`A`, `B`, `C`) with generic text columns and starts
with two empty body rows. The new sheet is always editable, even when the source
table was read-only or used numeric columns. The preview keeps the selected sheet
while settings are edited.

The **+** tab in the designer preview is always visible and adds a real sheet to
the workbook model, not a temporary runtime table, so it survives preview
rebuilds and is written to the `.sui` source. In a compiled app the **+** is
shown only when the workbook has an editable sheet. A workbook with a single
sheet shows no sheet tab, only the **+** in the designer (and no tab strip at all
at runtime when the workbook is read-only); the sheet tabs appear from the second
sheet on. Double-click a sheet tab, or use its context menu and **Rename
workbook**, to change the workbook name written on the tab. Each tab also has a
small **×**; clicking it, or choosing **Delete workbook** from the tab context
menu, removes that workbook from the model and the `.sui` source. In a compiled
app the **×** is added by the runtime to the tabs of editable sheets and removes
that sheet from the runtime workbook for the page session; the runtime tab
context menu offers the same **Delete workbook** action. The table control fills its host and has a **640 × 300 px**
minimum; when the host has no explicit size, that minimum is the default. The
workbook sheet area takes all the space inside the control border after the
title/gear row and the sheet tab strip, and it follows every control resize.
Resizing the control, not an individual sheet, is what changes the shared sheet
size. Workbook sheets without an explicit `selection` mode are editable; a click
on a cell, row, column or sheet tab keeps the inspector on the table settings. A
sheet that asks for row/column/cell selection stays read-only so the selection
interaction keeps working. Clicking a sheet tab only switches the visible sheet:
it does not draw a selection box or resize handles. Clicking the **table** node
in the elements panel outlines the whole control (title, gear, sheet area and
tabs) and shows the handles that resize it. If the columns are wider than the
sheet area, a horizontal scrollbar appears at the bottom of the sheet area, just
above the tabs; when rows exceed the height, only the body scrolls and the header
row stays fixed.

New tables are named automatically (`table1`, `table2`, ...), so their runtime
settings gear and `table.*` helpers are immediately available. The gear opens
the Excel-like panel described in **API reference → Runtime table settings**:
header sorting, multi-column sort levels and a funnel filter button in each
header that opens a searchable checklist of column values.

### Table selection

- Click a cell to select it.
- Click in the gutter just **above a column header** (outside the table) to select the column.
- Click in the gutter just **left of a row** (outside the table) to select the row.
- Hovering those gutters shows a thin blue strip **straddling the cell border**
  (half in the control padding, half over the cell edge). It is exactly the
  clickable zone, and it never overlaps the control resize handles. The strips
  live outside the scrolling sheet content.
- The small blue **triangle in the top-left corner of the first cell** selects
  all cells (all rows and columns together), not the outer control.
- Selecting the outer table control never steals a row/column gutter click: the
  hovered strip wins over a nearby cell.
- Clicking a sheet tab switches the workbook without drawing an extent box or
  resize handles; clicking the **table** node in the elements panel selects the
  whole control and shows its resize handles.
- New sheets start with a labelled header row and two empty body rows.
- **Shift+click** extends a cell/row/column range from the previous anchor.
- **Ctrl/Command+click** toggles individual cells, rows or columns.
- Drag in the left gutter to add contiguous rows; drag in the top gutter to add
  contiguous columns; drag across cells to select a rectangular area.
- Selecting the table itself (element list or **Table** selector) outlines the
  whole control. The resize handles adjust the control; rows and columns that
  exceed the sheet viewport produce the internal scrollbars.
- A click on a sheet tab switches to that sheet and updates the table settings
  without drawing an extent box or handles.

Table settings groups:

- **Appearance** - **theme** (Inherit global, or a Light/Dark variant from the
  same families as the general project themes), cell padding, horizontal/vertical
  alignment, zebra rows, hover row and the scroll/scale controls. **Inherit
  global** leaves the table on the active project theme; a forced theme writes
  that variant's `--sui-table-*` tokens on this table only. **Set default**
  stores the current choice for newly inserted tables. Zebra and hover colors
  are applied as a subtle tint so they stay readable on dark and light themes.
  **Scroll** wraps
  the table in a `tscroll` container with an optional max height, sticky header
  and custom scrollbar colors; **scale** zooms the whole table. In a workbook,
  **control padding** (in the Scroll & scale group, default 10 px) insets the
  viewport from the control border and provides the space for the row/column
  selection strips.
- **Borders** - preset, cell grid, outer border and radius.
- **Header** - header row mode, background/text color, weight, uppercase and
  sticky header.
- **Caption** - optional caption, size, weight, alignment and color. At runtime
  the caption is shown in the title row just above the table, on the same line
  as the settings gear.
- **Behavior** - **edit mode** and **cell text** mode. The edit mode select has
  three states: **non-editable** (the user can read, sort, filter and select but
  cannot edit cells, import data or change structure), **editable** (edits,
  imports and runtime-created data last for the page session and the source data
  returns on reload) and **editable persistent** (edits, imported data and
  runtime-created workbook sheets are stored in `localStorage` and survive
  reload/restart; if the table starts empty the data created by the user is
  kept). In persistent mode the runtime column widths, row heights, header
  labels, sort/filter/page state and cell text mode are saved too. The runtime
  table settings panel shows **Reset saved data** to discard them and restore the
  design-time table. The cell text mode is **truncate** (default; the column
  width never grows with the text, a single line with an ellipsis and the full
  text in a hover overlay) or **wrap** (the column width stays fixed, the text
  wraps and the row height grows). Runtime cell and
  header edits stay in the table store across re-renders and sheet switches for
  the page session; typing never rebuilds the table, the value is finalized on
  blur or after a short pause and change events run on blur. Cell input events
  stay inside the table, so ancestor `onInput` handlers are not triggered.
  Editable cells can be typed directly in the preview: click the cell to focus
  it, then type. In the designer, the committed value is also written back to the
  model and `.sui` source on blur or after the idle pause, without rebuilding the
  preview, so it survives a later resize or preview reload. Header labels are
  editable too: a click places the caret in the label, while sorting uses the
  dedicated header sort button.
- **Data** - runtime **table name** used by `table.display`, `table.data`,
  `table.json` and `table.exportText`.
- **Columns** - declarative `columns`, labels, widths, alignment, cell types,
  formats, select options and **Use first row** to derive keys/labels from the
  designed header. Columns remain optional; without them the designed header and
  cells are used as-is.
- **Data view** - `sortable`, `filterable`, `multi-select`, `virtual rows`
  (reserved), `show add table button` (runtime **+** tab), `runtime structure
  editing` (row/column context menu), `page size` and `selection` mode
  (row/column/cell). In `column` selection mode a header click selects the
  column; **Alt+click** sorts it when `sortable` is on. In editable tables a
  header click edits the label and sorting uses only the header sort button.

New table events are available in the overlay **Events** tab:
`onCellClick`, `onRowClick`, `onCellChange`, `onRowChange`, `onTableChange`,
`onRowSelect` and `onSort`. Details of the last change are available from
`table.lastChange(name)`.

Editable cells commit on blur/Enter. `columnTypes` validates `number`, `date`,
`boolean` and `select` values; invalid commits are restored and marked with
`aria-invalid`. `columnFormats` renders `number`, `currency:EUR`, `percent` and
`date` values while keeping the raw value for sorting, editing and export.

### Excel-like sheet tabs

For a workbook with several tables, use the normal **Tabs** control:

```sui
tabs "Workbook" position: "bottom" variant: "sheets" {
	tab "Users" {
		table name: "users" columns: "id,name" columnLabels: "ID,Name" sortable: true {
		}
	}
	tab "Orders" {
		table name: "orders" columns: "id,total" columnLabels: "ID,Total" sortable: true {
		}
	}
}
```

The **Tabs** inspector offers **tab position** (Top / Bottom) and **style**
(Standard / Sheet tabs). Bottom + Sheet tabs produce a compact Excel-like strip
under the panels. Each table keeps its own `name`, data, sort/filter/page state
and events.

See [API reference: tables](API_REFERENCE.md#tables-and-untrusted-data) and the
[sheet tabs example](../examples/table_sheets/README.md).

## Chart control

Insert **Chart** from **Controls**. A new chart starts empty as a bar chart: data
controls are containers, so no demo dataset is written unless you load it
explicitly with **Insert sample** or bind a state/expression. The chart fills its
container (`width: "100%"`, `height: "100%"`) and uses `overflow: "scroll"`. New
charts keep the table hidden until **Mode** is set to `table` or `both`.

The dedicated **Chart** inspector is ordered for quick access:

| Field | Meaning |
|---|---|
| name | runtime target used by `chart.*`; it must be unique among charts and graphs |
| type | bar, horizontal bar, stacked bar, grouped bar, line, step, smooth, area, combo, pie, donut, scatter, histogram, waterfall, radar, gauge, funnel, heatmap, candlestick, treemap, box plot, bullet, dumbbell or Gantt; the source attribute is `chartType` because `type` is reserved. Block diagrams use the separate **Graph** control |
| Mode | chart only, accessible data table only, or both; formerly **table view** and available for every type |
| settings | opens the unified **Chart settings** panel (appearance, points, tooltip, table, interactions, events, element hints and data fields) |
| bins | histogram bucket count (1-100) |
| min / max | gauge range (default 0-100) |
| edit mode | non-editable, editable for the page session, or editable persistent across reload/restart |
| show grid / show legend / show values | chart presentation toggles |
| preset | 14 classic color presets plus one preset for every theme family (Light/Dark) that write CSS variables into `style expr` |
| edit data | button that opens the movable, resizable chart-data overlay at the bottom of the inspector. Data columns and numeric series are auto-detected |

Runtime names must be unique within their family (`richtext`, `table`, `tree`,
`chart`/`graph`, `modal`). The designer runs the compiler check live, so a
duplicate name appears in **Problems** and `simpleui check` rejects it; names
from different families may repeat.

**Point size** and **point width** live in **Chart settings → Points**.
**Element hints** also appears in **Chart settings → Data**.
The separate **Chart attributes** button has been removed; unknown attributes
already present in a source are still preserved when the designer rewrites it.
There are no X/Y inspector fields: edit the data itself and let the chart detect
the label and numeric columns.
`overflow` and `initial zoom` are no longer inspector fields: normal scrollbars
appear when the frame is smaller than the chart.

The **Chart** inspector is type-aware: irrelevant fields are hidden when the
selected type cannot use them.

- **bins** only for histogram;
- **min/max** only for gauge;
- **show grid** hidden for pie, donut, radar, gauge, funnel, heatmap, treemap, bullet and state;
- **show legend** hidden for gauge, funnel and state;
- **show values** hidden for pie, donut, scatter, gauge, radar and state.

Changing **type** updates the visible fields immediately. If the current data is
not compatible with the selected type, a confirmation dialog warns that the
existing data will be ignored and replaced by a sample dataset suited to the new
type; cancelling restores the previous type. An empty chart stays empty when the
type changes: no sample is injected implicitly. Compatible data is kept. If
`chartType` is an expression, all fields stay visible and compatibility is
checked at runtime.

Gantt uses the same data overlay but reads richer rows: `task`, `start`, `end`
(or `duration`), `progress`, `milestone`, `group`, `parent`, `dependsOn`,
`dependencyType`, `lag`, `constraint`, `constraintDate`, `resource`, `status`,
`color`, `cost`, `actualStart`, `actualEnd`, `remaining`, `work`, `workDone` and
`resources` (an array or a comma/semicolon list).
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
immediately recalculates effective duration. **Mode** is available for
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

With **Mode** set to **both**, every chart docks a fixed-height table at
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
with explicit `x`/`y` coordinates available for manual placement. Use the
dedicated **Graph** control: `chartType: "state"` (and the `statemachine`
alias) is rejected with a diagnostic; the chart type menu does not offer the
block diagram.

The gear button in the chart header opens a tabbed **Settings** panel: **General**
(chart title, shape border/highlight colors, theme, show-anchors switch plus one
row of GRID/SHAPES/LINKS/TAGS buttons, save-as-default and reset),
**Versions** (pending changes, commit and version history) and **Data**
(import plus JSON/CSV/SVG/PNG export). Closing a GRID/SHAPES/LINKS/TAGS
sub-panel restores the General panel. The SHAPES defaults use two aligned
columns with matching row heights. The graph itself keeps only the diagram,
the shapes/links table and the add controls. **Export JSON** embeds the full
shapes, links and settings so re-importing it reconstructs the diagram exactly.
The **Edit mode** select in the Interactions tab controls the diagram:
**non-editable** removes selection, drag/resize and editing handles while keeping
pan/zoom; **editable** keeps them for the page session; **persistent** also saves
geometry/style changes across reload/restart. In non-editable mode every
interaction is queued by the chart; SUI consumes the queue with
`chart.pollEvents` / `chart.nextEvent` and decides which external function to
call. The queue item contains `event`, `kind`, `part`,
`id`, `label`, `tag` and pointer/modifier state. Direct callbacks are not
supported: the event list is the only interaction model. Shape ids are mandatory
and unique; the designer reports missing or duplicate ids as an import error.
**Hover hints** in the Interactions tab maps to `stateHints`; turn it off to hide
all tooltips. The **HINTS...** button opens the global tooltip style panel
with two groups: switches/colors on the left and size/spacing sliders on the
right. It also has a **Text align** option (left, center or right) for hints
whose text spans multiple lines. The background swatch includes its own opacity slider, so no separate
transparent button is needed; it writes the `stateHint*` chart attributes. Each
shape/link also accepts a `hint` field, editable from the table **Hint** column
or the context menu **Hint settings...**; use `|` for a line break and `\|` for a
literal pipe.
**Element hints** in **Chart settings → Data** lists every state and link found in
JSON, CSV or XML chart data (including a literal document held by a simple state
such as `data`) and edits their tag and hint text in single-line fields (`|`
breaks a line, `\|` a literal pipe). It writes the data back; CSV needs a header
row with `id`/`label` or `from`/`to` plus `tag`/`hint` columns, which are added
when missing. Computed expressions still use the data editor.
Chart attributes not exposed by the inspector (grid, routing, tag or link
styles, ...) are preserved verbatim when the document is opened and saved. The
separate **Chart attributes** panel has been removed; use the chart settings
panel for the supported options.
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

When **edit mode** is editable or persistent, Gantt Task, WBS parent, Start, End,
Progress, Group and Depends on cells are editable; non-Gantt Label/X/series cells
are also editable and call
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
tabs: Appearance, Points, Tooltip, Table, Interactions, Events and Data. It controls the
dark/light theme, **Editable colors** (color picker on/off), style preset, point size and point width, tooltip colors/width/font/delay, **Mode**, **Column
filters** on/off and table typography, **Edit mode**
(**non-editable**, **editable** or **editable persistent**), **Critical path**, **Auto schedule**,
**Holidays**, **Weekend**, **Timezone**, **Link color**, **Link width** (default
`1px`), **Min link gap** (default `4px`), **Line style**, **Arrow**, **Arrow
size**, **Bar border** and **Bar border color**; the Data tab has
**Element hints**. Apply writes attributes and
`--sui-chart-tooltip-*` / `--sui-chart-table-*` variables together. Dark and
light themes use explicit chart tokens; selection colors are shared and focus
tones adapt to the theme background. Click a series swatch in a table header or
legend to change that series color; the override is stored per chart. Chart
updates crossfade the old SVG into the new one unless reduced motion is
requested.

The **Edit mode** select in **Interactions** is the 3-state data-control
setting: **non-editable** lets the user read data and use tooltips/selection but
blocks table edits, imports, Gantt drag/resize and diagram edits; **editable**
allows those changes for the page session and restores the source data on reload;
**editable persistent** stores the changes (including imported data, empty
tables filled by the user, table layout, chart attributes and CSS variables) in
`localStorage` so they survive reload and restart. The gear menu and the diagram
settings panel expose **Reset saved data** to discard the persistent state and
return to the design-time source.

Charts can receive keyboard focus; use `ArrowLeft`/`ArrowRight` (or up/down) to
move through data marks, `Home`/`End` to jump, `Enter`/`Space` to select, and
`Escape` to clear the tooltip.

Tooltips, hover brightness, keyboard navigation and selection are available on
every chart. Gantt task rows are keyboard targets: Left/Right move,
Shift+Left/Right resize the finish, Alt+Left/Right resize the start, Enter opens
the task menu, and the dependency connector opens the target picker with Enter.
Focused dependency links support Enter/Shift+F10 to edit and Delete to remove.
**edit mode** editable/persistent additionally enables Gantt drag/resize, collapsible groups and
editable Gantt table cells. Task edits are auto-saved to a local draft and are
not versioned until the header **Commit** button is pressed. **Commit changes...**
and **Version history...** in the chart menu create and inspect versions with
field-level audit details. If the chart has `onTaskChange` in **Chart settings >
Events**, the
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

The inspector no longer exposes `overflow` or `initial zoom`: new charts use
scrollbars automatically when the frame is smaller than the chart. Legacy
`overflow`/`zoom` attributes already present in a source are preserved and still
work in the generated app.

The **preset** menu writes `--sui-chart-*` variables into the element's
**style expr** field under **Appearance**. **Classic** presets are **Default**,
**Ocean**, **Forest**, **Sunset**, **Neon**, **Mono**, **Pastel**,
**High contrast**, **Amber**, **Candy**, **Slate**, **Light**, **Outline** and
**Minimal**. **Theme presets** add one coordinated chart palette for every
shared style family, in both **Light** and **Dark** variants (Standard, Sketch,
Wireframe, Terminal, Glassmorphism, Brutalist, Blueprint, Neumorphic, High
contrast, Editorial, Cyberpunk and Geometric). Choosing **Custom** leaves the
current expression untouched; edit it directly for series colors, grid/axis
colors, spacing, line width and scrollbar colors. The selected preset name is
kept after **Apply**, because the panel recognises a preset even when tooltip
and table overrides are appended to its style expression.

The chart **Theme** combo is independent from the project theme. With the
**Default** preset, a dark chart on a light page keeps a solid dark surface and
light text, and its header, zoom and settings buttons use chart-specific
foreground, background and border colours instead of the light page palette.
Choosing the light chart theme with **Default** switches to the light default
palette. Older default style expressions are recognised as **Default** and are
upgraded on Apply, so a dark chart no longer turns its buttons into an
unreadable dark block.

A data-expression chart is reactive: changing the bound state re-renders the
chart. Malformed data or an empty payload shows a text error inside the chart,
not a broken preview. JSON literal braces must be written `\{` and `\}` in
source; the data overlay does that automatically when applying a literal.

See [API reference: chart](API_REFERENCE.md#chart) and the
[chart example](../examples/chart/README.md).

## Graph control

Insert **Graph** from **Controls**. A new graph starts as an empty block diagram:
states and transitions are data, so they are not injected as demo content; use
**Edit data** or bind a state/expression. It fills its container (`width:
"100%"`, `height: "100%"`). It is a leaf control: it cannot contain child
elements. The **Edit mode** select in **Graph settings  Interactions** makes
the diagram non-editable, editable for the page session, or editable persistent
(`localStorage` across reload/restart). It also controls table editing and the
editing handles, so there is no separate diagram editing switch.

The graph inspector is ordered as follows:
| Field | Meaning |
|---|---|
| name | runtime target used by `chart.*`; it must be unique among charts and graphs |
| Graph settings | first button; opens the same runtime panel as the chart gear (General plus GRID, SHAPES, LINKS, TAGS and HINTS) |
| Styles | color preset and **Edit styles...**, which opens the Graph style editor with Title, Frame, Buttons, Table and Scrollbars tabs |
| show table | shows or hides the synchronized data table |
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

When **edit mode** is editable or persistent, the table is editable: the Element
cell shows a color picker and text field, Tag and Hint are editable, Type becomes
a select, and the last column exposes text settings, element properties and row
deletion. In non-editable mode the table is read-only: cells show plain text, the
actions column is hidden and the Add row button disappears. While typing in a
cell its tooltip is suppressed. New graphs default to non-editable; choose the
mode in Graph settings.

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

## Desktop build dialog

**File → Build desktop executable.** opens the build dialog. The **Target**
selector chooses the window host:

- **Desktop WebView (native window)** - embedded WebView2 on Windows, WKWebView
  on macOS and WebKitGTK on Linux; no separate browser window.
- **Desktop (default browser)** - the executable starts the loopback server and
  opens the system browser.

The dialog preflights the Go toolchain and output directory, bakes the
`--allow-dir` and `--allow-exec` grants into the executable, shows the build log
and can download the result. WebView builds also need CGO and a C compiler; on
Linux install GTK 3 and WebKitGTK 4.0 development packages. The first build
downloads `github.com/webview/webview_go`. Runtime options are `--no-open` and
`--port N`.

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
