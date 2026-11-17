package main

import (
	"strings"
	"testing"
)

func TestXmlModuleWeb(t *testing.T) {
	src := `import { json, xml } from "std"

state out = ""
async fn run() -> null {
	doc, err = xml.parse("<people xmlns:x=\"urn:test\"><person id=\"1\"><name>Mario</name><desc>hello <b>world</b></desc><!--note--><![CDATA[raw <x>]]><?pi data?><x:tag>v</x:tag></person><person id=\"2\"><name>Ada</name></person></people>")
	root, rerr = xml.root(doc)
	people, perr = xml.children(root)
	let first = people[0]
	let second = people[1]
	nameEl, ne = xml.find(first, "name")
	nm, nme = xml.name(root)
	cnt, cnte = xml.childCount(root)
	txt, txte = xml.text(nameEl)
	descEl, de = xml.find(first, "desc")
	dtxt, dte = xml.text(descEl)
	dcnt, dcte = xml.content(descEl)
	attrs, ae = xml.attrs(first)
	idv, ide = xml.attr(first, "id")
	hasId, hide = xml.hasAttr(first, "id")
	pth, pthe = xml.path(nameEl)
	parentEl, pare = xml.parent(first)
	nextEl, nexe = xml.next(first)
	prevEl, pre = xml.previous(second)
	xp, xe = xml.xpath(doc, "//person[@id='2']/name", {})
	xtxt, xtext = xml.text(xp[0])
	xcount, xce = xml.xpath(doc, "count(//person)", {})
	nsx, nsxe = xml.findNs(root, "tag", "urn:test")
	nsv, nsve = xml.namespace(nsx[0])
	foundAttr, fae = xml.findAttr(root, "id", "2")
	foundText, fte = xml.findText(root, "Ada", false)
	descendants, dse = xml.descendants(root)
	cd, cde = xml.toCsv(doc, "person", ["@id", "name"])
	jd, jde = xml.fromJson({"@id": "9", "name": "X"}, "person")
	jds, jdse = xml.stringify(jd, {"declaration": false})
	jobj, joe = xml.toJson(jd)
	xbytes, xbe = xml.toBytes(doc, "utf-16le", {"indent": true})
	roundDoc, roundErr = xml.parseBytes(xbytes, "utf-16le")
	roundRoot, roundRootErr = xml.root(roundDoc)
	roundName, roundNameErr = xml.name(roundRoot)
	loss, le = xml.toJson(doc, {"lossless": true})
	lossText, lse = json.stringify(loss)
	fragDoc, fde = xml.createDocument("wrapper")
	fragRoot, frre = xml.root(fragDoc)
	inner, ine = xml.setInnerXml(fragRoot, "<a>1</a><b>2</b>")
	innerKids, ike = xml.children(fragRoot)
	fragment, fge = xml.parseFragment(doc, "<x/><y/>")
	fragNodes, fgne = xml.childNodes(fragment)
	replaced, rpe = xml.replaceChildren(fragRoot, fragNodes)
	replacedCount, rpce = xml.childCount(fragRoot)

	let validXsd = "<xs:schema xmlns:xs=\"http://www.w3.org/2001/XMLSchema\"><xs:element name=\"person\"><xs:complexType><xs:sequence><xs:element name=\"name\" type=\"xs:string\"/></xs:sequence><xs:attribute name=\"id\" type=\"xs:int\" use=\"required\"/></xs:complexType></xs:element></xs:schema>"
	vdoc, vde = xml.parse("<person id=\"1\"><name>Ada</name></person>")
	vok, verr = xml.validateXsd(vdoc, validXsd)
	let badXsd = "<xs:schema xmlns:xs=\"http://www.w3.org/2001/XMLSchema\"><xs:element name=\"person\"><xs:complexType><xs:sequence><xs:element name=\"name\" type=\"xs:string\"/></xs:sequence><xs:attribute name=\"id\" type=\"xs:int\" use=\"required\"/></xs:complexType></xs:element></xs:schema>"
	bdoc, bde = xml.parse("<person><name>Ada</name></person>")
	bok, berr = xml.validateXsd(bdoc, badXsd)
	let rules = {"person": {"required": ["name"], "attributes": ["id"], "types": {"id": "int"}}}
	rok, rerr2 = xml.validateRules(doc, rules)
	doctypeDoc, doctypeErr = xml.parse("<!DOCTYPE root><root/>")
	syntaxOk, syntaxErr = xml.validate("<root>")

	d1, d1e = xml.parse("<r><a id=\"1\">x</a></r>")
	d2, d2e = xml.parse("<r><a id=\"2\">y</a><b/></r>")
	changes, che = xml.diff(d1, d2)

	sa, sae = xml.setAttr(first, "status", "active")
	textNode, tne = xml.createText(doc, "!")
	ap, ape = xml.append(first, textNode)
	ren, rene = xml.rename(first, "human")
	cl, cle = xml.clone(ren, true)
	clr, clre = xml.remove(cl)
	ser, sere = xml.stringify(doc, {"indent": true})

	out = nm + "|" + str(cnt) + "|" + txt + "|" + dtxt + "|" + dcnt + "|" + str(idv) + "|" + str(hasId) + "|" + pth + "|" + nsv + "|" + xtxt + "|" + str(xcount) + "|" + str(len(xp)) + "|" + str(len(descendants)) + "|" + str(len(foundAttr)) + "|" + str(len(foundText)) + "|" + str(parentEl != null) + "|" + str(nextEl != null) + "|" + str(prevEl != null)
	out = out + "\nCSV=" + cd
	out = out + "\nJSON=" + str(joe) + " root=" + str(jobj)
	out = out + "\nFJ=" + jds
	out = out + "\nBYTES=" + str(len(xbytes)) + " BNAME=" + roundName
	out = out + "\nFRAG=" + str(len(innerKids)) + "/" + str(replacedCount) + "/" + str(replaced)
	out = out + "\nLOSSLESS=" + lossText
	out = out + "\nXSD=" + str(vok) + "/" + verr + " BAD=" + str(bok) + "/" + berr
	out = out + "\nRULES=" + str(rok) + "/" + rerr2 + " DIFF=" + str(len(changes))
	out = out + "\nDOCTYPE=" + str(doctypeErr != null) + " SYNTAX=" + str(syntaxOk) + "/" + str(syntaxErr != null)
	out = out + "\nPRETTY=" + ser
}
view main {
	pre "{out}"
	button "Run" onClick: run
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Run').click();
    for (var wi = 0; wi < 160 && document.body.innerText.indexOf('CSV=') < 0; wi++) await wait(50);
    var text = document.body.innerText;
    if (text.indexOf('people|2|Mario|hello |hello world|1|true|/people/person[1]/name|urn:test|Ada|2|1|7|1|1|true|true|true') < 0) throw new Error('navigation: ' + text);
    if (text.indexOf('"@id","name"') < 0 || text.indexOf('"1","Mario"') < 0) throw new Error('csv: ' + text);
    if (text.indexOf('FJ=<person id="9"><name>X</name></person>') < 0) throw new Error('fromJson: ' + text);
    if (text.indexOf('BYTES=0') >= 0 || text.indexOf('BNAME=people') < 0) throw new Error('bytes: ' + text);
    if (text.indexOf('FRAG=2/2/true') < 0) throw new Error('fragment: ' + text);
    var ascii = '<?xml version="1.0" encoding="ISO-8859-1"?><r>caf';
    var iso = [];
    for (var bi = 0; bi < ascii.length; bi++) iso.push(ascii.charCodeAt(bi));
    iso.push(233, 60, 47, 114, 62);
    var isoParsed = __SUI_XML.parseBytes(iso);
    if (isoParsed[1] !== null || !isoParsed[0] || isoParsed[0].documentElement.textContent.indexOf('café') < 0) throw new Error('encoding: ' + text);
    if (text.indexOf('"type":"document"') < 0 || text.indexOf('"type":"comment"') < 0 || text.indexOf('"type":"cdata"') < 0) throw new Error('lossless: ' + text);
    if (text.indexOf('XSD=true/null') < 0 || text.indexOf('BAD=false/') < 0) throw new Error('xsd: ' + text);
    if (text.indexOf('RULES=true/null') < 0 || text.indexOf('DOCTYPE=true') < 0) throw new Error('rules/doctype: ' + text);
    if (text.indexOf('SYNTAX=false/true') < 0) throw new Error('syntax: ' + text);
    if (text.indexOf('status="active"') < 0 || text.indexOf('human') < 0) throw new Error('mutation: ' + text);
`+featureScriptTail)
}

func TestXmlModuleErrors(t *testing.T) {
	bad := []struct{ name, src, want string }{
		{"parse type", `import { xml } from "std"
fn run() -> null { doc, err = xml.parse(1) }
view main { button "x" onClick: run }`, "must be string"},
		{"url allow-list", `import { xml } from "std"
fn run() -> null { doc, err = xml.loadUrlSafe("https://example.com/a.xml", allow) }
view main { button "x" onClick: run }`, "allow-list must be a literal string"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			p, d := ParseSource(tc.src, "xml.sui")
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

func TestXmlModuleXsltAndLoadFile(t *testing.T) {
	src := `import { xml } from "std"
state out = ""
fn transform() -> null {
	doc, derr = xml.parse("<r><a>1</a></r>")
	let xsl = "<xsl:stylesheet xmlns:xsl=\"http://www.w3.org/1999/XSL/Transform\" version=\"1.0\"><xsl:template match=\"/\"><out><xsl:value-of select=\"//a\"/></out></xsl:template></xsl:stylesheet>"
	result, xerr = xml.xslt(doc, xsl)
	out = result + "|" + str(xerr)
}
async fn load() -> null {
	doc, err = xml.loadFile("nope.xml")
	out = str(doc) + "|" + err
}
view main {
	pre "{out}"
	button "Transform" onClick: transform
	button "Load" onClick: load
}`
	runHeadlessFeature(t, src, featureScriptHead+`    await wait(300);
    button('Transform').click(); await wait(300);
    var text = document.body.innerText;
    if (text.indexOf('<out>1</out>|null') < 0 && text.indexOf('XSLT is not available in this environment') < 0) throw new Error('xslt: ' + text);
    button('Load').click();
    for (var wi = 0; wi < 80 && document.body.innerText.indexOf('not available on the web target') < 0; wi++) await wait(50);
    text = document.body.innerText;
    if (text.indexOf('not available on the web target') < 0) throw new Error('loadFile: ' + text);
`+featureScriptTail)
}
