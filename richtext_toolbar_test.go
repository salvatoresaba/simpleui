package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRichTextToolbarGroupsFromRuntime(t *testing.T) {
	groups := richTextToolbarGroups()
	if len(groups) < 6 {
		t.Fatalf("expected the default toolbar groups, got %d", len(groups))
	}
	ids := map[string]bool{}
	for _, group := range groups {
		if group.Name == "" {
			t.Fatal("toolbar group has no label")
		}
		if len(group.Items) == 0 {
			t.Fatalf("toolbar group %q has no items", group.Name)
		}
		for _, item := range group.Items {
			if item.ID == "" || item.Label == "" {
				t.Fatalf("toolbar item has an empty id or label: %+v", item)
			}
			ids[item.ID] = true
		}
	}
	for _, id := range []string{"undo", "styleSelect", "bold", "paragraphSettings", "link", "pageSettings", "zoom"} {
		if !ids[id] {
			t.Errorf("default toolbar item %q not exposed to the designer", id)
		}
	}
	for _, id := range []string{"lineHeight", "spaceBefore", "spaceAfter", "imageSettings", "documentSettings"} {
		if !ids[id] {
			t.Errorf("extra toolbar item %q not exposed to the designer", id)
		}
	}
}

func TestRichTextToolbarDesignerInjection(t *testing.T) {
	raw := richTextToolbarJSON()
	var groups []richTextToolbarGroup
	if err := json.Unmarshal([]byte(raw), &groups); err != nil {
		t.Fatalf("toolbar JSON is invalid: %v (%s)", err, raw)
	}
	if len(groups) == 0 {
		t.Fatal("toolbar JSON has no groups")
	}
	if !strings.Contains(guiPage, "window.SUI_RICH_TOOLBAR_GROUPS="+raw+";") {
		t.Fatal("designer page does not inject the toolbar groups")
	}
	if !strings.Contains(suiRichTextRuntime, "function noFormatToolbar(inst)") {
		t.Fatal("runtime does not implement toolbarItems: none")
	}
	if !strings.Contains(suiRichTextRuntime, "if (registered.length && !noToolbar)") {
		t.Fatal("toolbarItems: none still shows plugin toolbar buttons")
	}
	for _, marker := range []string{
		"if (noFormatToolbar(inst)) return;",
		"if (noFormatToolbar(inst)) { bar.hidden = true; return; }",
	} {
		if !strings.Contains(suiRichTextRuntime, marker) {
			t.Errorf("runtime missing none guard: %s", marker)
		}
	}
}

func TestRichTextThemeResolvedBeforeToolbarBuild(t *testing.T) {
	runtime := suiRichTextRuntime
	theme := strings.Index(runtime, "var explicitTheme =")
	toolbar := strings.Index(runtime, "buildToolbar(inst); buildMenu(inst)")
	if theme < 0 || toolbar < 0 || theme > toolbar {
		t.Fatal("initial theme must be resolved before buildToolbar overwrites the theme attribute")
	}
}

func richtextTestElement(t *testing.T, src string) (*Prog, *Element) {
	t.Helper()
	prog, dl := ParseSource(src, "richtext_toolbar_test.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	var find func([]ViewNode) *Element
	find = func(nodes []ViewNode) *Element {
		for i := range nodes {
			n := &nodes[i]
			if n.kind == "element" && n.elem != nil {
				if n.elem.tag == "richtext" {
					return n.elem
				}
				if found := find(n.elem.children); found != nil {
					return found
				}
			}
			for _, branch := range [][]*ViewNode{n.body, n.then_, n.else_} {
				for j := range branch {
					if found := find([]ViewNode{*branch[j]}); found != nil {
						return found
					}
				}
			}
		}
		return nil
	}
	for i := range prog.views {
		if found := find(prog.views[i].body); found != nil {
			return prog, found
		}
	}
	t.Fatal("richtext element not found")
	return nil, nil
}

func TestRichTextThemeAndDesignerRoundTrip(t *testing.T) {
	prog, elem := richtextTestElement(t, `state doc = ""
view main {
	richtext "Article" name: "editor" value: doc theme: "light" contentTheme: "dark" toolbarItems: "undo,bold|link"
}`)
	js := GenerateJs(prog)
	for _, marker := range []string{
		`data-sui-rich-theme="light" data-sui-rich-theme-explicit="1" data-sui-rich-content="light"`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS missing %s", marker)
		}
	}
	if strings.Contains(js, `data-sui-rich-content="dark"`) {
		t.Error("generated JS still emits a dark document page")
	}
	model := guiElementFrom(elem, map[string]bool{})
	for key, want := range map[string]string{
		"theme":        "light",
		"contentTheme": "dark",
		"toolbarItems": "undo,bold|link",
	} {
		if got := model.obj[key].s; got != want {
			t.Errorf("designer model %s = %q, want %q", key, got, want)
		}
	}

	prog, _ = richtextTestElement(t, `state doc = ""
view main {
	richtext "Article" name: "editor" value: doc contentTheme: "dark"
}`)
	js = GenerateJs(prog)
	if !strings.Contains(js, `data-sui-rich-theme="dark"`) {
		t.Error("default theme was not emitted")
	}
	if strings.Contains(js, `data-sui-rich-theme="dark" data-sui-rich-theme-explicit="1"`) {
		t.Error("theme flag emitted without a theme attribute")
	}
	if strings.Contains(js, `data-sui-rich-content="dark"`) {
		t.Error("contentTheme still changes the generated paper")
	}
}
