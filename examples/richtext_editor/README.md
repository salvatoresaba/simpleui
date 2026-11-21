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
  Bold/italic/underline/strike/superscript/subscript/inline code are on/off
  toggles. With a collapsed caret they become typing modes: the button stays
  highlighted and the next typed text uses that format until pressed again. The
  text-color control shows black and the highlight control shows white when no
  color was chosen; a color chosen at the caret stays in the swatch and applies
  to the text typed next, while a selection is colored immediately.
- Paragraph alignment, indentation, line height, spacing before/after and manual
  page breaks (Insert > Page break, `pageBreak` toolbar token, Ctrl+Enter) for
  printing and PDF export.
- Bullet/numbered lists, images, links, tables (with a draggable table toolbar
  for rows, columns, header, merge/split, whole-table selection, cell copy/paste, alignment and deletion) and rules.
  Enter creates one list item per line, and leaving a code block releases only
  the line with the caret while the other lines stay in code blocks. Word styles
  (bold, italic, color, highlight) inside code blocks are kept in the model and
  survive re-renders; indent/outdent adds or removes two spaces on the current
  code line. Inline code and code blocks use the page text color, so they are
  readable on the white page.
- DOCX, ODT, PDF and RTF export the inline styles inside code blocks; syntax
  highlighting and line numbers are exported as colored runs/numbered lines too.
  DOCX, ODT, PDF and RTF include rich header/footer text and images. DOCX, ODT
  and RTF write the table column widths and the automatic hyphenation setting. The PDF export uses its own text layout,
  so it does not hyphenate yet.
- Insert image accepts a file from the PC (up to 8 MB, stored as a data URL) or
  an http(s)/data URL.
- Table settings set the table width (fit to page inside the margins, pixels or
  percent). Dragging a column edge resizes only that column and stores the
  widths in the document; column and row borders can also be dragged on touch
  screens.
- Four toolbar icon styles (Line, Bold line, Duotone, Soft blocks) are available
  from Preferences > Icon style, with larger, clearer icons; the choice is
  remembered per editor.
- Document settings (menu) provides page defaults: size, orientation, margins,
  line height, document language and header/footer. The language drives
  hyphenation, spell checking and DOCX/ODT language metadata. Page settings
  overrides size/orientation/margins for a single page.
- Header and footer open in a floating mini editor from the single
  **Header and footer...** button in Document settings or Page settings, or from
  **Preferences > Header and footer...**. The settings panels show only that
  button; summaries and the variable help stay inside the editor. The editor
  always shows Header/Footer tabs, multi-line text, images, bold/italic/
  underline/alignment, token buttons and a variable menu with `{page}`,
  `{pages}`, `{title}`, `{name}`, `{date}`, `{time}`, `{datetime}`, `{words}` and
  `{chars}`. The **?** button opens the variable help. Rich header/footer height
  is measured so text and images do not overlap the page body.
Both feed paginated view, print and PDF/DOCX/ODT exports.
- Print preview shows the real page split on a white page while the editor
  chrome follows the UI light/dark theme;
  the continuous editor marks the actual print page starts with red dashed
  guides, so the lines match the paginated view and print preview. Hovering a
  guide shows a red
  "Page break - drag" label and a large grab band. The guides are anchored to
  document positions and follow the content while scrolling. Drag a guide, or
  the whole red line/label of an existing break (not just its handle), to move
  the cut; dragging near an edge auto-scrolls, so the target can go beyond the
  visible viewport and a simple click does nothing. Inside a long table the line
  snaps to a row edge and dropping it splits the table around the break,
  preserving rowspan groups, header rows and all data rows. A rowspan group
  that does not fit the remaining page continues across the page boundary, so
  the first page is filled instead of leaving blank space. Moving the break
  out or deleting it recomposes adjacent fragments into one table. Each break
  has a delete control. Every move/delete recalculates the following breaks,
  and oversized paragraphs, lists and tables are split so printed content never
  enters the margins.
- Paragraph settings panel (alignment, indent with explicit −/+ buttons, line
  height, space before/after, hyphenation), zoom controls in the toolbar and a
  print preview overlay. The hyphenation option inserts soft hyphens at syllable
  boundaries, so words break even without a browser dictionary; 57 languages use
  the full TeX hyph-utf8 pattern sets (Italian, English, German, French, Spanish,
  Portuguese, Dutch, Polish, Swedish, Russian, Greek, Turkish, Czech, Hungarian,
  Romanian, Nordic and the other European languages), selected from the document
  language, and the rest fall back to heuristics. The Inspector's **Rich text
  editor > Hyphenation languages...** overlay chooses which dictionaries are
  embedded at compile time with `hyphenLangs`; a small subset cuts `app.js` from
  about 3.4 MB to about 1 MB, while `none` keeps only the heuristic fallback.
  DOCX/ODT/RTF/PDF export optional or soft hyphens without changing paragraph
  line spacing. The
  alignment menu offers **Justify** and **Justify + hyphenation**. **Code tab
  size** (2/4/8 spaces) controls code-line indentation and is stored as
  `tabSize`; **Line numbers (code)** toggles the line-number gutter on code
  blocks (`lineNumbers: true`). **Code language** enables syntax highlighting
  for JavaScript, TypeScript, HTML, CSS, JSON, XML, Markdown, SQL, Python, Go,
  Java, C#, Bash and is stored as `codeLanguage`. The main toolbar and the
  contextual selection toolbar both expose indent/outdent for the current
  paragraph; inside a code block they indent the current code line. Paragraphs
  and lists split at line or item boundaries to fill each page; headings and
  images stay together.
- Settings panels use labels plus a **?** hint icon with the explanation in its
  tooltip instead of inline instruction paragraphs. The toolbar distinguishes
  text color (plain **A** with a color bar, default black) from background
  highlight (highlighted **A** with a swatch, default white). Selection colors
  and highlights are inline spans and also work inside code blocks. The color is
  previewed while the picker is open and is kept when the picker is closed,
  including by clicking outside it. A small **▾** opens a palette with presets
  and recent colors; **Custom...** opens the native picker.
- Font-size buttons A+/A− change the selected text (or the word at the caret);
  zoom only scales the view and never changes the document fonts.
- Image settings: double-click an image or right-click it and choose Edit/Replace/Delete;
  edit the URL, replace it with a file, change alt/width/alignment or remove it.
- Configurable toolbar groups, menu bar (File, Edit, Insert, Preferences; no
  redundant Format/Tools/View menus; Preferences holds Document settings,
  Page settings, Header and footer... and Icon style), contextual selection toolbar and status bar (words,
  characters, paragraphs, caret, zoom). The table toolbar is a dockable group
  in the main toolbar and appears when the caret is in a table. Toolbar categories are
  separate dockable toolbars: drag the ⣿ handle to float one, then drop it on the
  top bar to re-insert it at the nearest slot; the other toolbars move aside. A single theme button at the right of the toolbar
  cycles the interface between dark and light; the document page is always white.
- Undo/redo history, keyboard shortcuts, find and replace (Ctrl+F), sanitized
  paste from Word/Excel/HTML, print preview (Ctrl+P), a paginated A4 preview
  (Pages button) and JSON document binding.
- Import/export JSON, HTML, Markdown and XML.
- Open and save plain text, RTF, DOC (best effort), DOCX and ODT; RTF keeps
  tables (with merged cells and cell shading), fonts, header/footer pictures
  and Wingdings checkboxes, and images are imported/exported with RTF/DOCX/ODT.
  PDF export builds A4 pages with text styles, fonts, colors, tables and
  embedded PNG/JPEG pictures.
- Declarative plugins adding commands and toolbar buttons; JavaScript plugins
  can also register custom commands and block renderers on `window.__SUI_RICH`.
- Document templates: `article`, `letter`, `report`, `compact`.

Select the editor in the designer: the **Rich text editor** section sets the
interface theme (dark/light), toolbar items and the visible UI areas. Toolbar
items are picked in a multi-select overlay; `None` leaves only the theme button.
Height and width are set in the generic **Size** section, and the document page
is always white. The page break is available from Insert, the toolbar and
Ctrl+Enter.

## Component attributes

| Attribute | Meaning |
|---|---|
| `name` | Runtime instance name used by `richtext.*` (defaults to a generated name) |
| `value` | Bound state containing the JSON document |
| Element text | Document title for `{title}` tokens and fallback runtime name; it is not rendered above the menu |
| `onChange` / `onInput` | Change and live-typing handlers (state setter runs on change/blur) |
| `onReady` / `onSelectionChange` | Ready and selection callbacks |
| `height` | Content area height (`px`, `rem`, `vh`, `%`); the status bar stays pinned to the bottom border |
| `placeholder` | Empty-document placeholder |
| `readonly` | Display-only mode: the content cannot be edited and no menu or toolbar is shown |
| `toolbar`, `menu`, `contextToolbar`, `statusbar` | Legacy UI-area toggles, still honored in hand-written sources but not exposed in the designer |
| `toolbarItems` | Comma-separated commands, `|` separates groups; `none` shows no formatting buttons |
| `theme` | Interface theme: `dark` (default) or `light`; the attribute wins over a stored toolbar preference |
| Document paper | Always white; the old `contentTheme` attribute is ignored by the compiler |
| `header` / `footer` | Text or a rich block model repeated on every page; edited in a floating mini editor with text, images and variables (`{page}`, `{pages}`, `{title}`, `{name}`, `{date}`, `{time}`, `{datetime}`, `{words}`, `{chars}`) |
| `page break` | Insert > Page break, the `pageBreak` toolbar token or Ctrl+Enter; drag its handle to move it, use its control to delete it, and following breaks recalculate |
| `template` | `article`, `letter`, `report` or `compact` |

Toolbar items include `undo`, `redo`, `copy`, `cut`, `paste`, `styleSelect`,
`fontFamily`, `fontSize`, `bold`, `italic`, `underline`, `strike`, `textColor`,
`highlight`, `superscript`, `subscript`, `clearFormat`, `alignLeft`,
`alignCenter`, `alignRight`, `justify`, `indent`, `outdent`, `lineHeight`,
`spaceBefore`, `spaceAfter`, `bulletList`, `orderedList`, `blockquote`,
`codeBlock`, `link`, `unlink`, `image`, `table`, `symbol`, `hr`,
`pageSettings`, `printPreview`, `print`, `find`, `copyFormat`,
`fontSizeUp`, `fontSizeDown`,
`paragraphSettings`, `zoom`.
Set `toolbarItems: "none"` to hide every formatting button; the theme button
stays available. The default toolbar exposes a zoom group (−, percentage reset
to 100%, +).

## API

The `richtext` namespace is a core module (no import):

```sui
docJson, err = richtext.get("article")
ok, err = richtext.set("article", docJson)
ok, err = richtext.loadContent("article", "# Loaded from code", "markdown")
ok, err = richtext.loadContent("article", fileBytes, "report.docx")
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

`richtext.loadContent` is async. Besides strings (JSON, HTML, Markdown, XML or
plain text) it accepts an existing file path in desktop builds:

```sui
async fn openExisting() -> null {
	ok, err = richtext.loadContent("article", "C:/docs/report.docx")
}
```

The path must be inside a directory granted with `--allow-dir`. It also accepts
real file bytes (`array[int]` from `fs.readBytes` on desktop, or
`Uint8Array`/`ArrayBuffer`/`Blob`); pass the file name as the third or fourth
argument so the importer is detected from its extension. On the web a local path
cannot be read: use `richtext.loadContent("article")` or
`richtext.openFile("article")` to open the native picker. The example
demonstrates the byte path with a DOCX round-trip.

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
- PDF export keeps the visible layout: styles, fonts, colors, tables and images.
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

async fn loadFromCode() -> null {
	ok, err = richtext.loadContent("article", "# Loaded from SUI code\n\nThis document was passed by **loadContent** as Markdown.\n\n- no file picker\n- no HTML string", "markdown")
	out = ok ? "Content loaded from code" : str(err)
}

async fn loadFileBytes() -> null {
	bytes, berr = richtext.exportBytes("article", "docx")
	if berr != null {
		out = str(berr)
		return
	}
	ok, err = richtext.loadContent("article", bytes, "roundtrip.docx")
	out = ok ? "File bytes loaded (DOCX round-trip)" : str(err)
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
				button "Load from code" onClick: loadFromCode
				button "Load file bytes" onClick: loadFileBytes
			}
			richtext name: "article" value: doc height: "420px" theme: "dark" toolbarItems: "undo,redo,copy,cut,paste|styleSelect,fontFamily,fontSize,fontSizeUp,fontSizeDown,bold,italic,underline,strike,textColor,highlight,code,clearFormat,copyFormat|alignLeft,alignCenter,alignRight,justify,indent,outdent,paragraphSettings,bulletList,orderedList,blockquote,codeBlock|link,unlink,image,table,hr,pageBreak|pageSettings,printPreview,print,find|zoom" onChange: changed
			p "{status}" class: "muted"
			p "The editor is a reusable component: bind value to a state, listen to onChange/onInput, and call richtext.* APIs from any view, menu or plugin. Imported HTML/Markdown/XML is sanitized into the JSON model; scripts, event attributes and unsafe URLs are discarded." class: "muted"
		}
	}
}
```
