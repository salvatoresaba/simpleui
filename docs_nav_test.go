package main

import (
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
)

// docs_nav_test.go - keeps the human/agent documentation navigable:
// manifest paths and sections, internal links/anchors, recipe references and
// the embedded copy must all agree with the files in the repository.

var mdLinkRE = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func TestDocsManifestValid(t *testing.T) {
	m, err := loadDocsManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != 1 || m.Project == "" || len(m.Docs) == 0 || len(m.Recipes) == 0 {
		t.Fatalf("incomplete manifest: %#v", m)
	}
	ids := map[string]bool{}
	paths := map[string]bool{}
	for i := 0; i < len(m.Docs); i++ {
		d := &m.Docs[i]
		if d.ID == "" || d.Path == "" || d.Title == "" || d.Summary == "" {
			t.Errorf("incomplete doc entry: %#v", d)
		}
		if ids[d.ID] {
			t.Errorf("duplicate doc id %q", d.ID)
		}
		if paths[d.Path] {
			t.Errorf("duplicate doc path %q", d.Path)
		}
		ids[d.ID] = true
		paths[d.Path] = true
		if _, err := embeddedDocText(d.Path); err != nil {
			t.Errorf("manifest doc not embedded: %v", err)
		}
		if _, err := os.Stat(d.Path); err != nil {
			t.Errorf("manifest doc missing on disk: %v", err)
		}
		headings := markdownHeadings(readDocForTest(t, d.Path))
		for _, sec := range d.Sections {
			if !headings[sec] {
				t.Errorf("%s: section %q not found as a heading", d.Path, sec)
			}
		}
	}
	for name, p := range m.Entrypoints {
		if _, err := embeddedDocText(p); err != nil {
			t.Errorf("entrypoint %s (%s): %v", name, p, err)
		}
	}
	for i := 0; i < len(m.Invariants); i++ {
		inv := &m.Invariants[i]
		if inv.ID == "" || inv.Rule == "" {
			t.Errorf("incomplete invariant: %#v", inv)
		}
		for _, id := range inv.Docs {
			if !ids[id] {
				t.Errorf("invariant %q references unknown doc %q", inv.ID, id)
			}
		}
	}
	for i := 0; i < len(m.Recipes); i++ {
		r := &m.Recipes[i]
		if r.ID == "" || r.Title == "" || r.When == "" {
			t.Errorf("incomplete recipe: %#v", r)
		}
		if len(r.Docs) == 0 {
			t.Errorf("recipe %q has no docs", r.ID)
		}
		for _, id := range r.Docs {
			if !ids[id] {
				t.Errorf("recipe %q references unknown doc %q", r.ID, id)
			}
		}
		for _, ex := range r.Examples {
			if _, err := os.Stat(ex); err != nil {
				t.Errorf("recipe %q example missing: %v", r.ID, err)
			}
		}
	}
}

func TestDocsManifestCoversDocsDir(t *testing.T) {
	m, err := loadDocsManifest()
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, d := range m.Docs {
		known[d.Path] = true
	}
	entries, err := embeddedDocs.ReadDir("docs")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		p := "docs/" + e.Name()
		if !known[p] {
			t.Errorf("documentation file %s is not listed in docs/manifest.json", p)
		}
	}
}

func TestCanonicalDocsAreEnglish(t *testing.T) {
	m, err := loadDocsManifest()
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		"Impossibile", "documentazione", "Sicurezza", "Configurazione",
		"Rischi residui", "Tabella", "Tabelle", "linguaggio", "compilatore",
		"predefinito", "messaggio di errore", "esempio", "esempi",
	}
	for _, d := range m.Docs {
		if strings.HasSuffix(strings.ToLower(d.Path), "_it.md") ||
			strings.EqualFold(d.Path, "docs/INDEX_IT.md") {
			continue
		}
		text := readDocForTest(t, d.Path)
		for _, word := range forbidden {
			if strings.Contains(text, word) {
				t.Errorf("%s: canonical documentation contains Italian term %q", d.Path, word)
			}
		}
	}
	// Root docs are canonical English even when not listed in the manifest.
	for _, p := range []string{"README.md", "SECURITY.md", "SPEC.md", "DESIGN.md", "TODO.md", "STDLIB.md", "AGENTS.md", "CLAUDE.md", "llms.txt", "examples/README.md", "examples/security/README.md", "examples/spreadsheet_table/README.md", "examples/gantt_project/README.md", "examples/gantt_construction/README.md"} {
		text := readDocForTest(t, p)
		for _, word := range forbidden {
			if strings.Contains(text, word) {
				t.Errorf("%s: canonical documentation contains Italian term %q", p, word)
			}
		}
	}
}

func TestDocsEmbeddedInSync(t *testing.T) {
	m, err := loadDocsManifest()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	check := func(p string) {
		if seen[p] {
			return
		}
		seen[p] = true
		disk, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("read disk %s: %v", p, err)
			return
		}
		emb, err := embeddedDocs.ReadFile(p)
		if err != nil {
			t.Errorf("read embedded %s: %v", p, err)
			return
		}
		if string(disk) != string(emb) {
			t.Errorf("%s: embedded copy is stale; rebuild after editing documents", p)
		}
	}
	for _, d := range m.Docs {
		check(d.Path)
	}
	for _, p := range m.Entrypoints {
		check(p)
	}
	check("docs/manifest.json")
}

func TestDocsLinksAndAnchors(t *testing.T) {
	m, err := loadDocsManifest()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]bool{}
	for _, d := range m.Docs {
		sources[d.Path] = true
	}
	for _, p := range m.Entrypoints {
		sources[p] = true
	}
	// Root docs that are linked but not necessarily manifest entries.
	for _, p := range []string{
		"README.md", "AGENTS.md", "CLAUDE.md", "llms.txt",
		".github/copilot-instructions.md", ".cursor/rules/simpleui.mdc",
	} {
		sources[p] = true
	}
	var files []string
	for p := range sources {
		files = append(files, p)
	}
	sort.Strings(files)
	for _, src := range files {
		text := readDocForTest(t, src)
		for _, line := range strings.Split(stripFencedCode(text), "\n") {
			matches := mdLinkRE.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				target := strings.TrimSpace(m[1])
				if target == "" || strings.HasPrefix(target, "http://") ||
					strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "mailto:") ||
					strings.HasPrefix(target, "tel:") {
					continue
				}
				if i := strings.IndexAny(target, " \t"); i >= 0 {
					target = target[:i]
				}
				target = strings.Trim(target, "<>")
				pathPart, anchor := target, ""
				if i := strings.IndexByte(target, '#'); i >= 0 {
					pathPart, anchor = target[:i], target[i+1:]
				}
				if i := strings.IndexByte(pathPart, '?'); i >= 0 {
					pathPart = pathPart[:i]
				}
				targetPath := src
				if pathPart != "" {
					if strings.HasPrefix(pathPart, "/") {
						targetPath = strings.TrimPrefix(pathPart, "/")
					} else {
						targetPath = path.Join(path.Dir(src), pathPart)
					}
					targetPath = path.Clean(targetPath)
					if !docTargetExists(targetPath) {
						t.Errorf("%s: broken link %q -> %s", src, m[0], targetPath)
						continue
					}
				}
				if anchor == "" || !strings.HasSuffix(strings.ToLower(targetPath), ".md") {
					continue
				}
				headings := markdownAnchors(readDocForTest(t, targetPath))
				if !headings[anchor] {
					t.Errorf("%s: broken anchor %q -> %s#%s", src, m[0], targetPath, anchor)
				}
			}
		}
	}
}

func readDocForTest(t *testing.T, p string) string {
	t.Helper()
	if data, err := embeddedDocs.ReadFile(p); err == nil {
		return string(data)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(data)
}

func docTargetExists(p string) bool {
	if _, err := embeddedDocs.ReadFile(p); err == nil {
		return true
	}
	_, err := os.Stat(p)
	return err == nil
}

func stripFencedCode(s string) string {
	var b strings.Builder
	fenced := false
	for _, line := range strings.Split(s, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			fenced = !fenced
			b.WriteString("\n")
			continue
		}
		if !fenced {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func markdownHeadings(s string) map[string]bool {
	out := map[string]bool{}
	fenced := false
	for _, line := range strings.Split(s, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced || !strings.HasPrefix(trim, "#") {
			continue
		}
		title := strings.TrimSpace(strings.TrimLeft(trim, "#"))
		if title != "" {
			out[title] = true
		}
	}
	return out
}

func markdownAnchors(s string) map[string]bool {
	out := map[string]bool{}
	counts := map[string]int{}
	fenced := false
	for _, line := range strings.Split(s, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced || !strings.HasPrefix(trim, "#") {
			continue
		}
		title := strings.TrimSpace(strings.TrimLeft(trim, "#"))
		if title == "" {
			continue
		}
		slug := githubSlug(title)
		if n := counts[slug]; n > 0 {
			out[slug+"-"+itoa(n)] = true
			counts[slug] = n + 1
		} else {
			out[slug] = true
			counts[slug] = 1
		}
	}
	return out
}

func githubSlug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
