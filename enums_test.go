package sui

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const enumsSource = `enum Step { Idle, Loading, Done, Failed }

state step: Step = Step.Idle
state count = 0

fn stepLabel(s: Step) -> string {
	switch s {
		case Step.Idle {
			return "Ready"
		}
		case Step.Loading {
			return "Loading..."
		}
		case Step.Done {
			return "Done"
		}
		case Step.Failed {
			return "Failed"
		}
	}
	return "Unknown"
}

fn next() -> null {
	switch step {
		case Step.Idle {
			step = Step.Loading
		}
		case Step.Loading {
			step = Step.Done
			count += 1
		}
		case Step.Done {
			step = Step.Idle
		}
		case Step.Failed {
			step = Step.Idle
		}
	}
}

fn fail() -> null {
	step = Step.Failed
}

view main {
	card "" class: "card" {
		h1 "Enums and switch"
		p "step: {stepLabel(step)} ({step}) · completed: {str(count)}"
		row {
			button "Next" onClick: next
			button "Fail" onClick: fail
		}
	}
}
`

func TestEnumsTypecheckAndGenerate(t *testing.T) {
	prog, dl := ParseSource(enumsSource, "enums.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, marker := range []string{
		`const Step = Object.freeze({"Idle": "Idle", "Loading": "Loading", "Done": "Done", "Failed": "Failed"});`,
		`switch (step) {`,
		`case Step.Loading:`,
		`setStep(Step.Done);`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}
}

func TestEnumsErrors(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name: "non exhaustive",
			src: `enum Status { Idle, Loading, Error }
state s: Status = Status.Idle
fn run() -> null {
	switch s {
		case Status.Idle { }
		case Status.Loading { }
	}
}
view main { button "x" onClick: run }`,
			wantErr: "switch on Status is not exhaustive",
		},
		{
			name: "unknown member",
			src: `enum Status { Idle }
state s: Status = Status.Idle
fn run() -> null {
	switch s {
		case Status.Nope { }
		default { }
	}
}
view main { button "x" onClick: run }`,
			wantErr: "unknown enum member",
		},
		{
			name: "duplicate case",
			src: `enum Status { Idle, Loading }
state s: Status = Status.Idle
fn run() -> null {
	switch s {
		case Status.Idle { }
		case Status.Idle { }
		case Status.Loading { }
	}
}
view main { button "x" onClick: run }`,
			wantErr: "duplicate case value",
		},
		{
			name: "enum name as value",
			src: `enum Status { Idle }
fn run() -> null {
	let x = Status
}
view main { button "x" onClick: run }`,
			wantErr: "cannot be used as a value",
		},
		{
			name: "wrong enum assignment",
			src: `enum Status { Idle }
state s: Status = Status.Idle
fn run() -> null { s = "Idle" }
view main { button "x" onClick: run }`,
			wantErr: "cannot assign a value of type string",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "enums.sui")
			if dl.HasErrors() {
				t.Fatalf("parse: %v", dl.String())
			}
			dl = Check(prog)
			if !dl.HasErrors() || !strings.Contains(dl.String(), tc.wantErr) {
				t.Fatalf("expected %q, got: %v", tc.wantErr, dl.String())
			}
		})
	}
}

// TestEnumsDesignerRoundTrip keeps enums and switch bodies in the designer
// model; the visual editor does not edit them yet but must preserve them.
func TestEnumsDesignerRoundTrip(t *testing.T) {
	model, msgs, ok := guiOpen(enumsSource)
	if !ok {
		t.Fatalf("guiOpen: %s", StringifyJson(msgs))
	}
	enums := get(model, "enums")
	if len(enums.arr) != 1 || asStr(get(enums.arr[0], "name")) != "Step" {
		t.Fatalf("enum missing from designer model: %s", StringifyJson(enums))
	}
	var nextBody string
	for _, f := range get(model, "funcs").arr {
		if asStr(get(f, "name")) == "next" {
			nextBody = asStr(get(f, "body"))
		}
	}
	for _, want := range []string{"switch step {", "case Step.Loading {", "step = Step.Done"} {
		if !strings.Contains(nextBody, want) {
			t.Errorf("designer round-trip lost %q\\n---\\n%s", want, nextBody)
		}
	}
	// The regenerated body must still parse and check in its declaration context.
	wrapped := "enum Step { Idle, Loading, Done, Failed }\n" +
		"state step: Step = Step.Idle\nstate count = 0\n" +
		"fn next() -> null {\n" + nextBody + "\n}\n" +
		"view main { button \"x\" onClick: next }\n"
	prog, dl := ParseSource(wrapped, "next.sui")
	if dl.HasErrors() {
		t.Fatalf("designer body does not parse: %v\n%s", dl, wrapped)
	}
	if cdl := Check(prog); cdl.HasErrors() {
		t.Fatalf("designer body does not check: %v\n%s", cdl, wrapped)
	}
}

func TestEnumsRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping enum runtime test")
	}
	index, css, js, diags, ok := guiBuildFiles(enumsSource)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + enumsScript + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := guiTestMux()
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
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=15000", "--dump-dom", url,
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
	t.Fatalf("enum runtime e2e failed:\n%s", detail)
}

const enumsScript = `<script>
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
    await wait(400);
    var text = document.body.innerText;
    if (text.indexOf('step: Ready (Idle) · completed: 0') < 0) throw new Error('initial state: ' + text);
    button('Next').click(); await wait(250);
    text = document.body.innerText;
    if (text.indexOf('step: Loading... (Loading) · completed: 0') < 0) throw new Error('loading state: ' + text);
    button('Next').click(); await wait(250);
    text = document.body.innerText;
    if (text.indexOf('step: Done (Done) · completed: 1') < 0) throw new Error('done state: ' + text);
    button('Fail').click(); await wait(250);
    text = document.body.innerText;
    if (text.indexOf('step: Failed (Failed) · completed: 1') < 0) throw new Error('failed state: ' + text);
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`
