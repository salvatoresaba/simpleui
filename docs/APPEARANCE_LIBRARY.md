# Control appearance library

Every non-complex control has an appearance catalog in the designer. The only
exceptions are **table**, **graph** and **chart**, which keep their dedicated
data/visualization editors. The catalog separates the three things that make a
control look and feel the way it does:

| Axis | What it controls | Examples (button) |
|---|---|---|
| **Structure** | shape and drawing | Rectangular, Pill, Circular, Square, Wide |
| **Style** | colour palette | Midnight, Primary, Coral, Mint, Neon, Ghost |
| **Animation** | idle effect plus hover, click, focus and state feedback | None, Feedback, Pulse, Float, Glow, Jelly, Bounce |

A **theme** is a named structure + style + animation. Themes can be applied as a
whole, or the three axes can be mixed freely. The catalog ships in
[`components/appearance.json`](../components/appearance.json): that file lists the
structures, styles and animations for every control. The designer validates it at
startup and keeps a built-in fallback catalog, so a missing or rejected file can
never break the editor.

## The inspector section

Select a control (for example a button) and open its **name** section in the
Inspector - for a button the section is called **Button**; for a card it is
**Card**, for a dropdown **Dropdown menu**, and so on. Inside it there are four
main drop-downs in this order: **Theme**, **Structure**, **Style**, **Animation**.

The **Animation** menu selects a coherent set for the whole interaction: the idle
effect plus the feedback for **Hover**, **Click**, **Focus** and **State**
(checked, selected, expanded or open). The four selects under it override a
single phase without changing the others. Applying **Pulse**, for example, keeps
the pulse loop at rest, grows the control on hover, gives it a jelly reaction on
click and shows the focus ring plus a bounce on state. **None** disables every
phase, while **Feedback** keeps no idle effect and still reacts to hover, click,
focus and state.

Text-entry controls - **input**, **textarea**, **select**, **combobox**,
**date/time**, **datepicker**, **number**, **color** and **file** - only expose
**None** and **Feedback**. Their feedback uses ring, glow and box-shadow only:
no translate, scale or rotate, so the field never moves or changes size while
the user is typing. The phase menus for these controls list only the static
options.

- **Structure** — selects the shape/drawing; changes radius, padding, size and
  shape attributes as needed;
- **Style** — selects the border, text, fill and shadow colours;
- **Animation** — selects the full set: idle + hover + click + focus + state;
- **Hover / Click / Focus / State** — refine one phase of the current set;
- **Theme** — applies a structure + style + animation together. **Custom** means
  the current combination does not match a stored theme;
- **Browse…** — opens the Controls library navigator filtered to this control;
- **Save theme…** — stores the current combination and its custom declarations
  in the editor's favourite themes. Saved themes appear immediately in the
  **Theme** list for that control.

The appearance values are written as CSS custom properties (and, for structures
that need it, direct declarations or shape attributes) on **this element only**.
The project default for every control of that kind remains under
**Settings…** / Project preferences. Text content and type, border type and
width, events and alignment stay in their existing inspector sections and can be
combined with any appearance.

## Insert toolbar organization

The Insert toolbar is grouped so the list stays short:

| Group | Subcategories |
|---|---|
| Controls | Inputs, Toggles, Actions, Panels, Data, Feedback, **Advanced** |
| Layout | containers and the Navigation submenu |
| Text | text elements and Lists |
| Overlay / Content / Media | their own groups |

**Advanced** contains **table**, **graph** and **chart**. These three also do not
show the appearance section in the Inspector.

## The Controls library navigator

**Insert → Controls library…** opens the navigator. Its **Appearance** tab is the
appearance browser; the **Components** tab keeps the component presets.

In Appearance:

- the left column lists **every non-complex control** (Button first, then Input,
  Textarea, Select, Combobox, Date/Time, Number, Slider, Checkbox, Radio, Switch,
  Colour, Date, Time, File, Progress, Spinner, Badge, Alert, Card, Tabs,
  Accordion, Menu bar, Dropdown, Popover, Modal, Toast, Tree, Command palette,
  Drawer, Split panes, Dropzone, Timeline, Carousel, Stepper);
- the middle area shows three **side-by-side lists** - **Structures**,
  **Styles** and **Animations** - for the selected kind, so combinations are
  created without scrolling between sections. Each animation is a full set
  (idle + hover + click + focus + state), not only a resting effect;
- the lists are keyboard-navigable: **Up/Down** moves inside the focused list and
  **selects the reached option**, **Left/Right** switches list, **Home/End** jumps
  to the first/last option, **Enter** selects the focused option; the current list
  is highlighted;
- the right column compiles an **interactive preview** of the current
  combination, shows the **Theme** selector, and offers **Save theme…** and
  **Apply to selection**;
- the preview is a real SimpleUI app compiled by the server
  (`/__gui__/source-preview`), so inputs, toggles and tabs interact normally.

Moving in a list is the same as selecting: the highlight, the side lists and the
preview update on every **Up/Down/Home/End** key, without pressing **Enter**
first. **Enter** (or **Space**) re-applies the focused option. The preview runs
the selected animation with the control's own animation variable and keyframes,
so pulse, glow, float and the other effects are visible immediately. The preview
also plays the interaction phases once on load: hover, click, focus and state,
then it leaves the control interactive so you can try them yourself.

Select an option to update the preview immediately. Choose a saved theme to
restore its full combination. **Apply to selection** writes the current
combination to the selected control of that kind; select a matching control
first.

## The external appearance file

`components/appearance.json` is the complete catalog. The editor looks for it in:

1. `<project-dir>/components/appearance.json`
2. `<current-dir>/components/appearance.json`

The repository ships the full file, so every control starts with its structures,
styles and animations. The shape is:

```json
{
  "format": "simpleui-appearance-library",
  "version": 1,
  "feedback": {
    "default": {
      "hover": "sui-fx-lift .18s ease-out forwards",
      "press": "sui-fx-press .18s ease-out",
      "focus": "sui-fx-ring .9s ease-out",
      "state": "sui-fx-pop .24s ease-out"
    }
  },
  "controls": [
    {
      "kind": "button",
      "label": "Button",
      "structures": [
        { "id": "pill", "label": "Pill", "vars": { "--sui-button-radius": "999px" } }
      ],
      "styles": [
        { "id": "dark", "label": "Dark", "vars": { "--sui-button-bg": "#111827" } },
        { "id": "light", "label": "Light", "vars": { "--sui-button-bg": "#f8fafc" } }
      ],
      "animations": [
        { "id": "none", "label": "None", "value": "none" },
        { "id": "glow", "label": "Glow", "value": "sui-button-glow 2.2s ease-in-out infinite" }
      ]
    }
  ]
}
```

Rules:

- `kind` must be one of the controls that expose the appearance section (all
  except **table**, **graph** and **chart**); every kind appears at most once.
- `structures`, `styles` and `animations` each need at least one entry.
- `styles` must contain **Dark** and **Light** (ids `dark` and `light`). They are
  listed first for every control.
- `feedback` defines the interaction phases used by every animation set. Each
  variant needs `hover`, `press`, `focus` and `state`; the editor merges missing
  variants with the built-in feedback. `default`, `glow`, `energetic` and
  `playful` ship in the bundled file and are chosen from the idle effect name.
- `animations` must contain a **None** entry.
- `vars` values are CSS values; `attrs` is limited to the supported structure
  attributes (`shape`, `size`, `variant`, `tone`).
- A structure may also use direct CSS keys such as `border-radius` or `padding`.
- Values containing `;`, `{}`, `<>`, `url(...)`, `expression(...)`,
  `javascript:` or `!important` are rejected.
- Animation and feedback values reference the built-in `sui-fx-*` and
  `sui-appearance-*` keyframes. The JSON selects effects and phase assignments;
  it does not inject arbitrary `@keyframes`.

Validation is atomic: a malformed file rejects the whole catalog and the editor
keeps the built-in fallback, showing the reason in the status bar. After editing
the file, use **Preferences → Reload component and appearance libraries** (or the
**Reload** button in the navigator) to reload it without restarting.

Validate an edit with the same loader the GUI uses:

```bash
simpleui appearance check components/appearance.json --json
```

The MCP server exposes the same check as **`sui_appearance`**. The advisory
editor schema is [`components/appearance.schema.json`](../components/appearance.schema.json);
the CLI and the GUI loader remain authoritative.

Text-entry controls (**input**, **textarea**, **select**, **combobox**,
**date/time**, **datepicker**, **number**, **color**, **file**) must list only
the `None` animation. Phase feedback for them comes from `feedback.static` and
must not contain `sui-fx-lift`, `sui-fx-grow`, `sui-fx-tilt`, `sui-fx-press`,
`sui-fx-pop`, `sui-fx-jelly`, `sui-fx-bounce`, `sui-fx-shake`, `sui-fx-slide`,
`sui-appearance-pulse`, `sui-appearance-float`, `sui-appearance-fade` or
`sui-field-lift`. The validator rejects those values.

## Button example

1. Select a `button` in the Structure tree or preview.
2. In the **Button** section, choose **Circular** as the structure.
3. Choose **Coral** as the style (border, text, fill and shadow change).
4. Choose **Pulse** as the animation. The preview shows the pulse at rest,
   then the hover grow, the jelly click, the focus ring and the state bounce.
5. Open **Browse…**, check the live preview, then **Save theme…** as “Coral
   pulse”.
6. The theme now appears in the button's **Theme** list. Selecting it restores
   the circular shape, coral palette and pulse animation.

The existing inspector fields still apply: edit the label in **Content**, the
font in **Typography**, the border in **Border**, the click in **Events**, and the
alignment in **Alignment**.

## Where themes are stored

Favourite themes are stored in the editor's local storage under
`sui.appearance.themes.v1`. They are available in the inspector and in the
navigator across project loads and editor restarts on the same browser profile.
They are not written into the `.sui` file and do not travel with the project;
the project stores the applied CSS variables/attributes instead, so it renders
without the editor's theme list.

## Layout and contrast safeguards

- **Open overlays stay attached.** While a combobox popup, menu, dropdown,
  popover or tooltip is open, the editor removes transform/filter animations
  from the panel's ancestor chain. Fixed overlays then keep the 4 px gap from
  their trigger instead of jumping to the anchor's page position.
- **Text stays inside the shape.** Buttons, badges, tabs, menu items, cards,
  alerts, toasts and option rows use `min-width: 0` and `overflow-wrap:
  anywhere`; single-line controls ellipsize. Fixed-size shapes such as the
  circular button clip excess text instead of letting it spill outside.
- **Selected text keeps contrast.** Dark and Light palettes assign `*-fg`,
  `*-label` and `*-hover-fg` keys to text colours before hover/border colours.
  The combobox selected option, menu items and tabs therefore stay readable in
  both palettes.

## Catalog coverage and limits

- The bundled `components/appearance.json` covers **every** control that exposes
  the section. Rich catalogs ship for **button**, **checkbox/radio**, **switch**,
  **card**, **input/textarea/select**, **badge**, **progress**, **spinner**,
  **slider**, **alert**, **tabs** and **menu bar**; the other controls get a
  generic catalog with Default, Rounded, Square, Compact and Wide structures,
  **Dark** and **Light** plus Default, Accent and Outline styles, and None, Pulse,
  Glow, Float and Fade animations (`sui-appearance-*` keyframes written to the
  project CSS).
- **Dark** and **Light** are always the first two styles for every control;
  custom palettes follow them.
- Every non-None animation set gives feedback on hover, click, focus and state.
  The `feedback` block in `components/appearance.json` supplies the phase
  declarations and can be edited without touching the compiler. The **Feedback**
  set uses that block with no idle animation, so a control can react to actions
  without a permanent loop.
- Structures and styles are **validated catalog data**, not arbitrary CSS. The
  external `components/library.json` is a separate layer for insertable
  component presets; it does not replace the appearance catalog.
- Themes save the full appearance declarations for the selected element, so a
  theme can include custom values that are not in the three option lists; those
  values are restored with the theme.
- The navigator's preview is generated from the selected triple and compiled by
  the local compiler; the generated application still follows the normal SUI
  security rules.
