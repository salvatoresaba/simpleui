package sui

import (
	"strings"
	"testing"
)

func TestLooksLikeSuiFile(t *testing.T) {
	cases := map[string]bool{
		"app.sui":       true,
		"APP.SUI":       true,
		`C:\work\a.sui`: true,
		"app.txt":       false,
		"--port":        false,
		"gui":           false,
		"folder.sui/x":  false,
		"folder.sui":    true,
		"a.suix":        false,
	}
	for input, want := range cases {
		if got := looksLikeSuiFile(input); got != want {
			t.Errorf("looksLikeSuiFile(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestWindowsAssociateCommands(t *testing.T) {
	exe := `C:\Tools\simpleui.exe`
	cmds := windowsAssociateCommands(exe)
	if len(cmds) != 5 {
		t.Fatalf("associate commands = %d, want 5", len(cmds))
	}
	joined := make([]string, 0, len(cmds))
	for _, c := range cmds {
		joined = append(joined, strings.Join(c, " "))
	}
	all := strings.Join(joined, "\n")
	for _, want := range []string{
		`HKCU\Software\Classes\.sui`,
		suiProgID,
		`SimpleUI source file`,
		`"C:\Tools\simpleui.exe",0`,
		`"C:\Tools\simpleui.exe" gui "%1"`,
	} {
		if !strings.Contains(all, want) {
			t.Errorf("association commands missing %q", want)
		}
	}
}

func TestWindowsUninstallCommands(t *testing.T) {
	cmds := windowsUninstallCommands(true)
	if len(cmds) != 2 {
		t.Fatalf("uninstall commands = %d, want 2", len(cmds))
	}
	if !strings.Contains(strings.Join(cmds[0], " "), `.sui`) {
		t.Errorf("uninstall does not remove the extension key: %v", cmds[0])
	}
	cmds = windowsUninstallCommands(false)
	if len(cmds) != 1 {
		t.Fatalf("uninstall without extension = %d, want 1", len(cmds))
	}
}
