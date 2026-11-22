package core

import (
	"strings"
	"testing"
)

const dialogTestSource = `import { dialog } from "std"
state file = ""
state files = ""
state dir = ""
state status = ""

async fn pickFile() -> null {
	path, err = dialog.pickFile()
	if err != null {
		status = err
		return
	}
	file = path
}

async fn pickFiles() -> null {
	paths, err = dialog.pickFiles()
	if err != null {
		status = err
		return
	}
	files = join(paths, ",")
}

async fn pickDir() -> null {
	path, err = dialog.pickDirectory()
	if err != null {
		status = err
		return
	}
	dir = path
}

view main {
	button "file" onClick: pickFile
	button "files" onClick: pickFiles
	button "dir" onClick: pickDir
}
`

func TestDialogStdlib(t *testing.T) {
	prog, pdl := ParseSource(dialogTestSource, "dialog.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %s", tcDiagMsgs(pdl))
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %s", tcDiagMsgs(dl))
	}

	web := GenerateJsFor(prog, "web")
	for _, frag := range []string{
		"dialog.pickFile is available only in desktop builds",
		"dialog.pickFiles is available only in desktop builds",
		"dialog.pickDirectory is available only in desktop builds",
	} {
		jsContains(web, frag, t)
	}

	desktop := GenerateJsFor(prog, "desktop")
	for _, frag := range []string{
		"__sui_dialog_pick",
		"/__bridge__/pick-file",
		"/__bridge__/pick-directory",
		"dialog_pickfile",
		"dialog_pickfiles",
		"dialog_pickdirectory",
	} {
		jsContains(desktop, frag, t)
	}
	jsBalanced(desktop, t)
}

func TestDesktopDialogTemplate(t *testing.T) {
	main := DesktopMainSource(nil, nil)
	for _, frag := range []string{
		"/__bridge__/pick-file",
		"/__bridge__/pick-directory",
	} {
		if !strings.Contains(main, frag) {
			t.Errorf("desktop main missing %q", frag)
		}
	}
	for _, frag := range []string{
		"powershell.exe",
		"osascript",
		"zenity",
		"kdialog",
		"OpenFileDialog",
		"FolderBrowserDialog",
	} {
		if !strings.Contains(desktopDialogSource, frag) {
			t.Errorf("dialog source missing %q", frag)
		}
	}
}
