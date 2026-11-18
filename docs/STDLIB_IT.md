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
| `http.getWithHeaders(url, headers)` | string, any | (string, string) | true | GET HTTP request con mappa header personalizzati |
| `http.getWithHeadersSafe(url, headers, allow)` | string, any, string | (string, string) | true | GET con header e allow-list SSRF |
| `http.postWithHeaders(url, body, headers)` | string, string, any | (string, string) | true | POST HTTP request con mappa header personalizzati |
| `http.postWithHeadersSafe(url, body, headers, allow)` | string, string, any, string | (string, string) | true | POST con header e allow-list SSRF |
| `http.request(method, url, headers, body, timeoutMs)` | string, string, any, string, int | (string, string) | true | Richiesta HTTP generica con metodo, header, body e timeout in ms |
| `http.requestSafe(method, url, headers, body, timeoutMs, allow)` | string, string, any, string, int, string | (string, string) | true | Richiesta HTTP generica con allow-list SSRF |
| `http.upload(url, fieldName, filename, content)` | string, string, string, string | (string, string) | true | Caricamento multipart/form-data con nome campo e file |
| `http.uploadSafe(url, fieldName, filename, content, allow)` | string, string, string, string, string | (string, string) | true | Caricamento multipart con allow-list SSRF |
| `http.uploadWithHeaders(url, fieldName, filename, content, headers)` | string, string, string, string, any | (string, string) | true | Caricamento multipart con header personalizzati |
| `http.uploadWithHeadersSafe(url, fieldName, filename, content, headers, allow)` | string, string, string, string, any, string | (string, string) | true | Caricamento multipart con header e allow-list SSRF |

**SSRF:** `http.get/getStatus/post/postFiles/getWithHeaders/postWithHeaders/request/upload/uploadWithHeaders` accettano solo URL **letterali**. Un URL
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
| `ws.connectWithOptions(url, options)` | string, any | (any, string) | true | Connessione WebSocket con opzioni (`reconnect`, `delay`, `maxDelay`, `heartbeat`, `ping`) |
| `ws.connectWithOptionsSafe(url, options, allow)` | string, any, string | (any, string) | true | Connessione WebSocket con opzioni e allow-list SSRF |
| `ws.send(ws, data)` | any, string | (null, string) | true | Invia messaggio (fire-and-forget) |
| `ws.onMessage(ws, fn)` | any, string \| fn | (null, string) | true | Registra handler per messaggi (fire-and-forget) |
| `ws.onOpen(ws, fn)` | any, string \| fn | (null, string) | true | Registra handler per evento di apertura/connessione (fire-and-forget) |
| `ws.onClose(ws, fn)` | any, string \| fn | (null, string) | true | Registra handler per chiusura/disconnessione (fire-and-forget) |
| `ws.onError(ws, fn)` | any, string \| fn | (null, string) | true | Registra handler per errori (fire-and-forget) |
| `ws.status(ws)` | any | (string, string) | false | Restituisce stato: `"open"`, `"connecting"`, `"closing"`, `"closed"` |
| `ws.reconnect(ws)` | any | (null, string) | true | Forza riconnessione immediata (fire-and-forget) |
| `ws.close(ws)` | any | (null, string) | true | Chiude connessione e cancella timer (fire-and-forget) |

**Note:**
- Le chiamate `ws.send`, `ws.onMessage`, `ws.onOpen`, `ws.onClose`, `ws.onError`, `ws.reconnect`, `ws.close` sono *fire-and-forget* secondo `StdFunc` (ret0 = "null"). 
- `ws.onMessage` può essere usato in modo state-driven (vedi esempi) per aggiornare `state` e triggerare re-render.
- `ws.connectWithOptions` supporta riconnessione automatica con backoff esponenziale ed heartbeat periodico.

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
| `fs.copy(src, dst)` | string, string | (null, string) | true | Copia un file; entrambi i percorsi devono essere consentiti |
| `fs.move(src, dst)` | string, string | (null, string) | true | Sposta/rinomina file o directory; entrambi i percorsi devono essere consentiti |
| `fs.rename(src, dst)` | string, string | (null, string) | true | Alias di `fs.move` |
| `fs.isDir(path)` | string | (bool, string) | true | Indica se il percorso è una directory esistente |
| `fs.size(path)` | string | (int, string) | true | Dimensione in byte di un file (le directory sono rifiutate) |
| `fs.readBytes(path)` | string | (array[int], string) | true | Legge un file come byte grezzi |
| `fs.writeBytes(path, bytes)` | string, array[int] | (null, string) | true | Scrive byte grezzi (base64 sul bridge) |

Sorgente e destinazione sono controllati contro la capability `--allow-dir`.
`fs.copy` copia solo file regolari; crea prima la directory di destinazione e
usa `fs.move` per le directory. Sul target web queste chiamate restituiscono un
errore.

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

## Modulo JSON (`json`)

```sui
import { json } from "std"
```

| Funzione | Parametri | Ritorno | Async | Descrizione |
|---|---|---|---|---|
| `json.parse(s)` | string | (any, string) | true | Analizza testo JSON; rifiuta BOM iniziale e input troppo grande |
| `json.parseBytes(bytes, [encoding])` | array[int], string? | (any, string) | false | Analizza byte UTF-8/UTF-16LE/UTF-16BE; BOM rilevato automaticamente |
| `json.stringify(value, [options])` | any, map? | (string, string) | true | Serializza JSON. Opzioni: `indent` (bool/int/string), `sortKeys`, `newline` |
| `json.toBytes(value, encoding, [options])` | any, string, map? | (array[int], string) | false | Serializza in byte UTF-8/UTF-16LE/UTF-16BE |
| `json.validate(value, schema)` | any, any | (bool, string) | false | Valida con un sottoinsieme JSON Schema; restituisce il primo errore |
| `json.validateAll(value, schema)` | any, any | (array[string], string) | false | Valida e raccoglie fino a 100 errori |
| `json.pointer(value, pointer)` | any, string | (any, string) | false | Legge un JSON Pointer (RFC 6901) |
| `json.setPointer(value, pointer, next)` | any, string, any | (any, string) | false | Restituisce una copia con il pointer impostato o aggiunto |
| `json.removePointer(value, pointer)` | any, string | (any, string) | false | Restituisce una copia con il pointer rimosso |
| `json.clone(value)` | any | (any, string) | false | Copia profonda; rifiuta cicli e annidamento eccessivo |
| `json.equal(a, b)` | any, any | (bool, string) | false | Uguaglianza JSON profonda (l'ordine degli array conta) |
| `json.diff(a, b, [options])` | any, any, map? | (array[any], string) | false | Produce operazioni RFC 6902. Opzioni: `key`, `detectMoves` |
| `json.patch(value, operations)` | any, any | (any, string) | false | Applica operazioni RFC 6902 a una copia; si ferma al primo errore |
| `json.merge(target, patch)` | any, any | (any, string) | false | Applica un JSON Merge Patch RFC 7386 |
| `json.merge3(base, ours, theirs, [options])` | any, any, any, map? | (array[string], string) | false | Merge a tre vie dentro `ours`; restituisce i conflitti |
| `json.loadFile(path, [encoding])` | string, string? | (any, string) | true | Caricamento desktop sotto `--allow-dir` |
| `json.loadBytes(path, [encoding])` | string, string? | (any, string) | true | Caricamento binario desktop con BOM/encoding |
| `json.saveFile(path, value, [options])` | string, any, map? | (null, string) | true | Salvataggio desktop; testo UTF-8 |
| `json.saveBytes(path, value, encoding, [options])` | string, any, string, map? | (null, string) | true | Salvataggio binario desktop; UTF-8/UTF-16LE/UTF-16BE |

### JSON Pointer (RFC 6901)

I pointer indirizzano un solo valore. `""` è l'intero documento, `/store/books/1`
seleziona un elemento di array e `/a~1b` indirizza la chiave `a/b` (`~1` è `/`,
`~0` è `~`). Un percorso mancante restituisce `(null, errore)`; un valore `null`
memorizzato restituisce `(null, null)` ed è quindi distinguibile. `setPointer`
sostituisce un valore esistente, crea una chiave mancante o aggiunge un elemento
di array; non modifica mai l'input. `removePointer` rifiuta la radice.

### JSON Patch (RFC 6902)

`json.patch` supporta `add`, `remove`, `replace`, `move`, `copy` e `test`. Le
operazioni vengono applicate a una copia profonda, in ordine, e l'input non
viene mai modificato parzialmente: qualsiasi errore annulla e restituisce
`(null, messaggio)`. `test` consente controlli di concorrenza ottimistica.

```sui
patched, err = json.patch(doc, [
  {"op": "replace", "path": "/store/open", "value": false},
  {"op": "add", "path": "/store/books/-", "value": {"id": 3, "title": "C"}},
  {"op": "test", "path": "/store/open", "value": false}
])
```

### Diff e merge a tre vie

`json.diff` emette operazioni RFC 6902 applicabili direttamente. Senza `key`
confronta gli array per posizione. Con `{"key": "id"}` allinea gli elementi su
quel campo, emette `move` per il riordino e `add`/`remove` per aggiunte e
rimozioni. Ogni operazione porta anche un campo opzionale `key` per l'ispezione;
i client standard lo ignorano. Le operazioni strutturali precedono le modifiche
annidate, così i percorsi restano validi. `detectMoves: false` sostituisce
l'intero array riordinato invece di emettere `move`.

`json.merge` implementa RFC 7386: gli oggetti si fondono ricorsivamente, `null`
cancella una chiave e array/scalari sostituiscono il valore. `json.merge3`
calcola `base -> theirs`, applica le modifiche non in conflitto a `ours` sul
posto e restituisce descrizioni come `replace /items/0/value`. I nodi in
conflitto mantengono il valore di `ours`. La radice di `ours` deve essere un
oggetto o un array.

### Validazione

`json.validate` restituisce la prima violazione; `json.validateAll` fino a 100
messaggi. Lo schema può essere una mappa o una stringa JSON. Parole chiave
supportate: `type`, `enum`, `const`, `properties`, `required`,
`additionalProperties`, `items` (schema o tupla), `additionalItems`,
`minItems`/`maxItems`, `uniqueItems`, `minLength`/`maxLength`, `pattern`,
`format` (`email`, `uri`, `url`, `uuid`, `date`, `date-time`), limiti numerici,
`multipleOf`, `minProperties`/`maxProperties`, `allOf`/`anyOf`/`oneOf`/`not` e
`if`/`then`/`else`. `$ref` è rifiutato. Lo schema è codice fidato dello
sviluppatore; i dati non vengono mai eseguiti.

### Byte, encoding e file

`json.parseBytes` e `json.toBytes` supportano UTF-8, UTF-16LE e UTF-16BE,
rilevano o scrivono il BOM e rifiutano UTF-32 con un errore esplicito. Il testo
JSON è limitato a 20 milioni di caratteri/byte, l'annidamento a 128 livelli, le
patch a 10.000 operazioni e i diff a 10.000 modifiche. Gli helper file desktop
usano la stessa allow-list di `fs` e `xml`; sul web restituiscono un errore
esplicito. La chiave `__proto__` è rifiutata da patch, diff e clone per evitare
prototype pollution.

## Modulo XML (`xml`)

`import { xml } from "std"`. Il modulo usa il DOM XML del browser
(`DOMParser`, `XMLSerializer`, XPath e `XSLTProcessor`) su web e desktop.
Tutte le funzioni seguono la convenzione std: lega entrambi i valori con
`a, b = ...`. Documenti e nodi sono tipizzati `xmlDocument` / `xmlNode`; un
documento è assegnabile a un nodo.

```sui
import { xml } from "std"

doc, err = xml.parse("<people><person id=\"1\"><name>Ada</name></person></people>")
root, rerr = xml.root(doc)
nameEl, nerr = xml.find(root, "name")
name, textErr = xml.text(nameEl)
out = name
```

### A. Caricamento e sintassi

| Funzione | Parametri | Ritorno | Async | Descrizione |
|---|---|---|---|---|
| `xml.parse(text)` | string | (any, string) | false | Analizza una stringa; rifiuta DOCTYPE/ENTITY |
| `xml.parseBytes(bytes, [encoding])` | array[int], string | (any, string) | false | Decodifica i byte con BOM, encoding della dichiarazione XML o encoding esplicito |
| `xml.loadFile(path)` | string | (any, string) | true | Carica un file desktop sotto `--allow-dir` |
| `xml.loadBytes(path, [encoding])` | string, string? | (any, string) | true | Carica un file binario desktop; rispetta BOM/dichiarazione/encoding |
| `xml.loadUrlSafe(url, allow)` | string, string | (any, string) | true | Scarica XML con il guard SSRF; `allow` deve essere una allow-list letterale |
| `xml.validate(text)` | string | (bool, string) | false | Controllo sintattico; l'errore include riga/colonna quando disponibili |

Limiti: 5.000.000 caratteri e 200.000 elementi. Le dichiarazioni `DOCTYPE` e
`ENTITY` vengono rifiutate prima del parsing. Streaming reale e caricamento
incrementale non sono disponibili: il documento viene materializzato intero.

### B. Navigazione

| Funzione | Descrizione |
|---|---|
| `xml.root(doc)` | Elemento radice |
| `xml.parent(node)` | Nodo genitore (il documento per la radice) |
| `xml.children(node)` | Figli elemento |
| `xml.childNodes(node)` | Tutti i figli (testo, commenti, CDATA, PI, elementi) |
| `xml.childCount(node)` | Numero di figli elemento |
| `xml.next(node)` / `xml.previous(node)` | Fratello elemento successivo/precedente |
| `xml.find(node, name)` | Primo discendente con quel tag |
| `xml.findNs(node, localName, namespaceURI)` | Ricerca namespace-aware |
| `xml.findAll(node, name)` | Tutti i discendenti con quel tag |
| `xml.descendants(node)` | Tutti gli elementi discendenti |
| `xml.has(node, name)` | Se esiste un discendente |

### C. Lettura dati

| Funzione | Descrizione |
|---|---|
| `xml.type(node)` | `document`, `element`, `text`, `comment`, `cdata`, `pi`, ... |
| `xml.name(node)` / `xml.localName(node)` | Nome qualificato/locale |
| `xml.namespace(node)` / `xml.prefix(node)` | URI namespace e prefisso |
| `xml.text(node)` | Solo figli **diretti** Text/CDATA |
| `xml.content(node)` | `textContent` aggregato del sottoalbero |
| `xml.attr(node, name)` / `xml.attrNs(node, local, ns)` | Valore attributo o `null` |
| `xml.attrs(node)` | Mappa di tutti gli attributi |
| `xml.path(node)` | Percorso assoluto come `/people/person[1]/name` |

`text` e `content` sono intenzionalmente diversi: usa `text` per il valore
diretto di un elemento e `content` quando servono anche i discendenti.

### D. Modifica

Operazioni su elementi e albero:

| Funzione | Descrizione |
|---|---|
| `xml.createDocument([rootName])` | Nuovo documento XML vuoto |
| `xml.createElement(doc, name)` / `xml.createElementNs(doc, ns, qname)` | Nuovo elemento |
| `xml.createText(doc, text)` / `xml.createComment` / `xml.createCdata` / `xml.createPi` | Altri tipi di nodo |
| `xml.parseFragment(doc, text)` | Analizza nodi senza una radice unica (restituisce un fragment) |
| `xml.setInnerXml(node, text)` | Sostituisce i figli con il contenuto XML analizzato |
| `xml.replaceChildren(node, nodes)` | Sostituisce i figli con un array di nodi |
| `xml.append(parent, node)` / `xml.prepend` | Aggiunge un figlio alla fine/inizio |
| `xml.insertBefore(parent, node, ref)` / `xml.insertAfter` | Inserisce rispetto a un figlio |
| `xml.remove(node)` / `xml.replace(old, new)` / `xml.clear(node)` | Rimuove, sostituisce o svuota |
| `xml.clone(node, [deep])` | Copia un nodo |
| `xml.rename(node, name)` | Rinomina un elemento (restituisce il nuovo nodo) |
| `xml.setText(node, text)` | Sostituisce i figli con un nodo testo |
| `xml.move(node, parent)` | Sposta un nodo sotto un altro genitore |
| `xml.reorder(parent, indexes)` | Riordina i figli elemento con una permutazione di indici |

Operazioni attributi: `xml.setAttr`, `xml.setAttrNs`, `xml.removeAttr`,
`xml.renameAttr`, `xml.hasAttr`. I nodi di un altro documento vengono importati
automaticamente all'inserimento.

### E. Ricerca e XPath

- `xml.find`, `xml.findNs`, `xml.findAll`, `xml.descendants`, `xml.findAttr(node, name, value)` (null per qualsiasi valore)
- `xml.findText(node, term, [regex])` — ricerca nel testo diretto, con RegExp opzionale
- `xml.xpath(node, expression, [namespaces])` — XPath 1.0; `namespaces` è una mappa
  prefisso → URI. Restituisce un array per i node set, oppure string/number/bool.
- `xml.xpathOne(node, expression, [namespaces])` — primo nodo o valore scalare.

```sui
names, xerr = xml.xpath(doc, "//person[@id='2']/name", {})
```

### F. Validazione

| Funzione | Descrizione |
|---|---|
| `xml.validate(text)` | Sintassi XML |
| `xml.validateXsd(doc, xsdText)` | Sottoinsieme XSD; restituisce `(false, primoErrore)` alla prima violazione |
| `xml.validateXsdAll(doc, xsdText)` | Sottoinsieme XSD; restituisce `(errori, null)` con fino a 100 messaggi |
| `xml.validateRules(doc, rules)` | Regole personalizzate per nome elemento: figli/attributi obbligatori, min/max, tipi di campo |

La validazione DTD non è supportata: DOCTYPE/ENTITY sono rifiutati per
sicurezza. XSD è un sottoinsieme pratico, non l'intera specifica XSD 1.0/1.1.
Copre elementi globali/locali con `ref`, `sequence`/`all`/`choice`,
`minOccurs`/`maxOccurs`, attributi, attribute group e attributi globali,
restrizioni `simpleType` con enumerazione, facet length/pattern e numerici,
`complexType` con estensione/restrizione `simpleContent`/`complexContent`,
namespace tramite `targetNamespace` e modelli di contenuto ereditati da
`complexContent`. Non implementati: `$ref` a documenti o elementi esterni,
substitution group, `key`/`keyref`/`unique`, wildcard oltre `anyAttribute`,
derivazione da `xs:anyType`, import/include di schema. I controlli namespace
sono disponibili con `xml.namespace`, `xml.findNs` e `targetNamespace` XSD.

### G. Esportazione e salvataggio

| Funzione | Descrizione |
|---|---|
| `xml.stringify(node, [options])` | Serializza in stringa. Opzioni: `indent` (bool/int), `declaration`, `encoding`, `sortAttributes` |
| `xml.toBytes(node, encoding, [options])` | Serializza in byte UTF-8/UTF-16LE/UTF-16BE |
| `xml.saveFile(path, node, [options])` | Salvataggio desktop sotto `--allow-dir`; solo UTF-8 |
| `xml.saveBytes(path, node, encoding, [options])` | Salvataggio binario desktop; UTF-8/UTF-16LE/UTF-16BE |

Commenti, CDATA, processing instruction e namespace vengono preservati dal
serializer DOM. `indent: true` produce output leggibile; omettilo per XML
compatto. Il salvataggio incrementale non è implementato: il documento viene
serializzato di nuovo.

### H. Trasformazione e confronto

| Funzione | Descrizione |
|---|---|
| `xml.toJson(node, [options])` | JSON convenzionale di default; `{"lossless": true}` restituisce un albero tipizzato con nodi document/element/text/cdata/comment/pi |
| `xml.fromJson(value, [rootName])` | Costruisce un documento XML dalla stessa convenzione |
| `xml.toCsv(node, rowName, [columns])` | CSV da elementi ripetuti; le colonne possono essere `@attributo` o nomi di figli |
| `xml.xslt(node, stylesheet)` | Applica un foglio XSLT; richiede `XSLTProcessor` del browser |
| `xml.merge(target, source, [options])` | Di default aggiunge i figli clonati di source. `{"mode": "update", "key": "id"}` aggiorna i figli corrispondenti e aggiunge i nuovi; restituisce il numero di nodi modificati |
| `xml.merge3(base, ours, theirs, [options])` | Merge a tre vie dentro `ours`; applica le modifiche non in conflitto di `theirs` e restituisce un array di descrizioni dei conflitti |
| `xml.diff(a, b, [options])` | Diff strutturale con modifiche applicabili. Opzioni: `key` (attributo ID), `detectMoves` (default true), `includeXml` (default true), `maxXml` |
| `xml.patch(target, changes)` | Applica un diff prodotto da `xml.diff`; risolve tutti i nodi prima di modificare, quindi rimozioni e spostamenti non invalidano i percorsi |

Le modifiche di `xml.diff` contengono `op`, `path` e i dati necessari per
applicarle: `xml` per nodi aggiunti/sostituiti/rimossi, `name`/`old`/`new` per
attributi e testi, `parent`/`index` per modifiche strutturali e
`from`/`to`/`fromIndex` per gli spostamenti. Con `{"key": "id"}` i figli sono
allineati su quell'attributo e il riordino produce operazioni `moved`; senza
chiave il diff resta posizionale. I percorsi usano predicati di chiave come
`/root/item[@id="2"]`, quindi lo stesso diff si applica a un clone.
`xml.patch` supporta `attribute`, `text`, `added`, `removed`, `replaced` e
`moved` e segnala i percorsi non risolti invece di saltarli in silenzio.
`xml.merge3` applica le modifiche non in conflitto e lascia i nodi in conflitto
come sono in `ours`; controlla i conflitti restituiti prima di salvare.

XML → HTML e trasformazioni XML → XML arbitrarie si fanno con XSLT. La
conversione JSON è volutamente convenzionale e può perdere l'ordine del
mixed-content; usa le funzioni DOM quando l'ordine conta. Gli spostamenti sono
rilevati nello stesso genitore; quelli tra genitori diversi diventano una
rimozione più un'aggiunta.

Sicurezza: i dati XML non vengono mai inseriti in HTML; il serializer emette
solo testo XML. Espressioni XPath, schemi XSD e fogli XSLT sono codice dello
sviluppatore. Le dichiarazioni DTD/ENTITY sono rifiutate, i caricamenti remoti
richiedono una allow-list letterale e il parsing è limitato come sopra.

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
| `table.sort(name, column, dir)` | string, string, string | (int, string) | Ordina per key/indice colonna; `asc`, `desc` o vuoto |
| `table.filter(name, column, value)` | string, string, string | (int, string) | Filtra una colonna per testo |
| `table.filterText(name, text)` | string, string | (int, string) | Filtro testo globale |
| `table.page(name, index)` | string, int | (int, string) | Pagina zero-based |
| `table.setPageSize(name, size)` | string, int | (int, string) | Righe per pagina (`0` = tutte) |
| `table.selectedRows(name)` | string | (array[int], string) | Indici delle righe selezionate |
| `table.selectedCells(name)` | string | (string, string) | JSON `[{row,column}]` |
| `table.selectedColumns(name)` | string | (string, string) | Array JSON degli indici colonna |
| `table.lastChange(name)` | string | (string, string) | JSON `{op,row,column,old,new,...}` dell'ultima modifica |
| `table.importText(name, text, spec)` | string, string, string | (int, string) | Parse esplicito: `csv`, `tsv`, `ssv`, `psv`, `json`, `jsonl`, `auto` |
| `table.exportText(name, spec)` | string, string | (string, string) | Serializza in CSV/TSV/SSV/PSV/JSON/JSONL |
| `table.json(name)` | string | (string, string) | JSON con oggetti se c'è un'intestazione |
| `table.clear(name)` | string | (int, string) | Svuota il corpo |
| `table.clearPersisted(name)` | string | (int, string) | Cancella dati/layout persistenti e ripristina la tabella di design |

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
| `chart.focusRow(name, id)` | string, string | (bool, string) | Evidenzia, porta in vista e focalizza una riga tabella (anche non-editable) |
| `chart.pollEvents(name)` | string | (string, string) | Tutti gli eventi pendenti in JSON e svuota la lista |
| `chart.peekEvents(name)` | string | (string, string) | Eventi pendenti in JSON senza svuotare |
| `chart.nextEvent(name)` | string | (string, string) | Evento più vecchio in JSON e lo rimuove |
| `chart.clearEvents(name)` | string | (bool, string) | Svuota la lista eventi |
| `chart.clearPersisted(name)` | string | (int, string) | Cancella lo stato persistente di chart/graph e ripristina il sorgente di design |
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

Il chart è solo una sorgente di eventi. Le interazioni si consumano dalla coda
con `chart.pollEvents` / `chart.peekEvents` / `chart.nextEvent`; il codice SUI
decide quale funzione applicativa chiamare. I callback diretti (`chart.on`,
campi handler sugli elementi e `onShape*`/`onLink*`/`onTag*` chart-level) non
sono disponibili.

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

