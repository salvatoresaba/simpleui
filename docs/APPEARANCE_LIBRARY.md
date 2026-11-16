# Control appearance library

Every non-complex control has an appearance catalog in the designer. The only
exceptions are **table**, **graph** and **chart**, which keep their dedicated
data/visualization editors. The catalog separates the three things that make a
control look and feel the way it does:

| Axis | What it controls | Examples (button) |
|---|---|---|
| **Structure** | shape and drawing | Rectangular, Pill, Circular, Square, Wide |
| **Style** | colour palette | Midnight, Primary, Coral, Mint, Neon, Ghost |
| **Animation** | motion of the parts | None, Pulse, Float, Glow, Jelly, Bounce |

A **theme** is a named structure + style + animation. Themes can be applied as a
whole, or the three axes can be mixed freely. The catalog is derived from the
component definitions (`COMPONENT_DEFS`), so every control has at least a default
structure and style and, when its definition has one, its animation options.

## The inspector section

Select a control (for example a button) and open its **name** section in the
Inspector - for a button the section is called **Button**; for a card it is
**Card**, for a dropdown **Dropdown menu**, and so on. Inside it there are four
drop-downs in this order: **Theme**, **Structure**, **Style**, **Animation**.

- **Structure** — selects the shape/drawing; changes radius, padding, size and
  shape attributes as needed;
- **Style** — selects the border, text, fill and shadow colours;
- **Animation** — selects the effect applied to the control;
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
  Drawer, Split panes, Dropzone, Timeline, Carousel, Stepper, List);
- the middle area shows three **side-by-side lists** - **Structures**,
  **Styles** and **Animations** - for the selected kind, so combinations are
  created without scrolling between sections;
- the lists are keyboard-navigable: **Up/Down** moves inside the focused list,
  **Left/Right** switches list, **Home/End** jumps to the first/last option,
  **Enter** selects the focused option; the current list is highlighted;
- the right column compiles an **interactive preview** of the current
  combination, shows the **Theme** selector, and offers **Save theme…** and
  **Apply to selection**;
- the preview is a real SimpleUI app compiled by the server
  (`/__gui__/source-preview`), so inputs, toggles and tabs interact normally.

Select an option to update the preview immediately. Choose a saved theme to
restore its full combination. **Apply to selection** writes the current
combination to the selected control of that kind; select a matching control
first.

## Button example

1. Select a `button` in the Structure tree or preview.
2. In **Control appearance**, choose **Circular** as the structure.
3. Choose **Coral** as the style (border, text, fill and shadow change).
4. Choose **Pulse** as the animation.
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

## Catalog coverage and limits

- Rich catalogs ship for **button**, **checkbox/radio**, **switch**, **card**,
  **input/textarea/select**, **badge**, **progress**, **spinner**, **slider**,
  **alert**, **tabs** and **menu bar**. Other controls get a generic catalog with
  Default, Rounded, Square, Compact and Wide structures; **Dark**, **Light**,
  Default, Accent and Outline styles; and None, Pulse, Glow, Float and Fade
  animations (`sui-appearance-*` keyframes written to the project CSS).
- **Dark** and **Light** are always the first two styles for every control;
  custom palettes follow them.
- Structures and styles are **validated built-in data**, not arbitrary CSS. The
  external `components/library.json` is a separate layer for insertable
  component presets; it does not replace the appearance catalog.
- Themes save the full appearance declarations for the selected element, so a
  theme can include custom values that are not in the three option lists; those
  values are restored with the theme.
- The navigator's preview is generated from the selected triple and compiled by
  the local compiler; the generated application still follows the normal SUI
  security rules.
