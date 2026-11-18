# Rich text editor

A reusable rich text editor component for SimpleUI. The document is stored as
a JSON tree (`doc`), not as an HTML string. The editor core, rendering layer,
UI chrome, plugin registry and serializers are independent: host views only
bind `value`, listen to `onChange`/`onInput`, and call the `richtext.*` API.

## Run

```bash
simpleui check examples/richtext_editor/app.sui --json
simpleui serve examples/richtext_editor/app.sui --watch
# or
simpleui build examples/richtext_editor/app.sui --out dist
```

No desktop capability is required. The editor is also available in desktop
builds.

## What it demonstrates

- Bold, italic, underline, strikethrough, superscript/subscript, inline code,
  color, background color, font family/size, headings, quotes, code blocks.
- Paragraph alignment, indentation, line height and spacing before/after.
- Bullet/numbered lists, images, links, tables (with a draggable table toolbar
  for rows, columns, header, merge/split, whole-table selection, cell copy/paste, alignment and deletion) and rules.
- Configurable toolbar groups, menu bar, contextual selection toolbar and
  status bar (words, characters, paragraphs, caret, zoom).
- Undo/redo history, keyboard shortcuts and JSON document binding.
- Import/export JSON, HTML, Markdown and XML.
- Open and save plain text, RTF, DOC (best effort), DOCX and ODT; export PDF.
- Declarative plugins adding commands and toolbar buttons; JavaScript plugins
  can also register custom commands and block renderers on `window.__SUI_RICH`.
- Document templates: `article`, `letter`, `report`, `compact`.

## Component attributes

| Attribute | Meaning |
|---|---|
| `name` | Runtime instance name used by `richtext.*` (defaults to a generated name) |
| `value` | Bound state containing the JSON document |
| `onChange` / `onInput` | Change and live-typing handlers (state setter runs on change/blur) |
| `onReady` / `onSelectionChange` | Ready and selection callbacks |
| `height` | Minimum editor height (`px`, `rem`, `vh`, `%`) |
| `placeholder` | Empty-document placeholder |
| `readonly` | Disable editing |
| `toolbar`, `menu`, `contextToolbar`, `statusbar` | Toggle UI areas |
| `toolbarItems` | Comma-separated commands, `|` separates groups |
| `theme` | `dark` (default) or `light` |
| `template` | `article`, `letter`, `report` or `compact` |

Toolbar items include `undo`, `redo`, `copy`, `cut`, `paste`, `styleSelect`,
`fontFamily`, `fontSize`, `bold`, `italic`, `underline`, `strike`, `textColor`,
`highlight`, `superscript`, `subscript`, `clearFormat`, `alignLeft`,
`alignCenter`, `alignRight`, `justify`, `indent`, `outdent`, `lineHeight`,
`spaceBefore`, `spaceAfter`, `bulletList`, `orderedList`, `blockquote`,
`codeBlock`, `link`, `image`, `table`, `symbol`, `hr`.

## API

The `richtext` namespace is a core module (no import):

```sui
docJson, err = richtext.get("article")
ok, err = richtext.set("article", docJson)
ok, err = richtext.command("article", "bold")
ok, err = richtext.insertText("article", "Hello")
ok, err = richtext.insertImage("article", "https://example.com/a.png", "alt")
ok, err = richtext.insertLink("article", "https://example.com", "label")
ok, err = richtext.insertTable("article", 3, 3)
ok, err = richtext.undo("article")
ok, err = richtext.redo("article")
text, err = richtext.exportFormat("article", "markdown")
ok, err = richtext.importFormat("article", text, "markdown")
ok, err = richtext.registerPlugin("demo", pluginMap)
names, err = richtext.plugins()
meta, err = richtext.status("article")
sel, err = richtext.selection("article")
ok, err = richtext.setToolbar("article", "bold,italic|link")
ok, err = richtext.setTemplate("article", "letter")
ok, err = richtext.destroy("article")
```

### Declarative plugin

A plugin can register commands and toolbar entries without JavaScript:

```sui
let plugin = {
  "commands": {
    "shout": {"title": "Shout", "command": "bold"}
  },
  "toolbar": [{"id": "shout", "label": "Shout"}]
}
ok, err = richtext.registerPlugin("demo", plugin)
```

### JavaScript plugin

For custom block rendering or commands with logic, host JavaScript can call:

```js
window.__SUI_RICH.registerPlugin("chartBlock", {
  commands: { insertChart: function (inst) { /* ... */ } },
  toolbar: [{ id: "insertChart", label: "Chart" }],
  render: function (node) { /* return an HTMLElement */ }
})
```

## Serialization

- **JSON** is the native document model.
- **HTML** import is parsed into the model; scripts, styles, event attributes
  and unknown elements are discarded. Export emits only generated, trusted
  markup.
- **Markdown** supports headings, paragraphs, bold/italic/strike/code, links,
  images, lists, quotes, code fences, horizontal rules and pipe tables.
- **XML** uses the `<document>`, `<paragraph>`, `<heading>`, `<text marks="...">`,
  `<table>`, `<image>` structure. DOCTYPE/ENTITY are rejected.

## Security

Images allow `http(s)`, `blob:` and `data:image/...;base64`. Links allow
`http(s)`, `mailto:`, `tel:`, fragments and relative paths. `javascript:` and
other unsafe schemes are rejected. Imported content is rebuilt with DOM APIs,
never assigned as raw HTML. The editor does not require a desktop bridge.

## Limitations

- The engine uses `contenteditable` + browser selection APIs; behavior can vary
  slightly between Chromium, Firefox and WebKit.
- Table merging/splitting and tracked changes are not part of the core; add them
  through a plugin.
- Legacy binary `.doc` files are not parsed; use `.docx` or RTF. DOC export writes RTF content with a `.doc` extension.
- PDF export uses a text/table layout and represents images by their alt text.
- The editor intentionally exposes formatting through the toolbar/menu and API;
  host apps can build their own inspector with `richtext.command` and `richtext.get`.

## Complete source

```sui
import { json } from "std"

css "body \{ margin: 0; padding: 24px; background: #0f1117; color: #e6e6ef; font-family: system-ui, sans-serif; \}
.page \{ max-width: 1100px; margin: 0 auto; \}
.card \{ padding: 18px; border: 1px solid #2c3444; border-radius: 12px; background: #151922; \}
.row \{ display: flex; gap: 8px; flex-wrap: wrap; align-items: center; margin: 10px 0; \}
pre \{ white-space: pre-wrap; font-size: 12px; line-height: 1.5; max-height: 260px; overflow: auto; \}
.muted \{ color: #8fa0b8; font-size: 12px; line-height: 1.6; \}"

state doc = ""
state status = "Ready - select text and use the toolbar or menu."
state out = ""

fn changed() -> null {
	status = "Document changed"
}

async fn loadSample() -> null {
	let sample = {"type": "doc", "children": [
		{"type": "heading", "level": 1, "children": [{"type": "text", "text": "SimpleUI rich text", "marks": []}]},
		{"type": "paragraph", "children": [
			{"type": "text", "text": "This document is stored as JSON, not as HTML. ", "marks": []},
			{"type": "text", "text": "Bold", "marks": [{"type": "bold"}]},
			{"type": "text", "text": ", ", "marks": []},
			{"type": "text", "text": "italic", "marks": [{"type": "italic"}]},
			{"type": "text", "text": " and ", "marks": []},
			{"type": "text", "text": "links", "marks": [{"type": "link", "href": "https://example.com"}]},
			{"type": "text", "text": " are inline marks.", "marks": []}
		]},
		{"type": "bullet_list", "children": [
			{"type": "list_item", "children": [{"type": "paragraph", "children": [{"type": "text", "text": "Editable blocks and lists", "marks": []}]}]},
			{"type": "list_item", "children": [{"type": "paragraph", "children": [{"type": "text", "text": "Undo and redo history", "marks": []}]}]}
		]},
		{"type": "blockquote", "children": [{"type": "paragraph", "children": [{"type": "text", "text": "Plugins extend commands without changing the core.", "marks": []}]}]}
	]}
	ok, err = richtext.set("article", sample)
	out = ok ? "Sample document loaded" : str(err)
}

fn bold() -> null {
	ok, err = richtext.command("article", "bold")
	out = ok ? "Bold applied" : str(err)
}

async fn addImage() -> null {
	ok, err = richtext.insertImage("article", "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxNjAiIGhlaWdodD0iOTAiPjxyZWN0IHdpZHRoPSIxNjAiIGhlaWdodD0iOTAiIGZpbGw9IiM2ZjdmZmYiLz48dGV4dCB4PSI4MCIgeT0iNTIiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IndoaXRlIiBmb250LWZhbWlseT0ic2Fucy1zZXJpZiIgZm9udC1zaXplPSIxOCI+U2FtcGxlPC90ZXh0Pjwvc3ZnPg==", "Sample image")
	out = ok ? "Image inserted" : str(err)
}

fn addLink() -> null {
	ok, err = richtext.insertLink("article", "https://simpleui.dev", "SimpleUI")
	out = ok ? "Link inserted" : str(err)
}

fn addTable() -> null {
	ok, err = richtext.insertTable("article", 3, 3)
	out = ok ? "Table inserted" : str(err)
}

fn undo() -> null {
	ok, err = richtext.undo("article")
	out = ok ? "Undone" : str(err)
}

fn redo() -> null {
	ok, err = richtext.redo("article")
	out = ok ? "Redone" : str(err)
}

fn exportMarkdown() -> null {
	text, err = richtext.exportFormat("article", "markdown")
	out = err == null ? text : str(err)
}

async fn importMarkdown() -> null {
	ok, err = richtext.importFormat("article", "# Imported title\n\nThis was imported from **Markdown**.\n\n- first item\n- second item", "markdown")
	out = ok ? "Markdown imported" : str(err)
}

fn registerPlugin() -> null {
	let plugin = {"commands": {"shout": {"title": "Shout", "command": "bold"}}, "toolbar": [{"id": "shout", "label": "Shout"}]}
	ok, err = richtext.registerPlugin("demo", plugin)
	out = ok ? "Plugin registered" : str(err)
}

fn pluginCommand() -> null {
	ok, err = richtext.command("article", "shout")
	out = ok ? "Plugin command ran" : str(err)
}

async fn showStatus() -> null {
	value, err = richtext.status("article")
	data, perr = json.parse(value)
	if err != null or perr != null {
		out = str(err) + str(perr)
		return
	}
	out = "words=" + str(data["words"]) + " chars=" + str(data["chars"]) + " paragraphs=" + str(data["paragraphs"])
}

view main {
	div "" class: "page" {
		card "" class: "card" {
			row {
				h1 "Rich text editor"
				button "Load sample" onClick: loadSample
			}
			richtext "Article" name: "article" value: doc height: "420px" theme: "dark" menu: true toolbar: true contextToolbar: true statusbar: true toolbarItems: "undo,redo,copy,cut,paste|styleSelect,fontFamily,fontSize,bold,italic,underline,strike,textColor,highlight,clearFormat|alignLeft,alignCenter,alignRight,indent,outdent,lineHeight,bulletList,orderedList,blockquote,codeBlock|link,image,table,hr" onChange: changed
			p "{status}" class: "muted"
			p "The editor is a reusable component: bind value to a state, listen to onChange/onInput, and call richtext.* APIs from any view, menu or plugin. Imported HTML/Markdown/XML is sanitized into the JSON model; scripts, event attributes and unsafe URLs are discarded." class: "muted"
		}
	}
}
```
