package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// main.go — the SimpleUI compiler CLI.
//
//   simpleui version
//   simpleui help
//   simpleui parse <file.sui>
//   simpleui build <file.sui> [--target web|desktop|desktop-webview] [--out <dir>]
//
// The desktop targets wrap the generated app in a native executable:
// desktop opens the system browser, desktop-webview opens an embedded WebView.

func Main() {
	// A panic in any phase is a compiler bug: never dump a stack trace on the
	// user. Report it cleanly and exit with a distinct code. (os.Exit paths
	// below still exit immediately; this only catches unexpected panics.)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "%v: internal error: %v\n", AppName, r)
			fmt.Fprintf(os.Stderr, "%v: this is a bug in the compiler, not in your code — please report it\n", AppName)
			os.Exit(2)
		}
	}()
	args := os.Args
	cmd := "help"
	if len(args) > 1 {
		cmd = args[1]
	}
	if len(args) > 1 && looksLikeSuiFile(args[1]) {
		// Windows "Open with" and file associations pass the .sui path as the
		// first argument: open the designer on it directly.
		CmdGui(args[1:])
		return
	}
	if cmd == "version" || cmd == "--version" || cmd == "-v" {
		fmt.Printf("%v %v\n", AppName, AppVersion)
	} else if cmd == "help" || cmd == "-h" || cmd == "--help" || cmd == "" {
		PrintHelp()
	} else if cmd == "parse" {
		CmdParse(ArgsAfter(args, 2))
	} else if cmd == "check" {
		CmdCheck(ArgsAfter(args, 2))
	} else if cmd == "appearance" {
		CmdAppearance(ArgsAfter(args, 2))
	} else if cmd == "docs" {
		CmdDocs(ArgsAfter(args, 2))
	} else if cmd == "build" {
		CmdBuild(ArgsAfter(args, 2))
	} else if cmd == "serve" {
		CmdServe(ArgsAfter(args, 2))
	} else if cmd == "fmt" {
		CmdFmt(ArgsAfter(args, 2))
	} else if cmd == "lsp" {
		CmdLsp(ArgsAfter(args, 2))
	} else if cmd == "gui" {
		CmdGui(ArgsAfter(args, 2))
	} else if cmd == "associate" {
		CmdAssociate(ArgsAfter(args, 2))
	} else if cmd == "mcp" {
		CmdMcp(ArgsAfter(args, 2))
	} else {
		fmt.Fprintf(os.Stderr, "%v: unknown command %q\n", AppName, cmd)
		PrintHelp()
		os.Exit(1)
	}
}

func ArgsAfter(args []string, n int) []string {
	var rest []string
	for i := n; i < len(args); i++ {
		rest = append(rest, args[i])
	}
	return rest
}

func CmdParse(args []string) {
	if len(args) != 1 {
		UsageError("parse takes exactly one input file")
	}
	prog, errs := LoadProgram(args[0])
	if errs.HasErrors() {
		errs.Print()
		os.Exit(1)
	}
	fmt.Printf("%v\n", ProgToJson(prog))
}

func CmdBuild(args []string) {
	target := "web"
	out := "dist"
	input := ""
	single := false
	minify := false
	sourcemap := false
	var allowExec []string
	var allowDirs []string
	security := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			name, value, hasValue := SplitFlag(arg)
			needsValue := name == "target" || name == "out" || name == "allow-exec" ||
				name == "allow-dir" || name == "security"
			if !hasValue && needsValue && i+1 < len(args) {
				value, hasValue = args[i+1], true
				i++
			}
			if name == "target" {
				if !hasValue {
					UsageError("--target needs a value (web|desktop|desktop-webview)")
				}
				target = value
			} else if name == "out" {
				if !hasValue {
					UsageError("--out needs a value")
				}
				out = value
			} else if name == "allow-exec" {
				if !hasValue {
					UsageError("--allow-exec needs a command name")
				}
				allowExec = append(allowExec, value)
			} else if name == "allow-dir" {
				if !hasValue {
					UsageError("--allow-dir needs a directory")
				}
				allowDirs = append(allowDirs, value)
			} else if name == "security" {
				if !hasValue {
					UsageError("--security needs a JSON file")
				}
				security = value
			} else if name == "single" {
				single = true
			} else if name == "minify" {
				minify = true
			} else if name == "sourcemap" {
				sourcemap = true
			} else {
				UsageError("unknown flag --" + name)
			}
		} else if input == "" {
			input = arg
		} else {
			UsageError("unexpected argument " + arg)
		}
	}

	if input == "" {
		UsageError("missing input file")
	}
	if target != "web" && target != "desktop" && target != "desktop-webview" {
		UsageError("unknown target " + target)
	}
	if minify && !single {
		UsageError("--minify requires --single")
	}
	if sourcemap && minify {
		UsageError("--sourcemap cannot be combined with --minify")
	}
	if sourcemap && (target == "desktop" || target == "desktop-webview") {
		UsageError("--sourcemap is available for web builds and serve")
	}
	if single && (target == "desktop" || target == "desktop-webview") {
		UsageError("--single is for web builds; the desktop targets are already one executable")
	}
	if !single && (strings.HasSuffix(strings.ToLower(out), ".html") || strings.HasSuffix(strings.ToLower(out), ".htm")) {
		UsageError("--out ending in .html requires --single")
	}

	SetSecurityPath(security)
	prog, errs := LoadProgram(input)
	if errs.HasErrors() {
		errs.Print()
		os.Exit(1)
	}
	cerrs := Check(prog)
	errs.Extend(&cerrs)
	if errs.HasErrors() {
		errs.Print()
		os.Exit(1)
	}
	if errs.Count() > 0 {
		// Warnings only: report them but keep building.
		errs.Print()
	}

	if target == "desktop" {
		BuildDesktop(prog, out, allowExec, allowDirs)
		return
	}
	if target == "desktop-webview" {
		BuildDesktopWebview(prog, out, allowExec, allowDirs)
		return
	}
	if single {
		path := SingleOutputPath(out)
		parent := filepath.Dir(path)
		if parent != "." && parent != "" {
			if err := os.MkdirAll(parent, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "%v: cannot create output directory %q: %v\n", AppName, parent, err)
				os.Exit(1)
			}
		}
		html := BuildSingleHTMLOpts(prog, minify, sourcemap)
		if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "%v: cannot write %q: %v\n", AppName, path, err)
			os.Exit(1)
		}
		fmt.Printf("wrote single HTML to %v\n", path)
		return
	}

	appJs := GenerateJs(prog)
	indexHtml := GenerateIndexHtml(prog)
	styleCss := GenerateCss(prog)
	sourceBase := prog.dir
	if sourceBase == "" {
		sourceBase = filepath.Dir(prog.file)
	}
	var appMap, cssMap string
	if sourcemap {
		appJs, appMap = GenerateJsWithSourceMap(prog, "web", sourceBase)
		styleCss, cssMap = GenerateCssWithSourceMap(prog, sourceBase)
		appJs = appJs + "\n//# sourceMappingURL=app.js.map\n"
		styleCss = styleCss + "\n/*# sourceMappingURL=style.css.map */\n"
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "%v: cannot create output directory %q: %v\n", AppName, out, err)
		os.Exit(1)
	}
	WriteFileTo(out, "index.html", indexHtml)
	WriteFileTo(out, "style.css", styleCss)
	WriteFileTo(out, "app.js", appJs)
	if sourcemap {
		WriteFileTo(out, "app.js.map", appMap)
		WriteFileTo(out, "style.css.map", cssMap)
	}
	if sourcemap {
		fmt.Printf("wrote index.html, style.css, app.js and source maps to %v\n", out)
	} else {
		fmt.Printf("wrote index.html, style.css, app.js to %v\n", out)
	}
}

// SplitFlag splits "--name=value" into (name, value, hasValue).
func SplitFlag(arg string) (name, value string, hasValue bool) {
	eq := 0
	for eq < len(arg) && arg[eq] != '=' {
		eq++
	}
	if eq < len(arg) {
		return arg[2:eq], arg[eq+1:], true
	}
	return arg[2:], "", false
}

func WriteFileTo(dir, name, content string) {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%v: cannot write %q: %v\n", AppName, path, err)
		os.Exit(1)
	}
}

func UsageError(msg string) {
	fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, msg)
	PrintHelp()
	os.Exit(1)
}

func PrintHelp() {
	fmt.Printf("%v - %v\n\n", AppName, AppTagline)
	fmt.Printf("usage:\n")
	fmt.Printf("  %v version              print version\n", AppName)
	fmt.Printf("  %v help                 show this help\n", AppName)
	fmt.Printf("  %v parse <file.sui>     tokenize and parse, print the AST as JSON\n", AppName)
	fmt.Printf("  %v check <file.sui>     parse + typecheck; --json for structured diagnostics\n", AppName)
	fmt.Printf("  %v appearance check [f] validate the appearance catalog (default: components/appearance.json; --json)\n", AppName)
	fmt.Printf("  %v fmt <file.sui>       canonical format (--write to update in place)\n", AppName)
	fmt.Printf("  %v docs                 documentation index (also: --agent, --json, --read, --search)\n", AppName)
	fmt.Printf("  %v gui [file.sui]       visual design tool (palette + live preview + .sui export)\n", AppName)
	fmt.Printf("  %v <file.sui>           open the visual designer on that file\n", AppName)
	fmt.Printf("  %v associate            associate .sui files with the editor (Windows, HKCU)\n", AppName)
	fmt.Printf("         --uninstall       remove the .sui association\n")
	fmt.Printf("         --dry-run         print the registry commands without changing anything\n")
	fmt.Printf("         --security <f>   strict/relaxed policy file (default: ./simpleui.security.json)\n")
	fmt.Printf("  %v mcp                   Model Context Protocol server (for AI agents)\n", AppName)
	fmt.Printf("  %v serve <file.sui>     build + serve with browser hot-reload\n", AppName)
	fmt.Printf("         --port <n>       bind port (default: random)\n")
	fmt.Printf("         --watch           recompile and reload on source changes\n")
	fmt.Printf("         --sourcemap       emit JS/CSS source maps while serving\n")
	fmt.Printf("         --security <f>   policy file (default: next to the .sui)\n")
	fmt.Printf("  %v build <file.sui>     compile a .sui app to a web or desktop app\n", AppName)
	fmt.Printf("         --target web     generate index.html, style.css, app.js (default)\n")
	fmt.Printf("         --target desktop generate a desktop executable that opens the browser\n")
	fmt.Printf("         --target desktop-webview generate a desktop executable with an embedded WebView\n")
	fmt.Printf("         --out <dir>      output directory (default: ./dist)\n")
	fmt.Printf("         --single         web: emit one self-contained app.html\n")
	fmt.Printf("         --minify         with --single: strip comments/indentation\n")
	fmt.Printf("         --sourcemap      web: write app.js.map/style.css.map (inline with --single)\n")
	fmt.Printf("         --allow-exec <c> (desktop) grant the bridge the command name <c>\n")
	fmt.Printf("         --allow-dir <d>  (desktop) grant the bridge read/write under <d> (repeatable)\n")
	fmt.Printf("         --security <f>   policy file (default: simpleui.security.json next to <file.sui>)\n")
	fmt.Printf("\n")
}
