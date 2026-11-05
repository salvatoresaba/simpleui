package main

// modal_runtime.go - browser runtime for the modal/dialog control.
//
// A modal is rendered hidden in the view and opened by name through the
// modal.open(...) builtin. The runtime handles backdrop/Escape dismissal,
// focus restoration, focus trapping and the default modalbutton close action.
const suiModalRuntime = `(function(){
  var openName = '';
  var lastFocus = null;
  var lastFocusID = '';
  var lastClick = null;
  var lastClickID = '';
  var ID_ATTR = 'data-' + 'sui-id';
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function findModal(name) {
    var all = document.querySelectorAll('[data-sui-modal]');
    for (var i = 0; i < all.length; i++) {
      if (all[i].getAttribute('data-sui-modal') === name) return all[i];
    }
    return null;
  }
  function focusFirst(modal) {
    if (!modal || !modal.classList.contains('open')) return;
    var panel = modal.querySelector('.sui-modal-panel');
    if (!panel) return;
    var first = panel.querySelector('[autofocus], button, input, select, textarea, [tabindex]');
    if (first && first.focus) {
      try { first.focus({ preventScroll: true }); } catch (e) { try { first.focus(); } catch (e2) {} }
    } else if (panel.focus) {
      try { panel.focus({ preventScroll: true }); } catch (e) { try { panel.focus(); } catch (e2) {} }
    }
  }
  function setOpen(modal, on) {
    if (!modal) return;
    if (on) {
      modal.hidden = false;
      modal.classList.add('open');
      modal.setAttribute('aria-hidden', 'false');
      // Compensate for the scrollbar that the modal-open class hides. Without
      // this the page width changes and a static chart can visibly jump.
      if (!document.body.classList.contains('sui-modal-open')) {
        var scrollbarWidth = Math.max(0, window.innerWidth - document.documentElement.clientWidth);
        var bodyPadding = parseFloat(window.getComputedStyle(document.body).paddingRight) || 0;
        document.body.__suiModalPaddingRight = document.body.style.paddingRight;
        if (scrollbarWidth > 0) document.body.style.paddingRight = (bodyPadding + scrollbarWidth) + 'px';
      }
      document.body.classList.add('sui-modal-open');
    } else {
      modal.classList.remove('open');
      modal.hidden = true;
      modal.setAttribute('aria-hidden', 'true');
      var any = document.querySelector('.sui-modal.open');
      if (!any) {
        document.body.classList.remove('sui-modal-open');
        if (document.body.__suiModalPaddingRight !== undefined) {
          document.body.style.paddingRight = document.body.__suiModalPaddingRight;
          delete document.body.__suiModalPaddingRight;
        }
      }
    }
  }
  function focusTarget() {
    var target = null;
    if (lastFocusID) {
      try { target = document.querySelector('[' + ID_ATTR + '="' + lastFocusID + '"]'); } catch (e) {}
    }
    if (!target) target = lastFocus;
    if (target && target.focus) {
      try { target.focus({ preventScroll: true }); } catch (e) { try { target.focus(); } catch (e2) {} }
    }
  }
  function restoreFocus() {
    focusTarget();
    // Keep the id through the current event turn: an action handler may call
    // modal.close() and the delegated dismiss listener then runs too. Focus a
    // second time after the click/focus default action has settled.
    var oldFocus = lastFocus;
    var oldID = lastFocusID;
    setTimeout(function() { if (!document.querySelector('.sui-modal.open')) focusTarget(); }, 0);
    setTimeout(function() {
      if (lastFocus === oldFocus && lastFocusID === oldID) {
        lastFocus = null;
        lastFocusID = '';
      }
    }, 40);
  }
  function closeByName(name, restore) {
    var modal = findModal(name);
    if (!modal) return;
    setOpen(modal, false);
    if (openName === name) openName = '';
    if (!document.querySelector('.sui-modal.open')) {
      if (restore) restoreFocus(); else { lastFocus = null; lastFocusID = ''; }
    }
  }
  function openByName(name) {
    name = String(name === null || name === undefined ? '' : name);
    if (!name) return;
    var modal = findModal(name);
    if (!modal) return;
    var current = openName ? findModal(openName) : null;
    if (current && current !== modal) setOpen(current, false);
    lastFocus = document.activeElement;
    lastFocusID = '';
    var holder = closest(lastFocus, '[' + ID_ATTR + ']');
    if (holder) lastFocusID = holder.getAttribute(ID_ATTR) || '';
    // The click capture below remembers the actual trigger even when the
    // browser moves focus to the body before running the handler.
    if (lastClickID) { lastFocus = lastClick; lastFocusID = lastClickID; }
    openName = name;
    setOpen(modal, true);
    setTimeout(function() { focusFirst(modal); }, 0);
  }
  function closeAll(restore) {
    var open = document.querySelectorAll('.sui-modal.open');
    for (var i = 0; i < open.length; i++) setOpen(open[i], false);
    openName = '';
    if (restore) restoreFocus(); else { lastFocus = null; lastFocusID = ''; }
  }
  function applyOpen() {
    if (!openName) return;
    var modal = findModal(openName);
    if (modal && !modal.classList.contains('open')) setOpen(modal, true);
  }
  window.__SUI_MODAL_OPEN = openByName;
  window.__SUI_MODAL_CLOSE = closeByName;
  window.__sui_modal_open = openByName;
  window.__sui_modal_close = closeByName;
  document.addEventListener('click', function(ev) {
    // Capture the element that triggered a modal.open() call. Programmatic
    // clicks in embedded previews do not always focus the button first.
    if (document.querySelector('.sui-modal.open')) return;
    var trigger = closest(ev.target, '[' + ID_ATTR + ']');
    if (trigger) {
      lastClick = trigger;
      lastClickID = trigger.getAttribute(ID_ATTR) || '';
    }
  }, true);
  document.addEventListener('click', function(ev) {
    var backdrop = closest(ev.target, '[data-sui-modal-backdrop]');
    if (backdrop) {
      var modal = closest(backdrop, '.sui-modal');
      if (modal) { closeByName(modal.getAttribute('data-sui-modal'), true); }
      return;
    }
    var dismiss = closest(ev.target, '[data-sui-modal-dismiss]');
    if (dismiss) {
      var owner = closest(dismiss, '.sui-modal');
      if (owner) { closeByName(owner.getAttribute('data-sui-modal'), true); }
    }
  });
  document.addEventListener('keydown', function(ev) {
    var modal = document.querySelector('.sui-modal.open');
    if (!modal) return;
    if (ev.key === 'Escape') {
      ev.preventDefault();
      closeByName(modal.getAttribute('data-sui-modal'), true);
      return;
    }
    if (ev.key === 'Tab') {
      var panel = modal.querySelector('.sui-modal-panel');
      if (!panel) return;
      var nodes = panel.querySelectorAll('button, input, select, textarea, a[href], [tabindex]:not([tabindex="-1"])');
      var list = Array.prototype.filter.call(nodes, function(n) { return !n.disabled && n.offsetParent !== null; });
      if (!list.length) { ev.preventDefault(); if (panel.focus) panel.focus(); return; }
      var first = list[0], last = list[list.length - 1];
      if (ev.shiftKey && document.activeElement === first) { ev.preventDefault(); last.focus(); }
      else if (!ev.shiftKey && document.activeElement === last) { ev.preventDefault(); first.focus(); }
    }
  });
  if (window.MutationObserver) {
    var pending = false;
    new MutationObserver(function() {
      if (pending) return;
      pending = true;
      setTimeout(function() { pending = false; applyOpen(); }, 0);
    }).observe(document.documentElement, { childList: true, subtree: true });
  }
})();`
