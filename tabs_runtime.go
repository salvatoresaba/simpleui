package main

// tabs_runtime.go - browser runtime for the tabs control.
//
// Markup: <div class="sui-tabs" data-sui-tabs-key="..."><div class="sui-tablist">
// buttons... </div><div class="sui-tabpanels"> panels... </div></div>.
// The runtime activates tabs, supports arrow/Home/End keys, and remembers the
// selected tab by key across view re-renders.
const suiTabsRuntime = `(function(){
  var store = Object.create(null);
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function tabButtons(box) {
    var list = box ? box.querySelector('.sui-tablist') : null;
    if (!list) return [];
    return Array.prototype.filter.call(list.children, function(el) {
      return el.classList && el.classList.contains('sui-tab');
    });
  }
  function tabPanels(box) {
    var list = box ? box.querySelector('.sui-tabpanels') : null;
    if (!list) return [];
    return Array.prototype.filter.call(list.children, function(el) {
      return el.classList && el.classList.contains('sui-tabpanel');
    });
  }
  function activate(box, index, persist) {
    var tabs = tabButtons(box);
    var panels = tabPanels(box);
    if (!tabs.length) return;
    if (index < 0) index = tabs.length - 1;
    if (index >= tabs.length) index = 0;
    for (var i = 0; i < tabs.length; i++) {
      var on = i === index;
      tabs[i].classList.toggle('active', on);
      tabs[i].setAttribute('aria-selected', on ? 'true' : 'false');
      tabs[i].tabIndex = on ? 0 : -1;
      if (panels[i]) {
        panels[i].hidden = !on;
        panels[i].classList.toggle('active', on);
      }
    }
    var key = box.getAttribute('data-sui-tabs-key');
    if (key && persist) store[key] = index;
  }
  function applyStored() {
    var boxes = document.querySelectorAll('.sui-tabs');
    for (var i = 0; i < boxes.length; i++) {
      var key = boxes[i].getAttribute('data-sui-tabs-key');
      if (key && store[key] !== undefined) activate(boxes[i], store[key], false);
    }
  }
  document.addEventListener('click', function(ev) {
    var tab = closest(ev.target, '.sui-tab');
    if (!tab) return;
    var box = closest(tab, '.sui-tabs');
    if (!box) return;
    var tabs = tabButtons(box);
    activate(box, tabs.indexOf(tab), true);
  });
  document.addEventListener('keydown', function(ev) {
    var tab = closest(ev.target, '.sui-tab');
    if (!tab) return;
    var box = closest(tab, '.sui-tabs');
    if (!box) return;
    var tabs = tabButtons(box);
    var index = tabs.indexOf(tab);
    var next = -1;
    if (ev.key === 'ArrowRight') next = index + 1;
    else if (ev.key === 'ArrowLeft') next = index - 1;
    else if (ev.key === 'Home') next = 0;
    else if (ev.key === 'End') next = tabs.length - 1;
    if (next < 0 && ev.key === 'ArrowLeft') next = tabs.length - 1;
    if (next >= tabs.length) next = 0;
    if (next < 0) return;
    ev.preventDefault();
    activate(box, next, true);
    tabs[next].focus();
  });
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      var run = function() { pending = false; applyStored(); };
      // setTimeout (not rAF): background/embedded previews may throttle frames.
      setTimeout(run, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
  applyStored();
})();`
