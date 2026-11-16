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
function openLibraryNavigator(mode, kind) {
  var nav = document.getElementById("lib-nav");
  if (!nav) return;
  nav.classList.add("on");
  if (mode === "appearance" || mode === "components") LIB_NAV_MODE = mode;
  if (kind && appearanceKindSupported(kind)) {
    LIB_NAV_MODE = "appearance";
    LIB_APP_KIND = kind;
    LIB_APP_SEL = appearanceDefaultSelection(appearanceDefForKind(kind), kind);
    LIB_APP_THEME_ID = "";
  }
  if (LIB_NAV_MODE === "appearance" && (!LIB_APP_KIND || !appearanceKindSupported(LIB_APP_KIND))) {
    LIB_APP_KIND = APPEARANCE_KINDS[0] || "button";
    LIB_APP_SEL = appearanceDefaultSelection(appearanceDefForKind(LIB_APP_KIND), LIB_APP_KIND);
  }
  if (!LIB_NAV_SEL) {
    var all = libraryAllComponents();
    if (all.length) LIB_NAV_SEL = all[0].c.id;
  }
  if (LIB_NAV_MODE === "appearance") {
    LIB_APP_FOCUS_COL = "structure";
    LIB_APP_RESTORE_FOCUS = true;
  }
  renderLibraryNavigator();
  if (LIB_NAV_MODE === "components") {
    var s = document.getElementById("lib-search");
    if (s) s.focus();
  }
}
function closeLibraryNavigator() {
  var nav = document.getElementById("lib-nav");
  if (nav) nav.classList.remove("on");
}
function renderLibraryNavigator() {
  var src = document.getElementById("lib-source");
  if (src) src.textContent = LIB_NAV_MODE === "appearance" ? "Structures / styles / animations" : (COMPONENT_LIBRARY ? (COMPONENT_LIBRARY.source || "component library") : "");
  var st = document.getElementById("lib-status");
  if (st) {
    if (LIB_NAV_MODE === "appearance") {
      st.textContent = LIBRARY_REF_WARNINGS.length ? LIBRARY_REF_WARNINGS.join(" ") : "";
      st.style.color = LIBRARY_REF_WARNINGS.length ? "#fbbf24" : "";
    } else if (COMPONENT_LIBRARY_ERROR) {
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
  renderLibraryMode();
  if (LIB_NAV_MODE === "appearance") renderAppearanceNavigator();
  else {
    renderLibraryGroups();
    renderLibraryList();
    renderLibraryDetail();
  }
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

// ---- control appearance library -------------------------------------------
// Structures (shape), styles (colours) and animations for the built-in control
// kinds. A theme is a structure + style + animation triple. The data is derived
// from COMPONENT_DEFS so every control gets at least a default structure/style
// and its animation options; the curated entries below make the button example
// (and a few other controls) richer.
var APPEARANCE_FAV_KEY = "sui.appearance.themes.v1";
// Control kinds that expose the appearance section. Complex data/visualization
// components (table, graph, chart) are deliberately excluded.
var APPEARANCE_KINDS = ["button", "input", "textarea", "select", "combobox", "datepicker", "number", "slider", "checkbox", "radio", "switch", "color", "date", "time", "file", "progress", "spinner", "badge", "alert", "card", "tabs", "accordion", "menubar", "dropdown", "popover", "modal", "toast", "tree", "commandbar", "drawer", "splitter", "pane", "dropzone", "timeline", "carousel", "stepper", "list"];
function appearanceKindLabel(kind) {
  return PALETTE_LABEL[kind] || kind;
}
function appearanceKindSupported(kind) {
  return APPEARANCE_KINDS.indexOf(kind) >= 0;
}
function appearanceDefForKind(kind) {
  var def = componentDefForKind(kind);
  if (def) return def;
  return { id: "kind:" + kind, label: appearanceKindLabel(kind), kinds: [kind], categories: [], generic: true };
}
var APPEARANCE_CUSTOM = {
  button: {
    structures: [
      { id: "rect", label: "Rectangular", vars: { "--sui-button-radius": "8px" } },
      { id: "pill", label: "Pill", vars: { "--sui-button-radius": "999px", "--sui-button-padding": "8px 20px" } },
      { id: "circle", label: "Circular", vars: { "--sui-button-radius": "999px", "--sui-button-padding": "0px", "width": "56px", "height": "56px" } },
      { id: "square", label: "Square", vars: { "--sui-button-radius": "4px", "--sui-button-padding": "0px", "width": "56px", "height": "56px" } },
      { id: "wide", label: "Wide", vars: { "--sui-button-padding": "10px 28px", "--sui-button-radius": "12px" } }
    ],
    styles: [
      { id: "midnight", label: "Midnight", vars: { "--sui-button-bg": "#111827", "--sui-button-fg": "#e5e7eb", "--sui-button-border": "#334155", "--sui-button-hover": "#1f2937", "--sui-button-shadow": "0 2px 8px rgba(0,0,0,.28)" } },
      { id: "primary", label: "Primary", vars: { "--sui-button-bg": "linear-gradient(135deg, #4c5cf7, #7c3aed)", "--sui-button-fg": "#ffffff", "--sui-button-border": "#4c5cf7", "--sui-button-hover": "#5b6bff", "--sui-button-shadow": "0 4px 14px rgba(76,92,247,.45)" } },
      { id: "coral", label: "Coral", vars: { "--sui-button-bg": "#f43f5e", "--sui-button-fg": "#ffffff", "--sui-button-border": "#e11d48", "--sui-button-hover": "#fb7185", "--sui-button-shadow": "0 4px 14px rgba(244,63,94,.4)" } },
      { id: "mint", label: "Mint", vars: { "--sui-button-bg": "#10b981", "--sui-button-fg": "#052e16", "--sui-button-border": "#059669", "--sui-button-hover": "#34d399", "--sui-button-shadow": "0 4px 12px rgba(16,185,129,.35)" } },
      { id: "neon", label: "Neon", vars: { "--sui-button-bg": "transparent", "--sui-button-fg": "#22d3ee", "--sui-button-border": "#22d3ee", "--sui-button-hover": "rgba(34,211,238,.14)", "--sui-button-shadow": "0 0 14px rgba(34,211,238,.45)" } },
      { id: "ghost", label: "Ghost", vars: { "--sui-button-bg": "transparent", "--sui-button-fg": "#c7d2fe", "--sui-button-border": "transparent", "--sui-button-hover": "rgba(255,255,255,.08)", "--sui-button-shadow": "none" } }
    ]
  },
  check: {
    structures: [
      { id: "compact", label: "Compact", vars: { "--sui-check-size": "14px" } },
      { id: "default", label: "Default", vars: { "--sui-check-size": "18px" } },
      { id: "large", label: "Large", vars: { "--sui-check-size": "24px" } }
    ],
    styles: [
      { id: "midnight", label: "Midnight", vars: { "--sui-check-bg": "#1d2129", "--sui-check-border": "#3d4556", "--sui-check-checked": "#4c5cf7", "--sui-check-checked-border": "#4c5cf7", "--sui-check-mark": "#ffffff" } },
      { id: "success", label: "Success", vars: { "--sui-check-bg": "#0f172a", "--sui-check-border": "#334155", "--sui-check-checked": "#22c55e", "--sui-check-checked-border": "#16a34a", "--sui-check-mark": "#052e16" } },
      { id: "danger", label: "Danger", vars: { "--sui-check-bg": "#1f1215", "--sui-check-border": "#7f1d1d", "--sui-check-checked": "#ef4444", "--sui-check-checked-border": "#dc2626", "--sui-check-mark": "#ffffff" } },
      { id: "mono", label: "Mono", vars: { "--sui-check-bg": "#111827", "--sui-check-border": "#64748b", "--sui-check-checked": "#e5e7eb", "--sui-check-checked-border": "#e5e7eb", "--sui-check-mark": "#111827" } }
    ]
  },
  switch: {
    structures: [
      { id: "default", label: "Default", attrs: {}, vars: {} },
      { id: "rounded", label: "Rounded", attrs: { shape: "rounded" }, vars: {} },
      { id: "square", label: "Square", attrs: { shape: "square" }, vars: {} },
      { id: "pill", label: "Pill", attrs: { shape: "pill" }, vars: {} },
      { id: "ios", label: "iOS", attrs: { shape: "ios" }, vars: {} },
      { id: "material", label: "Material", attrs: { shape: "material" }, vars: {} },
      { id: "minimal", label: "Minimal", attrs: { shape: "minimal" }, vars: {} },
      { id: "soft", label: "Soft", attrs: { shape: "soft" }, vars: {} }
    ],
    styles: [
      { id: "midnight", label: "Midnight", vars: { "--sui-switch-bg": "#374151", "--sui-switch-checked": "#4c5cf7", "--sui-switch-thumb": "#ffffff" } },
      { id: "ios", label: "iOS green", vars: { "--sui-switch-bg": "#334155", "--sui-switch-checked": "#22c55e", "--sui-switch-thumb": "#ffffff" } },
      { id: "coral", label: "Coral", vars: { "--sui-switch-bg": "#3f1d24", "--sui-switch-checked": "#f43f5e", "--sui-switch-thumb": "#ffe4e6" } },
      { id: "mono", label: "Mono", vars: { "--sui-switch-bg": "#1f2937", "--sui-switch-checked": "#e5e7eb", "--sui-switch-thumb": "#111827" } }
    ]
  },
  card: {
    structures: [
      { id: "elevated", label: "Elevated", vars: { "--sui-card-radius": "12px", "--sui-card-padding": "16px" } },
      { id: "flat", label: "Flat", vars: { "--sui-card-shadow": "none", "--sui-card-radius": "8px" } },
      { id: "outline", label: "Outline", vars: { "--sui-card-bg": "transparent", "--sui-card-border-width": "2px" } },
      { id: "glass", label: "Glass", vars: { "--sui-card-bg": "rgba(15,23,42,.55)", "--sui-card-backdrop": "blur(10px)" } },
      { id: "square", label: "Square", vars: { "--sui-card-radius": "0px" } }
    ],
    styles: [
      { id: "dark", label: "Dark", vars: { "--sui-card-bg": "#1a1d26", "--sui-card-fg": "#e6e6ef", "--sui-card-border": "#3d4556", "--sui-card-shadow": "0 4px 12px rgba(0,0,0,.3)" } },
      { id: "light", label: "Light", vars: { "--sui-card-bg": "#f8fafc", "--sui-card-fg": "#0f172a", "--sui-card-border": "#cbd5e1", "--sui-card-shadow": "0 4px 12px rgba(15,23,42,.12)" } },
      { id: "glass", label: "Glass", vars: { "--sui-card-bg": "rgba(15,23,42,.55)", "--sui-card-fg": "#e2e8f0", "--sui-card-border": "rgba(148,163,184,.35)", "--sui-card-backdrop": "blur(10px)" } },
      { id: "neon", label: "Neon", vars: { "--sui-card-bg": "#0b1220", "--sui-card-fg": "#e0f2fe", "--sui-card-border": "#22d3ee", "--sui-card-shadow": "0 0 18px rgba(34,211,238,.28)" } }
    ]
  },
  form: {
    structures: [
      { id: "outline", label: "Outline", vars: { "--sui-field-radius": "8px", "--sui-field-border-width": "1px" } },
      { id: "filled", label: "Filled", vars: { "--sui-field-bg": "#1f2937", "--sui-field-border-width": "0px" } },
      { id: "pill", label: "Pill", vars: { "--sui-field-radius": "999px", "--sui-field-padding": "8px 16px" } },
      { id: "square", label: "Square", vars: { "--sui-field-radius": "0px" } }
    ],
    styles: [
      { id: "dark", label: "Dark", vars: { "--sui-field-bg": "#1d2129", "--sui-field-fg": "#e6e6ef", "--sui-field-border": "#3d4556", "--sui-field-focus": "#4c5cf7" } },
      { id: "filled", label: "Filled", vars: { "--sui-field-bg": "#1f2937", "--sui-field-fg": "#e5e7eb", "--sui-field-border": "#1f2937", "--sui-field-focus": "#60a5fa" } },
      { id: "outline", label: "Outline", vars: { "--sui-field-bg": "transparent", "--sui-field-fg": "#e6e6ef", "--sui-field-border": "#4c5cf7", "--sui-field-focus": "#818cf8" } },
      { id: "mono", label: "Mono", vars: { "--sui-field-bg": "#0f172a", "--sui-field-fg": "#e2e8f0", "--sui-field-border": "#64748b", "--sui-field-focus": "#e2e8f0" } }
    ]
  },
  badge: {
    structures: [
      { id: "pill", label: "Pill", vars: { "--sui-badge-radius": "999px", "--sui-badge-padding": "2px 10px" } },
      { id: "tag", label: "Tag", vars: { "--sui-badge-radius": "4px", "--sui-badge-padding": "2px 6px" } },
      { id: "square", label: "Square", vars: { "--sui-badge-radius": "0px" } },
      { id: "large", label: "Large", vars: { "--sui-badge-font-size": "14px", "--sui-badge-padding": "4px 12px" } }
    ],
    styles: [
      { id: "indigo", label: "Indigo", vars: { "--sui-badge-bg": "#252b45", "--sui-badge-fg": "#a5b4fc", "--sui-badge-border": "#4c5cf7" } },
      { id: "green", label: "Green", vars: { "--sui-badge-bg": "#052e16", "--sui-badge-fg": "#86efac", "--sui-badge-border": "#166534" } },
      { id: "red", label: "Red", vars: { "--sui-badge-bg": "#3f1d24", "--sui-badge-fg": "#fda4af", "--sui-badge-border": "#9f1239" } },
      { id: "amber", label: "Amber", vars: { "--sui-badge-bg": "#3b2a09", "--sui-badge-fg": "#fcd34d", "--sui-badge-border": "#b45309" } },
      { id: "mono", label: "Mono", vars: { "--sui-badge-bg": "#111827", "--sui-badge-fg": "#e5e7eb", "--sui-badge-border": "#64748b" } }
    ]
  },
  progress: {
    structures: [
      { id: "thin", label: "Thin line", vars: { "--sui-progress-height": "4px", "--sui-progress-radius": "999px" } },
      { id: "classic", label: "Classic", vars: { "--sui-progress-height": "8px", "--sui-progress-radius": "999px" } },
      { id: "thick", label: "Thick", vars: { "--sui-progress-height": "14px", "--sui-progress-radius": "999px" } },
      { id: "square", label: "Square", vars: { "--sui-progress-height": "10px", "--sui-progress-radius": "3px" } }
    ],
    styles: [
      { id: "indigo", label: "Indigo", vars: { "--sui-progress-track": "#1f2937", "--sui-progress-value": "#4c5cf7", "--sui-progress-value-end": "#6366f1" } },
      { id: "aurora", label: "Aurora", vars: { "--sui-progress-track": "#0b1220", "--sui-progress-value": "#8b5cf6", "--sui-progress-value-end": "#22d3ee" } },
      { id: "green", label: "Green", vars: { "--sui-progress-track": "#052e16", "--sui-progress-value": "#22c55e", "--sui-progress-value-end": "#16a34a" } },
      { id: "amber", label: "Amber", vars: { "--sui-progress-track": "#3b2a09", "--sui-progress-value": "#f59e0b", "--sui-progress-value-end": "#ef4444" } },
      { id: "mono", label: "Mono", vars: { "--sui-progress-track": "#1f2937", "--sui-progress-value": "#e5e7eb", "--sui-progress-value-end": "#94a3b8" } }
    ]
  },
  spinner: {
    structures: [
      { id: "small", label: "Small", vars: { "--sui-spinner-size": "16px", "--sui-spinner-thickness": "2px" } },
      { id: "default", label: "Default", vars: { "--sui-spinner-size": "24px", "--sui-spinner-thickness": "3px" } },
      { id: "large", label: "Large", vars: { "--sui-spinner-size": "40px", "--sui-spinner-thickness": "4px" } },
      { id: "thick", label: "Thick", vars: { "--sui-spinner-size": "28px", "--sui-spinner-thickness": "6px" } }
    ],
    styles: [
      { id: "accent", label: "Accent", vars: { "--sui-spinner-track": "#4b5563", "--sui-spinner-color": "#4c5cf7", "--sui-spinner-bg": "transparent" } },
      { id: "cyan", label: "Cyan", vars: { "--sui-spinner-track": "#164e63", "--sui-spinner-color": "#22d3ee", "--sui-spinner-color2": "#818cf8" } },
      { id: "green", label: "Green", vars: { "--sui-spinner-track": "#14532d", "--sui-spinner-color": "#22c55e", "--sui-spinner-bg": "transparent" } },
      { id: "mono", label: "Mono", vars: { "--sui-spinner-track": "#334155", "--sui-spinner-color": "#e5e7eb", "--sui-spinner-bg": "transparent" } }
    ]
  },
  slider: {
    structures: [
      { id: "thin", label: "Thin", vars: { "--sui-slider-track": "3px", "--sui-slider-thumb": "14px" } },
      { id: "classic", label: "Classic", vars: { "--sui-slider-track": "6px", "--sui-slider-thumb": "18px" } },
      { id: "thick", label: "Thick", vars: { "--sui-slider-track": "12px", "--sui-slider-thumb": "26px" } },
      { id: "square", label: "Square", vars: { "--sui-slider-radius": "3px", "--sui-slider-thumb-radius": "4px" } }
    ],
    styles: [
      { id: "indigo", label: "Indigo", vars: { "--sui-slider-bg": "#374151", "--sui-slider-thumb-bg": "#4c5cf7" } },
      { id: "cyan", label: "Cyan", vars: { "--sui-slider-bg": "#0f172a", "--sui-slider-thumb-bg": "#22d3ee" } },
      { id: "green", label: "Green", vars: { "--sui-slider-bg": "#14532d", "--sui-slider-thumb-bg": "#22c55e" } },
      { id: "coral", label: "Coral", vars: { "--sui-slider-bg": "#3f1d24", "--sui-slider-thumb-bg": "#f43f5e" } },
      { id: "mono", label: "Mono", vars: { "--sui-slider-bg": "#334155", "--sui-slider-thumb-bg": "#e5e7eb" } }
    ]
  },
  alert: {
    structures: [
      { id: "bar", label: "Accent bar", vars: { "--sui-alert-radius": "8px", "--sui-alert-border-width": "1px", "--sui-alert-padding": "12px 16px" } },
      { id: "filled", label: "Filled", vars: { "--sui-alert-border-width": "0px" } },
      { id: "outline", label: "Outline", vars: { "--sui-alert-bg": "transparent", "--sui-alert-border-width": "2px" } },
      { id: "square", label: "Square", vars: { "--sui-alert-radius": "0px" } }
    ],
    styles: [
      { id: "danger", label: "Danger", vars: { "--sui-alert-bg": "#3a2427", "--sui-alert-fg": "#fca5a5", "--sui-alert-border": "#7f3a3d" } },
      { id: "success", label: "Success", vars: { "--sui-alert-bg": "#0f2a1d", "--sui-alert-fg": "#86efac", "--sui-alert-border": "#166534" } },
      { id: "info", label: "Info", vars: { "--sui-alert-bg": "#10243a", "--sui-alert-fg": "#93c5fd", "--sui-alert-border": "#1d4ed8" } },
      { id: "warning", label: "Warning", vars: { "--sui-alert-bg": "#3b2a09", "--sui-alert-fg": "#fcd34d", "--sui-alert-border": "#b45309" } },
      { id: "mono", label: "Mono", vars: { "--sui-alert-bg": "#111827", "--sui-alert-fg": "#e5e7eb", "--sui-alert-border": "#64748b" } }
    ]
  },
  tabs: {
    structures: [
      { id: "underline", label: "Underline", vars: { "--sui-tabs-radius": "8px", "--sui-tabs-border-width": "1px" } },
      { id: "pills", label: "Pills", vars: { "--sui-tabs-radius": "999px", "--sui-tab-radius": "999px" } },
      { id: "boxed", label: "Boxed", vars: { "--sui-tabs-radius": "6px", "--sui-tabs-border-width": "1px" } },
      { id: "square", label: "Square", vars: { "--sui-tabs-radius": "0px", "--sui-tab-radius": "0px" } }
    ],
    styles: [
      { id: "dark", label: "Dark", vars: { "--sui-tabs-bg": "#1d2129", "--sui-tabs-fg": "#e6e6ef", "--sui-tablist-bg": "#1d2129", "--sui-tab-fg": "#9aa7b8", "--sui-tab-active-bg": "#14171c", "--sui-tab-active-fg": "#e6e6ef", "--sui-tab-accent": "#4c5cf7", "--sui-tabs-border": "#3d4556" } },
      { id: "light", label: "Light", vars: { "--sui-tabs-bg": "#f8fafc", "--sui-tabs-fg": "#0f172a", "--sui-tablist-bg": "#e2e8f0", "--sui-tab-fg": "#475569", "--sui-tab-active-bg": "#ffffff", "--sui-tab-active-fg": "#0f172a", "--sui-tab-accent": "#2563eb", "--sui-tabs-border": "#cbd5e1" } },
      { id: "neon", label: "Neon", vars: { "--sui-tabs-bg": "#0b1220", "--sui-tabs-fg": "#e0f2fe", "--sui-tablist-bg": "#0b1220", "--sui-tab-fg": "#7dd3fc", "--sui-tab-active-bg": "#082f49", "--sui-tab-active-fg": "#e0f2fe", "--sui-tab-accent": "#22d3ee", "--sui-tabs-border": "#164e63" } },
      { id: "mono", label: "Mono", vars: { "--sui-tabs-bg": "#111827", "--sui-tabs-fg": "#e5e7eb", "--sui-tablist-bg": "#111827", "--sui-tab-fg": "#94a3b8", "--sui-tab-active-bg": "#1f2937", "--sui-tab-active-fg": "#f8fafc", "--sui-tab-accent": "#e5e7eb", "--sui-tabs-border": "#374151" } }
    ]
  }
};
function appearanceFieldKey(def, re) {
  for (var i = 0; i < def.categories.length; i++) {
    var fs = def.categories[i].fields || [];
    for (var j = 0; j < fs.length; j++) if (re.test(fs[j].key || "")) return fs[j].key;
  }
  return "";
}
function appearanceAnimKey(def) {
  for (var i = 0; i < def.categories.length; i++) {
    var fs = def.categories[i].fields || [];
    for (var j = 0; j < fs.length; j++) {
      var f = fs[j];
      if ((f.key || "").indexOf("animation") >= 0 || (f.type === "select" && /^sui-|^none$/.test(String((f.opts || [["", ""]])[0][1] || "")))) return f.key || "";
    }
  }
  return "";
}
function appearanceStructures(def, kind) {
  var custom = APPEARANCE_CUSTOM[def.id];
  if (custom && custom.structures) return custom.structures;
  if (def.generic) {
    return [
      { id: "default", label: "Default", vars: {} },
      { id: "rounded", label: "Rounded", vars: { "border-radius": "12px" } },
      { id: "square", label: "Square", vars: { "border-radius": "0px" } },
      { id: "compact", label: "Compact", vars: { padding: "6px 10px" } },
      { id: "wide", label: "Wide", vars: { padding: "12px 20px" } }
    ];
  }
  var out = [{ id: "default", label: "Default", vars: {} }];
  var radius = appearanceFieldKey(def, /-radius$/);
  if (radius) {
    var round = {}; round[radius] = "999px";
    var square = {}; square[radius] = "0px";
    out.push({ id: "round", label: "Rounded", vars: round });
    out.push({ id: "square", label: "Square", vars: square });
  }
  return out;
}
function appearanceColorKeys(def) {
  var out = [];
  (def.categories || []).forEach(function(cat){
    (cat.fields || []).forEach(function(f){
      if (f.type === "color" && f.key && out.indexOf(f.key) < 0) out.push(f.key);
    });
  });
  return out;
}
function appearancePalette(def, mode) {
  var dark = mode === "dark";
  var vars = {};
  var keys = appearanceColorKeys(def);
  if (!keys.length) {
    vars["background-color"] = dark ? "#111827" : "#f8fafc";
    vars["color"] = dark ? "#e5e7eb" : "#0f172a";
    vars["border-color"] = dark ? "#334155" : "#cbd5e1";
    if (!dark) vars["box-shadow"] = "0 4px 12px rgba(15,23,42,.12)";
    return vars;
  }
  keys.forEach(function(k){
    var low = String(k).toLowerCase();
    var v;
    if (/hover/.test(low)) v = dark ? "#1f2937" : "#e2e8f0";
    else if (/checked|accent|focus|value|color/.test(low)) v = dark ? "#4c5cf7" : "#2563eb";
    else if (/fg|mark|thumb|text/.test(low)) v = dark ? "#e5e7eb" : "#0f172a";
    else if (/bg|track|tablist/.test(low)) v = dark ? "#111827" : "#f8fafc";
    else if (/border/.test(low)) v = dark ? "#334155" : "#cbd5e1";
    else v = dark ? "#111827" : "#f8fafc";
    vars[k] = v;
  });
  return vars;
}
function appearanceStyles(def, kind) {
  var custom = APPEARANCE_CUSTOM[def.id];
  var list = custom && custom.styles ? custom.styles.slice() : [];
  if (!list.length) {
    if (def.generic) {
      list = [
        { id: "accent", label: "Accent", vars: { "background-color": "#4c5cf7", "color": "#ffffff", "border-color": "#3730a3", "box-shadow": "0 4px 14px rgba(76,92,247,.35)" } },
        { id: "default", label: "Default", vars: {} },
        { id: "outline", label: "Outline", vars: { "background-color": "transparent", "color": "#a5b4fc", "border-color": "#4c5cf7", "border-width": "1px", "border-style": "solid" } }
      ];
    } else {
      list = [{ id: "default", label: "Default", vars: {} }];
    }
  }
  var dark = null, light = null;
  list = list.filter(function(s){
    if (s.id === "dark" || /^dark$/i.test(s.label)) { if (!dark) dark = s; return false; }
    if (s.id === "light" || /^light$/i.test(s.label)) { if (!light) light = s; return false; }
    return true;
  });
  if (!dark) dark = { id: "dark", label: "Dark", vars: appearancePalette(def, "dark") };
  if (!light) light = { id: "light", label: "Light", vars: appearancePalette(def, "light") };
  return [dark, light].concat(list);
}
function appearanceAnimations(def) {
  var field = null;
  for (var i = 0; i < def.categories.length && !field; i++) {
    var fs = def.categories[i].fields || [];
    for (var j = 0; j < fs.length; j++) {
      if ((fs[j].key || "").indexOf("animation") >= 0) { field = fs[j]; break; }
    }
  }
  if (!field) return genericAppearanceAnimations();
  var out = [];
  var seen = {};
  (field.opts || []).forEach(function(o){
    var value = String(o[1] === undefined ? "" : o[1]);
    if (value === "") return;
    var id = String(o[0] || value).toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "option";
    var base = id, n = 2;
    while (seen[id]) id = base + "-" + (n++);
    seen[id] = true;
    out.push({ id: id, label: String(o[0] || value), value: value });
  });
  if (!out.some(function(o){ return o.value === "none"; })) out.unshift({ id: "none", label: "None", value: "none" });
  return out;
}
function genericAppearanceAnimations() {
  return [
    { id: "none", label: "None", value: "none" },
    { id: "pulse", label: "Pulse", value: "sui-appearance-pulse 1.8s ease-in-out infinite" },
    { id: "glow", label: "Glow", value: "sui-appearance-glow 2.2s ease-in-out infinite" },
    { id: "float", label: "Float", value: "sui-appearance-float 2.6s ease-in-out infinite" },
    { id: "fade", label: "Fade", value: "sui-appearance-fade 2.4s ease-in-out infinite" }
  ];
}
function appearanceKeyframesCss() {
  return "@keyframes sui-appearance-pulse { 0%, 100% { transform: scale(1); opacity: 1; } 50% { transform: scale(1.06); opacity: .85; } }\n" +
    "@keyframes sui-appearance-glow { 0%, 100% { filter: brightness(1); } 50% { filter: brightness(1.35); } }\n" +
    "@keyframes sui-appearance-float { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-4px); } }\n" +
    "@keyframes sui-appearance-fade { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }";
}
function ensureAppearanceKeyframes() {
  var re = /\/\* sui-appearance-keyframes \*\/[\s\S]*?\/\* \/sui-appearance-keyframes \*\//;
  var block = "/* sui-appearance-keyframes */\n" + appearanceKeyframesCss() + "\n/* /sui-appearance-keyframes */";
  var css = String(root.css || "");
  if (re.test(css)) css = css.replace(re, block);
  else css = (css.trim() ? css.trim() + "\n" : "") + block;
  root.css = css.trim() + "\n";
}
function appearanceThemeLookup(def, kind, id) {
  var themes = appearanceThemes(def, kind);
  for (var i = 0; i < themes.length; i++) if (themes[i].id === id) return themes[i];
  return null;
}
function appearanceThemes(def, kind) {
  var S = appearanceStructures(def, kind);
  var T = appearanceStyles(def, kind);
  var A = appearanceAnimations(def);
  function pick(list, ids, fallback) {
    for (var i = 0; i < ids.length; i++) for (var j = 0; j < list.length; j++) if (list[j].id === ids[i]) return list[j].id;
    return fallback;
  }
  var s0 = S[0].id, t0 = T[0].id, a0 = A[0].id;
  var out = [];
  var seen = {};
  function add(id, label, structure, style, animation) {
    var key = structure + "|" + style + "|" + animation;
    if (seen[key]) return;
    seen[key] = true;
    out.push({ id: id, label: label, structure: structure, style: style, animation: animation, generated: true });
  }
  add("classic", "Classic", s0, t0, a0);
  var round = pick(S, ["pill", "round", "circle"], s0);
  var primary = pick(T, ["primary", "midnight"], t0);
  add("rounded", "Rounded", round, primary, a0);
  var neonStyle = pick(T, ["neon", "indigo", "primary"], t0);
  var glow = pick(A, ["glow", "pulse", "float"], a0);
  add("neon", "Neon", round, neonStyle, glow);
  var dangerStyle = pick(T, ["coral", "danger", "red"], t0);
  var pulse = pick(A, ["pulse", "bounce", "jelly"], a0);
  add("alert", "Alert", s0, dangerStyle, pulse);
  var compact = pick(S, ["compact", "square", "tag"], s0);
  add("compact", "Compact", compact, t0, a0);
  var favs = appearanceFavorites().filter(function(f){ return f.defId === def.id; });
  favs.forEach(function(f){
    out.push({ id: f.id, label: f.label, structure: f.structure || "", style: f.style || "", animation: f.animation || "", vars: f.vars || {}, favorite: true });
  });
  return out;
}
function appearanceFavorites() {
  try {
    var raw = localStorage.getItem(APPEARANCE_FAV_KEY);
    var list = raw ? JSON.parse(raw) : [];
    return Array.isArray(list) ? list.filter(function(f){ return f && f.id && f.label && f.defId; }) : [];
  } catch (e) { return []; }
}
function appearanceSaveFavorite(def, kind, node, sel) {
  function store(name) {
    name = String(name || "").trim();
    if (!name) return;
    var vars = {};
    if (node) {
      var map = styleDeclMap(node.style || "");
      appearanceKeyUnion(def).forEach(function(k){ if (map[k] !== undefined) vars[k] = map[k]; });
    } else {
      vars = appearanceVarsForSelection(def, kind, sel);
    }
    var list = appearanceFavorites();
    var item = {
      id: "fav-" + Date.now(),
      label: name,
      kind: kind,
      defId: def.id,
      structure: sel && sel.structure ? sel.structure : "",
      style: sel && sel.style ? sel.style : "",
      animation: sel && sel.animation ? sel.animation : "",
      vars: vars
    };
    list.push(item);
    try { localStorage.setItem(APPEARANCE_FAV_KEY, JSON.stringify(list)); } catch (e) {}
    setStatus("Saved appearance theme '" + name + "'.");
    renderInspector();
    if (document.getElementById("lib-nav") && document.getElementById("lib-nav").classList.contains("on")) renderLibraryNavigator();
    return item;
  }
  var initial = (def.label || kind) + " theme";
  if (typeof askPrompt === "function") askPrompt("Save appearance theme", "Name", initial, store, "Stored in this editor and available in the control inspector.");
  else store(initial);
}
function appearanceKeyUnion(def) {
  var keys = {}, out = [];
  function add(k) { if (k && !keys[k]) { keys[k] = true; out.push(k); } }
  appearanceStructures(def, "").forEach(function(s){ Object.keys(s.vars || {}).forEach(add); });
  appearanceStyles(def, "").forEach(function(s){ Object.keys(s.vars || {}).forEach(add); });
  var ak = appearanceAnimKey(def);
  if (ak) add(ak); else add("animation");
  return out;
}
function appearanceAttrsUnion(def) {
  var keys = {}, out = [];
  appearanceStructures(def, "").forEach(function(s){
    Object.keys(s.attrs || {}).forEach(function(k){ if (!keys[k]) { keys[k] = true; out.push(k); } });
  });
  return out;
}
function appearanceVariant(list, id) {
  for (var i = 0; i < list.length; i++) if (list[i].id === id) return list[i];
  return null;
}
function appearanceVarsForSelection(def, kind, sel) {
  var vars = {};
  sel = sel || {};
  if (sel.themeVars) {
    Object.keys(sel.themeVars).forEach(function(k){ vars[k] = sel.themeVars[k]; });
    return vars;
  }
  var st = appearanceVariant(appearanceStructures(def, kind), sel.structure);
  var sy = appearanceVariant(appearanceStyles(def, kind), sel.style);
  var an = appearanceVariant(appearanceAnimations(def), sel.animation);
  if (st) Object.keys(st.vars || {}).forEach(function(k){ vars[k] = st.vars[k]; });
  if (sy) Object.keys(sy.vars || {}).forEach(function(k){ vars[k] = sy.vars[k]; });
  var ak = appearanceAnimKey(def) || "animation";
  if (an && an.value && an.value !== "none") vars[ak] = an.value;
  return vars;
}
function appearanceMatchVariant(list, map, attrs, node) {
  for (var i = 0; i < list.length; i++) {
    var v = list[i], ok = true;
    Object.keys(v.vars || {}).forEach(function(k){ if (String(map[k] === undefined ? "" : map[k]) !== String(v.vars[k])) ok = false; });
    if (ok && v.attrs) {
      Object.keys(v.attrs).forEach(function(k){ if (String((node && node[k]) || "") !== String(v.attrs[k])) ok = false; });
    }
    if (ok) return v.id;
  }
  return "";
}
function appearanceCurrent(node, def, kind) {
  var map = styleDeclMap(node.style || "");
  var structs = appearanceStructures(def, kind);
  var styles = appearanceStyles(def, kind);
  var anims = appearanceAnimations(def);
  var ak = appearanceAnimKey(def) || "animation";
  var structure = appearanceMatchVariant(structs, map, null, node);
  var style = appearanceMatchVariant(styles, map, null, node);
  var animation = "";
  var av = String(map[ak] === undefined ? "" : map[ak]);
  for (var i = 0; i < anims.length; i++) {
    if (String(anims[i].value) === av || (av === "" && anims[i].value === "none")) { animation = anims[i].id; break; }
  }
  var theme = "";
  var themes = appearanceThemes(def, kind);
  for (var j = 0; j < themes.length; j++) {
    if (themes[j].structure === structure && themes[j].style === style && themes[j].animation === animation) { theme = themes[j].id; break; }
  }
  return { structure: structure, style: style, animation: animation, theme: theme, vars: map };
}
function appearanceApply(node, def, patch) {
  if (!node) return;
  patch = patch || {};
  var kind = node.kind;
  var map = styleDeclMap(node.style || "");
  var allKeys = appearanceKeyUnion(def);
  if (patch.themeVars) {
    allKeys.forEach(function(k){ delete map[k]; });
    Object.keys(patch.themeVars).forEach(function(k){ map[k] = patch.themeVars[k]; });
  } else {
    if (patch.structure) {
      var st = appearanceVariant(appearanceStructures(def, kind), patch.structure);
      appearanceStructures(def, kind).forEach(function(s){ Object.keys(s.vars || {}).forEach(function(k){ delete map[k]; }); });
      if (st) {
        Object.keys(st.vars || {}).forEach(function(k){ map[k] = st.vars[k]; });
        appearanceAttrsUnion(def).forEach(function(k){ delete node[k]; });
        Object.keys(st.attrs || {}).forEach(function(k){ node[k] = st.attrs[k]; });
      }
    }
    if (patch.style) {
      var sy = appearanceVariant(appearanceStyles(def, kind), patch.style);
      appearanceStyles(def, kind).forEach(function(s){ Object.keys(s.vars || {}).forEach(function(k){ delete map[k]; }); });
      if (sy) Object.keys(sy.vars || {}).forEach(function(k){ map[k] = sy.vars[k]; });
    }
    if (patch.animation) {
      var an = appearanceVariant(appearanceAnimations(def), patch.animation);
      var ak = appearanceAnimKey(def);
      if (ak) {
        if (an && an.value && an.value !== "none") map[ak] = an.value; else delete map[ak];
      } else if (an && an.value && an.value !== "none") {
        map["animation"] = an.value;
        ensureAppearanceKeyframes();
      } else {
        delete map["animation"];
      }
    }
  }
  if (patch.structure) {
    var st2 = appearanceVariant(appearanceStructures(def, kind), patch.structure);
    if (st2) {
      appearanceAttrsUnion(def).forEach(function(k){ delete node[k]; });
      Object.keys(st2.attrs || {}).forEach(function(k){ node[k] = st2.attrs[k]; });
    }
  }
  writeComponentNodeStyle(def, node, map);
  ensureComponentRules(def);
}
function appearanceVarsForTheme(def, kind, theme) {
  if (!theme) return {};
  if (theme.vars) {
    var vars = {};
    Object.keys(theme.vars).forEach(function(k){ vars[k] = theme.vars[k]; });
    return vars;
  }
  return appearanceVarsForSelection(def, kind, { structure: theme.structure, style: theme.style, animation: theme.animation });
}
function appearancePatchForTheme(theme) {
  if (!theme) return {};
  return {
    structure: theme.structure || "",
    style: theme.style || "",
    animation: theme.animation || "",
    themeVars: theme.vars || null
  };
}
function appearancePreviewNode(kind, attrs) {
  var attrStr = "";
  Object.keys(attrs || {}).forEach(function(k){ attrStr += " " + k + ": \"" + esc(String(attrs[k])) + "\""; });
  var cls = " class: \"appearance-preview\"";
  if (kind === "button") return "button \"Button\"" + cls + attrStr;
  if (kind === "card") return "card \"Card\"" + cls + attrStr + " {\n\t\tp \"Preview\"\n\t}";
  if (kind === "checkbox") return "checkbox \"Checkbox\" checked: true" + cls + attrStr;
  if (kind === "radio") return "radio \"Radio\" name: \"appearance-preview\" checked: true" + cls + attrStr;
  if (kind === "switch") return "switch \"Switch\" checked: true" + cls + attrStr;
  if (kind === "input") return "input \"Input\" value: \"Preview\"" + cls + attrStr;
  if (kind === "textarea") return "textarea \"Textarea\" value: \"Preview\"" + cls + attrStr;
  if (kind === "select") return "select \"Select\" options: \"One,Two\"" + cls + attrStr;
  if (kind === "badge") return "badge \"Live\"" + cls + attrStr;
  if (kind === "progress") return "progress value: 62 max: 100" + cls + attrStr;
  if (kind === "spinner") return "spinner" + cls + attrStr;
  if (kind === "tabs") return "tabs \"Tabs\"" + cls + attrStr + " {\n\t\ttab \"One\" { p \"One\" }\n\t\ttab \"Two\" { p \"Two\" }\n\t}";
  if (kind === "accordion") return "accordion \"Sections\"" + cls + attrStr + " {\n\t\taccordionitem \"First\" { p \"Content\" }\n\t}";
  if (kind === "menubar") return "menubar" + cls + attrStr + " {\n\t\tmenu \"File\" { menuitem \"Open\" }\n\t}";
  if (kind === "alert") return "alert \"Preview\"" + cls + attrStr;
  if (kind === "slider") return "slider value: 50" + cls + attrStr;
  if (kind === "number") return "number \"Number\" value: \"42\"" + cls + attrStr;
  if (kind === "color") return "color \"Colour\" value: \"#4c5cf7\"" + cls + attrStr;
  if (kind === "date") return "date \"Date\" value: \"2026-01-01\"" + cls + attrStr;
  if (kind === "time") return "time \"Time\" value: \"12:00\"" + cls + attrStr;
  if (kind === "file") return "file \"File\"" + cls + attrStr;
  if (kind === "combobox") return "combobox \"Combo\" options: \"One,Two\"" + cls + attrStr;
  if (kind === "datepicker") return "datepicker \"Date\"" + cls + attrStr;
  if (kind === "dropdown") return "dropdown \"Menu\"" + cls + attrStr + " {\n\t\tdropdownitem \"One\"\n\t}";
  if (kind === "popover") return "popover \"Popover\"" + cls + attrStr + " {\n\t\tp \"Panel\"\n\t}";
  if (kind === "modal") return "modal \"Modal\"" + cls + attrStr + " {\n\t\tp \"Content\"\n\t}";
  if (kind === "toast") return "toast \"Toast\"" + cls + attrStr + " {\n\t\tp \"Saved\"\n\t}";
  if (kind === "tree") return "tree \"Tree\"" + cls + attrStr + " {\n\t\ttreeitem \"One\"\n\t}";
  if (kind === "commandbar") return "commandbar \"Commands\"" + cls + attrStr + " {\n\t\tcommand \"One\"\n\t}";
  if (kind === "drawer") return "drawer \"Drawer\"" + cls + attrStr + " {\n\t\tp \"Panel\"\n\t}";
  if (kind === "splitter") return "splitter" + cls + attrStr + " {\n\t\tpane { p \"A\" }\n\t\tpane { p \"B\" }\n\t}";
  if (kind === "pane") return "splitter {\n\t\tpane" + cls + attrStr + " { p \"A\" }\n\t\tpane { p \"B\" }\n\t}";
  if (kind === "dropzone") return "dropzone \"Drop files\"" + cls + attrStr;
  if (kind === "timeline") return "timeline \"History\"" + cls + attrStr + " {\n\t\ttimelineitem \"Event\"\n\t}";
  if (kind === "carousel") return "carousel" + cls + attrStr + " {\n\t\tslide \"One\"\n\t}";
  if (kind === "stepper") return "stepper" + cls + attrStr + " {\n\t\tstep \"One\"\n\t}";
  if (kind === "list") return "list" + cls + attrStr + " {\n\t\tp \"One\"\n\t}";
  if (kind === "divider") return "divider" + cls + attrStr;
  return kind + cls + attrStr;
}
function appearancePreviewSource(def, kind, sel) {
  var vars = appearanceVarsForSelection(def, kind, sel);
  var st = appearanceVariant(appearanceStructures(def, kind), sel && sel.structure);
  var decls = Object.keys(vars).map(function(k){ return k + ": " + vars[k] + ";"; }).join(" ");
  var css = (def.css || "") + "\n" + appearanceKeyframesCss() + "\n.appearance-preview { " + decls + " }\n" +
    "body { margin: 0; padding: 24px; background: #0f1117; color: #e5e7eb; font-family: system-ui, -apple-system, 'Segoe UI', sans-serif; }";
  return "css \"" + esc(css) + "\"\n\nview main {\n\t" + appearancePreviewNode(kind, st ? st.attrs : null) + "\n}\n";
}
function appearanceLoadPreview(def, kind, sel, frame) {
  if (!frame) return;
  frame.srcdoc = "<!doctype html><html><body style='margin:0;display:flex;align-items:center;justify-content:center;height:100vh;background:#0f1117;color:#8fa0b8;font:13px system-ui,sans-serif'>Loading preview…</body></html>";
  var fd = new URLSearchParams();
  fd.set("sui", appearancePreviewSource(def, kind, sel));
  fetch("/__gui__/source-preview", { method: "POST", body: fd }).then(function(r){ return r.json(); }).then(function(j){
    if (j && j.ok && j.html) frame.srcdoc = j.html;
    else {
      var msg = (j && (j.error || (j.diagnostics && j.diagnostics.length))) ? (j.error || j.diagnostics[0]) : "Preview unavailable";
      frame.srcdoc = "<!doctype html><html><body style='margin:0;padding:18px;background:#0f1117;color:#f87171;font:12px ui-monospace,monospace'>" + libNavEscape(msg) + "</body></html>";
    }
  }).catch(function(){
    frame.srcdoc = "<!doctype html><html><body style='margin:0;padding:18px;background:#0f1117;color:#f87171;font:12px ui-monospace,monospace'>Preview request failed.</body></html>";
  });
}
function appearanceInspectorSection(def, node) {
  var kind = node.kind;
  var cur = appearanceCurrent(node, def, kind);
  var structs = appearanceStructures(def, kind);
  var styles = appearanceStyles(def, kind);
  var anims = appearanceAnimations(def);
  var themes = appearanceThemes(def, kind);
  var rows = [];
  function optionRow(label, list, value, onChange, title) {
    var row = document.createElement("div"); row.className = "rowline";
    var l = document.createElement("label"); l.textContent = label; l.title = title || ""; row.appendChild(l);
    var sel = document.createElement("select");
    var custom = document.createElement("option"); custom.value = ""; custom.textContent = "Custom"; sel.appendChild(custom);
    list.forEach(function(o){
      var op = document.createElement("option"); op.value = o.id; op.textContent = o.label; sel.appendChild(op);
    });
    sel.value = value || "";
    sel.onchange = function(){ onChange(sel.value); };
    row.appendChild(sel);
    rows.push(row);
    return sel;
  }
  function apply(patch, label) {
    withUndo(function(){ appearanceApply(node, def, patch); }, "appearance " + (label || kind));
  }
  // Order requested by the designer: Theme, Structure, Style, Animation.
  var themeSel = document.createElement("select");
  var themeRow = document.createElement("div"); themeRow.className = "rowline";
  var tl = document.createElement("label"); tl.textContent = "Theme"; themeRow.appendChild(tl);
  var tc = document.createElement("option"); tc.value = ""; tc.textContent = "Custom"; themeSel.appendChild(tc);
  themes.forEach(function(t){
    var op = document.createElement("option"); op.value = t.id; op.textContent = t.label + (t.favorite ? " (saved)" : ""); themeSel.appendChild(op);
  });
  themeSel.value = cur.theme || "";
  themeSel.onchange = function(){
    var t = appearanceThemeLookup(def, kind, themeSel.value);
    if (!t) return;
    apply(appearancePatchForTheme(t), "theme");
  };
  themeRow.appendChild(themeSel);
  rows.push(themeRow);
  optionRow("Structure", structs, cur.structure, function(v){
    apply({ structure: v }, "structure");
  }, "Shape and drawing of the control");
  optionRow("Style", styles, cur.style, function(v){
    apply({ style: v }, "style");
  }, "Colour palette of the control");
  optionRow("Animation", anims, cur.animation, function(v){
    apply({ animation: v }, "animation");
  }, "Animation of the control parts");
  var actions = document.createElement("div"); actions.className = "rowline";
  var browse = document.createElement("button"); browse.type = "button"; browse.className = "btn"; browse.textContent = "Browse…";
  browse.title = "Open the appearance library for this control";
  browse.onclick = function(){ openLibraryNavigator("appearance", kind); };
  var save = document.createElement("button"); save.type = "button"; save.className = "btn"; save.textContent = "Save theme…";
  save.title = "Save this structure + style + animation as a theme";
  save.onclick = function(){ appearanceSaveFavorite(def, kind, node, cur); };
  actions.appendChild(browse); actions.appendChild(save);
  rows.push(actions);
  var hint = document.createElement("div"); hint.className = "hint"; hint.style.fontSize = "10px";
  hint.textContent = "Applies to this element only. Use Settings… for the project default. Themes are saved in this editor.";
  rows.push(hint);
  var sec = acc(appearanceKindLabel(kind), rows);
  sec.classList.add("open");
  return sec;
}
var LIB_NAV_MODE = "appearance";
var LIB_APP_DEF_ID = "";
var LIB_APP_KIND = "";
var LIB_APP_SEL = null;
var LIB_APP_THEME_ID = "";
var LIB_APP_FOCUS_COL = "structure";
var LIB_APP_RESTORE_FOCUS = false;
function appearanceDefById(id) {
  for (var i = 0; i < COMPONENT_DEFS.length; i++) if (COMPONENT_DEFS[i].id === id) return COMPONENT_DEFS[i];
  return null;
}
function appearanceDefaultSelection(def, kind) {
  var t = appearanceThemeLookup(def, kind, "classic") || appearanceThemes(def, kind)[0] || null;
  return t ? { structure: t.structure, style: t.style, animation: t.animation, themeVars: t.vars || null } : { structure: "", style: "", animation: "" };
}
function renderLibraryMode() {
  var comp = document.getElementById("lib-body");
  var app = document.getElementById("lib-app-body");
  var cb = document.getElementById("lib-mode-components");
  var ab = document.getElementById("lib-mode-appearance");
  if (comp) comp.style.display = LIB_NAV_MODE === "appearance" ? "none" : "grid";
  if (app) app.style.display = LIB_NAV_MODE === "appearance" ? "grid" : "none";
  if (cb) cb.classList.toggle("on", LIB_NAV_MODE !== "appearance");
  if (ab) ab.classList.toggle("on", LIB_NAV_MODE === "appearance");
}
function setLibraryMode(mode) {
  LIB_NAV_MODE = mode === "components" ? "components" : "appearance";
  if (LIB_NAV_MODE === "appearance") {
    LIB_APP_FOCUS_COL = "structure";
    LIB_APP_RESTORE_FOCUS = true;
  }
  renderLibraryNavigator();
}
function renderAppearanceKinds() {
  var box = document.getElementById("lib-kinds");
  if (!box) return;
  box.textContent = "";
  APPEARANCE_KINDS.forEach(function(kind){
    var b = document.createElement("button");
    b.type = "button";
    b.className = "lib-group" + (LIB_APP_KIND === kind ? " on" : "");
    b.textContent = appearanceKindLabel(kind);
    b.onclick = function(){
      LIB_APP_KIND = kind;
      LIB_APP_SEL = appearanceDefaultSelection(appearanceDefForKind(kind), kind);
      LIB_APP_THEME_ID = "";
      renderAppearanceNavigator();
    };
    box.appendChild(b);
  });
}
function appearanceColumnBox(col) {
  var boxes = { structure: "lib-structures", style: "lib-styles", animation: "lib-animations" };
  return document.getElementById(boxes[col] || "lib-structures");
}
function appearanceFocusColumn(col) {
  var box = appearanceColumnBox(col);
  if (!box) return;
  LIB_APP_FOCUS_COL = col;
  var selected = (LIB_APP_SEL && LIB_APP_SEL[col]) || "";
  var target = selected ? box.querySelector('[data-id="' + selected + '"]') : null;
  if (!target) target = box.querySelector(".lib-app-option");
  if (target) target.focus();
}
function appearanceOptionKey(ev, btn, col) {
  var cols = ["structure", "style", "animation"];
  var buttons = Array.prototype.slice.call(btn.parentNode.querySelectorAll(".lib-app-option"));
  var at = buttons.indexOf(btn);
  if (ev.key === "ArrowDown" || ev.key === "ArrowUp" || ev.key === "Home" || ev.key === "End") {
    ev.preventDefault();
    ev.stopPropagation();
    var next = at;
    if (ev.key === "ArrowDown") next = Math.min(buttons.length - 1, at + 1);
    else if (ev.key === "ArrowUp") next = Math.max(0, at - 1);
    else if (ev.key === "Home") next = 0;
    else next = buttons.length - 1;
    if (buttons[next]) { LIB_APP_FOCUS_COL = col; buttons[next].focus(); }
    return;
  }
  if (ev.key === "ArrowRight" || ev.key === "ArrowLeft") {
    ev.preventDefault();
    ev.stopPropagation();
    var idx = cols.indexOf(col);
    var targetCol = cols[ev.key === "ArrowRight" ? Math.min(cols.length - 1, idx + 1) : Math.max(0, idx - 1)];
    if (!targetCol || targetCol === col) return;
    LIB_APP_FOCUS_COL = targetCol;
    var target = appearanceColumnBox(targetCol);
    var selected = (LIB_APP_SEL && LIB_APP_SEL[targetCol]) || "";
    var nextBtn = selected && target ? target.querySelector('[data-id="' + selected + '"]') : null;
    if (!nextBtn && target) nextBtn = target.querySelector(".lib-app-option");
    if (nextBtn) nextBtn.focus();
    return;
  }
  if (ev.key === "Enter" || ev.key === " " || ev.key === "Spacebar") {
    ev.preventDefault();
    ev.stopPropagation();
    btn.click();
  }
}
function renderAppearanceOptions() {
  var kind = LIB_APP_KIND || APPEARANCE_KINDS[0];
  var def = appearanceDefForKind(kind);
  if (!def) return;
  function list(boxId, items, selKey) {
    var box = document.getElementById(boxId);
    if (!box) return;
    box.textContent = "";
    box.setAttribute("role", "listbox");
    box.setAttribute("aria-label", selKey);
    items.forEach(function(item){
      var b = document.createElement("button");
      b.type = "button";
      b.className = "lib-app-option" + (LIB_APP_SEL && LIB_APP_SEL[selKey] === item.id ? " on" : "");
      b.textContent = item.label;
      b.title = item.label;
      b.setAttribute("data-id", item.id);
      b.setAttribute("data-col", selKey);
      b.setAttribute("role", "option");
      b.setAttribute("aria-selected", LIB_APP_SEL && LIB_APP_SEL[selKey] === item.id ? "true" : "false");
      b.onfocus = function(){ LIB_APP_FOCUS_COL = selKey; };
      b.onkeydown = function(ev){ appearanceOptionKey(ev, b, selKey); };
      b.onclick = function(){
        LIB_APP_FOCUS_COL = selKey;
        LIB_APP_RESTORE_FOCUS = true;
        LIB_APP_SEL[selKey] = item.id;
        LIB_APP_SEL.themeVars = null;
        LIB_APP_THEME_ID = "";
        renderAppearanceOptions();
        renderAppearanceTheme();
        appearanceLoadPreview(def, kind, LIB_APP_SEL, document.getElementById("lib-app-preview"));
      };
      box.appendChild(b);
    });
    var group = box.closest ? box.closest(".lib-app-group") : null;
    if (group && !group.__appColumnClick) {
      group.__appColumnClick = true;
      group.addEventListener("click", function(ev){
        if (ev.target === group || (ev.target && ev.target.tagName === "H4")) appearanceFocusColumn(selKey);
      });
    }
  }
  list("lib-structures", appearanceStructures(def, kind), "structure");
  list("lib-styles", appearanceStyles(def, kind), "style");
  list("lib-animations", appearanceAnimations(def), "animation");
  if (LIB_APP_RESTORE_FOCUS) {
    LIB_APP_RESTORE_FOCUS = false;
    appearanceFocusColumn(LIB_APP_FOCUS_COL || "structure");
  }
}
function renderAppearanceTheme() {
  var kind = LIB_APP_KIND || APPEARANCE_KINDS[0];
  var def = appearanceDefForKind(kind);
  if (!def) return;
  var sel = document.getElementById("lib-app-theme");
  if (!sel) return;
  sel.textContent = "";
  var custom = document.createElement("option"); custom.value = ""; custom.textContent = "Custom"; sel.appendChild(custom);
  var themes = appearanceThemes(def, kind);
  var current = "";
  themes.forEach(function(t){
    var op = document.createElement("option"); op.value = t.id; op.textContent = t.label + (t.favorite ? " (saved)" : ""); sel.appendChild(op);
    if (LIB_APP_SEL && t.structure === LIB_APP_SEL.structure && t.style === LIB_APP_SEL.style && t.animation === LIB_APP_SEL.animation) current = t.id;
  });
  sel.value = LIB_APP_THEME_ID || current;
  sel.onchange = function(){
    var t = appearanceThemeLookup(def, kind, sel.value);
    if (!t) return;
    LIB_APP_SEL = { structure: t.structure, style: t.style, animation: t.animation, themeVars: t.vars || null };
    LIB_APP_THEME_ID = t.id;
    renderAppearanceOptions();
    renderAppearanceTheme();
    appearanceLoadPreview(def, kind, LIB_APP_SEL, document.getElementById("lib-app-preview"));
  };
}
function renderAppearanceNavigator() {
  if (!LIB_APP_KIND || !appearanceKindSupported(LIB_APP_KIND)) LIB_APP_KIND = APPEARANCE_KINDS[0] || "button";
  var def = appearanceDefForKind(LIB_APP_KIND);
  if (!LIB_APP_SEL) LIB_APP_SEL = appearanceDefaultSelection(def, LIB_APP_KIND);
  var title = document.getElementById("lib-app-title");
  if (title) title.textContent = appearanceKindLabel(LIB_APP_KIND) + " appearance";
  renderAppearanceKinds();
  renderAppearanceOptions();
  renderAppearanceTheme();
  var kind = LIB_APP_KIND || def.kinds[0];
  var info = document.getElementById("lib-app-info");
  if (info) info.textContent = "Arrow keys move within a list, Left/Right switch list, Enter selects. The preview is a real compiled control.";
  var status = document.getElementById("lib-app-status");
  if (status) status.textContent = "";
  appearanceLoadPreview(def, kind, LIB_APP_SEL, document.getElementById("lib-app-preview"));
}
function applyAppearanceFromNavigator() {
  var kind = LIB_APP_KIND || APPEARANCE_KINDS[0];
  var def = appearanceDefForKind(kind);
  if (!def) return;
  var node = selPath ? byPath(selPath, root.children) : null;
  if (!node || node.kind !== kind) {
    var st = document.getElementById("lib-app-status");
    if (st) st.textContent = "Select a " + appearanceKindLabel(kind) + " element in the project first.";
    return;
  }
  withUndo(function(){ appearanceApply(node, def, LIB_APP_SEL); }, "appearance " + node.kind);
  var st2 = document.getElementById("lib-app-status");
  if (st2) st2.textContent = "Applied to the selected " + def.label + ".";
}
function saveAppearanceThemeFromNavigator() {
  var kind = LIB_APP_KIND || APPEARANCE_KINDS[0];
  var def = appearanceDefForKind(kind);
  if (!def) return;
  var node = selPath ? byPath(selPath, root.children) : null;
  if (node && node.kind !== kind) node = null;
  appearanceSaveFavorite(def, kind, node, LIB_APP_SEL);
  var st = document.getElementById("lib-app-status");
  if (st) st.textContent = "Theme saved. It is now available in the control inspector.";
}
`
