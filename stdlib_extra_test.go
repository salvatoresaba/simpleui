package sui

import (
	"strings"
	"testing"
)

const stdExtraTestSource = `import { bytes, bin, files, mime, color, units, print, history, zip, archive, csv, env, capabilities, events, bus, log, telemetry } from "std"

state out = ""
state payload = ""

fn onEvent(value: any) -> null {
	payload = str(value)
}

async fn run() -> null {
	raw, err1 = bytes.fromString("Ada", "utf-8")
	text, err2 = bytes.toString(raw, "utf-8")
	b64, err3 = bytes.toBase64(raw)
	binResult, binErr = bin.writeU16(258)
	n, err4 = bin.u16(binResult, 0)
	name, err5 = files.name("C:/tmp/report.csv")
	kind, err6 = mime.fromName("report.csv")
	parsed, err7 = color.parse("#336699")
	formatted, err8 = color.format(parsed, "hex")
	mm, err9 = units.convert(1.0, "in", "mm")
	paper, err10 = print.paper("A4", "landscape")
	hs, err11 = history.push("doc", {"n": 1})
	arc, err12 = zip.create()
	added, err13 = zip.addText(arc, "hello.txt", "world")
	packed, err14 = zip.finish(arc)
	opened, err15 = zip.open(packed)
	entries, err16 = zip.list(opened)
	unpacked, err17 = archive.unpackText(packed, "zip")
	rows, err18 = csv.parse("a,b\n1,2\n")
	serialized, err19 = csv.stringifyObjects([{"name": "Ada", "formula": "=1+1"}], {})
	target, err20 = env.target()
	caps, err21 = capabilities.list()
	listeners, err22 = events.on("demo", onEvent)
	emitted, err23 = bus.emit("demo", "ok")
	log.info("hello", {"token": "secret"})
	counter, err24 = telemetry.counter("requests")

	out = str(len(raw)) + "|" + text + "|" + b64 + "|" + str(n) + "|" + name + "|" + kind + "|" + formatted
	out = out + "|" + str(mm) + "|" + str(paper["width"]) + "|" + str(hs) + "|" + str(added) + "|" + str(len(entries))
	out = out + "|" + unpacked["hello.txt"] + "|" + str(len(rows)) + "|" + serialized + "|" + target
	out = out + "|" + str(len(caps)) + "|" + str(listeners) + "|" + str(emitted) + "|" + str(counter) + "|" + payload
}

view main {
	p "{out}"
	button "Run" onClick: run
}`

func TestStdlibExtraRegistry(t *testing.T) {
	for _, mod := range []string{
		"bytes", "bin", "files", "mime", "color", "units", "print", "history",
		"zip", "archive", "csv", "env", "capabilities", "events", "bus", "log", "telemetry",
	} {
		if !IsStdModule(mod) {
			t.Errorf("module %q is not importable", mod)
		}
		if !isStdlibExtraModule(mod) {
			t.Errorf("module %q is not in the extra runtime set", mod)
		}
	}
	names := map[string]bool{}
	keys := map[string]bool{}
	for i := 0; i < len(extraStdFuncs); i++ {
		f := &extraStdFuncs[i]
		full := f.mod + "." + f.name
		if names[full] {
			t.Errorf("duplicate std function %q", full)
		}
		if keys[f.jsKey] {
			t.Errorf("duplicate std key %q", f.jsKey)
		}
		names[full] = true
		keys[f.jsKey] = true
		if body := stdlibExtraBody(f.jsKey); !strings.Contains(body, "__SUI_STD_EXTRA."+f.mod+"."+f.name) {
			t.Errorf("missing generated body for %q: %q", full, body)
		}
	}
}

func TestStdlibExtraCompilation(t *testing.T) {
	prog, diags := ParseSource(stdExtraTestSource, "extra.sui")
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags.String())
	}
	if diags := Check(prog); diags.HasErrors() {
		t.Fatalf("check: %v", diags.String())
	}
	js := GenerateJsFor(prog, "web")
	for _, want := range []string{
		"var __SUI_STD_EXTRA = (function ()",
		"bytes_fromstring: function () { return __SUI_STD_EXTRA.bytes.fromString.apply(null, arguments); }",
		"telemetry_counter: function () { return __SUI_STD_EXTRA.telemetry.counter.apply(null, arguments); }",
		`TARGET = "web"`,
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("generated JS missing %q", want)
		}
	}
	desktop := GenerateJsFor(prog, "desktop")
	if !strings.Contains(desktop, `TARGET = "desktop"`) {
		t.Fatal("desktop target was not embedded in the runtime")
	}

	plain, plainDiags := ParseSource(`state out = ""
view main { p "{out}" }`, "plain.sui")
	if plainDiags.HasErrors() {
		t.Fatalf("parse plain: %v", plainDiags.String())
	}
	if js := GenerateJs(plain); strings.Contains(js, "__SUI_STD_EXTRA") {
		t.Fatal("extra runtime emitted for a program that imports no extra module")
	}
}

func TestStdlibExtraRuntimeWeb(t *testing.T) {
	runHeadlessFeature(t, stdExtraTestSource, featureScriptHead+`    await wait(300);
    button('Run').click();
    for (var wi = 0; wi < 160 && document.body.innerText.indexOf('world') < 0; wi++) await wait(50);
    var text = document.body.innerText;
    var expected = [
      '3|Ada|QWRh|258|report.csv|text/csv|#336699',
      '25.4|297|1|1|1',
      'world|2|name,formula',
      "Ada,'=1+1",
      'web|10|1|1|1|ok'
    ];
    for (var i = 0; i < expected.length; i++) {
      if (text.indexOf(expected[i]) < 0) throw new Error('missing ' + expected[i] + ' in ' + text);
    }
    if (text.indexOf('[object Object]') >= 0) throw new Error('unexpected object coercion: ' + text);
`+featureScriptTail)
}

const stdExtraSecuritySource = `import { files, mime, color, units, history, zip, archive, env, capabilities, log } from "std"

state result = ""

async fn run() -> null {
	badName, e1 = files.validate("../evil.txt", {})
	okName, e2 = files.validate("report.csv", {"allowedExt": "csv", "size": 10, "maxSize": 100})
	allowed, e3 = mime.isAllowed("image/png", "image/*")
	disallowed, e4 = mime.isAllowed("application/pdf", "image/*")
	envSecret, e5 = env.get("PATH")
	redacted, e6 = log.redact("token=abc123 Bearer xyz")
	black, e7 = color.parse("#000000")
	white, e8 = color.parse("#ffffff")
	ratio, e9 = color.contrast(black, white)
	h1, e10 = history.push("undo", "a")
	h2, e11 = history.push("undo", "b")
	prev, e12 = history.undo("undo")
	formatted, e13 = units.format(100.0, "px", 2)
	arc, e14 = zip.create()
	badZip, e15 = zip.addText(arc, "../evil.txt", "x")
	packed, e16 = zip.finish(arc)
	info, e17 = archive.info(packed)
	cap, e18 = capabilities.check("fs")

	result = str(badName == false) + "|" + str(okName) + "|" + str(allowed) + "|" + str(disallowed == false)
	result = result + "|" + str(len(envSecret) == 0) + "|" + str(len(e5) > 0)
	result = result + "|" + redacted + "|" + str(ratio > 20.0) + "|" + str(prev) + "|" + formatted
	result = str(result) + "|" + str(badZip == 0) + "|" + str(len(e15) > 0) + "|" + str(len(info) == 0)
	result = result + "|" + str(cap == false) + "|" + str(len(e18) > 0)
}

view main {
	p "{result}"
	button "Run" onClick: run
}`

func TestStdlibExtraSecurityWeb(t *testing.T) {
	runHeadlessFeature(t, stdExtraSecuritySource, featureScriptHead+`    await wait(300);
    button('Run').click();
    for (var wi = 0; wi < 160 && document.body.innerText.indexOf('|false') < 0; wi++) await wait(50);
    var text = document.body.innerText;
    var expected = [
      'true|true|true|true|true|true',
      'token=[REDACTED] Bearer [REDACTED]',
      '|a|100px|true|true|true|true|true'
    ];
    for (var i = 0; i < expected.length; i++) {
      if (text.indexOf(expected[i]) < 0) throw new Error('missing ' + expected[i] + ' in ' + text);
    }
`+featureScriptTail)
}
