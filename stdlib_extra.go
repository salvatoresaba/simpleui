package sui

import "strings"

// stdlib_extra.go registers the data/utility std modules added after v0.1.
//
// These modules are pure or in-memory by design: they never touch the network,
// the filesystem or the process environment. The browser/desktop runtime lives
// in stdlib_extra_runtime.go and is emitted only when one of these modules is
// imported. Every function follows the standard (value, err) convention.

var extraStdModules = map[string]bool{
	"bytes":        true,
	"bin":          true,
	"files":        true,
	"mime":         true,
	"color":        true,
	"units":        true,
	"print":        true,
	"history":      true,
	"zip":          true,
	"archive":      true,
	"csv":          true,
	"env":          true,
	"capabilities": true,
	"events":       true,
	"bus":          true,
	"log":          true,
	"telemetry":    true,
}

var extraStdFuncs = []StdFunc{
	// bytes
	StdFunc{mod: "bytes", name: "fromString", params: []string{"string", "string"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_fromstring", optional: 1},
	StdFunc{mod: "bytes", name: "toString", params: []string{"any", "string"}, ret0: "string", ret1: "string", jsKey: "bytes_tostring", optional: 1},
	StdFunc{mod: "bytes", name: "fromBase64", params: []string{"string"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_frombase64"},
	StdFunc{mod: "bytes", name: "toBase64", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "bytes_tobase64"},
	StdFunc{mod: "bytes", name: "fromHex", params: []string{"string"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_fromhex"},
	StdFunc{mod: "bytes", name: "toHex", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "bytes_tohex"},
	StdFunc{mod: "bytes", name: "concat", params: []string{"any", "any"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_concat"},
	StdFunc{mod: "bytes", name: "slice", params: []string{"any", "int", "int"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_slice", optional: 1},
	StdFunc{mod: "bytes", name: "indexOf", params: []string{"any", "any", "int"}, ret0: "int", ret1: "string", jsKey: "bytes_indexof", optional: 1},
	StdFunc{mod: "bytes", name: "equal", params: []string{"any", "any"}, ret0: "bool", ret1: "string", jsKey: "bytes_equal"},
	StdFunc{mod: "bytes", name: "length", params: []string{"any"}, ret0: "int", ret1: "string", jsKey: "bytes_length"},
	StdFunc{mod: "bytes", name: "fill", params: []string{"int", "int"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_fill"},
	StdFunc{mod: "bytes", name: "reverse", params: []string{"any"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_reverse"},
	StdFunc{mod: "bytes", name: "xor", params: []string{"any", "any"}, ret0: "array[int]", ret1: "string", jsKey: "bytes_xor"},
	// bin
	StdFunc{mod: "bin", name: "u8", params: []string{"any", "int"}, ret0: "int", ret1: "string", jsKey: "bin_u8"},
	StdFunc{mod: "bin", name: "u16", params: []string{"any", "int", "bool"}, ret0: "int", ret1: "string", jsKey: "bin_u16", optional: 1},
	StdFunc{mod: "bin", name: "u32", params: []string{"any", "int", "bool"}, ret0: "int", ret1: "string", jsKey: "bin_u32", optional: 1},
	StdFunc{mod: "bin", name: "i8", params: []string{"any", "int"}, ret0: "int", ret1: "string", jsKey: "bin_i8"},
	StdFunc{mod: "bin", name: "i16", params: []string{"any", "int", "bool"}, ret0: "int", ret1: "string", jsKey: "bin_i16", optional: 1},
	StdFunc{mod: "bin", name: "i32", params: []string{"any", "int", "bool"}, ret0: "int", ret1: "string", jsKey: "bin_i32", optional: 1},
	StdFunc{mod: "bin", name: "f32", params: []string{"any", "int", "bool"}, ret0: "float", ret1: "string", jsKey: "bin_f32", optional: 1},
	StdFunc{mod: "bin", name: "f64", params: []string{"any", "int", "bool"}, ret0: "float", ret1: "string", jsKey: "bin_f64", optional: 1},
	StdFunc{mod: "bin", name: "writeU8", params: []string{"int"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writeu8"},
	StdFunc{mod: "bin", name: "writeU16", params: []string{"int", "bool"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writeu16", optional: 1},
	StdFunc{mod: "bin", name: "writeU32", params: []string{"int", "bool"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writeu32", optional: 1},
	StdFunc{mod: "bin", name: "writeI8", params: []string{"int"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writei8"},
	StdFunc{mod: "bin", name: "writeI16", params: []string{"int", "bool"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writei16", optional: 1},
	StdFunc{mod: "bin", name: "writeI32", params: []string{"int", "bool"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writei32", optional: 1},
	StdFunc{mod: "bin", name: "writeF32", params: []string{"float", "bool"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writef32", optional: 1},
	StdFunc{mod: "bin", name: "writeF64", params: []string{"float", "bool"}, ret0: "array[int]", ret1: "string", jsKey: "bin_writef64", optional: 1},
	// files
	StdFunc{mod: "files", name: "name", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "files_name"},
	StdFunc{mod: "files", name: "stem", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "files_stem"},
	StdFunc{mod: "files", name: "ext", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "files_ext"},
	StdFunc{mod: "files", name: "split", params: []string{"string"}, ret0: "map[string]string", ret1: "string", jsKey: "files_split"},
	StdFunc{mod: "files", name: "sanitize", params: []string{"string", "int"}, ret0: "string", ret1: "string", jsKey: "files_sanitize", optional: 1},
	StdFunc{mod: "files", name: "validate", params: []string{"string", "any"}, ret0: "bool", ret1: "string", jsKey: "files_validate", optional: 1},
	StdFunc{mod: "files", name: "isHidden", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "files_ishidden"},
	StdFunc{mod: "files", name: "extAllowed", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "files_extallowed"},
	// The single exception in the data modules: it reads the bytes of a file
	// the user selected or dropped in this app (never a path, never the
	// network). The desktop target uses fs.readBytes(path) instead.
	StdFunc{mod: "files", name: "readBytes", params: []string{"string"}, ret0: "array[int]", ret1: "string", jsKey: "files_readbytes", async: true},
	// files.documentUrl builds a browser URL for a file the user selected,
	// so the PDF viewer can display it without a backend. The bytes are
	// checked and the type is forced to application/pdf: the native viewer
	// must never receive a document the browser could execute. The desktop
	// target uses fs.documentUrl(path).
	StdFunc{mod: "files", name: "documentUrl", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "files_documenturl", async: true},
	// mime
	StdFunc{mod: "mime", name: "fromName", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "mime_fromname"},
	StdFunc{mod: "mime", name: "fromBytes", params: []string{"any"}, ret0: "string", ret1: "string", jsKey: "mime_frombytes"},
	StdFunc{mod: "mime", name: "extension", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "mime_extension"},
	StdFunc{mod: "mime", name: "isAllowed", params: []string{"string", "string"}, ret0: "bool", ret1: "string", jsKey: "mime_isallowed"},
	StdFunc{mod: "mime", name: "validate", params: []string{"string", "string", "int", "any"}, ret0: "bool", ret1: "string", jsKey: "mime_validate", optional: 1},
	// color
	StdFunc{mod: "color", name: "parse", params: []string{"string"}, ret0: "any", ret1: "string", jsKey: "color_parse"},
	StdFunc{mod: "color", name: "format", params: []string{"any", "string"}, ret0: "string", ret1: "string", jsKey: "color_format", optional: 1},
	StdFunc{mod: "color", name: "rgb", params: []string{"int", "int", "int", "float"}, ret0: "any", ret1: "string", jsKey: "color_rgb", optional: 1},
	StdFunc{mod: "color", name: "hsl", params: []string{"float", "float", "float", "float"}, ret0: "any", ret1: "string", jsKey: "color_hsl", optional: 1},
	StdFunc{mod: "color", name: "mix", params: []string{"any", "any", "float"}, ret0: "any", ret1: "string", jsKey: "color_mix"},
	StdFunc{mod: "color", name: "withAlpha", params: []string{"any", "float"}, ret0: "any", ret1: "string", jsKey: "color_withalpha"},
	StdFunc{mod: "color", name: "luminance", params: []string{"any"}, ret0: "float", ret1: "string", jsKey: "color_luminance"},
	StdFunc{mod: "color", name: "contrast", params: []string{"any", "any"}, ret0: "float", ret1: "string", jsKey: "color_contrast"},
	StdFunc{mod: "color", name: "isDark", params: []string{"any"}, ret0: "bool", ret1: "string", jsKey: "color_isdark"},
	// units
	StdFunc{mod: "units", name: "parse", params: []string{"string"}, ret0: "any", ret1: "string", jsKey: "units_parse"},
	StdFunc{mod: "units", name: "convert", params: []string{"float", "string", "string"}, ret0: "float", ret1: "string", jsKey: "units_convert"},
	StdFunc{mod: "units", name: "toBase", params: []string{"float", "string"}, ret0: "float", ret1: "string", jsKey: "units_tobase"},
	StdFunc{mod: "units", name: "fromBase", params: []string{"float", "string"}, ret0: "float", ret1: "string", jsKey: "units_frombase"},
	StdFunc{mod: "units", name: "format", params: []string{"float", "string", "int"}, ret0: "string", ret1: "string", jsKey: "units_format", optional: 1},
	// print
	StdFunc{mod: "print", name: "pxToMm", params: []string{"float"}, ret0: "float", ret1: "string", jsKey: "print_pxtomm"},
	StdFunc{mod: "print", name: "mmToPx", params: []string{"float"}, ret0: "float", ret1: "string", jsKey: "print_mmtopx"},
	StdFunc{mod: "print", name: "ptToMm", params: []string{"float"}, ret0: "float", ret1: "string", jsKey: "print_pttomm"},
	StdFunc{mod: "print", name: "paper", params: []string{"string", "string"}, ret0: "map[string]any", ret1: "string", jsKey: "print_paper", optional: 1},
	StdFunc{mod: "print", name: "fitBox", params: []string{"float", "float", "float", "float"}, ret0: "map[string]any", ret1: "string", jsKey: "print_fitbox"},
	// history
	StdFunc{mod: "history", name: "push", params: []string{"string", "any", "int"}, ret0: "int", ret1: "string", jsKey: "history_push", optional: 1},
	StdFunc{mod: "history", name: "undo", params: []string{"string"}, ret0: "any", ret1: "string", jsKey: "history_undo"},
	StdFunc{mod: "history", name: "redo", params: []string{"string"}, ret0: "any", ret1: "string", jsKey: "history_redo"},
	StdFunc{mod: "history", name: "current", params: []string{"string"}, ret0: "any", ret1: "string", jsKey: "history_current"},
	StdFunc{mod: "history", name: "canUndo", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "history_canundo"},
	StdFunc{mod: "history", name: "canRedo", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "history_canredo"},
	StdFunc{mod: "history", name: "clear", params: []string{"string"}, ret0: "null", ret1: "string", jsKey: "history_clear"},
	StdFunc{mod: "history", name: "list", params: []string{}, ret0: "array[string]", ret1: "string", jsKey: "history_list"},
	StdFunc{mod: "history", name: "size", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "history_size"},
	StdFunc{mod: "history", name: "jump", params: []string{"string", "int"}, ret0: "any", ret1: "string", jsKey: "history_jump"},
	// zip
	StdFunc{mod: "zip", name: "create", params: []string{}, ret0: "any", ret1: "string", jsKey: "zip_create"},
	StdFunc{mod: "zip", name: "add", params: []string{"any", "string", "any"}, ret0: "int", ret1: "string", jsKey: "zip_add"},
	StdFunc{mod: "zip", name: "addText", params: []string{"any", "string", "string"}, ret0: "int", ret1: "string", jsKey: "zip_addtext"},
	StdFunc{mod: "zip", name: "finish", params: []string{"any"}, ret0: "array[int]", ret1: "string", jsKey: "zip_finish"},
	StdFunc{mod: "zip", name: "open", params: []string{"any"}, ret0: "any", ret1: "string", jsKey: "zip_open"},
	StdFunc{mod: "zip", name: "list", params: []string{"any"}, ret0: "array[any]", ret1: "string", jsKey: "zip_list"},
	StdFunc{mod: "zip", name: "read", params: []string{"any", "string"}, ret0: "array[int]", ret1: "string", jsKey: "zip_read"},
	StdFunc{mod: "zip", name: "readText", params: []string{"any", "string"}, ret0: "string", ret1: "string", jsKey: "zip_readtext"},
	StdFunc{mod: "zip", name: "has", params: []string{"any", "string"}, ret0: "bool", ret1: "string", jsKey: "zip_has"},
	StdFunc{mod: "zip", name: "count", params: []string{"any"}, ret0: "int", ret1: "string", jsKey: "zip_count"},
	// archive
	StdFunc{mod: "archive", name: "pack", params: []string{"any", "string"}, ret0: "array[int]", ret1: "string", jsKey: "archive_pack", optional: 1},
	StdFunc{mod: "archive", name: "unpack", params: []string{"any", "string"}, ret0: "map[string]any", ret1: "string", jsKey: "archive_unpack", optional: 1},
	StdFunc{mod: "archive", name: "packText", params: []string{"any", "string"}, ret0: "array[int]", ret1: "string", jsKey: "archive_packtext", optional: 1},
	StdFunc{mod: "archive", name: "unpackText", params: []string{"any", "string"}, ret0: "map[string]string", ret1: "string", jsKey: "archive_unpacktext", optional: 1},
	StdFunc{mod: "archive", name: "info", params: []string{"any"}, ret0: "array[any]", ret1: "string", jsKey: "archive_info"},
	// csv
	StdFunc{mod: "csv", name: "parse", params: []string{"string", "any"}, ret0: "array[any]", ret1: "string", jsKey: "csv_parse", optional: 1},
	StdFunc{mod: "csv", name: "parseObjects", params: []string{"string", "any"}, ret0: "array[any]", ret1: "string", jsKey: "csv_parseobjects", optional: 1},
	StdFunc{mod: "csv", name: "stringify", params: []string{"any", "any"}, ret0: "string", ret1: "string", jsKey: "csv_stringify", optional: 1},
	StdFunc{mod: "csv", name: "stringifyObjects", params: []string{"any", "any"}, ret0: "string", ret1: "string", jsKey: "csv_stringifyobjects", optional: 1},
	StdFunc{mod: "csv", name: "validate", params: []string{"string", "any"}, ret0: "bool", ret1: "string", jsKey: "csv_validate", optional: 1},
	// env
	StdFunc{mod: "env", name: "target", params: []string{}, ret0: "string", ret1: "string", jsKey: "env_target"},
	StdFunc{mod: "env", name: "platform", params: []string{}, ret0: "string", ret1: "string", jsKey: "env_platform"},
	StdFunc{mod: "env", name: "get", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "env_get"},
	StdFunc{mod: "env", name: "has", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "env_has"},
	StdFunc{mod: "env", name: "all", params: []string{}, ret0: "map[string]string", ret1: "string", jsKey: "env_all"},
	StdFunc{mod: "env", name: "language", params: []string{}, ret0: "string", ret1: "string", jsKey: "env_language"},
	StdFunc{mod: "env", name: "timezone", params: []string{}, ret0: "string", ret1: "string", jsKey: "env_timezone"},
	// capabilities
	StdFunc{mod: "capabilities", name: "has", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "capabilities_has"},
	StdFunc{mod: "capabilities", name: "check", params: []string{"string"}, ret0: "bool", ret1: "string", jsKey: "capabilities_check"},
	StdFunc{mod: "capabilities", name: "list", params: []string{}, ret0: "array[string]", ret1: "string", jsKey: "capabilities_list"},
	StdFunc{mod: "capabilities", name: "target", params: []string{}, ret0: "string", ret1: "string", jsKey: "capabilities_target"},
	// events
	StdFunc{mod: "events", name: "on", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "events_on"},
	StdFunc{mod: "events", name: "once", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "events_once"},
	StdFunc{mod: "events", name: "off", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "events_off"},
	StdFunc{mod: "events", name: "emit", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "events_emit"},
	StdFunc{mod: "events", name: "clear", params: []string{"string"}, ret0: "null", ret1: "string", jsKey: "events_clear"},
	StdFunc{mod: "events", name: "count", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "events_count"},
	StdFunc{mod: "events", name: "names", params: []string{}, ret0: "array[string]", ret1: "string", jsKey: "events_names"},
	// bus (alias of events)
	StdFunc{mod: "bus", name: "on", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "bus_on"},
	StdFunc{mod: "bus", name: "once", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "bus_once"},
	StdFunc{mod: "bus", name: "off", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "bus_off"},
	StdFunc{mod: "bus", name: "emit", params: []string{"string", "any"}, ret0: "int", ret1: "string", jsKey: "bus_emit"},
	StdFunc{mod: "bus", name: "clear", params: []string{"string"}, ret0: "null", ret1: "string", jsKey: "bus_clear"},
	StdFunc{mod: "bus", name: "count", params: []string{"string"}, ret0: "int", ret1: "string", jsKey: "bus_count"},
	StdFunc{mod: "bus", name: "names", params: []string{}, ret0: "array[string]", ret1: "string", jsKey: "bus_names"},
	// log
	StdFunc{mod: "log", name: "debug", params: []string{"string", "any"}, ret0: "null", ret1: "string", jsKey: "log_debug", optional: 1},
	StdFunc{mod: "log", name: "info", params: []string{"string", "any"}, ret0: "null", ret1: "string", jsKey: "log_info", optional: 1},
	StdFunc{mod: "log", name: "warn", params: []string{"string", "any"}, ret0: "null", ret1: "string", jsKey: "log_warn", optional: 1},
	StdFunc{mod: "log", name: "error", params: []string{"string", "any"}, ret0: "null", ret1: "string", jsKey: "log_error", optional: 1},
	StdFunc{mod: "log", name: "setLevel", params: []string{"string"}, ret0: "null", ret1: "string", jsKey: "log_setlevel"},
	StdFunc{mod: "log", name: "level", params: []string{}, ret0: "string", ret1: "string", jsKey: "log_level"},
	StdFunc{mod: "log", name: "clear", params: []string{}, ret0: "null", ret1: "string", jsKey: "log_clear"},
	StdFunc{mod: "log", name: "entries", params: []string{"int"}, ret0: "array[any]", ret1: "string", jsKey: "log_entries", optional: 1},
	StdFunc{mod: "log", name: "export", params: []string{}, ret0: "string", ret1: "string", jsKey: "log_export"},
	StdFunc{mod: "log", name: "redact", params: []string{"string"}, ret0: "string", ret1: "string", jsKey: "log_redact"},
	StdFunc{mod: "log", name: "redactObject", params: []string{"any"}, ret0: "any", ret1: "string", jsKey: "log_redactobject"},
	// telemetry
	StdFunc{mod: "telemetry", name: "counter", params: []string{"string", "float"}, ret0: "float", ret1: "string", jsKey: "telemetry_counter", optional: 1},
	StdFunc{mod: "telemetry", name: "gauge", params: []string{"string", "float"}, ret0: "float", ret1: "string", jsKey: "telemetry_gauge"},
	StdFunc{mod: "telemetry", name: "timing", params: []string{"string", "float"}, ret0: "float", ret1: "string", jsKey: "telemetry_timing"},
	StdFunc{mod: "telemetry", name: "snapshot", params: []string{}, ret0: "map[string]any", ret1: "string", jsKey: "telemetry_snapshot"},
	StdFunc{mod: "telemetry", name: "reset", params: []string{}, ret0: "null", ret1: "string", jsKey: "telemetry_reset"},
	StdFunc{mod: "telemetry", name: "names", params: []string{}, ret0: "array[string]", ret1: "string", jsKey: "telemetry_names"},
}

func init() {
	StdFuncs = append(StdFuncs, extraStdFuncs...)
}

// isStdlibExtraModule reports whether name belongs to the pure/in-memory std
// modules implemented by the shared runtime prelude.
func isStdlibExtraModule(name string) bool {
	return extraStdModules[name]
}

// stdlibExtraBody builds the generated __SUI_STD entry for an extra module.
// Returning an empty string means the key is not one of these modules.
func stdlibExtraBody(key string) string {
	for i := 0; i < len(extraStdFuncs); i++ {
		f := &extraStdFuncs[i]
		if f.jsKey != key {
			continue
		}
		return "function () { return __SUI_STD_EXTRA." + f.mod + "." + f.name + ".apply(null, arguments); }"
	}
	return ""
}

// extraRuntimeModuleImported reports whether the program imports at least one
// module that needs the shared stdlib extra runtime prelude.
func extraRuntimeModuleImported(p *Prog) bool {
	if p == nil {
		return false
	}
	for i := 0; i < len(p.imports); i++ {
		im := p.imports[i]
		if im.kind != "std" {
			continue
		}
		for j := 0; j < len(im.names); j++ {
			if isStdlibExtraModule(strings.TrimSpace(im.names[j])) {
				return true
			}
		}
	}
	return false
}
