# Component library folder

The designer's external component library is `components/library.json` in the
project directory (or next to the `simpleui` executable). The server looks for:

1. `<project-dir>/components/library.json`
2. `<current-dir>/components/library.json`

The file ships **empty** so the editor starts with only the built-in catalog. Add
groups/components and reload from the editor with
**Preferences → Reload component library**.

Minimal example:

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
          "class": "lib-metric-card",
          "css": ".lib-metric-card { padding: 18px; border-radius: 14px; background: #111827; color: #e5e7eb; }",
          "children": [
            { "kind": "p", "text": "Revenue" },
            { "kind": "h2", "text": "$12,480" }
          ]
        }
      ]
    }
  ]
}
```

Rules and validation are documented in
[`docs/COMPONENT_LIBRARY.md`](../docs/COMPONENT_LIBRARY.md).

## Appearance catalog

`components/appearance.json` (format `simpleui-appearance-library`) is the
complete control appearance catalog: structures, styles and animations for every
control. It is self-contained: `definitions.baseCss` is the full base stylesheet,
`definitions.phaseCss` defines the `sui-anim-*` classes,
`definitions.genericKeyframesCss` and `definitions.keyframes` define every
keyframe, and each control carries `markup`, `classes`, `css` and `keyframes`.
The designer validates it at startup and falls back to the built-in catalog when
the file is missing or rejected. Dark and Light styles are required for every
control; text-entry controls only allow the `None` animation and use the
non-motion `feedback.static` variant. The schema is documented in
[`docs/APPEARANCE_LIBRARY.md`](../docs/APPEARANCE_LIBRARY.md#the-external-appearance-file).

Validate an edit with the same loader the designer uses:

```bash
simpleui appearance check components/appearance.json --json
```

The advisory editor schema is `components/appearance.schema.json`. The CLI,
the GUI loader and the MCP tool `sui_appearance` are authoritative.
