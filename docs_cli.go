package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// docs_cli.go - canonical documentation access for humans and agents.
//
// The Markdown/JSON files are embedded into the compiler so `simpleui docs`
// and the MCP resources work from any directory, not only from the repository.
// docs_nav_test.go verifies that the embedded copy is in sync with the files on
// disk and that links/anchors/manifest entries are valid.

//go:embed docs/*.md docs/*.json AGENTS.md CLAUDE.md llms.txt README.md SECURITY.md SPEC.md DESIGN.md TODO.md examples/README.md examples/security/README.md examples/spreadsheet_table/README.md examples/gantt_project/README.md examples/gantt_construction/README.md examples/process_monitor/README.md examples/graph/README.md examples/motion_sequence/README.md examples/motion_features/README.md .github/copilot-instructions.md .cursor/rules/simpleui.mdc
var embeddedDocs embed.FS

type DocEntry struct {
	ID       string   `json:"id"`
	Path     string   `json:"path"`
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Audience []string `json:"audience"`
	Priority int      `json:"priority"`
	Tags     []string `json:"tags"`
	Sections []string `json:"sections"`
}

type DocInvariant struct {
	ID   string   `json:"id"`
	Rule string   `json:"rule"`
	Docs []string `json:"docs"`
}

type DocRecipe struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	When     string   `json:"when"`
	Target   string   `json:"target"`
	Docs     []string `json:"docs"`
	Examples []string `json:"examples"`
	Steps    []string `json:"steps"`
}

type DocsManifest struct {
	Version     int               `json:"version"`
	Project     string            `json:"project"`
	Description string            `json:"description"`
	Entrypoints map[string]string `json:"entrypoints"`
	Commands    map[string]string `json:"commands"`
	Invariants  []DocInvariant    `json:"invariants"`
	Docs        []DocEntry        `json:"docs"`
	Recipes     []DocRecipe       `json:"recipes"`
}

func loadDocsManifest() (*DocsManifest, error) {
	data, err := embeddedDocs.ReadFile("docs/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("embedded manifest: %v", err)
	}
	var m DocsManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("manifest JSON: %v", err)
	}
	return &m, nil
}

func findDoc(m *DocsManifest, key string) *DocEntry {
	k := strings.TrimSpace(key)
	for i := 0; i < len(m.Docs); i++ {
		if m.Docs[i].ID == k || m.Docs[i].Path == k {
			return &m.Docs[i]
		}
	}
	return nil
}

func embeddedDocText(path string) (string, error) {
	data, err := embeddedDocs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("document %q: %v", path, err)
	}
	return string(data), nil
}

// CmdDocs implements `simpleui docs [--...]`. It always reads the embedded
// documentation, so it is usable from arbitrary working directories.
func CmdDocs(args []string) {
	readID := ""
	search := ""
	jsonOut := false
	list := false
	agent := false
	full := false
	index := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "--") {
			UsageError("unexpected argument " + a)
		}
		name, value, hasValue := SplitFlag(a)
		if !hasValue && (name == "read" || name == "search") && i+1 < len(args) {
			value, hasValue = args[i+1], true
			i++
		}
		switch name {
		case "json":
			jsonOut = true
		case "list":
			list = true
		case "agent":
			agent = true
		case "full":
			full = true
		case "index":
			index = true
		case "read":
			if !hasValue {
				UsageError("--read needs a document id or path")
			}
			readID = value
		case "search":
			if !hasValue {
				UsageError("--search needs a query")
			}
			search = value
		case "help", "h":
			printDocsHelp()
			return
		default:
			UsageError("unknown docs flag --" + name)
		}
	}

	m, err := loadDocsManifest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
		os.Exit(1)
	}

	switch {
	case readID != "":
		d := findDoc(m, readID)
		if d == nil {
			fmt.Fprintf(os.Stderr, "%v: unknown document %q (try --list)\n", AppName, readID)
			os.Exit(1)
		}
		text, err := embeddedDocText(d.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
			os.Exit(1)
		}
		fmt.Print(text)
		if !strings.HasSuffix(text, "\n") {
			fmt.Println()
		}
	case search != "":
		fmt.Print(docsSearchText(m, search))
	case jsonOut:
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	case list:
		fmt.Print(docsListText(m))
	case agent:
		fmt.Print(docsAgentText(m))
	case full:
		text, err := docsFullText(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
			os.Exit(1)
		}
		fmt.Print(text)
	case index:
		printDocsIndex()
	default:
		printDocsIndex()
	}
}

func printDocsHelp() {
	fmt.Printf("usage: %v docs [--index|--json|--list|--agent|--full|--read <id>|--search <q>]\n", AppName)
	fmt.Println("  --index       print docs/INDEX.md (default)")
	fmt.Println("  --json        print docs/manifest.json")
	fmt.Println("  --list        list id, path and title")
	fmt.Println("  --agent       compact context bundle for a coding agent")
	fmt.Println("  --full        print every embedded document")
	fmt.Println("  --read <id>   print one document by id or path")
	fmt.Println("  --search <q>  search titles, tags, summaries and recipes")
}

func printDocsIndex() {
	text, err := embeddedDocText("docs/INDEX.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", AppName, err)
		os.Exit(1)
	}
	fmt.Print(text)
	if !strings.HasSuffix(text, "\n") {
		fmt.Println()
	}
}

func docsListText(m *DocsManifest) string {
	var b strings.Builder
	for i := 0; i < len(m.Docs); i++ {
		d := &m.Docs[i]
		fmt.Fprintf(&b, "%-20s %-38s %s\n", d.ID, d.Path, d.Title)
	}
	return b.String()
}

func docsSearchText(m *DocsManifest, q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return "empty search query\n"
	}
	hit := func(parts ...string) bool {
		for _, s := range parts {
			if strings.Contains(strings.ToLower(s), q) {
				return true
			}
		}
		return false
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Documentation matches for %q\n\n", q)
	found := false
	for i := 0; i < len(m.Docs); i++ {
		d := &m.Docs[i]
		meta := hit(d.ID, d.Path, d.Title, d.Summary, strings.Join(d.Tags, " "), strings.Join(d.Sections, " "))
		text, _ := embeddedDocText(d.Path)
		bodyMatch := text != "" && strings.Contains(strings.ToLower(text), q)
		if meta || bodyMatch {
			fmt.Fprintf(&b, "- docs:%-12s %s - %s\n", d.ID, d.Path, d.Title)
			if bodyMatch {
				lines := strings.Split(text, "\n")
				shown := 0
				for _, line := range lines {
					if !strings.Contains(strings.ToLower(line), q) {
						continue
					}
					snip := strings.TrimSpace(line)
					if len(snip) > 180 {
						snip = snip[:180] + "..."
					}
					fmt.Fprintf(&b, "  > %s\n", snip)
					shown++
					if shown >= 3 {
						break
					}
				}
			}
			found = true
		}
	}
	for i := 0; i < len(m.Recipes); i++ {
		r := &m.Recipes[i]
		if hit(r.ID, r.Title, r.When, r.Target, strings.Join(r.Docs, " "), strings.Join(r.Examples, " "), strings.Join(r.Steps, " ")) {
			fmt.Fprintf(&b, "- recipe:%-10s %s - %s\n", r.ID, r.Title, r.When)
			found = true
		}
	}
	if !found {
		b.WriteString("No matches. Try `simpleui docs --list` or a broader term.\n")
	}
	return b.String()
}

func docsAgentText(m *DocsManifest) string {
	var b strings.Builder
	b.WriteString("# SimpleUI agent context\n\n")
	b.WriteString(m.Description + "\n\n")
	b.WriteString("## Commands\n\n")
	keys := make([]string, 0, len(m.Commands))
	for k := range m.Commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("- `" + k + "`: `" + m.Commands[k] + "`\n")
	}
	b.WriteString("\n## Non-negotiable rules\n\n")
	for _, inv := range m.Invariants {
		b.WriteString("- **" + inv.ID + "**: " + inv.Rule + "\n")
	}
	b.WriteString("\n## Documents\n\n")
	for _, d := range m.Docs {
		fmt.Fprintf(&b, "- `%s` (%s) - %s - tags: %s\n", d.ID, d.Path, d.Title, strings.Join(d.Tags, ", "))
	}
	b.WriteString("\n## Recipes\n\n")
	for _, r := range m.Recipes {
		fmt.Fprintf(&b, "- `%s` [%s] - %s\n  docs: %s; examples: %s\n", r.ID, r.Target, r.Title,
			strings.Join(r.Docs, ", "), strings.Join(r.Examples, ", "))
	}
	b.WriteString("\n## Next\n\n")
	b.WriteString("1. Open the closest recipe in `docs/AGENT_RECIPES.md`.\n")
	b.WriteString("2. Read the canonical document(s) it references.\n")
	b.WriteString("3. Write the smallest `.sui` implementation with loading/error states.\n")
	b.WriteString("4. Run `simpleui check app.sui --json`; then build or serve.\n")
	b.WriteString("5. Before finishing, apply `SECURITY.md` and the security checklist.\n")
	b.WriteString("\nUse `simpleui docs --read <id>` for the full document.\n")
	return b.String()
}

func docsFullText(m *DocsManifest) (string, error) {
	var b strings.Builder
	for i := 0; i < len(m.Docs); i++ {
		d := &m.Docs[i]
		text, err := embeddedDocText(d.Path)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "\n<!-- doc:%s path:%s -->\n\n# %s\n\n%s\n", d.ID, d.Path, d.Title, text)
	}
	return b.String(), nil
}
