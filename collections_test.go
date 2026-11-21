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

const collectionSource = `state items = ["a", "bb", "ccc"]
state nums = [1, 2, 3, 4]
state dup = ["a", "a", "b"]
state comma = ","

fn isLong(s: string) -> bool { return len(s) > 1 }
fn upperIt(s: string) -> string { return upper(s) }
fn add(a: int, b: int) -> int { return a + b }
fn key(s: string) -> int { return len(s) }
fn even(n: int) -> bool { return n % 2 == 0 }

view main {
	p "filter: {join(filter(items, isLong), comma)}"
	p "map: {join(map(items, upperIt), comma)}"
	p "reduce: {str(reduce(nums, 0, add))}"
	p "find: {str(find(items, isLong))}"
	p "findIndex: {str(findIndex(nums, even))}"
	p "any: {str(any(nums, even))} all: {str(all(nums, even))}"
	p "sortBy: {join(sortBy(items, key), comma)}"
	p "unique: {join(unique(dup), comma)}"
}
`

// TestCollectionBuiltinsTypecheck guards the callback contract and the
// generated helper calls.
func TestCollectionBuiltinsTypecheck(t *testing.T) {
	prog, dl := ParseSource(collectionSource, "collections.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, marker := range []string{
		"function __sui_map(arr, fn)",
		"function __sui_filter(arr, fn)",
		"function __sui_reduce(arr, init, fn)",
		"__sui_filter(items, isLong)",
		"__sui_map(items, upperIt)",
		"__sui_reduce(nums, 0, add)",
		"__sui_find(items, isLong)",
		"__sui_find_index(nums, even)",
		"__sui_any(nums, even)",
		"__sui_all(nums, even)",
		"__sui_sort_by(items, key)",
		"__sui_unique(dup)",
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}
}

func TestCollectionBuiltinsRejectBadCallbacks(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name: "non function callback",
			src: `state items = [1, 2]
fn run() -> null {
	let x = filter(items, 3)
}
view main { button "Go" onClick: run }`,
			wantErr: "expects the name of a function",
		},
		{
			name: "async callback",
			src: `state items = [1, 2]
async fn check(n: int) -> bool { return n > 0 }
fn run() -> null {
	let x = filter(items, check)
}
view main { button "Go" onClick: run }`,
			wantErr: "must be synchronous",
		},
		{
			name: "filter callback return",
			src: `state items = [1, 2]
fn label(n: int) -> string { return str(n) }
fn run() -> null {
	let x = filter(items, label)
}
view main { button "Go" onClick: run }`,
			wantErr: "callback must return bool",
		},
		{
			name: "reduce return",
			src: `state items = [1, 2]
fn merge(a: int, n: int) -> string { return str(a) + str(n) }
fn run() -> null {
	let x = reduce(items, 0, merge)
}
view main { button "Go" onClick: run }`,
			wantErr: "reduce callback must return",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "collections.sui")
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

// TestCollectionBuiltinsRuntimeE2E exercises the helpers in a real browser so
// the callback invocation and result ordering are covered end to end.
func TestCollectionBuiltinsRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping collection runtime test")
	}
	index, css, js, diags, ok := guiBuildFiles(collectionSource)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + collectionScript + page[idx:]
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
		end := marker + 1600
		if end > len(text) {
			end = len(text)
		}
		detail = text[marker:end]
	}
	t.Fatalf("collection runtime e2e failed:\n%s", detail)
}

const collectionScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function assert(ok, message) { if (!ok) throw new Error(message); }
  try {
    await new Promise(function (r) { setTimeout(r, 500); });
    var text = document.body.innerText;
    assert(text.indexOf('filter: bb,ccc') >= 0, 'filter failed: ' + text);
    assert(text.indexOf('map: A,BB,CCC') >= 0, 'map failed: ' + text);
    assert(text.indexOf('reduce: 10') >= 0, 'reduce failed: ' + text);
    assert(text.indexOf('find: bb') >= 0, 'find failed: ' + text);
    assert(text.indexOf('findIndex: 1') >= 0, 'findIndex failed: ' + text);
    assert(text.indexOf('any: true all: false') >= 0, 'any/all failed: ' + text);
    assert(text.indexOf('sortBy: a,bb,ccc') >= 0, 'sortBy failed: ' + text);
    assert(text.indexOf('unique: a,b') >= 0, 'unique failed: ' + text);
    result(true);
  } catch (err) {
    result(false, err && err.message ? err.message : String(err));
  }
})();
</script>`
