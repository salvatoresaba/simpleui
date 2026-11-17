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

// runHeadlessFeature renders source, appends script before </body>, serves the
// page and runs it in a headless browser. The script must call result(true) or
// result(false, message) using the shared helper below.
func runHeadlessFeature(t *testing.T, source, script string) {
	t.Helper()
	browser := findHeadlessBrowser()
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
	t.Fatalf("feature e2e failed:\n%s", detail)
}

const featureScriptHead = `<script>
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

const featureScriptTail = `    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`

func TestRestParameters(t *testing.T) {
	src := `fn addAll(base: int, ...values: array[int]) -> int {
	let total = base
	loop v in values {
		total = total + v
	}
	return total
}
state out = ""
fn run() -> null {
	out = str(addAll(10, 1, 2, 3)) + " " + str(addAll(5))
}
view main {
	p "{out}"
	button "Run" onClick: run
}`
	prog, dl := ParseSource(src, "rest.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "function addAll(base, ...values)") {
		t.Errorf("rest parameter was not emitted as JavaScript rest syntax")
	}
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Run').click(); await wait(200);
    var text = document.body.innerText;
    if (text.indexOf('16 5') < 0) throw new Error('rest result: ' + text);
`+featureScriptTail)

	bad := []struct{ name, src, want string }{
		{"not last", `fn f(...values: array[int], extra: int) -> int { return extra }
fn run() -> null { f(1, 2) }
view main { button "x" onClick: run }`, "must be the last parameter"},
		{"not array", `fn f(...values: array[string]) -> int { return 0 }
fn run() -> null { f(1) }
view main { button "x" onClick: run }`, "must be string, got int"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "rest.sui")
			if d.HasErrors() {
				t.Fatalf("parse: %v", d.String())
			}
			d = Check(p)
			if !d.HasErrors() || !strings.Contains(d.String(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, d.String())
			}
		})
	}
}

func TestConstAndReadonly(t *testing.T) {
	src := `const LIMIT: int = 3
const NAME = "SUI"
type Job struct {
	readonly id: string
	mode: string
}
state count = LIMIT
state job: Job = Job { id: "a", mode: "idle" }
fn run() -> null {
	count = count + 1
	job.mode = "busy"
}
view main {
	p "{NAME} {str(count)} {job.id} {job.mode}"
	button "Run" onClick: run
}`
	prog, dl := ParseSource(src, "const.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "const LIMIT = 3;") || !strings.Contains(js, "const NAME = `SUI`;") {
		t.Errorf("const declarations were not emitted")
	}
	if !strings.Contains(js, "function run()") {
		t.Errorf("missing function")
	}
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen: %s", StringifyJson(msgs))
	}
	if len(get(model, "consts").arr) != 2 {
		t.Errorf("designer lost consts: %s", StringifyJson(get(model, "consts")))
	}
	types := get(model, "types")
	if len(types.arr) != 1 || !strings.Contains(StringifyJson(types), `"readonly":true`) {
		t.Errorf("designer lost readonly fields: %s", StringifyJson(types))
	}
	bad := []struct{ name, src, want string }{
		{"assign const", `const LIMIT = 3
fn run() -> null { LIMIT = 4 }
view main { button "x" onClick: run }`, "cannot assign to const LIMIT"},
		{"mutate const", `const ITEMS: array[string] = ["a"]
fn run() -> null { ITEMS[0] = "b" }
view main { button "x" onClick: run }`, "cannot mutate const ITEMS"},
		{"readonly field", `type J struct { readonly id: string }
state j: J = J { id: "a" }
fn run() -> null { j.id = "b" }
view main { button "x" onClick: run }`, `cannot assign to readonly field "id"`},
		{"const runtime value", `state n = 1
const C = n + 1
view main { p "{str(C)}" }`, "must be initialized with a compile-time value"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "const.sui")
			if d.HasErrors() {
				t.Fatalf("parse: %v", d.String())
			}
			d = Check(p)
			if !d.HasErrors() || !strings.Contains(d.String(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, d.String())
			}
		})
	}
}

func TestGenericFunctionsAndComponents(t *testing.T) {
	src := `fn identity<T>(value: T) -> T { return value }
fn first<T>(items: array[T]) -> T { return items[0] }
fn both<A, B>(a: A, b: B) -> array[A | B] { return [a, b] }
view List<T> (items: array[T]) {
	ul {
		list item in items key: item {
			li "{str(item)}"
		}
	}
}
state names: array[string] = ["a", "b"]
state out = ""
fn run() -> null {
	out = identity("x") + " " + str(identity(2)) + " " + str(first([7, 8])) + " " + str(both("a", 1))
}
view main {
	List items: names
	p "{out}"
	button "Run" onClick: run
}`
	prog, dl := ParseSource(src, "generic.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "function identity(value)") || !strings.Contains(js, "function first(items)") {
		t.Errorf("generic functions should erase type parameters in JavaScript")
	}
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen: %s", StringifyJson(msgs))
	}
	var identity string
	for _, f := range get(model, "funcs").arr {
		if asStr(get(f, "name")) == "identity" {
			identity = asStr(get(f, "typeParams"))
		}
	}
	if identity != "T" {
		t.Errorf("designer lost function type parameters: %q", identity)
	}
	comps := StringifyJson(get(model, "components"))
	if !strings.Contains(comps, "view List<T>") {
		t.Errorf("designer lost component type parameters: %s", comps)
	}

	bad := []struct{ name, src, want string }{
		{"conflicting inference", `fn both<T>(a: T, b: T) -> T { return a }
fn run() -> null { let x = both("a", 1) }
view main { button "x" onClick: run }`, "must be string, got int"},
		{"unbound parameter", `fn make<T>() -> T { return null }
fn run() -> null { let x = make() }
view main { button "x" onClick: run }`, "cannot infer generic type parameter T"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "generic.sui")
			if d.HasErrors() {
				t.Fatalf("parse: %v", d.String())
			}
			d = Check(p)
			if !d.HasErrors() || !strings.Contains(d.String(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, d.String())
			}
		})
	}
}

func TestFeatureFormatting(t *testing.T) {
	src := `const LIMIT:int=3
type Job struct{readonly id:string,mode:string}
computed total:int=LIMIT*2
fn identity<T>(value:T)->T{return value}
fn addAll(base:int,...values:array[int])->int{return base}
view List<T> (items:array[T]) { ul { list item in items key:item { li "{str(item)}" } } }
state items:array[string]=["a"]
view main {
	List items:items
	p "{identity(str(total))}"
}`
	out, dl := FormatSource(src, "features.sui")
	if dl != nil && dl.HasErrors() {
		t.Fatalf("fmt errors: %v", dl)
	}
	prog, pdl := ParseSource(out, "features.sui")
	if pdl.HasErrors() {
		t.Fatalf("formatted source does not parse: %v\n%s", pdl, out)
	}
	if cdl := Check(prog); cdl.HasErrors() {
		t.Fatalf("formatted source does not check: %v\n%s", cdl, out)
	}
	out2, _ := FormatSource(out, "features.sui")
	if out != out2 {
		t.Errorf("fmt is not idempotent for generic/const/computed/rest source")
	}
	for _, want := range []string{"const LIMIT: int = 3", "computed total: int = LIMIT * 2", "fn identity<T>", "...values: array[int]", "readonly id: string"} {
		if !strings.Contains(out, want) {
			t.Errorf("formatted source lost %q", want)
		}
	}
}

func TestComputedValuesAndBatching(t *testing.T) {
	src := `state price: int = 10
state qty: int = 2
computed total: int = price * qty
computed label = "total " + str(total)
fn inc() -> null {
	qty = qty + 1
}
view main {
	p "{str(price)} x {str(qty)} = {str(total)} ({label})"
	button "Inc" onClick: inc
}`
	prog, dl := ParseSource(src, "computed.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, marker := range []string{"__SUI_COMPUTED.push", "__sui_schedule", "__sui_recompute"} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}
	model, msgs, ok := guiOpen(src)
	if !ok {
		t.Fatalf("guiOpen: %s", StringifyJson(msgs))
	}
	if len(get(model, "computed").arr) != 2 {
		t.Errorf("designer lost computed values: %s", StringifyJson(get(model, "computed")))
	}
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    var text = document.body.innerText;
    if (text.indexOf('10 x 2 = 20 (total 20)') < 0) throw new Error('initial: ' + text);
    button('Inc').click(); button('Inc').click();
    await wait(200);
    text = document.body.innerText;
    if (text.indexOf('10 x 4 = 40 (total 40)') < 0) throw new Error('batched: ' + text);
`+featureScriptTail)
	p, d := ParseSource(`computed total = 1
fn run() -> null { total = 2 }
view main { button "x" onClick: run }`, "computed.sui")
	if d.HasErrors() {
		t.Fatalf("parse: %v", d.String())
	}
	d = Check(p)
	if !d.HasErrors() || !strings.Contains(d.String(), "cannot assign to computed total") {
		t.Fatalf("expected computed assignment error, got %v", d.String())
	}
}

func TestErrorBoundary(t *testing.T) {
	src := `fn broken() -> string {
	assert(false, "boom")
	return "unreachable"
}
view main {
	errorboundary fallback: "Recovered in boundary" {
		p "{broken()}"
	}
	p "after boundary"
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    var boundary = document.querySelector('.sui-error');
    var text = document.body.innerText;
    if (!boundary || boundary.textContent.indexOf('Recovered in boundary') < 0) throw new Error('fallback missing: ' + text);
    if (text.indexOf('after boundary') < 0) throw new Error('content after boundary missing: ' + text);
    if ((boundary.getAttribute('data-sui-error') || '').indexOf('boom') < 0) throw new Error('error detail missing');
`+featureScriptTail)
}

func TestListKeys(t *testing.T) {
	src := `state items: array[string] = ["a", "b"]
view main {
	ul {
		list item in items key: item {
			li "{item}"
		}
	}
}`
	prog, dl := ParseSource(src, "keys.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "__sui_keyed") || !strings.Contains(js, "data-sui-key") {
		t.Errorf("list key was not generated")
	}
	p, d := ParseSource(`state items: array[string] = ["a"]
view main {
	list item in items key: [item] {
		p "{item}"
	}
}`, "keys.sui")
	if d.HasErrors() {
		t.Fatalf("parse: %v", d.String())
	}
	d = Check(p)
	if !d.HasErrors() || !strings.Contains(d.String(), "list key must be") {
		t.Fatalf("expected list key type error, got %v", d.String())
	}
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(250);
    var items = document.querySelectorAll('li[data-sui-key]');
    if (items.length !== 2) throw new Error('keyed items: ' + items.length);
    if (items[0].getAttribute('data-sui-key') !== 'a' || items[1].getAttribute('data-sui-key') !== 'b') throw new Error('wrong keys');
`+featureScriptTail)
}

func TestIntlFormatting(t *testing.T) {
	src := `state out = ""
fn run() -> null {
	out = intl.number(1234.567, 2, "en-US") + "|" + intl.currency(1234.5, "EUR", "de-DE") + "|" + intl.percent(0.256, 1, "en-US") + "|" + intl.relative(-2, "day", "en-US") + "|" + intl.timezone()
}
view main {
	p "{out}"
	button "Run" onClick: run
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Run').click(); await wait(200);
    var text = document.body.innerText;
    if (text.indexOf('1,234.57') < 0) throw new Error('number: ' + text);
    if (text.indexOf('1.234,50') < 0) throw new Error('currency: ' + text);
    if (text.indexOf('25.6%') < 0) throw new Error('percent: ' + text);
    if (text.indexOf('2 days ago') < 0) throw new Error('relative: ' + text);
    if (!/[A-Za-z]+\/[A-Za-z_]+/.test(text)) throw new Error('timezone missing: ' + text);
`+featureScriptTail)

	bad := []struct{ name, src, want string }{
		{"number arg", `fn run() -> null { let x = intl.number("nope") }
view main { button "x" onClick: run }`, "intl.number requires a number"},
		{"relative unit", `fn run() -> null { let x = intl.relative(1, 2) }
view main { button "x" onClick: run }`, "intl.relative unit must be a string"},
		{"timezone args", `fn run() -> null { let x = intl.timezone(1) }
view main { button "x" onClick: run }`, "intl.timezone takes exactly 0 arguments"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "intl.sui")
			if d.HasErrors() {
				t.Fatalf("parse: %v", d.String())
			}
			d = Check(p)
			if !d.HasErrors() || !strings.Contains(d.String(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, d.String())
			}
		})
	}
}
