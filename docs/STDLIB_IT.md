# SimpleUI Stdlib Documentation

Lo standard library di SimpleUI è importabile da `"std"`. Le chiamate vengono tradotte in `__SUI_STD.<jsKey>(...)` e implementate nel runtime. Tutte le funzioni **non-fire-and-forget** restituiscono **due valori**: `value, err` (con `err = "null"` in caso di successo). Le funzioni marcate come *fire-and-forget* restituiscono `null` e non richiedono binding completo.

I moduli disponibili sono: `http`, `ws`, `fs`, `exec`, `json`, `table`, `tree`, `chart`, `dialog` (desktop).

## Import

```sui
import { http, json } from "std"
import { fs } from "std"
```

## HTTP Module (`http`)

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `http.get(url)` | string | (string, string) | true | GET HTTP request. Restituisce body come stringa |
| `http.getStatus(url)` | string | (int, string) | true | GET HTTP request. Restituisce status code come int |
| `http.post(url, body)` | string, string | (string, string) | true | POST HTTP request. Body come stringa (JSON/text) |
| `http.postFiles(url)` | string | (string, string) | true | POST dei file selezionati/drop come `multipart/form-data`, campo `files` |
| `http.postFilesSafe(url, allow)` | string, string | (string, string) | true | Upload file con allow-list SSRF |
| `http.getSafe(url, allow)` | string, string | (string, string) | true | GET con allow-list SSRF (host/schema esatti, no redirect) |
| `http.getStatusSafe(url, allow)` | string, string | (int, string) | true | Status code con allow-list SSRF |
| `http.postSafe(url, body, allow)` | string, string, string | (string, string) | true | POST con allow-list SSRF |

**SSRF:** `http.get/getStatus/post/postFiles` accettano solo URL **letterali**. Un URL
dinamico deve usare la variante `*Safe` con una allow-list **letterale** (il
compilatore rifiuta una variabile). Voci separate da virgole/spazi:

- `https://api.example.com` (tutti i path) o `https://api.example.com/v1`;
- `/api/` (path della stessa origine);
- per il WebSocket: `ws://host:porta` / `wss://host:porta`.

Il guard rifiuta credenziali, schemi non http(s)/ws(s), host/porta fuori lista
e redirect; timeout 30 s, risposta max 10 MB. Il requisito di usare `*Safe` per
gli URL dinamici è una policy di compilazione configurabile in
`simpleui.security.json` (default `safe`).

`http.postFiles` invia la **selezione corrente** di `pickFiles:`/`dropFiles:`
come multipart nel campo `files`. Invia anche un campo parallelo `paths` con i
percorsi relativi visibili al browser, cosi il backend puo mostrare il contesto
cartella/sottocartella. Carica i byte reali; il backend deve imporre limiti di
dimensione/numero e validare i nomi. Usare `postFilesSafe` con una allow-list
letterale per destinazioni dinamiche.

## WebSocket Module (`ws`)

WebSocket reali integrati con runtime. Richiede import `ws` da "std".

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `ws.connect(url)` | string | (any, string) | true | Apre connessione WebSocket. Restituisce istanza WS (gestita da runtime) |
| `ws.connectSafe(url, allow)` | string, string | (any, string) | true | WebSocket con allow-list `ws://`/`wss://` |
| `ws.send(ws, data)` | any, string | (null, string) | true | Invia messaggio. *Fire-and-forget*: richiede solo binding opzionale? Vedi note sotto |
| `ws.onMessage(ws, fn)` | any, string \| fn | (null, string) | true | Registra handler per messaggi. *Fire-and-forget* |
| `ws.close(ws)` | any | (null, string) | true | Chiude connessione. *Fire-and-forget* |

**Note:**
- Le chiamate `ws.send`, `ws.onMessage`, `ws.close` sono *fire-and-forget* secondo `StdFunc` (ret0 = "null"). 
- `ws.onMessage` può essere usato in modo state-driven (vedi esempi) per aggiornare `state` e triggerare re-render.
- `ws.connect` restituisce un'istanza utilizzabile con gli altri metodi ws.

## File System Module (`fs`)

FS accessibile via **Desktop Bridge** (target desktop) con capability-based security (`--allow-dir`). Su web target alcune operazioni potrebbero non essere disponibili a seconda del contesto. Tutte async.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `fs.read(path)` | string | (string, string) | true | Legge file come stringa |
| `fs.write(path, data)` | string, string | (null, string) | true | Scrive file. *Fire-and-forget* |
| `fs.exists(path)` | string | (bool, string) | true | Verifica esistenza (ritorna bool) |
| `fs.stat(path)` | string | (string, string) | true | Info file/dir (stringa descrittiva) |
| `fs.mkdir(path)` | string | (null, string) | true | Crea directory (mkdir -p style). *Fire-and-forget* |
| `fs.remove(path)` | string | (null, string) | true | Rimuove file/directory. *Fire-and-forget* |
| `fs.list(path)` | string | (array[string], string) | true | Elenca entry della directory |

## Exec Module (`exec`)

Esecuzione comandi via **Desktop Bridge** con allow-list (`--allow-exec`). 

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `exec.run(cmd, args)` | string, array[string] | (string, string) | true | Esegue comando. `cmd` risolto anche nella cartella dell'exe. Restituisce stdout/stderr combinati come stringa |

**Esempio (vedi `examples/userdata_table/`):**
```sui
import { exec, json } from "std"
async fn load() {
  out, err = exec.run("userdata.exe", [])
  if err != null { ... return }
  data, perr = json.parse(out)
  ...
}
```

## JSON Module (`json`)

Parsing/stringify JSON.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `json.parse(s)` | string | (any, string) | true | Parse JSON. Restituisce oggetto/array/valore |
| `json.stringify(v)` | any | (string, null) | true | Stringify JSON. *Fire-and-forget* in pratica? Secondo `StdFunc` ret1 = "null" quindi non richiede binding di errore separato nella logica usuale |

## Dialog Module (`dialog`, desktop)

Selezione nativa di file e cartelle. Importare con `import { dialog } from "std"`; sul target web le chiamate restituiscono errore. Il dialogo e guidato dall'utente; i percorsi possono essere passati a `exec.run`/`fs`, che applicano comunque le rispettive allow-list.

| Function | Params | Return | Async | Description |
|---|---|---|---|---|
| `dialog.pickFile()` | - | (string, string) | true | Dialogo file nativo; stringa vuota su annullamento |
| `dialog.pickFiles()` | - | (array[string], string) | true | Dialogo multi-file nativo |
| `dialog.pickDirectory()` | - | (string, string) | true | Dialogo cartella nativo; stringa vuota su annullamento |

Implementazione desktop:

- Windows: PowerShell + dialoghi `System.Windows.Forms`;
- macOS: `osascript` + `choose file` / `choose folder`;
- Linux: `zenity`, con fallback `kdialog`.
## Table Module (`table`, core)

`table` è un modulo **core**: è sempre disponibile e non richiede `import`.
Agisce sulla tabella che ha l'attributo `name:` corrispondente. Le funzioni
sono sincrone (solo DOM) e restituiscono `(valore, errore)`.

| Function | Params | Return | Description |
|---|---|---|---|
| `table.display(name, data)` | string, any | (int, string) | Popola da JSON, CSV/TSV/SSV/PSV, JSONL, array o oggetti |
| `table.data(name)` | string | (array[array[string]], string) | Righe visibili, incluse le modifiche |
| `table.header(name)` | string | (array[string], string) | Etichette dell'intestazione |
| `table.select(name, cols)` | string, array[int] | (array[array[string]], string) | Solo le colonne indicate |
| `table.slice(name, from, count)` | string, int, int | (array[array[string]], string) | Blocco di righe (`count <= 0` = fino alla fine) |
| `table.importText(name, text, spec)` | string, string, string | (int, string) | Parse esplicito: `csv`, `tsv`, `ssv`, `psv`, `json`, `jsonl`, `auto` |
| `table.exportText(name, spec)` | string, string | (string, string) | Serializza in CSV/TSV/SSV/PSV/JSON/JSONL |
| `table.json(name)` | string | (string, string) | JSON con oggetti se c'è un'intestazione |
| `table.clear(name)` | string | (int, string) | Svuota il corpo |

Spec di `importText`/`exportText`: `formato` + opzioni separate da `;`:

- `header=0|1`
- `sep=comma|semicolon|tab|pipe|<char>`
- `quote=<char>`
- `trim=1`
- `formula=escape` (solo export CSV: neutralizza `= + - @` con `'`)

```sui
state out = ""

fn load() -> null {
	n, e = table.display("utenti", "ID,Nome\n1,Alice\n2,Bob")
	if e != null { out = e } else { out = str(n) + " righe" }
}

fn saveCSV() -> null {
	csv, e = table.exportText("utenti", "csv;formula=escape")
	out = csv
}
```

Sicurezza:

- le celle sono scritte con `textContent`: HTML nei dati non viene interpretato;
- limiti: 100.000 righe, 10.000 colonne, 20 MB di testo per chiamata;
- `formula=escape` evita la formula injection quando il CSV viene aperto in
  Excel/LibreOffice (opt-in: `-5` diventa testo).

Dettagli completi in [`SECURITY.md`](../SECURITY_IT.md).

## Chart Module (`chart`, core)

`chart` è un modulo **core**: sempre disponibile, senza `import`. Agisce sul
grafico di stato con l'attributo `name:` corrispondente (o sul titolo). Le
funzioni sono sincrone e restituiscono `(valore, errore)`. Aggiornano testo e
colori dei tag di forme e link senza ricostruire tutto il grafico, così il
controllo può essere usato come monitor di processi o eventi. Le interazioni
static sono disponibili anche come lista eventi (`chart.pollEvents`,
`chart.nextEvent`, ...), così il chart non deve chiamare direttamente funzioni
SUI che modificano stato.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `chart.setTag(name, id, text, color, textColor)` | string, string, string, string, string | (bool, string) | Imposta testo, colore sfondo e colore testo del tag |
| `chart.update(name, id, patch)` | string, string, any | (bool, string) | Applica un oggetto patch o JSON a un elemento |
| `chart.updateMany(name, updates)` | string, any | (int, string) | Applica una mappa `{ id: patch }` o un array di patch |
| `chart.get(name, id)` | string, string | (string, string) | Attributi correnti di una forma/link in JSON |
| `chart.elements(name)` | string | (string, string) | Riepiloghi correnti di forme/link come array JSON |
| `chart.find(name, query)` | string, string | (string, string) | Cerca elementi per `campo:valore` o su tutti i campi |
| `chart.getAttr(name, id, field)` | string, string, string | (string, string) | Legge un attributo come stringa |
| `chart.setAttr(name, id, field, value)` | string, string, string, any | (bool, string) | Scrive un attributo validato |
| `chart.on(name, id, event, handler)` | string, string, string, string | (bool, string) | Associa una funzione SUI a un evento di un elemento |
| `chart.off(name, id, event)` | string, string, string | (bool, string) | Rimuove uno o tutti i binding dell’elemento |
| `chart.bindings(name)` | string | (string, string) | Binding `chart.on` attivi in JSON |
| `chart.focusRow(name, id)` | string, string | (bool, string) | Evidenzia, porta in vista e focalizza una riga tabella (anche in static) |
| `chart.pollEvents(name)` | string | (string, string) | Tutti gli eventi pendenti in JSON e svuota la lista |
| `chart.peekEvents(name)` | string | (string, string) | Eventi pendenti in JSON senza svuotare |
| `chart.nextEvent(name)` | string | (string, string) | Evento più vecchio in JSON e lo rimuove |
| `chart.clearEvents(name)` | string | (bool, string) | Svuota la lista eventi |
| `chart.eventCount(name)` | string | (int, string) | Numero di eventi pendenti |

Id elemento: `id` forma (`"a"`), obbligatorio e univoco; `id` link
(`"edge_1"`). I link possono omettere `id` e ricevono `edge_1`, `edge_2`, ... con
verifica di unicità.

Tutte le funzioni runtime accettano solo l’`id` esplicito. Prefissi (`shape:`,
`state:`, `link:`, `edge:`), label e `from->to` sono rifiutati.

Gli eventi riportano l’`id` esplicito. Per modificare un singolo attributo usare
gli helper tipizzati (`chart.setText`, `chart.setTagColor`, `chart.setHint`,
`chart.setColor`, ...); `chart.getAttr`/`chart.setAttr` coprono gli altri campi e
`chart.update` gestisce patch multi-attributo. Accessor aggiuntivi:
`getWidth`/`setWidth` (shape `w` o linea link), `getHeight`/`setHeight` (shape),
`getX`/`setX`, `getY`/`setY`, `getRotation`/`setRotation`,
`getDashStyle`/`setDashStyle`, `getStartMarker`/`setStartMarker`,
`getEndMarker`/`setEndMarker`, `getFrom`/`setFrom`, `getTo`/`setTo`,
`getWaypoints`/`setWaypoints`, `getTagKind`/`setTagKind`,
`getTagBox`/`setTagBox`, `getTagAlign`/`setTagAlign`,
`getTagLeader`/`setTagLeader`, `getTextPlacement`/`setTextPlacement`,
`getTagSide`/`setTagSide`, `getBorderDashLength`/`setBorderDashLength`,
`getBorderDashGap`/`setBorderDashGap`, `getSnap`/`setSnap`.

Per leggere i valori correnti (dopo modifiche/import):

```sui
shapeJson, err = chart.get("monitor", "process")
allJson, err2 = chart.elements("monitor")
```

`chart.get` restituisce `id`, `label`, `tag`, `hint`, geometria, colori e campi
dei gestori. `chart.elements` restituisce un array JSON con `kind`, `id`,
`label`, `tag`, `hint`, geometria/routing, colori e nomi
`onClick`/`onTagClick`.

Il chart è una sorgente di eventi. Ogni interazione static viene accodata; SUI
svuota la lista e chiama le proprie funzioni esterne. È il modello consigliato:

```sui
events, err = chart.pollEvents("monitor")
```

Il payload contiene `event`, `kind`, `part`, `id`, `label`, `tag`, `from`,
`to`, coordinate e modificatori. Usare `chart.pollEvents` (svuota),
`chart.peekEvents` (legge senza svuotare), `chart.nextEvent` (un evento),
`chart.clearEvents` e `chart.eventCount`.

I callback diretti legacy sono ancora disponibili con `chart.on` (o campi dati /
attributi chart-level), ma chiamano una funzione SUI dall’interno del runtime del
chart. Per il nuovo codice preferire la lista eventi:

```sui
ok, err = chart.on("monitor", "process", "click", "processClicked")
```

Eventi: `click`, `dblclick`, `mouseover`, `mouseout`, `contextmenu` e
`tagClick`, `tagDoubleClick`, `tagMouseOver`, `tagMouseOut`, `tagContextMenu`.
Usare `chart.off` per rimuovere i binding e `chart.bindings` per ispezionarli.

I dati dell’elemento possono anche portare gestori dichiarativi come fallback:

```json
{ "id": "process", "label": "Process", "onClick": "processClicked" }
```

Campi supportati: `onClick`, `onDoubleClick`, `onMouseOver`, `onMouseOut`,
`onContextMenu`, gli alias `onShape*`/`onLink*` e `onTagClick`,
`onTagDoubleClick`, `onTagMouseOver`, `onTagMouseOut`, `onTagContextMenu`.
Ordine di risoluzione: binding `chart.on`, campo dati, attributo chart-level. I
gestori devono essere funzioni SUI.

Campi patch per i tag: `tag`/`text`, `tagTextColor`/`textColor`,
`tagBoxColor`/`tagColor`/`backgroundColor`, `tagBorderColor`,
`tagLeaderColor`, `tagKind`, `tagAlign`, `tagBox`, `tagLeader`, `tagArrow`,
`tagArrowSize`, `tagArrowShape`, `tagFont`, `tagFontSize`, `tagFontWeight`,
`tagFontStyle`, `tagTextDecoration`, `tagTextAlign`, `tagLetterSpacing`,
`tagLineSpacing`, `tagTextTransform`, `tagTextShadow`,
`tagTextShadowColor`, `tagPadding`, `tagRotation`, `tagOffset`.
`chart.update` accetta anche `label`, `description`, `hint`, `color`,
`borderColor`, `highlightColor`, `borderStyle`, `borderWidth`, `cornerRadius`,
`shape`, `textPlacement`, `tagSide`, `snap`, `borderDashLength`,
`borderDashGap`, geometria forma `x`/`y`/`w`/`h`/`rotation` e, per i link,
`routing`, `width`, `dashStyle`, `dashLength`, `dashGap`, `arrowSize`,
`startMarker`, `endMarker`, `from`, `to`, `fromPort`, `toPort`, `fromAnchor`,
`toAnchor`, `waypoints`, `arcRadius`, `arcSweep`, `arcLarge` e controlli
`bend*`. In `setTag`/`tagColor`, `color` è lo sfondo del tag; in `update`,
`color` è il riempimento della forma o il colore della linea.

I colori accettano `#rgb`, `#rrggbb`, `rgb()/rgba()` e `transparent`/`none`.

```sui
fn updateMonitor() -> null {
	ok, e = chart.setTag("monitor", "a", "Running", "#22c55e", "#ffffff")
	ok2, e2 = chart.update("monitor", "edge_1", "\{\"text\":\"go\",\"tagColor\":\"#0ea5e9\",\"textColor\":\"#ffffff\"\}")
	n, e3 = chart.updateMany("monitor", "\{\"b\":\{\"tag\":\"Waiting\"\},\"c\":\{\"tag\":\"Done\"\}\}")
}

interval(1000, updateMonitor)
```

## Tree Module (`tree`, core)

`tree` è un modulo **core**: sempre disponibile, senza `import`. Agisce
sull'albero con l'attributo `name:` corrispondente (`name: "files"`). Le funzioni
restituiscono `(valore, errore)`.

| Funzione | Parametri | Ritorno | Async | Descrizione |
|---|---|---|---|---|
| `tree.importJson(name, text)` | string, string | (int, string) | no | Sostituisce i nodi da JSON; restituisce il conteggio |
| `tree.importXml(name, text)` | string, string | (int, string) | no | Sostituisce i nodi da XML; restituisce il conteggio |
| `tree.importJsonFile(name, path)` | string, string | (int, string) | sì | Legge un file selezionato e importa JSON |
| `tree.importXmlFile(name, path)` | string, string | (int, string) | sì | Legge un file selezionato e importa XML |
| `tree.exportJson(name)` | string | (string, string) | no | Esporta i nodi correnti come JSON |
| `tree.exportXml(name)` | string | (string, string) | no | Esporta i nodi correnti come XML |
| `tree.downloadJson(name, filename)` | string, string | (int, string) | no | Scarica il JSON (`1` = successo) |
| `tree.downloadXml(name, filename)` | string, string | (int, string) | no | Scarica l'XML (`1` = successo) |
| `tree.clear(name)` | string | (int, string) | no | Rimuove tutti i nodi |
| `tree.expandAll(name)` | string | (int, string) | no | Espande tutti i genitori; restituisce il numero |
| `tree.collapseAll(name)` | string | (int, string) | no | Comprime tutti i genitori |
| `tree.expandTo(name, level)` | string, int | (int, string) | no | Espande sopra `level` e comprime i livelli più profondi |
| `tree.setChildren(name, key, json)` | string, string, string | (int, string) | no | Aggiunge/sostituisce i figli di un elemento da JSON |
| `tree.setChildrenError(name, key, message)` | string, string, string | (int, string) | no | Mostra un errore di caricamento lazy |
| `tree.retryChildren(name, key)` | string, string | (int, string) | no | Riprova un caricamento lazy tramite `onLoad` |
| `tree.lastMove(name)` | string | (string, string) | no | JSON `{key, parent, index}` dell'ultimo riordino |
| `tree.lastChange(name)` | string | (string, string) | no | JSON `{op, key, parent, index, ...}` dell'ultima modifica dati |
| `tree.undoMove(name)` | string | (int, string) | no | Ripristina l'albero precedente dallo stack undo |
| `tree.selectedKeys(name)` | string | (string, string) | no | Array JSON delle key selezionate |
| `tree.checkedKeys(name)` | string | (string, string) | no | Array JSON delle key checked |
| `tree.search(name, term)` | string, string | (int, string) | no | Applica o pulisce il filtro inline |
| `tree.setColumns(name, json)` | string, string | (int, string) | no | Crea/sostituisce le colonne da un array JSON |
| `tree.getColumns(name)` | string | (string, string) | no | Colonne correnti come array JSON |
| `tree.moveColumn(name, from, to)` | string, int, int | (int, string) | no | Riordina una colonna |
| `tree.setCell(name, key, column, value)` | string, string, string, string | (int, string) | no | Imposta il valore di una cella |

### Formato JSON

La radice può essere un array oppure un oggetto con `nodes`, `items`, `tree` o
`children`. Ogni nodo supporta:

```json
[
  {
    "label": "src",
    "key": "src",
    "open": true,
    "editable": true,
    "renameIcon": "A",
    "showRenameIcon": false,
    "icon": "📁",
    "iconOpen": "📂",
    "badge": "12",
    "badgeType": "info",
    "meta": "2 KB",
    "checked": true,
    "tooltip": "Cartella sorgente",
    "lazy": false,
    "size": "2 KB",
    "date": "oggi",
    "status": "ok",
    "cells": { "owner": "Anna" },
    "children": [
      { "label": "main.go", "key": "main" }
    ]
  }
]
```

- `label` è preferito; sono accettati anche `text`, `name` e `title`.
- `key` è opzionale; se manca viene usato un path generato. Le key devono essere
  uniche.
- `open`, `editable`, `showRenameIcon` e `checked` accettano booleani o stringhe
  `true`/`false`/`1`/`0`.
- `icon`, `iconOpen`, `badge`, `badgeType`, `meta`, `tooltip`, `size`, `date` e
  `status` sono testo inerte. `badgeType` è `neutral`, `info`, `success`,
  `warning` o `danger`.
- `lazy: true` indica un elemento i cui figli vengono caricati dopo con
  `tree.setChildren`.
- `cells` è un oggetto di valori per colonne personalizzate
  (`{ "owner": "Anna" }`); l'attributo SUI accetta anche testo `key=value`.
- Stringhe e numeri sono accettati come etichette foglia.

### Formato XML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<tree name="files">
  <item label="src" key="src" open="true" icon="📁" iconOpen="📂" badge="12" badgeType="info" meta="2 KB" checked="true">
    <item label="main.go" key="main"/>
    <item label="README.md" key="readme" renameIcon="📝"/>
  </item>
  <item label="generated" key="generated" editable="false"/>
</tree>
```

`<tree>` è il contenitore; qualsiasi elemento è accettato come nodo, quindi
vanno bene `<item>`, `<node>` o tag applicativi. Se non ci sono attributi
`label`/`text`/`name`/`title`, viene usato il testo diretto dell'elemento. Gli
attributi XML `icon`, `iconOpen`, `badge`, `badgeType`, `meta`, `tooltip`,
`lazy`, `size`, `date`, `status` e `checked` vengono preservati.

### Lazy loading

Un elemento con `lazy: true` non ha figli e mostra la freccia. L'espansione
scrive la key nello stato collegato a `loadKey` e chiama `onLoad`. Il gestore può
aggiungere i figli in seguito:

```sui
state loadKey = ""

fn loaded() -> null {
	// avvia una richiesta asincrona usando loadKey
}

async fn finishLoad() -> null {
	n, e = tree.setChildren("files", loadKey, "[{\"label\":\"figlio\"}]")
	if e != null {
		tree.setChildrenError("files", loadKey, e)
	}
}
```

Usare `tree.selectedKeys`/`tree.checkedKeys` per leggere selezione e checkbox:
restituiscono array JSON.

### File

`importJsonFile`/`importXmlFile` leggono il file selezionato con `pickFiles:` o
trascinato nell'app (`drag.files()`), confrontando `path` o `name`; passare `""`
usa il primo file selezionato. Su desktop `dialog.pickFile()` restituisce solo un
path: usare `fs.read(path)` e passare il testo a `tree.importJson`/`importXml`.

```sui
state importPath = ""
state note = ""

async fn importJSON() -> null {
	n, e = tree.importJsonFile("files", importPath)
	if e != null { note = e } else { note = "importati " + str(n) + " nodi" }
}
```

`downloadJson`/`downloadXml` avviano il download dal browser; il nome viene
sanitizzato e l'estensione aggiunta se manca.

### Limiti e sicurezza

- limiti: 10.000 nodi, profondità 64, 2.000 caratteri per etichetta e tooltip,
  512 per key, 64 per icona e 128/256 per badge/metadata; il testo usa il limite
  globale `maxText` (default 20 MB);
- etichette, key, icone, badge e metadata sono scritti con
  `textContent`/attributi, mai HTML;
- dichiarazioni DTD ed entità XML vengono rifiutate prima del parsing;
- l'import sostituisce solo la struttura dei nodi: titolo, stile, `name` e
  binding del tree restano invariati;
- gli override locali di apertura/rinomina del tree vengono azzerati;
- i modelli importati sopravvivono ai re-render reattivi;
- l'export legge etichette e stato aperto correnti, incluse le rinomine inline.

## Error Handling Pattern

Tutte le funzioni con due valori seguono il pattern:
```sui
async fn fetchData() {
  body, err = http.get("https://api.example.com/data")
  if err != null {
    // handle error
    return
  }
  data, perr = json.parse(body)
  if perr != null {
    // handle parse error
    return
  }
  // use data
}
```

## Note Sicurezza (Desktop)

- **Capability-based**: permessi espliciti via CLI (`--allow-exec`, `--allow-dir`)
- **Path traversal protection**: normalizzazione, allowlist e risoluzione dei
  symlink (un link dentro la directory consentita non può uscire)
- **Per-run secret**: tutte le chiamate bridge autenticate con token casuale,
  confrontato in tempo costante
- **Localhost only**: bind su 127.0.0.1, Host header loopback obbligatorio
  (anti DNS-rebinding), richieste cross-site rifiutate
- **Command resolution**: comandi senza path separator cercati accanto all'exe;
  nessuna shell, quindi gli argomenti non vengono reinterpretati da `cmd`/`sh`
- **app.js**: `Cache-Control: no-store`, `X-Frame-Options: DENY`

