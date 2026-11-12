package main

// stepper_runtime.go - delegated clicks for state-bound steppers.
//
// The generated markup carries data-sui-step-index on each step and
// data-sui-stepper-set on the container. Clicking a step (or pressing Enter or
// Space on it) calls the state setter with the step index.
const suiStepperRuntime = `(function(){
  function closest(el, sel) { return el && el.closest ? el.closest(sel) : null; }
  function activate(el) {
    var step = closest(el, '[data-sui-step]');
    if (!step) return;
    var box = step.closest('[data-sui-stepper]');
    if (!box) return;
    var setter = box.getAttribute('data-sui-stepper-set');
    var idx = parseInt(step.getAttribute('data-sui-step-index'), 10);
    if (!isNaN(idx) && setter && typeof window[setter] === 'function') window[setter](idx);
  }
  document.addEventListener('click', function(ev) { activate(ev.target); });
  document.addEventListener('keydown', function(ev) {
    if (ev.key !== 'Enter' && ev.key !== ' ') return;
    var step = closest(ev.target, '[data-sui-step]');
    if (!step) return;
    ev.preventDefault();
    activate(step);
  });
})();`
