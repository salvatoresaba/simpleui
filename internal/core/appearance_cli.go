package core

// appearance_cli.go - validate the external control appearance catalog.
//
//	simpleui appearance check [file] [--json]
//
// The same validator runs when the GUI server loads components/appearance.json,
// so another tool or coding agent can check an edit without starting the
// designer. Without a file argument the command uses the normal lookup order:
// <project>/components/appearance.json, then ./components/appearance.json.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CmdAppearance(args []string) {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Println("usage: " + AppName + " appearance check [file] [--json]")
		fmt.Println("       validates components/appearance.json (format, CSS values, Dark/Light, text-entry rules)")
		return
	}
	if args[0] != "check" {
		UsageError("appearance supports only the check subcommand")
	}
	path := ""
	jsonOut := false
	for _, arg := range args[1:] {
		if arg == "--json" {
			jsonOut = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			UsageError("unknown appearance flag " + arg)
		}
		if path != "" {
			UsageError("appearance check takes at most one file")
		}
		path = arg
	}
	if path == "" {
		path = appearanceLibraryPath("")
		if path == "" {
			path = filepath.Join("components", "appearance.json")
		}
	}
	catalog, warnings, err := loadAppearanceLibraryFile(path)
	if jsonOut {
		if err != nil {
			writeAppearanceJSON(map[string]any{"ok": false, "file": path, "error": err.Error()})
			os.Exit(1)
		}
		writeAppearanceJSON(map[string]any{
			"ok":       true,
			"file":     path,
			"controls": len(catalog.Controls),
			"feedback": len(catalog.Feedback),
			"warnings": warnings,
		})
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: invalid appearance catalog %s: %v\n", AppName, path, err)
		os.Exit(1)
	}
	fmt.Printf("%s: OK (%d controls, %d feedback variants)\n", path, len(catalog.Controls), len(catalog.Feedback))
	for _, warning := range warnings {
		fmt.Printf("warning: %s\n", warning)
	}
}

func writeAppearanceJSON(v map[string]any) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: cannot encode result: %v\n", AppName, err)
		os.Exit(1)
	}
	fmt.Println(string(b))
}
