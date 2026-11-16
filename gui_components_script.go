package main

// guiComponentLibraryJS is served as /__gui__/components.js before the
// designer's main script. The catalog stays external to the page so the main
// script does not grow for every library feature.
const guiComponentLibraryJS = `// ---- external component library -------------------------------------------
// components/library.json is read and validated by the GUI server at startup.
// It may add presets over the built-in node kinds and may override a kind's
// palette entry (label/group). The editor never trusts the file: the server
// rejects the whole library on any schema, CSS or compile error and the
// built-in catalog remains active.
var COMPONENT_LIBRARY = null;
var COMPONENT_LIBRARY_ERROR = "";
function componentLibraryGroups() {
  return (COMPONENT_LIBRARY && COMPONENT_LIBRARY.groups) ? COMPONENT_LIBRARY.groups : [];
}
function loadComponentLibrary(reload) {
  var url = "/__gui__/components" + (reload ? "?reload=1" : "");
  return fetch(url).then(function(r){ return r.json(); }).then(function(j){
    if (j && j.ok) {
      COMPONENT_LIBRARY = { groups: j.groups || [], source: j.source || "", warnings: j.warnings || [] };
      COMPONENT_LIBRARY_ERROR = "";
      if (reload) {
        var n = 0;
        componentLibraryGroups().forEach(function(g){ n += (g.components || []).length; });
        setStatus("Component library reloaded: " + n + " component(s)" + (j.source ? " from " + j.source : ""));
        renderInsertMenu();
      }
    } else {
      COMPONENT_LIBRARY = null;
      COMPONENT_LIBRARY_ERROR = (j && j.error) ? String(j.error) : "invalid component library";
      setStatus("Component library rejected; built-in components remain available. " + COMPONENT_LIBRARY_ERROR);
    }
    return j;
  }).catch(function(err){
    COMPONENT_LIBRARY = null;
    COMPONENT_LIBRARY_ERROR = String(err);
    setStatus("Component library unavailable; built-in components remain available.");
    return null;
  });
}
function nodeFromLibraryNode(c) {
  var n = newNodeForKind(c.kind);
  if (c.text) n.text = c.text;
  if (c.class) n.class = c.class;
  if (c.attrs) {
    for (var k in c.attrs) {
      if (!Object.prototype.hasOwnProperty.call(c.attrs, k)) continue;
      n[k] = c.attrs[k];
    }
  }
  if (c.children && c.children.length && canNest(n.kind)) {
    n.children = c.children.map(nodeFromLibraryNode);
  }
  return n;
}
function ensureLibraryCss(c) {
  if (!c || !c.css || !c.class) return;
  var key = "sui-library:" + c.class;
  var escKey = key.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  var re = new RegExp("\\n?/\\* " + escKey + " \\*/[\\s\\S]*?/\\* /" + escKey + " \\*/");
  var block = "/* " + key + " */\n" + c.css + "\n/* /" + key + " */";
  var css = String(root.css || "");
  if (re.test(css)) css = css.replace(re, "\n" + block);
  else css = (css.trim() ? css.trim() + "\n" : "") + block;
  root.css = css.trim() + "\n";
}
function addLibraryComponent(c) {
  if (!c || !c.kind) return;
  withUndo(function(){
    ensureLibraryCss(c);
    insertNodeAtSelection(nodeFromLibraryNode(c), c.kind);
  }, "insert " + (c.label || c.id));
  scheduleCssSync();
  setStatus("Inserted " + (c.label || c.id) + " from the component library.");
}
function allInsertEntries() {
  var out = [];
  var overridden = {};
  var builtin = {};
  componentLibraryGroups().forEach(function(g){
    (g.components || []).forEach(function(c){
      out.push({ kind: c.kind, label: c.label || c.id, preset: c });
      if (c.override) overridden[c.kind] = true;
    });
  });
  PALETTE.forEach(function(g){
    (g.tags || []).forEach(function(t){ if (!overridden[t] && !builtin[t]) { builtin[t] = true; out.push({ kind: t, label: PALETTE_LABEL[t] || t, preset: null }); } });
    (g.subs || []).forEach(function(s){ (s.tags || []).forEach(function(t){ if (!overridden[t] && !builtin[t]) { builtin[t] = true; out.push({ kind: t, label: PALETTE_LABEL[t] || t, preset: null }); } }); });
  });
  return out;
}
`
