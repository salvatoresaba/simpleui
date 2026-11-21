# Rich text, HTML and Markup modes

This example shows the same rich text control used with three authoring
profiles and one read-only viewer:

- **rich** - all document features and all file formats;
- **html** - HTML tools and files only;
- **markup** - Markdown tools and files only;
- **viewer** - HTML rendered read-only, without editing chrome.

Load the samples to see different page backgrounds. Use the **Source** button
in the status bar or the **Source code** (`sourceMode`) tool in the **Document**
toolbar group of the HTML and Markup editors to edit the source directly or paste
external code. The source editor has a line-number gutter, lightweight syntax
highlighting and Tab/Shift+Tab indentation for selected lines. **Load external
HTML** and **Load external Markdown** use `richtext.loadContent` to sanitize and
import code from the page.

The page background is stored in the document JSON and can be changed from
Document settings and Page settings; dark backgrounds automatically use
readable text. Keyboard shortcuts and plugin buttons are filtered to the active
profile, so Markup mode never exposes page breaks, printing or underline.

## Run

``bash
simpleui check examples/richtext_modes/app.sui --json
simpleui serve examples/richtext_modes/app.sui --watch
simpleui build examples/richtext_modes/app.sui --out dist
``

## Complete source

``sui
css "body \{
  margin: 0;
  padding: 24px;
  background: #0f1117;
  color: #e6e6ef;
  font-family: system-ui, sans-serif;
\}
.page \{ max-width: 1180px; margin: 0 auto; \}
.card \{ padding: 16px; margin-bottom: 16px; border: 1px solid #2c3444; border-radius: 12px; background: #151922; \}
.row \{ display: flex; gap: 8px; flex-wrap: wrap; align-items: center; margin-bottom: 12px; \}
.grid \{ display: grid; grid-template-columns: 1fr 1fr; gap: 16px; \}
.muted \{ color: #8fa0b8; font-size: 12px; line-height: 1.5; \}
@media (max-width: 860px) \{ .grid \{ grid-template-columns: 1fr; \} \}"

state richDoc = ""
state htmlDoc = ""
state markdownDoc = ""
state viewerDoc = ""
state message = "Ready. Load the samples or paste source from the Source button."

fn loadSamples() -> null {
	ok1, err1 = richtext.set("rich", {"type": "doc", "page": {"background": "#ffffff"}, "children": [
		{"type": "heading", "level": 2, "children": [{"type": "text", "text": "Rich text editor", "marks": []}]},
		{"type": "paragraph", "children": [
			{"type": "text", "text": "All document features: ", "marks": []},
			{"type": "text", "text": "bold", "marks": [{"type": "bold"}]},
			{"type": "text", "text": ", colors, tables, page setup and file import/export.", "marks": []}
		]}
	]})
	ok2, err2 = richtext.set("html", {"type": "doc", "page": {"background": "#eef2ff"}, "children": [
		{"type": "heading", "level": 2, "children": [{"type": "text", "text": "HTML editor", "marks": []}]},
		{"type": "paragraph", "children": [{"type": "text", "text": "HTML profile with a light blue page background. Use Source to edit external HTML.", "marks": []}]}
	]})
	ok3, err3 = richtext.set("markup", {"type": "doc", "page": {"background": "#f1fbf4"}, "children": [
		{"type": "heading", "level": 2, "children": [{"type": "text", "text": "Markup editor", "marks": []}]},
		{"type": "paragraph", "children": [{"type": "text", "text": "Markdown profile. The Source button edits Markdown directly.", "marks": []}]}
	]})
	ok4, err4 = richtext.set("viewer", {"type": "doc", "page": {"background": "#111827"}, "children": [
		{"type": "heading", "level": 2, "children": [{"type": "text", "text": "Read-only viewer", "marks": []}]},
		{"type": "paragraph", "children": [{"type": "text", "text": "The same control in read-only mode renders content without editing chrome.", "marks": []}]}
	]})
	message = ok1 ? "Samples loaded" : str(err1)
}

async fn loadExternalHtml() -> null {
	ok, err = richtext.loadContent("html", "<h2>External HTML</h2><p>This code was pasted from outside the editor and sanitized into the document model.</p><ul><li>No scripts</li><li>No event attributes</li></ul>", "html")
	message = ok ? "External HTML loaded" : str(err)
}

async fn loadExternalMarkdown() -> null {
	ok, err = richtext.loadContent("markup", "# External Markdown\n\nThis content was pasted from outside the editor.\n\n- **Bold**\n- *Italic*\n- `Code`", "markdown")
	message = ok ? "External Markdown loaded" : str(err)
}

view main {
	div "" class: "page" {
		row {
			h1 "Rich text, HTML and Markup modes"
			button "Load samples" onClick: loadSamples
			button "Load external HTML" onClick: loadExternalHtml
			button "Load external Markdown" onClick: loadExternalMarkdown
		}
		p "{message}" class: "muted"
		div "" class: "grid" {
			card "" class: "card" {
				richtext name: "rich" mode: "richtext" value: richDoc height: "280px" theme: "dark"
			}
			card "" class: "card" {
				richtext name: "html" mode: "html" value: htmlDoc height: "280px" theme: "dark" toolbarItems: "undo,redo|bold,italic,link|documentSettings,sourceMode"
			}
			card "" class: "card" {
				richtext name: "markup" mode: "markup" value: markdownDoc height: "280px" theme: "dark" toolbarItems: "undo,redo|bold,italic,link|documentSettings,sourceMode"
			}
			card "" class: "card" {
				richtext name: "viewer" mode: "html" readonly: true value: viewerDoc height: "280px" theme: "dark"
			}
		}
		p "Use the Source button in the status bar or the sourceMode tool in the Document toolbar group to edit HTML/Markdown source. The editor shows line numbers and lightweight highlighting; Tab and Shift+Tab indent selected lines. The page background is set per document and per page in Document/Page settings." class: "muted"
	}
}

``
