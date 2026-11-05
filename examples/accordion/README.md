# Accordion example

This is the complete, runnable demonstration of the `accordion` control. It
shows every feature supported by the control and the normal designer workflow.

## Run it

A desktop build is included in `dist/`:

- **Windows:** double-click `examples/accordion/dist/app.exe`.
- Keep the `assets` folder next to `app.exe`.

To rebuild and launch it from this folder:

```powershell
./run.ps1
```

or, from `cmd.exe`:

```bat
run.bat
```

The scripts build `../accordion.sui` with:

```bash
simpleui build examples/accordion.sui --target desktop --out examples/accordion/dist
```

A web build can be produced with:

```bash
simpleui build examples/accordion.sui --single --out examples/accordion/dist/accordion.html
```

## Complete SUI source

The complete runnable source is `../accordion.sui` and is also copied to `dist/app.sui`.

```sui
# Complete accordion example.
# Open a section, then use "Re-render view" to verify that the open section is
# restored after a normal view re-render. Each accordion is independent.
state rerenders = 0

fn bump() -> null {
	rerenders = rerenders + 1
}

view main {
	title "Accordion - complete example"
	h1 "Accordion"
	p "An accordion groups content into disclosure sections. Only one section per accordion is open at a time."
	p "Try the sections, then click Re-render view: the open section is preserved."

	row {
		button "Re-render view" onClick: bump
		text "Re-renders: {rerenders}"
	}

	accordion "Frequently asked questions" {
		accordionitem "1. Basic usage" {
			p "Put an accordionitem inside an accordion. Its text is the visible header; its children are the panel content."
			p "The first section is open initially. Click a header to open it, or click the open header to close it."
		}
		accordionitem "2. Rich panel content" {
			h2 "Any view content works"
			ul {
				li "Headings, paragraphs and lists"
				li "Buttons, forms and inputs"
				li "Rows, columns and nested layouts"
			}
			row {
				button "A button inside the panel" onClick: bump
				text "This click also re-renders the view."
			}
		}
		accordionitem "3. Re-render persistence" {
			p "Open this section, then click Re-render view above or the button in section 2. The open section stays selected."
			p "This is the normal accordion state-restoration behavior in generated views."
		}
		accordionitem "4. Keyboard and accessibility" {
			p "Headers are native buttons. Use Tab to focus a header and Space or Enter to open or close it."
			p "The header exposes aria-expanded, and panels are hidden while closed."
		}
	}

	h2 "A second, independent accordion"
	p "Each accordion keeps its own open section; opening a section in this one does not affect the FAQ above."

	accordion "Advanced options" {
		accordionitem "How it is emitted" {
			p "The compiler emits .sui-accordion, .sui-accordion-item, .sui-accordion-head and .sui-accordion-panel."
			p "Labels and panel content are escaped normally; no raw HTML is accepted."
		}
		accordionitem "Designer workflow" {
			p "Insert Accordion from Layout > Navigation. Right-click an accordion row for Add accordion item, or an item row for Add item below."
			p "The Text field edits the accessible accordion name or the section header."
		}
	}
}
```

## What the example demonstrates


- `accordion "Accessible name"` creates the container.
- Each direct `accordionitem "Header"` creates one disclosure section.
- The first section is open initially.
- Clicking a closed header opens it and closes the other sections.
- Clicking the open header closes it, leaving all sections closed.
- The open section is restored after a normal view re-render (use **Re-render
  view** in the demo).
- Each accordion has independent state.
- Panel content can contain any normal view content: headings, paragraphs,
  lists, buttons, forms, nested rows/columns and so on.
- Headers are native buttons with `aria-expanded`; Tab + Space/Enter works.
- The second accordion in the demo shows independent state and the designer
  workflow.
- The Text field on an `accordion` is its accessible name; the Text field on an
  `accordionitem` is its visible header.

## Syntax

```sui
accordion "FAQ" {
	accordionitem "What is SimpleUI?" {
		p "A small reactive UI language."
	}
	accordionitem "Can I use my own CSS?" {
		p "Yes, with a css block or the designer."
	}
}
```

Only direct `accordionitem` children are sections. The `accordion` text is an
accessible name, not a visible title. The `accordionitem` text is the visible
header. Panels are hidden while closed.

## Designer

Insert **Accordion** under **Layout → Navigation**. It starts with two sections.
Right-click:

- an `accordion` row for **Add accordion item**, **Edit label**, **Duplicate**,
  **Move up/down**, **Delete**;
- an `accordionitem` row for **Add item below**, **Edit label**, **Duplicate**,
  **Move up/down**, **Delete**.

The generated classes are `.sui-accordion`, `.sui-accordion-item`,
`.sui-accordion-head` and `.sui-accordion-panel`.

The individual `accordionitem` rows are selected from the element tree; a click
in the preview selects the containing accordion.

## Security

The control is rendered through the normal view generator: header labels and
panel content are escaped text/nodes, never raw HTML. The runtime only toggles
classes, `aria-expanded` and the `hidden` attribute.

## Related documentation

- [API reference: accordion](../../docs/API_REFERENCE.md#accordion)
- [GUI editor: accordion control](../../docs/GUI_EDITOR.md#accordion-control)
- [Italian README](../../README_IT.md)
