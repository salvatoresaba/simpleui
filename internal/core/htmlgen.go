package core

import "encoding/json"
import "sort"
import "strconv"
import "strings"
import "sync/atomic"

// htmlgen.go — HTML/CSS code generation: index.html + style.css (SPEC §4
// "htmlgen" and §5 "Generated-web-app contract"). Owned by Agent D.
//
// This file only reads the Prog/ViewDecl/Element/Attr shapes from ast.go; it
// never validates tags or attributes (the typechecker owns that) and never
// panics. Interpolation parts in view titles and class values are ignored
// (their literal text is kept); the JS generator owns the reactive `${...}`
// output side of interpolation.

// The base stylesheet lives in css_parts.go as ordered, conditionally
// emitted slices (see writeBaseCss). Class names found in views are appended
// as pass-through rules by GenerateCss.

// concatTextParts joins the literal text parts of an interpolated string,
// skipping {expr} parts. Returns "" when there are no text parts.
func concatTextParts(parts []StrPart) string {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "text" {
			sb.WriteString(parts[i].text)
		}
	}
	return sb.String()
}

// htmlEscape escapes the characters that are special inside HTML text/attrs:
// & < > ". UTF-8 bytes pass through untouched. Escaping applies to generated
// text content only (PageTitle/text stay raw); separate concern.
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// PageTitle returns the literal text of the first view's title (interpolation
// parts are ignored), or "" when there is no view/title. No HTML escaping.
func PageTitle(p *Prog) string {
	if p == nil || len(p.views) == 0 {
		return ""
	}
	return concatTextParts(p.views[0].title)
}

// GenerateIndexHtml emits the index.html skeleton from SPEC §5 with the
// HTML-escaped PageTitle in <title>. A missing/empty title still emits an
// empty <title></title> element.
func GenerateIndexHtml(p *Prog) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>`)
	sb.WriteString(htmlEscape(PageTitle(p)))
	sb.WriteString(`</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
<div id="app"></div>
<script src="app.js"></script>
</body>
</html>
`)
	return sb.String()
}

// GenerateCss emits the base stylesheet plus one deduplicated (sorted)
// pass-through rule per literal class name found in any view element.
func GenerateCss(p *Prog) string {
	return generateCss(p, false)
}

// GenerateCssWithSourceMap builds the stylesheet with markers and returns the
// clean CSS plus a Source Map v3 document.
func GenerateCssWithSourceMap(p *Prog, sourceBase string) (string, string) {
	marked := generateCss(p, true)
	return SourceMapFromMarked(marked, SourceMapOptions{
		GeneratedFile:  "style.css",
		SourceBase:     sourceBase,
		IncludeContent: true,
	})
}

// writeBaseCss emits the base stylesheet slices used by feat, in the original
// order, so cascade and specificity are unchanged. Slices that only style an
// optional component are skipped when the program does not use it; the largest
// win is the chart stylesheet (about 105 KB).
func writeBaseCss(sb *strings.Builder, feat *progFeatures) {
	add := func(on bool, css string) {
		if on {
			sb.WriteString(css)
		}
	}
	add(true, suiCssFoundation)
	add(feat.needsCombobox(), suiCssCombobox)
	add(feat.needsDatePicker(), suiCssDatepicker)
	add(feat.needsTree(), suiCssTree)
	add(feat.needsChart(), suiCssChart)
	// Checkbox, radio, switch, slider, progress, badge, alert and divider
	// styles are part of the ordinary control vocabulary.
	add(true, suiCssControls)
	add(feat.needsSpinner(), suiCssSpinnerBase)
	add(feat.needsSpinner(), suiCssSpinnerSpin)
	add(feat.needsSpinner(), suiCssSpinnerVariants)
	add(true, suiCssBasicBlocks)
	add(feat.needsSpinner(), suiCssSpinnerImage)
	add(true, suiCssFoundation2)
	add(feat.needsTable(), suiCssTable)
	// The iframe error box is produced by the runtime for any iframe.
	add(true, suiCssIframeError)
	// The PDF viewer frame: the native viewer fills it.
	add(true, suiCssPdfViewer)
	add(feat.needsDnd(), suiCssDnd)
	add(feat.needsMenu() || feat.needsPopover(), suiCssMenuShared)
	add(feat.needsPopover(), suiCssPopover)
	add(feat.needsDrawer(), suiCssDrawer)
	add(feat.needsSplitter(), suiCssSplitter)
	add(feat.needsDropzone(), suiCssDropzone)
	add(feat.needsTimeline(), suiCssTimeline)
	add(feat.needsEmptyState(), suiCssEmptyState)
	add(feat.needsSkeleton(), suiCssSkeleton)
	add(feat.needsCarousel(), suiCssCarousel)
	add(feat.needsStepper(), suiCssStepper)
	add(feat.needsCommandBar(), suiCssCommandbar)
	add(feat.needsOverlay(), suiCssAnim)
	add(feat.needsMenu(), suiCssMenu)
	add(feat.needsTabs(), suiCssTabs)
	add(feat.needsAccordion(), suiCssAccordion)
	add(feat.needsToast(), suiCssToast)
	add(feat.needsTooltip(), suiCssTooltip)
	add(feat.needsModal(), suiCssModal)
}

func generateCss(p *Prog, markers bool) string {
	var sb strings.Builder
	writeBaseCss(&sb, scanProgFeatures(p))
	if progUsesRichText(p) {
		sb.WriteString(suiRichTextCss)
	}
	var classes []string
	if p != nil {
		for i := 0; i < len(p.views); i++ {
			collectClassesInView(p.views[i].body, &classes)
		}
	}
	sort.Strings(classes)
	prev := ""
	for i := 0; i < len(classes); i++ {
		c := classes[i]
		if c == "" || c == prev {
			continue
		}
		sb.WriteString("." + c + " { /* generated */ }\n")
		prev = c
	}
	if p != nil {
		for i := 0; i < len(p.css); i++ {
			sb.WriteString("\n/* user css */\n")
			if markers && i < len(p.cssSpans) {
				sb.WriteString(rawSrcMarker(p.cssSpans[i]))
			}
			sb.WriteString(p.css[i])
			sb.WriteString("\n")
		}
	}
	return applyTableTint(sb.String())
}

// applyTableTint keeps the designer-managed zebra/hover colors readable in
// dark themes. Older generated sources stored a solid light color (for example
// #f0f2f8) that painted a white row on a dark background; tint it instead.
func applyTableTint(css string) string {
	if !strings.Contains(css, "var(--t-zebra") && !strings.Contains(css, "var(--t-hover") {
		return css
	}
	if strings.Contains(css, "color-mix(in srgb, var(--t-zebra") || strings.Contains(css, "color-mix(in srgb, var(--t-hover") {
		return css
	}
	var b strings.Builder
	b.WriteString(css)
	if !strings.HasSuffix(css, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("table.suitbl tbody tr:nth-child(even) td { background: var(--t-zebra, transparent); background: color-mix(in srgb, var(--t-zebra, transparent) 14%, transparent); }\n")
	b.WriteString("table.suitbl tbody tr:hover td { background: var(--t-hover, transparent); background: color-mix(in srgb, var(--t-hover, transparent) 14%, transparent); }\n")
	return b.String()
}

// collectClassesInView walks the (dereferenced) nodes of a view body.
func collectClassesInView(nodes []ViewNode, out *[]string) {
	var ptrs []*ViewNode
	for i := 0; i < len(nodes); i++ {
		ptrs = append(ptrs, &nodes[i])
	}
	collectClassesPtrs(ptrs, out)
}

// collectClassesPtrs walks a []*ViewNode list, descending into element
// children, if/then-else branches and list bodies.
func collectClassesPtrs(nodes []*ViewNode, out *[]string) {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		if n == nil {
			continue
		}
		if n.kind == "element" {
			collectClassesElem(n.elem, out)
		} else if n.kind == "if" {
			collectClassesPtrs(n.then_, out)
			collectClassesPtrs(n.else_, out)
		} else if n.kind == "list" {
			collectClassesPtrs(n.body, out)
		}
	}
}

// collectClassesElem reads the element's own literal class attribute (if any)
// and then recurses into its children.
func collectClassesElem(e *Element, out *[]string) {
	if e == nil {
		return
	}
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		if a.key != "class" {
			continue
		}
		if !a.hasExpr {
			addClassTokens(concatTextParts(a.text), out)
		} else {
			collectExprClassTokens(a.expr, out)
		}
	}
	collectClassesInView(e.children, out)
}

// collectExprClassTokens collects the literal class tokens inside a class
// expression (templates, "a " + b, cond ? "on" : "off"). Dynamic values are
// filtered at runtime, but literals must remain usable without being declared
// in CSS.
func collectExprClassTokens(e *Expr, out *[]string) {
	if e == nil {
		return
	}
	switch e.kind {
	case "str":
		addClassTokens(concatTextParts(e.parts), out)
	case "binary":
		if e.op == "add" {
			collectExprClassTokens(e.l, out)
			collectExprClassTokens(e.r, out)
		}
	case "cond":
		collectExprClassTokens(e.l, out)
		collectExprClassTokens(e.r, out)
	}
}

// addClassTokens splits a class attribute value on whitespace and appends the
// tokens to out.
func addClassTokens(value string, out *[]string) {
	for _, t := range strings.Fields(value) {
		*out = append(*out, t)
	}
}

// GenerateHtmlPage renders the non-invoked views (pages) as a static HTML
// document (no JS): one .sui-view block per page, root view first. Used by the
// design tool's "HTML" tab as the human-readable source of the project.
func GenerateHtmlPage(p *Prog) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<title>")
	sb.WriteString(PageTitle(p))
	sb.WriteString("</title>\n</head>\n<body>\n<div id=\"app\">\n")
	if p != nil && len(p.views) > 0 {
		viewNames := map[string]bool{}
		for i := 0; i < len(p.views); i++ {
			viewNames[p.views[i].name] = true
		}
		invoked := map[string]bool{}
		for i := 0; i < len(p.views); i++ {
			collectTags(p.views[i].body, viewNames, invoked)
		}
		for i := 0; i < len(p.views); i++ {
			if invoked[p.views[i].name] {
				continue
			}
			sb.WriteString("  <div class=\"sui-view\" data-sui-view=\"" + p.views[i].name + "\">\n")
			writeStaticNodes(&sb, p.views[i].body, "    ")
			sb.WriteString("  </div>\n")
		}
	}
	sb.WriteString("</div>\n</body>\n</html>\n")
	return sb.String()
}

func writeStaticNodes(sb *strings.Builder, nodes []ViewNode, ind string) {
	for i := 0; i < len(nodes); i++ {
		if nodes[i].kind == "element" && nodes[i].elem != nil {
			writeStaticElem(sb, nodes[i].elem, ind)
		}
	}
}

// spinnerImageMode reports whether a spinner uses an image/GIF instead of the
// CSS ring. The source may be static text or a reactive expression.
func spinnerImageMode(e *Element) bool {
	if e == nil || e.tag != "spinner" {
		return false
	}
	a := findAttr(e, "src")
	if a == nil {
		return false
	}
	if a.hasExpr {
		return a.expr != nil
	}
	return strings.TrimSpace(partsContent(a.text)) != ""
}

// spinnerStaticImageMode is the literal-only variant used by the static
// renderer, which cannot evaluate expressions. A reactive src is rendered by
// the JavaScript generator instead.
func spinnerStaticImageMode(e *Element) bool {
	if e == nil || e.tag != "spinner" {
		return false
	}
	a := findAttr(e, "src")
	return a != nil && !a.hasExpr && strings.TrimSpace(partsContent(a.text)) != ""
}

// writeStaticSpinnerImage renders an image spinner as a wrapper with an <img>.
// The wrapper carries the control class so user styles keep working; the image
// inherits the size variables and never runs the CSS ring animation.
func writeStaticSpinnerImage(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-spinner sui-spinner-image\"")
	staticClassSuffix(sb, e)
	sb.WriteString(">")
	sb.WriteString("<img class=\"sui-spinner-img\"")
	if a := findAttr(e, "src"); a != nil && !a.hasExpr {
		sb.WriteString(" src=\"" + htmlEscape(strings.TrimSpace(partsContent(a.text))) + "\"")
	}
	if a := findAttr(e, "alt"); a != nil && !a.hasExpr {
		sb.WriteString(" alt=\"" + htmlEscape(partsContent(a.text)) + "\"")
	} else if len(e.text) > 0 {
		sb.WriteString(" alt=\"" + htmlEscape(partsContent(e.text)) + "\"")
	} else {
		sb.WriteString(" alt=\"\"")
	}
	sb.WriteString("></div>\n")
}

// staticDatalistSeq numbers the <datalist> elements emitted for static
// elements. Atomic: two concurrent HTTP handlers can generate HTML at the
// same time and a plain ++ would be a data race.
var staticDatalistSeq int64

func writeStaticElem(sb *strings.Builder, e *Element, ind string) {
	if e.tag == "spinner" && spinnerStaticImageMode(e) {
		writeStaticSpinnerImage(sb, e, ind)
		return
	}
	if e.tag == "dropdownitem" && datepickerBool(e, "separator") {
		sb.WriteString(ind + "<div class=\"sui-dropdown-sep\" role=\"separator\"></div>\n")
		return
	}
	if e.tag == "popover" {
		writeStaticPopover(sb, e, ind)
		return
	}
	if e.tag == "drawer" {
		writeStaticDrawer(sb, e, ind)
		return
	}
	if e.tag == "splitter" {
		writeStaticSplitter(sb, e, ind)
		return
	}
	if e.tag == "dropzone" {
		writeStaticDropzone(sb, e, ind)
		return
	}
	if e.tag == "pdfviewer" {
		writeStaticPdfViewer(sb, e, ind)
		return
	}
	if e.tag == "timeline" {
		writeStaticTimeline(sb, e, ind)
		return
	}
	if e.tag == "emptystate" {
		writeStaticEmptyState(sb, e, ind)
		return
	}
	if e.tag == "skeleton" {
		writeStaticSkeleton(sb, e, ind)
		return
	}
	if e.tag == "carousel" {
		writeStaticCarousel(sb, e, ind)
		return
	}
	if e.tag == "stepper" {
		writeStaticStepper(sb, e, ind)
		return
	}
	if e.tag == "commandbar" {
		writeStaticCommandbar(sb, e, ind)
		return
	}
	if e.tag == "modal" {
		writeStaticModal(sb, e, ind)
		return
	}
	if e.tag == "tabs" {
		writeStaticTabs(sb, e, ind)
		return
	}
	if e.tag == "accordion" {
		writeStaticAccordion(sb, e, ind)
		return
	}
	if e.tag == "toast" {
		writeStaticToastHost(sb, e, ind)
		return
	}
	if e.tag == "combobox" {
		writeStaticCombobox(sb, e, ind)
		return
	}
	if e.tag == "datepicker" {
		writeStaticDatePicker(sb, e, ind)
		return
	}
	if e.tag == "tree" {
		writeStaticTree(sb, e, ind)
		return
	}
	if e.tag == "tablelayout" {
		writeStaticTableLayout(sb, e, ind)
		return
	}
	if chartTag(e.tag) {
		writeStaticChart(sb, e, ind)
		return
	}
	if e.tag == "menubar" {
		writeStaticMenuBar(sb, e, ind)
		return
	}
	if e.tag == "menu" {
		writeStaticMenu(sb, e, ind)
		return
	}
	if e.tag == "dropdown" {
		writeStaticMenu(sb, e, ind)
		return
	}
	if e.tag == "dropdown" {
		writeStaticMenu(sb, e, ind)
		return
	}
	tag := elemHtmlTag(e.tag)
	if e.tag == "tcell" && isHeaderCell(e) {
		tag = "th"
	}
	var dlId string
	if e.tag == "input" || e.tag == "checkbox" || e.tag == "radio" || e.tag == "switch" || e.tag == "slider" || e.tag == "file" || e.tag == "color" || e.tag == "date" || e.tag == "time" || e.tag == "number" {
		if e.tag == "input" {
			if opts := inputOptions(e); len(opts) > 0 {
				n := atomic.AddInt64(&staticDatalistSeq, 1)
				dlId = "sui-dl-" + strconv.FormatInt(n, 10)
			}
		}
	}
	link := staticLinkOpen(e)
	if link != "" {
		sb.WriteString(ind)
		sb.WriteString(link)
		sb.WriteString("\n")
	}
	// checkbox/switch/radio render as input + label text inside a wrapper
	// <label>; the wrapper mirrors the element's class so user styling reaches
	// the whole widget.
	wrapBase := ""
	switch e.tag {
	case "checkbox":
		wrapBase = "sui-checkbox-wrap"
	case "radio":
		wrapBase = "sui-radio-wrap"
	case "switch":
		wrapBase = "sui-switch-wrap"
	}
	sb.WriteString(ind)
	if wrapBase != "" {
		sb.WriteString("<label class=\"" + wrapBase + "\"")
		staticClassSuffix(sb, e)
		sb.WriteString(">")
	}
	sb.WriteString("<")
	sb.WriteString(tag)
	switch e.tag {
	case "row", "col", "spacer":
		sb.WriteString(" class=\"")
		sb.WriteString(e.tag)
		staticClassSuffix(sb, e)
		sb.WriteString("\"")
		if e.tag == "row" || e.tag == "col" {
			layoutContainerStyleAttr(sb, e)
		}
	case "card":
		sb.WriteString(" class=\"sui-card\"")
		staticClassSuffix(sb, e)
	case "divider":
		sb.WriteString(" class=\"sui-divider\"")
		staticClassSuffix(sb, e)
	case "badge":
		sb.WriteString(" class=\"sui-badge\"")
		staticClassSuffix(sb, e)
	case "alert":
		sb.WriteString(" class=\"sui-alert\"")
		staticClassSuffix(sb, e)
	case "spinner":
		sb.WriteString(" class=\"sui-spinner\"")
		staticClassSuffix(sb, e)
	case "button":
		sb.WriteString(" class=\"sui-button\"")
		staticClassSuffix(sb, e)
		staticUserAttrsNoClass(sb, e, dlId)
	case "progress":
		sb.WriteString(" class=\"sui-progress\"")
		staticClassSuffix(sb, e)
	case "checkbox":
		sb.WriteString(" type=\"checkbox\" class=\"sui-checkbox" + togglePresetClasses(e) + "\"")
		staticClassSuffix(sb, e)
	case "radio":
		sb.WriteString(" type=\"radio\" class=\"sui-radio" + togglePresetClasses(e) + "\"")
		staticClassSuffix(sb, e)
	case "switch":
		sb.WriteString(" type=\"checkbox\" class=\"sui-switch" + togglePresetClasses(e) + "\"")
		staticClassSuffix(sb, e)
	case "slider":
		sb.WriteString(" type=\"range\" class=\"sui-slider\"")
		staticClassSuffix(sb, e)
	case "file":
		sb.WriteString(" type=\"file\" class=\"sui-file\"")
		staticClassSuffix(sb, e)
	case "color":
		sb.WriteString(" type=\"color\" class=\"sui-color\"")
		staticClassSuffix(sb, e)
	case "date":
		sb.WriteString(" type=\"date\" class=\"sui-date\"")
		staticClassSuffix(sb, e)
	case "time":
		sb.WriteString(" type=\"time\" class=\"sui-time\"")
		staticClassSuffix(sb, e)
	case "number":
		sb.WriteString(" type=\"number\" class=\"sui-number\"")
		staticClassSuffix(sb, e)
	case "textarea":
		sb.WriteString(" class=\"sui-textarea\"")
		staticClassSuffix(sb, e)
	case "select":
		sb.WriteString(" class=\"sui-select\"")
		staticClassSuffix(sb, e)
	case "menuitem":
		sb.WriteString(" class=\"sui-menuitem\"")
		staticClassSuffix(sb, e)
		sb.WriteString(" role=\"menuitem\" tabindex=\"-1\"")
		staticUserAttrsNoClass(sb, e, dlId)
	case "dropdownitem":
		sb.WriteString(" type=\"button\" class=\"sui-menuitem\"")
		staticClassSuffix(sb, e)
		sb.WriteString(" role=\"menuitem\" tabindex=\"-1\"")
		if datepickerBool(e, "disabled") {
			sb.WriteString(" disabled")
		}
		staticUserAttrsNoClass(sb, e, dlId)
	case "tab":
		sb.WriteString(" class=\"sui-tabpanel\"")
		staticClassSuffix(sb, e)
		staticUserAttrsNoClass(sb, e, dlId)
	case "modalbutton":
		sb.WriteString(" class=\"sui-modal-button")
		if v := modalVariant(e); v != "" {
			sb.WriteString(" " + v)
		}
		staticClassSuffix(sb, e)
		sb.WriteString("\"")
		staticUserAttrsNoClass(sb, e, dlId)
	default:
		staticUserAttrs(sb, e, dlId)
	}
	if e.tag == "checkbox" || e.tag == "radio" || e.tag == "switch" ||
		e.tag == "slider" || e.tag == "file" || e.tag == "color" ||
		e.tag == "date" || e.tag == "time" || e.tag == "number" ||
		e.tag == "textarea" || e.tag == "select" {
		staticUserAttrsNoClass(sb, e, dlId)
	}
	if e.tag == "table" {
		if v := treeTextAttr(e, "name"); v != "" {
			sb.WriteString(" name=\"" + htmlEscape(v) + "\"")
		}
		tableEditMode := treeTextAttr(e, "editMode")
		if tableEditMode == "editable" || tableEditMode == "persistent" {
			sb.WriteString(" data-sui-editable=\"1\"")
		}
		if tableEditMode != "" {
			sb.WriteString(" data-sui-edit-mode=\"" + htmlEscape(tableEditMode) + "\"")
		}
		if v := treeTextAttr(e, "cellText"); v != "" {
			sb.WriteString(" data-sui-cell-text=\"" + htmlEscape(v) + "\"")
			if v == "truncate" {
				sb.WriteString(" data-sui-fixed=\"1\"")
			}
		}
		for _, pair := range [][2]string{
			{"columns", "data-sui-columns"}, {"columnLabels", "data-sui-column-labels"},
			{"columnWidths", "data-sui-column-widths"}, {"columnAlign", "data-sui-column-align"},
			{"columnTypes", "data-sui-column-types"}, {"columnFormats", "data-sui-column-formats"},
			{"columnOptions", "data-sui-column-options"},
		} {
			if v := treeTextAttr(e, pair[0]); v != "" {
				sb.WriteString(" " + pair[1] + "=\"" + htmlEscape(v) + "\"")
			}
		}
		if datepickerBool(e, "sortable") {
			sb.WriteString(" data-sui-sortable=\"1\"")
		}
		if datepickerBool(e, "filterable") {
			sb.WriteString(" data-sui-filterable=\"1\"")
		}
		if datepickerBool(e, "multiSelect") {
			sb.WriteString(" data-sui-multi=\"1\"")
		}
		if datepickerBool(e, "virtual") {
			sb.WriteString(" data-sui-virtual=\"1\"")
		}
		if datepickerBool(e, "showAddTable") {
			sb.WriteString(" data-sui-show-add-table=\"1\"")
		}
		if datepickerBool(e, "structureEditable") {
			sb.WriteString(" data-sui-structure-editable=\"1\"")
		}
		if v := comboboxAttrNumber(e, "pageSize"); v != "" {
			sb.WriteString(" data-sui-page-size=\"" + htmlEscape(v) + "\"")
		}
		if v := treeTextAttr(e, "selection"); v != "" {
			sb.WriteString(" data-sui-selection=\"" + htmlEscape(v) + "\"")
		}
		for _, pair := range [][2]string{
			{"onCellChange", "data-sui-table-oncellchange"}, {"onRowChange", "data-sui-table-onrowchange"},
			{"onTableChange", "data-sui-table-ontablechange"}, {"onRowSelect", "data-sui-table-onrowselect"},
			{"onSort", "data-sui-table-onsort"}, {"onCellClick", "data-sui-table-oncellclick"},
			{"onRowClick", "data-sui-table-onrowclick"},
		} {
			if name := eventIdentOf(e, pair[0]); name != "" {
				sb.WriteString(" " + pair[1] + "=\"" + htmlEscape(name) + "\"")
			}
		}
	}
	sb.WriteString(">")
	if e.tag == "input" || e.tag == "img" || e.tag == "hr" || e.tag == "br" || e.tag == "checkbox" || e.tag == "radio" || e.tag == "switch" || e.tag == "slider" || e.tag == "file" || e.tag == "color" || e.tag == "date" || e.tag == "time" || e.tag == "number" || e.tag == "divider" || e.tag == "spinner" || e.tag == "progress" {
		// HTML void tags (input/img/hr/br) close naturally; "spinner" and
		// "progress" map to <div>/<progress>, which are NOT void: without an
		// explicit close tag every following element nests inside them.
		if tag != "input" && tag != "img" && tag != "hr" && tag != "br" {
			sb.WriteString("</")
			sb.WriteString(tag)
			sb.WriteString(">")
		}
		if wrapBase != "" {
			if len(e.text) > 0 {
				sb.WriteString("<span>")
				staticText(sb, e.text)
				sb.WriteString("</span>")
			}
			sb.WriteString("</label>")
		}
		if dlId != "" {
			sb.WriteString(datalistHtml(dlId, inputOptions(e)))
		}
		sb.WriteString("\n")
		if link != "" {
			sb.WriteString(ind)
			sb.WriteString("</a>\n")
		}
		return
	}
	staticText(sb, e.text)
	if len(e.children) > 0 {
		sb.WriteString("\n")
		writeStaticNodes(sb, e.children, ind+"  ")
		sb.WriteString(ind)
	}
	sb.WriteString("</")
	sb.WriteString(tag)
	sb.WriteString(">\n")
	if link != "" {
		sb.WriteString(ind)
		sb.WriteString("</a>\n")
	}
}

func modalNameOf(e *Element) string {
	a := findAttr(e, "name")
	if a == nil || a.hasExpr {
		return ""
	}
	return strings.TrimSpace(partsContent(a.text))
}

func modalVariant(e *Element) string {
	a := findAttr(e, "variant")
	if a == nil || a.hasExpr {
		return ""
	}
	v := strings.TrimSpace(partsContent(a.text))
	switch v {
	case "primary", "danger":
		return v
	}
	return ""
}

func modalButtonCloses(e *Element) bool {
	a := findAttr(e, "close")
	if a == nil {
		return true
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		return a.expr.b
	}
	return strings.TrimSpace(strings.ToLower(partsContent(a.text))) != "false"
}

// writeStaticTableLayout renders a structural table: each direct child is
// placed in its own <td>, flowing in a fixed number of columns.
func writeStaticTableLayout(sb *strings.Builder, e *Element, ind string) {
	grid := layoutTableGrid(e)
	style := layoutTableStyleWithUser(e)
	sb.WriteString(ind)
	sb.WriteString("<table class=\"sui-layout-table")
	staticClassSuffix(sb, e)
	if datepickerBool(e, "responsive") {
		sb.WriteString(" sui-layout-responsive")
	}
	sb.WriteString("\" role=\"")
	if datepickerBool(e, "semantic") {
		sb.WriteString("table")
	} else {
		sb.WriteString("presentation")
	}
	sb.WriteString("\"")
	if label := layoutTableTextAttr(e, "label", ""); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	if style != "" {
		sb.WriteString(" style=\"")
		sb.WriteString(htmlEscape(style))
		sb.WriteString("\"")
	}
	sb.WriteString(" data-sui-layout-cols=\"")
	sb.WriteString(strconv.Itoa(grid.cols))
	sb.WriteString("\" data-sui-layout-rows=\"")
	sb.WriteString(strconv.Itoa(grid.rows))
	sb.WriteString("\"")
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	if colgroup := layoutColgroupHTML(e, grid.cols); colgroup != "" {
		sb.WriteString(ind)
		sb.WriteString("  ")
		sb.WriteString(colgroup)
		sb.WriteString("\n")
	}
	sb.WriteString(ind)
	sb.WriteString("  <tbody>\n")
	for r := 0; r < grid.rows; r++ {
		sb.WriteString(ind)
		sb.WriteString("    <tr")
		sb.WriteString(layoutRowStyle(e, r))
		sb.WriteString(">\n")
		for c := 0; c < grid.cols; c++ {
			if !grid.occupied[r][c] {
				sb.WriteString(ind)
				sb.WriteString("      <td class=\"sui-layout-cell sui-layout-cell-empty\"")
				sb.WriteString(layoutCellPosAttrs(r, c))
				sb.WriteString(layoutEmptyCellAttrs(e))
				sb.WriteString("></td>\n")
				continue
			}
			ci, ok := grid.starts[[2]int{r, c}]
			if !ok {
				continue
			}
			cell := grid.cells[ci]
			sb.WriteString(ind)
			sb.WriteString("      <td class=\"sui-layout-cell")
			sb.WriteString(layoutCellClass(cell))
			if cell.node != nil && cell.node.kind == "element" && cell.node.elem != nil {
				staticClassSuffix(sb, cell.node.elem)
			}
			sb.WriteString("\"")
			sb.WriteString(layoutCellPosAttrs(r, c))
			sb.WriteString(layoutCellAttrs(cell, e, false))
			sb.WriteString(">\n")
			writeStaticNodes(sb, cell.content, ind+"        ")
			sb.WriteString(ind)
			sb.WriteString("      </td>\n")
		}
		sb.WriteString(ind)
		sb.WriteString("    </tr>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("  </tbody>\n")
	sb.WriteString(ind)
	sb.WriteString("</table>\n")
}

func writeStaticModal(sb *strings.Builder, e *Element, ind string) {
	name := modalNameOf(e)
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-modal")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-modal=\"")
	sb.WriteString(htmlEscape(name))
	sb.WriteString("\" aria-hidden=\"true\" hidden>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-modal-backdrop\" data-sui-modal-backdrop></div>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-modal-panel\" role=\"dialog\" aria-modal=\"true\" tabindex=\"-1\"")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	sb.WriteString(">\n")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(ind)
		sb.WriteString("    <div class=\"sui-modal-title\">")
		staticText(sb, e.text)
		sb.WriteString("</div>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("    <div class=\"sui-modal-body\">\n")
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "modalbutton" {
			continue
		}
		writeStaticNodes(sb, []ViewNode{*ch}, ind+"      ")
	}
	sb.WriteString(ind)
	sb.WriteString("    </div>\n")
	sb.WriteString(ind)
	sb.WriteString("    <div class=\"sui-modal-actions\">\n")
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind != "element" || ch.elem == nil || ch.elem.tag != "modalbutton" {
			continue
		}
		btn := ch.elem
		sb.WriteString(ind)
		sb.WriteString("      <button type=\"button\" class=\"sui-modal-button")
		if v := modalVariant(btn); v != "" {
			sb.WriteString(" " + v)
		}
		staticClassSuffix(sb, btn)
		sb.WriteString("\"")
		if modalButtonCloses(btn) {
			sb.WriteString(" data-sui-modal-dismiss=\"1\"")
		}
		sb.WriteString(">")
		staticText(sb, btn.text)
		sb.WriteString("</button>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("    </div>\n")
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

// tabsRuntimeKey returns the stable key shared by the static markup and the
// dynamic re-render, so the tabs runtime keeps the selected tab across both.
func tabsRuntimeKey(e *Element) string {
	if e == nil {
		return ""
	}
	return "sui-tabs-" + strconv.Itoa(e.span.line) + "-" + strconv.Itoa(e.span.col)
}

// writeStaticTabs emits a tablist plus tabpanels. The first tab is active in
// static HTML; the runtime keeps the selected tab across app re-renders.
func writeStaticTabs(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-tabs")
	staticClassSuffix(sb, e)
	sb.WriteString("\"")
	if v := treeTextAttr(e, "position"); v == "bottom" {
		sb.WriteString(" data-sui-tabs-position=\"bottom\"")
	}
	if v := treeTextAttr(e, "variant"); v == "sheets" {
		sb.WriteString(" data-sui-tabs-variant=\"sheets\"")
	}
	if key := tabsRuntimeKey(e); key != "" {
		sb.WriteString(" data-sui-tabs-key=\"")
		sb.WriteString(htmlEscape(key))
		sb.WriteString("\"")
	}
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	var tabs []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "tab" {
			tabs = append(tabs, ch.elem)
		}
	}
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-tablist\" role=\"tablist\">\n")
	for i, t := range tabs {
		active := i == 0
		sb.WriteString(ind)
		sb.WriteString("    <button type=\"button\" class=\"sui-tab")
		if active {
			sb.WriteString(" active")
		}
		sb.WriteString("\" role=\"tab\" aria-selected=\"")
		if active {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		sb.WriteString("\" tabindex=\"")
		if active {
			sb.WriteString("0")
		} else {
			sb.WriteString("-1")
		}
		sb.WriteString("\">")
		staticText(sb, t.text)
		sb.WriteString("</button>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-tabpanels\">\n")
	for i, t := range tabs {
		active := i == 0
		sb.WriteString(ind)
		sb.WriteString("    <div class=\"sui-tabpanel")
		staticClassSuffix(sb, t)
		if active {
			sb.WriteString(" active")
		}
		sb.WriteString("\" role=\"tabpanel\"")
		if !active {
			sb.WriteString(" hidden")
		}
		sb.WriteString(">\n")
		writeStaticNodes(sb, t.children, ind+"      ")
		sb.WriteString(ind)
		sb.WriteString("    </div>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

// writeStaticAccordion emits disclosure sections. The first section is open in
// static HTML; the browser runtime toggles the open item.
func writeStaticAccordion(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-accordion")
	staticClassSuffix(sb, e)
	sb.WriteString("\"")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	open := true
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind != "element" || ch.elem == nil || ch.elem.tag != "accordionitem" {
			continue
		}
		item := ch.elem
		isOpen := open
		open = false
		sb.WriteString(ind)
		sb.WriteString("  <div class=\"sui-accordion-item")
		staticClassSuffix(sb, item)
		if isOpen {
			sb.WriteString(" open")
		}
		sb.WriteString("\">\n")
		sb.WriteString(ind)
		sb.WriteString("    <button type=\"button\" class=\"sui-accordion-head\" aria-expanded=\"")
		if isOpen {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		sb.WriteString("\">")
		staticText(sb, item.text)
		sb.WriteString("</button>\n")
		sb.WriteString(ind)
		sb.WriteString("    <div class=\"sui-accordion-panel\"")
		if !isOpen {
			sb.WriteString(" hidden")
		}
		sb.WriteString(">\n")
		writeStaticNodes(sb, item.children, ind+"      ")
		sb.WriteString(ind)
		sb.WriteString("    </div>\n")
		sb.WriteString(ind)
		sb.WriteString("  </div>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

// writeStaticToastHost emits an empty live-region host. toast.show() appends
// text-only notifications to it at runtime.
func toastAttrValue(e *Element, key string) string {
	a := findAttr(e, key)
	if a == nil {
		return ""
	}
	if a.hasExpr && a.expr != nil {
		return strings.Trim(strings.TrimSpace(exprSource(a.expr)), "\"'")
	}
	return strings.TrimSpace(partsLiteral(a.text))
}

// toastPlacementClasses turns the compact designer/source attributes into
// static CSS hooks. Position and inset stay in CSS so they never fight the
// card style that the inspector writes for .sui-toast.
func toastAxisValues(e *Element) (h, v, x, y string) {
	h = strings.ToLower(strings.TrimSpace(toastAttrValue(e, "hAlign")))
	v = strings.ToLower(strings.TrimSpace(toastAttrValue(e, "vAlign")))
	x = strings.TrimSpace(toastAttrValue(e, "hOffset"))
	y = strings.TrimSpace(toastAttrValue(e, "vOffset"))
	// Compatibility with the first compact syntax: position="top-left",
	// offset="24px". Opening the inspector and changing an axis rewrites it to
	// hAlign/vAlign/hOffset/vOffset.
	if h == "" && v == "" {
		pos := strings.ToLower(strings.TrimSpace(toastAttrValue(e, "position")))
		if parts := strings.SplitN(pos, "-", 2); len(parts) == 2 {
			v, h = parts[0], parts[1]
		}
	}
	if x == "" {
		x = strings.TrimSpace(toastAttrValue(e, "offset"))
	}
	if y == "" {
		y = x
	}
	return
}

func toastPlacementClasses(e *Element) string {
	h, v, _, _ := toastAxisValues(e)
	classes := ""
	switch h {
	case "left", "center", "right":
		classes = "sui-toast-h-" + h
	}
	switch v {
	case "top", "center", "bottom":
		if classes != "" {
			classes += " "
		}
		classes += "sui-toast-v-" + v
	}
	return classes
}

func toastOffsetValue(e *Element, axis string) string {
	_, _, x, y := toastAxisValues(e)
	if axis == "y" {
		return y
	}
	return x
}

func comboboxAttrValue(e *Element, key string) string {
	a := findAttr(e, key)
	if a == nil || a.hasExpr {
		return ""
	}
	return partsContent(a.text)
}

func comboboxAttrNumber(e *Element, key string) string {
	a := findAttr(e, key)
	if a == nil {
		return ""
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
		return strconv.Itoa(a.expr.i)
	}
	if !a.hasExpr {
		return strings.TrimSpace(partsContent(a.text))
	}
	return ""
}

func treeKey(e *Element, path string) string {
	if a := findAttr(e, "key"); a != nil && !a.hasExpr {
		if k := strings.TrimSpace(partsContent(a.text)); k != "" {
			return k
		}
	}
	return path
}

// eventIdentOf returns the function name named by a handler attribute, or "".
func eventIdentOf(e *Element, key string) string {
	a := findAttr(e, key)
	if a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		return a.expr.name
	}
	return ""
}

// treeVisualOpts carries the tree-wide default glyphs through the recursive
// static/dynamic tree renderers. Per-item icon/iconOpen wins, then the
// parent/leaf default from the tree element.
type treeVisualOpts struct {
	IconLeaf       string
	IconParent     string
	IconParentOpen string
	ToggleClosed   string
	ToggleOpen     string
}

func treeVisualOptions(e *Element) treeVisualOpts {
	tv := treeVisualOpts{
		IconLeaf:       treeTextAttr(e, "iconLeaf"),
		IconParent:     treeTextAttr(e, "iconParent"),
		IconParentOpen: treeTextAttr(e, "iconParentOpen"),
		ToggleClosed:   treeTextAttr(e, "toggleClosed"),
		ToggleOpen:     treeTextAttr(e, "toggleOpen"),
	}
	if tv.ToggleClosed == "" {
		tv.ToggleClosed = "\u25b8"
	}
	if tv.ToggleOpen == "" {
		tv.ToggleOpen = "\u25be"
	}
	return tv
}

func treeItemIcon(it *Element, isParent, open bool, tv treeVisualOpts) string {
	icon := treeTextAttr(it, "icon")
	if icon == "" {
		if isParent {
			icon = tv.IconParent
		} else {
			icon = tv.IconLeaf
		}
	}
	if open && isParent {
		if v := treeTextAttr(it, "iconOpen"); v != "" {
			icon = v
		} else if tv.IconParentOpen != "" {
			icon = tv.IconParentOpen
		}
	}
	return icon
}

func treeToggleText(open bool, tv treeVisualOpts) string {
	if open {
		return tv.ToggleOpen
	}
	return tv.ToggleClosed
}

func treeItemRename(e *Element) string {
	a := findAttr(e, "rename")
	if a == nil {
		return "inherit"
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		if a.expr.b {
			return "1"
		}
		return "0"
	}
	v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
	if v == "true" || v == "yes" || v == "1" {
		return "1"
	}
	if v == "false" || v == "no" || v == "0" {
		return "0"
	}
	return "inherit"
}

func treeItemShowRenameIcon(e *Element) string {
	a := findAttr(e, "showRenameIcon")
	if a == nil {
		return "inherit"
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		if a.expr.b {
			return "1"
		}
		return "0"
	}
	v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
	if v == "true" || v == "1" {
		return "1"
	}
	if v == "false" || v == "0" {
		return "0"
	}
	return "inherit"
}

func treeItemRenameIcon(e *Element) string {
	a := findAttr(e, "renameIcon")
	if a == nil || a.hasExpr {
		return ""
	}
	return partsContent(a.text)
}

func treeOpen(e *Element) bool {
	a := findAttr(e, "open")
	return a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b
}

// treeTextAttr returns a literal text attribute, or "" when absent/dynamic.
// Icons, badges and metadata are rendered as inert text, never as markup.
func treeTextAttr(e *Element, key string) string {
	a := findAttr(e, key)
	if a == nil || a.hasExpr {
		return ""
	}
	return partsContent(a.text)
}

func treeChecked(e *Element) string {
	a := findAttr(e, "checked")
	if a == nil {
		return ""
	}
	on := false
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		on = a.expr.b
	} else {
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		on = v == "true" || v == "yes" || v == "1" || v == "on"
	}
	if on {
		return "1"
	}
	return "0"
}

func treeBadgeType(e *Element) string {
	v := strings.ToLower(strings.TrimSpace(treeTextAttr(e, "badgeType")))
	switch v {
	case "info", "success", "warning", "danger", "neutral":
		return v
	}
	return "neutral"
}

// treeColumns returns the configured tree-grid columns in display order.
// Column keys are free-form identifiers; size, date and status are only
// convenience names with built-in item attributes.
func treeColumns(e *Element) []string {
	v := treeTextAttr(e, "columns")
	if strings.TrimSpace(v) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		var clean strings.Builder
		for _, r := range name {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
				clean.WriteRune(r)
			}
		}
		if clean.Len() > 0 {
			out = append(out, clean.String())
		}
	}
	return out
}

func treeColumnLabel(name string) string {
	switch name {
	case "size":
		return "Size"
	case "date":
		return "Date"
	case "status":
		return "Status"
	}
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '_' || r == '-' })
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	if len(parts) == 0 {
		return name
	}
	return strings.Join(parts, " ")
}

func treeCellValue(it *Element, col string) string {
	switch col {
	case "size", "date", "status":
		return treeTextAttr(it, col)
	}
	raw := strings.TrimSpace(treeTextAttr(it, "cells"))
	if raw == "" {
		return ""
	}
	cells := treeCellsMap(raw)
	return cells[col]
}

func treeCellsMap(raw string) map[string]string {
	out := map[string]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	if strings.HasPrefix(raw, "{") {
		var obj map[string]any
		if err := json.Unmarshal([]byte(raw), &obj); err == nil {
			for k, v := range obj {
				switch x := v.(type) {
				case string:
					out[k] = x
				case float64:
					out[k] = strconv.FormatFloat(x, 'g', -1, 64)
				case bool:
					if x {
						out[k] = "true"
					} else {
						out[k] = "false"
					}
				}
			}
		}
		return out
	}
	sep := ";"
	if strings.Contains(raw, "|") && !strings.Contains(raw, ";") {
		sep = "|"
	}
	for _, pair := range strings.Split(raw, sep) {
		i := strings.Index(pair, "=")
		if i <= 0 {
			continue
		}
		k := strings.TrimSpace(pair[:i])
		if k != "" {
			out[k] = strings.TrimSpace(pair[i+1:])
		}
	}
	return out
}

func treeFontStep(e *Element) float64 {
	a := findAttr(e, "fontStep")
	if a == nil {
		return 0
	}
	if a.hasExpr && a.expr != nil {
		if a.expr.kind == "int" {
			return float64(a.expr.i)
		}
		if a.expr.kind == "float" {
			return a.expr.f
		}
		return 0
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(partsContent(a.text)), 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func treeFontStyle(level int, step float64) string {
	if step <= 0 || level <= 1 {
		return ""
	}
	offset := float64(level-1) * step
	return " style=\"font-size:calc(var(--sui-tree-font, 13px) - " + strconv.FormatFloat(offset, 'g', -1, 64) + "px)\""
}

func splitTreeColumnText(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		out = append(out, strings.TrimSpace(part))
	}
	return out
}

func splitTreeColumnWidths(v string) []int {
	var out []int
	for _, part := range strings.Split(v, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && n >= 20 && n <= 600 {
			out = append(out, n)
		} else {
			out = append(out, 0)
		}
	}
	return out
}

func treeColumnWidth(widths []int, i int) int {
	if i >= 0 && i < len(widths) {
		return widths[i]
	}
	return 0
}

func boolFlag(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func sortTreeElements(items []*Element, dir string) []*Element {
	if dir != "asc" && dir != "desc" {
		return items
	}
	out := append([]*Element(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		a := strings.ToLower(partsContent(out[i].text))
		b := strings.ToLower(partsContent(out[j].text))
		if dir == "asc" {
			return a < b
		}
		return a > b
	})
	return out
}

func treeChildren(e *Element) []*Element {
	var out []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "treeitem" {
			out = append(out, ch.elem)
		}
	}
	return out
}

func writeStaticTree(sb *strings.Builder, e *Element, ind string) {
	value := comboboxAttrValue(e, "value")
	title := partsContent(e.text)
	treeKeyName := strings.TrimSpace(title)
	if treeKeyName == "" {
		treeKeyName = "tree"
	}
	treeName := comboboxAttrValue(e, "name")
	if strings.TrimSpace(treeName) == "" {
		treeName = treeKeyName
	}
	editMode := strings.ToLower(strings.TrimSpace(treeTextAttr(e, "editMode")))
	if editMode != "editable" && editMode != "persistent" {
		editMode = "none"
	}
	depth := comboboxAttrNumber(e, "editDepth")
	renameIcon := comboboxAttrValue(e, "renameIcon")
	showRenameIcon := true
	if a := findAttr(e, "showRenameIcon"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		showRenameIcon = a.expr.b
	}
	guides := true
	if a := findAttr(e, "guides"); a != nil {
		guides = datepickerBool(e, "guides")
	}
	truncate := true
	if a := findAttr(e, "truncate"); a != nil {
		truncate = datepickerBool(e, "truncate")
	}
	checkbox := datepickerBool(e, "checkbox")
	counts := datepickerBool(e, "counts")
	virtual := datepickerBool(e, "virtual")
	tooltips := datepickerBool(e, "tooltips")
	rowHeight := comboboxAttrNumber(e, "rowHeight")
	tv := treeVisualOptions(e)
	onCheck := ""
	if a := findAttr(e, "onCheck"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onCheck = a.expr.name
	}
	onLoad := ""
	if a := findAttr(e, "onLoad"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onLoad = a.expr.name
	}
	onMove := ""
	if a := findAttr(e, "onMove"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onMove = a.expr.name
	}
	onContextMenu := ""
	if a := findAttr(e, "onContextMenu"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onContextMenu = a.expr.name
	}
	onTreeChange := ""
	if a := findAttr(e, "onTreeChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onTreeChange = a.expr.name
	}
	multiSelect := datepickerBool(e, "multiSelect")
	search := datepickerBool(e, "search")
	drag := datepickerBool(e, "drag")
	contextMenu := datepickerBool(e, "contextMenu")
	columns := treeColumns(e)
	columnsEnabled := true
	if a := findAttr(e, "columnsEnabled"); a != nil {
		columnsEnabled = datepickerBool(e, "columnsEnabled")
	}
	columnLabels := splitTreeColumnText(treeTextAttr(e, "columnLabels"))
	columnWidths := splitTreeColumnWidths(treeTextAttr(e, "columnWidths"))
	sortDir := strings.ToLower(strings.TrimSpace(treeTextAttr(e, "sort")))
	persistSelection := datepickerBool(e, "persistSelection")
	fontStep := treeFontStep(e)
	if !columnsEnabled {
		columns = nil
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-tree")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-tree=\"1\" data-sui-tree-name=\"" + htmlEscape(treeName) + "\" data-sui-tree-key=\"" + htmlEscape(treeKeyName) + "\" role=\"tree\"")
	if editMode != "none" {
		sb.WriteString(" data-sui-tree-edit-mode=\"" + htmlEscape(editMode) + "\"")
	}
	if depth != "" {
		sb.WriteString(" data-sui-tree-edit-depth=\"" + htmlEscape(depth) + "\"")
	}
	if renameIcon != "" {
		sb.WriteString(" data-sui-tree-rename-icon=\"" + htmlEscape(renameIcon) + "\"")
	}
	if !showRenameIcon {
		sb.WriteString(" data-sui-tree-show-rename-icon=\"0\"")
	}
	if guides {
		sb.WriteString(" data-sui-tree-guides=\"1\"")
	}
	if !truncate {
		sb.WriteString(" data-sui-tree-truncate=\"0\"")
	}
	if checkbox {
		sb.WriteString(" data-sui-tree-checkbox=\"1\"")
	}
	if counts {
		sb.WriteString(" data-sui-tree-counts=\"1\"")
	}
	if virtual {
		sb.WriteString(" data-sui-tree-virtual=\"1\"")
	}
	if tooltips {
		sb.WriteString(" data-sui-tree-tooltips=\"1\"")
	}
	if rowHeight != "" {
		sb.WriteString(" data-sui-tree-row-height=\"" + htmlEscape(rowHeight) + "\"")
	}

	if tv.IconLeaf != "" {
		sb.WriteString(" data-sui-tree-icon-leaf=\"" + htmlEscape(tv.IconLeaf) + "\"")
	}
	if tv.IconParent != "" {
		sb.WriteString(" data-sui-tree-icon-parent=\"" + htmlEscape(tv.IconParent) + "\"")
	}
	if tv.IconParentOpen != "" {
		sb.WriteString(" data-sui-tree-icon-parent-open=\"" + htmlEscape(tv.IconParentOpen) + "\"")
	}
	sb.WriteString(" data-sui-tree-toggle-closed=\"" + htmlEscape(tv.ToggleClosed) + "\"")
	sb.WriteString(" data-sui-tree-toggle-open=\"" + htmlEscape(tv.ToggleOpen) + "\"")
	if onCheck != "" {
		sb.WriteString(" data-sui-tree-oncheck=\"" + htmlEscape(onCheck) + "\"")
	}
	if onLoad != "" {
		sb.WriteString(" data-sui-tree-onload=\"" + htmlEscape(onLoad) + "\"")
	}
	if onMove != "" {
		sb.WriteString(" data-sui-tree-onmove=\"" + htmlEscape(onMove) + "\"")
	}
	if onContextMenu != "" {
		sb.WriteString(" data-sui-tree-oncontextmenu=\"" + htmlEscape(onContextMenu) + "\"")
	}
	if onTreeChange != "" {
		sb.WriteString(" data-sui-tree-ontreechange=\"" + htmlEscape(onTreeChange) + "\"")
	}
	if multiSelect {
		sb.WriteString(" data-sui-tree-multi=\"1\"")
	}
	if search {
		sb.WriteString(" data-sui-tree-search=\"1\"")
	}
	if drag {
		sb.WriteString(" data-sui-tree-drag=\"1\"")
	}
	if contextMenu {
		sb.WriteString(" data-sui-tree-context-menu=\"1\"")
	}
	if len(columns) > 0 {
		sb.WriteString(" data-sui-tree-columns=\"" + htmlEscape(strings.Join(columns, ",")) + "\"")
	}
	if !columnsEnabled {
		sb.WriteString(" data-sui-tree-columns-enabled=\"0\"")
	}
	if strings.TrimSpace(strings.Join(columnLabels, "")) != "" {
		sb.WriteString(" data-sui-tree-column-labels=\"" + htmlEscape(strings.Join(columnLabels, ",")) + "\"")
	}
	if len(columnWidths) > 0 {
		var ws []string
		for _, w := range columnWidths {
			ws = append(ws, strconv.Itoa(w))
		}
		sb.WriteString(" data-sui-tree-column-widths=\"" + htmlEscape(strings.Join(ws, ",")) + "\"")
	}
	if sortDir == "asc" || sortDir == "desc" {
		sb.WriteString(" data-sui-tree-sort=\"" + sortDir + "\"")
	}
	if persistSelection {
		sb.WriteString(" data-sui-tree-persist-selection=\"1\"")
	}
	if fontStep > 0 {
		sb.WriteString(" data-sui-tree-font-step=\"" + strconv.FormatFloat(fontStep, 'g', -1, 64) + "\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	indent := ind + "  "
	if title != "" {
		sb.WriteString(indent + "<div class=\"sui-tree-title\">" + htmlEscape(title) + "</div>\n")
	}
	sb.WriteString(indent + "<input type=\"hidden\" class=\"sui-tree-value\" value=\"" + htmlEscape(value) + "\">\n")
	if search {
		sb.WriteString(indent + "<div class=\"sui-tree-search\"><input type=\"search\" class=\"sui-tree-search-input\" placeholder=\"Search tree\" aria-label=\"Search tree\"><span class=\"sui-tree-search-count\" aria-live=\"polite\"></span></div>\n")
	}
	if len(columns) > 0 {
		sb.WriteString(indent + "<div class=\"sui-tree-grid-head\" role=\"presentation\"><span class=\"sui-tree-grid-name\">Name</span>")
		for i, col := range columns {
			label := treeColumnLabel(col)
			if i < len(columnLabels) && columnLabels[i] != "" {
				label = columnLabels[i]
			}
			sb.WriteString("<span class=\"sui-tree-cell sui-tree-cell-" + col + "\"")
			if w := treeColumnWidth(columnWidths, i); w > 0 {
				sb.WriteString(" style=\"width:" + strconv.Itoa(w) + "px\"")
			}
			sb.WriteString(">" + htmlEscape(label) + "</span>")
		}
		sb.WriteString("</div>\n")
	}
	sb.WriteString(indent + "<ul class=\"sui-tree-group\" role=\"group\">\n")
	writeStaticTreeItems(sb, treeChildren(e), indent+"  ", 1, "0", guides, checkbox, counts, tv, columns, columnWidths, sortDir, fontStep)
	sb.WriteString(indent + "</ul>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticTreeItems(sb *strings.Builder, items []*Element, ind string, level int, pathPrefix string, guides, checkbox, counts bool, tv treeVisualOpts, columns []string, columnWidths []int, sortDir string, fontStep float64) {
	items = sortTreeElements(items, sortDir)
	for i, it := range items {
		path := pathPrefix + "." + strconv.Itoa(i)
		key := treeKey(it, path)
		kids := treeChildren(it)
		open := treeOpen(it)
		lazy := datepickerBool(it, "lazy")
		parent := len(kids) > 0 || lazy
		label := partsContent(it.text)
		explicitIcon := treeTextAttr(it, "icon")
		explicitIconOpen := treeTextAttr(it, "iconOpen")
		icon := treeItemIcon(it, parent, open, tv)
		badge := treeTextAttr(it, "badge")
		meta := treeTextAttr(it, "meta")
		checked := treeChecked(it)
		size := treeTextAttr(it, "size")
		date := treeTextAttr(it, "date")
		status := treeTextAttr(it, "status")
		cells := treeTextAttr(it, "cells")
		sb.WriteString(ind)
		sb.WriteString("<li class=\"sui-tree-item")
		staticClassSuffix(sb, it)
		sb.WriteString("\" data-sui-tree-item=\"1\" data-sui-key=\"" + htmlEscape(key) +
			"\" data-sui-path=\"" + htmlEscape(path) + "\" data-sui-level=\"" + strconv.Itoa(level) +
			"\" data-sui-rename=\"" + htmlEscape(treeItemRename(it)) +
			"\" data-sui-show-rename-icon=\"" + htmlEscape(treeItemShowRenameIcon(it)) +
			"\" data-sui-rename-icon=\"" + htmlEscape(treeItemRenameIcon(it)) +
			"\" data-sui-icon-base=\"" + htmlEscape(explicitIcon) +
			"\" data-sui-icon-open-base=\"" + htmlEscape(explicitIconOpen) +
			"\" data-sui-icon=\"" + htmlEscape(icon) +
			"\" data-sui-badge=\"" + htmlEscape(badge) +
			"\" data-sui-badge-type=\"" + htmlEscape(treeBadgeType(it)) +
			"\" data-sui-meta=\"" + htmlEscape(meta) +
			"\" data-sui-size=\"" + htmlEscape(size) +
			"\" data-sui-date=\"" + htmlEscape(date) +
			"\" data-sui-status=\"" + htmlEscape(status) +
			"\" data-sui-cells=\"" + htmlEscape(cells) +
			"\" data-sui-lazy=\"" + boolFlag(lazy) +
			"\" data-sui-checked=\"" + htmlEscape(checked) + "\" data-sui-open=\"")
		if open {
			sb.WriteString("1")
		} else {
			sb.WriteString("0")
		}
		sb.WriteString("\" role=\"treeitem\" aria-level=\"" + strconv.Itoa(level) + "\" aria-selected=\"false\" tabindex=\"-1\"")
		if checkbox {
			st := "false"
			if checked == "1" {
				st = "true"
			}
			sb.WriteString(" aria-checked=\"" + st + "\"")
		}
		if len(kids) > 0 || lazy {
			sb.WriteString(" aria-expanded=\"")
			if open {
				sb.WriteString("true")
			} else {
				sb.WriteString("false")
			}
			sb.WriteString("\"")
		}
		staticUserAttrsNoClass(sb, it, "")
		sb.WriteString(">\n")
		sb.WriteString(ind + "  <div class=\"sui-tree-row\">")
		if guides {
			for g := 1; g < level; g++ {
				sb.WriteString("<span class=\"sui-tree-guide\" aria-hidden=\"true\"></span>")
			}
		}
		if len(kids) > 0 || lazy {
			sb.WriteString("<button type=\"button\" class=\"sui-tree-toggle\" tabindex=\"-1\" aria-label=\"Toggle " + htmlEscape(label) + "\">")
			sb.WriteString(htmlEscape(treeToggleText(open, tv)))
			sb.WriteString("</button>")
		} else {
			sb.WriteString("<span class=\"sui-tree-spacer\"></span>")
		}
		if checkbox {
			st := "false"
			if checked == "1" {
				st = "true"
			}
			sb.WriteString("<span class=\"sui-tree-check\" role=\"checkbox\" aria-checked=\"" + st + "\" tabindex=\"-1\" aria-label=\"Toggle " + htmlEscape(label) + "\"></span>")
		}
		if icon != "" {
			sb.WriteString("<span class=\"sui-tree-icon\" aria-hidden=\"true\">" + htmlEscape(icon) + "</span>")
		}
		sb.WriteString("<span class=\"sui-tree-label\"" + treeFontStyle(level, fontStep) + " title=\"" + htmlEscape(label) + "\">" + htmlEscape(label) + "</span>")
		if counts && len(kids) > 0 {
			sb.WriteString("<span class=\"sui-tree-badge sui-tree-badge-count\">" + strconv.Itoa(len(kids)) + "</span>")
		}
		if badge != "" {
			sb.WriteString("<span class=\"sui-tree-badge sui-tree-badge-" + treeBadgeType(it) + "\">" + htmlEscape(badge) + "</span>")
		}
		if meta != "" {
			sb.WriteString("<span class=\"sui-tree-meta\">" + htmlEscape(meta) + "</span>")
		}
		if len(columns) > 0 {
			for ci, col := range columns {
				val := treeCellValue(it, col)
				sb.WriteString("<span class=\"sui-tree-cell sui-tree-cell-" + col + "\"")
				if w := treeColumnWidth(columnWidths, ci); w > 0 {
					sb.WriteString(" style=\"width:" + strconv.Itoa(w) + "px\"")
				}
				sb.WriteString(">" + htmlEscape(val) + "</span>")
			}
		}
		sb.WriteString("</div>\n")
		if len(kids) > 0 {
			sb.WriteString(ind + "  <ul class=\"sui-tree-group\" role=\"group\"")
			if !open {
				sb.WriteString(" hidden")
			}
			sb.WriteString(">\n")
			writeStaticTreeItems(sb, kids, ind+"    ", level+1, path, guides, checkbox, counts, tv, columns, columnWidths, sortDir, fontStep)
			sb.WriteString(ind + "  </ul>\n")
		}
		sb.WriteString(ind + "</li>\n")
	}
}

func chartStaticValue(e *Element, key, def string) string {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	if a.hasExpr {
		if a.expr != nil && a.expr.kind == "int" {
			return strconv.Itoa(a.expr.i)
		}
		if a.expr != nil && a.expr.kind == "float" {
			return strconv.FormatFloat(a.expr.f, 'g', -1, 64)
		}
		return def
	}
	if v := partsContent(a.text); strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func chartStaticFlag(e *Element, key string, def bool) bool {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		return a.expr.b
	}
	v := strings.ToLower(strings.Trim(strings.TrimSpace(partsContent(a.text)), "\"'"))
	if v == "true" || v == "1" || v == "yes" || v == "on" {
		return true
	}
	if v == "false" || v == "0" || v == "no" || v == "off" {
		return false
	}
	return def
}

func writeStaticChart(sb *strings.Builder, e *Element, ind string) {
	name := chartNameValue(e)
	title := partsContent(e.text)
	ariaLabel := strings.TrimSpace(title)
	if ariaLabel == "" {
		if e.tag == "graph" {
			ariaLabel = "Graph"
		} else {
			ariaLabel = "Chart"
		}
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-chart")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-chart=\"1\" role=\"group\" aria-label=\"" + htmlEscape(ariaLabel) + "\" data-sui-chart-name=\"" + htmlEscape(name) +
		"\" data-sui-chart-title=\"" + htmlEscape(title) + "\"")
	if e.tag == "graph" {
		sb.WriteString(" data-sui-chart-type=\"state\"")
	} else {
		sb.WriteString(" data-sui-chart-type=\"" + htmlEscape(chartStaticValue(e, "chartType", "bar")) + "\"")
	}
	sb.WriteString(" data-sui-chart-theme=\"" + htmlEscape(chartStaticValue(e, "theme", "dark")) +
		"\" data-sui-chart-edit-mode=\"" + htmlEscape(chartStaticValue(e, "editMode", "")) +
		"\" data-sui-chart-data=\"" + htmlEscape(chartStaticValue(e, "data", "")) +
		"\" data-sui-chart-format=\"" + htmlEscape(chartStaticValue(e, "format", "auto")) +
		"\" data-sui-chart-x=\"" + htmlEscape(chartStaticValue(e, "x", "")) +
		"\" data-sui-chart-y=\"" + htmlEscape(chartStaticValue(e, "y", "")) + "\"")
	if findAttr(e, "width") != nil {
		sb.WriteString(" data-sui-chart-width=\"" + htmlEscape(chartStaticValue(e, "width", "640px")) + "\"")
	}
	if findAttr(e, "height") != nil {
		sb.WriteString(" data-sui-chart-height=\"" + htmlEscape(chartStaticValue(e, "height", "320px")) + "\"")
	}
	sb.WriteString(" data-sui-chart-overflow=\"" + htmlEscape(chartStaticValue(e, "overflow", "scroll")) +
		"\" data-sui-chart-zoom=\"" + htmlEscape(chartStaticValue(e, "zoom", "1")) +
		"\" data-sui-chart-point-width=\"" + htmlEscape(chartStaticValue(e, "pointWidth", "48")) +
		"\" data-sui-chart-point-size=\"" + htmlEscape(chartStaticValue(e, "pointSize", "3")) +
		"\" data-sui-chart-table-view=\"" + htmlEscape(chartStaticValue(e, "tableView", "chart")) +
		"\" data-sui-chart-table-filters=\"" + boolAttr01(chartStaticFlag(e, "tableFilters", true)) +
		"\" data-sui-chart-table-add-row=\"" + boolAttr01(chartStaticFlag(e, "tableAddRow", false)) +
		"\" data-sui-chart-table-delete-row=\"" + boolAttr01(chartStaticFlag(e, "tableDeleteRow", false)) +
		"\" data-sui-chart-table-steppers=\"" + boolAttr01(chartStaticFlag(e, "tableSteppers", false)) +
		"\" data-sui-chart-table-context-menu=\"" + boolAttr01(chartStaticFlag(e, "tableContextMenu", false)) +
		"\" data-sui-chart-table-selection=\"" + htmlEscape(chartStaticValue(e, "tableSelection", "row")) +
		"\" data-sui-chart-critical-path=\"" + boolAttr01(chartStaticFlag(e, "criticalPath", false)) +
		"\" data-sui-chart-holidays=\"" + htmlEscape(chartStaticValue(e, "holidays", "")) +
		"\" data-sui-chart-weekend=\"" + htmlEscape(chartStaticValue(e, "weekend", "")) +
		"\" data-sui-chart-timezone=\"" + htmlEscape(chartStaticValue(e, "timezone", "")) +
		"\" data-sui-chart-show-non-working-days=\"" + boolAttr01(chartStaticFlag(e, "showNonWorkingDays", true)) +
		"\" data-sui-chart-progress-column=\"" + boolAttr01(chartStaticFlag(e, "progressColumn", false)) +
		"\" data-sui-chart-bins=\"" + htmlEscape(chartStaticValue(e, "bins", "10")) +
		"\" data-sui-chart-min=\"" + htmlEscape(chartStaticValue(e, "min", "")) +
		"\" data-sui-chart-max=\"" + htmlEscape(chartStaticValue(e, "max", "")) +
		"\" data-sui-chart-color-picker=\"" + boolAttr01(chartStaticFlag(e, "colorPicker", true)) +
		"\" data-sui-chart-settings-button=\"" + boolAttr01(chartStaticFlag(e, "settingsButton", true)) +
		"\" data-sui-chart-show-zoom=\"" + boolAttr01(chartStaticFlag(e, "showZoom", true)) +
		"\" data-sui-chart-show-add=\"" + boolAttr01(chartStaticFlag(e, "showAdd", true)) +
		"\" data-sui-chart-show-history=\"" + boolAttr01(chartStaticFlag(e, "showHistory", true)) +
		"\" data-sui-chart-show-resources=\"" + boolAttr01(chartStaticFlag(e, "showResources", false)) +
		"\" data-sui-chart-show-progress=\"" + boolAttr01(chartStaticFlag(e, "showProgress", false)) +
		"\" data-sui-chart-show-slack=\"" + boolAttr01(chartStaticFlag(e, "showSlack", false)) +
		"\" data-sui-chart-show-costs=\"" + boolAttr01(chartStaticFlag(e, "showCosts", false)) +
		"\" data-sui-chart-status-colors=\"" + boolAttr01(chartStaticFlag(e, "statusColors", false)) +
		"\" data-sui-chart-auto-schedule=\"" + boolAttr01(chartStaticFlag(e, "autoSchedule", false)) +
		"\" data-sui-chart-gantt-scale=\"" + htmlEscape(chartStaticValue(e, "ganttScale", "auto")) +
		"\" data-sui-chart-link-color=\"" + htmlEscape(chartStaticValue(e, "linkColor", "#94a3b8")) +
		"\" data-sui-chart-state-shape-border-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeBorderColor", "")) +
		"\" data-sui-chart-state-selection-color=\"" + htmlEscape(chartStaticValue(e, "stateSelectionColor", "")) +
		"\" data-sui-chart-link-width=\"" + htmlEscape(chartStaticValue(e, "linkWidth", "1")) +
		"\" data-sui-chart-link-gap=\"" + htmlEscape(chartStaticValue(e, "linkGap", "4")) +
		"\" data-sui-chart-bar-border-color=\"" + htmlEscape(chartStaticValue(e, "barBorderColor", "")) +
		"\" data-sui-chart-link-style=\"" + htmlEscape(chartStaticValue(e, "linkStyle", "auto")) +
		"\" data-sui-chart-arrow-type=\"" + htmlEscape(chartStaticValue(e, "arrowType", "triangle")) +
		"\" data-sui-chart-arrow-size=\"" + htmlEscape(chartStaticValue(e, "arrowSize", "")) +
		"\" data-sui-chart-state-start-arrow=\"" + boolAttr01(chartStaticFlag(e, "stateStartArrow", false)) +
		"\" data-sui-chart-state-routing=\"" + htmlEscape(chartStaticValue(e, "stateRouting", "orthogonal")) +
		"\" data-sui-chart-state-tag-style=\"" + htmlEscape(chartStaticValue(e, "stateTagStyle", "pill")) +
		"\" data-sui-chart-state-tag-link=\"" + htmlEscape(chartStaticValue(e, "stateTagLink", "line")) +
		"\" data-sui-chart-state-link-tag-style=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagStyle", "")) +
		"\" data-sui-chart-state-link-tag-link=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLink", "")) +
		"\" data-sui-chart-state-shape-tag-style=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagStyle", "box")) +
		"\" data-sui-chart-state-shape-tag-link=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLink", "line")) +
		"\" data-sui-chart-state-shape-tag-kind=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagKind", "external")) +
		"\" data-sui-chart-state-shape-tag-align=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagAlign", "line")) +
		"\" data-sui-chart-state-shape-tag-rotation=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagRotation", "")) +
		"\" data-sui-chart-state-shape-tag-arrow-size=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagArrowSize", "")) +
		"\" data-sui-chart-state-link-tag-kind=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagKind", "external")) +
		"\" data-sui-chart-state-link-tag-align=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagAlign", "line")) +
		"\" data-sui-chart-state-link-tag-rotation=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagRotation", "")) +
		"\" data-sui-chart-state-link-tag-arrow-size=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagArrowSize", "")) +
		"\" data-sui-chart-state-tag-font=\"" + htmlEscape(chartStaticValue(e, "stateTagFont", "")) +
		"\" data-sui-chart-state-tag-font-size=\"" + htmlEscape(chartStaticValue(e, "stateTagFontSize", "11")) +
		"\" data-sui-chart-state-shape-tag-font=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFont", "")) +
		"\" data-sui-chart-state-shape-tag-font-size=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFontSize", "")) +
		"\" data-sui-chart-state-link-tag-font=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFont", "")) +
		"\" data-sui-chart-state-link-tag-font-size=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFontSize", "")) +
		"\" data-sui-chart-state-tag-font-weight=\"" + htmlEscape(chartStaticValue(e, "stateTagFontWeight", "400")) +
		"\" data-sui-chart-state-tag-font-style=\"" + htmlEscape(chartStaticValue(e, "stateTagFontStyle", "normal")) +
		"\" data-sui-chart-state-tag-text-decoration=\"" + htmlEscape(chartStaticValue(e, "stateTagTextDecoration", "none")) +
		"\" data-sui-chart-state-tag-text-align=\"" + htmlEscape(chartStaticValue(e, "stateTagTextAlign", "center")) +
		"\" data-sui-chart-state-tag-letter-spacing=\"" + htmlEscape(chartStaticValue(e, "stateTagLetterSpacing", "")) +
		"\" data-sui-chart-state-tag-line-spacing=\"" + htmlEscape(chartStaticValue(e, "stateTagLineSpacing", "")) +
		"\" data-sui-chart-state-tag-text-transform=\"" + htmlEscape(chartStaticValue(e, "stateTagTextTransform", "none")) +
		"\" data-sui-chart-state-tag-text-shadow=\"" + htmlEscape(chartStaticValue(e, "stateTagTextShadow", "none")) +
		"\" data-sui-chart-state-tag-text-shadow-color=\"" + htmlEscape(chartStaticValue(e, "stateTagTextShadowColor", "#000000")) +
		"\" data-sui-chart-state-shape-tag-font-weight=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFontWeight", "")) +
		"\" data-sui-chart-state-shape-tag-font-style=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFontStyle", "")) +
		"\" data-sui-chart-state-shape-tag-text-decoration=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextDecoration", "")) +
		"\" data-sui-chart-state-shape-tag-text-align=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextAlign", "")) +
		"\" data-sui-chart-state-shape-tag-letter-spacing=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLetterSpacing", "")) +
		"\" data-sui-chart-state-shape-tag-line-spacing=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLineSpacing", "")) +
		"\" data-sui-chart-state-shape-tag-text-transform=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextTransform", "")) +
		"\" data-sui-chart-state-shape-tag-text-shadow=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextShadow", "")) +
		"\" data-sui-chart-state-shape-tag-text-shadow-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextShadowColor", "")) +
		"\" data-sui-chart-state-link-tag-font-weight=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFontWeight", "")) +
		"\" data-sui-chart-state-link-tag-font-style=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFontStyle", "")) +
		"\" data-sui-chart-state-link-tag-text-decoration=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextDecoration", "")) +
		"\" data-sui-chart-state-link-tag-text-align=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextAlign", "")) +
		"\" data-sui-chart-state-link-tag-letter-spacing=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLetterSpacing", "")) +
		"\" data-sui-chart-state-link-tag-line-spacing=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLineSpacing", "")) +
		"\" data-sui-chart-state-link-tag-text-transform=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextTransform", "")) +
		"\" data-sui-chart-state-link-tag-text-shadow=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextShadow", "")) +
		"\" data-sui-chart-state-link-tag-text-shadow-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextShadowColor", "")) +
		"\" data-sui-chart-state-tag-text-color=\"" + htmlEscape(chartStaticValue(e, "stateTagTextColor", "")) +
		"\" data-sui-chart-state-tag-box-color=\"" + htmlEscape(chartStaticValue(e, "stateTagBoxColor", "")) +
		"\" data-sui-chart-state-tag-border-color=\"" + htmlEscape(chartStaticValue(e, "stateTagBorderColor", "")) +
		"\" data-sui-chart-state-tag-border-width=\"" + htmlEscape(chartStaticValue(e, "stateTagBorderWidth", "")) +
		"\" data-sui-chart-state-tag-border-style=\"" + htmlEscape(chartStaticValue(e, "stateTagBorderStyle", "solid")) +
		"\" data-sui-chart-state-tag-padding=\"" + htmlEscape(chartStaticValue(e, "stateTagPadding", "")) +
		"\" data-sui-chart-state-tag-leader-color=\"" + htmlEscape(chartStaticValue(e, "stateTagLeaderColor", "")) +
		"\" data-sui-chart-state-tag-arrow-shape=\"" + htmlEscape(chartStaticValue(e, "stateTagArrowShape", "triangle")) +
		"\" data-sui-chart-state-shape-tag-text-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextColor", "")) +
		"\" data-sui-chart-state-shape-tag-box-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBoxColor", "")) +
		"\" data-sui-chart-state-shape-tag-border-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBorderColor", "")) +
		"\" data-sui-chart-state-shape-tag-border-width=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBorderWidth", "")) +
		"\" data-sui-chart-state-shape-tag-border-style=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBorderStyle", "")) +
		"\" data-sui-chart-state-shape-tag-padding=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagPadding", "")) +
		"\" data-sui-chart-state-shape-tag-leader-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLeaderColor", "")) +
		"\" data-sui-chart-state-shape-tag-arrow-shape=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagArrowShape", "")) +
		"\" data-sui-chart-state-link-tag-text-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextColor", "")) +
		"\" data-sui-chart-state-link-tag-box-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBoxColor", "")) +
		"\" data-sui-chart-state-link-tag-border-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBorderColor", "")) +
		"\" data-sui-chart-state-link-tag-border-width=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBorderWidth", "")) +
		"\" data-sui-chart-state-link-tag-border-style=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBorderStyle", "")) +
		"\" data-sui-chart-state-link-tag-padding=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagPadding", "")) +
		"\" data-sui-chart-state-link-tag-leader-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLeaderColor", "")) +
		"\" data-sui-chart-state-link-tag-arrow-shape=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagArrowShape", "")) +
		"\" data-sui-chart-state-link-arc-radius=\"" + htmlEscape(chartStaticValue(e, "stateLinkArcRadius", "0")) +
		"\" data-sui-chart-state-grid=\"" + boolAttr01(chartStaticFlag(e, "stateGrid", true)) +
		"\" data-sui-chart-state-grid-gap=\"" + htmlEscape(chartStaticValue(e, "stateGridGap", "24")) +
		"\" data-sui-chart-state-grid-color=\"" + htmlEscape(chartStaticValue(e, "stateGridColor", "")) +
		"\" data-sui-chart-state-hints=\"" + boolAttr01(chartStaticFlag(e, "stateHints", true)) +
		"\" data-sui-chart-state-handles=\"" + boolAttr01(chartStaticFlag(e, "stateHandles", true)) +
		"\" data-sui-chart-state-tags-movable=\"" + boolAttr01(chartStaticFlag(e, "stateTagsMovable", false)) +
		"\" data-sui-chart-state-routes-editable=\"" + boolAttr01(chartStaticFlag(e, "stateRoutesEditable", false)) +
		"\" data-sui-chart-grid=\"" + boolAttr01(chartStaticFlag(e, "showGrid", true)) +
		"\" data-sui-chart-legend=\"" + boolAttr01(chartStaticFlag(e, "showLegend", true)) +
		"\" data-sui-chart-values=\"" + boolAttr01(chartStaticFlag(e, "showValues", false)) + "\"")
	if a := findAttr(e, "onTaskChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-ontaskchange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	if a := findAttr(e, "onCalendarChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-oncalendarchange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	if a := findAttr(e, "onDataChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-ondatachange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	if a := findAttr(e, "onCommit"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-oncommit=\"" + htmlEscape(a.expr.name) + "\"")
	}
	if a := findAttr(e, "onResourceChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-onresourcechange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	inner := ind + "  "
	sb.WriteString(inner + "<div class=\"sui-chart-head\"><div class=\"sui-chart-title\">" + htmlEscape(title) + "</div>")
	if chartStaticFlag(e, "settingsButton", true) {
		sb.WriteString("<button type=\"button\" class=\"sui-chart-menu-btn\" aria-label=\"Chart settings\" title=\"Chart settings\">&#9881;</button>")
	}
	sb.WriteString("<div class=\"sui-chart-zoom\" hidden><button type=\"button\" class=\"sui-chart-zoom-btn\" data-zoom=\"out\" aria-label=\"Zoom out\" title=\"Zoom out\">\u2212</button>")
	sb.WriteString("<span class=\"sui-chart-zoom-label\">100%</span><button type=\"button\" class=\"sui-chart-zoom-btn\" data-zoom=\"in\" aria-label=\"Zoom in\" title=\"Zoom in\">+</button>")
	sb.WriteString("<button type=\"button\" class=\"sui-chart-zoom-reset\" data-zoom=\"reset\" title=\"Reset zoom\">Reset</button>")
	sb.WriteString("<button type=\"button\" class=\"sui-chart-zoom-fit\" data-zoom=\"fit\" aria-label=\"Fit to width\" title=\"Fit to width\">Fit width</button>")
	sb.WriteString("<button type=\"button\" class=\"sui-chart-zoom-fit sui-chart-zoom-fit-all\" data-zoom=\"fitall\" aria-label=\"Fit all graph elements\" title=\"Fit all graph elements\">Fit all</button></div></div>\n")
	sb.WriteString(inner + "<div class=\"sui-chart-viewport\"><div class=\"sui-chart-stage\"><svg class=\"sui-chart-svg\" role=\"img\"></svg></div></div>\n")
	sb.WriteString(inner + "<div class=\"sui-chart-legend\"></div><div class=\"sui-chart-splitter\" role=\"separator\" tabindex=\"0\" aria-orientation=\"horizontal\" aria-valuemin=\"80\" aria-valuemax=\"320\" aria-valuenow=\"180\" aria-valuetext=\"180 pixels\" aria-label=\"Resize chart and table\" title=\"Drag, or use the arrow keys, to resize chart and table\"></div><div class=\"sui-chart-table\" role=\"region\" aria-label=\"Task table\" hidden></div><div class=\"sui-chart-error\" role=\"alert\" hidden></div><div class=\"sui-chart-tooltip\" role=\"tooltip\" aria-live=\"polite\" hidden></div>\n")
	sb.WriteString(ind + "</div>\n")
}

func boolAttr01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func datepickerBool(e *Element, key string) bool {
	a := findAttr(e, key)
	return a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b
}

func writeStaticDatePicker(sb *strings.Builder, e *Element, ind string) {
	mode := strings.ToLower(strings.TrimSpace(comboboxAttrValue(e, "mode")))
	if mode != "time" && mode != "datetime" {
		mode = "date"
	}
	inputType := mode
	if inputType == "datetime" {
		inputType = "datetime-local"
	}
	value := comboboxAttrValue(e, "value")
	min := comboboxAttrValue(e, "min")
	max := comboboxAttrValue(e, "max")
	step := comboboxAttrNumber(e, "step")
	disabled := datepickerBool(e, "disabled")
	required := datepickerBool(e, "required")
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-datepicker")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-datepicker=\"1\" data-sui-mode=\"" + htmlEscape(mode) + "\"")
	if disabled {
		sb.WriteString(" data-sui-disabled=\"1\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <label class=\"sui-datepicker-label\">")
	staticText(sb, e.text)
	sb.WriteString("</label>\n")
	sb.WriteString(ind + "  <div class=\"sui-datepicker-control\">\n")
	sb.WriteString(ind + "    <input type=\"" + htmlEscape(inputType) + "\" class=\"sui-datepicker-input\" autocomplete=\"off\"")
	if min != "" {
		sb.WriteString(" min=\"" + htmlEscape(min) + "\"")
	}
	if max != "" {
		sb.WriteString(" max=\"" + htmlEscape(max) + "\"")
	}
	if step != "" {
		sb.WriteString(" step=\"" + htmlEscape(step) + "\"")
	}
	if required {
		sb.WriteString(" required")
	}
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "    <button type=\"button\" class=\"sui-datepicker-clear\" aria-label=\"Clear value\" hidden>\u00d7</button>\n")
	sb.WriteString(ind + "  </div>\n")
	sb.WriteString(ind + "  <input type=\"hidden\" class=\"sui-datepicker-value\" value=\"" + htmlEscape(value) + "\"")
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <div class=\"sui-datepicker-error\" role=\"alert\" hidden></div>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticCombobox(sb *strings.Builder, e *Element, ind string) {
	options := comboboxAttrValue(e, "options")
	placeholder := comboboxAttrValue(e, "placeholder")
	if placeholder == "" {
		placeholder = "Search..."
	}
	empty := comboboxAttrValue(e, "empty")
	if empty == "" {
		empty = "No results"
	}
	value := comboboxAttrValue(e, "value")
	maxOptionLength := comboboxAttrNumber(e, "maxOptionLength")
	maxOptions := comboboxAttrNumber(e, "maxOptions")
	disabled := false
	if a := findAttr(e, "disabled"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		disabled = a.expr.b
	}
	required := false
	if a := findAttr(e, "required"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		required = a.expr.b
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-combobox")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-combobox=\"1\" data-sui-options=\"" + htmlEscape(options) +
		"\" data-sui-placeholder=\"" + htmlEscape(placeholder) +
		"\" data-sui-empty=\"" + htmlEscape(empty) + "\"")
	if maxOptionLength != "" {
		sb.WriteString(" data-sui-max-option-length=\"" + htmlEscape(maxOptionLength) + "\"")
	}
	if maxOptions != "" {
		sb.WriteString(" data-sui-max-options=\"" + htmlEscape(maxOptions) + "\"")
	}
	if required {
		sb.WriteString(" data-sui-required=\"1\"")
	}
	if disabled {
		sb.WriteString(" data-sui-disabled=\"1\"")
	}
	sb.WriteString(" data-sui-combobox-placement=\"" + attrChoice(e, "placement", []string{"auto", "bottom", "top", "left", "right"}, "auto") + "\"" + dataAnimAttrs(e, "fade"))
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <label class=\"sui-combobox-label\">")
	staticText(sb, e.text)
	sb.WriteString("</label>\n")
	sb.WriteString(ind + "  <div class=\"sui-combobox-control\">\n")
	sb.WriteString(ind + "    <input type=\"text\" class=\"sui-combobox-input\" role=\"combobox\" autocomplete=\"off\" spellcheck=\"false\" aria-expanded=\"false\" placeholder=\"" + htmlEscape(placeholder) + "\"")
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "    <button type=\"button\" class=\"sui-combobox-clear\" aria-label=\"Clear selection\" hidden>\u00d7</button>\n")
	sb.WriteString(ind + "    <button type=\"button\" class=\"sui-combobox-toggle\" aria-label=\"Open options\" aria-expanded=\"false\">\u25be</button>\n")
	sb.WriteString(ind + "  </div>\n")
	sb.WriteString(ind + "  <input type=\"hidden\" class=\"sui-combobox-value\" value=\"" + htmlEscape(value) + "\"")
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <div class=\"sui-combobox-popup\" role=\"listbox\" hidden></div>\n")
	sb.WriteString(ind + "  <div class=\"sui-combobox-error\" role=\"alert\" hidden></div>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticToastHost(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-toast-host")
	if pc := toastPlacementClasses(e); pc != "" {
		sb.WriteString(" " + pc)
	}
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-toast-host=\"1\" aria-live=\"polite\" aria-atomic=\"false\"")
	if v := toastAttrValue(e, "duration"); v != "" {
		sb.WriteString(" data-sui-toast-duration=\"" + htmlEscape(v) + "\"")
	}
	if v := toastAttrValue(e, "max"); v != "" {
		sb.WriteString(" data-sui-toast-max=\"" + htmlEscape(v) + "\"")
	}
	if v := toastAttrValue(e, "icons"); v != "" {
		sb.WriteString(" data-sui-toast-icons=\"" + htmlEscape(v) + "\"")
	}
	if v := toastOffsetValue(e, "x"); v != "" {
		sb.WriteString(" data-sui-toast-x=\"" + htmlEscape(v) + "\"")
	}
	if v := toastOffsetValue(e, "y"); v != "" {
		sb.WriteString(" data-sui-toast-y=\"" + htmlEscape(v) + "\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString("></div>\n")
}

// writeStaticMenuBar emits the semantic nav container used by the menu-bar
// control. The visible labels live on its `menu` children; the menubar text is
// an accessible name.
func writeStaticMenuBar(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<nav class=\"sui-menubar")
	staticClassSuffix(sb, e)
	sb.WriteString("\" role=\"menubar\"")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	writeStaticNodes(sb, e.children, ind+"  ")
	sb.WriteString(ind)
	sb.WriteString("</nav>\n")
}

// writeStaticMenu emits the button/popup structure for one top-level menu.
// The runtime adds the `open` class on click; without JavaScript the popup is
// hidden and the label still documents the structure.
func bool01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func writeStaticTimeline(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind + "<div class=\"sui-timeline")
	staticClassSuffix(sb, e)
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString("\">\n")
	for _, it := range childElemsByTag(e, "timelineitem") {
		status := strings.ToLower(attrLiteral(it, "status"))
		if status != "done" && status != "active" && status != "error" {
			status = "pending"
		}
		sb.WriteString(ind + "  <div class=\"sui-timeline-item sui-tl-" + status)
		staticClassSuffix(sb, it)
		sb.WriteString("\">\n")
		sb.WriteString(ind + "    <span class=\"sui-timeline-dot\"></span>\n")
		sb.WriteString(ind + "    <div class=\"sui-timeline-body\">\n")
		sb.WriteString(ind + "      <span class=\"sui-timeline-title\">" + htmlEscape(partsContent(it.text)) + "</span>\n")
		if date := attrLiteral(it, "date"); date != "" {
			sb.WriteString(ind + "      <span class=\"sui-timeline-date\">" + htmlEscape(date) + "</span>\n")
		}
		if len(it.children) > 0 {
			sb.WriteString(ind + "      <div class=\"sui-timeline-content\">\n")
			writeStaticNodes(sb, it.children, ind+"        ")
			sb.WriteString(ind + "      </div>\n")
		}
		sb.WriteString(ind + "    </div>\n")
		sb.WriteString(ind + "  </div>\n")
	}
	sb.WriteString(ind + "</div>\n")
}

func writeStaticEmptyState(sb *strings.Builder, e *Element, ind string) {
	icon := attrLiteral(e, "icon")
	if icon == "" {
		icon = "▢"
	}
	hint := attrLiteral(e, "hint")
	label := partsContent(e.text)
	if strings.TrimSpace(label) == "" {
		label = "Nothing here"
	}
	sb.WriteString(ind + "<div class=\"sui-empty")
	if attrBoolDefault(e, "compact", false) {
		sb.WriteString(" sui-empty-compact")
	}
	staticClassSuffix(sb, e)
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString("\">\n")
	sb.WriteString(ind + "  <span class=\"sui-empty-icon\" aria-hidden=\"true\">" + htmlEscape(icon) + "</span>\n")
	sb.WriteString(ind + "  <span class=\"sui-empty-title\">" + htmlEscape(label) + "</span>\n")
	if hint != "" {
		sb.WriteString(ind + "  <span class=\"sui-empty-hint\">" + htmlEscape(hint) + "</span>\n")
	}
	if len(e.children) > 0 {
		sb.WriteString(ind + "  <div class=\"sui-empty-body\">\n")
		writeStaticNodes(sb, e.children, ind+"    ")
		sb.WriteString(ind + "  </div>\n")
	}
	sb.WriteString(ind + "</div>\n")
}

func writeStaticSkeleton(sb *strings.Builder, e *Element, ind string) {
	lines := attrIntDefault(e, "lines", 3)
	if lines < 1 {
		lines = 1
	}
	if lines > 20 {
		lines = 20
	}
	width := layoutSafeCSSValue(attrLiteral(e, "width"))
	if width == "" {
		width = "100%"
	}
	height := layoutSafeCSSValue(attrLiteral(e, "height"))
	if height == "" {
		height = "12px"
	}
	gap := layoutSafeCSSValue(attrLiteral(e, "gap"))
	if gap == "" {
		gap = "8px"
	}
	duration := animationSafeValue(attrLiteral(e, "duration"))
	if duration == "" {
		duration = "1.4s"
	}
	circle := attrBoolDefault(e, "circle", false)
	animated := attrBoolDefault(e, "animated", true)
	sb.WriteString(ind + "<div class=\"sui-skeleton")
	if circle {
		sb.WriteString(" sui-skeleton-circle")
	}
	if !animated {
		sb.WriteString(" sui-skeleton-static")
	}
	staticClassSuffix(sb, e)
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString("\" style=\"--sui-sk-width:" + htmlEscape(width) + ";--sui-sk-height:" + htmlEscape(height) + ";--sui-sk-gap:" + htmlEscape(gap) + ";--sui-sk-duration:" + htmlEscape(duration) + "\">\n")
	if circle {
		sb.WriteString(ind + "  <span class=\"sui-skeleton-dot\"></span>\n")
	} else {
		for i := 0; i < lines; i++ {
			sb.WriteString(ind + "  <span class=\"sui-skeleton-line\"></span>\n")
		}
	}
	sb.WriteString(ind + "</div>\n")
}

func staticIntAttr(e *Element, key string, def int) int {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
		return a.expr.i
	}
	if n, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text))); err == nil {
		return n
	}
	return def
}

func writeStaticCarousel(sb *strings.Builder, e *Element, ind string) {
	slides := childElemsByTag(e, "slide")
	index := staticIntAttr(e, "value", 0)
	if index < 0 || index >= len(slides) {
		index = 0
	}
	controls := attrBoolDefault(e, "controls", true)
	indicators := attrBoolDefault(e, "indicators", true)
	loop := attrBoolDefault(e, "loop", true)
	autoplay := attrBoolDefault(e, "autoplay", false)
	interval := attrIntDefault(e, "interval", 4000)
	sb.WriteString(ind + "<div class=\"sui-carousel")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-carousel data-sui-carousel-index=\"" + strconv.Itoa(index) + "\" data-sui-carousel-count=\"" + strconv.Itoa(len(slides)) + "\"")
	if loop {
		sb.WriteString(" data-sui-carousel-loop=\"1\"")
	}
	if autoplay {
		sb.WriteString(" data-sui-carousel-autoplay=\"1\" data-sui-carousel-interval=\"" + strconv.Itoa(interval) + "\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	style := ""
	if h := layoutSafeCSSValue(attrLiteral(e, "height")); h != "" {
		style = " style=\"--sui-carousel-height:" + htmlEscape(h) + "\""
	}
	sb.WriteString(style + ">\n")
	sb.WriteString(ind + "  <div class=\"sui-carousel-frame\"><div class=\"sui-carousel-track\" style=\"transform:translateX(-" + strconv.Itoa(index*100) + "%)\">\n")
	for i, sl := range slides {
		label := attrLiteral(sl, "label")
		if label == "" {
			label = "Slide " + strconv.Itoa(i+1)
		}
		sb.WriteString(ind + "    <div class=\"sui-carousel-slide" + staticClassSuffixStr(sl) + "\" data-sui-slide-index=\"" + strconv.Itoa(i) + "\" role=\"group\" aria-label=\"" + htmlEscape(label) + "\">\n")
		writeStaticNodes(sb, sl.children, ind+"      ")
		sb.WriteString(ind + "    </div>\n")
	}
	sb.WriteString(ind + "  </div></div>\n")
	if controls || indicators {
		sb.WriteString(ind + "  <div class=\"sui-carousel-nav\">\n")
		if controls {
			sb.WriteString(ind + "    <button type=\"button\" class=\"sui-carousel-prev\" data-sui-carousel-prev aria-label=\"Previous slide\">&#8249;</button>\n")
		}
		if indicators {
			sb.WriteString(ind + "    <div class=\"sui-carousel-dots\" role=\"tablist\" aria-label=\"Slides\">\n")
			for i := range slides {
				cls := "sui-carousel-dot"
				if i == index {
					cls += " active"
				}
				sb.WriteString(ind + "      <button type=\"button\" class=\"" + cls + "\" data-sui-carousel-dot=\"" + strconv.Itoa(i) + "\" aria-label=\"Slide " + strconv.Itoa(i+1) + "\"></button>\n")
			}
			sb.WriteString(ind + "    </div>\n")
		}
		if controls {
			sb.WriteString(ind + "    <button type=\"button\" class=\"sui-carousel-next\" data-sui-carousel-next aria-label=\"Next slide\">&#8250;</button>\n")
		}
		sb.WriteString(ind + "  </div>\n")
	}
	sb.WriteString(ind + "</div>\n")
}

// staticClassSuffixStr returns the user class tokens for inline composition.
func staticClassSuffixStr(e *Element) string {
	a := findAttr(e, "class")
	if a == nil || a.hasExpr {
		return ""
	}
	v := strings.TrimSpace(partsContent(a.text))
	if v == "" {
		return ""
	}
	return " " + htmlEscape(v)
}

func writeStaticStepper(sb *strings.Builder, e *Element, ind string) {
	value := staticIntAttr(e, "value", 0)
	vertical := attrBoolDefault(e, "vertical", false)
	steps := childElemsByTag(e, "step")
	sb.WriteString(ind + "<div class=\"sui-stepper")
	if vertical {
		sb.WriteString(" sui-stepper-vertical")
	}
	staticClassSuffix(sb, e)
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString("\">\n")
	for i, st := range steps {
		cls := "sui-step"
		if value > i {
			cls += " done"
		} else if value == i {
			cls += " active"
		}
		sb.WriteString(ind + "  <div class=\"" + cls + staticClassSuffixStr(st) + "\" data-sui-step-index=\"" + strconv.Itoa(i) + "\">\n")
		sb.WriteString(ind + "    <span class=\"sui-step-num\">")
		if icon := attrLiteral(st, "icon"); icon != "" {
			sb.WriteString(htmlEscape(icon))
		} else if value > i {
			sb.WriteString("&#10003;")
		} else {
			sb.WriteString(strconv.Itoa(i + 1))
		}
		sb.WriteString("</span>\n")
		sb.WriteString(ind + "    <span class=\"sui-step-text\"><span class=\"sui-step-label\">" + htmlEscape(partsContent(st.text)) + "</span>")
		if desc := attrLiteral(st, "description"); desc != "" {
			sb.WriteString("<span class=\"sui-step-desc\">" + htmlEscape(desc) + "</span>")
		}
		sb.WriteString("</span>\n")
		sb.WriteString(ind + "  </div>\n")
		if i < len(steps)-1 {
			sb.WriteString(ind + "  <span class=\"sui-step-line\"></span>\n")
		}
	}
	sb.WriteString(ind + "</div>\n")
}

func writeStaticCommandbar(sb *strings.Builder, e *Element, ind string) {
	overlay := findAttr(e, "open") != nil
	open := false
	if overlay {
		if a := findAttr(e, "open"); a != nil && !a.hasExpr {
			v := strings.TrimSpace(strings.ToLower(partsContent(a.text)))
			open = v != "" && v != "false" && v != "0"
		}
	}
	placeholder := attrLiteral(e, "placeholder")
	if placeholder == "" {
		placeholder = strings.TrimSpace(partsContent(e.text))
	}
	if placeholder == "" {
		placeholder = "Type a command..."
	}
	hotkey := strings.ToLower(attrLiteral(e, "hotkey"))
	sb.WriteString(ind + "<div class=\"sui-commandbar")
	if overlay {
		sb.WriteString(" sui-commandbar-overlay")
	}
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-commandbar" + dataAnimAttrs(e, "fade"))
	if overlay {
		sb.WriteString(" data-sui-commandbar-overlay data-sui-commandbar-open=\"" + bool01(open) + "\"")
		if hotkey != "" {
			sb.WriteString(" data-sui-commandbar-hotkey=\"" + htmlEscape(hotkey[:1]) + "\"")
		}
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	if overlay {
		sb.WriteString(ind + "  <div class=\"sui-commandbar-backdrop\" data-sui-commandbar-close></div>\n")
	}
	sb.WriteString(ind + "  <div class=\"sui-commandbar-panel\" role=\"dialog\" aria-modal=\"" + bool01(overlay) + "\">\n")
	sb.WriteString(ind + "    <input class=\"sui-commandbar-input\" type=\"text\" placeholder=\"" + htmlEscape(placeholder) + "\" aria-label=\"" + htmlEscape(placeholder) + "\" autocomplete=\"off\" spellcheck=\"false\">\n")
	sb.WriteString(ind + "    <div class=\"sui-commandbar-list\" role=\"listbox\">\n")
	for i, cmd := range childElemsByTag(e, "command") {
		cls := "sui-command"
		if i == 0 {
			cls += " active"
		}
		label := partsContent(cmd.text)
		shortcut := attrLiteral(cmd, "shortcut")
		sb.WriteString(ind + "      <button type=\"button\" class=\"" + cls + staticClassSuffixStr(cmd) + "\" role=\"option\" data-sui-command-index=\"" + strconv.Itoa(i) + "\" data-sui-command-search=\"" + htmlEscape(strings.ToLower(label+" "+shortcut)) + "\" data-sui-command-label=\"" + htmlEscape(label) + "\">")
		sb.WriteString("<span class=\"sui-command-label\">" + htmlEscape(label) + "</span>")
		if shortcut != "" {
			sb.WriteString("<span class=\"sui-command-shortcut\">" + htmlEscape(shortcut) + "</span>")
		}
		sb.WriteString("</button>\n")
	}
	sb.WriteString(ind + "    </div>\n")
	sb.WriteString(ind + "    <div class=\"sui-commandbar-empty\" hidden>No matching commands</div>\n")
	sb.WriteString(ind + "  </div>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticDropzone(sb *strings.Builder, e *Element, ind string) {
	label := partsContent(e.text)
	if strings.TrimSpace(label) == "" {
		label = "Drop files here"
	}
	hint := "Drag files here or click to choose"
	if a := findAttr(e, "hint"); a != nil && !a.hasExpr {
		if v := strings.TrimSpace(partsContent(a.text)); v != "" {
			hint = v
		}
	} else if datepickerBool(e, "directory") {
		hint = "Drag a folder here or click to choose"
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-dropzone")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-dropzone=\"1\"")
	if a := findAttr(e, "dropFiles"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-dropfiles=\"" + setterName(a.expr.name) + "\"")
	}
	if a := findAttr(e, "pickFiles"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-pickfiles=\"" + setterName(a.expr.name) + "\"")
	} else if a := findAttr(e, "dropFiles"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-pickfiles=\"" + setterName(a.expr.name) + "\"")
	}
	if a := findAttr(e, "accept"); a != nil {
		sb.WriteString(" data-sui-pick-accept=\"" + htmlAttrValue(a, false) + "\"")
	}
	if datepickerBool(e, "multiple") {
		sb.WriteString(" data-sui-pick-multiple=\"1\"")
	}
	if datepickerBool(e, "directory") {
		sb.WriteString(" data-sui-pick-directory=\"1\"")
	}
	if a := findAttr(e, "pickMode"); a != nil && !a.hasExpr {
		if v := strings.ToLower(strings.TrimSpace(partsContent(a.text))); v != "" {
			sb.WriteString(" data-sui-pick-mode=\"" + v + "\"")
		}
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <span class=\"sui-dropzone-icon\" aria-hidden=\"true\">&#8681;</span>\n")
	sb.WriteString(ind + "  <span class=\"sui-dropzone-title\">" + htmlEscape(label) + "</span>\n")
	sb.WriteString(ind + "  <span class=\"sui-dropzone-hint\">" + htmlEscape(hint) + "</span>\n")
	writeStaticNodes(sb, e.children, ind+"  ")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticSplitter(sb *strings.Builder, e *Element, ind string) {
	direction := "horizontal"
	if a := findAttr(e, "direction"); a != nil && !a.hasExpr {
		if strings.ToLower(strings.TrimSpace(partsContent(a.text))) == "vertical" {
			direction = "vertical"
		}
	}
	initial := "50%"
	if a := findAttr(e, "initial"); a != nil && !a.hasExpr {
		if v := layoutSafeCSSValue(partsContent(a.text)); v != "" {
			initial = v
		}
	}
	min := "120px"
	if a := findAttr(e, "min"); a != nil && !a.hasExpr {
		if v := layoutSafeCSSValue(partsContent(a.text)); v != "" {
			min = v
		}
	}
	barOrientation := "vertical"
	if direction == "vertical" {
		barOrientation = "horizontal"
	}
	var panes []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "pane" {
			panes = append(panes, ch.elem)
		}
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-splitter sui-splitter-" + direction)
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-splitter data-sui-splitter-direction=\"" + direction + "\" data-sui-splitter-min=\"" + htmlEscape(min) + "\"")
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	for i := 0; i < len(panes); i++ {
		sb.WriteString(ind + "  <div class=\"sui-split-pane\"")
		if i == 0 {
			sb.WriteString(" style=\"flex-basis:" + htmlEscape(initial) + "\"")
		}
		sb.WriteString(">\n")
		writeStaticNodes(sb, panes[i].children, ind+"    ")
		sb.WriteString(ind + "  </div>\n")
		if i == 0 && len(panes) > 1 {
			sb.WriteString(ind + "  <div class=\"sui-split-bar\" role=\"separator\" tabindex=\"0\" aria-orientation=\"" + barOrientation + "\"></div>\n")
		}
	}
	sb.WriteString(ind + "</div>\n")
}

func writeStaticDrawer(sb *strings.Builder, e *Element, ind string) {
	open := false
	if a := findAttr(e, "open"); a != nil && !a.hasExpr {
		v := strings.TrimSpace(strings.ToLower(partsContent(a.text)))
		open = v != "" && v != "false" && v != "0"
	}
	side := "right"
	if a := findAttr(e, "side"); a != nil && !a.hasExpr {
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		if v == "left" || v == "right" {
			side = v
		}
	}
	width := "320px"
	if a := findAttr(e, "width"); a != nil && !a.hasExpr {
		if v := layoutSafeCSSValue(partsContent(a.text)); v != "" {
			width = v
		}
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-drawer-host")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-drawer data-sui-drawer-open=\"" + bool01(open) + "\"")
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <div class=\"sui-drawer-backdrop\" data-sui-drawer-backdrop></div>\n")
	sb.WriteString(ind + "  <aside class=\"sui-drawer sui-drawer-" + side + "\" role=\"dialog\" aria-modal=\"true\" style=\"--sui-drawer-width:" + htmlEscape(width) + "\">\n")
	sb.WriteString(ind + "    <div class=\"sui-drawer-head\"><b>")
	sb.WriteString(htmlEscape(partsContent(e.text)))
	sb.WriteString("</b><button type=\"button\" class=\"sui-drawer-close\" data-sui-drawer-close aria-label=\"Close\">&#10005;</button></div>\n")
	sb.WriteString(ind + "    <div class=\"sui-drawer-body\">\n")
	writeStaticNodes(sb, e.children, ind+"      ")
	sb.WriteString(ind + "    </div>\n")
	sb.WriteString(ind + "  </aside>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticPopover(sb *strings.Builder, e *Element, ind string) {
	pos := attrChoice(e, "position", []string{"auto", "top", "right", "bottom", "left"}, "auto")
	label := partsContent(e.text)
	if strings.TrimSpace(label) == "" {
		label = "Info"
	}
	sb.WriteString(ind)
	sb.WriteString("<span class=\"sui-popover")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-popover data-sui-popover-position=\"" + pos + "\"" + dataAnimAttrs(e, "scale"))
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <button type=\"button\" class=\"sui-popover-trigger\" data-sui-popover-trigger aria-haspopup=\"dialog\" aria-expanded=\"false\">")
	sb.WriteString(htmlEscape(label))
	sb.WriteString("</button>\n")
	sb.WriteString(ind + "  <div class=\"sui-popover-panel\" role=\"dialog\" hidden>\n")
	writeStaticNodes(sb, e.children, ind+"    ")
	sb.WriteString(ind + "  </div>\n")
	sb.WriteString(ind + "</span>\n")
}

func writeStaticMenu(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-menu")
	staticClassSuffix(sb, e)
	if e.tag == "dropdown" {
		if pos := findAttr(e, "position"); pos != nil && !pos.hasExpr && strings.EqualFold(strings.TrimSpace(partsContent(pos.text)), "right") {
			sb.WriteString(" sui-dropdown-right")
		}
	}
	sb.WriteString("\" role=\"none\"")
	if e.tag == "dropdown" {
		sb.WriteString(" data-sui-dropdown-placement=\"" + attrChoice(e, "placement", []string{"auto", "bottom", "top", "left", "right"}, "auto") + "\"" + dataAnimAttrs(e, "fade"))
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind)
	sb.WriteString("  <button type=\"button\" class=\"sui-menu-button\" role=\"menuitem\" aria-haspopup=\"true\" aria-expanded=\"false\">")
	staticText(sb, e.text)
	sb.WriteString("</button>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-menu-popup\" role=\"menu\">\n")
	writeStaticNodes(sb, e.children, ind+"    ")
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

// writeStaticPdfViewer renders the static form of the pdf viewer (the
// designer preview and the static HTML export). It is an <iframe> with the
// document type and without a sandbox attribute: every sandbox token
// disables the browser's native PDF viewer.
func writeStaticPdfViewer(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<iframe class=\"sui-pdfviewer")
	staticClassSuffix(sb, e)
	sb.WriteString("\" type=\"application/pdf\" referrerpolicy=\"no-referrer\"")
	if src := findAttr(e, "src"); src != nil && !src.hasExpr {
		sb.WriteString(" src=\"")
		sb.WriteString(htmlEscape(partsContent(src.text)))
		sb.WriteString("\"")
	}
	for _, key := range []string{"width", "height"} {
		if a := findAttr(e, key); a != nil && !a.hasExpr {
			sb.WriteString(" " + key + "=\"")
			sb.WriteString(htmlEscape(partsContent(a.text)))
			sb.WriteString("\"")
		}
	}
	sb.WriteString("></iframe>\n")
}

func staticLinkOpen(e *Element) string {
	a := findAttr(e, "link")
	if a == nil {
		return ""
	}
	if a.hasExpr {
		return "<a href=\"\">"
	}
	return "<a href=\"" + htmlEscape(partsContent(a.text)) + "\">"
}

func staticClassSuffix(sb *strings.Builder, e *Element) {
	if a := findAttr(e, "class"); a != nil && !a.hasExpr {
		if c := partsContent(a.text); c != "" {
			sb.WriteString(" ")
			sb.WriteString(htmlEscape(c))
		}
	}
}

func staticUserAttrs(sb *strings.Builder, e *Element, dlId string) {
	staticUserAttrsOpt(sb, e, dlId, false)
}

// staticUserAttrsNoClass is used by controls that already emitted their
// built-in class, so the user class is appended exactly once by
// staticClassSuffix.
func staticUserAttrsNoClass(sb *strings.Builder, e *Element, dlId string) {
	staticUserAttrsOpt(sb, e, dlId, true)
}

func staticUserAttrsOpt(sb *strings.Builder, e *Element, dlId string, skipClass bool) {
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		if _, isEv := eventAttrJs[a.key]; isEv {
			continue
		}
		if skipClass && a.key == "class" {
			continue
		}
		if (e.tag == "modal" && a.key == "name") ||
			(e.tag == "modalbutton" && (a.key == "variant" || a.key == "close")) {
			continue
		}
		if e.tag == "toast" && (a.key == "position" || a.key == "offset" ||
			a.key == "hAlign" || a.key == "vAlign" || a.key == "hOffset" || a.key == "vOffset" ||
			a.key == "duration" || a.key == "max" || a.key == "icons") {
			continue
		}
		if e.tag == "combobox" && (a.key == "value" || a.key == "options" || a.key == "placeholder" ||
			a.key == "empty" || a.key == "disabled" || a.key == "required" ||
			a.key == "maxOptionLength" || a.key == "maxOptions") {
			continue
		}
		if e.tag == "datepicker" && (a.key == "value" || a.key == "mode" || a.key == "min" ||
			a.key == "max" || a.key == "step" || a.key == "disabled" || a.key == "required") {
			continue
		}
		if e.tag == "tree" && (a.key == "value" || a.key == "editMode" || a.key == "editDepth" ||
			a.key == "name" || a.key == "renameKey" || a.key == "renameValue" || a.key == "renameIcon" || a.key == "showRenameIcon" ||
			a.key == "guides" || a.key == "truncate" || a.key == "checkbox" || a.key == "counts" || a.key == "rowHeight" ||
			a.key == "multiSelect" || a.key == "selection" ||
			a.key == "search" || a.key == "drag" || a.key == "contextMenu" || a.key == "columns" || a.key == "loadKey" ||
			a.key == "virtual" || a.key == "columnLabels" || a.key == "columnWidths" || a.key == "sort" || a.key == "persistSelection" ||
			a.key == "columnsEnabled" || a.key == "tooltips" || a.key == "fontStep" || a.key == "onLoad" || a.key == "onMove" || a.key == "onContextMenu") {
			continue
		}
		if e.tag == "treeitem" && (a.key == "key" || a.key == "open" || a.key == "rename" || a.key == "renameIcon" || a.key == "showRenameIcon" ||
			a.key == "icon" || a.key == "iconOpen" || a.key == "badge" || a.key == "badgeType" || a.key == "meta" || a.key == "checked" ||
			a.key == "lazy" || a.key == "size" || a.key == "date" || a.key == "status" || a.key == "cells") {
			continue
		}
		if e.tag == "tabs" && (a.key == "position" || a.key == "variant") {
			continue
		}
		if e.tag == "dropdown" && a.key == "position" {
			continue
		}
		if e.tag == "dropdownitem" && a.key == "separator" {
			continue
		}
		if e.tag == "dropdownitem" && a.key == "disabled" {
			continue
		}
		if e.tag == "popover" && a.key == "position" {
			continue
		}
		if e.tag == "drawer" && (a.key == "open" || a.key == "side" || a.key == "width" || a.key == "onClose") {
			continue
		}
		if e.tag == "splitter" && (a.key == "direction" || a.key == "initial" || a.key == "min") {
			continue
		}
		if e.tag == "dropzone" && (a.key == "accept" || a.key == "multiple" || a.key == "directory" || a.key == "hint") {
			continue
		}
		if e.tag == "timelineitem" && (a.key == "date" || a.key == "status") {
			continue
		}
		if e.tag == "emptystate" && (a.key == "icon" || a.key == "hint" || a.key == "compact") {
			continue
		}
		if e.tag == "skeleton" && (a.key == "lines" || a.key == "width" || a.key == "height" || a.key == "circle" || a.key == "animated" || a.key == "gap") {
			continue
		}
		if e.tag == "carousel" && (a.key == "value" || a.key == "autoplay" || a.key == "interval" || a.key == "loop" || a.key == "controls" || a.key == "indicators" || a.key == "height") {
			continue
		}
		if e.tag == "slide" && a.key == "label" {
			continue
		}
		if e.tag == "stepper" && (a.key == "value" || a.key == "clickable" || a.key == "vertical") {
			continue
		}
		if e.tag == "step" && (a.key == "description" || a.key == "icon") {
			continue
		}
		if e.tag == "commandbar" && (a.key == "open" || a.key == "hotkey" || a.key == "placeholder") {
			continue
		}
		if e.tag == "command" && a.key == "shortcut" {
			continue
		}
		if e.tag == "dropdown" && (a.key == "placement" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "popover" && (a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "combobox" && (a.key == "placement" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "drawer" && (a.key == "duration" || a.key == "easing") {
			continue
		}
		if e.tag == "carousel" && (a.key == "duration" || a.key == "easing") {
			continue
		}
		if e.tag == "commandbar" && (a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "skeleton" && a.key == "duration" {
			continue
		}
		if (e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio") &&
			(a.key == "shape" || a.key == "theme" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay" || a.key == "stagger") {
			continue
		}
		if e.tag == "tablelayout" && (a.key == "columns" || a.key == "rows" ||
			a.key == "width" || a.key == "height" ||
			a.key == "columnWidths" || a.key == "rowHeights" ||
			a.key == "gapX" || a.key == "gapY" ||
			a.key == "borderStyle" || a.key == "borderWidth" || a.key == "borderColor" ||
			a.key == "cellMinHeight" || a.key == "responsive" || a.key == "semantic" || a.key == "label" || a.key == "style") {
			continue
		}
		if e.tag == "table" && (a.key == "name" || a.key == "editMode" ||
			a.key == "columns" || a.key == "columnLabels" || a.key == "columnWidths" ||
			a.key == "columnAlign" || a.key == "columnTypes" || a.key == "columnFormats" || a.key == "columnOptions" ||
			a.key == "cellText" || a.key == "sortable" || a.key == "filterable" || a.key == "pageSize" || a.key == "selection" ||
			a.key == "multiSelect" || a.key == "virtual" || a.key == "showAddTable" || a.key == "structureEditable" ||
			a.key == "onCellChange" || a.key == "onRowChange" || a.key == "onTableChange" ||
			a.key == "onRowSelect" || a.key == "onSort" || a.key == "onCellClick" || a.key == "onRowClick") {
			continue
		}
		if chartTag(e.tag) && (a.key == "name" || a.key == "data" || a.key == "editMode" || a.key == "chartType" || a.key == "format" ||
			a.key == "x" || a.key == "y" || a.key == "width" || a.key == "height" || a.key == "overflow" ||
			a.key == "zoom" || a.key == "pointWidth" || a.key == "pointSize" || a.key == "bins" || a.key == "min" || a.key == "max" ||
			a.key == "tableView" || a.key == "tableFilters" || a.key == "tableAddRow" || a.key == "tableDeleteRow" || a.key == "tableSteppers" || a.key == "tableContextMenu" || a.key == "tableSelection" || a.key == "criticalPath" || a.key == "holidays" || a.key == "weekend" || a.key == "timezone" || a.key == "showNonWorkingDays" || a.key == "progressColumn" || a.key == "settingsButton" || a.key == "showZoom" || a.key == "showAdd" || a.key == "showHistory" || a.key == "stateHints" || a.key == "onTaskChange" || a.key == "onDataChange" || a.key == "onCalendarChange" || a.key == "onResourceChange" || a.key == "onCommit" ||

			a.key == "showGrid" || a.key == "showLegend" || a.key == "showValues" ||
			a.key == "showResources" || a.key == "showProgress" ||
			a.key == "showSlack" || a.key == "showCosts" || a.key == "statusColors" ||
			a.key == "ganttScale" || a.key == "autoSchedule" || a.key == "linkColor" ||
			a.key == "linkWidth" || a.key == "linkGap" || a.key == "linkStyle" || a.key == "arrowType" || a.key == "arrowSize" || a.key == "barBorderColor" || a.key == "stateRouting" || a.key == "stateTagStyle" || a.key == "stateTagLink" || a.key == "stateLinkTagStyle" || a.key == "stateLinkTagLink" || a.key == "stateShapeTagStyle" || a.key == "stateShapeTagLink" || a.key == "stateTagFont" || a.key == "stateTagFontSize" || a.key == "stateShapeTagFont" || a.key == "stateShapeTagFontSize" || a.key == "stateLinkTagFont" || a.key == "stateLinkTagFontSize" || a.key == "stateRouteAlgorithm" || a.key == "stateLinkArcRadius" || a.key == "stateGrid" || a.key == "stateGridGap" || a.key == "stateGridColor") {
			continue
		}
		if a.key == "tooltip" {
			if !a.hasExpr {
				sb.WriteString(" data-sui-tooltip=\"")
				sb.WriteString(htmlEscape(partsContent(a.text)))
				sb.WriteString("\"")
			}
			continue
		}
		if a.key == "onCheck" {
			continue
		}
		if a.key == "head" || a.key == "options" || a.key == "link" || a.key == "srcAllow" ||
			a.key == "dropFiles" || a.key == "pickFiles" || a.key == "pickMode" || a.key == "dragKey" || a.key == "dropKey" {
			continue
		}
		if a.key == "motion" || a.key == "motionKey" || a.key == "motionDuration" || a.key == "motionEasing" || a.key == "motionDelay" || a.key == "motionExit" || a.key == "motionGraph" {
			continue
		}
		if e.tag != "tcell" && (a.key == "colspan" || a.key == "rowspan") {
			continue
		}
		if a.key == "cellPadding" || a.key == "cellBackground" ||
			a.key == "cellBorderStyle" || a.key == "cellBorderWidth" || a.key == "cellBorderColor" {
			continue
		}
		if a.key == "required" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
				sb.WriteString(" required")
			}
			continue
		}
		if a.key == "directory" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
				sb.WriteString(" webkitdirectory")
			}
			continue
		}
		if a.key == "label" {
			continue
		}
		sb.WriteString(" ")
		if a.key == "inputtype" {
			sb.WriteString("type")
		} else {
			sb.WriteString(a.key)
		}
		sb.WriteString("=\"")
		if a.hasExpr {
			if a.expr != nil && a.expr.kind == "int" {
				sb.WriteString(strconv.Itoa(a.expr.i))
			} else if a.expr != nil && a.expr.kind == "bool" {
				if a.expr.b {
					sb.WriteString("true")
				} else {
					sb.WriteString("false")
				}
			}
			// dynamic bindings left as-is in static output
		} else {
			sb.WriteString(htmlEscape(partsContent(a.text)))
		}
		sb.WriteString("\"")
	}
	if dlId != "" {
		sb.WriteString(" list=\"" + dlId + "\"")
	}
}

func staticText(sb *strings.Builder, parts []StrPart) {
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "text" {
			sb.WriteString(htmlEscape(parts[i].text))
		}
	}
}
