package main

import "testing"
import "strings"

// htmlgen_test.go — tests for PageTitle / GenerateIndexHtml / GenerateCss
// (Agent D, codegen). Sources are parsed with the already-present ParseSource;
// the typechecker is out of scope here.

// diagText joins diagnostic messages for failure reporting.
func diagText(dl DiagList) string {
	var sb strings.Builder
	for i := 0; i < len(dl.list); i++ {
		sb.WriteString("|")
		sb.WriteString(dl.list[i].msg)
	}
	return sb.String()
}

// htmlSampleSrc is a view exercising every structure htmlgen must traverse:
// title, h1, a row with two buttons sharing a class, an input with a
// placeholder, and one if/else and one list node.
func htmlSampleSrc() string {
	return "view main {\n" +
		"\ttitle \"My App\"\n" +
		"\th1 \"Hello {name}\" class: \"hero\"\n" +
		"\trow {\n" +
		"\t\tbutton \"Save\" class: \"btn btn-primary\"\n" +
		"\t\tbutton \"Cancel\" class: \"btn btn-primary\"\n" +
		"\t}\n" +
		"\tinput \"Enter name\" placeholder: \"Your name\"\n" +
		"\tif ready {\n" +
		"\t\ttext \"Ready\" class: \"status-ok\"\n" +
		"\t} else {\n" +
		"\t\ttext \"Busy\"\n" +
		"\t}\n" +
		"\tlist item in items {\n" +
		"\t\tp \"{item}\" class: \"list-item\"\n" +
		"\t}\n" +
		"}\n"
}

func TestBaseElementsUseThemeVariables(t *testing.T) {
	for _, want := range []string{
		`.sui-view {`,
		`background: var(--sui-panel, #1b1f27);`,
		`color: var(--sui-fg, #e6e6e6);`,
		`h1, h2, h3, h4, h5, h6 {`,
		`color: var(--sui-fg, #f2f4f6);`,
		`button {`,
		`background: var(--sui-button-bg, #2d3340);`,
		`color: var(--sui-button-fg, #e6e6e6);`,
		`:where(button):hover { background: var(--sui-button-hover, #3a4356); }`,
		`a {`,
		`color: var(--sui-link, var(--sui-accent, #7db3ff));`,
		`code, pre {`,
		`background: var(--sui-panel2, #10131a);`,
		`thead th {`,
		`background: var(--sui-panel, #202732);`,
		`.sui-menubar {`,
		`background: var(--sui-panel, #1d2129);`,
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("base theme variable missing %q", want)
		}
	}
}

func TestSpinnerStructuralAnimations(t *testing.T) {
	for _, want := range []string{
		`.sui-spinner[variant="hourglass"]`,
		`.sui-spinner[variant="hourglass"]::after`,
		`@keyframes sui-spinner-hourglass`,
		`@keyframes sui-spinner-hourglass-sand`,
		`@keyframes sui-spinner-hourglass-flow`,
		`82%, 88%, 94% { transform: rotate(180deg); background-size: 100% 42%, 100% 100%; }`,
		`82%, 100% { height: 0%; }`,
		`.sui-spinner[variant="fill-flip"]`,
		`@keyframes sui-spinner-fill-flip`,
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("spinner base CSS missing %q", want)
		}
	}
}

func TestChartDarkThemeControlsUseChartForeground(t *testing.T) {
	for _, want := range []string{
		`--sui-chart-control-bg: #1d2129;`,
		`--sui-chart-control-border: #3d4556;`,
		`--sui-chart-control-fg: var(--sui-chart-fg, #e6e6e6);`,
		`--sui-chart-control-hover-bg: #2d3340;`,
		`--sui-chart-control-hover-fg: #ffffff;`,
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("dark chart control theme missing %q", want)
		}
	}
}

func TestHtmlPageTitle(t *testing.T) {
	prog, dl := ParseSource(htmlSampleSrc(), "t.sui")
	if dl.Count() != 0 {
		t.Errorf("parse: %v diagnostics: %v", dl.Count(), diagText(dl))
		return
	}
	if PageTitle(prog) != "My App" {
		t.Errorf("PageTitle: got %q, want %q", PageTitle(prog), "My App")
	}
}

func TestHtmlIndex(t *testing.T) {
	prog, dl := ParseSource(htmlSampleSrc(), "t.sui")
	if dl.Count() != 0 {
		t.Errorf("parse: %v diagnostics: %v", dl.Count(), diagText(dl))
		return
	}
	html := GenerateIndexHtml(prog)
	if html != `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>My App</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
<div id="app"></div>
<script src="app.js"></script>
</body>
</html>
` {
		t.Errorf("index: skeleton does not match SPEC §5 exactly")
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("index: missing <!DOCTYPE html>")
	}
	if !strings.Contains(html, "<title>My App</title>") {
		t.Errorf("index: missing <title>My App</title>")
	}
	if !strings.Contains(html, `<div id="app"></div>`) {
		t.Errorf("index: missing app div")
	}
	if !strings.Contains(html, "style.css") {
		t.Errorf("index: missing style.css link")
	}
	if !strings.Contains(html, "app.js") {
		t.Errorf("index: missing app.js script")
	}
}

func TestHtmlCss(t *testing.T) {
	prog, dl := ParseSource(htmlSampleSrc(), "t.sui")
	if dl.Count() != 0 {
		t.Errorf("parse: %v diagnostics: %v", dl.Count(), diagText(dl))
		return
	}
	css := GenerateCss(prog)
	needles := []string{".sui-view", ".row", ".col", ".spacer", "button", "input", ".btn", ".btn-primary"}
	for i := 0; i < len(needles); i++ {
		if !strings.Contains(css, needles[i]) {
			t.Errorf("css: missing %q", needles[i])
		}
	}
	if !strings.Contains(css, ".status-ok { /* generated */ }") {
		t.Errorf("css: no class placeholder collected from if branch")
	}
	if !strings.Contains(css, ".list-item { /* generated */ }") {
		t.Errorf("css: no class placeholder collected from list branch")
	}
	if !strings.Contains(css, ".btn-primary { /* generated */ }") {
		t.Errorf("css: no .btn-primary placeholder rule")
	}
	if !strings.Contains(css, ".btn { /* generated */ }") {
		t.Errorf("css: no .btn placeholder rule")
	}
}

func TestHtmlTitleEscaping(t *testing.T) {
	src := "view v {\n\ttitle \"a <b> & c\"\n\th1 \"t\"\n}\n"
	prog, dl := ParseSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("parse: %v diagnostics: %v", dl.Count(), diagText(dl))
		return
	}
	if PageTitle(prog) != "a <b> & c" {
		t.Errorf("PageTitle should stay raw: got %q", PageTitle(prog))
	}
	html := GenerateIndexHtml(prog)
	if !strings.Contains(html, "<title>a &lt;b&gt; &amp; c</title>") {
		t.Errorf("index: title not escaped, got %q", html)
	}
}

func TestHtmlEmptyTitle(t *testing.T) {
	src := "view v {\n\th1 \"no title\"\n}\n"
	prog, dl := ParseSource(src, "t.sui")
	if dl.Count() != 0 {
		t.Errorf("parse: %v diagnostics: %v", dl.Count(), diagText(dl))
		return
	}
	if PageTitle(prog) != "" {
		t.Errorf("PageTitle: got %q, want empty", PageTitle(prog))
	}
	if !strings.Contains(GenerateIndexHtml(prog), "<title></title>") {
		t.Errorf("index: expected an empty <title></title>")
	}
}

func TestHtmlNoProg(t *testing.T) {
	if PageTitle(nil) != "" {
		t.Errorf("PageTitle(nil): want empty")
	}
	if !strings.Contains(GenerateIndexHtml(nil), "<title></title>") {
		t.Errorf("GenerateIndexHtml(nil): want empty title")
	}
	if strings.Contains(GenerateCss(nil), ".btn") {
		t.Errorf("GenerateCss(nil): no classes expected")
	}
}
