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

func TestMotionPagesAttrs(t *testing.T) {
	src := `fn openDetail() -> null {
	navigate("detail")
}

view home {
	card motion: "layout" motionKey: "hero" {
		p "Hero"
	}
	button "Open" onClick: openDetail
}

view detail {
	card motion: "layout" motionKey: "hero" {
		p "Hero"
	}
}`
	prog, dl := ParseSource(src, "pages.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"function __sui_page(n)",
		"__SUI_MOTION.capturePage",
		"__SUI_MOTION.flushShared",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("shared page JS missing %q", want)
		}
	}
	if strings.Count(js, `data-sui-motion-key="hero"`) < 2 {
		t.Error("expected the same shared key on both pages")
	}
}

func TestMotionPagesE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping shared page motion test")
	}
	const source = `fn openDetail() -> null {
	navigate("detail")
}

fn backHome() -> null {
	navigate("home")
}

css ".hero \{ background: #202733; border: 1px solid #3a4152; border-radius: 12px; padding: 14px; \} .hero-home \{ width: 280px; height: 140px; \} .hero-detail \{ width: 640px; height: 260px; \}"

view home {
	card class: "hero hero-home" motion: "layout" motionKey: "hero" motionDuration: "400ms" motionEasing: "linear" {
		p "Hero"
	}
	button "Open" onClick: openDetail
}

view detail {
	button "Back" onClick: backHome
	card class: "hero hero-detail" motion: "layout" motionKey: "hero" motionDuration: "400ms" motionEasing: "linear" {
		p "Hero"
	}
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + motionPagesScript + page[idx:]
	} else {
		t.Fatal("compiled page has no closing body")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	mux := http.NewServeMux()
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
		"--no-first-run", "--no-default-browser-check", "--window-size=1000,700",
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
	t.Fatalf("shared page motion e2e failed:\n%s", detail)
}

const motionPagesScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.appendChild(el); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function button(label) {
    var list = document.querySelectorAll('button');
    for (var i = 0; i < list.length; i++) if (list[i].textContent.trim() === label) return list[i];
    return null;
  }
  function hero() {
    var list = document.querySelectorAll('[data-sui-motion-key="hero"]');
    for (var i = 0; i < list.length; i++) {
      if (list[i].offsetParent !== null) {
        var r = list[i].getBoundingClientRect();
        return { w: Math.round(r.width), h: Math.round(r.height), left: Math.round(r.left), top: Math.round(r.top), transform: getComputedStyle(list[i]).transform, page: window.__SUI_PAGE };
      }
    }
    return null;
  }
  try {
    for (var i = 0; i < 100; i++) {
      if (hero()) break;
      await wait(100);
    }
    var home = hero();
    if (!home || home.w !== 280 || home.h !== 140) throw new Error('unexpected home hero: ' + JSON.stringify(home));

    button('Open').click();
    await wait(120);
    var toDetail = hero();
    if (!toDetail || toDetail.page !== 'detail') throw new Error('did not navigate to detail');
    if (toDetail.transform === 'none') throw new Error('shared element did not FLIP');
    if (!(toDetail.w > home.w && toDetail.w < 640)) throw new Error('shared width did not tween: ' + JSON.stringify(toDetail));
    await wait(500);
    var detail = hero();
    if (!detail || detail.w !== 640 || detail.h !== 260 || detail.transform !== 'none') throw new Error('detail did not settle: ' + JSON.stringify(detail));

    button('Back').click();
    await wait(120);
    var toHome = hero();
    if (!toHome || toHome.page !== 'home') throw new Error('did not navigate home');
    if (toHome.transform === 'none') throw new Error('reverse shared element did not FLIP');
    if (!(toHome.w > 280 && toHome.w < 640)) throw new Error('reverse shared width did not tween: ' + JSON.stringify(toHome));
    await wait(500);
    var back = hero();
    if (!back || back.w !== 280 || back.h !== 140 || back.transform !== 'none') throw new Error('home did not settle: ' + JSON.stringify(back));
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
