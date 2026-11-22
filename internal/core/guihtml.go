package core

import _ "embed"

// guihtml.go - the embedded visual design tool (structured CSS engine UI).
// Plain HTML/CSS/JS shipped inside the single compiler binary.
//
// Pillars:
//   - CSS is a data model (AST), not a string: /__css__/parse + /__css__/set
//     let the browser parse/generate/edit the stylesheet.
//   - two-way binding with a single source of truth (root.css); programmatic
//     edits set .value (which does not fire oninput) so no sync loops.
//   - sandboxed live preview compiled by the compiler (/__gui__/render).
//
// Layout (dockable panels, dark theme):
//   - top bar: breakpoints, export, status
//   - left:  Structure (palette + tree + element attrs) | Tokens
//   - center: Preview | SUI | CSS (editor with line numbers)
//   - right: Inspector (state, specificity, visual styler)
//   - bottom: Problems | History (collapsible)
//
// The page lives in assets/designer.html so it can be edited with HTML/CSS/JS
// tooling; go:embed keeps the single-binary distribution unchanged.

//go:embed assets/designer.html
var guiHTML string
