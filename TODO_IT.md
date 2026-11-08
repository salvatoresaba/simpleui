# SimpleUI TODO List (Italiano)

> Questa è la copia italiana. La fonte canonica e verificata è
> [`TODO.md`](TODO.md); in caso di divergenza vale `TODO.md`.

**Stato:** 0.1.0-alpha  
**Ultima revisione:** 2026-11-07

Legenda: `[x]` implementato e verificato · `[~]` parziale · `[ ]` mancante.

## Snapshot

Verificato con:

```bash
go build -o simpleui.exe .
go vet ./...
go test -count=1 ./...
```

Tutti verdi su Windows con Go 1.24.2.

Completati di recente:

- [x] Binding reattivi `class:`/`style:` e ternary.
- [x] Tabelle con data binding, CSV/TSV/SSV/PSV/JSONL, editing, fixed e
      formula escape.
- [x] Conversione XLSX/ODS a monte.
- [x] Configurazione di sicurezza rigorosa.
- [x] Iframe: sandbox di default, `srcAllow`, validazione runtime e messaggio
      visibile in inglese.
- [x] `simpleui check --json`.
- [x] Documentazione embedded, manifest, guide per agenti, recipe, `llms.txt`,
      `simpleui docs` e MCP docs.
- [x] Export HTML singolo (`--single`) e `--minify` conservativo.
- [x] Dialog di build desktop con preflight, capability, log, output locale e
      download opzionale.
- [x] Target desktop WebView incorporata (`desktop-webview`) con WebView2 /
      WKWebView / WebKitGTK, stesso bridge e modello di capability.
- [x] Selettore target nel designer e avvio da file (`simpleui gui app.sui`) con
      associazione `.sui` su Windows.
- [x] Sistema temi condiviso con varianti Light/Dark, pagina ed elementi base,
      tooltip, media, hover, preset chart/graph, export/import temi e contrasto
      automatico.
- [x] Audit di contrasto computed-style per tutte le varianti e golden screenshot
      Standard Light/Dark.
- [x] Filtro del pannello Problems per classi runtime, keyframe, blocchi `:root`
      e override di stato/pseudo-elementi.
- [x] Tree control completo: selezione dell'intero albero, modifica strutturale
      da menu contestuale, icone leaf/parent/open, simboli di ramo aperto/chiuso,
      picker icone filtrato, `onTreeChange` e `tree.lastChange`.
- [x] Inspector chart semplificato: type, Mode e Chart settings in alto, dati in
      fondo, point width nelle impostazioni, pannello attributes rimosso e campi
      X/Y rimossi a favore della modifica diretta dei dati.
- [x] Controlli delle preferenze di progetto a tema e card per categoria in una
      griglia a due colonne.

## Correzioni di accuratezza

- HTTP custom headers/timeout **non** sono nel contratto stdlib.
- `http.upload`, `fs` avanzato, crypto, reconnect WebSocket e source maps **non**
  sono implementati; le voci pubbliche non supportate sono state rimosse.
- LSP supporta completion di base e diagnostica live, non hover/definition/
  references/symbols/rename/code actions.
- `--single --minify` è un minifier conservativo, non un minifier JS completo.

## Prossime priorità

1. Timeout/audit/process-kill del bridge desktop.
2. HTTP headers/multipart e reconnect WebSocket.
3. LSP avanzato (hover, definition, references, symbols, rename).
4. Source maps e decisione sul minifier completo.
5. Packaging/firma desktop: installer, icona/metadati, bootstrap WebView2 e build
   firmate/notarizzate.
6. `simpleui test` e CI cross-platform.
7. Moduli/import utente.
8. Completezza temi: collegamento automatico progetto → chart/graph e override
   scrollbar per singolo elemento.
9. Associazioni file su macOS/Linux.

## Aree principali ancora aperte

- Import/moduli utente, export di simboli, alias.
- Enum, union, type guards, generics, literal types, const, readonly.
- Rest/named parameters, destructuring, spread.
- Computed/effects/batching/fine-grained reactivity.
- Error boundaries, list keys, suspense, context.
- HTTP multipart, WebSocket reconnect, fs avanzato, crypto, path utilities,
  JSON Schema, i18n.
- Source maps, tree shaking, bundle splitting, PWA/SSR/WASM.
- Component library, tokens JSON, multi-device preview, asset manager.
- HMR, preserve state, browser error overlay.
- Test runner, fuzzing, snapshot, benchmark, CI.
- Bridge audit/timeout, `build --csp`, permission prompts, `postMessage`
  helper.
- Tutorial, cookbook esteso, migration guide, changelog.

## Gap noti

- Il tema di progetto non aggiorna automaticamente chart/graph: usare i preset
  tema o il pannello chart.
- Override scrollbar per singolo elemento non esposto.
- La WebView desktop richiede il motore web di sistema, non e una toolkit nativa.
- Mancano installer, firma e auto-update per gli eseguibili desktop.
- I golden screenshot coprono Standard Light/Dark; tutte le 24 varianti sono
  coperte dall'audit di contrasto.
- L'associazione `.sui` e solo Windows.
- Un cambio tema progetto non riscrive lo stile dei chart già esistenti.
