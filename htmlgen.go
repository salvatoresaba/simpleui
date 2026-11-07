package main

import "sort"
import "strconv"
import "strings"
import "sync/atomic"

// htmlgen.go — HTML/CSS code generation: index.html + style.css (SPEC §4
// "htmlgen" and §5 "Generated-web-app contract"). Owned by Agent D.
//
// This file only reads the Prog/ViewDecl/Element/Attr shapes from ast.go; it
// never validates tags or attributes (the typechecker owns that) and never
// panics. Interpolation parts in view titles and class values are ignored
// (their literal text is kept); the JS generator owns the reactive `${...}`
// output side of interpolation.

// suiCssBase is the fixed base stylesheet: universal reset, dark palette,
// .sui-view card, base element styles, flex helpers and a code/pre block.
// Class names found in views are appended as pass-through rules by
// GenerateCss.
var suiCssBase = `* { box-sizing: border-box; }

:root {
  --sui-bg: #14171c;
  --sui-hover: #2d3340;
  --sui-table-hover-bg: #20252e;
  --sui-border: #3d4556;
  --sui-fg: #e6e6e6;
  --sui-radius: 8px;
  --sui-ac-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
}

body {
  margin: 0;
  font-family: system-ui, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background: var(--sui-bg);
  color: var(--sui-fg);
  padding: 0;
  line-height: 1.5;
}

/* Form controls - modern minimal style */
.sui-input, .sui-textarea, .sui-select {
  appearance: none;
  width: 100%;
  background: #1d2129;
  color: var(--sui-fg);
  border: 1px solid var(--sui-border);
  border-radius: var(--sui-radius);
  padding: 8px 12px;
  font-family: inherit;
  font-size: 14px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.sui-input::placeholder, .sui-textarea::placeholder {
  color: #6b7280;
}
:where(.sui-input):focus, :where(.sui-textarea):focus, :where(.sui-select):focus {
  outline: none;
  border-color: #4c5cf7;
  box-shadow: 0 0 0 3px rgba(76, 92, 247, 0.25);
}
:where(.sui-input):disabled, :where(.sui-textarea):disabled, :where(.sui-select):disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.sui-textarea {
  resize: vertical;
  min-height: 80px;
}
.sui-select {
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 8px center;
  background-repeat: no-repeat;
  background-size: 16px;
  padding-right: 32px;
}
.sui-combobox { display: flex; flex-direction: column; gap: 4px; position: relative; min-width: 180px; }
.sui-combobox-label { font-size: 12px; color: var(--sui-muted, #8fa0b8); }
.sui-combobox-control { position: relative; display: flex; align-items: center; }
.sui-combobox-input {
  appearance: none;
  width: 100%;
  min-width: 0;
  background: #1d2129;
  color: var(--sui-fg, #e6e6ef);
  border: 1px solid var(--sui-border, #3a4152);
  border-radius: 8px;
  padding: 8px 58px 8px 12px;
  font-family: inherit;
  font-size: 14px;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.sui-combobox-input::placeholder { color: #6b7280; }
.sui-combobox-input:focus { outline: none; border-color: #4c5cf7; box-shadow: 0 0 0 3px rgba(76,92,247,.25); }
.sui-combobox-input:disabled { opacity: .5; cursor: not-allowed; }
.sui-combobox-toggle,
.sui-combobox-clear {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  border: 0;
  background: transparent;
  color: var(--sui-muted, #8fa0b8);
  font: inherit;
  cursor: pointer;
  padding: 2px 5px;
  line-height: 1;
}
.sui-combobox-toggle { right: 5px; font-size: 13px; }
.sui-combobox-clear { right: 28px; font-size: 16px; }
.sui-combobox-toggle:hover, .sui-combobox-clear:hover { color: var(--sui-fg, #e6e6ef); }
.sui-combobox-value { display: none; }
.sui-combobox-popup {
  position: absolute;
  z-index: 4000;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 220px;
  overflow: auto;
  padding: 4px;
  background: #1d2129;
  color: var(--sui-fg, #e6e6ef);
  border: 1px solid var(--sui-border, #3a4152);
  border-radius: 8px;
  box-shadow: var(--sui-ac-shadow, 0 8px 24px rgba(0,0,0,.45));
}
.sui-combobox-popup[hidden] { display: none; }
.sui-combobox-option { padding: 7px 9px; border-radius: 5px; font-size: 13px; cursor: pointer; }
.sui-combobox-option:hover, .sui-combobox-option.active { background: var(--sui-hover, #2a3143); }
.sui-combobox-empty { padding: 8px 9px; font-size: 12px; color: var(--sui-muted, #8fa0b8); }
.sui-combobox-error { font-size: 12px; color: #f87171; }
.sui-combobox-error[hidden] { display: none; }
.sui-datepicker { display: flex; flex-direction: column; gap: 4px; position: relative; min-width: 180px; }
.sui-datepicker-label { font-size: 12px; color: var(--sui-muted, #8fa0b8); }
.sui-datepicker-control { position: relative; display: flex; align-items: center; }
.sui-datepicker-input {
  appearance: none;
  width: 100%;
  min-width: 0;
  background: #1d2129;
  color: var(--sui-fg, #e6e6ef);
  border: 1px solid var(--sui-border, #3a4152);
  border-radius: 8px;
  padding: 8px 36px 8px 12px;
  font-family: inherit;
  font-size: 14px;
  color-scheme: dark;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.sui-datepicker-input:focus { outline: none; border-color: #4c5cf7; box-shadow: 0 0 0 3px rgba(76,92,247,.25); }
.sui-datepicker-input:disabled { opacity: .5; cursor: not-allowed; }
.sui-datepicker-clear {
  position: absolute;
  top: 50%;
  right: 5px;
  transform: translateY(-50%);
  border: 0;
  background: transparent;
  color: var(--sui-muted, #8fa0b8);
  font: inherit;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  padding: 2px 5px;
}
.sui-datepicker-clear:hover { color: var(--sui-fg, #e6e6ef); }
.sui-datepicker-value { display: none; }
.sui-datepicker-error { font-size: 12px; color: #f87171; }
.sui-datepicker-error[hidden] { display: none; }
.sui-tree {
  display: flex;
  flex-direction: column;
  gap: var(--sui-tree-gap, 2px);
  color: var(--sui-fg, #e6e6ef);
  background: var(--sui-tree-panel-bg, transparent);
  border: var(--sui-tree-panel-border, 0);
  border-radius: var(--sui-tree-panel-radius, 0);
  padding: var(--sui-tree-panel-padding, 0);
}
.sui-tree-title { font-size: var(--sui-tree-title-size, 13px); color: var(--sui-tree-title-color, #e6e6ef); font-weight: 600; padding: 2px 4px 6px; }
.sui-tree-group { list-style: none; margin: 0; padding-left: var(--sui-tree-indent, 16px); }
.sui-tree > .sui-tree-group { padding-left: 0; }
.sui-tree-item { margin: 1px 0; }
.sui-tree-row { display: flex; align-items: center; gap: 4px; min-height: 28px; border-radius: var(--sui-tree-row-radius, 5px); padding: var(--sui-tree-row-padding, 2px 5px); background: var(--sui-tree-row-bg, transparent); cursor: pointer; }
.sui-tree-row:hover { background: var(--sui-tree-hover-bg, var(--sui-hover, #2a3143)); }
.sui-tree-item.selected > .sui-tree-row,
.sui-tree-item.sui-tree-pending > .sui-tree-row { background: var(--sui-tree-selected-bg, rgba(76, 92, 247, .28)); color: var(--sui-tree-selected-fg, inherit); }
.sui-tree-item:focus { outline: none; }
.sui-tree-item:focus-visible > .sui-tree-row { outline: 2px solid #4c5cf7; outline-offset: -2px; }
.sui-tree-toggle, .sui-tree-spacer { width: 20px; height: 20px; flex: none; display: inline-flex; align-items: center; justify-content: center; }
.sui-tree-toggle { border: 0; background: transparent; color: var(--sui-muted, #8fa0b8); padding: 0; cursor: pointer; font: inherit; }
.sui-tree-toggle:hover { color: var(--sui-fg, #e6e6ef); }
.sui-tree-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sui-tree-rename-btn {
  border: 0;
  background: transparent;
  color: var(--sui-muted, #8fa0b8);
  font: inherit;
  font-size: 13px;
  line-height: 1;
  cursor: pointer;
  padding: 1px 4px;
  opacity: .45;
}
.sui-tree-row:hover .sui-tree-rename-btn,
.sui-tree-rename-btn:focus { opacity: 1; }
.sui-tree-rename-btn:hover { color: #a5b4fc; }
.sui-tree-group[hidden] { display: none; }
.sui-tree-value { display: none; }
.sui-tree-editable > .sui-tree-row { cursor: text; }
.sui-tree-rename {
  min-width: 80px;
  flex: 1;
  background: #1d2129;
  color: var(--sui-fg, #e6e6ef);
  border: 1px solid #4c5cf7;
  border-radius: 4px;
  padding: 2px 5px;
  font: inherit;
}
.sui-tree-rename.invalid { border-color: #f87171; box-shadow: 0 0 0 2px rgba(248,113,113,.25); }
.sui-chart {
  color-scheme: dark;
  --sui-chart-gantt-period-a: rgba(148,163,184,.045);
  --sui-chart-gantt-period-b: rgba(148,163,184,.10);
  --sui-chart-table-selection-bg: rgba(56,189,248,.22);
  --sui-chart-table-selection-border: #38bdf8;
  --sui-chart-table-focus-row-bg: rgba(100,116,139,.14);
  --sui-chart-table-focus-cell-bg: rgba(100,116,139,.22);
  --sui-chart-table-focus-color: #64748b;
  --sui-chart-table-flash-bg: rgba(56,189,248,.20);
  --sui-chart-state-selection: #f59e0b;
  position: relative;
  display: flex;
  flex-direction: column;
  width: var(--sui-chart-width, 100%);
  height: var(--sui-chart-height, 100%);
  min-height: var(--sui-chart-min-height, 240px);
  max-width: 100%;
  min-width: 0;
  color: var(--sui-chart-fg, var(--sui-fg, #e6e6ef));
  background: var(--sui-chart-bg, transparent);
  border: var(--sui-chart-border, 0);
  border-radius: var(--sui-chart-radius, 0);
  padding: var(--sui-chart-padding, 0);
  overflow: hidden;
  font-family: var(--sui-chart-font, inherit);
}
.sui-chart.sui-chart-fixed-height { min-height: 0; }
.sui-chart[data-sui-chart-theme="dark"] {
  color-scheme: dark;
  --sui-chart-bg: var(--sui-bg, #14171c);
  --sui-chart-fg: var(--sui-fg, #e6e6e6);
  --sui-chart-title-color: var(--sui-fg, #e6e6e6);
  --sui-chart-muted: #8fa0b8;
  --sui-chart-grid: var(--sui-border, #3d4556);
  --sui-chart-axis: #6b7280;
  --sui-chart-control-bg: #1d2129;
  --sui-chart-control-border: var(--sui-border, #3d4556);
  --sui-chart-control-fg: var(--sui-fg, #e6e6e6);
  --sui-chart-control-hover-bg: var(--sui-hover, #2d3340);
  --sui-chart-control-hover-border: #556075;
  --sui-chart-control-hover-fg: #ffffff;
  --sui-chart-control-muted: #8fa0b8;
  --sui-chart-tooltip-bg: var(--sui-chart-control-bg, #1d2129);
  --sui-chart-tooltip-fg: var(--sui-chart-control-fg, var(--sui-fg, #e6e6e6));
  --sui-chart-tooltip-border: var(--sui-chart-control-border, var(--sui-border, #3d4556));
  --sui-chart-legend-color: var(--sui-chart-muted, #8fa0b8);
  --sui-chart-value-color: var(--sui-chart-fg, var(--sui-fg, #e6e6e6));
  --sui-chart-scroll-track: #1a1e26;
  --sui-chart-scroll-thumb: #3d4556;
  --sui-chart-table-head-bg: #1d2129;
  --sui-chart-table-sticky-bg: #1d2129;
  --sui-chart-table-input-bg: var(--sui-chart-control-bg, #1d2129);
  --sui-chart-table-selection-bg: rgba(56,189,248,.22);
  --sui-chart-table-selection-border: #38bdf8;
  --sui-chart-table-summary-bg: var(--sui-chart-control-bg, #1d2129);
  --sui-chart-table-flash-bg: rgba(56,189,248,.20);
  --sui-chart-gantt-period-a: rgba(148,163,184,.045);
  --sui-chart-gantt-period-b: rgba(148,163,184,.10);
  --sui-chart-gantt-weekend: #20252e;
  --sui-chart-gantt-group-bg: #2d3340;
  --sui-chart-gantt-dependency: #8fa0b8;
  --sui-chart-gantt-today: #fbbf24;
  --sui-chart-critical: #fbbf24;
  --sui-chart-error: #f87171;
  --sui-chart-table-focus-row-bg: rgba(100,116,139,.24);
  --sui-chart-table-focus-cell-bg: rgba(100,116,139,.34);
  --sui-chart-table-focus-color: #94a3b8;
  --sui-table-hover-bg: transparent;
}
.sui-chart[data-sui-chart-theme="light"] {
  color-scheme: light;
  --sui-chart-bg: #ffffff;
  --sui-chart-fg: #172033;
  --sui-chart-title-color: #172033;
  --sui-chart-muted: #5b6779;
  --sui-chart-grid: #dbe2ec;
  --sui-chart-axis: #9aa8ba;
  --sui-chart-control-bg: #f7f9fc;
  --sui-chart-control-border: #c5cfdd;
  --sui-chart-control-fg: #172033;
  --sui-chart-control-hover-bg: #e9eef6;
  --sui-chart-control-hover-border: #8fa3bb;
  --sui-chart-control-hover-fg: #0b1220;
  --sui-chart-control-muted: #6b7788;
  --sui-chart-tooltip-bg: var(--sui-chart-control-bg, #ffffff);
  --sui-chart-tooltip-fg: var(--sui-chart-control-fg, #172033);
  --sui-chart-tooltip-border: var(--sui-chart-control-border, #b8c4d4);
  --sui-chart-legend-color: var(--sui-chart-muted, #5b6779);
  --sui-chart-value-color: var(--sui-chart-fg, #172033);
  --sui-chart-scroll-track: #eef2f7;
  --sui-chart-scroll-thumb: #b7c3d3;
  --sui-chart-table-head-bg: #eef3fa;
  --sui-chart-table-sticky-bg: #eef3fa;
  --sui-chart-table-input-bg: var(--sui-chart-control-bg, #ffffff);
  --sui-chart-table-selection-bg: rgba(2,132,199,.18);
  --sui-chart-table-selection-border: #0284c7;
  --sui-chart-table-summary-bg: var(--sui-chart-table-head-bg, #f2f8ff);
  --sui-chart-table-flash-bg: rgba(56,189,248,.18);
  --sui-chart-gantt-period-a: #f2f7fc;
  --sui-chart-gantt-period-b: #e7f0f8;
  --sui-chart-gantt-weekend: #dbe2ec;
  --sui-chart-gantt-group-bg: #cbd5e1;
  --sui-chart-gantt-dependency: #7c8aa0;
  --sui-chart-gantt-today: #d97706;
  --sui-chart-critical: #d97706;
  --sui-chart-error: #dc2626;
  --sui-chart-table-focus-row-bg: #f1f5f9;
  --sui-chart-table-focus-cell-bg: #e2e8f0;
  --sui-chart-table-focus-color: #64748b;
  --sui-table-hover-bg: transparent;
}
.sui-chart:focus-visible { outline: 2px solid var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); outline-offset: 2px; }
.sui-chart-head button, .sui-chart-state-create-tools button, .sui-chart-zoom-btn, .sui-chart-zoom-reset, .sui-chart-zoom-fit, .sui-chart-menu-btn { font-family: var(--sui-chart-control-font, inherit); box-shadow: var(--sui-chart-control-shadow, none); }
.sui-chart-head button:not(.sui-chart-menu-btn), .sui-chart-state-create-tools button, .sui-chart-zoom-btn, .sui-chart-zoom-reset, .sui-chart-zoom-fit { font-size: var(--sui-chart-control-font-size, 12px); }
.sui-chart-head button:focus-visible, .sui-chart-state-create-tools button:focus-visible, .sui-chart-zoom-btn:focus-visible, .sui-chart-zoom-reset:focus-visible, .sui-chart-zoom-fit:focus-visible { outline: 2px solid var(--sui-chart-control-focus-border, var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7))); outline-offset: 1px; }
.sui-chart-head button:active, .sui-chart-state-create-tools button:active, .sui-chart-zoom-btn:active, .sui-chart-zoom-reset:active, .sui-chart-zoom-fit:active { background: var(--sui-chart-control-active-bg, var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444))); }
.sui-chart button { border-width: var(--sui-chart-control-border-width, 1px) !important; }
.sui-chart-table tbody tr { background: var(--sui-chart-table-row-bg, transparent); }
.sui-chart-table th { color: var(--sui-chart-table-head-fg, var(--sui-chart-fg, #e6e6ef)); }
.sui-chart-table th, .sui-chart-table td { border-bottom-color: var(--sui-chart-table-row-border, var(--sui-chart-grid, #2c3444)); }
.sui-chart-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: none;
  padding: var(--sui-chart-head-padding, 2px 4px 6px);
  background: var(--sui-chart-head-bg, transparent);
  min-height: var(--sui-chart-head-min-height, 0px);
}
.sui-chart-head button,
.sui-chart-head .sui-chart-state-create-tools button {
  height: 28px;
  min-height: 28px;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.sui-chart-title {
  font-weight: var(--sui-chart-title-weight, 600);
  font-family: var(--sui-chart-title-font, var(--sui-chart-font, inherit));
  font-size: var(--sui-chart-title-size, 14px);
  font-style: var(--sui-chart-title-style, normal);
  text-decoration: var(--sui-chart-title-decoration, none);
  text-shadow: var(--sui-chart-title-shadow, none);
  letter-spacing: var(--sui-chart-title-letter-spacing, normal);
  color: var(--sui-chart-title-color, var(--sui-chart-fg, #e6e6ef));
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sui-chart-title:empty { display: none; }
.sui-chart-menu-btn { margin-left: auto; border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, #475569)); background: var(--sui-chart-control-bg, transparent); color: var(--sui-chart-control-fg, var(--sui-chart-fg, #e6e6ef)); border-radius: var(--sui-chart-control-radius, 5px); font: inherit; font-size: 15px; line-height: 1; padding: 2px 6px; min-height: 24px; cursor: pointer; }
.sui-chart-menu-btn:hover { background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); }
.sui-chart-zoom { display: flex; align-items: center; gap: 3px; flex: none; }
.sui-chart-zoom[hidden] { display: none; }
.sui-chart-zoom-btn, .sui-chart-zoom-reset {
  border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, var(--sui-chart-muted, #3a4152)));
  background: var(--sui-chart-control-bg, transparent);
  color: var(--sui-chart-control-fg, var(--sui-chart-fg, #e6e6ef));
  border-radius: var(--sui-chart-control-radius, 5px);
  font: inherit;
  font-size: 12px;
  line-height: 1;
  padding: 4px 6px;
  min-height: 24px;
  cursor: pointer;
}
.sui-chart-zoom-btn:hover, .sui-chart-zoom-reset:hover {
  background: var(--sui-chart-control-hover-bg, transparent);
  color: var(--sui-chart-control-hover-fg, var(--sui-chart-control-fg, var(--sui-chart-fg, #e6e6ef)));
  border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7));
}
.sui-chart-zoom-btn:focus-visible, .sui-chart-zoom-reset:focus-visible, .sui-chart-zoom-fit:focus-visible {
  outline: 2px solid var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7));
  outline-offset: 1px;
}
.sui-chart-gantt-commit {
  border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, #475569));
  background: var(--sui-chart-control-bg, transparent);
  color: var(--sui-chart-control-fg, var(--sui-chart-fg, #e6e6ef));
  border-radius: var(--sui-chart-control-radius, 5px);
  font: inherit;
  font-size: 11px;
  line-height: 1;
  padding: 5px 8px;
  min-height: 24px;
  cursor: pointer;
}
.sui-chart-gantt-commit:hover:not(:disabled) { border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); }
.sui-chart-gantt-commit:disabled { opacity: .45; cursor: default; }
.sui-chart-gantt-commit:focus-visible { outline: 2px solid var(--sui-chart-series-1, #38bdf8); outline-offset: 1px; }
.sui-chart-gantt-commit[hidden] { display: none; }
.sui-chart-zoom-label {
  min-width: 38px;
  text-align: center;
  font-size: 11px;
  white-space: nowrap;
  color: var(--sui-chart-control-muted, var(--sui-chart-muted, #8fa0b8));
}
.sui-chart-zoom-fit {
  font-size: 11px;
  padding: 2px 8px;
  margin-left: 6px;
  border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, #475569));
  border-radius: var(--sui-chart-control-radius, 5px);
  background: var(--sui-chart-control-bg, transparent);
  color: var(--sui-chart-control-fg, var(--sui-chart-fg, #e6e6ef));
  cursor: pointer;
}
.sui-chart-zoom-fit:hover { background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); }
.sui-chart:not([data-sui-chart-type="state"]) .sui-chart-zoom-fit-all { display: none; }
.sui-chart-viewport {
  flex: 1;
  min-height: 0;
  min-width: 0;
  overflow: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--sui-chart-scroll-thumb, var(--sui-chart-axis, var(--sui-chart-muted, #3d4556))) var(--sui-chart-scroll-track, transparent);
}
.sui-chart[data-sui-chart-type="gantt"] .sui-chart-viewport { touch-action: pan-x pan-y; }
.sui-chart-viewport::-webkit-scrollbar,
.sui-chart-table::-webkit-scrollbar,
.sui-chart-state-tab-panel::-webkit-scrollbar { width: var(--sui-chart-scroll-size, 9px); height: var(--sui-chart-scroll-size, 9px); }
.sui-chart { scrollbar-color: var(--sui-chart-scroll-thumb, var(--sui-chart-axis, var(--sui-chart-muted, #3d4556))) var(--sui-chart-scroll-track, transparent); }
.sui-chart::-webkit-scrollbar { width: var(--sui-chart-scroll-size, 9px); height: var(--sui-chart-scroll-size, 9px); }
.sui-chart::-webkit-scrollbar-track { background: var(--sui-chart-scroll-track, transparent); }
.sui-chart::-webkit-scrollbar-thumb { background: var(--sui-chart-scroll-thumb, var(--sui-chart-axis, var(--sui-chart-muted, #3d4556))); border-radius: var(--sui-chart-scroll-radius, 6px); }
.sui-chart-viewport::-webkit-scrollbar-track,
.sui-chart-table::-webkit-scrollbar-track,
.sui-chart-state-tab-panel::-webkit-scrollbar-track { background: var(--sui-chart-scroll-track, transparent); }
.sui-chart-viewport::-webkit-scrollbar-thumb,
.sui-chart-table::-webkit-scrollbar-thumb,
.sui-chart-state-tab-panel::-webkit-scrollbar-thumb { background: var(--sui-chart-scroll-thumb, var(--sui-chart-axis, var(--sui-chart-muted, #3d4556))); border-radius: var(--sui-chart-scroll-radius, 6px); }
.sui-chart-stage { position: relative; min-width: 100%; min-height: 100%; }
.sui-chart-svg { display: block; width: 100%; height: 100%; overflow: visible; }
.sui-chart-ghost { position: absolute; inset: 0; display: block; width: 100% !important; height: 100% !important; pointer-events: none; opacity: 1; transition: opacity .2s ease; }
.sui-chart-ghost-out { opacity: 0; }
@media (prefers-reduced-motion: reduce) { .sui-chart-ghost { display: none; } }
.sui-chart-svg text { user-select: none; }
.sui-chart-grid { stroke: var(--sui-chart-grid, #2c3444); stroke-width: var(--sui-chart-grid-width, 1); }
.sui-chart-axis { stroke: var(--sui-chart-axis, #3a4152); stroke-width: var(--sui-chart-axis-width, 1); }
.sui-chart-axis-label { fill: var(--sui-chart-muted, #8fa0b8); font-size: var(--sui-chart-label-size, 11px); font-family: var(--sui-chart-font, inherit); }
.sui-chart-value-label { fill: var(--sui-chart-value-color, var(--sui-chart-fg, #e6e6ef)); font-size: var(--sui-chart-value-size, 10px); font-family: var(--sui-chart-font, inherit); }
.sui-chart-gantt-progress-label { paint-order: stroke fill; stroke: var(--sui-chart-gantt-label-bg, var(--sui-chart-bg, #14171c)); stroke-width: 3px; stroke-linejoin: round; }
.sui-chart-bar { opacity: var(--sui-chart-bar-opacity, .9); }
.sui-chart-line { fill: none !important; stroke-width: var(--sui-chart-line-width, 2.2); stroke-linejoin: round; stroke-linecap: round; }
.sui-chart-area { opacity: var(--sui-chart-area-opacity, .22); }
.sui-chart-point { stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; }
.sui-chart-waterfall-link { stroke: var(--sui-chart-muted, #8fa0b8); stroke-width: 1; stroke-dasharray: 3 3; }
.sui-chart-radar-area { opacity: var(--sui-chart-radar-opacity, .16); }
.sui-chart-gauge-track { fill: none; stroke: var(--sui-chart-axis, #3a4152); stroke-width: 12; stroke-linecap: round; }
.sui-chart-gauge-progress { fill: none; stroke: var(--sui-chart-series-1, #4c5cf7); stroke-width: 12; stroke-linecap: round; }
.sui-chart-gauge-value { fill: var(--sui-chart-fg, #e6e6ef); font-size: 26px; font-weight: 700; font-family: var(--sui-chart-font, inherit); }
.sui-chart-gauge-label { fill: var(--sui-chart-muted, #8fa0b8); font-size: 11px; font-family: var(--sui-chart-font, inherit); }
.sui-chart-funnel-stage { opacity: var(--sui-chart-funnel-opacity, .9); }
.sui-chart-heatmap-cell { stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; }
.sui-chart-heatmap-value { fill: var(--sui-chart-fg, #e6e6ef); font-size: 10px; font-weight: 600; paint-order: stroke; stroke: var(--sui-chart-bg, #14171c); stroke-width: 2px; font-family: var(--sui-chart-font, inherit); }
.sui-chart-treemap-cell { stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; opacity: var(--sui-chart-treemap-opacity, .9); }
.sui-chart-treemap-label { fill: var(--sui-chart-fg, #e6e6ef); font-size: 11px; font-weight: 600; paint-order: stroke; stroke: var(--sui-chart-bg, #14171c); stroke-width: 2px; font-family: var(--sui-chart-font, inherit); }
.sui-chart-boxplot-box { opacity: var(--sui-chart-box-opacity, .72); }
.sui-chart-boxplot-median { stroke: var(--sui-chart-fg, #e6e6ef); stroke-width: 2; }
.sui-chart-bullet-track { fill: var(--sui-chart-grid, #2c3444); opacity: .35; }
.sui-chart-bullet-target { stroke: var(--sui-chart-series-4, #ef4444); stroke-width: 2; }
.sui-chart-dumbbell-link { stroke: var(--sui-chart-muted, #8fa0b8); stroke-width: 2; }
.sui-chart-gantt-month-band,
.sui-chart-gantt-week-band,
.sui-chart-gantt-day-band { fill: var(--sui-chart-gantt-period-a, rgba(148,163,184,.045)); }
.sui-chart-gantt-month-band-alt,
.sui-chart-gantt-week-band-alt,
.sui-chart-gantt-day-band-alt { fill: var(--sui-chart-gantt-period-b, rgba(148,163,184,.10)); }
.sui-chart-gantt-weekend { fill: var(--sui-chart-gantt-weekend, var(--sui-chart-grid, #2c3444)); opacity: .18; }
.sui-chart-gantt-holiday { fill: var(--sui-chart-gantt-holiday, var(--sui-chart-series-4, #ef4444)); opacity: .13; }
.sui-chart-gantt-critical .sui-chart-gantt-task { stroke: var(--sui-chart-critical, #f59e0b); stroke-width: 3; }
.sui-chart-gantt-warning .sui-chart-gantt-task { stroke: var(--sui-chart-error, #f87171) !important; stroke-width: 3 !important; }
.sui-chart-gantt-warning-banner { display: flex; flex-wrap: wrap; gap: 3px 8px; align-items: baseline; margin: 0 4px 4px; padding: 5px 8px; border: 1px solid var(--sui-chart-error, #f87171); border-radius: 4px; background: color-mix(in srgb, var(--sui-chart-error, #f87171) 12%, transparent); color: var(--sui-chart-fg, #e6e6ef); font-size: 11px; flex: none; }
.sui-chart-gantt-warning-banner strong { color: var(--sui-chart-error, #f87171); flex: none; }
.sui-chart-gantt-warning-banner span { color: var(--sui-chart-muted, #8fa0b8); min-width: 0; overflow-wrap: anywhere; }
.sui-chart-gantt-warning { fill: var(--sui-chart-error, #f87171); }
.sui-chart-gantt-deadline { stroke: var(--sui-chart-error, #f87171); stroke-width: 3; stroke-dasharray: 3 2; pointer-events: none; }
.sui-chart-gantt-deadline-flag { fill: var(--sui-chart-error, #f87171); pointer-events: none; }
.sui-chart-gantt-deadline-group { cursor: default; }
.sui-chart-gantt-deadline-hit { fill: transparent; pointer-events: none; }
.sui-chart-interactive .sui-chart-gantt-deadline-group { cursor: ew-resize; }
.sui-chart-interactive .sui-chart-gantt-deadline-hit { pointer-events: all; }
.sui-chart-gantt-actual { fill: var(--sui-chart-fg, #e6e6ef); opacity: 0.9; pointer-events: none; }
.sui-chart-table tbody tr.sui-chart-table-row-warning td:first-child { border-left: 3px solid var(--sui-chart-error, #f87171); }
.sui-chart-bar-border .sui-chart-gantt-task,
.sui-chart-bar-border .sui-chart-gantt-summary,
.sui-chart-bar-border .sui-chart-gantt-milestone { stroke: var(--sui-chart-gantt-bar-border, transparent); }
.sui-chart-gantt-critical-dependency { stroke: var(--sui-chart-gantt-dependency-critical, #cbd5e1) !important; stroke-width: var(--sui-chart-gantt-link-critical-width, 1) !important; }
.sui-chart-gantt-arrow-open { fill: none !important; stroke: var(--sui-chart-gantt-dependency, #94a3b8); stroke-width: var(--sui-chart-gantt-link-width, 1); }
.sui-chart-gantt-arrow-open.sui-chart-gantt-critical-dependency { stroke: var(--sui-chart-gantt-dependency-critical, #cbd5e1) !important; stroke-width: var(--sui-chart-gantt-link-critical-width, 1) !important; }
.sui-chart-gantt-critical-dependency.sui-chart-gantt-dependency-head,
.sui-chart-gantt-dependency-dot.sui-chart-gantt-critical-dependency { fill: var(--sui-chart-gantt-dependency-critical, #cbd5e1) !important; }
.sui-chart-gantt-dependency-head.sui-chart-gantt-arrow-open { fill: none !important; }
.sui-chart-table tbody tr.sui-chart-table-row-critical td:first-child { border-left: 3px solid var(--sui-chart-critical, #f59e0b); }
.sui-chart-table tbody tr.sui-chart-table-row-summary td { font-weight: 700; background: var(--sui-chart-table-summary-bg, rgba(56,189,248,.07)); }
.sui-chart-gantt-deps { display: inline-flex; flex-wrap: wrap; gap: 3px; align-items: center; min-width: 90px; }
.sui-chart-gantt-dep-chip { display: inline-flex; align-items: center; gap: 2px; padding: 1px 4px; border-radius: 3px; background: var(--sui-chart-table-selection-bg, rgba(56,189,248,.22)); border: 1px solid var(--sui-chart-table-selection-border, #38bdf8); font-size: 10px; }
.sui-chart-gantt-dep-remove { border: 0; background: transparent; color: inherit; cursor: pointer; padding: 0 1px; font-size: 11px; line-height: 1; }
.sui-chart-gantt-dep-add { min-width: 82px; max-width: 120px; }
.sui-chart-gantt-resources { display: inline-flex; flex-wrap: wrap; gap: 3px; align-items: center; min-width: 100px; }
.sui-chart-gantt-resource-chip { display: inline-flex; align-items: center; gap: 2px; padding: 1px 4px; border-radius: 3px; background: var(--sui-chart-table-selection-bg, rgba(56,189,248,.22)); border: 1px solid var(--sui-chart-table-selection-border, #38bdf8); font-size: 10px; }
.sui-chart-gantt-resource-units { width: 34px !important; min-width: 0 !important; height: 18px; padding: 0 2px !important; text-align: center; font-size: 10px; }
.sui-chart-gantt-resource-remove { border: 0; background: transparent; color: inherit; cursor: pointer; padding: 0 1px; font-size: 11px; line-height: 1; }
.sui-chart-gantt-resource-add-btn { flex: none; width: 18px; height: 18px; padding: 0; border: 1px solid var(--sui-chart-axis, #475569); border-radius: 3px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-fg, #e6e6ef); cursor: pointer; line-height: 16px; }
.sui-chart-gantt-group-toggles { display: flex; flex-wrap: wrap; gap: 4px; padding: 5px 6px 2px; border-bottom: 1px solid var(--sui-chart-grid, #2c3444); }
.sui-chart-gantt-group-toggle { border: 1px solid var(--sui-chart-grid, #2c3444); border-radius: 3px; background: transparent; color: var(--sui-chart-muted, #8fa0b8); cursor: pointer; padding: 2px 6px; font-size: 10px; }
.sui-chart-gantt-group-toggle:hover { color: var(--sui-chart-fg, #e6e6ef); border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); }
.sui-chart-gantt-wbs-table-toggle { border: 0; background: transparent; color: var(--sui-chart-muted, #8fa0b8); cursor: pointer; padding: 0 3px 0 0; font-size: 11px; line-height: 1; }
.sui-chart-gantt-wbs-table-toggle:hover { color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-gantt-progress { fill: var(--sui-chart-gantt-progress, var(--sui-chart-fg, #e6e6ef)); opacity: .68; pointer-events: none; }
.sui-chart-gantt-sticky { pointer-events: auto; }
.sui-chart-gantt-sticky-bg { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-grid, #2c3444); stroke-width: .5; }
.sui-chart-gantt-name-resize { cursor: col-resize; touch-action: none; }
.sui-chart-gantt-name-resize-hit { fill: transparent; pointer-events: all; }
.sui-chart-gantt-name-resize-line { stroke: var(--sui-chart-grid, #2c3444); stroke-width: 1; pointer-events: none; }
.sui-chart-gantt-name-resize:hover .sui-chart-gantt-name-resize-line,
.sui-chart-gantt-name-resizing .sui-chart-gantt-name-resize-line { stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2; }
.sui-chart-gantt-name-resizing, .sui-chart-gantt-name-resizing * { cursor: col-resize !important; user-select: none; }
.sui-chart-gantt-task.sui-chart-gantt-status-not-started { fill: var(--sui-chart-gantt-status-not-started, #94a3b8) !important; }
.sui-chart-gantt-task.sui-chart-gantt-status-in-progress { fill: var(--sui-chart-gantt-status-in-progress, #38bdf8) !important; }
.sui-chart-gantt-task.sui-chart-gantt-status-completed { fill: var(--sui-chart-gantt-status-completed, #22c55e) !important; }
.sui-chart-gantt-task.sui-chart-gantt-status-blocked { fill: var(--sui-chart-gantt-status-blocked, #f97316) !important; }
.sui-chart-gantt-task.sui-chart-gantt-status-late { fill: var(--sui-chart-gantt-status-late, #ef4444) !important; }
.sui-chart-gantt-milestone { stroke: var(--sui-chart-gantt-bar-border, var(--sui-chart-bg, #14171c)); stroke-width: 1; }
.sui-chart-gantt-dependency { stroke: var(--sui-chart-gantt-dependency, #94a3b8); stroke-width: var(--sui-chart-gantt-link-width, 1); stroke-linecap: round; stroke-linejoin: round; opacity: .92; pointer-events: stroke; }
.sui-chart-gantt-dependency-dot { fill: var(--sui-chart-gantt-dependency, #94a3b8); stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; }
.sui-chart-gantt-link-ss { stroke-dasharray: 5 3; }
.sui-chart-gantt-link-ff { stroke-dasharray: 2 3; }
.sui-chart-gantt-link-sf { stroke-dasharray: 7 2 2 2; }
.sui-chart-gantt-dependency-head { fill: var(--sui-chart-gantt-dependency, #94a3b8); stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; }
.sui-chart-gantt-link-group { cursor: context-menu; outline: none; }
.sui-chart-gantt-link-hit { fill: none; stroke: transparent; stroke-width: 14; pointer-events: stroke; }
.sui-chart-gantt-link-group:hover .sui-chart-gantt-dependency, .sui-chart-gantt-link-group:focus .sui-chart-gantt-dependency { stroke-width: 2.5; filter: drop-shadow(0 0 2px var(--sui-chart-series-1, #38bdf8)); }
.sui-chart-gantt-link-group:focus .sui-chart-gantt-link-hit { stroke: var(--sui-chart-series-1, #38bdf8); opacity: .18; }
.sui-chart-gantt-link-menu { position: fixed; z-index: 10030; min-width: 190px; padding: 4px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); border: 1px solid var(--sui-chart-axis, #475569); border-radius: 5px; box-shadow: 0 8px 20px rgba(0,0,0,.35); }
.sui-chart-gantt-link-menu-title { padding: 4px 7px 5px; color: var(--sui-chart-muted, #8fa0b8); font-size: 10px; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 280px; }
.sui-chart-gantt-link-menu button { display: block; width: 100%; padding: 5px 8px; border: 0; border-radius: 4px; background: transparent; color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 12px; text-align: left; cursor: pointer; }
.sui-chart-gantt-link-menu button:hover, .sui-chart-gantt-link-menu button:focus-visible { background: var(--sui-chart-control-hover-bg, #2c3444); outline: none; }
.sui-chart-gantt-today { stroke: var(--sui-chart-gantt-today, var(--sui-chart-series-3, #f59e0b)); stroke-width: 2; stroke-dasharray: 4 3; }
.sui-chart-gantt-group-band { fill: var(--sui-chart-gantt-group-bg, var(--sui-chart-grid, #2c3444)); opacity: .16; }
.sui-chart-gantt-summary { stroke: var(--sui-chart-gantt-bar-border, var(--sui-chart-bg, #14171c)); stroke-width: 1.5; }
.sui-chart-gantt-wbs-header { font-weight: 700; fill: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-gantt-wbs-band { opacity: .12; }
.sui-chart-gantt-group { font-weight: 700; fill: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-gantt-year { font-weight: 700; fill: var(--sui-chart-muted, #8fa0b8); letter-spacing: .04em; }
.sui-chart-gantt-progress-column { fill: var(--sui-chart-muted, #8fa0b8); font-variant-numeric: tabular-nums; }
.sui-chart-state-shape { stroke: var(--sui-chart-bg, #14171c); stroke-width: 1.5; opacity: var(--sui-chart-state-opacity, .92); }

.sui-chart-state-label { fill: var(--sui-chart-fg, #e6e6ef); font-weight: 700; font-size: var(--sui-chart-state-label-size, 12.5px); paint-order: stroke fill; stroke: var(--sui-chart-bg, #14171c); stroke-width: 2.5px; stroke-linejoin: round; }
.sui-chart-state-description { fill: var(--sui-chart-muted, #8fa0b8); font-size: 10.5px; paint-order: stroke fill; stroke: var(--sui-chart-bg, #14171c); stroke-width: 2px; }
.sui-chart-state-edge-line { stroke: var(--sui-chart-gantt-dependency, var(--sui-chart-link-color, #94a3b8)); stroke-width: 1.5; fill: none; }
.sui-chart-state-edge-hit { stroke: transparent; stroke-width: 14; fill: none; cursor: pointer; }
.sui-chart-state-edge-label { fill: var(--sui-chart-fg, #e6e6ef); font-size: 11px; font-weight: 600; paint-order: stroke fill; stroke: var(--sui-chart-bg, #14171c); stroke-width: 3px; stroke-linejoin: round; }
.sui-chart-state-arrow { fill: var(--sui-chart-gantt-dependency, #94a3b8); }
.sui-chart-state-arrow-open { fill: none; stroke: var(--sui-chart-gantt-dependency, #94a3b8); stroke-width: 1.5; }
.sui-chart-state-node-group:focus, .sui-chart-state-edge-label:focus { outline: none; }
.sui-chart-state-node-group:focus-visible .sui-chart-state-shape { stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 3; }
.sui-chart-state-edge-label:focus-visible { text-decoration: underline; }
.sui-chart-state-tag-bg { fill: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); stroke: var(--sui-chart-control-border, var(--sui-chart-grid, #2c3444)); stroke-width: .9; }
.sui-chart-state-tag-internal .sui-chart-state-tag-bg { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-muted, #8fa0b8); stroke-width: 1; }
.sui-chart-state-tag-internal .sui-chart-state-edge-label { paint-order: normal; stroke: none; }
.sui-chart-state-tag-anchor { fill: var(--sui-chart-state-selection, #f59e0b); stroke: var(--sui-chart-bg, #14171c); stroke-width: 1.5; opacity: 0; pointer-events: none; cursor: crosshair; }
.sui-chart-state-selected .sui-chart-state-tag-anchor { opacity: 1; pointer-events: auto; }
.sui-chart-state-tag-rotate { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-state-selection, #f59e0b); stroke-width: 1.5; opacity: 0; pointer-events: none; cursor: grab; }
.sui-chart-state-tag-selected .sui-chart-state-tag-rotate { opacity: 1; pointer-events: auto; }
.sui-chart-state-tag-selected .sui-chart-state-tag-rotate:hover { fill: var(--sui-chart-state-selection, #f59e0b); }
.sui-chart-state-tag-link { stroke: var(--sui-chart-muted, #8fa0b8); stroke-width: 1; stroke-dasharray: 2 2; pointer-events: none; }
.sui-chart-state-tag-underline { stroke: var(--sui-chart-control-border, #64748b); stroke-width: 1.2; pointer-events: none; }
.sui-chart-state-tag-arrow { fill: var(--sui-chart-muted, #8fa0b8); pointer-events: none; }
.sui-chart-state-color { width: 18px; height: 18px; padding: 0; margin: 0 5px 0 0; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; background: transparent; cursor: pointer; vertical-align: middle; }
.sui-chart-state-color-readonly { display: inline-block; width: 18px; height: 18px; margin: 0 5px 0 0; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; vertical-align: middle; }
.sui-chart-state-cell-text { vertical-align: middle; }
.sui-chart-state-color::-webkit-color-swatch-wrapper { padding: 2px; }
.sui-chart-state-color::-webkit-color-swatch { border: 0; border-radius: 2px; }
.sui-chart-state-color::-moz-color-swatch { border: 0; border-radius: 2px; }
.sui-chart-state-node-group { filter: drop-shadow(0 1px 2px rgba(0,0,0,.16)); }
.sui-chart-state-selected .sui-chart-state-shape { stroke: var(--sui-chart-state-selection, #f59e0b) !important; stroke-width: 3 !important; filter: drop-shadow(0 0 4px var(--sui-chart-state-selection, #f59e0b)); }
.sui-chart-state-edge.sui-chart-state-selected .sui-chart-state-edge-line { stroke: var(--sui-chart-state-selection, #f59e0b) !important; stroke-width: 3 !important; }
.sui-chart-state-edge.sui-chart-state-selected .sui-chart-state-arrow { fill: var(--sui-chart-state-selection, #f59e0b) !important; stroke: var(--sui-chart-state-selection, #f59e0b) !important; }
.sui-chart-state-edge.sui-chart-state-selected .sui-chart-state-arrow-open { fill: none !important; stroke: var(--sui-chart-state-selection, #f59e0b) !important; }
.sui-chart-state-link-endpoint { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2; opacity: 0; pointer-events: none; cursor: grab; transition: opacity .12s ease; }
.sui-chart-state-link-endpoint.sui-chart-state-handle-visible { opacity: 1; pointer-events: all; }
.sui-chart-state-link-endpoint:active { cursor: grabbing; }
.sui-chart-state-link-endpoint:focus-visible { outline: none; stroke-width: 3; }
.sui-chart-state-resize { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 1.5; opacity: 0; pointer-events: none; cursor: nwse-resize; }
.sui-chart-state-selected .sui-chart-state-resize { opacity: 1; pointer-events: auto; }
.sui-chart-state-rotate-line { stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 1; opacity: 0; pointer-events: none; }
.sui-chart-state-rotate { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-state-selection, #f59e0b); stroke-width: 1.5; cursor: grab; opacity: 0; pointer-events: none; }
.sui-chart-state-selected .sui-chart-state-rotate-line { opacity: .55; }
.sui-chart-state-selected .sui-chart-state-rotate { opacity: 1; pointer-events: auto; }
.sui-chart-state-selected .sui-chart-state-rotate:hover { fill: var(--sui-chart-state-selection, #f59e0b); }
.sui-chart-state-grip, .sui-chart-state-anchor { fill: var(--sui-chart-series-1, #38bdf8); stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; opacity: 0; pointer-events: none; }
.sui-chart-state-selected .sui-chart-state-grip,
.sui-chart-state-selected .sui-chart-state-anchor { opacity: .9; }
/* View-only diagrams: no anchors, grips or editing handles. */
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-anchor,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-grip,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-resize,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-rotate,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-rotate-line,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-link-endpoint,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-waypoint,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-edge-handle,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-tag-anchor,
.sui-chart[data-sui-chart-state-handles="0"] .sui-chart-state-tag-rotate { display: none !important; }
.sui-chart-state-marquee {
  position: fixed;
  z-index: 10000;
  pointer-events: none;
  border: 1px solid var(--sui-chart-series-1, #38bdf8);
  background: rgba(56,189,248,.14);
  border-radius: 2px;
}
.sui-chart-state-guides { pointer-events: none; }
.sui-chart-state-align-guide {
  stroke: var(--sui-chart-series-1, #38bdf8);
  stroke-width: 1;
  stroke-dasharray: 4 4;
  opacity: .85;
  pointer-events: none;
}
.sui-chart-state-align-toolbar {
  position: absolute;
  top: 46px;
  right: 12px;
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px;
  border: 1px solid var(--sui-chart-control-border, #475569);
  border-radius: 9px;
  background: var(--sui-chart-control-bg, #1f2633);
  box-shadow: 0 8px 24px rgba(2,6,23,.28);
}
.sui-chart-state-align-toolbar[hidden] { display: none; }
.sui-chart-state-align-toolbar-group { display: flex; gap: 2px; }
.sui-chart-state-align-toolbar-group + .sui-chart-state-align-toolbar-group {
  padding-left: 5px;
  border-left: 1px solid var(--sui-chart-control-border, #475569);
}
.sui-chart-state-align-toolbar button {
  min-width: 24px;
  height: 24px;
  padding: 0 5px;
  border: 1px solid transparent;
  border-radius: 5px;
  background: transparent;
  color: var(--sui-chart-fg, #e6e6ef);
  font-size: 10.5px;
  font-weight: 700;
  cursor: pointer;
}
.sui-chart-state-align-toolbar button:hover,
.sui-chart-state-align-toolbar button:focus-visible {
  border-color: var(--sui-chart-series-1, #38bdf8);
  background: rgba(56,189,248,.14);
  outline: none;
}
.sui-chart-state-align-toolbar button:disabled { opacity: .35; cursor: not-allowed; }
.sui-chart-state-align-icon { width: 16px; height: 16px; display: block; }
.sui-chart-state-align-icon-block { fill: currentColor; opacity: .82; }
.sui-chart-state-align-icon-guide { stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 1.3; stroke-linecap: round; }
.sui-chart-state-link-mode .sui-chart-state-node-group:hover .sui-chart-state-anchor,
.sui-chart-state-link-mode .sui-chart-state-selected .sui-chart-state-anchor { opacity: .55; }
.sui-chart-state-link-mode .sui-chart-state-anchor-active {
  opacity: 1 !important;
  fill: var(--sui-chart-state-selection, #f59e0b);
  stroke: var(--sui-chart-bg, #14171c);
  stroke-width: 1.5;
}
.sui-chart-state-grid { stroke: var(--sui-chart-grid, #2c3444); stroke-width: .5; opacity: .35; pointer-events: none; }
.sui-chart-state-properties-dialog { width: min(720px, 100%); }
.sui-chart-state-properties-dialog .sui-chart-link-fields { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px 14px; }
.sui-chart-state-properties-dialog .sui-chart-link-fields label { min-width: 0; }
.sui-chart-state-properties-dialog .sui-chart-state-cell-input, .sui-chart-state-properties-dialog .sui-chart-state-color { width: 100%; box-sizing: border-box; }
.sui-chart-state-properties-dialog .sui-chart-state-color { max-width: 72px; }
.sui-chart-calendar-dialog.sui-chart-state-tag-dialog { width: min(720px, 100%); padding: 16px 18px; border-radius: 14px; box-shadow: 0 26px 80px rgba(2,6,23,.42); }
.sui-chart-calendar-dialog.sui-chart-link-settings-dialog { width: min(600px, 100%); }
.sui-chart-calendar-dialog.sui-chart-shape-settings-dialog { width: min(600px, 100%); }
.sui-chart-shape-settings-grid { display: flex; flex-direction: column; gap: 7px; }
.sui-chart-shape-settings-row { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; align-items: stretch; }
.sui-chart-shape-settings-row > label {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: .02em;
  color: var(--sui-chart-muted, #8fa0b8);
}
.sui-chart-shape-settings-row > label .sui-chart-state-cell-input,
.sui-chart-shape-settings-row > label .sui-chart-number-field,
.sui-chart-shape-settings-row > label .sui-chart-color-field { width: 100%; }
.sui-chart-shape-settings-row > label .sui-chart-switch { align-self: flex-start; }
.sui-chart-shape-settings-placeholder { visibility: hidden; }
.sui-chart-state-settings-dialog .sui-chart-tag-panes { width: 100%; min-width: 0; }
.sui-chart-state-settings-dialog .sui-chart-tag-pane { width: 100%; box-sizing: border-box; }
.sui-chart-settings-actions-row { display: grid; width: 100%; grid-template-columns: repeat(auto-fit, minmax(64px, 1fr)); gap: 6px; }
.sui-chart-settings-actions-row .sui-chart-settings-action { text-align: center; padding: 7px 4px; }
.sui-chart-calendar-dialog.sui-chart-grid-settings-dialog { width: min(480px, 100%); }
.sui-chart-calendar-dialog.sui-chart-state-settings-dialog { width: min(560px, 100%); }
.sui-chart-calendar-dialog.sui-chart-state-theme-dialog { width: min(720px, 100%); }
.sui-chart-state-settings-dialog .sui-chart-tag-pane { display: flex; flex-direction: column; gap: 7px; }
.sui-chart-state-settings-dialog .sui-chart-tag-pane > label {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: .02em;
  color: var(--sui-chart-muted, #8fa0b8);
}
.sui-chart-state-settings-dialog .sui-chart-tag-pane > label .sui-chart-switch { align-self: flex-start; }
.sui-chart-settings-action {
  display: block;
  width: 100%;
  box-sizing: border-box;
  text-align: left;
  padding: 8px 10px;
  border: 1px solid var(--sui-chart-control-border, #475569);
  border-radius: 8px;
  background: var(--sui-chart-control-bg, #1f2633);
  color: var(--sui-chart-fg, #e6e6ef);
  font-size: 11.5px;
  font-weight: 600;
  letter-spacing: .01em;
  cursor: pointer;
}
.sui-chart-settings-action:hover, .sui-chart-settings-action:focus-visible {
  border-color: var(--sui-chart-series-1, #38bdf8);
  background: rgba(56,189,248,.12);
  outline: none;
}
.sui-chart-settings-action.primary {
  border-color: var(--sui-chart-series-1, #38bdf8);
  background: rgba(56,189,248,.18);
}
.sui-chart-link-settings-body, .sui-chart-grid-settings-body { display: block; }
.sui-chart-link-settings-body .sui-chart-tag-editor, .sui-chart-grid-settings-body .sui-chart-tag-editor { max-height: none; padding-right: 0; }
.sui-chart-calendar-dialog.sui-chart-state-tag-dialog h3 { margin: 0 0 2px; font-size: 17px; letter-spacing: -.01em; }
.sui-chart-calendar-dialog.sui-chart-state-tag-dialog .sui-chart-calendar-hint { margin: 0 0 10px; line-height: 1.35; font-size: 11.5px; }
.sui-chart-state-tag-body { display: grid; grid-template-columns: minmax(0, 1fr) 280px; gap: 12px; align-items: start; }
.sui-chart-tag-editor { display: flex; flex-direction: column; gap: 8px; min-width: 0; max-height: min(68vh, 620px); overflow: auto; padding: 0 4px 0 0; }
.sui-chart-tag-section { border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 10px; padding: 8px 10px 10px; background: rgba(148,163,184,.045); }
.sui-chart-tag-section h4 { margin: 0 0 7px; font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: var(--sui-chart-fg, #111827); }
.sui-chart-tag-tabs { display: flex; flex-wrap: wrap; gap: 3px; padding: 3px; margin-bottom: 7px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 10px; background: rgba(148,163,184,.06); }
.sui-chart-tag-tab { border: 1px solid transparent; border-radius: 7px; padding: 5px 8px; background: transparent; color: var(--sui-chart-fg, #111827); opacity: .62; font-size: 10.5px; font-weight: 600; letter-spacing: .02em; cursor: pointer; }
.sui-chart-tag-tab:hover, .sui-chart-tag-tab:focus-visible { opacity: 1; background: rgba(148,163,184,.12); outline: none; }
.sui-chart-tag-tab-active { opacity: 1; background: var(--sui-chart-control-bg, #1f2633); border-color: var(--sui-chart-series-1, #38bdf8); }
.sui-chart-tag-tab-disabled { opacity: .35; }
.sui-chart-tag-panes { display: grid; min-width: 0; }
.sui-chart-tag-pane { grid-area: 1 / 1; min-width: 0; visibility: hidden; pointer-events: none; }
.sui-chart-tag-pane-active { visibility: visible; pointer-events: auto; }
.sui-chart-tag-context { display: grid; grid-template-columns: 1fr; gap: 6px; margin-bottom: 7px; }
.sui-chart-tag-section.sui-chart-section-disabled { opacity: .45; }
.sui-chart-tag-section.sui-chart-section-disabled h4 { color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-field-disabled { opacity: .45; }
.sui-chart-tag-editor select:disabled, .sui-chart-tag-editor input:disabled { cursor: not-allowed; }
.sui-chart-tag-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 7px 10px; align-items: start; }
.sui-chart-tag-col-left { grid-column: 1; }
.sui-chart-tag-col-right { grid-column: 2; }
.sui-chart-tag-grid label, .sui-chart-tag-context label { display: flex; flex-direction: column; gap: 3px; min-width: 0; font-size: 10.5px; font-weight: 600; letter-spacing: .02em; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-tag-grid .sui-chart-tag-field-wide { grid-column: 1 / -1; }
.sui-chart-tag-grid label.sui-chart-setting-row {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr);
  gap: 10px;
  align-items: center;
  min-width: 0;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: .02em;
  color: var(--sui-chart-muted, #8fa0b8);
}
.sui-chart-setting-label { min-width: 0; overflow-wrap: anywhere; }
.sui-chart-state-theme-dialog .sui-chart-tag-grid label.sui-chart-setting-row { grid-column: auto; display: grid; grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.5fr); gap: 8px; align-items: center; }
.sui-chart-two-col-fields label.sui-chart-setting-row { grid-column: auto; display: grid; grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.5fr); gap: 8px; align-items: center; }
.sui-chart-two-col-fields, .sui-chart-state-theme-dialog .sui-chart-tag-grid { grid-auto-flow: row dense; }
.sui-chart-two-col-fields label.sui-chart-setting-row.sui-chart-field-color,
.sui-chart-state-theme-dialog .sui-chart-tag-grid label.sui-chart-setting-row.sui-chart-field-color { grid-column: 1; }
.sui-chart-two-col-fields label.sui-chart-setting-row.sui-chart-field-number,
.sui-chart-state-theme-dialog .sui-chart-tag-grid label.sui-chart-setting-row.sui-chart-field-number { grid-column: 2; }
.sui-chart-setting-control { display: flex; align-items: center; justify-content: flex-end; min-width: 0; }
.sui-chart-setting-control > * { width: 100%; min-width: 0; }
.sui-chart-setting-control > .sui-chart-switch { width: auto; margin-left: 0; margin-right: auto; }
.sui-chart-setting-control .sui-chart-color-field,
.sui-chart-setting-control .sui-chart-number-field { width: 100%; }
.sui-chart-hint-settings-body { display: block; }
.sui-chart-hint-settings-body .sui-chart-tag-editor { max-height: none; padding-right: 0; }
.sui-chart-calendar-dialog.sui-chart-hint-settings-dialog { width: min(780px, 100%); }
.sui-chart-hint-groups { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; align-items: start; }
.sui-chart-hint-group { min-width: 0; display: flex; flex-direction: column; gap: 6px; padding: 8px 10px 10px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 10px; background: rgba(148,163,184,.045); }
.sui-chart-hint-group label.sui-chart-setting-row { column-gap: 4px; }
.sui-chart-hint-group-title { margin: 0; font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: var(--sui-chart-muted, #8fa0b8); }
@media (max-width: 640px) { .sui-chart-hint-groups { grid-template-columns: 1fr; } }
.sui-chart-tag-editor .sui-chart-state-cell-input, .sui-chart-tag-editor select { width: 100%; box-sizing: border-box; height: 26px; border-radius: 6px; font-size: 11px; }
select.sui-chart-state-cell-input, .sui-chart-tag-editor select { appearance: none; -webkit-appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6' viewBox='0 0 10 6'%3E%3Cpath d='M1 1l4 4 4-4' fill='none' stroke='%2364748b' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 7px center; background-size: 9px 5px; padding-right: 22px; cursor: pointer; }
.sui-chart-tag-editor .sui-chart-state-color { width: 100%; max-width: none; height: 26px; box-sizing: border-box; border-radius: 6px; }
.sui-chart-color-field { display: grid; grid-template-columns: 30px minmax(0, 1fr); gap: 6px; align-items: center; }
.sui-chart-color-alpha-field { display: grid; grid-template-columns: 30px minmax(0, 1fr) 48px; gap: 6px; align-items: center; width: 100%; box-sizing: border-box; }
.sui-chart-color-alpha-range { width: 100%; min-width: 0; accent-color: var(--sui-chart-series-1, #38bdf8); }
.sui-chart-color-alpha-text { width: 100%; min-width: 0; text-align: right; font-variant-numeric: tabular-nums; padding-left: 2px; padding-right: 2px; }
.sui-chart-color-swatch { width: 30px; height: 26px; padding: 0; box-sizing: border-box; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 6px; background-color: transparent; background-image: linear-gradient(45deg, #cbd5e1 25%, transparent 25%), linear-gradient(-45deg, #cbd5e1 25%, transparent 25%), linear-gradient(45deg, transparent 75%, #cbd5e1 75%), linear-gradient(-45deg, transparent 75%, #cbd5e1 75%); background-size: 8px 8px; background-position: 0 0, 0 4px, 4px -4px, -4px 0; cursor: pointer; }
.sui-chart-color-text { width: 100%; min-width: 0; font-family: ui-monospace, SFMono-Regular, Consolas, monospace; font-size: 10.5px; }
.sui-chart-color-picker { position: fixed; z-index: 10060; width: 230px; box-sizing: border-box; padding: 10px; display: flex; flex-direction: column; gap: 8px; background: var(--sui-chart-control-bg, #1e293b); color: var(--sui-chart-control-fg, #f8fafc); border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 10px; box-shadow: 0 18px 50px rgba(0,0,0,.45); }
.sui-chart-color-picker-input { width: 100%; height: 36px; padding: 2px; box-sizing: border-box; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 8px; background: transparent; cursor: pointer; }
.sui-chart-color-picker-alpha { display: flex; align-items: center; gap: 8px; font-size: 10px; font-weight: 600; color: var(--sui-chart-control-fg, #f8fafc); }
.sui-chart-color-picker-alpha input[type="range"] { flex: 1 1 auto; min-width: 0; accent-color: var(--sui-chart-series-1, #38bdf8); }
.sui-chart-color-picker-alpha-text { width: 46px; min-width: 0; text-align: right; font-variant-numeric: tabular-nums; }
.sui-chart-color-picker-close { align-self: flex-end; height: 24px; padding: 0 10px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 6px; background: var(--sui-chart-control-bg, #1e293b); color: var(--sui-chart-control-fg, #f8fafc); font: inherit; font-size: 10px; font-weight: 600; cursor: pointer; }
.sui-chart-number-field { display: grid; grid-template-columns: minmax(0, 1fr) 54px; gap: 6px; align-items: center; }
.sui-chart-number-range { width: 100%; min-width: 0; box-sizing: border-box; accent-color: var(--sui-chart-series-1, #38bdf8); }
.sui-chart-number-input { min-width: 0; text-align: right; font-variant-numeric: tabular-nums; }
.sui-chart-tag-editor .sui-chart-number-input { min-width: 0; }
.sui-chart-link-preview-field { display: grid; grid-template-columns: minmax(0, 1fr) 118px; gap: 8px; align-items: center; }
.sui-chart-link-preview-field select { min-width: 0; }
.sui-chart-link-preview { width: 100%; height: 20px; color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-link-preview-line { stroke: currentColor; stroke-width: 2; opacity: .72; }
.sui-chart-link-preview-shape { fill: currentColor; }
.sui-chart-link-preview-open { fill: none; stroke: currentColor; stroke-width: 2; stroke-linejoin: round; }
.sui-chart-direction-button { text-align: left; cursor: pointer; }
.sui-chart-direction-button:hover, .sui-chart-direction-button:focus-visible { background: transparent; border-color: var(--sui-chart-series-1, #38bdf8); outline: none; }
.sui-chart-state-cell-input.sui-chart-direction-button:focus { background: transparent; }
.sui-chart-direction-field { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.sui-chart-field-caption { font-size: 10px; font-weight: 600; opacity: .75; }
.sui-chart-link-arrow-preview-box { display: flex; flex-direction: column; gap: 6px; grid-row: span 3; min-width: 0; }
.sui-chart-link-arrow-preview { width: 100%; height: 58px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 8px; background: rgba(148,163,184,.06); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-link-preview-caption { font-size: 9px; text-transform: uppercase; letter-spacing: .08em; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-shape-preview-box { margin-top: 10px; display: flex; flex-direction: column; gap: 6px; }
.sui-chart-shape-preview { width: 100%; height: 130px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 8px; background: rgba(148,163,184,.06); }
.sui-chart-shape-preview-shape { vector-effect: non-scaling-stroke; }
.sui-chart-switch { display: inline-flex; align-items: center; gap: 6px; cursor: pointer; min-height: 26px; }
.sui-chart-switch-input { position: absolute; width: 1px; height: 1px; opacity: 0; }
.sui-chart-switch-track { position: relative; flex: 0 0 auto; width: 32px; height: 18px; border-radius: 999px; border: 1px solid var(--sui-chart-control-border, #475569); background: var(--sui-chart-control-bg, #1f2633); transition: background .15s, border-color .15s; }
.sui-chart-switch-knob { position: absolute; left: 2px; top: 2px; width: 12px; height: 12px; border-radius: 50%; background: var(--sui-chart-muted, #8fa0b8); transition: transform .15s, background .15s; }
.sui-chart-switch-input:checked + .sui-chart-switch-track { background: var(--sui-chart-series-1, #38bdf8); border-color: var(--sui-chart-series-1, #38bdf8); }
.sui-chart-switch-input:checked + .sui-chart-switch-track .sui-chart-switch-knob { transform: translateX(14px); background: #fff; }
.sui-chart-switch-input:focus-visible + .sui-chart-switch-track { outline: 2px solid var(--sui-chart-series-1, #38bdf8); outline-offset: 2px; }
.sui-chart-switch-label { font-size: 10.5px; min-width: 18px; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-font-field { position: relative; display: block; }
.sui-chart-font-list { display: none; position: fixed; z-index: 10000; max-height: 170px; overflow: auto; padding: 3px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 8px; background: var(--sui-chart-control-bg, #1f2633); box-shadow: 0 14px 30px rgba(2,6,23,.35); }
.sui-chart-font-list.sui-chart-font-list-open { display: block; }
.sui-chart-font-option { display: block; width: 100%; padding: 5px 7px; border: 0; border-radius: 5px; background: transparent; color: var(--sui-chart-fg, #e6e6ef); font-size: 11px; text-align: left; cursor: pointer; }
.sui-chart-font-option:hover, .sui-chart-font-option:focus { background: var(--sui-chart-series-1, #38bdf8); color: #fff; outline: none; }
.sui-chart-state-tag-preview-wrap { position: sticky; top: 0; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 10px; background-color: var(--sui-chart-bg, #14171c); background-image: linear-gradient(rgba(148,163,184,.10) 1px, transparent 1px), linear-gradient(90deg, rgba(148,163,184,.10) 1px, transparent 1px); background-size: 14px 14px; overflow: hidden; box-shadow: inset 0 1px 0 rgba(255,255,255,.04), 0 10px 28px rgba(2,6,23,.18); padding: 4px; }
.sui-chart-state-tag-preview { display: block; width: 100%; height: auto; border-radius: 8px; background: rgba(15,23,42,.06); touch-action: none; user-select: none; }
.sui-chart-state-tag-preview .sui-chart-state-tag-content { cursor: move; }
.sui-chart-state-tag-preview .sui-chart-state-tag-rotate { opacity: .9; pointer-events: auto; }
.sui-chart-state-tag-preview .sui-chart-state-tag-rotate:hover, .sui-chart-state-tag-preview .sui-chart-state-tag-rotate:focus { opacity: 1; fill: var(--sui-chart-state-selection, #f59e0b); outline: none; }
.sui-chart-state-tag-preview-hint { padding: 5px 6px 1px; text-align: center; font-size: 9.5px; line-height: 1.3; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-state-tag-preview-bg { fill: transparent; }
.sui-chart-state-tag-preview-label { position: absolute; left: 10px; top: 8px; padding: 2px 6px; border-radius: 999px; border: 1px solid var(--sui-chart-control-border, #475569); background: var(--sui-chart-control-bg, rgba(15,23,42,.72)); font-size: 9px; letter-spacing: .06em; text-transform: uppercase; color: var(--sui-chart-muted, #8fa0b8); pointer-events: none; }
.sui-chart-state-tag-settings-row { display: flex; align-items: center; gap: 10px; margin-top: 12px; flex-wrap: wrap; }
.sui-chart-state-tag-settings-row span { color: var(--sui-chart-muted, #8fa0b8); font-size: 12px; }
@media (max-width: 760px) { .sui-chart-state-tag-body { grid-template-columns: 1fr; } .sui-chart-state-tag-preview-wrap { position: relative; } .sui-chart-tag-editor { max-height: none; } }
.sui-chart-state-edge-handle { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 1.5; cursor: move; opacity: 0; pointer-events: none; transition: opacity .12s ease; }
.sui-chart-state-edge-handle.sui-chart-state-handle-visible { opacity: .85; pointer-events: auto; }
.sui-chart-state-edge-handle:hover, .sui-chart-state-edge-handle:focus { opacity: 1; }
.sui-chart-state-waypoint { fill: var(--sui-chart-bg, #14171c); stroke: var(--sui-chart-state-selection, #f59e0b); stroke-width: 1.5; cursor: move; opacity: 0; pointer-events: none; transition: opacity .12s ease; }
.sui-chart-state-waypoint.sui-chart-state-handle-visible { opacity: 1; pointer-events: auto; }
.sui-chart-state-waypoint:hover, .sui-chart-state-waypoint:focus { fill: var(--sui-chart-state-selection, #f59e0b); }
.sui-chart-state-table-tabs { display: flex; gap: 4px; align-items: center; margin: 0 0 6px; padding: 3px; width: 100%; box-sizing: border-box; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 7px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); }
.sui-chart-state-table-tablist { display: flex; gap: 4px; align-items: center; }
.sui-chart-state-table-tabs button { border: 1px solid transparent; border-radius: 4px; background: transparent; color: var(--sui-chart-muted, #8fa0b8); font: inherit; font-size: 11px; padding: 4px 9px; height: 28px; min-height: 28px; box-sizing: border-box; cursor: pointer; }
.sui-chart-state-table-tabs button:hover, .sui-chart-state-table-tabs button:focus-visible { border-color: var(--sui-chart-control-border, #475569); color: var(--sui-chart-fg, #e6e6ef); outline: none; }
.sui-chart-state-table-tabs button.active { border-color: var(--sui-chart-state-selection, #f59e0b); background: rgba(245,158,11,.14); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-state-tab-panel[hidden] { display: none; }
.sui-chart-table.sui-chart-state-table-scroll { display: flex; flex-direction: column; overflow: hidden; }
.sui-chart-state-table-scroll .sui-chart-state-table-tabs { flex: none; }
.sui-chart-state-table-scroll .sui-chart-state-tab-panel {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
}
.sui-chart-state-table-scroll .sui-chart-state-table th { z-index: 3; }
.sui-chart-state-table-tabs .sui-chart-state-add-row { margin-left: auto; border: 1px solid var(--sui-chart-control-border, #475569); background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-fg, #e6e6ef); border-radius: 4px; padding: 3px 7px; font: inherit; font-size: 11px; font-weight: 650; height: 28px; min-height: 28px; box-sizing: border-box; cursor: pointer; }
.sui-chart-state-table-tabs .sui-chart-state-add-row:hover, .sui-chart-state-table-tabs .sui-chart-state-add-row:focus-visible { border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); color: var(--sui-chart-fg, #e6e6ef); outline: none; }
.sui-chart-state-create-tools { margin-left: auto; display: flex; flex: none; align-items: center; position: static; padding: 0; border: 0; background: transparent; box-shadow: none; }
.sui-chart-state-create-tools { display: flex; align-items: center; gap: 3px; flex: none; margin-left: auto; }
.sui-chart-state-create-tools + .sui-chart-menu-btn { margin-left: 0; }
.sui-chart-state-create-tools button { width: 26px; height: 24px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, #475569)); border-radius: var(--sui-chart-control-radius, 5px); background: var(--sui-chart-control-bg, transparent); color: var(--sui-chart-control-fg, var(--sui-chart-fg, #e6e6ef)); font: inherit; font-size: 17px; line-height: 1; padding: 0; cursor: pointer; }
.sui-chart-state-create-tools .sui-chart-state-icon-btn { width: 28px; height: 28px; padding: 0; }
.sui-chart-state-create-tools .sui-chart-state-icon-btn:disabled { opacity: .4; cursor: default; }
.sui-chart-state-create-tools .sui-chart-state-icon-btn:disabled:hover { border-color: var(--sui-chart-control-border, var(--sui-chart-axis, #475569)); background: var(--sui-chart-control-bg, transparent); }
.sui-chart-state-create-tools button:hover, .sui-chart-state-create-tools button:focus-visible, .sui-chart-state-create-tools button[aria-expanded="true"] { border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #38bdf8)); background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); outline: none; }
.sui-chart-state-create-tools .sui-chart-state-add-btn { width: auto; min-width: 62px; padding: 3px 10px; font-size: 12px; font-weight: 600; gap: 6px; }
.sui-chart-state-create-section { padding: 5px 8px 2px; color: var(--sui-chart-muted, #8fa0b8); font-size: 9px; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; }
.sui-chart-table-only .sui-chart-state-create-tools { display: none; }
/* Placement mode preview */
.sui-chart-state-preview-shape { fill: none; stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2; stroke-dasharray: 6 4; pointer-events: none; opacity: .85; }
.sui-chart-state-preview-link { fill: none; stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2; stroke-dasharray: 6 4; pointer-events: none; opacity: .8; }
.sui-chart-state-preview-anchor { fill: var(--sui-chart-series-1, #38bdf8); stroke: #ffffff; stroke-width: 1.5; pointer-events: none; opacity: .9; }
.sui-chart-state-anchor-hover { fill: #ffffff; stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2; cursor: crosshair; }
.sui-chart-state-cursor-crosshair { cursor: crosshair !important; }
.sui-chart-state-panning .sui-chart-svg { cursor: grabbing !important; }
.sui-chart-state-tooltip-hint { position: absolute; background: var(--sui-chart-control-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; padding: 4px 8px; font-size: 11px; pointer-events: none; z-index: 100; white-space: nowrap; }
.sui-chart-state-actions-header, .sui-chart-state-actions { width: 88px !important; min-width: 88px !important; max-width: 88px !important; padding: 2px 3px !important; text-align: center; white-space: nowrap; }
.sui-chart-state-actions { text-align: center; vertical-align: middle; }
.sui-chart-state-actions .sui-chart-state-icon-btn + .sui-chart-state-icon-btn { margin-left: 2px; }
.sui-chart-state-actions .sui-chart-state-icon-btn { width: 24px; height: 24px; padding: 0; border-radius: 5px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid transparent; background: transparent; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-state-actions .sui-chart-state-icon-btn:hover, .sui-chart-state-actions .sui-chart-state-icon-btn:focus-visible { border-color: var(--sui-chart-control-border, #475569); background: var(--sui-chart-control-hover-bg, #2c3444); color: var(--sui-chart-fg, #e6e6ef); outline: none; }
.sui-chart-state-actions .sui-chart-state-delete:hover, .sui-chart-state-actions .sui-chart-state-delete:focus-visible { color: var(--sui-chart-error, #f87171); border-color: var(--sui-chart-error, #f87171); background: rgba(248,113,113,.12); }
.sui-chart-state-actions .sui-chart-state-delete svg { pointer-events: none; }
.sui-chart-table.sui-chart-state-table { border-collapse: separate; border-spacing: 0; table-layout: auto; overflow: visible; max-height: none; }
.sui-chart-state-table thead th { position: sticky; top: 0; z-index: 3; background: var(--sui-chart-table-sticky-bg, var(--sui-chart-table-head-bg, var(--sui-chart-bg, #14171c))); }
.sui-chart-state-table th { position: relative; white-space: nowrap; }
.sui-chart-state-col-resize { position: absolute; right: -4px; top: 0; bottom: 0; width: 8px; cursor: col-resize; z-index: 4; touch-action: none; }
.sui-chart-state-col-resize:focus-visible { outline: 2px solid var(--sui-chart-series-1, #38bdf8); outline-offset: 1px; }
.sui-chart-state-cell-input { box-sizing: border-box; width: 100%; min-width: 70px; border: 1px solid transparent; background: transparent; color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 11px; padding: 3px 4px; border-radius: 3px; }
.sui-chart-link-fields textarea { box-sizing: border-box; padding: 4px 7px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; background: var(--sui-chart-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 11px; }
.sui-chart-state-cell-input:focus { border-color: var(--sui-chart-series-1, #38bdf8); outline: none; background: var(--sui-chart-bg, #14171c); }
.sui-chart-state-row td .sui-chart-state-color + .sui-chart-state-cell-input { width: calc(100% - 25px); }
.sui-chart-state-row { cursor: pointer; }
.sui-chart-state-row-selected td { background: var(--sui-chart-table-selection-bg, var(--sui-chart-table-selected-row-bg, rgba(245,158,11,.18))); box-shadow: inset 3px 0 0 var(--sui-chart-state-selection, #f59e0b); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-state-row button { border: 0; background: transparent; color: var(--sui-chart-muted, #8fa0b8); cursor: pointer; font-size: 15px; }
.sui-chart-gantt-progress-sep { stroke: var(--sui-chart-grid, #2c3444); stroke-width: .5; }
.sui-chart-gantt-wbs-summary-item .sui-chart-gantt-summary { stroke-width: 2; }
.sui-chart-gantt-item { touch-action: none; }
.sui-chart-gantt-item:focus, .sui-chart-gantt-link-group:focus { outline: none; }
.sui-chart-gantt-item:focus-visible .sui-chart-gantt-task,
.sui-chart-gantt-item:focus-visible .sui-chart-gantt-summary,
.sui-chart-gantt-item:focus-visible .sui-chart-gantt-milestone { stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2.5; }
.sui-chart-gantt-link-handle:focus-visible { opacity: 1; stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2; }
.sui-chart-gantt-link-group:focus-visible .sui-chart-gantt-dependency { stroke-width: 3 !important; }
.sui-chart-gantt-live { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
.sui-chart-gantt-handle { fill: var(--sui-chart-fg, #e6e6ef); stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; opacity: 0; pointer-events: none; transition: opacity .12s; cursor: ew-resize; }
.sui-chart-gantt-link-handle { fill: var(--sui-chart-series-1, #38bdf8); stroke: var(--sui-chart-bg, #14171c); stroke-width: 1; opacity: .42; cursor: crosshair; transition: opacity .12s; }
.sui-chart-gantt-item:hover .sui-chart-gantt-link-handle, .sui-chart-gantt-item:focus-within .sui-chart-gantt-link-handle { opacity: 1; }
.sui-chart-gantt-link-target .sui-chart-gantt-task, .sui-chart-gantt-link-target .sui-chart-gantt-summary, .sui-chart-gantt-link-target .sui-chart-gantt-milestone { stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 2.5; filter: drop-shadow(0 0 3px var(--sui-chart-series-1, #38bdf8)); }
.sui-chart-gantt-link-preview { fill: none; stroke: var(--sui-chart-series-1, #38bdf8); stroke-width: 1.5; stroke-dasharray: 5 4; pointer-events: none; }
.sui-chart-interactive .sui-chart-gantt-handle { opacity: .45; pointer-events: all; }
.sui-chart-interactive .sui-chart-gantt-item:hover .sui-chart-gantt-handle,
.sui-chart-interactive .sui-chart-gantt-item.sui-chart-gantt-dragging .sui-chart-gantt-handle { opacity: .95; }
.sui-chart-gantt-dragging { user-select: none; }
.sui-chart-table { position: relative; flex: 0 1 auto; min-height: 0; max-height: 45%; overflow: auto; margin-top: 10px; border-top: 1px solid var(--sui-chart-table-row-border, var(--sui-chart-grid, #2c3444)); background: var(--sui-chart-table-bg, transparent); }
.sui-chart-table-selection-outline { position: absolute; z-index: 2; pointer-events: none; box-sizing: border-box; border: 2px solid var(--sui-chart-table-selection-border, var(--sui-chart-series-1, #38bdf8)); border-radius: 2px; }
.sui-chart-table[hidden] { display: none; }
.sui-chart-table table { width: 100%; border-collapse: collapse; color: var(--sui-chart-table-fg, var(--sui-chart-fg, #e6e6ef)); font-family: var(--sui-chart-table-font-family, inherit); font-size: var(--sui-chart-table-font-size, 11px); }
.sui-chart-table caption { display: var(--sui-chart-table-caption-display, block); caption-side: top; text-align: left; padding: 6px 4px 4px; color: var(--sui-chart-muted, #8fa0b8); font-weight: 600; }
.sui-chart-table th, .sui-chart-table td { padding: var(--sui-chart-table-cell-padding, 4px 6px); text-align: left; border-bottom: 1px solid var(--sui-chart-grid, #2c3444); white-space: nowrap; }
.sui-chart-table td[data-col="0"], .sui-chart-table th[data-col="0"] { white-space: normal; overflow-wrap: anywhere; }
.sui-chart-gantt-table td[data-col="0"], .sui-chart-gantt-table th[data-col="0"] { min-width: 190px; }
.sui-chart-table.sui-chart-gantt-table th { top: var(--sui-chart-tabs-height, 0px); }
.sui-chart-table th { position: var(--sui-chart-table-header-position, sticky); top: 0; background: var(--sui-chart-table-head-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-table-head-fg, var(--sui-chart-title-color, var(--sui-chart-fg, #e6e6ef))); font-weight: var(--sui-chart-table-head-weight, 600); z-index: 1; }
.sui-chart-table tbody tr:nth-child(even) td { background: rgba(127,127,127,var(--sui-chart-table-stripe-opacity, .06)); }
.sui-chart-table tbody tr.sui-chart-table-row-selected td,
.sui-chart-table tbody tr.sui-chart-table-row-multi td,
.sui-chart-table th.sui-chart-table-col-multi,
.sui-chart-table td.sui-chart-table-col-multi,
.sui-chart-table td.sui-chart-table-cell-multi { background: var(--sui-chart-table-selection-bg, var(--sui-chart-table-selected-row-bg, var(--sui-chart-table-selected-bg, rgba(56,189,248,.22)))) !important; }
.sui-chart-table-stepper { display: inline-flex; align-items: center; justify-content: flex-end; gap: 2px; width: 100%; }
.sui-chart-table-stepper .sui-chart-table-input { width: auto; flex: 1 1 auto; min-width: 0; }
.sui-chart-table-step-btn { flex: none; width: 18px; height: 18px; padding: 0; border: 1px solid var(--sui-chart-axis, #475569); border-radius: 3px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-fg, #e6e6ef); font: inherit; line-height: 16px; cursor: pointer; opacity: 0; pointer-events: none; transition: opacity .12s ease; }
.sui-chart-table tbody td:focus-within .sui-chart-table-step-btn,
.sui-chart-table-step-btn:focus { opacity: 1; pointer-events: auto; }
.sui-chart-table.sui-chart-table-dragging .sui-chart-table-step-btn { opacity: 0 !important; pointer-events: none !important; }
.sui-chart-table-step-btn:hover { background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); }
.sui-chart-table-first-cell { position: relative; padding-left: 22px !important; }
.sui-chart-table-row-select { position: absolute; left: 1px; top: 1px; bottom: 1px; width: 17px; display: inline-flex; align-items: center; justify-content: center; border-right: 1px solid var(--sui-chart-grid, #2c3444); color: var(--sui-chart-muted, #8fa0b8); font-size: 9px; font-weight: 400; cursor: pointer; user-select: none; }
.sui-chart-table-row-select:hover { color: var(--sui-chart-fg, #e6e6ef); background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); }
.sui-chart-table-col-select { position: absolute; left: 0; right: 0; top: 0; height: 6px; cursor: col-resize; user-select: none; }
.sui-chart-table-col-select:hover { background: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); }
.sui-chart-table-row-delete { margin-left: 3px; width: 18px; height: 18px; padding: 0; border: 1px solid transparent; border-radius: 3px; background: transparent; color: var(--sui-chart-muted, #8fa0b8); font: inherit; line-height: 16px; cursor: pointer; }
.sui-chart-table-row-delete:hover { color: var(--sui-chart-error, #f87171); border-color: var(--sui-chart-error, #f87171); }
.sui-chart-table-action-cell { white-space: nowrap; }
.sui-chart-table-context { position: fixed; z-index: 20; min-width: 190px; padding: 4px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); border: 1px solid var(--sui-chart-axis, #475569); border-radius: 5px; box-shadow: 0 8px 20px rgba(0,0,0,.35); }
.sui-chart-table-context[hidden] { display: none; }
.sui-chart-table-input[type="number"] { appearance: textfield; -moz-appearance: textfield; }
.sui-chart-table-input[type="number"]::-webkit-inner-spin-button,
.sui-chart-table-input[type="number"]::-webkit-outer-spin-button { -webkit-appearance: none; margin: 0; }
.sui-chart-table-virtual tbody tr { height: 30px; }
.sui-chart-table-spacer td { padding: 0 !important; border: 0 !important; }
.sui-chart-table tbody tr.sui-chart-table-row-flash td { animation: sui-chart-table-flash .9s ease-out; }
.sui-chart-table tbody td.sui-chart-table-cell-flash { animation: sui-chart-cell-flash .9s ease-out; }
.sui-chart-table,
.sui-chart-table tr,
.sui-chart-table th,
.sui-chart-table td { -webkit-tap-highlight-color: transparent; }
.sui-chart-table ::selection { background: rgba(56,189,248,.28); color: inherit; }
@keyframes sui-chart-table-flash { 0% { background: var(--sui-chart-table-flash-bg, rgba(76,92,247,.42)); } 100% { background: transparent; } }
@keyframes sui-chart-cell-flash { 0% { box-shadow: inset 0 0 0 2px var(--sui-chart-table-selection-border, var(--sui-chart-series-1, #4c5cf7)); } 100% { box-shadow: inset 0 0 0 0 transparent; } }
.sui-chart-table td:first-child { font-weight: 600; }
.sui-chart-table-tabs { position: sticky; top: 0; left: 0; z-index: 7; display: flex; flex-wrap: wrap; align-items: center; gap: 3px; padding: 6px 4px 4px; flex: none; background: var(--sui-chart-bg, #14171c); border-bottom: 1px solid var(--sui-chart-grid, #2c3444); }
.sui-chart-table-tab { padding: 4px 10px; border: 1px solid transparent; border-bottom: 0; border-radius: 5px 5px 0 0; background: transparent; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; font-weight: 600; cursor: pointer; }
.sui-chart-table-tab:hover { background: var(--sui-chart-control-hover-bg, rgba(127,127,127,.12)); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-table-tab[aria-selected="true"] { border-color: var(--sui-chart-control-border, var(--sui-chart-axis, #475569)); background: var(--sui-chart-table-selection-bg, rgba(56,189,248,.22)); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-table-tab:focus-visible { outline: 2px solid var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); outline-offset: -2px; }
.sui-chart-table-tools { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; padding: 4px; flex: none; }
.sui-chart-table-hint { color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; }
.sui-chart-table-empty { padding: 14px 8px !important; color: var(--sui-chart-muted, #8fa0b8); font-style: italic; }
.sui-chart-resource-table { width: 100%; }
.sui-chart-resource-table td[data-col="2"] { min-width: 100px; }
.sui-chart-resource-table td[data-col="3"] { min-width: 130px; }
.sui-chart-resource-table td[data-col="4"] { min-width: 180px; }
.sui-chart-resource-table td[data-col="5"] { width: 32px; text-align: center !important; }
.sui-chart-resource-table .sui-chart-table-input { min-width: 0; }
.sui-chart-resource-delete { width: 22px; height: 22px; padding: 0; border: 1px solid transparent; border-radius: 4px; background: transparent; color: var(--sui-chart-error, #f87171); font-size: 14px; line-height: 1; cursor: pointer; }
.sui-chart-resource-delete:hover, .sui-chart-resource-delete:focus-visible { border-color: var(--sui-chart-error, #f87171); outline: none; }
.sui-chart-calendar-dialog.sui-chart-resource-picker-dialog { width: min(380px, 100%); }
.sui-chart-calendar-dialog.sui-chart-group-dialog { width: min(400px, 100%); }
.sui-chart-resource-picker-fields { display: grid; gap: 10px; margin: 14px 0 10px; }
.sui-chart-resource-picker-fields label { display: grid; gap: 4px; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; font-weight: 600; }
.sui-chart-resource-picker-select, .sui-chart-resource-picker-qty { width: 100%; box-sizing: border-box; padding: 6px 7px; border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, #475569)); border-radius: 4px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 12px; }
.sui-chart-resource-picker-select option, .sui-chart-calendar-dialog option, .sui-chart-link-fields option { background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-resource-picker-qty { width: 110px; }
.sui-chart-resource-picker-error { margin: 8px 0 0; color: var(--sui-chart-error, #f87171); font-size: 11px; }
.sui-chart-resource-picker-actions { display: flex; justify-content: flex-end; gap: 7px; margin-top: 14px; }
.sui-chart-resource-picker-actions button { padding: 5px 12px; border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, #475569)); border-radius: 4px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 12px; cursor: pointer; }
.sui-chart-resource-picker-actions button:hover:not(:disabled) { border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); }
.sui-chart-resource-picker-actions button.primary { border-color: var(--sui-chart-series-1, #4c5cf7); background: var(--sui-chart-series-1, #4c5cf7); color: #fff; }
.sui-chart-resource-picker-actions button:disabled { opacity: .55; cursor: default; }
.sui-chart-table-filter { min-width: 120px; flex: 1 1 150px; max-width: 260px; box-sizing: border-box; padding: 3px 6px; border: 1px solid var(--sui-chart-axis, #475569); border-radius: 4px; background: var(--sui-chart-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); font-size: 11px; }
.sui-chart-table-btn { padding: 3px 7px; border: 1px solid var(--sui-chart-control-border, var(--sui-chart-axis, #475569)); border-radius: 4px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-control-fg, var(--sui-chart-fg, #e6e6ef)); font-size: 11px; cursor: pointer; }
.sui-chart-table-settings-btn { width: 28px; height: 24px; padding: 0; font-size: 14px; line-height: 22px; text-align: center; }
.sui-chart-table-btn:hover { background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); }
.sui-chart-table-cols-wrap { position: relative; }
.sui-chart-table-cols { position: absolute; top: 100%; left: 0; z-index: 3; min-width: 150px; padding: 6px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); border: 1px solid var(--sui-chart-axis, #475569); border-radius: 5px; box-shadow: 0 6px 16px rgba(0,0,0,.3); }
.sui-chart-table-cols[hidden] { display: none; }
.sui-chart-table-cols label { display: flex; align-items: center; gap: 6px; padding: 2px 3px; font-size: 11px; color: var(--sui-chart-fg, #e6e6ef); white-space: nowrap; }
.sui-chart-table-submenu-wrap + .sui-chart-table-submenu-wrap { border-top: 1px solid var(--sui-chart-grid, #2c3444); margin-top: 3px; padding-top: 3px; }
.sui-chart-table-submenu-trigger { display: flex; justify-content: space-between; gap: 10px; font-weight: 600; }
.sui-chart-table-submenu { margin: 2px 0 3px 8px; padding-left: 7px; border-left: 1px solid var(--sui-chart-axis, #475569); }
.sui-chart-table-submenu[hidden] { display: none; }
.sui-chart-table-menu-title { padding: 5px 3px 2px; color: var(--sui-chart-muted, #8fa0b8); font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: .05em; }
.sui-chart-table-menu-action { display: block; width: 100%; padding: 4px 6px; border: 0; border-radius: 3px; background: transparent; color: var(--sui-chart-fg, #e6e6ef); font-size: 11px; text-align: left; cursor: pointer; }
.sui-chart-table-menu-action:hover { background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); }
.sui-chart-table-head-label { display: inline-block; margin-right: 4px; }
.sui-chart-table-col-resize { position: absolute; top: 0; right: -4px; z-index: 5; width: 8px; height: 100%; cursor: col-resize; touch-action: none; }
.sui-chart-table-col-resize::after { content: ""; position: absolute; top: 18%; bottom: 18%; left: 3px; width: 1px; background: var(--sui-chart-axis, #475569); opacity: .5; }
.sui-chart-table-col-resize:hover::after, .sui-chart-table-col-resize:focus-visible::after { width: 2px; left: 2px; opacity: 1; background: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); }
.sui-chart-table-col-resize:focus-visible { outline: 2px solid var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); outline-offset: -2px; }
.sui-chart-table-filter-btn { display: inline-flex; align-items: center; justify-content: center; width: 18px; height: 17px; padding: 0; border: 1px solid transparent; border-radius: 3px; background: transparent; color: var(--sui-chart-muted, #8fa0b8); font: inherit; font-size: 10px; line-height: 1; cursor: pointer; }
.sui-chart-table-filter-btn:hover, .sui-chart-table-filter-btn.active { background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); border-color: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart-table-filter-pop { position: absolute; top: 100%; left: 0; z-index: 4; width: 230px; max-height: 360px; overflow: auto; padding: 5px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); border: 1px solid var(--sui-chart-axis, #475569); border-radius: 4px; box-shadow: 0 6px 16px rgba(0,0,0,.3); }
.sui-chart-table-filter-pop[hidden] { display: none; }
.sui-chart-table-col-filter { display: block; width: 100%; box-sizing: border-box; padding: 2px 4px; border: 1px solid var(--sui-chart-axis, #475569); border-radius: 3px; background: var(--sui-chart-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); font-size: 10px; font-weight: 400; }
.sui-chart-table-filter-range { display: flex; gap: 4px; margin-top: 4px; }
.sui-chart-table-filter-values { margin-top: 5px; padding-top: 4px; border-top: 1px solid var(--sui-chart-grid, #2c3444); max-height: 150px; overflow: auto; }
.sui-chart-table-filter-title { padding: 1px 2px 3px; color: var(--sui-chart-muted, #8fa0b8); font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: .05em; }
.sui-chart-table-filter-values label { display: flex; align-items: center; gap: 5px; padding: 1px 2px; color: var(--sui-chart-fg, #e6e6ef); font-size: 10px; font-weight: 400; white-space: nowrap; }
.sui-chart-table-add-row { font-weight: 600; }
.sui-chart-table-status { margin-left: auto; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; }
.sui-chart-table th[data-col="0"], .sui-chart-table td[data-col="0"] { position: sticky; left: 0; z-index: 1; background: var(--sui-chart-table-sticky-bg, var(--sui-chart-bg, #14171c)); }
.sui-chart-table th[data-col="0"] { z-index: 2; }
.sui-chart-table tbody tr.sui-chart-table-row-selected td[data-col="0"],
.sui-chart-table tbody tr.sui-chart-table-row-multi td[data-col="0"],
.sui-chart-table th[data-col="0"].sui-chart-table-col-multi,
.sui-chart-table td[data-col="0"].sui-chart-table-col-multi,
.sui-chart-table td[data-col="0"].sui-chart-table-cell-multi { background: var(--sui-chart-table-selection-bg, var(--sui-chart-table-selected-row-bg, var(--sui-chart-table-selected-bg, rgba(56,189,248,.22)))) !important; }
.sui-chart-table table.sui-chart-table-mode-cell tbody tr.sui-chart-table-row-selected td { background: transparent !important; }
.sui-chart-table table.sui-chart-table-mode-cell tbody tr.sui-chart-table-row-selected td.sui-chart-table-cell-multi { background: var(--sui-chart-table-selection-bg, var(--sui-chart-table-selected-row-bg, var(--sui-chart-table-selected-bg, rgba(56,189,248,.22)))) !important; }
.sui-chart-table th { cursor: pointer; }
.sui-chart-table th:focus-visible, .sui-chart-table tbody tr:focus-visible,
.sui-chart-table tbody td:focus-visible,
.sui-chart-table tbody td:has(:focus-visible) { outline: 1px solid var(--sui-chart-table-focus-color, #cbd5e1); outline-offset: -1px; }
.sui-chart-table:not(.sui-chart-table-dragging) tbody tr:not(.sui-chart-table-row-selected):not(.sui-chart-table-row-multi):hover td:not(.sui-chart-table-cell-multi):not(.sui-chart-table-col-multi) { background: var(--sui-chart-table-focus-row-bg, rgba(255,255,255,.055)) !important; }
.sui-chart-table:not(.sui-chart-table-dragging) tbody tr:not(.sui-chart-table-row-selected):not(.sui-chart-table-row-multi) td:not(.sui-chart-table-cell-multi):not(.sui-chart-table-col-multi):hover { background: var(--sui-chart-table-focus-cell-bg, rgba(255,255,255,.10)) !important; outline: 1px solid var(--sui-chart-table-focus-color, #cbd5e1); outline-offset: -1px; }
.sui-chart-table-step-btn:focus-visible { outline: 2px solid var(--sui-chart-table-focus-color, #cbd5e1); outline-offset: 1px; }
.sui-chart-table th[aria-sort="ascending"]::after { content: " ▲" attr(data-sort-rank); font-size: 9px; }
.sui-chart-table th[aria-sort="descending"]::after { content: " ▼" attr(data-sort-rank); font-size: 9px; }
.sui-chart-table th[draggable="true"] { cursor: grab; }
.sui-chart-table th[draggable="true"]:active { cursor: grabbing; }
.sui-chart-table-input { width: 100%; min-width: 64px; box-sizing: border-box; padding: 1px 3px; border: 1px solid transparent; border-radius: 3px; background: transparent; color: inherit; font: inherit; }
.sui-chart-table-input option, .sui-chart-gantt-dep-add option { background: var(--sui-chart-table-input-bg, #1b2230); color: var(--sui-chart-fg, #e6e6ef); }
.sui-chart[data-sui-chart-theme="light"] .sui-chart-table-input option,
.sui-chart[data-sui-chart-theme="light"] .sui-chart-gantt-dep-add option { background: #ffffff; color: #172033; }
.sui-chart-table-input:hover { border-color: var(--sui-chart-axis, #475569); }
.sui-chart-table-input:focus { outline: none; border-color: var(--sui-chart-table-focus-color, #cbd5e1); background: var(--sui-chart-table-input-bg, rgba(127,127,127,.08)); }
.sui-chart-table-input[aria-invalid="true"] { border-color: var(--sui-chart-error, #f87171); }
.sui-chart-table td:has(.sui-chart-table-input) { padding-top: 2px; padding-bottom: 2px; }
.sui-chart-table-only .sui-chart-viewport,
.sui-chart-table-only .sui-chart-zoom,
.sui-chart-table-only .sui-chart-legend { display: none; }
.sui-chart-table-only .sui-chart-table { flex: 1 1 auto; max-height: 70vh; overflow: auto; margin-top: 0; border-top: 0; }
.sui-chart-table-both { overflow: hidden; }
.sui-chart-table-both .sui-chart-viewport { flex: 1 1 auto; height: auto; min-height: 0; }
.sui-chart-table-both > .sui-chart-table { flex: 0 0 var(--sui-chart-table-height, 180px); height: var(--sui-chart-table-height, 180px); max-height: calc(100% - var(--sui-chart-head-height, 38px) - 80px); overflow: auto; margin-top: 0; width: 100%; }
.sui-chart-table-both > .sui-chart-table.sui-chart-state-table-scroll { overflow: hidden; }
.sui-chart-table-both > .sui-chart-table:not(.sui-chart-state-table-scroll) > .sui-chart-table-tools,
.sui-chart-table-both > .sui-chart-table:not(.sui-chart-state-table-scroll) > .sui-chart-gantt-group-toggles { position: sticky; z-index: 6; background: var(--sui-chart-table-head-bg, var(--sui-chart-bg, #14171c)); }
.sui-chart-table-both > .sui-chart-table:not(.sui-chart-state-table-scroll) > .sui-chart-table-tools { top: var(--sui-chart-table-tools-top, 0px); }
.sui-chart-table-both > .sui-chart-table:not(.sui-chart-state-table-scroll) > .sui-chart-gantt-group-toggles { top: var(--sui-chart-table-group-top, 0px); }
.sui-chart-table-both > .sui-chart-table:not(.sui-chart-state-table-scroll) th { top: var(--sui-chart-table-head-top, 0px); }
.sui-chart-table-both .sui-chart-legend { flex: none; }
.sui-chart-splitter { display: none; flex: none; height: 9px; margin-top: 4px; cursor: row-resize; align-items: center; justify-content: center; touch-action: none; }
.sui-chart-splitter::before { content: ""; display: block; width: 44px; height: 2px; border-radius: 2px; background: var(--sui-chart-axis, #475569); opacity: .65; }
.sui-chart-splitter:hover::before, .sui-chart-splitter.sui-chart-splitter-active::before { background: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); opacity: 1; }
.sui-chart-splitter:focus-visible { outline: 2px solid var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); outline-offset: 2px; border-radius: 3px; }
.sui-chart-splitter:focus-visible::before { background: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); opacity: 1; }
.sui-chart-table-both .sui-chart-splitter { display: flex; }
.sui-chart-slice { stroke: var(--sui-chart-bg, #14171c); stroke-width: var(--sui-chart-slice-stroke, 2); }
.sui-chart-pie-hole { fill: var(--sui-chart-bg, #14171c); }
.sui-chart-series-1 { fill: var(--sui-chart-series-1, #4c5cf7); }
.sui-chart-series-2 { fill: var(--sui-chart-series-2, #22c55e); }
.sui-chart-series-3 { fill: var(--sui-chart-series-3, #f59e0b); }
.sui-chart-series-4 { fill: var(--sui-chart-series-4, #ef4444); }
.sui-chart-series-5 { fill: var(--sui-chart-series-5, #06b6d4); }
.sui-chart-series-6 { fill: var(--sui-chart-series-6, #a855f7); }
.sui-chart-series-7 { fill: var(--sui-chart-series-7, #84cc16); }
.sui-chart-series-8 { fill: var(--sui-chart-series-8, #ec4899); }
.sui-chart-line.sui-chart-series-1 { stroke: var(--sui-chart-series-1, #4c5cf7); }
.sui-chart-line.sui-chart-series-2 { stroke: var(--sui-chart-series-2, #22c55e); }
.sui-chart-line.sui-chart-series-3 { stroke: var(--sui-chart-series-3, #f59e0b); }
.sui-chart-line.sui-chart-series-4 { stroke: var(--sui-chart-series-4, #ef4444); }
.sui-chart-line.sui-chart-series-5 { stroke: var(--sui-chart-series-5, #06b6d4); }
.sui-chart-line.sui-chart-series-6 { stroke: var(--sui-chart-series-6, #a855f7); }
.sui-chart-line.sui-chart-series-7 { stroke: var(--sui-chart-series-7, #84cc16); }
.sui-chart-line.sui-chart-series-8 { stroke: var(--sui-chart-series-8, #ec4899); }
.sui-chart-legend { display: var(--sui-chart-legend-display, flex); flex-wrap: wrap; gap: var(--sui-chart-legend-gap, 4px 12px); flex: none; max-height: 24%; overflow: auto; padding: var(--sui-chart-legend-padding, 4px 4px 0); font-size: var(--sui-chart-legend-size, 11px); color: var(--sui-chart-legend-color, var(--sui-chart-muted, #8fa0b8)); background: var(--sui-chart-legend-bg, transparent); }
.sui-chart-legend[hidden] { display: none; }
.sui-chart-legend-item { display: inline-flex; align-items: center; gap: 5px; min-width: 0; }
.sui-chart-legend-item > span:last-child { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sui-chart-legend-colorizable { cursor: pointer; border-radius: 3px; }
.sui-chart-legend-colorizable:hover { background: var(--sui-chart-control-hover-bg, var(--sui-chart-grid, #2c3444)); }
.sui-chart-swatch { display: inline-block; width: 10px; height: 10px; border-radius: 2px; flex: none; border: 1px solid var(--sui-chart-axis, #475569); }
.sui-chart-swatch.sui-chart-series-1 { background: var(--sui-chart-series-1, #4c5cf7); }
.sui-chart-swatch.sui-chart-series-2 { background: var(--sui-chart-series-2, #22c55e); }
.sui-chart-swatch.sui-chart-series-3 { background: var(--sui-chart-series-3, #f59e0b); }
.sui-chart-swatch.sui-chart-series-4 { background: var(--sui-chart-series-4, #ef4444); }
.sui-chart-swatch.sui-chart-series-5 { background: var(--sui-chart-series-5, #06b6d4); }
.sui-chart-swatch.sui-chart-series-6 { background: var(--sui-chart-series-6, #a855f7); }
.sui-chart-swatch.sui-chart-series-7 { background: var(--sui-chart-series-7, #84cc16); }
.sui-chart-swatch.sui-chart-series-8 { background: var(--sui-chart-series-8, #ec4899); }
.sui-chart-table-series-color { display: inline-block; width: 12px; height: 12px; margin: 0 4px 0 0; padding: 0; vertical-align: -1px; border: 1px solid var(--sui-chart-axis, #475569); border-radius: 3px; cursor: pointer; }
.sui-chart-table-series-color:hover { outline: 2px solid var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); outline-offset: 1px; }
.sui-chart-gantt-label { display: flex; align-items: center; gap: 4px; min-width: 0; }
.sui-chart-gantt-label .sui-chart-table-input { flex: 1 1 auto; min-width: 0; }
.sui-chart-gantt-label textarea.sui-chart-table-input { resize: none; overflow: hidden; line-height: 1.25; white-space: pre-wrap; overflow-wrap: anywhere; }
.sui-chart-gantt-task-color { display: inline-block; flex: none; width: 12px; height: 12px; padding: 0; vertical-align: -1px; border: 1px solid var(--sui-chart-axis, #475569); border-radius: 3px; cursor: pointer; }
.sui-chart-gantt-task-color:hover { outline: 2px solid var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)); outline-offset: 1px; }
.sui-chart-error { flex: none; font-size: 12px; color: var(--sui-chart-error, #f87171); padding: 4px; }
.sui-chart-error[hidden] { display: none; }
.sui-chart-tooltip {
  position: absolute; z-index: 5; pointer-events: none;
  max-width: var(--sui-chart-tooltip-max-width, 240px);
  background: var(--sui-chart-tooltip-bg, #111827);
  color: var(--sui-chart-tooltip-fg, #f8fafc);
  border: var(--sui-chart-tooltip-border-width, 1px) solid var(--sui-chart-tooltip-border, #475569);
  border-radius: var(--sui-chart-tooltip-radius, 6px);
  padding: var(--sui-chart-tooltip-padding, 6px 8px);
  font-family: var(--sui-chart-tooltip-font-family, system-ui, sans-serif);
  font-size: var(--sui-chart-tooltip-font-size, 11px);
  line-height: 1.4; white-space: pre-line;
  box-shadow: var(--sui-chart-tooltip-shadow, 0 6px 18px rgba(0,0,0,.35));
}
.sui-chart-tooltip[hidden] { display: none; }
.sui-chart-table-cell-tip { position: fixed; z-index: 60; pointer-events: none; max-width: 320px; background: var(--sui-chart-tooltip-bg, #111827); color: var(--sui-chart-tooltip-fg, #f8fafc); border: var(--sui-chart-tooltip-border-width, 1px) solid var(--sui-chart-tooltip-border, #475569); border-radius: var(--sui-chart-tooltip-radius, 6px); padding: var(--sui-chart-tooltip-padding, 6px 8px); font-family: var(--sui-chart-tooltip-font-family, system-ui, sans-serif); font-size: var(--sui-chart-tooltip-font-size, 11px); line-height: 1.4; white-space: pre-line; box-shadow: var(--sui-chart-tooltip-shadow, 0 6px 18px rgba(0,0,0,.35)); }
.sui-chart-table-cell-tip[hidden] { display: none; }
.sui-chart-hint-preview { grid-column: 1 / -1; margin-top: 10px; padding: 14px; border: 1px dashed var(--sui-chart-control-border, #475569); border-radius: 8px; background: var(--sui-chart-control-bg, #1d2129); text-align: center; }
.sui-chart-hint-preview-label { font-size: 11px; color: var(--sui-chart-muted, #8fa0b8); margin-bottom: 8px; }
.sui-chart-hint-preview-box { display: inline-block; max-width: 240px; white-space: pre-line; line-height: 1.4; margin-left: auto; margin-right: auto; }
.sui-chart-menu { position: fixed; z-index: 50; min-width: 170px; padding: 4px; background: var(--sui-chart-control-bg, #1d2129); border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 6px; box-shadow: 0 8px 24px rgba(0,0,0,.4); }
.sui-chart-calendar-overlay { position: fixed; inset: 0; z-index: 10020; display: flex; align-items: center; justify-content: center; padding: 16px; background: rgba(0,0,0,.58); }
.sui-chart-calendar-dialog { width: min(560px, 100%); max-height: min(720px, 92vh); overflow: auto; box-sizing: border-box; padding: 16px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 8px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-fg, #e6e6ef); box-shadow: 0 18px 50px rgba(0,0,0,.5); }
.sui-chart-calendar-dialog h3 { margin: 0 0 5px; font-size: 16px; color: var(--sui-chart-title-color, var(--sui-chart-fg, #e6e6ef)); }
.sui-chart-calendar-hint { margin: 0 0 12px; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; line-height: 1.4; }
.sui-chart-workload-dialog { width: min(1180px, 100%); }
.sui-chart-workload-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.sui-chart-workload-head h3 { margin: 0; }
.sui-chart-workload-close { flex: none; width: 26px; height: 26px; padding: 0; border: 1px solid transparent; border-radius: 4px; background: transparent; color: inherit; font-size: 18px; line-height: 1; cursor: pointer; }
.sui-chart-workload-close:hover, .sui-chart-workload-close:focus-visible { border-color: var(--sui-chart-control-hover-border, #4c5cf7); outline: none; }
.sui-chart-workload-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.sui-chart-workload-table th, .sui-chart-workload-table td { padding: 6px 7px; border-bottom: 1px solid var(--sui-chart-grid, #2c3444); text-align: left; vertical-align: middle; }
.sui-chart-workload-table th { background: var(--sui-chart-table-head-bg, var(--sui-chart-bg, #14171c)); color: var(--sui-chart-muted, #8fa0b8); font-size: 10px; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; }
.sui-chart-workload-table tbody tr:hover td { background: var(--sui-chart-control-hover-bg, rgba(127,127,127,.10)); }
.sui-chart-workload-table tfoot td { border-bottom: 0; border-top: 1px solid var(--sui-chart-axis, #475569); font-weight: 700; }
.sui-chart-workload-number { text-align: right !important; font-variant-numeric: tabular-nums; }
.sui-chart-workload-role { color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-workload-bar-cell { min-width: 220px; }
.sui-chart-workload-bar { display: inline-block; width: calc(100% - 52px); height: 10px; overflow: hidden; vertical-align: middle; border-radius: 3px; background: var(--sui-chart-grid, #2c3444); }
.sui-chart-workload-fill { display: block; height: 100%; border-radius: 3px; background: var(--sui-chart-series-1, #38bdf8); }
.sui-chart-workload-value { display: inline-block; width: 44px; margin-left: 8px; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; text-align: right; font-variant-numeric: tabular-nums; }
.sui-chart-calendar-weekend { display: flex; flex-wrap: wrap; gap: 7px 12px; margin-bottom: 12px; font-size: 11px; }
.sui-chart-calendar-weekend label { display: inline-flex; align-items: center; gap: 4px; }
.sui-chart-calendar-section { margin: 10px 0 5px; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: .05em; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-calendar-list { max-height: 150px; overflow: auto; border: 1px solid var(--sui-chart-grid, #2c3444); border-radius: 4px; padding: 4px; }
.sui-chart-calendar-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 3px 5px; font-size: 11px; border-bottom: 1px solid var(--sui-chart-grid, #2c3444); }
.sui-chart-calendar-row:last-child { border-bottom: 0; }
.sui-chart-calendar-row button { border: 0; background: transparent; color: var(--sui-chart-error, #f87171); cursor: pointer; font-size: 14px; }
.sui-chart-calendar-empty { padding: 6px; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; }
.sui-chart-calendar-add { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; margin-top: 7px; }
.sui-chart-datepicker { position: relative; }
.sui-chart-datepicker-trigger { min-width: 210px; text-align: left; }
.sui-chart-datepicker-trigger[aria-expanded="true"] { border-color: var(--sui-chart-control-hover-border, #4c5cf7); }
.sui-chart-datepicker-pop { position: absolute; z-index: 2; top: calc(100% + 4px); left: 0; width: 238px; padding: 6px; background: var(--sui-chart-control-bg, var(--sui-chart-bg, #14171c)); border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 5px; box-shadow: 0 8px 24px rgba(0,0,0,.4); }
.sui-chart-datepicker-pop[hidden] { display: none; }
.sui-chart-datepicker-head { display: flex; align-items: center; justify-content: space-between; gap: 5px; margin-bottom: 4px; font-size: 11px; }
.sui-chart-datepicker-head button { width: 24px; padding: 2px; border: 1px solid transparent; border-radius: 3px; background: transparent; color: var(--sui-chart-fg, #e6e6ef); cursor: pointer; }
.sui-chart-datepicker-head button:hover { border-color: var(--sui-chart-control-hover-border, #4c5cf7); }
.sui-chart-datepicker-weekdays, .sui-chart-datepicker-grid { display: grid; grid-template-columns: repeat(7, 1fr); gap: 2px; }
.sui-chart-datepicker-weekdays span { padding: 2px 0; color: var(--sui-chart-muted, #8fa0b8); font-size: 9px; text-align: center; }
.sui-chart-datepicker-grid button, .sui-chart-datepicker-empty { min-height: 25px; padding: 2px; border: 1px solid transparent; border-radius: 3px; background: transparent; color: var(--sui-chart-fg, #e6e6ef); font-size: 10px; text-align: center; }
.sui-chart-datepicker-grid button { cursor: pointer; }
.sui-chart-datepicker-grid button:hover { border-color: var(--sui-chart-control-hover-border, #4c5cf7); }
.sui-chart-datepicker-grid button.selected { background: var(--sui-chart-control-hover-border, #4c5cf7); color: #fff; }
.sui-chart-datepicker-grid button.in-range { background: var(--sui-chart-table-selection-bg, rgba(56,189,248,.22)); }
.sui-chart-calendar-add input, .sui-chart-calendar-add button, .sui-chart-calendar-actions button { box-sizing: border-box; padding: 4px 7px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; background: var(--sui-chart-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 11px; }
.sui-chart-calendar-add button, .sui-chart-calendar-actions button { cursor: pointer; }
.sui-chart-calendar-add button:hover, .sui-chart-calendar-actions button:hover { border-color: var(--sui-chart-control-hover-border, #4c5cf7); }
.sui-chart-link-fields { display: flex; flex-wrap: wrap; gap: 14px; margin-top: 12px; }
.sui-chart-link-fields label { display: flex; flex-direction: column; gap: 4px; font-size: 11px; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-link-fields input, .sui-chart-link-fields select { box-sizing: border-box; padding: 4px 7px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; background: var(--sui-chart-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 11px; }
.sui-chart-calendar-actions { display: flex; justify-content: flex-end; gap: 7px; margin-top: 14px; }
.sui-chart-calendar-actions .primary { border-color: var(--sui-chart-control-hover-border, #4c5cf7); background: var(--sui-chart-control-hover-border, #4c5cf7); color: #fff; }
.sui-chart-gantt-commit-dialog label { display: flex; flex-direction: column; gap: 4px; margin-top: 10px; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; }
.sui-chart-gantt-commit-dialog input, .sui-chart-gantt-commit-dialog textarea { box-sizing: border-box; width: 100%; padding: 5px 7px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; background: var(--sui-chart-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 12px; }
.sui-chart-gantt-history { display: flex; flex-direction: column; gap: 8px; margin-top: 10px; }
.sui-chart-gantt-history-item { padding: 9px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 5px; background: var(--sui-chart-bg, #14171c); }
.sui-chart-gantt-history-meta { margin: 3px 0 6px; color: var(--sui-chart-muted, #8fa0b8); font-size: 11px; }
.sui-chart-gantt-history-item details { margin: 4px 0 8px; color: var(--sui-chart-fg, #e6e6ef); font-size: 11px; }
.sui-chart-gantt-history-item summary { cursor: pointer; color: var(--sui-chart-muted, #8fa0b8); }
.sui-chart-gantt-history-item ul { margin: 5px 0 0; padding-left: 18px; max-height: 160px; overflow: auto; }
.sui-chart-gantt-history-item button { padding: 4px 7px; border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 4px; background: var(--sui-chart-bg, #14171c); color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 11px; cursor: pointer; }
.sui-chart-menu[hidden] { display: none; }
.sui-chart-menu button { display: block; width: 100%; padding: 5px 8px; border: 0; border-radius: 4px; background: transparent; color: var(--sui-chart-fg, #e6e6ef); font: inherit; font-size: 12px; text-align: left; cursor: pointer; }
.sui-chart-menu button:hover, .sui-chart-menu button:focus-visible { background: var(--sui-chart-control-hover-bg, #2c3444); outline: none; }
.sui-chart-menu-group { position: relative; }
.sui-chart-menu-submenu-trigger { display: flex !important; justify-content: space-between; align-items: center; }
.sui-chart-menu-submenu { position: absolute; left: calc(100% - 2px); top: 0; min-width: 150px; padding: 4px; background: var(--sui-chart-control-bg, #1d2129); border: 1px solid var(--sui-chart-control-border, #475569); border-radius: 6px; box-shadow: 0 8px 24px rgba(0,0,0,.4); }
.sui-chart-menu-submenu[hidden] { display: none; }
.sui-chart-menu-submenu.align-left { left: auto; right: calc(100% - 2px); }
.sui-chart-hit { cursor: pointer; }
.sui-chart-active { filter: brightness(1.18); }
.sui-chart-interactive .sui-chart-hit:hover,
.sui-chart-hit.sui-chart-active { filter: brightness(1.18); }
.sui-chart-hit.sui-chart-selected {
  stroke: var(--sui-chart-control-hover-border, var(--sui-chart-series-1, #4c5cf7)) !important;
  stroke-width: 2.5 !important;
}
.sui-chart-gantt-task { cursor: grab; stroke: var(--sui-chart-gantt-bar-border, transparent); stroke-width: 1; }
.sui-chart-gantt-task:active { cursor: grabbing; }
.sui-chart-gantt-item.sui-chart-gantt-dragging .sui-chart-gantt-task { cursor: grabbing; }
.sui-chart-gantt-item.sui-chart-gantt-selected .sui-chart-gantt-task { stroke: var(--sui-chart-table-selection-border, var(--sui-chart-series-1, #4c5cf7)); stroke-width: 3; }
.sui-chart-gantt-label-selected { fill: var(--sui-chart-table-selection-border, var(--sui-chart-series-1, #38bdf8)) !important; stroke: var(--sui-chart-table-selection-border, var(--sui-chart-series-1, #38bdf8)); stroke-width: .8; paint-order: stroke fill; font-weight: 700; }
.sui-chart-gantt-label-selected .sui-chart-gantt-wbs-header { fill: inherit; }
.sui-chart-gantt-label-selected.sui-chart-gantt-wbs-header { fill: var(--sui-chart-fg, #e6e6ef) !important; }
.sui-chart-gantt-item.sui-chart-gantt-flash .sui-chart-gantt-task { animation: sui-chart-task-flash .9s ease-out; }
@keyframes sui-chart-task-flash { 0% { filter: brightness(1.7); stroke: var(--sui-chart-control-hover-border, #4c5cf7); stroke-width: 3; } 100% { filter: brightness(1); } }
.sui-chart-gantt-group { cursor: pointer; font-weight: 600; }
.sui-chart-gantt-group-hit { fill: transparent; cursor: pointer; }
.sui-chart-gantt-summary[data-sui-chart-gantt-group],
.sui-chart-gantt-summary[data-sui-chart-gantt-wbs-toggle],
.sui-chart-gantt-wbs-header[data-sui-chart-gantt-wbs-toggle] { cursor: pointer; }
.sui-checkbox-wrap, .sui-radio-wrap, .sui-switch-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.sui-checkbox, .sui-radio {
  appearance: none;
  width: 18px;
  height: 18px;
  border: 1px solid var(--sui-border);
  background: #1d2129;
  border-radius: 4px;
  position: relative;
  cursor: pointer;
  transition: all 0.15s ease;
}
.sui-radio {
  border-radius: 50%;
}
:where(.sui-checkbox):checked, :where(.sui-radio):checked {
  background: #4c5cf7;
  border-color: #4c5cf7;
}
.sui-checkbox:checked::after {
  content: "";
  position: absolute;
  left: 5px;
  top: 2px;
  width: 5px;
  height: 9px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}
.sui-radio:checked::after {
  content: "";
  position: absolute;
  left: 5px;
  top: 5px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: white;
}
.sui-switch {
  appearance: none;
  width: 42px;
  height: 24px;
  background: #374151;
  border-radius: 999px;
  position: relative;
  cursor: pointer;
  transition: background 0.2s ease;
}
.sui-switch::after {
  content: "";
  position: absolute;
  left: 2px;
  top: 2px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s ease;
}
:where(.sui-switch):checked {
  background: #4c5cf7;
}
.sui-switch:checked::after {
  transform: translateX(18px);
}
.sui-slider {
  appearance: none;
  width: 100%;
  height: 6px;
  background: #374151;
  border-radius: 999px;
  outline: none;
}
.sui-slider::-webkit-slider-thumb {
  appearance: none;
  width: 18px;
  height: 18px;
  background: #4c5cf7;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(76, 92, 247, 0.4);
}
.sui-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  background: #4c5cf7;
  border-radius: 50%;
  cursor: pointer;
  border: none;
}
.sui-progress {
  width: 100%;
  height: 8px;
  border-radius: 999px;
  background: #374151;
  overflow: hidden;
}
.sui-progress::-webkit-progress-bar {
  background: #374151;
}
.sui-progress::-webkit-progress-value {
  background: linear-gradient(90deg, #4c5cf7, #6366f1);
}
.sui-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  background: rgba(76, 92, 247, 0.15);
  color: #a5b4fc;
}
.sui-alert {
  padding: 12px 16px;
  border-radius: var(--sui-radius);
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #fca5a5;
}
.sui-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid rgba(156, 163, 175, 0.3);
  border-top-color: #4c5cf7;
  border-radius: 50%;
  animation: sui-spin 0.8s linear infinite;
}
@keyframes sui-spin {
  to { transform: rotate(360deg); }
}
.sui-card {
  background: #1a1d26;
  border: 1px solid var(--sui-border);
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}
.sui-divider {
  border: none;
  border-top: 1px solid var(--sui-border);
  margin: 16px 0;
}

.sui-view {
  width: 100%;
  margin: 0 auto;
  background: var(--sui-panel, #1b1f27);
  color: var(--sui-fg, #e6e6e6);
  border-radius: 12px;
  padding: 24px;
  box-shadow: var(--sui-ac-shadow, 0 4px 18px rgba(0, 0, 0, 0.45));
}

h1, h2, h3, h4, h5, h6 {
  color: var(--sui-fg, #f2f4f6);
  line-height: 1.3;
}

h1 { font-size: 1.75em; }
h2 { font-size: 1.4em; }
h3 { font-size: 1.15em; }

p { line-height: 1.6; }

.row {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.col {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.spacer { height: 16px; }

button {
  background: var(--sui-button-bg, #2d3340);
  color: var(--sui-button-fg, #e6e6e6);
  border: 1px solid var(--sui-button-border, #3d4556);
  border-radius: 8px;
  padding: 8px 16px;
  font-size: 14px;
  cursor: pointer;
}

:where(button):hover { background: var(--sui-button-hover, #3a4356); }

input {
  background: var(--sui-bg);
  color: var(--sui-fg);
  border: 1px solid var(--sui-border);
  border-radius: var(--sui-radius);
  padding: 8px 12px;
  font-size: 14px;
}

.sui-ac {
  position: absolute;
  z-index: 1000;
  background: var(--sui-bg);
  border: 1px solid var(--sui-border);
  border-radius: var(--sui-radius);
  max-height: 220px;
  overflow-y: auto;
  box-shadow: var(--sui-ac-shadow);
  padding: 4px 0;
}
.sui-ac-item {
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
  color: var(--sui-fg);
}
.sui-ac-item:hover, .sui-ac-item.active { background: var(--sui-hover); }

a {
  color: var(--sui-accent, #7db3ff);
  text-decoration: none;
}

:where(a):hover { text-decoration: underline; }

code, pre {
  font-family: ui-monospace, "SF Mono", Consolas, Menlo, monospace;
  background: var(--sui-panel2, #10131a);
  color: var(--sui-fg, #e6e6e6);
  border-radius: 6px;
  padding: 2px 6px;
}

pre {
  display: block;
  padding: 14px 16px;
  overflow-x: auto;
  line-height: 1.5;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin: 8px 0 16px;
  font-size: 14px;
}

thead th {
  text-align: left;
  color: var(--sui-fg, #f2f4f6);
  background: var(--sui-panel, #202732);
  border-bottom: 2px solid var(--sui-border, #3d4556);
}

th, td {
  padding: 10px 14px;
  min-width: 90px;
  min-height: 36px;
  border: 1px solid var(--sui-border, #262c38);
  vertical-align: middle;
}

tbody tr:hover { background: var(--sui-table-hover-bg, #20252e); }

/* fixed-size tables: long cell text is truncated; click a cell to see it all */
table[data-sui-fixed] { table-layout: fixed; }
table[data-sui-fixed] th, table[data-sui-fixed] td { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
table[data-sui-fixed] td:focus, table[data-sui-fixed] th:focus { outline: 2px solid #4c5cf7; outline-offset: -2px; }
/* editable tables: every cell can be edited by the end user */
table[data-sui-editable] td, table[data-sui-editable] th { cursor: text; }
table[data-sui-editable] td:focus, table[data-sui-editable] th:focus { outline: 2px solid #4c5cf7; outline-offset: -2px; background: var(--sui-table-focus-bg, rgba(76,92,247,.10)); }
table[data-sui-editable] td, table[data-sui-editable] th { -webkit-tap-highlight-color: transparent; }
table[data-sui-editable] ::selection { background: rgba(76,92,247,.22); color: inherit; }
.sui-cell-pop { position: fixed; z-index: 9999; max-width: 420px; background: var(--sui-panel, #202732); color: var(--sui-fg, #f2f4f6); border: 1px solid var(--sui-border, #3d4556); border-radius: 8px; padding: 8px 10px; box-shadow: var(--sui-ac-shadow, 0 8px 22px rgba(0,0,0,.45)); font-size: 13px; line-height: 1.45; white-space: pre-wrap; word-break: break-word; }
.sui-iframe-error { padding: 10px 14px; border: 1px solid #a33; border-radius: 8px; background: #2b1719; color: #ffb4b4; font-size: 13px; line-height: 1.4; }
/* Drag and drop: the runtime toggles these classes on the current source and
   drop target. The error is inserted as a sibling with textContent. */
.sui-dragging { opacity: .55; }
.sui-drop-active { outline: 2px dashed #4c5cf7 !important; outline-offset: 2px; background: rgba(76, 92, 247, .10) !important; }
.sui-drop-error { color: #ffb4b4; background: #2b1719; border: 1px solid #a33; border-radius: 6px; padding: 6px 8px; margin-top: 4px; font-size: 12px; line-height: 1.4; }

/* Menu bar: a horizontal nav with button/popup menus. The runtime toggles
   the open class on .sui-menu; the popup is inside the normal flow only when open. */
.sui-menubar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 2px;
  padding: 4px;
  background: var(--sui-panel, #1d2129);
  border: 1px solid var(--sui-border);
  border-radius: var(--sui-radius);
  position: relative;
  z-index: 20;
}
.sui-menu { position: relative; display: inline-block; }
.sui-menu-button {
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--sui-fg);
  border-radius: 6px;
  padding: 7px 12px;
  font: inherit;
  cursor: pointer;
  white-space: nowrap;
}
.sui-menu-button:hover,
.sui-menu-button:focus-visible,
.sui-menu.open > .sui-menu-button { background: var(--sui-hover); outline: none; }
.sui-menu-popup {
  display: none;
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 190px;
  padding: 4px;
  background: #1d2129;
  border: 1px solid var(--sui-border);
  border-radius: 8px;
  box-shadow: var(--sui-ac-shadow);
  z-index: 1000;
}
.sui-menu.open > .sui-menu-popup { display: block; }
.sui-menu-popup > a { display: block; color: inherit; text-decoration: none; }
.sui-menuitem {
  display: block;
  width: 100%;
  padding: 8px 12px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--sui-fg);
  font: inherit;
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
}
.sui-menuitem:hover,
.sui-menuitem:focus { background: var(--sui-hover); outline: none; }
.sui-menu-popup .sui-divider { margin: 4px 0; }
/* A menu nested inside a popup is a submenu: same markup, side placement. */
.sui-menu-popup .sui-menu { display: block; }
.sui-menu-popup .sui-menu > .sui-menu-button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 8px 12px;
  text-align: left;
}
.sui-menu-popup .sui-menu > .sui-menu-button::after {
  content: "\25B8";
  color: var(--sui-muted, #8fa0b8);
  font-size: 10px;
  margin-left: 12px;
}
.sui-menu-popup .sui-menu > .sui-menu-popup {
  left: calc(100% + 4px);
  top: -4px;
}
/* Tabs: tablist + panels. The runtime toggles active/hidden and restores the
   selected tab after a view re-render when data-sui-tabs-key is present. */
.sui-tabs {
  border: 1px solid var(--sui-border);
  border-radius: var(--sui-radius);
  overflow: hidden;
}
.sui-tablist {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  padding: 4px 4px 0;
  background: #1d2129;
  border-bottom: 1px solid var(--sui-border);
}
.sui-tab {
  appearance: none;
  border: 0;
  background: transparent;
  color: #9aa7b8;
  border-radius: 6px 6px 0 0;
  padding: 8px 14px;
  font: inherit;
  cursor: pointer;
  white-space: nowrap;
}
.sui-tab:hover,
.sui-tab:focus-visible { background: var(--sui-hover); color: var(--sui-fg); outline: none; }
.sui-tab.active {
  color: var(--sui-fg);
  background: #14171c;
  box-shadow: inset 0 -2px 0 #4c5cf7;
}
.sui-tabpanels { padding: 12px; }
.sui-tabpanel[hidden] { display: none; }

/* Accordion: stacked disclosure sections. The runtime keeps one panel open
   and preserves the open index across view re-renders when keyed. */
.sui-accordion {
  border: 1px solid var(--sui-border);
  border-radius: var(--sui-radius);
  overflow: hidden;
  background: #1d2129;
}
.sui-accordion-item + .sui-accordion-item { border-top: 1px solid var(--sui-border); }
.sui-accordion-head {
  appearance: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  border: 0;
  background: transparent;
  color: var(--sui-fg);
  padding: 11px 14px;
  font: inherit;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
}
.sui-accordion-head::after {
  content: "\25B8";
  color: #9aa7b8;
  font-size: 13px;
  line-height: 1;
}
.sui-accordion-item.open > .sui-accordion-head::after { content: "\25BE"; }
.sui-accordion-head:hover,
.sui-accordion-head:focus-visible { background: var(--sui-hover); outline: none; }
.sui-accordion-panel { padding: 0 14px 12px; color: var(--sui-fg); }
.sui-accordion-panel[hidden] { display: none; }

/* Toast notifications: a host receives text-only .sui-toast elements. */
.sui-toast-host {
  display: flex;
  flex-direction: column;
  gap: 8px;
  position: fixed;
  top: var(--sui-toast-offset-y, 16px);
  right: var(--sui-toast-offset-x, 16px);
  left: auto;
  bottom: auto;
  width: max-content;
  max-width: calc(100vw - var(--sui-toast-offset-x, 16px) - var(--sui-toast-offset-x, 16px));
  z-index: 5000;
  transform: translate(var(--sui-toast-tx, 0), var(--sui-toast-ty, 0));
  pointer-events: none;
}
.sui-toast-host:empty { display: none; }
.sui-toast-host-fixed {
  position: fixed;
  top: var(--sui-toast-offset-y, 16px);
  right: var(--sui-toast-offset-x, 16px);
  left: auto;
  bottom: auto;
  width: max-content;
  max-width: calc(100vw - var(--sui-toast-offset-x, 16px) - var(--sui-toast-offset-x, 16px));
  z-index: 5000;
  transform: translate(var(--sui-toast-tx, 0), var(--sui-toast-ty, 0));
}
/* Horizontal and vertical anchors are independent. An axis set to center has
   no edge offset; its translate variable recenters the host. */
.sui-toast-host.sui-toast-h-left { left: var(--sui-toast-offset-x, 16px); right: auto; --sui-toast-tx: 0; }
.sui-toast-host.sui-toast-h-center { left: 50%; right: auto; --sui-toast-tx: -50%; }
.sui-toast-host.sui-toast-h-right { right: var(--sui-toast-offset-x, 16px); left: auto; --sui-toast-tx: 0; }
.sui-toast-host.sui-toast-v-top { top: var(--sui-toast-offset-y, 16px); bottom: auto; --sui-toast-ty: 0; }
.sui-toast-host.sui-toast-v-center { top: 50%; bottom: auto; --sui-toast-ty: -50%; }
.sui-toast-host.sui-toast-v-bottom { bottom: var(--sui-toast-offset-y, 16px); top: auto; --sui-toast-ty: 0; }
/* Bottom-anchored stacks grow upwards: the newest toast must sit above the
   previous one. column-reverse keeps the DOM append order (newest last) while
   the oldest message stays nearest the bottom edge. */
.sui-toast-host.sui-toast-v-bottom { flex-direction: column-reverse; }
.sui-toast {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  width: var(--sui-toast-width, 360px);
  max-width: 100%;
  padding: 11px 12px;
  border: 1px solid var(--sui-border);
  border-left: 4px solid var(--sui-toast-accent, #6b8afd);
  border-radius: var(--sui-toast-radius, 8px);
  background: var(--sui-toast-bg, #1d2129);
  color: var(--sui-toast-fg, var(--sui-fg));
  font-size: 13px;
  line-height: 1.45;
  box-shadow: var(--sui-ac-shadow);
  pointer-events: auto;
  animation: sui-toast-in .18s ease-out;
}
@keyframes sui-toast-in {
  from { opacity: 0; transform: translateY(-6px); }
  to { opacity: 1; transform: translateY(0); }
}
.sui-toast-success { --sui-toast-accent: #3fb950; --sui-toast-icon-bg: #1f6f36; }
.sui-toast-warning { --sui-toast-accent: #d29922; --sui-toast-icon-bg: #7a5a12; }
.sui-toast-error { --sui-toast-accent: #f85149; --sui-toast-icon-bg: #8f2424; }
.sui-toast-icon {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 19px;
  height: 19px;
  margin-top: 1px;
  border-radius: 50%;
  background: var(--sui-toast-icon-bg, #2b3444);
  color: var(--sui-toast-icon-fg, var(--sui-fg));
  font-size: 12px;
  font-weight: 700;
}
.sui-toast-message {
  flex: 1;
  min-width: 0;
  font-size: inherit;
  line-height: inherit;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
  color: var(--sui-toast-message-fg, inherit);
}
.sui-toast-close {
  flex: none;
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--sui-toast-close-fg, #9aa7b8);
  font: inherit;
  font-size: 17px;
  line-height: 1;
  padding: 0 2px;
  cursor: pointer;
}
.sui-toast-close:hover, .sui-toast-close:focus-visible { color: var(--sui-fg); outline: none; }

/* Tooltips: text-only bubbles created by the runtime and positioned in the
   viewport, so they work on every element even when it is a void control. */
.sui-tooltip-bubble {
  position: fixed;
  left: 0;
  top: 0;
  z-index: 6000;
  width: max-content;
  max-width: var(--sui-tooltip-max-width, 260px);
  padding: 6px 9px;
  border: 1px solid var(--sui-tooltip-border, var(--sui-border, #3a4152));
  border-radius: var(--sui-tooltip-radius, 6px);
  background: var(--sui-tooltip-bg, #1d2129);
  color: var(--sui-tooltip-fg, var(--sui-fg, #e6e6ef));
  box-shadow: var(--sui-ac-shadow, 0 8px 24px rgba(0,0,0,.45));
  font: inherit;
  font-size: 12px;
  line-height: 1.35;
  text-align: left;
  white-space: normal;
  overflow-wrap: anywhere;
  pointer-events: none;
}
.sui-tooltip-bubble[hidden] { display: none; }
.sui-tooltip-bubble.open { animation: sui-tooltip-in .12s ease-out; }
@keyframes sui-tooltip-in { from { opacity: 0; } to { opacity: 1; } }
/* Modal dialogs */
.sui-modal {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: none;
  align-items: center;
  justify-content: center;
  padding: 16px;
}
.sui-modal.open { display: flex; }
.sui-modal[hidden] { display: none !important; }
.sui-modal-backdrop { position: absolute; inset: 0; background: rgba(0, 0, 0, 0.58); }
.sui-modal-panel {
  position: relative;
  z-index: 1;
  width: min(520px, 100%);
  max-height: calc(100vh - 32px);
  overflow: auto;
  background: #1d2129;
  color: var(--sui-fg);
  border: 1px solid var(--sui-border);
  border-radius: 10px;
  box-shadow: var(--sui-ac-shadow);
}
.sui-modal-title { padding: 16px 18px 0; font-size: 17px; font-weight: 700; }
.sui-modal-body { padding: 14px 18px; }
.sui-modal-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0 18px 16px;
}
.sui-modal-button {
  appearance: none;
  border: 1px solid var(--sui-border);
  background: #252b36;
  color: var(--sui-fg);
  border-radius: 6px;
  padding: 8px 14px;
  font: inherit;
  cursor: pointer;
}
.sui-modal-button:hover, .sui-modal-button:focus-visible { background: var(--sui-hover); outline: none; }
.sui-modal-button.primary { background: #4c5cf7; border-color: #4c5cf7; color: #fff; }
.sui-modal-button.danger { background: #a33; border-color: #a33; color: #fff; }
body.sui-modal-open { overflow: hidden; }
`

// suiTableRuntime is the table runtime: it binds tabular data to the table
// named with `name:` (table.display / table.data / ...), keeps the designer's
// row and cell styling, re-applies displayed data after every render, and
// wires the editable/fixed behaviors (contenteditable, truncation overlay,
// draggable row/column boundaries).
const suiTableRuntime = `const __SUI_TBL = (function(){
  // ---- data binding --------------------------------------------------------
  // store[name] = { header: [] | null, rows: [[...]] }. It survives view
  // re-renders: afterRender() re-applies it to the freshly rendered table.
  var store = Object.create(null);
  var tpls = new WeakMap(); // table element -> { heads, bodies } row templates
  // Resource limits: the table is a DOM widget, so huge payloads are refused
  // instead of freezing the page. Applications should paginate.
  var MAX_ROWS = (__SUI_SEC && __SUI_SEC.maxRows) || 100000;
  var MAX_COLS = (__SUI_SEC && __SUI_SEC.maxCols) || 10000;
  var MAX_TEXT = (__SUI_SEC && __SUI_SEC.maxText) || 20000000;

  function isObj(v){ return v !== null && typeof v === 'object' && !Array.isArray(v); }
  function cellStr(v){
    if (v === null || v === undefined) return '';
    if (typeof v === 'object') { try { return JSON.stringify(v); } catch(e){ return String(v); } }
    return String(v);
  }
  // normValue normalizes already-parsed data: arrays of arrays, arrays of
  // objects, {columns, rows} payloads, key/value objects and scalars.
  function normValue(data){
    if (data === null || data === undefined) return { header: null, rows: [] };
    if (Array.isArray(data)) {
      if (!data.length) return { header: null, rows: [] };
      if (data.length > MAX_ROWS) return { error: 'table.display: too many rows (limit ' + MAX_ROWS + ')' };
      var rows = [], header = null, i, j;
      if (Array.isArray(data[0])) {
        for (i = 0; i < data.length; i++) {
          var src = Array.isArray(data[i]) ? data[i] : [data[i]], rr = [];
          if (src.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
          for (j = 0; j < src.length; j++) rr.push(cellStr(src[j]));
          rows.push(rr);
        }
      } else if (isObj(data[0])) {
        header = [];
        for (i = 0; i < data.length; i++) {
          if (!isObj(data[i])) continue;
          for (var k in data[i]) {
            if (Object.prototype.hasOwnProperty.call(data[i], k) && header.indexOf(k) < 0) header.push(k);
          }
          if (header.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
        }
        for (i = 0; i < data.length; i++) {
          var o = isObj(data[i]) ? data[i] : {}, r2 = [];
          for (j = 0; j < header.length; j++) r2.push(cellStr(o[header[j]]));
          rows.push(r2);
        }
      } else {
        for (i = 0; i < data.length; i++) rows.push([cellStr(data[i])]);
      }
      return { header: header, rows: rows };
    }
    if (isObj(data)) {
      if (Array.isArray(data.rows)) {
        var sub = normValue(data.rows);
        if (sub.error) return sub;
        var h = null;
        if (Array.isArray(data.columns)) h = data.columns.map(cellStr);
        else if (Array.isArray(data.headers)) h = data.headers.map(cellStr);
        else h = sub.header;
        if (h && h.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
        return { header: h, rows: sub.rows };
      }
      var ks = Object.keys(data), kv = [];
      if (ks.length > MAX_COLS) return { error: 'table.display: too many columns (limit ' + MAX_COLS + ')' };
      for (var x = 0; x < ks.length; x++) kv.push([ks[x], cellStr(data[ks[x]])]);
      return { header: ['key', 'value'], rows: kv };
    }
    return { header: null, rows: [[cellStr(data)]] };
  }

  // ---- delimited text (CSV, TSV, semicolon/pipe separated) -----------------
  function countSep(line, sep){
    var n = 0, q = false, i, ch;
    for (i = 0; i < line.length; i++) {
      ch = line.charAt(i);
      if (ch === '"') q = !q;
      else if (!q && ch === sep) n++;
    }
    return n;
  }
  function detectSep(text){
    var line = String(text).split(/\r?\n/)[0] || '';
    var cands = [',', ';', '\t', '|'], best = '', bestN = 0, i, n;
    for (i = 0; i < cands.length; i++) {
      n = countSep(line, cands[i]);
      if (n > bestN) { bestN = n; best = cands[i]; }
    }
    return best || ',';
  }
  function delimFor(fmt){
    if (fmt === 'tsv') return '\t';
    if (fmt === 'ssv') return ';';
    if (fmt === 'psv') return '|';
    return ',';
  }
  // RFC4180-style parser: quoted fields, "" escapes, embedded delimiters and
  // newlines inside quotes, CRLF, UTF-8 BOM.
  function parseDelimited(text, sep, quote, trim){
    var rows = [], row = [], field = '', inQ = false, i, ch;
    text = String(text).replace(/^\uFEFF/, '');
    for (i = 0; i < text.length; i++) {
      ch = text.charAt(i);
      if (inQ) {
        if (ch === quote) {
          if (text.charAt(i + 1) === quote) { field += quote; i++; }
          else inQ = false;
        } else field += ch;
      } else if (quote && ch === quote && field === '') {
        inQ = true;
      } else if (ch === sep) {
        row.push(trim ? field.trim() : field); field = '';
        if (row.length > MAX_COLS) return null;
      } else if (ch === '\n') {
        row.push(trim ? field.trim() : field); rows.push(row); row = []; field = '';
        if (rows.length > MAX_ROWS) return null;
      } else if (ch === '\r') {
        // CRLF: the following \n ends the row
      } else {
        field += ch;
      }
    }
    if (field !== '' || row.length) {
      row.push(trim ? field.trim() : field);
      if (row.length > MAX_COLS) return null;
      rows.push(row);
      if (rows.length > MAX_ROWS) return null;
    }
    if (rows.length && rows[rows.length - 1].length === 1 && rows[rows.length - 1][0] === '') rows.pop();
    return rows;
  }
  function normDelimited(text, header, sep, quote, trim){
    var s = sep || detectSep(text);
    var rows = parseDelimited(text, s, quote === undefined ? '"' : quote, !!trim);
    if (!rows) return { error: 'table.importText: too many rows or columns (limit ' + MAX_ROWS + 'x' + MAX_COLS + ')' };
    if (!rows.length) return { header: null, rows: [] };
    var useHeader = header === null || header === undefined ? rows.length > 1 : !!header;
    var h = useHeader ? rows.shift() : null;
    return { header: h, rows: rows, textHeader: useHeader, sep: s };
  }
  function tryJsonLines(text){
    var lines = String(text).split(/\r?\n/), vals = [], i, line;
    for (i = 0; i < lines.length; i++) {
      line = lines[i].trim();
      if (!line) continue;
      try { vals.push(JSON.parse(line)); } catch(e) { return null; }
      if (vals.length > MAX_ROWS) return { error: 'table.display: too many rows (limit ' + MAX_ROWS + ')' };
    }
    if (vals.length < 2) return null;
    return normValue(vals);
  }
  // startsJSON checks the first character without embedding raw delimiters in
  // the runtime (keeps the generated source's bracket count balanced).
  function startsJSON(s){
    var c = String(s).charCodeAt(0);
    return c === 91 || c === 123; // opening bracket or brace
  }
  // spec syntax: "csv", "tsv", "ssv" (;), "psv" (|), "json", "jsonl", "auto",
  // plus ;header=0|1 ;sep=comma|semicolon|tab|pipe|<char> ;quote=<char> ;trim=1
  function parseSpec(spec, defFmt){
    var s = String(spec === undefined || spec === null ? '' : spec).trim().toLowerCase();
    var bits = s ? s.split(';') : [];
    var fmt = bits.length ? bits.shift().trim() : (defFmt || 'auto');
    var o = { format: fmt || 'auto', header: null, sep: '', quote: '"', trim: false,
      formula: (typeof __SUI_SEC !== 'undefined' && __SUI_SEC.csvFormulaEscape) ? 'escape' : 'raw' };
    for (var i = 0; i < bits.length; i++) {
      var b = bits[i].trim();
      if (!b) continue;
      var eq = b.indexOf('=');
      var k = (eq < 0 ? b : b.slice(0, eq)).trim();
      var v = eq < 0 ? '' : b.slice(eq + 1).trim();
      if (k === 'header') o.header = !(v === '0' || v === 'false' || v === 'no' || v === 'off');
      else if (k === 'sep') {
        if (v === 'tab' || v === '\\t') o.sep = '\t';
        else if (v === 'comma') o.sep = ',';
        else if (v === 'semicolon') o.sep = ';';
        else if (v === 'pipe') o.sep = '|';
        else o.sep = v;
      }
      else if (k === 'quote') o.quote = v;
      else if (k === 'trim') o.trim = (v === '1' || v === 'true' || v === 'yes' || v === 'on');
      else if (k === 'formula' || k === 'safe') {
        o.formula = (v === 'escape' || v === 'safe' || v === '1' || v === 'true' || v === 'yes') ? 'escape' : 'raw';
      }
    }
    return o;
  }
  // normText parses an explicit text payload according to spec.
  function normText(text, spec){
    var o = parseSpec(spec, 'auto');
    var s = String(text === undefined || text === null ? '' : text).trim();
    if (!s) return { header: null, rows: [] };
    if (s.length > MAX_TEXT) return { error: 'table.importText: text too large (limit ' + MAX_TEXT + ' characters)' };
    if (o.format === 'json') {
      try { return normValue(JSON.parse(s)); }
      catch(e) { return { error: 'table.importText: invalid JSON (' + e.message + ')' }; }
    }
    if (o.format === 'jsonl' || o.format === 'ndjson') {
      var jl = tryJsonLines(s);
      return jl || { error: 'table.importText: invalid JSON Lines' };
    }
    if (o.format === 'auto') {
      if (startsJSON(s)) {
        var jerr = null;
        try { return normValue(JSON.parse(s)); } catch(e) { jerr = e; }
        var auto = tryJsonLines(s);
        if (auto) return auto;
        return { error: 'table.importText: invalid JSON (' + (jerr ? jerr.message : 'parse error') + ')' };
      }
    }
    var sep = o.sep || (o.format === 'auto' ? detectSep(s) : delimFor(o.format));
    var useHeader = o.header === null ? true : !!o.header;
    return normDelimited(s, useHeader, sep, o.quote, o.trim);
  }
  // norm is the table.display() entry point: it sniffs JSON, JSON Lines and
  // delimited text, and treats the first row of a multi-row delimited payload
  // as the header.
  function norm(data){
    if (typeof data !== 'string') return normValue(data);
    var s = data.trim();
    if (!s) return { header: null, rows: [] };
    if (s.length > MAX_TEXT) return { error: 'table.display: text too large (limit ' + MAX_TEXT + ' characters)' };
    if (startsJSON(s)) {
      var err = null;
      try { return normValue(JSON.parse(s)); } catch(e) { err = e; }
      var jl = tryJsonLines(s);
      if (jl) return jl;
      return { error: 'table.display: invalid JSON (' + (err ? err.message : 'parse error') + ')' };
    }
    return normDelimited(s, null, detectSep(s), '"', false);
  }

  function findTables(name){
    var out = [], ts = document.querySelectorAll('table');
    for (var i = 0; i < ts.length; i++) { if (ts[i].getAttribute('name') === name) out.push(ts[i]); }
    return out;
  }
  function cellsOf(tr){
    var out = [], ch = tr.children;
    for (var i = 0; i < ch.length; i++) { if (ch[i].tagName === 'TD' || ch[i].tagName === 'TH') out.push(ch[i]); }
    return out;
  }
  function tbodyOf(tbl){
    if (tbl.tBodies && tbl.tBodies.length) return tbl.tBodies[0];
    var tb = tbl.querySelector('tbody');
    if (!tb) { tb = document.createElement('tbody'); tbl.appendChild(tb); }
    return tb;
  }
  function bodyRows(tbl){
    var tb = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
    if (tb) return Array.prototype.slice.call(tb.rows);
    var out = [], trs = tbl.querySelectorAll('tr');
    for (var i = 0; i < trs.length; i++) { if (!trs[i].parentNode || trs[i].parentNode.tagName !== 'THEAD') out.push(trs[i]); }
    return out;
  }
  function templates(tbl){
    var t = tpls.get(tbl);
    if (t) return t;
    var bodies = [], heads = [], tb = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
    if (tb) for (var i = 0; i < tb.rows.length; i++) bodies.push(tb.rows[i].cloneNode(true));
    var hr = tbl.querySelector('thead tr');
    if (hr) heads.push(hr.cloneNode(true));
    if (!bodies.length) bodies.push(hr ? hr.cloneNode(true) : document.createElement('tr'));
    t = { heads: heads, bodies: bodies };
    tpls.set(tbl, t);
    return t;
  }
  function toTd(cell){
    if (cell.tagName !== 'TH') return cell;
    var td = document.createElement('td');
    if (cell.className) td.className = cell.className;
    var st = cell.getAttribute('style');
    if (st) td.setAttribute('style', st);
    return td;
  }
  function prepRow(tr, n, want){
    var cells = cellsOf(tr);
    while (cells.length > n) { tr.removeChild(cells.pop()); }
    while (cells.length < n) {
      var c = cells.length ? cells[cells.length - 1].cloneNode(false) : document.createElement(want === 'td' ? 'td' : 'th');
      tr.appendChild(c); cells.push(c);
    }
    for (var i = 0; i < cells.length; i++) {
      cells[i].removeAttribute('colspan');
      cells[i].removeAttribute('rowspan');
      cells[i].removeAttribute('data-' + 'sui-id');
      cells[i].removeAttribute('contenteditable');
      if (want === 'td') {
        var td = toTd(cells[i]);
        if (td !== cells[i]) tr.replaceChild(td, cells[i]);
        cells[i] = td;
      }
    }
    return cells;
  }
  function applyTo(tbl, ds){
    var t = templates(tbl);
    var n = 0, i, j;
    var hr = tbl.querySelector('thead tr');
    var rows = ds.rows;
    // Delimited text always carries its first row as the header. When the
    // designed table has no <thead>, keep that row visible as the first body
    // row instead of dropping it.
    if (ds.textHeader && ds.header && !hr) rows = [ds.header].concat(rows);
    if (ds.header) n = ds.header.length;
    for (i = 0; i < rows.length; i++) n = Math.max(n, rows[i].length);
    if (!ds.header) {
      // No header information in the data: keep the designed column count and
      // labels (arrays/JSON arrays carry no column names).
      if (hr) n = Math.max(n, cellsOf(hr).length);
      if (t.bodies.length && t.bodies[0].children) n = Math.max(n, cellsOf(t.bodies[0]).length);
    }
    if (!n) {
      var tb0 = tbl.tBodies && tbl.tBodies.length ? tbl.tBodies[0] : null;
      if (tb0) while (tb0.firstChild) tb0.removeChild(tb0.firstChild);
      return;
    }
    if (hr) {
      var hc = prepRow(hr, n, 'th');
      if (ds.header) {
        for (j = 0; j < hc.length; j++) hc[j].textContent = j < ds.header.length ? ds.header[j] : '';
      } else if (t.heads.length) {
        // restore the labels as designed (a previous object payload may have
        // replaced them with key names)
        var orig = cellsOf(t.heads[0]);
        for (j = 0; j < hc.length; j++) hc[j].textContent = j < orig.length ? orig[j].textContent : '';
      }
    }
    var tb = tbodyOf(tbl);
    while (tb.firstChild) tb.removeChild(tb.firstChild);
    for (i = 0; i < rows.length; i++) {
      var tr = t.bodies[i % t.bodies.length].cloneNode(true);
      tr.removeAttribute('data-' + 'sui-id');
      var cs = prepRow(tr, n, 'td');
      for (j = 0; j < cs.length; j++) cs[j].textContent = rows[i][j] !== undefined ? rows[i][j] : '';
      tb.appendChild(tr);
    }
  }
  function storeApply(name, ds){
    store[name] = ds;
    var ts = findTables(name);
    for (var i = 0; i < ts.length; i++) {
      try { applyTo(ts[i], ds); } catch(e) { return { error: String(e) }; }
    }
    init();
    return { n: ds.rows.length };
  }
  function display(name, data){
    var ds = norm(data);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    return r.error ? [0, r.error] : [r.n, null];
  }
  // importText parses text explicitly: see parseSpec for the spec syntax.
  function importText(name, text, spec){
    var ds = normText(text, spec);
    if (ds.error) return [0, ds.error];
    var r = storeApply(name, ds);
    return r.error ? [0, r.error] : [r.n, null];
  }
  function data(name){
    var ts = findTables(name);
    if (!ts.length) {
      var st = store[name];
      return [st ? st.rows.map(function(r){ return r.slice(); }) : [], null];
    }
    var out = [], rows = bodyRows(ts[0]);
    for (var i = 0; i < rows.length; i++) {
      var cs = cellsOf(rows[i]), r = [];
      for (var j = 0; j < cs.length; j++) r.push((cs[j].textContent || '').trim());
      out.push(r);
    }
    return [out, null];
  }
  function header(name){
    var ts = findTables(name), out = [];
    if (!ts.length) {
      var st = store[name];
      return [st && st.header ? st.header.slice() : [], null];
    }
    var hr = ts[0].querySelector('thead tr');
    if (hr) {
      var cs = cellsOf(hr);
      for (var i = 0; i < cs.length; i++) out.push((cs[i].textContent || '').trim());
    }
    return [out, null];
  }
  function json(name){
    var h = header(name)[0], rows = data(name)[0], out = [], i, j;
    if (h && h.length) {
      for (i = 0; i < rows.length; i++) {
        var o = Object.create(null);
        for (j = 0; j < rows[i].length; j++) o[h[j] || ('col' + (j + 1))] = rows[i][j];
        out.push(o);
      }
      return [JSON.stringify(out), null];
    }
    return [JSON.stringify(rows), null];
  }
  function select(name, cols){
    if (!Array.isArray(cols)) return [[], 'table.select: columns must be an array of indexes'];
    var rows = data(name)[0], out = [];
    for (var i = 0; i < rows.length; i++) {
      var r = [];
      for (var j = 0; j < cols.length; j++) {
        var c = Number(cols[j]);
        r.push(c >= 0 && c < rows[i].length ? rows[i][c] : '');
      }
      out.push(r);
    }
    return [out, null];
  }
  function clear(name){
    delete store[name];
    var ts = findTables(name);
    for (var i = 0; i < ts.length; i++) {
      var tb = ts[i].tBodies && ts[i].tBodies.length ? ts[i].tBodies[0] : null;
      if (tb) while (tb.firstChild) tb.removeChild(tb.firstChild);
    }
    return [0, null];
  }
  function slice(name, from, count){
    var rows = data(name)[0];
    var a = Math.max(0, Number(from) || 0);
    var n = Number(count) || 0;
    return [n > 0 ? rows.slice(a, a + n) : rows.slice(a), null];
  }
  // exportText serializes the visible table as CSV/TSV/SSV/PSV/JSON/JSONL.
  function exportText(name, spec){
    var o = parseSpec(spec, 'csv');
    if (o.format === 'json') return json(name);
    var h = header(name)[0], rows = data(name)[0], i, j, ob;
    if (o.format === 'jsonl' || o.format === 'ndjson') {
      var objs = [];
      if (h.length) {
        for (i = 0; i < rows.length; i++) {
          ob = Object.create(null);
          for (j = 0; j < rows[i].length; j++) ob[h[j] || ('col' + (j + 1))] = rows[i][j];
          objs.push(ob);
        }
      } else {
        objs = rows;
      }
      var txt = objs.map(function(v){ return JSON.stringify(v); }).join('\n');
      return [txt ? txt + '\n' : '', null];
    }
    var sep = o.sep || delimFor(o.format);
    var q = o.quote === undefined ? '"' : o.quote;
    function esc(v){
      v = v === null || v === undefined ? '' : String(v);
      // CSV/Excel formula injection: opt-in prefix for cells that a
      // spreadsheet would evaluate as a formula.
      if (o.formula === 'escape' && /^[=+\-@\t\r]/.test(v)) v = "'" + v;
      if (q && (v.indexOf(sep) >= 0 || v.indexOf(q) >= 0 || v.indexOf('\n') >= 0 || v.indexOf('\r') >= 0)) {
        return q + v.split(q).join(q + q) + q;
      }
      return v;
    }
    var lines = [];
    var useHeader = o.header === null ? true : !!o.header;
    if (useHeader && h.length) lines.push(h.map(esc).join(sep));
    for (i = 0; i < rows.length; i++) lines.push(rows[i].map(esc).join(sep));
    return [lines.join('\n'), null];
  }
  function applyAll(){
    for (var k in store) {
      var ts = findTables(k);
      for (var i = 0; i < ts.length; i++) { try { applyTo(ts[i], store[k]); } catch(e) {} }
    }
  }
  function afterRender(){ applyAll(); init(); }

  // ---- editable / fixed behaviors -----------------------------------------
  function init(){
    var eds = document.querySelectorAll('table[data-sui-editable] td, table[data-sui-editable] th');
    for (var i=0;i<eds.length;i++){ var c=eds[i]; if (c.getAttribute('contenteditable') !== 'true') { c.setAttribute('contenteditable','true'); c.spellcheck=false; } }
    var fix = document.querySelectorAll('table[data-sui-fixed] td, table[data-sui-fixed] th');
    for (var j=0;j<fix.length;j++){
      var d=fix[j];
      if (!d.hasAttribute('tabindex')) d.tabIndex=0;
      var isEd = d.getAttribute('contenteditable') === 'true';
      if (!isEd) { var txt = d.textContent; if (d.getAttribute('data-full') !== txt) d.setAttribute('data-full', txt); }
    }
  }
  var pop=null, popCell=null;
  function closePop(){ if(pop){ if(pop.parentNode) pop.parentNode.removeChild(pop); pop=null; popCell=null; } }
  function openPop(c, txt){
    if (popCell === c) return;
    closePop();
    pop=document.createElement('div'); pop.className='sui-cell-pop'; pop.textContent=txt;
    var cs=getComputedStyle(c);
    var bg=cs.getPropertyValue('--pop-bg').trim(); if(bg) pop.style.background=bg;
    var fg=cs.getPropertyValue('--pop-color').trim(); if(fg) pop.style.color=fg;
    var fs=cs.getPropertyValue('--pop-size').trim(); if(fs) pop.style.fontSize=fs;
    var mw=cs.getPropertyValue('--pop-maxw').trim(); if(mw) pop.style.maxWidth=mw;
    var br=cs.getPropertyValue('--pop-radius').trim(); if(br) pop.style.borderRadius=br;
    document.body.appendChild(pop);
    var r=c.getBoundingClientRect();
    var pr=pop.getBoundingClientRect();
    var left=Math.max(8, Math.min(window.innerWidth - pr.width - 8, r.left));
    var top=r.top - pr.height - 8;
    if (top < 8) top = Math.min(window.innerHeight - pr.height - 8, r.bottom + 6);
    pop.style.left=left+'px'; pop.style.top=top+'px';
    popCell=c;
  }
  document.addEventListener('mouseover', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-fixed] td, table[data-sui-fixed] th') : null;
    if(!c){ return; }
    if(document.activeElement === c) return; // being edited: text is already visible
    if(c.scrollWidth <= c.clientWidth + 1 && c.scrollHeight <= c.clientHeight + 1) { closePop(); return; }
    openPop(c, c.textContent);
  }, true);
  document.addEventListener('mouseout', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-fixed] td, table[data-sui-fixed] th') : null;
    if(c) closePop();
  }, true);
  document.addEventListener('keydown', function(ev){ if(ev.key==='Escape') closePop(); });

  // ---- draggable row/column boundaries (editable tables only) --------------
  var drag=null;
  function columnCells(cell){
    var tr=cell.parentNode, idx=Array.prototype.indexOf.call(tr.children, cell);
    var rows=cell.closest('table').querySelectorAll('tr'), out=[];
    for(var i=0;i<rows.length;i++){ var c=rows[i].children[idx]; if(c) out.push(c); }
    return out;
  }
  function edge(el, ev){
    var r=el.getBoundingClientRect();
    if ((r.right - ev.clientX) <= 7 && ev.clientX >= r.left + 10) return 'col';
    if ((r.bottom - ev.clientY) <= 7 && ev.clientY >= r.top + 8) return 'row';
    return '';
  }
  document.addEventListener('mousemove', function(ev){
    if(drag) return;
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-editable] td, table[data-sui-editable] th') : null;
    if(!c) return;
    var e=edge(c, ev);
    c.style.cursor = e === 'col' ? 'col-resize' : e === 'row' ? 'row-resize' : '';
  }, true);
  document.addEventListener('mousedown', function(ev){
    var t=ev.target;
    var c = t && t.closest ? t.closest('table[data-sui-editable] td, table[data-sui-editable] th') : null;
    if(!c) return;
    var e=edge(c, ev);
    if(!e) return;
    ev.preventDefault(); ev.stopPropagation();
    var r=c.getBoundingClientRect();
    drag = { kind:e, cell:c, start: e==='col' ? ev.clientX : ev.clientY, size: e==='col' ? r.width : r.height };
    document.body.style.cursor = e==='col' ? 'col-resize' : 'row-resize';
  }, true);
  document.addEventListener('mousemove', function(ev){
    if(!drag) return;
    ev.preventDefault();
    if(drag.kind==='col'){
      var w=Math.max(24, drag.size + (ev.clientX - drag.start));
      var cells=columnCells(drag.cell);
      for(var i=0;i<cells.length;i++){ cells[i].style.width=w+'px'; cells[i].style.minWidth=w+'px'; }
    } else {
      var h=Math.max(18, drag.size + (ev.clientY - drag.start));
      var tr=drag.cell.parentNode;
      tr.style.height=h+'px';
      for(var j=0;j<tr.children.length;j++){ tr.children[j].style.height=h+'px'; }
    }
  }, true);
  document.addEventListener('mouseup', function(){ if(drag){ drag=null; document.body.style.cursor=''; } }, true);

  // ---- keep edited cells in the store so re-renders do not lose them -------
  document.addEventListener('input', function(ev){
    var c = ev.target;
    if (!c || !c.closest) return;
    c = c.closest('td,th');
    if (!c) return;
    var tbl = c.closest('table');
    if (!tbl) return;
    var name = tbl.getAttribute('name');
    if (!name || !store[name]) return;
    var tr = c.parentNode, col = Array.prototype.indexOf.call(tr.children, c);
    if (col < 0) return;
    if (tr.parentNode && tr.parentNode.tagName === 'THEAD') {
      if (store[name].header) store[name].header[col] = c.textContent || '';
    } else {
      var tb = tr.parentNode, r = Array.prototype.indexOf.call(tb.rows, tr);
      if (r >= 0) {
        if (!store[name].rows[r]) store[name].rows[r] = [];
        store[name].rows[r][col] = c.textContent || '';
      }
    }
  }, true);

  if(window.MutationObserver){
    var pending=false;
    new MutationObserver(function(){ if(pending) return; pending=true; setTimeout(function(){ pending=false; init(); }, 60); }).observe(document.body,{childList:true,subtree:true,characterData:true});
  }
  init();
  return {
    display: display, data: data, header: header, json: json,
    select: select, slice: slice, clear: clear, importText: importText,
    exportText: exportText, applyAll: applyAll, afterRender: afterRender
  };
})();
window.__SUI_TBL = __SUI_TBL;`

// concatTextParts joins the literal text parts of an interpolated string,
// skipping {expr} parts. Returns "" when there are no text parts.
func concatTextParts(parts []StrPart) string {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "text" {
			sb.WriteString(parts[i].text)
		}
	}
	return sb.String()
}

// htmlEscape escapes the characters that are special inside HTML text/attrs:
// & < > ". UTF-8 bytes pass through untouched. Escaping applies to generated
// text content only (PageTitle/text stay raw); separate concern.
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// PageTitle returns the literal text of the first view's title (interpolation
// parts are ignored), or "" when there is no view/title. No HTML escaping.
func PageTitle(p *Prog) string {
	if p == nil || len(p.views) == 0 {
		return ""
	}
	return concatTextParts(p.views[0].title)
}

// GenerateIndexHtml emits the index.html skeleton from SPEC §5 with the
// HTML-escaped PageTitle in <title>. A missing/empty title still emits an
// empty <title></title> element.
func GenerateIndexHtml(p *Prog) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>`)
	sb.WriteString(htmlEscape(PageTitle(p)))
	sb.WriteString(`</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
<div id="app"></div>
<script src="app.js"></script>
</body>
</html>
`)
	return sb.String()
}

// GenerateCss emits the base stylesheet plus one deduplicated (sorted)
// pass-through rule per literal class name found in any view element.
func GenerateCss(p *Prog) string {
	var sb strings.Builder
	sb.WriteString(suiCssBase)
	var classes []string
	if p != nil {
		for i := 0; i < len(p.views); i++ {
			collectClassesInView(p.views[i].body, &classes)
		}
	}
	sort.Strings(classes)
	prev := ""
	for i := 0; i < len(classes); i++ {
		c := classes[i]
		if c == "" || c == prev {
			continue
		}
		sb.WriteString("." + c + " { /* generated */ }\n")
		prev = c
	}
	if p != nil {
		for i := 0; i < len(p.css); i++ {
			sb.WriteString("\n/* user css */\n")
			sb.WriteString(p.css[i])
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// collectClassesInView walks the (dereferenced) nodes of a view body.
func collectClassesInView(nodes []ViewNode, out *[]string) {
	var ptrs []*ViewNode
	for i := 0; i < len(nodes); i++ {
		ptrs = append(ptrs, &nodes[i])
	}
	collectClassesPtrs(ptrs, out)
}

// collectClassesPtrs walks a []*ViewNode list, descending into element
// children, if/then-else branches and list bodies.
func collectClassesPtrs(nodes []*ViewNode, out *[]string) {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		if n == nil {
			continue
		}
		if n.kind == "element" {
			collectClassesElem(n.elem, out)
		} else if n.kind == "if" {
			collectClassesPtrs(n.then_, out)
			collectClassesPtrs(n.else_, out)
		} else if n.kind == "list" {
			collectClassesPtrs(n.body, out)
		}
	}
}

// collectClassesElem reads the element's own literal class attribute (if any)
// and then recurses into its children.
func collectClassesElem(e *Element, out *[]string) {
	if e == nil {
		return
	}
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		if a.key != "class" {
			continue
		}
		if !a.hasExpr {
			addClassTokens(concatTextParts(a.text), out)
		} else {
			collectExprClassTokens(a.expr, out)
		}
	}
	collectClassesInView(e.children, out)
}

// collectExprClassTokens collects the literal class tokens inside a class
// expression (templates, "a " + b, cond ? "on" : "off"). Dynamic values are
// filtered at runtime, but literals must remain usable without being declared
// in CSS.
func collectExprClassTokens(e *Expr, out *[]string) {
	if e == nil {
		return
	}
	switch e.kind {
	case "str":
		addClassTokens(concatTextParts(e.parts), out)
	case "binary":
		if e.op == "add" {
			collectExprClassTokens(e.l, out)
			collectExprClassTokens(e.r, out)
		}
	case "cond":
		collectExprClassTokens(e.l, out)
		collectExprClassTokens(e.r, out)
	}
}

// addClassTokens splits a class attribute value on whitespace and appends the
// tokens to out.
func addClassTokens(value string, out *[]string) {
	for _, t := range strings.Fields(value) {
		*out = append(*out, t)
	}
}

// GenerateHtmlPage renders the non-invoked views (pages) as a static HTML
// document (no JS): one .sui-view block per page, root view first. Used by the
// design tool's "HTML" tab as the human-readable source of the project.
func GenerateHtmlPage(p *Prog) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<title>")
	sb.WriteString(PageTitle(p))
	sb.WriteString("</title>\n</head>\n<body>\n<div id=\"app\">\n")
	if p != nil && len(p.views) > 0 {
		viewNames := map[string]bool{}
		for i := 0; i < len(p.views); i++ {
			viewNames[p.views[i].name] = true
		}
		invoked := map[string]bool{}
		for i := 0; i < len(p.views); i++ {
			collectTags(p.views[i].body, viewNames, invoked)
		}
		for i := 0; i < len(p.views); i++ {
			if invoked[p.views[i].name] {
				continue
			}
			sb.WriteString("  <div class=\"sui-view\" data-sui-view=\"" + p.views[i].name + "\">\n")
			writeStaticNodes(&sb, p.views[i].body, "    ")
			sb.WriteString("  </div>\n")
		}
	}
	sb.WriteString("</div>\n</body>\n</html>\n")
	return sb.String()
}

func writeStaticNodes(sb *strings.Builder, nodes []ViewNode, ind string) {
	for i := 0; i < len(nodes); i++ {
		if nodes[i].kind == "element" && nodes[i].elem != nil {
			writeStaticElem(sb, nodes[i].elem, ind)
		}
	}
}

// staticDatalistSeq numbers the <datalist> elements emitted for static
// elements. Atomic: two concurrent HTTP handlers can generate HTML at the
// same time and a plain ++ would be a data race.
var staticDatalistSeq int64

func writeStaticElem(sb *strings.Builder, e *Element, ind string) {
	if e.tag == "modal" {
		writeStaticModal(sb, e, ind)
		return
	}
	if e.tag == "tabs" {
		writeStaticTabs(sb, e, ind)
		return
	}
	if e.tag == "accordion" {
		writeStaticAccordion(sb, e, ind)
		return
	}
	if e.tag == "toast" {
		writeStaticToastHost(sb, e, ind)
		return
	}
	if e.tag == "combobox" {
		writeStaticCombobox(sb, e, ind)
		return
	}
	if e.tag == "datepicker" {
		writeStaticDatePicker(sb, e, ind)
		return
	}
	if e.tag == "tree" {
		writeStaticTree(sb, e, ind)
		return
	}
	if chartTag(e.tag) {
		writeStaticChart(sb, e, ind)
		return
	}
	if e.tag == "menubar" {
		writeStaticMenuBar(sb, e, ind)
		return
	}
	if e.tag == "menu" {
		writeStaticMenu(sb, e, ind)
		return
	}
	tag := elemHtmlTag(e.tag)
	if e.tag == "tcell" && isHeaderCell(e) {
		tag = "th"
	}
	var dlId string
	if e.tag == "input" || e.tag == "checkbox" || e.tag == "radio" || e.tag == "switch" || e.tag == "slider" || e.tag == "file" || e.tag == "color" || e.tag == "date" || e.tag == "time" || e.tag == "number" {
		if e.tag == "input" {
			if opts := inputOptions(e); len(opts) > 0 {
				n := atomic.AddInt64(&staticDatalistSeq, 1)
				dlId = "sui-dl-" + strconv.FormatInt(n, 10)
			}
		}
	}
	link := staticLinkOpen(e)
	if link != "" {
		sb.WriteString(ind)
		sb.WriteString(link)
		sb.WriteString("\n")
	}
	// checkbox/switch/radio render as input + label text inside a wrapper
	// <label>; the wrapper mirrors the element's class so user styling reaches
	// the whole widget.
	wrapBase := ""
	switch e.tag {
	case "checkbox":
		wrapBase = "sui-checkbox-wrap"
	case "radio":
		wrapBase = "sui-radio-wrap"
	case "switch":
		wrapBase = "sui-switch-wrap"
	}
	sb.WriteString(ind)
	if wrapBase != "" {
		sb.WriteString("<label class=\"" + wrapBase + "\"")
		staticClassSuffix(sb, e)
		sb.WriteString(">")
	}
	sb.WriteString("<")
	sb.WriteString(tag)
	switch e.tag {
	case "row", "col", "spacer":
		sb.WriteString(" class=\"")
		sb.WriteString(e.tag)
		staticClassSuffix(sb, e)
		sb.WriteString("\"")
	case "card":
		sb.WriteString(" class=\"sui-card\"")
		staticClassSuffix(sb, e)
	case "divider":
		sb.WriteString(" class=\"sui-divider\"")
		staticClassSuffix(sb, e)
	case "badge":
		sb.WriteString(" class=\"sui-badge\"")
		staticClassSuffix(sb, e)
	case "alert":
		sb.WriteString(" class=\"sui-alert\"")
		staticClassSuffix(sb, e)
	case "spinner":
		sb.WriteString(" class=\"sui-spinner\"")
		staticClassSuffix(sb, e)
	case "button":
		sb.WriteString(" class=\"sui-button\"")
		staticClassSuffix(sb, e)
		staticUserAttrsNoClass(sb, e, dlId)
	case "progress":
		sb.WriteString(" class=\"sui-progress\"")
		staticClassSuffix(sb, e)
	case "checkbox":
		sb.WriteString(" type=\"checkbox\" class=\"sui-checkbox\"")
		staticClassSuffix(sb, e)
	case "radio":
		sb.WriteString(" type=\"radio\" class=\"sui-radio\"")
		staticClassSuffix(sb, e)
	case "switch":
		sb.WriteString(" type=\"checkbox\" class=\"sui-switch\"")
		staticClassSuffix(sb, e)
	case "slider":
		sb.WriteString(" type=\"range\" class=\"sui-slider\"")
		staticClassSuffix(sb, e)
	case "file":
		sb.WriteString(" type=\"file\" class=\"sui-file\"")
		staticClassSuffix(sb, e)
	case "color":
		sb.WriteString(" type=\"color\" class=\"sui-color\"")
		staticClassSuffix(sb, e)
	case "date":
		sb.WriteString(" type=\"date\" class=\"sui-date\"")
		staticClassSuffix(sb, e)
	case "time":
		sb.WriteString(" type=\"time\" class=\"sui-time\"")
		staticClassSuffix(sb, e)
	case "number":
		sb.WriteString(" type=\"number\" class=\"sui-number\"")
		staticClassSuffix(sb, e)
	case "textarea":
		sb.WriteString(" class=\"sui-textarea\"")
		staticClassSuffix(sb, e)
	case "select":
		sb.WriteString(" class=\"sui-select\"")
		staticClassSuffix(sb, e)
	case "menuitem":
		sb.WriteString(" class=\"sui-menuitem\"")
		staticClassSuffix(sb, e)
		sb.WriteString(" role=\"menuitem\" tabindex=\"-1\"")
		staticUserAttrsNoClass(sb, e, dlId)
	case "tab":
		sb.WriteString(" class=\"sui-tabpanel\"")
		staticClassSuffix(sb, e)
		staticUserAttrsNoClass(sb, e, dlId)
	case "modalbutton":
		sb.WriteString(" class=\"sui-modal-button")
		if v := modalVariant(e); v != "" {
			sb.WriteString(" " + v)
		}
		staticClassSuffix(sb, e)
		sb.WriteString("\"")
		staticUserAttrsNoClass(sb, e, dlId)
	default:
		staticUserAttrs(sb, e, dlId)
	}
	if e.tag == "checkbox" || e.tag == "radio" || e.tag == "switch" ||
		e.tag == "slider" || e.tag == "file" || e.tag == "color" ||
		e.tag == "date" || e.tag == "time" || e.tag == "number" ||
		e.tag == "textarea" || e.tag == "select" {
		staticUserAttrsNoClass(sb, e, dlId)
	}
	if e.tag == "table" {
		if a := findAttr(e, "editable"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
			sb.WriteString(" data-sui-editable=\"1\"")
		}
		if a := findAttr(e, "fixed"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
			sb.WriteString(" data-sui-fixed=\"1\"")
		}
	}
	sb.WriteString(">")
	if e.tag == "input" || e.tag == "img" || e.tag == "hr" || e.tag == "br" || e.tag == "checkbox" || e.tag == "radio" || e.tag == "switch" || e.tag == "slider" || e.tag == "file" || e.tag == "color" || e.tag == "date" || e.tag == "time" || e.tag == "number" || e.tag == "divider" || e.tag == "spinner" || e.tag == "progress" {
		// HTML void tags (input/img/hr/br) close naturally; "spinner" and
		// "progress" map to <div>/<progress>, which are NOT void: without an
		// explicit close tag every following element nests inside them.
		if tag != "input" && tag != "img" && tag != "hr" && tag != "br" {
			sb.WriteString("</")
			sb.WriteString(tag)
			sb.WriteString(">")
		}
		if wrapBase != "" {
			if len(e.text) > 0 {
				sb.WriteString("<span>")
				staticText(sb, e.text)
				sb.WriteString("</span>")
			}
			sb.WriteString("</label>")
		}
		if dlId != "" {
			sb.WriteString(datalistHtml(dlId, inputOptions(e)))
		}
		sb.WriteString("\n")
		if link != "" {
			sb.WriteString(ind)
			sb.WriteString("</a>\n")
		}
		return
	}
	staticText(sb, e.text)
	if len(e.children) > 0 {
		sb.WriteString("\n")
		writeStaticNodes(sb, e.children, ind+"  ")
		sb.WriteString(ind)
	}
	sb.WriteString("</")
	sb.WriteString(tag)
	sb.WriteString(">\n")
	if link != "" {
		sb.WriteString(ind)
		sb.WriteString("</a>\n")
	}
}

func modalNameOf(e *Element) string {
	a := findAttr(e, "name")
	if a == nil || a.hasExpr {
		return ""
	}
	return strings.TrimSpace(partsContent(a.text))
}

func modalVariant(e *Element) string {
	a := findAttr(e, "variant")
	if a == nil || a.hasExpr {
		return ""
	}
	v := strings.TrimSpace(partsContent(a.text))
	switch v {
	case "primary", "danger":
		return v
	}
	return ""
}

func modalButtonCloses(e *Element) bool {
	a := findAttr(e, "close")
	if a == nil {
		return true
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		return a.expr.b
	}
	return strings.TrimSpace(strings.ToLower(partsContent(a.text))) != "false"
}

func writeStaticModal(sb *strings.Builder, e *Element, ind string) {
	name := modalNameOf(e)
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-modal")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-modal=\"")
	sb.WriteString(htmlEscape(name))
	sb.WriteString("\" aria-hidden=\"true\" hidden>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-modal-backdrop\" data-sui-modal-backdrop></div>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-modal-panel\" role=\"dialog\" aria-modal=\"true\" tabindex=\"-1\"")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	sb.WriteString(">\n")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(ind)
		sb.WriteString("    <div class=\"sui-modal-title\">")
		staticText(sb, e.text)
		sb.WriteString("</div>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("    <div class=\"sui-modal-body\">\n")
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "modalbutton" {
			continue
		}
		writeStaticNodes(sb, []ViewNode{*ch}, ind+"      ")
	}
	sb.WriteString(ind)
	sb.WriteString("    </div>\n")
	sb.WriteString(ind)
	sb.WriteString("    <div class=\"sui-modal-actions\">\n")
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind != "element" || ch.elem == nil || ch.elem.tag != "modalbutton" {
			continue
		}
		btn := ch.elem
		sb.WriteString(ind)
		sb.WriteString("      <button type=\"button\" class=\"sui-modal-button")
		if v := modalVariant(btn); v != "" {
			sb.WriteString(" " + v)
		}
		staticClassSuffix(sb, btn)
		sb.WriteString("\"")
		if modalButtonCloses(btn) {
			sb.WriteString(" data-sui-modal-dismiss=\"1\"")
		}
		sb.WriteString(">")
		staticText(sb, btn.text)
		sb.WriteString("</button>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("    </div>\n")
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

// writeStaticTabs emits a tablist plus tabpanels. The first tab is active in
// static HTML; the runtime keeps the selected tab across app re-renders.
func writeStaticTabs(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-tabs")
	staticClassSuffix(sb, e)
	sb.WriteString("\"")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	var tabs []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "tab" {
			tabs = append(tabs, ch.elem)
		}
	}
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-tablist\" role=\"tablist\">\n")
	for i, t := range tabs {
		active := i == 0
		sb.WriteString(ind)
		sb.WriteString("    <button type=\"button\" class=\"sui-tab")
		if active {
			sb.WriteString(" active")
		}
		sb.WriteString("\" role=\"tab\" aria-selected=\"")
		if active {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		sb.WriteString("\" tabindex=\"")
		if active {
			sb.WriteString("0")
		} else {
			sb.WriteString("-1")
		}
		sb.WriteString("\">")
		staticText(sb, t.text)
		sb.WriteString("</button>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-tabpanels\">\n")
	for i, t := range tabs {
		active := i == 0
		sb.WriteString(ind)
		sb.WriteString("    <div class=\"sui-tabpanel")
		staticClassSuffix(sb, t)
		if active {
			sb.WriteString(" active")
		}
		sb.WriteString("\" role=\"tabpanel\"")
		if !active {
			sb.WriteString(" hidden")
		}
		sb.WriteString(">\n")
		writeStaticNodes(sb, t.children, ind+"      ")
		sb.WriteString(ind)
		sb.WriteString("    </div>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

// writeStaticAccordion emits disclosure sections. The first section is open in
// static HTML; the browser runtime toggles the open item.
func writeStaticAccordion(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-accordion")
	staticClassSuffix(sb, e)
	sb.WriteString("\"")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	open := true
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind != "element" || ch.elem == nil || ch.elem.tag != "accordionitem" {
			continue
		}
		item := ch.elem
		isOpen := open
		open = false
		sb.WriteString(ind)
		sb.WriteString("  <div class=\"sui-accordion-item")
		staticClassSuffix(sb, item)
		if isOpen {
			sb.WriteString(" open")
		}
		sb.WriteString("\">\n")
		sb.WriteString(ind)
		sb.WriteString("    <button type=\"button\" class=\"sui-accordion-head\" aria-expanded=\"")
		if isOpen {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		sb.WriteString("\">")
		staticText(sb, item.text)
		sb.WriteString("</button>\n")
		sb.WriteString(ind)
		sb.WriteString("    <div class=\"sui-accordion-panel\"")
		if !isOpen {
			sb.WriteString(" hidden")
		}
		sb.WriteString(">\n")
		writeStaticNodes(sb, item.children, ind+"      ")
		sb.WriteString(ind)
		sb.WriteString("    </div>\n")
		sb.WriteString(ind)
		sb.WriteString("  </div>\n")
	}
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

// writeStaticToastHost emits an empty live-region host. toast.show() appends
// text-only notifications to it at runtime.
func toastAttrValue(e *Element, key string) string {
	a := findAttr(e, key)
	if a == nil {
		return ""
	}
	if a.hasExpr && a.expr != nil {
		return strings.Trim(strings.TrimSpace(exprSource(a.expr)), "\"'")
	}
	return strings.TrimSpace(partsLiteral(a.text))
}

// toastPlacementClasses turns the compact designer/source attributes into
// static CSS hooks. Position and inset stay in CSS so they never fight the
// card style that the inspector writes for .sui-toast.
func toastAxisValues(e *Element) (h, v, x, y string) {
	h = strings.ToLower(strings.TrimSpace(toastAttrValue(e, "hAlign")))
	v = strings.ToLower(strings.TrimSpace(toastAttrValue(e, "vAlign")))
	x = strings.TrimSpace(toastAttrValue(e, "hOffset"))
	y = strings.TrimSpace(toastAttrValue(e, "vOffset"))
	// Compatibility with the first compact syntax: position="top-left",
	// offset="24px". Opening the inspector and changing an axis rewrites it to
	// hAlign/vAlign/hOffset/vOffset.
	if h == "" && v == "" {
		pos := strings.ToLower(strings.TrimSpace(toastAttrValue(e, "position")))
		if parts := strings.SplitN(pos, "-", 2); len(parts) == 2 {
			v, h = parts[0], parts[1]
		}
	}
	if x == "" {
		x = strings.TrimSpace(toastAttrValue(e, "offset"))
	}
	if y == "" {
		y = x
	}
	return
}

func toastPlacementClasses(e *Element) string {
	h, v, _, _ := toastAxisValues(e)
	classes := ""
	switch h {
	case "left", "center", "right":
		classes = "sui-toast-h-" + h
	}
	switch v {
	case "top", "center", "bottom":
		if classes != "" {
			classes += " "
		}
		classes += "sui-toast-v-" + v
	}
	return classes
}

func toastOffsetValue(e *Element, axis string) string {
	_, _, x, y := toastAxisValues(e)
	if axis == "y" {
		return y
	}
	return x
}

func comboboxAttrValue(e *Element, key string) string {
	a := findAttr(e, key)
	if a == nil || a.hasExpr {
		return ""
	}
	return partsContent(a.text)
}

func comboboxAttrNumber(e *Element, key string) string {
	a := findAttr(e, key)
	if a == nil {
		return ""
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
		return strconv.Itoa(a.expr.i)
	}
	if !a.hasExpr {
		return strings.TrimSpace(partsContent(a.text))
	}
	return ""
}

func treeKey(e *Element, path string) string {
	if a := findAttr(e, "key"); a != nil && !a.hasExpr {
		if k := strings.TrimSpace(partsContent(a.text)); k != "" {
			return k
		}
	}
	return path
}

func treeItemEditable(e *Element) string {
	a := findAttr(e, "editable")
	if a == nil {
		return "inherit"
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		if a.expr.b {
			return "1"
		}
		return "0"
	}
	v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
	if v == "true" || v == "yes" || v == "1" {
		return "1"
	}
	if v == "false" || v == "no" || v == "0" {
		return "0"
	}
	return "inherit"
}

func treeItemShowRenameIcon(e *Element) string {
	a := findAttr(e, "showRenameIcon")
	if a == nil {
		return "inherit"
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		if a.expr.b {
			return "1"
		}
		return "0"
	}
	v := strings.ToLower(strings.TrimSpace(partsContent(a.text)))
	if v == "true" || v == "1" {
		return "1"
	}
	if v == "false" || v == "0" {
		return "0"
	}
	return "inherit"
}

func treeItemRenameIcon(e *Element) string {
	a := findAttr(e, "renameIcon")
	if a == nil || a.hasExpr {
		return ""
	}
	return partsContent(a.text)
}

func treeOpen(e *Element) bool {
	a := findAttr(e, "open")
	return a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b
}

func treeChildren(e *Element) []*Element {
	var out []*Element
	for i := 0; i < len(e.children); i++ {
		ch := &e.children[i]
		if ch.kind == "element" && ch.elem != nil && ch.elem.tag == "treeitem" {
			out = append(out, ch.elem)
		}
	}
	return out
}

func writeStaticTree(sb *strings.Builder, e *Element, ind string) {
	value := comboboxAttrValue(e, "value")
	title := partsContent(e.text)
	treeKeyName := strings.TrimSpace(title)
	if treeKeyName == "" {
		treeKeyName = "tree"
	}
	treeName := comboboxAttrValue(e, "name")
	if strings.TrimSpace(treeName) == "" {
		treeName = treeKeyName
	}
	editable := datepickerBool(e, "editable")
	depth := comboboxAttrNumber(e, "editableDepth")
	renameIcon := comboboxAttrValue(e, "renameIcon")
	showRenameIcon := true
	if a := findAttr(e, "showRenameIcon"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		showRenameIcon = a.expr.b
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-tree")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-tree=\"1\" data-sui-tree-name=\"" + htmlEscape(treeName) + "\" data-sui-tree-key=\"" + htmlEscape(treeKeyName) + "\" role=\"tree\"")
	if editable {
		sb.WriteString(" data-sui-tree-editable=\"1\"")
	}
	if depth != "" {
		sb.WriteString(" data-sui-tree-editable-depth=\"" + htmlEscape(depth) + "\"")
	}
	if renameIcon != "" {
		sb.WriteString(" data-sui-tree-rename-icon=\"" + htmlEscape(renameIcon) + "\"")
	}
	if !showRenameIcon {
		sb.WriteString(" data-sui-tree-show-rename-icon=\"0\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	indent := ind + "  "
	if title != "" {
		sb.WriteString(indent + "<div class=\"sui-tree-title\">" + htmlEscape(title) + "</div>\n")
	}
	sb.WriteString(indent + "<input type=\"hidden\" class=\"sui-tree-value\" value=\"" + htmlEscape(value) + "\">\n")
	sb.WriteString(indent + "<ul class=\"sui-tree-group\" role=\"group\">\n")
	writeStaticTreeItems(sb, treeChildren(e), indent+"  ", 1, "0")
	sb.WriteString(indent + "</ul>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticTreeItems(sb *strings.Builder, items []*Element, ind string, level int, pathPrefix string) {
	for i, it := range items {
		path := pathPrefix + "." + strconv.Itoa(i)
		key := treeKey(it, path)
		kids := treeChildren(it)
		open := treeOpen(it)
		label := partsContent(it.text)
		sb.WriteString(ind)
		sb.WriteString("<li class=\"sui-tree-item")
		staticClassSuffix(sb, it)
		sb.WriteString("\" data-sui-tree-item=\"1\" data-sui-key=\"" + htmlEscape(key) +
			"\" data-sui-path=\"" + htmlEscape(path) + "\" data-sui-level=\"" + strconv.Itoa(level) +
			"\" data-sui-editable=\"" + htmlEscape(treeItemEditable(it)) +
			"\" data-sui-show-rename-icon=\"" + htmlEscape(treeItemShowRenameIcon(it)) +
			"\" data-sui-rename-icon=\"" + htmlEscape(treeItemRenameIcon(it)) + "\" data-sui-open=\"")
		if open {
			sb.WriteString("1")
		} else {
			sb.WriteString("0")
		}
		sb.WriteString("\" role=\"treeitem\" aria-level=\"" + strconv.Itoa(level) + "\" aria-selected=\"false\" tabindex=\"-1\"")
		if len(kids) > 0 {
			sb.WriteString(" aria-expanded=\"")
			if open {
				sb.WriteString("true")
			} else {
				sb.WriteString("false")
			}
			sb.WriteString("\"")
		}
		staticUserAttrsNoClass(sb, it, "")
		sb.WriteString(">\n")
		sb.WriteString(ind + "  <div class=\"sui-tree-row\">")
		if len(kids) > 0 {
			sb.WriteString("<button type=\"button\" class=\"sui-tree-toggle\" tabindex=\"-1\" aria-label=\"Toggle " + htmlEscape(label) + "\">")
			if open {
				sb.WriteString("\u25be")
			} else {
				sb.WriteString("\u25b8")
			}
			sb.WriteString("</button>")
		} else {
			sb.WriteString("<span class=\"sui-tree-spacer\"></span>")
		}
		sb.WriteString("<span class=\"sui-tree-label\">" + htmlEscape(label) + "</span></div>\n")
		if len(kids) > 0 {
			sb.WriteString(ind + "  <ul class=\"sui-tree-group\" role=\"group\"")
			if !open {
				sb.WriteString(" hidden")
			}
			sb.WriteString(">\n")
			writeStaticTreeItems(sb, kids, ind+"    ", level+1, path)
			sb.WriteString(ind + "  </ul>\n")
		}
		sb.WriteString(ind + "</li>\n")
	}
}

func chartStaticValue(e *Element, key, def string) string {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	if a.hasExpr {
		if a.expr != nil && a.expr.kind == "int" {
			return strconv.Itoa(a.expr.i)
		}
		if a.expr != nil && a.expr.kind == "float" {
			return strconv.FormatFloat(a.expr.f, 'g', -1, 64)
		}
		return def
	}
	if v := partsContent(a.text); strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func chartStaticFlag(e *Element, key string, def bool) bool {
	a := findAttr(e, key)
	if a == nil {
		return def
	}
	if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		return a.expr.b
	}
	v := strings.ToLower(strings.Trim(strings.TrimSpace(partsContent(a.text)), "\"'"))
	if v == "true" || v == "1" || v == "yes" || v == "on" {
		return true
	}
	if v == "false" || v == "0" || v == "no" || v == "off" {
		return false
	}
	return def
}

var stateChartEventAttrs = []string{
	"onShapeClick", "onShapeDoubleClick", "onShapeMouseOver", "onShapeMouseOut", "onShapeContextMenu",
	"onLinkClick", "onLinkDoubleClick", "onLinkMouseOver", "onLinkMouseOut", "onLinkContextMenu",
	"onTagClick", "onTagDoubleClick", "onTagMouseOver", "onTagMouseOut", "onTagContextMenu",
}

func writeStaticChart(sb *strings.Builder, e *Element, ind string) {
	name := chartNameValue(e)
	title := partsContent(e.text)
	ariaLabel := strings.TrimSpace(title)
	if ariaLabel == "" {
		if e.tag == "graph" {
			ariaLabel = "Graph"
		} else {
			ariaLabel = "Chart"
		}
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-chart")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-chart=\"1\" role=\"group\" aria-label=\"" + htmlEscape(ariaLabel) + "\" data-sui-chart-name=\"" + htmlEscape(name) +
		"\" data-sui-chart-title=\"" + htmlEscape(title) + "\"")
	if e.tag == "graph" {
		sb.WriteString(" data-sui-chart-type=\"state\"")
	} else {
		sb.WriteString(" data-sui-chart-type=\"" + htmlEscape(chartStaticValue(e, "chartType", "bar")) + "\"")
	}
	sb.WriteString(" data-sui-chart-theme=\"" + htmlEscape(chartStaticValue(e, "theme", "dark")) +
		"\" data-sui-chart-data=\"" + htmlEscape(chartStaticValue(e, "data", "")) +
		"\" data-sui-chart-format=\"" + htmlEscape(chartStaticValue(e, "format", "auto")) +
		"\" data-sui-chart-x=\"" + htmlEscape(chartStaticValue(e, "x", "")) +
		"\" data-sui-chart-y=\"" + htmlEscape(chartStaticValue(e, "y", "")) + "\"")
	if findAttr(e, "width") != nil {
		sb.WriteString(" data-sui-chart-width=\"" + htmlEscape(chartStaticValue(e, "width", "640px")) + "\"")
	}
	if findAttr(e, "height") != nil {
		sb.WriteString(" data-sui-chart-height=\"" + htmlEscape(chartStaticValue(e, "height", "320px")) + "\"")
	}
	sb.WriteString(" data-sui-chart-overflow=\"" + htmlEscape(chartStaticValue(e, "overflow", "scroll")) +
		"\" data-sui-chart-zoom=\"" + htmlEscape(chartStaticValue(e, "zoom", "1")) +
		"\" data-sui-chart-point-width=\"" + htmlEscape(chartStaticValue(e, "pointWidth", "48")) +
		"\" data-sui-chart-point-size=\"" + htmlEscape(chartStaticValue(e, "pointSize", "3")) +
		"\" data-sui-chart-table-view=\"" + htmlEscape(chartStaticValue(e, "tableView", "chart")) +
		"\" data-sui-chart-table-filters=\"" + boolAttr01(chartStaticFlag(e, "tableFilters", true)) +
		"\" data-sui-chart-table-add-row=\"" + boolAttr01(chartStaticFlag(e, "tableAddRow", false)) +
		"\" data-sui-chart-table-delete-row=\"" + boolAttr01(chartStaticFlag(e, "tableDeleteRow", false)) +
		"\" data-sui-chart-table-steppers=\"" + boolAttr01(chartStaticFlag(e, "tableSteppers", false)) +
		"\" data-sui-chart-table-context-menu=\"" + boolAttr01(chartStaticFlag(e, "tableContextMenu", false)) +
		"\" data-sui-chart-table-selection=\"" + htmlEscape(chartStaticValue(e, "tableSelection", "row")) +
		"\" data-sui-chart-table-edit=\"" + boolAttr01(chartStaticFlag(e, "tableEdit", false)) +
		"\" data-sui-chart-critical-path=\"" + boolAttr01(chartStaticFlag(e, "criticalPath", false)) +
		"\" data-sui-chart-holidays=\"" + htmlEscape(chartStaticValue(e, "holidays", "")) +
		"\" data-sui-chart-weekend=\"" + htmlEscape(chartStaticValue(e, "weekend", "")) +
		"\" data-sui-chart-timezone=\"" + htmlEscape(chartStaticValue(e, "timezone", "")) +
		"\" data-sui-chart-show-non-working-days=\"" + boolAttr01(chartStaticFlag(e, "showNonWorkingDays", true)) +
		"\" data-sui-chart-progress-column=\"" + boolAttr01(chartStaticFlag(e, "progressColumn", false)) +
		"\" data-sui-chart-bins=\"" + htmlEscape(chartStaticValue(e, "bins", "10")) +
		"\" data-sui-chart-min=\"" + htmlEscape(chartStaticValue(e, "min", "")) +
		"\" data-sui-chart-max=\"" + htmlEscape(chartStaticValue(e, "max", "")) +
		"\" data-sui-chart-interactive=\"" + boolAttr01(chartStaticFlag(e, "interactive", false)) +
		"\" data-sui-chart-color-picker=\"" + boolAttr01(chartStaticFlag(e, "colorPicker", true)) +
		"\" data-sui-chart-settings-button=\"" + boolAttr01(chartStaticFlag(e, "settingsButton", true)) +
		"\" data-sui-chart-show-zoom=\"" + boolAttr01(chartStaticFlag(e, "showZoom", true)) +
		"\" data-sui-chart-show-add=\"" + boolAttr01(chartStaticFlag(e, "showAdd", true)) +
		"\" data-sui-chart-show-history=\"" + boolAttr01(chartStaticFlag(e, "showHistory", true)) +
		"\" data-sui-chart-show-resources=\"" + boolAttr01(chartStaticFlag(e, "showResources", false)) +
		"\" data-sui-chart-show-progress=\"" + boolAttr01(chartStaticFlag(e, "showProgress", false)) +
		"\" data-sui-chart-show-slack=\"" + boolAttr01(chartStaticFlag(e, "showSlack", false)) +
		"\" data-sui-chart-show-costs=\"" + boolAttr01(chartStaticFlag(e, "showCosts", false)) +
		"\" data-sui-chart-status-colors=\"" + boolAttr01(chartStaticFlag(e, "statusColors", false)) +
		"\" data-sui-chart-auto-schedule=\"" + boolAttr01(chartStaticFlag(e, "autoSchedule", false)) +
		"\" data-sui-chart-gantt-scale=\"" + htmlEscape(chartStaticValue(e, "ganttScale", "auto")) +
		"\" data-sui-chart-link-color=\"" + htmlEscape(chartStaticValue(e, "linkColor", "#94a3b8")) +
		"\" data-sui-chart-state-shape-border-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeBorderColor", "")) +
		"\" data-sui-chart-state-selection-color=\"" + htmlEscape(chartStaticValue(e, "stateSelectionColor", "")) +
		"\" data-sui-chart-link-width=\"" + htmlEscape(chartStaticValue(e, "linkWidth", "1")) +
		"\" data-sui-chart-link-gap=\"" + htmlEscape(chartStaticValue(e, "linkGap", "4")) +
		"\" data-sui-chart-bar-border-color=\"" + htmlEscape(chartStaticValue(e, "barBorderColor", "")) +
		"\" data-sui-chart-link-style=\"" + htmlEscape(chartStaticValue(e, "linkStyle", "auto")) +
		"\" data-sui-chart-arrow-type=\"" + htmlEscape(chartStaticValue(e, "arrowType", "triangle")) +
		"\" data-sui-chart-arrow-size=\"" + htmlEscape(chartStaticValue(e, "arrowSize", "")) +
		"\" data-sui-chart-state-start-arrow=\"" + boolAttr01(chartStaticFlag(e, "stateStartArrow", false)) +
		"\" data-sui-chart-state-routing=\"" + htmlEscape(chartStaticValue(e, "stateRouting", "orthogonal")) +
		"\" data-sui-chart-state-tag-style=\"" + htmlEscape(chartStaticValue(e, "stateTagStyle", "pill")) +
		"\" data-sui-chart-state-tag-link=\"" + htmlEscape(chartStaticValue(e, "stateTagLink", "line")) +
		"\" data-sui-chart-state-link-tag-style=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagStyle", "")) +
		"\" data-sui-chart-state-link-tag-link=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLink", "")) +
		"\" data-sui-chart-state-shape-tag-style=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagStyle", "box")) +
		"\" data-sui-chart-state-shape-tag-link=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLink", "line")) +
		"\" data-sui-chart-state-shape-tag-kind=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagKind", "external")) +
		"\" data-sui-chart-state-shape-tag-align=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagAlign", "line")) +
		"\" data-sui-chart-state-shape-tag-rotation=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagRotation", "")) +
		"\" data-sui-chart-state-shape-tag-arrow-size=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagArrowSize", "")) +
		"\" data-sui-chart-state-link-tag-kind=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagKind", "external")) +
		"\" data-sui-chart-state-link-tag-align=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagAlign", "line")) +
		"\" data-sui-chart-state-link-tag-rotation=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagRotation", "")) +
		"\" data-sui-chart-state-link-tag-arrow-size=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagArrowSize", "")) +
		"\" data-sui-chart-state-tag-font=\"" + htmlEscape(chartStaticValue(e, "stateTagFont", "")) +
		"\" data-sui-chart-state-tag-font-size=\"" + htmlEscape(chartStaticValue(e, "stateTagFontSize", "11")) +
		"\" data-sui-chart-state-shape-tag-font=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFont", "")) +
		"\" data-sui-chart-state-shape-tag-font-size=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFontSize", "")) +
		"\" data-sui-chart-state-link-tag-font=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFont", "")) +
		"\" data-sui-chart-state-link-tag-font-size=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFontSize", "")) +
		"\" data-sui-chart-state-tag-font-weight=\"" + htmlEscape(chartStaticValue(e, "stateTagFontWeight", "400")) +
		"\" data-sui-chart-state-tag-font-style=\"" + htmlEscape(chartStaticValue(e, "stateTagFontStyle", "normal")) +
		"\" data-sui-chart-state-tag-text-decoration=\"" + htmlEscape(chartStaticValue(e, "stateTagTextDecoration", "none")) +
		"\" data-sui-chart-state-tag-text-align=\"" + htmlEscape(chartStaticValue(e, "stateTagTextAlign", "center")) +
		"\" data-sui-chart-state-tag-letter-spacing=\"" + htmlEscape(chartStaticValue(e, "stateTagLetterSpacing", "")) +
		"\" data-sui-chart-state-tag-line-spacing=\"" + htmlEscape(chartStaticValue(e, "stateTagLineSpacing", "")) +
		"\" data-sui-chart-state-tag-text-transform=\"" + htmlEscape(chartStaticValue(e, "stateTagTextTransform", "none")) +
		"\" data-sui-chart-state-tag-text-shadow=\"" + htmlEscape(chartStaticValue(e, "stateTagTextShadow", "none")) +
		"\" data-sui-chart-state-tag-text-shadow-color=\"" + htmlEscape(chartStaticValue(e, "stateTagTextShadowColor", "#000000")) +
		"\" data-sui-chart-state-shape-tag-font-weight=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFontWeight", "")) +
		"\" data-sui-chart-state-shape-tag-font-style=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagFontStyle", "")) +
		"\" data-sui-chart-state-shape-tag-text-decoration=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextDecoration", "")) +
		"\" data-sui-chart-state-shape-tag-text-align=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextAlign", "")) +
		"\" data-sui-chart-state-shape-tag-letter-spacing=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLetterSpacing", "")) +
		"\" data-sui-chart-state-shape-tag-line-spacing=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLineSpacing", "")) +
		"\" data-sui-chart-state-shape-tag-text-transform=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextTransform", "")) +
		"\" data-sui-chart-state-shape-tag-text-shadow=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextShadow", "")) +
		"\" data-sui-chart-state-shape-tag-text-shadow-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextShadowColor", "")) +
		"\" data-sui-chart-state-link-tag-font-weight=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFontWeight", "")) +
		"\" data-sui-chart-state-link-tag-font-style=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagFontStyle", "")) +
		"\" data-sui-chart-state-link-tag-text-decoration=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextDecoration", "")) +
		"\" data-sui-chart-state-link-tag-text-align=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextAlign", "")) +
		"\" data-sui-chart-state-link-tag-letter-spacing=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLetterSpacing", "")) +
		"\" data-sui-chart-state-link-tag-line-spacing=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLineSpacing", "")) +
		"\" data-sui-chart-state-link-tag-text-transform=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextTransform", "")) +
		"\" data-sui-chart-state-link-tag-text-shadow=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextShadow", "")) +
		"\" data-sui-chart-state-link-tag-text-shadow-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextShadowColor", "")) +
		"\" data-sui-chart-state-tag-text-color=\"" + htmlEscape(chartStaticValue(e, "stateTagTextColor", "")) +
		"\" data-sui-chart-state-tag-box-color=\"" + htmlEscape(chartStaticValue(e, "stateTagBoxColor", "")) +
		"\" data-sui-chart-state-tag-border-color=\"" + htmlEscape(chartStaticValue(e, "stateTagBorderColor", "")) +
		"\" data-sui-chart-state-tag-border-width=\"" + htmlEscape(chartStaticValue(e, "stateTagBorderWidth", "")) +
		"\" data-sui-chart-state-tag-border-style=\"" + htmlEscape(chartStaticValue(e, "stateTagBorderStyle", "solid")) +
		"\" data-sui-chart-state-tag-padding=\"" + htmlEscape(chartStaticValue(e, "stateTagPadding", "")) +
		"\" data-sui-chart-state-tag-leader-color=\"" + htmlEscape(chartStaticValue(e, "stateTagLeaderColor", "")) +
		"\" data-sui-chart-state-tag-arrow-shape=\"" + htmlEscape(chartStaticValue(e, "stateTagArrowShape", "triangle")) +
		"\" data-sui-chart-state-shape-tag-text-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagTextColor", "")) +
		"\" data-sui-chart-state-shape-tag-box-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBoxColor", "")) +
		"\" data-sui-chart-state-shape-tag-border-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBorderColor", "")) +
		"\" data-sui-chart-state-shape-tag-border-width=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBorderWidth", "")) +
		"\" data-sui-chart-state-shape-tag-border-style=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagBorderStyle", "")) +
		"\" data-sui-chart-state-shape-tag-padding=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagPadding", "")) +
		"\" data-sui-chart-state-shape-tag-leader-color=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagLeaderColor", "")) +
		"\" data-sui-chart-state-shape-tag-arrow-shape=\"" + htmlEscape(chartStaticValue(e, "stateShapeTagArrowShape", "")) +
		"\" data-sui-chart-state-link-tag-text-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagTextColor", "")) +
		"\" data-sui-chart-state-link-tag-box-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBoxColor", "")) +
		"\" data-sui-chart-state-link-tag-border-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBorderColor", "")) +
		"\" data-sui-chart-state-link-tag-border-width=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBorderWidth", "")) +
		"\" data-sui-chart-state-link-tag-border-style=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagBorderStyle", "")) +
		"\" data-sui-chart-state-link-tag-padding=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagPadding", "")) +
		"\" data-sui-chart-state-link-tag-leader-color=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagLeaderColor", "")) +
		"\" data-sui-chart-state-link-tag-arrow-shape=\"" + htmlEscape(chartStaticValue(e, "stateLinkTagArrowShape", "")) +
		"\" data-sui-chart-state-link-arc-radius=\"" + htmlEscape(chartStaticValue(e, "stateLinkArcRadius", "0")) +
		"\" data-sui-chart-state-grid=\"" + boolAttr01(chartStaticFlag(e, "stateGrid", true)) +
		"\" data-sui-chart-state-grid-gap=\"" + htmlEscape(chartStaticValue(e, "stateGridGap", "24")) +
		"\" data-sui-chart-state-grid-color=\"" + htmlEscape(chartStaticValue(e, "stateGridColor", "")) +
		"\" data-sui-chart-state-mode=\"" + htmlEscape(chartStaticValue(e, "stateMode", func() string {
		if e.tag == "graph" {
			return "static"
		}
		return "editing"
	}())) +
		"\" data-sui-chart-state-hints=\"" + boolAttr01(chartStaticFlag(e, "stateHints", true)) +
		"\" data-sui-chart-state-handles=\"" + boolAttr01(chartStaticFlag(e, "stateHandles", true)) +
		"\" data-sui-chart-state-tags-movable=\"" + boolAttr01(chartStaticFlag(e, "stateTagsMovable", false)) +
		"\" data-sui-chart-state-routes-editable=\"" + boolAttr01(chartStaticFlag(e, "stateRoutesEditable", false)) +
		"\" data-sui-chart-grid=\"" + boolAttr01(chartStaticFlag(e, "showGrid", true)) +
		"\" data-sui-chart-legend=\"" + boolAttr01(chartStaticFlag(e, "showLegend", true)) +
		"\" data-sui-chart-values=\"" + boolAttr01(chartStaticFlag(e, "showValues", false)) + "\"")
	if a := findAttr(e, "onTaskChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-ontaskchange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	if a := findAttr(e, "onCalendarChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-oncalendarchange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	if a := findAttr(e, "onDataChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-ondatachange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	for _, eventName := range stateChartEventAttrs {
		if a := findAttr(e, eventName); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
			sb.WriteString(" data-sui-chart-" + strings.ToLower(eventName) + "=\"" + htmlEscape(a.expr.name) + "\"")
		}
	}
	if a := findAttr(e, "onCommit"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-oncommit=\"" + htmlEscape(a.expr.name) + "\"")
	}
	if a := findAttr(e, "onResourceChange"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
		sb.WriteString(" data-sui-chart-onresourcechange=\"" + htmlEscape(a.expr.name) + "\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	inner := ind + "  "
	sb.WriteString(inner + "<div class=\"sui-chart-head\"><div class=\"sui-chart-title\">" + htmlEscape(title) + "</div>")
	if chartStaticFlag(e, "settingsButton", true) {
		sb.WriteString("<button type=\"button\" class=\"sui-chart-menu-btn\" aria-label=\"Chart settings\" title=\"Chart settings\">&#9881;</button>")
	}
	sb.WriteString("<div class=\"sui-chart-zoom\" hidden><button type=\"button\" class=\"sui-chart-zoom-btn\" data-zoom=\"out\" aria-label=\"Zoom out\" title=\"Zoom out\">\u2212</button>")
	sb.WriteString("<span class=\"sui-chart-zoom-label\">100%</span><button type=\"button\" class=\"sui-chart-zoom-btn\" data-zoom=\"in\" aria-label=\"Zoom in\" title=\"Zoom in\">+</button>")
	sb.WriteString("<button type=\"button\" class=\"sui-chart-zoom-reset\" data-zoom=\"reset\" title=\"Reset zoom\">Reset</button>")
	sb.WriteString("<button type=\"button\" class=\"sui-chart-zoom-fit\" data-zoom=\"fit\" aria-label=\"Fit to width\" title=\"Fit to width\">Fit width</button>")
	sb.WriteString("<button type=\"button\" class=\"sui-chart-zoom-fit sui-chart-zoom-fit-all\" data-zoom=\"fitall\" aria-label=\"Fit all graph elements\" title=\"Fit all graph elements\">Fit all</button></div></div>\n")
	sb.WriteString(inner + "<div class=\"sui-chart-viewport\"><div class=\"sui-chart-stage\"><svg class=\"sui-chart-svg\" role=\"img\"></svg></div></div>\n")
	sb.WriteString(inner + "<div class=\"sui-chart-legend\"></div><div class=\"sui-chart-splitter\" role=\"separator\" tabindex=\"0\" aria-orientation=\"horizontal\" aria-valuemin=\"80\" aria-valuemax=\"320\" aria-valuenow=\"180\" aria-valuetext=\"180 pixels\" aria-label=\"Resize chart and table\" title=\"Drag, or use the arrow keys, to resize chart and table\"></div><div class=\"sui-chart-table\" role=\"region\" aria-label=\"Task table\" hidden></div><div class=\"sui-chart-error\" role=\"alert\" hidden></div><div class=\"sui-chart-tooltip\" role=\"tooltip\" aria-live=\"polite\" hidden></div>\n")
	sb.WriteString(ind + "</div>\n")
}

func boolAttr01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func datepickerBool(e *Element, key string) bool {
	a := findAttr(e, key)
	return a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b
}

func writeStaticDatePicker(sb *strings.Builder, e *Element, ind string) {
	mode := strings.ToLower(strings.TrimSpace(comboboxAttrValue(e, "mode")))
	if mode != "time" && mode != "datetime" {
		mode = "date"
	}
	inputType := mode
	if inputType == "datetime" {
		inputType = "datetime-local"
	}
	value := comboboxAttrValue(e, "value")
	min := comboboxAttrValue(e, "min")
	max := comboboxAttrValue(e, "max")
	step := comboboxAttrNumber(e, "step")
	disabled := datepickerBool(e, "disabled")
	required := datepickerBool(e, "required")
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-datepicker")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-datepicker=\"1\" data-sui-mode=\"" + htmlEscape(mode) + "\"")
	if disabled {
		sb.WriteString(" data-sui-disabled=\"1\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <label class=\"sui-datepicker-label\">")
	staticText(sb, e.text)
	sb.WriteString("</label>\n")
	sb.WriteString(ind + "  <div class=\"sui-datepicker-control\">\n")
	sb.WriteString(ind + "    <input type=\"" + htmlEscape(inputType) + "\" class=\"sui-datepicker-input\" autocomplete=\"off\"")
	if min != "" {
		sb.WriteString(" min=\"" + htmlEscape(min) + "\"")
	}
	if max != "" {
		sb.WriteString(" max=\"" + htmlEscape(max) + "\"")
	}
	if step != "" {
		sb.WriteString(" step=\"" + htmlEscape(step) + "\"")
	}
	if required {
		sb.WriteString(" required")
	}
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "    <button type=\"button\" class=\"sui-datepicker-clear\" aria-label=\"Clear value\" hidden>\u00d7</button>\n")
	sb.WriteString(ind + "  </div>\n")
	sb.WriteString(ind + "  <input type=\"hidden\" class=\"sui-datepicker-value\" value=\"" + htmlEscape(value) + "\"")
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <div class=\"sui-datepicker-error\" role=\"alert\" hidden></div>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticCombobox(sb *strings.Builder, e *Element, ind string) {
	options := comboboxAttrValue(e, "options")
	placeholder := comboboxAttrValue(e, "placeholder")
	if placeholder == "" {
		placeholder = "Search..."
	}
	empty := comboboxAttrValue(e, "empty")
	if empty == "" {
		empty = "No results"
	}
	value := comboboxAttrValue(e, "value")
	maxOptionLength := comboboxAttrNumber(e, "maxOptionLength")
	maxOptions := comboboxAttrNumber(e, "maxOptions")
	disabled := false
	if a := findAttr(e, "disabled"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		disabled = a.expr.b
	}
	required := false
	if a := findAttr(e, "required"); a != nil && a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
		required = a.expr.b
	}
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-combobox")
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-combobox=\"1\" data-sui-options=\"" + htmlEscape(options) +
		"\" data-sui-placeholder=\"" + htmlEscape(placeholder) +
		"\" data-sui-empty=\"" + htmlEscape(empty) + "\"")
	if maxOptionLength != "" {
		sb.WriteString(" data-sui-max-option-length=\"" + htmlEscape(maxOptionLength) + "\"")
	}
	if maxOptions != "" {
		sb.WriteString(" data-sui-max-options=\"" + htmlEscape(maxOptions) + "\"")
	}
	if required {
		sb.WriteString(" data-sui-required=\"1\"")
	}
	if disabled {
		sb.WriteString(" data-sui-disabled=\"1\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <label class=\"sui-combobox-label\">")
	staticText(sb, e.text)
	sb.WriteString("</label>\n")
	sb.WriteString(ind + "  <div class=\"sui-combobox-control\">\n")
	sb.WriteString(ind + "    <input type=\"text\" class=\"sui-combobox-input\" role=\"combobox\" autocomplete=\"off\" spellcheck=\"false\" aria-expanded=\"false\" placeholder=\"" + htmlEscape(placeholder) + "\"")
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "    <button type=\"button\" class=\"sui-combobox-clear\" aria-label=\"Clear selection\" hidden>\u00d7</button>\n")
	sb.WriteString(ind + "    <button type=\"button\" class=\"sui-combobox-toggle\" aria-label=\"Open options\" aria-expanded=\"false\">\u25be</button>\n")
	sb.WriteString(ind + "  </div>\n")
	sb.WriteString(ind + "  <input type=\"hidden\" class=\"sui-combobox-value\" value=\"" + htmlEscape(value) + "\"")
	if disabled {
		sb.WriteString(" disabled")
	}
	sb.WriteString(">\n")
	sb.WriteString(ind + "  <div class=\"sui-combobox-popup\" role=\"listbox\" hidden></div>\n")
	sb.WriteString(ind + "  <div class=\"sui-combobox-error\" role=\"alert\" hidden></div>\n")
	sb.WriteString(ind + "</div>\n")
}

func writeStaticToastHost(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-toast-host")
	if pc := toastPlacementClasses(e); pc != "" {
		sb.WriteString(" " + pc)
	}
	staticClassSuffix(sb, e)
	sb.WriteString("\" data-sui-toast-host=\"1\" aria-live=\"polite\" aria-atomic=\"false\"")
	if v := toastAttrValue(e, "duration"); v != "" {
		sb.WriteString(" data-sui-toast-duration=\"" + htmlEscape(v) + "\"")
	}
	if v := toastAttrValue(e, "max"); v != "" {
		sb.WriteString(" data-sui-toast-max=\"" + htmlEscape(v) + "\"")
	}
	if v := toastAttrValue(e, "icons"); v != "" {
		sb.WriteString(" data-sui-toast-icons=\"" + htmlEscape(v) + "\"")
	}
	if v := toastOffsetValue(e, "x"); v != "" {
		sb.WriteString(" data-sui-toast-x=\"" + htmlEscape(v) + "\"")
	}
	if v := toastOffsetValue(e, "y"); v != "" {
		sb.WriteString(" data-sui-toast-y=\"" + htmlEscape(v) + "\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString("></div>\n")
}

// writeStaticMenuBar emits the semantic nav container used by the menu-bar
// control. The visible labels live on its `menu` children; the menubar text is
// an accessible name.
func writeStaticMenuBar(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<nav class=\"sui-menubar")
	staticClassSuffix(sb, e)
	sb.WriteString("\" role=\"menubar\"")
	if label := partsContent(e.text); label != "" {
		sb.WriteString(" aria-label=\"")
		sb.WriteString(htmlEscape(label))
		sb.WriteString("\"")
	}
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	writeStaticNodes(sb, e.children, ind+"  ")
	sb.WriteString(ind)
	sb.WriteString("</nav>\n")
}

// writeStaticMenu emits the button/popup structure for one top-level menu.
// The runtime adds the `open` class on click; without JavaScript the popup is
// hidden and the label still documents the structure.
func writeStaticMenu(sb *strings.Builder, e *Element, ind string) {
	sb.WriteString(ind)
	sb.WriteString("<div class=\"sui-menu")
	staticClassSuffix(sb, e)
	sb.WriteString("\" role=\"none\"")
	staticUserAttrsNoClass(sb, e, "")
	sb.WriteString(">\n")
	sb.WriteString(ind)
	sb.WriteString("  <button type=\"button\" class=\"sui-menu-button\" role=\"menuitem\" aria-haspopup=\"true\" aria-expanded=\"false\">")
	staticText(sb, e.text)
	sb.WriteString("</button>\n")
	sb.WriteString(ind)
	sb.WriteString("  <div class=\"sui-menu-popup\" role=\"menu\">\n")
	writeStaticNodes(sb, e.children, ind+"    ")
	sb.WriteString(ind)
	sb.WriteString("  </div>\n")
	sb.WriteString(ind)
	sb.WriteString("</div>\n")
}

func staticLinkOpen(e *Element) string {
	a := findAttr(e, "link")
	if a == nil {
		return ""
	}
	if a.hasExpr {
		return "<a href=\"\">"
	}
	return "<a href=\"" + htmlEscape(partsContent(a.text)) + "\">"
}

func staticClassSuffix(sb *strings.Builder, e *Element) {
	if a := findAttr(e, "class"); a != nil && !a.hasExpr {
		if c := partsContent(a.text); c != "" {
			sb.WriteString(" ")
			sb.WriteString(htmlEscape(c))
		}
	}
}

func staticUserAttrs(sb *strings.Builder, e *Element, dlId string) {
	staticUserAttrsOpt(sb, e, dlId, false)
}

// staticUserAttrsNoClass is used by controls that already emitted their
// built-in class, so the user class is appended exactly once by
// staticClassSuffix.
func staticUserAttrsNoClass(sb *strings.Builder, e *Element, dlId string) {
	staticUserAttrsOpt(sb, e, dlId, true)
}

func staticUserAttrsOpt(sb *strings.Builder, e *Element, dlId string, skipClass bool) {
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		if _, isEv := eventAttrJs[a.key]; isEv {
			continue
		}
		if skipClass && a.key == "class" {
			continue
		}
		if (e.tag == "modal" && a.key == "name") ||
			(e.tag == "modalbutton" && (a.key == "variant" || a.key == "close")) {
			continue
		}
		if e.tag == "toast" && (a.key == "position" || a.key == "offset" ||
			a.key == "hAlign" || a.key == "vAlign" || a.key == "hOffset" || a.key == "vOffset" ||
			a.key == "duration" || a.key == "max" || a.key == "icons") {
			continue
		}
		if e.tag == "combobox" && (a.key == "value" || a.key == "options" || a.key == "placeholder" ||
			a.key == "empty" || a.key == "disabled" || a.key == "required" ||
			a.key == "maxOptionLength" || a.key == "maxOptions") {
			continue
		}
		if e.tag == "datepicker" && (a.key == "value" || a.key == "mode" || a.key == "min" ||
			a.key == "max" || a.key == "step" || a.key == "disabled" || a.key == "required") {
			continue
		}
		if e.tag == "tree" && (a.key == "value" || a.key == "editable" || a.key == "editableDepth" ||
			a.key == "name" || a.key == "renameKey" || a.key == "renameValue" || a.key == "renameIcon" || a.key == "showRenameIcon") {
			continue
		}
		if e.tag == "treeitem" && (a.key == "key" || a.key == "open" || a.key == "editable" || a.key == "renameIcon" || a.key == "showRenameIcon") {
			continue
		}
		if chartTag(e.tag) && (a.key == "name" || a.key == "data" || a.key == "chartType" || a.key == "format" ||
			a.key == "x" || a.key == "y" || a.key == "width" || a.key == "height" || a.key == "overflow" ||
			a.key == "zoom" || a.key == "pointWidth" || a.key == "pointSize" || a.key == "bins" || a.key == "min" || a.key == "max" ||
			a.key == "tableView" || a.key == "tableFilters" || a.key == "tableAddRow" || a.key == "tableDeleteRow" || a.key == "tableSteppers" || a.key == "tableContextMenu" || a.key == "tableSelection" || a.key == "criticalPath" || a.key == "holidays" || a.key == "weekend" || a.key == "timezone" || a.key == "showNonWorkingDays" || a.key == "progressColumn" || a.key == "interactive" || a.key == "settingsButton" || a.key == "showZoom" || a.key == "showAdd" || a.key == "showHistory" || a.key == "stateMode" || a.key == "stateHints" || a.key == "onTaskChange" || a.key == "onDataChange" || a.key == "onCalendarChange" || a.key == "onResourceChange" || a.key == "onCommit" ||
			a.key == "onShapeClick" || a.key == "onShapeDoubleClick" || a.key == "onShapeMouseOver" || a.key == "onShapeMouseOut" || a.key == "onShapeContextMenu" ||
			a.key == "onLinkClick" || a.key == "onLinkDoubleClick" || a.key == "onLinkMouseOver" || a.key == "onLinkMouseOut" || a.key == "onLinkContextMenu" ||
			a.key == "onTagClick" || a.key == "onTagDoubleClick" || a.key == "onTagMouseOver" || a.key == "onTagMouseOut" || a.key == "onTagContextMenu" ||
			a.key == "showGrid" || a.key == "showLegend" || a.key == "showValues" ||
			a.key == "showResources" || a.key == "showProgress" ||
			a.key == "showSlack" || a.key == "showCosts" || a.key == "statusColors" ||
			a.key == "ganttScale" || a.key == "autoSchedule" || a.key == "linkColor" ||
			a.key == "linkWidth" || a.key == "linkGap" || a.key == "linkStyle" || a.key == "arrowType" || a.key == "arrowSize" || a.key == "barBorderColor" || a.key == "stateRouting" || a.key == "stateTagStyle" || a.key == "stateTagLink" || a.key == "stateLinkTagStyle" || a.key == "stateLinkTagLink" || a.key == "stateShapeTagStyle" || a.key == "stateShapeTagLink" || a.key == "stateTagFont" || a.key == "stateTagFontSize" || a.key == "stateShapeTagFont" || a.key == "stateShapeTagFontSize" || a.key == "stateLinkTagFont" || a.key == "stateLinkTagFontSize" || a.key == "stateRouteAlgorithm" || a.key == "stateLinkArcRadius" || a.key == "stateGrid" || a.key == "stateGridGap" || a.key == "stateGridColor") {
			continue
		}
		if a.key == "tooltip" {
			if !a.hasExpr {
				sb.WriteString(" data-sui-tooltip=\"")
				sb.WriteString(htmlEscape(partsContent(a.text)))
				sb.WriteString("\"")
			}
			continue
		}
		if a.key == "head" || a.key == "options" || a.key == "link" ||
			a.key == "dropFiles" || a.key == "pickFiles" || a.key == "pickMode" || a.key == "dragKey" || a.key == "dropKey" {
			continue
		}
		if a.key == "required" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
				sb.WriteString(" required")
			}
			continue
		}
		if a.key == "directory" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b {
				sb.WriteString(" webkitdirectory")
			}
			continue
		}
		if a.key == "label" {
			continue
		}
		sb.WriteString(" ")
		if a.key == "inputtype" {
			sb.WriteString("type")
		} else {
			sb.WriteString(a.key)
		}
		sb.WriteString("=\"")
		if a.hasExpr {
			if a.expr != nil && a.expr.kind == "int" {
				sb.WriteString(strconv.Itoa(a.expr.i))
			} else if a.expr != nil && a.expr.kind == "bool" {
				if a.expr.b {
					sb.WriteString("true")
				} else {
					sb.WriteString("false")
				}
			}
			// dynamic bindings left as-is in static output
		} else {
			sb.WriteString(htmlEscape(partsContent(a.text)))
		}
		sb.WriteString("\"")
	}
	if dlId != "" {
		sb.WriteString(" list=\"" + dlId + "\"")
	}
}

func staticText(sb *strings.Builder, parts []StrPart) {
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "text" {
			sb.WriteString(htmlEscape(parts[i].text))
		}
	}
}
