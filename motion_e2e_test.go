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

func TestMotionRuntimeE2E(t *testing.T) {
	browser := findHeadlessBrowser()
	if browser == "" {
		t.Skip("no Chrome/Edge found; skipping motion runtime test")
	}
	const source = `state on = true
state shape = "pill"
state theme = "accent"
state effect = "morph"

fn toggle() -> null {
	on = on == false
}

fn setMaterial() -> null {
	shape = "material"
}

fn setSuccess() -> null {
	theme = "success"
}

fn setFade() -> null {
	effect = "fade"
}

fn setScale() -> null {
	effect = "scale"
}

fn setMorph() -> null {
	effect = "morph"
}

view main {
	switch "Modalità" shape: shape theme: theme checked: on animation: effect duration: "400ms" easing: "linear" onChange: toggle
	checkbox "Accetta" checked: on animation: "morph" duration: "400ms" easing: "linear" onChange: toggle
	button "Material" onClick: setMaterial
	button "Success" onClick: setSuccess
	button "Fade" onClick: setFade
	button "Scale" onClick: setScale
	button "Morph" onClick: setMorph
	p "on {on} shape {shape} theme {theme} effect {effect}"
}`
	index, css, js, diags, ok := guiBuildFiles(source)
	if !ok {
		t.Fatalf("build files: %s", StringifyJson(diags))
	}
	page := CombineSingleHTML(index, css, js, false)
	if idx := strings.LastIndex(page, "</body>"); idx >= 0 {
		page = page[:idx] + motionRuntimeScript + page[idx:]
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
	t.Fatalf("motion runtime e2e failed:\n%s", detail)
}

const motionRuntimeScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) { el = document.createElement('div'); el.id = 'e2e-result'; document.body.insertBefore(el, document.body.firstChild); }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function wait(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
  function press(el) { el.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, cancelable: true })); }
  function num(v) { var n = parseFloat(v); return isNaN(n) ? null : n; }
  function button(label) {
    var list = document.querySelectorAll('button');
    for (var i = 0; i < list.length; i++) if (list[i].textContent.trim() === label) return list[i];
    return null;
  }
  try {
    for (var i = 0; i < 100; i++) {
      if (document.querySelector('.sui-switch') && document.querySelector('.sui-checkbox')) break;
      await wait(100);
    }
    var sw = document.querySelector('.sui-switch');
    if (!sw) throw new Error('switch missing');
    var beforeX = num(getComputedStyle(sw).getPropertyValue('--sui-sw-thumb-x'));
    press(sw);
    sw.click();
    await wait(120);
    sw = document.querySelector('.sui-switch');
    var liveX = num(sw.style.getPropertyValue('--sui-sw-thumb-x'));
    if (liveX == null) throw new Error('switch thumb position did not tween');
    if (beforeX != null && Math.abs(liveX - beforeX) < 0.5) throw new Error('switch thumb did not move during tween: ' + beforeX + ' -> ' + liveX);
    await wait(420);
    sw = document.querySelector('.sui-switch');
    if (sw.style.getPropertyValue('--sui-sw-thumb-x') !== '') throw new Error('switch inline vars were not cleared: style=' + sw.getAttribute('style') + ' raf=' + (sw.__suiMotionRaf || 0) + ' count=' + document.querySelectorAll('.sui-switch').length);

    var mat = null;
    var buttons = document.querySelectorAll('button');
    for (var b = 0; b < buttons.length; b++) if (buttons[b].textContent.trim() === 'Material') mat = buttons[b];
    if (!mat) throw new Error('material button missing');
    var beforeW = num(getComputedStyle(sw).getPropertyValue('--sui-sw-w'));
    mat.click();
    await wait(120);
    sw = document.querySelector('.sui-switch');
    var liveW = num(sw.style.getPropertyValue('--sui-sw-w'));
    if (liveW == null) throw new Error('switch shape did not tween');
    if (beforeW != null && Math.abs(liveW - beforeW) < 0.5) throw new Error('switch width did not change during tween');
    await wait(420);
    sw = document.querySelector('.sui-switch');
    if (sw.className.indexOf('sui-sw-shape-material') < 0) throw new Error('switch shape class missing: ' + sw.className);

    press(sw);
    sw.click();
    await wait(450);
    sw = document.querySelector('.sui-switch');
    var success = null;
    buttons = document.querySelectorAll('button');
    for (var c = 0; c < buttons.length; c++) if (buttons[c].textContent.trim() === 'Success') success = buttons[c];
    if (!success) throw new Error('success button missing');
    success.click();
    await wait(120);
    sw = document.querySelector('.sui-switch');
    if (sw.style.getPropertyValue('--sui-sw-color') === '') throw new Error('switch theme color did not tween');
    await wait(420);
    sw = document.querySelector('.sui-switch');
    if (sw.className.indexOf('sui-sw-theme-success') < 0) throw new Error('switch theme class missing: ' + sw.className);

    var cb = document.querySelector('.sui-checkbox');
    if (!cb) throw new Error('checkbox missing');
    var beforeOpacity = num(getComputedStyle(cb).getPropertyValue('--sui-cb-mark-opacity'));
    press(cb);
    cb.click();
    await wait(120);
    cb = document.querySelector('.sui-checkbox');
    var liveOpacity = num(cb.style.getPropertyValue('--sui-cb-mark-opacity'));
    if (liveOpacity == null) throw new Error('checkbox mark did not tween');
    if (beforeOpacity != null && Math.abs(liveOpacity - beforeOpacity) < 0.05) throw new Error('checkbox opacity did not change during tween');
    await wait(420);
    cb = document.querySelector('.sui-checkbox');
    if (cb.style.getPropertyValue('--sui-cb-mark-opacity') !== '') throw new Error('checkbox inline vars were not cleared');

    async function sampleEffect(label, want) {
      button(label).click();
      await wait(80);
      var s = document.querySelector('.sui-switch');
      var attr = s.getAttribute('data-sui-motion-effect');
      s.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, cancelable: true }));
      s.click();
      await wait(100);
      s = document.querySelector('.sui-switch');
      var cs = getComputedStyle(s);
      var res = { attr: attr, opacity: parseFloat(cs.opacity), transform: cs.transform };
      await wait(420);
      return res;
    }
    var fade = await sampleEffect('Fade', 'fade');
    if (fade.attr !== 'fade' || fade.opacity <= 0.35 || fade.opacity >= 1) throw new Error('fade effect missing: ' + JSON.stringify(fade));
    var scale = await sampleEffect('Scale', 'scale');
    if (scale.attr !== 'scale' || scale.opacity < 0.99 || scale.transform === 'none') throw new Error('scale effect missing: ' + JSON.stringify(scale));
    var morph = await sampleEffect('Morph', 'morph');
    if (morph.attr !== 'morph' || morph.opacity < 0.99 || morph.transform !== 'none') throw new Error('morph effect changed opacity/transform: ' + JSON.stringify(morph));
    result(true);
  } catch (e) {
    result(false, e && e.message ? e.message : String(e));
  }
})();
</script>`
