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
  var sharedMap = {};
  var pageSwitching = false;
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
  function snapshot(el, withClone) {
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
    var snap = { vars: vars, props: props, rect: { left: r.left, top: r.top, width: r.width, height: r.height }, opts: optionsOf(el), el: el, clone: null };
    if (withClone !== false && !names.length && el.getAttribute('data-sui-motion-exit') === '1') {
      snap.clone = el.cloneNode(true);
      stripMotionClone(snap.clone);
    }
    return snap;
  }
  function stripMotionClone(node) {
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
    if (el.__suiMotionRaf) {
      if (el.__suiMotionRafType === 'raf' && el.__suiMotionRaf.raf) {
        cancelAnimationFrame(el.__suiMotionRaf.raf);
        clearTimeout(el.__suiMotionRaf.timer);
      } else {
        clearTimeout(el.__suiMotionRaf);
      }
      el.__suiMotionRaf = 0;
    }
    el.__suiMotionRafType = '';
    if (el.__suiMotionVars) {
      for (var i = 0; i < el.__suiMotionVars.length; i++) el.style.removeProperty(el.__suiMotionVars[i]);
      el.__suiMotionVars = null;
    }
    if (el.__suiMotionProps) {
      el.style.removeProperty('opacity');
      el.style.removeProperty('transform');
      el.style.removeProperty('width');
      el.style.removeProperty('height');
      el.__suiMotionProps = false;
    }
    if (el.__suiMotionAnim) { try { el.__suiMotionAnim.cancel(); } catch (e) {} el.__suiMotionAnim = null; }
  }
  function scheduleFrame(el, fn) {
    // requestAnimationFrame is smoother in a visible browser. A timer fallback
    // keeps the tween moving when the tab is hidden or the driver throttles
    // rAF. Tests can force timer mode with ?motion=timer.
    var forceTimer = (typeof window !== 'undefined' && window.__SUI_MOTION_TIMER) ||
      (typeof location !== 'undefined' && location.search.indexOf('motion=timer') >= 0);
    var useRaf = typeof requestAnimationFrame === 'function' && !forceTimer && !(typeof navigator !== 'undefined' && navigator.webdriver);
    if (!useRaf) {
      el.__suiMotionRafType = 'timer';
      el.__suiMotionRaf = setTimeout(fn, 16);
      return;
    }
    var fired = false;
    function run() {
      if (fired) return;
      fired = true;
      if (el.__suiMotionRaf && el.__suiMotionRaf.raf) cancelAnimationFrame(el.__suiMotionRaf.raf);
      if (el.__suiMotionRaf && el.__suiMotionRaf.timer) clearTimeout(el.__suiMotionRaf.timer);
      el.__suiMotionRaf = 0;
      el.__suiMotionRafType = '';
      fn();
    }
    el.__suiMotionRafType = 'raf';
    el.__suiMotionRaf = { raf: requestAnimationFrame(run), timer: setTimeout(run, 34) };
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
    var effectProps = o.effect === 'fade' || o.effect === 'scale' || o.effect === 'slide';
    function applyEffect(t) {
      if (!effectProps) return;
      var e = ease(Math.max(0, Math.min(1, t)), o.easing);
      if (o.effect === 'fade') {
        // Never go fully transparent: a state toggle should stay readable.
        el.style.opacity = 0.35 + 0.65 * e;
        el.style.transform = 'none';
      } else if (o.effect === 'scale') {
        // Scale is a real, visible size change without fading out.
        el.style.opacity = '1';
        el.style.transform = 'scale(' + (0.82 + 0.18 * e) + ')';
      } else if (o.effect === 'slide') {
        el.style.opacity = 0.4 + 0.6 * e;
        el.style.transform = 'translateY(' + ((1 - e) * 10) + 'px)';
      }
    }
    el.__suiMotionVars = items.map(function(it){ return it.name; });
    el.__suiMotionProps = effectProps;
    // Apply the first frame synchronously: without this the browser paints the
    // new state for one frame before the tween rewinds to the old value, which
    // is visible as flicker.
    for (var init = 0; init < items.length; init++) {
      el.style.setProperty(items[init].name, interpolate(items[init].a, items[init].b, 0));
    }
    applyEffect(0);
    function frame() {
      var t = (performance.now() - start) / o.duration;
      for (var k = 0; k < items.length; k++) {
        var local = Math.max(0, Math.min(1, (t - k * step) / span));
        var val = interpolate(items[k].a, items[k].b, ease(local, o.easing));
        if (val != null) el.style.setProperty(items[k].name, val);
      }
      applyEffect(t);
      if (t < 1) {
        scheduleFrame(el, frame);
      } else {
        el.__suiMotionRaf = 0;
        el.__suiMotionRafType = '';
        el.__suiMotionVars = null;
        for (var v = 0; v < items.length; v++) el.style.removeProperty(items[v].name);
        if (el.__suiMotionProps) {
          el.style.removeProperty('opacity');
          el.style.removeProperty('transform');
          el.__suiMotionProps = false;
        }
      }
    }
    scheduleFrame(el, frame);
  }
  function animateMovement(el, dx, dy, o, fromOpacity, toOpacity, fromScale, toScale, fromY, toY, done, size) {
    var has = Math.abs(dx) > 0.5 || Math.abs(dy) > 0.5 || fromOpacity !== toOpacity || fromScale !== toScale || fromY !== toY;
    var hasSize = !!(size && (Math.abs(size.fromW - size.toW) > 0.5 || Math.abs(size.fromH - size.toH) > 0.5));
    var useOpacity = fromOpacity !== 1 || toOpacity !== 1;
    if (!has && !hasSize) { if (done) done(); return; }
    var start = performance.now() + o.delay;
    el.__suiMotionProps = true;
    function apply(t) {
      var e = ease(Math.max(0, Math.min(1, t)), o.easing);
      var tx = dx * (1 - e);
      var ty = dy * (1 - e);
      var sc = fromScale + (toScale - fromScale) * e;
      var yy = fromY + (toY - fromY) * e;
      var op = fromOpacity + (toOpacity - fromOpacity) * e;
      var tf = '';
      if (Math.abs(tx) > 0.01 || Math.abs(ty) > 0.01) tf += 'translate(' + tx + 'px,' + ty + 'px)';
      if (Math.abs(sc - 1) > 0.001) tf += (tf ? ' ' : '') + 'scale(' + sc + ')';
      if (Math.abs(yy) > 0.01) tf += (tf ? ' ' : '') + 'translateY(' + yy + 'px)';
      el.style.transform = tf;
      if (useOpacity) el.style.opacity = String(op);
      if (hasSize) {
        el.style.width = (size.fromW + (size.toW - size.fromW) * e) + 'px';
        el.style.height = (size.fromH + (size.toH - size.fromH) * e) + 'px';
      }
    }
    apply(0);
    function frame() {
      var t = (performance.now() - start) / o.duration;
      apply(t);
      if (t < 1) {
        scheduleFrame(el, frame);
      } else {
        el.__suiMotionRaf = 0;
        el.__suiMotionRafType = '';
        el.style.removeProperty('transform');
        if (useOpacity) el.style.removeProperty('opacity');
        if (hasSize) {
          el.style.removeProperty('width');
          el.style.removeProperty('height');
        }
        el.__suiMotionProps = false;
        if (done) done();
      }
    }
    scheduleFrame(el, frame);
  }
  function animateElement(old, el, o) {
    // Elements with custom-property presets run their whole transition in
    // tweenVars (including the fade/scale/slide effect), so it stays in sync
    // with the geometry tween.
    if (varList(el).length > 0) return;
    var r = el.getBoundingClientRect();
    var dx = old.rect.left - r.left;
    var dy = old.rect.top - r.top;
    var fromOpacity = 1, toOpacity = 1, fromScale = 1, toScale = 1, fromY = 0, toY = 0;
    if (o.effect === 'fade') fromOpacity = 0;
    else if (o.effect === 'scale') { fromOpacity = 0; fromScale = 0.96; }
    else if (o.effect === 'slide') { fromOpacity = 0; fromY = 6; }
    var size = { fromW: old.rect.width, fromH: old.rect.height, toW: r.width, toH: r.height };
    animateMovement(el, dx, dy, o, fromOpacity, toOpacity, fromScale, toScale, fromY, toY, null, size);
    if (old.props) {
      var ns = getComputedStyle(el);
      var from = {}, to = {};
      for (var i = 0; i < STANDARD.length; i++) {
        var p = STANDARD[i];
        if (p === 'opacity' || p === 'transform' || p === 'width' || p === 'height') continue;
        var nv = ns[p];
        if (nv === undefined) continue;
        if (String(old.props[p]) !== String(nv)) { from[p] = old.props[p]; to[p] = nv; }
      }
      if (Object.keys(from).length) {
        el.__suiMotionAnim = el.animate([from, to], { duration: o.duration, easing: o.easing, delay: o.delay, fill: 'none' });
      }
    }
  }
  function enter(el, o) {
    if (o.effect === 'none') return;
    var fromScale = o.effect === 'scale' ? 0.96 : 1;
    var fromY = o.effect === 'slide' ? 10 : 0;
    animateMovement(el, 0, 0, o, 0, 1, fromScale, 1, fromY, 0, null);
  }
  function exitClone(snap) {
    var clone = snap.clone;
    if (!clone) return;
    var o = snap.opts || { effect: 'layout', duration: 220, easing: 'ease-out', delay: 0 };
    var r = snap.rect;
    clone.classList.add('sui-motion-exit');
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
    clone.style.opacity = '1';
    document.body.appendChild(clone);
    var toScale = o.effect === 'scale' ? 0.96 : 1;
    var toY = o.effect === 'slide' ? 10 : 0;
    animateMovement(clone, 0, 0, o, 1, 0, 1, toScale, 0, toY, function() {
      if (clone.parentNode) clone.remove();
    });
    setTimeout(function() {
      if (clone.parentNode) clone.remove();
    }, o.duration + o.delay + 250);
  }
  function before(root) {
    if (pageSwitching) { hadSnapshot = false; return; }
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
    if (pageSwitching) return;
    if (!hadSnapshot) { beforeMap = {}; return; }
    var els = root.querySelectorAll('[data-sui-motion]');
    var seen = {};
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      var k = keyOf(el);
      if (k) seen[k] = 1;
      var o = optionsOf(el);
      if (o.effect === 'none') continue;
      var old = beforeMap[k];
      cancel(el);
      if (old) {
        animateElement(old, el, o);
        tweenVars(el, old, o);
      } else {
        enter(el, o);
      }
    }
    for (var key in beforeMap) {
      if (seen[key]) continue;
      var snap = beforeMap[key];
      if (snap && snap.clone) exitClone(snap);
    }
    beforeMap = {};
  }
  function visiblePageBoxes() {
    var boxes = document.querySelectorAll('[data-sui-pagebox]');
    var out = [];
    for (var i = 0; i < boxes.length; i++) {
      if (getComputedStyle(boxes[i]).display !== 'none') out.push(boxes[i]);
    }
    return out;
  }
  function partRects(el) {
    // Measure the initial/final state of every direct visual part. The shared
    // transition interpolates these two measured states directly instead of
    // letting the browser reflow the content while the box changes size.
    var out = [];
    var kids = el.children || [];
    for (var i = 0; i < kids.length; i++) {
      if (kids[i].nodeType !== 1) continue;
      var r = kids[i].getBoundingClientRect();
      out.push({ left: r.left, top: r.top, width: r.width, height: r.height });
    }
    return out;
  }
  function paintBox(el, rect) {
    // A content-free layer that only draws the surface (background, border,
    // radius, shadow). Its rectangle is animated with explicit pixel values,
    // so growing the card never reflows or rewraps the text.
    var cs = getComputedStyle(el);
    var box = document.createElement('div');
    box.className = 'sui-motion-shared-box';
    box.setAttribute('aria-hidden', 'true');
    box.style.position = 'fixed';
    box.style.left = rect.left + 'px';
    box.style.top = rect.top + 'px';
    box.style.width = rect.width + 'px';
    box.style.height = rect.height + 'px';
    box.style.margin = '0';
    box.style.boxSizing = 'border-box';
    box.style.pointerEvents = 'none';
    box.style.zIndex = '9997';
    // The animated rectangle also clips the final-layout content, so text never
    // spills outside the card while the box is still growing.
    box.style.overflow = 'hidden';
    var props = ['background-color', 'background-image', 'background-size', 'background-position', 'background-repeat',
      'border-top-left-radius', 'border-top-right-radius', 'border-bottom-right-radius', 'border-bottom-left-radius',
      'border-top-width', 'border-right-width', 'border-bottom-width', 'border-left-width',
      'border-top-style', 'border-right-style', 'border-bottom-style', 'border-left-style',
      'border-top-color', 'border-right-color', 'border-bottom-color', 'border-left-color',
      'box-shadow', 'outline-color', 'outline-style', 'outline-width', 'backdrop-filter', '-webkit-backdrop-filter', 'filter'];
    for (var i = 0; i < props.length; i++) {
      var v = cs.getPropertyValue(props[i]);
      if (v) box.style.setProperty(props[i], v);
    }
    document.body.appendChild(box);
    return box;
  }
  function contentLayer(el, rect) {
    var content = el.cloneNode(true);
    stripMotionClone(content);
    content.classList.add('sui-motion-shared-clone');
    content.style.position = 'absolute';
    content.style.width = rect.width + 'px';
    content.style.height = rect.height + 'px';
    content.style.margin = '0';
    content.style.boxSizing = 'border-box';
    content.style.pointerEvents = 'none';
    // The separate box layer owns the surface. Keep border widths (they affect
    // the final layout) but make the paint transparent so the two layers do not
    // draw the same background twice.
    content.style.background = 'transparent';
    content.style.backgroundImage = 'none';
    content.style.boxShadow = 'none';
    content.style.borderTopColor = 'transparent';
    content.style.borderRightColor = 'transparent';
    content.style.borderBottomColor = 'transparent';
    content.style.borderLeftColor = 'transparent';
    return content;
  }
  function animateShared(box, fromRect, toRect, content, oldParts, newParts, o, done) {
    var pairs = [];
    var count = Math.min(oldParts.length, newParts.length, content.children.length);
    for (var i = 0; i < count; i++) {
      var child = content.children[i];
      pairs.push({
        el: child,
        dx: oldParts[i].left - newParts[i].left,
        dy: oldParts[i].top - newParts[i].top
      });
    }
    var cs = getComputedStyle(box);
    var borderLeft = parseFloat(cs.borderLeftWidth) || 0;
    var borderTop = parseFloat(cs.borderTopWidth) || 0;
    var start = performance.now() + o.delay;
    function apply(t) {
      var e = ease(Math.max(0, Math.min(1, t)), o.easing);
      var left = fromRect.left + (toRect.left - fromRect.left) * e;
      var top = fromRect.top + (toRect.top - fromRect.top) * e;
      box.style.left = left + 'px';
      box.style.top = top + 'px';
      box.style.width = (fromRect.width + (toRect.width - fromRect.width) * e) + 'px';
      box.style.height = (fromRect.height + (toRect.height - fromRect.height) * e) + 'px';
      // Keep the content anchored to its final viewport position while the
      // clipping box grows and moves.
      content.style.left = (toRect.left - left - borderLeft) + 'px';
      content.style.top = (toRect.top - top - borderTop) + 'px';
      for (var p = 0; p < pairs.length; p++) {
        pairs[p].el.style.transform = 'translate(' + (pairs[p].dx * (1 - e)) + 'px,' + (pairs[p].dy * (1 - e)) + 'px)';
      }
    }
    apply(0);
    function frame() {
      var t = (performance.now() - start) / o.duration;
      apply(t);
      if (t < 1) {
        scheduleFrame(box, frame);
      } else {
        box.__suiMotionRaf = 0;
        box.__suiMotionRafType = '';
        for (var p = 0; p < pairs.length; p++) pairs[p].el.style.removeProperty('transform');
        done();
      }
    }
    scheduleFrame(box, frame);
  }
  function capturePage() {
    sharedMap = {};
    pageSwitching = true;
    var boxes = visiblePageBoxes();
    for (var b = 0; b < boxes.length; b++) {
      var els = boxes[b].querySelectorAll('[data-sui-motion]');
      for (var i = 0; i < els.length; i++) {
        var k = keyOf(els[i]);
        if (!k) continue;
        var snap = snapshot(els[i], false);
        snap.parts = partRects(els[i]);
        sharedMap[k] = snap;
      }
    }
  }
  function flushShared() {
    var boxes = visiblePageBoxes();
    for (var b = 0; b < boxes.length; b++) {
      var els = boxes[b].querySelectorAll('[data-sui-motion]');
      for (var i = 0; i < els.length; i++) {
        var el = els[i];
        var k = keyOf(el);
        var old = sharedMap[k];
        if (!old) continue;
        var o = optionsOf(el);
        if (o.effect === 'none') continue;
        cancel(el);
        var r = el.getBoundingClientRect();
        if (!old.parts) {
          animateElement(old, el, o);
          delete sharedMap[k];
          continue;
        }
        // Evaluate the two measured states up front: the box rectangle and
        // every part rectangle. Then fuse them with one easing function, so
        // each part travels in a straight line while the box grows.
        var box = paintBox(el, old.rect);
        var content = contentLayer(el, r);
        box.appendChild(content);
        var bcs = getComputedStyle(box);
        content.style.left = (r.left - old.rect.left - (parseFloat(bcs.borderLeftWidth) || 0)) + 'px';
        content.style.top = (r.top - old.rect.top - (parseFloat(bcs.borderTopWidth) || 0)) + 'px';
        var newParts = partRects(content);
        el.style.visibility = 'hidden';
        var finished = false;
        function finishShared() {
          if (finished) return;
          finished = true;
          if (box.parentNode) box.remove();
          if (content.parentNode) content.remove();
          el.style.removeProperty('visibility');
        }
        animateShared(box, old.rect, r, content, old.parts, newParts, o, finishShared);
        setTimeout(finishShared, o.delay + o.duration + 250);
        delete sharedMap[k];
      }
    }
    sharedMap = {};
    pageSwitching = false;
  }
  function captureBeforeToggle(ev) {
    var el = motionTarget(ev.target);
    if (!el) return;
    var k = keyOf(el);
    if (!k) return;
    var snap = snapshot(el);
    preMap[k] = snap;
    // Freeze state-bound toggles until the re-render starts the tween. This
    // removes the one-frame native flash some browsers paint between the click
    // and the reactive update.
    if (el.hasAttribute('data-sui-motion-state')) {
      var names = varList(el);
      for (var i = 0; i < names.length; i++) {
        if (snap.vars[names[i]] != null && snap.vars[names[i]] !== '') el.style.setProperty(names[i], snap.vars[names[i]]);
      }
      setTimeout(function() {
        if (!document.body.contains(el)) return;
        for (var j = 0; j < names.length; j++) el.style.removeProperty(names[j]);
      }, 700);
    }
  }
  if (typeof document !== 'undefined') {
    document.addEventListener('pointerdown', captureBeforeToggle, true);
    document.addEventListener('keydown', function(ev) {
      if (ev.key === ' ' || ev.key === 'Enter') captureBeforeToggle(ev);
    }, true);
  }
  return { before: before, after: after, cancel: cancel, capturePage: capturePage, flushShared: flushShared };
})();`
