package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// security_test.go checks the generated app's first line of defence against
// untrusted data: interpolations are HTML-escaped, href/src are URL-sanitized
// and static text cannot inject markup.

func TestViewTextIsEscaped(t *testing.T) {
	src := "state evil = \"<img src=x onerror=alert(1)>\"\n" +
		"view main {\n" +
		"\th1 \"{evil}\"\n" +
		"\tp \"value: {evil}\"\n" +
		"\ttable {\n" +
		"\t\tthead {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"A\" head: true\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t\ttbody {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"{evil}\"\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
	prog, dl := ParseSource(src, "sec.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if strings.Contains(js, ">${evil}<") || strings.Contains(js, "<h1>${evil}") {
		t.Error("raw interpolation found in view text")
	}
	for _, want := range []string{
		"${__sui_esc(evil)}",
		"function __sui_esc(v)",
		"function __sui_url(v)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
}

func TestAttributeEscapingAndURLSanitizing(t *testing.T) {
	src := "state evil = \"x\"\n" +
		"state url = \"javascript:alert(1)\"\n" +
		"view main {\n" +
		"\ta \"go\" href: url\n" +
		"\timg src: url\n" +
		"\tinput value: evil\n" +
		"\tdiv \"d\" class: evil style: cssSafe(evil)\n" +
		"\tdiv \"l\" link: url\n" +
		"}\n"
	prog, dl := ParseSource(src, "secattr.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"href=\"${__sui_esc(__sui_url(url))}\"",
		"src=\"${__sui_esc(__sui_url(url))}\"",
		"value=\"${__sui_esc(evil)}\"",
		"class=\"${__sui_class(evil)}\"",
		"style=\"${__sui_esc(__sui_css(evil))}\"",
		"function __sui_css(v)",
		"function __sui_class(v)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	if strings.Contains(js, "href=\"${url}\"") || strings.Contains(js, "src=\"${url}\"") {
		t.Error("URL attribute interpolated without __sui_url")
	}
}

func TestStaticTextEscaped(t *testing.T) {
	src := "view main {\n\tp \"1 < 2 & 3 > 0\"\n}\n"
	prog, dl := ParseSource(src, "secstatic.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "1 &lt; 2 &amp; 3 &gt; 0") {
		t.Errorf("static text is not HTML-escaped: %s", js)
	}
}

// TestSecurityExamples keeps every documented example compilable.
func TestSecurityExamples(t *testing.T) {
	files, err := filepath.Glob("examples/security/*.sui")
	if err != nil || len(files) == 0 {
		t.Fatalf("no security examples found: %v", err)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		prog, dl := ParseSource(string(raw), f)
		if dl.HasErrors() {
			t.Errorf("%s: parse: %v", f, dl.String())
			continue
		}
		if dl := Check(prog); dl.HasErrors() {
			t.Errorf("%s: check: %v", f, dl.String())
		}
	}
}

func TestDynamicStyleRequiresCssSafe(t *testing.T) {
	src := "state c = \"red\"\n" +
		"view main {\n" +
		"\tdiv \"x\" style: c\n" +
		"}\n"
	prog, dl := ParseSource(src, "style.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	dl = Check(prog)
	if !dl.HasErrors() || !strings.Contains(dl.String(), "cssSafe") {
		t.Fatalf("expected a cssSafe diagnostic, got: %v", dl.String())
	}
}

func TestClassAllowList(t *testing.T) {
	src := "css \".active { color: red; }\"\n" +
		"state cls = \"active evil\"\n" +
		"view main {\n" +
		"\tdiv \"a\" class: \"btn {cls}\"\n" +
		"\tdiv \"b\" class: cls\n" +
		"\tdiv \"c\" class: \"btn active\"\n" +
		"}\n"
	prog, dl := ParseSource(src, "class.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"class=\"${__sui_class(`btn ${cls}`)}\"",
		"class=\"${__sui_class(cls)}\"",
		"__SUI_CLASSES[\"active\"] = 1;",
		"__SUI_CLASSES[\"btn\"] = 1;",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q", want)
		}
	}
	if strings.Contains(js, "__SUI_CLASSES[\"evil\"]") {
		t.Error("undeclared class leaked into the allow-list")
	}
}

func TestSSRFCompileChecks(t *testing.T) {
	dynamic := "import { http } from \"std\"\n" +
		"state u = \"https://x\"\n" +
		"async fn f() -> null {\n" +
		"\tb, e = http.get(u)\n" +
		"}\n" +
		"view main {\n\tp \"x\"\n}\n"
	prog, dl := ParseSource(dynamic, "ssrf.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); !dl.HasErrors() || !strings.Contains(dl.String(), "getSafe") {
		t.Fatalf("expected a getSafe diagnostic, got: %v", dl.String())
	}

	safe := "import { http } from \"std\"\n" +
		"state u = \"https://x\"\n" +
		"async fn f() -> null {\n" +
		"\tb, e = http.getSafe(u, \"https://x\")\n" +
		"}\n" +
		"view main {\n\tp \"x\"\n}\n"
	prog, dl = ParseSource(safe, "ssrf.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("safe call rejected: %v", dl.String())
	}

	dynamicAllow := "import { http } from \"std\"\n" +
		"state u = \"https://x\"\n" +
		"state allow = \"https://x\"\n" +
		"async fn f() -> null {\n" +
		"\tb, e = http.getSafe(u, allow)\n" +
		"}\n" +
		"view main {\n\tp \"x\"\n}\n"
	prog, dl = ParseSource(dynamicAllow, "ssrf.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); !dl.HasErrors() || !strings.Contains(dl.String(), "allow-list must be a literal") {
		t.Fatalf("expected a literal allow-list diagnostic, got: %v", dl.String())
	}
}

func TestStaticActiveURLRejected(t *testing.T) {
	bad := "view main {\n\ta \"x\" href: \"javascript:alert(1)\"\n}\n"
	prog, dl := ParseSource(bad, "badurl.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); !dl.HasErrors() || !strings.Contains(dl.String(), "active URL scheme") {
		t.Fatalf("expected an active URL diagnostic, got: %v", dl.String())
	}

	good := "view main {\n\ta \"x\" href: \"https://example.com/\"\n}\n"
	prog, dl = ParseSource(good, "goodurl.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("safe URL rejected: %v", dl.String())
	}
}
