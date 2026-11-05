package main

// guihtml.go - the embedded visual design tool (structured CSS engine UI).
// Plain HTML/CSS/JS shipped inside the single compiler binary.
//
// Pillars:
//   - CSS is a data model (AST), not a string: /__css__/parse + /__css__/set
//     let the browser parse/generate/edit the stylesheet.
//   - two-way binding with a single source of truth (root.css); programmatic
//     edits set .value (which does not fire oninput) so no sync loops.
//   - sandboxed live preview compiled by the compiler (/__gui__/render).
//
// Layout (dockable panels, dark theme):
//   - top bar: breakpoints, export, status
//   - left:  Structure (palette + tree + element attrs) | Tokens
//   - center: Preview | SUI | CSS (editor with line numbers)
//   - right: Inspector (state, specificity, visual styler)
//   - bottom: Problems | History (collapsible)

var guiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>SimpleUI Design Tool</title>
<style>
:root {
  --fs: 14px;            /* base UI font size */
  --bg: #0f1117; --panel: #12141c; --panel2: #161922; --panel3: #1b1f2a;
  --border: #22242f; --border2: #3a4152;
  --fg: #e6e6ef; --muted: #8fa0b8; --dim: #6b7690;
  --accent: #4c5cf7;
  --editorbg: #141824; --editorfg: #dfe6f0;
  --edtok-fn: #e5c07b; --edtok-var: #56b6c2;
  --btn-bg: #262b3a; --btn-hover: #313849; --btn-on: #33405c;
  --palbtn-bg: #222735; --sel: #2a3050; --input-bg: #10121a;
  --gutter-bg: #10121a; --gutter-fg: #4a5570;
}
* { box-sizing: border-box; }
html, body { margin:0; height:100%; }
body { font-family: system-ui, "Segoe UI", Roboto, sans-serif; background: var(--bg); color: var(--fg); display:flex; flex-direction:column; overflow:hidden; font-size: var(--fs); }
::-webkit-scrollbar { width:11px; height:11px; }
::-webkit-scrollbar-thumb { background:#2a2f3d; border-radius:6px; }

/* top bar */
#topbar { display:flex; align-items:center; gap:7px; padding:7px 12px; border-bottom:1px solid var(--border); background: var(--panel2); flex-wrap:wrap; }
#topbar h1 { font-size:14px; margin:0 6px 0 0; font-weight:600; }

/* menu bar */
.menu { position:relative; }
.mitem { display:block; padding:5px 10px; font-size:13px; border-radius:5px; cursor:pointer; user-select:none; color:var(--fg); }
.mitem:hover, .menu.open .mitem { background:var(--btn-bg); }
.drop { display:none; position:absolute; top:100%; left:0; margin-top:2px; background:var(--panel2); border:1px solid var(--border2); border-radius:6px; min-width:200px; padding:4px; z-index:60; box-shadow:0 8px 24px rgba(0,0,0,.45); }
.menu.open .drop { display:block; }
.mi { padding:6px 10px; font-size:13px; border-radius:5px; cursor:pointer; color:var(--fg); white-space:nowrap; }
/* cascade submenu: every menu row is a "has-sub" group; the first flyout
   level opens to the right of the top row, deeper levels chain further */
.mi.has-sub { display:flex; align-items:center; position:relative; font-weight:600; }
.mi-sub > .mi { font-weight:400; }
.mi-sub > .mi.has-sub { font-weight:600; }
.sub-arrow { margin-left:auto; padding-left:12px; font-size:9px; color:var(--dim); }
.mi.has-sub:hover { background:var(--btn-hover); }
.mi-sub { display:none; position:absolute; left:calc(100% - 2px); top:-5px; min-width:130px; background:var(--panel2); border:1px solid var(--border2); border-radius:6px; padding:3px; z-index:70; box-shadow:0 8px 24px rgba(0,0,0,.45); }
.mi.has-sub:hover .mi-sub { display:block; }
.mi-sub .mi { display:block; }
.mi-sub .mi.has-sub { display:flex; align-items:center; }
.mi-sub .mi-sub { top:-6px; }
/* dock (element bar) submenu variant */
.eb-menu-item { position:relative; }
.eb-menu-item .sub-arrow { margin-left:auto; }
.eb-sub { display:none; position:absolute; left:calc(100% + 2px); top:-8px; min-width:130px; flex-direction:column; background:var(--panel); border:1px solid var(--border2); border-radius:8px; padding:4px; z-index:210; box-shadow:0 8px 24px rgba(0,0,0,0.5); }
.eb-menu-item.has-sub:hover { background:var(--btn-hover); }
.eb-menu-item.has-sub:hover .eb-sub { display:flex; }
.eb-sub .eb-menu-item { display:flex; }
.mi:hover { background:var(--btn-hover); }
.mi.grp { color:var(--muted); font-size:11px; font-weight:600; padding:6px 10px 2px; cursor:default; }
.mi.grp:hover { background:transparent; }
.mi.on { color:var(--accent); font-weight:600; }
.mi.on::before { content:"\2713 "; }
.dropsep { height:1px; background:var(--border2); margin:4px 0; }

/* snippet menu (fixed overlay anchored to the Snippets button) */
#snip-menu { padding:0; background:var(--panel2); border:1px solid var(--border2); border-radius:8px; box-shadow:0 14px 34px rgba(0,0,0,.45); }
.snip-head { display:flex; align-items:baseline; gap:8px; padding:7px 11px 6px; border-bottom:1px solid var(--border); background:var(--panel3); border-radius:8px 8px 0 0; }
.snip-head .st { font-size:11px; font-weight:700; text-transform:uppercase; letter-spacing:.07em; color:var(--fg); }
.snip-head .grily { margin-left:auto; font-size:10px; color:var(--muted); }
.snip-grp { padding:7px 11px 2px; font-size:9.5px; font-weight:600; text-transform:uppercase; letter-spacing:.06em; color:var(--dim); }
#snip-menu .snip-item { display:flex; align-items:center; gap:9px; padding:5px 11px; font-size:12px; cursor:pointer; color:var(--fg); }
#snip-menu .snip-item:hover, #snip-menu .snip-item.kbd { background:var(--btn-hover); }
#snip-menu .snip-item.kbd { box-shadow:inset 2px 0 0 var(--accent); }
.snip-ic { flex:none; width:26px; height:18px; display:flex; align-items:center; justify-content:center; border-radius:4px; background:var(--btn-bg); border:1px solid var(--border2); color:var(--accent); font-size:9px; font-weight:700; letter-spacing:.03em; font-family:ui-monospace, Consolas, monospace; }
.snip-lb { flex:none; font-weight:600; }
.snip-ds { color:var(--muted); font-size:10.5px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; min-width:0; }
.snip-ftr { display:flex; align-items:center; gap:8px; border-top:1px solid var(--border); background:var(--panel3); padding:5px 11px; border-radius:0 0 8px 8px; color:var(--muted); font-size:10px; }
.snip-ftr .kkey { flex:none; padding:0 5px; border:1px solid var(--border2); border-radius:3px; background:var(--btn-bg); font-family:ui-monospace, Consolas, monospace; font-size:9px; line-height:15px; }
#snip-preview { border-top:1px solid var(--border); background:var(--input-bg); font-family:ui-monospace, SFMono-Regular, Consolas, monospace; font-size:10.5px; line-height:1.5; color:var(--fg); padding:8px 11px; max-height:118px; overflow:auto; white-space:pre-wrap; word-break:break-word; }
#snip-preview:empty { display:none; }

.btn { background:var(--btn-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; padding:5px 11px; font-size:13px; cursor:pointer; }
.btn:hover { background:var(--btn-hover); }
.btn.on { background:var(--btn-on); border-color: var(--accent); }
.sep { width:1px; height:16px; background:var(--border2); margin:0 4px; }
.bp { display:flex; align-items:center; gap:2px; }
#status { font-size:12px; color:var(--muted); margin-left:auto; white-space:pre-wrap; max-width:40%; }

/* docked toolbar */
#topbars { display:flex; flex-direction:row; flex-wrap:wrap; align-items:stretch; }
#toolbar { display:inline-flex; align-items:center; gap:3px; padding:5px 12px; border-bottom:1px solid var(--border); background:var(--panel); max-width:100%; vertical-align:top; }
.tbtn { display:inline-flex; align-items:center; justify-content:center; width:30px; height:30px; padding:0; border:1px solid transparent; border-radius:6px; background:transparent; color:var(--fg); cursor:pointer; }
.tbtn:hover { background:var(--btn-hover); border-color:var(--border2); }
.tbtn:disabled { opacity:.35; cursor:default; }
.tbtn:disabled:hover { background:transparent; border-color:transparent; }
.tbtn svg { width:17px; height:17px; }
.tbtnsep { width:1px; height:18px; background:var(--border2); margin:0 4px; }

/* main 3-column workspace */
main { flex:1; display:flex; min-height:0; }

/* element insert toolbar (layout) */
#elementbar { display:inline-flex; align-items:center; gap:6px; padding:6px 14px; border-bottom:1px solid var(--border); background:var(--panel2); flex-wrap:wrap; max-width:100%; vertical-align:top; }
#elementbar .ebtn { display:inline-flex; align-items:center; justify-content:center; width:36px; height:36px; padding:0; border:1px solid var(--border2); border-radius:8px; background:var(--btn-bg); color:var(--fg); cursor:pointer; transition: all 0.15s ease; }
#elementbar .ebtn:hover { background:var(--btn-hover); border-color:#4c5cf7; transform: translateY(-1px); box-shadow: 0 2px 8px rgba(76, 92, 247, 0.25); }
#elementbar .ebtn:active { transform: translateY(0); }
#elementbar .ebtn svg { width:22px; height:22px; }
#elementbar .esep { width:1px; height:24px; background:var(--border2); margin:0 6px; }
#elementbar .elabel { font-size:12px; color:var(--muted); margin:0 6px 0 4px; font-weight: 500; }
.eb-group { position: relative; display: inline-flex; align-items: center; }
.eb-group .ebtn { width:42px; }
.eb-menu { position: absolute; top: calc(100% + 4px); left: 0; display: none; flex-direction: column; background: var(--panel); border: 1px solid var(--border2); border-radius: 8px; box-shadow: 0 8px 24px rgba(0,0,0,0.5); z-index: 200; min-width: 160px; padding: 6px; }
.eb-menu.open { display: flex; }
.eb-menu-item { display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 6px; border: none; background: transparent; color: var(--fg); cursor: pointer; text-align: left; font-size: 13px; }
.eb-menu-item:hover { background: var(--btn-hover); }
.eb-menu-item svg { width:18px; height:18px; }
/* draggable toolbar grip */
.bar-grip { cursor:grab; color:var(--dim); font-size:15px; line-height:1; padding:0 6px; user-select:none; flex:none; }
.bar-grip:active { cursor:grabbing; }
.bar-grip:hover { color:var(--accent); }
#toolbar.drop-hot, #elementbar.drop-hot { outline:2px dashed var(--accent); outline-offset:-2px; }

/* left panel (structure tree) */
#left { width:248px; min-width:190px; border-right:1px solid var(--border); display:flex; flex-direction:column; background: var(--panel); }
.tabs { display:flex; border-bottom:1px solid var(--border); }
.tab { flex:1; text-align:center; padding:8px 4px; font-size:12px; color:var(--muted); cursor:pointer; border-right:1px solid var(--border); }
.tab.on { color:var(--fg); background:var(--panel3); border-bottom:2px solid var(--accent); }
#leftbody { flex:1; overflow-y:auto; padding:8px 10px; }

/* center */
#center { flex:1; display:flex; flex-direction:column; min-width:0; }
#ctabs { display:flex; border-bottom:1px solid var(--border); background: var(--panel); }
#cbody { flex:1; display:flex; min-height:0; position:relative; }
#cpane-preview, #cpane-sui { position:absolute; inset:0; display:none; }
#cpane-preview.on, #cpane-sui.on { display:flex; flex-direction:column; }
#stage { flex:1; display:flex; justify-content:center; align-items:stretch; overflow:auto; background:#0d0d14; }
#frame { flex:1; position:relative; }
#prev { width:100%; height:100%; border:0; background:#0d0d14; }
#stage.tablet #frame { max-width:768px; margin:0 auto; }
#stage.mobile #frame { max-width:375px; margin:0 auto; }
#stage.custom { align-items:center; }
#stage.custom #frame { flex:none; margin:0 auto; }
#vbar { display:flex; align-items:center; gap:6px; padding:5px 10px; border-bottom:1px solid var(--border); background:var(--panel); flex-wrap:wrap; }
#vbar label { font-size:11px; color:var(--muted); }
#vbar input { width:70px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:5px; padding:4px 6px; font-size:12px; }
#vbar .vsep { width:1px; height:16px; background:var(--border2); }

/* code editors */
.editor {
  display:flex; flex:1; min-height:0;
  font-family: var(--editor-font, ui-monospace, Consolas, Menlo, monospace);
  font-size: var(--editor-size, 16px);
}
.editor.no-line-numbers .gutter { display:none; }
.gutter {
  width:48px; background:var(--gutter-bg); color:var(--gutter-fg);
  text-align:right; padding:10px 8px 10px 0; overflow:hidden;
  user-select:none; white-space:pre;
  line-height: var(--editor-lh, 1.6);
  font-size: inherit;
}
  .codescroll { position:relative; flex:1; min-height:0; overflow:hidden; background:var(--editorbg); }
  #sui_hl .tok-fn { color:#82aaff; }
  .cm-error { text-decoration: wavy underline #f7768e; text-underline-offset: 2px; }
  .cm-warn { text-decoration: wavy underline #e0af68; text-underline-offset: 2px; }
  .sui-ac { position:absolute; background:var(--panel); border:1px solid var(--border); border-radius:6px; box-shadow:0 8px 24px rgba(0,0,0,.4); max-height:220px; overflow:auto; z-index:10000; display:none; min-width:220px; }
  .sui-ac-item { padding:4px 8px; font-size:12px; cursor:pointer; display:flex; gap:6px; align-items:baseline; }
  .sui-ac-item.active, .sui-ac-item:hover { background:var(--btn-hover); }
  .sui-ac-item .ac-desc { color:var(--muted); font-size:11px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
  .ctx-sep { height:1px; background:var(--border); margin:4px 2px; }
  .ac-kind { color:var(--muted); font-size:10px; text-transform:uppercase; letter-spacing:.04em; flex:none; }
  .sui-doc { position:absolute; background:var(--panel); border:1px solid var(--border); border-radius:6px; box-shadow:0 8px 24px rgba(0,0,0,.45); padding:8px 10px; z-index:10001; display:none; max-width:420px; font-size:12px; color:var(--fg); pointer-events:none; }
  .sui-doc .doc-sig { font-family:ui-monospace,monospace; color:#82aaff; font-size:12.5px; }
  .sui-doc .doc-desc { color:var(--muted); margin-top:4px; line-height:1.45; }
  .sui-doc .doc-ret { color:#9ece6a; margin-top:3px; font-size:11px; }
  .doc-kind { display:inline-block; font-size:9px; font-weight:700; text-transform:uppercase; letter-spacing:.05em; padding:1px 6px; border-radius:3px; margin-bottom:5px; }
  .doc-kind.warn { background:#e0af68; color:#14161f; }
  .doc-kind.err { background:#e5484d; color:#fff; }
  .doc-desc.hintline { color:var(--dim); margin-top:2px; }
  .cm-ref { background:rgba(158,206,106,.28); border-radius:2px; }
  .cm-def { background:rgba(255,203,107,.30); border-radius:2px; }
  .cm-elline { background:rgba(76,92,247,.16); }
  .outline { max-height:150px; overflow:auto; border-top:1px solid var(--border); border-bottom:1px solid var(--border); padding:4px 6px; background:var(--panel2); }
  .ol-item { display:flex; gap:6px; align-items:baseline; padding:2px 6px; border-radius:4px; cursor:pointer; font-size:12px; color:var(--fg); white-space:nowrap; }
  .ol-item:hover { background:var(--btn-hover); }
  .ol-kind { font-size:9px; text-transform:uppercase; letter-spacing:.04em; color:var(--muted); flex:none; width:44px; }
  .ol-name { color:var(--fg); }
  .ol-detail { color:var(--muted); font-size:11px; overflow:hidden; text-overflow:ellipsis; }
  .ol-empty { color:var(--dim); font-size:11px; padding:4px 6px; }
.hl, .editor textarea {
  margin:0; padding:10px 12px;
  font-family:inherit; font-size:inherit; font-weight:400; font-style:normal;
  line-height: var(--editor-lh, 1.6);
  letter-spacing:0;
  font-kerning:none;
  font-variant-ligatures:none;
  font-feature-settings:"liga" 0, "calt" 0;
  tab-size: var(--editor-tab, 4);
  -moz-tab-size: var(--editor-tab, 4);
  white-space: var(--editor-wrap, pre);
  overflow-wrap: anywhere;
  word-break: normal;
}
.hl { position:absolute; inset:0; overflow:hidden; color:var(--editorfg); background:transparent; pointer-events:none; }
.editor textarea { position:absolute; inset:0; background:transparent; color:transparent; caret-color:#7db3ff; border:0; outline:0; resize:none; overflow:auto; }
.editor.wrap .hl, .editor.wrap textarea { white-space:pre-wrap; overflow-wrap:anywhere; }
.tok-kw { color:#7aa2f7; }
.tok-str { color:#9ece6a; }
.tok-num { color:#ff9e64; }
.tok-cm { color:#565f89; }
.tok-fnname { color: var(--edtok-fn, #e5c07b); }
.tok-var { color: var(--edtok-var, #56b6c2); }
.editor.plain-code .tok-kw,
.editor.plain-code .tok-str,
.editor.plain-code .tok-num,
.editor.plain-code .tok-cm,
.editor.plain-code .tok-fnname,
.editor.plain-code .tok-var { color: var(--editorfg); }
.tok-tag { color:#7db3ff; }
.tok-attr { color:#ffcb6b; }
.tok-prop { color:#82aaff; }
.tok-sel { color:#ffcb6b; }
.br-match { background:rgba(122,162,247,.30); border-radius:2px; box-shadow:0 0 0 1px rgba(122,162,247,.55); }

/* right inspector */
#right { width:272px; min-width:210px; border-left:1px solid var(--border); display:flex; flex-direction:column; background: var(--panel); position:relative; }
#right-grip { position:absolute; left:-3px; top:0; bottom:0; width:6px; cursor:col-resize; z-index:6; }
#right-grip:hover, #right-grip.drag { background:rgba(76,92,247,.35); }
#inspector { flex:1; overflow-y:auto; padding:10px 12px; }
.states { display:flex; gap:4px; margin-bottom:7px; }
.states .btn { font-size:12px; padding:4px 8px; }
.spec { font-size:12px; color:var(--muted); margin:5px 0; }
.spec b { color:var(--fg); }
.spec .warn { color:#ffb454; }
.acc { border:1px solid var(--border2); border-radius:7px; margin-bottom:7px; }
.acc > .acchead { padding:7px 10px; font-size:12px; cursor:pointer; background:var(--panel3); display:flex; justify-content:space-between; }
.acc > .accbody { padding:7px 10px; display:none; }
.acc.open > .accbody { display:block; }
#inspector .ins-sec { border:1px solid var(--border2); border-radius:7px; margin-bottom:7px; }
#inspector .ins-sec-h { padding:7px 10px; font-size:12px; cursor:pointer; background:var(--panel3); display:flex; align-items:center; justify-content:space-between; }
#inspector .ins-sec-h::after { content:"\25BE"; font-size:9px; color:var(--dim); transition:transform .12s ease; }
#inspector .ins-sec.open .ins-sec-h::after { transform:rotate(180deg); }
#inspector .ins-sec-body { padding:7px 10px; display:none; }
#inspector .ins-sec.open .ins-sec-body { display:block; }
.rowline { display:flex; gap:5px; margin:4px 0; align-items:center; }
.rowline label { font-size:12px; color:var(--muted); width:72px; flex:none; }
.rowline input, .rowline select { flex:1; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:5px; padding:5px 7px; font-size:13px; min-width:0; }
input[type=color] { padding:0; height:26px; }
.rowline input[type=range] { padding:0; }
.hint { font-size:11px; color:var(--dim); margin:5px 0; }
.applyhot { background: var(--accent) !important; color: #fff !important; font-weight: 700; box-shadow: 0 0 0 2px rgba(76, 92, 247, .35); }
/* inspector switch: graphical on/off control for binary attribute fields */
label.isw { box-sizing:border-box; position:relative; width:36px; height:20px; flex:none; display:inline-block; }
label.isw input { position:absolute; inset:0; width:100%; height:100%; opacity:0; margin:0; cursor:pointer; z-index:2; }
.iswTrk { position:absolute; inset:0; background:var(--input-bg); border:1px solid var(--border2); border-radius:10px; transition:background .15s,border-color .15s; }
.iswKnob { position:absolute; top:2px; left:2px; width:14px; height:14px; border-radius:50%; background:var(--muted); transition:left .15s,background .15s; }
.isw input:checked ~ .iswTrk { background:var(--accent); border-color:var(--accent); }
.isw input:checked ~ .iswTrk .iswKnob { left:18px; background:#fff; }
.isw.mixed .iswKnob { left:10px; opacity:.65; }
.histitem { padding:4px 6px; border-radius:4px; cursor:pointer; font-size:12px; color:var(--fg); }
.histitem:hover { background:var(--btn-hover); }

/* tree */
.node { border-radius:4px; padding:4px 6px; margin:1px 0; cursor:pointer; font-size:13px; color:var(--fg); display:flex; align-items:center; gap:4px; }
.node.sel { background:var(--sel); outline:1px solid var(--accent); }
.node.multi { background:var(--btn-hover); }
.node.unused { outline:1px dashed #ff7b72; }
.node.twist {
  cursor:pointer; color:var(--muted); font-size:15px; font-weight:700;
  line-height:1; flex:none; width:16px; height:18px; padding:0;
  justify-content:center; align-items:center; border-radius:3px;
  user-select:none;
}
.node.twist:hover { color:var(--fg); background:var(--btn-hover); }
.node.hot { outline:2px dashed var(--accent); }
.node-type {
  flex:none;
  font:600 9px/1 var(--mono, ui-monospace, Consolas, monospace);
  letter-spacing:.2px;
  text-transform:lowercase;
  color:var(--accent);
  background:var(--btn-hover);
  border:1px solid var(--border2);
  border-radius:4px;
  padding:3px 4px;
  opacity:.92;
}
.node.sel .node-type { color:var(--fg); border-color:var(--accent); }
.sel-type {
  display:flex; align-items:center; gap:6px;
  margin:0 0 8px; padding:6px 8px;
  border:1px solid var(--border2); border-radius:6px;
  background:var(--panel3); color:var(--fg); font-size:12px;
}
.sel-type code {
  font:600 11px/1 var(--mono, ui-monospace, Consolas, monospace);
  color:var(--accent); background:var(--btn-hover);
  border:1px solid var(--border2); border-radius:4px; padding:2px 5px;
}
.rowth { font-size:12px; color:var(--muted); margin:8px 0 3px; text-transform:uppercase; letter-spacing:.4px; }

/* palette */
#palette { display:flex; flex-wrap:wrap; gap:4px; margin-bottom:8px; }
.palbtn { background:var(--palbtn-bg); border:1px solid var(--border2); color:var(--fg); border-radius:5px; padding:5px 9px; font-size:12px; cursor:pointer; }
.palbtn:hover { background:var(--btn-hover); }

/* bottom panel */
#bottom { display:none; height:190px; border-top:1px solid var(--border); background:var(--panel); flex-direction:column; }
/* project color palette — visible color bar + selection popup per field */
.colorbar { display:block; width:100%; height:26px; border:1px solid var(--border2); border-radius:6px; cursor:pointer; padding:0; position:relative; background:#000; }
.colorbar.unset { background: repeating-linear-gradient(45deg, var(--input-bg), var(--input-bg) 6px, var(--panel3) 6px, var(--panel3) 12px); }
.colorbar-lab { position:absolute; left:8px; top:50%; transform:translateY(-50%); font-size:10.5px; font-weight:700; letter-spacing:.04em; pointer-events:none; }
.color-pop { display:none; position:fixed; z-index:96; background:var(--panel); border:1px solid var(--border2); border-radius:9px; box-shadow:0 8px 22px rgba(0,0,0,.4); padding:10px; width:216px; min-height:64px; }
.color-pop.on { display:block; }
.color-pop-ttl { font-size:10px; font-weight:700; text-transform:uppercase; letter-spacing:.06em; color:var(--muted); margin-bottom:5px; }
.color-pop-row { display:flex; align-items:center; gap:6px; margin-bottom:7px; }
.color-pop-row span { font-size:11.5px; color:var(--muted); }
.color-pop-row input[type=color] { width:40px; height:24px; padding:0; border:1px solid var(--border2); border-radius:6px; background:var(--input-bg); cursor:pointer; flex:none; }
.color-hex { flex:1; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; padding:4px 7px; font-size:12px; min-width:0; font-family:ui-monospace, monospace; }
.swstrip { display:flex; flex-wrap:wrap; gap:4px; max-width:100%; }
.swstrip:empty { display:none; }
.sw { width:20px; height:20px; border-radius:4px; border:1px solid var(--border2); cursor:pointer; padding:0; flex:none; }
.sw:hover { outline:2px solid var(--accent); outline-offset:1px; }
#bottom.on { display:flex; }
#btabs { display:flex; border-bottom:1px solid var(--border); }
#bbody { flex:1; overflow-y:auto; padding:8px 12px; font-size:13px; }
#problems li { margin:4px 0; cursor:pointer; }
#problems .unused { color:#ff7b72; }
#problems .override { color:#ffb454; }
#problems .duplicate { color:#7db3ff; }
#problems .syntax { color:#ff7b72; }
#history div { margin:3px 0; cursor:pointer; color:var(--fg); }

/* tokens */
.tok { display:flex; align-items:center; gap:7px; margin:4px 0; }
.sw { width:18px; height:18px; border-radius:4px; border:1px solid var(--border2); }
.tok .tkname { font-size:12px; font-family:ui-monospace, monospace; color:var(--fg); }
.tok .tkval { font-size:12px; color:var(--muted); margin-left:auto; }

/* preferences modal */
#pref { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:10; }
#pref.on { display:flex; }
#prefbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:18px 20px; width:420px; max-width:92vw; max-height:90vh; overflow:auto; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#prefbox h2 { font-size:15px; margin:0 0 12px; }
.pfsec { font-size:11px; font-weight:700; text-transform:uppercase; letter-spacing:.06em; color:var(--muted); margin:16px 0 6px; padding-top:10px; border-top:1px solid var(--border); }
.pfrow input[type=checkbox] { flex:none; }
.pfrow { display:flex; align-items:center; gap:10px; margin:10px 0; }
.pfrow label { font-size:13px; width:120px; flex:none; color:var(--muted); }
.pfrow select, .pfrow input[type=range] { flex:1; font-size:13px; }
.pfbtns { display:flex; justify-content:flex-end; gap:8px; margin-top:14px; }

/* tooltip global settings modal */
#ttcfg { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:10; }
#ttcfg.on { display:flex; }
#ttcfgbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:18px 20px; width:440px; max-width:92vw; max-height:90vh; overflow:auto; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#ttcfgbox h2 { font-size:15px; margin:0 0 12px; }
#ttcfgbox .pfrow input[type=text], #ttcfgbox .pfrow input[type=number] { flex:1; font-size:13px; padding:6px 8px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
#ttcfgbox .pfrow input[type=color] { width:52px; height:28px; padding:2px; background:var(--input-bg); border:1px solid var(--border2); border-radius:6px; }

/* per-chart tooltip mini panel */
#cttcfg { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:10; }
#cttcfg.on { display:flex; }
#cttcfgbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:16px 18px; width:390px; max-width:92vw; max-height:90vh; overflow:auto; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#cttcfgbox h2 { font-size:15px; margin:0 0 8px; }
#cttcfgbox .pfrow input[type=text], #cttcfgbox .pfrow input[type=number] { flex:1; font-size:13px; padding:6px 8px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
#cttcfgbox .pfrow input[type=color] { width:52px; height:28px; padding:2px; background:var(--input-bg); border:1px solid var(--border2); border-radius:6px; }
#ctt-preview { margin:10px 0 2px; padding:6px 8px; background:#111827; color:#f8fafc; border:1px solid #475569; border-radius:6px; max-width:240px; font-size:11px; line-height:1.4; white-space:pre-line; box-sizing:border-box; }

/* per-chart table mini panel */
#ctablecfg { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:10; }
#ctablecfg.on { display:flex; }
#ctablecfgbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:16px 18px; width:390px; max-width:92vw; max-height:90vh; overflow:auto; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#ctablecfgbox h2 { font-size:15px; margin:0 0 8px; }
#ctablecfgbox .pfrow input[type=text], #ctablecfgbox .pfrow input[type=number] { flex:1; font-size:13px; padding:6px 8px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
#ctablecfgbox .pfrow input[type=color] { width:52px; height:28px; padding:2px; background:var(--input-bg); border:1px solid var(--border2); border-radius:6px; }
#ctable-preview { margin:10px 0 2px; border:1px solid var(--border2); border-radius:5px; overflow:hidden; font-size:11px; }
#ctable-preview .hd { padding:4px 6px; background:#14171c; color:#e6e6ef; font-weight:600; }
#ctable-preview .rw { padding:4px 6px; color:var(--fg); border-top:1px solid var(--border); }

/* per-chart state hint mini panel */
#chintscfg { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:11; }
#chintscfg.on { display:flex; }
#chintscfgbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:16px 18px; width:780px; max-width:94vw; max-height:90vh; overflow:auto; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#chintscfgbox h2 { font-size:15px; margin:0 0 8px; }
#chintscfgbox .pfrow input[type=text], #chintscfgbox .pfrow input[type=number] { flex:1; font-size:13px; padding:6px 8px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
#chintscfgbox .pfrow input[type=color] { width:52px; height:28px; padding:2px; background:var(--input-bg); border:1px solid var(--border2); border-radius:6px; }
#chints-groups { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:10px; align-items:start; }
.chints-group { min-width:0; padding:8px 10px 10px; border:1px solid var(--border2); border-radius:10px; background:rgba(148,163,184,.045); }
.chints-group-title { margin:0 0 6px; font-size:10px; font-weight:700; text-transform:uppercase; letter-spacing:.08em; color:var(--muted); }
.chints-group .pfrow { margin:6px 0; gap:4px; }
.chints-group .pfrow label { width:105px; font-size:12px; }
.chints-group .pfrow input[type=range] { flex:1; min-width:0; accent-color:var(--accent); }
.chints-bg-field { display:grid; grid-template-columns:52px minmax(0,1fr) 34px; gap:6px; align-items:center; flex:1; min-width:0; }
.chints-bg-field input[type=range] { width:100%; min-width:0; }
.chints-alpha-label { font-size:11px; text-align:right; color:var(--muted); font-variant-numeric:tabular-nums; }
#chints-preview { display:block; width:max-content; max-width:280px; margin:4px auto 2px; padding:6px; border:1px solid #3d4556; border-radius:6px; background:#1d2129; color:#e6e6e6; font-size:11px; line-height:1.4; white-space:pre-line; text-align:left; }
#chints-preview-label { margin-top:10px; text-align:center; }
@media (max-width: 640px) { #chints-groups { grid-template-columns:1fr; } }

/* per-element state hint editor */
#celemhints { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:12; }
#celemhints.on { display:flex; }
#celemhintsbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:16px 18px; width:720px; max-width:94vw; max-height:92vh; display:flex; flex-direction:column; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#celemhintsbox h2 { font-size:15px; margin:0 0 8px; }
#celemhints-rows { overflow:auto; max-height:55vh; margin:8px 0; border:1px solid var(--border); border-radius:8px; }
.celemhints-row { display:grid; grid-template-columns:170px minmax(0,1fr) minmax(0,1fr); gap:8px; align-items:center; padding:5px 8px; border-bottom:1px solid var(--border); }
.celemhints-row:last-child { border-bottom:0; }
.celemhints-head { background:var(--input-bg); font-size:11px; font-weight:600; color:var(--muted); position:sticky; top:0; z-index:1; }
.celemhints-row .lbl { font-size:11px; color:var(--muted); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.celemhints-input { width:100%; box-sizing:border-box; font-size:12px; padding:5px 7px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
#cextrattrs { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:13; }
#cextrattrs.on { display:flex; }
#cextrattrsbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:16px 18px; width:640px; max-width:94vw; max-height:92vh; display:flex; flex-direction:column; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#cextrattrsbox h2 { font-size:15px; margin:0 0 8px; }
#cextrattrs-rows { overflow:auto; max-height:50vh; margin:8px 0; border:1px solid var(--border); border-radius:8px; }
.cextra-row { display:grid; grid-template-columns:180px minmax(0,1fr) 28px; gap:6px; align-items:center; padding:5px 7px; border-bottom:1px solid var(--border); }
.cextra-row:last-child { border-bottom:0; }
.cextra-row input { width:100%; box-sizing:border-box; font-size:12px; padding:5px 7px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
.cextra-remove { width:26px; height:26px; padding:0; border:1px solid var(--border2); border-radius:6px; background:transparent; color:var(--muted); cursor:pointer; }
.cextra-remove:hover { color:var(--fg); border-color:var(--accent); }

/* unified chart settings panel */
#csettings { position:fixed; inset:0; background:rgba(0,0,0,.5); display:none; align-items:center; justify-content:center; z-index:10; }
#csettings.on { display:flex; }
#csettingsbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:16px 18px; width:460px; max-width:94vw; max-height:90vh; overflow:auto; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#csettingsbox h2 { font-size:15px; margin:0 0 10px; }
.cset-tabs { display:flex; flex-wrap:wrap; gap:4px; border-bottom:1px solid var(--border); margin-bottom:10px; }
.cset-tab { border:0; border-bottom:2px solid transparent; background:transparent; color:var(--muted); padding:6px 8px; font-size:12px; cursor:pointer; }
.cset-tab.on { color:var(--fg); border-bottom-color:var(--accent); }
.cset-pane { display:none; }
.cset-pane.on { display:block; }
.cset-pane .pfrow label { width:128px; }
.cset-pane .pfrow input[type=text], .cset-pane .pfrow input[type=number], .cset-pane .pfrow select { flex:1; font-size:13px; padding:6px 8px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
.cset-pane .pfrow input[type=color] { width:52px; height:28px; padding:2px; background:var(--input-bg); border:1px solid var(--border2); border-radius:6px; }
.cset-pane .pfrow input[type=checkbox] { flex:none; }
.cset-hint { font-size:11px; color:var(--muted); margin:4px 0 8px; }

/* styled prompt modal (rename, new function, ...) — same card style as prefs */
#askdlg { position:fixed; inset:0; background:rgba(0,0,0,.55); display:none; align-items:center; justify-content:center; z-index:30; }
#askdlg.on { display:flex; }
#askbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:18px 20px; width:380px; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#askbox h2 { font-size:15px; margin:0 0 12px; }
#askbox .pfrow input { flex:1; font-size:13px; padding:6px 8px; }
#ask-extra { font-size:11px; color:var(--muted); margin-top:2px; }
#askbox input:focus { outline:1px solid var(--accent); }
/* confirmation / message modal (guard quick templates, show template errors) */
#confdlg { position:fixed; inset:0; background:rgba(0,0,0,.55); display:none; align-items:center; justify-content:center; z-index:34; }
#confdlg.on { display:flex; }
#confbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:18px 20px; width:420px; max-width:92vw; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#confbox h2 { font-size:15px; margin:0 0 10px; }
#conf-text { font-size:13px; color:var(--fg); white-space:pre-wrap; max-height:50vh; overflow:auto; line-height:1.5; }
/* desktop build dialog */
#deskdlg { position:fixed; inset:0; background:rgba(0,0,0,.55); display:none; align-items:center; justify-content:center; z-index:35; }
#deskdlg.on { display:flex; }
#deskbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:18px 20px; width:620px; max-width:94vw; max-height:88vh; display:flex; flex-direction:column; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#deskbox h2 { font-size:15px; margin:0 0 12px; }
#deskbox .pfrow { align-items:flex-start; }
#deskbox .pfrow label { padding-top:5px; }
#deskbox input[type=text], #deskbox textarea { flex:1; font-size:12px; padding:6px 8px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; font-family:ui-monospace,monospace; }
#deskbox textarea { min-height:48px; resize:vertical; }
#desk-env { font-size:11px; color:var(--muted); margin:2px 0 8px; }
#desk-status { font-size:12px; color:var(--muted); margin:4px 0; white-space:pre-wrap; }
#desk-status.err { color:#ff9a9a; }
#desk-status.ok { color:#8fe3a4; }
#desk-log { flex:1; min-height:90px; max-height:220px; overflow:auto; margin:6px 0; padding:8px 10px; background:#11141a; color:#c9d1e8; border:1px solid var(--border); border-radius:6px; font-family:ui-monospace,monospace; font-size:11px; line-height:1.4; white-space:pre-wrap; }
#deskdl { display:none; }
/* new-page dialog */
#pagedlg { position:fixed; inset:0; background:rgba(0,0,0,.55); display:none; align-items:center; justify-content:center; z-index:32; }
#pagedlg.on { display:flex; }
#pagebox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; padding:18px 20px; width:420px; max-width:92vw; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#pagebox h2 { font-size:15px; margin:0 0 12px; }
#pagebox .pfrow input, #pagebox .pfrow select { flex:1; font-size:13px; padding:6px 8px; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:6px; }
#page-tpls { margin:4px 0 2px; display:flex; flex-wrap:wrap; gap:5px; max-height:180px; overflow:auto; padding:4px; border:1px dashed var(--border2); border-radius:8px; }
#page-tpls .palbtn { cursor:pointer; }
#page-tpls .palbtn.on { outline:2px solid var(--accent); font-weight:600; }
.p-hint { font-size:11px; color:var(--muted); width:100%; padding:4px 2px 0; }
.pagechip.on .palbtn { outline:1px solid var(--accent); font-weight:600; }
/* page tabs live inside the preview pane */
#pagetabs { display:flex; flex-wrap:wrap; align-items:center; gap:3px; padding:5px 8px; border-bottom:1px solid var(--border); background:var(--panel); }
#pagetabs:empty { display:none; padding:0; border-bottom:none; }
#pagetabs .pagechip .palbtn { padding:3px 10px; font-size:12px; cursor:pointer; }
#pagetabs .pagechip.on .palbtn { outline:1px solid var(--accent); font-weight:600; color:var(--fg); }

/* filesystem browser modal */
#fsdialog { position:fixed; inset:0; background:rgba(0,0,0,.55); display:none; align-items:center; justify-content:center; z-index:11; }
#media-editor { position:fixed; inset:0; background:rgba(0,0,0,.55); display:none; align-items:center; justify-content:center; z-index:12; }
#media-editor.on { display:flex; }
#media-box { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; width:660px; max-width:94vw; height:60vh; display:flex; flex-direction:column; padding:0; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#media-head { display:flex; align-items:center; gap:6px; padding:12px 14px; border-bottom:1px solid var(--border); font-size:14px; }
#media-head span { flex:1; }
#hist-editor { position:fixed; inset:0; background:rgba(0,0,0,.55); display:none; align-items:center; justify-content:center; z-index:12; }
#hist-editor.on { display:flex; }
#hist-box { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; width:520px; max-width:94vw; max-height:70vh; display:flex; flex-direction:column; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#hist-head { display:flex; align-items:center; gap:6px; padding:12px 14px; border-bottom:1px solid var(--border); font-size:14px; }
#hist-head span { flex:1; }
#hist-list { overflow:auto; padding:8px 12px; }
.histv { display:flex; align-items:center; gap:8px; padding:6px 8px; border-radius:6px; font-size:13px; color:var(--fg); }
.histv:hover { background:var(--btn-hover); }
.histv .histvname { flex:1; cursor:pointer; }
#tpl-list { display:flex; flex-wrap:wrap; gap:4px; }
.tplitem { display:inline-flex; align-items:center; gap:4px; }
.tplitem .btn { font-size:11px; padding:3px 7px; }
#fsdialog.on { display:flex; }
#fsbox { background:var(--panel2); border:1px solid var(--border2); border-radius:10px; width:560px; max-width:94vw; height:72vh; display:flex; flex-direction:column; box-shadow:0 10px 40px rgba(0,0,0,.5); }
#fshead { display:flex; justify-content:space-between; align-items:center; padding:12px 14px; border-bottom:1px solid var(--border); font-size:14px; }
#fsbar { display:flex; gap:6px; padding:8px 12px; border-bottom:1px solid var(--border); }
#fsbar input { flex:1; min-width:0; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:5px; padding:5px 8px; font-size:13px; }
#fslist { flex:1; overflow:auto; padding:6px 10px; }
.fsentry { display:flex; align-items:center; gap:8px; padding:6px 8px; border-radius:6px; cursor:pointer; font-size:13px; color:var(--fg); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.fsentry:hover { background:var(--btn-hover); }
.fsentry.dir { color:var(--accent); }
.fsentry .ficon { flex:none; }
#fsfoot { padding:10px 14px; border-top:1px solid var(--border); display:flex; gap:8px; align-items:center; }
#fsfoot input { flex:1; min-width:0; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:5px; padding:5px 8px; font-size:13px; }

/* States & Functions floating panel */
.sf-panel {
  position: fixed; top: 90px; left: 260px;
  width: 340px; height: 300px;
  min-width: 220px; min-height: 120px;
  max-width: 92vw; max-height: 88vh;
  background: var(--panel2); border: 1px solid var(--border2); border-radius: 10px;
  display: none; flex-direction: column; overflow: hidden;
  box-shadow: 0 10px 40px rgba(0,0,0,.5); z-index: 40;
}
.sf-hdr {
  display: flex; align-items: center; gap: 4px;
  padding: 8px 10px; border-bottom: 1px solid var(--border);
  font-size: 13px; font-weight: 600; cursor: move;
  background: var(--panel); border-radius: 10px 10px 0 0; flex: none;
  user-select: none;
}
.sf-body { flex: 1; overflow: auto; padding: 8px 10px; min-height: 0; }
.sf-panel.hidden { display: none !important; }
.sf-resize {
  position: absolute; right: 0; bottom: 0; width: 16px; height: 16px;
  cursor: nwse-resize; z-index: 2;
  background: linear-gradient(135deg, transparent 0 45%, var(--border2) 45% 55%, transparent 55% 70%, var(--border2) 70% 80%, transparent 80%);
}
/* resizable chart data overlay */
#chart-data-layer { position:fixed; inset:0; background:rgba(0,0,0,.38); display:none; z-index:45; }
#chart-data-layer.on { display:block; }
.cd-panel {
  position:absolute; top:80px; left:80px;
  width:760px; height:540px; min-width:380px; min-height:300px;
  max-width:96vw; max-height:92vh;
  background:var(--panel2); border:1px solid var(--border2); border-radius:10px;
  box-shadow:0 14px 50px rgba(0,0,0,.6);
  display:flex; flex-direction:column; overflow:hidden;
}
.cd-head {
  display:flex; align-items:center; gap:8px; padding:9px 10px 9px 14px;
  border-bottom:1px solid var(--border); background:var(--panel);
  font-size:13px; font-weight:600; cursor:move; user-select:none; flex:none;
}
.cd-head #cd-title { flex:none; }
.cd-head .cd-subtitle { flex:1; min-width:0; color:var(--muted); font-size:11px; font-weight:400; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.cd-head .btn { cursor:pointer; }
.cd-body { flex:1; min-height:0; display:flex; flex-direction:column; gap:8px; padding:10px 12px; overflow:hidden; }
.cd-mode { display:flex; align-items:center; gap:5px; flex-wrap:wrap; flex:none; }
.cd-mode-label { color:var(--muted); font-size:12px; margin-right:3px; }
.cd-mode select { background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:5px; padding:5px 7px; font-size:12px; }
.cd-mode .btn.on { outline:1px solid var(--accent); background:var(--btn-hover); color:var(--fg); }
.cd-literal-wrap { flex:1; min-height:0; display:flex; flex-direction:column; gap:7px; }
.cd-literal-wrap[hidden], .cd-expr-wrap[hidden] { display:none; }
.cd-toolbar { display:flex; align-items:center; gap:6px; flex-wrap:wrap; flex:none; }
.cd-toolbar label { color:var(--muted); font-size:12px; }
.cd-toolbar select { background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:5px; padding:5px 7px; font-size:12px; }
.cd-stats { margin-left:auto; color:var(--muted); font-size:11px; white-space:nowrap; }
#cd-text {
  flex:1; min-height:140px; width:100%; resize:none;
  background:#11141a; color:#d7e3ff; border:1px solid var(--border2); border-radius:7px;
  padding:10px 12px; font-family:ui-monospace, Consolas, Menlo, monospace;
  font-size:13px; line-height:1.5; tab-size:2; white-space:pre; overflow:auto;
}
#cd-text:focus { outline:none; border-color:var(--accent); box-shadow:0 0 0 2px rgba(76,92,247,.25); }
.cd-expr-wrap { display:flex; flex-direction:column; gap:7px; padding:4px 0; }
.cd-expr-wrap label { color:var(--muted); font-size:12px; }
#cd-expr { width:100%; background:var(--input-bg); color:var(--fg); border:1px solid var(--border2); border-radius:7px; padding:8px 10px; font-family:ui-monospace, Consolas, Menlo, monospace; font-size:13px; }
#cd-expr:focus { outline:none; border-color:var(--accent); box-shadow:0 0 0 2px rgba(76,92,247,.25); }
.cd-status { flex:none; min-height:18px; font-size:12px; color:var(--muted); }
.cd-status.ok { color:#8fe3a4; }
.cd-status.err { color:#ff9a9a; }
.cd-status.warn { color:#f7c873; }
.cd-foot { display:flex; align-items:center; gap:8px; padding:9px 12px; border-top:1px solid var(--border); flex:none; }
.cd-foot .hint { flex:1; min-width:0; }
.cd-resize {
  position:absolute; right:0; bottom:0; width:18px; height:18px; cursor:nwse-resize; z-index:3;
  background:linear-gradient(135deg, transparent 0 44%, var(--border2) 44% 54%, transparent 54% 68%, var(--border2) 68% 78%, transparent 78%);
}
</style>
</head>
<body>
<div id="topbar">
  <div class="menu" data-menu="file">
    <span class="mitem">File</span>
    <div class="drop" id="drop-file">
      <div class="mi" onclick="newSui()">New SUI</div>
      <div class="mi" onclick="openFsDialog('open')">Open SUI file…</div>
      <div class="dropsep"></div>
      <div class="mi" onclick="saveSui()">Save</div>
      <div class="mi" onclick="saveSuiAs()">Save As…</div>
      <div class="dropsep"></div>
      <div class="mi" onclick="downloadSui()">Export .sui</div>
      <div class="mi" onclick="exportZip()">Export build (.zip)…</div>
      <div class="mi" onclick="exportSingleHtml()">Export single HTML…</div>
      <div class="mi" onclick="openDesktopBuild()">Build desktop executable…</div>
      <div class="dropsep"></div>
      <div class="mi" onclick="openHistory()">Project history…</div>
    </div>
  </div>
  <div class="menu" data-menu="edit">
    <span class="mitem">Edit</span>
    <div class="drop" id="drop-edit">
      <div class="mi" onclick="undo()">Undo</div>
      <div class="mi" onclick="redo()">Redo</div>
      <div class="dropsep"></div>
      <div class="mi" onclick="duplicateSelected()">Duplicate</div>
      <div class="mi" onclick="deleteSelected()">Delete selected</div>
    </div>
  </div>
  <div class="menu" data-menu="insert">
    <span class="mitem">Insert</span>
    <div class="drop" id="drop-insert"></div>
  </div>
  <div class="menu" data-menu="view">
    <span class="mitem">View</span>
    <div class="drop" id="drop-view">
      <div class="mi grp">Device</div>
      <div class="mi" data-bp-mi="desktop" onclick="setBp('desktop')">Desktop</div>
      <div class="mi" data-bp-mi="tablet" onclick="setBp('tablet')">Tablet</div>
      <div class="mi" data-bp-mi="mobile" onclick="setBp('mobile')">Mobile</div>
    </div>
  </div>
  <div class="menu" data-menu="prefs">
    <span class="mitem">Preferences</span>
    <div class="drop" id="drop-prefs">
      <div class="mi" onclick="openPrefs()">Editor settings…</div>
      <div class="mi" onclick="openTooltipCfg()">Tooltip settings.</div>
      <div class="dropsep"></div>
      <div class="mi" id="probbtn" onclick="toggleBottom()">Problems panel</div>
    </div>
  </div>
  <span id="status"></span>
  <span id="dbg" style="font-size:11px;color:#6b7690;margin-left:8px"></span>
</div>
<div id="topbars">
  <div id="toolbar">
  <span class="bar-grip" draggable="true" data-bar="toolbar" title="Drag to move">&#10495;</span>
  <button class="tbtn" id="tb-undo" title="Undo" onclick="undo()"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7v6h6"/><path d="M21 17a9 9 0 0 0-9-9 9 9 0 0 0-6 2.3L3 13"/></svg></button>
  <button class="tbtn" id="tb-redo" title="Redo" onclick="redo()"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 7v6h-6"/><path d="M3 17a9 9 0 0 1 9-9 9 9 0 0 1 6 2.3L21 13"/></svg></button>
  <span class="tbtnsep"></span>
  <button class="tbtn" id="tb-dup" title="Duplicate" onclick="duplicateSelected()"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg></button>
  <span class="tbtnsep"></span>
  <button class="tbtn" id="tb-preview" title="Build & preview in new window" onclick="openPreviewWindow()"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg></button>
  <button class="tbtn" id="tb-desktop" title="Build desktop executable" onclick="openDesktopBuild()"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="12" rx="1.5"/><line x1="8" y1="20" x2="16" y2="20"/><line x1="12" y1="16" x2="12" y2="20"/></svg></button>
</div>
<div id="elementbar"></div>
</div>
<main>
  <div id="left">
    <div class="tabs">
      <div class="tab on" data-tab="structure" onclick="switchLeft('structure')">Structure</div>
      <div class="tab" data-tab="tokens" onclick="switchLeft('tokens')">Tokens</div>
    </div>
    <div id="leftbody">
      <div id="l-structure">
        <div class="rowth" id="elmcount">Elements</div>
        <div id="nodes"></div>
        <div class="rowth" id="elemhead" style="display:none">Element</div>
        <div id="sec-elem" style="display:none"></div>
        <div class="rowth">Blocks</div>
        <div class="rowline" style="margin-top:4px"><button class="btn" style="font-size:11px" onclick="saveBlockFlow()">Save selection as block</button></div>
        <div id="tpl-list"></div>
      </div>
      <div id="l-tokens" style="display:none">
        <div class="rowth">Design tokens</div>
        <div id="tokens"></div>
        <div class="rowline" style="margin-top:8px"><button class="btn" onclick="addToken()">+ token</button></div>
        <div class="hint">Tokens are custom properties (--name) declared in the CSS. Define colors/spacing here and use var(--name) anywhere.</div>
      </div>
    </div>
  </div>

  <div id="center">
    <div id="ctabs">
      <div class="tab on" data-tab="preview" onclick="switchC('preview')">Preview</div>
      <div class="tab" data-tab="sui" onclick="switchC('sui')">SUI</div>
    </div>
    <div id="cbody">
      <div id="cpane-preview" class="on">
        <div id="pagetabs"></div>
        <div id="vbar">
          <label>Viewport</label>
          <span class="vsep"></span>
          <label>W</label><input id="vpw" type="number" min="100" placeholder="auto" oninput="onViewportInput()">
          <label>H</label><input id="vph" type="number" min="100" placeholder="auto" oninput="onViewportInput()">
          <span class="vsep"></span>
          <button class="btn" style="font-size:11px;padding:3px 8px" onclick="setBp('desktop')">Desktop</button>
          <button class="btn" style="font-size:11px;padding:3px 8px" onclick="setBp('tablet')">Tablet</button>
          <button class="btn" style="font-size:11px;padding:3px 8px" onclick="setBp('mobile')">Mobile</button>
          <span class="vsep"></span>
          <button class="btn" id="marquee-btn" title="Marquee select: drag on the preview to select multiple elements" style="font-size:11px;padding:3px 8px" onclick="toggleMarquee()">&#9618; Marquee</button>
          <span class="vsep"></span>
          <button class="btn" id="guides-btn" style="font-size:11px;padding:3px 8px" onclick="toggleGuides()">Guides</button>
          <input id="grid-step" type="number" min="2" max="100" value="8" style="width:44px" title="snap step (px)" onchange="updateGuides()">
        </div>
        <div id="stage" class="desktop">
          <div id="frame"><iframe id="prev" src="/__gui__/preview?ts=0"></iframe></div>
        </div>
      </div>
      <div id="cpane-sui">
        <div style="overflow:auto;padding:8px 10px">
          <div style="margin-top:4px;display:flex;gap:6px;align-items:center;flex-wrap:wrap">
            <button class="btn" style="font-size:11px" onclick="toggleStateFuncPanel()">States &amp; Functions</button>
            <button class="btn" style="font-size:11px" id="snipbtn" onclick="toggleSnippetMenu(event)">Snippets &#9662;</button>
            <button class="btn" style="font-size:11px" onclick="gotoDefinition()" title="F12">Go to def</button>
            <button class="btn" style="font-size:11px" onclick="findReferences()" title="Shift+F12">References</button>
            <button class="btn" style="font-size:11px" onclick="renameSymbol()" title="F2">Rename</button>
            <button class="btn" style="font-size:11px" onclick="toggleComment()" title="Ctrl+/">Comment</button>
            <button class="btn" style="font-size:11px" onclick="toggleOutlinePanel()" title="File structure">Outline</button>
          </div>
        </div>
        <div id="outline" class="outline" style="display:none"></div>
        <div class="rowth" style="padding:0 10px;display:flex;align-items:center;justify-content:space-between">
          <span style="margin:0">SUI source</span>
          <button class="btn" data-apply="sui" onclick="applySui()" style="font-size:12px;display:none">Apply</button>
        </div>
        <div class="editor">
          <div class="gutter" id="suigutter"></div>
          <div class="codescroll"><pre class="hl" id="sui_hl"></pre><textarea id="out" spellcheck="false" oninput="onSuiEdit()"></textarea></div>
        </div>
      </div>
    </div>
  </div>

  <div id="right">
    <div id="right-grip" title="Drag to resize the inspector"></div>
    <div id="inspector"></div>
  </div>
    <div id="sf-panel" class="sf-panel">
      <div class="sf-hdr" id="sf-hdr">
        <span>States &amp; Functions</span>
        <div style="flex:1"></div>
        <button class="btn" style="font-size:11px;padding:2px 6px" onclick="addState()">+ state</button>
        <button class="btn" style="font-size:11px;padding:2px 6px;margin-left:4px" onclick="addFn()">+ fn</button>
        <button class="btn" style="font-size:11px;padding:2px 6px;margin-left:8px" onclick="toggleStateFuncPanel()">×</button>
      </div>
      <div class="sf-body">
        <div id="stateslist2"></div>
        <div id="fnslist2" style="margin-top:8px"></div>
      </div>
      <div class="sf-resize" id="sf-resize" title="Drag to resize"></div>
    </div>
</main>
<div id="bottom">
  <div id="btabs">
    <div class="tab on" data-tab="problems" onclick="switchB('problems')">Problems</div>
    <div class="tab" data-tab="history" onclick="switchB('history')">History</div>
  </div>
  <div id="bbody">
    <ul id="problems" style="margin:0;padding-left:16px"></ul>
    <div id="history" style="display:none"></div>
  </div>
</div>
<input type="file" id="filein" accept=".sui,text/plain" style="display:none">
<input type="file" id="fontfile" accept=".ttf,.otf,.woff,.woff2" style="display:none">
<div id="fsdialog">
  <div id="fsbox">
    <div id="fshead"><span id="fstitle">Open project</span><button class="btn" onclick="closeFsDialog()">&#10005;</button></div>
    <div id="fsbar">
      <button class="btn" title="Up" onclick="fsUp()">&#8593;</button>
      <input id="fspath" placeholder="path" onkeydown="if(event.key==='Enter')fsGo()">
      <button class="btn" onclick="fsGo()">Go</button>
    </div>
    <div id="fslist"></div>
    <div id="fsfoot"></div>
  </div>
</div>
<div id="media-editor">
  <div id="media-box">
    <div id="media-head"><span>Media queries</span><button class="btn" onclick="commitMediaEditor()">Apply</button><button class="btn" onclick="closeMediaEditor()">&#10005;</button></div>
    <textarea id="media-css" spellcheck="false" style="flex:1;width:100%;resize:none;background:var(--input-bg);color:var(--fg);border:1px solid var(--border2);border-radius:6px;padding:10px;font-family:ui-monospace,monospace;font-size:13px"></textarea>
  </div>
</div>
<div id="hist-editor">
  <div id="hist-box">
    <div id="hist-head"><span>Project history</span><button class="btn" onclick="snapshotNow()">Snapshot now</button><button class="btn" onclick="closeHistory()">&#10005;</button></div>
    <div id="hist-list"></div>
  </div>
</div>
<div id="chart-data-layer" aria-hidden="true">
  <div id="cd-panel" class="cd-panel" role="dialog" aria-modal="true" aria-labelledby="cd-title">
    <div id="cd-head" class="cd-head">
      <span id="cd-title">Chart data</span>
      <span id="cd-subtitle" class="cd-subtitle"></span>
      <button class="btn" id="cd-close" title="Close without applying">&#10005;</button>
    </div>
    <div class="cd-body">
      <div class="cd-mode">
        <span class="cd-mode-label">Source</span>
        <button class="btn on" id="cd-mode-literal" type="button">Literal data</button>
        <button class="btn" id="cd-mode-expr" type="button">State expression</button>
        <span class="cd-mode-label" style="margin-left:10px">Type</span>
        <select id="cd-type"></select>
      </div>
      <div id="cd-literal-wrap" class="cd-literal-wrap">
        <div class="cd-toolbar">
          <label for="cd-format">Format</label>
          <select id="cd-format">
            <option value="auto">Auto detect</option>
            <option value="json">JSON</option>
            <option value="xml">XML</option>
            <option value="csv">CSV</option>
          </select>
          <button class="btn" id="cd-load" type="button" title="Read a .json, .xml or .csv file into the editor">Load file.</button>
          <button class="btn" id="cd-sample" type="button" title="Insert a sample matching the selected format">Insert sample</button>
          <button class="btn" id="cd-pretty" type="button" title="Pretty-print valid JSON">Pretty JSON</button>
          <span class="cd-stats" id="cd-stats"></span>
        </div>
        <textarea id="cd-text" spellcheck="false" autocapitalize="off" autocomplete="off" placeholder="Paste JSON, XML or CSV here."></textarea>
      </div>
      <div id="cd-expr-wrap" class="cd-expr-wrap" hidden>
        <label for="cd-expr">State or expression</label>
        <input id="cd-expr" spellcheck="false" autocomplete="off" placeholder="e.g. chartData">
        <div class="hint">The generated app evaluates this expression and redraws the chart when the state changes. The data content is not evaluated inside the designer.</div>
      </div>
      <div class="cd-status" id="cd-status" role="status" aria-live="polite"></div>
    </div>
    <div class="cd-foot">
      <span class="hint">Drag the header to move this panel; resize from the bottom-right corner. Ctrl+Enter applies.</span>
      <button class="btn" id="cd-cancel" type="button">Cancel</button>
      <button class="btn" id="cd-apply" type="button">Apply</button>
    </div>
    <div class="cd-resize" id="cd-resize" title="Drag to resize"></div>
    <input type="file" id="cd-fileinput" accept=".json,.xml,.csv,.txt,application/json,text/xml,text/csv,text/plain" hidden>
  </div>
</div>
<div id="pref">
  <div id="prefbox">
    <h2>Preferences</h2>
    <div class="pfsec">Interface</div>
    <div class="pfrow"><label>Theme</label><select id="pf-theme"><option value="dark">Dark</option><option value="light">Light</option><option value="contrast">High contrast</option></select></div>
    <div class="pfrow"><label>Font size</label><input type="range" id="pf-size" min="12" max="20" value="14"></div>
    <div class="pfrow"><label>Density</label><select id="pf-density"><option value="comfortable">Comfortable</option><option value="compact">Compact</option></select></div>
    <div class="pfrow"><label>Accent color</label><input type="color" id="pf-accent" value="#4c5cf7"></div>
    <div class="pfrow"><label>Toolbars</label>
      <span style="display:flex;flex-direction:column;gap:4px;font-size:12px">
        <label style="display:flex;align-items:center;gap:6px"><input type="checkbox" id="pf-left"> Structure / Tokens</label>
        <label style="display:flex;align-items:center;gap:6px"><input type="checkbox" id="pf-right"> Inspector</label>
        <label style="display:flex;align-items:center;gap:6px"><input type="checkbox" id="pf-bottom"> Problems / History</label>
      </span>
    </div>
    <div class="pfsec">General editor</div>
    <div class="pfrow"><label>Font family</label>
      <select id="pf-editor-font">
        <option value="ui-monospace, Consolas, Menlo, monospace">System monospace</option>
        <option value="'Cascadia Code', 'Cascadia Mono', Consolas, monospace">Cascadia Code</option>
        <option value="'Fira Code', Consolas, monospace">Fira Code</option>
        <option value="'JetBrains Mono', Consolas, monospace">JetBrains Mono</option>
        <option value="Consolas, 'Courier New', monospace">Consolas</option>
        <option value="'Courier New', Courier, monospace">Courier New</option>
        <option value="monospace">Browser monospace</option>
        <option value="__custom__">Custom…</option>
      </select>
    </div>
    <div class="pfrow"><label>Custom font</label><input type="text" id="pf-editor-font-custom" placeholder="'My Mono', monospace"></div>
    <div class="pfrow"><label>Editor size</label><input type="range" id="pf-editor-size" min="12" max="28" value="16"></div>
    <div class="pfrow"><label>Line height</label><input type="range" id="pf-editor-lh" min="1.2" max="2.1" step="0.05" value="1.6"></div>
    <div class="pfrow"><label>Tab size</label><select id="pf-editor-tab"><option value="2">2</option><option value="4">4</option><option value="8">8</option></select></div>
    <div class="pfrow"><label>Word wrap</label><input type="checkbox" id="pf-editor-wrap"></div>
    <div class="pfrow"><label>Line numbers</label><input type="checkbox" id="pf-editor-lines" checked></div>
    <div class="pfsec">SUI code appearance</div>
    <div class="pfrow"><label>Syntax colors</label><input type="checkbox" id="pf-editor-syntax" checked></div>
    <div class="pfrow"><label>Bracket match</label><input type="checkbox" id="pf-editor-brackets" checked></div>
    <div class="pfrow"><label>Selected line</label><input type="checkbox" id="pf-editor-linehl" checked></div>
    <div class="pfbtns">
      <button class="btn" onclick="resetPrefs()">Reset defaults</button>
      <span style="flex:1"></span>
      <button class="btn" onclick="closePrefs()">Cancel</button>
      <button class="btn" onclick="applyPrefs()">Apply</button>
    </div>
  </div>
</div>
<div id="ttcfg">
  <div id="ttcfgbox">
    <h2>Tooltip settings</h2>
    <div class="hint">Global defaults for every element with a tooltip property.</div>
    <div class="pfrow"><label>Place</label><select id="tt-place">
      <option value="top">top</option><option value="right">right</option><option value="bottom">bottom</option><option value="left">left</option>
    </select></div>
    <div class="pfrow"><label>Delay (ms)</label><input type="number" id="tt-delay" min="0" max="5000" step="20" value="140"></div>
    <div class="pfrow"><label>Background</label><input type="color" id="tt-bg" value="#1d2129"></div>
    <div class="pfrow"><label>Text color</label><input type="color" id="tt-fg" value="#e6e6ef"></div>
    <div class="pfrow"><label>Border color</label><input type="color" id="tt-border" value="#3a4152"></div>
    <div class="pfrow"><label>Radius</label><input type="text" id="tt-radius" value="6px" placeholder="6px"></div>
    <div class="pfrow"><label>Max width</label><input type="text" id="tt-maxwidth" value="260px" placeholder="260px"></div>
    <div class="pfbtns">
      <button class="btn" onclick="closeTooltipCfg()">Cancel</button>
      <button class="btn" onclick="applyTooltipCfg()">Apply</button>
    </div>
  </div>
</div>
<div id="cttcfg">
  <div id="cttcfgbox">
    <h2>Chart tooltip</h2>
    <div class="hint">Appearance for the selected chart only. Enable <b>interactive</b> to show tooltips.</div>
    <div id="ctt-preview">Jan
value: 12</div>
    <div class="pfrow"><label>Background</label><input type="color" id="ctt-bg" value="#111827" oninput="updateChartTooltipPreview()"></div>
    <div class="pfrow"><label>Text</label><input type="color" id="ctt-fg" value="#f8fafc" oninput="updateChartTooltipPreview()"></div>
    <div class="pfrow"><label>Border</label><input type="color" id="ctt-border" value="#475569" oninput="updateChartTooltipPreview()"></div>
    <div class="pfrow"><label>Radius</label><input type="text" id="ctt-radius" value="6px" placeholder="6px" oninput="updateChartTooltipPreview()"></div>
    <div class="pfrow"><label>Max width</label><input type="text" id="ctt-maxwidth" value="240px" placeholder="240px" oninput="updateChartTooltipPreview()"></div>
    <div class="pfrow"><label>Font size</label><input type="text" id="ctt-fontsize" value="11px" placeholder="11px" oninput="updateChartTooltipPreview()"></div>
    <div class="pfrow"><label>Delay</label><input type="number" id="ctt-delay" min="0" max="5000" step="20" value="0" oninput="updateChartTooltipPreview()"></div>
    <div class="pfbtns">
      <button class="btn" onclick="resetChartTooltipCfg()">Reset</button>
      <span style="flex:1"></span>
      <button class="btn" onclick="closeChartTooltipCfg()">Cancel</button>
      <button class="btn" onclick="applyChartTooltipCfg()">Apply</button>
    </div>
  </div>
</div>
<div id="ctablecfg">
  <div id="ctablecfgbox">
    <h2>Chart table</h2>
    <div class="hint">Table options for the selected chart. Gantt cells are editable when <b>interactive</b> is on.</div>
    <div id="ctable-preview"><div class="hd">Task &nbsp; Start &nbsp; End &nbsp; Progress</div><div class="rw">Design &nbsp; 2026-11-02 &nbsp; 2026-11-06 &nbsp; 80%</div></div>
    <div class="pfrow"><label>Font size</label><input type="text" id="ctb-fontsize" value="11px" placeholder="11px" oninput="updateChartTablePreview()"></div>
    <div class="pfrow"><label>Cell padding</label><input type="text" id="ctb-padding" value="4px 6px" placeholder="4px 6px" oninput="updateChartTablePreview()"></div>
    <div class="pfrow"><label>Header background</label><input type="color" id="ctb-headbg" value="#14171c" oninput="updateChartTablePreview()"></div>
    <div class="pfrow"><label>Stripe opacity</label><input type="number" id="ctb-stripe" min="0" max="1" step="0.01" value="0.06" oninput="updateChartTablePreview()"></div>
    <div class="pfrow"><label>Show caption</label><input type="checkbox" id="ctb-caption" checked onchange="updateChartTablePreview()"></div>
    <div class="pfrow"><label>Sticky header</label><input type="checkbox" id="ctb-sticky" checked onchange="updateChartTablePreview()"></div>
    <div class="pfbtns">
      <button class="btn" onclick="resetChartTableCfg()">Reset</button>
      <span style="flex:1"></span>
      <button class="btn" onclick="closeChartTableCfg()">Cancel</button>
      <button class="btn" onclick="applyChartTableCfg()">Apply</button>
    </div>
  </div>
</div>
<div id="chintscfg">
  <div id="chintscfgbox">
    <h2>HINTS settings</h2>
    <div class="hint">Global tooltip style for the selected state chart. Values left empty use the theme defaults. In tag and hint text <b>|</b> breaks a line and <b>\\|</b> is a literal pipe.</div>
    <div id="chints-groups">
      <div class="chints-group">
        <div class="chints-group-title">Hints and colors</div>
        <div class="pfrow"><label>Show hints</label><input type="checkbox" id="chints-enabled" checked></div>
        <div class="pfrow"><label>Background</label><span class="chints-bg-field">
          <input type="color" id="chints-bg" value="#1d2129" oninput="updateChartHintsPreview()">
          <input type="range" id="chints-opacity" min="0" max="1" step="0.05" value="1" oninput="updateChartHintsPreview()">
          <span id="chints-opacity-label" class="chints-alpha-label">100%</span>
        </span></div>
        <div class="pfrow"><label>Text color</label><input type="text" id="chints-text" placeholder="#rrggbb" oninput="updateChartHintsPreview()"></div>
        <div class="pfrow"><label>Border color</label><input type="text" id="chints-border" placeholder="#rrggbb" oninput="updateChartHintsPreview()"></div>
        <div class="pfrow"><label>Font</label><input type="text" id="chints-font" placeholder="system-ui, sans-serif" oninput="updateChartHintsPreview()"></div>
      </div>
      <div class="chints-group">
        <div class="chints-group-title">Size and spacing</div>
        <div class="pfrow"><label>Font size</label><input type="number" id="chints-fontsize" min="8" max="32" step="1" value="11" oninput="updateChartHintsPreview()"></div>
        <div class="pfrow"><label>Border width</label><input type="number" id="chints-borderwidth" min="0" max="6" step="0.5" value="1" oninput="updateChartHintsPreview()"></div>
        <div class="pfrow"><label>Text padding</label><input type="number" id="chints-padding" min="0" max="24" step="1" value="6" oninput="updateChartHintsPreview()"></div>
        <div class="pfrow"><label>Corner radius</label><input type="number" id="chints-radius" min="0" max="32" step="1" value="6" oninput="updateChartHintsPreview()"></div>
      </div>
    </div>
    <div id="chints-preview-label" class="cset-hint">Preview</div>
    <div id="chints-preview">Process
RUNNING</div>
    <div class="pfbtns">
      <button class="btn" onclick="resetChartHintsCfg()">Reset</button>
      <span style="flex:1"></span>
      <button class="btn" onclick="closeChartHintsCfg()">Cancel</button>
      <button class="btn" onclick="applyChartHintsCfg()">Apply</button>
    </div>
  </div>
</div>
<div id="celemhints">
  <div id="celemhintsbox">
    <h2>Element hints</h2>
    <div class="hint">Edit tag and hint text for every state and link in JSON, CSV or XML chart data. Use <b>|</b> for a line break and <b>\\|</b> for a literal pipe. The data is rewritten when you apply; CSV needs a header row with id/label or from/to plus tag/hint.</div>
    <div id="celemhints-msg" class="hint"></div>
    <div id="celemhints-rows"></div>
    <div class="pfbtns">
      <span style="flex:1"></span>
      <button class="btn" onclick="closeChartElementHints()">Cancel</button>
      <button class="btn" onclick="applyChartElementHints()">Apply</button>
    </div>
  </div>
</div>
<div id="cextrattrs">
  <div id="cextrattrsbox">
    <h2>Chart attributes</h2>
    <div class="hint">Attributes not exposed by the inspector. Values use SUI source syntax: strings in quotes, numbers and booleans bare. Example: <b>"orthogonal"</b>, <b>24</b>, <b>true</b>.</div>
    <div id="cextrattrs-msg" class="hint"></div>
    <div id="cextrattrs-rows"></div>
    <div class="pfbtns">
      <button class="btn" onclick="addChartExtraRow()">Add attribute</button>
      <span style="flex:1"></span>
      <button class="btn" onclick="closeChartExtraCfg()">Cancel</button>
      <button class="btn" onclick="applyChartExtraCfg()">Apply</button>
    </div>
  </div>
</div>
<div id="csettings">
  <div id="csettingsbox">
    <h2 id="csettings-title">Chart settings</h2>
    <div class="cset-tabs">
      <button class="cset-tab on" data-cset="appearance" onclick="showChartSettingsTab('appearance')">Appearance</button>
      <button class="cset-tab" data-cset="points" onclick="showChartSettingsTab('points')">Points</button>
      <button class="cset-tab" data-cset="tooltip" onclick="showChartSettingsTab('tooltip')">Tooltip</button>
      <button class="cset-tab" data-cset="table" onclick="showChartSettingsTab('table')">Table</button>
      <button class="cset-tab" data-cset="interactions" onclick="showChartSettingsTab('interactions')">Interactions</button>
      <button class="cset-tab" data-cset="data" onclick="showChartSettingsTab('data')">Data</button>
    </div>
    <div class="cset-pane on" data-pane="appearance">
      <div class="cset-hint">Presets write chart colors into style expr. Tooltip and table variables are preserved.</div>
      <div class="pfrow"><label>Theme</label><select id="cs-theme"><option value="dark">Dark</option><option value="light">Light</option></select></div>
      <div class="pfrow"><label>Editable colors</label><input type="checkbox" id="cs-colorpicker" checked></div>
      <div class="pfrow"><label>Style preset</label><select id="cs-preset"></select></div>
    </div>
    <div class="cset-pane" data-pane="points">
      <div class="cset-hint">Point size applies to line, step, smooth, area, scatter, combo, radar and dumbbell.</div>
      <div class="pfrow"><label>Point size</label><input type="text" id="cs-pointsize" value="3" placeholder="3"></div>
    </div>
    <div class="cset-pane" data-pane="tooltip">
      <div class="cset-hint">The tooltip is shown when interactive is on.</div>
      <div class="pfrow"><label>Background</label><input type="color" id="cs-tt-bg" value="#111827"></div>
      <div class="pfrow"><label>Text</label><input type="color" id="cs-tt-fg" value="#f8fafc"></div>
      <div class="pfrow"><label>Border</label><input type="color" id="cs-tt-border" value="#475569"></div>
      <div class="pfrow"><label>Radius</label><input type="text" id="cs-tt-radius" value="6px"></div>
      <div class="pfrow"><label>Max width</label><input type="text" id="cs-tt-maxwidth" value="240px"></div>
      <div class="pfrow"><label>Font size</label><input type="text" id="cs-tt-fontsize" value="11px"></div>
      <div class="pfrow"><label>Delay (ms)</label><input type="number" id="cs-tt-delay" min="0" max="5000" step="20" value="0"></div>
    </div>
    <div class="cset-pane" data-pane="table">
      <div class="cset-hint">The accessible table is available for every chart type. Gantt cells are editable when interactive is on.</div>
      <div class="pfrow"><label>Table view</label><select id="cs-tableview"></select></div>
      <div class="pfrow"><label>Column filters</label><input type="checkbox" id="cs-tb-filters" checked></div>
      <div class="pfrow"><label>Add row button</label><input type="checkbox" id="cs-tb-addrow"></div>
      <div class="pfrow"><label>Delete row button</label><input type="checkbox" id="cs-tb-deleterow"></div>
      <div class="pfrow"><label>Number steppers</label><input type="checkbox" id="cs-tb-steppers"></div>
      <div class="pfrow"><label>Context menu</label><input type="checkbox" id="cs-tb-contextmenu"></div>
      <div class="pfrow"><label>Selection</label><select id="cs-tb-selection"><option value="none">None</option><option value="row">Rows</option><option value="column">Columns</option><option value="cell">Cells</option></select></div>
      <div class="pfrow"><label>Font size</label><input type="text" id="cs-tb-fontsize" value="11px"></div>
      <div class="pfrow"><label>Cell padding</label><input type="text" id="cs-tb-padding" value="4px 6px"></div>
      <div class="pfrow"><label>Header background</label><input type="color" id="cs-tb-headbg" value="#14171c"></div>
      <div class="pfrow"><label>Stripe opacity</label><input type="number" id="cs-tb-stripe" min="0" max="1" step="0.01" value="0.06"></div>
      <div class="pfrow"><label>Show caption</label><input type="checkbox" id="cs-tb-caption" checked></div>
      <div class="pfrow"><label>Sticky header</label><input type="checkbox" id="cs-tb-sticky" checked></div>
    </div>
    <div class="cset-pane" data-pane="interactions">
      <div class="cset-hint">Interactive enables tooltips, selection, table editing and Gantt drag/resize. Keyboard: focus the chart and use arrow keys; Enter selects.</div>
      <div class="pfrow"><label>Interactive</label><input type="checkbox" id="cs-interactive"></div>
      <div class="pfrow"><label>Edit mode</label><input type="checkbox" id="cs-state-edit" checked></div>
      <div class="pfrow"><label>Hover hints</label><input type="checkbox" id="cs-state-hints" checked></div>
      <div class="pfrow"><label>Hint style</label><button class="btn" onclick="openChartHintsCfg()">HINTS...</button></div>
      <div class="pfrow"><label>Show resources</label><input type="checkbox" id="cs-gantt-resources"></div>
      <div class="pfrow"><label>Show progress %</label><input type="checkbox" id="cs-gantt-progress"></div>
      <div class="pfrow"><label>Progress % column</label><input type="checkbox" id="cs-gantt-progress-column"></div>
      <div class="pfrow"><label>Show slack</label><input type="checkbox" id="cs-gantt-slack"></div>
      <div class="pfrow"><label>Show costs</label><input type="checkbox" id="cs-gantt-costs"></div>
      <div class="pfrow"><label>Status colors</label><input type="checkbox" id="cs-gantt-statuscolors"></div>
      <div class="pfrow"><label>Show non-working days</label><input type="checkbox" id="cs-gantt-shownonworking" checked></div>
      <div class="pfrow"><label>Auto schedule</label><input type="checkbox" id="cs-gantt-autoschedule"></div>
      <div class="pfrow"><label>Time scale</label><select id="cs-gantt-scale"><option value="auto">Auto</option><option value="day">Days</option><option value="week">Weeks</option><option value="month">Months</option><option value="year">Years</option></select></div>
      <div class="cset-hint">Link appearance</div>
      <div class="pfrow"><label>Link color</label><input type="color" id="cs-gantt-linkcolor" value="#94a3b8"></div>
      <div class="pfrow"><label>Link width</label><input type="number" id="cs-gantt-linkwidth" min="0.5" max="8" step="0.1" value="1"></div>
      <div class="pfrow"><label>Min link gap (px)</label><input type="number" id="cs-gantt-linkgap" min="0" max="24" step="1" value="4"></div>
      <div class="pfrow"><label>Line style</label><select id="cs-gantt-linkstyle"><option value="auto">Auto (by type)</option><option value="solid">Solid</option><option value="dashed">Dashed</option><option value="dotted">Dotted</option></select></div>
      <div class="pfrow"><label>Arrow</label><select id="cs-gantt-arrowtype"><option value="triangle">Triangle</option><option value="open">Open</option><option value="none">None</option></select></div>
      <div class="pfrow"><label>Arrow size</label><input type="number" id="cs-gantt-arrowsize" min="2" max="20" step="1" value="5"></div>
      <div class="pfrow"><label>Bar border</label><input type="checkbox" id="cs-gantt-barborder-enable"></div>
      <div class="pfrow"><label>Bar border color</label><input type="color" id="cs-gantt-barborder" value="#475569"></div>
      <div class="pfrow"><label>Critical path</label><input type="checkbox" id="cs-critical"></div>
      <div class="pfrow"><label>Holidays</label><input type="text" id="cs-holidays" placeholder="2026-12-25, *-01-01, 2026-12-24..2026-12-26"></div>
      <div class="pfrow"><label>Weekend</label><input type="text" id="cs-weekend" placeholder="sat,sun or none"></div>
      <div class="pfrow"><label>Timezone</label><input type="text" id="cs-timezone" placeholder="UTC or Europe/Rome"></div>
    </div>
    <div class="cset-pane" data-pane="data">
      <div class="cset-hint">The data button in the inspector opens the full JSON/XML/CSV editor.</div>
      <div class="pfrow"><label>X field</label><input type="text" id="cs-x" placeholder="auto"></div>
      <div class="pfrow"><label>Y field</label><input type="text" id="cs-y" placeholder="auto"></div>
      <div class="pfbtns"><button class="btn" onclick="openDataFromChartSettings()">Open data editor...</button></div>
    </div>
    <div class="pfbtns">
      <button class="btn" onclick="resetChartSettings()">Reset styles</button>
      <span style="flex:1"></span>
      <button class="btn" onclick="closeChartSettings()">Cancel</button>
      <button class="btn" onclick="applyChartSettings()">Apply</button>
    </div>
  </div>
</div>
<div id="askdlg">
  <div id="askbox">
    <h2 id="ask-title">Prompt</h2>
    <div class="pfrow"><label id="ask-label">New name</label><input id="ask-input" autocomplete="off" spellcheck="false"></div>
    <div id="ask-extra"></div>
    <div class="pfbtns">
      <button class="btn" id="ask-cancel">Cancel</button>
      <button class="btn" id="ask-ok">OK</button>
    </div>
  </div>
</div>
<div id="confdlg">
  <div id="confbox">
    <h2 id="conf-title">Confirm</h2>
    <div id="conf-text"></div>
    <div class="pfbtns">
      <button class="btn" id="conf-cancel">Cancel</button>
      <button class="btn" id="conf-ok">OK</button>
    </div>
  </div>
</div>
<div id="pagedlg">
  <div id="pagebox">
    <h2>New page</h2>
    <div class="pfrow"><label>Page name</label><input id="page-name" autocomplete="off" spellcheck="false"></div>
    <div class="pfrow"><label>Content</label>
      <select id="page-mode" onchange="pageModeChanged()">
        <option value="empty">Empty</option>
        <option value="template">From template</option>
      </select>
    </div>
    <div id="page-tpls" style="display:none"></div>
    <div class="pfbtns">
      <button class="btn" id="page-cancel">Cancel</button>
      <button class="btn" id="page-ok">OK</button>
    </div>
  </div>
</div>
<div id="deskdlg">
  <div id="deskbox">
    <h2>Build desktop executable</h2>
    <div class="pfrow"><label>Output directory</label><input type="text" id="desk-out" value="dist/desktop" spellcheck="false"></div>
    <div class="pfrow"><label>Allowed dirs</label><textarea id="desk-dirs" placeholder="One directory per line; leave empty for no filesystem access." spellcheck="false"></textarea></div>
    <div class="pfrow"><label>Allowed commands</label><textarea id="desk-exec" placeholder="One command name or path per line; leave empty for no exec access." spellcheck="false"></textarea></div>
    <div class="pfrow"><label>Download</label><label style="width:auto;display:flex;align-items:center;gap:6px;color:var(--fg)"><input type="checkbox" id="desk-download"> download the executable after a successful build</label></div>
    <div id="desk-env"></div>
    <div id="desk-status">Checking build prerequisites…</div>
    <pre id="desk-log"></pre>
    <div class="pfbtns">
      <a id="desk-dl" class="btn" href="#" download>Download executable</a>
      <span style="flex:1"></span>
      <button class="btn" id="desk-cancel">Cancel</button>
      <button class="btn" id="desk-build">Build</button>
    </div>
  </div>
</div>
<script>
"use strict";
// Designer resilience: server restarts re-bind one stable port and an open
// tab recovers by itself. The first unrecoverable fetch failure flips the tab
// into a "waiting" state and a lightweight ping reloads the page when the
// server is back (a stale preview iframe should never stay dead).
(function () {
  var orig = window.fetch.bind(window);
  var gone = false;
  function markLost() {
    if (gone) return;
    gone = true;
    try { document.getElementById("status").textContent = "server unreachable: the page will reload automatically when it returns…"; } catch (e2) {}
  }
  window.fetch = function () {
    return orig.apply(null, arguments).catch(function () { markLost(); throw arguments[0]; });
  };
  // Sentinel and old-port rediscovery: tabs left on an older ephemeral port
  // (server previously rebound) scan the stable default port and jump back.
  var disco = false;
  // Watchdog: the ping endpoint always answers {"ok":true} while the server
  // lives. A network failure OR a non-ok answer means it went down; when it
  // answers ok again the page reloads itself (new build, fresh preview).
  setInterval(function () {
    orig("/__gui__/ping", { cache: "no-store" }).then(function (r) {
      if (!r.ok) { markLost(); return; }
      return r.json().then(function (j) {
        if (j && j.ok && gone && !disco) { disco = true; location.reload(); }
      });
    }).catch(function () { markLost(); });
    if (gone && !disco && location.port !== "32330") {
      disco = true;
      orig("http://127.0.0.1:32330/__gui__/ping", { mode: "no-cors", cache: "no-store" }).then(function () {
        location.href = "http://127.0.0.1:32330/";
      }).catch(function () { disco = false; });
    }
  }, 4000);
})();
// Pages model: a project owns one or more pages (SUI views). "children" is
// always an ALIAS of the active page's element list (remount() keeps it in
// sync after every model rebuild), so the rest of the designer keeps working
// against root.children / root.title.
function pagesModelEmpty() {
  var r = { imports: [], states: [], funcs: [], css: "", components: [], pages: [{ name: "main", title: "", onLoad: "", children: [] }], page: 0 };
  r.children = r.pages[0].children;
  return r;
}
function remount() {
  if (!root.pages || !root.pages.length) root.pages = [{ name: "main", title: root.title || "", onLoad: "", children: root.children || [] }];
  if (root.page === undefined || root.page < 0 || root.page >= root.pages.length) root.page = 0;
  root.children = root.pages[root.page].children;
  root.title = root.pages[root.page].title || "";
  return root;
}
function curPage() { remount(); return root.pages[root.page]; }
var root = pagesModelEmpty();
var suiSource = "";   // the canonical SUI source text (shown in the SUI tab)
var selPath = null;
var past = [];
var pastLabels = [];
var future = [];
var futureLabels = [];
var __textHistoryActive = false;
var __textHistoryTimer = null;
var pickMap = [];
var ast = [];
var tokens = [];
var problems = [];
var cssGen = "";
var selState = "";   // "", ":hover", ":focus", ":active"
var selCell = null;  // {r, c} selected cell inside a table
var selRow = null;   // r selected row inside a table
var selCol = null;   // c selected column inside a table
var selMode = "cell"; // "cell" | "row" | "col"
var selBody = false; // the virtual root "body" is selected
var multiSel = [];   // extra selected paths (multi-select)
var clipboard = [];  // cloned nodes for copy/paste
var marqueeMode = false;
var syncing = false;
var cssTimer = null;

var IDENT = /^[a-zA-Z_][a-zA-Z0-9_]*$/;
// Reserved SUI literals: these can never name a state or a function, and a
// binding typed as "true"/"false" must stay a literal instead of silently
// creating an invalid "state true = ..." declaration.
var RESERVED_NAME = /^(?:true|false|null)$/;

// Palette grouped by HTML category.
var PALETTE = [
  { group: "Page", tags: [], newpage: true },
  { group: "Layout", tags: ["div", "header", "footer", "section", "row", "col", "spacer", "card"],
    subs: [ { name: "Navigation", tags: ["menubar", "menu", "menuitem", "tabs", "tab", "accordion", "accordionitem", "divider"] } ] },
  { group: "Overlay", tags: ["modal", "modalbutton", "toast"] },
  { group: "Text", tags: ["h1", "h2", "h3", "p", "text", "hr", "blockquote", "pre"],
    subs: [ { name: "Lists", tags: ["list", "ul", "ol"] } ] },
  { group: "Controls", tags: ["button", "input", "textarea", "checkbox", "switch", "radio", "select", "combobox", "datepicker", "number", "slider", "file", "color", "date", "time", "tree", "treeitem", "chart", "graph"],
    subs: [ { name: "Feedback", tags: ["progress", "badge", "alert", "spinner"] } ] },
  { group: "Tables", tags: ["table"] },
  { group: "Media", tags: ["img", "video", "audio", "iframe"] }
];

var PALETTE_LABEL = { 
  div: "block", col: "column", list: "list", text: "Text", 
  textarea: "Textarea", input: "Text input", checkbox: "Checkbox", 
  radio: "Radio", switch: "Switch/Toggle", select: "Select", combobox: "Combobox", datepicker: "Date/Time picker",
  number: "Number input", slider: "Slider/Range", file: "File input",
  color: "Color picker", date: "Date", time: "Time", button: "Button",
  card: "Card", tabs: "Tabs", tab: "Tab", accordion: "Accordion", accordionitem: "Accordion item",
  tree: "Tree", treeitem: "Tree item", chart: "Chart", graph: "Graph",
  menubar: "Menu bar", menu: "Menu", menuitem: "Menu item", modal: "Modal", modalbutton: "Modal button", toast: "Toast host",
  progress: "Progress",
  badge: "Badge", alert: "Alert", spinner: "Spinner", divider: "Divider"
};

// Font choices for the Typography inspector: web-safe stacks plus any fonts
// the user has uploaded through the designer.
var BUILTIN_FONTS = [
  "system-ui, sans-serif",
  "'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
  "Arial, Helvetica, sans-serif",
  "'Helvetica Neue', Helvetica, Arial, sans-serif",
  "Verdana, Geneva, sans-serif",
  "Tahoma, Geneva, sans-serif",
  "'Trebuchet MS', Helvetica, sans-serif",
  "Roboto, sans-serif",
  "Georgia, serif",
  "'Times New Roman', Times, serif",
  "Garamond, serif",
  "'Palatino Linotype', 'Book Antiqua', serif",
  "ui-monospace, Menlo, Consolas, monospace",
  "'Courier New', Courier, monospace",
  "'Lucida Console', Monaco, monospace",
  "'Comic Sans MS', cursive, sans-serif",
  "Impact, Charcoal, sans-serif",
  "'Lucida Sans Unicode', 'Lucida Grande', sans-serif",
  "cursive",
  "fantasy",
  "monospace",
  "serif",
  "sans-serif"
];
var uploadedFonts = [];
var PREF_DEFAULTS = {
  theme: "dark", font: 14, density: "comfortable", accent: "#4c5cf7",
  showLeft: true, showRight: true, showBottom: false,
  barOrder: ["toolbar", "elementbar"], palette: [],
  editorFont: "ui-monospace, Consolas, Menlo, monospace",
  editorSize: 16, editorLineHeight: 1.6, editorTabSize: 4,
  editorWrap: false, editorLineNumbers: true,
  editorSyntaxColors: true, editorBracketMatch: true, editorSelectedLine: false
};
var prefs = Object.assign({}, PREF_DEFAULTS);

// Which SUI attributes each element exposes in the "Element" panel.
var TAG_FIELDS = {
  body: [],
  div: ["text"], section: ["text"], article: ["text"], main: ["text"], header: ["text"], footer: ["text"], nav: ["text"], aside: ["text"], span: ["text"],
  h1: ["text"], h2: ["text"], h3: ["text"], h4: ["text"], h5: ["text"], h6: ["text"],
  p: ["text"], blockquote: ["text"], pre: ["text"], code: ["text"], mark: ["text"], strong: ["text"], em: ["text"], small: ["text"],
  li: ["text"], legend: ["text"], tcell: ["text"], option: ["text"], text: [],
  list: [],
  button: ["text", "onClick"], checkbox: ["label", "checked", "disabled", "onClick", "onChange"], switch: ["label", "checked", "disabled", "onClick", "onChange"], radio: ["label", "value", "name", "checked", "disabled", "onClick", "onChange"], input: ["value", "placeholder", "inputtype", "disabled", "onClick", "onChange"], textarea: ["value", "placeholder", "disabled", "onClick", "onChange"], select: ["value", "options", "disabled", "onClick", "onChange"], number: ["value", "placeholder", "min", "max", "step", "onClick", "onChange"], slider: ["value", "min", "max", "step", "onClick", "onChange"], file: ["accept", "multiple", "onClick", "onChange"], color: ["value", "onClick", "onChange"], date: ["value", "onClick", "onChange"], time: ["value", "onClick", "onChange"], card: ["text"], a: ["text", "link", "onClick"], progress: ["value", "max"], badge: ["text"], alert: ["text"], spinner: [], divider: [], tabs: ["text"], tab: ["text"], menubar: ["text"], menu: ["text"], menuitem: ["text", "link", "onClick"], modal: ["text", "name"], modalbutton: ["text", "onClick"], 
  img: ["src"], video: ["src"], audio: ["src"], iframe: ["src", "srcAllow", "sandbox", "allow", "referrerpolicy", "loading", "width", "height"],
  input: ["value", "placeholder", "inputtype", "disabled", "required"],
  textarea: ["value", "placeholder", "disabled", "required"],
  checkbox: ["label", "checked", "disabled"],
  switch: ["label", "checked", "disabled"],
  radio: ["label", "value", "name", "checked", "disabled", "required"],
  select: ["value", "options", "disabled", "required"],
  combobox: ["value", "options", "maxOptions", "maxOptionLength", "placeholder", "empty", "required"],
  datepicker: ["value", "mode", "min", "max", "step", "disabled", "required"],
  number: ["value", "placeholder", "min", "max", "step", "required"],
  slider: ["value", "min", "max", "step"],
  file: ["accept", "multiple"],
  color: ["value"],
  date: ["value", "required"],
  time: ["value", "required"],
  card: ["text"],
  tabs: ["text"],
  tab: ["text"],
  accordion: ["text"],
  accordionitem: ["text"],
  tree: ["text", "name", "value", "editable", "editableDepth", "renameIcon", "showRenameIcon", "renameKey", "renameValue", "onChange", "onRename"],
  treeitem: ["text", "key", "open", "editable", "renameIcon", "showRenameIcon"],
  chart: ["text", "name"],
  graph: ["text", "name"],
  modal: ["text", "name"],
  modalbutton: ["text", "onClick"],
  toast: [],
  menubar: ["text"],
  menu: ["text"],
  menuitem: ["text", "link", "onClick"],
  progress: ["value", "max"],
  badge: ["text"],
  alert: ["text"],
  spinner: [],
  divider: [],
  row: [], col: [], spacer: [], ul: [], ol: [], table: [], thead: [], tbody: [], tfoot: [], trow: [], form: [], fieldset: [], hr: [], br: [], a: ["text", "link"]
};

var TAG_LABELS = { text: "Text", onClick: "onClick", placeholder: "placeholder", value: "value", src: "src / href", link: "link", inputtype: "type", pattern: "pattern", minlength: "min length", maxlength: "max length", required: "required", options: "options (comma-separated, e.g. A,B,C)", label: "label", checked: "checked (state/bool)", disabled: "disabled", name: "name / modal id", min: "min", max: "max", step: "step", accept: "accept", multiple: "multiple", directory: "directory picker", draggable: "draggable", dragKey: "drag key expression", dropKey: "drop key expression", dropFiles: "drop files into state", pickFiles: "choose files into state", pickMode: "picker API mode", variant: "variant", close: "close after click", srcAllow: "src allow-list", sandbox: "sandbox tokens", allow: "permissions policy", referrerpolicy: "referrer policy", loading: "loading", empty: "empty result text", maxOptionLength: "max option text length", maxOptions: "max option count", mode: "date/time mode" };

// Cross-browser expansion: a friendly control (e.g. "Trasparenza") maps to
// one or more real CSS declarations, incl. vendor prefixes.
var COMPAT = {
  "opacity": function(v) { return [{ prop: "opacity", value: String(v) }]; },
  "border-radius": function(v) { return [
    { prop: "-webkit-border-radius", value: v },
    { prop: "-moz-border-radius", value: v },
    { prop: "border-radius", value: v }
  ]; },
  "box-shadow": function(v) { return [{ prop: "-webkit-box-shadow", value: v }, { prop: "-moz-box-shadow", value: v }, { prop: "box-shadow", value: v }]; },
  "transform": function(v) { return [{ prop: "-webkit-transform", value: v }, { prop: "-moz-transform", value: v }, { prop: "transform", value: v }]; },
  "transition": function(v) { return [{ prop: "-webkit-transition", value: v }, { prop: "-moz-transition", value: v }, { prop: "transition", value: v }]; },
  "user-select": function(v) { return [{ prop: "-webkit-user-select", value: v }, { prop: "-moz-user-select", value: v }, { prop: "-ms-user-select", value: v }, { prop: "user-select", value: v }]; }
};
function compatDecls(prop, value) {
  var fn = COMPAT[prop];
  if (!fn) return [{ prop: prop, value: value }];
  return fn(value);
}

function newNode(kind) {
  return { kind: kind, text: "", class: "", onClick: "", placeholder: "", value: "", src: "", name: "", variant: "", close: true, draggable: false, dragKey: "", dropKey: "", dropFiles: "", pickFiles: "", pickMode: "native", accept: "", multiple: false, directory: false, children: [] };
}
// "Leaf" tags cannot contain children; everything else can nest.
var LEAF = { h1:1, h2:1, h3:1, h4:1, h5:1, h6:1, p:1, button:1, input:1, textarea:1, img:1, video:1, audio:1, iframe:1, a:1, br:1, hr:1, spacer:1, text:1, code:1, mark:1, strong:1, em:1, small:1, legend:1, option:1, list:1, checkbox:1, radio:1, switch:1, select:1, combobox:1, datepicker:1, number:1, slider:1, file:1, color:1, date:1, time:1, progress:1, badge:1, spinner:1, divider:1, menuitem:1, modalbutton:1, toast:1, chart:1, graph:1 };
function canNest(k) { return !LEAF[k]; }
function isContainer(k) { return k === "row" || k === "col"; }
function normalizeTables(nodes) {
  // inheritMatrix: when the parsed model carries style matrices (cstyles /
  // rowstyles), adopt them so cell styling survives every reparse.
  nodes.forEach(function(n) {
    if (n.kind === "table" && n.cols) {
      if (Array.isArray(n.cstyles)) n.cellStyles = n.cstyles;
      if (Array.isArray(n.rowstyles)) n.rowStyles = n.rowstyles;
      if (!n.cellStyles) {
        n.cellStyles = [];
        for (var r = 0; r < n.rows; r++) { var row = []; for (var c = 0; c < n.cols; c++) row.push({ bg: "", color: "", extra: "" }); n.cellStyles.push(row); }
      }
      if (!n.rowStyles) {
        n.rowStyles = [];
        for (var r2 = 0; r2 < n.rows; r2++) n.rowStyles.push({ bg: "", color: "", extra: "" });
      }
      if (!n.cells) n.cells = [];
      if (!n.spans) n.spans = [];
      for (var r3 = 0; r3 < n.rows; r3++) {
        if (!n.cells[r3]) n.cells[r3] = [];
        if (!n.spans[r3]) n.spans[r3] = [];
        if (!n.cellStyles[r3]) n.cellStyles[r3] = [];
        for (var c2 = 0; c2 < n.cols; c2++) {
          if (n.cells[r3][c2] === undefined) n.cells[r3][c2] = "";
          if (!n.spans[r3][c2]) n.spans[r3][c2] = { cs: 1, rs: 1 };
          if (!n.cellStyles[r3][c2]) n.cellStyles[r3][c2] = { bg: "", color: "", extra: "" };
          else if (!n.cellStyles[r3][c2].extra) n.cellStyles[r3][c2].extra = "";
        }
        if (!n.rowStyles[r3]) n.rowStyles[r3] = { bg: "", color: "", extra: "" };
        else if (!n.rowStyles[r3].extra) n.rowStyles[r3].extra = "";
      }
      // Delete the raw matrices: the structured fields are the working copy.
      delete n.cstyles;
      delete n.rowstyles;
      if (n.caption === undefined) n.caption = "";
      if (n.capstyle === undefined) n.capstyle = "";
      if (n.captionShow === undefined) n.captionShow = !!(n.caption && n.caption.trim());
      if (n.editable === undefined) n.editable = false;
      if (n.fixed === undefined) n.fixed = false;
      if (n.name === undefined) n.name = "";
    }
    if (n.kind === "list") {
      if (!Array.isArray(n.items)) n.items = [];
      if (n.ordered === undefined) n.ordered = false;
      if (n.marker === undefined) n.marker = "";
    }
    if (n.kind === "text" && Array.isArray(n.runs)) {
      if (!n.tag) n.tag = "p";
      if (n.align === undefined) n.align = "";
      if (n.size === undefined) n.size = "";
      n.runs.forEach(function(r){ if (r.b === undefined) r.b = false; if (r.i === undefined) r.i = false; if (r.u === undefined) r.u = false; if (r.color === undefined) r.color = ""; if (r.bg === undefined) r.bg = ""; if (r.font === undefined) r.font = ""; });
    }
    if (n.children && n.children.length) normalizeTables(n.children);
  });
}
function newTableNode() {
  var n = newNode("table");
  n.cols = 3;
  n.header = 1;   // number of header rows
  n.rows = 3;     // total rows (header + body)
  n.class = "suitbl"; // hooks the auto-managed table base rules (zebra, padding...)
  n.caption = "";
  n.capstyle = "";
  n.captionShow = false;
  n.editable = false;
  n.fixed = false;
  n.name = "";
  ensureTableBaseCss();
  n.cells = [];
  n.spans = [];
  n.cellStyles = [];
  n.rowStyles = [];
  for (var r = 0; r < n.rows; r++) {
    n.cells.push([]);
    n.spans.push([]);
    n.cellStyles.push([]);
    n.rowStyles.push({ bg: "", color: "", extra: "" });
    for (var c = 0; c < n.cols; c++) {
      n.cells[r].push("");
      n.spans[r].push({ cs: 1, rs: 1 });
      n.cellStyles[r].push({ bg: "", color: "", extra: "" });
    }
  }
  return n;
}
function newListNode() {
  var n = newNode("list");
  n.ordered = false;   // false -> <ul>, true -> <ol>
  n.marker = "";       // CSS list-style-type (disc, decimal, …)
  n.items = ["Item 1", "Item 2", "Item 3"];
  return n;
}
// newMenuNode builds one top-level menu with free-form items. Actions are
// intentionally empty: the inspector lets the user set onClick/link per item.
function newMenuNode(label, items) {
  var n = newNode("menu");
  n.text = label || "Menu";
  for (var i = 0; items && i < items.length; i++) {
    var item = newNode("menuitem");
    item.text = items[i];
    n.children.push(item);
  }
  return n;
}
function newMenuBarNode() {
  var n = newNode("menubar");
  n.text = "Main menu";  // accessible name (aria-label), not a visible label
  n.children = [
    newMenuNode("File", ["New", "Open", "Save"]),
    newMenuNode("Edit", ["Undo", "Redo"]),
    newMenuNode("Help", ["About"])
  ];
  return n;
}
function newTabNode(label) {
  var n = newNode("tab");
  n.text = label || "Tab";
  var body = newTextNode();
  body.tag = "p";
  body.runs[0].text = "Content";
  n.children = [body];
  return n;
}
function newTabsNode() {
  var n = newNode("tabs");
  n.text = "Tabs";  // accessible name (aria-label)
  n.children = [newTabNode("Tab 1"), newTabNode("Tab 2")];
  return n;
}
function newAccordionItemNode(label) {
  var n = newNode("accordionitem");
  n.text = label || "Section";
  var body = newTextNode();
  body.tag = "p";
  body.runs[0].text = "Content";
  n.children = [body];
  return n;
}
function newAccordionNode() {
  var n = newNode("accordion");
  n.text = "Accordion";  // accessible name (aria-label)
  n.children = [newAccordionItemNode("Section 1"), newAccordionItemNode("Section 2")];
  return n;
}
function newModalButtonNode(label, variant) {
  var n = newNode("modalbutton");
  n.text = label || "Button";
  n.variant = variant || "";
  n.close = true;
  return n;
}
function nextModalName() {
  var used = {};
  (root.pages || []).forEach(function(p) {
    (function walk(nodes) {
      (nodes || []).forEach(function(n) {
        if (n.kind === "modal" && n.name) used[n.name] = true;
        if (n.children) walk(n.children);
      });
    })(p.children || []);
  });
  var base = "dialog", i = 1;
  while (used[base]) { i++; base = "dialog" + i; }
  return base;
}
function newModalNode() {
  var n = newNode("modal");
  n.text = "Dialog";
  n.name = nextModalName();
  var body = newTextNode();
  body.tag = "p";
  body.runs[0].text = "Dialog content";
  n.children = [
    body,
    newModalButtonNode("Cancel"),
    newModalButtonNode("OK", "primary")
  ];
  return n;
}
function uniqueStateName(base) {
  var i = 1;
  while (hasState(base + i)) i++;
  return base + i;
}
function newComboboxNode() {
  var n = newNode("combobox");
  n.text = "Choose";
  // Start with an empty, editable options state. Adding rows in the inspector
  // updates this state's initial value; application code can replace it later
  // at runtime and the view re-render refreshes the combobox.
  var sname = uniqueStateName("options");
  if (!hasState(sname)) root.states.push({ name: sname, init: "\"\"" });
  n.options = sname;
  n.optionsExpr = true;
  n.placeholder = "Search...";
  n.empty = "No results";
  return n;
}
function newDatepickerNode() {
  var n = newNode("datepicker");
  n.text = "Date";
  n.mode = "date";
  n.value = "";
  return n;
}
function newTreeItemNode(label) {
  var n = newNode("treeitem");
  n.text = label || "Item";
  n.key = "";
  n.open = false;
  return n;
}
function newTreeNode() {
  var n = newNode("tree");
  n.text = "Tree";
  var parent = newTreeItemNode("Parent");
  parent.open = true;
  parent.children = [newTreeItemNode("Child")];
  n.children = [parent];
  n.value = "";
  n.editable = false;
  n.editableDepth = 0;
  n.styleExpr = treePresetStyleExpr("default");
  return n;
}
function newChartNode() {
  var n = newNode("chart");
  n.text = "Chart";
  n.chartType = "bar";
  n.chartTheme = "dark";
  n.chartData = "Month,Revenue\nJan,12\nFeb,18\nMar,15\nApr,24\nMay,21\nJun,29";
  n.chartDataExpr = false;
  n.chartFormat = "auto";
  n.chartX = "";
  n.chartY = "";
  n.chartWidth = "640px";
  n.chartHeight = "320px";
  n.chartOverflow = "scroll";
  n.chartZoom = "1";
  n.chartPointWidth = "48";
  n.chartPointSize = "3";
  n.chartTableView = "chart";
  n.chartTableFilters = true;
  n.chartTableAddRow = false;
  n.chartTableDeleteRow = false;
  n.chartTableSteppers = false;
  n.chartTableContextMenu = false;
  n.chartTableSelection = "row";
  n.chartCriticalPath = false;
  n.chartHolidays = "";
  n.chartWeekend = "";
  n.chartTimezone = "";
  n.chartShowNonWorkingDays = true;
  n.chartBins = "10";
  n.chartMin = "";
  n.chartMax = "";
  n.chartInteractive = false;
  n.chartStateMode = "editing";
  n.chartStateHints = true;
  n.chartStateHintBg = "";
  n.chartStateHintOpacity = "";
  n.chartStateHintTextColor = "";
  n.chartStateHintBorderColor = "";
  n.chartStateHintFont = "";
  n.chartStateHintFontSize = "";
  n.chartStateHintBorderWidth = "";
  n.chartStateHintPadding = "";
  n.chartStateHintRadius = "";
  n.chartColorPicker = true;
  n.chartShowResources = false;
  n.chartShowProgress = false;
  n.chartProgressColumn = false;
  n.chartShowSlack = false;
  n.chartShowCosts = false;
  n.chartStatusColors = false;
  n.chartAutoSchedule = false;
  n.chartGanttScale = "auto";
  n.chartLinkColor = "#94a3b8";
  n.chartLinkWidth = "1";
  n.chartLinkGap = "4";
  n.chartLinkStyle = "auto";
  n.chartArrowType = "triangle";
  n.chartArrowSize = "5";
  n.chartBarBorderColor = "";
  n.chartGrid = true;
  n.chartLegend = true;
  n.chartValues = false;
  n.styleExpr = chartPresetStyleExpr("default");
  return n;
}
function isChartKind(n) {
  return !!n && (n.kind === "chart" || n.kind === "graph");
}
function isGraphKind(n) {
  return !!n && n.kind === "graph";
}
function newGraphNode() {
  var n = newChartNode();
  n.kind = "graph";
  n.text = "Graph";
  n.chartType = "state";
  n.chartData = CHART_GRAPH_SAMPLE_JSON;
  n.chartDataExpr = false;
  n.chartFormat = "auto";
  return n;
}
function newNodeForKind(kind) {
  if (kind === "combobox") return newComboboxNode();
  if (kind === "datepicker") return newDatepickerNode();
  if (kind === "tree") return newTreeNode();
  if (kind === "treeitem") return newTreeItemNode("Item");
  if (kind === "chart") return newChartNode();
  if (kind === "graph") return newGraphNode();
  if (kind === "table") return newTableNode();
  if (kind === "modal") return newModalNode();
  if (kind === "modalbutton") return newModalButtonNode("Button");
  if (kind === "list") return newListNode();
  if (kind === "text") return newTextNode();
  if (kind === "tabs") return newTabsNode();
  if (kind === "tab") return newTabNode("Tab");
  if (kind === "accordion") return newAccordionNode();
  if (kind === "accordionitem") return newAccordionItemNode("Section");
  if (kind === "menubar") return newMenuBarNode();
  if (kind === "menu") return newMenuNode("Menu", ["Item 1", "Item 2"]);
  if (kind === "menuitem") {
    var item = newNode("menuitem");
    item.text = "Item";
    return item;
  }
  return newNode(kind);
}
function newTextNode() {
  var n = newNode("text");
  n.tag = "p";         // block element: p / h1..h6 / blockquote / pre
  n.align = "";        // text-align (left/center/right/justify)
  n.size = "";         // base font-size for the whole block
  n.runs = [{ text: "", b: false, i: false, u: false, color: "", bg: "", font: "" }];
  return n;
}
function byPath(path, arr) {
  var cur = { children: arr };
  for (var i = 0; i < path.length; i++) cur = cur.children[path[i]];
  return cur;
}
function currentSourceForHistory() {
  var ta = document.getElementById("out");
  return ta ? ta.value : (suiSource || "");
}
// A history entry stores the designer model AND the SUI text. Restoring both is
// what makes one undo stack cover property edits, element changes, source
// typing and component insertion.
function historyState(source) {
  return JSON.stringify({ root: root, source: source === undefined ? currentSourceForHistory() : source });
}
function restoreHistoryEntry(raw, label) {
  var st = JSON.parse(raw);
  root = st.root;
  remount();
  computeElementLines();
  selPath = null; multiSel = []; selBody = false; selState = "";
  selCell = null; selRow = null; selCol = null;
  buildPickMap();
  renderTree();
  renderView();
  buildDatalist();
  renderPages();
  clearFields();
  renderInspector();
  renderProblems();
  renderTokens();
  suiSource = (st.source === undefined || st.source === null) ? generate() : st.source;
  setSuiEditor();
  updateApply();
  autoPreview();
  scheduleCssSync();
  hideSuiAc();
  closeSnippetMenu();
  closeContextMenus();
  setStatus(label || "undone");
}
function endTextHistoryBurst() {
  __textHistoryActive = false;
  if (__textHistoryTimer) { clearTimeout(__textHistoryTimer); __textHistoryTimer = null; }
}
function snap(label, source, fromText) {
  if (!fromText) endTextHistoryBurst();
  past.push(historyState(source));
  pastLabels.push(label || "");
  if (past.length > 80) { past.shift(); pastLabels.shift(); }
  future = []; futureLabels = [];
  renderHistory();
}
function undo() {
  if (!past.length) { setStatus("nothing to undo"); return; }
  endTextHistoryBurst();
  future.push(historyState());
  futureLabels.push(pastLabels[pastLabels.length - 1] || "");
  var raw = past.pop();
  pastLabels.pop();
  restoreHistoryEntry(raw, "undone");
  renderHistory();
}
function redo() {
  if (!future.length) { setStatus("nothing to redo"); return; }
  endTextHistoryBurst();
  past.push(historyState());
  pastLabels.push(futureLabels[futureLabels.length - 1] || "");
  var raw = future.pop();
  futureLabels.pop();
  restoreHistoryEntry(raw, "ripetuto");
  renderHistory();
}
function jumpTo(i) {
  if (i < 0 || i >= past.length) return;
  endTextHistoryBurst();
  var steps = past.length - i, last = null;
  for (var k = 0; k < steps && past.length; k++) {
    future.push(historyState());
    futureLabels.push(pastLabels[pastLabels.length - 1] || "");
    last = past.pop();
    pastLabels.pop();
  }
  if (last) restoreHistoryEntry(last, "reverted");
  renderHistory();
}
function withUndo(fn, label) { snap(label); fn(); renderAll(); autoPreview(); reparseSui(); }

// reparseSui re-derives the tree/preview from the SUI text (single source of
// truth), guaranteeing the structure always matches the generated source.
// reparseSui serializes model round-trips: each withUndo fires one, and with
// quick successive edits a stale response would roll the tree back (eating
// user changes — the "parameters have no effect" bug). If a request is already
// in flight we don't cancel it: we remember to re-run right after, always with
// the latest suiSource. A response whose source predates the current one is
// never applied.
var reparseInFlight = false;
var reparseReRun = false;
function reparseSui() {
  if (reparseInFlight) { reparseReRun = true; return; }
  reparseInFlight = true;
  var src = suiSource;
  var fd = new URLSearchParams();
  fd.set("sui", suiSource);
  fetch("/__gui__/open", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    reparseInFlight = false;
    if (j.ok && suiSource === src) {
      var prevName = curPage().name;
      root = { states: j.states || [], funcs: j.funcs || [], css: j.css || "", components: j.components || [], page: 0,
        pages: (j.pages && j.pages.length ? j.pages.map(function(p){ return { name: p.name || "main", title: p.title || "", children: p.elements || [] }; }) : [{ name: "main", title: j.title || "", children: j.elements || [] }]) };
      root.pages.forEach(function(p){ normalizeTables(p.children); });
      for (var i = 0; i < root.pages.length; i++) { if (root.pages[i].name === prevName) { root.page = i; break; } }
      remount();
      computeElementLines();
      renderAll();
      renderPages();
      autoPreview();
      scheduleLiveCheck();
    }
    if (reparseReRun) { reparseReRun = false; reparseSui(); }
  }).catch(function(){
    reparseInFlight = false;
    if (reparseReRun) { reparseReRun = false; reparseSui(); }
  });
}

function renderAll() {
  buildPickMap();
  renderTree();
  renderView();
  buildDatalist();
  renderGutters();
  if (selPath) fillPanel(); else clearFields();
  setOut();
  renderInspector();
  renderPages();
  scheduleCssSync();
}

function setOut() { suiSource = generate(); setSuiEditor(); dbg("setOut children=" + root.children.length + " kinds=[" + root.children.map(function(n){ return n.kind; }).join(",") + "]"); }

// ---- panel switching ------------------------------------------------------
function switchLeft(t) {
  document.querySelectorAll("#left .tab").forEach(function(x){ x.classList.toggle("on", x.getAttribute("data-tab")===t); });
  document.getElementById("l-structure").style.display = t === "structure" ? "" : "none";
  document.getElementById("l-tokens").style.display = t === "tokens" ? "" : "none";
}
function switchC(t) {
  document.querySelectorAll("#ctabs .tab").forEach(function(x){ x.classList.toggle("on", x.getAttribute("data-tab")===t); });
  document.getElementById("cpane-preview").classList.toggle("on", t==="preview");
  document.getElementById("cpane-sui").classList.toggle("on", t==="sui");
  if (t === "sui") setSuiEditor();
}
function switchB(t) {
  document.querySelectorAll("#btabs .tab").forEach(function(x){ x.classList.toggle("on", x.getAttribute("data-tab")===t); });
  document.getElementById("problems").style.display = t === "problems" ? "" : "none";
  document.getElementById("history").style.display = t === "history" ? "" : "none";
}
function toggleBottom() {
  closeMenus();
  var b = document.getElementById("bottom");
  b.classList.toggle("on");
  var on = b.classList.contains("on");
  prefs.showBottom = on;
  savePrefs();
  document.getElementById("probbtn").textContent = on ? "Hide problems panel" : "Problems panel";
}
function setBp(bp) {
  document.querySelectorAll('#drop-view .mi[data-bp-mi]').forEach(function(x){ x.classList.toggle("on", x.getAttribute("data-bp-mi")===bp); });
  var st = document.getElementById("stage");
  st.className = bp;
  var f = document.getElementById("frame");
  f.style.width = ""; f.style.height = "";
  var w = bp === "tablet" ? 768 : bp === "mobile" ? 375 : "";
  document.getElementById("vpw").value = w;
  document.getElementById("vph").value = "";
  closeMenus();
}
function onViewportInput() {
  var w = parseInt(document.getElementById("vpw").value, 10) || 0;
  var h = parseInt(document.getElementById("vph").value, 10) || 0;
  applyViewport(w, h);
}
function applyViewport(w, h) {
  var f = document.getElementById("frame");
  var st = document.getElementById("stage");
  f.style.width = w > 0 ? w + "px" : "";
  f.style.height = h > 0 ? h + "px" : "";
  if (w > 0 || h > 0) st.classList.add("custom"); else st.classList.remove("custom");
  document.querySelectorAll('#drop-view .mi[data-bp-mi]').forEach(function(x){ x.classList.remove("on"); });
}

// ---- tree -----------------------------------------------------------------
// pickMap mirrors the compiler's DFS data-sui-id numbering exactly: regular
// nodes are one entry, and a table expands into thead/trow/tcell/tbody
// entries so clicks/highlights map to cells/rows too. Multi-page previews
// keep every page in the DOM at once, so the counter must NOT restart per
// page (the compiled preview continues it across views): every entry is
// tagged with its page (pg) and paths stay local to that page.
function buildPickMap() {
  pickMap = [];
  var pages = root.pages || [root];
  for (var pi = 0; pi < pages.length; pi++) {
    (function(pi) {
      (function walk(nodes, path) {
        for (var i = 0; i < nodes.length; i++) {
          var n = nodes[i];
          var p = path.concat([i]);
          if (n.raw) {
            // component invocation: rendered without a data-sui-id, so it must
            // not consume a pickMap slot (keeps ids aligned with the preview)
            continue;
          }
          pickMap.push({ t: "n", pg: pi, p: p });
          if (n.kind === "list") {
            for (var k = 0; k < (n.items || []).length; k++) {
              pickMap.push({ t: "item", pg: pi, p: p, k: k });
            }
          } else if (n.kind === "text" && n.runs) {
            // each run renders as a leaf plus a wrapper per bold/italic flag
            (n.runs || []).forEach(function(r) {
              var extra = 1 + (r.b ? 1 : 0) + (r.i ? 1 : 0);
              for (var k = 0; k < extra; k++) pickMap.push({ t: "n", pg: pi, p: p });
            });
          } else if (n.kind === "table" && n.cols) {
            // caption is emitted as the table's first child when shown
            if (n.captionShow !== false && n.caption && n.caption.trim()) pickMap.push({ t: "sec", pg: pi, p: p }); // caption
            pickMap.push({ t: "sec", pg: pi, p: p }); // thead
            for (var r = 0; r < n.header; r++) {
              pickMap.push({ t: "row", pg: pi, p: p, r: r });
              for (var c = 0; c < n.cols; c++) {
                if (cellCovered(n, r, c)) continue;
                pickMap.push({ t: "cell", pg: pi, p: p, r: r, c: c });
              }
            }
            pickMap.push({ t: "sec", pg: pi, p: p }); // tbody
            for (var r = n.header; r < n.rows; r++) {
              pickMap.push({ t: "row", pg: pi, p: p, r: r });
              for (var c = 0; c < n.cols; c++) {
                if (cellCovered(n, r, c)) continue;
                pickMap.push({ t: "cell", pg: pi, p: p, r: r, c: c });
              }
            }
          } else if (n.children && n.children.length) {
            walk(n.children, p);
          }
        }
      })(pages[pi].children, []);
    })(pi);
  }
}
function selectByDfs(id) {
  var t = pickMap[id];
  if (!t || t.pg !== root.page) return; // ignore picks from other (hidden) pages
  selBody = false;
  multiSel = [];
  if (t.t === "cell") {
    selPath = t.p; selCell = null; selRow = null; selCol = null;
    // "table" mode: clicking anywhere inside the table selects the whole table.
    if (selMode === "table") { /* whole table */ }
    else if (selMode === "row") selRow = t.r;
    else if (selMode === "col") selCol = t.c;
    else selCell = { r: t.r, c: t.c };
  } else if (t.t === "row") {
    selPath = t.p; selRow = null; selCell = null; selCol = null;
    if (selMode !== "table") selRow = t.r;
  } else {
    selPath = t.p; selCell = null; selRow = null; selCol = null;
  }
  renderTree(); fillPanel(); renderInspector(); postHighlight();
  dbg("select id=" + id + " path=[" + (selPath ? selPath.join(",") : "?") + "] children=" + root.children.length);
}
function currentPeerIds() {
  if (selCell && selPath) {
    for (var i = 0; i < pickMap.length; i++) {
      var e = pickMap[i];
      if (e.pg === root.page && e.t === "cell" && samePath(e.p, selPath) && e.r === selCell.r && e.c === selCell.c) return [i];
    }
  }
  if (selRow !== null && selPath) {
    for (var i = 0; i < pickMap.length; i++) {
      var e2 = pickMap[i];
      if (e2.pg === root.page && e2.t === "row" && samePath(e2.p, selPath) && e2.r === selRow) return [i];
    }
  }
  if (selCol !== null && selPath) {
    var ids = [];
    for (var i = 0; i < pickMap.length; i++) {
      var e3 = pickMap[i];
      if (e3.pg === root.page && e3.t === "cell" && samePath(e3.p, selPath) && e3.c === selCol) ids.push(i);
    }
    if (ids.length) return ids;
  }
  if (selPath) {
    for (var i = 0; i < pickMap.length; i++) {
      if (pickMap[i].pg === root.page && pickMap[i].t === "n" && samePath(pickMap[i].p, selPath)) return [i];
    }
  }
  return [];
}
function postHighlight() {
  var f = document.getElementById("prev");
  if (f && f.contentWindow) {
    f.contentWindow.postMessage({ ty: "sui-hl", id: currentPeerIds(), mode: selMode }, "*");
  }
  editorSyncFromSelection();
}
// editorSyncFromSelection reveals, in the SUI editor, the block of the element
// currently selected in the tree/preview. It is the tree -> editor half of the
// two-way synchronisation.
function editorSyncFromSelection() {
  if (selBody || !selPath || selPath.length === 0) { return; }
  var it = __elLines.filter(function(x){ return samePath(x.path, selPath); })[0];
  if (!it) return;
  var ta = document.getElementById("out");
  if (!ta || !ta.value) return;
  __elline = { line: it.start };
  // Only move the caret when the editor is not focused, so typing isn't
  // interrupted while exploring the tree.
  if (document.activeElement !== ta) {
    var src = ta.value;
    var off = lineColOffset(src, it.start, 0);
    ta.selectionStart = ta.selectionEnd = off;
    var lh = editorLineHeightPx();
    ta.scrollTop = Math.max(0, it.start * lh - ta.clientHeight / 2 + lh);
    var hl = document.getElementById("sui_hl");
    if (hl) hl.scrollTop = ta.scrollTop;
  }
  repaintBracketMatch();
}
// computeElementLines records the start (and best-effort end) line of every
// element in the model, so a selection can be mapped onto the editor text.
// Uses the line emitted by the designer on each element.
function computeElementLines() {
  __elLines = [];
  function walk(nodes, path) {
    for (var i = 0; i < nodes.length; i++) {
      var n = nodes[i], p = path.concat([i]);
      var ln = (typeof n.line === "number") ? n.line : 0;
      __elLines.push({ path: p, start: ln, end: undefined, kind: n.kind });
      if (n.children && n.children.length) walk(n.children, p);
    }
  }
  walk(root.children || [], []);
  // Fill block ends: an element's block ends just before the next element
  // declared at the same or a shallower depth in source order.
  var sorted = __elLines.slice().sort(function(a, b){ return a.start - b.start; });
  for (var j = 0; j < sorted.length; j++) {
    var cur = sorted[j];
    if (!cur.start) continue;
    // Find the next element whose line is greater; if it is at the same or a
    // shallower depth (shorter path), this element's block ends before it.
    for (var k = j + 1; k < sorted.length; k++) {
      if (sorted[k].start > cur.start) {
        if (sorted[k].path.length <= cur.path.length) { cur.end = sorted[k].start - 1; }
        break;
      }
    }
  }
}
var __treeCollapsed = Object.create(null);
function treeCollapseKey(path) { return String(root.page || 0) + ":" + path.join(","); }
function isTreeCollapsed(path) { return !!__treeCollapsed[treeCollapseKey(path)]; }
function toggleTreeCollapsed(path) {
  var key = treeCollapseKey(path);
  if (__treeCollapsed[key]) delete __treeCollapsed[key];
  else __treeCollapsed[key] = true;
  renderTree();
}
function renderTree() {
  var box = document.getElementById("nodes");
  box.innerHTML = "";
  var bl = document.createElement("div");
  bl.className = "node" + (selBody ? " sel" : "");
  bl.style.cssText = "color:#8fa0b8;font-weight:600;cursor:pointer";
  bl.textContent = "body";
  bl.onclick = function() { selBody = true; selPath = null; selCell = null; selRow = null; selCol = null; renderTree(); clearFields(); renderInspector(); postHighlight(); };
  bl.oncontextmenu = function(ev) {
    ev.preventDefault();
    ev.stopPropagation();
    selBody = true; selPath = null; selCell = null; selRow = null; selCol = null;
    renderTree(); renderInspector();
    showNodeContextMenu(ev, [], { kind: "body" });
  };
  box.appendChild(bl);
  var ec = document.getElementById("elmcount");
  if (ec) ec.textContent = "Elements (" + root.children.length + ")";
  renderNodes(root.children, 1, box, []);
  updateToolbar();
}
function unusedClasses() {
  var set = {};
  problems.forEach(function(p){ if (p.kind === "unused") set[p.sel.replace(/^\./,"")] = true; });
  return set;
}
function renderNodes(children, depth, parent, path) {
  var unused = unusedClasses();
  for (var i = 0; i < children.length; i++) {
    var p = path.concat([i]);
    var node = children[i];
    var row = document.createElement("div");
    var cls = node.class ? node.class.trim().split(/\s+/)[0] : "";
    var isUnused = cls && unused[cls];
    var isSel = samePath(selPath, p);
    var isMulti = !isSel && multiHas(p);
    row.className = "node" + (isSel ? " sel" : "") + (isMulti ? " multi" : "") + (isUnused ? " unused" : "");
    row.draggable = true;
    row.style.paddingLeft = (4 + depth * 12) + "px";
    row.setAttribute("data-path", p.join(","));
    bindRowEvents(row, p, node, children, i);
    var hasChildren = !!(node.children && node.children.length);
    var collapsed = hasChildren && isTreeCollapsed(p);
    var twist = document.createElement("span");
    twist.className = "node twist";
    twist.textContent = hasChildren ? (collapsed ? "\u25B8" : "\u25BE") : " ";
    if (hasChildren) {
      twist.title = collapsed ? "Expand" : "Collapse";
      twist.setAttribute("role", "button");
      twist.setAttribute("aria-expanded", collapsed ? "false" : "true");
      twist.onmousedown = function(ev) { ev.stopPropagation(); };
      // Capture this row's path per iteration. A plain var would be shared by
      // every closure and clicking one arrow would toggle the last row instead.
      twist.onclick = (function(treePath) {
        return function(ev) {
          ev.preventDefault();
          ev.stopPropagation();
          toggleTreeCollapsed(treePath);
        };
      })(p.slice());
    }
    row.appendChild(twist);
    var lab = document.createElement("span");
    var labelText;
    if (node.kind === "text") {
      var txt = (node.runs || []).map(function(r){ return r.text; }).join("");
      labelText = (node.tag && node.tag !== "p" ? node.tag + " '" : "text ") + (txt ? txt.slice(0, 18) : "");
      if (node.tag && node.tag !== "p" && txt) labelText += "'";
    } else if (node.kind === "modal") {
      labelText = "modal " + (node.text ? "'" + node.text + "'" : "") + (node.name ? " #" + node.name : "");
    } else if (node.kind === "modalbutton") {
      labelText = "modalbutton '" + (node.text || "") + "'";
    } else if (node.kind === "tabs") {
      labelText = "tabs " + (node.text ? "'" + node.text + "'" : "''");
    } else if (node.kind === "tab") {
      labelText = "tab '" + (node.text || "") + "'";
    } else if (node.kind === "accordion") {
      labelText = "accordion " + (node.text ? "'" + node.text + "'" : "''");
    } else if (node.kind === "accordionitem") {
      labelText = "item '" + (node.text || "") + "'";
    } else if (node.kind === "toast") {
      labelText = "toast host";
    } else if (node.kind === "combobox") {
      labelText = "combobox " + (node.text ? "'" + node.text + "'" : "''");
    } else if (node.kind === "datepicker") {
      labelText = "datepicker " + (node.text ? "'" + node.text + "'" : "''");
    } else if (node.kind === "tree") {
      labelText = "tree " + (node.text ? "'" + node.text + "'" : "''");
    } else if (node.kind === "treeitem") {
      labelText = "treeitem " + (node.text ? "'" + node.text + "'" : "''");
    } else if (node.kind === "chart") {
      labelText = "chart " + (node.chartType || "bar") + (node.text ? " '" + node.text + "'" : "");
    } else if (node.kind === "graph") {
      labelText = "graph block diagram" + (node.text ? " '" + node.text + "'" : "");
    } else if (node.kind === "menubar") {
      labelText = "menubar " + (node.text ? "'" + node.text + "'" : "''");
    } else if (node.kind === "menu") {
      labelText = "menu '" + (node.text || "") + "'";
    } else if (node.kind === "menuitem") {
      labelText = "item '" + (node.text || "") + "'";
    } else {
      labelText = node.text || node.kind;
    }
    var typeChip = document.createElement("span");
    typeChip.className = "node-type";
    typeChip.textContent = node.kind === "text" && node.tag ? node.tag : node.kind;
    typeChip.title = "Exact element type";
    row.appendChild(typeChip);
    lab.textContent = labelText + (cls ? " ." + cls : "") + (node.component ? "  \u2022" : "");
    row.appendChild(lab);
    var ops = document.createElement("span");
    ops.style.marginLeft = "auto";
    ops.innerHTML = '<span class="btn" style="font-size:9px;padding:0 4px" onclick="event.stopPropagation();delSelPath(\'' + p.join(",") + '\')">&#215;</span>';
    row.appendChild(ops);
    parent.appendChild(row);
    if (hasChildren && !collapsed) renderNodes(node.children, depth + 1, parent, p);
  }
}
function bindRowEvents(row, p, node, tray, i) {
  row.ondragstart = function (ev) { ev.dataTransfer.setData("text/plain", p.join(",")); ev.dataTransfer.effectAllowed = "move"; };
  row.ondragover = function (ev) { ev.preventDefault(); ev.currentTarget.classList.add("hot"); ev.dataTransfer.dropEffect = "move"; };
  row.ondragleave = function (ev) { ev.currentTarget.classList.remove("hot"); };
  row.ondrop = function (ev) {
    ev.preventDefault();
    ev.currentTarget.classList.remove("hot");
    var srcStr = ev.dataTransfer.getData("text/plain");
    dropAt(srcStr, tray, i, node, p);
  };
  row.onclick = function (ev) {
    ev.stopPropagation();
    selCell = null; selRow = null; selCol = null; selBody = false;
    if (ev.shiftKey) {
      rangeSelect((selPath || p).slice(), p);
    } else if (ev.ctrlKey || ev.metaKey) {
      // Seed the current single selection into the multi-set before toggling.
      if (!multiSel.length && selPath && !samePath(selPath, p)) multiAdd(selPath);
      if (multiHas(p)) multiRemove(p); else multiAdd(p);
      if (multiHas(p)) selPath = p.slice();
      else selPath = multiSel.length ? multiSel[multiSel.length - 1].slice() : null;
    } else {
      multiSel = [];
      selPath = p;
    }
    renderTree(); fillPanel(); renderInspector();
  };
  // Right-click opens the element-list context menu. It is also the quick
  // menu editor: add/remove menu nodes without leaving the tree.
  row.oncontextmenu = function (ev) {
    ev.preventDefault();
    ev.stopPropagation();
    selCell = null; selRow = null; selCol = null; selBody = false;
    multiSel = [];
    selPath = p.slice();
    renderTree(); fillPanel(); renderInspector(); postHighlight();
    showNodeContextMenu(ev, p, node);
  };
}
// ---- element-list context menu / menu editor -----------------------------
// The tree is the menu editor: right-click any element for structural actions.
// Menu nodes get Add menu / Add item / Add submenu; all rows get duplicate,
// move and delete.
function showNodeContextMenu(ev, p, node) {
  closeMenus();
  var old = document.getElementById("node-menu");
  if (old) old.remove();
  var actions = nodeContextActions(p, node);
  if (!actions.length) return;
  var m = document.createElement("div");
  m.id = "node-menu";
  m.className = "sui-ac";
  m.style.display = "block";
  m.style.minWidth = "210px";
  actions.forEach(function(it) {
    if (it.sep) {
      var sep = document.createElement("div");
      sep.className = "ctx-sep";
      m.appendChild(sep);
      return;
    }
    var d = document.createElement("div");
    d.className = "sui-ac-item";
    d.textContent = it.label;
    d.onclick = function() { m.remove(); it.fn(); };
    m.appendChild(d);
  });
  document.body.appendChild(m);
  var z = menuZoom();
  m.style.left = (ev.pageX / z) + "px";
  m.style.top = (ev.pageY / z) + "px";
  var close = function() {
    m.remove();
    document.removeEventListener("click", close);
  };
  setTimeout(function() { document.addEventListener("click", close); }, 0);
}

function nodeContextActions(p, node) {
  var out = [];
  function add(label, fn) { out.push({ label: label, fn: fn }); }
  function sep() { if (out.length && !out[out.length - 1].sep) out.push({ sep: true }); }

  if (node.kind === "body") {
    add("Add menu bar", function() {
      selBody = true; selPath = null; multiSel = [];
      addEl("menubar");
    });
    return out;
  }

  if (node.kind === "modal") {
    add("Add modal button", function() { insertModalButton(p, node); });
    sep();
  } else if (node.kind === "modalbutton") {
    add("Add button below", function() { insertModalButton(p, node); });
    sep();
  } else if (node.kind === "tabs") {
    add("Add tab", function() { insertTabNode(p, node); });
    sep();
  } else if (node.kind === "tab") {
    add("Add tab below", function() { insertTabNode(p, node); });
    sep();
  } else if (node.kind === "accordion") {
    add("Add accordion item", function() { insertAccordionItem(p, node); });
    sep();
  } else if (node.kind === "accordionitem") {
    add("Add item below", function() { insertAccordionItem(p, node); });
    sep();
  } else if (node.kind === "menubar") {
    add("Add menu", function() { insertMenuNode(p, node, "menu"); });
    sep();
  } else if (node.kind === "menu") {
    add("Add menu item", function() { insertMenuNode(p, node, "menuitem"); });
    add("Add submenu", function() { insertMenuNode(p, node, "menu"); });
    sep();
  } else if (node.kind === "menuitem") {
    add("Add item below", function() { insertMenuNode(p, node, "menuitem"); });
    add("Add submenu below", function() { insertMenuNode(p, node, "menu"); });
    sep();
  }

  var fields = TAG_FIELDS[node.kind] || [];
  if (fields.indexOf("text") >= 0) {
    add("Edit label", function() {
      setTimeout(function() {
        var f = document.getElementById("f_text");
        if (f) { f.focus(); f.select(); }
      }, 0);
    });
  }
  add("Duplicate", function() { selPath = p.slice(); multiSel = []; duplicateSelected(); });

  var parentPath = p.slice(0, -1);
  var siblings = parentPath.length === 0 ? root.children : byPath(parentPath, root.children).children;
  var idx = p.length ? p[p.length - 1] : -1;
  if (idx > 0) add("Move up", function() { moveNodeAt(p, -1); });
  if (idx >= 0 && idx < siblings.length - 1) add("Move down", function() { moveNodeAt(p, 1); });

  sep();
  add("Delete", function() { selPath = p.slice(); multiSel = []; deleteSelected(); });
  return out;
}

// insertMenuNode adds a child to a menubar/menu, or a sibling after a
// menuitem. kind "menu" creates a nested submenu.
function insertModalButton(p, node) {
  withUndo(function() {
    var next = newModalButtonNode("Button");
    var arr, idx, newPath;
    if (node.kind === "modal") {
      arr = node.children;
      idx = arr.length;
      arr.splice(idx, 0, next);
      newPath = p.concat([idx]);
    } else {
      var parentPath = p.slice(0, -1);
      arr = parentPath.length === 0 ? root.children : byPath(parentPath, root.children).children;
      idx = p[p.length - 1] + 1;
      arr.splice(idx, 0, next);
      newPath = parentPath.concat([idx]);
    }
    selPath = newPath;
    multiSel = [];
  }, "add modal button");
}

function insertTabNode(p, node) {
  withUndo(function() {
    var next = newTabNode("Tab");
    var arr, idx, newPath;
    if (node.kind === "tabs") {
      arr = node.children;
      idx = arr.length;
      arr.splice(idx, 0, next);
      newPath = p.concat([idx]);
    } else {
      var parentPath = p.slice(0, -1);
      arr = parentPath.length === 0 ? root.children : byPath(parentPath, root.children).children;
      idx = p[p.length - 1] + 1;
      arr.splice(idx, 0, next);
      newPath = parentPath.concat([idx]);
    }
    selPath = newPath;
    multiSel = [];
  }, "add tab");
}

function insertAccordionItem(p, node) {
  withUndo(function() {
    var next = newAccordionItemNode("Section");
    var arr, idx, newPath;
    if (node.kind === "accordion") {
      arr = node.children;
      idx = arr.length;
      arr.splice(idx, 0, next);
      newPath = p.concat([idx]);
    } else {
      var parentPath = p.slice(0, -1);
      arr = parentPath.length === 0 ? root.children : byPath(parentPath, root.children).children;
      idx = p[p.length - 1] + 1;
      arr.splice(idx, 0, next);
      newPath = parentPath.concat([idx]);
    }
    selPath = newPath;
    multiSel = [];
  }, "add accordion item");
}

function insertMenuNode(p, node, kind) {
  withUndo(function() {
    var arr, idx, newPath;
    var next = kind === "menu"
      ? newMenuNode(node.kind === "menubar" ? "Menu" : "Submenu", node.kind === "menubar" ? ["Item 1", "Item 2"] : ["Item"])
      : newNodeForKind("menuitem");
    if (node.kind === "menubar" || node.kind === "menu") {
      arr = node.children;
      idx = arr.length;
      arr.splice(idx, 0, next);
      newPath = p.concat([idx]);
    } else {
      var parentPath = p.slice(0, -1);
      arr = parentPath.length === 0 ? root.children : byPath(parentPath, root.children).children;
      idx = p[p.length - 1] + 1;
      arr.splice(idx, 0, next);
      newPath = parentPath.concat([idx]);
    }
    selPath = newPath;
    multiSel = [];
  }, kind === "menu" ? "add submenu" : "add menu item");
}

function moveNodeAt(p, delta) {
  if (!p.length) return;
  var parentPath = p.slice(0, -1);
  var arr = parentPath.length === 0 ? root.children : byPath(parentPath, root.children).children;
  var idx = p[p.length - 1];
  var to = idx + delta;
  if (to < 0 || to >= arr.length) return;
  withUndo(function() {
    var node = arr.splice(idx, 1)[0];
    arr.splice(to, 0, node);
    selPath = parentPath.concat([to]);
    multiSel = [];
  }, "move element");
}

function samePath(a, b) {
  if (!a || !b || a.length !== b.length) return false;
  for (var i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}
function sameArray(a, b) {
  if (!a || !b) return false;
  if (a.length !== b.length) return false;
  for (var i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}
function multiHas(p) { for (var i = 0; i < multiSel.length; i++) if (samePath(multiSel[i], p)) return true; return false; }
function multiAdd(p) { if (!multiHas(p)) multiSel.push(p.slice()); }
function multiRemove(p) { multiSel = multiSel.filter(function(q){ return !samePath(q, p); }); }
function flattenPaths() {
  var out = [];
  (function walk(nodes, path) {
    for (var i = 0; i < nodes.length; i++) {
      var p = path.concat([i]);
      out.push(p);
      if (nodes[i].children && nodes[i].children.length) walk(nodes[i].children, p);
    }
  })(root.children, []);
  return out;
}
function indexOfPath(list, p) { for (var i = 0; i < list.length; i++) if (samePath(list[i], p)) return i; return -1; }
function rangeSelect(from, to) {
  var all = flattenPaths();
  var fi = indexOfPath(all, from), ti = indexOfPath(all, to);
  if (fi < 0 || ti < 0) { multiSel = [to.slice()]; selPath = to; return; }
  var lo = Math.min(fi, ti), hi = Math.max(fi, ti);
  multiSel = all.slice(lo, hi + 1).map(function(p){ return p.slice(); });
  selPath = to.slice();
}
function sortByDfs(paths) {
  var order = {};
  var all = flattenPaths();
  for (var i = 0; i < all.length; i++) order[all[i].join(",")] = i;
  return paths.slice().sort(function(a, b){ var ai = order[a.join(",")], bi = order[b.join(",")]; if (ai === undefined) ai = 1e9; if (bi === undefined) bi = 1e9; return ai - bi; });
}
function normalizeSelection(paths) {
  var out = [];
  paths.forEach(function(p) {
    var under = false;
    for (var i = 0; i < out.length; i++) if (isDescendant(out[i], p)) { under = true; break; }
    if (under) return;
    out = out.filter(function(q){ return !isDescendant(p, q); });
    out.push(p.slice());
  });
  return sortByDfs(out);
}
function selectedPaths() {
  var base = multiSel.slice();
  if (selPath && !base.some(function(q){ return samePath(q, selPath); })) base.push(selPath);
  return normalizeSelection(base);
}
function deletePaths(paths) {
  var sorted = paths.slice().sort(function(a, b) {
    if (a.length !== b.length) return b.length - a.length;
    for (var i = a.length - 1; i >= 0; i--) if (a[i] !== b[i]) return b[i] - a[i];
    return 0;
  });
  withUndo(function() {
    for (var i = 0; i < sorted.length; i++) {
      var p = sorted[i];
      byPath(p.slice(0, p.length - 1), root.children).children.splice(p[p.length - 1], 1);
    }
    selPath = null; multiSel = [];
  }, "delete elements");
}
function copySelected() {
  var paths = selectedPaths();
  if (!paths.length) { setStatus("nothing selected"); return false; }
  clipboard = paths.map(function(p){ return JSON.parse(JSON.stringify(byPath(p, root.children))); });
  setStatus("copied " + clipboard.length + " element(s)");
  return true;
}
function cutSelected() {
  var paths = selectedPaths();
  if (!paths.length) { setStatus("nothing selected"); return; }
  clipboard = paths.map(function(p){ return JSON.parse(JSON.stringify(byPath(p, root.children))); });
  deletePaths(paths);
}
function pasteClipboard() {
  if (!clipboard.length) { setStatus("clipboard empty"); return; }
  var clones = clipboard.map(function(c){ return JSON.parse(JSON.stringify(c)); });
  var idx, container;
  if (selPath) {
    idx = selPath[selPath.length - 1] + 1;
    container = byPath(selPath.slice(0, -1), root.children).children;
  } else {
    container = root.children;
    idx = container.length;
  }
  withUndo(function() {
    for (var i = 0; i < clones.length; i++) container.splice(idx + i, 0, clones[i]);
  }, "paste");
  setStatus("pasted " + clones.length + " element(s)");
}
function delSelPath(pathStr) {
  var p = pathStr.split(",").map(Number);
  withUndo(function() {
    dbg("delete path=[" + p.join(",") + "] before=" + root.children.length);
    byPath(p.slice(0, p.length - 1), root.children).children.splice(p[p.length - 1], 1);
    if (samePath(selPath, p)) selPath = null;
  }, "delete element");
}
function dropAt(srcStr, tray, index, targetNode, targetPath) {
  var src = srcStr.split(",").map(Number);
  if (!src.length) return;
  // no-op when dropping on itself or inside its own subtree
  if (isDescendant(src, targetPath)) return;
  withUndo(function() {
    var srcPass = JSON.parse(JSON.stringify(byPath(src, root.children)));
    var srcParent = src.slice(0, -1);
    var srcIdx = src[src.length - 1];
    byPath(srcParent, root.children).children.splice(srcIdx, 1);
    if (targetNode && canNest(targetNode.kind)) {
      // nest inside the target container
      targetNode.children.push(srcPass);
    } else {
      // reorder as a sibling, right before the target
      if (sameArray(srcParent, pathOf(tray)) && srcIdx < index) index = index - 1;
      tray.splice(index, 0, srcPass);
    }
    selPath = null;
  }, "move element");
}
function isDescendant(ancestor, maybe) {
  if (maybe.length < ancestor.length) return false;
  for (var i = 0; i < ancestor.length; i++) if (ancestor[i] !== maybe[i]) return false;
  return true;
}
function pathOf(tray) {
  if (tray === root.children) return [];
  var found = null;
  (function walk(children, path) {
    for (var i = 0; i < children.length && !found; i++) {
      if (children[i].children === tray) { found = path.concat([i]); return; }
      if (children[i].children) walk(children[i].children, path.concat([i]));
    }
  })(root.children, []);
  return found || [];
}

// KIND_PLACEHOLDER gives text-bearing elements a visible default label when
// inserted from the palette. Without it an empty button/code/badge is invisible
// in the preview (an empty dark pill on a dark page) and looks like the insert
// did nothing.
var KIND_PLACEHOLDER = {
  button: "Button", h1: "Title", h2: "Title 2", h3: "Title 3",
  h4: "Title 4", h5: "Title 5", h6: "Title 6", p: "Paragraph",
  code: "code", mark: "highlight", strong: "bold text", em: "italic text",
  small: "small text", legend: "Legend", badge: "badge"
};
function addEl(kind) {
  withUndo(function() {
    var node = newNodeForKind(kind);
    if (KIND_PLACEHOLDER[kind] && node.text === "") node.text = KIND_PLACEHOLDER[kind];
    var insPath = null;
    if (selPath) {
      var sel = byPath(selPath, root.children);
      if (canNest(sel.kind)) {
        sel.children.push(node);
        insPath = selPath.concat([sel.children.length - 1]);
      } else {
        // leaf selected: insert as the next sibling
        var par = selPath.slice(0, -1);
        var arr = par.length === 0 ? root.children : byPath(par, root.children).children;
        var pos = selPath[selPath.length - 1] + 1;
        arr.splice(pos, 0, node);
        insPath = par.concat([pos]);
      }
    } else {
      // no selection: append at the root (inside the virtual body)
      root.children.push(node);
      insPath = [root.children.length - 1];
    }
    // Tables are wrapped in a scroll container so a fixed-size parent shows
    // scrollbars only when rows/columns are actually hidden.
    selPath = (kind === "table") ? wrapTableInScroll(insPath) : insPath;
  }, "insert " + kind);
}

function upd(field, value) {
  var paths = selectedPaths();
  if (!paths.length) return;
  snap("set " + field);
  paths.forEach(function(p) {
    var node = byPath(p, root.children);
    node[field] = value;
    if (field === "src") {
      var sv = String(node[field] || "").trim();
      var looksURL = /^(?:[a-z][a-z0-9+.-]*:|\/|\.{1,2}\/|#)/i.test(sv);
      node.srcExpr = !looksURL && (/^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$/.test(sv) || /[?{}]|\+/.test(sv));
    }
    if (field === "onClick" && IDENT.test(node[field]) && !RESERVED_NAME.test(node[field]) && !hasFn(node[field])) root.funcs.push({ name: node[field], params: "", ret: "null" });
    if ((field === "dropFiles" || field === "pickFiles") && IDENT.test(node[field]) && !RESERVED_NAME.test(node[field]) && !hasState(node[field])) root.states.push({ name: node[field], init: "\"\"" });
    if (field === "value" && IDENT.test(node[field]) && !RESERVED_NAME.test(node[field]) && !hasState(node[field])) root.states.push({ name: node[field], init: "\"\"" });
    if (field === "checked" && IDENT.test(node[field]) && !RESERVED_NAME.test(node[field]) && !hasState(node[field])) root.states.push({ name: node[field], init: "false" });
  });
  renderTree(); renderView(); setOut(); autoPreview();
}
function updView(field) {
  snap();
  if (field === "onLoad") curPage().onLoad = (document.getElementById("f_onload").value || "").trim();
  else curPage().title = document.getElementById("f_title").value;
  root.title = curPage().title;
  setOut(); renderPages(); autoPreview();
}
function updStateName(s, val) { snap(); s.name = val; setOut(); autoPreview(); }
function updStateInit(s, val) { snap(); s.init = val; setOut(); autoPreview(); }
function updFnName(f, val) { snap(); f.name = val; setOut(); autoPreview(); }
function updFnParams(f, val) { snap(); f.params = val; setOut(); autoPreview(); }
function updFnRet(f, val) { snap(); f.ret = val; setOut(); autoPreview(); }

function fillPanel() {
  var box = document.getElementById("sec-elem");
  box.innerHTML = "";
  if (!selPath) { clearFields(); return; }
  var n = byPath(selPath, root.children);
  // Element properties moved to right inspector for better UX
  document.getElementById("sec-elem").style.display = "none";
  document.getElementById("elemhead").style.display = "none";
  postHighlight();
}
function onEventSelect(nodes, evType) {
  evType = evType || "onClick";
  var s = document.createElement("select");
  s.id = "f_" + evType;
  var cur = commonVal(nodes, evType);
  var mixed = nodes.length > 1 && nodes.some(function(m){ return (m[evType] || "") !== ""; }) && !cur;
  var empty = document.createElement("option"); empty.value = ""; empty.textContent = mixed ? "(valori misti)" : "(no handler)"; s.appendChild(empty);
  if (root.funcs) {
    root.funcs.forEach(function(fn) {
      if (!fn.name) return;
      var o = document.createElement("option"); o.value = fn.name; o.textContent = fn.name;
      if (cur === fn.name) o.selected = true;
      s.appendChild(o);
    });
  }
  var add = document.createElement("option"); add.value = "__new__"; add.textContent = "+ new function\u2026"; s.appendChild(add);
  s.onchange = function(ev) {
    if (ev.target.value === "__new__") {
      askPrompt("New function", "Name", "handler", function(raw) {
        var name = raw.replace(/[^A-Za-z0-9_]/g, "");
        if (name && IDENT.test(name) && !RESERVED_NAME.test(name)) {
          snap();
          if (!hasFn(name)) { if (!root.funcs) root.funcs = []; root.funcs.push({ name: name, params: "", ret: "null" }); }
          nodes.forEach(function(m){ m[evType] = name; });
          renderAll(); autoPreview();
          return;
        }
      });
      s.value = cur;
      return;
    }
    snap();
    if (ev.target.value === "") {
      nodes.forEach(function(m){ delete m[evType]; });
      renderAll(); autoPreview(); return;
    }
    upd(evType, ev.target.value);
  };
  return s;
}
function clearFields() {
  document.getElementById("sec-elem").innerHTML = "";
  document.getElementById("sec-elem").style.display = "none";
  document.getElementById("elemhead").style.display = "none";
  postHighlight();
}

// ---- view model -----------------------------------------------------------
function hasState(name) { return root.states.some(function(s) { return s.name === name; }); }
function hasFn(name) { return root.funcs.some(function(f) { return f.name === name; }); }
function renderView() {
  var sl = document.getElementById("stateslist"); if (sl) sl.innerHTML = ""; var sl2 = document.getElementById("stateslist2"); if (sl2) sl2.innerHTML = ""; if (sl && !sl2) sl2 = sl; if (!sl) sl = sl2; if (!sl2) sl2 = sl;
  if (!sl) sl = document.getElementById("stateslist2");
  if (sl) {
    sl.innerHTML = "";
    // std imports: preserved across designer edits and emitted first in the
    // generated SUI so http/json/exec (or table) calls keep compiling.
    var impRow = document.createElement("div"); impRow.className = "rowline";
    var impLbl = document.createElement("label"); impLbl.textContent = "imports";
    impLbl.title = "Comma-separated std modules, e.g. http, json, exec";
    impRow.appendChild(impLbl);
    var impInp = document.createElement("input");
    impInp.value = (root.imports || []).join(", ");
    impInp.placeholder = "http, json, exec";
    impInp.spellcheck = false;
    impInp.onchange = function(){ withUndo(function(){ root.imports = parseStdImports(impInp.value); }); };
    impRow.appendChild(impInp);
    sl.appendChild(impRow);
  }
  root.states.forEach(function(s, i) {
    var row = document.createElement("div"); row.className = "rowline";
    var ni = mkInp(s.name); ni.oninput = function(ev){ updStateName(s, ev.target.value); }; row.appendChild(ni);
    var ii = mkInp(s.init, "init"); ii.oninput = function(ev){ updStateInit(s, ev.target.value); }; row.appendChild(ii);
    row.appendChild(mkDel(function(){ withUndo(function(){ root.states.splice(i, 1); }); }));
    if (sl) sl.appendChild(row); if (sl2 && sl2 !== sl) sl2.appendChild(row);
  });
  var fl = document.getElementById("fnslist"); if (fl) fl.innerHTML = ""; var fl2 = document.getElementById("fnslist2"); if (fl2) fl2.innerHTML = ""; if (fl && !fl2) fl2 = fl; if (!fl) fl = fl2; if (!fl2) fl2 = fl;
  if (!fl) fl = document.getElementById("fnslist2");
  if (fl) {
    fl.innerHTML = "";
  }
  root.funcs.forEach(function(f, i) {
    var row = document.createElement("div"); row.className = "rowline";
    var ni = mkInp(f.name); ni.oninput = function(ev){ updFnName(f, ev.target.value); }; row.appendChild(ni);
    var pi = mkInp(f.params || "", "params"); pi.oninput = function(ev){ updFnParams(f, ev.target.value); }; row.appendChild(pi);
    var ri = mkInp(f.ret || "", "ret"); ri.oninput = function(ev){ updFnRet(f, ev.target.value); }; row.appendChild(ri);
    var rl = document.createElement("label"); rl.textContent = "ret"; row.appendChild(rl);
    row.appendChild(mkDel(function(){ withUndo(function(){ root.funcs.splice(i, 1); }); }));
    if (fl) fl.appendChild(row); if (fl2 && fl2 !== fl) fl2.appendChild(row);
  });
}
function mkInp(value, ph) { var i = document.createElement("input"); i.value = value; if (ph) i.placeholder = ph; return i; }
function parseStdImports(v) {
  var out = [], seen = {};
  (v || "").split(/[\s,]+/).forEach(function(x){
    x = (x || "").trim();
    if (!x) return;
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(x)) return;
    if (seen[x]) return;
    seen[x] = true; out.push(x);
  });
  return out;
}
function mkDel(fn) { var d = document.createElement("span"); d.className = "btn"; d.style.cssText = "padding:1px 5px;font-size:10px"; d.textContent = "\u2212"; d.onclick = fn; return d; }
function addState() { withUndo(function(){ root.states.push({ name: "s" + (root.states.length + 1), init: "" }); }); }
function addFn() { withUndo(function(){ root.funcs.push({ name: "handler" + (root.funcs.length + 1), params: "", ret: "null" }); }); }

// ---- CSS engine (two-way binding) ----------------------------------------
function updateApply() {
  var btn = document.querySelector('[data-apply="sui"]');
  if (!btn) return;
  var dirty = document.getElementById("out").value !== suiSource;
  btn.style.display = dirty ? '' : 'none';
  btn.classList.toggle("applyhot", dirty);
}
function buildDatalist() {
  var dl = document.getElementById("classlist"); if (!dl) return;
  dl.innerHTML = "";
  var re = /\.([A-Za-z_][A-Za-z0-9_-]*)\s*\{/g, seen = {}, m;
  while ((m = re.exec(root.css)) !== null) {
    if (m[1] === "suitbl" || seen[m[1]]) continue; // designer-managed base class: not user-selectable
    seen[m[1]] = true; var o = document.createElement("option"); o.value = m[1]; dl.appendChild(o);
  }
}
function scheduleCssSync() {
  if (cssTimer) clearTimeout(cssTimer);
  cssTimer = setTimeout(cssSync, 200);
}
function cssSync() {
  var fd = new URLSearchParams();
  fd.set("css", root.css);
  fd.set("used", JSON.stringify(usedClassesList()));
  fetch("/__css__/parse", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) return;
    ast = j.ast || [];
    tokens = j.tokens || [];
    problems = j.problems || [];
    cssGen = j.generated || root.css;
    renderProblems();
    renderTokens();
    renderTree();
    renderInspector();
    var d = j.diagnostics || [];
    setStatus(d.length ? d.length + " css issue(s)" : "synced");
  });
}
function usedClassesList() {
  var out = [];
  function walk(nodes) { nodes.forEach(function(n){ if (n.class) n.class.trim().split(/\s+/).forEach(function(c){ if (c && out.indexOf(c)<0) out.push(c); }); if (n.children) walk(n.children); }); }
  (root.pages || []).forEach(function(p){ walk(p.children || []); });
  return out;
}

// Skin class names the compiled app injects automatically or uses for its own
// chrome. The renderer prepends the built-in class to native controls (an
// input always renders class="sui-checkbox <userclass>"), so handing one of
// these names to a user element would make its style rules leak into EVERY
// same-kind control, even plain ones with no class attribute.
var SKIN_CLASSES = ["sui-input", "sui-textarea", "sui-select", "sui-checkbox", "sui-radio", "sui-switch", "sui-slider", "sui-progress", "sui-badge", "sui-alert", "sui-spinner", "sui-card", "sui-divider", "sui-tabs", "sui-tab", "sui-accordion", "sui-accordion-item", "sui-accordion-head", "sui-accordion-panel", "sui-toast", "sui-toast-host", "sui-toast-fixed", "sui-toast-info", "sui-toast-success", "sui-toast-warning", "sui-toast-error", "sui-tooltip-bubble", "sui-combobox", "sui-combobox-label", "sui-combobox-control", "sui-combobox-input", "sui-combobox-toggle", "sui-combobox-clear", "sui-combobox-value", "sui-combobox-popup", "sui-combobox-option", "sui-combobox-empty", "sui-combobox-error", "sui-datepicker", "sui-datepicker-label", "sui-datepicker-control", "sui-datepicker-input", "sui-datepicker-clear", "sui-datepicker-value", "sui-datepicker-error", "sui-tree", "sui-tree-title", "sui-tree-group", "sui-tree-item", "sui-tree-row", "sui-tree-toggle", "sui-tree-spacer", "sui-tree-label", "sui-tree-value", "sui-chart", "sui-chart-head", "sui-chart-title", "sui-chart-zoom", "sui-chart-zoom-btn", "sui-chart-zoom-label", "sui-chart-zoom-reset", "sui-chart-viewport", "sui-chart-stage", "sui-chart-svg", "sui-chart-grid", "sui-chart-axis", "sui-chart-axis-label", "sui-chart-value-label", "sui-chart-bar", "sui-chart-line", "sui-chart-area", "sui-chart-point", "sui-chart-slice", "sui-chart-pie-hole", "sui-chart-legend", "sui-chart-legend-item", "sui-chart-swatch", "sui-chart-error", "sui-view", "sui-hl", "sui-drag-over", "sui-rsz", "sui-grid", "sui-ac", "sui-ac-item", "suitbl"];

// assignedClassNames collects the classes already claimed by any element on
// any page, so two controls never share an auto-generated class (a shared
// class would style them together).
function assignedClassNames() {
  var used = {};
  function walk(nodes) {
    (nodes || []).forEach(function(n){
      if (n.class) n.class.trim().split(/\s+/).forEach(function(c){ if (c) used[c] = 1; });
      if (n.children) walk(n.children);
    });
  }
  (root.pages || []).forEach(function(p){ walk(p.children || []); });
  return used;
}

// set a CSS property on the selected class (visual edit -> text).
function ensureClass(n) {
  if (n.class && n.class.trim()) {
    // "suitbl" is the designer-managed table base hook: never return it as
    // the user's editable class, and skip it when picking the first name.
    var cs = n.class.trim().split(/\s+/).filter(function(c) { return c && c !== "suitbl"; });
    if (cs.length) return cs[0];
    n.class = "";
  }
  var base = ((n.text || n.kind) + "").toLowerCase().replace(/[^a-z0-9_-]+/g, "").replace(/^[^a-z_]+/, "") || "el";
  var used = assignedClassNames();
  var taken = function(c){ return used[c] || SKIN_CLASSES.indexOf(c) >= 0 || root.css.indexOf("." + c + " ") >= 0 || root.css.indexOf("." + c + "{") >= 0; };
  var cls = base, i = 2;
  while (taken(cls)) { cls = base + "-" + (i++); }
  n.class = (n.kind === "table" ? "suitbl " : "") + cls;
  return cls;
}

function setCssProp(prop, value) {
  var obj = {}; obj[prop] = value; setCssProps(obj);
}

// Writes one or more CSS properties, expanding vendor prefixes as needed.
// Applies to every selected element (sequentially, to avoid racing writes).
function setCssProps(map) {
  var decls = [];
  for (var k in map) {
    compatDecls(k, map[k]).forEach(function(d){ decls.push(d); });
  }
  snap(snapLabelFromMap(map));
  if (selBody) { setCssOnClass("body" + selState, decls); return; }
  var paths = selectedPaths();
  if (!paths.length) return;
  var i = 0;
  (function next() {
    if (i >= paths.length) return;
    var n = byPath(paths[i], root.children);
    var cls = ensureClass(n) + selState;
    if (n.kind === "toast") {
      // Toast is a component: the inspector styles the card inside the host
      // rather than the empty host wrapper.
      cls = ensureClass(n) + " .sui-toast" + selState;
    }
    var useDecls = decls;
    // Tables: a height maps to an even per-row height (header included) so the
    // extra space is shared by every row instead of being dumped on one row.
    if (n.kind === "table" && n.rows && map["height"] !== undefined) {
      var hv = String(map["height"] === null ? "" : map["height"]).trim();
      useDecls = decls.filter(function(d){ return d.prop !== "height"; });
      if (hv === "") {
        useDecls = useDecls.concat([{ prop: "--t-rowh", value: "" }]);
      } else {
        var hpx = parseFloat(hv);
        if (!isNaN(hpx) && hpx > 0) useDecls = useDecls.concat([{ prop: "--t-rowh", value: Math.max(1, Math.round(hpx / n.rows)) + "px" }]);
      }
    }
    i++;
    setCssOnClass(cls, useDecls, next);
  })();
}
function snapLabelFromMap(map) {
  var keys = Object.keys(map).filter(function(k){ return map[k] !== ""; });
  if (!keys.length) return "clear style";
  return "set " + keys.join("/");
}
function setCssOnClass(cls, decls, after) {
  var fd = new URLSearchParams();
  fd.set("css", root.css);
  fd.set("class", cls);
  fd.set("decls", JSON.stringify(decls));
  fetch("/__css__/setbatch", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) return;
    syncing = true;
    root.css = j.css;
    ast = j.ast || ast;
    renderTree(); renderGutters(); setOut(); autoPreview(); renderInspector();
    syncing = false;
    scheduleCssSync();
    if (after) after();
  });
}

// ---- alignment toolbar -----------------------------------------------------
// Aligns the selected element inside its parent container. Nine buttons form a
// 3x3 square; each corresponding to a corner/edge/center position. The parent
// is made a flex container (if it is not already) and the element is placed
// with a mix of align-self and auto margins so both axes are respected.
var ALIGN_POINTS = [
  { dir: "tl", row: 1, col: 1, label: "Align top-left" },
  { dir: "tc", row: 1, col: 2, label: "Align top-center" },
  { dir: "tr", row: 1, col: 3, label: "Align top-right" },
  { dir: "cl", row: 2, col: 1, label: "Align center-left" },
  { dir: "cc", row: 2, col: 2, label: "Center" },
  { dir: "cr", row: 2, col: 3, label: "Align center-right" },
  { dir: "bl", row: 3, col: 1, label: "Align bottom-left" },
  { dir: "bc", row: 3, col: 2, label: "Align bottom-center" },
  { dir: "br", row: 3, col: 3, label: "Align bottom-right" }
];
function alignElement(dir) {
  if (!selPath) { setStatus("nothing selected"); return; }
  var parent = selPath.length >= 2 ? byPath(selPath.slice(0, -1), root.children) : null;
  // row => main axis horizontal; col/generic/body => main axis vertical.
  var mainH = parent && parent.kind === "row";
  var h = dir.charAt(1); // 'l', 'c', 'r'
  var v = dir.charAt(0); // 't', 'c', 'b'
  var map = {};
  // Horizontal component: main axis uses auto margins, cross axis uses align-self.
  if (mainH) {
    map["margin-left"] = h === "r" ? "auto" : "0";
    map["margin-right"] = h === "l" ? "auto" : "0";
    if (h === "c") { map["margin-left"] = "auto"; map["margin-right"] = "auto"; }
  } else {
    map["align-self"] = h === "l" ? "flex-start" : h === "r" ? "flex-end" : "center";
  }
  // Vertical component.
  if (mainH) {
    map["align-self"] = v === "t" ? "flex-start" : v === "b" ? "flex-end" : "center";
  } else {
    map["margin-top"] = v === "b" ? "auto" : "0";
    map["margin-bottom"] = v === "t" ? "auto" : "0";
    if (v === "c") { map["margin-top"] = "auto"; map["margin-bottom"] = "auto"; }
  }
  function applyChild() { setCssProps(map); }
  var needFlex = !parent || (parent.kind !== "row" && parent.kind !== "col");
  if (needFlex) {
    // Top-level elements are children of the page view (.sui-view), not of
    // <body>: making body flex does not affect them, the page view is the box
    // that must become a flex column for align-self / auto margins to work.
    var flexCls = parent ? ensureClass(parent) : "sui-view";
    setCssOnClass(flexCls, [{ prop: "display", value: "flex" }, { prop: "flex-direction", value: "column" }], applyChild);
  } else {
    applyChild();
  }
}
function alignClear() {
  if (!selPath) { setStatus("nothing selected"); return; }
  setCssProps({ "margin-left": "", "margin-right": "", "margin-top": "", "margin-bottom": "", "align-self": "" });
}
function alignControls() {
  var a = document.createElement("div"); a.className = "acc open";
  var h = document.createElement("div"); h.className = "acchead";
  h.innerHTML = '<span>Alignment</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div"); b.className = "accbody";
  if (!selPath) {
    var hint = document.createElement("div"); hint.className = "hint";
    hint.textContent = "Align an element inside its parent container.";
    b.appendChild(hint);
  } else {
    var grid = document.createElement("div");
    grid.style.cssText = "display:grid;grid-template-columns:repeat(3,26px);grid-template-rows:repeat(3,26px);gap:3px;justify-content:center";
    ALIGN_POINTS.forEach(function(p) {
      var cell = document.createElement("button");
      cell.className = "btn";
      cell.style.cssText = "padding:0;display:flex;align-items:center;justify-content:center;font-size:8px;color:var(--dim);grid-row:" + p.row + ";grid-column:" + p.col;
      cell.title = p.label;
      cell.textContent = "\u25a0";
      cell.onclick = function(){ alignElement(p.dir); };
      grid.appendChild(cell);
    });
    b.appendChild(grid);
    var btns = document.createElement("div"); btns.style.cssText = "display:flex;gap:4px;margin-top:4px";
    var cls = document.createElement("button"); cls.className = "btn"; cls.style.cssText = "font-size:11px;padding:4px 8px"; cls.textContent = "\u2715 Clear";
    cls.onclick = function(){ alignClear(); };
    btns.appendChild(cls);
    b.appendChild(btns);
  }
  a.appendChild(b);
  return a;
}

// ---- inspector ------------------------------------------------------------
function selClass() {
  if (selBody) return "body" + selState;
  if (!selPath) return "";
  var n = byPath(selPath, root.children);
  var all = n.class ? n.class.trim().split(/\s+/).filter(Boolean) : [];
  var usable = all.filter(function(c){ return c !== "suitbl"; });
  var cls = "";
  for (var i = 0; i < usable.length; i++) {
    if (findRule(usable[i] + selState)) { cls = usable[i]; break; }
  }
  if (!cls) cls = usable[0] || all[0] || "";
  return cls + selState;
}
function selectorMatchesSel(s, sel) {
  if (!s || !sel) return false;
  if (s === sel || s === "." + sel) return true;
  var state = "", base = sel;
  var sm = /:[A-Za-z-]+/.exec(sel);
  if (sm) { state = sm[0]; base = sel.slice(0, sm.index); }
  function escRe(x) { return x.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"); }
  if (base === "body") {
    if (s.indexOf("body") < 0) return false;
    return !state || s.indexOf(state) >= 0;
  }
  var re = new RegExp("(^|[^A-Za-z0-9_-])\\." + escRe(base) + "(?![A-Za-z0-9_-])");
  if (!re.test(s)) return false;
  if (state) return s.indexOf(state) >= 0;
  return s.indexOf(":") < 0;
}
function findRule(sel) {
  if (!sel) return null;
  for (var i = 0; i < ast.length; i++) {
    var r = ast[i];
    if (r.kind === "rule") {
      for (var j = 0; j < r.selectors.length; j++) {
        if (selectorMatchesSel(r.selectors[j].trim(), sel)) return r;
      }
    } else if (r.nodes) {
      var sub = findRuleIn(r.nodes, sel);
      if (sub) return sub;
    }
  }
  return null;
}
function findRuleIn(nodes, sel) {
  for (var i = 0; i < nodes.length; i++) {
    var r = nodes[i];
    if (r.kind === "rule") {
      for (var j = 0; j < r.selectors.length; j++) { if (selectorMatchesSel(r.selectors[j].trim(), sel)) return r; }
    } else if (r.nodes) { var sub = findRuleIn(r.nodes, sel); if (sub) return sub; }
  }
  return null;
}
// Exact selector lookup is needed for part-scoped toast rules such as
// ".demo-toast .sui-toast-icon", which are not a single class.
function findRuleBySelector(sel) {
  var want = String(sel || "").trim();
  if (!want) return null;
  function walk(nodes) {
    for (var i = 0; i < nodes.length; i++) {
      var r = nodes[i];
      if (r.kind === "rule") {
        for (var j = 0; j < r.selectors.length; j++) {
          if (r.selectors[j].trim() === want) return r;
        }
      } else if (r.nodes) {
        var sub = walk(r.nodes);
        if (sub) return sub;
      }
    }
    return null;
  }
  return walk(ast);
}
function declOf(rule, prop) {
  if (!rule) return "";
  for (var i = 0; i < rule.decls.length; i++) if (rule.decls[i].prop === prop) return rule.decls[i].value;
  return "";
}

// Container-target rows: justify-content, align-items, flex-direction, gap and
// align-content only act on flex/grid CONTAINERS, so when the selection is a
// plain element (a button dropped on the page, a text run...) these rows are
// redirected to the parent container — or to the page view (.sui-view) for
// top-level elements — where they actually lay out children. The chip
// "→ parent" marks a redirected row.
function selParentCls() {
  var paths = selectedPaths();
  if (!paths.length) return "";
  var pnode = paths[0].length >= 2 ? byPath(paths[0].slice(0, -1), root.children) : null;
  return pnode ? ensureClass(pnode) : "sui-view";
}
function selIsContainer(nodes) {
  return selBody || (nodes.length > 0 && everyKind(nodes, function(k){ return k === "row" || k === "col" || k === "card"; }));
}
function applyToParent(prop, val) {
  var cls = selParentCls();
  if (!cls) return;
  snap("parent set " + prop);
  setCssOnClass(cls, compatDecls(prop, val));
  setStatus(prop + " applied to parent container " + cls);
}
function parentChip(d) {
  var chip = document.createElement("span"); chip.textContent = "\u2192 parent";
  chip.title = "Applied to the parent container (this property only lays out children)";
  chip.style.cssText = "font-size:10px;color:var(--accent);flex:none";
  d.appendChild(chip);
}
function containerSelBox(prop, label, opts, nodes) {
  if (selIsContainer(nodes)) return selBox(prop, label, opts);
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var s = document.createElement("select");
  var cur = cssFieldValue(prop, findRule(selParentCls()), true);
  var list = opts.slice();
  if (cur && list.indexOf(cur) < 0) list.push(cur);
  list.forEach(function(o){ var e = document.createElement("option"); e.value = o; e.textContent = o === "" ? "(unset)" : o; if (o === cur) e.selected = true; s.appendChild(e); });
  s.onchange = function(){ applyToParent(prop, s.value); };
  d.appendChild(s);
  parentChip(d);
  return d;
}
function containerTxt(prop, label, ph, nodes) {
  if (selIsContainer(nodes)) return txt(prop, label, ph);
  var d = document.createElement("div"); d.className = "rowline";
  d.appendChild(lbl(label));
  var i = document.createElement("input"); i.value = cssFieldValue(prop, findRule(selParentCls()), true);
  i.placeholder = ph || "auto";
  i.onchange = function(){ applyToParent(prop, normLen(i.value)); };
  d.appendChild(i);
  parentChip(d);
  return d;
}
function positionHasEffect() {
  var cur = cssFieldValue("position", currentRule());
  return cur !== "" && cur !== "static";
}

// Effective-style cache. The preview iframe reports the authored and computed
// CSS values of the selected element, so the inspector also shows values that
// come from inherited rules, compound selectors, the table base skin or inline
// styles rather than only from the selected class rule.
var LIVE_STYLE_PROPS = [
  "display", "flex-direction", "justify-content", "align-items", "align-self", "align-content",
  "vertical-align", "gap", "grid-template-columns", "grid-template-rows", "grid-column",
  "grid-row", "grid-area", "place-items", "place-content", "position", "top", "right",
  "bottom", "left", "z-index", "margin", "padding", "border-style", "border-width",
  "border-color", "border-radius", "font-size", "font-family", "color", "font-weight",
  "font-style", "text-transform", "text-decoration", "text-align", "line-height",
  "letter-spacing", "word-spacing", "text-indent", "text-shadow", "white-space",
  "background-color", "opacity", "box-shadow", "width", "height", "max-width"
];
var __liveStyle = { key: "", id: -1, values: {}, parentValues: {}, pending: false };
var __runColorHint = "";
function commonRunColor(n) {
  var c = null;
  (n && n.runs ? n.runs : []).forEach(function(r) {
    if (!r || !r.text) return;
    if (c === null) c = r.color || "";
    else if (c !== (r.color || "")) c = "";
  });
  return c || "";
}
function liveStyleKey() {
  if (selBody) return "body|" + selState;
  if (!selPath) return "";
  return "p" + selPath.join(",") + "|" + (selCell ? ("c" + selCell.r + "," + selCell.c) : "") +
    "|" + (selRow !== null ? ("r" + selRow) : "") + "|" + (selCol !== null ? ("k" + selCol) : "") + "|" + selState;
}
function requestLiveStyles(force) {
  // Toast style rows read the authored part-scoped rules; the preview bridge
  // exposes computed styles for the host wrapper only, not its internal parts.
  if (toastHostFromSelection()) {
    __liveStyle = { key: "", id: -1, values: {}, parentValues: {}, pending: false };
    return;
  }
  var key = liveStyleKey();
  if (!key) { __liveStyle = { key: "", id: -1, values: {}, parentValues: {}, pending: false }; return; }
  if (!force && __liveStyle.key === key) return;
  var ids = currentPeerIds();
  var id = selBody ? -1 : (ids.length ? ids[0] : -2);
  if (id === -2) return;
  var f = document.getElementById("prev");
  __liveStyle = { key: key, id: id, values: {}, parentValues: {}, pending: true };
  if (f && f.contentWindow) {
    f.contentWindow.postMessage({ ty: "sui-style-req", key: key, id: id, props: LIVE_STYLE_PROPS }, "*");
  }
}
function cssColorToHex(v) {
  v = String(v || "").trim();
  if (/^#[0-9a-f]{3,8}$/i.test(v)) return v;
  if (!v || v === "transparent" || v === "none" || v === "inherit" || v === "initial" || v === "unset") return "";
  var probe = document.createElement("span");
  probe.style.cssText = "position:absolute;left:-10000px;top:0;color:" + v;
  document.body.appendChild(probe);
  var rgb = getComputedStyle(probe).color;
  probe.remove();
  var m = /rgba?\(([^)]+)\)/.exec(rgb);
  if (!m) return "";
  var p = m[1].split(/[,\s\/]+/).map(function(x){ return parseFloat(x); });
  if (p.length < 3 || p[3] === 0) return "";
  function h2(n) { var s = Math.round(Math.max(0, Math.min(255, n))).toString(16); return s.length < 2 ? "0" + s : s; }
  return "#" + h2(p[0]) + h2(p[1]) + h2(p[2]);
}
function normalizeLiveField(prop, v) {
  if (/(^|-)color$/.test(prop) || prop === "background-color") return cssColorToHex(v);
  return String(v || "").trim();
}
function liveStyleValue(prop, parent) {
  if (!__liveStyle || __liveStyle.key !== liveStyleKey()) return "";
  var map = parent ? __liveStyle.parentValues : __liveStyle.values;
  var x = map && map[prop];
  if (!x) return "";
  var a = x.a || "", c = x.c || "";
  if (a && !/^var\(/i.test(a) && !/^(inherit|initial|unset)$/i.test(a)) return normalizeLiveField(prop, a);
  if (c) return normalizeLiveField(prop, c);
  return normalizeLiveField(prop, a);
}
function cssFieldValue(prop, rule, useParent) {
  if (prop === "color" && __runColorHint) return cssColorToHex(__runColorHint) || __runColorHint;
  var direct = rule ? declOf(rule, prop) : "";
  if (direct && !/^var\(/i.test(direct)) return normalizeLiveField(prop, direct) || direct;
  var live = liveStyleValue(prop, !!useParent);
  return live || direct;
}
function applyLiveStyleMessage(d) {
  if (!__liveStyle || d.key !== __liveStyle.key) return;
  __liveStyle.values = d.values || {};
  __liveStyle.parentValues = d.parentValues || {};
  __liveStyle.pending = false;
  renderInspector();
}
function renderInspector() {
  var box = document.getElementById("inspector");
  box.innerHTML = "";
  // Document section: the page title is a page-level property (the browser tab
  // title of the generated app), not an element property; it pins to the top of
  // the inspector and stays editable whatever is selected.
  var dsec = document.createElement("div"); dsec.className = "ins-sec";
  var dh = document.createElement("div"); dh.className = "ins-sec-h"; dh.textContent = "View";
  dsec.appendChild(dh);
  var dr = document.createElement("div"); dr.className = "rowline";
  var dl = document.createElement("label"); dl.textContent = "title"; dl.title = "Browser tab title in the generated app (<title>)";
  var di = document.createElement("input"); di.id = "f_title"; di.value = curPage().title || "";
  di.oninput = function(ev){ updView("title"); };
  dr.appendChild(dl); dr.appendChild(di);
  dsec.appendChild(dr);
  var pr = document.createElement("div"); pr.className = "rowline";
  var pl2 = document.createElement("label"); pl2.textContent = "page"; pl2.title = "Name of the active page (SUI view)";
  var pi = document.createElement("input"); pi.id = "f_page"; pi.value = curPage().name;
  pi.onchange = function(){ 
    var want = pi.value.trim();
    var cur = curPage();
    if (want === cur.name) { pi.value = curPage().name; return; }
    askPrompt("Rename page", "New name", want, function(name) {
      if (!/^[a-z_][a-z0-9_]*$/i.test(name)) { showMsg("Invalid name", "Use only letters, digits and _ , not starting with a digit."); return; }
      if (/^render_/.test(name)) { showMsg("Invalid name", "Names starting with render_ are reserved."); return; }
      if (tagUsedAnywhere(cur.name)) { showMsg("Page in use", "\"" + cur.name + "\" is invoked as a component by other elements."); return; }
      snap("rename page " + cur.name);
      curPage().name = uniquePageName(name);
      setOut(); renderPages(); autoPreview();
      var fj = document.getElementById("f_page"); if (fj) fj.value = curPage().name;
    }, "Letters, digits and _ only");
    pi.value = curPage().name;
  };
  pr.appendChild(pl2); pr.appendChild(pi);
  dsec.appendChild(pr);
  var olr = document.createElement("div"); olr.className = "rowline";
  var oll = document.createElement("label"); oll.textContent = "onLoad"; oll.title = "Function called once when the page loads (e.g. a table loader)";
  var oli = document.createElement("input"); oli.id = "f_onload"; oli.value = curPage().onLoad || ""; oli.placeholder = "function name"; oli.spellcheck = false;
  oli.onchange = function(){ updView("onLoad"); };
  olr.appendChild(oll); olr.appendChild(oli);
  dsec.appendChild(olr);
  // Page actions live with the page attributes: duplicate / delete / save as
  // template operate on the ACTIVE page.
  var pbr = document.createElement("div"); pbr.className = "rowline";
  var pdbtn = document.createElement("button"); pdbtn.className = "btn"; pdbtn.style.fontSize = "11px"; pdbtn.textContent = "Duplicate"; pdbtn.title = "Duplicate the active page"; pdbtn.onclick = duplicatePage;
  var pxbtn = document.createElement("button"); pxbtn.className = "btn"; pxbtn.style.fontSize = "11px"; pxbtn.textContent = "Delete"; pxbtn.title = "Delete the active page"; pxbtn.onclick = delPageFlow;
  var psbtn = document.createElement("button"); psbtn.className = "btn"; psbtn.style.fontSize = "11px"; psbtn.textContent = "Save as template"; psbtn.title = "Save the active page to the library"; psbtn.onclick = savePageTemplateFlow;
  pbr.appendChild(pdbtn); pbr.appendChild(pxbtn); pbr.appendChild(psbtn);
  dsec.appendChild(pbr);
  box.appendChild(dsec);
  var paths = selectedPaths();
  var nodes = selBody ? [] : paths.map(function(p){ return byPath(p, root.children); }).filter(Boolean);
  var n = selPath ? byPath(selPath, root.children) : null;
  var isToastHost = !!(nodes.length === 1 && n && n.kind === "toast");
  if (!isToastHost) __toastPreviewPending = null;
  __runColorHint = (nodes.length === 1 && n && n.kind === "text") ? commonRunColor(n) : "";
  if (!nodes.length && !selBody) { box.insertAdjacentHTML("beforeend", '<div class="hint">Select an element (tree or preview) to style it.</div>'); return; }
  if (nodes.some(function(m){ return m && m.component; })) {
    box.insertAdjacentHTML("beforeend", '<div class="hint">This is a parameterized view (component). It is kept in the SUI as-is.</div>');
    return;
  }

  // Exact-type banner: the element kind is always visible, independently of
  // the friendly label shown in the structure tree.
  var exactTypes = [];
  if (selBody) exactTypes.push("body");
  else nodes.forEach(function(m){
    var t = (m.kind === "text" && m.tag) ? m.tag : m.kind;
    if (exactTypes.indexOf(t) < 0) exactTypes.push(t);
  });
  if (exactTypes.length) {
    var typeBar = document.createElement("div");
    typeBar.className = "sel-type";
    var typeLead = document.createElement("span");
    typeLead.textContent = exactTypes.length === 1 ? "Selected element" : "Selected elements";
    typeBar.appendChild(typeLead);
    exactTypes.forEach(function(t){
      var code = document.createElement("code");
      code.textContent = t;
      typeBar.appendChild(code);
    });
    typeBar.title = selPath ? "Path: " + selPath.join(".") : "Page body";
    box.appendChild(typeBar);
  }

  var states = ["", ":hover", ":focus", ":active"];
  var sd = document.createElement("div"); sd.className = "states";
  states.forEach(function(s){
    var b = document.createElement("button");
    b.className = "btn" + (selState === s ? " on" : "");
    b.textContent = s === "" ? "base" : s;
    b.onclick = function(){ selState = s; renderInspector(); };
    sd.appendChild(b);
  });
  box.appendChild(sd);

  if (selBody) {
    var hb = document.createElement("div"); hb.className = "hint";
    hb.textContent = "Page body (root container).";
    box.appendChild(hb);
  }

  // Content section: only elements whose kind actually carries text show the
  // Text field (media elements get none, a rule has no label, ...); media
  // kinds show src instead. With a multi-selection the field appears only if
  // EVERY selected element supports it, showing the shared value.
  var capText = allKindsHave(nodes, TEXT_KINDS);
  var capSrc = allKindsHave(nodes, SRC_KINDS);
  if (capText || capSrc) {
    var sec1 = document.createElement("div"); sec1.className = "ins-sec"; sec1.innerHTML = '<div class="ins-sec-h">Content</div>';
    if (capText) {
      var tr = document.createElement("div"); tr.className = "rowline";
      var tl = document.createElement("label"); tl.textContent = "Text"; tr.appendChild(tl);
      var ti = document.createElement("input"); ti.id = "f_text"; ti.value = commonVal(nodes, "text");
      ti.oninput = function(ev){ upd("text", ev.target.value); };
      tr.appendChild(ti);
      sec1.appendChild(tr);
    }
    if (capSrc) {
      var sr = document.createElement("div"); sr.className = "rowline";
      var sl2 = document.createElement("label"); sl2.textContent = "src"; sl2.title = "Resource path or URL"; sr.appendChild(sl2);
      var si = document.createElement("input"); si.value = commonVal(nodes, "src"); si.placeholder = "https://…";
      si.onchange = function(){ upd("src", si.value); };
      sr.appendChild(si);
      sec1.appendChild(sr);
    }
    box.appendChild(sec1);
  }

  if (nodes.length && everyKind(nodes, function(k){ return k === "iframe"; })) {
    var ifSec = document.createElement("div"); ifSec.className = "ins-sec"; ifSec.innerHTML = '<div class="ins-sec-h">Iframe security</div>';
    var ifFields = [
      ["sandbox", "sandbox tokens", "allow-scripts allow-forms", "Default empty sandbox blocks scripts, forms, popups and same-origin."],
      ["allow", "permissions", "camera; microphone https://example.com", "Permissions Policy features."],
      ["referrerpolicy", "referrer policy", "no-referrer", "no-referrer, origin, strict-origin, ..."],
      ["loading", "loading", "lazy", "lazy or eager"],
      ["width", "width", "640", "Integer pixels"],
      ["height", "height", "360", "Integer pixels"],
      ["srcAllow", "src allow-list", "https://example.com", "Literal origin/path allow-list, required for dynamic src."]
    ];
    ifFields.forEach(function(f) {
      var row = document.createElement("div"); row.className = "rowline";
      var lb = document.createElement("label"); lb.textContent = f[1]; lb.title = f[3]; row.appendChild(lb);
      var inp = document.createElement("input");
      inp.value = commonVal(nodes, f[0]) || "";
      inp.placeholder = f[2];
      inp.oninput = function(ev){ upd(f[0], ev.target.value); };
      row.appendChild(inp);
      ifSec.appendChild(row);
    });
    var ifh = document.createElement("div"); ifh.className = "hint";
    ifh.style.cssText = "font-size:10px;margin-top:4px";
    ifh.textContent = "Empty sandbox = no scripts/forms/popups/same-origin. Dynamic src is rejected unless srcAllow lists its origin.";
    ifSec.appendChild(ifh);
    box.appendChild(ifSec);
  }

  var sel = selClass();
  var rule = findRule(sel);
  var spec = rule ? (rule.specificity && rule.specificity[0] ? rule.specificity[0] : "0,0,0") : "";
  var sp = document.createElement("div"); sp.className = "spec";
  var overridden = problems.some(function(p){ return p.kind === "override" && p.sel === "." + sel; });
  if (sel) sp.innerHTML = 'selector <b>' + (sel === "body" ? "body" : "." + sel) + '</b>' + (spec ? ' &nbsp; specificity <b>' + spec + '</b>' : '') + (overridden ? '<div class="warn">overridden — check Problems</div>' : '');
  else sp.innerHTML = 'assign a class to style this element';
  box.appendChild(sp);

  if (nodes.length) {
    var clsSec = document.createElement("div"); clsSec.className = "ins-sec"; clsSec.innerHTML = '<div class="ins-sec-h">Appearance</div>';
    var cls = document.createElement("div"); cls.className = "rowline";
    var cl = document.createElement("label"); cl.textContent = "class"; cls.appendChild(cl);
    var ci = document.createElement("input"); ci.id = "f_class"; ci.setAttribute("list", "classlist");
    ci.value = commonVal(nodes, "class");
    ci.oninput = function(ev){ snap(); nodes.forEach(function(m){ m.class = ev.target.value; }); renderTree(); setOut(); autoPreview(); buildDatalist(); };
    cls.appendChild(ci);
    var dl = document.createElement("datalist"); dl.id = "classlist"; cls.appendChild(dl);
    clsSec.appendChild(cls);
    buildDatalist();
    // Reactive class: any SUI expression (e.g. active ? "on" : ""). It is
    // appended to the element's built-in class and updates with the state.
    var cxs = document.createElement("div"); cxs.className = "rowline";
    var cxl = document.createElement("label"); cxl.textContent = "class expr"; cxl.title = 'Reactive class expression, e.g. active ? "card on" : "card"'; cxs.appendChild(cxl);
    var cxi = document.createElement("input"); cxi.id = "f_classexpr"; cxi.placeholder = 'active ? "on" : ""';
    cxi.value = commonVal(nodes, "classExpr");
    cxi.oninput = function(ev){ snap("set class expr"); nodes.forEach(function(m){ m.classExpr = ev.target.value; }); setOut(); autoPreview(); };
    cxs.appendChild(cxi);
    clsSec.appendChild(cxs);
    // Reactive style: a SUI expression producing a CSS declaration string.
    var sxs = document.createElement("div"); sxs.className = "rowline";
    var sxl = document.createElement("label"); sxl.textContent = "style expr"; sxl.title = 'Reactive style expression, e.g. "background:" + bg'; sxs.appendChild(sxl);
    var sxi = document.createElement("input"); sxi.id = "f_styleexpr";
    if (nodes.length === 1 && n && n.kind === "toast") {
      sxi.placeholder = "width: 420px; border-width: 3px;";
      sxi.title = "Toast card CSS declarations. Edit these to evaluate and change the card style; custom properties such as --sui-toast-icon-bg are inherited by the toast parts.";
      sxi.value = toastCardCss(n);
      sxi.onchange = function(ev){ applyToastCardCss(n, ev.target.value); };
    } else {
      sxi.placeholder = '"background:" + bg';
      sxi.value = n.styleExpr || "";
      sxi.oninput = function(ev){ snap("set style expr"); n.styleExpr = ev.target.value; setOut(); autoPreview(); };
    }
    sxs.appendChild(sxi);
    clsSec.appendChild(sxs);
    var lk = document.createElement("div"); lk.className = "rowline";
    var ll = document.createElement("label"); ll.textContent = "link"; lk.appendChild(ll);
    var li = document.createElement("input"); li.id = "f_link"; li.value = n.link || ""; li.placeholder = "https://…";
    li.oninput = function(ev){ snap("set link"); n.link = ev.target.value; n.linkExpr = false; setOut(); autoPreview(); };
    lk.appendChild(li);
    clsSec.appendChild(lk);
    box.appendChild(clsSec);
  }

  // Value bindings and events exist only for kinds whose tables define them
  // explicitly: the sections carry the INTERSECTION of the selection and
  // disappear entirely when an element (or one of the selected ones) has none.
  if (nodes.length && everyKind(nodes, function(k){ return k !== "table"; })) {
    var bc = bindingControls(nodes);
    if (bc) box.appendChild(bc);
    var evSec = eventsControls(nodes);
    if (evSec) box.appendChild(evSec);
  }

  if (nodes.length === 1 && n && (n.kind === "menubar" || n.kind === "menu" || n.kind === "menuitem")) {
    var menuHint = document.createElement("div");
    menuHint.className = "hint";
    menuHint.textContent = "Menu editor: right-click this row in the element tree for Add menu/item/submenu, Duplicate, Move and Delete.";
    box.appendChild(menuHint);
  }
  if (nodes.length === 1 && n && (n.kind === "tabs" || n.kind === "tab")) {
    var tabsHint = document.createElement("div");
    tabsHint.className = "hint";
    tabsHint.textContent = "Tabs editor: right-click a tabs/tab row for Add tab, Edit label, Duplicate, Move and Delete.";
    box.appendChild(tabsHint);
  }
  if (nodes.length === 1 && n && (n.kind === "accordion" || n.kind === "accordionitem")) {
    var accHint = document.createElement("div");
    accHint.className = "hint";
    accHint.textContent = "Accordion editor: right-click an accordion/item row for Add accordion item, Edit label, Duplicate, Move and Delete.";
    box.appendChild(accHint);
  }
  if (nodes.length === 1 && n && (n.kind === "modal" || n.kind === "modalbutton")) {
    var modalHint = document.createElement("div");
    modalHint.className = "hint";
    modalHint.textContent = "Modal editor: right-click a modal/modalbutton row for Add modal button, Edit label, Duplicate, Move and Delete.";
    box.appendChild(modalHint);
  }
  if (nodes.length === 1 && n && n.kind === "combobox") {
    box.appendChild(comboboxControls(n));
  }
  if (nodes.length === 1 && n && n.kind === "datepicker") {
    box.appendChild(datepickerControls(n));
  }
  if (nodes.length === 1 && n && n.kind === "tree") {
    box.appendChild(treeControls(n));
  }
  if (nodes.length === 1 && n && n.kind === "treeitem") {
    box.appendChild(treeItemControls(n));
  }
  if (nodes.length === 1 && n && isChartKind(n)) {
    box.appendChild(chartControls(n));
  }
  if (nodes.length === 1 && n && !n.raw && n.kind !== "toast" && n.kind !== "text") {
    box.appendChild(tooltipPropertyControls(n));
  }
  if (nodes.length === 1 && n && n.kind === "toast") {
    box.appendChild(toastControls(n));
    previewToastSoon(n);
  }

  // Node-internal structures (table cells, list items, rich-text runs) are
  // edited one node at a time: single-selection only.
  if (nodes.length === 1 && n && n.kind === "table" && n.cols) {
    box.appendChild(tableControls(n));
  }
  if (nodes.length === 1 && n && n.kind === "list") {
    box.appendChild(listControls(n));
  }
  if (nodes.length === 1 && n && n.kind === "text" && n.runs) {
    box.appendChild(textEditor(n));
  }
  if (nodes.length && everyKind(nodes, function(k){ return k === "input" || k === "textarea"; })) {
    box.appendChild(inputControls(nodes));
  }
  var reqSec = requiredControls(nodes);
  if (reqSec) box.appendChild(reqSec);
  if (nodes.length === 1 && n && (n.kind === "modal" || n.kind === "modalbutton")) {
    box.appendChild(modalControls(n));
  }

  var controls = [
    (!selBody && nodes.length && everyKind(nodes, capAlign)) ? alignControls() : null,
    everyKind(nodes, capLayout) ? acc("Layout", (function(){
      var rows = [
        selBox("display", "display", ["", "block", "flex", "grid", "inline", "none"]),
        containerSelBox("flex-direction", "flex-direction", ["", "row", "column", "row-reverse", "column-reverse"], nodes),
        containerSelBox("justify-content", "justify-content", ["", "flex-start", "center", "flex-end", "space-between", "space-around"], nodes),
        containerSelBox("align-items", "align-items", ["", "stretch", "center", "flex-start", "flex-end", "baseline"], nodes),
        selBox("align-self", "align-self", ["", "auto", "stretch", "center", "flex-start", "flex-end", "baseline"]),
        containerSelBox("align-content", "align-content", ["", "stretch", "center", "flex-start", "flex-end", "space-between", "space-around"], nodes),
        selBox("vertical-align", "vertical-align", ["", "baseline", "top", "middle", "bottom", "sub", "super", "text-top", "text-bottom"]),
        containerTxt("gap", "gap", "e.g. 12px", nodes)
      ];
      if (!selIsContainer(nodes)) {
        var ph = document.createElement("div"); ph.className = "hint";
        ph.style.cssText = "font-size:10px;margin-top:2px";
        ph.textContent = "flex-direction, justify-content, align-items, align-content and gap lay out CHILDREN: they are applied to the parent container (or the page view).";
        rows.push(ph);
      }
      return rows;
    })()) : null,
    everyKind(nodes, capGrid) ? acc("Grid", [
      txt("grid-template-columns", "columns", "e.g. 1fr 1fr"),
      txt("grid-template-rows", "rows"),
      txt("grid-column", "grid-column", "e.g. 1 / 3"),
      txt("grid-row", "grid-row"),
      txt("grid-area", "grid-area"),
      selBox("place-items", "place-items", ["", "start", "center", "end", "stretch"]),
      selBox("place-content", "place-content", ["", "start", "center", "end", "stretch", "space-between", "space-around"])
    ]) : null,
    acc("Position", (function(){
      var r = [
        selBox("position", "position", ["", "static", "relative", "absolute", "fixed", "sticky"]),
        txt("top", "top"),
        txt("right", "right"),
        txt("bottom", "bottom"),
        txt("left", "left"),
        txt("z-index", "z-index")
      ];
      if (!positionHasEffect() && !selBody) {
        var h = document.createElement("div"); h.className = "hint";
        h.style.cssText = "font-size:10px;color:#ffb454;margin-top:2px";
        h.textContent = "top / right / bottom / left / z-index have no effect until position is set (relative, absolute, fixed).";
        r.push(h);
      }
      return r;
    })()),
    responsiveControls(),
    acc("Box model", [
      txt("margin", "margin"),
      txt("padding", "padding")
    ]),
    acc("Border", [
      selBox("border-style", "style", ["", "none", "solid", "dashed", "dotted", "double", "groove", "ridge", "inset", "outset"]),
      txt("border-width", "width"),
      color("border-color", "color"),
      roundness("border-radius", "radius")
    ]),
    everyKind(nodes, capTypography) ? acc("Typography", [
      range("font-size", "font-size", "8", "72", "px"),
      fontBox(),
      color("color", "color"),
      selBox("font-weight", "font-weight", ["", "400", "500", "600", "700"]),
      selBox("font-style", "font-style", ["", "normal", "italic", "oblique"]),
      selBox("text-transform", "text-transform", ["", "none", "uppercase", "lowercase", "capitalize"]),
      selBox("text-decoration", "text-decoration", ["", "none", "underline", "line-through", "overline"]),
      selBox("text-align", "text-align", ["", "left", "center", "right", "justify"]),
      txt("line-height", "line-height"),
      txt("letter-spacing", "letter-spacing"),
      txt("word-spacing", "word-spacing"),
      txt("text-indent", "text-indent"),
      shadowBox("text-shadow", "text-shadow"),
      selBox("white-space", "white-space", ["", "normal", "nowrap", "pre", "pre-wrap", "pre-line"])
    ]) : null,
    acc("Background", [
      color("background-color", "background-color"),
      transparency("opacity", "Opacity")
    ]),
    acc("Effects", [
      shadowBox("box-shadow", "box-shadow")
    ]),
    acc("Size", [
      txt("width", "width"),
      txt("height", "height"),
      txt("max-width", "max-width"),
      fillControls()
    ])
  ];
  if (isToastHost) {
    var toastSections = { "Box model": 1, "Border": 1, "Typography": 1, "Background": 1, "Effects": 1, "Size": 1 };
    controls = controls.filter(function(a){
      if (!a) return false;
      var h = a.querySelector(".acchead span");
      return !!(h && toastSections[h.textContent.trim()]);
    });
  }
  controls.forEach(function(a){ if (a) box.appendChild(a); });

  // Drag & drop is a dedicated section at the END of the inspector, after all
  // visual/style sections. It contains the source/target keys, file state and
  // picker options; the matching onDrag*/onDrop handlers stay in Events.
  if (nodes.length && everyKind(nodes, function(k){ return k !== "table"; })) {
    var dndSec = dndControls(nodes);
    if (dndSec) box.appendChild(dndSec);
  }
  if (selCell) {
    var ct = document.getElementById("celltext");
    if (ct) { ct.focus(); ct.select(); }
  }
  syncInspectorSections();
  requestLiveStyles();
}

// Typical events per element kind.
var ELEMENT_EVENTS = {
  button:   ["onClick", "onMouseEnter", "onMouseLeave", "onFocus", "onBlur", "onDoubleClick"],
  a:        ["onClick", "onMouseEnter", "onMouseLeave", "onFocus", "onBlur"],
  input:    ["onInput", "onChange", "onFocus", "onBlur", "onKeyDown", "onKeyUp", "onClick", "onMouseEnter", "onMouseLeave"],
  textarea: ["onInput", "onChange", "onFocus", "onBlur", "onKeyDown", "onKeyUp"],
  checkbox: ["onChange", "onClick", "onFocus", "onBlur"],
  switch:   ["onChange", "onClick", "onFocus", "onBlur"],
  radio:    ["onChange", "onClick", "onFocus", "onBlur"],
  select:   ["onChange", "onFocus", "onBlur", "onClick"],
  combobox: ["onChange"],
  datepicker: ["onChange"],
  tree:     ["onChange", "onRename"],
  chart:    ["onTaskChange", "onDataChange", "onCalendarChange", "onResourceChange", "onCommit",
             "onShapeClick", "onShapeDoubleClick", "onShapeMouseOver", "onShapeMouseOut", "onShapeContextMenu",
             "onLinkClick", "onLinkDoubleClick", "onLinkMouseOver", "onLinkMouseOut", "onLinkContextMenu",
             "onTagClick", "onTagDoubleClick", "onTagMouseOver", "onTagMouseOut", "onTagContextMenu"],
  graph:    ["onTaskChange", "onDataChange", "onCalendarChange", "onResourceChange", "onCommit",
             "onShapeClick", "onShapeDoubleClick", "onShapeMouseOver", "onShapeMouseOut", "onShapeContextMenu",
             "onLinkClick", "onLinkDoubleClick", "onLinkMouseOver", "onLinkMouseOut", "onLinkContextMenu",
             "onTagClick", "onTagDoubleClick", "onTagMouseOver", "onTagMouseOut", "onTagContextMenu"],
  slider:   ["onInput", "onChange", "onFocus", "onBlur"],
  number:   ["onInput", "onChange", "onFocus", "onBlur"],
  file:     ["onChange", "onFocus", "onBlur"],
  color:    ["onInput", "onChange", "onFocus", "onBlur"],
  date:     ["onInput", "onChange", "onFocus", "onBlur"],
  time:     ["onInput", "onChange", "onFocus", "onBlur"],
  form:     ["onSubmit", "onReset"],
  img:      ["onClick", "onMouseEnter", "onMouseLeave"],
  row:      ["onClick", "onMouseEnter", "onMouseLeave"],
  col:      ["onClick", "onMouseEnter", "onMouseLeave"],
  div:      ["onClick", "onMouseEnter", "onMouseLeave"],
  card:     ["onClick", "onMouseEnter", "onMouseLeave"],
  alert:    ["onClick"],
  badge:    ["onClick", "onMouseEnter", "onMouseLeave"],
  tabs:     ["onClick"],
  menuitem: ["onClick", "onMouseEnter", "onMouseLeave", "onFocus", "onBlur"],
  modalbutton: ["onClick", "onMouseEnter", "onMouseLeave", "onFocus", "onBlur"],
};
// Drag events are valid on every element: containers can be drop zones, and
// leaves can be drag handles. The runtime still decides what is useful.
var DRAG_EVENTS = ["onDragStart", "onDragEnter", "onDragOver", "onDragLeave", "onDrop", "onDragEnd"];
Object.keys(ELEMENT_EVENTS).forEach(function(k){
  DRAG_EVENTS.forEach(function(ev){ if (ELEMENT_EVENTS[k].indexOf(ev) < 0) ELEMENT_EVENTS[k].push(ev); });
});
var EVENT_LABELS = {
  onClick: "onClick", onDoubleClick: "onDoubleClick", onMouseEnter: "onMouseEnter", onMouseLeave: "onMouseLeave",
  onFocus: "onFocus", onBlur: "onBlur", onInput: "onInput", onChange: "onChange",
  onRename: "onRename", onTaskChange: "onTaskChange", onDataChange: "onDataChange", onCalendarChange: "onCalendarChange", onResourceChange: "onResourceChange", onCommit: "onCommit",
  onKeyDown: "onKeyDown", onKeyUp: "onKeyUp", onSubmit: "onSubmit", onReset: "onReset",
  onDragStart: "onDragStart", onDragEnter: "onDragEnter", onDragOver: "onDragOver",
  onDragLeave: "onDragLeave", onDrop: "onDrop", onDragEnd: "onDragEnd"
};
// ELEMENT_EVENTS[kind] is itself the per-kind event list (an empty list means
// the element cannot fire any handler, so the inspector hides the section).

// ---- element capability tables ---------------------------------------------
// The inspector is DYNAMIC: a field/section must appear only for elements
// that can actually have that attribute (a media element has no text, a rule
// cannot be aligned, a spacer accepts no events, ...). These tables are the
// single source of truth; when two or more elements are selected together the
// inspector intersects them and shows only the capabilities they share.
var TEXT_KINDS = { div:1, section:1, article:1, main:1, header:1, footer:1, nav:1, aside:1, span:1,
  h1:1, h2:1, h3:1, h4:1, h5:1, h6:1, p:1, blockquote:1, pre:1, code:1, mark:1, strong:1, em:1, small:1,
  li:1, legend:1, label:1, tcell:1, option:1, button:1, a:1, checkbox:1, switch:1, radio:1, card:1, badge:1, alert:1,
  tabs:1, tab:1, accordion:1, accordionitem:1, tree:1, treeitem:1, chart:1,
  modal:1, modalbutton:1, menubar:1, menu:1, menuitem:1 };
var LINK_KINDS = { a:1 };
var SRC_KINDS = { img:1, video:1, audio:1, iframe:1 };
// Flow-level shaping (display/flex/grid internals) only makes sense on
// containers and block/lexical elements.
var BOX_KINDS = { div:1, section:1, article:1, main:1, header:1, footer:1, nav:1, aside:1, block:1, row:1, col:1, card:1, form:1, fieldset:1, tabs:1, tab:1, accordion:1, accordionitem:1, toast:1, modal:1, menubar:1, menu:1, ul:1, ol:1, list:1, body:1, chart:1 };
var LEX_KINDS = { h1:1, h2:1, h3:1, h4:1, h5:1, h6:1, p:1, pre:1, blockquote:1, code:1, mark:1, strong:1, em:1, small:1, span:1, legend:1, li:1, a:1, button:1, badge:1, alert:1, text:1, menuitem:1 };
var TEXTWIDGET_KINDS = { input:1, textarea:1, select:1, combobox:1, datepicker:1, number:1, date:1, time:1 };
// Pure layout rules/spacers: their whole job is positioning, they cannot be
// aligned inside a flow and their font/text settings are meaningless.
var RULE_KINDS = { hr:1, br:1, spacer:1, divider:1 };
function capLayout(kind) { return !!(BOX_KINDS[kind] || LEX_KINDS[kind]); }
function capGrid(kind) { return !!BOX_KINDS[kind]; }
function capTypography(kind) { return !!(BOX_KINDS[kind] || LEX_KINDS[kind] || TEXTWIDGET_KINDS[kind]); }
function capAlign(kind) { return !(RULE_KINDS[kind] || kind === "thead" || kind === "tbody" || kind === "tfoot" || kind === "trow"); }
// everyKind / allKindsHave evaluate a capability across the (multi-)selection.
function selNodes() {
  var paths = multiSel.slice();
  if (selPath && !paths.some(function(q){ return samePath(q, selPath); })) paths.push(selPath);
  return paths.map(function(p){ return byPath(p, root.children); }).filter(Boolean);
}
function everyKind(nodes, fn) { return nodes.every(function(m){ return fn(m.kind); }); }
function allKindsHave(nodes, table) {
  if (!nodes.length) return false;
  for (var i = 0; i < nodes.length; i++) if (!table[nodes[i].kind]) return false;
  return true;
}
// intersectFields: a per-kind FIELD LIST (events, bindings) common to every
// selected node; null-safe — one kind without the field kills the entry.
function commonFieldList(nodes, table) {
  if (!nodes.length) return null;
  var cur = null;
  for (var i = 0; i < nodes.length; i++) {
    var f = table[nodes[i].kind];
    if (!f || !f.length) return null;
    if (cur === null) { cur = f.slice(); continue; }
    cur = cur.filter(function(x){ return f.indexOf(x) >= 0; });
  }
  return cur;
}
// commonVal: the shared value of an attribute across the selection ("" when
// the selections differ, shown as an empty field the user can overwrite).
function commonVal(nodes, field) {
  var v = nodes[0][field] || "";
  for (var i = 1; i < nodes.length; i++) { if ((nodes[i][field] || "") !== v) return ""; }
  return v;
}
// BINDING_FIELDS lists the value/state binding attributes each control kind
// offers in the inspector. A binding holds a SUI expression (usually the name
// of a state), which is emitted unquoted into the generated source.
var BINDING_FIELDS = {
  checkbox: ["checked", "disabled"],
  switch:   ["checked", "disabled"],
  radio:    ["checked", "name", "disabled"],
  number:   ["value", "min", "max", "step", "disabled"],
  slider:   ["value", "min", "max", "step", "disabled"],
  select:   ["value", "disabled"],
  file:     ["accept", "multiple", "disabled"],
  color:    ["value", "disabled"],
  date:     ["value", "disabled"],
  time:     ["value", "disabled"]
};
var BINDING_LABELS = { checked: "checked", value: "value", disabled: "disabled", name: "name (radio group)", min: "min", max: "max", step: "step", accept: "accept", multiple: "multiple" };
// Binary binding fields rendered as a graphical switch. Literal-only: when the
// current value is a binding expression (e.g. checked: active) the plain text
// input stays so the binding can still be edited.
var BINARY_BINDING = { checked: 1, disabled: 1, multiple: 1 };

// mkSwitch builds a small on/off toggle. 'on' sets the initial state, 'mixed'
// shows an indeterminate knob (multi-selection with different values), and
// 'onchg(bool)' receives each toggle.
function mkSwitch(on, mixed, onchg) {
  var w = document.createElement("label"); w.className = "isw" + (mixed ? " mixed" : "");
  w.title = "on / off";
  var ios = document.createElement("input"); ios.type = "checkbox"; ios.checked = !!on;
  ios.onchange = function(){ w.classList.remove("mixed"); onchg(ios.checked); };
  var trk = document.createElement("span"); trk.className = "iswTrk";
  var kn = document.createElement("span"); kn.className = "iswKnob";
  trk.appendChild(kn);
  w.appendChild(ios); w.appendChild(trk);
  return w;
}

// bindingControls renders the "Value" section for value-carrying controls
// (switch/checkbox/radio/number/...). Each field is a raw SUI expression, and
// the datalist offers the declared states (plus true/false for checked), so
// binding a control to a state is one click. With a multi-selection only the
// binding fields common to EVERY selected control appear, reading the shared
// value. Returns null when no common binding exists.
function bindingControls(nodes) {
  var fields = commonFieldList(nodes, BINDING_FIELDS);
  if (!fields || !fields.length) return null;
  var wrap = document.createElement("div"); wrap.className = "ins-sec";
  wrap.innerHTML = '<div class="ins-sec-h">Value</div>';
  var dl = document.createElement("datalist"); dl.id = "binding-states";
  (root.states || []).forEach(function(s){ if (s && s.name) { var o = document.createElement("option"); o.value = s.name; dl.appendChild(o); } });
  wrap.appendChild(dl);
  fields.forEach(function(f) {
    var row = document.createElement("div"); row.className = "rowline";
    var lb = document.createElement("label"); lb.textContent = BINDING_LABELS[f] || f; row.appendChild(lb);
    var cur = commonVal(nodes, f);
    var sv = cur == null ? "" : String(cur);
    var literal = sv === "" || sv === "true" || sv === "false";
    if (BINARY_BINDING[f] && literal) {
      // Binary field with a literal value: render the on/off switch.
      var mixed = false;
      if (nodes.length > 1) {
        var vals = nodes.map(function(m){ var v = m[f]; return v == null ? "" : String(v); });
        mixed = vals.some(function(v){ return v !== vals[0]; });
      }
      var ios = mkSwitch(sv === "true", mixed, function(on){ upd(f, on ? "true" : ""); });
      row.appendChild(ios);
    } else {
      // Expression binding (e.g. checked: active) or free-form field: text input.
      var i = document.createElement("input");
      i.value = cur;
      i.setAttribute("list", "binding-states");
      if (f === "checked") i.placeholder = "state name / true / false";
      i.oninput = function(){ upd(f, i.value); };
      i.onchange = function(){ upd(f, i.value); };
      row.appendChild(i);
    }
    wrap.appendChild(row);
  });
  var hint = document.createElement("div"); hint.className = "hint";
  hint.style.cssText = "font-size:10px;margin-top:4px";
  hint.textContent = "Bind to a state (e.g. active). A missing state is created automatically; the control then reflects and updates it.";
  wrap.appendChild(hint);
  return wrap;
}

// eventsControls shows only the event handler fields the selected elements
// can ALL react to: the per-kind event table is intersected, and kinds with
// no events (rules, spacers, ... simply cannot fire any) hide the section.
function eventsControls(nodes) {
  var evs = commonFieldList(nodes, ELEMENT_EVENTS);
  if (!evs || !evs.length) return null;
  var rows = [];
  evs.forEach(function(evType) {
    var row = document.createElement("div"); row.className = "rowline";
    var lb = document.createElement("label"); lb.textContent = EVENT_LABELS[evType] || evType; row.appendChild(lb);
    row.appendChild(onEventSelect(nodes, evType));
    rows.push(row);
  });
  var hint = document.createElement("div"); hint.className = "hint";
  hint.style.cssText = "font-size:10px;margin-top:4px";
  hint.textContent = "Handlers must be declared with fn in the SUI. Use + new function to create and attach one.";
  rows.push(hint);
  return acc("Events", rows);
}

// dndControls edits the drag-and-drop attributes. The runtime is generic, so
// the same controls are offered for every element; a row element can be both
// a drag source and a drop target.
function dndControls(nodes) {
  if (!nodes.length) return null;
  var rows = [];

  var dr = document.createElement("div"); dr.className = "rowline";
  var dl = document.createElement("label"); dl.textContent = "draggable"; dl.title = "Make this element a drag source. dragKey supplies the payload."; dr.appendChild(dl);
  var dOn = nodes.every(function(m){ return !!m.draggable; });
  var dMixed = nodes.length > 1 && !dOn && nodes.some(function(m){ return !!m.draggable; });
  dr.appendChild(mkSwitch(dOn, dMixed, function(on){ upd("draggable", on); }));
  rows.push(dr);

  rows.push(rowInput("dragKey", commonVal(nodes, "dragKey"), function(v){ upd("dragKey", v); }, "item.id / index"));
  rows.push(rowInput("dropKey", commonVal(nodes, "dropKey"), function(v){ upd("dropKey", v); }, "item.id / index"));

  var fr = document.createElement("div"); fr.className = "rowline";
  var fl = document.createElement("label"); fl.textContent = "dropFiles"; fl.title = "Name of a string state that receives dropped file paths."; fr.appendChild(fl);
  var fi = document.createElement("input"); fi.value = commonVal(nodes, "dropFiles"); fi.placeholder = "state name"; fi.spellcheck = false;
  var fdl = document.createElement("datalist"); fdl.id = "drop-states";
  (root.states || []).forEach(function(s){ if (s && s.name) { var o = document.createElement("option"); o.value = s.name; fdl.appendChild(o); } });
  fi.setAttribute("list", "drop-states");
  fi.onchange = function(){ upd("dropFiles", fi.value.trim()); };
  fr.appendChild(fi); fr.appendChild(fdl); rows.push(fr);

  var pr = document.createElement("div"); pr.className = "rowline";
  var pl = document.createElement("label"); pl.textContent = "pickFiles"; pl.title = "Name of a string state that receives files chosen from the native dialog."; pr.appendChild(pl);
  var pi = document.createElement("input"); pi.value = commonVal(nodes, "pickFiles"); pi.placeholder = "state name"; pi.spellcheck = false;
  var pdl = document.createElement("datalist"); pdl.id = "pick-states";
  (root.states || []).forEach(function(s){ if (s && s.name) { var o = document.createElement("option"); o.value = s.name; pdl.appendChild(o); } });
  pi.setAttribute("list", "pick-states");
  pi.onchange = function(){ upd("pickFiles", pi.value.trim()); };
  pr.appendChild(pi); pr.appendChild(pdl); rows.push(pr);

  var pmr = document.createElement("div"); pmr.className = "rowline";
  var pml = document.createElement("label"); pml.textContent = "pickMode"; pml.title = "native: webkitdirectory; fs: File System Access API; auto: fs when available, otherwise native."; pmr.appendChild(pml);
  var pms = document.createElement("select");
  ["native", "auto", "fs"].forEach(function(mode){
    var o = document.createElement("option"); o.value = mode; o.textContent = mode;
    if ((commonVal(nodes, "pickMode") || "native") === mode) o.selected = true;
    pms.appendChild(o);
  });
  pms.onchange = function(){ upd("pickMode", pms.value); };
  pmr.appendChild(pms); rows.push(pmr);

  var allFile = nodes.every(function(m){ return m.kind === "file"; });
  if (!allFile) {
    rows.push(rowInput("accept", commonVal(nodes, "accept"), function(v){ upd("accept", v); }, ".csv,.txt,image/*"));
    var mr = document.createElement("div"); mr.className = "rowline";
    var ml = document.createElement("label"); ml.textContent = "multiple"; mr.appendChild(ml);
    var mOn = nodes.every(function(m){ return !!m.multiple; });
    var mMixed = nodes.length > 1 && !mOn && nodes.some(function(m){ return !!m.multiple; });
    mr.appendChild(mkSwitch(mOn, mMixed, function(on){ upd("multiple", on); }));
    rows.push(mr);
    var dr2 = document.createElement("div"); dr2.className = "rowline";
    var dl2 = document.createElement("label"); dl2.textContent = "directory"; dr2.appendChild(dl2);
    var dOn2 = nodes.every(function(m){ return !!m.directory; });
    var dMixed2 = nodes.length > 1 && !dOn2 && nodes.some(function(m){ return !!m.directory; });
    dr2.appendChild(mkSwitch(dOn2, dMixed2, function(on){ upd("directory", on); }));
    rows.push(dr2);
  } else {
    var ir = document.createElement("div"); ir.className = "rowline";
    var il = document.createElement("label"); il.textContent = "directory"; il.title = "Use a directory picker (webkitdirectory)."; ir.appendChild(il);
    var iOn = nodes.every(function(m){ return !!m.directory; });
    var iMixed = nodes.length > 1 && !iOn && nodes.some(function(m){ return !!m.directory; });
    ir.appendChild(mkSwitch(iOn, iMixed, function(on){ upd("directory", on); }));
    rows.push(ir);
  }

  var hint = document.createElement("div"); hint.className = "hint";
  hint.style.cssText = "font-size:10px;margin-top:4px";
  hint.textContent = "dropFiles accepts dragged files/folders. pickFiles opens the native dialog when the element is clicked (use a button or a file control). Browsers expose names/relative paths, not absolute paths.";
  rows.push(hint);
  return acc("Drag &amp; drop", rows);
}

function modalControls(n) {
  var wrap = document.createElement("div"); wrap.className = "ins-sec";
  wrap.innerHTML = '<div class="ins-sec-h">Modal</div>';
  if (n.kind === "modal") {
    var nr = document.createElement("div"); nr.className = "rowline";
    var nl = document.createElement("label"); nl.textContent = "name"; nl.title = "Stable id used by modal.open(name) and modal.close(name)."; nr.appendChild(nl);
    var ni = document.createElement("input"); ni.value = n.name || ""; ni.placeholder = "dialog"; ni.spellcheck = false;
    ni.onchange = function(){ upd("name", ni.value.trim()); };
    nr.appendChild(ni); wrap.appendChild(nr);

    var br = document.createElement("div"); br.className = "rowline";
    var ob = document.createElement("button"); ob.className = "btn"; ob.style.fontSize = "11px"; ob.textContent = "Open in preview";
    ob.title = "Open this modal in the live preview without leaving the designer.";
    ob.onclick = function() {
      var ifr = document.getElementById("prev");
      var w = ifr && ifr.contentWindow;
      if (w && typeof w.__sui_modal_open === "function") w.__sui_modal_open(n.name || "dialog");
      else setStatus("modal runtime not ready; refresh the preview");
    };
    br.appendChild(ob); wrap.appendChild(br);

    var hint = document.createElement("div"); hint.className = "hint";
    hint.textContent = "Open from app code with modal.open(\"" + (n.name || "dialog") + "\"). Modal buttons close the dialog by default; set close: false to keep it open.";
    wrap.appendChild(hint);
  } else {
    var vr = document.createElement("div"); vr.className = "rowline";
    var vl = document.createElement("label"); vl.textContent = "variant"; vr.appendChild(vl);
    var vs = document.createElement("select");
    ["", "primary", "danger"].forEach(function(v){
      var o = document.createElement("option"); o.value = v; o.textContent = v || "(default)";
      if ((n.variant || "") === v) o.selected = true;
      vs.appendChild(o);
    });
    vs.onchange = function(){ upd("variant", vs.value); };
    vr.appendChild(vs); wrap.appendChild(vr);

    var cr = document.createElement("div"); cr.className = "rowline";
    var cl = document.createElement("label"); cl.textContent = "close"; cl.title = "Close the modal after this button is clicked."; cr.appendChild(cl);
    cr.appendChild(mkSwitch(n.close !== false, false, function(on){ upd("close", on); }));
    wrap.appendChild(cr);

    var bh = document.createElement("div"); bh.className = "hint";
    bh.textContent = "Modal button onClick runs before the default close. Set close: false when the action should keep the dialog open.";
    wrap.appendChild(bh);
  }
  return wrap;
}

function comboboxState(n) {
  if (!n || !n.optionsExpr || !n.options) return null;
  for (var i = 0; i < root.states.length; i++) if (root.states[i].name === n.options) return root.states[i];
  return null;
}
function comboboxStateString(s) {
  if (!s || s.init == null || s.init === "") return "";
  try {
    var v = JSON.parse(s.init);
    return typeof v === "string" ? v : null;
  } catch (e) {
    return null;
  }
}
// comboboxSplitRaw splits a runtime option string on unescaped commas while
// preserving the escape characters for round-tripping back into the state.
function comboboxSplitRaw(text) {
  var out = [], cur = "", escaped = false;
  String(text || "").split("").forEach(function(ch) {
    if (escaped) {
      cur += "\\" + ch;
      escaped = false;
    } else if (ch === "\\") {
      escaped = true;
    } else if (ch === ",") {
      if (cur.trim()) out.push(cur.trim());
      cur = "";
    } else {
      cur += ch;
    }
  });
  if (escaped) cur += "\\";
  if (cur.trim()) out.push(cur.trim());
  return out;
}
function comboboxDisplayText(raw) {
  return String(raw || "").replace(/\\([\s\S])/g, "$1");
}
function comboboxOptionList(n) {
  if (n.optionsExpr) {
    var s = comboboxState(n);
    if (!s) return [];
    var text = comboboxStateString(s);
    if (text == null) return null;
    return comboboxSplitRaw(text);
  }
  return comboboxSplitRaw(n.options);
}
function comboboxWriteOptions(n, list) {
  var text = list.join(",");
  if (n.optionsExpr) {
    var s = comboboxState(n);
    if (!s) return;
    s.init = JSON.stringify(text);
  } else {
    n.options = text;
  }
}
var TREE_STYLE_PRESETS = [
  { key: "default", label: "Default", css: "--sui-tree-gap: 2px; --sui-tree-row-padding: 2px 5px; --sui-tree-row-radius: 5px; --sui-tree-row-bg: transparent; --sui-tree-hover-bg: #2a3143; --sui-tree-selected-bg: #4c5cf747; --sui-tree-selected-fg: #ffffff; --sui-tree-title-color: #e6e6ef; --sui-tree-title-size: 13px; --sui-tree-panel-bg: transparent; --sui-tree-panel-border: 0; --sui-tree-panel-radius: 0px; --sui-tree-panel-padding: 0px; --sui-tree-indent: 16px;" },
  { key: "compact", label: "Compact", css: "--sui-tree-gap: 0px; --sui-tree-row-padding: 0px 4px; --sui-tree-row-radius: 4px; --sui-tree-hover-bg: #252b38; --sui-tree-selected-bg: #4c5cf738; --sui-tree-selected-fg: #ffffff; --sui-tree-title-size: 12px; --sui-tree-indent: 12px;" },
  { key: "bordered", label: "Bordered", css: "--sui-tree-gap: 1px; --sui-tree-row-padding: 4px 6px; --sui-tree-row-radius: 4px; --sui-tree-hover-bg: #252b38; --sui-tree-selected-bg: #3949d6; --sui-tree-selected-fg: #ffffff; --sui-tree-panel-bg: #191d25; --sui-tree-panel-border: 1px solid #3a4152; --sui-tree-panel-radius: 8px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "soft", label: "Soft rounded", css: "--sui-tree-gap: 2px; --sui-tree-row-padding: 4px 8px; --sui-tree-row-radius: 9px; --sui-tree-hover-bg: #4c5cf724; --sui-tree-selected-bg: #4c5cf761; --sui-tree-selected-fg: #ffffff; --sui-tree-title-color: #dbe4ff; --sui-tree-panel-bg: #1c2230; --sui-tree-panel-border: 1px solid #303b55; --sui-tree-panel-radius: 12px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "minimal", label: "Minimal", css: "--sui-tree-gap: 0px; --sui-tree-row-padding: 1px 4px; --sui-tree-row-radius: 3px; --sui-tree-hover-bg: transparent; --sui-tree-selected-bg: #4c5cf72e; --sui-tree-selected-fg: #a5b4fc; --sui-tree-title-color: #aab4c8; --sui-tree-title-size: 12px; --sui-tree-panel-bg: transparent; --sui-tree-panel-border: 0; --sui-tree-panel-padding: 0px; --sui-tree-indent: 14px;" },
  { key: "comfortable", label: "Comfortable", css: "--sui-tree-gap: 4px; --sui-tree-row-padding: 6px 9px; --sui-tree-row-radius: 7px; --sui-tree-hover-bg: #2b3448; --sui-tree-selected-bg: #4c5cf74d; --sui-tree-selected-fg: #ffffff; --sui-tree-title-size: 14px; --sui-tree-panel-padding: 4px; --sui-tree-indent: 20px;" },
  { key: "dense", label: "Dense", css: "--sui-tree-gap: 0px; --sui-tree-row-padding: 0px 3px; --sui-tree-row-radius: 2px; --sui-tree-hover-bg: #232a38; --sui-tree-selected-bg: #4c5cf72e; --sui-tree-selected-fg: #dbe4ff; --sui-tree-title-size: 11px; --sui-tree-indent: 10px;" },
  { key: "neon", label: "Neon", css: "--sui-tree-gap: 2px; --sui-tree-row-padding: 3px 7px; --sui-tree-row-radius: 6px; --sui-tree-hover-bg: #0e3b44; --sui-tree-selected-bg: #00e5ff; --sui-tree-selected-fg: #04161b; --sui-tree-title-color: #7df9ff; --sui-tree-panel-bg: #08161d; --sui-tree-panel-border: 1px solid #00e5ff; --sui-tree-panel-radius: 10px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "forest", label: "Forest", css: "--sui-tree-gap: 2px; --sui-tree-row-padding: 3px 7px; --sui-tree-row-radius: 6px; --sui-tree-hover-bg: #143d2b; --sui-tree-selected-bg: #1f7a4d; --sui-tree-selected-fg: #ecfdf5; --sui-tree-title-color: #bbf7d0; --sui-tree-panel-bg: #0f241a; --sui-tree-panel-border: 1px solid #24543a; --sui-tree-panel-radius: 10px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "amber", label: "Amber", css: "--sui-tree-gap: 2px; --sui-tree-row-padding: 3px 7px; --sui-tree-row-radius: 6px; --sui-tree-hover-bg: #4a3510; --sui-tree-selected-bg: #b45309; --sui-tree-selected-fg: #fffbeb; --sui-tree-title-color: #fde68a; --sui-tree-panel-bg: #241a0c; --sui-tree-panel-border: 1px solid #5a4218; --sui-tree-panel-radius: 10px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "ocean", label: "Ocean", css: "--sui-tree-gap: 2px; --sui-tree-row-padding: 3px 7px; --sui-tree-row-radius: 6px; --sui-tree-hover-bg: #123a5c; --sui-tree-selected-bg: #0e7490; --sui-tree-selected-fg: #ecfeff; --sui-tree-title-color: #a5f3fc; --sui-tree-panel-bg: #0b1f33; --sui-tree-panel-border: 1px solid #1c4b6e; --sui-tree-panel-radius: 10px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "highcontrast", label: "High contrast", css: "--sui-tree-gap: 1px; --sui-tree-row-padding: 4px 8px; --sui-tree-row-radius: 0px; --sui-tree-hover-bg: #333333; --sui-tree-selected-bg: #ffff00; --sui-tree-selected-fg: #000000; --sui-tree-title-color: #ffffff; --sui-tree-title-size: 14px; --sui-tree-panel-bg: #000000; --sui-tree-panel-border: 2px solid #ffffff; --sui-tree-panel-radius: 0px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "outline", label: "Outline", css: "--sui-tree-gap: 3px; --sui-tree-row-padding: 4px 8px; --sui-tree-row-radius: 6px; --sui-tree-row-bg: transparent; --sui-tree-hover-bg: #252d3d; --sui-tree-selected-bg: transparent; --sui-tree-selected-fg: #93c5fd; --sui-tree-panel-bg: transparent; --sui-tree-panel-border: 1px solid #3a4152; --sui-tree-panel-radius: 8px; --sui-tree-panel-padding: 8px; --sui-tree-indent: 18px;" },
  { key: "mono", label: "Mono", css: "--sui-tree-gap: 1px; --sui-tree-row-padding: 2px 4px; --sui-tree-row-radius: 2px; --sui-tree-hover-bg: #2a2a2a; --sui-tree-selected-bg: #d4d4d4; --sui-tree-selected-fg: #111111; --sui-tree-title-color: #e5e5e5; --sui-tree-panel-bg: #171717; --sui-tree-panel-border: 1px solid #404040; --sui-tree-panel-radius: 0px; --sui-tree-panel-padding: 6px; --sui-tree-indent: 14px;" },
  { key: "candy", label: "Candy", css: "--sui-tree-gap: 3px; --sui-tree-row-padding: 4px 9px; --sui-tree-row-radius: 999px; --sui-tree-hover-bg: #4a1d54; --sui-tree-selected-bg: #db2777; --sui-tree-selected-fg: #fdf2f8; --sui-tree-title-color: #f9a8d4; --sui-tree-panel-bg: #2a1033; --sui-tree-panel-border: 1px solid #6b21a8; --sui-tree-panel-radius: 16px; --sui-tree-panel-padding: 10px; --sui-tree-indent: 20px;" },
  { key: "cards", label: "Cards", css: "--sui-tree-gap: 6px; --sui-tree-row-padding: 7px 10px; --sui-tree-row-radius: 9px; --sui-tree-row-bg: #202838; --sui-tree-hover-bg: #2b3549; --sui-tree-selected-bg: #4c5cf7; --sui-tree-selected-fg: #ffffff; --sui-tree-title-color: #dbe4ff; --sui-tree-panel-bg: transparent; --sui-tree-panel-border: 0; --sui-tree-panel-padding: 0px; --sui-tree-indent: 20px;" }
];
function treePresetCss(key) {
  for (var i = 0; i < TREE_STYLE_PRESETS.length; i++) {
    if (TREE_STYLE_PRESETS[i].key === key) return TREE_STYLE_PRESETS[i].css;
  }
  return "";
}
function treePresetStyleExpr(key) {
  var css = treePresetCss(key);
  return css ? "cssSafe(" + JSON.stringify(css) + ")" : "";
}
function treePresetKey(n) {
  var cur = n && n.styleExpr ? String(n.styleExpr) : "";
  for (var i = 0; i < TREE_STYLE_PRESETS.length; i++) {
    if (cur === treePresetStyleExpr(TREE_STYLE_PRESETS[i].key)) return TREE_STYLE_PRESETS[i].key;
  }
  return "custom";
}
function treeControls(n) {
  var rows = [];
  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Selecting a treeitem writes its key to value state. Editable items can be renamed with double-click or F2; the key is never changed.";
  rows.push(hint);
  var nameRow = document.createElement("div"); nameRow.className = "rowline";
  var nameLabel = document.createElement("label"); nameLabel.textContent = "tree name"; nameRow.appendChild(nameLabel);
  var nameInput = document.createElement("input"); nameInput.value = n.name || ""; nameInput.placeholder = "files"; nameInput.spellcheck = false;
  nameInput.oninput = function(ev) { snap("tree name"); n.name = ev.target.value.trim(); setOut(); autoPreview(); };
  nameRow.appendChild(nameInput); rows.push(nameRow);
  var nameHint = document.createElement("div"); nameHint.className = "hint";
  nameHint.textContent = "Runtime target used by tree.importJson/importXml/exportJson/exportXml and the file helpers.";
  rows.push(nameHint);
  var ioRow = document.createElement("div"); ioRow.className = "rowline";
  var ioLabel = document.createElement("label"); ioLabel.textContent = "JSON/XML data"; ioRow.appendChild(ioLabel);
  [["Import JSON", "json", 1], ["Import XML", "xml", 1], ["Export JSON", "json", 0], ["Export XML", "xml", 0]].forEach(function(cfg) {
    var b = document.createElement("button"); b.className = "btn"; b.textContent = cfg[0]; b.style.fontSize = "10px"; b.style.padding = "2px 5px";
    b.onclick = function() { if (cfg[2]) treePickData(n, cfg[1]); else treeDownloadData(n, cfg[1]); };
    ioRow.appendChild(b);
  });
  rows.push(ioRow);
  var ioHint = document.createElement("div"); ioHint.className = "hint";
  ioHint.textContent = "Design-time import replaces the tree items in this source. Runtime import/export is available through the tree.* core module.";
  rows.push(ioHint);
  function stateField(label, value, placeholder, key) {
    var r = document.createElement("div"); r.className = "rowline";
    var l = document.createElement("label"); l.textContent = label; r.appendChild(l);
    var i = document.createElement("input"); i.value = value || ""; i.placeholder = placeholder || ""; i.spellcheck = false;
    i.oninput = function(ev) { snap("set tree " + key); n[key] = ev.target.value; setOut(); autoPreview(); };
    r.appendChild(i); rows.push(r);
  }
  stateField("value state", n.value, "state name", "value");
  var styleRow = document.createElement("div"); styleRow.className = "rowline";
  var styleLabel = document.createElement("label"); styleLabel.textContent = "style preset"; styleRow.appendChild(styleLabel);
  var styleSel = document.createElement("select");
  TREE_STYLE_PRESETS.forEach(function(p) {
    var o = document.createElement("option"); o.value = p.key; o.textContent = p.label;
    if (treePresetKey(n) === p.key) o.selected = true;
    styleSel.appendChild(o);
  });
  var customOpt = document.createElement("option"); customOpt.value = "custom"; customOpt.textContent = "Custom / edit style expr";
  if (treePresetKey(n) === "custom") customOpt.selected = true;
  styleSel.appendChild(customOpt);
  styleSel.onchange = function() {
    if (styleSel.value === "custom") {
      styleSel.value = treePresetKey(n);
      return;
    }
    snap("tree style preset");
    n.styleExpr = treePresetStyleExpr(styleSel.value);
    setOut();
    renderInspector();
    autoPreview();
  };
  styleRow.appendChild(styleSel); rows.push(styleRow);
  var styleHint = document.createElement("div"); styleHint.className = "hint";
  styleHint.textContent = "Presets write CSS variables into style expr. Edit the style expr field in Appearance for full control.";
  rows.push(styleHint);
  var iconRow = document.createElement("div"); iconRow.className = "rowline";
  var iconLabel = document.createElement("label"); iconLabel.textContent = "rename icon"; iconRow.appendChild(iconLabel);
  var iconInput = document.createElement("input"); iconInput.value = n.renameIcon || ""; iconInput.placeholder = "\u270e (default)"; iconInput.maxLength = 16;
  iconInput.oninput = function(ev) { snap("tree rename icon"); n.renameIcon = ev.target.value; setOut(); autoPreview(); };
  iconRow.appendChild(iconInput); rows.push(iconRow);
  var iconShow = document.createElement("div"); iconShow.className = "rowline";
  var iconShowLabel = document.createElement("label"); iconShowLabel.textContent = "show rename icon"; iconShow.appendChild(iconShowLabel);
  iconShow.appendChild(mkSwitch(n.showRenameIcon !== false, false, function(on){
    snap("tree show rename icon"); n.showRenameIcon = on; setOut(); autoPreview();
  }));
  rows.push(iconShow);
  var er = document.createElement("div"); er.className = "rowline";
  var el = document.createElement("label"); el.textContent = "editable"; er.appendChild(el);
  er.appendChild(mkSwitch(!!n.editable, false, function(on){ snap("tree editable"); n.editable = on; setOut(); autoPreview(); }));
  rows.push(er);
  var dr = document.createElement("div"); dr.className = "rowline";
  var dl = document.createElement("label"); dl.textContent = "editable depth"; dr.appendChild(dl);
  var di = document.createElement("input"); di.type = "number"; di.min = "0"; di.step = "1";
  di.value = n.editableDepth || ""; di.placeholder = "0 = all levels";
  di.oninput = function(ev) { snap("tree editable depth"); n.editableDepth = ev.target.value; setOut(); autoPreview(); };
  dr.appendChild(di); rows.push(dr);
  stateField("rename key state", n.renameKey, "state name", "renameKey");
  stateField("rename value state", n.renameValue, "state name", "renameValue");
  return acc("Tree", rows);
}
function treeRawLabel(raw) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return raw === null || raw === undefined ? "" : String(raw);
  var keys = ["label", "text", "name", "title"];
  for (var i = 0; i < keys.length; i++) {
    if (raw[keys[i]] !== undefined && raw[keys[i]] !== null && raw[keys[i]] !== "") return String(raw[keys[i]]);
  }
  return "";
}
function treeBool(v, fallback) {
  if (v === undefined || v === null || v === "") return fallback;
  if (typeof v === "boolean") return v;
  var s = String(v).trim().toLowerCase();
  if (s === "true" || s === "1" || s === "yes") return true;
  if (s === "false" || s === "0" || s === "no") return false;
  return fallback;
}
function treeDataRoot(data) {
  if (Array.isArray(data)) return data;
  if (data && typeof data === "object") {
    if (Array.isArray(data.nodes)) return data.nodes;
    if (Array.isArray(data.items)) return data.items;
    if (Array.isArray(data.tree)) return data.tree;
    if (Array.isArray(data.children)) return treeRawLabel(data) ? [data] : data.children;
    if (treeRawLabel(data)) return [data];
  }
  throw new Error("expected an array of nodes or an object with nodes/children");
}
function treeDataModel(raw, path, depth, ctx) {
  if (depth > 64) throw new Error("nesting is too deep (limit 64)");
  if (ctx.count >= 5000) throw new Error("too many nodes (limit 5000)");
  ctx.count++;
  var obj = raw && typeof raw === "object" && !Array.isArray(raw) ? raw : null;
  var label = String(treeRawLabel(raw) || "").trim();
  var key = obj && obj.key !== undefined && obj.key !== null ? String(obj.key).trim() : "";
  if (!label) label = key || String(path);
  var n = newTreeItemNode(label);
  n.key = key;
  if (obj) {
    n.open = treeBool(obj.open, false);
    if (obj.editable !== undefined) n.editable = treeBool(obj.editable, false);
    if (obj.showRenameIcon !== undefined) n.showRenameIcon = treeBool(obj.showRenameIcon, false);
    if (obj.renameIcon !== undefined) n.renameIcon = String(obj.renameIcon);
    var kids = obj.children !== undefined ? obj.children : obj.items;
    if (kids !== undefined && kids !== null && !Array.isArray(kids)) throw new Error("children must be an array");
    if (kids) {
      n.children = [];
      for (var i = 0; i < kids.length; i++) n.children.push(treeDataModel(kids[i], path ? path + "." + i : String(i), depth + 1, ctx));
    }
  }
  return n;
}
function treeXmlToRaw(el) {
  var o = {};
  var label = el.getAttribute("label") || el.getAttribute("text") || el.getAttribute("name") || el.getAttribute("title") || "";
  if (!label) {
    for (var t = 0; t < el.childNodes.length; t++) {
      if (el.childNodes[t].nodeType === 3) label += el.childNodes[t].nodeValue || "";
    }
    label = label.trim();
  }
  if (label) o.label = label;
  var key = el.getAttribute("key");
  if (key) o.key = key;
  var attrs = ["open", "editable", "showRenameIcon", "renameIcon"];
  for (var i = 0; i < attrs.length; i++) {
    var v = el.getAttribute(attrs[i]);
    if (v !== null) o[attrs[i]] = v;
  }
  var kids = [];
  for (var j = 0; j < el.childNodes.length; j++) {
    if (el.childNodes[j].nodeType === 1) kids.push(treeXmlToRaw(el.childNodes[j]));
  }
  if (kids.length) o.children = kids;
  return o;
}
function treeDataFromXml(text) {
  if (/<!DOCTYPE|<!ENTITY/i.test(text)) throw new Error("DTD and entities are not allowed");
  var doc = new DOMParser().parseFromString(text, "application/xml");
  var pe = doc.getElementsByTagName("parsererror");
  if (pe.length) throw new Error("invalid XML");
  var root = doc.documentElement;
  if (!root) throw new Error("empty XML");
  var tag = (root.localName || root.nodeName || "").toLowerCase(), raw = [];
  if (tag === "tree") {
    for (var i = 0; i < root.childNodes.length; i++) {
      if (root.childNodes[i].nodeType === 1) raw.push(treeXmlToRaw(root.childNodes[i]));
    }
  } else {
    raw.push(treeXmlToRaw(root));
  }
  return raw;
}
function treeImportData(n, fmt, text) {
  try {
    var raw;
    if (fmt === "xml") raw = treeDataFromXml(text);
    else {
      var data = JSON.parse(text);
      raw = treeDataRoot(data);
    }
    var ctx = { count: 0 };
    var nodes = [];
    for (var i = 0; i < raw.length; i++) nodes.push(treeDataModel(raw[i], String(i), 0, ctx));
    snap("import tree " + fmt);
    n.children = nodes;
    setOut(); renderTree(); renderView(); renderInspector(); autoPreview();
    setStatus("imported " + ctx.count + " tree nodes from " + fmt.toUpperCase());
  } catch (e) {
    showMsg("Tree import failed", (e && e.message) ? e.message : String(e));
  }
}
function treePickData(n, fmt) {
  var inp = document.createElement("input");
  inp.type = "file";
  inp.accept = fmt === "xml" ? ".xml,application/xml,text/xml" : ".json,application/json";
  inp.style.cssText = "position:fixed;left:-9999px;top:0;width:1px;height:1px;opacity:0";
  inp.onchange = function() {
    var f = inp.files && inp.files[0];
    if (!f) { if (inp.parentNode) inp.parentNode.removeChild(inp); return; }
    var r = new FileReader();
    r.onload = function() { treeImportData(n, fmt, String(r.result || "")); };
    r.onerror = function() { showMsg("Tree import failed", "cannot read the selected file"); };
    r.readAsText(f);
    if (inp.parentNode) inp.parentNode.removeChild(inp);
  };
  document.body.appendChild(inp);
  inp.click();
}
function treeModelNode(n) {
  var o = { label: n.text || "", key: n.key || "" };
  if (n.children && n.children.length) {
    o.open = !!n.open;
    o.children = [];
    for (var i = 0; i < n.children.length; i++) o.children.push(treeModelNode(n.children[i]));
  }
  if (n.editable === true || n.editable === false) o.editable = n.editable;
  if (n.showRenameIcon === true || n.showRenameIcon === false) o.showRenameIcon = n.showRenameIcon;
  if (n.renameIcon) o.renameIcon = n.renameIcon;
  return o;
}
function treeXmlEscape(v) {
  return String(v === null || v === undefined ? "" : v).replace(/&/g, "&amp;").replace(/</g, "&lt;")
    .replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&apos;");
}
function treeModelXml(items, indent) {
  var out = [];
  for (var i = 0; i < items.length; i++) {
    var n = items[i], attrs = " label=\"" + treeXmlEscape(n.label) + "\" key=\"" + treeXmlEscape(n.key) + "\"";
    if (n.open !== undefined) attrs += " open=\"" + (n.open ? "true" : "false") + "\"";
    if (n.editable !== undefined) attrs += " editable=\"" + (n.editable ? "true" : "false") + "\"";
    if (n.showRenameIcon !== undefined) attrs += " showRenameIcon=\"" + (n.showRenameIcon ? "true" : "false") + "\"";
    if (n.renameIcon) attrs += " renameIcon=\"" + treeXmlEscape(n.renameIcon) + "\"";
    if (n.children && n.children.length) {
      out.push(indent + "<item" + attrs + ">");
      out = out.concat(treeModelXml(n.children, indent + "  "));
      out.push(indent + "</item>");
    } else {
      out.push(indent + "<item" + attrs + "/>");
    }
  }
  return out;
}
function treeDownloadData(n, fmt) {
  var items = (n.children || []).map(treeModelNode);
  var text;
  if (fmt === "xml") {
    var lines = treeModelXml(items, "  ");
    text = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<tree name=\"" + treeXmlEscape(n.name || n.text || "tree") + "\">\n" +
      (lines.length ? lines.join("\n") + "\n" : "") + "</tree>\n";
  } else {
    text = JSON.stringify(items, null, 2);
  }
  try {
    var blob = new Blob([text], { type: fmt === "xml" ? "application/xml;charset=utf-8" : "application/json;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    var base = String(n.name || "tree").replace(/[\\/:*?"<>|\x00-\x1f]/g, "_").trim() || "tree";
    a.href = url; a.download = base + "." + fmt;
    a.style.display = "none"; document.body.appendChild(a); a.click();
    if (a.parentNode) a.parentNode.removeChild(a);
    setTimeout(function(){ try { URL.revokeObjectURL(url); } catch(e) {} }, 0);
    setStatus("exported tree " + fmt.toUpperCase());
  } catch (e) {
    showMsg("Tree export failed", String(e));
  }
}
var TREE_ITEM_STYLE_PRESETS = [
  { key: "inherit", label: "Inherit tree style", css: "" },
  { key: "muted", label: "Muted", css: "--sui-tree-row-bg: #1b2130; --sui-tree-selected-bg: #334155; --sui-tree-selected-fg: #e2e8f0;" },
  { key: "accent", label: "Accent", css: "--sui-tree-hover-bg: #26346b; --sui-tree-selected-bg: #4c5cf7; --sui-tree-selected-fg: #ffffff;" },
  { key: "success", label: "Success", css: "--sui-tree-hover-bg: #123f2c; --sui-tree-selected-bg: #16794a; --sui-tree-selected-fg: #ecfdf5;" },
  { key: "warning", label: "Warning", css: "--sui-tree-hover-bg: #4a3510; --sui-tree-selected-bg: #b45309; --sui-tree-selected-fg: #fffbeb;" },
  { key: "danger", label: "Danger", css: "--sui-tree-hover-bg: #4c1d1d; --sui-tree-selected-bg: #b91c1c; --sui-tree-selected-fg: #fef2f2;" },
  { key: "folder", label: "Folder", css: "--sui-tree-row-bg: #1a253b; --sui-tree-row-radius: 7px; --sui-tree-selected-bg: #2a3a5f; --sui-tree-selected-fg: #dbeafe;" },
  { key: "compactrow", label: "Compact row", css: "--sui-tree-row-padding: 0px 4px; --sui-tree-row-radius: 3px;" }
];
var CHART_STYLE_PRESETS = [
  { key: "default", label: "Default", css: "--sui-chart-bg: transparent; --sui-chart-fg: #e6e6ef; --sui-chart-title-color: #e6e6ef; --sui-chart-muted: #8fa0b8; --sui-chart-grid: #2c3444; --sui-chart-axis: #3a4152; --sui-chart-series-1: #4c5cf7; --sui-chart-series-2: #22c55e; --sui-chart-series-3: #f59e0b; --sui-chart-series-4: #ef4444; --sui-chart-series-5: #06b6d4; --sui-chart-series-6: #a855f7; --sui-chart-series-7: #84cc16; --sui-chart-series-8: #ec4899;" },
  { key: "ocean", label: "Ocean", css: "--sui-chart-bg: #0b1f33; --sui-chart-fg: #ecfeff; --sui-chart-title-color: #a5f3fc; --sui-chart-muted: #7dd3fc; --sui-chart-grid: #164e63; --sui-chart-axis: #0e7490; --sui-chart-series-1: #22d3ee; --sui-chart-series-2: #38bdf8; --sui-chart-series-3: #2dd4bf; --sui-chart-series-4: #818cf8; --sui-chart-series-5: #67e8f9; --sui-chart-series-6: #a5b4fc; --sui-chart-series-7: #14b8a6; --sui-chart-series-8: #7dd3fc; --sui-chart-border: 1px solid #164e63; --sui-chart-radius: 10px; --sui-chart-padding: 8px;" },
  { key: "forest", label: "Forest", css: "--sui-chart-bg: #0f241a; --sui-chart-fg: #ecfdf5; --sui-chart-title-color: #bbf7d0; --sui-chart-muted: #86efac; --sui-chart-grid: #24543a; --sui-chart-axis: #2f6b49; --sui-chart-series-1: #22c55e; --sui-chart-series-2: #84cc16; --sui-chart-series-3: #14b8a6; --sui-chart-series-4: #eab308; --sui-chart-series-5: #10b981; --sui-chart-series-6: #65a30d; --sui-chart-series-7: #4ade80; --sui-chart-series-8: #a3e635; --sui-chart-border: 1px solid #24543a; --sui-chart-radius: 10px; --sui-chart-padding: 8px;" },
  { key: "sunset", label: "Sunset", css: "--sui-chart-bg: #2a1420; --sui-chart-fg: #fff1f2; --sui-chart-title-color: #fecdd3; --sui-chart-muted: #fda4af; --sui-chart-grid: #5c243b; --sui-chart-axis: #7f2947; --sui-chart-series-1: #fb7185; --sui-chart-series-2: #fb923c; --sui-chart-series-3: #f59e0b; --sui-chart-series-4: #e879f9; --sui-chart-series-5: #f97316; --sui-chart-series-6: #f43f5e; --sui-chart-series-7: #facc15; --sui-chart-series-8: #d946ef; --sui-chart-border: 1px solid #5c243b; --sui-chart-radius: 12px; --sui-chart-padding: 8px;" },
  { key: "neon", label: "Neon", css: "--sui-chart-bg: #08161d; --sui-chart-fg: #e0ffff; --sui-chart-title-color: #7df9ff; --sui-chart-muted: #67e8f9; --sui-chart-grid: #0e3b44; --sui-chart-axis: #00e5ff; --sui-chart-series-1: #00e5ff; --sui-chart-series-2: #ff2d95; --sui-chart-series-3: #a3ff12; --sui-chart-series-4: #d946ef; --sui-chart-series-5: #22d3ee; --sui-chart-series-6: #f472b6; --sui-chart-series-7: #bef264; --sui-chart-series-8: #c084fc; --sui-chart-border: 1px solid #00e5ff; --sui-chart-radius: 10px; --sui-chart-padding: 8px; --sui-chart-line-width: 2.6; --sui-chart-title-size: 15px;" },
  { key: "mono", label: "Mono", css: "--sui-chart-bg: #171717; --sui-chart-fg: #f5f5f5; --sui-chart-title-color: #ffffff; --sui-chart-muted: #a3a3a3; --sui-chart-grid: #404040; --sui-chart-axis: #737373; --sui-chart-series-1: #e5e5e5; --sui-chart-series-2: #a3a3a3; --sui-chart-series-3: #737373; --sui-chart-series-4: #d4d4d4; --sui-chart-series-5: #525252; --sui-chart-series-6: #fafafa; --sui-chart-series-7: #8a8a8a; --sui-chart-series-8: #c4c4c4; --sui-chart-border: 1px solid #404040; --sui-chart-radius: 4px; --sui-chart-padding: 6px;" },
  { key: "pastel", label: "Pastel", css: "--sui-chart-bg: #fdf8ff; --sui-chart-fg: #3b2f4a; --sui-chart-title-color: #4c1d95; --sui-chart-muted: #7c6a8f; --sui-chart-grid: #eadcf5; --sui-chart-axis: #cbb8dd; --sui-chart-series-1: #a78bfa; --sui-chart-series-2: #6ee7b7; --sui-chart-series-3: #fcd34d; --sui-chart-series-4: #f9a8d4; --sui-chart-series-5: #7dd3fc; --sui-chart-series-6: #c4b5fd; --sui-chart-series-7: #86efac; --sui-chart-series-8: #fda4af; --sui-chart-border: 1px solid #eadcf5; --sui-chart-radius: 14px; --sui-chart-padding: 10px; --sui-chart-scroll-thumb: #c4b5fd; --sui-chart-scroll-track: #f5edff; --sui-chart-control-bg: #f7efff; --sui-chart-control-fg: #3b2f4a; --sui-chart-control-border: #cbb8dd; --sui-chart-control-hover-bg: #f0e2ff; --sui-chart-control-hover-border: #a78bfa; --sui-chart-error: #b91c1c;" },
  { key: "highcontrast", label: "High contrast", css: "--sui-chart-bg: #000000; --sui-chart-fg: #ffffff; --sui-chart-title-color: #ffff00; --sui-chart-muted: #ffffff; --sui-chart-grid: #666666; --sui-chart-axis: #ffffff; --sui-chart-series-1: #ffff00; --sui-chart-series-2: #00ffff; --sui-chart-series-3: #ff00ff; --sui-chart-series-4: #00ff00; --sui-chart-series-5: #ff8800; --sui-chart-series-6: #ffffff; --sui-chart-series-7: #ff4444; --sui-chart-series-8: #66ccff; --sui-chart-border: 2px solid #ffffff; --sui-chart-radius: 0; --sui-chart-padding: 8px; --sui-chart-label-size: 12px; --sui-chart-title-size: 16px; --sui-chart-control-bg: #000000; --sui-chart-control-fg: #ffffff; --sui-chart-control-border: #ffffff; --sui-chart-control-hover-bg: #ffff00; --sui-chart-control-hover-fg: #000000; --sui-chart-control-hover-border: #ffff00; --sui-chart-control-radius: 0;" },
  { key: "amber", label: "Amber", css: "--sui-chart-bg: #241a0c; --sui-chart-fg: #fffbeb; --sui-chart-title-color: #fde68a; --sui-chart-muted: #fcd34d; --sui-chart-grid: #5a4218; --sui-chart-axis: #7c5a1e; --sui-chart-series-1: #f59e0b; --sui-chart-series-2: #fbbf24; --sui-chart-series-3: #ea580c; --sui-chart-series-4: #d97706; --sui-chart-series-5: #facc15; --sui-chart-series-6: #b45309; --sui-chart-series-7: #fde047; --sui-chart-series-8: #fb923c; --sui-chart-border: 1px solid #5a4218; --sui-chart-radius: 10px; --sui-chart-padding: 8px;" },
  { key: "candy", label: "Candy", css: "--sui-chart-bg: #2a1033; --sui-chart-fg: #fdf2f8; --sui-chart-title-color: #f9a8d4; --sui-chart-muted: #f0abfc; --sui-chart-grid: #6b21a8; --sui-chart-axis: #86198f; --sui-chart-series-1: #ec4899; --sui-chart-series-2: #a855f7; --sui-chart-series-3: #38bdf8; --sui-chart-series-4: #f472b6; --sui-chart-series-5: #c084fc; --sui-chart-series-6: #22d3ee; --sui-chart-series-7: #fb7185; --sui-chart-series-8: #d946ef; --sui-chart-border: 1px solid #6b21a8; --sui-chart-radius: 16px; --sui-chart-padding: 10px;" },
  { key: "slate", label: "Slate", css: "--sui-chart-bg: #1e293b; --sui-chart-fg: #f1f5f9; --sui-chart-title-color: #e2e8f0; --sui-chart-muted: #94a3b8; --sui-chart-grid: #334155; --sui-chart-axis: #475569; --sui-chart-series-1: #60a5fa; --sui-chart-series-2: #34d399; --sui-chart-series-3: #fbbf24; --sui-chart-series-4: #f87171; --sui-chart-series-5: #a78bfa; --sui-chart-series-6: #22d3ee; --sui-chart-series-7: #a3e635; --sui-chart-series-8: #fb7185; --sui-chart-border: 1px solid #334155; --sui-chart-radius: 9px; --sui-chart-padding: 8px;" },
  { key: "light", label: "Light", css: "--sui-chart-bg: #ffffff; --sui-chart-fg: #1f2937; --sui-chart-title-color: #111827; --sui-chart-muted: #6b7280; --sui-chart-grid: #e5e7eb; --sui-chart-axis: #9ca3af; --sui-chart-series-1: #2563eb; --sui-chart-series-2: #16a34a; --sui-chart-series-3: #d97706; --sui-chart-series-4: #dc2626; --sui-chart-series-5: #0891b2; --sui-chart-series-6: #7c3aed; --sui-chart-series-7: #65a30d; --sui-chart-series-8: #db2777; --sui-chart-border: 1px solid #e5e7eb; --sui-chart-radius: 10px; --sui-chart-padding: 10px; --sui-chart-scroll-thumb: #9ca3af; --sui-chart-scroll-track: #f3f4f6; --sui-chart-control-bg: #f3f4f6; --sui-chart-control-fg: #111827; --sui-chart-control-border: #9ca3af; --sui-chart-control-hover-bg: #e5e7eb; --sui-chart-control-hover-border: #2563eb; --sui-chart-error: #b91c1c;" },
  { key: "outline", label: "Outline", css: "--sui-chart-bg: transparent; --sui-chart-fg: #dbe4ff; --sui-chart-title-color: #93c5fd; --sui-chart-muted: #93c5fd; --sui-chart-grid: #2c3444; --sui-chart-axis: #93c5fd; --sui-chart-series-1: #93c5fd; --sui-chart-series-2: #86efac; --sui-chart-series-3: #fde68a; --sui-chart-series-4: #fca5a5; --sui-chart-series-5: #67e8f9; --sui-chart-series-6: #c4b5fd; --sui-chart-series-7: #bef264; --sui-chart-series-8: #f9a8d4; --sui-chart-border: 1px solid #3a4152; --sui-chart-radius: 8px; --sui-chart-padding: 8px;" },
  { key: "minimal", label: "Minimal", css: "--sui-chart-bg: transparent; --sui-chart-fg: #e6e6ef; --sui-chart-title-color: #aab4c8; --sui-chart-muted: #7b879c; --sui-chart-grid: transparent; --sui-chart-axis: #3a4152; --sui-chart-series-1: #93c5fd; --sui-chart-series-2: #86efac; --sui-chart-series-3: #fde68a; --sui-chart-series-4: #fca5a5; --sui-chart-series-5: #67e8f9; --sui-chart-series-6: #c4b5fd; --sui-chart-series-7: #bef264; --sui-chart-series-8: #f9a8d4; --sui-chart-line-width: 1.5; --sui-chart-bar-opacity: .75; --sui-chart-padding: 0;" }
];
function chartPresetCss(key) {
  for (var i = 0; i < CHART_STYLE_PRESETS.length; i++) if (CHART_STYLE_PRESETS[i].key === key) return CHART_STYLE_PRESETS[i].css;
  return CHART_STYLE_PRESETS[0].css;
}
function chartPresetStyleExpr(key) {
  var css = chartPresetCss(key);
  return "cssSafe(" + JSON.stringify(css) + ")";
}
function chartPresetKey(n) {
  var cur = n && n.styleExpr ? String(n.styleExpr) : "";
  for (var i = 0; i < CHART_STYLE_PRESETS.length; i++) {
    if (cur === chartPresetStyleExpr(CHART_STYLE_PRESETS[i].key)) return CHART_STYLE_PRESETS[i].key;
  }
  return "custom";
}
function chartClearParam(n, field) {
  if (n.chartExprs && n.chartExprs[field]) delete n.chartExprs[field];
}
function chartParamExpr(n, field) {
  return n.chartExprs && n.chartExprs[field] ? String(n.chartExprs[field]) : "";
}
var chartDataEditorNode = null;
var chartDataDraft = null;
var chartDataEditorMode = "literal";
var chartDataPanelGeom = null;
// Sample datasets are grouped by data shape. When a chart type change needs a
// new dataset the designer keeps the current literal format and picks the
// sample for the target type, so the replacement is immediately renderable.
var CHART_GRAPH_SAMPLE_JSON = '{"states":[\n  {"id":"input","label":"Input","shape":"rectangle","x":40,"y":80,"w":120,"h":56},\n  {"id":"process","label":"Process","shape":"rectangle","x":240,"y":80,"w":120,"h":56},\n  {"id":"output","label":"Output","shape":"rectangle","x":440,"y":80,"w":120,"h":56}\n],"transitions":[\n  {"from":"input","to":"process","label":"data"},\n  {"from":"process","to":"output","label":"result"}\n]}';
var CHART_SAMPLE_DATA = {
  state: {
    json: CHART_GRAPH_SAMPLE_JSON,
    csv: 'From,To,Label\ninput,process,data\nprocess,output,result',
    xml: '<states>\n  <state id="input" label="Input" shape="rectangle" x="40" y="80" w="120" h="56"/>\n  <state id="process" label="Process" shape="rectangle" x="240" y="80" w="120" h="56"/>\n  <state id="output" label="Output" shape="rectangle" x="440" y="80" w="120" h="56"/>\n  <transition from="input" to="process" label="data"/>\n  <transition from="process" to="output" label="result"/>\n</states>'
  },
  single: {
    json: '[{"label":"Jan","value":12},{"label":"Feb","value":18},{"label":"Mar","value":15},{"label":"Apr","value":24}]',
    csv: 'Label,Value\nJan,12\nFeb,18\nMar,15\nApr,24',
    xml: '<data>\n  <row label="Jan" value="12"/>\n  <row label="Feb" value="18"/>\n  <row label="Mar" value="15"/>\n  <row label="Apr" value="24"/>\n</data>'
  },
  multi: {
    json: '[{"label":"Jan","revenue":12,"cost":7,"profit":5},{"label":"Feb","revenue":18,"cost":9,"profit":9},{"label":"Mar","revenue":15,"cost":6,"profit":9},{"label":"Apr","revenue":24,"cost":11,"profit":13}]',
    csv: 'Label,Revenue,Cost,Profit\nJan,12,7,5\nFeb,18,9,9\nMar,15,6,9\nApr,24,11,13',
    xml: '<data>\n  <row label="Jan" revenue="12" cost="7" profit="5"/>\n  <row label="Feb" revenue="18" cost="9" profit="9"/>\n  <row label="Mar" revenue="15" cost="6" profit="9"/>\n  <row label="Apr" revenue="24" cost="11" profit="13"/>\n</data>'
  },
  heatmap: {
    json: '[{"label":"A","row":1,"value":5},{"label":"B","row":1,"value":8},{"label":"A","row":2,"value":3},{"label":"B","row":2,"value":6}]',
    csv: 'Label,Row,Value\nA,1,5\nB,1,8\nA,2,3\nB,2,6',
    xml: '<data>\n  <row label="A" row="1" value="5"/>\n  <row label="B" row="1" value="8"/>\n  <row label="A" row="2" value="3"/>\n  <row label="B" row="2" value="6"/>\n</data>'
  },
  ohlc: {
    json: '[{"date":"Jan 2","open":10,"high":14,"low":9,"close":13},{"date":"Jan 3","open":13,"high":16,"low":11,"close":15},{"date":"Jan 4","open":15,"high":18,"low":12,"close":14}]',
    csv: 'Date,Open,High,Low,Close\nJan 2,10,14,9,13\nJan 3,13,16,11,15\nJan 4,15,18,12,14',
    xml: '<data>\n  <row label="Jan 2" open="10" high="14" low="9" close="13"/>\n  <row label="Jan 3" open="13" high="16" low="11" close="15"/>\n  <row label="Jan 4" open="15" high="18" low="12" close="14"/>\n</data>'
  },
  boxplot: {
    json: '[{"label":"A","min":2,"q1":5,"median":8,"q3":12,"max":16},{"label":"B","min":3,"q1":6,"median":10,"q3":14,"max":18}]',
    csv: 'Label,Min,Q1,Median,Q3,Max\nA,2,5,8,12,16\nB,3,6,10,14,18',
    xml: '<data>\n  <row label="A" min="2" q1="5" median="8" q3="12" max="16"/>\n  <row label="B" min="3" q1="6" median="10" q3="14" max="18"/>\n</data>'
  },
  bullet: {
    json: '[{"label":"Revenue","value":72,"target":80,"max":100},{"label":"Costs","value":48,"target":45,"max":100}]',
    csv: 'Label,Value,Target,Max\nRevenue,72,80,100\nCosts,48,45,100',
    xml: '<data>\n  <row label="Revenue" value="72" target="80" max="100"/>\n  <row label="Costs" value="48" target="45" max="100"/>\n</data>'
  },
  dumbbell: {
    json: '[{"label":"A","before":12,"after":18},{"label":"B","before":9,"after":15},{"label":"C","before":14,"after":11}]',
    csv: 'Label,Before,After\nA,12,18\nB,9,15\nC,14,11',
    xml: '<data>\n  <row label="A" before="12" after="18"/>\n  <row label="B" before="9" after="15"/>\n  <row label="C" before="14" after="11"/>\n</data>'
  },
  gantt: {
    json: '[{"task":"Design","start":"2026-01-05","end":"2026-01-12"},{"task":"Build","start":"2026-01-12","end":"2026-01-23"},{"task":"Test","start":"2026-01-23","end":"2026-01-30"}]',
    csv: 'Task,Start,End\nDesign,2026-01-05,2026-01-12\nBuild,2026-01-12,2026-01-23\nTest,2026-01-23,2026-01-30',
    xml: '<data>\n  <task label="Design" start="2026-01-05" end="2026-01-12"/>\n  <task label="Build" start="2026-01-12" end="2026-01-23"/>\n  <task label="Test" start="2026-01-23" end="2026-01-30"/>\n</data>'
  },
  histogram: {
    json: '[3,5,5,6,7,8,8,9,12,14,15,18]',
    csv: 'Value\n3\n5\n5\n6\n7\n8\n8\n9\n12\n14\n15\n18',
    xml: '<data>\n  <row value="3"/>\n  <row value="5"/>\n  <row value="5"/>\n  <row value="6"/>\n  <row value="7"/>\n  <row value="8"/>\n  <row value="8"/>\n  <row value="9"/>\n  <row value="12"/>\n  <row value="14"/>\n  <row value="15"/>\n  <row value="18"/>\n</data>'
  },
  waterfall: {
    json: '[{"label":"Start","delta":20},{"label":"Sales","delta":-5},{"label":"Costs","delta":-3},{"label":"Final","delta":8}]',
    csv: 'Label,Delta\nStart,20\nSales,-5\nCosts,-3\nFinal,8',
    xml: '<data>\n  <row label="Start" delta="20"/>\n  <row label="Sales" delta="-5"/>\n  <row label="Costs" delta="-3"/>\n  <row label="Final" delta="8"/>\n</data>'
  }
};
var CHART_TYPE_SAMPLE_FAMILY = {
  bar: "single", barhorizontal: "single", line: "single", step: "single",
  smooth: "single", area: "single", pie: "single", donut: "single",
  scatter: "single", gauge: "single", funnel: "single", treemap: "single",
  barstacked: "multi", bargrouped: "multi", combo: "multi", radar: "multi",
  heatmap: "heatmap", candlestick: "ohlc", boxplot: "boxplot",
  bullet: "bullet", dumbbell: "dumbbell", gantt: "gantt",
  histogram: "histogram", waterfall: "waterfall",
  state: "state", statemachine: "state"
};
function chartSampleDataset(type, format) {
  var family = CHART_TYPE_SAMPLE_FAMILY[type] || "single";
  var set = CHART_SAMPLE_DATA[family] || CHART_SAMPLE_DATA.single;
  var fmt = String(format || "").toLowerCase();
  if (fmt !== "csv" && fmt !== "xml") fmt = "json";
  return set[fmt] || set.json;
}
function chartSampleFormatFor(format) {
  var fmt = String(format || "").toLowerCase();
  return (fmt === "csv" || fmt === "xml") ? fmt : "json";
}

var CHART_TYPE_OPTIONS = [
  ["bar", "Bar"], ["barhorizontal", "Horizontal bar"], ["barstacked", "Stacked bar"],
  ["bargrouped", "Grouped bar"], ["line", "Line"], ["step", "Step line"],
  ["smooth", "Smooth line"], ["area", "Area"], ["combo", "Combo bar + line"],
  ["pie", "Pie"], ["donut", "Donut"], ["scatter", "Scatter"],
  ["histogram", "Histogram"], ["waterfall", "Waterfall"],
  ["radar", "Radar"], ["gauge", "Gauge"], ["funnel", "Funnel"],
  ["heatmap", "Heatmap"], ["candlestick", "Candlestick"],
  ["treemap", "Treemap"], ["boxplot", "Box plot"], ["bullet", "Bullet"],
  ["dumbbell", "Dumbbell"], ["gantt", "Gantt"]
];
var CHART_TYPE_LABEL = {};
CHART_TYPE_OPTIONS.forEach(function(v){ CHART_TYPE_LABEL[v[0]] = v[1]; });
// The block-diagram geometry is now the dedicated Graph control. Keep the
// labels for legacy chartType sources so existing documents stay editable.
CHART_TYPE_LABEL.state = "Block diagram";
CHART_TYPE_LABEL.statemachine = "Block diagram";
function looksNumeric(v) {
  if (typeof v === "number") return isFinite(v);
  var s = String(v === null || v === undefined ? "" : v).trim().replace(/\s+/g, "").replace(",", ".");
  if (!s) return false;
  var n = Number(s);
  return isFinite(n);
}
function chartTypeLabel(t) { return CHART_TYPE_LABEL[t] || t; }
function chartTypeIsMulti(t) { return t === "barstacked" || t === "bargrouped" || t === "combo"; }
function chartTypeSupportsMulti(t) { return chartTypeIsMulti(t) || t === "radar"; }
function chartDataCsvRows(text) {
  var s = String(text || "").replace(/\r\n/g, "\n").replace(/\r/g, "\n");
  var sep = chartDataDetectSep(s.split("\n")[0] || "");
  var rows = [], row = [], cur = "", q = false;
  for (var i = 0; i < s.length; i++) {
    var ch = s.charAt(i);
    if (q) {
      if (ch === '"') {
        if (s.charAt(i + 1) === '"') { cur += '"'; i++; }
        else q = false;
      } else cur += ch;
    } else if (ch === '"') q = true;
    else if (ch === sep) { row.push(cur); cur = ""; }
    else if (ch === "\n") { row.push(cur); cur = ""; rows.push(row); row = []; }
    else cur += ch;
  }
  row.push(cur);
  if (row.length > 1 || row[0] !== "") rows.push(row);
  return rows.filter(function(r){ return r.some(function(c){ return String(c).trim() !== ""; }); });
}
function chartDataColumnProfile(rows, columns, xField, yField) {
  var head = columns || [], count = head.length, r, i, c;
  for (i = 0; i < rows.length; i++) count = Math.max(count, (rows[i] || []).length);
  var x = -1;
  if (xField) {
    for (i = 0; i < head.length; i++) if (String(head[i]).trim().toLowerCase() === String(xField).trim().toLowerCase()) x = i;
  }
  if (x < 0) {
    for (c = 0; c < count; c++) {
      var textCells = 0, numCells = 0;
      for (i = 0; i < rows.length; i++) {
        var v = (rows[i] || [])[c];
        if (v === undefined || String(v).trim() === "") continue;
        if (looksNumeric(v)) numCells++; else textCells++;
      }
      if (textCells > 0 && numCells === 0) { x = c; break; }
    }
  }
  var yIdx = [];
  var specs = chartDataSplitFields(yField);
  for (i = 0; i < specs.length; i++) {
    for (c = 0; c < head.length; c++) {
      if (String(head[c]).trim().toLowerCase() === specs[i].toLowerCase() && c !== x && yIdx.indexOf(c) < 0) yIdx.push(c);
    }
  }
  if (!yIdx.length) {
    for (c = 0; c < count; c++) {
      if (c === x) continue;
      for (i = 0; i < rows.length; i++) if (looksNumeric((rows[i] || [])[c])) { yIdx.push(c); break; }
    }
  }
  var values = yIdx.length ? rows.map(function(row){ return (row || [])[yIdx[0]]; }).filter(function(v){ return v !== undefined && looksNumeric(v); }).map(Number) : [];
  return { seriesCount: yIdx.length, labels: rows.length, values: values };
}
function chartDataSplitFields(v) {
  return String(v || "").split(",").map(function(x){ return x.trim(); }).filter(Boolean);
}
function chartDataProfile(text, format, xField, yField) {
  var t = String(text || "");
  if (!t.trim()) return { seriesCount: 0, labels: 0, values: [] };
  var f = format;
  if (f === "auto" || f === "__expr__") f = chartDataDetectFormat(t);
  if (f === "json") {
    try {
      var data = JSON.parse(t);
      if (Array.isArray(data)) {
        if (data.every(function(v){ return typeof v === "number" || (typeof v === "string" && looksNumeric(v)); })) {
          return { seriesCount: 1, labels: data.length, values: data.map(Number) };
        }
        if (Array.isArray(data[0])) {
          var rows = data, columns = null;
          if (data.length > 1 && data[0].every(function(v){ return !looksNumeric(v); })) { columns = data[0]; rows = data.slice(1); }
          return chartDataColumnProfile(rows, columns, xField, yField);
        }
        var keys = [];
        data.forEach(function(o){ if (o && typeof o === "object" && !Array.isArray(o)) Object.keys(o).forEach(function(k){ if (keys.indexOf(k) < 0) keys.push(k); }); });
        var xKey = xField || "";
        if (!xKey) {
          for (var ki = 0; ki < keys.length; ki++) {
            var nonNum = data.some(function(o){ return o && o[keys[ki]] !== undefined && o[keys[ki]] !== "" && !looksNumeric(o[keys[ki]]); });
            if (nonNum) { xKey = keys[ki]; break; }
          }
        }
        var yKeys = [];
        var specs = chartDataSplitFields(yField);
        if (specs.length) specs.forEach(function(name){ if (keys.indexOf(name) >= 0 && name !== xKey) yKeys.push(name); });
        if (!yKeys.length) keys.forEach(function(k){ if (k !== xKey && data.some(function(o){ return o && looksNumeric(o[k]); })) yKeys.push(k); });
        var vals = yKeys.length ? data.map(function(o){ return o ? o[yKeys[0]] : null; }).filter(function(v){ return v !== null && v !== undefined && v !== "" && looksNumeric(v); }).map(Number) : [];
        return { seriesCount: yKeys.length, labels: data.length, values: vals };
      }
      if (data && typeof data === "object") {
        if (Array.isArray(data.values)) return { seriesCount: 1, labels: data.values.length, values: data.values.filter(function(v){ return looksNumeric(v); }).map(Number) };
        if (Array.isArray(data.rows)) return chartDataColumnProfile(data.rows, Array.isArray(data.columns) ? data.columns : null, xField, yField);
        if (data.values && typeof data.values === "object") {
          var vv = Object.keys(data.values).map(function(k){ return data.values[k]; }).filter(function(v){ return looksNumeric(v); }).map(Number);
          return { seriesCount: 1, labels: vv.length, values: vv };
        }
        var mv = Object.keys(data).map(function(k){ return data[k]; }).filter(function(v){ return looksNumeric(v); }).map(Number);
        return { seriesCount: 1, labels: mv.length, values: mv };
      }
    } catch (e) {}
    return { seriesCount: 0, labels: 0, values: [] };
  }
  if (f === "xml") {
    try {
      var doc = new DOMParser().parseFromString(t, "application/xml");
      if (doc.getElementsByTagName("parsererror").length) return { seriesCount: 0, labels: 0, values: [] };
      var root = doc.documentElement, nodes = [];
      for (var xi = 0; xi < root.childNodes.length; xi++) if (root.childNodes[xi].nodeType === 1) nodes.push(root.childNodes[xi]);
      if (!nodes.length) nodes = [root];
      var names = [], rows2 = [];
      nodes.forEach(function(el){
        var vals = {}, j;
        for (j = 0; j < el.attributes.length; j++) {
          var at = el.attributes[j], ak = at.name;
          if (["label","x","name","key","id","class","style"].indexOf(ak) >= 0) continue;
          if (looksNumeric(at.value)) vals[ak === "y" ? "Value" : ak] = at.value;
        }
        for (j = 0; j < el.childNodes.length; j++) {
          var ch = el.childNodes[j];
          if (ch.nodeType !== 1) continue;
          var tag = (ch.localName || ch.nodeName || "").toLowerCase();
          if (["label","x","name"].indexOf(tag) >= 0) continue;
          if (looksNumeric(ch.textContent)) vals[tag === "y" ? "Value" : tag] = ch.textContent;
        }
        if (!Object.keys(vals).length && looksNumeric(el.textContent)) vals.Value = el.textContent;
        if (!Object.keys(vals).length) return;
        Object.keys(vals).forEach(function(k){ if (names.indexOf(k) < 0) names.push(k); });
        rows2.push(vals);
      });
      var wantedXml = chartDataSplitFields(yField);
      if (wantedXml.length) {
        var filteredXml = names.filter(function(name){ return wantedXml.indexOf(name) >= 0; });
        if (filteredXml.length) names = filteredXml;
      }
      var vals2 = names.length ? rows2.map(function(r){ return r[names[0]]; }).filter(function(v){ return v !== undefined && looksNumeric(v); }).map(Number) : [];
      return { seriesCount: names.length, labels: rows2.length, values: vals2 };
    } catch (e) { return { seriesCount: 0, labels: 0, values: [] }; }
  }
  var rows3 = chartDataCsvRows(t);
  if (!rows3.length) return { seriesCount: 0, labels: 0, values: [] };
  var header = rows3[0], dataRows = rows3.slice(1), hasHeader = false;
  for (var h = 0; h < header.length && !hasHeader; h++) {
    if (looksNumeric(header[h])) continue;
    for (var hr = 1; hr < rows3.length && hr < 30; hr++) {
      if (dataRows[hr - 1] && looksNumeric(dataRows[hr - 1][h])) { hasHeader = true; break; }
    }
  }
  return chartDataColumnProfile(hasHeader ? dataRows : rows3, hasHeader ? header : null, xField, yField);
}
function chartDataGanttField(obj, names) {
  if (!obj || typeof obj !== "object") return undefined;
  var keys = Object.keys(obj);
  for (var n = 0; n < names.length; n++) for (var k = 0; k < keys.length; k++) {
    if (keys[k].toLowerCase() === names[n].toLowerCase()) return obj[keys[k]];
  }
  return undefined;
}
function chartDataGanttIndex(head, names, fallback) {
  for (var i = 0; i < head.length; i++) for (var n = 0; n < names.length; n++) {
    if (String(head[i]).trim().toLowerCase() === names[n].toLowerCase()) return i;
  }
  return fallback;
}
function chartDataGanttProfile(text, format) {
  var t = String(text || "");
  if (!t.trim()) return { seriesCount: 0, labels: 0, values: [] };
  var f = format;
  if (f === "auto" || f === "__expr__") f = chartDataDetectFormat(t);
  var rows = 0;
  try {
    if (f === "json") {
      var data = JSON.parse(t);
      var arr = Array.isArray(data) ? data : (data && Array.isArray(data.tasks) ? data.tasks : []);
      if (arr.length && Array.isArray(arr[0])) {
        var h = arr[0];
        var si = chartDataGanttIndex(h, ["start", "from"], 1), ei = chartDataGanttIndex(h, ["end", "to"], 2);
        for (var i = 1; i < arr.length; i++) if (arr[i] && String(arr[i][si] || "").trim() !== "" && String(arr[i][ei] || "").trim() !== "") rows++;
      } else {
        for (var j = 0; j < arr.length; j++) {
          if (!arr[j] || typeof arr[j] !== "object") continue;
          var st = chartDataGanttField(arr[j], ["start", "from", "begin"]);
          var en = chartDataGanttField(arr[j], ["end", "to", "finish"]);
          var du = chartDataGanttField(arr[j], ["duration"]);
          if (st !== undefined && (en !== undefined || du !== undefined)) rows++;
        }
      }
    } else if (f === "csv") {
      var cr = chartDataCsvRows(t);
      if (cr.length) {
        var head = cr[0], body = cr.slice(1);
        var has = head.some(function(c){ var s=String(c).trim().toLowerCase(); return s==="start"||s==="end"||s==="task"||s==="label"; });
        if (!has) { head = ["Task","Start","End"]; body = cr; }
        var xs = chartDataGanttIndex(head, ["start", "from"], 1), xe = chartDataGanttIndex(head, ["end", "to"], 2);
        for (var r = 0; r < body.length; r++) if (body[r] && String(body[r][xs] || "").trim() !== "" && String(body[r][xe] || "").trim() !== "") rows++;
      }
    } else if (f === "xml") {
      var doc = new DOMParser().parseFromString(t, "application/xml");
      if (!doc.getElementsByTagName("parsererror").length) {
        var root = doc.documentElement, nodes = [];
        for (var n = 0; n < root.childNodes.length; n++) if (root.childNodes[n].nodeType === 1) nodes.push(root.childNodes[n]);
        if (!nodes.length) nodes = [root];
        nodes.forEach(function(el){
          var st = "", en = "";
          ["start","from"].forEach(function(k){ if (!st && el.getAttribute(k)) st = el.getAttribute(k); });
          ["end","to"].forEach(function(k){ if (!en && el.getAttribute(k)) en = el.getAttribute(k); });
          for (var c = 0; c < el.childNodes.length; c++) {
            var ch = el.childNodes[c]; if (ch.nodeType !== 1) continue;
            var tag = (ch.localName || ch.nodeName || "").toLowerCase();
            if ((tag === "start" || tag === "from") && !st) st = ch.textContent;
            if ((tag === "end" || tag === "to") && !en) en = ch.textContent;
          }
          if (String(st).trim() !== "" && String(en).trim() !== "") rows++;
        });
      }
    }
  } catch (e) {}
  return { seriesCount: rows ? 2 : 0, labels: rows, values: [] };
}
function chartDataStateProfile(text, format) {
  var t = String(text || "");
  var out = { seriesCount: 0, labels: 0, values: [], stateCount: 0, transitionCount: 0 };
  if (!t.trim()) return out;
  var f = format;
  if (f === "auto" || f === "__expr__") f = chartDataDetectFormat(t);
  try {
    if (f === "json") {
      var d = JSON.parse(t);
      var hasId = function(x) { return x && typeof x === "object" && (x.id !== undefined || x.key !== undefined); };
      var hasEdge = function(x) { return x && typeof x === "object" && (x.from !== undefined || x.source !== undefined); };
      var states = [];
      var transitions = [];
      if (Array.isArray(d)) {
        if (d.some(hasEdge)) transitions = d;
        else if (d.some(hasId)) states = d;
      } else if (d && typeof d === "object") {
        var sr = Array.isArray(d.states) ? d.states : (Array.isArray(d.nodes) ? d.nodes : []);
        states = sr.filter(hasId);
        var tr = Array.isArray(d.transitions) ? d.transitions : (Array.isArray(d.edges) ? d.edges : []);
        transitions = tr.filter(hasEdge);
      }
      out.stateCount = states.length;
      out.transitionCount = transitions.length;
      out.labels = Math.max(states.length, transitions.length);
    } else if (f === "xml") {
      var doc = new DOMParser().parseFromString(t, "application/xml");
      if (!doc.getElementsByTagName("parsererror").length) {
        out.stateCount = doc.getElementsByTagName("state").length + doc.getElementsByTagName("node").length;
        out.transitionCount = doc.getElementsByTagName("transition").length + doc.getElementsByTagName("edge").length;
        out.labels = Math.max(out.stateCount, out.transitionCount);
      }
    } else {
      var rows = chartDataCsvRows(t);
      var body = rows.length > 1 ? rows.slice(1) : rows;
      var head = rows.length ? rows[0] : [];
      var lower = head.map(function(c){ return String(c).trim().toLowerCase(); });
      var hasEdges = lower.some(function(c){ return ["from","source","fromstate","from_state","to","target","tostate","to_state"].indexOf(c) >= 0; });
      var hasStates = lower.some(function(c){ return ["id","key","state","type","kind","label","name"].indexOf(c) >= 0; });
      var hasNumericValue = lower.some(function(c){ return ["value","values","amount","total","count","y"].indexOf(c) >= 0; });
      if (hasEdges || (hasStates && !hasNumericValue)) {
        out.stateCount = hasEdges ? 0 : body.length;
        out.transitionCount = hasEdges ? body.length : 0;
        out.labels = body.length;
      }
    }
  } catch (e) {}
  return out;
}
function chartDataCompatibility(profile, type) {
  if (!profile || type === "__expr__") return { ok: true, level: "warn", msg: "Chart type is an expression; compatibility is checked at runtime." };
  if (type === "state" || type === "statemachine") {
    if (profile.stateCount || profile.transitionCount) return { ok: true, level: "ok", msg: "Compatible with block diagram (" + profile.stateCount + " states, " + profile.transitionCount + " transitions)." };
    return { ok: false, level: "err", msg: "Block diagram data needs at least one state or transition." };
  }
  if (!profile.labels) return { ok: false, level: "err", msg: "No " + chartTypeLabel(type) + " data rows were found in the current dataset." };
  if (!profile.seriesCount) return { ok: false, level: "err", msg: "No numeric series found. Add a numeric column or check the x/y fields." };
  var multi = chartTypeIsMulti(type);
  if (multi && profile.seriesCount < 2) {
    return { ok: false, level: "err", msg: chartTypeLabel(type) + " needs at least two numeric series. Add a second numeric column or choose a single-series type." };
  }
  if (type === "heatmap" && profile.seriesCount !== 2) {
    return { ok: false, level: "err", msg: "Heatmap needs exactly two numeric series: y coordinate and value." };
  }
  if (type === "candlestick" && profile.seriesCount !== 4) {
    return { ok: false, level: "err", msg: "Candlestick needs exactly four numeric series: open, high, low, close." };
  }
  if (type === "boxplot" && profile.seriesCount !== 5) {
    return { ok: false, level: "err", msg: "Box plot needs exactly five series: min, q1, median, q3, max." };
  }
  if (type === "bullet" && (profile.seriesCount < 2 || profile.seriesCount > 3)) {
    return { ok: false, level: "err", msg: "Bullet needs two or three series: value, target, optional max." };
  }
  if ((type === "dumbbell" || type === "gantt") && profile.seriesCount !== 2) {
    return { ok: false, level: "err", msg: chartTypeLabel(type) + " needs exactly two numeric series." };
  }
  if (type === "radar" && profile.labels < 3) {
    return { ok: false, level: "err", msg: "Radar needs at least three categories." };
  }
  if (["heatmap","candlestick","boxplot","bullet","dumbbell","gantt"].indexOf(type) < 0 && !chartTypeSupportsMulti(type) && profile.seriesCount > 1) {
    return { ok: false, level: "err", msg: "This data has " + profile.seriesCount + " numeric series. Choose a multi-series chart or set y to one field." };
  }
  if ((type === "pie" || type === "donut" || type === "funnel" || type === "treemap") && !profile.values.some(function(v){ return v > 0; })) {
    return { ok: false, level: "err", msg: chartTypeLabel(type) + " needs at least one positive value." };
  }
  if (type === "histogram" && !profile.values.length) {
    return { ok: false, level: "err", msg: "Histogram needs numeric values." };
  }
  if (type === "waterfall" && profile.values.length && !profile.values.some(function(v){ return v < 0; })) {
    return { ok: true, level: "warn", msg: "Waterfall has no negative deltas; it will only show increases." };
  }
  return { ok: true, level: "ok", msg: "Compatible with " + chartTypeLabel(type) + " (" + profile.seriesCount + " numeric series)." };
}
// Pick the parser that matches the target geometry. State/graph data is a
// shapes/links document, not a numeric series, and Gantt has its own date
// profile, so a plain chart profile would report a false mismatch.
function chartDataProfileForType(text, format, type, xField, yField) {
  if (type === "state" || type === "statemachine") return chartDataStateProfile(text, format);
  if (type === "gantt") return chartDataGanttProfile(text, format);
  return chartDataProfile(text, format, xField, yField);
}
function chartTypeCompatibilityForData(text, format, type, xField, yField) {
  var profile = chartDataProfileForType(text, format, type, xField, yField);
  return { profile: profile, compat: chartDataCompatibility(profile, type) };
}
// Legacy chartType values stay selectable only on documents that already use
// them, so the normal chart type menu no longer offers "State machine".
function chartTypeMenuOptions(n) {
  var out = CHART_TYPE_OPTIONS.slice();
  var current = n && n.chartType ? String(n.chartType) : "";
  if (current === "state" || current === "statemachine") {
    out.push([current, "Block diagram (legacy chart)"]);
  }
  return out;
}
// Decide what a chart type change implies for the current dataset. Numeric
// types keep compatible data; an incompatible or empty dataset gets the
// sample that matches the target geometry.
function chartTypeChangePlan(n, newType) {
  var sourceFormat = n.chartFormat || "auto";
  var sampleFormat = chartSampleFormatFor(sourceFormat);
  var sample = chartSampleDataset(newType, sampleFormat);
  var appliedFormat = sampleFormat === "json" ? "auto" : sampleFormat;
  if (n.chartDataExpr) return { needsReplace: false, fillSample: false, sample: "", format: "", reason: "" };
  var source = String(n.chartData || "");
  if (!source.trim()) return { needsReplace: false, fillSample: true, sample: sample, format: appliedFormat, reason: "" };
  var check = chartTypeCompatibilityForData(
    source,
    sourceFormat === "__expr__" ? "auto" : sourceFormat,
    newType,
    n.chartX,
    n.chartY
  );
  if (check.compat.ok) return { needsReplace: false, fillSample: false, sample: "", format: "", reason: "" };
  return { needsReplace: true, fillSample: false, sample: sample, format: appliedFormat, reason: check.compat.msg };
}
function chartApplyTypeChange(n, newType, plan) {
  snap("chart type");
  chartClearParam(n, "chartType");
  n.chartType = newType;
  if (plan && (plan.needsReplace || plan.fillSample)) {
    n.chartData = plan.sample;
    n.chartDataExpr = false;
    chartClearParam(n, "chartFormat");
    n.chartFormat = plan.format || "auto";
  }
  setOut();
  autoPreview();
  renderInspector();
}
function chartTypeWarningText(newType, plan) {
  var label = chartTypeLabel(newType);
  var why = plan && plan.reason ? " " + plan.reason : "";
  return "The current data is not compatible with " + label + "." + why +
    " Changing the chart type will ignore the existing data and replace it with a sample dataset suited to " + label + ".";
}
function chartDataLines(text) {
  var t = String(text || "");
  if (!t) return 0;
  return t.split(/\r?\n/).filter(function(l){ return l.trim() !== ""; }).length;
}
function chartDataButtonText(n) {
  if (n.chartDataExpr) return "Edit data... (" + (n.chartData || "expression") + ")";
  var fmt = n.chartFormat || "auto";
  var lines = chartDataLines(n.chartData);
  return "Edit data... (" + String(fmt).toUpperCase() + " · " + lines + " lines)";
}
function chartDataDetectFormat(text) {
  var t = String(text || "").replace(/^\s+/, "");
  if (!t) return "csv";
  var c = t.charAt(0);
  if (c === "{" || c === "[") return "json";
  if (c === "<") return "xml";
  return "csv";
}
function chartDataDetectSep(line) {
  var cands = [",", ";", "\t", "|"], best = ",", bestN = -1;
  for (var i = 0; i < cands.length; i++) {
    var n = 0, q = false;
    for (var j = 0; j < line.length; j++) {
      var ch = line.charAt(j);
      if (ch === '"') q = !q;
      else if (!q && ch === cands[i]) n++;
    }
    if (n > bestN) { bestN = n; best = cands[i]; }
  }
  return best;
}
function chartDataJsonRows(data) {
  if (Array.isArray(data)) return data.length;
  if (data && typeof data === "object") {
    if (Array.isArray(data.values)) return data.values.length;
    if (Array.isArray(data.labels)) return data.labels.length;
    if (Array.isArray(data.rows)) return data.rows.length;
    return Object.keys(data).length;
  }
  return 0;
}
function chartDataValidateText(text, fmt) {
  var t = String(text || "");
  if (!t.trim()) {
    return { ok: true, level: "warn", format: fmt === "auto" || fmt === "__expr__" ? "auto" : fmt,
      rows: 0, chars: 0, msg: "Empty data: the chart will show 'No data'." };
  }
  var f = fmt;
  if (f === "auto" || f === "__expr__") f = chartDataDetectFormat(t);
  if (f === "json") {
    try {
      var data = JSON.parse(t);
      var rows = chartDataJsonRows(data);
      if (!rows) return { ok: false, level: "err", format: "json", chars: t.length,
        msg: "JSON parsed, but no data rows were found." };
      return { ok: true, level: "ok", format: "json", rows: rows, chars: t.length,
        msg: "Valid JSON · " + rows + " rows." };
    } catch (e) {
      return { ok: false, level: "err", format: "json", chars: t.length,
        msg: "Invalid JSON: " + (e && e.message ? e.message : String(e)) };
    }
  }
  if (f === "xml") {
    if (/<!DOCTYPE|<!ENTITY/i.test(t)) {
      return { ok: false, level: "err", format: "xml", chars: t.length, msg: "DTD/entity declarations are not allowed." };
    }
    if (!window.DOMParser) return { ok: false, level: "err", format: "xml", chars: t.length, msg: "XML parser is not available." };
    var doc = new DOMParser().parseFromString(t, "application/xml");
    if (doc.getElementsByTagName("parsererror").length) {
      return { ok: false, level: "err", format: "xml", chars: t.length, msg: "Invalid XML." };
    }
    var root = doc.documentElement;
    var xrows = 0;
    if (root) {
      for (var i = 0; i < root.childNodes.length; i++) if (root.childNodes[i].nodeType === 1) xrows++;
      if (!xrows) xrows = 1;
    }
    return { ok: true, level: "ok", format: "xml", rows: xrows, chars: t.length,
      msg: "Valid XML · " + xrows + " row elements." };
  }
  var lines = t.split(/\r?\n/).filter(function(l){ return l.trim() !== ""; });
  if (!lines.length) return { ok: true, level: "warn", format: "csv", rows: 0, chars: t.length, msg: "Empty CSV." };
  var sep = chartDataDetectSep(lines[0]);
  var cols = lines[0].split(sep).length;
  if (cols < 2) {
    return { ok: false, level: "err", format: "csv", chars: t.length,
      msg: "CSV needs at least two columns (label and value)." };
  }
  var rows = Math.max(0, lines.length - 1);
  return { ok: true, level: "ok", format: "csv", rows: rows, chars: t.length,
    msg: "Valid CSV · " + rows + " data rows, " + cols + " columns." };
}
function chartDataEls() {
  return {
    layer: document.getElementById("chart-data-layer"),
    panel: document.getElementById("cd-panel"),
    head: document.getElementById("cd-head"),
    subtitle: document.getElementById("cd-subtitle"),
    literalBtn: document.getElementById("cd-mode-literal"),
    exprBtn: document.getElementById("cd-mode-expr"),
    literalWrap: document.getElementById("cd-literal-wrap"),
    exprWrap: document.getElementById("cd-expr-wrap"),
    text: document.getElementById("cd-text"),
    expr: document.getElementById("cd-expr"),
    format: document.getElementById("cd-format"),
    type: document.getElementById("cd-type"),
    stats: document.getElementById("cd-stats"),
    status: document.getElementById("cd-status"),
    apply: document.getElementById("cd-apply"),
    pretty: document.getElementById("cd-pretty"),
    file: document.getElementById("cd-fileinput")
  };
}
function chartDataSetMode(mode, focus) {
  chartDataEditorMode = mode === "expr" ? "expr" : "literal";
  var e = chartDataEls();
  if (!e.layer) return;
  e.literalBtn.classList.toggle("on", chartDataEditorMode === "literal");
  e.exprBtn.classList.toggle("on", chartDataEditorMode === "expr");
  e.literalWrap.hidden = chartDataEditorMode !== "literal";
  e.exprWrap.hidden = chartDataEditorMode !== "expr";
  chartDataUpdateStatus();
  if (focus) {
    if (chartDataEditorMode === "expr") e.expr.focus();
    else e.text.focus();
  }
}
function chartDataUpdateStatus() {
  var e = chartDataEls();
  if (!e.status || !chartDataDraft) return null;
  var r, profile = null, compat = null;
  if (chartDataEditorMode === "expr") {
    var expr = String(chartDataDraft.expression || "").trim();
    if (!expr) {
      r = { ok: false, level: "err", msg: "Enter a state or expression." };
    } else {
      r = { ok: true, level: "ok", format: "expression", rows: 0, chars: expr.length,
        msg: "Expression mode: the app evaluates this source and redraws the chart." };
    }
    e.stats.textContent = "Expression";
  } else {
    var fmt = chartDataDraft.format || "auto";
    r = chartDataValidateText(chartDataDraft.literal, fmt);
    var lines = chartDataLines(chartDataDraft.literal);
    var chars = String(chartDataDraft.literal || "").length;
    if (r.ok && String(chartDataDraft.literal || "").trim() !== "") {
      profile = chartDataDraft.type === "gantt"
        ? chartDataGanttProfile(chartDataDraft.literal, fmt === "__expr__" ? "auto" : fmt)
        : (chartDataDraft.type === "state" || chartDataDraft.type === "statemachine"
          ? chartDataStateProfile(chartDataDraft.literal, fmt === "__expr__" ? "auto" : fmt)
          : chartDataProfile(chartDataDraft.literal, fmt === "__expr__" ? "auto" : fmt, chartDataDraft.xField, chartDataDraft.yField));
      compat = chartDataCompatibility(profile, chartDataDraft.type);
    }
    e.stats.textContent = chartTypeLabel(chartDataDraft.type) + " · " +
      String(fmt === "__expr__" ? "format expression" : fmt).toUpperCase() + " · " + lines + " lines" +
      (profile ? " · " + profile.seriesCount + " series" : "");
  }
  if (r.ok && compat) {
    if (!compat.ok) {
      r = { ok: false, level: "err", format: r.format, rows: r.rows, chars: r.chars,
        msg: r.msg + " " + compat.msg };
    } else if (compat.level === "warn") {
      r = { ok: r.ok, level: "warn", format: r.format, rows: r.rows, chars: r.chars,
        msg: r.msg + " " + compat.msg };
    } else {
      r = { ok: r.ok, level: "ok", format: r.format, rows: r.rows, chars: r.chars,
        msg: r.msg + " " + compat.msg };
    }
  }
  e.status.className = "cd-status" + (r.level === "ok" ? " ok" : r.level === "err" ? " err" : r.level === "warn" ? " warn" : "");
  e.status.textContent = r.msg || "";
  if (e.apply) e.apply.disabled = !r.ok;
  if (e.pretty) e.pretty.disabled = !(r.ok && r.format === "json");
  return r;
}
function chartDataSaveGeom() {
  var p = document.getElementById("cd-panel");
  if (!p) return;
  chartDataPanelGeom = { left: p.offsetLeft, top: p.offsetTop, width: p.offsetWidth, height: p.offsetHeight };
  try { localStorage.setItem("sui.chartDataPanelGeom", JSON.stringify(chartDataPanelGeom)); } catch (e) {}
}
function chartDataPlacePanel() {
  var p = document.getElementById("cd-panel");
  if (!p) return;
  if (!chartDataPanelGeom) {
    try { chartDataPanelGeom = JSON.parse(localStorage.getItem("sui.chartDataPanelGeom") || "null"); } catch (e) { chartDataPanelGeom = null; }
  }
  var vw = Math.max(420, window.innerWidth || document.documentElement.clientWidth || 1024);
  var vh = Math.max(320, window.innerHeight || document.documentElement.clientHeight || 768);
  var w = chartDataPanelGeom && chartDataPanelGeom.width ? chartDataPanelGeom.width : Math.min(820, vw - 70);
  var h = chartDataPanelGeom && chartDataPanelGeom.height ? chartDataPanelGeom.height : Math.min(560, vh - 70);
  w = Math.max(380, Math.min(w, vw - 24));
  h = Math.max(300, Math.min(h, vh - 24));
  var left = chartDataPanelGeom && isFinite(chartDataPanelGeom.left) ? chartDataPanelGeom.left : Math.max(12, (vw - w) / 2);
  var top = chartDataPanelGeom && isFinite(chartDataPanelGeom.top) ? chartDataPanelGeom.top : Math.max(12, (vh - h) / 2);
  left = Math.max(0, Math.min(left, vw - w - 8));
  top = Math.max(0, Math.min(top, vh - h - 8));
  p.style.width = w + "px";
  p.style.height = h + "px";
  p.style.left = left + "px";
  p.style.top = top + "px";
}
function chartDataFillFormat(n) {
  var sel = document.getElementById("cd-format");
  if (!sel) return;
  sel.innerHTML = "";
  [["auto", "Auto detect"], ["json", "JSON"], ["xml", "XML"], ["csv", "CSV"]].forEach(function(v){
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1]; sel.appendChild(o);
  });
  var fmtExpr = chartParamExpr(n, "chartFormat");
  if (fmtExpr) {
    var o = document.createElement("option");
    o.value = "__expr__";
    o.textContent = "Expression: " + fmtExpr;
    sel.appendChild(o);
    sel.value = "__expr__";
  } else {
    sel.value = n.chartFormat || "auto";
  }
}
function chartDataFillType(n) {
  var sel = document.getElementById("cd-type");
  if (!sel) return;
  sel.innerHTML = "";
  if (isGraphKind(n)) {
    var graphOpt = document.createElement("option");
    graphOpt.value = "state";
    graphOpt.textContent = "Graph — block diagram";
    sel.appendChild(graphOpt);
    sel.value = "state";
    sel.disabled = true;
    return;
  }
  sel.disabled = false;
  chartTypeMenuOptions(n).forEach(function(v){
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1]; sel.appendChild(o);
  });
  var typeExpr = chartParamExpr(n, "chartType");
  if (typeExpr) {
    var o = document.createElement("option");
    o.value = "__expr__";
    o.textContent = "Expression: " + typeExpr;
    sel.appendChild(o);
    sel.value = "__expr__";
  } else {
    sel.value = n.chartType || "bar";
  }
}
function chartDataSetLiteral(text) {
  chartDataDraft.literal = String(text || "");
  document.getElementById("cd-text").value = chartDataDraft.literal;
}
function openChartDataEditor(n) {
  if (!isChartKind(n)) return;
  chartDataEditorNode = n;
  var fmtExpr = chartParamExpr(n, "chartFormat");
  chartDataDraft = {
    literal: n.chartDataExpr ? "" : (n.chartData || ""),
    expression: n.chartDataExpr ? (n.chartData || "") : "",
    format: fmtExpr ? "__expr__" : (n.chartFormat || "auto"),
    type: isGraphKind(n) ? "state" : (chartParamExpr(n, "chartType") ? "__expr__" : (n.chartType || "bar")),
    xField: n.chartX || "",
    yField: n.chartY || ""
  };
  var e = chartDataEls();
  e.subtitle.textContent = (n.text || "Chart") + (n.name ? "  (#" + n.name + ")" : "");
  document.getElementById("cd-text").value = chartDataDraft.literal;
  document.getElementById("cd-expr").value = chartDataDraft.expression;
  chartDataFillFormat(n);
  chartDataFillType(n);
  chartDataPlacePanel();
  e.layer.classList.add("on");
  e.layer.setAttribute("aria-hidden", "false");
  chartDataSetMode(n.chartDataExpr ? "expr" : "literal", false);
  chartDataUpdateStatus();
  setTimeout(function(){ if (chartDataEditorMode === "expr") e.expr.focus(); else e.text.focus(); }, 0);
}
function closeChartDataEditor() {
  var e = chartDataEls();
  if (e.layer && e.layer.classList.contains("on")) chartDataSaveGeom();
  if (e.layer) { e.layer.classList.remove("on"); e.layer.setAttribute("aria-hidden", "true"); }
  chartDataEditorNode = null;
  chartDataDraft = null;
}
function chartDataInferFormat(name) {
  var m = String(name || "").toLowerCase().match(/\.([a-z0-9]+)$/);
  var ext = m ? m[1] : "";
  if (ext === "json") return "json";
  if (ext === "xml") return "xml";
  if (ext === "csv" || ext === "tsv") return "csv";
  return "auto";
}
function chartDataApply() {
  var r = chartDataUpdateStatus();
  if (!r || !r.ok || !chartDataEditorNode) return;
  var n = chartDataEditorNode;
  snap("chart data");
  if (chartDataEditorMode === "expr") {
    n.chartData = String(chartDataDraft.expression || "").trim();
    n.chartDataExpr = true;
  } else {
    n.chartData = String(chartDataDraft.literal || "");
    n.chartDataExpr = false;
  }
  if (chartDataDraft.format !== "__expr__") {
    chartClearParam(n, "chartFormat");
    n.chartFormat = chartDataDraft.format || "auto";
  }
  if (chartDataDraft.type !== "__expr__") {
    chartClearParam(n, "chartType");
    n.chartType = chartDataDraft.type || "bar";
  }
  setOut();
  autoPreview();
  renderInspector();
  closeChartDataEditor();
  setStatus("chart data updated");
}
function chartDataInsertText(text, insert) {
  var start = text.selectionStart, end = text.selectionEnd;
  text.value = text.value.slice(0, start) + insert + text.value.slice(end);
  text.selectionStart = text.selectionEnd = start + insert.length;
  chartDataDraft.literal = text.value;
  chartDataUpdateStatus();
}
function wireChartDataEditor() {
  var e = chartDataEls();
  if (!e.layer || e.layer.__wired) return;
  e.layer.__wired = true;
  makeDraggable(e.panel, e.head);
  makeResizable(e.panel, document.getElementById("cd-resize"), 380, 300);
  document.getElementById("cd-close").onclick = closeChartDataEditor;
  document.getElementById("cd-cancel").onclick = closeChartDataEditor;
  e.apply.onclick = chartDataApply;
  e.literalBtn.onclick = function(){ chartDataSetMode("literal", true); };
  e.exprBtn.onclick = function(){ chartDataSetMode("expr", true); };
  e.text.oninput = function(){ chartDataDraft.literal = e.text.value; chartDataUpdateStatus(); };
  e.expr.oninput = function(){ chartDataDraft.expression = e.expr.value; chartDataUpdateStatus(); };
  e.format.onchange = function(){ chartDataDraft.format = e.format.value; chartDataUpdateStatus(); };
  e.type.onchange = function(){
    var newType = e.type.value;
    if (newType === chartDataDraft.type) return;
    if (newType === "__expr__") { chartDataDraft.type = newType; chartDataUpdateStatus(); return; }
    var probe = {
      chartData: chartDataDraft.literal,
      chartDataExpr: chartDataEditorMode === "expr",
      chartFormat: chartDataDraft.format,
      chartX: chartDataDraft.xField,
      chartY: chartDataDraft.yField
    };
    var plan = chartTypeChangePlan(probe, newType);
    if (plan.needsReplace) {
      var previous = chartDataDraft.type;
      e.type.value = previous;
      confirmBox("Replace chart data?", chartTypeWarningText(newType, plan), function(){
        chartDataDraft.type = newType;
        e.type.value = newType;
        chartDataSetLiteral(plan.sample);
        chartDataSetMode("literal", false);
        if (plan.format) { e.format.value = plan.format; chartDataDraft.format = plan.format; }
        chartDataUpdateStatus();
      });
      return;
    }
    chartDataDraft.type = newType;
    if (plan.fillSample) {
      chartDataSetLiteral(plan.sample);
      chartDataSetMode("literal", false);
      if (plan.format) { e.format.value = plan.format; chartDataDraft.format = plan.format; }
    }
    chartDataUpdateStatus();
  };
  e.text.addEventListener("keydown", function(ev){
    if (ev.key === "Tab") { ev.preventDefault(); chartDataInsertText(e.text, "  "); }
  });
  document.getElementById("cd-sample").onclick = function(){
    var fmt = e.format.value;
    var sampleFormat = chartSampleFormatFor(fmt === "__expr__" ? "auto" : fmt);
    chartDataSetLiteral(chartSampleDataset(chartDataDraft.type, sampleFormat));
    chartDataSetMode("literal", false);
    chartDataUpdateStatus();
  };
  e.pretty.onclick = function(){
    try {
      var pretty = JSON.stringify(JSON.parse(chartDataDraft.literal), null, 2);
      chartDataSetLiteral(pretty);
      chartDataUpdateStatus();
      setStatus("JSON formatted");
    } catch (err) {
      chartDataUpdateStatus();
    }
  };
  document.getElementById("cd-load").onclick = function(){ e.file.value = ""; e.file.click(); };
  e.file.onchange = function(){
    var f = e.file.files && e.file.files[0];
    if (!f) return;
    var reader = new FileReader();
    reader.onload = function(){
      chartDataSetLiteral(String(reader.result || ""));
      var fmt = chartDataInferFormat(f.name);
      if (fmt !== "auto") {
        var has = false;
        for (var i = 0; i < e.format.options.length; i++) if (e.format.options[i].value === fmt) has = true;
        if (has) { e.format.value = fmt; chartDataDraft.format = fmt; }
      }
      chartDataSetMode("literal", false);
      chartDataUpdateStatus();
      setStatus("loaded " + f.name);
    };
    reader.onerror = function(){ e.status.className = "cd-status err"; e.status.textContent = "Cannot read the selected file."; };
    reader.readAsText(f);
  };
  e.layer.addEventListener("mousedown", function(ev){ if (ev.target === e.layer) closeChartDataEditor(); });
  document.addEventListener("keydown", function(ev){
    if (!e.layer.classList.contains("on")) return;
    if (ev.key === "Escape") { ev.preventDefault(); closeChartDataEditor(); }
    else if (ev.key === "Enter" && (ev.ctrlKey || ev.metaKey)) { ev.preventDefault(); chartDataApply(); }
  });
}
function chartFieldRow(labelText, control, hintText) {
  var row = document.createElement("div");
  row.className = "rowline";
  var label = document.createElement("label");
  label.textContent = labelText;
  row.appendChild(label);
  row.appendChild(control);
  if (hintText) label.title = hintText;
  return row;
}
function chartControls(n) {
  var a = document.createElement("div");
  a.className = "acc open";
  var h = document.createElement("div");
  h.className = "acchead";
  var graph = isGraphKind(n);
  h.innerHTML = '<span>' + (graph ? "Graph" : "Chart") + '</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div");
  b.className = "accbody";

  var hint = document.createElement("div");
  hint.className = "hint";
  hint.textContent = graph
    ? "Block diagram data: states as boxes and transitions as links. JSON, XML or CSV; use Element hints for shapes and links."
    : "Data can be JSON, XML or CSV. Width/height define the frame; scroll shows bars, zoom adds +/- controls when the chart is larger than its frame.";
  b.appendChild(hint);

  var typeSel = document.createElement("select");
  chartTypeMenuOptions(n).forEach(function(v){
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1]; typeSel.appendChild(o);
  });
  typeSel.value = n.chartType || "bar";
  if (!typeSel.value) typeSel.value = "bar";
  var typeRow = chartFieldRow("type", typeSel, "Chart geometry. Fields below follow the selected type.");
  if (!graph) b.appendChild(typeRow);
  var typeHint = document.createElement("div");
  typeHint.className = "hint";
  if (!graph) b.appendChild(typeHint);

  var dataBtn = document.createElement("button");
  dataBtn.type = "button";
  dataBtn.className = "btn chart-data-btn";
  dataBtn.style.cssText = "flex:1;text-align:left;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap";
  dataBtn.textContent = chartDataButtonText(n);
  dataBtn.title = "Open the resizable chart data editor";
  dataBtn.onclick = function(){ openChartDataEditor(n); };
  b.appendChild(chartFieldRow("data", dataBtn, "Open the overlay editor for JSON, XML or CSV data, or a reactive state expression."));
  var dataHint = document.createElement("div");
  dataHint.className = "hint";
  dataHint.textContent = "The data button opens a movable, resizable editor with format detection, validation, file loading and samples.";
  b.appendChild(dataHint);

  var fields = {};
  function textField(label, key, placeholder, hintText) {
    var input = document.createElement("input");
    var ex = chartParamExpr(n, key);
    input.placeholder = ex ? "expression: " + ex : (placeholder || "");
    input.value = n[key] || "";
    input.oninput = function(ev){ snap("chart " + label); chartClearParam(n, key); n[key] = ev.target.value; setOut(); autoPreview(); };
    var row = chartFieldRow(label, input, hintText);
    b.appendChild(row);
    return row;
  }
  fields.x = textField("x field", "chartX", "month", "Object/CSV column used for labels. Empty auto-detects.");
  fields.y = textField("y field", "chartY", "value", "Object/CSV column used for numeric values. Empty auto-detects.");
  fields.pointWidth = textField("point width", "chartPointWidth", "48", "Horizontal pixels reserved per point in scroll mode.");
  fields.pointSize = textField("point size", "chartPointSize", "3", "Radius of data points in pixels (1-24). Used by point-based charts.");
  var tableSel = document.createElement("select");
  [["chart", "Chart only"], ["table", "Table only"], ["both", "Chart + table"]].forEach(function(v){
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1]; tableSel.appendChild(o);
  });
  tableSel.value = n.chartTableView === "gantt" ? "chart" : (n.chartTableView || "chart");
  tableSel.onchange = function(){ snap("chart table view"); chartClearParam(n, "chartTableView"); n.chartTableView = tableSel.value; setOut(); autoPreview(); };
  fields.tableView = chartFieldRow("table view", tableSel, "Chart only, accessible data table only, or both.");
  b.appendChild(fields.tableView);
  fields.setTableLabels = function(type) {
    var gantt = type === "gantt";
    tableSel.options[0].textContent = gantt ? "Timeline" : "Chart only";
    tableSel.options[1].textContent = "Table only";
    tableSel.options[2].textContent = gantt ? "Timeline + table" : "Chart + table";
  };
  fields.zoom = textField("initial zoom", "chartZoom", "1", "Starting zoom for zoom mode (0.5 to 4; Gantt timeline 0.5 to 24).");
  fields.bins = textField("bins", "chartBins", "10", "Histogram bins (1-100). Ignored by other chart types.");
  fields.min = textField("min", "chartMin", "0", "Gauge minimum. Empty uses 0.");
  fields.max = textField("max", "chartMax", "100", "Gauge maximum. Empty uses 100 or the value.");

  var ovSel = document.createElement("select");
  [["scroll", "Scrollbars"], ["zoom", "Zoom controls"], ["fit", "Fit to frame"]].forEach(function(v){
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1]; ovSel.appendChild(o);
  });
  ovSel.value = n.chartOverflow || "scroll";
  ovSel.onchange = function(){ snap("chart overflow"); chartClearParam(n, "chartOverflow"); n.chartOverflow = ovSel.value; setOut(); autoPreview(); };
  b.appendChild(chartFieldRow("overflow", ovSel, "Scroll: grow and scroll. Zoom: add zoom controls. Fit: squeeze into the frame."));
  var sizeHint = document.createElement("div");
  sizeHint.className = "hint";
  sizeHint.textContent = "Use the standard Size section for width and height. The chart frame follows the component size.";
  b.appendChild(sizeHint);

  function boolRow(label, key, hintText) {
    var row = document.createElement("div");
    row.className = "rowline";
    var lab = document.createElement("label");
    lab.textContent = label;
    lab.title = hintText || "";
    row.appendChild(lab);
    row.appendChild(mkSwitch(!!n[key], false, function(on){ snap("chart " + label); chartClearParam(n, key); n[key] = on; setOut(); autoPreview(); }));
    b.appendChild(row);
    return row;
  }
  fields.grid = boolRow("show grid", "chartGrid", "Draw horizontal grid lines.");
  fields.legend = boolRow("show legend", "chartLegend", "Show a legend under the chart.");
  fields.values = boolRow("show values", "chartValues", "Print numeric values on bars/lines.");
  fields.interactive = boolRow("interactive", "chartInteractive", "Enable Gantt drag/resize and editable Gantt table cells. Tooltips and selection stay available.");
  fields.interactive.title = "Enable chart hover/selection and Gantt editing when a task-change handler is set.";

  var settingsBtn = document.createElement("button");
  settingsBtn.type = "button";
  settingsBtn.className = "btn";
  settingsBtn.style.fontSize = "11px";
  settingsBtn.textContent = graph ? "Graph settings..." : "Chart settings...";
  settingsBtn.title = graph ? "Open the unified graph configuration panel" : "Open the unified chart configuration panel";
  settingsBtn.onclick = function(){ openChartSettings(n); };
  fields.chartSettings = chartFieldRow("settings", settingsBtn, "Appearance, points, tooltip, table, interactions and data fields in one panel.");
  b.appendChild(fields.chartSettings);

  var elemHintsBtn = document.createElement("button");
  elemHintsBtn.type = "button";
  elemHintsBtn.className = "btn";
  elemHintsBtn.style.fontSize = "11px";
  elemHintsBtn.textContent = "Element hints...";
  elemHintsBtn.title = "Edit tag and hint text for each state and link";
  elemHintsBtn.onclick = function(){ openChartElementHints(n); };
  fields.chartElementHints = chartFieldRow("element hints", elemHintsBtn, "Per-element tag and hint text. Uses | for line breaks in JSON, CSV or XML state data.");
  b.appendChild(fields.chartElementHints);

  var extraAttrsBtn = document.createElement("button");
  extraAttrsBtn.type = "button";
  extraAttrsBtn.className = "btn";
  extraAttrsBtn.style.fontSize = "11px";
  extraAttrsBtn.textContent = "Chart attributes...";
  extraAttrsBtn.title = "Edit chart attributes that the inspector does not expose";
  extraAttrsBtn.onclick = function(){ openChartExtraCfg(n); };
  fields.chartExtraAttrs = chartFieldRow("attributes", extraAttrsBtn, "Grid, routing, tag/link style and other chart attributes preserved by the designer.");
  b.appendChild(fields.chartExtraAttrs);

  var sep = document.createElement("div");
  sep.style.cssText = "border-top:1px solid var(--border);margin:8px 0 6px";
  b.appendChild(sep);
  var presetLabel = document.createElement("div");
  presetLabel.className = "rowth";
  presetLabel.textContent = "Style presets";
  b.appendChild(presetLabel);
  var styleSel = document.createElement("select");
  CHART_STYLE_PRESETS.forEach(function(p){
    var o = document.createElement("option"); o.value = p.key; o.textContent = p.label;
    if (chartPresetKey(n) === p.key) o.selected = true;
    styleSel.appendChild(o);
  });
  var customOpt = document.createElement("option");
  customOpt.value = "custom"; customOpt.textContent = "Custom (edit style expr)";
  if (chartPresetKey(n) === "custom") customOpt.selected = true;
  styleSel.appendChild(customOpt);
  styleSel.onchange = function(){
    if (styleSel.value === "custom") { styleSel.value = chartPresetKey(n); return; }
    snap("chart style preset");
    n.styleExpr = chartPresetStyleExpr(styleSel.value);
    setOut(); autoPreview(); renderInspector();
  };
  b.appendChild(chartFieldRow("preset", styleSel, "Writes CSS variables into style expr. You can then edit them freely."));
  var styleHint = document.createElement("div");
  styleHint.className = "hint";
  styleHint.textContent = "Presets set --sui-chart-series-1..8, --sui-chart-bg, --sui-chart-grid and related variables. Edit them in Appearance > style expr.";
  b.appendChild(styleHint);

  var TYPE_HINTS = {
    bar: "Data: one numeric series.",
    barhorizontal: "Data: one numeric series; labels stay readable on the horizontal axis.",
    barstacked: "Data: two or more numeric series; values are stacked.",
    bargrouped: "Data: two or more numeric series; bars sit side by side.",
    combo: "Data: two or more series; first series bars, the others lines.",
    line: "Data: one numeric series.",
    step: "Data: one numeric series; values change at thresholds.",
    smooth: "Data: one numeric series with a curved line.",
    area: "Data: one numeric series with a filled area.",
    pie: "Data: labels and positive values.",
    donut: "Data: labels and positive values.",
    scatter: "Data: one numeric x/y series.",
    histogram: "Data: numeric values; bins controls the bucket count.",
    waterfall: "Data: one series with positive and negative deltas.",
    radar: "Data: one or more series and at least three categories.",
    gauge: "Data: one value; min/max define the arc range.",
    funnel: "Data: one positive series, ordered by stage.",
    heatmap: "Data: exactly two series, y row and value; x is the label column.",
    candlestick: "Data: exactly four series: open, high, low, close.",
    treemap: "Data: one positive series; areas are proportional to values.",
    boxplot: "Data: exactly five series: min, q1, median, q3, max.",
    bullet: "Data: two or three series: value, target, optional max.",
    dumbbell: "Data: exactly two series; dots show the before/after values.",
    gantt: "Data: exactly two numeric series: start and end.",
    state: "Data: states and transitions (JSON object with states/transitions, XML state nodes, or CSV from/to rows)."
  };
  function setRowVisible(row, show) {
    if (row) row.style.display = show ? "" : "none";
  }
  function applyChartTypeVisibility() {
    var type = graph ? "state" : (typeSel.value || "bar");
    var all = !!chartParamExpr(n, "chartType");
    var noPoint = ["pie", "donut", "radar", "gauge", "funnel", "treemap", "bullet", "dumbbell", "gantt", "state"].indexOf(type) >= 0;
    var hasPoints = ["line", "step", "smooth", "area", "scatter", "combo", "radar", "dumbbell"].indexOf(type) >= 0;
    var noGrid = ["pie", "donut", "radar", "gauge", "funnel", "heatmap", "treemap", "bullet"].indexOf(type) >= 0;
    var noValues = ["pie", "donut", "scatter", "gauge", "radar"].indexOf(type) >= 0;
    setRowVisible(fields.x, !graph && (all || (type !== "gauge" && type !== "histogram")));
    setRowVisible(fields.y, !graph && (all || type !== "gauge"));
    setRowVisible(fields.pointWidth, !graph && (all || !noPoint));
    setRowVisible(fields.pointSize, !graph && (all || hasPoints));
    if (fields.setTableLabels) fields.setTableLabels(type);
    setRowVisible(fields.tableView, true);
    setRowVisible(fields.zoom, true);
    setRowVisible(fields.bins, !graph && (all || type === "histogram"));
    setRowVisible(fields.min, !graph && (all || type === "gauge"));
    setRowVisible(fields.max, !graph && (all || type === "gauge"));
    setRowVisible(fields.grid, !graph && (all || !noGrid));
    setRowVisible(fields.legend, !graph && (all || (type !== "gauge" && type !== "funnel")));
    setRowVisible(fields.values, !graph && (all || !noValues));
    if (graph) typeHint.textContent = "Block diagram: states and links.";
    else typeHint.textContent = all
      ? "Chart type is an expression; all parameters are shown and compatibility is checked at runtime."
      : (TYPE_HINTS[type] || "Configure the fields relevant to the selected chart type.");
  }
  typeSel.onchange = function(){
    if (graph) return;
    var newType = typeSel.value;
    var oldType = n.chartType || "bar";
    if (newType === oldType) return;
    var plan = chartTypeChangePlan(n, newType);
    if (plan.needsReplace) {
      typeSel.value = oldType;
      confirmBox("Replace chart data?", chartTypeWarningText(newType, plan), function(){
        chartApplyTypeChange(n, newType, plan);
      });
      return;
    }
    chartApplyTypeChange(n, newType, plan);
  };
  applyChartTypeVisibility();

  a.appendChild(b);
  return a;
}

function treeItemPresetCss(key) {
  for (var i = 0; i < TREE_ITEM_STYLE_PRESETS.length; i++) {
    if (TREE_ITEM_STYLE_PRESETS[i].key === key) return TREE_ITEM_STYLE_PRESETS[i].css;
  }
  return "";
}
function treeItemPresetStyleExpr(key) {
  var css = treeItemPresetCss(key);
  return css ? "cssSafe(" + JSON.stringify(css) + ")" : "";
}
function treeItemPresetKey(n) {
  var cur = n && n.styleExpr ? String(n.styleExpr) : "";
  if (!cur) return "inherit";
  for (var i = 0; i < TREE_ITEM_STYLE_PRESETS.length; i++) {
    if (cur === treeItemPresetStyleExpr(TREE_ITEM_STYLE_PRESETS[i].key)) return TREE_ITEM_STYLE_PRESETS[i].key;
  }
  return "custom";
}
function treeItemControls(n) {
  var rows = [];
  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Keys identify selected items and persist expansion/rename state. Leave empty to use the item path.";
  rows.push(hint);
  var r = document.createElement("div"); r.className = "rowline";
  var l = document.createElement("label"); l.textContent = "key"; r.appendChild(l);
  var i = document.createElement("input"); i.value = n.key || ""; i.placeholder = "stable item key"; i.spellcheck = false;
  i.oninput = function(ev) { snap("set treeitem key"); n.key = ev.target.value; setOut(); autoPreview(); };
  r.appendChild(i); rows.push(r);
  var or = document.createElement("div"); or.className = "rowline";
  var ol = document.createElement("label"); ol.textContent = "initially open"; or.appendChild(ol);
  or.appendChild(mkSwitch(!!n.open, false, function(on){ snap("treeitem open"); n.open = on; setOut(); autoPreview(); }));
  rows.push(or);
  var nsr = document.createElement("div"); nsr.className = "rowline";
  var nsl = document.createElement("label"); nsl.textContent = "node style preset"; nsr.appendChild(nsl);
  var nss = document.createElement("select");
  TREE_ITEM_STYLE_PRESETS.forEach(function(p) {
    var o = document.createElement("option"); o.value = p.key; o.textContent = p.label;
    if (treeItemPresetKey(n) === p.key) o.selected = true;
    nss.appendChild(o);
  });
  var customItem = document.createElement("option"); customItem.value = "custom"; customItem.textContent = "Custom / edit style expr";
  if (treeItemPresetKey(n) === "custom") customItem.selected = true;
  nss.appendChild(customItem);
  nss.onchange = function() {
    if (nss.value === "custom") { nss.value = treeItemPresetKey(n); return; }
    snap("treeitem style preset");
    n.styleExpr = treeItemPresetStyleExpr(nss.value);
    setOut(); renderInspector(); autoPreview();
  };
  nsr.appendChild(nss); rows.push(nsr);
  var nodeStyleHint = document.createElement("div"); nodeStyleHint.className = "hint";
  nodeStyleHint.textContent = "Node presets override row colors, hover/selected colors, padding or radius through style expr.";
  rows.push(nodeStyleHint);
  var er = document.createElement("div"); er.className = "rowline";
  var el = document.createElement("label"); el.textContent = "editable"; er.appendChild(el);
  var es = document.createElement("select");
  [["inherit", "inherit"], ["yes", "yes"], ["no", "read-only"]].forEach(function(v) {
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1];
    var cur = n.editable === true ? "yes" : n.editable === false ? "no" : "inherit";
    if (cur === v[0]) o.selected = true;
    es.appendChild(o);
  });
  es.onchange = function() {
    snap("treeitem editable");
    if (es.value === "inherit") delete n.editable;
    else n.editable = es.value === "yes";
    setOut(); autoPreview();
  };
  er.appendChild(es); rows.push(er);
  var ir = document.createElement("div"); ir.className = "rowline";
  var il = document.createElement("label"); il.textContent = "show rename icon"; ir.appendChild(il);
  var is = document.createElement("select");
  [["inherit", "inherit"], ["yes", "show"], ["no", "hide"]].forEach(function(v) {
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1];
    var cur = n.showRenameIcon === true ? "yes" : n.showRenameIcon === false ? "no" : "inherit";
    if (cur === v[0]) o.selected = true;
    is.appendChild(o);
  });
  is.onchange = function() {
    snap("treeitem show rename icon");
    if (is.value === "inherit") delete n.showRenameIcon;
    else n.showRenameIcon = is.value === "yes";
    setOut(); autoPreview();
  };
  ir.appendChild(is); rows.push(ir);
  var ir2 = document.createElement("div"); ir2.className = "rowline";
  var il2 = document.createElement("label"); il2.textContent = "rename icon"; ir2.appendChild(il2);
  var ii2 = document.createElement("input"); ii2.value = n.renameIcon || ""; ii2.placeholder = "inherit / e.g. \u270e"; ii2.maxLength = 16;
  ii2.oninput = function(ev) { snap("treeitem rename icon"); n.renameIcon = ev.target.value; setOut(); autoPreview(); };
  ir2.appendChild(ii2); rows.push(ir2);
  return acc("Tree item", rows);
}

function datepickerControls(n) {
  var rows = [];
  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Uses the browser's native date/time picker. Values are stored as ISO strings: date YYYY-MM-DD, time HH:MM, datetime YYYY-MM-DDTHH:MM. min/max/step are passed to the native input.";
  rows.push(hint);
  function textField(label, value, placeholder, key) {
    var r = document.createElement("div"); r.className = "rowline";
    var l = document.createElement("label"); l.textContent = label; r.appendChild(l);
    var i = document.createElement("input"); i.value = value || ""; i.placeholder = placeholder || ""; i.spellcheck = false;
    i.oninput = function(ev) { snap("set datepicker " + key); n[key] = ev.target.value; setOut(); autoPreview(); };
    r.appendChild(i); rows.push(r);
  }
  textField("value state", n.value, "state name", "value");
  var mr = document.createElement("div"); mr.className = "rowline";
  var ml = document.createElement("label"); ml.textContent = "mode"; mr.appendChild(ml);
  var ms = document.createElement("select");
  [["date", "date"], ["time", "time"], ["datetime", "datetime"]].forEach(function(m) {
    var o = document.createElement("option"); o.value = m[0]; o.textContent = m[1];
    if ((n.mode || "date") === m[0]) o.selected = true;
    ms.appendChild(o);
  });
  ms.onchange = function() { snap("set datepicker mode"); n.mode = ms.value; setOut(); autoPreview(); };
  mr.appendChild(ms); rows.push(mr);
  textField("min", n.min, "e.g. 2026-01-01", "min");
  textField("max", n.max, "e.g. 2026-12-31", "max");
  var sr = document.createElement("div"); sr.className = "rowline";
  var sl = document.createElement("label"); sl.textContent = "step"; sr.appendChild(sl);
  var si = document.createElement("input"); si.type = "number"; si.min = "1"; si.step = "1";
  si.value = n.step || ""; si.placeholder = "seconds (time/datetime)";
  si.oninput = function(ev) { snap("set datepicker step"); n.step = ev.target.value; setOut(); autoPreview(); };
  sr.appendChild(si); rows.push(sr);
  var dr = document.createElement("div"); dr.className = "rowline";
  var dl = document.createElement("label"); dl.textContent = "disabled"; dr.appendChild(dl);
  dr.appendChild(mkSwitch(!!n.disabled, false, function(on){ snap("datepicker disabled"); n.disabled = on; setOut(); autoPreview(); }));
  rows.push(dr);
  var rr = document.createElement("div"); rr.className = "rowline";
  var rl = document.createElement("label"); rl.textContent = "required"; rr.appendChild(rl);
  rr.appendChild(mkSwitch(!!n.required, false, function(on){ snap("datepicker required"); n.required = on; setOut(); autoPreview(); }));
  rows.push(rr);
  return acc("Date/Time picker", rows);
}

function comboboxControls(n) {
  var rows = [];
  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Use a string state for the option list so application code can replace it at runtime. Label=value separates display text from the stored value; use \\, for a comma inside a label. Search supports * and ? wildcards.";
  rows.push(hint);

  // Options binding: empty means a literal list, a name means a state whose
  // string value is the comma-separated list.
  var sr = document.createElement("div"); sr.className = "rowline";
  var sl = document.createElement("label"); sl.textContent = "options state"; sr.appendChild(sl);
  var si = document.createElement("input");
  si.value = n.optionsExpr ? (n.options || "") : "";
  si.placeholder = "state name (empty = literal)";
  si.spellcheck = false;
  si.onchange = function() {
    var v = si.value.trim();
    if (!v) {
      withUndo(function(){ n.options = ""; n.optionsExpr = false; }, "combobox literal options");
      return;
    }
    if (!IDENT.test(v) || RESERVED_NAME.test(v)) {
      setStatus("options state must be a simple identifier");
      si.value = n.optionsExpr ? (n.options || "") : "";
      return;
    }
    withUndo(function(){
      n.options = v;
      n.optionsExpr = true;
      if (!hasState(v)) root.states.push({ name: v, init: "\"\"" });
    }, "bind combobox options state");
  };
  sr.appendChild(si); rows.push(sr);

  var opts = comboboxOptionList(n);
  var listHeader = document.createElement("div"); listHeader.className = "rowth";
  listHeader.textContent = n.optionsExpr ? "Options in " + n.options : "Literal options";
  rows.push(listHeader);

  if (opts == null) {
    var exprHint = document.createElement("div"); exprHint.className = "hint";
    exprHint.textContent = "This state has a non-literal initializer. Its runtime list is managed by code; add rows after giving it a string literal initial value.";
    rows.push(exprHint);
  } else {
    if (!opts.length) {
      var emptyHint = document.createElement("div"); emptyHint.className = "hint";
      emptyHint.textContent = "The list is empty. Add options below; they update the state's initial value.";
      rows.push(emptyHint);
    }
    opts.forEach(function(opt, idx) {
      var r = document.createElement("div"); r.className = "rowline";
      var txt = document.createElement("span"); txt.textContent = comboboxDisplayText(opt); txt.title = comboboxDisplayText(opt); txt.style.flex = "1"; txt.style.minWidth = "0";
      r.appendChild(txt);
      var del = document.createElement("button"); del.className = "btn"; del.textContent = "\u00d7";
      del.title = "Remove option";
      del.onclick = function() {
        var next = opts.slice(); next.splice(idx, 1);
        withUndo(function(){ comboboxWriteOptions(n, next); }, "remove combobox option");
      };
      r.appendChild(del); rows.push(r);
    });

    var ar = document.createElement("div"); ar.className = "rowline";
    var ai = document.createElement("input");
    ai.placeholder = "Add option, e.g. Italy or Spain=es";
    ai.spellcheck = false;
    var ab = document.createElement("button"); ab.className = "btn"; ab.textContent = "Add";
    function addOption() {
      var vals = comboboxSplitRaw(ai.value);
      if (!vals.length) return;
      var next = opts.slice();
      vals.forEach(function(v){ if (next.indexOf(v) < 0) next.push(v); });
      var maxCount = parseInt(n.maxOptions, 10);
      if (n.maxOptions && !isNaN(maxCount) && maxCount > 0 && next.length > maxCount) {
        setStatus("option count exceeds max options (" + maxCount + ")");
        return;
      }
      var maxLen = parseInt(n.maxOptionLength, 10);
      if (n.maxOptionLength && !isNaN(maxLen) && maxLen > 0) {
        for (var i = 0; i < next.length; i++) {
          if (comboboxDisplayText(next[i]).length > maxLen) {
            setStatus("option text exceeds max option length (" + maxLen + ")");
            return;
          }
        }
      }
      withUndo(function(){ comboboxWriteOptions(n, next); }, "add combobox option");
    }
    ab.onclick = addOption;
    ai.onkeydown = function(ev){ if (ev.key === "Enter") { ev.preventDefault(); addOption(); } };
    ar.appendChild(ai); ar.appendChild(ab); rows.push(ar);
    var runtimeHint = document.createElement("div"); runtimeHint.className = "hint";
    runtimeHint.textContent = "At runtime assign a new comma-separated string to " +
      (n.optionsExpr ? n.options : "the bound state") + "; the view re-render refreshes the list.";
    rows.push(runtimeHint);
  }

  var vr = document.createElement("div"); vr.className = "rowline";
  var vl = document.createElement("label"); vl.textContent = "value state"; vr.appendChild(vl);
  var vi = document.createElement("input");
  vi.value = n.value || ""; vi.placeholder = "state name"; vi.spellcheck = false;
  vi.oninput = function(ev) {
    snap("set combobox value state");
    n.value = ev.target.value;
    setOut(); autoPreview();
  };
  vr.appendChild(vi); rows.push(vr);

  var pr = document.createElement("div"); pr.className = "rowline";
  var pl = document.createElement("label"); pl.textContent = "placeholder"; pr.appendChild(pl);
  var pi = document.createElement("input"); pi.value = n.placeholder || ""; pi.placeholder = "Search...";
  pi.oninput = function(ev) { snap("set combobox placeholder"); n.placeholder = ev.target.value; setOut(); autoPreview(); };
  pr.appendChild(pi); rows.push(pr);

  var er = document.createElement("div"); er.className = "rowline";
  var el = document.createElement("label"); el.textContent = "empty text"; er.appendChild(el);
  var ei = document.createElement("input"); ei.value = n.empty || ""; ei.placeholder = "No results";
  ei.oninput = function(ev) { snap("set combobox empty text"); n.empty = ev.target.value; setOut(); autoPreview(); };
  er.appendChild(ei); rows.push(er);

  var mr = document.createElement("div"); mr.className = "rowline";
  var ml = document.createElement("label"); ml.textContent = "max option length"; mr.appendChild(ml);
  var mi = document.createElement("input"); mi.type = "number"; mi.min = "0"; mi.step = "1";
  mi.value = n.maxOptionLength || ""; mi.placeholder = "0 = no limit";
  mi.oninput = function(ev) { snap("set combobox max option length"); n.maxOptionLength = ev.target.value; setOut(); autoPreview(); };
  mr.appendChild(mi); rows.push(mr);
  var mh = document.createElement("div"); mh.className = "hint";
  mh.textContent = "Maximum characters accepted for each option label/value. If any loaded item exceeds it, the whole list is rejected and an error is shown. 0 or empty means no limit.";
  rows.push(mh);

  var cr = document.createElement("div"); cr.className = "rowline";
  var cl = document.createElement("label"); cl.textContent = "max options"; cr.appendChild(cl);
  var ci = document.createElement("input"); ci.type = "number"; ci.min = "0"; ci.step = "1";
  ci.value = n.maxOptions || ""; ci.placeholder = "0 = no limit";
  ci.oninput = function(ev) { snap("set combobox max options"); n.maxOptions = ev.target.value; setOut(); autoPreview(); };
  cr.appendChild(ci); rows.push(cr);
  var ch = document.createElement("div"); ch.className = "hint";
  ch.textContent = "Maximum number of options accepted when the list is formed. If the loaded list is larger, the whole list is rejected and an error is shown.";
  rows.push(ch);

  var rrq = document.createElement("div"); rrq.className = "rowline";
  var rlq = document.createElement("label"); rlq.textContent = "required"; rrq.appendChild(rlq);
  rrq.appendChild(mkSwitch(!!n.required, false, function(on){
    snap("combobox required"); n.required = on; setOut(); autoPreview();
  }));
  rows.push(rrq);

  var dr = document.createElement("div"); dr.className = "rowline";
  var dl = document.createElement("label"); dl.textContent = "disabled"; dr.appendChild(dl);
  dr.appendChild(mkSwitch(!!n.disabled, false, function(on){
    snap("combobox disabled"); n.disabled = on; setOut(); autoPreview();
  }));
  rows.push(dr);
  return acc("Combobox", rows);
}

function tooltipPropertyControls(n) {
  var rows = [];
  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Text shown when this element is hovered or receives keyboard focus.";
  rows.push(hint);
  var r = document.createElement("div"); r.className = "rowline";
  var l = document.createElement("label"); l.textContent = "tooltip text"; r.appendChild(l);
  var i = document.createElement("input");
  i.id = "f_tooltip";
  i.value = n.tooltip || "";
  i.placeholder = "e.g. Saves the current document";
  i.oninput = function(ev) {
    snap("set tooltip");
    n.tooltip = ev.target.value;
    setOut();
    autoPreview();
  };
  r.appendChild(i); rows.push(r);
  var br = document.createElement("div"); br.className = "rowline";
  var b = document.createElement("button"); b.className = "btn"; b.style.fontSize = "11px";
  b.textContent = "Tooltip settings...";
  b.title = "Open global tooltip appearance and timing settings";
  b.onclick = function(){ openTooltipCfg(); };
  br.appendChild(b); rows.push(br);
  var a = acc("Tooltip", rows);
  if (n.tooltip) a.classList.add("open");
  return a;
}

// ---- toast host editor and live preview ---------------------------------
var __toastPreviewKind = "success";
var __toastPreviewMessage = "Preview: saved successfully.";
var __toastPreviewTimer = null;
var __toastPreviewPending = null;

var TOAST_H = ["left", "center", "right"];
var TOAST_V = ["top", "center", "bottom"];

function toastCfg(n) {
  var h = n.toastHAlign || "";
  var v = n.toastVAlign || "";
  var hx = n.toastHOffset || "";
  var vy = n.toastVOffset || "";
  // Compatibility with the earlier position/offset pair.
  if (n.toastPosition) {
    var parts = String(n.toastPosition).split("-");
    if (parts.length === 2) {
      if (!v) v = parts[0];
      if (!h) h = parts[1];
    }
  }
  if (n.toastOffset) {
    if (!hx) hx = n.toastOffset;
    if (!vy) vy = n.toastOffset;
  }
  // Compatibility with the first panel's inline style.
  var m = styleDeclMap(n.style || "");
  if (!h) {
    if (String(m.left || "") === "50%" || String(m.transform || "").indexOf("translateX") >= 0) h = "center";
    else if (m.left && m.left !== "auto") h = "left";
    else h = "right";
  }
  if (!v) {
    if (String(m.top || "") === "50%" || String(m.transform || "").indexOf("translateY") >= 0) v = "center";
    else if (m.bottom && m.bottom !== "auto") v = "bottom";
    else v = "top";
  }
  if (!hx) {
    var lx = String(m.left || "").trim();
    var rx = String(m.right || "").trim();
    hx = (lx && lx !== "auto" && lx !== "50%") ? lx : ((rx && rx !== "auto") ? rx : "16px");
  }
  if (!vy) {
    var ty = String(m.top || "").trim();
    var by = String(m.bottom || "").trim();
    vy = (ty && ty !== "auto" && ty !== "50%") ? ty : ((by && by !== "auto") ? by : "16px");
  }
  if (TOAST_H.indexOf(h) < 0) h = "right";
  if (TOAST_V.indexOf(v) < 0) v = "top";
  return { h: h, v: v, hx: hx || "16px", vy: vy || "16px" };
}
function toastSetAxes(n, cfg, label) {
  snap(label || "toast position");
  n.toastHAlign = cfg.h;
  n.toastVAlign = cfg.v;
  if (cfg.h === "center") delete n.toastHOffset;
  else n.toastHOffset = cfg.hx || "16px";
  if (cfg.v === "center") delete n.toastVOffset;
  else n.toastVOffset = cfg.vy || "16px";
  delete n.toastPosition;
  delete n.toastOffset;
  // Remove placement declarations written by the old panel so the generated
  // CSS classes are not overridden by stale inline top/right/bottom/left.
  var m = styleDeclMap(n.style || "");
  ["position", "top", "right", "bottom", "left", "transform"].forEach(function(k){ delete m[k]; });
  n.style = styleDeclStr(m);
  setOut();
  autoPreview();
  previewToastSoon(n);
}
function toastHostFromSelection() {
  if (selBody || !selPath) return null;
  var n = byPath(selPath, root.children);
  return n && n.kind === "toast" ? n : null;
}
// The preview contains every page and every host in source order. Target the
// selected host by ordinal as well as by data-sui-id: the latter can be offset
// when non-rendered helper nodes are present, and the designer must never show
// a sample in a different toast host.
function toastHostOrdinal() {
  if (selBody || !selPath) return -1;
  var pages = root.pages || [{ children: root.children }];
  var ordinal = 0;
  for (var pi = 0; pi < pages.length; pi++) {
    var found = false;
    (function walk(nodes, path) {
      for (var i = 0; i < nodes.length && !found; i++) {
        var n = nodes[i];
        var p = path.concat([i]);
        if (n.kind === "toast") {
          if (pi === root.page && samePath(p, selPath)) { found = true; return; }
          ordinal++;
        }
        if (n.children && n.children.length) walk(n.children, p);
      }
    })(pages[pi].children || [], []);
    if (found) return ordinal;
  }
  return -1;
}
function toastSelectorFor(n) {
  var all = n && n.class ? n.class.trim().split(/\s+/).filter(Boolean) : [];
  var cls = "";
  for (var i = 0; i < all.length; i++) {
    if (all[i] !== "suitbl") { cls = all[i]; break; }
  }
  if (!cls) return "";
  return "." + cls + " .sui-toast";
}
function toastCardRule(n) {
  var sel = toastSelectorFor(n);
  return sel ? findRuleBySelector(sel) : null;
}
function toastCardCss(n) {
  var r = toastCardRule(n);
  if (!r || !r.decls) return "";
  return r.decls.map(function(d){ return d.prop + ": " + d.value + ";"; }).join(" ");
}
function applyToastCardCss(n, text) {
  var r = toastCardRule(n);
  var next = styleDeclMap(text || "");
  var map = {};
  if (r && r.decls) {
    r.decls.forEach(function(d){ if (!(d.prop in next)) map[d.prop] = ""; });
  }
  Object.keys(next).forEach(function(k){ map[k] = next[k]; });
  if (Object.keys(map).length) setCssProps(map);
}
function showToastPreviewNow(n) {
  var f = document.getElementById("prev");
  var w = f && f.contentWindow;
  if (!n || !w || typeof w.__sui_toast_show !== "function") return false;
  try {
    w.__sui_toast_clear();
    var ord = toastHostOrdinal();
    if (ord >= 0 && typeof w.__sui_toast_show_at === "function") {
      w.__sui_toast_show_at(ord, __toastPreviewMessage || "Preview toast", __toastPreviewKind || "success", 0);
    } else {
      w.__sui_toast_show(__toastPreviewMessage || "Preview toast", __toastPreviewKind || "success", 0);
    }
  } catch (e) { return false; }
  return true;
}
function bindToastPreviewLoad() {
  var f = document.getElementById("prev");
  if (!f || f.__suiToastLoadHook) return;
  f.__suiToastLoadHook = true;
  f.addEventListener("load", function(){
    var n = __toastPreviewPending;
    if (!n) return;
    if (__toastPreviewTimer) clearTimeout(__toastPreviewTimer);
    // Let the freshly loaded runtime create its toast host first.
    __toastPreviewTimer = setTimeout(function(){ showToastPreviewNow(n); }, 80);
  });
}
function previewToastSoon(n) {
  if (!n || n.kind !== "toast") { __toastPreviewPending = null; return; }
  __toastPreviewPending = n;
  bindToastPreviewLoad();
  if (__toastPreviewTimer) clearTimeout(__toastPreviewTimer);
  var tries = 0;
  function tick() {
    if (showToastPreviewNow(n)) return;
    if (++tries < 40) __toastPreviewTimer = setTimeout(tick, 100);
  }
  // Wait for the render request started by autoPreview/setCssOnClass; the
  // iframe load hook above also refreshes the sample on the new document.
  __toastPreviewTimer = setTimeout(tick, 260);
}
function toastControls(n) {
  var cfg = toastCfg(n);
  var rows = [];
  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Choose the horizontal and vertical anchors. An offset field appears only when that axis is anchored to an edge; centered axes need no offset. A bare number is px; write %, em, rem or vw explicitly for other units.";
  rows.push(hint);

  function row(label) {
    var d = document.createElement("div"); d.className = "rowline";
    var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
    rows.push(d); return d;
  }

  var hr = row("horizontal position");
  var hs = document.createElement("select");
  [["left", "left"], ["center", "center"], ["right", "right"]].forEach(function(o){
    var e = document.createElement("option"); e.value = o[0]; e.textContent = o[1];
    if (cfg.h === o[0]) e.selected = true;
    hs.appendChild(e);
  });
  hs.onchange = function(){
    cfg.h = hs.value;
    toastSetAxes(n, cfg, "horizontal position");
    renderInspector();
  };
  hr.appendChild(hs);

  var vr = row("vertical position");
  var vs = document.createElement("select");
  [["top", "top"], ["center", "center"], ["bottom", "bottom"]].forEach(function(o){
    var e = document.createElement("option"); e.value = o[0]; e.textContent = o[1];
    if (cfg.v === o[0]) e.selected = true;
    vs.appendChild(e);
  });
  vs.onchange = function(){
    cfg.v = vs.value;
    toastSetAxes(n, cfg, "vertical position");
    renderInspector();
  };
  vr.appendChild(vs);

  function offsetRow(label, key, value) {
    var r = row(label);
    var input = document.createElement("input");
    input.value = value;
    input.placeholder = "16px (or 5%)";
    var timer = null;
    function commit(normalize) {
      cfg[key] = normLen(input.value) || "16px";
      // While typing, never rewrite the input value: a debounced number such
      // as "2" must not become "2px" before the user types the second digit.
      if (normalize) input.value = cfg[key];
      toastSetAxes(n, cfg, label);
    }
    input.onchange = function(){ commit(true); };
    input.oninput = function() {
      if (timer) clearTimeout(timer);
      timer = setTimeout(function(){ commit(false); }, 400);
    };
    r.appendChild(input);
  }
  if (cfg.h !== "center") offsetRow("horizontal offset", "hx", cfg.hx);
  if (cfg.v !== "center") offsetRow("vertical offset", "vy", cfg.vy);
  var dr = row("default duration");
  var di = document.createElement("input"); di.type="number"; di.min="0"; di.step="100"; di.style.width="88px"; di.value=n.toastDuration || 4000;
  var dTimer = null;
  function commitDuration(normalize) {
    var v = parseInt(di.value, 10);
    if (isNaN(v) || v < 0) v = 4000;
    if (normalize) di.value = v;
    n.toastDuration = v;
    snap("toast duration"); setOut(); autoPreview(); previewToastSoon(n);
  }
  di.onchange = function(){ commitDuration(true); };
  di.oninput = function(){ if (dTimer) clearTimeout(dTimer); dTimer = setTimeout(function(){ commitDuration(false); }, 400); };
  dr.appendChild(di);

  var mr = row("max visible");
  var mi = document.createElement("input"); mi.type="number"; mi.min="1"; mi.max="8"; mi.style.width="64px"; mi.value=n.toastMax || 4;
  var mTimer = null;
  function commitMax(normalize) {
    var v = parseInt(mi.value, 10);
    if (isNaN(v)) v = 4;
    v = Math.max(1, Math.min(8, v));
    if (normalize) mi.value = v;
    n.toastMax = v;
    snap("toast max"); setOut(); autoPreview(); previewToastSoon(n);
  }
  mi.onchange = function(){ commitMax(true); };
  mi.oninput = function(){ if (mTimer) clearTimeout(mTimer); mTimer = setTimeout(function(){ commitMax(false); }, 400); };
  mr.appendChild(mi);

  var ir = row("show icons");
  ir.appendChild(mkSwitch(n.toastIcons !== false, false, function(on){
    snap("toast icons"); n.toastIcons=on; setOut(); autoPreview(); previewToastSoon(n);
  }));

  var pkr = row("preview kind");
  var pk = document.createElement("select");
  [["info","info"],["success","success"],["warning","warning"],["error","error"]].forEach(function(o){
    var e=document.createElement("option"); e.value=o[0]; e.textContent=o[0]; if(__toastPreviewKind===o[0]) e.selected=true; pk.appendChild(e);
  });
  pk.onchange=function(){ __toastPreviewKind=pk.value; previewToastSoon(n); };
  pkr.appendChild(pk);

  var pmr = row("preview message");
  var pmi = document.createElement("input"); pmi.value=__toastPreviewMessage;
  pmi.onchange=function(){ __toastPreviewMessage=pmi.value || "Preview toast"; previewToastSoon(n); };
  pmr.appendChild(pmi);

  var btr = document.createElement("div"); btr.className = "rowline";
  var show = document.createElement("button"); show.className="btn"; show.style.fontSize="11px"; show.textContent="Show preview";
  show.onclick=function(){ previewToastSoon(n); };
  var clear = document.createElement("button"); clear.className="btn"; clear.style.fontSize="11px"; clear.textContent="Clear preview";
  clear.onclick=function(){ __toastPreviewPending=null; var f=document.getElementById("prev"); if(f && f.contentWindow && f.contentWindow.__sui_toast_clear) f.contentWindow.__sui_toast_clear(); };
  btr.appendChild(show); btr.appendChild(clear); rows.push(btr);

  var a = acc("Toast", rows);
  a.classList.add("open");
  return a;
}

// Inspector section disclosure state. Every section starts closed; opening or
// closing one is remembered across selection changes, re-renders and reloads.
var __inspectorSectionOpen = {};
try {
  __inspectorSectionOpen = JSON.parse(localStorage.getItem("sui.inspector.sections") || "{}") || {};
} catch (e) { __inspectorSectionOpen = {}; }
function saveInspectorSections() {
  try { localStorage.setItem("sui.inspector.sections", JSON.stringify(__inspectorSectionOpen)); } catch (e) {}
}
function directInspectorChild(el, cls) {
  for (var i = 0; i < el.children.length; i++) {
    if (el.children[i].classList && el.children[i].classList.contains(cls)) return el.children[i];
  }
  return null;
}
function syncInspectorSections() {
  var sections = document.querySelectorAll("#inspector .acc, #inspector .ins-sec");
  for (var i = 0; i < sections.length; i++) {
    var a = sections[i];
    var head = directInspectorChild(a, a.classList.contains("acc") ? "acchead" : "ins-sec-h");
    if (!head) continue;
    var titleEl = head.querySelector("span") || head;
    var title = (titleEl.textContent || "").trim();
    if (a.classList.contains("ins-sec")) {
      var body = directInspectorChild(a, "ins-sec-body");
      if (!body) {
        body = document.createElement("div");
        body.className = "ins-sec-body";
        var moved = [];
        var child = head.nextSibling;
        while (child) {
          moved.push(child);
          child = child.nextSibling;
        }
        moved.forEach(function(x){ body.appendChild(x); });
        a.appendChild(body);
      }
    }
    a.classList.toggle("open", !!__inspectorSectionOpen[title]);
    (function(section, sectionTitle, header) {
      header.onclick = function() {
        section.classList.toggle("open");
        __inspectorSectionOpen[sectionTitle] = section.classList.contains("open");
        saveInspectorSections();
      };
    })(a, title, head);
  }
}

function acc(title, rows) {
  var a = document.createElement("div"); a.className = "acc";
  var h = document.createElement("div"); h.className = "acchead";
  h.innerHTML = '<span>' + title + '</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div"); b.className = "accbody";
  rows.forEach(function(r){ b.appendChild(r); });
  a.appendChild(b);
  return a;
}
function currentRule() {
  var th = toastHostFromSelection();
  if (th) {
    var partSel = toastSelectorFor(th);
    return partSel ? findRuleBySelector(partSel) : null;
  }
  var sel = selClass();
  return findRule(sel);
}
function txt(prop, label, ph) {
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var i = document.createElement("input"); i.value = cssFieldValue(prop, currentRule());
  i.placeholder = ph || "auto";
  i.onchange = function(){ setCssProp(prop, NO_PX[prop] ? i.value : normLen(i.value)); };
  d.appendChild(i);
  return d;
}
function fillControls() {
  var d = document.createElement("div"); d.className = "rowline";
  d.appendChild(lbl("fill"));
  var w = document.createElement("button"); w.className = "btn"; w.style.cssText = "font-size:11px;padding:3px 8px"; w.title = "Occupy all available width"; w.textContent = "width";
  w.onclick = function(){ setCssProp("width", "100%"); };
  var h = document.createElement("button"); h.className = "btn"; h.style.cssText = "font-size:11px;padding:3px 8px"; h.title = "Occupy all available height"; h.textContent = "height";
  h.onclick = function(){ setCssProp("height", "100%"); };
  d.appendChild(w); d.appendChild(h);
  return d;
}
// ---- responsive / media-query editor ---------------------------------------
var BREAKPOINTS = [
  { key: "desktop", label: "Desktop", query: "(min-width: 1024px)" },
  { key: "tablet",  label: "Tablet",  query: "(min-width: 768px) and (max-width: 1023px)" },
  { key: "mobile",  label: "Mobile",  query: "(max-width: 767px)" }
];
function responsiveControls() {
  var a = document.createElement("div"); a.className = "acc";
  var h = document.createElement("div"); h.className = "acchead";
  h.innerHTML = '<span>Responsive</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div"); b.className = "accbody";

  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Add custom @media rules for the selected element.";
  b.appendChild(hint);

  var sel = document.createElement("select");
  BREAKPOINTS.forEach(function(bp){
    var o = document.createElement("option"); o.value = bp.key; o.textContent = bp.label + "  " + bp.query; sel.appendChild(o);
  });
  b.appendChild(sel);

  var prop = document.createElement("input"); prop.placeholder = "property (e.g. font-size)"; b.appendChild(prop);
  var val = document.createElement("input"); val.placeholder = "value (e.g. 12px)"; b.appendChild(val);

  var add = document.createElement("button"); add.className = "btn"; add.style.cssText = "font-size:11px;margin-top:4px"; add.textContent = "+ Add media rule";
  add.onclick = function() {
    if (!selPath && !selBody) { setStatus("select an element first"); return; }
    var bp = null;
    BREAKPOINTS.forEach(function(x){ if (x.key === sel.value) bp = x; });
    if (!bp || !prop.value || !val.value) { setStatus("fill property and value"); return; }
    addMediaRule(bp, prop.value.trim(), val.value.trim());
  };
  b.appendChild(add);

  var view = document.createElement("button"); view.className = "btn"; view.style.cssText = "font-size:11px;margin-top:4px"; view.textContent = "Edit raw media queries";
  view.onclick = function(){ openMediaEditor(); };
  b.appendChild(view);

  a.appendChild(b);
  return a;
}
function addMediaRule(bp, prop, value) {
  snap("media " + bp.label + " " + prop);
  var cls = selBody ? "body" : (selPath ? ensureClass(byPath(selPath, root.children)) : "");
  if (!cls) return;
  var block = "\n@media " + bp.query + " {\n\t." + cls + " { " + prop + ": " + value + "; }\n}\n";
  root.css += block;
  scheduleCssSync();
}
function openMediaEditor() {
  var m = document.getElementById("media-editor"); if (!m) return;
  m.classList.add("on");
  document.getElementById("media-css").value = extractMediaBlocks(root.css);
}
function extractMediaBlocks(css) {
  var out = [], depth = 0, cur = "", i = 0, n = (css || "").length;
  var text = css || "";
  while (i < n) {
    var c = text[i];
    if (text.slice(i, i+7) === "@media ") {
      var start = i;
      var j = i;
      var brace = 0;
      while (j < n) {
        if (text[j] === '{') brace++;
        else if (text[j] === '}') { brace--; if (brace === 0) { j++; break; } }
        j++;
      }
      out.push(text.slice(start, j));
      i = j;
      continue;
    }
    i++;
  }
  return out.join("").trim();
}
function commitMediaEditor() {
  var text = document.getElementById("media-css").value;
  var existing = extractMediaBlocks(root.css);
  var clean = root.css;
  if (existing) clean = clean.replace(existing, "");
  snap("media rules");
  root.css = (clean.trim() + "\n" + text).trim();
  closeMediaEditor();
  scheduleCssSync(); setOut(); autoPreview();
}
function closeMediaEditor() {
  var m = document.getElementById("media-editor");
  if (m) m.classList.remove("on");
}
// Length-ish values accept px/em/%/... units; a bare number means px. Props
// where a unitless value is meaningful (line-height, shorthand border) are left
// untouched.
var NO_PX = { "line-height": 1, border: 1, "z-index": 1, "grid-column": 1, "grid-row": 1, opacity: 1 };
function normLen(v) {
  v = (v || "").trim();
  if (v === "" || /^-?\d*\.?\d+$/.test(v) === false) return v;
  return v + "px";
}
function selBox(prop, label, opts) {
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var s = document.createElement("select");
  var cur = cssFieldValue(prop, currentRule());
  var list = opts.slice();
  if (cur && list.indexOf(cur) < 0) list.push(cur);
  list.forEach(function(o){ var e = document.createElement("option"); e.value = o; e.textContent = o === "" ? "(unset)" : o; if (o === cur) e.selected = true; s.appendChild(e); });
  s.onchange = function(){ setCssProp(prop, s.value); };
  d.appendChild(s);
  return d;
}
function color(prop, label) {
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var f = makeColorField(cssFieldValue(prop, currentRule()), function(v){ setCssProp(prop, v); });
  d.appendChild(f);
  return d;
}
function fontBox() {
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = "font"; d.appendChild(l);
  var s = document.createElement("select");
  var cur = cssFieldValue("font-family", currentRule());
  var w = document.createElement("div"); w.style.cssText = "display:flex;flex:1;gap:4px;align-items:center;min-width:0";
  var up = document.createElement("button"); up.className = "btn"; up.textContent = "+"; up.title = "Upload font"; up.style.cssText = "padding:4px 9px;font-size:12px;flex:none";
  up.onclick = function(){ var f = document.getElementById("fontfile"); f.value = ""; f.click(); };
  function fill() {
    var list = [""].concat(BUILTIN_FONTS, uploadedFonts);
    if (cur && list.indexOf(cur) < 0) list.push(cur);
    s.innerHTML = "";
    list.forEach(function(o){ var e = document.createElement("option"); e.value = o; e.textContent = o === "" ? "(unset)" : o; if (o === cur) e.selected = true; s.appendChild(e); });
  }
  fill();
  s.onchange = function(){ setCssProp("font-family", s.value); };
  w.appendChild(s);
  w.appendChild(up);
  d.appendChild(w);
  return d;
}
function uploadFont(file) {
  if (!file) return;
  var family = file.name.replace(/\.[^.]+$/, "").replace(/[^a-zA-Z0-9_-]+/g, "");
  if (!family) family = "CustomFont";
  var fd = new FormData();
  fd.append("font", file);
  fetch("/__gui__/font?name=" + encodeURIComponent(family), { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) { setStatus("font upload failed: " + (j.error || "unknown")); return; }
    if (uploadedFonts.indexOf(j.family) < 0) uploadedFonts.push(j.family);
    setCssProp("font-family", j.family);
    setStatus("font " + j.family + " uploaded");
  }).catch(function(){ setStatus("font upload failed"); });
}
function loadRemoteFonts() {
  fetch("/__gui__/fonts").then(function(r){ return r.json(); }).then(function(j){
    if (j && j.fonts) uploadedFonts = j.fonts.slice();
  }).catch(function(){});
}
function range(prop, label, min, max, unit) {
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var v = cssFieldValue(prop, currentRule()); var num = parseFloat(v) || 0;
  var r = document.createElement("input"); r.type = "range"; r.min = min; r.max = max; r.value = num;
  var t = document.createElement("input"); t.style.width = "52px"; t.value = v;
  r.oninput = function(){ t.value = r.value + unit; };
  r.onchange = function(){ setCssProp(prop, r.value + unit); };
  t.onchange = function(){ var raw = t.value.trim(); if (/^-?\d*\.?\d+$/.test(raw)) raw = raw + unit; var nn = parseFloat(raw) || 0; r.value = nn; setCssProp(prop, raw); };
  d.appendChild(r); d.appendChild(t);
  return d;
}
function normColor(v) { return /^#[0-9a-fA-F]{3,8}$/.test(v || "") ? v : "#000000"; }
function lbl(text) { var l = document.createElement("label"); l.textContent = text; return l; }
// ---- color field: visible color bar + selection popup -----------------------
// A color selector is a wide BAR painting the chosen color (like the native
// input looked). Clicking the bar opens a popup panel with: the native OS
// color picker, a hex field, and the reusable palette of previously chosen
// project colors. Palette swattches and any pick route through one apply()
// which also registers the color in the palette.
var palColors = [];
function registerSwatch(v) {
  v = String(v || "").trim().toLowerCase();
  if (!/^#[0-9a-f]{3,8}$/i.test(v)) return;
  var i = palColors.indexOf(v);
  if (i === 0) return;
  if (i > 0) palColors.splice(i, 1);
  palColors.unshift(v);
  if (palColors.length > 28) palColors.length = 28;
  prefs.palette = palColors.slice();
  savePrefs();
  renderPalStrips();
}
function fillPalStrip(s, onPick) {
  s.innerHTML = "";
  palColors.forEach(function(v) {
    var sw = document.createElement("button");
    sw.className = "sw"; sw.style.background = v;
    sw.title = v + "  \u2014 click: apply to this field; right-click: remove from palette";
    sw.onclick = function(){ onPick(v); };
    sw.oncontextmenu = function(ev) {
      ev.preventDefault();
      var i = palColors.indexOf(v);
      if (i >= 0) palColors.splice(i, 1);
      prefs.palette = palColors.slice();
      savePrefs();
      renderPalStrips();
    };
    s.appendChild(sw);
  });
  s.style.display = palColors.length ? "" : "none";
}
function renderPalStrips() {
  document.querySelectorAll(".color-pop .swstrip").forEach(function(s) {
    if (s.__onpick) fillPalStrip(s, s.__onpick);
  });
}
function closeColorPops() {
  document.querySelectorAll(".color-pop.on").forEach(function(p){ p.classList.remove("on"); });
}
(function () {
  var ins = document.getElementById("inspector");
  if (ins) ins.addEventListener("scroll", closeColorPops);
  window.addEventListener("resize", closeColorPops);
})();
document.addEventListener("mousedown", function(ev) {
  var t = ev.target;
  if (t && (t.closest(".color-pop") || t.closest(".colorbar"))) return;
  closeColorPops();
}, true);
document.addEventListener("keydown", function(ev){ if (ev.key === "Escape") closeColorPops(); });
function palSwatchTone(v) {
  var h = v.replace("#", "");
  if (h.length === 3) h = h.charAt(0) + h.charAt(0) + h.charAt(1) + h.charAt(1) + h.charAt(2) + h.charAt(2);
  if (h.length >= 6) {
    var r = parseInt(h.substr(0, 2), 16) || 0;
    var g = parseInt(h.substr(2, 2), 16) || 0;
    var b = parseInt(h.substr(4, 2), 16) || 0;
    return (0.299 * r + 0.587 * g + 0.114 * b) > 150;
  }
  return false;
}
function colorPicker(name, value, onchange) {
  return makeColorField(normColor(value), onchange);
}
// makeColorField(cur, onchange): the visible color bar + the popup panel.
// onchange(v) runs for every applied pick (native picker, hex field, palette
// swatch) AFTER the color has been registered in the project palette. The
// wrapper carries __inp — a hidden mirror input — so legacy readers (the
// shadow composite) keep working unchanged.
function makeColorField(cur, onchange) {
  var wrap = document.createElement("div");
  wrap.style.cssText = "display:flex;flex-direction:column;gap:3px;min-width:0;flex:1;position:relative";
  var bar = document.createElement("button");
  bar.type = "button";
  bar.className = "colorbar";
  bar.title = "Click: opens the color picker (panel with the project palette)";
  var lab = document.createElement("span"); lab.className = "colorbar-lab";
  bar.appendChild(lab);
  var hidden = document.createElement("input");
  hidden.type = "color";
  hidden.style.display = "none";
  wrap.__inp = hidden;
  var pop = document.createElement("div");
  pop.className = "color-pop";
  var ttl = document.createElement("div"); ttl.className = "color-pop-ttl"; ttl.textContent = "Project colors"; pop.appendChild(ttl);
  var strip = document.createElement("div"); strip.className = "swstrip"; pop.appendChild(strip);
  var nat = document.createElement("input"); nat.type = "color"; nat.title = "Open the system color picker";
  var natRow = document.createElement("div"); natRow.className = "color-pop-row";
  var natLab = document.createElement("span"); natLab.textContent = "Picker";
  natRow.appendChild(natLab); natRow.appendChild(nat);
  pop.appendChild(natRow);
  var hexIn = document.createElement("input"); hexIn.className = "color-hex"; hexIn.spellcheck = false; hexIn.placeholder = "#rrggbb";
  var hexRow = document.createElement("div"); hexRow.className = "color-pop-row";
  var hexLab = document.createElement("span"); hexLab.textContent = "Hex";
  hexRow.appendChild(hexLab); hexRow.appendChild(hexIn);
  pop.appendChild(hexRow);
  strip.__onpick = function(v) { applyColor(v); closeColorPops(); };

  function valid(v) { return /^#[0-9a-f]{3,8}$/i.test(String(v || "")); }
  var curVal = valid(cur) ? normColor(cur) : "";
  function paint() {
    if (valid(curVal)) {
      bar.style.background = curVal;
      bar.classList.remove("unset");
      lab.textContent = curVal.toUpperCase();
      lab.style.color = palSwatchTone(curVal) ? "#1b1b1b" : "#f2f2f2";
      hexIn.value = curVal.toUpperCase();
    } else {
      bar.style.background = "";
      bar.classList.add("unset");
      lab.textContent = "not set";
      lab.style.color = "";
      hexIn.value = "";
    }
  }
  function applyColor(v) {
    v = String(v || "").trim().toLowerCase();
    if (!valid(v)) return;
    if (curVal === v) { closeColorPops(); return; }
    curVal = v;
    hidden.value = v;
    paint();
    registerSwatch(v);
    onchange(v);
  }
  nat.onchange = function(){ applyColor(nat.value); };
  hexIn.onchange = function(){
    var raw = hexIn.value.trim();
    if (raw.charAt(0) !== "#") raw = "#" + raw;
    if (valid(raw)) applyColor(raw.toLowerCase()); else paint();
  };
  hexIn.onkeydown = function(ev){ if (ev.key === "Enter") { ev.preventDefault(); hexIn.onchange(); } };
  bar.addEventListener("click", function(ev) {
    ev.stopPropagation();
    var wasOpen = pop.classList.contains("on");
    closeColorPops();
    if (wasOpen) return;
    pop.classList.add("on");
    fillPalStrip(strip, strip.__onpick);
    // body zoom: rect measurements are in zoomed pixels while left/top are
    // interpreted unzoomed, so divide by menuZoom() (same trick as menus).
    var z = menuZoom();
    var r = bar.getBoundingClientRect();
    var left = r.left / z;
    var top = r.bottom / z;
    var vw = window.innerWidth / z;
    var vh = window.innerHeight / z;
    left = Math.max(8, Math.min(left, vw - pop.offsetWidth - 8));
    if (top + pop.offsetHeight > vh - 8) top = Math.max(8, r.top / z - pop.offsetHeight - 4);
    pop.style.left = left + "px";
    pop.style.top = top + "px";
  });
  wrap.appendChild(bar);
  wrap.appendChild(hidden);
  wrap.appendChild(pop);
  hidden.value = valid(cur) ? normColor(cur) : "#000000";
  paint();
  return wrap;
}

// "Roundness": a 0..100px slider mapped to prefixed border-radius rules.
function roundness(prop, label) {
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var cur = cssFieldValue("border-radius", currentRule()) || declOf(currentRule(), "-webkit-border-radius") || "0px";
  var num = parseInt(cur, 10) || 0;
  var r = document.createElement("input"); r.type = "range"; r.min = 0; r.max = 100; r.value = num;
  var t = document.createElement("input"); t.style.width = "52px"; t.value = num + "px";
  r.oninput = function(){ t.value = r.value + "px"; };
  r.onchange = function(){ setCssProps({ "border-radius": r.value + "px" }); };
  t.onchange = function(){ var nn = parseInt(t.value, 10) || 0; r.value = nn; setCssProps({ "border-radius": t.value }); };
  d.appendChild(r); d.appendChild(t);
  return d;
}
// "Opacity": a 0..100 slider mapped to an opacity rule (100 = fully opaque).
function transparency(prop, label) {
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var raw = cssFieldValue("opacity", currentRule());
  var num = raw === "" ? 100 : Math.round((parseFloat(raw) || 0) * 100);
  var r = document.createElement("input"); r.type = "range"; r.min = 0; r.max = 100; r.value = num;
  var t = document.createElement("input"); t.style.width = "52px"; t.value = num + "%";
  r.oninput = function(){ t.value = r.value + "%"; };
  r.onchange = function(){ setCssProps({ "opacity": (r.value / 100).toFixed(2) }); };
  t.onchange = function(){ var nn = parseInt(t.value, 10); r.value = nn; setCssProps({ "opacity": (nn / 100).toFixed(2) }); };
  d.appendChild(r); d.appendChild(t);
  return d;
}

// Shadow editor: splits a CSS shadow (text-shadow / box-shadow) into its
// component factors (offset x/y, blur, optional spread, color) so the user can
// fill each one separately instead of typing the full value.
function parseShadow(v, spread) {
  var res = { x: "0px", y: "0px", blur: "0px", color: "#000000" };
  if (spread) res.spread = "0px";
  if (!v) return res;
  var color = "";
  var m = v.match(/#[0-9a-fA-F]{3,8}|rgba?\([^)]*\)|hsla?\([^)]*\)/);
  if (m) { color = m[0]; v = v.replace(m[0], ""); }
  var nums = v.match(/-?\d*\.?\d+(?:px|em|rem|%|pt|cm|mm|in|vw|vh)?/g) || [];
  res.x = nums[0] || res.x;
  res.y = nums[1] || res.y;
  res.blur = nums[2] || res.blur;
  if (spread) res.spread = nums[3] || res.spread;
  res.color = color || res.color;
  return res;
}
function shadowBox(prop, label) {
  var spread = prop === "box-shadow";
  var d = document.createElement("div"); d.className = "rowline";
  var l = document.createElement("label"); l.textContent = label; d.appendChild(l);
  var cur = cssFieldValue(prop, currentRule());
  var s = parseShadow(cur, spread);
  var wrap = document.createElement("div"); wrap.style.cssText = "display:flex;flex-direction:column;gap:4px;flex:1;min-width:0";
  var row = document.createElement("div"); row.style.cssText = "display:flex;gap:5px;align-items:center;flex-wrap:wrap";
  function mk(txt, val, width) {
    var cell = document.createElement("span"); cell.style.cssText = "display:flex;align-items:center;gap:2px";
    var t = document.createElement("span"); t.textContent = txt; t.style.cssText = "font-size:10px;color:var(--muted)";
    var i = document.createElement("input"); i.style.cssText = "width:" + width + "px;min-width:0"; i.value = val;
    cell.appendChild(t); cell.appendChild(i);
    row.appendChild(cell);
    return i;
  }
  var xi = mk("X", s.x, 44);
  var yi = mk("Y", s.y, 44);
  var bi = mk("blur", s.blur, 44);
  var si = spread ? mk("spread", s.spread, 52) : null;
  var cp = colorPicker("", s.color, function(){ apply(); });
  row.appendChild(cp);
  wrap.appendChild(row);
  function apply() {
    var parts = [normLen(xi.value) || "0px", normLen(yi.value) || "0px", normLen(bi.value) || "0px"];
    if (spread) parts.push(normLen(si.value) || "0px");
    parts.push(cp.__inp.value);
    setCssProp(prop, parts.join(" "));
  }
  xi.onchange = apply;
  yi.onchange = apply;
  bi.onchange = apply;
  if (spread) si.onchange = apply;
  d.appendChild(wrap);
  return d;
}

// ---- problems / tokens / history -----------------------------------------
// ---- table style: presets + CSS-variable options ---------------------------
// Table-level options (padding, zebra, hover, header theme, borders) ride the
// table's inline style as CSS custom properties, consumed by the auto-managed
// base rules (table.suitbl ...). Variables round-trip as the table's style
// attribute, so every option survives save/reopen without new SUI syntax.
function styleDeclMap(style) {
  var m = {};
  (style || "").split(";").forEach(function(d) {
    var i = d.indexOf(":");
    if (i < 0) return;
    var k = d.slice(0, i).trim();
    var v = d.slice(i + 1).trim();
    if (k) m[k] = v;
  });
  return m;
}
function styleDeclStr(m) {
  var s = "";
  Object.keys(m).forEach(function(k){ s += k + ":" + m[k] + ";"; });
  return s;
}
function tblVar(n, key) {
  var v = styleDeclMap(n.style || "")[key];
  return v === undefined ? "" : v;
}
function tblEnsureClass(n) {
  var cs = (n.class || "").trim().split(/\s+/).filter(Boolean);
  if (cs.indexOf("suitbl") >= 0) return;
  cs.unshift("suitbl");
  n.class = cs.join(" ");
}
// ensureTableBaseCss: inserts the base rule set once (checked by selector, not
// by a comment marker: the css round-trip regenerates text without comments).
function ensureTableBaseCss() {
  if (root.css.indexOf("table.suitbl") >= 0) return;
  var block = [
    "table.suitbl { border-collapse: collapse; }",
    "table.suitbl th, table.suitbl td { padding: var(--t-pad, 6px 10px); text-align: var(--t-align, left); vertical-align: var(--t-valign, middle); height: var(--t-rowh, 34px); }",
    "table.suitbl thead th { background: var(--t-headbg, #202732); color: var(--t-headfg, #f2f4f6); font-weight: var(--t-headfw, 700); text-transform: var(--t-headtt, none); border-bottom: var(--t-headbb, 2px solid #3d4556); }",
    "table.suitbl tbody tr:nth-child(even) td { background: var(--t-zebra, transparent); }",
    "table.suitbl tbody tr:hover td { background: var(--t-hover, transparent); }",
    "table.suitbl td, table.suitbl th { border: var(--t-grid, 1px solid #262c38); }",
    "table.suitbl { border: var(--t-ring, none); border-radius: var(--t-radius, 0); }",
    "table.suitbl { box-shadow: var(--t-shadow, none); }"
  ].join("\n");
  root.css = (root.css.trim() ? root.css.trim() + "\n" : "") + block + "\n";
}
// ensureScrollCss: rules for the optional scroll wrapper around a table. The
// wrapper is a real div in the model (".tscroll"), so the table keeps its layout
// and the wrapper can scroll; "sticky" keeps the header pinned while scrolling.
function ensureScrollCss() {
  if (root.css.indexOf("::-webkit-scrollbar") >= 0) return;
  var block = [
    ".tscroll { overflow: auto; max-width: 100%; max-height: 100%; }",
    ".tscroll > table { margin: 0; }",
    ".tscroll { scrollbar-width: thin; scrollbar-color: var(--sb-thumb, #3d4556) var(--sb-track, #14171c); }",
    ".tscroll::-webkit-scrollbar { width: var(--sb-size, 10px); height: var(--sb-size, 10px); }",
    ".tscroll::-webkit-scrollbar-track { background: var(--sb-track, #14171c); }",
    ".tscroll::-webkit-scrollbar-thumb { background: var(--sb-thumb, #3d4556); border-radius: var(--sb-radius, 6px); }",
    ".tscroll.sticky thead th { position: sticky; top: 0; z-index: 2; }"
  ].join("\n");
  root.css = (root.css.trim() ? root.css.trim() + "\n" : "") + block + "\n";
}
// wrapTableInScroll puts a table inside a .tscroll container (overflow:auto).
// The container takes the parent's size (max-width/max-height 100%), so a fixed
// parent smaller than the table shows scrollbars automatically — vertical only
// when rows are hidden, horizontal only when columns are hidden.
function wrapTableInScroll(path) {
  var node = byPath(path, root.children);
  if (!node || node.kind !== "table") return path;
  var parPath = path.slice(0, -1);
  var par = parPath.length ? byPath(parPath, root.children) : null;
  var arr = par ? par.children : root.children;
  var idx = path[path.length - 1];
  var w = newNode("div");
  w.class = "tscroll";
  w.children = [node];
  arr.splice(idx, 1, w);
  ensureScrollCss();
  return parPath.concat([idx, 0]);
}
// findTableWrap returns the .tscroll container around a table node, if any.
// The wrapper is used to clip the table (rounded corners, scrollbars): CSS
// border-radius on a collapsed table does not clip the cells, but the wrapper
// with overflow:auto does.
function findTableWrap(n) {
  var found = null;
  (function walk(nodes) {
    for (var i = 0; i < nodes.length; i++) {
      var x = nodes[i];
      if (x.kind === "div" && (x.class || "").split(/\s+/).indexOf("tscroll") >= 0 && x.children && x.children[0] === n) { found = x; return; }
      if (x.children && x.children.length) walk(x.children);
    }
  })(root.children);
  return found;
}
// tblApplyVars sets/clears several custom properties atomically and hooks the
// base class + rules. Empty value removes the property (var default kicks in).
function tblApplyVars(n, map, label) {
  var m = styleDeclMap(n.style || "");
  var chg = false;
  Object.keys(map).forEach(function(k) {
    var v = map[k];
    if (v) { if (m[k] !== v) { m[k] = v; chg = true; } }
    else if (k in m) { delete m[k]; chg = true; }
  });
  if (!chg) return;
  var next = styleDeclStr(m);
  withUndo(function(){
    if (n.style === undefined) n.style = "";
    n.style = next;
    tblEnsureClass(n);
    // Outer radius: mirror it on the scroll wrapper so the rounded corners
    // actually clip the cell corners (collapsed tables ignore border-radius).
    if (Object.prototype.hasOwnProperty.call(map, "--t-radius")) {
      var wr = findTableWrap(n);
      if (wr) {
        var wm = styleDeclMap(wr.style || "");
        var rv = map["--t-radius"];
        if (rv) wm["border-radius"] = rv; else delete wm["border-radius"];
        wr.style = styleDeclStr(wm);
      }
    }
    ensureTableBaseCss();
  }, label);
}
var TABLE_PRESETS = [
  ["None", { "--t-pad": "", "--t-align": "", "--t-valign": "", "--t-zebra": "", "--t-hover": "", "--t-grid": "", "--t-ring": "", "--t-radius": "", "--t-shadow": "", "--t-headbg": "", "--t-headfg": "", "--t-headfw": "", "--t-headtt": "", "--t-headbb": "" }],
  ["Minimal", { "--t-pad": "8px 10px", "--t-grid": "", "--t-ring": "", "--t-radius": "", "--t-shadow": "", "--t-headbb": "" }],
  ["Lines",   { "--t-pad": "8px 10px", "--t-grid": "", "--t-ring": "", "--t-radius": "", "--t-shadow": "", "--t-headbb": "2px solid #c9cfe0" }],
  ["Grid", { "--t-pad": "6px 10px", "--t-grid": "1px solid #d3d7e2", "--t-ring": "1px solid #b9c0d4", "--t-radius": "", "--t-shadow": "", "--t-headbb": "" }],
  ["Card",    { "--t-pad": "9px 12px", "--t-grid": "", "--t-ring": "1px solid #e3e6ef", "--t-radius": "10px", "--t-shadow": "0 2px 12px rgba(20,25,55,.16)", "--t-headbb": "" }],
  ["Dotted",  { "--t-pad": "7px 11px", "--t-grid": "1px dotted #9aa4bb", "--t-ring": "1px dotted #9aa4bb", "--t-radius": "", "--t-shadow": "", "--t-headbb": "" }],
  ["Bold",    { "--t-pad": "9px 12px", "--t-grid": "2px solid #b9c0d4", "--t-ring": "2px solid #9aa4bb", "--t-radius": "", "--t-shadow": "", "--t-headbb": "" }],
  ["Rounded", { "--t-pad": "8px 12px", "--t-grid": "", "--t-ring": "1px solid #c9cfe0", "--t-radius": "12px", "--t-shadow": "", "--t-headbb": "" }],
  ["Shadow",  { "--t-pad": "9px 12px", "--t-grid": "", "--t-ring": "1px solid #e3e6ef", "--t-radius": "10px", "--t-shadow": "0 6px 18px rgba(20,25,55,.20)", "--t-headbb": "" }]
];
// cellExtraWrite: structured per-cell declarations (padding, alignment) that
// join the cell's bg/color in the inline style string.
function cellExtraWrite(n, r, c, key, v, label) {
  var sp = n.spans[r][c];
  if (!sp || sp.cs === 0) return;
  var st = n.cellStyles[r][c];
  var m = styleDeclMap(st.extra || "");
  if (v) { if (m[key] === v) return; m[key] = v; }
  else if (!(key in m)) return;
  var next = styleDeclStr(m);
  withUndo(function(){ st.extra = next; }, label);
}
// colExtraWrite: column-wide extras materialize on every visible cell of the
// column (per-cell style strings survive every reparse).
function colExtraWrite(n, col, key, v, label) {
  withUndo(function(){
    for (var rr = 0; rr < n.rows; rr++) {
      var s = n.spans[rr][col];
      if (!s || s.cs === 0) continue;
      var st = n.cellStyles[rr][col];
      var m = styleDeclMap(st.extra || "");
      if (v) m[key] = v; else delete m[key];
      st.extra = styleDeclStr(m);
    }
  }, label);
}
function colExtraVal(n, col, key) {
  for (var rr = 0; rr < n.rows; rr++) {
    var s = n.spans[rr][col];
    if (s && s.cs !== 0) return styleDeclMap(n.cellStyles[rr][col].extra || "")[key] || "";
  }
  return "";
}
function tableControls(n) {
  var a = document.createElement("div");
  a.className = "acc open";
  var h = document.createElement("div"); h.className = "acchead";
  h.innerHTML = '<span>Table</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div"); b.className = "accbody";

  var selLabel = document.createElement("div");
  selLabel.className = "rowth";
  selLabel.textContent = "Table selectors";
  b.appendChild(selLabel);

  // selection mode toggle
  var mode = document.createElement("div");
  mode.style.cssText = "display:flex;gap:4px;margin-bottom:6px";
  [["Table", null], ["Cell", "cell"], ["Row", "row"], ["Column", "col"]].forEach(function(mt){
    var mb = document.createElement("button");
    // The active state follows the MODE only, so a selector stays highlighted
    // right after being chosen (even before anything is picked in the preview).
    var active = (mt[1] === null) ? (selMode === "table") : (selMode === mt[1]);
    mb.className = "btn" + (active ? " on" : "");
    mb.style.fontSize = "11px";
    mb.textContent = mt[0];
    mb.onclick = function(){
      selMode = mt[1] ? mt[1] : "table";
      selCell = null; selRow = null; selCol = null;
      renderInspector();
    };
    mode.appendChild(mb);
  });
  b.appendChild(mode);

  // structure buttons: visually separated from the selectors, on a single row
  var ssep = document.createElement("div");
  ssep.style.cssText = "border-top:1px solid var(--border);margin:8px 0 2px";
  b.appendChild(ssep);
  var slabel = document.createElement("div"); slabel.className = "rowth"; slabel.textContent = "Structure";
  b.appendChild(slabel);
  var btns = document.createElement("div");
  btns.style.cssText = "display:flex;gap:6px;margin-bottom:8px";
  [["+ Col", function(){ tableAddCol(n); }],
   ["\u2212 Col", function(){ tableDelCol(n); }],
   ["+ Row", function(){ tableAddRow(n); }],
   ["\u2212 Row", function(){ tableDelRow(n); }]
  ].forEach(function(bt){
    var x = document.createElement("button");
    x.className = "btn";
    x.style.cssText = "flex:1;font-size:11px;padding:6px 2px;white-space:nowrap";
    x.textContent = bt[0];
    x.onclick = bt[1];
    btns.appendChild(x);
  });
  b.appendChild(btns);

  // ---- opzioni raggruppate in fisarmoniche ----------------------------------
  // ---- scale and scroll: table-specific controls ----------------------------
  // Alignment and sizing are generic element properties: edit them in the
  // Alignment / Size sections like any other element (no duplicates here).
  function tableAutoClass(nn) {
    var cs = (nn.class || "").trim().split(/\s+/).filter(Boolean).filter(function(c){ return c !== "suitbl"; });
    return cs[0] || "";
  }
  // scale the whole table (zoom) + optional scroll wrapper with sticky header
  function tableZoomVal(nn) {
    var cls = tableAutoClass(nn);
    return cls ? declOf(findRule(cls), "zoom") : "";
  }
  function tableWrapInfo() {
    if (!selPath || selPath.length < 2) return null;
    var par = byPath(selPath.slice(0, -1), root.children);
    if (par && par.kind === "div" && (par.class || "").split(/\s+/).indexOf("tscroll") >= 0) return { node: par, path: selPath.slice(0, -1) };
    return null;
  }
  function tableToggleScroll(nn, on) {
    withUndo(function(){
      var parPath = selPath.slice(0, -1);
      var parent = parPath.length ? byPath(parPath, root.children) : null;
      var arr = parent ? parent.children : root.children;
      var idx = selPath[selPath.length - 1];
      var wrap = tableWrapInfo();
      if (on && !wrap) {
        var w = newNode("div"); w.class = "tscroll"; w.style = "max-height:260px";
        w.children = [nn];
        arr.splice(idx, 1, w);
        selPath = parPath.concat([idx, 0]);
        ensureScrollCss();
      } else if (!on && wrap) {
        arr.splice(idx, 1, nn);
        selPath = parPath.concat([idx]);
      }
    });
  }
  function tableSetScrollMaxH(wn, v) {
    withUndo(function(){
      var m = styleDeclMap(wn.style || "");
      if (v) m["max-height"] = v; else delete m["max-height"];
      wn.style = styleDeclStr(m);
    });
  }
  function tableSetScrollVar(wn, prop, v) {
    withUndo(function(){
      var m = styleDeclMap(wn.style || "");
      if (v) m[prop] = v; else delete m[prop];
      wn.style = styleDeclStr(m);
    });
  }
  function tableSetSticky(wn, on) {
    withUndo(function(){
      var cs = (wn.class || "").split(/\s+/).filter(Boolean).filter(function(c){ return c !== "sticky"; });
      if (on) cs.push("sticky");
      wn.class = cs.join(" ");
    });
  }
  // scale: zoom of the whole table. The numeric field always means PERCENT
  // (typing 120 -> 120%, never a 120x zoom), the slider is the quick control.
  function zoomPercent(v) {
    v = (v || "").trim();
    if (!v) return 100;
    var n = parseFloat(v.replace("%", "").replace(",", "."));
    if (isNaN(n)) return 100;
    if (v.indexOf("%") < 0 && n <= 10) n = n * 100; // 1.2 -> 120%
    return Math.round(n);
  }
  function applyZoom(val) {
    var v = (val === null || val === undefined) ? "" : String(val).trim();
    if (!v) { setCssProps({ "zoom": "" }); return; }
    var n = parseFloat(v.replace("%", "").replace(",", "."));
    if (isNaN(n)) return;
    if (v.indexOf("%") < 0 && n <= 10) n = n * 100;
    n = Math.max(10, Math.min(500, Math.round(n)));
    setCssProps({ "zoom": n + "%" });
  }
  var curZoom = zoomPercent(tableZoomVal(n));
  var scr = document.createElement("div"); scr.className = "rowline";
  scr.appendChild(lbl("scale"));
  var srange = document.createElement("input");
  srange.type = "range"; srange.min = "50"; srange.max = "200"; srange.step = "5";
  srange.value = String(Math.max(50, Math.min(200, curZoom)));
  srange.style.cssText = "flex:1;min-width:0;padding:0;accent-color:var(--accent)";
  var snum = document.createElement("input");
  snum.type = "text"; snum.value = curZoom + "%"; snum.spellcheck = false;
  snum.style.cssText = "flex:none;width:56px;text-align:right";
  srange.oninput = function(){ snum.value = srange.value + "%"; };
  srange.onchange = function(){ applyZoom(srange.value + "%"); };
  snum.onchange = function(){
    var p = zoomPercent(snum.value);
    snum.value = p + "%";
    srange.value = String(Math.max(50, Math.min(200, p)));
    applyZoom(p + "%");
  };
  scr.appendChild(srange); scr.appendChild(snum);
  var scw = document.createElement("div"); scw.className = "rowline";
  scw.appendChild(lbl("scroll"));
  var scchk = document.createElement("input"); scchk.type = "checkbox"; scchk.style.flex = "none";
  var winfo = tableWrapInfo();
  scchk.checked = !!winfo;
  scchk.onchange = function(){ tableToggleScroll(n, scchk.checked); };
  scw.appendChild(scchk);
  var lrows = [scr, scw];
  var stickyRow = null;
  if (winfo) {
    var mhr = document.createElement("div"); mhr.className = "rowline";
    mhr.appendChild(lbl("max height"));
    var mhi = document.createElement("input"); mhi.value = styleDeclMap(winfo.node.style || "")["max-height"] || ""; mhi.spellcheck = false; mhi.placeholder = "260px";
    mhi.onchange = function(){ tableSetScrollMaxH(winfo.node, mhi.value.trim()); };
    mhr.appendChild(mhi); lrows.push(mhr);
    // sticky header lives in the Header section (it belongs to the header),
    // but it only has an effect when the table is inside a scroll container.
    stickyRow = document.createElement("div"); stickyRow.className = "rowline";
    stickyRow.appendChild(lbl("sticky header"));
    var stk = document.createElement("input"); stk.type = "checkbox"; stk.style.flex = "none";
    stk.checked = (winfo.node.class || "").split(/\s+/).indexOf("sticky") >= 0;
    stk.onchange = function(){ tableSetSticky(winfo.node, stk.checked); };
    stickyRow.appendChild(stk);
    // scrollbar look: size, corners, thumb and track colors (CSS vars on the
    // wrapper; the defaults follow the app's dark UI).
    [["scrollbar size", "--sb-size", "10px"], ["scrollbar corners", "--sb-radius", "6px"]].forEach(function(sb){
      var rr = document.createElement("div"); rr.className = "rowline";
      rr.appendChild(lbl(sb[0]));
      var ii = document.createElement("input");
      ii.value = styleDeclMap(winfo.node.style || "")[sb[1]] || "";
      ii.spellcheck = false; ii.placeholder = sb[2];
      ii.onchange = function(){ tableSetScrollVar(winfo.node, sb[1], ii.value.trim()); };
      rr.appendChild(ii); lrows.push(rr);
    });
    [["scrollbar thumb", "--sb-thumb", "#3d4556"], ["scrollbar track", "--sb-track", "#14171c"]].forEach(function(sb){
      var rr = document.createElement("div"); rr.className = "rowline";
      rr.appendChild(lbl(sb[0]));
      var cp = colorPicker("color", styleDeclMap(winfo.node.style || "")[sb[1]] || sb[2], function(v){
        tableSetScrollVar(winfo.node, sb[1], v);
      });
      rr.appendChild(cp); lrows.push(rr);
    });
  }
  var lgrp = acc("Scroll & scale", lrows);
  lgrp.classList.add("open");

  function textRow(labelText, key, placeholder) {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl(labelText));
    var inp = document.createElement("input");
    inp.value = tblVar(n, key);
    inp.spellcheck = false;
    inp.placeholder = placeholder || "";
    inp.onchange = function(){ var v = inp.value.trim(); tblApplyVars(n, (function(){ var mm = {}; mm[key] = v; return mm; })(), key.replace("--t-", "")); };
    row.appendChild(inp);
    return row;
  }
  function mkSelRowUI(labelText, key, opts) {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl(labelText));
    var sel = document.createElement("select");
    opts.forEach(function(o) {
      var eo = document.createElement("option"); eo.value = o[1]; eo.textContent = o[0];
      if ((tblVar(n, key) || "") === o[1]) eo.selected = true;
      sel.appendChild(eo);
    });
    sel.onchange = function(){ tblApplyVars(n, (function(){ var mm = {}; mm[key] = sel.value; return mm; })(), key.replace("--t-", "")); };
    row.appendChild(sel);
    return row;
  }
  function borderRow(labelText, key) {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl(labelText));
    function compose(wsty, color) {
      if (!wsty) return "";
      var dsh = wsty.indexOf("dashed") >= 0;
      var w = dsh ? wsty.replace(" dashed", "") : wsty;
      return w + " " + (dsh ? "dashed" : "solid") + " " + color;
    }
    var t3 = tblVar(n, key).split(/\s+/);
    var wsty = t3.length >= 2 ? t3[0] + (t3[1] === "dashed" ? " dashed" : "") : "";
    var col = t3.length >= 3 ? t3[t3.length - 1] : "#c9cfe0";
    var sel = document.createElement("select");
    [["-", ""], ["1px", "1px"], ["2px", "2px"], ["3px", "3px"], ["1px tratteggiato", "1px dashed"]].forEach(function(o) {
      var eo = document.createElement("option"); eo.value = o[1]; eo.textContent = o[0];
      if (o[1] === wsty) eo.selected = true;
      sel.appendChild(eo);
    });
    sel.style.flex = "1"; sel.style.minWidth = "0";
    var cp = colorPicker("color", col, function(v){ if (wsty) tblApplyVars(n, (function(){ var mm = {}; mm[key] = compose(wsty, v); return mm; })(), "table border"); });
    cp.style.flex = "none"; cp.style.width = "54px";
    sel.onchange = function(){
      wsty = sel.value;
      tblApplyVars(n, (function(){ var mm = {}; mm[key] = compose(wsty, col); return mm; })(), "table border");
    };
    row.appendChild(sel); row.appendChild(cp);
    return row;
  }
  function headerRowUI() {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl("header"));
    var sel = document.createElement("select");
    [["none", "0"], ["first row", "1"], ["+ new row", "new"]].forEach(function(o) {
      var eo = document.createElement("option"); eo.value = o[1]; eo.textContent = o[0];
      if ((n.header || 0) === 0 && o[1] === "0") eo.selected = true;
      else if ((n.header || 0) > 0 && o[1] === "1") eo.selected = true;
      sel.appendChild(eo);
    });
    sel.onchange = function(){ tableSetHeader(n, sel.value); };
    row.appendChild(sel);
    return row;
  }
  function toggleRow(labelText, key, defColor) {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl(labelText));
    var cur = tblVar(n, key);
    var chk = document.createElement("input");
    chk.type = "checkbox";
    chk.checked = cur !== "";
    chk.style.flex = "none";
    var cp = colorPicker("color", cur || defColor, function(v){
      cur = v;
      chk.checked = true;
      tblApplyVars(n, (function(){ var mm = {}; mm[key] = v; return mm; })(), labelText);
    });
    chk.onchange = function(){ tblApplyVars(n, (function(){ var mm = {}; mm[key] = chk.checked ? (cur || defColor) : ""; return mm; })(), labelText); };
    cp.style.flex = "1"; cp.style.minWidth = "0";
    row.appendChild(chk); row.appendChild(cp);
    return row;
  }
  // preset dropdown: presets are mutually exclusive, so a select is the right
  // control. Selecting one overwrites the presets' own keys (borders included).
  function currentPresetIndex(nn) {
    for (var i = 0; i < TABLE_PRESETS.length; i++) {
      var map = TABLE_PRESETS[i][1], ok = true;
      for (var k in map) { if (tblVar(nn, k) !== (map[k] || "")) { ok = false; break; } }
      if (ok) return i;
    }
    return -1;
  }
  function presetRow(nn) {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl("preset"));
    var sel = document.createElement("select");
    var cur = currentPresetIndex(nn);
    TABLE_PRESETS.forEach(function(pd, i) {
      var o = document.createElement("option"); o.value = String(i); o.textContent = pd[0];
      if (i === cur) o.selected = true;
      sel.appendChild(o);
    });
    if (cur < 0) {
      var oc = document.createElement("option"); oc.value = "-1"; oc.textContent = "Custom"; oc.selected = true;
      sel.appendChild(oc);
    }
    sel.onchange = function() {
      var i = Number(sel.value);
      if (i >= 0) tblApplyVars(nn, TABLE_PRESETS[i][1], "table preset " + TABLE_PRESETS[i][0]);
    };
    row.appendChild(sel);
    return row;
  }
  // group(title, rows, keys): a nested accordion; auto-opens when the group
  // owns at least one active option, or when nobody does (first discovery).
  function group(title, rows, keys, forceOpen) {
    var g = acc(title, rows);
    var anyActive = keys.some(function(k){ return tblVar(n, k) !== ""; });
    if (forceOpen) g.classList.add("open");
    else if (anyActive) g.classList.add("open");
    return g;
  }
  function noGroupActive(keys) { return !keys.some(function(k){ return tblVar(n, k) !== ""; }); }
  var ASPECTO_KEYS = ["--t-pad", "--t-align", "--t-valign", "--t-zebra", "--t-hover"];
  var BORDI_KEYS = ["--t-grid", "--t-ring", "--t-radius", "--t-shadow"];
  var HEAD_KEYS = ["--t-headbg", "--t-headfg", "--t-headfw", "--t-headtt", "--t-headbb"];
  var anythingActive = !noGroupActive(ASPECTO_KEYS.concat(BORDI_KEYS, HEAD_KEYS));
  b.appendChild(group("Appearance", [
    textRow("cell padding", "--t-pad", "6px 10px"),
    mkSelRowUI("align", "--t-align", [["default", ""], ["left", "left"], ["center", "center"], ["right", "right"], ["justify", "justify"]]),
    mkSelRowUI("vertical", "--t-valign", [["default", ""], ["top", "top"], ["middle", "middle"], ["bottom", "bottom"]]),
    toggleRow("zebra rows", "--t-zebra", "#f0f2f8"),
    toggleRow("hover row", "--t-hover", "#eef1fa"),
    lgrp
  ], ASPECTO_KEYS, !anythingActive));
  var radiusRow = textRow("outer radius", "--t-radius", "8px");
  (function(){ var rl = radiusRow.querySelector("label"); if (rl) rl.title = "Rounds the table's outer corners (not the single cells)."; })();
  b.appendChild(group("Borders", [
    presetRow(n),
    borderRow("cell grid", "--t-grid"),
    borderRow("outer border", "--t-ring"),
    radiusRow
  ], BORDI_KEYS, false));
  var headerRows = [
    headerRowUI(),
    toggleRow("header background", "--t-headbg", "#e5e9f4"),
    toggleRow("header text", "--t-headfg", "#1c2340"),
    mkSelRowUI("header weight", "--t-headfw", [["bold (default)", ""], ["medium", "600"], ["normal", "400"]]),
    mkSelRowUI("header uppercase", "--t-headtt", [["no", ""], ["yes", "uppercase"]])
  ];
  if (stickyRow) {
    headerRows.push(stickyRow);
  } else {
    var shint = document.createElement("div"); shint.className = "hint";
    shint.textContent = "Enable scroll (Appearance) to pin the header while scrolling.";
    headerRows.push(shint);
  }
  b.appendChild(group("Header", headerRows, HEAD_KEYS, false));

  // ---- caption: optional title above the table ------------------------------
  function capStyleGet(prop) { return styleDeclMap(n.capstyle || "")[prop] || ""; }
  function capStyleSet(prop, v) {
    withUndo(function(){
      var m = styleDeclMap(n.capstyle || "");
      if (v) m[prop] = v; else delete m[prop];
      n.capstyle = styleDeclStr(m);
    });
  }
  var capShow = document.createElement("div"); capShow.className = "rowline";
  capShow.appendChild(lbl("show caption"));
  var capChk = document.createElement("input"); capChk.type = "checkbox"; capChk.style.flex = "none";
  capChk.checked = !!n.captionShow;
  capChk.onchange = function(){ withUndo(function(){ n.captionShow = capChk.checked; }); };
  capShow.appendChild(capChk);
  var capRow = document.createElement("div"); capRow.className = "rowline";
  capRow.appendChild(lbl("caption"));
  var capInp = document.createElement("input");
  capInp.value = n.caption || ""; capInp.spellcheck = false; capInp.placeholder = "Table title";
  capInp.onchange = function(){ var v = capInp.value; withUndo(function(){ n.caption = v; if (v.trim()) n.captionShow = true; }); };
  capRow.appendChild(capInp);
  var capSizeRow = document.createElement("div"); capSizeRow.className = "rowline";
  capSizeRow.appendChild(lbl("caption size"));
  var capSize = document.createElement("input"); capSize.value = capStyleGet("font-size"); capSize.spellcheck = false; capSize.placeholder = "15px";
  capSize.onchange = function(){ capStyleSet("font-size", capSize.value.trim()); };
  capSizeRow.appendChild(capSize);
  var capWeightRow = document.createElement("div"); capWeightRow.className = "rowline";
  capWeightRow.appendChild(lbl("caption weight"));
  var capWeight = document.createElement("select");
  [["default", ""], ["bold", "700"], ["semibold", "600"], ["normal", "400"]].forEach(function(o){
    var eo = document.createElement("option"); eo.value = o[1]; eo.textContent = o[0];
    if (capStyleGet("font-weight") === o[1]) eo.selected = true;
    capWeight.appendChild(eo);
  });
  capWeight.onchange = function(){ capStyleSet("font-weight", capWeight.value); };
  capWeightRow.appendChild(capWeight);
  var capAlignRow = document.createElement("div"); capAlignRow.className = "rowline";
  capAlignRow.appendChild(lbl("caption align"));
  var capAlign = document.createElement("select");
  [["left", "left"], ["center", "center"], ["right", "right"]].forEach(function(o){
    var eo = document.createElement("option"); eo.value = o[1]; eo.textContent = o[0];
    if ((capStyleGet("text-align") || "left") === o[1]) eo.selected = true;
    capAlign.appendChild(eo);
  });
  capAlign.onchange = function(){ capStyleSet("text-align", capAlign.value); };
  capAlignRow.appendChild(capAlign);
  var capColorRow = document.createElement("div"); capColorRow.className = "rowline";
  capColorRow.appendChild(lbl("caption color"));
  var capColor = colorPicker("color", capStyleGet("color"), function(v){ capStyleSet("color", v); });
  capColorRow.appendChild(capColor);
  b.appendChild(group("Caption", [capShow, capRow, capSizeRow, capWeightRow, capAlignRow, capColorRow], [], false));

  // ---- behavior: editable cells / fixed sizing ------------------------------
  function boolRow(labelText, getv, setv, titleText) {
    var row = document.createElement("div"); row.className = "rowline";
    var lb = lbl(labelText); if (titleText) lb.title = titleText;
    row.appendChild(lb);
    var chk = document.createElement("input"); chk.type = "checkbox"; chk.style.flex = "none";
    chk.checked = !!getv();
    chk.onchange = function(){ withUndo(function(){ setv(chk.checked); }); };
    row.appendChild(chk);
    return row;
  }
  var behRows = [
    boolRow("editable", function(){ return n.editable; }, function(v){ n.editable = v; },
      "Cells can be edited by the end user in the rendered app."),
    boolRow("fixed sizes", function(){ return n.fixed; }, function(v){ n.fixed = v; },
      "Rows and columns keep their size; long cell text is truncated and shown in full in an overlay when hovering.")
  ];
  // overlay box look (shown while hovering a truncated cell in fixed mode)
  function popTextRow(labelText, key, ph) {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl(labelText));
    var inp = document.createElement("input");
    inp.value = tblVar(n, key); inp.spellcheck = false; inp.placeholder = ph;
    inp.onchange = function(){ var mm = {}; mm[key] = inp.value.trim(); tblApplyVars(n, mm, "overlay"); };
    row.appendChild(inp);
    return row;
  }
  function popColorRow(labelText, key, def) {
    var row = document.createElement("div"); row.className = "rowline";
    row.appendChild(lbl(labelText));
    var cp = colorPicker("color", tblVar(n, key) || def, function(v){ var mm = {}; mm[key] = v; tblApplyVars(n, mm, "overlay"); });
    row.appendChild(cp);
    return row;
  }
  behRows.push(popColorRow("overlay background", "--pop-bg", "#202732"));
  behRows.push(popColorRow("overlay text", "--pop-color", "#f2f4f6"));
  behRows.push(popTextRow("overlay size", "--pop-size", "13px"));
  behRows.push(popTextRow("overlay max width", "--pop-maxw", "420px"));
  behRows.push(popTextRow("overlay radius", "--pop-radius", "8px"));
  var bhint = document.createElement("div"); bhint.className = "hint";
  bhint.textContent = "Flexible (off): cells adapt to their content. Fixed: sizes stay as set; overflowing text is shortened and shown in the overlay box on hover.";
  behRows.push(bhint);
  b.appendChild(group("Behavior", behRows, [], false));

  // ---- data binding --------------------------------------------------------
  // A name turns the table into a runtime target: table.display() fills it
  // from any tabular payload while keeping these rows/cells as style
  // templates; table.data()/table.json() read the visible values back.
  var dataRows = [];
  var dnRow = document.createElement("div"); dnRow.className = "rowline";
  dnRow.appendChild(lbl("table name"));
  var dname = document.createElement("input");
  dname.value = n.name || ""; dname.spellcheck = false; dname.placeholder = "users";
  dname.onchange = function(){ var v = dname.value.trim(); withUndo(function(){ n.name = v; }); };
  dnRow.appendChild(dname);
  dataRows.push(dnRow);
  var dh = document.createElement("div"); dh.className = "hint";
  dh.textContent = "Fill from code: n, err = table.display(\"users\", data) - data can be JSON, CSV/TSV text, arrays or objects. Read back: table.data, table.json, table.exportText(\"users\", \"csv\"). For HTTP data add the std imports http, json.";
  dataRows.push(dh);
  b.appendChild(group("Data", dataRows, [], false));

  var selCellRowText = [];
  if (selCell && !cellCovered(n, selCell.r, selCell.c)) {
    var tr = document.createElement("div"); tr.className = "rowline";
    var tl = document.createElement("label"); tl.textContent = "text"; tr.appendChild(tl);
    var ti = document.createElement("input"); ti.value = n.cells[selCell.r][selCell.c];
    ti.oninput = function(ev){ snap(); n.cells[selCell.r][selCell.c] = ev.target.value; setOut(); autoPreview(); };
    tr.appendChild(ti); selCellRowText.push(tr);
    // cell styling
    var st = n.cellStyles[selCell.r][selCell.c];
    var c1 = colorPicker("bg", st.bg, function(v){ snap(); st.bg = v; setOut(); autoPreview(); renderInspector(); });
    var c2 = colorPicker("color", st.color, function(v){ snap(); st.color = v; setOut(); autoPreview(); renderInspector(); });
    var cr1 = document.createElement("div"); cr1.className = "rowline"; cr1.appendChild(lbl("background")); cr1.appendChild(c1); selCellRowText.push(cr1);
    var cr2 = document.createElement("div"); cr2.className = "rowline"; cr2.appendChild(lbl("text color")); cr2.appendChild(c2); selCellRowText.push(cr2);
    // per-cell extras: padding and text alignment (stored in the cell's style)
    var cellMap = styleDeclMap(st.extra || "");
    var cpad = document.createElement("div"); cpad.className = "rowline";
    cpad.appendChild(lbl("padding"));
    var pi = document.createElement("input"); pi.value = cellMap["padding"] || ""; pi.spellcheck = false; pi.placeholder = "4px 8px";
    pi.onchange = function(){ cellExtraWrite(n, selCell.r, selCell.c, "padding", pi.value.trim(), "cell padding"); };
    cpad.appendChild(pi); selCellRowText.push(cpad);
    var cal = document.createElement("div"); cal.className = "rowline";
    cal.appendChild(lbl("align"));
    var cs2 = document.createElement("select");
    [["default", ""], ["left", "left"], ["center", "center"], ["right", "right"], ["justify", "justify"]].forEach(function(o) {
      var eo = document.createElement("option"); eo.value = o[1]; eo.textContent = o[0];
      if ((cellMap["text-align"] || "") === o[1]) eo.selected = true;
      cs2.appendChild(eo);
    });
    cs2.onchange = function(){ cellExtraWrite(n, selCell.r, selCell.c, "text-align", cs2.value, "cell align"); };
    cal.appendChild(cs2); selCellRowText.push(cal);
    // column styling (apply to all cells in this column)
    var colBg = n.cellStyles[0][selCell.c] ? n.cellStyles[0][selCell.c].bg : "";
    var c3 = colorPicker("bg", colBg, function(v){ snap(); for (var rr=0; rr<n.rows; rr++) n.cellStyles[rr][selCell.c].bg = v; setOut(); autoPreview(); renderInspector(); });
    var cr3 = document.createElement("div"); cr3.className = "rowline"; cr3.appendChild(lbl("column: background")); cr3.appendChild(c3); selCellRowText.push(cr3);
    // Merge belongs to the cell context: one selection is enough — the buttons
    // join the selected cell with its neighbour to the right/below (chainable).
    var mrow = document.createElement("div");
    mrow.style.cssText = "display:flex;gap:4px;flex-wrap:wrap;margin-top:4px";
    [["Merge \u2192", function(){ tableMergeRight(n); }],
     ["Merge \u2193", function(){ tableMergeDown(n); }],
     ["Unmerge", function(){ tableUnmerge(n); }]
    ].forEach(function(bt){
      var mb2 = document.createElement("button"); mb2.className = "btn"; mb2.style.fontSize = "10.5px"; mb2.textContent = bt[0]; mb2.onclick = bt[1]; mrow.appendChild(mb2);
    });
    selCellRowText.push(mrow);
    var mhint = document.createElement("div"); mhint.className = "hint";
    mhint.textContent = "Merge \u2192 / \u2193 joins this cell with the neighbour on the right or below (one selection is enough; repeat for more cells).";
    selCellRowText.push(mhint);
    b.appendChild(group("Selected cell: r" + (selCell.r + 1) + " c" + (selCell.c + 1), selCellRowText, [], true));
  }
  var selRowRows = [];
  if (selRow !== null && selRow !== undefined) {
    var headerRow = selRow < n.header;
    var rs = n.rowStyles[selRow];
    var bgVal = headerRow ? (n.cellStyles[selRow][0] ? n.cellStyles[selRow][0].bg : "") : rs.bg;
    var colorVal = headerRow ? (n.cellStyles[selRow][0] ? n.cellStyles[selRow][0].color : "") : rs.color;
    var c4 = colorPicker("bg", bgVal, function(v){ snap(); if (headerRow) { for (var cc = 0; cc < n.cols; cc++) { var s = n.spans[selRow][cc]; if (s && s.cs !== 0) n.cellStyles[selRow][cc].bg = v; } } else { rs.bg = v; } setOut(); autoPreview(); renderInspector(); });
    var c5 = colorPicker("color", colorVal, function(v){ snap(); if (headerRow) { for (var cc = 0; cc < n.cols; cc++) { var s = n.spans[selRow][cc]; if (s && s.cs !== 0) n.cellStyles[selRow][cc].color = v; } } else { rs.color = v; } setOut(); autoPreview(); renderInspector(); });
    var cr4 = document.createElement("div"); cr4.className = "rowline"; cr4.appendChild(lbl("row background")); cr4.appendChild(c4); selRowRows.push(cr4);
    var cr5 = document.createElement("div"); cr5.className = "rowline"; cr5.appendChild(lbl("row text color")); cr5.appendChild(c5); selRowRows.push(cr5);
    b.appendChild(group("Selected row: " + (selRow + 1) + (headerRow ? " (header)" : ""), selRowRows, [], true));
  }
  var selColRows = [];
  if (selCol !== null && selCol !== undefined) {
    var colBg = n.cellStyles[0][selCol] ? n.cellStyles[0][selCol].bg : "";
    var c6 = colorPicker("bg", colBg, function(v){ snap(); for (var rr=0; rr<n.rows; rr++) { var s = n.spans[rr][selCol]; if (s && s.cs !== 0) n.cellStyles[rr][selCol].bg = v; } setOut(); autoPreview(); renderInspector(); });
    var c7 = colorPicker("color", n.cellStyles[0][selCol] ? n.cellStyles[0][selCol].color : "", function(v){ snap(); for (var rr=0; rr<n.rows; rr++) { var s = n.spans[rr][selCol]; if (s && s.cs !== 0) n.cellStyles[rr][selCol].color = v; } setOut(); autoPreview(); renderInspector(); });
    var cr6 = document.createElement("div"); cr6.className = "rowline"; cr6.appendChild(lbl("background")); cr6.appendChild(c6); selColRows.push(cr6);
    var cr7 = document.createElement("div"); cr7.className = "rowline"; cr7.appendChild(lbl("text color")); cr7.appendChild(c7); selColRows.push(cr7);
    // column width + alignment (materialize on the column's cells)
    var ccw = document.createElement("div"); ccw.className = "rowline";
    ccw.appendChild(lbl("width"));
    var cw = document.createElement("input"); cw.value = colExtraVal(n, selCol, "width"); cw.spellcheck = false; cw.placeholder = "120px / 25%";
    cw.onchange = function(){ colExtraWrite(n, selCol, "width", cw.value.trim(), "column width"); };
    ccw.appendChild(cw); selColRows.push(ccw);
    var cca = document.createElement("div"); cca.className = "rowline";
    cca.appendChild(lbl("align"));
    var cas = document.createElement("select");
    [["default", ""], ["left", "left"], ["center", "center"], ["right", "right"]].forEach(function(o) {
      var eo = document.createElement("option"); eo.value = o[1]; eo.textContent = o[0];
      if (colExtraVal(n, selCol, "text-align") === o[1]) eo.selected = true;
      cas.appendChild(eo);
    });
    cas.onchange = function(){ colExtraWrite(n, selCol, "text-align", cas.value, "column align"); };
    cca.appendChild(cas); selColRows.push(cca);
    b.appendChild(group("Selected column: " + String.fromCharCode(65 + selCol), selColRows, [], true));
  }

  a.appendChild(b);
  return a;
}
function tableAddCol(n) {
  withUndo(function(){
    n.cols++;
    for (var r = 0; r < n.rows; r++) { n.cells[r].push(""); n.spans[r].push({ cs: 1, rs: 1 }); n.cellStyles[r].push({ bg: "", color: "", extra: "" }); }
  });
}
function tableDelCol(n) {
  if (n.cols <= 1) return;
  withUndo(function(){
    n.cols--;
    for (var r = 0; r < n.rows; r++) { n.cells[r].pop(); n.spans[r].pop(); n.cellStyles[r].pop(); }
    clampSpans(n);
    if (selCell && selCell.c >= n.cols) selCell = null;
  });
}
function tableAddRow(n) {
  withUndo(function(){
    n.rows++;
    var cr = [], sr = [], csr = [];
    for (var c = 0; c < n.cols; c++) { cr.push(""); sr.push({ cs: 1, rs: 1 }); csr.push({ bg: "", color: "", extra: "" }); }
    n.cells.push(cr); n.spans.push(sr); n.cellStyles.push(csr); n.rowStyles.push({ bg: "", color: "", extra: "" });
  });
}
function tableDelRow(n) {
  if (n.rows <= n.header) return;
  withUndo(function(){
    n.rows--;
    n.cells.pop(); n.spans.pop(); n.cellStyles.pop(); n.rowStyles.pop();
    clampSpans(n);
    if (selCell && selCell.r >= n.rows) selCell = null;
    if (selRow !== null && selRow >= n.rows) selRow = null;
  });
}
function clampSpans(n) {
  for (var r = 0; r < n.rows; r++) for (var c = 0; c < n.cols; c++) {
    var s = n.spans[r][c];
    if (s.cs === 0) continue;
    if (c + s.cs > n.cols) s.cs = n.cols - c;
    if (r + s.rs > n.rows) s.rs = n.rows - r;
  }
}
function tableMergeRight(n) {
  if (!selCell) return;
  var r = selCell.r, c = selCell.c;
  var s = n.spans[r][c];
  if (!s || s.cs === 0 || s.rs !== 1) return;
  var nc = c + s.cs;                 // next cell after the current span
  if (nc >= n.cols) return;
  var t = n.spans[r][nc];
  if (!t || t.cs !== 1 || t.rs !== 1) return;
  withUndo(function(){
    s.cs += 1;                       // chainable: merge 2, then 3, ...
    n.cells[r][c] = combineCell(n.cells[r][c], n.cells[r][nc]);
    n.cells[r][nc] = "";
    t.cs = 0; t.rs = 0;
  });
}
function tableMergeDown(n) {
  if (!selCell) return;
  var r = selCell.r, c = selCell.c;
  var s = n.spans[r][c];
  if (!s || s.cs !== 1 || s.rs === 0) return;
  var nr = r + s.rs;                 // next row after the current span
  if (nr >= n.rows) return;
  var t = n.spans[nr][c];
  if (!t || t.cs !== 1 || t.rs !== 1) return;
  withUndo(function(){
    s.rs += 1;                       // chainable: merge 2, then 3, ...
    n.cells[r][c] = combineCell(n.cells[r][c], n.cells[nr][c]);
    n.cells[nr][c] = "";
    t.cs = 0; t.rs = 0;
  });
}
function tableUnmerge(n) {
  if (!selCell) return;
  var r = selCell.r, c = selCell.c;
  var s = n.spans[r][c];
  if (!s || s.cs === 0) return;
  withUndo(function(){
    if (s.cs > 1) { for (var cc = c + 1; cc < c + s.cs; cc++) { n.spans[r][cc] = { cs: 1, rs: 1 }; n.cells[r][cc] = ""; } s.cs = 1; }
    if (s.rs > 1) { for (var rr = r + 1; rr < r + s.rs; rr++) { n.spans[rr][c] = { cs: 1, rs: 1 }; n.cells[rr][c] = ""; } s.rs = 1; }
  });
}
function combineCell(a, b) { if (a && b) return a + " " + b; return a || b; }
// tableSetHeader switches the header treatment: "0" removes the header (the
// first row becomes body), "1" makes the first existing row the header, "new"
// inserts a brand-new header row at the top.
function tableSetHeader(n, mode) {
  withUndo(function(){
    if (mode === "0") { n.header = 0; return; }
    if (mode === "1") { n.header = Math.min(1, n.rows); return; }
    if (mode === "new") {
      n.rows += 1;
      var cr = [], sr = [], csr = [];
      for (var c = 0; c < n.cols; c++) { cr.push(""); sr.push({ cs: 1, rs: 1 }); csr.push({ bg: "", color: "", extra: "" }); }
      n.cells.unshift(cr); n.spans.unshift(sr); n.cellStyles.unshift(csr);
      n.rowStyles.unshift({ bg: "", color: "", extra: "" });
      n.header = 1;
      clampSpans(n);
      if (selCell) selCell = { r: selCell.r + 1, c: selCell.c };
      if (selRow !== null) selRow = selRow + 1;
    }
  });
}
function listControls(n) {
  var a = document.createElement("div"); a.className = "acc open";
  var h = document.createElement("div"); h.className = "acchead";
  h.innerHTML = '<span>List</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div"); b.className = "accbody";

  var ty = document.createElement("div"); ty.className = "rowline";
  ty.appendChild(lbl("type"));
  var ts = document.createElement("select");
  [["Unordered", false], ["Ordered", true]].forEach(function(o){
    var e = document.createElement("option"); e.value = o[1]; e.textContent = o[0]; if (n.ordered === o[1]) e.selected = true; ts.appendChild(e);
  });
  ts.onchange = function(){ withUndo(function(){ n.ordered = ts.value === "true"; }); };
  ty.appendChild(ts); b.appendChild(ty);

  var mk = document.createElement("div"); mk.className = "rowline";
  mk.appendChild(lbl("marker"));
  var ms = document.createElement("select");
  [["default", ""], ["disc", "disc"], ["circle", "circle"], ["square", "square"], ["none", "none"],
   ["decimal", "decimal"], ["decimal-leading-zero", "decimal-leading-zero"], ["lower-alpha", "lower-alpha"],
   ["upper-alpha", "upper-alpha"], ["lower-roman", "lower-roman"], ["upper-roman", "upper-roman"]].forEach(function(o){
    var e = document.createElement("option"); e.value = o[1]; e.textContent = o[0]; if ((n.marker || "") === o[1]) e.selected = true; ms.appendChild(e);
  });
  ms.onchange = function(){ withUndo(function(){ n.marker = ms.value; }); };
  mk.appendChild(ms); b.appendChild(mk);

  var il = document.createElement("div"); il.className = "rowth"; il.textContent = "Items"; b.appendChild(il);
  n.items.forEach(function(it, idx) {
    var row = document.createElement("div"); row.className = "rowline";
    var inp = document.createElement("input");
    inp.value = it || "";
    inp.oninput = function(){ snap(); n.items[idx] = inp.value; setOut(); autoPreview(); };
    row.appendChild(inp);
    row.appendChild(mkSmall("\u2191", idx === 0, function(){ withUndo(function(){ var t = n.items[idx]; n.items[idx] = n.items[idx - 1]; n.items[idx - 1] = t; }); }));
    row.appendChild(mkSmall("\u2193", idx === n.items.length - 1, function(){ withUndo(function(){ var t = n.items[idx]; n.items[idx] = n.items[idx + 1]; n.items[idx + 1] = t; }); }));
    row.appendChild(mkSmall("\u2212", false, function(){ withUndo(function(){ n.items.splice(idx, 1); }); }));
    b.appendChild(row);
  });
  var add = document.createElement("button"); add.className = "btn"; add.style.cssText = "font-size:11px;margin-top:4px"; add.textContent = "+ Item";
  add.onclick = function(){ withUndo(function(){ n.items.push("Item"); }); };
  b.appendChild(add);

  a.appendChild(b);
  return a;
}
function mkSmall(label, disabled, fn) {
  var x = document.createElement("button"); x.className = "btn"; x.style.cssText = "padding:2px 6px;font-size:10px;flex:none";
  x.textContent = label; x.disabled = disabled; x.onclick = fn;
  return x;
}

// ---- text input / text area parameters ------------------------------------
// inputControls edits the input/textarea-specific fields (placeholder, value,
// type, validation). With a multi-selection the shared value is shown and the
// edit applies to every selected element; input-only rows (type/options)
// appear only when every selected node is an input.
function requiredControls(nodes) {
  if (!nodes.length) return null;
  var allowed = { select: 1, number: 1, radio: 1, date: 1, time: 1 };
  if (!everyKind(nodes, function(k){ return !!allowed[k]; })) return null;
  var rr = document.createElement("div"); rr.className = "rowline";
  var lblReq = document.createElement("label"); lblReq.textContent = "required"; rr.appendChild(lblReq);
  var reqOn = nodes.every(function(m){ return !!m.required; });
  var reqMixed = nodes.length > 1 && !reqOn && nodes.some(function(m){ return !!m.required; });
  rr.appendChild(mkSwitch(reqOn, reqMixed, function(on){ upd("required", on); }));
  var hint = document.createElement("div"); hint.className = "hint";
  hint.textContent = "Native required validation for the selected control.";
  return acc("Validation", [rr, hint]);
}

function inputControls(nodes) {
  var allInput = everyKind(nodes, function(k){ return k === "input"; });
  var a = document.createElement("div"); a.className = "acc open";
  var h = document.createElement("div"); h.className = "acchead";
  h.innerHTML = '<span>' + (allInput ? "Text input" : "Text area") + '</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div"); b.className = "accbody";

  b.appendChild(rowInput("placeholder", commonVal(nodes, "placeholder"), function(v){ upd("placeholder", v); }));
  b.appendChild(rowInput("value", commonVal(nodes, "value"), function(v){ upd("value", v); }));

  if (allInput) {
    var tr = document.createElement("div"); tr.className = "rowline";
    tr.appendChild(lbl("type"));
    var s = document.createElement("select");
    ["text", "email", "number", "password", "tel", "url", "date", "search"].forEach(function(t){
      var o = document.createElement("option"); o.value = t; o.textContent = t; if ((commonVal(nodes, "inputtype") || "text") === t) o.selected = true; s.appendChild(o);
    });
    s.onchange = function(){ upd("inputtype", s.value); };
    tr.appendChild(s); b.appendChild(tr);
  }

  b.appendChild(rowInput("pattern", commonVal(nodes, "pattern"), function(v){ upd("pattern", v); }));
  b.appendChild(rowInput("minlength", commonVal(nodes, "minlength"), function(v){ upd("minlength", v); }));
  b.appendChild(rowInput("maxlength", commonVal(nodes, "maxlength"), function(v){ upd("maxlength", v); }));

  var rr = document.createElement("div"); rr.className = "rowline";
  rr.appendChild(lbl("required"));
  var reqOn = nodes.length > 0 && nodes.every(function(m){ return !!m.required; });
  var reqMixed = nodes.length > 1 && !reqOn && nodes.some(function(m){ return !!m.required; });
  var rsw = mkSwitch(reqOn, reqMixed, function(on){ upd("required", on); });
  rr.appendChild(rsw); b.appendChild(rr);

  if (allInput) {
    var or = document.createElement("div"); or.className = "rowline";
    or.appendChild(lbl("options"));
    var oi = document.createElement("input"); oi.value = commonVal(nodes, "options"); oi.placeholder = "a, b, c";
    oi.oninput = function(){ upd("options", oi.value); };
    or.appendChild(oi); b.appendChild(or);
    var hint = document.createElement("div"); hint.className = "hint";
    hint.textContent = "comma-separated — enables autocomplete";
    b.appendChild(hint);
  }

  a.appendChild(b);
  return a;
}
function rowInput(label, val, onchange, placeholder) {
  var d = document.createElement("div"); d.className = "rowline";
  d.appendChild(lbl(label));
  var i = document.createElement("input"); i.value = val || "";
  if (placeholder) i.placeholder = placeholder;
  i.oninput = function(){ onchange(i.value); };
  d.appendChild(i);
  return d;
}

// ---- rich text (WYSIWYG) editor -------------------------------------------
function styleToObj(s) {
  var o = {};
  String(s || "").split(";").forEach(function(part) {
    var kv = part.split(":");
    if (kv.length < 2) return;
    var k = kv[0].trim(); var v = kv.slice(1).join(":").trim();
    if (k && v) o[k] = v;
  });
  return o;
}
function runsToHtml(runs) {
  var out = "";
  (runs || []).forEach(function(r) {
    var txt = (r.text || "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    if (txt === "") return;
    var style = "";
    if (r.color) style += "color:" + r.color + ";";
    if (r.bg) style += "background:" + r.bg + ";";
    if (r.font) style += "font-family:" + r.font + ";";
    var s = style ? '<span style="' + style + '">' + txt + '</span>' : txt;
    if (r.u) s = '<u>' + s + '</u>';
    if (r.i) s = '<i>' + s + '</i>';
    if (r.b) s = '<b>' + s + '</b>';
    out += s;
  });
  return out;
}
function htmlToRuns(html) {
  var div = document.createElement("div");
  div.innerHTML = html || "";
  var runs = [];
  var cur = null, curF = null;
  function key(f) { return (f.b?1:0) + "|" + (f.i?1:0) + "|" + (f.u?1:0) + "|" + f.color + "|" + f.bg + "|" + f.font; }
  function append(text, f) {
    if (text === "") return;
    if (cur && key(curF) === key(f)) { cur.text += text; }
    else { cur = { text: text, b: f.b, i: f.i, u: f.u, color: f.color, bg: f.bg, font: f.font }; runs.push(cur); curF = f; }
  }
  (function visit(el, f) {
    for (var i = 0; i < el.childNodes.length; i++) {
      var node = el.childNodes[i];
      if (node.nodeType === 3) { append(node.nodeValue, f); continue; }
      if (node.nodeType !== 1) continue;
      var tag = node.tagName.toLowerCase();
      if (tag === "br") { append("\n", f); continue; }
      var nf = { b: f.b || tag === "b" || tag === "strong", i: f.i || tag === "i" || tag === "em", u: f.u || tag === "u", color: f.color, bg: f.bg, font: f.font };
      var st = node.getAttribute && node.getAttribute("style");
      if (st) {
        var p = styleToObj(st);
        if (p.color) nf.color = p.color;
        if (p.background || p["background-color"]) nf.bg = p.background || p["background-color"];
        if (p["font-family"]) nf.font = p["font-family"];
        if (p["font-weight"] === "bold" || parseInt(p["font-weight"], 10) >= 600) nf.b = true;
        if (p["font-style"] === "italic") nf.i = true;
        var dec = p["text-decoration"] || p["text-decoration-line"] || "";
        if (dec.indexOf("underline") >= 0) nf.u = true;
      }
      if (tag === "font") {
        if (node.getAttribute("color")) nf.color = node.getAttribute("color");
        if (node.getAttribute("face")) nf.font = node.getAttribute("face");
      }
      visit(node, nf);
    }
  })(div, { b: false, i: false, u: false, color: "", bg: "", font: "" });
  return runs;
}
function textEditor(n) {
  var a = document.createElement("div"); a.className = "acc open";
  var h = document.createElement("div"); h.className = "acchead";
  h.innerHTML = '<span>Text</span><span>\u25BE</span>';
  h.onclick = function(){ a.classList.toggle("open"); };
  a.appendChild(h);
  var b = document.createElement("div"); b.className = "accbody";

  var r1 = document.createElement("div"); r1.className = "rowline";
  r1.appendChild(lbl("block"));
  var ts = document.createElement("select");
  ["p", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "pre"].forEach(function(t){
    var o = document.createElement("option"); o.value = t; o.textContent = t; if (n.tag === t) o.selected = true; ts.appendChild(o);
  });
  ts.onchange = function(){ withUndo(function(){ n.tag = ts.value; }); };
  r1.appendChild(ts); b.appendChild(r1);

  var r2 = document.createElement("div"); r2.className = "rowline";
  r2.appendChild(lbl("align"));
  var asel = document.createElement("select");
  [["default", ""], ["left", "left"], ["center", "center"], ["right", "right"], ["justify", "justify"]].forEach(function(t){
    var o = document.createElement("option"); o.value = t[1]; o.textContent = t[0]; if (n.align === t[1]) o.selected = true; asel.appendChild(o);
  });
  asel.onchange = function(){ withUndo(function(){ n.align = asel.value; }); };
  r2.appendChild(asel); b.appendChild(r2);

  var r3 = document.createElement("div"); r3.className = "rowline";
  r3.appendChild(lbl("size"));
  var sz = document.createElement("input"); sz.style.width = "48px"; sz.value = n.size || "";
  sz.placeholder = "px";
  function setSize(v) { withUndo(function(){ n.size = v; }); }
  sz.onchange = function(){ setSize(sz.value); };
  r3.appendChild(sz);
  r3.appendChild(mkSmall("\u2212", false, function(){ setSize(Math.max(6, (parseInt(n.size, 10) || 16) - 1) + "px"); }));
  r3.appendChild(mkSmall("+", false, function(){ setSize(((parseInt(n.size, 10) || 16) + 1) + "px"); }));
  b.appendChild(r3);

  // toolbar
  var tb = document.createElement("div"); tb.style.cssText = "display:flex;gap:4px;flex-wrap:wrap;align-items:center;margin:6px 0";
  function tbBtn(label, fn, title) {
    var x = document.createElement("button"); x.className = "btn"; x.style.cssText = "font-size:12px;padding:3px 8px";
    x.innerHTML = label; x.title = title; x.onmousedown = function(ev){ ev.preventDefault(); }; x.onclick = fn; tb.appendChild(x);
    return x;
  }
  function mkColor(title, fg) {
    var c = document.createElement("input"); c.type = "color"; c.value = "#ffffff"; c.title = title;
    c.style.cssText = "width:26px;height:24px;padding:0;border:1px solid var(--border2);border-radius:4px;background:none";
    c.oninput = function(){ restore(); document.execCommand(fg ? "foreColor" : "hiliteColor", false, c.value); commit(); };
    tb.appendChild(c);
  }
  tbBtn("<b>B</b>", function(){ restore(); document.execCommand("bold"); commit(); }, "Bold");
  tbBtn("<i>I</i>", function(){ restore(); document.execCommand("italic"); commit(); }, "Italic");
  tbBtn("<u>U</u>", function(){ restore(); document.execCommand("underline"); commit(); }, "Underline");
  mkColor("Text color", true);
  mkColor("Highlight", false);
  var fsel = document.createElement("select"); fsel.style.cssText = "max-width:110px";
  var fo = document.createElement("option"); fo.value = ""; fo.textContent = "font"; fsel.appendChild(fo);
  ["Arial", "Georgia", "Courier New", "Times New Roman", "Verdana", "Tahoma", "Impact", "Comic Sans MS"].concat(uploadedFonts).forEach(function(f){
    var o = document.createElement("option"); o.value = f; o.textContent = f; fsel.appendChild(o);
  });
  fsel.onchange = function(){ if (fsel.value) { restore(); document.execCommand("fontName", false, fsel.value); commit(); } };
  tb.appendChild(fsel);

  var ed = document.createElement("div"); ed.className = "richtext"; ed.contentEditable = "true";
  ed.innerHTML = runsToHtml(n.runs);
  ed.style.cssText = "min-height:64px;border:1px solid var(--border2);border-radius:6px;padding:8px;font-size:" + (n.size || "14px") + ";line-height:1.5;outline:none;background:var(--input-bg);color:var(--fg);white-space:pre-wrap";
  b.appendChild(ed);
  a.appendChild(b);

  var timer = null;
  var savedRange = null;
  function save() { var s = window.getSelection(); if (s.rangeCount > 0) savedRange = s.getRangeAt(0).cloneRange(); }
  function restore() {
    ed.focus();
    if (savedRange) { var s = window.getSelection(); s.removeAllRanges(); s.addRange(savedRange); }
  }
  function commit() {
    clearTimeout(timer);
    timer = setTimeout(function(){ _commit(); }, 150);
  }
  function _commit() {
    n.runs = htmlToRuns(ed.innerHTML);
    if (!n.runs.length) n.runs = [{ text: "", b: false, i: false, u: false, color: "", bg: "", font: "" }];
    setOut(); autoPreview(); renderTree();
  }
  ed.oninput = commit;
  ed.onkeyup = function(){ save(); };
  ed.onmouseup = function(){ save(); };
  ed.addEventListener("blur", function(){ save(); _commit(); });

  return a;
}
function renderProblems() {
  var ul = document.getElementById("problems"); ul.innerHTML = "";
  var items = [];
  // SUI diagnostics (compiler) first so they can be clicked to jump.
  (__suiDiags || []).forEach(function(d){
    items.push({ kind: (d.kind === "warning" ? "warn" : "error"), line: d.line, message: d.msg, col: d.col });
  });
  problems.forEach(function(p){
    items.push({ kind: p.kind, line: p.line, message: p.message, css: true });
  });
  if (!items.length) { var li = document.createElement("li"); li.textContent = "No problems."; ul.appendChild(li); }
  items.forEach(function(p){
    var li = document.createElement("li");
    li.className = p.kind;
    li.textContent = (p.line ? "L" + p.line + ": " : "") + p.message;
    if (!p.css && p.line) {
      li.style.cursor = "pointer";
      li.title = "Go to line " + p.line;
      li.onclick = function(){ revealLine(p.line - 1, Math.max(0, (p.col || 1) - 1)); };
    }
    ul.appendChild(li);
  });
  var n = items.length;
  var pb = document.getElementById("probbtn");
  if (pb) pb.textContent = (document.getElementById("bottom").classList.contains("on") ? "Hide" : "Problems") + " panel" + (n ? " (" + n + ")" : "");
}
function renderTokens() {
  var box = document.getElementById("tokens"); box.innerHTML = "";
  tokens.forEach(function(t){
    var d = document.createElement("div"); d.className = "tok";
    var sw = document.createElement("span"); sw.className = "sw"; sw.style.background = isColor(t.value) ? t.value : "#2a2f3d";
    d.appendChild(sw);
    var nm = document.createElement("span"); nm.className = "tkname"; nm.textContent = t.name;
    d.appendChild(nm);
    var vl = document.createElement("span"); vl.className = "tkval"; vl.textContent = t.value;
    d.appendChild(vl);
    box.appendChild(d);
  });
}
function isColor(v) { return /^#([0-9a-fA-F]{3,8})$/.test(v) || /^rgb/.test(v) || /^hsl/.test(v); }
function addToken() {
  var name = prompt("Token name", "--primary");
  var value = prompt("Token value", "#4c5cf7");
  if (!name || !value) return;
  if (name.indexOf("--") !== 0) name = "--" + name;
  snap();
  var prefix = ":root { " + name + ": " + value + "; }\n";
  if (root.css.indexOf(":root") >= 0) {
    root.css = root.css.replace(/:root\s*\{/, ":root { " + name + ": " + value + "; ");
  } else {
    root.css = prefix + root.css;
  }
  setOut(); autoPreview(); scheduleCssSync();
}
function renderHistory() {
  var h = document.getElementById("history"); if (!h) return;
  h.innerHTML = "";
  var title = document.createElement("div"); title.className = "hint"; title.textContent = "Click a step to revert";
  h.appendChild(title);
  if (!past.length) { updateToolbar(); return; }
  for (var i = past.length - 1; i >= 0; i--) {
    (function(idx){
      var d = document.createElement("div");
      d.className = "histitem";
      d.textContent = (pastLabels[idx] || "change") + (idx === past.length - 1 ? "  \u25cf" : "");
      d.onclick = function(){ jumpTo(idx); };
      h.appendChild(d);
    })(i);
  }
  updateToolbar();
}
function updateToolbar() {
  var u = document.getElementById("tb-undo");
  var r = document.getElementById("tb-redo");
  var d = document.getElementById("tb-dup");
  if (u) u.disabled = !past.length;
  if (r) r.disabled = !future.length;
  if (d) d.disabled = !selPath;
}
function openPreviewWindow() {
  var src = document.getElementById("out").value || suiSource || generate();
  setStatus("building…");
  var win = window.open("about:blank", "_blank");
  var fd = new URLSearchParams(); fd.set("sui", src);
  fetch("/__gui__/build", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) { if (win) win.close(); setStatus((j.diagnostics || []).join(" | ")); return; }
    if (win) { win.location.href = "/__gui__/built/"; }
    else { window.open("/__gui__/built/", "_blank"); }
    setStatus("preview opened");
  }).catch(function(){ if (win) win.close(); setStatus("build failed"); });
}
function renderGutters() {
  highlightSui();
}
function lineNumbers(text) {
  var n = text.split("\n").length, out = "";
  for (var i = 1; i <= n; i++) out += i + "\n";
  return out;
}
function editorLineHeightPx() {
  var ta = document.getElementById("out");
  if (!ta) return 21.7;
  var lh = parseFloat(getComputedStyle(ta).lineHeight);
  if (isFinite(lh) && lh > 0) return lh;
  var fs = parseFloat(getComputedStyle(ta).fontSize) || 16;
  return fs * 1.6;
}
function editorCharWidthPx() {
  var ta = document.getElementById("out");
  if (!ta) return 8.4;
  var cs = getComputedStyle(ta);
  var probe = document.createElement("span");
  probe.textContent = "MMMMMMMMMM";
  probe.style.cssText = "position:fixed;left:-10000px;top:0;visibility:hidden;white-space:pre;";
  ["fontFamily","fontSize","fontWeight","fontStyle","letterSpacing","tabSize"].forEach(function(k){ probe.style[k] = cs[k]; });
  document.body.appendChild(probe);
  var w = probe.getBoundingClientRect().width / 10;
  probe.remove();
  return w || 8.4;
}
function setStatus(s) { document.getElementById("status").textContent = s; }
function dbg(s) {
  var d = document.getElementById("dbg");
  if (d) d.textContent = s;
  if (window.console) console.log("[dbg]", s);
}
function setSuiEditor() {
  var v = suiSource || generate();
  // External replacement (designer edit, apply, load): the editor history
  // restarts so undo never rewinds across an unrelated regeneration.
  undoPrepareExternal(v);
  document.getElementById("out").value = v;
  highlightSui();
  renderOutline();
}
function highlightSui() {
  var s = document.getElementById("out").value;
  // Use the overlay renderer so inline diagnostics, bracket match and element
  // highlighting survive every keystroke; it falls back to plain highlighting
  // when no overlay is active.
  document.getElementById("sui_hl").innerHTML = supHighlight(s) + "\n";
  document.getElementById("suigutter").textContent = lineNumbers(s);
  updateBracketMatch();
}

// ---- snippets ------------------------------------------------------------
// Each snippet is a template inserted at the cursor. "$0" marks the final
// caret position; "${1}".."${n}" are tab stops visited in order while typing
// (Tab jumps to the next stop). Lines are indented to the current line.
var SNIPPETS = {
  fn: "fn ${1:name}(${2}) -> ${3:null} {\n\t$0\n}",
  fnasync: "async fn ${1:name}(${2}) -> ${3:null} {\n\t$0\n}",
  state: "state ${1:name} = ${2:0}",
  "if": "if ${1:condition} {\n\t$0\n} else {\n\t\n}",
  loop: "loop ${1:item} in ${2:items} {\n\t$0\n}",
  handler: "fn ${1:onClick}() -> null {\n\t$0\n}",
  component: "view ${1:Name}(${2}) {\n\t$0\n}",
  list: "list ${1:item} in ${2:items} {\n\t${3:row \"{item}\"}\n}"
};

// ---- snippet menu ---------------------------------------------------------
// The Snippets menu is rendered as a fixed-position overlay anchored to its
// button. Doing it this way (instead of an absolutely-positioned .drop child)
// means it can never be clipped by the toolbar's overflow:auto wrapper or by
// a short pane, and it does not depend on event propagation.
// Grouped, described entries drive the Snippets menu rendering. "grp" opens a
// small section header; "icon" is the monogram chip shown before the label.
var SNIPPET_MENU = [
  { grp: "Structure" },
  { id: "fn", label: "Function", icon: "fn", desc: "fn name(args) -> ret" },
  { id: "fnasync", label: "Async function", icon: "as", desc: "await http.get(...)" },
  { id: "state", label: "State", icon: "st", desc: "state name = init" },
  { id: "if", label: "If / else", icon: "if", desc: "conditional statement" },
  { id: "loop", label: "Loop", icon: "lp", desc: "loop item in items" },
  { id: "handler", label: "Event handler", icon: "ev", desc: "fn onClick() -> null" },
  { grp: "Views" },
  { id: "component", label: "Component view", icon: "vw", desc: "view Name(props)" },
  { id: "list", label: "List", icon: "ls", desc: "list item in items" }
];
function snippetMenuEl() {
  var m = document.getElementById("snip-menu");
  if (!m) {
    m = document.createElement("div");
    m.id = "snip-menu";
    m.className = "sui-ac";
    m.style.position = "fixed";
    m.style.minWidth = "270px";
    m.style.display = "none";
    document.body.appendChild(m);
  }
  return m;
}
function closeSnippetMenu() {
  var m = document.getElementById("snip-menu");
  if (m) m.style.display = "none";
}
// snippetDemo expands a snippet template into the code it will insert, so the
// preview strip at the bottom of the menu can show real code instead of
// placeholder markers: ${1:name} -> name, bare ${n} -> removed, $0 -> removed.
function snippetDemo(tpl) {
  return tpl.replace(/\$\{(\d+):([^}]*)\}/g, "$2").replace(/\$\{\d+\}/g, "").replace(/\$0/g, "");
}
// snipShowPreview fills the menu's preview strip with a highlighted rendering
// of the snippet with the given id.
function snipShowPreview(id) {
  var p = document.getElementById("snip-preview");
  if (!p) return;
  var tpl = SNIPPETS[id];
  p.innerHTML = tpl === undefined ? "" : highlight(snippetDemo(tpl), "sui");
}
// snipItems lists the selectable rows of the open snippet menu.
function snipItems() {
  return Array.prototype.slice.call(document.querySelectorAll("#snip-menu .snip-item"));
}
// snipSetKbd highlights item i as the keyboard's current selection and previews
// its template; negative i clears the highlight.
function snipSetKbd(i) {
  var its = snipItems();
  its.forEach(function(it, k) { it.classList.toggle("kbd", k === i); });
  if (i >= 0 && its[i]) {
    its[i].scrollIntoView({ block: "nearest" });
    snipShowPreview(its[i].getAttribute("data-snip"));
  }
}
// snipMove moves the keyboard selection by dir items (wrapping around).
function snipMove(dir) {
  var its = snipItems();
  if (!its.length) return;
  var idx = 0;
  for (var i = 0; i < its.length; i++) { if (its[i].classList.contains("kbd")) { idx = i; break; } }
  snipSetKbd((idx + dir + its.length) % its.length);
}
// snipPick inserts the snippet currently highlighted by the keyboard.
function snipPick() {
  var its = snipItems();
  if (!its.length) return;
  var pick = null;
  for (var i = 0; i < its.length; i++) { if (its[i].classList.contains("kbd")) { pick = its[i]; break; } }
  if (!pick) return;
  var id = pick.getAttribute("data-snip");
  closeSnippetMenu();
  insertSnippet(id);
}
// menuZoom returns the effective CSS zoom applied to the designer root. The
// whole UI is scaled via body zoom (applyTheme), and getBoundingClientRect()
// returns zoomed pixels while left/top are interpreted unzoomed, so overlay
// menus must divide measured coordinates by this factor to stay anchored.
function menuZoom() {
  var z = parseFloat(getComputedStyle(document.body).zoom);
  if (!z || isNaN(z) || z <= 0) z = 1;
  return z;
}
function toggleSnippetMenu(ev) {
  if (ev && ev.stopPropagation) ev.stopPropagation();
  if (ev && ev.preventDefault) ev.preventDefault();
  var m = snippetMenuEl();
  var open = m.style.display === "block";
  closeSnippetMenu();
  closeMenus();
  if (open) return;
  // Header, one section per group, and a live preview strip at the bottom.
  var head = document.createElement("div");
  head.className = "snip-head";
  var ht = document.createElement("span"); ht.className = "st"; ht.textContent = "Snippets";
  var hk = document.createElement("span"); hk.className = "grily"; hk.textContent = SNIPPET_MENU.filter(function(x){ return x.id; }).length + " modelli";
  head.appendChild(ht); head.appendChild(hk);
  m.appendChild(head);
  SNIPPET_MENU.forEach(function(it) {
    if (it.grp) { var g = document.createElement("div"); g.className = "snip-grp"; g.textContent = it.grp; m.appendChild(g); return; }
    var d = document.createElement("div");
    d.className = "snip-item";
    d.setAttribute("data-snip", it.id);
    var ic = document.createElement("span"); ic.className = "snip-ic"; ic.textContent = it.icon || "";
    var lb = document.createElement("span"); lb.className = "snip-lb"; lb.textContent = it.label;
    var ds = document.createElement("span"); ds.className = "snip-ds"; ds.textContent = it.desc || "";
    // Hover and keyboard selection are the same state: highlight + preview.
    d.onmouseenter = function() { snipSetKbd(snipItems().indexOf(d)); };
    // Use mousedown so the menu acts before the editor/textarea loses focus.
    d.onmousedown = function(e) { e.preventDefault(); };
    d.onclick = function(e) { e.preventDefault(); e.stopPropagation(); closeSnippetMenu(); insertSnippet(it.id); };
    d.appendChild(ic); d.appendChild(lb); d.appendChild(ds);
    m.appendChild(d);
  });
  var pv = document.createElement("div");
  pv.id = "snip-preview";
  m.appendChild(pv);
  var ftr = document.createElement("div");
  ftr.className = "snip-ftr";
  var k1 = document.createElement("span"); k1.className = "kkey"; k1.textContent = "\u2191\u2193";
  var t1 = document.createElement("span"); t1.textContent = "choose";
  var k2 = document.createElement("span"); k2.className = "kkey"; k2.textContent = "\u21b5";
  var t2 = document.createElement("span"); t2.textContent = "insert";
  var k3 = document.createElement("span"); k3.className = "kkey"; k3.textContent = "esc";
  var t3 = document.createElement("span"); t3.textContent = "close";
  [k1, t1, k2, t2, k3, t3].forEach(function(n) { ftr.appendChild(n); });
  m.appendChild(ftr);
  snipSetKbd(0);
  var btn = document.getElementById("snipbtn") || (ev && ev.currentTarget);
  var r = btn ? btn.getBoundingClientRect() : { left: 20, top: 60, bottom: 80 };
  m.style.display = "block";
  // getBoundingClientRect() reports *scaled* viewport pixels, but a fixed/absolute
  // left/top is interpreted in *unscaled* CSS pixels. The designer root applies a
  // zoom (body zoom), so divide the measured coordinates by the effective zoom
  // before assigning them, otherwise the menu lands offset from its button.
  var z = menuZoom();
  var mw = m.offsetWidth, mh = m.offsetHeight;
  var vw = window.innerWidth / z, vh = window.innerHeight / z;
  var left = Math.max(4, Math.min(vw - mw - 8, r.left / z));
  var top = Math.min(vh - mh - 8, r.bottom / z + 4);
  m.style.left = left + "px";
  m.style.top = Math.max(4, top) + "px";
  // Close on an outside pointer press or Escape. A short delay avoids the
  // very click that opened the menu from immediately closing it.
  setTimeout(function() {
    var h = function(e) {
      var mm = document.getElementById("snip-menu");
      if (!mm || mm.style.display !== "block") { document.removeEventListener("mousedown", h, true); return; }
      if (mm.contains(e.target)) return;
      closeSnippetMenu();
      document.removeEventListener("mousedown", h, true);
    };
    document.addEventListener("mousedown", h, true);
  }, 0);
}

// indentOf returns the leading whitespace of the line containing pos.
function indentOf(val, pos) {
  var ls = val.lastIndexOf("\n", pos - 1) + 1;
  var m = /^[ \t]*/.exec(val.slice(ls, pos));
  return m ? m[0] : "";
}

// insertSnippet expands a named snippet at the caret, applying the indentation
// of the current line to every embedded newline.
function insertSnippet(name) {
  var tpl = SNIPPETS[name];
  if (tpl === undefined) return;
  var ta = document.getElementById("out");
  var val = ta.value, start = ta.selectionStart, end = ta.selectionEnd;
  var indent = indentOf(val, start);
  var text = tpl.replace(/\n/g, "\n" + indent);
  // Strip placeholders into plain text, remembering the caret. A placeholder
  // is ${n:default} (default may be empty) or the bare ${n} / $0 form.
  var re = /\$\{(\d+)(?::([^}]*))?\}|\$0/g, m;
  var clean = "", last = 0, caret = -1, selEnd = -1;
  while ((m = re.exec(text)) !== null) {
    clean += text.slice(last, m.index);
    if (m[0] === "$0") {
      if (caret < 0) caret = start + clean.length;
    } else {
      var off = start + clean.length;
      var def = m[2] || "";
      clean += def;
      if (caret < 0) { caret = off; selEnd = off + def.length; }
    }
    last = re.lastIndex;
  }
  clean += text.slice(last);
  ta.value = val.slice(0, start) + clean + val.slice(end);
  ta.focus();
  if (selEnd >= 0) { ta.selectionStart = caret; ta.selectionEnd = selEnd; }
  else { ta.selectionStart = ta.selectionEnd = caret < 0 ? start + clean.length : caret; }
  onSuiEdit();
  closeSnippetMenu();
  setStatus("snippet: " + name);
}


// ---- source generation ---------------------------------------------------
function esc(s) {
  return s.replace(/\\/g, "\\\\").replace(/"/g, "\\\"").replace(/\n/g, "\\n").replace(/\{/g, "\\{").replace(/\}/g, "\\}");
}
// escText is like esc but leaves interpolation braces {expr} intact so
// reactive text such as "Likes: {likes}" round-trips without being escaped.
function escText(s) {
  if (s === undefined || s === null) return "";
  var out = "";
  var i = 0, n = s.length;
  while (i < n) {
    var c = s[i];
    if (c === '{') {
      var j = s.indexOf('}', i);
      if (j >= 0) { out += s.slice(i, j + 1); i = j + 1; continue; }
    }
    if (c === '\\') out += '\\\\';
    else if (c === '"') out += '\\"';
    else if (c === '\n') out += '\\n';
    else out += c;
    i++;
  }
  return out;
}
function bondOut(v) {
  var s = String(v && v.trim ? v.trim() : v);
  if (IDENT.test(s) && !RESERVED_NAME.test(s) && !/^-?\d*\.?\d+$/.test(s) && !hasState(s)) {
    return "\"" + esc(s) + "\"";
  }
  return s;
}
function cssSafeExpr(expr) {
  var s = String(expr || "").trim();
  if (/^cssSafe\s*\(/.test(s)) return s;
  return "cssSafe(" + s + ")";
}
function chartParamAttr(n, key, field, value, def, quote) {
  var ex = n.chartExprs && n.chartExprs[field] ? String(n.chartExprs[field]) : "";
  if (ex) return " " + key + ": " + ex;
  if (value === undefined || value === null || String(value) === "" || String(value) === String(def)) return "";
  return " " + key + ": " + (quote ? "\"" + esc(String(value)) + "\"" : bondOut(String(value)));
}
function chartBoolAttr(n, key, field, value, def) {
  var ex = n.chartExprs && n.chartExprs[field] ? String(n.chartExprs[field]) : "";
  if (ex) return " " + key + ": " + ex;
  if (def ? value === false : value === true) return " " + key + ": " + (value ? "true" : "false");
  return "";
}
function nameOut(v) {
  var s = String(v && v.trim ? v.trim() : v);
  if (IDENT.test(s) && !RESERVED_NAME.test(s) && hasState(s)) return s;
  return "\"" + esc(s) + "\"";
}
function attrsOf(n) {
  var a = "";
  if (n.placeholder) a += " placeholder: \"" + esc(n.placeholder) + "\"";
  if (n.value) a += " value: " + n.value;
  // Control bindings are raw expressions (a state name, a comparison, a
  // bool...). They are emitted unquoted exactly like value/onClick so the
  // compiler re-parses them as expressions, not as strings. A plain string
  // the compiler cannot resolve as a declared state (like a radio group
  // name "grp1") must be quoted: bare it parses as an unknown variable and
  // the render fails.
  if (n.checked) a += " checked: " + n.checked;
  if (n.disabled) a += " disabled: " + bondOut(n.disabled);
  if (n.name) a += " name: " + nameOut(n.name);
  if (n.variant) a += " variant: \"" + esc(n.variant) + "\"";
  if (n.close === false) a += " close: false";
  if (n.min) a += " min: " + (n.kind === "datepicker" ? "\"" + esc(n.min) + "\"" : bondOut(n.min));
  if (n.max) a += " max: " + (n.kind === "datepicker" ? "\"" + esc(n.max) + "\"" : bondOut(n.max));
  if (n.step) a += " step: " + bondOut(n.step);
  if (n.kind === "datepicker" && n.mode) a += " mode: \"" + esc(n.mode) + "\"";
  if (n.accept) a += " accept: \"" + esc(n.accept) + "\"";
  if (n.multiple) a += " multiple: true";
  if (n.src) a += " src: " + (n.srcExpr ? n.src : "\"" + esc(n.src) + "\"");
  if (n.link && n.linkExpr) a += " link: " + n.link;
  else if (n.link) a += " link: \"" + esc(n.link) + "\"";
  if (n.inputtype && n.kind === "input") a += " inputtype: \"" + esc(n.inputtype) + "\"";
  if (n.pattern) a += " pattern: \"" + esc(n.pattern) + "\"";
  if (n.minlength) a += " minlength: \"" + esc(String(n.minlength)) + "\"";
  if (n.maxlength) a += " maxlength: \"" + esc(String(n.maxlength)) + "\"";
  if (n.required) a += " required: true";
  if (n.options) a += n.optionsExpr ? " options: " + n.options : " options: \"" + esc(n.options) + "\"";
  if (n.maxOptionLength) a += " maxOptionLength: " + n.maxOptionLength;
  if (n.maxOptions) a += " maxOptions: " + n.maxOptions;
  if (isChartKind(n)) {
    if (n.chartData) a += " data: " + (n.chartDataExpr ? n.chartData : "\"" + esc(n.chartData) + "\"");
    if (n.kind !== "graph") a += chartParamAttr(n, "chartType", "chartType", n.chartType, "bar", true);
    a += chartParamAttr(n, "theme", "chartTheme", n.chartTheme, "dark", true);
    a += chartParamAttr(n, "format", "chartFormat", n.chartFormat, "auto", true);
    a += chartParamAttr(n, "x", "chartX", n.chartX, "", true);
    a += chartParamAttr(n, "y", "chartY", n.chartY, "", true);
    a += chartParamAttr(n, "width", "chartWidth", n.chartWidth, "640px", true);
    a += chartParamAttr(n, "height", "chartHeight", n.chartHeight, "320px", true);
    a += chartParamAttr(n, "overflow", "chartOverflow", n.chartOverflow, "scroll", true);
    a += chartParamAttr(n, "zoom", "chartZoom", n.chartZoom, "1", false);
    a += chartParamAttr(n, "pointWidth", "chartPointWidth", n.chartPointWidth, "48", false);
    a += chartParamAttr(n, "pointSize", "chartPointSize", n.chartPointSize, "3", false);
    a += chartParamAttr(n, "tableView", "chartTableView", n.chartTableView, "chart", true);
    a += chartBoolAttr(n, "tableFilters", "chartTableFilters", n.chartTableFilters, true);
    a += chartBoolAttr(n, "tableAddRow", "chartTableAddRow", n.chartTableAddRow, false);
    a += chartBoolAttr(n, "tableDeleteRow", "chartTableDeleteRow", n.chartTableDeleteRow, false);
    a += chartBoolAttr(n, "tableSteppers", "chartTableSteppers", n.chartTableSteppers, false);
    a += chartBoolAttr(n, "tableContextMenu", "chartTableContextMenu", n.chartTableContextMenu, false);
    a += chartParamAttr(n, "tableSelection", "chartTableSelection", n.chartTableSelection, "row", true);
    a += chartBoolAttr(n, "criticalPath", "chartCriticalPath", n.chartCriticalPath, false);
    a += chartParamAttr(n, "holidays", "chartHolidays", n.chartHolidays, "", true);
    a += chartParamAttr(n, "weekend", "chartWeekend", n.chartWeekend, "", true);
    a += chartParamAttr(n, "timezone", "chartTimezone", n.chartTimezone, "", true);
    a += chartBoolAttr(n, "showNonWorkingDays", "chartShowNonWorkingDays", n.chartShowNonWorkingDays, true);
    a += chartBoolAttr(n, "progressColumn", "chartProgressColumn", n.chartProgressColumn, false);
    a += chartParamAttr(n, "bins", "chartBins", n.chartBins, "10", false);
    a += chartParamAttr(n, "min", "chartMin", n.chartMin, "", false);
    a += chartParamAttr(n, "max", "chartMax", n.chartMax, "", false);
    if (n.chartInteractive) a += " interactive: true";
    if (n.chartStateMode && n.chartStateMode !== "editing") a += " stateMode: \"" + esc(n.chartStateMode) + "\"";
    a += chartBoolAttr(n, "stateHints", "chartStateHints", n.chartStateHints, true);
  a += chartParamAttr(n, "stateHintBg", "chartStateHintBg", n.chartStateHintBg, "", true);
  a += chartParamAttr(n, "stateHintOpacity", "chartStateHintOpacity", n.chartStateHintOpacity, "", true);
  a += chartParamAttr(n, "stateHintTextColor", "chartStateHintTextColor", n.chartStateHintTextColor, "", true);
  a += chartParamAttr(n, "stateHintBorderColor", "chartStateHintBorderColor", n.chartStateHintBorderColor, "", true);
  a += chartParamAttr(n, "stateHintFont", "chartStateHintFont", n.chartStateHintFont, "", true);
  a += chartParamAttr(n, "stateHintFontSize", "chartStateHintFontSize", n.chartStateHintFontSize, "", true);
  a += chartParamAttr(n, "stateHintBorderWidth", "chartStateHintBorderWidth", n.chartStateHintBorderWidth, "", true);
  a += chartParamAttr(n, "stateHintPadding", "chartStateHintPadding", n.chartStateHintPadding, "", true);
  a += chartParamAttr(n, "stateHintRadius", "chartStateHintRadius", n.chartStateHintRadius, "", true);
    if (n.chartColorPicker === false) a += " colorPicker: false";
    a += chartBoolAttr(n, "showResources", "chartShowResources", n.chartShowResources, false);
    a += chartBoolAttr(n, "showProgress", "chartShowProgress", n.chartShowProgress, false);
    a += chartBoolAttr(n, "showSlack", "chartShowSlack", n.chartShowSlack, false);
    a += chartBoolAttr(n, "showCosts", "chartShowCosts", n.chartShowCosts, false);
    a += chartBoolAttr(n, "statusColors", "chartStatusColors", n.chartStatusColors, false);
    a += chartBoolAttr(n, "autoSchedule", "chartAutoSchedule", n.chartAutoSchedule, false);
    a += chartParamAttr(n, "ganttScale", "chartGanttScale", n.chartGanttScale, "auto", true);
    a += chartParamAttr(n, "linkColor", "chartLinkColor", n.chartLinkColor, "#94a3b8", true);
    a += chartParamAttr(n, "linkWidth", "chartLinkWidth", n.chartLinkWidth, "1", false);
    a += chartParamAttr(n, "linkGap", "chartLinkGap", n.chartLinkGap, "4", false);
    a += chartParamAttr(n, "linkStyle", "chartLinkStyle", n.chartLinkStyle, "auto", true);
    a += chartParamAttr(n, "arrowType", "chartArrowType", n.chartArrowType, "triangle", true);
    a += chartParamAttr(n, "arrowSize", "chartArrowSize", n.chartArrowSize, "5", false);
    a += chartParamAttr(n, "barBorderColor", "chartBarBorderColor", n.chartBarBorderColor, "", true);
    if (n.onTaskChange) a += " onTaskChange: " + n.onTaskChange;
    if (n.onDataChange) a += " onDataChange: " + n.onDataChange;
    if (n.onCalendarChange) a += " onCalendarChange: " + n.onCalendarChange;
    if (n.onResourceChange) a += " onResourceChange: " + n.onResourceChange;
    if (n.onCommit) a += " onCommit: " + n.onCommit;
    ["onShapeClick", "onShapeDoubleClick", "onShapeMouseOver", "onShapeMouseOut", "onShapeContextMenu",
     "onLinkClick", "onLinkDoubleClick", "onLinkMouseOver", "onLinkMouseOut", "onLinkContextMenu",
     "onTagClick", "onTagDoubleClick", "onTagMouseOver", "onTagMouseOut", "onTagContextMenu"].forEach(function(k) {
      if (n[k]) a += " " + k + ": " + n[k];
    });
    a += chartBoolAttr(n, "showGrid", "chartGrid", n.chartGrid, true);
    a += chartBoolAttr(n, "showLegend", "chartLegend", n.chartLegend, true);
    a += chartBoolAttr(n, "showValues", "chartValues", n.chartValues, false);
    if (n.chartExtra && typeof n.chartExtra === "object") {
      Object.keys(n.chartExtra).forEach(function(k) {
        var v = n.chartExtra[k];
        if (v === undefined || v === null || v === "") return;
        a += " " + k + ": " + v;
      });
    }
  }
  if (n.kind === "treeitem" && n.key) a += " key: \"" + esc(n.key) + "\"";
  if (n.kind === "treeitem" && n.open) a += " open: true";
  if (n.kind === "treeitem" && n.editable !== undefined && n.editable !== null && n.editable !== "inherit") {
    a += " editable: " + (n.editable ? "true" : "false");
  }
  if (n.kind === "treeitem" && n.renameIcon) a += " renameIcon: \"" + esc(n.renameIcon) + "\"";
  if (n.kind === "treeitem" && n.showRenameIcon !== undefined && n.showRenameIcon !== null) {
    a += " showRenameIcon: " + (n.showRenameIcon ? "true" : "false");
  }
  if (n.kind === "tree" && n.editable) a += " editable: true";
  if (n.kind === "tree" && n.renameIcon) a += " renameIcon: \"" + esc(n.renameIcon) + "\"";
  if (n.kind === "tree" && n.showRenameIcon === false) a += " showRenameIcon: false";
  if (n.kind === "tree" && n.editableDepth) a += " editableDepth: " + n.editableDepth;
  if (n.kind === "tree" && n.renameKey) a += " renameKey: " + n.renameKey;
  if (n.kind === "tree" && n.renameValue) a += " renameValue: " + n.renameValue;
  if (n.onClick) a += " onClick: " + n.onClick;
  if (n.onDoubleClick) a += " onDoubleClick: " + n.onDoubleClick;
  if (n.onMouseEnter) a += " onMouseEnter: " + n.onMouseEnter;
  if (n.onMouseLeave) a += " onMouseLeave: " + n.onMouseLeave;
  if (n.onFocus) a += " onFocus: " + n.onFocus;
  if (n.onBlur) a += " onBlur: " + n.onBlur;
  if (n.onInput) a += " onInput: " + n.onInput;
  if (n.onChange) a += " onChange: " + n.onChange;
  if (n.onRename) a += " onRename: " + n.onRename;
  if (n.onKeyDown) a += " onKeyDown: " + n.onKeyDown;
  if (n.onKeyUp) a += " onKeyUp: " + n.onKeyUp;
  if (n.onSubmit) a += " onSubmit: " + n.onSubmit;
  if (n.onReset) a += " onReset: " + n.onReset;
  if (n.onDragStart) a += " onDragStart: " + n.onDragStart;
  if (n.onDragEnter) a += " onDragEnter: " + n.onDragEnter;
  if (n.onDragOver) a += " onDragOver: " + n.onDragOver;
  if (n.onDragLeave) a += " onDragLeave: " + n.onDragLeave;
  if (n.onDrop) a += " onDrop: " + n.onDrop;
  if (n.onDragEnd) a += " onDragEnd: " + n.onDragEnd;
  if (n.draggable) a += " draggable: true";
  // These are SUI expressions (item.id, index, a state). Keep them raw so a
  // list variable such as item is not quoted as a string by bondOut.
  if (n.dragKey) a += " dragKey: " + n.dragKey;
  if (n.dropKey) a += " dropKey: " + n.dropKey;
  if (n.dropFiles) a += " dropFiles: " + n.dropFiles;
  if (n.pickFiles) a += " pickFiles: " + n.pickFiles;
  if (n.pickMode && n.pickMode !== "native") a += " pickMode: \"" + esc(n.pickMode) + "\"";
  if (n.directory) a += " directory: true";
  if (n.kind === "toast") {
    if (n.toastHAlign) a += " hAlign: \"" + esc(n.toastHAlign) + "\"";
    if (n.toastVAlign) a += " vAlign: \"" + esc(n.toastVAlign) + "\"";
    if (n.toastHOffset) a += " hOffset: \"" + esc(n.toastHOffset) + "\"";
    if (n.toastVOffset) a += " vOffset: \"" + esc(n.toastVOffset) + "\"";
    // Keep the first compact syntax round-tripping until the user changes an
    // axis; toastSetAxes then clears the legacy fields.
    if (!n.toastHAlign && !n.toastVAlign) {
      if (n.toastPosition) a += " position: \"" + esc(n.toastPosition) + "\"";
      if (n.toastOffset) a += " offset: \"" + esc(n.toastOffset) + "\"";
    }
    if (n.toastDuration) a += " duration: " + n.toastDuration;
    if (n.toastMax) a += " max: " + n.toastMax;
    if (n.toastIcons === false) a += " icons: false";
  }
  if (n.kind === "combobox" && n.empty) a += " empty: \"" + esc(n.empty) + "\"";
  if (n.tooltip) a += " tooltip: \"" + escText(n.tooltip) + "\"";
  if (n.classExpr) a += " class: " + n.classExpr;
  else if (n.class) a += " class: \"" + esc(n.class) + "\"";
  if (n.styleExpr) a += " style: " + cssSafeExpr(n.styleExpr);
  else if (n.style) a += " style: \"" + esc(n.style) + "\"";
  if (n.kind === "iframe") {
    if (n.srcAllow) a += " srcAllow: \"" + esc(n.srcAllow) + "\"";
    if (n.sandbox !== undefined && n.sandbox !== null) a += " sandbox: \"" + esc(n.sandbox) + "\"";
    if (n.allow) a += " allow: \"" + esc(n.allow) + "\"";
    if (n.referrerpolicy) a += " referrerpolicy: \"" + esc(n.referrerpolicy) + "\"";
    if (n.loading) a += " loading: \"" + esc(n.loading) + "\"";
    if (n.width) a += " width: \"" + esc(String(n.width)) + "\"";
    if (n.height) a += " height: \"" + esc(String(n.height)) + "\"";
  }
  return a;
}
function toSui(n, depth) {
  if (n.raw) {
    var indr = ""; for (var i = 0; i < depth; i++) indr += "\t";
    return indr + n.raw.split("\n").join("\n" + indr) + "\n";
  }
  if (n.kind === "table" && n.cols) return tableToSui(n, depth);
  if (n.kind === "list") return listToSui(n, depth);
  if (n.kind === "text" && n.runs) return textToSui(n, depth);
  var ind = ""; for (var i = 0; i < depth; i++) ind += "\t";
  var line = ind + n.kind;
  // A toast host has no visible text and is emitted as toast class: ...
  if (!(n.kind === "toast" && !n.text)) line += " \"" + escText(n.text) + "\"";
  line += attrsOf(n);
  if (n.children && n.children.length) {
    line += " {\n";
    for (var j = 0; j < n.children.length; j++) line += toSui(n.children[j], depth + 1);
    line += ind + "}\n";
  } else line += "\n";
  return line;
}
function cellCovered(n, r, c) {
  var s = n.spans[r][c];
  return s && s.cs === 0;
}
function cellStyleStr(n, r, c) {
  var s = n.cellStyles && n.cellStyles[r] && n.cellStyles[r][c];
  if (!s) return "";
  var out = "";
  if (s.bg) out += "background:" + s.bg + ";";
  if (s.color) out += "color:" + s.color + ";";
  if (s.extra) out += s.extra;
  return out;
}
function rowStyleStr(n, r) {
  var s = n.rowStyles && n.rowStyles[r];
  if (!s) return "";
  var out = "";
  if (s.bg) out += "background:" + s.bg + ";";
  if (s.color) out += "color:" + s.color + ";";
  if (s.extra) out += s.extra;
  return out;
}
function tcellToSui(n, r, c) {
  var s = n.spans[r][c] || { cs: 1, rs: 1 };
  var a = "";
  if (r < n.header) a += " head: true";
  if (s.cs > 1) a += " colspan: " + s.cs;
  if (s.rs > 1) a += " rowspan: " + s.rs;
  var st = cellStyleStr(n, r, c);
  if (st) a += " style: \"" + esc(st) + "\"";
  return "tcell \"" + escText(n.cells[r][c]) + "\"" + a;
}
function tableToSui(n, depth) {
  function ind(level) { var s = ""; for (var i = 0; i < level; i++) s += "\t"; return s; }
  var extra = (n.editable ? " editable: true" : "") + (n.fixed ? " fixed: true" : "");
  var line = ind(depth) + "table" + attrsOf(n) + extra + " {\n";
  if (n.captionShow !== false && n.caption && n.caption.trim()) {
    line += ind(depth + 1) + "caption \"" + escText(n.caption) + "\"" + (n.capstyle ? " style: \"" + esc(n.capstyle) + "\"" : "") + "\n";
  }
  // header rows (thead)
  line += ind(depth + 1) + "thead {\n";
  for (var r = 0; r < n.header; r++) {
    var rstyle = rowStyleStr(n, r);
    line += ind(depth + 2) + "trow" + (rstyle ? " style: \"" + esc(rstyle) + "\"" : "") + " {\n";
    for (var c = 0; c < n.cols; c++) {
      if (cellCovered(n, r, c)) continue;
      line += ind(depth + 3) + tcellToSui(n, r, c) + "\n";
    }
    line += ind(depth + 2) + "}\n";
  }
  line += ind(depth + 1) + "}\n";
  // body rows (tbody)
  line += ind(depth + 1) + "tbody {\n";
  for (var r = n.header; r < n.rows; r++) {
    var rstyle2 = rowStyleStr(n, r);
    line += ind(depth + 2) + "trow" + (rstyle2 ? " style: \"" + esc(rstyle2) + "\"" : "") + " {\n";
    for (var c = 0; c < n.cols; c++) {
      if (cellCovered(n, r, c)) continue;
      line += ind(depth + 3) + tcellToSui(n, r, c) + "\n";
    }
    line += ind(depth + 2) + "}\n";
  }
  line += ind(depth + 1) + "}\n";
  line += ind(depth) + "}\n";
  return line;
}
function listToSui(n, depth) {
  function ind(level) { var s = ""; for (var i = 0; i < level; i++) s += "\t"; return s; }
  var tag = n.ordered ? "ol" : "ul";
  var style = n.marker ? ' style: "list-style-type: ' + esc(n.marker) + '"' : "";
  var line = ind(depth) + tag + style + attrsOf(n) + " {\n";
  for (var i = 0; i < n.items.length; i++) {
    line += ind(depth + 1) + "li \"" + escText(n.items[i]) + "\"\n";
  }
  line += ind(depth) + "}\n";
  return line;
}
function runToSui(r) {
  var style = "";
  if (r.u) style += "text-decoration: underline;";
  if (r.color) style += "color: " + r.color + ";";
  if (r.bg) style += "background: " + r.bg + ";";
  if (r.font) style += "font-family: " + r.font + ";";
  var s = "text" + (style ? ' style: "' + esc(style) + '"' : '') + ' "' + escText(r.text) + '"';
  if (r.b) s = "strong { " + s + " }";
  if (r.i) s = "em { " + s + " }";
  return s;
}
function textToSui(n, depth) {
  function ind(level) { var s = ""; for (var i = 0; i < level; i++) s += "\t"; return s; }
  var style = "";
  if (n.align) style += "text-align: " + n.align + ";";
  if (n.size) style += "font-size: " + n.size + ";";
  var line = ind(depth) + (n.tag || "p") + (style ? ' style: "' + esc(style) + '"' : "") + attrsOf(n) + " {\n";
  var runs = n.runs || [];
  for (var i = 0; i < runs.length; i++) {
    line += ind(depth + 1) + runToSui(runs[i]) + "\n";
  }
  line += ind(depth) + "}\n";
  return line;
}
// indentFunctionBody gives every generated function body one level of
// indentation. The designer stores the body without its outer indentation
// (stmtsToSui starts at column 0), so this is the one place that must add it
// back when regenerating source.
function indentFunctionBody(body) {
  var lines = String(body || "").replace(/\r\n?/g, "\n").split("\n");
  while (lines.length && !lines[0].trim()) lines.shift();
  while (lines.length && !lines[lines.length - 1].trim()) lines.pop();
  if (!lines.length) return "\t";
  return lines.map(function(line) { return line ? "\t" + line : ""; }).join("\n");
}
function generate() {
  var src = "";
  if (root.imports && root.imports.length) src += "import { " + root.imports.join(", ") + " } from \"std\"\n";
  if (root.css && root.css.trim()) src += "css \"" + esc(root.css) + "\"\n";
  root.states.forEach(function(s) {
    if (!s.name) return;
    var init = s.init;
    if (!init) init = (s.type === "int" || s.type === "float") ? "0" : "\"\"";
    src += "state " + s.name + (s.type ? ": " + s.type : "") + " = " + init + "\n";
  });
  (root.components || []).forEach(function(c) { if (c) src += c + "\n"; });
  // Every page is emitted as its own view: the first page comes first so it
  // stays the project's root page.
  (root.pages || [{ name: "main", title: root.title || "", onLoad: "", children: root.children }]).forEach(function(p, i) {
    var nm = p.name || ("page" + (i + 1));
    src += "view " + nm + " {\n";
    if (p.title) src += "\ttitle \"" + esc(p.title) + "\"\n";
    if (p.onLoad) src += "\tonLoad: " + p.onLoad + "\n";
    var kids = p.children || [];
    for (var k = 0; k < kids.length; k++) src += toSui(kids[k], 1);
    src += "}\n";
  });
  root.funcs.forEach(function(f) {
    if (!f.name) return;
    src += (f.async ? "async " : "") + "fn " + f.name + "(" + (f.params || "") + ") -> " + (f.ret || "null") + " {\n" + indentFunctionBody(f.body) + "\n}\n";
  });
  return src;
}
var pvSeq = 0;
function autoPreview() {
  var src = document.getElementById("out").value;
  if (!src) return;
  // Serialize preview loads: a stale render response must not reload the
  // iframe after a newer one (it would show an older version of the project).
  var mySeq = ++pvSeq;
  var fd = new URLSearchParams();
  fd.set("sui", src);
  fetch("/__gui__/render", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (mySeq !== pvSeq) return;
    if (j.ok) { document.getElementById("prev").src = "/__gui__/preview?ts=" + Date.now(); }
    else { setStatus((j.diagnostics || []).join(" | ")); }
  }).catch(function(){
    if (mySeq === pvSeq) setStatus("preview failed: server unavailable");
  });
}
function downloadSui() {
  var blob = new Blob([suiSource || generate()], { type: "text/plain" });
  var a = document.createElement("a"); a.href = URL.createObjectURL(blob); a.download = "app.sui"; a.click();
}
function loadTree(model) {
  root = { imports: model.imports || [], states: model.states || [], funcs: model.funcs || [], css: model.css || "", components: model.components || [], page: 0,
    pages: (model.pages && model.pages.length ? model.pages.map(function(p){ return { name: p.name || "main", title: p.title || "", onLoad: p.onLoad || "", children: p.elements || p.children || [] }; }) : [{ name: "main", title: model.title || "", onLoad: model.onLoad || "", children: model.elements || [] }]) };
  root.pages.forEach(function(p){ normalizeTables(p.children); });
  remount();
  computeElementLines();
  selPath = null; past = []; future = []; selState = ""; selCell = null; selRow = null;
  __treeCollapsed = Object.create(null);
  endTextHistoryBurst();
  renderAll(); autoPreview();
}

// ---- pages ----------------------------------------------------------------
// A page is an editable view layer: the first page (main) is the project's
// root view, further pages are additional non-invoked views. Page chips are
// rendered as TABS inside the preview pane; clicking one activates it (tree,
// inspector and preview all follow) and the inspector carries the page
// attributes (name/title + duplicate/delete/save-as-template actions).
function renderPages() {
  var box = document.getElementById("pagetabs"); if (!box) return;
  box.innerHTML = "";
  (root.pages || []).forEach(function(p, i) {
    var chip = document.createElement("span"); chip.className = "pagechip" + (i === root.page ? " on" : "");
    var b2 = document.createElement("button"); b2.className = "palbtn";
    b2.textContent = p.name || ("page" + (i + 1));
    b2.title = (p.title ? p.title + " — " : "") + "click to activate, double-click to rename";
    b2.onclick = function(){ if (root.page !== i) switchPage(i); };
    b2.ondblclick = function(){ renamePageFlow(); };
    chip.appendChild(b2);
    box.appendChild(chip);
  });
}
function switchPage(i, skipSnap) {
  if (i < 0 || i >= root.pages.length) return;
  if (!skipSnap) snap("page " + root.pages[i].name);
  root.page = i;
  remount();
  selPath = null; multiSel = []; selCell = null; selRow = null; selCol = null;
  computeElementLines();
  renderAll();
  autoPreview();
  previewPageSync();
}
// uniquePageName: valid SUI ident, unused, not prefixed by the reserved
// render_ namespace (the typechecker rejects render_* names).
function uniquePageName(base) {
  base = (base || "page").toLowerCase().replace(/[^a-z0-9_]/g, "");
  if (!base || /^render_/.test(base + "_")) { if (/^render_/.test(base)) base = "p_" + base; }
  if (!base) base = "page";
  if (!/^[a-z_]/.test(base)) base = "p" + base;
  var names = (root.pages || []).map(function(p){ return p.name; });
  if (names.indexOf(base) < 0) return base;
  var n = 2;
  while (names.indexOf(base + n) >= 0) n++;
  return base + n;
}
// tagUsedAnywhere reports whether an element somewhere invokes "tag" as a
// component (renaming/deleting such a view would break the callers).
function tagUsedAnywhere(tag) {
  function walk(nodes) {
    if (!nodes) return false;
    for (var i = 0; i < nodes.length; i++) {
      if (nodes[i].kind === tag) return true;
      if (walk(nodes[i].children)) return true;
    }
    return false;
  }
  for (var i = 0; i < root.pages.length; i++) {
    if (i === root.page) continue;
    if (walk(root.pages[i].children)) return true;
  }
  if (walk(root.children)) return true;
  // component declarations are raw strings: scanning their text is cheap
  for (var c = 0; c < (root.components || []).length; c++) {
    if (new RegExp("\\b" + tag + "\\b").test(String(root.components[c]))) return true;
  }
  return false;
}
function renamePageFlow() {
  var p = curPage();
  askPrompt("Rename page", "New name", p.name, function(raw) {
    var name = uniquePageName(raw);
    if (!/^[a-z_][a-z0-9_]*$/i.test(name)) { showMsg("Invalid name", "Use only letters, digits and _ , not starting with a digit."); return; }
    if (/^render_/.test(name)) { showMsg("Invalid name", "Names starting with render_ are reserved."); return; }
    if (tagUsedAnywhere(p.name)) { showMsg("Page in use", "\"" + p.name + "\" is invoked as a component by other elements: remove those invocations first."); return; }
    snap("rename page " + p.name);
    root.pages[root.page].name = name;
    setOut(); renderPages(); autoPreview();
    var fi = document.getElementById("f_page"); if (fi) fi.value = curPage().name;
  }, "Letters, digits and _ only");
}
// mergeFromClientModel folds a built-in page template's states/functions/css
// into the project, skipping names that already exist.
function mergeFromClientModel(m) {
  (m.states || []).forEach(function(s) {
    if (!s.name || RESERVED_NAME.test(s.name) || !IDENT.test(s.name)) return;
    if (!root.states.some(function(x){ return x.name === s.name; })) root.states.push(s);
  });
  (m.funcs || []).forEach(function(f) {
    if (!f.name || RESERVED_NAME.test(f.name) || !IDENT.test(f.name)) return;
    if (!root.funcs.some(function(x){ return x.name === f.name; })) root.funcs.push(f);
  });
  var extra = (m.css || "").trim();
  if (extra && (root.css || "").indexOf(extra) < 0) root.css = ((root.css || "") + (root.css && root.css.trim() ? "\n" : "") + extra);
}
// Built-in page templates (also selectable in the new-page dialog).
function builtinPageTemplates() {
  return [
    { name: "Dashboard", model: { title: "Dashboard", states: [], funcs: [{ name: "refresh", params: "", ret: "null" }], css: ".card { padding: 16px; border-radius: 10px; background: #1b1f27; }\n.row { gap: 12px; }",
      elements: [ tx("Dashboard", "h1"), tx("Overview", "p"), nodeIn("row", [b("Refresh", "refresh")], "card"), nodeIn("col", [tx("status", "p")]) ] } },
    { name: "Form", model: { title: "Form", states: [{ name: "name", init: "\"\"" }], funcs: [{ name: "submit", params: "", ret: "null" }], css: ".field { padding: 10px; border: 1px solid #3a4152; border-radius: 6px; }",
      elements: [ tx("Contact", "h1"), nodeIn("col", [bindInp("name"), b("Submit", "submit")], "field") ] } },
    { name: "CLI wrapper", model: { title: "CLI wrapper", states: [{ name: "cmd", init: "\"\"" }, { name: "output", init: "\"\"" }], funcs: [{ name: "run", params: "", ret: "null" }], css: ".mono { font-family: monospace; }",
      elements: [ tx("Command", "h1"), bindInp("cmd"), nodeIn("row", [b("Run", "run")]), tx("Output appears here", "p") ] } }
  ];
}
var __pageTplChoice = null;
function addPageFlow() {
  var dlg = document.getElementById("pagedlg");
  if (!dlg) { addPageEmpty(); return; }
  document.getElementById("page-name").value = "";
  document.getElementById("page-mode").value = "empty";
  __pageTplChoice = null;
  document.getElementById("page-tpls").style.display = "none";
  renderPageTplChoices();
  dlg.classList.add("on");
  document.getElementById("page-name").focus();
}
function renderPageTplChoices() {
  var box = document.getElementById("page-tpls");
  box.innerHTML = "";
  builtinPageTemplates().forEach(function(t) {
    var c = document.createElement("button"); c.className = "palbtn"; c.textContent = "built-in: " + t.name;
    c.onclick = function(){ pickPageTpl(t.name, c); };
    box.appendChild(c);
  });
  var hd = document.createElement("div"); hd.className = "p-hint"; hd.textContent = "My templates (library):";
  box.appendChild(hd);
  var found = false;
  fetch("/__gui__/templates").then(function(r){ return r.json(); }).then(function(j) {
    (j.templates || []).forEach(function(t) {
      if (t.kind !== "page") return;
      found = true;
      var c = document.createElement("button"); c.className = "palbtn"; c.textContent = t.name;
      c.onclick = function(){ pickPageTpl(t.name, c, true); };
      box.appendChild(c);
    });
    if (!found) {
      var none = document.createElement("div"); none.className = "p-hint"; none.textContent = "(no saved page template — use \"Save as template\")";
      box.appendChild(none);
    }
  });
}
function pickPageTpl(name, chip, isUser) {
  __pageTplChoice = { name: name, user: !!isUser };
  var all = document.querySelectorAll("#page-tpls .palbtn");
  for (var i = 0; i < all.length; i++) all[i].classList.toggle("on", all[i] === chip);
}
function pageModeChanged() {
  var m = document.getElementById("page-mode").value;
  document.getElementById("page-tpls").style.display = m === "template" ? "" : "none";
}
function pageSubmit() {
  var rawName = document.getElementById("page-name").value.trim();
  var name = uniquePageName(rawName || "new page");
  var mode = document.getElementById("page-mode").value;
  var chosen = __pageTplChoice; // closePage() clears the module state
  closePage();
  if (mode === "empty") { addPageEmpty(name); return; }
  if (!chosen) { showMsg("Choose a template", "Select \"Empty\" or a template from the list."); return; }
  if (chosen.user) { addUserPageTemplate(chosen.name, name); return; }
  var t = null;
  builtinPageTemplates().forEach(function(b){ if (b.name === chosen.name) t = b; });
  if (!t) { addPageEmpty(name); return; }
  mergeFromClientModel(t.model);
  addPage(name, t.model.title || "", (t.model.elements || []).map(function(e){ return JSON.parse(JSON.stringify(e)); }));
}
function closePage() {
  document.getElementById("pagedlg").classList.remove("on");
  __pageTplChoice = null;
}
function addPageEmpty(name) {
  snap("new page");
  root.pages.push({ name: uniquePageName(name), title: "", onLoad: "", children: [] });
  switchPage(root.pages.length - 1, true);
  setStatus("page added: " + curPage().name);
}
function addPage(name, title, els) {
  snap("new page from template");
  root.pages.push({ name: name || uniquePageName("new page"), title: title || "", onLoad: "", children: els || [] });
  normalizeTables(root.pages[root.pages.length - 1].children);
  switchPage(root.pages.length - 1, true);
  setStatus("page added: " + curPage().name);
}
function addUserPageTemplate(tplName, pageName) {
  fetch("/__gui__/template?name=" + encodeURIComponent(tplName)).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) { setStatus("template not found"); return; }
    var fd = new URLSearchParams(); fd.set("sui", j.sui); fd.set("project", typeof suiSource === "string" ? suiSource : "");
    fetch("/__gui__/template/open", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(m){
      if (!m.ok) { showMsg("Template cannot be inserted", "The template is not valid in the current project:\n\n" + ((m.diagnostics || []).map(function(d){ return "L" + d.line + ": " + d.msg + (d.hint ? " — " + d.hint : ""); }).join("\n") || "invalid template")); return; }
      var els = m.elements || [];
      if (!els.length) { setStatus("template is empty"); return; }
      addPage(pageName, m.title || "", els);
    });
  });
}
function delPageFlow() {
  if (root.pages.length <= 1) { showMsg("Cannot delete", "The project must keep at least one page."); return; }
  var p = curPage();
  confirmBox("Delete the page?", "\"" + (p.name || "main") + "\" will be removed from the project (the SUI source will lose it as well).", function(){
    if (tagUsedAnywhere(p.name)) { showMsg("Page in use", "\"" + p.name + "\" is invoked as a component by other elements."); return; }
    snap("delete page " + p.name);
    var name = p.name;
    root.pages.splice(root.page, 1);
    if (root.page >= root.pages.length) root.page = root.pages.length - 1;
    remount();
    selPath = null; multiSel = [];
    computeElementLines();
    renderAll();
    autoPreview();
    setStatus("page deleted: " + name);
  });
}
function duplicatePage() {
  var p = curPage();
  snap("duplicate page " + p.name);
  var name = uniquePageName(p.name);
  var copy = { name: name, title: (p.title || "") + (p.title ? " (copia)" : ""), onLoad: p.onLoad || "", children: JSON.parse(JSON.stringify(p.children || [])) };
  normalizeTables(copy.children);
  root.pages.splice(root.page + 1, 0, copy);
  switchPage(root.page + 1, true);
  setStatus("page duplicated: " + name);
}
function savePageTemplateFlow() {
  var p = curPage();
  var snippet = "";
  if (p.title) snippet += "title \"" + esc(p.title) + "\"\n";
  for (var i = 0; i < p.children.length; i++) snippet += toSui(p.children[i], 1);
  if (!snippet.trim()) { setStatus("empty page: nothing to save"); return; }
  askPrompt("Save page as template", "Template name", p.title || (p.name === "main" ? "My page" : p.name), function(name){
    var fd = new URLSearchParams(); fd.set("name", name); fd.set("kind", "page"); fd.set("sui", snippet);
    fetch("/__gui__/template/save", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
      if (!j.ok) { setStatus("save failed: " + (j.error || "")); return; }
      renderTemplates();
      setStatus("saved page template: " + name);
    });
  }, "It will be available when adding a new page, in every project.");
}
function previewPageSync() {
  var ifr = document.getElementById("prev"); if (!ifr) return;
  var name = curPage().name;
  var tries = 0;
  (function tick() {
    var w; try { w = ifr.contentWindow; } catch(e) { return; }
    if (w && typeof w.__sui_page === "function") { try { w.__sui_page(name); } catch(e){} return; }
    if (++tries < 25) setTimeout(tick, 120);
  })();
}
(function wirePreviewPageSync() {
  var ifr = document.getElementById("prev");
  ifr.addEventListener("load", function(){ previewPageSync(); requestLiveStyles(true); });
})();

// ---- blocks (reusable element groups) -------------------------------------
// saveBlockFlow stores the selected elements as a library block (persisted in
// %APPDATA%, usable from every project). insertTemplate places one back into
// the current tree, validated inside the project environment.
function h(kind, text) { var n = newNode(kind); n.text = text; return n; }
function tx(text, tag) { var n = newTextNode(); n.tag = tag || "p"; n.runs = [{ text: text || "", b: false, i: false, u: false, color: "", bg: "", font: "" }]; return n; }
function nodeWith(kind, text) { return h(kind, text); }
function nodeIn(kind, children, cls) { var n = newNode(kind); n.children = children.slice(); if (cls) n.class = cls; return n; }
function b(text, onClick) { var n = newNode("button"); n.text = text; n.onClick = onClick; return n; }
function bindInp(state) { var n = newNode("input"); n.value = state; n.placeholder = state; return n; }

// open .sui ----------------------------------------------------------------
function openFile() {
  var f = document.getElementById("filein").files[0];
  if (!f) return;
  var rd = new FileReader();
  rd.onload = function() {
    var fd = new URLSearchParams(); fd.set("sui", rd.result);
    fetch("/__gui__/open", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
      if (j.ok) { loadTree(j); }
      else { setStatus((j.diagnostics || []).join("\n")); }
    });
  };
  rd.readAsText(f);
}
document.getElementById("filein").onchange = openFile;
document.getElementById("fontfile").onchange = function(){ uploadFont(this.files[0]); };

// preview click -> select node ----------------------------------------------
window.addEventListener("message", function(ev) {
  if (ev.source !== document.getElementById("prev").contentWindow) return;
  if (!ev.data) return;
  // Clicking the preview iframe must dismiss a context menu just like a
  // click elsewhere in the parent document.
  closeContextMenus();
  if (ev.data.ty === "sui-style") { applyLiveStyleMessage(ev.data); return; }
  if (ev.data.ty === "sui") { selectByDfs(ev.data.id); return; }
  if (ev.data.ty === "sui-mselect") { selectPaths(idsToPaths(ev.data.ids || [])); setStatus((ev.data.ids || []).length ? ("selected " + ev.data.ids.length + " element(s)") : "nothing selected"); return; }
  if (ev.data.ty === "sui-drop") { dropFromPalette(ev.data.id); return; }
  if (ev.data.ty === "sui-key") { typeToEdit(String(ev.data.key || "")); return; }
  if (ev.data.ty === "sui-resize") { resizeById(ev.data.id, ev.data.w, ev.data.h, ev.data.l, ev.data.t); return; }
  if (ev.data.ty === "sui-colw") {
    var pc = pickMap[ev.data.id];
    if (pc && pc.t === "cell" && pc.pg === root.page) {
      var tn = byPath(pc.p, root.children);
      colExtraWrite(tn, pc.c, "width", Math.max(8, Math.round(ev.data.w)) + "px", "column width");
    }
    return;
  }
});
function dropFromPalette(targetId) {
  if (!dragKind) return;
  var kind = dragKind;
  var t = pickMap[targetId];
  if (!t || !t.p || t.pg !== root.page) { setStatus("cannot drop here"); return; }
  var node = byPath(t.p, root.children);
  withUndo(function() {
    var n = newNodeForKind(kind);
    var insPath;
    if (canNest(node.kind)) {
      node.children.push(n);
      insPath = t.p.concat([node.children.length - 1]);
    } else {
      var par = t.p.slice(0, -1);
      var arr = par.length === 0 ? root.children : byPath(par, root.children).children;
      var pos = t.p[t.p.length - 1] + 1;
      arr.splice(pos, 0, n);
      insPath = par.concat([pos]);
    }
    selPath = (kind === "table") ? wrapTableInScroll(insPath) : insPath;
  }, "insert " + kind);
  dragKind = null;
}
function resizeById(id, w, h, l, t) {
  var pt = pickMap[id];
  if (!pt || !pt.p || pt.pg !== root.page) return;
  var n = byPath(pt.p, root.children);
  var px = function(v) {
    if (v > 0) return Math.round(v) + "px";
    return "";
  };
  var decls = [];
  if (n.kind === "table" && n.rows) {
    // Distribute the requested height evenly across ALL rows (header too):
    // the base rule gives td/th height from var(--t-rowh). No total height is
    // set on the table (that would make the browser add the extra twice).
    var per = h > 0 ? Math.max(1, Math.round(h / n.rows)) : 0;
    decls.push({ prop: "width", value: px(w) });
    decls.push({ prop: "--t-rowh", value: per ? per + "px" : "" });
    decls.push({ prop: "height", value: "" });
  } else {
    decls.push({ prop: "width", value: px(w) });
    decls.push({ prop: "height", value: px(h) });
  }
  if (l || t) {
    decls.push({ prop: "position", value: "relative" });
    decls.push({ prop: "left", value: px(l || 0) });
    decls.push({ prop: "top", value: px(t || 0) });
  }
  var cls = ensureClass(n);
  snap("resize");
  setCssOnClass(cls, decls);
}

// keyboard -----------------------------------------------------------------
document.addEventListener("keydown", function(ev) {
  var mod = ev.ctrlKey || ev.metaKey;
  var ae = document.activeElement;
  var tag = ae && ae.tagName;
  var editing = tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT" || (ae && ae.isContentEditable);
  // The event target also tells where the key really came from (for synthetic
  // events and when focus moved between phases).
  var ttag = ev.target && ev.target.tagName;
  var tediting = ttag === "INPUT" || ttag === "TEXTAREA" || ttag === "SELECT" || (ev.target && ev.target.isContentEditable);
  var inInspector = !!(ev.target && ev.target.closest && ev.target.closest("#inspector"));
  // Undo/redo use the unified designer history. The SUI editor intercepts the
  // keys in editorKeydown and stops propagation; inspector fields are included
  // because a property edit is a designer action, while prompt/modal text
  // fields keep their native undo.
  var modelUndoScope = (!editing && !tediting) || inInspector;
  if (mod && (ev.key === "z" || ev.key === "Z" || ev.key === "u" || ev.key === "U") && !ev.shiftKey && modelUndoScope) {
    ev.preventDefault(); undo(); return;
  }
  if (mod && ((ev.shiftKey && (ev.key === "z" || ev.key === "Z")) || ev.key === "y" || ev.key === "Y") && modelUndoScope) {
    ev.preventDefault(); redo(); return;
  }
  if (mod && !(editing || tediting)) {
    if (ev.key === "c") { ev.preventDefault(); copySelected(); return; }
    if (ev.key === "x") { ev.preventDefault(); cutSelected(); return; }
    if (ev.key === "v") { ev.preventDefault(); pasteClipboard(); return; }
  }
  if ((ev.key === "Delete" || ev.key === "Backspace") && !editing) { ev.preventDefault(); deleteSelected(); return; }
  // Type-to-edit: if a text-capable element is selected and the user starts
  // typing, redirect the keystroke into its "text" field.
  if (!selPath || mod || ev.altKey || ev.isComposing || ev.key.length !== 1) return;
  if (editing) return;
  if (typeToEdit(ev.key)) ev.preventDefault();
});
// typeToEdit appends one typed character to the current selection: a text node's
// rich field, the selected table cell, or the generic element text field. Shared
// by the document handler (focus in the designer) and the preview message bridge
// (focus inside the preview iframe, where clicks land).
function typeToEdit(key) {
  if (!key || key.length !== 1) return false;
  if (!selPath) return false;
  var n = byPath(selPath, root.children);
  if (!n) return false;
  if (n.kind === "text" && n.runs) {
    var ed = document.querySelector("#inspector .richtext");
    if (!ed) return false;
    placeCaretAtEnd(ed);
    document.execCommand("insertText", false, key);
    ed.dispatchEvent(new Event("input", { bubbles: true }));
    return true;
  }
  // Table type-to-edit: when a cell is selected (table selection mode), typing
  // appends the keystroke to that cell's text (model cells are string matrices).
  if (selMode === "cell" && selCell && n.kind === "table" && n.cells && n.cells[selCell.r]) {
    n.cells[selCell.r][selCell.c] = (n.cells[selCell.r][selCell.c] || "") + key;
    snap("type cell " + key);
    setOut();
    autoPreview();
    renderInspector();
    return true;
  }
  var fields = n.component ? [] : (TAG_FIELDS[n.kind] || ["text"]);
  if (fields.indexOf("text") < 0) return false;
  var fi = document.getElementById("f_text");
  if (!fi) return false;
  var next = fi.value + key;
  fi.focus();
  fi.value = next;
  try { fi.setSelectionRange(next.length, next.length); } catch (e) {}
  upd("text", next);
  return true;
}
function placeCaretAtEnd(el) {
  el.focus();
  var range = document.createRange();
  range.selectNodeContents(el);
  range.collapse(false);
  var sel = window.getSelection();
  sel.removeAllRanges();
  sel.addRange(range);
}

// ---- marquee (rectangle) selection on the preview --------------------------
function toggleMarquee() {
  marqueeMode = !marqueeMode;
  document.getElementById("marquee-btn").classList.toggle("on", marqueeMode);
  var f = document.getElementById("prev");
  if (f && f.contentWindow) f.contentWindow.postMessage({ ty: "sui-marquee", on: marqueeMode }, "*");
  setStatus(marqueeMode ? "Marquee: drag on the preview to select multiple elements (click selects one)" : "Marquee selection off");
}
var guidesOn = false;
function toggleGuides() {
  guidesOn = !guidesOn;
  document.getElementById("guides-btn").classList.toggle("on", guidesOn);
  updateGuides();
}
function toggleStateFuncPanel() {
  var p = document.getElementById('sf-panel');
  if (!p) return;
  var showing = p.style.display !== 'none' && getComputedStyle(p).display !== 'none';
  p.style.display = showing ? 'none' : 'flex';
}

// makeResizable lets the user drag the bottom-right grip to resize a panel.
function makeResizable(el, grip, minW, minH){
  if(!el || !grip) return;
  var startX=0, startY=0, startW=0, startH=0;
  var minWidth = minW || 200, minHeight = minH || 110;
  grip.addEventListener('mousedown', function(e){
    e.preventDefault(); e.stopPropagation();
    startX = e.clientX; startY = e.clientY;
    startW = el.offsetWidth; startH = el.offsetHeight;
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  });
  function onMove(e){
    var w = Math.max(minWidth, startW + (e.clientX - startX));
    var h = Math.max(minHeight, startH + (e.clientY - startY));
    el.style.width = w + 'px';
    el.style.height = h + 'px';
  }
  function onUp(){
    document.removeEventListener('mousemove', onMove);
    document.removeEventListener('mouseup', onUp);
  }
}

function makeDraggable(el, handle){
  if(!el || !handle) return;
  var pos1=0,pos2=0,pos3=0,pos4=0;
  handle.onmousedown = dragMouseDown;
  function dragMouseDown(e){
    if (e.target && e.target.closest && e.target.closest("button,input,select,textarea,a")) return;
    e.preventDefault();
    pos3 = e.clientX; pos4 = e.clientY;
    document.onmouseup = closeDragElement;
    document.onmousemove = elementDrag;
  }
  function elementDrag(e){
    e.preventDefault();
    pos1 = pos3 - e.clientX;
    pos2 = pos4 - e.clientY;
    pos3 = e.clientX;
    pos4 = e.clientY;
    el.style.top = (el.offsetTop - pos2) + 'px';
    el.style.left = (el.offsetLeft - pos1) + 'px';
    el.style.bottom = 'auto'; el.style.right = 'auto';
  }
  function closeDragElement(){ document.onmouseup = null; document.onmousemove = null; }
}

function updateGuides() {
  var step = parseInt(document.getElementById("grid-step").value, 10) || 8;
  var f = document.getElementById("prev");
  if (f && f.contentWindow) {
    f.contentWindow.postMessage({ ty: "sui-guides", on: guidesOn, step: step }, "*");
  }
}
function idsToPaths(ids) {
  var out = [];
  ids.forEach(function(id) {
    var t = pickMap[id];
    if (!t || !t.p || t.pg !== root.page) return;
    var p = t.p.slice();
    if (!out.some(function(q){ return samePath(q, p); })) out.push(p);
  });
  return out;
}
function selectPaths(paths) {
  selCell = null; selRow = null; selCol = null; selBody = false;
  if (!paths.length) { multiSel = []; selPath = null; }
  else { multiSel = paths.map(function(p){ return p.slice(); }); selPath = paths[paths.length - 1].slice(); }
  renderTree(); fillPanel(); renderInspector(); postHighlight();
}

// ---- palette + preferences ------------------------------------------------
var dragKind = null;   // element kind being dragged from the palette
// Compact SVG glyphs for the layout toolbar.
var PAL_ICON = {
  newpage: '<path d="M7 4h7l4 4v11a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1z"/><polyline points="13,4 13,8 17,8"/><line x1="12" y1="11.5" x2="12" y2="17"/><line x1="9.2" y1="14.2" x2="14.8" y2="14.2"/>',
  div: '<rect x="4" y="5" width="16" height="14" rx="1.5"/>',
  header: '<rect x="4" y="5" width="16" height="14" rx="1.5"/><rect x="4" y="5" width="16" height="4.5" rx="1.5" fill="currentColor" stroke="none" opacity="0.55"/>',
  footer: '<rect x="4" y="5" width="16" height="14" rx="1.5"/><rect x="4" y="14.5" width="16" height="4.5" rx="1.5" fill="currentColor" stroke="none" opacity="0.55"/>',
  section: '<rect x="4" y="5" width="16" height="14" rx="1.5"/><line x1="4" y1="12" x2="20" y2="12"/><rect x="4" y="10.5" width="6" height="3" rx="1" fill="currentColor" stroke="none" opacity="0.55"/>',
  row: '<rect x="3" y="7" width="8" height="10" rx="1"/><rect x="13" y="7" width="8" height="10" rx="1"/>',
  col: '<rect x="7" y="3" width="10" height="8" rx="1"/><rect x="7" y="13" width="10" height="8" rx="1"/>',
  spacer: '<line x1="12" y1="7" x2="12" y2="17"/><path d="M8 10l4-4 4 4"/><path d="M8 14l4 4 4-4"/>',
  text: '<path d="M5 7h14M12 7v11"/>',
  textarea: '<rect x="4" y="5" width="16" height="14" rx="1.5"/><line x1="7" y1="9" x2="17" y2="9"/><line x1="7" y1="12" x2="17" y2="12"/><line x1="7" y1="15" x2="13" y2="15"/>',
  input: '<rect x="3" y="9" width="18" height="6" rx="3"/><line x1="6.5" y1="12" x2="10" y2="12"/>',
  list: '<line x1="9" y1="7" x2="19" y2="7"/><line x1="9" y1="12" x2="19" y2="12"/><line x1="9" y1="17" x2="19" y2="17"/><circle cx="5.5" cy="7" r="1"/><circle cx="5.5" cy="12" r="1"/><circle cx="5.5" cy="17" r="1"/>',
  ul: '<line x1="9" y1="7" x2="19" y2="7"/><line x1="9" y1="12" x2="19" y2="12"/><line x1="9" y1="17" x2="19" y2="17"/><circle cx="5.5" cy="7" r="1"/><circle cx="5.5" cy="12" r="1"/><circle cx="5.5" cy="17" r="1"/>',
  ol: '<line x1="9" y1="7" x2="19" y2="7"/><line x1="9" y1="12" x2="19" y2="12"/><line x1="9" y1="17" x2="19" y2="17"/><text x="3.5" y="9.2" font-size="6" fill="currentColor" stroke="none">1</text><text x="3.5" y="14.2" font-size="6" fill="currentColor" stroke="none">2</text><text x="3.5" y="19.2" font-size="6" fill="currentColor" stroke="none">3</text>',
  table: '<rect x="4" y="5" width="16" height="14" rx="1"/><line x1="4" y1="10" x2="20" y2="10"/><line x1="4" y1="14" x2="20" y2="14"/><line x1="12" y1="5" x2="12" y2="19"/>',
  img: '<rect x="3" y="5" width="18" height="14" rx="1.5"/><circle cx="8.5" cy="10" r="1.5"/><path d="M4 17l5-5 4 4 3-3 4 4"/>',
  video: '<rect x="3" y="6" width="18" height="12" rx="1.5"/><path d="M10 9.5l5 2.5-5 2.5z"/>',
  audio: '<path d="M4 9v6h4l5 4V5L8 9H4z"/><path d="M16 9a3 3 0 0 1 0 6"/>',
  iframe: '<rect x="3" y="5" width="18" height="14" rx="1.5"/><path d="M9 9l-2 3 2 3M15 9l2 3-2 3"/>',
  hr: '<line x1="4" y1="12" x2="20" y2="12"/><line x1="4.5" y1="9.5" x2="4.5" y2="14.5"/><line x1="19.5" y1="9.5" x2="19.5" y2="14.5"/>',
  br: '<path d="M8 5v9a3 3 0 0 0 3 3h6"/><path d="M15 14l3 3-3 3"/>',
  button: '<rect x="4" y="8" width="16" height="8" rx="2"/><line x1="8" y1="12" x2="16" y2="12"/>',
  a: '<path d="M10 14a3 3 0 0 1 0-6h4a3 3 0 0 1 0 6h-1"/><path d="M14 10a3 3 0 0 1 0 6h-4a3 3 0 0 1 0-6h1"/>',
  checkbox: '<rect x="4.5" y="5.5" width="13" height="13" rx="2.5" stroke-width="2"/><path d="M8 12.5l2.5 2.5 5.5-6" stroke-width="2"/>',
  radio: '<circle cx="12" cy="12" r="7" stroke-width="2"/><circle cx="12" cy="12" r="3" fill="currentColor" stroke="none"/>',
  switch: '<rect x="3.5" y="8" width="17" height="8" rx="4" stroke-width="2"/><circle cx="16" cy="12" r="3" fill="currentColor" stroke="none"/>',
  select: '<rect x="3" y="9" width="18" height="6" rx="1.5"/><path d="M15 11l2 2 2-2"/>',
  number: '<rect x="3" y="9" width="18" height="6" rx="3"/><text x="12" y="13.8" font-size="7" font-weight="700" text-anchor="middle" fill="currentColor" stroke="none">#</text>',
  slider: '<line x1="4" y1="12" x2="20" y2="12"/><circle cx="14" cy="12" r="2.5"/>',
  file: '<path d="M6 6h8l4 4v8a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1z"/><polyline points="14,6 14,10 18,10"/>',
  color: '<path d="M12 4.2c3 4 4.8 6.6 4.8 8.8a4.8 4.8 0 0 1-9.6 0C7.2 10.8 9 8.2 12 4.2z"/><circle cx="10.2" cy="13.2" r="1.1" fill="currentColor" stroke="none"/>',
  date: '<rect x="4" y="7" width="16" height="10" rx="1.5"/><line x1="4" y1="10.5" x2="20" y2="10.5"/><line x1="8" y1="5" x2="8" y2="8"/><line x1="16" y1="5" x2="16" y2="8"/>',
  time: '<circle cx="12" cy="12" r="5"/><polyline points="12,9 12,12 14,13"/>',
  card: '<rect x="4" y="5" width="16" height="14" rx="2"/><rect x="6" y="7" width="6" height="4.5" rx="1" fill="currentColor" stroke="none" opacity="0.5"/><line x1="6" y1="14" x2="18" y2="14"/><line x1="6" y1="16.5" x2="13" y2="16.5"/>',
  tabs: '<rect x="4" y="7" width="16" height="10" rx="1.5"/><line x1="4" y1="10" x2="20" y2="10"/><rect x="6" y="7" width="6" height="3" rx="1" opacity="0.6"/>',
  tab: '<rect x="4" y="7" width="16" height="10" rx="1.5"/><line x1="4" y1="10" x2="20" y2="10"/><rect x="8" y="7" width="5" height="3" rx="1" fill="currentColor" stroke="none" opacity="0.6"/>',
  modal: '<rect x="4" y="5" width="16" height="14" rx="1.5"/><line x1="4" y1="9" x2="20" y2="9"/><circle cx="17" cy="7" r="0.8" fill="currentColor" stroke="none"/><line x1="7" y1="13" x2="17" y2="13"/><line x1="7" y1="16" x2="13" y2="16"/>',
  modalbutton: '<rect x="4" y="8" width="16" height="8" rx="2"/><line x1="8" y1="12" x2="16" y2="12"/>',
  menubar: '<rect x="3" y="6" width="18" height="12" rx="1.5"/><line x1="3" y1="10" x2="21" y2="10"/><line x1="6" y1="8" x2="10" y2="8"/><line x1="13" y1="8" x2="17" y2="8"/>',
  menu: '<rect x="4" y="6" width="16" height="12" rx="1.5"/><line x1="7" y1="9" x2="17" y2="9"/><line x1="7" y1="12" x2="17" y2="12"/><line x1="7" y1="15" x2="13" y2="15"/>',
  menuitem: '<rect x="3" y="8" width="18" height="8" rx="1.5"/><line x1="7" y1="12" x2="15" y2="12"/>',
  tree: '<rect x="9" y="3" width="6" height="4" rx="1"/><rect x="3" y="16" width="5" height="4" rx="1"/><rect x="9.5" y="16" width="5" height="4" rx="1"/><rect x="16" y="16" width="5" height="4" rx="1"/><path d="M12 7v4M5.5 16v-3h13v3M12 13v3"/>',
  treeitem: '<rect x="7" y="9" width="10" height="6" rx="1.5"/><line x1="10" y1="12" x2="14" y2="12"/>',
  chart: '<path d="M4 20V4M4 20h16"/><rect x="7" y="12" width="3" height="5" rx=".5"/><rect x="12" y="8" width="3" height="9" rx=".5"/><rect x="17" y="5" width="3" height="12" rx=".5"/><path d="M6.5 7.5l4-2 4 1 5-2.5" opacity=".55"/>',
  graph: '<rect x="3" y="8" width="6" height="5" rx="1"/><rect x="15" y="4" width="6" height="5" rx="1"/><rect x="15" y="15" width="6" height="5" rx="1"/><path d="M9 10.5h3v-4h3M12 10.5v7h3"/>',
  progress: '<rect x="4" y="9" width="16" height="6" rx="3" opacity="0.35"/><rect x="4" y="9" width="10" height="6" rx="3" fill="currentColor" stroke="none"/>',
  badge: '<path d="M12.6 2.6A2 2 0 0 0 11.2 2H4a2 2 0 0 0-2 2v7.2a2 2 0 0 0 .6 1.4l8.7 8.7a2.4 2.4 0 0 0 3.4 0l6.6-6.6a2.4 2.4 0 0 0 0-3.4z"/><circle cx="7.5" cy="7.5" r="0.6" fill="currentColor" stroke="none"/>',
  alert: '<path d="M12 7l7 12H5z"/><line x1="12" y1="11" x2="12" y2="14"/><circle cx="12" cy="16" r="0.5" fill="currentColor"/>',
  spinner: '<circle cx="12" cy="12" r="7" stroke-width="3" stroke-dasharray="31 13" stroke-dashoffset="14"/><circle cx="12" cy="12" r="1.8" fill="currentColor" stroke="none"/>',
  divider: '<line x1="4" y1="12" x2="20" y2="12" stroke-dasharray="3.5 3.5"/>',
  h1: '<text x="2.5" y="18" font-size="13" font-weight="700" fill="currentColor" stroke="none">H1</text>',
  h2: '<text x="2.5" y="17.5" font-size="11.5" font-weight="700" fill="currentColor" stroke="none">H2</text>',
  h3: '<text x="2.5" y="17" font-size="10" font-weight="700" fill="currentColor" stroke="none">H3</text>',
  p: '<line x1="6" y1="9" x2="18" y2="9"/><line x1="6" y1="12" x2="18" y2="12"/><line x1="6" y1="15" x2="14" y2="15"/>',
  blockquote: '<line x1="8" y1="8" x2="8" y2="16"/><line x1="12" y1="10" x2="18" y2="10"/><line x1="12" y1="14" x2="18" y2="14"/>',
  pre: '<rect x="4" y="6" width="16" height="12" rx="1.5"/><line x1="7" y1="9" x2="10" y2="9"/><line x1="7" y1="12" x2="14" y2="12"/><line x1="7" y1="15" x2="12" y2="15"/>',
  nav: '<rect x="4" y="7" width="16" height="10" rx="1.5"/><circle cx="8" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="16" cy="12" r="1"/>',
  aside: '<rect x="4" y="6" width="16" height="12" rx="1.5"/><rect x="4" y="6" width="5" height="12" rx="1.5"/><line x1="6.5" y1="10" x2="6.5" y2="14"/>',
  form: '<rect x="4" y="6" width="16" height="12" rx="1.5"/><line x1="7" y1="9" x2="17" y2="9"/><line x1="7" y1="12" x2="17" y2="12"/><line x1="7" y1="15" x2="13" y2="15"/>',
  fieldset: '<rect x="4" y="7" width="16" height="10" rx="1.5" stroke-dasharray="2 2"/><text x="6.5" y="11" font-size="5" fill="currentColor" stroke="none">F</text>'
};
// Head icon per palette GROUP: decoupled from tags[0] so a group button can't
// pass as the wrong element (e.g. "Controls" shown as a plain rect).
var PAL_HEAD_ICON = { Page: "newpage", Layout: "div", Overlay: "modal", Text: "text", Controls: "button", Tables: "table", Media: "img" };
function renderPalette() {
  var box = document.getElementById("elementbar");
  if (!box) return;
  box.innerHTML = "";
  addBarGrip(box, "elementbar");
  PALETTE.forEach(function(g, gi) {
    if (gi > 0) { var sep = document.createElement("span"); sep.className = "esep"; box.appendChild(sep); }
    var grp = document.createElement("div");
    grp.className = "eb-group";
    var main = document.createElement("button");
    main.className = "ebtn";
    main.title = g.group;
    main.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">' + (PAL_ICON[PAL_HEAD_ICON[g.group]] || PAL_ICON.div) + '</svg>';
    var menu = document.createElement("div");
    menu.className = "eb-menu";
    main.onclick = function(e) {
      e.stopPropagation();
      document.querySelectorAll(".eb-menu").forEach(function(m){ if (m !== menu) m.classList.remove("open"); });
      menu.classList.toggle("open");
    };
    if (g.newpage) {
      var np2 = document.createElement("button"); np2.className = "eb-menu-item"; np2.innerHTML = '<span>New page\u2026</span>';
      np2.onclick = function(ev){ ev.stopPropagation(); menu.classList.remove("open"); addPageFlow(); };
      menu.appendChild(np2);
    }
    g.tags.forEach(function(t) {
      var mi = document.createElement("button");
      mi.className = "eb-menu-item";
      mi.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">' + (PAL_ICON[t] || PAL_ICON.div) + '</svg><span>' + (PALETTE_LABEL[t] || t) + '</span>';
      mi.onclick = function(ev) {
        ev.stopPropagation();
        addEl(t);
        menu.classList.remove("open");
      };
      mi.draggable = true;
      mi.ondragstart = function(ev2) { dragKind = t; ev2.dataTransfer.setData("text/plain", "palette:" + t); ev2.dataTransfer.effectAllowed = "copy"; };
      mi.ondragend = function() { dragKind = null; };
      menu.appendChild(mi);
    });
    (g.subs || []).forEach(function(sub) {
      // Submenu row: hover opens a flyout with the nested items (icons and
      // drag sources preserved).
      var mi = document.createElement("button");
      mi.className = "eb-menu-item has-sub";
      mi.innerHTML = '<span>' + sub.name + '</span><span class="sub-arrow">\u25b8</span>';
      var subPanel = document.createElement("div");
      subPanel.className = "eb-sub";
      sub.tags.forEach(function(t) {
        var sm = document.createElement("button");
        sm.className = "eb-menu-item";
        sm.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">' + (PAL_ICON[t] || PAL_ICON.div) + '</svg><span>' + (PALETTE_LABEL[t] || t) + '</span>';
        sm.onclick = function(ev) {
          ev.stopPropagation();
          addEl(t);
          menu.classList.remove("open");
        };
        sm.draggable = true;
        sm.ondragstart = function(ev2) { dragKind = t; ev2.dataTransfer.setData("text/plain", "palette:" + t); ev2.dataTransfer.effectAllowed = "copy"; };
        sm.ondragend = function() { dragKind = null; };
        subPanel.appendChild(sm);
      });
      mi.appendChild(subPanel);
      mi.addEventListener("mouseenter", function(){ subPanel.style.display = "flex"; });
      mi.addEventListener("mouseleave", function(){ subPanel.style.display = ""; });
      menu.appendChild(mi);
    });
    grp.appendChild(main);
    grp.appendChild(menu);
    box.appendChild(grp);
  });
  document.addEventListener("click", function(){ document.querySelectorAll(".eb-menu").forEach(function(m){ m.classList.remove("open"); }); });
}

// ---- movable toolbars -----------------------------------------------------
function addBarGrip(box, id) {
  var grip = document.createElement("span");
  grip.className = "bar-grip";
  grip.draggable = true;
  grip.title = "Drag to move";
  grip.innerHTML = "&#10495;";
  grip.setAttribute("data-bar", id);
  box.appendChild(grip);
  return grip;
}
var dragBar = null;
function applyBarOrder() {
  var order = (prefs.barOrder && prefs.barOrder.length) ? prefs.barOrder : PREF_DEFAULTS.barOrder;
  var cont = document.getElementById("topbars");
  if (!cont) return;
  order.forEach(function(id) {
    var bar = document.getElementById(id);
    if (bar) cont.appendChild(bar);
  });
  // any bar not listed keeps its position
  ["toolbar", "elementbar"].forEach(function(id) {
    if (order.indexOf(id) < 0) {
      var bar = document.getElementById(id);
      if (bar) cont.appendChild(bar);
    }
  });
}
function saveBarOrder() {
  var bars = [];
  document.querySelectorAll("#toolbar, #elementbar").forEach(function(b){ bars.push(b.id); });
  prefs.barOrder = bars;
  savePrefs();
}
function initDraggableBars() {
  document.addEventListener("dragstart", function(e) {
    var g = e.target.closest ? e.target.closest(".bar-grip") : null;
    if (!g) return;
    dragBar = g.getAttribute("data-bar");
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", "bar:" + dragBar);
  });
  document.addEventListener("dragend", function() {
    dragBar = null;
    document.querySelectorAll(".drop-hot").forEach(function(x){ x.classList.remove("drop-hot"); });
  });
  document.addEventListener("dragover", function(e) {
    if (!dragBar) return;
    var bar = e.target.closest ? e.target.closest("#toolbar, #elementbar") : null;
    if (!bar) return;
    e.preventDefault();
    bar.classList.add("drop-hot");
  });
  document.addEventListener("dragleave", function(e) {
    if (!dragBar) return;
    var bar = e.target.closest ? e.target.closest("#toolbar, #elementbar") : null;
    if (bar) bar.classList.remove("drop-hot");
  });
  document.addEventListener("drop", function(e) {
    if (!dragBar) return;
    var bar = e.target.closest ? e.target.closest("#toolbar, #elementbar") : null;
    if (!bar) return;
    e.preventDefault();
    bar.classList.remove("drop-hot");
    if (bar.id !== dragBar) {
      bar.parentNode.insertBefore(document.getElementById(dragBar), bar);
      saveBarOrder();
    }
    dragBar = null;
  });
}

// ---- menu bar -------------------------------------------------------------
var currentFilename = "app.sui";
var currentDir = "";   // filesystem directory of the opened/saved file ("" = unsaved)
function initMenus() {
  var menus = document.querySelectorAll('#topbar .menu');
  menus.forEach(function(m) {
    var head = m.querySelector('.mitem');
    head.onclick = function(ev) {
      ev.stopPropagation();
      var wasOpen = m.classList.contains('open');
      closeMenus();
      if (!wasOpen) m.classList.add('open');
    };
  });
  document.addEventListener('click', closeMenus);
  document.addEventListener('keydown', function(ev) {
    if (ev.key === 'Escape') closeMenus();
  });
}
function closeContextMenus() {
  var ed = document.getElementById('ed-menu'); if (ed) ed.remove();
  var nm = document.getElementById('node-menu'); if (nm) nm.remove();
}
function closeMenus() {
  document.querySelectorAll('#topbar .menu').forEach(function(m){ m.classList.remove('open'); });
  closeSnippetMenu();
  closeContextMenus();
}
// Context menus close on any pointer press outside their own panel. Capture
// phase is deliberate: many toolbar/buttons stop regular click propagation,
// which used to leave the editor/tree menu on screen after clicking elsewhere.
(function wireContextMenus() {
  document.addEventListener("pointerdown", function(ev) {
    if (!document.getElementById("ed-menu") && !document.getElementById("node-menu")) return;
    var t = ev.target;
    if (t && t.closest && (t.closest("#ed-menu") || t.closest("#node-menu"))) return;
    closeContextMenus();
  }, true);
  document.addEventListener("keydown", function(ev) {
    if (ev.key === "Escape") closeContextMenus();
  });
  window.addEventListener("blur", closeContextMenus);
})();
function renderInsertMenu() {
  var drop = document.getElementById('drop-insert');
  drop.innerHTML = '';
  // The top level shows only GROUP rows; hovering a row opens its flyout to
  // the right. Sub-groups (Navigation, Lists, Feedback) cascade the same way
  // one level deeper.
  function addItem(panel, t) {
    var mi = document.createElement('div'); mi.className = 'mi'; mi.textContent = PALETTE_LABEL[t] || t;
    mi.onclick = function(){ addEl(t); closeMenus(); };
    panel.appendChild(mi);
  }
  function addSubRow(panel, label, build) {
    var row = document.createElement('div'); row.className = 'mi has-sub';
    var lab = document.createElement('span'); lab.textContent = label;
    var arrow = document.createElement('span'); arrow.className = 'sub-arrow'; arrow.textContent = '\u25b8';
    row.appendChild(lab); row.appendChild(arrow);
    var sub = document.createElement('div'); sub.className = 'mi-sub';
    build(sub);
    row.appendChild(sub);
    // Open/close via JS as well, so the flyout also reacts to synthetic
    // hover (CSS :hover alone is not always honoured: zoom, touch).
    row.addEventListener("mouseenter", function(){ sub.style.display = "block"; });
    row.addEventListener("mouseleave", function(){ sub.style.display = ""; });
    panel.appendChild(row);
  }
  PALETTE.forEach(function(g) {
    addSubRow(drop, g.group, function(sub) {
      if (g.newpage) {
        var np = document.createElement('div'); np.className = 'mi'; np.textContent = 'New page\u2026';
        np.onclick = function(){ closeMenus(); addPageFlow(); };
        sub.appendChild(np);
        return;
      }
      g.tags.forEach(function(t) { addItem(sub, t); });
      (g.subs || []).forEach(function(sg) {
        addSubRow(sub, sg.name, function(sub2) {
          sg.tags.forEach(function(t) { addItem(sub2, t); });
        });
      });
    });
  });
}
function newSui() {
  closeMenus();
  root = pagesModelEmpty();
  selPath = null; selCell = null; selRow = null; selCol = null; selBody = false; selState = "";
  past = []; future = [];
  setOut();
  renderAll();
  autoPreview();
  currentFilename = "app.sui";
  currentDir = "";
  setStatus("new SUI");
}
function saveSui() {
  closeMenus();
  if (currentDir) {
    fsWriteFile(currentDir, currentFilename, suiSource || generate());
  } else {
    openFsDialog("save");
  }
}
function saveSuiAs() {
  closeMenus();
  openFsDialog("save");
}
function fsWriteFile(dir, name, content) {
  var fd = new URLSearchParams();
  fd.set("path", dir);
  fd.set("name", name);
  fd.set("content", content);
  fetch("/__gui__/fs/save", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) { setStatus("save failed: " + (j.error || "")); return; }
    currentDir = dir;
    currentFilename = name;
    setStatus("saved " + j.path);
  });
}

// ---- filesystem browser dialog --------------------------------------------
var fsMode = "";
var fsPath = "";
var fsParent = "";
function openFsDialog(mode) {
  closeMenus();
  fsMode = mode;
  document.getElementById("fstitle").textContent = mode === "open" ? "Open project" : "Save project";
  var foot = document.getElementById("fsfoot");
  foot.innerHTML = "";
  if (mode === "save") {
    foot.innerHTML = '<input id="fsname" placeholder="name.sui" value="' + currentFilename + '"><button class="btn" onclick="fsSave()">Save</button>';
  } else {
    var h = document.createElement("div"); h.className = "hint";
    h.textContent = "Choose a .sui file to open.";
    foot.appendChild(h);
  }
  document.getElementById("fsdialog").classList.add("on");
  fsLoad("");
}
function closeFsDialog() { document.getElementById("fsdialog").classList.remove("on"); }
function fsLoad(path) {
  fsPath = path || "";
  fetch("/__gui__/fs/list?path=" + encodeURIComponent(fsPath)).then(function(r){ return r.json(); }).then(function(j){
    fsParent = j.parent || "";
    document.getElementById("fspath").value = j.path || "";
    renderFsEntries(j.entries || [], j.error || "");
  });
}
function renderFsEntries(entries, err) {
  var box = document.getElementById("fslist");
  box.innerHTML = "";
  if (err) { var e = document.createElement("div"); e.className = "hint"; e.textContent = err; box.appendChild(e); return; }
  if (!entries.length) { var h = document.createElement("div"); h.className = "hint"; h.textContent = "(empty)"; box.appendChild(h); return; }
  entries.forEach(function(en){
    var d = document.createElement("div");
    d.className = "fsentry" + (en.dir ? " dir" : "");
    d.innerHTML = '<span class="ficon">' + (en.dir ? "&#128193;" : "&#128196;") + '</span>' + escHtml(en.name);
    d.title = en.path;
    d.onclick = function(){ if (en.dir) fsLoad(en.path); else fsSelect(en); };
    box.appendChild(d);
  });
}
function fsSelect(en) {
  if (fsMode === "open") {
    fetch("/__gui__/fs/open?path=" + encodeURIComponent(en.path)).then(function(r){ return r.json(); }).then(function(j){
      if (!j.ok) { setStatus("open failed: " + (j.error || "")); return; }
      var fd = new URLSearchParams(); fd.set("sui", j.content);
      fetch("/__gui__/open", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(m){
        if (!m.ok) { setStatus((m.diagnostics || []).join(" | ")); return; }
        loadTree(m);
        currentDir = fsPath;
        currentFilename = en.name;
        closeFsDialog();
        setStatus("opened " + en.path);
      });
    });
  } else {
    var nameInput = document.getElementById("fsname");
    if (nameInput) nameInput.value = en.name;
  }
}
function fsUp() { fsLoad(fsParent); }
function fsGo() { fsLoad(document.getElementById("fspath").value); }
function fsSave() {
  var name = (document.getElementById("fsname").value || "").trim();
  if (!name) { setStatus("enter a file name"); return; }
  if (!/\.sui$/.test(name)) name += ".sui";
  currentFilename = name;
  fsWriteFile(fsPath, name, suiSource || generate());
  closeFsDialog();
}

// ---- export build (.zip) --------------------------------------------------
function exportZip() {
  closeMenus();
  setStatus("building…");
  var fd = new URLSearchParams(); fd.set("sui", suiSource || generate());
  fetch("/__gui__/export", { method: "POST", body: fd }).then(function(r){
    var ct = r.headers.get("Content-Type") || "";
    if (ct.indexOf("application/zip") < 0) {
      return r.json().then(function(j){ setStatus((j.diagnostics || ["export failed"]).join(" | ")); });
    }
    return r.blob().then(function(b){
      var a = document.createElement("a");
      a.href = URL.createObjectURL(b);
      a.download = "simpleui-build.zip";
      a.click();
      setStatus("exported simpleui-build.zip");
    });
  }).catch(function(){ setStatus("export failed"); });
}

function exportSingleHtml() {
  closeMenus();
  setStatus("building single HTML…");
  var fd = new URLSearchParams(); fd.set("sui", suiSource || generate());
  fetch("/__gui__/export-single", { method: "POST", body: fd }).then(function(r){
    var ct = r.headers.get("Content-Type") || "";
    if (ct.indexOf("text/html") < 0) {
      return r.json().then(function(j){ setStatus((j.diagnostics || ["export failed"]).join(" | ")); });
    }
    return r.blob().then(function(b){
      var a = document.createElement("a");
      a.href = URL.createObjectURL(b);
      a.download = "app.html";
      a.click();
      setStatus("exported app.html");
    });
  }).catch(function(){ setStatus("export failed"); });
}

// ---- desktop executable build --------------------------------------------
var deskBuilding = false;
function deskStatus(text, cls) {
  var el = document.getElementById("desk-status");
  el.textContent = text || "";
  el.className = cls || "";
}
function openDesktopBuild() {
  if (deskBuilding) return;
  closeMenus();
  deskStatus("Checking build prerequisites…", "");
  document.getElementById("desk-log").textContent = "";
  var dl = document.getElementById("desk-dl");
  dl.style.display = "none";
  dl.removeAttribute("href");
  fetch("/__gui__/desktop-check").then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) {
      showMsg("Desktop build unavailable", j.error || "The desktop build prerequisites are not satisfied.");
      return;
    }
    document.getElementById("desk-env").textContent = "Go " + j.goVersion + " | " + j.goos + "/" + j.goarch + " | output is local to the SimpleUI server";
    deskStatus("Ready. Preflight will run again before building.", "");
    deskBuilding = false;
    document.getElementById("desk-build").disabled = false;
    document.getElementById("desk-cancel").disabled = false;
    document.getElementById("deskdlg").classList.add("on");
    document.getElementById("desk-out").focus();
  }).catch(function(){
    showMsg("Desktop build unavailable", "Cannot reach the editor server to check the Go toolchain.");
  });
}
function closeDesktopBuild() {
  if (deskBuilding) {
    deskStatus("Build in progress: wait for it to finish.", "err");
    return;
  }
  document.getElementById("deskdlg").classList.remove("on");
}
function startDesktopBuild() {
  if (deskBuilding) return;
  var out = (document.getElementById("desk-out").value || "").trim();
  if (!out) {
    deskStatus("Output directory is required.", "err");
    return;
  }
  var log = document.getElementById("desk-log");
  var dl = document.getElementById("desk-dl");
  dl.style.display = "none";
  dl.removeAttribute("href");
  deskBuilding = true;
  document.getElementById("desk-build").disabled = true;
  document.getElementById("desk-cancel").disabled = true;
  deskStatus("Preflight and build in progress…", "");
  log.textContent = "Preflight: checking source, Go toolchain, output directory and capabilities…\n";
  var fd = new URLSearchParams();
  fd.set("sui", suiSource || generate());
  fd.set("out", out);
  fd.set("allowDirs", document.getElementById("desk-dirs").value || "");
  fd.set("allowExec", document.getElementById("desk-exec").value || "");
  fetch("/__gui__/build-desktop", { method: "POST", body: fd }).then(function(r){
    return r.json();
  }).then(function(j){
    deskBuilding = false;
    document.getElementById("desk-build").disabled = false;
    document.getElementById("desk-cancel").disabled = false;
    log.textContent = j.log || "";
    if (j.diagnostics && j.diagnostics.length) {
      log.textContent += "\n";
      j.diagnostics.forEach(function(d){
        var at = d.line ? d.line + ":" + (d.col || 1) + ": " : "";
        log.textContent += at + (d.msg || String(d)) + "\n";
      });
    }
    if (!j.ok) {
      deskStatus(j.error || "Desktop build failed.", "err");
      setStatus("desktop build failed");
      return;
    }
    deskStatus("Built: " + j.path, "ok");
    setStatus("desktop build ok: " + j.path);
    if (j.downloadUrl) {
      dl.href = j.downloadUrl;
      dl.setAttribute("download", j.file || "app.exe");
      dl.style.display = "inline-flex";
      if (document.getElementById("desk-download").checked) {
        dl.click();
      }
    }
  }).catch(function(){
    deskBuilding = false;
    document.getElementById("desk-build").disabled = false;
    document.getElementById("desk-cancel").disabled = false;
    deskStatus("Build request failed: the editor server is unreachable.", "err");
    log.textContent += "Build request failed.\n";
  });
}
(function wireDesktopDialog(){
  var dlg = document.getElementById("deskdlg");
  document.getElementById("desk-cancel").onclick = closeDesktopBuild;
  document.getElementById("desk-build").onclick = startDesktopBuild;
  dlg.addEventListener("mousedown", function(ev){ if (ev.target === dlg) closeDesktopBuild(); });
  dlg.addEventListener("keydown", function(ev){
    if (ev.key === "Escape") { ev.preventDefault(); closeDesktopBuild(); }
  });
})();

// ---- project history ------------------------------------------------------
function openHistory() {
  closeMenus();
  document.getElementById("hist-editor").classList.add("on");
  renderHistoryList();
}
function closeHistory() { document.getElementById("hist-editor").classList.remove("on"); }
function renderHistoryList() {
  var box = document.getElementById("hist-list");
  box.innerHTML = "";
  fetch("/__gui__/history").then(function(r){ return r.json(); }).then(function(j){
    var vs = j.versions || [];
    if (!vs.length) { var h = document.createElement("div"); h.className = "hint"; h.textContent = "No snapshots yet."; box.appendChild(h); return; }
    vs.forEach(function(name){
      var row = document.createElement("div"); row.className = "histv";
      var nm = document.createElement("span"); nm.className = "histvname"; nm.textContent = name;
      nm.title = "Click to restore";
      nm.onclick = function(){ restoreVersion(name); };
      row.appendChild(nm);
      box.appendChild(row);
    });
  }).catch(function(){});
}
function snapshotNow() {
  var fd = new URLSearchParams(); fd.set("sui", suiSource || generate());
  fetch("/__gui__/snapshot", { method: "POST", body: fd }).then(function(){ renderHistoryList(); setStatus("snapshot saved"); });
}
function restoreVersion(name) {
  fetch("/__gui__/restore?file=" + encodeURIComponent(name)).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok || !j.sui) { setStatus("cannot read version"); return; }
    var fd = new URLSearchParams(); fd.set("sui", j.sui);
    fetch("/__gui__/open", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(m){
      if (!m.ok) { setStatus((m.diagnostics || []).join(" | ")); return; }
      loadTree(m); autoPreview();
      closeHistory();
      setStatus("restored " + name);
    });
  });
}

// ---- reusable templates ---------------------------------------------------
// A saved block must be self-contained: the snippet carries the CSS rules of
// every class it uses, otherwise re-inserting it in another project (or after
// deleting the original) yields an empty, zero-height, invisible element.
function collectClasses(n, acc) {
  String(n.class || "").trim().split(/\s+/).forEach(function(c){ if (c) acc[c] = 1; });
  (n.children || []).forEach(function(c){ collectClasses(c, acc); });
}
function cssRuleToSuiChunk(rule) {
  var body = rule.decls.map(function(d){ return "  " + d.prop + ": " + d.value + ";\n"; }).join("");
  var sel = rule.selectors.join(", ");
  return "css \"" + esc(sel + " {\n" + body + "}") + "\"\n";
}
// unescapeSuiStr is the inverse of esc() for css chunks: braces first, then
// newline, quote, backslash last.
function unescapeSuiStr(s) {
  return s.replace(/\\([{}])/g, "$1").replace(/\\n/g, "\n").replace(/\\"/g, "\"").replace(/\\\\/g, "\\");
}
function templateCssChunks(t) {
  var re = /^css "((?:\\.|[^"\\])*)"$/mg;
  var out = [], m;
  while ((m = re.exec(t)) !== null) out.push(unescapeSuiStr(m[1]));
  return out;
}
function renderTemplates() {
  var box = document.getElementById("tpl-list"); if (!box) return;
  box.innerHTML = "";
  fetch("/__gui__/templates").then(function(r){ return r.json(); }).then(function(j){
    (j.templates || []).forEach(function(t) {
      if (t.kind === "page") return; // page templates live in the new-page dialog
      var name = t.name;
      var wrap = document.createElement("span"); wrap.className = "tplitem";
      var b = document.createElement("button"); b.className = "btn"; b.textContent = name;
      b.title = "Insert block";
      b.onclick = function(){ insertTemplate(name); };
      wrap.appendChild(b);
      var d = document.createElement("button"); d.className = "btn"; d.textContent = "\u2212"; d.title = "Delete template";
      d.onclick = function(){ deleteTemplate(name); };
      wrap.appendChild(d);
      box.appendChild(wrap);
    });
  }).catch(function(){});
}
function saveBlockFlow() {
  var paths = selectedPaths();
  if (!paths.length) { setStatus("select an element first"); return; }
  var nodes = paths.map(function(p){ return byPath(p, root.children); });
  var snippet = nodes.map(function(n){ return toSui(n, 1); }).join("");
  // Carry the CSS rules of every class the snippet uses: without them the
  // block re-inserts as an invisible zero-height element.
  var clsSet = {};
  nodes.forEach(function(n){ collectClasses(n, clsSet); });
  Object.keys(clsSet).forEach(function(cls){
    var r = findRule(cls);
    if (r) snippet = cssRuleToSuiChunk(r) + snippet;
  });
  askPrompt("Save as block", "Block name", "My block", function(name){
    var fd = new URLSearchParams(); fd.set("name", name); fd.set("kind", "block"); fd.set("sui", snippet);
    fetch("/__gui__/template/save", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
      if (!j.ok) { setStatus("save failed: " + (j.error || "")); return; }
      renderTemplates();
      setStatus("saved block: " + name);
    });
  }, "Blocks are saved in the library and stay available in every project.");
}
function deleteTemplate(name) {
  confirmBox("Delete the block?", "\"" + name + "\" will be removed from the library (all projects).", function(){
    var fd = new URLSearchParams(); fd.set("name", name);
    fetch("/__gui__/template/delete", { method: "POST", body: fd }).then(function(){ renderTemplates(); });
  });
}
function insertTemplate(name) {
  fetch("/__gui__/template?name=" + encodeURIComponent(name)).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) { setStatus("template not found"); return; }
    // Validate the snippet INSIDE the live project: a saved template routinely
    // references project states/functions (value: name, onClick: save), so
    // checking it in isolation always fails with "unknown variable".
    var fd = new URLSearchParams(); fd.set("sui", j.sui); fd.set("project", typeof suiSource === "string" ? suiSource : "");
    fetch("/__gui__/template/open", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(m){
      if (!m.ok) { showMsg("Template cannot be inserted", "The template is not valid in the current project:\n\n" + ((m.diagnostics || []).map(function(d){ return "L" + d.line + ": " + d.msg + (d.hint ? " — " + d.hint : ""); }).join("\n") || "invalid template")); return; }
      var els = m.elements || [];
      if (!els.length) { setStatus("template is empty"); return; }
      // Merge the template's CSS into the project, remapping the block's
      // classes that would collide with existing rules/local classes so an
      // inserted block can never restyle unrelated elements.
      var map = {};
      var clsSet = {};
      els.forEach(function(e){ collectClasses(e, clsSet); });
      var used = {};
      usedClassesList().forEach(function(c){ used[c] = 1; });
      var takens = {};
      Object.keys(clsSet).forEach(function(cls){
        if (!used[cls] && !findRule(cls)) return;
        var nm = cls, k = 2;
        while (findRule(nm) || used[nm] || takens[nm]) { nm = cls + "-" + k; k++; }
        map[cls] = nm; takens[nm] = 1;
      });
      function remapNode(n) {
        var toks = String(n.class || "").trim().split(/\s+/).filter(function(t){ return t; });
        if (toks.length) n.class = toks.map(function(t){ return map[t] || t; }).join(" ");
        (n.children || []).forEach(remapNode);
      }
      var cssBlocks = templateCssChunks(j.sui).map(function(block){
        return block.replace(/\.([A-Za-z_][\w-]*)/g, function(m0, t){ return map[t] ? "." + map[t] : m0; });
      });
      withUndo(function(){
        els.forEach(function(e){ remapNode(e); });
        if (cssBlocks.length) root.css = (root.css ? root.css + "\n" : "") + cssBlocks.join("\n").trim();
        if (selPath) {
          var sel = byPath(selPath, root.children);
          if (canNest(sel.kind)) {
            els.forEach(function(e){ sel.children.push(e); });
            selPath = selPath.concat([sel.children.length - 1]);
            return;
          }
          var par = selPath.slice(0, -1);
          var arr = par.length === 0 ? root.children : byPath(par, root.children).children;
          var pos = selPath[selPath.length - 1] + 1;
          els.forEach(function(e, i){ arr.splice(pos + i, 0, e); });
          selPath = par.concat([pos]);
        } else {
          els.forEach(function(e){ root.children.push(e); });
          selPath = [root.children.length - 1];
        }
      }, "insert template " + name);
    });
  });
}
function deleteSelected() {
  closeMenus();
  if (multiSel.length) { deletePaths(multiSel); return; }
  if (selPath) {
    delSelPath(selPath.join(","));
    return;
  }
  setStatus("nothing selected");
}
function duplicateSelected() {
  closeMenus();
  if (!selPath) { setStatus("nothing selected"); return; }
  withUndo(function() {
    var parentPath = selPath.slice(0, -1);
    var idx = selPath[selPath.length - 1];
    var clone = JSON.parse(JSON.stringify(byPath(selPath, root.children)));
    byPath(parentPath, root.children).children.splice(idx + 1, 0, clone);
    selPath = parentPath.concat([idx + 1]);
    setStatus("duplicated " + (clone.kind || "element"));
  }, "duplicate");
}


function savePrefs() {
  try {
    var fd = new URLSearchParams();
    fd.set("prefs", JSON.stringify(prefs));
    fetch("/__gui__/prefs", { method: "POST", body: fd });
  } catch (e) {}
}
function loadRemotePrefs() {
  fetch("/__gui__/prefs").then(function(r){ return r.json(); }).then(function(p){
    if (!p || typeof p !== "object") return;
    prefs.theme = p.theme || PREF_DEFAULTS.theme;
    prefs.font = p.font || PREF_DEFAULTS.font;
    prefs.density = p.density || PREF_DEFAULTS.density;
    prefs.accent = p.accent || PREF_DEFAULTS.accent;
    prefs.showLeft = p.showLeft !== undefined ? p.showLeft : PREF_DEFAULTS.showLeft;
    prefs.showRight = p.showRight !== undefined ? p.showRight : PREF_DEFAULTS.showRight;
    prefs.showBottom = p.showBottom !== undefined ? p.showBottom : PREF_DEFAULTS.showBottom;
    prefs.barOrder = (p.barOrder && p.barOrder.length) ? p.barOrder : PREF_DEFAULTS.barOrder;
    prefs.editorFont = p.editorFont || PREF_DEFAULTS.editorFont;
    prefs.editorSize = p.editorSize || PREF_DEFAULTS.editorSize;
    prefs.editorLineHeight = p.editorLineHeight || PREF_DEFAULTS.editorLineHeight;
    prefs.editorTabSize = p.editorTabSize || PREF_DEFAULTS.editorTabSize;
    prefs.editorWrap = p.editorWrap !== undefined ? p.editorWrap : PREF_DEFAULTS.editorWrap;
    prefs.editorLineNumbers = p.editorLineNumbers !== undefined ? p.editorLineNumbers : PREF_DEFAULTS.editorLineNumbers;
    prefs.editorSyntaxColors = p.editorSyntaxColors !== undefined ? p.editorSyntaxColors : PREF_DEFAULTS.editorSyntaxColors;
    prefs.editorBracketMatch = p.editorBracketMatch !== undefined ? p.editorBracketMatch : PREF_DEFAULTS.editorBracketMatch;
    prefs.editorSelectedLine = p.editorSelectedLine !== undefined ? p.editorSelectedLine : PREF_DEFAULTS.editorSelectedLine;
    applyEditorPrefs();
    prefs.palette = (p.palette && p.palette.length) ? p.palette.slice() : [];
    palColors = prefs.palette;
    renderPalStrips();
    applyTheme();
    applyPanelVisibility();
    applyBarOrder();
  }).catch(function(){});
}

var chartTooltipNode = null;
var CHART_TOOLTIP_VARS = [
  "--sui-chart-tooltip-bg", "--sui-chart-tooltip-fg", "--sui-chart-tooltip-border",
  "--sui-chart-tooltip-radius", "--sui-chart-tooltip-max-width",
  "--sui-chart-tooltip-font-size", "--sui-chart-tooltip-delay"
];
function chartStyleCss(n) {
  var s = n && n.styleExpr ? String(n.styleExpr).trim() : "";
  if (!s) return "";
  var m = s.match(/^cssSafe\s*\(\s*("(?:[^"\\]|\\.)*")\s*\)$/);
  if (!m) return null;
  try { return JSON.parse(m[1]); } catch (e) { return null; }
}
function chartTooltipCssValue(n, name, fallback) {
  var css = chartStyleCss(n);
  if (css === null) return fallback;
  var re = new RegExp(name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&") + "\\s*:\\s*([^;]+);");
  var m = String(css || "").match(re);
  return m ? m[1].trim() : fallback;
}
function chartTooltipColor(n, name, fallback) {
  var v = chartTooltipCssValue(n, name, fallback);
  return /^#[0-9a-f]{6}$/i.test(v) ? v : fallback;
}
function chartTooltipLength(n, name, fallback) {
  var v = chartTooltipCssValue(n, name, fallback).trim();
  return /^(0|(?:\d*\.)?\d+(?:px|rem|em|%))$/.test(v) ? v : fallback;
}
function updateChartTooltipPreview() {
  var p = document.getElementById("ctt-preview");
  if (!p) return;
  var bg = document.getElementById("ctt-bg").value || "#111827";
  var fg = document.getElementById("ctt-fg").value || "#f8fafc";
  var border = document.getElementById("ctt-border").value || "#475569";
  var radius = normLen(document.getElementById("ctt-radius").value) || "6px";
  var maxWidth = normLen(document.getElementById("ctt-maxwidth").value) || "240px";
  var fontSize = normLen(document.getElementById("ctt-fontsize").value) || "11px";
  p.style.background = bg;
  p.style.color = fg;
  p.style.borderColor = border;
  p.style.borderRadius = radius;
  p.style.maxWidth = maxWidth;
  p.style.fontSize = fontSize;
}
function openChartTooltipCfg(n) {
  chartTooltipNode = n;
  document.getElementById("ctt-bg").value = chartTooltipColor(n, "--sui-chart-tooltip-bg", "#111827");
  document.getElementById("ctt-fg").value = chartTooltipColor(n, "--sui-chart-tooltip-fg", "#f8fafc");
  document.getElementById("ctt-border").value = chartTooltipColor(n, "--sui-chart-tooltip-border", "#475569");
  document.getElementById("ctt-radius").value = chartTooltipLength(n, "--sui-chart-tooltip-radius", "6px");
  document.getElementById("ctt-maxwidth").value = chartTooltipLength(n, "--sui-chart-tooltip-max-width", "240px");
  document.getElementById("ctt-fontsize").value = chartTooltipLength(n, "--sui-chart-tooltip-font-size", "11px");
  var delay = parseInt(chartTooltipCssValue(n, "--sui-chart-tooltip-delay", "0ms"), 10);
  document.getElementById("ctt-delay").value = isNaN(delay) ? 0 : Math.max(0, Math.min(5000, delay));
  updateChartTooltipPreview();
  document.getElementById("cttcfg").classList.add("on");
}
function closeChartTooltipCfg() {
  document.getElementById("cttcfg").classList.remove("on");
  chartTooltipNode = null;
}
function chartTooltipCssWithoutVars(css) {
  var out = [];
  String(css || "").split(";").forEach(function(part) {
    var t = part.trim();
    if (!t) return;
    for (var i = 0; i < CHART_TOOLTIP_VARS.length; i++) {
      if (t.indexOf(CHART_TOOLTIP_VARS[i] + ":") === 0) return;
    }
    out.push(t);
  });
  return out;
}
function applyChartTooltipCfg() {
  var n = chartTooltipNode;
  if (!n) return;
  var css = chartStyleCss(n);
  if (css === null) {
    alert("This chart uses a custom style expression. Edit it manually in Appearance.");
    return;
  }
  var bg = document.getElementById("ctt-bg").value || "#111827";
  var fg = document.getElementById("ctt-fg").value || "#f8fafc";
  var border = document.getElementById("ctt-border").value || "#475569";
  var radius = normLen(document.getElementById("ctt-radius").value) || "6px";
  var maxWidth = normLen(document.getElementById("ctt-maxwidth").value) || "240px";
  var fontSize = normLen(document.getElementById("ctt-fontsize").value) || "11px";
  var delay = parseInt(document.getElementById("ctt-delay").value, 10);
  if (isNaN(delay) || delay < 0) delay = 0;
  if (delay > 5000) delay = 5000;
  var parts = chartTooltipCssWithoutVars(css);
  parts.push("--sui-chart-tooltip-bg: " + bg);
  parts.push("--sui-chart-tooltip-fg: " + fg);
  parts.push("--sui-chart-tooltip-border: " + border);
  parts.push("--sui-chart-tooltip-radius: " + radius);
  parts.push("--sui-chart-tooltip-max-width: " + maxWidth);
  parts.push("--sui-chart-tooltip-font-size: " + fontSize);
  parts.push("--sui-chart-tooltip-delay: " + delay + "ms");
  snap("chart tooltip settings");
  n.styleExpr = "cssSafe(" + JSON.stringify(parts.join("; ")) + ")";
  closeChartTooltipCfg();
  setOut();
  autoPreview();
  renderInspector();
}
function resetChartTooltipCfg() {
  var n = chartTooltipNode;
  if (!n) return;
  var css = chartStyleCss(n);
  if (css === null) {
    alert("This chart uses a custom style expression. Edit it manually in Appearance.");
    return;
  }
  var parts = chartTooltipCssWithoutVars(css);
  snap("reset chart tooltip settings");
  n.styleExpr = parts.length ? "cssSafe(" + JSON.stringify(parts.join("; ")) + ")" : "";
  closeChartTooltipCfg();
  setOut();
  autoPreview();
  renderInspector();
}
var chartTableNode = null;
var CHART_TABLE_VARS = [
  "--sui-chart-table-font-size", "--sui-chart-table-cell-padding",
  "--sui-chart-table-head-bg", "--sui-chart-table-stripe-opacity",
  "--sui-chart-table-caption-display", "--sui-chart-table-header-position"
];
function chartTableCssValue(n, name, fallback) {
  return chartTooltipCssValue(n, name, fallback);
}
function updateChartTablePreview() {
  var p = document.getElementById("ctable-preview");
  if (!p) return;
  var fontSize = normLen(document.getElementById("ctb-fontsize").value) || "11px";
  var padding = document.getElementById("ctb-padding").value || "4px 6px";
  var headBg = document.getElementById("ctb-headbg").value || "#14171c";
  var stripe = parseFloat(document.getElementById("ctb-stripe").value);
  if (!isFinite(stripe) || stripe < 0) stripe = 0;
  if (stripe > 1) stripe = 1;
  p.style.fontSize = fontSize;
  var hd = p.querySelector(".hd"), rw = p.querySelector(".rw");
  if (hd) { hd.style.background = headBg; hd.style.padding = padding; }
  if (rw) { rw.style.padding = padding; rw.style.background = "rgba(127,127,127," + stripe + ")"; }
}
function openChartTableCfg(n) {
  chartTableNode = n;
  document.getElementById("ctb-fontsize").value = chartTableCssValue(n, "--sui-chart-table-font-size", "11px");
  document.getElementById("ctb-padding").value = chartTableCssValue(n, "--sui-chart-table-cell-padding", "4px 6px");
  document.getElementById("ctb-headbg").value = chartTooltipColor(n, "--sui-chart-table-head-bg", "#14171c");
  var stripe = parseFloat(chartTableCssValue(n, "--sui-chart-table-stripe-opacity", "0.06"));
  document.getElementById("ctb-stripe").value = isFinite(stripe) ? stripe : 0.06;
  document.getElementById("ctb-caption").checked = chartTableCssValue(n, "--sui-chart-table-caption-display", "block") !== "none";
  document.getElementById("ctb-sticky").checked = chartTableCssValue(n, "--sui-chart-table-header-position", "sticky") !== "static";
  updateChartTablePreview();
  document.getElementById("ctablecfg").classList.add("on");
}
function closeChartTableCfg() {
  document.getElementById("ctablecfg").classList.remove("on");
  chartTableNode = null;
}
function chartTableCssWithoutVars(css) {
  var out = [];
  String(css || "").split(";").forEach(function(part) {
    var t = part.trim();
    if (!t) return;
    for (var i = 0; i < CHART_TABLE_VARS.length; i++) {
      if (t.indexOf(CHART_TABLE_VARS[i] + ":") === 0) return;
    }
    out.push(t);
  });
  return out;
}
function applyChartTableCfg() {
  var n = chartTableNode;
  if (!n) return;
  var css = chartStyleCss(n);
  if (css === null) {
    alert("This chart uses a custom style expression. Edit it manually in Appearance.");
    return;
  }
  var fontSize = normLen(document.getElementById("ctb-fontsize").value) || "11px";
  var padding = document.getElementById("ctb-padding").value.trim() || "4px 6px";
  var headBg = document.getElementById("ctb-headbg").value || "#14171c";
  var stripe = parseFloat(document.getElementById("ctb-stripe").value);
  if (!isFinite(stripe) || stripe < 0) stripe = 0;
  if (stripe > 1) stripe = 1;
  var parts = chartTableCssWithoutVars(css);
  parts.push("--sui-chart-table-font-size: " + fontSize);
  parts.push("--sui-chart-table-cell-padding: " + padding);
  parts.push("--sui-chart-table-head-bg: " + headBg);
  parts.push("--sui-chart-table-stripe-opacity: " + stripe);
  parts.push("--sui-chart-table-caption-display: " + (document.getElementById("ctb-caption").checked ? "block" : "none"));
  parts.push("--sui-chart-table-header-position: " + (document.getElementById("ctb-sticky").checked ? "sticky" : "static"));
  snap("chart table settings");
  n.styleExpr = "cssSafe(" + JSON.stringify(parts.join("; ")) + ")";
  closeChartTableCfg();
  setOut();
  autoPreview();
  renderInspector();
}
function resetChartTableCfg() {
  var n = chartTableNode;
  if (!n) return;
  var css = chartStyleCss(n);
  if (css === null) {
    alert("This chart uses a custom style expression. Edit it manually in Appearance.");
    return;
  }
  var parts = chartTableCssWithoutVars(css);
  snap("reset chart table settings");
  n.styleExpr = parts.length ? "cssSafe(" + JSON.stringify(parts.join("; ")) + ")" : "";
  closeChartTableCfg();
  setOut();
  autoPreview();
  renderInspector();
}
var chartHintsNode = null;
var chartHintsUseDefaults = false;
function chartHintsColor(value, fallback) {
  var v = String(value || "").trim();
  return v || fallback;
}
function chartHintHex(value, fallback) {
  var v = String(value || "").trim();
  if (!v || /^(transparent|none)$/i.test(v)) return fallback;
  var direct = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(v);
  var h = direct ? direct[1] : "";
  if (!h) {
    var css = cssColorToHex(v);
    var m = /^#([0-9a-f]{3}|[0-9a-f]{6})/i.exec(css);
    if (m) h = m[1];
  }
  if (!h) return fallback;
  if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
  return "#" + h.slice(0, 6).toLowerCase();
}
function chartHintsPreviewColor(value, opacity, fallback) {
  var v = String(value || "").trim();
  if (!v) return fallback;
  if (v.toLowerCase() === "transparent") return "transparent";
  var m = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(v);
  if (!m) return fallback;
  var h = m[1];
  if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
  var r = parseInt(h.slice(0, 2), 16), g = parseInt(h.slice(2, 4), 16), b = parseInt(h.slice(4, 6), 16);
  var a = Number(opacity);
  if (!isFinite(a)) a = 1;
  a = Math.max(0, Math.min(1, a));
  if (a <= 0) return "transparent";
  return a >= 1 ? "rgb(" + r + "," + g + "," + b + ")" : "rgba(" + r + "," + g + "," + b + "," + a + ")";
}
function updateChartHintsPreview() {
  var el = document.getElementById("chints-preview");
  if (!el) return;
  var opacityRaw = document.getElementById("chints-opacity").value;
  var opacity = opacityRaw === "" ? 1 : Number(opacityRaw);
  if (!isFinite(opacity)) opacity = 1;
  opacity = Math.max(0, Math.min(1, opacity));
  var opacityLabel = document.getElementById("chints-opacity-label");
  if (opacityLabel) opacityLabel.textContent = Math.round(opacity * 100) + "%";
  var bg = chartHintsPreviewColor(document.getElementById("chints-bg").value, opacity, "#1d2129");
  var fg = chartHintsPreviewColor(document.getElementById("chints-text").value, 1, "#e6e6e6");
  var border = chartHintsPreviewColor(document.getElementById("chints-border").value, 1, "#3d4556");
  var font = chartHintsColor(document.getElementById("chints-font").value, "system-ui, sans-serif");
  var size = Number(document.getElementById("chints-fontsize").value);
  if (!isFinite(size) || size < 8) size = 11;
  var bw = Number(document.getElementById("chints-borderwidth").value);
  if (!isFinite(bw) || bw < 0) bw = 1;
  var pad = Number(document.getElementById("chints-padding").value);
  if (!isFinite(pad) || pad < 0) pad = 6;
  var radius = Number(document.getElementById("chints-radius").value);
  if (!isFinite(radius) || radius < 0) radius = 6;
  el.style.background = bg;
  el.style.color = fg;
  el.style.borderColor = border;
  el.style.borderWidth = bw + "px";
  el.style.padding = pad + "px";
  el.style.borderRadius = radius + "px";
  el.style.fontFamily = font;
  el.style.fontSize = size + "px";
  el.style.opacity = document.getElementById("chints-enabled").checked ? "1" : ".45";
  el.textContent = "Process\nRUNNING";
}
function openChartHintsCfg(n) {
  var node = n || chartSettingsNode;
  if (!node) return;
  chartHintsNode = node;
  chartHintsUseDefaults = false;
  var settingsOpen = document.getElementById("csettings").classList.contains("on");
  document.getElementById("chints-enabled").checked = settingsOpen ? document.getElementById("cs-state-hints").checked : node.chartStateHints !== false;
  var bgRaw = String(node.chartStateHintBg || "").trim();
  var bgTransparent = /^(transparent|none)$/i.test(bgRaw);
  var bgHex = chartHintHex(bgRaw, "#1d2129");
  var opacity = Number(node.chartStateHintOpacity);
  if (!isFinite(opacity)) opacity = 1;
  if (bgTransparent) opacity = 0;
  document.getElementById("chints-bg").value = bgHex;
  document.getElementById("chints-opacity").value = String(Math.max(0, Math.min(1, opacity)));
  document.getElementById("chints-text").value = node.chartStateHintTextColor || "";
  document.getElementById("chints-border").value = node.chartStateHintBorderColor || "";
  document.getElementById("chints-font").value = node.chartStateHintFont || "";
  document.getElementById("chints-fontsize").value = node.chartStateHintFontSize || "";
  document.getElementById("chints-borderwidth").value = node.chartStateHintBorderWidth || "";
  document.getElementById("chints-padding").value = node.chartStateHintPadding || "";
  document.getElementById("chints-radius").value = node.chartStateHintRadius || "";
  var box = document.getElementById("chintscfgbox");
  if (box && !box.__chintsDefaultsBound) {
    box.__chintsDefaultsBound = true;
    var markEdited = function() { chartHintsUseDefaults = false; };
    box.addEventListener("input", markEdited);
    box.addEventListener("change", markEdited);
  }
  updateChartHintsPreview();
  document.getElementById("chintscfg").classList.add("on");
}
function closeChartHintsCfg() {
  document.getElementById("chintscfg").classList.remove("on");
  chartHintsNode = null;
  chartHintsUseDefaults = false;
}
function chartHintsField(id) {
  return document.getElementById(id).value.trim();
}
function applyChartHintsCfg() {
  var n = chartHintsNode;
  if (!n) return;
  snap("chart hint settings");
  n.chartStateHints = document.getElementById("chints-enabled").checked;
  document.getElementById("cs-state-hints").checked = n.chartStateHints;
  if (chartHintsUseDefaults) {
    n.chartStateHintBg = "";
    n.chartStateHintOpacity = "";
    n.chartStateHintTextColor = "";
    n.chartStateHintBorderColor = "";
    n.chartStateHintFont = "";
    n.chartStateHintFontSize = "";
    n.chartStateHintBorderWidth = "";
    n.chartStateHintPadding = "";
    n.chartStateHintRadius = "";
  } else {
    n.chartStateHintBg = document.getElementById("chints-bg").value || "";
    n.chartStateHintOpacity = String(Number(document.getElementById("chints-opacity").value));
    n.chartStateHintTextColor = chartHintsField("chints-text");
    n.chartStateHintBorderColor = chartHintsField("chints-border");
    n.chartStateHintFont = chartHintsField("chints-font");
    n.chartStateHintFontSize = chartHintsField("chints-fontsize");
    n.chartStateHintBorderWidth = chartHintsField("chints-borderwidth");
    n.chartStateHintPadding = chartHintsField("chints-padding");
    n.chartStateHintRadius = chartHintsField("chints-radius");
  }
  closeChartHintsCfg();
  setOut();
  autoPreview();
  renderInspector();
}
function resetChartHintsCfg() {
  chartHintsUseDefaults = true;
  document.getElementById("chints-bg").value = "#1d2129";
  document.getElementById("chints-opacity").value = "1";
  ["chints-text", "chints-border", "chints-font"].forEach(function(id) {
    document.getElementById(id).value = "";
  });
  document.getElementById("chints-fontsize").value = "11";
  document.getElementById("chints-borderwidth").value = "1";
  document.getElementById("chints-padding").value = "6";
  document.getElementById("chints-radius").value = "6";
  updateChartHintsPreview();
}
var chartElementHintsNode = null;
var chartElementHintsDraft = null;
function chartElementLineText(value) {
  return String(value === null || value === undefined ? "" : value).replace(/\r\n?|\n/g, "|");
}
function chartElementRows(data) {
  var rows = [];
  var addStates = function(states) {
    if (Array.isArray(states)) states.forEach(function(item, key) { if (item && typeof item === "object") rows.push({ kind: "state", key: key, item: item }); });
    else if (states && typeof states === "object") Object.keys(states).forEach(function(key) { if (states[key] && typeof states[key] === "object") rows.push({ kind: "state", key: key, item: states[key] }); });
  };
  var addTransitions = function(transitions) {
    if (Array.isArray(transitions)) transitions.forEach(function(item, key) { if (item && typeof item === "object") rows.push({ kind: "transition", key: key, item: item }); });
    else if (transitions && typeof transitions === "object") Object.keys(transitions).forEach(function(key) { if (transitions[key] && typeof transitions[key] === "object") rows.push({ kind: "transition", key: key, item: transitions[key] }); });
  };
  if (Array.isArray(data)) addStates(data);
  else if (data && typeof data === "object") {
    addStates(data.states || data.nodes);
    addTransitions(data.transitions || data.edges);
  }
  return rows;
}
function chartElementJsonDraft(source) {
  var data;
  try { data = JSON.parse(source); } catch (e) { return { error: "Invalid JSON: " + e.message }; }
  var items = chartElementRows(data);
  var rows = items.map(function(ref) {
    var item = ref.item;
    var label = ref.kind === "state"
      ? (item.label || item.id || String(ref.key))
      : ((item.from || "?") + " \u2192 " + (item.to || "?") + (item.id ? " (" + item.id + ")" : ""));
    return {
      kind: ref.kind,
      key: ref.key,
      label: label,
      getTag: function() { return item.tag || ""; },
      getHint: function() { return item.hint || ""; },
      setTag: function(value) { if (value !== "") item.tag = value; else delete item.tag; },
      setHint: function(value) { if (value !== "") item.hint = value; else delete item.hint; }
    };
  });
  return {
    kind: "json",
    data: data,
    rows: rows,
    serialize: function() { return JSON.stringify(data, null, 2); }
  };
}
function chartElementCsvQuote(value, sep) {
  var s = String(value === null || value === undefined ? "" : value);
  if (s.indexOf(sep) >= 0 || s.indexOf('"') >= 0 || /[\r\n]/.test(s)) {
    return '"' + s.replace(/"/g, '""') + '"';
  }
  return s;
}
function chartElementCsvDraft(source) {
  var rows = chartDataCsvRows(source);
  if (!rows.length) return { error: "The CSV chart data is empty." };
  var sep = chartDataDetectSep(source);
  var header = rows[0].map(function(cell) { return String(cell === null || cell === undefined ? "" : cell).trim(); });
  var lower = header.map(function(cell) { return cell.toLowerCase(); });
  var known = ["id", "state", "key", "from", "source", "to", "target", "label", "name", "tag", "hint", "tooltip", "tip"];
  var hasHeader = lower.some(function(cell) { return known.indexOf(cell) >= 0; });
  if (!hasHeader) return { error: "CSV chart data needs a header row (id/label/tag/hint or from/to/tag/hint)." };
  var body = rows.slice(1).map(function(row) { return row.slice(); });
  var indexOf = function(names) {
    for (var i = 0; i < lower.length; i++) if (names.indexOf(lower[i]) >= 0) return i;
    return -1;
  };
  var fromIdx = indexOf(["from", "source", "fromstate", "from_state"]);
  var toIdx = indexOf(["to", "target", "tostate", "to_state"]);
  var isTransition = fromIdx >= 0 && toIdx >= 0;
  var idIdx = indexOf(["id", "state", "key"]);
  var labelIdx = indexOf(["label", "name", "title"]);
  var tagIdx = indexOf(["tag", "comment"]);
  var hintIdx = indexOf(["hint", "tooltip", "tip"]);
  var width = header.length;
  body.forEach(function(row) { while (row.length < width) row.push(""); });
  if (tagIdx < 0) {
    header.push("tag");
    lower.push("tag");
    tagIdx = header.length - 1;
    body.forEach(function(row) { row.push(""); });
  }
  if (hintIdx < 0) {
    header.push("hint");
    lower.push("hint");
    hintIdx = header.length - 1;
    body.forEach(function(row) { row.push(""); });
  }
  var refs = body.map(function(row, index) {
    var kind = isTransition ? "transition" : "state";
    var label;
    if (kind === "transition") {
      label = (row[fromIdx] || "?") + " \u2192 " + (row[toIdx] || "?") + (idIdx >= 0 && row[idIdx] ? " (" + row[idIdx] + ")" : "");
    } else {
      label = (labelIdx >= 0 && row[labelIdx]) || (idIdx >= 0 && row[idIdx]) || ("#" + (index + 1));
    }
    return {
      kind: kind,
      key: index,
      label: label,
      getTag: function() { return row[tagIdx] || ""; },
      getHint: function() { return row[hintIdx] || ""; },
      setTag: function(value) { row[tagIdx] = value; },
      setHint: function(value) { row[hintIdx] = value; }
    };
  });
  return {
    kind: "csv",
    rows: refs,
    serialize: function() {
      return [header].concat(body).map(function(row) {
        return row.map(function(cell) { return chartElementCsvQuote(cell, sep); }).join(sep);
      }).join("\n");
    }
  };
}
function chartElementXmlDraft(source) {
  var doc;
  try { doc = new DOMParser().parseFromString(source, "application/xml"); } catch (e) { return { error: "Invalid XML." }; }
  if (!doc || !doc.documentElement || doc.getElementsByTagName("parsererror").length) return { error: "Invalid XML." };
  var refs = [];
  var seen = [];
  var add = function(el, kind) {
    if (!el || !el.getAttribute || seen.indexOf(el) >= 0) return;
    seen.push(el);
    var id = el.getAttribute("id") || el.getAttribute("key") || "";
    var label = kind === "transition"
      ? (el.getAttribute("from") || "?") + " \u2192 " + (el.getAttribute("to") || "?") + (id ? " (" + id + ")" : "")
      : (el.getAttribute("label") || el.getAttribute("name") || id || "state");
    var tagChild = el.querySelector ? el.querySelector("tag") : null;
    var hintChild = el.querySelector ? el.querySelector("hint") : null;
    refs.push({
      kind: kind,
      key: refs.length,
      label: label,
      getTag: function() { return el.getAttribute("tag") || el.getAttribute("note") || (tagChild ? tagChild.textContent : "") || ""; },
      getHint: function() { return el.getAttribute("hint") || el.getAttribute("tooltip") || (hintChild ? hintChild.textContent : "") || ""; },
      setTag: function(value) {
        el.removeAttribute("note");
        if (tagChild) {
          if (value !== "") tagChild.textContent = value;
          else if (tagChild.parentNode) { tagChild.parentNode.removeChild(tagChild); tagChild = null; }
        }
        if (value !== "") el.setAttribute("tag", value); else el.removeAttribute("tag");
      },
      setHint: function(value) {
        el.removeAttribute("tooltip");
        if (hintChild) {
          if (value !== "") hintChild.textContent = value;
          else if (hintChild.parentNode) { hintChild.parentNode.removeChild(hintChild); hintChild = null; }
        }
        if (value !== "") el.setAttribute("hint", value); else el.removeAttribute("hint");
      }
    });
  };
  var rootName = (doc.documentElement.localName || doc.documentElement.nodeName || "").toLowerCase();
  if (rootName === "state" || rootName === "node") add(doc.documentElement, "state");
  if (rootName === "transition" || rootName === "edge") add(doc.documentElement, "transition");
  var states = doc.getElementsByTagName("state");
  for (var si = 0; si < states.length; si++) add(states[si], "state");
  var nodes = doc.getElementsByTagName("node");
  for (var ni = 0; ni < nodes.length; ni++) add(nodes[ni], "state");
  var transitions = doc.getElementsByTagName("transition");
  for (var ti = 0; ti < transitions.length; ti++) add(transitions[ti], "transition");
  var edges = doc.getElementsByTagName("edge");
  for (var ei = 0; ei < edges.length; ei++) add(edges[ei], "transition");
  if (!refs.length) return { error: "No states or links found in the XML data." };
  return {
    kind: "xml",
    rows: refs,
    serialize: function() { return new XMLSerializer().serializeToString(doc); }
  };
}
function chartElementDraft(source, format) {
  var f = String(format || "").toLowerCase();
  if (!f || f === "auto") {
    var first = String(source || "").replace(/^\s+/, "").charAt(0);
    f = first === "<" ? "xml" : (first === "{" || first === "[" ? "json" : "csv");
  }
  if (f === "json") return chartElementJsonDraft(source);
  if (f === "csv") return chartElementCsvDraft(source);
  if (f === "xml") return chartElementXmlDraft(source);
  return { error: "Unsupported chart data format: " + f };
}
function chartElementLiteralText(expr) {
  var s = String(expr || "").trim();
  if (s.length < 2 || s.charAt(0) !== '"' || s.charAt(s.length - 1) !== '"') return null;
  var raw = s.slice(1, -1), out = "";
  for (var i = 0; i < raw.length; i++) {
    var ch = raw.charAt(i);
    if (ch === "\\" && i + 1 < raw.length) {
      var next = raw.charAt(i + 1);
      if (next === "n") { out += "\n"; i++; continue; }
      if (next === "t") { out += "\t"; i++; continue; }
      if (next === "\\" || next === "\"" || next === "{" || next === "}") { out += next; i++; continue; }
    }
    out += ch;
  }
  return out;
}
function chartElementResolveData(node) {
  if (!node) return { error: "" };
  if (!node.chartDataExpr) return { source: String(node.chartData || "").trim(), state: null, name: "" };
  var expr = String(node.chartData || "").trim();
  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(expr)) return { error: "The chart data is an expression; edit it in the data editor." };
  var states = (typeof root !== "undefined" && root && Array.isArray(root.states)) ? root.states : [];
  for (var i = 0; i < states.length; i++) {
    if (states[i] && states[i].name === expr) {
      var value = chartElementLiteralText(states[i].init);
      if (value === null) return { error: "State '" + expr + "' is not a literal string; edit it in the data editor." };
      return { source: value, state: states[i], name: expr };
    }
  }
  return { error: "The chart data is an expression; edit it in the data editor." };
}
function openChartElementHints(n) {
  var node = n || chartSettingsNode;
  if (!node) return;
  chartElementHintsNode = node;
  chartElementHintsDraft = null;
  var msg = document.getElementById("celemhints-msg");
  var box = document.getElementById("celemhints-rows");
  msg.textContent = "";
  box.textContent = "";
  var resolved = chartElementResolveData(node);
  if (resolved.error) {
    msg.textContent = resolved.error;
  } else {
    var source = resolved.source;
    if (!source) msg.textContent = "The chart has no data.";
    else {
      var draft = chartElementDraft(source, node.chartFormat);
      if (draft.error) msg.textContent = draft.error;
      else if (!draft.rows.length) msg.textContent = "No states or links found in the chart data.";
      else {
        draft.state = resolved.state || null;
        chartElementHintsDraft = draft;
        if (resolved.state) msg.textContent = "Editing the value of state '" + resolved.name + "'.";
        var head = document.createElement("div");
        head.className = "celemhints-row celemhints-head";
        ["Element", "Tag", "Hint"].forEach(function(label) { var cell = document.createElement("div"); cell.textContent = label; head.appendChild(cell); });
        box.appendChild(head);
        draft.rows.forEach(function(ref, index) {
          var row = document.createElement("div");
          row.className = "celemhints-row";
          row.setAttribute("data-elem-index", String(index));
          var label = document.createElement("div");
          label.className = "lbl";
          label.textContent = ref.label;
          var tag = document.createElement("input");
          tag.type = "text";
          tag.className = "celemhints-input";
          tag.value = chartElementLineText(ref.getTag());
          tag.setAttribute("aria-label", "Tag for " + label.textContent);
          var hint = document.createElement("input");
          hint.type = "text";
          hint.className = "celemhints-input";
          hint.value = chartElementLineText(ref.getHint());
          hint.setAttribute("aria-label", "Hint for " + label.textContent);
          row.appendChild(label);
          row.appendChild(tag);
          row.appendChild(hint);
          box.appendChild(row);
        });
      }
    }
  }
  document.getElementById("celemhints").classList.add("on");
}
function closeChartElementHints() {
  document.getElementById("celemhints").classList.remove("on");
  chartElementHintsNode = null;
  chartElementHintsDraft = null;
}
function applyChartElementHints() {
  var n = chartElementHintsNode;
  var draft = chartElementHintsDraft;
  if (!n || !draft) { closeChartElementHints(); return; }
  snap("chart element hints");
  var rows = document.querySelectorAll("#celemhints-rows .celemhints-row[data-elem-index]");
  for (var i = 0; i < rows.length; i++) {
    var index = Number(rows[i].getAttribute("data-elem-index"));
    var ref = draft.rows[index];
    if (!ref) continue;
    var inputs = rows[i].querySelectorAll("input.celemhints-input");
    var tag = inputs[0] ? inputs[0].value : "";
    var hint = inputs[1] ? inputs[1].value : "";
    ref.setTag(tag);
    ref.setHint(hint);
  }
  var serialized = draft.serialize();
  if (draft.state) {
    draft.state.init = '"' + esc(serialized) + '"';
  } else {
    n.chartData = serialized;
    n.chartDataExpr = false;
  }
  closeChartElementHints();
  setOut();
  autoPreview();
  renderInspector();
}
var chartExtraNode = null;
function addChartExtraRow(key, value) {
  var rows = document.getElementById("cextrattrs-rows");
  if (!rows) return;
  var row = document.createElement("div");
  row.className = "cextra-row";
  var name = document.createElement("input");
  name.type = "text";
  name.className = "cextra-key";
  name.placeholder = "attribute name";
  name.value = key || "";
  var val = document.createElement("input");
  val.type = "text";
  val.className = "cextra-value";
  val.placeholder = '"text", 24, true';
  val.value = value || "";
  var remove = document.createElement("button");
  remove.type = "button";
  remove.className = "cextra-remove";
  remove.textContent = "\u00d7";
  remove.title = "Remove attribute";
  remove.addEventListener("click", function() { if (row.parentNode) row.parentNode.removeChild(row); });
  row.appendChild(name);
  row.appendChild(val);
  row.appendChild(remove);
  rows.appendChild(row);
}
function openChartExtraCfg(n) {
  var node = n || chartSettingsNode;
  if (!node) return;
  chartExtraNode = node;
  var rows = document.getElementById("cextrattrs-rows");
  var msg = document.getElementById("cextrattrs-msg");
  rows.textContent = "";
  msg.textContent = "";
  var extra = node.chartExtra && typeof node.chartExtra === "object" ? node.chartExtra : {};
  Object.keys(extra).sort().forEach(function(key) { addChartExtraRow(key, extra[key]); });
  document.getElementById("cextrattrs").classList.add("on");
}
function closeChartExtraCfg() {
  document.getElementById("cextrattrs").classList.remove("on");
  chartExtraNode = null;
}
function applyChartExtraCfg() {
  var n = chartExtraNode;
  if (!n) return;
  var rows = document.querySelectorAll("#cextrattrs-rows .cextra-row");
  var extra = {};
  var error = "";
  for (var i = 0; i < rows.length; i++) {
    var keyInput = rows[i].querySelector(".cextra-key");
    var valueInput = rows[i].querySelector(".cextra-value");
    var key = keyInput ? keyInput.value.trim() : "";
    var value = valueInput ? valueInput.value.trim() : "";
    if (!key && !value) continue;
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(key)) { error = "Invalid attribute name: " + key; break; }
    if (!value) { error = "Attribute '" + key + "' has no value."; break; }
    extra[key] = value;
  }
  var msg = document.getElementById("cextrattrs-msg");
  if (error) { msg.textContent = error; return; }
  snap("chart attributes");
  n.chartExtra = extra;
  closeChartExtraCfg();
  setOut();
  autoPreview();
  renderInspector();
}
var chartSettingsNode = null;
function showChartSettingsTab(name) {
  document.querySelectorAll(".cset-tab").forEach(function(b) { b.classList.toggle("on", b.getAttribute("data-cset") === name); });
  document.querySelectorAll("#csettings .cset-pane").forEach(function(p) { p.classList.toggle("on", p.getAttribute("data-pane") === name); });
}
function openChartSettings(n) {
  chartSettingsNode = n;
  document.getElementById("csettings-title").textContent = isGraphKind(n) ? "Graph settings" : "Chart settings";
  var theme = document.getElementById("cs-theme");
  theme.value = n.chartTheme === "light" ? "light" : "dark";
  document.getElementById("cs-colorpicker").checked = n.chartColorPicker !== false;
  var preset = document.getElementById("cs-preset");
  preset.innerHTML = "";
  CHART_STYLE_PRESETS.forEach(function(p) {
    var o = document.createElement("option"); o.value = p.key; o.textContent = p.label; preset.appendChild(o);
  });
  var custom = document.createElement("option"); custom.value = "custom"; custom.textContent = "Custom (keep style expr)"; preset.appendChild(custom);
  preset.value = chartPresetKey(n);
  var tv = document.getElementById("cs-tableview");
  tv.innerHTML = "";
  var gantt = (n.chartType || "bar") === "gantt";
  [["chart", gantt ? "Timeline" : "Chart only"], ["table", "Table only"], ["both", gantt ? "Timeline + table" : "Chart + table"]].forEach(function(v) {
    var o = document.createElement("option"); o.value = v[0]; o.textContent = v[1]; tv.appendChild(o);
  });
  tv.value = n.chartTableView === "gantt" ? "chart" : (n.chartTableView || "chart");
  document.getElementById("cs-pointsize").value = n.chartPointSize || "3";
  document.getElementById("cs-interactive").checked = !!n.chartInteractive;
  document.getElementById("cs-state-edit").checked = n.chartStateMode !== "static";
  document.getElementById("cs-state-hints").checked = n.chartStateHints !== false;
  document.getElementById("cs-gantt-resources").checked = !!n.chartShowResources;
  document.getElementById("cs-gantt-progress").checked = !!n.chartShowProgress;
  document.getElementById("cs-gantt-progress-column").checked = !!n.chartProgressColumn;
  document.getElementById("cs-gantt-slack").checked = !!n.chartShowSlack;
  document.getElementById("cs-gantt-costs").checked = !!n.chartShowCosts;
  document.getElementById("cs-gantt-shownonworking").checked = n.chartShowNonWorkingDays !== false;
  document.getElementById("cs-gantt-statuscolors").checked = !!n.chartStatusColors;
  document.getElementById("cs-gantt-autoschedule").checked = !!n.chartAutoSchedule;
  document.getElementById("cs-gantt-scale").value = n.chartGanttScale || "auto";
  document.getElementById("cs-gantt-linkcolor").value = /^#[0-9a-f]{6}$/i.test(n.chartLinkColor || "") ? n.chartLinkColor : "#94a3b8";
  document.getElementById("cs-gantt-linkwidth").value = n.chartLinkWidth || "1";
  document.getElementById("cs-gantt-linkgap").value = n.chartLinkGap || "4";
  document.getElementById("cs-gantt-linkstyle").value = n.chartLinkStyle || "auto";
  document.getElementById("cs-gantt-arrowtype").value = n.chartArrowType || "triangle";
  document.getElementById("cs-gantt-arrowsize").value = n.chartArrowSize || "5";
  document.getElementById("cs-gantt-barborder-enable").checked = !!n.chartBarBorderColor;
  document.getElementById("cs-gantt-barborder").value = /^#[0-9a-f]{6}$/i.test(n.chartBarBorderColor || "") ? n.chartBarBorderColor : "#475569";
  document.getElementById("cs-critical").checked = !!n.chartCriticalPath;
  document.getElementById("cs-holidays").value = n.chartHolidays || "";
  document.getElementById("cs-weekend").value = n.chartWeekend || "";
  document.getElementById("cs-timezone").value = n.chartTimezone || "";
  document.getElementById("cs-x").value = n.chartX || "";
  document.getElementById("cs-y").value = n.chartY || "";
  document.getElementById("cs-tt-bg").value = chartTooltipColor(n, "--sui-chart-tooltip-bg", "#111827");
  document.getElementById("cs-tt-fg").value = chartTooltipColor(n, "--sui-chart-tooltip-fg", "#f8fafc");
  document.getElementById("cs-tt-border").value = chartTooltipColor(n, "--sui-chart-tooltip-border", "#475569");
  document.getElementById("cs-tt-radius").value = chartTooltipLength(n, "--sui-chart-tooltip-radius", "6px");
  document.getElementById("cs-tt-maxwidth").value = chartTooltipLength(n, "--sui-chart-tooltip-max-width", "240px");
  document.getElementById("cs-tt-fontsize").value = chartTooltipLength(n, "--sui-chart-tooltip-font-size", "11px");
  var delay = parseInt(chartTooltipCssValue(n, "--sui-chart-tooltip-delay", "0ms"), 10);
  document.getElementById("cs-tt-delay").value = isNaN(delay) ? 0 : Math.max(0, Math.min(5000, delay));
  document.getElementById("cs-tb-filters").checked = n.chartTableFilters !== false;
  document.getElementById("cs-tb-addrow").checked = !!n.chartTableAddRow;
  document.getElementById("cs-tb-deleterow").checked = !!n.chartTableDeleteRow;
  document.getElementById("cs-tb-steppers").checked = !!n.chartTableSteppers;
  document.getElementById("cs-tb-contextmenu").checked = !!n.chartTableContextMenu;
  document.getElementById("cs-tb-selection").value = n.chartTableSelection || "row";
  document.getElementById("cs-tb-fontsize").value = chartTableCssValue(n, "--sui-chart-table-font-size", "11px");
  document.getElementById("cs-tb-padding").value = chartTableCssValue(n, "--sui-chart-table-cell-padding", "4px 6px");
  document.getElementById("cs-tb-headbg").value = chartTooltipColor(n, "--sui-chart-table-head-bg", "#14171c");
  var stripe = parseFloat(chartTableCssValue(n, "--sui-chart-table-stripe-opacity", "0.06"));
  document.getElementById("cs-tb-stripe").value = isFinite(stripe) ? stripe : 0.06;
  document.getElementById("cs-tb-caption").checked = chartTableCssValue(n, "--sui-chart-table-caption-display", "block") !== "none";
  document.getElementById("cs-tb-sticky").checked = chartTableCssValue(n, "--sui-chart-table-header-position", "sticky") !== "static";
  showChartSettingsTab("appearance");
  document.getElementById("csettings").classList.add("on");
}
function closeChartSettings() {
  document.getElementById("csettings").classList.remove("on");
  chartSettingsNode = null;
}
function openDataFromChartSettings() {
  var n = chartSettingsNode;
  if (!n) return;
  closeChartSettings();
  openChartDataEditor(n);
}
function chartSettingsStyleCss(n, base) {
  var parts = chartTableCssWithoutVars(chartTooltipCssWithoutVars(base));
  var ttBg = document.getElementById("cs-tt-bg").value || "#111827";
  var ttFg = document.getElementById("cs-tt-fg").value || "#f8fafc";
  var ttBorder = document.getElementById("cs-tt-border").value || "#475569";
  var ttRadius = normLen(document.getElementById("cs-tt-radius").value) || "6px";
  var ttMax = normLen(document.getElementById("cs-tt-maxwidth").value) || "240px";
  var ttFont = normLen(document.getElementById("cs-tt-fontsize").value) || "11px";
  var ttDelay = parseInt(document.getElementById("cs-tt-delay").value, 10);
  if (isNaN(ttDelay) || ttDelay < 0) ttDelay = 0;
  if (ttDelay > 5000) ttDelay = 5000;
  parts.push("--sui-chart-tooltip-bg: " + ttBg);
  parts.push("--sui-chart-tooltip-fg: " + ttFg);
  parts.push("--sui-chart-tooltip-border: " + ttBorder);
  parts.push("--sui-chart-tooltip-radius: " + ttRadius);
  parts.push("--sui-chart-tooltip-max-width: " + ttMax);
  parts.push("--sui-chart-tooltip-font-size: " + ttFont);
  parts.push("--sui-chart-tooltip-delay: " + ttDelay + "ms");
  parts.push("--sui-chart-table-font-size: " + (normLen(document.getElementById("cs-tb-fontsize").value) || "11px"));
  parts.push("--sui-chart-table-cell-padding: " + (document.getElementById("cs-tb-padding").value.trim() || "4px 6px"));
  parts.push("--sui-chart-table-head-bg: " + (document.getElementById("cs-tb-headbg").value || "#14171c"));
  var stripe2 = parseFloat(document.getElementById("cs-tb-stripe").value);
  if (!isFinite(stripe2) || stripe2 < 0) stripe2 = 0;
  if (stripe2 > 1) stripe2 = 1;
  parts.push("--sui-chart-table-stripe-opacity: " + stripe2);
  parts.push("--sui-chart-table-caption-display: " + (document.getElementById("cs-tb-caption").checked ? "block" : "none"));
  parts.push("--sui-chart-table-header-position: " + (document.getElementById("cs-tb-sticky").checked ? "sticky" : "static"));
  return parts.join("; ");
}
function applyChartSettings() {
  var n = chartSettingsNode;
  if (!n) return;
  var css = chartStyleCss(n);
  if (css === null) {
    alert("This chart uses a custom style expression. Edit it manually in Appearance.");
    return;
  }
  var presetKey = document.getElementById("cs-preset").value;
  var base = presetKey === "custom" ? css : chartPresetCss(presetKey);
  snap("chart settings");
  n.styleExpr = "cssSafe(" + JSON.stringify(chartSettingsStyleCss(n, base)) + ")";
  n.chartTheme = document.getElementById("cs-theme").value === "light" ? "light" : "dark";
  n.chartColorPicker = document.getElementById("cs-colorpicker").checked;
  n.chartPointSize = document.getElementById("cs-pointsize").value || "3";
  n.chartInteractive = document.getElementById("cs-interactive").checked;
  n.chartStateMode = document.getElementById("cs-state-edit").checked ? "editing" : "static";
  n.chartStateHints = document.getElementById("cs-state-hints").checked;
  n.chartShowResources = document.getElementById("cs-gantt-resources").checked;
  n.chartShowProgress = document.getElementById("cs-gantt-progress").checked;
  n.chartProgressColumn = document.getElementById("cs-gantt-progress-column").checked;
  n.chartShowSlack = document.getElementById("cs-gantt-slack").checked;
  n.chartShowCosts = document.getElementById("cs-gantt-costs").checked;
  n.chartShowNonWorkingDays = document.getElementById("cs-gantt-shownonworking").checked;
  n.chartStatusColors = document.getElementById("cs-gantt-statuscolors").checked;
  n.chartAutoSchedule = document.getElementById("cs-gantt-autoschedule").checked;
  n.chartGanttScale = document.getElementById("cs-gantt-scale").value || "auto";
  n.chartLinkColor = document.getElementById("cs-gantt-linkcolor").value || "#94a3b8";
  n.chartLinkWidth = document.getElementById("cs-gantt-linkwidth").value || "1";
  n.chartLinkGap = document.getElementById("cs-gantt-linkgap").value || "4";
  n.chartLinkStyle = document.getElementById("cs-gantt-linkstyle").value || "auto";
  n.chartArrowType = document.getElementById("cs-gantt-arrowtype").value || "triangle";
  n.chartArrowSize = document.getElementById("cs-gantt-arrowsize").value || "5";
  n.chartBarBorderColor = document.getElementById("cs-gantt-barborder-enable").checked ? document.getElementById("cs-gantt-barborder").value : "";
  n.chartTableView = document.getElementById("cs-tableview").value || "chart";
  n.chartTableFilters = document.getElementById("cs-tb-filters").checked;
  n.chartTableAddRow = document.getElementById("cs-tb-addrow").checked;
  n.chartTableDeleteRow = document.getElementById("cs-tb-deleterow").checked;
  n.chartTableSteppers = document.getElementById("cs-tb-steppers").checked;
  n.chartTableContextMenu = document.getElementById("cs-tb-contextmenu").checked;
  n.chartTableSelection = document.getElementById("cs-tb-selection").value || "row";
  n.chartCriticalPath = document.getElementById("cs-critical").checked;
  n.chartHolidays = document.getElementById("cs-holidays").value.trim();
  n.chartWeekend = document.getElementById("cs-weekend").value.trim();
  n.chartTimezone = document.getElementById("cs-timezone").value.trim();
  n.chartX = document.getElementById("cs-x").value || "";
  n.chartY = document.getElementById("cs-y").value || "";
  closeChartSettings();
  setOut();
  autoPreview();
  renderInspector();
}
function resetChartSettings() {
  var n = chartSettingsNode;
  if (!n) return;
  var css = chartStyleCss(n);
  if (css === null) {
    alert("This chart uses a custom style expression. Edit it manually in Appearance.");
    return;
  }
  var parts = chartTableCssWithoutVars(chartTooltipCssWithoutVars(css));
  snap("reset chart styles");
  n.styleExpr = parts.length ? "cssSafe(" + JSON.stringify(parts.join("; ")) + ")" : "";
  closeChartSettings();
  setOut();
  autoPreview();
  renderInspector();
}
function tooltipCssValue(name, fallback) {
  var re = new RegExp(name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&") + "\\s*:\\s*([^;]+);");
  var m = String(root.css || "").match(re);
  return m ? m[1].trim() : fallback;
}
function openTooltipCfg() {
  document.getElementById("tt-place").value = tooltipCssValue("--sui-tooltip-place", "top");
  document.getElementById("tt-delay").value = parseInt(tooltipCssValue("--sui-tooltip-delay", "140"), 10) || 0;
  document.getElementById("tt-bg").value = /^#[0-9a-f]{6}$/i.test(tooltipCssValue("--sui-tooltip-bg", "#1d2129")) ? tooltipCssValue("--sui-tooltip-bg", "#1d2129") : "#1d2129";
  document.getElementById("tt-fg").value = /^#[0-9a-f]{6}$/i.test(tooltipCssValue("--sui-tooltip-fg", "#e6e6ef")) ? tooltipCssValue("--sui-tooltip-fg", "#e6e6ef") : "#e6e6ef";
  document.getElementById("tt-border").value = /^#[0-9a-f]{6}$/i.test(tooltipCssValue("--sui-tooltip-border", "#3a4152")) ? tooltipCssValue("--sui-tooltip-border", "#3a4152") : "#3a4152";
  document.getElementById("tt-radius").value = tooltipCssValue("--sui-tooltip-radius", "6px");
  document.getElementById("tt-maxwidth").value = tooltipCssValue("--sui-tooltip-max-width", "260px");
  document.getElementById("ttcfg").classList.add("on");
}
function closeTooltipCfg() {
  document.getElementById("ttcfg").classList.remove("on");
}
function applyTooltipCfg() {
  var place = document.getElementById("tt-place").value;
  if (["top", "right", "bottom", "left"].indexOf(place) < 0) place = "top";
  var delay = parseInt(document.getElementById("tt-delay").value, 10);
  if (isNaN(delay) || delay < 0) delay = 140;
  if (delay > 5000) delay = 5000;
  var bg = document.getElementById("tt-bg").value || "#1d2129";
  var fg = document.getElementById("tt-fg").value || "#e6e6ef";
  var border = document.getElementById("tt-border").value || "#3a4152";
  var radius = normLen(document.getElementById("tt-radius").value) || "6px";
  var maxWidth = normLen(document.getElementById("tt-maxwidth").value) || "260px";
  snap("tooltip settings");
  var clean = String(root.css || "").replace(/\/\* sui-tooltip-settings \*\/\s*:root\s*\{[^}]*\}\s*/g, "").trim();
  var block = "/* sui-tooltip-settings */\n:root {\n" +
    "  --sui-tooltip-place: " + place + ";\n" +
    "  --sui-tooltip-delay: " + delay + "ms;\n" +
    "  --sui-tooltip-bg: " + bg + ";\n" +
    "  --sui-tooltip-fg: " + fg + ";\n" +
    "  --sui-tooltip-border: " + border + ";\n" +
    "  --sui-tooltip-radius: " + radius + ";\n" +
    "  --sui-tooltip-max-width: " + maxWidth + ";\n" +
    "}";
  root.css = block + (clean ? "\n" + clean : "");
  closeTooltipCfg();
  scheduleCssSync();
  setOut();
  autoPreview();
}
function openPrefs() {
  var p = prefs;
  document.getElementById("pf-theme").value = p.theme;
  document.getElementById("pf-size").value = p.font;
  document.getElementById("pf-density").value = p.density;
  document.getElementById("pf-accent").value = p.accent;
  document.getElementById("pf-left").checked = p.showLeft;
  document.getElementById("pf-right").checked = p.showRight;
  document.getElementById("pf-bottom").checked = p.showBottom;

  var fsel = document.getElementById("pf-editor-font");
  var fcustom = document.getElementById("pf-editor-font-custom");
  var known = false;
  for (var i = 0; i < fsel.options.length; i++) {
    if (fsel.options[i].value === p.editorFont) { known = true; break; }
  }
  fsel.value = known ? p.editorFont : "__custom__";
  fcustom.value = known ? "" : (p.editorFont || "");
  document.getElementById("pf-editor-size").value = p.editorSize;
  document.getElementById("pf-editor-lh").value = p.editorLineHeight;
  document.getElementById("pf-editor-tab").value = String(p.editorTabSize);
  document.getElementById("pf-editor-wrap").checked = !!p.editorWrap;
  document.getElementById("pf-editor-lines").checked = p.editorLineNumbers !== false;
  document.getElementById("pf-editor-syntax").checked = p.editorSyntaxColors !== false;
  document.getElementById("pf-editor-brackets").checked = p.editorBracketMatch !== false;
  document.getElementById("pf-editor-linehl").checked = p.editorSelectedLine !== false;
  document.getElementById("pref").classList.add("on");
}
function closePrefs() { document.getElementById("pref").classList.remove("on"); }

// ---- styled prompt modal --------------------------------------------------
// askPrompt replaces window.prompt so editor-facing questions (rename, new
// handler function, ...) match the designer theme and stay centered. The
// callback receives the trimmed input; cancelling (Cancel, Esc, backdrop,
// empty input) just closes the dialog without calling it.
var __askCb = null;
function askPrompt(title, label, value, onOk, extra) {
  if (!onOk) return;
  document.getElementById("ask-title").textContent = title;
  document.getElementById("ask-label").textContent = label;
  document.getElementById("ask-extra").textContent = extra || "";
  var inp = document.getElementById("ask-input");
  inp.value = value === undefined || value === null ? "" : value;
  document.getElementById("askdlg").classList.add("on");
  __askCb = onOk;
  inp.focus();
  inp.select();
}
function closeAsk() {
  document.getElementById("askdlg").classList.remove("on");
  __askCb = null;
}
function askSubmit() {
  var cb = __askCb;
  if (!cb) { closeAsk(); return; }
  var v = document.getElementById("ask-input").value.trim();
  closeAsk();
  if (v) cb(v);
}
(function wireAskDialog() {
  var dlg = document.getElementById("askdlg");
  var inp = document.getElementById("ask-input");
  document.getElementById("ask-ok").onclick = askSubmit;
  document.getElementById("ask-cancel").onclick = closeAsk;
  // Clicking the dark backdrop cancels; clicking the card does not.
  dlg.addEventListener("mousedown", function(ev){ if (ev.target === dlg) closeAsk(); });
  inp.addEventListener("keydown", function(ev){
    if (ev.key === "Enter") { ev.preventDefault(); askSubmit(); }
    else if (ev.key === "Escape") { ev.preventDefault(); closeAsk(); }
  });
})();

// ---- confirmation / message modal -----------------------------------------
// confirmBox guards destructive template actions (Dash/Form/CLI replace the
// whole project, delete removes a saved template). showMsg surfaces hard
// failures (a template that does not type-check inside the project) as a real
// centered dialog — transient status-bar text was too easy to miss.
var __confCb = null;
function showModal(title, text, onOk, showCancel) {
  document.getElementById("conf-title").textContent = title;
  document.getElementById("conf-text").textContent = text;
  document.getElementById("conf-cancel").style.display = showCancel ? "" : "none";
  document.getElementById("confdlg").classList.add("on");
  __confCb = onOk || null;
  document.getElementById("conf-ok").focus();
}
function confirmBox(title, text, onOk) { showModal(title, text, onOk, true); }
function showMsg(title, text) { showModal(title, text, null, false); }
function closeConf() { document.getElementById("confdlg").classList.remove("on"); __confCb = null; }
function confOk() { var cb = __confCb; closeConf(); if (cb) cb(); }
(function wireConfDialog() {
  var dlg = document.getElementById("confdlg");
  document.getElementById("conf-ok").onclick = confOk;
  document.getElementById("conf-cancel").onclick = closeConf;
  dlg.addEventListener("mousedown", function(ev){ if (ev.target === dlg) closeConf(); });
  dlg.addEventListener("keydown", function(ev){
    if (ev.key === "Escape") { ev.preventDefault(); closeConf(); }
  });
})();
(function wirePageDialog() {
  var dlg = document.getElementById("pagedlg");
  if (!dlg) return;
  document.getElementById("page-ok").onclick = pageSubmit;
  document.getElementById("page-cancel").onclick = closePage;
  dlg.addEventListener("mousedown", function(ev){ if (ev.target === dlg) closePage(); });
  document.getElementById("page-name").addEventListener("keydown", function(ev){
    if (ev.key === "Enter") { ev.preventDefault(); pageSubmit(); }
    else if (ev.key === "Escape") { ev.preventDefault(); closePage(); }
  });
})();

function applyPrefs() {
  prefs.theme = document.getElementById("pf-theme").value;
  prefs.font = parseInt(document.getElementById("pf-size").value, 10) || 14;
  prefs.density = document.getElementById("pf-density").value;
  prefs.accent = document.getElementById("pf-accent").value;
  prefs.showLeft = document.getElementById("pf-left").checked;
  prefs.showRight = document.getElementById("pf-right").checked;
  prefs.showBottom = document.getElementById("pf-bottom").checked;
  var fsel = document.getElementById("pf-editor-font").value;
  var fcustom = document.getElementById("pf-editor-font-custom").value.trim();
  prefs.editorFont = safeEditorFont(fsel === "__custom__" ? (fcustom || PREF_DEFAULTS.editorFont) : fsel);
  prefs.editorSize = parseInt(document.getElementById("pf-editor-size").value, 10) || 16;
  prefs.editorLineHeight = parseFloat(document.getElementById("pf-editor-lh").value) || 1.6;
  prefs.editorTabSize = parseInt(document.getElementById("pf-editor-tab").value, 10) || 4;
  prefs.editorWrap = document.getElementById("pf-editor-wrap").checked;
  prefs.editorLineNumbers = document.getElementById("pf-editor-lines").checked;
  prefs.editorSyntaxColors = document.getElementById("pf-editor-syntax").checked;
  prefs.editorBracketMatch = document.getElementById("pf-editor-brackets").checked;
  prefs.editorSelectedLine = document.getElementById("pf-editor-linehl").checked;
  savePrefs();
  closePrefs();
  applyTheme();
  applyPanelVisibility();
}

function resetPrefs() {
  prefs = Object.assign({}, PREF_DEFAULTS);
  savePrefs();
  document.getElementById("pf-theme").value = prefs.theme;
  document.getElementById("pf-size").value = prefs.font;
  document.getElementById("pf-density").value = prefs.density;
  document.getElementById("pf-accent").value = prefs.accent;
  document.getElementById("pf-left").checked = prefs.showLeft;
  document.getElementById("pf-right").checked = prefs.showRight;
  document.getElementById("pf-bottom").checked = prefs.showBottom;
  document.getElementById("pf-editor-font").value = "__custom__";
  document.getElementById("pf-editor-font-custom").value = prefs.editorFont;
  document.getElementById("pf-editor-size").value = prefs.editorSize;
  document.getElementById("pf-editor-lh").value = prefs.editorLineHeight;
  document.getElementById("pf-editor-tab").value = String(prefs.editorTabSize);
  document.getElementById("pf-editor-wrap").checked = prefs.editorWrap;
  document.getElementById("pf-editor-lines").checked = prefs.editorLineNumbers;
  document.getElementById("pf-editor-syntax").checked = prefs.editorSyntaxColors;
  document.getElementById("pf-editor-brackets").checked = prefs.editorBracketMatch;
  document.getElementById("pf-editor-linehl").checked = prefs.editorSelectedLine;
  applyTheme();
  applyPanelVisibility();
}

function applyPanelVisibility() {
  document.getElementById("left").style.display = prefs.showLeft ? "" : "none";
  document.getElementById("right").style.display = prefs.showRight ? "" : "none";
  var b = document.getElementById("bottom");
  if (prefs.showBottom) b.classList.add("on"); else b.classList.remove("on");
}

function safeEditorFont(v) {
  var s = String(v || "").replace(/[;{}<>]/g, "").trim().slice(0, 200);
  return s || PREF_DEFAULTS.editorFont;
}
function applyEditorPrefs() {
  var r = document.documentElement.style;
  var z = (parseFloat(prefs.font) || 14) / 14;
  if (!z || isNaN(z) || z <= 0) z = 1;
  var size = Math.max(10, Math.min(40, parseFloat(prefs.editorSize) || 16));
  var lh = Math.max(1, Math.min(3, parseFloat(prefs.editorLineHeight) || 1.6));
  var tab = Math.max(1, Math.min(12, parseInt(prefs.editorTabSize, 10) || 4));
  r.setProperty("--editor-font", safeEditorFont(prefs.editorFont));
  // Keep the requested editor size independent from the global UI zoom.
  r.setProperty("--editor-size", (size / z) + "px");
  r.setProperty("--editor-lh", String(lh));
  r.setProperty("--editor-tab", String(tab));
  var ed = document.querySelector(".editor");
  if (ed) {
    ed.classList.toggle("no-line-numbers", prefs.editorLineNumbers === false);
    ed.classList.toggle("wrap", !!prefs.editorWrap);
    ed.classList.toggle("plain-code", prefs.editorSyntaxColors === false);
  }
  if (prefs.editorBracketMatch === false) {
    __brMatch = { open: -1, close: -1 };
  }
  if (document.getElementById("sui_hl")) highlightSui();
}

function applyTheme() {
  var r = document.documentElement.style;
  r.setProperty("--fs", prefs.font + "px");
  r.setProperty("--accent", prefs.accent);
  // Scale the whole UI so the font-size preference has a visible effect even
  // though individual controls hard-code px sizes.
  document.body.style.zoom = String(Math.round((prefs.font / 14) * 100) / 100);
  var light = {
    "--bg":"#f5f6fa", "--panel":"#eceef4", "--panel2":"#e2e5ee", "--panel3":"#d8dce8",
    "--border":"#c8ccd8", "--border2":"#aab0c0",
    "--fg":"#1a1d26", "--muted":"#5a6072", "--dim":"#7a8094",
    "--btn-bg":"#e2e5ee", "--btn-hover":"#cdd3e2", "--btn-on":"#c4d0ff",
    "--palbtn-bg":"#eef0f6", "--sel":"#c9d6ff", "--input-bg":"#ffffff",
    "--gutter-bg":"#e9ecf4", "--gutter-fg":"#7a8094"
  };
  var contrast = {
    "--bg":"#000000", "--panel":"#0a0a0a", "--panel2":"#111111", "--panel3":"#1a1a1a",
    "--border":"#555555", "--border2":"#999999",
    "--fg":"#ffffff", "--muted":"#e0e0e0", "--dim":"#c0c0c0",
    "--btn-bg":"#000000", "--btn-hover":"#333333", "--btn-on":"#0044cc",
    "--palbtn-bg":"#000000", "--sel":"#0044cc", "--input-bg":"#000000",
    "--gutter-bg":"#000000", "--gutter-fg":"#aaaaaa"
  };
  if (prefs.theme === "dark") {
    ["--bg","--panel","--panel2","--panel3","--border","--border2","--fg","--muted","--dim","--btn-bg","--btn-hover","--btn-on","--palbtn-bg","--sel","--input-bg","--gutter-bg","--gutter-fg"].forEach(function(k){ r.removeProperty(k); });
  } else {
    var t = prefs.theme === "light" ? light : contrast;
    for (var k in t) r.setProperty(k, t[k]);
  }
  applyEditorPrefs();
}

// ---- code editing: highlight + two-way sync -------------------------------
var KW_SUI = { view:1, fn:1, state:1, async:1, if:1, else:1, loop:1, while:1, in:1, return:1, break:1, continue:1, let:1, import:1, from:1, title:1, true:1, false:1, null:1, and:1, or:1, not:1, type:1, struct:1, css:1, component:1, props:1 };
var KW_JS = { function:1, const:1, let:1, var:1, return:1, if:1, else:1, for:1, while:1, of:1, in:1, new:1, null:1, true:1, false:1, async:1, await:1, await:1, typeof:1, undefined:1, class:1, this:1, switch:1, case:1, break:1, continue:1, delete:1, void:1, instanceof:1 };
var CSS_PROPS = { display:1,"flex-direction":1,"justify-content":1,"align-items":1,"align-self":1,"align-content":1,"vertical-align":1,gap:1,margin:1,padding:1,border:1,"border-radius":1,"border-style":1,"border-width":1,"border-color":1,width:1,height:1,"max-width":1,"min-width":1,color:1,background:1,"background-color":1,opacity:1,"font-size":1,"font-family":1,"font-weight":1,"font-style":1,"text-align":1,"text-transform":1,"text-decoration":1,"line-height":1,"letter-spacing":1,"word-spacing":1,"text-indent":1,"text-shadow":1,"box-shadow":1,"white-space":1,position:1,top:1,left:1,right:1,bottom:1,"z-index":1,"grid-template-columns":1,"grid-template-rows":1,"grid-column":1,"grid-row":1,"grid-area":1,"place-items":1,"place-content":1,"justify-items":1,"align-contentx":1 };
var CSS_KW = { important:1, media:1, keyframes:1, import:1, charset:1, supports:1 };
function escHtml(s) { return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;"); }

// suiHighlightSymbols collects the names declared in the current buffer:
// functions, states, parameters, local let bindings and loop variables. The
// tokenizer uses these maps to give declarations and their uses a stable,
// readable colour without changing glyph metrics.
function suiHighlightSymbols(code) {
  var fns = Object.create(null), vars = Object.create(null), m, re;
  function addVar(raw) {
    var pm = /^([A-Za-z_]\w*)/.exec(String(raw || "").trim());
    if (pm) vars[pm[1]] = 1;
  }
  re = /^[ \t]*(?:async[ \t]+)?fn[ \t]+([A-Za-z_]\w*)/gm;
  while ((m = re.exec(code)) !== null) fns[m[1]] = 1;
  re = /^[ \t]*state[ \t]+([A-Za-z_]\w*)/gm;
  while ((m = re.exec(code)) !== null) vars[m[1]] = 1;
  re = /\blet[ \t]+([A-Za-z_]\w*)/g;
  while ((m = re.exec(code)) !== null) vars[m[1]] = 1;
  re = /\bloop[ \t]+([A-Za-z_]\w*)(?:[ \t]*,[ \t]*([A-Za-z_]\w*))?[ \t]+in\b/g;
  while ((m = re.exec(code)) !== null) {
    vars[m[1]] = 1;
    if (m[2]) vars[m[2]] = 1;
  }
  // Function and component parameters are variables too.
  re = /^[ \t]*(?:async[ \t]+)?fn[ \t]+[A-Za-z_]\w*[ \t]*\(([^)]*)\)/gm;
  while ((m = re.exec(code)) !== null) {
    m[1].split(",").forEach(addVar);
  }
  re = /^[ \t]*view[ \t]+[A-Za-z_]\w*[ \t]*\(([^)]*)\)/gm;
  while ((m = re.exec(code)) !== null) {
    m[1].split(",").forEach(addVar);
  }
  return { fns: fns, vars: vars };
}

function highlight(code, lang) {
  if (!code) return "";
  var out = "";
  var toks = tokenizeHighlight(code, lang);
  for (var t = 0; t < toks.length; t++) {
    var tk = toks[t];
    if (tk.cls) out += '<span class="' + tk.cls + '">' + escHtml(tk.text) + '</span>';
    else out += escHtml(tk.text);
  }
  return out;
}

// tokenizeHighlight splits source into coloured runs so any renderer (plain or
// overlaid) can share the exact same syntax classification. Each entry is
// { text, cls } where cls is a token class ("tok-kw", "tok-str", ...) or "".
function tokenizeHighlight(code, lang, symbolsOverride) {
  var out = [], i = 0, n = code.length;
  var symbols = symbolsOverride || (lang === 'sui' ? suiHighlightSymbols(code) : null);
  function push(cls, s) { if (s !== "") out.push({ text: s, cls: cls ? "tok-" + cls : "" }); }
  function interpolationEnd(start, max) {
    var depth = 1;
    for (var j = start + 1; j < max; j++) {
      if (code[j] === '\\') { j++; continue; }
      if (code[j] === '{') depth++;
      else if (code[j] === '}') { depth--; if (depth === 0) return j; }
    }
    return -1;
  }
  while (i < n) {
    var c = code[i];
    if (c === '/' && code[i+1] === '/') { var j = i; while (j < n && code[j] !== '\n') j++; push('cm', code.slice(i, j)); i = j; continue; }
    if (c === '/' && code[i+1] === '*') { var k = i + 2; while (k < n && !(code[k] === '*' && code[k+1] === '/')) k++; k = Math.min(n, k + 2); push('cm', code.slice(i, k)); i = k; continue; }
    if (c === '"' || c === "'") {
      var q = c, e = i + 1;
      while (e < n && code[e] !== q) { if (code[e] === '\\') e++; e++; }
      e = Math.min(n, e + 1);
      var closed = e > i + 1 && code[e - 1] === q;
      var contentEnd = closed ? e - 1 : e;
      // SUI text strings can interpolate expressions: "Count {count}". Keep
      // the text green but let variables/functions inside braces keep their
      // semantic colour. The CSS string (css "...") is not an interpolation.
      if (lang === 'sui' && q === '"' && !/\bcss[ \t]*$/.test(code.slice(0, i))) {
        push('str', code.slice(i, i + 1));
        var pos = i + 1, last = pos;
        while (pos < contentEnd) {
          if (code[pos] === '\\') { pos += 2; continue; }
          if (code[pos] === '{') {
            var close = interpolationEnd(pos, contentEnd);
            if (close < 0) break;
            if (last < pos) push('str', code.slice(last, pos));
            push('str', '{');
            var inner = tokenizeHighlight(code.slice(pos + 1, close), 'sui', symbols);
            for (var ii = 0; ii < inner.length; ii++) out.push(inner[ii]);
            push('str', '}');
            pos = close + 1;
            last = pos;
          } else {
            pos++;
          }
        }
        if (last < contentEnd) push('str', code.slice(last, contentEnd));
        if (closed) push('str', code.slice(contentEnd, e));
        i = e;
        continue;
      }
      push('str', code.slice(i, e));
      i = e;
      continue;
    }
    if (/[0-9]/.test(c) || (c === '#' && /[0-9a-fA-F]/.test(code[i+1] || ''))) {
      var j2 = i;
      if (c === '#') { j2++; while (j2 < n && /[0-9a-fA-F]/.test(code[j2])) j2++; }
      else { while (j2 < n && /[0-9a-zA-Z.%]/.test(code[j2])) j2++; }
      push('num', code.slice(i, j2)); i = j2; continue;
    }
    if (/[A-Za-z_@-]/.test(c)) {
      // The start class includes @ (for CSS at-rules), so the inner class must
      // accept it too. Otherwise a stray "@" never advances i and freezes the
      // editor while invalid source is being highlighted.
      var j3 = i + 1;
      while (j3 < n && /[A-Za-z0-9_@-]/.test(code[j3])) j3++;
      var w = code.slice(i, j3);
      var t = "";
      if (lang === 'sui') {
        var afterDot = i > 0 && code[i - 1] === '.';
        if (KW_SUI[w]) t = 'kw';
        else if (symbols && !afterDot && symbols.fns[w]) t = 'fnname';
        else if (symbols && !afterDot && symbols.vars[w]) t = 'var';
        else if (DOCS[w] || w === 'str' || w === 'len' || w === 'int' || w === 'float' || w === 'bool' || w.startsWith('http.') || w.startsWith('time.') || w.startsWith('timers.') || w.startsWith('url.') || w.startsWith('encoding.') || w.startsWith('storage.') || w.startsWith('clipboard.') || w.startsWith('validation.') || w === 'match' || w === 'matchAll' || w === 'replaceRegex' || w === 'replaceFirst' || w === 'search' || w === 'test' || w === 'splitRegex' || w === 'extract' || w === 'extractAll') t = 'fn';
      }
      else if (lang === 'css') { if (CSS_KW[w]) t = 'kw'; else if (CSS_PROPS[w]) t = 'prop'; else if (w[0] === '@') t = 'kw'; }
      else if (lang === 'js') t = (KW_JS[w] ? 'kw' : '');
      else t = (w === 'true' || w === 'false' || w === 'null') ? 'kw' : '';
      push(t, w); i = j3; continue;
    }
    push('', c); i++;
  }
  return out;
}
// supHighlight renders SUI highlight HTML with overlays:
//   - the matched bracket pair (.br-match)
//   - reference marks (.cm-ref) and definition marks (.cm-def) for
//     find-references / go-to-definition
//   - the line range of the element selected in the tree (.cm-elline)
//   - inline diagnostics (.cm-error / .cm-warn) from the compiler
// When no overlay is active it falls back to the plain highlighter.
function supHighlight(code) {
  var br = (prefs.editorBracketMatch !== false && __brMatch.open >= 0 && __brMatch.close >= 0);
  var hasMarks = (__cmMarks && __cmMarks.length) || (prefs.editorSelectedLine && __elline && __elline.line >= 0);
  var hasDiags = __suiDiags && __suiDiags.length;
  if (!br && !hasMarks && !hasDiags) return highlight(code, "sui");
  return overlayHighlight(code);
}
function overlayHighlight(code) {
  var marks = {};   // char index -> overlay class name(s)
  function addClass(i, cls) {
    if (i < 0 || i >= code.length) return;
    marks[i] = marks[i] ? (marks[i] + " " + cls) : cls;
  }
  if (__brMatch.open >= 0 && __brMatch.close >= 0) {
    addClass(__brMatch.open, "br-match");
    addClass(__brMatch.close, "br-match");
  }
  if (__cmMarks) {
    __cmMarks.forEach(function(m) {
      for (var i = m.start; i < m.end && i < code.length; i++) addClass(i, m.cls || "cm-ref");
    });
  }
  if (prefs.editorSelectedLine && __elline && __elline.line >= 0) {
    var ls = 0; for (var k = 0; k < __elline.line; k++) { var nl = code.indexOf("\n", ls); if (nl < 0) { ls = code.length; break; } ls = nl + 1; }
    var le = code.indexOf("\n", ls); if (le < 0) le = code.length;
    for (var j = ls; j < le; j++) if (!marks[j]) marks[j] = "cm-elline";
  }
  if (__suiDiags) {
    __suiDiags.forEach(function(d) {
      var cls = (d.kind === "warning") ? "cm-warn" : "cm-error";
      var start = lineColOffset(code, d.line - 1, Math.max(0, (d.col || 1) - 1));
      // Underline the token at the position, or the rest of the line.
      var end = start;
      while (end < code.length && /[A-Za-z0-9_.]/.test(code[end])) end++;
      if (end === start) { while (end < code.length && code[end] !== "\n") end++; }
      for (var i = start; i < end; i++) addClass(i, cls);
    });
  }
  // Render by walking the syntax tokens and splitting each one at overlay
  // boundaries, so every character keeps its token colour AND gets the overlay
  // class. This is what keeps formatting visible while an overlay is active.
  var toks = tokenizeHighlight(code, "sui");
  var html = "", off = 0;
  function span(cls, text) {
    if (text === "") return "";
    if (cls) return '<span class="' + cls + '">' + escHtml(text) + '</span>';
    return escHtml(text);
  }
  for (var t = 0; t < toks.length; t++) {
    var tk = toks[t], text = tk.text, base = tk.cls;
    var seg = "", segCls = null, pos = off;
    for (var ci = 0; ci < text.length; ci++, pos++) {
      var cls2 = marks[pos] ? (base ? base + " " + marks[pos] : marks[pos]) : base;
      if (segCls === null) { segCls = cls2; seg = text[ci]; }
      else if (cls2 === segCls) { seg += text[ci]; }
      else { html += span(segCls, seg); segCls = cls2; seg = text[ci]; }
    }
    html += span(segCls, seg);
    off += text.length;
  }
  return html;
}

var __suiAc = { items:[], idx:0, visible:false, manual:false };
function hideSuiAc(){ var p=document.getElementById('sui-ac'); if(p){ p.style.display='none'; __suiAc.visible=false; __suiAc.manual=false; } }
// suiTextRange maps absolute character offsets in the highlighted <pre> to a
// DOM Range. The overlay is the visual source of truth, so measuring its
// ranges keeps the completion popup attached to the real caret even after
// scrolling or with a custom editor font.
function suiTextRange(root, from, to) {
  var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
  var n, acc = 0, startNode = null, startOff = 0, endNode = null, endOff = 0;
  while ((n = walker.nextNode())) {
    var len = n.nodeValue.length;
    if (startNode === null && acc + len >= from) {
      startNode = n;
      startOff = Math.max(0, Math.min(len, from - acc));
    }
    if (endNode === null && acc + len >= to) {
      endNode = n;
      endOff = Math.max(0, Math.min(len, to - acc));
      break;
    }
    acc += len;
  }
  if (!startNode) return null;
  if (!endNode) { endNode = startNode; endOff = startOff; }
  var r = document.createRange();
  try { r.setStart(startNode, startOff); r.setEnd(endNode, endOff); } catch (e) { return null; }
  return r;
}
function suiCaretRect(pos) {
  var hl = document.getElementById("sui_hl");
  var ta = document.getElementById("out");
  var rect = null;
  if (hl) {
    var r = suiTextRange(hl, pos, pos);
    if (r) rect = r.getBoundingClientRect();
    if ((!rect || (!rect.width && !rect.height && !rect.left && !rect.top)) && pos > 0) {
      var r2 = suiTextRange(hl, pos - 1, pos);
      if (r2) {
        var rr = r2.getBoundingClientRect();
        if (rr) rect = { left: rr.right, right: rr.right, top: rr.top, bottom: rr.bottom, width: 0, height: rr.height || 1 };
      }
    }
  }
  if (rect && (rect.left || rect.top || rect.bottom)) return rect;
  var tr = ta ? ta.getBoundingClientRect() : { left: 20, top: 60 };
  return { left: tr.left + 12, right: tr.left + 12, top: tr.top + 12, bottom: tr.top + 28, width: 0, height: 16 };
}
function showSuiAc(list, manual){
  var p=document.getElementById('sui-ac'); if(!p){ p=document.createElement('div'); p.id='sui-ac'; p.className='sui-ac'; document.body.appendChild(p); }
  __suiAc.items=list||[]; __suiAc.idx=0;
  if(__suiAc.items.length===0){ hideSuiAc(); return; }
  __suiAc.manual = !!manual;
  p.innerHTML='';
  __suiAc.items.forEach(function(it, i){
    var d=document.createElement('div'); d.className='sui-ac-item'+(i===0?' active':'');
    var lab=document.createElement('span'); lab.textContent=it.label;
    d.appendChild(lab);
    var right=document.createElement('span'); right.style.cssText='margin-left:auto;flex:none;display:flex;gap:6px;align-items:baseline';
    if(it.desc){ var ds=document.createElement('span'); ds.className='ac-desc'; ds.textContent=it.desc; right.appendChild(ds); }
    if(it.kind){ var kd=document.createElement('span'); kd.className='ac-kind'; kd.textContent=it.kind; right.appendChild(kd); }
    d.appendChild(right);
    d.onmousedown=function(e){ e.preventDefault(); };
    d.onclick=function(e){ e.preventDefault(); applySuiAc(it); };
    p.appendChild(d);
  });
  var ta=document.getElementById('out');
  p.style.display='block';
  // Anchor the popup just below the caret. Menu coordinates are unscaled CSS
  // pixels while getBoundingClientRect() reports zoomed pixels (see menuZoom).
  var rect=suiCaretRect(ta.selectionStart);
  var z=menuZoom();
  var mw=p.offsetWidth, mh=p.offsetHeight;
  var vw=window.innerWidth/z, vh=window.innerHeight/z;
  var left=Math.max(4, Math.min(vw-mw-8, rect.left/z));
  var top=rect.bottom/z+3;
  if(top+mh > vh-4) top=rect.top/z-mh-4;
  if(top < 4) top=4;
  p.style.left=left+'px';
  p.style.top=top+'px';
  __suiAc.visible=true;
}
function moveSuiAc(dir){ if(!__suiAc.visible) return; __suiAc.idx = (__suiAc.idx+dir+__suiAc.items.length)%__suiAc.items.length; var items=document.querySelectorAll('#sui-ac .sui-ac-item'); items.forEach(function(it,i){ it.classList.toggle('active', i===__suiAc.idx); }); items[__suiAc.idx].scrollIntoView({block:'nearest'}); }
function applySuiAc(it){ if(!it) return; var ta=document.getElementById('out'); var start=ta.selectionStart, end=ta.selectionEnd; var val=ta.value;
  if(it.mkfn){ createFunctionForHandler(start); return; }
  // Replace the word already typed before the caret (if any) so suggestions
  // don't duplicate the prefix.
  var wi=start-1; while(wi>=0 && /[A-Za-z0-9_.]/.test(val[wi])) wi--; wi++;
  // With the caret before an existing identifier (for example at the start of
  // "onClick: inc"), replace that identifier too instead of producing "incinc".
  var we=end;
  while(we<val.length && /[A-Za-z0-9_.]/.test(val[we])) we++;
  if(wi===start && we>end) end=we;
  var word=it.insert||it.label;
  val = val.slice(0,wi) + word + val.slice(end);
  ta.value=val; ta.focus(); ta.selectionStart=ta.selectionEnd=wi+word.length; onSuiEdit(); hideSuiAc();
}

// createFunctionForHandler adds a fresh handler function and binds the event
// attribute at the caret to it. It is invoked from the "+ new function…"
// entry of the handler autocomplete so a handler can be authored on the fly.
function createFunctionForHandler(pos) {
  askPrompt("New handler function", "Name", "onClick", function(name) {
    name = name.replace(/[^A-Za-z0-9_]/g, "");
    if (!name) return;
    var ta = document.getElementById("out");
    var val = ta.value;
    // Bind: replace the (possibly empty) word after the attribute colon.
    var wi = pos - 1; while (wi >= 0 && /[A-Za-z0-9_.]/.test(val[wi])) wi--; wi++;
    val = val.slice(0, wi) + name + val.slice(pos);
    // Append a new async-less function at the end of the file.
    var sep = /\n\s*$/.test(val) ? "" : "\n";
    var fnsrc = sep + "fn " + name + "() -> null {\n\t\n}\n";
    val = val + fnsrc;
    ta.value = val;
    ta.focus();
    ta.selectionStart = ta.selectionEnd = wi + name.length;
    hideSuiAc();
    onSuiEdit();
    setStatus("created function " + name);
  });
}
// ---- integrated docs (builtins + stdlib) ---------------------------------
// Mirrors docs/BUILTINS.md and docs/STDLIB.md so the editor can show a
// signature and a short description on hover and in the autocomplete list.
var DOCS = {
  len: { sig: "len(s)", ret: "int", desc: "Length of a string or array." },
  str: { sig: "str(v)", ret: "string", desc: "Converts a value to string." },
  int: { sig: "int(v)", ret: "int", desc: "Converts a value to int." },
  float: { sig: "float(v)", ret: "float", desc: "Converts a value to float." },
  bool: { sig: "bool(v)", ret: "bool", desc: "Converts a value to bool." },
  upper: { sig: "upper(s)", ret: "string", desc: "Converts to uppercase." },
  lower: { sig: "lower(s)", ret: "string", desc: "Converts to lowercase." },
  trim: { sig: "trim(s)", ret: "string", desc: "Trims leading and trailing spaces." },
  trimLeft: { sig: "trimLeft(s)", ret: "string", desc: "Trims leading spaces." },
  trimRight: { sig: "trimRight(s)", ret: "string", desc: "Trims trailing spaces." },
  contains: { sig: "contains(s, sub)", ret: "bool", desc: "Checks whether it contains the substring." },
  startsWith: { sig: "startsWith(s, sub)", ret: "bool", desc: "Checks whether it starts with the substring." },
  endsWith: { sig: "endsWith(s, sub)", ret: "bool", desc: "Checks whether it ends with the substring." },
  indexOf: { sig: "indexOf(s, sub)", ret: "int", desc: "Index of the first occurrence (-1 if absent)." },
  lastIndexOf: { sig: "lastIndexOf(s, sub)", ret: "int", desc: "Index of the last occurrence (-1 if absent)." },
  slice: { sig: "slice(s, start, end)", ret: "string", desc: "Extracts a slice of the string." },
  substr: { sig: "substr(s, start, len)", ret: "string", desc: "Substring (legacy)." },
  repeat: { sig: "repeat(s, count)", ret: "string", desc: "Repeats the string n times." },
  padLeft: { sig: "padLeft(s, len, pad)", ret: "string", desc: "Left padding." },
  padRight: { sig: "padRight(s, len, pad)", ret: "string", desc: "Right padding." },
  chars: { sig: "chars(s)", ret: "array[string]", desc: "Splits the string into characters." },
  split: { sig: "split(s, sep)", ret: "array[string]", desc: "Splits by separator." },
  splitRegex: { sig: "splitRegex(s, pattern, [limit])", ret: "array[string]", desc: "Splits using a regular expression." },
  join: { sig: "join(arr, sep)", ret: "string", desc: "Joins an array with a separator." },
  replace: { sig: "replace(s, old, new)", ret: "string", desc: "Replaces all occurrences." },
  replaceFirst: { sig: "replaceFirst(s, old, new)", ret: "string", desc: "Replaces the first occurrence." },
  replaceRegex: { sig: "replaceRegex(s, pattern, repl, [flags])", ret: "string", desc: "Replacement using a regular expression." },
  match: { sig: "match(s, pattern, [flags])", ret: "array[string]", desc: "Finds regex matches." },
  matchAll: { sig: "matchAll(s, pattern, [flags])", ret: "array[any]", desc: "All regex matches." },
  search: { sig: "search(s, pattern)", ret: "int", desc: "Index of the first regex match (-1 if absent)." },
  test: { sig: "test(s, pattern)", ret: "bool", desc: "Checks the regex match." },
  extract: { sig: "extract(s, pattern)", ret: "string", desc: "Extracts the first captured group." },
  extractAll: { sig: "extractAll(s, pattern)", ret: "array[string]", desc: "Extracts all captured groups." },
  fmt: { sig: "fmt(n, decimals)", ret: "string", desc: "Formats a number with the given decimals." },
  abs: { sig: "abs(n)", ret: "number", desc: "Absolute value." },
  min: { sig: "min(a, b)", ret: "number", desc: "Minimum of two numbers." },
  max: { sig: "max(a, b)", ret: "number", desc: "Maximum of two numbers." },
  round: { sig: "round(n)", ret: "int", desc: "Rounds to the nearest integer." },
  floor: { sig: "floor(n)", ret: "int", desc: "Rounds down." },
  ceil: { sig: "ceil(n)", ret: "int", desc: "Rounds up." },
  sqrt: { sig: "sqrt(n)", ret: "float", desc: "Square root." },
  pow: { sig: "pow(base, exp)", ret: "float", desc: "Raises to a power." },
  random: { sig: "random()", ret: "float", desc: "Random number between 0 and 1." },
  randint: { sig: "randint(min, max)", ret: "int", desc: "Random integer between min and max." },
  "time.now": { sig: "time.now()", ret: "int", desc: "Current timestamp in milliseconds." },
  "time.format": { sig: "time.format(ms, pattern)", ret: "string", desc: "Formats a timestamp with a pattern." },
  "timers.setTimeout": { sig: "timers.setTimeout(fn, ms)", ret: "null", desc: "Runs fn after ms milliseconds." },
  "http.get": { sig: "http.get(url)", ret: "(string, string)", desc: "HTTP GET; returns body and error.", async: true },
  "http.getStatus": { sig: "http.getStatus(url)", ret: "(int, string)", desc: "HTTP GET; returns the status code.", async: true },
  "http.post": { sig: "http.post(url, body)", ret: "(string, string)", desc: "HTTP POST with body.", async: true },
  "http.getSafe": { sig: "http.getSafe(url, allow)", ret: "(string, string)", desc: "SSRF-guarded GET; allow is a literal origin/path list.", async: true },
  "http.getStatusSafe": { sig: "http.getStatusSafe(url, allow)", ret: "(int, string)", desc: "SSRF-guarded GET status.", async: true },
  "http.postSafe": { sig: "http.postSafe(url, body, allow)", ret: "(string, string)", desc: "SSRF-guarded POST.", async: true },
  "ws.connect": { sig: "ws.connect(url)", ret: "(any, string)", desc: "Opens a WebSocket connection.", async: true },
  "ws.connectSafe": { sig: "ws.connectSafe(url, allow)", ret: "(any, string)", desc: "SSRF-guarded WebSocket (ws/wss origins).", async: true },
  "ws.send": { sig: "ws.send(ws, data)", ret: "(null, string)", desc: "Sends a message on the socket.", async: true },
  "ws.onMessage": { sig: "ws.onMessage(ws, fn)", ret: "(null, string)", desc: "Registers a message handler.", async: true },
  "ws.close": { sig: "ws.close(ws)", ret: "(null, string)", desc: "Closes the WebSocket connection.", async: true },
  "fs.read": { sig: "fs.read(path)", ret: "(string, string)", desc: "Reads a file as a string.", async: true },
  "fs.write": { sig: "fs.write(path, data)", ret: "(null, string)", desc: "Writes a file.", async: true },
  "fs.exists": { sig: "fs.exists(path)", ret: "(bool, string)", desc: "Checks whether a file or directory exists.", async: true },
  "fs.stat": { sig: "fs.stat(path)", ret: "(string, string)", desc: "File or directory info.", async: true },
  "fs.mkdir": { sig: "fs.mkdir(path)", ret: "(null, string)", desc: "Creates a directory (mkdir -p).", async: true },
  "fs.remove": { sig: "fs.remove(path)", ret: "(null, string)", desc: "Removes a file or directory.", async: true },
  "fs.list": { sig: "fs.list(path)", ret: "(array[string], string)", desc: "Lists directory entries.", async: true },
  "exec.run": { sig: "exec.run(cmd, args)", ret: "(string, string)", desc: "Runs a command and returns its output.", async: true },
  "json.parse": { sig: "json.parse(s)", ret: "(any, string)", desc: "Parses a JSON string.", async: true },
  "json.stringify": { sig: "json.stringify(v)", ret: "(string, null)", desc: "Serializes a value to JSON.", async: true },
  "table.display": { sig: "table.display(name, data)", ret: "(int, string)", desc: "Fills the named table from JSON text, arrays or objects; keeps the design.", core: true },
  "table.data": { sig: "table.data(name)", ret: "(array[array[string]], string)", desc: "Reads the visible rows (including edits).", core: true },
  "table.header": { sig: "table.header(name)", ret: "(array[string], string)", desc: "Reads the header labels.", core: true },
  "table.select": { sig: "table.select(name, cols)", ret: "(array[array[string]], string)", desc: "Reads only the given column indexes.", core: true },
  "table.slice": { sig: "table.slice(name, from, count)", ret: "(array[array[string]], string)", desc: "Reads a block of rows (count <= 0 means to the end).", core: true },
  "table.importText": { sig: "table.importText(name, text, spec)", ret: "(int, string)", desc: "Parses CSV/TSV/SSV/PSV/JSON/JSONL text into the table (spec: csv, tsv, json, jsonl, ...;header=0;sep=...).", core: true },
  "table.exportText": { sig: "table.exportText(name, spec)", ret: "(string, string)", desc: "Serializes the visible table to CSV/TSV/SSV/PSV/JSON/JSONL.", core: true },
  "table.json": { sig: "table.json(name)", ret: "(string, string)", desc: "Exports the rows as JSON (objects when a header exists).", core: true },
  "table.clear": { sig: "table.clear(name)", ret: "(int, string)", desc: "Empties the visible rows.", core: true },
  "storage.get": { sig: "storage.get(key)", ret: "string", desc: "Reads a value from local storage." },
  "storage.set": { sig: "storage.set(key, value)", ret: "null", desc: "Writes a value to local storage." },
  "clipboard.writeText": { sig: "clipboard.writeText(s)", ret: "null", desc: "Copies the text to the clipboard." },
  "toast.show": { sig: "toast.show(message, [kind], [duration])", ret: "null", desc: "Shows a text toast. kind: info, success, warning, error; duration in ms (0 keeps it open)." },
  "toast.clear": { sig: "toast.clear()", ret: "null", desc: "Dismisses all visible toast notifications." },
  "validation.isEmail": { sig: "validation.isEmail(s)", ret: "bool", desc: "Checks whether the string is a valid email." }
};

// ---- contextual autocomplete ---------------------------------------------
// SUI is a small language but the useful completion set changes with the
// surrounding block. These helpers build a lightweight lexical scope from the
// braces before the caret (ignoring strings/comments) so the list can offer
// declarations at top level, elements inside a view, statements inside a
// function, attributes inside an element, and expressions inside statements.
var SUI_TOP_WORDS = ["import", "css", "state", "type", "struct", "fn", "async", "view"];
var SUI_STMT_WORDS = ["let", "if", "else", "loop", "while", "return", "break", "continue"];
var SUI_EXPR_WORDS = ["true", "false", "null", "and", "or", "not"];
var SUI_TYPES = ["any", "string", "int", "float", "bool", "null"];
var SUI_TAG_SET = null;

function suiAllTags() {
  if (SUI_TAG_SET) return SUI_TAG_SET;
  var seen = {}, out = [];
  Object.keys(TAG_FIELDS).forEach(function(k) {
    if (k === "body" || !/^[A-Za-z_][A-Za-z0-9_-]*$/.test(k) || seen[k]) return;
    seen[k] = 1; out.push(k);
  });
  SUI_TAG_SET = out.sort();
  return SUI_TAG_SET;
}

// suiDeclsFromSource reads declarations directly from the editor buffer, so
// completion stays useful before the 700ms auto-apply refreshes root.funcs.
function suiDeclsFromSource(src) {
  var out = { states: [], funcs: [], views: [], imports: [] }, m, re;
  re = /^\s*state\s+([A-Za-z_]\w*)\s*(?::\s*([^=\n]+?))?\s*=\s*([^\n]*)/gm;
  while ((m = re.exec(src)) !== null) {
    out.states.push({ name: m[1], type: (m[2] || "").trim(), init: (m[3] || "").trim() });
  }
  re = /^\s*(async\s+)?fn\s+([A-Za-z_]\w*)\s*\(([^)]*)\)\s*(?:->\s*([^\n{]+))?/gm;
  while ((m = re.exec(src)) !== null) {
    out.funcs.push({ name: m[2], async: !!m[1], params: (m[3] || "").trim(), ret: (m[4] || "null").trim() });
  }
  re = /^\s*view\s+([A-Za-z_]\w*)\s*(?:\(([^)]*)\))?/gm;
  while ((m = re.exec(src)) !== null) {
    out.views.push({ name: m[1], params: (m[2] || "").trim() });
  }
  re = /^\s*import\s*\{([^}]*)\}/gm;
  while ((m = re.exec(src)) !== null) {
    (m[1] || "").split(",").forEach(function(name) {
      name = name.trim();
      if (/^[A-Za-z_]\w*$/.test(name) && out.imports.indexOf(name) < 0) out.imports.push(name);
    });
  }
  return out;
}

function suiParamNames(raw) {
  var out = [];
  (raw || "").split(",").forEach(function(p) {
    p = p.trim();
    if (!p) return;
    var m = /^([A-Za-z_]\w*)/.exec(p);
    if (m) out.push(m[1]);
  });
  return out;
}

function suiSourceClasses(src) {
  var out = [], seen = {}, m, re = /\.([A-Za-z_][A-Za-z0-9_-]*)\s*\{/g;
  while ((m = re.exec(src)) !== null) {
    if (!seen[m[1]]) { seen[m[1]] = 1; out.push(m[1]); }
  }
  return out.sort();
}

// suiScopeAt returns the brace stack before pos. Each entry records the block
// kind and the header (function parameters, loop items, ...). Strings and
// comments are skipped so braces inside text do not corrupt the scope.
function suiScopeAt(src, pos) {
  var stack = [], inStr = false, inComment = false, esc = false, lineStart = 0;
  for (var i = 0; i < pos; i++) {
    var c = src[i];
    if (inComment) {
      if (c === "\n") { inComment = false; lineStart = i + 1; }
      continue;
    }
    if (inStr) {
      if (esc) { esc = false; continue; }
      if (c === "\\") { esc = true; continue; }
      if (c === "\"") { inStr = false; continue; }
      if (c === "\n") lineStart = i + 1;
      continue;
    }
    if (c === "#") { inComment = true; continue; }
    if (c === "\"") { inStr = true; continue; }
    if (c === "\n") { lineStart = i + 1; continue; }
    if (c === "{") {
      var head = src.slice(lineStart, i).trim(), block = { type: "element", header: head, start: i }, m;
      if ((m = /\b(?:async\s+)?fn\s+([A-Za-z_]\w*)\s*\(([^)]*)\)/.exec(head))) {
        block.type = "fn"; block.name = m[1]; block.params = m[2];
      } else if ((m = /\bview\s+([A-Za-z_]\w*)\s*(?:\(([^)]*)\))?/.exec(head))) {
        block.type = "view"; block.name = m[1]; block.params = m[2] || "";
      } else if ((m = /\bloop\s+([A-Za-z_]\w*)(?:\s*,\s*([A-Za-z_]\w*))?\s+in\b/.exec(head))) {
        block.type = "loop";
        block.items = [m[1]]; if (m[2]) block.items.push(m[2]);
      } else if (/\bwhile\b/.test(head)) {
        block.type = "while";
      } else if (/\bif\b/.test(head)) {
        block.type = "if";
      } else if (/\belse\b/.test(head)) {
        block.type = "else";
      }
      stack.push(block);
      lineStart = i + 1;
    } else if (c === "}") {
      stack.pop();
      lineStart = i + 1;
    }
  }
  return { stack: stack, top: stack.length ? stack[stack.length - 1] : null, inStr: inStr, inComment: inComment, lineStart: lineStart };
}

function suiExprPosition(before) {
  var s = before.replace(/\s+$/, "");
  if (/(?:^|[^A-Za-z0-9_)\]}])(?:return|let|if|while|loop|in|and|or|not)\s+[A-Za-z0-9_.]*$/.test(s)) return true;
  if (/[=+\-*\/%<>!,:(]\s*[A-Za-z0-9_.]*$/.test(s)) return true;
  if (/->\s*[A-Za-z0-9_\[\], ]*$/.test(s)) return true;
  return false;
}

function suiContext(val, pos) {
  var lineStart = val.lastIndexOf("\n", pos - 1) + 1;
  var before = val.slice(lineStart, pos);
  var wi = before.length - 1;
  while (wi >= 0 && /[A-Za-z0-9_.]/.test(before[wi])) wi--;
  var ctx = { kind: "expr", tag: "", attr: "", word: before.slice(wi + 1), lineStart: lineStart, pos: pos };
  var sc = suiScopeAt(val, pos);
  ctx.stack = sc.stack; ctx.block = sc.top;
  if (sc.inComment || /^\s*#/.test(before)) { ctx.kind = "none"; return ctx; }
  if (sc.inStr) {
    // class: "my-... still benefits from the CSS class list.
    var ccm = /\bclass\s*:\s*"?([A-Za-z0-9_-]*)$/.exec(before);
    if (ccm) { ctx.kind = "class"; ctx.word = ccm[1]; return ctx; }
    // Interpolations inside strings ("Count {count}") are expressions.
    ctx.kind = /\{[^}]*$/.test(before) ? "expr" : "none";
    return ctx;
  }
  var trimmed = before.replace(/\s+$/, "");
  // Handler values are declared functions, not generic expressions.
  var evm = /(on[A-Z][A-Za-z]*)\s*:\s*[A-Za-z0-9_.]*$/.exec(trimmed);
  if (evm) { ctx.kind = "handler"; ctx.attr = evm[1]; return ctx; }
  // class: "..." is backed by the CSS classes present in the source.
  var cm = /\bclass\s*:\s*"?([A-Za-z0-9_-]*)$/.exec(trimmed);
  if (cm) { ctx.kind = "class"; ctx.word = cm[1]; return ctx; }
  // A known element followed by whitespace is in attribute position, even
  // after a quoted label ("button \"Save\" on...").
  var noStr = before.replace(/"(?:\\.|[^"\\])*"/g, "\"\"");
  var tagm = /^\s*([A-Za-z_][A-Za-z0-9_-]*)\b/.exec(noStr);
  if (tagm) {
    var rest = noStr.slice(tagm[0].length);
    var tag = tagm[1];
    if (rest && rest.indexOf(":") < 0 && rest.indexOf("{") < 0 &&
        (suiAllTags().indexOf(tag) >= 0 || /^[A-Z]/.test(tag))) {
      ctx.kind = "attr"; ctx.tag = tag; return ctx;
    }
  }
  // Type positions: state/let annotations and function return types.
  if (/^\s*(?:state|let)\s+[A-Za-z_]\w*\s*:\s*[A-Za-z0-9_\[\], ]*$/.test(before) ||
      /(?:->|\b(?:async\s+)?fn\s+\w+\s*\([^)]*?[A-Za-z_]\w*\s*:)\s*[A-Za-z0-9_\[\], ]*$/.test(before)) {
    ctx.kind = "type"; return ctx;
  }
  if (suiExprPosition(before)) { ctx.kind = "expr"; return ctx; }
  var lineTrim = before.replace(/^\s+/, "");
  if (/^[A-Za-z0-9_.]*$/.test(lineTrim)) {
    var hasFn = false, hasView = false;
    sc.stack.forEach(function(b) { if (b.type === "fn") hasFn = true; if (b.type === "view" || b.type === "element") hasView = true; });
    ctx.kind = hasView ? "view" : (hasFn ? "stmt" : "top");
  }
  return ctx;
}

function attrSuggestions(tag) {
  var out = [], seen = {};
  function add(label, kind, insert, desc) {
    if (seen[label]) return;
    seen[label] = 1;
    out.push({ label: label, kind: kind, insert: insert, desc: desc || "" });
  }
  ["draggable", "dragKey", "dropKey", "dropFiles", "pickFiles", "pickMode"].forEach(function(f) {
    add(f, "attr", f + ": ", TAG_LABELS[f] || "");
  });
  if (tag === "file") add("directory", "attr", "directory: true", TAG_LABELS.directory);
  var fields = TAG_FIELDS[tag];
  if (fields) fields.forEach(function(f) { add(f, /^on[A-Z]/.test(f) ? "event" : "attr", f + ": ", TAG_LABELS[f] || ""); });
  var evs = ELEMENT_EVENTS[tag];
  if (evs) evs.forEach(function(e) { add(e, "event", e + ": ", "handler"); });
  add("class", "attr", "class: ", "CSS class");
  add("style", "attr", "style: ", "inline style");
  return out;
}

function declaredFns() { return (root && root.funcs) ? root.funcs : []; }
function declaredStates() { return (root && root.states) ? root.states : []; }

function suiMatchScore(label, word) {
  if (!word) return 0;
  return label.toLowerCase().indexOf(word.toLowerCase()) === 0 ? 0 : -1;
}
function suiRank(items, word) {
  var seen = {}, ranked = [];
  (items || []).forEach(function(it) {
    if (!it || !it.label) return;
    var score = suiMatchScore(it.label, word);
    if (score < 0 || seen[it.label]) return;
    seen[it.label] = 1;
    ranked.push({ it: it, score: score });
  });
  ranked.sort(function(a, b) { return a.score - b.score; });
  return ranked.map(function(x) { return x.it; });
}

function suiLocalNames(src, pos, scope) {
  var out = [], seen = {};
  function add(name, desc) {
    if (!name || seen[name]) return;
    seen[name] = 1;
    out.push({ label: name, kind: "var", insert: name, desc: desc || "local" });
  }
  var block = null;
  for (var i = scope.stack.length - 1; i >= 0; i--) {
    if (scope.stack[i].type === "fn" || scope.stack[i].type === "view") { block = scope.stack[i]; break; }
  }
  if (block) {
    suiParamNames(block.params).forEach(function(n) { add(n, "parameter"); });
    var seg = src.slice(block.start, pos), m;
    var reLet = /\blet\s+([A-Za-z_]\w*)/g;
    while ((m = reLet.exec(seg)) !== null) add(m[1], "local");
    var reLoop = /\bloop\s+([A-Za-z_]\w*)(?:\s*,\s*([A-Za-z_]\w*))?\s+in\b/g;
    while ((m = reLoop.exec(seg)) !== null) { add(m[1], "loop item"); if (m[2]) add(m[2], "loop index"); }
  }
  scope.stack.forEach(function(b) {
    if (b.type === "loop" && b.items) b.items.forEach(function(n) { add(n, "loop item"); });
  });
  return out;
}

function suiSuggestionItems(ctx, src, decls) {
  var items = [], word = ctx.word || "";
  if (ctx.kind === "handler") {
    var fns = decls.funcs.length ? decls.funcs : declaredFns();
    fns.forEach(function(f) {
      items.push({ label: f.name, kind: "fn", insert: f.name, desc: "fn(" + (f.params || "") + ") -> " + (f.ret || "null") });
    });
    items.push({ label: "+ new function…", kind: "fn", insert: "", desc: "create a new function", mkfn: true });
    return items;
  }
  if (ctx.kind === "class") {
    suiSourceClasses(src).forEach(function(c) { items.push({ label: c, kind: "class", insert: c, desc: "CSS class" }); });
    return items;
  }
  if (ctx.kind === "attr") return attrSuggestions(ctx.tag);
  if (ctx.kind === "type") {
    SUI_TYPES.forEach(function(t) { items.push({ label: t, kind: "type", insert: t, desc: "type" }); });
    return items;
  }
  if (ctx.kind === "top") {
    SUI_TOP_WORDS.forEach(function(k) {
      var ins = k + " ";
      if (k === "css") ins = "css \"";
      if (k === "import") ins = "import { ";
      items.push({ label: k, kind: "kw", insert: ins, desc: "top-level declaration" });
    });
    return items;
  }
  if (ctx.kind === "view") {
    items.push({ label: "title", kind: "kw", insert: "title \"", desc: "page title" });
    items.push({ label: "onLoad", kind: "event", insert: "onLoad: ", desc: "page load handler" });
    var tags = suiAllTags().slice();
    decls.views.forEach(function(v) { if (v.name && v.name !== "main") tags.push(v.name); });
    tags.forEach(function(t) { items.push({ label: t, kind: "tag", insert: t + " ", desc: "element" }); });
    return items;
  }
  if (ctx.kind === "stmt") {
    SUI_STMT_WORDS.forEach(function(k) { items.push({ label: k, kind: "kw", insert: k + " ", desc: "statement" }); });
  }
  var locals = suiLocalNames(src, ctx.pos, { stack: ctx.stack });
  locals.forEach(function(x) { items.push(x); });
  decls.states.forEach(function(s) {
    items.push({ label: s.name, kind: "state", insert: s.name, desc: "state " + (s.type || "") + " = " + (s.init || "") });
  });
  decls.funcs.forEach(function(f) {
    items.push({ label: f.name, kind: "fn", insert: f.name, desc: "fn -> " + (f.ret || "null") });
  });
  decls.imports.forEach(function(m) {
    items.push({ label: m, kind: "module", insert: m + ".", desc: "std module" });
  });
  Object.keys(DOCS).forEach(function(k) {
    items.push({ label: k, kind: "fn", insert: k + "(", desc: DOCS[k].sig });
  });
  if (ctx.kind !== "stmt") {
    SUI_EXPR_WORDS.forEach(function(k) { items.push({ label: k, kind: "kw", insert: k, desc: "literal/operator" }); });
  }
  return items;
}

function suiSuggest(force) {
  var ta = document.getElementById("out");
  if (!ta) return;
  var val = ta.value, pos = ta.selectionStart;
  var ctx = suiContext(val, pos);
  if (ctx.kind === "none") { hideSuiAc(); return; }
  var word = ctx.word || "";
  var manual = !!force || (__suiAc.visible && __suiAc.manual);
  var autoMin = 2;
  if (ctx.kind === "handler") autoMin = 0;
  else if (ctx.kind === "attr") autoMin = 1;
  if (!manual && word.length < autoMin) { hideSuiAc(); return; }
  var decls = suiDeclsFromSource(val);
  var ranked = suiRank(suiSuggestionItems(ctx, val, decls), word);
  if (!ranked.length) { hideSuiAc(); return; }
  showSuiAc(ranked.slice(0, 25), manual);
}

// findDefsInSource scans the editor text for declarations of a function/state.
function findDefsInSource(src, name) {
  var res = [], lines = src.split("\n");
  var esc = name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  var reFn = new RegExp("^\\s*(async\\s+)?fn\\s+" + esc + "\\b");
  var reSt = new RegExp("^\\s*state\\s+" + esc + "\\b");
  for (var i = 0; i < lines.length; i++) {
    if (reFn.test(lines[i])) res.push({ line: i, col: 0, kind: "fn" });
    else if (reSt.test(lines[i])) res.push({ line: i, col: 0, kind: "state" });
  }
  return res;
}
function findRefsInSource(src, name) {
  var res = [], re = new RegExp("\\b" + name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&") + "\\b", "g");
  var m;
  while ((m = re.exec(src)) !== null) {
    var idx = m.index, pre = src.slice(0, idx);
    res.push({ line: pre.split("\n").length - 1, col: idx - (pre.lastIndexOf("\n") + 1) });
  }
  return res;
}

function captureTextHistoryBurst() {
  if (__textHistoryActive) return;
  __textHistoryActive = true;
  // The model still reflects the last applied source; that is the baseline
  // this burst must restore. Typing is coalesced into one undo step until the
  // auto-apply finishes (or another designer action starts).
  snap("edit text", suiSource || currentSourceForHistory(), true);
}
function onSuiEdit() {
  captureTextHistoryBurst();
  undoCaptureSoon();
  highlightSui();
  updateApply();
  setStatus("editing SUI - auto-apply on pause, or click Apply");
  suiSuggest();
  scheduleLiveCheck();
  renderOutline();
  scheduleAutoApply();
}
// applySuiCommon compiles an edited SUI source and adopts the resulting model.
// opts.keepUndo keeps the designer undo stack (auto-apply), opts.keepSel
// restores the previous selection by class/kind after the remount, opts.quiet
// skips the "applied" status message.
function applySuiCommon(src, o) {
  o = o || {};
  var fd = new URLSearchParams(); fd.set("sui", src);
  if (!o.quiet) setStatus("applying...");
  fetch("/__gui__/apply", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (!j.ok) { setStatus((j.diagnostics || []).join(" | ")); return; }
    var model = j.model;
    var keepKind = null, keepCls = null, keepState = selState;
    if (o.keepSel && selPath) {
      var sn = byPath(selPath, root.children);
      if (sn) { keepKind = sn.kind; keepCls = (sn.class || "").trim().split(/\s+/)[0]; }
    }
    var prevName = curPage().name;
    root = { states: model.states || [], funcs: model.funcs || [], css: model.css || "", components: model.components || [], page: 0,
      pages: (model.pages && model.pages.length ? model.pages.map(function(p){ return { name: p.name || "main", title: p.title || "", children: p.elements || p.children || [] }; }) : [{ name: "main", title: model.title || "", children: model.elements || [] }]) };
    root.pages.forEach(function(p){ normalizeTables(p.children); });
    for (var i = 0; i < root.pages.length; i++) { if (root.pages[i].name === prevName) { root.page = i; break; } }
    remount();
    computeElementLines();
    if (!o.keepUndo) { selPath = null; selCell = null; selRow = null; selCol = null; selState = ""; past = []; future = []; }
    buildPickMap(); renderTree(); renderPages();
    document.getElementById("f_title").value = root.title;
    var fol = document.getElementById("f_onload"); if (fol) fol.value = curPage().onLoad || "";
    renderView();
    buildDatalist();
    clearFields(); renderInspector(); renderProblems(); renderTokens();
    if (o.keepSel && keepKind) {
      var rp = findElementPath(root.children, keepKind, keepCls);
      if (rp) { selPath = rp; multiSel = []; selState = keepState; renderTree(); renderView(); renderInspector(); }
      else { selPath = null; selState = ""; renderTree(); }
    }
    postHighlight();
    scheduleCssSync();
    suiSource = j.sui;
    setSuiEditor();
    endTextHistoryBurst();
    updateApply();
    if (!o.quiet) setStatus("applied");
    // The source is valid at this point; refresh diagnostics (clears errors).
    runLiveCheck();
    var mySeq = ++pvSeq;
    fetch("/__gui__/render", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(rj){
      if (mySeq !== pvSeq || !rj.ok) return;
      document.getElementById("prev").src = "/__gui__/preview?ts=" + Date.now();
    }).catch(function(){});
  }).catch(function(){
    setStatus("apply failed: server unavailable");
    updateApply();
  });
}
// findElementPath locates a node by kind + first class token and returns its
// path for selection restore after a code re-apply (null when gone).
function findElementPath(nodes, kind, cls) {
  for (var i = 0; i < nodes.length; i++) {
    var n = nodes[i];
    if (n.kind === kind && (!cls || (n.class || "").trim().split(/\s+/)[0] === cls)) return [i];
    var sub = (n.children || []).length ? findElementPath(n.children, kind, cls) : null;
    if (sub) return [i].concat(sub);
  }
  return null;
}
// Auto-apply: while the user edits the code, after a short pause the source is
// compiled automatically when it is valid, keeping the elements list, the
// inspector and the preview in sync without a manual Apply. Invalid sources
// leave everything untouched (the live diagnostics already flag the errors).
// Inspector edits are safe: setOut() keeps the textarea and suiSource in step,
// so a pending timer sees an unchanged source and does nothing.
var autoApplyTimer = null;
function scheduleAutoApply() {
  if (autoApplyTimer) clearTimeout(autoApplyTimer);
  autoApplyTimer = setTimeout(function(){
    autoApplyTimer = null;
    var src = document.getElementById("out").value;
    if (src === suiSource) { updateApply(); return; }
    applySuiCommon(src, { keepUndo: true, keepSel: true, quiet: true });
  }, 700);
}
function applySui() {
  applySuiCommon(document.getElementById("out").value, {});
}
function setupEditorScroll() {
  [["out", "sui_hl"]].forEach(function(p) {
    var ta = document.getElementById(p[0]);
    ta.addEventListener("scroll", function() {
      var hl = document.getElementById(p[1]);
      hl.scrollTop = ta.scrollTop;
      hl.scrollLeft = ta.scrollLeft;
      hideSuiAc();
    });
    ta.addEventListener("keydown", function(ev){ editorKeydown(ev); });
    ta.addEventListener("input", function(){ undoCaptureSoon(); updateBracketMatch(); });
    ta.addEventListener("click", function(){ updateBracketMatch(); });
    ta.addEventListener("mouseup", function(){ updateBracketMatch(); syncEditorToTreeSoon(); });
    ta.addEventListener("contextmenu", function(ev){ showEditorMenu(ev); });
    ta.addEventListener("mousemove", function(ev){ updateHoverDoc(ev); });
    ta.addEventListener("mouseleave", function(){ hideDoc(); });
    ta.addEventListener("keyup", function(e){
      // Caret movement with the arrow keys / Home / End should refresh the
      // bracket highlight too.
      if(e.key && e.key.indexOf('Arrow') === 0) { updateBracketMatch(); syncEditorToTreeSoon(); updateHoverDoc(); }
      else if(e.key === 'Home' || e.key === 'End') { updateBracketMatch(); syncEditorToTreeSoon(); updateHoverDoc(); }
    });
    ta.addEventListener("blur", function(){ setTimeout(hideSuiAc, 150); hideDoc(); });
  });
}

var BR_OPEN = { "(": ")", "[": "]", "{": "}" };
var BR_CLOSE = { ")": "(", "]": "[", "}": "{" };

// ---- editor undo/redo -----------------------------------------------------
// A textarea has native undo, but the editor performs most edits by assigning
// ta.value programmatically (bracket auto-close, smart Enter, Tab, snippets,
// autocomplete, designer regeneration), and every assignment wipes the native
// undo stack. So the editor keeps its own snapshot history: each change --
// typed or programmatic -- records the editor state, and Ctrl+Z / Ctrl+Y
// (Ctrl+U works as an undo alias; Ctrl+Shift+Z as redo) walk through it.
// Copy/cut/paste stay fully native (nothing ever intercepts Ctrl+C/X/V).
var __undo = { stack: [], idx: -1, timer: null, busy: false };
// undoSnap reads the editor state as a history entry.
function undoSnap() {
  var ta = document.getElementById("out");
  if (!ta) return null;
  return { v: ta.value, s: ta.selectionStart, e: ta.selectionEnd };
}
// undoReset restarts history from the editor's current content. Called on load
// and whenever the content is replaced externally (designer edit, apply).
function undoReset(resetValue) {
  var s = undoSnap();
  if (resetValue !== undefined) s = { v: resetValue, s: 0, e: 0 };
  if (!s) return;
  __undo.stack = [s];
  __undo.idx = 0;
  if (__undo.timer) { clearTimeout(__undo.timer); __undo.timer = null; }
}
// undoPrepareExternal resets history when the editor is about to be replaced
// with external content that differs from the current history top. A same-text
// call (e.g. re-entering the SUI tab) keeps the history intact.
function undoPrepareExternal(v) {
  var top = __undo.stack[__undo.idx];
  if (!top || top.v !== v) undoReset(v);
}
// undoCaptureSoon records the current state as a new history step, debounced
// so a burst of keystrokes coalesces into one undo step.
function undoCaptureSoon() {
  if (__undo.busy) return;
  if (__undo.timer) clearTimeout(__undo.timer);
  __undo.timer = setTimeout(undoFlush, 300);
}
// undoFlush pushes the pending state immediately (skipped if unchanged).
function undoFlush() {
  if (__undo.timer) { clearTimeout(__undo.timer); __undo.timer = null; }
  var s = undoSnap();
  if (!s) return;
  var top = __undo.stack[__undo.idx];
  if (top && top.v === s.v && top.s === s.s && top.e === s.e) return;
  // A new step abandons any redo tail.
  __undo.stack.length = __undo.idx + 1;
  __undo.stack.push(s);
  if (__undo.stack.length > 200) { __undo.stack.shift(); }
  __undo.idx = __undo.stack.length - 1;
}
// undoRestore applies a snapshot to the editor and repaints the overlay.
function undoRestore(s, label) {
  __undo.busy = true;
  var ta = document.getElementById("out");
  if (!ta) { __undo.busy = false; return; }
  ta.value = s.v;
  ta.selectionStart = Math.min(s.s, s.v.length);
  ta.selectionEnd = Math.min(s.e, s.v.length);
  highlightSui();
  renderOutline();
  updateApply();
  scheduleLiveCheck();
  setStatus(label);
  __undo.busy = false;
}
// editorUndo steps one history entry back (flushing any pending capture first,
// so the very latest typing lands on the stack before stepping away from it).
function editorUndo() {
  if (__undo.busy) return;
  undoFlush();
  if (__undo.idx <= 0) { setStatus("nothing to undo"); return; }
  __undo.idx--;
  undoRestore(__undo.stack[__undo.idx], "undone");
  hideSuiAc();
  closeSnippetMenu();
}
// editorRedo steps one history entry forward.
function editorRedo() {
  if (__undo.busy) return;
  undoFlush();
  if (__undo.idx >= __undo.stack.length - 1) { setStatus("nothing to redo"); return; }
  __undo.idx++;
  undoRestore(__undo.stack[__undo.idx], "ripetuto");
  hideSuiAc();
  closeSnippetMenu();
}

// editorKeydown implements the editor's typing intelligence: autocomplete
// navigation, bracket/quote auto-closing, smart Enter indentation and
// backspace that removes both halves of an empty pair.
function editorKeydown(ev) {
  var ta = document.getElementById("out");
  if (!ta) return;
  // Navigation shortcuts win even while the popup is open.
  if (ev.key === "F12") { ev.preventDefault(); if (ev.shiftKey) findReferences(); else gotoDefinition(); return; }
  if (ev.key === "F2") { ev.preventDefault(); renameSymbol(); return; }
  // Ctrl+Space opens the context-aware completion popup even when no word has
  // been typed yet. It is also the manual refresh while the popup is open.
  if ((ev.ctrlKey || ev.metaKey) && !ev.altKey && (ev.code === "Space" || ev.key === " " || ev.key === "Spacebar")) {
    ev.preventDefault();
    suiSuggest(true);
    return;
  }
  if ((ev.ctrlKey || ev.metaKey) && ev.key === "/") { ev.preventDefault(); toggleComment(); return; }
  // Undo/redo through the editor's own snapshot history. Ctrl+U is accepted as
  // an undo alias; Ctrl+Y and Ctrl+Shift+Z are redo. Copy/cut/paste
  // (Ctrl+C/X/V) are native textarea behavior and are never intercepted here.
  if ((ev.ctrlKey || ev.metaKey) && !ev.altKey) {
    var uk = (ev.key || "").toLowerCase();
    if ((uk === "z" || uk === "u") && !ev.shiftKey) {
      ev.preventDefault(); ev.stopPropagation(); undo(); return;
    }
    if (uk === "y" || (uk === "z" && ev.shiftKey)) {
      ev.preventDefault(); ev.stopPropagation(); redo(); return;
    }
  }
  // Autocomplete popup owns the arrow/enter/tab keys while visible.
  if (__suiAc.visible) {
    if (ev.key === 'ArrowDown') { ev.preventDefault(); moveSuiAc(1); return; }
    if (ev.key === 'ArrowUp') { ev.preventDefault(); moveSuiAc(-1); return; }
    if (ev.key === 'Enter' || ev.key === 'Tab') { if (__suiAc.items.length > 0) { ev.preventDefault(); applySuiAc(__suiAc.items[__suiAc.idx]); return; } }
    if (ev.key === 'Escape') { hideSuiAc(); closeSnippetMenu(); return; }
  }
  // The snippet menu owns arrows/enter/escape while open (keyboard snippet
  // picking: highlight moves, Enter inserts, Escape closes, preview follows).
  var snm = document.getElementById("snip-menu");
  if (snm && snm.style.display === "block") {
    if (ev.key === 'ArrowDown') { ev.preventDefault(); snipMove(1); return; }
    if (ev.key === 'ArrowUp') { ev.preventDefault(); snipMove(-1); return; }
    if (ev.key === 'Enter') { ev.preventDefault(); snipPick(); return; }
    if (ev.key === 'Escape') { ev.preventDefault(); closeSnippetMenu(); return; }
  }

  var start = ta.selectionStart, end = ta.selectionEnd, val = ta.value;

  // Auto-close brackets and quotes when a single character is typed.
  if (BR_OPEN[ev.key] && start === end) {
    ev.preventDefault();
    var close = BR_OPEN[ev.key];
    var sel = val.slice(start, end);
    val = val.slice(0, start) + ev.key + sel + close + val.slice(end);
    ta.value = val;
    ta.selectionStart = ta.selectionEnd = start + 1;
    if (sel) ta.selectionEnd = start + 1 + sel.length;
    onSuiEdit();
    return;
  }
  if ((ev.key === '"' || ev.key === "'") && start === end) {
    var next = val[start];
    // Skip over an existing identical quote instead of doubling it.
    if (next === ev.key) {
      ev.preventDefault();
      ta.selectionStart = ta.selectionEnd = start + 1;
      return;
    }
    var prev = val[start - 1];
    // Only auto-close when starting a new string (not inside a word).
    if (!/[A-Za-z0-9_]/.test(prev || "")) {
      ev.preventDefault();
      val = val.slice(0, start) + ev.key + ev.key + val.slice(end);
      ta.value = val;
      ta.selectionStart = ta.selectionEnd = start + 1;
      onSuiEdit();
      return;
    }
  }

  // Smart Enter: keep the current indentation; add a level after an opening
  // bracket and dedent a lone closing bracket typed on its own line.
  if (ev.key === 'Enter' && !ev.shiftKey) {
    var lineStart = val.lastIndexOf("\n", start - 1) + 1;
    var indent = (/^[ \t]*/.exec(val.slice(lineStart, start)) || [""])[0];
    var before = val.slice(0, start).replace(/[ \t]+$/, "");
    var lastCh = before[before.length - 1];
    var nextCh = val[start];
    ev.preventDefault();
    if (lastCh === "{" || lastCh === "[" || lastCh === "(") {
      if (nextCh === "}" || nextCh === "]" || nextCh === ")") {
        // Expand an empty pair into an indented block: {|} -> {\n\t|\n}
        var pair = val.slice(start, start + 1);
        val = val.slice(0, start) + "\n" + indent + "\t" + "\n" + indent + pair + val.slice(start + 1);
        ta.value = val;
        ta.selectionStart = ta.selectionEnd = start + 1 + indent.length + 1;
        onSuiEdit();
        return;
      }
      val = val.slice(0, start) + "\n" + indent + "\t" + val.slice(end);
      ta.value = val;
      ta.selectionStart = ta.selectionEnd = start + 1 + indent.length + 1;
      onSuiEdit();
      return;
    }
    // Default: newline + same indentation.
    val = val.slice(0, start) + "\n" + indent + val.slice(end);
    ta.value = val;
    ta.selectionStart = ta.selectionEnd = start + 1 + indent.length;
    onSuiEdit();
    return;
  }

  // Backspace between an empty pair deletes both halves.
  if (ev.key === 'Backspace' && start === end && start > 0) {
    var a = val[start - 1], b = val[start];
    if (BR_OPEN[a] && BR_OPEN[a] === b) {
      ev.preventDefault();
      ta.value = val.slice(0, start - 1) + val.slice(start + 1);
      ta.selectionStart = ta.selectionEnd = start - 1;
      onSuiEdit();
      return;
    }
    if ((a === '"' || a === "'") && a === b) {
      ev.preventDefault();
      ta.value = val.slice(0, start - 1) + val.slice(start + 1);
      ta.selectionStart = ta.selectionEnd = start - 1;
      onSuiEdit();
      return;
    }
  }

  // Tab inserts a tab (or indents the selection) rather than moving focus.
  if (ev.key === 'Tab' && !ev.shiftKey) {
    ev.preventDefault();
    if (start === end) {
      ta.value = val.slice(0, start) + "\t" + val.slice(end);
      ta.selectionStart = ta.selectionEnd = start + 1;
    } else {
      var ls = val.lastIndexOf("\n", start - 1) + 1;
      var block = val.slice(ls, end);
      var indented = block.replace(/\n/g, "\n\t");
      ta.value = val.slice(0, ls) + "\t" + indented + val.slice(end);
      ta.selectionStart = start + 1; ta.selectionEnd = end + indented.length - block.length + 1;
    }
    onSuiEdit();
    return;
  }
}

// updateBracketMatch highlights the bracket pair adjacent to the caret. It
// works on the visible, colour-plus-transparent-text editor: we can't paint a
// background inside the <pre> without shifting the layout, so we mark the
// matching characters with a dedicated span class in the overlay.
var __brMatch = { open: -1, close: -1 };
function updateBracketMatch() {
  var ta = document.getElementById("out");
  if (!ta) return;
  if (prefs.editorBracketMatch === false) {
    if (__brMatch.open >= 0) { __brMatch = { open: -1, close: -1 }; repaintBracketMatch(); }
    return;
  }
  var val = ta.value, pos = ta.selectionStart;
  var cand = -1;
  // Prefer the char just before the caret, then the one after.
  if (BR_OPEN[val[pos - 1]] || BR_CLOSE[val[pos - 1]]) cand = pos - 1;
  else if (BR_OPEN[val[pos]] || BR_CLOSE[val[pos]]) cand = pos;
  if (cand < 0) { if (__brMatch.open >= 0) { __brMatch = { open: -1, close: -1 }; repaintBracketMatch(); } return; }
  var m = matchBracket(val, cand);
  if (m < 0) { if (__brMatch.open >= 0) { __brMatch = { open: -1, close: -1 }; repaintBracketMatch(); } return; }
  var a = Math.min(cand, m), b = Math.max(cand, m);
  if (__brMatch.open === a && __brMatch.close === b) return;
  __brMatch = { open: a, close: b };
  repaintBracketMatch();
}

// matchBracket finds the index of the bracket matching the one at at.
function matchBracket(val, at) {
  var ch = val[at], dir, open, close;
  if (BR_OPEN[ch]) { dir = 1; open = ch; close = BR_OPEN[ch]; }
  else if (BR_CLOSE[ch]) { dir = -1; open = BR_CLOSE[ch]; close = ch; }
  else return -1;
  var depth = 0, i = at;
  while (i >= 0 && i < val.length) {
    var c = val[i];
    if (dir === 1) {
      if (c === open) depth++;
      else if (c === close) { depth--; if (depth === 0) return i; }
    } else {
      if (c === close) depth++;
      else if (c === open) { depth--; if (depth === 0) return i; }
    }
    i += dir;
  }
  return -1;
}

function repaintBracketMatch() { repaintEditor(); }
function repaintEditor() {
  var el = document.getElementById("out");
  if (!el) return;
  document.getElementById("sui_hl").innerHTML = supHighlight(el.value);
}

// ---- go to definition / find references ----------------------------------
// Overlay marks for cross-navigation, plus the line range of the element
// selected in the designer tree (keeps editor and tree in sync).
var __cmMarks = [];
var __elline = { line: -1 };
// __suiDiags holds the compiler diagnostics for the current editor text so
// they can be underlined inline. Each entry: {line, col, kind, msg}.
var __suiDiags = [];
// liveCheck debounces a POST to /__gui__/check while typing, updating the
// inline underlines. Cheap enough to run on every pause.
var __checkTimer = null;
var __checkSeq = 0;
function scheduleLiveCheck() {
  if (__checkTimer) clearTimeout(__checkTimer);
  __checkTimer = setTimeout(runLiveCheck, 350);
}
function runLiveCheck() {
  __checkTimer = null;
  var ta = document.getElementById("out");
  if (!ta) return;
  var seq = ++__checkSeq;
  var fd = new URLSearchParams();
  fd.set("sui", ta.value);
  fetch("/__gui__/check", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (seq !== __checkSeq) return;
    __suiDiags = j.diagDetails || [];
    renderProblems();
    repaintEditor();
  }).catch(function(){
    if (seq !== __checkSeq) return;
    __suiDiags = [];
    renderProblems();
    repaintEditor();
    setStatus("live check failed: server unavailable");
  });
}
function clearEditorMarks(repaint) {
  if (__cmMarks.length === 0) { if (repaint) repaintBracketMatch(); return; }
  __cmMarks = [];
  if (repaint !== false) repaintBracketMatch();
}
// lineColOffset returns the absolute character index of a (line, col) pair.
function lineColOffset(src, line, col) {
  var idx = 0;
  for (var i = 0; i < line; i++) {
    var nl = src.indexOf("\n", idx);
    if (nl < 0) return src.length;
    idx = nl + 1;
  }
  return idx + (col || 0);
}
// revealLine scrolls the editor to a line and places the caret there,
// selecting the given column range when provided.
function revealLine(line, colStart, colEnd) {
  var ta = document.getElementById("out");
  if (!ta) return;
  var src = ta.value;
  var off = lineColOffset(src, line, colStart || 0);
  ta.focus();
  ta.selectionStart = off;
  ta.selectionEnd = (colEnd !== undefined && colEnd !== null) ? lineColOffset(src, line, colEnd) : off;
  // Approximate scroll: keep the target line near the vertical middle.
  var lh = editorLineHeightPx();
  ta.scrollTop = Math.max(0, line * lh - ta.clientHeight / 2 + lh);
  var hl = document.getElementById("sui_hl");
  if (hl) hl.scrollTop = ta.scrollTop;
}
// wordAt returns the identifier (optionally dotted) surrounding an index.
function wordAt(src, idx) {
  var a = idx, b = idx;
  while (a > 0 && /[A-Za-z0-9_.]/.test(src[a - 1])) a--;
  while (b < src.length && /[A-Za-z0-9_.]/.test(src[b])) b++;
  return { word: src.slice(a, b), start: a, end: b };
}
// gotoDefinition jumps to the declaration of the word at the caret.
function gotoDefinition() {
  var ta = document.getElementById("out");
  var src = ta.value, pos = ta.selectionStart;
  var w = wordAt(src, pos > 0 && !/[A-Za-z0-9_.]/.test(src[pos]) ? pos - 1 : pos);
  var name = w.word.replace(/^.*\./, "");
  if (!name) { setStatus("no symbol under the cursor"); return; }
  var defs = findDefsInSource(src, name);
  if (defs.length === 0) { setStatus("definition of '" + name + "' not found"); return; }
  var d = defs[0];
  clearEditorMarks(false);
  __cmMarks = [{ start: lineColOffset(src, d.line, 0), end: lineColOffset(src, d.line, 9999), cls: "cm-def" }];
  revealLine(d.line, 0, 0);
  repaintBracketMatch();
  setStatus("go to definition: " + name + " (line " + (d.line + 1) + ")");
}
// findReferences highlights every use of the word at the caret.
function findReferences() {
  var ta = document.getElementById("out");
  var src = ta.value, pos = ta.selectionStart;
  var w = wordAt(src, pos > 0 && !/[A-Za-z0-9_.]/.test(src[pos]) ? pos - 1 : pos);
  var name = w.word.replace(/^.*\./, "");
  if (!name) { setStatus("no symbol under the cursor"); return; }
  var refs = findRefsInSource(src, name);
  if (refs.length === 0) { setStatus("no references to '" + name + "'"); return; }
  __cmMarks = refs.map(function(r) {
    var st = lineColOffset(src, r.line, r.col);
    var isDef = findDefsInSource(src, name).some(function(d){ return d.line === r.line; });
    return { start: st, end: st + name.length, cls: isDef ? "cm-def" : "cm-ref" };
  });
  repaintBracketMatch();
  setStatus(refs.length + " references to '" + name + "'");
}

// renameSymbol renames the identifier under the caret across the whole file.
// It targets declared functions/states/views so accidental edits are avoided.
function renameSymbol() {
  var ta = document.getElementById("out");
  var src = ta.value, pos = ta.selectionStart;
  var w = wordAt(src, pos > 0 && !/[A-Za-z0-9_.]/.test(src[pos]) ? pos - 1 : pos);
  var name = w.word.replace(/^.*\./, "");
  if (!name) { setStatus("no symbol under the cursor"); return; }
  var defs = findDefsInSource(src, name);
  if (defs.length === 0) { setStatus("'" + name + "' is not a renameable declaration"); return; }
  var refs = findRefsInSource(src, name);
  askPrompt("Rename '" + name + "'", "New name", name, function(nn) {
    nn = nn.replace(/[^A-Za-z0-9_]/g, "");
    if (!nn || nn === name) return;
    var esc = name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    var re = new RegExp("\\b" + esc + "\\b", "g");
    var count = 0;
    var out = src.replace(re, function() { count++; return nn; });
    ta.value = out;
    ta.focus();
    onSuiEdit();
    setStatus("renamed " + count + " uses of '" + name + "' to '" + nn + "'");
  }, refs.length + " occurrences in the file will be replaced");
}

// ---- document outline (document symbols) ---------------------------------
// A compact, clickable list of the top-level declarations in the file. It is
// built by scanning the editor text (not the model) so it works while typing
// and does not depend on a successful compile.
function parseOutline(src) {
  var out = [], lines = src.split("\n");
  for (var i = 0; i < lines.length; i++) {
    var l = lines[i];
    var m;
    if ((m = /^\s*async\s+fn\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(?:->\s*([A-Za-z0-9_,\s\[\]]+))?/.exec(l)))
      out.push({ kind: "async fn", name: m[1], line: i, detail: "(" + m[2].trim() + ")" + (m[3] ? " -> " + m[3].trim() : "") });
    else if ((m = /^\s*fn\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(?:->\s*([A-Za-z0-9_,\s\[\]]+))?/.exec(l)))
      out.push({ kind: "fn", name: m[1], line: i, detail: "(" + m[2].trim() + ")" + (m[3] ? " -> " + m[3].trim() : "") });
    else if ((m = /^\s*state\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?::\s*([A-Za-z0-9_\[\]]+))?\s*(?:=\s*(.*))?$/.exec(l)))
      out.push({ kind: "state", name: m[1], line: i, detail: (m[2] ? m[2] + " " : "") + (m[3] ? "= " + m[3].trim() : "") });
    else if ((m = /^\s*view\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)/.exec(l)))
      out.push({ kind: "view", name: m[1], line: i, detail: "(" + m[2].trim() + ")" });
    else if ((m = /^\s*view\s+([A-Za-z_][A-Za-z0-9_]*)/.exec(l)))
      out.push({ kind: "view", name: m[1], line: i, detail: "" });
    else if ((m = /^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)/.exec(l)))
      out.push({ kind: "type", name: m[1], line: i, detail: "" });
    else if ((m = /^\s*import\s+\{([^}]*)\}\s+from\s+"([^"]+)"/.exec(l)))
      out.push({ kind: "import", name: m[2], line: i, detail: m[1].trim() });
  }
  return out;
}
function renderOutline() {
  var box = document.getElementById("outline");
  if (!box || box.style.display === "none") return;
  var ta = document.getElementById("out");
  var syms = parseOutline(ta ? ta.value : "");
  box.innerHTML = "";
  if (!syms.length) { var e = document.createElement("div"); e.className = "ol-empty"; e.textContent = "No symbols."; box.appendChild(e); return; }
  syms.forEach(function(s) {
    var d = document.createElement("div"); d.className = "ol-item";
    var k = document.createElement("span"); k.className = "ol-kind"; k.textContent = s.kind;
    var n = document.createElement("span"); n.className = "ol-name"; n.textContent = s.name;
    var det = document.createElement("span"); det.className = "ol-detail"; det.textContent = s.detail || "";
    d.appendChild(k); d.appendChild(n); d.appendChild(det);
    d.title = "Go to line " + (s.line + 1);
    d.onclick = function() { revealLine(s.line, 0, 0); };
    box.appendChild(d);
  });
}
function toggleOutlinePanel() {
  var box = document.getElementById("outline");
  if (!box) return;
  var on = box.style.display !== "none";
  box.style.display = on ? "none" : "block";
  if (!on) renderOutline();
}

// ---- hover docs ----------------------------------------------------------
// A floating panel showing signature and description for a builtin/stdlib
// call, a declared function or state under the pointer.
var __docEl = null;
function hideDoc() { if (__docEl) __docEl.style.display = "none"; }
function showDoc(html, x, y) {
  if (!__docEl) { __docEl = document.createElement("div"); __docEl.className = "sui-doc"; document.body.appendChild(__docEl); }
  __docEl.innerHTML = html;
  __docEl.style.display = "block";
  // x/y are scaled viewport pixels; divide by the root zoom (see menuZoom).
  var z = menuZoom();
  var w = __docEl.offsetWidth, h = __docEl.offsetHeight;
  var vw = window.innerWidth / z;
  __docEl.style.left = Math.max(4, Math.min(vw - w - 12, x / z + 12)) + "px";
  __docEl.style.top = Math.max(8, y / z - h - 12) + "px";
}
// docHtml builds the hover panel content for a symbol name.
function docHtml(name) {
  var d = DOCS[name];
  if (d) {
    return '<div class="doc-sig">' + escHtml(d.sig) + (d.async ? ' <span style="color:#e0af68">async</span>' : '') + '</div>' +
           '<div class="doc-ret">&#8594; ' + escHtml(d.ret) + '</div>' +
           '<div class="doc-desc">' + escHtml(d.desc) + '</div>';
  }
  var f = declaredFns().filter(function(fn){ return fn.name === name; })[0];
  if (f) {
    return '<div class="doc-sig">' + (f.async ? 'async ' : '') + 'fn ' + escHtml(f.name) + '(' + escHtml(f.params || '') + ') -&gt; ' + escHtml(f.ret || 'null') + '</div>' +
           '<div class="doc-desc">function declared in the file</div>';
  }
  var s = declaredStates().filter(function(sd){ return sd.name === name; })[0];
  if (s) {
    return '<div class="doc-sig">state ' + escHtml(s.name) + (s.type ? ': ' + escHtml(s.type) : '') + '</div>' +
           '<div class="doc-desc">' + escHtml(s.init !== undefined && s.init !== '' ? ('initial value ' + s.init) : 'reactive state') + '</div>';
  }
  return null;
}
// diagRangeFor returns the underlined byte range of a diagnostic — the same
// range overlayHighlight paints — so the hover tooltip matches exactly what is
// marked wavy in the editor.
function diagRangeFor(code, d) {
  var start = lineColOffset(code, d.line - 1, Math.max(0, (d.col || 1) - 1));
  var end = start;
  while (end < code.length && /[A-Za-z0-9_.]/.test(code[end])) end++;
  if (end === start) { while (end < code.length && code[end] !== "\n") end++; }
  return { start: start, end: end };
}
// __docCanvas measures real glyph widths for the editor's font. The editor uses
// tabs (rendered at tabSize columns) and a proportional-ish font, so relying on
// a fixed per-character width misses the marked range — especially on indented
// lines, where a single tab is several cell widths off.
var __docCanvas = null;
// textWidthUpTo measures the pixel width of lineText up to column col, giving
// each tab its effective tabSize-column width.
function textWidthUpTo(ctx, lineText, col, tabCols) {
  var w = 0, seg = "";
  for (var i = 0; i < col && i < lineText.length; i++) {
    if (lineText[i] === "\t") {
      if (seg) { w += ctx.measureText(seg).width; seg = ""; }
      w += ctx.measureText(" ").width * tabCols;
    } else {
      seg += lineText[i];
    }
  }
  if (seg) w += ctx.measureText(seg).width;
  return w;
}
// diagHitAt returns the diagnostic whose underlined range contains the mouse
// position, or null. It derives the row from the real line height and the
// column from measured text widths (tab-aware), matching what is painted.
function diagHitAt(ta, clientX, clientY) {
  var rect = ta.getBoundingClientRect();
  var z = menuZoom() || 1;
  var mx = (clientX - rect.left) / z + ta.scrollLeft;
  var my = (clientY - rect.top) / z + ta.scrollTop;
  var cs = getComputedStyle(ta);
  var lh = parseFloat(cs.lineHeight) || 21.7;
  var gline = Math.floor(my / lh);
  if (gline < 0) return null;
  var lineStart = lineColOffset(ta.value, gline, 0);
  var lineEnd = ta.value.indexOf("\n", lineStart); if (lineEnd < 0) lineEnd = ta.value.length;
  var lineText = ta.value.slice(lineStart, lineEnd);
  if (!__docCanvas) __docCanvas = document.createElement("canvas");
  var ctx = __docCanvas.getContext("2d");
  ctx.font = cs.fontStyle + " " + cs.fontWeight + " " + cs.fontSize + " " + cs.fontFamily;
  var tabCols = parseFloat(cs.tabSize) || 8;
  var PADX = 12; // horizontal padding shared by textarea and overlay
  for (var i = 0; i < __suiDiags.length; i++) {
    var d = __suiDiags[i];
    if (d.line - 1 !== gline) continue;
    var rr = diagRangeFor(ta.value, d);
    var sCol = rr.start - lineStart, eCol = rr.end - lineStart;
    var wS = textWidthUpTo(ctx, lineText, sCol, tabCols) + PADX;
    var wE = textWidthUpTo(ctx, lineText, eCol, tabCols) + PADX;
    // Small slack both sides: the marked keyword may be bold (wider) and the
    // marked span may be a full line when no word follows the position.
    if (mx >= wS - 6 && mx <= wE + 10) return d;
  }
  return null;
}
function updateHoverDoc(ev) {
  var ta = document.getElementById("out");
  if (!ta) { hideDoc(); return; }
  // 1) Diagnostic hover: hit-test the character under the mouse against the
  //    underlined compiler diagnostics. Works whether or not the editor has
  //    focus, so a wavy underline immediately explains itself.
  if (ev && __suiDiags && __suiDiags.length) {
    var d = diagHitAt(ta, ev.clientX, ev.clientY);
    if (d) {
      var arect = ta.getBoundingClientRect();
      var alh = parseFloat(getComputedStyle(ta).lineHeight) || 21.7;
      var kindTag = (d.kind === "warning") ? "warning" : (d.kind === "error" ? "error" : d.kind);
      var html = '<div class="doc-kind ' + (d.kind === "warning" ? "warn" : "err") + '">' + escHtml(kindTag) + '</div>';
      html += '<div class="doc-desc">' + escHtml(d.msg || "") + '</div>';
      if (d.hint) html += '<div class="doc-desc hintline">' + escHtml(d.hint) + '</div>';
      showDoc(html, ev.clientX, arect.top + (d.line - 1) * alh + 10);
      return;
    }
  }
  // 2) Symbol documentation (only while the user keeps the caret in the editor).
  if (document.activeElement !== ta) { hideDoc(); return; }
  var src = ta.value, pos = ta.selectionStart;
  var probe = pos < src.length && /[A-Za-z0-9_.]/.test(src[pos]) ? pos : pos - 1;
  if (probe < 0 || !/[A-Za-z0-9_.]/.test(src[probe] || "")) { hideDoc(); return; }
  var w = wordAt(src, probe);
  // Prefer the dotted stdlib name (http.get) when present.
  var name = w.word;
  var html = docHtml(name);
  if (!html) { var short = name.replace(/^.*\./, ""); html = docHtml(short); if (!html) { hideDoc(); return; } }
  var rect = ta.getBoundingClientRect();
  // Estimate the caret position from the line/column.
  var before = src.slice(0, w.start);
  var line = before.split("\n").length - 1;
  var col = w.start - (before.lastIndexOf("\n") + 1);
  var lh = editorLineHeightPx(), cw = editorCharWidthPx();
  showDoc(html, rect.left + 12 + col * cw - ta.scrollLeft, rect.top + 10 + (line + 1) * lh - ta.scrollTop);
}

// ---- editor context menu (right-click) -----------------------------------
function showEditorMenu(ev) {
  ev.preventDefault();
  closeSnippetMenu();
  closeContextMenus();
  var m = document.createElement("div");
  m.id = "ed-menu"; m.className = "sui-ac"; m.style.display = "block"; m.style.minWidth = "170px";
  var items = [
    { label: "Go to definition", fn: gotoDefinition, key: "F12" },
    { label: "Find references", fn: findReferences, key: "Shift+F12" },
    { label: "Rename symbol", fn: renameSymbol, key: "F2" },
    { label: "Toggle comment", fn: toggleComment, key: "Ctrl+/" }
  ];
  items.forEach(function(it) {
    var d = document.createElement("div"); d.className = "sui-ac-item";
    d.innerHTML = '<span>' + it.label + '</span><span class="ac-kind" style="margin-left:auto">' + it.key + '</span>';
    d.onclick = function(){ m.remove(); it.fn(); };
    m.appendChild(d);
  });
  document.body.appendChild(m);
  var z = menuZoom();
  m.style.left = (ev.pageX / z) + "px";
  m.style.top = (ev.pageY / z) + "px";
  var close = function(){ m.remove(); document.removeEventListener("click", close); };
  setTimeout(function(){ document.addEventListener("click", close); }, 0);
}

// ---- toggle line comment -------------------------------------------------
function toggleComment() {
  var ta = document.getElementById("out");
  var src = ta.value;
  var s = ta.selectionStart, e = ta.selectionEnd;
  var ls = src.lastIndexOf("\n", s - 1) + 1;
  var le = src.indexOf("\n", e); if (le < 0) le = src.length;
  var block = src.slice(ls, le);
  var lines = block.split("\n");
  var commented = lines.every(function(l){ return /^\s*#/.test(l) || l.trim() === ""; });
  var outLines = lines.map(function(l) {
    if (l.trim() === "") return l;
    if (commented) return l.replace(/^(\s*)#\s?/, "$1");
    return l.replace(/^(\s*)/, "$1# ");
  });
  var nb = outLines.join("\n");
  ta.value = src.slice(0, ls) + nb + src.slice(le);
  ta.focus();
  ta.selectionStart = ls; ta.selectionEnd = ls + nb.length;
  onSuiEdit();
}

// ---- tree <-> editor synchronisation -------------------------------------
// Editor caret position -> select the element block that contains it.
function syncEditorToTree() {
  var ta = document.getElementById("out");
  if (!ta) return;
  var src = ta.value, pos = ta.selectionStart;
  var before = src.slice(0, pos);
  var line = before.split("\n").length - 1;
  var loc = elementBlockAtLine(line);
  if (loc) {
    selPath = loc.path.slice(); selCell = null; selRow = null; selCol = null;
    renderTree(); fillPanel(); renderInspector(); postHighlight();
  }
}
// elementBlockAtLine returns the element whose declaration line is at or
// before the line, using the line range recorded on each element when the
// designer renders the model. Returns null when nothing matches.
function elementBlockAtLine(line) {
  if (!__elLines || !__elLines.length) return null;
  var best = null;
  for (var i = 0; i < __elLines.length; i++) {
    var it = __elLines[i];
    if (it.start <= line && (it.end === undefined || line <= it.end)) {
      if (!best || it.start >= best.start) best = it;
    }
  }
  return best;
}
var __elLines = [];
// syncEditorToTreeSoon debounces the caret -> tree selection update so moving
// the caret does not rebuild the tree on every key repeat.
var __syncTimer = null;
function syncEditorToTreeSoon() {
  if (__syncTimer) clearTimeout(__syncTimer);
  __syncTimer = setTimeout(function(){ __syncTimer = null; syncEditorToTree(); }, 120);
}


renderPalette();
renderInsertMenu();
renderTemplates();
applyTheme();
applyPanelVisibility();
loadRemotePrefs();
loadRemoteFonts();
loadCurrentProject();
setupEditorScroll();
initMenus();
initDraggableBars();
setBp("desktop");

  try { makeDraggable(document.getElementById('sf-panel'), document.getElementById('sf-hdr')); } catch(e) {}
  try { makeResizable(document.getElementById('sf-panel'), document.getElementById('sf-resize')); } catch(e) {}
  try { wireChartDataEditor(); } catch(e) {}
// Inspector resize: the whole right panel can be widened/narrowed at will by
// dragging its left edge. The width is remembered across reloads.
(function(){
  var right = document.getElementById("right"), grip = document.getElementById("right-grip");
  if (!right || !grip) return;
  try {
    var st = parseInt(localStorage.getItem("sui.rightW") || "", 10);
    if (st >= 210) right.style.width = st + "px";
  } catch (e) {}
  grip.addEventListener("mousedown", function(e){
    e.preventDefault();
    grip.classList.add("drag");
    var startX = e.clientX, startW = right.offsetWidth;
    function move(ev){
      var w = Math.max(210, Math.min(760, startW - (ev.clientX - startX)));
      right.style.width = w + "px";
    }
    function up(){
      document.removeEventListener("mousemove", move);
      document.removeEventListener("mouseup", up);
      grip.classList.remove("drag");
      try { localStorage.setItem("sui.rightW", String(right.offsetWidth)); } catch (e) {}
    }
    document.addEventListener("mousemove", move);
    document.addEventListener("mouseup", up);
  });
})();
document.getElementById("prev").onload = function() { postHighlight(); updateGuides(); };
window.addEventListener("error", function(e) { setStatus("JS ERROR: " + (e.message || e)); });
window.addEventListener("unhandledrejection", function(e) { setStatus("PROMISE ERROR: " + (e.reason && e.reason.message ? e.reason.message : e.reason)); });
// Restore the in-memory project kept by the server (survives a page reload).
function loadCurrentProject() {
  fetch("/__gui__/current").then(function(r){ return r.json(); }).then(function(j){
    if (j && j.ok && j.sui) {
      var fd = new URLSearchParams(); fd.set("sui", j.sui);
      fetch("/__gui__/open", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(m){
        if (m && m.ok) { loadTree(m); autoPreview(); scheduleLiveCheck(); }
        else { renderAll(); scheduleLiveCheck(); }
      }).catch(function(){ renderAll(); });
    } else {
      renderAll();
    }
  }).catch(function(){ renderAll(); });
}
</script>
</body>
</html>
`
