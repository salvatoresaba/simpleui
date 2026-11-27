package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// designer_workflow_test.go covers the designer workflow that an author uses
// on a file with conditional and repeated blocks:
//
//   - the CSS analysis must not call the classes of those blocks unused;
//   - a rule repeated inside a media query is a refinement, not a duplicate;
//   - picking an element in the tree must outline it in the preview;
//   - a manual edit of the source must apply (the formatter and the parser
//     must agree on the import list) and its errors must be underlined.

// ---------------------------------------------------------------------------
// Parser and formatter
// ---------------------------------------------------------------------------

func TestParseMultiLineImportList(t *testing.T) {
	src := "import {\n  http, json\n} from \"std\"\n\n" +
		"view main {\n\tp \"hi\"\n}\n"
	prog, dl := ParseSource(src, "multi.sui")
	if dl.HasErrors() {
		t.Fatalf("multi-line import must parse: %s", diagMsgs(dl))
	}
	if len(prog.imports) != 1 {
		t.Fatalf("expected one import, got %d", len(prog.imports))
	}
	imp := prog.imports[0]
	if len(imp.names) != 2 || imp.names[0] != "http" || imp.names[1] != "json" {
		t.Fatalf("unexpected import names: %v", imp.names)
	}
	if imp.kind != "std" {
		t.Fatalf("expected a std import, got %q", imp.kind)
	}
}

// TestFormatSourceRoundTripKeepsImport is the invariant that the designer
// relies on: what the formatter writes must parse again, otherwise every
// manual edit of a file with an import fails to apply.
func TestFormatSourceRoundTripKeepsImport(t *testing.T) {
	src := "import { http, json } from \"std\"\n\n" +
		"state counter = 0\n\n" +
		"view main {\n\tp \"count {counter}\"\n}\n"
	formatted, fdl := FormatSource(src, "round.sui")
	if fdl != nil && fdl.HasErrors() {
		t.Fatalf("format failed: %s", diagMsgs(*fdl))
	}
	if !strings.Contains(formatted, "from \"std\"") {
		t.Fatalf("the formatted source lost the import:\n%s", formatted)
	}
	_, dl := ParseSource(formatted, "round.sui")
	if dl.HasErrors() {
		t.Fatalf("the formatted source does not parse: %s\n%s", diagMsgs(dl), formatted)
	}
}

// ---------------------------------------------------------------------------
// CSS analysis
// ---------------------------------------------------------------------------

func TestAnalyzeCssIgnoresAtRuleRefinements(t *testing.T) {
	src := ".card { padding: 10px; gap: 4px; }\n" +
		".card:last-child { padding-bottom: 0; }\n" +
		".row-item { display: flex; }\n" +
		"@media (max-width: 600px) {\n" +
		"  .card { padding: 4px; }\n" +
		"  .row-item { flex-direction: column; }\n" +
		"}\n"
	used := map[string]bool{"card": true, "row-item": true}
	probs := AnalyzeCss(ParseCss(src), used)
	for _, p := range probs {
		if p.kind == "duplicate" {
			t.Errorf("a media query must not be a duplicate: %s", p.message)
		}
		if p.kind == "override" {
			t.Errorf("a refinement must not be reported as an override: %s", p.message)
		}
	}
}

// TestAnalyzeCssKeepsRealDuplicates makes sure the relaxed rules still catch
// a genuine mistake: the same selector twice in the same context.
func TestAnalyzeCssKeepsRealDuplicates(t *testing.T) {
	src := ".card { padding: 10px; }\n.card { padding: 20px; }\n"
	probs := AnalyzeCss(ParseCss(src), map[string]bool{"card": true})
	dup := false
	for _, p := range probs {
		if p.kind == "duplicate" {
			dup = true
		}
	}
	if !dup {
		t.Errorf("a repeated selector in the same context must be reported: %+v", probs)
	}
}

// ---------------------------------------------------------------------------
// Designer endpoints
// ---------------------------------------------------------------------------

func TestGuiApplySourceReportsDiagnostics(t *testing.T) {
	res := StringifyJson(guiApplySource("view main {\n\tp \"unterminated\n}\n"))
	if !strings.Contains(res, `"ok":false`) {
		t.Fatalf("a broken source must not apply: %s", res)
	}
	if !strings.Contains(res, `"diagDetails":[{`) {
		t.Fatalf("the failure must carry structured diagnostics: %s", res)
	}
}

func TestGuiApplySourceAcceptsFormattedImport(t *testing.T) {
	src := "import { http, json } from \"std\"\n\nview main {\n\tp \"hi\"\n}\n"
	res := StringifyJson(guiApplySource(src))
	if !strings.Contains(res, `"ok":true`) {
		t.Fatalf("a valid source must apply: %s", res)
	}
}

// ---------------------------------------------------------------------------
// End to end (real designer page in headless Chrome)
// ---------------------------------------------------------------------------

const designerWorkflowSource = `import { http } from "std"

state items = ["a", "b"]

view main {
	card "Card" class: "card" {
		text "Hello" class: "label"
		if items != null {
			div class: "inner" {
				p "inside" class: "note"
			}
		}
		button "After" class: "after-btn" {
		}
		list item in items {
			row class: "row-item" {
				p "{item}"
			}
		}
	}
}`

const designerWorkflowCss = `css ".card { padding: 10px; gap: 4px; } .label { font-weight: 700; } .inner { color: #f00; } .note { color: #00f; } .row-item { display: flex; } @media (max-width: 600px) { .card { padding: 4px; } .row-item { flex-direction: column; } }"`

func TestGuiDesignerWorkflowE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping designer end-to-end test")
	}
	source := designerWorkflowSource + "\n" + designerWorkflowCss
	page := injectGuiTestScript(designerWorkflowScript)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}

	ln, err := ListenLocal()
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	logPath := filepath.Join(t.TempDir(), "requests.log")
	var lastHTML string

	mux := http.NewServeMux()

	// The designer page talks to the server with a handful of endpoints. Every
	// data endpoint is registered for both methods and reads its payload from
	// the request body or from the query string, so the harness does not
	// depend on the verb the page uses.
	formValue := func(r *http.Request, key string) string {
		if v := r.PostFormValue(key); v != "" {
			return v
		}
		return r.URL.Query().Get(key)
	}
	on := func(path string, fn func(w http.ResponseWriter, r *http.Request)) {
		mux.HandleFunc("POST "+path, fn)
		mux.HandleFunc("GET "+path, fn)
	}
	json := func(w http.ResponseWriter, body string) {
		writeRawHandler(w, "application/json; charset=utf-8", body)
	}

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	mux.HandleFunc("GET /__gui__/ping", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true}`)
	})
	mux.HandleFunc("GET /__gui__/appearance", func(w http.ResponseWriter, r *http.Request) {
		json(w, string(loadAppearanceLibraryPayload("")))
	})
	mux.HandleFunc("GET /__gui__/components.js", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/javascript; charset=utf-8", guiComponentLibraryJS)
	})
	mux.HandleFunc("GET /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		json(w, loadGuiPrefs())
	})
	mux.HandleFunc("GET /__gui__/templates", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"templates":[]}`)
	})
	mux.HandleFunc("GET /__gui__/fonts", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"fonts":[]}`)
	})
	mux.HandleFunc("GET /__gui__/assets", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiAssetList()))
	})
	mux.HandleFunc("GET /__gui__/history", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"versions":[]}`)
	})
	mux.HandleFunc("GET /__gui__/restore", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"sui":""}`)
	})
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		json(w, `{"ok":true,"sui":`+StringifyJson(JStr(source))+`}`)
	})
	on("/__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		model, msgs, details, ok := guiOpenDiags(formValue(r, "sui"))
		if !ok {
			json(w, `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`,"diagDetails":`+StringifyJson(details)+`}`)
			return
		}
		body := StringifyJson(model)
		json(w, `{"ok":true,`+body[1:])
	})
	on("/__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		ok, html, msgs := renderSui(formValue(r, "sui"))
		if ok {
			lastHTML = html
			json(w, `{"ok":true}`)
			return
		}
		json(w, `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", lastHTML)
	})
	on("/__gui__/apply", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiApplySource(formValue(r, "sui"))))
	})
	on("/__gui__/check", func(w http.ResponseWriter, r *http.Request) {
		_, msgs, details, ok := guiOpenDiags(formValue(r, "sui"))
		okStr := "false"
		if ok {
			okStr = "true"
		}
		json(w, `{"ok":`+okStr+`,"diagnostics":`+StringifyJson(msgs)+`,"diagDetails":`+StringifyJson(details)+`}`)
	})
	on("/__css__/parse", func(w http.ResponseWriter, r *http.Request) {
		json(w, StringifyJson(guiCssParse(formValue(r, "css"), formValue(r, "used"))))
	})

	go func() {
		_ = http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err == nil {
				fmt.Fprintf(f, "%s %s\n", r.Method, r.URL.Path)
				_ = f.Close()
			}
			mux.ServeHTTP(w, r)
		}))
	}()

	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=1200,800",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=60000", "--dump-dom",
		"--enable-logging=stderr", "--log-level=0", url,
	}
	out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
	text := string(out)
	if err != nil {
		t.Fatalf("headless browser failed: %v\n%s", err, text)
	}
	if strings.Contains(text, `data-e2e="done">OK</div>`) {
		return
	}
	detail := text
	if marker := strings.Index(text, `data-e2e="done">`); marker >= 0 {
		end := marker + 400
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	if title := strings.Index(text, "<title>"); title >= 0 {
		if end := strings.Index(text[title:], "</title>"); end > 0 {
			detail = "title: " + text[title+7:title+end] + "\n" + detail
		}
	}
	if stepIdx := strings.Index(text, `id="e2e-step"`); stepIdx >= 0 {
		end := stepIdx + 200
		if end > len(text) {
			end = len(text)
		}
		detail = "step: " + text[stepIdx:end] + "\n" + detail
	}
	if reqLog, err := os.ReadFile(logPath); err == nil {
		detail = "requests:\n" + string(reqLog) + "\n" + detail
	}
	t.Fatalf("designer workflow e2e failed:\n%s", detail)
}

const designerWorkflowScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-result';
      document.body.insertBefore(el, document.body.firstChild);
    }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    el.setAttribute('data-e2e', 'done');
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function step(name) {
    var el = document.getElementById('e2e-step');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-step';
      document.body.insertBefore(el, document.body.firstChild);
    }
    el.textContent = name;
    document.title = 'E2E:' + name;
  }
  var wait = function (ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); };
  try {
    step('start');
    var ta = null;
    for (var i = 0; i < 150; i++) {
      ta = document.getElementById('out');
      if (ta && ta.value && document.querySelectorAll('#nodes .node').length) break;
      await wait(100);
    }
    if (!ta) throw new Error('the source editor is missing');
    await wait(1200);
    step('loaded');

    // 1. the CSS analysis must not report the classes of the if/list blocks
    var unused = (problems || []).filter(function (p) { return p.kind === 'unused'; });
    if (unused.length) {
      throw new Error('false unused classes: ' + unused.map(function (p) { return p.message; }).join(' | '));
    }
    var dups = (problems || []).filter(function (p) { return p.kind === 'duplicate' || p.kind === 'override'; });
    if (dups.length) {
      throw new Error('false duplicate/override problems: ' + dups.map(function (p) { return p.message; }).join(' | '));
    }
    step('css-ok');

    // 2. picking an element in the tree must outline it in the preview.
    //    The picked element comes AFTER a dynamic block: the preview ids of
    //    the block children must be counted, or the tree points at the wrong
    //    preview element.
    var nodes = Array.prototype.slice.call(document.querySelectorAll('#nodes .node'));
    if (!nodes.length) {
      var status0 = (document.getElementById('status') || {}).textContent || '';
      var pages = (typeof root !== 'undefined' && root && root.pages) ? root.pages.length : '?';
      var srcLen = (typeof suiSource !== 'undefined') ? String(suiSource || '').length : '?';
      throw new Error('the element tree is empty (status: ' + status0 + ', pages: ' + pages + ', source length: ' + srcLen + ')');
    }
    var target = null;
    for (var j = 0; j < nodes.length; j++) {
      if ((nodes[j].textContent || '').indexOf('after-btn') >= 0) { target = nodes[j]; break; }
    }
    if (!target) throw new Error('the element after the dynamic block is not in the tree');
    target.click();
    await wait(700);
    var frame = document.getElementById('prev');
    var doc = frame && frame.contentDocument;
    var outlined = doc ? doc.querySelectorAll('.sui-hl') : [];
    if (!outlined.length) throw new Error('the picked element is not highlighted in the preview');
    var outlinedClass = outlined[0].getAttribute('class') || '';
    if (outlinedClass.indexOf('after-btn') < 0) {
      throw new Error('the preview highlights the wrong element: ' + outlinedClass);
    }
    step('hl-ok');

    // the inspector on the right must show the picked element
    var inspector = document.getElementById('inspector');
    var inspectorText = inspector ? (inspector.textContent || '').replace(/\s+/g, ' ') : '';
    if (inspectorText.indexOf('after-btn') < 0) {
      throw new Error('the inspector does not show the picked element: ' + inspectorText.slice(0, 160));
    }
    step('inspector-ok');

    // 3. picking an element in the preview must select it in the tree and
    //    show its inspector. An element inside a dynamic block has a row of
    //    its own: the element itself is selected, not the block container.
    //    The deepest match is the text node inside the if block: its parent
    //    div holds the same text.
    var picks = Array.prototype.slice.call(doc.querySelectorAll('[data-sui-id]')).filter(function (e) {
      return (e.textContent || '').indexOf('inside') >= 0;
    });
    var pick = picks.length ? picks[picks.length - 1] : null;
    if (!pick) throw new Error('the preview has no pickable element');
    pick.click();
    await wait(700);
    var selected = document.querySelectorAll('#nodes .node.sel');
    if (!selected.length) {
      throw new Error('the element picked in the preview is not selected in the tree');
    }
    var selectedText = (selected[0].textContent || '').replace(/\s+/g, ' ');
    if (selectedText.indexOf('inside') < 0) {
      throw new Error('the tree selects the wrong element: ' + selectedText.slice(0, 160));
    }
    var inspector2 = document.getElementById('inspector');
    var inspectorText2 = inspector2 ? (inspector2.textContent || '').replace(/\s+/g, ' ') : '';
    if (inspectorText2.indexOf('note') < 0) {
      throw new Error('the inspector does not show the element picked in the preview: ' + inspectorText2.slice(0, 160));
    }
    step('reverse-ok');

    // 4. a manual edit must apply: the formatted source (with the import on
    //    several lines) must parse again
    ta.value = ta.value.replace('text "Hello"', 'text "Hello there"');
    ta.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(2500);
    var status = (document.getElementById('status') || {}).textContent || '';
    if (status.indexOf('expected') >= 0 || status.indexOf('invalid') >= 0) {
      throw new Error('the manual edit did not apply: ' + status);
    }
    var applied = document.getElementById('out').value;
    if (applied.indexOf('Hello there') < 0) throw new Error('the manual edit was reverted: ' + applied.slice(0, 200));
    step('applied');

    // 5. a broken edit must be underlined in the file, not only in the status bar
    ta.value = ta.value.replace('text "Hello there"', 'text "Hello there');
    ta.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(2500);
    var marks = document.querySelectorAll('#sui_hl .cm-error').length;
    if (!marks) throw new Error('the errors of the file are not underlined');
    step('errors-ok');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`
