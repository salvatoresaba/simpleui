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
  function matchesSource(node, id) {
    if (!node || !node.getAttribute) return false;
    return node.getAttribute('data-sui-id') === id || node.getAttribute('data-sui-motion-key') === id;
  }
  function sourceById(id) {
    var s = String(id || '');
    if (!/^[A-Za-z0-9_.:-]+$/.test(s)) return null;
    var list = document.querySelectorAll('[data-sui-id],[data-sui-motion-key]');
    for (var i = 0; i < list.length; i++) if (matchesSource(list[i], s)) return list[i];
    return null;
  }
  function linkIterations(link) {
    var v = link && link.iterations;
    if (v === Infinity || String(v).toLowerCase() === 'infinite') return Infinity;
    var n = Number(v);
    if (!isFinite(n) || n < 1) return 1;
    return Math.min(1000, Math.floor(n));
  }
  function linkDirection(link) {
    var d = String(link && link.direction || 'normal').toLowerCase();
    if (d === 'reverse' || d === 'alternate' || d === 'alternate-reverse') return d;
    return 'normal';
  }
  function linkPlaybackRate(link) {
    var r = Number(link && link.playbackRate);
    if (!isFinite(r) || r <= 0) return 1;
    return Math.min(10, Math.max(0.1, r));
  }
  function clampNum(v, fallback, min, max) {
    var n = Number(v);
    if (!isFinite(n)) return fallback;
    return Math.min(max, Math.max(min, n));
  }
  function springEasing(spec, durationMs) {
    if (!spec) return '';
    var stiffness = clampNum(spec.stiffness, 170, 1, 1000);
    var damping = clampNum(spec.damping, 18, 0, 100);
    var mass = clampNum(spec.mass, 1, 0.1, 20);
    var seconds = Math.max(0.05, durationMs / 1000);
    var w0 = Math.sqrt(stiffness / mass);
    var zeta = damping / (2 * Math.sqrt(stiffness * mass));
    var wd = w0 * Math.sqrt(Math.max(0, 1 - zeta * zeta));
    function valueAt(t) {
      var s = t * seconds;
      if (zeta < 1 && wd > 0.001) {
        return 1 - Math.exp(-zeta * w0 * s) * (Math.cos(wd * s) + (zeta * w0 / wd) * Math.sin(wd * s));
      }
      return 1 - Math.exp(-w0 * s) * (1 + w0 * s);
    }
    var end = valueAt(1);
    if (!isFinite(end) || Math.abs(end) < 0.001) return 'cubic-bezier(0.34, 1.56, 0.64, 1)';
    var points = [];
    var steps = 24;
    for (var i = 0; i <= steps; i++) {
      var x = i === 0 ? 0 : (i === steps ? 1 : valueAt(i / steps) / end);
      points.push(Math.round(x * 1000) / 1000);
    }
    return 'linear(' + points.join(', ') + ')';
  }
  function scheduleFallback(el) {
    if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
    if (!el.__suiGraphFinishFn || el.__suiGraphInfinite || el.__suiGraphPaused) return;
    var anim = el.__suiGraphAnim;
    var total = Number(el.__suiGraphTotal);
    var current = anim ? Number(anim.currentTime) || 0 : 0;
    var remaining = isFinite(total) ? Math.max(0, total - current) : 0;
    el.__suiGraphFinishTimer = setTimeout(function() {
      if (el.__suiGraphPaused) { scheduleFallback(el); return; }
      if (el.__suiGraphFinishFn) el.__suiGraphFinishFn();
    }, remaining + 80);
  }
  function transitionDetail(link, el) {
    return {
      from: link && link.from ? link.from : '',
      to: link && link.to ? link.to : '',
      state: el ? el.__suiGraphState : (link && link.to) || '',
      link: link && link.id ? link.id : ''
    };
  }
  function emitEvent(el, name, detail) {
    if (!el || !name || !/^[A-Za-z][A-Za-z0-9_.:-]*$/.test(String(name))) return;
    var ev = null;
    try { ev = new CustomEvent(name, { detail: detail || {}, bubbles: true, composed: true }); } catch (e) { return; }
    try { el.dispatchEvent(ev); } catch (e) {}
  }
  function emitSpec(el, link, when, detail) {
    var spec = link && link.emit;
    if (!spec) return;
    var name = typeof spec === 'string' ? spec : spec.name;
    var on = typeof spec === 'string' ? 'end' : (spec.on || 'end');
    if (!name || on !== when) return;
    emitEvent(el, name, detail);
  }
  function stateById(cfg, id) {
    var list = cfg && cfg.states ? cfg.states : [];
    for (var i = 0; i < list.length; i++) if (list[i] && list[i].id === id) return list[i];
    return null;
  }
  // Safe guard expressions: identifiers (SUI states), booleans, numbers,
  // strings, parentheses, !, &&, ||, ==, !=, <, <=, >, >=. No eval and no
  // function calls.
  function guardTokens(src) {
    var out = [], i = 0;
    while (i < src.length) {
      var ch = src.charAt(i);
      if (/\s/.test(ch)) { i++; continue; }
      var three = src.slice(i, i + 3);
      if (three === '===' || three === '!==') { out.push({ t: 'op', v: three }); i += 3; continue; }
      var two = src.slice(i, i + 2);
      if (two === '&&' || two === '||' || two === '==' || two === '!=' || two === '<=' || two === '>=') {
        out.push({ t: 'op', v: two }); i += 2; continue;
      }
      if (ch === '!' || ch === '(' || ch === ')') {
        out.push({ t: ch, v: ch }); i++; continue;
      }
      if (ch === '<' || ch === '>') {
        out.push({ t: 'op', v: ch }); i++; continue;
      }
      if (ch === '"' || ch === "'") {
        var quote = ch, value = '';
        i++;
        while (i < src.length) {
          var c = src.charAt(i);
          if (c === quote) { i++; break; }
          if (c === '\\' && i + 1 < src.length) { value += src.charAt(i + 1); i += 2; continue; }
          value += c; i++;
        }
        out.push({ t: 'str', v: value }); continue;
      }
      var num = /^-?\d+(?:\.\d+)?/.exec(src.slice(i));
      if (num) { out.push({ t: 'num', v: parseFloat(num[0]) }); i += num[0].length; continue; }
      var id = /^[A-Za-z_][A-Za-z0-9_]*/.exec(src.slice(i));
      if (id) { out.push({ t: 'id', v: id[0] }); i += id[0].length; continue; }
      throw new Error('invalid guard');
    }
    return out;
  }
  function evalGuard(src) {
    var tokens = guardTokens(String(src || ''));
    var pos = 0;
    function peek() { return tokens[pos] || null; }
    function next() { return tokens[pos++]; }
    function isOp(v) { var t = peek(); return t && t.t === 'op' && t.v === v; }
    function primary() {
      var t = next();
      if (!t) throw new Error('missing value');
      if (t.t === '(') {
        var v = orExpr();
        if (!peek() || peek().t !== ')') throw new Error('missing close paren');
        next();
        return v;
      }
      if (t.t === 'num' || t.t === 'str') return t.v;
      if (t.t === 'id') {
        if (t.v === 'true') return true;
        if (t.v === 'false') return false;
        if (t.v === 'null') return null;
        var get = typeof window !== 'undefined' && window.__SUI_STATE_GET && window.__SUI_STATE_GET[t.v];
        return get ? get() : undefined;
      }
      throw new Error('unexpected token');
    }
    function unary() {
      if (peek() && peek().t === '!') { next(); return !unary(); }
      return primary();
    }
    function compare() {
      var left = unary();
      var t = peek();
      if (t && t.t === 'op' && (t.v === '==' || t.v === '===' || t.v === '!=' || t.v === '!==' || t.v === '<' || t.v === '<=' || t.v === '>' || t.v === '>=')) {
        next();
        var right = unary();
        if (t.v === '==' || t.v === '===') return left === right;
        if (t.v === '!=' || t.v === '!==') return left !== right;
        if (t.v === '<') return left < right;
        if (t.v === '<=') return left <= right;
        if (t.v === '>') return left > right;
        return left >= right;
      }
      return left;
    }
    function andExpr() {
      var value = compare();
      while (isOp('&&')) { next(); var right = compare(); value = !!value && !!right; }
      return value;
    }
    function orExpr() {
      var value = andExpr();
      while (isOp('||')) { next(); var right = andExpr(); value = !!value || !!right; }
      return value;
    }
    if (!tokens.length) return false;
    var result = orExpr();
    if (pos !== tokens.length) throw new Error('trailing guard tokens');
    return result;
  }
  function guardPass(link) {
    var raw = link && link.guard;
    if (!raw) return true;
    try { return !!evalGuard(String(raw)); } catch (e) { return false; }
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
    var iterations = linkIterations(link);
    var direction = linkDirection(link);
    var playbackRate = linkPlaybackRate(link);
    var spring = link && link.spring;
    var springUsed = false;
    if (spring) {
      var generated = springEasing(spring, duration);
      if (generated) { easing = generated; springUsed = true; }
    }
    if (reduced || instant || !link || duration <= 0) {
      duration = 0; delay = 0; iterations = 1; direction = 'normal'; playbackRate = 1;
    }
    var infinite = iterations === Infinity && duration > 0;
    var totalMs = infinite ? Infinity : (delay + duration * iterations) / playbackRate;
    var detail = transitionDetail(link, el);
    var token = (el.__suiGraphToken || 0) + 1;
    el.__suiGraphToken = token;
    el.__suiGraphPaused = false;
    el.__suiGraphInfinite = infinite;
    el.__suiGraphTotal = totalMs;
    if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
    if (link && !instant) emitEvent(el, 'sui:motionstate', detail);
    if (token !== el.__suiGraphToken) return;
    var settled = false;
    function finish() {
      if (token !== el.__suiGraphToken || settled) return;
      settled = true;
      if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
      var active = el.__suiGraphAnim;
      el.__suiGraphAnim = null;
      el.__suiGraphInfinite = false;
      el.__suiGraphPaused = false;
      for (var p = 0; p < all.length; p++) if (!next[all[p]]) el.style.removeProperty(all[p]);
      for (var d = 0; d < decls.length; d++) el.style.setProperty(decls[d].name, decls[d].value);
      el.__suiGraphProps = decls.map(function(x){ return x.name; });
      if (active) { try { active.cancel(); } catch (e) {} }
      if (link) {
        emitEvent(el, 'sui:motionend', detail);
        emitSpec(el, link, 'end', detail);
      }
      if (token !== el.__suiGraphToken) return;
      if (el.__suiGraphLink === link) el.__suiGraphLink = null;
      el.__suiGraphFinishFn = null;
      if (link && link.to) {
        el.__suiGraphAfterTimer = setTimeout(function() {
          el.__suiGraphAfterTimer = 0;
          if (token !== el.__suiGraphToken) return;
          fireAfter(el, cfg, link.to);
        }, 0);
      }
    }
    function cancelOld() {
      if (!el.__suiGraphAnim) return;
      var oldLink = el.__suiGraphLink;
      var oldDetail = transitionDetail(oldLink, null);
      try { el.__suiGraphAnim.cancel(); } catch (e) {}
      el.__suiGraphAnim = null;
      el.__suiGraphLink = null;
      if (oldLink) {
        emitEvent(el, 'sui:motioncancel', oldDetail);
        emitSpec(el, oldLink, 'cancel', oldDetail);
      }
    }
    function emitStart() {
      if (!link || instant) return;
      emitEvent(el, 'sui:motionstart', detail);
      emitSpec(el, link, 'start', detail);
    }
    if (duration <= 0 || !all.length) {
      cancelOld();
      if (token !== el.__suiGraphToken) return;
      emitStart();
      if (token !== el.__suiGraphToken) return;
      finish();
      return;
    }
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
    cancelOld();
    if (token !== el.__suiGraphToken) return;
    emitStart();
    if (token !== el.__suiGraphToken) return;
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
    var timing = {
      duration: duration, delay: delay, fill: 'both',
      iterations: iterations === Infinity ? Infinity : iterations,
      direction: direction, easing: easing
    };
    var anim = null;
    try {
      anim = el.animate([from, to], timing);
    } catch (err) {
      if (springUsed) {
        try {
          timing.easing = 'cubic-bezier(0.34, 1.56, 0.64, 1)';
          anim = el.animate([from, to], timing);
        } catch (err2) { anim = null; }
      }
      if (!anim) { finish(); return; }
    }
    try { anim.playbackRate = playbackRate; } catch (e) {}
    el.__suiGraphAnim = anim;
    el.__suiGraphLink = link;
    el.__suiGraphFinishFn = finish;
    if (infinite) {
      anim.onfinish = null;
      return;
    }
    scheduleFallback(el);
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
    if (link.wheel && link.wheel !== 'any' && e.deltaY != null) {
      var dy = Number(e.deltaY) || 0;
      if (link.wheel === 'down' && dy <= 0) return false;
      if (link.wheel === 'up' && dy >= 0) return false;
    }
    if (link.axis && e.axis && String(link.axis) !== String(e.axis)) return false;
    if (link.distance != null && link.distance !== '' && e.distance != null && Number(e.distance) < Number(link.distance)) return false;
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
      if (name !== eventName || !eventMatches(l, e) || !guardPass(l)) continue;
      var score = (l.key ? 2 : 0) + (l.button != null && l.button !== '' ? 2 : 0) +
        (l.ctrl ? 1 : 0) + (l.alt ? 1 : 0) + (l.shift ? 1 : 0) + (l.meta ? 1 : 0) +
        (l.guard ? 1 : 0) + (l.axis ? 1 : 0) + (l.wheel && l.wheel !== 'any' ? 1 : 0);
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
      var now = stateCondition(l) && guardPass(l);
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
      var now = !!(src && src.__suiGraphState && src.__suiGraphState === l.whenState) && guardPass(l);
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
          if (token !== el.__suiGraphToken || !guardPass(link)) return;
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
    if (!link || el.__suiGraphUnbinding || !guardPass(link)) return;
    var to = stateById(cfg, link.to);
    if (!to) return;
    clearTimers(el);
    if (el.__suiGraphAfterTimer) { clearTimeout(el.__suiGraphAfterTimer); el.__suiGraphAfterTimer = 0; }
    el.__suiGraphState = link.to;
    var key = keyOf(el);
    if (key) stateMemory[key] = link.to;
    var seq = (el.__suiGraphRunSeq || 0) + 1;
    el.__suiGraphRunSeq = seq;
    applyState(el, cfg, link.to, link, false);
    if (el.__suiGraphRunSeq !== seq) return;
    scheduleTimers(el, cfg, link.to);
  }
  function unbind(el) {
    el.__suiGraphUnbinding = true;
    if (el.__suiGraphGestureCleanup) {
      try { el.__suiGraphGestureCleanup(); } catch (e) {}
      el.__suiGraphGestureCleanup = null;
    }
    clearTimers(el);
    if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
    if (el.__suiGraphAfterTimer) { clearTimeout(el.__suiGraphAfterTimer); el.__suiGraphAfterTimer = 0; }
    var handlers = el.__suiGraphHandlers || [];
    for (var i = 0; i < handlers.length; i++) {
      var h = handlers[i];
      var target = h.target || el;
      target.removeEventListener(h.event, h.fn, !!h.capture);
    }
    el.__suiGraphHandlers = [];
    el.__suiGraphBound = false;
    if (el.__suiGraphAnim) {
      var oldLink = el.__suiGraphLink;
      var oldDetail = transitionDetail(oldLink, null);
      try { el.__suiGraphAnim.cancel(); } catch (e) {}
      el.__suiGraphAnim = null;
      if (oldLink) {
        emitEvent(el, 'sui:motioncancel', oldDetail);
        emitSpec(el, oldLink, 'cancel', oldDetail);
      }
    }
    el.__suiGraphInfinite = false;
    el.__suiGraphPaused = false;
    el.__suiGraphLink = null;
    el.__suiGraphFinishFn = null;
    delete el.__suiMotion;
  }
  function bindGestures(el, cfg, links) {
    var cleanups = [];
    var swipeLinks = [], longLinks = [], visLinks = [];
    for (var i = 0; i < links.length; i++) {
      var l = links[i];
      if (!l) continue;
      if (l.event === 'swipe') swipeLinks.push(l);
      else if (l.event === 'longpress') longLinks.push(l);
      else if (l.event === 'visible' || l.event === 'hidden') visLinks.push(l);
    }
    if (swipeLinks.length) {
      var swipeStart = null;
      var swipeDown = function(e) {
        if (e.button != null && e.button !== 0) return;
        swipeStart = { x: e.clientX, y: e.clientY, id: e.pointerId };
      };
      var swipeUp = function(e) {
        if (!swipeStart) return;
        if (swipeStart.id != null && e.pointerId != null && swipeStart.id !== e.pointerId) return;
        var dx = e.clientX - swipeStart.x, dy = e.clientY - swipeStart.y;
        swipeStart = null;
        var ax = Math.abs(dx), ay = Math.abs(dy);
        var distance = Math.max(ax, ay);
        if (distance < 10) return;
        var axis = ax >= ay ? (dx < 0 ? 'left' : 'right') : (dy < 0 ? 'up' : 'down');
        var found = findEventLink(cfg, el.__suiGraphState, 'swipe', '', {
          axis: axis, distance: distance, dx: dx, dy: dy, pointerType: e.pointerType
        });
        if (found) run(el, cfg, found);
      };
      var swipeCancel = function(){ swipeStart = null; };
      el.addEventListener('pointerdown', swipeDown, { passive: true });
      document.addEventListener('pointerup', swipeUp, true);
      document.addEventListener('pointercancel', swipeCancel, true);
      cleanups.push(function(){
        el.removeEventListener('pointerdown', swipeDown);
        document.removeEventListener('pointerup', swipeUp, true);
        document.removeEventListener('pointercancel', swipeCancel, true);
      });
    }
    if (longLinks.length) {
      var longStart = null, longTimers = [];
      function clearLong() {
        for (var t = 0; t < longTimers.length; t++) clearTimeout(longTimers[t]);
        longTimers = [];
        longStart = null;
      }
      var longDown = function(e) {
        if (e.button != null && e.button !== 0) return;
        clearLong();
        longStart = { x: e.clientX, y: e.clientY, id: e.pointerId };
        var from = el.__suiGraphState;
        for (var i = 0; i < longLinks.length; i++) {
          (function(link) {
            if (link.from !== from || !guardPass(link)) return;
            var hold = Math.max(0, timeMs(link.hold, 500));
            longTimers.push(setTimeout(function() {
              if (!longStart || el.__suiGraphState !== link.from || !guardPass(link)) return;
              clearLong();
              run(el, cfg, link);
            }, hold));
          })(longLinks[i]);
        }
      };
      var longUp = function(){ clearLong(); };
      var longMove = function(e) {
        if (!longStart) return;
        var dx = e.clientX - longStart.x, dy = e.clientY - longStart.y;
        if (dx * dx + dy * dy > 100) clearLong();
      };
      el.addEventListener('pointerdown', longDown, { passive: true });
      document.addEventListener('pointerup', longUp, true);
      document.addEventListener('pointercancel', longUp, true);
      document.addEventListener('pointermove', longMove, true);
      cleanups.push(function(){
        clearLong();
        el.removeEventListener('pointerdown', longDown);
        document.removeEventListener('pointerup', longUp, true);
        document.removeEventListener('pointercancel', longUp, true);
        document.removeEventListener('pointermove', longMove, true);
      });
    }
    if (visLinks.length) {
      var visState = el.__suiGraphVisState || (el.__suiGraphVisState = Object.create(null));
      function applyVisibility(entry) {
        var ratio = entry.intersectionRatio || 0;
        for (var i = 0; i < visLinks.length; i++) {
          (function(link) {
            var key = link.id || (link.event + '|' + link.to);
            var st = visState[key] || (visState[key] = { vis: false, fired: false });
            var fallback = link.event === 'hidden' ? 0.2 : 0.5;
            var threshold = link.threshold != null && link.threshold !== '' ? Number(link.threshold) : fallback;
            if (!isFinite(threshold)) threshold = fallback;
            var visNow = !!(entry.isIntersecting && ratio >= threshold);
            var wasVisible = st.vis;
            if (link.event === 'visible') {
              if (visNow && !wasVisible && !(link.once && st.fired) && link.from === el.__suiGraphState && guardPass(link)) {
                st.fired = true;
                run(el, cfg, link);
              }
            } else if (!visNow && wasVisible && !(link.once && st.fired) && link.from === el.__suiGraphState && guardPass(link)) {
              st.fired = true;
              run(el, cfg, link);
            }
            st.vis = visNow;
          })(visLinks[i]);
        }
      }
      // Expose the evaluator for deterministic tests and for hosts that
      // already know the visible ratio (for example a virtualized list).
      // It is an internal hook and is cleared when the graph is unbound.
      el.__suiGraphApplyVisibility = applyVisibility;
      cleanups.push(function() { el.__suiGraphApplyVisibility = null; });
      // Initial visibility check: the observer normally reports the first
      // entry after the next frame, but an enter animation should start as
      // soon as the element is mounted. The once flag keeps a later observer
      // entry from firing the same link twice.
      function initialVisibility() {
        if (!el.isConnected || typeof el.getBoundingClientRect !== 'function') return;
        var r = el.getBoundingClientRect();
        var win = window;
        var vw = win.innerWidth || document.documentElement.clientWidth || 0;
        var vh = win.innerHeight || document.documentElement.clientHeight || 0;
        var left = Math.max(0, r.left), top = Math.max(0, r.top);
        var right = Math.min(vw, r.right), bottom = Math.min(vh, r.bottom);
        var area = Math.max(0, r.width) * Math.max(0, r.height);
        var visible = Math.max(0, right - left) * Math.max(0, bottom - top);
        var ratio = area > 0 ? Math.min(1, visible / area) : 0;
        applyVisibility({ isIntersecting: visible > 0, intersectionRatio: ratio });
      }
      if (typeof IntersectionObserver !== 'undefined') {
        var observedFirst = false;
        var observer = new IntersectionObserver(function(entries) {
          observedFirst = true;
          for (var e = 0; e < entries.length; e++) applyVisibility(entries[e]);
        }, { threshold: [0, 0.1, 0.25, 0.5, 0.75, 0.9, 1] });
        observer.observe(el);
        var initialTimer = setTimeout(function() {
          if (observedFirst) return;
          initialVisibility();
        }, 0);
        cleanups.push(function(){ clearTimeout(initialTimer); observer.disconnect(); });
      } else {
        var measure = function() {
          if (el.isConnected) initialVisibility();
        };
        var fallbackTimer = setTimeout(measure, 0);
        var fallbackInterval = setInterval(measure, 350);
        document.addEventListener('scroll', measure, true);
        window.addEventListener('resize', measure);
        cleanups.push(function(){
          clearTimeout(fallbackTimer);
          clearInterval(fallbackInterval);
          document.removeEventListener('scroll', measure, true);
          window.removeEventListener('resize', measure);
        });
      }
    }
    if (cleanups.length) {
      el.__suiGraphGestureCleanup = function() {
        for (var i = 0; i < cleanups.length; i++) { try { cleanups[i](); } catch (e) {} }
      };
    }
  }
  function bind(el, cfg) {
    if (el.__suiGraphBound) unbind(el);
    el.__suiGraphUnbinding = false;
    var handlers = [];
    var seen = Object.create(null);
    var links = cfg && cfg.links ? cfg.links : [];
    for (var i = 0; i < links.length; i++) {
      var link = links[i];
      var rawEv = link && link.event;
      if (!rawEv || rawEv === 'after' || rawEv === 'state' || rawEv === 'timer' || rawEv === 'elementState' ||
          rawEv === 'swipe' || rawEv === 'longpress' || rawEv === 'visible' || rawEv === 'hidden') continue;
      var ev = rawEv === 'custom' ? String(link.customEvent || '') : rawEv;
      if (!ev || !/^[A-Za-z][A-Za-z0-9_.:-]*$/.test(ev)) continue;
      var sourceId = link.sourceId || '';
      if (sourceId && !/^[A-Za-z0-9_.:-]+$/.test(sourceId)) continue;
      var key = sourceId + '|' + ev;
      if (seen[key]) continue;
      seen[key] = 1;
      (function(eventName, source) {
        var fn = function(e) {
          if (el.__suiGraphEditLocked) return;
          if (source) {
            var t = e.target;
            if (!t || !t.getAttribute) return;
            var foundSource = false;
            var node = t;
            while (node && node !== document) {
              if (matchesSource(node, source)) { foundSource = true; break; }
              node = node.parentElement;
            }
            if (!foundSource) return;
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
    bindGestures(el, cfg, links);
    el.__suiGraphHandlers = handlers;
    el.__suiGraphBound = true;
  }
  function runTo(el, cfg, state, opts) {
    if (!el || !cfg || !stateById(cfg, state)) return false;
    opts = opts || {};
    var link = {
      id: 'api',
      from: el.__suiGraphState || '',
      to: state,
      duration: opts.duration != null ? opts.duration : '240ms',
      easing: opts.easing != null ? opts.easing : 'ease-out',
      delay: opts.delay != null ? opts.delay : '0ms'
    };
    if (opts.iterations != null) link.iterations = opts.iterations;
    if (opts.direction != null) link.direction = opts.direction;
    if (opts.playbackRate != null) link.playbackRate = opts.playbackRate;
    if (opts.spring) link.spring = opts.spring;
    if (opts.emit) link.emit = opts.emit;
    run(el, cfg, link);
    return true;
  }
  function attachApi(el, cfg) {
    el.__suiMotion = {
      to: function(state, opts) { return runTo(el, cfg, state, opts); },
      pause: function() {
        if (!el.__suiGraphAnim) return false;
        try { el.__suiGraphAnim.pause(); } catch (e) { return false; }
        el.__suiGraphPaused = true;
        if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
        return true;
      },
      resume: function() {
        if (!el.__suiGraphAnim) return false;
        try { el.__suiGraphAnim.play(); } catch (e) { return false; }
        el.__suiGraphPaused = false;
        scheduleFallback(el);
        return true;
      },
      seek: function(ms) {
        if (!el.__suiGraphAnim) return false;
        try { el.__suiGraphAnim.currentTime = Math.max(0, Number(ms) || 0); } catch (e) { return false; }
        scheduleFallback(el);
        return true;
      },
      get state() { return el.__suiGraphState || ''; },
      get paused() { return !!el.__suiGraphPaused; },
      get progress() {
        var anim = el.__suiGraphAnim;
        if (!anim) return el.__suiGraphState ? 1 : 0;
        var total = el.__suiGraphTotal;
        if (el.__suiGraphInfinite) {
          var duration = Math.max(1, timeMs(el.__suiGraphLink && el.__suiGraphLink.duration, 240) / linkPlaybackRate(el.__suiGraphLink));
          return ((Number(anim.currentTime) || 0) % duration) / duration;
        }
        if (!isFinite(total) || total <= 0) return 0;
        return Math.min(1, Math.max(0, (Number(anim.currentTime) || 0) / total));
      }
    };
    return el.__suiMotion;
  }
  var exitSnapshots = [];
  function stripExitClone(node) {
    if (!node || node.nodeType !== 1) return;
    node.setAttribute('aria-hidden', 'true');
    var all = [node];
    var found = node.querySelectorAll('*');
    for (var i = 0; i < found.length; i++) all.push(found[i]);
    for (var j = 0; j < all.length; j++) {
      var n = all[j];
      n.removeAttribute('id');
      n.removeAttribute('data-sui-id');
      var attrs = Array.prototype.slice.call(n.attributes || []);
      for (var a = 0; a < attrs.length; a++) {
        if (attrs[a].name.indexOf('data-sui-motion') === 0) n.removeAttribute(attrs[a].name);
      }
    }
  }
  function beforeReplace(root) {
    exitSnapshots = [];
    if (reduced || !root || !root.querySelectorAll) return;
    var els = root.querySelectorAll('[data-sui-motion-graph]');
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      var cfg = el.__suiGraphCfg;
      if (!cfg) {
        try { cfg = JSON.parse(el.getAttribute('data-sui-motion-graph') || 'null'); } catch (e) { cfg = null; }
      }
      var ex = cfg && cfg.exit;
      if (!ex || !ex.state || !stateById(cfg, ex.state)) continue;
      var key = keyOf(el);
      if (!key) continue;
      var r = el.getBoundingClientRect();
      if (!r.width && !r.height) continue;
      var clone = el.cloneNode(true);
      stripExitClone(clone);
      exitSnapshots.push({
        key: key, clone: clone, cfg: cfg, exit: ex, state: stateById(cfg, ex.state),
        rect: { left: r.left, top: r.top, width: r.width, height: r.height }
      });
    }
  }
  function afterReplace(root) {
    if (!exitSnapshots.length) return;
    var seen = Object.create(null);
    var els = root && root.querySelectorAll ? root.querySelectorAll('[data-sui-motion-graph]') : [];
    for (var i = 0; i < els.length; i++) {
      var k = keyOf(els[i]);
      if (k) seen[k] = 1;
    }
    var pending = exitSnapshots;
    exitSnapshots = [];
    for (var j = 0; j < pending.length; j++) {
      if (seen[pending[j].key]) continue;
      animateExitClone(pending[j]);
    }
  }
  function animateExitClone(snap) {
    var clone = snap.clone, state = snap.state, ex = snap.exit || {};
    var decls = parseDecls(state.css || '');
    if (!decls.length) return;
    var r = snap.rect;
    clone.classList.add('sui-motion-graph-exit');
    clone.style.position = 'fixed';
    clone.style.left = r.left + 'px';
    clone.style.top = r.top + 'px';
    clone.style.width = r.width + 'px';
    clone.style.height = r.height + 'px';
    clone.style.margin = '0';
    clone.style.boxSizing = 'border-box';
    clone.style.pointerEvents = 'none';
    clone.style.zIndex = '9998';
    clone.style.transform = 'none';
    document.body.appendChild(clone);
    var starts = {}, was = {}, prio = {};
    for (var a = 0; a < decls.length; a++) {
      var p0 = decls[a].name;
      was[p0] = clone.style.getPropertyValue(p0);
      prio[p0] = clone.style.getPropertyPriority(p0);
      starts[p0] = getComputedStyle(clone).getPropertyValue(p0);
    }
    for (var b = 0; b < decls.length; b++) clone.style.setProperty(decls[b].name, decls[b].value);
    var ends = {};
    for (var c = 0; c < decls.length; c++) ends[decls[c].name] = getComputedStyle(clone).getPropertyValue(decls[c].name);
    for (var d = 0; d < decls.length; d++) {
      var pd = decls[d].name;
      if (was[pd]) clone.style.setProperty(pd, was[pd], prio[pd]);
      else clone.style.removeProperty(pd);
    }
    var duration = Math.max(0, timeMs(ex.duration, 240));
    var delay = Math.max(0, timeMs(ex.delay, 0));
    var easing = ex.easing || 'ease-out';
    if (ex.spring) {
      var generated = springEasing(ex.spring, duration);
      if (generated) easing = generated;
    }
    function removeClone() {
      if (clone.parentNode) clone.remove();
    }
    if (reduced || duration <= 0 || typeof clone.animate !== 'function') {
      for (var e = 0; e < decls.length; e++) clone.style.setProperty(decls[e].name, decls[e].value);
      setTimeout(removeClone, 20);
      return;
    }
    var from = {}, to = {};
    for (var f = 0; f < decls.length; f++) {
      var pf = decls[f].name;
      var key = pf.indexOf('--') === 0 ? pf : pf.replace(/-([a-z])/g, function(m, ch){ return ch.toUpperCase(); });
      from[key] = starts[pf];
      to[key] = ends[pf];
    }
    var anim = null;
    try {
      anim = clone.animate([from, to], { duration: duration, delay: delay, easing: easing, fill: 'both' });
    } catch (err) {
      for (var g = 0; g < decls.length; g++) clone.style.setProperty(decls[g].name, decls[g].value);
      setTimeout(removeClone, 20);
      return;
    }
    var done = false;
    function finishExit() {
      if (done) return;
      done = true;
      if (timer) clearTimeout(timer);
      for (var h = 0; h < decls.length; h++) clone.style.setProperty(decls[h].name, decls[h].value);
      try { anim.cancel(); } catch (err2) {}
      removeClone();
    }
    var timer = setTimeout(finishExit, duration + delay + 80);
    anim.onfinish = finishExit;
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
    // Resolve the starting state before binding: IntersectionObserver can
    // deliver its first entry in the same task, and visibility links compare
    // against the element's current state.
    var key = keyOf(el);
    var id = key && stateMemory[key] ? stateMemory[key] : (cfg.initial || cfg.states[0].id);
    if (!stateById(cfg, id)) id = cfg.initial || cfg.states[0].id;
    el.__suiGraphState = id;
    bind(el, cfg);
    attachApi(el, cfg);
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
  if (typeof window !== 'undefined') {
    window.SimpleUIMotion = {
      get: function(target) {
        if (typeof target === 'string') {
          try { target = document.querySelector(target); } catch (e) { return null; }
        }
        return target && target.__suiMotion ? target.__suiMotion : null;
      },
      to: function(target, state, opts) {
        var api = this.get(target);
        return api ? api.to(state, opts) : false;
      }
    };
  }
  return { scan: scan, run: run, init: init, beforeReplace: beforeReplace, afterReplace: afterReplace };
})();
if (typeof window !== 'undefined') window.__SUI_MOTION_GRAPH = __SUI_MOTION_GRAPH;`
