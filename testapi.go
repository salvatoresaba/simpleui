package sui

// This file exposes the shared compiler and browser-test helpers to control
// packages under components/. It is intentionally part of the importable sui
// package: component tests use an external test package to avoid an import
// cycle, so test-only helpers defined in root *_test.go files are not visible
// to them.

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// RepoPath resolves a repository-relative path regardless of the test package's
// working directory.
func RepoPath(parts ...string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(parts...)
	}
	return filepath.Join(append([]string{filepath.Dir(file)}, parts...)...)
}

// GuiHTML is the designer page used by designer and component tests.
var GuiHTML = guiHTML

// GUIComponentLibraryJS is the injected external-library loader served by the
// designer test mux.
const GUIComponentLibraryJS = guiComponentLibraryJS

// GuiModelJSON returns the designer model as a JSON value.
func GuiModelJSON(prog *Prog) JVal { return guiModelJson(prog) }

// GUIBuildFiles compiles SUI source to index, CSS and JS for tests.
func GUIBuildFiles(src string) (string, string, string, JVal, bool) { return guiBuildFiles(src) }

// RenderSui renders SUI source to preview HTML for tests.
func RenderSui(src string) (bool, string, JVal) { return renderSui(src) }

// LoadAppearanceLibraryPayload loads the appearance catalog for tests.
func LoadAppearanceLibraryPayload(projectDir string) []byte {
	return loadAppearanceLibraryPayload(projectDir)
}

// WriteRawHandler writes a raw HTTP response for test servers.
func WriteRawHandler(w http.ResponseWriter, ct, body string) { writeRawHandler(w, ct, body) }

// FeatureScriptHead and FeatureScriptTail wrap browser feature assertions.
const FeatureScriptHead = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function button(label) {
    var all = Array.prototype.slice.call(document.querySelectorAll('button'));
    for (var i = 0; i < all.length; i++) if (all[i].textContent === label) return all[i];
    return null;
  }
  try {
`

const FeatureScriptTail = `    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`

// FindHeadlessBrowser locates an installed Chrome/Edge for headless tests.
func FindHeadlessBrowser() string {
	candidates := []string{
		os.Getenv("CHROME_PATH"),
		os.Getenv("EDGE_PATH"),
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`/usr/bin/google-chrome`,
		`/usr/bin/chromium`,
		`/usr/bin/chromium-browser`,
		`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
	}
	for _, path := range candidates {
		if path != "" {
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				return path
			}
		}
	}
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// GUITestMux serves the designer endpoints used by browser tests.
func GUITestMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /__gui__/components.js", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/javascript; charset=utf-8", guiComponentLibraryJS)
	})
	mux.HandleFunc("GET /__gui__/appearance", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", string(loadAppearanceLibraryPayload("")))
	})
	mux.HandleFunc("GET /__gui__/ping", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	mux.HandleFunc("POST /__gui__/check", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"diagDetails":[]}`)
	})
	mux.HandleFunc("POST /__gui__/component-preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"html":"<!doctype html><html><body><button>Preview</button></body></html>"}`)
	})
	mux.HandleFunc("POST /__gui__/source-preview", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ok, html, msgs := renderSui(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"html":`+StringifyJson(JStr(html))+`}`)
	})
	return mux
}

// InjectGuiTestScript inserts a browser script into the designer page.
func InjectGuiTestScript(script string) string {
	return strings.Replace(guiHTML, "<body>", "<body>"+script, 1)
}

// SuiCssBase, GuiPage and GuiPreviewPicker are read-only assets used by
// component tests to assert generated styles and designer markup.
const SuiCssBase = suiCssBase

var GuiPage = guiPage

const GuiPreviewPicker = guiPreviewPicker

// Views exposes the parsed view list to external tests.
func (p *Prog) Views() []ViewDecl { return p.views }

// Body exposes a view's root nodes to external tests.
func (v *ViewDecl) Body() []ViewNode { return v.body }

// FindElementByTag walks parsed views and returns the first element with the
// requested tag, including nested conditional branches.
func FindElementByTag(p *Prog, tag string) *Element {
	var find func([]ViewNode) *Element
	find = func(nodes []ViewNode) *Element {
		for i := range nodes {
			n := &nodes[i]
			if n.kind == "element" && n.elem != nil {
				if n.elem.tag == tag {
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
	for i := range p.views {
		if found := find(p.views[i].body); found != nil {
			return found
		}
	}
	return nil
}

// DesignerElementFrom builds the designer model for one AST element.
func DesignerElementFrom(elem *Element, seen map[string]bool) JVal {
	return guiElementFrom(elem, seen)
}

// ViewNodesSource serializes designer nodes for model assertions.
func ViewNodesSource(nodes []ViewNode, ind int) string { return viewNodesSource(nodes, ind) }

// GuiCssSetBatch applies a CSS declaration batch through the designer API.
func GuiCssSetBatch(css, class, declsJson string) JVal {
	return guiCssSetBatch(css, class, declsJson)
}

// GuiCssSetSelectors applies a CSS declaration batch to exact selectors.
func GuiCssSetSelectors(css, selectorsJson, atRule, declsJson string) JVal {
	return guiCssSetSelectors(css, selectorsJson, atRule, declsJson)
}

// GuiOpen builds the designer model for SUI source.
func GuiOpen(src string) (JVal, JVal, bool) { return guiOpen(src) }

// AsBool, AsStr and Get are JSON helpers shared by component tests.
func AsBool(v JVal) bool {
	if v.kind == "bool" {
		return v.b
	}
	return false
}

func AsStr(v JVal) string { return asStr(v) }

func Get(o JVal, key string) JVal { return get(o, key) }

// GenerateJsFromInline parses, checks and generates JS from a SUI string.
func GenerateJsFromInline(t *testing.T, src string) string {
	t.Helper()
	prog, pdl := ParseSource(src, "feature_inline.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse inline: %v", pdl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check inline: %v", dl.String())
	}
	return GenerateJs(prog)
}

// RunHeadlessFeature renders source, injects a feature script and runs it in a
// headless browser. The script must call result(true) or result(false, msg).
func RunHeadlessFeature(t *testing.T, source, script string) {
	t.Helper()
	browser := FindHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found")
	}
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + script + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := GUITestMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	go func() { _ = http.Serve(ln, mux) }()
	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=900,700",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=90000", "--dump-dom", url,
	}
	out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
	text := string(out)
	if err != nil {
		t.Fatalf("headless browser failed: %v\n%s", err, text)
	}
	if strings.Contains(text, `id="e2e-result">OK<`) {
		return
	}
	marker := strings.Index(text, `id="e2e-result"`)
	detail := text
	if marker >= 0 {
		end := marker + 1800
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("feature e2e failed:\n%s", detail)
}
