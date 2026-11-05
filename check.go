package main

import (
	"fmt"
	"os"
	"strings"
)

// check.go - `simpleui check`: the fast parse/typecheck loop intended for
// editors, CI and coding agents. With --json it emits structured diagnostics
// that an agent can consume without scraping human output.
func CmdCheck(args []string) {
	jsonOut := false
	security := ""
	input := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--") {
			name, value, hasValue := SplitFlag(a)
			if !hasValue && (name == "security") && i+1 < len(args) {
				value, hasValue = args[i+1], true
				i++
			}
			switch name {
			case "json":
				jsonOut = true
			case "security":
				if !hasValue {
					UsageError("--security needs a JSON file")
				}
				security = value
			case "help", "h":
				fmt.Printf("usage: %v check <file.sui> [--json] [--security <file>]\n", AppName)
				return
			default:
				UsageError("unknown check flag --" + name)
			}
		} else if input == "" {
			input = a
		} else {
			UsageError("unexpected argument " + a)
		}
	}
	if input == "" {
		UsageError("missing input file")
	}
	SetSecurityPath(security)
	prog, dl := LoadProgram(input)
	cerrs := Check(prog)
	dl.Extend(&cerrs)
	ok := !dl.HasErrors()

	if jsonOut {
		fmt.Println(StringifyJson(JObj(
			"ok", JBool(ok),
			"file", JStr(input),
			"diagnostics", guiDiags(dl),
		)))
	} else {
		if dl.Count() > 0 {
			dl.Print()
		} else {
			fmt.Printf("%s: ok\n", input)
		}
	}
	if !ok {
		os.Exit(1)
	}
}
