# SimpleUI (SUI) — Specifica tecnica (Italiano)

> La specifica normativa completa è [`SPEC.md`](SPEC.md) in inglese. Questa è
> la sintesi italiana aggiornata; in caso di divergenza vale `SPEC.md`.

**Versione:** 0.1.0-alpha  
**Target:** web e desktop.

## 1. Struttura del repository

La radice del modulo è un package Go importabile `sui` con lexer, parser, AST,
typechecker, codegen JS/HTML/CSS, runtime desktop, CLI, GUI, LSP, MCP e
documentazione. L'eseguibile è un sottile wrapper `package main` in
`cmd/simpleui`. I runtime browser e i test dei controlli stanno in
`components/<controllo>/`; `component_runtimes.go` espone gli alias usati dal
compilatore. I test trasversali restano nella radice, quelli dei controlli sono
`package <controllo>_test`. La pagina del designer e i font PDF sono file in
`assets/` inclusi con `go:embed`; `scripts/verify.ps1` esegue tutte le verifiche.

## 2. Linguaggio SUI

- File `.sui`, statement separati da newline; blocchi con `{}`.
- Tipi: `int`, `float`, `string`, `bool`, `null`, `any`, `array[T]`,
  `map[string]T`, `struct`.
- `state` per lo stato reattivo; `let` per valori locali; `fn`/`async fn` per
  funzioni; `view` per le viste e i componenti parametrici.
- Espressioni: operatori aritmetici/logici/confronto, ternary,
  interpolazione `{expr}`, chiamate, accesso a membri/indici, range `lo..hi`.
- `if`/`else`, `while`, `loop` su range/array/map, `break`, `continue`,
  `return`.
- Le funzioni I/O sono `async`; le chiamate restituiscono `(valore, errore)`.
- Gli import utente non sono ancora supportati; è supportato
  `import { ... } from "std"`.

## 3. Viste e sicurezza

- Testo e attributi sono escapati automaticamente; non esiste HTML raw.
- `style:` dinamico richiede `cssSafe(...)`.
- Le classi dinamiche sono filtrate sulla allow-list dell'app.
- Gli URL dinamici richiedono `*Safe` + allow-list letterale.
- Le celle tabella usano `textContent`.
- Gli iframe hanno sandbox/referrer di default, `srcAllow` per URL dinamici e
  validazione runtime con messaggio visibile.
- Pagine: ogni `view` root non invocata è una pagina. Le app multi-pagina
  tengono tutte le pagine nel DOM ma mostrano la prima; il builtin
  `navigate(name)` (senza import) cambia pagina ed è un no-op nelle app a
  pagina singola o con nomi sconosciuti. Lo stato è condiviso.
- Modali: `modal` (titolo, `name`) contiene contenuto e pulsanti
  `modalbutton` (etichetta, `variant`, `close`, `onClick`). `modal.open(name)`/
  `modal.close(name)` mostrano/nascondono il dialogo; il click fuori, Escape o
  un pulsante lo chiudono di default. Il focus entra e, se disponibile, ritorna
  all'opener.
- Tabs: `tabs` (nome accessibile) contiene figli diretti `tab`; il testo di
  `tab` è l'etichetta del pulsante e i figli sono il pannello. La prima scheda
  è attiva; click e frecce cambiano pannello; la selezione viene ripristinata
  dopo un re-render di pagina.
- Menu bar: `menubar` (nome accessibile) contiene `menu` (etichetta del
  pulsante) e `menuitem` (etichetta, `link`, `onClick`). Un `menu` annidato in
  un altro `menu` è un sottomenu e può annidarsi ulteriormente. Il runtime
  apre/chiude i popup, supporta Escape e click esterno. Etichette escapate e
  link passano dall'URL guard.
- Drag and drop: `dragKey`/`dropKey` e gli eventi `onDrag*`/`onDrop`;
  `dropFiles:` richiede uno state stringa e riceve i percorsi. `pickFiles:`
  apre il dialogo file nativo da un `file`, `button` o altro elemento
  cliccabile; `accept`/`multiple`/`directory` configurano la selezione.
  `pickMode: "fs"` usa la File System Access API (Chromium, contesto sicuro,
  sola lettura) e permette di conservare il nome anche di una cartella vuota;
  `auto` ripiega sul picker nativo. `directory: true` restituisce tutti i file
  della cartella e `drag.folder()` espone il nome della cartella comune. I file sono solo metadati (2.000 file,
  profondita 16, 4.096 caratteri per percorso); il browser non espone percorsi
  assoluti e i nomi restano dati non fidati.
- La configurazione di sicurezza è rigorosa: JSON/campi/valori non validi
  bloccano il build.

## 4. Standard library

Moduli: `http`, `ws`, `fs`, `exec`, `json`, `table` (core), `dialog` (desktop). Le funzioni con due
valori restituiscono `(value, error)`; le funzioni fire-and-forget restituiscono
`null`. HTTP dinamico: `getSafe`, `getStatusSafe`, `postSafe`, `postFilesSafe`;
upload file: `postFiles`; WebSocket:
`connectSafe`. Table: `display`, `data`, `header`, `select`, `slice`,
`importText`, `exportText`, `json`, `clear`. Dialog desktop: `pickFile`,
`pickFiles`, `pickDirectory` (dialoghi nativi, solo target desktop).

## 5. Compilatore e target

Fasi: lexer → parser → typechecker → codegen. Gli errori confluiscono in
`DiagList` con file/riga/colonna/tipo/messaggio/hint.

- **Web:** `index.html`, `style.css`, `app.js`; `--single` produce un HTML con
  CSS/JS/runtime inline; `--minify` è conservativo.
- **Desktop:** eseguibile singolo con asset `go:embed`, micro-server
  `127.0.0.1`, bridge con capability e apertura browser. La WebView nativa è un
  follow-up.

## 6. CLI

Comandi: `version`, `help`, `parse`, `check`, `fmt`, `docs`, `build`, `serve`,
`gui`, `lsp`, `mcp`.

`check --json` restituisce `{ok, file, diagnostics}` ed esce con codice 1 in caso
di errori. `docs` espone la documentazione embedded; `build --single` e
`--target desktop` producono rispettivamente un HTML unico e un eseguibile.

## 7. Runtime generato

`index.html` contiene `#app`; `app.js` contiene il runtime e le funzioni
`__SUI_views`, `__sui_render`, `__sui_bindView`, `__SUI_STD`,
`__SUI_TBL` e i renderer delle viste. Gli id `data-sui-id` seguono l'ordine DFS
delle viste root. Il target desktop sostituisce le intrinsic `fs`/`exec` con il
bridge e inietta un token per-run.

## 8. Sicurezza del target desktop

- Token casuale per-run, confronto in tempo costante.
- Host loopback obbligatorio, richieste cross-site rifiutate.
- `X-Frame-Options: DENY` + `frame-ancestors 'none'`.
- Allow-list esatta dei comandi, `exec.Command` senza shell.
- Allow-dir con risoluzione dei symlink.
- Nessun bridge nel target web.

## 9. Stato

La fonte aggiornata e verificata è [`TODO.md`](TODO.md). Le API non ancora
disponibili (HTTP headers/multipart, WebSocket reconnect, fs avanzato, crypto)
non devono essere considerate stabili. Le source map JS/CSS sono opt-in con
`--sourcemap`.
