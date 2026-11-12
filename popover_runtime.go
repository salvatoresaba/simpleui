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
      if (panel) panel.hidden = true;
      var trigger = open[i].querySelector('[data-sui-popover-trigger]');
      if (trigger) trigger.setAttribute('aria-expanded', 'false');
    }
  }
  document.addEventListener('click', function(ev) {
    var trigger = closest(ev.target, '[data-sui-popover-trigger]');
    if (trigger) {
      var pop = trigger.closest('[data-sui-popover]');
      var wasOpen = pop && pop.classList.contains('open');
      closeAll(null);
      if (pop && !wasOpen) {
        pop.classList.add('open');
        var panel = pop.querySelector('.sui-popover-panel');
        if (panel) panel.hidden = false;
        trigger.setAttribute('aria-expanded', 'true');
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
