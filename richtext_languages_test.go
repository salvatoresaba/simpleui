package main

import (
	"strings"
	"testing"
)

func patternsObject(js string) string {
	i := strings.Index(js, hyphenPatternsStart)
	if i < 0 {
		return ""
	}
	start := i + len(hyphenPatternsStart)
	e := strings.Index(js[start:], hyphenPatternsEnd)
	if e < 0 {
		return ""
	}
	return js[start : start+e]
}

func TestHyphenationLanguagesFromRuntime(t *testing.T) {
	if len(richHyphenPatterns) != 57 {
		t.Fatalf("expected 57 embedded hyphenation dictionaries, got %d", len(richHyphenPatterns))
	}
	if len(richHyphenOptions) != len(richHyphenPatterns) {
		t.Fatalf("expected one designer option per dictionary (%d), got %d", len(richHyphenPatterns), len(richHyphenOptions))
	}
	options := map[string]hyphenLangOption{}
	for _, o := range richHyphenOptions {
		options[o.Code] = o
		if len(o.Patterns) != 1 || o.Patterns[0] != o.Code || o.Bytes != len(richHyphenPatterns[o.Code]) {
			t.Errorf("designer option %q does not describe its dictionary: %+v", o.Code, o)
		}
	}
	for _, code := range []string{"it", "en", "de", "ru", "el-monoton", "el-polyton", "tr", "en-gb", "sh-latn", "sr-cyrl", "sa"} {
		if _, ok := options[code]; !ok {
			t.Errorf("designer language list is missing %q", code)
		}
	}
	for _, tc := range []struct{ tag, pattern string }{
		{"el", "el-monoton"}, {"de-CH", "de"}, {"sr-Latn", "sh-latn"},
		{"no", "nn"}, {"nb", "nn"}, {"mn", "mn-cyrl"}, {"bs", "sh-latn"},
	} {
		got := hyphenPatternCodes([]string{tc.tag})
		if len(got) != 1 || got[0] != tc.pattern {
			t.Errorf("language %q: got patterns %v, want %q", tc.tag, got, tc.pattern)
		}
	}
	if got := hyphenPatternCodes([]string{"xx"}); len(got) != 0 {
		t.Errorf("unknown language must not select a dictionary, got %v", got)
	}
}

func TestRichTextRuntimeSubset(t *testing.T) {
	small := richTextRuntimeSubset([]string{"it", "en", "el"})
	if len(small) >= len(suiRichTextRuntime)-100000 {
		t.Fatalf("subset runtime is not smaller: full=%d subset=%d", len(suiRichTextRuntime), len(small))
	}
	if !strings.Contains(small, ";  var hyphenationLanguage = '';") {
		t.Error("subset runtime lost the language variable declaration")
	}
	patterns := patternsObject(small)
	for _, want := range []string{`"it":`, `"en":`, `"el-monoton":`} {
		if !strings.Contains(patterns, want) {
			t.Errorf("subset runtime is missing %s", want)
		}
	}
	for _, unwanted := range []string{`"ru":`, `"de":`, `"sr-cyrl":`} {
		if strings.Contains(patterns, unwanted) {
			t.Errorf("subset runtime still embeds %s", unwanted)
		}
	}
	empty := patternsObject(richTextRuntimeSubset([]string{"none"}))
	if empty != "{}" {
		t.Errorf("empty selection should embed no dictionaries, got %.40q", empty)
	}
}

func TestDesignerExposesHyphenationLanguages(t *testing.T) {
	if !strings.Contains(guiPage, "window.SUI_HYPHEN_LANGUAGES=[") {
		t.Error("designer page does not expose the hyphenation language list")
	}
	if !strings.Contains(guiHTML, "Hyphenation languages...") {
		t.Error("richtext inspector is missing the hyphenation overlay button")
	}
	if !strings.Contains(guiHTML, `if (n.hyphenLangs)`) {
		t.Error("designer SUI serializer drops the hyphenLangs attribute")
	}
}

func TestGeneratedJsHyphenLangsAttribute(t *testing.T) {
	cases := []struct {
		name   string
		attr   string
		want   []string
		absent []string
	}{
		{"subset", ` hyphenLangs: "it,en,en-GB,el"`,
			[]string{`"it":`, `"en":`, `"en-gb":`, `"el-monoton":`},
			[]string{`"ru":`, `"de":`, `"sr-cyrl":`}},
		{"none", ` hyphenLangs: "none"`,
			[]string{"{}"}, []string{`"it":`, `"ru":`}},
		{"default keeps all dictionaries", "",
			[]string{`"it":`, `"ru":`, `"sr-cyrl":`}, nil},
		{"wildcard keeps all dictionaries", ` hyphenLangs: "*"`,
			[]string{`"ru":`}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			js := generateJsFromInline(t, `view main { richtext "Article" name: "editor"`+tc.attr+` }`)
			if !strings.Contains(js, richTextRuntimeMarker) {
				t.Fatal("generated app.js is missing the rich text runtime")
			}
			patterns := patternsObject(js)
			for _, want := range tc.want {
				if !strings.Contains(patterns, want) {
					t.Errorf("generated app.js is missing %s in HYPHEN_PATTERNS", want)
				}
			}
			for _, unwanted := range tc.absent {
				if strings.Contains(patterns, unwanted) {
					t.Errorf("generated app.js embeds unselected %s", unwanted)
				}
			}
		})
	}
}

func TestGeneratedJsHyphenLangsUnionAndReactiveFallback(t *testing.T) {
	js := generateJsFromInline(t, `view main {
  richtext "One" name: "one" hyphenLangs: "it"
  richtext "Two" name: "two" hyphenLangs: "ru,el"
}`)
	patterns := patternsObject(js)
	for _, want := range []string{`"it":`, `"ru":`, `"el-monoton":`} {
		if !strings.Contains(patterns, want) {
			t.Errorf("union of editor dictionaries is missing %s", want)
		}
	}
	if strings.Contains(patterns, `"de":`) {
		t.Error("union embedded an unselected dictionary")
	}
	reactive := generateJsFromInline(t, `view main {
  state langs = "it"
  richtext "One" name: "one" hyphenLangs: langs
}`)
	if !strings.Contains(patternsObject(reactive), `"ru":`) {
		t.Error("a reactive hyphenLangs value cannot be resolved at build time and must keep all dictionaries")
	}
}
