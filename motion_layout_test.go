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

func TestMotionLayoutAttrs(t *testing.T) {
	src := `state items = ["a", "b"]

view main {
	card motion: "layout" motionKey: "static-a" motionDuration: "420ms" motionEasing: "ease-out" {
		p "A"
	}
	card motion: "fade" motionKey: "static-b" motionExit: false {
		p "B"
	}
	list item in items {
		card motion: "layout" motionKey: item {
			p "{item}"
		}
	}
}`
	prog, dl := ParseSource(src, "motion.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		`data-sui-motion`,
		`data-sui-motion-key="static-a"`,
		`data-sui-motion-key="static-b"`,
		`data-sui-motion-effect="layout"`,
		`data-sui-motion-effect="fade"`,
		`data-sui-motion-duration="420ms"`,
		`data-sui-motion-easing="ease-out"`,
		`data-sui-motion-key="${__sui_esc(item)}"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("motion layout JS missing %q", want)
		}
	}
	// static-b opts out of the exit clone.
	if strings.Count(js, `data-sui-motion-exit="1"`) < 2 {
		t.Error("expected exit attrs for motion elements without motionExit false")
	}
}

func TestMotionLayoutE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion layout test")
	}
	const source = `state showB = true
state order = 0

fn toggleB() -> null {
	showB = showB == false
}

fn flipOrder() -> null {
	order = 1 - order
}

view main {
	row {
		button "Toggle B" onClick: toggleB
		button "Flip order" onClick: flipOrder
	}
	col class: "cards" {
		if order == 0 {
			if showB {
				card motion: "layout" motionKey: "b" motionDuration: "400ms" motionEasing: "ease-out" {
					h3 "B"
					p "Second card"
				}
			}
			card motion: "layout" motionKey: "a" motionDuration: "400ms" motionEasing: "ease-out" {
				h3 "A"
				p "First card"
			}
		} else {
			card motion: "layout" motionKey: "a" motionDuration: "400ms" motionEasing: "ease-out" {
				h3 "A"
				p "First card"
			}
			if showB {
				card motion: "layout" motionKey: "b" motionDuration: "400ms" motionEasing: "ease-out" {
					h3 "B"
					p "Second card"
				}
			}
		}
	}
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + motionLayoutScript + page[idx:]
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
	url := "http://" + ln.Addr().String() + "/?motion=timer"
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows", "--disable-renderer-backgrounding",
		"--no-first-run", "--no-default-browser-check", "--window-size=800,600",
		"--user-data-dir=" + t.TempDir(), "--virtual-time-budget=20000", "--dump-dom", url,
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
	t.Fatalf("motion layout e2e failed:\n%s", detail)
}

const motionLayoutScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function button(label) {
    var list = document.querySelectorAll('button');
    for (var i = 0; i < list.length; i++) if (list[i].textContent.trim() === label) return list[i];
    return null;
  }
  function cards() {
    var out = [];
    var list = document.querySelectorAll('[data-sui-motion-key]');
    for (var i = 0; i < list.length; i++) {
      var r = list[i].getBoundingClientRect();
      out.push({ key: list[i].getAttribute('data-sui-motion-key'), top: Math.round(r.top), opacity: parseFloat(getComputedStyle(list[i]).opacity), transform: getComputedStyle(list[i]).transform });
    }
    return out;
  }
  try {
    for (var i = 0; i < 100; i++) {
      if (document.querySelector('[data-sui-motion-key="a"]')) break;
      await wait(100);
    }
    var initial = cards();
    if (initial.length !== 2 || initial[0].key !== 'b' || initial[1].key !== 'a') throw new Error('unexpected initial cards: ' + JSON.stringify(initial));

    button('Toggle B').click();
    await wait(100);
    var mid = cards();
    var clones = document.querySelectorAll('.sui-motion-exit').length;
    if (clones !== 1) throw new Error('exit clone missing: ' + clones);
    if (mid.length !== 1 || mid[0].key !== 'a' || mid[0].transform === 'none') throw new Error('remaining card did not FLIP: ' + JSON.stringify(mid));
    await wait(450);
    var afterHide = cards();
    if (afterHide.length !== 1 || afterHide[0].key !== 'a' || Math.abs(afterHide[0].top - initial[0].top) > 40) throw new Error('card did not settle after exit: ' + JSON.stringify(afterHide));
    if (document.querySelectorAll('.sui-motion-exit').length !== 0) throw new Error('exit clone was not removed');

    button('Toggle B').click();
    await wait(100);
    var entering = cards();
    var entered = null;
    for (var e = 0; e < entering.length; e++) if (entering[e].key === 'b' && entering[e].opacity < 1) entered = entering[e];
    if (!entered) throw new Error('entering card did not fade in: ' + JSON.stringify(entering));
    await wait(450);

    button('Flip order').click();
    await wait(100);
    var flipping = cards();
    if (flipping.length !== 2) throw new Error('cards missing during flip');
    var moving = 0;
    for (var f = 0; f < flipping.length; f++) if (flipping[f].transform !== 'none') moving++;
    if (moving !== 2) throw new Error('FLIP did not move both cards: ' + JSON.stringify(flipping));
    await wait(450);
    var flipped = cards();
    if (flipped[0].key !== 'a' || flipped[1].key !== 'b') throw new Error('order did not change: ' + JSON.stringify(flipped));
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
