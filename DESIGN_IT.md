# Documento di Progetto: Linguaggio "SimpleUI" (SUI)

This file is the original design document that scopes the project. The
technical contracts derived from it (and the decisions taken while
implementing) live in `SPEC.md`.

---

## 1. Visione e Principi Fondamentali

**Obiettivo:** Creare un linguaggio di programmazione per lo sviluppo rapido di
interfacce grafiche (front-end) per applicazioni desktop locali e web remote. Il
focus è su tool, dashboard e GUI per wrapping di CLI o dati in tempo reale.

**Principi:**
1. **Zero Dipendenze:** L'utente non deve installare Node.js, npm, Python o
   librerie C++. Il compilatore è un singolo binario.
2. **Un codice, due target:** Lo stesso file sorgente (`.sui`) compila un'app
   desktop nativa (tramite WebView) o un'app web (HTML/JS/CSS statico).
3. **Sicurezza prima di tutto (Static Typing):** Tipizzazione statica con
   inferenza per prevenire errori a runtime.
4. **Sviluppo guidato dai dati (Reattività Implicita):** Il programmatore
   aggiorna le variabili, la UI si aggiorna da sola.
5. **Batterie Incluse:** Accesso nativo a File System, HTTP, WebSockets, JSON e
   CLI senza importare librerie esterne.

---

## 2. Specifiche del Linguaggio

### 2.1 Sintassi e Struttura
Un file `.sui` è diviso in due concetti logici, ma può mescolarli: la **Logica**
(stato, funzioni, moduli) e le **Viste** (dichiarazione della UI).

### 2.2 Sistema di Tipizzazione
- **Statica con Inferenza:** Le variabili dichiarano il tipo implicitamente al
  primo assegnamento.
- Esempio: `state counter = 0` (il compilatore deduce `int`). Se
  successivamente si fa `counter = "test"`, il compilatore restituisce errore.
- Tipi supportati: `int`, `float`, `string`, `bool`, `array[T]`, `map[Key]Value`.

### 2.3 Gestione dello Stato e Reattività
- Le variabili di stato devono essere dichiarate con la keyword `state`.
- Quando una variabile `state` viene riassegnata, il compilatore genera un
  trigger che aggiorna il DOM.
- **Regola di scope:** Le modifiche allo stato fuori dalle funzioni di callback
  devono avvenire all'interno di funzioni specifiche.

### 2.4 Gestione degli Errori e Asincronia
- Ispirato a Go. Niente `try/catch`.
- Le funzioni che eseguono I/O (rete, file, CLI) devono usare la keyword `async`.
- **Sintassi:** `risultato, errore = funzione()`. Se `errore != null`,
  l'operazione è fallita.

### 2.5 Moduli
- Parola chiave: `import`.
- Permette di dividere l'app in componenti:
  `import { Button } from "./ui_components.sui"`.
- Permette di importare la libreria standard:
  `import { http, ws, fs } from "std"`.

---

## 3. Architettura del Compilatore (Scritto in Go)

Il compilatore è il cuore del sistema. Riceve un file `.sui` e produce gli
artefatti necessari. È scritto interamente in Go per garantire cross-compilation
e distribuzione come singolo binario.

### 3.1 Fasi del Compilatore
1. **Lexer (Tokenizzazione):** Legge il testo sorgente e lo scompone in token
   (parole chiave, stringhe, operatori).
2. **Parser (AST):** Analizza i token e costruisce l'Abstract Syntax Tree,
   validando la grammatica del linguaggio.
3. **Type Checker:** Attraversa l'AST per verificare che i tipi assegnati siano
   coerenti, applicando l'inferenza.
4. **Code Generator:**
   - Traduce l'AST in codice JavaScript Vanilla (per la logica).
   - Traduce l'AST in HTML/CSS (per la UI).
   - Genera il codice Go aggiuntivo (per il bridge desktop).

### 3.2 I Target di Compilazione

**Target Web (`simpleui build --target web`)**
- Output: Una cartella contenente `index.html`, `style.css`, `app.js`.
- Comportamento: Tutte le chiamate native (es. `fs.read`) vengono transpilate in
  chiamate API Fetch verso un backend (presumendo che l'app venga ospitata su un
  server Node/Go/PHP gestito dall'utente).

**Target Desktop (`simpleui build --target desktop`)**
- Output: Un singolo file eseguibile (es. `app.exe` o `app` su Linux/Mac).
- Implementazione: Il compilatore Go genera un file `main.go` che incorpora
  (`go:embed`) HTML/CSS/JS e avvia un micro-server web invisibile su
  `127.0.0.1:porta`. L'eseguibile usa le API WebView native del SO (Edge WebView2
  su Windows, WebKit su Mac/Linux) per renderizzare l'interfaccia in una finestra
  nativa.

---

## 4. Gestione della Logica (Il Bridge Nativo)

La logica dell'utente viene transpilata in JavaScript. Tuttavia, sul desktop,
JavaScript è rinchiuso nella sandbox della WebView. Per accedere al sistema
operativo (File System, Esecuzione CLI), si usa un'architettura a Bridge.

### 4.1 Come funziona il Bridge Desktop
1. Il codice utente `output, err = exec.run("ls -la")` viene tradotto dal
   compilatore.
2. Il compilatore genera codice JS che esegue una richiesta `fetch` verso
   `http://127.0.0.1:porta/__bridge__/exec`.
3. Il server Go incorporato nel binario desktop intercetta la richiesta, esegue
   il vero comando `ls -la` tramite il pacchetto `os/exec` di Go, cattura
   l'output e lo rispedisce al JS in formato JSON.

### 4.2 Standard Library (Batterie Incluse)
Queste funzioni sono parte del linguaggio e il compilatore sa come tradurle:
- `http.get(url)`, `http.post(url, data)`: Tradotte in `fetch` JS.
- `ws.connect(url)`: Tradotte in `WebSocket` JS nativo.
- `fs.read(path)`, `fs.write(path, data)`: Su web usano IndexedDB/API; su
  desktop passano per il Bridge Go.
- `exec.run(cmd, args)`: Disabilitata su web, tradotta via Bridge Go su desktop.
- `json.parse(str)`, `json.stringify(obj)`: Tradotte in `JSON.parse` /
  `JSON.stringify`.

---

## 5. Soluzioni Tecniche ai Problemi Principali

### Problema 1: Reattività Implicita in JS Vanilla
- **Problema:** JavaScript non è reattivo di default. Come fa la UI ad
  aggiornarsi quando si cambia una variabile?
- **Soluzione:** Il Code Generator in Go analizza le assegnazioni alle variabili
  `state`. Se vede `state counter = 0`, genera in JS una variabile associata a
  una funzione di rendering. Ogni volta che nel codice compare `counter = ...`,
  il compilatore lo trasforma in `setCounter(...)` (pattern di reattività simile
  a Svelte), aggiornando l'`innerHTML` del nodo DOM interessato.

### Problema 2: Valori Multipli (Multiple Return Values) in JS
- **Problema:** Il linguaggio usa `val, err = funzione()`, ma JS non supporta
  valori multipli di ritorno.
- **Soluzione:** Il compilatore traduce la chiamata in una `Promise` che risolve
  un array. `val, err = await funzione()` diventa
  `let [val, err] = await __funzione_bridge()`. Le funzioni intrinseche del
  bridge restituiranno sempre array `[null, error]` o `[value, null]`.

### Problema 3: Sicurezza e Permessi Desktop
- **Problema:** Se l'app desktop esegue un comando CLI locale, un utente
  malintenzionato potrebbe sfruttare l'app per eseguire comandi dannosi (Command
  Injection).
- **Soluzione:** Il Bridge Go implementerà un sistema di "Capability-based
  security". L'utente potrà dichiarare nel file di configurazione (o nel codice)
  quali path o comandi sono permessi. Inoltre, le chiamate al bridge locale
  devono includere un hash/segreto generato a runtime noto solo al binario Go e
  alla WebView.

---

## 6. Roadmap di Sviluppo (Step by Step)

**Fase 1: Sviluppo del Lexer e Parser (Settimana 1-2)**
- Scrivere in Go la grammatica formale (magari usando `goyacc` o scrivendo un
  parser a discesa ricorsiva manuale).
- Obiettivo: Leggere un file `.sui` e stampare a video l'AST in formato JSON per
  verificare la correttezza.

**Fase 2: Code Generator HTML/CSS (Settimana 3)**
- Tradurre i blocchi `view` in file `index.html` e `style.css` statici.
- Obiettivo: Aprendo il file HTML in un browser, si deve vedere l'interfaccia
  priva di logica.

**Fase 3: Code Generator Logica JS (Settimana 4-5)**
- Implementare l'inferenza dei tipi.
- Tradurre funzioni, `if`, `loop` e `state` in JavaScript.
- Implementare la reattività implicita (trasformazione delle assegnazioni in
  funzioni di update).

**Fase 4: Il Desktop Wrapper e Bridge (Settimana 6-7)**
- Scrivere il template Go `main.go` che incorpora gli asset (`go:embed`).
- Implementare il server HTTP locale e l'apertura della WebView (usando `cgo`
  con i binding nativi o librerie Go minime per WebView).
- Mappare le chiamate API `fs`, `exec`, `http` al server locale.

**Fase 5: Moduli e Standard Library (Settimana 8)**
- Implementare il risolutore di `import`.
- Scrivere le implementazioni native della standard library `std`.

**Fase 6: Testing, E2E e Rilascio MVP**
- Creare 3 app di esempio (es. monitoraggio CPU, wrapper per comando CLI, client
  WebSocket).
- Rilascio della versione `v0.1.0-alpha`.

---

## 7. Struttura del Repository (Consigliata)

```text
/simpleui-compiler/
├── cmd/
│   └── simpleui/
│       └── main.go          # Entry point della CLI del compilatore
├── pkg/
│   ├── lexer/               # Analizzatore lessicale
│   ├── parser/              # Costruzione AST
│   ├── typechecker/         # Inferenza e validazione tipi
│   ├── codegen/
│   │   ├── jsgen/           # Generatore codice JavaScript
│   │   ├── htmlgen/         # Generatore HTML/CSS
│   │   └── gogen/           # Generatore del main.go wrapper per desktop
│   └── stdlib/              # Mappe per le funzioni intrinseche (http, ws, ecc)
├── runtime/                 # File Go precompilati per l'embed (il bridge server)
│   └── desktop_bridge.go
└── examples/                # Applicazioni di esempio in .sui
```

> Adopted layout note: the Go 1.24 toolchain resolves packages by directory and
> only resolves std + required modules for import paths, so this layout was
> adapted to a flat single-package module (see `SPEC.md` §1). Revisit when the
> toolchain supports local subpackages.