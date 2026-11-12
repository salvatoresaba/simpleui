package main

// commandbar_runtime.go - search, keyboard navigation and overlay control for
// the command palette.
//
// Inline command bars are always visible. Overlay bars render a backdrop and
// are shown when data-sui-commandbar-open="1"; the optional hotkey (Ctrl/Cmd +
// key) toggles the bound state. Typing filters commands, Arrow Up/Down moves
// the active row, Enter invokes it and Escape closes the overlay.
const suiCommandbarRuntime = `(function(){
  function closest(el, sel) { return el && el.closest ? el.closest(sel) : null; }
  function isOverlay(el) { return el && el.hasAttribute('data-sui-commandbar-overlay'); }
  function setOpen(el, value) {
    if (!el) return;
    var setter = el.getAttribute('data-sui-commandbar-set');
    if (setter && typeof window[setter] === 'function') { window[setter](value); return; }
    el.setAttribute('data-sui-commandbar-open', value ? '1' : '0');
  }
  function inputOf(el) { return el ? el.querySelector('.sui-commandbar-input') : null; }
  function focusInput(el) {
    var inp = inputOf(el);
    if (!inp) return;
    setTimeout(function() { try { inp.focus(); inp.select(); } catch (e) {} }, 0);
  }
  function focusOpen() {
    var open = document.querySelectorAll('[data-sui-commandbar-overlay][data-sui-commandbar-open="1"]');
    for (var i = 0; i < open.length; i++) focusInput(open[i]);
  }
  function filter(el) {
    var inp = inputOf(el);
    var q = inp && inp.value ? inp.value.toLowerCase().trim() : '';
    var list = el.querySelectorAll('.sui-command');
    var visible = 0;
    var first = null;
    for (var i = 0; i < list.length; i++) {
      var hay = list[i].getAttribute('data-sui-command-search') || '';
      var show = !q || hay.indexOf(q) >= 0;
      list[i].hidden = !show;
      list[i].classList.remove('active');
      if (show) { visible++; if (!first) first = list[i]; }
    }
    if (first) first.classList.add('active');
    var empty = el.querySelector('.sui-commandbar-empty');
    if (empty) empty.hidden = visible > 0;
  }
  function visibleCommands(el) {
    var list = el.querySelectorAll('.sui-command');
    var out = [];
    for (var i = 0; i < list.length; i++) if (!list[i].hidden) out.push(list[i]);
    return out;
  }
  function move(el, delta) {
    var list = visibleCommands(el);
    if (!list.length) return;
    var idx = 0;
    for (var i = 0; i < list.length; i++) if (list[i].classList.contains('active')) idx = i;
    idx = (idx + delta + list.length) % list.length;
    for (var j = 0; j < list.length; j++) list[j].classList.toggle('active', j === idx);
    if (list[idx].scrollIntoView) list[idx].scrollIntoView({ block: 'nearest' });
  }
  document.addEventListener('input', function(ev) {
    var inp = closest(ev.target, '.sui-commandbar-input');
    if (!inp) return;
    var el = closest(inp, '[data-sui-commandbar]');
    if (el) filter(el);
  });
  document.addEventListener('keydown', function(ev) {
    var el = closest(ev.target, '[data-sui-commandbar]');
    if (el) {
      if (ev.key === 'ArrowDown') { ev.preventDefault(); move(el, 1); return; }
      if (ev.key === 'ArrowUp') { ev.preventDefault(); move(el, -1); return; }
      if (ev.key === 'Enter') {
        var list = visibleCommands(el);
        var active = null;
        for (var i = 0; i < list.length; i++) if (list[i].classList.contains('active')) active = list[i];
        if (!active && list.length) active = list[0];
        if (active) { ev.preventDefault(); active.click(); }
        return;
      }
      if (ev.key === 'Escape' && isOverlay(el)) { ev.preventDefault(); setOpen(el, false); return; }
    }
    if (ev.key === 'Escape') {
      var openBar = document.querySelector('[data-sui-commandbar-overlay][data-sui-commandbar-open="1"]');
      if (openBar) { ev.preventDefault(); setOpen(openBar, false); return; }
    }
    if (ev.ctrlKey || ev.metaKey) {
      var key = (ev.key || '').toLowerCase();
      var bars = document.querySelectorAll('[data-sui-commandbar-overlay]');
      for (var b = 0; b < bars.length; b++) {
        if ((bars[b].getAttribute('data-sui-commandbar-hotkey') || '') !== key) continue;
        ev.preventDefault();
        var wasOpen = bars[b].getAttribute('data-sui-commandbar-open') === '1';
        setOpen(bars[b], !wasOpen);
        if (!wasOpen) setTimeout(focusOpen, 0);
        return;
      }
    }
  });
  document.addEventListener('click', function(ev) {
    var close = closest(ev.target, '[data-sui-commandbar-close]');
    if (close) {
      var bar = closest(close, '[data-sui-commandbar]');
      if (bar) { ev.preventDefault(); setOpen(bar, false); }
      return;
    }
    var cmd = closest(ev.target, '.sui-command');
    if (cmd) {
      var box = closest(cmd, '[data-sui-commandbar]');
      if (box && isOverlay(box)) setTimeout(function() { setOpen(box, false); }, 0);
    }
  });
  if (typeof document !== 'undefined') {
    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', focusOpen);
    else focusOpen();
    if (typeof MutationObserver === 'function') {
      new MutationObserver(function(list) {
        for (var i = 0; i < list.length; i++) {
          var el = list[i].target;
          if (list[i].attributeName === 'data-sui-commandbar-open' && el.getAttribute('data-sui-commandbar-open') === '1') focusInput(el);
        }
      }).observe(document.body, { attributes: true, subtree: true, attributeFilter: ['data-sui-commandbar-open'] });
    }
  }
})();`
