# SimpleUI TODO List (Italiano)

> Questa è la copia italiana. La fonte canonica e verificata è
> [`TODO.md`](TODO.md); in caso di divergenza vale `TODO.md`.

**Stato:** 0.1.0-alpha  
**Ultima revisione:** 2026-10-25

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
5. WebView nativa.
6. `simpleui test` e CI cross-platform.
7. Moduli/import utente.

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
  helper, WebView.
- Tutorial, cookbook esteso, migration guide, changelog.
