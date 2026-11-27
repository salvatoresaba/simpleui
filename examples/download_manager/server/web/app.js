"use strict";

const __SUI_views  = Object.create(null);
const __SUI_bind   = Object.create(null);
const __SUI_inst   = Object.create(null);

function __sui_bindView(stateVar, viewName) {
  (__SUI_bind[stateVar] ||= []).push(viewName);
}
function __sui_preset(value, prefix, allowed) {
  var v = String(value == null ? '' : value);
  if (v === 'default' || v === 'accent' || !v) return '';
  for (var i = 0; i < allowed.length; i++) if (allowed[i] === v) return ' ' + prefix + v;
  return '';
}
const __SUI_COMPUTED = [];
var __sui_pendingRender = false;
var __sui_pendingViews = null;
function __sui_recompute() {
  for (var i = 0; i < __SUI_COMPUTED.length; i++) __SUI_COMPUTED[i]();
}
function __sui_schedule(views) {
  if (!views || !views.length) return;
  if (!__sui_pendingViews) __sui_pendingViews = Object.create(null);
  for (var i = 0; i < views.length; i++) __sui_pendingViews[views[i]] = 1;
  if (__sui_pendingRender) return;
  __sui_pendingRender = true;
  var flush = function () {
    __sui_pendingRender = false;
    var names = Object.keys(__sui_pendingViews || {});
    __sui_pendingViews = null;
    __sui_recompute();
    for (var j = 0; j < names.length; j++) if (__SUI_views[names[j]]) __SUI_views[names[j]]();
  };
  if (typeof queueMicrotask === 'function') queueMicrotask(flush);
  else setTimeout(flush, 0);
}
function __sui_react(stateVar) {
  // Computed values are refreshed immediately so a handler that reads one
  // after a state change sees the current value. Only views bound to this
  // state are scheduled, and updates in the same tick are batched.
  __sui_recompute();
  __sui_schedule(__SUI_bind[stateVar] || []);
}
function __sui_render() {
  __sui_recompute();
  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();
}
function __sui_refresh() {
  __sui_recompute();
  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();
}
function __sui_intl_locale(locale) {
  var l = locale ? String(locale) : '';
  if (!l && typeof __sui_i18n_locale === 'function') l = __sui_i18n_locale();
  if (!l) { try { l = (navigator.languages && navigator.languages[0]) || navigator.language || 'en'; } catch (e) { l = 'en'; } }
  return l || 'en';
}
function __sui_intl_date_value(value) {
  if (value instanceof Date) return value;
  var d = typeof value === 'number' ? new Date(value) : new Date(String(value));
  return isNaN(d.getTime()) ? null : d;
}
function __sui_intl_style(style, kind) {
  var s = String(style || 'medium').toLowerCase();
  if (s === 'iso') return 'iso';
  if (s !== 'short' && s !== 'long' && s !== 'full') s = 'medium';
  var opts = {};
  if (kind === 'time') opts.timeStyle = s;
  else if (kind === 'datetime') { opts.dateStyle = s; opts.timeStyle = s; }
  else opts.dateStyle = s;
  return opts;
}
function __sui_intl_format_date(value, style, locale, timeZone, kind) {
  var d = __sui_intl_date_value(value);
  if (!d) return 'Invalid date';
  var opts = __sui_intl_style(style, kind);
  if (opts === 'iso') {
    if (kind === 'date') return d.toISOString().slice(0, 10);
    if (kind === 'time') return d.toISOString().slice(11, 19);
    return d.toISOString();
  }
  if (timeZone) opts.timeZone = String(timeZone);
  try { return new Intl.DateTimeFormat(__sui_intl_locale(locale), opts).format(d); } catch (e) { return 'Invalid date'; }
}
function __sui_intl_date(value, style, locale, timeZone) { return __sui_intl_format_date(value, style, locale, timeZone, 'date'); }
function __sui_intl_time(value, style, locale, timeZone) { return __sui_intl_format_date(value, style, locale, timeZone, 'time'); }
function __sui_intl_datetime(value, style, locale, timeZone) { return __sui_intl_format_date(value, style, locale, timeZone, 'datetime'); }
function __sui_intl_digits(decimals) {
  if (decimals === undefined || decimals === null || decimals === '') return null;
  var d = Number(decimals);
  return isFinite(d) && d >= 0 ? Math.floor(d) : null;
}
function __sui_intl_number(value, decimals, locale) {
  var n = Number(value);
  if (!isFinite(n)) return '';
  var opts = {}, d = __sui_intl_digits(decimals);
  if (d !== null) { opts.minimumFractionDigits = d; opts.maximumFractionDigits = d; }
  try { return new Intl.NumberFormat(__sui_intl_locale(locale), opts).format(n); } catch (e) { return String(n); }
}
function __sui_intl_currency(value, currency, locale) {
  var n = Number(value), code = String(currency || 'USD').toUpperCase();
  if (!isFinite(n)) return '';
  try { return new Intl.NumberFormat(__sui_intl_locale(locale), { style: 'currency', currency: code }).format(n); } catch (e) { return code + ' ' + n; }
}
function __sui_intl_percent(value, decimals, locale) {
  var n = Number(value);
  if (!isFinite(n)) return '';
  var opts = { style: 'percent' }, d = __sui_intl_digits(decimals);
  if (d !== null) { opts.minimumFractionDigits = d; opts.maximumFractionDigits = d; }
  try { return new Intl.NumberFormat(__sui_intl_locale(locale), opts).format(n); } catch (e) { return String(n); }
}
function __sui_intl_relative(value, unit, locale, numeric) {
  var u = String(unit || ''), n = Number(value);
  var allowed = { second: 1, seconds: 1, minute: 1, minutes: 1, hour: 1, hours: 1, day: 1, days: 1, week: 1, weeks: 1, month: 1, months: 1, quarter: 1, quarters: 1, year: 1, years: 1 };
  if (!allowed[u] || !isFinite(n)) return '';
  var mode = String(numeric || 'auto');
  if (mode !== 'always' && mode !== 'auto') mode = 'auto';
  try { return new Intl.RelativeTimeFormat(__sui_intl_locale(locale), { numeric: mode }).format(Math.trunc(n), u); } catch (e) { return ''; }
}
function __sui_intl_timezone() {
  try { return Intl.DateTimeFormat().resolvedOptions().timeZone || ''; } catch (e) { return ''; }
}
function __sui_keyed(html, key) {
  var k = __sui_esc(key);
  return String(html).replace(/^\s*<([a-zA-Z][\w-]*)/, '<$1 data-sui-key="' + k + '"');
}
function __sui_rowkey(el, index) {
  var row = el && el.closest ? el.closest('[data-sui-key]') : null;
  if (row) {
    var key = row.getAttribute('data-sui-key');
    if (key !== null && key !== '') return key;
  }
  return String(index);
}
function __sui_assert(cond, msg) {
  if (!cond) throw new Error(msg == null || String(msg) === '' ? 'assertion failed' : String(msg));
}
function __sui_boundary(fn, fallback) {
  try { return fn(); } catch (e) {
    try { console.error('SUI render error:', e); } catch (ignored) {}
    var msg = fallback == null || String(fallback) === '' ? 'Something went wrong.' : String(fallback);
    var detail = e && e.message ? String(e.message) : String(e);
    return '<div class="sui-error" role="alert" data-sui-error="' + __sui_esc(detail) + '">' + __sui_esc(msg) + '</div>';
  }
}
// Collection helpers take a declared function (or its name for an fn
// parameter) and call it once per item. Extra arguments are ignored by
// callbacks that declare fewer parameters.
function __sui_callback(fn, name) {
  if (typeof fn === 'function') return fn;
  if (typeof fn === 'string' && typeof window[fn] === 'function') return window[fn];
  throw new Error(name + ' expects a function name');
}
function __sui_map(arr, fn) {
  var f = __sui_callback(fn, 'map'), list = arr || [], out = [];
  for (var i = 0; i < list.length; i++) out.push(f(list[i], i));
  return out;
}
function __sui_filter(arr, fn) {
  var f = __sui_callback(fn, 'filter'), list = arr || [], out = [];
  for (var i = 0; i < list.length; i++) if (f(list[i], i)) out.push(list[i]);
  return out;
}
function __sui_reduce(arr, init, fn) {
  var f = __sui_callback(fn, 'reduce'), list = arr || [], acc = init;
  for (var i = 0; i < list.length; i++) acc = f(acc, list[i], i);
  return acc;
}
function __sui_find(arr, fn) {
  var f = __sui_callback(fn, 'find'), list = arr || [];
  for (var i = 0; i < list.length; i++) if (f(list[i], i)) return list[i];
  return null;
}
function __sui_find_index(arr, fn) {
  var f = __sui_callback(fn, 'findIndex'), list = arr || [];
  for (var i = 0; i < list.length; i++) if (f(list[i], i)) return i;
  return -1;
}
function __sui_any(arr, fn) {
  var f = __sui_callback(fn, 'any'), list = arr || [];
  for (var i = 0; i < list.length; i++) if (f(list[i], i)) return true;
  return false;
}
function __sui_all(arr, fn) {
  var f = __sui_callback(fn, 'all'), list = arr || [];
  for (var i = 0; i < list.length; i++) if (!f(list[i], i)) return false;
  return true;
}
function __sui_sort_by(arr, fn) {
  var f = __sui_callback(fn, 'sortBy'), list = (arr || []).slice();
  list.sort(function (a, b) { var ka = f(a), kb = f(b); if (ka < kb) return -1; if (ka > kb) return 1; return 0; });
  return list;
}
function __sui_unique(arr) {
  var list = arr || [], out = [], seen = new Set();
  for (var i = 0; i < list.length; i++) { if (!seen.has(list[i])) { seen.add(list[i]); out.push(list[i]); } }
  return out;
}
// Internationalization: catalogs are emitted by i18n declarations. The
// active locale comes from ?locale=, then localStorage, then the browser.
var __SUI_I18N = { catalogs: {}, locale: '', fallback: '', dir: 'ltr' };
function __sui_i18n_rtl(locale) { return /^(ar|he|fa|ur|yi|dv|ps|sd)(-|$)/i.test(String(locale || '')); }
function __sui_i18n_match(locale) {
  var want = String(locale || '').toLowerCase();
  if (!want) return '';
  var keys = Object.keys(__SUI_I18N.catalogs || {});
  for (var i = 0; i < keys.length; i++) if (keys[i].toLowerCase() === want) return keys[i];
  var base = want.split('-')[0];
  for (var j = 0; j < keys.length; j++) if (keys[j].toLowerCase().split('-')[0] === base) return keys[j];
  return '';
}
function __sui_i18n_apply(locale) {
  __SUI_I18N.locale = locale || __SUI_I18N.fallback;
  __SUI_I18N.dir = __sui_i18n_rtl(__SUI_I18N.locale) ? 'rtl' : 'ltr';
  if (typeof document !== 'undefined' && document.documentElement) {
    document.documentElement.lang = __SUI_I18N.locale;
    document.documentElement.dir = __SUI_I18N.dir;
  }
}
function __sui_i18n_init(catalogs, fallback) {
  __SUI_I18N.catalogs = catalogs || {};
  __SUI_I18N.fallback = fallback || '';
  var chosen = '';
  try {
    var q = new URLSearchParams(location.search).get('locale');
    if (q) chosen = __sui_i18n_match(q);
    if (!chosen) { var saved = localStorage.getItem('sui.locale'); if (saved) chosen = __sui_i18n_match(saved); }
  } catch (e) {}
  if (!chosen) chosen = __sui_i18n_detect();
  if (!chosen) chosen = __SUI_I18N.fallback;
  __sui_i18n_apply(chosen);
}
function __sui_i18n_set(locale, persist) {
  var want = String(locale || '');
  var chosen = want.toLowerCase() === 'auto' ? __sui_i18n_detect() : __sui_i18n_match(want);
  if (!chosen) return false;
  var changed = chosen !== __SUI_I18N.locale;
  __sui_i18n_apply(chosen);
  if (persist) { try { localStorage.setItem('sui.locale', chosen); } catch (e) {} }
  if (changed && typeof __sui_render === 'function') __sui_render();
  return true;
}
function __sui_i18n_locale() { return __SUI_I18N.locale || __SUI_I18N.fallback; }
function __sui_i18n_detect() {
  var langs = [];
  try { langs = navigator.languages || [navigator.language || '']; } catch (e) {}
  for (var i = 0; i < langs.length; i++) { var m = __sui_i18n_match(langs[i]); if (m) return m; }
  return '';
}
function __sui_i18n_available() { return Object.keys(__SUI_I18N.catalogs || {}); }
function __sui_i18n_dir() { return __SUI_I18N.dir || (__sui_i18n_rtl(__sui_i18n_locale()) ? 'rtl' : 'ltr'); }
function __sui_i18n_entry(key) {
  var k = String(key == null ? '' : key);
  var cur = __SUI_I18N.catalogs[__SUI_I18N.locale];
  if (cur && Object.prototype.hasOwnProperty.call(cur, k)) return cur[k];
  var fb = __SUI_I18N.catalogs[__SUI_I18N.fallback];
  if (fb && Object.prototype.hasOwnProperty.call(fb, k)) return fb[k];
  return null;
}
function __sui_i18n_has(key) { return __sui_i18n_entry(key) != null; }
function __sui_i18n_missing() {
  var fb = __SUI_I18N.catalogs[__SUI_I18N.fallback] || {};
  var cur = __SUI_I18N.catalogs[__SUI_I18N.locale] || {};
  var out = [];
  for (var k in fb) if (Object.prototype.hasOwnProperty.call(fb, k) && !Object.prototype.hasOwnProperty.call(cur, k)) out.push(k);
  return out;
}
function __sui_i18n_vars(text, vars) {
  if (!vars) return text;
  return String(text).replace(/\{([A-Za-z0-9_.-]+)\}/g, function (_, name) {
    var v = vars[name];
    return v === null || v === undefined ? '' : String(v);
  });
}
function __sui_t(key, vars) {
  var k = String(key == null ? '' : key);
  var entry = __sui_i18n_entry(k);
  if (entry === null || entry === undefined) return k;
  if (typeof entry === 'object') {
    var cat = 'other';
    if (vars && vars.count !== null && vars.count !== undefined) {
      try {
        if (typeof Intl !== 'undefined' && Intl.PluralRules) cat = new Intl.PluralRules(__sui_i18n_locale()).select(Number(vars.count));
        else cat = Number(vars.count) === 1 ? 'one' : 'other';
      } catch (e) { cat = Number(vars.count) === 1 ? 'one' : 'other'; }
    }
    entry = entry[cat] !== undefined ? entry[cat] : (entry.other !== undefined ? entry.other : entry.one);
    if (entry === undefined || entry === null) return k;
  }
  return __sui_i18n_vars(String(entry), vars);
}
function __sui_replace_view(el, html) {
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
      idx: -1,
      start: null,
      end: null,
      scrollX: window.scrollX,
      scrollY: window.scrollY
    };
    // A row of a list can hold more than one field: the position of the
    // field inside its row tells which one has the focus, so the focus does
    // not jump to the first field of the row at every re-render.
    if (focusKeyHolder && focusHolder !== focusEl) {
      var sameTag = focusKeyHolder.querySelectorAll(focusEl.tagName.toLowerCase());
      for (var fi = 0; fi < sameTag.length; fi++) {
        if (sameTag[fi] === focusEl) { focusInfo.idx = fi; break; }
      }
    }
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
      if (focusInfo.idx >= 0) {
        // The row holds more than one field: the field that had the focus is
        // the one at the same position.
        var sameTag = fHolder.querySelectorAll(focusInfo.tag.toLowerCase());
        if (focusInfo.idx < sameTag.length) fTarget = sameTag[focusInfo.idx];
      }
      if (!fTarget) {
        if (fHolder.tagName === focusInfo.tag) fTarget = fHolder;
        else fTarget = fHolder.querySelector(focusInfo.tag.toLowerCase());
      }
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
// Desktop native picker bridge. The dialog module is desktop-only; the web
// branch of stdBodyJs never calls this helper.
function __sui_dialog_pick(endpoint, multiple) {
  const fd = new URLSearchParams();
  fd.set('token', __SUI_TOKEN);
  if (multiple) fd.set('multiple', '1');
  return fetch(endpoint, { method: 'POST', body: fd }).then(async (r) => {
    const j = await r.json();
    if (!j.ok) return [null, j.message || 'dialog failed'];
    return [j.data || [], null];
  }).catch((e) => [null, String(e)]);
}
// __sui_navigate switches between non-invoked root views (pages). It is a
// no-op in single-page apps and for unknown page names.
function __sui_navigate(name) {
  var p = String(name === null || name === undefined ? '' : name);
  if (!p || typeof __sui_page !== 'function') return;
  if (!__SUI_views[p]) return;
  __sui_page(p);
}
function __sui_vm(key, name) {
  const o = __SUI_inst[key];
  const f = (o || {})[name];
  if (f) f();
}
// __sui_esc escapes interpolated values for HTML text and attribute context.
// Views always go through it, so untrusted data cannot inject markup.
function __sui_esc(v) {
  return String(v === null || v === undefined ? '' : v).replace(/[&<>"']/g, function (c) {
    return c === '&' ? '&amp;' : c === '<' ? '&lt;' : c === '>' ? '&gt;' : c === '"' ? '&quot;' : '&#39;';
  });
}
// __sui_class keeps only class tokens declared by the app (its CSS and
// static class attributes), so data cannot activate arbitrary classes.
function __sui_class(v) {
  var toks = String(v === null || v === undefined ? '' : v).split(/\s+/);
  var out = [];
  for (var i = 0; i < toks.length; i++) {
    if (toks[i] && __SUI_CLASSES[toks[i]]) out.push(toks[i]);
  }
  return out.join(' ');
}
// __sui_url allows only safe URL schemes and relative URLs. It blocks
// javascript:, vbscript:, data: and other active schemes; use an allow-list
// in the app when the destination itself is untrusted.
function __sui_url(v) {
  var s = String(v === null || v === undefined ? '' : v).trim();
  var probe = s.replace(/[\u0000-\u0020\u007f]+/g, '');
  if (probe === '') return '';
  if (/^(https?:|mailto:|tel:)/i.test(probe)) return s;
  if (/^[a-z][a-z0-9+.-]*:/i.test(probe)) return '#';
  return s;
}
// __sui_img_url is the image-source variant of __sui_url. Raster images and
// GIFs are embedded as data:image/... URIs by the designer asset manager;
// scripts never run in an <img>, so those values are safe there.
function __sui_img_url(v) {
  var s = String(v === null || v === undefined ? '' : v).trim();
  var probe = s.replace(/[\u0000-\u0020\u007f]+/g, '');
  if (/^data:image\//i.test(probe)) return s;
  return __sui_url(s);
}
// __sui_iframe_src validates iframe URLs. Malformed data is rejected with a
// reason that the renderer shows in a visible alert (never loaded).
function __sui_iframe_src(v, allow, required) {
  var s = String(v === null || v === undefined ? '' : v).trim();
  if (!s) return ['', 'missing URL'];
  if (s.length > 4096) return ['', 'URL too long'];
  for (var i = 0; i < s.length; i++) {
    var c = s.charCodeAt(i);
    if (c <= 32 || c === 127) return ['', 'invalid URL'];
  }
  if (s.indexOf('\\') >= 0) return ['', 'invalid URL'];
  var u;
  try { u = new URL(s, location.origin); } catch(e) { return ['', 'invalid URL']; }
  var scheme = u.protocol.toLowerCase();
  if (scheme !== 'http:' && scheme !== 'https:') return ['', 'scheme not allowed'];
  if (u.username || u.password) return ['', 'credentials are not allowed'];
  var items = String(allow === null || allow === undefined ? '' : allow).split(/[\s,]+/);
  var any = false;
  for (var j = 0; j < items.length; j++) {
    var a = items[j].trim();
    if (!a) continue;
    any = true;
    if (a.charAt(0) === '/') {
      if (a.charAt(1) === '/' || a.indexOf('?') >= 0 || a.indexOf('#') >= 0) continue;
      if (u.origin !== location.origin) continue;
      var ap = a;
      if (ap.length > 1 && ap.charAt(ap.length - 1) === '/') ap = ap.slice(0, -1);
      if (ap === '/' || u.pathname === ap || u.pathname.indexOf(ap + '/') === 0) return [u.href, null];
      continue;
    }
    var p;
    try { p = new URL(a); } catch(e) { continue; }
    if (p.protocol !== u.protocol || p.host !== u.host) continue;
    if (p.username || p.password) continue;
    var pp = p.pathname || '/';
    if (pp !== '/') {
      if (u.pathname.indexOf(pp) !== 0) continue;
      if (pp.charAt(pp.length - 1) !== '/' && u.pathname.length > pp.length && u.pathname.charAt(pp.length) !== '/') continue;
    }
    return [u.href, null];
  }
  if (!any && !required) return [u.href, null];
  return ['', 'URL not allowed'];
}
function __sui_iframe_error(reason) {
  return '<div class="sui-iframe-error" role="alert" data-sui-iframe-error>Cannot accept the data: ' +
    __sui_esc(reason || 'invalid URL') + '</div>';
}
function __sui_pdf_error(reason) {
  return '<div class="sui-iframe-error" role="alert" data-sui-pdf-error>Cannot display the document: ' +
    __sui_esc(reason || 'invalid URL') + '</div>';
}
function __sui_pdf_src(v, allow, required, blob) {
  var s = String(v === null || v === undefined ? '' : v).trim();
  if (s.slice(0, 5).toLowerCase() !== 'blob:') return __sui_iframe_src(v, allow, required);
  if (!blob) return ['', 'blob URLs are not allowed'];
  var inner;
  try {
    var u = new URL(s);
    inner = new URL(u.pathname);
  } catch (e) { return ['', 'invalid URL']; }
  if (inner.origin !== location.origin) return ['', 'URL not allowed'];
  var checked = __sui_iframe_src(inner.href, allow, required);
  if (checked[1]) return checked;
  return [s, null];
}
// __sui_url_guard is the SSRF guard used by http.getSafe/postSafe/
// getStatusSafe and ws.connectSafe. The allow-list is a developer-written
// literal: absolute origins (optionally with a path prefix), ws(s) origins
// for WebSockets, or same-origin path prefixes starting with '/'.
function __sui_url_guard(url, allow, ws) {
  var raw = String(url === null || url === undefined ? '' : url).trim();
  if (!raw) return ['', 'empty URL'];
  if (/[\u0000-\u001f\u007f]/.test(raw)) return ['', 'invalid URL'];
  var u;
  try { u = new URL(raw, location.origin); } catch(e) { return ['', 'invalid URL']; }
  var scheme = u.protocol.toLowerCase();
  if (ws) {
    if (scheme !== 'ws:' && scheme !== 'wss:') return ['', 'only ws/wss URLs are allowed'];
  } else if (scheme !== 'http:' && scheme !== 'https:') {
    return ['', 'only http/https URLs are allowed'];
  }
  if (u.username || u.password) return ['', 'URL credentials are not allowed'];
  var items = String(allow === null || allow === undefined ? '' : allow).split(/[\s,]+/);
  for (var i = 0; i < items.length; i++) {
    var a = items[i].trim();
    if (!a) continue;
    if (a.charAt(0) === '/' && !ws) {
      var prefix = location.origin + a;
      var target = u.origin + u.pathname + u.search + u.hash;
      if (target.indexOf(prefix) === 0) return [u.href, null];
      continue;
    }
    var p;
    try { p = new URL(a); } catch(e) { continue; }
    if (p.protocol !== u.protocol || p.host !== u.host) continue;
    var pp = p.pathname || '/';
    if (pp !== '/') {
      if (u.pathname.indexOf(pp) !== 0) continue;
      if (pp.charAt(pp.length - 1) !== '/' && u.pathname.length > pp.length && u.pathname.charAt(pp.length) !== '/') continue;
    }
    return [u.href, null];
  }
  return ['', 'URL is not in the allow-list'];
}
// __sui_play_sound plays a sound effect. The URL is checked like the
// other network calls: only http/https, without credentials, and
// against the allow-list when the caller uses the *Safe variant.
function __sui_play_sound(url, allow) {
  return new Promise(function (resolve) {
    var raw = String(url === null || url === undefined ? '' : url).trim();
    if (!raw) { resolve([false, 'empty URL']); return; }
    var href = raw;
    if (allow !== null && allow !== undefined && String(allow) !== '') {
      var g = __sui_url_guard(raw, allow, false);
      if (g[1]) { resolve([false, g[1]]); return; }
      href = g[0];
    } else {
      var u;
      try { u = new URL(raw, location.origin); } catch (e) { resolve([false, 'invalid URL']); return; }
      if (u.protocol !== 'http:' && u.protocol !== 'https:') { resolve([false, 'only http/https URLs are allowed']); return; }
      if (u.username || u.password) { resolve([false, 'URL credentials are not allowed']); return; }
      href = u.href;
    }
    try {
      var a = new Audio();
      a.preload = 'auto';
      a.src = href;
      var p = a.play();
      if (p && p.then) {
        p.then(function () { resolve([true, null]); }, function (e) { resolve([false, String(e)]); });
      } else {
        resolve([true, null]);
      }
    } catch (e) {
      resolve([false, String(e)]);
    }
  });
}
function __sui_clean_headers(h) {
  if (!h) return {};
  if (typeof h === 'string') { try { h = JSON.parse(h); } catch(e) { return {}; } }
  if (typeof h !== 'object') return {};
  var out = {};
  for (var k in h) { if (Object.prototype.hasOwnProperty.call(h, k)) out[String(k)] = String(h[k]); }
  return out;
}
function __sui_upload_form(fieldName, fileName, content) {
  var fd = new FormData();
  var blob = new Blob([String(content === null || content === undefined ? '' : content)], { type: 'application/octet-stream' });
  fd.append(fieldName || 'file', blob, fileName || 'upload.bin');
  return fd;
}
async function __sui_fetch_guarded(url, allow, init, timeoutMs) {
  var g = __sui_url_guard(url, allow, false);
  if (g[1]) return [null, g[1]];
  var opts = init || {};
  opts.redirect = 'error';
  var tLimit = (typeof timeoutMs === 'number' && timeoutMs > 0) ? timeoutMs : 30000;
  var ctrl = null, timer = null;
  if (typeof AbortController !== 'undefined') { ctrl = new AbortController(); opts.signal = ctrl.signal; timer = setTimeout(function(){ ctrl.abort(); }, tLimit); }
  try {
    var r = await fetch(g[0], opts);
    var len = Number(r.headers.get('content-length') || 0);
    if (len > 10485760) { if (r.body && r.body.cancel) { try { r.body.cancel(); } catch(e) {} } return [null, 'response too large']; }
    var text = await r.text();
    if (text.length > 10485760) return [null, 'response too large'];
    return [text, null];
  } catch(e) {
    if (ctrl && ctrl.signal && ctrl.signal.aborted) return [null, 'request timed out after ' + tLimit + 'ms'];
    return [null, String(e)];
  } finally {
    if (timer) clearTimeout(timer);
  }
}
async function __sui_fetch_bytes_guarded(url, allow) {
  var g = __sui_url_guard(url, allow, false);
  if (g[1]) return [null, g[1]];
  var opts = { redirect: 'error' };
  var ctrl = null, timer = null;
  if (typeof AbortController !== 'undefined') { ctrl = new AbortController(); opts.signal = ctrl.signal; timer = setTimeout(function(){ ctrl.abort(); }, 30000); }
  try {
    var r = await fetch(g[0], opts);
    var len = Number(r.headers.get('content-length') || 0);
    if (len > 10485760) { if (r.body && r.body.cancel) { try { r.body.cancel(); } catch(e) {} } return [null, 'response too large']; }
    var buf = await r.arrayBuffer();
    if (buf.byteLength > 10485760) return [null, 'response too large'];
    var v = new Uint8Array(buf);
    var out = new Array(v.length);
    for (var i = 0; i < v.length; i++) out[i] = v[i];
    return [out, null];
  } catch(e) {
    if (ctrl && ctrl.signal && ctrl.signal.aborted) return [null, 'request timed out after 30000ms'];
    return [null, String(e)];
  } finally {
    if (timer) clearTimeout(timer);
  }
}
async function __sui_fetch_status_guarded(url, allow) {
  var g = __sui_url_guard(url, allow, false);
  if (g[1]) return [null, g[1]];
  var opts = { redirect: 'error' };
  var ctrl = null, timer = null;
  if (typeof AbortController !== 'undefined') { ctrl = new AbortController(); opts.signal = ctrl.signal; timer = setTimeout(function(){ ctrl.abort(); }, 30000); }
  try {
    var r = await fetch(g[0], opts);
    if (r.body && r.body.cancel) { try { r.body.cancel(); } catch(e) {} }
    return [r.status, null];
  } catch(e) {
    return [null, String(e)];
  } finally {
    if (timer) clearTimeout(timer);
  }
}
// __sui_css sanitizes a dynamic style value: comma-free declarations only,
// resource loading (url(), paint(), @import), CSS escapes, expression() and
// positioning/overlay properties are dropped.
function __sui_css(v) {
  var text = String(v === null || v === undefined ? '' : v).replace(/\/\*[\s\S]*?\*\//g, '');
  var blocked = { position: 1, top: 1, right: 1, bottom: 1, left: 1, inset: 1, 'z-index': 1, content: 1, 'pointer-events': 1, all: 1, behavior: 1, quotes: 1 };
  var out = [];
  var parts = text.split(';');
  for (var i = 0; i < parts.length; i++) {
    var p = parts[i].trim();
    if (!p) continue;
    var idx = p.indexOf(':');
    if (idx <= 0) continue;
    var name = p.slice(0, idx).trim().toLowerCase();
    var val = p.slice(idx + 1).trim();
    if (!/^(?:--[a-z][a-z0-9-]*|[a-z][a-z0-9-]*)$/.test(name)) continue;
    if (name.indexOf('inset-') === 0) continue;
    if (blocked[name]) continue;
    if (!val) continue;
    if (/url\s*[^a-z0-9_-]|expression\s*[^a-z0-9_-]|javascript\s*:|vbscript\s*:|@import|behavior\s*:|paint\s*[^a-z0-9_-]|image-set\s*[^a-z0-9_-]|cross-fade\s*[^a-z0-9_-]|image\s*[^a-z0-9_-]|element\s*[^a-z0-9_-]|attr\s*[^a-z0-9_-]|var\s*[^a-z0-9_-]|\\/i.test(val)) continue;
    out.push(name + ': ' + val);
  }
  return out.join('; ');
}
function __sui_ac(input, options) {
  if (!input || !options || options[0] === undefined) return;
  const box = document.createElement('div');
  box.className = 'sui-ac';
  box.style.display = 'none';
  document.body.appendChild(box);
  let active = -1;
  let count = 0;
  function hide() { box.style.display = 'none'; }
  function show() {
    const r = input.getBoundingClientRect();
    box.style.top = (r.bottom + window.scrollY) + 'px';
    box.style.left = (r.left + window.scrollX) + 'px';
    box.style.minWidth = r.width + 'px';
    box.style.display = 'block';
  }
  function render() {
    const q = (input.value || '').toLowerCase();
    box.innerHTML = '';
    active = -1;
    count = 0;
    options.forEach(it => {
      if (it.toLowerCase().indexOf(q) < 0) return;
      count++;
      const d = document.createElement('div');
      d.className = 'sui-ac-item';
      d.textContent = it;
      d.addEventListener('mousedown', ev => { ev.preventDefault(); input.value = it; hide(); });
      box.appendChild(d);
    });
  }
  function highlight() {
    const items = box.querySelectorAll('.sui-ac-item');
    let i = 0;
    items.forEach(d => { d.classList.toggle('active', i === active); i++; });
  }
  input.addEventListener('focus', () => { render(); show(); });
  input.addEventListener('input', () => { render(); show(); });
  input.addEventListener('blur', () => setTimeout(hide, 150));
  input.addEventListener('keydown', ev => {
    if (!count) return;
    const items = box.querySelectorAll('.sui-ac-item');
    if (ev.key === 'ArrowDown') { ev.preventDefault(); active = Math.min(active + 1, count - 1); highlight(); }
    else if (ev.key === 'ArrowUp') { ev.preventDefault(); active = Math.max(active - 1, 0); highlight(); }
    else if (ev.key === 'Enter') { ev.preventDefault(); if (active >= 0) input.value = items[active].textContent; hide(); }
    else if (ev.key === 'Escape') { hide(); }
  });
}

function __sui_timers_debounce(fn, ms) {
  let t = null;
  const wrapped = () => {
    if (t) clearTimeout(t);
    t = setTimeout(() => fn(), ms || 0);
  };
  wrapped._id = Date.now() + Math.floor(Math.random()*1000);
  return setTimeout(() => wrapped(), 0);
}
function __sui_timers_throttle(fn, ms) {
  let last = 0;
  const wrapped = () => {
    const now = Date.now();
    if (now - last >= (ms||0)) { last = now; fn(); }
  };
  return setTimeout(() => wrapped(), 0);
}

function __sui_time_now() { return new Date(); }
function __sui_time_nowMs() { return Date.now(); }
function __sui_time_format(dt, fmt) {
  const d = dt instanceof Date ? dt : (dt ? new Date(dt) : new Date());
  if (isNaN(d.getTime())) return '';
  if (!fmt || fmt === 'iso') return d.toISOString();
  const p = (fmt || '').replace('YYYY', String(d.getFullYear()))
    .replace('MM', String(d.getMonth()+1).padStart(2,'0'))
    .replace('DD', String(d.getDate()).padStart(2,'0'))
    .replace('HH', String(d.getHours()).padStart(2,'0'))
    .replace('mm', String(d.getMinutes()).padStart(2,'0'))
    .replace('ss', String(d.getSeconds()).padStart(2,'0'));
  return p;
}
function __sui_time_parse(s, fmt) {
  try {
    if (!fmt || fmt === 'iso') { const d = s ? new Date(s) : null; if (!d || isNaN(d.getTime())) return [null, 'parse error']; return [d, null]; }
    const m = s.match(/(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})/);
    if (!m) return [null, 'format mismatch'];
    const d = new Date(+m[1], +m[2]-1, +m[3], +m[4], +m[5], +m[6]);
    if (isNaN(d.getTime())) return [null, 'parse error'];
    return [d, null];
  } catch(e){ return [null, String(e)]; }
}
function __sui_time_add(dt, amt, unit) {
  const d = dt instanceof Date ? dt : new Date(dt||Date.now());
  const v = Number(amt)||0; const u = (unit||'ms').toLowerCase();
  if (u==='ms') d.setMilliseconds(d.getMilliseconds()+v);
  else if (u==='s'||u==='sec'||u==='second') d.setSeconds(d.getSeconds()+v);
  else if (u==='m'||u==='min'||u==='minute') d.setMinutes(d.getMinutes()+v);
  else if (u==='h'||u==='hour') d.setHours(d.getHours()+v);
  else if (u==='d'||u==='day') d.setDate(d.getDate()+v);
  else if (u==='w'||u==='week') d.setDate(d.getDate()+v*7);
  else if (u==='mo'||u==='month') d.setMonth(d.getMonth()+v);
  else if (u==='y'||u==='year') d.setFullYear(d.getFullYear()+v);
  else d.setMilliseconds(d.getMilliseconds()+v);
  return d;
}
function __sui_time_diff(a,b,unit){
  const da=a instanceof Date?a:new Date(a||0); const db=b instanceof Date?b:new Date(b||0);
  let ms=Math.max(0,da.getTime()-db.getTime()); const u=(unit||'ms').toLowerCase();
  if(u==='s'||u==='sec') return Math.floor(ms/1000);
  if(u==='m'||u==='min') return Math.floor(ms/60000);
  if(u==='h'||u==='hour') return Math.floor(ms/3600000);
  if(u==='d'||u==='day') return Math.floor(ms/86400000);
  return ms;
}
function __sui_time_toISO(dt){ const d=dt instanceof Date?dt:new Date(dt||0); return isNaN(d.getTime())?'':d.toISOString(); }
function __sui_time_toMs(dt){ const d=dt instanceof Date?dt:new Date(dt||0); return isNaN(d.getTime())?0:d.getTime(); }

function __sui_url_encode(s){ try{return encodeURI(String(s||''))}catch(e){return String(s||'')} }
function __sui_url_decode(s){ try{return [decodeURI(String(s||'')), null]}catch(e){return ['', String(e)]} }
function __sui_url_decodeComponent(s){ try{return [decodeURIComponent(String(s||'')), null]}catch(e){return ['', String(e)]} }
function __sui_b64_encode(s){ try{return btoa(unescape(encodeURIComponent(String(s||''))))}catch(e){return ''} }
function __sui_b64_decode(s){ try{return [decodeURIComponent(escape(atob(String(s||'')))) , null]}catch(e){return ['', String(e)]} }
function __sui_bytes_to_b64(bytes){ var s=''; var a=bytes||[]; for(var i=0;i<a.length;i++) s+=String.fromCharCode(a[i]&255); try{return btoa(s)}catch(e){return ''} }
function __sui_b64_to_bytes(s){ try{ var t=atob(String(s||'')); var a=[]; for(var i=0;i<t.length;i++) a.push(t.charCodeAt(i)); return a; }catch(e){ return []; } }

function __sui_path_join() {
  var parts = [];
  for (var i = 0; i < arguments.length; i++) { var v = arguments[i]; if (v === undefined || v === null || v === '') continue; parts.push(String(v)); }
  var all = parts.join('/').replace(/\\/g, '/');
  var drive = ''; var dm = /^([A-Za-z]:)(\/|$)/.exec(all); if (dm) { drive = dm[1]; all = all.slice(drive.length); }
  var abs = all.charAt(0) === '/'; var segs = all.split('/'); var out = [];
  for (var j = 0; j < segs.length; j++) {
    var seg = segs[j];
    if (seg === '' || seg === '.') continue;
    if (seg === '..') { if (out.length && out[out.length - 1] !== '..') out.pop(); else if (!abs) out.push('..'); }
    else out.push(seg);
  }
  var res = out.join('/'); if (abs) res = '/' + res; if (drive) res = drive + (res && res.charAt(0) !== '/' ? '/' : '') + res;
  if (!res) res = abs ? '/' : (drive ? drive + '/' : '.');
  return res;
}
function __sui_path_dirname(p) {
  var v = String(p == null ? '' : p).replace(/\\/g, '/').replace(/\/+$/, '');
  if (!v) return '.';
  var drive = ''; var dm = /^([A-Za-z]:)(\/|$)/.exec(v); if (dm) { drive = dm[1]; if (v === drive) return drive + '/'; }
  var idx = v.lastIndexOf('/');
  if (idx < 0) return drive ? drive + '/' : '.';
  if (idx === 0) return '/';
  var d = v.slice(0, idx); if (drive && d === drive) return drive + '/'; return d || '/';
}
function __sui_path_basename(p) {
  var v = String(p == null ? '' : p).replace(/\\/g, '/').replace(/\/+$/, '');
  var idx = v.lastIndexOf('/'); return idx >= 0 ? v.slice(idx + 1) : v;
}
function __sui_path_extname(p) {
  var b = __sui_path_basename(p); var idx = b.lastIndexOf('.');
  return idx > 0 ? b.slice(idx) : '';
}
function __sui_crypto_status() {
  try {
    var c = typeof crypto !== 'undefined' ? crypto : null;
    var hasRandom = !!c && typeof c.getRandomValues === 'function';
    var hasHash = !!c && !!c.subtle && typeof c.subtle.digest === 'function';
    if (!hasRandom) return 'unavailable';
    if (!hasHash) return 'random-only';
    return 'ready';
  } catch (e) { return 'unavailable'; }
}
function __sui_crypto_available() { return __sui_crypto_status() === 'ready'; }
function __sui_crypto_uuid() {
  try { if (typeof crypto !== 'undefined' && crypto.randomUUID) return String(crypto.randomUUID()); } catch (e) {}
  if (__sui_crypto_status() === 'unavailable') throw new Error('crypto.uuid requires crypto.getRandomValues (Web Crypto is unavailable)');
  try {
    var b = new Uint8Array(16); crypto.getRandomValues(b); b[6] = (b[6] & 15) | 64; b[8] = (b[8] & 63) | 128;
    var h = []; for (var i = 0; i < b.length; i++) h.push((b[i] + 256).toString(16).slice(1));
    return h.slice(0,4).join('') + '-' + h.slice(4,6).join('') + '-' + h.slice(6,8).join('') + '-' + h.slice(8,10).join('') + '-' + h.slice(10).join('');
  } catch (e) { throw new Error('crypto.uuid failed: ' + (e && e.message ? e.message : String(e))); }
}
function __sui_crypto_random_bytes(n) {
  var size = Math.floor(Number(n));
  if (!(size > 0) || size > 1024) throw new Error('crypto.randomBytes requires a length between 1 and 1024');
  if (__sui_crypto_status() === 'unavailable') throw new Error('crypto.randomBytes requires crypto.getRandomValues (Web Crypto is unavailable)');
  try { var b = new Uint8Array(size); crypto.getRandomValues(b); return Array.prototype.slice.call(b); }
  catch (e) { throw new Error('crypto.randomBytes failed: ' + (e && e.message ? e.message : String(e))); }
}
async function __sui_crypto_hash(algo, text) {
  var status = __sui_crypto_status();
  if (status !== 'ready') {
    if (status === 'random-only') throw new Error('crypto.hash requires a secure context (HTTPS or localhost): crypto.subtle is unavailable');
    throw new Error('crypto.hash is unavailable: Web Crypto is not supported');
  }
  var name = String(algo == null ? 'sha256' : algo).toUpperCase().replace(/-/g, '');
  if (name === 'SHA2' || name === 'SHA256') name = 'SHA-256';
  else if (name === 'SHA384') name = 'SHA-384';
  else if (name === 'SHA512') name = 'SHA-512';
  else throw new Error('crypto.hash: unsupported algorithm "' + String(algo) + '" (use sha256, sha384 or sha512)');
  try {
    var data = new TextEncoder().encode(String(text == null ? '' : text));
    var buf = await crypto.subtle.digest(name, data);
    var out = []; var bytes = new Uint8Array(buf);
    for (var i = 0; i < bytes.length; i++) out.push((bytes[i] + 256).toString(16).slice(1));
    return out.join('');
  } catch (e) {
    throw new Error('crypto.hash failed: ' + (e && e.message ? e.message : String(e)));
  }
}

// JSON Schema validation: a practical subset (type, properties, required,
// additionalProperties, items, enum/const, bounds, pattern, formats, allOf/anyOf/oneOf/not).
function __sui_schema_kind(v) {
  if (v === null) return 'null';
  if (Array.isArray(v)) return 'array';
  if (typeof v === 'number') return Number.isInteger(v) ? 'integer' : 'number';
  return typeof v;
}
function __sui_schema_equal(a, b) { try { return JSON.stringify(a) === JSON.stringify(b); } catch (e) { return false; } }
function __sui_schema_type_ok(v, t) {
  if (t === 'any') return true;
  if (t === 'null') return v === null;
  if (t === 'array') return Array.isArray(v);
  if (t === 'object') return v !== null && typeof v === 'object' && !Array.isArray(v);
  if (t === 'integer') return typeof v === 'number' && Number.isInteger(v);
  if (t === 'number') return typeof v === 'number' && !isNaN(v);
  return typeof v === t;
}
function __sui_schema_check(v, s, path, depth) {
  if (depth > 64) return path + ': schema nesting is too deep';
  if (s === true) return null;
  if (s === false) return path + ': rejected by false schema';
  if (s === null || typeof s !== 'object' || Array.isArray(s)) return path + ': schema must be an object or boolean';
  if (s.$ref !== undefined) return path + ': $ref is not supported';
  if (s.type !== undefined) {
    var types = Array.isArray(s.type) ? s.type : [s.type];
    var ok = false;
    for (var ti = 0; ti < types.length; ti++) if (__sui_schema_type_ok(v, types[ti])) { ok = true; break; }
    if (!ok) return path + ': expected ' + types.join(' or ') + ', got ' + __sui_schema_kind(v);
  }
  if (s.enum !== undefined) {
    if (!Array.isArray(s.enum)) return path + ': enum must be an array';
    var found = false;
    for (var ei = 0; ei < s.enum.length; ei++) if (__sui_schema_equal(v, s.enum[ei])) { found = true; break; }
    if (!found) return path + ': value is not in enum';
  }
  if (Object.prototype.hasOwnProperty.call(s, 'const') && !__sui_schema_equal(v, s.const)) return path + ': value does not match const';
  if (typeof v === 'number') {
    if (typeof s.minimum === 'number' && v < s.minimum) return path + ': must be >= ' + s.minimum;
    if (typeof s.maximum === 'number' && v > s.maximum) return path + ': must be <= ' + s.maximum;
    if (typeof s.exclusiveMinimum === 'number' && v <= s.exclusiveMinimum) return path + ': must be > ' + s.exclusiveMinimum;
    if (typeof s.exclusiveMaximum === 'number' && v >= s.exclusiveMaximum) return path + ': must be < ' + s.exclusiveMaximum;
    if (typeof s.multipleOf === 'number' && s.multipleOf > 0) {
      var q = v / s.multipleOf;
      if (Math.abs(q - Math.round(q)) > 1e-9) return path + ': must be a multiple of ' + s.multipleOf;
    }
  }
  if (typeof v === 'string') {
    if (typeof s.minLength === 'number' && v.length < s.minLength) return path + ': must have at least ' + s.minLength + ' characters';
    if (typeof s.maxLength === 'number' && v.length > s.maxLength) return path + ': must have at most ' + s.maxLength + ' characters';
    if (typeof s.pattern === 'string') {
      if (s.pattern.length > 512) return path + ': pattern is too long';
      try { if (!(new RegExp(s.pattern)).test(v)) return path + ': does not match the pattern'; } catch (e) { return path + ': invalid pattern'; }
    }
    if (typeof s.format === 'string') {
      if (s.format === 'email' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) return path + ': must be an email';
      if ((s.format === 'uri' || s.format === 'url') && !__sui_valid_url(v)) return path + ': must be a URI';
      if (s.format === 'uuid' && !/^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$/.test(v)) return path + ': must be a UUID';
      if (s.format === 'date' && !/^\d{4}-\d{2}-\d{2}$/.test(v)) return path + ': must be a date';
      if (s.format === 'date-time' && isNaN(Date.parse(v))) return path + ': must be a date-time';
    }
  }
  if (Array.isArray(v)) {
    if (typeof s.minItems === 'number' && v.length < s.minItems) return path + ': must have at least ' + s.minItems + ' items';
    if (typeof s.maxItems === 'number' && v.length > s.maxItems) return path + ': must have at most ' + s.maxItems + ' items';
    if (s.uniqueItems === true) {
      var seen = {};
      for (var ui = 0; ui < v.length; ui++) {
        var ukey = JSON.stringify(v[ui]);
        if (seen[ukey]) return path + ': items must be unique';
        seen[ukey] = true;
      }
    }
    if (Array.isArray(s.items)) {
      for (var ai = 0; ai < s.items.length && ai < v.length; ai++) {
        var aerr = __sui_schema_check(v[ai], s.items[ai], path + '[' + ai + ']', depth + 1);
        if (aerr !== null) return aerr;
      }
      if (v.length > s.items.length) {
        if (s.additionalItems === false) return path + ': additional items are not allowed';
        if (s.additionalItems && typeof s.additionalItems === 'object') {
          for (var aj = s.items.length; aj < v.length; aj++) {
            var ajerr = __sui_schema_check(v[aj], s.additionalItems, path + '[' + aj + ']', depth + 1);
            if (ajerr !== null) return ajerr;
          }
        }
      }
    } else if (s.items !== undefined) {
      for (var ii = 0; ii < v.length; ii++) {
        var ierr = __sui_schema_check(v[ii], s.items, path + '[' + ii + ']', depth + 1);
        if (ierr !== null) return ierr;
      }
    }
  } else if (v !== null && typeof v === 'object') {
    if (typeof s.minProperties === 'number' && Object.keys(v).length < s.minProperties) return path + ': must have at least ' + s.minProperties + ' properties';
    if (typeof s.maxProperties === 'number' && Object.keys(v).length > s.maxProperties) return path + ': must have at most ' + s.maxProperties + ' properties';
    if (Array.isArray(s.required)) {
      for (var ri = 0; ri < s.required.length; ri++) {
        if (!Object.prototype.hasOwnProperty.call(v, s.required[ri])) return path + ': missing required property "' + s.required[ri] + '"';
      }
    }
    var props = (s.properties && typeof s.properties === 'object') ? s.properties : {};
    var keys = Object.keys(v);
    for (var ki = 0; ki < keys.length; ki++) {
      var key = keys[ki];
      if (props[key] !== undefined) {
        var perr = __sui_schema_check(v[key], props[key], path + '.' + key, depth + 1);
        if (perr !== null) return perr;
      } else if (s.additionalProperties === false) {
        return path + ': unexpected property "' + key + '"';
      } else if (s.additionalProperties && typeof s.additionalProperties === 'object') {
        var xerr = __sui_schema_check(v[key], s.additionalProperties, path + '.' + key, depth + 1);
        if (xerr !== null) return xerr;
      }
    }
  }
  if (Array.isArray(s.allOf)) {
    for (var oi = 0; oi < s.allOf.length; oi++) {
      var oerr = __sui_schema_check(v, s.allOf[oi], path, depth + 1);
      if (oerr !== null) return oerr;
    }
  }
  if (Array.isArray(s.anyOf)) {
    var anyOk = false;
    for (var yi = 0; yi < s.anyOf.length; yi++) if (__sui_schema_check(v, s.anyOf[yi], path, depth + 1) === null) { anyOk = true; break; }
    if (!anyOk) return path + ': does not match anyOf';
  }
  if (Array.isArray(s.oneOf)) {
    var oneCount = 0;
    for (var wi = 0; wi < s.oneOf.length; wi++) if (__sui_schema_check(v, s.oneOf[wi], path, depth + 1) === null) oneCount++;
    if (oneCount !== 1) return path + ': must match exactly one oneOf schema';
  }
  if (s.not !== undefined) {
    if (__sui_schema_check(v, s.not, path, depth + 1) === null) return path + ': must not match the not schema';
  }
  if (s.if !== undefined) {
    var cond = __sui_schema_check(v, s.if, path, depth + 1);
    if (cond === null && s.then !== undefined) {
      var terr = __sui_schema_check(v, s.then, path, depth + 1);
      if (terr !== null) return terr;
    } else if (cond !== null && s.else !== undefined) {
      var eerr = __sui_schema_check(v, s.else, path, depth + 1);
      if (eerr !== null) return eerr;
    }
  }
  return null;
}
function __sui_json_validate(value, schema) {
  try { if (typeof schema === 'string') schema = JSON.parse(schema); } catch (e) { return [false, '$: schema is not valid JSON']; }
  var err = __sui_schema_check(value, schema, '$', 0);
  return err === null ? [true, ''] : [false, err];
}

const __SUI_MEM_STORE = Object.create(null);
var __SUI_STORAGE_SYNC = false;
function __sui_storage_get(k, def) {
  var key = String(k); var fallback = def === undefined ? null : def;
  try { var v = localStorage.getItem(key); if (v === null) return fallback; try { return JSON.parse(v); } catch (e) { return v; } }
  catch (e) { if (key in __SUI_MEM_STORE) { try { return JSON.parse(__SUI_MEM_STORE[key]); } catch (ee) { return __SUI_MEM_STORE[key]; } } return fallback; }
}
function __sui_storage_set(k, v) {
  var key = String(k); var sv = typeof v === 'string' ? v : JSON.stringify(v);
  try { localStorage.setItem(key, sv); return true; } catch (e) { __SUI_MEM_STORE[key] = sv; return false; }
}
function __sui_storage_remove(k) {
  var key = String(k);
  try { localStorage.removeItem(key); return true; } catch (e) { delete __SUI_MEM_STORE[key]; return false; }
}
function __sui_storage_clear() {
  try { localStorage.clear(); return true; } catch (e) { for (var kk in __SUI_MEM_STORE) delete __SUI_MEM_STORE[kk]; return false; }
}
function __sui_storage_keys() {
  try { var r = []; for (var i = 0; i < localStorage.length; i++) { var k = localStorage.key(i); if (k) r.push(k); } return r; }
  catch (e) { return Object.keys(__SUI_MEM_STORE); }
}
async function __sui_storage_quota() {
  var used = 0, quota = 0;
  try { for (var i = 0; i < localStorage.length; i++) { var k = localStorage.key(i) || ''; used += (k.length + String(localStorage.getItem(k) || '').length) * 2; } } catch (e) {}
  try {
    if (navigator.storage && navigator.storage.estimate) {
      var estimate = navigator.storage.estimate();
      var timeout = new Promise(function (resolve) { setTimeout(function () { resolve(null); }, 1000); });
      var est = await Promise.race([estimate, timeout]);
      if (est) { if (est.usage !== undefined) used = Number(est.usage) || used; if (est.quota !== undefined) quota = Number(est.quota) || 0; }
    }
  } catch (e) {}
  return { used: used, quota: quota };
}
function __sui_storage_sync() {
  if (__SUI_STORAGE_SYNC) return true;
  try { if (typeof window === 'undefined' || !window.addEventListener) return false;
    window.addEventListener('storage', function () { if (typeof __sui_render === 'function') __sui_render(); });
    __SUI_STORAGE_SYNC = true; return true;
  } catch (e) { return false; }
}

function __sui_clip_write_multi(items){try{return {written:true};}catch(e){return {written:false,error:String(e)};}}
function __sui_clip_write(t){ return (async()=>{ try{ if(navigator.clipboard&&navigator.clipboard.writeText){ await navigator.clipboard.writeText(String(t||'')); return [true,null] } return [false,'clipboard not available'] }catch(e){ return [false,String(e)] } })(); }

function __sui_valid_email(s){ const v=String(s||''); return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v); }
function __sui_valid_url(s){ try{ new URL(String(s||'')); return true }catch(e){ return false } }
function __sui_valid_empty(v){ if(v===null||v===undefined) return true; if(typeof v==='string') return v.trim()===''; if(Array.isArray(v)) return v.length===0; if(typeof v==='object') return Object.keys(v).length===0; return false; }
function __sui_valid_required(v){ return !__sui_valid_empty(v); }
function __sui_valid_minlen(v,l){ const n=Array.isArray(v)?v.length:(typeof v==='string'?v.length:Object.keys(v||{}).length); return n>=Number(l||0); }
function __sui_valid_maxlen(v,l){ const n=Array.isArray(v)?v.length:(typeof v==='string'?v.length:Object.keys(v||{}).length); return n<=Number(l||0); }
function __sui_valid_length(v,a,b){ const n=Array.isArray(v)?v.length:(typeof v==='string'?v.length:Object.keys(v||{}).length); const lo=Number(a||0), hi=b===undefined?lo:Number(b); if(b===undefined) return n===lo; return n>=lo&&n<=hi; }
function __sui_valid_pattern(s,pat){ try{ const r=new RegExp(pat); return r.test(String(s||'')) }catch(e){ return false } }
function __sui_valid_isnum(v){ return typeof v==='number'&&!isNaN(v); }
function __sui_valid_isbool(v){ return typeof v==='boolean'; }
function __sui_valid_isstr(v){ return typeof v==='string'; }
function __sui_valid_isarr(v){ return Array.isArray(v); }
function __sui_valid_isobj(v){ return v!==null&&typeof v==='object'&&!Array.isArray(v); }

function __sui_regex_flags(f){ if(!f) return 'g'; let s=''; if(f.indexOf('i')>=0) s+='i'; if(f.indexOf('m')>=0) s+='m'; if(f.indexOf('s')>=0) s+='s'; if(f.indexOf('u')>=0) s+='u'; if(f.indexOf('y')>=0) s+='y'; if(s.indexOf('g')<0) s+='g'; return s; }
function __sui_regex(pat,flags){ try{ return new RegExp(String(pat||''), __sui_regex_flags(flags)); }catch(e){ try{return new RegExp(String(pat||''),'g')}catch(ee){ return null } } }
function __sui_replace_regex(s,pat,repl,flags){ const r=__sui_regex(pat,flags||'g'); if(!r) return String(s||''); try{ return String(s||'').replace(r, repl===undefined?'':String(repl)) }catch(e){ return String(s||'') } }
function __sui_match(s,pat,flags){ const r=__sui_regex(pat,flags||''); if(!r) return []; try{ const m=String(s||'').match(r); return m||[] }catch(e){ return [] } }
function __sui_match_all(s,pat,flags){ const f=flags||'g'; if(f.indexOf('g')<0){ const r=__sui_regex(pat,f+'g'); if(!r) return []; try{ const it=String(s||'').matchAll(r); return Array.from(it) }catch(e){ return [] } } const r2=__sui_regex(pat,f); if(!r2) return []; try{ const it=String(s||'').matchAll(r2); return Array.from(it) }catch(e){ return [] } }
function __sui_search(s,pat){ try{ const r=__sui_regex(pat,''); if(!r) return -1; return String(s||'').search(r) }catch(e){ return -1 } }
function __sui_test(s,pat){ try{ const r=__sui_regex(pat,''); if(!r) return false; return r.test(String(s||'')) }catch(e){ return false } }
function __sui_split_regex(s,pat,limit){ try{ const r=__sui_regex(pat,'g'); if(!r) return String(s||'').split(''); return String(s||'').split(r, limit===undefined?undefined:Number(limit)) }catch(e){ return String(s||'').split(String(pat||''), limit===undefined?undefined:Number(limit)) } }
function __sui_extract(s,pat){ const m=__sui_match(s,pat,''); if(m&&m.length>1) return m[1]; if(m&&m.length===1) return m[0]; return '' }
function __sui_extract_all(s,pat){ const ms=__sui_match_all(s,pat,'g'); const out=[]; for(let i=0;i<ms.length;i++){ const mm=ms[i]; if(mm&&mm.length>1) out.push(mm[1]); else if(mm&&mm.length===1) out.push(mm[0]); } return out }

const __SUI_SEC = {maxRows:100000,maxCols:10000,maxText:20000000,csvFormulaEscape:false};

const __SUI_TBL = (function(){
  // ---- data binding --------------------------------------------------------
  // store[name] = { header: [] | null, rows: [[...]] }. It survives view
  // re-renders: afterRender() re-applies it to the freshly rendered table.
  var store = Object.create(null);
  var changes = Object.create(null);
  var tableClip = Object.create(null);
  var tableBooks = Object.create(null);
  var tpls = new WeakMap(); // table element -> { heads, bodies } row templates
  // Resource limits: the table is a DOM widget, so huge payloads are refused
  // instead of freezing the page. Applications should paginate.
  var MAX_ROWS = (__SUI_SEC && __SUI_SEC.maxRows) || 100000;
  var MAX_COLS = (__SUI_SEC && __SUI_SEC.maxCols) || 10000;
  var MAX_TEXT = (__SUI_SEC && __SUI_SEC.maxText) || 20000000;
  // Split so the generated component JS never contains the literal attribute
  // name; component-only apps rely on that (inline handlers, no query ids).
  var TBL_ID_ATTR = 'data-' + 'sui-id';

  function splitList(v){ return String(v || '').split(',').map(function(s){ return s.trim(); }).filter(function(s){ return s !== ''; }); }
  function intList(v){ return splitList(v).map(function(s){ return parseInt(s, 10) || 0; }); }
  function parseColumnOptions(v){
    var out = {};
    var parts = String(v || '').split(';');
    for (var i = 0; i < parts.length; i++) {
      var eq = parts[i].indexOf('=');
      if (eq <= 0) continue;
      var k = parts[i].slice(0, eq).trim();
      var vals = parts[i].slice(eq + 1).split('|').map(function(s){ return s.trim(); });
      if (k) out[k] = vals;
    }
    return out;
  }
  function cfgOf(tbl){
    function attr(k){ return tbl.getAttribute('data-sui-' + k) || ''; }
    var mode = attr('edit-mode');
    if (mode !== 'editable' && mode !== 'persistent') mode = 'none';
    var editable = mode !== 'none';
    var cfg = {
      columns: splitList(attr('columns')),
      labels: splitList(attr('column-labels')),
      widths: intList(attr('column-widths')),
      align: splitList(attr('column-align')),
      types: splitList(attr('column-types')),
      formats: splitList(attr('column-formats')),
      options: parseColumnOptions(attr('column-options')),
      sortable: attr('sortable') === '1',
      filterable: attr('filterable') === '1',
      pageSize: parseInt(attr('page-size'), 10) || 0,
      selection: attr('selection') || 'none',
      multi: attr('multi') === '1',
      editMode: mode,
      editable: editable,
      fixed: attr('fixed') === '1',
      showAddTable: editable && attr('show-add-table') === '1',
      structureEditable: editable && attr('structure-editable') === '1'
    };
    return cfg;
  }
  function tableName(tbl){ return tbl.getAttribute('name') || ''; }
  function firstTable(name){ var ts = findTables(name); return ts.length ? ts[0] : null; }
  function tableValue(tbl, key){
    var a = tbl.getAttribute('data-sui-table-' + key) || '';
    return a;
  }
  function fireTableEvent(name, key){
    var tbl = firstTable(name);
    if (!tbl) return;
    var fn = tableValue(tbl, key);
    if (fn && typeof window[fn] === 'function') window[fn]();
  }
  // ---- persistent edit mode ------------------------------------------------
  // Persistent tables store their header and rows in localStorage, keyed by
  // table name and page. Editable (session) tables keep the in-memory store;
  // they return to the source data on reload.
  var tableBootAt = Date.now();
  var tableUserTouched = false;
  function markTableUserTouched(){ tableUserTouched = true; }
  document.addEventListener('pointerdown', markTableUserTouched, true);
  document.addEventListener('keydown', markTableUserTouched, true);
  function tablePersistKey(name){
    var path = (window.location && window.location.pathname) || '';
    return 'sui.table.persist.' + path + '.' + name;
  }
  function captureColumnWidths(tbl){
    var out = [];
    if (!tbl) return out;
    var row = tbl.querySelector('thead tr') || (tbl.tBodies && tbl.tBodies[0] ? tbl.tBodies[0].rows[0] : null);
    if (!row) return out;
    for (var i = 0; i < row.cells.length; i++) {
      var w = parseFloat(row.cells[i].style.width || row.cells[i].style.minWidth || '');
      out.push(isFinite(w) ? Math.round(w) : 0);
    }
    return out;
  }
  function captureRowHeights(tbl){
    var out = [];
    if (!tbl) return out;
    var rows = bodyRows(tbl);
    for (var i = 0; i < rows.length; i++) {
      var h = parseFloat(rows[i].style.height || '');
      out.push(isFinite(h) ? Math.round(h) : 0);
    }
    return out;
  }
  function persistTableState(name, st, cfg){
    if (!cfg || cfg.editMode !== 'persistent') return;
    if (!window.localStorage) return;
    var tbl = firstTable(name);
    try {
      window.localStorage.setItem(tablePersistKey(name), JSON.stringify({
        version: 2,
        header: st && st.header ? st.header : [],
        rows: st && st.allRows ? st.allRows : [],
        textHeader: !!(st && st.textHeader),
        columnWidths: captureColumnWidths(tbl),
        rowHeights: captureRowHeights(tbl),
        cellText: tbl ? (tbl.getAttribute('data-sui-cell-text') || '') : '',
        sortKeys: st ? sortKeysOf(st) : [],
        filter: st ? (st.filter || '') : '',
        colFilters: st ? (st.colFilters || {}) : {},
        colValueFilters: st ? (st.colValueFilters || {}) : {},
        page: st ? (st.page || 0) : 0
      }));
    } catch (e) {}
  }
  function persistTable(name){
    var tbl = firstTable(name);
    if (!tbl) return;
    persistTableState(name, store[name], cfgOf(tbl));
  }
  function loadPersistedTable(name){
    try {
      var raw = window.localStorage && window.localStorage.getItem(tablePersistKey(name));
      if (!raw) return null;
      var p = JSON.parse(raw);
      if (!p || !Array.isArray(p.rows)) return null;
      return p;
    } catch (e) { return null; }
  }
  function applyPersistedTable(st, p){
    st.header = Array.isArray(p.header) ? p.header.map(function(v){ return String(v); }) : [];
    st.allRows = p.rows.map(function(r){ return Array.isArray(r) ? r.slice() : []; });
    st.rows = st.allRows;
    st.textHeader = !!p.textHeader;
    st.__fromDom = false;
    st.__colStruct = false;
    st.persistWidths = Array.isArray(p.columnWidths) ? p.columnWidths.slice() : [];
    st.persistRowHeights = Array.isArray(p.rowHeights) ? p.rowHeights.slice() : [];
    st.persistCellText = p.cellText || '';
    setSortKeys(st, Array.isArray(p.sortKeys) ? p.sortKeys : []);
    st.filter = p.filter || '';
    st.colFilters = p.colFilters && typeof p.colFilters === 'object' ? p.colFilters : {};
    st.colValueFilters = p.colValueFilters && typeof p.colValueFilters === 'object' ? p.colValueFilters : {};
    st.page = p.page || 0;
    clearSelection(st);
  }
  function applyPersistedTableAttrs(tbl, st){
    if (!tbl || !st) return;
    if (st.persistWidths && st.persistWidths.length) {
      tbl.setAttribute('data-sui-column-widths', st.persistWidths.join(','));
    }
    if (st.persistCellText === 'truncate' || st.persistCellText === 'wrap') {
      tbl.setAttribute('data-sui-cell-text', st.persistCellText);
      if (st.persistCellText === 'truncate') tbl.setAttribute('data-sui-fixed', '1');
    }
  }
  function captureTableSource(tbl, st){
    if (!st || st.__sourceRows) return;
    st.__sourceHeader = (st.header || []).slice();
    st.__sourceRows = (st.allRows || []).map(function(r){ return r.slice(); });
    st.__sourceWidthsAttr = tbl ? (tbl.getAttribute('data-sui-column-widths') || '') : '';
    st.__sourceCellText = tbl ? (tbl.getAttribute('data-sui-cell-text') || '') : '';
    if (!st.__sourceWidths) st.__sourceWidths = captureColumnWidths(tbl);
  }
  function restoreTableSourceAttrs(tbl, st){
    if (!tbl || !st) return;
    if (st.__sourceWidthsAttr) tbl.setAttribute('data-sui-column-widths', st.__sourceWidthsAttr);
    else tbl.removeAttribute('data-sui-column-widths');
    if (st.__sourceCellText) {
      tbl.setAttribute('data-sui-cell-text', st.__sourceCellText);
      if (st.__sourceCellText === 'truncate') tbl.setAttribute('data-sui-fixed', '1');
      else tbl.removeAttribute('data-sui-fixed');
    } else {
      tbl.removeAttribute('data-sui-cell-text');
      tbl.removeAttribute('data-sui-fixed');
    }
  }
  // During the initial boot a script-driven table.display/importText must not
  // overwrite saved persistent data: the user's saved rows win on reload.
  function persistentBootBlocks(name){
    var tbl = firstTable(name);
    if (!tbl) return false;
    var st = store[name];
    if (!st || !st.__persistLoaded) return false;
    if (cfgOf(tbl).editMode !== 'persistent') return false;
    return !tableUserTouched && (Date.now() - tableBootAt) < 2000;
  }
  function recordChange(name, rec){
    rec.time = Date.now();
    changes[name] = rec;
    if (rec && /^(cell|display|import|clear|row-|column-|sort|filter|page|resize)/.test(String(rec.op || ''))) persistTable(name);
  }
  function clearPersisted(name){
    if (!name) return [0, 'table.clearPersisted: name is required'];
    var removed = 0;
    try {
      if (window.localStorage && window.localStorage.getItem(tablePersistKey(name)) !== null) {
        window.localStorage.removeItem(tablePersistKey(name));
        removed = 1;
      }
    } catch (e) {}
    var st = store[name];
    var tbl = firstTable(name);
    if (st) {
      if (st.__sourceRows) {
        st.header = (st.__sourceHeader || []).slice();
        st.allRows = (st.__sourceRows || []).map(function(r){ return r.slice(); });
        st.rows = st.allRows;
        st.textHeader = false;
        st.persistWidths = (st.__sourceWidths || []).slice();
        st.persistRowHeights = [];
        st.persistCellText = st.__sourceCellText || '';
        st.filter = '';
        st.colFilters = {};
        st.colValueFilters = {};
        st.colFilter = null;
        setSortKeys(st, []);
        st.page = 0;
        clearSelection(st);
        restoreTableSourceAttrs(tbl, st);
        renderName(name);
      }
      delete st.__persistLoaded;
      delete st.persistWidths;
      delete st.persistRowHeights;
      delete st.persistCellText;
    }
    return [removed, null];
  }
  // Workbooks in persistent mode also keep the runtime-created sheet list.
  function tableBookPersistKey(box){
    var path = (window.location && window.location.pathname) || '';
    return 'sui.table.book.' + path + '.' + tableBookKey(box);
  }
  function workbookHasPersistentSheet(box){
    var panels = box ? box.querySelectorAll('.sui-tabpanels .sui-tabpanel') : [];
    for (var i = 0; i < panels.length; i++) {
      var t = panels[i].querySelector('table');
      if (t && cfgOf(t).editMode === 'persistent') return true;
    }
    return false;
  }
  function workbookPersistentMode(box){
    return workbookHasPersistentSheet(box) ? 'persistent' : '';
  }
  function saveTableBookDefs(box){
    if (!box || !window.localStorage) return;
    try {
      var key = tableBookPersistKey(box);
      if (workbookHasPersistentSheet(box)) {
        window.localStorage.setItem(key, JSON.stringify(tableBooks[tableBookKey(box)] || []));
      } else if (window.localStorage.getItem(key)) {
        window.localStorage.removeItem(key);
      }
    } catch (e) {}
  }
  function loadTableBookDefs(box){
    if (!box || !workbookHasPersistentSheet(box)) return [];
    try {
      var raw = window.localStorage && window.localStorage.getItem(tableBookPersistKey(box));
      var defs = raw ? JSON.parse(raw) : [];
      return Array.isArray(defs) ? defs : [];
    } catch (e) { return []; }
  }
  function tableStore(name){
    if (!store[name]) store[name] = { header: [], allRows: [], rows: [], page: 0,
      sortCol: null, sortDir: '', sortKeys: [], filter: '', colFilter: null,
      colFilters: {}, colValueFilters: {}, headerEdits: {}, filtersOn: null, sortOn: null,
      selectedRows: {}, selectedCells: {}, selectedCols: {} };
    return store[name];
  }
  function sortKeysOf(st){
    if (st.sortKeys && st.sortKeys.length) return st.sortKeys;
    if (st.sortCol !== null && st.sortCol !== undefined && st.sortDir) return [{ col: st.sortCol, dir: st.sortDir }];
    return [];
  }
  function setSortKeys(st, keys){
    st.sortKeys = keys || [];
    if (st.sortKeys.length) { st.sortCol = st.sortKeys[0].col; st.sortDir = st.sortKeys[0].dir; }
    else { st.sortCol = null; st.sortDir = ''; }
  }
  function colFiltersOf(st){ return st.colFilters || (st.colFilters = {}); }
  function colIndex(cfg, key){
    var s = String(key === undefined || key === null ? '' : key).trim();
    if (s === '') return -1;
    if (/^\d+$/.test(s)) return parseInt(s, 10);
    for (var i = 0; i < cfg.columns.length; i++) if (cfg.columns[i] === s) return i;
    for (var j = 0; j < cfg.labels.length; j++) if (cfg.labels[j] === s) return j;
    return -1;
  }
  function colOpts(cfg, col){
    if (!cfg.options) return null;
    return cfg.options[col] || cfg.options[String(col)] || null;
  }
  function validValue(v, type, opts){
    var s = String(v === undefined || v === null ? '' : v).trim();
    if (s === '') return '';
    type = String(type || 'text').toLowerCase();
    if (type === 'number') return isFinite(Number(s.replace(',', '.'))) ? '' : 'number';
    if (type === 'date') return isNaN(Date.parse(s)) ? 'date' : '';
    if (type === 'boolean') {
      var low = s.toLowerCase();
      return (low === 'true' || low === 'false' || low === 'yes' || low === 'no' || low === '1' || low === '0') ? '' : 'boolean';
    }
    if (type === 'select' && opts && opts.length) return opts.indexOf(s) >= 0 ? '' : 'select';
    return '';
  }
  function formatValue(v, fmt){
    var s = String(v === undefined || v === null ? '' : v);
    if (!fmt || fmt === 'text' || s === '') return s;
    var n = Number(s.replace(',', '.'));
    if (fmt === 'number') return isFinite(n) ? n.toLocaleString() : s;
    if (fmt === 'percent') return isFinite(n) ? n.toLocaleString(undefined, { style: 'percent', maximumFractionDigits: 2 }) : s;
    if (fmt.indexOf('currency') === 0) {
      var cur = fmt.split(':')[1] || 'USD';
      return isFinite(n) ? n.toLocaleString(undefined, { style: 'currency', currency: cur }) : s;
    }
    if (fmt === 'date') {
      var d = new Date(s);
      return isNaN(d.getTime()) ? s : d.toLocaleDateString();
    }
    return s;
  }
  function clearSelection(st){
    st.selectedRows = {}; st.selectedCells = {}; st.selectedCols = {};
  }
  function rowText(row){ return (row || []).join(' ').toLowerCase(); }
  function applyView(name){
    var st = tableStore(name);
    var tbl = firstTable(name);
    var cfg = tbl ? cfgOf(tbl) : { pageSize: 0, columns: [], labels: [] };
    st.filtersOn = (st.filtersOn === null || st.filtersOn === undefined) ? !!cfg.filterable : !!st.filtersOn;
    st.sortOn = (st.sortOn === null || st.sortOn === undefined) ? !!cfg.sortable : !!st.sortOn;
    var rows = st.allRows.slice();
    if (cfg.columns.length && !st.header.length) st.header = cfg.labels.length ? cfg.labels.slice() : cfg.columns.slice();
    if (st.filter) {
      var term = String(st.filter).toLowerCase();
      rows = rows.filter(function(r){ return rowText(r).indexOf(term) >= 0; });
    }
    if (st.colFilter) {
      var ci = colIndex(cfg, st.colFilter.column);
      var cv = String(st.colFilter.value || '').toLowerCase();
      if (ci >= 0 && cv) rows = rows.filter(function(r){ return String(r[ci] === undefined ? '' : r[ci]).toLowerCase().indexOf(cv) >= 0; });
    }
    // Per-column filters (the Excel-like header filter row).
    var cf = colFiltersOf(st);
    for (var cfk in cf) {
      if (!Object.prototype.hasOwnProperty.call(cf, cfk)) continue;
      var cfIdx = parseInt(cfk, 10);
      var cfv = String(cf[cfk] === undefined || cf[cfk] === null ? '' : cf[cfk]).toLowerCase();
      if (cfIdx >= 0 && cfv) {
        rows = rows.filter(function(r){ return String(r[cfIdx] === undefined ? '' : r[cfIdx]).toLowerCase().indexOf(cfv) >= 0; });
      }
    }
    // Excel-like value filters from the header filter popup: exact membership.
    var cvf = st.colValueFilters || {};
    for (var cvk in cvf) {
      if (!Object.prototype.hasOwnProperty.call(cvf, cvk)) continue;
      var cvIdx = parseInt(cvk, 10);
      var selected = cvf[cvk] || [];
      if (cvIdx >= 0) {
        rows = rows.filter(function(r){
          return selected.indexOf(String(r[cvIdx] === undefined ? '' : r[cvIdx])) >= 0;
        });
      }
    }
    var keys = sortKeysOf(st);
    if (keys.length) {
      rows.sort(function(a, b){
        for (var k = 0; k < keys.length; k++) {
          var key = keys[k];
          if (!key || !key.dir) continue;
          var sd = key.dir === 'desc' ? -1 : 1;
          var av = a[key.col] === undefined ? '' : a[key.col], bv = b[key.col] === undefined ? '' : b[key.col];
          var an = Number(av), bn = Number(bv);
          var cmp = 0;
          if (isFinite(an) && isFinite(bn) && String(av).trim() !== '' && String(bv).trim() !== '') cmp = an - bn;
          else {
            var as = String(av).toLowerCase(), bs = String(bv).toLowerCase();
            cmp = as < bs ? -1 : as > bs ? 1 : 0;
          }
          if (cmp) return cmp * sd;
        }
        return 0;
      });
    }
    var pageSize = (st.pageSize !== undefined && st.pageSize !== null) ? st.pageSize : cfg.pageSize;
    pageSize = pageSize > 0 ? pageSize : 0;
    var pageCount = pageSize > 0 ? Math.max(1, Math.ceil(rows.length / pageSize)) : 1;
    if (st.page >= pageCount) st.page = pageCount - 1;
    if (st.page < 0) st.page = 0;
    st.pageCount = pageCount;
    st.rows = pageSize > 0 ? rows.slice(st.page * pageSize, st.page * pageSize + pageSize) : rows;
    return { st: st, cfg: cfg };
  }

  function isObj(v){ return v !== null && typeof v === 'object' && !Array.isArray(v); }
  function cellStr(v){
    if (v === null || v === undefined) return '';
    if (typeof v === 'object') { try { return JSON.stringify(v); } catch(e){ return String(v); } }
    return String(v);
  }
  // normValue normalizes already-parsed data: arrays of arrays, arrays of
  // objects, {columns, rows} payloads, key/value objects and scalars.
  function normValue(data){
    if (data === null || data === undefined) return { header: null, rows: [] };
    if (Array.isArray(data)) {
      if (!data.length) return { header: null, rows: [] };
      if (data.length > MAX_ROWS) return { error: 'table.display: too many rows (limit ' + MAX_ROWS + ')' };
      var rows = [], header = null, i, j;
      if (Array.isArray(data[0])) {
        for (i = 0; i < data.length; i++) {
          var src = Array.isArray(data[i]) ? data[i] : [data[i]], rr = [];
          if (src.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
          for (j = 0; j < src.length; j++) rr.push(cellStr(src[j]));
          rows.push(rr);
        }
      } else if (isObj(data[0])) {
        header = [];
        for (i = 0; i < data.length; i++) {
          if (!isObj(data[i])) continue;
          for (var k in data[i]) {
            if (Object.prototype.hasOwnProperty.call(data[i], k) && header.indexOf(k) < 0) header.push(k);
          }
          if (header.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
        }
        for (i = 0; i < data.length; i++) {
          var o = isObj(data[i]) ? data[i] : {}, r2 = [];
          for (j = 0; j < header.length; j++) r2.push(cellStr(o[header[j]]));
          rows.push(r2);
        }
      } else {
        for (i = 0; i < data.length; i++) rows.push([cellStr(data[i])]);
      }
      return { header: header, rows: rows };
    }
    if (isObj(data)) {
      if (Array.isArray(data.rows)) {
        var sub = normValue(data.rows);
        if (sub.error) return sub;
        var h = null;
        if (Array.isArray(data.columns)) h = data.columns.map(cellStr);
        else if (Array.isArray(data.headers)) h = data.headers.map(cellStr);
        else h = sub.header;
        if (h && h.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
        return { header: h, rows: sub.rows };
      }
      var ks = Object.keys(data), kv = [];
      if (ks.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
      for (var x = 0; x < ks.length; x++) kv.push([ks[x], cellStr(data[ks[x]])]);
      return { header: ['key', 'value'], rows: kv };
    }
    return { header: null, rows: [[cellStr(data)]] };
  }

  // ---- delimited text (CSV, TSV, semicolon/pipe separated) -----------------
  function countSep(line, sep){
    var n = 0, q = false, i, ch;
    for (i = 0; i < line.length; i++) {
      ch = line.charAt(i);
      if (ch === '"') q = !q;
      else if (!q && ch === sep) n++;
    }
    return n;
  }
  function detectSep(text){
    var line = String(text).split(/\r?\n/)[0] || '';
    var cands = [',', ';', '\t', '|'], best = '', bestN = 0, i, n;
    for (i = 0; i < cands.length; i++) {
      n = countSep(line, cands[i]);
      if (n > bestN) { bestN = n; best = cands[i]; }
    }
    return best || ',';
  }
  function delimFor(fmt){
    if (fmt === 'tsv') return '\t';
    if (fmt === 'ssv') return ';';
    if (fmt === 'psv') return '|';
    return ',';
  }
  // RFC4180-style parser: quoted fields, "" escapes, embedded delimiters and
  // newlines inside quotes, CRLF, UTF-8 BOM.
  function parseDelimited(text, sep, quote, trim){
    var rows = [], row = [], field = '', inQ = false, i, ch;
    text = String(text).replace(/^\uFEFF/, '');
    for (i = 0; i < text.length; i++) {
      ch = text.charAt(i);
      if (inQ) {
        if (ch === quote) {
          if (text.charAt(i + 1) === quote) { field += quote; i++; }
          else inQ = false;
        } else field += ch;
      } else if (quote && ch === quote && field === '') {
        inQ = true;
      } else if (ch === sep) {
        row.push(trim ? field.trim() : field); field = '';
        if (row.length > MAX_COLS) return null;
      } else if (ch === '\n') {
        row.push(trim ? field.trim() : field); rows.push(row); row = []; field = '';
        if (rows.length > MAX_ROWS) return null;
      } else if (ch === '\r') {
        // CRLF: the following \n ends the row
      } else {
        field += ch;
      }
    }
    if (field !== '' || row.length) {
      row.push(trim ? field.trim() : field);
      if (row.length > MAX_COLS) return null;
      rows.push(row);
      if (rows.length > MAX_ROWS) return null;
    }
    if (rows.length && rows[rows.length - 1].length === 1 && rows[rows.length - 1][0] === '') rows.pop();
    return rows;
  }
  function normDelimited(text, header, sep, quote, trim){
    var s = sep || detectSep(text);
    var rows = parseDelimited(text, s, quote === undefined ? '"' : quote, !!trim);
    if (!rows) return { error: 'table.importText: too many rows or columns (limit ' + MAX_ROWS + 'x' + MAX_COLS + ')' };
    if (!rows.length) return { header: null, rows: [] };
    var useHeader = header === null || header === undefined ? rows.length > 1 : !!header;
    var h = useHeader ? rows.shift() : null;
    return { header: h, rows: rows, textHeader: useHeader, sep: s };
  }
  function tryJsonLines(text){
    var lines = String(text).split(/\r?\n/), vals = [], i, line;
    for (i = 0; i < lines.length; i++) {
      line = lines[i].trim();
      if (!line) continue;
      try { vals.push(JSON.parse(line)); } catch(e) { return null; }
      if (vals.length > MAX_ROWS) return { error: 'table.display: too many rows (limit ' + MAX_ROWS + ')' };
    }
    if (vals.length < 2) return null;
    return normValue(vals);
  }
  // startsJSON checks the first character without embedding raw delimiters in
  // the runtime (keeps the generated source's bracket count balanced).
  function startsJSON(s){
    var c = String(s).charCodeAt(0);
    return c === 91 || c === 123; // opening bracket or brace
  }
  // spec syntax: "csv", "tsv", "ssv" (;), "psv" (|), "json", "jsonl", "auto",
  // plus ;header=0|1 ;sep=comma|semicolon|tab|pipe|<char> ;quote=<char> ;trim=1
  function parseSpec(spec, defFmt){
    var s = String(spec === undefined || spec === null ? '' : spec).trim().toLowerCase();
    var bits = s ? s.split(';') : [];
    var fmt = bits.length ? bits.shift().trim() : (defFmt || 'auto');
    var o = { format: fmt || 'auto', header: null, sep: '', quote: '"', trim: false,
      formula: (typeof __SUI_SEC !== 'undefined' && __SUI_SEC.csvFormulaEscape) ? 'escape' : 'raw' };
    for (var i = 0; i < bits.length; i++) {
      var b = bits[i].trim();
      if (!b) continue;
      var eq = b.indexOf('=');
      var k = (eq < 0 ? b : b.slice(0, eq)).trim();
      var v = eq < 0 ? '' : b.slice(eq + 1).trim();
      if (k === 'header') o.header = !(v === '0' || v === 'false' || v === 'no' || v === 'off');
      else if (k === 'sep') {
        if (v === 'tab' || v === '\\t') o.sep = '\t';
        else if (v === 'comma') o.sep = ',';
        else if (v === 'semicolon') o.sep = ';';
        else if (v === 'pipe') o.sep = '|';
        else o.sep = v;
      }
      else if (k === 'quote') o.quote = v;
      else if (k === 'trim') o.trim = (v === '1' || v === 'true' || v === 'yes' || v === 'on');
      else if (k === 'formula' || k === 'safe') {
        o.formula = (v === 'escape' || v === 'safe' || v === '1' || v === 'true' || v === 'yes') ? 'escape' : 'raw';
      }
    }
    return o;
  }
  // normText parses an explicit text payload according to spec.
  function normText(text, spec){
    var o = parseSpec(spec, 'auto');
    var s = String(text === undefined || text === null ? '' : text).trim();
    if (!s) return { header: null, rows: [] };
    if (s.length > MAX_TEXT) return { error: 'table.importText: text too large (limit ' + MAX_TEXT + ' characters)' };
    if (o.format === 'json') {
      try { return normValue(JSON.parse(s)); }
      catch(e) { return { error: 'table.importText: invalid JSON (' + e.message + ')' }; }
    }
    if (o.format === 'jsonl' || o.format === 'ndjson') {
      var jl = tryJsonLines(s);
      return jl || { error: 'table.importText: invalid JSON Lines' };
    }
    if (o.format === 'auto') {
      if (startsJSON(s)) {
        var jerr = null;
        try { return normValue(JSON.parse(s)); } catch(e) { jerr = e; }
        var auto = tryJsonLines(s);
        if (auto) return auto;
        return { error: 'table.importText: invalid JSON (' + (jerr ? jerr.message : 'parse error') + ')' };
      }
    }
    var sep = o.sep || (o.format === 'auto' ? detectSep(s) : delimFor(o.format));
    var useHeader = o.header === null ? true : !!o.header;
    return normDelimited(s, useHeader, sep, o.quote, o.trim);
  }
  // norm is the table.display() entry point: it sniffs JSON, JSON Lines and
  // delimited text, and treats the first row of a multi-row delimited payload
  // as the header.
  function norm(data){
    if (typeof data !== 'string') return normValue(data);
    var s = data.trim();
    if (!s) return { header: null, rows: [] };
    if (s.length > MAX_TEXT) return { error: 'table.display: text too large (limit ' + MAX_TEXT + ' characters)' };
    if (startsJSON(s)) {
      var err = null;
      try { return normValue(JSON.parse(s)); } catch(e) { err = e; }
      var jl = tryJsonLines(s);
      if (jl) return jl;
      return { error: 'table.display: invalid JSON (' + (err ? err.message : 'parse error') + ')' };
    }
    return normDelimited(s, null, detectSep(s), '"', false);
  }

  function findTables(name){
    var out = [], ts = document.querySelectorAll('table');
    for (var i = 0; i < ts.length; i++) { if (ts[i].getAttribute('name') === name) out.push(ts[i]); }
    return out;
  }
  function cellsOf(tr){
    var out = [], ch = tr.children;
    for (var i = 0; i < ch.length; i++) { if (ch[i].tagName === 'TD' || ch[i].tagName === 'TH') out.push(ch[i]); }
    return out;
  }
  function tbodyOf(tbl){
    if (tbl.tBodies && tbl.tBodies.length) return tbl.tBodies[0];
    var tb = tbl.querySelector('tbody');
    if (!tb) { tb = document.createElement('tbody'); tbl.appendChild(tb); }
    return tb;
  }
  function bodyRows(tbl){
    var tb = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
    if (tb) return Array.prototype.slice.call(tb.rows);
    var out = [], trs = tbl.querySelectorAll('tr');
    for (var i = 0; i < trs.length; i++) { if (!trs[i].parentNode || trs[i].parentNode.tagName !== 'THEAD') out.push(trs[i]); }
    return out;
  }
  function templates(tbl){
    var t = tpls.get(tbl);
    if (t) return t;
    var bodies = [], heads = [], tb = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
    if (tb) for (var i = 0; i < tb.rows.length; i++) bodies.push(tb.rows[i].cloneNode(true));
    var hr = tbl.querySelector('thead tr');
    if (hr) heads.push(hr.cloneNode(true));
    if (!bodies.length) bodies.push(hr ? hr.cloneNode(true) : document.createElement('tr'));
    // Keep the designed ids so a runtime re-render (sort/filter/settings) does
    // not break the designer's preview picking.
    var rowIds = [], headCells = [];
    if (hr) {
      var hcs = cellsOf(hr);
      for (var hc = 0; hc < hcs.length; hc++) headCells.push(hcs[hc].getAttribute(TBL_ID_ATTR));
    }
    for (var b = 0; b < bodies.length; b++) {
      var bcs = cellsOf(bodies[b]), ids = [];
      for (var bi = 0; bi < bcs.length; bi++) ids.push(bcs[bi].getAttribute(TBL_ID_ATTR));
      rowIds.push({ row: bodies[b].getAttribute(TBL_ID_ATTR), cells: ids });
    }
    t = { heads: heads, bodies: bodies, headCells: headCells, rowIds: rowIds };
    tpls.set(tbl, t);
    return t;
  }
  function toTd(cell){
    if (cell.tagName !== 'TH') return cell;
    var td = document.createElement('td');
    if (cell.className) td.className = cell.className;
    var st = cell.getAttribute('style');
    if (st) td.setAttribute('style', st);
    return td;
  }
  function prepRow(tr, n, want){
    var cells = cellsOf(tr);
    while (cells.length > n) { tr.removeChild(cells.pop()); }
    while (cells.length < n) {
      var c = cells.length ? cells[cells.length - 1].cloneNode(false) : document.createElement(want === 'td' ? 'td' : 'th');
      tr.appendChild(c); cells.push(c);
    }
    for (var i = 0; i < cells.length; i++) {
      cells[i].removeAttribute('colspan');
      cells[i].removeAttribute('rowspan');
      cells[i].removeAttribute('data-' + 'sui-id');
      cells[i].removeAttribute('contenteditable');
      if (want === 'td') {
        var td = toTd(cells[i]);
        if (td !== cells[i]) tr.replaceChild(td, cells[i]);
        cells[i] = td;
      }
    }
    return cells;
  }
  function applyTo(tbl, ds){
    if (tableFilterPop && tableFilterPop.name === tableName(tbl)) closeTableFilterPop();
    var t = templates(tbl);
    var n = 0, i, j;
    var hr = tbl.querySelector('thead tr');
    if (!hr && ds.header && ds.header.length) {
      var thead = tbl.querySelector('thead');
      if (!thead) { thead = document.createElement('thead'); tbl.insertBefore(thead, tbl.firstChild); }
      hr = document.createElement('tr');
      thead.appendChild(hr);
    }
    var rows = ds.rows;
    // Delimited text always carries its first row as the header. When the
    // designed table has no <thead>, keep that row visible as the first body
    // row instead of dropping it.
    if (ds.textHeader && ds.header && !hr) rows = [ds.header].concat(rows);
    if (ds.header) n = ds.header.length;
    for (i = 0; i < rows.length; i++) n = Math.max(n, rows[i].length);
    if (!ds.header) {
      // No header information in the data: keep the designed column count and
      // labels (arrays/JSON arrays carry no column names).
      if (hr) n = Math.max(n, cellsOf(hr).length);
      if (t.bodies.length && t.bodies[0].children) n = Math.max(n, cellsOf(t.bodies[0]).length);
    }
    if (!n) {
      var tb0 = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
      if (tb0) while (tb0.firstChild) tb0.removeChild(tb0.firstChild);
      return;
    }
    if (hr) {
      var hc = prepRow(hr, n, 'th');
      if (ds.header) {
        for (j = 0; j < hc.length; j++) hc[j].textContent = j < ds.header.length ? ds.header[j] : '';
      } else if (t.heads.length) {
        // restore the labels as designed (a previous object payload may have
        // replaced them with key names)
        var orig = cellsOf(t.heads[0]);
        for (j = 0; j < hc.length; j++) hc[j].textContent = j < orig.length ? orig[j].textContent : '';
      }
      for (j = 0; j < hc.length; j++) {
        if (t.headCells && t.headCells[j]) hc[j].setAttribute(TBL_ID_ATTR, t.headCells[j]);
      }
    }
    var tb = tbodyOf(tbl);
    while (tb.firstChild) tb.removeChild(tb.firstChild);
    for (i = 0; i < rows.length; i++) {
      var origIdx = i;
      if (ds.allRows) {
        var found = ds.allRows.indexOf(rows[i]);
        if (found >= 0) origIdx = found;
      }
      var tr = t.bodies[i % t.bodies.length].cloneNode(true);
      tr.removeAttribute(TBL_ID_ATTR);
      var cs = prepRow(tr, n, 'td');
      for (j = 0; j < cs.length; j++) cs[j].textContent = rows[i][j] !== undefined ? rows[i][j] : '';
      var ids = t.rowIds && t.rowIds[origIdx];
      if (ids) {
        if (ids.row) tr.setAttribute(TBL_ID_ATTR, ids.row);
        for (j = 0; j < cs.length; j++) {
          if (ids.cells[j]) cs[j].setAttribute(TBL_ID_ATTR, ids.cells[j]);
        }
      }
      tb.appendChild(tr);
    }
  }
  function selectionChanged(name, rec){
    recordChange(name, rec);
    renderName(name);
    fireTableEvent(name, 'onrowselect');
  }
  function selectRow(name, row, multi){
    var st = tableStore(name);
    if (!multi) clearSelection(st);
    if (st.selectedRows[row]) delete st.selectedRows[row]; else st.selectedRows[row] = true;
    selectionChanged(name, { op: 'select-row', row: row });
  }
  function selectCell(name, row, col, multi){
    var st = tableStore(name);
    if (!multi) clearSelection(st);
    var key = row + ':' + col;
    if (st.selectedCells[key]) delete st.selectedCells[key]; else st.selectedCells[key] = true;
    selectionChanged(name, { op: 'select-cell', row: row, column: col });
  }
  function selectColumn(name, col, multi){
    var st = tableStore(name);
    if (!multi) clearSelection(st);
    if (st.selectedCols[col]) delete st.selectedCols[col]; else st.selectedCols[col] = true;
    selectionChanged(name, { op: 'select-column', column: col });
  }
  function structRowList(st, row){
    var keys = [];
    for (var k in st.selectedRows) if (st.selectedRows[k]) keys.push(parseInt(k, 10));
    keys.sort(function(a,b){ return a-b; });
    if (!keys.length && row >= 0) keys = [row];
    return keys;
  }
  function structColList(st, col){
    var keys = [];
    for (var k in st.selectedCols) if (st.selectedCols[k]) keys.push(parseInt(k, 10));
    keys.sort(function(a,b){ return a-b; });
    if (!keys.length && col >= 0) keys = [col];
    return keys;
  }
  function structApply(name, st, rec){
    st.page = 0;
    clearSelection(st);
    recordChange(name, rec);
    renderName(name);
    fireTableEvent(name, 'ontablechange');
  }
  function structInsertRows(name, at, rows){
    var st = tableStore(name);
    var w = st.header.length;
    for (var i = 0; i < rows.length; i++) {
      var r = rows[i].slice();
      while (r.length < w) r.push('');
      st.allRows.splice(at + i, 0, r);
    }
    structApply(name, st, { op: 'row-insert', row: at, count: rows.length });
  }
  function structDeleteRows(name, indexes){
    var st = tableStore(name);
    indexes = indexes.slice().sort(function(a,b){ return b-a; });
    for (var i = 0; i < indexes.length; i++) st.allRows.splice(indexes[i], 1);
    structApply(name, st, { op: 'row-delete', row: indexes.length ? indexes[indexes.length-1] : 0, count: indexes.length });
  }
  function structInsertColumns(name, at, labels, values){
    var st = tableStore(name);
    for (var i = 0; i < labels.length; i++) st.header.splice(at + i, 0, labels[i]);
    for (var r = 0; r < st.allRows.length; r++) {
      for (var c = 0; c < labels.length; c++) st.allRows[r].splice(at + c, 0, values[r] && values[r][c] !== undefined ? values[r][c] : '');
    }
    st.__colStruct = true;
    structApply(name, st, { op: 'column-insert', column: at, count: labels.length });
  }
  function structDeleteColumns(name, indexes){
    var st = tableStore(name);
    indexes = indexes.slice().sort(function(a,b){ return b-a; });
    for (var i = 0; i < indexes.length; i++) {
      var col = indexes[i];
      if (st.header.length) st.header.splice(col, 1);
      for (var r = 0; r < st.allRows.length; r++) st.allRows[r].splice(col, 1);
    }
    st.__colStruct = true;
    structApply(name, st, { op: 'column-delete', column: indexes.length ? indexes[indexes.length-1] : 0, count: indexes.length });
  }
  function openTableStructMenu(ev, tbl, name, row, col, cfg){
    var old = document.getElementById('sui-table-struct');
    if (old) old.remove();
    var st = tableStore(name);
    var menu = document.createElement('div');
    menu.id = 'sui-table-struct';
    menu.className = 'sui-table-struct-menu';
    menu.setAttribute('role', 'menu');
    function add(label, fn, disabled) {
      var b = document.createElement('button');
      b.type = 'button'; b.className = 'sui-table-struct-item'; b.textContent = label;
      b.disabled = !!disabled;
      b.addEventListener('click', function(e){ e.stopPropagation(); menu.remove(); fn(); });
      menu.appendChild(b);
    }
    function sep(){ var d = document.createElement('div'); d.className = 'sui-table-struct-sep'; menu.appendChild(d); }
    var rows = structRowList(st, row), cols = structColList(st, col);
    if (row >= 0) {
      add('Insert row above', function(){ structInsertRows(name, Math.min.apply(null, rows), [[]]); });
      add('Insert row below', function(){ structInsertRows(name, Math.max.apply(null, rows) + 1, [[]]); });
      add('Delete row' + (rows.length > 1 ? 's' : ''), function(){ structDeleteRows(name, rows); });
      add('Copy row' + (rows.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'rows', rows: rows.map(function(i){ return (st.allRows[i] || []).slice(); }) };
      });
      add('Cut row' + (rows.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'rows', rows: rows.map(function(i){ return (st.allRows[i] || []).slice(); }) };
        structDeleteRows(name, rows);
      });
      if (tableClip[name] && tableClip[name].kind === 'rows') add('Paste rows', function(){ structInsertRows(name, row, tableClip[name].rows); });
      sep();
    }
    if (col >= 0) {
      add('Insert column left', function(){ structInsertColumns(name, Math.min.apply(null, cols), ['Column'], []); });
      add('Insert column right', function(){ structInsertColumns(name, Math.max.apply(null, cols) + 1, ['Column'], []); });
      add('Delete column' + (cols.length > 1 ? 's' : ''), function(){ structDeleteColumns(name, cols); });
      add('Copy column' + (cols.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'cols', labels: cols.map(function(c){ return st.header[c] || ''; }),
          values: st.allRows.map(function(r){ return cols.map(function(c){ return r[c] === undefined ? '' : r[c]; }); }) };
      });
      add('Cut column' + (cols.length > 1 ? 's' : ''), function(){
        tableClip[name] = { kind: 'cols', labels: cols.map(function(c){ return st.header[c] || ''; }),
          values: st.allRows.map(function(r){ return cols.map(function(c){ return r[c] === undefined ? '' : r[c]; }); }) };
        structDeleteColumns(name, cols);
      });
      if (tableClip[name] && tableClip[name].kind === 'cols') add('Paste columns', function(){ structInsertColumns(name, col, tableClip[name].labels, tableClip[name].values); });
      sep();
    }
    add('Add row at end', function(){ structInsertRows(name, st.allRows.length, [[]]); });
    add('Add column at end', function(){ structInsertColumns(name, st.header.length || (st.allRows[0] ? st.allRows[0].length : 0), ['Column'], []); });
    var x = ev.clientX, y = ev.clientY;
    document.body.appendChild(menu);
    var rect = menu.getBoundingClientRect();
    if (x + rect.width > window.innerWidth - 8) x = Math.max(8, window.innerWidth - rect.width - 8);
    if (y + rect.height > window.innerHeight - 8) y = Math.max(8, window.innerHeight - rect.height - 8);
    menu.style.left = x + 'px'; menu.style.top = y + 'px';
    var close = function(){ menu.remove(); document.removeEventListener('click', close); };
    setTimeout(function(){ document.addEventListener('click', close); }, 0);
  }
  function tableBookKey(box){
    return box.getAttribute('data-sui-tabs-key') || ('tabs:' + (box.getAttribute('aria-label') || '') + ':' + Array.prototype.indexOf.call(document.querySelectorAll('.sui-tabs'), box));
  }
  function addRuntimeTableDom(box, name, label, templatePanel){
    var list = box.querySelector('.sui-tablist');
    var panels = box.querySelector('.sui-tabpanels');
    if (!list || !panels || !templatePanel) return null;
    if (list.querySelector('[data-sui-table-sheet="' + name + '"]')) return null;
    var clone = templatePanel.cloneNode(true);
    clone.classList.remove('active');
    clone.hidden = true;
    clone.setAttribute('data-sui-table-sheet', name);
    Array.prototype.forEach.call(clone.querySelectorAll('[' + 'data-' + 'sui-id' + ']'), function(el){ el.removeAttribute('data-' + 'sui-id'); });
    Array.prototype.forEach.call(clone.querySelectorAll('.sui-table-filter, .sui-table-pager, .sui-table-add, .sui-table-topbar, .sui-table-tools'), function(el){ if (el.parentNode) el.parentNode.removeChild(el); });
      var newTable = clone.querySelector('table');
    if (newTable) {
      newTable.setAttribute('name', name);
      // A new sheet gets its own empty title row, not the source caption.
      var clonedCaption = newTable.querySelector('caption');
      if (clonedCaption && clonedCaption.parentNode) clonedCaption.parentNode.removeChild(clonedCaption);
      var tb = newTable.tBodies && newTable.tBodies.length ? newTable.tBodies[0] : null;
      // Keep a real body row as the template so the new sheet starts with two
      // styled empty rows (not header clones, which could look unstyled/white).
      var bodyTemplate = tb && tb.rows.length ? tb.rows[0].cloneNode(true) : null;
      if (tb) while (tb.firstChild) tb.removeChild(tb.firstChild);
      if (tb) {
        for (var rIdx = 0; rIdx < 2; rIdx++) {
          var tr = bodyTemplate ? bodyTemplate.cloneNode(true) : document.createElement('tr');
          if (!bodyTemplate) {
            var colCount = newTable.querySelectorAll('thead th').length;
            for (var cIdx = 0; cIdx < colCount; cIdx++) tr.appendChild(document.createElement('td'));
          }
          Array.prototype.forEach.call(tr.querySelectorAll('[' + TBL_ID_ATTR + ']'), function(el){ el.removeAttribute(TBL_ID_ATTR); });
          tr.removeAttribute(TBL_ID_ATTR);
          Array.prototype.forEach.call(tr.querySelectorAll('td,th'), function(el){
            el.removeAttribute(TBL_ID_ATTR);
            el.removeAttribute('data-sui-raw');
            el.textContent = '';
          });
          tb.appendChild(tr);
        }
      }
      // A new sheet keeps its own generic header instead of inheriting the
      // source table's labels, types and formats: a blank page must accept any
      // text, even when it was cloned from a read-only or numeric table.
      var headCells = newTable.querySelectorAll('thead tr:first-child th, thead tr:first-child td');
      // On a re-render the template panel may not be decorated yet, so derive
      // the column count from the declared columns when there are no th cells.
      var colCount = headCells.length;
      if (!colCount) {
        var colAttr = newTable.getAttribute('data-sui-columns') || '';
        if (colAttr) colCount = colAttr.split(',').filter(function(x){ return x.trim() !== ''; }).length;
      }
      // A runtime sheet may have gained columns before a re-render; keep the
      // stored width so the header and the data stay aligned.
      var existingStore = store[name];
      if (existingStore && existingStore.header && existingStore.header.length > colCount) colCount = existingStore.header.length;
      if (colCount) {
        var headLabels = [], headKeys = [];
        for (var hi = 0; hi < colCount; hi++) {
          var headLabel = hi < 26 ? String.fromCharCode(65 + hi) : ('Col ' + (hi + 1));
          headLabels.push(headLabel);
          headKeys.push(String.fromCharCode(97 + (hi % 26)) + (hi >= 26 ? String(hi + 1) : ''));
          if (headCells[hi]) {
            headCells[hi].removeAttribute('data-sui-raw');
            headCells[hi].textContent = headLabel;
          }
        }
        newTable.setAttribute('data-sui-column-labels', headLabels.join(','));
        newTable.setAttribute('data-sui-columns', headKeys.join(','));
      }
      newTable.removeAttribute('data-sui-column-types');
      newTable.removeAttribute('data-sui-column-formats');
      newTable.removeAttribute('data-sui-column-options');
      // A runtime sheet is for manual data entry, so it is always editable.
      newTable.setAttribute('data-sui-editable', '1');
    }
    panels.appendChild(clone);
    var tab = document.createElement('button');
    tab.type = 'button'; tab.className = 'sui-tab';
    tab.setAttribute('role', 'tab');
    tab.setAttribute('aria-selected', 'false');
    tab.setAttribute('data-sui-table-sheet', name);
    tab.tabIndex = -1;
    tab.textContent = label;
    var addTab = list.querySelector('.sui-tab-add');
    if (addTab) list.insertBefore(tab, addTab); else list.appendChild(tab);
    return { tab: tab, panel: clone, table: newTable };
  }
  function restoreTableBooks(){
    var boxes = document.querySelectorAll('.sui-tabs');
    for (var b = 0; b < boxes.length; b++) {
      var box = boxes[b], key = tableBookKey(box), defs = tableBooks[key] || [];
      if (!defs.length) {
        var savedDefs = loadTableBookDefs(box);
        if (savedDefs.length) { tableBooks[key] = savedDefs; defs = savedDefs; }
      }
      if (!defs.length) continue;
      var panels = box.querySelector('.sui-tabpanels');
      var template = panels ? panels.querySelector('.sui-tabpanel') : null;
      if (!template) continue;
      for (var i = 0; i < defs.length; i++) {
        var made = addRuntimeTableDom(box, defs[i].name, defs[i].label, template);
        var bookMode = workbookPersistentMode(box);
        if (made && made.table && bookMode) made.table.setAttribute('data-sui-edit-mode', bookMode);
        if (made && store[defs[i].name]) {
          try { applyTo(made.table, store[defs[i].name]); } catch (e) {}
        }
      }
    }
  }
  function ensureAddTableButton(tbl, name, st, cfg){
    if (!cfg.showAddTable) return;
    var box = tbl.closest ? tbl.closest('.sui-tabs') : null;
    if (!box) return;
    var list = box.querySelector('.sui-tablist');
    if (!list) return;
    var btn = list.querySelector('.sui-tab-add');
    if (!btn) {
      btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'sui-tab-add';
      btn.setAttribute('data-sui-tab-add', '1');
      btn.textContent = '+';
      btn.title = 'Add table';
      btn.setAttribute('aria-label', 'Add table');
      btn.addEventListener('click', function(ev){
        ev.preventDefault();
        ev.stopPropagation();
        var panels = box.querySelector('.sui-tabpanels');
        var panel = panels ? (panels.querySelector('.sui-tabpanel.active') || panels.querySelector('.sui-tabpanel:not([hidden])')) : null;
        if (!panel) panel = tbl.closest('.sui-tabpanel');
        if (!panel) return;
        var activeTable = panel.querySelector('table');
        var activeCfg = activeTable ? cfgOf(activeTable) : cfg;
        var n = 1, newName = 'table-' + n;
        while (findTables(newName).length) { n++; newName = 'table-' + n; }
        var label = 'Table ' + (box.querySelectorAll('.sui-tabpanel').length + 1);
        var made = addRuntimeTableDom(box, newName, label, panel);
        if (!made) return;
        if (made.table) {
          // A new sheet inherits the workbook edit mode: persistent sheets
          // keep their data after a reload, editable ones stay per session.
          made.table.setAttribute('data-sui-edit-mode', activeCfg.editMode);
          made.table.setAttribute('data-sui-editable', '1');
          made.table.setAttribute('data-sui-show-add-table', '1');
        }
        if (activeCfg.structureEditable && made.table) {
          made.table.setAttribute('data-sui-editable', '1');
          made.table.setAttribute('data-sui-structure-editable', '1');
        }
        var key = tableBookKey(box);
        // A runtime sheet name may be recycled after a deletion.
        if (deletedWorkbookSheets[key]) delete deletedWorkbookSheets[key]['n:' + newName];
        (tableBooks[key] || (tableBooks[key] = [])).push({ name: newName, label: label });
        saveTableBookDefs(box);
        made.tab.click();
        init();
      });
      list.appendChild(btn);
    } else if (list.lastElementChild !== btn) {
      list.appendChild(btn);
    }
  }
  function applyHeaderSort(name, tbl, idx, cell, addLevel){
    var st = tableStore(name);
    var cfg = cfgOf(tbl);
    if (cfg.selection === 'column') {
      // The caller already handled plain column selection clicks.
    }
    if (!st.sortOn) return;
    var keys = sortKeysOf(st).map(function(k){ return { col: k.col, dir: k.dir }; });
    var pos = -1;
    for (var q = 0; q < keys.length; q++) if (keys[q].col === idx) { pos = q; break; }
    if (addLevel) {
      if (pos >= 0) {
        if (keys[pos].dir === 'asc') keys[pos].dir = 'desc';
        else keys.splice(pos, 1);
      } else keys.push({ col: idx, dir: 'asc' });
    } else if (keys.length === 1 && pos === 0) {
      keys = keys[0].dir === 'asc' ? [{ col: idx, dir: 'desc' }] : [];
    } else {
      keys = [{ col: idx, dir: 'asc' }];
    }
    setSortKeys(st, keys);
    st.page = 0;
    clearSelection(st);
    var lbl = cell.querySelector('.sui-table-hlabel');
    recordChange(name, { op: 'sort', column: idx, dir: keys.length ? keys[keys.length - 1].dir : '',
      keys: keys.map(function(k){ return { column: k.col, dir: k.dir }; }),
      label: lbl ? lbl.textContent : cell.textContent });
    renderName(name);
    fireTableEvent(name, 'onsort');
    fireTableEvent(name, 'ontablechange');
  }
  function decorateTable(tbl, st, cfg){
    var name = tableName(tbl);
    // Cell input is internal to the table. Stop it at the table capture phase
    // so application-level onInput handlers on the table or its ancestors do
    // not re-render the view (and blur the edited cell) on every letter.
    if (tbl && !tbl.__suiInputGuard) {
      tbl.__suiInputGuard = true;
      tbl.addEventListener('input', function(ev){ ev.stopPropagation(); }, true);
    }
    if (st.filtersOn === null || st.filtersOn === undefined) st.filtersOn = !!cfg.filterable;
    if (st.sortOn === null || st.sortOn === undefined) st.sortOn = !!cfg.sortable;
    var headerRow = tbl.querySelector('thead tr');
    var hrCells = headerRow ? cellsOf(headerRow) : [];
    for (var h = 0; h < hrCells.length; h++) {
      var hc = hrCells[h];
      hc.setAttribute('data-col', String(h));
      var hwrap = hc.querySelector('.sui-table-hwrap');
      if (!hwrap) {
        hwrap = document.createElement('div');
        hwrap.className = 'sui-table-hwrap';
        while (hc.firstChild) hwrap.appendChild(hc.firstChild);
        hc.appendChild(hwrap);
      }
      var hlabel = hwrap.querySelector('.sui-table-hlabel');
      if (!hlabel) {
        hlabel = document.createElement('span');
        hlabel.className = 'sui-table-hlabel';
        var moveNodes = [];
        for (var mi = 0; mi < hwrap.childNodes.length; mi++) {
          var mn = hwrap.childNodes[mi];
          if (mn.classList && mn.classList.contains('sui-table-hbtns')) continue;
          moveNodes.push(mn);
        }
        for (var mj = 0; mj < moveNodes.length; mj++) hlabel.appendChild(moveNodes[mj]);
        hwrap.insertBefore(hlabel, hwrap.firstChild);
      }
      var editedLabel = st.headerEdits && Object.prototype.hasOwnProperty.call(st.headerEdits, h);
      var targetLabel = editedLabel ? st.headerEdits[h] : (cfg.labels[h] !== undefined && cfg.labels[h] !== '' ? cfg.labels[h] : (st.header && st.header[h] !== undefined ? st.header[h] : ""));
      if (targetLabel === undefined || targetLabel === null) targetLabel = "";
      if (!st.__colStruct && hc !== editCell && hlabel.textContent !== targetLabel) hlabel.textContent = targetLabel;
      if (targetLabel && hlabel.title !== targetLabel) hlabel.title = targetLabel;
      else if (!targetLabel && hlabel.title) hlabel.removeAttribute('title');
      if (cfg.widths[h] > 0) { hc.style.width = cfg.widths[h] + 'px'; hc.style.minWidth = cfg.widths[h] + 'px'; }
      hlabel.style.textAlign = cfg.align[h] || '';
      hc.classList.toggle('sui-table-sortable', !!st.sortOn);
      if (st.sortOn) hc.setAttribute('tabindex', '0'); else hc.removeAttribute('tabindex');
      if (!hc.__suiSortBound) {
        hc.__suiSortBound = true;
        hc.addEventListener('click', function(ev){
          if (ev.target && ev.target.closest && ev.target.closest('.sui-table-hbtns')) return;
          var cell = ev.currentTarget;
          var idx = parseInt(cell.getAttribute('data-col'), 10) || 0;
          var st2 = tableStore(name);
          var cfg2 = cfgOf(tbl);
          // Editable headers are text fields: a plain click places the caret.
          // Sorting stays on the explicit header sort button so it cannot
          // re-render the table (and blur the header) while editing.
          if (cfg2.editable) return;
          if (cfg2.selection === 'column' && !ev.altKey) {
            selectColumn(name, idx, ev.shiftKey || ev.ctrlKey || ev.metaKey || cfg2.multi);
            return;
          }
          if (!st2.sortOn) return;
          applyHeaderSort(name, tbl, idx, cell, ev.ctrlKey || ev.metaKey || ev.shiftKey);
        });
      }
      var sortKeys = sortKeysOf(st);
      var sortDir = '', sortRank = -1;
      for (var sk = 0; sk < sortKeys.length; sk++) {
        if (sortKeys[sk].col === h) { sortDir = sortKeys[sk].dir; sortRank = sk; break; }
      }
      hc.setAttribute('aria-sort', sortDir === 'asc' ? 'ascending' : sortDir === 'desc' ? 'descending' : 'none');
      hc.classList.toggle('sui-table-sort-asc', sortDir === 'asc');
      hc.classList.toggle('sui-table-sort-desc', sortDir === 'desc');
      if (sortRank >= 0 && sortKeys.length > 1) hc.setAttribute('data-sort-rank', String(sortRank + 1));
      else hc.removeAttribute('data-sort-rank');
      ensureHeaderTools(hc, hwrap, hlabel, tbl, name, h, st, cfg, sortDir);
      if (cfg.structureEditable && !hc.__suiStructBound) {
        hc.__suiStructBound = true;
        hc.addEventListener('contextmenu', function(ev){
          ev.preventDefault();
          var col = parseInt(ev.currentTarget.getAttribute('data-col'), 10) || 0;
          openTableStructMenu(ev, tbl, name, -1, col, cfg);
        });
      }
    }
    // Keep cfgOf/header()/column pickers aligned with runtime header edits.
    var hasHeaderEdits = st.headerEdits && Object.keys(st.headerEdits).length > 0;
    if (hrCells.length && (tbl.hasAttribute('data-sui-column-labels') || hasHeaderEdits)) {
      var runtimeLabels = [];
      for (var rl = 0; rl < hrCells.length; rl++) {
        var rlLabel = hrCells[rl].querySelector('.sui-table-hlabel');
        runtimeLabels.push(rlLabel ? (rlLabel.textContent || '') : (hrCells[rl].textContent || '').trim());
      }
      var runtimeLabelsStr = runtimeLabels.join(',');
      if ((tbl.getAttribute('data-sui-column-labels') || '') !== runtimeLabelsStr) {
        tbl.setAttribute('data-sui-column-labels', runtimeLabelsStr);
      }
    }
    var rows = bodyRows(tbl);
    for (var i = 0; i < rows.length; i++) {
      var tr = rows[i];
      tr.setAttribute('data-row', String(i));
      var rowSelected = !!st.selectedRows[i];
      tr.classList.toggle('sui-table-row-selected', rowSelected);
      tr.setAttribute('aria-selected', rowSelected ? 'true' : 'false');
      var cs = cellsOf(tr);
      for (var j = 0; j < cs.length; j++) {
        var c = cs[j];
        c.setAttribute('data-col', String(j));
        if (cfg.widths[j] > 0) { c.style.width = cfg.widths[j] + 'px'; c.style.minWidth = cfg.widths[j] + 'px'; }
        if (cfg.align[j]) c.style.textAlign = cfg.align[j];
        var raw = st.rows[i] && st.rows[i][j] !== undefined ? st.rows[i][j] : (c.getAttribute('data-sui-raw') || c.textContent || '');
        if (st.rows[i]) {
          c.setAttribute('data-sui-raw', raw);
          if (c !== editCell) {
            var fv = formatValue(raw, cfg.formats[j]);
            if (c.textContent !== fv) c.textContent = fv;
          }
        } else if (cfg.formats[j]) {
          c.setAttribute('data-sui-raw', raw);
          if (c !== editCell) {
            var fv2 = formatValue(raw, cfg.formats[j]);
            if (c.textContent !== fv2) c.textContent = fv2;
          }
        }
        var cellKey = i + ':' + j;
        var cellSel = !!st.selectedCells[cellKey];
        c.classList.toggle('sui-table-cell-selected', cellSel);
        c.classList.toggle('sui-table-col-selected', !!st.selectedCols[j]);
        c.setAttribute('aria-selected', (rowSelected || cellSel || !!st.selectedCols[j]) ? 'true' : 'false');
        if (!c.__suiClickBound) {
          c.__suiClickBound = true;
          c.addEventListener('click', function(ev){
            var cell = ev.currentTarget;
            var row = parseInt(cell.parentNode.getAttribute('data-row'), 10) || 0;
            var col = parseInt(cell.getAttribute('data-col'), 10) || 0;
            recordChange(name, { op: 'click-cell', row: row, column: col });
            fireTableEvent(name, 'oncellclick');
            fireTableEvent(name, 'onrowclick');
          });
        }
        if (cfg.structureEditable && !c.__suiStructBound) {
          c.__suiStructBound = true;
          c.addEventListener('contextmenu', function(ev){
            ev.preventDefault();
            var cell = ev.currentTarget;
            var row = parseInt(cell.parentNode.getAttribute('data-row'), 10) || 0;
            var col = parseInt(cell.getAttribute('data-col'), 10) || 0;
            openTableStructMenu(ev, tbl, name, row, col, cfg);
          });
        }
        if (cfg.selection !== 'none' && !cfg.editable && !c.__suiSelectBound) {
          c.__suiSelectBound = true;
          c.addEventListener('click', function(ev){
            var cell = ev.currentTarget;
            var cfg2 = cfgOf(tbl);
            var row = parseInt(cell.parentNode.getAttribute('data-row'), 10) || 0;
            var col = parseInt(cell.getAttribute('data-col'), 10) || 0;
            if (cfg2.selection === 'row') selectRow(name, row, cfg2.multi && (ev.ctrlKey || ev.metaKey || ev.shiftKey));
            else if (cfg2.selection === 'cell') selectCell(name, row, col, cfg2.multi && (ev.ctrlKey || ev.metaKey || ev.shiftKey));
            else if (cfg2.selection === 'column') selectColumn(name, col, cfg2.multi && (ev.ctrlKey || ev.metaKey || ev.shiftKey));
          });
        }
      }
    }
    if (st.persistRowHeights && st.persistRowHeights.length) {
      var heightRows = bodyRows(tbl);
      for (var hr2 = 0; hr2 < heightRows.length && hr2 < st.persistRowHeights.length; hr2++) {
        var ph = st.persistRowHeights[hr2];
        if (!(ph > 0)) continue;
        heightRows[hr2].style.height = ph + 'px';
        for (var hc2 = 0; hc2 < heightRows[hr2].children.length; hc2++) heightRows[hr2].children[hc2].style.height = ph + 'px';
      }
    }
    updateWorkbookTableWidth(tbl);
    ensureTableTools(tbl, name);
  }
  function ensureFilterInput(tbl, name, st, cfg){
    if (!cfg.filterable) return;
    var host = tbl.previousElementSibling;
    var toolsHost = (host && host.classList && host.classList.contains('sui-table-tools')) ? host : null;
    var prev = toolsHost ? toolsHost.previousElementSibling : host;
    var inp = prev && prev.classList && prev.classList.contains('sui-table-filter') &&
      prev.getAttribute('data-sui-table-filter-for') === name ? prev : null;
    if (!inp) {
      inp = document.createElement('input');
      inp.type = 'text';
      inp.className = 'sui-table-filter';
      inp.setAttribute('data-sui-table-filter-for', name);
      inp.setAttribute('placeholder', 'Filter table\u2026');
      inp.setAttribute('aria-label', 'Filter table ' + (name || ''));
      inp.value = st.filter || '';
      inp.addEventListener('input', function(){
        var st2 = tableStore(name);
        st2.filter = inp.value;
        st2.page = 0;
        clearSelection(st2);
        renderName(name);
      });
      if (tbl.parentNode) tbl.parentNode.insertBefore(inp, toolsHost || tbl);
    } else if (document.activeElement !== inp && inp.value !== (st.filter || '')) {
      inp.value = st.filter || '';
    }
  }
  // ---- per-column filters and the runtime settings panel ------------------
  var tableCfg = null;
  function cfgEl(tag, cls, text){
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined) e.textContent = text;
    return e;
  }
  function columnChoices(tbl, cfg){
    var out = [], hr = tbl.querySelector('thead tr'), cells = hr ? cellsOf(hr) : [];
    var n = Math.max(cells.length, cfg.columns.length, cfg.labels.length);
    for (var i = 0; i < n; i++) {
      var lbl = '';
      if (cells[i]) {
        var span = cells[i].querySelector('.sui-table-hlabel');
        lbl = span ? span.textContent : (cells[i].getAttribute('data-sui-raw') || cells[i].textContent);
      }
      if (!lbl) lbl = cfg.labels[i] || cfg.columns[i] || ('Column ' + (i + 1));
      out.push({ index: i, label: String(lbl).trim() || ('Column ' + (i + 1)) });
    }
    return out;
  }
  function hasValueFilter(st, col){
    return Object.prototype.hasOwnProperty.call(st.colValueFilters || {}, String(col));
  }
  function columnValuePairs(tbl, name, col, cfg){
    var st = tableStore(name), seen = {}, out = [];
    function add(v){
      var raw = String(v === undefined || v === null ? '' : v);
      if (Object.prototype.hasOwnProperty.call(seen, raw)) return;
      seen[raw] = 1;
      var label = raw === '' ? '(Blanks)' : formatValue(raw, cfg.formats[col]);
      out.push({ raw: raw, label: label === '' ? '(Blanks)' : label });
    }
    var rows = st.allRows && st.allRows.length ? st.allRows : null;
    if (rows) {
      for (var i = 0; i < rows.length; i++) add(rows[i][col]);
    } else {
      var tb = tbl.tBodies && tbl.tBodies[0];
      if (tb) {
        for (var r = 0; r < tb.rows.length; r++) {
          var c = tb.rows[r].cells[col];
          if (!c) continue;
          var rawAttr = c.getAttribute('data-sui-raw');
          add(rawAttr !== null ? rawAttr : c.textContent);
        }
      }
    }
    out.sort(function(a, b){
      if (a.raw === '') return 1;
      if (b.raw === '') return -1;
      return a.label.localeCompare(b.label, undefined, { numeric: true, sensitivity: 'base' });
    });
    return out;
  }
  var tableFilterPop = null;
  function closeTableFilterPop(){
    if (!tableFilterPop) return;
    if (tableFilterPop.el && tableFilterPop.el.parentNode) tableFilterPop.el.parentNode.removeChild(tableFilterPop.el);
    if (tableFilterPop.btn) tableFilterPop.btn.setAttribute('aria-expanded', 'false');
    document.removeEventListener('click', tableFilterPop.outside, true);
    document.removeEventListener('keydown', tableFilterPop.esc, true);
    window.removeEventListener('resize', tableFilterPop.resize);
    window.removeEventListener('scroll', tableFilterPop.resize, true);
    tableFilterPop = null;
  }
  function openTableFilterPop(btn, tbl, name, col){
    closeTableFilterPop();
    var st = tableStore(name), cfg = cfgOf(tbl);
    if (!st.colValueFilters) st.colValueFilters = {};
    var pairs = columnValuePairs(tbl, name, col, cfg);
    var allValues = pairs.map(function(p){ return p.raw; });
    var active = hasValueFilter(st, col);
    var selected = active ? (st.colValueFilters[String(col)] || []).slice() : null;
    var el = cfgEl('div', 'sui-table-fpop');
    el.setAttribute('role', 'dialog');
    el.setAttribute('aria-label', 'Filter column');
    el.appendChild(cfgEl('div', 'sui-table-fpop-title', 'Filter'));
    var search = document.createElement('input');
    search.type = 'search';
    search.className = 'sui-table-fpop-search';
    search.setAttribute('placeholder', 'Search');
    search.setAttribute('aria-label', 'Search filter values');
    el.appendChild(search);
    var list = cfgEl('div', 'sui-table-fpop-list');
    el.appendChild(list);
    var allRow = cfgEl('div', 'sui-table-fpop-item');
    var allChk = document.createElement('input');
    allChk.type = 'checkbox';
    allChk.setAttribute('aria-label', 'Select all');
    allRow.appendChild(allChk);
    allRow.appendChild(cfgEl('span', 'sui-table-fpop-lbl', '(Select All)'));
    list.appendChild(allRow);
    var rows = [];
    pairs.forEach(function(pair){
      var row = cfgEl('div', 'sui-table-fpop-item');
      var chk = document.createElement('input');
      chk.type = 'checkbox';
      chk.setAttribute('aria-label', 'Filter value ' + pair.label);
      chk.checked = !active || selected.indexOf(pair.raw) >= 0;
      var lbl = cfgEl('span', 'sui-table-fpop-lbl', pair.label);
      lbl.title = pair.label;
      row.appendChild(chk);
      row.appendChild(lbl);
      row.__pair = pair;
      row.__chk = chk;
      row.addEventListener('click', function(ev){
        if (ev.target !== chk) chk.checked = !chk.checked;
        updateAll();
      });
      chk.addEventListener('change', updateAll);
      list.appendChild(row);
      rows.push(row);
    });
    if (!pairs.length) list.appendChild(cfgEl('div', 'sui-table-fpop-empty', '(No values)'));
    function visibleRows(){
      return rows.filter(function(row){ return row.style.display !== 'none'; });
    }
    function updateAll(){
      var vis = visibleRows(), checked = 0;
      vis.forEach(function(row){ if (row.__chk.checked) checked++; });
      allChk.checked = vis.length > 0 && checked === vis.length;
      allChk.indeterminate = checked > 0 && checked < vis.length;
    }
    allChk.addEventListener('change', function(){
      var target = allChk.checked;
      visibleRows().forEach(function(row){ row.__chk.checked = target; });
      updateAll();
    });
    search.addEventListener('input', function(){
      var q = search.value.toLowerCase();
      rows.forEach(function(row){
        var match = !q || row.__pair.label.toLowerCase().indexOf(q) >= 0;
        row.style.display = match ? '' : 'none';
      });
      updateAll();
    });
    var actions = cfgEl('div', 'sui-table-fpop-actions');
    var clear = cfgEl('button', 'sui-table-fpop-btn', 'Clear Filter');
    clear.type = 'button';
    clear.addEventListener('click', function(){
      var st2 = tableStore(name);
      delete st2.colValueFilters[String(col)];
      st2.page = 0;
      clearSelection(st2);
      recordChange(name, { op: 'filter', column: col });
      closeTableFilterPop();
      renderName(name);
    });
    var apply = cfgEl('button', 'sui-table-fpop-btn', 'Apply');
    apply.type = 'button';
    apply.addEventListener('click', function(){
      var st2 = tableStore(name), chosen = [];
      rows.forEach(function(row){ if (row.__chk.checked) chosen.push(row.__pair.raw); });
      if (chosen.length === allValues.length) delete st2.colValueFilters[String(col)];
      else st2.colValueFilters[String(col)] = chosen;
      st2.page = 0;
      clearSelection(st2);
      recordChange(name, { op: 'filter', column: col });
      closeTableFilterPop();
      renderName(name);
    });
    actions.appendChild(clear);
    actions.appendChild(apply);
    el.appendChild(actions);
    document.body.appendChild(el);
    var outside = function(ev){
      if (!tableFilterPop) return;
      if (el.contains(ev.target) || btn.contains(ev.target)) return;
      closeTableFilterPop();
    };
    var esc = function(ev){ if (ev.key === 'Escape') closeTableFilterPop(); };
    var resize = function(){ closeTableFilterPop(); };
    tableFilterPop = { el: el, btn: btn, name: name, col: col, outside: outside, esc: esc, resize: resize };
    btn.setAttribute('aria-expanded', 'true');
    updateAll();
    var r = btn.getBoundingClientRect();
    el.style.left = Math.max(8, Math.min(window.innerWidth - el.offsetWidth - 8, r.right - el.offsetWidth)) + 'px';
    var top = r.bottom + 4;
    if (top + el.offsetHeight > window.innerHeight - 8) top = Math.max(8, r.top - el.offsetHeight - 4);
    el.style.top = top + 'px';
    setTimeout(function(){ document.addEventListener('click', outside, true); }, 0);
    document.addEventListener('keydown', esc, true);
    window.addEventListener('resize', resize);
    window.addEventListener('scroll', resize, true);
    search.focus();
  }
  function ensureHeaderTools(hc, hwrap, hlabel, tbl, name, col, st, cfg, sortDir){
    // The header cell is contenteditable for renaming; its sort/filter controls
    // must stay outside the editable text so a caret or Ctrl+A cannot eat them.
    var hbtns = hwrap.querySelector('.sui-table-hbtns');
    if (hbtns) hbtns.setAttribute('contenteditable', 'false');
    function ensureHbtns(){
      if (!hbtns) {
        hbtns = document.createElement('span');
        hbtns.className = 'sui-table-hbtns';
        hbtns.setAttribute('contenteditable', 'false');
        hwrap.appendChild(hbtns);
      }
      return hbtns;
    }
    var fbtn = hbtns ? hbtns.querySelector('.sui-table-filterbtn') : null;
    if (st.filtersOn) {
      if (!fbtn) {
        ensureHbtns();
        fbtn = document.createElement('button');
        fbtn.type = 'button';
        fbtn.className = 'sui-table-filterbtn';
        fbtn.addEventListener('click', function(ev){
          ev.preventDefault();
          ev.stopPropagation();
          var c = parseInt(hc.getAttribute('data-col'), 10) || 0;
          if (tableFilterPop && tableFilterPop.btn === fbtn) { closeTableFilterPop(); return; }
          openTableFilterPop(fbtn, tbl, name, c);
        });
        hbtns.appendChild(fbtn);
      }
      var active = hasValueFilter(st, col);
      fbtn.classList.toggle('active', active);
      fbtn.title = active ? 'Filter (active)' : 'Filter';
      fbtn.setAttribute('aria-label', 'Filter column ' + (col + 1));
      fbtn.setAttribute('aria-expanded', (tableFilterPop && tableFilterPop.btn === fbtn) ? 'true' : 'false');
    } else if (fbtn && fbtn.parentNode) {
      if (tableFilterPop && tableFilterPop.btn === fbtn) closeTableFilterPop();
      fbtn.parentNode.removeChild(fbtn);
    }
    var sbtn = hbtns ? hbtns.querySelector('.sui-table-sortbtn') : null;
    if (st.sortOn) {
      if (!sbtn) {
        ensureHbtns();
        sbtn = document.createElement('button');
        sbtn.type = 'button';
        sbtn.className = 'sui-table-sortbtn';
        sbtn.addEventListener('click', function(ev){
          ev.preventDefault();
          ev.stopPropagation();
          var c = parseInt(hc.getAttribute('data-col'), 10) || 0;
          applyHeaderSort(name, tbl, c, hc, ev.ctrlKey || ev.metaKey || ev.shiftKey);
        });
        hbtns.appendChild(sbtn);
      }
      sbtn.textContent = '';
      sbtn.classList.toggle('asc', sortDir === 'asc');
      sbtn.classList.toggle('desc', sortDir === 'desc');
      sbtn.classList.toggle('none', sortDir !== 'asc' && sortDir !== 'desc');
      sbtn.title = sortDir === 'asc' ? 'Sorted ascending' : sortDir === 'desc' ? 'Sorted descending' : 'Sort column';
      sbtn.setAttribute('aria-label', (sortDir === 'asc' ? 'Sort column descending' : sortDir === 'desc' ? 'Clear sort' : 'Sort column ascending') + ' (' + (col + 1) + ')');
    } else if (sbtn && sbtn.parentNode) {
      sbtn.parentNode.removeChild(sbtn);
    }
    var rank = hc.getAttribute('data-sort-rank') || '';
    var rankEl = hbtns ? hbtns.querySelector('.sui-table-sort-rank') : null;
    if (rank && st.sortOn) {
      ensureHbtns();
      if (!rankEl) {
        rankEl = document.createElement('span');
        rankEl.className = 'sui-table-sort-rank';
        hbtns.appendChild(rankEl);
      }
      if (rankEl.textContent !== rank) rankEl.textContent = rank;
      rankEl.title = 'Sort level ' + rank;
    } else if (rankEl && rankEl.parentNode) {
      rankEl.parentNode.removeChild(rankEl);
    }
    if (hbtns && !hbtns.children.length && hbtns.parentNode) hbtns.parentNode.removeChild(hbtns);
    hc.classList.toggle('sui-table-filter-active', hasValueFilter(st, col));
  }
  function positionTableCfg(){
    if (!tableCfg) return;
    var btn = tableCfg.btn, el = tableCfg.el;
    if (!btn || !el) return;
    var r = btn.getBoundingClientRect();
    el.style.left = Math.max(8, Math.min(window.innerWidth - el.offsetWidth - 8, r.right - el.offsetWidth)) + 'px';
    var top = r.bottom + 6;
    if (top + el.offsetHeight > window.innerHeight - 8) top = Math.max(8, r.top - el.offsetHeight - 6);
    el.style.top = top + 'px';
  }
  function closeTableCfg(){
    if (!tableCfg) return;
    if (tableCfg.el && tableCfg.el.parentNode) tableCfg.el.parentNode.removeChild(tableCfg.el);
    if (tableCfg.btn) tableCfg.btn.setAttribute('aria-expanded', 'false');
    document.removeEventListener('click', tableCfg.outside, true);
    document.removeEventListener('keydown', tableCfg.esc, true);
    window.removeEventListener('resize', tableCfg.resize);
    window.removeEventListener('scroll', tableCfg.resize, true);
    tableCfg = null;
  }
  function sortKeysCopy(st){
    return sortKeysOf(st).map(function(k){ return { col: k.col, dir: k.dir }; });
  }
  function renderTableCfg(){
    if (!tableCfg) return;
    var name = tableCfg.name, tbl = firstTable(name);
    if (!tbl) { closeTableCfg(); return; }
    var st = tableStore(name), cfg = cfgOf(tbl), box = tableCfg.el;
    box.textContent = '';
    var head = cfgEl('div', 'sui-table-cfg-head');
    head.appendChild(cfgEl('span', 'sui-table-cfg-title', 'Table settings'));
    var close = cfgEl('button', 'sui-table-cfg-close', '\u00d7');
    close.type = 'button';
    close.setAttribute('aria-label', 'Close table settings');
    close.addEventListener('click', closeTableCfg);
    head.appendChild(close);
    box.appendChild(head);

    var frow = cfgEl('div', 'sui-table-cfg-row');
    var fchk = document.createElement('input');
    fchk.type = 'checkbox';
    fchk.checked = !!st.filtersOn;
    fchk.setAttribute('aria-label', 'Filter on column headers');
    fchk.addEventListener('change', function(){
      st.filtersOn = fchk.checked;
      if (!st.filtersOn) {
        st.colFilters = {};
        st.colValueFilters = {};
        st.colFilter = null;
        closeTableFilterPop();
      }
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    frow.appendChild(fchk);
    frow.appendChild(cfgEl('span', '', 'Filter on column headers'));
    box.appendChild(frow);
    box.appendChild(cfgEl('div', 'sui-table-cfg-empty', 'Use the funnel button in each header to pick column values.'));

    var srow = cfgEl('div', 'sui-table-cfg-row');
    var schk = document.createElement('input');
    schk.type = 'checkbox';
    schk.checked = !!st.sortOn;
    schk.setAttribute('aria-label', 'Sort by clicking headers');
    schk.addEventListener('change', function(){
      st.sortOn = schk.checked;
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    srow.appendChild(schk);
    srow.appendChild(cfgEl('span', '', 'Sort by clicking headers'));
    box.appendChild(srow);

    var sect = cfgEl('div', 'sui-table-cfg-sect');
    sect.appendChild(cfgEl('div', 'sui-table-cfg-sec', 'Sort levels'));
    var keys = sortKeysCopy(st);
    var choices = columnChoices(tbl, cfg);
    if (!keys.length) sect.appendChild(cfgEl('div', 'sui-table-cfg-empty', 'No sort levels. Add one, or click a header (Ctrl/Shift+click adds a level).'));
    keys.forEach(function(key, i){
      var row = cfgEl('div', 'sui-table-cfg-level');
      var colSel = document.createElement('select');
      colSel.setAttribute('aria-label', 'Sort column ' + (i + 1));
      choices.forEach(function(ch){
        var o = document.createElement('option');
        o.value = String(ch.index);
        o.textContent = ch.label;
        colSel.appendChild(o);
      });
      colSel.value = String(key.col);
      colSel.addEventListener('change', function(){
        key.col = parseInt(colSel.value, 10) || 0;
        setSortKeys(st, keys);
        st.page = 0;
        clearSelection(st);
        renderName(name);
        renderTableCfg();
      });
      var dirSel = document.createElement('select');
      dirSel.setAttribute('aria-label', 'Sort direction ' + (i + 1));
      [['asc', 'Ascending'], ['desc', 'Descending']].forEach(function(d){
        var o = document.createElement('option');
        o.value = d[0];
        o.textContent = d[1];
        dirSel.appendChild(o);
      });
      dirSel.value = key.dir === 'desc' ? 'desc' : 'asc';
      dirSel.addEventListener('change', function(){
        key.dir = dirSel.value === 'desc' ? 'desc' : 'asc';
        setSortKeys(st, keys);
        st.page = 0;
        clearSelection(st);
        renderName(name);
        renderTableCfg();
      });
      var rm = cfgEl('button', 'sui-table-cfg-btn', '\u00d7');
      rm.type = 'button';
      rm.title = 'Remove sort level';
      rm.setAttribute('aria-label', 'Remove sort level ' + (i + 1));
      rm.addEventListener('click', function(){
        keys.splice(i, 1);
        setSortKeys(st, keys);
        st.page = 0;
        clearSelection(st);
        renderName(name);
        renderTableCfg();
      });
      row.appendChild(colSel);
      row.appendChild(dirSel);
      row.appendChild(rm);
      sect.appendChild(row);
    });
    var add = cfgEl('button', 'sui-table-cfg-btn', '+ Add level');
    add.type = 'button';
    add.addEventListener('click', function(){
      var k2 = sortKeysCopy(st);
      var used = {};
      k2.forEach(function(k){ used[k.col] = 1; });
      var pick = -1;
      for (var c = 0; c < choices.length; c++) {
        if (!used[choices[c].index]) { pick = choices[c].index; break; }
      }
      if (pick < 0 && choices.length) pick = choices[0].index;
      if (pick < 0) return;
      k2.push({ col: pick, dir: 'asc' });
      setSortKeys(st, k2);
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    sect.appendChild(add);
    box.appendChild(sect);

    var acts = cfgEl('div', 'sui-table-cfg-actions');
    var clearF = cfgEl('button', 'sui-table-cfg-btn', 'Clear filters');
    clearF.type = 'button';
    clearF.addEventListener('click', function(){
      st.colFilters = {};
      st.colValueFilters = {};
      st.colFilter = null;
      st.filter = '';
      st.page = 0;
      clearSelection(st);
      closeTableFilterPop();
      renderName(name);
      renderTableCfg();
    });
    var clearS = cfgEl('button', 'sui-table-cfg-btn', 'Clear sort');
    clearS.type = 'button';
    clearS.addEventListener('click', function(){
      setSortKeys(st, []);
      st.page = 0;
      clearSelection(st);
      renderName(name);
      renderTableCfg();
    });
    var done = cfgEl('button', 'sui-table-cfg-btn', 'Done');
    done.type = 'button';
    done.addEventListener('click', closeTableCfg);
    acts.appendChild(clearF);
    acts.appendChild(clearS);
    if (cfg.editMode === 'persistent') {
      var reset = cfgEl('button', 'sui-table-cfg-btn', 'Reset saved data');
      reset.type = 'button';
      reset.addEventListener('click', function(){
        if (!window.confirm('Discard saved table data and restore the design-time table?')) return;
        clearPersisted(name);
        closeTableCfg();
      });
      acts.appendChild(reset);
    }
    acts.appendChild(done);
    box.appendChild(acts);
    positionTableCfg();
  }
  function openTableCfg(tbl, name, btn){
    closeTableCfg();
    var el = cfgEl('div', 'sui-table-cfg');
    el.setAttribute('role', 'dialog');
    el.setAttribute('aria-label', 'Table settings');
    document.body.appendChild(el);
    var outside = function(ev){
      if (!tableCfg) return;
      if (el.contains(ev.target) || (btn && btn.contains(ev.target))) return;
      closeTableCfg();
    };
    var esc = function(ev){ if (ev.key === 'Escape') closeTableCfg(); };
    var resize = function(){ positionTableCfg(); };
    tableCfg = { el: el, name: name, btn: btn, outside: outside, esc: esc, resize: resize };
    btn.setAttribute('aria-expanded', 'true');
    renderTableCfg();
    setTimeout(function(){ document.addEventListener('click', outside, true); }, 0);
    document.addEventListener('keydown', esc, true);
    window.addEventListener('resize', resize);
    window.addEventListener('scroll', resize, true);
  }
  function ensureTableTools(tbl, name){
    var wrap = tbl.closest ? tbl.closest('.tscroll') : null;
    var anchor = wrap || tbl;
    var parent = anchor.parentNode;
    if (!parent) return;
    var bar = null, kids = parent.children;
    for (var i = 0; i < kids.length; i++) {
      if (kids[i].classList && kids[i].classList.contains('sui-table-topbar') &&
          kids[i].getAttribute('data-sui-table-topbar-for') === name) { bar = kids[i]; break; }
    }
    if (!bar) {
      bar = document.createElement('div');
      bar.className = 'sui-table-topbar';
      bar.setAttribute('data-sui-table-topbar-for', name);
      parent.insertBefore(bar, anchor);
    } else if (bar.nextElementSibling !== anchor) {
      parent.insertBefore(bar, anchor);
    }
    // The caption becomes the title row above the table, on the same line as
    // the settings gear. The original <caption> is hidden to avoid duplicating
    // the title inside the table.
    var cap = tbl.querySelector('caption');
    var title = bar.querySelector('.sui-table-title');
    if (cap && (cap.textContent || '').trim()) {
      if (!title) {
        title = document.createElement('div');
        title.className = 'sui-table-title';
        bar.insertBefore(title, bar.firstChild);
      }
      var capText = cap.textContent;
      if (title.textContent !== capText) title.textContent = capText;
      if (cap.getAttribute('data-sui-caption-style') === null) {
        cap.setAttribute('data-sui-caption-style', cap.getAttribute('style') || '');
      }
      var capStyle = cap.getAttribute('data-sui-caption-style') || '';
      if ((title.getAttribute('style') || '') !== capStyle) {
        if (capStyle) title.setAttribute('style', capStyle); else title.removeAttribute('style');
        title.style.display = '';
      }
      if (cap.style.display !== 'none') cap.style.display = 'none';
    } else {
      if (title && title.parentNode) title.parentNode.removeChild(title);
      if (cap && cap.style.display === 'none') cap.style.display = '';
    }
    var host = bar.querySelector('.sui-table-tools');
    if (!host) {
      host = document.createElement('div');
      host.className = 'sui-table-tools';
      bar.appendChild(host);
    }
    var btn = host.querySelector('.sui-table-settings');
    if (!btn) {
      btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'sui-table-settings';
      btn.textContent = '\u2699';
      btn.title = 'Table settings';
      btn.setAttribute('aria-label', 'Table settings');
      btn.setAttribute('aria-expanded', 'false');
      host.appendChild(btn);
      btn.addEventListener('click', function(ev){
        ev.preventDefault();
        ev.stopPropagation();
        if (tableCfg && tableCfg.name === name && tableCfg.btn === btn) { closeTableCfg(); return; }
        openTableCfg(tbl, name, btn);
      });
    }
    var w = (wrap || tbl).getBoundingClientRect().width;
    if (w > 0 && Math.abs(bar.offsetWidth - w) > 1) bar.style.width = Math.round(w) + 'px';
  }
  function ensurePager(tbl, name, st, cfg){
    var next = tbl.nextElementSibling;
    var pager = next && next.classList && next.classList.contains('sui-table-pager') &&
      next.getAttribute('data-sui-table-pager-for') === name ? next : null;
    var pages = st.pageCount || 1;
    if (!cfg.pageSize || pages <= 1) {
      if (pager) pager.style.display = 'none';
      return;
    }
    if (!pager) {
      pager = document.createElement('div');
      pager.className = 'sui-table-pager';
      pager.setAttribute('data-sui-table-pager-for', name);
      if (tbl.parentNode) tbl.parentNode.insertBefore(pager, tbl.nextSibling);
    }
    pager.style.display = '';
    pager.textContent = '';
    function btn(label, delta, disabled) {
      var b = document.createElement('button');
      b.type = 'button';
      b.className = 'sui-table-page-btn';
      b.textContent = label;
      b.disabled = !!disabled;
      b.addEventListener('click', function(){
        var st2 = tableStore(name);
        var p = (st2.page || 0) + delta;
        if (p < 0 || p >= (st2.pageCount || 1)) return;
        st2.page = p;
        clearSelection(st2);
        renderName(name);
      });
      pager.appendChild(b);
    }
    btn('\u2039 Prev', -1, (st.page || 0) <= 0);
    var info = document.createElement('span');
    info.className = 'sui-table-page-info';
    info.textContent = 'Page ' + ((st.page || 0) + 1) + ' / ' + pages;
    pager.appendChild(info);
    btn('Next \u203a', 1, (st.page || 0) >= pages - 1);
  }
  function renderName(name){
    var ts = findTables(name);
    if (!ts.length) return;
    var view = applyView(name), st = view.st, cfg = view.cfg;
    for (var i = 0; i < ts.length; i++) {
      applyTo(ts[i], st);
      decorateTable(ts[i], st, cfg);
      ensureFilterInput(ts[i], name, st, cfg);
      ensurePager(ts[i], name, st, cfg);
      ensureAddTableButton(ts[i], name, st, cfg);
    }
    init();
  }
  function storeApply(name, ds){
    var st = tableStore(name);
    st.header = ds.header ? ds.header.map(function(v){ return String(v); }) : [];
    st.allRows = ds.rows ? ds.rows.map(function(r){ return r.slice(); }) : [];
    st.textHeader = ds.textHeader;
    st.sep = ds.sep;
    st.page = 0;
    st.__fromDom = false;
    st.__colStruct = false;
    clearSelection(st);
    store[name] = st;
    renderName(name);
    return { n: st.allRows.length };
  }
  function display(name, data){
    if (persistentBootBlocks(name)) {
      var kept = store[name] ? store[name].allRows.length : 0;
      return [kept, null];
    }
    var ds = norm(data);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    if (r.error) return [0, r.error];
    recordChange(name, { op: 'display', count: r.n });
    fireTableEvent(name, 'ontablechange');
    return [r.n, null];
  }
  // importText parses text explicitly: see parseSpec for the spec syntax.
  function importText(name, text, spec){
    if (persistentBootBlocks(name)) {
      var keptImport = store[name] ? store[name].allRows.length : 0;
      return [keptImport, null];
    }
    var ds = normText(text, spec);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    if (r.error) return [0, r.error];
    recordChange(name, { op: 'import', count: r.n });
    fireTableEvent(name, 'ontablechange');
    return [r.n, null];
  }
  function data(name){
    var st = store[name];
    if (st) return [st.rows.map(function(r){ return r.slice(); }), null];
    var ts = findTables(name);
    if (!ts.length) return [[], null];
    var out = [], rows = bodyRows(ts[0]);
    for (var i = 0; i < rows.length; i++) {
      var cs = cellsOf(rows[i]), r = [];
      for (var j = 0; j < cs.length; j++) r.push(cs[j].hasAttribute('data-sui-raw') ? (cs[j].getAttribute('data-sui-raw') || '') : (cs[j].textContent || '').trim());
      out.push(r);
    }
    return [out, null];
  }
  function header(name){
    var st = store[name];
    if (st && st.header && st.header.length) return [st.header.slice(), null];
    var ts = findTables(name), out = [];
    if (!ts.length) return [out, null];
    var hr = ts[0].querySelector('thead tr');
    if (hr) {
      var cs = cellsOf(hr);
      for (var i = 0; i < cs.length; i++) out.push((cs[i].textContent || '').trim());
    }
    return [out, null];
  }
  function json(name){
    var h = header(name)[0], rows = data(name)[0], out = [], i, j;
    if (h && h.length) {
      for (i = 0; i < rows.length; i++) {
        var o = Object.create(null);
        for (j = 0; j < rows[i].length; j++) o[h[j] || ('col' + (j + 1))] = rows[i][j];
        out.push(o);
      }
      return [JSON.stringify(out), null];
    }
    return [JSON.stringify(rows), null];
  }
  function select(name, cols){
    if (!Array.isArray(cols)) return [[], 'table.select: columns must be an array of indexes'];
    var rows = data(name)[0], out = [];
    for (var i = 0; i < rows.length; i++) {
      var r = [];
      for (var j = 0; j < cols.length; j++) {
        var c = Number(cols[j]);
        r.push(c >= 0 && c < rows[i].length ? rows[i][c] : '');
      }
      out.push(r);
    }
    return [out, null];
  }
  function clear(name){
    var clearTbl = firstTable(name);
    var clearCfg = clearTbl ? cfgOf(clearTbl) : null;
    delete store[name];
    if (tableCfg && tableCfg.name === name) closeTableCfg();
    if (tableFilterPop && tableFilterPop.name === name) closeTableFilterPop();
    var ts = findTables(name);
    for (var i = 0; i < ts.length; i++) {
      var tb = ts[i].tBodies && ts[i].tBodies.length ? ts[i].tBodies[0] : null;
      if (tb) while (tb.firstChild) tb.removeChild(tb.firstChild);
      var wrap = ts[i].closest ? ts[i].closest('.tscroll') : null;
      var anchor = wrap || ts[i];
      if (anchor.parentNode) {
        var bar = null, kids = anchor.parentNode.children;
        for (var b = 0; b < kids.length; b++) {
          if (kids[b].classList && kids[b].classList.contains('sui-table-topbar') && kids[b].getAttribute('data-sui-table-topbar-for') === name) { bar = kids[b]; break; }
        }
        if (bar) {
          var cap = ts[i].querySelector('caption');
          if (cap) cap.style.display = '';
          bar.parentNode.removeChild(bar);
        }
      }
      var prev = ts[i].previousElementSibling;
      if (prev && prev.classList && prev.classList.contains('sui-table-tools') && prev.getAttribute('data-sui-table-tools-for') === name) {
        var fprev = prev.previousElementSibling;
        prev.parentNode.removeChild(prev);
        prev = fprev;
      }
      if (prev && prev.classList && prev.classList.contains('sui-table-filter') && prev.getAttribute('data-sui-table-filter-for') === name) prev.parentNode.removeChild(prev);
      var next = ts[i].nextElementSibling;
      if (next && next.classList && next.classList.contains('sui-table-pager') && next.getAttribute('data-sui-table-pager-for') === name) next.parentNode.removeChild(next);
    }
    persistTableState(name, { header: [], allRows: [], textHeader: false }, clearCfg);
    recordChange(name, { op: 'clear' });
    fireTableEvent(name, 'ontablechange');
    return [0, null];
  }
  function sort(name, column, dir){
    var st = tableStore(name);
    var tbl = firstTable(name), cfg = tbl ? cfgOf(tbl) : { columns: [], labels: [] };
    var idx = colIndex(cfg, column);
    if (idx < 0) return [0, 'table.sort: column not found'];
    var d = String(dir || '').toLowerCase();
    if (d !== 'asc' && d !== 'desc' && d !== '') return [0, 'table.sort: direction must be asc or desc'];
    setSortKeys(st, d ? [{ col: idx, dir: d }] : []);
    st.page = 0;
    clearSelection(st);
    recordChange(name, { op: 'sort', column: idx, dir: d });
    renderName(name);
    fireTableEvent(name, 'onsort');
    return [1, null];
  }
  function filter(name, column, value){
    var st = tableStore(name);
    var tbl = firstTable(name), cfg = tbl ? cfgOf(tbl) : { columns: [], labels: [] };
    var idx = colIndex(cfg, column);
    if (idx < 0) return [0, 'table.filter: column not found'];
    var fv = String(value === undefined || value === null ? '' : value);
    st.colFilters = {};
    if (fv) st.colFilters[idx] = fv;
    st.colFilter = { column: idx, value: fv };
    st.page = 0;
    clearSelection(st);
    recordChange(name, { op: 'filter', column: idx, value: fv });
    renderName(name);
    return [1, null];
  }
  function filterText(name, text){
    var st = tableStore(name);
    st.filter = String(text === undefined || text === null ? '' : text);
    st.page = 0;
    clearSelection(st);
    recordChange(name, { op: 'filter', column: '', value: st.filter });
    renderName(name);
    return [1, null];
  }
  function page(name, index){
    var st = tableStore(name);
    var p = parseInt(index, 10);
    if (!isFinite(p) || p < 0) return [0, 'table.page: invalid page'];
    st.page = p;
    clearSelection(st);
    renderName(name);
    return [1, null];
  }
  function setPageSize(name, size){
    var st = tableStore(name);
    var n = parseInt(size, 10);
    if (!isFinite(n) || n < 0) return [0, 'table.setPageSize: invalid size'];
    st.pageSize = n;
    st.page = 0;
    renderName(name);
    return [1, null];
  }
  function selectedRows(name){ var st = tableStore(name), out = []; for (var k in st.selectedRows) if (st.selectedRows[k]) out.push(parseInt(k, 10)); return [out, null]; }
  function selectedCells(name){ var st = tableStore(name), out = []; for (var k in st.selectedCells) if (st.selectedCells[k]) { var p = k.split(':'); out.push({ row: parseInt(p[0], 10), column: parseInt(p[1], 10) }); } return [JSON.stringify(out), null]; }
  function selectedColumns(name){ var st = tableStore(name), out = []; for (var k in st.selectedCols) if (st.selectedCols[k]) out.push(parseInt(k, 10)); return [JSON.stringify(out), null]; }
  function lastChange(name){ return [JSON.stringify(changes[name] || null), null]; }
  function applyAll(){
    for (var k in store) {
      try { renderName(k); } catch(e) {}
    }
  }
  function slice(name, from, count){
    var rows = data(name)[0];
    var a = Math.max(0, Number(from) || 0);
    var n = Number(count) || 0;
    return [n > 0 ? rows.slice(a, a + n) : rows.slice(a), null];
  }
  // exportText serializes the visible table as CSV/TSV/SSV/PSV/JSON/JSONL.
  function exportText(name, spec){
    var o = parseSpec(spec, 'csv');
    if (o.format === 'json') return json(name);
    var h = header(name)[0], rows = data(name)[0], i, j, ob;
    if (o.format === 'jsonl' || o.format === 'ndjson') {
      var objs = [];
      if (h.length) {
        for (i = 0; i < rows.length; i++) {
          ob = Object.create(null);
          for (j = 0; j < rows[i].length; j++) ob[h[j] || ('col' + (j + 1))] = rows[i][j];
          objs.push(ob);
        }
      } else {
        objs = rows;
      }
      var txt = objs.map(function(v){ return JSON.stringify(v); }).join('\n');
      return [txt ? txt + '\n' : '', null];
    }
    var sep = o.sep || delimFor(o.format);
    var q = o.quote === undefined ? '"' : o.quote;
    function esc(v){
      v = v === null || v === undefined ? '' : String(v);
      // CSV/Excel formula injection: opt-in prefix for cells that a
      // spreadsheet would evaluate as a formula.
      if (o.formula === 'escape' && /^[=+\-@\t\r]/.test(v)) v = "'" + v;
      if (q && (v.indexOf(sep) >= 0 || v.indexOf(q) >= 0 || v.indexOf('\n') >= 0 || v.indexOf('\r') >= 0)) {
        return q + v.split(q).join(q + q) + q;
      }
      return v;
    }
    var lines = [];
    var useHeader = o.header === null ? true : !!o.header;
    if (useHeader && h.length) lines.push(h.map(esc).join(sep));
    for (i = 0; i < rows.length; i++) lines.push(rows[i].map(esc).join(sep));
    return [lines.join('\n'), null];
  }
  function afterRender(){
    restoreTableBooks();
    applyAll();
    init();
    // Restore the selected tab synchronously, before the browser can paint the
    // regenerated view with the first tab active.
    if (window.__SUI_TABS && window.__SUI_TABS.restore) window.__SUI_TABS.restore();
  }

  // ---- editable / fixed behaviors -----------------------------------------
  function ensureDeclaredHeader(tbl, cfg){
    if (tbl.querySelector('thead tr')) return;
    var labels = cfg.labels.length ? cfg.labels : cfg.columns;
    if (!labels.length) return;
    var thead = tbl.querySelector('thead');
    if (!thead) { thead = document.createElement('thead'); tbl.insertBefore(thead, tbl.firstChild); }
    var tr = document.createElement('tr');
    for (var i = 0; i < labels.length; i++) {
      var th = document.createElement('th');
      th.textContent = labels[i];
      if (cfg.widths[i] > 0) { th.style.width = cfg.widths[i] + 'px'; th.style.minWidth = cfg.widths[i] + 'px'; }
      if (cfg.align[i]) th.style.textAlign = cfg.align[i];
      tr.appendChild(th);
    }
    thead.appendChild(tr);
  }
  function syncStoreFromDom(tbl, name){
    var st = tableStore(name);
    if (st.__fromDom || st.allRows.length) return;
    st.__fromDom = true;
    st.header = [];
    st.allRows = [];
    var hr = tbl.querySelector('thead tr');
    if (hr) cellsOf(hr).forEach(function(c){ st.header.push(c.getAttribute('data-sui-raw') || c.textContent || ''); });
    var rows = bodyRows(tbl);
    for (var i = 0; i < rows.length; i++) {
      var cs = cellsOf(rows[i]), r = [];
      for (var j = 0; j < cs.length; j++) r.push(cs[j].getAttribute('data-sui-raw') || cs[j].textContent || '');
      st.allRows.push(r);
    }
    st.rows = st.allRows;
    store[name] = st;
  }
  // Runtime workbook sheet deletion is kept for the page session. Source sheets
  // removed at runtime stay hidden across view re-renders; runtime sheets are
  // also dropped from tableBooks so restoreTableBooks cannot recreate them.
  var deletedWorkbookSheets = {};
  function workbookTabList(box){
    var list = box ? box.querySelector('.sui-tablist') : null;
    return list ? Array.prototype.filter.call(list.children, function(el){
      return el.classList && el.classList.contains('sui-tab');
    }) : [];
  }
  function workbookPanelList(box){
    var panels = box ? box.querySelector('.sui-tabpanels') : null;
    return panels ? Array.prototype.filter.call(panels.children, function(el){
      return el.classList && el.classList.contains('sui-tabpanel');
    }) : [];
  }
  function workbookSheetLabel(tab){
    return (tab.getAttribute('data-sui-sheet-label') || tab.textContent.replace(/\u00d7\s*$/,'').trim());
  }
  function workbookSheetKey(table, label){
    var name = table ? (table.getAttribute('name') || '') : '';
    return name ? ('n:' + name) : ('l:' + label);
  }
  function ensureWorkbookSheetLabels(box){
    var tabs = workbookTabList(box);
    for (var i = 0; i < tabs.length; i++) {
      if (!tabs[i].hasAttribute('data-sui-sheet-label')) {
        tabs[i].setAttribute('data-sui-sheet-label', workbookSheetLabel(tabs[i]));
      }
    }
  }
  function applyWorkbookDeletions(box){
    var dkey = tableBookKey(box);
    var marks = deletedWorkbookSheets[dkey];
    if (!marks) return;
    var tabs = workbookTabList(box), panels = workbookPanelList(box);
    var activeRemoved = false;
    for (var i = tabs.length - 1; i >= 0; i--) {
      var panel = panels[i];
      var table = panel ? panel.querySelector('table') : null;
      var key = workbookSheetKey(table, workbookSheetLabel(tabs[i]));
      if (!marks[key]) continue;
      if (tabs[i].classList.contains('active')) activeRemoved = true;
      if (tabs[i].parentNode) tabs[i].parentNode.removeChild(tabs[i]);
      if (panel && panel.parentNode) panel.parentNode.removeChild(panel);
    }
    if (activeRemoved) {
      var rest = workbookTabList(box);
      if (rest.length) { try { rest[0].click(); } catch (e) {} }
    }
  }
  function deleteWorkbookSheet(box, index){
    var tabs = workbookTabList(box);
    if (tabs.length <= 1) return;
    var tab = tabs[index];
    if (!tab) return;
    ensureWorkbookSheetLabels(box);
    var panels = workbookPanelList(box);
    var panel = panels[index];
    var table = panel ? panel.querySelector('table') : null;
    var key = workbookSheetKey(table, workbookSheetLabel(tab));
    var dkey = tableBookKey(box);
    var marks = deletedWorkbookSheets[dkey] || (deletedWorkbookSheets[dkey] = {});
    marks[key] = true;
    var name = table ? (table.getAttribute('name') || '') : '';
    if (name && tableBooks[dkey]) {
      tableBooks[dkey] = tableBooks[dkey].filter(function(def){ return def.name !== name; });
    }
    var wasActive = tab.classList.contains('active');
    if (tab.parentNode) tab.parentNode.removeChild(tab);
    if (panel && panel.parentNode) panel.parentNode.removeChild(panel);
    if (name) {
      try { if (window.localStorage) window.localStorage.removeItem(tablePersistKey(name)); } catch (e) {}
    }
    saveTableBookDefs(box);
    // Update the visible tab count, editable + state and the x buttons.
    normalizeWorkbookRuntime();
    var rest = workbookTabList(box);
    if (rest.length) {
      var active = -1;
      for (var i = 0; i < rest.length; i++) if (rest[i].classList.contains('active')) { active = i; break; }
      if (active < 0) active = wasActive ? Math.min(index, rest.length - 1) : 0;
      // Re-click the active tab so the tabs runtime stores the shifted index.
      try { rest[active].click(); } catch (e) {}
    }
  }
  function ensureWorkbookDeleteButtons(box){
    var designer = !!(document.body && document.body.classList && document.body.classList.contains('sui-designer-preview'));
    var tabs = workbookTabList(box), panels = workbookPanelList(box);
    ensureWorkbookSheetLabels(box);
    for (var i = 0; i < tabs.length; i++) {
      var panel = panels[i];
      var table = panel ? panel.querySelector('table') : null;
      var editable = !!(table && table.getAttribute('data-sui-editable') === '1');
      var x = tabs[i].querySelector('.sui-tab-x');
      if (!editable && !designer) {
        if (x && x.parentNode) x.parentNode.removeChild(x);
        continue;
      }
      if (x) continue;
      x = document.createElement('span');
      x.className = 'sui-tab-x';
      x.textContent = '\u00d7';
      x.title = 'Delete workbook';
      x.setAttribute('aria-label', 'Delete workbook');
      x.contentEditable = 'false';
      x.addEventListener('click', function(ev){
        ev.preventDefault();
        ev.stopPropagation();
        var tab = this.parentNode;
        var book = tab && tab.closest ? tab.closest('.sui-tabs') : null;
        if (!book) return;
        var list = workbookTabList(book);
        var idx = Array.prototype.indexOf.call(list, tab);
        if (idx >= 0) deleteWorkbookSheet(book, idx);
      });
      tabs[i].appendChild(x);
    }
  }
  // Runtime tab context menu: Delete workbook on an editable sheet. The
  // designer picker intercepts the contextmenu event first in its own preview.
  document.addEventListener('contextmenu', function(ev){
    var tab = ev.target && ev.target.closest ? ev.target.closest('.sui-tab') : null;
    if (!tab) return;
    var box = tab.closest('.sui-tabs');
    if (!box) return;
    var tabs = workbookTabList(box);
    if (tabs.length <= 1) return;
    var idx = Array.prototype.indexOf.call(tabs, tab);
    var panels = workbookPanelList(box);
    var table = panels[idx] ? panels[idx].querySelector('table') : null;
    if (!table || table.getAttribute('data-sui-editable') !== '1') return;
    ev.preventDefault();
    ev.stopPropagation();
    var old = document.getElementById('sui-tab-menu-runtime');
    if (old) old.remove();
    var menu = document.createElement('div');
    menu.id = 'sui-tab-menu-runtime';
    menu.className = 'sui-table-struct-menu';
    menu.setAttribute('role', 'menu');
    var item = document.createElement('button');
    item.type = 'button';
    item.className = 'sui-table-struct-item';
    item.textContent = 'Delete workbook';
    item.addEventListener('click', function(e){
      e.stopPropagation();
      menu.remove();
      deleteWorkbookSheet(box, idx);
    });
    menu.appendChild(item);
    document.body.appendChild(menu);
    var x = ev.clientX, y = ev.clientY;
    var rect = menu.getBoundingClientRect();
    if (x + rect.width > window.innerWidth - 8) x = Math.max(8, window.innerWidth - rect.width - 8);
    if (y + rect.height > window.innerHeight - 8) y = Math.max(8, window.innerHeight - rect.height - 8);
    menu.style.left = x + 'px';
    menu.style.top = y + 'px';
    var close = function(){ menu.remove(); document.removeEventListener('click', close); };
    setTimeout(function(){ document.addEventListener('click', close); }, 0);
  });
  // A tabs control whose panels all hold a table is a workbook: every sheet is
  // a data-entry surface, shares one visible table size and offers the + tab.
  // This also upgrades sources built before the designer normalized workbooks.
  function normalizeWorkbookRuntime(){
    var boxes = document.querySelectorAll('.sui-tabs');
    for (var b = 0; b < boxes.length; b++) {
      var panels = boxes[b].querySelectorAll('.sui-tabpanels .sui-tabpanel');
      if (!panels.length) continue;
      applyWorkbookDeletions(boxes[b]);
      panels = boxes[b].querySelectorAll('.sui-tabpanels .sui-tabpanel');
      if (!panels.length) continue;
      var tables = [], all = true;
      for (var p = 0; p < panels.length; p++) {
        var t = panels[p].querySelector('table');
        if (!t) { all = false; break; }
        tables.push(t);
      }
      if (!all || !tables.length) continue;
      boxes[b].setAttribute('data-sui-workbook', '1');
      // A single workbook has no sheet tabs; the + tab remains available to
      // create the second sheet. Two or more sheets show their tabs.
      var tabBtns = boxes[b].querySelectorAll('.sui-tablist .sui-tab');
      boxes[b].setAttribute('data-sui-tabcount', String(tabBtns.length));
      var anyEditable = false;
      for (var tj = 0; tj < tables.length; tj++) {
        var t2 = tables[tj];
        var sheetMode = t2.getAttribute('data-sui-edit-mode') || 'none';
        if (sheetMode === 'none') {
          // Non-editable sheet: keep it read-only even inside a workbook that is
          // otherwise editable.
          t2.removeAttribute('data-sui-editable');
          t2.removeAttribute('data-sui-show-add-table');
        } else {
          t2.setAttribute('data-sui-editable', '1');
          // Keep the + in the DOM for workbooks: CSS hides it when no sheet is
          // editable (the designer preview shows it).
          if (t2.getAttribute('name')) t2.setAttribute('data-sui-show-add-table', '1');
        }
        if (t2.getAttribute('data-sui-editable') === '1') anyEditable = true;
        // Stable cell sizing: truncate is the default, wrap is explicit.
        var cellText = t2.getAttribute('data-sui-cell-text') || '';
        if (!cellText) cellText = 'truncate';
        t2.setAttribute('data-sui-cell-text', cellText);
        if (cellText === 'truncate') t2.setAttribute('data-sui-fixed', '1');
        var w2 = t2.parentElement;
        if (w2 && w2.classList && w2.classList.contains('tscroll')) {
          // The sheet area follows the control size; per-sheet fixed sizes made
          // the sheets diverge and broke the shared workbook layout.
          w2.style.width = '';
          w2.style.height = '';
          w2.style.maxWidth = '';
          w2.style.maxHeight = '';
        }
      }
      // The runtime + is shown only for an editable workbook; the designer
      // preview overrides this with its body class and always offers it.
      boxes[b].setAttribute('data-sui-workbook-editable', anyEditable ? '1' : '0');
      ensureWorkbookDeleteButtons(boxes[b]);
    }
  }
  // The workbook table width is driven by the declared/default column widths,
  // never by cell text. When the sum exceeds the viewport the table overflows
  // and the sheet area shows the horizontal scrollbar.
  function updateWorkbookTableWidth(tbl){
    if (!tbl || !tbl.closest) return;
    var box = tbl.closest('.sui-tabs');
    if (!box || box.getAttribute('data-sui-workbook') !== '1') return;
    var row = (tbl.tHead && tbl.tHead.rows.length) ? tbl.tHead.rows[0] :
      ((tbl.tBodies && tbl.tBodies.length && tbl.tBodies[0].rows.length) ? tbl.tBodies[0].rows[0] : null);
    if (!row) return;
    var sum = 0;
    for (var i = 0; i < row.cells.length; i++) {
      var c = row.cells[i];
      var sw = (c.style && (c.style.width || c.style.minWidth)) || '';
      var w = parseFloat(sw);
      if (!isFinite(w) || sw.indexOf('%') >= 0) w = 160;
      sum += w * (c.colSpan || 1);
    }
    if (sum > 0) tbl.style.minWidth = Math.round(sum) + 'px';
  }
  function init(){
    normalizeWorkbookRuntime();
    var edTables = document.querySelectorAll('table[data-sui-editable]');
    for (var et = 0; et < edTables.length; et++) {
      var etMode = cfgOf(edTables[et]).editMode;
      var etCells = edTables[et].querySelectorAll('td,th');
      for (var ec = 0; ec < etCells.length; ec++) {
        if (etMode === 'none') {
          if (etCells[ec].getAttribute('contenteditable') === 'true') etCells[ec].removeAttribute('contenteditable');
        } else if (etCells[ec].getAttribute('contenteditable') !== 'true') {
          etCells[ec].setAttribute('contenteditable', 'true');
          etCells[ec].spellcheck = false;
        }
      }
    }
    var fix = document.querySelectorAll('table[data-sui-fixed] td, table[data-sui-fixed] th');
    for (var j=0;j<fix.length;j++){
      var d=fix[j];
      if (!d.hasAttribute('tabindex')) d.tabIndex=0;
      var isEd = d.getAttribute('contenteditable') === 'true';
      if (!isEd) { var txt = d.textContent; if (d.getAttribute('data-full') !== txt) d.setAttribute('data-full', txt); }
    }
    restoreTableBooks();
    // Restored runtime sheets change the tab count: update the workbook tab
    // visibility before decorating the restored tables.
    normalizeWorkbookRuntime();
    var named = document.querySelectorAll('table[name]');
    for (var t = 0; t < named.length; t++) {
      var tbl = named[t], name = tableName(tbl);
      if (!name) continue;
      var cfg = cfgOf(tbl);
      ensureDeclaredHeader(tbl, cfg);
      syncStoreFromDom(tbl, name);
      var st = store[name];
      captureTableSource(tbl, st);
      if (cfg.editMode === 'persistent' && !st.__persistLoaded) {
        var savedTable = loadPersistedTable(name);
        if (savedTable) applyPersistedTable(st, savedTable);
        st.__persistLoaded = true;
      }
      applyPersistedTableAttrs(tbl, st);
      var view = applyView(name);
      cfg = view.cfg;
      st = view.st;
      decorateTable(tbl, st, cfg);
      ensureFilterInput(tbl, name, st, cfg);
      ensurePager(tbl, name, st, cfg);
      ensureAddTableButton(tbl, name, st, cfg);
    }
  }
  var pop=null, popCell=null;
  function closePop(){ if(pop){ if(pop.parentNode) pop.parentNode.removeChild(pop); pop=null; popCell=null; } }
  function openPop(c, txt){
    if (popCell === c) return;
    closePop();
    pop=document.createElement('div'); pop.className='sui-cell-pop'; pop.textContent=txt;
    var cs=getComputedStyle(c);
    var bg=cs.getPropertyValue('--pop-bg').trim(); if(bg) pop.style.background=bg;
    var fg=cs.getPropertyValue('--pop-color').trim(); if(fg) pop.style.color=fg;
    var fs=cs.getPropertyValue('--pop-size').trim(); if(fs) pop.style.fontSize=fs;
    var mw=cs.getPropertyValue('--pop-maxw').trim(); if(mw) pop.style.maxWidth=mw;
    var br=cs.getPropertyValue('--pop-radius').trim(); if(br) pop.style.borderRadius=br;
    document.body.appendChild(pop);
    var r=c.getBoundingClientRect();
    var pr=pop.getBoundingClientRect();
    var left=Math.max(8, Math.min(window.innerWidth - pr.width - 8, r.left));
    var top=r.top - pr.height - 8;
    if (top < 8) top = Math.min(window.innerHeight - pr.height - 8, r.bottom + 6);
    pop.style.left=left+'px'; pop.style.top=top+'px';
    popCell=c;
  }
  document.addEventListener('mouseover', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-fixed] td, table[data-sui-fixed] th') : null;
    if(!c){ return; }
    if(document.activeElement === c) return; // being edited: text is already visible
    if(c.scrollWidth <= c.clientWidth + 1 && c.scrollHeight <= c.clientHeight + 1) { closePop(); return; }
    openPop(c, c.textContent);
  }, true);
  document.addEventListener('mouseout', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-fixed] td, table[data-sui-fixed] th') : null;
    if(c) closePop();
  }, true);
  document.addEventListener('keydown', function(ev){ if(ev.key==='Escape') closePop(); });

  // ---- draggable row/column boundaries (editable tables only) --------------
  var drag=null;
  function columnCells(cell){
    var tr=cell.parentNode, idx=Array.prototype.indexOf.call(tr.children, cell);
    var rows=cell.closest('table').querySelectorAll('tr'), out=[];
    for(var i=0;i<rows.length;i++){ var c=rows[i].children[idx]; if(c) out.push(c); }
    return out;
  }
  function edge(el, ev){
    var r=el.getBoundingClientRect();
    if ((r.right - ev.clientX) <= 7 && ev.clientX >= r.left + 10) return 'col';
    if ((r.bottom - ev.clientY) <= 7 && ev.clientY >= r.top + 8) return 'row';
    return '';
  }
  document.addEventListener('mousemove', function(ev){
    if(drag) return;
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-editable] td, table[data-sui-editable] th') : null;
    if(!c) return;
    var e=edge(c, ev);
    c.style.cursor = e === 'col' ? 'col-resize' : e === 'row' ? 'row-resize' : '';
  }, true);
  document.addEventListener('mousedown', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-editable] td, table[data-sui-editable] th') : null;
    if(!c) return;
    var e=edge(c, ev);
    if(!e) return;
    ev.preventDefault(); ev.stopPropagation();
    var r=c.getBoundingClientRect();
    drag = { kind:e, cell:c, start: e==='col' ? ev.clientX : ev.clientY, size: e==='col' ? r.width : r.height };
    document.body.style.cursor = e==='col' ? 'col-resize' : 'row-resize';
  }, true);
  document.addEventListener('mousemove', function(ev){
    if(!drag) return;
    ev.preventDefault();
    if(drag.kind==='col'){
      var w=Math.max(24, drag.size + (ev.clientX - drag.start));
      var cells=columnCells(drag.cell);
      for(var i=0;i<cells.length;i++){ cells[i].style.width=w+'px'; cells[i].style.minWidth=w+'px'; }
      updateWorkbookTableWidth(drag.cell.closest('table'));
    } else {
      var h=Math.max(18, drag.size + (ev.clientY - drag.start));
      var tr=drag.cell.parentNode;
      tr.style.height=h+'px';
      for(var j=0;j<tr.children.length;j++){ tr.children[j].style.height=h+'px'; }
    }
  }, true);
  document.addEventListener('mouseup', function(){
    if (!drag) return;
    var dragTable = drag.cell && drag.cell.closest ? drag.cell.closest('table') : null;
    var dragName = dragTable ? (dragTable.getAttribute('name') || '') : '';
    drag = null;
    document.body.style.cursor = '';
    if (dragName) {
      recordChange(dragName, { op: 'resize' });
      persistTable(dragName);
    }
  }, true);

  // ---- keep edited cells in the store so re-renders do not lose them -------
  // Header cells also contain sort/filter controls: the editable text is the
  // .sui-table-hlabel span, never the whole th (which includes the controls).
  function editTextOf(c){
    var lbl = c && c.querySelector ? c.querySelector('.sui-table-hlabel') : null;
    return lbl ? (lbl.textContent || '') : (c ? (c.textContent || '') : '');
  }
  function setCellEditText(c, txt){
    var lbl = c && c.querySelector ? c.querySelector('.sui-table-hlabel') : null;
    if (lbl) lbl.textContent = txt;
    else if (c) c.textContent = txt;
  }
  function focusInside(c){
    return !!(c && c.contains && document.activeElement && c.contains(document.activeElement));
  }
  // Keep the declarative label attribute in sync after a header edit so tools
  // that read data-sui-column-labels see the new text before the next render.
  function syncHeaderLabelsAttr(tbl){
    if (!tbl || !tbl.hasAttribute('data-sui-column-labels')) return;
    var cells = tbl.querySelectorAll('thead tr:first-child th, thead tr:first-child td');
    var labels = [];
    for (var i = 0; i < cells.length; i++) {
      labels.push(editTextOf(cells[i]));
    }
    var str = labels.join(',');
    if (tbl.getAttribute('data-sui-column-labels') !== str) tbl.setAttribute('data-sui-column-labels', str);
  }
  document.addEventListener('input', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    if (!c) return;
    var tbl = c.closest('table');
    if (!tbl) return;
    var name = tbl.getAttribute('name');
    if (!name || !store[name]) return;
    if (editCell !== c) {
      editCell = c;
      var committed = c.getAttribute('data-sui-raw');
      editOld = committed === null ? editTextOf(c) : committed;
    }
    var st = store[name];
    var tr = c.parentNode, col = Array.prototype.indexOf.call(tr.children, c);
    if (col < 0) return;
    var text = editTextOf(c);
    // Commit data-sui-raw on blur. Updating it on every keystroke would
    // spam the MutationObserver and make the table look like it is redrawing.
    if (tr.parentNode && tr.parentNode.tagName === 'THEAD') {
      if (!st.header) st.header = [];
      if (!st.headerEdits) st.headerEdits = {};
      st.header[col] = text;
      st.headerEdits[col] = text;
    } else {
      var tb = tr.parentNode, r = Array.prototype.indexOf.call(tb.rows, tr);
      if (r >= 0) {
        if (!st.allRows[r]) st.allRows[r] = [];
        st.allRows[r][col] = text;
        if (st.rows[r]) st.rows[r][col] = text;
      }
    }
    // The insertion phase ends after a pause even if the cell stays focused.
    clearEditTimer();
    editTimer = setTimeout(function(){
      editTimer = null;
      if (!c.isConnected) { if (editCell === c) { editCell = null; editOld = ''; } return; }
      commitCellEdit(c, tbl, name, cfgOf(tbl), false);
      if (editCell === c && (!c.isConnected || !focusInside(c))) { editCell = null; editOld = ''; }
    }, 900);
  }, true);
  var editCell = null, editOld = '', editTimer = null;
  function clearEditTimer(){ if (editTimer) { clearTimeout(editTimer); editTimer = null; } }
  // commitCellEdit stores the current cell text and fires the change events.
  // finalize=true (blur) also validates, restores invalid values and formats
  // the cell. finalize=false (idle timeout) commits valid text without moving
  // the caret or re-formatting the focused cell.
  function commitCellEdit(c, tbl, name, cfg, finalize){
    if (!c || !c.isConnected) {
      if (editCell === c) { editCell = null; editOld = ''; }
      return false;
    }
    var tr = c.parentNode, col = Array.prototype.indexOf.call(tr.children, c);
    if (col < 0) return false;
    var raw = (editTextOf(c) || '').trim();
    var oldVal = editOld === null || editOld === undefined ? raw : String(editOld);
    var rowIdx = tr.parentNode && tr.parentNode.tagName === 'THEAD' ? -1 : Array.prototype.indexOf.call(tr.parentNode.rows, tr);
    var type = cfg.types[col] || 'text';
    var err = rowIdx < 0 ? '' : validValue(raw, type, colOpts(cfg, col));
    var st = store[name];
    if (err) {
      if (!finalize) return false;
      setCellEditText(c, oldVal);
      c.classList.add('sui-table-invalid');
      c.setAttribute('aria-invalid', 'true');
      c.title = 'Invalid ' + err + ' value';
      if (st) {
        if (tr.parentNode && tr.parentNode.tagName === 'THEAD') {
          if (st.header) st.header[col] = oldVal;
        } else {
          var r0 = Array.prototype.indexOf.call(tr.parentNode.rows, tr);
          if (r0 >= 0) { if (!st.allRows[r0]) st.allRows[r0] = []; st.allRows[r0][col] = oldVal; if (st.rows[r0]) st.rows[r0][col] = oldVal; }
        }
      }
      return false;
    }
    c.classList.remove('sui-table-invalid');
    c.removeAttribute('aria-invalid');
    c.removeAttribute('title');
    if (raw !== oldVal) {
      if (st) {
        if (rowIdx < 0) {
          if (!st.header) st.header = [];
          if (!st.headerEdits) st.headerEdits = {};
          st.header[col] = raw;
          st.headerEdits[col] = raw;
        } else {
          if (!st.allRows[rowIdx]) st.allRows[rowIdx] = [];
          st.allRows[rowIdx][col] = raw;
          if (st.rows[rowIdx]) st.rows[rowIdx][col] = raw;
        }
      }
      c.setAttribute('data-sui-raw', raw);
      if (finalize && rowIdx < 0) {
        var labelEl = c.querySelector('.sui-table-hlabel');
        if (labelEl) {
          if (labelEl.textContent !== raw) labelEl.textContent = raw;
        } else if (c.textContent !== raw) {
          c.textContent = raw;
        }
      }
      if (finalize) {
        recordChange(name, { op: 'cell', row: rowIdx, column: col, old: oldVal, new: raw, value: raw });
        fireTableEvent(name, 'oncellchange');
        if (rowIdx >= 0) fireTableEvent(name, 'onrowchange');
        fireTableEvent(name, 'ontablechange');
        editOld = raw;
      }
    }
    if (finalize && rowIdx >= 0) {
      var fv = formatValue(raw, cfg.formats[col]);
      if (c.textContent !== fv) c.textContent = fv;
    }
    if (rowIdx < 0) syncHeaderLabelsAttr(tbl);
    return true;
  }
  document.addEventListener('focusin', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    var tbl = c && c.closest ? c.closest('table[data-sui-editable]') : null;
    if (!tbl) return;
    clearEditTimer();
    editCell = c;
    editOld = c.getAttribute('data-sui-raw');
    if (editOld === null) editOld = editTextOf(c);
    else if ((editTextOf(c) || '') !== editOld) setCellEditText(c, editOld);
    c.classList.add('sui-table-editing');
    c.classList.remove('sui-table-invalid');
    c.removeAttribute('aria-invalid');
    c.removeAttribute('title');
  }, true);
  document.addEventListener('focusout', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    if (!c || c !== editCell) return;
    clearEditTimer();
    var tbl = c.closest('table[data-sui-editable]');
    var name = tbl ? tbl.getAttribute('name') : '';
    if (!tbl || !name) { c.classList.remove('sui-table-editing'); editCell = null; editOld = ''; return; }
    commitCellEdit(c, tbl, name, cfgOf(tbl), true);
    c.classList.remove('sui-table-editing');
    editCell = null; editOld = '';
  }, true);

  if(window.MutationObserver){
    var pending=false;
    new MutationObserver(function(){
      if (editCell && !editCell.isConnected) { editCell = null; editOld = ''; }
      var ae = document.activeElement;
      if (editCell || (ae && ae.closest && ae.closest('td[contenteditable="true"], th[contenteditable="true"]'))) return;
      if(pending) return; pending=true; setTimeout(function(){
        pending=false;
        var a2 = document.activeElement;
        if (!editCell && !(a2 && a2.closest && a2.closest('td[contenteditable="true"], th[contenteditable="true"]'))) init();
      }, 60);
    }).observe(document.body,{childList:true,subtree:true,characterData:true});
  }
  init();
  return {
    display: display, data: data, header: header, json: json,
    select: select, slice: slice, clear: clear, importText: importText,
    exportText: exportText, applyAll: applyAll, afterRender: afterRender,
    sort: sort, filter: filter, filterText: filterText, page: page, setPageSize: setPageSize,
    selectedRows: selectedRows, selectedCells: selectedCells, selectedColumns: selectedColumns,
    lastChange: lastChange, clearPersisted: clearPersisted
  };
})();
window.__SUI_TBL = __SUI_TBL;

const __SUI_OVERLAY = (function(){
  var tracked = [];
  function clamp(v, lo, hi) { return Math.max(lo, Math.min(hi, v)); }
  function visible(panel) {
    if (!panel || panel.hidden) return false;
    try { return getComputedStyle(panel).display !== 'none'; } catch (e) { return true; }
  }
  function timeValue(v, fallback) {
    if (v == null || v === '') return fallback;
    var s = String(v).trim();
    if (/^[0-9]+$/.test(s)) return s + 'ms';
    if (/^[0-9]*\.?[0-9]+(ms|s)$/.test(s)) return s;
    return fallback;
  }
  function markOverlayChain(panel) {
    if (!panel || panel.__suiMarked) return;
    var marked = [];
    var el = panel.parentElement;
    while (el && el !== document.documentElement) {
      var openCount = Number(el.getAttribute('data-sui-overlay-count') || '0') + 1;
      el.setAttribute('data-sui-overlay-count', String(openCount));
      el.classList.add('sui-overlay-anchor-open');
      marked.push(el);
      el = el.parentElement;
    }
    panel.__suiMarked = marked;
  }
  function unmarkOverlayChain(panel) {
    if (!panel || !panel.__suiMarked) return;
    var marked = panel.__suiMarked;
    panel.__suiMarked = null;
    for (var i = 0; i < marked.length; i++) {
      var el = marked[i];
      var openCount = Math.max(0, Number(el.getAttribute('data-sui-overlay-count') || '0') - 1);
      if (openCount > 0) {
        el.setAttribute('data-sui-overlay-count', String(openCount));
      } else {
        el.removeAttribute('data-sui-overlay-count');
        el.classList.remove('sui-overlay-anchor-open');
      }
    }
  }
  function place(anchor, panel, opts) {
    if (!anchor || !panel) return '';
    // A transformed/filtered ancestor makes a fixed panel use that ancestor as
    // its containing block: the overlay would jump far from its trigger. Keep
    // every ancestor of the panel animation-free while the overlay is open.
    markOverlayChain(panel);
    opts = opts || {};
    var margin = opts.margin == null ? 8 : opts.margin;
    var gap = opts.gap == null ? 4 : opts.gap;
    var vw = window.innerWidth || document.documentElement.clientWidth || 0;
    var vh = window.innerHeight || document.documentElement.clientHeight || 0;
    if (!vw || !vh) return '';
    panel.style.position = 'fixed';
    panel.style.right = 'auto';
    panel.style.left = '0px';
    panel.style.top = '0px';
    panel.style.maxWidth = opts.preserveMaxWidth ? '' : Math.max(80, vw - margin * 2) + 'px';
    panel.style.maxHeight = '';
    panel.style.overflowY = '';
    var ar = anchor.getBoundingClientRect();
    if (opts.matchWidth) panel.style.width = Math.round(ar.width) + 'px';
    else panel.style.width = '';
    var pw = panel.offsetWidth || panel.getBoundingClientRect().width || 0;
    var ph = panel.offsetHeight || panel.getBoundingClientRect().height || 0;
    var fitsBottom = ar.bottom + gap + ph <= vh - margin;
    var fitsTop = ar.top - gap - ph >= margin;
    var fitsRight = ar.right + gap + pw <= vw - margin;
    var fitsLeft = ar.left - gap - pw >= margin;
    var preferred = String(opts.placement || 'auto').toLowerCase();
    var placed = preferred;
    if (preferred === 'auto') {
      if (fitsBottom) placed = 'bottom';
      else if (fitsTop) placed = 'top';
      else if (fitsRight) placed = 'right';
      else if (fitsLeft) placed = 'left';
      else placed = 'bottom';
    } else if (preferred === 'bottom' && !fitsBottom && fitsTop) {
      placed = 'top';
    } else if (preferred === 'top' && !fitsTop && fitsBottom) {
      placed = 'bottom';
    } else if (preferred === 'right' && !fitsRight && fitsLeft) {
      placed = 'left';
    } else if (preferred === 'left' && !fitsLeft && fitsRight) {
      placed = 'right';
    }
    var align = String(opts.align || 'start').toLowerCase();
    function alignLeft() {
      if (align === 'end') return ar.right - pw;
      if (align === 'center') return ar.left + (ar.width - pw) / 2;
      return ar.left;
    }
    function alignTop() {
      if (align === 'end') return ar.bottom - ph;
      if (align === 'center') return ar.top + (ar.height - ph) / 2;
      return ar.top;
    }
    var left = alignLeft();
    var top = ar.bottom + gap;
    if (placed === 'top') top = ar.top - gap - ph;
    else if (placed === 'right') { left = ar.right + gap; top = alignTop(); }
    else if (placed === 'left') { left = ar.left - gap - pw; top = alignTop(); }
    // Main menus and dropdowns normally align to the trigger's left edge. If
    // that would clip the panel on the right, align its right edge to the
    // trigger instead. The reverse applies to explicitly right-aligned menus.
    if (placed === 'bottom' || placed === 'top') {
      if (align === 'start' && left + pw > vw - margin) {
        var rightAligned = ar.right - pw;
        if (rightAligned >= margin) left = rightAligned;
      } else if (align === 'end' && left < margin) {
        var leftAligned = ar.left;
        if (leftAligned + pw <= vw - margin) left = leftAligned;
      }
    }
    if (placed === 'bottom') {
      var availBottom = vh - margin - top;
      if (ph > availBottom) {
        panel.style.maxHeight = Math.max(60, availBottom) + 'px';
        panel.style.overflowY = 'auto';
        ph = Math.min(ph, Math.max(60, availBottom));
      }
    } else if (placed === 'top') {
      var availTop = top - margin;
      if (ph > availTop) {
        panel.style.maxHeight = Math.max(60, availTop) + 'px';
        panel.style.overflowY = 'auto';
        ph = Math.min(ph, Math.max(60, availTop));
      }
    } else if (ph > vh - margin * 2) {
      panel.style.maxHeight = Math.max(60, vh - margin * 2) + 'px';
      panel.style.overflowY = 'auto';
      ph = Math.min(ph, vh - margin * 2);
    }
    left = clamp(left, margin, Math.max(margin, vw - pw - margin));
    top = clamp(top, margin, Math.max(margin, vh - ph - margin));
    panel.style.left = Math.round(left) + 'px';
    panel.style.top = Math.round(top) + 'px';
    panel.__suiOverlay = { anchor: anchor, opts: opts };
    if (tracked.indexOf(panel) < 0) tracked.push(panel);
    return placed;
  }
  function reset(panel) {
    if (!panel) return;
    unmarkOverlayChain(panel);
    panel.style.position = '';
    panel.style.left = '';
    panel.style.top = '';
    panel.style.right = '';
    panel.style.width = '';
    panel.style.maxHeight = '';
    panel.style.maxWidth = '';
    panel.style.overflowY = '';
    panel.removeAttribute('data-sui-anim');
    panel.removeAttribute('data-sui-anim-shift');
    panel.__suiOverlay = null;
    var i = tracked.indexOf(panel);
    if (i >= 0) tracked.splice(i, 1);
  }
  function animate(panel, effect, duration, easing, delay, placed) {
    if (!panel) return;
    var fx = String(effect || 'fade').toLowerCase();
    if (['fade', 'scale', 'slide', 'none'].indexOf(fx) < 0) fx = 'fade';
    if (fx === 'none') {
      panel.removeAttribute('data-sui-anim');
      panel.removeAttribute('data-sui-anim-shift');
      return;
    }
    panel.style.setProperty('--sui-anim-duration', timeValue(duration, '140ms'));
    panel.style.setProperty('--sui-anim-easing', String(easing || 'ease-out'));
    panel.style.setProperty('--sui-anim-delay', timeValue(delay, '0ms'));
    var shift = 'down';
    if (placed === 'top') shift = 'up';
    else if (placed === 'right') shift = 'right';
    else if (placed === 'left') shift = 'left';
    panel.removeAttribute('data-sui-anim');
    panel.removeAttribute('data-sui-anim-shift');
    void panel.offsetWidth;
    panel.setAttribute('data-sui-anim', fx);
    panel.setAttribute('data-sui-anim-shift', shift);
  }
  function reposition() {
    for (var i = 0; i < tracked.length; i++) {
      var p = tracked[i];
      if (!document.body.contains(p)) { reset(p); tracked.splice(i, 1); i--; continue; }
      if (!p.__suiOverlay || !p.__suiOverlay.anchor || !p.__suiOverlay.anchor.isConnected) { reset(p); tracked.splice(i, 1); i--; continue; }
      if (!visible(p) || !p.__suiOverlay) continue;
      place(p.__suiOverlay.anchor, p, p.__suiOverlay.opts);
    }
  }
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', reposition);
    document.addEventListener('scroll', reposition, true);
  }
  return { place: place, reset: reset, animate: animate, reposition: reposition };
})();

(function(){
  var store = Object.create(null);
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function tabButtons(box) {
    var list = box ? box.querySelector('.sui-tablist') : null;
    if (!list) return [];
    return Array.prototype.filter.call(list.children, function(el) {
      return el.classList && el.classList.contains('sui-tab');
    });
  }
  function tabPanels(box) {
    var list = box ? box.querySelector('.sui-tabpanels') : null;
    if (!list) return [];
    return Array.prototype.filter.call(list.children, function(el) {
      return el.classList && el.classList.contains('sui-tabpanel');
    });
  }
  function activate(box, index, persist) {
    var tabs = tabButtons(box);
    var panels = tabPanels(box);
    if (!tabs.length) return;
    if (index < 0) index = tabs.length - 1;
    if (index >= tabs.length) index = 0;
    for (var i = 0; i < tabs.length; i++) {
      var on = i === index;
      tabs[i].classList.toggle('active', on);
      tabs[i].setAttribute('aria-selected', on ? 'true' : 'false');
      tabs[i].tabIndex = on ? 0 : -1;
      if (panels[i]) {
        panels[i].hidden = !on;
        panels[i].classList.toggle('active', on);
      }
    }
    var key = box.getAttribute('data-sui-tabs-key');
    if (key && persist) store[key] = index;
  }
  function applyStored() {
    var boxes = document.querySelectorAll('.sui-tabs');
    for (var i = 0; i < boxes.length; i++) {
      var key = boxes[i].getAttribute('data-sui-tabs-key');
      if (key && store[key] !== undefined) activate(boxes[i], store[key], false);
    }
  }
  document.addEventListener('click', function(ev) {
    var tab = closest(ev.target, '.sui-tab');
    if (!tab) return;
    var box = closest(tab, '.sui-tabs');
    if (!box) return;
    var tabs = tabButtons(box);
    activate(box, tabs.indexOf(tab), true);
  });
  document.addEventListener('keydown', function(ev) {
    var tab = closest(ev.target, '.sui-tab');
    if (!tab) return;
    var box = closest(tab, '.sui-tabs');
    if (!box) return;
    var tabs = tabButtons(box);
    var index = tabs.indexOf(tab);
    var next = -1;
    if (ev.key === 'ArrowRight') next = index + 1;
    else if (ev.key === 'ArrowLeft') next = index - 1;
    else if (ev.key === 'Home') next = 0;
    else if (ev.key === 'End') next = tabs.length - 1;
    if (next < 0 && ev.key === 'ArrowLeft') next = tabs.length - 1;
    if (next >= tabs.length) next = 0;
    if (next < 0) return;
    ev.preventDefault();
    activate(box, next, true);
    tabs[next].focus();
  });
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      // Do not re-activate tabs while a table cell is being edited: toggling
      // panel visibility during typing can steal focus or make the table flicker.
      var ae = document.activeElement;
      if (ae && ae.closest && ae.closest('td[contenteditable="true"], th[contenteditable="true"]')) return;
      if (pending) return;
      pending = true;
      // Restore before the browser paints the regenerated view. A setTimeout
      // here made the first sheet visible for a frame before the stored tab
      // was re-activated (visible flicker when editing a later sheet).
      var run = function() { pending = false; applyStored(); };
      if (typeof Promise !== 'undefined') Promise.resolve().then(run);
      else run();
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
  window.__SUI_TABS = { restore: applyStored };
  applyStored();
})();

(function(){
  var store = Object.create(null);
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function items(box) {
    if (!box) return [];
    return Array.prototype.filter.call(box.children, function(el) {
      return el.classList && el.classList.contains('sui-accordion-item');
    });
  }
  function setOpen(box, index, persist) {
    var list = items(box);
    for (var i = 0; i < list.length; i++) {
      var open = i === index;
      list[i].classList.toggle('open', open);
      var head = list[i].querySelector('.sui-accordion-head');
      var panel = list[i].querySelector('.sui-accordion-panel');
      if (head) head.setAttribute('aria-expanded', open ? 'true' : 'false');
      if (panel) panel.hidden = !open;
    }
    var key = box.getAttribute('data-sui-accordion-key');
    if (key && persist) store[key] = index;
  }
  function applyStored() {
    var boxes = document.querySelectorAll('.sui-accordion');
    for (var i = 0; i < boxes.length; i++) {
      var key = boxes[i].getAttribute('data-sui-accordion-key');
      if (key && store[key] !== undefined) setOpen(boxes[i], store[key], false);
    }
  }
  document.addEventListener('click', function(ev) {
    var head = closest(ev.target, '.sui-accordion-head');
    if (!head) return;
    var item = closest(head, '.sui-accordion-item');
    var box = closest(head, '.sui-accordion');
    if (!item || !box) return;
    var list = items(box);
    var index = list.indexOf(item);
    if (index < 0) return;
    setOpen(box, item.classList.contains('open') ? -1 : index, true);
  });
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      setTimeout(function() { pending = false; applyStored(); }, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
  applyStored();
})();

(function(){
  var MAX = 4;
  var DEFAULT_MS = 4000;
  var KINDS = { info:1, success:1, warning:1, error:1 };
  var ICONS = { info:'i', success:'\u2713', warning:'!', error:'\u00d7' };

  function findHost() {
    var host = document.querySelector('[data-sui-toast-host]');
    if (!host) host = document.querySelector('.sui-toast-host');
    if (host) return host;
    host = document.createElement('div');
    host.className = 'sui-toast-host sui-toast-host-fixed';
    host.setAttribute('data-sui-toast-host', '1');
    document.body.appendChild(host);
    return host;
  }
  function findHostByIndex(index) {
    var hosts = document.querySelectorAll('[data-sui-toast-host]');
    var i = parseInt(index, 10);
    return i >= 0 && i < hosts.length ? hosts[i] : null;
  }
  function syncHost(host) {
    if (!host) return;
    var x = host.getAttribute('data-sui-toast-x');
    var y = host.getAttribute('data-sui-toast-y');
    if (x) host.style.setProperty('--sui-toast-offset-x', x);
    if (y) host.style.setProperty('--sui-toast-offset-y', y);
  }
  function normalizeKind(kind) {
    kind = String(kind === null || kind === undefined ? '' : kind).toLowerCase();
    return KINDS[kind] ? kind : 'info';
  }
  function readInt(host, attr, fallback) {
    var n = parseInt(host ? host.getAttribute(attr) : '', 10);
    return isNaN(n) ? fallback : n;
  }
  function removeToast(el) {
    if (!el) return;
    if (el.parentNode) el.parentNode.removeChild(el);
  }
  function clear() {
    var hosts = document.querySelectorAll('[data-sui-toast-host], .sui-toast-host');
    for (var i = 0; i < hosts.length; i++) {
      while (hosts[i].firstChild) hosts[i].removeChild(hosts[i].firstChild);
    }
  }
  function showOn(host, message, kind, duration) {
    if (!host) host = findHost();
    syncHost(host);
    if (!host.hasAttribute('aria-live')) host.setAttribute('aria-live', 'polite');
    host.setAttribute('aria-atomic', 'false');
    var k = normalizeKind(kind);
    var el = document.createElement('div');
    el.className = 'sui-toast sui-toast-' + k;
    el.setAttribute('role', 'status');
    var icons = host.getAttribute('data-sui-toast-icons') !== 'false';
    var icon = null;
    if (icons) {
      icon = document.createElement('span');
      icon.className = 'sui-toast-icon';
      icon.setAttribute('aria-hidden', 'true');
      icon.textContent = ICONS[k] || ICONS.info;
    }
    var text = document.createElement('span');
    text.className = 'sui-toast-message';
    text.textContent = String(message === null || message === undefined ? '' : message);
    var close = document.createElement('button');
    close.type = 'button';
    close.className = 'sui-toast-close';
    close.setAttribute('aria-label', 'Dismiss notification');
    close.textContent = '\u00d7';
    close.addEventListener('click', function() { removeToast(el); });
    if (icon) el.appendChild(icon);
    el.appendChild(text);
    el.appendChild(close);
    var max = readInt(host, 'data-sui-toast-max', MAX);
    if (max < 1) max = 1;
    if (max > 8) max = 8;
    while (host.children.length >= max) removeToast(host.firstChild);
    host.appendChild(el);
    var hostDefault = readInt(host, 'data-sui-toast-duration', DEFAULT_MS);
    var ms = duration === undefined || duration === null ? hostDefault : Number(duration);
    if (isNaN(ms)) ms = hostDefault;
    if (ms > 0) {
      setTimeout(function() { removeToast(el); }, ms);
    }
    return el;
  }
  function show(message, kind, duration) {
    return showOn(findHost(), message, kind, duration);
  }
  var initialHosts = document.querySelectorAll('[data-sui-toast-host]');
  for (var hi = 0; hi < initialHosts.length; hi++) syncHost(initialHosts[hi]);
  window.__sui_toast_show = show;
  window.__sui_toast_show_at = function(index, message, kind, duration) {
    return showOn(findHostByIndex(index) || findHost(), message, kind, duration);
  };
  window.__sui_toast_clear = clear;
  // Close buttons inside a re-rendered host still work through delegation.
  document.addEventListener('click', function(ev) {
    var close = ev.target && ev.target.closest ? ev.target.closest('.sui-toast-close') : null;
    if (!close) return;
    var toast = close.closest ? close.closest('.sui-toast') : null;
    if (toast) removeToast(toast);
  });
})();

(function(){
  var bubble = null;
  var current = null;
  var timer = null;
  var seq = 0;

  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function rootValue(name, fallback) {
    var v = '';
    try { v = getComputedStyle(document.documentElement).getPropertyValue(name).trim(); } catch (e) {}
    if (!v) {
      try { v = getComputedStyle(document.body).getPropertyValue(name).trim(); } catch (e) {}
    }
    return v || fallback;
  }
  function delayFor() {
    var v = parseInt(rootValue('--sui-tooltip-delay', '140'), 10);
    return isNaN(v) || v < 0 ? 140 : v;
  }
  function placeFor() {
    var v = rootValue('--sui-tooltip-place', 'auto').toLowerCase();
    return ['auto', 'top', 'right', 'bottom', 'left'].indexOf(v) >= 0 ? v : 'auto';
  }
  function ensureBubble() {
    if (bubble) return bubble;
    bubble = document.createElement('div');
    bubble.className = 'sui-tooltip-bubble';
    bubble.setAttribute('role', 'tooltip');
    bubble.hidden = true;
    document.body.appendChild(bubble);
    return bubble;
  }
  function position(el, b) {
    var place = placeFor();
    b.style.visibility = 'hidden';
    var placed = __SUI_OVERLAY.place(el, b, {
      placement: place,
      align: 'center',
      gap: 8,
      margin: 8,
      preserveMaxWidth: true
    });
    b.setAttribute('data-sui-tooltip-placed', placed || '');
    b.style.visibility = '';
    return placed;
  }
  function hide(el) {
    if (!el) return;
    if (timer) { clearTimeout(timer); timer = null; }
    if (current === el || !el) {
      current = null;
      if (bubble) {
        bubble.hidden = true;
        bubble.classList.remove('open');
        __SUI_OVERLAY.reset(bubble);
      }
    }
    if (el.getAttribute('aria-describedby') && bubble && el.getAttribute('aria-describedby') === bubble.id) {
      el.removeAttribute('aria-describedby');
    }
  }
  function show(el) {
    if (!el) return;
    var b = ensureBubble();
    if (!b.id) b.id = 'sui-tooltip-' + (++seq);
    b.textContent = el.getAttribute('data-sui-tooltip') || '';
    b.hidden = false;
    b.classList.add('open');
    current = el;
    el.setAttribute('aria-describedby', b.id);
    var placed = position(el, b);
    __SUI_OVERLAY.animate(
      b,
      rootValue('--sui-tooltip-animation', 'fade'),
      rootValue('--sui-tooltip-duration', '140ms'),
      rootValue('--sui-tooltip-easing', 'ease-out'),
      rootValue('--sui-tooltip-anim-delay', '0ms'),
      placed || 'top'
    );
  }
  function schedule(el, immediate) {
    if (!el) return;
    if (timer) { clearTimeout(timer); timer = null; }
    var delay = delayFor();
    if (immediate) delay = Math.min(delay, 80);
    timer = setTimeout(function() {
      timer = null;
      show(el);
    }, delay);
  }
  function targetFrom(ev) {
    return closest(ev.target, '[data-sui-tooltip]');
  }
  function inside(el, other) {
    return !!(el && other && (el === other || (el.contains && el.contains(other))));
  }

  document.addEventListener('pointerover', function(ev) {
    var el = targetFrom(ev);
    if (!el || inside(el, ev.relatedTarget)) return;
    el.__suiTooltipHover = true;
    schedule(el, false);
  }, true);
  document.addEventListener('pointerout', function(ev) {
    var el = targetFrom(ev);
    if (!el || inside(el, ev.relatedTarget)) return;
    el.__suiTooltipHover = false;
    if (!el.__suiTooltipFocus) hide(el);
  }, true);
  document.addEventListener('focusin', function(ev) {
    var el = targetFrom(ev);
    if (!el) return;
    el.__suiTooltipFocus = true;
    schedule(el, true);
  }, true);
  document.addEventListener('focusout', function(ev) {
    var el = targetFrom(ev);
    if (!el || inside(el, ev.relatedTarget)) return;
    el.__suiTooltipFocus = false;
    if (!el.__suiTooltipHover) hide(el);
  }, true);
  document.addEventListener('keydown', function(ev) {
    if (ev.key !== 'Escape' || !current) return;
    current.__suiTooltipHover = false;
    current.__suiTooltipFocus = false;
    hide(current);
  });
})();

(function(){
  var seq = 0;
  var MAX_OPTIONS = 2000;
  var MAX_QUERY = 256;
  var MAX_OPTION_LIMIT = 10000;
  var MAX_OPTION_TEXT = 10000;
  var MAX_OPTIONS_TEXT = 500000;
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  // parseOptions returns {list, error}. It never throws and never returns
  // partially-parsed data: any malformed escape, empty label/value or control
  // character makes the whole list invalid so stale options are not shown.
  function parseOptions(raw) {
    var text = String(raw == null ? '' : raw);
    if (text.trim() === '') return { list: [], error: '' };
    if (text.length > MAX_OPTIONS_TEXT) return { list: [], error: 'Invalid options data: list too long' };
    var out = [];
    var label = '';
    var value = '';
    var hasValue = false;
    var escaped = false;
    var error = '';
    function addOption() {
      var l = label.trim();
      var v = hasValue ? value.trim() : l;
      if (!l) { error = 'Invalid options data: empty option label'; return false; }
      if (hasValue && !v) { error = 'Invalid options data: empty option value'; return false; }
      if (/[\u0000-\u001f]/.test(l) || /[\u0000-\u001f]/.test(v)) {
        error = 'Invalid options data: control character'; return false;
      }
      if (l.length > MAX_OPTION_TEXT || v.length > MAX_OPTION_TEXT) {
        error = 'Invalid options data: option too long'; return false;
      }
      out.push({ label: l, value: v });
      if (out.length > MAX_OPTIONS) {
        error = 'Invalid options data: too many options'; return false;
      }
      return true;
    }
    for (var i = 0; i < text.length; i++) {
      var ch = text.charAt(i);
      if (escaped) {
        if (hasValue) value += ch; else label += ch;
        escaped = false;
        continue;
      }
      if (ch === '\\') { escaped = true; continue; }
      if (ch === ',') {
        if (!addOption()) return { list: [], error: error };
        label = ''; value = ''; hasValue = false;
        continue;
      }
      if (ch === '=' && !hasValue) { hasValue = true; continue; }
      if (hasValue) value += ch; else label += ch;
    }
    if (escaped) return { list: [], error: 'Invalid options data: dangling escape' };
    if (!addOption()) return { list: [], error: error };
    return { list: out, error: '' };
  }
  // hasGlob checks for an unescaped * or ? in a search pattern.
  function hasGlob(pattern) {
    var escaped = false;
    for (var i = 0; i < pattern.length; i++) {
      var ch = pattern.charAt(i);
      if (escaped) { escaped = false; continue; }
      if (ch === '\\') { escaped = true; continue; }
      if (ch === '*' || ch === '?') return true;
    }
    return false;
  }
  function literalPattern(pattern) {
    return pattern.replace(/\\([\s\S])/g, '$1');
  }
  // globMatch is an iterative glob matcher. It supports * (any sequence), ?
  // (one character) and backslash escapes. No user input is compiled as a
  // regular expression, avoiding regex injection and catastrophic backtracking.
  function globMatch(text, pattern) {
    var ti = 0, pi = 0, star = -1, mark = 0;
    while (ti < text.length) {
      var current = pi < pattern.length ? pattern.charAt(pi) : '';
      var literal = null;
      if (current === '\\') {
        if (pi + 1 < pattern.length) {
          literal = pattern.charAt(pi + 1);
          pi += 2;
        } else {
          literal = '\\';
          pi += 1;
        }
      } else if (current === '?') {
        ti++; pi++; continue;
      } else if (current === '*') {
        star = pi; mark = ti; pi++; continue;
      } else if (current !== '') {
        literal = current;
        pi++;
      }
      if (literal !== null && ti < text.length && text.charAt(ti) === literal) {
        ti++;
        continue;
      }
      if (star >= 0) { pi = star + 1; ti = ++mark; }
      else return false;
    }
    while (pi < pattern.length && pattern.charAt(pi) === '*') pi++;
    return pi === pattern.length;
  }
  function matches(label, query) {
    var q = String(query || '').trim();
    if (!q) return true;
    var text = String(label || '').toLowerCase();
    q = q.toLowerCase().slice(0, MAX_QUERY);
    if (hasGlob(q)) return globMatch(text, q);
    return text.indexOf(literalPattern(q)) >= 0;
  }
  function init(root) {
    if (!root || root.__suiCombobox) return;
    root.__suiCombobox = true;
    var input = root.querySelector('.sui-combobox-input');
    var hidden = root.querySelector('.sui-combobox-value');
    var popup = root.querySelector('.sui-combobox-popup');
    var label = root.querySelector('.sui-combobox-label');
    var toggle = root.querySelector('.sui-combobox-toggle');
    var clear = root.querySelector('.sui-combobox-clear');
    var errorEl = root.querySelector('.sui-combobox-error');
    if (!input || !hidden || !popup) return;
    var parsed = parseOptions(root.getAttribute('data-sui-options'));
    var options = [];
    var configError = parsed.error;
    var emptyText = root.getAttribute('data-sui-empty') || 'No results';
    var setter = root.getAttribute('data-sui-combobox-setter') || '';
    var handler = root.getAttribute('data-sui-combobox-onchange') || '';
    var disabled = root.getAttribute('data-sui-disabled') === '1';
    var required = root.getAttribute('data-sui-required') === '1';
    var missingRequired = false;
    var maxRaw = String(root.getAttribute('data-sui-max-option-length') || '').trim();
    var maxLength = 0;
    if (maxRaw !== '') {
      if (!/^[0-9]+$/.test(maxRaw) || parseInt(maxRaw, 10) > MAX_OPTION_LIMIT) {
        configError = 'Invalid max option length';
      } else {
        maxLength = parseInt(maxRaw, 10);
      }
    }
    var maxCountRaw = String(root.getAttribute('data-sui-max-options') || '').trim();
    var maxCount = 0;
    if (maxCountRaw !== '') {
      if (!/^[0-9]+$/.test(maxCountRaw) || parseInt(maxCountRaw, 10) > MAX_OPTIONS) {
        configError = 'Invalid max options';
      } else {
        maxCount = parseInt(maxCountRaw, 10);
      }
    }
    // Validate the complete candidate list before publishing it. A single
    // overlong item (label or stored value) or an excessive option count
    // rejects the whole list, so no partial data reaches the DOM.
    if (!configError) {
      for (var vi = 0; vi < parsed.list.length; vi++) {
        var candidate = parsed.list[vi];
        if (maxLength > 0 && (candidate.label.length > maxLength || candidate.value.length > maxLength)) {
          configError = 'Invalid options data: option ' + (vi + 1) + ' exceeds max option length';
          break;
        }
      }
      if (!configError && maxCount > 0 && parsed.list.length > maxCount) {
        configError = 'Invalid options data: too many options (max ' + maxCount + ')';
      }
      if (!configError) options = parsed.list;
    }
    var filtered = options.slice();
    var active = -1;
    var current = hidden.value || '';
    if (!input.id) input.id = 'sui-combobox-input-' + (++seq);
    if (!popup.id) popup.id = 'sui-combobox-list-' + (++seq);
    if (!label.id) label.id = 'sui-combobox-label-' + (++seq);
    label.setAttribute('for', input.id);
    input.setAttribute('aria-controls', popup.id);
    input.setAttribute('aria-autocomplete', 'list');
    input.setAttribute('aria-haspopup', 'listbox');
    if (required) input.setAttribute('aria-required', 'true');
    if (disabled) {
      input.disabled = true;
      toggle.disabled = true;
      if (clear) clear.disabled = true;
    }
    function labelOf(value) {
      for (var i = 0; i < options.length; i++) if (options[i].value === value) return options[i].label;
      return '';
    }
    function renderError() {
      var message = configError;
      if (!message && required && !disabled && missingRequired && !current) {
        message = 'This field is required';
      }
      if (errorEl) {
        errorEl.textContent = message;
        errorEl.hidden = !message;
      }
      if (message) input.setAttribute('aria-invalid', 'true');
      else input.removeAttribute('aria-invalid');
      if (toggle) toggle.disabled = disabled || !!configError;
      if (configError) {
        popup.hidden = true;
        input.setAttribute('aria-expanded', 'false');
      }
    }
    function setError(message) {
      configError = message || '';
      renderError();
    }
    function syncInput() {
      input.value = labelOf(current);
      if (clear) clear.hidden = !current;
    }
    function clearMissingSelection() {
      if (configError || !current) return;
      for (var i = 0; i < options.length; i++) {
        if (options[i].value === current) return;
      }
      // The selected value is no longer present in the replacement list.
      // Clear the visible value and the bound state; do not run onChange,
      // because this is an automatic invalidation rather than a user action.
      current = '';
      hidden.value = '';
      input.value = '';
      if (clear) clear.hidden = true;
      if (setter && typeof window[setter] === 'function') {
        setTimeout(function() {
          if (typeof window[setter] === 'function') window[setter]('');
        }, 0);
      }
    }
    function close() {
      popup.hidden = true;
      __SUI_OVERLAY.reset(popup);
      input.setAttribute('aria-expanded', 'false');
      if (toggle) toggle.setAttribute('aria-expanded', 'false');
      input.removeAttribute('aria-activedescendant');
      active = -1;
    }
    function positionPopup(withAnim) {
      if (popup.hidden) return;
      var placed = __SUI_OVERLAY.place(root, popup, {
        placement: root.getAttribute('data-sui-combobox-placement') || 'auto',
        align: 'start',
        gap: 4,
        margin: 8,
        matchWidth: true
      });
      if (withAnim) {
        __SUI_OVERLAY.animate(
          popup,
          root.getAttribute('data-sui-animation') || 'fade',
          root.getAttribute('data-sui-duration'),
          root.getAttribute('data-sui-easing'),
          root.getAttribute('data-sui-delay'),
          placed
        );
      }
    }
    function open() {
      if (disabled || configError) return;
      popup.hidden = false;
      input.setAttribute('aria-expanded', 'true');
      if (toggle) toggle.setAttribute('aria-expanded', 'true');
      render('');
      positionPopup(true);
    }
    function render(query) {
      if (configError) return;
      var q = String(query || '').trim();
      filtered = options.filter(function(o) { return matches(o.label, q); });
      popup.textContent = '';
      if (!filtered.length) {
        var empty = document.createElement('div');
        empty.className = 'sui-combobox-empty';
        empty.textContent = emptyText;
        popup.appendChild(empty);
        active = -1;
        input.removeAttribute('aria-activedescendant');
        positionPopup(false);
        return;
      }
      filtered.forEach(function(o, i) {
        var row = document.createElement('div');
        row.className = 'sui-combobox-option';
        row.id = popup.id + '-option-' + i;
        row.setAttribute('role', 'option');
        row.setAttribute('data-value', o.value);
        row.setAttribute('aria-selected', o.value === current ? 'true' : 'false');
        row.textContent = o.label;
        row.addEventListener('mousedown', function(ev) { ev.preventDefault(); });
        row.addEventListener('click', function() { commit(o.value, o.label); });
        popup.appendChild(row);
      });
      active = -1;
      for (var i = 0; i < filtered.length; i++) {
        if (filtered[i].value === current) { active = i; break; }
      }
      if (active < 0) active = 0;
      highlight();
      positionPopup(false);
    }
    function highlight() {
      var rows = popup.querySelectorAll('.sui-combobox-option');
      for (var i = 0; i < rows.length; i++) {
        var on = i === active;
        rows[i].classList.toggle('active', on);
        rows[i].setAttribute('aria-selected', filtered[i] && filtered[i].value === current ? 'true' : 'false');
      }
      if (active >= 0 && rows[active]) input.setAttribute('aria-activedescendant', rows[active].id);
      else input.removeAttribute('aria-activedescendant');
    }
    function move(delta) {
      if (popup.hidden) open();
      if (!filtered.length) return;
      active = (active + delta + filtered.length) % filtered.length;
      highlight();
      var rows = popup.querySelectorAll('.sui-combobox-option');
      if (rows[active] && rows[active].scrollIntoView) rows[active].scrollIntoView({ block: 'nearest' });
    }
    function commit(value, label) {
      current = value;
      hidden.value = value;
      input.value = label || labelOf(value);
      if (clear) clear.hidden = !value;
      missingRequired = false;
      renderError();
      close();
      if (setter && typeof window[setter] === 'function') {
        window[setter](value);
      } else if (handler && typeof window[handler] === 'function') {
        window[handler]();
      }
    }
    input.addEventListener('focus', function() {
      missingRequired = false;
      renderError();
      if (disabled || configError) return;
      input.select();
      open();
    });
    input.addEventListener('click', function() {
      if (!disabled && !configError && popup.hidden) open();
    });
    input.addEventListener('input', function() {
      missingRequired = false;
      renderError();
      if (configError) return;
      open();
      render(input.value);
    });
    input.addEventListener('keydown', function(ev) {
      if (configError) return;
      if (ev.key === 'ArrowDown') {
        ev.preventDefault();
        if (popup.hidden) open(); else move(1);
      } else if (ev.key === 'ArrowUp') {
        ev.preventDefault();
        if (popup.hidden) open(); else move(-1);
      } else if (ev.key === 'Enter') {
        if (!popup.hidden && active >= 0 && filtered[active]) {
          ev.preventDefault();
          commit(filtered[active].value, filtered[active].label);
        }
      } else if (ev.key === 'Escape') {
        if (!popup.hidden) {
          ev.preventDefault();
          close();
          syncInput();
        }
      } else if (ev.key === 'Tab') {
        close();
      }
    });
    if (toggle) toggle.addEventListener('click', function() {
      if (disabled || configError) return;
      if (popup.hidden) { input.focus(); open(); } else close();
    });
    if (clear) clear.addEventListener('click', function(ev) {
      ev.preventDefault();
      if (configError) return;
      commit('', '');
      input.focus();
    });
    document.addEventListener('pointerdown', function(ev) {
      if (!root.contains(ev.target)) close();
    });
    root.addEventListener('focusout', function(ev) {
      if (!root.contains(ev.relatedTarget)) {
        setTimeout(close, 100);
        if (required && !disabled && !current) {
          missingRequired = true;
          renderError();
        }
      }
    });
    setError(configError);
    clearMissingSelection();
    syncInput();
  }
  function initAll() {
    var all = document.querySelectorAll('.sui-combobox');
    for (var i = 0; i < all.length; i++) init(all[i]);
  }
  initAll();
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      setTimeout(function() { pending = false; initAll(); }, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
})();

(function(){
  var openName = '';
  var lastFocus = null;
  var lastFocusID = '';
  var lastClick = null;
  var lastClickID = '';
  var ID_ATTR = 'data-' + 'sui-id';
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function findModal(name) {
    var all = document.querySelectorAll('[data-sui-modal]');
    for (var i = 0; i < all.length; i++) {
      if (all[i].getAttribute('data-sui-modal') === name) return all[i];
    }
    return null;
  }
  function focusFirst(modal) {
    if (!modal || !modal.classList.contains('open')) return;
    var panel = modal.querySelector('.sui-modal-panel');
    if (!panel) return;
    var first = panel.querySelector('[autofocus], button, input, select, textarea, [tabindex]');
    if (first && first.focus) {
      try { first.focus({ preventScroll: true }); } catch (e) { try { first.focus(); } catch (e2) {} }
    } else if (panel.focus) {
      try { panel.focus({ preventScroll: true }); } catch (e) { try { panel.focus(); } catch (e2) {} }
    }
  }
  function setOpen(modal, on) {
    if (!modal) return;
    if (on) {
      modal.hidden = false;
      modal.classList.add('open');
      modal.setAttribute('aria-hidden', 'false');
      // Compensate for the scrollbar that the modal-open class hides. Without
      // this the page width changes and a static chart can visibly jump.
      if (!document.body.classList.contains('sui-modal-open')) {
        var scrollbarWidth = Math.max(0, window.innerWidth - document.documentElement.clientWidth);
        var bodyPadding = parseFloat(window.getComputedStyle(document.body).paddingRight) || 0;
        document.body.__suiModalPaddingRight = document.body.style.paddingRight;
        if (scrollbarWidth > 0) document.body.style.paddingRight = (bodyPadding + scrollbarWidth) + 'px';
      }
      document.body.classList.add('sui-modal-open');
    } else {
      modal.classList.remove('open');
      modal.hidden = true;
      modal.setAttribute('aria-hidden', 'true');
      var any = document.querySelector('.sui-modal.open');
      if (!any) {
        document.body.classList.remove('sui-modal-open');
        if (document.body.__suiModalPaddingRight !== undefined) {
          document.body.style.paddingRight = document.body.__suiModalPaddingRight;
          delete document.body.__suiModalPaddingRight;
        }
      }
    }
  }
  function focusTarget() {
    var target = null;
    if (lastFocusID) {
      try { target = document.querySelector('[' + ID_ATTR + '="' + lastFocusID + '"]'); } catch (e) {}
    }
    if (!target) target = lastFocus;
    if (target && target.focus) {
      try { target.focus({ preventScroll: true }); } catch (e) { try { target.focus(); } catch (e2) {} }
    }
  }
  function restoreFocus() {
    focusTarget();
    // Keep the id through the current event turn: an action handler may call
    // modal.close() and the delegated dismiss listener then runs too. Focus a
    // second time after the click/focus default action has settled.
    var oldFocus = lastFocus;
    var oldID = lastFocusID;
    setTimeout(function() { if (!document.querySelector('.sui-modal.open')) focusTarget(); }, 0);
    setTimeout(function() {
      if (lastFocus === oldFocus && lastFocusID === oldID) {
        lastFocus = null;
        lastFocusID = '';
      }
    }, 40);
  }
  function closeByName(name, restore) {
    var modal = findModal(name);
    if (!modal) return;
    setOpen(modal, false);
    if (openName === name) openName = '';
    if (!document.querySelector('.sui-modal.open')) {
      if (restore) restoreFocus(); else { lastFocus = null; lastFocusID = ''; }
    }
  }
  function openByName(name) {
    name = String(name === null || name === undefined ? '' : name);
    if (!name) return;
    var modal = findModal(name);
    if (!modal) return;
    var current = openName ? findModal(openName) : null;
    if (current && current !== modal) setOpen(current, false);
    lastFocus = document.activeElement;
    lastFocusID = '';
    var holder = closest(lastFocus, '[' + ID_ATTR + ']');
    if (holder) lastFocusID = holder.getAttribute(ID_ATTR) || '';
    // The click capture below remembers the actual trigger even when the
    // browser moves focus to the body before running the handler.
    if (lastClickID) { lastFocus = lastClick; lastFocusID = lastClickID; }
    openName = name;
    setOpen(modal, true);
    setTimeout(function() { focusFirst(modal); }, 0);
  }
  function closeAll(restore) {
    var open = document.querySelectorAll('.sui-modal.open');
    for (var i = 0; i < open.length; i++) setOpen(open[i], false);
    openName = '';
    if (restore) restoreFocus(); else { lastFocus = null; lastFocusID = ''; }
  }
  function applyOpen() {
    if (!openName) return;
    var modal = findModal(openName);
    if (modal && !modal.classList.contains('open')) setOpen(modal, true);
  }
  window.__SUI_MODAL_OPEN = openByName;
  window.__SUI_MODAL_CLOSE = closeByName;
  window.__sui_modal_open = openByName;
  window.__sui_modal_close = closeByName;
  document.addEventListener('click', function(ev) {
    // Capture the element that triggered a modal.open() call. Programmatic
    // clicks in embedded previews do not always focus the button first.
    if (document.querySelector('.sui-modal.open')) return;
    var trigger = closest(ev.target, '[' + ID_ATTR + ']');
    if (trigger) {
      lastClick = trigger;
      lastClickID = trigger.getAttribute(ID_ATTR) || '';
    }
  }, true);
  document.addEventListener('click', function(ev) {
    var backdrop = closest(ev.target, '[data-sui-modal-backdrop]');
    if (backdrop) {
      var modal = closest(backdrop, '.sui-modal');
      if (modal) { closeByName(modal.getAttribute('data-sui-modal'), true); }
      return;
    }
    var dismiss = closest(ev.target, '[data-sui-modal-dismiss]');
    if (dismiss) {
      var owner = closest(dismiss, '.sui-modal');
      if (owner) { closeByName(owner.getAttribute('data-sui-modal'), true); }
    }
  });
  document.addEventListener('keydown', function(ev) {
    var modal = document.querySelector('.sui-modal.open');
    if (!modal) return;
    if (ev.key === 'Escape') {
      ev.preventDefault();
      closeByName(modal.getAttribute('data-sui-modal'), true);
      return;
    }
    if (ev.key === 'Tab') {
      var panel = modal.querySelector('.sui-modal-panel');
      if (!panel) return;
      var nodes = panel.querySelectorAll('button, input, select, textarea, a[href], [tabindex]:not([tabindex="-1"])');
      var list = Array.prototype.filter.call(nodes, function(n) { return !n.disabled && n.offsetParent !== null; });
      if (!list.length) { ev.preventDefault(); if (panel.focus) panel.focus(); return; }
      var first = list[0], last = list[list.length - 1];
      if (ev.shiftKey && document.activeElement === first) { ev.preventDefault(); last.focus(); }
      else if (!ev.shiftKey && document.activeElement === last) { ev.preventDefault(); first.focus(); }
    }
  });
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      setTimeout(function() { pending = false; applyOpen(); }, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
})();

const __SUI_CLASSES = Object.create(null);
__SUI_CLASSES["bar-part"] = 1;
__SUI_CLASSES["bar-target"] = 1;
__SUI_CLASSES["bar-total"] = 1;
__SUI_CLASSES["btn-action"] = 1;
__SUI_CLASSES["btn-add"] = 1;
__SUI_CLASSES["btn-cancel"] = 1;
__SUI_CLASSES["btn-ghost"] = 1;
__SUI_CLASSES["btn-icon"] = 1;
__SUI_CLASSES["btn-icon-sm"] = 1;
__SUI_CLASSES["btn-play"] = 1;
__SUI_CLASSES["btn-prio"] = 1;
__SUI_CLASSES["btn-stop"] = 1;
__SUI_CLASSES["btn-warn"] = 1;
__SUI_CLASSES["controls"] = 1;
__SUI_CLASSES["count-active"] = 1;
__SUI_CLASSES["count-done"] = 1;
__SUI_CLASSES["count-label"] = 1;
__SUI_CLASSES["count-value"] = 1;
__SUI_CLASSES["dl-file"] = 1;
__SUI_CLASSES["dl-link"] = 1;
__SUI_CLASSES["field-file"] = 1;
__SUI_CLASSES["file-bytes"] = 1;
__SUI_CLASSES["file-name"] = 1;
__SUI_CLASSES["file-row"] = 1;
__SUI_CLASSES["files-list"] = 1;
__SUI_CLASSES["hint"] = 1;
__SUI_CLASSES["mode-switch"] = 1;
__SUI_CLASSES["muted"] = 1;
__SUI_CLASSES["panel"] = 1;
__SUI_CLASSES["panel-targets"] = 1;
__SUI_CLASSES["part-bytes"] = 1;
__SUI_CLASSES["part-count"] = 1;
__SUI_CLASSES["part-counts"] = 1;
__SUI_CLASSES["part-error"] = 1;
__SUI_CLASSES["part-index"] = 1;
__SUI_CLASSES["part-row"] = 1;
__SUI_CLASSES["part-state"] = 1;
__SUI_CLASSES["parts-scroll"] = 1;
__SUI_CLASSES["row"] = 1;
__SUI_CLASSES["sb-item"] = 1;
__SUI_CLASSES["sb-label"] = 1;
__SUI_CLASSES["sb-value"] = 1;
__SUI_CLASSES["setting"] = 1;
__SUI_CLASSES["setting-control"] = 1;
__SUI_CLASSES["setting-text"] = 1;
__SUI_CLASSES["settings"] = 1;
__SUI_CLASSES["state-badge"] = 1;
__SUI_CLASSES["status"] = 1;
__SUI_CLASSES["statusbar"] = 1;
__SUI_CLASSES["statusbar-row"] = 1;
__SUI_CLASSES["statusbar-status"] = 1;
__SUI_CLASSES["sui-button"] = 1;
__SUI_CLASSES["sui-progress"] = 1;
__SUI_CLASSES["sui-tabpanels"] = 1;
__SUI_CLASSES["sui-tabs"] = 1;
__SUI_CLASSES["sui-toast"] = 1;
__SUI_CLASSES["target-bytes"] = 1;
__SUI_CLASSES["target-foot"] = 1;
__SUI_CLASSES["target-locked"] = 1;
__SUI_CLASSES["target-name"] = 1;
__SUI_CLASSES["target-row"] = 1;
__SUI_CLASSES["target-url"] = 1;
__SUI_CLASSES["targets-head"] = 1;
__SUI_CLASSES["toast-host"] = 1;
__SUI_CLASSES["topbar"] = 1;
__SUI_CLASSES["topbar-actions"] = 1;
__SUI_CLASSES["topbar-row"] = 1;
__SUI_CLASSES["verify"] = 1;


let partSize = `1 MiB`;
function setPartSize(v) { partSize = v; __sui_react("partSize"); }

let workers = `4`;
function setWorkers(v) { workers = v; __sui_react("workers"); }

let single = false;
function setSingle(v) { single = v; __sui_react("single"); }

let jobId = ``;
function setJobId(v) { jobId = v; __sui_react("jobId"); }

let jobState = `idle`;
function setJobState(v) { jobState = v; __sui_react("jobState"); }

let jobSingle = false;
function setJobSingle(v) { jobSingle = v; __sui_react("jobSingle"); }

let phase = `idle`;
function setPhase(v) { phase = v; __sui_react("phase"); }

let percent = 0;
function setPercent(v) { percent = v; __sui_react("percent"); }

let totalText = `0 B`;
function setTotalText(v) { totalText = v; __sui_react("totalText"); }

let receivedText = `0 B`;
function setReceivedText(v) { receivedText = v; __sui_react("receivedText"); }

let etaText = `--`;
function setEtaText(v) { etaText = v; __sui_react("etaText"); }

let verifyText = ``;
function setVerifyText(v) { verifyText = v; __sui_react("verifyText"); }

let message = ``;
function setMessage(v) { message = v; __sui_react("message"); }

let partsView = [];
function setPartsView(v) { partsView = v; __sui_react("partsView"); }

let partsTable = [];
function setPartsTable(v) { partsTable = v; __sui_react("partsTable"); }

let selectedPart = ``;
function setSelectedPart(v) { selectedPart = v; __sui_react("selectedPart"); }

let failedParts = 0;
function setFailedParts(v) { failedParts = v; __sui_react("failedParts"); }

let partsSignature = ``;
function setPartsSignature(v) { partsSignature = v; __sui_react("partsSignature"); }

let lastState = ``;
function setLastState(v) { lastState = v; __sui_react("lastState"); }

let lastError = ``;
function setLastError(v) { lastError = v; __sui_react("lastError"); }

let targets = [];
function setTargets(v) { targets = v; __sui_react("targets"); }

let targetSeq = 0;
function setTargetSeq(v) { targetSeq = v; __sui_react("targetSeq"); }

let partTotal = 0;
function setPartTotal(v) { partTotal = v; __sui_react("partTotal"); }

let partActive = 0;
function setPartActive(v) { partActive = v; __sui_react("partActive"); }

let partDone = 0;
function setPartDone(v) { partDone = v; __sui_react("partDone"); }

let files = [];
function setFiles(v) { files = v; __sui_react("files"); }

let filesTable = [];
function setFilesTable(v) { filesTable = v; __sui_react("filesTable"); }

let filesSignature = ``;
function setFilesSignature(v) { filesSignature = v; __sui_react("filesSignature"); }

let apiRoot = ``;
function setApiRoot(v) { apiRoot = v; __sui_react("apiRoot"); }

let apiFallback = `http://127.0.0.1:8090`;
function setApiFallback(v) { apiFallback = v; __sui_react("apiFallback"); }

let backendConfig = null;
function setBackendConfig(v) { backendConfig = v; __sui_react("backendConfig"); }

let jobsList = null;
function setJobsList(v) { jobsList = v; __sui_react("jobsList"); }

window.__SUI_STATE_GET = Object.create(null);
window.__SUI_STATE_GET["partSize"] = function(){ return partSize; };
window.__SUI_STATE_GET["workers"] = function(){ return workers; };
window.__SUI_STATE_GET["single"] = function(){ return single; };
window.__SUI_STATE_GET["jobId"] = function(){ return jobId; };
window.__SUI_STATE_GET["jobState"] = function(){ return jobState; };
window.__SUI_STATE_GET["jobSingle"] = function(){ return jobSingle; };
window.__SUI_STATE_GET["phase"] = function(){ return phase; };
window.__SUI_STATE_GET["percent"] = function(){ return percent; };
window.__SUI_STATE_GET["totalText"] = function(){ return totalText; };
window.__SUI_STATE_GET["receivedText"] = function(){ return receivedText; };
window.__SUI_STATE_GET["etaText"] = function(){ return etaText; };
window.__SUI_STATE_GET["verifyText"] = function(){ return verifyText; };
window.__SUI_STATE_GET["message"] = function(){ return message; };
window.__SUI_STATE_GET["partsView"] = function(){ return partsView; };
window.__SUI_STATE_GET["partsTable"] = function(){ return partsTable; };
window.__SUI_STATE_GET["selectedPart"] = function(){ return selectedPart; };
window.__SUI_STATE_GET["failedParts"] = function(){ return failedParts; };
window.__SUI_STATE_GET["partsSignature"] = function(){ return partsSignature; };
window.__SUI_STATE_GET["lastState"] = function(){ return lastState; };
window.__SUI_STATE_GET["lastError"] = function(){ return lastError; };
window.__SUI_STATE_GET["targets"] = function(){ return targets; };
window.__SUI_STATE_GET["targetSeq"] = function(){ return targetSeq; };
window.__SUI_STATE_GET["partTotal"] = function(){ return partTotal; };
window.__SUI_STATE_GET["partActive"] = function(){ return partActive; };
window.__SUI_STATE_GET["partDone"] = function(){ return partDone; };
window.__SUI_STATE_GET["files"] = function(){ return files; };
window.__SUI_STATE_GET["filesTable"] = function(){ return filesTable; };
window.__SUI_STATE_GET["filesSignature"] = function(){ return filesSignature; };
window.__SUI_STATE_GET["apiRoot"] = function(){ return apiRoot; };
window.__SUI_STATE_GET["apiFallback"] = function(){ return apiFallback; };
window.__SUI_STATE_GET["backendConfig"] = function(){ return backendConfig; };
window.__SUI_STATE_GET["jobsList"] = function(){ return jobsList; };


function apiUrl(path) {
  return ((apiRoot + `/api`) + path);
}

function backendUrl(path) {
  return (apiRoot + path);
}

function isConfig(body) {
  let text = String(body);
  if ((text === ``)) {
    return false;
  }
  return (text.includes(`partSize`) && text.includes(`workers`));
}

function backendDown() {
  return `the download manager is not running: start it with run.ps1`;
}

async function loadConfig() {
  ensureFirstTarget();
  setBackendConfig(null);
  const __sui_r = await __SUI_STD.http_getsafe(apiUrl(`/config`), `/api http://127.0.0.1:8090/api`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err === null)) {
    const __sui_r2 = await __SUI_STD.json_parse(body);
    const parsed = __sui_r2[0];
    const jerr = __sui_r2[1];
    if (((jerr === null) && isConfig(body))) {
      setBackendConfig(parsed);
    }
  }
  if ((backendConfig === null)) {
    setApiRoot(apiFallback);
    const __sui_r3 = await __SUI_STD.http_getsafe(apiUrl(`/config`), `/api http://127.0.0.1:8090/api`);
    const body2 = __sui_r3[0];
    const err2 = __sui_r3[1];
    if ((err2 === null)) {
      const __sui_r4 = await __SUI_STD.json_parse(body2);
      const parsed2 = __sui_r4[0];
      const jerr2 = __sui_r4[1];
      if (((jerr2 === null) && isConfig(body2))) {
        setBackendConfig(parsed2);
      }
    }
  }
  if ((backendConfig === null)) {
    setApiRoot(``);
    notifyError(backendDown());
    return;
  }
  let cfg = backendConfig;
  if ((targets.length > 0)) {
    let first = targetById(String(targets[0][`id`]));
    if (((String(first[`fileId`]) === ``) && (String(first[`url`]) === ``))) {
      setTarget(String(first[`id`]), `url`, String(cfg[`url`]));
      syncTargetName(String(first[`id`]));
    }
  }
  let cfgSize = String(cfg[`partSize`]);
  if ((cfgSize !== ``)) {
    setPartSize(cfgSize);
  }
  let cfgWorkers = String(cfg[`workers`]);
  if ((cfgWorkers !== ``)) {
    setWorkers(cfgWorkers);
  }
}

function fileNameFromUrl(u) {
  let clean = u.trim();
  if ((clean === ``)) {
    return `download.bin`;
  }
  let noHash = clean.split(`#`);
  let noQuery = noHash[0].split(`?`);
  let segments = noQuery[0].split(`/`);
  let last = segments[(segments.length - 1)].trim();
  if ((last === ``)) {
    return `download.bin`;
  }
  if (last.includes(`:`)) {
    return `download.bin`;
  }
  let safe = __sui_replace_regex(last, `[^A-Za-z0-9._-]`, `_`);
  if ((safe === ``)) {
    return `download.bin`;
  }
  if ((safe === `.`)) {
    return `download.bin`;
  }
  if ((safe === `..`)) {
    return `download.bin`;
  }
  return safe;
}

function emptyTarget(id) {
  return ({"id": id, "fileId": ``, "url": ``, "name": ``, "named": `no`, "priority": `normal`, "state": `new`, "percent": 0, "received": `0 B`, "total": `0 B`, "merged": ``});
}

function newTarget() {
  setTargetSeq((targetSeq + 1));
  return emptyTarget((`t` + String(targetSeq)));
}

function targetById(id) {
  let i = 0;
  while ((i < targets.length)) {
    if ((String(targets[i][`id`]) === id)) {
      return targets[i];
    }
    i = (i + 1);
  }
  return emptyTarget(id);
}

function targetByFileId(fileId) {
  let i = 0;
  while ((i < targets.length)) {
    if ((String(targets[i][`fileId`]) === fileId)) {
      return targets[i];
    }
    i = (i + 1);
  }
  return emptyTarget(``);
}

function setTarget(id, field, value) {
  let out = [];
  let i = 0;
  while ((i < targets.length)) {
    let r = targets[i];
    if ((String(r[`id`]) === id)) {
      r[field] = value;
    }
    out = [...out, r];
    i = (i + 1);
  }
  setTargets(out);
}

function removeTarget(id) {
  let out = [];
  let i = 0;
  while ((i < targets.length)) {
    if ((String(targets[i][`id`]) !== id)) {
      out = [...out, targets[i]];
    }
    i = (i + 1);
  }
  setTargets(out);
}

function addTarget() {
  setTargets([...targets, newTarget()]);
}

function ensureFirstTarget() {
  if ((targets.length > 0)) {
    return;
  }
  addTarget();
}

function syncTargetName(id) {
  let r = targetById(id);
  if ((String(r[`named`]) === `yes`)) {
    return;
  }
  let u = String(r[`url`]).trim();
  if ((u === ``)) {
    return;
  }
  let wanted = fileNameFromUrl(u);
  if (((wanted === ``) || (wanted === String(r[`name`])))) {
    return;
  }
  setTarget(id, `name`, wanted);
}

function onTargetUrl(ev, id) {
  setTarget(id, `url`, String(ev[`target`][`value`]));
  syncTargetName(id);
}

function onTargetName(ev, id) {
  setTarget(id, `name`, String(ev[`target`][`value`]));
  setTarget(id, `named`, `yes`);
}

function prioIcon(p) {
  if ((p === `high`)) {
    return `↑`;
  }
  if ((p === `low`)) {
    return `↓`;
  }
  return `→`;
}

function prioLabel(p) {
  if ((p === `high`)) {
    return `High priority`;
  }
  if ((p === `low`)) {
    return `Low priority`;
  }
  return `Normal priority`;
}

function nextPriority(p) {
  if ((p === `low`)) {
    return `normal`;
  }
  if ((p === `normal`)) {
    return `high`;
  }
  return `low`;
}

async function onTargetPriority(ev, id) {
  let r = targetById(id);
  let want = nextPriority(String(r[`priority`]));
  setTarget(id, `priority`, want);
  if ((String(r[`fileId`]) !== ``)) {
    setFilePriority(String(r[`fileId`]), want);
  }
}

function playIcon(t) {
  let s = String(t[`state`]);
  if ((s === `running`)) {
    return `❚❚`;
  }
  if ((s === `done`)) {
    return `✓`;
  }
  if ((s === `error`)) {
    return `!`;
  }
  return `▶`;
}

function playHint(t) {
  let s = String(t[`state`]);
  if ((s === `running`)) {
    return `Pause the download of this file`;
  }
  if ((s === `paused`)) {
    return `Resume the download of this file`;
  }
  if ((s === `stopped`)) {
    return `Start the download again with the current URL`;
  }
  if ((s === `done`)) {
    return `This file is complete`;
  }
  return `Start the download of this file`;
}

function stopIcon(t) {
  let s = String(t[`state`]);
  if ((((s === `running`) || (s === `paused`)) || (s === `done`))) {
    return `■`;
  }
  return `↻`;
}

function stopHint(t) {
  let s = String(t[`state`]);
  if (((s === `running`) || (s === `paused`))) {
    return `Stop this file: the URL can be changed again`;
  }
  if ((s === `done`)) {
    return `Close this file: the URL can be changed again`;
  }
  return `Start the download again with the current URL`;
}

async function onTargetPlay(ev, id) {
  let r = targetById(id);
  let s = String(r[`state`]);
  if ((String(r[`fileId`]) === ``)) {
    startTarget(id);
    return;
  }
  if ((s === `running`)) {
    fileAction(String(r[`fileId`]), `pause`);
    return;
  }
  if ((s === `paused`)) {
    fileAction(String(r[`fileId`]), `resume`);
    return;
  }
  if ((s === `stopped`)) {
    restartTarget(id);
    return;
  }
}

async function onTargetStop(ev, id) {
  let r = targetById(id);
  let s = String(r[`state`]);
  if ((String(r[`fileId`]) === ``)) {
    return;
  }
  if ((((s === `running`) || (s === `paused`)) || (s === `done`))) {
    fileAction(String(r[`fileId`]), `stop`);
    return;
  }
  restartTarget(id);
}

async function onTargetCancel(ev, id) {
  let r = targetById(id);
  let fileId = String(r[`fileId`]);
  if ((fileId === ``)) {
    removeTarget(id);
    return;
  }
  const __sui_r = await __SUI_STD.http_postsafe(apiUrl(((((`/downloads/` + jobId) + `/files/`) + fileId) + `/cancel`)), ``, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err !== null)) {
    notifyError((`the cancel failed: ` + err));
    return;
  }
  const __sui_r2 = await __SUI_STD.json_parse(body);
  const job = __sui_r2[0];
  const jerr = __sui_r2[1];
  if ((jerr !== null)) {
    notifyError((`invalid response: ` + jerr));
    return;
  }
  if ((Object.prototype.hasOwnProperty.call(job, `error`) && (String(job[`error`]) !== ``))) {
    notifyError((`the cancel failed: ` + String(job[`error`])));
    return;
  }
  removeTarget(id);
  updateJob(job);
}

function freeTargetId() {
  let i = 0;
  while ((i < targets.length)) {
    let r = targets[i];
    if (((String(r[`fileId`]) === ``) && (String(r[`url`]) === ``))) {
      return String(r[`id`]);
    }
    i = (i + 1);
  }
  return ``;
}

function syncTargets(jobFiles) {
  let i = 0;
  while ((i < jobFiles.length)) {
    let f = jobFiles[i];
    let fileId = String(f[`id`]);
    let row = targetByFileId(fileId);
    let id = String(row[`id`]);
    if ((id === ``)) {
      let free = freeTargetId();
      if ((free !== ``)) {
        id = free;
      } else {
        let fresh = newTarget();
        setTargets([...targets, fresh]);
        id = String(fresh[`id`]);
      }
      setTarget(id, `fileId`, fileId);
      setTarget(id, `url`, String(f[`url`]));
      setTarget(id, `name`, String(f[`file`]));
      setTarget(id, `named`, `yes`);
      setTarget(id, `priority`, String(f[`priority`]));
    }
    setTarget(id, `state`, String(f[`state`]));
    setTarget(id, `percent`, f[`percent`]);
    setTarget(id, `received`, String(f[`receivedText`]));
    setTarget(id, `total`, String(f[`totalText`]));
    setTarget(id, `merged`, String(f[`merged`]));
    i = (i + 1);
  }
}

function notifyError(text) {
  if ((text === lastError)) {
    return;
  }
  setLastError(text);
  setMessage(text);
  __sui_toast_show(text, `error`, 5000);
}

function notifyInfo(text) {
  setMessage(text);
  __sui_toast_show(text, `info`, 3000);
}

async function refresh() {
  setJobsList(null);
  const __sui_r = await __SUI_STD.http_getsafe(apiUrl(`/downloads`), `/api http://127.0.0.1:8090/api`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err === null)) {
    const __sui_r2 = await __SUI_STD.json_parse(body);
    const parsed = __sui_r2[0];
    const jerr = __sui_r2[1];
    if ((jerr === null)) {
      setJobsList(parsed[`jobs`]);
    }
  }
  if (((jobsList === null) && (apiRoot === ``))) {
    setApiRoot(apiFallback);
    const __sui_r3 = await __SUI_STD.http_getsafe(apiUrl(`/downloads`), `/api http://127.0.0.1:8090/api`);
    const body2 = __sui_r3[0];
    const err2 = __sui_r3[1];
    if ((err2 === null)) {
      const __sui_r4 = await __SUI_STD.json_parse(body2);
      const parsed2 = __sui_r4[0];
      const jerr2 = __sui_r4[1];
      if ((jerr2 === null)) {
        setJobsList(parsed2[`jobs`]);
      }
    }
    if ((jobsList === null)) {
      setApiRoot(``);
    }
  }
  if ((jobsList === null)) {
    notifyError(backendDown());
    return;
  }
  let jobs = jobsList;
  const __sui_r5 = __SUI_STD.table_display(`jobs`, jobs);
  const n = __sui_r5[0];
  const terr = __sui_r5[1];
  if ((terr !== null)) {
    notifyError((`jobs table not updated: ` + terr));
  }
  if ((jobs.length === 0)) {
    return;
  }
  updateJob(jobs[0]);
}

async function updateJob(job) {
  setJobId(String(job[`id`]));
  setJobState(String(job[`state`]));
  setJobSingle(job[`single`]);
  setPercent(job[`percent`]);
  setTotalText(String(job[`totalText`]));
  setReceivedText(String(job[`receivedText`]));
  setEtaText(String(job[`etaText`]));
  setVerifyText(String(job[`verified`]));
  setFailedParts(job[`failedParts`]);
  let parts = job[`parts`];
  setPartsView([...__sui_filter(parts, isStarted)].reverse());
  setPartsTable(parts);
  setPartTotal(parts.length);
  setPartActive(countParts(parts, `downloading`));
  setPartDone(countParts(parts, `done`));
  let jobFiles = job[`files`];
  setFiles(jobFiles);
  setFilesTable(fileTable(jobFiles));
  refreshPartsTable(((((((jobId + `|`) + jobState) + `|`) + String(failedParts)) + `|`) + String(partDone)));
  refreshFilesTable(((jobId + `|`) + filesStamp(jobFiles)));
  syncTargets(jobFiles);
  syncPhase();
  let previous = lastState;
  setLastState(jobState);
  if ((jobState !== previous)) {
    if ((jobState === `done`)) {
      announceDone(job);
    }
    if ((jobState === `error`)) {
      notifyError((`the download failed: ` + String(job[`error`])));
    }
  }
}

function fileRow(f) {
  return [String(f[`id`]), String(f[`file`]), String(f[`priority`]), String(f[`state`]), String(f[`partCount`]), String(f[`totalText`])];
}

function fileRows(jobFiles) {
  return __sui_map(jobFiles, fileRow);
}

function fileTable(jobFiles) {
  return ({"columns": [`Id`, `File`, `Priority`, `State`, `Parts`, `Size`], "rows": fileRows(jobFiles)});
}

function filesStamp(jobFiles) {
  let stamp = ``;
  let i = 0;
  while ((i < jobFiles.length)) {
    let f = jobFiles[i];
    stamp = ((((((stamp + String(f[`id`])) + `:`) + String(f[`priority`])) + `:`) + String(f[`state`])) + `;`);
    i = (i + 1);
  }
  return stamp;
}

function syncPhase() {
  if ((jobState === `running`)) {
    setPhase(`running`);
    return;
  }
  if ((jobState === `paused`)) {
    setPhase(`paused`);
    return;
  }
  setPhase(`idle`);
}

function phaseIcon() {
  if ((phase === `running`)) {
    return `❚❚`;
  }
  return `▶`;
}

function phaseHint() {
  if ((phase === `running`)) {
    return `Pause every download`;
  }
  if ((phase === `paused`)) {
    return `Resume every download`;
  }
  return `Pause or resume every download`;
}

function targetLocked(t) {
  let s = String(t[`state`]);
  if ((s === `new`)) {
    return false;
  }
  if ((s === `stopped`)) {
    return false;
  }
  return true;
}

function isStarted(part) {
  if ((String(part[`State`]) === `pending`)) {
    return (String(part[`Received`]) !== `0`);
  }
  return true;
}

function countParts(parts, want) {
  let n = 0;
  let i = 0;
  while ((i < parts.length)) {
    if ((String(parts[i][`State`]) === want)) {
      n = (n + 1);
    }
    i = (i + 1);
  }
  return n;
}

function refreshPartsTable(signature) {
  if ((signature === partsSignature)) {
    return;
  }
  setPartsSignature(signature);
  const __sui_r = __SUI_STD.table_display(`parts`, partsTable);
  const n = __sui_r[0];
  const terr = __sui_r[1];
  if ((terr !== null)) {
    notifyError((`parts table not updated: ` + terr));
  }
}

function refreshFilesTable(signature) {
  if ((signature === filesSignature)) {
    return;
  }
  setFilesSignature(signature);
  const __sui_r = __SUI_STD.table_display(`files`, filesTable);
  const n = __sui_r[0];
  const terr = __sui_r[1];
  if ((terr !== null)) {
    notifyError((`files table not updated: ` + terr));
  }
}

async function announceDone(job) {
  __sui_toast_show(((`Download complete: ` + String(job[`files`].length)) + ` files`), `success`, 4000);
  const __sui_r = await __SUI_STD.sound_playsafe(backendUrl(`/sound.wav`), `/sound.wav http://127.0.0.1:8090/sound.wav`);
  const ok = __sui_r[0];
  const err = __sui_r[1];
  if ((err !== null)) {
    setMessage(`download complete (the browser blocked the sound)`);
  }
}

async function onAction() {
  if ((jobId === ``)) {
    return;
  }
  if ((phase === `running`)) {
    runAction(`pause`);
    return;
  }
  if ((phase === `paused`)) {
    runAction(`resume`);
  }
}

async function startTarget(id) {
  let r = targetById(id);
  let u = String(r[`url`]).trim();
  if ((u === ``)) {
    notifyError(`type the URL of the file to download`);
    return;
  }
  let name = String(r[`name`]);
  let prio = String(r[`priority`]);
  if ((jobId === ``)) {
    const __sui_r = await __SUI_STD.json_stringify(({"url": u, "partSize": partSize, "workers": workers, "single": single, "file": name, "priority": prio}));
    const payload = __sui_r[0];
    const jerr = __sui_r[1];
    if ((jerr !== null)) {
      notifyError((`invalid request: ` + jerr));
      return;
    }
    const __sui_r2 = await __SUI_STD.http_postsafe(apiUrl(`/downloads`), payload, `/api http://127.0.0.1:8090/api`);
    const body = __sui_r2[0];
    const err = __sui_r2[1];
    if ((err !== null)) {
      notifyError((`could not start: ` + err));
      return;
    }
    const __sui_r3 = await __SUI_STD.json_parse(body);
    const job = __sui_r3[0];
    const perr = __sui_r3[1];
    if ((perr !== null)) {
      notifyError((`invalid response: ` + perr));
      return;
    }
    if ((Object.prototype.hasOwnProperty.call(job, `error`) && (String(job[`error`]) !== ``))) {
      notifyError((`could not start: ` + String(job[`error`])));
      return;
    }
    bindTargetFile(id, job);
    setSelectedPart(``);
    updateJob(job);
    return;
  }
  const __sui_r4 = await __SUI_STD.json_stringify(({"url": u, "file": name, "priority": prio}));
  const payload = __sui_r4[0];
  const jerr = __sui_r4[1];
  if ((jerr !== null)) {
    notifyError((`invalid request: ` + jerr));
    return;
  }
  const __sui_r5 = await __SUI_STD.http_postsafe(apiUrl(((`/downloads/` + jobId) + `/files`)), payload, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r5[0];
  const err = __sui_r5[1];
  if ((err !== null)) {
    notifyError((`could not add the file: ` + err));
    return;
  }
  const __sui_r6 = await __SUI_STD.json_parse(body);
  const job = __sui_r6[0];
  const perr = __sui_r6[1];
  if ((perr !== null)) {
    notifyError((`invalid response: ` + perr));
    return;
  }
  if ((Object.prototype.hasOwnProperty.call(job, `error`) && (String(job[`error`]) !== ``))) {
    notifyError((`could not add the file: ` + String(job[`error`])));
    return;
  }
  bindTargetFile(id, job);
  updateJob(job);
}

function bindTargetFile(id, job) {
  let jobFiles = job[`files`];
  if ((jobFiles.length === 0)) {
    return;
  }
  let f = jobFiles[(jobFiles.length - 1)];
  setTarget(id, `fileId`, String(f[`id`]));
  setTarget(id, `state`, String(f[`state`]));
}

async function restartTarget(id) {
  let r = targetById(id);
  let fileId = String(r[`fileId`]);
  if ((fileId === ``)) {
    return;
  }
  const __sui_r = await __SUI_STD.json_stringify(({"url": String(r[`url`]).trim(), "file": String(r[`name`])}));
  const payload = __sui_r[0];
  const jerr = __sui_r[1];
  if ((jerr !== null)) {
    notifyError((`invalid request: ` + jerr));
    return;
  }
  const __sui_r2 = await __SUI_STD.http_postsafe(apiUrl(((((`/downloads/` + jobId) + `/files/`) + fileId) + `/restart`)), payload, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r2[0];
  const err = __sui_r2[1];
  if ((err !== null)) {
    notifyError((`could not start the file again: ` + err));
    return;
  }
  const __sui_r3 = await __SUI_STD.json_parse(body);
  const job = __sui_r3[0];
  const perr = __sui_r3[1];
  if ((perr !== null)) {
    notifyError((`invalid response: ` + perr));
    return;
  }
  if ((Object.prototype.hasOwnProperty.call(job, `error`) && (String(job[`error`]) !== ``))) {
    notifyError((`could not start the file again: ` + String(job[`error`])));
    return;
  }
  updateJob(job);
}

async function fileAction(fileId, act) {
  if ((jobId === ``)) {
    notifyError(`no download to manage`);
    return;
  }
  const __sui_r = await __SUI_STD.http_postsafe(apiUrl((((((`/downloads/` + jobId) + `/files/`) + fileId) + `/`) + act)), ``, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err !== null)) {
    notifyError((((`the ` + act) + ` failed: `) + err));
    return;
  }
  const __sui_r2 = await __SUI_STD.json_parse(body);
  const job = __sui_r2[0];
  const jerr = __sui_r2[1];
  if ((jerr !== null)) {
    notifyError((`invalid response: ` + jerr));
    return;
  }
  if ((Object.prototype.hasOwnProperty.call(job, `error`) && (String(job[`error`]) !== ``))) {
    notifyError((((`the ` + act) + ` failed: `) + String(job[`error`])));
    return;
  }
  updateJob(job);
}

async function setFilePriority(fileId, want) {
  if ((jobId === ``)) {
    return;
  }
  const __sui_r = await __SUI_STD.json_stringify(({"priority": want}));
  const payload = __sui_r[0];
  const jerr = __sui_r[1];
  if ((jerr !== null)) {
    notifyError((`invalid request: ` + jerr));
    return;
  }
  const __sui_r2 = await __SUI_STD.http_postsafe(apiUrl(((((`/downloads/` + jobId) + `/files/`) + fileId) + `/priority`)), payload, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r2[0];
  const err = __sui_r2[1];
  if ((err !== null)) {
    notifyError((`could not change the priority: ` + err));
    return;
  }
  const __sui_r3 = await __SUI_STD.json_parse(body);
  const job = __sui_r3[0];
  const perr = __sui_r3[1];
  if ((perr !== null)) {
    notifyError((`invalid response: ` + perr));
    return;
  }
  if ((Object.prototype.hasOwnProperty.call(job, `error`) && (String(job[`error`]) !== ``))) {
    notifyError((`could not change the priority: ` + String(job[`error`])));
    return;
  }
  updateJob(job);
}

async function runAction(act) {
  if ((jobId === ``)) {
    notifyError(`no download to manage`);
    return;
  }
  const __sui_r = await __SUI_STD.http_postsafe(apiUrl((((`/downloads/` + jobId) + `/`) + act)), ``, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err !== null)) {
    notifyError((((`operation ` + act) + ` failed: `) + err));
    return;
  }
  const __sui_r2 = await __SUI_STD.json_parse(body);
  const job = __sui_r2[0];
  const jerr = __sui_r2[1];
  if ((jerr !== null)) {
    notifyError((`invalid response: ` + jerr));
    return;
  }
  updateJob(job);
}

function openSettings() {
  __sui_modal_open(`settings`);
}

function closeSettings() {
  __sui_modal_close(`settings`);
}

function onPartSelected() {
  const __sui_r = __SUI_STD.table_selectedrows(`parts`);
  const sel = __sui_r[0];
  const e = __sui_r[1];
  if ((e !== null)) {
    notifyError((`cannot read the selection: ` + e));
    return;
  }
  if ((sel.length === 0)) {
    return;
  }
  const __sui_r2 = __SUI_STD.table_data(`parts`);
  const rows = __sui_r2[0];
  const e2 = __sui_r2[1];
  if ((e2 !== null)) {
    notifyError((`cannot read the parts table: ` + e2));
    return;
  }
  let idx = sel[0];
  if (((idx < 0) || (idx >= rows.length))) {
    return;
  }
  setSelectedPart(rows[idx][0]);
  setMessage(((`Part ` + selectedPart) + ` selected: press Retry to download it again.`));
}

async function retrySelectedPart() {
  if ((jobId === ``)) {
    notifyError(`no download to manage`);
    return;
  }
  if ((selectedPart === ``)) {
    notifyError(`select a row in the Data tab first`);
    return;
  }
  const __sui_r = await __SUI_STD.http_postsafe(apiUrl(((((`/downloads/` + jobId) + `/parts/`) + selectedPart) + `/retry`)), ``, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err !== null)) {
    notifyError((`retry failed: ` + err));
    return;
  }
  const __sui_r2 = await __SUI_STD.json_parse(body);
  const job = __sui_r2[0];
  const jerr = __sui_r2[1];
  if ((jerr !== null)) {
    notifyError((`invalid response: ` + jerr));
    return;
  }
  updateJob(job);
  notifyInfo(((`part ` + selectedPart) + ` queued again`));
}

async function retryFailedParts() {
  if ((jobId === ``)) {
    notifyError(`no download to manage`);
    return;
  }
  const __sui_r = await __SUI_STD.http_postsafe(apiUrl(((`/downloads/` + jobId) + `/retry-failed`)), ``, `/api http://127.0.0.1:8090/api`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err !== null)) {
    notifyError((`retry failed: ` + err));
    return;
  }
  const __sui_r2 = await __SUI_STD.json_parse(body);
  const job = __sui_r2[0];
  const jerr = __sui_r2[1];
  if ((jerr !== null)) {
    notifyError((`invalid response: ` + jerr));
    return;
  }
  updateJob(job);
  notifyInfo(`the failed parts are queued again`);
}

async function boot() {
  ensureFirstTarget();
  setInterval(() => refresh(), 500);
  loadConfig();
}

function refreshData() {
  setPartsSignature(``);
  refreshPartsTable(`manual`);
  setFilesSignature(``);
  refreshFilesTable(`manual`);
}

window.__SUI_USER_FUNCS = Object.create(null);
window.__SUI_USER_FUNCS["apiUrl"] = 1;
window.__SUI_USER_FUNCS["backendUrl"] = 1;
window.__SUI_USER_FUNCS["isConfig"] = 1;
window.__SUI_USER_FUNCS["backendDown"] = 1;
window.__SUI_USER_FUNCS["loadConfig"] = 1;
window.__SUI_USER_FUNCS["fileNameFromUrl"] = 1;
window.__SUI_USER_FUNCS["emptyTarget"] = 1;
window.__SUI_USER_FUNCS["newTarget"] = 1;
window.__SUI_USER_FUNCS["targetById"] = 1;
window.__SUI_USER_FUNCS["targetByFileId"] = 1;
window.__SUI_USER_FUNCS["setTarget"] = 1;
window.__SUI_USER_FUNCS["removeTarget"] = 1;
window.__SUI_USER_FUNCS["addTarget"] = 1;
window.__SUI_USER_FUNCS["ensureFirstTarget"] = 1;
window.__SUI_USER_FUNCS["syncTargetName"] = 1;
window.__SUI_USER_FUNCS["onTargetUrl"] = 1;
window.__SUI_USER_FUNCS["onTargetName"] = 1;
window.__SUI_USER_FUNCS["prioIcon"] = 1;
window.__SUI_USER_FUNCS["prioLabel"] = 1;
window.__SUI_USER_FUNCS["nextPriority"] = 1;
window.__SUI_USER_FUNCS["onTargetPriority"] = 1;
window.__SUI_USER_FUNCS["playIcon"] = 1;
window.__SUI_USER_FUNCS["playHint"] = 1;
window.__SUI_USER_FUNCS["stopIcon"] = 1;
window.__SUI_USER_FUNCS["stopHint"] = 1;
window.__SUI_USER_FUNCS["onTargetPlay"] = 1;
window.__SUI_USER_FUNCS["onTargetStop"] = 1;
window.__SUI_USER_FUNCS["onTargetCancel"] = 1;
window.__SUI_USER_FUNCS["freeTargetId"] = 1;
window.__SUI_USER_FUNCS["syncTargets"] = 1;
window.__SUI_USER_FUNCS["notifyError"] = 1;
window.__SUI_USER_FUNCS["notifyInfo"] = 1;
window.__SUI_USER_FUNCS["refresh"] = 1;
window.__SUI_USER_FUNCS["updateJob"] = 1;
window.__SUI_USER_FUNCS["fileRow"] = 1;
window.__SUI_USER_FUNCS["fileRows"] = 1;
window.__SUI_USER_FUNCS["fileTable"] = 1;
window.__SUI_USER_FUNCS["filesStamp"] = 1;
window.__SUI_USER_FUNCS["syncPhase"] = 1;
window.__SUI_USER_FUNCS["phaseIcon"] = 1;
window.__SUI_USER_FUNCS["phaseHint"] = 1;
window.__SUI_USER_FUNCS["targetLocked"] = 1;
window.__SUI_USER_FUNCS["isStarted"] = 1;
window.__SUI_USER_FUNCS["countParts"] = 1;
window.__SUI_USER_FUNCS["refreshPartsTable"] = 1;
window.__SUI_USER_FUNCS["refreshFilesTable"] = 1;
window.__SUI_USER_FUNCS["announceDone"] = 1;
window.__SUI_USER_FUNCS["onAction"] = 1;
window.__SUI_USER_FUNCS["startTarget"] = 1;
window.__SUI_USER_FUNCS["bindTargetFile"] = 1;
window.__SUI_USER_FUNCS["restartTarget"] = 1;
window.__SUI_USER_FUNCS["fileAction"] = 1;
window.__SUI_USER_FUNCS["setFilePriority"] = 1;
window.__SUI_USER_FUNCS["runAction"] = 1;
window.__SUI_USER_FUNCS["openSettings"] = 1;
window.__SUI_USER_FUNCS["closeSettings"] = 1;
window.__SUI_USER_FUNCS["onPartSelected"] = 1;
window.__SUI_USER_FUNCS["retrySelectedPart"] = 1;
window.__SUI_USER_FUNCS["retryFailedParts"] = 1;
window.__SUI_USER_FUNCS["boot"] = 1;
window.__SUI_USER_FUNCS["refreshData"] = 1;


var __SUI_JSON = (function () {
  var MAX_BYTES = 20000000;
  var MAX_DEPTH = 128;
  var MAX_ERRORS = 100;
  function fail(message) { return [null, String(message == null ? 'JSON error' : message).slice(0, 500)]; }
  function ok(value) { return [value, null]; }
  function hasOwn(obj, key) { return Object.prototype.hasOwnProperty.call(obj, key); }
  function isObject(value) { return value !== null && typeof value === 'object' && !Array.isArray(value); }
  function setKey(obj, key, value) {
    if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
    Object.defineProperty(obj, key, { value: value, writable: true, enumerable: true, configurable: true });
  }
  function cloneValue(value, depth, seen) {
    if (value === null || typeof value !== 'object') return value;
    if (depth > MAX_DEPTH) throw new Error('JSON value is nested too deeply (max ' + MAX_DEPTH + ')');
    seen = seen || [];
    if (seen.indexOf(value) >= 0) throw new Error('JSON value contains a cycle');
    seen.push(value);
    var out, i, keys;
    if (Array.isArray(value)) {
      out = [];
      for (i = 0; i < value.length; i++) out.push(cloneValue(value[i], depth + 1, seen));
    } else {
      out = {};
      keys = Object.keys(value);
      for (i = 0; i < keys.length; i++) setKey(out, keys[i], cloneValue(value[keys[i]], depth + 1, seen));
    }
    seen.pop();
    return out;
  }
  function clone(value) {
    try { return ok(cloneValue(value, 0)); }
    catch (e) { return fail('json.clone: ' + (e && e.message ? e.message : String(e))); }
  }
  function equalValue(a, b, depth) {
    if (a === b) return true;
    if (a === null || b === null || typeof a !== 'object' || typeof b !== 'object') return false;
    if (depth > MAX_DEPTH) return false;
    if (Array.isArray(a) !== Array.isArray(b)) return false;
    var i, keys;
    if (Array.isArray(a)) {
      if (a.length !== b.length) return false;
      for (i = 0; i < a.length; i++) if (!equalValue(a[i], b[i], depth + 1)) return false;
      return true;
    }
    keys = Object.keys(a);
    if (keys.length !== Object.keys(b).length) return false;
    for (i = 0; i < keys.length; i++) {
      if (!hasOwn(b, keys[i])) return false;
      if (!equalValue(a[keys[i]], b[keys[i]], depth + 1)) return false;
    }
    return true;
  }
  function equal(a, b) {
    try { return ok(equalValue(a, b, 0)); }
    catch (e) { return fail('json.equal: ' + (e && e.message ? e.message : String(e))); }
  }
  function parsePointer(pointer) {
    if (typeof pointer !== 'string') throw new Error('JSON Pointer must be a string');
    if (pointer === '') return [];
    if (pointer.charAt(0) !== '/') throw new Error('JSON Pointer must be empty or start with "/"');
    var raw = pointer.slice(1).split('/'), out = [];
    for (var i = 0; i < raw.length; i++) {
      var token = raw[i];
      if (/~(?![01])/.test(token)) throw new Error('invalid JSON Pointer escape in "' + token + '"');
      out.push(token.replace(/~1/g, '/').replace(/~0/g, '~'));
    }
    return out;
  }
  function pointerString(tokens) {
    if (!tokens || tokens.length === 0) return '';
    var parts = [];
    for (var i = 0; i < tokens.length; i++) {
      parts.push(String(tokens[i]).replace(/~/g, '~0').replace(/\//g, '~1'));
    }
    return '/' + parts.join('/');
  }
  function arrayIndex(token, length, allowEnd) {
    if (token === '-') {
      if (allowEnd) return length;
      throw new Error('"-" is only valid as the last token of an add operation');
    }
    if (!/^(0|[1-9][0-9]*)$/.test(token)) throw new Error('invalid array index "' + token + '"');
    var index = Number(token);
    if (index > length || (!allowEnd && index === length)) throw new Error('array index ' + index + ' is out of range');
    return index;
  }
  function getTokens(root, tokens, pointerText) {
    var current = root;
    for (var i = 0; i < tokens.length; i++) {
      var token = tokens[i];
      if (Array.isArray(current)) current = current[arrayIndex(token, current.length, false)];
      else if (isObject(current)) {
        if (!hasOwn(current, token)) throw new Error('path not found at "' + pointerString(tokens.slice(0, i + 1)) + '"');
        current = current[token];
      } else throw new Error('cannot traverse "' + pointerText + '" through a scalar');
    }
    return current;
  }
  function getPointer(root, pointer) {
    try {
      var tokens = parsePointer(pointer);
      return ok(getTokens(root, tokens, pointer));
    } catch (e) { return fail('json.pointer: ' + (e && e.message ? e.message : String(e))); }
  }
  function addTokens(root, tokens, value) {
    if (tokens.length === 0) return value;
    var parent = getTokens(root, tokens.slice(0, -1), pointerString(tokens));
    var last = tokens[tokens.length - 1];
    if (Array.isArray(parent)) {
      var index = arrayIndex(last, parent.length, true);
      parent.splice(index, 0, value);
    } else if (isObject(parent)) {
      setKey(parent, last, value);
    } else throw new Error('cannot add a value under a scalar');
    return root;
  }
  function replaceTokens(root, tokens, value) {
    if (tokens.length === 0) return value;
    var parent = getTokens(root, tokens.slice(0, -1), pointerString(tokens));
    var last = tokens[tokens.length - 1];
    if (Array.isArray(parent)) {
      var index = arrayIndex(last, parent.length, false);
      parent[index] = value;
    } else if (isObject(parent)) {
      if (!hasOwn(parent, last)) throw new Error('path not found at "' + pointerString(tokens) + '"');
      setKey(parent, last, value);
    } else throw new Error('cannot replace a value under a scalar');
    return root;
  }
  function removeTokens(root, tokens) {
    if (tokens.length === 0) throw new Error('cannot remove the document root');
    var parent = getTokens(root, tokens.slice(0, -1), pointerString(tokens));
    var last = tokens[tokens.length - 1];
    if (Array.isArray(parent)) {
      var index = arrayIndex(last, parent.length, false);
      parent.splice(index, 1);
    } else if (isObject(parent)) {
      if (!hasOwn(parent, last)) throw new Error('path not found at "' + pointerString(tokens) + '"');
      delete parent[last];
    } else throw new Error('cannot remove a value under a scalar');
    return root;
  }
  function pointer(value, path) { return getPointer(value, path); }
  function setPointer(value, path, newValue) {
    try {
      var copy = cloneValue(value, 0);
      var tokens = parsePointer(path);
      var exists = true;
      try { getTokens(copy, tokens, path); } catch (e) { exists = false; }
      var next = cloneValue(newValue, 0);
      copy = exists ? replaceTokens(copy, tokens, next) : addTokens(copy, tokens, next);
      return ok(copy);
    } catch (e) { return fail('json.setPointer: ' + (e && e.message ? e.message : String(e))); }
  }
  function removePointer(value, path) {
    try {
      var copy = cloneValue(value, 0);
      removeTokens(copy, parsePointer(path));
      return ok(copy);
    } catch (e) { return fail('json.removePointer: ' + (e && e.message ? e.message : String(e))); }
  }
  function requireValue(op, index) {
    if (!hasOwn(op, 'value')) throw new Error('operation ' + (index + 1) + ' requires "value"');
    return cloneValue(op.value, 0);
  }
  function patch(value, operations) {
    try {
      if (typeof operations === 'string') {
        var parsed = parse(operations);
        if (parsed[1] !== null) return [null, 'json.patch: ' + parsed[1]];
        operations = parsed[0];
      }
      if (!Array.isArray(operations)) throw new Error('operations must be an array');
      if (operations.length > 10000) throw new Error('too many operations (max 10000)');
      var root = cloneValue(value, 0);
      for (var i = 0; i < operations.length; i++) {
        var op = operations[i];
        if (!op || typeof op !== 'object' || Array.isArray(op)) throw new Error('operation ' + (i + 1) + ' must be an object');
        var kind = op.op;
        if (typeof kind !== 'string') throw new Error('operation ' + (i + 1) + ' requires an "op" string');
        if (typeof op.path !== 'string') throw new Error('operation ' + (i + 1) + ' requires a "path" string');
        var tokens = parsePointer(op.path);
        if (kind === 'add') root = addTokens(root, tokens, requireValue(op, i));
        else if (kind === 'replace') root = replaceTokens(root, tokens, requireValue(op, i));
        else if (kind === 'remove') root = removeTokens(root, tokens);
        else if (kind === 'test') {
          var actual = getTokens(root, tokens, op.path);
          if (!equalValue(actual, op.value, 0)) throw new Error('test failed at "' + op.path + '"');
        } else if (kind === 'copy') {
          if (typeof op.from !== 'string') throw new Error('operation ' + (i + 1) + ' requires a "from" string');
          var copied = cloneValue(getTokens(root, parsePointer(op.from), op.from), 0);
          root = addTokens(root, tokens, copied);
        } else if (kind === 'move') {
          if (typeof op.from !== 'string') throw new Error('operation ' + (i + 1) + ' requires a "from" string');
          if (op.from === op.path) continue;
          var fromTokens = parsePointer(op.from);
          if (fromTokens.length === 0) throw new Error('cannot move the document root');
          if (op.path.indexOf(op.from + '/') === 0) throw new Error('cannot move a value into its own child');
          var moved = getTokens(root, fromTokens, op.from);
          root = removeTokens(root, fromTokens);
          root = addTokens(root, tokens, moved);
        } else throw new Error('operation ' + (i + 1) + ' has unsupported op "' + kind + '"');
      }
      return ok(root);
    } catch (e) { return fail('json.patch: ' + (e && e.message ? e.message : String(e))); }
  }
  function keyMeta(item, keyName) {
    if (!keyName || item === null || typeof item !== 'object' || Array.isArray(item)) return undefined;
    if (!hasOwn(item, keyName)) return undefined;
    var value = item[keyName];
    if (value === null || typeof value === 'object') return undefined;
    return value;
  }
  function itemKey(item, keyName) {
    var value = keyMeta(item, keyName);
    return value === undefined ? null : typeof value + ':' + String(value);
  }
  function diff(a, b, options) {
    var opts = options && typeof options === 'object' ? options : {};
    var keyName = typeof opts.key === 'string' ? opts.key : '';
    var detectMoves = opts.detectMoves !== false;
    var changes = [];
    function push(change) {
      if (changes.length >= 10000) throw new Error('diff is too large (more than 10000 operations)');
      changes.push(change);
    }
    function at(tokens) { return pointerString(tokens); }
    function addOp(tokens, value, keyValue) {
      var change = { op: 'add', path: at(tokens), value: cloneValue(value, 0) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function removeOp(tokens, keyValue) {
      var change = { op: 'remove', path: at(tokens) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function replaceOp(tokens, value, keyValue) {
      var change = { op: 'replace', path: at(tokens), value: cloneValue(value, 0) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function moveOp(fromTokens, toTokens, keyValue) {
      var change = { op: 'move', from: at(fromTokens), path: at(toTokens) };
      if (keyValue !== undefined && keyValue !== null) change.key = keyValue;
      push(change);
    }
    function compare(x, y, tokens) {
      if (equalValue(x, y, 0)) return;
      var xo = x !== null && typeof x === 'object';
      var yo = y !== null && typeof y === 'object';
      if (!xo || !yo || Array.isArray(x) !== Array.isArray(y)) { replaceOp(tokens, y); return; }
      if (Array.isArray(x)) compareArrays(x, y, tokens);
      else compareObjects(x, y, tokens);
    }
    function compareObjects(x, y, tokens) {
      var keysX = Object.keys(x), keysY = Object.keys(y), i, key;
      for (i = 0; i < keysX.length; i++) {
        key = keysX[i];
        if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
        if (!hasOwn(y, key)) removeOp(tokens.concat([key]));
      }
      for (i = 0; i < keysY.length; i++) {
        key = keysY[i];
        if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
        if (!hasOwn(x, key)) addOp(tokens.concat([key]), y[key]);
        else compare(x[key], y[key], tokens.concat([key]));
      }
    }
    function compareArrays(x, y, tokens) {
      if (!keyName) {
        var common = Math.min(x.length, y.length), i;
        for (i = 0; i < common; i++) compare(x[i], y[i], tokens.concat([String(i)]));
        if (x.length > y.length) {
          for (i = x.length - 1; i >= y.length; i--) removeOp(tokens.concat([String(i)]));
        } else if (y.length > x.length) {
          for (i = common; i < y.length; i++) addOp(tokens.concat([String(i)]), y[i]);
        }
        return;
      }
      var queues = Object.create(null), j, k;
      for (j = 0; j < y.length; j++) {
        k = itemKey(y[j], keyName);
        if (k !== null) (queues[k] || (queues[k] = [])).push(j);
      }
      var entries = [], used = [];
      for (j = 0; j < x.length; j++) {
        k = itemKey(x[j], keyName);
        var match = -1;
        if (k !== null && queues[k] && queues[k].length > 0) match = queues[k].shift();
        entries.push({ a: x[j], b: match >= 0 ? y[match] : null, bIndex: match, key: k });
        if (match >= 0) used[match] = true;
      }
      for (var r = entries.length - 1; r >= 0; r--) {
        if (!entries[r].b) {
          removeOp(tokens.concat([String(r)]), keyMeta(entries[r].a, keyName));
          entries.splice(r, 1);
        }
      }
      var reordered = false;
      for (var t = 0; t < y.length; t++) {
        if (used[t]) {
          var current = -1;
          for (var e = 0; e < entries.length; e++) if (entries[e].bIndex === t) { current = e; break; }
          if (current !== t) {
            reordered = true;
            if (!detectMoves) { replaceOp(tokens, y); return; }
            moveOp(tokens.concat([String(current)]), tokens.concat([String(t)]), keyMeta(y[t], keyName));
            var moved = entries.splice(current, 1)[0];
            entries.splice(t, 0, moved);
          }
        } else {
          addOp(tokens.concat([String(t)]), y[t], keyMeta(y[t], keyName));
          entries.splice(t, 0, { a: null, b: y[t], bIndex: t, key: itemKey(y[t], keyName) });
        }
      }
      for (var n = 0; n < entries.length; n++) {
        if (entries[n].a !== null && entries[n].b !== null) compare(entries[n].a, entries[n].b, tokens.concat([String(n)]));
      }
    }
    try {
      compare(a, b, []);
      return ok(changes);
    } catch (e) { return fail('json.diff: ' + (e && e.message ? e.message : String(e))); }
  }
  function merge(target, patchValue) {
    try {
      function apply(current, patchDoc) {
        if (!isObject(patchDoc)) return cloneValue(patchDoc, 0);
        var base = isObject(current) ? cloneValue(current, 0) : {};
        var keys = Object.keys(patchDoc);
        for (var i = 0; i < keys.length; i++) {
          var key = keys[i];
          if (key === '__proto__') throw new Error('unsafe object key "__proto__"');
          if (patchDoc[key] === null) delete base[key];
          else setKey(base, key, apply(base[key], patchDoc[key]));
        }
        return base;
      }
      return ok(apply(target, patchValue));
    } catch (e) { return fail('json.merge: ' + (e && e.message ? e.message : String(e))); }
  }
  function tryGet(root, path) {
    try { return { found: true, value: getTokens(root, parsePointer(path), path) }; }
    catch (e) { return { found: false, value: null }; }
  }
  function replaceInPlace(target, merged) {
    if (Array.isArray(target) && Array.isArray(merged)) {
      target.length = 0;
      for (var i = 0; i < merged.length; i++) target.push(merged[i]);
      return;
    }
    if (isObject(target) && isObject(merged)) {
      var keys = Object.keys(target), j;
      for (j = 0; j < keys.length; j++) delete target[keys[j]];
      keys = Object.keys(merged);
      for (j = 0; j < keys.length; j++) setKey(target, keys[j], merged[keys[j]]);
      return;
    }
    throw new Error('requires object or array roots that keep their type');
  }
  function merge3(base, ours, theirs, options) {
    var changes = diff(base, theirs, options);
    if (changes[1] !== null) return [[], changes[1]];
    var conflicts = [], safe = [];
    for (var i = 0; i < changes[0].length; i++) {
      var op = changes[0][i];
      var mine = tryGet(ours, op.path);
      var before = tryGet(base, op.path);
      var conflict = false;
      if (op.op === 'replace') {
        if (!mine.found) conflict = before.found;
        else if (equalValue(mine.value, op.value, 0)) continue;
        else if (before.found && equalValue(mine.value, before.value, 0)) conflict = false;
        else conflict = true;
      } else if (op.op === 'remove') {
        if (!mine.found) continue;
        if (before.found && equalValue(mine.value, before.value, 0)) conflict = false;
        else conflict = true;
      } else if (op.op === 'add') {
        if (mine.found) {
          if (equalValue(mine.value, op.value, 0)) continue;
          conflict = true;
        }
      } else if (op.op === 'move') {
        var mineFrom = tryGet(ours, op.from);
        var baseFrom = tryGet(base, op.from);
        if (!mineFrom.found) conflict = true;
        else if (baseFrom.found && !equalValue(mineFrom.value, baseFrom.value, 0)) conflict = true;
      } else if (op.op === 'test') {
        if (mine.found && equalValue(mine.value, op.value, 0)) continue;
        conflict = true;
      }
      if (conflict) conflicts.push(op.op + ' at ' + op.path);
      else safe.push(op);
    }
    var applied = patch(ours, safe);
    if (applied[1] !== null) return [[], applied[1]];
    try {
      if (!equalValue(applied[0], ours, 0)) replaceInPlace(ours, applied[0]);
    } catch (e) { return [[], 'json.merge3: ' + (e && e.message ? e.message : String(e))]; }
    return ok(conflicts);
  }
  function parse(text) {
    try {
      if (typeof text !== 'string') throw new Error('json.parse expects a string');
      if (text.length > MAX_BYTES) throw new Error('JSON text is too large (max ' + MAX_BYTES + ' characters)');
      if (text.charCodeAt(0) === 0xFEFF) text = text.slice(1);
      var value = JSON.parse(text);
      return ok(value);
    } catch (e) { return fail('json.parse: ' + (e && e.message ? e.message : String(e))); }
  }
  function decodeBytes(bytes, encoding) {
    if (!bytes || typeof bytes.length !== 'number') throw new Error('expected a byte array');
    if (bytes.length > MAX_BYTES) throw new Error('JSON file is too large (max ' + MAX_BYTES + ' bytes)');
    var data = new Uint8Array(bytes.length);
    for (var i = 0; i < bytes.length; i++) {
      var value = Number(bytes[i]);
      if (!isFinite(value) || value < 0 || value > 255) throw new Error('invalid byte value at index ' + i);
      data[i] = value;
    }
    if (data.length >= 4 && data[0] === 0x00 && data[1] === 0x00 && data[2] === 0xFE && data[3] === 0xFF) throw new Error('UTF-32BE JSON is not supported');
    if (data.length >= 4 && data[0] === 0xFF && data[1] === 0xFE && data[2] === 0x00 && data[3] === 0x00) throw new Error('UTF-32LE JSON is not supported');
    var label = encoding ? String(encoding).toLowerCase().replace(/_/g, '-') : '';
    if (!label) {
      if (data.length >= 3 && data[0] === 0xEF && data[1] === 0xBB && data[2] === 0xBF) label = 'utf-8';
      else if (data.length >= 2 && data[0] === 0xFF && data[1] === 0xFE) label = 'utf-16le';
      else if (data.length >= 2 && data[0] === 0xFE && data[1] === 0xFF) label = 'utf-16be';
      else label = 'utf-8';
    }
    if (label === 'utf8') label = 'utf-8';
    if (label === 'utf16le') label = 'utf-16le';
    if (label === 'utf16be') label = 'utf-16be';
    if (label !== 'utf-8' && label !== 'utf-16le' && label !== 'utf-16be') throw new Error('unsupported JSON encoding "' + label + '"');
    if (label === 'utf-16le' || label === 'utf-16be') {
      var little = label === 'utf-16le';
      var offset = 0;
      if (data.length >= 2 && ((little && data[0] === 0xFF && data[1] === 0xFE) || (!little && data[0] === 0xFE && data[1] === 0xFF))) offset = 2;
      var text = '';
      for (var j = offset; j + 1 < data.length; j += 2) {
        var code = little ? data[j] | (data[j + 1] << 8) : (data[j] << 8) | data[j + 1];
        text += String.fromCharCode(code);
      }
      return text;
    }
    if (typeof TextDecoder === 'undefined') throw new Error('TextDecoder is not available');
    return new TextDecoder('utf-8').decode(data).replace(/^\uFEFF/, '');
  }
  function parseBytes(bytes, encoding) {
    try {
      var text = decodeBytes(bytes, encoding);
      return parse(text);
    } catch (e) { return fail('json.parseBytes: ' + (e && e.message ? e.message : String(e))); }
  }
  function encodeBytes(text, encoding) {
    var label = encoding ? String(encoding).toLowerCase().replace(/_/g, '-') : 'utf-8';
    if (label === 'utf8') label = 'utf-8';
    if (label === 'utf16le') label = 'utf-16le';
    if (label === 'utf16be') label = 'utf-16be';
    if (label === 'utf-8') {
      if (typeof TextEncoder === 'undefined') throw new Error('TextEncoder is not available');
      return Array.prototype.slice.call(new TextEncoder().encode(text));
    }
    if (label !== 'utf-16le' && label !== 'utf-16be') throw new Error('unsupported JSON encoding "' + label + '"');
    var little = label === 'utf-16le', out = [0xFF, 0xFE];
    if (!little) out = [0xFE, 0xFF];
    for (var i = 0; i < text.length; i++) {
      var code = text.charCodeAt(i);
      if (little) { out.push(code & 255, (code >> 8) & 255); }
      else { out.push((code >> 8) & 255, code & 255); }
    }
    return out;
  }
  function toBytes(value, encoding, options) {
    var result = stringify(value, options);
    if (result[1] !== null) return result;
    try { return ok(encodeBytes(result[0], encoding || 'utf-8')); }
    catch (e) { return fail('json.toBytes: ' + (e && e.message ? e.message : String(e))); }
  }
  function stringify(value, options) {
    try {
      var opts = options && typeof options === 'object' ? options : {};
      var indent = opts.indent;
      if (indent === true) indent = 2;
      if (typeof indent === 'number') indent = Math.max(0, Math.min(10, Math.floor(indent)));
      else if (typeof indent === 'string') indent = indent.slice(0, 10);
      else indent = 0;
      var valueToWrite = value;
      if (opts.sortKeys) {
        function sort(v) {
          if (Array.isArray(v)) return v.map(sort);
          if (!isObject(v)) return v;
          var out = {}, keys = Object.keys(v).sort();
          for (var i = 0; i < keys.length; i++) setKey(out, keys[i], sort(v[keys[i]]));
          return out;
        }
        valueToWrite = sort(value);
      }
      var text = JSON.stringify(valueToWrite, null, indent || undefined);
      if (typeof text !== 'string') throw new Error('value is not JSON serializable');
      if (opts.newline === true) text += '\n';
      return ok(text);
    } catch (e) { return fail('json.stringify: ' + (e && e.message ? e.message : String(e))); }
  }
  function validateAll(value, schema) {
    var errors = [];
    function add(path, message) { if (errors.length < MAX_ERRORS) errors.push(path + ': ' + message); }
    try { if (typeof schema === 'string') schema = JSON.parse(schema); }
    catch (e) { return [[], '$: schema is not valid JSON']; }
    function check(v, s, path, depth) {
      if (errors.length >= MAX_ERRORS) return;
      if (depth > 64) { add(path, 'schema nesting is too deep'); return; }
      if (s === true) return;
      if (s === false) { add(path, 'rejected by false schema'); return; }
      if (s === null || typeof s !== 'object' || Array.isArray(s)) { add(path, 'schema must be an object or boolean'); return; }
      if (s.$ref !== undefined) { add(path, '$ref is not supported'); return; }
      var i, key;
      if (s.type !== undefined) {
        var types = Array.isArray(s.type) ? s.type : [s.type], typeOk = false;
        for (i = 0; i < types.length; i++) if (__sui_schema_type_ok(v, types[i])) { typeOk = true; break; }
        if (!typeOk) add(path, 'expected ' + types.join(' or ') + ', got ' + __sui_schema_kind(v));
      }
      if (s.enum !== undefined) {
        if (!Array.isArray(s.enum)) add(path, 'enum must be an array');
        else {
          var found = false;
          for (i = 0; i < s.enum.length; i++) if (__sui_schema_equal(v, s.enum[i])) { found = true; break; }
          if (!found) add(path, 'value is not in enum');
        }
      }
      if (hasOwn(s, 'const') && !__sui_schema_equal(v, s.const)) add(path, 'value does not match const');
      if (typeof v === 'number') {
        if (typeof s.minimum === 'number' && v < s.minimum) add(path, 'must be >= ' + s.minimum);
        if (typeof s.maximum === 'number' && v > s.maximum) add(path, 'must be <= ' + s.maximum);
        if (typeof s.exclusiveMinimum === 'number' && v <= s.exclusiveMinimum) add(path, 'must be > ' + s.exclusiveMinimum);
        if (typeof s.exclusiveMaximum === 'number' && v >= s.exclusiveMaximum) add(path, 'must be < ' + s.exclusiveMaximum);
        if (typeof s.multipleOf === 'number' && s.multipleOf > 0) {
          var q = v / s.multipleOf;
          if (Math.abs(q - Math.round(q)) > 1e-9) add(path, 'must be a multiple of ' + s.multipleOf);
        }
      }
      if (typeof v === 'string') {
        if (typeof s.minLength === 'number' && v.length < s.minLength) add(path, 'must have at least ' + s.minLength + ' characters');
        if (typeof s.maxLength === 'number' && v.length > s.maxLength) add(path, 'must have at most ' + s.maxLength + ' characters');
        if (typeof s.pattern === 'string') {
          if (s.pattern.length > 512) add(path, 'pattern is too long');
          else { try { if (!(new RegExp(s.pattern)).test(v)) add(path, 'does not match the pattern'); } catch (e) { add(path, 'invalid pattern'); } }
        }
        if (typeof s.format === 'string') {
          if (s.format === 'email' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) add(path, 'must be an email');
          if ((s.format === 'uri' || s.format === 'url') && !__sui_valid_url(v)) add(path, 'must be a URI');
          if (s.format === 'uuid' && !/^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$/.test(v)) add(path, 'must be a UUID');
          if (s.format === 'date' && !/^\d{4}-\d{2}-\d{2}$/.test(v)) add(path, 'must be a date');
          if (s.format === 'date-time' && isNaN(Date.parse(v))) add(path, 'must be a date-time');
        }
      }
      if (Array.isArray(v)) {
        if (typeof s.minItems === 'number' && v.length < s.minItems) add(path, 'must have at least ' + s.minItems + ' items');
        if (typeof s.maxItems === 'number' && v.length > s.maxItems) add(path, 'must have at most ' + s.maxItems + ' items');
        if (s.uniqueItems === true) {
          var seen = {}, duplicates = 0;
          for (i = 0; i < v.length; i++) {
            var ukey = JSON.stringify(v[i]);
            if (seen[ukey]) { duplicates++; break; }
            seen[ukey] = true;
          }
          if (duplicates) add(path, 'items must be unique');
        }
        if (Array.isArray(s.items)) {
          for (i = 0; i < s.items.length && i < v.length; i++) check(v[i], s.items[i], path + '[' + i + ']', depth + 1);
          if (v.length > s.items.length) {
            if (s.additionalItems === false) add(path, 'additional items are not allowed');
            else if (s.additionalItems && typeof s.additionalItems === 'object') {
              for (i = s.items.length; i < v.length; i++) check(v[i], s.additionalItems, path + '[' + i + ']', depth + 1);
            }
          }
        } else if (s.items !== undefined) {
          for (i = 0; i < v.length; i++) check(v[i], s.items, path + '[' + i + ']', depth + 1);
        }
      } else if (isObject(v)) {
        var keys = Object.keys(v);
        if (typeof s.minProperties === 'number' && keys.length < s.minProperties) add(path, 'must have at least ' + s.minProperties + ' properties');
        if (typeof s.maxProperties === 'number' && keys.length > s.maxProperties) add(path, 'must have at most ' + s.maxProperties + ' properties');
        if (Array.isArray(s.required)) {
          for (i = 0; i < s.required.length; i++) if (!hasOwn(v, s.required[i])) add(path, 'missing required property "' + s.required[i] + '"');
        }
        var props = s.properties && typeof s.properties === 'object' ? s.properties : {};
        for (i = 0; i < keys.length; i++) {
          key = keys[i];
          if (props[key] !== undefined) check(v[key], props[key], path + '.' + key, depth + 1);
          else if (s.additionalProperties === false) add(path, 'unexpected property "' + key + '"');
          else if (s.additionalProperties && typeof s.additionalProperties === 'object') check(v[key], s.additionalProperties, path + '.' + key, depth + 1);
        }
      }
      if (Array.isArray(s.allOf)) for (i = 0; i < s.allOf.length; i++) check(v, s.allOf[i], path, depth + 1);
      if (Array.isArray(s.anyOf)) {
        var anyOk = false;
        for (i = 0; i < s.anyOf.length; i++) if (__sui_schema_check(v, s.anyOf[i], path, depth + 1) === null) { anyOk = true; break; }
        if (!anyOk) add(path, 'does not match anyOf');
      }
      if (Array.isArray(s.oneOf)) {
        var oneCount = 0;
        for (i = 0; i < s.oneOf.length; i++) if (__sui_schema_check(v, s.oneOf[i], path, depth + 1) === null) oneCount++;
        if (oneCount !== 1) add(path, 'must match exactly one oneOf schema');
      }
      if (s.not !== undefined && __sui_schema_check(v, s.not, path, depth + 1) === null) add(path, 'must not match the not schema');
      if (s.if !== undefined) {
        var condition = __sui_schema_check(v, s.if, path, depth + 1);
        if (condition === null && s.then !== undefined) check(v, s.then, path, depth + 1);
        else if (condition !== null && s.else !== undefined) check(v, s.else, path, depth + 1);
      }
    }
    check(value, schema, '$', 0);
    return ok(errors);
  }
  return {
    parse: parse,
    parseBytes: parseBytes,
    decodeBytes: decodeBytes,
    encodeBytes: encodeBytes,
    stringify: stringify,
    toBytes: toBytes,
    validateAll: validateAll,
    pointer: pointer,
    setPointer: setPointer,
    removePointer: removePointer,
    clone: clone,
    equal: equal,
    patch: patch,
    diff: diff,
    merge: merge,
    merge3: merge3
  };
})();


const __SUI_STD = {
  http_get: async (url) => { try { const r = await fetch(url); return [await r.text(), null]; } catch (e) { return [null, String(e)]; } },
  http_getstatus: async (url) => { try { const r = await fetch(url); return [r.status, null]; } catch (e) { return [null, String(e)]; } },
  http_getbytes: async (url) => { try { const r = await fetch(url); if (!r.ok) return [null, 'http error ' + r.status]; const buf = await r.arrayBuffer(); if (buf.byteLength > 10485760) return [null, 'response too large']; const v = new Uint8Array(buf); const out = new Array(v.length); for (var i = 0; i < v.length; i++) out[i] = v[i]; return [out, null]; } catch (e) { return [null, String(e)]; } },
  http_getbytessafe: async (url, allow) => __sui_fetch_bytes_guarded(url, allow),
  http_post: async (url, data) => { try { const r = await fetch(url, {method: "POST", headers: {"Content-Type": "application/json"}, body: data}); return [await r.text(), null]; } catch (e) { return [null, String(e)]; } },
  http_postfiles: async (url) => { const fd = __sui_files_form(); if (!fd) return [null, 'no files selected']; try { const abs = new URL(url, location.origin).href; return __sui_fetch_guarded(abs, abs, {method: 'POST', body: fd}); } catch (e) { return [null, String(e)]; } },
  http_postfilessafe: async (url, allow) => { const fd = __sui_files_form(); if (!fd) return [null, 'no files selected']; return __sui_fetch_guarded(url, allow, {method: 'POST', body: fd}); },
  http_getsafe: async (url, allow) => __sui_fetch_guarded(url, allow, {}),
  http_getstatussafe: async (url, allow) => __sui_fetch_status_guarded(url, allow),
  http_postsafe: async (url, data, allow) => __sui_fetch_guarded(url, allow, {method: "POST", headers: {"Content-Type": "application/json"}, body: data}),
  http_getwithheaders: async (url, headers) => { try { const abs = new URL(url, location.origin).href; return __sui_fetch_guarded(abs, abs, {method: 'GET', headers: __sui_clean_headers(headers)}); } catch (e) { return [null, String(e)]; } },
  http_getwithheaderssafe: async (url, headers, allow) => __sui_fetch_guarded(url, allow, {method: 'GET', headers: __sui_clean_headers(headers)}),
  http_postwithheaders: async (url, data, headers) => { try { const abs = new URL(url, location.origin).href; const h = Object.assign({"Content-Type": "application/json"}, __sui_clean_headers(headers)); return __sui_fetch_guarded(abs, abs, {method: 'POST', headers: h, body: data}); } catch (e) { return [null, String(e)]; } },
  http_postwithheaderssafe: async (url, data, headers, allow) => { const h = Object.assign({"Content-Type": "application/json"}, __sui_clean_headers(headers)); return __sui_fetch_guarded(url, allow, {method: 'POST', headers: h, body: data}); },
  http_request: async (method, url, headers, data, timeoutMs) => { try { const abs = new URL(url, location.origin).href; const m = String(method || 'GET').toUpperCase(); const init = {method: m, headers: __sui_clean_headers(headers)}; if (m !== 'GET' && m !== 'HEAD' && data !== undefined && data !== null && data !== '') init.body = data; return __sui_fetch_guarded(abs, abs, init, timeoutMs); } catch (e) { return [null, String(e)]; } },
  http_requestsafe: async (method, url, headers, data, timeoutMs, allow) => { const m = String(method || 'GET').toUpperCase(); const init = {method: m, headers: __sui_clean_headers(headers)}; if (m !== 'GET' && m !== 'HEAD' && data !== undefined && data !== null && data !== '') init.body = data; return __sui_fetch_guarded(url, allow, init, timeoutMs); },
  http_upload: async (url, fieldName, fileName, content) => { try { const abs = new URL(url, location.origin).href; const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(abs, abs, {method: 'POST', body: fd}); } catch (e) { return [null, String(e)]; } },
  http_uploadsafe: async (url, fieldName, fileName, content, allow) => { const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(url, allow, {method: 'POST', body: fd}); },
  http_uploadwithheaders: async (url, fieldName, fileName, content, headers) => { try { const abs = new URL(url, location.origin).href; const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(abs, abs, {method: 'POST', headers: __sui_clean_headers(headers), body: fd}); } catch (e) { return [null, String(e)]; } },
  http_uploadwithheaderssafe: async (url, fieldName, fileName, content, headers, allow) => { const fd = __sui_upload_form(fieldName, fileName, content); return __sui_fetch_guarded(url, allow, {method: 'POST', headers: __sui_clean_headers(headers), body: fd}); },
  sound_play: async (url) => __sui_play_sound(url, null),
  sound_playsafe: async (url, allow) => __sui_play_sound(url, allow),
  json_parse: async (s) => __SUI_JSON.parse(s),
  json_parse_bytes: (bytes, encoding) => __SUI_JSON.parseBytes(bytes, encoding),
  json_stringify: async (v, options) => __SUI_JSON.stringify(v, options),
  json_to_bytes: (value, encoding, options) => __SUI_JSON.toBytes(value, encoding, options),
  json_validate: (value, schema) => __sui_json_validate(value, schema),
  json_validate_all: (value, schema) => __SUI_JSON.validateAll(value, schema),
  json_pointer: (value, path) => __SUI_JSON.pointer(value, path),
  json_set_pointer: (value, path, next) => __SUI_JSON.setPointer(value, path, next),
  json_remove_pointer: (value, path) => __SUI_JSON.removePointer(value, path),
  json_clone: (value) => __SUI_JSON.clone(value),
  json_equal: (a, b) => __SUI_JSON.equal(a, b),
  json_diff: (a, b, options) => __SUI_JSON.diff(a, b, options),
  json_patch: (value, changes) => __SUI_JSON.patch(value, changes),
  json_merge: (value, patch) => __SUI_JSON.merge(value, patch),
  json_merge3: (base, ours, theirs, options) => __SUI_JSON.merge3(base, ours, theirs, options),
  json_load_file: async (path, encoding) => [null, "json.loadFile is not available on the web target; use fs.read or fetch and json.parse"],
  json_load_bytes: async (path, encoding) => [null, "json.loadBytes is not available on the web target"],
  json_save_file: async (path, value, options) => [null, "json.saveFile is not available on the web target"],
  json_save_bytes: async (path, value, encoding, options) => [null, "json.saveBytes is not available on the web target"],
  richtext_get: (name) => __SUI_RICH.get(name),
  richtext_set: (name, value) => __SUI_RICH.set(name, value),
  richtext_loadcontent: async (name, content, fmt, filename) => __SUI_RICH.loadContent(name, content, fmt, filename),
  richtext_command: (name, cmd, value) => __SUI_RICH.command(name, cmd, value),
  richtext_format: (name, fmt, value) => __SUI_RICH.format(name, fmt, value),
  richtext_inserttext: (name, text) => __SUI_RICH.insertText(name, text),
  richtext_insertimage: (name, src, alt) => __SUI_RICH.insertImage(name, src, alt),
  richtext_insertlink: (name, href, text) => __SUI_RICH.insertLink(name, href, text),
  richtext_inserttable: (name, rows, cols) => __SUI_RICH.insertTable(name, rows, cols),
  richtext_undo: (name) => __SUI_RICH.undo(name),
  richtext_redo: (name) => __SUI_RICH.redo(name),
  richtext_focus: (name) => __SUI_RICH.focus(name),
  richtext_export: (name, fmt) => __SUI_RICH.exportFormat(name, fmt),
  richtext_exportbytes: (name, fmt) => __SUI_RICH.exportBytes(name, fmt),
  richtext_openfile: async (name) => __SUI_RICH.openFile(name),
  richtext_importbytes: async (name, bytes, filename) => __SUI_RICH.importBytes(name, bytes, filename),
  richtext_saveas: (name, fmt) => __SUI_RICH.saveAs(name, fmt),
  richtext_import: (name, text, fmt) => __SUI_RICH.importFormat(name, text, fmt),
  richtext_register: (name, plugin) => __SUI_RICH.registerPlugin(name, plugin),
  richtext_plugins: () => __SUI_RICH.plugins(),
  richtext_status: (name) => __SUI_RICH.status(name),
  richtext_selection: (name) => __SUI_RICH.selection(name),
  richtext_settoolbar: (name, items) => __SUI_RICH.setToolbar(name, items),
  richtext_settemplate: (name, template) => __SUI_RICH.setTemplate(name, template),
  richtext_settheme: (name, theme) => __SUI_RICH.setTheme(name, theme),
  richtext_destroy: (name) => __SUI_RICH.destroy(name),
  table_display: (name, data) => __SUI_TBL.display(name, data),
  table_data: (name) => __SUI_TBL.data(name),
  table_header: (name) => __SUI_TBL.header(name),
  table_select: (name, cols) => __SUI_TBL.select(name, cols),
  table_sort: (name, column, dir) => __SUI_TBL.sort(name, column, dir),
  table_filter: (name, column, value) => __SUI_TBL.filter(name, column, value),
  table_filtertext: (name, text) => __SUI_TBL.filterText(name, text),
  table_page: (name, page) => __SUI_TBL.page(name, page),
  table_setpagesize: (name, size) => __SUI_TBL.setPageSize(name, size),
  table_selectedrows: (name) => __SUI_TBL.selectedRows(name),
  table_selectedcells: (name) => __SUI_TBL.selectedCells(name),
  table_selectedcolumns: (name) => __SUI_TBL.selectedColumns(name),
  table_lastchange: (name) => __SUI_TBL.lastChange(name),
  table_slice: (name, from, count) => __SUI_TBL.slice(name, from, count),
  table_importtext: (name, text, spec) => __SUI_TBL.importText(name, text, spec),
  table_exporttext: (name, spec) => __SUI_TBL.exportText(name, spec),
  table_json: (name) => __SUI_TBL.json(name),
  table_clear: (name) => __SUI_TBL.clear(name),
  table_clearpersisted: (name) => __SUI_TBL.clearPersisted(name),
  chart_settag: (name, id, text, color, textColor) => __SUI_CHART.setTag(name, id, text, color, textColor),
  chart_update: (name, id, patch) => __SUI_CHART.update(name, id, patch),
  chart_updatemany: (name, updates) => __SUI_CHART.updateMany(name, updates),
  chart_get: (name, id) => __SUI_CHART.get(name, id),
  chart_elements: (name) => __SUI_CHART.elements(name),
  chart_find: (name, query) => __SUI_CHART.find(name, query),
  chart_getattr: (name, id, field) => __SUI_CHART.getAttr(name, id, field),
  chart_setattr: (name, id, field, value) => __SUI_CHART.setAttr(name, id, field, value),
  chart_getlabel: (name, id) => __SUI_CHART.getAttr(name, id, "label"),
  chart_setlabel: (name, id, value) => __SUI_CHART.setAttr(name, id, "label", value),
  chart_gettext: (name, id) => __SUI_CHART.getAttr(name, id, "tag"),
  chart_settext: (name, id, value) => __SUI_CHART.setAttr(name, id, "tag", value),
  chart_gethint: (name, id) => __SUI_CHART.getAttr(name, id, "hint"),
  chart_sethint: (name, id, value) => __SUI_CHART.setAttr(name, id, "hint", value),
  chart_getdescription: (name, id) => __SUI_CHART.getAttr(name, id, "description"),
  chart_setdescription: (name, id, value) => __SUI_CHART.setAttr(name, id, "description", value),
  chart_getcolor: (name, id) => __SUI_CHART.getAttr(name, id, "color"),
  chart_setcolor: (name, id, value) => __SUI_CHART.setAttr(name, id, "color", value),
  chart_gettagcolor: (name, id) => __SUI_CHART.getAttr(name, id, "tagBoxColor"),
  chart_settagcolor: (name, id, value) => __SUI_CHART.setAttr(name, id, "tagBoxColor", value),
  chart_gettextcolor: (name, id) => __SUI_CHART.getAttr(name, id, "tagTextColor"),
  chart_settextcolor: (name, id, value) => __SUI_CHART.setAttr(name, id, "tagTextColor", value),
  chart_getshape: (name, id) => __SUI_CHART.getAttr(name, id, "shape"),
  chart_setshape: (name, id, value) => __SUI_CHART.setAttr(name, id, "shape", value),
  chart_getrouting: (name, id) => __SUI_CHART.getAttr(name, id, "routing"),
  chart_setrouting: (name, id, value) => __SUI_CHART.setAttr(name, id, "routing", value),
  chart_getbordercolor: (name, id) => __SUI_CHART.getAttr(name, id, "borderColor"),
  chart_setbordercolor: (name, id, value) => __SUI_CHART.setAttr(name, id, "borderColor", value),
  chart_getborderstyle: (name, id) => __SUI_CHART.getAttr(name, id, "borderStyle"),
  chart_setborderstyle: (name, id, value) => __SUI_CHART.setAttr(name, id, "borderStyle", value),
  chart_getborderwidth: (name, id) => __SUI_CHART.getAttr(name, id, "borderWidth"),
  chart_setborderwidth: (name, id, value) => __SUI_CHART.setAttr(name, id, "borderWidth", value),
  chart_getcornerradius: (name, id) => __SUI_CHART.getAttr(name, id, "cornerRadius"),
  chart_setcornerradius: (name, id, value) => __SUI_CHART.setAttr(name, id, "cornerRadius", value),
  chart_getwidth: (name, id) => __SUI_CHART.getAttr(name, id, "width"),
  chart_setwidth: (name, id, value) => __SUI_CHART.setAttr(name, id, "width", value),
  chart_getheight: (name, id) => __SUI_CHART.getAttr(name, id, "height"),
  chart_setheight: (name, id, value) => __SUI_CHART.setAttr(name, id, "height", value),
  chart_getx: (name, id) => __SUI_CHART.getAttr(name, id, "x"),
  chart_setx: (name, id, value) => __SUI_CHART.setAttr(name, id, "x", value),
  chart_gety: (name, id) => __SUI_CHART.getAttr(name, id, "y"),
  chart_sety: (name, id, value) => __SUI_CHART.setAttr(name, id, "y", value),
  chart_getrotation: (name, id) => __SUI_CHART.getAttr(name, id, "rotation"),
  chart_setrotation: (name, id, value) => __SUI_CHART.setAttr(name, id, "rotation", value),
  chart_getdashstyle: (name, id) => __SUI_CHART.getAttr(name, id, "dashStyle"),
  chart_setdashstyle: (name, id, value) => __SUI_CHART.setAttr(name, id, "dashStyle", value),
  chart_getstartmarker: (name, id) => __SUI_CHART.getAttr(name, id, "startMarker"),
  chart_setstartmarker: (name, id, value) => __SUI_CHART.setAttr(name, id, "startMarker", value),
  chart_getendmarker: (name, id) => __SUI_CHART.getAttr(name, id, "endMarker"),
  chart_setendmarker: (name, id, value) => __SUI_CHART.setAttr(name, id, "endMarker", value),
  chart_setfrom: (name, id, value) => __SUI_CHART.setAttr(name, id, "from", value),
  chart_setto: (name, id, value) => __SUI_CHART.setAttr(name, id, "to", value),
  chart_gettagkind: (name, id) => __SUI_CHART.getAttr(name, id, "tagKind"),
  chart_settagkind: (name, id, value) => __SUI_CHART.setAttr(name, id, "tagKind", value),
  chart_gettagbox: (name, id) => __SUI_CHART.getAttr(name, id, "tagBox"),
  chart_settagbox: (name, id, value) => __SUI_CHART.setAttr(name, id, "tagBox", value),
  chart_gettagalign: (name, id) => __SUI_CHART.getAttr(name, id, "tagAlign"),
  chart_settagalign: (name, id, value) => __SUI_CHART.setAttr(name, id, "tagAlign", value),
  chart_gettagleader: (name, id) => __SUI_CHART.getAttr(name, id, "tagLeader"),
  chart_settagleader: (name, id, value) => __SUI_CHART.setAttr(name, id, "tagLeader", value),
  chart_getwaypoints: (name, id) => __SUI_CHART.getAttr(name, id, "waypoints"),
  chart_setwaypoints: (name, id, value) => __SUI_CHART.setAttr(name, id, "waypoints", value),
  chart_gettextplacement: (name, id) => __SUI_CHART.getAttr(name, id, "textPlacement"),
  chart_settextplacement: (name, id, value) => __SUI_CHART.setAttr(name, id, "textPlacement", value),
  chart_gettagside: (name, id) => __SUI_CHART.getAttr(name, id, "tagSide"),
  chart_settagside: (name, id, value) => __SUI_CHART.setAttr(name, id, "tagSide", value),
  chart_getborderdashlength: (name, id) => __SUI_CHART.getAttr(name, id, "borderDashLength"),
  chart_setborderdashlength: (name, id, value) => __SUI_CHART.setAttr(name, id, "borderDashLength", value),
  chart_getborderdashgap: (name, id) => __SUI_CHART.getAttr(name, id, "borderDashGap"),
  chart_setborderdashgap: (name, id, value) => __SUI_CHART.setAttr(name, id, "borderDashGap", value),
  chart_getsnap: (name, id) => __SUI_CHART.getAttr(name, id, "snap"),
  chart_setsnap: (name, id, value) => __SUI_CHART.setAttr(name, id, "snap", value),
  chart_getkind: (name, id) => __SUI_CHART.getAttr(name, id, "kind"),
  chart_getfrom: (name, id) => __SUI_CHART.getAttr(name, id, "from"),
  chart_getto: (name, id) => __SUI_CHART.getAttr(name, id, "to"),
  chart_focusrow: (name, id) => __SUI_CHART.focusRow(name, id),
  chart_peekevents: (name) => __SUI_CHART.peekEvents(name),
  chart_pollevents: (name) => __SUI_CHART.pollEvents(name),
  chart_nextevent: (name) => __SUI_CHART.nextEvent(name),
  chart_clearevents: (name) => __SUI_CHART.clearEvents(name),
  chart_clearpersisted: (name) => __SUI_CHART.clearPersisted(name),
  chart_eventcount: (name) => __SUI_CHART.eventCount(name),
  tree_importjson: (name, text) => __SUI_TREE.importJson(name, text),
  tree_importxml: (name, text) => __SUI_TREE.importXml(name, text),
  tree_importjsonfile: async (name, path) => __SUI_TREE.importJsonFile(name, path),
  tree_importxmlfile: async (name, path) => __SUI_TREE.importXmlFile(name, path),
  tree_exportjson: (name) => __SUI_TREE.exportJson(name),
  tree_exportxml: (name) => __SUI_TREE.exportXml(name),
  tree_downloadjson: (name, filename) => __SUI_TREE.downloadJson(name, filename),
  tree_downloadxml: (name, filename) => __SUI_TREE.downloadXml(name, filename),
  tree_clear: (name) => __SUI_TREE.clear(name),
  tree_clearpersisted: (name) => __SUI_TREE.clearPersisted(name),
  tree_expandall: (name) => __SUI_TREE.expandAll(name),
  tree_collapseall: (name) => __SUI_TREE.collapseAll(name),
  tree_expandto: (name, level) => __SUI_TREE.expandTo(name, level),
  tree_setchildren: (name, key, text) => __SUI_TREE.setChildren(name, key, text),
  tree_setchildrenerror: (name, key, message) => __SUI_TREE.setChildrenError(name, key, message),
  tree_retrychildren: (name, key) => __SUI_TREE.retryChildren(name, key),
  tree_lastmove: (name) => __SUI_TREE.lastMove(name),
  tree_lastchange: (name) => __SUI_TREE.lastChange(name),
  tree_undomove: (name) => __SUI_TREE.undoMove(name),
  tree_setcolumns: (name, json) => __SUI_TREE.setColumns(name, json),
  tree_getcolumns: (name) => __SUI_TREE.getColumns(name),
  tree_movecolumn: (name, from, to) => __SUI_TREE.moveColumn(name, from, to),
  tree_setcell: (name, key, column, value) => __SUI_TREE.setCell(name, key, column, value),
  tree_selectedkeys: (name) => __SUI_TREE.selectedKeys(name),
  tree_checkedkeys: (name) => __SUI_TREE.checkedKeys(name),
  tree_search: (name, term) => __SUI_TREE.search(name, term)
};

function render_main() {
  const el = document.getElementById("app");
  const __viewHtml = `<div class="sui-view" data-sui-view="main">`
    + `<div class="sui-toast-host sui-toast-h-right sui-toast-v-top toast-host" data-sui-toast-host="1" aria-live="polite" aria-atomic="false" data-sui-toast-x="20px" data-sui-toast-y="20px" data-sui-id="0"></div>` + `<div class="sui-modal" data-sui-modal="settings" aria-hidden="true" hidden data-sui-id="1"><div class="sui-modal-backdrop" data-sui-modal-backdrop></div><div class="sui-modal-panel" role="dialog" aria-modal="true" tabindex="-1" aria-label="Download settings"><div class="sui-modal-title">Download settings</div><div class="sui-modal-body">` + (`<div class="settings" data-sui-id="2">` + (`<div class="row setting" data-sui-id="3">` + (`<label data-sui-id="4" class="sui-switch-wrap mode-switch">` + (`<input type="checkbox" class="sui-switch mode-switch" ${(single) ? "checked" : ""}>` + (`<span>` + `Download the file in one piece` + `</span>`)) + `</label>` + `<div class="col setting-text" data-sui-id="5">` + (`<p class="hint" data-sui-id="6">On: one request for the whole file, no parts and no reassembly.</p>` + `<p class="hint" data-sui-id="7">Off: the file is split into parts downloaded in parallel.</p>`) + `</div>`) + `</div>` + ((single === false) ? `<div class="row setting" data-sui-id="8">` + (`<div class="sui-combobox setting-control" data-sui-id="9" data-sui-combobox-placement="auto" data-sui-animation="fade" data-sui-combobox="1" data-sui-options="256 KiB,512 KiB,1 MiB,2 MiB,4 MiB,8 MiB,16 MiB" data-sui-placeholder="Search..." data-sui-empty="No results" data-sui-combobox-setter="setPartSize"><label class="sui-combobox-label">Part size</label><div class="sui-combobox-control"><input type="text" class="sui-combobox-input" role="combobox" autocomplete="off" spellcheck="false" aria-expanded="false" placeholder="Search..."><button type="button" class="sui-combobox-clear" aria-label="Clear selection" hidden>×</button><button type="button" class="sui-combobox-toggle" aria-label="Open options" aria-expanded="false">▾</button></div><input type="hidden" class="sui-combobox-value" value="${__sui_esc(partSize)}"><div class="sui-combobox-popup" role="listbox" hidden></div><div class="sui-combobox-error" role="alert" hidden></div></div>` + `<div class="col setting-text" data-sui-id="10">` + (`<p class="hint" data-sui-id="11">The file is split into parts of this size.</p>`) + `</div>`) + `</div>` + `<div class="row setting" data-sui-id="12">` + (`<input type="number" class="sui-number setting-control" value="${__sui_esc(workers)}" min="1" max="32" step="1" data-sui-id="13">` + `<div class="col setting-text" data-sui-id="14">` + (`<p class="hint" data-sui-id="15">How many parts are downloaded at the same time.</p>`) + `</div>`) + `</div>` : '') + ((single === true) ? `<p class="hint" data-sui-id="16">Now: the file is downloaded in one piece.</p>` : `<p class="hint" data-sui-id="17">Now: parts of ${__sui_esc(partSize)}, ${__sui_esc(workers)} at a time.</p>`)) + `</div>`) + `</div><div class="sui-modal-actions"><button type="button" class="sui-modal-button" data-sui-modal-dismiss="1">Close</button><button type="button" class="sui-modal-button primary" data-sui-modal-dismiss="1" onclick="closeSettings()">Apply</button></div></div></div>` + `<div class="sui-card topbar" data-sui-id="18">` + (`<div class="row topbar-row" data-sui-id="19">` + (`<h1 data-sui-id="20">Download manager</h1>` + `<div class="row topbar-actions" data-sui-id="21">` + (`<button type="button" class="sui-button btn-icon btn-icon-sm" data-sui-tooltip="${__sui_esc(`${phaseHint()}`)}" data-sui-id="22">${__sui_esc(phaseIcon())}</button>` + `<button type="button" class="sui-button btn-icon btn-icon-sm" data-sui-tooltip="Download settings" data-sui-id="23">⚙</button>`) + `</div>`) + `</div>`) + `</div>` + `<div class="sui-card panel panel-targets" data-sui-id="24">` + (`<div class="row targets-head" data-sui-id="25">` + (`<h3 data-sui-id="26">Targets</h3>` + `<button type="button" class="sui-button btn-icon btn-add btn-icon-sm" data-sui-tooltip="Add a target" data-sui-id="27">+</button>`) + `</div>` + ((targets.length === 0) ? `<p class="muted" data-sui-id="28">Add a target: paste the URL and press play.</p>` : '') + (() => { let s = ""; for (const t of (targets)) { s += __sui_keyed(`<div class="row target-row" data-sui-id="29">` + ((targetLocked(t) ? `<p class="target-url target-locked" data-sui-tooltip="The URL cannot be changed while the file is downloading" data-sui-id="30">${__sui_esc(t[`url`])}</p>` : `<input class="sui-input target-url" value="${__sui_esc(t[`url`])}" placeholder="https://example.com/file.bin" data-sui-id="31">`) + `<input class="sui-input target-name" value="${__sui_esc(t[`name`])}" placeholder="download.bin" data-sui-id="32">` + `<button type="button" class="sui-button btn-icon btn-prio" data-sui-tooltip="${__sui_esc(`${prioLabel(t[`priority`])}`)}" data-sui-id="33">${__sui_esc(prioIcon(t[`priority`]))}</button>` + `<button type="button" class="sui-button btn-icon btn-play" data-sui-tooltip="${__sui_esc(`${playHint(t)}`)}" data-sui-id="34">${__sui_esc(playIcon(t))}</button>` + `<button type="button" class="sui-button btn-icon btn-stop" data-sui-tooltip="${__sui_esc(`${stopHint(t)}`)}" data-sui-id="35">${__sui_esc(stopIcon(t))}</button>` + `<button type="button" class="sui-button btn-icon btn-cancel" data-sui-tooltip="Cancel this target" data-sui-id="36">✕</button>`) + `</div>` + `<div class="row target-foot" data-sui-id="37">` + (`<progress class="sui-progress bar-target" value="${__sui_esc(t[`percent`])}" max="${__sui_esc(100)}" data-sui-id="38"></progress>` + `<p class="target-bytes" data-sui-id="39">${__sui_esc(t[`received`])} / ${__sui_esc(t[`total`])}</p>`) + `</div>`, t[`id`]); } return s; })()) + `</div>` + `<div class="sui-tabs" aria-label="Download" data-sui-tabs-key="sui-tabs-1119-2" data-sui-id="40"><div class="sui-tablist" role="tablist"><button type="button" class="sui-tab active" role="tab" aria-selected="true" tabindex="0">Progress</button><button type="button" class="sui-tab" role="tab" aria-selected="false" tabindex="-1">Files</button><button type="button" class="sui-tab" role="tab" aria-selected="false" tabindex="-1">Data</button></div><div class="sui-tabpanels"><div class="sui-tabpanel active" role="tabpanel">` + (`<div class="sui-card panel" data-sui-id="41">` + (`<h3 data-sui-id="42">Total progress</h3>` + `<progress class="sui-progress bar-total" value="${__sui_esc(percent)}" max="${__sui_esc(100)}" data-sui-id="43"></progress>` + `<p class="verify" data-sui-id="44">${__sui_esc(verifyText)}</p>` + ((jobState === `done`) ? `<a class="dl-link" href="${__sui_esc(__sui_url(apiUrl(((`/downloads/` + jobId) + `/file`))))}" data-sui-id="45">Download the reassembled file</a>` : '') + ((jobSingle === false) ? `<div class="sui-accordion" aria-label="Parts" data-sui-accordion-key="sui-accordion-46" data-sui-id="46"><div class="sui-accordion-item open"><button type="button" class="sui-accordion-head" aria-expanded="true">${__sui_esc(partsView.length)} parts started - the most recent one first</button><div class="sui-accordion-panel">` + (`<div class="row part-counts" data-sui-id="47">` + (`<div class="col part-count" data-sui-id="48">` + (`<p class="count-value" data-sui-id="49">${__sui_esc(partTotal)}</p>` + `<p class="count-label" data-sui-id="50">parts</p>`) + `</div>` + `<div class="col part-count" data-sui-id="51">` + (`<p class="count-value count-active" data-sui-id="52">${__sui_esc(partActive)}</p>` + `<p class="count-label" data-sui-id="53">in download</p>`) + `</div>` + `<div class="col part-count" data-sui-id="54">` + (`<p class="count-value count-done" data-sui-id="55">${__sui_esc(partDone)}</p>` + `<p class="count-label" data-sui-id="56">completed</p>`) + `</div>`) + `</div>` + ((partsView.length === 0) ? `<p class="muted" data-sui-id="57">No download yet: the bars of the parts appear here.</p>` : `<div class="parts-scroll" data-sui-id="58">` + ((() => { let s = ""; for (const part of (partsView)) { s += `<div class="row part-row" data-sui-id="59">` + (`<p class="part-index" data-sui-id="60">Part ${__sui_esc(part[`Part`])}</p>` + `<progress class="sui-progress bar-part" value="${__sui_esc(part[`Percent`])}" max="${__sui_esc(100)}" data-sui-id="61"></progress>` + `<p class="part-state" data-sui-id="62">${__sui_esc(part[`State`])}</p>` + `<p class="part-bytes" data-sui-id="63">${__sui_esc(part[`Received`])} / ${__sui_esc(part[`Size`])}</p>` + ((part[`State`] === `error`) ? `<p class="part-error" data-sui-id="64">${__sui_esc(part[`Error`])}</p>` : '')) + `</div>`; } return s; })()) + `</div>`)) + `</div></div></div>` : '')) + `</div>`) + `</div><div class="sui-tabpanel" role="tabpanel" hidden>` + (`<div class="sui-card panel" data-sui-id="65">` + (`<h3 data-sui-id="66">Files</h3>` + ((files.length === 0) ? `<p class="muted" data-sui-id="67">No download yet: the files of the download appear here.</p>` : `<table name="files" data-sui-id="68">` + (`<thead data-sui-id="69">` + (`<tr data-sui-id="70">` + (`<th data-sui-id="71">Id</th>` + `<th data-sui-id="72">File</th>` + `<th data-sui-id="73">Priority</th>` + `<th data-sui-id="74">State</th>` + `<th data-sui-id="75">Parts</th>` + `<th data-sui-id="76">Size</th>`) + `</tr>`) + `</thead>` + `<tbody data-sui-id="77">` + (`<tr data-sui-id="78">` + (`<td data-sui-id="79">-</td>` + `<td data-sui-id="80">-</td>` + `<td data-sui-id="81">-</td>` + `<td data-sui-id="82">-</td>` + `<td data-sui-id="83">-</td>` + `<td data-sui-id="84">-</td>`) + `</tr>`) + `</tbody>`) + `</table>` + `<div class="files-list" data-sui-id="85">` + ((() => { let s = ""; for (const f of (files)) { s += `<div class="row file-row" data-sui-id="86">` + (`<p class="file-name" data-sui-id="87">${__sui_esc(f[`file`])}</p>` + ((f[`merged`] !== ``) ? `<a class="dl-file" href="${__sui_esc(__sui_url(apiUrl(((((`/downloads/` + jobId) + `/files/`) + String(f[`id`])) + `/file`))))}" data-sui-id="88">Download</a>` : '')) + `</div>`; } return s; })()) + `</div>`)) + `</div>`) + `</div><div class="sui-tabpanel" role="tabpanel" hidden>` + (`<div class="sui-card panel" data-sui-id="89">` + (`<h3 data-sui-id="90">Parts</h3>` + `<div class="row controls" data-sui-id="91">` + (`<button type="button" class="sui-button btn-warn" data-sui-id="92">Retry part ${__sui_esc(selectedPart)}</button>` + `<button type="button" class="sui-button btn-warn" data-sui-id="93">Retry all failed</button>` + `<button type="button" class="sui-button btn-ghost" data-sui-id="94">Refresh data</button>` + `<p class="muted" data-sui-id="95">Click a row to choose the part to download again.</p>`) + `</div>` + `<table name="parts" data-sui-selection="row" data-sui-table-onrowselect="onPartSelected" data-sui-id="96">` + (`<thead data-sui-id="97">` + (`<tr data-sui-id="98">` + (`<th data-sui-id="99">File</th>` + `<th data-sui-id="100">Priority</th>` + `<th data-sui-id="101">Part</th>` + `<th data-sui-id="102">From</th>` + `<th data-sui-id="103">To</th>` + `<th data-sui-id="104">Size</th>` + `<th data-sui-id="105">Received</th>` + `<th data-sui-id="106">Percent</th>` + `<th data-sui-id="107">State</th>`) + `</tr>`) + `</thead>` + `<tbody data-sui-id="108">` + (`<tr data-sui-id="109">` + (`<td data-sui-id="110">-</td>` + `<td data-sui-id="111">-</td>` + `<td data-sui-id="112">-</td>` + `<td data-sui-id="113">-</td>` + `<td data-sui-id="114">-</td>` + `<td data-sui-id="115">-</td>` + `<td data-sui-id="116">-</td>` + `<td data-sui-id="117">-</td>` + `<td data-sui-id="118">-</td>`) + `</tr>`) + `</tbody>`) + `</table>`) + `</div>` + `<div class="sui-card panel" data-sui-id="119">` + (`<h3 data-sui-id="120">Jobs</h3>` + `<table name="jobs" data-sui-id="121">` + (`<thead data-sui-id="122">` + (`<tr data-sui-id="123">` + (`<th data-sui-id="124">Id</th>` + `<th data-sui-id="125">File</th>` + `<th data-sui-id="126">State</th>` + `<th data-sui-id="127">Percent</th>` + `<th data-sui-id="128">Total</th>`) + `</tr>`) + `</thead>` + `<tbody data-sui-id="129">` + (`<tr data-sui-id="130">` + (`<td data-sui-id="131">-</td>` + `<td data-sui-id="132">-</td>` + `<td data-sui-id="133">-</td>` + `<td data-sui-id="134">-</td>` + `<td data-sui-id="135">-</td>`) + `</tr>`) + `</tbody>`) + `</table>`) + `</div>`) + `</div></div></div>` + `<div class="statusbar" data-sui-id="136">` + (`<div class="row statusbar-status" data-sui-id="137">` + (`<span class="sui-badge state-badge" data-sui-id="138">${__sui_esc(jobState)}</span>` + `<p class="status" data-sui-id="139">${__sui_esc(message)}</p>`) + `</div>` + `<div class="row statusbar-row" data-sui-id="140">` + (`<div class="col sb-item" data-sui-id="141">` + (`<p class="sb-value" data-sui-id="142">${__sui_esc(percent)}%</p>` + `<p class="sb-label" data-sui-id="143">progress</p>`) + `</div>` + `<div class="col sb-item" data-sui-id="144">` + (`<p class="sb-value" data-sui-id="145">${__sui_esc(totalText)}</p>` + `<p class="sb-label" data-sui-id="146">size</p>`) + `</div>` + `<div class="col sb-item" data-sui-id="147">` + (`<p class="sb-value" data-sui-id="148">${__sui_esc(receivedText)}</p>` + `<p class="sb-label" data-sui-id="149">downloaded</p>`) + `</div>` + `<div class="col sb-item" data-sui-id="150">` + (`<p class="sb-value" data-sui-id="151">${__sui_esc(etaText)}</p>` + `<p class="sb-label" data-sui-id="152">time left</p>`) + `</div>`) + `</div>`) + `</div>`
    + `</div>`;
  __sui_replace_view(el, __viewHtml);
  const e4 = document.querySelector('[data-sui-id="4"]');
  if (e4) {
    e4.onchange = (ev) => setSingle(ev.target.checked);
  }
  const e13 = document.querySelector('[data-sui-id="13"]');
  if (e13) {
    e13.oninput = (ev) => setWorkers(ev.target.value);
  }
  const e22 = document.querySelector('[data-sui-id="22"]');
  if (e22) {
    e22.onclick = onAction;
  }
  const e23 = document.querySelector('[data-sui-id="23"]');
  if (e23) {
    e23.onclick = openSettings;
  }
  const e27 = document.querySelector('[data-sui-id="27"]');
  if (e27) {
    e27.onclick = addTarget;
  }
  const e31 = document.querySelectorAll('[data-sui-id="31"]');
  e31.forEach((__el, __i) => {
    __el.oninput = (ev) => (onTargetUrl)(ev, __sui_rowkey(__el, __i));
  });
  const e32 = document.querySelectorAll('[data-sui-id="32"]');
  e32.forEach((__el, __i) => {
    __el.oninput = (ev) => (onTargetName)(ev, __sui_rowkey(__el, __i));
  });
  const e33 = document.querySelectorAll('[data-sui-id="33"]');
  e33.forEach((__el, __i) => {
    __el.onclick = (ev) => (onTargetPriority)(ev, __sui_rowkey(__el, __i));
  });
  const e34 = document.querySelectorAll('[data-sui-id="34"]');
  e34.forEach((__el, __i) => {
    __el.onclick = (ev) => (onTargetPlay)(ev, __sui_rowkey(__el, __i));
  });
  const e35 = document.querySelectorAll('[data-sui-id="35"]');
  e35.forEach((__el, __i) => {
    __el.onclick = (ev) => (onTargetStop)(ev, __sui_rowkey(__el, __i));
  });
  const e36 = document.querySelectorAll('[data-sui-id="36"]');
  e36.forEach((__el, __i) => {
    __el.onclick = (ev) => (onTargetCancel)(ev, __sui_rowkey(__el, __i));
  });
  const e92 = document.querySelector('[data-sui-id="92"]');
  if (e92) {
    e92.onclick = retrySelectedPart;
  }
  const e93 = document.querySelector('[data-sui-id="93"]');
  if (e93) {
    e93.onclick = retryFailedParts;
  }
  const e94 = document.querySelector('[data-sui-id="94"]');
  if (e94) {
    e94.onclick = refreshData;
  }
  __SUI_TBL.afterRender();
  if (window.__SUI_RICH) __SUI_RICH.afterRender();
}
__SUI_views["main"] = render_main;
__sui_bindView("etaText", "main");
__sui_bindView("files", "main");
__sui_bindView("jobId", "main");
__sui_bindView("jobSingle", "main");
__sui_bindView("jobState", "main");
__sui_bindView("message", "main");
__sui_bindView("partActive", "main");
__sui_bindView("partDone", "main");
__sui_bindView("partSize", "main");
__sui_bindView("partTotal", "main");
__sui_bindView("partsView", "main");
__sui_bindView("percent", "main");
__sui_bindView("receivedText", "main");
__sui_bindView("selectedPart", "main");
__sui_bindView("single", "main");
__sui_bindView("targets", "main");
__sui_bindView("totalText", "main");
__sui_bindView("verifyText", "main");
__sui_bindView("workers", "main");


__sui_render();

boot();

