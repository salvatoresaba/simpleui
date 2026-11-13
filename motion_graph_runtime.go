package main

// motion_graph_runtime.go - declarative state graphs for element animation.
//
// A designer-authored graph is serialized as JSON in the motionGraph attribute.
// The runtime keeps the current state on the element, applies that state's CSS
// declarations and lets CSS transitions (or an instant jump for reduced motion)
// interpolate to the next state. Links carry the trigger and timing:
//
//	{"initial":"closed","states":[...],"links":[
//	  {"from":"closed","to":"open","event":"click","duration":"320ms"}]}
//
// Events are ordinary DOM events; "after" links continue a chain when the
// previous transition completes. The runtime does not use eval and only reads
// literal, author-written JSON that the designer emitted.
const suiMotionGraphRuntime = `const __SUI_MOTION_GRAPH = (function(){
  var stateMemory = Object.create(null);
  var reduced = false;
  try { reduced = !!(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches); } catch (e) {}

  function timeMs(v, fallback) {
    if (v == null || v === '') return fallback;
    var s = String(v).trim();
    if (/^[0-9]+$/.test(s)) return parseInt(s, 10);
    var m = /^([0-9]*\.?[0-9]+)(ms|s)$/.exec(s);
    if (!m) return fallback;
    return m[2] === 's' ? parseFloat(m[1]) * 1000 : parseFloat(m[1]);
  }
  function keyOf(el) {
    return el.getAttribute('data-sui-motion-key') || el.getAttribute('data-sui-id') || '';
  }
  function stateById(cfg, id) {
    var list = cfg && cfg.states ? cfg.states : [];
    for (var i = 0; i < list.length; i++) if (list[i] && list[i].id === id) return list[i];
    return null;
  }
  function parseDecls(text) {
    var out = [];
    var buf = '';
    var depth = 0;
    var quote = '';
    for (var i = 0; i < String(text || '').length; i++) {
      var ch = text.charAt(i);
      if (quote) {
        buf += ch;
        if (ch === quote && text.charAt(i - 1) !== '\\') quote = '';
        continue;
      }
      if (ch === '"' || ch === "'") { quote = ch; buf += ch; continue; }
      if (ch === '(') depth++;
      else if (ch === ')' && depth > 0) depth--;
      if (ch === ';' && depth === 0) {
        addDecl(out, buf);
        buf = '';
      } else {
        buf += ch;
      }
    }
    addDecl(out, buf);
    return out;
  }
  function addDecl(out, raw) {
    var s = String(raw || '').trim();
    if (!s) return;
    var at = s.indexOf(':');
    if (at <= 0) return;
    var name = s.slice(0, at).trim().toLowerCase();
    var value = s.slice(at + 1).trim();
    if (!name || !value) return;
    out.push({ name: name, value: value });
  }
  function parseValue(v) {
    if (v == null) return null;
    var s = String(v).trim();
    if (!s) return null;
    var m = /^(-?\d*\.?\d+)([a-z%]*)$/i.exec(s);
    if (m) return { kind: 'num', n: parseFloat(m[1]), unit: m[2] || '' };
    var cm = /^rgba?\(([0-9.,\s]+)\)$/i.exec(s);
    if (!cm) return null;
    var parts = cm[1].split(',').map(function(x){ return parseFloat(x); });
    if (parts.length < 3 || isNaN(parts[0])) return null;
    return { kind: 'color', r: parts[0], g: parts[1], b: parts[2], a: parts.length > 3 ? parts[3] : 1 };
  }
  function interpolate(a, b, t) {
    if (!a || !b) return null;
    if (a.kind === 'color' && b.kind === 'color') {
      var r = Math.round(a.r + (b.r - a.r) * t);
      var g = Math.round(a.g + (b.g - a.g) * t);
      var bl = Math.round(a.b + (b.b - a.b) * t);
      var al = a.a + (b.a - a.a) * t;
      return 'rgba(' + r + ',' + g + ',' + bl + ',' + (Math.round(al * 1000) / 1000) + ')';
    }
    if (a.kind === 'num' && b.kind === 'num' && a.unit === b.unit) {
      return (a.n + (b.n - a.n) * t) + a.unit;
    }
    return null;
  }
  function ease(t, name) {
    if (t <= 0) return 0;
    if (t >= 1) return 1;
    name = String(name || 'ease-out').toLowerCase();
    if (name === 'linear') return t;
    if (name === 'ease-in') return t * t;
    if (name === 'ease-out') return 1 - (1 - t) * (1 - t);
    if (name === 'ease-in-out') return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
    var m = /^cubic-bezier\(([0-9.,\s-]+)\)$/.exec(name);
    if (m) {
      var p = m[1].split(',').map(function(n){ return parseFloat(n); });
      if (p.length === 4) {
        function sx(u) { return 3 * (1 - u) * (1 - u) * u * p[0] + 3 * (1 - u) * u * u * p[2] + u * u * u; }
        function sy(u) { return 3 * (1 - u) * (1 - u) * u * p[1] + 3 * (1 - u) * u * u * p[3] + u * u * u; }
        var lo = 0, hi = 1;
        for (var i = 0; i < 20; i++) { var mid = (lo + hi) / 2; if (sx(mid) < t) lo = mid; else hi = mid; }
        return sy((lo + hi) / 2);
      }
    }
    return 1 - (1 - t) * (1 - t);
  }
  function applyState(el, cfg, id, link, instant) {
    var state = stateById(cfg, id);
    if (!state) return;
    var decls = parseDecls(state.css || '');
    var next = Object.create(null);
    var target = Object.create(null);
    for (var i = 0; i < decls.length; i++) { next[decls[i].name] = 1; target[decls[i].name] = decls[i].value; }
    var prev = el.__suiGraphProps || [];
    var all = prev.slice();
    for (var j = 0; j < decls.length; j++) if (all.indexOf(decls[j].name) < 0) all.push(decls[j].name);
    var duration = Math.max(0, timeMs(link && link.duration, 240));
    var delay = Math.max(0, timeMs(link && link.delay, 0));
    var easing = (link && link.easing) || 'ease-out';
    if (reduced || instant || !link || duration <= 0) { duration = 0; delay = 0; }
    var token = (el.__suiGraphToken || 0) + 1;
    el.__suiGraphToken = token;
    function finish() {
      if (token !== el.__suiGraphToken) return;
      for (var p = 0; p < prev.length; p++) if (!next[prev[p]]) el.style.removeProperty(prev[p]);
      for (var d = 0; d < decls.length; d++) el.style.setProperty(decls[d].name, decls[d].value);
      el.__suiGraphProps = decls.map(function(x){ return x.name; });
    }
    if (duration <= 0 || !all.length) { finish(); return; }
    // Evaluate both measured states before starting: compute the start values,
    // temporarily apply the target to read its computed end values, then put
    // the start values back and tween the difference.
    var starts = {}, was = {}, explicit = Object.create(null);
    for (var a = 0; a < all.length; a++) {
      var p0 = all[a];
      was[p0] = el.style.getPropertyValue(p0);
      starts[p0] = getComputedStyle(el).getPropertyValue(p0);
      explicit[p0] = el.style.getPropertyPriority(p0);
    }
    for (var r = 0; r < prev.length; r++) if (!next[prev[r]]) el.style.removeProperty(prev[r]);
    for (var t = 0; t < decls.length; t++) el.style.setProperty(decls[t].name, decls[t].value);
    var ends = {};
    for (var e = 0; e < all.length; e++) ends[all[e]] = getComputedStyle(el).getPropertyValue(all[e]);
    for (var b = 0; b < all.length; b++) {
      var pb = all[b];
      if (was[pb]) el.style.setProperty(pb, was[pb], explicit[pb]);
      else el.style.removeProperty(pb);
    }
    var items = [];
    for (var c = 0; c < all.length; c++) {
      var pc = all[c];
      var va = parseValue(starts[pc]);
      var vb = parseValue(ends[pc]);
      if (va && vb) items.push({ name: pc, a: va, b: vb });
    }
    var nonAnim = [];
    for (var n = 0; n < all.length; n++) {
      var pn = all[n];
      var found = false;
      for (var q = 0; q < items.length; q++) if (items[q].name === pn) { found = true; break; }
      if (!found) nonAnim.push(pn);
    }
    for (var q2 = 0; q2 < nonAnim.length; q2++) {
      if (target[nonAnim[q2]] != null) el.style.setProperty(nonAnim[q2], target[nonAnim[q2]]);
      else el.style.removeProperty(nonAnim[q2]);
    }
    var start = performance.now() + delay;
    function frame() {
      if (token !== el.__suiGraphToken) return;
      var t = (performance.now() - start) / duration;
      var k = ease(Math.max(0, Math.min(1, t)), easing);
      for (var m = 0; m < items.length; m++) {
        var v = interpolate(items[m].a, items[m].b, k);
        if (v != null) el.style.setProperty(items[m].name, v);
      }
      if (t < 1) {
        setTimeout(frame, 16);
      } else {
        finish();
      }
    }
    setTimeout(frame, 16);
  }
  function findLink(cfg, from, event) {
    var links = cfg && cfg.links ? cfg.links : [];
    for (var i = 0; i < links.length; i++) {
      var l = links[i];
      if (l && l.from === from && l.event === event) return l;
    }
    return null;
  }
  function stateCondition(link) {
    var w = String(link && link.when || '').trim();
    var neg = false;
    if (w.charAt(0) === '!') { neg = true; w = w.slice(1).trim(); }
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(w)) return false;
    var get = window.__SUI_STATE_GET && window.__SUI_STATE_GET[w];
    if (!get) return false;
    var v = false;
    try { v = !!get(); } catch (e) { return false; }
    return neg ? !v : v;
  }
  function checkStateLinks(el, cfg) {
    if (!el.isConnected || el.__suiGraphEditLocked) return;
    var seen = el.__suiGraphSeen || (el.__suiGraphSeen = Object.create(null));
    var links = cfg && cfg.links ? cfg.links : [];
    for (var i = 0; i < links.length; i++) {
      var l = links[i];
      if (!l || l.event !== 'state' || l.from !== el.__suiGraphState) continue;
      var now = stateCondition(l);
      if (now && !seen[l.id]) {
        seen[l.id] = true;
        run(el, cfg, l);
        return;
      }
      seen[l.id] = now;
    }
  }
  function fireAfter(el, cfg, from) {
    var l = findLink(cfg, from, 'after');
    if (l) run(el, cfg, l);
  }
  function run(el, cfg, link) {
    if (!link) return;
    var to = stateById(cfg, link.to);
    if (!to) return;
    el.__suiGraphState = link.to;
    var key = keyOf(el);
    if (key) stateMemory[key] = link.to;
    applyState(el, cfg, link.to, link, false);
    var wait = Math.max(0, timeMs(link.duration, 240)) + Math.max(0, timeMs(link.delay, 0));
    var token = el.__suiGraphToken;
    setTimeout(function() {
      if (token !== el.__suiGraphToken) return;
      fireAfter(el, cfg, link.to);
    }, wait + 60);
  }
  function unbind(el) {
    var handlers = el.__suiGraphHandlers || [];
    for (var i = 0; i < handlers.length; i++) el.removeEventListener(handlers[i].event, handlers[i].fn);
    el.__suiGraphHandlers = [];
    el.__suiGraphBound = false;
  }
  function bind(el, cfg) {
    if (el.__suiGraphBound) unbind(el);
    var handlers = [];
    var seen = Object.create(null);
    var links = cfg && cfg.links ? cfg.links : [];
    for (var i = 0; i < links.length; i++) {
      var ev = links[i] && links[i].event;
      if (!ev || ev === 'after' || ev === 'state' || seen[ev]) continue;
      seen[ev] = 1;
      (function(eventName) {
        var fn = function(e) {
          var link = findLink(cfg, el.__suiGraphState, eventName);
          if (!link) return;
          if (el.__suiGraphEditLocked) return;
          run(el, cfg, link);
        };
        el.addEventListener(eventName, fn);
        handlers.push({ event: eventName, fn: fn });
      })(ev);
    }
    el.__suiGraphHandlers = handlers;
    el.__suiGraphBound = true;
  }
  function init(el) {
    var raw = el.getAttribute('data-sui-motion-graph');
    if (!raw) return;
    if (el.__suiGraphJson === raw && el.__suiGraphBound) return;
    var cfg = null;
    try { cfg = JSON.parse(raw); } catch (e) { return; }
    if (!cfg || !cfg.states || !cfg.states.length) return;
    el.__suiGraphJson = raw;
    el.__suiGraphCfg = cfg;
    bind(el, cfg);
    var key = keyOf(el);
    var id = key && stateMemory[key] ? stateMemory[key] : (cfg.initial || cfg.states[0].id);
    if (!stateById(cfg, id)) id = cfg.initial || cfg.states[0].id;
    el.__suiGraphState = id;
    applyState(el, cfg, id, null, true);
  }
  function scan(root) {
    var scope = root && root.querySelectorAll ? root : document;
    var els = scope.querySelectorAll('[data-sui-motion-graph]');
    for (var i = 0; i < els.length; i++) init(els[i]);
  }
  if (typeof document !== 'undefined') {
    var schedule = function() {
      if (schedule.pending) return;
      schedule.pending = true;
      setTimeout(function() { schedule.pending = false; scan(document); }, 16);
    };
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', function() { scan(document); });
    } else {
      scan(document);
    }
    try {
      new MutationObserver(schedule).observe(document.body, { childList: true, subtree: true });
    } catch (e) {}
    setTimeout(function() { scan(document); }, 0);
    setInterval(function() {
      var els = document.querySelectorAll('[data-sui-motion-graph]');
      for (var i = 0; i < els.length; i++) {
        var el = els[i];
        if (el.__suiGraphCfg) checkStateLinks(el, el.__suiGraphCfg);
      }
    }, 120);
  }
  return { scan: scan, run: run, init: init };
})();
if (typeof window !== 'undefined') window.__SUI_MOTION_GRAPH = __SUI_MOTION_GRAPH;`
