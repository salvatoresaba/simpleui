package main

// overlay_runtime.go - shared viewport placement and animation helper.
//
// Every anchored overlay (menu/dropdown, popover, tooltip, combobox popup)
// uses this helper so it stays inside the visible window: preferred placement
// is flipped when there is no space, coordinates are clamped to the viewport,
// and tall panels scroll instead of overflowing. It also applies configurable
// open animations (effect, duration, easing, delay) through data attributes
// and CSS custom properties.
const suiOverlayRuntime = `const __SUI_OVERLAY = (function(){
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
})();`
