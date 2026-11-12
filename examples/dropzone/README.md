# File dropzone

A drop area that accepts files or folders and opens the picker when clicked.
It reuses the existing drag-and-drop and picker runtime, so paths are validated
and stored in a string state.

## Complete source

```sui
state files = ""
state folder = ""

css ".drop-page \\{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \\} .drop-note \\{ color: #8fa0b8; font-size: 13px; \\}"

view main {
	title "File dropzone"
	col class: "drop-page" {
		h1 "File dropzone"
		p "Drag files or click to choose. The state receives the selected paths; drop the same zone into the same state." class: "drop-note"
		dropzone "CSV files" dropFiles: files accept: ".csv,text/csv" multiple: true hint: "Drop CSV files or click to choose" {
			p "The CSV paths are stored in the files state." class: "drop-note"
		}
		p "Files: {files}" class: "drop-note"
		h2 "Folder picker"
		dropzone "Folder" dropFiles: folder directory: true pickMode: "auto" hint: "Drop a folder or click to choose" {
			p "Folder paths are stored in the folder state." class: "drop-note"
		}
		p "Folder: {folder}" class: "drop-note"
	}
}
```

## Attributes

| Attribute | Meaning |
|---|---|
| `dropFiles` | Required string state that receives dragged/dropped paths. |
| `pickFiles` | Optional separate state for the picker; defaults to `dropFiles`. |
| `accept` | Picker filter, for example `.csv` or `image/*`. |
| `multiple` | Allows more than one file. |
| `directory` | Opens the folder picker and accepts dropped folders. |
| `pickMode` | `native`, `fs` (File System Access API) or `auto`. |
| `hint` | Secondary line under the title. |

The element text is the title. Child elements render inside the zone. The status
line shows the bound state value and hides itself while empty. Every value comes
from the drag/drop runtime, so names and paths are escaped as text.

## Designer

Choose **Controls > File dropzone** in the palette. Use the **Dropzone** section
to set the state, `accept`, `multiple`, `directory`, `pickMode` and hint. The
same fields also appear in the shared drag-and-drop section of the inspector.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```
