package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// desktop_backend.go — the optional Go backend of a desktop build.
//
// `simpleui build app.sui --target desktop --backend <dir>` links a Go module
// into the generated desktop executable, so the user runs ONE process instead
// of two (the app and a separate server).
//
// <dir> is the ROOT DIRECTORY of a Go module (it holds a go.mod with a module
// directive) and its ROOT PACKAGE must export:
//
//	func Handler(fallback http.Handler) http.Handler
//
// The backend is mounted IN FRONT of the compiled app assets: every request it
// does not handle itself is delegated to the fallback handler, which serves
// the SimpleUI app. The backend package name does not matter: the generated
// code imports it with an explicit alias.

const (
	// desktopBaseGoVersion is the go directive of the generated module when
	// the backend does not need a newer one.
	desktopBaseGoVersion = "1.24.2"

	// backendAlias is the import alias used for the backend package in the
	// generated main.go.
	backendAlias = "backend"
)

// backendRequire is one require directive of the backend go.mod.
type backendRequire struct {
	path    string
	version string
}

// backendModule describes the backend module linked into a desktop build.
type backendModule struct {
	dir        string           // the directory as given by the user
	absDir     string           // its absolute path (the replace target)
	modulePath string           // the module path declared by dir/go.mod
	goVersion  string           // the go directive of dir/go.mod
	requires   []backendRequire // the require directives of dir/go.mod
}

// prepareBackend validates the backend contract and describes the module.
// An empty dir means "no backend" and returns (nil, nil).
func prepareBackend(dir string) (*backendModule, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, nil
	}
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("backend: no go.mod in %v", dir)
		}
		return nil, fmt.Errorf("backend: cannot read %v: %v", goModPath, err)
	}
	modulePath, goVersion, requires, err := parseBackendGoMod(string(data))
	if err != nil {
		return nil, fmt.Errorf("backend: cannot parse %v: %v", goModPath, err)
	}
	if modulePath == "" {
		return nil, fmt.Errorf("backend: %v has no module directive in go.mod", dir)
	}
	if err := validateBackendContract(dir); err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("backend: cannot resolve %v: %v", dir, err)
	}
	return &backendModule{
		dir:        dir,
		absDir:     abs,
		modulePath: modulePath,
		goVersion:  goVersion,
		requires:   requires,
	}, nil
}

// goMod renders the go.mod of the generated desktop module. The backend is
// required and replaced by its local directory, so the build needs no
// network access; the requirements of the backend are carried over as
// indirect requirements.
func (b *backendModule) goMod(directRequires []string) string {
	var sb strings.Builder
	sb.WriteString("module desktopapp\n\n")
	sb.WriteString("go " + newerGoVersion(desktopBaseGoVersion, b.goVersion) + "\n")
	if len(directRequires) > 0 {
		sb.WriteString("\nrequire (\n")
		for i := 0; i < len(directRequires); i++ {
			sb.WriteString("\t" + directRequires[i] + "\n")
		}
		sb.WriteString(")\n")
	}
	sb.WriteString("\nrequire " + b.modulePath + " v0.0.0\n")
	sb.WriteString("\nreplace " + b.modulePath + " => " + goModPathValue(b.absDir) + "\n")
	if len(b.requires) > 0 {
		sb.WriteString("\nrequire (\n")
		for i := 0; i < len(b.requires); i++ {
			sb.WriteString("\t" + b.requires[i].path + " " + b.requires[i].version + " // indirect\n")
		}
		sb.WriteString(")\n")
	}
	return sb.String()
}

// copyGoSum copies the go.sum of the backend next to the generated module,
// so its external dependencies resolve with the same checksums.
func (b *backendModule) copyGoSum(outDir string) error {
	data, err := os.ReadFile(filepath.Join(b.dir, "go.sum"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("backend: cannot read go.sum in %v: %v", b.dir, err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "go.sum"), data, 0o644); err != nil {
		return err
	}
	return nil
}

// parseBackendGoMod extracts the module path, the go directive and the
// require directives of a backend go.mod. The replace and exclude directives
// of the backend are not carried over: the backend is replaced by its local
// directory instead.
func parseBackendGoMod(text string) (string, string, []backendRequire, error) {
	var modulePath string
	var goVersion string
	var requires []backendRequire
	block := ""
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if block != "" {
			if line == ")" {
				block = ""
				continue
			}
			if block == "require" {
				if r, ok := parseRequireLine(line); ok {
					requires = append(requires, r)
				}
			}
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "module":
			modulePath = moduleValue(line)
		case "go":
			if len(fields) > 1 {
				goVersion = fields[1]
			}
		case "require":
			rest := strings.TrimSpace(strings.TrimPrefix(line, "require"))
			if strings.HasPrefix(rest, "(") {
				block = "require"
				continue
			}
			if r, ok := parseRequireLine(rest); ok {
				requires = append(requires, r)
			}
		case "replace", "exclude", "retract", "toolchain", "godebug":
			if len(fields) == 1 {
				block = fields[0]
			}
		}
	}
	return modulePath, goVersion, requires, nil
}

// parseRequireLine parses one require entry ("path version"), ignoring the
// trailing comment (e.g. "// indirect").
func parseRequireLine(line string) (backendRequire, bool) {
	if i := strings.Index(line, "//"); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return backendRequire{}, false
	}
	return backendRequire{path: fields[0], version: fields[1]}, true
}

// moduleValue extracts the module path of a "module" directive.
func moduleValue(line string) string {
	line = strings.TrimSpace(strings.TrimPrefix(line, "module"))
	if i := strings.Index(line, "//"); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	return strings.TrimSpace(line)
}

// validateBackendContract checks that the root package of the backend
// directory exports func Handler(fallback http.Handler) http.Handler.
func validateBackendContract(dir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return fmt.Errorf("backend: cannot parse %v: %v", dir, err)
	}
	for name := range pkgs {
		if name == "main" {
			return fmt.Errorf("backend: %v has a package main in its root: a backend must be an importable package", dir)
		}
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || fn.Name.Name != "Handler" {
					continue
				}
				if !isBackendHandlerSignature(fn) {
					return fmt.Errorf("backend: %v must export func Handler(fallback http.Handler) http.Handler", dir)
				}
				return nil
			}
		}
	}
	return fmt.Errorf("backend: %v must export func Handler(fallback http.Handler) http.Handler", dir)
}

// isBackendHandlerSignature reports whether fn has the shape
// func Handler(fallback http.Handler) http.Handler.
func isBackendHandlerSignature(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	return isHTTPHandlerType(fn.Type.Params.List[0].Type) &&
		isHTTPHandlerType(fn.Type.Results.List[0].Type)
}

// isHTTPHandlerType reports whether expr is the type http.Handler.
func isHTTPHandlerType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "http" && sel.Sel.Name == "Handler"
}

// goModPathValue renders a filesystem path as a go.mod replace target:
// forward slashes (go.mod accepts them on Windows too) and quotes when the
// path contains spaces.
func goModPathValue(path string) string {
	value := filepath.ToSlash(path)
	if strings.ContainsAny(value, " \"") {
		return strconv.Quote(value)
	}
	return value
}

// newerGoVersion returns the newer of two go directives ("1.24.2" style).
func newerGoVersion(a string, b string) string {
	if compareGoVersions(a, b) >= 0 {
		return a
	}
	return b
}

// compareGoVersions compares two go directives: -1, 0 or 1.
func compareGoVersions(a string, b string) int {
	av := parseGoVersion(a)
	bv := parseGoVersion(b)
	for i := 0; i < 3; i++ {
		if av[i] != bv[i] {
			if av[i] < bv[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// parseGoVersion parses a go directive into its major/minor/patch numbers.
func parseGoVersion(v string) [3]int {
	var out [3]int
	v = strings.TrimSpace(v)
	if i := strings.IndexAny(v, " \t"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	for i := 0; i < 3 && i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			break
		}
		out[i] = n
	}
	return out
}
