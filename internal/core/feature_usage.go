package core

import "strings"

// Component feature detection.
//
// A generated app only needs the JavaScript runtime of the components it
// actually uses. The scanner walks the whole program (views, attributes,
// state initializers, functions and methods) and records the tags, attribute
// names and intrinsic module roots it finds, so generateJsFor can embed only
// the relevant runtimes. This keeps app.js small: the chart runtime alone is
// about 1.2 MB.

type progFeatures struct {
	tags  map[string]bool
	attrs map[string]bool
	apis  map[string]bool
	// toggleMotion records checkbox/switch/radio controls with an `animation`
	// attribute: they are animated by the motion runtime even though they do
	// not carry `motion`/`motionKey`.
	toggleMotion bool
	// fileUpload records http.postFiles/postFilesSafe: they read the files
	// selected through the drag-and-drop/file-picker runtime.
	fileUpload bool
	// fileRead records files.readBytes: it reads the bytes of a file in the
	// current selection, so the drag-and-drop/file-picker runtime is needed.
	fileRead bool
	// hyphenLangs is the union of the dictionaries requested by richtext tags
	// through `hyphenLangs`. Restricted is true only when every richtext tag
	// carries a literal selection, so no editor still needs the full set.
	hyphenLangs           map[string]bool
	hyphenLangsRestricted bool
}

func (f *progFeatures) hasTag(names ...string) bool {
	for _, name := range names {
		if f.tags[name] {
			return true
		}
	}
	return false
}

func (f *progFeatures) hasAttr(names ...string) bool {
	for _, name := range names {
		if f.attrs[name] {
			return true
		}
	}
	return false
}

func (f *progFeatures) hasAPI(names ...string) bool {
	for _, name := range names {
		if f.apis[name] {
			return true
		}
	}
	return false
}

// needsTable reports whether the data-table runtime is required. The trow /
// tcell family belongs to the same control, and the table.* core module calls
// into __SUI_TBL.
func (f *progFeatures) needsTable() bool {
	return f.hasTag("table", "caption", "thead", "tbody", "tfoot", "trow", "tcell") || f.hasAPI("table")
}

// needsDnd reports whether the drag-and-drop/file-picker runtime is required.
func (f *progFeatures) needsDnd() bool {
	return f.fileUpload || f.fileRead || f.hasTag("dropzone") ||
		f.hasAttr("draggable", "dragKey", "dropKey", "dropFiles", "pickFiles", "pickMode", "directory") ||
		f.hasAPI("drag")
}

// needsMotion reports whether the FLIP/enter/exit motion engine is required.
func (f *progFeatures) needsMotion() bool {
	return f.toggleMotion || f.hasAttr("motion", "motionExit", "motionDuration", "motionEasing", "motionDelay")
}

// needsMotionGraph reports whether the state-machine motion runtime is needed.
func (f *progFeatures) needsMotionGraph() bool {
	return f.hasAttr("motionGraph")
}

// needsMenu reports whether the menubar/dropdown runtime is needed. Dropdowns
// are menus with a different markup, so they share the same runtime.
func (f *progFeatures) needsMenu() bool {
	return f.hasTag("menubar", "menu", "menuitem", "dropdown", "dropdownitem")
}

func (f *progFeatures) needsTabs() bool {
	return f.hasTag("tabs", "tab")
}

func (f *progFeatures) needsAccordion() bool {
	return f.hasTag("accordion", "accordionitem")
}

func (f *progFeatures) needsToast() bool {
	return f.hasTag("toast") || f.hasAPI("toast")
}

func (f *progFeatures) needsTooltip() bool {
	return f.hasAttr("tooltip")
}

func (f *progFeatures) needsCombobox() bool {
	return f.hasTag("combobox")
}

func (f *progFeatures) needsDatePicker() bool {
	return f.hasTag("datepicker")
}

func (f *progFeatures) needsTree() bool {
	return f.hasTag("tree", "treeitem") || f.hasAPI("tree")
}

func (f *progFeatures) needsRichText() bool {
	return f.hasTag("richtext") || f.hasAPI("richtext")
}

func (f *progFeatures) needsChart() bool {
	return f.hasTag("chart", "graph") || f.hasAPI("chart")
}

func (f *progFeatures) needsModal() bool {
	return f.hasTag("modal", "modalbutton") || f.hasAPI("modal")
}

func (f *progFeatures) needsPopover() bool {
	return f.hasTag("popover")
}

func (f *progFeatures) needsDrawer() bool {
	return f.hasTag("drawer")
}

func (f *progFeatures) needsSplitter() bool {
	return f.hasTag("splitter", "pane")
}

func (f *progFeatures) needsStepper() bool {
	return f.hasTag("stepper", "step")
}

func (f *progFeatures) needsCarousel() bool {
	return f.hasTag("carousel", "slide")
}

func (f *progFeatures) needsCommandBar() bool {
	return f.hasTag("commandbar", "command")
}

func (f *progFeatures) needsDropzone() bool {
	return f.hasTag("dropzone")
}

func (f *progFeatures) needsTimeline() bool {
	return f.hasTag("timeline")
}

func (f *progFeatures) needsEmptyState() bool {
	return f.hasTag("emptystate")
}

func (f *progFeatures) needsSkeleton() bool {
	return f.hasTag("skeleton")
}

func (f *progFeatures) needsSpinner() bool {
	return f.hasTag("spinner")
}

// needsOverlay reports whether the shared anchored-overlay helper is required.
// Menu, combobox, tooltip, popover and command bar call __SUI_OVERLAY directly,
// so the helper must be emitted with any of them.
func (f *progFeatures) needsOverlay() bool {
	return f.needsMenu() || f.needsCombobox() || f.needsTooltip() || f.needsPopover() || f.needsCommandBar()
}

// scanProgFeatures collects the component features used by p. It never
// returns nil.
func scanProgFeatures(p *Prog) *progFeatures {
	f := &progFeatures{
		tags:                  map[string]bool{},
		attrs:                 map[string]bool{},
		apis:                  map[string]bool{},
		hyphenLangs:           map[string]bool{},
		hyphenLangsRestricted: true,
	}
	if p == nil {
		return f
	}

	var expr func(*Expr)
	expr = func(e *Expr) {
		if e == nil {
			return
		}
		// Intrinsic modules (table.*, tree.*, chart.*, richtext.*) and core
		// builtins (toast.show, modal.open, drag.*) are referenced as members
		// of a namespace identifier.
		if e.kind == "member" && e.obj != nil && e.obj.kind == "ident" {
			f.apis[e.obj.name] = true
			if e.obj.name == "http" && (e.memb == "postFiles" || e.memb == "postFilesSafe") {
				f.fileUpload = true
			}
			if e.obj.name == "files" && e.memb == "readBytes" {
				f.fileRead = true
			}
		}
		expr(e.obj)
		expr(e.target)
		expr(e.operand)
		expr(e.l)
		expr(e.r)
		expr(e.cond)
		expr(e.idx)
		expr(e.lo)
		expr(e.hi)
		for _, a := range e.args {
			expr(a)
		}
		for _, a := range e.elements {
			expr(a)
		}
		for _, v := range e.vals {
			expr(v)
		}
	}

	parts := func(ps []StrPart) {
		for i := 0; i < len(ps); i++ {
			if ps[i].kind == "expr" {
				expr(ps[i].expr)
			}
		}
	}

	var stmts func([]*Stmt, func(*Expr))
	stmts = func(list []*Stmt, scanExpr func(*Expr)) {
		for _, st := range list {
			if st == nil {
				continue
			}
			scanExpr(st.init)
			scanExpr(st.rhs)
			scanExpr(st.lvalue)
			scanExpr(st.cond)
			scanExpr(st.iterable)
			scanExpr(st.value)
			for _, e := range st.values {
				scanExpr(e)
			}
			stmts(st.then_, scanExpr)
			stmts(st.else_, scanExpr)
			stmts(st.body, scanExpr)
			stmts(st.default_, scanExpr)
			for _, c := range st.cases {
				scanExpr(c.expr)
				stmts(c.body, scanExpr)
			}
		}
	}

	var nodes func([]*ViewNode)
	nodes = func(list []*ViewNode) {
		for _, n := range list {
			if n == nil {
				continue
			}
			if n.kind == "element" && n.elem != nil {
				e := n.elem
				if e.tag != "" {
					f.tags[e.tag] = true
				}
				parts(e.text)
				foundHyphenLangs := false
				for ai := range e.attrs {
					a := &e.attrs[ai]
					if a.key != "" {
						f.attrs[a.key] = true
					}
					parts(a.text)
					if a.hasExpr {
						expr(a.expr)
					}
					if a.key == "animation" && (e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio") {
						f.toggleMotion = true
					}
					if e.tag == "richtext" && a.key == "hyphenLangs" {
						foundHyphenLangs = true
						if a.hasExpr || strPartsHaveExpr(a.text) {
							// A reactive selection cannot be resolved at build time.
							f.hyphenLangsRestricted = false
						} else {
							value, _ := staticPartsLiteral(a.text)
							value = strings.ToLower(strings.TrimSpace(value))
							if value == "" || value == "*" || value == "all" {
								f.hyphenLangsRestricted = false
							} else {
								for _, code := range strings.Split(value, ",") {
									code = strings.TrimSpace(code)
									if code != "" && code != "none" {
										f.hyphenLangs[code] = true
									}
								}
							}
						}
					}
				}
				if e.tag == "richtext" && !foundHyphenLangs {
					// No restriction on this editor: it can hyphenate in any language.
					f.hyphenLangsRestricted = false
				}
				nodes(childPtrs(e.children))
			}
			expr(n.cond)
			expr(n.iterable)
			expr(n.key)
			nodes(n.then_)
			nodes(n.else_)
			nodes(n.body)
		}
	}

	scanExpr := func(e *Expr) { expr(e) }

	for i := range p.consts {
		expr(p.consts[i].init)
	}
	for i := range p.computed {
		expr(p.computed[i].init)
	}
	for i := range p.states {
		expr(p.states[i].init)
	}
	for i := range p.funcs {
		fn := &p.funcs[i]
		for j := range fn.params {
			expr(fn.params[j].dflt)
		}
		stmts(fn.body, scanExpr)
	}
	for i := range p.views {
		v := &p.views[i]
		for j := range v.params {
			expr(v.params[j].dflt)
		}
		for j := range v.inst_state {
			expr(v.inst_state[j].init)
		}
		nodes(childPtrs(v.body))
		for j := range v.methods {
			m := &v.methods[j]
			for k := range m.params {
				expr(m.params[k].dflt)
			}
			stmts(m.body, scanExpr)
		}
	}
	return f
}
