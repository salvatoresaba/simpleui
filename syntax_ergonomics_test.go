package sui

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const syntaxErgonomicsSource = `type Person struct { name: string, age: int }
state out = ""

fn greet(name: string, excited: bool = false, punctuation: string = "!") -> string {
	let mood = "hello"
	if excited { mood = "HELLO" }
	return mood + " " + name + punctuation
}

fn run() -> null {
	let p = Person{ name: "Anna", age: 30 }
	let {name, age} = p
	let [first, second] = [10, 20, 30]
	let nums = [...[1, 2], ...[3, 4]]
	let defaults = {"a": 1, "b": 2}
	let merged = {...defaults, "b": 9, "c": 3}
	let m = {"x": 7, "y": 8}
	let {x, y} = m
	out = greet(punctuation: "?", name: name, excited: true) + " " + str(age) + " " + str(first) + " " + str(second) + " " + str(nums) + " " + str(merged["a"]) + str(merged["b"]) + str(merged["c"]) + " " + str(x) + str(y)
}

view main {
	p "{out}"
	button "Run" onClick: run
}
`

func TestSyntaxErgonomicsTypecheck(t *testing.T) {
	prog, dl := ParseSource(syntaxErgonomicsSource, "syntax.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, marker := range []string{
		"let {name, age} = p;",
		"let [first, second] = [10, 20, 30];",
		"let nums = [...[1, 2], ...[3, 4]];",
		"let merged = ({...defaults, \"b\": 9, \"c\": 3});",
		"greet(name, true, `?`)",
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}
}

func TestSyntaxErgonomicsErrors(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name: "object pattern on array",
			src: `fn run() -> null {
	let {a} = [1, 2]
}
view main { button "x" onClick: run }`,
			wantErr: "object destructuring requires a struct or a map",
		},
		{
			name: "array pattern on struct",
			src: `type P struct { a: int }
fn run() -> null {
	let [a] = P{ a: 1 }
}
view main { button "x" onClick: run }`,
			wantErr: "array destructuring requires an array",
		},
		{
			name: "array spread of non array",
			src: `fn run() -> null {
	let x = [...3]
}
view main { button "x" onClick: run }`,
			wantErr: "spread in an array literal requires an array",
		},
		{
			name: "map spread of non map",
			src: `fn run() -> null {
	let x = {...3}
}
view main { button "x" onClick: run }`,
			wantErr: "spread in a map literal requires a map",
		},
		{
			name: "mixed named and positional",
			src: `fn f(a: int, b: int) -> int { return a + b }
fn run() -> null {
	let x = f(1, b: 2)
}
view main { button "x" onClick: run }`,
			wantErr: "cannot mix positional and named arguments",
		},
		{
			name: "unknown named argument",
			src: `fn f(a: int) -> int { return a }
fn run() -> null {
	let x = f(nope: 1)
}
view main { button "x" onClick: run }`,
			wantErr: "unknown named argument",
		},
		{
			name: "duplicate destructuring name",
			src: `fn run() -> null {
	let {a, a} = {"a": 1}
}
view main { button "x" onClick: run }`,
			wantErr: "duplicate name",
		},
		{
			name: "duplicate named argument",
			src: `fn f(a: int) -> int { return a }
fn run() -> null {
	let x = f(a: 1, a: 2)
}
view main { button "x" onClick: run }`,
			wantErr: "duplicate named argument",
		},
		{
			name: "named builtin",
			src: `view main {
	p "{upper(s: "x")}"
}`,
			wantErr: "named arguments are only supported for user functions",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "syntax.sui")
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

// TestSyntaxErgonomicsDesignerRoundTrip keeps the visual designer lossless for
// the new forms: function bodies are stored as source text and are regenerated
// by stmtToSui/exprSource.
func TestSyntaxErgonomicsDesignerRoundTrip(t *testing.T) {
	model, msgs, ok := guiOpen(syntaxErgonomicsSource)
	if !ok {
		t.Fatalf("guiOpen failed: %s", StringifyJson(msgs))
	}
	var runBody string
	for _, f := range get(model, "funcs").arr {
		if asStr(get(f, "name")) == "run" {
			runBody = asStr(get(f, "body"))
		}
	}
	if runBody == "" {
		t.Fatal("run function body missing from the designer model")
	}
	for _, want := range []string{
		"let {name, age} = p",
		"let [first, second] = [10, 20, 30]",
		"let nums = [...[1, 2], ...[3, 4]]",
		"let merged = { ...defaults, \"b\": 9, \"c\": 3 }",
		"greet(punctuation: \"?\", name: name, excited: true)",
	} {
		if !strings.Contains(runBody, want) {
			t.Errorf("designer round-trip lost %q\n---\n%s", want, runBody)
		}
	}
}

func TestSyntaxErgonomicsFmt(t *testing.T) {
	out, dl := FormatSource(syntaxErgonomicsSource, "syntax.sui")
	if dl != nil && dl.HasErrors() {
		t.Fatalf("fmt errors: %v", dl)
	}
	prog, pdl := ParseSource(out, "syntax.sui")
	if pdl.HasErrors() {
		t.Fatalf("formatted output does not parse: %v\n%s", pdl, out)
	}
	if cdl := Check(prog); cdl.HasErrors() {
		t.Fatalf("formatted output does not check: %v\n%s", cdl, out)
	}
	out2, _ := FormatSource(out, "syntax.sui")
	if out != out2 {
		t.Errorf("fmt is not idempotent for destructuring/spread/named args")
	}
}

func TestSyntaxErgonomicsRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping syntax runtime test")
	}
	index, css, js, diags, ok := guiBuildFiles(syntaxErgonomicsSource)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + syntaxErgonomicsScript + page[idx:]
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
	t.Fatalf("syntax ergonomics e2e failed:\n%s", detail)
}

const syntaxErgonomicsScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  try {
    await wait(500);
    var button = document.querySelector('button');
    if (!button) throw new Error('Run button missing');
    button.click();
    await wait(400);
    var text = document.body.innerText;
    var expected = 'HELLO Anna? 30 10 20 1,2,3,4 193 78';
    if (text.indexOf(expected) < 0) throw new Error('unexpected output: ' + text);
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`
