package main

// datepicker_runtime.go - runtime for native and customizable date pickers.
const suiDatePickerRuntime = `(function(){
  var seq = 0;
  function validFormat(value, mode) {
    if (!value) return true;
    if (mode === 'time') return /^\d{2}:\d{2}(:\d{2})?$/.test(value);
    if (mode === 'datetime') return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?$/.test(value);
    return /^\d{4}-\d{2}-\d{2}$/.test(value);
  }
  function parseDate(value) {
    var match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value || '');
    if (!match) return null;
    var date = new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3])));
    if (date.getUTCFullYear() !== Number(match[1]) || date.getUTCMonth() !== Number(match[2]) - 1 || date.getUTCDate() !== Number(match[3])) return null;
    return date;
  }
  function dateValue(date) {
    return date.getUTCFullYear() + '-' + String(date.getUTCMonth() + 1).padStart(2, '0') + '-' + String(date.getUTCDate()).padStart(2, '0');
  }
  function localToday() {
    var date = new Date();
    return date.getFullYear() + '-' + String(date.getMonth() + 1).padStart(2, '0') + '-' + String(date.getDate()).padStart(2, '0');
  }
  function datePart(value) {
    return value && value.indexOf('T') >= 0 ? value.slice(0, value.indexOf('T')) : value;
  }
  function displayCalendarValue(value, mode) {
    if (!value) return '';
    var match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(datePart(value));
    if (!match) return value;
    var display = match[3] + '/' + match[2] + '/' + match[1];
    if (mode === 'datetime' && value.indexOf('T') >= 0) display += ' ' + value.slice(value.indexOf('T') + 1);
    return display;
  }
  function maskCalendarInput(value, mode) {
    var limit = mode === 'datetime' ? 14 : 8;
    var digits = String(value || '').replace(/\D/g, '').slice(0, limit);
    var masked = digits.slice(0, 2);
    if (digits.length >= 2) masked += '/';
    if (digits.length > 2) masked += digits.slice(2, 4);
    if (digits.length >= 4) masked += '/';
    if (digits.length > 4) masked += digits.slice(4, 8);
    if (mode === 'datetime' && digits.length >= 8) masked += ' ';
    if (mode === 'datetime' && digits.length > 8) masked += digits.slice(8, 10);
    if (mode === 'datetime' && digits.length >= 10) masked += ':';
    if (mode === 'datetime' && digits.length > 10) masked += digits.slice(10, 12);
    if (mode === 'datetime' && digits.length > 12) masked += ':' + digits.slice(12, 14);
    return masked;
  }
  function parseCalendarInput(value, mode) {
    var match;
    if (mode === 'datetime') {
      match = /^(\d{2})\/(\d{2})\/(\d{4}) (\d{2}):(\d{2})(?::(\d{2}))?$/.exec(value || '');
    } else {
      match = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(value || '');
    }
    if (!match) return '';
    var date = match[3] + '-' + match[2] + '-' + match[1];
    if (!parseDate(date)) return '';
    if (mode !== 'datetime') return date;
    var hour = Number(match[4]);
    var minute = Number(match[5]);
    var second = match[6] ? Number(match[6]) : -1;
    if (hour > 23 || minute > 59 || second > 59) return '';
    return date + 'T' + match[4] + ':' + match[5] + (match[6] ? ':' + match[6] : '');
  }
  function calendarInputError(mode) {
    return mode === 'datetime' ? 'Enter a valid date and time as DD/MM/YYYY HH:MM' : 'Enter a valid date as DD/MM/YYYY';
  }
  function cursorAfterDigits(value, count) {
    var index = 0;
    var digits = 0;
    while (index < value.length && digits < count) {
      if (/\d/.test(value.charAt(index))) digits++;
      index++;
    }
    while (index < value.length && !/\d/.test(value.charAt(index))) index++;
    return index;
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
    var variant = (root.getAttribute('variant') || 'native').toLowerCase();
    var customCalendar = variant === 'calendar' && mode !== 'time';
    var setter = root.getAttribute('data-sui-datepicker-setter') || '';
    var handler = root.getAttribute('data-sui-datepicker-onchange') || '';
    var disabled = root.getAttribute('data-sui-disabled') === '1';
    var min = input.getAttribute('min') || '';
    var max = input.getAttribute('max') || '';
    var required = input.required;
    var current = hidden.value || input.value || '';
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
      if (!value) return required && customCalendar && !disabled ? 'A value is required' : '';
      if (!validFormat(value, mode)) return 'Invalid ' + mode + ' value';
      if (min && value < min) return 'Value is before the minimum';
      if (max && value > max) return 'Value is after the maximum';
      return '';
    }
    function sync() {
      input.value = customCalendar ? displayCalendarValue(current, mode) : current;
      setError(validate(current));
      if (clear) clear.hidden = !current;
    }
    function commit(value) {
      current = value;
      hidden.value = value;
      var display = customCalendar ? displayCalendarValue(value, mode) : value;
      if (input.value !== display) input.value = display;
      if (clear) clear.hidden = !value;
      setError(validate(value));
      if (setter && typeof window[setter] === 'function') window[setter](value);
      if (handler && typeof window[handler] === 'function') window[handler]();
    }
    if (customCalendar) {
      root.setAttribute('data-sui-custom-calendar', '1');
      input.type = 'text';
      input.readOnly = false;
      input.setAttribute('placeholder', mode === 'datetime' ? 'DD/MM/YYYY HH:MM' : 'DD/MM/YYYY');
      input.setAttribute('aria-haspopup', 'dialog');
      input.setAttribute('aria-expanded', 'false');
      input.setAttribute('autocomplete', 'off');
      var control = input.parentNode;
      var open = document.createElement('button');
      open.type = 'button';
      open.className = 'sui-datepicker-open';
      open.textContent = '▦';
      open.disabled = disabled;
      open.setAttribute('aria-label', 'Choose date');
      open.setAttribute('aria-haspopup', 'dialog');
      open.setAttribute('aria-expanded', 'false');
      var popup = document.createElement('div');
      popup.className = 'sui-datepicker-calendar';
      popup.hidden = true;
      popup.setAttribute('role', 'dialog');
      popup.setAttribute('aria-label', 'Choose date');
      popup.id = 'sui-datepicker-calendar-' + (++seq);
      input.setAttribute('aria-controls', popup.id);
      open.setAttribute('aria-controls', popup.id);
      var header = document.createElement('div');
      header.className = 'sui-datepicker-calendar-header';
      var previous = document.createElement('button');
      previous.type = 'button';
      previous.className = 'sui-datepicker-month-previous';
      previous.textContent = '‹';
      previous.setAttribute('aria-label', 'Previous month');
      var monthTitle = document.createElement('strong');
      monthTitle.className = 'sui-datepicker-month-title';
      monthTitle.setAttribute('aria-live', 'polite');
      var next = document.createElement('button');
      next.type = 'button';
      next.className = 'sui-datepicker-month-next';
      next.textContent = '›';
      next.setAttribute('aria-label', 'Next month');
      header.appendChild(previous);
      header.appendChild(monthTitle);
      header.appendChild(next);
      var weekdays = document.createElement('div');
      weekdays.className = 'sui-datepicker-weekdays';
      weekdays.setAttribute('aria-hidden', 'true');
      var monday = new Date(Date.UTC(2024, 0, 1));
      for (var wd = 0; wd < 7; wd++) {
        var weekday = document.createElement('span');
        weekday.textContent = new Intl.DateTimeFormat(undefined, { weekday: 'short', timeZone: 'UTC' }).format(monday);
        weekdays.appendChild(weekday);
        monday.setUTCDate(monday.getUTCDate() + 1);
      }
      var grid = document.createElement('div');
      grid.className = 'sui-datepicker-days';
      grid.setAttribute('role', 'grid');
      popup.appendChild(header);
      popup.appendChild(weekdays);
      popup.appendChild(grid);
      if (control) {
        if (clear && clear.parentNode === control) control.insertBefore(open, clear);
        else control.appendChild(open);
      }
      root.appendChild(popup);
      var timeInput = null;
      if (mode === 'datetime') {
        var timeWrap = document.createElement('label');
        timeWrap.className = 'sui-datepicker-time-wrap';
        timeWrap.textContent = 'Time ';
        timeInput = document.createElement('input');
        timeInput.type = 'time';
        timeInput.className = 'sui-datepicker-time';
        var step = input.getAttribute('step');
        if (step) timeInput.setAttribute('step', step);
        timeInput.disabled = disabled;
        timeWrap.appendChild(timeInput);
        popup.appendChild(timeWrap);
      }
      var viewing = parseDate(datePart(current)) || parseDate(localToday());
      viewing = new Date(Date.UTC(viewing.getUTCFullYear(), viewing.getUTCMonth(), 1));
      function monthText(date) {
        return new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(date);
      }
      function focusDay(value) {
        var buttons = grid.querySelectorAll('button[data-date]');
        var fallback = null;
        for (var i = 0; i < buttons.length; i++) {
          if (buttons[i].disabled) continue;
          if (!fallback) fallback = buttons[i];
          if (buttons[i].getAttribute('data-date') === value) {
            buttons[i].focus();
            return;
          }
        }
        if (fallback) fallback.focus();
      }
      function renderCalendar() {
        while (grid.firstChild) grid.removeChild(grid.firstChild);
        monthTitle.textContent = monthText(viewing);
        var year = viewing.getUTCFullYear();
        var month = viewing.getUTCMonth();
        var first = new Date(Date.UTC(year, month, 1));
        var offset = (first.getUTCDay() + 6) % 7;
        var selected = datePart(current);
        var today = localToday();
        var minDate = datePart(min);
        var maxDate = datePart(max);
        for (var dayIndex = 0; dayIndex < 42; dayIndex++) {
          var day = new Date(Date.UTC(year, month, 1 - offset + dayIndex));
          var value = dateValue(day);
          var button = document.createElement('button');
          button.type = 'button';
          button.className = 'sui-datepicker-day';
          button.textContent = String(day.getUTCDate());
          button.setAttribute('data-date', value);
          button.setAttribute('aria-label', new Intl.DateTimeFormat(undefined, { dateStyle: 'full', timeZone: 'UTC' }).format(day));
          button.setAttribute('aria-pressed', value === selected ? 'true' : 'false');
          button.tabIndex = -1;
          if (day.getUTCMonth() !== month) button.classList.add('is-outside-month');
          if (value === selected) button.classList.add('is-selected');
          if (value === today) {
            button.classList.add('is-today');
            button.setAttribute('aria-current', 'date');
          }
          button.disabled = !!((minDate && value < minDate) || (maxDate && value > maxDate));
          button.addEventListener('click', function(ev) {
            var chosen = ev.currentTarget.getAttribute('data-date');
            var time = timeInput ? (timeInput.value || '09:00') : '';
            if (timeInput) {
              updateTimeBounds(chosen);
              var selectedMin = timeInput.min || '';
              var selectedMax = timeInput.max || '';
              if (selectedMin && time < selectedMin) time = selectedMin;
              if (selectedMax && time > selectedMax) time = selectedMax;
              timeInput.value = time;
            }
            commit(mode === 'datetime' ? chosen + 'T' + time : chosen);
            closeCalendar(true);
          });
          grid.appendChild(button);
        }
      }
      function updateTimeBounds(value) {
        if (!timeInput) return;
        var selectedMin = datePart(min) === value && min.indexOf('T') >= 0 ? min.slice(min.indexOf('T') + 1) : '';
        var selectedMax = datePart(max) === value && max.indexOf('T') >= 0 ? max.slice(max.indexOf('T') + 1) : '';
        if (selectedMin) timeInput.min = selectedMin;
        else timeInput.removeAttribute('min');
        if (selectedMax) timeInput.max = selectedMax;
        else timeInput.removeAttribute('max');
      }
      function closeCalendar(restoreFocus) {
        popup.hidden = true;
        root.removeAttribute('data-sui-calendar-open');
        input.setAttribute('aria-expanded', 'false');
        open.setAttribute('aria-expanded', 'false');
        if (restoreFocus) input.focus();
      }
      function openCalendar() {
        var selected = parseDate(datePart(current));
        if (selected) viewing = new Date(Date.UTC(selected.getUTCFullYear(), selected.getUTCMonth(), 1));
        renderCalendar();
        if (timeInput) timeInput.value = current.indexOf('T') >= 0 ? current.slice(current.indexOf('T') + 1) : '09:00';
        updateTimeBounds(datePart(current));
        popup.hidden = false;
        root.setAttribute('data-sui-calendar-open', '1');
        input.setAttribute('aria-expanded', 'true');
        open.setAttribute('aria-expanded', 'true');
        focusDay(datePart(current) || localToday());
      }
      function moveMonth(amount) {
        viewing = new Date(Date.UTC(viewing.getUTCFullYear(), viewing.getUTCMonth() + amount, 1));
        renderCalendar();
        focusDay(dateValue(viewing));
      }
      open.addEventListener('click', openCalendar);
      function handleCalendarInput(finalize) {
        var caret = input.selectionStart || 0;
        var digitsBeforeCaret = (input.value.slice(0, caret).match(/\d/g) || []).length;
        var masked = maskCalendarInput(input.value, mode);
        if (input.value !== masked) {
          input.value = masked;
          var nextCaret = cursorAfterDigits(masked, digitsBeforeCaret);
          input.setSelectionRange(nextCaret, nextCaret);
        }
        var parsed = parseCalendarInput(masked, mode);
        if (parsed) {
          if (parsed !== current) commit(parsed);
          else setError(validate(parsed));
          return;
        }
        if (!masked) {
          if (current) commit('');
          else setError(validate(''));
          return;
        }
        setError('');
        if (finalize) {
          if (current) commit('');
          input.value = masked;
          setError(calendarInputError(mode));
        }
      }
      input.addEventListener('input', function() { handleCalendarInput(false); });
      input.addEventListener('change', function() { handleCalendarInput(true); });
      input.addEventListener('blur', function() { handleCalendarInput(true); });
      input.addEventListener('paste', function(ev) {
        var clipboard = ev.clipboardData;
        var pasted = clipboard ? clipboard.getData('text').trim() : '';
        var iso = /^(\d{4})-(\d{2})-(\d{2})(?:T(\d{2}):(\d{2})(?::(\d{2}))?)?$/.exec(pasted);
        if (!iso) return;
        ev.preventDefault();
        input.value = iso[3] + '/' + iso[2] + '/' + iso[1];
        if (mode === 'datetime' && iso[4]) input.value += ' ' + iso[4] + ':' + iso[5] + (iso[6] ? ':' + iso[6] : '');
        input.setSelectionRange(input.value.length, input.value.length);
        handleCalendarInput(false);
      });
      input.addEventListener('keydown', function(ev) {
        if (ev.key === 'Enter' || ev.key === ' ' || ev.key === 'ArrowDown') {
          ev.preventDefault();
          openCalendar();
        }
      });
      previous.addEventListener('click', function() { moveMonth(-1); });
      next.addEventListener('click', function() { moveMonth(1); });
      grid.addEventListener('keydown', function(ev) {
        var value = ev.target.getAttribute('data-date');
        var date = parseDate(value);
        if (!date) return;
        var amount = 0;
        if (ev.key === 'ArrowLeft') amount = -1;
        else if (ev.key === 'ArrowRight') amount = 1;
        else if (ev.key === 'ArrowUp') amount = -7;
        else if (ev.key === 'ArrowDown') amount = 7;
        else if (ev.key === 'Home') amount = -((date.getUTCDay() + 6) % 7);
        else if (ev.key === 'End') amount = 6 - ((date.getUTCDay() + 6) % 7);
        else if (ev.key === 'PageUp' || ev.key === 'PageDown') {
          ev.preventDefault();
          viewing = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + (ev.key === 'PageUp' ? -1 : 1), 1));
          renderCalendar();
          focusDay(dateValue(viewing));
          return;
        } else return;
        ev.preventDefault();
        date.setUTCDate(date.getUTCDate() + amount);
        if (date.getUTCMonth() !== viewing.getUTCMonth()) {
          viewing = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1));
          renderCalendar();
        }
        focusDay(dateValue(date));
      });
      popup.addEventListener('keydown', function(ev) {
        if (ev.key === 'Escape') {
          ev.preventDefault();
          closeCalendar(true);
        }
      });
      if (timeInput) timeInput.addEventListener('change', function() {
        var date = datePart(current);
        if (date) commit(date + 'T' + timeInput.value);
      });
      document.addEventListener('pointerdown', function outsideClick(ev) {
        if (!root.isConnected) {
          document.removeEventListener('pointerdown', outsideClick);
          return;
        }
        if (!root.contains(ev.target) && !popup.hidden) closeCalendar(false);
      });
    } else {
      input.addEventListener('change', function() { commit(input.value); });
    }
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
