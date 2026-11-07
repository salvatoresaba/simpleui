package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// associate.go - Windows file association for .sui sources.
//
// `simpleui associate` registers the editor under HKEY_CURRENT_USER, so it
// needs no administrator rights. `--uninstall` removes the keys it created and
// `--dry-run` prints the exact reg.exe commands without changing anything.
//
// Windows may keep a per-user UserChoice for an extension that was already
// opened with another application. The command sets the normal HKCU classes
// association; if Windows keeps the old choice, select SimpleUI once from
// "Open with" or reset the default app in Windows Settings.

const suiProgID = "SimpleUI.sui"

func looksLikeSuiFile(name string) bool {
	if strings.HasPrefix(name, "-") {
		return false
	}
	return strings.HasSuffix(strings.ToLower(name), ".sui")
}

func windowsAssociateCommands(exe string) [][]string {
	quotedExe := `"` + exe + `"`
	return [][]string{
		{"reg", "add", `HKCU\Software\Classes\.sui`, "/ve", "/t", "REG_SZ", "/d", suiProgID, "/f"},
		{"reg", "add", `HKCU\Software\Classes\` + suiProgID, "/ve", "/t", "REG_SZ", "/d", "SimpleUI source file", "/f"},
		{"reg", "add", `HKCU\Software\Classes\` + suiProgID + `\DefaultIcon`, "/ve", "/t", "REG_SZ", "/d", quotedExe + ",0", "/f"},
		{"reg", "add", `HKCU\Software\Classes\` + suiProgID + `\shell\open`, "/ve", "/t", "REG_SZ", "/d", "Open with SimpleUI", "/f"},
		{"reg", "add", `HKCU\Software\Classes\` + suiProgID + `\shell\open\command`, "/ve", "/t", "REG_SZ", "/d", quotedExe + ` gui "%1"`, "/f"},
	}
}

func windowsUninstallCommands(removeExtension bool) [][]string {
	var cmds [][]string
	if removeExtension {
		cmds = append(cmds, []string{"reg", "delete", `HKCU\Software\Classes\.sui`, "/f"})
	}
	cmds = append(cmds, []string{"reg", "delete", `HKCU\Software\Classes\` + suiProgID, "/f"})
	return cmds
}

func windowsAssociationPresent(extKey string) bool {
	out, err := exec.Command("reg", "query", extKey, "/ve").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), strings.ToLower(suiProgID))
}

// CmdAssociate implements `simpleui associate [--uninstall] [--dry-run]`.
func CmdAssociate(args []string) {
	uninstall := false
	dryRun := false
	for _, a := range args {
		switch a {
		case "--uninstall", "-u":
			uninstall = true
		case "--dry-run", "-n":
			dryRun = true
		case "--help", "-h":
			fmt.Printf("usage: %v associate [--uninstall] [--dry-run]\n", AppName)
			return
		default:
			UsageError("unknown flag " + a)
		}
	}
	if runtime.GOOS != "windows" {
		fmt.Fprintf(os.Stderr, "%v: .sui file associations are implemented for Windows only\n", AppName)
		fmt.Fprintf(os.Stderr, "On macOS/Linux, associate .sui with: %v gui \"%%f\"\n", AppName)
		os.Exit(1)
	}
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: cannot locate the SimpleUI executable: %v\n", AppName, err)
		os.Exit(1)
	}
	exe, _ = filepath.Abs(exe)

	extKey := `HKCU\Software\Classes\.sui`
	if uninstall {
		removeExt := windowsAssociationPresent(extKey)
		if !removeExt {
			fmt.Printf("leaving %s untouched (it is not associated with %s)\n", extKey, suiProgID)
		}
		for _, cmd := range windowsUninstallCommands(removeExt) {
			runRegistryCommand(dryRun, cmd)
		}
		if dryRun {
			fmt.Println("dry run: no registry changes were made")
		} else {
			fmt.Printf("removed .sui association for %s\n", AppName)
		}
		return
	}

	commands := windowsAssociateCommands(exe)
	for _, cmd := range commands {
		runRegistryCommand(dryRun, cmd)
	}
	if dryRun {
		fmt.Println("dry run: no registry changes were made")
		return
	}
	fmt.Printf("associated .sui with %s\n", exe)
	fmt.Printf("double-click a .sui file to open the designer\n")
	fmt.Printf("undo with: %v associate --uninstall\n", AppName)
}

func runRegistryCommand(dryRun bool, cmd []string) {
	if dryRun {
		fmt.Println(strings.Join(cmd, " "))
		return
	}
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: registry command failed: %v\n%s", AppName, err, string(out))
		os.Exit(1)
	}
}
