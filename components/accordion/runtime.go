package accordion

// accordion_runtime.go - browser runtime for the accordion control.
//
// Markup: <div class="sui-accordion" data-sui-accordion-key="...">
//
//	<div class="sui-accordion-item open">
//	  <button class="sui-accordion-head" aria-expanded="true">Title</button>
//	  <div class="sui-accordion-panel">content</div>
//	</div>
//
// </div>
// The runtime toggles one section at a time and remembers the open item across
// view re-renders when the accordion has a key.
const Runtime = `(function(){
  var store = Object.create(null);
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function items(box) {
    if (!box) return [];
    return Array.prototype.filter.call(box.children, function(el) {
      return el.classList && el.classList.contains('sui-accordion-item');
    });
  }
  function setOpen(box, index, persist) {
    var list = items(box);
    for (var i = 0; i < list.length; i++) {
      var open = i === index;
      list[i].classList.toggle('open', open);
      var head = list[i].querySelector('.sui-accordion-head');
      var panel = list[i].querySelector('.sui-accordion-panel');
      if (head) head.setAttribute('aria-expanded', open ? 'true' : 'false');
      if (panel) panel.hidden = !open;
    }
    var key = box.getAttribute('data-sui-accordion-key');
    if (key && persist) store[key] = index;
  }
  function applyStored() {
    var boxes = document.querySelectorAll('.sui-accordion');
    for (var i = 0; i < boxes.length; i++) {
      var key = boxes[i].getAttribute('data-sui-accordion-key');
      if (key && store[key] !== undefined) setOpen(boxes[i], store[key], false);
    }
  }
  document.addEventListener('click', function(ev) {
    var head = closest(ev.target, '.sui-accordion-head');
    if (!head) return;
    var item = closest(head, '.sui-accordion-item');
    var box = closest(head, '.sui-accordion');
    if (!item || !box) return;
    var list = items(box);
    var index = list.indexOf(item);
    if (index < 0) return;
    setOpen(box, item.classList.contains('open') ? -1 : index, true);
  });
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      setTimeout(function() { pending = false; applyStored(); }, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
  applyStored();
})();`
