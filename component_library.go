package main

// External component library.
//
// The designer's Insert catalog can be extended by a JSON file that external
// tools edit. Loading is atomic and validated: schema, identifiers, allowed
// tags/attributes, CSS declarations, tree size and a real compiler check per
// component. Any failure rejects the whole library and the editor keeps using
// the built-in catalog, so a broken file can never break the designer.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	componentLibraryFormat  = "simpleui-component-library"
	componentLibraryVersion = 1
	componentLibraryMaxSize = 1 << 20 // 1 MB
	componentMaxDepth       = 6
	componentMaxNodes       = 100
	componentMaxGroups      = 20
	componentMaxComponents  = 200
)

var (
	libraryIDRe    = regexp.MustCompile(`^[a-z][a-z0-9-]{0,39}$`)
	libraryClassRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]{0,59}$`)
	libraryNameRe  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	libraryTokenRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	libraryVarRe   = regexp.MustCompile(`^--[A-Za-z0-9_-]+$`)
)

// ComponentLibraryJSON is the on-disk and wire shape.
type ComponentLibraryJSON struct {
	Format  string               `json:"format"`
	Version int                  `json:"version"`
	Groups  []ComponentGroupJSON `json:"groups"`
}

type ComponentGroupJSON struct {
	ID         string               `json:"id"`
	Label      string               `json:"label"`
	Components []ComponentEntryJSON `json:"components"`
}

type ComponentEntryJSON struct {
	ID          string              `json:"id"`
	Label       string              `json:"label"`
	Kind        string              `json:"kind"`
	Description string              `json:"description,omitempty"`
	Class       string              `json:"class,omitempty"`
	CSS         string              `json:"css,omitempty"`
	Override    bool                `json:"override,omitempty"`
	Default     bool                `json:"default,omitempty"`
	Text        string              `json:"text,omitempty"`
	Attrs       map[string]any      `json:"attrs,omitempty"`
	Children    []ComponentNodeJSON `json:"children,omitempty"`
}

type ComponentNodeJSON struct {
	Kind     string              `json:"kind"`
	Text     string              `json:"text,omitempty"`
	Class    string              `json:"class,omitempty"`
	Attrs    map[string]any      `json:"attrs,omitempty"`
	Children []ComponentNodeJSON `json:"children,omitempty"`
}

var libraryKinds = map[string]bool{
	"button": true, "input": true, "textarea": true, "select": true, "checkbox": true,
	"switch": true, "radio": true, "number": true, "slider": true, "color": true,
	"date": true, "time": true, "combobox": true, "datepicker": true,
	"div": true, "row": true, "col": true, "card": true, "section": true, "header": true,
	"footer": true, "nav": true, "main": true, "article": true, "aside": true,
	"form": true, "fieldset": true, "legend": true,
	"text": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"p": true, "strong": true, "em": true, "code": true, "mark": true, "small": true,
	"blockquote": true, "pre": true, "hr": true, "divider": true, "spacer": true,
	"badge": true, "alert": true, "progress": true, "spinner": true,
	"tabs": true, "tab": true, "accordion": true, "accordionitem": true,
	"dropdown": true, "dropdownitem": true, "popover": true, "modal": true,
	"modalbutton": true, "toast": true, "list": true, "ul": true, "ol": true,
	"menubar": true, "menu": true, "menuitem": true, "drawer": true, "splitter": true,
	"pane": true, "dropzone": true, "timeline": true, "timelineitem": true,
	"emptystate": true, "skeleton": true, "carousel": true, "slide": true,
	"stepper": true, "step": true, "commandbar": true, "command": true,
	"tree": true, "treeitem": true,
}

var libraryContainers = map[string]bool{
	"div": true, "row": true, "col": true, "card": true, "section": true, "header": true,
	"footer": true, "nav": true, "main": true, "article": true, "aside": true,
	"form": true, "fieldset": true, "tabs": true, "tab": true, "accordion": true,
	"accordionitem": true, "dropdown": true, "popover": true, "modal": true,
	"toast": true, "list": true, "ul": true, "ol": true, "menubar": true, "menu": true,
	"drawer": true, "splitter": true, "pane": true, "dropzone": true, "timeline": true,
	"carousel": true, "stepper": true, "commandbar": true, "tree": true,
}

var libraryAttrType = map[string]string{
	"placeholder": "string",
	"value":       "string",
	"checked":     "bool",
	"disabled":    "bool",
	"name":        "string",
	"options":     "string",
	"min":         "number",
	"max":         "number",
	"step":        "number",
	"open":        "bool",
	"shape":       "string",
	"theme":       "string",
	"variant":     "string",
	"size":        "string",
}

var libraryCssProps = map[string]bool{
	"color": true, "background": true, "background-color": true,
	"background-size": true, "background-position": true, "background-repeat": true,
	"border": true, "border-width": true, "border-style": true, "border-color": true,
	"border-top": true, "border-right": true, "border-bottom": true, "border-left": true,
	"border-top-width": true, "border-right-width": true, "border-bottom-width": true,
	"border-left-width": true, "border-top-color": true, "border-right-color": true,
	"border-bottom-color": true, "border-left-color": true, "border-top-style": true,
	"border-right-style": true, "border-bottom-style": true, "border-left-style": true,
	"border-radius": true, "border-top-left-radius": true, "border-top-right-radius": true,
	"border-bottom-left-radius": true, "border-bottom-right-radius": true,
	"padding": true, "padding-top": true, "padding-right": true, "padding-bottom": true,
	"padding-left": true, "margin": true, "margin-top": true, "margin-right": true,
	"margin-bottom": true, "margin-left": true, "gap": true, "row-gap": true,
	"column-gap": true, "display": true, "flex": true, "flex-direction": true,
	"flex-wrap": true, "flex-grow": true, "flex-shrink": true, "flex-basis": true,
	"align-items": true, "align-content": true, "align-self": true,
	"justify-content": true, "justify-items": true, "justify-self": true,
	"grid": true, "grid-template-columns": true, "grid-template-rows": true,
	"grid-auto-flow": true, "grid-column": true, "grid-row": true,
	"place-items": true, "place-content": true, "width": true, "height": true,
	"min-width": true, "min-height": true, "max-width": true, "max-height": true,
	"box-sizing": true, "font": true, "font-family": true, "font-size": true,
	"font-weight": true, "font-style": true, "line-height": true,
	"letter-spacing": true, "text-align": true, "text-decoration": true,
	"text-transform": true, "white-space": true, "text-overflow": true,
	"opacity": true, "overflow": true, "overflow-x": true, "overflow-y": true,
	"object-fit": true, "object-position": true, "aspect-ratio": true,
	"box-shadow": true, "text-shadow": true, "transform": true,
	"transform-origin": true, "transition": true, "transition-property": true,
	"transition-duration": true, "transition-timing-function": true,
	"transition-delay": true, "cursor": true, "outline": true, "outline-width": true,
	"outline-style": true, "outline-color": true, "outline-offset": true,
	"vertical-align": true, "list-style": true, "list-style-type": true,
	"position": true, "z-index": true, "visibility": true, "float": true,
}

// loadComponentLibraryPayload reads, validates and serializes the library for
// the designer. Missing file is not an error: it returns an empty catalog with
// a warning so the editor uses the built-in components.
func loadComponentLibraryPayload(projectDir string) []byte {
	path := componentLibraryPath(projectDir)
	if path == "" {
		return libraryPayload(false, "", nil, nil, "no external component library found; using built-in components")
	}
	groups, warnings, err := loadComponentLibraryFile(path)
	if err != nil {
		b, _ := json.Marshal(map[string]any{
			"ok":     false,
			"source": path,
			"error":  err.Error(),
		})
		return b
	}
	return libraryPayload(true, path, groups, warnings, "")
}

func libraryPayload(ok bool, source string, groups []ComponentGroupJSON, warnings []string, warning string) []byte {
	if groups == nil {
		groups = []ComponentGroupJSON{}
	}
	if warnings == nil {
		warnings = []string{}
	}
	if warning != "" {
		warnings = append(warnings, warning)
	}
	b, _ := json.Marshal(map[string]any{
		"ok":       ok,
		"source":   source,
		"groups":   groups,
		"warnings": warnings,
	})
	return b
}

func componentLibraryPath(projectDir string) string {
	if projectDir != "" {
		p := filepath.Join(projectDir, "components", "library.json")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	p := filepath.Join("components", "library.json")
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}

func loadComponentLibraryFile(path string) ([]ComponentGroupJSON, []string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read %s: %v", path, err)
	}
	if st.Size() > componentLibraryMaxSize {
		return nil, nil, fmt.Errorf("%s is larger than %d bytes", path, componentLibraryMaxSize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read %s: %v", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var lib ComponentLibraryJSON
	if err := dec.Decode(&lib); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON: %v", err)
	}
	if dec.More() {
		return nil, nil, fmt.Errorf("unexpected data after the library object")
	}
	if lib.Format != componentLibraryFormat {
		return nil, nil, fmt.Errorf("format must be %q", componentLibraryFormat)
	}
	if lib.Version != componentLibraryVersion {
		return nil, nil, fmt.Errorf("unsupported version %d (expected %d)", lib.Version, componentLibraryVersion)
	}
	if len(lib.Groups) == 0 {
		return nil, nil, fmt.Errorf("the library has no groups")
	}
	if len(lib.Groups) > componentMaxGroups {
		return nil, nil, fmt.Errorf("too many groups (max %d)", componentMaxGroups)
	}
	warnings := []string{}
	seenGroups := map[string]bool{}
	seenComponents := map[string]bool{}
	total := 0
	for gi := range lib.Groups {
		g := &lib.Groups[gi]
		if !libraryIDRe.MatchString(g.ID) {
			return nil, nil, fmt.Errorf("group %d: invalid id %q (use lowercase letters, digits and -)", gi+1, g.ID)
		}
		if seenGroups[g.ID] {
			return nil, nil, fmt.Errorf("duplicate group id %q", g.ID)
		}
		seenGroups[g.ID] = true
		if strings.TrimSpace(g.Label) == "" || len(g.Label) > 60 {
			return nil, nil, fmt.Errorf("group %q: label must be 1..60 characters", g.ID)
		}
		if len(g.Components) == 0 {
			return nil, nil, fmt.Errorf("group %q has no components: every control type needs at least one version", g.ID)
		}
		for ci := range g.Components {
			c := &g.Components[ci]
			total++
			if total > componentMaxComponents {
				return nil, nil, fmt.Errorf("too many components (max %d)", componentMaxComponents)
			}
			if !libraryIDRe.MatchString(c.ID) {
				return nil, nil, fmt.Errorf("component %d/%d: invalid id %q", gi+1, ci+1, c.ID)
			}
			if seenComponents[c.ID] {
				return nil, nil, fmt.Errorf("duplicate component id %q", c.ID)
			}
			seenComponents[c.ID] = true
			if strings.TrimSpace(c.Label) == "" || len(c.Label) > 60 {
				return nil, nil, fmt.Errorf("component %q: label must be 1..60 characters", c.ID)
			}
			if len(c.Description) > 240 {
				return nil, nil, fmt.Errorf("component %q: description is longer than 240 characters", c.ID)
			}
			if !libraryKinds[c.Kind] {
				return nil, nil, fmt.Errorf("component %q: unsupported kind %q", c.ID, c.Kind)
			}
			if c.Class != "" && !libraryClassRe.MatchString(c.Class) {
				return nil, nil, fmt.Errorf("component %q: invalid class %q", c.ID, c.Class)
			}
			if strings.HasPrefix(c.Class, "sui-") {
				return nil, nil, fmt.Errorf("component %q: class cannot start with sui-", c.ID)
			}
			if len(c.Text) > 300 {
				return nil, nil, fmt.Errorf("component %q: text is longer than 300 characters", c.ID)
			}
			if err := validateLibraryAttrs(c.ID, c.Attrs); err != nil {
				return nil, nil, err
			}
			if c.CSS != "" {
				if err := validateLibraryCss(c.ID, c.Class, c.CSS); err != nil {
					return nil, nil, err
				}
			}
			if len(c.Children) > 0 {
				if !libraryContainers[c.Kind] {
					return nil, nil, fmt.Errorf("component %q: kind %q cannot contain children", c.ID, c.Kind)
				}
				count := 0
				if err := validateLibraryNodes(c.ID, c.Children, 1, &count); err != nil {
					return nil, nil, err
				}
			}
			src := libraryComponentSource(c)
			prog, dl := ParseSource(src, "components/library.json")
			if dl.HasErrors() {
				return nil, nil, fmt.Errorf("component %q does not compile: %s", c.ID, firstLibraryError(&dl))
			}
			dl = Check(prog)
			if dl.HasErrors() {
				return nil, nil, fmt.Errorf("component %q does not type-check: %s", c.ID, firstLibraryError(&dl))
			}
		}
	}
	return lib.Groups, warnings, nil
}

func firstLibraryError(dl *DiagList) string {
	for _, d := range dl.list {
		if d.kind == "warning" {
			continue
		}
		if d.line > 0 {
			return fmt.Sprintf("line %d:%d: %s", d.line, d.col, d.msg)
		}
		return d.msg
	}
	return "unknown error"
}

func validateLibraryAttrs(id string, attrs map[string]any) error {
	if len(attrs) > 20 {
		return fmt.Errorf("component %q: too many attributes (max 20)", id)
	}
	for k, v := range attrs {
		want, allowed := libraryAttrType[k]
		if !allowed {
			return fmt.Errorf("component %q: attribute %q is not allowed", id, k)
		}
		switch want {
		case "string":
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("component %q: attribute %q must be a string", id, k)
			}
			if len(s) > 500 {
				return fmt.Errorf("component %q: attribute %q is too long", id, k)
			}
			if k == "name" && !libraryNameRe.MatchString(s) {
				return fmt.Errorf("component %q: attribute name must be an identifier", id)
			}
			if (k == "shape" || k == "theme" || k == "variant" || k == "size") && s != "" && !libraryTokenRe.MatchString(s) {
				return fmt.Errorf("component %q: attribute %q must use letters, digits, _ and -", id, k)
			}
			if k == "options" && strings.ContainsAny(s, "\n\r{}") {
				return fmt.Errorf("component %q: options cannot contain braces or newlines", id)
			}
		case "bool":
			if _, ok := v.(bool); !ok {
				return fmt.Errorf("component %q: attribute %q must be true or false", id, k)
			}
		case "number":
			if _, ok := v.(float64); !ok {
				return fmt.Errorf("component %q: attribute %q must be a number", id, k)
			}
		}
	}
	return nil
}

func validateLibraryNodes(id string, nodes []ComponentNodeJSON, depth int, count *int) error {
	if depth > componentMaxDepth {
		return fmt.Errorf("component %q: children are nested too deeply (max %d)", id, componentMaxDepth)
	}
	for i := range nodes {
		n := &nodes[i]
		*count = *count + 1
		if *count > componentMaxNodes {
			return fmt.Errorf("component %q: too many child nodes (max %d)", id, componentMaxNodes)
		}
		if !libraryKinds[n.Kind] {
			return fmt.Errorf("component %q: child %d has unsupported kind %q", id, i+1, n.Kind)
		}
		if n.Class != "" && !libraryClassRe.MatchString(n.Class) {
			return fmt.Errorf("component %q: child %d has an invalid class", id, i+1)
		}
		if strings.HasPrefix(n.Class, "sui-") {
			return fmt.Errorf("component %q: child class cannot start with sui-", id)
		}
		if len(n.Text) > 300 {
			return fmt.Errorf("component %q: child %d text is too long", id, i+1)
		}
		if err := validateLibraryAttrs(id, n.Attrs); err != nil {
			return err
		}
		if len(n.Children) > 0 {
			if !libraryContainers[n.Kind] {
				return fmt.Errorf("component %q: child kind %q cannot contain children", id, n.Kind)
			}
			if err := validateLibraryNodes(id, n.Children, depth+1, count); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateLibraryCss(id, class, css string) error {
	if class == "" {
		return fmt.Errorf("component %q: css requires a class", id)
	}
	if len(css) > 4000 {
		return fmt.Errorf("component %q: css is longer than 4000 characters", id)
	}
	if strings.Contains(css, "/*") || strings.Contains(css, "*/") {
		return fmt.Errorf("component %q: css comments are not allowed", id)
	}
	if strings.Count(css, "{") != 1 || strings.Count(css, "}") != 1 {
		return fmt.Errorf("component %q: css must be a single rule", id)
	}
	open := strings.Index(css, "{")
	close := strings.LastIndex(css, "}")
	if open < 0 || close < open || strings.TrimSpace(css[close+1:]) != "" {
		return fmt.Errorf("component %q: css must be a single rule", id)
	}
	if strings.TrimSpace(css[:open]) != "."+class {
		return fmt.Errorf("component %q: css selector must be .%s", id, class)
	}
	for _, decl := range strings.Split(css[open+1:close], ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		colon := strings.Index(decl, ":")
		if colon <= 0 {
			return fmt.Errorf("component %q: malformed css declaration %q", id, decl)
		}
		prop := strings.ToLower(strings.TrimSpace(decl[:colon]))
		val := strings.TrimSpace(decl[colon+1:])
		if !libraryCssProps[prop] && !libraryVarRe.MatchString(prop) {
			return fmt.Errorf("component %q: css property %q is not allowed", id, prop)
		}
		if len(val) == 0 || len(val) > 300 {
			return fmt.Errorf("component %q: css value for %q is invalid", id, prop)
		}
		low := strings.ToLower(val)
		if strings.Contains(low, "url(") || strings.Contains(low, "@") ||
			strings.Contains(low, "expression") || strings.Contains(low, "javascript:") ||
			strings.Contains(low, "behavior:") || strings.ContainsAny(val, "<>\\;{}") {
			return fmt.Errorf("component %q: css value for %q is not allowed", id, prop)
		}
		if prop == "position" && low != "relative" && low != "static" {
			return fmt.Errorf("component %q: position must be relative or static", id)
		}
		if prop == "z-index" {
			var z int
			if _, err := fmt.Sscanf(low, "%d", &z); err == nil && z < 0 {
				return fmt.Errorf("component %q: negative z-index is not allowed", id)
			}
		}
	}
	return nil
}

func libraryComponentSource(c *ComponentEntryJSON) string {
	var b strings.Builder
	if c.CSS != "" {
		b.WriteString("css \"")
		b.WriteString(suiLibraryEscape(c.CSS))
		b.WriteString("\"\n\n")
	}
	b.WriteString("view main {\n")
	writeLibraryNode(&b, c.Kind, c.Text, c.Class, c.Attrs, c.Children, 1)
	b.WriteString("}\n")
	return b.String()
}

func writeLibraryNode(b *strings.Builder, kind, text, class string, attrs map[string]any, children []ComponentNodeJSON, depth int) {
	indent := strings.Repeat("\t", depth)
	b.WriteString(indent)
	b.WriteString(kind)
	if text != "" {
		b.WriteString(" \"")
		b.WriteString(suiLibraryEscape(text))
		b.WriteString("\"")
	}
	if class != "" {
		b.WriteString(" class: \"")
		b.WriteString(suiLibraryEscape(class))
		b.WriteString("\"")
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(" ")
		b.WriteString(k)
		b.WriteString(": ")
		switch v := attrs[k].(type) {
		case string:
			b.WriteString("\"")
			b.WriteString(suiLibraryEscape(v))
			b.WriteString("\"")
		case bool:
			if v {
				b.WriteString("true")
			} else {
				b.WriteString("false")
			}
		case float64:
			b.WriteString(strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", v), "0"), "."))
		}
	}
	if len(children) > 0 {
		b.WriteString(" {\n")
		for i := range children {
			writeLibraryNode(b, children[i].Kind, children[i].Text, children[i].Class, children[i].Attrs, children[i].Children, depth+1)
		}
		b.WriteString(indent)
		b.WriteString("}")
	}
	b.WriteString("\n")
}

func suiLibraryEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`, "{", `\{`, "}", `\}`)
	return r.Replace(s)
}

func libraryComponentByID(groups []ComponentGroupJSON, id string) *ComponentEntryJSON {
	for gi := range groups {
		for ci := range groups[gi].Components {
			if groups[gi].Components[ci].ID == id {
				return &groups[gi].Components[ci]
			}
		}
	}
	return nil
}

// libraryPreviewSource builds a small standalone app for the navigator. It adds
// a neutral preview background without changing the component itself.
func libraryPreviewSource(c *ComponentEntryJSON) string {
	css := c.CSS
	if css != "" {
		css += "\n"
	}
	css += "body { margin: 0; padding: 24px; background: #0f1117; color: #e5e7eb; font-family: system-ui, -apple-system, 'Segoe UI', sans-serif; }"
	var b strings.Builder
	b.WriteString("css \"")
	b.WriteString(suiLibraryEscape(css))
	b.WriteString("\"\n\nview main {\n")
	writeLibraryNode(&b, c.Kind, c.Text, c.Class, c.Attrs, c.Children, 1)
	b.WriteString("}\n")
	return b.String()
}
