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
      queueLibraryAudit();
      if (document.getElementById("lib-nav") && document.getElementById("lib-nav").classList.contains("on")) renderLibraryNavigator();
    } else {
      COMPONENT_LIBRARY = null;
      COMPONENT_LIBRARY_ERROR = (j && j.error) ? String(j.error) : "invalid component library";
      setStatus("Component library rejected; built-in components remain available. " + COMPONENT_LIBRARY_ERROR);
      queueLibraryAudit();
      if (document.getElementById("lib-nav") && document.getElementById("lib-nav").classList.contains("on")) renderLibraryNavigator();
    }
    return j;
  }).catch(function(err){
    COMPONENT_LIBRARY = null;
    COMPONENT_LIBRARY_ERROR = String(err);
    setStatus("Component library unavailable; built-in components remain available.");
    queueLibraryAudit();
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
function libraryCssBlock(c) {
  var key = "sui-library:" + c.class;
  var meta = (c.id ? " component=" + c.id : "") + (c.kind ? " kind=" + c.kind : "");
  return "/* " + key + meta + " */\n" + (c.css || "") + "\n/* /" + key + " */";
}
function ensureLibraryCss(c) {
  if (!c || !c.class) return;
  var key = "sui-library:" + c.class;
  var escKey = key.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  var re = new RegExp("\\n?/\\* " + escKey + "[^*]*\\*/[\\s\\S]*?/\\* /" + escKey + " \\*/");
  var block = libraryCssBlock(c);
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
var LIBRARY_REF_WARNINGS = [];
var LIB_NAV_SEL = "";
var LIB_NAV_GROUP = "";
var LIB_NAV_PREVIEW_ID = "";
function libraryAllComponents() {
  var out = [];
  componentLibraryGroups().forEach(function(g){
    (g.components || []).forEach(function(c){ out.push({ c: c, group: g }); });
  });
  return out;
}
function libraryComponentById(id) {
  var all = libraryAllComponents();
  for (var i = 0; i < all.length; i++) if (all[i].c.id === id) return all[i].c;
  return null;
}
function libraryDefaultForKind(kind) {
  var all = libraryAllComponents();
  var first = null;
  for (var i = 0; i < all.length; i++) {
    if (all[i].c.kind !== kind) continue;
    if (!first) first = all[i].c;
    if (all[i].c.default) return all[i].c;
  }
  return first;
}
function libraryIsDefault(c) {
  if (!c) return false;
  var explicit = null;
  var all = libraryAllComponents();
  for (var i = 0; i < all.length; i++) {
    if (all[i].c.kind === c.kind && all[i].c.default) { explicit = all[i].c; break; }
  }
  if (explicit) return explicit.id === c.id;
  return libraryDefaultForKind(c.kind) === c;
}
function walkLibraryNodes(cb) {
  function walk(list) {
    (list || []).forEach(function(n){
      cb(n);
      if (n.children) walk(n.children);
    });
  }
  walk(root.children);
  (root.pages || []).forEach(function(p){ if (p.children !== root.children) walk(p.children); });
}
function auditLibraryReferences() {
  LIBRARY_REF_WARNINGS = [];
  if (!root || typeof root.css !== "string" || !root.css) return;
  var css = root.css;
  var re = /\/\* sui-library:([A-Za-z_][A-Za-z0-9_-]*) component=([a-z][a-z0-9-]{0,39}) kind=([A-Za-z0-9_-]+) \*\/([\s\S]*?)\/\* \/sui-library:\1 \*\//g;
  var refs = [];
  var m;
  while ((m = re.exec(css)) !== null) refs.push({ cls: m[1], id: m[2], kind: m[3], full: m[0] });
  var warnings = [];
  var changed = false;
  refs.forEach(function(ref){
    if (libraryComponentById(ref.id)) return;
    var fb = libraryDefaultForKind(ref.kind);
    if (!fb) {
      warnings.push("Library component '" + ref.id + "' (class ." + ref.cls + ") is missing and there is no default for " + ref.kind + "; the current styles are kept.");
      return;
    }
    warnings.push("Library component '" + ref.id + "' (class ." + ref.cls + ") is missing; using the default " + (fb.label || fb.id) + " (" + fb.kind + ").");
    if (fb.class && fb.class !== ref.cls) {
      var existing = css.indexOf("/* sui-library:" + fb.class);
      css = css.replace(ref.full, existing >= 0 ? "" : libraryCssBlock(fb));
    } else {
      css = css.replace(ref.full, fb.class ? libraryCssBlock(fb) : "");
    }
    if (fb.class !== ref.cls) {
      walkLibraryNodes(function(n){
        var cs = String(n.class || "").trim().split(/\s+/).filter(Boolean);
        var at = cs.indexOf(ref.cls);
        if (at < 0) return;
        if (fb.class) cs[at] = fb.class;
        else cs.splice(at, 1);
        n.class = cs.join(" ");
      });
    }
    changed = true;
  });
  LIBRARY_REF_WARNINGS = warnings;
  if (warnings.length) setStatus(warnings[0] + (warnings.length > 1 ? " (+" + (warnings.length - 1) + " more)" : ""));
  if (changed) {
    root.css = css.trim() + (css.trim() ? "\n" : "");
    setOut();
    renderAll();
    autoPreview();
  }
  if (document.getElementById("lib-nav") && document.getElementById("lib-nav").classList.contains("on")) renderLibraryNavigator();
}
var __libraryAuditTimer = null;
function queueLibraryAudit() {
  if (__libraryAuditTimer) return;
  __libraryAuditTimer = setTimeout(function(){
    __libraryAuditTimer = null;
    auditLibraryReferences();
  }, 900);
}
function libNavEscape(s) {
  return String(s || "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}
function libraryNavSelected() {
  var all = libraryAllComponents();
  for (var i = 0; i < all.length; i++) if (all[i].c.id === LIB_NAV_SEL) return all[i];
  return all.length ? all[0] : null;
}
function openLibraryNavigator() {
  var nav = document.getElementById("lib-nav");
  if (!nav) return;
  nav.classList.add("on");
  if (!LIB_NAV_SEL) {
    var all = libraryAllComponents();
    if (all.length) LIB_NAV_SEL = all[0].c.id;
  }
  renderLibraryNavigator();
  var s = document.getElementById("lib-search");
  if (s) s.focus();
}
function closeLibraryNavigator() {
  var nav = document.getElementById("lib-nav");
  if (nav) nav.classList.remove("on");
}
function renderLibraryNavigator() {
  var src = document.getElementById("lib-source");
  if (src) src.textContent = COMPONENT_LIBRARY ? (COMPONENT_LIBRARY.source || "component library") : "";
  var st = document.getElementById("lib-status");
  if (st) {
    if (COMPONENT_LIBRARY_ERROR) {
      st.textContent = "Library rejected: " + COMPONENT_LIBRARY_ERROR;
      st.style.color = "#f87171";
    } else if (LIBRARY_REF_WARNINGS.length) {
      st.textContent = LIBRARY_REF_WARNINGS.join(" ");
      st.style.color = "#fbbf24";
    } else {
      st.textContent = COMPONENT_LIBRARY ? "" : "No external component library loaded. Add components/library.json and reload.";
      st.style.color = "";
    }
  }
  renderLibraryGroups();
  renderLibraryList();
  renderLibraryDetail();
}
function renderLibraryGroups() {
  var box = document.getElementById("lib-groups");
  if (!box) return;
  box.textContent = "";
  var all = libraryAllComponents();
  var groups = [{ id: "", label: "All", n: all.length }];
  componentLibraryGroups().forEach(function(g){
    groups.push({ id: g.id, label: g.label, n: (g.components || []).length });
  });
  groups.forEach(function(g){
    var b = document.createElement("button");
    b.type = "button";
    b.className = "lib-group" + (LIB_NAV_GROUP === g.id ? " on" : "");
    b.textContent = g.label + " (" + g.n + ")";
    b.onclick = function(){ LIB_NAV_GROUP = g.id; renderLibraryGroups(); renderLibraryList(); renderLibraryDetail(); };
    box.appendChild(b);
  });
}
function renderLibraryList() {
  var box = document.getElementById("lib-list");
  if (!box) return;
  box.textContent = "";
  var search = document.getElementById("lib-search");
  var q = search && search.value ? search.value.toLowerCase() : "";
  var all = libraryAllComponents().filter(function(e){
    if (LIB_NAV_GROUP && e.group.id !== LIB_NAV_GROUP) return false;
    var hay = (e.c.label + " " + e.c.id + " " + e.c.kind + " " + (e.c.description || "") + " " + e.group.label).toLowerCase();
    return !q || hay.indexOf(q) >= 0;
  });
  if (!all.length) {
    var empty = document.createElement("div");
    empty.className = "hint";
    empty.textContent = COMPONENT_LIBRARY ? "No components match the filter." : "The component library is not available.";
    box.appendChild(empty);
    return;
  }
  if (!all.some(function(e){ return e.c.id === LIB_NAV_SEL; })) LIB_NAV_SEL = all[0].c.id;
  all.forEach(function(e){
    var b = document.createElement("button");
    b.type = "button";
    b.className = "lib-item" + (e.c.id === LIB_NAV_SEL ? " on" : "");
    var title = document.createElement("span");
    title.className = "lib-item-title";
    title.textContent = e.c.label || e.c.id;
    var meta = document.createElement("span");
    meta.className = "lib-item-meta";
    meta.textContent = e.group.label + " · " + e.c.kind;
    b.appendChild(title);
    b.appendChild(meta);
    if (libraryIsDefault(e.c)) {
      var def = document.createElement("span");
      def.className = "lib-default";
      def.textContent = "Default";
      b.appendChild(def);
    }
    b.onclick = function(){ LIB_NAV_SEL = e.c.id; renderLibraryList(); renderLibraryDetail(); };
    box.appendChild(b);
  });
}
function renderLibraryDetail() {
  var title = document.getElementById("lib-detail-title");
  var badges = document.getElementById("lib-badges");
  var desc = document.getElementById("lib-detail-desc");
  var btn = document.getElementById("lib-insert");
  if (!title || !badges) return;
  badges.textContent = "";
  var sel = libraryNavSelected();
  if (!sel) {
    title.textContent = "Select a component";
    desc.textContent = "";
    if (btn) btn.disabled = true;
    return;
  }
  title.textContent = sel.c.label || sel.c.id;
  function badge(text, cls) {
    var s = document.createElement("span");
    s.className = "lib-badge" + (cls ? " " + cls : "");
    s.textContent = text;
    badges.appendChild(s);
  }
  badge(sel.c.kind);
  if (sel.c.class) badge("." + sel.c.class);
  badge(sel.group.label);
  if (libraryIsDefault(sel.c)) badge("Default", "lib-default-badge");
  desc.textContent = sel.c.description || "";
  if (btn) {
    btn.disabled = false;
    btn.textContent = "Insert " + (sel.c.label || sel.c.id);
  }
  if (LIB_NAV_PREVIEW_ID !== sel.c.id) loadLibraryPreview(sel.c);
}
function loadLibraryPreview(c) {
  var frame = document.getElementById("lib-preview");
  if (!frame) return;
  LIB_NAV_PREVIEW_ID = c.id;
  frame.srcdoc = "<!doctype html><html><body style='margin:0;display:flex;align-items:center;justify-content:center;height:100vh;background:#0f1117;color:#8fa0b8;font:13px system-ui,sans-serif'>Loading preview…</body></html>";
  var fd = new URLSearchParams();
  fd.set("id", c.id);
  fetch("/__gui__/component-preview", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (LIB_NAV_PREVIEW_ID !== c.id) return;
    if (j && j.ok && j.html) {
      frame.srcdoc = j.html;
    } else {
      var msg = (j && (j.error || (j.diagnostics && j.diagnostics.length))) ? (j.error || j.diagnostics[0]) : "Preview unavailable";
      frame.srcdoc = "<!doctype html><html><body style='margin:0;padding:18px;background:#0f1117;color:#f87171;font:12px ui-monospace,monospace'>" + libNavEscape(msg) + "</body></html>";
    }
  }).catch(function(){
    frame.srcdoc = "<!doctype html><html><body style='margin:0;padding:18px;background:#0f1117;color:#f87171;font:12px ui-monospace,monospace'>Preview request failed.</body></html>";
  });
}
function insertLibraryFromNavigator() {
  var sel = libraryNavSelected();
  if (!sel) return;
  addLibraryComponent(sel.c);
  var st = document.getElementById("lib-status");
  if (st) {
    st.textContent = "Inserted " + (sel.c.label || sel.c.id) + " into the project.";
    st.style.color = "#10b981";
  }
}
document.addEventListener("keydown", function(ev){
  if (ev.key !== "Escape") return;
  var nav = document.getElementById("lib-nav");
  if (nav && nav.classList.contains("on")) { ev.preventDefault(); closeLibraryNavigator(); }
});
document.addEventListener("click", function(ev){
  var nav = document.getElementById("lib-nav");
  if (nav && ev.target === nav) closeLibraryNavigator();
});
`
