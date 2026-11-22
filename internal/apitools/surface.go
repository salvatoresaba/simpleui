// Package apitools extracts the exported API surface of a Go package from its
// source files.
//
// It exists so the public contract of the `sui` package can be snapshotted
// before the implementation moves into `internal/core` and compared against
// the facade that replaces it. The extraction is purely syntactic (go/ast), so
// it works on a package that does not compile yet and always reports the same
// shape for the original sources and for the generated facade.
package apitools

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Param is one parameter or result of a function signature.
type Param struct {
	Name string // empty when the source declares an unnamed parameter
	Type string // rendered type, e.g. "string" or "...string"
}

// Func describes an exported top-level function.
type Func struct {
	Name     string
	Params   []Param
	Results  []Param
	Variadic bool
	Doc      string
}

// Signature renders the declaration as it would appear in source, without the
// body: "func Name(a int, b ...string) (ok bool, err error)".
func (f Func) Signature() string {
	var b strings.Builder
	b.WriteString("func ")
	b.WriteString(f.Name)
	b.WriteByte('(')
	for i, p := range f.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p.Name)
		b.WriteByte(' ')
		b.WriteString(p.Type)
	}
	b.WriteByte(')')
	if len(f.Results) == 0 {
		return b.String()
	}
	b.WriteByte(' ')
	if len(f.Results) == 1 && f.Results[0].Name == "" {
		b.WriteString(f.Results[0].Type)
		return b.String()
	}
	b.WriteByte('(')
	for i, r := range f.Results {
		if i > 0 {
			b.WriteString(", ")
		}
		if r.Name != "" {
			b.WriteString(r.Name)
			b.WriteByte(' ')
		}
		b.WriteString(r.Type)
	}
	b.WriteByte(')')
	return b.String()
}

// Call renders the argument list used to forward a call: the parameter names,
// with "..." appended to the variadic one.
func (f Func) Call() string {
	var parts []string
	for i, p := range f.Params {
		if p.Name == "" {
			parts = append(parts, fmt.Sprintf("arg%d", i+1))
			continue
		}
		if f.Variadic && i == len(f.Params)-1 {
			parts = append(parts, p.Name+"...")
			continue
		}
		parts = append(parts, p.Name)
	}
	return strings.Join(parts, ", ")
}

// Type describes an exported named type.
type Type struct {
	Name string
	Doc  string
}

// Var describes an exported package-level variable.
type Var struct {
	Name string
	Type string
	Doc  string
}

// Const describes an exported package-level constant.
type Const struct {
	Name string
	Doc  string
}

// Method describes an exported method and the type that carries it.
type Method struct {
	Receiver string
	Pointer  bool // the method is declared on the pointer receiver
	Name     string
	Doc      string
}

// Expression renders the method expression used to reference the method from
// outside the package, e.g. "sui.DiagList.Add" or "(*sui.DiagList).Add".
func (m Method) Expression(pkgAlias string) string {
	if m.Pointer {
		return "(*" + pkgAlias + "." + m.Receiver + ")." + m.Name
	}
	return pkgAlias + "." + m.Receiver + "." + m.Name
}

// Surface is the exported API of a package.
type Surface struct {
	Funcs   []Func
	Types   []Type
	Vars    []Var
	Consts  []Const
	Methods []Method
}

// ExtractDir parses every non-test Go file in dir and returns the exported
// API surface of the package it declares.
func ExtractDir(dir string) (*Surface, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", dir, err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("parse %s: expected exactly one package, got %d", dir, len(pkgs))
	}
	var s Surface
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			if err := collectFile(fset, file, &s); err != nil {
				return nil, err
			}
		}
	}
	s.sort()
	return &s, nil
}

func collectFile(fset *token.FileSet, file *ast.File, s *Surface) error {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				if !d.Name.IsExported() {
					continue
				}
				fn, err := collectFunc(fset, d)
				if err != nil {
					return err
				}
				s.Funcs = append(s.Funcs, fn)
				continue
			}
			if !d.Name.IsExported() || len(d.Recv.List) == 0 {
				continue
			}
			recv, pointer := receiverName(d.Recv.List[0].Type)
			s.Methods = append(s.Methods, Method{
				Receiver: recv,
				Pointer:  pointer,
				Name:     d.Name.Name,
				Doc:      docText(d.Doc),
			})
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch sp := spec.(type) {
				case *ast.TypeSpec:
					if !sp.Name.IsExported() {
						continue
					}
					s.Types = append(s.Types, Type{Name: sp.Name.Name, Doc: docText(sp.Doc)})
				case *ast.ValueSpec:
					for _, name := range sp.Names {
						if !name.IsExported() {
							continue
						}
						switch d.Tok {
						case token.CONST:
							s.Consts = append(s.Consts, Const{Name: name.Name, Doc: docText(sp.Doc)})
						case token.VAR:
							s.Vars = append(s.Vars, Var{
								Name: name.Name,
								Type: exprText(fset, sp.Type),
								Doc:  docText(sp.Doc),
							})
						}
					}
				}
			}
		}
	}
	return nil
}

func collectFunc(fset *token.FileSet, d *ast.FuncDecl) (Func, error) {
	fn := Func{Name: d.Name.Name, Doc: docText(d.Doc)}
	if d.Type.Params != nil {
		for _, field := range d.Type.Params.List {
			text := exprText(fset, field.Type)
			if _, ok := field.Type.(*ast.Ellipsis); ok {
				fn.Variadic = true
			}
			if len(field.Names) == 0 {
				fn.Params = append(fn.Params, Param{Type: text})
				continue
			}
			for _, name := range field.Names {
				fn.Params = append(fn.Params, Param{Name: name.Name, Type: text})
			}
		}
	}
	if d.Type.Results != nil {
		for _, field := range d.Type.Results.List {
			text := exprText(fset, field.Type)
			if len(field.Names) == 0 {
				fn.Results = append(fn.Results, Param{Type: text})
				continue
			}
			for _, name := range field.Names {
				fn.Results = append(fn.Results, Param{Name: name.Name, Type: text})
			}
		}
	}
	return fn, nil
}

func receiverName(expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		name, _ := receiverName(t.X)
		return name, true
	case *ast.Ident:
		return t.Name, false
	case *ast.IndexExpr:
		return receiverName(t.X)
	case *ast.IndexListExpr:
		return receiverName(t.X)
	}
	return "", false
}

func exprText(fset *token.FileSet, expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, expr); err != nil {
		return ""
	}
	return buf.String()
}

func docText(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return strings.TrimSpace(doc.Text())
}

func (s *Surface) sort() {
	sort.Slice(s.Funcs, func(i, j int) bool { return s.Funcs[i].Name < s.Funcs[j].Name })
	sort.Slice(s.Types, func(i, j int) bool { return s.Types[i].Name < s.Types[j].Name })
	sort.Slice(s.Vars, func(i, j int) bool { return s.Vars[i].Name < s.Vars[j].Name })
	sort.Slice(s.Consts, func(i, j int) bool { return s.Consts[i].Name < s.Consts[j].Name })
	sort.Slice(s.Methods, func(i, j int) bool {
		if s.Methods[i].Receiver != s.Methods[j].Receiver {
			return s.Methods[i].Receiver < s.Methods[j].Receiver
		}
		return s.Methods[i].Name < s.Methods[j].Name
	})
}

// Golden renders the surface in the stable textual form used by the snapshot
// file. Functions keep their full signature, types and variables keep their
// name only: the facade exposes aliases, so the underlying definition is
// pinned by the type identity rather than by the rendered shape.
func (s *Surface) Golden() string {
	var b strings.Builder
	writeSection := func(title string, lines []string) {
		b.WriteString("# " + title + "\n")
		for _, line := range lines {
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}
	var funcs []string
	for _, f := range s.Funcs {
		funcs = append(funcs, f.Signature())
	}
	writeSection("functions", funcs)
	var types []string
	for _, t := range s.Types {
		types = append(types, "type "+t.Name)
	}
	writeSection("types", types)
	var consts []string
	for _, c := range s.Consts {
		consts = append(consts, "const "+c.Name)
	}
	writeSection("constants", consts)
	var vars []string
	for _, v := range s.Vars {
		vars = append(vars, "var "+v.Name)
	}
	writeSection("variables", vars)
	var methods []string
	for _, m := range s.Methods {
		if m.Pointer {
			methods = append(methods, "(*"+m.Receiver+")."+m.Name)
			continue
		}
		methods = append(methods, m.Receiver+"."+m.Name)
	}
	writeSection("methods", methods)
	return b.String()
}

// WriteGolden writes the surface snapshot to path.
func WriteGolden(path string, s *Surface) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(s.Golden()), 0o644)
}

// ReadGolden parses a snapshot produced by Golden back into its sections.
func ReadGolden(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sections := map[string][]string{}
	current := ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "# ") {
			current = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			sections[current] = nil
			continue
		}
		if current == "" || strings.TrimSpace(line) == "" {
			continue
		}
		sections[current] = append(sections[current], line)
	}
	return sections, nil
}

// Compare checks a live surface against a golden snapshot. The methods are
// compared by name only: the facade exposes the original types through
// aliases, so the method set travels with the type identity and is verified
// at compile time by the generated method assertions.
func Compare(live *Surface, golden map[string][]string) []string {
	var diffs []string
	liveFuncs := map[string]bool{}
	for _, f := range live.Funcs {
		liveFuncs[f.Signature()] = true
	}
	for _, want := range golden["functions"] {
		if !liveFuncs[want] {
			diffs = append(diffs, "missing function: "+want)
		}
	}
	for _, have := range live.Funcs {
		found := false
		for _, want := range golden["functions"] {
			if want == have.Signature() {
				found = true
				break
			}
		}
		if !found {
			diffs = append(diffs, "unexpected function: "+have.Signature())
		}
	}
	liveNames := func(kind string, names []string) map[string]bool {
		set := map[string]bool{}
		for _, n := range names {
			set[kind+" "+n] = true
		}
		return set
	}
	liveTypes := liveNames("type", func() []string {
		out := []string{}
		for _, t := range live.Types {
			out = append(out, t.Name)
		}
		return out
	}())
	liveConsts := liveNames("const", func() []string {
		out := []string{}
		for _, c := range live.Consts {
			out = append(out, c.Name)
		}
		return out
	}())
	liveVars := liveNames("var", func() []string {
		out := []string{}
		for _, v := range live.Vars {
			out = append(out, v.Name)
		}
		return out
	}())
	for _, section := range []struct {
		name string
		live map[string]bool
	}{
		{"types", liveTypes},
		{"constants", liveConsts},
		{"variables", liveVars},
	} {
		for _, want := range golden[section.name] {
			if !section.live[want] {
				diffs = append(diffs, "missing "+section.name+": "+strings.TrimPrefix(want, section.name+" "))
			}
		}
		for have := range section.live {
			found := false
			for _, want := range golden[section.name] {
				if want == have {
					found = true
					break
				}
			}
			if !found {
				diffs = append(diffs, "unexpected "+section.name+": "+strings.TrimPrefix(have, section.name+" "))
			}
		}
	}
	return diffs
}
