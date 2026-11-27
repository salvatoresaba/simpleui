package core

import (
	"strings"
	"testing"
)

// The designer preview can force a branch of a conditional block. Both
// branches must show up in the preview and the ids of the elements must not
// change, or the tree would point at the wrong preview element.
func TestGuiPreviewForcesIfBranches(t *testing.T) {
	src := "state flag = true\nview main {\n\tif flag == true {\n\t\tp \"then\"\n\t} else {\n\t\tp \"else\"\n\t}\n}\n"
	_, plain, _ := renderSui(src)
	ok, both, msgs := renderSuiModes(src, map[int]string{3: "both"})
	if !ok {
		t.Fatalf("preview failed: %v", StringifyJson(msgs))
	}
	if !strings.Contains(both, ">then<") || !strings.Contains(both, ">else<") {
		t.Errorf("the preview must show both branches:\n%s", both)
	}
	if strings.Contains(both, "(flag === true) ?") {
		t.Errorf("the preview must not evaluate the condition:\n%s", both)
	}
	if !strings.Contains(plain, "(flag === true) ?") {
		t.Errorf("the normal preview must evaluate the condition:\n%s", plain)
	}
	if strings.Count(both, "data-sui-id") != strings.Count(plain, "data-sui-id") {
		t.Errorf("the ids of the elements must not change: %d vs %d",
			strings.Count(both, "data-sui-id"), strings.Count(plain, "data-sui-id"))
	}
}

// A single branch can be forced too, to check the effect of the other one.
func TestGuiPreviewForcesOneIfBranch(t *testing.T) {
	src := "state flag = true\nview main {\n\tif flag == true {\n\t\tp \"then\"\n\t} else {\n\t\tp \"else\"\n\t}\n}\n"
	ok, only, msgs := renderSuiModes(src, map[int]string{3: "then"})
	if !ok {
		t.Fatalf("preview failed: %v", StringifyJson(msgs))
	}
	if !strings.Contains(only, ">then<") {
		t.Errorf("the preview must show the then branch:\n%s", only)
	}
	if strings.Contains(only, ">else<") {
		t.Errorf("the preview must not show the else branch:\n%s", only)
	}
}

// The branch overrides travel with the preview request as "line:mode" pairs.
func TestParseIfPreviewModes(t *testing.T) {
	modes := parseIfPreviewModes("12:both, 30:then ,40:else")
	if len(modes) != 3 {
		t.Fatalf("expected 3 modes, got %d: %v", len(modes), modes)
	}
	if modes[12] != "both" || modes[30] != "then" || modes[40] != "else" {
		t.Errorf("wrong modes: %v", modes)
	}
	if got := parseIfPreviewModes(""); len(got) != 0 {
		t.Errorf("an empty request must not force any branch: %v", got)
	}
	if got := parseIfPreviewModes("12:maybe,xx:both"); len(got) != 0 {
		t.Errorf("an invalid request must not force any branch: %v", got)
	}
}

// The line of the block is the one the designer shows: the preview must find
// the right block among the nested ones.
func TestFindIfNodeByLine(t *testing.T) {
	src := "view main {\n\tdiv {\n\t\tif a == 1 {\n\t\t\tp \"one\"\n\t\t}\n\t\tif b == 2 {\n\t\t\tp \"two\"\n\t\t}\n\t}\n}\n"
	prog, dl := ParseSource(src, "test.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	node := findIfNodeByLine(prog, 6)
	if node == nil {
		t.Fatal("the block at line 6 was not found")
	}
	if node.cond == nil || !strings.Contains(exprSource(node.cond), "b == 2") {
		t.Errorf("the wrong block was found: %+v", node.cond)
	}
	if findIfNodeByLine(prog, 99) != nil {
		t.Error("a line without a block must not match")
	}
}
