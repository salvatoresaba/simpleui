package core

import (
	"strings"
	"testing"
)

// layout_tracks_test.go - tests for the tracks of a row or of a column
// container: the declared columns, the size of every cell and the validation
// of the percentages.

// layoutTracksProg parses a source and fails the test when the parser reports
// something: the assertions only concern the generated layout.
func layoutTracksProg(t *testing.T, src string) *Prog {
	t.Helper()
	prog, dl := ParseSource(src, "layout.sui")
	if dl.Count() != 0 {
		t.Fatalf("unexpected parse diagnostics: %s", tcDiagMsgs(dl))
	}
	return prog
}

func TestRowLayoutTracksInGeneratedApp(t *testing.T) {
	src := "view main {\n" +
		"\trow columns: \"2\" columnWidths: \"60%, 40%\" width: \"100%\" height: \"200px\" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t\tdiv { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	prog := layoutTracksProg(t, src)
	js := GenerateJs(prog)
	if !strings.Contains(js, "display:grid;grid-template-columns:60% 40%;width:100%;height:200px;") {
		t.Fatalf("the row does not lay its cells out as columns:\n%s", js)
	}
	html := GenerateHtmlPage(prog)
	if !strings.Contains(html, "display:grid;grid-template-columns:60% 40%;width:100%;height:200px;") {
		t.Fatalf("the static page does not lay the cells out as columns:\n%s", html)
	}
}

func TestColumnLayoutTracksInGeneratedApp(t *testing.T) {
	src := "view main {\n" +
		"\tcol rows: \"2\" rowHeights: \"100px, \" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t\tdiv { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	prog := layoutTracksProg(t, src)
	js := GenerateJs(prog)
	// A cell without a size takes the space that the other cells leave free.
	if !strings.Contains(js, "display:grid;grid-template-rows:100px minmax(0,1fr);") {
		t.Fatalf("the column container does not lay its cells out as rows:\n%s", js)
	}
}

func TestRowLayoutWithoutTracksKeepsFlex(t *testing.T) {
	src := "view main {\n" +
		"\trow {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t\tdiv { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	prog := layoutTracksProg(t, src)
	js := GenerateJs(prog)
	if strings.Contains(js, "grid-template-columns") {
		t.Fatalf("a row without tracks must keep the default layout:\n%s", js)
	}
	if !strings.Contains(js, "class=\"row\"") {
		t.Fatalf("the row lost its class:\n%s", js)
	}
}

func TestRowLayoutTrackCountFillsTheMissingCells(t *testing.T) {
	src := "view main {\n" +
		"\trow columns: \"3\" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t}\n" +
		"}\n"
	prog := layoutTracksProg(t, src)
	js := GenerateJs(prog)
	if !strings.Contains(js, "grid-template-columns:minmax(0,1fr) minmax(0,1fr) minmax(0,1fr);") {
		t.Fatalf("the declared columns are not all automatic:\n%s", js)
	}
}

func TestRowLayoutPercentagesCannotExceed100(t *testing.T) {
	src := "view main {\n" +
		"\trow columns: \"2\" columnWidths: \"60%, 50%\" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t\tdiv { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	diagnostics := checkSrc(t, src)
	if !hasMsg(diagnostics, "columnWidths cannot exceed 100%") {
		t.Fatalf("expected the percentage sum diagnostic, got: %s", diagnostics.String())
	}
}

func TestRowLayoutAcceptsValidPercentages(t *testing.T) {
	src := "view main {\n" +
		"\trow columns: \"2\" columnWidths: \"60%, 40%\" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t\tdiv { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	if diagnostics := checkSrc(t, src); diagnostics.HasErrors() {
		t.Fatalf("valid column widths were rejected: %s", diagnostics.String())
	}
}

func TestColumnLayoutPercentagesCannotExceed100(t *testing.T) {
	src := "view main {\n" +
		"\tcol rows: \"2\" rowHeights: \"70%, 40%\" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t\tdiv { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	diagnostics := checkSrc(t, src)
	if !hasMsg(diagnostics, "rowHeights cannot exceed 100%") {
		t.Fatalf("expected the percentage sum diagnostic, got: %s", diagnostics.String())
	}
}

func TestTableLayoutPercentagesCannotExceed100(t *testing.T) {
	src := "view main {\n" +
		"\ttablelayout rows: \"1\" columns: \"2\" columnWidths: \"60%, 60%\" {\n" +
		"\t\tlayoutcell row: 0 col: 0 { p \"one\" }\n" +
		"\t\tlayoutcell row: 0 col: 1 { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	diagnostics := checkSrc(t, src)
	if !hasMsg(diagnostics, "columnWidths cannot exceed 100%") {
		t.Fatalf("expected the percentage sum diagnostic, got: %s", diagnostics.String())
	}
}

func TestRowLayoutRejectsInvalidSizes(t *testing.T) {
	src := "view main {\n" +
		"\trow columns: \"2\" columnWidths: \"60vw, 40%\" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t\tdiv { p \"two\" }\n" +
		"\t}\n" +
		"}\n"
	diagnostics := checkSrc(t, src)
	if !hasMsg(diagnostics, "columnWidths accepts only sizes in px or %") {
		t.Fatalf("expected the invalid size diagnostic, got: %s", diagnostics.String())
	}
}

func TestRowLayoutColumnCountMustBeAnInteger(t *testing.T) {
	src := "view main {\n" +
		"\trow columns: \"0\" {\n" +
		"\t\tdiv { p \"one\" }\n" +
		"\t}\n" +
		"}\n"
	diagnostics := checkSrc(t, src)
	if !hasMsg(diagnostics, "columns must be an integer between 1 and 24") {
		t.Fatalf("expected the column count diagnostic, got: %s", diagnostics.String())
	}
}
