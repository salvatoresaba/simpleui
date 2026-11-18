package main

import (
	"strings"
	"testing"
)

func TestJsonModuleWeb(t *testing.T) {
	src := `import { json } from "std"
state out = ""
async fn run() -> null {
	let doc = {"store": {"books": [{"id": 1, "title": "A"}, {"id": 2, "title": "B"}], "open": true}, "a/b": 7, "~": 9}
	bookTitle, te = json.pointer(doc, "/store/books/1/title")
	escaped, ee = json.pointer(doc, "/a~1b")
	tilde, tie = json.pointer(doc, "/~0")
	rootDoc, rte0 = json.pointer(doc, "")
	missing, me = json.pointer(doc, "/missing")
	setDoc, se = json.setPointer(doc, "/store/open", false)
	setText, ste = json.stringify(setDoc, {"sortKeys": true})
	removedDoc, re = json.removePointer(doc, "/store/open")
	removedText, rte = json.stringify(removedDoc, {"sortKeys": true})
	patched, pe = json.patch(doc, [{"op": "replace", "path": "/store/open", "value": false}, {"op": "add", "path": "/store/books/-", "value": {"id": 3, "title": "C"}}])
	patchedText, pte = json.stringify(patched, {"sortKeys": true})
	bad, be = json.patch(doc, [{"op": "remove", "path": "/missing"}])
	moved, mpe = json.patch(doc, [{"op": "copy", "from": "/store/books/0", "path": "/store/books/-"}, {"op": "move", "from": "/store/books/0", "path": "/store/books/1"}, {"op": "remove", "path": "/store/books/0"}])
	movedText, mve = json.stringify(moved, {"sortKeys": true})
	testOk, toe = json.patch(doc, [{"op": "test", "path": "/store/open", "value": true}])
	testBad, tbe = json.patch(doc, [{"op": "test", "path": "/store/open", "value": false}])
	cloned, ce = json.clone(doc)
	equalDoc, eqe = json.equal(doc, cloned)

	let before = {"root": {"items": [{"id": 1, "v": "A"}, {"id": 2, "v": "B"}, {"id": 3, "v": "C"}]}}
	let after = {"root": {"items": [{"id": 3, "v": "C2"}, {"id": 1, "v": "A"}, {"id": 4, "v": "D"}]}}
	changes, che = json.diff(before, after, {"key": "id"})
	changesText, cte = json.stringify(changes)
	diffPatched, dpe = json.patch(before, changes)
	diffEqual, deq = json.equal(diffPatched, after)

	merged, mge = json.merge({"a": 1, "b": {"x": 1, "y": 2}}, {"b": {"y": null, "z": 3}, "c": [1, 2]})
	mergedText, mte = json.stringify(merged, {"sortKeys": true})

	let base = {"items": [{"id": 1, "v": "A"}, {"id": 2, "v": "B"}]}
	let ours = {"items": [{"id": 1, "v": "ours"}, {"id": 2, "v": "B"}]}
	let theirs = {"items": [{"id": 1, "v": "A"}, {"id": 2, "v": "theirs"}]}
	conflicts, cfe = json.merge3(base, ours, theirs, {"key": "id"})
	oursText, ote = json.stringify(ours)

	let conflictOurs = {"items": [{"id": 1, "v": "left"}]}
	let conflictBase = {"items": [{"id": 1, "v": "A"}]}
	let conflictTheirs = {"items": [{"id": 1, "v": "right"}]}
	conflictList, cle = json.merge3(conflictBase, conflictOurs, conflictTheirs, {"key": "id"})

	let schema = {"type": "object", "required": ["name", "age"], "properties": {"name": {"type": "string", "minLength": 3}, "age": {"type": "integer", "minimum": 18}, "tags": {"type": "array", "items": {"type": "string"}}}, "additionalProperties": false}
	errors, vae = json.validateAll({"name": "A", "age": 12, "tags": ["x", 2], "extra": true}, schema)

	utf16, u16e = json.toBytes(doc, "utf-16le")
	utf8, u8e = json.toBytes(doc, "utf-8", {"indent": 2})
	from16, f16e = json.parseBytes(utf16)
	from8, f8e = json.parseBytes(utf8)
	book16, t16e = json.pointer(from16, "/store/books/0/title")
	book8, t8e = json.pointer(from8, "/store/books/0/title")
	pretty, pre = json.stringify(doc, {"indent": true, "sortKeys": true, "newline": true})

	out = "PTR=" + str(bookTitle) + "|" + str(escaped) + "|" + str(tilde) + "|" + str(te == null) + "|" + str(me != null) + "|" + str(rootDoc != null)
	out = out + "\nSET=" + setText
	out = out + "\nREMOVED=" + removedText
	out = out + "\nPATCH=" + patchedText
	out = out + "\nMOVED=" + movedText
	out = out + "\nBAD=" + str(be != null) + "|" + str(toe == null) + "|" + str(tbe != null)
	out = out + "\nCLONE=" + str(eqe == null) + "|" + str(equalDoc)
	out = out + "\nDIFF=" + changesText + "|" + str(deq == null) + "|" + str(diffEqual)
	out = out + "\nMERGE=" + mergedText
	out = out + "\nTHREE=" + str(len(conflicts)) + "|" + oursText
	out = out + "\nCONFLICT=" + str(len(conflictList)) + "|" + conflictList[0]
	out = out + "\nERRORS=" + str(len(errors)) + "|" + errors[0] + "|" + errors[1]
	out = out + "\nBYTES=" + str(book16) + "|" + str(book8) + "|" + str(u16e == null) + "|" + str(u8e == null) + "|" + str(f16e == null) + "|" + str(f8e == null)
	out = out + "\nPRETTY=" + pretty
}
view main {
	pre "{out}"
	button "Run" onClick: run
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Run').click();
    for (var wi = 0; wi < 200 && document.body.innerText.indexOf('PRETTY=') < 0; wi++) await wait(50);
    var text = document.body.innerText;
    if (text.indexOf('PTR=B|7|9|true|true|true') < 0) throw new Error('pointer: ' + text);
    if (text.indexOf('SET={"a/b":7,"store":{"books":[{"id":1,"title":"A"},{"id":2,"title":"B"}],"open":false},"~":9}') < 0) throw new Error('setPointer: ' + text);
    if (text.indexOf('REMOVED={"a/b":7,"store":{"books":[{"id":1,"title":"A"},{"id":2,"title":"B"}]},"~":9}') < 0) throw new Error('removePointer: ' + text);
    if (text.indexOf('"id":3') < 0 || text.indexOf('"open":false') < 0) throw new Error('patch: ' + text);
    if (text.indexOf('BAD=true|true|true') < 0) throw new Error('patch errors: ' + text);
    if (text.indexOf('MOVED=') < 0 || text.indexOf('"title":"A"') < 0) throw new Error('patch move/copy: ' + text);
    if (text.indexOf('CLONE=true|true') < 0) throw new Error('clone: ' + text);
    if (text.indexOf('"op":"move"') < 0 || text.indexOf('"op":"add"') < 0 || text.indexOf('"op":"remove"') < 0 || text.indexOf('"op":"replace"') < 0) throw new Error('diff ops: ' + text);
    if (text.indexOf('|true|true') < 0) throw new Error('diff patch: ' + text);
    if (text.indexOf('MERGE={"a":1,"b":{"x":1,"z":3},"c":[1,2]}') < 0) throw new Error('merge: ' + text);
    if (text.indexOf('THREE=0|') < 0 || text.indexOf('"v":"ours"') < 0 || text.indexOf('"v":"theirs"') < 0) throw new Error('merge3: ' + text);
    if (text.indexOf('CONFLICT=1|') < 0) throw new Error('merge3 conflict: ' + text);
    if (text.indexOf('ERRORS=') < 0) throw new Error('validateAll missing: ' + text);
    var m = /ERRORS=(\d+)/.exec(text);
    if (!m || Number(m[1]) < 3) throw new Error('validateAll count: ' + text);
    if (text.indexOf('BYTES=A|A|true|true|true|true') < 0) throw new Error('bytes: ' + text);
    if (text.indexOf('PRETTY=') < 0 || text.indexOf('"a/b": 7') < 0) throw new Error('pretty: ' + text);
`+featureScriptTail)
}

func TestJsonBytesWebUnavailable(t *testing.T) {
	src := `import { json } from "std"
state out = ""
async fn run() -> null {
	let value = {"x": 1}
	loaded, loadErr = json.loadFile("a.json")
	loadedBytes, loadBytesErr = json.loadBytes("a.json", "utf-16le")
	saved, saveErr = json.saveFile("a.json", value)
	savedBytes, saveBytesErr = json.saveBytes("a.json", value, "utf-16le")
	out = loadErr + "|" + loadBytesErr + "|" + saveErr + "|" + saveBytesErr
}
view main {
	p "{out}"
	button "Run" onClick: run
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Run').click(); await wait(300);
    var text = document.body.innerText;
    if (text.indexOf('json.loadFile is not available on the web target') < 0) throw new Error('json loadFile: ' + text);
    if (text.indexOf('json.loadBytes is not available on the web target') < 0) throw new Error('json loadBytes: ' + text);
    if (text.indexOf('json.saveFile is not available on the web target') < 0) throw new Error('json saveFile: ' + text);
    if (text.indexOf('json.saveBytes is not available on the web target') < 0) throw new Error('json saveBytes: ' + text);
`+featureScriptTail)
}

func TestJsonArgumentErrors(t *testing.T) {
	bad := []struct{ name, src, want string }{
		{"pointer args", `import { json } from "std"
fn run() -> null { v, e = json.pointer({"a": 1}, 2) }
view main { button "x" onClick: run }`, "argument 2 of json.pointer must be string, got int"},
		{"patch too many args", `import { json } from "std"
fn run() -> null { v, e = json.patch({}, [], []) }
view main { button "x" onClick: run }`, "std function \"json.patch\" expects 2 arguments, got 3"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "json.sui")
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
