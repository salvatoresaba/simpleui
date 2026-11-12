package main

// combobox_runtime.go - browser runtime for the searchable combobox control.
//
// Markup:
//
//	<div class="sui-combobox" data-sui-options="Italy,France=fr">
//	  <label class="sui-combobox-label">Country</label>
//	  <div class="sui-combobox-control">
//	    <input class="sui-combobox-input" ...>
//	    <button class="sui-combobox-clear">×</button>
//	    <button class="sui-combobox-toggle">▾</button>
//	  </div>
//	  <input type="hidden" class="sui-combobox-value">
//	  <div class="sui-combobox-popup" role="listbox" hidden></div>
//	  <div class="sui-combobox-error" role="alert" hidden></div>
//	</div>
//
// Options are text and are written with textContent. Options may use
// "Label=value" to separate display text and stored value. A backslash escapes
// the next character in option data, so a comma inside a label is written
// "Washington\, D.C." at runtime. In SUI source the backslash itself is
// escaped, e.g. "Washington\\, D.C.". Invalid option data is reported in the
// combobox error region instead of being interpreted as partial data.
const suiComboboxRuntime = `(function(){
  var seq = 0;
  var MAX_OPTIONS = 2000;
  var MAX_QUERY = 256;
  var MAX_OPTION_LIMIT = 10000;
  var MAX_OPTION_TEXT = 10000;
  var MAX_OPTIONS_TEXT = 500000;
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  // parseOptions returns {list, error}. It never throws and never returns
  // partially-parsed data: any malformed escape, empty label/value or control
  // character makes the whole list invalid so stale options are not shown.
  function parseOptions(raw) {
    var text = String(raw == null ? '' : raw);
    if (text.trim() === '') return { list: [], error: '' };
    if (text.length > MAX_OPTIONS_TEXT) return { list: [], error: 'Invalid options data: list too long' };
    var out = [];
    var label = '';
    var value = '';
    var hasValue = false;
    var escaped = false;
    var error = '';
    function addOption() {
      var l = label.trim();
      var v = hasValue ? value.trim() : l;
      if (!l) { error = 'Invalid options data: empty option label'; return false; }
      if (hasValue && !v) { error = 'Invalid options data: empty option value'; return false; }
      if (/[\u0000-\u001f]/.test(l) || /[\u0000-\u001f]/.test(v)) {
        error = 'Invalid options data: control character'; return false;
      }
      if (l.length > MAX_OPTION_TEXT || v.length > MAX_OPTION_TEXT) {
        error = 'Invalid options data: option too long'; return false;
      }
      out.push({ label: l, value: v });
      if (out.length > MAX_OPTIONS) {
        error = 'Invalid options data: too many options'; return false;
      }
      return true;
    }
    for (var i = 0; i < text.length; i++) {
      var ch = text.charAt(i);
      if (escaped) {
        if (hasValue) value += ch; else label += ch;
        escaped = false;
        continue;
      }
      if (ch === '\\') { escaped = true; continue; }
      if (ch === ',') {
        if (!addOption()) return { list: [], error: error };
        label = ''; value = ''; hasValue = false;
        continue;
      }
      if (ch === '=' && !hasValue) { hasValue = true; continue; }
      if (hasValue) value += ch; else label += ch;
    }
    if (escaped) return { list: [], error: 'Invalid options data: dangling escape' };
    if (!addOption()) return { list: [], error: error };
    return { list: out, error: '' };
  }
  // hasGlob checks for an unescaped * or ? in a search pattern.
  function hasGlob(pattern) {
    var escaped = false;
    for (var i = 0; i < pattern.length; i++) {
      var ch = pattern.charAt(i);
      if (escaped) { escaped = false; continue; }
      if (ch === '\\') { escaped = true; continue; }
      if (ch === '*' || ch === '?') return true;
    }
    return false;
  }
  function literalPattern(pattern) {
    return pattern.replace(/\\([\s\S])/g, '$1');
  }
  // globMatch is an iterative glob matcher. It supports * (any sequence), ?
  // (one character) and backslash escapes. No user input is compiled as a
  // regular expression, avoiding regex injection and catastrophic backtracking.
  function globMatch(text, pattern) {
    var ti = 0, pi = 0, star = -1, mark = 0;
    while (ti < text.length) {
      var current = pi < pattern.length ? pattern.charAt(pi) : '';
      var literal = null;
      if (current === '\\') {
        if (pi + 1 < pattern.length) {
          literal = pattern.charAt(pi + 1);
          pi += 2;
        } else {
          literal = '\\';
          pi += 1;
        }
      } else if (current === '?') {
        ti++; pi++; continue;
      } else if (current === '*') {
        star = pi; mark = ti; pi++; continue;
      } else if (current !== '') {
        literal = current;
        pi++;
      }
      if (literal !== null && ti < text.length && text.charAt(ti) === literal) {
        ti++;
        continue;
      }
      if (star >= 0) { pi = star + 1; ti = ++mark; }
      else return false;
    }
    while (pi < pattern.length && pattern.charAt(pi) === '*') pi++;
    return pi === pattern.length;
  }
  function matches(label, query) {
    var q = String(query || '').trim();
    if (!q) return true;
    var text = String(label || '').toLowerCase();
    q = q.toLowerCase().slice(0, MAX_QUERY);
    if (hasGlob(q)) return globMatch(text, q);
    return text.indexOf(literalPattern(q)) >= 0;
  }
  function init(root) {
    if (!root || root.__suiCombobox) return;
    root.__suiCombobox = true;
    var input = root.querySelector('.sui-combobox-input');
    var hidden = root.querySelector('.sui-combobox-value');
    var popup = root.querySelector('.sui-combobox-popup');
    var label = root.querySelector('.sui-combobox-label');
    var toggle = root.querySelector('.sui-combobox-toggle');
    var clear = root.querySelector('.sui-combobox-clear');
    var errorEl = root.querySelector('.sui-combobox-error');
    if (!input || !hidden || !popup) return;
    var parsed = parseOptions(root.getAttribute('data-sui-options'));
    var options = [];
    var configError = parsed.error;
    var emptyText = root.getAttribute('data-sui-empty') || 'No results';
    var setter = root.getAttribute('data-sui-combobox-setter') || '';
    var handler = root.getAttribute('data-sui-combobox-onchange') || '';
    var disabled = root.getAttribute('data-sui-disabled') === '1';
    var required = root.getAttribute('data-sui-required') === '1';
    var missingRequired = false;
    var maxRaw = String(root.getAttribute('data-sui-max-option-length') || '').trim();
    var maxLength = 0;
    if (maxRaw !== '') {
      if (!/^[0-9]+$/.test(maxRaw) || parseInt(maxRaw, 10) > MAX_OPTION_LIMIT) {
        configError = 'Invalid max option length';
      } else {
        maxLength = parseInt(maxRaw, 10);
      }
    }
    var maxCountRaw = String(root.getAttribute('data-sui-max-options') || '').trim();
    var maxCount = 0;
    if (maxCountRaw !== '') {
      if (!/^[0-9]+$/.test(maxCountRaw) || parseInt(maxCountRaw, 10) > MAX_OPTIONS) {
        configError = 'Invalid max options';
      } else {
        maxCount = parseInt(maxCountRaw, 10);
      }
    }
    // Validate the complete candidate list before publishing it. A single
    // overlong item (label or stored value) or an excessive option count
    // rejects the whole list, so no partial data reaches the DOM.
    if (!configError) {
      for (var vi = 0; vi < parsed.list.length; vi++) {
        var candidate = parsed.list[vi];
        if (maxLength > 0 && (candidate.label.length > maxLength || candidate.value.length > maxLength)) {
          configError = 'Invalid options data: option ' + (vi + 1) + ' exceeds max option length';
          break;
        }
      }
      if (!configError && maxCount > 0 && parsed.list.length > maxCount) {
        configError = 'Invalid options data: too many options (max ' + maxCount + ')';
      }
      if (!configError) options = parsed.list;
    }
    var filtered = options.slice();
    var active = -1;
    var current = hidden.value || '';
    if (!input.id) input.id = 'sui-combobox-input-' + (++seq);
    if (!popup.id) popup.id = 'sui-combobox-list-' + (++seq);
    if (!label.id) label.id = 'sui-combobox-label-' + (++seq);
    label.setAttribute('for', input.id);
    input.setAttribute('aria-controls', popup.id);
    input.setAttribute('aria-autocomplete', 'list');
    input.setAttribute('aria-haspopup', 'listbox');
    if (required) input.setAttribute('aria-required', 'true');
    if (disabled) {
      input.disabled = true;
      toggle.disabled = true;
      if (clear) clear.disabled = true;
    }
    function labelOf(value) {
      for (var i = 0; i < options.length; i++) if (options[i].value === value) return options[i].label;
      return '';
    }
    function renderError() {
      var message = configError;
      if (!message && required && !disabled && missingRequired && !current) {
        message = 'This field is required';
      }
      if (errorEl) {
        errorEl.textContent = message;
        errorEl.hidden = !message;
      }
      if (message) input.setAttribute('aria-invalid', 'true');
      else input.removeAttribute('aria-invalid');
      if (toggle) toggle.disabled = disabled || !!configError;
      if (configError) {
        popup.hidden = true;
        input.setAttribute('aria-expanded', 'false');
      }
    }
    function setError(message) {
      configError = message || '';
      renderError();
    }
    function syncInput() {
      input.value = labelOf(current);
      if (clear) clear.hidden = !current;
    }
    function clearMissingSelection() {
      if (configError || !current) return;
      for (var i = 0; i < options.length; i++) {
        if (options[i].value === current) return;
      }
      // The selected value is no longer present in the replacement list.
      // Clear the visible value and the bound state; do not run onChange,
      // because this is an automatic invalidation rather than a user action.
      current = '';
      hidden.value = '';
      input.value = '';
      if (clear) clear.hidden = true;
      if (setter && typeof window[setter] === 'function') {
        setTimeout(function() {
          if (typeof window[setter] === 'function') window[setter]('');
        }, 0);
      }
    }
    function close() {
      popup.hidden = true;
      __SUI_OVERLAY.reset(popup);
      input.setAttribute('aria-expanded', 'false');
      if (toggle) toggle.setAttribute('aria-expanded', 'false');
      input.removeAttribute('aria-activedescendant');
      active = -1;
    }
    function positionPopup(withAnim) {
      if (popup.hidden) return;
      var placed = __SUI_OVERLAY.place(root, popup, {
        placement: root.getAttribute('data-sui-combobox-placement') || 'auto',
        align: 'start',
        gap: 4,
        margin: 8,
        matchWidth: true
      });
      if (withAnim) {
        __SUI_OVERLAY.animate(
          popup,
          root.getAttribute('data-sui-animation') || 'fade',
          root.getAttribute('data-sui-duration'),
          root.getAttribute('data-sui-easing'),
          root.getAttribute('data-sui-delay'),
          placed
        );
      }
    }
    function open() {
      if (disabled || configError) return;
      popup.hidden = false;
      input.setAttribute('aria-expanded', 'true');
      if (toggle) toggle.setAttribute('aria-expanded', 'true');
      render('');
      positionPopup(true);
    }
    function render(query) {
      if (configError) return;
      var q = String(query || '').trim();
      filtered = options.filter(function(o) { return matches(o.label, q); });
      popup.textContent = '';
      if (!filtered.length) {
        var empty = document.createElement('div');
        empty.className = 'sui-combobox-empty';
        empty.textContent = emptyText;
        popup.appendChild(empty);
        active = -1;
        input.removeAttribute('aria-activedescendant');
        positionPopup(false);
        return;
      }
      filtered.forEach(function(o, i) {
        var row = document.createElement('div');
        row.className = 'sui-combobox-option';
        row.id = popup.id + '-option-' + i;
        row.setAttribute('role', 'option');
        row.setAttribute('data-value', o.value);
        row.setAttribute('aria-selected', o.value === current ? 'true' : 'false');
        row.textContent = o.label;
        row.addEventListener('mousedown', function(ev) { ev.preventDefault(); });
        row.addEventListener('click', function() { commit(o.value, o.label); });
        popup.appendChild(row);
      });
      active = -1;
      for (var i = 0; i < filtered.length; i++) {
        if (filtered[i].value === current) { active = i; break; }
      }
      if (active < 0) active = 0;
      highlight();
      positionPopup(false);
    }
    function highlight() {
      var rows = popup.querySelectorAll('.sui-combobox-option');
      for (var i = 0; i < rows.length; i++) {
        var on = i === active;
        rows[i].classList.toggle('active', on);
        rows[i].setAttribute('aria-selected', filtered[i] && filtered[i].value === current ? 'true' : 'false');
      }
      if (active >= 0 && rows[active]) input.setAttribute('aria-activedescendant', rows[active].id);
      else input.removeAttribute('aria-activedescendant');
    }
    function move(delta) {
      if (popup.hidden) open();
      if (!filtered.length) return;
      active = (active + delta + filtered.length) % filtered.length;
      highlight();
      var rows = popup.querySelectorAll('.sui-combobox-option');
      if (rows[active] && rows[active].scrollIntoView) rows[active].scrollIntoView({ block: 'nearest' });
    }
    function commit(value, label) {
      current = value;
      hidden.value = value;
      input.value = label || labelOf(value);
      if (clear) clear.hidden = !value;
      missingRequired = false;
      renderError();
      close();
      if (setter && typeof window[setter] === 'function') {
        window[setter](value);
      } else if (handler && typeof window[handler] === 'function') {
        window[handler]();
      }
    }
    input.addEventListener('focus', function() {
      missingRequired = false;
      renderError();
      if (disabled || configError) return;
      input.select();
      open();
    });
    input.addEventListener('click', function() {
      if (!disabled && !configError && popup.hidden) open();
    });
    input.addEventListener('input', function() {
      missingRequired = false;
      renderError();
      if (configError) return;
      open();
      render(input.value);
    });
    input.addEventListener('keydown', function(ev) {
      if (configError) return;
      if (ev.key === 'ArrowDown') {
        ev.preventDefault();
        if (popup.hidden) open(); else move(1);
      } else if (ev.key === 'ArrowUp') {
        ev.preventDefault();
        if (popup.hidden) open(); else move(-1);
      } else if (ev.key === 'Enter') {
        if (!popup.hidden && active >= 0 && filtered[active]) {
          ev.preventDefault();
          commit(filtered[active].value, filtered[active].label);
        }
      } else if (ev.key === 'Escape') {
        if (!popup.hidden) {
          ev.preventDefault();
          close();
          syncInput();
        }
      } else if (ev.key === 'Tab') {
        close();
      }
    });
    if (toggle) toggle.addEventListener('click', function() {
      if (disabled || configError) return;
      if (popup.hidden) { input.focus(); open(); } else close();
    });
    if (clear) clear.addEventListener('click', function(ev) {
      ev.preventDefault();
      if (configError) return;
      commit('', '');
      input.focus();
    });
    document.addEventListener('pointerdown', function(ev) {
      if (!root.contains(ev.target)) close();
    });
    root.addEventListener('focusout', function(ev) {
      if (!root.contains(ev.relatedTarget)) {
        setTimeout(close, 100);
        if (required && !disabled && !current) {
          missingRequired = true;
          renderError();
        }
      }
    });
    setError(configError);
    clearMissingSelection();
    syncInput();
  }
  function initAll() {
    var all = document.querySelectorAll('.sui-combobox');
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
