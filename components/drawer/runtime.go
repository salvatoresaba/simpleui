package drawer

// drawer_runtime.go - close behavior for the state-controlled drawer.
//
// Visibility is rendered from the `open` binding; this runtime only dispatches
// the close action (state setter or onClose function) from the backdrop, the
// close button or Escape.
const Runtime = `(function(){
  function closeLayer(layer) {
    if (!layer) return;
    var setter = layer.getAttribute('data-sui-drawer-set');
    if (setter && typeof window[setter] === 'function') { window[setter](false); return; }
    var fn = layer.getAttribute('data-sui-drawer-onclose');
    if (fn && typeof window[fn] === 'function') window[fn]();
  }
  document.addEventListener('click', function(ev) {
    var target = ev.target && ev.target.closest ? ev.target.closest('[data-sui-drawer-close],[data-sui-drawer-backdrop]') : null;
    if (!target) return;
    var layer = target.closest('[data-sui-drawer]');
    if (layer) closeLayer(layer);
  });
  document.addEventListener('keydown', function(ev) {
    if (ev.key !== 'Escape') return;
    var open = document.querySelector('[data-sui-drawer][data-sui-drawer-open="1"]');
    if (open) closeLayer(open);
  });
})();`
