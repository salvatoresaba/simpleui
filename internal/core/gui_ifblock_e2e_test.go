package core

import "testing"

// A conditional block with two branches: the designer must keep the elements
// in the branch they belong to, let the condition be edited and show the two
// branches apart in the tree.
const designerIfBlockSource = `state flag = true
state count = 0

view main {
	title "If block"
	col {
		if flag {
			p "then branch"
		} else {
			p "else branch"
		}
	}
}`

func TestGuiIfBlockBranchesE2E(t *testing.T) {
	runDesignerE2E(t, designerIfBlockSource, designerIfBlockScript)
}

// countOf counts the occurrences of a marker in a piece of source.
const designerIfBlockScript = `<script>
(async function () {
  function result(ok, message) {
    var el = document.getElementById('e2e-result');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-result';
      document.body.insertBefore(el, document.body.firstChild);
    }
    el.textContent = ok ? 'OK' : 'FAIL: ' + message;
    el.setAttribute('data-e2e', 'done');
    document.title = ok ? 'E2E_OK' : 'E2E_FAIL';
  }
  function step(name) {
    var el = document.getElementById('e2e-step');
    if (!el) {
      el = document.createElement('div');
      el.id = 'e2e-step';
      document.body.insertBefore(el, document.body.firstChild);
    }
    el.textContent = name;
    document.title = 'E2E:' + name;
  }
  var wait = function (ms) { return new Promise(function (resolve) { setTimeout(resolve, ms); }); };
  async function until(fn, timeoutMs) {
    var deadline = Date.now() + (timeoutMs || 5000);
    for (;;) {
      var value = fn();
      if (value) return value;
      if (Date.now() >= deadline) return null;
      await wait(20);
    }
  }
  function source() { return document.getElementById('out').value || ''; }
  function countOf(text, marker) {
    var total = 0, at = text.indexOf(marker);
    while (at >= 0) { total++; at = text.indexOf(marker, at + marker.length); }
    return total;
  }
  function branches(src) {
    var elseAt = src.indexOf('else {');
    var thenAt = src.indexOf('if ');
    if (elseAt < 0 || thenAt < 0) return null;
    return { then: src.slice(thenAt, elseAt), else: src.slice(elseAt) };
  }
  try {
    step('start');
    var ta = null;
    for (var i = 0; i < 150; i++) {
      ta = document.getElementById('out');
      if (ta && ta.value && document.querySelectorAll('#nodes .node').length) break;
      await wait(100);
    }
    if (!ta) throw new Error('the source editor is missing');
    await wait(1200);
    step('loaded');

    // The tree separates the two branches of the conditional block.
    var sep = await until(function () {
      return document.querySelector('#nodes .node.branchsep') ? true : null;
    }, 5000);
    if (!sep) throw new Error('the else branch is not marked in the tree');
    step('separator');

    // The condition of the block is editable.
    selPath = [0, 0];
    multiSel = [];
    selBody = false;
    renderTree();
    renderInspector();
    await wait(200);
    var blockSec = null;
    var secs = document.querySelectorAll('#inspector .ins-sec');
    for (var s = 0; s < secs.length; s++) {
      var head = secs[s].querySelector('.ins-sec-h');
      if (head && head.textContent.indexOf('Conditional block') === 0) { blockSec = secs[s]; break; }
    }
    if (!blockSec) throw new Error('the conditional block section is missing');
    var cond = blockSec.querySelector('input');
    if (!cond) throw new Error('the condition field is missing');
    if (cond.readOnly) throw new Error('the condition is read-only');
    cond.value = 'count > 2';
    cond.dispatchEvent(new Event('change', { bubbles: true }));
    await wait(400);
    if (source().indexOf('if count > 2') < 0) throw new Error('the condition was not written: ' + source());
    step('condition');

    // An element inserted next to an element of the else branch joins the
    // else branch, not the then branch.
    selPath = [0, 0, 1];
    multiSel = [];
    renderTree();
    renderInspector();
    await wait(200);
    addEl('p');
    await wait(600);
    var src = source();
    var br = branches(src);
    if (!br) throw new Error('the branches are missing: ' + src);
    if (countOf(br.else, 'Paragraph') !== 1) throw new Error('the element did not land in the else branch: ' + src);
    if (countOf(br.then, 'Paragraph') !== 0) throw new Error('the element landed in the then branch: ' + src);
    // The model keeps the branch flag of the new element.
    var block = await until(function () {
      var b = byPath([0, 0], root.children);
      return b && b.children && b.children.length >= 3 ? b : null;
    }, 5000);
    if (!block) throw new Error('the block has no children: ' + JSON.stringify(byPath([0, 0], root.children)));
    if (!block.children[2].else) throw new Error('the new element is not in the else branch: ' + JSON.stringify(block.children));
    step('branch');

    // An element moved onto an element of the other branch follows it.
    dropAt([0, 0, 0].join(','), block.children, 2, block.children[2], [0, 0, 2]);
    await wait(600);
    var after = source();
    var br2 = branches(after);
    if (!br2) throw new Error('the branches are missing after the move: ' + after);
    if (countOf(br2.then, 'then branch') !== 0) throw new Error('the element did not move to the else branch: ' + after);
    if (countOf(br2.else, 'then branch') !== 1) throw new Error('the element is not in the else branch: ' + after);
    step('moved');

    result(true, '');
  } catch (error) {
    result(false, error && error.message ? error.message : String(error));
  }
})();
</script>`
