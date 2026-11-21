package carousel

// carousel_runtime.go - slide switching for the carousel control.
//
// A carousel renders all slides in one track and moves it with a transform.
// When the carousel has a data-sui-carousel-set state setter the runtime calls
// it with the next index; otherwise it updates the DOM directly. Arrow buttons,
// dots, Left/Right keys and optional autoplay all share the same apply path.
const Runtime = `(function(){
  function closest(el, sel) { return el && el.closest ? el.closest(sel) : null; }
  function indexOf(el) {
    var n = parseInt(el.getAttribute('data-sui-carousel-index'), 10);
    return isNaN(n) ? 0 : n;
  }
  function slideCount(el) {
    var box = el.querySelector('.sui-carousel-track');
    return box ? box.querySelectorAll('.sui-carousel-slide').length : 0;
  }
  function apply(el, n) {
    var count = slideCount(el);
    if (!count) return;
    var loop = el.getAttribute('data-sui-carousel-loop') === '1';
    if (loop) { n = ((n % count) + count) % count; }
    else { n = Math.max(0, Math.min(count - 1, n)); }
    var setter = el.getAttribute('data-sui-carousel-set');
    if (setter && typeof window[setter] === 'function') { window[setter](n); return; }
    el.setAttribute('data-sui-carousel-index', String(n));
    var track = el.querySelector('.sui-carousel-track');
    if (track) track.style.transform = 'translateX(' + (-n * 100) + '%)';
    var dots = el.querySelectorAll('[data-sui-carousel-dot]');
    for (var i = 0; i < dots.length; i++) dots[i].classList.toggle('active', i === n);
  }
  document.addEventListener('click', function(ev) {
    var el = closest(ev.target, '[data-sui-carousel]');
    if (!el) return;
    if (closest(ev.target, '[data-sui-carousel-prev]')) { ev.preventDefault(); apply(el, indexOf(el) - 1); return; }
    if (closest(ev.target, '[data-sui-carousel-next]')) { ev.preventDefault(); apply(el, indexOf(el) + 1); return; }
    var dot = closest(ev.target, '[data-sui-carousel-dot]');
    if (dot) {
      ev.preventDefault();
      var n = parseInt(dot.getAttribute('data-sui-carousel-dot'), 10);
      apply(el, isNaN(n) ? 0 : n);
    }
  });
  document.addEventListener('keydown', function(ev) {
    var el = closest(ev.target, '[data-sui-carousel]');
    if (!el) return;
    if (ev.key === 'ArrowLeft') { ev.preventDefault(); apply(el, indexOf(el) - 1); }
    else if (ev.key === 'ArrowRight') { ev.preventDefault(); apply(el, indexOf(el) + 1); }
  });
  function startAutoplay() {
    var list = document.querySelectorAll('[data-sui-carousel-autoplay="1"]');
    for (var i = 0; i < list.length; i++) {
      var el = list[i];
      if (el.__suiTimer) continue;
      var interval = parseInt(el.getAttribute('data-sui-carousel-interval'), 10);
      if (isNaN(interval) || interval < 500) interval = 4000;
      el.__suiTimer = setInterval(function() {
        if (!document.body.contains(el)) { clearInterval(el.__suiTimer); return; }
        apply(el, indexOf(el) + 1);
      }, interval);
    }
  }
  if (typeof document !== 'undefined') {
    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', startAutoplay);
    else startAutoplay();
    if (typeof MutationObserver === 'function') {
      new MutationObserver(startAutoplay).observe(document.body, { childList: true, subtree: true });
    }
  }
})();`
