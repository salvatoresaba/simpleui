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

const i18nSource = `i18n "it" {
	"save": "Salva"
	"items": { one: "{count} elemento", other: "{count} elementi" }
	"onlyIt": "Solo italiano"
}
i18n "en" {
	"save": "Save"
	"items": { one: "{count} item", other: "{count} items" }
}
i18n "ar" {
	"save": "حفظ"
	"items": { zero: "zero", one: "one", two: "two", few: "few", many: "many", other: "other" }
}
state count = 1
state keySave = "save"
state comma = ","

fn itemText(n: int) -> string { return t("items", {"count": n}) }
fn setIt() -> null { i18n.setLocale("it") }
fn setEn() -> null { i18n.setLocale("en") }
fn setAr() -> null { i18n.setLocale("ar") }
fn inc() -> null { count += 1 }

view main {
	p "locale: {i18n.locale()}"
	p "save: {t(keySave)}"
	p "items: {itemText(count)}"
	p "dir: {i18n.dir()}"
	p "available: {join(i18n.available(), comma)}"
	p "has: {str(i18n.has(keySave))}"
	p "missing: {join(i18n.missing(), comma)}"
	button "Italiano" onClick: setIt
	button "English" onClick: setEn
	button "Arabic" onClick: setAr
	button "+1" onClick: inc
}
`

func TestI18nParseAndValidate(t *testing.T) {
	prog, dl := ParseSource(i18nSource, "i18n.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if len(prog.i18n) != 3 {
		t.Fatalf("expected 3 catalogs, got %d", len(prog.i18n))
	}
	if prog.i18n[0].locale != "it" || prog.i18n[1].locale != "en" {
		t.Fatalf("catalog order not preserved")
	}
	if prog.i18n[0].messages[1].plural["one"] != "{count} elemento" {
		t.Fatalf("plural message was not parsed: %#v", prog.i18n[0].messages[1])
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, marker := range []string{
		`__sui_i18n_init({`,
		`"it":{"save":"Salva"`,
		`"items":{"one":"{count} elemento","other":"{count} elementi"}`,
		`"en":{"save":"Save"`,
		`}, "it");`,
		`__sui_t(keySave, undefined)`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}
}

func TestI18nValidationErrors(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name: "duplicate locale",
			src: `i18n "it" { "a": "A" }
i18n "it" { "b": "B" }
view main { p "x" }`,
			wantErr: "duplicate i18n locale",
		},
		{
			name: "duplicate key",
			src: `i18n "it" {
	"a": "A"
	"a": "B"
}
view main { p "x" }`,
			wantErr: "duplicate i18n message key",
		},
		{
			name: "missing other",
			src: `i18n "it" {
	"a": { one: "uno" }
}
view main { p "x" }`,
			wantErr: `must define the "other" category`,
		},
		{
			name: "unknown category",
			src: `i18n "it" {
	"a": { one: "uno", other: "molti", lots: "tanti" }
}
view main { p "x" }`,
			wantErr: "unknown plural category",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "i18n.sui")
			if dl.HasErrors() {
				if strings.Contains(dl.String(), tc.wantErr) {
					return
				}
				t.Fatalf("parse: %v", dl.String())
			}
			dl = Check(prog)
			if !dl.HasErrors() || !strings.Contains(dl.String(), tc.wantErr) {
				t.Fatalf("expected %q, got: %v", tc.wantErr, dl.String())
			}
		})
	}
}

func TestI18nRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping i18n runtime test")
	}
	index, css, js, diags, ok := guiBuildFiles(i18nSource)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + i18nScript + page[idx:]
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
		end := marker + 2000
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("i18n runtime e2e failed:\n%s", detail)
}

const i18nScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function assert(ok, message) { if (!ok) throw new Error(message); }
  function button(label) {
    var all = Array.prototype.slice.call(document.querySelectorAll('button'));
    for (var i = 0; i < all.length; i++) if (all[i].textContent === label) return all[i];
    return null;
  }
  try {
    await wait(500);
    var text = document.body.innerText;
    assert(text.indexOf('locale: it') >= 0, 'default locale is not it: ' + text);
    assert(text.indexOf('save: Salva') >= 0, 'italian message missing: ' + text);
    assert(text.indexOf('items: 1 elemento') >= 0, 'italian singular missing: ' + text);
    assert(text.indexOf('available: it,en,ar') >= 0, 'available locales missing: ' + text);
    assert(text.indexOf('has: true') >= 0, 'has() failed: ' + text);
    assert(text.indexOf('missing:') >= 0, 'missing() failed: ' + text);

    button('+1').click(); await wait(300);
    text = document.body.innerText;
    assert(text.indexOf('items: 2 elementi') >= 0, 'italian plural missing: ' + text);

    button('English').click(); await wait(400);
    text = document.body.innerText;
    assert(text.indexOf('locale: en') >= 0, 'english locale missing: ' + text);
    assert(text.indexOf('save: Save') >= 0, 'english message missing: ' + text);
    assert(text.indexOf('items: 2 items') >= 0, 'english plural missing: ' + text);
    assert(text.indexOf('missing: onlyIt') >= 0, 'english missing() should report onlyIt: ' + text);

    button('Arabic').click(); await wait(400);
    text = document.body.innerText;
    assert(text.indexOf('locale: ar') >= 0, 'arabic locale missing: ' + text);
    assert(text.indexOf('save: حفظ') >= 0, 'arabic message missing: ' + text);
    assert(text.indexOf('dir: rtl') >= 0, 'arabic direction missing: ' + text);
    assert(text.indexOf('items: two') >= 0, 'arabic two-category missing: ' + text);
    button('+1').click(); await wait(300);
    text = document.body.innerText;
    assert(text.indexOf('items: few') >= 0, 'arabic few-category missing: ' + text);
    assert(text.indexOf('missing: onlyIt') >= 0, 'arabic missing() is wrong: ' + text);
    assert(document.documentElement.dir === 'rtl', 'document dir was not set to rtl');
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`

// TestOptionalBuiltinArguments guards the codegen padding that keeps optional
// builtin arguments (time.format(dt), t(key), ...) valid JavaScript.
func TestOptionalBuiltinArguments(t *testing.T) {
	prog, dl := ParseSource(`view main { p "{time.format(time.now())}" }`, "optional.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "__sui_time_format(__sui_time_now(), undefined)") {
		t.Errorf("optional time.format argument was not padded with undefined")
	}
	if strings.Contains(js, "{1}") {
		t.Errorf("generated JS still contains an unsubstituted placeholder")
	}
}

// TestI18nLocaleQueryParam checks the ?locale= override used by the designer
// preview without touching localStorage.
func TestI18nLocaleQueryParam(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping i18n query test")
	}
	prog, dl := ParseSource(i18nSource, "i18n.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	page := CombineSingleHTML(GenerateIndexHtml(prog), GenerateCss(prog), GenerateJs(prog), false)
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
	url := "http://" + ln.Addr().String() + "/?locale=en"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--no-first-run", "--no-default-browser-check", "--window-size=900,700",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=8000", "--dump-dom", url,
	}
	out, err := exec.CommandContext(ctx, browser, args...).CombinedOutput()
	text := string(out)
	if err != nil {
		t.Fatalf("headless browser failed: %v\n%s", err, text)
	}
	if !strings.Contains(text, "locale: en") || !strings.Contains(text, "save: Save") {
		t.Fatalf("?locale=en did not select English:\n%s", text)
	}
}
