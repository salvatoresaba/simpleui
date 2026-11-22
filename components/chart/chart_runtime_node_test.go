package chart_test

import . "github.com/salvatoresaba/simpleui"

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// chart_runtime_node_test.go runs the generated chart runtime in Node with a
// minimal DOM stub. It covers the parts that pure Go string checks cannot:
// parser output, explicit-id lookup, patch coverage, hint suppression and the
// event queue.
func TestChartRuntimeNodeBehaviour(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed; skipping runtime behaviour test")
	}
	dir := t.TempDir()
	runtimePath := filepath.Join(dir, "chart_runtime.js")
	if err := os.WriteFile(runtimePath, []byte(ChartRuntime), 0o600); err != nil {
		t.Fatalf("write runtime: %v", err)
	}
	harnessPath := filepath.Join(dir, "harness.js")
	if err := os.WriteFile(harnessPath, []byte(chartRuntimeNodeHarness), 0o600); err != nil {
		t.Fatalf("write harness: %v", err)
	}
	cmd := exec.Command(nodePath, harnessPath, runtimePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node runtime test failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "FAIL") {
		t.Fatalf("runtime assertions failed:\n%s", out)
	}
}

const chartRuntimeNodeHarness = `
const fs = require('fs');
let src = fs.readFileSync(process.argv[2], 'utf8');
const marker = src.lastIndexOf('})();');
if (marker < 0) throw new Error('runtime IIFE end not found');
src = src.slice(0, marker) + '\n  return { parseStateData: parseStateData, stateChartFindElement: stateChartFindElement, stateChartApplyPatch: stateChartApplyPatch, stateChartHint: stateChartHint, stateChartHintColor: stateChartHintColor, stateChartFireEvent: stateChartFireEvent, stateChartEventQueue: stateChartEventQueue, stateLineText: stateLineText, stateLineParts: stateLineParts, MAX_TAG_LINES: MAX_TAG_LINES, updateStateGridBackground: updateStateGridBackground };\n})();';
const window = { addEventListener: function(){}, __SUI_USER_FUNCS: Object.create(null) };
const document = { querySelectorAll: function(){ return []; }, addEventListener: function(){}, createElement: function(){ return {}; }, documentElement: {} };
global.getComputedStyle = function(){ return { getPropertyValue: function(){ return ""; } }; };
const mod = { exports: {} };
new Function('module', 'window', 'document', 'module.exports = ' + src)(mod, window, document);
const api = mod.exports;
function ok(cond, msg) {
  if (!cond) { console.error('FAIL: ' + msg); process.exitCode = 1; }
  else console.log('ok: ' + msg);
}

const data = api.parseStateData(JSON.stringify({
  states: [
    { id: 'a', label: 'A', hint: 'HA' },
    { id: 'b', label: 'B', hint: 'none' }
  ],
  transitions: [
    { id: 'e1', from: 'a', to: 'b', hint: 'HE' }
  ]
}), 'json');
ok(data.stateNodes[0].hint === 'HA', 'JSON shape hint parsed');
ok(data.stateNodes[1].hint === 'none', 'JSON hint=none preserved');
ok(data.stateTransitions[0].hint === 'HE', 'JSON link hint parsed');

const csvNode = api.parseStateData('id,label,hint\na,A,CSV A\nb,B,CSV B', 'csv');
ok(csvNode.stateNodes[0].hint === 'CSV A', 'CSV shape hint parsed');
const csvTag = api.parseStateData('id,label,tag,hint\na,A,TAG A,HINT A', 'csv');
ok(csvTag.stateNodes[0].tag === 'TAG A' && csvTag.stateNodes[0].hint === 'HINT A', 'CSV shape tag and hint parsed');
const csvLink = api.parseStateData('From,To,Label,Hint\nA,B,go,CSV link', 'csv');
ok(csvLink.stateTransitions[0].hint === 'CSV link', 'CSV link hint parsed');

ok(api.stateChartFindElement(data, 'a') !== null, 'explicit shape id resolves');
ok(api.stateChartFindElement(data, 'e1') !== null, 'explicit link id resolves');
ok(api.stateChartFindElement(data, 'shape:a') === null, 'prefixed id rejected');
ok(api.stateChartFindElement(data, 'A') === null, 'label rejected');
ok(api.stateChartFindElement(data, 'a->b') === null, 'endpoint pair rejected');

const shape = data.stateNodes[0];
const shapePatch = api.stateChartApplyPatch({ kind: 'shape', item: shape, index: 0 }, {
  width: 200, height: 90, x: 120, y: 100, rotation: 15,
  hint: 'H', tagKind: 'external', tagBox: 'rounded'
}, data);
ok(shapePatch.error === null && shapePatch.changed > 0, 'shape patch applied');
ok(shape.w === 200 && shape.h === 90 && shape.x === 120 && shape.y === 100, 'shape geometry patched');
ok(shape.rotation === 15 && shape.hint === 'H' && shape.tagKind === 'external' && shape.tagBox === 'rounded', 'shape attributes patched');

const link = data.stateTransitions[0];
const linkPatch = api.stateChartApplyPatch({ kind: 'link', item: link, index: 0 }, {
  width: 3, dashStyle: 'dashed', startMarker: 'triangle', endMarker: 'open',
  from: 'a', to: 'b', hint: 'L', waypoints: [{ x: 10, y: 20 }, { x: 30, y: 40 }]
}, data);
ok(linkPatch.error === null && linkPatch.changed > 0, 'link patch applied');
ok(link.width === 3 && link.dashStyle === 'dashed', 'link width and dash patched');
ok(link.startMarker === 'triangle' && link.endMarker === 'open', 'link markers patched');
ok(link.from === 'a' && link.to === 'b', 'link endpoints patched');
ok(Array.isArray(link.waypoints) && link.waypoints.length === 2 && link.waypoints[1].x === 30, 'link waypoints patched');
ok(link.hint === 'L', 'link hint patched');

const onRoot = { getAttribute: function(n) { return n === 'data-sui-chart-state-hints' ? '1' : ''; } };
const offRoot = { getAttribute: function(n) { return n === 'data-sui-chart-state-hints' ? '0' : ''; } };
ok(api.stateChartHint(onRoot, { hint: 'custom' }, 'default') === 'custom', 'custom hint wins');
ok(api.stateChartHint(onRoot, {}, 'default') === 'default', 'empty hint uses default');
ok(api.stateChartHint(onRoot, { hint: 'none' }, 'default') === '', 'hint=none suppresses tooltip');
ok(api.stateChartHint(onRoot, { hint: 'off' }, 'default') === '', 'hint=off suppresses tooltip');
ok(api.stateChartHint(onRoot, { hint: '-' }, 'default') === '', 'hint=- suppresses tooltip');
ok(api.stateChartHint(offRoot, { hint: 'custom' }, 'default') === '', 'global stateHints=false suppresses all');

ok(api.stateChartHintColor('#ff0000', 0.5) === 'rgba(255,0,0,0.5)', 'hint background applies opacity');
ok(api.stateChartHintColor('#ff0000', 1) === 'rgb(255,0,0)', 'hint background is opaque at 1');
ok(api.stateChartHintColor('#ff0000', 0) === 'transparent', 'opacity 0 makes the hint transparent');
ok(api.stateChartHintColor('#0f0', 1) === 'rgb(0,255,0)', 'short hex colors are expanded');
ok(api.stateChartHintColor('transparent', 1) === 'transparent', 'explicit transparent stays transparent');
ok(api.stateChartHintColor('url(https://example.invalid/x)', 1) === 'transparent', 'non-color values cannot inject CSS');

ok(api.stateLineText('A\nB') === 'A|B', 'real newline normalises to pipe');
ok(api.stateLineText('A\r\nB') === 'A|B', 'CRLF normalises to pipe');
ok(api.stateLineParts('A|B').join('/') === 'A/B', 'pipe splits lines');
ok(api.stateLineParts('A||B').join('/') === 'A//B', 'empty lines are preserved');
ok(api.stateLineParts('A\\|B').join('/') === 'A|B', 'escaped pipe stays literal');
const backslashBreak = api.stateLineParts('A\\\\|B');
ok(backslashBreak.length === 2 && backslashBreak[0] === 'A\\' && backslashBreak[1] === 'B', 'escaped backslash before a break');
var manyLines = [];
for (var mi = 0; mi < 150; mi++) manyLines.push('L' + mi);
var cappedLines = api.stateLineParts(manyLines.join('|'), api.MAX_TAG_LINES);
ok(cappedLines.length <= api.MAX_TAG_LINES, 'tag lines are capped');
ok(cappedLines[cappedLines.length - 1].slice(-1) === '…', 'cap marks the truncation');
ok(api.stateChartHint(onRoot, { hint: 'A|B' }, 'default') === 'A\nB', 'pipe becomes a newline in hints');
ok(api.stateChartHint(onRoot, { hint: 'A\\|B' }, 'default') === 'A|B', 'escaped pipe stays literal in hints');
const piped = api.parseStateData(JSON.stringify({ states: [{ id: 'p', label: 'P', hint: 'H1\nH2', tag: 'T1\nT2' }] }), 'json');
ok(piped.stateNodes[0].hint === 'H1|H2', 'parsed hint normalises newlines');
ok(piped.stateNodes[0].tag === 'T1|T2', 'parsed tag normalises newlines');
const modernPiped = api.parseStateData(JSON.stringify({ version: 2, states: [{ id: 'p', label: 'P', hint: 'A|B' }] }), 'json');
ok(modernPiped.stateNodes[0].hint === 'A|B', 'version 2 keeps pipe line breaks');
const pipedPatch = api.stateChartApplyPatch({ kind: 'shape', item: shape, index: 0 }, { hint: 'P1\nP2', tag: 'T1\nT2' }, data);
ok(pipedPatch.error === null && shape.hint === 'P1|P2' && shape.tag === 'T1|T2', 'patch normalises newlines');

const qroot = {
  __suiChartEvents: [],
  getAttribute: function(n) {
    if (n === 'data-sui-chart-state-mode') return 'static';
    if (n === 'data-sui-chart-name') return 'monitor';
    return '';
  }
};
api.stateChartFireEvent(qroot, 'shape', { id: 'a', label: 'A', tag: 'T' }, 'body', 'click', { clientX: 1, clientY: 2, button: 0 });
const queue = api.stateChartEventQueue(qroot);
ok(queue.length === 1, 'event queued');
ok(queue[0].id === 'a' && queue[0].event === 'click' && queue[0].key === undefined, 'event payload uses explicit id only');

const gridViewport = { scrollLeft: 0, scrollTop: 0, style: {} };
const gridRoot = {
  style: {},
  __suiStateWorldPadX: 20,
  __suiStateWorldPadY: 20,
  __suiStateGridZoom: 1,
  __suiStatePanX: 0,
  __suiStatePanY: 0,
  getAttribute: function(n) {
    if (n === 'data-sui-chart-type') return 'state';
    if (n === 'data-sui-chart-state-grid-color') return '#2c3444';
    return null;
  },
  querySelector: function(sel) { return sel === '.sui-chart-viewport' ? gridViewport : null; }
};
api.updateStateGridBackground(gridRoot);
ok(gridViewport.style.backgroundSize === '24px 24px', 'screen grid uses the world step');
ok(gridViewport.style.backgroundPosition === '20px 20px', 'screen grid aligns with the board origin');
gridViewport.scrollLeft = 50;
gridViewport.scrollTop = 30;
api.updateStateGridBackground(gridRoot);
ok(gridViewport.style.backgroundPosition === '-30px -10px', 'screen grid follows scrolling');
gridRoot.__suiStatePanX = 10;
gridRoot.__suiStatePanY = 5;
api.updateStateGridBackground(gridRoot);
ok(gridViewport.style.backgroundPosition === '-20px -5px', 'screen grid follows transform pan');
gridRoot.getAttribute = function(n) {
  if (n === 'data-sui-chart-type') return 'state';
  if (n === 'data-sui-chart-state-grid') return '0';
  return null;
};
api.updateStateGridBackground(gridRoot);
ok(gridViewport.style.backgroundImage === '', 'disabled grid clears the background');
`
