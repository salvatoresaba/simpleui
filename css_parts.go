package sui

// Base stylesheet slices. generateCss emits only the slices a program
// needs, in this exact order, so cascade and specificity are preserved.
// The text is the original monolithic stylesheet, split at component
// boundaries; suiCssBase is the concatenation of every slice.

const suiCssFoundation = `* { box-sizing: border-box; }

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

/* An open anchored overlay must not inherit a transform/filter animation:
   those properties make the fixed panel use the anchor as its containing
   block, which moves the menu far away from its trigger. */
.sui-overlay-anchor-open,
.sui-overlay-anchor-open:hover,
.sui-overlay-anchor-open:focus,
.sui-overlay-anchor-open:active {
  animation: none !important;
  transition: none !important;
  translate: none !important;
  scale: none !important;
  rotate: none !important;
  transform: none !important;
  perspective: none !important;
  filter: none !important;
  backdrop-filter: none !important;
  will-change: auto !important;
}

/* Text containment: whatever the control shape, the label stays inside. */
.sui-button, .sui-badge, .sui-alert, .sui-card, .sui-tab, .sui-menuitem,
.sui-dropdown-item, .sui-combobox-option, .sui-toast-message, .sui-tree-label,
.sui-timeline-title, .sui-timeline-content, .sui-accordion-head,
.sui-modal-title, .sui-modal-body, .sui-drawer-body, .sui-command,
.sui-step-label, .sui-carousel-slide, .sui-field-label {
  min-width: 0;
  max-width: 100%;
  overflow-wrap: anywhere;
  word-break: break-word;
}
.sui-button { overflow: hidden; text-align: center; line-height: 1.15; }
.sui-badge, .sui-tab { overflow: hidden; text-overflow: ellipsis; }

/* Form controls - modern minimal style */
.sui-input, .sui-textarea, .sui-select {
  appearance: none;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  background: var(--sui-field-bg, #1d2129);
  color: var(--sui-field-fg, var(--sui-fg, #e6e6ef));
  border: var(--sui-field-border-width, 1px) var(--sui-field-border-style, solid) var(--sui-field-border, var(--sui-border, #3a4152));
  border-radius: var(--sui-field-radius, var(--sui-radius, 8px));
  padding: var(--sui-field-padding, 8px 12px);
  font-family: var(--sui-field-font, inherit);
  font-size: var(--sui-field-font-size, 14px);
  letter-spacing: var(--sui-field-tracking, normal);
  box-shadow: var(--sui-field-shadow, none);
  text-overflow: ellipsis;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.sui-input::placeholder, .sui-textarea::placeholder {
  color: var(--sui-field-placeholder, #6b7280);
}
.sui-input:hover, .sui-textarea:hover, .sui-select:hover {
  border-color: var(--sui-field-hover, var(--sui-field-border, var(--sui-border, #3a4152)));
}
:where(.sui-input):focus, :where(.sui-textarea):focus, :where(.sui-select):focus {
  outline: none;
  border-color: var(--sui-field-focus, #4c5cf7);
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
.sui-number, .sui-color, .sui-date, .sui-time, .sui-file {
  appearance: none;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  background: var(--sui-field-bg, #1d2129);
  color: var(--sui-field-fg, var(--sui-fg, #e6e6ef));
  border: var(--sui-field-border-width, 1px) var(--sui-field-border-style, solid) var(--sui-field-border, var(--sui-border, #3a4152));
  border-radius: var(--sui-field-radius, var(--sui-radius, 8px));
  padding: var(--sui-field-padding, 8px 12px);
  font-family: var(--sui-field-font, inherit);
  font-size: var(--sui-field-font-size, 14px);
  letter-spacing: var(--sui-field-tracking, normal);
  box-shadow: var(--sui-field-shadow, none);
  text-overflow: ellipsis;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.sui-color { padding: 4px; min-height: 36px; }
.sui-file { padding: 6px 10px; }
.sui-file::file-selector-button { margin-right: 8px; }
.sui-number:hover, .sui-color:hover, .sui-date:hover, .sui-time:hover, .sui-file:hover {
  border-color: var(--sui-field-hover, var(--sui-field-border, var(--sui-border, #3a4152)));
}
.sui-number:focus, .sui-color:focus, .sui-date:focus, .sui-time:focus, .sui-file:focus {
  outline: none;
  border-color: var(--sui-field-focus, #4c5cf7);
  box-shadow: 0 0 0 3px rgba(76, 92, 247, 0.25);
}
`

const suiCssCombobox = `.sui-combobox { display: flex; flex-direction: column; gap: 4px; position: relative; min-width: min(180px, 100%); max-width: 100%; }
.sui-combobox-label { font-size: 12px; color: var(--sui-combobox-label, var(--sui-muted, #8fa0b8)); }
.sui-combobox-control { position: relative; display: flex; align-items: center; min-width: 0; }
.sui-combobox-input {
  appearance: none;
  width: 100%;
  min-width: 0;
  background: var(--sui-combobox-bg, #1d2129);
  color: var(--sui-combobox-fg, var(--sui-fg, #e6e6ef));
  border: var(--sui-combobox-border-width, 1px) var(--sui-combobox-border-style, solid) var(--sui-combobox-border, var(--sui-border, #3a4152));
  border-radius: var(--sui-combobox-radius, 8px);
  padding: var(--sui-combobox-padding, 8px 12px);
  padding-right: var(--sui-combobox-padding-right, calc(2.6em + 26px));
  font-family: var(--sui-combobox-font, inherit);
  font-size: var(--sui-combobox-font-size, 14px);
  letter-spacing: var(--sui-combobox-tracking, normal);
  box-shadow: var(--sui-combobox-shadow, none);
  text-overflow: ellipsis;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.sui-combobox-input::placeholder { color: var(--sui-combobox-placeholder, #6b7280); }
.sui-combobox-input:focus { outline: none; border-color: var(--sui-combobox-focus, #4c5cf7); box-shadow: 0 0 0 3px rgba(76,92,247,.25); }
.sui-combobox-input:disabled { opacity: .5; cursor: not-allowed; }
.sui-combobox-toggle,
.sui-combobox-clear {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  border: 0;
  background: transparent;
  color: var(--sui-combobox-muted, var(--sui-muted, #8fa0b8));
  font: inherit;
  cursor: pointer;
  padding: 2px 5px;
  line-height: 1;
}
.sui-combobox-toggle { right: 5px; font-size: 13px; }
.sui-combobox-clear { right: 28px; font-size: 16px; }
.sui-combobox-toggle:hover, .sui-combobox-clear:hover { color: var(--sui-combobox-fg, var(--sui-fg, #e6e6ef)); }
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
  background: var(--sui-combobox-popup-bg, var(--sui-combobox-bg, #1d2129));
  color: var(--sui-combobox-fg, var(--sui-fg, #e6e6ef));
  border: var(--sui-combobox-border-width, 1px) var(--sui-combobox-border-style, solid) var(--sui-combobox-border, var(--sui-border, #3a4152));
  border-radius: var(--sui-combobox-radius, 8px);
  box-shadow: var(--sui-ac-shadow, 0 8px 24px rgba(0,0,0,.45));
  box-sizing: border-box;
}
.sui-combobox-popup[hidden] { display: none; }
.sui-combobox-option { padding: 7px 9px; border-radius: 5px; font-size: 13px; cursor: pointer; min-width: 0; overflow-wrap: anywhere; white-space: normal; }
.sui-combobox-option:hover { background: var(--sui-combobox-hover, var(--sui-hover, #2a3143)); color: var(--sui-combobox-hover-fg, var(--sui-combobox-fg, var(--sui-fg, #e6e6ef))); }
.sui-combobox-option.active { background: var(--sui-combobox-hover, var(--sui-hover, #2a3143)); color: var(--sui-combobox-hover-fg, var(--sui-combobox-fg, var(--sui-fg, #e6e6ef))); outline: 2px solid var(--sui-combobox-focus, #4c5cf7); outline-offset: -2px; }
.sui-combobox-option[aria-selected="true"] { background: var(--sui-combobox-selected-bg, rgba(76,92,247,.18)); color: var(--sui-combobox-selected-fg, var(--sui-combobox-fg, var(--sui-fg, #e6e6ef))); font-weight: 600; box-shadow: inset 3px 0 0 var(--sui-combobox-selected-accent, var(--sui-accent, #4c5cf7)); }
.sui-combobox-empty { padding: 8px 9px; font-size: 12px; color: var(--sui-combobox-muted, var(--sui-muted, #8fa0b8)); }
.sui-combobox-error { font-size: 12px; color: #f87171; }
.sui-combobox-error[hidden] { display: none; }
`

const suiCssDatepicker = `.sui-datepicker { display: flex; flex-direction: column; gap: 4px; position: relative; min-width: min(180px, 100%); max-width: 100%; }
.sui-datepicker-label { font-size: 12px; color: var(--sui-datepicker-label, var(--sui-muted, #8fa0b8)); }
.sui-datepicker-control { position: relative; display: flex; align-items: center; min-width: 0; }
.sui-datepicker-input {
  appearance: none;
  width: 100%;
  min-width: 0;
  background: var(--sui-datepicker-bg, #1d2129);
  color: var(--sui-datepicker-fg, var(--sui-fg, #e6e6ef));
  border: var(--sui-datepicker-border-width, 1px) var(--sui-datepicker-border-style, solid) var(--sui-datepicker-border, var(--sui-border, #3a4152));
  border-radius: var(--sui-datepicker-radius, 8px);
  padding: var(--sui-datepicker-padding, 8px 12px);
  padding-right: var(--sui-datepicker-padding-right, 36px);
  font-family: var(--sui-datepicker-font, inherit);
  font-size: var(--sui-datepicker-font-size, 14px);
  letter-spacing: var(--sui-datepicker-tracking, normal);
  box-shadow: var(--sui-datepicker-shadow, none);
  color-scheme: dark;
  text-overflow: ellipsis;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.sui-datepicker-input:focus { outline: none; border-color: var(--sui-datepicker-focus, #4c5cf7); box-shadow: 0 0 0 3px rgba(76,92,247,.25); }
.sui-datepicker-input:disabled { opacity: .5; cursor: not-allowed; }
.sui-datepicker-clear {
  position: absolute;
  top: 50%;
  right: 5px;
  transform: translateY(-50%);
  border: 0;
  background: transparent;
  color: var(--sui-datepicker-muted, var(--sui-muted, #8fa0b8));
  font: inherit;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  padding: 2px 5px;
}
.sui-datepicker-clear:hover { color: var(--sui-datepicker-fg, var(--sui-fg, #e6e6ef)); }
.sui-datepicker-value { display: none; }
.sui-datepicker-error { font-size: 12px; color: #f87171; }
.sui-datepicker-error[hidden] { display: none; }
.sui-datepicker[data-sui-calendar-open="1"] { z-index: 4100; }
.sui-datepicker-open {
  position: absolute;
  right: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 28px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--sui-datepicker-muted, var(--sui-muted, #8fa0b8));
  font: inherit;
  cursor: pointer;
}
.sui-datepicker-open:hover { color: var(--sui-datepicker-fg, var(--sui-fg, #e6e6ef)); background: var(--sui-datepicker-hover, rgba(148,163,184,.12)); }
.sui-datepicker-input[aria-haspopup="dialog"] { padding-right: 64px; cursor: text; }
.sui-datepicker-calendar {
  position: absolute;
  z-index: 4000;
  top: calc(100% + 6px);
  left: 0;
  width: min(320px, calc(100vw - 32px));
  padding: 12px;
  border: 1px solid var(--sui-datepicker-border, var(--sui-border, #3a4152));
  border-radius: var(--sui-datepicker-calendar-radius, 12px);
  background: var(--sui-datepicker-calendar-bg, var(--sui-datepicker-bg, var(--sui-panel, #1d2129)));
  color: var(--sui-datepicker-fg, var(--sui-fg, #e6e6ef));
  box-shadow: var(--sui-datepicker-calendar-shadow, 0 12px 32px rgba(0,0,0,.36));
}
.sui-datepicker-calendar[hidden] { display: none; }
.sui-datepicker-calendar-header { display: grid; grid-template-columns: 36px 1fr 36px; align-items: center; gap: 6px; margin-bottom: 10px; }
.sui-datepicker-month-title { text-align: center; font-size: 13px; font-weight: 650; }
.sui-datepicker-month-previous, .sui-datepicker-month-next {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 30px;
  border: 1px solid transparent;
  border-radius: 7px;
  background: transparent;
  color: inherit;
  font-family: inherit;
  font-size: 22px;
  line-height: 1;
  cursor: pointer;
}
.sui-datepicker-month-previous:hover, .sui-datepicker-month-next:hover { background: var(--sui-datepicker-hover, rgba(148,163,184,.12)); }
.sui-datepicker-weekdays, .sui-datepicker-days { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 3px; }
.sui-datepicker-weekdays { margin-bottom: 4px; color: var(--sui-datepicker-muted, var(--sui-muted, #8fa0b8)); text-align: center; font-size: 10px; font-weight: 600; }
.sui-datepicker-weekdays span { padding: 4px 0; }
.sui-datepicker-day {
  min-width: 0;
  height: 34px;
  border: 1px solid transparent;
  border-radius: var(--sui-datepicker-day-radius, 8px);
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: 12px;
  cursor: pointer;
}
.sui-datepicker-day:hover:not(:disabled) { background: var(--sui-datepicker-hover, rgba(148,163,184,.14)); }
.sui-datepicker-day.is-outside-month { color: var(--sui-datepicker-muted, var(--sui-muted, #8fa0b8)); opacity: .58; }
.sui-datepicker-day.is-today { border-color: var(--sui-datepicker-today, #818cf8); }
.sui-datepicker-day.is-selected { background: var(--sui-datepicker-selected, #4c5cf7); color: var(--sui-datepicker-selected-fg, #fff); font-weight: 700; }
.sui-datepicker-day:focus-visible, .sui-datepicker-month-previous:focus-visible, .sui-datepicker-month-next:focus-visible, .sui-datepicker-open:focus-visible { outline: 2px solid var(--sui-datepicker-focus, #818cf8); outline-offset: 2px; }
.sui-datepicker-day:disabled { opacity: .32; cursor: not-allowed; }
.sui-datepicker-time-wrap { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: 10px; color: var(--sui-datepicker-muted, var(--sui-muted, #8fa0b8)); font-size: 12px; }
.sui-datepicker-time { width: 112px; padding: 6px 8px; border: 1px solid var(--sui-datepicker-border, var(--sui-border, #3a4152)); border-radius: 7px; background: var(--sui-datepicker-bg, #1d2129); color: var(--sui-datepicker-fg, var(--sui-fg, #e6e6ef)); color-scheme: inherit; }
@media (prefers-reduced-motion: reduce) { .sui-datepicker-calendar *, .sui-datepicker-open { scroll-behavior: auto !important; transition-duration: .01ms !important; } }
`

const suiCssTree = `.sui-tree {
  display: flex;
  flex-direction: column;
  gap: var(--sui-tree-gap, 2px);
  color: var(--sui-fg, #e6e6ef);
  font-size: var(--sui-tree-font, 13px);
  font-family: var(--sui-tree-font-family, inherit);
  font-weight: var(--sui-tree-font-weight, 400);
  letter-spacing: var(--sui-tree-tracking, normal);
  background: var(--sui-tree-panel-bg, transparent);
  border: var(--sui-tree-panel-border, 0);
  border-radius: var(--sui-tree-panel-radius, 0);
  padding: var(--sui-tree-panel-padding, 0);
}
.sui-tree-title { font-size: var(--sui-tree-title-size, 13px); color: var(--sui-tree-title-color, #e6e6ef); font-weight: 600; padding: 2px 4px 6px; }
.sui-tree-group { list-style: none; margin: 0; padding-left: var(--sui-tree-indent, 16px); }
.sui-tree > .sui-tree-group { padding-left: 0; }
.sui-tree-item { margin: 1px 0; }
.sui-tree-row { display: flex; align-items: center; gap: 4px; min-height: var(--sui-tree-row-height, 28px); border-radius: var(--sui-tree-row-radius, 5px); padding: var(--sui-tree-row-padding, 2px 5px); background: var(--sui-tree-row-bg, transparent); cursor: pointer; }
.sui-tree-row:hover { background: var(--sui-tree-hover-bg, var(--sui-hover, #2a3143)); }
.sui-tree-item.selected > .sui-tree-row,
.sui-tree-item.sui-tree-pending > .sui-tree-row { background: var(--sui-tree-selected-bg, rgba(76, 92, 247, .28)); color: var(--sui-tree-selected-fg, inherit); }
.sui-tree-item:focus { outline: none; }
.sui-tree-item:focus-visible > .sui-tree-row { outline: 2px solid #4c5cf7; outline-offset: -2px; }
.sui-tree-toggle, .sui-tree-spacer { height: 20px; flex: none; display: inline-flex; align-items: center; justify-content: center; }
.sui-tree-spacer { width: 20px; }
.sui-tree-toggle { min-width: 20px; border: 0; background: transparent; color: var(--sui-muted, #8fa0b8); padding: 0 1px; cursor: pointer; font: inherit; }
.sui-tree-toggle:hover { color: var(--sui-fg, #e6e6ef); }
.sui-tree-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
.sui-tree[data-sui-tree-truncate="0"] .sui-tree-label { white-space: normal; overflow: visible; text-overflow: clip; }
.sui-tree[data-sui-tree-guides="1"] .sui-tree-group { padding-left: 0; }
.sui-tree-guide {
  flex: none;
  align-self: stretch;
  width: max(1px, calc(var(--sui-tree-indent, 16px) - 4px));
  border-left: 1px solid var(--sui-tree-guide-color, rgba(143, 160, 184, .30));
  pointer-events: none;
}
.sui-tree-icon {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.1em;
  color: var(--sui-tree-icon-color, inherit);
}
.sui-tree-badge {
  flex: none;
  border-radius: 999px;
  padding: 0 6px;
  font-size: 10px;
  font-weight: 600;
  line-height: 16px;
  white-space: nowrap;
  background: var(--sui-tree-badge-bg, rgba(148, 163, 184, .18));
  color: var(--sui-tree-badge-fg, #cbd5e1);
}
.sui-tree-badge-info { background: rgba(56, 189, 248, .20); color: #7dd3fc; }
.sui-tree-badge-success { background: rgba(34, 197, 94, .20); color: #86efac; }
.sui-tree-badge-warning { background: rgba(245, 158, 11, .22); color: #fcd34d; }
.sui-tree-badge-danger { background: rgba(239, 68, 68, .22); color: #fca5a5; }
.sui-tree-badge-count {
  background: var(--sui-tree-count-bg, rgba(148, 163, 184, .14));
  color: var(--sui-tree-count-fg, var(--sui-tree-meta-color, #94a3b8));
  border: var(--sui-tree-count-border-width, 0) solid var(--sui-tree-count-border, transparent);
  border-radius: var(--sui-tree-count-radius, 999px);
  padding: var(--sui-tree-count-padding, 0 6px);
  min-width: var(--sui-tree-count-min-width, 18px);
  font-family: var(--sui-tree-count-font, inherit);
  font-size: var(--sui-tree-count-size, 10px);
  font-weight: var(--sui-tree-count-weight, 600);
}
.sui-tree-meta {
  flex: none;
  margin-left: auto;
  padding-left: 6px;
  color: var(--sui-tree-meta-color, #94a3b8);
  font-size: 11px;
  white-space: nowrap;
}
.sui-tree-check {
  flex: none;
  width: 14px;
  height: 14px;
  border: 1.5px solid var(--sui-tree-check-border, #64748b);
  border-radius: 3px;
  background: var(--sui-tree-check-bg, transparent);
  position: relative;
  cursor: pointer;
}
.sui-tree-check[aria-checked="true"],
.sui-tree-check[aria-checked="mixed"] {
  background: var(--sui-tree-check-on, #4c5cf7);
  border-color: var(--sui-tree-check-on, #4c5cf7);
}
.sui-tree-check[aria-checked="true"]::after {
  content: "";
  position: absolute;
  left: 4px;
  top: 1px;
  width: 3px;
  height: 7px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}
.sui-tree-check[aria-checked="mixed"]::after {
  content: "";
  position: absolute;
  left: 2px;
  top: 5px;
  width: 7px;
  height: 2px;
  border-radius: 1px;
  background: #fff;
}
.sui-tree-check:focus-visible { outline: 2px solid #93c5fd; outline-offset: 2px; }
.sui-tree-search { display: flex; align-items: center; gap: 8px; padding: 0 0 6px; }
.sui-tree-search-input {
  flex: 1;
  min-width: 0;
  background: var(--input-bg, #1d2129);
  color: var(--sui-fg, #e6e6ef);
  border: 1px solid var(--sui-tree-border, #3d4556);
  border-radius: var(--sui-tree-radius, 5px);
  padding: 6px 8px;
  font: inherit;
}
.sui-tree-search-input:focus { outline: 2px solid #4c5cf7; outline-offset: 1px; }
.sui-tree-search-count { flex: none; color: var(--sui-tree-meta-color, #94a3b8); font-size: 11px; }
.sui-tree-empty { padding: 16px 8px; color: var(--sui-tree-meta-color, #94a3b8); font-size: 12px; text-align: center; }
.sui-tree-empty[hidden] { display: none; }
.sui-tree-match { background: rgba(245, 158, 11, .28); color: inherit; border-radius: 2px; }
.sui-tree-tooltip {
  position: fixed;
  z-index: 1001;
  max-width: var(--sui-tree-tooltip-max-width, 260px);
  background: var(--sui-tree-tooltip-bg, #111827);
  color: var(--sui-tree-tooltip-fg, #f8fafc);
  border: 1px solid var(--sui-tree-tooltip-border, #475569);
  border-radius: var(--sui-tree-tooltip-radius, 6px);
  padding: var(--sui-tree-tooltip-padding, 6px 8px);
  font-size: var(--sui-tree-tooltip-font-size, 11px);
  line-height: 1.35;
  pointer-events: none;
  opacity: 0;
  transition: opacity .12s ease;
  white-space: normal;
  overflow-wrap: anywhere;
}
.sui-tree-tooltip.on { opacity: 1; }
.sui-tree-item.sui-tree-drop-before > .sui-tree-row { box-shadow: inset 0 2px 0 #4c5cf7; }
.sui-tree-item.sui-tree-drop-after > .sui-tree-row { box-shadow: inset 0 -2px 0 #4c5cf7; }
.sui-tree-item.sui-tree-drop-inside > .sui-tree-row { outline: 2px solid #4c5cf7; outline-offset: -2px; }
.sui-tree-item.sui-tree-drop-invalid > .sui-tree-row { outline: 2px dashed #f87171; outline-offset: -2px; }
.sui-tree-item.sui-tree-dragging > .sui-tree-row { opacity: .45; }
.sui-tree-item[data-sui-lazy="1"] > .sui-tree-row .sui-tree-toggle { color: var(--sui-tree-toggle-color, #94a3b8); }
.sui-tree-loading { margin-left: auto; color: var(--sui-tree-meta-color, #94a3b8); font-size: 11px; }
.sui-tree-error { margin-left: auto; color: #f87171; font-size: 11px; }
.sui-tree-retry {
  border: 1px solid currentColor;
  background: transparent;
  color: #f87171;
  border-radius: 4px;
  font: inherit;
  font-size: 10px;
  padding: 1px 5px;
  cursor: pointer;
}
.sui-tree-retry:hover { background: rgba(248, 113, 113, .15); }
.sui-tree-context {
  position: fixed;
  z-index: 1000;
  min-width: 170px;
  padding: 4px;
  background: var(--sui-panel, #1d2129);
  color: var(--sui-fg, #e6e6ef);
  border: 1px solid var(--sui-border, #3d4556);
  border-radius: 6px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, .45);
}
.sui-tree-context-item {
  display: block;
  width: 100%;
  text-align: left;
  background: transparent;
  color: inherit;
  border: 0;
  border-radius: 4px;
  padding: 6px 8px;
  font: inherit;
  font-size: 12px;
  cursor: pointer;
}
.sui-tree-context-item:hover,
.sui-tree-context-item:focus { background: rgba(76, 92, 247, .28); outline: none; }
.sui-tree-context-item:disabled { opacity: .45; cursor: default; }
.sui-tree-grid-head {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: var(--sui-tree-row-padding, 2px 5px);
  color: var(--sui-tree-meta-color, #94a3b8);
  border-bottom: 1px solid var(--sui-tree-guide-color, rgba(143, 160, 184, .30));
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: .04em;
}
.sui-tree-grid-name { flex: 1; min-width: 0; }
.sui-tree-cell {
  flex: none;
  width: 104px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--sui-tree-meta-color, #94a3b8);
  font-size: 11px;
}
.sui-tree-cell-size { font-variant-numeric: tabular-nums; }
.sui-tree[data-sui-tree-columns] .sui-tree-meta { margin-left: 0; }
.sui-tree[data-sui-tree-columns] .sui-tree-label { flex: 1; }
.sui-tree[data-sui-tree-virtual="1"] .sui-tree-item {
  content-visibility: auto;
  contain-intrinsic-size: auto var(--sui-tree-row-height, 28px);
}
.sui-tree-rename-btn {
  border: var(--sui-tree-rename-border-width, 0) solid var(--sui-tree-rename-border, transparent);
  background: var(--sui-tree-rename-bg, transparent);
  color: var(--sui-tree-rename-fg, var(--sui-tree-rename-color, var(--sui-muted, #8fa0b8)));
  font-family: var(--sui-tree-rename-font, inherit);
  font-size: var(--sui-tree-rename-size, 13px);
  line-height: 1;
  border-radius: var(--sui-tree-rename-radius, 4px);
  cursor: pointer;
  padding: var(--sui-tree-rename-padding, 1px 4px);
  opacity: .45;
}
.sui-tree-row:hover .sui-tree-rename-btn,
.sui-tree-rename-btn:focus { opacity: 1; }
.sui-tree-rename-btn:hover { background: var(--sui-tree-rename-hover-bg, transparent); color: var(--sui-tree-rename-hover, #a5b4fc); }
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
`

const suiCssChart = `.sui-chart {
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
  --sui-chart-control-border: #3d4556;
  --sui-chart-control-fg: var(--sui-chart-fg, #e6e6e6);
  --sui-chart-control-hover-bg: #2d3340;
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
.sui-chart-table td.sui-chart-table-cell-multi { background: var(--sui-chart-table-selection-bg, var(--sui-chart-table-selected-row-bg, rgba(56,189,248,.22))) !important; }
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
.sui-chart-table td[data-col="0"].sui-chart-table-cell-multi { background: var(--sui-chart-table-selection-bg, var(--sui-chart-table-selected-row-bg, rgba(56,189,248,.22))) !important; }
.sui-chart-table table.sui-chart-table-mode-cell tbody tr.sui-chart-table-row-selected td { background: transparent !important; }
.sui-chart-table table.sui-chart-table-mode-cell tbody tr.sui-chart-table-row-selected td.sui-chart-table-cell-multi { background: var(--sui-chart-table-selection-bg, var(--sui-chart-table-selected-row-bg, rgba(56,189,248,.22))) !important; }
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
  text-align: var(--sui-chart-tooltip-text-align, left);
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
`

const suiCssControls = `.sui-checkbox-wrap, .sui-radio-wrap, .sui-switch-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.sui-checkbox {
  --sui-cb-size:18px;
  --sui-cb-radius:4px;
  --sui-cb-off-bg:#1d2129;
  --sui-cb-on-bg:#4c5cf7;
  --sui-cb-color:var(--sui-cb-off-bg);
  --sui-cb-border-width:1px;
  --sui-cb-border-color:var(--sui-border);
  --sui-cb-mark-color:#fff;
  --sui-cb-mark-w:5px;
  --sui-cb-mark-h:9px;
  --sui-cb-mark-opacity:0;
  --sui-cb-mark-rotate:45deg;
  appearance:none;
  width:var(--sui-cb-size);
  height:var(--sui-cb-size);
  border:var(--sui-cb-border-width) solid var(--sui-cb-border-color);
  background:var(--sui-cb-color);
  border-radius:var(--sui-cb-radius);
  position:relative;
  cursor:pointer;
  transition:all 0.15s ease;
}
.sui-checkbox:checked {
  --sui-cb-color:var(--sui-cb-on-bg);
  --sui-cb-border-color:var(--sui-cb-on-bg);
  --sui-cb-mark-opacity:1;
}
.sui-checkbox::after {
  content:"";
  position:absolute;
  left:50%;
  top:50%;
  width:var(--sui-cb-mark-w);
  height:var(--sui-cb-mark-h);
  border:solid var(--sui-cb-mark-color);
  border-width:0 2px 2px 0;
  transform:translate(-50%,-50%) rotate(var(--sui-cb-mark-rotate));
  opacity:var(--sui-cb-mark-opacity);
}
.sui-radio {
  --sui-rd-size:18px;
  --sui-rd-radius:50%;
  --sui-rd-off-bg:#1d2129;
  --sui-rd-on-bg:#4c5cf7;
  --sui-rd-color:var(--sui-rd-off-bg);
  --sui-rd-border-width:1px;
  --sui-rd-border-color:var(--sui-border);
  --sui-rd-dot-size:6px;
  --sui-rd-dot-color:#fff;
  --sui-rd-dot-opacity:0;
  appearance:none;
  width:var(--sui-rd-size);
  height:var(--sui-rd-size);
  border:var(--sui-rd-border-width) solid var(--sui-rd-border-color);
  background:var(--sui-rd-color);
  border-radius:var(--sui-rd-radius,50%);
  position:relative;
  cursor:pointer;
  transition:all 0.15s ease;
}
.sui-radio:checked {
  --sui-rd-color:var(--sui-rd-on-bg);
  --sui-rd-border-color:var(--sui-rd-on-bg);
  --sui-rd-dot-opacity:1;
}
.sui-radio::after {
  content:"";
  position:absolute;
  left:50%;
  top:50%;
  width:var(--sui-rd-dot-size);
  height:var(--sui-rd-dot-size);
  border-radius:50%;
  background:var(--sui-rd-dot-color);
  transform:translate(-50%,-50%);
  opacity:var(--sui-rd-dot-opacity);
}
.sui-switch {
  --sui-sw-w:42px;
  --sui-sw-h:24px;
  --sui-sw-radius:999px;
  --sui-sw-off-bg:#374151;
  --sui-sw-on-bg:#4c5cf7;
  --sui-sw-color:var(--sui-sw-off-bg);
  --sui-sw-thumb-size:20px;
  --sui-sw-thumb-radius:10px;
  --sui-sw-thumb-off-bg:#fff;
  --sui-sw-thumb-on-bg:#fff;
  --sui-sw-thumb-color:var(--sui-sw-thumb-off-bg);
  --sui-sw-thumb-x-off:2px;
  --sui-sw-thumb-x-on:20px;
  --sui-sw-thumb-x:var(--sui-sw-thumb-x-off);
  --sui-sw-thumb-y:2px;
  --sui-sw-border-width:0px;
  --sui-sw-border-color:transparent;
  appearance:none;
  width:var(--sui-sw-w);
  height:var(--sui-sw-h);
  background:var(--sui-sw-color);
  border:var(--sui-sw-border-width) solid var(--sui-sw-border-color);
  border-radius:var(--sui-sw-radius);
  position:relative;
  cursor:pointer;
  transition:background 0.2s ease;
}
.sui-switch::after {
  content:"";
  position:absolute;
  left:0;
  top:0;
  width:var(--sui-sw-thumb-size);
  height:var(--sui-sw-thumb-size);
  background:var(--sui-sw-thumb-color);
  border-radius:var(--sui-sw-thumb-radius);
  transform:translate(var(--sui-sw-thumb-x), var(--sui-sw-thumb-y));
  transition:transform .22s cubic-bezier(.2,.8,.2,1), width .22s ease, height .22s ease, border-radius .22s ease, background-color .22s ease;
  will-change:transform, width, height, background-color;
}
.sui-switch:checked {
  --sui-sw-color:var(--sui-sw-on-bg);
  --sui-sw-thumb-color:var(--sui-sw-thumb-on-bg);
  --sui-sw-thumb-x:var(--sui-sw-thumb-x-on);
}
[data-sui-motion], [data-sui-motion]::before, [data-sui-motion]::after { transition:none !important; }
[data-sui-motion] { will-change:transform, opacity, background-color, border-color, border-radius; }
.sui-motion-exit { pointer-events:none !important; }
.sui-motion-shared-clone { pointer-events:none !important; }
.sui-switch.sui-sw-shape-rounded { --sui-sw-w:44px; --sui-sw-h:26px; --sui-sw-radius:10px; --sui-sw-thumb-size:18px; --sui-sw-thumb-radius:6px; --sui-sw-thumb-x-off:4px; --sui-sw-thumb-x-on:22px; --sui-sw-thumb-y:4px; }
.sui-switch.sui-sw-shape-square { --sui-sw-w:42px; --sui-sw-h:24px; --sui-sw-radius:3px; --sui-sw-thumb-size:16px; --sui-sw-thumb-radius:3px; --sui-sw-thumb-x-off:4px; --sui-sw-thumb-x-on:22px; --sui-sw-thumb-y:4px; }
.sui-switch.sui-sw-shape-pill { --sui-sw-w:54px; --sui-sw-h:24px; --sui-sw-radius:999px; --sui-sw-thumb-size:16px; --sui-sw-thumb-radius:8px; --sui-sw-thumb-x-off:4px; --sui-sw-thumb-x-on:34px; --sui-sw-thumb-y:4px; }
.sui-switch.sui-sw-shape-ios { --sui-sw-w:51px; --sui-sw-h:31px; --sui-sw-radius:999px; --sui-sw-thumb-size:27px; --sui-sw-thumb-radius:14px; --sui-sw-thumb-x-off:2px; --sui-sw-thumb-x-on:22px; --sui-sw-thumb-y:2px; }
.sui-switch.sui-sw-shape-material { --sui-sw-w:36px; --sui-sw-h:14px; --sui-sw-radius:999px; --sui-sw-thumb-size:20px; --sui-sw-thumb-x-off:-2px; --sui-sw-thumb-x-on:18px; --sui-sw-thumb-y:-3px; }
.sui-switch.sui-sw-shape-minimal { --sui-sw-w:38px; --sui-sw-h:20px; --sui-sw-radius:4px; --sui-sw-thumb-size:14px; --sui-sw-thumb-radius:7px; --sui-sw-thumb-x-off:3px; --sui-sw-thumb-x-on:21px; --sui-sw-thumb-y:3px; }
.sui-switch.sui-sw-shape-soft { --sui-sw-w:46px; --sui-sw-h:26px; --sui-sw-radius:13px; --sui-sw-thumb-size:18px; --sui-sw-thumb-radius:9px; --sui-sw-thumb-x-off:4px; --sui-sw-thumb-x-on:24px; --sui-sw-thumb-y:4px; }
.sui-switch.sui-sw-theme-success { --sui-sw-on-bg:#16a34a; }
.sui-switch.sui-sw-theme-warning { --sui-sw-on-bg:#d97706; }
.sui-switch.sui-sw-theme-danger { --sui-sw-on-bg:#dc2626; }
.sui-switch.sui-sw-theme-mono { --sui-sw-off-bg:#4b5563; --sui-sw-on-bg:#e5e7eb; --sui-sw-thumb-off-bg:#9ca3af; --sui-sw-thumb-on-bg:#111827; }
.sui-switch.sui-sw-theme-dark { --sui-sw-off-bg:#1f2937; --sui-sw-on-bg:#64748b; --sui-sw-thumb-off-bg:#e5e7eb; --sui-sw-thumb-on-bg:#f8fafc; }
.sui-switch.sui-sw-theme-light { --sui-sw-off-bg:#cbd5e1; --sui-sw-on-bg:#2563eb; --sui-sw-thumb-off-bg:#fff; --sui-sw-thumb-on-bg:#fff; }
.sui-checkbox.sui-cb-shape-rounded { --sui-cb-radius:6px; }
.sui-checkbox.sui-cb-shape-square { --sui-cb-radius:2px; }
.sui-checkbox.sui-cb-shape-circle { --sui-cb-radius:9px; --sui-cb-mark-w:5px; --sui-cb-mark-h:8px; }
.sui-checkbox.sui-cb-shape-ios { --sui-cb-size:22px; --sui-cb-radius:11px; --sui-cb-mark-w:5px; --sui-cb-mark-h:9px; }
.sui-checkbox.sui-cb-shape-minimal { --sui-cb-size:16px; --sui-cb-radius:3px; --sui-cb-mark-w:4px; --sui-cb-mark-h:7px; }
.sui-checkbox.sui-cb-shape-soft { --sui-cb-size:20px; --sui-cb-radius:6px; --sui-cb-mark-w:5px; --sui-cb-mark-h:9px; }
.sui-checkbox.sui-cb-symbol-plus:checked::after, .sui-checkbox.sui-cb-symbol-cross:checked::after, .sui-checkbox.sui-cb-symbol-star:checked::after, .sui-checkbox.sui-cb-symbol-heart:checked::after { display:grid; place-items:center; left:0; top:0; width:100%; height:100%; border:0; transform:none; opacity:1; color:var(--sui-cb-mark-color,#fff); font:700 14px/1 system-ui,sans-serif; }
.sui-checkbox.sui-cb-symbol-plus:checked::after { content:"+"; }
.sui-checkbox.sui-cb-symbol-cross:checked::after { content:"×"; font-size:16px; }
.sui-checkbox.sui-cb-symbol-star:checked::after { content:"★"; font-size:12px; }
.sui-checkbox.sui-cb-symbol-heart:checked::after { content:"♥"; font-size:12px; }
.sui-checkbox.sui-cb-theme-success { --sui-cb-on-bg:#16a34a; }
.sui-checkbox.sui-cb-theme-warning { --sui-cb-on-bg:#d97706; }
.sui-checkbox.sui-cb-theme-danger { --sui-cb-on-bg:#dc2626; }
.sui-checkbox.sui-cb-theme-mono { --sui-cb-off-bg:#4b5563; --sui-cb-on-bg:#e5e7eb; --sui-cb-mark-color:#111827; }
.sui-checkbox.sui-cb-theme-dark { --sui-cb-off-bg:#1f2937; --sui-cb-on-bg:#64748b; }
.sui-checkbox.sui-cb-theme-light { --sui-cb-off-bg:#e5e7eb; --sui-cb-on-bg:#2563eb; --sui-cb-mark-color:#fff; }
.sui-radio.sui-rd-shape-compact { --sui-rd-size:14px; --sui-rd-dot-size:5px; }
.sui-radio.sui-rd-shape-large { --sui-rd-size:22px; --sui-rd-dot-size:8px; }
.sui-radio.sui-rd-shape-soft { --sui-rd-size:20px; --sui-rd-dot-size:7px; }
.sui-radio.sui-rd-shape-minimal { --sui-rd-size:16px; --sui-rd-dot-size:5px; --sui-rd-border-width:2px; }
.sui-radio.sui-rd-shape-check:checked::after { width:5px; height:9px; border:solid var(--sui-rd-dot-color,#fff); border-width:0 2px 2px 0; border-radius:0; background:transparent; transform:translate(-50%,-50%) rotate(45deg); }
.sui-radio.sui-rd-theme-success { --sui-rd-on-bg:#16a34a; }
.sui-radio.sui-rd-theme-warning { --sui-rd-on-bg:#d97706; }
.sui-radio.sui-rd-theme-danger { --sui-rd-on-bg:#dc2626; }
.sui-radio.sui-rd-theme-mono { --sui-rd-off-bg:#4b5563; --sui-rd-on-bg:#e5e7eb; --sui-rd-dot-color:#111827; }
.sui-radio.sui-rd-theme-dark { --sui-rd-off-bg:#1f2937; --sui-rd-on-bg:#64748b; }
.sui-radio.sui-rd-theme-light { --sui-rd-off-bg:#e5e7eb; --sui-rd-on-bg:#2563eb; }
.sui-slider {
  appearance: none;
  width: 100%;
  height: var(--sui-slider-track, 6px);
  background: var(--sui-slider-bg, #374151);
  border: var(--sui-slider-border-width, 0) solid var(--sui-slider-border, transparent);
  border-radius: var(--sui-slider-radius, 999px);
  box-shadow: var(--sui-slider-shadow, none);
  outline: none;
}
.sui-slider::-webkit-slider-thumb {
  appearance: none;
  width: var(--sui-slider-thumb-width, var(--sui-slider-thumb, 18px));
  height: var(--sui-slider-thumb-height, var(--sui-slider-thumb, 18px));
  background: var(--sui-slider-thumb-bg, #4c5cf7);
  border: var(--sui-slider-thumb-border-width, 0) solid var(--sui-slider-thumb-border, transparent);
  border-radius: var(--sui-slider-thumb-radius, 50%);
  cursor: pointer;
  box-shadow: var(--sui-slider-thumb-shadow, 0 2px 6px rgba(76, 92, 247, 0.4));
}
.sui-slider::-moz-range-thumb {
  width: var(--sui-slider-thumb-width, var(--sui-slider-thumb, 18px));
  height: var(--sui-slider-thumb-height, var(--sui-slider-thumb, 18px));
  background: var(--sui-slider-thumb-bg, #4c5cf7);
  border: var(--sui-slider-thumb-border-width, 0) solid var(--sui-slider-thumb-border, transparent);
  border-radius: var(--sui-slider-thumb-radius, 50%);
  cursor: pointer;
  box-shadow: var(--sui-slider-thumb-shadow, 0 2px 6px rgba(76, 92, 247, 0.4));
}
.sui-slider::-webkit-slider-runnable-track { height:var(--sui-slider-track,6px); border-radius:var(--sui-slider-radius,999px); background:var(--sui-slider-bg,#374151); }
.sui-slider::-moz-range-track { height:var(--sui-slider-track,6px); border-radius:var(--sui-slider-radius,999px); background:var(--sui-slider-bg,#374151); }
.sui-slider::-moz-range-progress { height:var(--sui-slider-track,6px); border-radius:var(--sui-slider-radius,999px); background:var(--sui-slider-thumb-bg,#4c5cf7); }
.sui-slider.sui-slider-thin { --sui-slider-track:3px; --sui-slider-thumb:14px; }
.sui-slider.sui-slider-classic { --sui-slider-track:6px; --sui-slider-thumb:18px; }
.sui-slider.sui-slider-pill { --sui-slider-track:12px; --sui-slider-thumb:18px; --sui-slider-thumb-radius:999px; --sui-slider-shadow:0 4px 10px rgba(76,92,247,.28); }
.sui-slider.sui-slider-glass { --sui-slider-track:10px; --sui-slider-bg:rgba(148,163,184,.18); --sui-slider-thumb:20px; --sui-slider-thumb-border-width:2px; --sui-slider-thumb-border:rgba(255,255,255,.55); --sui-slider-shadow:0 8px 18px rgba(76,92,247,.24); }
.sui-slider.sui-slider-square { --sui-slider-track:8px; --sui-slider-radius:3px; --sui-slider-thumb:17px; --sui-slider-thumb-radius:4px; }
.sui-slider.sui-slider-diamond::-webkit-slider-thumb { border-radius:3px; transform:rotate(45deg); }
.sui-slider.sui-slider-diamond::-moz-range-thumb { border-radius:3px; transform:rotate(45deg); }
.sui-slider.sui-slider-triangle::-webkit-slider-thumb { border-radius:0; clip-path:polygon(0 0,100% 0,50% 100%); }
.sui-slider.sui-slider-triangle::-moz-range-thumb { border-radius:0; clip-path:polygon(0 0,100% 0,50% 100%); }
.sui-slider.sui-slider-segment { --sui-slider-thumb-width:4px; --sui-slider-thumb-height:24px; --sui-slider-thumb-radius:2px; }
.sui-slider.sui-slider-neon { --sui-slider-track:5px; --sui-slider-thumb-bg:#22d3ee; --sui-slider-shadow:0 0 12px rgba(34,211,238,.32); --sui-slider-thumb-shadow:0 0 12px rgba(34,211,238,.72); }
.sui-progress {
  appearance: none;
  width: var(--sui-progress-width, 100%);
  height: var(--sui-progress-height, 8px);
  border: var(--sui-progress-border-width, 0) solid var(--sui-progress-border, transparent);
  border-radius: var(--sui-progress-radius, 999px);
  background: var(--sui-progress-track, #374151);
  box-shadow: var(--sui-progress-shadow, none);
  overflow: hidden;
}
.sui-progress::-webkit-progress-bar {
  background: var(--sui-progress-track, #374151);
}
.sui-progress::-webkit-progress-value {
  background: var(--sui-progress-fill, linear-gradient(90deg, var(--sui-progress-value, #4c5cf7), var(--sui-progress-value-end, #6366f1)));
  background-size: var(--sui-progress-fill-size, auto);
  border-radius: var(--sui-progress-radius, 999px);
  box-shadow: var(--sui-progress-fill-shadow, none);
  animation: var(--sui-progress-animation, none);
}
.sui-progress::-moz-progress-bar {
  background: var(--sui-progress-fill, linear-gradient(90deg, var(--sui-progress-value, #4c5cf7), var(--sui-progress-value-end, #6366f1)));
  background-size: var(--sui-progress-fill-size, auto);
  border-radius: var(--sui-progress-radius, 999px);
  box-shadow: var(--sui-progress-fill-shadow, none);
  animation: var(--sui-progress-animation, none);
}
.sui-progress.sui-progress-thin { --sui-progress-height:4px; --sui-progress-radius:999px; }
.sui-progress.sui-progress-classic { --sui-progress-height:8px; --sui-progress-radius:999px; }
.sui-progress.sui-progress-segmented { --sui-progress-height:12px; --sui-progress-radius:999px; --sui-progress-fill:repeating-linear-gradient(90deg,var(--sui-progress-value,#4c5cf7) 0 9px,transparent 9px 12px); }
.sui-progress.sui-progress-steps { --sui-progress-height:14px; --sui-progress-radius:3px; --sui-progress-fill:repeating-linear-gradient(90deg,var(--sui-progress-value,#4c5cf7) 0 15px,transparent 15px 19px); }
.sui-progress.sui-progress-neon { --sui-progress-height:10px; --sui-progress-radius:999px; --sui-progress-fill-shadow:0 0 10px var(--sui-progress-value,#4c5cf7); }
.sui-progress.sui-progress-square { --sui-progress-height:12px; --sui-progress-radius:3px; --sui-progress-border-width:1px; --sui-progress-border:var(--sui-progress-value,#4c5cf7); }
.sui-badge {
  display: inline-flex;
  align-items: center;
  padding: var(--sui-badge-padding, 2px 10px);
  border: var(--sui-badge-border-width, 1px) solid var(--sui-badge-border, transparent);
  border-radius: var(--sui-badge-radius, 999px);
  font-size: var(--sui-badge-font-size, 12px);
  font-weight: 500;
  background: var(--sui-badge-bg, rgba(76, 92, 247, 0.15));
  color: var(--sui-badge-fg, #a5b4fc);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sui-alert {
  padding: var(--sui-alert-padding, 12px 16px);
  border-radius: var(--sui-alert-radius, var(--sui-radius));
  background: var(--sui-alert-bg, rgba(239, 68, 68, 0.1));
  border: var(--sui-alert-border-width, 1px) var(--sui-alert-border-style, solid) var(--sui-alert-border, rgba(239, 68, 68, 0.3));
  color: var(--sui-alert-fg, #fca5a5);
  overflow-wrap: anywhere;
}
`

const suiCssSpinnerBase = `.sui-spinner {
  width: var(--sui-spinner-size, 24px);
  height: var(--sui-spinner-size, 24px);
  border: var(--sui-spinner-thickness, 3px) var(--sui-spinner-border-style, solid) var(--sui-spinner-track, rgba(156, 163, 175, 0.3));
  border-top-color: var(--sui-spinner-color, #4c5cf7);
  border-bottom-color: var(--sui-spinner-color2, transparent);
  background: var(--sui-spinner-bg, transparent);
  border-radius: var(--sui-spinner-radius, 50%);
  animation-name: var(--sui-spinner-animation, sui-spin), var(--sui-spinner-glow-animation, none);
  animation-duration: var(--sui-spinner-speed, .8s), 1.8s;
  animation-timing-function: var(--sui-spinner-easing, linear), ease-in-out;
  animation-iteration-count: infinite, infinite;
  animation-play-state: var(--sui-spinner-play, running), var(--sui-spinner-play, running);
}
`

const suiCssSpinnerSpin = `@keyframes sui-spin {
  to { transform: rotate(360deg); }
}
`

const suiCssSpinnerVariants = `@keyframes sui-spinner-glow {
  0%, 100% { filter: drop-shadow(0 0 2px var(--sui-spinner-color, #4c5cf7)); }
  50% { filter: drop-shadow(0 0 7px var(--sui-spinner-color, #4c5cf7)); }
}
@keyframes sui-spinner-glow-pulse {
  0%, 100% { filter: drop-shadow(0 0 1px var(--sui-spinner-color, #4c5cf7)); }
  50% { filter: drop-shadow(0 0 5px var(--sui-spinner-color, #4c5cf7)) brightness(1.25); }
}
@keyframes sui-spinner-fill-flip {
  0%, 35% { transform: perspective(90px) rotateY(0deg); background-color: var(--sui-spinner-color, #4c5cf7); }
  50%, 85% { transform: perspective(90px) rotateY(180deg); background-color: var(--sui-spinner-color2, #22d3ee); }
  100% { transform: perspective(90px) rotateY(360deg); background-color: var(--sui-spinner-color, #4c5cf7); }
}
@keyframes sui-spinner-hourglass {
  0% { transform: rotate(0deg); background-size: 100% 42%, 100% 100%; }
  38%, 44% { transform: rotate(0deg); background-size: 100% 0%, 100% 100%; }
  50% { transform: rotate(180deg); background-size: 100% 0%, 100% 100%; }
  82%, 88%, 94% { transform: rotate(180deg); background-size: 100% 42%, 100% 100%; }
  100% { transform: rotate(360deg); background-size: 100% 42%, 100% 100%; }
}
@keyframes sui-spinner-hourglass-sand {
  0%, 6% { height: 0%; }
  38%, 50% { height: 42%; }
  82%, 100% { height: 0%; }
}
@keyframes sui-spinner-hourglass-flow {
  0%, 5%, 42%, 54%, 92%, 100% { opacity: 0; transform: scaleY(0); }
  8%, 35%, 60%, 84% { opacity: 1; transform: scaleY(1); }
}
.sui-spinner[variant="hourglass"] {
  border: none;
  border-radius: 0;
  background:
    linear-gradient(var(--sui-spinner-color, #f59e0b), var(--sui-spinner-color, #f59e0b)) top / 100% 42% no-repeat,
    linear-gradient(var(--sui-spinner-track, rgba(148, 163, 184, .35)), var(--sui-spinner-track, rgba(148, 163, 184, .35))) top / 100% 100% no-repeat;
  clip-path: polygon(10% 0, 90% 0, 90% 8%, 58% 42%, 58% 58%, 90% 92%, 90% 100%, 10% 100%, 10% 92%, 42% 58%, 42% 42%, 10% 8%);
  animation-name: var(--sui-spinner-animation, sui-spinner-hourglass), var(--sui-spinner-glow-animation, none);
  animation-timing-function: linear, ease-in-out;
}
.sui-spinner[variant="hourglass"]::before {
  content: "";
  position: absolute;
  left: 48%;
  top: 43%;
  width: 6%;
  height: 14%;
  border-radius: 99px;
  background: linear-gradient(var(--sui-spinner-color, #f59e0b), var(--sui-spinner-color2, #fcd34d));
  transform-origin: top;
  animation: sui-spinner-hourglass-flow var(--sui-spinner-speed, 5.4s) linear infinite;
}
.sui-spinner[variant="hourglass"]::after {
  content: "";
  position: absolute;
  left: 10%;
  bottom: 0;
  width: 80%;
  height: 0;
  background: var(--sui-spinner-color2, #fcd34d);
  clip-path: polygon(0 100%, 50% 0, 100% 100%);
  animation: sui-spinner-hourglass-sand var(--sui-spinner-speed, 5.4s) ease-in-out infinite;
}
.sui-spinner[variant="fill-flip"] {
  border: none;
  background: var(--sui-spinner-color, #4c5cf7);
  transform-style: preserve-3d;
  backface-visibility: hidden;
  animation-name: var(--sui-spinner-animation, sui-spinner-fill-flip), var(--sui-spinner-glow-animation, none);
}
.sui-spinner[variant="fill-flip"]::before {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: var(--sui-spinner-color2, #22d3ee);
  transform: rotateY(180deg);
  backface-visibility: hidden;
}
`

const suiCssBasicBlocks = `.sui-card {
  background: var(--sui-card-bg, #1a1d26);
  color: var(--sui-card-fg, var(--sui-fg, #e6e6ef));
  border: var(--sui-card-border-width, 1px) var(--sui-card-border-style, solid) var(--sui-card-border, var(--sui-border, #3d4556));
  border-radius: var(--sui-card-radius, 12px);
  padding: var(--sui-card-padding, 16px);
  box-shadow: var(--sui-card-shadow, 0 4px 12px rgba(0, 0, 0, 0.3));
  backdrop-filter: var(--sui-card-backdrop, none);
  min-width: 0;
  max-width: 100%;
  overflow-wrap: anywhere;
}
.sui-divider {
  border: none;
  border-top: 1px solid var(--sui-border);
  margin: 16px 0;
}

/* Optional hover states for non-interactive controls. */
.sui-progress:hover { background: var(--sui-progress-hover-track, var(--sui-progress-track, #374151)); filter: var(--sui-progress-hover-filter, none); }
.sui-progress:hover::-webkit-progress-bar { background: var(--sui-progress-hover-track, var(--sui-progress-track, #374151)); }
.sui-progress:hover::-webkit-progress-value { background: var(--sui-progress-hover-fill, var(--sui-progress-fill, linear-gradient(90deg, var(--sui-progress-value, #4c5cf7), var(--sui-progress-value-end, #6366f1)))); }
.sui-progress:hover::-moz-progress-bar { background: var(--sui-progress-hover-fill, var(--sui-progress-fill, linear-gradient(90deg, var(--sui-progress-value, #4c5cf7), var(--sui-progress-value-end, #6366f1)))); }
`

const suiCssSpinnerImage = `.sui-spinner:hover {
  border-color: var(--sui-spinner-hover-track, var(--sui-spinner-track, #4b5563));
  border-top-color: var(--sui-spinner-hover-color, var(--sui-spinner-color, #4c5cf7));
  border-bottom-color: var(--sui-spinner-hover-color2, var(--sui-spinner-color2, transparent));
  filter: var(--sui-spinner-hover-filter, none);
}
/* Image/GIF spinner: the src attribute replaces the CSS ring and its
   animation. The wrapper keeps the user class so the designer can still
   style it. */
.sui-spinner.sui-spinner-image {
  display: inline-block;
  width: auto;
  height: auto;
  min-width: 0;
  min-height: 0;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  line-height: 0;
  animation: none;
  vertical-align: middle;
}
.sui-spinner.sui-spinner-image::before,
.sui-spinner.sui-spinner-image::after {
  content: none;
}
.sui-spinner.sui-spinner-image:hover {
  border: 0;
  background: transparent;
  filter: none;
}
.sui-spinner.sui-spinner-image > img.sui-spinner-img {
  display: block;
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: var(--sui-spinner-size, 96px);
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  object-fit: contain;
}
`

const suiCssFoundation2 = `.sui-divider:hover {
  border-top-color: var(--sui-divider-hover-color, var(--sui-divider-color, #3d4556));
  border-top-width: var(--sui-divider-hover-thickness, var(--sui-divider-thickness, 1px));
}

/* Media controls follow the project theme like every other control. */
img, video, audio, iframe {
  max-width: 100%;
  background: var(--sui-media-bg, var(--sui-panel, #1d2129));
  border: var(--sui-media-border-width, 0) var(--sui-media-border-style, solid) var(--sui-media-border, var(--sui-border, #3d4556));
  border-radius: var(--sui-media-radius, var(--sui-radius, 8px));
  box-shadow: var(--sui-media-shadow, none);
  box-sizing: border-box;
  padding: var(--sui-media-padding, 0px);
}
img, video { object-fit: var(--sui-media-fit, cover); }
iframe { display: block; width: var(--sui-media-width, 100%); }
audio { width: var(--sui-media-width, 100%); }

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

/* Structural layout table: real <table>/<tr>/<td> wrappers but no data-table
   styling. Children flow into fixed columns. */
.sui-layout-table {
  width: 100%;
  margin: 0;
  border-collapse: separate;
  border-spacing: var(--sui-layout-gap-x, var(--sui-layout-gap, 12px)) var(--sui-layout-gap-y, var(--sui-layout-gap, 12px));
  table-layout: fixed;
}
.sui-layout-cell {
  padding: 0;
  min-width: var(--sui-layout-min-cell-width, 40px);
  min-height: var(--sui-layout-min-cell-height, 24px);
  border: 0;
  background: transparent;
  vertical-align: top;
}
.sui-layout-cell-empty {
  height: var(--sui-layout-cell-min-height, 72px);
}
.sui-layout-table > tbody > tr:hover { background: transparent; }
.sui-layout-cell > * { min-width: 0; }
@media (max-width: 640px) {
  .sui-layout-responsive { display: block; border-spacing: 0; }
  .sui-layout-responsive > colgroup { display: none; }
  .sui-layout-responsive > tbody,
  .sui-layout-responsive > tbody > tr,
  .sui-layout-responsive > tbody > tr > .sui-layout-cell { display: block; width: 100% !important; }
  .sui-layout-responsive > tbody > tr > .sui-layout-cell.sui-layout-cell-aligned {
    display: flex; flex-direction: column;
    align-items: var(--sui-cell-align, stretch);
    justify-content: var(--sui-cell-justify, flex-start);
  }
  .sui-layout-responsive > tbody > tr > .sui-layout-cell { margin-bottom: var(--sui-layout-gap-y, 12px); }
}

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
  color: var(--sui-link, var(--sui-accent, #7db3ff));
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

`

const suiCssTable = `tbody tr:hover { background: var(--sui-table-hover-bg, #20252e); }

/* fixed-size tables: long cell text is truncated; click a cell to see it all */
table[data-sui-fixed] { table-layout: fixed; }
table[data-sui-fixed] th, table[data-sui-fixed] td { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
table[data-sui-fixed] td:focus, table[data-sui-fixed] th:focus { outline: 2px solid #4c5cf7; outline-offset: -2px; }
/* fixed-width wrapping: the column width stays stable, rows grow in height */
table[data-sui-cell-text="wrap"] { table-layout: fixed; }
table[data-sui-cell-text="wrap"] th, table[data-sui-cell-text="wrap"] td { white-space: normal; overflow-wrap: anywhere; word-break: break-word; }
/* editable tables: every cell can be edited by the end user */
table[data-sui-editable] td, table[data-sui-editable] th { cursor: text; }
table[data-sui-editable] td:focus, table[data-sui-editable] th:focus { outline: 2px solid #4c5cf7; outline-offset: -2px; background: var(--sui-table-focus-bg, rgba(76,92,247,.10)); }
table[data-sui-editable] td, table[data-sui-editable] th { -webkit-tap-highlight-color: transparent; }
table[data-sui-editable] ::selection { background: rgba(76,92,247,.22); color: inherit; }
.sui-table-filter { display: block; box-sizing: border-box; width: 100%; margin: 0 0 6px; padding: 6px 9px; font: inherit; font-size: 12px; color: var(--sui-fg, #f2f4f6); background: var(--sui-input-bg, #14171c); border: 1px solid var(--sui-border, #3d4556); border-radius: 6px; }
.sui-table-filter:focus { outline: 2px solid #4c5cf7; outline-offset: 1px; }
.sui-table-pager { display: flex; align-items: center; justify-content: flex-end; gap: 6px; margin-top: 6px; font-size: 11px; color: var(--sui-muted, #8fa0b8); }
.sui-table-page-btn { appearance: none; border: 1px solid var(--sui-border, #3d4556); border-radius: 5px; background: var(--sui-panel, #202732); color: var(--sui-fg, #f2f4f6); padding: 3px 8px; font: inherit; font-size: 11px; cursor: pointer; }
.sui-table-page-btn:hover:not(:disabled) { border-color: #4c5cf7; }
.sui-table-page-btn:disabled { opacity: .45; cursor: default; }
.sui-table-page-info { min-width: 74px; text-align: center; }
th.sui-table-sortable { cursor: pointer; user-select: none; }
th.sui-table-sortable:hover { color: var(--sui-fg, #f2f4f6); }
tr.sui-table-row-selected > td, tr.sui-table-row-selected > th { background: var(--sui-table-selected-bg, rgba(76,92,247,.18)); }
td.sui-table-cell-selected, th.sui-table-cell-selected { background: var(--sui-table-cell-selected-bg, rgba(76,92,247,.22)); }
td.sui-table-col-selected, th.sui-table-col-selected { background: var(--sui-table-col-selected-bg, rgba(76,92,247,.10)); }
td.sui-table-invalid, th.sui-table-invalid { background: rgba(239,68,68,.18); box-shadow: inset 0 0 0 1px #ef4444; }
.sui-table-add { display: inline-flex; align-items: center; gap: 4px; margin: 6px 0 0; padding: 4px 9px; border: 1px dashed var(--sui-border, #3d4556); border-radius: 6px; background: transparent; color: var(--sui-muted, #8fa0b8); font: inherit; font-size: 11px; cursor: pointer; }
.sui-table-add:hover { color: var(--sui-fg, #f2f4f6); border-color: #4c5cf7; }
.sui-table-topbar { display: flex; align-items: center; gap: 8px; width: 100%; min-height: 26px; margin: 0 0 4px; }
.sui-table-title { flex: 1 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 15px; font-weight: 600; }
.sui-table-tools { display: flex; justify-content: flex-end; flex: 0 0 auto; margin-left: auto; pointer-events: none; }
.sui-table-tools .sui-table-settings { margin: 0; }
.sui-table-settings { pointer-events: auto; display: inline-flex; align-items: center; justify-content: center; width: 24px; height: 24px; margin: 0 6px 0 0; padding: 0; border: 1px solid var(--sui-border, #3d4556); border-radius: 6px; background: var(--sui-panel, #202732); color: var(--sui-muted, #8fa0b8); font: inherit; font-size: 14px; line-height: 1; cursor: pointer; opacity: .72; box-shadow: 0 1px 4px rgba(0,0,0,.35); }
.sui-table-settings:hover, .sui-table-settings:focus, .sui-table-settings[aria-expanded="true"] { opacity: 1; color: var(--sui-fg, #f2f4f6); border-color: #4c5cf7; outline: none; }
.sui-table-hwrap { display: flex; align-items: center; gap: 6px; min-width: 0; flex-wrap: nowrap; }
.sui-table-hlabel { flex: 1 1 auto; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sui-table-hbtns { display: inline-flex; align-items: center; gap: 2px; flex: 0 0 auto; align-self: center; }
.sui-table-filterbtn, .sui-table-sortbtn { display: inline-flex; align-items: center; justify-content: center; width: 18px; height: 18px; padding: 0; border: 0; border-radius: 4px; background: transparent; color: inherit; opacity: .55; cursor: pointer; font: inherit; }
.sui-table-filterbtn:hover, .sui-table-sortbtn:hover, .sui-table-filterbtn:focus, .sui-table-sortbtn:focus { opacity: 1; background: rgba(127,127,127,.18); outline: none; }
.sui-table-filterbtn.active { opacity: 1; color: #4c5cf7; }
.sui-table-filterbtn::before { content: ""; display: block; width: 11px; height: 9px; background: currentColor; clip-path: polygon(0 0, 100% 0, 62% 46%, 62% 100%, 38% 80%, 38% 46%); }
.sui-table-sortbtn { font-size: 9px; line-height: 1; }
.sui-table-sortbtn::before { display: block; font-size: 9px; line-height: 1; }
.sui-table-sortbtn.asc::before { content: "\25B2"; }
.sui-table-sortbtn.desc::before { content: "\25BC"; }
.sui-table-sortbtn.none::before { content: "\21C5"; }
th.sui-table-filter-active { box-shadow: inset 0 -2px 0 #4c5cf7; }
.sui-table-sort-rank { display: inline-flex; align-items: center; justify-content: center; min-width: 12px; height: 14px; padding: 0 3px; border-radius: 7px; background: #4c5cf7; color: #fff; font-size: 9px; line-height: 1; }
.sui-table-cfg { position: fixed; z-index: 10050; width: 336px; max-width: calc(100vw - 16px); max-height: min(72vh, 540px); overflow: auto; padding: 10px 12px 12px; background: var(--sui-panel, #1d2129); color: var(--sui-fg, #e6e6ef); border: 1px solid var(--sui-border, #475569); border-radius: 10px; box-shadow: 0 12px 32px rgba(0,0,0,.5); font-size: 12px; }
.sui-table-cfg-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 8px; }
.sui-table-cfg-title { font-size: 13px; font-weight: 700; }
.sui-table-cfg-close { border: 0; background: transparent; color: var(--sui-muted, #8fa0b8); font: inherit; font-size: 16px; line-height: 1; cursor: pointer; padding: 2px 4px; }
.sui-table-cfg-close:hover { color: var(--sui-fg, #f2f4f6); }
.sui-table-cfg-row { display: flex; align-items: center; gap: 8px; margin: 6px 0; }
.sui-table-cfg-row label { display: flex; align-items: center; gap: 6px; cursor: pointer; }
.sui-table-cfg-sect { margin-top: 10px; padding-top: 8px; border-top: 1px solid var(--sui-border, #3d4556); }
.sui-table-cfg-sec { margin-bottom: 6px; font-size: 10px; letter-spacing: .05em; text-transform: uppercase; color: var(--sui-muted, #8fa0b8); }
.sui-table-cfg-level { display: grid; grid-template-columns: minmax(0, 1fr) 84px 24px; gap: 6px; align-items: center; margin: 5px 0; }
.sui-table-cfg select, .sui-table-cfg-btn { box-sizing: border-box; padding: 4px 6px; font: inherit; font-size: 11px; color: var(--sui-fg, #e6e6ef); background: var(--sui-input-bg, #14171c); border: 1px solid var(--sui-border, #3d4556); border-radius: 5px; }
.sui-table-cfg select:focus, .sui-table-cfg-btn:focus { outline: 2px solid #4c5cf7; outline-offset: 1px; }
.sui-table-cfg-btn { cursor: pointer; }
.sui-table-cfg-btn:hover { border-color: #4c5cf7; }
.sui-table-cfg-actions { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 10px; }
.sui-table-cfg-empty { margin: 4px 0; color: var(--sui-muted, #8fa0b8); font-size: 11px; }
.sui-table-fpop { position: fixed; z-index: 10060; width: 240px; max-width: calc(100vw - 16px); max-height: 380px; display: flex; flex-direction: column; padding: 10px; background: var(--sui-panel, #1d2129); color: var(--sui-fg, #e6e6ef); border: 1px solid var(--sui-border, #475569); border-radius: 8px; box-shadow: 0 12px 32px rgba(0,0,0,.45); font-size: 12px; }
.sui-table-fpop-title { font-weight: 700; margin-bottom: 6px; }
.sui-table-fpop-search { box-sizing: border-box; width: 100%; margin-bottom: 6px; padding: 5px 7px; font: inherit; font-size: 12px; color: var(--sui-fg, #e6e6ef); background: var(--sui-input-bg, #14171c); border: 1px solid var(--sui-border, #3d4556); border-radius: 4px; }
.sui-table-fpop-search:focus { outline: 2px solid #4c5cf7; outline-offset: 1px; }
.sui-table-fpop-list { flex: 1 1 auto; min-height: 48px; max-height: 240px; overflow: auto; padding: 4px 0; border-top: 1px solid var(--sui-border, #3d4556); border-bottom: 1px solid var(--sui-border, #3d4556); }
.sui-table-fpop-item { display: flex; align-items: center; gap: 6px; padding: 2px 4px; border-radius: 4px; cursor: pointer; user-select: none; }
.sui-table-fpop-item:hover { background: rgba(76,92,247,.14); }
.sui-table-fpop-item input { flex: 0 0 auto; margin: 0; }
.sui-table-fpop-lbl { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sui-table-fpop-empty { padding: 4px; color: var(--sui-muted, #8fa0b8); }
.sui-table-fpop-actions { display: flex; justify-content: space-between; gap: 8px; margin-top: 8px; }
.sui-table-fpop-btn { box-sizing: border-box; padding: 4px 8px; font: inherit; font-size: 11px; color: var(--sui-fg, #e6e6ef); background: var(--sui-input-bg, #14171c); border: 1px solid var(--sui-border, #3d4556); border-radius: 5px; cursor: pointer; }
.sui-table-fpop-btn:hover { border-color: #4c5cf7; }
.sui-table-fpop-btn:focus { outline: 2px solid #4c5cf7; outline-offset: 1px; }
.sui-table-struct-menu { position: fixed; z-index: 10045; min-width: 190px; padding: 4px; background: var(--sui-panel, #1d2129); border: 1px solid var(--sui-border, #475569); border-radius: 7px; box-shadow: 0 8px 24px rgba(0,0,0,.45); }
.sui-table-struct-item { display: block; width: 100%; text-align: left; border: 0; border-radius: 5px; background: transparent; color: var(--sui-fg, #e6e6ef); padding: 5px 8px; font: inherit; font-size: 12px; cursor: pointer; }
.sui-table-struct-item:hover, .sui-table-struct-item:focus { background: rgba(76,92,247,.25); outline: none; }
.sui-table-struct-item:disabled { opacity: .45; cursor: default; }
.sui-table-struct-sep { height: 1px; margin: 4px 3px; background: var(--sui-border, #3d4556); }
.sui-cell-pop { position: fixed; z-index: 9999; max-width: 420px; background: var(--sui-panel, #202732); color: var(--sui-fg, #f2f4f6); border: 1px solid var(--sui-border, #3d4556); border-radius: 8px; padding: 8px 10px; box-shadow: var(--sui-ac-shadow, 0 8px 22px rgba(0,0,0,.45)); font-size: 13px; line-height: 1.45; white-space: pre-wrap; word-break: break-word; }
`

const suiCssIframeError = `.sui-iframe-error { padding: 10px 14px; border: 1px solid #a33; border-radius: 8px; background: #2b1719; color: #ffb4b4; font-size: 13px; line-height: 1.4; }
`

const suiCssDnd = `/* Drag and drop: the runtime toggles these classes on the current source and
   drop target. The error is inserted as a sibling with textContent. */
.sui-dragging { opacity: .55; }
.sui-drop-active { outline: 2px dashed #4c5cf7 !important; outline-offset: 2px; background: rgba(76, 92, 247, .10) !important; }
.sui-drop-error { color: #ffb4b4; background: #2b1719; border: 1px solid #a33; border-radius: 6px; padding: 6px 8px; margin-top: 4px; font-size: 12px; line-height: 1.4; }

`

const suiCssMenuShared = `/* Menu bar: a horizontal nav with button/popup menus. The runtime toggles
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
  border: var(--sui-menu-button-border-width, 0px) var(--sui-menu-border-style, solid) var(--sui-menu-border, transparent);
  background: var(--sui-menu-button-bg, transparent);
  color: var(--sui-menu-fg, var(--sui-fg, #e6e6ef));
  border-radius: var(--sui-menu-radius, 6px);
  padding: var(--sui-menu-padding, 7px 12px);
  font: inherit;
  font-size: var(--sui-menu-font-size, 14px);
  cursor: pointer;
  white-space: nowrap;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sui-menu-button:hover,
.sui-menu-button:focus-visible,
.sui-menu.open > .sui-menu-button {
  background: var(--sui-menu-hover, var(--sui-hover));
  color: var(--sui-menu-hover-fg, var(--sui-menu-fg, var(--sui-fg, #e6e6ef)));
  outline: none;
}
.sui-menu-popup {
  display: none;
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 190px;
  max-width: min(360px, 92vw);
  padding: 4px;
  background: var(--sui-menu-popup-bg, #1d2129);
  color: var(--sui-menu-fg, var(--sui-fg, #e6e6ef));
  border: var(--sui-menu-border-width, 1px) var(--sui-menu-border-style, solid) var(--sui-menu-border, var(--sui-border));
  border-radius: var(--sui-menu-radius, 8px);
  box-shadow: var(--sui-menu-shadow, var(--sui-ac-shadow));
  z-index: 1000;
  overflow-wrap: anywhere;
}
.sui-menu.open > .sui-menu-popup { display: block; }
.sui-dropdown-right > .sui-menu-popup, [data-sui-dropdown-right] > .sui-menu-popup { left: auto; right: 0; }
.sui-dropdown-sep { height: 1px; margin: 4px 6px; background: var(--sui-border, #3d4556); }
[data-sui-popover] { position: relative; display: inline-block; }
.sui-popover-trigger { appearance:none; border:1px solid var(--sui-menu-border, var(--sui-border)); background:var(--sui-menu-button-bg, var(--sui-panel)); color:var(--sui-menu-fg, var(--sui-fg)); border-radius:var(--sui-menu-radius, 6px); padding:var(--sui-menu-padding, 6px 10px); font:inherit; cursor:pointer; }
.sui-popover-trigger:hover, [data-sui-popover].open > .sui-popover-trigger { border-color:var(--sui-accent, #4c5cf7); color:var(--sui-menu-hover-fg, var(--sui-menu-fg, var(--sui-fg))); }
.sui-popover-panel { position:absolute; z-index:1000; min-width:180px; max-width:min(360px, 92vw); padding:10px; background:var(--sui-menu-popup-bg, var(--sui-panel, #1d2129)); color:var(--sui-menu-fg, var(--sui-fg)); border:1px solid var(--sui-menu-border, var(--sui-border)); border-radius:var(--sui-menu-radius, 8px); box-shadow:var(--sui-ac-shadow, 0 10px 30px rgba(0,0,0,.45)); overflow-wrap:anywhere; }
`

const suiCssPopover = `.sui-popover-panel[hidden] { display:none; }
[data-sui-popover-position="top"] > .sui-popover-panel { left:0; bottom:calc(100% + 6px); }
[data-sui-popover-position="right"] > .sui-popover-panel { left:calc(100% + 6px); top:0; }
[data-sui-popover-position="bottom"] > .sui-popover-panel { left:0; top:calc(100% + 6px); }
[data-sui-popover-position="left"] > .sui-popover-panel { right:calc(100% + 6px); top:0; }
`

const suiCssDrawer = `[data-sui-drawer] { position:fixed; inset:0; z-index:1200; pointer-events:none; }
[data-sui-drawer][data-sui-drawer-open="1"] { pointer-events:auto; }
.sui-drawer-backdrop { position:absolute; inset:0; background:rgba(0,0,0,.45); opacity:0; transition:opacity var(--sui-drawer-duration,.2s) var(--sui-drawer-easing,ease); }
[data-sui-drawer][data-sui-drawer-open="1"] > .sui-drawer-backdrop { opacity:1; }
.sui-drawer { position:absolute; top:0; bottom:0; width:var(--sui-drawer-width,320px); max-width:92vw; display:flex; flex-direction:column; background:var(--sui-panel,#1d2129); color:var(--sui-fg); border:1px solid var(--sui-border); box-shadow:0 14px 50px rgba(0,0,0,.55); transition:transform var(--sui-drawer-duration,.2s) var(--sui-drawer-easing,ease); }
.sui-drawer-right { right:0; transform:translateX(100%); }
.sui-drawer-left { left:0; transform:translateX(-100%); }
[data-sui-drawer][data-sui-drawer-open="1"] > .sui-drawer { transform:none; }
.sui-drawer-head { flex:none; display:flex; align-items:center; gap:8px; padding:10px 12px; border-bottom:1px solid var(--sui-border); }
.sui-drawer-head b { flex:1; font-size:13px; }
.sui-drawer-close { border:0; background:transparent; color:var(--sui-muted); font-size:16px; cursor:pointer; }
.sui-drawer-close:hover { color:var(--sui-fg); }
.sui-drawer-body { flex:1; min-height:0; overflow:auto; padding:12px; }
`

const suiCssSplitter = `.sui-splitter { display:flex; width:100%; min-height:0; overflow:hidden; border:1px solid var(--sui-border); border-radius:8px; background:var(--sui-panel); }
.sui-splitter-vertical { flex-direction:column; }
.sui-split-pane { flex:1 1 0; min-width:0; min-height:0; overflow:auto; padding:10px; }
.sui-splitter-horizontal > .sui-split-pane { min-width:0; }
.sui-splitter-vertical > .sui-split-pane { min-height:0; }
.sui-split-bar { flex:0 0 8px; background:var(--sui-border); cursor:col-resize; position:relative; }
.sui-split-bar:hover, .sui-split-bar:focus-visible { background:var(--sui-accent,#4c5cf7); outline:none; }
.sui-splitter-vertical > .sui-split-bar { cursor:row-resize; }
.sui-splitbar-active, .sui-splitbar-active * { cursor:col-resize !important; user-select:none; }
.sui-splitter-vertical.sui-splitbar-active, .sui-splitter-vertical.sui-splitbar-active * { cursor:row-resize !important; }
`

const suiCssDropzone = `.sui-dropzone { display:flex; flex-direction:column; align-items:center; justify-content:center; gap:4px; width:100%; min-height:120px; padding:18px 14px; border:2px dashed var(--sui-border,#475569); border-radius:10px; background:var(--sui-panel); color:var(--sui-fg); text-align:center; cursor:pointer; transition:border-color .15s ease, background .15s ease; }
.sui-dropzone:hover, .sui-dropzone:focus-visible, .sui-dropzone.sui-drop-active { border-color:var(--sui-accent,#4c5cf7); background:rgba(76,92,247,.08); outline:none; }
.sui-dropzone-icon { font-size:24px; line-height:1; color:var(--sui-accent,#4c5cf7); }
.sui-dropzone-title { font-size:13px; font-weight:600; }
.sui-dropzone-hint { font-size:12px; color:var(--sui-muted,#8fa0b8); }
.sui-dropzone-status { font:11px/1.4 ui-monospace,Consolas,monospace; color:var(--sui-fg); word-break:break-all; max-height:72px; overflow:auto; }
.sui-dropzone-status:empty { display:none; }
`

const suiCssTimeline = `.sui-timeline { display:flex; flex-direction:column; }
.sui-timeline-item { position:relative; display:flex; gap:10px; padding:0 0 14px; }
.sui-tl-pending { opacity:.65; }
.sui-timeline-item::before { content:""; position:absolute; left:6px; top:16px; bottom:0; width:1px; background:var(--sui-border); }
.sui-timeline-item:last-child::before { display:none; }
.sui-timeline-dot { flex:none; width:13px; height:13px; border-radius:50%; background:var(--sui-panel); border:2px solid var(--sui-border); margin-top:2px; z-index:1; }
.sui-tl-done .sui-timeline-dot { background:#34d399; border-color:#34d399; }
.sui-tl-active .sui-timeline-dot { background:var(--sui-accent,#4c5cf7); border-color:var(--sui-accent,#4c5cf7); box-shadow:0 0 0 3px rgba(76,92,247,.25); }
.sui-tl-error .sui-timeline-dot { background:#ef4444; border-color:#ef4444; }
.sui-timeline-body { display:flex; flex-direction:column; gap:2px; min-width:0; }
.sui-timeline-title { font-size:13px; font-weight:600; color:var(--sui-fg); }
.sui-timeline-date { font-size:11px; color:var(--sui-muted,#8fa0b8); }
.sui-timeline-content { font-size:12px; color:var(--sui-muted,#8fa0b8); }
`

const suiCssEmptyState = `.sui-empty { display:flex; flex-direction:column; align-items:center; justify-content:center; gap:6px; padding:28px 16px; text-align:center; color:var(--sui-muted,#8fa0b8); border:1px dashed var(--sui-border); border-radius:10px; background:var(--sui-panel); }
.sui-empty-icon { font-size:32px; line-height:1; color:var(--sui-accent,#4c5cf7); }
.sui-empty-title { font-size:14px; font-weight:600; color:var(--sui-fg); }
.sui-empty-hint { font-size:12px; max-width:44ch; }
.sui-empty-body { display:flex; flex-direction:column; align-items:center; gap:8px; margin-top:4px; }
.sui-empty-compact { padding:16px 12px; }
.sui-empty-compact .sui-empty-icon { font-size:22px; }
`

const suiCssSkeleton = `.sui-skeleton { display:flex; flex-direction:column; gap:var(--sui-sk-gap,8px); width:var(--sui-sk-width,100%); }
.sui-skeleton-line, .sui-skeleton-dot { display:block; width:100%; height:var(--sui-sk-height,12px); border-radius:6px; background:linear-gradient(90deg,rgba(148,163,184,.12) 25%,rgba(148,163,184,.28) 37%,rgba(148,163,184,.12) 63%); background-size:400% 100%; animation:sui-sk var(--sui-sk-duration,1.4s) ease infinite; }
.sui-skeleton-dot { width:var(--sui-sk-height,48px); height:var(--sui-sk-height,48px); border-radius:50%; }
.sui-skeleton-static .sui-skeleton-line, .sui-skeleton-static .sui-skeleton-dot { animation:none; }
@keyframes sui-sk { 0% { background-position:100% 50%; } 100% { background-position:0 50%; } }
`

const suiCssCarousel = `.sui-carousel { display:flex; flex-direction:column; gap:6px; width:100%; }
.sui-carousel-frame { overflow:hidden; border:1px solid var(--sui-border); border-radius:8px; background:var(--sui-panel); }
.sui-carousel-track { display:flex; transition:transform var(--sui-carousel-duration,.28s) var(--sui-carousel-easing,ease); }
.sui-carousel-slide { flex:0 0 100%; min-width:100%; min-height:var(--sui-carousel-height,160px); display:flex; flex-direction:column; gap:8px; padding:12px; }
.sui-carousel-nav { display:flex; align-items:center; justify-content:center; gap:10px; }
.sui-carousel-prev, .sui-carousel-next { border:1px solid var(--sui-border); background:var(--sui-panel); color:var(--sui-fg); border-radius:6px; padding:4px 10px; font:inherit; cursor:pointer; }
.sui-carousel-prev:hover, .sui-carousel-next:hover { border-color:var(--sui-accent,#4c5cf7); }
.sui-carousel-dots { display:flex; gap:4px; }
.sui-carousel-dot { width:8px; height:8px; padding:0; border-radius:50%; border:1px solid var(--sui-border); background:transparent; cursor:pointer; }
.sui-carousel-dot.active { background:var(--sui-accent,#4c5cf7); border-color:var(--sui-accent,#4c5cf7); }
`

const suiCssStepper = `.sui-stepper { display:flex; align-items:center; gap:8px; width:100%; }
.sui-stepper-vertical { flex-direction:column; align-items:stretch; gap:6px; }
.sui-step { display:flex; align-items:center; gap:8px; color:var(--sui-muted,#8fa0b8); font-size:12px; }
.sui-step-num { flex:none; width:24px; height:24px; display:flex; align-items:center; justify-content:center; border:1px solid var(--sui-border); border-radius:50%; font-size:11px; }
.sui-step-text { display:flex; flex-direction:column; }
.sui-step-label { font-weight:600; }
.sui-step-desc { font-size:11px; }
.sui-step.active { color:var(--sui-fg); }
.sui-step.active .sui-step-num { background:var(--sui-accent,#4c5cf7); border-color:var(--sui-accent,#4c5cf7); color:#fff; }
.sui-step.done .sui-step-num { background:rgba(52,211,153,.18); border-color:#34d399; color:#34d399; }
.sui-step-line { flex:1 1 auto; min-width:12px; height:1px; background:var(--sui-border); }
.sui-stepper-vertical .sui-step-line { flex:none; width:1px; height:12px; margin-left:12px; }
[data-sui-step] { cursor:pointer; }
[data-sui-step]:focus-visible { outline:2px solid var(--sui-accent,#4c5cf7); outline-offset:2px; border-radius:6px; }
`

const suiCssCommandbar = `.sui-commandbar { position:relative; width:100%; }
.sui-commandbar-panel { background:var(--sui-panel); border:1px solid var(--sui-border); border-radius:10px; overflow:hidden; box-shadow:0 12px 36px rgba(0,0,0,.35); }
.sui-commandbar-input { width:100%; border:0; border-bottom:1px solid var(--sui-border); background:transparent; color:var(--sui-fg); padding:11px 12px; font:inherit; outline:none; }
.sui-commandbar-list { max-height:320px; overflow:auto; padding:4px; }
.sui-command { display:flex; width:100%; align-items:center; justify-content:space-between; gap:10px; border:0; background:transparent; color:var(--sui-fg); text-align:left; padding:8px 10px; border-radius:6px; font:inherit; cursor:pointer; }
.sui-command-label { min-width:0; max-width:100%; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.sui-command:hover { background:var(--sui-hover,#2d3545); }
.sui-command:focus { background:var(--sui-hover,#2d3545); outline:2px solid var(--sui-accent,#4c5cf7); outline-offset:-2px; }
.sui-command.active { background:var(--sui-command-selected-bg, rgba(76,92,247,.2)); color:var(--sui-command-selected-fg, var(--sui-fg,#e6e6ef)); box-shadow:inset 3px 0 0 var(--sui-command-selected-accent, var(--sui-accent,#4c5cf7)); font-weight:600; }
.sui-command[hidden] { display:none; }
.sui-command-shortcut { font:11px ui-monospace,Consolas,monospace; color:var(--sui-muted,#8fa0b8); border:1px solid var(--sui-border); border-radius:4px; padding:1px 5px; }
.sui-commandbar-empty { padding:10px 12px; color:var(--sui-muted,#8fa0b8); font-size:12px; }
.sui-commandbar-empty[hidden] { display:none; }
.sui-commandbar-overlay { position:fixed; inset:0; z-index:1300; display:none; align-items:flex-start; justify-content:center; padding-top:12vh; }
.sui-commandbar-overlay[data-sui-commandbar-open="1"] { display:flex; }
.sui-commandbar-overlay .sui-commandbar-panel { position:relative; z-index:1; width:min(520px,92vw); background:var(--sui-panel,#1d2129); }
.sui-commandbar-backdrop { position:absolute; inset:0; background:rgba(0,0,0,.5); }
`

const suiCssAnim = `@keyframes sui-anim-fade { from { opacity:0; } to { opacity:1; } }
@keyframes sui-anim-scale { from { opacity:0; transform:scale(.96); } to { opacity:1; transform:scale(1); } }
@keyframes sui-anim-slide-down { from { opacity:0; transform:translateY(-6px); } to { opacity:1; transform:none; } }
@keyframes sui-anim-slide-up { from { opacity:0; transform:translateY(6px); } to { opacity:1; transform:none; } }
@keyframes sui-anim-slide-right { from { opacity:0; transform:translateX(-6px); } to { opacity:1; transform:none; } }
@keyframes sui-anim-slide-left { from { opacity:0; transform:translateX(6px); } to { opacity:1; transform:none; } }
[data-sui-anim="fade"] { animation: sui-anim-fade var(--sui-anim-duration,140ms) var(--sui-anim-easing,ease-out) var(--sui-anim-delay,0ms); }
[data-sui-anim="scale"] { animation: sui-anim-scale var(--sui-anim-duration,140ms) var(--sui-anim-easing,ease-out) var(--sui-anim-delay,0ms); }
[data-sui-anim="slide"][data-sui-anim-shift="down"] { animation: sui-anim-slide-down var(--sui-anim-duration,140ms) var(--sui-anim-easing,ease-out) var(--sui-anim-delay,0ms); }
[data-sui-anim="slide"][data-sui-anim-shift="up"] { animation: sui-anim-slide-up var(--sui-anim-duration,140ms) var(--sui-anim-easing,ease-out) var(--sui-anim-delay,0ms); }
[data-sui-anim="slide"][data-sui-anim-shift="right"] { animation: sui-anim-slide-right var(--sui-anim-duration,140ms) var(--sui-anim-easing,ease-out) var(--sui-anim-delay,0ms); }
[data-sui-anim="slide"][data-sui-anim-shift="left"] { animation: sui-anim-slide-left var(--sui-anim-duration,140ms) var(--sui-anim-easing,ease-out) var(--sui-anim-delay,0ms); }
`

const suiCssMenu = `.sui-menu-popup > a { display: block; color: inherit; text-decoration: none; }
.sui-menuitem {
  display: block;
  width: 100%;
  padding: var(--sui-menu-item-padding, 8px 12px);
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--sui-menu-fg, var(--sui-fg, #e6e6ef));
  font: inherit;
  font-size: var(--sui-menu-font-size, 14px);
  text-align: left;
  cursor: pointer;
  white-space: normal;
  overflow-wrap: anywhere;
}
.sui-menuitem:hover {
  background: var(--sui-menu-hover, var(--sui-hover));
  color: var(--sui-menu-hover-fg, var(--sui-menu-fg, var(--sui-fg, #e6e6ef)));
  outline: none;
}
.sui-menuitem:focus {
  background: var(--sui-menu-hover, var(--sui-hover));
  color: var(--sui-menu-hover-fg, var(--sui-menu-fg, var(--sui-fg, #e6e6ef)));
  outline: 2px solid var(--sui-accent, #4c5cf7);
  outline-offset: -2px;
}
.sui-menuitem.active,
.sui-menuitem.selected,
.sui-menuitem[aria-selected="true"] {
  background: var(--sui-menu-selected-bg, rgba(76,92,247,.2));
  color: var(--sui-menu-selected-fg, var(--sui-menu-fg, var(--sui-fg, #e6e6ef)));
  box-shadow: inset 3px 0 0 var(--sui-menu-selected-accent, var(--sui-accent, #4c5cf7));
  font-weight: 600;
}
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
`

const suiCssTabs = `/* Tabs: tablist + panels. The runtime toggles active/hidden and restores the
   selected tab after a view re-render when data-sui-tabs-key is present. */
.sui-tabs {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--sui-tabs-border, var(--sui-border));
  border-radius: var(--sui-tabs-radius, var(--sui-radius));
  background: var(--sui-tabs-bg, transparent);
  overflow: hidden;
}
.sui-tablist {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  padding: 4px 4px 0;
  background: var(--sui-tablist-bg, var(--sui-tabs-bg, #1d2129));
  border-bottom: 1px solid var(--sui-tabs-border, var(--sui-border));
  min-width: 0;
}
.sui-tab {
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--sui-tab-fg, var(--sui-tabs-fg, var(--sui-fg, #9aa7b8)));
  border-radius: var(--sui-tab-radius, 6px 6px 0 0);
  padding: var(--sui-tab-padding, 8px 14px);
  font: inherit;
  font-size: var(--sui-tab-font-size, inherit);
  cursor: pointer;
  white-space: nowrap;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sui-tab:hover,
.sui-tab:focus-visible {
  background: var(--sui-tab-hover-bg, var(--sui-hover));
  color: var(--sui-tab-hover-fg, var(--sui-tabs-fg, var(--sui-fg, #e6e6ef)));
  outline: none;
}
.sui-tab.active {
  color: var(--sui-tab-active-fg, var(--sui-tabs-fg, var(--sui-fg, #e6e6ef)));
  background: var(--sui-tab-active-bg, #14171c);
  box-shadow: inset 0 -2px 0 var(--sui-tab-accent, #4c5cf7);
}
.sui-tab-add {
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--sui-tab-fg, var(--sui-tabs-fg, var(--sui-fg, #9aa7b8)));
  border-radius: var(--sui-tab-radius, 6px 6px 0 0);
  padding: 8px 12px;
  font: inherit;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  white-space: nowrap;
}
.sui-tab-add:hover,
.sui-tab-add:focus-visible { background: var(--sui-tab-hover-bg, var(--sui-hover)); color: var(--sui-tab-hover-fg, var(--sui-tabs-fg, var(--sui-fg))); outline: none; }
.sui-tab-x { display: inline-flex; align-items: center; justify-content: center; width: 14px; height: 14px; margin-left: 6px; border-radius: 3px; font-size: 12px; line-height: 1; opacity: .65; cursor: pointer; }
.sui-tab-x:hover { opacity: 1; background: rgba(239,68,68,.25); }
.sui-tabs[data-sui-tabs-position="bottom"] .sui-tab-add { border-radius: 0 0 6px 6px; }
.sui-tabs[data-sui-tabs-variant="sheets"] .sui-tab-add { border: 1px solid transparent; border-bottom: 0; border-radius: 6px 6px 0 0; padding: 5px 12px; font-size: 15px; }
.sui-tabs[data-sui-tabs-variant="sheets"][data-sui-tabs-position="bottom"] .sui-tab-add { border-bottom: 0; border-top: 1px solid transparent; border-radius: 0 0 6px 6px; }
.sui-tabpanels { padding: 12px; }
.sui-tabpanel[hidden] { display: none; }
.sui-tabs.sui-tabs-rail { flex-direction: row; }
.sui-tabs.sui-tabs-rail > .sui-tablist {
  flex: 0 0 auto;
  flex-direction: column;
  flex-wrap: nowrap;
  border-bottom: 0;
  border-right: 1px solid var(--sui-tabs-border, var(--sui-border));
  padding: 6px;
}
.sui-tabs.sui-tabs-rail .sui-tab { border-radius: 6px; text-align: left; }
.sui-tabs.sui-tabs-rail .sui-tab.active { box-shadow: inset -2px 0 0 var(--sui-tab-accent, #4c5cf7); }
.sui-tabs.sui-tabs-rail > .sui-tabpanels { flex: 1 1 auto; min-width: 0; }
.sui-tabs[data-sui-tabs-position="bottom"] .sui-tablist { order: 2; border-bottom: 0; border-top: 1px solid var(--sui-border); padding: 4px 4px 0; }
.sui-tabs[data-sui-tabs-position="bottom"] .sui-tabpanels { order: 1; }
.sui-tabs[data-sui-tabs-position="bottom"] .sui-tab { border-radius: 0 0 6px 6px; }
.sui-tabs[data-sui-tabs-position="bottom"] .sui-tab.active { box-shadow: inset 0 2px 0 #4c5cf7; }
.sui-tabs[data-sui-tabs-variant="sheets"] .sui-tablist { gap: 1px; background: #14171c; padding: 4px 6px 0; }
.sui-tabs[data-sui-tabs-variant="sheets"] .sui-tab { border: 1px solid transparent; border-bottom: 0; border-radius: 6px 6px 0 0; padding: 5px 12px; font-size: 12px; }
.sui-tabs[data-sui-tabs-variant="sheets"][data-sui-tabs-position="bottom"] .sui-tab { border-bottom: 0; border-top: 1px solid transparent; border-radius: 0 0 6px 6px; }
.sui-tabs[data-sui-tabs-variant="sheets"] .sui-tab.active { background: var(--sui-panel, #1d2129); border-color: var(--sui-border); }
.sui-tabs[data-sui-tabs-variant="sheets"] .sui-tabpanels { padding: 0; }
/* Workbook: one table control. It fills the host (minimum 640x300) and the
   active sheet takes all the space between the title row and the bottom tabs.
   The sheet area scrolls horizontally at its bottom, just above the tabs; the
   header row stays sticky while the body rows scroll vertically. */
.sui-tabs[data-sui-workbook="1"] {
  width: 100%;
  height: 100%;
  min-width: 640px;
  min-height: 300px;
  box-sizing: border-box;
}
.sui-tabs[data-sui-workbook="1"] .sui-tabpanels {
  flex: 1 1 0;
  min-width: 0;
  min-height: 0;
  display: flex;
}
.sui-tabs[data-sui-workbook="1"] .sui-tabpanel {
  flex: 1 1 0;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  /* The viewport inset is outside the scroll container, so the row/column
     gutter strips live in real padding instead of over scrolled cells. */
  padding: var(--sui-workbook-pad, 10px) var(--sui-workbook-pad, 10px) 0;
  gap: var(--sui-workbook-pad, 10px);
  box-sizing: border-box;
}
.sui-tabs[data-sui-workbook="1"] .sui-tabpanel[hidden] { display: none; }
.sui-tabs[data-sui-workbook="1"][data-sui-tabcount="1"] .sui-tablist .sui-tab { display: none; }
.sui-tabs[data-sui-workbook="1"] .sui-table-topbar { flex: 0 0 auto; }
.sui-tabs[data-sui-workbook="1"] .tscroll {
  flex: 1 1 0;
  min-width: 0;
  min-height: 0;
  width: 100%;
  height: 100%;
  max-height: none;
  box-sizing: border-box;
}
.sui-tabs[data-sui-workbook="1"] .tscroll > table {
  margin: 0;
  table-layout: fixed;
  width: 100%;
}
.sui-tabs[data-sui-workbook="1"] .tscroll th,
.sui-tabs[data-sui-workbook="1"] .tscroll td {
  width: 160px;
  min-width: 160px;
}
.sui-tabs[data-sui-workbook="1"] .tscroll thead th {
  position: sticky;
  top: 0;
  z-index: 2;
}
/* While editing, a truncated cell shows its full text without changing size. */
.sui-tabs[data-sui-workbook="1"] table[data-sui-fixed] td.sui-table-editing,
.sui-tabs[data-sui-workbook="1"] table[data-sui-fixed] th.sui-table-editing {
  overflow: visible;
  text-overflow: clip;
  white-space: nowrap;
  position: relative;
  z-index: 3;
}
.sui-tabs[data-sui-workbook="1"][data-sui-workbook-editable="0"] .sui-tab-add { display: none; }
.sui-tabs[data-sui-workbook="1"][data-sui-tabcount="1"][data-sui-workbook-editable="0"] .sui-tablist { display: none; }
/* The designer preview always offers + to create a workbook, even when the
   source table is read-only. */
body.sui-designer-preview .sui-tabs[data-sui-workbook="1"] .sui-tab-add { display: inline-block; }
body.sui-designer-preview .sui-tabs[data-sui-workbook="1"][data-sui-tabcount="1"] .sui-tablist { display: flex; }
`

const suiCssAccordion = `
/* Accordion: stacked disclosure sections. The runtime keeps one panel open
   and preserves the open index across view re-renders when keyed. */
.sui-accordion {
  border: 1px solid var(--sui-accordion-border, var(--sui-border));
  border-radius: var(--sui-accordion-radius, var(--sui-radius));
  overflow: hidden;
  background: var(--sui-accordion-bg, #1d2129);
  color: var(--sui-accordion-fg, var(--sui-fg));
}
.sui-accordion-item + .sui-accordion-item { border-top: 1px solid var(--sui-accordion-border, var(--sui-border)); }
.sui-accordion-head {
  appearance: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  border: 0;
  background: transparent;
  color: var(--sui-accordion-fg, var(--sui-fg));
  padding: 11px 14px;
  font: inherit;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  min-width: 0;
  overflow-wrap: anywhere;
}
.sui-accordion-head::after {
  content: "\25B8";
  color: var(--sui-accordion-muted, var(--sui-muted, #9aa7b8));
  font-size: 13px;
  line-height: 1;
}
.sui-accordion-item.open > .sui-accordion-head::after { content: "\25BE"; }
.sui-accordion-head:hover,
.sui-accordion-head:focus-visible {
  background: var(--sui-accordion-hover-bg, var(--sui-hover));
  color: var(--sui-accordion-hover-fg, var(--sui-accordion-fg, var(--sui-fg)));
  outline: none;
}
.sui-accordion-panel { padding: 0 14px 12px; color: var(--sui-accordion-fg, var(--sui-fg)); overflow-wrap: anywhere; }
.sui-accordion-panel[hidden] { display: none; }
.sui-accordion.sui-accordion-separated {
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: visible;
  border: 0;
  background: transparent;
}
.sui-accordion-separated .sui-accordion-item {
  overflow: hidden;
  border: 1px solid var(--sui-accordion-border, var(--sui-border));
  border-radius: var(--sui-accordion-radius, 10px);
  background: var(--sui-accordion-bg, #1d2129);
}
.sui-accordion-separated .sui-accordion-item + .sui-accordion-item { border-top: 1px solid var(--sui-accordion-border, var(--sui-border)); }
`

const suiCssToast = `
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
`

const suiCssTooltip = `
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
  letter-spacing: var(--sui-tooltip-tracking, normal);
  font-size: 12px;
  line-height: 1.35;
  text-align: left;
  white-space: normal;
  overflow-wrap: anywhere;
  pointer-events: none;
}
.sui-tooltip-bubble[hidden] { display: none; }
.sui-tooltip-bubble.open:not([data-sui-anim]) { animation: sui-tooltip-in var(--sui-tooltip-duration,.12s) var(--sui-tooltip-easing,ease-out) var(--sui-tooltip-anim-delay,0ms); }
@keyframes sui-tooltip-in { from { opacity: 0; } to { opacity: 1; } }
`

const suiCssModal = `/* Modal dialogs */
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
  background: var(--sui-modal-bg, #1d2129);
  color: var(--sui-modal-fg, var(--sui-fg));
  border: 1px solid var(--sui-modal-border, var(--sui-border));
  border-radius: var(--sui-modal-radius, 10px);
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
  border: 1px solid var(--sui-modal-border, var(--sui-border));
  background: var(--sui-modal-button-bg, var(--sui-panel, #252b36));
  color: var(--sui-modal-button-fg, var(--sui-modal-fg, var(--sui-fg)));
  border-radius: 6px;
  padding: 8px 14px;
  font: inherit;
  cursor: pointer;
}
.sui-modal-button:hover, .sui-modal-button:focus-visible {
  background: var(--sui-modal-button-hover, var(--sui-hover));
  color: var(--sui-modal-button-hover-fg, var(--sui-modal-button-fg, var(--sui-modal-fg, var(--sui-fg))));
  outline: none;
}
.sui-modal-button.primary { background: #4c5cf7; border-color: #4c5cf7; color: #fff; }
.sui-modal-button.danger { background: #a33; border-color: #a33; color: #fff; }
body.sui-modal-open { overflow: hidden; }
`

// suiCssBase is the complete base stylesheet (all slices).
const suiCssBase = suiCssFoundation +
	suiCssCombobox +
	suiCssDatepicker +
	suiCssTree +
	suiCssChart +
	suiCssControls +
	suiCssSpinnerBase +
	suiCssSpinnerSpin +
	suiCssSpinnerVariants +
	suiCssBasicBlocks +
	suiCssSpinnerImage +
	suiCssFoundation2 +
	suiCssTable +
	suiCssIframeError +
	suiCssDnd +
	suiCssMenuShared +
	suiCssPopover +
	suiCssDrawer +
	suiCssSplitter +
	suiCssDropzone +
	suiCssTimeline +
	suiCssEmptyState +
	suiCssSkeleton +
	suiCssCarousel +
	suiCssStepper +
	suiCssCommandbar +
	suiCssAnim +
	suiCssMenu +
	suiCssTabs +
	suiCssAccordion +
	suiCssToast +
	suiCssTooltip +
	suiCssModal
