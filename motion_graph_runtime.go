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
    if (!link || el.__suiGraphUnbinding) return;
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
  function bind(el, cfg) {
    if (el.__suiGraphBound) unbind(el);
    el.__suiGraphUnbinding = false;
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
    attachApi(el, cfg);
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
  return { scan: scan, run: run, init: init };
})();
if (typeof window !== 'undefined') window.__SUI_MOTION_GRAPH = __SUI_MOTION_GRAPH;`
