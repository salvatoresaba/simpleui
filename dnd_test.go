package sui

import (
	"os"
	"strings"
	"testing"
)

const dndTestSource = `state items = ["a", "b"]
state attachment = ""

fn reorder() -> null {
	let src = indexOf(items, drag.source())
	let dst = indexOf(items, drag.target())
	if src < 0 or dst < 0 or src == dst {
		return
	}
	items = moveItem(items, src, dst)
}

fn inspect() -> null {
	let names = []
	loop f in drag.files() {
		names = append(names, f["path"])
	}
	attachment = join(names, ", ")
}

view main {
	list item in items {
		row draggable: true dragKey: item dropKey: item onDrop: reorder {
			text "{item}"
		}
	}
	input value: attachment dropFiles: attachment onDrop: inspect
	div "drop" draggable: true dragKey: attachment dropKey: attachment onDrop: reorder
	button "pick" pickFiles: attachment accept: ".txt" multiple: true directory: true pickMode: "fs"
	file directory: true multiple: true
}
`

func TestDndTypecheckAndGenerate(t *testing.T) {
	prog, pdl := ParseSource(dndTestSource, "dnd.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %s", tcDiagMsgs(dl))
	}
	js := GenerateJs(prog)
	jsBalanced(js, t)
	for _, frag := range []string{
		`data-sui-drag-key="${__sui_esc(item)}"`,
		`data-sui-drop-key="${__sui_esc(item)}"`,
		`data-sui-dropfiles="setAttachment"`,
		`ondrop="reorder()"`,
		`ondrop="inspect()"`,
		`webkitdirectory`,
		`function __sui_move_item`,
		`function __sui_drag_source`,
		`function __sui_drag_folder`,
		`const __SUI_DND`,
	} {
		jsContains(js, frag, t)
	}
	if strings.Contains(js, "e2.ondrop = reorder") {
		t.Errorf("list row DnD must use inline handlers, not query wiring")
	}
}

func TestPickFilesTypecheckAndGenerate(t *testing.T) {
	src := `state attachment = ""
view main {
	button "Choose" pickFiles: attachment accept: ".csv,.txt" multiple: true directory: true pickMode: "auto"
	file accept: ".csv" multiple: true pickFiles: attachment
	input value: attachment
}
`
	prog, pdl := ParseSource(src, "pick.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %s", tcDiagMsgs(dl))
	}
	js := GenerateJs(prog)
	jsBalanced(js, t)
	for _, frag := range []string{
		`data-sui-pickfiles="setAttachment"`,
		`data-sui-pick-accept=".csv,.txt"`,
		`data-sui-pick-multiple="1"`,
		`data-sui-pick-directory="1"`,
		`data-sui-pick-mode="auto"`,
		`function onPickClick`,
		`function handlePicked`,
		`showDirectoryPicker`,
	} {
		jsContains(js, frag, t)
	}
}

func TestHttpPostFilesTypecheckAndGenerate(t *testing.T) {
	src := `import { http } from "std"
async fn upload() -> null {
	body, err = http.postFiles("/api/analyze")
	if err != null {
		return
	}
	body2, err2 = http.postFilesSafe("/api/analyze", "/api/")
	if err2 != null {
		return
	}
}
view main {
	button "go" onClick: upload
}
`
	prog, pdl := ParseSource(src, "upload.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %s", tcDiagMsgs(dl))
	}
	js := GenerateJs(prog)
	jsBalanced(js, t)
	for _, frag := range []string{`http_postfiles`, `http_postfilessafe`, `function __sui_files_form`} {
		jsContains(js, frag, t)
	}
}

func TestPickModeRejectsUnknownValue(t *testing.T) {
	src := `state attachment = ""
view main {
	button "Choose" pickFiles: attachment directory: true pickMode: "magic"
}
`
	prog, pdl := ParseSource(src, "badmode.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if !hasMsg(dl, "pickMode must be one of: native, fs, auto") {
		t.Fatalf("expected pickMode diagnostic, got: %s", tcDiagMsgs(dl))
	}
}

func TestPickFilesRequiresState(t *testing.T) {
	src := `state attachment = ""
view main {
	button "Choose" pickFiles: "not-a-state"
}
`
	prog, pdl := ParseSource(src, "badpick.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if !hasMsg(dl, "pickFiles must name a declared state") {
		t.Fatalf("expected pickFiles diagnostic, got: %s", tcDiagMsgs(dl))
	}
}

func TestPickFilesExampleChecks(t *testing.T) {
	src, err := os.ReadFile("examples/file_picker.sui")
	if err != nil {
		t.Fatal(err)
	}
	prog, pdl := ParseSource(string(src), "examples/file_picker.sui")
	if pdl.HasErrors() {
		t.Fatalf("example parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("example check diagnostics: %s", tcDiagMsgs(dl))
	}
}

func TestDndDropFilesRequiresState(t *testing.T) {
	src := `state attachment = ""
view main {
	input value: attachment dropFiles: "not-a-state"
}
`
	prog, pdl := ParseSource(src, "bad.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if !hasMsg(dl, "dropFiles must name a declared state") {
		t.Fatalf("expected dropFiles diagnostic, got: %s", tcDiagMsgs(dl))
	}
}

func TestDndDropFilesRejectsNonStringState(t *testing.T) {
	src := `state count = 0
view main {
	input value: count dropFiles: count
}
`
	prog, pdl := ParseSource(src, "badtype.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if !hasMsg(dl, "dropFiles state must be a string") {
		t.Fatalf("expected string-state diagnostic, got: %s", tcDiagMsgs(dl))
	}
}

func TestDndDragBuiltinTypes(t *testing.T) {
	src := `fn inspect() -> string {
	let names = []
	loop f in drag.files() {
		names = append(names, f["path"])
	}
	if drag.active() {
		return join(names, ",") + drag.error()
	}
	return drag.source() + drag.target() + drag.data() + drag.folder()
}
view main {
	button "inspect" onClick: inspect
}
`
	prog, pdl := ParseSource(src, "types.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %s", tcDiagMsgs(dl))
	}
}

func TestMoveItemBuiltinTyping(t *testing.T) {
	src := `state items = ["a", "b"]
fn bad() -> null {
	items = moveItem("no", 0, 1)
}
view main {
	button "go" onClick: bad
}
`
	prog, pdl := ParseSource(src, "move.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if !hasMsg(dl, "moveItem expects an array") {
		t.Fatalf("expected moveItem diagnostic, got: %s", tcDiagMsgs(dl))
	}
}

func TestDndExampleChecks(t *testing.T) {
	src, err := os.ReadFile("examples/drag_drop.sui")
	if err != nil {
		t.Fatal(err)
	}
	prog, pdl := ParseSource(string(src), "examples/drag_drop.sui")
	if pdl.HasErrors() {
		t.Fatalf("example parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("example check diagnostics: %s", tcDiagMsgs(dl))
	}
}

func TestDndGuiRoundTrip(t *testing.T) {
	prog, pdl := ParseSource(dndTestSource, "dnd.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	model := StringifyJson(guiModelJson(prog))
	for _, frag := range []string{
		`"draggable":true`,
		`"dragKey":"attachment"`,
		`"dropKey":"attachment"`,
		`"dropFiles":"attachment"`,
		`"pickFiles":"attachment"`,
		`"pickMode":"fs"`,
		`"onDrop":"reorder"`,
		`"directory":true`,
		`"accept":".txt"`,
	} {
		if !strings.Contains(model, frag) {
			t.Errorf("designer model missing %q\n%s", frag, model)
		}
	}
}
