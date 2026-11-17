package main

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const literalSource = `state mode: "idle" | "loading" | "error" = "idle"
state level: 1 | 2 | 3 = 1
state mixed: array["idle" | "loading"] = ["idle", "loading"]
state message = ""

fn describe() -> null {
	switch mode {
	case "idle" {
		message = "idle"
	}
	case "loading" {
		message = "loading"
	}
	case "error" {
		message = "error"
	}
	}
}

fn nextMode() -> null {
	if mode == "idle" {
		mode = "loading"
	} else if mode == "loading" {
		mode = "error"
	} else {
		mode = "idle"
	}
	describe()
}

fn raiseLevel() -> null {
	if level == 1 {
		level = 2
	} else if level == 2 {
		level = 3
	} else {
		level = 1
	}
	message = "level " + str(level)
}

fn guard() -> null {
	if mode is "error" {
		message = "guard-error"
	} else {
		message = "guard-other"
	}
	let first = mixed[0]
	if first is "idle" {
		message = message + "-idle"
	}
}

view main {
	p "{message}"
	button "Next" onClick: nextMode
	button "Raise" onClick: raiseLevel
	button "Guard" onClick: guard
}
`

func TestLiteralTypecheckAndGenerate(t *testing.T) {
	prog, dl := ParseSource(literalSource, "literal.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, marker := range []string{
		`switch (mode)`,
		`mode === "error"`,
		"case `idle`:",
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}
}

func TestLiteralNumericAndBoolean(t *testing.T) {
	src := `state x: 1.0 | 2.0 = 1.0
state flag: true | false = true
state out = ""
fn run() -> null {
	switch x {
	case 1.0 {
		out = "one"
	}
	case 2.0 {
		out = "two"
	}
	}
	switch flag {
	case true {
		out = out + " t"
	}
	case false {
		out = out + " f"
	}
	}
}
view main {
	p "{out}"
	button "Run" onClick: run
}`
	prog, dl := ParseSource(src, "literal.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
}

func TestLiteralTypeErrors(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name: "wrong string literal",
			src: `state mode: "idle" | "loading" = "run"
view main { p "{mode}" }`,
			wantErr: `declared as "idle" | "loading"`,
		},
		{
			name: "wrong number literal",
			src: `state level: 1 | 2 = 3
view main { p "{str(level)}" }`,
			wantErr: "declared as 1 | 2",
		},
		{
			name: "wrong boolean literal",
			src: `state flag: true | false = 1
view main { p "{str(flag)}" }`,
			wantErr: "declared as true | false",
		},
		{
			name: "wrong array element",
			src: `state xs: array["a" | "b"] = ["a", "c"]
view main { p "{str(xs)}" }`,
			wantErr: `declared as array["a" | "b"]`,
		},
		{
			name: "wrong argument",
			src: `fn take(m: "a" | "b") -> null {
	message = m
}
state message = ""
fn run() -> null {
	take("c")
}
view main { p "{message}" }`,
			wantErr: `argument 1 of take must be "a" | "b"`,
		},
		{
			name: "wrong return",
			src: `fn pick() -> "yes" | "no" {
	return "maybe"
}
view main { p "{pick()}" }`,
			wantErr: `cannot return a value of type "maybe"`,
		},
		{
			name: "wrong struct field",
			src: `type Job struct {
	mode: "idle" | "busy"
}
state job: Job = Job { mode: "nope" }
view main { p "{job.mode}" }`,
			wantErr: `field "mode" of Job expects "idle" | "busy", got "nope"`,
		},
		{
			name: "float is not an int literal",
			src: `state level: 1 | 2 = 1.0
view main { p "{str(level)}" }`,
			wantErr: "declared as 1 | 2",
		},
		{
			name: "missing switch case",
			src: `state mode: "idle" | "loading" = "idle"
fn run() -> null {
	switch mode {
	case "idle" {
		mode = "loading"
	}
	}
}
view main { button "Run" onClick: run }`,
			wantErr: `not exhaustive; missing: "loading"`,
		},
		{
			name: "duplicate case",
			src: `state mode: "idle" | "loading" = "idle"
fn run() -> null {
	switch mode {
	case "idle" {
		mode = "loading"
	}
	case "idle" {
		mode = "loading"
	}
	case "loading" {
		mode = "idle"
	}
	}
}
view main { button "Run" onClick: run }`,
			wantErr: "duplicate case value in switch",
		},
		{
			name: "impossible guard",
			src: `state level: 1 | 2 = 1
fn run() -> null {
	if level is 3 {
		level = 1
	}
}
view main { button "Run" onClick: run }`,
			wantErr: "can never succeed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "literal.sui")
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

func TestLiteralDesignerRoundTrip(t *testing.T) {
	model, msgs, ok := guiOpen(literalSource)
	if !ok {
		t.Fatalf("guiOpen: %s", StringifyJson(msgs))
	}
	got := map[string]string{}
	for _, st := range get(model, "states").arr {
		got[asStr(get(st, "name"))] = asStr(get(st, "type"))
	}
	if got["mode"] != `"idle" | "loading" | "error"` {
		t.Errorf("designer lost the string literal union: %q", got["mode"])
	}
	if got["level"] != "1 | 2 | 3" {
		t.Errorf("designer lost the numeric literal union: %q", got["level"])
	}
	if got["mixed"] != `array["idle" | "loading"]` {
		t.Errorf("designer lost the array literal type: %q", got["mixed"])
	}
	var describe, guard string
	for _, f := range get(model, "funcs").arr {
		switch asStr(get(f, "name")) {
		case "describe":
			describe = asStr(get(f, "body"))
		case "guard":
			guard = asStr(get(f, "body"))
		}
	}
	if !strings.Contains(describe, "switch mode") || !strings.Contains(describe, `case "error"`) {
		t.Errorf("designer lost the switch body:\n%s", describe)
	}
	if !strings.Contains(guard, `if mode is "error"`) {
		t.Errorf("designer lost the literal guard:\n%s", guard)
	}
}

func TestLiteralFmt(t *testing.T) {
	out, dl := FormatSource(literalSource, "literal.sui")
	if dl != nil && dl.HasErrors() {
		t.Fatalf("fmt errors: %v", dl)
	}
	prog, pdl := ParseSource(out, "literal.sui")
	if pdl.HasErrors() {
		t.Fatalf("formatted source does not parse: %v\n%s", pdl, out)
	}
	if cdl := Check(prog); cdl.HasErrors() {
		t.Fatalf("formatted source does not check: %v\n%s", cdl, out)
	}
	out2, _ := FormatSource(out, "literal.sui")
	if out != out2 {
		t.Errorf("fmt is not idempotent for literal types")
	}
	for _, want := range []string{
		`state mode: "idle" | "loading" | "error" = "idle"`,
		`state level: 1 | 2 | 3 = 1`,
		`state mixed: array["idle" | "loading"]`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("formatted source lost %q", want)
		}
	}
}

func TestLiteralRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping literal runtime test")
	}
	index, css, js, diags, ok := guiBuildFiles(literalSource)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + literalScript + page[idx:]
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
	t.Fatalf("literal runtime e2e failed:\n%s", detail)
}

const literalScript = `<script>
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
    button('Next').click(); await wait(150);
    var text = document.body.innerText;
    if (text.indexOf('loading') < 0) throw new Error('next: ' + text);
    button('Next').click(); await wait(150);
    text = document.body.innerText;
    if (text.indexOf('error') < 0) throw new Error('error: ' + text);
    button('Guard').click(); await wait(150);
    text = document.body.innerText;
    if (text.indexOf('guard-error-idle') < 0) throw new Error('guard: ' + text);
    button('Raise').click(); await wait(150);
    text = document.body.innerText;
    if (text.indexOf('level 2') < 0) throw new Error('level: ' + text);
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`
