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
[`docs/COMPONENT_LIBRARY.md`](../docs/COMPONENT_LIBRARY.md). The per-control
appearance catalog (structures, styles, animations, themes) is built into the
editor and is documented in
[`docs/APPEARANCE_LIBRARY.md`](../docs/APPEARANCE_LIBRARY.md); it is a separate
layer from this component-preset file.
