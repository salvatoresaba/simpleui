package sui

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const unionSource = `enum Kind { Text, Number }

state value: string | int = "hello"
state maybe: string? = null
state kind: Kind = Kind.Text
state mixed: array[string | int] = ["a", 1]
state message = ""

fn inspect() -> null {
	if value is string {
		message = "string:" + value
	} else {
		message = "number:" + str(value)
	}
	if maybe is null {
		message = message + " empty"
	} else {
		message = message + " " + maybe
	}
	if kind is Kind {
		message = message + " enum"
	}
	let first = mixed[0]
	if first is string {
		message = message + " first-string"
	} else {
		message = message + " first-number"
	}
}

fn setText() -> null {
	value = "text"
	maybe = "ready"
}

fn setNumber() -> null {
	value = 42
	maybe = null
}

view main {
	p "{message}"
	button "Inspect" onClick: inspect
	button "Text" onClick: setText
	button "Number" onClick: setNumber
}
`

func TestUnionTypecheckAndGenerate(t *testing.T) {
	prog, dl := ParseSource(unionSource, "union.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, marker := range []string{
		`typeof value === "string"`,
		`maybe === null`,
		`kind === Kind.Text || kind === Kind.Number`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}
}

func TestUnionErrors(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name: "wrong union assignment",
			src: `state v: string | int = true
view main { p "x" }`,
			wantErr: "declared as string | int",
		},
		{
			name: "impossible guard",
			src: `state n: int = 1
fn run() -> null {
	if n is string {
		n = 2
	}
}
view main { button "x" onClick: run }`,
			wantErr: "can never succeed",
		},
		{
			name: "wrong call argument",
			src: `fn show(v: string | int) -> string { return str(v) }
fn run() -> null {
	let x = show(true)
}
view main { button "x" onClick: run }`,
			wantErr: "must be string | int",
		},
		{
			name: "struct guard rejected",
			src: `type P struct { name: string }
state p: P | null = null
fn run() -> null {
	if p is P {
		p = null
	}
}
view main { button "x" onClick: run }`,
			wantErr: "type guard for P is not supported at runtime",
		},
		{
			name: "wrong return",
			src: `fn pick() -> string | int { return true }
view main { p "{str(pick())}" }`,
			wantErr: "cannot return a value of type bool",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "union.sui")
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

func TestUnionShortCircuitNarrowing(t *testing.T) {
	src := `state v: string | int = "a"
state maybe: string | null = null
state out = ""
fn run() -> null {
	if v is string and len(v) > 0 {
		out = "non-empty " + v
	}
	if maybe is null or len(maybe) == 0 {
		out = out + " empty"
	} else {
		out = out + " " + maybe
	}
}
view main {
	p "{out}"
	button "Run" onClick: run
}`
	prog, dl := ParseSource(src, "union.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
}

func TestUnionDesignerRoundTrip(t *testing.T) {
	model, msgs, ok := guiOpen(unionSource)
	if !ok {
		t.Fatalf("guiOpen: %s", StringifyJson(msgs))
	}
	states := get(model, "states")
	var valueType, maybeType string
	for _, st := range states.arr {
		switch asStr(get(st, "name")) {
		case "value":
			valueType = asStr(get(st, "type"))
		case "maybe":
			maybeType = asStr(get(st, "type"))
		}
	}
	if valueType != "string | int" || maybeType != "string | null" {
		t.Fatalf("designer lost union types: value=%q maybe=%q", valueType, maybeType)
	}
	var inspectBody string
	for _, f := range get(model, "funcs").arr {
		if asStr(get(f, "name")) == "inspect" {
			inspectBody = asStr(get(f, "body"))
		}
	}
	for _, want := range []string{"if value is string", "if maybe is null", "if kind is Kind", "if first is string"} {
		if !strings.Contains(inspectBody, want) {
			t.Errorf("designer lost %q\\n---\\n%s", want, inspectBody)
		}
	}
}

func TestUnionFmt(t *testing.T) {
	out, dl := FormatSource(unionSource, "union.sui")
	if dl != nil && dl.HasErrors() {
		t.Fatalf("fmt errors: %v", dl)
	}
	prog, pdl := ParseSource(out, "union.sui")
	if pdl.HasErrors() {
		t.Fatalf("formatted source does not parse: %v\n%s", pdl, out)
	}
	if cdl := Check(prog); cdl.HasErrors() {
		t.Fatalf("formatted source does not check: %v\n%s", cdl, out)
	}
	out2, _ := FormatSource(out, "union.sui")
	if out != out2 {
		t.Errorf("fmt is not idempotent for unions/guards")
	}
}

func TestUnionRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping union runtime test")
	}
	index, css, js, diags, ok := guiBuildFiles(unionSource)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + unionScript + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}
	ln, err := ListenLocal()
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
	t.Fatalf("union runtime e2e failed:\n%s", detail)
}

const unionScript = `<script>
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
    button('Inspect').click(); await wait(200);
    var text = document.body.innerText;
    if (text.indexOf('string:hello empty enum first-string') < 0) throw new Error('initial: ' + text);
    button('Number').click(); await wait(150);
    button('Inspect').click(); await wait(200);
    text = document.body.innerText;
    if (text.indexOf('number:42 empty enum first-string') < 0) throw new Error('number: ' + text);
    button('Text').click(); await wait(150);
    button('Inspect').click(); await wait(200);
    text = document.body.innerText;
    if (text.indexOf('string:text ready enum first-string') < 0) throw new Error('text: ' + text);
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`
