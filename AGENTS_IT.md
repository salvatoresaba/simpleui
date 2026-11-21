# Guida per coding agent SimpleUI

SimpleUI (SUI) è un piccolo linguaggio reattivo per interfacce: un file `.sui`
compila in un'app web autonoma oppure in un eseguibile desktop. Questo documento
è la guida italiana per gli agenti; la versione canonica in inglese è
[`AGENTS.md`](AGENTS.md).

## Da dove iniziare

1. Esegui `simpleui docs --agent` per un contesto compatto.
2. Leggi `docs/manifest.json` (o `simpleui docs --json`).
3. Scegli una recipe in `docs/AGENT_RECIPES_IT.md` e apri i documenti canonici.
4. Leggi l'esempio più vicino in `examples/`.
5. Scrivi/modifica il file `.sui`, poi esegui
   `simpleui check app.sui --json`.
6. Compila con `simpleui build app.sui --out dist` e usa
   `simpleui serve app.sui --watch` per il browser.

Se il binario non esiste ancora:

```bash
go build -o simpleui.exe ./cmd/simpleui     # Windows
go build -o simpleui ./cmd/simpleui         # Linux/macOS
```

## Lingua

La documentazione canonica, i commenti del codice e le stringhe UI sono in
inglese. Questa guida e gli altri file con suffisso `_IT` sono le copie
italiane: quando la versione inglese cambia, aggiorna anche la copia `_IT`.

## Regole non negoziabili

- **Commit obbligatorio per ogni modifica tracciata.** Al termine di qualsiasi
  modifica ai file tracciati esegui `git add -A` e crea un commit descrittivo
  prima di chiudere il turno. Non lasciare modifiche tracciate non committate.
  Se il lavoro è incompleto o i test falliscono, committa con prefisso `wip:`
  (su un branch quando possibile) così da poter recuperare il lavoro. Gli
  artefatti generati sono ignorati: non forzarne l'aggiunta. Verifica con
  `git status --short` e riporta l'hash del commit.
- **Niente HTML raw.** Testo e attributi sono escapati automaticamente; non
  esiste l'interpolazione HTML raw.
- **Valida i dati non fidati.** Corpi API, file, CSV/JSON, URL e iframe sono
  dati non fidati: controlla ogni `(valore, errore)` e mostra l'errore.
- **CSS dinamico con `cssSafe`.** `style: cssSafe(expr)` è obbligatorio per lo
  stile dinamico; gli stili statici sono fidati.
- **Classi dinamiche filtrate** sulla allow-list dell'app.
- **URL dinamici con `*Safe`.** Usa `http.getSafe/postSafe/getStatusSafe` o
  `ws.connectSafe` con allow-list **letterale**.
- **Iframe sandbox di default.** Se manca `sandbox` viene aggiunto
  `sandbox=""`; se manca `referrerpolicy` viene aggiunto `no-referrer`. Un
  `src` dinamico richiede `srcAllow` e viene validato a runtime. Non usare
  `srcdoc` né markup di fallback.
- **Dati tabella inerti.** Le celle usano `textContent`; usa `table.display` e
  controlla l'errore. Per CSV aperti in Excel/LibreOffice usa
  `;formula=escape`.
- **Capability desktop esplicite.** File/comandi solo con `--allow-dir` /
  `--allow-exec`; valida comunque i nomi.
- **Configurazione di sicurezza rigorosa.** Un file malformato o un valore
  sconosciuto blocca il build.
- **Non modificare l'output generato** (`dist/`, preview, file temporanei).
- **Mantenere gli asset embedded come file separati**: la pagina del designer è
  `assets/designer.html` e i font PDF sono `assets/pdf_fonts.js`, inclusi con
  `go:embed`. Non riportarli dentro stringhe Go raw.
- **Non inventare API**: verifica su `docs/API_REFERENCE.md`,
  `docs/STDLIB.md`, `docs/BUILTINS.md` o compila con `simpleui check`.

## Checklist di consegna dei componenti

Quando si aggiunge un nuovo componente, tag o controllo, consegnarlo come
funzionalità completa, non solo come codice runtime:

1. Aggiungere supporto a compilatore/type-checker, HTML/CSS/JS generati, voce
   nella palette del designer e campi dell'inspector. Mettere il runtime browser
   in `components/<controllo>/runtime.go` come `Runtime` esportato e i test in
   `components/<controllo>/`; la radice conserva solo alias di compatibilità.
2. Aggiungere almeno un esempio eseguibile che mostri **tutte** le funzionalità:
   contenuto, stato, tastiera/accessibilità, flusso nel designer e limiti.
   Preferire `examples/<componente>/app.sui` con commenti e un
   `examples/<componente>/README.md`; si può mantenere anche un piccolo esempio
   `examples/<componente>.sui`.
3. Includere il **codice SUI completo** nel README dell'esempio (in un blocco
   `sui`), non solo uno snippet breve. Spiegare cosa viene dimostrato, come si
   usa il componente, come modificarlo nel designer e come ricompilarlo.
4. Quando possibile, includere una build pronta all'uso (ad esempio
   `examples/<componente>/dist/app.exe`) e `run.ps1` / `run.bat`, più una copia
   del sorgente completo come `dist/app.sui`, così un principiante può avviare
   e ispezionare la demo senza usare comandi del compilatore.
5. Aggiornare documentazione inglese e italiana, manifest, indice e test nello
   stesso cambiamento. Eseguire `simpleui check`, la build pertinente e
   `go test ./...`.

## Mappa del repository

| Percorso | Scopo |
|---|---|
| `cmd/simpleui/main.go` | Entry point eseguibile (`package main`) |
| `*.go` nella radice (`package sui`) | Compilatore, designer, stdlib e CLI importabili |
| `components/<controllo>/` | Runtime browser (`Runtime`), helper e test per controllo |
| `component_runtimes.go` | Alias nella radice per mantenere stabili i call site |
| `testapi.go` | Helper designer/browser esportati per i test esterni |
| `assets/` | Pagina designer, font PDF e asset runtime embedded |
| `scripts/verify.ps1` | Unico comando di verifica (format, build, vet, test, check) |
| `lexer.go`, `parser.go` | Tokenizzazione, AST, caricamento sorgenti |
| `ast.go`, `typechecker.go` | AST, semantica e controlli di sicurezza |
| `jsgen.go`, `htmlgen.go` | Generazione JS/CSS e runtime |
| `dnd.go` | Runtime browser drag and drop (chiavi componenti e file/cartelle) |
| `runtime.go`, `desktop_dialog.go` | Bridge desktop, runtime servito e dialoghi nativi |
| `builtins.go`, `stdlib.go` | Builtin e moduli std |
| `main.go`, `serve.go`, `docs_cli.go` | CLI, dev server, comando docs |
| `gui.go`, `guihtml.go` | Designer visuale, preview, export |
| `mcp.go`, `lsp.go` | Server MCP e language server |
| `Test/` | Fixture e artefatti di test spostati (nessun package Go) |
| `examples/` | Esempi compilabili |
| `docs/` | Navigazione, recipe, API, stdlib, builtin |
| `vscode-hermes-extension/` | Estensione VS Code (Hermes): comandi check e build, `.vsix` pacchettato |
| `SECURITY.md` | Minacce, protezioni e rischi residui |
| `SPEC.md` | Contratti normativi |

## Build e verifica

```bash
# Dopo modifiche al compilatore/linguaggio
go build -o simpleui.exe ./cmd/simpleui
go vet ./...
go test -count=1 ./...

# Dopo modifiche a un file .sui
./simpleui.exe check app.sui --json
./simpleui.exe build app.sui --out dist
./simpleui.exe build app.sui --single --out app.html
./simpleui.exe build app.sui --single --minify --out app.html
./simpleui.exe serve app.sui --watch
```

Per tutte le verifiche in un solo comando usare `pwsh scripts/verify.ps1`
(include i test E2E headless).

Per capability desktop:

```bash
./simpleui.exe build app.sui --target desktop --out dist \
  --allow-dir C:/data --allow-exec helper.exe
```

## Task routing

| Esigenza | Documento | Esempio |
|---|---|---|
| Form locale, stato, tabella | `docs/AGENT_RECIPES_IT.md`, `docs/API_REFERENCE_IT.md` | `examples/todo.sui`, `examples/table_data.sui` |
| Moduli dati puri/in-memory (bytes, bin, files, mime, color, units, print, history, ZIP, CSV, env, capabilities, eventi, log, telemetria) | `docs/STDLIB_DATA_IT.md` | `examples/stdlib_data.sui` |
| Workbook Excel con più tabelle in sheet tab in basso | `docs/API_REFERENCE_IT.md#sheet-tabs-in-stile-excel`, `docs/GUI_EDITOR_IT.md#sheet-tabs-in-stile-excel` | `examples/table_sheets/README.md` |
| Menu bar / navigazione | `docs/API_REFERENCE_IT.md#menu-bar`, `docs/GUI_EDITOR_IT.md#menu-bar-control` | `examples/menu_bar.sui` |
| Accordion / sezioni comprimibili | `docs/API_REFERENCE_IT.md#accordion`, `docs/GUI_EDITOR_IT.md#controllo-accordion` | `examples/accordion/README.md`, `examples/accordion/dist/app.exe` |
| Notifiche toast / feedback | `docs/API_REFERENCE_IT.md#notifiche-toast`, `docs/GUI_EDITOR_IT.md#controllo-toast` | `examples/toast/README.md`, `examples/toast/dist/app.exe` |
| Date/time picker | `docs/API_REFERENCE_IT.md#datetime-picker`, `docs/GUI_EDITOR_IT.md#controllo-datetime-picker` | `examples/datepicker/README.md`, `examples/datepicker/dist/app.exe` |
| Tree / navigazione annidata | `docs/API_REFERENCE_IT.md#tree`, `docs/GUI_EDITOR_IT.md#controllo-tree` | `examples/tree/README.md`, `examples/tree/dist/app.exe` |
| Import/export JSON/XML del tree | `docs/STDLIB_IT.md#tree-module-tree-core`, `docs/API_REFERENCE_IT.md#tree` | `examples/tree/README.md` |
| Grafici da JSON/XML/CSV con preset, scroll/zoom | `docs/API_REFERENCE_IT.md#chart`, `docs/GUI_EDITOR_IT.md#controllo-chart` | `examples/chart/README.md`, `examples/chart/dist/app.exe` |
| Navigazione multi-pagina | `docs/API_REFERENCE_IT.md#pagine-e-navigazione`, `docs/GUI_EDITOR_IT.md#pagine` | `examples/pages/README.md` |
| Dialoghi modali | `docs/API_REFERENCE_IT.md#dialoghi-modali`, `docs/GUI_EDITOR_IT.md#modal-control` | `examples/modal.sui` |
| Loop, spring ed eventi tra grafi | `docs/API_REFERENCE_IT.md#motion-graph-stati-e-trigger`, `docs/GUI_EDITOR_IT.md#animation-panel` | `examples/motion_sequence/README.md` |
| Drag/drop o dialogo file nativo | `docs/GUI_EDITOR_IT.md`, `docs/AGENT_RECIPES_IT.md#drag-and-drop`, `docs/API_REFERENCE_IT.md#drag-and-drop` | `examples/drag_drop.sui`, `examples/file_picker.sui` |
| API JSON remota | `docs/AGENT_RECIPES_IT.md`, `docs/STDLIB_IT.md` | `examples/users/app.sui` |
| HTTP header, multipart upload e WebSocket avanzato | `docs/STDLIB_IT.md`, `docs/API_REFERENCE_IT.md` | `examples/networking_suite/README.md` |
| CRUD remoto | `docs/AGENT_RECIPES_IT.md`, `SECURITY_IT.md` | `examples/users/app.sui` |
| Dashboard polling | `docs/AGENT_RECIPES_IT.md` | `examples/monitor.sui` |
| File/comandi desktop | `docs/AGENT_RECIPES_IT.md`, `SECURITY_IT.md` | `examples/security/desktop_files.sui` |
| Dialoghi nativi file/cartelle | `docs/STDLIB_IT.md#dialog-module-dialog-desktop` | `examples/desktop_picker.sui` |
| XLSX/ODS | `examples/spreadsheet_table/README.md` | `examples/spreadsheet_table/app.sui` |
| Iframe | `docs/AGENT_RECIPES_IT.md`, `SECURITY_IT.md` | `examples/security/untrusted_iframe.sui` |
| Security review | `SECURITY_IT.md`, `examples/security/README.md` | `examples/security/*.sui` |

Indice completo: `docs/INDEX_IT.md`. Manifest: `docs/manifest.json`.

## MCP e automazione

`simpleui mcp` espone:

- tool: `sui_check`, `sui_model`, `sui_compile`, `sui_render`, `sui_docs`;
- risorse `sui://docs/<path>`;
- prompt `build_simpleui_interface` e `review_simpleui_security`.

Configurazione tipica:

```json
{
  "mcpServers": {
    "simpleui": {
      "command": "simpleui",
      "args": ["mcp"]
    }
  }
}
```

Configurazioni pronte da copiare per Claude Code, Cursor e VS Code, con
verifica, risoluzione dei problemi e note di sicurezza, sono in
[`docs/MCP_SETUP.md`](docs/MCP_SETUP.md). Questo repository include anche una
configurazione di progetto funzionante in [`.mcp.json`](.mcp.json).

Documentazione offline:

```bash
simpleui docs --agent
simpleui docs --json
simpleui docs --list
simpleui docs --read api
simpleui docs --search table
simpleui docs --full
```

## Contratto di output

Al termine riporta:

1. file `.sui` modificati (e, se necessario, file del compilatore);
2. esito di `simpleui check` e comando di build/serve usato;
3. decisioni di sicurezza (allow-list, token sandbox, capability desktop);
4. rischi residui o header di deployment ancora necessari;
5. commit git creato per la modifica (hash e messaggio).

Non dichiarare funzionante un'interfaccia se `simpleui check` fallisce o se il
build non è stato eseguito.
