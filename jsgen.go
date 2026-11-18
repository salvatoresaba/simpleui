package main

import "fmt"
import "regexp"
import "sort"
import "strconv"
import "strings"
import "sync"

var SuiPrelude = "\"use strict\";\n" +
	"\n" +
	"const __SUI_views  = Object.create(null);\n" +
	"const __SUI_bind   = Object.create(null);\n" +
	"const __SUI_inst   = Object.create(null);\n" +
	"\n" +
	"function __sui_bindView(stateVar, viewName) {\n" +
	"  (__SUI_bind[stateVar] ||= []).push(viewName);\n" +
	"}\n" +
	"function __sui_preset(value, prefix, allowed) {\n" +
	"  var v = String(value == null ? '' : value);\n" +
	"  if (v === 'default' || v === 'accent' || !v) return '';\n" +
	"  for (var i = 0; i < allowed.length; i++) if (allowed[i] === v) return ' ' + prefix + v;\n" +
	"  return '';\n" +
	"}\n" +
	"const __SUI_COMPUTED = [];\n" +
	"var __sui_pendingRender = false;\n" +
	"var __sui_pendingViews = null;\n" +
	"function __sui_recompute() {\n" +
	"  for (var i = 0; i < __SUI_COMPUTED.length; i++) __SUI_COMPUTED[i]();\n" +
	"}\n" +
	"function __sui_schedule(views) {\n" +
	"  if (!views || !views.length) return;\n" +
	"  if (!__sui_pendingViews) __sui_pendingViews = Object.create(null);\n" +
	"  for (var i = 0; i < views.length; i++) __sui_pendingViews[views[i]] = 1;\n" +
	"  if (__sui_pendingRender) return;\n" +
	"  __sui_pendingRender = true;\n" +
	"  var flush = function () {\n" +
	"    __sui_pendingRender = false;\n" +
	"    var names = Object.keys(__sui_pendingViews || {});\n" +
	"    __sui_pendingViews = null;\n" +
	"    __sui_recompute();\n" +
	"    for (var j = 0; j < names.length; j++) if (__SUI_views[names[j]]) __SUI_views[names[j]]();\n" +
	"  };\n" +
	"  if (typeof queueMicrotask === 'function') queueMicrotask(flush);\n" +
	"  else setTimeout(flush, 0);\n" +
	"}\n" +
	"function __sui_react(stateVar) {\n" +
	"  // Computed values are refreshed immediately so a handler that reads one\n" +
	"  // after a state change sees the current value. Only views bound to this\n" +
	"  // state are scheduled, and updates in the same tick are batched.\n" +
	"  __sui_recompute();\n" +
	"  __sui_schedule(__SUI_bind[stateVar] || []);\n" +
	"}\n" +
	"function __sui_render() {\n" +
	"  __sui_recompute();\n" +
	"  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();\n" +
	"}\n" +
	"function __sui_refresh() {\n" +
	"  __sui_recompute();\n" +
	"  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();\n" +
	"}\n" +
	"function __sui_intl_locale(locale) {\n" +
	"  var l = locale ? String(locale) : '';\n" +
	"  if (!l && typeof __sui_i18n_locale === 'function') l = __sui_i18n_locale();\n" +
	"  if (!l) { try { l = (navigator.languages && navigator.languages[0]) || navigator.language || 'en'; } catch (e) { l = 'en'; } }\n" +
	"  return l || 'en';\n" +
	"}\n" +
	"function __sui_intl_date_value(value) {\n" +
	"  if (value instanceof Date) return value;\n" +
	"  var d = typeof value === 'number' ? new Date(value) : new Date(String(value));\n" +
	"  return isNaN(d.getTime()) ? null : d;\n" +
	"}\n" +
	"function __sui_intl_style(style, kind) {\n" +
	"  var s = String(style || 'medium').toLowerCase();\n" +
	"  if (s === 'iso') return 'iso';\n" +
	"  if (s !== 'short' && s !== 'long' && s !== 'full') s = 'medium';\n" +
	"  var opts = {};\n" +
	"  if (kind === 'time') opts.timeStyle = s;\n" +
	"  else if (kind === 'datetime') { opts.dateStyle = s; opts.timeStyle = s; }\n" +
	"  else opts.dateStyle = s;\n" +
	"  return opts;\n" +
	"}\n" +
	"function __sui_intl_format_date(value, style, locale, timeZone, kind) {\n" +
	"  var d = __sui_intl_date_value(value);\n" +
	"  if (!d) return 'Invalid date';\n" +
	"  var opts = __sui_intl_style(style, kind);\n" +
	"  if (opts === 'iso') {\n" +
	"    if (kind === 'date') return d.toISOString().slice(0, 10);\n" +
	"    if (kind === 'time') return d.toISOString().slice(11, 19);\n" +
	"    return d.toISOString();\n" +
	"  }\n" +
	"  if (timeZone) opts.timeZone = String(timeZone);\n" +
	"  try { return new Intl.DateTimeFormat(__sui_intl_locale(locale), opts).format(d); } catch (e) { return 'Invalid date'; }\n" +
	"}\n" +
	"function __sui_intl_date(value, style, locale, timeZone) { return __sui_intl_format_date(value, style, locale, timeZone, 'date'); }\n" +
	"function __sui_intl_time(value, style, locale, timeZone) { return __sui_intl_format_date(value, style, locale, timeZone, 'time'); }\n" +
	"function __sui_intl_datetime(value, style, locale, timeZone) { return __sui_intl_format_date(value, style, locale, timeZone, 'datetime'); }\n" +
	"function __sui_intl_digits(decimals) {\n" +
	"  if (decimals === undefined || decimals === null || decimals === '') return null;\n" +
	"  var d = Number(decimals);\n" +
	"  return isFinite(d) && d >= 0 ? Math.floor(d) : null;\n" +
	"}\n" +
	"function __sui_intl_number(value, decimals, locale) {\n" +
	"  var n = Number(value);\n" +
	"  if (!isFinite(n)) return '';\n" +
	"  var opts = {}, d = __sui_intl_digits(decimals);\n" +
	"  if (d !== null) { opts.minimumFractionDigits = d; opts.maximumFractionDigits = d; }\n" +
	"  try { return new Intl.NumberFormat(__sui_intl_locale(locale), opts).format(n); } catch (e) { return String(n); }\n" +
	"}\n" +
	"function __sui_intl_currency(value, currency, locale) {\n" +
	"  var n = Number(value), code = String(currency || 'USD').toUpperCase();\n" +
	"  if (!isFinite(n)) return '';\n" +
	"  try { return new Intl.NumberFormat(__sui_intl_locale(locale), { style: 'currency', currency: code }).format(n); } catch (e) { return code + ' ' + n; }\n" +
	"}\n" +
	"function __sui_intl_percent(value, decimals, locale) {\n" +
	"  var n = Number(value);\n" +
	"  if (!isFinite(n)) return '';\n" +
	"  var opts = { style: 'percent' }, d = __sui_intl_digits(decimals);\n" +
	"  if (d !== null) { opts.minimumFractionDigits = d; opts.maximumFractionDigits = d; }\n" +
	"  try { return new Intl.NumberFormat(__sui_intl_locale(locale), opts).format(n); } catch (e) { return String(n); }\n" +
	"}\n" +
	"function __sui_intl_relative(value, unit, locale, numeric) {\n" +
	"  var u = String(unit || ''), n = Number(value);\n" +
	"  var allowed = { second: 1, seconds: 1, minute: 1, minutes: 1, hour: 1, hours: 1, day: 1, days: 1, week: 1, weeks: 1, month: 1, months: 1, quarter: 1, quarters: 1, year: 1, years: 1 };\n" +
	"  if (!allowed[u] || !isFinite(n)) return '';\n" +
	"  var mode = String(numeric || 'auto');\n" +
	"  if (mode !== 'always' && mode !== 'auto') mode = 'auto';\n" +
	"  try { return new Intl.RelativeTimeFormat(__sui_intl_locale(locale), { numeric: mode }).format(Math.trunc(n), u); } catch (e) { return ''; }\n" +
	"}\n" +
	"function __sui_intl_timezone() {\n" +
	"  try { return Intl.DateTimeFormat().resolvedOptions().timeZone || ''; } catch (e) { return ''; }\n" +
	"}\n" +
	"function __sui_keyed(html, key) {\n" +
	"  var k = __sui_esc(key);\n" +
	"  return String(html).replace(/^\\s*<([a-zA-Z][\\w-]*)/, '<$1 data-sui-key=\"' + k + '\"');\n" +
	"}\n" +
	"function __sui_assert(cond, msg) {\n" +
	"  if (!cond) throw new Error(msg == null || String(msg) === '' ? 'assertion failed' : String(msg));\n" +
	"}\n" +
	"function __sui_boundary(fn, fallback) {\n" +
	"  try { return fn(); } catch (e) {\n" +
	"    try { console.error('SUI render error:', e); } catch (ignored) {}\n" +
	"    var msg = fallback == null || String(fallback) === '' ? 'Something went wrong.' : String(fallback);\n" +
	"    var detail = e && e.message ? String(e.message) : String(e);\n" +
	"    return '<div class=\"sui-error\" role=\"alert\" data-sui-error=\"' + __sui_esc(detail) + '\">' + __sui_esc(msg) + '</div>';\n" +
	"  }\n" +
	"}\n" +
	"// Collection helpers take a declared function (or its name for an fn\n" +
	"// parameter) and call it once per item. Extra arguments are ignored by\n" +
	"// callbacks that declare fewer parameters.\n" +
	"function __sui_callback(fn, name) {\n" +
	"  if (typeof fn === 'function') return fn;\n" +
	"  if (typeof fn === 'string' && typeof window[fn] === 'function') return window[fn];\n" +
	"  throw new Error(name + ' expects a function name');\n" +
	"}\n" +
	"function __sui_map(arr, fn) {\n" +
	"  var f = __sui_callback(fn, 'map'), list = arr || [], out = [];\n" +
	"  for (var i = 0; i < list.length; i++) out.push(f(list[i], i));\n" +
	"  return out;\n" +
	"}\n" +
	"function __sui_filter(arr, fn) {\n" +
	"  var f = __sui_callback(fn, 'filter'), list = arr || [], out = [];\n" +
	"  for (var i = 0; i < list.length; i++) if (f(list[i], i)) out.push(list[i]);\n" +
	"  return out;\n" +
	"}\n" +
	"function __sui_reduce(arr, init, fn) {\n" +
	"  var f = __sui_callback(fn, 'reduce'), list = arr || [], acc = init;\n" +
	"  for (var i = 0; i < list.length; i++) acc = f(acc, list[i], i);\n" +
	"  return acc;\n" +
	"}\n" +
	"function __sui_find(arr, fn) {\n" +
	"  var f = __sui_callback(fn, 'find'), list = arr || [];\n" +
	"  for (var i = 0; i < list.length; i++) if (f(list[i], i)) return list[i];\n" +
	"  return null;\n" +
	"}\n" +
	"function __sui_find_index(arr, fn) {\n" +
	"  var f = __sui_callback(fn, 'findIndex'), list = arr || [];\n" +
	"  for (var i = 0; i < list.length; i++) if (f(list[i], i)) return i;\n" +
	"  return -1;\n" +
	"}\n" +
	"function __sui_any(arr, fn) {\n" +
	"  var f = __sui_callback(fn, 'any'), list = arr || [];\n" +
	"  for (var i = 0; i < list.length; i++) if (f(list[i], i)) return true;\n" +
	"  return false;\n" +
	"}\n" +
	"function __sui_all(arr, fn) {\n" +
	"  var f = __sui_callback(fn, 'all'), list = arr || [];\n" +
	"  for (var i = 0; i < list.length; i++) if (!f(list[i], i)) return false;\n" +
	"  return true;\n" +
	"}\n" +
	"function __sui_sort_by(arr, fn) {\n" +
	"  var f = __sui_callback(fn, 'sortBy'), list = (arr || []).slice();\n" +
	"  list.sort(function (a, b) { var ka = f(a), kb = f(b); if (ka < kb) return -1; if (ka > kb) return 1; return 0; });\n" +
	"  return list;\n" +
	"}\n" +
	"function __sui_unique(arr) {\n" +
	"  var list = arr || [], out = [], seen = new Set();\n" +
	"  for (var i = 0; i < list.length; i++) { if (!seen.has(list[i])) { seen.add(list[i]); out.push(list[i]); } }\n" +
	"  return out;\n" +
	"}\n" + "// Internationalization: catalogs are emitted by i18n declarations. The\n" +
	"// active locale comes from ?locale=, then localStorage, then the browser.\n" +
	"var __SUI_I18N = { catalogs: {}, locale: '', fallback: '', dir: 'ltr' };\n" +
	"function __sui_i18n_rtl(locale) { return /^(ar|he|fa|ur|yi|dv|ps|sd)(-|$)/i.test(String(locale || '')); }\n" +
	"function __sui_i18n_match(locale) {\n" +
	"  var want = String(locale || '').toLowerCase();\n" +
	"  if (!want) return '';\n" +
	"  var keys = Object.keys(__SUI_I18N.catalogs || {});\n" +
	"  for (var i = 0; i < keys.length; i++) if (keys[i].toLowerCase() === want) return keys[i];\n" +
	"  var base = want.split('-')[0];\n" +
	"  for (var j = 0; j < keys.length; j++) if (keys[j].toLowerCase().split('-')[0] === base) return keys[j];\n" +
	"  return '';\n" +
	"}\n" +
	"function __sui_i18n_apply(locale) {\n" +
	"  __SUI_I18N.locale = locale || __SUI_I18N.fallback;\n" +
	"  __SUI_I18N.dir = __sui_i18n_rtl(__SUI_I18N.locale) ? 'rtl' : 'ltr';\n" +
	"  if (typeof document !== 'undefined' && document.documentElement) {\n" +
	"    document.documentElement.lang = __SUI_I18N.locale;\n" +
	"    document.documentElement.dir = __SUI_I18N.dir;\n" +
	"  }\n" +
	"}\n" +
	"function __sui_i18n_init(catalogs, fallback) {\n" +
	"  __SUI_I18N.catalogs = catalogs || {};\n" +
	"  __SUI_I18N.fallback = fallback || '';\n" +
	"  var chosen = '';\n" +
	"  try {\n" +
	"    var q = new URLSearchParams(location.search).get('locale');\n" +
	"    if (q) chosen = __sui_i18n_match(q);\n" +
	"    if (!chosen) { var saved = localStorage.getItem('sui.locale'); if (saved) chosen = __sui_i18n_match(saved); }\n" +
	"  } catch (e) {}\n" +
	"  if (!chosen) chosen = __sui_i18n_detect();\n" +
	"  if (!chosen) chosen = __SUI_I18N.fallback;\n" +
	"  __sui_i18n_apply(chosen);\n" +
	"}\n" +
	"function __sui_i18n_set(locale, persist) {\n" +
	"  var want = String(locale || '');\n" +
	"  var chosen = want.toLowerCase() === 'auto' ? __sui_i18n_detect() : __sui_i18n_match(want);\n" +
	"  if (!chosen) return false;\n" +
	"  var changed = chosen !== __SUI_I18N.locale;\n" +
	"  __sui_i18n_apply(chosen);\n" +
	"  if (persist) { try { localStorage.setItem('sui.locale', chosen); } catch (e) {} }\n" +
	"  if (changed && typeof __sui_render === 'function') __sui_render();\n" +
	"  return true;\n" +
	"}\n" +
	"function __sui_i18n_locale() { return __SUI_I18N.locale || __SUI_I18N.fallback; }\n" +
	"function __sui_i18n_detect() {\n" +
	"  var langs = [];\n" +
	"  try { langs = navigator.languages || [navigator.language || '']; } catch (e) {}\n" +
	"  for (var i = 0; i < langs.length; i++) { var m = __sui_i18n_match(langs[i]); if (m) return m; }\n" +
	"  return '';\n" +
	"}\n" +
	"function __sui_i18n_available() { return Object.keys(__SUI_I18N.catalogs || {}); }\n" +
	"function __sui_i18n_dir() { return __SUI_I18N.dir || (__sui_i18n_rtl(__sui_i18n_locale()) ? 'rtl' : 'ltr'); }\n" +
	"function __sui_i18n_entry(key) {\n" +
	"  var k = String(key == null ? '' : key);\n" +
	"  var cur = __SUI_I18N.catalogs[__SUI_I18N.locale];\n" +
	"  if (cur && Object.prototype.hasOwnProperty.call(cur, k)) return cur[k];\n" +
	"  var fb = __SUI_I18N.catalogs[__SUI_I18N.fallback];\n" +
	"  if (fb && Object.prototype.hasOwnProperty.call(fb, k)) return fb[k];\n" +
	"  return null;\n" +
	"}\n" +
	"function __sui_i18n_has(key) { return __sui_i18n_entry(key) != null; }\n" +
	"function __sui_i18n_missing() {\n" +
	"  var fb = __SUI_I18N.catalogs[__SUI_I18N.fallback] || {};\n" +
	"  var cur = __SUI_I18N.catalogs[__SUI_I18N.locale] || {};\n" +
	"  var out = [];\n" +
	"  for (var k in fb) if (Object.prototype.hasOwnProperty.call(fb, k) && !Object.prototype.hasOwnProperty.call(cur, k)) out.push(k);\n" +
	"  return out;\n" +
	"}\n" +
	"function __sui_i18n_vars(text, vars) {\n" +
	"  if (!vars) return text;\n" +
	"  return String(text).replace(/\\{([A-Za-z0-9_.-]+)\\}/g, function (_, name) {\n" +
	"    var v = vars[name];\n" +
	"    return v === null || v === undefined ? '' : String(v);\n" +
	"  });\n" +
	"}\n" +
	"function __sui_t(key, vars) {\n" +
	"  var k = String(key == null ? '' : key);\n" +
	"  var entry = __sui_i18n_entry(k);\n" +
	"  if (entry === null || entry === undefined) return k;\n" +
	"  if (typeof entry === 'object') {\n" +
	"    var cat = 'other';\n" +
	"    if (vars && vars.count !== null && vars.count !== undefined) {\n" +
	"      try {\n" +
	"        if (typeof Intl !== 'undefined' && Intl.PluralRules) cat = new Intl.PluralRules(__sui_i18n_locale()).select(Number(vars.count));\n" +
	"        else cat = Number(vars.count) === 1 ? 'one' : 'other';\n" +
	"      } catch (e) { cat = Number(vars.count) === 1 ? 'one' : 'other'; }\n" +
	"    }\n" +
	"    entry = entry[cat] !== undefined ? entry[cat] : (entry.other !== undefined ? entry.other : entry.one);\n" +
	"    if (entry === undefined || entry === null) return k;\n" +
	"  }\n" +
	"  return __sui_i18n_vars(String(entry), vars);\n" +
	"}\n" +
	`function __sui_replace_view(el, html) {
  if (el.__suiLastHtml === html) return;
  el.__suiLastHtml = html;
  var focusInfo = null, focusEl = document.activeElement;
  if (focusEl && focusEl !== document.body && el.contains(focusEl)) {
    var focusHolder = focusEl.closest ? (focusEl.closest('[data-sui-id]') || focusEl) : focusEl;
    var focusKeyHolder = focusEl.closest ? focusEl.closest('[data-sui-key]') : null;
    focusInfo = {
      id: focusHolder.getAttribute('data-sui-id') || '',
      key: focusKeyHolder ? (focusKeyHolder.getAttribute('data-sui-key') || '') : '',
      tag: focusEl.tagName,
      start: null,
      end: null,
      scrollX: window.scrollX,
      scrollY: window.scrollY
    };
    try { focusInfo.start = focusEl.selectionStart; focusInfo.end = focusEl.selectionEnd; } catch (e) {}
  }
  var scrollX = window.scrollX, scrollY = window.scrollY;
  var keepMedia = [];
  var mediaNodes = el.querySelectorAll('video,audio,img,iframe');
  for (var mi = 0; mi < mediaNodes.length; mi++) {
    var mn = mediaNodes[mi];
    var mHolder = mn.closest ? (mn.closest('[data-sui-id]') || mn) : mn;
    var mId = mHolder.getAttribute('data-sui-id');
    var mSrc = mn.getAttribute('src') || '';
    if (!mId || !mSrc) continue;
    keepMedia.push({ id: mId, tag: mn.tagName, src: mSrc, node: mn });
  }
  var keepAnims = [];
  if (document.getAnimations && typeof CSSAnimation !== 'undefined') {
    var liveAnims = document.getAnimations();
    for (var ai = 0; ai < liveAnims.length; ai++) {
      var an = liveAnims[ai];
      var at = an.effect && an.effect.target;
      if (!at || !el.contains(at) || !(an instanceof CSSAnimation)) continue;
      var aHolder = at.closest ? (at.closest('[data-sui-id]') || at) : at;
      var aId = aHolder.getAttribute('data-sui-id');
      if (!aId) continue;
      keepAnims.push({ id: aId, tag: at.tagName, name: an.animationName, time: an.currentTime, state: an.playState });
    }
  }
  var oldCharts = Array.prototype.slice.call(el.querySelectorAll('[data-sui-chart]'));
  var saved = [];
  for (var i = 0; i < oldCharts.length; i++) {
    saved.push({ node: oldCharts[i], key: oldCharts[i].getAttribute('data-sui-chart-name') || oldCharts[i].getAttribute('data-sui-chart-title') || '' });
  }
  var active = document.activeElement, activeInChart = null;
  if (active) for (var a = 0; a < saved.length; a++) if (saved[a].node.contains(active)) { activeInChart = active; break; }
  if (typeof __SUI_MOTION_GRAPH !== 'undefined' && __SUI_MOTION_GRAPH.beforeReplace) __SUI_MOTION_GRAPH.beforeReplace(el);
  if (typeof __SUI_MOTION !== 'undefined') __SUI_MOTION.before(el);
  el.innerHTML = html;
  if (typeof __SUI_MOTION !== 'undefined') __SUI_MOTION.after(el);
  if (typeof __SUI_MOTION_GRAPH !== 'undefined' && __SUI_MOTION_GRAPH.afterReplace) __SUI_MOTION_GRAPH.afterReplace(el);
  for (var km = 0; km < keepMedia.length; km++) {
    var keep = keepMedia[km];
    var kHolder = el.querySelector('[data-sui-id="' + keep.id + '"]');
    if (!kHolder) continue;
    var kFresh = kHolder.tagName === keep.tag ? kHolder : kHolder.querySelector(keep.tag.toLowerCase());
    if (!kFresh || kFresh === keep.node || (kFresh.getAttribute('src') || '') !== keep.src) continue;
    if (!kFresh.parentNode) continue;
    for (var ka = 0; ka < kFresh.attributes.length; ka++) {
      var kattr = kFresh.attributes[ka];
      if (kattr.name === 'src') continue;
      keep.node.setAttribute(kattr.name, kattr.value);
    }
    kFresh.parentNode.replaceChild(keep.node, kFresh);
  }
  var valueSelects = el.querySelectorAll('select[value]');
  for (var vs = 0; vs < valueSelects.length; vs++) {
    var wantValue = valueSelects[vs].getAttribute('value');
    if (wantValue && valueSelects[vs].value !== wantValue) {
      try { valueSelects[vs].value = wantValue; } catch (e) {}
    }
  }
  for (var kn = 0; kn < keepAnims.length; kn++) {
    var ka2 = keepAnims[kn];
    var anHolder = el.querySelector('[data-sui-id="' + ka2.id + '"]');
    if (!anHolder || !anHolder.getAnimations) continue;
    var anTargets = [];
    if (anHolder.tagName === ka2.tag) anTargets.push(anHolder);
    else anTargets = Array.prototype.slice.call(anHolder.querySelectorAll(ka2.tag.toLowerCase()));
    for (var ati = 0; ati < anTargets.length; ati++) {
      var targetAnims = anTargets[ati].getAnimations();
      for (var tai = 0; tai < targetAnims.length; tai++) {
        if (targetAnims[tai].animationName !== ka2.name) continue;
        try {
          targetAnims[tai].currentTime = ka2.time;
          if (ka2.state === 'paused') targetAnims[tai].pause();
        } catch (e) {}
      }
    }
  }
  if (focusInfo && (focusInfo.id || focusInfo.key)) {
    var fHolder = null;
    if (focusInfo.key) {
      var keyed = el.querySelectorAll('[data-sui-key]');
      for (var fki = 0; fki < keyed.length; fki++) {
        if (keyed[fki].getAttribute('data-sui-key') === focusInfo.key) { fHolder = keyed[fki]; break; }
      }
    }
    if (!fHolder && focusInfo.id) fHolder = el.querySelector('[data-sui-id="' + focusInfo.id + '"]');
    var fTarget = null;
    if (fHolder) {
      if (fHolder.tagName === focusInfo.tag) fTarget = fHolder;
      else fTarget = fHolder.querySelector(focusInfo.tag.toLowerCase());
      if (!fTarget && fHolder.matches && fHolder.matches('input,textarea,select')) fTarget = fHolder;
    }
    if (fTarget && fTarget.focus) {
      try { fTarget.focus({ preventScroll: true }); } catch (e) { try { fTarget.focus(); } catch (e2) {} }
      if (focusInfo.start != null && fTarget.setSelectionRange) {
        try { fTarget.setSelectionRange(focusInfo.start, focusInfo.end); } catch (e) {}
      }
    }
  }
  if (window.scrollX !== scrollX || window.scrollY !== scrollY) {
    window.scrollTo(scrollX, scrollY);
  }
  if (!saved.length) return;
  var fresh = Array.prototype.slice.call(el.querySelectorAll('[data-sui-chart]'));
  var used = [], changed = [];
  for (var s = 0; s < saved.length; s++) {
    var target = null;
    for (var j = 0; j < fresh.length; j++) {
      if (used.indexOf(j) >= 0) continue;
      var key = fresh[j].getAttribute('data-sui-chart-name') || fresh[j].getAttribute('data-sui-chart-title') || '';
      if (key === saved[s].key) { target = fresh[j]; used.push(j); break; }
    }
    if (!target) continue;
    var old = saved[s].node, diff = false;
    var redrawAttr = /^data-sui-chart-(data|type|format|x|y|width|height|overflow|zoom|point-width|point-size|bins|min|max|table-view|table-filters|table-add-row|table-delete-row|table-steppers|table-context-menu|table-selection|critical-path|holidays|weekend|timezone|show-non-working-days|progress-column|grid|legend|values|title)$/;
    for (var k = 0; k < target.attributes.length; k++) {
      var attr = target.attributes[k];
      if (attr.name === 'class') {
        var cls = (old.getAttribute('class') || '').split(/\s+/), add = attr.value.split(/\s+/);
        for (var c = 0; c < add.length; c++) if (add[c] && cls.indexOf(add[c]) < 0) cls.push(add[c]);
        old.setAttribute('class', cls.join(' '));
        continue;
      }
      if (attr.name === 'style') {
        var styleProbe = document.createElement('div');
        styleProbe.setAttribute('style', attr.value);
        for (var sp = 0; sp < styleProbe.style.length; sp++) {
          var styleProp = styleProbe.style[sp];
          old.style.setProperty(styleProp, styleProbe.style.getPropertyValue(styleProp));
        }
        continue;
      }
      if (old.getAttribute(attr.name) !== attr.value) {
        old.setAttribute(attr.name, attr.value);
        if (redrawAttr.test(attr.name)) diff = true;
      }
    }
    var oldTitle = old.querySelector('.sui-chart-title'), newTitle = target.querySelector('.sui-chart-title');
    if (oldTitle && newTitle && oldTitle.textContent !== newTitle.textContent) { oldTitle.textContent = newTitle.textContent; diff = true; }
    if (!target.parentNode) continue;
    target.parentNode.replaceChild(old, target);
    if (diff) changed.push(old);
  }
  if (activeInChart && activeInChart.focus) { try { activeInChart.focus({ preventScroll: true }); } catch (e) {} }
  if (changed.length && window.__SUI_CHART) {
    clearTimeout(window.__SUI_CHART_REDRAW_TIMER);
    window.__SUI_CHART_REDRAW_TIMER = setTimeout(function() {
      var redrawOne = window.__SUI_CHART.redrawRoot;
      for (var r = 0; r < changed.length; r++) {
        if (!changed[r].isConnected) continue;
        if (typeof redrawOne === 'function') redrawOne(changed[r]);
        else if (typeof window.__SUI_CHART.redraw === 'function') window.__SUI_CHART.redraw();
      }
    }, 0);
  }
}
` +
	"// Desktop native picker bridge. The dialog module is desktop-only; the web\n" +
	"// branch of stdBodyJs never calls this helper.\n" +
	"function __sui_dialog_pick(endpoint, multiple) {\n" +
	"  const fd = new URLSearchParams();\n" +
	"  fd.set('token', __SUI_TOKEN);\n" +
	"  if (multiple) fd.set('multiple', '1');\n" +
	"  return fetch(endpoint, { method: 'POST', body: fd }).then(async (r) => {\n" +
	"    const j = await r.json();\n" +
	"    if (!j.ok) return [null, j.message || 'dialog failed'];\n" +
	"    return [j.data || [], null];\n" +
	"  }).catch((e) => [null, String(e)]);\n" +
	"}\n" +
	"// __sui_navigate switches between non-invoked root views (pages). It is a\n" +
	"// no-op in single-page apps and for unknown page names.\n" +
	"function __sui_navigate(name) {\n" +
	"  var p = String(name === null || name === undefined ? '' : name);\n" +
	"  if (!p || typeof __sui_page !== 'function') return;\n" +
	"  if (!__SUI_views[p]) return;\n" +
	"  __sui_page(p);\n" +
	"}\n" +
	"function __sui_vm(key, name) {\n" +
	"  const o = __SUI_inst[key];\n" +
	"  const f = (o || {})[name];\n" +
	"  if (f) f();\n" +
	"}\n" +
	"// __sui_esc escapes interpolated values for HTML text and attribute context.\n" +
	"// Views always go through it, so untrusted data cannot inject markup.\n" +
	"function __sui_esc(v) {\n" +
	"  return String(v === null || v === undefined ? '' : v).replace(/[&<>\"']/g, function (c) {\n" +
	"    return c === '&' ? '&amp;' : c === '<' ? '&lt;' : c === '>' ? '&gt;' : c === '\"' ? '&quot;' : '&#39;';\n" +
	"  });\n" +
	"}\n" +
	"// __sui_class keeps only class tokens declared by the app (its CSS and\n" +
	"// static class attributes), so data cannot activate arbitrary classes.\n" +
	"function __sui_class(v) {\n" +
	"  var toks = String(v === null || v === undefined ? '' : v).split(/\\s+/);\n" +
	"  var out = [];\n" +
	"  for (var i = 0; i < toks.length; i++) {\n" +
	"    if (toks[i] && __SUI_CLASSES[toks[i]]) out.push(toks[i]);\n" +
	"  }\n" +
	"  return out.join(' ');\n" +
	"}\n" +
	"// __sui_url allows only safe URL schemes and relative URLs. It blocks\n" +
	"// javascript:, vbscript:, data: and other active schemes; use an allow-list\n" +
	"// in the app when the destination itself is untrusted.\n" +
	"function __sui_url(v) {\n" +
	"  var s = String(v === null || v === undefined ? '' : v).trim();\n" +
	"  var probe = s.replace(/[\\u0000-\\u0020\\u007f]+/g, '');\n" +
	"  if (probe === '') return '';\n" +
	"  if (/^(https?:|mailto:|tel:)/i.test(probe)) return s;\n" +
	"  if (/^[a-z][a-z0-9+.-]*:/i.test(probe)) return '#';\n" +
	"  return s;\n" +
	"}\n" +
	"// __sui_img_url is the image-source variant of __sui_url. Raster images and\n" +
	"// GIFs are embedded as data:image/... URIs by the designer asset manager;\n" +
	"// scripts never run in an <img>, so those values are safe there.\n" +
	"function __sui_img_url(v) {\n" +
	"  var s = String(v === null || v === undefined ? '' : v).trim();\n" +
	"  var probe = s.replace(/[\\u0000-\\u0020\\u007f]+/g, '');\n" +
	"  if (/^data:image\\//i.test(probe)) return s;\n" +
	"  return __sui_url(s);\n" +
	"}\n" +
	"// __sui_iframe_src validates iframe URLs. Malformed data is rejected with a\n" +
	"// reason that the renderer shows in a visible alert (never loaded).\n" +
	"function __sui_iframe_src(v, allow, required) {\n" +
	"  var s = String(v === null || v === undefined ? '' : v).trim();\n" +
	"  if (!s) return ['', 'missing URL'];\n" +
	"  if (s.length > 4096) return ['', 'URL too long'];\n" +
	"  for (var i = 0; i < s.length; i++) {\n" +
	"    var c = s.charCodeAt(i);\n" +
	"    if (c <= 32 || c === 127) return ['', 'invalid URL'];\n" +
	"  }\n" +
	"  if (s.indexOf('\\\\') >= 0) return ['', 'invalid URL'];\n" +
	"  var u;\n" +
	"  try { u = new URL(s, location.origin); } catch(e) { return ['', 'invalid URL']; }\n" +
	"  var scheme = u.protocol.toLowerCase();\n" +
	"  if (scheme !== 'http:' && scheme !== 'https:') return ['', 'scheme not allowed'];\n" +
	"  if (u.username || u.password) return ['', 'credentials are not allowed'];\n" +
	"  var items = String(allow === null || allow === undefined ? '' : allow).split(/[\\s,]+/);\n" +
	"  var any = false;\n" +
	"  for (var j = 0; j < items.length; j++) {\n" +
	"    var a = items[j].trim();\n" +
	"    if (!a) continue;\n" +
	"    any = true;\n" +
	"    if (a.charAt(0) === '/') {\n" +
	"      if (a.charAt(1) === '/' || a.indexOf('?') >= 0 || a.indexOf('#') >= 0) continue;\n" +
	"      if (u.origin !== location.origin) continue;\n" +
	"      var ap = a;\n" +
	"      if (ap.length > 1 && ap.charAt(ap.length - 1) === '/') ap = ap.slice(0, -1);\n" +
	"      if (ap === '/' || u.pathname === ap || u.pathname.indexOf(ap + '/') === 0) return [u.href, null];\n" +
	"      continue;\n" +
	"    }\n" +
	"    var p;\n" +
	"    try { p = new URL(a); } catch(e) { continue; }\n" +
	"    if (p.protocol !== u.protocol || p.host !== u.host) continue;\n" +
	"    if (p.username || p.password) continue;\n" +
	"    var pp = p.pathname || '/';\n" +
	"    if (pp !== '/') {\n" +
	"      if (u.pathname.indexOf(pp) !== 0) continue;\n" +
	"      if (pp.charAt(pp.length - 1) !== '/' && u.pathname.length > pp.length && u.pathname.charAt(pp.length) !== '/') continue;\n" +
	"    }\n" +
	"    return [u.href, null];\n" +
	"  }\n" +
	"  if (!any && !required) return [u.href, null];\n" +
	"  return ['', 'URL not allowed'];\n" +
	"}\n" +
	"function __sui_iframe_error(reason) {\n" +
	"  return '<div class=\"sui-iframe-error\" role=\"alert\" data-sui-iframe-error>Cannot accept the data: ' +\n" +
	"    __sui_esc(reason || 'invalid URL') + '</div>';\n" +
	"}\n" +
	"// __sui_url_guard is the SSRF guard used by http.getSafe/postSafe/\n" +
	"// getStatusSafe and ws.connectSafe. The allow-list is a developer-written\n" +
	"// literal: absolute origins (optionally with a path prefix), ws(s) origins\n" +
	"// for WebSockets, or same-origin path prefixes starting with '/'.\n" +
	"function __sui_url_guard(url, allow, ws) {\n" +
	"  var raw = String(url === null || url === undefined ? '' : url).trim();\n" +
	"  if (!raw) return ['', 'empty URL'];\n" +
	"  if (/[\\u0000-\\u001f\\u007f]/.test(raw)) return ['', 'invalid URL'];\n" +
	"  var u;\n" +
	"  try { u = new URL(raw, location.origin); } catch(e) { return ['', 'invalid URL']; }\n" +
	"  var scheme = u.protocol.toLowerCase();\n" +
	"  if (ws) {\n" +
	"    if (scheme !== 'ws:' && scheme !== 'wss:') return ['', 'only ws/wss URLs are allowed'];\n" +
	"  } else if (scheme !== 'http:' && scheme !== 'https:') {\n" +
	"    return ['', 'only http/https URLs are allowed'];\n" +
	"  }\n" +
	"  if (u.username || u.password) return ['', 'URL credentials are not allowed'];\n" +
	"  var items = String(allow === null || allow === undefined ? '' : allow).split(/[\\s,]+/);\n" +
	"  for (var i = 0; i < items.length; i++) {\n" +
	"    var a = items[i].trim();\n" +
	"    if (!a) continue;\n" +
	"    if (a.charAt(0) === '/' && !ws) {\n" +
	"      var prefix = location.origin + a;\n" +
	"      var target = u.origin + u.pathname + u.search + u.hash;\n" +
	"      if (target.indexOf(prefix) === 0) return [u.href, null];\n" +
	"      continue;\n" +
	"    }\n" +
	"    var p;\n" +
	"    try { p = new URL(a); } catch(e) { continue; }\n" +
	"    if (p.protocol !== u.protocol || p.host !== u.host) continue;\n" +
	"    var pp = p.pathname || '/';\n" +
	"    if (pp !== '/') {\n" +
	"      if (u.pathname.indexOf(pp) !== 0) continue;\n" +
	"      if (pp.charAt(pp.length - 1) !== '/' && u.pathname.length > pp.length && u.pathname.charAt(pp.length) !== '/') continue;\n" +
	"    }\n" +
	"    return [u.href, null];\n" +
	"  }\n" +
	"  return ['', 'URL is not in the allow-list'];\n" +
	"}\n" +
	"function __sui_clean_headers(h) {\n" +
	"  if (!h) return {};\n" +
	"  if (typeof h === 'string') { try { h = JSON.parse(h); } catch(e) { return {}; } }\n" +
	"  if (typeof h !== 'object') return {};\n" +
	"  var out = {};\n" +
	"  for (var k in h) { if (Object.prototype.hasOwnProperty.call(h, k)) out[String(k)] = String(h[k]); }\n" +
	"  return out;\n" +
	"}\n" +
	"function __sui_upload_form(fieldName, fileName, content) {\n" +
	"  var fd = new FormData();\n" +
	"  var blob = new Blob([String(content === null || content === undefined ? '' : content)], { type: 'application/octet-stream' });\n" +
	"  fd.append(fieldName || 'file', blob, fileName || 'upload.bin');\n" +
	"  return fd;\n" +
	"}\n" +
	"async function __sui_fetch_guarded(url, allow, init, timeoutMs) {\n" +
	"  var g = __sui_url_guard(url, allow, false);\n" +
	"  if (g[1]) return [null, g[1]];\n" +
	"  var opts = init || {};\n" +
	"  opts.redirect = 'error';\n" +
	"  var tLimit = (typeof timeoutMs === 'number' && timeoutMs > 0) ? timeoutMs : 30000;\n" +
	"  var ctrl = null, timer = null;\n" +
	"  if (typeof AbortController !== 'undefined') { ctrl = new AbortController(); opts.signal = ctrl.signal; timer = setTimeout(function(){ ctrl.abort(); }, tLimit); }\n" +
	"  try {\n" +
	"    var r = await fetch(g[0], opts);\n" +
	"    var len = Number(r.headers.get('content-length') || 0);\n" +
	"    if (len > 10485760) { if (r.body && r.body.cancel) { try { r.body.cancel(); } catch(e) {} } return [null, 'response too large']; }\n" +
	"    var text = await r.text();\n" +
	"    if (text.length > 10485760) return [null, 'response too large'];\n" +
	"    return [text, null];\n" +
	"  } catch(e) {\n" +
	"    if (ctrl && ctrl.signal && ctrl.signal.aborted) return [null, 'request timed out after ' + tLimit + 'ms'];\n" +
	"    return [null, String(e)];\n" +
	"  } finally {\n" +
	"    if (timer) clearTimeout(timer);\n" +
	"  }\n" +
	"}\n" +
	"async function __sui_fetch_status_guarded(url, allow) {\n" +
	"  var g = __sui_url_guard(url, allow, false);\n" +
	"  if (g[1]) return [null, g[1]];\n" +
	"  var opts = { redirect: 'error' };\n" +
	"  var ctrl = null, timer = null;\n" +
	"  if (typeof AbortController !== 'undefined') { ctrl = new AbortController(); opts.signal = ctrl.signal; timer = setTimeout(function(){ ctrl.abort(); }, 30000); }\n" +
	"  try {\n" +
	"    var r = await fetch(g[0], opts);\n" +
	"    if (r.body && r.body.cancel) { try { r.body.cancel(); } catch(e) {} }\n" +
	"    return [r.status, null];\n" +
	"  } catch(e) {\n" +
	"    return [null, String(e)];\n" +
	"  } finally {\n" +
	"    if (timer) clearTimeout(timer);\n" +
	"  }\n" +
	"}\n" +
	"// __sui_css sanitizes a dynamic style value: comma-free declarations only,\n" +
	"// resource loading (url(), paint(), @import), CSS escapes, expression() and\n" +
	"// positioning/overlay properties are dropped.\n" +
	"function __sui_css(v) {\n" +
	"  var text = String(v === null || v === undefined ? '' : v).replace(/\\/\\*[\\s\\S]*?\\*\\//g, '');\n" +
	"  var blocked = { position: 1, top: 1, right: 1, bottom: 1, left: 1, inset: 1, 'z-index': 1, content: 1, 'pointer-events': 1, all: 1, behavior: 1, quotes: 1 };\n" +
	"  var out = [];\n" +
	"  var parts = text.split(';');\n" +
	"  for (var i = 0; i < parts.length; i++) {\n" +
	"    var p = parts[i].trim();\n" +
	"    if (!p) continue;\n" +
	"    var idx = p.indexOf(':');\n" +
	"    if (idx <= 0) continue;\n" +
	"    var name = p.slice(0, idx).trim().toLowerCase();\n" +
	"    var val = p.slice(idx + 1).trim();\n" +
	"    if (!/^(?:--[a-z][a-z0-9-]*|[a-z][a-z0-9-]*)$/.test(name)) continue;\n" +
	"    if (name.indexOf('inset-') === 0) continue;\n" +
	"    if (blocked[name]) continue;\n" +
	"    if (!val) continue;\n" +
	"    if (/url\\s*[^a-z0-9_-]|expression\\s*[^a-z0-9_-]|javascript\\s*:|vbscript\\s*:|@import|behavior\\s*:|paint\\s*[^a-z0-9_-]|image-set\\s*[^a-z0-9_-]|cross-fade\\s*[^a-z0-9_-]|image\\s*[^a-z0-9_-]|element\\s*[^a-z0-9_-]|attr\\s*[^a-z0-9_-]|var\\s*[^a-z0-9_-]|\\\\/i.test(val)) continue;\n" +
	"    out.push(name + ': ' + val);\n" +
	"  }\n" +
	"  return out.join('; ');\n" +
	"}\n" +
	"function __sui_ac(input, options) {\n" +
	"  if (!input || !options || options[0] === undefined) return;\n" +
	"  const box = document.createElement('div');\n" +
	"  box.className = 'sui-ac';\n" +
	"  box.style.display = 'none';\n" +
	"  document.body.appendChild(box);\n" +
	"  let active = -1;\n" +
	"  let count = 0;\n" +
	"  function hide() { box.style.display = 'none'; }\n" +
	"  function show() {\n" +
	"    const r = input.getBoundingClientRect();\n" +
	"    box.style.top = (r.bottom + window.scrollY) + 'px';\n" +
	"    box.style.left = (r.left + window.scrollX) + 'px';\n" +
	"    box.style.minWidth = r.width + 'px';\n" +
	"    box.style.display = 'block';\n" +
	"  }\n" +
	"  function render() {\n" +
	"    const q = (input.value || '').toLowerCase();\n" +
	"    box.innerHTML = '';\n" +
	"    active = -1;\n" +
	"    count = 0;\n" +
	"    options.forEach(it => {\n" +
	"      if (it.toLowerCase().indexOf(q) < 0) return;\n" +
	"      count++;\n" +
	"      const d = document.createElement('div');\n" +
	"      d.className = 'sui-ac-item';\n" +
	"      d.textContent = it;\n" +
	"      d.addEventListener('mousedown', ev => { ev.preventDefault(); input.value = it; hide(); });\n" +
	"      box.appendChild(d);\n" +
	"    });\n" +
	"  }\n" +
	"  function highlight() {\n" +
	"    const items = box.querySelectorAll('.sui-ac-item');\n" +
	"    let i = 0;\n" +
	"    items.forEach(d => { d.classList.toggle('active', i === active); i++; });\n" +
	"  }\n" +
	"  input.addEventListener('focus', () => { render(); show(); });\n" +
	"  input.addEventListener('input', () => { render(); show(); });\n" +
	"  input.addEventListener('blur', () => setTimeout(hide, 150));\n" +
	"  input.addEventListener('keydown', ev => {\n" +
	"    if (!count) return;\n" +
	"    const items = box.querySelectorAll('.sui-ac-item');\n" +
	"    if (ev.key === 'ArrowDown') { ev.preventDefault(); active = Math.min(active + 1, count - 1); highlight(); }\n" +
	"    else if (ev.key === 'ArrowUp') { ev.preventDefault(); active = Math.max(active - 1, 0); highlight(); }\n" +
	"    else if (ev.key === 'Enter') { ev.preventDefault(); if (active >= 0) input.value = items[active].textContent; hide(); }\n" +
	"    else if (ev.key === 'Escape') { hide(); }\n" +
	"  });\n" +
	"}\n" +
	"\n" +
	"function __sui_timers_debounce(fn, ms) {\n" +
	"  let t = null;\n" +
	"  const wrapped = () => {\n" +
	"    if (t) clearTimeout(t);\n" +
	"    t = setTimeout(() => fn(), ms || 0);\n" +
	"  };\n" +
	"  wrapped._id = Date.now() + Math.floor(Math.random()*1000);\n" +
	"  return setTimeout(() => wrapped(), 0);\n" +
	"}\n" +
	"function __sui_timers_throttle(fn, ms) {\n" +
	"  let last = 0;\n" +
	"  const wrapped = () => {\n" +
	"    const now = Date.now();\n" +
	"    if (now - last >= (ms||0)) { last = now; fn(); }\n" +
	"  };\n" +
	"  return setTimeout(() => wrapped(), 0);\n" +
	"}\n" +
	"\n" +
	"function __sui_time_now() { return new Date(); }\n" +
	"function __sui_time_nowMs() { return Date.now(); }\n" +
	"function __sui_time_format(dt, fmt) {\n" +
	"  const d = dt instanceof Date ? dt : (dt ? new Date(dt) : new Date());\n" +
	"  if (isNaN(d.getTime())) return '';\n" +
	"  if (!fmt || fmt === 'iso') return d.toISOString();\n" +
	"  const p = (fmt || '').replace('YYYY', String(d.getFullYear()))\n" +
	"    .replace('MM', String(d.getMonth()+1).padStart(2,'0'))\n" +
	"    .replace('DD', String(d.getDate()).padStart(2,'0'))\n" +
	"    .replace('HH', String(d.getHours()).padStart(2,'0'))\n" +
	"    .replace('mm', String(d.getMinutes()).padStart(2,'0'))\n" +
	"    .replace('ss', String(d.getSeconds()).padStart(2,'0'));\n" +
	"  return p;\n" +
	"}\n" +
	"function __sui_time_parse(s, fmt) {\n" +
	"  try {\n" +
	"    if (!fmt || fmt === 'iso') { const d = s ? new Date(s) : null; if (!d || isNaN(d.getTime())) return [null, 'parse error']; return [d, null]; }\n" +
	"    const m = s.match(/(\\d{4})-(\\d{2})-(\\d{2}) (\\d{2}):(\\d{2}):(\\d{2})/);\n" +
	"    if (!m) return [null, 'format mismatch'];\n" +
	"    const d = new Date(+m[1], +m[2]-1, +m[3], +m[4], +m[5], +m[6]);\n" +
	"    if (isNaN(d.getTime())) return [null, 'parse error'];\n" +
	"    return [d, null];\n" +
	"  } catch(e){ return [null, String(e)]; }\n" +
	"}\n" +
	"function __sui_time_add(dt, amt, unit) {\n" +
	"  const d = dt instanceof Date ? dt : new Date(dt||Date.now());\n" +
	"  const v = Number(amt)||0; const u = (unit||'ms').toLowerCase();\n" +
	"  if (u==='ms') d.setMilliseconds(d.getMilliseconds()+v);\n" +
	"  else if (u==='s'||u==='sec'||u==='second') d.setSeconds(d.getSeconds()+v);\n" +
	"  else if (u==='m'||u==='min'||u==='minute') d.setMinutes(d.getMinutes()+v);\n" +
	"  else if (u==='h'||u==='hour') d.setHours(d.getHours()+v);\n" +
	"  else if (u==='d'||u==='day') d.setDate(d.getDate()+v);\n" +
	"  else if (u==='w'||u==='week') d.setDate(d.getDate()+v*7);\n" +
	"  else if (u==='mo'||u==='month') d.setMonth(d.getMonth()+v);\n" +
	"  else if (u==='y'||u==='year') d.setFullYear(d.getFullYear()+v);\n" +
	"  else d.setMilliseconds(d.getMilliseconds()+v);\n" +
	"  return d;\n" +
	"}\n" +
	"function __sui_time_diff(a,b,unit){\n" +
	"  const da=a instanceof Date?a:new Date(a||0); const db=b instanceof Date?b:new Date(b||0);\n" +
	"  let ms=Math.max(0,da.getTime()-db.getTime()); const u=(unit||'ms').toLowerCase();\n" +
	"  if(u==='s'||u==='sec') return Math.floor(ms/1000);\n" +
	"  if(u==='m'||u==='min') return Math.floor(ms/60000);\n" +
	"  if(u==='h'||u==='hour') return Math.floor(ms/3600000);\n" +
	"  if(u==='d'||u==='day') return Math.floor(ms/86400000);\n" +
	"  return ms;\n" +
	"}\n" +
	"function __sui_time_toISO(dt){ const d=dt instanceof Date?dt:new Date(dt||0); return isNaN(d.getTime())?'':d.toISOString(); }\n" +
	"function __sui_time_toMs(dt){ const d=dt instanceof Date?dt:new Date(dt||0); return isNaN(d.getTime())?0:d.getTime(); }\n" +
	"\n" +
	"function __sui_url_encode(s){ try{return encodeURI(String(s||''))}catch(e){return String(s||'')} }\n" +
	"function __sui_url_decode(s){ try{return [decodeURI(String(s||'')), null]}catch(e){return ['', String(e)]} }\n" +
	"function __sui_url_decodeComponent(s){ try{return [decodeURIComponent(String(s||'')), null]}catch(e){return ['', String(e)]} }\n" +
	"function __sui_b64_encode(s){ try{return btoa(unescape(encodeURIComponent(String(s||''))))}catch(e){return ''} }\n" +
	"function __sui_b64_decode(s){ try{return [decodeURIComponent(escape(atob(String(s||'')))) , null]}catch(e){return ['', String(e)]} }\n" +
	"function __sui_bytes_to_b64(bytes){ var s=''; var a=bytes||[]; for(var i=0;i<a.length;i++) s+=String.fromCharCode(a[i]&255); try{return btoa(s)}catch(e){return ''} }\n" +
	"function __sui_b64_to_bytes(s){ try{ var t=atob(String(s||'')); var a=[]; for(var i=0;i<t.length;i++) a.push(t.charCodeAt(i)); return a; }catch(e){ return []; } }\n" +
	"\n" +
	"function __sui_path_join() {\n" +
	"  var parts = [];\n" +
	"  for (var i = 0; i < arguments.length; i++) { var v = arguments[i]; if (v === undefined || v === null || v === '') continue; parts.push(String(v)); }\n" +
	"  var all = parts.join('/').replace(/\\\\/g, '/');\n" +
	"  var drive = ''; var dm = /^([A-Za-z]:)(\\/|$)/.exec(all); if (dm) { drive = dm[1]; all = all.slice(drive.length); }\n" +
	"  var abs = all.charAt(0) === '/'; var segs = all.split('/'); var out = [];\n" +
	"  for (var j = 0; j < segs.length; j++) {\n" +
	"    var seg = segs[j];\n" +
	"    if (seg === '' || seg === '.') continue;\n" +
	"    if (seg === '..') { if (out.length && out[out.length - 1] !== '..') out.pop(); else if (!abs) out.push('..'); }\n" +
	"    else out.push(seg);\n" +
	"  }\n" +
	"  var res = out.join('/'); if (abs) res = '/' + res; if (drive) res = drive + (res && res.charAt(0) !== '/' ? '/' : '') + res;\n" +
	"  if (!res) res = abs ? '/' : (drive ? drive + '/' : '.');\n" +
	"  return res;\n" +
	"}\n" +
	"function __sui_path_dirname(p) {\n" +
	"  var v = String(p == null ? '' : p).replace(/\\\\/g, '/').replace(/\\/+$/, '');\n" +
	"  if (!v) return '.';\n" +
	"  var drive = ''; var dm = /^([A-Za-z]:)(\\/|$)/.exec(v); if (dm) { drive = dm[1]; if (v === drive) return drive + '/'; }\n" +
	"  var idx = v.lastIndexOf('/');\n" +
	"  if (idx < 0) return drive ? drive + '/' : '.';\n" +
	"  if (idx === 0) return '/';\n" +
	"  var d = v.slice(0, idx); if (drive && d === drive) return drive + '/'; return d || '/';\n" +
	"}\n" +
	"function __sui_path_basename(p) {\n" +
	"  var v = String(p == null ? '' : p).replace(/\\\\/g, '/').replace(/\\/+$/, '');\n" +
	"  var idx = v.lastIndexOf('/'); return idx >= 0 ? v.slice(idx + 1) : v;\n" +
	"}\n" +
	"function __sui_path_extname(p) {\n" +
	"  var b = __sui_path_basename(p); var idx = b.lastIndexOf('.');\n" +
	"  return idx > 0 ? b.slice(idx) : '';\n" +
	"}\n" +
	"function __sui_crypto_status() {\n" +
	"  try {\n" +
	"    var c = typeof crypto !== 'undefined' ? crypto : null;\n" +
	"    var hasRandom = !!c && typeof c.getRandomValues === 'function';\n" +
	"    var hasHash = !!c && !!c.subtle && typeof c.subtle.digest === 'function';\n" +
	"    if (!hasRandom) return 'unavailable';\n" +
	"    if (!hasHash) return 'random-only';\n" +
	"    return 'ready';\n" +
	"  } catch (e) { return 'unavailable'; }\n" +
	"}\n" +
	"function __sui_crypto_available() { return __sui_crypto_status() === 'ready'; }\n" +
	"function __sui_crypto_uuid() {\n" +
	"  try { if (typeof crypto !== 'undefined' && crypto.randomUUID) return String(crypto.randomUUID()); } catch (e) {}\n" +
	"  if (__sui_crypto_status() === 'unavailable') throw new Error('crypto.uuid requires crypto.getRandomValues (Web Crypto is unavailable)');\n" +
	"  try {\n" +
	"    var b = new Uint8Array(16); crypto.getRandomValues(b); b[6] = (b[6] & 15) | 64; b[8] = (b[8] & 63) | 128;\n" +
	"    var h = []; for (var i = 0; i < b.length; i++) h.push((b[i] + 256).toString(16).slice(1));\n" +
	"    return h.slice(0,4).join('') + '-' + h.slice(4,6).join('') + '-' + h.slice(6,8).join('') + '-' + h.slice(8,10).join('') + '-' + h.slice(10).join('');\n" +
	"  } catch (e) { throw new Error('crypto.uuid failed: ' + (e && e.message ? e.message : String(e))); }\n" +
	"}\n" +
	"function __sui_crypto_random_bytes(n) {\n" +
	"  var size = Math.floor(Number(n));\n" +
	"  if (!(size > 0) || size > 1024) throw new Error('crypto.randomBytes requires a length between 1 and 1024');\n" +
	"  if (__sui_crypto_status() === 'unavailable') throw new Error('crypto.randomBytes requires crypto.getRandomValues (Web Crypto is unavailable)');\n" +
	"  try { var b = new Uint8Array(size); crypto.getRandomValues(b); return Array.prototype.slice.call(b); }\n" +
	"  catch (e) { throw new Error('crypto.randomBytes failed: ' + (e && e.message ? e.message : String(e))); }\n" +
	"}\n" +
	"async function __sui_crypto_hash(algo, text) {\n" +
	"  var status = __sui_crypto_status();\n" +
	"  if (status !== 'ready') {\n" +
	"    if (status === 'random-only') throw new Error('crypto.hash requires a secure context (HTTPS or localhost): crypto.subtle is unavailable');\n" +
	"    throw new Error('crypto.hash is unavailable: Web Crypto is not supported');\n" +
	"  }\n" +
	"  var name = String(algo == null ? 'sha256' : algo).toUpperCase().replace(/-/g, '');\n" +
	"  if (name === 'SHA2' || name === 'SHA256') name = 'SHA-256';\n" +
	"  else if (name === 'SHA384') name = 'SHA-384';\n" +
	"  else if (name === 'SHA512') name = 'SHA-512';\n" +
	"  else throw new Error('crypto.hash: unsupported algorithm \"' + String(algo) + '\" (use sha256, sha384 or sha512)');\n" +
	"  try {\n" +
	"    var data = new TextEncoder().encode(String(text == null ? '' : text));\n" +
	"    var buf = await crypto.subtle.digest(name, data);\n" +
	"    var out = []; var bytes = new Uint8Array(buf);\n" +
	"    for (var i = 0; i < bytes.length; i++) out.push((bytes[i] + 256).toString(16).slice(1));\n" +
	"    return out.join('');\n" +
	"  } catch (e) {\n" +
	"    throw new Error('crypto.hash failed: ' + (e && e.message ? e.message : String(e)));\n" +
	"  }\n" +
	"}\n" +
	"\n" +
	"// JSON Schema validation: a practical subset (type, properties, required,\n" +
	"// additionalProperties, items, enum/const, bounds, pattern, formats, allOf/anyOf/oneOf/not).\n" +
	"function __sui_schema_kind(v) {\n" +
	"  if (v === null) return 'null';\n" +
	"  if (Array.isArray(v)) return 'array';\n" +
	"  if (typeof v === 'number') return Number.isInteger(v) ? 'integer' : 'number';\n" +
	"  return typeof v;\n" +
	"}\n" +
	"function __sui_schema_equal(a, b) { try { return JSON.stringify(a) === JSON.stringify(b); } catch (e) { return false; } }\n" +
	"function __sui_schema_type_ok(v, t) {\n" +
	"  if (t === 'any') return true;\n" +
	"  if (t === 'null') return v === null;\n" +
	"  if (t === 'array') return Array.isArray(v);\n" +
	"  if (t === 'object') return v !== null && typeof v === 'object' && !Array.isArray(v);\n" +
	"  if (t === 'integer') return typeof v === 'number' && Number.isInteger(v);\n" +
	"  if (t === 'number') return typeof v === 'number' && !isNaN(v);\n" +
	"  return typeof v === t;\n" +
	"}\n" +
	"function __sui_schema_check(v, s, path, depth) {\n" +
	"  if (depth > 64) return path + ': schema nesting is too deep';\n" +
	"  if (s === true) return null;\n" +
	"  if (s === false) return path + ': rejected by false schema';\n" +
	"  if (s === null || typeof s !== 'object' || Array.isArray(s)) return path + ': schema must be an object or boolean';\n" +
	"  if (s.$ref !== undefined) return path + ': $ref is not supported';\n" +
	"  if (s.type !== undefined) {\n" +
	"    var types = Array.isArray(s.type) ? s.type : [s.type];\n" +
	"    var ok = false;\n" +
	"    for (var ti = 0; ti < types.length; ti++) if (__sui_schema_type_ok(v, types[ti])) { ok = true; break; }\n" +
	"    if (!ok) return path + ': expected ' + types.join(' or ') + ', got ' + __sui_schema_kind(v);\n" +
	"  }\n" +
	"  if (s.enum !== undefined) {\n" +
	"    if (!Array.isArray(s.enum)) return path + ': enum must be an array';\n" +
	"    var found = false;\n" +
	"    for (var ei = 0; ei < s.enum.length; ei++) if (__sui_schema_equal(v, s.enum[ei])) { found = true; break; }\n" +
	"    if (!found) return path + ': value is not in enum';\n" +
	"  }\n" +
	"  if (Object.prototype.hasOwnProperty.call(s, 'const') && !__sui_schema_equal(v, s.const)) return path + ': value does not match const';\n" +
	"  if (typeof v === 'number') {\n" +
	"    if (typeof s.minimum === 'number' && v < s.minimum) return path + ': must be >= ' + s.minimum;\n" +
	"    if (typeof s.maximum === 'number' && v > s.maximum) return path + ': must be <= ' + s.maximum;\n" +
	"    if (typeof s.exclusiveMinimum === 'number' && v <= s.exclusiveMinimum) return path + ': must be > ' + s.exclusiveMinimum;\n" +
	"    if (typeof s.exclusiveMaximum === 'number' && v >= s.exclusiveMaximum) return path + ': must be < ' + s.exclusiveMaximum;\n" +
	"    if (typeof s.multipleOf === 'number' && s.multipleOf > 0) {\n" +
	"      var q = v / s.multipleOf;\n" +
	"      if (Math.abs(q - Math.round(q)) > 1e-9) return path + ': must be a multiple of ' + s.multipleOf;\n" +
	"    }\n" +
	"  }\n" +
	"  if (typeof v === 'string') {\n" +
	"    if (typeof s.minLength === 'number' && v.length < s.minLength) return path + ': must have at least ' + s.minLength + ' characters';\n" +
	"    if (typeof s.maxLength === 'number' && v.length > s.maxLength) return path + ': must have at most ' + s.maxLength + ' characters';\n" +
	"    if (typeof s.pattern === 'string') {\n" +
	"      if (s.pattern.length > 512) return path + ': pattern is too long';\n" +
	"      try { if (!(new RegExp(s.pattern)).test(v)) return path + ': does not match the pattern'; } catch (e) { return path + ': invalid pattern'; }\n" +
	"    }\n" +
	"    if (typeof s.format === 'string') {\n" +
	"      if (s.format === 'email' && !/^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/.test(v)) return path + ': must be an email';\n" +
	"      if ((s.format === 'uri' || s.format === 'url') && !__sui_valid_url(v)) return path + ': must be a URI';\n" +
	"      if (s.format === 'uuid' && !/^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$/.test(v)) return path + ': must be a UUID';\n" +
	"      if (s.format === 'date' && !/^\\d{4}-\\d{2}-\\d{2}$/.test(v)) return path + ': must be a date';\n" +
	"      if (s.format === 'date-time' && isNaN(Date.parse(v))) return path + ': must be a date-time';\n" +
	"    }\n" +
	"  }\n" +
	"  if (Array.isArray(v)) {\n" +
	"    if (typeof s.minItems === 'number' && v.length < s.minItems) return path + ': must have at least ' + s.minItems + ' items';\n" +
	"    if (typeof s.maxItems === 'number' && v.length > s.maxItems) return path + ': must have at most ' + s.maxItems + ' items';\n" +
	"    if (s.uniqueItems === true) {\n" +
	"      var seen = {};\n" +
	"      for (var ui = 0; ui < v.length; ui++) {\n" +
	"        var ukey = JSON.stringify(v[ui]);\n" +
	"        if (seen[ukey]) return path + ': items must be unique';\n" +
	"        seen[ukey] = true;\n" +
	"      }\n" +
	"    }\n" +
	"    if (Array.isArray(s.items)) {\n" +
	"      for (var ai = 0; ai < s.items.length && ai < v.length; ai++) {\n" +
	"        var aerr = __sui_schema_check(v[ai], s.items[ai], path + '[' + ai + ']', depth + 1);\n" +
	"        if (aerr !== null) return aerr;\n" +
	"      }\n" +
	"      if (v.length > s.items.length) {\n" +
	"        if (s.additionalItems === false) return path + ': additional items are not allowed';\n" +
	"        if (s.additionalItems && typeof s.additionalItems === 'object') {\n" +
	"          for (var aj = s.items.length; aj < v.length; aj++) {\n" +
	"            var ajerr = __sui_schema_check(v[aj], s.additionalItems, path + '[' + aj + ']', depth + 1);\n" +
	"            if (ajerr !== null) return ajerr;\n" +
	"          }\n" +
	"        }\n" +
	"      }\n" +
	"    } else if (s.items !== undefined) {\n" +
	"      for (var ii = 0; ii < v.length; ii++) {\n" +
	"        var ierr = __sui_schema_check(v[ii], s.items, path + '[' + ii + ']', depth + 1);\n" +
	"        if (ierr !== null) return ierr;\n" +
	"      }\n" +
	"    }\n" +
	"  } else if (v !== null && typeof v === 'object') {\n" +
	"    if (typeof s.minProperties === 'number' && Object.keys(v).length < s.minProperties) return path + ': must have at least ' + s.minProperties + ' properties';\n" +
	"    if (typeof s.maxProperties === 'number' && Object.keys(v).length > s.maxProperties) return path + ': must have at most ' + s.maxProperties + ' properties';\n" +
	"    if (Array.isArray(s.required)) {\n" +
	"      for (var ri = 0; ri < s.required.length; ri++) {\n" +
	"        if (!Object.prototype.hasOwnProperty.call(v, s.required[ri])) return path + ': missing required property \"' + s.required[ri] + '\"';\n" +
	"      }\n" +
	"    }\n" +
	"    var props = (s.properties && typeof s.properties === 'object') ? s.properties : {};\n" +
	"    var keys = Object.keys(v);\n" +
	"    for (var ki = 0; ki < keys.length; ki++) {\n" +
	"      var key = keys[ki];\n" +
	"      if (props[key] !== undefined) {\n" +
	"        var perr = __sui_schema_check(v[key], props[key], path + '.' + key, depth + 1);\n" +
	"        if (perr !== null) return perr;\n" +
	"      } else if (s.additionalProperties === false) {\n" +
	"        return path + ': unexpected property \"' + key + '\"';\n" +
	"      } else if (s.additionalProperties && typeof s.additionalProperties === 'object') {\n" +
	"        var xerr = __sui_schema_check(v[key], s.additionalProperties, path + '.' + key, depth + 1);\n" +
	"        if (xerr !== null) return xerr;\n" +
	"      }\n" +
	"    }\n" +
	"  }\n" +
	"  if (Array.isArray(s.allOf)) {\n" +
	"    for (var oi = 0; oi < s.allOf.length; oi++) {\n" +
	"      var oerr = __sui_schema_check(v, s.allOf[oi], path, depth + 1);\n" +
	"      if (oerr !== null) return oerr;\n" +
	"    }\n" +
	"  }\n" +
	"  if (Array.isArray(s.anyOf)) {\n" +
	"    var anyOk = false;\n" +
	"    for (var yi = 0; yi < s.anyOf.length; yi++) if (__sui_schema_check(v, s.anyOf[yi], path, depth + 1) === null) { anyOk = true; break; }\n" +
	"    if (!anyOk) return path + ': does not match anyOf';\n" +
	"  }\n" +
	"  if (Array.isArray(s.oneOf)) {\n" +
	"    var oneCount = 0;\n" +
	"    for (var wi = 0; wi < s.oneOf.length; wi++) if (__sui_schema_check(v, s.oneOf[wi], path, depth + 1) === null) oneCount++;\n" +
	"    if (oneCount !== 1) return path + ': must match exactly one oneOf schema';\n" +
	"  }\n" +
	"  if (s.not !== undefined) {\n" +
	"    if (__sui_schema_check(v, s.not, path, depth + 1) === null) return path + ': must not match the not schema';\n" +
	"  }\n" +
	"  if (s.if !== undefined) {\n" +
	"    var cond = __sui_schema_check(v, s.if, path, depth + 1);\n" +
	"    if (cond === null && s.then !== undefined) {\n" +
	"      var terr = __sui_schema_check(v, s.then, path, depth + 1);\n" +
	"      if (terr !== null) return terr;\n" +
	"    } else if (cond !== null && s.else !== undefined) {\n" +
	"      var eerr = __sui_schema_check(v, s.else, path, depth + 1);\n" +
	"      if (eerr !== null) return eerr;\n" +
	"    }\n" +
	"  }\n" +
	"  return null;\n" +
	"}\n" +
	"function __sui_json_validate(value, schema) {\n" +
	"  try { if (typeof schema === 'string') schema = JSON.parse(schema); } catch (e) { return [false, '$: schema is not valid JSON']; }\n" +
	"  var err = __sui_schema_check(value, schema, '$', 0);\n" +
	"  return err === null ? [true, ''] : [false, err];\n" +
	"}\n" +
	"\n" +
	"const __SUI_MEM_STORE = Object.create(null);\n" +
	"var __SUI_STORAGE_SYNC = false;\n" +
	"function __sui_storage_get(k, def) {\n" +
	"  var key = String(k); var fallback = def === undefined ? null : def;\n" +
	"  try { var v = localStorage.getItem(key); if (v === null) return fallback; try { return JSON.parse(v); } catch (e) { return v; } }\n" +
	"  catch (e) { if (key in __SUI_MEM_STORE) { try { return JSON.parse(__SUI_MEM_STORE[key]); } catch (ee) { return __SUI_MEM_STORE[key]; } } return fallback; }\n" +
	"}\n" +
	"function __sui_storage_set(k, v) {\n" +
	"  var key = String(k); var sv = typeof v === 'string' ? v : JSON.stringify(v);\n" +
	"  try { localStorage.setItem(key, sv); return true; } catch (e) { __SUI_MEM_STORE[key] = sv; return false; }\n" +
	"}\n" +
	"function __sui_storage_remove(k) {\n" +
	"  var key = String(k);\n" +
	"  try { localStorage.removeItem(key); return true; } catch (e) { delete __SUI_MEM_STORE[key]; return false; }\n" +
	"}\n" +
	"function __sui_storage_clear() {\n" +
	"  try { localStorage.clear(); return true; } catch (e) { for (var kk in __SUI_MEM_STORE) delete __SUI_MEM_STORE[kk]; return false; }\n" +
	"}\n" +
	"function __sui_storage_keys() {\n" +
	"  try { var r = []; for (var i = 0; i < localStorage.length; i++) { var k = localStorage.key(i); if (k) r.push(k); } return r; }\n" +
	"  catch (e) { return Object.keys(__SUI_MEM_STORE); }\n" +
	"}\n" +
	"async function __sui_storage_quota() {\n" +
	"  var used = 0, quota = 0;\n" +
	"  try { for (var i = 0; i < localStorage.length; i++) { var k = localStorage.key(i) || ''; used += (k.length + String(localStorage.getItem(k) || '').length) * 2; } } catch (e) {}\n" +
	"  try {\n" +
	"    if (navigator.storage && navigator.storage.estimate) {\n" +
	"      var estimate = navigator.storage.estimate();\n" +
	"      var timeout = new Promise(function (resolve) { setTimeout(function () { resolve(null); }, 1000); });\n" +
	"      var est = await Promise.race([estimate, timeout]);\n" +
	"      if (est) { if (est.usage !== undefined) used = Number(est.usage) || used; if (est.quota !== undefined) quota = Number(est.quota) || 0; }\n" +
	"    }\n" +
	"  } catch (e) {}\n" +
	"  return { used: used, quota: quota };\n" +
	"}\n" +
	"function __sui_storage_sync() {\n" +
	"  if (__SUI_STORAGE_SYNC) return true;\n" +
	"  try { if (typeof window === 'undefined' || !window.addEventListener) return false;\n" +
	"    window.addEventListener('storage', function () { if (typeof __sui_render === 'function') __sui_render(); });\n" +
	"    __SUI_STORAGE_SYNC = true; return true;\n" +
	"  } catch (e) { return false; }\n" +
	"}\n" +
	"\n" +
	"function __sui_clip_write_multi(items){try{return {written:true};}catch(e){return {written:false,error:String(e)};}}\n" +
	"function __sui_clip_write(t){ return (async()=>{ try{ if(navigator.clipboard&&navigator.clipboard.writeText){ await navigator.clipboard.writeText(String(t||'')); return [true,null] } return [false,'clipboard not available'] }catch(e){ return [false,String(e)] } })(); }\n" +
	"\n" +
	"function __sui_valid_email(s){ const v=String(s||''); return /^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/.test(v); }\n" +
	"function __sui_valid_url(s){ try{ new URL(String(s||'')); return true }catch(e){ return false } }\n" +
	"function __sui_valid_empty(v){ if(v===null||v===undefined) return true; if(typeof v==='string') return v.trim()===''; if(Array.isArray(v)) return v.length===0; if(typeof v==='object') return Object.keys(v).length===0; return false; }\n" +
	"function __sui_valid_required(v){ return !__sui_valid_empty(v); }\n" +
	"function __sui_valid_minlen(v,l){ const n=Array.isArray(v)?v.length:(typeof v==='string'?v.length:Object.keys(v||{}).length); return n>=Number(l||0); }\n" +
	"function __sui_valid_maxlen(v,l){ const n=Array.isArray(v)?v.length:(typeof v==='string'?v.length:Object.keys(v||{}).length); return n<=Number(l||0); }\n" +
	"function __sui_valid_length(v,a,b){ const n=Array.isArray(v)?v.length:(typeof v==='string'?v.length:Object.keys(v||{}).length); const lo=Number(a||0), hi=b===undefined?lo:Number(b); if(b===undefined) return n===lo; return n>=lo&&n<=hi; }\n" +
	"function __sui_valid_pattern(s,pat){ try{ const r=new RegExp(pat); return r.test(String(s||'')) }catch(e){ return false } }\n" +
	"function __sui_valid_isnum(v){ return typeof v==='number'&&!isNaN(v); }\n" +
	"function __sui_valid_isbool(v){ return typeof v==='boolean'; }\n" +
	"function __sui_valid_isstr(v){ return typeof v==='string'; }\n" +
	"function __sui_valid_isarr(v){ return Array.isArray(v); }\n" +
	"function __sui_valid_isobj(v){ return v!==null&&typeof v==='object'&&!Array.isArray(v); }\n" +
	"\n" +
	"function __sui_regex_flags(f){ if(!f) return 'g'; let s=''; if(f.indexOf('i')>=0) s+='i'; if(f.indexOf('m')>=0) s+='m'; if(f.indexOf('s')>=0) s+='s'; if(f.indexOf('u')>=0) s+='u'; if(f.indexOf('y')>=0) s+='y'; if(s.indexOf('g')<0) s+='g'; return s; }\n" +
	"function __sui_regex(pat,flags){ try{ return new RegExp(String(pat||''), __sui_regex_flags(flags)); }catch(e){ try{return new RegExp(String(pat||''),'g')}catch(ee){ return null } } }\n" +
	"function __sui_replace_regex(s,pat,repl,flags){ const r=__sui_regex(pat,flags||'g'); if(!r) return String(s||''); try{ return String(s||'').replace(r, repl===undefined?'':String(repl)) }catch(e){ return String(s||'') } }\n" +
	"function __sui_match(s,pat,flags){ const r=__sui_regex(pat,flags||''); if(!r) return []; try{ const m=String(s||'').match(r); return m||[] }catch(e){ return [] } }\n" +
	"function __sui_match_all(s,pat,flags){ const f=flags||'g'; if(f.indexOf('g')<0){ const r=__sui_regex(pat,f+'g'); if(!r) return []; try{ const it=String(s||'').matchAll(r); return Array.from(it) }catch(e){ return [] } } const r2=__sui_regex(pat,f); if(!r2) return []; try{ const it=String(s||'').matchAll(r2); return Array.from(it) }catch(e){ return [] } }\n" +
	"function __sui_search(s,pat){ try{ const r=__sui_regex(pat,''); if(!r) return -1; return String(s||'').search(r) }catch(e){ return -1 } }\n" +
	"function __sui_test(s,pat){ try{ const r=__sui_regex(pat,''); if(!r) return false; return r.test(String(s||'')) }catch(e){ return false } }\n" +
	"function __sui_split_regex(s,pat,limit){ try{ const r=__sui_regex(pat,'g'); if(!r) return String(s||'').split(''); return String(s||'').split(r, limit===undefined?undefined:Number(limit)) }catch(e){ return String(s||'').split(String(pat||''), limit===undefined?undefined:Number(limit)) } }\n" +
	"function __sui_extract(s,pat){ const m=__sui_match(s,pat,''); if(m&&m.length>1) return m[1]; if(m&&m.length===1) return m[0]; return '' }\n" +
	"function __sui_extract_all(s,pat){ const ms=__sui_match_all(s,pat,'g'); const out=[]; for(let i=0;i<ms.length;i++){ const mm=ms[i]; if(mm&&mm.length>1) out.push(mm[1]); else if(mm&&mm.length===1) out.push(mm[0]); } return out }"

var SuiWsClass = "class __SUI_WS {\n" +
	"  constructor(url, opts) {\n" +
	"    this.url = url;\n" +
	"    this.opts = opts || {};\n" +
	"    this.autoReconnect = Boolean(this.opts.reconnect === true || this.opts.reconnect === 'true' || this.opts.reconnect === 1 || this.opts.reconnect === '1');\n" +
	"    this.reconnectDelay = Number(this.opts.delay || this.opts.reconnectDelay || 1000);\n" +
	"    this.currentDelay = this.reconnectDelay;\n" +
	"    this.maxReconnectDelay = Number(this.opts.maxDelay || this.opts.maxReconnectDelay || 30000);\n" +
	"    this.heartbeatInterval = Number(this.opts.heartbeat || this.opts.heartbeatInterval || 0);\n" +
	"    this.pingMsg = String(this.opts.ping || this.opts.pingMessage || 'ping');\n" +
	"    this.manualClose = false;\n" +
	"    this.reconnectTimer = null;\n" +
	"    this.heartbeatTimer = null;\n" +
	"    this._openHandlers = [];\n" +
	"    this._closeHandlers = [];\n" +
	"    this._errorHandlers = [];\n" +
	"    this.sock = new WebSocket(url);\n" +
	"    this._bindSock();\n" +
	"  }\n" +
	"  _bindSock(prev) {\n" +
	"    if (!this.sock) return;\n" +
	"    if (prev && prev.onmessage) this.sock.onmessage = prev.onmessage;\n" +
	"    var oldOpen = this.sock.onopen;\n" +
	"    this.sock.onopen = (ev) => {\n" +
	"      this.currentDelay = this.reconnectDelay;\n" +
	"      this._startHeartbeat();\n" +
	"      if (oldOpen) { try { oldOpen(ev); } catch(e) {} }\n" +
	"      for (var i = 0; i < this._openHandlers.length; i++) { try { this._openHandlers[i](); } catch(e) {} }\n" +
	"    };\n" +
	"    this.sock.onclose = (ev) => {\n" +
	"      this._stopHeartbeat();\n" +
	"      for (var i = 0; i < this._closeHandlers.length; i++) { try { this._closeHandlers[i](); } catch(e) {} }\n" +
	"      if (!this.manualClose && this.autoReconnect) { this._scheduleReconnect(); }\n" +
	"    };\n" +
	"    this.sock.onerror = (ev) => {\n" +
	"      this._handleError(String(ev));\n" +
	"    };\n" +
	"  }\n" +
	"  _init() {\n" +
	"    if (this.manualClose) return;\n" +
	"    try {\n" +
	"      var prev = this.sock;\n" +
	"      var url = this.url;\n" +
	"      this.sock = new WebSocket(url);\n" +
	"      this._bindSock(prev);\n" +
	"    } catch(e) {\n" +
	"      this._handleError(String(e));\n" +
	"      this._scheduleReconnect();\n" +
	"    }\n" +
	"  }\n" +
	"  _scheduleReconnect() {\n" +
	"    if (this.manualClose || this.reconnectTimer) return;\n" +
	"    this.reconnectTimer = setTimeout(() => {\n" +
	"      this.reconnectTimer = null;\n" +
	"      this._init();\n" +
	"    }, this.currentDelay);\n" +
	"    this.currentDelay = Math.min(this.currentDelay * 1.5, this.maxReconnectDelay);\n" +
	"  }\n" +
	"  _startHeartbeat() {\n" +
	"    this._stopHeartbeat();\n" +
	"    if (this.heartbeatInterval > 0) {\n" +
	"      this.heartbeatTimer = setInterval(() => {\n" +
	"        if (this.open()) { try { this.sock.send(this.pingMsg); } catch(e) {} }\n" +
	"      }, this.heartbeatInterval);\n" +
	"    }\n" +
	"  }\n" +
	"  _stopHeartbeat() {\n" +
	"    if (this.heartbeatTimer) { clearInterval(this.heartbeatTimer); this.heartbeatTimer = null; }\n" +
	"  }\n" +
	"  _handleError(err) {\n" +
	"    for (var i = 0; i < this._errorHandlers.length; i++) { try { this._errorHandlers[i](err); } catch(e) {} }\n" +
	"  }\n" +
	"  send(data) { if (this.sock) this.sock.send(data); }\n" +
	"  close() {\n" +
	"    this.manualClose = true;\n" +
	"    if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null; }\n" +
	"    this._stopHeartbeat();\n" +
	"    if (this.sock) this.sock.close();\n" +
	"  }\n" +
	"  reconnect() {\n" +
	"    this.manualClose = false;\n" +
	"    if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null; }\n" +
	"    if (this.sock) { try { this.sock.close(); } catch(e) {} }\n" +
	"    this._init();\n" +
	"  }\n" +
	"  open() { return Boolean(this.sock && this.sock.readyState === WebSocket.OPEN); }\n" +
	"  status() {\n" +
	"    if (!this.sock) return 'closed';\n" +
	"    if (this.sock.readyState === 0) return 'connecting';\n" +
	"    if (this.sock.readyState === 1) return 'open';\n" +
	"    if (this.sock.readyState === 2) return 'closing';\n" +
	"    return 'closed';\n" +
	"  }\n" +
	"  onMessage(fn) { this.sock.onmessage = (ev) => fn(ev.data); }\n" +
	"  onClose(fn) { this._closeHandlers.push(fn); }\n" +
	"  onError(fn) { this._errorHandlers.push(fn); }\n" +
	"  onOpen(fn) { this._openHandlers.push(fn); if (this.open()) { try { fn(); } catch(e) {} } }\n" +
	"}"

type JsCtx struct {
	states []string
	scope  []string
	tuples int
}

type Wire struct {
	idx   int
	lines []string
}

type ViewGen struct {
	count int
	wires []Wire
	refs  []string
}

func GenerateJs(p *Prog) string {
	return GenerateJsFor(p, "web")
}

// genUserFuncs holds the user function names of the program currently being
// generated; a user declaration shadows a same-named built-in.
var genUserFuncs = map[string]bool{}

func isUserFunc(name string) bool {
	return genUserFuncs[name]
}

// genAsyncFuncs tracks which user functions are async, so call sites that
// must await (tuple binding) can skip the await for sync callees.
var genAsyncFuncs = map[string]bool{}

// genFnDefaults holds, per user function, the JS expression of each
// parameter's default value ("" when the parameter is required). Call sites
// append the missing trailing defaults.
var genFnDefaults = map[string][]string{}

// genFnParams holds parameter names in declaration order for named calls.
var genFnParams = map[string][]string{}

// genEnums holds enum members for `value is Enum` runtime guards.
var genEnums = map[string][]string{}

// genComputedDeps maps a computed value to the states/computed values it
// reads. Views bind to those underlying states so a computed change still
// re-renders the right views without rendering unrelated ones.
var genComputedDeps = map[string][]string{}
var genStateNames = map[string]bool{}

// genEmitSourceMarkers enables /*@sui-src:...*/ markers during generation.
// It is set only by GenerateJsWithSourceMap and reset by generateJsFor.
var genEmitSourceMarkers bool

// Component/view bookkeeping shared by generation of one program.
var genViewParams = map[string][]Param{} // view name -> component parameters
var genViews = map[string]ViewDecl{}     // view name -> declaration (for refs/cycles)
var genCompBody = false                  // generating a component chunk body
var genCompFn = map[string]bool{}        // fn-typed params of current component
var genCompInst = map[string]bool{}      // instance-state vars of current component
var genCompMethods = map[string]bool{}   // view-method names of current component
var genViewInst = map[string]bool{}      // view name -> has instance states
var genClassFilter = true                // __sui_class allow-list (security config)
var genIframeCfg = DefaultSecurity()     // iframe defaults (security config)
var genSiteId = 0                        // global invocation-site counter
var refExcl = map[string]bool{}          // component parameter exclusion for refs

var refExclSaved = map[string]bool{}

// GenerateJsFor renders the application script for a build target ("web" or
// "desktop"). The two differ only in the std intrinsics: on desktop, fs and
// exec operations go through the native bridge (same-origin POST endpoints),
// guarded by a runtime secret that the server injects at serve time.
//
// Label JS generation: the generator keeps its bookkeeping in package-level
// maps (genUserFuncs, genViews, ...), so two concurrent HTTP handlers
// generating at once would race on them — a concurrent map write kills the
// whole process. Serialization is enough: generation is fast and local.
var genMu sync.Mutex

// cssClassTokenRe finds class selectors in user CSS (also inside :hover,
// :not(), media queries, etc.). False positives (e.g. a dot in a URL) only
// widen the allow-list, never the opposite.
var cssClassTokenRe = regexp.MustCompile(`\.(-?[_a-zA-Z][_a-zA-Z0-9-]*)`)

// securityClassNames collects every class the app can legitimately use:
// classes defined in its CSS and literal tokens in static class attributes
// (including the static parts of "btn {state}" templates).
func securityClassNames(p *Prog) []string {
	set := map[string]bool{}
	if p != nil {
		for _, css := range p.css {
			for _, m := range cssClassTokenRe.FindAllStringSubmatch(css, -1) {
				if len(m) > 1 && m[1] != "" {
					set[m[1]] = true
				}
			}
		}
		for i := 0; i < len(p.views); i++ {
			var classes []string
			collectClassesInView(p.views[i].body, &classes)
			for _, c := range classes {
				if c != "" {
					set[c] = true
				}
			}
		}
	}
	out := make([]string, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

func GenerateJsFor(p *Prog, target string) string {
	return generateJsFor(p, target, false)
}

// GenerateJsWithSourceMap renders the script with source markers, strips them
// and returns the clean JavaScript plus a Source Map v3 JSON document.
func GenerateJsWithSourceMap(p *Prog, target, sourceBase string) (string, string) {
	marked := generateJsFor(p, target, true)
	return SourceMapFromMarked(marked, SourceMapOptions{
		GeneratedFile:  "app.js",
		SourceBase:     sourceBase,
		IncludeContent: true,
	})
}

func generateJsFor(p *Prog, target string, markers bool) string {
	genMu.Lock()
	defer genMu.Unlock()
	genEmitSourceMarkers = markers
	defer func() { genEmitSourceMarkers = false }()
	sec := SecurityConfig{}
	if p != nil {
		sec = p.security.Normalized()
	} else {
		sec = DefaultSecurity()
	}
	genClassFilter = sec.ClassDynamic != "allow"
	genIframeCfg = sec
	genUserFuncs = map[string]bool{}
	genAsyncFuncs = map[string]bool{}
	genFnDefaults = map[string][]string{}
	genFnParams = map[string][]string{}
	genStateNames = map[string]bool{}
	for i := 0; i < len(p.states); i++ {
		genStateNames[p.states[i].name] = true
	}
	genComputedDeps = map[string][]string{}
	for i := 0; i < len(p.computed); i++ {
		cd := &p.computed[i]
		acc := &RefAcc{names: map[string]bool{}}
		collectExprRefs(cd.init, acc)
		var deps []string
		for name := range acc.names {
			if genStateNames[name] || genComputedDeps[name] != nil {
				deps = append(deps, name)
			}
		}
		genComputedDeps[cd.name] = sortStringList(deps)
	}
	genEnums = map[string][]string{}
	for i := 0; i < len(p.enums); i++ {
		genEnums[p.enums[i].name] = p.enums[i].members
	}
	for i := 0; i < len(p.funcs); i++ {
		f := p.funcs[i]
		genUserFuncs[f.name] = true
		if f.async {
			genAsyncFuncs[f.name] = true
		}
		var defs []string
		var names []string
		for j := 0; j < len(f.params); j++ {
			names = append(names, f.params[j].name)
			if f.params[j].dflt != nil {
				defs = append(defs, ExprToJs(f.params[j].dflt))
			} else {
				defs = append(defs, "")
			}
		}
		genFnDefaults[f.name] = defs
		genFnParams[f.name] = names
	}
	// View-local methods are called like functions inside their component and
	// support the same named-argument reordering.
	for i := 0; i < len(p.views); i++ {
		for j := 0; j < len(p.views[i].methods); j++ {
			m := &p.views[i].methods[j]
			genUserFuncs[m.name] = true
			if m.async {
				genAsyncFuncs[m.name] = true
			}
			var defs []string
			var names []string
			for k := 0; k < len(m.params); k++ {
				names = append(names, m.params[k].name)
				if m.params[k].dflt != nil {
					defs = append(defs, ExprToJs(m.params[k].dflt))
				} else {
					defs = append(defs, "")
				}
			}
			genFnDefaults[m.name] = defs
			genFnParams[m.name] = names
		}
	}
	genViewParams = map[string][]Param{}
	genViews = map[string]ViewDecl{}
	genViewInst = map[string]bool{}
	for i := 0; i < len(p.views); i++ {
		genViewParams[p.views[i].name] = p.views[i].params
		genViews[p.views[i].name] = p.views[i]
		if len(p.views[i].inst_state) > 0 {
			genViewInst[p.views[i].name] = true
		}
	}
	genSiteId = 0
	refExcl = map[string]bool{}

	// Component detection: a view is a *component* when another view invokes
	// it (its name appears as a tag); every non-invoked view is a *root* that
	// renders into #app and fires its onLoad.
	invoked := map[string]bool{}
	for i := 0; i < len(p.views); i++ {
		var kids []string
		collectComponentNamesOf(p.views[i].body, &kids)
		for j := 0; j < len(kids); j++ {
			invoked[kids[j]] = true
		}
	}
	isRoot := func(i int) bool {
		return !invoked[p.views[i].name]
	}

	var ctx JsCtx
	for i := 0; i < len(p.states); i++ {
		ctx.states = append(ctx.states, p.states[i].name)
	}
	var sb strings.Builder
	sb.WriteString(SuiPrelude)
	if wsImported(p) {
		sb.WriteString("\n\n" + SuiWsClass)
	}
	if target == "desktop" {
		sb.WriteString("\n\nconst __SUI_TOKEN = \"" + TokenPlaceholder + "\";")
	}
	// Table helpers must exist before the first render: every renderer ends
	// with __SUI_TBL.afterRender(), which re-applies displayed rows and wires
	// editable cells after the view HTML has been replaced.
	// Security runtime policy (table limits, CSV default). Emitted before the
	// table runtime, whose constants read it.
	sb.WriteString("\n\nconst __SUI_SEC = {maxRows:" + strconv.Itoa(sec.MaxRows) +
		",maxCols:" + strconv.Itoa(sec.MaxCols) +
		",maxText:" + strconv.Itoa(sec.MaxText) +
		",csvFormulaEscape:" + strconv.FormatBool(sec.CSVFormulaEscape) + "};")
	// Only the runtimes of the components the program actually uses are
	// embedded, so an app without charts does not ship the 1.2 MB chart
	// runtime. scanProgFeatures walks the AST once; rendererJs uses the same
	// result to guard the after-render hooks.
	feat := scanProgFeatures(p)
	if feat.needsTable() {
		sb.WriteString("\n\n")
		sb.WriteString(suiTableRuntime)
	}
	// Drag-and-drop runtime. It is independent of the table runtime: the
	// generated app wires drag sources/targets by data attributes, while the
	// language exposes the captured keys through drag.source()/drag.target().
	if feat.needsDnd() {
		sb.WriteString("\n\n")
		sb.WriteString(suiDndRuntime)
	}
	// Menu bars use the same delegated runtime in every view and across
	// rerenders; labels/actions are generated as normal elements, not HTML.
	// Anchored overlays share viewport placement and animation helpers.
	if feat.needsMotion() {
		sb.WriteString("\n\n")
		sb.WriteString(suiMotionRuntime)
	}
	if feat.needsMotionGraph() {
		sb.WriteString("\n\n")
		sb.WriteString(suiMotionGraphRuntime)
	}
	if feat.needsOverlay() {
		sb.WriteString("\n\n")
		sb.WriteString(suiOverlayRuntime)
	}
	if feat.needsMenu() {
		sb.WriteString("\n\n")
		sb.WriteString(suiMenuRuntime)
	}
	// Tabs share the same delegated/runtime approach and survive view rerenders
	// through data-sui-tabs-key.
	if feat.needsTabs() {
		sb.WriteString("\n\n")
		sb.WriteString(suiTabsRuntime)
	}
	// Accordions use the same delegated pattern and keep the open item keyed
	// across view re-renders.
	if feat.needsAccordion() {
		sb.WriteString("\n\n")
		sb.WriteString(suiAccordionRuntime)
	}
	// Toast hosts receive safe text notifications through toast.show().
	if feat.needsToast() {
		sb.WriteString("\n\n")
		sb.WriteString(suiToastRuntime)
	}
	// Tooltips use delegated hover/focus events and text-only bubbles.
	if feat.needsTooltip() {
		sb.WriteString("\n\n")
		sb.WriteString(suiTooltipRuntime)
	}
	// Searchable combobox controls filter their own listbox and sync a hidden
	// value binding back to application state.
	if feat.needsCombobox() {
		sb.WriteString("\n\n")
		sb.WriteString(suiComboboxRuntime)
	}
	// Date/time pickers wrap a native input and synchronize the bound state.
	if feat.needsDatePicker() {
		sb.WriteString("\n\n")
		sb.WriteString(suiDatePickerRuntime)
	}
	// Trees own expansion, roving focus and selection; open state is persisted.
	if feat.needsTree() {
		sb.WriteString("\n\n")
		sb.WriteString(suiTreeRuntime)
	}
	// Rich text editors keep a JSON document model, history and a configurable
	// toolbar/plugin registry. The embedded PDF fonts travel with the editor.
	if feat.needsRichText() {
		sb.WriteString("\n\n")
		sb.WriteString(suiPdfFontsJS)
		sb.WriteString("\n\n")
		sb.WriteString(suiRichTextRuntime)
	}
	// Charts parse JSON/XML/CSV data and draw inert SVG nodes. Large charts
	// scroll their viewport or can be zoomed by the user.
	if feat.needsChart() {
		sb.WriteString("\n\n")
		sb.WriteString(suiChartRuntime)
	}
	// Modal dialogs are opened by data-sui-modal name and use delegated events.
	if feat.needsModal() {
		sb.WriteString("\n\n")
		sb.WriteString(suiModalRuntime)
	}
	// Popovers toggle an anchored panel and close on outside click/Escape.
	if feat.needsPopover() {
		sb.WriteString("\n\n")
		sb.WriteString(suiPopoverRuntime)
	}
	// State-controlled drawers close through the state setter or onClose.
	if feat.needsDrawer() {
		sb.WriteString("\n\n")
		sb.WriteString(suiDrawerRuntime)
	}
	// Splitters resize their two panes with pointer drags or arrow keys.
	if feat.needsSplitter() {
		sb.WriteString("\n\n")
		sb.WriteString(suiSplitterRuntime)
	}
	// Steppers switch the bound step index.
	if feat.needsStepper() {
		sb.WriteString("\n\n")
		sb.WriteString(suiStepperRuntime)
	}
	// Carousels move a slide track and optionally bind the active slide.
	if feat.needsCarousel() {
		sb.WriteString("\n\n")
		sb.WriteString(suiCarouselRuntime)
	}
	// Command palettes filter commands and handle their overlay/hotkey.
	if feat.needsCommandBar() {
		sb.WriteString("\n\n")
		sb.WriteString(suiCommandbarRuntime)
	}
	// Class allow-list: dynamic class expressions are filtered to these names
	// so untrusted data cannot activate arbitrary app classes.
	sb.WriteString("\n\nconst __SUI_CLASSES = Object.create(null);\n")
	for _, cn := range securityClassNames(p) {
		sb.WriteString("__SUI_CLASSES[" + jsString(cn) + "] = 1;\n")
	}
	// Translation catalogs are available before state initializers and views,
	// so t(...) also works in initial expressions.
	if init := i18nInitJs(p); init != "" {
		sb.WriteString("\n\n")
		sb.WriteString(init)
	}
	// Enums are frozen objects of string members so Enum.Member expressions
	// work as ordinary JS values and switch uses strict equality.
	for i := 0; i < len(p.enums); i++ {
		e := p.enums[i]
		sb.WriteString("\n\n" + srcMarker(e.span))
		sb.WriteString("const " + e.name + " = Object.freeze({")
		for j := 0; j < len(e.members); j++ {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(jsString(e.members[j]) + ": " + jsString(e.members[j]))
		}
		sb.WriteString("});")
	}
	for i := 0; i < len(p.consts); i++ {
		cd := p.consts[i]
		sb.WriteString("\n\n" + srcMarker(cd.span))
		sb.WriteString("const " + cd.name + " = " + ExprToJs(cd.init) + ";")
	}
	for i := 0; i < len(p.states); i++ {
		sd := p.states[i]
		sb.WriteString("\n\n")
		sb.WriteString(srcMarker(sd.span))
		sb.WriteString("let " + sd.name + " = " + ExprToJs(sd.init) + ";\n")
		sb.WriteString("function " + setterName(sd.name) + "(v) { " + sd.name + " = v; __sui_react(" + jsString(sd.name) + "); }")
	}
	for i := 0; i < len(p.computed); i++ {
		cd := p.computed[i]
		sb.WriteString("\n\n" + srcMarker(cd.span))
		sb.WriteString("let " + cd.name + ";")
		sb.WriteString("\n__SUI_COMPUTED.push(function(){ " + cd.name + " = (" + ExprToJs(cd.init) + "); });")
	}
	if len(p.states) > 0 {
		sb.WriteString("\n\nwindow.__SUI_STATE_GET = Object.create(null);\n")
		for i := 0; i < len(p.states); i++ {
			sd := p.states[i]
			sb.WriteString("window.__SUI_STATE_GET[" + jsString(sd.name) + "] = function(){ return " + sd.name + "; };\n")
		}
	}
	for i := 0; i < len(p.funcs); i++ {
		sb.WriteString("\n\n")
		sb.WriteString(funcJs(&p.funcs[i], &ctx))
	}
	sb.WriteString("\n\nwindow.__SUI_USER_FUNCS = Object.create(null);\n")
	for i := 0; i < len(p.funcs); i++ {
		sb.WriteString("window.__SUI_USER_FUNCS[" + jsString(p.funcs[i].name) + "] = 1;\n")
	}
	stdObj := stdObjJs(p, target)
	if stdObj != "" {
		sb.WriteString("\n\n" + stdObj)
	}
	// Rendering functions: chunks for components, full renderers for roots.
	for i := 0; i < len(p.views); i++ {
		if isRoot(i) {
			continue
		}
		sb.WriteString("\n\n")
		sb.WriteString(componentChunkJs(p, &p.views[i], &ctx))
	}
	// Page support: with several non-invoked views each root view renders into
	// its own page box; the program boots on the FIRST root view and further
	// pages become visible through __sui_page(name).
	rootNames := []string{}
	for i := 0; i < len(p.views); i++ {
		if isRoot(i) {
			rootNames = append(rootNames, p.views[i].name)
		}
	}
	multi := len(rootNames) > 1
	// Multi-page previews keep every root view in the DOM at once; carrying
	// the DFS counter across views keeps data-sui-id unique document-wide.
	// Without this, page 2+ re-uses 0.. and the design tool's picker, the
	// runtime wiring queries and highlight targeting all resolve the wrong
	// element (page 1's node with the same id, which is display:none).
	base := 0
	for i := 0; i < len(p.views); i++ {
		if !isRoot(i) {
			continue
		}
		sb.WriteString("\n\n")
		sb.WriteString(rendererJs(p, &p.views[i], &ctx, multi, &base, feat))
	}
	if multi {
		sb.WriteString("\n\nwindow.__SUI_PAGE = " + jsString(rootNames[0]) + ";\n")
		sb.WriteString("function __sui_page(n) {\n")
		sb.WriteString("  if (typeof __SUI_MOTION !== 'undefined' && __SUI_MOTION.capturePage) __SUI_MOTION.capturePage();\n")
		sb.WriteString("  window.__SUI_PAGE = String(n);\n")
		sb.WriteString("  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();\n")
		sb.WriteString("  if (typeof __SUI_MOTION !== 'undefined' && __SUI_MOTION.flushShared) __SUI_MOTION.flushShared();\n")
		sb.WriteString("}\n")
	}
	sb.WriteString("\n\n__sui_render();\n")
	// A root view may declare `onLoad: fn` — invoke it once after the first
	// render, so async data-loaders can populate the UI immediately.
	for i := 0; i < len(p.views); i++ {
		if isRoot(i) && p.views[i].onLoad != "" {
			sb.WriteString("\n")
			sb.WriteString(p.views[i].onLoad)
			sb.WriteString("();\n")
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

// i18nInitJs serializes the i18n declarations into the runtime catalogs. The
// first declared locale is the fallback (the designer keeps it first).
func i18nInitJs(p *Prog) string {
	if p == nil || len(p.i18n) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("__sui_i18n_init({")
	for i := 0; i < len(p.i18n); i++ {
		d := p.i18n[i]
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(jsString(d.locale))
		sb.WriteString(":{")
		for j := 0; j < len(d.messages); j++ {
			m := d.messages[j]
			if j > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(jsString(m.key))
			sb.WriteString(":")
			if len(m.plural) > 0 {
				sb.WriteString("{")
				first := true
				for _, cat := range i18nPluralCategories {
					v, ok := m.plural[cat]
					if !ok {
						continue
					}
					if !first {
						sb.WriteString(",")
					}
					first = false
					sb.WriteString(jsString(cat))
					sb.WriteString(":")
					sb.WriteString(jsString(v))
				}
				sb.WriteString("}")
			} else {
				sb.WriteString(jsString(m.text))
			}
		}
		sb.WriteString("}")
	}
	sb.WriteString("}, ")
	sb.WriteString(jsString(p.i18n[0].locale))
	sb.WriteString(");")
	return sb.String()
}

// collectComponentNamesOf gathers the names of components invoked in a view
// body (elements whose tag names a declared component).
func collectComponentNamesOf(nodes []ViewNode, out *[]string) {
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		if n.kind == "element" && n.elem != nil {
			if genViewParams[n.elem.tag] != nil {
				*out = append(*out, n.elem.tag)
			}
			collectComponentNamesOf(n.elem.children, out)
		} else if n.kind == "if" {
			collectComponentNamesOf(derefViewNodes(n.then_), out)
			collectComponentNamesOf(derefViewNodes(n.else_), out)
		} else if n.kind == "list" {
			collectComponentNamesOf(derefViewNodes(n.body), out)
		}
	}
}

// derefViewNodes turns a []*ViewNode into a []ViewNode for the value-based
// walkers.
func derefViewNodes(nodes []*ViewNode) []ViewNode {
	var out []ViewNode
	for i := 0; i < len(nodes); i++ {
		if nodes[i] != nil {
			out = append(out, *nodes[i])
		}
	}
	return out
}

// orderNamedArgsJs rewrites named call arguments into declaration order for
// code generation. The typechecker has already validated names and gaps.
func orderNamedArgsJs(args []*Expr, names []string, params []string) []*Expr {
	if !hasNamedArgs(names) {
		return args
	}
	pos := map[string]int{}
	for i := 0; i < len(params); i++ {
		pos[params[i]] = i
	}
	out := make([]*Expr, len(args))
	for i := 0; i < len(args); i++ {
		if i >= len(names) || names[i] == "" {
			return args
		}
		idx, ok := pos[names[i]]
		if !ok || idx >= len(out) {
			return args
		}
		out[idx] = args[i]
	}
	last := len(out) - 1
	for last >= 0 && out[last] == nil {
		last--
	}
	return out[:last+1]
}

// typeGuardJs lowers `value is Type` to an inert runtime test. Primitive
// checks use typeof/Number.isInteger; enums test membership in the generated
// frozen object; arrays/maps use Array.isArray/object checks.
func typeGuardJs(e *Expr, name string) string {
	x := ExprToJs(e)
	switch name {
	case "string":
		return "(typeof " + x + " === \"string\")"
	case "int":
		return "(typeof " + x + " === \"number\" && Number.isInteger(" + x + "))"
	case "float":
		return "(typeof " + x + " === \"number\" && !Number.isInteger(" + x + "))"
	case "bool":
		return "(typeof " + x + " === \"boolean\")"
	case "null":
		return "(" + x + " === null)"
	case "fn":
		return "(typeof " + x + " === \"function\")"
	case "any", "unknown":
		return "true"
	}
	// Literal types compare against their value: strings arrive already quoted
	// by the parser, numbers and booleans are valid JavaScript literals.
	if strings.HasPrefix(name, "\"") {
		return "(" + x + " === " + name + ")"
	}
	if name == "true" || name == "false" {
		return "(" + x + " === " + name + ")"
	}
	if _, err := strconv.ParseFloat(name, 64); err == nil {
		return "(" + x + " === " + name + ")"
	}
	if strings.HasPrefix(name, "array[") {
		return "Array.isArray(" + x + ")"
	}
	if strings.HasPrefix(name, "map[") {
		return "(" + x + " !== null && typeof " + x + " === \"object\" && !Array.isArray(" + x + "))"
	}
	if members, ok := genEnums[name]; ok {
		var tests []string
		for i := 0; i < len(members); i++ {
			tests = append(tests, x+" === "+name+"."+members[i])
		}
		if len(tests) == 0 {
			return "false"
		}
		return "(" + strings.Join(tests, " || ") + ")"
	}
	// Structs and unknown user types: a non-null object is the best runtime
	// check available without class tags.
	return "(" + x + " !== null && typeof " + x + " === \"object\")"
}

func ExprToJs(e *Expr) string {
	if e == nil {
		return "null"
	}
	js := exprToJsRaw(e)
	if genEmitSourceMarkers {
		return srcMarker(e.span) + js
	}
	return js
}

func exprToJsRaw(e *Expr) string {
	if e == nil {
		return "null"
	}
	k := e.kind
	if k == "int" {
		return strconv.Itoa(e.i)
	}
	if k == "float" {
		return fmt.Sprintf("%v", e.f)
	}
	if k == "bool" {
		if e.b {
			return "true"
		}
		return "false"
	}
	if k == "null" {
		return "null"
	}
	if k == "str" {
		return templateParts(e.parts)
	}
	if k == "cond" {
		elseJs := "null"
		if e.r != nil {
			elseJs = ExprToJs(e.r)
		}
		return "(" + ExprToJs(e.cond) + " ? " + ExprToJs(e.l) + " : " + elseJs + ")"
	}
	if k == "ident" {
		return e.name
	}
	if k == "member" {
		return ExprToJs(e.obj) + "." + e.memb
	}
	if k == "call" {
		// Fire-and-forget std intrinsics (ws.send/ws.onMessage/ws.close/...):
		// emitted as __SUI_STD.<key>(...), no tuple binding needed.
		if e.target != nil && e.target.kind == "member" {
			if f, ok := StdLookup(memberFullName(e.target)); ok && f != nil && f.ret0 == "null" {
				return "__SUI_STD." + f.jsKey + "(" + exprListJs(e.args) + ")"
			}
		}
		// Built-in functions (len, append, upper, ...) translate into native
		// JS expressions; a user-declared function wins over a same-named
		// builtin. Std intrinsics and user async calls go through the normal
		// expression form.
		builtinName := ""
		if e.target != nil && e.target.kind == "ident" && IsBuiltin(e.target.name) && !isUserFunc(e.target.name) {
			builtinName = e.target.name
		} else if e.target != nil && e.target.kind == "member" && IsBuiltin(memberFullName(e.target)) {
			builtinName = memberFullName(e.target)
		}
		if builtinName != "" {
			var argJs []string
			for i := 0; i < len(e.args); i++ {
				argJs = append(argJs, ExprToJs(e.args[i]))
			}
			if builtinName == "toast.show" {
				// kind and duration are optional; the runtime treats undefined
				// as the default info/4000ms values.
				for len(argJs) < 3 {
					argJs = append(argJs, "undefined")
				}
				return "__sui_toast_show(" + strings.Join(argJs, ", ") + ")"
			}
			if js := builtinToJs(builtinName, argJs); js != "" {
				if IsAsyncBuiltin(builtinName) {
					return "(await " + js + ")"
				}
				return js
			}
		}
		// User function calls append the missing trailing default values.
		if e.target != nil && e.target.kind == "ident" && isUserFunc(e.target.name) {
			args := orderNamedArgsJs(e.args, e.argNames, genFnParams[e.target.name])
			var argJs []string
			for i := 0; i < len(args); i++ {
				argJs = append(argJs, ExprToJs(args[i]))
			}
			if defs := genFnDefaults[e.target.name]; len(defs) > len(argJs) {
				for i := len(argJs); i < len(defs) && defs[i] != ""; i++ {
					argJs = append(argJs, defs[i])
				}
			}
			return ExprToJs(e.target) + "(" + strings.Join(argJs, ", ") + ")"
		}
		return ExprToJs(e.target) + "(" + exprListJs(e.args) + ")"
	}
	if k == "array" {
		var parts []string
		for i := 0; i < len(e.elements); i++ {
			prefix := ""
			if i < len(e.spreads) && e.spreads[i] {
				prefix = "..."
			}
			parts = append(parts, prefix+ExprToJs(e.elements[i]))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	if k == "map" {
		var parts []string
		for i := 0; i < len(e.keys); i++ {
			if i < len(e.spreads) && e.spreads[i] {
				parts = append(parts, "..."+ExprToJs(e.vals[i]))
				continue
			}
			parts = append(parts, jsString(e.keys[i])+": "+ExprToJs(e.vals[i]))
		}
		return "({" + strings.Join(parts, ", ") + "})"
	}
	if k == "structlit" {
		var parts []string
		for i := 0; i < len(e.keys); i++ {
			parts = append(parts, e.keys[i]+": "+ExprToJs(e.vals[i]))
		}
		return "({" + strings.Join(parts, ", ") + "})"
	}
	if k == "index" {
		return ExprToJs(e.obj) + "[" + ExprToJs(e.idx) + "]"
	}
	if k == "unary" {
		if e.op == "neg" {
			return "(-" + ExprToJs(e.operand) + ")"
		}
		return "(!" + ExprToJs(e.operand) + ")"
	}
	if k == "is" {
		return typeGuardJs(e.l, e.name)
	}
	if k == "binary" {
		if e.op == "div" {
			// SUI integer division truncates (Go-like); floats use `fdiv`.
			return "(Math.trunc(" + ExprToJs(e.l) + " / " + ExprToJs(e.r) + "))"
		}
		if e.op == "fdiv" {
			return "(" + ExprToJs(e.l) + " / " + ExprToJs(e.r) + ")"
		}
		return "(" + ExprToJs(e.l) + " " + jsBinOp(e.op) + " " + ExprToJs(e.r) + ")"
	}
	return "null"
}

func StatementsToJs(stmts []*Stmt) string {
	var ctx JsCtx
	return stmtsToJs(stmts, "", &ctx)
}

func EscapeJsStr(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' {
			sb.WriteString("\\\\")
		} else if c == '"' {
			sb.WriteString("\\\"")
		} else if c == '\n' {
			sb.WriteString("\\n")
		} else if c == '\t' {
			sb.WriteString("\\t")
		} else if c == '\r' {
			sb.WriteString("\\r")
		} else if c == 0 {
			sb.WriteString("\\u0000")
		} else {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func EscapeJsTemplate(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' {
			sb.WriteString("\\\\")
		} else if c == '`' {
			sb.WriteString("\\`")
		} else if c == '$' {
			sb.WriteString("\\$")
		} else if c == '\n' {
			sb.WriteString("\\n")
		} else if c == '\t' {
			sb.WriteString("\\t")
		} else if c == '\r' {
			sb.WriteString("\\r")
		} else if c == 0 {
			sb.WriteString("\\u0000")
		} else {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func wsImported(p *Prog) bool {
	for i := 0; i < len(p.imports); i++ {
		im := p.imports[i]
		if im.kind != "std" {
			continue
		}
		for j := 0; j < len(im.names); j++ {
			if im.names[j] == "ws" {
				return true
			}
		}
	}
	return false
}

func capFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	c := s[0]
	if c >= 'a' && c <= 'z' {
		return string([]byte{c - 32}) + s[1:]
	}
	return s
}

func setterName(name string) string {
	return "set" + capFirst(name)
}

func isNameIn(names []string, name string) bool {
	for i := 0; i < len(names); i++ {
		if names[i] == name {
			return true
		}
	}
	return false
}

func ctxIsState(ctx *JsCtx, name string) bool {
	return isNameIn(ctx.states, name)
}

func ctxTruncateScope(ctx *JsCtx, n int) {
	var sc []string
	for i := 0; i < n && i < len(ctx.scope); i++ {
		sc = append(sc, ctx.scope[i])
	}
	ctx.scope = sc
}

func jsString(s string) string {
	return "\"" + EscapeJsStr(s) + "\""
}

func sortStringList(a []string) []string {
	var out []string
	for i := 0; i < len(a); i++ {
		out = append(out, a[i])
	}
	n := len(out)
	for i := 1; i < n; i++ {
		j := i
		for j > 0 && out[j] < out[j-1] {
			tmp := out[j]
			out[j] = out[j-1]
			out[j-1] = tmp
			j = j - 1
		}
	}
	return out
}

func jsBinOp(op string) string {
	if op == "add" {
		return "+"
	}
	if op == "sub" {
		return "-"
	}
	if op == "mul" {
		return "*"
	}
	if op == "div" {
		return "/"
	}
	if op == "mod" {
		return "%"
	}
	if op == "==" {
		return "==="
	}
	if op == "!=" {
		return "!=="
	}
	if op == "lt" {
		return "<"
	}
	if op == "lte" {
		return "<="
	}
	if op == "gt" {
		return ">"
	}
	if op == "gte" {
		return ">="
	}
	if op == "and" {
		return "&&"
	}
	if op == "or" {
		return "||"
	}
	return op
}

// lvalueRootNameJs returns the base identifier of a field/index lvalue.
func lvalueRootNameJs(e *Expr) string {
	if e != nil && (e.kind == "member" || e.kind == "index") {
		return lvalueRootNameJs(e.obj)
	}
	if e != nil && e.kind == "ident" {
		return e.name
	}
	return ""
}

func exprListJs(es []*Expr) string {
	var parts []string
	for i := 0; i < len(es); i++ {
		parts = append(parts, ExprToJs(es[i]))
	}
	return strings.Join(parts, ", ")
}

func templateParts(parts []StrPart) string {
	var sb strings.Builder
	sb.WriteString("`")
	partsContentInto(&sb, parts)
	sb.WriteString("`")
	return sb.String()
}

func partsContentInto(sb *strings.Builder, parts []StrPart) {
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if p.kind == "expr" {
			sb.WriteString("${")
			if p.expr == nil {
				sb.WriteString("null")
			} else {
				sb.WriteString(ExprToJs(p.expr))
			}
			sb.WriteString("}")
		} else {
			sb.WriteString(EscapeJsTemplate(p.text))
		}
	}
}

func partsContent(parts []StrPart) string {
	var sb strings.Builder
	partsContentInto(&sb, parts)
	return sb.String()
}

// htmlTextParts renders element text with HTML escaping. Static text is
// escaped and every interpolated expression goes through __sui_esc, so
// untrusted values can never be parsed as markup.
func htmlTextParts(parts []StrPart) string {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		p := &parts[i]
		if p.kind == "expr" {
			sb.WriteString("${__sui_esc(")
			if p.expr == nil {
				sb.WriteString("null")
			} else {
				sb.WriteString(ExprToJs(p.expr))
			}
			sb.WriteString(")}")
		} else {
			sb.WriteString(htmlEscape(EscapeJsTemplate(p.text)))
		}
	}
	return sb.String()
}

func classExprJs(expr string) string {
	if genClassFilter {
		return "__sui_class(" + expr + ")"
	}
	return "__sui_esc(" + expr + ")"
}

// classPartsJs renders a class attribute: static tokens are kept (escaped),
// interpolated values pass through __sui_class (allow-list filter, default) or
// __sui_esc when the project explicitly disabled class filtering.
func classPartsJs(parts []StrPart) string {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		p := &parts[i]
		if p.kind == "expr" {
			e := "null"
			if p.expr != nil {
				e = ExprToJs(p.expr)
			}
			sb.WriteString("${" + classExprJs(e) + "}")
		} else {
			sb.WriteString(htmlEscape(EscapeJsTemplate(p.text)))
		}
	}
	return sb.String()
}

// htmlAttrValue renders an attribute value for a double-quoted attribute.
// Expressions are escaped; href/src values also pass through __sui_url, which
// rejects active URL schemes (javascript:, data:, ...).
func htmlAttrValue(a *Attr, url bool) string {
	if a.hasExpr {
		expr := "null"
		if a.expr != nil {
			expr = ExprToJs(a.expr)
		}
		if url {
			return "${__sui_esc(__sui_url(" + expr + "))}"
		}
		return "${__sui_esc(" + expr + ")}"
	}
	if url {
		// `partsContent` keeps expressions raw inside the template literal;
		// the whole URL is sanitized and then escaped once.
		return "${__sui_esc(__sui_url(`" + partsContent(a.text) + "`))}"
	}
	return htmlTextParts(a.text)
}

// htmlAttrValueImage is htmlAttrValue for an <img> src. It permits the
// data:image/... URIs produced by the designer asset manager while delegating
// every other value to the normal URL guard.
func htmlAttrValueImage(a *Attr) string {
	if a.hasExpr {
		expr := "null"
		if a.expr != nil {
			expr = ExprToJs(a.expr)
		}
		return "${__sui_esc(__sui_img_url(" + expr + "))}"
	}
	return "${__sui_esc(__sui_img_url(`" + partsContent(a.text) + "`))}"
}

func memberFullName(e *Expr) string {
	if e == nil {
		return "."
	}
	if e.kind == "ident" {
		return e.name
	}
	if e.kind == "member" && e.obj != nil {
		return memberFullName(e.obj) + "." + e.memb
	}
	return e.memb
}

func tupleRhsJs(rhs *Expr) string {
	if rhs != nil && rhs.kind == "call" && rhs.target != nil && rhs.target.kind == "member" {
		f, ok := StdLookup(memberFullName(rhs.target))
		if ok && f != nil {
			return "__SUI_STD." + f.jsKey + "(" + exprListJs(rhs.args) + ")"
		}
	}
	return ExprToJs(rhs)
}

// tupleNeedAwait reports whether a tuple-RHS must be awaited: std intrinsics
// (async by contract) and user async functions. A sync user multi-return fn
// returns its array directly, so `const r = f(...)` would be valid JS even
// inside a non-async function.
func tupleNeedAwait(rhs *Expr) bool {
	if rhs == nil || rhs.kind != "call" {
		return false
	}
	t := rhs.target
	if t == nil {
		return false
	}
	if t.kind == "member" {
		if t.obj != nil {
			if f, ok := StdLookup(t.obj.name + "." + t.memb); ok && f != nil {
				return f.async
			}
		}
		return true
	}
	return t.kind == "ident" && genAsyncFuncs[t.name]
}

func stmtsToJs(stmts []*Stmt, ind string, ctx *JsCtx) string {
	var out []string
	for i := 0; i < len(stmts); i++ {
		out = append(out, srcMarker(stmts[i].span)+stmtToJs(stmts[i], ind, ctx))
	}
	return strings.Join(out, "\n")
}

func stmtToJs(st *Stmt, ind string, ctx *JsCtx) string {
	if st == nil {
		return ""
	}
	k := st.kind
	if k == "let" {
		if len(st.pattern) > 0 {
			for i := 0; i < len(st.pattern); i++ {
				ctx.scope = append(ctx.scope, st.pattern[i])
			}
			open, close := "{", "}"
			if st.patternKind == "array" {
				open, close = "[", "]"
			}
			return ind + "let " + open + strings.Join(st.pattern, ", ") + close + " = " + ExprToJs(st.init) + ";"
		}
		ctx.scope = append(ctx.scope, st.name)
		return ind + "let " + st.name + " = " + ExprToJs(st.init) + ";"
	}
	if k == "switch" {
		var b strings.Builder
		b.WriteString(ind + "switch (" + ExprToJs(st.init) + ") {\n")
		for i := 0; i < len(st.cases); i++ {
			cse := st.cases[i]
			b.WriteString(ind + "  case " + ExprToJs(cse.expr) + ":\n")
			for j := 0; j < len(cse.body); j++ {
				b.WriteString(stmtToJs(cse.body[j], ind+"    ", ctx) + "\n")
			}
			b.WriteString(ind + "    break;\n")
		}
		if st.hasDefault {
			b.WriteString(ind + "  default:\n")
			for j := 0; j < len(st.default_); j++ {
				b.WriteString(stmtToJs(st.default_[j], ind+"    ", ctx) + "\n")
			}
		}
		b.WriteString(ind + "}")
		return b.String()
	}
	if k == "assign" {
		if st.lvalue != nil {
			// Struct field / array / map element mutation: JS assigned in
			// place; when the root is a state, re-sync via its setter so the
			// reactive re-render fires.
			lhs := ExprToJs(st.lvalue)
			if st.rhs != nil {
				out := ind + lhs + " = " + ExprToJs(st.rhs) + ";"
				if root := lvalueRootNameJs(st.lvalue); root != "" && ctxIsState(ctx, root) {
					out = out + "\n" + ind + setterName(root) + "(" + root + ");"
				}
				return out
			}
			return ind + lhs + ";"
		}
		if len(st.targets) >= 2 {
			var lines []string
			tmpName := "__sui_r"
			if ctx.tuples > 0 {
				tmpName = tmpName + strconv.Itoa(ctx.tuples+1)
			}
			prefix := "await "
			if !tupleNeedAwait(st.rhs) {
				prefix = ""
			}
			lines = append(lines, ind+"const "+tmpName+" = "+prefix+tupleRhsJs(st.rhs)+";")
			ctx.tuples = ctx.tuples + 1
			for i := 0; i < len(st.targets); i++ {
				lines = append(lines, ind+"const "+st.targets[i]+" = "+tmpName+"["+strconv.Itoa(i)+"];")
			}
			return strings.Join(lines, "\n")
		}
		if len(st.targets) == 1 {
			name := st.targets[0]
			if genCompBody && genCompInst[name] {
				// Per-instance state: write through to inst and refresh.
				return ind + name + " = " + ExprToJs(st.rhs) + ";\n" +
					ind + "inst." + name + " = " + name + ";\n" +
					ind + "__sui_refresh();"
			}
			if ctxIsState(ctx, name) {
				return ind + setterName(name) + "(" + ExprToJs(st.rhs) + ");"
			}
			return ind + name + " = " + ExprToJs(st.rhs) + ";"
		}
		return ""
	}
	if k == "expr" {
		return ind + ExprToJs(st.init) + ";"
	}
	if k == "if" {
		var sb strings.Builder
		sb.WriteString(ind)
		writeIfTree(&sb, st, ind, ctx)
		return sb.String()
	}
	if k == "loop" {
		item := st.item
		iterable := st.iterable
		saved := len(ctx.scope)
		if st.item2 != "" {
			// Map iteration binds key and value.
			tmp := "__sui_" + item
			head := ind + "for (const " + tmp + " of Object.entries(" + ExprToJs(iterable) + ")) {"
			ctx.scope = append(ctx.scope, item)
			ctx.scope = append(ctx.scope, st.item2)
			body := stmtsToJs(st.body, ind+"  ", ctx)
			ctxTruncateScope(ctx, saved)
			var mb strings.Builder
			mb.WriteString(head + "\n")
			mb.WriteString(ind + "  const " + item + " = " + tmp + "[0];\n")
			mb.WriteString(ind + "  const " + st.item2 + " = " + tmp + "[1];\n")
			mb.WriteString(body)
			mb.WriteString(ind + "}")
			return mb.String()
		}
		var head string
		if iterable != nil && iterable.kind == "range" {
			head = ind + "for (let " + item + " = " + ExprToJs(iterable.lo) + "; " + item + " < " + ExprToJs(iterable.hi) + "; " + item + "++) {"
		} else {
			head = ind + "for (const " + item + " of (" + ExprToJs(iterable) + ")) {"
		}
		ctx.scope = append(ctx.scope, item)
		body := stmtsToJs(st.body, ind+"  ", ctx)
		ctxTruncateScope(ctx, saved)
		return head + "\n" + body + "\n" + ind + "}"
	}
	if k == "while" {
		saved := len(ctx.scope)
		body := stmtsToJs(st.body, ind+"  ", ctx)
		ctxTruncateScope(ctx, saved)
		return ind + "while (" + ExprToJs(st.cond) + ") {\n" + body + "\n" + ind + "}"
	}
	if k == "return" {
		if len(st.values) > 0 {
			// Multi-return: the JS function returns a tuple array.
			var parts []string
			for i := 0; i < len(st.values); i++ {
				parts = append(parts, ExprToJs(st.values[i]))
			}
			return ind + "return [" + strings.Join(parts, ", ") + "];"
		}
		if st.value == nil {
			return ind + "return;"
		}
		return ind + "return " + ExprToJs(st.value) + ";"
	}
	if k == "break" {
		return ind + "break;"
	}
	if k == "continue" {
		return ind + "continue;"
	}
	return ""
}

func writeIfTree(sb *strings.Builder, st *Stmt, ind string, ctx *JsCtx) {
	sb.WriteString("if (" + ExprToJs(st.cond) + ") {\n")
	saved := len(ctx.scope)
	sb.WriteString(stmtsToJs(st.then_, ind+"  ", ctx))
	ctxTruncateScope(ctx, saved)
	sb.WriteString("\n" + ind)
	if len(st.else_) == 1 && st.else_[0].kind == "if" {
		sb.WriteString("} else ")
		writeIfTree(sb, st.else_[0], ind, ctx)
	} else if len(st.else_) > 0 {
		sb.WriteString("} else {\n")
		saved := len(ctx.scope)
		sb.WriteString(stmtsToJs(st.else_, ind+"  ", ctx))
		ctxTruncateScope(ctx, saved)
		sb.WriteString("\n" + ind + "}")
	} else {
		sb.WriteString("}")
	}
}

func funcJs(f *FuncDecl, ctx *JsCtx) string {
	var paramNames []string
	for i := 0; i < len(f.params); i++ {
		name := f.params[i].name
		if f.params[i].rest {
			name = "..." + name
		}
		paramNames = append(paramNames, name)
	}
	prefix := "function "
	if f.async {
		prefix = "async function "
	}
	saved := len(ctx.scope)
	savedTuples := ctx.tuples
	ctx.tuples = 0
	for i := 0; i < len(paramNames); i++ {
		ctx.scope = append(ctx.scope, paramNames[i])
	}
	body := stmtsToJs(f.body, "  ", ctx)
	ctxTruncateScope(ctx, saved)
	ctx.tuples = savedTuples
	return srcMarker(f.span) + prefix + f.name + "(" + strings.Join(paramNames, ", ") + ") {\n" + body + "\n}"
}

func stdObjJs(p *Prog, target string) string {
	var mods []string
	for i := 0; i < len(p.imports); i++ {
		im := p.imports[i]
		if im.kind != "std" {
			continue
		}
		for j := 0; j < len(im.names); j++ {
			if !isNameIn(mods, im.names[j]) {
				mods = append(mods, im.names[j])
			}
		}
	}
	var entries []string
	for i := 0; i < len(StdFuncs); i++ {
		f := StdFuncs[i]
		if isNameIn(mods, f.mod) || IsCoreModule(f.mod) {
			entries = append(entries, "  "+f.jsKey+": "+stdBodyJs(f.jsKey, target))
		}
	}
	if len(entries) == 0 {
		return ""
	}
	prefix := ""
	if xmlModuleImported(p) {
		prefix += xmlRuntimeSource + "\n\n"
	}
	if jsonModuleImported(p) {
		prefix += jsonRuntimeSource + "\n\n"
	}
	return prefix + "const __SUI_STD = {\n" + strings.Join(entries, ",\n") + "\n};"
}

var chartAccessorFields = map[string]string{
	"label": "label", "text": "tag", "hint": "hint", "description": "description",
	"color": "color", "tagcolor": "tagBoxColor", "textcolor": "tagTextColor",
	"shape": "shape", "routing": "routing", "bordercolor": "borderColor",
	"borderstyle": "borderStyle", "borderwidth": "borderWidth", "cornerradius": "cornerRadius",
	"kind": "kind", "from": "from", "to": "to",
	"width": "width", "height": "height", "x": "x", "y": "y", "rotation": "rotation",
	"dashstyle": "dashStyle", "startmarker": "startMarker", "endmarker": "endMarker",
	"tagkind": "tagKind", "tagbox": "tagBox", "tagalign": "tagAlign", "tagleader": "tagLeader",
	"waypoints": "waypoints", "textplacement": "textPlacement", "tagside": "tagSide",
	"borderdashlength": "borderDashLength", "borderdashgap": "borderDashGap", "snap": "snap",
}

func stdBodyJs(key string, target string) string {
	if strings.HasPrefix(key, "xml_") {
		return xmlStdBodyJs(key, target)
	}
	if key == "http_get" {
		return "async (url) => { try { const r = await fetch(url); return [await r.text(), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_getstatus" {
		return "async (url) => { try { const r = await fetch(url); return [r.status, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_post" {
		return "async (url, data) => { try { const r = await fetch(url, {method: \"POST\", headers: {\"Content-Type\": \"application/json\"}, body: data}); return [await r.text(), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_postfiles" {
		return "async (url) => { const fd = __sui_files_form(); if (!fd) return [null, 'no files selected']; try { const abs = new URL(url, location.origin).href; return __sui_fetch_guarded(abs, abs, {method: 'POST', body: fd}); } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_postfilessafe" {
		return "async (url, allow) => { const fd = __sui_files_form(); if (!fd) return [null, 'no files selected']; return __sui_fetch_guarded(url, allow, {method: 'POST', body: fd}); }"
	}
	if key == "http_getsafe" {
		return "async (url, allow) => __sui_fetch_guarded(url, allow, {})"
	}
	if key == "http_getstatussafe" {
		return "async (url, allow) => __sui_fetch_status_guarded(url, allow)"
	}
	if key == "http_postsafe" {
		return "async (url, data, allow) => __sui_fetch_guarded(url, allow, {method: \"POST\", headers: {\"Content-Type\": \"application/json\"}, body: data})"
	}
	if key == "http_getwithheaders" {
		return "async (url, headers) => { try { const abs = new URL(url, location.origin).href; return __sui_fetch_guarded(abs, abs, {method: 'GET', headers: __sui_clean_headers(headers)}); } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_getwithheaderssafe" {
		return "async (url, headers, allow) => __sui_fetch_guarded(url, allow, {method: 'GET', headers: __sui_clean_headers(headers)})"
	}
	if key == "http_postwithheaders" {
		return "async (url, data, headers) => { try { const abs = new URL(url, location.origin).href; const h = Object.assign({\"Content-Type\": \"application/json\"}, __sui_clean_headers(headers)); return __sui_fetch_guarded(abs, abs, {method: 'POST', headers: h, body: data}); } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_postwithheaderssafe" {
		return "async (url, data, headers, allow) => { const h = Object.assign({\"Content-Type\": \"application/json\"}, __sui_clean_headers(headers)); return __sui_fetch_guarded(url, allow, {method: 'POST', headers: h, body: data}); }"
	}
	if key == "http_request" {
		return "async (method, url, headers, data, timeoutMs) => { try { const abs = new URL(url, location.origin).href; const m = String(method || 'GET').toUpperCase(); const init = {method: m, headers: __sui_clean_headers(headers)}; if (m !== 'GET' && m !== 'HEAD' && data !== undefined && data !== null && data !== '') init.body = data; return __sui_fetch_guarded(abs, abs, init, timeoutMs); } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_requestsafe" {
		return "async (method, url, headers, data, timeoutMs, allow) => { const m = String(method || 'GET').toUpperCase(); const init = {method: m, headers: __sui_clean_headers(headers)}; if (m !== 'GET' && m !== 'HEAD' && data !== undefined && data !== null && data !== '') init.body = data; return __sui_fetch_guarded(url, allow, init, timeoutMs); }"
	}
	if key == "http_upload" {
		return "async (url, fieldName, fileName, content) => { try { const abs = new URL(url, location.origin).href; const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(abs, abs, {method: 'POST', body: fd}); } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_uploadsafe" {
		return "async (url, fieldName, fileName, content, allow) => { const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(url, allow, {method: 'POST', body: fd}); }"
	}
	if key == "http_uploadwithheaders" {
		return "async (url, fieldName, fileName, content, headers) => { try { const abs = new URL(url, location.origin).href; const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(abs, abs, {method: 'POST', headers: __sui_clean_headers(headers), body: fd}); } catch (e) { return [null, String(e)]; } }"
	}
	if key == "http_uploadwithheaderssafe" {
		return "async (url, fieldName, fileName, content, headers, allow) => { const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(url, allow, {method: 'POST', headers: __sui_clean_headers(headers), body: fd}); }"
	}
	if key == "json_validate" {
		return "(value, schema) => __sui_json_validate(value, schema)"
	}
	if key == "json_parse" {
		return "async (s) => __SUI_JSON.parse(s)"
	}
	if key == "json_parse_bytes" {
		return "(bytes, encoding) => __SUI_JSON.parseBytes(bytes, encoding)"
	}
	if key == "json_stringify" {
		return "async (v, options) => __SUI_JSON.stringify(v, options)"
	}
	if key == "json_to_bytes" {
		return "(value, encoding, options) => __SUI_JSON.toBytes(value, encoding, options)"
	}
	if key == "json_validate_all" {
		return "(value, schema) => __SUI_JSON.validateAll(value, schema)"
	}
	if key == "json_pointer" {
		return "(value, path) => __SUI_JSON.pointer(value, path)"
	}
	if key == "json_set_pointer" {
		return "(value, path, next) => __SUI_JSON.setPointer(value, path, next)"
	}
	if key == "json_remove_pointer" {
		return "(value, path) => __SUI_JSON.removePointer(value, path)"
	}
	if key == "json_clone" {
		return "(value) => __SUI_JSON.clone(value)"
	}
	if key == "json_equal" {
		return "(a, b) => __SUI_JSON.equal(a, b)"
	}
	if key == "json_diff" {
		return "(a, b, options) => __SUI_JSON.diff(a, b, options)"
	}
	if key == "json_patch" {
		return "(value, changes) => __SUI_JSON.patch(value, changes)"
	}
	if key == "json_merge" {
		return "(value, patch) => __SUI_JSON.merge(value, patch)"
	}
	if key == "json_merge3" {
		return "(base, ours, theirs, options) => __SUI_JSON.merge3(base, ours, theirs, options)"
	}
	if key == "json_load_file" || key == "json_load_bytes" {
		if target == "desktop" {
			return "async (path, encoding) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/file/readbytes\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? __SUI_JSON.parseBytes(__sui_b64_to_bytes(j.data), encoding) : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		if key == "json_load_file" {
			return "async (path, encoding) => [null, \"json.loadFile is not available on the web target; use fs.read or fetch and json.parse\"]"
		}
		return "async (path, encoding) => [null, \"json.loadBytes is not available on the web target\"]"
	}
	if key == "json_save_file" {
		if target == "desktop" {
			return "async (path, value, options) => { const res = __SUI_JSON.stringify(value, options || {}); if (res[1] !== null) return res; const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); fd.set(\"data\", res[0]); try { const r = await fetch(\"/__bridge__/file/write\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path, value, options) => [null, \"json.saveFile is not available on the web target\"]"
	}
	if key == "json_save_bytes" {
		if target == "desktop" {
			return "async (path, value, encoding, options) => { const res = __SUI_JSON.stringify(value, options || {}); if (res[1] !== null) return res; let bytes; try { bytes = __SUI_JSON.encodeBytes(res[0], encoding || \"utf-8\"); } catch (e) { return [null, String(e)]; } const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); fd.set(\"data\", __sui_bytes_to_b64(bytes)); try { const r = await fetch(\"/__bridge__/file/writebytes\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path, value, encoding, options) => [null, \"json.saveBytes is not available on the web target\"]"
	}
	if key == "fs_read" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/file/read\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [j.data, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path) => { try { const r = await fetch(\"/api/fs/read?path=\" + encodeURIComponent(path)); if (!r.ok) return [null, \"fs.read: HTTP \" + r.status]; return [await r.text(), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "fs_write" {
		if target == "desktop" {
			return "async (path, data) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); fd.set(\"data\", data); try { const r = await fetch(\"/__bridge__/file/write\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path, data) => { try { const r = await fetch(\"/api/fs/write\", {method: \"POST\", headers: {\"Content-Type\": \"application/json\"}, body: JSON.stringify({path, data})}); if (!r.ok) return [null, \"fs.write: HTTP \" + r.status]; return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "fs_exists" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/fs/stat\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [j.data !== \"missing\", null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path) => [false, \"fs.exists is not available on the web target\"]"
	}
	if key == "fs_stat" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/fs/stat\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [j.data, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path) => [\"\", \"fs.stat is not available on the web target\"]"
	}
	if key == "fs_mkdir" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/fs/mkdir\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path) => [null, \"fs.mkdir is not available on the web target\"]"
	}
	if key == "fs_remove" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/fs/remove\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path) => [null, \"fs.remove is not available on the web target\"]"
	}
	if key == "fs_list" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/fs/list\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [j.data.split(\"\\n\").filter((x) => x !== \"\"), null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (path) => [\"\", \"fs.list is not available on the web target\"]"
	}
	if key == "fs_readbytes" {
		if target == "desktop" {
			return "async function (path) { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/file/readbytes\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [__sui_b64_to_bytes(j.data), null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async function (path) { return [null, \"fs.readBytes is not available on the web target\"]; }"
	}
	if key == "fs_writebytes" {
		if target == "desktop" {
			return "async function (path, bytes) { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); fd.set(\"data\", __sui_bytes_to_b64(bytes)); try { const r = await fetch(\"/__bridge__/file/writebytes\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async function (path, bytes) { return [null, \"fs.writeBytes is not available on the web target\"]; }"
	}
	if key == "fs_copy" {
		if target == "desktop" {
			return "async (src, dst) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"src\", src); fd.set(\"dst\", dst); try { const r = await fetch(\"/__bridge__/fs/copy\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (src, dst) => [null, \"fs.copy is not available on the web target\"]"
	}
	if key == "fs_move" {
		if target == "desktop" {
			return "async (src, dst) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"src\", src); fd.set(\"dst\", dst); try { const r = await fetch(\"/__bridge__/fs/move\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [null, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (src, dst) => [null, \"fs.move is not available on the web target\"]"
	}
	if key == "fs_isdir" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/fs/stat\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [j.data === \"dir\", null] : [false, j.message]; } catch (e) { return [false, String(e)]; } }"
		}
		return "async (path) => [false, \"fs.isDir is not available on the web target\"]"
	}
	if key == "fs_size" {
		if target == "desktop" {
			return "async (path) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"path\", path); try { const r = await fetch(\"/__bridge__/fs/size\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [Number(j.data) || 0, null] : [0, j.message]; } catch (e) { return [0, String(e)]; } }"
		}
		return "async (path) => [0, \"fs.size is not available on the web target\"]"
	}
	if key == "exec_run" {
		if target == "desktop" {
			return "async (cmd, args) => { const fd = new URLSearchParams(); fd.set(\"token\", __SUI_TOKEN); fd.set(\"cmd\", cmd); for (const a of args) fd.append(\"arg\", a); try { const r = await fetch(\"/__bridge__/exec\", {method: \"POST\", body: fd}); const j = await r.json(); return j.ok ? [j.data, null] : [null, j.message]; } catch (e) { return [null, String(e)]; } }"
		}
		return "async (cmd, args) => [null, \"exec.run is disabled on the web target\"]"
	}
	if key == "ws_connect" {
		return "async (url) => { try { return [new __SUI_WS(url), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_connectsafe" {
		return "async (url, allow) => { const g = __sui_url_guard(url, allow, true); if (g[1]) return [null, g[1]]; try { return [new __SUI_WS(g[0]), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_connectwithoptions" {
		return "async (url, opts) => { try { return [new __SUI_WS(url, opts), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_connectwithoptionssafe" {
		return "async (url, opts, allow) => { const g = __sui_url_guard(url, allow, true); if (g[1]) return [null, g[1]]; try { return [new __SUI_WS(g[0], opts), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_send" {
		return "async (conn, data) => { try { conn.send(data); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_onmessage" {
		return "async (conn, fn) => { try { conn.onMessage((data) => fn(data)); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_onopen" {
		return "async (conn, fn) => { try { conn.onOpen(() => fn()); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_onclose" {
		return "async (conn, fn) => { try { conn.onClose(() => fn()); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_onerror" {
		return "async (conn, fn) => { try { conn.onError((err) => fn(err)); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_status" {
		return "(conn) => { try { return [conn ? conn.status() : 'closed', null]; } catch (e) { return ['closed', String(e)]; } }"
	}
	if key == "ws_reconnect" {
		return "async (conn) => { try { if (conn) conn.reconnect(); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_close" {
		return "async (conn) => { try { conn.close(); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "dialog_pickfile" {
		if target != "desktop" {
			return "async () => [null, 'dialog.pickFile is available only in desktop builds']"
		}
		return "async () => { const r = await __sui_dialog_pick('/__bridge__/pick-file', false); if (r[1]) return r; return [r[0] && r[0].length ? r[0][0] : '', null]; }"
	}
	if key == "dialog_pickfiles" {
		if target != "desktop" {
			return "async () => [null, 'dialog.pickFiles is available only in desktop builds']"
		}
		return "async () => __sui_dialog_pick('/__bridge__/pick-file', true)"
	}
	if key == "dialog_pickdirectory" {
		if target != "desktop" {
			return "async () => [null, 'dialog.pickDirectory is available only in desktop builds']"
		}
		return "async () => { const r = await __sui_dialog_pick('/__bridge__/pick-directory', false); if (r[1]) return r; return [r[0] && r[0].length ? r[0][0] : '', null]; }"
	}
	// Core table helpers: synchronous, no I/O. They operate on the rendered
	// table carrying the matching name attribute.
	if key == "richtext_get" {
		return "(name) => __SUI_RICH.get(name)"
	}
	if key == "richtext_set" {
		return "(name, value) => __SUI_RICH.set(name, value)"
	}
	if key == "richtext_command" {
		return "(name, cmd, value) => __SUI_RICH.command(name, cmd, value)"
	}
	if key == "richtext_format" {
		return "(name, fmt, value) => __SUI_RICH.format(name, fmt, value)"
	}
	if key == "richtext_inserttext" {
		return "(name, text) => __SUI_RICH.insertText(name, text)"
	}
	if key == "richtext_insertimage" {
		return "(name, src, alt) => __SUI_RICH.insertImage(name, src, alt)"
	}
	if key == "richtext_insertlink" {
		return "(name, href, text) => __SUI_RICH.insertLink(name, href, text)"
	}
	if key == "richtext_inserttable" {
		return "(name, rows, cols) => __SUI_RICH.insertTable(name, rows, cols)"
	}
	if key == "richtext_undo" {
		return "(name) => __SUI_RICH.undo(name)"
	}
	if key == "richtext_redo" {
		return "(name) => __SUI_RICH.redo(name)"
	}
	if key == "richtext_focus" {
		return "(name) => __SUI_RICH.focus(name)"
	}
	if key == "richtext_export" {
		return "(name, fmt) => __SUI_RICH.exportFormat(name, fmt)"
	}
	if key == "richtext_exportbytes" {
		return "(name, fmt) => __SUI_RICH.exportBytes(name, fmt)"
	}
	if key == "richtext_openfile" {
		return "async (name) => __SUI_RICH.openFile(name)"
	}
	if key == "richtext_importbytes" {
		return "async (name, bytes, filename) => __SUI_RICH.importBytes(name, bytes, filename)"
	}
	if key == "richtext_saveas" {
		return "(name, fmt) => __SUI_RICH.saveAs(name, fmt)"
	}
	if key == "richtext_import" {
		return "(name, text, fmt) => __SUI_RICH.importFormat(name, text, fmt)"
	}
	if key == "richtext_register" {
		return "(name, plugin) => __SUI_RICH.registerPlugin(name, plugin)"
	}
	if key == "richtext_plugins" {
		return "() => __SUI_RICH.plugins()"
	}
	if key == "richtext_status" {
		return "(name) => __SUI_RICH.status(name)"
	}
	if key == "richtext_selection" {
		return "(name) => __SUI_RICH.selection(name)"
	}
	if key == "richtext_settoolbar" {
		return "(name, items) => __SUI_RICH.setToolbar(name, items)"
	}
	if key == "richtext_settemplate" {
		return "(name, template) => __SUI_RICH.setTemplate(name, template)"
	}
	if key == "richtext_settheme" {
		return "(name, theme) => __SUI_RICH.setTheme(name, theme)"
	}
	if key == "richtext_destroy" {
		return "(name) => __SUI_RICH.destroy(name)"
	}
	if key == "table_display" {
		return "(name, data) => __SUI_TBL.display(name, data)"
	}
	if key == "table_data" {
		return "(name) => __SUI_TBL.data(name)"
	}
	if key == "table_header" {
		return "(name) => __SUI_TBL.header(name)"
	}
	if key == "table_select" {
		return "(name, cols) => __SUI_TBL.select(name, cols)"
	}
	if key == "table_sort" {
		return "(name, column, dir) => __SUI_TBL.sort(name, column, dir)"
	}
	if key == "table_filter" {
		return "(name, column, value) => __SUI_TBL.filter(name, column, value)"
	}
	if key == "table_filtertext" {
		return "(name, text) => __SUI_TBL.filterText(name, text)"
	}
	if key == "table_page" {
		return "(name, page) => __SUI_TBL.page(name, page)"
	}
	if key == "table_setpagesize" {
		return "(name, size) => __SUI_TBL.setPageSize(name, size)"
	}
	if key == "table_selectedrows" {
		return "(name) => __SUI_TBL.selectedRows(name)"
	}
	if key == "table_selectedcells" {
		return "(name) => __SUI_TBL.selectedCells(name)"
	}
	if key == "table_selectedcolumns" {
		return "(name) => __SUI_TBL.selectedColumns(name)"
	}
	if key == "table_lastchange" {
		return "(name) => __SUI_TBL.lastChange(name)"
	}
	if key == "table_slice" {
		return "(name, from, count) => __SUI_TBL.slice(name, from, count)"
	}
	if key == "table_importtext" {
		return "(name, text, spec) => __SUI_TBL.importText(name, text, spec)"
	}
	if key == "table_exporttext" {
		return "(name, spec) => __SUI_TBL.exportText(name, spec)"
	}
	if key == "table_json" {
		return "(name) => __SUI_TBL.json(name)"
	}
	if key == "table_clear" {
		return "(name) => __SUI_TBL.clear(name)"
	}
	if key == "table_clearpersisted" {
		return "(name) => __SUI_TBL.clearPersisted(name)"
	}
	// Core chart helpers: update rendered state-chart tags and element styles.
	if key == "chart_settag" {
		return "(name, id, text, color, textColor) => __SUI_CHART.setTag(name, id, text, color, textColor)"
	}
	if key == "chart_update" {
		return "(name, id, patch) => __SUI_CHART.update(name, id, patch)"
	}
	if key == "chart_updatemany" {
		return "(name, updates) => __SUI_CHART.updateMany(name, updates)"
	}
	if key == "chart_get" {
		return "(name, id) => __SUI_CHART.get(name, id)"
	}
	if key == "chart_elements" {
		return "(name) => __SUI_CHART.elements(name)"
	}
	if key == "chart_find" {
		return "(name, query) => __SUI_CHART.find(name, query)"
	}
	if key == "chart_getattr" {
		return "(name, id, field) => __SUI_CHART.getAttr(name, id, field)"
	}
	if key == "chart_setattr" {
		return "(name, id, field, value) => __SUI_CHART.setAttr(name, id, field, value)"
	}
	if strings.HasPrefix(key, "chart_get") {
		if field, ok := chartAccessorFields[strings.TrimPrefix(key, "chart_get")]; ok {
			return fmt.Sprintf("(name, id) => __SUI_CHART.getAttr(name, id, %q)", field)
		}
	}
	if strings.HasPrefix(key, "chart_set") {
		if field, ok := chartAccessorFields[strings.TrimPrefix(key, "chart_set")]; ok {
			return fmt.Sprintf("(name, id, value) => __SUI_CHART.setAttr(name, id, %q, value)", field)
		}
	}
	if key == "chart_focusrow" {
		return "(name, id) => __SUI_CHART.focusRow(name, id)"
	}
	if key == "chart_peekevents" {
		return "(name) => __SUI_CHART.peekEvents(name)"
	}
	if key == "chart_pollevents" {
		return "(name) => __SUI_CHART.pollEvents(name)"
	}
	if key == "chart_nextevent" {
		return "(name) => __SUI_CHART.nextEvent(name)"
	}
	if key == "chart_clearevents" {
		return "(name) => __SUI_CHART.clearEvents(name)"
	}
	if key == "chart_clearpersisted" {
		return "(name) => __SUI_CHART.clearPersisted(name)"
	}
	if key == "chart_eventcount" {
		return "(name) => __SUI_CHART.eventCount(name)"
	}
	// Core tree helpers: text import/export plus browser file read/download.
	if key == "tree_importjson" {
		return "(name, text) => __SUI_TREE.importJson(name, text)"
	}
	if key == "tree_importxml" {
		return "(name, text) => __SUI_TREE.importXml(name, text)"
	}
	if key == "tree_importjsonfile" {
		return "async (name, path) => __SUI_TREE.importJsonFile(name, path)"
	}
	if key == "tree_importxmlfile" {
		return "async (name, path) => __SUI_TREE.importXmlFile(name, path)"
	}
	if key == "tree_exportjson" {
		return "(name) => __SUI_TREE.exportJson(name)"
	}
	if key == "tree_exportxml" {
		return "(name) => __SUI_TREE.exportXml(name)"
	}
	if key == "tree_downloadjson" {
		return "(name, filename) => __SUI_TREE.downloadJson(name, filename)"
	}
	if key == "tree_downloadxml" {
		return "(name, filename) => __SUI_TREE.downloadXml(name, filename)"
	}
	if key == "tree_clearpersisted" {
		return "(name) => __SUI_TREE.clearPersisted(name)"
	}
	if key == "tree_clear" {
		return "(name) => __SUI_TREE.clear(name)"
	}
	if key == "tree_expandall" {
		return "(name) => __SUI_TREE.expandAll(name)"
	}
	if key == "tree_collapseall" {
		return "(name) => __SUI_TREE.collapseAll(name)"
	}
	if key == "tree_expandto" {
		return "(name, level) => __SUI_TREE.expandTo(name, level)"
	}
	if key == "tree_setchildren" {
		return "(name, key, text) => __SUI_TREE.setChildren(name, key, text)"
	}
	if key == "tree_setchildrenerror" {
		return "(name, key, message) => __SUI_TREE.setChildrenError(name, key, message)"
	}
	if key == "tree_retrychildren" {
		return "(name, key) => __SUI_TREE.retryChildren(name, key)"
	}
	if key == "tree_lastmove" {
		return "(name) => __SUI_TREE.lastMove(name)"
	}
	if key == "tree_lastchange" {
		return "(name) => __SUI_TREE.lastChange(name)"
	}
	if key == "tree_undomove" {
		return "(name) => __SUI_TREE.undoMove(name)"
	}
	if key == "tree_setcolumns" {
		return "(name, json) => __SUI_TREE.setColumns(name, json)"
	}
	if key == "tree_getcolumns" {
		return "(name) => __SUI_TREE.getColumns(name)"
	}
	if key == "tree_movecolumn" {
		return "(name, from, to) => __SUI_TREE.moveColumn(name, from, to)"
	}
	if key == "tree_setcell" {
		return "(name, key, column, value) => __SUI_TREE.setCell(name, key, column, value)"
	}
	if key == "tree_selectedkeys" {
		return "(name) => __SUI_TREE.selectedKeys(name)"
	}
	if key == "tree_checkedkeys" {
		return "(name) => __SUI_TREE.checkedKeys(name)"
	}
	if key == "tree_search" {
		return "(name, term) => __SUI_TREE.search(name, term)"
	}
	return "async () => [null, null]"
}

func childPtrs(a []ViewNode) []*ViewNode {
	var r []*ViewNode
	for i := 0; i < len(a); i++ {
		r = append(r, &a[i])
	}
	return r
}

// chartTag reports whether a tag renders through the shared chart runtime.
// "graph" is the dedicated block-diagram control; it always renders the state
// (shapes/links) geometry while chart keeps the explicit chartType attribute.
func chartTag(tag string) bool {
	return tag == "chart" || tag == "graph"
}

// layoutTableIntAttr reads a static integer attribute, clamped to [min,max].
func layoutTableIntAttr(e *Element, key string, def, min, max int) int {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	n := 0
	if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
		n = a.expr.i
	} else {
		v, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text)))
		if err != nil {
			return def
		}
		n = v
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func layoutTableColumns(e *Element) int {
	return layoutTableIntAttr(e, "columns", 2, 1, 12)
}

// layoutTableRows returns the explicit rows attribute, or 0 when rows are
// derived from the children.
func layoutTableRows(e *Element) int {
	return layoutTableIntAttr(e, "rows", 0, 0, 50)
}

func layoutTableTextAttr(e *Element, key, def string) string {
	a := findAttr(e, key)
	if a == nil || a.hasExpr {
		return def
	}
	v := strings.TrimSpace(partsContent(a.text))
	if v == "" {
		return def
	}
	return v
}

// layoutSafeSize keeps CSS length values written by the developer. Only
// straightforward lengths, percentages and intrinsic keywords are accepted.
func layoutSafeSize(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	switch v {
	case "auto", "fit-content", "min-content", "max-content":
		return v
	}
	if regexp.MustCompile(`^\d+(\.\d+)?$`).MatchString(v) {
		return v + "px"
	}
	if regexp.MustCompile(`^\d+(\.\d+)?(px|%|em|rem|vw|vh|pt)$`).MatchString(v) {
		return v
	}
	return ""
}

func layoutSafeCSSValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || strings.ContainsAny(v, ";{}<>\"'\n\r") {
		return ""
	}
	return v
}

func layoutSizeList(e *Element, key string) []string {
	a := findAttr(e, key)
	if a == nil || a.hasExpr {
		return nil
	}
	raw := strings.Split(partsContent(a.text), ",")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		out = append(out, layoutSafeSize(item))
	}
	return out
}

func layoutBorderCSS(e *Element) string {
	style := strings.ToLower(layoutTableTextAttr(e, "borderStyle", "none"))
	switch style {
	case "solid", "dashed", "dotted", "double":
	default:
		return ""
	}
	width := layoutSafeSize(layoutTableTextAttr(e, "borderWidth", "1px"))
	if width == "" {
		width = "1px"
	}
	color := layoutSafeCSSValue(layoutTableTextAttr(e, "borderColor", "var(--sui-border, #3d4556)"))
	if color == "" {
		color = "var(--sui-border, #3d4556)"
	}
	return "border:" + width + " " + style + " " + color + ";"
}

func layoutTableStyle(e *Element) string {
	var parts []string
	if v := layoutSafeSize(layoutTableTextAttr(e, "width", "100%")); v != "" {
		parts = append(parts, "width:"+v)
	}
	if v := layoutSafeSize(layoutTableTextAttr(e, "height", "")); v != "" {
		parts = append(parts, "height:"+v)
	}
	gapX := layoutSafeSize(layoutTableTextAttr(e, "gapX", "12px"))
	if gapX == "" {
		gapX = "12px"
	}
	gapY := layoutSafeSize(layoutTableTextAttr(e, "gapY", gapX))
	if gapY == "" {
		gapY = gapX
	}
	parts = append(parts, "border-spacing:"+gapX+" "+gapY)
	if v := layoutSafeSize(layoutTableTextAttr(e, "cellMinHeight", "72px")); v != "" {
		parts = append(parts, "--sui-layout-cell-min-height:"+v)
	}
	return strings.Join(parts, ";") + ";"
}

func layoutTableStyleWithUser(e *Element) string {
	base := layoutTableStyle(e)
	a := findAttr(e, "style")
	if a == nil || a.hasExpr {
		return base
	}
	user := strings.TrimSpace(partsContent(a.text))
	if user == "" {
		return base
	}
	return base + user + ";"
}

func layoutTableStyleAttrValue(e *Element) string {
	base := htmlEscape(layoutTableStyle(e))
	a := findAttr(e, "style")
	if a == nil {
		return base
	}
	if a.hasExpr {
		return base + "${__sui_esc(" + ExprToJs(a.expr) + ")}"
	}
	return base + htmlTextParts(a.text)
}

type layoutGridCell struct {
	node             *ViewNode
	content          []ViewNode
	explicit         bool
	row, col         int
	rowspan, colspan int
}

type layoutGrid struct {
	rows     int
	cols     int
	cells    []layoutGridCell
	starts   map[[2]int]int
	occupied [][]bool
}

func isLayoutCellNode(n *ViewNode) bool {
	return n != nil && n.kind == "element" && n.elem != nil && n.elem.tag == "layoutcell"
}

func layoutCellElement(cell layoutGridCell) *Element {
	if cell.node != nil && cell.node.kind == "element" {
		return cell.node.elem
	}
	return nil
}

func layoutSpanValue(e *Element, key string, max int) int {
	n := layoutTableIntAttr(e, key, 1, 1, 50)
	if max > 0 && n > max {
		n = max
	}
	return n
}

func layoutTableGrid(e *Element) layoutGrid {
	cols := layoutTableColumns(e)
	explicitRows := layoutTableRows(e)
	explicitCells := []int{}
	for i := range e.children {
		if isLayoutCellNode(&e.children[i]) {
			explicitCells = append(explicitCells, i)
		}
	}
	if len(explicitCells) > 0 {
		maxCol, maxRow := cols, explicitRows
		for _, i := range explicitCells {
			cell := e.children[i].elem
			row := layoutTableIntAttr(cell, "row", 0, 0, 49)
			col := layoutTableIntAttr(cell, "col", 0, 0, 11)
			rs := layoutSpanValue(cell, "rowspan", 50)
			cs := layoutSpanValue(cell, "colspan", 12)
			if col+cs > maxCol {
				maxCol = col + cs
			}
			if row+rs > maxRow {
				maxRow = row + rs
			}
		}
		cols = maxCol
		if maxRow < 1 {
			maxRow = 1
		}
		explicitRows = maxRow
	}
	occ := [][]bool{}
	ensure := func(n int) {
		for len(occ) < n {
			occ = append(occ, make([]bool, cols))
		}
	}
	if explicitRows > 0 {
		ensure(explicitRows)
	}
	ensure(1)
	g := layoutGrid{cols: cols, starts: map[[2]int]int{}}
	place := func(node *ViewNode, explicit bool, row, col, cs, rs int) {
		ensure(row + rs)
		for rr := 0; rr < rs; rr++ {
			for cc := 0; cc < cs; cc++ {
				occ[row+rr][col+cc] = true
			}
		}
		g.starts[[2]int{row, col}] = len(g.cells)
		cell := layoutGridCell{node: node, explicit: explicit, row: row, col: col, rowspan: rs, colspan: cs}
		if explicit && node != nil && node.elem != nil {
			cell.content = node.elem.children
		} else if node != nil {
			cell.content = []ViewNode{*node}
		}
		g.cells = append(g.cells, cell)
	}
	for _, i := range explicitCells {
		node := &e.children[i]
		cell := node.elem
		row := layoutTableIntAttr(cell, "row", 0, 0, 49)
		col := layoutTableIntAttr(cell, "col", 0, 0, 11)
		if col >= cols {
			col = cols - 1
		}
		cs := layoutSpanValue(cell, "colspan", cols-col)
		rs := layoutSpanValue(cell, "rowspan", 50)
		if _, exists := g.starts[[2]int{row, col}]; exists {
			continue
		}
		place(node, true, row, col, cs, rs)
	}
	r, c := 0, 0
	for i := range e.children {
		if isLayoutCellNode(&e.children[i]) {
			continue
		}
		node := &e.children[i]
		for {
			ensure(r + 1)
			if c >= cols {
				r++
				c = 0
				continue
			}
			if occ[r][c] {
				c++
				continue
			}
			break
		}
		cs, rs := 1, 1
		if node.kind == "element" && node.elem != nil {
			cs = layoutSpanValue(node.elem, "colspan", cols-c)
			rs = layoutSpanValue(node.elem, "rowspan", 50)
		}
		if explicitRows > 0 && r+rs > explicitRows {
			rs = explicitRows - r
			if rs < 1 {
				rs = 1
			}
		}
		place(node, false, r, c, cs, rs)
		c += cs
	}
	if explicitRows > len(occ) {
		ensure(explicitRows)
	}
	g.rows = len(occ)
	g.occupied = occ
	return g
}

func layoutCellPosAttrs(row, col int) string {
	return " data-sui-layout-row=\"" + strconv.Itoa(row) + "\" data-sui-layout-col=\"" + strconv.Itoa(col) + "\""
}

func layoutCellBorderCSS(child, table *Element) string {
	style := layoutTableTextAttr(child, "cellBorderStyle", "")
	width := layoutTableTextAttr(child, "cellBorderWidth", "")
	color := layoutTableTextAttr(child, "cellBorderColor", "")
	if style == "" && width == "" && color == "" {
		return layoutBorderCSS(table)
	}
	if style == "" {
		style = layoutTableTextAttr(table, "borderStyle", "none")
	}
	switch strings.ToLower(style) {
	case "solid", "dashed", "dotted", "double":
	default:
		return ""
	}
	if width == "" {
		width = layoutTableTextAttr(table, "borderWidth", "1px")
	}
	width = layoutSafeSize(width)
	if width == "" {
		width = "1px"
	}
	if color == "" {
		color = layoutTableTextAttr(table, "borderColor", "var(--sui-border, #3d4556)")
	}
	color = layoutSafeCSSValue(color)
	if color == "" {
		color = "var(--sui-border, #3d4556)"
	}
	return "border:" + width + " " + strings.ToLower(style) + " " + color + ";"
}

func layoutCellStyle(cell layoutGridCell, table *Element) string {
	if cell.node == nil || cell.node.kind != "element" || cell.node.elem == nil {
		return layoutBorderCSS(table)
	}
	child := cell.node.elem
	var parts []string
	if v := layoutSafeCSSValue(layoutTableTextAttr(child, "cellPadding", "")); v != "" {
		parts = append(parts, "padding:"+v)
	}
	if v := layoutSafeCSSValue(layoutTableTextAttr(child, "cellBackground", "")); v != "" {
		parts = append(parts, "background:"+v)
	}
	if border := layoutCellBorderCSS(child, table); border != "" {
		parts = append(parts, border)
	}
	return strings.Join(parts, ";")
}

func layoutCellClass(cell layoutGridCell) string {
	if len(cell.content) == 0 {
		return " sui-layout-cell-empty"
	}
	return ""
}

// layoutCellClassSuffix appends the user class assigned by the inspector to the
// generated <td>, so generic sections (background, border, effects, ...) can
// target the layout cell itself.
func layoutCellClassSuffix(cell layoutGridCell) string {
	if e := layoutCellElement(cell); e != nil {
		if a := findAttr(e, "class"); a != nil {
			if a.hasExpr {
				return " ${" + classExprJs(ExprToJs(a.expr)) + "}"
			}
			if c := classPartsJs(a.text); c != "" {
				return " " + c
			}
		}
	}
	return ""
}

func layoutCellStyleAttrValue(cell layoutGridCell, table *Element, dynamic bool) string {
	base := layoutCellStyle(cell, table)
	if base != "" {
		base += ";"
	}
	base = htmlEscape(base)
	var child *Element
	if cell.node != nil && cell.node.kind == "element" {
		child = cell.node.elem
	}
	if child == nil {
		return base
	}
	a := findAttr(child, "style")
	if a == nil {
		return base
	}
	if a.hasExpr {
		if dynamic {
			return base + "${__sui_esc(" + ExprToJs(a.expr) + ")}"
		}
		return base
	}
	return base + htmlTextParts(a.text)
}

func writeLayoutCellExtras(sb *strings.Builder, e *Element, ctx *JsCtx) {
	writeDndAttrs(sb, e, ctx)
	writePickerAttrs(sb, e, ctx)
	writeMotionAttrs(sb, e, ctx)
	if a := findAttr(e, "tooltip"); a != nil {
		sb.WriteString(" data-sui-tooltip=\"")
		sb.WriteString(htmlAttrValue(a, false))
		sb.WriteString("\"")
	}
	keyboard := false
	for _, evKey := range elemEventOrder {
		ev := findAttr(e, evKey)
		if ev == nil || !ev.hasExpr || ev.expr == nil || ev.expr.kind != "ident" {
			continue
		}
		if evKey == "onKeyDown" || evKey == "onKeyUp" || evKey == "onFocus" || evKey == "onBlur" {
			keyboard = true
		}
		sb.WriteString(" " + eventAttrJs[evKey] + "=\"var __t=event.target.closest('[data-sui-id]'); if (!__t || !this.contains(__t)) " + ev.expr.name + "()\"")
	}
	if ev := findAttr(e, "onContextMenu"); ev != nil && ev.hasExpr && ev.expr != nil && ev.expr.kind == "ident" {
		sb.WriteString(" oncontextmenu=\"var __t=event.target.closest('[data-sui-id]'); if (!__t || !this.contains(__t)) { event.preventDefault(); " + ev.expr.name + "(); }\"")
	}
	if keyboard {
		sb.WriteString(" tabindex=\"0\"")
	}
}

func layoutCellAttrs(cell layoutGridCell, e *Element, dynamic bool) string {
	var b strings.Builder
	if cell.colspan > 1 {
		b.WriteString(" colspan=\"")
		b.WriteString(strconv.Itoa(cell.colspan))
		b.WriteString("\"")
	}
	if cell.rowspan > 1 {
		b.WriteString(" rowspan=\"")
		b.WriteString(strconv.Itoa(cell.rowspan))
		b.WriteString("\"")
	}
	if style := layoutCellStyleAttrValue(cell, e, dynamic); style != "" {
		b.WriteString(" style=\"")
		b.WriteString(style)
		b.WriteString("\"")
	}
	return b.String()
}

func layoutEmptyCellAttrs(e *Element) string {
	if border := layoutBorderCSS(e); border != "" {
		return " style=\"" + htmlEscape(border) + "\""
	}
	return ""
}

func layoutColgroupHTML(e *Element, cols int) string {
	widths := layoutSizeList(e, "columnWidths")
	if len(widths) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<colgroup>")
	for i := 0; i < cols; i++ {
		w := ""
		if i < len(widths) {
			w = widths[i]
		}
		if w != "" {
			b.WriteString("<col style=\"width:" + htmlEscape(w) + "\">")
		} else {
			b.WriteString("<col>")
		}
	}
	b.WriteString("</colgroup>")
	return b.String()
}

func layoutRowStyle(e *Element, row int) string {
	heights := layoutSizeList(e, "rowHeights")
	if row < len(heights) && heights[row] != "" {
		return " style=\"height:" + htmlEscape(heights[row]) + "\""
	}
	return ""
}

func elemHtmlTag(tag string) string {
	if tag == "text" {
		return "span"
	}
	if tag == "body" {
		return "div"
	}
	if tag == "row" || tag == "col" || tag == "spacer" || tag == "card" || tag == "tabs" || tag == "accordion" || tag == "accordionitem" || tag == "toast" {
		return "div"
	}
	if tag == "tablelayout" {
		return "table"
	}
	if tag == "menubar" {
		return "nav"
	}
	if tag == "combobox" || tag == "datepicker" || tag == "tree" || tag == "chart" || tag == "graph" {
		return "div"
	}
	if tag == "menu" || tag == "menuitem" || tag == "tab" || tag == "modal" || tag == "dropdown" {
		return "div"
	}
	if tag == "popover" {
		return "span"
	}
	if tag == "drawer" {
		return "div"
	}
	if tag == "splitter" {
		return "div"
	}
	if tag == "pane" {
		return "div"
	}
	if tag == "dropzone" {
		return "div"
	}
	if tag == "timeline" || tag == "timelineitem" || tag == "emptystate" || tag == "skeleton" ||
		tag == "carousel" || tag == "slide" || tag == "stepper" || tag == "step" || tag == "commandbar" {
		return "div"
	}
	if tag == "command" {
		return "button"
	}
	if tag == "dropdownitem" {
		return "button"
	}
	if tag == "modalbutton" {
		return "button"
	}
	if tag == "divider" {
		return "hr"
	}
	if tag == "checkbox" || tag == "radio" || tag == "switch" || tag == "slider" || tag == "file" || tag == "color" || tag == "date" || tag == "time" || tag == "number" {
		return "input"
	}
	if tag == "textarea" {
		return "textarea"
	}
	if tag == "select" {
		return "select"
	}
	if tag == "progress" {
		return "progress"
	}
	if tag == "badge" {
		return "span"
	}
	if tag == "alert" {
		return "div"
	}
	if tag == "spinner" {
		return "div"
	}
	if tag == "trow" {
		return "tr"
	}
	if tag == "tcell" {
		return "td"
	}
	return tag
}

// isHeaderCell reports whether a tcell is declared `head: true`.
func isHeaderCell(e *Element) bool {
	a := findAttr(e, "head")
	return a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b
}

func isVoidTag(tag string) bool {
	if tag == "input" || tag == "img" || tag == "hr" || tag == "br" {
		return true
	}
	if tag == "checkbox" || tag == "radio" || tag == "switch" || tag == "slider" || tag == "file" || tag == "color" || tag == "date" || tag == "time" || tag == "number" {
		return true
	}
	if tag == "divider" {
		return true
	}
	return false
}

// inputOptions returns the comma-separated `options` attribute values for an
// <input> (used to build the autocomplete dropdown).
func inputOptions(e *Element) []string {
	a := findAttr(e, "options")
	if a == nil || a.hasExpr {
		return nil
	}
	raw := partsLiteral(a.text)
	var out []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func datalistHtml(id string, opts []string) string {
	var sb strings.Builder
	sb.WriteString("<datalist id=\"" + id + "\">")
	for _, o := range opts {
		sb.WriteString("<option value=\"" + htmlEscape(o) + "\"></option>")
	}
	sb.WriteString("</datalist>")
	return sb.String()
}

func findAttr(e *Element, key string) *Attr {
	for i := 0; i < len(e.attrs); i++ {
		if e.attrs[i].key == key {
			return &e.attrs[i]
		}
	}
	return nil
}

// eventAttrJs lists the event handler attributes supported on elements. Each
// maps to the corresponding DOM property used for query-based wiring.
var eventAttrJs = map[string]string{
	"onClick":       "onclick",
	"onDoubleClick": "ondblclick",
	"onChange":      "onchange",
	"onInput":       "oninput",
	"onSubmit":      "onsubmit",
	"onReset":       "onreset",
	"onFocus":       "onfocus",
	"onBlur":        "onblur",
	"onMouseEnter":  "onmouseenter",
	"onMouseLeave":  "onmouseleave",
	"onKeyDown":     "onkeydown",
	"onKeyUp":       "onkeyup",
	"onDragStart":   "ondragstart",
	"onDragEnter":   "ondragenter",
	"onDragOver":    "ondragover",
	"onDragLeave":   "ondragleave",
	"onDrop":        "ondrop",
	"onDragEnd":     "ondragend",
}

// elemEventOrder is a stable iteration order for event attributes.
var elemEventOrder = []string{
	"onClick", "onDoubleClick", "onChange", "onInput", "onSubmit", "onReset",
	"onFocus", "onBlur", "onMouseEnter", "onMouseLeave", "onKeyDown", "onKeyUp",
	"onDragStart", "onDragEnter", "onDragOver", "onDragLeave", "onDrop", "onDragEnd",
}

// dndEventOrder is a subset used for inline event handlers. A list body is
// rendered once per item, so query-based wiring would only reach the first
// row; inline handlers make every generated row a real drag source/target.
var dndEventOrder = []string{
	"onDragStart", "onDragEnter", "onDragOver", "onDragLeave", "onDrop", "onDragEnd",
}

func isDndEventKey(key string) bool {
	for _, k := range dndEventOrder {
		if k == key {
			return true
		}
	}
	return false
}

// hasEventAttr reports whether the element carries any event handler attribute.
func hasEventAttr(e *Element) bool {
	for i := 0; i < len(e.attrs); i++ {
		if _, ok := eventAttrJs[e.attrs[i].key]; ok {
			return true
		}
	}
	return false
}

func isInteractiveElem(e *Element, ctx *JsCtx) bool {
	if e.tag == "button" || e.tag == "input" || e.tag == "a" || e.tag == "textarea" || e.tag == "select" {
		return true
	}
	// Value-carrying controls: these need wiring so that a `value:`/`checked:`
	// binding both renders from state AND writes back to it. Without this a
	// lone "switch checked: on" (no explicit handler) would never update state.
	switch e.tag {
	case "checkbox", "switch", "radio", "slider", "number", "file", "color", "date", "time":
		return true
	}
	if hasEventAttr(e) {
		return true
	}
	if e.tag == "img" {
		src := findAttr(e, "src")
		if src != nil && src.hasExpr && src.expr != nil && src.expr.kind == "ident" && ctxIsState(ctx, src.expr.name) {
			return true
		}
	}
	return false
}

func openTagHtml(e *Element, tag string, interactive bool, idx int, ctx *JsCtx) string {
	return openTagHtmlWithSrc(e, tag, interactive, idx, ctx, "")
}

func openTagHtmlWithSrc(e *Element, tag string, interactive bool, idx int, ctx *JsCtx, iframeSrc string) string {
	var sb strings.Builder
	sb.WriteString("<" + tag)
	if e.tag == "button" || e.tag == "dropdownitem" || e.tag == "command" {
		sb.WriteString(" type=\"button\"")
	}
	if e.tag == "row" {
		writeContainerClass(&sb, e, ctx, "row")
		writeDndAttrs(&sb, e, ctx)
		writePickerAttrs(&sb, e, ctx)
	} else if e.tag == "col" {
		writeContainerClass(&sb, e, ctx, "col")
		writeDndAttrs(&sb, e, ctx)
		writePickerAttrs(&sb, e, ctx)
	} else if e.tag == "spacer" {
		writeContainerClass(&sb, e, ctx, "spacer")
		writeDndAttrs(&sb, e, ctx)
		writePickerAttrs(&sb, e, ctx)
	} else {
		// all other elements handled by writeUserAttrs (classes/types for new controls)
		writeUserAttrsWithSrc(&sb, e, ctx, iframeSrc)
	}
	if e.tag == "modal" {
		sb.WriteString(" data-sui-modal=\"" + htmlEscape(modalNameOf(e)) + "\" aria-hidden=\"true\" hidden")
	} else if e.tag == "tablelayout" {
		role := "presentation"
		if datepickerBool(e, "semantic") {
			role = "table"
		}
		sb.WriteString(" role=\"" + role + "\"")
		if label := layoutTableTextAttr(e, "label", ""); label != "" {
			sb.WriteString(" aria-label=\"" + htmlEscape(label) + "\"")
		}
	} else if e.tag == "tabs" {
		if v := treeTextAttr(e, "position"); v == "bottom" {
			sb.WriteString(" data-sui-tabs-position=\"bottom\"")
		}
		if v := treeTextAttr(e, "variant"); v == "sheets" {
			sb.WriteString(" data-sui-tabs-variant=\"sheets\"")
		}
		if len(e.text) > 0 {
			sb.WriteString(" aria-label=\"" + htmlTextParts(e.text) + "\"")
		}
		if idx >= 0 {
			if key := tabsRuntimeKey(e); key != "" {
				sb.WriteString(" data-sui-tabs-key=\"" + htmlEscape(key) + "\"")
			}
		}
	} else if e.tag == "accordion" {
		if len(e.text) > 0 {
			sb.WriteString(" aria-label=\"" + htmlTextParts(e.text) + "\"")
		}
		if idx >= 0 {
			sb.WriteString(" data-sui-accordion-key=\"sui-accordion-" + strconv.Itoa(idx) + "\"")
		}
	} else if e.tag == "toast" {
		sb.WriteString(" data-sui-toast-host=\"1\" aria-live=\"polite\" aria-atomic=\"false\"")
		if v := toastAttrValue(e, "duration"); v != "" {
			sb.WriteString(" data-sui-toast-duration=\"" + htmlEscape(v) + "\"")
		}
		if v := toastAttrValue(e, "max"); v != "" {
			sb.WriteString(" data-sui-toast-max=\"" + htmlEscape(v) + "\"")
		}
		if v := toastAttrValue(e, "icons"); v != "" {
			sb.WriteString(" data-sui-toast-icons=\"" + htmlEscape(v) + "\"")
		}
		if v := toastOffsetValue(e, "x"); v != "" {
			sb.WriteString(" data-sui-toast-x=\"" + htmlEscape(v) + "\"")
		}
		if v := toastOffsetValue(e, "y"); v != "" {
			sb.WriteString(" data-sui-toast-y=\"" + htmlEscape(v) + "\"")
		}
	} else if e.tag == "menubar" {
		sb.WriteString(" role=\"menubar\"")
		if len(e.text) > 0 {
			sb.WriteString(" aria-label=\"" + htmlTextParts(e.text) + "\"")
		}
	} else if e.tag == "menu" {
		sb.WriteString(" role=\"none\"")
	} else if e.tag == "dropdown" {
		sb.WriteString(" role=\"none\" data-sui-dropdown")
		pos := findAttr(e, "position")
		if pos != nil && !pos.hasExpr && strings.EqualFold(strings.TrimSpace(partsContent(pos.text)), "right") {
			sb.WriteString(" data-sui-dropdown-right")
		}
		sb.WriteString(" data-sui-dropdown-placement=\"" + attrChoice(e, "placement", []string{"auto", "bottom", "top", "left", "right"}, "auto") + "\"")
		sb.WriteString(dataAnimAttrs(e, "fade"))
	} else if e.tag == "menuitem" || e.tag == "dropdownitem" {
		sb.WriteString(" role=\"menuitem\" tabindex=\"-1\"")
		if e.tag == "dropdownitem" && datepickerBool(e, "disabled") {
			sb.WriteString(" disabled")
		}
	}
	if e.tag == "iframe" {
		// Secure defaults: if the author omitted sandbox, deny scripts, forms,
		// popups and same-origin access. Referrer is removed by default.
		if findAttr(e, "sandbox") == nil && genIframeCfg.IframeSandbox != "allow" {
			sb.WriteString(" sandbox=\"\"")
		}
		if findAttr(e, "referrerpolicy") == nil && genIframeCfg.IframeReferrer != "unset" {
			sb.WriteString(" referrerpolicy=\"" + htmlEscape(genIframeCfg.IframeReferrer) + "\"")
		}
	}
	// Table behavior flags: rendered as data attributes so the base CSS and the
	// small runtime script can make cells editable / truncate long content.
	if e.tag == "table" {
		if v := treeTextAttr(e, "name"); v != "" {
			sb.WriteString(" name=\"" + htmlEscape(v) + "\"")
		}
		tableEditMode := treeTextAttr(e, "editMode")
		if tableEditMode == "editable" || tableEditMode == "persistent" {
			sb.WriteString(" data-sui-editable=\"1\"")
		}
		if tableEditMode != "" {
			sb.WriteString(" data-sui-edit-mode=\"" + htmlEscape(tableEditMode) + "\"")
		}
		if v := treeTextAttr(e, "cellText"); v != "" {
			sb.WriteString(" data-sui-cell-text=\"" + htmlEscape(v) + "\"")
			if v == "truncate" {
				sb.WriteString(" data-sui-fixed=\"1\"")
			}
		}
		for _, pair := range [][2]string{
			{"columns", "data-sui-columns"}, {"columnLabels", "data-sui-column-labels"},
			{"columnWidths", "data-sui-column-widths"}, {"columnAlign", "data-sui-column-align"},
			{"columnTypes", "data-sui-column-types"}, {"columnFormats", "data-sui-column-formats"},
			{"columnOptions", "data-sui-column-options"},
		} {
			if v := treeTextAttr(e, pair[0]); v != "" {
				sb.WriteString(" " + pair[1] + "=\"" + htmlEscape(v) + "\"")
			}
		}
		if datepickerBool(e, "sortable") {
			sb.WriteString(" data-sui-sortable=\"1\"")
		}
		if datepickerBool(e, "filterable") {
			sb.WriteString(" data-sui-filterable=\"1\"")
		}
		if datepickerBool(e, "multiSelect") {
			sb.WriteString(" data-sui-multi=\"1\"")
		}
		if datepickerBool(e, "virtual") {
			sb.WriteString(" data-sui-virtual=\"1\"")
		}
		if datepickerBool(e, "showAddTable") {
			sb.WriteString(" data-sui-show-add-table=\"1\"")
		}
		if datepickerBool(e, "structureEditable") {
			sb.WriteString(" data-sui-structure-editable=\"1\"")
		}
		if v := comboboxAttrNumber(e, "pageSize"); v != "" {
			sb.WriteString(" data-sui-page-size=\"" + htmlEscape(v) + "\"")
		}
		if v := treeTextAttr(e, "selection"); v != "" {
			sb.WriteString(" data-sui-selection=\"" + htmlEscape(v) + "\"")
		}
		for _, pair := range [][2]string{
			{"onCellChange", "data-sui-table-oncellchange"}, {"onRowChange", "data-sui-table-onrowchange"},
			{"onTableChange", "data-sui-table-ontablechange"}, {"onRowSelect", "data-sui-table-onrowselect"},
			{"onSort", "data-sui-table-onsort"}, {"onCellClick", "data-sui-table-oncellclick"},
			{"onRowClick", "data-sui-table-onrowclick"},
		} {
			if name := eventIdentOf(e, pair[0]); name != "" {
				sb.WriteString(" " + pair[1] + "=\"" + htmlEscape(name) + "\"")
			}
		}
	}
	// `input "text"` uses the element text as its placeholder (see SPEC §2.5).
	// An explicit placeholder: attribute wins.
	if e.tag == "input" && findAttr(e, "placeholder") == nil && len(e.text) > 0 {
		if c := partsContent(e.text); c != "" {
			sb.WriteString(" placeholder=\"" + htmlTextParts(e.text) + "\"")
		}
	}
	if !genCompBody {
		// In a root view handlers are wired through the generated JS
		// (collectWiring), keyed by this dfs-order id; the design tool also
		// uses it to translate a click inside the live preview into a tree
		// selection. Wrapped label controls (checkbox/switch/radio) pass -1
		// here: the id goes on their wrapper <label> instead.
		if idx >= 0 {
			sb.WriteString(" data-sui-id=\"" + strconv.Itoa(idx) + "\"")
		}
	} else if interactive {
		// Inside a component, instances may repeat: use inline handlers on
		// the global functions instead of query-based wiring with ids.
		for _, evKey := range elemEventOrder {
			dom := eventAttrJs[evKey]
			ev := findAttr(e, evKey)
			if ev == nil || !ev.hasExpr || ev.expr == nil || ev.expr.kind != "ident" {
				continue
			}
			if genCompMethods[ev.expr.name] {
				// A view-local method: dispatch to THIS instance store.
				sb.WriteString(" " + dom + "=\"__sui_vm(${JSON.stringify(__key)}, '" + ev.expr.name + "')\"")
			} else if genCompFn[ev.expr.name] {
				// A "fn" component parameter: interpolate its (runtime)
				// handler-name string into the inline attribute. Escape it so a
				// value cannot break out of the attribute.
				sb.WriteString(" " + dom + "=\"${__sui_esc(" + ev.expr.name + ")}()\"")
			} else {
				sb.WriteString(" " + dom + "=\"" + ev.expr.name + "()\"")
			}
		}
		if e.tag == "input" || e.tag == "textarea" || e.tag == "select" {
			val := findAttr(e, "value")
			if val != nil && val.hasExpr && val.expr != nil && val.expr.kind == "ident" && ctxIsState(ctx, val.expr.name) {
				sb.WriteString(" oninput=\"" + setterName(val.expr.name) + "(this.value)\"")
			}
		} else if e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio" {
			val := findAttr(e, "checked")
			if val != nil && val.hasExpr && val.expr != nil && val.expr.kind == "ident" && ctxIsState(ctx, val.expr.name) {
				sb.WriteString(" onchange=\"" + setterName(val.expr.name) + "(this.checked)\"")
			}
		} else if e.tag == "slider" || e.tag == "number" {
			val := findAttr(e, "value")
			if val != nil && val.hasExpr && val.expr != nil && val.expr.kind == "ident" && ctxIsState(ctx, val.expr.name) {
				sb.WriteString(" oninput=\"" + setterName(val.expr.name) + "(this.value)\"")
			}
		}
	}
	if !genCompBody {
		for _, evKey := range dndEventOrder {
			ev := findAttr(e, evKey)
			if ev == nil || !ev.hasExpr || ev.expr == nil || ev.expr.kind != "ident" {
				continue
			}
			sb.WriteString(" " + eventAttrJs[evKey] + "=\"" + ev.expr.name + "()\"")
		}
	}
	sb.WriteByte('>')
	return sb.String()
}

// newControlClass maps a control tag to its built-in base class and input type.
var newControlClass = map[string]string{
	"checkbox":      "sui-checkbox",
	"radio":         "sui-radio",
	"switch":        "sui-switch",
	"slider":        "sui-slider",
	"file":          "sui-file",
	"color":         "sui-color",
	"date":          "sui-date",
	"time":          "sui-time",
	"number":        "sui-number",
	"textarea":      "sui-textarea",
	"select":        "sui-select",
	"input":         "sui-input",
	"button":        "sui-button",
	"progress":      "sui-progress",
	"badge":         "sui-badge",
	"alert":         "sui-alert",
	"spinner":       "sui-spinner",
	"card":          "sui-card",
	"tablelayout":   "sui-layout-table",
	"divider":       "sui-divider",
	"tabs":          "sui-tabs",
	"tab":           "sui-tabpanel",
	"accordion":     "sui-accordion",
	"accordionitem": "sui-accordion-item",
	"toast":         "sui-toast-host",
	"combobox":      "sui-combobox",
	"datepicker":    "sui-datepicker",
	"tree":          "sui-tree",
	"treeitem":      "sui-tree-item",
	"chart":         "sui-chart",
	"graph":         "sui-chart",
	"modal":         "sui-modal",
	"modalbutton":   "sui-modal-button",
	"menubar":       "sui-menubar",
	"menu":          "sui-menu",
	"menuitem":      "sui-menuitem",
	"dropdown":      "sui-menu",
	"dropdownitem":  "sui-menuitem",
}

// newControlType maps a control tag to the HTML input type it renders with.
var newControlType = map[string]string{
	"checkbox": "checkbox",
	"radio":    "radio",
	"switch":   "checkbox",
	"slider":   "range",
	"file":     "file",
	"color":    "color",
	"date":     "date",
	"time":     "time",
	"number":   "number",
}

// writeDndAttrs emits the drag-and-drop data attributes shared by every
// element, including row/col/spacer which use a specialized class writer.
func writeDndAttrs(sb *strings.Builder, e *Element, ctx *JsCtx) {
	if findAttr(e, "onDrop") != nil || findAttr(e, "dropKey") != nil || findAttr(e, "dropFiles") != nil {
		sb.WriteString(" data-sui-dropzone=\"1\"")
	}
	if findAttr(e, "dragKey") != nil && findAttr(e, "draggable") == nil {
		// A drag key is enough to make the element a drag source; requiring a
		// second `draggable: true` attribute would be easy to forget.
		sb.WriteString(" draggable=\"true\"")
	}
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		switch a.key {
		case "draggable":
			sb.WriteString(" draggable=\"")
			if a.hasExpr && a.expr != nil {
				if a.expr.kind == "bool" {
					if a.expr.b {
						sb.WriteString("true")
					} else {
						sb.WriteString("false")
					}
				} else {
					sb.WriteString("${(")
					sb.WriteString(ExprToJs(a.expr))
					sb.WriteString(") ? \"true\" : \"false\"}")
				}
			} else {
				v := strings.TrimSpace(strings.ToLower(partsContent(a.text)))
				if v == "false" || v == "0" {
					sb.WriteString("false")
				} else {
					sb.WriteString("true")
				}
			}
			sb.WriteString("\"")
		case "directory":
			// Native directory input only. Non-file picker triggers carry the
			// flag as data-sui-pick-directory (see writePickerAttrs).
			if e.tag == "file" {
				if a.hasExpr && a.expr != nil {
					if a.expr.kind == "bool" {
						if a.expr.b {
							sb.WriteString(" webkitdirectory")
						}
					} else {
						sb.WriteString("${(")
						sb.WriteString(ExprToJs(a.expr))
						sb.WriteString(") ? \" webkitdirectory\" : \"\"}")
					}
				} else {
					v := strings.TrimSpace(strings.ToLower(partsContent(a.text)))
					if v != "false" && v != "0" && v != "" {
						sb.WriteString(" webkitdirectory")
					}
				}
			}
		case "dragKey", "dropKey":
			if a.key == "dragKey" {
				sb.WriteString(" data-sui-drag-key=\"")
			} else {
				sb.WriteString(" data-sui-drop-key=\"")
			}
			sb.WriteString(htmlAttrValue(&a, false))
			sb.WriteString("\"")
		case "dropFiles":
			if a.hasExpr && a.expr != nil && a.expr.kind == "ident" && ctxIsState(ctx, a.expr.name) {
				sb.WriteString(" data-sui-dropfiles=\"" + setterName(a.expr.name) + "\"")
			}
		}
	}
}

func writeDataBoolAttr(sb *strings.Builder, name string, a *Attr) {
	if a == nil {
		return
	}
	sb.WriteString(" " + name + "=\"")
	if a.hasExpr && a.expr != nil {
		if a.expr.kind == "bool" {
			if a.expr.b {
				sb.WriteString("1")
			} else {
				sb.WriteString("0")
			}
		} else {
			sb.WriteString("${(")
			sb.WriteString(ExprToJs(a.expr))
			sb.WriteString(") ? \"1\" : \"0\"}")
		}
	} else {
		v := strings.TrimSpace(strings.ToLower(partsContent(a.text)))
		if v != "" && v != "false" && v != "0" {
			sb.WriteString("1")
		} else {
			sb.WriteString("0")
		}
	}
	sb.WriteString("\"")
}

// writePickerAttrs emits the native-dialog selector metadata. On a <file>
// control the browser's native input is used directly; on a button/container
// the runtime creates a hidden file input when the element is clicked.
func writePickerAttrs(sb *strings.Builder, e *Element, ctx *JsCtx) {
	pf := findAttr(e, "pickFiles")
	if pf == nil || !pf.hasExpr || pf.expr == nil || pf.expr.kind != "ident" || !ctxIsState(ctx, pf.expr.name) {
		return
	}
	sb.WriteString(" data-sui-pickfiles=\"" + setterName(pf.expr.name) + "\"")
	if pm := findAttr(e, "pickMode"); pm != nil && !pm.hasExpr {
		v := strings.ToLower(strings.TrimSpace(partsContent(pm.text)))
		if v != "" {
			sb.WriteString(" data-sui-pick-mode=\"" + v + "\"")
		}
	}
	if e.tag == "file" {
		return
	}
	if a := findAttr(e, "accept"); a != nil {
		sb.WriteString(" data-sui-pick-accept=\"")
		sb.WriteString(htmlAttrValue(a, false))
		sb.WriteString("\"")
	}
	writeDataBoolAttr(sb, "data-sui-pick-multiple", findAttr(e, "multiple"))
	writeDataBoolAttr(sb, "data-sui-pick-directory", findAttr(e, "directory"))
}

func writeUserAttrs(sb *strings.Builder, e *Element, ctx *JsCtx) {
	writeUserAttrsWithSrc(sb, e, ctx, "")
}

// writeUserAttrsWithSrc is writeUserAttrs with an optional iframe src
// override (a JS expression yielding an already validated URL).
func writeUserAttrsWithSrc(sb *strings.Builder, e *Element, ctx *JsCtx, iframeSrc string) {
	// Auto-apply the base class and input type for new controls. The base class
	// always stays; a user class (static or reactive) is appended to it.
	if t, ok := newControlType[e.tag]; ok {
		sb.WriteString(" type=\"" + t + "\"")
	}
	if base, ok := newControlClass[e.tag]; ok {
		if e.tag == "toast" {
			if pc := toastPlacementClasses(e); pc != "" {
				base += " " + pc
			}
		}
		if e.tag == "spinner" && spinnerImageMode(e) {
			base += " sui-spinner-image"
		}
		if e.tag == "tablelayout" && datepickerBool(e, "responsive") {
			base += " sui-layout-responsive"
		}
		if e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio" {
			base += togglePresetClasses(e)
		}
		writeContainerClass(sb, e, ctx, base)
	} else {
		// Plain elements (p, div, span, ...): emit the class attribute from the
		// user's static or reactive expression.
		writePlainClass(sb, e)
	}

	writeDndAttrs(sb, e, ctx)
	writePickerAttrs(sb, e, ctx)
	writeMotionAttrs(sb, e, ctx)
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		if _, isEv := eventAttrJs[a.key]; isEv {
			continue
		}
		if e.tag == "errorboundary" && a.key == "fallback" {
			continue
		}
		if e.tag == "spinner" && spinnerImageMode(e) && (a.key == "src" || a.key == "alt") {
			// The image source and its alternative text belong to the inner
			// <img>, emitted by spinnerImageElemJs.
			continue
		}
		if e.tag == "img" && a.key == "src" {
			// The designer asset manager embeds uploaded images as
			// data:image/... URIs; the image guard accepts them, while the
			// generic URL guard would replace them with "#".
			sb.WriteString(" src=\"")
			sb.WriteString(htmlAttrValueImage(&a))
			sb.WriteString("\"")
			continue
		}
		if (e.tag == "modal" && a.key == "name") ||
			(e.tag == "modalbutton" && (a.key == "variant" || a.key == "close")) {
			continue
		}
		if a.key == "tooltip" {
			sb.WriteString(" data-sui-tooltip=\"")
			sb.WriteString(htmlAttrValue(&a, false))
			sb.WriteString("\"")
			continue
		}
		if e.tag == "combobox" && (a.key == "value" || a.key == "options" || a.key == "placeholder" ||
			a.key == "empty" || a.key == "disabled" || a.key == "required" ||
			a.key == "maxOptionLength" || a.key == "maxOptions") {
			continue
		}
		if e.tag == "datepicker" && (a.key == "value" || a.key == "mode" || a.key == "min" ||
			a.key == "max" || a.key == "step" || a.key == "disabled" || a.key == "required") {
			continue
		}
		if chartTag(e.tag) && (a.key == "name" || a.key == "data" || a.key == "editMode" || a.key == "chartType" || a.key == "format" || a.key == "theme" ||
			a.key == "x" || a.key == "y" || a.key == "width" || a.key == "height" || a.key == "overflow" ||
			a.key == "zoom" || a.key == "pointWidth" || a.key == "pointSize" || a.key == "bins" || a.key == "min" || a.key == "max" ||
			a.key == "tableView" || a.key == "tableFilters" || a.key == "tableAddRow" || a.key == "tableDeleteRow" || a.key == "tableSteppers" || a.key == "tableContextMenu" || a.key == "tableSelection" || a.key == "criticalPath" || a.key == "holidays" || a.key == "weekend" || a.key == "timezone" || a.key == "showNonWorkingDays" || a.key == "progressColumn" || a.key == "stateHints" || a.key == "stateHintBg" || a.key == "stateHintOpacity" || a.key == "stateHintTextColor" || a.key == "stateHintTextAlign" || a.key == "stateHintBorderColor" || a.key == "stateHintFont" || a.key == "stateHintFontSize" || a.key == "stateHintBorderWidth" || a.key == "stateHintPadding" || a.key == "stateHintRadius" ||
			a.key == "colorPicker" || a.key == "settingsButton" || a.key == "showZoom" || a.key == "showAdd" || a.key == "showHistory" || a.key == "onTaskChange" || a.key == "onDataChange" || a.key == "onCalendarChange" || a.key == "onResourceChange" || a.key == "onCommit" ||
			a.key == "stateRouting" || a.key == "stateTagStyle" || a.key == "stateTagLink" || a.key == "stateLinkTagStyle" || a.key == "stateLinkTagLink" || a.key == "stateShapeTagStyle" || a.key == "stateShapeTagLink" || a.key == "stateTagFont" || a.key == "stateTagFontSize" || a.key == "stateShapeTagFont" || a.key == "stateShapeTagFontSize" || a.key == "stateLinkTagFont" || a.key == "stateLinkTagFontSize" || a.key == "stateRouteAlgorithm" || a.key == "stateLinkArcRadius" || a.key == "stateGrid" || a.key == "stateGridGap" || a.key == "stateGridColor" ||
			a.key == "showGrid" || a.key == "showLegend" || a.key == "showValues" ||
			a.key == "showResources" || a.key == "showProgress" ||
			a.key == "showSlack" || a.key == "showCosts" || a.key == "statusColors" || a.key == "ganttScale" || a.key == "autoSchedule" ||
			a.key == "linkColor" || a.key == "linkWidth" || a.key == "linkGap" || a.key == "linkStyle" || a.key == "arrowType" || a.key == "arrowSize" || a.key == "barBorderColor") {
			continue
		}
		if e.tag == "tree" && (a.key == "value" || a.key == "editMode" || a.key == "editDepth" ||
			a.key == "name" || a.key == "renameKey" || a.key == "renameValue" || a.key == "renameIcon" || a.key == "showRenameIcon" ||
			a.key == "guides" || a.key == "truncate" || a.key == "checkbox" || a.key == "counts" || a.key == "rowHeight" ||
			a.key == "multiSelect" || a.key == "selection" ||
			a.key == "search" || a.key == "drag" || a.key == "contextMenu" || a.key == "columns" || a.key == "loadKey" ||
			a.key == "virtual" || a.key == "columnLabels" || a.key == "columnWidths" || a.key == "sort" || a.key == "persistSelection" ||
			a.key == "columnsEnabled" || a.key == "tooltips" || a.key == "fontStep" || a.key == "onCheck" || a.key == "onLoad" || a.key == "onMove" || a.key == "onContextMenu") {
			continue
		}
		if e.tag == "treeitem" && (a.key == "key" || a.key == "open" || a.key == "rename" || a.key == "renameIcon" || a.key == "showRenameIcon" ||
			a.key == "icon" || a.key == "iconOpen" || a.key == "badge" || a.key == "badgeType" || a.key == "meta" || a.key == "checked" ||
			a.key == "lazy" || a.key == "size" || a.key == "date" || a.key == "status" || a.key == "cells") {
			continue
		}
		if e.tag == "tabs" && (a.key == "position" || a.key == "variant") {
			continue
		}
		if e.tag == "dropdown" && a.key == "position" {
			continue
		}
		if e.tag == "popover" && a.key == "position" {
			continue
		}
		if e.tag == "drawer" && (a.key == "open" || a.key == "side" || a.key == "width" || a.key == "onClose") {
			continue
		}
		if e.tag == "splitter" && (a.key == "direction" || a.key == "initial" || a.key == "min") {
			continue
		}
		if e.tag == "dropzone" && (a.key == "accept" || a.key == "multiple" || a.key == "directory" || a.key == "hint") {
			continue
		}
		if e.tag == "timelineitem" && (a.key == "date" || a.key == "status") {
			continue
		}
		if e.tag == "emptystate" && (a.key == "icon" || a.key == "hint" || a.key == "compact") {
			continue
		}
		if e.tag == "skeleton" && (a.key == "lines" || a.key == "width" || a.key == "height" || a.key == "circle" || a.key == "animated" || a.key == "gap") {
			continue
		}
		if e.tag == "carousel" && (a.key == "value" || a.key == "autoplay" || a.key == "interval" || a.key == "loop" || a.key == "controls" || a.key == "indicators" || a.key == "height") {
			continue
		}
		if e.tag == "slide" && a.key == "label" {
			continue
		}
		if e.tag == "stepper" && (a.key == "value" || a.key == "clickable" || a.key == "vertical") {
			continue
		}
		if e.tag == "step" && (a.key == "description" || a.key == "icon") {
			continue
		}
		if e.tag == "commandbar" && (a.key == "open" || a.key == "hotkey" || a.key == "placeholder") {
			continue
		}
		if e.tag == "command" && a.key == "shortcut" {
			continue
		}
		if e.tag == "dropdown" && (a.key == "placement" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "popover" && (a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "combobox" && (a.key == "placement" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "drawer" && (a.key == "duration" || a.key == "easing") {
			continue
		}
		if e.tag == "carousel" && (a.key == "duration" || a.key == "easing") {
			continue
		}
		if e.tag == "commandbar" && (a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			continue
		}
		if e.tag == "skeleton" && a.key == "duration" {
			continue
		}
		if (e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio") &&
			(a.key == "shape" || a.key == "theme" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay" || a.key == "stagger") {
			continue
		}
		if e.tag == "dropdownitem" && a.key == "separator" {
			continue
		}
		if e.tag == "dropdownitem" && a.key == "disabled" {
			continue
		}
		if e.tag == "tablelayout" && (a.key == "columns" || a.key == "rows" ||
			a.key == "width" || a.key == "height" ||
			a.key == "columnWidths" || a.key == "rowHeights" ||
			a.key == "gapX" || a.key == "gapY" ||
			a.key == "borderStyle" || a.key == "borderWidth" || a.key == "borderColor" ||
			a.key == "cellMinHeight" || a.key == "responsive" || a.key == "semantic" || a.key == "label" || a.key == "style") {
			continue
		}
		if e.tag == "table" && (a.key == "name" || a.key == "editMode" ||
			a.key == "columns" || a.key == "columnLabels" || a.key == "columnWidths" ||
			a.key == "columnAlign" || a.key == "columnTypes" || a.key == "columnFormats" || a.key == "columnOptions" ||
			a.key == "cellText" || a.key == "sortable" || a.key == "filterable" || a.key == "pageSize" || a.key == "selection" ||
			a.key == "multiSelect" || a.key == "virtual" || a.key == "showAddTable" || a.key == "structureEditable" ||
			a.key == "onCellChange" || a.key == "onRowChange" || a.key == "onTableChange" ||
			a.key == "onRowSelect" || a.key == "onSort" || a.key == "onCellClick" || a.key == "onRowClick") {
			continue
		}
		if a.key == "head" || a.key == "options" || a.key == "link" || a.key == "srcAllow" ||
			a.key == "draggable" || a.key == "directory" || a.key == "dragKey" ||
			a.key == "dropKey" || a.key == "dropFiles" || a.key == "pickFiles" || a.key == "pickMode" {
			continue
		}
		if a.key == "motion" || a.key == "motionKey" || a.key == "motionDuration" || a.key == "motionEasing" || a.key == "motionDelay" || a.key == "motionExit" {
			continue
		}
		if e.tag != "tcell" && (a.key == "colspan" || a.key == "rowspan") {
			continue
		}
		if a.key == "cellPadding" || a.key == "cellBackground" ||
			a.key == "cellBorderStyle" || a.key == "cellBorderWidth" || a.key == "cellBorderColor" {
			continue
		}
		if e.tag != "file" && findAttr(e, "pickFiles") != nil &&
			(a.key == "accept" || a.key == "multiple" || a.key == "directory") {
			continue
		}
		if e.tag == "iframe" && a.key == "src" && iframeSrc != "" {
			sb.WriteString(" src=\"${__sui_esc(" + iframeSrc + ")}\"")
			continue
		}
		if e.tag == "toast" && (a.key == "position" || a.key == "offset" ||
			a.key == "hAlign" || a.key == "vAlign" || a.key == "hOffset" || a.key == "vOffset" ||
			a.key == "duration" || a.key == "max" || a.key == "icons") {
			continue
		}
		if a.key == "class" {
			// handled up-front so it composes with any built-in base class
			continue
		}
		if e.tag == "textarea" && a.key == "value" {
			// <textarea> has no value attribute: the value is its text content
			// (emitted by elemJs).
			continue
		}
		if a.key == "required" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
				sb.WriteString(" required")
			}
			continue
		}
		if a.key == "label" {
			// label is for editor - skip or could wrap, but keep simple
			continue
		}
		sb.WriteByte(' ')
		if a.key == "inputtype" {
			sb.WriteString("type")
		} else if a.key == "checked" && (e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio") {
			// "checked" is a boolean HTML attribute: its mere presence means
			// ON, so checked="false" would still render as checked and the
			// control would never visibly toggle. Emit the attribute only when
			// the expression is truthy (this runs inside a template literal).
			if a.hasExpr && a.expr != nil {
				sb.WriteString("${(")
				sb.WriteString(ExprToJs(a.expr))
				sb.WriteString(") ? \"checked\" : \"\"}")
			} else {
				var raw strings.Builder
				for _, p := range a.text {
					if p.kind == "text" {
						raw.WriteString(p.text)
					}
				}
				v := strings.TrimSpace(strings.ToLower(raw.String()))
				if v != "" && v != "false" && v != "0" {
					sb.WriteString("checked")
				}
			}
			continue
		} else {
			sb.WriteString(a.key)
		}
		sb.WriteByte('=')
		sb.WriteByte('"')
		urlAttr := a.key == "href" || a.key == "src"
		sb.WriteString(htmlAttrValue(&a, urlAttr))
		sb.WriteByte('"')
	}
	if e.tag == "tablelayout" {
		sb.WriteString(" style=\"")
		sb.WriteString(layoutTableStyleAttrValue(e))
		sb.WriteString("\"")
		grid := layoutTableGrid(e)
		sb.WriteString(" data-sui-layout-cols=\"" + strconv.Itoa(grid.cols) + "\" data-sui-layout-rows=\"" + strconv.Itoa(grid.rows) + "\"")
	}
}

// writeContainerClass emits the class attribute for container tags (row/col/
// spacer) whose base class is built in. The custom class may be a static
// string or a reactive expression (e.g. class: active ? "on" : ""), in which
// case it is interpolated so the element updates when the state changes.
func writeContainerClass(sb *strings.Builder, e *Element, ctx *JsCtx, base string) {
	a := findAttr(e, "class")
	if a == nil {
		sb.WriteString(" class=\"" + base + "\"")
		return
	}
	if a.hasExpr {
		sb.WriteString(" class=\"" + base + " ${" + classExprJs(ExprToJs(a.expr)) + "}\"")
		return
	}
	if partsContent(a.text) == "" {
		sb.WriteString(" class=\"" + base + "\"")
		return
	}
	sb.WriteString(" class=\"" + base + " " + classPartsJs(a.text) + "\"")
}

// writePlainClass emits the class attribute for a plain element from a static
// string or a reactive expression.
func writePlainClass(sb *strings.Builder, e *Element) {
	a := findAttr(e, "class")
	if a == nil {
		return
	}
	if a.hasExpr {
		sb.WriteString(" class=\"${" + classExprJs(ExprToJs(a.expr)) + "}\"")
		return
	}
	if partsContent(a.text) != "" {
		sb.WriteString(" class=\"" + classPartsJs(a.text) + "\"")
	}
}

func classSuffix(e *Element) string {
	a := findAttr(e, "class")
	if a == nil || a.hasExpr {
		return ""
	}
	if partsContent(a.text) == "" {
		return ""
	}
	return " " + classPartsJs(a.text)
}

// elemLinkOpen returns the `<a href="...">` wrapper when the element carries a
// `link` attribute (renders this element as a hyperlink), or "" otherwise.
func elemLinkOpen(e *Element) string {
	a := findAttr(e, "link")
	if a == nil {
		return ""
	}
	if a.hasExpr {
		return "<a href=\"${__sui_esc(__sui_url(" + ExprToJs(a.expr) + "))}\">"
	}
	return "<a href=\"${__sui_esc(__sui_url(`" + partsContent(a.text) + "`))}\">"
}

// wrapLabelOpen builds the <label> wrapper that hosts a checkbox/switch/radio
// widget: it carries the picker data-sui-id and mirrors the element's class
// (static or reactive) so a user class styles the whole widget, input and
// label text alike.
func wrapLabelOpen(e *Element, idx int) string {
	base := "sui-checkbox-wrap"
	if e.tag == "switch" {
		base = "sui-switch-wrap"
	} else if e.tag == "radio" {
		base = "sui-radio-wrap"
	}
	var sb strings.Builder
	sb.WriteString("<label")
	if idx >= 0 {
		sb.WriteString(" data-sui-id=\"" + strconv.Itoa(idx) + "\"")
	}
	sb.WriteString(" class=\"" + base)
	if a := findAttr(e, "class"); a != nil {
		if a.hasExpr && a.expr != nil {
			sb.WriteString(" ${" + classExprJs(ExprToJs(a.expr)) + "}")
		} else if partsContent(a.text) != "" {
			sb.WriteString(" " + classPartsJs(a.text))
		}
	}
	sb.WriteString("\">")
	return sb.String()
}

// errorBoundaryElemJs wraps the boundary content in a try/catch at render
// time. Child component and expression errors show the fallback text instead
// of breaking the whole page.
func errorBoundaryElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	if !strings.Contains(open, "sui-error-boundary") {
		open = addOpenTagClass(open, "sui-error-boundary")
	}
	fallback := "`Something went wrong.`"
	if a := findAttr(e, "fallback"); a != nil {
		if a.hasExpr && a.expr != nil {
			fallback = "String(" + ExprToJs(a.expr) + ")"
		} else if len(a.text) > 0 {
			fallback = templateParts(a.text)
		}
	}
	inner := viewNodesJs(childPtrs(e.children), g, ctx)
	return "`" + open + "${__sui_boundary(() => (" + inner + "), " + fallback + ")}</div>`"
}

func elemJs(e *Element, g *ViewGen, ctx *JsCtx) string {
	if e == nil {
		return "''"
	}
	if e.tag == "dropdownitem" && datepickerBool(e, "separator") {
		return "`<div class=\"sui-dropdown-sep\" role=\"separator\"></div>`"
	}
	// The `children` slot marker renders the caller-supplied block.
	if e.tag == "children" {
		return "`${children}`"
	}
	// A component invocation: the tag names a parameterized view.
	if genViewParams[e.tag] != nil {
		params := genViewParams[e.tag]
		var parts []string
		// Per-instance state: a stable per-site key into the runtime store.
		if genViewInst[e.tag] {
			siteKey := strconv.Itoa(genSiteId)
			genSiteId = genSiteId + 1
			parts = append(parts, "__inst: (__SUI_inst[\""+siteKey+"\"] ||= {})")
			parts = append(parts, "__key: \""+siteKey+"\"")
		}
		// The caller's block becomes the children chunk, generated in the
		// caller's scope (states/loop vars and wired handlers live there).
		if len(e.children) > 0 {
			parts = append(parts, "children: "+viewNodesJs(childPtrs(e.children), g, ctx))
		}
		for i := 0; i < len(e.attrs); i++ {
			a := e.attrs[i]
			paramType := ""
			for j := 0; j < len(params); j++ {
				if params[j].name == a.key {
					paramType = params[j].typename
					break
				}
			}
			val := "''"
			if paramType == "fn" && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
				// Handlers are passed by name; the chunk interpolates the name
				// into an inline onclick attribute at runtime.
				val = jsString(a.expr.name)
			} else if a.hasExpr {
				val = ExprToJs(a.expr)
			} else {
				val = templateParts(a.text)
			}
			parts = append(parts, a.key+": "+val)
		}
		return " __html_" + e.tag + "({" + strings.Join(parts, ", ") + "})"
	}
	tag := elemHtmlTag(e.tag)
	if e.tag == "tcell" && isHeaderCell(e) {
		tag = "th"
	}
	interactive := isInteractiveElem(e, ctx)
	idx := -1
	if !genCompBody {
		// Every element in a root view gets a unique DFS-order id (used for
		// query wiring AND for the design tool's click-to-select mapping).
		idx = g.count
		g.count = g.count + 1
	}
	addElemRefs(e, g)
	if interactive && !genCompBody && e.tag != "richtext" {
		collectWiring(e, idx, g, ctx)
	}
	if e.tag == "errorboundary" {
		return errorBoundaryElemJs(e, g, ctx, idx)
	}
	if e.tag == "spinner" && spinnerImageMode(e) {
		return spinnerImageElemJs(e, idx, ctx)
	}
	if e.tag == "iframe" {
		if src := findAttr(e, "src"); src != nil {
			allow := findAttr(e, "srcAllow")
			if attrIsDynamic(src) || allow != nil {
				srcJs := "''"
				if src.hasExpr && src.expr != nil {
					srcJs = ExprToJs(src.expr)
				} else if lit, ok := staticPartsLiteral(src.text); ok {
					srcJs = jsString(lit)
				}
				allowJs, required := "''", "false"
				if allow != nil {
					if lit, ok := staticPartsLiteral(allow.text); ok {
						allowJs = jsString(lit)
						required = "true"
					}
				}
				guard := "__sui_iframe_src(" + srcJs + ", " + allowJs + ", " + required + ")"
				open := openTagHtmlWithSrc(e, tag, interactive, idx, ctx, "__g[0]")
				return "`${(function(){var __g=" + guard +
					";if(__g[1])return __sui_iframe_error(__g[1]);return `" + open + "</iframe>`;})()}`"
			}
		}
	}
	if e.tag == "modal" {
		return modalElemJs(e, g, ctx, idx)
	}
	if e.tag == "combobox" {
		return comboboxElemJs(e, g, ctx, idx)
	}
	if e.tag == "datepicker" {
		return datepickerElemJs(e, g, ctx, idx)
	}
	if e.tag == "tree" {
		return treeElemJs(e, g, ctx, idx)
	}
	if e.tag == "richtext" {
		return richTextElemJs(e, g, ctx, idx)
	}
	if e.tag == "tablelayout" {
		return tableLayoutElemJs(e, g, ctx, idx)
	}
	if chartTag(e.tag) {
		return chartElemJs(e, g, ctx, idx)
	}
	if e.tag == "tabs" {
		return tabsElemJs(e, g, ctx, idx)
	}
	if e.tag == "accordion" {
		return accordionElemJs(e, g, ctx, idx)
	}
	if e.tag == "menu" {
		return menuElemJs(e, g, ctx, idx)
	}
	if e.tag == "dropdown" {
		return menuElemJs(e, g, ctx, idx)
	}
	if e.tag == "popover" {
		return popoverElemJs(e, g, ctx, idx)
	}
	if e.tag == "drawer" {
		return drawerElemJs(e, g, ctx, idx)
	}
	if e.tag == "splitter" {
		return splitterElemJs(e, g, ctx, idx)
	}
	if e.tag == "dropzone" {
		return dropzoneElemJs(e, g, ctx, idx)
	}
	if e.tag == "timeline" {
		return timelineElemJs(e, g, ctx, idx)
	}
	if e.tag == "emptystate" {
		return emptyStateElemJs(e, g, ctx, idx)
	}
	if e.tag == "skeleton" {
		return skeletonElemJs(e, g, ctx, idx)
	}
	if e.tag == "carousel" {
		return carouselElemJs(e, g, ctx, idx)
	}
	if e.tag == "stepper" {
		return stepperElemJs(e, g, ctx, idx)
	}
	if e.tag == "commandbar" {
		return commandbarElemJs(e, g, ctx, idx)
	}
	if e.tag == "menubar" {
		open := openTagHtml(e, "nav", false, idx, ctx)
		var kids string
		if len(e.children) > 0 {
			kids = viewNodesJs(childPtrs(e.children), g, ctx)
		} else {
			kids = "''"
		}
		return "`" + open + "` + (" + kids + ") + `</nav>`"
	}
	open := openTagHtml(e, tag, interactive, idx, ctx)
	var inner string
	if e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio" {
		// These controls carry a label: render input + label text inside a
		// wrapper <label>. The picker data-sui-id moves onto the wrapper (so a
		// click on the text selects the element too) and the wiring code below
		// still finds it there — the input's DOM events bubble to the label.
		var sb strings.Builder
		inputOpen := openTagHtml(e, tag, interactive, -1, ctx)
		inputOpen = strings.TrimSuffix(inputOpen, ">") + toggleMotionAttrs(e, idx) + ">"
		sb.WriteString("`" + wrapLabelOpen(e, idx) + "` + (`" + inputOpen + "`")
		if len(e.text) > 0 {
			var tb strings.Builder
			tb.WriteString(htmlTextParts(e.text))
			sb.WriteString(" + (`<span>` + `" + tb.String() + "` + `</span>`)")
		}
		sb.WriteString(") + `</label>`")
		inner = sb.String()
	} else if isVoidTag(e.tag) {
		inner = "`" + open + "`"
	} else if e.tag == "textarea" {
		var sb strings.Builder
		sb.WriteString("`")
		sb.WriteString(open)
		if va := findAttr(e, "value"); va != nil {
			sb.WriteString(htmlAttrValue(va, false))
		}
		sb.WriteString("</textarea>`")
		inner = sb.String()
	} else if len(e.children) > 0 {
		kids := viewNodesJs(childPtrs(e.children), g, ctx)
		inner = "`" + open + "` + (" + kids + ") + `</" + tag + ">`"
	} else {
		var sb strings.Builder
		sb.WriteString("`")
		sb.WriteString(open)
		if e.tag != "spacer" {
			sb.WriteString(htmlTextParts(e.text))
		}
		sb.WriteString("</" + tag + ">`")
		inner = sb.String()
	}
	if lo := elemLinkOpen(e); lo != "" {
		return "`" + lo + "` + (" + inner + ") + `</a>`"
	}
	return inner
}

// spinnerImageElemJs renders a spinner whose `src` points at a GIF or image.
// The wrapper keeps the control class and user attributes; the inner <img>
// receives the guarded source and the label as alternative text.
func spinnerImageElemJs(e *Element, idx int, ctx *JsCtx) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString("<img class=\"sui-spinner-img\"")
	if src := findAttr(e, "src"); src != nil {
		b.WriteString(" src=\"")
		b.WriteString(htmlAttrValueImage(src))
		b.WriteString("\"")
	}
	b.WriteString(" alt=\"")
	if alt := findAttr(e, "alt"); alt != nil {
		b.WriteString(htmlAttrValue(alt, false))
	} else if len(e.text) > 0 {
		b.WriteString(htmlTextParts(e.text))
	}
	b.WriteString("\"></div>`")
	return b.String()
}

// modalElemJs renders a hidden modal with a title, body content and a footer
// of modalbutton children. The runtime opens it by data-sui-modal name.
// tableLayoutElemJs renders a structural HTML table whose cells host the
// element's children. Children flow left-to-right, top-to-bottom in fixed
// columns; wrappers are static markup while every child goes through the
// normal escaped element renderer.
func tableLayoutElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "table", false, idx, ctx)
	grid := layoutTableGrid(e)
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString(layoutColgroupHTML(e, grid.cols))
	b.WriteString("<tbody>")
	for r := 0; r < grid.rows; r++ {
		b.WriteString("<tr")
		b.WriteString(layoutRowStyle(e, r))
		b.WriteString(">")
		for c := 0; c < grid.cols; c++ {
			if !grid.occupied[r][c] {
				b.WriteString("<td class=\"sui-layout-cell sui-layout-cell-empty\"")
				b.WriteString(layoutCellPosAttrs(r, c))
				b.WriteString(layoutEmptyCellAttrs(e))
				b.WriteString("></td>")
				continue
			}
			ci, ok := grid.starts[[2]int{r, c}]
			if !ok {
				continue // covered by a rowspan from an earlier row
			}
			cell := grid.cells[ci]
			b.WriteString("<td class=\"sui-layout-cell")
			b.WriteString(layoutCellClass(cell))
			b.WriteString(layoutCellClassSuffix(cell))
			b.WriteString("\"")
			b.WriteString(layoutCellPosAttrs(r, c))
			b.WriteString(layoutCellAttrs(cell, e, true))
			if cell.node != nil && cell.node.kind == "element" && cell.node.elem != nil {
				writeLayoutCellExtras(&b, cell.node.elem, ctx)
			}
			b.WriteString(">` + (")
			b.WriteString(viewNodesJs(childPtrs(cell.content), g, ctx))
			b.WriteString(") + `</td>")
		}
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table>`")
	return b.String()
}

func modalElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	label := htmlTextParts(e.text)
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString("<div class=\"sui-modal-backdrop\" data-sui-modal-backdrop></div>")
	b.WriteString("<div class=\"sui-modal-panel\" role=\"dialog\" aria-modal=\"true\" tabindex=\"-1\"")
	if label != "" {
		b.WriteString(" aria-label=\"")
		b.WriteString(label)
		b.WriteString("\"")
	}
	b.WriteString(">")
	if label != "" {
		b.WriteString("<div class=\"sui-modal-title\">")
		b.WriteString(label)
		b.WriteString("</div>")
	}
	b.WriteString("<div class=\"sui-modal-body\">")
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "modalbutton" {
			continue
		}
		b.WriteString("` + (")
		b.WriteString(viewNodesJs([]*ViewNode{ch}, g, ctx))
		b.WriteString(") + `")
	}
	b.WriteString("</div><div class=\"sui-modal-actions\">")
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind != "element" || ch.elem == nil || ch.elem.tag != "modalbutton" {
			continue
		}
		btn := ch.elem
		b.WriteString("<button type=\"button\" class=\"sui-modal-button")
		if v := modalVariant(btn); v != "" {
			b.WriteString(" " + v)
		}
		b.WriteString(classSuffix(btn))
		b.WriteString("\"")
		if modalButtonCloses(btn) {
			b.WriteString(" data-sui-modal-dismiss=\"1\"")
		}
		if a := findAttr(btn, "onClick"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
			b.WriteString(" onclick=\"")
			b.WriteString(a.expr.name)
			b.WriteString("()\"")
		}
		b.WriteString(">")
		b.WriteString(htmlTextParts(btn.text))
		b.WriteString("</button>")
	}
	b.WriteString("</div></div></div>`")
	return b.String()
}

// chartAttrValue returns a chart data attribute value. The value can be a
// literal (JSON/XML/CSV text or a CSS size) or a reactive expression.
func chartAttrValue(e *Element, key, def string) string {
	a := findAttr(e, key)
	if a == nil {
		return htmlEscape(def)
	}
	if !a.hasExpr && strings.TrimSpace(partsContent(a.text)) == "" {
		return htmlEscape(def)
	}
	if a.hasExpr && a.expr != nil {
		if a.expr.kind == "int" {
			return strconv.Itoa(a.expr.i)
		}
		if a.expr.kind == "float" {
			return strconv.FormatFloat(a.expr.f, 'g', -1, 64)
		}
	}
	return htmlAttrValue(a, false)
}

func chartFlagValue(e *Element, key string, def bool) string {
	a := findAttr(e, key)
	if a == nil {
		if def {
			return "1"
		}
		return "0"
	}
	if a.hasExpr && a.expr != nil {
		if a.expr.kind == "bool" {
			if a.expr.b {
				return "1"
			}
			return "0"
		}
		return "${((" + ExprToJs(a.expr) + ") ? \"1\" : \"0\")}"
	}
	v := strings.ToLower(strings.Trim(strings.TrimSpace(partsContent(a.text)), "\"'"))
	if v == "true" || v == "1" || v == "yes" || v == "on" {
		return "1"
	}
	if v == "false" || v == "0" || v == "no" || v == "off" {
		return "0"
	}
	if def {
		return "1"
	}
	return "0"
}

func chartNameValue(e *Element) string {
	if a := findAttr(e, "name"); a != nil && !a.hasExpr {
		if v := strings.TrimSpace(partsContent(a.text)); v != "" {
			return v
		}
	}
	if v := strings.TrimSpace(partsContent(e.text)); v != "" {
		return v
	}
	return "chart"
}

func chartDataAttrs(e *Element) string {
	var b strings.Builder
	label := strings.TrimSpace(partsContent(e.text))
	if label == "" {
		if e.tag == "graph" {
			label = "Graph"
		} else {
			label = "Chart"
		}
	}
	b.WriteString(" data-sui-chart=\"1\" role=\"group\" aria-label=\"")
	b.WriteString(htmlEscape(label))
	b.WriteString("\"")
	b.WriteString(" data-sui-chart-name=\"")
	b.WriteString(htmlEscape(chartNameValue(e)))
	b.WriteString("\" data-sui-chart-title=\"")
	b.WriteString(htmlEscape(partsContent(e.text)))
	b.WriteString("\" data-sui-chart-type=\"")
	if e.tag == "graph" {
		b.WriteString("state")
	} else {
		b.WriteString(chartAttrValue(e, "chartType", "bar"))
	}
	b.WriteString("\" data-sui-chart-theme=\"")
	b.WriteString(chartAttrValue(e, "theme", "dark"))
	b.WriteString("\" data-sui-chart-edit-mode=\"")
	b.WriteString(chartAttrValue(e, "editMode", ""))
	b.WriteString("\" data-sui-chart-data=\"")
	b.WriteString(chartAttrValue(e, "data", ""))
	b.WriteString("\" data-sui-chart-format=\"")
	b.WriteString(chartAttrValue(e, "format", "auto"))
	b.WriteString("\" data-sui-chart-x=\"")
	b.WriteString(chartAttrValue(e, "x", ""))
	b.WriteString("\" data-sui-chart-y=\"")
	b.WriteString(chartAttrValue(e, "y", ""))
	b.WriteString("\"")
	// Width/height are the standard component size in the designer. Keep the
	// explicit chart attributes only when the source declares them.
	if findAttr(e, "width") != nil {
		b.WriteString(" data-sui-chart-width=\"")
		b.WriteString(chartAttrValue(e, "width", "640px"))
		b.WriteString("\"")
	}
	if findAttr(e, "height") != nil {
		b.WriteString(" data-sui-chart-height=\"")
		b.WriteString(chartAttrValue(e, "height", "320px"))
		b.WriteString("\"")
	}
	b.WriteString(" data-sui-chart-overflow=\"")
	b.WriteString(chartAttrValue(e, "overflow", "scroll"))
	b.WriteString("\" data-sui-chart-zoom=\"")
	b.WriteString(chartAttrValue(e, "zoom", "1"))
	b.WriteString("\" data-sui-chart-point-width=\"")
	b.WriteString(chartAttrValue(e, "pointWidth", "48"))
	b.WriteString("\" data-sui-chart-point-size=\"")
	b.WriteString(chartAttrValue(e, "pointSize", "3"))
	b.WriteString("\" data-sui-chart-table-view=\"")
	b.WriteString(chartAttrValue(e, "tableView", "chart"))
	b.WriteString("\" data-sui-chart-table-filters=\"")
	b.WriteString(chartFlagValue(e, "tableFilters", true))
	b.WriteString("\" data-sui-chart-table-add-row=\"")
	b.WriteString(chartFlagValue(e, "tableAddRow", false))
	b.WriteString("\" data-sui-chart-table-delete-row=\"")
	b.WriteString(chartFlagValue(e, "tableDeleteRow", false))
	b.WriteString("\" data-sui-chart-table-steppers=\"")
	b.WriteString(chartFlagValue(e, "tableSteppers", false))
	b.WriteString("\" data-sui-chart-table-context-menu=\"")
	b.WriteString(chartFlagValue(e, "tableContextMenu", false))
	b.WriteString("\" data-sui-chart-table-selection=\"")
	b.WriteString(chartAttrValue(e, "tableSelection", "row"))
	b.WriteString("\" data-sui-chart-critical-path=\"")
	b.WriteString(chartFlagValue(e, "criticalPath", false))
	b.WriteString("\" data-sui-chart-holidays=\"")
	b.WriteString(chartAttrValue(e, "holidays", ""))
	b.WriteString("\" data-sui-chart-weekend=\"")
	b.WriteString(chartAttrValue(e, "weekend", ""))
	b.WriteString("\" data-sui-chart-timezone=\"")
	b.WriteString(chartAttrValue(e, "timezone", ""))
	b.WriteString("\" data-sui-chart-show-non-working-days=\"")
	b.WriteString(chartFlagValue(e, "showNonWorkingDays", true))
	b.WriteString("\" data-sui-chart-progress-column=\"")
	b.WriteString(chartFlagValue(e, "progressColumn", false))
	b.WriteString("\" data-sui-chart-bins=\"")
	b.WriteString(chartAttrValue(e, "bins", "10"))
	b.WriteString("\" data-sui-chart-min=\"")
	b.WriteString(chartAttrValue(e, "min", ""))
	b.WriteString("\" data-sui-chart-max=\"")
	b.WriteString(chartAttrValue(e, "max", ""))
	b.WriteString("\" data-sui-chart-color-picker=\"")
	b.WriteString(chartFlagValue(e, "colorPicker", true))
	b.WriteString("\" data-sui-chart-settings-button=\"")
	b.WriteString(chartFlagValue(e, "settingsButton", true))
	b.WriteString("\" data-sui-chart-show-zoom=\"")
	b.WriteString(chartFlagValue(e, "showZoom", true))
	b.WriteString("\" data-sui-chart-show-add=\"")
	b.WriteString(chartFlagValue(e, "showAdd", true))
	b.WriteString("\" data-sui-chart-show-history=\"")
	b.WriteString(chartFlagValue(e, "showHistory", true))
	b.WriteString("\" data-sui-chart-show-resources=\"")
	b.WriteString(chartFlagValue(e, "showResources", false))
	b.WriteString("\" data-sui-chart-show-progress=\"")
	b.WriteString(chartFlagValue(e, "showProgress", false))
	b.WriteString("\" data-sui-chart-show-slack=\"")
	b.WriteString(chartFlagValue(e, "showSlack", false))
	b.WriteString("\" data-sui-chart-show-costs=\"")
	b.WriteString(chartFlagValue(e, "showCosts", false))
	b.WriteString("\" data-sui-chart-status-colors=\"")
	b.WriteString(chartFlagValue(e, "statusColors", false))
	b.WriteString("\" data-sui-chart-auto-schedule=\"")
	b.WriteString(chartFlagValue(e, "autoSchedule", false))
	b.WriteString("\" data-sui-chart-gantt-scale=\"")
	b.WriteString(chartAttrValue(e, "ganttScale", "auto"))
	b.WriteString("\" data-sui-chart-link-color=\"")
	b.WriteString(chartAttrValue(e, "linkColor", "#94a3b8"))
	b.WriteString("\" data-sui-chart-state-shape-border-color=\"")
	b.WriteString(chartAttrValue(e, "stateShapeBorderColor", ""))
	b.WriteString("\" data-sui-chart-state-selection-color=\"")
	b.WriteString(chartAttrValue(e, "stateSelectionColor", ""))
	b.WriteString("\" data-sui-chart-link-width=\"")
	b.WriteString(chartAttrValue(e, "linkWidth", "1"))
	b.WriteString("\" data-sui-chart-link-gap=\"")
	b.WriteString(chartAttrValue(e, "linkGap", "4"))
	b.WriteString("\" data-sui-chart-bar-border-color=\"")
	b.WriteString(chartAttrValue(e, "barBorderColor", ""))
	b.WriteString("\" data-sui-chart-link-style=\"")
	b.WriteString(chartAttrValue(e, "linkStyle", "auto"))
	b.WriteString("\" data-sui-chart-arrow-type=\"")
	b.WriteString(chartAttrValue(e, "arrowType", "triangle"))
	b.WriteString("\" data-sui-chart-arrow-size=\"")
	b.WriteString(chartAttrValue(e, "arrowSize", ""))
	b.WriteString("\" data-sui-chart-state-start-arrow=\"")
	b.WriteString(chartFlagValue(e, "stateStartArrow", false))
	b.WriteString("\" data-sui-chart-state-routing=\"")
	b.WriteString(chartAttrValue(e, "stateRouting", "orthogonal"))
	b.WriteString("\" data-sui-chart-state-tag-style=\"")
	b.WriteString(chartAttrValue(e, "stateTagStyle", "pill"))
	b.WriteString("\" data-sui-chart-state-tag-link=\"")
	b.WriteString(chartAttrValue(e, "stateTagLink", "line"))
	b.WriteString("\" data-sui-chart-state-link-tag-style=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagStyle", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-link=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagLink", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-style=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagStyle", "box"))
	b.WriteString("\" data-sui-chart-state-shape-tag-link=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagLink", "line"))
	b.WriteString("\" data-sui-chart-state-shape-tag-kind=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagKind", "external"))
	b.WriteString("\" data-sui-chart-state-shape-tag-align=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagAlign", "line"))
	b.WriteString("\" data-sui-chart-state-shape-tag-rotation=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagRotation", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-arrow-size=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagArrowSize", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-kind=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagKind", "external"))
	b.WriteString("\" data-sui-chart-state-link-tag-align=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagAlign", "line"))
	b.WriteString("\" data-sui-chart-state-link-tag-rotation=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagRotation", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-arrow-size=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagArrowSize", ""))
	b.WriteString("\" data-sui-chart-state-tag-font=\"")
	b.WriteString(chartAttrValue(e, "stateTagFont", ""))
	b.WriteString("\" data-sui-chart-state-tag-font-size=\"")
	b.WriteString(chartAttrValue(e, "stateTagFontSize", "11"))
	b.WriteString("\" data-sui-chart-state-shape-tag-font=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagFont", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-font-size=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagFontSize", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-font=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagFont", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-font-size=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagFontSize", ""))
	b.WriteString("\" data-sui-chart-state-tag-font-weight=\"")
	b.WriteString(chartAttrValue(e, "stateTagFontWeight", "400"))
	b.WriteString("\" data-sui-chart-state-tag-font-style=\"")
	b.WriteString(chartAttrValue(e, "stateTagFontStyle", "normal"))
	b.WriteString("\" data-sui-chart-state-tag-text-decoration=\"")
	b.WriteString(chartAttrValue(e, "stateTagTextDecoration", "none"))
	b.WriteString("\" data-sui-chart-state-tag-text-align=\"")
	b.WriteString(chartAttrValue(e, "stateTagTextAlign", "center"))
	b.WriteString("\" data-sui-chart-state-tag-letter-spacing=\"")
	b.WriteString(chartAttrValue(e, "stateTagLetterSpacing", ""))
	b.WriteString("\" data-sui-chart-state-tag-line-spacing=\"")
	b.WriteString(chartAttrValue(e, "stateTagLineSpacing", ""))
	b.WriteString("\" data-sui-chart-state-tag-text-transform=\"")
	b.WriteString(chartAttrValue(e, "stateTagTextTransform", "none"))
	b.WriteString("\" data-sui-chart-state-tag-text-shadow=\"")
	b.WriteString(chartAttrValue(e, "stateTagTextShadow", "none"))
	b.WriteString("\" data-sui-chart-state-tag-text-shadow-color=\"")
	b.WriteString(chartAttrValue(e, "stateTagTextShadowColor", "#000000"))
	b.WriteString("\" data-sui-chart-state-shape-tag-font-weight=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagFontWeight", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-font-style=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagFontStyle", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-text-decoration=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagTextDecoration", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-text-align=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagTextAlign", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-letter-spacing=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagLetterSpacing", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-line-spacing=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagLineSpacing", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-text-transform=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagTextTransform", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-text-shadow=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagTextShadow", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-text-shadow-color=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagTextShadowColor", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-font-weight=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagFontWeight", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-font-style=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagFontStyle", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-text-decoration=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagTextDecoration", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-text-align=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagTextAlign", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-letter-spacing=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagLetterSpacing", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-line-spacing=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagLineSpacing", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-text-transform=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagTextTransform", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-text-shadow=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagTextShadow", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-text-shadow-color=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagTextShadowColor", ""))
	b.WriteString("\" data-sui-chart-state-tag-text-color=\"")
	b.WriteString(chartAttrValue(e, "stateTagTextColor", ""))
	b.WriteString("\" data-sui-chart-state-tag-box-color=\"")
	b.WriteString(chartAttrValue(e, "stateTagBoxColor", ""))
	b.WriteString("\" data-sui-chart-state-tag-border-color=\"")
	b.WriteString(chartAttrValue(e, "stateTagBorderColor", ""))
	b.WriteString("\" data-sui-chart-state-tag-border-width=\"")
	b.WriteString(chartAttrValue(e, "stateTagBorderWidth", ""))
	b.WriteString("\" data-sui-chart-state-tag-border-style=\"")
	b.WriteString(chartAttrValue(e, "stateTagBorderStyle", "solid"))
	b.WriteString("\" data-sui-chart-state-tag-padding=\"")
	b.WriteString(chartAttrValue(e, "stateTagPadding", ""))
	b.WriteString("\" data-sui-chart-state-tag-leader-color=\"")
	b.WriteString(chartAttrValue(e, "stateTagLeaderColor", ""))
	b.WriteString("\" data-sui-chart-state-tag-arrow-shape=\"")
	b.WriteString(chartAttrValue(e, "stateTagArrowShape", "triangle"))
	b.WriteString("\" data-sui-chart-state-shape-tag-text-color=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagTextColor", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-box-color=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagBoxColor", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-border-color=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagBorderColor", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-border-width=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagBorderWidth", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-border-style=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagBorderStyle", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-padding=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagPadding", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-leader-color=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagLeaderColor", ""))
	b.WriteString("\" data-sui-chart-state-shape-tag-arrow-shape=\"")
	b.WriteString(chartAttrValue(e, "stateShapeTagArrowShape", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-text-color=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagTextColor", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-box-color=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagBoxColor", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-border-color=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagBorderColor", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-border-width=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagBorderWidth", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-border-style=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagBorderStyle", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-padding=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagPadding", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-leader-color=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagLeaderColor", ""))
	b.WriteString("\" data-sui-chart-state-link-tag-arrow-shape=\"")
	b.WriteString(chartAttrValue(e, "stateLinkTagArrowShape", ""))
	b.WriteString("\" data-sui-chart-state-link-arc-radius=\"")
	b.WriteString(chartAttrValue(e, "stateLinkArcRadius", "0"))
	b.WriteString("\" data-sui-chart-state-grid=\"")
	b.WriteString(chartFlagValue(e, "stateGrid", true))
	b.WriteString("\" data-sui-chart-state-grid-gap=\"")
	b.WriteString(chartAttrValue(e, "stateGridGap", "24"))
	b.WriteString("\" data-sui-chart-state-grid-color=\"")
	b.WriteString(chartAttrValue(e, "stateGridColor", ""))
	b.WriteString("\" data-sui-chart-state-hints=\"")
	b.WriteString(chartFlagValue(e, "stateHints", true))
	b.WriteString("\" data-sui-chart-state-hint-bg=\"")
	b.WriteString(chartAttrValue(e, "stateHintBg", ""))
	b.WriteString("\" data-sui-chart-state-hint-opacity=\"")
	b.WriteString(chartAttrValue(e, "stateHintOpacity", ""))
	b.WriteString("\" data-sui-chart-state-hint-text-color=\"")
	b.WriteString(chartAttrValue(e, "stateHintTextColor", ""))
	b.WriteString("\" data-sui-chart-state-hint-text-align=\"")
	b.WriteString(chartAttrValue(e, "stateHintTextAlign", ""))
	b.WriteString("\" data-sui-chart-state-hint-border-color=\"")
	b.WriteString(chartAttrValue(e, "stateHintBorderColor", ""))
	b.WriteString("\" data-sui-chart-state-hint-font=\"")
	b.WriteString(chartAttrValue(e, "stateHintFont", ""))
	b.WriteString("\" data-sui-chart-state-hint-font-size=\"")
	b.WriteString(chartAttrValue(e, "stateHintFontSize", ""))
	b.WriteString("\" data-sui-chart-state-hint-border-width=\"")
	b.WriteString(chartAttrValue(e, "stateHintBorderWidth", ""))
	b.WriteString("\" data-sui-chart-state-hint-padding=\"")
	b.WriteString(chartAttrValue(e, "stateHintPadding", ""))
	b.WriteString("\" data-sui-chart-state-hint-radius=\"")
	b.WriteString(chartAttrValue(e, "stateHintRadius", ""))
	b.WriteString("\" data-sui-chart-state-handles=\"")
	b.WriteString(chartFlagValue(e, "stateHandles", true))
	b.WriteString("\" data-sui-chart-state-tags-movable=\"")
	b.WriteString(chartFlagValue(e, "stateTagsMovable", false))
	b.WriteString("\" data-sui-chart-state-routes-editable=\"")
	b.WriteString(chartFlagValue(e, "stateRoutesEditable", false))
	if a := findAttr(e, "onTaskChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		b.WriteString("\" data-sui-chart-ontaskchange=\"")
		b.WriteString(htmlEscape(a.expr.name))
	}
	if a := findAttr(e, "onCalendarChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		b.WriteString("\" data-sui-chart-oncalendarchange=\"")
		b.WriteString(htmlEscape(a.expr.name))
	}
	if a := findAttr(e, "onDataChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		b.WriteString("\" data-sui-chart-ondatachange=\"")
		b.WriteString(htmlEscape(a.expr.name))
	}
	if a := findAttr(e, "onResourceChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		b.WriteString("\" data-sui-chart-onresourcechange=\"")
		b.WriteString(htmlEscape(a.expr.name))
	}
	if a := findAttr(e, "onCommit"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		b.WriteString("\" data-sui-chart-oncommit=\"")
		b.WriteString(htmlEscape(a.expr.name))
	}
	b.WriteString("\" data-sui-chart-grid=\"")
	b.WriteString(chartFlagValue(e, "showGrid", true))
	b.WriteString("\" data-sui-chart-legend=\"")
	b.WriteString(chartFlagValue(e, "showLegend", true))
	b.WriteString("\" data-sui-chart-values=\"")
	b.WriteString(chartFlagValue(e, "showValues", false))
	b.WriteString("\"")
	return b.String()
}

// chartElemJs renders the chart shell. The chart runtime owns parsing and SVG
// generation; keeping the shell semantic makes it usable in the designer and
// in static builds (where the SVG is empty until the runtime loads).
func chartElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	clean := *e
	var keep []Attr
	for _, a := range e.attrs {
		switch a.key {
		case "class", "style", "id", "tooltip", "motion", "motionKey", "motionDuration", "motionEasing", "motionDelay", "motionExit", "motionGraph":
			keep = append(keep, a)
		}
	}
	clean.attrs = keep
	open := openTagHtml(&clean, "div", false, idx, ctx)
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">") + chartDataAttrs(e) + ">"
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString(`<div class="sui-chart-head"><div class="sui-chart-title">`)
	b.WriteString(htmlTextParts(e.text))
	b.WriteString(`</div>`)
	if chartFlagValue(e, "settingsButton", true) != "0" {
		b.WriteString(`<button type="button" class="sui-chart-menu-btn" aria-label="Chart settings" title="Chart settings">&#9881;</button>`)
	}
	b.WriteString(`<div class="sui-chart-zoom" hidden><button type="button" class="sui-chart-zoom-btn" data-zoom="out" aria-label="Zoom out" title="Zoom out">`)
	b.WriteString("\u2212")
	b.WriteString(`</button><span class="sui-chart-zoom-label">100%</span><button type="button" class="sui-chart-zoom-btn" data-zoom="in" aria-label="Zoom in" title="Zoom in">+</button><button type="button" class="sui-chart-zoom-reset" data-zoom="reset" title="Reset zoom">Reset</button><button type="button" class="sui-chart-zoom-fit" data-zoom="fit" aria-label="Fit to width" title="Fit to width">Fit width</button><button type="button" class="sui-chart-zoom-fit sui-chart-zoom-fit-all" data-zoom="fitall" aria-label="Fit all graph elements" title="Fit all graph elements">Fit all</button></div></div>`)
	b.WriteString(`<div class="sui-chart-viewport"><div class="sui-chart-stage"><svg class="sui-chart-svg" role="img"></svg></div></div>`)
	b.WriteString(`<div class="sui-chart-legend"></div><div class="sui-chart-splitter" role="separator" tabindex="0" aria-orientation="horizontal" aria-valuemin="80" aria-valuemax="320" aria-valuenow="180" aria-valuetext="180 pixels" aria-label="Resize chart and table" title="Drag, or use the arrow keys, to resize chart and table"></div><div class="sui-chart-table" role="region" aria-label="Chart data table" hidden></div><div class="sui-chart-error" role="alert" hidden></div><div class="sui-chart-tooltip" role="tooltip" aria-live="polite" hidden></div></div>`)
	b.WriteString("`")
	return b.String()
}

// richTextElemJs renders the host shell for the rich text editor. The
// document JSON is carried in a hidden input so it follows normal state
// bindings; the runtime builds the toolbar, menus and editable surface.
func richTextElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	title := strings.TrimSpace(partsContent(e.text))
	name := richTextTextAttr(e, "name")
	if name == "" {
		name = title
	}
	if name == "" {
		name = "richtext"
	}
	if richTextTextAttr(e, "name") == "" {
		name = name + "-" + strconv.Itoa(idx)
	}
	valueJs := "{}"
	if a := findAttr(e, "value"); a != nil {
		valueJs = htmlAttrValue(a, false)
	}
	setter := ""
	if a := findAttr(e, "value"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" && ctxIsState(ctx, a.expr.name) {
		setter = setterName(a.expr.name)
	}
	onChange := richTextHandlerAttr(e, "onChange")
	onInput := richTextHandlerAttr(e, "onInput")
	onReady := richTextHandlerAttr(e, "onReady")
	onSelection := richTextHandlerAttr(e, "onSelectionChange")
	toolbarItems := richTextTextAttr(e, "toolbarItems")
	template := strings.ToLower(richTextTextAttr(e, "template"))
	height := richTextTextAttr(e, "height")
	placeholder := richTextTextAttr(e, "placeholder")
	theme := strings.ToLower(richTextTextAttr(e, "theme"))
	if theme != "light" && theme != "auto" {
		theme = "dark"
	}
	contentTheme := strings.ToLower(richTextTextAttr(e, "contentTheme"))
	if contentTheme != "dark" {
		contentTheme = "light"
	}
	headerText := richTextTextAttr(e, "header")
	footerText := richTextTextAttr(e, "footer")
	clean := *e
	var keep []Attr
	for _, a := range e.attrs {
		switch a.key {
		case "class", "style", "id", "tooltip", "motion", "motionKey", "motionDuration", "motionEasing", "motionDelay", "motionExit", "motionGraph":
			keep = append(keep, a)
		}
	}
	clean.attrs = keep
	open := openTagHtml(&clean, "div", false, idx, ctx)
	if strings.Contains(open, " class=\"") {
		open = strings.Replace(open, " class=\"", " class=\"sui-rich ", 1)
	} else if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">") + " class=\"sui-rich\">"
	}
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">")
		open += " data-sui-rich=\"1\" data-sui-rich-name=\"" + htmlEscape(name) + "\""
		open += " data-sui-rich-setter=\"" + htmlEscape(setter) + "\""
		open += " data-sui-rich-onchange=\"" + htmlEscape(onChange) + "\""
		open += " data-sui-rich-oninput=\"" + htmlEscape(onInput) + "\""
		open += " data-sui-rich-onready=\"" + htmlEscape(onReady) + "\""
		open += " data-sui-rich-onselection=\"" + htmlEscape(onSelection) + "\""
		open += " data-sui-rich-toolbar-items=\"" + htmlEscape(toolbarItems) + "\""
		open += " data-sui-rich-height=\"" + htmlEscape(height) + "\""
		open += " data-sui-rich-placeholder=\"" + htmlEscape(placeholder) + "\""
		open += " data-sui-rich-theme=\"" + theme + "\""
		open += " data-sui-rich-content=\"" + contentTheme + "\""
		if headerText != "" {
			open += " data-sui-rich-header=\"" + htmlEscape(headerText) + "\""
		}
		if footerText != "" {
			open += " data-sui-rich-footer=\"" + htmlEscape(footerText) + "\""
		}
		if template != "" {
			open += " data-sui-rich-template=\"" + htmlEscape(template) + "\""
		}
		if !richTextBoolAttr(e, "toolbar", true) {
			open += " data-sui-rich-toolbar=\"0\""
		}
		if !richTextBoolAttr(e, "menu", true) {
			open += " data-sui-rich-menu=\"0\""
		}
		if !richTextBoolAttr(e, "contextToolbar", true) {
			open += " data-sui-rich-context=\"0\""
		}
		if !richTextBoolAttr(e, "statusbar", true) {
			open += " data-sui-rich-statusbar=\"0\""
		}
		if richTextBoolAttr(e, "readonly", false) {
			open += " data-sui-rich-readonly=\"1\""
		}
		open += ">"
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	if title != "" {
		b.WriteString("<div class=\"sui-rich-title\">")
		b.WriteString(htmlTextParts(e.text))
		b.WriteString("</div>")
	}
	b.WriteString("<input type=\"hidden\" class=\"sui-rich-value\" value=\"")
	b.WriteString(valueJs)
	b.WriteString("\">")
	b.WriteString("<div class=\"sui-rich-menu\" role=\"menubar\" aria-label=\"Document menu\"></div>")
	b.WriteString("<div class=\"sui-rich-toolbar\" role=\"toolbar\" aria-label=\"Text formatting\"></div>")
	b.WriteString("<div class=\"sui-rich-tablebar\" role=\"toolbar\" aria-label=\"Table actions\" hidden></div>")
	b.WriteString("<div class=\"sui-rich-context\" role=\"toolbar\" aria-label=\"Selection formatting\" hidden></div>")
	b.WriteString("<div class=\"sui-rich-body\"><div class=\"sui-rich-editor\" contenteditable=\"true\" role=\"textbox\" aria-multiline=\"true\" spellcheck=\"true\" aria-label=\"")
	if title != "" {
		b.WriteString(htmlEscape(title))
	} else {
		b.WriteString(htmlEscape(name))
	}
	b.WriteString("\"></div>")
	b.WriteString("</div>")
	b.WriteString("<div class=\"sui-rich-status\" role=\"status\" aria-live=\"polite\"></div>")
	b.WriteString("</div>`")
	return b.String()
}

func richTextTextAttr(e *Element, key string) string {
	if a := findAttr(e, key); a != nil && !a.hasExpr {
		return strings.TrimSpace(partsContent(a.text))
	}
	return ""
}

func richTextBoolAttr(e *Element, key string, fallback bool) bool {
	if a := findAttr(e, key); a != nil {
		if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
			return a.expr.b
		}
		if !a.hasExpr {
			v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
			return v != "0" && v != "false" && v != "no"
		}
	}
	return fallback
}

func richTextHandlerAttr(e *Element, key string) string {
	if a := findAttr(e, key); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		return a.expr.name
	}
	return ""
}

// treeElemJs renders a nested role=tree/group structure. Expansion is handled
// by the runtime (and persisted per tree path); selection is synchronized with
// a hidden value input and an optional state setter + onChange handler.
func treeElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	title := partsContent(e.text)
	treeKeyName := strings.TrimSpace(title)
	if treeKeyName == "" {
		treeKeyName = "tree"
	}
	treeName := ""
	if a := findAttr(e, "name"); a != nil && !a.hasExpr {
		treeName = strings.TrimSpace(partsContent(a.text))
	}
	if treeName == "" {
		treeName = treeKeyName
	}
	valueJs := ""
	if a := findAttr(e, "value"); a != nil {
		valueJs = htmlAttrValue(a, false)
	}
	selectionJs := ""
	if a := findAttr(e, "selection"); a != nil {
		selectionJs = htmlAttrValue(a, false)
	}
	handler := ""
	if a := findAttr(e, "onChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		handler = a.expr.name
	}
	renameHandler := ""
	if a := findAttr(e, "onRename"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		renameHandler = a.expr.name
	}
	renameKeySetter := ""
	if a := findAttr(e, "renameKey"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" &&
		ctxIsState(ctx, a.expr.name) {
		renameKeySetter = setterName(a.expr.name)
	}
	renameValueSetter := ""
	if a := findAttr(e, "renameValue"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" &&
		ctxIsState(ctx, a.expr.name) {
		renameValueSetter = setterName(a.expr.name)
	}
	setter := ""
	if a := findAttr(e, "value"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" &&
		ctxIsState(ctx, a.expr.name) {
		setter = setterName(a.expr.name)
	}
	editMode := strings.ToLower(strings.TrimSpace(treeTextAttr(e, "editMode")))
	if editMode != "editable" && editMode != "persistent" {
		editMode = "none"
	}
	renameIcon := ""
	if a := findAttr(e, "renameIcon"); a != nil && !a.hasExpr {
		renameIcon = partsContent(a.text)
	}
	showRenameIcon := true
	if a := findAttr(e, "showRenameIcon"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		showRenameIcon = a.expr.b
	}
	depth := ""
	if a := findAttr(e, "editDepth"); a != nil {
		if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
			depth = strconv.Itoa(a.expr.i)
		} else if !a.hasExpr {
			depth = strings.TrimSpace(partsContent(a.text))
		}
	}
	guides := true
	if a := findAttr(e, "guides"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		guides = a.expr.b
	}
	truncate := true
	if a := findAttr(e, "truncate"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		truncate = a.expr.b
	}
	checkbox := false
	if a := findAttr(e, "checkbox"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		checkbox = a.expr.b
	}
	counts := false
	if a := findAttr(e, "counts"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		counts = a.expr.b
	}
	virtual := false
	if a := findAttr(e, "virtual"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		virtual = a.expr.b
	}
	tooltips := false
	if a := findAttr(e, "tooltips"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		tooltips = a.expr.b
	}
	rowHeight := ""
	if a := findAttr(e, "rowHeight"); a != nil {
		if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
			rowHeight = strconv.Itoa(a.expr.i)
		} else if !a.hasExpr {
			rowHeight = strings.TrimSpace(partsContent(a.text))
		}
	}
	tv := treeVisualOptions(e)
	onCheck := ""
	if a := findAttr(e, "onCheck"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onCheck = a.expr.name
	}
	onLoad := ""
	if a := findAttr(e, "onLoad"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onLoad = a.expr.name
	}
	onMove := ""
	if a := findAttr(e, "onMove"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onMove = a.expr.name
	}
	onContextMenu := ""
	if a := findAttr(e, "onContextMenu"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onContextMenu = a.expr.name
	}
	onTreeChange := ""
	if a := findAttr(e, "onTreeChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		onTreeChange = a.expr.name
	}
	multiSelect := false
	if a := findAttr(e, "multiSelect"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		multiSelect = a.expr.b
	}
	search := false
	if a := findAttr(e, "search"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		search = a.expr.b
	}
	dragEnabled := false
	if a := findAttr(e, "drag"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		dragEnabled = a.expr.b
	}
	contextMenu := false
	if a := findAttr(e, "contextMenu"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		contextMenu = a.expr.b
	}
	columns := treeColumns(e)
	columnsEnabled := true
	if a := findAttr(e, "columnsEnabled"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		columnsEnabled = a.expr.b
	}
	if !columnsEnabled {
		columns = nil
	}
	columnLabels := splitTreeColumnText(treeTextAttr(e, "columnLabels"))
	columnWidths := splitTreeColumnWidths(treeTextAttr(e, "columnWidths"))
	sortDir := strings.ToLower(strings.TrimSpace(treeTextAttr(e, "sort")))
	persistSelection := false
	if a := findAttr(e, "persistSelection"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		persistSelection = a.expr.b
	}
	fontStep := treeFontStep(e)
	selectionSetter := ""
	if a := findAttr(e, "selection"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" &&
		ctxIsState(ctx, a.expr.name) {
		selectionSetter = setterName(a.expr.name)
	}
	loadKeySetter := ""
	if a := findAttr(e, "loadKey"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" &&
		ctxIsState(ctx, a.expr.name) {
		loadKeySetter = setterName(a.expr.name)
	}
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">")
		open += " data-sui-tree=\"1\" data-sui-tree-name=\"" + htmlEscape(treeName) + "\" data-sui-tree-key=\"" + htmlEscape(treeKeyName) + "\" role=\"tree\""
		if editMode != "none" {
			open += " data-sui-tree-edit-mode=\"" + htmlEscape(editMode) + "\""
		}
		if depth != "" {
			open += " data-sui-tree-edit-depth=\"" + htmlEscape(depth) + "\""
		}
		if renameIcon != "" {
			open += " data-sui-tree-rename-icon=\"" + htmlEscape(renameIcon) + "\""
		}
		if !showRenameIcon {
			open += " data-sui-tree-show-rename-icon=\"0\""
		}
		if guides {
			open += " data-sui-tree-guides=\"1\""
		}
		if !truncate {
			open += " data-sui-tree-truncate=\"0\""
		}
		if checkbox {
			open += " data-sui-tree-checkbox=\"1\""
		}
		if counts {
			open += " data-sui-tree-counts=\"1\""
		}
		if virtual {
			open += " data-sui-tree-virtual=\"1\""
		}
		if tooltips {
			open += " data-sui-tree-tooltips=\"1\""
		}
		if rowHeight != "" {
			open += " data-sui-tree-row-height=\"" + htmlEscape(rowHeight) + "\""
		}

		if tv.IconLeaf != "" {
			open += " data-sui-tree-icon-leaf=\"" + htmlEscape(tv.IconLeaf) + "\""
		}
		if tv.IconParent != "" {
			open += " data-sui-tree-icon-parent=\"" + htmlEscape(tv.IconParent) + "\""
		}
		if tv.IconParentOpen != "" {
			open += " data-sui-tree-icon-parent-open=\"" + htmlEscape(tv.IconParentOpen) + "\""
		}
		open += " data-sui-tree-toggle-closed=\"" + htmlEscape(tv.ToggleClosed) + "\""
		open += " data-sui-tree-toggle-open=\"" + htmlEscape(tv.ToggleOpen) + "\""
		if setter != "" {
			open += " data-sui-tree-setter=\"" + setter + "\""
		}
		if handler != "" {
			open += " data-sui-tree-onchange=\"" + handler + "\""
		}
		if onCheck != "" {
			open += " data-sui-tree-oncheck=\"" + onCheck + "\""
		}
		if onLoad != "" {
			open += " data-sui-tree-onload=\"" + onLoad + "\""
		}
		if onMove != "" {
			open += " data-sui-tree-onmove=\"" + onMove + "\""
		}
		if onContextMenu != "" {
			open += " data-sui-tree-oncontextmenu=\"" + onContextMenu + "\""
		}
		if onTreeChange != "" {
			open += " data-sui-tree-ontreechange=\"" + onTreeChange + "\""
		}
		if multiSelect {
			open += " data-sui-tree-multi=\"1\""
		}
		if search {
			open += " data-sui-tree-search=\"1\""
		}
		if dragEnabled {
			open += " data-sui-tree-drag=\"1\""
		}
		if contextMenu {
			open += " data-sui-tree-context-menu=\"1\""
		}
		if len(columns) > 0 {
			open += " data-sui-tree-columns=\"" + htmlEscape(strings.Join(columns, ",")) + "\""
		}
		if !columnsEnabled {
			open += " data-sui-tree-columns-enabled=\"0\""
		}
		if strings.TrimSpace(strings.Join(columnLabels, "")) != "" {
			open += " data-sui-tree-column-labels=\"" + htmlEscape(strings.Join(columnLabels, ",")) + "\""
		}
		if len(columnWidths) > 0 {
			var ws []string
			for _, w := range columnWidths {
				ws = append(ws, strconv.Itoa(w))
			}
			open += " data-sui-tree-column-widths=\"" + htmlEscape(strings.Join(ws, ",")) + "\""
		}
		if sortDir == "asc" || sortDir == "desc" {
			open += " data-sui-tree-sort=\"" + sortDir + "\""
		}
		if persistSelection {
			open += " data-sui-tree-persist-selection=\"1\""
		}
		if fontStep > 0 {
			open += " data-sui-tree-font-step=\"" + strconv.FormatFloat(fontStep, 'g', -1, 64) + "\""
		}
		if selectionSetter != "" {
			open += " data-sui-tree-selection-setter=\"" + selectionSetter + "\""
		}
		if loadKeySetter != "" {
			open += " data-sui-tree-load-key-setter=\"" + loadKeySetter + "\""
		}
		if renameKeySetter != "" {
			open += " data-sui-tree-rename-key-setter=\"" + renameKeySetter + "\""
		}
		if renameValueSetter != "" {
			open += " data-sui-tree-rename-value-setter=\"" + renameValueSetter + "\""
		}
		if renameHandler != "" {
			open += " data-sui-tree-onrename=\"" + renameHandler + "\""
		}
		open += ">"
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	if title != "" {
		b.WriteString("<div class=\"sui-tree-title\">")
		b.WriteString(htmlTextParts(e.text))
		b.WriteString("</div>")
	}
	b.WriteString("<input type=\"hidden\" class=\"sui-tree-value\" value=\"")
	b.WriteString(valueJs)
	b.WriteString("\">")
	if multiSelect {
		b.WriteString("<input type=\"hidden\" class=\"sui-tree-selection\" value=\"")
		b.WriteString(selectionJs)
		b.WriteString("\">")
	}
	if search {
		b.WriteString("<div class=\"sui-tree-search\"><input type=\"search\" class=\"sui-tree-search-input\" placeholder=\"Search tree\" aria-label=\"Search tree\"><span class=\"sui-tree-search-count\" aria-live=\"polite\"></span></div>")
	}
	if len(columns) > 0 {
		b.WriteString("<div class=\"sui-tree-grid-head\" role=\"presentation\"><span class=\"sui-tree-grid-name\">Name</span>")
		for i, col := range columns {
			label := treeColumnLabel(col)
			if i < len(columnLabels) && columnLabels[i] != "" {
				label = columnLabels[i]
			}
			b.WriteString("<span class=\"sui-tree-cell sui-tree-cell-" + col + "\"")
			if w := treeColumnWidth(columnWidths, i); w > 0 {
				b.WriteString(" style=\"width:" + strconv.Itoa(w) + "px\"")
			}
			b.WriteString(">" + htmlEscape(label) + "</span>")
		}
		b.WriteString("</div>")
	}
	b.WriteString("<ul class=\"sui-tree-group\" role=\"group\">")
	seq := 0
	for _, child := range sortTreeElements(treeChildren(e), sortDir) {
		b.WriteString(treeItemElemJs(child, ctx, idx*1000+seq+1, 1, "0."+strconv.Itoa(seq), guides, checkbox, counts, tv, columns, columnWidths, sortDir, fontStep))
		seq++
	}
	b.WriteString("</ul></div>`")
	return b.String()
}

func treeItemElemJs(it *Element, ctx *JsCtx, itemID int, level int, path string, guides, checkbox, counts bool, tv treeVisualOpts, columns []string, columnWidths []int, sortDir string, fontStep float64) string {
	open := openTagHtml(it, "li", false, itemID, ctx)
	kids := sortTreeElements(treeChildren(it), sortDir)
	isOpen := treeOpen(it)
	lazy := false
	if a := findAttr(it, "lazy"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		lazy = a.expr.b
	}
	key := treeKey(it, path)
	parent := len(kids) > 0 || lazy
	explicitIcon := treeTextAttr(it, "icon")
	explicitIconOpen := treeTextAttr(it, "iconOpen")
	icon := treeItemIcon(it, parent, isOpen, tv)
	badge := treeTextAttr(it, "badge")
	meta := treeTextAttr(it, "meta")
	size := treeTextAttr(it, "size")
	date := treeTextAttr(it, "date")
	status := treeTextAttr(it, "status")
	cells := treeTextAttr(it, "cells")
	badgeType := strings.ToLower(strings.TrimSpace(treeTextAttr(it, "badgeType")))
	switch badgeType {
	case "info", "success", "warning", "danger", "neutral":
	default:
		badgeType = "neutral"
	}
	checked := ""
	if a := findAttr(it, "checked"); a != nil {
		if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
			if a.expr.b {
				checked = "1"
			} else {
				checked = "0"
			}
		} else if !a.hasExpr {
			v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
			if v == "true" || v == "yes" || v == "1" || v == "on" {
				checked = "1"
			} else {
				checked = "0"
			}
		}
	}
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">")
		open += " data-sui-tree-item=\"1\" data-sui-key=\"" + htmlEscape(key) +
			"\" data-sui-path=\"" + htmlEscape(path) + "\" data-sui-level=\"" + strconv.Itoa(level) +
			"\" data-sui-rename=\"" + htmlEscape(treeItemRename(it)) +
			"\" data-sui-show-rename-icon=\"" + htmlEscape(treeItemShowRenameIcon(it)) +
			"\" data-sui-rename-icon=\"" + htmlEscape(treeItemRenameIcon(it)) +
			"\" data-sui-icon-base=\"" + htmlEscape(explicitIcon) +
			"\" data-sui-icon-open-base=\"" + htmlEscape(explicitIconOpen) +
			"\" data-sui-icon=\"" + htmlEscape(icon) +
			"\" data-sui-badge=\"" + htmlEscape(badge) +
			"\" data-sui-badge-type=\"" + htmlEscape(badgeType) +
			"\" data-sui-meta=\"" + htmlEscape(meta) +
			"\" data-sui-size=\"" + htmlEscape(size) +
			"\" data-sui-date=\"" + htmlEscape(date) +
			"\" data-sui-status=\"" + htmlEscape(status) +
			"\" data-sui-cells=\"" + htmlEscape(cells) +
			"\" data-sui-lazy=\"" + boolFlag(lazy) +
			"\" data-sui-checked=\"" + htmlEscape(checked) + "\" data-sui-open=\""
		if isOpen {
			open += "1"
		} else {
			open += "0"
		}
		open += "\" role=\"treeitem\" aria-level=\"" + strconv.Itoa(level) + "\" aria-selected=\"false\" tabindex=\"-1\""
		if checkbox {
			st := "false"
			if checked == "1" {
				st = "true"
			}
			open += " aria-checked=\"" + st + "\""
		}
		if len(kids) > 0 || lazy {
			if isOpen {
				open += " aria-expanded=\"true\""
			} else {
				open += " aria-expanded=\"false\""
			}
		}
		open += ">"
	}
	var b strings.Builder
	b.WriteString(open)
	b.WriteString("<div class=\"sui-tree-row\">")
	if guides {
		for g := 1; g < level; g++ {
			b.WriteString("<span class=\"sui-tree-guide\" aria-hidden=\"true\"></span>")
		}
	}
	if len(kids) > 0 || lazy {
		b.WriteString("<button type=\"button\" class=\"sui-tree-toggle\" tabindex=\"-1\" aria-label=\"Toggle item\">")
		b.WriteString(htmlEscape(treeToggleText(isOpen, tv)))
		b.WriteString("</button>")
	} else {
		b.WriteString("<span class=\"sui-tree-spacer\"></span>")
	}
	if checkbox {
		st := "false"
		if checked == "1" {
			st = "true"
		}
		b.WriteString("<span class=\"sui-tree-check\" role=\"checkbox\" aria-checked=\"" + st + "\" tabindex=\"-1\" aria-label=\"Toggle item\"></span>")
	}
	if icon != "" {
		b.WriteString("<span class=\"sui-tree-icon\" aria-hidden=\"true\">")
		b.WriteString(htmlEscape(icon))
		b.WriteString("</span>")
	}
	labelText := partsContent(it.text)
	b.WriteString("<span class=\"sui-tree-label\"" + treeFontStyle(level, fontStep) + " title=\"" + htmlEscape(labelText) + "\">")
	b.WriteString(htmlTextParts(it.text))
	b.WriteString("</span>")
	if counts && len(kids) > 0 {
		b.WriteString("<span class=\"sui-tree-badge sui-tree-badge-count\">")
		b.WriteString(strconv.Itoa(len(kids)))
		b.WriteString("</span>")
	}
	if badge != "" {
		b.WriteString("<span class=\"sui-tree-badge sui-tree-badge-" + htmlEscape(badgeType) + "\">")
		b.WriteString(htmlEscape(badge))
		b.WriteString("</span>")
	}
	if meta != "" {
		b.WriteString("<span class=\"sui-tree-meta\">")
		b.WriteString(htmlEscape(meta))
		b.WriteString("</span>")
	}
	if len(columns) > 0 {
		for ci, col := range columns {
			val := treeCellValue(it, col)
			b.WriteString("<span class=\"sui-tree-cell sui-tree-cell-" + col + "\"")
			if w := treeColumnWidth(columnWidths, ci); w > 0 {
				b.WriteString(" style=\"width:" + strconv.Itoa(w) + "px\"")
			}
			b.WriteString(">")
			b.WriteString(htmlEscape(val))
			b.WriteString("</span>")
		}
	}
	b.WriteString("</div>")
	if len(kids) > 0 {
		b.WriteString("<ul class=\"sui-tree-group\" role=\"group\"")
		if !isOpen {
			b.WriteString(" hidden")
		}
		b.WriteString(">")
		for i, child := range kids {
			b.WriteString(treeItemElemJs(child, ctx, itemID*1000+i+1, level+1, path+"."+strconv.Itoa(i), guides, checkbox, counts, tv, columns, columnWidths, sortDir, fontStep))
		}
		b.WriteString("</ul>")
	}
	b.WriteString("</li>")
	return b.String()
}

// datepickerElemJs renders a native date/time/datetime-local input with a
// label, clear action, hidden value binding and an inline error region. The
// browser owns the calendar/clock UI and keyboard handling.
func datepickerElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	mode := "date"
	if a := findAttr(e, "mode"); a != nil && !a.hasExpr {
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		if v == "time" || v == "datetime" {
			mode = v
		}
	}
	inputType := mode
	if inputType == "datetime" {
		inputType = "datetime-local"
	}
	attrValue := func(key string) string {
		if a := findAttr(e, key); a != nil {
			if key == "step" && a.hasExpr && a.expr != nil && a.expr.kind == "int" {
				return strconv.Itoa(a.expr.i)
			}
			return htmlAttrValue(a, false)
		}
		return ""
	}
	valueJs := ""
	if a := findAttr(e, "value"); a != nil {
		valueJs = htmlAttrValue(a, false)
	}
	handler := ""
	if a := findAttr(e, "onChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		handler = a.expr.name
	}
	setter := ""
	if a := findAttr(e, "value"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" &&
		ctxIsState(ctx, a.expr.name) {
		setter = setterName(a.expr.name)
	}
	disabled := false
	if a := findAttr(e, "disabled"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		disabled = a.expr.b
	}
	required := false
	if a := findAttr(e, "required"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		required = a.expr.b
	}
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">")
		open += " data-sui-datepicker=\"1\" data-sui-mode=\"" + htmlEscape(mode) + "\""
		if disabled {
			open += " data-sui-disabled=\"1\""
		}
		if setter != "" {
			open += " data-sui-datepicker-setter=\"" + setter + "\""
		}
		if handler != "" {
			open += " data-sui-datepicker-onchange=\"" + handler + "\""
		}
		open += ">"
	}
	min := attrValue("min")
	max := attrValue("max")
	step := attrValue("step")
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString("<label class=\"sui-datepicker-label\">")
	b.WriteString(htmlTextParts(e.text))
	b.WriteString("</label><div class=\"sui-datepicker-control\"><input type=\"")
	b.WriteString(htmlEscape(inputType))
	b.WriteString("\" class=\"sui-datepicker-input\" autocomplete=\"off\"")
	if min != "" {
		b.WriteString(" min=\"" + min + "\"")
	}
	if max != "" {
		b.WriteString(" max=\"" + max + "\"")
	}
	if step != "" {
		b.WriteString(" step=\"" + step + "\"")
	}
	if required {
		b.WriteString(" required")
	}
	if disabled {
		b.WriteString(" disabled")
	}
	b.WriteString("><button type=\"button\" class=\"sui-datepicker-clear\" aria-label=\"Clear value\" hidden>\u00d7</button>")
	b.WriteString("</div><input type=\"hidden\" class=\"sui-datepicker-value\" value=\"")
	b.WriteString(valueJs)
	b.WriteString("\"")
	if disabled {
		b.WriteString(" disabled")
	}
	b.WriteString("><div class=\"sui-datepicker-error\" role=\"alert\" hidden></div></div>`")
	return b.String()
}

// comboboxElemJs renders an input + listbox widget. The runtime owns filtering,
// keyboard navigation and commit; the generated hidden input carries the value
// binding and the optional onChange handler as data attributes.
func comboboxElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	open = strings.TrimSuffix(open, ">") + " data-sui-combobox-placement=\"" + attrChoice(e, "placement", []string{"auto", "bottom", "top", "left", "right"}, "auto") + "\"" + dataAnimAttrs(e, "fade") + ">"
	// options is a comma-separated string. A literal is embedded as-is; an
	// expression (usually a state string) is evaluated on every render, so a
	// function can replace the list at runtime without rebuilding the view.
	optionsHtml := ""
	if a := findAttr(e, "options"); a != nil {
		if a.hasExpr && a.expr != nil {
			optionsHtml = "${__sui_esc(" + ExprToJs(a.expr) + ")}"
		} else {
			optionsHtml = htmlAttrValue(a, false)
		}
	}
	placeholderAttr := htmlEscape("Search...")
	if a := findAttr(e, "placeholder"); a != nil {
		placeholderAttr = htmlAttrValue(a, false)
	}
	emptyAttr := htmlEscape("No results")
	if a := findAttr(e, "empty"); a != nil {
		emptyAttr = htmlAttrValue(a, false)
	}
	disabled := false
	if a := findAttr(e, "disabled"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		disabled = a.expr.b
	}
	required := false
	if a := findAttr(e, "required"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		required = a.expr.b
	}
	maxOptionLength := ""
	if a := findAttr(e, "maxOptionLength"); a != nil {
		if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
			maxOptionLength = strconv.Itoa(a.expr.i)
		} else if a.hasExpr && a.expr != nil {
			maxOptionLength = "${__sui_esc(" + ExprToJs(a.expr) + ")}"
		} else {
			maxOptionLength = htmlAttrValue(a, false)
		}
	}
	maxOptions := ""
	if a := findAttr(e, "maxOptions"); a != nil {
		if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
			maxOptions = strconv.Itoa(a.expr.i)
		} else if a.hasExpr && a.expr != nil {
			maxOptions = "${__sui_esc(" + ExprToJs(a.expr) + ")}"
		} else {
			maxOptions = htmlAttrValue(a, false)
		}
	}
	val := findAttr(e, "value")
	valueJs := ""
	if val != nil {
		if val.hasExpr && val.expr != nil {
			valueJs = "${__sui_esc(" + ExprToJs(val.expr) + ")}"
		} else {
			valueJs = htmlAttrValue(val, false)
		}
	}
	handler := ""
	if a := findAttr(e, "onChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		handler = a.expr.name
	}
	setter := ""
	if val != nil && val.hasExpr && val.expr != nil && val.expr.kind == "ident" &&
		ctxIsState(ctx, val.expr.name) {
		setter = setterName(val.expr.name)
	}
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">")
		open += " data-sui-combobox=\"1\" data-sui-options=\"" + optionsHtml +
			"\" data-sui-placeholder=\"" + placeholderAttr +
			"\" data-sui-empty=\"" + emptyAttr + "\""
		if disabled {
			open += " data-sui-disabled=\"1\""
		}
		if maxOptionLength != "" {
			open += " data-sui-max-option-length=\"" + maxOptionLength + "\""
		}
		if maxOptions != "" {
			open += " data-sui-max-options=\"" + maxOptions + "\""
		}
		if required {
			open += " data-sui-required=\"1\""
		}
		if setter != "" {
			open += " data-sui-combobox-setter=\"" + setter + "\""
		}
		if handler != "" {
			open += " data-sui-combobox-onchange=\"" + handler + "\""
		}
		open += ">"
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString("<label class=\"sui-combobox-label\">")
	b.WriteString(htmlTextParts(e.text))
	b.WriteString("</label><div class=\"sui-combobox-control\">")
	b.WriteString("<input type=\"text\" class=\"sui-combobox-input\" role=\"combobox\" autocomplete=\"off\" spellcheck=\"false\" aria-expanded=\"false\" placeholder=\"" + placeholderAttr + "\"")
	if disabled {
		b.WriteString(" disabled")
	}
	b.WriteString(">")
	b.WriteString("<button type=\"button\" class=\"sui-combobox-clear\" aria-label=\"Clear selection\" hidden>\u00d7</button>")
	b.WriteString("<button type=\"button\" class=\"sui-combobox-toggle\" aria-label=\"Open options\" aria-expanded=\"false\">\u25be</button>")
	b.WriteString("</div><input type=\"hidden\" class=\"sui-combobox-value\" value=\"" + valueJs + "\"")
	if disabled {
		b.WriteString(" disabled")
	}
	b.WriteString("><div class=\"sui-combobox-popup\" role=\"listbox\" hidden></div><div class=\"sui-combobox-error\" role=\"alert\" hidden></div></div>`")
	return b.String()
}

// tabsElemJs renders a tab set. Each direct `tab` child supplies the button
// label (its text) and the panel content (its children). The runtime restores
// the selected tab per data-sui-tabs-key after a view re-render.
func tabsElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	var tabs []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "tab" {
			tabs = append(tabs, ch.elem)
		}
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString("<div class=\"sui-tablist\" role=\"tablist\">")
	for i, t := range tabs {
		active := i == 0
		b.WriteString("<button type=\"button\" class=\"sui-tab")
		if active {
			b.WriteString(" active")
		}
		b.WriteString("\" role=\"tab\" aria-selected=\"")
		if active {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		b.WriteString("\" tabindex=\"")
		if active {
			b.WriteString("0")
		} else {
			b.WriteString("-1")
		}
		b.WriteString("\">")
		b.WriteString(htmlTextParts(t.text))
		b.WriteString("</button>")
	}
	b.WriteString("</div><div class=\"sui-tabpanels\">")
	for i, t := range tabs {
		active := i == 0
		b.WriteString("<div class=\"sui-tabpanel")
		b.WriteString(classSuffix(t))
		if active {
			b.WriteString(" active")
		}
		b.WriteString("\" role=\"tabpanel\"")
		if !active {
			b.WriteString(" hidden")
		}
		b.WriteString(">")
		if len(t.children) > 0 {
			// Close the HTML template, concatenate the child chunks (which
			// contain their own template literals), then reopen it.
			b.WriteString("` + (")
			b.WriteString(viewNodesJs(childPtrs(t.children), g, ctx))
			b.WriteString(") + `")
		}
		b.WriteString("</div>")
	}
	b.WriteString("</div></div>`")
	return b.String()
}

// accordionElemJs renders disclosure sections. The first section starts open;
// the delegated runtime toggles one section at a time and remembers the open
// index through data-sui-accordion-key after a re-render.
func accordionElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	var items []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "accordionitem" {
			items = append(items, ch.elem)
		}
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	for i, item := range items {
		isOpen := i == 0
		b.WriteString("<div class=\"sui-accordion-item")
		b.WriteString(classSuffix(item))
		if isOpen {
			b.WriteString(" open")
		}
		b.WriteString("\"><button type=\"button\" class=\"sui-accordion-head\" aria-expanded=\"")
		if isOpen {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		b.WriteString("\">")
		b.WriteString(htmlTextParts(item.text))
		b.WriteString("</button><div class=\"sui-accordion-panel\"")
		if !isOpen {
			b.WriteString(" hidden")
		}
		b.WriteString(">")
		if len(item.children) > 0 {
			b.WriteString("` + (")
			b.WriteString(viewNodesJs(childPtrs(item.children), g, ctx))
			b.WriteString(") + `")
		}
		b.WriteString("</div></div>")
	}
	b.WriteString("</div>")
	b.WriteString("`")
	return b.String()
}

// menuElemJs renders a menu as a real button plus a popup. The popup is hidden
// by CSS until the runtime adds the `open` class; the menu label remains
// keyboard-focusable and the items keep their normal wiring (`onClick`,
// `link`) and designer ids.
func drawerElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	openExpr := "false"
	if a := findAttr(e, "open"); a != nil {
		if a.hasExpr && a.expr != nil {
			openExpr = ExprToJs(a.expr)
		} else {
			v := strings.TrimSpace(strings.ToLower(partsContent(a.text)))
			openExpr = strconv.FormatBool(v != "" && v != "false" && v != "0")
		}
	}
	setter := ""
	if a := findAttr(e, "open"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" && ctxIsState(ctx, a.expr.name) {
		setter = setterName(a.expr.name)
	}
	onClose := eventIdentOf(e, "onClose")
	side := "right"
	if a := findAttr(e, "side"); a != nil && !a.hasExpr {
		v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
		if v == "left" || v == "right" {
			side = v
		}
	}
	width := "320px"
	if a := findAttr(e, "width"); a != nil && !a.hasExpr {
		if v := layoutSafeCSSValue(partsContent(a.text)); v != "" {
			width = v
		}
	}
	open := openTagHtml(e, "div", false, idx, ctx)
	open = strings.TrimSuffix(open, ">") + " data-sui-drawer data-sui-drawer-open=\"${(" + openExpr + ") ? \"1\" : \"0\"}\""
	if setter != "" {
		open += " data-sui-drawer-set=\"" + setter + "\""
	}
	if onClose != "" {
		open += " data-sui-drawer-onclose=\"" + onClose + "\""
	}
	open += ">"
	var drawerStyle strings.Builder
	if v := animationSafeValue(attrLiteral(e, "duration")); v != "" {
		drawerStyle.WriteString("--sui-drawer-duration:" + v + ";")
	}
	if v := animationSafeValue(attrLiteral(e, "easing")); v != "" {
		drawerStyle.WriteString("--sui-drawer-easing:" + v + ";")
	}
	if drawerStyle.Len() > 0 {
		open = addOpenTagStyle(open, drawerStyle.String())
	}
	kids := "''"
	if len(e.children) > 0 {
		kids = viewNodesJs(childPtrs(e.children), g, ctx)
	}
	label := htmlTextParts(e.text)
	var sb strings.Builder
	sb.WriteString("`")
	sb.WriteString(open)
	sb.WriteString("<div class=\"sui-drawer-backdrop\" data-sui-drawer-backdrop></div>")
	sb.WriteString("<aside class=\"sui-drawer sui-drawer-" + side + "\" role=\"dialog\" aria-modal=\"true\" style=\"--sui-drawer-width:" + htmlEscape(width) + "\">")
	sb.WriteString("<div class=\"sui-drawer-head\"><b>")
	sb.WriteString(label)
	sb.WriteString("</b><button type=\"button\" class=\"sui-drawer-close\" data-sui-drawer-close aria-label=\"Close\">&#10005;</button></div>")
	sb.WriteString("<div class=\"sui-drawer-body\">` + (")
	sb.WriteString(kids)
	sb.WriteString(") + `</div></aside></div>`")
	return sb.String()
}

func addOpenTagClass(open, cls string) string {
	if cls == "" || !strings.HasSuffix(open, ">") {
		return open
	}
	body := strings.TrimSuffix(open, ">")
	i := strings.Index(body, " class=\"")
	if i >= 0 {
		return body[:i+8] + cls + " " + body[i+8:] + ">"
	}
	j := strings.Index(body, " ")
	if j < 0 {
		j = len(body)
	}
	return body[:j] + " class=\"" + cls + "\"" + body[j:] + ">"
}

func splitterElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	direction := "horizontal"
	if a := findAttr(e, "direction"); a != nil && !a.hasExpr {
		if strings.ToLower(strings.TrimSpace(partsContent(a.text))) == "vertical" {
			direction = "vertical"
		}
	}
	initial := "50%"
	if a := findAttr(e, "initial"); a != nil && !a.hasExpr {
		if v := layoutSafeCSSValue(partsContent(a.text)); v != "" {
			initial = v
		}
	}
	min := "120px"
	if a := findAttr(e, "min"); a != nil && !a.hasExpr {
		if v := layoutSafeCSSValue(partsContent(a.text)); v != "" {
			min = v
		}
	}
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-splitter sui-splitter-"+direction)
	open = strings.TrimSuffix(open, ">") + " data-sui-splitter data-sui-splitter-direction=\"" + direction + "\" data-sui-splitter-min=\"" + htmlEscape(min) + "\">"
	var panes []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "pane" {
			panes = append(panes, ch.elem)
		}
	}
	barOrientation := "vertical"
	if direction == "vertical" {
		barOrientation = "horizontal"
	}
	var sb strings.Builder
	sb.WriteString("`")
	sb.WriteString(open)
	for i := 0; i < len(panes); i++ {
		style := ""
		if i == 0 {
			style = " style=\"flex:0 0 " + htmlEscape(initial) + "\""
		}
		sb.WriteString("<div class=\"sui-split-pane\"" + style + ">` + (")
		sb.WriteString(viewNodesJs(childPtrs(panes[i].children), g, ctx))
		sb.WriteString(") + `</div>")
		if i == 0 && len(panes) > 1 {
			sb.WriteString("<div class=\"sui-split-bar\" role=\"separator\" tabindex=\"0\" aria-orientation=\"" + barOrientation + "\"></div>")
		}
	}
	sb.WriteString("</div>`")
	return sb.String()
}

func dropzoneElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-dropzone")
	if !strings.Contains(open, "data-sui-dropzone") {
		open = strings.TrimSuffix(open, ">") + " data-sui-dropzone=\"1\">"
	}
	// A dropzone is also a picker target. When only dropFiles is set, clicking
	// the area opens the picker into the same state.
	if findAttr(e, "pickFiles") == nil {
		if a := findAttr(e, "dropFiles"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" && ctxIsState(ctx, a.expr.name) {
			extra := " data-sui-pickfiles=\"" + setterName(a.expr.name) + "\""
			if acc := findAttr(e, "accept"); acc != nil {
				extra += " data-sui-pick-accept=\"" + htmlAttrValue(acc, false) + "\""
			}
			extra += dataBoolAttrValue("data-sui-pick-multiple", findAttr(e, "multiple"))
			extra += dataBoolAttrValue("data-sui-pick-directory", findAttr(e, "directory"))
			open = strings.TrimSuffix(open, ">") + extra + ">"
		}
	}
	stateRead := ""
	if a := findAttr(e, "dropFiles"); a != nil && a.hasExpr && a.expr != nil {
		stateRead = ExprToJs(a.expr)
	} else if a := findAttr(e, "pickFiles"); a != nil && a.hasExpr && a.expr != nil {
		stateRead = ExprToJs(a.expr)
	}
	label := htmlTextParts(e.text)
	if strings.TrimSpace(partsContent(e.text)) == "" {
		label = "Drop files here"
	}
	hint := "Drag files here or click to choose"
	if a := findAttr(e, "hint"); a != nil && !a.hasExpr {
		if v := strings.TrimSpace(partsContent(a.text)); v != "" {
			hint = v
		}
	} else if datepickerBool(e, "directory") {
		hint = "Drag a folder here or click to choose"
	}
	var sb strings.Builder
	sb.WriteString("`")
	sb.WriteString(open)
	sb.WriteString("<span class=\"sui-dropzone-icon\" aria-hidden=\"true\">&#8681;</span>")
	sb.WriteString("<span class=\"sui-dropzone-title\">")
	sb.WriteString(label)
	sb.WriteString("</span>")
	sb.WriteString("<span class=\"sui-dropzone-hint\">")
	sb.WriteString(htmlEscape(hint))
	sb.WriteString("</span>")
	if stateRead != "" {
		sb.WriteString("<span class=\"sui-dropzone-status\">${__sui_esc(" + stateRead + ")}</span>")
	}
	if len(e.children) > 0 {
		sb.WriteString("` + (")
		sb.WriteString(viewNodesJs(childPtrs(e.children), g, ctx))
		sb.WriteString(") + `")
	}
	sb.WriteString("</div>`")
	return sb.String()
}

func childElemsByTag(e *Element, tag string) []*Element {
	var out []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == tag {
			out = append(out, ch.elem)
		}
	}
	return out
}

func attrLiteral(e *Element, key string) string {
	if a := findAttr(e, key); a != nil && !a.hasExpr {
		return strings.TrimSpace(partsContent(a.text))
	}
	return ""
}

func attrChoice(e *Element, key string, allowed []string, def string) string {
	v := strings.ToLower(attrLiteral(e, key))
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	return def
}

// animationSafeValue is layoutSafeCSSValue plus a check for characters that
// would break the generated JavaScript template literal.
func animationSafeValue(v string) string {
	v = layoutSafeCSSValue(v)
	if strings.ContainsAny(v, "`$") {
		return ""
	}
	return v
}

// dataAnimAttrs emits the shared overlay animation parameters. Values are
// sanitized by the compiler and applied by __SUI_OVERLAY.animate at runtime.
// The animation name may be a state expression, so an app can change the
// effect while it runs; the runtime still validates it against its own list.
func dataAnimAttrs(e *Element, defEffect string) string {
	var b strings.Builder
	if a := findAttr(e, "animation"); a != nil && a.hasExpr && a.expr != nil {
		b.WriteString(" data-sui-animation=\"${__sui_esc(" + ExprToJs(a.expr) + ")}\"")
	} else {
		effect := attrChoice(e, "animation", []string{"fade", "scale", "slide", "none"}, defEffect)
		b.WriteString(" data-sui-animation=\"" + effect + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "duration")); v != "" {
		b.WriteString(" data-sui-duration=\"" + htmlEscape(v) + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "easing")); v != "" {
		b.WriteString(" data-sui-easing=\"" + htmlEscape(v) + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "delay")); v != "" {
		b.WriteString(" data-sui-delay=\"" + htmlEscape(v) + "\"")
	}
	return b.String()
}

var toggleShapeChoices = map[string][]string{
	"switch":   {"default", "rounded", "square", "pill", "ios", "material", "minimal", "soft"},
	"checkbox": {"default", "rounded", "square", "circle", "ios", "minimal", "soft"},
	"radio":    {"default", "compact", "large", "soft", "minimal"},
}

var toggleThemeChoices = []string{"default", "accent", "success", "warning", "danger", "mono", "dark", "light"}

func jsStringList(items []string) string {
	parts := make([]string, len(items))
	for i := 0; i < len(items); i++ {
		parts[i] = jsString(items[i])
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func togglePresetClasses(e *Element) string {
	prefix := ""
	switch e.tag {
	case "switch":
		prefix = "sui-sw"
	case "checkbox":
		prefix = "sui-cb"
	case "radio":
		prefix = "sui-rd"
	default:
		return ""
	}
	shapes := toggleShapeChoices[e.tag]
	var b strings.Builder
	if a := findAttr(e, "shape"); a != nil {
		if a.hasExpr && a.expr != nil {
			b.WriteString(" ${__sui_preset(" + ExprToJs(a.expr) + ", \"" + prefix + "-shape-\", " + jsStringList(shapes) + ")}")
		} else {
			shape := attrChoice(e, "shape", shapes, "default")
			if shape != "" && shape != "default" {
				b.WriteString(" " + prefix + "-shape-" + shape)
			}
		}
	}
	if a := findAttr(e, "theme"); a != nil {
		if a.hasExpr && a.expr != nil {
			b.WriteString(" ${__sui_preset(" + ExprToJs(a.expr) + ", \"" + prefix + "-theme-\", " + jsStringList(toggleThemeChoices) + ")}")
		} else {
			theme := attrChoice(e, "theme", toggleThemeChoices, "default")
			if theme != "" && theme != "default" && theme != "accent" {
				b.WriteString(" " + prefix + "-theme-" + theme)
			}
		}
	}
	return b.String()
}

func toggleMotionVars(tag string) string {
	switch tag {
	case "switch":
		return "--sui-sw-w,--sui-sw-h,--sui-sw-radius,--sui-sw-color,--sui-sw-thumb-size,--sui-sw-thumb-radius,--sui-sw-thumb-color,--sui-sw-thumb-x,--sui-sw-thumb-y,--sui-sw-border-width,--sui-sw-border-color"
	case "checkbox":
		return "--sui-cb-size,--sui-cb-radius,--sui-cb-color,--sui-cb-border-width,--sui-cb-border-color,--sui-cb-mark-color,--sui-cb-mark-w,--sui-cb-mark-h,--sui-cb-mark-left,--sui-cb-mark-top,--sui-cb-mark-opacity,--sui-cb-mark-rotate"
	case "radio":
		return "--sui-rd-size,--sui-rd-color,--sui-rd-border-width,--sui-rd-border-color,--sui-rd-dot-size,--sui-rd-dot-color,--sui-rd-dot-opacity"
	}
	return ""
}

// toggleMotionAttrs emits the data attributes used by __SUI_MOTION. The key is
// the DFS id of the enclosing label so the same control is matched across
// renders even though the input itself has no data-sui-id.
func toggleMotionAttrs(e *Element, idx int) string {
	if findAttr(e, "animation") == nil || idx < 0 {
		return ""
	}
	effectAttr := ""
	if a := findAttr(e, "animation"); a != nil && a.hasExpr && a.expr != nil {
		effectAttr = " data-sui-motion-effect=\"${__sui_esc(" + ExprToJs(a.expr) + ")}\""
	} else {
		effect := attrChoice(e, "animation", []string{"auto", "morph", "fade", "scale", "slide", "none"}, "morph")
		effectAttr = " data-sui-motion-effect=\"" + effect + "\""
	}
	var b strings.Builder
	b.WriteString(" data-sui-motion data-sui-motion-key=\"" + strconv.Itoa(idx) + "\"")
	b.WriteString(effectAttr)
	b.WriteString(" data-sui-motion-vars=\"" + toggleMotionVars(e.tag) + "\"")
	if a := findAttr(e, "checked"); a != nil && a.hasExpr && a.expr != nil {
		b.WriteString(" data-sui-motion-state=\"1\"")
	}
	if v := animationSafeValue(attrLiteral(e, "duration")); v != "" {
		b.WriteString(" data-sui-motion-duration=\"" + htmlEscape(v) + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "easing")); v != "" {
		b.WriteString(" data-sui-motion-easing=\"" + htmlEscape(v) + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "delay")); v != "" {
		b.WriteString(" data-sui-motion-delay=\"" + htmlEscape(v) + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "stagger")); v != "" {
		b.WriteString(" data-sui-motion-stagger=\"" + htmlEscape(v) + "\"")
	}
	return b.String()
}

// writeMotionAttrs opts any element into the shared FLIP/enter/exit engine.
// motionKey is required for repeated list items because data-sui-id is the same
// for every iteration of a list node.
func writeMotionAttrs(sb *strings.Builder, e *Element, ctx *JsCtx) {
	if g := findAttr(e, "motionGraph"); g != nil && !g.hasExpr {
		if v := strings.TrimSpace(partsContent(g.text)); v != "" {
			sb.WriteString(" data-sui-motion-graph=\"" + htmlEscape(v) + "\"")
		}
	}
	if e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio" {
		return
	}
	if k := findAttr(e, "motionKey"); k != nil {
		if k.hasExpr && k.expr != nil {
			sb.WriteString(" data-sui-motion-key=\"${__sui_esc(" + ExprToJs(k.expr) + ")}\"")
		} else if v := strings.TrimSpace(partsContent(k.text)); v != "" {
			sb.WriteString(" data-sui-motion-key=\"" + htmlEscape(v) + "\"")
		}
	}
	a := findAttr(e, "motion")
	if a == nil {
		return
	}
	effectAttr := ""
	if a.hasExpr && a.expr != nil {
		effectAttr = " data-sui-motion-effect=\"${__sui_esc(" + ExprToJs(a.expr) + ")}\""
	} else {
		effect := attrChoice(e, "motion", []string{"layout", "fade", "scale", "slide", "none"}, "layout")
		if effect == "none" {
			return
		}
		effectAttr = " data-sui-motion-effect=\"" + effect + "\""
	}
	var b strings.Builder
	b.WriteString(" data-sui-motion")
	b.WriteString(effectAttr)
	if attrBoolDefault(e, "motionExit", true) {
		b.WriteString(" data-sui-motion-exit=\"1\"")
	}
	if v := animationSafeValue(attrLiteral(e, "motionDuration")); v != "" {
		b.WriteString(" data-sui-motion-duration=\"" + htmlEscape(v) + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "motionEasing")); v != "" {
		b.WriteString(" data-sui-motion-easing=\"" + htmlEscape(v) + "\"")
	}
	if v := animationSafeValue(attrLiteral(e, "motionDelay")); v != "" {
		b.WriteString(" data-sui-motion-delay=\"" + htmlEscape(v) + "\"")
	}
	sb.WriteString(b.String())
}

func attrBoolDefault(e *Element, key string, def bool) bool {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	if a.hasExpr && a.expr != nil {
		if a.expr.kind == "bool" {
			return a.expr.b
		}
		return def
	}
	v := strings.TrimSpace(strings.ToLower(partsContent(a.text)))
	if v == "" {
		return true
	}
	return v != "false" && v != "0"
}

func attrIntDefault(e *Element, key string, def int) int {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
		return a.expr.i
	}
	if n, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text))); err == nil {
		return n
	}
	return def
}

func addOpenTagStyle(open, decl string) string {
	if decl == "" || !strings.HasSuffix(open, ">") {
		return open
	}
	body := strings.TrimSuffix(open, ">")
	if i := strings.Index(body, " style=\""); i >= 0 {
		if j := strings.Index(body[i+8:], "\""); j >= 0 {
			k := i + 8 + j
			return body[:k] + ";" + decl + body[k:] + ">"
		}
	}
	j := strings.Index(body, " ")
	if j < 0 {
		j = len(body)
	}
	return body[:j] + " style=\"" + decl + "\"" + body[j:] + ">"
}

// renderElementOpen renders one child element with its normal DFS id and
// wiring. Custom containers use it for interactive child elements that they
// lay out themselves (for example command palette entries).
func renderElementOpen(e *Element, tag string, interactive bool, g *ViewGen, ctx *JsCtx) string {
	idx := -1
	if !genCompBody {
		idx = g.count
		g.count = g.count + 1
	}
	addElemRefs(e, g)
	if interactive && !genCompBody {
		collectWiring(e, idx, g, ctx)
	}
	return openTagHtml(e, tag, interactive, idx, ctx)
}

func timelineElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-timeline")
	items := childElemsByTag(e, "timelineitem")
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	for _, it := range items {
		status := strings.ToLower(attrLiteral(it, "status"))
		if status != "done" && status != "active" && status != "error" {
			status = "pending"
		}
		b.WriteString("<div class=\"sui-timeline-item sui-tl-" + status)
		b.WriteString(classSuffix(it))
		b.WriteString("\"><span class=\"sui-timeline-dot\"></span><div class=\"sui-timeline-body\"><span class=\"sui-timeline-title\">")
		b.WriteString(htmlTextParts(it.text))
		b.WriteString("</span>")
		if date := attrLiteral(it, "date"); date != "" {
			b.WriteString("<span class=\"sui-timeline-date\">" + htmlEscape(date) + "</span>")
		}
		if len(it.children) > 0 {
			b.WriteString("<div class=\"sui-timeline-content\">` + (")
			b.WriteString(viewNodesJs(childPtrs(it.children), g, ctx))
			b.WriteString(") + `</div>")
		}
		b.WriteString("</div></div>")
	}
	b.WriteString("</div>`")
	return b.String()
}

func emptyStateElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	icon := attrLiteral(e, "icon")
	if icon == "" {
		icon = "▢"
	}
	hint := attrLiteral(e, "hint")
	label := htmlTextParts(e.text)
	if strings.TrimSpace(partsContent(e.text)) == "" {
		label = "Nothing here"
	}
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-empty")
	if attrBoolDefault(e, "compact", false) {
		open = addOpenTagClass(open, "sui-empty-compact")
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString("<span class=\"sui-empty-icon\" aria-hidden=\"true\">" + htmlEscape(icon) + "</span>")
	b.WriteString("<span class=\"sui-empty-title\">" + label + "</span>")
	if hint != "" {
		b.WriteString("<span class=\"sui-empty-hint\">" + htmlEscape(hint) + "</span>")
	}
	if len(e.children) > 0 {
		b.WriteString("<div class=\"sui-empty-body\">` + (")
		b.WriteString(viewNodesJs(childPtrs(e.children), g, ctx))
		b.WriteString(") + `</div>")
	}
	b.WriteString("</div>`")
	return b.String()
}

func skeletonElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	lines := attrIntDefault(e, "lines", 3)
	if lines < 1 {
		lines = 1
	}
	if lines > 20 {
		lines = 20
	}
	width := layoutSafeCSSValue(attrLiteral(e, "width"))
	if width == "" {
		width = "100%"
	}
	height := layoutSafeCSSValue(attrLiteral(e, "height"))
	if height == "" {
		height = "12px"
	}
	gap := layoutSafeCSSValue(attrLiteral(e, "gap"))
	if gap == "" {
		gap = "8px"
	}
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-skeleton")
	if attrBoolDefault(e, "circle", false) {
		open = addOpenTagClass(open, "sui-skeleton-circle")
	}
	if !attrBoolDefault(e, "animated", true) {
		open = addOpenTagClass(open, "sui-skeleton-static")
	}
	open = addOpenTagStyle(open, "--sui-sk-width:"+width+";--sui-sk-height:"+height+";--sui-sk-gap:"+gap)
	if v := animationSafeValue(attrLiteral(e, "duration")); v != "" {
		open = addOpenTagStyle(open, "--sui-sk-duration:"+v)
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	if attrBoolDefault(e, "circle", false) {
		b.WriteString("<span class=\"sui-skeleton-dot\"></span>")
	} else {
		for i := 0; i < lines; i++ {
			b.WriteString("<span class=\"sui-skeleton-line\"></span>")
		}
	}
	b.WriteString("</div>`")
	return b.String()
}

func stepperElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	valueExpr := "0"
	setter := ""
	if a := findAttr(e, "value"); a != nil {
		if a.hasExpr && a.expr != nil {
			valueExpr = ExprToJs(a.expr)
			if a.expr.kind == "ident" && ctxIsState(ctx, a.expr.name) {
				setter = setterName(a.expr.name)
			}
		} else if n, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text))); err == nil {
			valueExpr = strconv.Itoa(n)
		}
	}
	clickable := setter != "" && attrBoolDefault(e, "clickable", true)
	vertical := attrBoolDefault(e, "vertical", false)
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-stepper")
	if vertical {
		open = addOpenTagClass(open, "sui-stepper-vertical")
	}
	if clickable {
		open = strings.TrimSuffix(open, ">") + " data-sui-stepper data-sui-stepper-set=\"" + setter + "\">"
	}
	steps := childElemsByTag(e, "step")
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	for i, st := range steps {
		b.WriteString("<div class=\"sui-step${(")
		b.WriteString(valueExpr)
		b.WriteString(") > ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(" ? \" done\" : ((")
		b.WriteString(valueExpr)
		b.WriteString(") === ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(" ? \" active\" : \"\")}")
		b.WriteString(classSuffix(st))
		b.WriteString("\"")
		if clickable {
			b.WriteString(" data-sui-step=\"1\" data-sui-step-index=\"" + strconv.Itoa(i) + "\" role=\"button\" tabindex=\"0\"")
		}
		b.WriteString("><span class=\"sui-step-num\">")
		if icon := attrLiteral(st, "icon"); icon != "" {
			b.WriteString(htmlEscape(icon))
		} else {
			b.WriteString("${(")
			b.WriteString(valueExpr)
			b.WriteString(") > ")
			b.WriteString(strconv.Itoa(i))
			b.WriteString(" ? \"✓\" : \"" + strconv.Itoa(i+1) + "\"}")
		}
		b.WriteString("</span><span class=\"sui-step-text\"><span class=\"sui-step-label\">")
		b.WriteString(htmlTextParts(st.text))
		b.WriteString("</span>")
		if desc := attrLiteral(st, "description"); desc != "" {
			b.WriteString("<span class=\"sui-step-desc\">" + htmlEscape(desc) + "</span>")
		}
		b.WriteString("</span></div>")
		if i < len(steps)-1 {
			b.WriteString("<span class=\"sui-step-line\"></span>")
		}
	}
	b.WriteString("</div>`")
	return b.String()
}

func carouselElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	slides := childElemsByTag(e, "slide")
	valueExpr := "0"
	setter := ""
	if a := findAttr(e, "value"); a != nil {
		if a.hasExpr && a.expr != nil {
			valueExpr = ExprToJs(a.expr)
			if a.expr.kind == "ident" && ctxIsState(ctx, a.expr.name) {
				setter = setterName(a.expr.name)
			}
		} else if n, err := strconv.Atoi(strings.TrimSpace(partsContent(a.text))); err == nil {
			valueExpr = strconv.Itoa(n)
		}
	}
	loop := attrBoolDefault(e, "loop", true)
	controls := attrBoolDefault(e, "controls", true)
	indicators := attrBoolDefault(e, "indicators", true)
	autoplay := attrBoolDefault(e, "autoplay", false)
	interval := attrIntDefault(e, "interval", 4000)
	if interval < 500 {
		interval = 500
	}
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-carousel")
	open = strings.TrimSuffix(open, ">") + " data-sui-carousel data-sui-carousel-index=\"${(" + valueExpr + ")}\" data-sui-carousel-count=\"" + strconv.Itoa(len(slides)) + "\""
	if loop {
		open += " data-sui-carousel-loop=\"1\""
	}
	if autoplay {
		open += " data-sui-carousel-autoplay=\"1\" data-sui-carousel-interval=\"" + strconv.Itoa(interval) + "\""
	}
	if setter != "" {
		open += " data-sui-carousel-set=\"" + setter + "\""
	}
	open += ">"
	var carouselStyle strings.Builder
	if h := layoutSafeCSSValue(attrLiteral(e, "height")); h != "" {
		carouselStyle.WriteString("--sui-carousel-height:" + h + ";")
	}
	if v := animationSafeValue(attrLiteral(e, "duration")); v != "" {
		carouselStyle.WriteString("--sui-carousel-duration:" + v + ";")
	}
	if v := animationSafeValue(attrLiteral(e, "easing")); v != "" {
		carouselStyle.WriteString("--sui-carousel-easing:" + v + ";")
	}
	if carouselStyle.Len() > 0 {
		open = addOpenTagStyle(open, carouselStyle.String())
	}
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	b.WriteString("<div class=\"sui-carousel-frame\"><div class=\"sui-carousel-track\" style=\"transform:translateX(${(" + valueExpr + ") * -100}%)\">")
	for i, sl := range slides {
		label := attrLiteral(sl, "label")
		if label == "" {
			label = "Slide " + strconv.Itoa(i+1)
		}
		b.WriteString("<div class=\"sui-carousel-slide")
		b.WriteString(classSuffix(sl))
		b.WriteString("\" data-sui-slide-index=\"" + strconv.Itoa(i) + "\" role=\"group\" aria-label=\"" + htmlEscape(label) + "\">")
		if len(sl.children) > 0 {
			b.WriteString("` + (")
			b.WriteString(viewNodesJs(childPtrs(sl.children), g, ctx))
			b.WriteString(") + `")
		}
		b.WriteString("</div>")
	}
	b.WriteString("</div></div>")
	if controls || indicators {
		b.WriteString("<div class=\"sui-carousel-nav\">")
		if controls {
			b.WriteString("<button type=\"button\" class=\"sui-carousel-prev\" data-sui-carousel-prev aria-label=\"Previous slide\">&#8249;</button>")
		}
		if indicators {
			b.WriteString("<div class=\"sui-carousel-dots\" role=\"tablist\" aria-label=\"Slides\">")
			for i := range slides {
				b.WriteString("<button type=\"button\" class=\"sui-carousel-dot${(" + valueExpr + ") === " + strconv.Itoa(i) + " ? \" active\" : \"\"}\" data-sui-carousel-dot=\"" + strconv.Itoa(i) + "\" aria-label=\"Slide " + strconv.Itoa(i+1) + "\"></button>")
			}
			b.WriteString("</div>")
		}
		if controls {
			b.WriteString("<button type=\"button\" class=\"sui-carousel-next\" data-sui-carousel-next aria-label=\"Next slide\">&#8250;</button>")
		}
		b.WriteString("</div>")
	}
	b.WriteString("</div>`")
	return b.String()
}

func commandbarElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	openAttr := findAttr(e, "open")
	overlay := openAttr != nil
	openExpr := "false"
	setter := ""
	if openAttr != nil {
		if openAttr.hasExpr && openAttr.expr != nil {
			openExpr = ExprToJs(openAttr.expr)
			if openAttr.expr.kind == "ident" && ctxIsState(ctx, openAttr.expr.name) {
				setter = setterName(openAttr.expr.name)
			}
		} else {
			v := strings.TrimSpace(strings.ToLower(partsContent(openAttr.text)))
			openExpr = strconv.FormatBool(v != "" && v != "false" && v != "0")
		}
	}
	hotkey := strings.ToLower(attrLiteral(e, "hotkey"))
	if len(hotkey) > 1 {
		hotkey = hotkey[:1]
	}
	placeholder := attrLiteral(e, "placeholder")
	if placeholder == "" {
		placeholder = strings.TrimSpace(partsContent(e.text))
	}
	if placeholder == "" {
		placeholder = "Type a command..."
	}
	open := addOpenTagClass(openTagHtml(e, "div", false, idx, ctx), "sui-commandbar")
	if overlay {
		open = addOpenTagClass(open, "sui-commandbar-overlay")
	}
	open = strings.TrimSuffix(open, ">") + " data-sui-commandbar" + dataAnimAttrs(e, "fade")
	if overlay {
		open += " data-sui-commandbar-overlay data-sui-commandbar-open=\"${(" + openExpr + ") ? \"1\" : \"0\"}\""
		if setter != "" {
			open += " data-sui-commandbar-set=\"" + setter + "\""
		}
		if hotkey != "" {
			open += " data-sui-commandbar-hotkey=\"" + htmlEscape(hotkey) + "\""
		}
	}
	open += ">"
	role := "search"
	modal := "false"
	if overlay {
		role = "dialog"
		modal = "true"
	}
	cmds := childElemsByTag(e, "command")
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(open)
	if overlay {
		b.WriteString("<div class=\"sui-commandbar-backdrop\" data-sui-commandbar-close></div>")
	}
	b.WriteString("<div class=\"sui-commandbar-panel\" role=\"" + role + "\" aria-modal=\"" + modal + "\">")
	b.WriteString("<input class=\"sui-commandbar-input\" type=\"text\" placeholder=\"" + htmlEscape(placeholder) + "\" aria-label=\"" + htmlEscape(placeholder) + "\" autocomplete=\"off\" spellcheck=\"false\">")
	b.WriteString("<div class=\"sui-commandbar-list\" role=\"listbox\">")
	for i, cmd := range cmds {
		label := partsContent(cmd.text)
		shortcut := attrLiteral(cmd, "shortcut")
		search := strings.ToLower(label)
		if shortcut != "" {
			search += " " + strings.ToLower(shortcut)
		}
		cmdOpen := addOpenTagClass(renderElementOpen(cmd, "button", true, g, ctx), "sui-command")
		if i == 0 {
			cmdOpen = addOpenTagClass(cmdOpen, "active")
		}
		cmdOpen = strings.TrimSuffix(cmdOpen, ">") + " role=\"option\" data-sui-command-index=\"" + strconv.Itoa(i) + "\" data-sui-command-search=\"" + htmlEscape(search) + "\" data-sui-command-label=\"" + htmlEscape(label) + "\">"
		b.WriteString(cmdOpen)
		b.WriteString("<span class=\"sui-command-label\">" + htmlTextParts(cmd.text) + "</span>")
		if shortcut != "" {
			b.WriteString("<span class=\"sui-command-shortcut\">" + htmlEscape(shortcut) + "</span>")
		}
		b.WriteString("</button>")
	}
	b.WriteString("</div><div class=\"sui-commandbar-empty\" hidden>No matching commands</div></div></div>`")
	return b.String()
}

func dataBoolAttrValue(name string, a *Attr) string {
	if a == nil {
		return ""
	}
	var sb strings.Builder
	writeDataBoolAttr(&sb, name, a)
	return sb.String()
}

func popoverElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "span", false, idx, ctx)
	open = strings.TrimSuffix(open, ">") + " data-sui-popover"
	pos := attrChoice(e, "position", []string{"auto", "top", "right", "bottom", "left"}, "auto")
	open += " data-sui-popover-position=\"" + pos + "\"" + dataAnimAttrs(e, "scale") + ">"
	label := htmlTextParts(e.text)
	if strings.TrimSpace(partsContent(e.text)) == "" {
		label = "Info"
	}
	kids := "''"
	if len(e.children) > 0 {
		kids = viewNodesJs(childPtrs(e.children), g, ctx)
	}
	var sb strings.Builder
	sb.WriteString("`")
	sb.WriteString(open)
	sb.WriteString("<button type=\"button\" class=\"sui-popover-trigger\" data-sui-popover-trigger aria-haspopup=\"dialog\" aria-expanded=\"false\">")
	sb.WriteString(label)
	sb.WriteString("</button><div class=\"sui-popover-panel\" role=\"dialog\" hidden>` + (")
	sb.WriteString(kids)
	sb.WriteString(") + `</div></span>`")
	return sb.String()
}

func menuElemJs(e *Element, g *ViewGen, ctx *JsCtx, idx int) string {
	open := openTagHtml(e, "div", false, idx, ctx)
	kids := "''"
	if len(e.children) > 0 {
		kids = viewNodesJs(childPtrs(e.children), g, ctx)
	}
	var sb strings.Builder
	sb.WriteString("`")
	sb.WriteString(open)
	sb.WriteString("<button type=\"button\" class=\"sui-menu-button\" role=\"menuitem\" aria-haspopup=\"true\" aria-expanded=\"false\">")
	sb.WriteString(htmlTextParts(e.text))
	sb.WriteString("</button><div class=\"sui-menu-popup\" role=\"menu\">` + (")
	sb.WriteString(kids)
	sb.WriteString(") + `</div></div>`")
	return sb.String()
}

func collectWiring(e *Element, idx int, g *ViewGen, ctx *JsCtx) {
	var lines []string
	// Bind each declared event handler to the matching DOM property. Done in a
	// fixed order so generated output is deterministic.
	for _, key := range elemEventOrder {
		if isDndEventKey(key) {
			// DnD handlers are inline so repeated list rows each receive one.
			continue
		}
		dom := eventAttrJs[key]
		a := findAttr(e, key)
		if a != nil && a.hasExpr && a.expr != nil {
			lines = append(lines, "e"+strconv.Itoa(idx)+"."+dom+" = "+ExprToJs(a.expr)+";")
		}
	}
	// hasEv reports whether the element declares an explicit handler for the
	// given source event (e.g. "onChange"). When it does, the user is managing
	// this control's value themselves, so the automatic state setter below must
	// NOT run: it would overwrite the handler's effect (and could even coerce a
	// non-boolean state, such as an int counter, to true/false).
	hasEv := func(key string) bool {
		a := findAttr(e, key)
		return a != nil && a.hasExpr && a.expr != nil
	}
	if e.tag == "input" || e.tag == "textarea" || e.tag == "select" || e.tag == "number" || e.tag == "slider" {
		val := findAttr(e, "value")
		if val != nil && val.hasExpr && val.expr != nil && val.expr.kind == "ident" && ctxIsState(ctx, val.expr.name) && !hasEv("onInput") {
			lines = append(lines, "e"+strconv.Itoa(idx)+".oninput = (ev) => "+setterName(val.expr.name)+"(ev.target.value);")
		}
	}
	if e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio" {
		val := findAttr(e, "checked")
		if val != nil && val.hasExpr && val.expr != nil && val.expr.kind == "ident" && ctxIsState(ctx, val.expr.name) && !hasEv("onChange") {
			lines = append(lines, "e"+strconv.Itoa(idx)+".onchange = (ev) => "+setterName(val.expr.name)+"(ev.target.checked);")
		}
	}
	if e.tag == "input" {
		if opts := inputOptions(e); len(opts) > 0 {
			var quoted []string
			for _, o := range opts {
				quoted = append(quoted, jsString(o))
			}
			lines = append(lines, "__sui_ac(e"+strconv.Itoa(idx)+", ["+strings.Join(quoted, ", ")+"]);")
		}
	}
	if len(lines) > 0 {
		g.wires = append(g.wires, Wire{idx: idx, lines: lines})
	}
}

func viewNodesJs(nodes []*ViewNode, g *ViewGen, ctx *JsCtx) string {
	var chunks []string
	for i := 0; i < len(nodes); i++ {
		chunks = append(chunks, viewNodeJs(nodes[i], g, ctx))
	}
	if len(chunks) == 0 {
		return "''"
	}
	return strings.Join(chunks, " + ")
}

func viewNodeJs(n *ViewNode, g *ViewGen, ctx *JsCtx) string {
	if n == nil {
		return "''"
	}
	js := viewNodeJsRaw(n, g, ctx)
	if genEmitSourceMarkers {
		return srcMarker(n.span) + js
	}
	return js
}

func viewNodeJsRaw(n *ViewNode, g *ViewGen, ctx *JsCtx) string {
	if n == nil {
		return "''"
	}
	if n.kind == "element" {
		return elemJs(n.elem, g, ctx)
	}
	if n.kind == "if" {
		thenC := viewNodesJs(n.then_, g, ctx)
		elseC := "''"
		if len(n.else_) > 0 {
			elseC = viewNodesJs(n.else_, g, ctx)
		}
		return "(" + ExprToJs(n.cond) + " ? " + thenC + " : " + elseC + ")"
	}
	if n.kind == "list" {
		bodyC := viewNodesJs(n.body, g, ctx)
		itemC := bodyC
		if n.key != nil {
			itemC = "__sui_keyed(" + bodyC + ", " + ExprToJs(n.key) + ")"
		}
		if n.item2 != "" {
			tmp := "__sui_" + n.item
			return "(() => { let s = \"\"; for (const " + tmp + " of Object.entries(" + ExprToJs(n.iterable) + ")) { const " + n.item + " = " + tmp + "[0]; const " + n.item2 + " = " + tmp + "[1]; s += " + itemC + "; } return s; })()"
		}
		return "(() => { let s = \"\"; for (const " + n.item + " of (" + ExprToJs(n.iterable) + ")) { s += " + itemC + "; } return s; })()"
	}
	return "''"
}

func addPartsRefs(parts []StrPart, g *ViewGen) {
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "expr" {
			addExprRefs(parts[i].expr, g)
		}
	}
}

func addExprRefs(e *Expr, g *ViewGen) {
	if e == nil {
		return
	}
	k := e.kind
	if k == "ident" {
		if refExcl[e.name] {
			return
		}
		g.refs = append(g.refs, e.name)
		return
	}
	if k == "str" {
		addPartsRefs(e.parts, g)
		return
	}
	if k == "member" {
		addExprRefs(e.obj, g)
		return
	}
	if k == "call" {
		addExprRefs(e.target, g)
		for i := 0; i < len(e.args); i++ {
			addExprRefs(e.args[i], g)
		}
		return
	}
	if k == "array" {
		for i := 0; i < len(e.elements); i++ {
			addExprRefs(e.elements[i], g)
		}
		return
	}
	if k == "map" {
		for i := 0; i < len(e.vals); i++ {
			addExprRefs(e.vals[i], g)
		}
		return
	}
	if k == "index" {
		addExprRefs(e.obj, g)
		addExprRefs(e.idx, g)
		return
	}
	if k == "unary" {
		addExprRefs(e.operand, g)
		return
	}
	if k == "binary" {
		addExprRefs(e.l, g)
		addExprRefs(e.r, g)
		return
	}
	if k == "cond" {
		addExprRefs(e.cond, g)
		addExprRefs(e.l, g)
		addExprRefs(e.r, g)
		return
	}
	if k == "range" {
		addExprRefs(e.lo, g)
		addExprRefs(e.hi, g)
		return
	}
}

func addNodesRefs(nodes []ViewNode, g *ViewGen) {
	for i := 0; i < len(nodes); i++ {
		addNodeRefs(nodes[i], g)
	}
}

func addPtrNodesRefs(nodes []*ViewNode, g *ViewGen) {
	for i := 0; i < len(nodes); i++ {
		addNodeRefs(*nodes[i], g)
	}
}

func addNodeRefs(n ViewNode, g *ViewGen) {
	if n.kind == "element" {
		addElemRefs(n.elem, g)
		return
	}
	if n.kind == "if" {
		addExprRefs(n.cond, g)
		addPtrNodesRefs(n.then_, g)
		addPtrNodesRefs(n.else_, g)
		return
	}
	if n.kind == "list" {
		addExprRefs(n.iterable, g)
		if n.key != nil {
			addExprRefs(n.key, g)
		}
		addPtrNodesRefs(n.body, g)
		return
	}
}

func addElemRefs(e *Element, g *ViewGen) {
	if e == nil {
		return
	}
	// A component invocation: walk the component's OWN body so the states it
	// references are bound on the root view. Component parameters are locals
	// and must not be collected (they may shadow a state of the same name).
	if genViewParams[e.tag] != nil {
		vd, ok := genViews[e.tag]
		if ok {
			saved := refExcl
			excl := map[string]bool{}
			for i := 0; i < len(vd.params); i++ {
				excl[vd.params[i].name] = true
			}
			refExcl = excl
			addNodesRefs(vd.body, g)
			refExcl = saved
		}
		// The caller's block is ALSO generated in this (root) scope: bind the
		// states it references.
		addNodesRefs(e.children, g)
		return
	}
	addPartsRefs(e.text, g)
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		// value:/checked: on a form control is a write-only binding: the
		// control updates itself and the state setter runs, so the view does
		// not need to be re-bound to that state. Skipping the ref avoids a
		// full re-render (and the resulting jank/flicker) on every keystroke
		// or drag threshold. If the state is read anywhere else, that read
		// still binds the view.
		if bindingOnlyAttr(e, &a) {
			continue
		}
		addPartsRefs(a.text, g)
		if a.hasExpr {
			addExprRefs(a.expr, g)
		}
	}
	addNodesRefs(e.children, g)
}

// bindingOnlyAttr reports whether a is an automatic form binding whose only
// effect is to write state from the control (the control already shows the
// value, and an explicit handler is absent).
func bindingOnlyAttr(e *Element, a *Attr) bool {
	if e == nil || a == nil || !a.hasExpr || a.expr == nil || a.expr.kind != "ident" {
		return false
	}
	if a.key == "value" {
		switch e.tag {
		case "input", "textarea", "select", "number", "slider", "datepicker", "combobox", "tree":
		default:
			return false
		}
		if h := findAttr(e, "onInput"); h != nil && h.hasExpr && h.expr != nil {
			return false
		}
		return true
	}
	if a.key == "checked" {
		switch e.tag {
		case "checkbox", "switch", "radio":
		default:
			return false
		}
		if h := findAttr(e, "onChange"); h != nil && h.hasExpr && h.expr != nil {
			return false
		}
		// A toggle with an animation needs the re-render so the motion engine
		// can tween from the previous to the new checked appearance. Without an
		// animation the native state change is already final and the view does
		// not need to re-render.
		if an := findAttr(e, "animation"); an != nil {
			if !an.hasExpr && strings.EqualFold(strings.TrimSpace(attrChoice(e, "animation", []string{"auto", "morph", "fade", "scale", "slide", "none"}, "morph")), "none") {
				return true
			}
			return false
		}
		return true
	}
	return false
}

func viewBindings(states []string, viewName string, g *ViewGen) string {
	seen := map[string]bool{}
	var names []string
	var addRef func(name string)
	addRef = func(name string) {
		if seen[name] {
			return
		}
		seen[name] = true
		if isNameIn(states, name) {
			if !isNameIn(names, name) {
				names = append(names, name)
			}
			return
		}
		if deps := genComputedDeps[name]; deps != nil {
			for _, dep := range deps {
				addRef(dep)
			}
		}
	}
	for i := 0; i < len(g.refs); i++ {
		addRef(g.refs[i])
	}
	var out []string
	sorted := sortStringList(names)
	for i := 0; i < len(sorted); i++ {
		out = append(out, "__sui_bindView("+jsString(sorted[i])+", "+jsString(viewName)+");")
	}
	return strings.Join(out, "\n")
}

func rendererJs(p *Prog, v *ViewDecl, ctx *JsCtx, multi bool, base *int, feat *progFeatures) string {
	var g ViewGen
	// Continue the DFS numbering from prior root views (multi-page mode);
	// single-page programs pass 0, so their ids are unchanged.
	g.count = *base
	addPtrNodesRefs(childPtrs(v.body), &g)
	bodyJs := viewNodesJs(childPtrs(v.body), &g, ctx)
	*base = g.count
	var sb strings.Builder
	sb.WriteString(srcMarker(v.span))
	sb.WriteString("function render_" + v.name + "() {\n")
	if multi {
		// Multi-page programs render every root view into its own box under
		// #app; only the active page (__SUI_PAGE) is visible. __sui_page(n)
		// switches the visible page and re-renders.
		sb.WriteString("  const rootel = document.getElementById(\"app\");\n")
		sb.WriteString("  let box = rootel.querySelector('[data-sui-pagebox=\"" + v.name + "\"]');\n")
		sb.WriteString("  if (!box) { box = document.createElement('div'); box.className = 'sui-view'; box.setAttribute('data-sui-pagebox', " + jsString(v.name) + "); rootel.appendChild(box); }\n")
		sb.WriteString("  const __viewHtml = `<div class=\"sui-view\" data-sui-view=\"" + v.name + "\">`\n")
		sb.WriteString("    + " + bodyJs + "\n")
		sb.WriteString("    + `</div>`;\n")
		sb.WriteString("  __sui_replace_view(box, __viewHtml);\n")
		sb.WriteString("  box.style.display = (String(window.__SUI_PAGE) === \"" + v.name + "\") ? '' : 'none';\n")
	} else {
		sb.WriteString("  const el = document.getElementById(\"app\");\n")
		sb.WriteString("  const __viewHtml = `<div class=\"sui-view\" data-sui-view=\"" + v.name + "\">`\n")
		sb.WriteString("    + " + bodyJs + "\n")
		sb.WriteString("    + `</div>`;\n")
		sb.WriteString("  __sui_replace_view(el, __viewHtml);\n")
	}
	for i := 0; i < len(g.wires); i++ {
		w := g.wires[i]
		sb.WriteString("  const e" + strconv.Itoa(w.idx) + " = document.querySelector('[data-sui-id=\"" + strconv.Itoa(w.idx) + "\"]');\n")
		// Conditional branches render only one side, so an element that belongs
		// to the inactive branch is absent after a re-render. Guard the wiring
		// instead of throwing and skipping every handler that follows it.
		sb.WriteString("  if (e" + strconv.Itoa(w.idx) + ") {\n")
		for j := 0; j < len(w.lines); j++ {
			sb.WriteString("    " + w.lines[j] + "\n")
		}
		sb.WriteString("  }\n")
	}
	if feat.needsTable() {
		sb.WriteString("  __SUI_TBL.afterRender();\n")
	}
	if feat.needsDnd() {
		sb.WriteString("  __SUI_DND.afterRender();\n")
	}
	sb.WriteString("  if (window.__SUI_RICH) __SUI_RICH.afterRender();\n")
	sb.WriteString("}\n")
	sb.WriteString("__SUI_views[" + jsString(v.name) + "] = render_" + v.name + ";\n")
	binds := viewBindings(ctx.states, v.name, &g)
	if binds != "" {
		sb.WriteString(binds + "\n")
	}
	return sb.String()
}

// componentChunkJs renders a component as a reusable chunk function:
//
//	function __html_<Name>(p1, p2) { return <html chunk>; }
//
// The chunk is concatenated into the invoking view template at the call site
// (__html_<Name>({ ... })). Interactive elements inside the chunk use inline
// handlers on global functions (no ids, safe across repeated instances).
func componentChunkJs(p *Prog, v *ViewDecl, ctx *JsCtx) string {
	saved := genCompBody
	savedFn := genCompFn
	savedInst := genCompInst
	genCompBody = true
	genCompFn = map[string]bool{}
	genCompInst = map[string]bool{}
	for i := 0; i < len(v.params); i++ {
		if v.params[i].typename == "fn" {
			genCompFn[v.params[i].name] = true
		}
	}
	for i := 0; i < len(v.inst_state); i++ {
		genCompInst[v.inst_state[i].name] = true
	}
	savedMethods := genCompMethods
	genCompMethods = map[string]bool{}
	for i := 0; i < len(v.methods); i++ {
		genCompMethods[v.methods[i].name] = true
	}
	usesChildren := compUsesChildren(v.body)
	savedScope := len(ctx.scope)
	for i := 0; i < len(v.params); i++ {
		ctx.scope = append(ctx.scope, v.params[i].name)
	}
	if usesChildren {
		ctx.scope = append(ctx.scope, "children")
	}
	if len(v.inst_state) > 0 {
		ctx.scope = append(ctx.scope, "__key")
	}
	var g ViewGen
	bodyJs := viewNodesJs(childPtrs(v.body), &g, ctx)
	var methodLines []string
	for i := 0; i < len(v.methods); i++ {
		methodLines = append(methodLines, funcJs(&v.methods[i], ctx))
		methodLines = append(methodLines, "inst."+v.methods[i].name+" = "+v.methods[i].name+";")
	}
	ctxTruncateScope(ctx, savedScope)
	genCompFn = savedFn
	genCompInst = savedInst
	genCompMethods = savedMethods
	genCompBody = saved
	var paramNames []string
	if len(v.inst_state) > 0 {
		paramNames = append(paramNames, "__inst: inst")
		paramNames = append(paramNames, "__key")
	}
	for i := 0; i < len(v.params); i++ {
		paramNames = append(paramNames, v.params[i].name)
	}
	if usesChildren {
		paramNames = append(paramNames, "children")
	}
	var pre []string
	for i := 0; i < len(v.inst_state); i++ {
		sd := v.inst_state[i]
		pre = append(pre, "if (!(\""+sd.name+"\" in inst)) inst."+sd.name+" = "+ExprToJs(sd.init)+";")
		pre = append(pre, "let "+sd.name+" = inst."+sd.name+";")
	}
	head := ""
	if len(pre) > 0 {
		head = strings.Join(pre, "\n") + "\n"
	}
	methodBlock := ""
	if len(methodLines) > 0 {
		methodBlock = strings.Join(methodLines, "\n") + "\n"
	}
	return srcMarker(v.span) + "function __html_" + v.name + "({ " + strings.Join(paramNames, ", ") + " }) {" + "\n" +
		head +
		methodBlock +
		"  return " + bodyJs + ";" + "\n" +
		"}"
}

// compUsesChildren reports whether a component body contains the `children`
// slot marker (directly or inside if/list branches).
func compUsesChildren(nodes []ViewNode) bool {
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		if n.kind == "element" && n.elem != nil {
			if n.elem.tag == "children" {
				return true
			}
			if compUsesChildren(n.elem.children) {
				return true
			}
		} else if n.kind == "if" {
			if compUsesChildren(derefViewNodes(n.then_)) || compUsesChildren(derefViewNodes(n.else_)) {
				return true
			}
		} else if n.kind == "list" {
			if compUsesChildren(derefViewNodes(n.body)) {
				return true
			}
		}
	}
	return false
}

/* Source map deep: mapping SUI line -> JS line generated at build */
