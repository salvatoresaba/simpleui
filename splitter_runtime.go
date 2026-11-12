package main

// splitter_runtime.go - resizable two-pane splitter.
//
// The generated markup is a flex container with two .sui-split-pane children
// and a .sui-split-bar between them. The first pane carries an initial
// flex-basis; the bar is the drag handle. Pointer drags and Left/Right or
// Up/Down keys (with Home/End) resize the panes while respecting the configured
// minimum size.
const suiSplitterRuntime = `(function(){
  function axisOf(split) {
    return split.getAttribute('data-sui-splitter-direction') === 'vertical' ? 'y' : 'x';
  }
  function minSize(el) {
    var m = el.getAttribute('data-sui-splitter-min');
    if (!m) return 0;
    var num = parseFloat(m);
    if (isNaN(num)) return 0;
    return /%$/.test(m) ? 0 : num;
  }
  function panes(split) {
    return split.querySelectorAll(':scope > .sui-split-pane');
  }
  function applySplit(split, first, pct) {
    var total = panes(split);
    if (total.length < 2) return;
    var r = split.getBoundingClientRect();
    var limit = minSize(split);
    var available = axisOf(split) === 'y' ? r.height : r.width;
    if (available <= 0) return;
    var minPct = limit > 0 ? (limit / available) * 100 : 0;
    var maxPct = 100 - minPct;
    pct = Math.max(minPct, Math.min(maxPct, pct));
    first.style.flexBasis = pct + '%';
  }
  function beginDrag(split, bar, ev) {
    var first = panes(split)[0];
    if (!first) return;
    var vertical = axisOf(split) === 'y';
    var r = split.getBoundingClientRect();
    var startPos = vertical ? ev.clientY : ev.clientX;
    var startPct = (vertical ? first.offsetHeight : first.offsetWidth) / (vertical ? r.height : r.width) * 100;
    var size = vertical ? r.height : r.width;
    split.classList.add('sui-splitbar-active');
    function move(e) {
      var pos = vertical ? e.clientY : e.clientX;
      if (size <= 0) return;
      applySplit(split, first, startPct + ((pos - startPos) / size) * 100);
    }
    function end() {
      split.classList.remove('sui-splitbar-active');
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', end);
      window.removeEventListener('pointercancel', end);
    }
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', end);
    window.addEventListener('pointercancel', end);
  }
  function keyStep(split, bar, ev) {
    var first = panes(split)[0];
    if (!first) return;
    var vertical = axisOf(split) === 'y';
    var r = split.getBoundingClientRect();
    var size = vertical ? r.height : r.width;
    if (size <= 0) return;
    var current = (vertical ? first.offsetHeight : first.offsetWidth) / size * 100;
    var step = ev.shiftKey ? 10 : 2;
    var next = current;
    if (ev.key === 'Home') next = 0;
    else if (ev.key === 'End') next = 100;
    else if (vertical && ev.key === 'ArrowUp') next = current - step;
    else if (vertical && ev.key === 'ArrowDown') next = current + step;
    else if (!vertical && ev.key === 'ArrowLeft') next = current - step;
    else if (!vertical && ev.key === 'ArrowRight') next = current + step;
    else return;
    ev.preventDefault();
    applySplit(split, first, next);
  }
  document.addEventListener('pointerdown', function(ev) {
    var bar = ev.target && ev.target.closest ? ev.target.closest('.sui-split-bar') : null;
    if (!bar) return;
    var split = bar.closest('[data-sui-splitter]');
    if (!split) return;
    ev.preventDefault();
    beginDrag(split, bar, ev);
  });
  document.addEventListener('keydown', function(ev) {
    var bar = ev.target && ev.target.closest ? ev.target.closest('.sui-split-bar') : null;
    if (!bar) return;
    var split = bar.closest('[data-sui-splitter]');
    if (!split) return;
    keyStep(split, bar, ev);
  });
})();`
