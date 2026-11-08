# SimpleUI (SUI)

SimpleUI è un linguaggio DSL reattivo per la creazione di interfacce utente, compilato in applicazioni web e desktop standalone scritte in Go.

## Cos'è SimpleUI

SimpleUI ti permette di scrivere UI in un file `.sui` utilizzando una sintassi semplice e dichiarativa. Il compilatore traduce automaticamente il tuo codice in **HTML, CSS e JavaScript vanilla**, senza dipendenze esterne.

Le principali caratteristiche:
- **Reattivo**: le assegnazioni a `state` aggiornano automaticamente l'interfaccia
- **Drag and drop / file picker**: riordina elementi, trascina file/cartelle o usa il dialogo file nativo
- **Menu bar**: inserisci una barra dei menù configurabile con sottomenu, link e azioni
- **Tabs**: inserisci set di schede funzionanti con pannelli cliccabili e navigazione da tastiera
- **Preferenze editor**: font, dimensione, interlinea, tab, wrapping e numeri di riga personalizzabili
- **Completamento contestuale**: Ctrl+Space mostra dichiarazioni, elementi, attributi, handler ed espressioni vicino al cursore
- **Corpo funzioni indentato**: le funzioni generate mantengono l'indentazione annidata dopo le modifiche al designer
- **Evidenziazione simboli**: nomi delle funzioni e variabili con colori distinti in dichiarazioni, riferimenti e interpolazioni
- **Undo/redo unificato**: testo, proprietà, elementi e azioni del designer condividono una cronologia con pulsanti, Ctrl+Z e Ctrl+U
- **Stili effettivi nell'inspector**: larghezza, altezza, colori e altri campi riflettono valori dichiarati, ereditati, inline e calcolati
- **Albero elementi comprimibile**: le righe contenitore si espandono e comprimono con una freccia
- **Controllo Accordion**: sezioni espandibili annidate con intestazioni accessibili e stato aperto ripristinato
- **Controllo Tree**: selezione dell'intero albero, modifica strutturale da menu contestuale, icone leaf/parent/open, simboli di ramo aperto/chiuso, picker icone filtrato, colonne tree-grid e `onTreeChange` con `tree.lastChange`
- **Controllo Table**: colonne dichiarative, ordinamento, filtri, paginazione, selezione riga/colonna/cella, celle editable validate, eventi tabella, overlay Table settings a tab, più tabelle in sheet tab in basso stile Excel, creazione tabelle e modifica struttura righe/colonne a runtime
- **Controllo Chart**: grafici SVG barre, barre orizzontali/stacked/grouped, linee/step/smooth, area, combo, torta/ciambella, scatter, istogramma, waterfall, radar, gauge, funnel, heatmap, candlestick, treemap, box plot, bullet, dumbbell e Gantt da JSON/XML/CSV con controllo compatibilita, scroll/zoom, preset, editor dati ridimensionabile, tabelle accessibili sincronizzate, inserimento/modifica/eliminazione righe, selezione righe/colonne/celle, stepper numerici visibili, menu contestuale copia/export, export della selezione e scrolling virtuale, critical path, festivi, weekend configurabili e timezone IANA
- **Badge tipo esatto**: ogni riga della struttura mostra il tipo preciso dell'elemento (`div`, `treeitem`, `chart`, ...)
- **Notifiche toast**: feedback testuale sicuro con varianti info/success/warning/error, timeout e host opzionale
- **Date/Time picker**: controlli nativi data, ora e data/ora con binding di stato, min/max, clear e validazione
- **Combobox**: lista ricercabile con aggiornamenti runtime, escape e ricerca wildcard
- **Pagine**: più view con `navigate("pagina")` e chip pagina nell'editor
- **Dialoghi modali**: dialoghi configurabili con pulsanti primary/danger e `modal.open()` / `modal.close()`
- **Componenti**: crea componenti riutilizzabili con parametri, slot e stato per istanza
- **Tipizzato**: type checking statico con inferenza per `state` e `let`
- **Multi-target**: genera app web o eseguibili desktop standalone
- **Zero-dipendenze**: il codice generato funziona ovunque (browser moderno)
- **Sicuro per desktop**: bridge nativo con permessi basati su capability e dialoghi file/cartelle nativi

## Requisiti

- [Go 1.24.2+](https://go.dev/dl/) per compilare il compilatore e le build desktop
- Un browser moderno per eseguire le app web
- Windows/Linux/macOS per le build desktop

## Compilazione del compilatore

Clona o scarica il progetto e compila il binario:

```bash
# Clona il repository (se da git)
git clone <repo-url>
cd GO_PROJECT_LANG

# Compila il compilatore
go build -o simpleui

# Oppure su Windows
go build -o simpleui.exe
```

Verifica l'installazione:

```bash
./simpleui version
# Output: simpleui 0.1.0-alpha
```

## Uso rapido

### 1. Creare un'app di esempio

Crea un file `counter.sui`:

```sui
state count = 0

fn increment() {
  count = count + 1
}

fn decrement() {
  count = count - 1
}

view App {
  col {
    h1 "Contatore"
    text "Valore: " + str(count)
    row {
      button "Decrementa" onClick: decrement
      button "Incrementa" onClick: increment
    }
  }
}
```

### 2. Compilare per web

```bash
./simpleui build counter.sui --target web --out dist
```

Verranno generati:
- `dist/index.html`
- `dist/style.css`
- `dist/app.js`

Apri `dist/index.html` nel browser per vedere l'app.

Per esportare tutto in **un unico file HTML** (CSS, JS, runtime e font del
designer inline):

```bash
./simpleui build counter.sui --single --out app.html
./simpleui build counter.sui --single --minify --out app.html
```

Con `--out dist` il file diventa `dist/app.html`. `--minify` è conservativo:
rimuove commenti e indentazione senza cambiare la semantica.

### 3. Compilare per desktop

```bash
./simpleui build counter.sui --target desktop --out dist
./dist/app.exe  # Windows
# oppure ./dist/app su Linux/macOS
```

L'app si aprirà automaticamente nel browser di default, oppure puoi usare `--no-open` per avviarla in modalità headless.

### 4. Modalità sviluppo

Usa il dev server con live-reload:

```bash
./simpleui serve counter.sui --watch
```

Apri l'URL mostrato nel terminale per vedere l'app con aggiornamenti automatici ad ogni modifica del file `.sui`.

### 5. Strumenti aggiuntivi

```bash
# Verifica parse + typecheck (--json per agenti/CI)
./simpleui check counter.sui --json

# Visualizza AST
./simpleui parse counter.sui

# Formatta il codice (canonical)
./simpleui fmt counter.sui --write

# Documentazione embedded (indice, contesto agente, ricerca, lettura)
./simpleui docs
./simpleui docs --agent
./simpleui docs --search table
./simpleui docs --read api

# Editor visuale GUI (comando diretto)
./simpleui gui

# Language Server per editor (VSCode, Neovim, etc.)
./simpleui lsp

# MCP server per agenti (tool, risorse docs, prompt)
./simpleui mcp
```

#### Editor grafico GUI

L'editor grafico permette di creare e modificare file `.sui` in modo visivo, con preview live, tree editor drag&drop, styler CSS e molto altro.

Da `File` puoi esportare il sorgente (`.sui`), la build web (`.zip` con HTML,
CSS e JS) oppure **Export single HTML…**: un solo `app.html` con CSS, JS,
runtime e font caricati nel designer inline.

In `File → Build desktop executable…` puoi compilare l'eseguibile desktop
direttamente dall'editor. Il dialog accetta la directory di output, le
directory/comandi consentiti dal bridge e il download opzionale; prima di
avviare la build verifica Go 1.24.2+, il sorgente, la scrivibilità della
directory e le capability. Il log di build e il percorso locale dell'eseguibile
restano visibili nel dialog.

**Windows:**

Fai doppio clic su `GUI_editor.bat` oppure esegui da terminale:

```cmd
GUI_editor.bat
```

**Linux/macOS:**

Rendi eseguibile lo script (solo la prima volta):

```bash
chmod +x GUI_editor.sh
```

Avvialo:

```bash
./GUI_editor.sh
```

In alternativa, puoi sempre usare il comando diretto:

```bash
./simpleui gui
```

Il browser si aprirà automaticamente all'indirizzo locale mostrato nel terminale.

## Concetti fondamentali

### State e reattività

```sui
state nome = "Mario"
state contatore = 0

fn cambia() {
  nome = "Luigi"      // Aggiorna automaticamente la UI
  contatore++
}
```

### Componenti

```sui
view Card(titolo: string) {
  col {
    h2 titolo
    children  // Contenuto passato dal chiamante
  }
}

view App {
  Card titolo: "Benvenuto" {
    text "Questo è il contenuto dello slot"
  }
}
```

### Componenti con stato per istanza

```sui
view Contatore(start: int) {
  state count = start
  
  fn inc() { count++ }
  
  row {
    text str(count)
    button "+" onClick: inc
  }
}
```

### Chiamate asincrone e API

```sui
import { http, json } from "std"

state dati = []

async fn carica() {
  body, err = http.get("https://api.example.com/users")
  if err != null {
    return
  }
  parsed, perr = json.parse(body)
  if perr == null {
    dati = parsed
  }
}

view App {
  onLoad: carica
  list item in dati {
    text item.name
  }
}
```

### Tabelle popolate da dati (`table.display` / `table.data`)

Dai un nome alla tabella (inspector → Data, oppure `name:` nel codice). Le righe
di esempio restano come modello grafico:

```sui
table name: "utenti" {
  thead { trow { tcell "ID" head: true ... } }
  tbody { trow { tcell "1" ... } }
}
```

Poi riempila con qualunque payload tabellare e rileggi i valori visibili:

```sui
fn carica() -> null {
  n, err = table.display("utenti", [["1", "Alice"], ["2", "Bob"]])
  // accetta anche testo JSON, array di oggetti o {columns, rows}
}

fn caricaCSV() -> null {
  n, err = table.display("utenti", "ID,Nome\n1,Alice\n2,Bob")
  // CSV/TSV/SSV/PSV rilevati automaticamente, con campi tra virgolette
}

fn caricaSSV() -> null {
  n, err = table.importText("utenti", "1;Alice\n2;Bob", "ssv;header=0")
  // formato esplicito e opzioni: header, sep, quote, trim
}

fn esporta() -> null {
  csv, err = table.exportText("utenti", "csv")   // csv, tsv, ssv, psv, json, jsonl
  righe, e2 = table.data("utenti")
  parte, e3 = table.select("utenti", [0, 2])     // solo alcune colonne
  pagina, e3b = table.slice("utenti", 0, 50)     // solo un blocco di righe
  body, e4 = table.json("utenti")                // JSON pronto da inviare
}

async fn invia() -> null {
  body, e = table.json("utenti")
  out, e2 = http.post("/api/utenti", body)
}
```

- `table` è un modulo core: non serve importarlo.
- Le righe generate copiano classi e stili delle righe/celle di esempio; caption,
  preset, wrapper di scroll e flag editable/fixed restano invariati.
- I testi vengono riconosciuti: prima JSON, poi JSON Lines, poi testo delimitato.
- Delimitatori auto-rilevati: `,` `;` tab `|`; gestisce virgolette, `""`, CRLF e BOM.
- La prima riga del testo delimitato è l'intestazione; se la tabella non ha
  `<thead>`, resta visibile come prima riga del corpo.
- `exportText` include l'intestazione (salvo `header=0`) e mette tra virgolette i
  campi che contengono delimitatore, virgolette o a capo.
- Per CSV aperti in Excel/LibreOffice: `table.exportText("utenti",
  "csv;formula=escape")` neutralizza le formule (`= + - @`).
- Le celle sono testo, mai HTML. Limiti runtime: 100.000 righe, 10.000 colonne,
  20 MB di testo per chiamata.
- I dati restano dopo i re-render dello stato; le celle editable mantengono le
  modifiche anche dopo un re-render.

### Excel / ODS

`.xlsx` e `.ods` sono archivi binari: non vanno passati direttamente a
`table.display`. L'esempio `examples/spreadsheet_table` include `sheet2json`,
un convertitore Go (solo standard library) che li trasforma in JSON/CSV/JSONL:

```bash
sheet2json data.xlsx                 # JSON pronto per table.display
sheet2json data.ods -format csv
sheet2json -serve -dir ./data -web-dir ./web   # endpoint /api/sheet
```

L'app può usare `exec.run("sheet2json.exe", [path])` (desktop) oppure
`http.get("/api/sheet?file=...")` (web) e passare il risultato a
`table.display`. Dettagli in `examples/spreadsheet_table/README.md`.

### Accesso file/exec (solo desktop)

Per usare `fs` ed `exec` nelle build desktop devi concedere i permessi espliciti:

```bash
./simpleui build app.sui --target desktop --out dist \
  --allow-exec "node" \
  --allow-dir "C:/dati"
```

```sui
import { fs } from "std"

async fn leggi() {
  contenuto, err = fs.read("C:/dati/file.txt")
  // ...
}
```

### Sicurezza dei dati non fidati

- Le interpolazioni `{...}` in testo e attributi sono **escapate
  automaticamente**: i dati non possono iniettare markup.
- `href`/`src`/`link` filtrano gli schemi: `javascript:`, `data:`, `vbscript:`
  e simili diventano `#`; gli schemi attivi **statici** sono rifiutati già in
  compilazione.
- `style:` dinamico richiede `cssSafe(...)`; il sanitizer blocca `url()`,
  posizionamento/overlay e proprietà rischiose.
- Le classi dinamiche sono filtrate sulla allow-list delle classi dell'app
  (CSS + `class:` statiche): `"active evil"` diventa `"active"` se `evil` non è
  dichiarata.
- Gli URL dinamici passano da `http.getSafe`/`postSafe`/`getStatusSafe` (o
  `ws.connectSafe`) con una allow-list letterale: niente SSRF, redirect o
  credenziali. `http.get` resta per URL letterali scritti nel codice.
- `table.*` scrive le celle come testo (`textContent`), mai come HTML; JSON,
  CSV e fogli convertiti non possono eseguire markup.
- I CSV destinati a Excel/LibreOffice possono neutralizzare le formule:
  `table.exportText("t", "csv;formula=escape")`.
- Il bridge desktop usa token casuale, Host loopback, allow-list per comandi e
  directory (con risoluzione dei symlink), nessuna shell.

Modello completo, limiti e rischi residui: [`SECURITY_IT.md`](SECURITY_IT.md).
Le policy di compilazione sono modificabili con `simpleui.security.json`
(accanto al `.sui`) o `--security <file>`: default severi, override esplicito.
Gli iframe hanno `sandbox=""` e `referrerpolicy="no-referrer"` per default;
`src` dinamico richiede un'allow-list `srcAllow` e viene validato a runtime.
Esempi eseguibili: [`examples/security`](examples/security/README.md).

## Struttura del progetto

```
.
├── lexer.go, parser.go    # Tokenizzazione e parsing
├── ast.go, typechecker.go # AST e type checking
├── jsgen.go, htmlgen.go   # Generazione codice
├── dnd.go                 # Runtime drag and drop nel browser
├── runtime.go             # Runtime e bridge desktop
├── builtins.go, stdlib.go # Librerie built-in e standard
├── main.go                # CLI
├── gui.go                 # Editor visuale
├── lsp.go                 # Language Server
└── examples/              # Esempi di utilizzo
```

## Esempi inclusi

- `examples/counter.sui` - Contatore semplice
- `examples/users/` - Fetch API JSON e tabella reattiva
- `examples/userdata_table/` - Esecuzione comando esterno via bridge desktop
- `examples/table_data.sui` - Popolamento ed estrazione dati di una tabella
- `examples/spreadsheet_table/` - Conversione XLSX/ODS in JSON/CSV + tabella
- `examples/security/` - Testo, tabelle, URL, file e CSV non fidati

## Documentazione

- [`docs/INDEX_IT.md`](docs/INDEX_IT.md) - Mappa unica della documentazione, per persone e agenti.
- [`AGENTS_IT.md`](AGENTS_IT.md) - Guida operativa per coding agent: regole, comandi, task routing.
- [`docs/AGENT_RECIPES_IT.md`](docs/AGENT_RECIPES_IT.md) - Pattern per applicazioni locali, remote, desktop e iframe.
- [`docs/GUI_EDITOR_IT.md`](docs/GUI_EDITOR_IT.md) - Drag and drop del designer e dell'applicazione, proprietà dell'inspector e picker.
- [`docs/manifest.json`](docs/manifest.json) - Indice machine-readable di documenti, sezioni, tag e recipe.
- [`llms.txt`](llms.txt) - Discovery compatta per LLM.
- [`docs/BUILTINS_IT.md`](docs/BUILTINS_IT.md) - Tutte le funzioni built-in (stringhe, array, math, time, regex, storage, etc.)
- [`docs/STDLIB_IT.md`](docs/STDLIB_IT.md) - Standard library (`http`, `ws`, `fs`, `exec`, `json`, `table`, `dialog` desktop)
- [`docs/API_REFERENCE_IT.md`](docs/API_REFERENCE_IT.md) - Riferimento completo API e sintassi
- [`SECURITY_IT.md`](SECURITY_IT.md) - Modello di minaccia, escaping, bridge, limiti e rischi residui
- [`SPEC_IT.md`](SPEC_IT.md) - Specifica tecnica del linguaggio
- [`DESIGN.md`](DESIGN.md) - Architettura e design del compilatore
- [`TODO_IT.md`](TODO_IT.md) - Roadmap di sviluppo

Per un agente:

```bash
simpleui docs --agent
simpleui docs --read <id>
simpleui check app.sui --json
simpleui mcp
```

## Licenza

Questo progetto è in stato **0.1.0-alpha**. Usa con consapevolezza in ambienti di sviluppo/sperimentazione.

