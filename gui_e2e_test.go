package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// gui_e2e_test.go drives the real designer page in headless Chrome/Edge and
// exercises the HINTS panel plus the per-element hint editor.
func TestGuiDesignerE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping designer end-to-end test")
	}

	const source = `view main {
	chart "Bars" name: "b" data: "Month,Value\nJan,12\nFeb,18" chartType: "bar"
}`

	page := strings.Replace(guiHTML, "</body>", guiE2EScript+"</body>", 1)
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
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(source))+`}`)
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

	// The page embeds no external resources; wait briefly for the listener.
	url := "http://" + ln.Addr().String() + "/"
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	profile := t.TempDir()
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--no-first-run",
		"--no-default-browser-check",
		"--user-data-dir=" + profile,
		"--virtual-time-budget=20000",
		"--dump-dom",
		url,
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
		end := marker + 400
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("designer e2e failed:\n%s", detail)
}

// TestGuiMotionDesignerE2E drives the real animation panel: it edits the exit
// overlay and a link guard through the same controls a user clicks.
func TestGuiMotionDesignerE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion designer end-to-end test")
	}

	const source = `state gate = true
state mode = "edit"
view main {
	card "Animated" class: "card" motionGraph: "\{\"initial\":\"a\",\"states\":[\{\"id\":\"a\",\"name\":\"A\",\"css\":\"width: 120px; height: 80px; background-color: #223344;\"\},\{\"id\":\"b\",\"name\":\"B\",\"css\":\"width: 200px; height: 100px; background-color: #557799;\"\}],\"links\":[\{\"id\":\"l1\",\"from\":\"a\",\"to\":\"b\",\"event\":\"click\",\"duration\":\"200ms\"\}]\}"
}`

	page := strings.Replace(guiHTML, "</body>", guiMotionDesignerScript+"</body>", 1)
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
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(source))+`}`)
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

	url := "http://" + ln.Addr().String() + "/"
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
		url,
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
		end := marker + 600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("motion designer e2e failed:\n%s", detail)
}

func findHeadlessBrowser() string {
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

const guiE2EScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-result';
      document.body.appendChild(el);
    }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  var wait = function (ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); };
  try {
    for (var i = 0; i < 120 && !document.querySelector('#nodes .node'); i++) await wait(100);
    var chartNode = Array.prototype.slice.call(document.querySelectorAll('#nodes .node')).filter(function (el) {
      return (el.textContent || '').replace(/\s+/g, ' ').indexOf('Bars') >= 0;
    })[0];
    if (!chartNode) throw new Error('chart node not found');
    chartNode.click();
    await wait(400);
    var settings = Array.prototype.slice.call(document.querySelectorAll('#inspector button')).filter(function (b) {
      var t = b.textContent.trim();
      return t === 'Graph settings...' || t === 'Chart settings...';
    })[0];
    if (!settings) throw new Error('chart settings button not found');
    settings.click();
    await wait(400);
    var interactions = Array.prototype.slice.call(document.querySelectorAll('#csettings .cset-tab')).filter(function (b) { return b.textContent.trim() === 'Interactions'; })[0];
    interactions.click();
    await wait(250);
    var hintsButton = Array.prototype.slice.call(document.querySelectorAll('#csettings button')).filter(function (b) { return b.textContent.trim() === 'HINTS...'; })[0];
    if (!hintsButton) throw new Error('HINTS button not found');
    hintsButton.click();
    await wait(400);
    var bg = document.getElementById('chints-bg');
    var opacity = document.getElementById('chints-opacity');
    if (!bg || bg.type !== 'color') throw new Error('background color picker missing');
    if (!opacity || opacity.type !== 'range') throw new Error('background alpha range missing');
    if (!document.getElementById('chints-groups')) throw new Error('HINTS groups missing');
    opacity.value = '0.4';
    opacity.dispatchEvent(new Event('input', { bubbles: true }));
    await wait(150);
    if (document.getElementById('chints-opacity-label').textContent !== '40%') throw new Error('alpha label not updated');
    var apply = Array.prototype.slice.call(document.querySelectorAll('#chintscfgbox .pfbtns button')).filter(function (b) { return b.textContent.trim() === 'Apply'; })[0];
    if (!apply) throw new Error('HINTS Apply missing');
    apply.click();
    await wait(1000);
    if (document.getElementById('out').value.indexOf('stateHintOpacity: "0.4"') < 0) throw new Error('hint opacity not applied');
    var settingsCancel = Array.prototype.slice.call(document.querySelectorAll('#csettings .pfbtns button')).filter(function (b) { return b.textContent.trim() === 'Cancel'; })[0];
    if (settingsCancel) settingsCancel.click();
    await wait(400);
    var typeSelect = null;
    var selects = document.querySelectorAll('#inspector select');
    for (var s = 0; s < selects.length; s++) {
      if (selects[s].querySelector('option[value="candlestick"]')) { typeSelect = selects[s]; break; }
    }
    if (!typeSelect) throw new Error('chart type selector not found');
    var stateOption = typeSelect.querySelector('option[value="state"]');
    if (stateOption && stateOption.textContent.indexOf('legacy') < 0) throw new Error('state machine still listed as a normal chart type');
    typeSelect.value = 'candlestick';
    typeSelect.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    if (!document.getElementById('confdlg').classList.contains('on')) throw new Error('incompatible type warning not shown');
    if (document.getElementById('conf-text').textContent.indexOf('sample dataset') < 0) throw new Error('warning does not mention sample data');
    document.getElementById('conf-ok').click();
    await wait(1200);
    var finalSource = document.getElementById('out').value;
    if (finalSource.indexOf('chartType: "candlestick"') < 0) throw new Error('candlestick type not applied');
    if (finalSource.indexOf('"open"') < 0 && finalSource.indexOf('open') < 0) throw new Error('OHLC sample data not applied');
    if (finalSource.indexOf('"close"') < 0 && finalSource.indexOf('close') < 0) throw new Error('OHLC sample data not applied');
    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`

const guiMotionDesignerScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-result';
      document.body.appendChild(el);
    }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  var wait = function (ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); };
  function buttonByText(root, text) {
    var scope = typeof root === 'string' ? document.querySelectorAll(root + ' button') : root.querySelectorAll('button');
    for (var i = 0; i < scope.length; i++) {
      if ((scope[i].textContent || '').trim() === text) return scope[i];
    }
    return null;
  }
  function fieldInput(text) {
    var labels = document.querySelectorAll('#mg-param-body label');
    for (var i = 0; i < labels.length; i++) {
      if ((labels[i].textContent || '').trim() === text) return labels[i].nextElementSibling;
    }
    return null;
  }
  try {
    for (var i = 0; i < 120 && !document.querySelector('#nodes .node'); i++) await wait(100);
    var graphNode = Array.prototype.slice.call(document.querySelectorAll('#nodes .node')).filter(function (el) {
      return (el.textContent || '').replace(/\s+/g, ' ').indexOf('Animated') >= 0;
    })[0];
    if (!graphNode) throw new Error('graph node not found');
    graphNode.click();
    await wait(500);

    var open = buttonByText('#inspector', 'Open animation panel');
    if (!open) throw new Error('Open animation panel button not found');
    open.click();
    await wait(600);
    var panel = document.getElementById('mg-panel');
    if (!panel || !panel.classList.contains('on')) throw new Error('animation panel did not open');

    var exitButton = buttonByText('#mg-panel', 'Exit animation');
    if (!exitButton) throw new Error('Exit animation toolbar button not found');
    exitButton.click();
    await wait(400);
    var title = document.getElementById('mg-param-title');
    if (!title || title.textContent.trim() !== 'Exit animation') throw new Error('exit overlay title missing');
    var exitState = document.querySelector('#mg-param-body select');
    if (!exitState) throw new Error('exit state selector missing');
    if (!exitState.querySelector('option[value="b"]')) throw new Error('exit destination state missing');
    exitState.value = 'b';
    exitState.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(300);
    var duration = fieldInput('duration');
    if (!duration) throw new Error('exit duration field missing');
    duration.value = '180ms';
    duration.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(300);
    if (!mgCfg || !mgCfg.exit || mgCfg.exit.state !== 'b') throw new Error('exit state was not stored');
    if (mgCfg.exit.duration !== '180ms') throw new Error('exit duration was not stored');
    var out = document.getElementById('out');
    if (!out || out.value.indexOf('exit') < 0 || out.value.indexOf('180ms') < 0) throw new Error('exit settings missing from source');
    var ok = buttonByText('#mg-param-body', 'OK');
    if (!ok) throw new Error('exit OK button missing');
    ok.click();
    await wait(500);

    var link = document.querySelector('#mg-lines path[data-link]');
    if (!link) throw new Error('graph link not found');
    link.dispatchEvent(new MouseEvent('dblclick', { bubbles: true, cancelable: true }));
    await wait(500);
    if (!title || title.textContent.trim() !== 'Trigger parameters') throw new Error('trigger overlay did not open');
    var guard = fieldInput('guard');
    if (!guard) throw new Error('guard field missing');
    guard.value = 'gate &&';
    guard.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(300);
    var notice = document.getElementById('mg-notify');
    if (!notice || notice.textContent.indexOf('operator') < 0) throw new Error('invalid guard was not rejected');
    if (mgCfg.links[0].guard) throw new Error('invalid guard was stored');
    guard.value = 'gate && mode == "edit"';
    guard.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    if (mgCfg.links[0].guard !== 'gate && mode == "edit"') throw new Error('valid guard was not stored');
    out = document.getElementById('out');
    if (!out || out.value.indexOf('guard') < 0 || out.value.indexOf('mode ==') < 0) throw new Error('guard missing from source');
    var linkOk = buttonByText('#mg-param-body', 'OK');
    if (!linkOk) throw new Error('trigger OK button missing');
    linkOk.click();
    await wait(300);
    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`
