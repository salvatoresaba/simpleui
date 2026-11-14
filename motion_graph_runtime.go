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
  function sourceById(id) {
    var s = String(id || '');
    if (!/^[A-Za-z0-9_-]+$/.test(s)) return null;
    var list = document.querySelectorAll('[data-sui-id]');
    for (var i = 0; i < list.length; i++) if (list[i].getAttribute('data-sui-id') === s) return list[i];
    return null;
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
  function applyState(el, cfg, id, link, instant) {
    var state = stateById(cfg, id);
    if (!state) return;
    var decls = parseDecls(state.css || '');
    var next = Object.create(null);
    for (var i = 0; i < decls.length; i++) next[decls[i].name] = 1;
    var prev = el.__suiGraphProps || [];
    var all = prev.slice();
    for (var j = 0; j < decls.length; j++) if (all.indexOf(decls[j].name) < 0) all.push(decls[j].name);
    // Remember every property touched by the graph, including those from a
    // transition that may be interrupted. The next transition starts from the
    // current computed values and removes any carried property not in its target.
    el.__suiGraphProps = all.slice();
    var duration = Math.max(0, timeMs(link && link.duration, 240));
    var delay = Math.max(0, timeMs(link && link.delay, 0));
    var easing = (link && link.easing) || 'ease-out';
    if (reduced || instant || !link || duration <= 0) { duration = 0; delay = 0; }
    var token = (el.__suiGraphToken || 0) + 1;
    el.__suiGraphToken = token;
    if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
    function finish() {
      if (token !== el.__suiGraphToken) return;
      if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
      var active = el.__suiGraphAnim;
      el.__suiGraphAnim = null;
      for (var p = 0; p < all.length; p++) if (!next[all[p]]) el.style.removeProperty(all[p]);
      for (var d = 0; d < decls.length; d++) el.style.setProperty(decls[d].name, decls[d].value);
      el.__suiGraphProps = decls.map(function(x){ return x.name; });
      if (active) { try { active.cancel(); } catch (e) {} }
    }
    if (duration <= 0 || !all.length) { finish(); return; }
    // Evaluate both measured states before starting: compute the start values,
    // temporarily apply the target to read its computed end values, then put
    // the start values back and let the browser animate between them.
    var starts = {}, was = {}, explicit = Object.create(null);
    for (var a = 0; a < all.length; a++) {
      var p0 = all[a];
      was[p0] = el.style.getPropertyValue(p0);
      starts[p0] = getComputedStyle(el).getPropertyValue(p0);
      explicit[p0] = el.style.getPropertyPriority(p0);
    }
    if (el.__suiGraphAnim) { try { el.__suiGraphAnim.cancel(); } catch (e) {} el.__suiGraphAnim = null; }
    for (var r = 0; r < all.length; r++) if (!next[all[r]]) el.style.removeProperty(all[r]);
    for (var t = 0; t < decls.length; t++) el.style.setProperty(decls[t].name, decls[t].value);
    var ends = {};
    for (var e = 0; e < all.length; e++) ends[all[e]] = getComputedStyle(el).getPropertyValue(all[e]);
    for (var b = 0; b < all.length; b++) {
      var pb = all[b];
      if (was[pb]) el.style.setProperty(pb, was[pb], explicit[pb]);
      else el.style.removeProperty(pb);
    }
    if (typeof el.animate !== 'function') { finish(); return; }
    var from = {}, to = {};
    for (var c = 0; c < all.length; c++) {
      var pc = all[c];
      var key = pc.indexOf('--') === 0 ? pc : pc.replace(/-([a-z])/g, function(m, ch){ return ch.toUpperCase(); });
      from[key] = starts[pc];
      to[key] = ends[pc];
    }
    var anim = null;
    try {
      anim = el.animate([from, to], { duration: duration, delay: delay, easing: easing, fill: 'both' });
    } catch (err) { finish(); return; }
    el.__suiGraphAnim = anim;
    el.__suiGraphFinishTimer = setTimeout(function() {
      if (token !== el.__suiGraphToken) return;
      finish();
    }, duration + delay + 50);
    anim.onfinish = function() {
      if (token !== el.__suiGraphToken) return;
      if (el.__suiGraphAnim === anim) el.__suiGraphAnim = null;
      finish();
      try { anim.cancel(); } catch (e) {}
    };
  }
  function findLink(cfg, from, event, sourceId) {
    var want = sourceId || '';
    var links = cfg && cfg.links ? cfg.links : [];
    for (var i = 0; i < links.length; i++) {
      var l = links[i];
      if (l && l.from === from && l.event === event && (l.sourceId || '') === want) return l;
    }
    return null;
  }
  function eventMatches(link, e) {
    if (!e) return true;
    if (link.key && String(link.key) !== String(e.key || '')) return false;
    if (link.ctrl && !e.ctrlKey) return false;
    if (link.alt && !e.altKey) return false;
    if (link.shift && !e.shiftKey) return false;
    if (link.meta && !e.metaKey) return false;
    if (link.button != null && link.button !== '') {
      var b = Number(link.button);
      if (!isNaN(b) && b >= 0 && e.button !== b) return false;
    }
    return true;
  }
  function findEventLink(cfg, from, eventName, sourceId, e) {
    var want = sourceId || '';
    var links = cfg && cfg.links ? cfg.links : [];
    var best = null, bestScore = -1;
    for (var i = 0; i < links.length; i++) {
      var l = links[i];
      if (!l || l.from !== from || (l.sourceId || '') !== want) continue;
      var name = l.event === 'custom' ? l.customEvent : l.event;
      if (name !== eventName || !eventMatches(l, e)) continue;
      var score = (l.key ? 2 : 0) + (l.button != null && l.button !== '' ? 2 : 0) +
        (l.ctrl ? 1 : 0) + (l.alt ? 1 : 0) + (l.shift ? 1 : 0) + (l.meta ? 1 : 0);
      if (score > bestScore) { best = l; bestScore = score; }
    }
    return best;
  }
  function stateCondition(link) {
    var w = String(link && link.when || '').trim();
    var neg = false;
    if (w.charAt(0) === '!') { neg = true; w = w.slice(1).trim(); }
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(w)) return false;
    var get = window.__SUI_STATE_GET && window.__SUI_STATE_GET[w];
    if (!get) return false;
    var v;
    try { v = get(); } catch (e) { return false; }
    var expected = link && link.value != null ? String(link.value) : '';
    var active;
    if (expected === '') {
      active = !!v;
    } else if (typeof v === 'boolean') {
      active = (expected.toLowerCase() === 'true') === v;
    } else if (typeof v === 'number') {
      active = v === Number(expected);
    } else {
      active = String(v) === expected;
    }
    return neg ? !active : active;
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
  function checkExternalStateLinks(el, cfg) {
    if (!el.isConnected || el.__suiGraphEditLocked) return;
    var seen = el.__suiGraphSeenElement || (el.__suiGraphSeenElement = Object.create(null));
    var links = cfg && cfg.links ? cfg.links : [];
    for (var i = 0; i < links.length; i++) {
      var l = links[i];
      if (!l || l.event !== 'elementState' || l.from !== el.__suiGraphState) continue;
      var src = sourceById(l.sourceId);
      var now = !!(src && src.__suiGraphState && src.__suiGraphState === l.whenState);
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
  function clearTimers(el) {
    var timers = el.__suiGraphTimers || [];
    for (var i = 0; i < timers.length; i++) clearTimeout(timers[i]);
    el.__suiGraphTimers = [];
  }
  function scheduleTimers(el, cfg, from) {
    clearTimers(el);
    var links = cfg && cfg.links ? cfg.links : [];
    var token = el.__suiGraphToken;
    for (var i = 0; i < links.length; i++) {
      var l = links[i];
      if (!l || l.event !== 'timer' || l.from !== from) continue;
      (function(link) {
        var wait = Math.max(0, timeMs(link.timer, 1000));
        var id = setTimeout(function() {
          if (token !== el.__suiGraphToken) return;
          run(el, cfg, link);
        }, wait);
        el.__suiGraphTimers.push(id);
      })(l);
    }
  }
  // The state changes immediately when a link fires. An event that arrives
  // during the tween is evaluated against the new state and starts a new tween
  // from the current computed values (latest event wins); it is not queued.
  // A new run invalidates the previous token, timers and pending after link, so
  // an interrupted transition cannot leave stale graph properties behind.
  function run(el, cfg, link) {
    if (!link) return;
    var to = stateById(cfg, link.to);
    if (!to) return;
    clearTimers(el);
    el.__suiGraphState = link.to;
    var key = keyOf(el);
    if (key) stateMemory[key] = link.to;
    applyState(el, cfg, link.to, link, false);
    scheduleTimers(el, cfg, link.to);
    var wait = Math.max(0, timeMs(link.duration, 240)) + Math.max(0, timeMs(link.delay, 0));
    var token = el.__suiGraphToken;
    setTimeout(function() {
      if (token !== el.__suiGraphToken) return;
      fireAfter(el, cfg, link.to);
    }, wait + 60);
  }
  function unbind(el) {
    clearTimers(el);
    var handlers = el.__suiGraphHandlers || [];
    for (var i = 0; i < handlers.length; i++) {
      var h = handlers[i];
      var target = h.target || el;
      target.removeEventListener(h.event, h.fn, !!h.capture);
    }
    el.__suiGraphHandlers = [];
    el.__suiGraphBound = false;
  }
  function bind(el, cfg) {
    if (el.__suiGraphBound) unbind(el);
    var handlers = [];
    var seen = Object.create(null);
    var links = cfg && cfg.links ? cfg.links : [];
    for (var i = 0; i < links.length; i++) {
      var link = links[i];
      var rawEv = link && link.event;
      if (!rawEv || rawEv === 'after' || rawEv === 'state' || rawEv === 'timer' || rawEv === 'elementState') continue;
      var ev = rawEv === 'custom' ? String(link.customEvent || '') : rawEv;
      if (!ev || !/^[A-Za-z][A-Za-z0-9_.:-]*$/.test(ev)) continue;
      var sourceId = link.sourceId || '';
      if (sourceId && !/^[A-Za-z0-9_-]+$/.test(sourceId)) continue;
      var key = sourceId + '|' + ev;
      if (seen[key]) continue;
      seen[key] = 1;
      (function(eventName, source) {
        var fn = function(e) {
          if (el.__suiGraphEditLocked) return;
          if (source) {
            var t = e.target;
            if (!t || !t.closest) return;
            var src = t.closest('[data-sui-id]');
            if (!src || src.getAttribute('data-sui-id') !== source) return;
          }
          var found = findEventLink(cfg, el.__suiGraphState, eventName, source, e);
          if (found) run(el, cfg, found);
        };
        var target = source ? document : el;
        var capture = !!source;
        target.addEventListener(eventName, fn, capture);
        handlers.push({ event: eventName, fn: fn, target: target, capture: capture });
      })(ev, sourceId);
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
    scheduleTimers(el, cfg, id);
    // Apply the remembered state and evaluate state triggers in the same DOM
    // mutation cycle: the browser has not painted the re-rendered base element
    // yet, so a switch change never flashes the default size.
    checkStateLinks(el, cfg);
    checkExternalStateLinks(el, cfg);
  }
  function scan(root) {
    var scope = root && root.querySelectorAll ? root : document;
    var els = scope.querySelectorAll('[data-sui-motion-graph]');
    for (var i = 0; i < els.length; i++) init(els[i]);
  }
  if (typeof document !== 'undefined') {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', function() { scan(document); });
    } else {
      scan(document);
    }
    try {
      // MutationObserver callbacks run before the browser paints. Scan here
      // instead of waiting for a timer, so re-rendered graph elements receive
      // their remembered state in the same frame.
      new MutationObserver(function() { scan(document); }).observe(document.body, { childList: true, subtree: true });
    } catch (e) {}
    setTimeout(function() { scan(document); }, 0);
    setInterval(function() {
      var els = document.querySelectorAll('[data-sui-motion-graph]');
      for (var i = 0; i < els.length; i++) {
        var el = els[i];
        if (el.__suiGraphCfg) {
          checkStateLinks(el, el.__suiGraphCfg);
          checkExternalStateLinks(el, el.__suiGraphCfg);
        }
      }
    }, 120);
  }
  return { scan: scan, run: run, init: init };
})();
if (typeof window !== 'undefined') window.__SUI_MOTION_GRAPH = __SUI_MOTION_GRAPH;`
