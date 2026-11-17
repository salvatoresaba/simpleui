package main

import "strconv"
import "strings"

// builtins.go - language-level built-in functions (no `import` needed).
//
// These are the batteries for data manipulation most apps need: collection
// sizing/extension, string helpers and number conversions. Typing rules live
// in the typechecker (some are polymorphic: len/append/str/int); the JS
// generator translates the caller into the JS expression in `js`.
//
// Placeholders in js: {0} {1} {2} = transpiled argument expressions.

type Builtin struct {
	name string
	js   string
}

var Builtins = []Builtin{
	Builtin{name: "assert", js: "__sui_assert({0}, {1})"},
	// ---- locale formatting (Intl) ----
	Builtin{name: "intl.date", js: "__sui_intl_date({0}, {1}, {2}, {3})"},
	Builtin{name: "intl.time", js: "__sui_intl_time({0}, {1}, {2}, {3})"},
	Builtin{name: "intl.datetime", js: "__sui_intl_datetime({0}, {1}, {2}, {3})"},
	Builtin{name: "intl.number", js: "__sui_intl_number({0}, {1}, {2})"},
	Builtin{name: "intl.currency", js: "__sui_intl_currency({0}, {1}, {2})"},
	Builtin{name: "intl.percent", js: "__sui_intl_percent({0}, {1}, {2})"},
	Builtin{name: "intl.relative", js: "__sui_intl_relative({0}, {1}, {2}, {3})"},
	Builtin{name: "intl.timezone", js: "__sui_intl_timezone()"},
	Builtin{name: "len", js: "{0}.length"},
	Builtin{name: "append", js: "[...{0}, {1}]"},
	Builtin{name: "upper", js: "{0}.toUpperCase()"},
	Builtin{name: "lower", js: "{0}.toLowerCase()"},
	Builtin{name: "trim", js: "{0}.trim()"},
	Builtin{name: "cssSafe", js: "__sui_css({0})"},
	Builtin{name: "contains", js: "{0}.includes({1})"},
	Builtin{name: "split", js: "{0}.split({1})"},
	Builtin{name: "join", js: "{0}.join({1})"},
	Builtin{name: "substring", js: "{0}.substring({1}, {1}+{2})"},
	Builtin{name: "str", js: "String({0})"},
	Builtin{name: "int", js: "Math.trunc({0})"},
	Builtin{name: "float", js: "Number({0})"},
	Builtin{name: "keys", js: "Object.keys({0})"},
	Builtin{name: "values", js: "Object.values({0})"},
	Builtin{name: "merge", js: "Object.assign({}, {0}, {1})"},
	Builtin{name: "has", js: "Object.prototype.hasOwnProperty.call({0}, {1})"},
	// ---- strings ----
	Builtin{name: "replace", js: "{0}.replaceAll({1}, {2})"},
	Builtin{name: "replaceFirst", js: "{0}.replace({1}, {2})"},
	Builtin{name: "replaceRegex", js: "__sui_replace_regex({0}, {1}, {2})"},
	Builtin{name: "match", js: "__sui_match({0}, {1}, {2})"},
	Builtin{name: "matchAll", js: "__sui_match_all({0}, {1}, {2})"},
	Builtin{name: "search", js: "__sui_search({0}, {1})"},
	Builtin{name: "test", js: "__sui_test({0}, {1})"},
	Builtin{name: "splitRegex", js: "__sui_split_regex({0}, {1}, {2})"},
	Builtin{name: "extract", js: "__sui_extract({0}, {1})"},
	Builtin{name: "extractAll", js: "__sui_extract_all({0}, {1})"},
	Builtin{name: "startsWith", js: "{0}.startsWith({1})"},
	Builtin{name: "endsWith", js: "{0}.endsWith({1})"},
	Builtin{name: "indexOf", js: "{0}.indexOf({1})"},
	Builtin{name: "lastIndexOf", js: "{0}.lastIndexOf({1})"},
	Builtin{name: "slice", js: "{0}.slice({1}, {2})"},
	Builtin{name: "repeat", js: "{0}.repeat({1})"},
	Builtin{name: "padLeft", js: "{0}.padStart({1}, {2})"},
	Builtin{name: "padRight", js: "{0}.padEnd({1}, {2})"},
	Builtin{name: "trimLeft", js: "{0}.trimStart()"},
	Builtin{name: "trimRight", js: "{0}.trimEnd()"},
	Builtin{name: "chars", js: "[...{0}]"},
	Builtin{name: "fmt", js: "{0}.toFixed({1})"},
	// ---- collections ----
	Builtin{name: "concat", js: "[...{0}, ...{1}]"},
	Builtin{name: "remove", js: "[...{0}.slice(0, {1}), ...{0}.slice({1}+1)]"},
	Builtin{name: "includes", js: "{0}.includes({1})"},
	Builtin{name: "sort", js: "[...{0}].sort()"},
	Builtin{name: "reverse", js: "[...{0}].reverse()"},
	Builtin{name: "removeKey", js: "Object.fromEntries(Object.entries({0}).filter((e) => e[0] !== {1}))"},
	Builtin{name: "moveItem", js: "__sui_move_item({0}, {1}, {2})"},
	Builtin{name: "map", js: "__sui_map({0}, {1})"},
	Builtin{name: "filter", js: "__sui_filter({0}, {1})"},
	Builtin{name: "reduce", js: "__sui_reduce({0}, {1}, {2})"},
	Builtin{name: "find", js: "__sui_find({0}, {1})"},
	Builtin{name: "findIndex", js: "__sui_find_index({0}, {1})"},
	Builtin{name: "any", js: "__sui_any({0}, {1})"},
	Builtin{name: "all", js: "__sui_all({0}, {1})"},
	Builtin{name: "sortBy", js: "__sui_sort_by({0}, {1})"},
	Builtin{name: "unique", js: "__sui_unique({0})"},
	// ---- drag and drop ----
	// DnD handlers are zero-argument functions: these builtins read the
	// payload captured by the browser event runtime.
	Builtin{name: "drag.source", js: "__sui_drag_source()"},
	Builtin{name: "drag.target", js: "__sui_drag_target()"},
	Builtin{name: "drag.data", js: "__sui_drag_data()"},
	Builtin{name: "drag.files", js: "__sui_drag_files()"},
	Builtin{name: "drag.folder", js: "__sui_drag_folder()"},
	Builtin{name: "drag.error", js: "__sui_drag_error()"},
	Builtin{name: "drag.active", js: "__sui_drag_active()"},
	Builtin{name: "drag.clear", js: "__sui_drag_clear()"},
	// ---- pages ----
	// navigate switches to another non-invoked root view (a page). It is a
	// no-op in single-page apps; unknown page names are ignored at runtime.
	Builtin{name: "navigate", js: "__sui_navigate({0})"},
	// ---- modal ----
	Builtin{name: "modal.open", js: "__sui_modal_open({0})"},
	Builtin{name: "modal.close", js: "__sui_modal_close({0})"},
	Builtin{name: "toast.show", js: "__sui_toast_show({0}, {1}, {2})"},
	Builtin{name: "toast.clear", js: "__sui_toast_clear()"},
	// ---- timers ----
	Builtin{name: "after", js: "setTimeout(() => {1}(), {0})"},
	Builtin{name: "interval", js: "setInterval(() => {1}(), {0})"},
	Builtin{name: "timers.cancel", js: "clearTimeout({0})"},
	Builtin{name: "timers.debounce", js: "__sui_timers_debounce({0}, {1})"},
	Builtin{name: "timers.throttle", js: "__sui_timers_throttle({0}, {1})"},
	// ---- math ----
	Builtin{name: "min", js: "Math.min({0}, {1})"},
	Builtin{name: "max", js: "Math.max({0}, {1})"},
	Builtin{name: "abs", js: "Math.abs({0})"},
	Builtin{name: "round", js: "Math.round({0})"},
	Builtin{name: "floor", js: "Math.floor({0})"},
	Builtin{name: "ceil", js: "Math.ceil({0})"},
	Builtin{name: "pow", js: "Math.pow({0}, {1})"},
	Builtin{name: "sqrt", js: "Math.sqrt({0})"},
	Builtin{name: "random", js: "Math.random()"},
	Builtin{name: "randint", js: "(Math.floor(Math.random() * ({1} - {0} + 1)) + {0})"},
	// ---- time/date ----
	Builtin{name: "time.now", js: "__sui_time_now()"},
	Builtin{name: "time.nowMs", js: "__sui_time_nowMs()"},
	Builtin{name: "time.format", js: "__sui_time_format({0}, {1})"},
	Builtin{name: "time.parse", js: "__sui_time_parse({0}, {1})"},
	Builtin{name: "time.add", js: "__sui_time_add({0}, {1}, {2})"},
	Builtin{name: "time.diff", js: "__sui_time_diff({0}, {1}, {2})"},
	Builtin{name: "time.toISO", js: "__sui_time_toISO({0})"},
	Builtin{name: "time.toMs", js: "__sui_time_toMs({0})"},
	// ---- internationalization ----
	Builtin{name: "t", js: "__sui_t({0}, {1})"},
	Builtin{name: "i18n.t", js: "__sui_t({0}, {1})"},
	Builtin{name: "i18n.locale", js: "__sui_i18n_locale()"},
	Builtin{name: "i18n.setLocale", js: "__sui_i18n_set({0}, true)"},
	Builtin{name: "i18n.detect", js: "__sui_i18n_detect()"},
	Builtin{name: "i18n.available", js: "__sui_i18n_available()"},
	Builtin{name: "i18n.dir", js: "__sui_i18n_dir()"},
	Builtin{name: "i18n.has", js: "__sui_i18n_has({0})"},
	Builtin{name: "i18n.missing", js: "__sui_i18n_missing()"},
	// ---- url/encoding ----
	Builtin{name: "url.encode", js: "__sui_url_encode({0})"},
	Builtin{name: "url.decode", js: "__sui_url_decode({0})"},
	Builtin{name: "url.encodeComponent", js: "encodeURIComponent({0})"},
	Builtin{name: "url.decodeComponent", js: "__sui_url_decodeComponent({0})"},
	Builtin{name: "encoding.base64Encode", js: "__sui_b64_encode({0})"},
	Builtin{name: "encoding.base64Decode", js: "__sui_b64_decode({0})"},
	// ---- storage (cross-target) ----
	Builtin{name: "storage.get", js: "__sui_storage_get({0}, {1})"},
	Builtin{name: "storage.set", js: "__sui_storage_set({0}, {1})"},
	Builtin{name: "storage.remove", js: "__sui_storage_remove({0})"},
	Builtin{name: "storage.clear", js: "__sui_storage_clear()"},
	Builtin{name: "storage.keys", js: "__sui_storage_keys()"},
	// ---- clipboard ----
	Builtin{name: "clipboard.readText", js: "__sui_clip_read()"},
	Builtin{name: "clipboard.writeText", js: "__sui_clip_write({0})"},
	// ---- validation ----
	Builtin{name: "validation.isEmail", js: "__sui_valid_email({0})"},
	Builtin{name: "validation.isUrl", js: "__sui_valid_url({0})"},
	Builtin{name: "validation.isEmpty", js: "__sui_valid_empty({0})"},
	Builtin{name: "validation.required", js: "__sui_valid_required({0})"},
	Builtin{name: "validation.minLen", js: "__sui_valid_minlen({0}, {1})"},
	Builtin{name: "validation.maxLen", js: "__sui_valid_maxlen({0}, {1})"},
	Builtin{name: "validation.length", js: "__sui_valid_length({0}, {1}, {2})"},
	Builtin{name: "validation.pattern", js: "__sui_valid_pattern({0}, {1})"},
	Builtin{name: "validation.isNumber", js: "__sui_valid_isnum({0})"},
	Builtin{name: "validation.isBool", js: "__sui_valid_isbool({0})"},
	Builtin{name: "validation.isString", js: "__sui_valid_isstr({0})"},
	Builtin{name: "validation.isArray", js: "__sui_valid_isarr({0})"},
	Builtin{name: "validation.isObject", js: "__sui_valid_isobj({0})"},
}

// IsBuiltin reports whether `name` is a built-in function.
func IsBuiltin(name string) bool {
	for i := 0; i < len(Builtins); i++ {
		if Builtins[i].name == name {
			return true
		}
	}
	return false
}

// builtinToJs renders the JS translation for a built-in call, substituting
// argument expressions for {0}..{n} placeholders.
func builtinToJs(name string, args []string) string {
	js := ""
	for i := 0; i < len(Builtins); i++ {
		if Builtins[i].name == name {
			js = Builtins[i].js
			break
		}
	}
	if js == "" {
		return ""
	}
	for i := 0; i < len(args); i++ {
		js = strings.ReplaceAll(js, "{"+strconv.Itoa(i)+"}", args[i])
	}
	// Optional arguments (t(key), time.format(dt), ...) leave higher
	// placeholders in the template. Pass undefined so the generated JS stays
	// valid and the runtime helper applies its default.
	for i := len(args); i < 10; i++ {
		js = strings.ReplaceAll(js, "{"+strconv.Itoa(i)+"}", "undefined")
	}
	return js
}
