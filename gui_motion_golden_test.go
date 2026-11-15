package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// gui_motion_golden_test.go captures a stable structural snapshot of the two
// motion overlays: the exit editor and the trigger overlay with the guard
// field. It uses text, field order, option lists and visibility flags rather
// than pixels, so it stays deterministic across fonts and platforms while
// still catching designer regressions. Set SUI_UPDATE_GUI_MOTION_GOLDEN=1 to
// regenerate testdata/gui_motion_golden/panels.json.
type guiMotionGolden struct {
	PanelOn bool `json:"panelOn"`
	Exit    struct {
		Title                 string   `json:"title"`
		StateValues           []string `json:"stateValues"`
		StateLabels           []string `json:"stateLabels"`
		Labels                []string `json:"labels"`
		Buttons               []string `json:"buttons"`
		SpringInitiallyHidden bool     `json:"springInitiallyHidden"`
		SpringVisible         bool     `json:"springVisible"`
	} `json:"exit"`
	Guard struct {
		Title            string   `json:"title"`
		HasGuard         bool     `json:"hasGuard"`
		GuardType        string   `json:"guardType"`
		GuardPlaceholder string   `json:"guardPlaceholder"`
		GuardHint        string   `json:"guardHint"`
		Buttons          []string `json:"buttons"`
	} `json:"guard"`
}

func TestGuiMotionPanelsGolden(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion panel golden test")
	}

	page := strings.Replace(guiHTML, "</body>", guiMotionGoldenScript+"</body>", 1)
	if page == guiHTML {
		t.Fatal("designer page has no closing body")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	var lastHTML string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", page)
	})
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(guiMotionDesignerSource))+`}`)
	})
	mux.HandleFunc("POST /__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ok, html, msgs := renderSui(r.PostFormValue("sui"))
		if ok {
			lastHTML = html
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
	})
	mux.HandleFunc("POST /__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		model, msgs, ok := guiOpen(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		body := StringifyJson(model)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,`+body[1:])
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", lastHTML)
	})
	go func() { _ = http.Serve(ln, mux) }()

	baseURL := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--no-first-run",
		"--no-default-browser-check",
		"--user-data-dir=" + t.TempDir(),
		"--virtual-time-budget=20000",
		"--dump-dom",
		baseURL,
	}
	out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
	text := string(out)
	if err != nil {
		t.Fatalf("headless browser failed: %v\n%s", err, text)
	}
	if msg := goldenAttribute(text, "data-golden-error"); msg != "" {
		t.Fatalf("motion panel golden script failed: %s", msg)
	}
	encoded := goldenAttribute(text, "data-golden")
	if encoded == "" {
		t.Fatalf("golden snapshot missing from designer page")
	}
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		t.Fatalf("decode golden attribute: %v", err)
	}
	var got guiMotionGolden
	if err := json.Unmarshal([]byte(decoded), &got); err != nil {
		t.Fatalf("decode golden JSON: %v\n%s", err, decoded)
	}
	if err := validateGuiMotionGolden(got); err != nil {
		t.Fatalf("invalid golden snapshot: %v", err)
	}

	goldenPath := filepath.Join("testdata", "gui_motion_golden", "panels.json")
	update := os.Getenv("SUI_UPDATE_GUI_MOTION_GOLDEN") == "1"
	wantJSON, marshalErr := json.MarshalIndent(got, "", "  ")
	if marshalErr != nil {
		t.Fatalf("encode golden: %v", marshalErr)
	}
	wantJSON = append(wantJSON, '\n')
	if update {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, wantJSON, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden %s missing; run with SUI_UPDATE_GUI_MOTION_GOLDEN=1", goldenPath)
	}
	var want guiMotionGolden
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("decode golden %s: %v", goldenPath, err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("motion panel golden mismatch\nwant: %s\ngot:  %s", string(raw), string(wantJSON))
	}
}

func goldenAttribute(text, name string) string {
	marker := name + `="`
	start := strings.Index(text, marker)
	if start < 0 {
		return ""
	}
	start += len(marker)
	end := strings.Index(text[start:], `"`)
	if end < 0 {
		return ""
	}
	return text[start : start+end]
}

func validateGuiMotionGolden(g guiMotionGolden) error {
	if !g.PanelOn {
		return errString("animation panel is not open")
	}
	if g.Exit.Title != "Exit animation" || len(g.Exit.Labels) == 0 || len(g.Exit.Buttons) == 0 {
		return errString("exit overlay was not captured")
	}
	if !g.Exit.SpringInitiallyHidden || !g.Exit.SpringVisible {
		return errString("exit spring visibility was not captured")
	}
	if g.Guard.Title != "Trigger parameters" || !g.Guard.HasGuard || g.Guard.GuardType != "text" {
		return errString("guard overlay was not captured")
	}
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }

const guiMotionGoldenScript = `<script>
(async function () {
  function wait(ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); }
  function buttonByText(text) {
    var list = document.querySelectorAll('#mg-param-body button');
    for (var i = 0; i < list.length; i++) if ((list[i].textContent || '').trim() === text) return list[i];
    return null;
  }
  function labels() {
    return Array.prototype.map.call(document.querySelectorAll('#mg-param-body label'), function (el) { return (el.textContent || '').trim(); });
  }
  function buttons() {
    return Array.prototype.map.call(document.querySelectorAll('#mg-param-body button'), function (el) { return (el.textContent || '').trim(); });
  }
  function inputAfter(labelText) {
    var list = document.querySelectorAll('#mg-param-body label');
    for (var i = 0; i < list.length; i++) {
      if ((list[i].textContent || '').trim() === labelText) return list[i].nextElementSibling;
    }
    return null;
  }
  try {
    for (var i = 0; i < 120 && !document.querySelector('#nodes .node'); i++) await wait(100);
    var node = Array.prototype.slice.call(document.querySelectorAll('#nodes .node')).filter(function (el) {
      return (el.textContent || '').replace(/\s+/g, ' ').indexOf('Animated') >= 0;
    })[0];
    if (!node) throw new Error('graph node not found');
    node.click();
    await wait(500);
    var open = document.querySelectorAll('#inspector button');
    var openPanel = null;
    for (var o = 0; o < open.length; o++) if ((open[o].textContent || '').trim() === 'Open animation panel') openPanel = open[o];
    if (!openPanel) throw new Error('Open animation panel button not found');
    openPanel.click();
    await wait(600);
    var exitButton = null;
    var toolbar = document.querySelectorAll('#mg-panel button');
    for (var b = 0; b < toolbar.length; b++) if ((toolbar[b].textContent || '').trim() === 'Exit animation') exitButton = toolbar[b];
    if (!exitButton) throw new Error('Exit animation button not found');
    exitButton.click();
    await wait(400);
    var exitTitle = (document.getElementById('mg-param-title') || {}).textContent || '';
    var springInput = document.getElementById('mg-exit-stiffness');
    var springHidden = !springInput || getComputedStyle(springInput).display === 'none';
    var stateSel = document.getElementById('mg-exit-state');
    if (!stateSel) throw new Error('exit state select missing');
    var stateValues = [], stateLabels = [];
    for (var s = 0; s < stateSel.options.length; s++) {
      stateValues.push(stateSel.options[s].value);
      stateLabels.push((stateSel.options[s].textContent || '').trim());
    }
    stateSel.value = 'b';
    stateSel.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(300);
    var spring = document.getElementById('mg-exit-spring');
    if (!spring) throw new Error('exit spring toggle missing');
    spring.checked = true;
    spring.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    var springVisible = getComputedStyle(document.getElementById('mg-exit-stiffness')).display !== 'none';
    var exitLabels = labels();
    var exitButtons = buttons();
    var ok = buttonByText('OK');
    if (ok) ok.click();
    await wait(500);
    var link = document.querySelector('#mg-lines path[data-link]');
    if (!link) throw new Error('graph link missing');
    link.dispatchEvent(new MouseEvent('dblclick', { bubbles: true, cancelable: true }));
    await wait(500);
    var guard = inputAfter('guard');
    var guardHint = '';
    var hints = document.querySelectorAll('#mg-param-body .hint');
    for (var h = 0; h < hints.length; h++) {
      if ((hints[h].textContent || '').indexOf('Optional condition using SUI state names') >= 0) guardHint = (hints[h].textContent || '').trim();
    }
    var snapshot = {
      panelOn: !!document.getElementById('mg-panel') && document.getElementById('mg-panel').classList.contains('on'),
      exit: {
        title: exitTitle,
        stateValues: stateValues,
        stateLabels: stateLabels,
        labels: exitLabels,
        buttons: exitButtons,
        springInitiallyHidden: springHidden,
        springVisible: springVisible
      },
      guard: {
        title: (document.getElementById('mg-param-title') || {}).textContent || '',
        hasGuard: !!guard,
        guardType: guard ? guard.type : '',
        guardPlaceholder: guard ? guard.placeholder : '',
        guardHint: guardHint,
        buttons: buttons()
      }
    };
    document.body.setAttribute('data-golden', encodeURIComponent(JSON.stringify(snapshot)));
    document.title = 'GOLDEN';
  } catch (e) {
    document.body.setAttribute('data-golden-error', (e && e.message) ? e.message : String(e));
    document.title = 'GOLDEN_FAIL';
  }
})();
</script>`
