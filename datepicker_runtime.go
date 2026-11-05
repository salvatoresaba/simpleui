package main

// datepicker_runtime.go - browser runtime for the native date/time picker
// wrapper. The native <input type=date|time|datetime-local> supplies the
// platform picker and keyboard behavior; this runtime synchronizes the value
// with the hidden state-binding input and reports invalid/out-of-range state.
//
// Markup:
//
//	<div class="sui-datepicker" data-sui-datepicker="1" data-sui-mode="date">
//	  <label class="sui-datepicker-label">Appointment</label>
//	  <div class="sui-datepicker-control">
//	    <input class="sui-datepicker-input" type="date" ...>
//	    <button class="sui-datepicker-clear">×</button>
//	  </div>
//	  <input type="hidden" class="sui-datepicker-value">
//	  <div class="sui-datepicker-error" role="alert" hidden></div>
//	</div>
const suiDatePickerRuntime = `(function(){
  var seq = 0;
  function validFormat(value, mode) {
    if (!value) return true;
    if (mode === 'time') return /^\d{2}:\d{2}(:\d{2})?$/.test(value);
    if (mode === 'datetime') return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?$/.test(value);
    return /^\d{4}-\d{2}-\d{2}$/.test(value);
  }
  function init(root) {
    if (!root || root.__suiDatepicker) return;
    root.__suiDatepicker = true;
    var input = root.querySelector('.sui-datepicker-input');
    var hidden = root.querySelector('.sui-datepicker-value');
    var label = root.querySelector('.sui-datepicker-label');
    var clear = root.querySelector('.sui-datepicker-clear');
    var errorEl = root.querySelector('.sui-datepicker-error');
    if (!input || !hidden) return;
    var mode = root.getAttribute('data-sui-mode') || 'date';
    var setter = root.getAttribute('data-sui-datepicker-setter') || '';
    var handler = root.getAttribute('data-sui-datepicker-onchange') || '';
    var disabled = root.getAttribute('data-sui-disabled') === '1';
    var min = input.getAttribute('min') || '';
    var max = input.getAttribute('max') || '';
    var current = hidden.value || '';
    if (!input.id) input.id = 'sui-datepicker-input-' + (++seq);
    if (label) label.setAttribute('for', input.id);
    if (disabled) {
      input.disabled = true;
      if (clear) clear.disabled = true;
    }
    function setError(message) {
      var msg = message || '';
      if (errorEl) {
        errorEl.textContent = msg;
        errorEl.hidden = !msg;
      }
      if (msg) input.setAttribute('aria-invalid', 'true');
      else input.removeAttribute('aria-invalid');
    }
    function validate(value) {
      if (!value) return '';
      if (!validFormat(value, mode)) return 'Invalid ' + mode + ' value';
      if (min && value < min) return 'Value is before the minimum';
      if (max && value > max) return 'Value is after the maximum';
      return '';
    }
    function sync() {
      input.value = current;
      setError(validate(current));
      if (clear) clear.hidden = !current;
    }
    function commit(value) {
      current = value;
      hidden.value = value;
      input.value = value;
      if (clear) clear.hidden = !value;
      setError(validate(value));
      if (setter && typeof window[setter] === 'function') window[setter](value);
      if (handler && typeof window[handler] === 'function') window[handler]();
    }
    input.addEventListener('change', function() { commit(input.value); });
    if (clear) clear.addEventListener('click', function(ev) {
      ev.preventDefault();
      if (disabled) return;
      commit('');
      input.focus();
    });
    sync();
  }
  function initAll() {
    var all = document.querySelectorAll('.sui-datepicker');
    for (var i = 0; i < all.length; i++) init(all[i]);
  }
  initAll();
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      setTimeout(function() { pending = false; initAll(); }, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
})();`
