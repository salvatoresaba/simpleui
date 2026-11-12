package main

// tooltip_runtime.go - browser runtime for the tooltip element property.
//
// Any element can carry data-sui-tooltip="Text". The runtime creates one
// portal bubble appended to <body>, shows it on pointer hover or keyboard
// focus, positions it in the viewport and removes it on leave/blur/Escape.
// Text is assigned with textContent and is never interpreted as HTML.
const suiTooltipRuntime = `(function(){
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
})();`
