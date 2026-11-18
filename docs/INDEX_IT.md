# Documentazione SimpleUI

> Mappa unica per persone e agenti AI. Se stai aprendo il repository per la
> prima volta, parti da [README_IT](../README_IT.md) o, se sei un agente,
> da [AGENTS_IT](../AGENTS_IT.md).

SimpleUI compila file `.sui` in applicazioni web autonome o eseguibili desktop.
La documentazione è separata tra **guida operativa** (come costruire
un'interfaccia), **riferimento** (sintassi e API) e **specifica** (contratti
interni). Questa pagina indica quale documento leggere e in quale ordine.

## Start here

| Se sei... | Parti da | Poi |
|---|---|---|
| Una persona che vuole provare SimpleUI | [README_IT](../README_IT.md) | [Esempi](../examples/README.md) e [API](API_REFERENCE_IT.md) |
| Un coding agent che deve creare un'interfaccia | [AGENTS_IT](../AGENTS_IT.md) | [Recipe locali/remote](AGENT_RECIPES_IT.md) |
| Un agente che usa MCP | [AGENTS.md](../AGENTS.md#mcp-and-automation) | `simpleui mcp`, risorse `sui://docs/...` |
| Devi correggere un problema di sicurezza | [SECURITY_IT](../SECURITY_IT.md) | [Esempi security](../examples/security/README.md) |
| Devi modificare il compilatore | [SPEC_IT](../SPEC_IT.md) | [DESIGN_IT](../DESIGN_IT.md) |

Ordine consigliato per un agente:

1. `simpleui docs --agent` — contesto compatto e regole.
2. `docs/manifest.json` — indice machine-readable di documenti e recipe.
3. `simpleui docs --read <id>` — documento canonico completo.
4. `simpleui check app.sui --json` — verifica dopo ogni modifica.

## What are you building?

| Scenario | Recipe | Documenti canonici | Esempio |
|---|---|---|---|
| Form locale con stato e tabella | [local-form-table](AGENT_RECIPES_IT.md#local-form--reactive-table) | [API](API_REFERENCE_IT.md), [STDLIB](STDLIB_IT.md) | `examples/todo.sui`, `examples/table_data.sui` |
| Temi e strutture per spinner, checkbox, radio, progress, slider, schede, accordion e toast | [Libreria aspetto](APPEARANCE_LIBRARY.md) | [Editor GUI: aspetto dei controlli](GUI_EDITOR_IT.md#aspetto-dei-controlli) | `examples/appearance_gallery.html` |
| Spinner con immagine / GIF | [API: spinner](API_REFERENCE_IT.md#spinner) | [Editor GUI: modalità immagine/GIF](GUI_EDITOR_IT.md#modalità-immaginegif-dello-spinner) | `examples/spinner_image/README.md` |
| Traduzioni, plurali e cambio lingua | [API: i18n](API_REFERENCE_IT.md#internazionalizzazione-i18n) | [Editor GUI: lingue](GUI_EDITOR_IT.md#lingue) | `examples/i18n/README.md` |
| Filtrare, trasformare e ridurre liste con funzioni | [Builtins: collections](BUILTINS_IT.md#collection-functions) | [API](API_REFERENCE_IT.md) | `examples/collections.sui` |
| Destructuring, spread e argomenti nominati | [API: sintassi](API_REFERENCE_IT.md#destructuring-spread-e-argomenti-nominati) | [SPEC](../SPEC.md) | `examples/syntax_ergonomics.sui` |
| Enum nominali e switch esaustivo | [API: enum](API_REFERENCE_IT.md#enum-e-switch) | [SPEC](../SPEC.md) | `examples/enums.sui` |
| Tipi union, `?` e type guard `is` | [API: union](API_REFERENCE_IT.md#tipi-union-e-type-guard-is) | [SPEC](../SPEC.md) | `examples/unions.sui` |
| Literal types e switch esaustivo sui literal | [API: literal types](API_REFERENCE_IT.md#literal-types) | [SPEC](../SPEC.md) | `examples/literal_types.sui` |
| Generici per funzioni/componenti e parametri rest | [API: generici](API_REFERENCE_IT.md#generici-e-parametri-rest) | [SPEC](../SPEC.md) | `examples/generics.sui` |
| Costanti, computed e campi readonly | [API: costanti](API_REFERENCE_IT.md#costanti-valori-computed-e-campi-readonly) | [SPEC](../SPEC.md) | `examples/reactivity.sui` |
| Batching, error boundary e chiavi delle liste | [API: reattività](API_REFERENCE_IT.md#aggiornamenti-reattivi-batching-error-boundary-e-chiavi-delle-liste) | [SPEC](../SPEC.md) | `examples/reactivity.sui` |
| Formattazione locale con Intl | [API: Intl](API_REFERENCE_IT.md#formattazione-locale-intl) | [BUILTINS](BUILTINS_IT.md#assertion-e-formattazione-locale) | `examples/intl_format.sui` |
| Helper path | [BUILTINS: path](BUILTINS_IT.md#funzioni-path) | [STDLIB](STDLIB_IT.md) | `examples/stdlib_advanced.sui` |
| Crypto hash/UUID/random bytes | [BUILTINS: crypto](BUILTINS_IT.md#funzioni-crypto) | [SECURITY](../SECURITY_IT.md) | `examples/stdlib_advanced.sui` |
| JSON Pointer, JSON Patch, Merge Patch, diff con chiavi e validazione JSON Schema | [STDLIB: json](STDLIB_IT.md#modulo-json-json) | [BUILTINS](BUILTINS_IT.md) | `examples/json_library.sui` |
| Quota storage e sync cross-tab | [BUILTINS: storage](BUILTINS_IT.md#funzioni-storage-cross-target) | [STDLIB](STDLIB_IT.md) | `examples/stdlib_advanced.sui` |
| fs copy/move/size/isDir desktop | [STDLIB: fs](STDLIB_IT.md#file-system-module-fs) | [SECURITY](../SECURITY_IT.md#bridge-desktop) | `examples/fs_operations.sui` |
| XML DOM, XPath, XSD subset, diff/patch/merge con chiavi e conversione | [STDLIB: XML](STDLIB_IT.md#modulo-xml-xml) | [SECURITY](../SECURITY_IT.md#bridge-desktop) | `examples/xml_library.sui` |
| Creare una barra dei menù / navigazione | [API: menu bar](API_REFERENCE_IT.md#menu-bar) | [Editor GUI: menu bar](GUI_EDITOR_IT.md#menu-bar-control) | `examples/menu_bar.sui` |
| Contenuti a schede | [API: tabs](API_REFERENCE_IT.md#tabs) | [Editor GUI: tabs](GUI_EDITOR_IT.md#tabs-control) | `examples/tabs.sui` |
| Sezioni comprimibili / FAQ | [API: accordion](API_REFERENCE_IT.md#accordion) | [Editor GUI: accordion](GUI_EDITOR_IT.md#controllo-accordion) | `examples/accordion/README.md` |
| Feedback toast | [API: toast](API_REFERENCE_IT.md#notifiche-toast) | [Editor GUI: toast](GUI_EDITOR_IT.md#controllo-toast) | `examples/toast/README.md` |
| Suggerimento hover/focus | [API: tooltip](API_REFERENCE_IT.md#proprietà-tooltip) | [Editor GUI: tooltip](GUI_EDITOR_IT.md#proprietà-tooltip) | `examples/tooltip/README.md` |
| Scelta singola ricercabile, opzioni runtime, wildcard | [API: combobox](API_REFERENCE_IT.md#combobox) | [Editor GUI: combobox](GUI_EDITOR_IT.md#controllo-combobox) | `examples/combobox/README.md` |
| Picker nativo data/ora/data-ora | [API: date/time picker](API_REFERENCE_IT.md#datetime-picker) | [Editor GUI: date/time picker](GUI_EDITOR_IT.md#controllo-datetime-picker) | `examples/datepicker/README.md` |
| Albero con rinomina, icone e stili | [API: tree](API_REFERENCE_IT.md#tree) | [Editor GUI: tree](GUI_EDITOR_IT.md#controllo-tree) | `examples/tree/README.md` |
| Import/export JSON e XML del tree | [STDLIB: tree](STDLIB_IT.md#tree-module-tree-core) | [API: tree](API_REFERENCE_IT.md#tree) | `examples/tree/README.md` |
| Grafici da JSON/XML/CSV con preset, scroll e zoom | [API: chart](API_REFERENCE_IT.md#chart) | [Editor GUI: chart](GUI_EDITOR_IT.md#controllo-chart) | `examples/chart/README.md` |
| Controllo Graph / diagrammi a blocchi | [API: Graph](API_REFERENCE_IT.md#controllo-graph-diagrammi-a-blocchi) | [Editor GUI: Graph](GUI_EDITOR_IT.md#controllo-graph) | `examples/graph/README.md`, `examples/state_machine/README.md` |
| Più pagine e navigazione | [API: pagine](API_REFERENCE_IT.md#pagine-e-navigazione) | [Editor GUI: pagine](GUI_EDITOR_IT.md#pagine) | `examples/pages/README.md` |
| Dialoghi modali | [API: modali](API_REFERENCE_IT.md#dialoghi-modali) | [Editor GUI: modal](GUI_EDITOR_IT.md#modal-control) | `examples/modal.sui` |
| Loop, spring ed eventi tra grafi | [API: motion graph](API_REFERENCE_IT.md#motion-graph-stati-e-trigger) | [Editor GUI: animation panel](GUI_EDITOR_IT.md#animation-panel) | `examples/motion_sequence/README.md` |
| Riordinare righe o trascinare/scegliere file/cartelle | [drag-and-drop](AGENT_RECIPES_IT.md#drag-and-drop) | [API](API_REFERENCE_IT.md#drag-and-drop), [BUILTINS](BUILTINS_IT.md#drag-and-drop-functions), [Editor GUI](GUI_EDITOR_IT.md) | `examples/drag_drop.sui`, `examples/file_picker.sui` |
| Analizzare i tipi di file con backend Go | [file-type-analyzer](AGENT_RECIPES_IT.md#analisi-dei-tipi-di-file-con-backend-go) | [STDLIB](STDLIB_IT.md#http-module-http), [API](API_REFERENCE_IT.md#drag-and-drop) | `examples/file_analyzer/README.md` |
| Leggere un'API JSON remota | [remote-api-table](AGENT_RECIPES_IT.md#remote-json-api--table) | [STDLIB](STDLIB_IT.md#http-module-http), [SECURITY_IT](../SECURITY_IT.md) | `examples/users/app.sui` |
| Creare/modificare dati via HTTP | [remote-crud](AGENT_RECIPES_IT.md#remote-crud-form) | [STDLIB](STDLIB_IT.md#http-module-http), [SECURITY_IT](../SECURITY_IT.md) | `examples/users/app.sui` |
| Dashboard con refresh periodico | [polling-dashboard](AGENT_RECIPES_IT.md#polling-dashboard) | [BUILTINS](BUILTINS_IT.md#timer-functions), [API](API_REFERENCE_IT.md) | `examples/monitor.sui` |
| File e comandi su desktop | [desktop-files](AGENT_RECIPES_IT.md#desktop-files-and-commands) | [SECURITY](../SECURITY_IT.md#bridge-desktop), [STDLIB](STDLIB_IT.md#file-system-module-fs) | `examples/security/desktop_files.sui`, `examples/desktop_picker.sui` |
| Importare XLSX/ODS | [spreadsheet-data](AGENT_RECIPES_IT.md#spreadsheet-data) | [Spreadsheet README](../examples/spreadsheet_table/README.md) | `examples/spreadsheet_table/app.sui` |
| Incorporare video/mappe/documenti | [iframe-embed](AGENT_RECIPES_IT.md#iframe) | [API](API_REFERENCE_IT.md#iframe), [SECURITY](../SECURITY_IT.md#iframe-rischi-dettagliati) | `examples/security/untrusted_iframe.sui` |
| Rivedere la sicurezza | [security-review](AGENT_RECIPES_IT.md#security-checklist) | [SECURITY_IT](../SECURITY_IT.md), [security examples](../examples/security/README.md) | `examples/security/*.sui` |

Per l'elenco completo delle recipe, dei tag e dei percorsi, usa
`docs/manifest.json` o `simpleui docs --json`.

## Documentation map

### Iniziare

- [README_IT](../README_IT.md) — installazione, comandi, concetti e struttura.
- [AGENTS_IT](../AGENTS_IT.md) — workflow per agenti, regole e task routing.
- [CLAUDE_IT](../CLAUDE_IT.md) — puntatore per Claude Code.
- [llms.txt](../llms.txt) — discovery compatta per LLM.
- [DOCS index](INDEX.md) — questa pagina.

### Linguaggio e API

- [API_REFERENCE](API_REFERENCE_IT.md) — sintassi essenziale, tabelle, iframe, CLI.
- [SPEC_IT](../SPEC_IT.md) — specifica normativa di linguaggio, compilatore e runtime.
- [BUILTINS](BUILTINS_IT.md) — funzioni senza import.
- [STDLIB](STDLIB_IT.md) — `http`, `ws`, `fs`, `exec`, `json`, `table` e il picker desktop `dialog`.

### Dati e interfacce

- [Tabelle e dati non fidati](API_REFERENCE_IT.md#tabelle-e-dati-non-fidati)
- [Table module](STDLIB_IT.md#table-module-table-core)
- [Spreadsheet XLSX/ODS](../examples/spreadsheet_table/README.md)
- [Iframe sicuri](API_REFERENCE_IT.md#iframe)

### Sicurezza

- [Modello di minaccia](../SECURITY_IT.md#modello-di-minaccia)
- [Protezioni del linguaggio](../SECURITY_IT.md#protezioni-del-linguaggio)
- [Configurazione della sicurezza](../SECURITY_IT.md#configurazione-della-sicurezza)
- [Iframe: rischi dettagliati](../SECURITY_IT.md#iframe-rischi-dettagliati)
- [Rischi residui](../SECURITY_IT.md#rischi-residui)
- [Esempi security](../examples/security/README.md)

### Tooling e architettura

- [Editor GUI: drag and drop](GUI_EDITOR_IT.md) — DnD del designer e dell'applicazione, proprietà, file e picker.
- [CLI](API_REFERENCE_IT.md#cli)
- [Design del compilatore](../DESIGN_IT.md)
- [Roadmap](../TODO_IT.md)
- [MCP e automazione](../AGENTS.md#mcp-and-automation)

### Esempi

- [Esempi root](../examples/README.md)
- [Security](../examples/security/README.md)
- [Spreadsheet table](../examples/spreadsheet_table/README.md)

## Machine interfaces

SimpleUI espone la documentazione in formati stabili e automatizzabili:

| Interfaccia | Uso |
|---|---|
| `docs/manifest.json` | Indice JSON con documenti, sezioni, tag, invarianti e recipe |
| `simpleui docs` | Stampa questa pagina (anche embedded nel binario) |
| `simpleui docs --json` | Manifest completo |
| `simpleui docs --list` | Elenco compatto `id → path` |
| `simpleui docs --read api` | Contenuto di un documento per id |
| `simpleui docs --search table` | Ricerca per titolo, tag e riepilogo |
| `simpleui docs --agent` | Contesto operativo compatto per un coding agent |
| `simpleui docs --full` | Tutta la documentazione in un unico flusso |
| `simpleui mcp` | Tool e risorse MCP per agenti (docs incluse) |

File di integrazione per client diversi: [`AGENTS.md`](../AGENTS.md),
[`CLAUDE.md`](../CLAUDE.md), [`llms.txt`](../llms.txt),
[`.github/copilot-instructions.md`](../.github/copilot-instructions.md),
[`.cursor/rules/simpleui.mdc`](../.cursor/rules/simpleui.mdc).

Gli agenti possono anche leggere direttamente `AGENTS.md` e
`docs/AGENT_RECIPES_IT.md`; i link sono relativi alla root del repository.

## Conventions

- **Fonte di verità**: `SPEC_IT.md`, `SECURITY_IT.md`, `API_REFERENCE_IT.md`,
  `STDLIB_IT.md`, `BUILTINS_IT.md` e `docs/manifest.json`.
- **Docs per compito**: `AGENTS.md` e `docs/AGENT_RECIPES_IT.md` spiegano *come*
  costruire; i riferimenti spiegano *cosa* è disponibile.
- **Non inventare API**: se una funzione non è nei riferimenti, verifica nel
  codice o compila con `simpleui check`.
- **Link stabili**: preferisci link a file e heading esistenti; il test
  `docs_nav_test.go` controlla i link locali e la coerenza del manifest.
- **Sicurezza first**: ogni esempio che consuma dati esterni deve mostrare
  anche il percorso di errore e le protezioni (`cssSafe`, `*Safe`, `srcAllow`,
  `table`, capability desktop).
- **Lingua**: la documentazione canonica è in inglese; le copie italiane usano il suffisso `_IT` e vanno aggiornate insieme alla versione inglese.





