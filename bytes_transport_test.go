package sui

import (
	"strings"
	"testing"
)

// bytes_transport_test.go - binary transport for the PDF pipeline.
//
// The three sources a document viewer needs are a folder the app can read
// (fs.readBytes on desktop), a remote document (http.getBytes) and a file the
// user picked (files.readBytes). These tests keep the binary path intact:
// text transports corrupt PDF bytes.

const bytesTransportSource = `import { http, files } from "std"

state picked = ""
state size = 0
state status = ""

async fn loadRemote() -> null {
	data, err = http.getBytes("https://example.com/report.pdf")
	if err != null {
		status = "remote error: " + err
		return
	}
	size = len(data)
}

async fn loadRemoteSafe() -> null {
	data, err = http.getBytesSafe("https://example.com/report.pdf", "https://example.com")
	if err != null {
		status = "safe error: " + err
		return
	}
	size = len(data)
}

async fn loadPicked() -> null {
	data, err = files.readBytes(picked)
	if err != null {
		status = "local error: " + err
		return
	}
	size = len(data)
}

view main {
	row {
		button "Remote" onClick: loadRemote
		button "Remote safe" onClick: loadRemoteSafe
		button "Picked" onClick: loadPicked
	}
	p "{status}"
}
`

func TestBytesTransportTypecheck(t *testing.T) {
	prog, dl := ParseSource(bytesTransportSource, "bytes_transport.sui")
	if dl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(dl))
	}
	if errs := Check(prog); errs.HasErrors() {
		t.Fatalf("check diagnostics: %s", tcDiagMsgs(errs))
	}
	js := GenerateJs(prog)
	for _, frag := range []string{
		"http_getbytes:",
		"http_getbytessafe:",
		"files_readbytes:",
		"await __SUI_STD.http_getbytes(",
		"await __SUI_STD.files_readbytes(",
	} {
		if !strings.Contains(js, frag) {
			t.Errorf("generated app.js missing %q", frag)
		}
	}
}

// The binary response must be read as bytes, never decoded as text.
func TestBytesTransportUsesArrayBuffer(t *testing.T) {
	prog, dl := ParseSource(bytesTransportSource, "bytes_transport.sui")
	if dl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(dl))
	}
	js := GenerateJs(prog)
	for _, frag := range []string{
		"arrayBuffer()",
		"__sui_fetch_bytes_guarded",
		"new Uint8Array(buf)",
	} {
		if !strings.Contains(js, frag) {
			t.Errorf("generated app.js missing %q", frag)
		}
	}
	if strings.Contains(js, "http_getbytes: async (url) => { try { const r = await fetch(url); return [await r.text(), null];") {
		t.Errorf("http.getBytes must not decode the response as text")
	}
}

// files.readBytes reads the current selection, so the drag-and-drop/file
// picker runtime must be part of the generated app.
func TestFilesReadBytesEmbedsDndRuntime(t *testing.T) {
	prog, dl := ParseSource(bytesTransportSource, "bytes_transport.sui")
	if dl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(dl))
	}
	js := GenerateJs(prog)
	for _, frag := range []string{
		"__sui_dnd_read_bytes",
		"file.arrayBuffer()",
		"__SUI_STD_EXTRA.files.readBytes",
	} {
		if !strings.Contains(js, frag) {
			t.Errorf("generated app.js missing %q", frag)
		}
	}
}

// On desktop the file system is the source of truth: the picker returns paths,
// so the app must be told to use fs.readBytes instead.
func TestFilesReadBytesDesktopHint(t *testing.T) {
	prog, dl := ParseSource(bytesTransportSource, "bytes_transport.sui")
	if dl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(dl))
	}
	js := GenerateJsFor(prog, "desktop")
	if !strings.Contains(js, "use fs.readBytes(path)") {
		t.Errorf("desktop build does not point to fs.readBytes(path)")
	}
	if strings.Contains(js, "http.getBytes is not available") {
		t.Errorf("desktop build must keep http.getBytes available")
	}
}

// The byte functions return array[int]: member access on them is rejected like
// on every other array.
func TestBytesTransportResultType(t *testing.T) {
	src := `import { http } from "std"
async fn run() -> null {
	data, err = http.getBytes("https://example.com/a.pdf")
	if err == null {
		size = data.length
	}
}
`
	prog, dl := ParseSource(src, "bytes_type.sui")
	if dl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(dl))
	}
	errs := Check(prog)
	if !hasMsg(errs, "member access is not allowed on a value of type array[int]") {
		t.Fatalf("expected the array[int] member-access diagnostic, got: %s", tcDiagMsgs(errs))
	}
}
