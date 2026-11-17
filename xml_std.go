package main

import "strings"

// xmlStdMethods maps generated __SUI_STD keys to methods on __SUI_XML.
// Keeping this table in one place avoids a hand-written case for every
// navigation/mutation helper.
var xmlStdMethods = map[string]string{
	"xml_parse":             "parse",
	"xml_parse_bytes":       "parseBytes",
	"xml_validate":          "validate",
	"xml_root":              "root",
	"xml_parent":            "parent",
	"xml_children":          "children",
	"xml_child_nodes":       "childNodes",
	"xml_child_count":       "childCount",
	"xml_next":              "next",
	"xml_previous":          "previous",
	"xml_find":              "find",
	"xml_find_ns":           "findNs",
	"xml_find_all":          "findAll",
	"xml_descendants":       "descendants",
	"xml_has":               "has",
	"xml_type":              "type",
	"xml_name":              "name",
	"xml_local_name":        "localName",
	"xml_namespace":         "namespace",
	"xml_prefix":            "prefix",
	"xml_text":              "text",
	"xml_content":           "content",
	"xml_attr":              "attr",
	"xml_attr_ns":           "attrNs",
	"xml_attrs":             "attrs",
	"xml_path":              "path",
	"xml_create_document":   "createDocument",
	"xml_create_element":    "createElement",
	"xml_create_element_ns": "createElementNs",
	"xml_create_text":       "createText",
	"xml_create_comment":    "createComment",
	"xml_create_cdata":      "createCdata",
	"xml_create_pi":         "createPi",
	"xml_parse_fragment":    "parseFragment",
	"xml_set_inner_xml":     "setInnerXml",
	"xml_replace_children":  "replaceChildren",
	"xml_append":            "append",
	"xml_prepend":           "prepend",
	"xml_insert_before":     "insertBefore",
	"xml_insert_after":      "insertAfter",
	"xml_remove":            "remove",
	"xml_replace":           "replace",
	"xml_clear":             "clear",
	"xml_clone":             "clone",
	"xml_rename":            "rename",
	"xml_set_text":          "setText",
	"xml_move":              "move",
	"xml_reorder":           "reorder",
	"xml_set_attr":          "setAttr",
	"xml_set_attr_ns":       "setAttrNs",
	"xml_remove_attr":       "removeAttr",
	"xml_rename_attr":       "renameAttr",
	"xml_has_attr":          "hasAttr",
	"xml_xpath":             "xpath",
	"xml_xpath_one":         "xpathOne",
	"xml_find_attr":         "findAttr",
	"xml_find_text":         "findText",
	"xml_stringify":         "stringify",
	"xml_to_bytes":          "toBytes",
	"xml_to_json":           "toJson",
	"xml_from_json":         "fromJson",
	"xml_to_csv":            "toCsv",
	"xml_xslt":              "xslt",
	"xml_merge":             "merge",
	"xml_merge3":            "merge3",
	"xml_diff":              "diff",
	"xml_patch":             "patch",
	"xml_validate_xsd":      "validateXsd",
	"xml_validate_xsd_all":  "validateXsdAll",
	"xml_validate_rules":    "validateRules",
}

// xmlStdBodyJs returns the JS function for one xml std intrinsic. File
// loading/saving need target-aware bridge calls; the rest delegate to the
// __SUI_XML runtime emitted only when the module is imported.
func xmlStdBodyJs(key, target string) string {
	if key == "xml_load_file" {
		if target == "desktop" {
			return `async function (path) {
  var fd = new URLSearchParams();
  fd.set("token", __SUI_TOKEN);
  fd.set("path", path);
  try {
    var r = await fetch("/__bridge__/file/read", { method: "POST", body: fd });
    var j = await r.json();
    if (!j.ok) return [null, j.message];
    return __SUI_XML.parse(j.data);
  } catch (e) { return [null, String(e)]; }
}`
		}
		return `async function (path) { return [null, "xml.loadFile is not available on the web target; read the file with fs.read or fetch it, then call xml.parse"]; }`
	}
	if key == "xml_load_url_safe" {
		return `async function (url, allow) {
  var fetched = await __sui_fetch_guarded(url, allow, { method: "GET" });
  if (fetched[1] !== null) return [null, fetched[1]];
  return __SUI_XML.parse(fetched[0]);
}`
	}
	if key == "xml_load_bytes" {
		if target == "desktop" {
			return `async function (path, encoding) {
  var fd = new URLSearchParams();
  fd.set("token", __SUI_TOKEN);
  fd.set("path", path);
  try {
    var r = await fetch("/__bridge__/file/readbytes", { method: "POST", body: fd });
    var j = await r.json();
    if (!j.ok) return [null, j.message];
    return __SUI_XML.parseBytes(__sui_b64_to_bytes(j.data), encoding);
  } catch (e) { return [null, String(e)]; }
}`
		}
		return `async function (path, encoding) { return [null, "xml.loadBytes is not available on the web target"]; }`
	}
	if key == "xml_save_bytes" {
		if target == "desktop" {
			return `async function (path, node, encoding, options) {
  var res = __SUI_XML.toBytes(node, encoding, options || {});
  if (res[1] !== null) return res;
  var fd = new URLSearchParams();
  fd.set("token", __SUI_TOKEN);
  fd.set("path", path);
  fd.set("data", __sui_bytes_to_b64(res[0]));
  try {
    var r = await fetch("/__bridge__/file/writebytes", { method: "POST", body: fd });
    var j = await r.json();
    return j.ok ? [null, null] : [null, j.message];
  } catch (e) { return [null, String(e)]; }
}`
		}
		return `async function (path, node, encoding, options) { return [null, "xml.saveBytes is not available on the web target"]; }`
	}
	if key == "xml_save_file" {
		if target == "desktop" {
			return `async function (path, node, options) {
  var res = __SUI_XML.stringify(node, options || {});
  if (res[1] !== null) return res;
  var enc = String((options && options.encoding) || "UTF-8").toLowerCase();
  if (enc !== "utf-8" && enc !== "utf8") return [null, "xml.saveFile supports UTF-8 only; use xml.toBytes for other encodings"];
  var fd = new URLSearchParams();
  fd.set("token", __SUI_TOKEN);
  fd.set("path", path);
  fd.set("data", res[0]);
  try {
    var r = await fetch("/__bridge__/file/write", { method: "POST", body: fd });
    var j = await r.json();
    return j.ok ? [null, null] : [null, j.message];
  } catch (e) { return [null, String(e)]; }
}`
		}
		return `async function (path, node, options) { return [null, "xml.saveFile is not available on the web target"]; }`
	}
	method := xmlStdMethods[key]
	if method == "" {
		return ""
	}
	return "function () { return __SUI_XML." + method + ".apply(null, arguments); }"
}

// xmlModuleImported reports whether the program imports the xml std module.
func xmlModuleImported(p *Prog) bool {
	if p == nil {
		return false
	}
	for i := 0; i < len(p.imports); i++ {
		im := p.imports[i]
		if im.kind != "std" {
			continue
		}
		for j := 0; j < len(im.names); j++ {
			if strings.TrimSpace(im.names[j]) == "xml" {
				return true
			}
		}
	}
	return false
}
