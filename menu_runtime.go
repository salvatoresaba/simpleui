package main

// menu_runtime.go - browser runtime for the menu-bar control.
//
// The generated markup is a <nav class="sui-menubar"> containing
// <div class="sui-menu"> items. Each menu is a real <button> plus a hidden
// popup. A menu nested in another popup is a submenu and uses the same markup.
// This script toggles the `open` class, keeps ARIA state in sync and positions
// every popup inside the visible viewport (flipping above and clamping left/
// right when the trigger is near an edge).
const suiMenuRuntime = `(function(){
  function closest(el, sel) {
    return el && el.closest ? el.closest(sel) : null;
  }
  function popupOf(menu) {
    if (!menu) return null;
    for (var i = 0; i < menu.children.length; i++) {
      var c = menu.children[i];
      if (c.classList && c.classList.contains('sui-menu-popup')) return c;
    }
    return null;
  }
  function resetPopup(popup) {
    if (popup) __SUI_OVERLAY.reset(popup);
  }
  // positionMenu uses the shared overlay helper: it flips above when there is
  // no space below, clamps at the viewport edges and scrolls when too tall.
  function positionMenu(menu) {
    if (!menu || !menu.classList.contains('open')) return;
    var popup = popupOf(menu);
    if (!popup) return;
    var trigger = menu.querySelector('.sui-menu-button') || menu;
    var rightAligned = menu.hasAttribute('data-sui-dropdown-right') || menu.classList.contains('sui-dropdown-right');
    var placement = isNested(menu) ? 'right' : (menu.getAttribute('data-sui-dropdown-placement') || 'auto');
    var placed = __SUI_OVERLAY.place(trigger, popup, {
      placement: placement,
      align: rightAligned ? 'end' : 'start',
      gap: 4,
      margin: 8
    });
    __SUI_OVERLAY.animate(
      popup,
      menu.getAttribute('data-sui-animation') || 'fade',
      menu.getAttribute('data-sui-duration'),
      menu.getAttribute('data-sui-easing'),
      menu.getAttribute('data-sui-delay'),
      placed
    );
  }
  function closeMenus(except) {
    var open = document.querySelectorAll('.sui-menu.open');
    for (var i = 0; i < open.length; i++) {
      if (open[i] === except) continue;
      open[i].classList.remove('open');
      var p = popupOf(open[i]);
      if (p) resetPopup(p);
      var b = open[i].querySelector('.sui-menu-button');
      if (b) b.setAttribute('aria-expanded', 'false');
    }
  }
  function openMenu(menu) {
    if (!menu) return;
    var deeper = menu.querySelectorAll('.sui-menu.open');
    for (var i = 0; i < deeper.length; i++) {
      deeper[i].classList.remove('open');
      var dp = popupOf(deeper[i]);
      if (dp) resetPopup(dp);
      var db = deeper[i].querySelector('.sui-menu-button');
      if (db) db.setAttribute('aria-expanded', 'false');
    }
    menu.classList.add('open');
    var b = menu.querySelector('.sui-menu-button');
    if (b) b.setAttribute('aria-expanded', 'true');
    positionMenu(menu);
  }
  function isNested(menu) {
    return !!(menu && menu.parentNode && menu.parentNode.classList &&
      menu.parentNode.classList.contains('sui-menu-popup'));
  }
  function closeBranch(menu) {
    if (!menu) return;
    menu.classList.remove('open');
    var bp = popupOf(menu);
    if (bp) resetPopup(bp);
    var b = menu.querySelector('.sui-menu-button');
    if (b) b.setAttribute('aria-expanded', 'false');
    var nested = menu.querySelectorAll('.sui-menu.open');
    for (var i = 0; i < nested.length; i++) {
      nested[i].classList.remove('open');
      var np = popupOf(nested[i]);
      if (np) resetPopup(np);
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
      var sp = popupOf(open[i]);
      if (sp) resetPopup(sp);
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
      var menubar = closest(btn, '.sui-menubar');
      var topLevel = !!(menubar && !closest(btn, '.sui-menu-popup'));
      if (topLevel && (ev.key === 'ArrowLeft' || ev.key === 'ArrowRight')) {
        ev.preventDefault();
        var buttons = [];
        var allButtons = menubar.querySelectorAll('.sui-menu-button');
        for (var bi = 0; bi < allButtons.length; bi++) {
          if (!closest(allButtons[bi], '.sui-menu-popup')) buttons.push(allButtons[bi]);
        }
        var ci = buttons.indexOf(btn);
        if (ci >= 0) {
          var ni = (ci + (ev.key === 'ArrowRight' ? 1 : -1) + buttons.length) % buttons.length;
          var nextButton = buttons[ni];
          nextButton.focus();
          if (document.querySelector('.sui-menu.open')) {
            closeMenus(null);
            openMenu(nextButton.parentNode);
            focusItem(nextButton.parentNode, 0);
          }
        }
        return;
      }
      if (ev.key === 'Home' || ev.key === 'End') {
        ev.preventDefault();
        var listMenu = menu;
        if (isNested(menu) && !menu.classList.contains('open')) {
          listMenu = closest(menu.parentNode, '.sui-menu') || menu;
        }
        if (listMenu && listMenu.classList.contains('open')) {
          focusItem(listMenu, ev.key === 'Home' ? 0 : -1);
        }
        return;
      }
      if (ev.key === 'ArrowDown' || ev.key === 'ArrowUp' || (ev.key === 'ArrowRight' && !topLevel)) {
        ev.preventDefault();
        if (isNested(menu)) closeSiblingMenus(menu); else closeMenus(null);
        openMenu(menu);
        focusItem(menu, ev.key === 'ArrowUp' ? -1 : 0);
        return;
      }
      if (ev.key === 'ArrowLeft' && isNested(menu)) {
        ev.preventDefault();
        closeBranch(menu);
        var own = menu.querySelector('.sui-menu-button');
        if (own) own.focus();
        return;
      }
    }
    var item = closest(ev.target, '.sui-menuitem');
    if (item && (ev.key === 'Enter' || ev.key === ' ' || ev.key === 'Spacebar')) {
      ev.preventDefault();
      item.click();
      return;
    }
    if (item && (ev.key === 'ArrowDown' || ev.key === 'ArrowUp' || ev.key === 'Home' || ev.key === 'End')) {
      ev.preventDefault();
      var itemMenu = closest(item, '.sui-menu');
      var items = directItems(itemMenu);
      var index = Array.prototype.indexOf.call(items, item);
      if (ev.key === 'Home') focusItem(itemMenu, 0);
      else if (ev.key === 'End') focusItem(itemMenu, -1);
      else focusItem(itemMenu, index + (ev.key === 'ArrowUp' ? -1 : 1));
      return;
    }
    if (item && ev.key === 'ArrowLeft') {
      var itemMenu2 = closest(item, '.sui-menu');
      if (isNested(itemMenu2)) {
        ev.preventDefault();
        var own2 = itemMenu2.querySelector('.sui-menu-button');
        closeBranch(itemMenu2);
        if (own2) own2.focus();
      }
      return;
    }
    if (ev.key === 'Escape') {
      var openNow = document.querySelectorAll('.sui-menu.open');
      if (openNow.length) {
        var firstButton = null;
        for (var oi = 0; oi < openNow.length; oi++) {
          if (!isNested(openNow[oi])) { firstButton = openNow[oi].querySelector('.sui-menu-button'); break; }
        }
        closeMenus(null);
        if (firstButton) firstButton.focus();
        ev.preventDefault();
      }
    }
  });
})();`
