package richtext

import (
	"encoding/json"
	"regexp"
	"strings"
)

// The toolbar picker inside the designer must expose the same commands as the
// rich text runtime. The default groups and the command labels are parsed from
// the embedded runtime instead of being duplicated by hand; a small table adds
// the select-like controls that have no TOOL_LABELS entry.

type ToolbarItem struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Default bool   `json:"default"`
}

type ToolbarGroup struct {
	Name  string        `json:"name"`
	Items []ToolbarItem `json:"items"`
}

var ToolbarExtraLabels = map[string]string{
	"styleSelect": "Block style",
	"fontFamily":  "Font family",
	"fontSize":    "Font size",
	"lineHeight":  "Line height",
	"spaceBefore": "Space before",
	"spaceAfter":  "Space after",
	"zoom":        "Zoom controls",
	"sourceMode":  "Source code",
}

var ToolbarExtraGroups = map[string]string{
	"lineHeight":       "Paragraph",
	"spaceBefore":      "Paragraph",
	"spaceAfter":       "Paragraph",
	"imageSettings":    "Insert",
	"documentSettings": "Document",
	"sourceMode":       "Document",
}

var (
	richTextHistoryTools = map[string]bool{
		"undo": true, "redo": true, "copy": true, "cut": true, "paste": true,
	}
	richTextFormatTools = map[string]bool{
		"styleSelect": true, "fontFamily": true, "fontSize": true, "fontSizeUp": true,
		"fontSizeDown": true, "bold": true, "italic": true, "underline": true,
		"strike": true, "textColor": true, "highlight": true, "superscript": true,
		"subscript": true, "code": true, "clearFormat": true, "copyFormat": true,
	}
	richTextParagraphTools = map[string]bool{
		"indent": true, "outdent": true, "paragraphSettings": true, "bulletList": true,
		"orderedList": true, "blockquote": true, "codeBlock": true,
	}
	richTextInsertTools = map[string]bool{
		"link": true, "unlink": true, "image": true, "imageSettings": true,
		"table": true, "symbol": true, "hr": true, "pageBreak": true,
	}
	richTextDocumentTools = map[string]bool{
		"documentSettings": true, "pageSettings": true, "print": true,
		"printPreview": true, "find": true, "sourceMode": true,
	}
)

func DefaultToolbar() string {
	const marker = "var defaultToolbar = '"
	i := strings.Index(Runtime, marker)
	if i < 0 {
		return ""
	}
	rest := Runtime[i+len(marker):]
	end := strings.Index(rest, "'")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func ToolLabels() map[string]string {
	labels := map[string]string{}
	const marker = "var TOOL_LABELS = {"
	i := strings.Index(Runtime, marker)
	if i < 0 {
		return labels
	}
	chunk := Runtime[i:]
	if end := strings.Index(chunk, "\n  };"); end >= 0 {
		chunk = chunk[:end]
	}
	re := regexp.MustCompile(`\b([A-Za-z][A-Za-z0-9]*): '([^']*)'`)
	for _, m := range re.FindAllStringSubmatch(chunk, -1) {
		labels[m[1]] = m[2]
	}
	for id, label := range ToolbarExtraLabels {
		if labels[id] == "" {
			labels[id] = label
		}
	}
	return labels
}

func ToolbarGroupName(tokens []string) string {
	has := func(set map[string]bool) bool {
		for _, token := range tokens {
			if set[token] {
				return true
			}
		}
		return false
	}
	if has(richTextHistoryTools) {
		return "History and clipboard"
	}
	if has(richTextFormatTools) {
		return "Text formatting"
	}
	for _, token := range tokens {
		if strings.HasPrefix(token, "align") || token == "justify" {
			return "Paragraph"
		}
	}
	if has(richTextParagraphTools) {
		return "Paragraph"
	}
	if has(richTextInsertTools) {
		return "Insert"
	}
	if has(richTextDocumentTools) {
		return "Document"
	}
	for _, token := range tokens {
		if token == "zoom" {
			return "Zoom"
		}
	}
	return "Tools"
}

// ToolbarGroups returns the default toolbar plus the extra controls the
// runtime accepts but does not show by default. Every entry carries its default
// flag so the designer can send an empty toolbarItems attribute when the user
// selects exactly the default set.
func ToolbarGroups() []ToolbarGroup {
	labels := ToolLabels()
	groups := []ToolbarGroup{}
	seen := map[string]bool{}
	for _, raw := range strings.Split(DefaultToolbar(), "|") {
		tokens := []string{}
		items := []ToolbarItem{}
		for _, token := range strings.Split(raw, ",") {
			id := strings.TrimSpace(token)
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			tokens = append(tokens, id)
			label := labels[id]
			if label == "" {
				label = id
			}
			items = append(items, ToolbarItem{ID: id, Label: label, Default: true})
		}
		if len(items) > 0 {
			groups = append(groups, ToolbarGroup{Name: ToolbarGroupName(tokens), Items: items})
		}
	}
	for _, id := range []string{"lineHeight", "spaceBefore", "spaceAfter", "imageSettings", "documentSettings", "sourceMode"} {
		if seen[id] {
			continue
		}
		seen[id] = true
		label := labels[id]
		if label == "" {
			label = id
		}
		item := ToolbarItem{ID: id, Label: label, Default: false}
		name := ToolbarExtraGroups[id]
		added := false
		for i := range groups {
			if groups[i].Name == name {
				groups[i].Items = append(groups[i].Items, item)
				added = true
				break
			}
		}
		if !added {
			groups = append(groups, ToolbarGroup{Name: name, Items: []ToolbarItem{item}})
		}
	}
	return groups
}

// ToolbarJSON feeds the designer's multi-select toolbar picker.
func ToolbarJSON() string {
	data, err := json.Marshal(ToolbarGroups())
	if err != nil {
		return "[]"
	}
	return string(data)
}
