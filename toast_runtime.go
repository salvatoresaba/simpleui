package main

// toast_runtime.go - browser runtime for toast notifications.
//
// toast.show(message [, kind [, duration]]) adds a safe text notification to
// the first .sui-toast-host in the page, or to a fixed host created on demand.
// toast.clear() removes all notifications. Message text is assigned with
// textContent; kinds are restricted to info, success, warning and error.
const suiToastRuntime = `(function(){
  var MAX = 4;
  var DEFAULT_MS = 4000;
  var KINDS = { info:1, success:1, warning:1, error:1 };
  var ICONS = { info:'i', success:'\u2713', warning:'!', error:'\u00d7' };

  function findHost() {
    var host = document.querySelector('[data-sui-toast-host]');
    if (!host) host = document.querySelector('.sui-toast-host');
    if (host) return host;
    host = document.createElement('div');
    host.className = 'sui-toast-host sui-toast-host-fixed';
    host.setAttribute('data-sui-toast-host', '1');
    document.body.appendChild(host);
    return host;
  }
  function findHostByIndex(index) {
    var hosts = document.querySelectorAll('[data-sui-toast-host]');
    var i = parseInt(index, 10);
    return i >= 0 && i < hosts.length ? hosts[i] : null;
  }
  function syncHost(host) {
    if (!host) return;
    var x = host.getAttribute('data-sui-toast-x');
    var y = host.getAttribute('data-sui-toast-y');
    if (x) host.style.setProperty('--sui-toast-offset-x', x);
    if (y) host.style.setProperty('--sui-toast-offset-y', y);
  }
  function normalizeKind(kind) {
    kind = String(kind === null || kind === undefined ? '' : kind).toLowerCase();
    return KINDS[kind] ? kind : 'info';
  }
  function readInt(host, attr, fallback) {
    var n = parseInt(host ? host.getAttribute(attr) : '', 10);
    return isNaN(n) ? fallback : n;
  }
  function removeToast(el) {
    if (!el) return;
    if (el.parentNode) el.parentNode.removeChild(el);
  }
  function clear() {
    var hosts = document.querySelectorAll('[data-sui-toast-host], .sui-toast-host');
    for (var i = 0; i < hosts.length; i++) {
      while (hosts[i].firstChild) hosts[i].removeChild(hosts[i].firstChild);
    }
  }
  function showOn(host, message, kind, duration) {
    if (!host) host = findHost();
    syncHost(host);
    if (!host.hasAttribute('aria-live')) host.setAttribute('aria-live', 'polite');
    host.setAttribute('aria-atomic', 'false');
    var k = normalizeKind(kind);
    var el = document.createElement('div');
    el.className = 'sui-toast sui-toast-' + k;
    el.setAttribute('role', 'status');
    var icons = host.getAttribute('data-sui-toast-icons') !== 'false';
    var icon = null;
    if (icons) {
      icon = document.createElement('span');
      icon.className = 'sui-toast-icon';
      icon.setAttribute('aria-hidden', 'true');
      icon.textContent = ICONS[k] || ICONS.info;
    }
    var text = document.createElement('span');
    text.className = 'sui-toast-message';
    text.textContent = String(message === null || message === undefined ? '' : message);
    var close = document.createElement('button');
    close.type = 'button';
    close.className = 'sui-toast-close';
    close.setAttribute('aria-label', 'Dismiss notification');
    close.textContent = '\u00d7';
    close.addEventListener('click', function() { removeToast(el); });
    if (icon) el.appendChild(icon);
    el.appendChild(text);
    el.appendChild(close);
    var max = readInt(host, 'data-sui-toast-max', MAX);
    if (max < 1) max = 1;
    if (max > 8) max = 8;
    while (host.children.length >= max) removeToast(host.firstChild);
    host.appendChild(el);
    var hostDefault = readInt(host, 'data-sui-toast-duration', DEFAULT_MS);
    var ms = duration === undefined || duration === null ? hostDefault : Number(duration);
    if (isNaN(ms)) ms = hostDefault;
    if (ms > 0) {
      setTimeout(function() { removeToast(el); }, ms);
    }
    return el;
  }
  function show(message, kind, duration) {
    return showOn(findHost(), message, kind, duration);
  }
  var initialHosts = document.querySelectorAll('[data-sui-toast-host]');
  for (var hi = 0; hi < initialHosts.length; hi++) syncHost(initialHosts[hi]);
  window.__sui_toast_show = show;
  window.__sui_toast_show_at = function(index, message, kind, duration) {
    return showOn(findHostByIndex(index) || findHost(), message, kind, duration);
  };
  window.__sui_toast_clear = clear;
  // Close buttons inside a re-rendered host still work through delegation.
  document.addEventListener('click', function(ev) {
    var close = ev.target && ev.target.closest ? ev.target.closest('.sui-toast-close') : null;
    if (!close) return;
    var toast = close.closest ? close.closest('.sui-toast') : null;
    if (toast) removeToast(toast);
  });
})();`
