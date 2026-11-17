package main

import (
	"strings"
	"testing"
)

func TestAdvancedStdlibWeb(t *testing.T) {
	src := `import { json } from "std"

state out = ""
async fn run() -> null {
	out = crypto.status() + "|" + str(crypto.available())
	storage.set("sui.demo", {"n": 1})
	storage.set("sui.text", "hello")
	let okSync = storage.sync()
	let q = storage.quota()
	let bytes = crypto.randomBytes(4)
	let hash = crypto.hash("sha256", "hello")
	let id = crypto.uuid()
	let joined = path.join("a", "b", "..", "c")
	let dir = path.dirname("C:/x/y.txt")
	let base = path.basename("C:/x/y.txt")
	let ext = path.extname("C:/x/y.txt")

	let schema = {"type": "object", "required": ["name", "age"], "additionalProperties": false, "properties": {"name": {"type": "string", "minLength": 2}, "age": {"type": "integer", "minimum": 18}}}
	valid, validErr = json.validate({"name": "Ada", "age": 36}, schema)
	invalid, invalidErr = json.validate({"name": "A", "age": 12, "extra": true}, schema)

	out = joined + "|" + dir + "|" + base + "|" + ext
	out = out + "|" + hash + "|" + str(len(id)) + "|" + str(len(bytes))
	out = out + "|" + str(storage.get("sui.text")) + "|" + str(len(storage.keys()))
	out = out + "|" + str(q["used"] > 0) + "|" + str(q["quota"] >= 0) + "|" + str(okSync)
	out = out + "|" + str(valid) + "|" + validErr + "|" + str(invalid) + "|" + invalidErr
	out = out + "|" + crypto.status() + "|" + str(crypto.available())
	storage.remove("sui.text")
	out = out + "|" + str(storage.get("sui.text", "missing"))
}
view main {
	p "{out}"
	button "Run" onClick: run
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Run').click();
    for (var wi = 0; wi < 160 && document.body.innerText.indexOf('a/c|C:/x|y.txt|.txt') < 0; wi++) await wait(50);
    var text = document.body.innerText;
    if (text.indexOf('a/c|C:/x|y.txt|.txt') < 0) throw new Error('path: ' + text);
    if (text.indexOf('2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824') < 0) throw new Error('hash: ' + text);
    if (text.indexOf('|36|4|hello|') < 0) throw new Error('uuid/bytes/storage: ' + text);
    if (text.indexOf('|true|true|true|true|') < 0) throw new Error('quota/sync/valid: ' + text);
    if (text.indexOf('$.name: must have at least 2 characters') < 0) throw new Error('schema detail: ' + text);
    if (text.indexOf('|ready|true|missing') < 0) throw new Error('crypto/status/storage: ' + text);
`+featureScriptTail)

	bad := []struct{ name, src, want string }{
		{"hash async", `fn run() -> null { let h = crypto.hash("sha256", "x") }
view main { button "x" onClick: run }`, "crypto.hash is async"},
		{"quota async", `fn run() -> null { let q = storage.quota() }
view main { button "x" onClick: run }`, "storage.quota is async"},
		{"path arg", `fn run() -> null { let p = path.join("a", 1) }
view main { button "x" onClick: run }`, "path.join arguments must be strings"},
		{"random bytes arg", `fn run() -> null { let b = crypto.randomBytes("x") }
view main { button "x" onClick: run }`, "crypto.randomBytes requires an int"},
		{"available args", `fn run() -> null { let a = crypto.available(1) }
view main { button "x" onClick: run }`, "crypto.available takes exactly 0 arguments"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "std.sui")
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

func TestAdvancedStdlibWebFsUnavailable(t *testing.T) {
	src := `import { fs } from "std"
state out = ""
async fn run() -> null {
	copyOk, err = fs.copy("a", "b")
	bytes, byteErr = fs.readBytes("a")
	writeOk, writeErr = fs.writeBytes("a", [1, 2, 3])
	out = str(copyOk) + " " + err + "|" + str(bytes) + " " + byteErr + "|" + str(writeOk) + " " + writeErr
}
view main {
	p "{out}"
	button "Run" onClick: run
}`
	prog, dl := ParseSource(src, "fs.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Run').click(); await wait(200);
    var text = document.body.innerText;
    if (text.indexOf('fs.copy is not available on the web target') < 0) throw new Error('web fs: ' + text);
    if (text.indexOf('fs.readBytes is not available on the web target') < 0) throw new Error('web fs bytes read: ' + text);
    if (text.indexOf('fs.writeBytes is not available on the web target') < 0) throw new Error('web fs bytes write: ' + text);
`+featureScriptTail)
}
