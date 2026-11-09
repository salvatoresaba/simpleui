# Chart / Gantt table colors

Complete color reference for the synchronized chart table, dark and light themes. A visual swatch version is available in [`table_colors.html`](table_colors.html).

Alpha values are overlays: the final color depends on the underlying cell/theme background. `fallback` means the value is used when the author has not overridden the token.

## Struttura e superfici

| Element / state | Dark | Light | Code / token | Note |
|---|---|---|---|---|
| Testo tabella | `#e6e6ef` | `#172033` | `--sui-chart-fg` | testo base di celle, menu e controlli |
| Testo secondario | `#8fa0b8` | `#5b6779` | `--sui-chart-muted` | caption, status, etichette filtro |
| Griglia e bordi | `#2c3444` | `#dbe2ec` | `--sui-chart-grid` | righe, separatori, bordi tabella |
| Sfondo tabella | `transparent` | `#ffffff` | `chart background` | le celle normali sono trasparenti |
| Intestazione: sfondo | `#14171c (fallback)` | `#eef3fa` | `--sui-chart-table-head-bg` | sticky in alto |
| Intestazione: testo | `#e6e6ef` | `#172033` | `--sui-chart-title-color` | include frecce di ordinamento |
| Prima colonna sticky | `#14171c (fallback)` | `#eef3fa` | `--sui-chart-table-sticky-bg` | colonna Task |
| Riga pari | `rgba(127,127,127,.06)` | `rgba(127,127,127,.06)` | `rgba(127,127,127, var(--sui-chart-table-stripe-opacity,.06))` | strisce alternate |
| Cella normale | `transparent` | `transparent` | `background` | mostra lo sfondo del grafico/pagina |
| Contorno selezione | `#38bdf8` | `#0284c7` | `--sui-chart-table-selection-border` | bordo 2px attorno al blocco selezionato |

## Righe e stati di selezione

| Element / state | Dark | Light | Code / token | Note |
|---|---|---|---|---|
| Riga hover | `rgba(100,116,139,.14)` | `#f1f5f9` | `--sui-chart-table-focus-row-bg` | solo se non selezionata |
| Cella hover | `rgba(100,116,139,.22)` | `#e2e8f0` | `--sui-chart-table-focus-cell-bg` | con outline 1px |
| Outline hover/focus | `#64748b` | `#64748b` | `--sui-chart-table-focus-color` | 1px su celle/righe/colonne |
| Riga selezionata / multi / colonna / cella | `rgba(56,189,248,.22)` | `#e0f2fe` | `--sui-chart-table-selection-bg` | |

| Riga selezionata in modalità cella | `transparent` | `transparent` | `table.sui-chart-table-mode-cell ... row-selected td` | solo la cella selezionata viene riempita |
| Riga summary | `rgba(56,189,248,.07)` | `#f2f8ff` | `--sui-chart-table-summary-bg` | testo in grassetto |
| Warning: bordo sinistro | `#f87171` | `#dc2626` | `--sui-chart-error` | 3px sulla prima cella |
| Critical: bordo sinistro | `#f59e0b` | `#d97706` | `--sui-chart-critical` | 3px sulla prima cella |
| Flash riga, inizio | `rgba(56,189,248,.20)` | `rgba(56,189,248,.18)` | `--sui-chart-table-flash-bg` | sfuma a trasparente in .9s |
| Flash cella, anello | `#38bdf8` | `#0284c7` | `--sui-chart-table-selection-border` | box-shadow interno 2px, sfuma a trasparente |

## Schede Tasks / Resources

| Element / state | Dark | Light | Code / token | Note |
|---|---|---|---|---|
| Scheda attiva: sfondo | `rgba(56,189,248,.22)` | `#e0f2fe` | `--sui-chart-table-selection-bg` | stessa selezione delle righe |
| Scheda attiva: testo | `#e6e6ef` | `#172033` | `--sui-chart-fg` | bordo `--sui-chart-control-border` |
| Scheda hover: sfondo | `rgba(127,127,127,.12)` | `rgba(127,127,127,.12)` | `--sui-chart-control-hover-bg` | solo schede non attive |
| Registro: celle/input | come tabella Task | come tabella Task | tabelle e input condivisi | stessi token di celle, hover e focus |
| Elimina risorsa | `#f87171` | `#dc2626` | `--sui-chart-error` | bordo e testo al hover/focus |

## Editor e controlli delle celle

| Element / state | Dark | Light | Code / token | Note |
|---|---|---|---|---|
| Input: testo | `#e6e6ef` | `#172033` | `color: inherit` | nessun colore separato |
| Input: sfondo normale | `transparent` | `transparent` | `background` | eredita la cella |
| Input: bordo hover | `#475569` | `#9aa8ba` | `--sui-chart-axis` | 1px |
| Input: bordo focus | `#64748b` | `#64748b` | `--sui-chart-table-focus-color` | 1px |
| Input: sfondo focus | `rgba(127,127,127,.08)` | `#ffffff` | `--sui-chart-table-input-bg` | solo al focus |
| Input: bordo invalido | `#f87171` | `#dc2626` | `--sui-chart-error` | aria-invalid=true |
| Option dei select: sfondo | `#1b2230` | `#ffffff` | `--sui-chart-table-input-bg (option)` | testo: --sui-chart-fg |
| Stepper: sfondo | `#14171c` | `#f7f9fc` | `--sui-chart-control-bg` | visibile solo al focus/hover cella |
| Stepper: bordo | `#475569` | `#9aa8ba` | `--sui-chart-axis` | 1px |
| Stepper: testo | `#e6e6ef` | `#172033` | `--sui-chart-fg` | simboli + / - |
| Stepper: hover | `#2c3444` | `#e9eef6` | `--sui-chart-control-hover-bg` | sfondo |
| Stepper: focus | `#64748b` | `#64748b` | `--sui-chart-table-focus-color` | outline 2px |
| Gutter riga: testo | `#8fa0b8` | `#5b6779` | `--sui-chart-muted` | numero riga |
| Gutter riga: hover | `#e6e6ef / #2c3444` | `#172033 / #e9eef6` | `--sui-chart-fg + --sui-chart-control-hover-bg` | testo / sfondo |
| Delete riga: normale | `#8fa0b8` | `#5b6779` | `--sui-chart-muted` | pulsante × |
| Delete riga: hover | `#f87171` | `#dc2626` | `--sui-chart-error` | testo e bordo |

## Filtri, menu e controlli della tabella

| Element / state | Dark | Light | Code / token | Note |
|---|---|---|---|---|
| Group toggle: bordo | `#2c3444` | `#dbe2ec` | `--sui-chart-grid` | pulsanti gruppo |
| Group toggle: testo | `#8fa0b8` | `#5b6779` | `--sui-chart-muted` | freccia ▼/▶ inclusa |
| Group toggle: hover testo | `#e6e6ef` | `#172033` | `--sui-chart-fg` | testo |
| Group toggle: hover bordo | `#4c5cf7 (fallback series-1)` | `#8fa3bb` | `--sui-chart-control-hover-border` | può essere sovrascritto dal tema/autore |
| WBS table toggle | `#8fa0b8 → #e6e6ef` | `#5b6779 → #172033` | `--sui-chart-muted → --sui-chart-fg` | normale / hover |
| Dependency chip | `rgba(56,189,248,.22) / #38bdf8` | `#e0f2fe / #0284c7` | `selection tokens` | sfondo / bordo; testo inherit |
| Filter button: normale | `#8fa0b8` | `#5b6779` | `--sui-chart-muted` | icona filtro |
| Filter button: hover/active | `#2c3444 / #4c5cf7 / #e6e6ef` | `#e9eef6 / #8fa3bb / #172033` | `control hover tokens` | sfondo / bordo / testo |
| Table button: normale | `#14171c / #475569 / #e6e6ef` | `#f7f9fc / #c5cfdd / #172033` | `control tokens` | sfondo / bordo / testo |
| Table button: hover | `#2c3444 / #4c5cf7` | `#e9eef6 / #8fa3bb` | `control hover tokens` | sfondo / bordo |
| Filter/select input: sfondo | `#14171c` | `#ffffff` | `chart bg` | bordo --sui-chart-axis; testo --sui-chart-fg |
| Popup/menu: sfondo | `#14171c` | `#f7f9fc` | `--sui-chart-control-bg` | bordo --sui-chart-axis |
| Popup/menu: ombra | `rgba(0,0,0,.30–.35)` | `rgba(0,0,0,.30–.35)` | `box-shadow` | filtri, colonne, menu contestuale |
| Menu action: testo | `#e6e6ef` | `#172033` | `--sui-chart-fg` | hover: --sui-chart-control-hover-bg |
| Column resize: linea | `#475569 · 50%` | `#9aa8ba · 50%` | `--sui-chart-axis` | hover: --sui-chart-control-hover-border |
| Sort arrow | `#e6e6ef` | `#172033` | `inherits header text` | ▲ / ▼ accanto all'intestazione |
| Readonly / derived | `#e6e6ef` | `#172033` | `inherits --sui-chart-fg` | nessun fondo dedicato |

## Theme tokens

| Token | Dark | Light |
|---|---|---|
| `--sui-chart-table-selection-bg` | `rgba(56,189,248,.22)` | `#e0f2fe` |
| `--sui-chart-table-selection-border` | `#38bdf8` | `#0284c7` |
| `--sui-chart-table-summary-bg` | `rgba(56,189,248,.07)` | `#f2f8ff` |
| `--sui-chart-table-flash-bg` | `rgba(56,189,248,.20)` | `rgba(56,189,248,.18)` |
| `--sui-chart-table-focus-row-bg` | `rgba(100,116,139,.14)` | `#f1f5f9` |
| `--sui-chart-table-focus-cell-bg` | `rgba(100,116,139,.22)` | `#e2e8f0` |
| `--sui-chart-table-focus-color` | `#64748b` | `#64748b` |
| `--sui-chart-table-head-bg` | `#14171c (fallback)` | `#eef3fa` |
| `--sui-chart-table-sticky-bg` | `#14171c (fallback)` | `#eef3fa` |
| `--sui-chart-table-input-bg` | `rgba(127,127,127,.08) on focus` | `#ffffff` |
