package main

// popover_runtime.go - browser runtime for the popover control.
//
// The generated markup is a relative container with a trigger button and a
// panel. The panel content is normal generated markup; this runtime only
// toggles the open class/hidden attribute, keeps ARIA in sync and closes on
// outside clicks or Escape.
const suiPopoverRuntime = `(function(){
  function closest(el, sel) { return el && el.closest ? el.closest(sel) : null; }
  function closeAll(except) {
    var open = document.querySelectorAll('[data-sui-popover].open');
    for (var i = 0; i < open.length; i++) {
      if (open[i] === except) continue;
      open[i].classList.remove('open');
      var panel = open[i].querySelector('.sui-popover-panel');
      if (panel) { panel.hidden = true; __SUI_OVERLAY.reset(panel); }
      var trigger = open[i].querySelector('[data-sui-popover-trigger]');
      if (trigger) trigger.setAttribute('aria-expanded', 'false');
    }
  }
  function openPanel(pop, trigger, panel) {
    pop.classList.add('open');
    panel.hidden = false;
    trigger.setAttribute('aria-expanded', 'true');
    var pos = pop.getAttribute('data-sui-popover-position') || 'auto';
    var placed = __SUI_OVERLAY.place(trigger, panel, { placement: pos, align: 'start', gap: 6, margin: 8 });
    __SUI_OVERLAY.animate(panel, pop.getAttribute('data-sui-animation') || 'scale', pop.getAttribute('data-sui-duration'), pop.getAttribute('data-sui-easing'), pop.getAttribute('data-sui-delay'), placed);
  }
  document.addEventListener('click', function(ev) {
    var trigger = closest(ev.target, '[data-sui-popover-trigger]');
    if (trigger) {
      var pop = trigger.closest('[data-sui-popover]');
      var wasOpen = pop && pop.classList.contains('open');
      closeAll(null);
      if (pop && !wasOpen) {
        var panel = pop.querySelector('.sui-popover-panel');
        if (panel) openPanel(pop, trigger, panel);
      }
      ev.preventDefault();
      return;
    }
    if (!closest(ev.target, '[data-sui-popover]')) closeAll(null);
  });
  document.addEventListener('keydown', function(ev) {
    if (ev.key === 'Escape') closeAll(null);
  });
})();`
