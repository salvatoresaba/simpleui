package main

// motion_runtime.go - state-to-state motion for elements and control presets.
//
// Elements opt in with data-sui-motion. Before a view is replaced the runtime
// snapshots the element's geometry, selected CSS properties and the custom
// properties listed in data-sui-motion-vars. After the replacement it tweens
// those custom properties frame by frame (colors, lengths, radii, positions),
// applies a FLIP transform for geometry changes and adds the requested enter
// effect (fade, scale, slide). This is what makes switch/checkbox shape and
// theme changes morph instead of jumping.
const suiMotionRuntime = `const __SUI_MOTION = (function(){
  var beforeMap = {};
  var preMap = {};
  var hadSnapshot = false;
  var probe = null;
  var STANDARD = ['opacity', 'backgroundColor', 'color', 'borderColor', 'borderRadius', 'width', 'height', 'transform', 'filter', 'letterSpacing', 'fontSize', 'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft', 'marginTop', 'marginRight', 'marginBottom', 'marginLeft', 'boxShadow'];

  function closest(el, sel) { return el && el.closest ? el.closest(sel) : null; }
  function keyOf(el) {
    return el.getAttribute('data-sui-motion-key') || el.getAttribute('data-sui-id') || '';
  }
  function motionTarget(t) {
    var el = closest(t, '[data-sui-motion]');
    if (el) return el;
    var wrap = closest(t, '.sui-checkbox-wrap, .sui-switch-wrap, .sui-radio-wrap');
    if (wrap) return wrap.querySelector('[data-sui-motion]');
    return null;
  }
  function varList(el) {
    var raw = el.getAttribute('data-sui-motion-vars') || '';
    if (!raw) return [];
    return raw.split(',').map(function(s){ return s.trim(); }).filter(Boolean);
  }
  function timeValue(v, fallback) {
    if (v == null || v === '') return fallback;
    var s = String(v).trim();
    if (/^[0-9]+$/.test(s)) return parseInt(s, 10);
    var m = /^([0-9]*\.?[0-9]+)(ms|s)$/.exec(s);
    if (!m) return fallback;
    return m[2] === 's' ? parseFloat(m[1]) * 1000 : parseFloat(m[1]);
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
      if (p.length === 4) return cubicBezier(t, p[0], p[1], p[2], p[3]);
    }
    return 1 - (1 - t) * (1 - t);
  }
  function cubicBezier(t, x1, y1, x2, y2) {
    function sx(u) { return 3 * (1 - u) * (1 - u) * u * x1 + 3 * (1 - u) * u * u * x2 + u * u * u; }
    function sy(u) { return 3 * (1 - u) * (1 - u) * u * y1 + 3 * (1 - u) * u * u * y2 + u * u * u; }
    var lo = 0, hi = 1;
    for (var i = 0; i < 20; i++) {
      var mid = (lo + hi) / 2;
      if (sx(mid) < t) lo = mid; else hi = mid;
    }
    return sy((lo + hi) / 2);
  }
  function normalizeColor(v) {
    if (!probe) {
      probe = document.createElement('span');
      probe.style.display = 'none';
      document.body.appendChild(probe);
    }
    probe.style.color = '';
    probe.style.color = v;
    return getComputedStyle(probe).color || '';
  }
  function parseValue(v) {
    if (v == null) return null;
    var s = String(v).trim();
    if (!s) return null;
    var m = /^(-?\d*\.?\d+)([a-z%]*)$/i.exec(s);
    if (m) return { kind: 'num', n: parseFloat(m[1]), unit: m[2] || '' };
    var c = normalizeColor(s);
    var cm = /^rgba?\(([0-9.,\s]+)\)$/i.exec(c);
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
  function snapshot(el) {
    var cs = getComputedStyle(el);
    var vars = {};
    var names = varList(el);
    for (var i = 0; i < names.length; i++) vars[names[i]] = cs.getPropertyValue(names[i]).trim();
    var props = null;
    if (!names.length) {
      props = {};
      for (var j = 0; j < STANDARD.length; j++) props[STANDARD[j]] = cs[STANDARD[j]];
    }
    var r = el.getBoundingClientRect();
    return { vars: vars, props: props, rect: { left: r.left, top: r.top, width: r.width, height: r.height } };
  }
  function optionsOf(el) {
    return {
      effect: String(el.getAttribute('data-sui-motion-effect') || 'auto').toLowerCase(),
      duration: Math.max(1, timeValue(el.getAttribute('data-sui-motion-duration'), 220)),
      easing: el.getAttribute('data-sui-motion-easing') || 'ease-out',
      delay: Math.max(0, timeValue(el.getAttribute('data-sui-motion-delay'), 0)),
      stagger: Math.max(0, Math.min(0.5, parseFloat(el.getAttribute('data-sui-motion-stagger') || '0') || 0))
    };
  }
  function cancel(el) {
    if (el.__suiMotionRaf) { clearTimeout(el.__suiMotionRaf); el.__suiMotionRaf = 0; }
    if (el.__suiMotionVars) {
      for (var i = 0; i < el.__suiMotionVars.length; i++) el.style.removeProperty(el.__suiMotionVars[i]);
      el.__suiMotionVars = null;
    }
    if (el.__suiMotionAnim) { try { el.__suiMotionAnim.cancel(); } catch (e) {} el.__suiMotionAnim = null; }
  }
  function tweenVars(el, old, o) {
    var names = varList(el);
    if (!names.length) return;
    var cs = getComputedStyle(el);
    var items = [];
    for (var i = 0; i < names.length; i++) {
      var to = cs.getPropertyValue(names[i]).trim();
      var a = parseValue(old.vars[names[i]]);
      var b = parseValue(to);
      if (a && b) items.push({ name: names[i], a: a, b: b });
    }
    if (!items.length) return;
    var start = performance.now() + o.delay;
    var n = items.length;
    var spread = o.stagger;
    var step = n > 1 ? spread / (n - 1) : 0;
    var span = Math.max(0.05, 1 - spread);
    el.__suiMotionVars = items.map(function(it){ return it.name; });
    function frame() {
      var t = (performance.now() - start) / o.duration;
      for (var k = 0; k < items.length; k++) {
        var local = Math.max(0, Math.min(1, (t - k * step) / span));
        var val = interpolate(items[k].a, items[k].b, ease(local, o.easing));
        if (val != null) el.style.setProperty(items[k].name, val);
      }
      if (t < 1) {
        el.__suiMotionRaf = setTimeout(frame, 16);
      } else {
        el.__suiMotionRaf = 0;
        el.__suiMotionVars = null;
        for (var v = 0; v < items.length; v++) el.style.removeProperty(items[v].name);
      }
    }
    el.__suiMotionRaf = setTimeout(frame, 16);
  }
  function animateElement(old, el, o) {
    var r = el.getBoundingClientRect();
    var dx = old.rect.left - r.left;
    var dy = old.rect.top - r.top;
    var sx = r.width ? old.rect.width / r.width : 1;
    var sy = r.height ? old.rect.height / r.height : 1;
    var from = {};
    var to = {};
    if (old.props) {
      var ns = getComputedStyle(el);
      for (var i = 0; i < STANDARD.length; i++) {
        var p = STANDARD[i];
        var nv = ns[p];
        if (nv === undefined) continue;
        if (String(old.props[p]) !== String(nv)) { from[p] = old.props[p]; to[p] = nv; }
      }
    }
    var transform = '';
    if (Math.abs(dx) > 0.5 || Math.abs(dy) > 0.5 || Math.abs(sx - 1) > 0.01 || Math.abs(sy - 1) > 0.01) {
      transform += ' translate(' + dx + 'px,' + dy + 'px) scale(' + sx + ',' + sy + ')';
    }
    if (o.effect === 'scale') transform += ' scale(.96)';
    if (o.effect === 'slide') transform += ' translateY(6px)';
    if (transform) { from.transform = transform.trim(); to.transform = to.transform || 'none'; }
    if (o.effect === 'fade' || o.effect === 'scale' || o.effect === 'slide') { from.opacity = 0; to.opacity = to.opacity || 1; }
    if (!Object.keys(from).length) return;
    el.__suiMotionAnim = el.animate([from, to], { duration: o.duration, easing: o.easing, delay: o.delay, fill: 'none' });
  }
  function enter(el, o) {
    if (o.effect === 'none') return;
    var from = { opacity: 0 };
    if (o.effect === 'scale') from.transform = 'scale(.96)';
    if (o.effect === 'slide') from.transform = 'translateY(6px)';
    var to = { opacity: 1, transform: 'none' };
    el.__suiMotionAnim = el.animate([from, to], { duration: o.duration, easing: o.easing, delay: o.delay, fill: 'none' });
  }
  function before(root) {
    beforeMap = {};
    hadSnapshot = false;
    var els = root.querySelectorAll('[data-sui-motion]');
    if (!els.length) return;
    hadSnapshot = true;
    for (var i = 0; i < els.length; i++) {
      var k = keyOf(els[i]);
      if (!k) continue;
      if (preMap[k]) { beforeMap[k] = preMap[k]; delete preMap[k]; }
      else beforeMap[k] = snapshot(els[i]);
    }
  }
  function after(root) {
    var els = root.querySelectorAll('[data-sui-motion]');
    if (!els.length) { beforeMap = {}; return; }
    if (!hadSnapshot) { beforeMap = {}; return; }
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      var o = optionsOf(el);
      if (o.effect === 'none') continue;
      var old = beforeMap[keyOf(el)];
      cancel(el);
      if (old) {
        animateElement(old, el, o);
        tweenVars(el, old, o);
      } else {
        enter(el, o);
      }
    }
    beforeMap = {};
  }
  function captureBeforeToggle(ev) {
    var el = motionTarget(ev.target);
    if (!el) return;
    var k = keyOf(el);
    if (k) preMap[k] = snapshot(el);
  }
  if (typeof document !== 'undefined') {
    document.addEventListener('pointerdown', captureBeforeToggle, true);
    document.addEventListener('keydown', function(ev) {
      if (ev.key === ' ' || ev.key === 'Enter') captureBeforeToggle(ev);
    }, true);
  }
  return { before: before, after: after, cancel: cancel };
})();`
