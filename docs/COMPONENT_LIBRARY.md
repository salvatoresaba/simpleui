# External component library

The designer can load its Insert catalog from a JSON file that external tools
edit. The file is read when the GUI starts and can be reloaded without restarting
the editor. Loading is validated and atomic: a malformed or unsafe file is
rejected as a whole, the reason is shown in the Insert menu and the status bar,
and the built-in catalog stays available. A broken library cannot break the
designer or the generated application.

## Where the file lives

The server looks for the library next to the project, then in the working
directory:

1. `<project-dir>/components/library.json`
2. `<current-dir>/components/library.json`

The repository ships an example in
[`components/library.json`](../components/library.json). When no file is found,
the designer uses the built-in components and shows a soft warning; this is not
an error.

Use **Preferences → Reload component library** after editing the file. The
server validates it again and the Insert menu is rebuilt. The editor fetches the
library once at startup (`/__gui__/components`) and the JavaScript itself is
served as `/__gui__/components.js`.

## Format

```json
{
  "format": "simpleui-component-library",
  "version": 1,
  "groups": [
    {
      "id": "custom-cards",
      "label": "Custom cards",
      "components": [
        {
          "id": "metric-card",
          "label": "Metric card",
          "kind": "card",
          "default": true,
          "description": "A compact KPI card.",
          "class": "lib-metric-card",
          "css": ".lib-metric-card { display: flex; flex-direction: column; gap: 6px; padding: 18px; border: 1px solid #334155; border-radius: 14px; background: #111827; color: #e5e7eb; }",
          "children": [
            { "kind": "p", "text": "Revenue" },
            { "kind": "h2", "text": "$12,480" },
            { "kind": "p", "text": "+8.2% vs last month" }
          ]
        }
      ]
    }
  ]
}
```

### Groups

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Unique, `^[a-z][a-z0-9-]{0,39}$` |
| `label` | yes | Group name shown in the Insert menu (1–60 characters) |
| `components` | yes | Components of this control family |

The groups are the control families of the Insert menu. They are shown as
flyout rows before the **Built-in** groups, so a library can organize custom
components by type without removing access to standard controls.

Every group must contain at least one component: a control type always has a
version available. Empty groups are rejected at load. Each kind used by the
library gets a default component: the one marked `"default": true`, or the first
component of that kind when no explicit default exists.

### Components

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Unique, lowercase letters, digits and `-` |
| `label` | yes | Insert-menu label (1–60 characters) |
| `kind` | yes | A supported SimpleUI element kind (see limits) |
| `default` | no | Marks the default version of its `kind` for fallback; the first component of a kind is the implicit default |
| `description` | no | Tooltip text (max 240 characters) |
| `class` | no | CSS class applied to the component; required when `css` is set |
| `css` | no | One rule for that class, validated against a safe property list |
| `override` | no | When true, this entry replaces the built-in palette slot for the same `kind` |
| `text` | no | Text literal for the component (max 300 characters) |
| `attrs` | no | Literal attributes (strings, booleans or numbers) |
| `children` | no | Nested elements; only for container kinds |

`children` entries use the same shape except `css` and `override`: `kind`,
`text`, `class`, `attrs`, `children`.

Allowed `attrs` are `placeholder`, `value`, `checked`, `disabled`, `name`,
`options`, `min`, `max`, `step`, `open`, `shape`, `theme`, `variant` and `size`.
Values must be literal; expressions, `onClick`, `href`, `src`, `styleExpr`,
`classExpr`, motion graphs and raw markup are intentionally rejected.

## How the editor uses it

- The library groups appear under **Component library** in **Insert**.
- The component search dialog (Insert → Search components…) includes library
  components and built-ins.
- Inserting a component creates the node tree (`kind` + `text` + `attrs` +
  `children`) and writes the CSS as a managed block:

```css
/* sui-library:lib-metric-card */
.lib-metric-card { ... }
/* /sui-library:lib-metric-card */
```

- `override: true` on a component with a built-in `kind` removes that kind's
  built-in palette row and uses the library label/group instead. It does not
  change the kind's implementation or inspector fields.
- The library cannot remove built-in components that no entry overrides. This
  keeps the editor usable if the external file is incomplete.

## Fallback for missing components

Inserted library components are tracked in the managed CSS block:

```css
/* sui-library:lib-metric-card component=metric-card kind=card */
.lib-metric-card { ... }
/* /sui-library:lib-metric-card */
```

When a project is opened (or the library is reloaded), the editor checks every
`component=<id>` marker against the loaded library. If the component no longer
exists:

1. a warning is shown in the status bar, in **Insert** and in the Controls
   library navigator;
2. the default version for that `kind` is selected (explicit `default`, else the
   first component of the kind);
3. the element's class and the managed CSS block are rewritten to the default
   version, so the control keeps a working style;
4. the element content (text and children) is preserved.

If the library has no component at all for the kind, the warning says so and the
current styles are kept. When the whole library is unavailable, the project still
renders with its materialized CSS; the warning explains that no fallback could be
chosen.

## Controls library navigator

**Insert → Controls library…** opens the library browser:

- groups on the left, filtered by control type;
- search across label, id, kind, description and group;
- a **Default** badge on the fallback version of each kind;
- an **interactive preview** of the selected control, compiled by the server and
  shown in a sandboxed iframe;
- **Insert** to add the selected component to the project at the current
  selection;
- **Reload** to re-read `components/library.json` and **Reload preview** for the
  current selection.

The navigator also shows the library source, rejection errors, fallback warnings
and the current group counts. It is the quickest way for an external tool author
to verify the catalog before opening the project.

## Security and stability checks

Loading is **all-or-nothing**. The server rejects the file if any check fails
and returns the reason; the editor keeps the built-in catalog and shows
`Library rejected: ...` in the Insert menu.

Checks:

- file size at most 1 MB, valid UTF-8 JSON, unknown fields rejected;
- `format` must be `simpleui-component-library`, `version` must be `1`;
- at most 20 groups and 200 components, unique group and component ids;
- every group has at least one component; a kind's explicit or implicit default is
  used for fallback;
- component labels, descriptions, classes and texts are length-limited;
- id/class patterns are validated; classes starting with `sui-` are rejected so
  a library cannot override the component skin;
- `kind` must be in the supported element list (no script, style, iframe, raw
  SVG, chart/table internals or media with external sources);
- `attrs` are restricted to the literal allow-list above, with type checks;
- children are limited to container kinds, depth 6 and 100 nodes per component;
- `css` is a single rule `.class { ... }`; every declaration is checked against
  a property allow-list, `url(...)`, `@` rules, `expression`, `behavior`,
  `javascript:`, braces and semicolons in values are rejected, `position` is
  limited to `relative`/`static`, negative `z-index` is rejected;
- every component is compiled and type-checked with the real compiler (errors
  reject the library; warnings do not);
- a missing file is a soft fallback, not a failure.

The generated application never receives a new script, event handler or external
resource from the library. CSS is written as ordinary author CSS and classes are
validated before insertion.

## Limits

- A library component is a **preset over an existing SimpleUI kind**, not
  arbitrary SUI source. For new behavior, add a compiler feature first.
- The supported kinds are the standard layout, text, form and feedback elements.
  Charts, tables, trees and media are intentionally outside the custom-preset
  surface.
- `css` is one rule for the component class; global selectors, media queries,
  keyframes and animations are rejected.
- Built-in components are protected; `override` can only rename/re-group a kind,
  not replace its implementation.

## Workflow for external tools

1. Write or update `components/library.json`.
2. Run the editor and use **Preferences → Reload component library**.
3. If the Insert menu shows `Library rejected: ...`, read the message, fix the
   file, reload. The built-in catalog remains usable in the meantime.
4. Optionally validate the file with the test suite:
   `go test -run TestBundledComponentLibrary .`

The bundled example is intentionally small: a metric card, a profile card, a
search bar and a status pill. Copy one entry and change its `id`, `label`,
`kind`, `class`, `css` and children.
