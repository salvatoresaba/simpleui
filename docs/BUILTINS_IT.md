# SimpleUI Builtins Documentation

Questo documento elenca tutte le funzioni builtin disponibili in SimpleUI (non richiedono `import`).

## String Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `len(s)` | string \| array | int | Restituisce la lunghezza di una stringa o array |
| `str(v)` | any | string | Converte un valore in stringa |
| `upper(s)` | string | string | Converte in maiuscolo |
| `lower(s)` | string | string | Converte in minuscolo |
| `trim(s)` | string | string | Rimuove spazi iniziali/finali |
| `cssSafe(s)` | string | string | Sanitizza CSS dinamico per `style:` (blocca `url()`, positioning, vendor props) |
| `trimLeft(s)` | string | string | Rimuove spazi iniziali |
| `trimRight(s)` | string | string | Rimuove spazi finali |
| `contains(s, sub)` | string, string | bool | Verifica se contiene sottostringa |
| `startsWith(s, sub)` | string, string | bool | Verifica se inizia con |
| `endsWith(s, sub)` | string, string | bool | Verifica se termina con |
| `indexOf(s, sub)` | string, string (o array, any) | int | Indice della prima occorrenza (-1 se non trovato) |
| `lastIndexOf(s, sub)` | string, string | int | Indice dell'ultima occorrenza (-1 se non trovato) |
| `slice(s, start, end)` | string, int, int | string | Estrae porzione di stringa |
| `substring(s, start, len)` | string, int, int | string | Estrae `len` caratteri da `start` |
| `repeat(s, count)` | string, int | string | Ripete stringa n volte |
| `padLeft(s, len, pad)` | string, int, string | string | Padding a sinistra |
| `padRight(s, len, pad)` | string, int, string | string | Padding a destra |
| `chars(s)` | string | array[string] | Divide stringa in caratteri |
| `split(s, sep)` | string, string | array[string] | Divide per separatore |
| `splitRegex(s, pattern, [limit])` | string, string, int? | array[string] | Divide con regex (supporta flag g interno) |
| `join(arr, sep)` | array[string], string | string | Unisce array con separatore |
| `replace(s, old, new)` | string, string, string | string | Sostituisce tutte le occorrenze (replaceAll) |
| `replaceFirst(s, old, new)` | string, string, string | string | Sostituisce prima occorrenza |
| `replaceRegex(s, pattern, repl, [flags])` | string, string, string, string? | string | Sostituisce con regex (flags: i,m,s,u,y,g) |
| `match(s, pattern, [flags])` | string, string, string? | array[string] | Trova corrispondenze (array di gruppi/corrispondenza) |
| `matchAll(s, pattern, [flags])` | string, string, string? | array[any] | Tutte le corrispondenze (iteratore-like, restituisce array) |
| `search(s, pattern)` | string, string | int | Indice prima corrispondenza regex (-1 se non trovato) |
| `test(s, pattern)` | string, string | bool | Verifica corrispondenza regex |
| `extract(s, pattern)` | string, string | string | Estrae primo gruppo/capture (o match intero) |
| `extractAll(s, pattern)` | string, string | array[string] | Estrae tutti i gruppi/capture |
| `fmt(n, decimals)` | number, int | string | Formatta numero con decimali (toFixed) |

## Collection Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `append(arr, v)` | array[T], T | array[T] | Aggiunge elemento a array (restituisce nuovo array) |
| `concat(a, b)` | array, array | array | Concatena due array |
| `remove(arr, idx)` | array, int | array | Rimuove elemento all'indice (nuovo array) |
| `includes(arr, v)` | array, any | bool | Verifica presenza elemento |
| `sort(arr)` | array | array | Ordina array (copia) |
| `reverse(arr)` | array | array | Inverte array (copia) |
| `moveItem(arr, from, to)` | array[T], int, int | array[T] | Sposta un elemento in un altro indice (nuovo array) |
| `map(arr, fn)` | array[T], fn(T, [int]) | array[U] | Trasforma ogni elemento; la callback può ricevere `(item)` o `(item, index)` |
| `filter(arr, fn)` | array[T], fn(T, [int]) -> bool | array[T] | Mantiene gli elementi per cui la callback restituisce true |
| `reduce(arr, init, fn)` | array[T], U, fn(U, T, [int]) -> U | U | Riduce l'array a un solo valore |
| `find(arr, fn)` | array[T], fn(T, [int]) -> bool | any | Primo elemento corrispondente, oppure `null` |
| `findIndex(arr, fn)` | array[T], fn(T, [int]) -> bool | int | Indice del primo elemento corrispondente, oppure `-1` |
| `any(arr, fn)` | array[T], fn(T, [int]) -> bool | bool | Se almeno un elemento corrisponde |
| `all(arr, fn)` | array[T], fn(T, [int]) -> bool | bool | Se tutti gli elementi corrispondono |
| `sortBy(arr, fn)` | array[T], fn(T, [int]) -> any | array[T] | Copia ordinata in base alla chiave restituita |
| `unique(arr)` | array[T] | array[T] | Rimuove i duplicati per valore/identità |

## Map/Object Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `keys(m)` | map | array[string] | Chiavi della mappa |
| `values(m)` | map | array[T] | Valori della mappa |
| `merge(m1, m2)` | map, map | map | Unisce mappe |
| `has(m, k)` | map, string | bool | Verifica presenza chiave |
| `removeKey(m, k)` | map, string | map | Rimuove chiave (nuova mappa) |

## Math Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `int(v)` | number | int | Tronca a intero (Math.trunc) |
| `float(v)` | number | float | Converte a numero (Number) |
| `min(a, b)` | number, number | number | Minimo |
| `max(a, b)` | number, number | number | Massimo |
| `abs(n)` | number | number | Valore assoluto |
| `round(n)` | number | int | Arrotonda |
| `floor(n)` | number | int | Parte intera inferiore |
| `ceil(n)` | number | int | Parte intera superiore |
| `pow(a, b)` | number, number | float | Potenza (a^b) |
| `sqrt(n)` | number | float | Radice quadrata |
| `random()` | - | float | Numero casuale [0,1) |
| `randint(min, max)` | int, int | int | Intero casuale [min,max] |

## Timer Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `after(ms, fn)` | number, fn() | null | Esegue funzione dopo delay (ms). fn = nome funzione dichiarata |
| `interval(ms, fn)` | number, fn() | null | Esegue funzione a intervalli regolari. fn = nome funzione dichiarata |
| `timers.cancel(id)` | number | null | Cancella timeout/interval |
| `timers.debounce(fn, ms)` | fn(), number | int | Crea debounced wrapper, restituisce id timer |
| `timers.throttle(fn, ms)` | fn(), number | int | Crea throttled wrapper, restituisce id timer |

## Time/Date Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `time.now()` | - | any (Date) | Data/ora corrente |
| `time.nowMs()` | - | int | Timestamp millisecondi (Date.now) |
| `time.format(dt, [fmt])` | any, string? | string | Formatta data. fmt='iso' o omesso → ISO string |
| `time.parse(s, [fmt])` | string, string? | (any, any) | Parsea stringa data. Restituisce (Date/null, err/null) |
| `time.add(dt, amt, [unit])` | any, number, string? | any (Date) | Aggiunge quantità. unit: ms,s/m/sec, m/min, h/hour, d/day, w/week, mo/month, y/year (default ms) |
| `time.diff(a, b, [unit])` | any, any, string? | int | Differenza a-b. unit: ms,s/sec,m/min,h/hour,d/day (default ms) |
| `time.toISO(dt)` | any | string | Converte a ISO 8601 string |
| `time.toMs(dt)` | any | int | Converte a timestamp ms |

## URL/Encoding Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `url.encode(s)` | string | string | encodeURI |
| `url.decode(s)` | string | (string, any) | decodeURI. Restituisce (value, err/null) |
| `url.encodeComponent(s)` | string | string | encodeURIComponent |
| `url.decodeComponent(s)` | string | (string, any) | decodeURIComponent. Restituisce (value, err/null) |
| `encoding.base64Encode(s)` | string | string | Codifica Base64 (UTF-8 safe) |
| `encoding.base64Decode(s)` | string | (string, any) | Decodifica Base64. Restituisce (value, err/null) |

## Funzioni path

Helper puramente testuali; non accedono al filesystem. I separatori vengono
normalizzati a `/`, i prefissi Windows vengono conservati e i segmenti `.`/`..`
vengono risolti.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `path.join(a, b, ...)` | 1-8 stringhe | string | Unisce e normalizza i segmenti |
| `path.dirname(p)` | string | string | Directory padre (`.` se non esiste) |
| `path.basename(p)` | string | string | Ultimo segmento del percorso |
| `path.extname(p)` | string | string | Estensione con il punto, o `""` |

## Funzioni crypto

Web Crypto sicuro nel browser/desktop. Non esiste fallback a `Math.random`: se
l'API non è disponibile, `crypto.hash` restituisce `""` e `crypto.randomBytes`
restituisce un array vuoto.

| Funzione | Parametri | Ritorno | Async | Descrizione |
|---|---|---|---|---|
| `crypto.available()` | - | bool | false | `true` solo se sono disponibili sia `crypto.subtle` sia `getRandomValues` |
| `crypto.status()` | - | string | false | `ready`, `random-only` o `unavailable` |
| `crypto.hash(algo, text)` | string, string | string | true | Digest hex; `sha256`, `sha384` o `sha512` |
| `crypto.uuid()` | - | string | false | UUID v4 casuale |
| `crypto.randomBytes(n)` | int | array[int] | false | `n` byte sicuri (1-1024), valori 0-255 |

`crypto.hash` è async; chiamalo da una `async fn`. Usa `len()` e `str()` per
ispezionare l'array di byte.

`crypto.subtle` esiste solo in un contesto sicuro (`https://`,
`http://localhost`, `127.0.0.1` o un'app desktop). `crypto.status()` distingue
i casi:

- `ready`: hashing e valori casuali disponibili.
- `random-only`: `getRandomValues` funziona (`uuid`/`randomBytes` funzionano),
  ma `hash` non può girare perché il contesto non è sicuro.
- `unavailable`: Web Crypto assente.

Le funzioni crypto ora falliscono in modo esplicito invece di restituire valori
vuoti:

```sui
if not crypto.available() {
	status = "Crypto hashing richiede HTTPS o localhost: " + crypto.status()
	return
}
```

`crypto.hash` solleva un `Error` per contesto non sicuro o algoritmo non
supportato; `crypto.uuid`/`crypto.randomBytes` sollevano un errore se il CSPRNG
della piattaforma non è disponibile o la lunghezza è fuori da 1-1024. Non esiste
un fallback a `Math.random`.

## Funzioni storage (cross-target)

Web usa `localStorage`; desktop ripiega sulla memoria se non disponibile.
Queste funzioni restituiscono un valore singolo, non coppie `(value, err)`.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `storage.get(k, [def])` | string, any? | any | Legge un valore (JSON parse se possibile); assente -> def/null |
| `storage.set(k, v)` | string, any | bool | Salva un valore (JSON se non stringa) |
| `storage.remove(k)` | string | bool | Rimuove una chiave |
| `storage.clear()` | - | bool | Pulisce lo storage |
| `storage.keys()` | - | array[string] | Elenca le chiavi |
| `storage.quota()` | - | map[string]int | Async; `{used, quota}` da `navigator.storage.estimate` (fallback dimensione localStorage) |
| `storage.sync()` | - | bool | Installa un listener `storage` cross-tab; le altre schede ri-renderizzano al cambio di una chiave |

`storage.quota` è async; chiamalo da una `async fn`. `storage.sync` è
sincronizzazione locale tra schede, non un'API cloud: ascolta gli eventi di
storage del browser e ri-renderizza le view; l'applicazione decide come unire i
dati.

## Clipboard Functions

Web: richiede permessi/HTTPS (Clipboard API). Desktop: via bridge (se disponibile).

| Function | Params | Return | Description |
|---|---|---|---|
| `clipboard.readText()` | - | (string, any) | Legge testo dagli appunti. Restituisce (text, err/null) |
| `clipboard.writeText(t)` | string | (bool, any) | Scrive testo negli appunti. Restituisce (ok, err/null) |

## Validation Functions

| Function | Params | Return | Description |
|---|---|---|---|
| `validation.isEmail(s)` | string | bool | Valida formato email (regex) |
| `validation.isUrl(s)` | string | bool | Valida URL (new URL) |
| `validation.isEmpty(v)` | any | bool | Vuoto: null/undefined, stringa vuota/trimmata, array vuoto, oggetto vuoto |
| `validation.required(v)` | any | bool | Obbligatorio: !isEmpty |
| `validation.minLen(v, n)` | any, number | bool | Lunghezza minima (string/array/oggetto: numero chiavi/elementi) |
| `validation.maxLen(v, n)` | any, number | bool | Lunghezza massima |
| `validation.length(v, a, [b])` | any, number, number? | bool | Lunghezza esatta (2 args) o range [a,b] (3 args) |
| `validation.pattern(s, regex)` | string, string | bool | Test regex (new RegExp) |
| `validation.isNumber(v)` | any | bool | typeof number e non NaN |
| `validation.isBool(v)` | any | bool | typeof boolean |
| `validation.isString(v)` | any | bool | typeof string |
| `validation.isArray(v)` | any | bool | Array.isArray |
| `validation.isObject(v)` | any | bool | Object non-null non-array |

## Drag and Drop Functions

Gli handler drag non ricevono argomenti; questi accessor leggono il payload
catturato dal runtime browser. Vedi
[`API_REFERENCE_IT.md`](API_REFERENCE_IT.md#drag-and-drop) per gli attributi e la
[guida all'editor GUI](GUI_EDITOR_IT.md) per la configurazione nell'inspector.

| Function | Params | Return | Description |
|---|---|---|---|
| `drag.source()` | - | string | Chiave sorgente da `dragKey` |
| `drag.target()` | - | string | Chiave target da `dropKey` |
| `drag.data()` | - | string | Payload `text/plain` |
| `drag.files()` | - | array[map] | Metadati file (`name`, `path`, `size`, `type`, `isDirectory`) |
| `drag.folder()` | - | string | Cartella top-level comune della selezione, o `""` |
| `drag.error()` | - | string | Ultimo motivo di rifiuto |
| `drag.active()` | - | bool | Drag attivo |
| `drag.clear()` | - | null | Azzera lo stato drag |

`drag.files()` contiene i metadati immediati. La raccolta ricorsiva delle
directory e asincrona e riempie uno state `dropFiles:` al termine. Anche una
selezione da `pickFiles:` popola `drag.files()`.

## Navigation

| Function | Params | Return | Description |
|---|---|---|---|
| `navigate(name)` | string | null | Mostra un'altra pagina (view root non invocata) |
| `modal.open(name)` | string | null | Apre il modale con quel `name` |
| `modal.close(name)` | string | null | Chiude il modale con quel `name` |

## Funzioni di internazionalizzazione

`i18n "locale" { ... }` dichiara un catalogo; `t` e gli helper `i18n.*` non
richiedono import. Il primo catalogo dichiarato è il fallback. Vedi
[Riferimento API: i18n](API_REFERENCE_IT.md#internazionalizzazione-i18n).

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `t(key, [vars])` | string, map? | string | Traduce una chiave; `vars` fornisce i valori `{placeholder}` e `count` per i plurali |
| `i18n.t(key, [vars])` | string, map? | string | Alias di `t` |
| `i18n.locale()` | - | string | Codice della lingua attiva |
| `i18n.setLocale(code)` | string | bool | Cambia lingua e salva la scelta; `"auto"` riesegue il rilevamento |
| `i18n.detect()` | - | string | Migliore corrispondenza caricata per le lingue del browser |
| `i18n.available()` | - | array[string] | Codici lingua caricati |
| `i18n.dir()` | - | string | `"ltr"` o `"rtl"` |
| `i18n.has(key)` | string | bool | Se la chiave esiste nel catalogo attivo o di fallback |
| `i18n.missing()` | - | array[string] | Chiavi presenti nel fallback ma assenti nella lingua attiva |

`navigate` non richiede import. È un no-op nelle app a pagina singola e con nomi
di pagina sconosciuti. Usarlo da un handler senza argomenti:

```sui
fn openSettings() -> null {
	navigate("settings")
}
```

## Assertion e formattazione locale

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `assert(cond, message)` | bool, string | null | Solleva un `Error` se `cond` è falso; utile con `errorboundary` |
| `intl.date(value, [style], [locale], [timeZone])` | string \| int, string?, string?, string? | string | Formatta una data; `style` è `short`, `medium`, `long`, `full` o `iso` |
| `intl.time(value, [style], [locale], [timeZone])` | string \| int, string?, string?, string? | string | Formatta un orario |
| `intl.datetime(value, [style], [locale], [timeZone])` | string \| int, string?, string?, string? | string | Formatta data e orario |
| `intl.number(value, [decimals], [locale])` | number, int?, string? | string | Formattazione decimale con separatori |
| `intl.currency(value, currency, [locale])` | number, string, string? | string | Formattazione valuta con codice ISO 4217 |
| `intl.percent(value, [decimals], [locale])` | number, int?, string? | string | Percentuale da una frazione |
| `intl.relative(offset, unit, [locale], [numeric])` | int, string, string?, string? | string | Tempo relativo come `2 giorni fa` |
| `intl.timezone()` | - | string | Fuso orario IANA risolto |

Le locale omesse usano la locale i18n attiva, poi la lingua del browser. Vedi
[API reference: Intl](API_REFERENCE_IT.md#formattazione-locale-intl).

## Sicurezza

`cssSafe` è obbligatorio per i valori **dinamici** di `style:`:

```sui
div "x" style: cssSafe(dynamicCss)   # ok
div "x" style: dynamicCss            # errore di compilazione
```

Il sanitizer è conservativo: elimina `url()`, `paint()`, `expression()`, escape
CSS, proprietà di posizionamento/overlay e prefissi vendor. Gli stili statici
scritti nella vista sono considerati codice fidato. Vedi
[`SECURITY.md`](../SECURITY_IT.md).

## HTTP Headers avanzati — non disponibili
Le firme attuali sono `http.get(url)` e `http.post(url, body)`. Custom headers e timeout override non sono ancora parte del contratto stdlib.


