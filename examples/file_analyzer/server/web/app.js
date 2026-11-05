"use strict";

const __SUI_views  = Object.create(null);
const __SUI_bind   = Object.create(null);
const __SUI_inst   = Object.create(null);

function __sui_bindView(stateVar, viewName) {
  (__SUI_bind[stateVar] ||= []).push(viewName);
}
function __sui_react(stateVar) {
  for (const v of (__SUI_bind[stateVar] || [])) __SUI_views[v]();
}
function __sui_render() {
  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();
}
function __sui_refresh() {
  for (const k of Object.keys(__SUI_views)) __SUI_views[k]();
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
async function __sui_fetch_guarded(url, allow, init) {
  var g = __sui_url_guard(url, allow, false);
  if (g[1]) return [null, g[1]];
  var opts = init || {};
  opts.redirect = 'error';
  var ctrl = null, timer = null;
  if (typeof AbortController !== 'undefined') { ctrl = new AbortController(); opts.signal = ctrl.signal; timer = setTimeout(function(){ ctrl.abort(); }, 30000); }
  try {
    var r = await fetch(g[0], opts);
    var len = Number(r.headers.get('content-length') || 0);
    if (len > 10485760) { if (r.body && r.body.cancel) { try { r.body.cancel(); } catch(e) {} } return [null, 'response too large']; }
    var text = await r.text();
    if (text.length > 10485760) return [null, 'response too large'];
    return [text, null];
  } catch(e) {
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
    if (!/^[a-z][a-z0-9-]*$/.test(name)) continue;
    if (name.charAt(0) === '-' || name.indexOf('inset-') === 0) continue;
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

const __SUI_MEM_STORE = Object.create(null);
function __sui_storage_get(k, def){ try{ const v=localStorage.getItem(k); if(v===null) return [def===undefined?null:def, null]; try{return [JSON.parse(v), null]}catch(e){return [v,null]} }catch(e){ if(k in __SUI_MEM_STORE){ try{return [JSON.parse(__SUI_MEM_STORE[k]),null]}catch(ee){return [__SUI_MEM_STORE[k],null]} } return [def===undefined?null:def,null] } }
function __sui_storage_set(k,v){ const sv=typeof v==='string'?v:JSON.stringify(v); try{ localStorage.setItem(k,sv); return [true,null] }catch(e){ __SUI_MEM_STORE[k]=sv; return [false,String(e)] } }
function __sui_storage_remove(k){ try{ localStorage.removeItem(k); return [true,null] }catch(e){ delete __SUI_MEM_STORE[k]; return [false,String(e)] } }
function __sui_storage_clear(){ try{ localStorage.clear(); return [true,null] }catch(e){ for(const kk in __SUI_MEM_STORE) delete __SUI_MEM_STORE[kk]; return [false,String(e)] } }
function __sui_storage_quota(){ try{ return [JSON.stringify(localStorage).length,null]; }catch(e){ return [0,String(e)]; } }
function __sui_storage_sync(){ try{ return [true,null]; }catch(e){ return [false,String(e)]; } }
function __sui_storage_keys(){ try{ const r=[]; for(let i=0;i<localStorage.length;i++){ const k=localStorage.key(i); if(k) r.push(k); } return [r,null] }catch(e){ return [Object.keys(__SUI_MEM_STORE),null] } }

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
  var tpls = new WeakMap(); // table element -> { heads, bodies } row templates
  // Resource limits: the table is a DOM widget, so huge payloads are refused
  // instead of freezing the page. Applications should paginate.
  var MAX_ROWS = (__SUI_SEC && __SUI_SEC.maxRows) || 100000;
  var MAX_COLS = (__SUI_SEC && __SUI_SEC.maxCols) || 10000;
  var MAX_TEXT = (__SUI_SEC && __SUI_SEC.maxText) || 20000000;

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
    t = { heads: heads, bodies: bodies };
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
    var t = templates(tbl);
    var n = 0, i, j;
    var hr = tbl.querySelector('thead tr');
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
    }
    var tb = tbodyOf(tbl);
    while (tb.firstChild) tb.removeChild(tb.firstChild);
    for (i = 0; i < rows.length; i++) {
      var tr = t.bodies[i % t.bodies.length].cloneNode(true);
      tr.removeAttribute('data-' + 'sui-id');
      var cs = prepRow(tr, n, 'td');
      for (j = 0; j < cs.length; j++) cs[j].textContent = rows[i][j] !== undefined ? rows[i][j] : '';
      tb.appendChild(tr);
    }
  }
  function storeApply(name, ds){
    store[name] = ds;
    var ts = findTables(name);
    for (var i = 0; i < ts.length; i++) {
      try { applyTo(ts[i], ds); } catch(e) { return { error: String(e) }; }
    }
    init();
    return { n: ds.rows.length };
  }
  function display(name, data){
    var ds = norm(data);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    return r.error ? [0, r.error] : [r.n, null];
  }
  // importText parses text explicitly: see parseSpec for the spec syntax.
  function importText(name, text, spec){
    var ds = normText(text, spec);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    return r.error ? [0, r.error] : [r.n, null];
  }
  function data(name){
    var ts = findTables(name);
    if (!ts.length) {
      var st = store[name];
      return [st ? st.rows.map(function(r){ return r.slice(); }) : [], null];
    }
    var out = [], rows = bodyRows(ts[0]);
    for (var i = 0; i < rows.length; i++) {
      var cs = cellsOf(rows[i]), r = [];
      for (var j = 0; j < cs.length; j++) r.push((cs[j].textContent || '').trim());
      out.push(r);
    }
    return [out, null];
  }
  function header(name){
    var ts = findTables(name), out = [];
    if (!ts.length) {
      var st = store[name];
      return [st && st.header ? st.header.slice() : [], null];
    }
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
    delete store[name];
    var ts = findTables(name);
    for (var i = 0; i < ts.length; i++) {
      var tb = ts[i].tBodies && ts[i].tBodies.length ? ts[i].tBodies[0] : null;
      if (tb) while (tb.firstChild) tb.removeChild(tb.firstChild);
    }
    return [0, null];
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
  function applyAll(){
    for (var k in store) {
      var ts = findTables(k);
      for (var i = 0; i < ts.length; i++) { try { applyTo(ts[i], store[k]); } catch(e) {} }
    }
  }
  function afterRender(){ applyAll(); init(); }

  // ---- editable / fixed behaviors -----------------------------------------
  function init(){
    var eds = document.querySelectorAll('table[data-sui-editable] td, table[data-sui-editable] th');
    for (var i=0;i<eds.length;i++){ var c=eds[i]; if (c.getAttribute('contenteditable') !== 'true') { c.setAttribute('contenteditable','true'); c.spellcheck=false; } }
    var fix = document.querySelectorAll('table[data-sui-fixed] td, table[data-sui-fixed] th');
    for (var j=0;j<fix.length;j++){
      var d=fix[j];
      if (!d.hasAttribute('tabindex')) d.tabIndex=0;
      var isEd = d.getAttribute('contenteditable') === 'true';
      if (!isEd) { var txt = d.textContent; if (d.getAttribute('data-full') !== txt) d.setAttribute('data-full', txt); }
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
    } else {
      var h=Math.max(18, drag.size + (ev.clientY - drag.start));
      var tr=drag.cell.parentNode;
      tr.style.height=h+'px';
      for(var j=0;j<tr.children.length;j++){ tr.children[j].style.height=h+'px'; }
    }
  }, true);
  document.addEventListener('mouseup', function(){ if(drag){ drag=null; document.body.style.cursor=''; } }, true);

  // ---- keep edited cells in the store so re-renders do not lose them -------
  document.addEventListener('input', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    if (!c) return;
    var tbl = c.closest('table');
    if (!tbl) return;
    var name = tbl.getAttribute('name');
    if (!name || !store[name]) return;
    var tr = c.parentNode, col = Array.prototype.indexOf.call(tr.children, c);
    if (col < 0) return;
    if (tr.parentNode && tr.parentNode.tagName === 'THEAD') {
      if (store[name].header) store[name].header[col] = c.textContent || '';
    } else {
      var tb = tr.parentNode, r = Array.prototype.indexOf.call(tb.rows, tr);
      if (r >= 0) {
        if (!store[name].rows[r]) store[name].rows[r] = [];
        store[name].rows[r][col] = c.textContent || '';
      }
    }
  }, true);

  if(window.MutationObserver){
    var pending=false;
    new MutationObserver(function(){ if(pending) return; pending=true; setTimeout(function(){ pending=false; init(); }, 60); }).observe(document.body,{childList:true,subtree:true,characterData:true});
  }
  init();
  return {
    display: display, data: data, header: header, json: json,
    select: select, slice: slice, clear: clear, importText: importText,
    exportText: exportText, applyAll: applyAll, afterRender: afterRender
  };
})();
window.__SUI_TBL = __SUI_TBL;

const __SUI_DND = (function(){
  var state = { from: '', to: '', data: '', files: [], objects: [], folder: '', error: '', active: false, source: null };
  var MAX_FILES = 2000;
  var MAX_DEPTH = 16;
  var MAX_PATH = 4096;
  var lastTarget = null;
  var dropSeq = 0;
  var bound = false;

  function text(v) {
    return v === null || v === undefined ? '' : String(v);
  }
  function cleanPath(v) {
    var p = text(v);
    while (p.charAt(0) === '/') p = p.slice(1);
    return p;
  }
  function closest(el, selector) {
    var n = el && el.nodeType === 3 ? el.parentNode : el;
    while (n && n.nodeType === 1) {
      if (n.matches && n.matches(selector)) return n;
      n = n.parentNode;
    }
    return null;
  }
  function dropZone(el) {
    return closest(el, '[data-sui-dropzone], [data-sui-dropfiles], [data-sui-drop-key]');
  }
  function isFileDrag(dt) {
    if (!dt) return false;
    if (dt.files && dt.files.length) return true;
    if (!dt.types) return false;
    for (var i = 0; i < dt.types.length; i++) {
      if (dt.types[i] === 'Files') return true;
    }
    return false;
  }
  function fileMeta(file, entryPath) {
    if (!file) return null;
    var p = text(file.path || file.webkitRelativePath || '');
    if (!p) p = cleanPath(entryPath || file.name || '');
    return {
      name: text(file.name),
      path: p,
      size: file.size || 0,
      type: text(file.type),
      isDirectory: false,
      lastModified: file.lastModified || 0
    };
  }
  function filesFromList(list) {
    var out = [];
    if (!list) return out;
    for (var i = 0; i < list.length && out.length < MAX_FILES; i++) {
      var m = fileMeta(list[i], '');
      if (m && m.path.length <= MAX_PATH) out.push(m);
    }
    return out;
  }
  function immediateFiles(dt) {
    var out = filesFromList(dt && dt.files);
    if (!out.length && dt && dt.items) {
      for (var i = 0; i < dt.items.length; i++) {
        var item = dt.items[i];
        if (!item || item.kind !== 'file' || !item.webkitGetAsEntry) continue;
        var entry = null;
        try { entry = item.webkitGetAsEntry(); } catch (e) { entry = null; }
        if (entry && entry.isDirectory) {
          out.push({
            name: text(entry.name),
            path: cleanPath(entry.fullPath),
            size: 0,
            type: '',
            isDirectory: true,
            lastModified: 0
          });
        }
      }
    }
    return out;
  }
  function readAll(reader) {
    return new Promise(function(resolve, reject) {
      var all = [];
      function step() {
        reader.readEntries(function(batch) {
          if (!batch || !batch.length) { resolve(all); return; }
          all = all.concat(Array.prototype.slice.call(batch));
          step();
        }, function() { reject('cannot list a directory'); });
      }
      step();
    });
  }
  function walkEntry(entry, out, depth, objects) {
    if (out.length >= MAX_FILES) return Promise.reject('too many files (limit ' + MAX_FILES + ')');
    if (depth > MAX_DEPTH) return Promise.reject('directory nesting is too deep (limit ' + MAX_DEPTH + ')');
    if (entry && entry.isFile) {
      return new Promise(function(resolve, reject) {
        entry.file(function(file) {
          var m = fileMeta(file, entry.fullPath);
          if (!m || !m.path) { resolve(); return; }
          if (m.path.length > MAX_PATH) { reject('path is too long (limit ' + MAX_PATH + ')'); return; }
          out.push(m);
          if (objects) objects.push(file);
          resolve();
        }, function() { reject('cannot read ' + text(entry.name)); });
      });
    }
    if (entry && entry.isDirectory) {
      return readAll(entry.createReader()).then(function(entries) {
        var chain = Promise.resolve();
        entries.forEach(function(child) {
          chain = chain.then(function() { return walkEntry(child, out, depth + 1, objects); });
        });
        return chain;
      });
    }
    return Promise.resolve();
  }
  function collectFiles(dt, objects) {
    if (!dt) return Promise.resolve([]);
    var entries = [];
    if (dt.items && dt.items.length) {
      for (var i = 0; i < dt.items.length; i++) {
        var item = dt.items[i];
        if (!item || item.kind !== 'file' || !item.webkitGetAsEntry) continue;
        var entry = null;
        try { entry = item.webkitGetAsEntry(); } catch (e) { entry = null; }
        if (entry) entries.push(entry);
      }
    }
    if (!entries.length) {
      if (objects && dt.files) {
        for (var j = 0; j < dt.files.length; j++) objects.push(dt.files[j]);
      }
      return Promise.resolve(filesFromList(dt.files));
    }
    var out = [];
    var chain = Promise.resolve();
    entries.forEach(function(entry) {
      chain = chain.then(function() { return walkEntry(entry, out, 0, objects); });
    });
    return chain.then(function() {
      if (objects && !objects.length && dt.files) {
        for (var i = 0; i < dt.files.length; i++) objects.push(dt.files[i]);
      }
      if (!out.length) return immediateFiles(dt);
      return out;
    });
  }
  function setTarget(el) {
    if (lastTarget && lastTarget !== el) lastTarget.classList.remove('sui-drop-active');
    lastTarget = el;
    if (el) {
      el.classList.add('sui-drop-active');
      state.to = text(el.getAttribute('data-sui-drop-key'));
    } else {
      state.to = '';
    }
  }
  function clearTarget() {
    if (lastTarget) lastTarget.classList.remove('sui-drop-active');
    lastTarget = null;
  }
  function showError(el, msg) {
    state.error = text(msg);
    if (el && el.parentNode) {
      var old = el.__suiDropError;
      if (old && old.parentNode) old.parentNode.removeChild(old);
      var box = document.createElement('div');
      box.className = 'sui-drop-error';
      box.setAttribute('role', 'alert');
      box.textContent = state.error;
      el.parentNode.insertBefore(box, el.nextSibling);
      el.__suiDropError = box;
    }
  }
  function applyDropFiles(el, setter, list) {
    var paths = [];
    for (var i = 0; i < list.length; i++) {
      var p = text(list[i] && (list[i].path || list[i].name));
      if (p) paths.push(p);
    }
    var sep = el && el.tagName === 'TEXTAREA' ? '\n' : ', ';
    var value = paths.join(sep);
    if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA')) {
      try { el.value = value; } catch (e) {}
    }
    var fn = window[setter];
    if (typeof fn !== 'function') {
      showError(el, 'Cannot accept the files: drop target state is not available');
      return;
    }
    fn(value);
  }
  function clearError(el) {
    if (el && el.__suiDropError && el.__suiDropError.parentNode) {
      el.__suiDropError.parentNode.removeChild(el.__suiDropError);
    }
    if (el) el.__suiDropError = null;
  }
  function pickedFiles(input) {
    var files = input && input.files ? input.files : [];
    if (files.length > MAX_FILES) {
      return { files: [], error: 'too many files (limit ' + MAX_FILES + ')' };
    }
    var out = [];
    for (var i = 0; i < files.length; i++) {
      var m = fileMeta(files[i], files[i].webkitRelativePath || files[i].name);
      if (!m || !m.path) continue;
      if (m.path.length > MAX_PATH) {
        return { files: [], error: 'path is too long (limit ' + MAX_PATH + ')' };
      }
      out.push(m);
    }
    return { files: out, error: '' };
  }
  function handlePicked(input, trigger, setter) {
    var res = pickedFiles(input);
    state.folder = '';
    state.error = '';
    clearError(trigger);
    if (res.error) {
      state.files = [];
      state.objects = [];
      showError(trigger, 'Cannot accept the files: ' + res.error);
      return;
    }
    state.files = res.files;
    state.objects = input && input.files ? Array.prototype.slice.call(input.files) : [];
    if (!res.files.length) {
      state.objects = [];
      showError(trigger, 'Cannot accept the files: no files were selected');
      return;
    }
    applyDropFiles(trigger, setter, res.files);
  }
  function openPicker(trigger) {
    var inp = document.createElement('input');
    inp.type = 'file';
    inp.style.cssText = 'position:fixed;left:-9999px;top:0;width:1px;height:1px;opacity:0';
    var accept = trigger.getAttribute('data-sui-pick-accept');
    if (accept) inp.accept = accept;
    if (trigger.getAttribute('data-sui-pick-multiple') === '1') inp.multiple = true;
    if (trigger.getAttribute('data-sui-pick-directory') === '1') {
      inp.webkitdirectory = true;
      inp.setAttribute('webkitdirectory', '');
    }
    var setter = trigger.getAttribute('data-sui-pickfiles');
    if (!setter) return;
    function cleanup() {
      if (inp.parentNode) inp.parentNode.removeChild(inp);
    }
    inp.addEventListener('change', function() {
      handlePicked(inp, trigger, setter);
      cleanup();
    });
    inp.addEventListener('cancel', cleanup);
    document.body.appendChild(inp);
    try {
      inp.click();
    } catch (e) {
      cleanup();
      showError(trigger, 'Cannot accept the files: file picker is not available');
    }
    setTimeout(cleanup, 120000);
  }
  async function collectDirectoryHandle(handle) {
    var files = [];
    var objects = [];
    async function walk(dir, prefix, depth) {
      if (depth > MAX_DEPTH) throw new Error('directory nesting is too deep (limit ' + MAX_DEPTH + ')');
      for await (const entry of dir.values()) {
        if (entry.kind === 'file') {
          if (files.length >= MAX_FILES) throw new Error('too many files (limit ' + MAX_FILES + ')');
          var file = await entry.getFile();
          var p = prefix + entry.name;
          var m = fileMeta(file, p);
          if (!m || !m.path) continue;
          if (m.path.length > MAX_PATH) throw new Error('path is too long (limit ' + MAX_PATH + ')');
          files.push(m);
          objects.push(file);
        } else if (entry.kind === 'directory') {
          await walk(entry, prefix + entry.name + '/', depth + 1);
        }
      }
    }
    await walk(handle, handle.name + '/', 0);
    return { files: files, objects: objects };
  }
  function applyDirectoryHandle(handle, res, trigger, setter) {
    state.folder = text(handle.name);
    state.files = res.files;
    state.objects = res.objects;
    state.error = '';
    clearError(trigger);
    if (!res.files.length) {
      // Empty folder: keep the folder name usable even though there are no
      // file objects to upload.
      var fn = window[setter];
      if (typeof fn === 'function') fn(state.folder);
      return;
    }
    applyDropFiles(trigger, setter, res.files);
  }
  async function openFSPicker(trigger) {
    if (typeof window.showDirectoryPicker !== 'function') {
      showError(trigger, 'Cannot accept the folder: File System Access API is not available in this browser');
      return;
    }
    var setter = trigger.getAttribute('data-sui-pickfiles');
    if (!setter) return;
    try {
      var handle = await window.showDirectoryPicker({ mode: 'read' });
      var res = await collectDirectoryHandle(handle);
      applyDirectoryHandle(handle, res, trigger, setter);
    } catch (e) {
      if (e && e.name === 'AbortError') return;
      showError(trigger, 'Cannot accept the folder: ' + text(e && e.message ? e.message : e));
    }
  }
  function onPickClick(e) {
    var trigger = closest(e.target, '[data-sui-pickfiles]');
    if (!trigger) return;
    var mode = text(trigger.getAttribute('data-sui-pick-mode')).toLowerCase();
    if (mode === 'fs' || (mode === 'auto' && typeof window.showDirectoryPicker === 'function')) {
      e.preventDefault();
      openFSPicker(trigger);
      return;
    }
    // The native <input type=file> opens its own dialog and is handled by the
    // change listener below.
    if (trigger.tagName === 'INPUT' && trigger.type === 'file') return;
    e.preventDefault();
    openPicker(trigger);
  }
  function onPickChange(e) {
    var input = e.target;
    if (!input || input.tagName !== 'INPUT' || input.type !== 'file') return;
    var setter = input.getAttribute('data-sui-pickfiles');
    if (!setter) return;
    handlePicked(input, input, setter);
  }
  function onDragStart(e) {
    dropSeq++;
    var el = closest(e.target, '[data-sui-drag-key]');
    state.from = el ? text(el.getAttribute('data-sui-drag-key')) : '';
    state.to = '';
    state.data = state.from;
    state.error = '';
    state.active = true;
    state.source = el;
    if (el) el.classList.add('sui-dragging');
    if (e.dataTransfer) {
      if (state.from) e.dataTransfer.setData('text/plain', state.from);
      e.dataTransfer.effectAllowed = 'move';
    }
  }
  function onDragOver(e) {
    var el = dropZone(e.target);
    if (el) {
      e.preventDefault();
      if (e.dataTransfer) {
        e.dataTransfer.dropEffect = isFileDrag(e.dataTransfer) ? 'copy' : 'move';
      }
      setTarget(el);
    } else {
      // A file dragged from the desktop must never navigate away from the app.
      if (isFileDrag(e.dataTransfer)) e.preventDefault();
      setTarget(null);
    }
  }
  function onDragLeave(e) {
    var el = dropZone(e.target);
    if (el && el === lastTarget) {
      if (!e.relatedTarget || !el.contains(e.relatedTarget)) setTarget(null);
    }
  }
  function onDrop(e) {
    var el = dropZone(e.target);
    var dt = e.dataTransfer;
    if (!el) {
      if (isFileDrag(dt)) e.preventDefault();
      dropSeq++;
      return;
    }
    e.preventDefault();
    setTarget(el);
    state.error = '';
    clearError(el);
    if (dt) {
      var plain = '';
      try { plain = text(dt.getData('text/plain')); } catch (err) { plain = ''; }
      if (plain) state.data = plain;
      if (!state.from && plain) state.from = plain;
      if (isFileDrag(dt)) {
        state.files = immediateFiles(dt);
        state.objects = [];
        state.folder = '';
      }
    }
    if (isFileDrag(dt)) state.active = true;
    var setter = el.getAttribute('data-sui-dropfiles');
    if (dt && isFileDrag(dt)) {
      var seq = ++dropSeq;
      collectFiles(dt, state.objects).then(function(list) {
        if (seq !== dropSeq) return;
        state.files = list;
        if (!list.length) {
          if (setter) showError(el, 'Cannot accept the files: no files were found');
          return;
        }
        if (setter) applyDropFiles(el, setter, list);
      }).catch(function(err) {
        if (seq !== dropSeq) return;
        state.objects = [];
        showError(el, 'Cannot accept the files: ' + text(err || 'unreadable data'));
      });
    } else {
      dropSeq++;
    }
    setTimeout(function() {
      clearTarget();
      state.active = false;
    }, 0);
  }
  function onDragEnd(e) {
    if (state.source) state.source.classList.remove('sui-dragging');
    clearTarget();
    setTimeout(function() {
      state.from = '';
      state.data = '';
      state.active = false;
      state.source = null;
      // Keep state.files/state.objects: they are the last selection and are
      // consumed by http.postFiles() after the drag/drop gesture has ended.
    }, 0);
  }
  function init() {
    if (bound || typeof document === 'undefined') return;
    bound = true;
    document.addEventListener('dragstart', onDragStart, true);
    document.addEventListener('dragenter', onDragOver, true);
    document.addEventListener('dragover', onDragOver, true);
    document.addEventListener('dragleave', onDragLeave, true);
    document.addEventListener('drop', onDrop, true);
    document.addEventListener('dragend', onDragEnd, true);
    document.addEventListener('click', onPickClick, true);
    document.addEventListener('change', onPickChange, true);
  }
  function afterRender() {
    clearTarget();
  }
  function folderName() {
    if (state.folder) return state.folder;
    // A directory dropped without files is represented by a synthetic entry.
    if (state.files.length === 1 && state.files[0] && state.files[0].isDirectory) {
      return text(state.files[0].name || state.files[0].path);
    }
    var root = '';
    for (var i = 0; i < state.files.length; i++) {
      var p = text(state.files[i] && state.files[i].path);
      var slash = p.indexOf('/');
      if (slash <= 0) return '';
      var seg = p.slice(0, slash);
      if (!root) root = seg;
      else if (root !== seg) return '';
    }
    return root;
  }
  function clear() {
    dropSeq++;
    if (state.source) state.source.classList.remove('sui-dragging');
    clearTarget();
    state.from = '';
    state.to = '';
    state.data = '';
    state.files = [];
    state.objects = [];
    state.folder = '';
    state.error = '';
    state.active = false;
    state.source = null;
  }
  init();
  return {
    afterRender: afterRender,
    from: function() { return state.from; },
    to: function() { return state.to; },
    data: function() { return state.data; },
    files: function() { return state.files; },
    objects: function() { return state.objects; },
    folder: function() { return folderName(); },
    error: function() { return state.error; },
    active: function() { return state.active; },
    clear: clear
  };
})();
window.__SUI_DND = __SUI_DND;

function __sui_drag_source() { return __SUI_DND.from(); }
function __sui_drag_target() { return __SUI_DND.to(); }
function __sui_drag_data() { return __SUI_DND.data(); }
function __sui_drag_files() { return __SUI_DND.files(); }
function __sui_drag_folder() { return __SUI_DND.folder(); }
function __sui_drag_error() { return __SUI_DND.error(); }
function __sui_drag_active() { return __SUI_DND.active(); }
function __sui_drag_clear() { return __SUI_DND.clear(); }
function __sui_files_form() {
  var files = __SUI_DND.objects();
  if (!files || !files.length) return null;
  var meta = __SUI_DND.files();
  var fd = new FormData();
  for (var i = 0; i < files.length; i++) {
    var f = files[i];
    fd.append('files', f, f.name || 'file');
    // Parallel field: browsers do not expose absolute paths, but the relative
    // path (folder/sub/file) is useful to the backend for display/grouping.
    var p = '';
    if (meta && meta[i]) p = meta[i].path || meta[i].name || '';
    fd.append('paths', p || f.name || '');
  }
  return fd;
}
function __sui_move_item(arr, from, to) {
  var src = Array.isArray(arr) ? arr.slice() : [];
  var a = parseInt(from, 10);
  var b = parseInt(to, 10);
  if (!isFinite(a) || !isFinite(b) || a < 0 || a >= src.length) return src;
  if (b < 0) b = 0;
  if (b >= src.length) b = src.length - 1;
  var item = src.splice(a, 1)[0];
  src.splice(b, 0, item);
  return src;
}

const __SUI_CLASSES = Object.create(null);


let attachment = ``;
function setAttachment(v) { attachment = v; __sui_react("attachment"); }

let status = `idle`;
function setStatus(v) { status = v; __sui_react("status"); }

let errorMsg = ``;
function setErrorMsg(v) { errorMsg = v; __sui_react("errorMsg"); }

let summary = ``;
function setSummary(v) { summary = v; __sui_react("summary"); }

function clearSelection() {
  __sui_drag_clear();
  setAttachment(``);
  setStatus(`idle`);
  setErrorMsg(``);
  setSummary(``);
}

async function showResponse(body) {
  const __sui_r = await __SUI_STD.json_parse(body);
  const data = __sui_r[0];
  const perr = __sui_r[1];
  if ((perr !== null)) {
    setStatus(`error`);
    setErrorMsg((`Invalid server response: ` + perr));
    return;
  }
  if (Object.prototype.hasOwnProperty.call(data, `error`)) {
    setStatus(`error`);
    setErrorMsg(data[`error`]);
    return;
  }
  setStatus(`ok`);
  const __sui_r2 = __SUI_STD.table_display(`analysis`, data);
  const n = __sui_r2[0];
  const terr = __sui_r2[1];
  if ((terr !== null)) {
    setStatus(`error`);
    setErrorMsg(terr);
    return;
  }
  if ((n === 1)) {
    setSummary(`1 result`);
  } else {
    setSummary((String(n) + ` results`));
  }
}

async function analyzeFiles() {
  const __sui_r = await __SUI_STD.http_postfiles(`/api/analyze`);
  const body = __sui_r[0];
  const err = __sui_r[1];
  if ((err !== null)) {
    setStatus(`error`);
    setErrorMsg(err);
    return;
  }
  showResponse(body);
}

async function analyzeFolder(folder) {
  const __sui_r = await __SUI_STD.json_stringify(({"folder": folder}));
  const payload = __sui_r[0];
  const jerr = __sui_r[1];
  if ((jerr !== null)) {
    setStatus(`error`);
    setErrorMsg(String(jerr));
    return;
  }
  const __sui_r2 = await __SUI_STD.http_post(`/api/analyze-folder`, payload);
  const body = __sui_r2[0];
  const err = __sui_r2[1];
  if ((err !== null)) {
    setStatus(`error`);
    setErrorMsg(err);
    return;
  }
  showResponse(body);
}

async function analyze() {
  setStatus(`loading`);
  setErrorMsg(``);
  setSummary(``);
  let files = __sui_drag_files();
  let folder = __sui_drag_folder();
  if (((files.length === 0) || ((files.length === 1) && (files[0][`isDirectory`] === true)))) {
    if ((folder === ``)) {
      setStatus(`error`);
      setErrorMsg(`Select files or a folder first`);
      return;
    }
    analyzeFolder(folder);
  } else {
    analyzeFiles();
  }
}

const __SUI_STD = {
  http_get: async (url) => { try { const r = await fetch(url); return [await r.text(), null]; } catch (e) { return [null, String(e)]; } },
  http_getstatus: async (url) => { try { const r = await fetch(url); return [r.status, null]; } catch (e) { return [null, String(e)]; } },
  http_post: async (url, data) => { try { const r = await fetch(url, {method: "POST", headers: {"Content-Type": "application/json"}, body: data}); return [await r.text(), null]; } catch (e) { return [null, String(e)]; } },
  http_postfiles: async (url) => { const fd = __sui_files_form(); if (!fd) return [null, 'no files selected']; try { const abs = new URL(url, location.origin).href; return __sui_fetch_guarded(abs, abs, {method: 'POST', body: fd}); } catch (e) { return [null, String(e)]; } },
  http_postfilessafe: async (url, allow) => { const fd = __sui_files_form(); if (!fd) return [null, 'no files selected']; return __sui_fetch_guarded(url, allow, {method: 'POST', body: fd}); },
  http_getsafe: async (url, allow) => __sui_fetch_guarded(url, allow, {}),
  http_getstatussafe: async (url, allow) => __sui_fetch_status_guarded(url, allow),
  http_postsafe: async (url, data, allow) => __sui_fetch_guarded(url, allow, {method: "POST", headers: {"Content-Type": "application/json"}, body: data}),
  json_parse: async (s) => { try { return [JSON.parse(s), null]; } catch (e) { return [null, String(e)]; } },
  json_stringify: async (v) => { try { return [JSON.stringify(v), null]; } catch (e) { return [null, String(e)]; } },
  table_display: (name, data) => __SUI_TBL.display(name, data),
  table_data: (name) => __SUI_TBL.data(name),
  table_header: (name) => __SUI_TBL.header(name),
  table_select: (name, cols) => __SUI_TBL.select(name, cols),
  table_slice: (name, from, count) => __SUI_TBL.slice(name, from, count),
  table_importtext: (name, text, spec) => __SUI_TBL.importText(name, text, spec),
  table_exporttext: (name, spec) => __SUI_TBL.exportText(name, spec),
  table_json: (name) => __SUI_TBL.json(name),
  table_clear: (name) => __SUI_TBL.clear(name)
};

function render_main() {
  const el = document.getElementById("app");
  el.innerHTML = `<div class="sui-view" data-sui-view="main">`
    + `<h1 data-sui-id="0">File analyzer</h1>` + `<p data-sui-id="1">Choose files with the native dialog or drag them onto the field. The Go backend detects their type.</p>` + `<div class="row" data-sui-id="2">` + (`<button type="button" data-sui-pickfiles="setAttachment" data-sui-pick-multiple="1" data-sui-id="3">Choose files...</button>` + `<button type="button" data-sui-pickfiles="setAttachment" data-sui-pick-mode="auto" data-sui-pick-multiple="1" data-sui-pick-directory="1" data-sui-id="4">Choose folder...</button>` + `<button type="button" data-sui-id="5">Analyze</button>` + `<button type="button" data-sui-id="6">Clear</button>`) + `</div>` + `<input class="sui-input" data-sui-dropzone="1" data-sui-dropfiles="setAttachment" value="${__sui_esc(attachment)}" placeholder="Selected files" data-sui-id="7">` + ((__sui_drag_folder() !== ``) ? `<p data-sui-id="8">Folder: ${__sui_esc(__sui_drag_folder())}</p>` : '') + ((status === `loading`) ? `<p data-sui-id="9">Analyzing files on the Go backend...</p>` : '') + ((status === `error`) ? `<div class="sui-alert" data-sui-id="10">${__sui_esc(errorMsg)}</div>` : '') + ((status === `ok`) ? `<p data-sui-id="11">${__sui_esc(summary)}</p>` + `<table name="analysis" data-sui-id="12">` + (`<thead data-sui-id="13">` + (`<tr data-sui-id="14">` + (`<th data-sui-id="15">File</th>` + `<th data-sui-id="16">Path</th>` + `<th data-sui-id="17">Size</th>` + `<th data-sui-id="18">MIME</th>` + `<th data-sui-id="19">Category</th>` + `<th data-sui-id="20">Detected by</th>` + `<th data-sui-id="21">Detail</th>`) + `</tr>`) + `</thead>` + `<tbody data-sui-id="22">` + (`<tr data-sui-id="23">` + (`<td data-sui-id="24"></td>` + `<td data-sui-id="25"></td>` + `<td data-sui-id="26"></td>` + `<td data-sui-id="27"></td>` + `<td data-sui-id="28"></td>` + `<td data-sui-id="29"></td>` + `<td data-sui-id="30"></td>`) + `</tr>`) + `</tbody>`) + `</table>` : '')
    + `</div>`;
  const e5 = document.querySelector('[data-sui-id="5"]');
  e5.onclick = analyze;
  const e6 = document.querySelector('[data-sui-id="6"]');
  e6.onclick = clearSelection;
  const e7 = document.querySelector('[data-sui-id="7"]');
  e7.oninput = (ev) => setAttachment(ev.target.value);
  __SUI_TBL.afterRender();
  __SUI_DND.afterRender();
}
__SUI_views["main"] = render_main;
__sui_bindView("attachment", "main");
__sui_bindView("errorMsg", "main");
__sui_bindView("status", "main");
__sui_bindView("summary", "main");


__sui_render();

