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
	"function __sui_react(stateVar) {\n" +
	"  for (const v of (__SUI_bind[stateVar] || [])) __SUI_views[v]();\n" +
	"}\n" +
	"function __sui_render() {\n" +
	"  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();\n" +
	"}\n" +
	"function __sui_refresh() {\n" +
	"  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();\n" +
	"}\n" +
	`function __sui_replace_view(el, html) {
  var oldCharts = Array.prototype.slice.call(el.querySelectorAll('[data-sui-chart]'));
  var saved = [];
  for (var i = 0; i < oldCharts.length; i++) {
    saved.push({ node: oldCharts[i], key: oldCharts[i].getAttribute('data-sui-chart-name') || oldCharts[i].getAttribute('data-sui-chart-title') || '' });
  }
  var active = document.activeElement, activeInChart = null;
  if (active) for (var a = 0; a < saved.length; a++) if (saved[a].node.contains(active)) { activeInChart = active; break; }
  el.innerHTML = html;
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
    var redrawAttr = /^data-sui-chart-(data|type|format|x|y|width|height|overflow|zoom|point-width|point-size|bins|min|max|table-view|table-filters|table-add-row|table-delete-row|table-steppers|table-context-menu|table-selection|critical-path|holidays|weekend|timezone|show-non-working-days|progress-column|interactive|state-mode|grid|legend|values|title)$/;
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
	"async function __sui_fetch_guarded(url, allow, init) {\n" +
	"  var g = __sui_url_guard(url, allow, false);\n" +
	"  if (g[1]) return [null, g[1]];\n" +
	"  var opts = init || {};\n" +
	"  opts.redirect = 'error';\n" +
	"  var ctrl = null, timer = null;\n" +
	"  if (typeof AbortController !== 'undefined') { ctrl = new AbortController(); opts.signal = ctrl.signal; timer = setTimeout(function(){ ctrl.abort(); }, 30000); }\n" +
	"  try {\n" +
	"    var r = await fetch(g[0], opts);\n" +
	"    var len = Number(r.headers.get('content-length') || 0);\n" +
	"    if (len > 10485760) { if (r.body && r.body.cancel) { try { r.body.cancel(); } catch(e) {} } return [null, 'response too large']; }\n" +
	"    var text = await r.text();\n" +
	"    if (text.length > 10485760) return [null, 'response too large'];\n" +
	"    return [text, null];\n" +
	"  } catch(e) {\n" +
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
	"\n" +
	"const __SUI_MEM_STORE = Object.create(null);\n" +
	"function __sui_storage_get(k, def){ try{ const v=localStorage.getItem(k); if(v===null) return [def===undefined?null:def, null]; try{return [JSON.parse(v), null]}catch(e){return [v,null]} }catch(e){ if(k in __SUI_MEM_STORE){ try{return [JSON.parse(__SUI_MEM_STORE[k]),null]}catch(ee){return [__SUI_MEM_STORE[k],null]} } return [def===undefined?null:def,null] } }\n" +
	"function __sui_storage_set(k,v){ const sv=typeof v==='string'?v:JSON.stringify(v); try{ localStorage.setItem(k,sv); return [true,null] }catch(e){ __SUI_MEM_STORE[k]=sv; return [false,String(e)] } }\n" +
	"function __sui_storage_remove(k){ try{ localStorage.removeItem(k); return [true,null] }catch(e){ delete __SUI_MEM_STORE[k]; return [false,String(e)] } }\n" +
	"function __sui_storage_clear(){ try{ localStorage.clear(); return [true,null] }catch(e){ for(const kk in __SUI_MEM_STORE) delete __SUI_MEM_STORE[kk]; return [false,String(e)] } }\n" +
	"function __sui_storage_quota(){ try{ return [JSON.stringify(localStorage).length,null]; }catch(e){ return [0,String(e)]; } }\n" +
	"function __sui_storage_sync(){ try{ return [true,null]; }catch(e){ return [false,String(e)]; } }\n" +
	"function __sui_storage_keys(){ try{ const r=[]; for(let i=0;i<localStorage.length;i++){ const k=localStorage.key(i); if(k) r.push(k); } return [r,null] }catch(e){ return [Object.keys(__SUI_MEM_STORE),null] } }\n" +

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
	"  constructor(url) {\n" +
	"    this.sock = new WebSocket(url);\n" +
	"  }\n" +
	"  send(data) { this.sock.send(data); }\n" +
	"  close() { this.sock.close(); }\n" +
	"  open() { return this.sock.readyState === WebSocket.OPEN; }\n" +
	"  onMessage(fn) { this.sock.onmessage = (ev) => fn(ev.data); }\n" +
	"  onClose(fn) { this.sock.onclose = () => fn(); }\n" +
	"  onError(fn) { this.sock.onerror = (ev) => fn(String(ev)); }\n" +
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
	genMu.Lock()
	defer genMu.Unlock()
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
	for i := 0; i < len(p.funcs); i++ {
		f := p.funcs[i]
		genUserFuncs[f.name] = true
		if f.async {
			genAsyncFuncs[f.name] = true
		}
		var defs []string
		for j := 0; j < len(f.params); j++ {
			if f.params[j].dflt != nil {
				defs = append(defs, ExprToJs(f.params[j].dflt))
			} else {
				defs = append(defs, "")
			}
		}
		genFnDefaults[f.name] = defs
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
	sb.WriteString("\n\n")
	sb.WriteString(suiTableRuntime)
	// Drag-and-drop runtime. It is independent of the table runtime: the
	// generated app wires drag sources/targets by data attributes, while the
	// language exposes the captured keys through drag.source()/drag.target().
	sb.WriteString("\n\n")
	sb.WriteString(suiDndRuntime)
	// Menu bars use the same delegated runtime in every view and across
	// rerenders; labels/actions are generated as normal elements, not HTML.
	sb.WriteString("\n\n")
	sb.WriteString(suiMenuRuntime)
	// Tabs share the same delegated/runtime approach and survive view rerenders
	// through data-sui-tabs-key.
	sb.WriteString("\n\n")
	sb.WriteString(suiTabsRuntime)
	// Accordions use the same delegated pattern and keep the open item keyed
	// across view re-renders.
	sb.WriteString("\n\n")
	sb.WriteString(suiAccordionRuntime)
	// Toast hosts receive safe text notifications through toast.show().
	sb.WriteString("\n\n")
	sb.WriteString(suiToastRuntime)
	// Tooltips use delegated hover/focus events and text-only bubbles.
	sb.WriteString("\n\n")
	sb.WriteString(suiTooltipRuntime)
	// Searchable combobox controls filter their own listbox and sync a hidden
	// value binding back to application state.
	sb.WriteString("\n\n")
	sb.WriteString(suiComboboxRuntime)
	// Date/time pickers wrap a native input and synchronize the bound state.
	sb.WriteString("\n\n")
	sb.WriteString(suiDatePickerRuntime)
	// Trees own expansion, roving focus and selection; open state is persisted.
	sb.WriteString("\n\n")
	sb.WriteString(suiTreeRuntime)
	// Charts parse JSON/XML/CSV data and draw inert SVG nodes. Large charts
	// scroll their viewport or can be zoomed by the user.
	sb.WriteString("\n\n")
	sb.WriteString(suiChartRuntime)
	// Modal dialogs are opened by data-sui-modal name and use delegated events.
	sb.WriteString("\n\n")
	sb.WriteString(suiModalRuntime)
	// Class allow-list: dynamic class expressions are filtered to these names
	// so untrusted data cannot activate arbitrary app classes.
	sb.WriteString("\n\nconst __SUI_CLASSES = Object.create(null);\n")
	for _, cn := range securityClassNames(p) {
		sb.WriteString("__SUI_CLASSES[" + jsString(cn) + "] = 1;\n")
	}
	for i := 0; i < len(p.states); i++ {
		sd := p.states[i]
		sb.WriteString("\n\n")
		sb.WriteString("let " + sd.name + " = " + ExprToJs(sd.init) + ";\n")
		sb.WriteString("function " + setterName(sd.name) + "(v) { " + sd.name + " = v; __sui_react(" + jsString(sd.name) + "); }")
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
		sb.WriteString(rendererJs(p, &p.views[i], &ctx, multi, &base))
	}
	if multi {
		sb.WriteString("\n\nwindow.__SUI_PAGE = " + jsString(rootNames[0]) + ";\n")
		sb.WriteString("function __sui_page(n) {\n")
		sb.WriteString("  window.__SUI_PAGE = String(n);\n")
		sb.WriteString("  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();\n")
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

func ExprToJs(e *Expr) string {
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
				return js
			}
		}
		// User function calls append the missing trailing default values.
		if e.target != nil && e.target.kind == "ident" && isUserFunc(e.target.name) {
			var argJs []string
			for i := 0; i < len(e.args); i++ {
				argJs = append(argJs, ExprToJs(e.args[i]))
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
		return "[" + exprListJs(e.elements) + "]"
	}
	if k == "map" {
		var parts []string
		for i := 0; i < len(e.keys); i++ {
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
		out = append(out, stmtToJs(stmts[i], ind, ctx))
	}
	return strings.Join(out, "\n")
}

func stmtToJs(st *Stmt, ind string, ctx *JsCtx) string {
	if st == nil {
		return ""
	}
	k := st.kind
	if k == "let" {
		ctx.scope = append(ctx.scope, st.name)
		return ind + "let " + st.name + " = " + ExprToJs(st.init) + ";"
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
		paramNames = append(paramNames, f.params[i].name)
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
	return prefix + f.name + "(" + strings.Join(paramNames, ", ") + ") {\n" + body + "\n}"
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
	return "const __SUI_STD = {\n" + strings.Join(entries, ",\n") + "\n};"
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
	if key == "ws_send" {
		return "async (conn, data) => { try { conn.send(data); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_onmessage" {
		return "async (conn, fn) => { try { conn.onMessage((data) => fn(data)); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "ws_close" {
		return "async (conn) => { try { conn.close(); return [null, null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "json_parse" {
		return "async (s) => { try { return [JSON.parse(s), null]; } catch (e) { return [null, String(e)]; } }"
	}
	if key == "json_stringify" {
		return "async (v) => { try { return [JSON.stringify(v), null]; } catch (e) { return [null, String(e)]; } }"
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
	if key == "chart_on" {
		return "(name, id, event, handler) => __SUI_CHART.on(name, id, event, handler)"
	}
	if key == "chart_off" {
		return "(name, id, event) => __SUI_CHART.off(name, id, event)"
	}
	if key == "chart_bindings" {
		return "(name) => __SUI_CHART.bindings(name)"
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
	if key == "tree_clear" {
		return "(name) => __SUI_TREE.clear(name)"
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
	if tag == "menubar" {
		return "nav"
	}
	if tag == "combobox" || tag == "datepicker" || tag == "tree" || tag == "chart" || tag == "graph" {
		return "div"
	}
	if tag == "menu" || tag == "menuitem" || tag == "tab" || tag == "modal" {
		return "div"
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
	if e.tag == "button" {
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
	} else if e.tag == "tabs" {
		if len(e.text) > 0 {
			sb.WriteString(" aria-label=\"" + htmlTextParts(e.text) + "\"")
		}
		if idx >= 0 {
			sb.WriteString(" data-sui-tabs-key=\"sui-tabs-" + strconv.Itoa(idx) + "\"")
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
	} else if e.tag == "menuitem" {
		sb.WriteString(" role=\"menuitem\" tabindex=\"-1\"")
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
		if a := findAttr(e, "editable"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
			sb.WriteString(" data-sui-editable=\"1\"")
		}
		if a := findAttr(e, "fixed"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
			sb.WriteString(" data-sui-fixed=\"1\"")
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
	"progress":      "sui-progress",
	"badge":         "sui-badge",
	"alert":         "sui-alert",
	"spinner":       "sui-spinner",
	"card":          "sui-card",
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
		writeContainerClass(sb, e, ctx, base)
	} else {
		// Plain elements (p, div, span, ...): emit the class attribute from the
		// user's static or reactive expression.
		writePlainClass(sb, e)
	}

	writeDndAttrs(sb, e, ctx)
	writePickerAttrs(sb, e, ctx)
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		if _, isEv := eventAttrJs[a.key]; isEv {
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
		if chartTag(e.tag) && (a.key == "name" || a.key == "data" || a.key == "chartType" || a.key == "format" || a.key == "theme" ||
			a.key == "x" || a.key == "y" || a.key == "width" || a.key == "height" || a.key == "overflow" ||
			a.key == "zoom" || a.key == "pointWidth" || a.key == "pointSize" || a.key == "bins" || a.key == "min" || a.key == "max" ||
			a.key == "tableView" || a.key == "tableFilters" || a.key == "tableAddRow" || a.key == "tableDeleteRow" || a.key == "tableSteppers" || a.key == "tableContextMenu" || a.key == "tableSelection" || a.key == "criticalPath" || a.key == "holidays" || a.key == "weekend" || a.key == "timezone" || a.key == "showNonWorkingDays" || a.key == "progressColumn" || a.key == "interactive" || a.key == "stateMode" || a.key == "stateHints" || a.key == "stateHintBg" || a.key == "stateHintOpacity" || a.key == "stateHintTextColor" || a.key == "stateHintBorderColor" || a.key == "stateHintFont" || a.key == "stateHintFontSize" || a.key == "stateHintBorderWidth" || a.key == "stateHintPadding" || a.key == "stateHintRadius" ||
			a.key == "colorPicker" || a.key == "settingsButton" || a.key == "showZoom" || a.key == "showAdd" || a.key == "showHistory" || a.key == "onTaskChange" || a.key == "onDataChange" || a.key == "onCalendarChange" || a.key == "onResourceChange" || a.key == "onCommit" ||
			a.key == "onShapeClick" || a.key == "onShapeDoubleClick" || a.key == "onShapeMouseOver" || a.key == "onShapeMouseOut" || a.key == "onShapeContextMenu" ||
			a.key == "onLinkClick" || a.key == "onLinkDoubleClick" || a.key == "onLinkMouseOver" || a.key == "onLinkMouseOut" || a.key == "onLinkContextMenu" ||
			a.key == "onTagClick" || a.key == "onTagDoubleClick" || a.key == "onTagMouseOver" || a.key == "onTagMouseOut" || a.key == "onTagContextMenu" || a.key == "stateRouting" || a.key == "stateTagStyle" || a.key == "stateTagLink" || a.key == "stateLinkTagStyle" || a.key == "stateLinkTagLink" || a.key == "stateShapeTagStyle" || a.key == "stateShapeTagLink" || a.key == "stateTagFont" || a.key == "stateTagFontSize" || a.key == "stateShapeTagFont" || a.key == "stateShapeTagFontSize" || a.key == "stateLinkTagFont" || a.key == "stateLinkTagFontSize" || a.key == "stateRouteAlgorithm" || a.key == "stateLinkArcRadius" || a.key == "stateGrid" || a.key == "stateGridGap" || a.key == "stateGridColor" ||
			a.key == "showGrid" || a.key == "showLegend" || a.key == "showValues" ||
			a.key == "showResources" || a.key == "showProgress" ||
			a.key == "showSlack" || a.key == "showCosts" || a.key == "statusColors" || a.key == "ganttScale" || a.key == "autoSchedule" ||
			a.key == "linkColor" || a.key == "linkWidth" || a.key == "linkGap" || a.key == "linkStyle" || a.key == "arrowType" || a.key == "arrowSize" || a.key == "barBorderColor") {
			continue
		}
		if e.tag == "tree" && (a.key == "value" || a.key == "editable" || a.key == "editableDepth" ||
			a.key == "name" || a.key == "renameKey" || a.key == "renameValue" || a.key == "renameIcon" || a.key == "showRenameIcon") {
			continue
		}
		if e.tag == "treeitem" && (a.key == "key" || a.key == "open" || a.key == "editable" || a.key == "renameIcon" || a.key == "showRenameIcon") {
			continue
		}
		if a.key == "head" || a.key == "options" || a.key == "link" || a.key == "srcAllow" ||
			a.key == "draggable" || a.key == "directory" || a.key == "dragKey" ||
			a.key == "dropKey" || a.key == "dropFiles" || a.key == "pickFiles" || a.key == "pickMode" {
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

func elemJs(e *Element, g *ViewGen, ctx *JsCtx) string {
	if e == nil {
		return "''"
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
	if interactive && !genCompBody {
		collectWiring(e, idx, g, ctx)
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
		sb.WriteString("`" + wrapLabelOpen(e, idx) + "` + (`" + openTagHtml(e, tag, interactive, -1, ctx) + "`")
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

// modalElemJs renders a hidden modal with a title, body content and a footer
// of modalbutton children. The runtime opens it by data-sui-modal name.
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
	// legacy chart attributes only when an existing source explicitly has them.
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
	b.WriteString("\" data-sui-chart-interactive=\"")
	b.WriteString(chartFlagValue(e, "interactive", false))
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
	b.WriteString("\" data-sui-chart-state-mode=\"")
	b.WriteString(chartAttrValue(e, "stateMode", "editing"))
	b.WriteString("\" data-sui-chart-state-hints=\"")
	b.WriteString(chartFlagValue(e, "stateHints", true))
	b.WriteString("\" data-sui-chart-state-hint-bg=\"")
	b.WriteString(chartAttrValue(e, "stateHintBg", ""))
	b.WriteString("\" data-sui-chart-state-hint-opacity=\"")
	b.WriteString(chartAttrValue(e, "stateHintOpacity", ""))
	b.WriteString("\" data-sui-chart-state-hint-text-color=\"")
	b.WriteString(chartAttrValue(e, "stateHintTextColor", ""))
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
	for _, eventName := range stateChartEventAttrs {
		if a := findAttr(e, eventName); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
			b.WriteString("\" data-sui-chart-" + strings.ToLower(eventName) + "=\"")
			b.WriteString(htmlEscape(a.expr.name))
		}
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
	open := openTagHtml(e, "div", false, idx, ctx)
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
	b.WriteString(`<div class="sui-chart-legend"></div><div class="sui-chart-table" role="region" aria-label="Chart data table" hidden></div><div class="sui-chart-error" role="alert" hidden></div><div class="sui-chart-tooltip" role="tooltip" aria-live="polite" hidden></div></div>`)
	b.WriteString("`")
	return b.String()
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
	editable := false
	if a := findAttr(e, "editable"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		editable = a.expr.b
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
	if a := findAttr(e, "editableDepth"); a != nil {
		if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
			depth = strconv.Itoa(a.expr.i)
		} else if !a.hasExpr {
			depth = strings.TrimSpace(partsContent(a.text))
		}
	}
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">")
		open += " data-sui-tree=\"1\" data-sui-tree-name=\"" + htmlEscape(treeName) + "\" data-sui-tree-key=\"" + htmlEscape(treeKeyName) + "\" role=\"tree\""
		if editable {
			open += " data-sui-tree-editable=\"1\""
		}
		if depth != "" {
			open += " data-sui-tree-editable-depth=\"" + htmlEscape(depth) + "\""
		}
		if renameIcon != "" {
			open += " data-sui-tree-rename-icon=\"" + htmlEscape(renameIcon) + "\""
		}
		if !showRenameIcon {
			open += " data-sui-tree-show-rename-icon=\"0\""
		}
		if setter != "" {
			open += " data-sui-tree-setter=\"" + setter + "\""
		}
		if handler != "" {
			open += " data-sui-tree-onchange=\"" + handler + "\""
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
	b.WriteString("\"><ul class=\"sui-tree-group\" role=\"group\">")
	seq := 0
	for _, child := range treeChildren(e) {
		b.WriteString(treeItemElemJs(child, ctx, idx*1000+seq+1, 1, "0."+strconv.Itoa(seq)))
		seq++
	}
	b.WriteString("</ul></div>`")
	return b.String()
}

func treeItemElemJs(it *Element, ctx *JsCtx, itemID int, level int, path string) string {
	open := openTagHtml(it, "li", false, itemID, ctx)
	kids := treeChildren(it)
	isOpen := treeOpen(it)
	key := treeKey(it, path)
	if strings.HasSuffix(open, ">") {
		open = strings.TrimSuffix(open, ">")
		open += " data-sui-tree-item=\"1\" data-sui-key=\"" + htmlEscape(key) +
			"\" data-sui-path=\"" + htmlEscape(path) + "\" data-sui-level=\"" + strconv.Itoa(level) +
			"\" data-sui-editable=\"" + htmlEscape(treeItemEditable(it)) +
			"\" data-sui-show-rename-icon=\"" + htmlEscape(treeItemShowRenameIcon(it)) +
			"\" data-sui-rename-icon=\"" + htmlEscape(treeItemRenameIcon(it)) + "\" data-sui-open=\""
		if isOpen {
			open += "1"
		} else {
			open += "0"
		}
		open += "\" role=\"treeitem\" aria-level=\"" + strconv.Itoa(level) + "\" aria-selected=\"false\" tabindex=\"-1\""
		if len(kids) > 0 {
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
	if len(kids) > 0 {
		b.WriteString("<button type=\"button\" class=\"sui-tree-toggle\" tabindex=\"-1\" aria-label=\"Toggle item\">")
		if isOpen {
			b.WriteString("\u25be")
		} else {
			b.WriteString("\u25b8")
		}
		b.WriteString("</button>")
	} else {
		b.WriteString("<span class=\"sui-tree-spacer\"></span>")
	}
	b.WriteString("<span class=\"sui-tree-label\">")
	b.WriteString(htmlTextParts(it.text))
	b.WriteString("</span></div>")
	if len(kids) > 0 {
		b.WriteString("<ul class=\"sui-tree-group\" role=\"group\"")
		if !isOpen {
			b.WriteString(" hidden")
		}
		b.WriteString(">")
		for i, child := range kids {
			b.WriteString(treeItemElemJs(child, ctx, itemID*1000+i+1, level+1, path+"."+strconv.Itoa(i)))
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
		if n.item2 != "" {
			tmp := "__sui_" + n.item
			return "(() => { let s = \"\"; for (const " + tmp + " of Object.entries(" + ExprToJs(n.iterable) + ")) { const " + n.item + " = " + tmp + "[0]; const " + n.item2 + " = " + tmp + "[1]; s += " + bodyC + "; } return s; })()"
		}
		return "(() => { let s = \"\"; for (const " + n.item + " of (" + ExprToJs(n.iterable) + ")) { s += " + bodyC + "; } return s; })()"
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
		addPartsRefs(a.text, g)
		if a.hasExpr {
			addExprRefs(a.expr, g)
		}
	}
	addNodesRefs(e.children, g)
}

func viewBindings(states []string, viewName string, g *ViewGen) string {
	var names []string
	for i := 0; i < len(g.refs); i++ {
		name := g.refs[i]
		if !isNameIn(states, name) {
			continue
		}
		if !isNameIn(names, name) {
			names = append(names, name)
		}
	}
	var out []string
	sorted := sortStringList(names)
	for i := 0; i < len(sorted); i++ {
		out = append(out, "__sui_bindView("+jsString(sorted[i])+", "+jsString(viewName)+");")
	}
	return strings.Join(out, "\n")
}

func rendererJs(p *Prog, v *ViewDecl, ctx *JsCtx, multi bool, base *int) string {
	var g ViewGen
	// Continue the DFS numbering from prior root views (multi-page mode);
	// single-page programs pass 0, so their ids are unchanged.
	g.count = *base
	addPtrNodesRefs(childPtrs(v.body), &g)
	bodyJs := viewNodesJs(childPtrs(v.body), &g, ctx)
	*base = g.count
	var sb strings.Builder
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
		for j := 0; j < len(w.lines); j++ {
			sb.WriteString("  " + w.lines[j] + "\n")
		}
	}
	sb.WriteString("  __SUI_TBL.afterRender();\n")
	sb.WriteString("  __SUI_DND.afterRender();\n")
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
	return "function __html_" + v.name + "({ " + strings.Join(paramNames, ", ") + " }) {" + "\n" +
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
