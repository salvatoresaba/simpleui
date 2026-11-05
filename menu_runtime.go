package main

// menu_runtime.go - browser runtime for the menu-bar control.
//
// The generated markup is a <nav class="sui-menubar"> containing
// <div class="sui-menu"> items. Each menu is a real <button> plus a hidden
// popup. A menu nested in another popup is a submenu and uses the same markup;
// CSS positions it beside its parent. This script only toggles the `open` class
// and keeps ARIA state in sync; labels and links are rendered/escaped by the
// normal generator.
const suiMenuRuntime = `(function(){
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function closeMenus(except) {
    var open = document.querySelectorAll('.sui-menu.open');
    for (var i = 0; i < open.length; i++) {
      if (open[i] === except) continue;
      open[i].classList.remove('open');
      var b = open[i].querySelector('.sui-menu-button');
      if (b) b.setAttribute('aria-expanded', 'false');
    }
  }
  function openMenu(menu) {
    if (!menu) return;
    var deeper = menu.querySelectorAll('.sui-menu.open');
    for (var i = 0; i < deeper.length; i++) {
      deeper[i].classList.remove('open');
      var db = deeper[i].querySelector('.sui-menu-button');
      if (db) db.setAttribute('aria-expanded', 'false');
    }
    menu.classList.add('open');
    var b = menu.querySelector('.sui-menu-button');
    if (b) b.setAttribute('aria-expanded', 'true');
  }
  function isNested(menu) {
    return !!(menu && menu.parentNode && menu.parentNode.classList &&
      menu.parentNode.classList.contains('sui-menu-popup'));
  }
  function closeBranch(menu) {
    if (!menu) return;
    menu.classList.remove('open');
    var b = menu.querySelector('.sui-menu-button');
    if (b) b.setAttribute('aria-expanded', 'false');
    var nested = menu.querySelectorAll('.sui-menu.open');
    for (var i = 0; i < nested.length; i++) {
      nested[i].classList.remove('open');
      var nb = nested[i].querySelector('.sui-menu-button');
      if (nb) nb.setAttribute('aria-expanded', 'false');
    }
  }
  function closeSiblingMenus(menu) {
    var popup = isNested(menu) ? menu.parentNode : null;
    if (!popup) {
      closeMenus(menu);
      return;
    }
    var open = popup.querySelectorAll('.sui-menu.open');
    for (var i = 0; i < open.length; i++) {
      if (open[i] === menu) continue;
      open[i].classList.remove('open');
      var b = open[i].querySelector('.sui-menu-button');
      if (b) b.setAttribute('aria-expanded', 'false');
    }
  }
  // directItems returns only the entries of THIS menu popup, not nested
  // submenu entries. Submenu buttons are included so arrow keys can reach them.
  function directItems(menu) {
    if (!menu) return [];
    var popup = null;
    for (var i = 0; i < menu.children.length; i++) {
      var c = menu.children[i];
      if (c.classList && c.classList.contains('sui-menu-popup')) { popup = c; break; }
    }
    if (!popup) return [];
    var out = [];
    for (var j = 0; j < popup.children.length; j++) {
      var item = popup.children[j];
      if (item.classList && item.classList.contains('sui-menuitem')) {
        out.push(item);
      } else if (item.classList && item.classList.contains('sui-menu')) {
        var b = item.querySelector('.sui-menu-button');
        if (b) out.push(b);
      } else if (item.tagName === 'A') {
        var mi = item.querySelector('.sui-menuitem');
        if (mi) out.push(mi);
      }
    }
    return out;
  }
  function focusItem(menu, index) {
    var items = directItems(menu);
    if (!items.length) return;
    if (index < 0) index = items.length - 1;
    if (index >= items.length) index = 0;
    items[index].focus();
  }
  document.addEventListener('click', function(ev) {
    var btn = closest(ev.target, '.sui-menu-button');
    if (btn) {
      var menu = btn.parentNode;
      var wasOpen = menu && menu.classList && menu.classList.contains('open');
      if (isNested(menu)) {
        closeSiblingMenus(menu);
        if (wasOpen) closeBranch(menu); else openMenu(menu);
      } else {
        closeMenus(null);
        if (!wasOpen) openMenu(menu);
      }
      ev.preventDefault();
      return;
    }
    if (closest(ev.target, '.sui-menuitem')) {
      closeMenus(null);
      return;
    }
    if (!closest(ev.target, '.sui-menu')) closeMenus(null);
  });
  document.addEventListener('keydown', function(ev) {
    var btn = closest(ev.target, '.sui-menu-button');
    if (btn) {
      var menu = btn.parentNode;
      if (ev.key === 'ArrowDown' || ev.key === 'ArrowUp' || ev.key === 'ArrowRight') {
        ev.preventDefault();
        if (isNested(menu)) closeSiblingMenus(menu); else closeMenus(null);
        openMenu(menu);
        focusItem(menu, ev.key === 'ArrowUp' ? -1 : 0);
        return;
      }
      if (ev.key === 'ArrowLeft' && isNested(menu)) {
        ev.preventDefault();
        closeBranch(menu);
        var parent = closest(menu.parentNode, '.sui-menu');
        if (parent) {
          var pb = parent.querySelector('.sui-menu-button');
          if (pb) pb.focus();
        }
        return;
      }
    }
    var item = closest(ev.target, '.sui-menuitem');
    if (item && (ev.key === 'ArrowDown' || ev.key === 'ArrowUp')) {
      ev.preventDefault();
      var itemMenu = closest(item, '.sui-menu');
      var items = directItems(itemMenu);
      var index = Array.prototype.indexOf.call(items, item);
      focusItem(itemMenu, index + (ev.key === 'ArrowUp' ? -1 : 1));
      return;
    }
    if (item && ev.key === 'ArrowLeft') {
      var itemMenu2 = closest(item, '.sui-menu');
      if (isNested(itemMenu2)) {
        ev.preventDefault();
        closeBranch(itemMenu2);
        var parent2 = closest(itemMenu2.parentNode, '.sui-menu');
        if (parent2) {
          var pb2 = parent2.querySelector('.sui-menu-button');
          if (pb2) pb2.focus();
        }
      }
      return;
    }
    if (ev.key === 'Escape') {
      var top = document.querySelector('.sui-menubar > .sui-menu.open');
      if (top) {
        var button = top.querySelector('.sui-menu-button');
        closeMenus(null);
        if (button) button.focus();
        ev.preventDefault();
      }
    }
  });
})();`
