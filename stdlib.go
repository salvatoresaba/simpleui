package main

// stdlib.go is the contract for the SimpleUI standard library.
//
// The "batteries included" functions are recognized syntactically: a call to
// `http.get(...)` is only valid if `http` was imported from "std". The
// typechecker uses the parameter/return types below; the JS generator uses
// `jsKey` to emit `await __SUI_STD.<jsKey>(...)`, whose implementations live in
// the generated runtime prelude (see jsgen.go).
//
// Return convention (v0.1): every std function returns exactly TWO values:
//   ret0  the primary value (may be "null", e.g. fs.write)
//   ret1  an error string, or "null" when the call succeeded
// In SUI source: `value, err = http.get(url)`.

// StdFunc describes one intrinsic std function.
type StdFunc struct {
	mod    string   // module: "http" | "ws" | "fs" | "exec" | "json"
	name   string   // function name within the module
	params []string // SUI type names: "int" "float" "string" "bool"
	// "array[string]" "any" "null"
	ret0     string // primary return type (may be "null")
	ret1     string // error return type (may be "null")
	jsKey    string // key into the generated __SUI_STD object
	async    bool
	optional int // number of trailing parameters that may be omitted
}

// StdModules is the set of modules that may be imported with
// `import { a, b } from "std"`.
var StdModules = []string{
	"http",
	"ws",
	"fs",
	"exec",
	"json",
	"table",
	"tree",
	"chart",
	"richtext",
	"dialog",
	"xml",
}

// CoreModules are intrinsic namespaces that are always available: they need no
// `import` because they act on the program itself (the rendered UI) rather than
// on an external service. `table.display`/`table.data` target the table with a
// matching `name:` attribute.
var CoreModules = []string{
	"table",
	"tree",
	"chart",
	"richtext",
}

// IsCoreModule reports whether name is an always-available intrinsic module.
func IsCoreModule(name string) bool {
	for _, m := range CoreModules {
		if m == name {
			return true
		}
	}
	return false
}

// StdFuncs is the complete v0.1 standard library table.
var StdFuncs = []StdFunc{
	// http
	StdFunc{mod: "http", name: "get", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "http_get", async: true},
	StdFunc{mod: "http", name: "getStatus", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "http_getstatus", async: true},
	StdFunc{mod: "http", name: "post", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "http_post", async: true},
	StdFunc{mod: "http", name: "postFiles", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "http_postfiles", async: true},
	StdFunc{mod: "http", name: "postFilesSafe", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "http_postfilessafe", async: true},
	StdFunc{mod: "http", name: "getSafe", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "http_getsafe", async: true},
	StdFunc{mod: "http", name: "getStatusSafe", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "http_getstatussafe", async: true},
	StdFunc{mod: "http", name: "postSafe", params: []string{"string", "string", "string"}, ret0: "string", ret1: "string", jsKey: "http_postsafe", async: true},
	StdFunc{mod: "http", name: "getWithHeaders", params: []string{"string", "any"}, ret0: "string", ret1: "string", jsKey: "http_getwithheaders", async: true},
	StdFunc{mod: "http", name: "getWithHeadersSafe", params: []string{"string", "any", "string"}, ret0: "string", ret1: "string", jsKey: "http_getwithheaderssafe", async: true},
	StdFunc{mod: "http", name: "postWithHeaders", params: []string{"string", "string", "any"}, ret0: "string", ret1: "string", jsKey: "http_postwithheaders", async: true},
	StdFunc{mod: "http", name: "postWithHeadersSafe", params: []string{"string", "string", "any", "string"}, ret0: "string", ret1: "string", jsKey: "http_postwithheaderssafe", async: true},
	StdFunc{mod: "http", name: "request", params: []string{"string", "string", "any", "string", "int"}, ret0: "string", ret1: "string", jsKey: "http_request", async: true},
	StdFunc{mod: "http", name: "requestSafe", params: []string{"string", "string", "any", "string", "int", "string"}, ret0: "string", ret1: "string", jsKey: "http_requestsafe", async: true},
	StdFunc{mod: "http", name: "upload", params: []string{"string", "string", "string", "string"}, ret0: "string", ret1: "string", jsKey: "http_upload", async: true},
	StdFunc{mod: "http", name: "uploadSafe", params: []string{"string", "string", "string", "string", "string"}, ret0: "string", ret1: "string", jsKey: "http_uploadsafe", async: true},
	StdFunc{mod: "http", name: "uploadWithHeaders", params: []string{"string", "string", "string", "string", "any"}, ret0: "string", ret1: "string", jsKey: "http_uploadwithheaders", async: true},
	StdFunc{mod: "http", name: "uploadWithHeadersSafe", params: []string{"string", "string", "string", "string", "any", "string"}, ret0: "string", ret1: "string", jsKey: "http_uploadwithheaderssafe", async: true},
	// ws
	StdFunc{mod: "ws", name: "connect", params: []string{"string"}, ret0: "any", ret1: "string", jsKey: "ws_connect", async: true},
	StdFunc{mod: "ws", name: "connectSafe", params: []string{"string", "string"}, ret0: "any", ret1: "string", jsKey: "ws_connectsafe", async: true},
	StdFunc{mod: "ws", name: "connectWithOptions", params: []string{"string", "any"}, ret0: "any", ret1: "string", jsKey: "ws_connectwithoptions", async: true},
	StdFunc{mod: "ws", name: "connectWithOptionsSafe", params: []string{"string", "any", "string"}, ret0: "any", ret1: "string", jsKey: "ws_connectwithoptionssafe", async: true},
	StdFunc{mod: "ws", name: "send", params: []string{"any", "string"}, ret0: "null", ret1: "string", jsKey: "ws_send", async: true},
	StdFunc{mod: "ws", name: "onMessage", params: []string{"any", "string"}, ret0: "null", ret1: "string", jsKey: "ws_onmessage", async: true},
	StdFunc{mod: "ws", name: "onOpen", params: []string{"any", "string"}, ret0: "null", ret1: "string", jsKey: "ws_onopen", async: true},
	StdFunc{mod: "ws", name: "onClose", params: []string{"any", "string"}, ret0: "null", ret1: "string", jsKey: "ws_onclose", async: true},
	StdFunc{mod: "ws", name: "onError", params: []string{"any", "string"}, ret0: "null", ret1: "string", jsKey: "ws_onerror", async: true},
	StdFunc{mod: "ws", name: "status", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "ws_status", async: false},
	StdFunc{mod: "ws", name: "reconnect", params: []string{"any"}, ret0: "null", ret1: "string", jsKey: "ws_reconnect", async: true},
	StdFunc{mod: "ws", name: "close", params: []string{"any"}, ret0: "null", ret1: "string", jsKey: "ws_close", async: true},
	// fs
	StdFunc{mod: "fs", name: "read", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "fs_read", async: true},
	StdFunc{mod: "fs", name: "write", params: []string{"string", "string"}, ret0: "null", ret1: "string", jsKey: "fs_write", async: true},
	StdFunc{mod: "fs", name: "exists", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "fs_exists", async: true},
	StdFunc{mod: "fs", name: "stat", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "fs_stat", async: true},
	StdFunc{mod: "fs", name: "mkdir", params: []string{"string"}, ret0: "null", ret1: "string", jsKey: "fs_mkdir", async: true},
	StdFunc{mod: "fs", name: "remove", params: []string{"string"}, ret0: "null", ret1: "string", jsKey: "fs_remove", async: true},
	StdFunc{mod: "fs", name: "list", params: []string{"string"}, ret0: "array[string]", ret1: "string", jsKey: "fs_list", async: true},
	StdFunc{mod: "fs", name: "copy", params: []string{"string", "string"}, ret0: "null", ret1: "string", jsKey: "fs_copy", async: true},
	StdFunc{mod: "fs", name: "move", params: []string{"string", "string"}, ret0: "null", ret1: "string", jsKey: "fs_move", async: true},
	StdFunc{mod: "fs", name: "rename", params: []string{"string", "string"}, ret0: "null", ret1: "string", jsKey: "fs_move", async: true},
	StdFunc{mod: "fs", name: "isDir", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "fs_isdir", async: true},
	StdFunc{mod: "fs", name: "size", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "fs_size", async: true},
	StdFunc{mod: "fs", name: "readBytes", params: []string{"string"}, ret0: "array[int]", ret1: "string", jsKey: "fs_readbytes", async: true},
	StdFunc{mod: "fs", name: "writeBytes", params: []string{"string", "any"}, ret0: "null", ret1: "string", jsKey: "fs_writebytes", async: true},
	// exec
	StdFunc{mod: "exec", name: "run", params: []string{"string", "array[string]"}, ret0: "string", ret1: "string", jsKey: "exec_run", async: true},
	// json
	StdFunc{mod: "json", name: "parse", params: []string{"string"}, ret0: "any", ret1: "string", jsKey: "json_parse", async: true},
	StdFunc{mod: "json", name: "parseBytes", params: []string{"any", "string"}, ret0: "any", ret1: "string", jsKey: "json_parse_bytes", async: false, optional: 1},
	StdFunc{mod: "json", name: "stringify", params: []string{"any", "any"}, ret0: "string", ret1: "string", jsKey: "json_stringify", async: true, optional: 1},
	StdFunc{mod: "json", name: "toBytes", params: []string{"any", "string", "any"}, ret0: "array[int]", ret1: "string", jsKey: "json_to_bytes", async: false, optional: 1},
	StdFunc{mod: "json", name: "validate", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "json_validate", async: false},
	StdFunc{mod: "json", name: "validateAll", params: []string{"any", "any"}, ret0: "array[string]", ret1: "string", jsKey: "json_validate_all", async: false},
	StdFunc{mod: "json", name: "pointer", params: []string{"any", "string"}, ret0: "any", ret1: "string", jsKey: "json_pointer", async: false},
	StdFunc{mod: "json", name: "setPointer", params: []string{"any", "string", "any"}, ret0: "any", ret1: "string", jsKey: "json_set_pointer", async: false},
	StdFunc{mod: "json", name: "removePointer", params: []string{"any", "string"}, ret0: "any", ret1: "string", jsKey: "json_remove_pointer", async: false},
	StdFunc{mod: "json", name: "clone", params: []string{"any"}, ret0: "any", ret1: "string", jsKey: "json_clone", async: false},
	StdFunc{mod: "json", name: "equal", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "json_equal", async: false},
	StdFunc{mod: "json", name: "diff", params: []string{"any", "any", "any"}, ret0: "array[any]", ret1: "string", jsKey: "json_diff", async: false, optional: 2},
	StdFunc{mod: "json", name: "patch", params: []string{"any", "any"}, ret0: "any", ret1: "string", jsKey: "json_patch", async: false},
	StdFunc{mod: "json", name: "merge", params: []string{"any", "any"}, ret0: "any", ret1: "string", jsKey: "json_merge", async: false},
	StdFunc{mod: "json", name: "merge3", params: []string{"any", "any", "any", "any"}, ret0: "array[string]", ret1: "string", jsKey: "json_merge3", async: false, optional: 3},
	StdFunc{mod: "json", name: "loadFile", params: []string{"string", "string"}, ret0: "any", ret1: "string", jsKey: "json_load_file", async: true, optional: 1},
	StdFunc{mod: "json", name: "loadBytes", params: []string{"string", "string"}, ret0: "any", ret1: "string", jsKey: "json_load_bytes", async: true, optional: 1},
	StdFunc{mod: "json", name: "saveFile", params: []string{"string", "any", "any"}, ret0: "null", ret1: "string", jsKey: "json_save_file", async: true, optional: 1},
	StdFunc{mod: "json", name: "saveBytes", params: []string{"string", "any", "string", "any"}, ret0: "null", ret1: "string", jsKey: "json_save_bytes", async: true, optional: 1},
	// xml (browser DOM, XPath, XSLT; see xml_runtime.go)
	StdFunc{mod: "xml", name: "parse", params: []string{"string"}, ret0: "xmlDocument", ret1: "string", jsKey: "xml_parse", async: false},
	StdFunc{mod: "xml", name: "parseBytes", params: []string{"any", "string"}, ret0: "xmlDocument", ret1: "string", jsKey: "xml_parse_bytes", async: false, optional: 1},
	StdFunc{mod: "xml", name: "validate", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "xml_validate", async: false},
	StdFunc{mod: "xml", name: "loadFile", params: []string{"string"}, ret0: "xmlDocument", ret1: "string", jsKey: "xml_load_file", async: true},
	StdFunc{mod: "xml", name: "loadUrlSafe", params: []string{"string", "string"}, ret0: "xmlDocument", ret1: "string", jsKey: "xml_load_url_safe", async: true},
	StdFunc{mod: "xml", name: "saveFile", params: []string{"string", "any", "any"}, ret0: "null", ret1: "string", jsKey: "xml_save_file", async: true, optional: 1},
	StdFunc{mod: "xml", name: "loadBytes", params: []string{"string", "string"}, ret0: "xmlDocument", ret1: "string", jsKey: "xml_load_bytes", async: true, optional: 1},
	StdFunc{mod: "xml", name: "saveBytes", params: []string{"string", "any", "string", "any"}, ret0: "null", ret1: "string", jsKey: "xml_save_bytes", async: true, optional: 1},
	StdFunc{mod: "xml", name: "root", params: []string{"any"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_root", async: false},
	StdFunc{mod: "xml", name: "parent", params: []string{"any"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_parent", async: false},
	StdFunc{mod: "xml", name: "children", params: []string{"any"}, ret0: "array[xmlNode]", ret1: "string", jsKey: "xml_children", async: false},
	StdFunc{mod: "xml", name: "childNodes", params: []string{"any"}, ret0: "array[xmlNode]", ret1: "string", jsKey: "xml_child_nodes", async: false},
	StdFunc{mod: "xml", name: "childCount", params: []string{"any"}, ret0: "int", ret1: "string", jsKey: "xml_child_count", async: false},
	StdFunc{mod: "xml", name: "next", params: []string{"any"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_next", async: false},
	StdFunc{mod: "xml", name: "previous", params: []string{"any"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_previous", async: false},
	StdFunc{mod: "xml", name: "find", params: []string{"any", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_find", async: false},
	StdFunc{mod: "xml", name: "findNs", params: []string{"any", "string", "string"}, ret0: "array[xmlNode]", ret1: "string", jsKey: "xml_find_ns", async: false},
	StdFunc{mod: "xml", name: "findAll", params: []string{"any", "string"}, ret0: "array[xmlNode]", ret1: "string", jsKey: "xml_find_all", async: false},
	StdFunc{mod: "xml", name: "descendants", params: []string{"any"}, ret0: "array[xmlNode]", ret1: "string", jsKey: "xml_descendants", async: false},
	StdFunc{mod: "xml", name: "has", params: []string{"any", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_has", async: false},
	StdFunc{mod: "xml", name: "type", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_type", async: false},
	StdFunc{mod: "xml", name: "name", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_name", async: false},
	StdFunc{mod: "xml", name: "localName", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_local_name", async: false},
	StdFunc{mod: "xml", name: "namespace", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_namespace", async: false},
	StdFunc{mod: "xml", name: "prefix", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_prefix", async: false},
	StdFunc{mod: "xml", name: "text", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_text", async: false},
	StdFunc{mod: "xml", name: "content", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_content", async: false},
	StdFunc{mod: "xml", name: "attr", params: []string{"any", "string"}, ret0: "any", ret1: "string", jsKey: "xml_attr", async: false},
	StdFunc{mod: "xml", name: "attrNs", params: []string{"any", "string", "string"}, ret0: "any", ret1: "string", jsKey: "xml_attr_ns", async: false},
	StdFunc{mod: "xml", name: "attrs", params: []string{"any"}, ret0: "map[string]string", ret1: "string", jsKey: "xml_attrs", async: false},
	StdFunc{mod: "xml", name: "path", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "xml_path", async: false},
	StdFunc{mod: "xml", name: "createDocument", params: []string{"string"}, ret0: "xmlDocument", ret1: "string", jsKey: "xml_create_document", async: false, optional: 1},
	StdFunc{mod: "xml", name: "createElement", params: []string{"any", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_create_element", async: false},
	StdFunc{mod: "xml", name: "createElementNs", params: []string{"any", "string", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_create_element_ns", async: false},
	StdFunc{mod: "xml", name: "createText", params: []string{"any", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_create_text", async: false},
	StdFunc{mod: "xml", name: "createComment", params: []string{"any", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_create_comment", async: false},
	StdFunc{mod: "xml", name: "createCdata", params: []string{"any", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_create_cdata", async: false},
	StdFunc{mod: "xml", name: "createPi", params: []string{"any", "string", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_create_pi", async: false},
	StdFunc{mod: "xml", name: "parseFragment", params: []string{"any", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_parse_fragment", async: false},
	StdFunc{mod: "xml", name: "setInnerXml", params: []string{"any", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_set_inner_xml", async: false},
	StdFunc{mod: "xml", name: "replaceChildren", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_replace_children", async: false},
	StdFunc{mod: "xml", name: "append", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_append", async: false},
	StdFunc{mod: "xml", name: "prepend", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_prepend", async: false},
	StdFunc{mod: "xml", name: "insertBefore", params: []string{"any", "any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_insert_before", async: false},
	StdFunc{mod: "xml", name: "insertAfter", params: []string{"any", "any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_insert_after", async: false},
	StdFunc{mod: "xml", name: "remove", params: []string{"any"}, ret0: "bool", ret1: "string", jsKey: "xml_remove", async: false},
	StdFunc{mod: "xml", name: "replace", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_replace", async: false},
	StdFunc{mod: "xml", name: "clear", params: []string{"any"}, ret0: "bool", ret1: "string", jsKey: "xml_clear", async: false},
	StdFunc{mod: "xml", name: "clone", params: []string{"any", "bool"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_clone", async: false, optional: 1},
	StdFunc{mod: "xml", name: "rename", params: []string{"any", "string"}, ret0: "xmlNode", ret1: "string", jsKey: "xml_rename", async: false},
	StdFunc{mod: "xml", name: "setText", params: []string{"any", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_set_text", async: false},
	StdFunc{mod: "xml", name: "move", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_move", async: false},
	StdFunc{mod: "xml", name: "reorder", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_reorder", async: false},
	StdFunc{mod: "xml", name: "setAttr", params: []string{"any", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_set_attr", async: false},
	StdFunc{mod: "xml", name: "setAttrNs", params: []string{"any", "string", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_set_attr_ns", async: false},
	StdFunc{mod: "xml", name: "removeAttr", params: []string{"any", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_remove_attr", async: false},
	StdFunc{mod: "xml", name: "renameAttr", params: []string{"any", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_rename_attr", async: false},
	StdFunc{mod: "xml", name: "hasAttr", params: []string{"any", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_has_attr", async: false},
	StdFunc{mod: "xml", name: "xpath", params: []string{"any", "string", "any"}, ret0: "any", ret1: "string", jsKey: "xml_xpath", async: false, optional: 1},
	StdFunc{mod: "xml", name: "xpathOne", params: []string{"any", "string", "any"}, ret0: "any", ret1: "string", jsKey: "xml_xpath_one", async: false, optional: 1},
	StdFunc{mod: "xml", name: "findAttr", params: []string{"any", "string", "any"}, ret0: "array[xmlNode]", ret1: "string", jsKey: "xml_find_attr", async: false, optional: 1},
	StdFunc{mod: "xml", name: "findText", params: []string{"any", "string", "bool"}, ret0: "array[xmlNode]", ret1: "string", jsKey: "xml_find_text", async: false, optional: 1},
	StdFunc{mod: "xml", name: "stringify", params: []string{"any", "any"}, ret0: "string", ret1: "string", jsKey: "xml_stringify", async: false, optional: 1},
	StdFunc{mod: "xml", name: "toBytes", params: []string{"any", "string", "any"}, ret0: "any", ret1: "string", jsKey: "xml_to_bytes", async: false, optional: 1},
	StdFunc{mod: "xml", name: "toJson", params: []string{"any", "any"}, ret0: "any", ret1: "string", jsKey: "xml_to_json", async: false, optional: 1},
	StdFunc{mod: "xml", name: "fromJson", params: []string{"any", "string"}, ret0: "xmlDocument", ret1: "string", jsKey: "xml_from_json", async: false, optional: 1},
	StdFunc{mod: "xml", name: "toCsv", params: []string{"any", "string", "any"}, ret0: "string", ret1: "string", jsKey: "xml_to_csv", async: false, optional: 1},
	StdFunc{mod: "xml", name: "xslt", params: []string{"any", "string"}, ret0: "string", ret1: "string", jsKey: "xml_xslt", async: false},
	StdFunc{mod: "xml", name: "merge", params: []string{"any", "any", "any"}, ret0: "int", ret1: "string", jsKey: "xml_merge", async: false, optional: 1},
	StdFunc{mod: "xml", name: "merge3", params: []string{"any", "any", "any", "any"}, ret0: "array[string]", ret1: "string", jsKey: "xml_merge3", async: false, optional: 3},
	StdFunc{mod: "xml", name: "diff", params: []string{"any", "any", "any"}, ret0: "array[any]", ret1: "string", jsKey: "xml_diff", async: false, optional: 2},
	StdFunc{mod: "xml", name: "patch", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_patch", async: false},
	StdFunc{mod: "xml", name: "validateXsd", params: []string{"any", "string"}, ret0: "bool", ret1: "string", jsKey: "xml_validate_xsd", async: false},
	StdFunc{mod: "xml", name: "validateXsdAll", params: []string{"any", "string"}, ret0: "array[string]", ret1: "string", jsKey: "xml_validate_xsd_all", async: false},
	StdFunc{mod: "xml", name: "validateRules", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "xml_validate_rules", async: false},
	// dialog (desktop target only)
	StdFunc{mod: "dialog", name: "pickFile", params: []string{}, ret0: "string", ret1: "string", jsKey: "dialog_pickfile", async: true},
	StdFunc{mod: "dialog", name: "pickFiles", params: []string{}, ret0: "array[string]", ret1: "string", jsKey: "dialog_pickfiles", async: true},
	StdFunc{mod: "dialog", name: "pickDirectory", params: []string{}, ret0: "string", ret1: "string", jsKey: "dialog_pickdirectory", async: true},
	// table (core: always available, acts on the rendered table named with `name:`)
	StdFunc{mod: "richtext", name: "get", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "richtext_get", async: false},
	StdFunc{mod: "richtext", name: "set", params: []string{"string", "any"}, ret0: "bool", ret1: "string", jsKey: "richtext_set", async: false},
	StdFunc{mod: "richtext", name: "loadContent", params: []string{"string", "any", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_loadcontent", async: true, optional: 2},
	StdFunc{mod: "richtext", name: "command", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "richtext_command", async: false, optional: 1},
	StdFunc{mod: "richtext", name: "format", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "richtext_format", async: false, optional: 1},
	StdFunc{mod: "richtext", name: "insertText", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_inserttext", async: false},
	StdFunc{mod: "richtext", name: "insertImage", params: []string{"string", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_insertimage", async: false, optional: 1},
	StdFunc{mod: "richtext", name: "insertLink", params: []string{"string", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_insertlink", async: false, optional: 1},
	StdFunc{mod: "richtext", name: "insertTable", params: []string{"string", "int", "int"}, ret0: "bool", ret1: "string", jsKey: "richtext_inserttable", async: false},
	StdFunc{mod: "richtext", name: "undo", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "richtext_undo", async: false},
	StdFunc{mod: "richtext", name: "redo", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "richtext_redo", async: false},
	StdFunc{mod: "richtext", name: "focus", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "richtext_focus", async: false},
	StdFunc{mod: "richtext", name: "exportFormat", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "richtext_export", async: false},
	StdFunc{mod: "richtext", name: "exportBytes", params: []string{"string", "string"}, ret0: "array[int]", ret1: "string", jsKey: "richtext_exportbytes", async: false},
	StdFunc{mod: "richtext", name: "openFile", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "richtext_openfile", async: true},
	StdFunc{mod: "richtext", name: "importBytes", params: []string{"string", "any", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_importbytes", async: true},
	StdFunc{mod: "richtext", name: "saveAs", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_saveas", async: false},
	StdFunc{mod: "richtext", name: "importFormat", params: []string{"string", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_import", async: false},
	StdFunc{mod: "richtext", name: "registerPlugin", params: []string{"string", "any"}, ret0: "bool", ret1: "string", jsKey: "richtext_register", async: false},
	StdFunc{mod: "richtext", name: "plugins", params: []string{}, ret0: "array[string]", ret1: "string", jsKey: "richtext_plugins", async: false},
	StdFunc{mod: "richtext", name: "status", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "richtext_status", async: false},
	StdFunc{mod: "richtext", name: "selection", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "richtext_selection", async: false},
	StdFunc{mod: "richtext", name: "setToolbar", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_settoolbar", async: false},
	StdFunc{mod: "richtext", name: "setTemplate", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_settemplate", async: false},
	StdFunc{mod: "richtext", name: "setTheme", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "richtext_settheme", async: false},
	StdFunc{mod: "richtext", name: "destroy", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "richtext_destroy", async: false},

	StdFunc{mod: "table", name: "display", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "table_display", async: false},
	StdFunc{mod: "table", name: "data", params: []string{"string"}, ret0: "array[array[string]]", ret1: "string", jsKey: "table_data", async: false},
	StdFunc{mod: "table", name: "header", params: []string{"string"}, ret0: "array[string]", ret1: "string", jsKey: "table_header", async: false},
	StdFunc{mod: "table", name: "select", params: []string{"string", "array[int]"}, ret0: "array[array[string]]", ret1: "string", jsKey: "table_select", async: false},
	StdFunc{mod: "table", name: "sort", params: []string{"string", "string", "string"}, ret0: "int", ret1: "string", jsKey: "table_sort", async: false},
	StdFunc{mod: "table", name: "filter", params: []string{"string", "string", "string"}, ret0: "int", ret1: "string", jsKey: "table_filter", async: false},
	StdFunc{mod: "table", name: "filterText", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "table_filtertext", async: false},
	StdFunc{mod: "table", name: "page", params: []string{"string", "int"}, ret0: "int", ret1: "string", jsKey: "table_page", async: false},
	StdFunc{mod: "table", name: "setPageSize", params: []string{"string", "int"}, ret0: "int", ret1: "string", jsKey: "table_setpagesize", async: false},
	StdFunc{mod: "table", name: "selectedRows", params: []string{"string"}, ret0: "array[int]", ret1: "string", jsKey: "table_selectedrows", async: false},
	StdFunc{mod: "table", name: "selectedCells", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "table_selectedcells", async: false},
	StdFunc{mod: "table", name: "selectedColumns", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "table_selectedcolumns", async: false},
	StdFunc{mod: "table", name: "lastChange", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "table_lastchange", async: false},
	StdFunc{mod: "table", name: "slice", params: []string{"string", "int", "int"}, ret0: "array[array[string]]", ret1: "string", jsKey: "table_slice", async: false},
	StdFunc{mod: "table", name: "importText", params: []string{"string", "string", "string"}, ret0: "int", ret1: "string", jsKey: "table_importtext", async: false},
	StdFunc{mod: "table", name: "exportText", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "table_exporttext", async: false},
	StdFunc{mod: "table", name: "json", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "table_json", async: false},
	StdFunc{mod: "table", name: "clear", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "table_clear", async: false},
	StdFunc{mod: "table", name: "clearPersisted", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "table_clearpersisted", async: false},
	// chart (core: always available, acts on the rendered state chart named with `name:`)
	StdFunc{mod: "chart", name: "setTag", params: []string{"string", "string", "string", "string", "string"}, ret0: "bool", ret1: "string", jsKey: "chart_settag", async: false},
	StdFunc{mod: "chart", name: "update", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_update", async: false},
	StdFunc{mod: "chart", name: "updateMany", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "chart_updatemany", async: false},
	StdFunc{mod: "chart", name: "get", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_get", async: false},
	StdFunc{mod: "chart", name: "elements", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "chart_elements", async: false},
	StdFunc{mod: "chart", name: "find", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_find", async: false},
	StdFunc{mod: "chart", name: "getAttr", params: []string{"string", "string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getattr", async: false},
	StdFunc{mod: "chart", name: "setAttr", params: []string{"string", "string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setattr", async: false},
	StdFunc{mod: "chart", name: "getLabel", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getlabel", async: false},
	StdFunc{mod: "chart", name: "setLabel", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setlabel", async: false},
	StdFunc{mod: "chart", name: "getText", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettext", async: false},
	StdFunc{mod: "chart", name: "setText", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settext", async: false},
	StdFunc{mod: "chart", name: "getHint", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gethint", async: false},
	StdFunc{mod: "chart", name: "setHint", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_sethint", async: false},
	StdFunc{mod: "chart", name: "getDescription", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getdescription", async: false},
	StdFunc{mod: "chart", name: "setDescription", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setdescription", async: false},
	StdFunc{mod: "chart", name: "getColor", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getcolor", async: false},
	StdFunc{mod: "chart", name: "setColor", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setcolor", async: false},
	StdFunc{mod: "chart", name: "getTagColor", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettagcolor", async: false},
	StdFunc{mod: "chart", name: "setTagColor", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settagcolor", async: false},
	StdFunc{mod: "chart", name: "getTextColor", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettextcolor", async: false},
	StdFunc{mod: "chart", name: "setTextColor", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settextcolor", async: false},
	StdFunc{mod: "chart", name: "getShape", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getshape", async: false},
	StdFunc{mod: "chart", name: "setShape", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setshape", async: false},
	StdFunc{mod: "chart", name: "getRouting", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getrouting", async: false},
	StdFunc{mod: "chart", name: "setRouting", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setrouting", async: false},
	StdFunc{mod: "chart", name: "getBorderColor", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getbordercolor", async: false},
	StdFunc{mod: "chart", name: "setBorderColor", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setbordercolor", async: false},
	StdFunc{mod: "chart", name: "getBorderStyle", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getborderstyle", async: false},
	StdFunc{mod: "chart", name: "setBorderStyle", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setborderstyle", async: false},
	StdFunc{mod: "chart", name: "getBorderWidth", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getborderwidth", async: false},
	StdFunc{mod: "chart", name: "setBorderWidth", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setborderwidth", async: false},
	StdFunc{mod: "chart", name: "getCornerRadius", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getcornerradius", async: false},
	StdFunc{mod: "chart", name: "setCornerRadius", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setcornerradius", async: false},
	StdFunc{mod: "chart", name: "getWidth", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getwidth", async: false},
	StdFunc{mod: "chart", name: "setWidth", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setwidth", async: false},
	StdFunc{mod: "chart", name: "getHeight", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getheight", async: false},
	StdFunc{mod: "chart", name: "setHeight", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setheight", async: false},
	StdFunc{mod: "chart", name: "getX", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getx", async: false},
	StdFunc{mod: "chart", name: "setX", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setx", async: false},
	StdFunc{mod: "chart", name: "getY", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gety", async: false},
	StdFunc{mod: "chart", name: "setY", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_sety", async: false},
	StdFunc{mod: "chart", name: "getRotation", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getrotation", async: false},
	StdFunc{mod: "chart", name: "setRotation", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setrotation", async: false},
	StdFunc{mod: "chart", name: "getDashStyle", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getdashstyle", async: false},
	StdFunc{mod: "chart", name: "setDashStyle", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setdashstyle", async: false},
	StdFunc{mod: "chart", name: "getStartMarker", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getstartmarker", async: false},
	StdFunc{mod: "chart", name: "setStartMarker", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setstartmarker", async: false},
	StdFunc{mod: "chart", name: "getEndMarker", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getendmarker", async: false},
	StdFunc{mod: "chart", name: "setEndMarker", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setendmarker", async: false},
	StdFunc{mod: "chart", name: "setFrom", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setfrom", async: false},
	StdFunc{mod: "chart", name: "setTo", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setto", async: false},
	StdFunc{mod: "chart", name: "getTagKind", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettagkind", async: false},
	StdFunc{mod: "chart", name: "setTagKind", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settagkind", async: false},
	StdFunc{mod: "chart", name: "getTagBox", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettagbox", async: false},
	StdFunc{mod: "chart", name: "setTagBox", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settagbox", async: false},
	StdFunc{mod: "chart", name: "getTagAlign", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettagalign", async: false},
	StdFunc{mod: "chart", name: "setTagAlign", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settagalign", async: false},
	StdFunc{mod: "chart", name: "getTagLeader", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettagleader", async: false},
	StdFunc{mod: "chart", name: "setTagLeader", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settagleader", async: false},
	StdFunc{mod: "chart", name: "getWaypoints", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getwaypoints", async: false},
	StdFunc{mod: "chart", name: "setWaypoints", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setwaypoints", async: false},
	StdFunc{mod: "chart", name: "getTextPlacement", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettextplacement", async: false},
	StdFunc{mod: "chart", name: "setTextPlacement", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settextplacement", async: false},
	StdFunc{mod: "chart", name: "getTagSide", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_gettagside", async: false},
	StdFunc{mod: "chart", name: "setTagSide", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_settagside", async: false},
	StdFunc{mod: "chart", name: "getBorderDashLength", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getborderdashlength", async: false},
	StdFunc{mod: "chart", name: "setBorderDashLength", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setborderdashlength", async: false},
	StdFunc{mod: "chart", name: "getBorderDashGap", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getborderdashgap", async: false},
	StdFunc{mod: "chart", name: "setBorderDashGap", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setborderdashgap", async: false},
	StdFunc{mod: "chart", name: "getSnap", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getsnap", async: false},
	StdFunc{mod: "chart", name: "setSnap", params: []string{"string", "string", "any"}, ret0: "bool", ret1: "string", jsKey: "chart_setsnap", async: false},
	StdFunc{mod: "chart", name: "getKind", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getkind", async: false},
	StdFunc{mod: "chart", name: "getFrom", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getfrom", async: false},
	StdFunc{mod: "chart", name: "getTo", params: []string{"string", "string"}, ret0: "string", ret1: "string", jsKey: "chart_getto", async: false},
	StdFunc{mod: "chart", name: "focusRow", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "chart_focusrow", async: false},
	StdFunc{mod: "chart", name: "peekEvents", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "chart_peekevents", async: false},
	StdFunc{mod: "chart", name: "pollEvents", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "chart_pollevents", async: false},
	StdFunc{mod: "chart", name: "nextEvent", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "chart_nextevent", async: false},
	StdFunc{mod: "chart", name: "clearEvents", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "chart_clearevents", async: false},
	StdFunc{mod: "chart", name: "clearPersisted", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "chart_clearpersisted", async: false},
	StdFunc{mod: "chart", name: "eventCount", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "chart_eventcount", async: false},
	// tree (core: always available, acts on the rendered tree named with `name:`)
	StdFunc{mod: "tree", name: "importJson", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_importjson", async: false},
	StdFunc{mod: "tree", name: "importXml", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_importxml", async: false},
	StdFunc{mod: "tree", name: "importJsonFile", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_importjsonfile", async: true},
	StdFunc{mod: "tree", name: "importXmlFile", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_importxmlfile", async: true},
	StdFunc{mod: "tree", name: "exportJson", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "tree_exportjson", async: false},
	StdFunc{mod: "tree", name: "exportXml", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "tree_exportxml", async: false},
	StdFunc{mod: "tree", name: "downloadJson", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_downloadjson", async: false},
	StdFunc{mod: "tree", name: "downloadXml", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_downloadxml", async: false},
	StdFunc{mod: "tree", name: "clear", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "tree_clear", async: false},
	StdFunc{mod: "tree", name: "clearPersisted", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "tree_clearpersisted", async: false},
	StdFunc{mod: "tree", name: "expandAll", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "tree_expandall", async: false},
	StdFunc{mod: "tree", name: "collapseAll", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "tree_collapseall", async: false},
	StdFunc{mod: "tree", name: "expandTo", params: []string{"string", "int"}, ret0: "int", ret1: "string", jsKey: "tree_expandto", async: false},
	StdFunc{mod: "tree", name: "setChildren", params: []string{"string", "string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_setchildren", async: false},
	StdFunc{mod: "tree", name: "setChildrenError", params: []string{"string", "string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_setchildrenerror", async: false},
	StdFunc{mod: "tree", name: "retryChildren", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_retrychildren", async: false},
	StdFunc{mod: "tree", name: "lastMove", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "tree_lastmove", async: false},
	StdFunc{mod: "tree", name: "lastChange", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "tree_lastchange", async: false},
	StdFunc{mod: "tree", name: "undoMove", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "tree_undomove", async: false},
	StdFunc{mod: "tree", name: "setColumns", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_setcolumns", async: false},
	StdFunc{mod: "tree", name: "getColumns", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "tree_getcolumns", async: false},
	StdFunc{mod: "tree", name: "moveColumn", params: []string{"string", "int", "int"}, ret0: "int", ret1: "string", jsKey: "tree_movecolumn", async: false},
	StdFunc{mod: "tree", name: "setCell", params: []string{"string", "string", "string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_setcell", async: false},
	StdFunc{mod: "tree", name: "selectedKeys", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "tree_selectedkeys", async: false},
	StdFunc{mod: "tree", name: "checkedKeys", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "tree_checkedkeys", async: false},
	StdFunc{mod: "tree", name: "search", params: []string{"string", "string"}, ret0: "int", ret1: "string", jsKey: "tree_search", async: false},
}

// IsStdModule reports whether a module name can be imported from "std".
func IsStdModule(name string) bool {
	for _, m := range StdModules {
		if m == name {
			return true
		}
	}
	return false
}

// StdLookup finds the StdFunc for a fully-qualified name, e.g. "http.get".
func StdLookup(full string) (*StdFunc, bool) {
	dot := 0
	for dot < len(full) && full[dot] != '.' {
		dot++
	}
	if dot == len(full) {
		return nil, false
	}
	modPart := full[:dot]
	namePart := full[dot+1:]
	for i := 0; i < len(StdFuncs); i++ {
		f := &StdFuncs[i]
		if f.mod == modPart && f.name == namePart {
			return f, true
		}
	}
	return nil, false
}
