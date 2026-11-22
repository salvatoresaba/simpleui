# Documentazione Stdlib Dati di SimpleUI

Questo documento copre i moduli std puri e in-memory. Non accedono alla rete, al filesystem, alle variabili di processo o al DOM. Sono disponibili sia per target web sia desktop e seguono la convenzione std abituale: `value, err = modulo.funzione(...)` con `err = null` in caso di successo. Tutte le funzioni di questi moduli sono sincrone, con due eccezioni: `files.readBytes` e `files.documentUrl` sono asincrone perché leggono i byte di un file selezionato dall'utente.

I moduli sono `bytes`, `bin`, `files`, `mime`, `color`, `units`, `print`, `history`, `zip`, `archive`, `csv`, `env`, `capabilities`, `events`, `bus`, `log` e `telemetry`. Per i moduli I/O (`http`, `ws`, `fs`, `exec`, `dialog`) vedere [STDLIB](STDLIB_IT.md).

## Import

```sui
import { bytes, csv, zip, color } from "std"
```

Vengono emessi solo i moduli importati nel runtime generato. Importare un solo modulo di questo documento non trascina gli altri.

## Modulo Bytes (`bytes`)

Gli array di byte sono valori `array[int]` contenenti 0-255. Gli input vengono validati; il buffer massimo è di 10.000.000 byte.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `bytes.fromString(text, [encoding])` | string, string? | (array[int], string) | Codifica il testo come `utf-8` (predefinito), `utf-16le` o `utf-16be` |
| `bytes.toString(bytes, [encoding])` | any, string? | (string, string) | Decodifica i byte; UTF-8 non valido restituisce un errore |
| `bytes.fromBase64(text)` | string | (array[int], string) | Decodifica base64 |
| `bytes.toBase64(bytes)` | any | (string, string) | Codifica base64 a blocchi (nessun limite dello stack) |
| `bytes.fromHex(text)` | string | (array[int], string) | Decodifica esadecimale, ignorando gli spazi |
| `bytes.toHex(bytes)` | any | (string, string) | Codifica esadecimale minuscola |
| `bytes.concat(a, b)` | any, any | (array[int], string) | Concatena due buffer |
| `bytes.slice(bytes, start, [end])` | any, int, int? | (array[int], string) | Copia un intervallo di byte |
| `bytes.indexOf(bytes, needle, [start])` | any, any, int? | (int, string) | Primo indice di una sequenza di byte, oppure `-1` |
| `bytes.equal(a, b)` | any, any | (bool, string) |uguaglianza byte per byte |
| `bytes.length(bytes)` | any | (int, string) | Lunghezza del buffer |
| `bytes.fill(length, byte)` | int, int | (array[int], string) | Crea un buffer riempito |
| `bytes.reverse(bytes)` | any | (array[int], string) | Copia e inverte un buffer |
| `bytes.xor(a, b)` | any, any | (array[int], string) | XOR byte per byte; le lunghezze devono corrispondere |

## Modulo Binary (`bin`)

`bin` legge e scrive interi e float IEEE-754 con `DataView`. L'argomento `little` predefinito è `true` per i valori multi-byte. Gli offset vengono controllati.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `bin.u8(bytes, offset)` | any, int | (int, string) | Legge un byte unsigned |
| `bin.u16(bytes, offset, [little])` | any, int, bool? | (int, string) | Legge `uint16` |
| `bin.u32(bytes, offset, [little])` | any, int, bool? | (int, string) | Legge `uint32` |
| `bin.i8(bytes, offset)` | any, int | (int, string) | Legge `int8` |
| `bin.i16(bytes, offset, [little])` | any, int, bool? | (int, string) | Legge `int16` |
| `bin.i32(bytes, offset, [little])` | any, int, bool? | (int, string) | Legge `int32` |
| `bin.f32(bytes, offset, [little])` | any, int, bool? | (float, string) | Legge `float32` |
| `bin.f64(bytes, offset, [little])` | any, int, bool? | (float, string) | Legge `float64` |
| `bin.writeU8(value)` | int | (array[int], string) | Scrive un byte unsigned |
| `bin.writeU16(value, [little])` | int, bool? | (array[int], string) | Scrive `uint16` |
| `bin.writeU32(value, [little])` | int, bool? | (array[int], string) | Scrive `uint32` |
| `bin.writeI8(value)` | int | (array[int], string) | Scrive `int8` |
| `bin.writeI16(value, [little])` | int, bool? | (array[int], string) | Scrive `int16` |
| `bin.writeI32(value, [little])` | int, bool? | (array[int], string) | Scrive `int32` |
| `bin.writeF32(value, [little])` | float, bool? | (array[int], string) | Scrive `float32` |
| `bin.writeF64(value, [little])` | float, bool? | (array[int], string) | Scrive `float64` |

## Modulo Files (`files`)

Questi helper validano i nomi e **non** leggono né scrivono file. Per vero I/O usare il modulo desktop `fs`. I nomi file sono considerati non fidati.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `files.name(path)` | string | (string, string) | Nome base; vengono accettati sia `/` sia `\` |
| `files.stem(path)` | string | (string, string) | Nome base senza ultima estensione |
| `files.ext(path)` | string | (string, string) | Estensione minuscola inclusa di `.` |
| `files.split(path)` | string | (map[string]string, string) | Mappa con `name`, `stem` e `ext` |
| `files.sanitize(name, [maxLen])` | string, int? | (string, string) | Rimuove separatori, controlli e punti iniziali; limite predefinito 120 |
| `files.validate(name, [options])` | string, any? | (bool, string) | Valida nome, flag hidden, estensione e dimensione |
| `files.isHidden(name)` | string | (bool, string) | Se il nome base inizia con `.` |
| `files.extAllowed(name, allow)` | string, string | (bool, string) | Controlla una allow-list di estensioni |
| `files.readBytes(name)` | string | (array[int], string) | Legge i byte di un file della selezione corrente (solo web) |
| `files.documentUrl(name)` | string | (string, string) | URL del browser (`blob:`) di un PDF della selezione corrente (solo web) |

`files.readBytes(name)` legge i byte di un file della selezione corrente
(`pickFiles:`/`dropFiles:`). Il browser espone solo i file scelti o trascinati,
quindi non legge mai un percorso: `name` corrisponde al nome o al percorso
relativo del file selezionato, e un nome vuoto legge l'unico file selezionato.
I file oltre 10 MB vengono rifiutati. Su desktop il picker restituisce percorsi:
usare li `fs.readBytes(path)` (richiede la capability `--allow-dir`).

`files.documentUrl(name)` trasforma il file selezionato in un URL che il
visualizzatore PDF può caricare, così l'utente apre un documento da questo
dispositivo senza backend. I byte devono iniziare con il magic PDF `%PDF-` e il
tipo del blob è forzato a `application/pdf`: il visualizzatore nativo non
riceve mai un documento che il browser potrebbe eseguire. L'URL vive finché
vive la pagina. Su desktop usare invece `fs.documentUrl(path)`.

```sui
url, err = files.documentUrl(picked)
if err == null {
	docUrl = url
}
```

```sui
pdfviewer src: docUrl srcAllow: "/"
```

Le opzioni di `files.validate` sono `maxName` (predefinito 200), `allowHidden` (predefinito false), `allowedExt` o `allow` (elenco separato da virgole/spazi o array), `size` e `maxSize`. Traversal, separatori di percorso, lettere di unità, caratteri di controllo e nomi nascosti vengono rifiutati per predefinito.

## Modulo MIME (`mime`)

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `mime.fromName(name)` | string | (string, string) | Tipo MIME dall'estensione; fallback `application/octet-stream` |
| `mime.fromBytes(bytes)` | any | (string, string) | Rileva firme comuni di immagini, PDF, ZIP, XML e JSON |
| `mime.extension(mimeType)` | string | (string, string) | Estensione suggerita inclusa di `.`, oppure stringa vuota |
| `mime.isAllowed(mimeType, allow)` | string, string | (bool, string) | Allow-list esatta, con wildcard (`image/*`) o `*/*` |
| `mime.validate(name, mimeType, size, [options])` | string, string, int, any? | (bool, string) | Combina controlli su nome, dimensione, estensione e MIME |

`mime.validate` riutilizza le opzioni di `files.validate` e aggiunge `allow` per i tipi MIME. Entrambi gli helper restituiscono `(false, error)` in caso di rifiuto.

## Modulo Color (`color`)

I colori sono mappe con `r`, `g`, `b` numerici (0-255) e `a` (0-1). Il parsing supporta `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa`, `rgb()`, `rgba()`, `hsl()`, `hsla()` e un piccolo insieme di colori nominati. I valori non supportati generano un errore; nessun CSS viene iniettato o valutato.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `color.parse(value)` | string | (any, string) | Analizza in una mappa colore |
| `color.format(color, [format])` | any, string? | (string, string) | Formatta come `hex` (predefinito), `rgb`, `rgba`, `hsl` o `hsla` |
| `color.rgb(r, g, b, [a])` | int, int, int, float? | (any, string) | Crea una mappa colore |
| `color.hsl(h, s, l, [a])` | float, float, float, float? | (any, string) | Converte HSL in una mappa colore |
| `color.mix(a, b, t)` | any, any, float | (any, string) | Interpola con `t` limitato a 0-1 |
| `color.withAlpha(color, alpha)` | any, float | (any, string) | Sostituisce il canale alpha |
| `color.luminance(color)` | any | (float, string) | Luminanza relativa WCAG |
| `color.contrast(a, b)` | any, any | (float, string) | Rapporto di contrasto WCAG |
| `color.isDark(color)` | any | (bool, string) | Se la luminanza è inferiore a 0.5 |

## Modulo Units (`units`)

Sono supportate le unità CSS assolute: `px`, `mm`, `cm`, `in`, `pt`, `pc` e `q`. La base pixel è 96 dpi. Le unità relative come `em`, `rem` e `%` vengono rifiutate dalle funzioni di conversione perché la loro base dipende dal documento.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `units.parse(value)` | string | (any, string) | Divide una lunghezza in `value` e `unit` |
| `units.convert(value, from, to)` | float, string, string | (float, string) | Converte tra unità assolute |
| `units.toBase(value, unit)` | float, string | (float, string) | Converte in pixel CSS |
| `units.fromBase(value, unit)` | float, string | (float, string) | Converte dai pixel CSS |
| `units.format(value, unit, [precision])` | float, string, int? | (string, string) | Testo a precisione fissa senza zeri finali |

## Modulo Print (`print`)

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `print.pxToMm(px)` | float | (float, string) | Pixel in millimetri a 96 dpi |
| `print.mmToPx(mm)` | float | (float, string) | Millimetri in pixel a 96 dpi |
| `print.ptToMm(pt)` | float | (float, string) | Punti in millimetri |
| `print.paper(name, [orientation])` | string, string? | (map[string]any, string) | `a3`, `a4`, `a5`, `letter`, `legal`, `tabloid`; predefinito `portrait` |
| `print.fitBox(width, height, maxW, maxH)` | float, float, float, float | (map[string]any, string) | Riduce un rettangolo mantenendo le proporzioni |

`print.paper` restituisce `width`, `height`, `unit` (`"mm"`) e `orientation`. `print.fitBox` restituisce `width`, `height` e `scale`.

## Modulo History (`history`)

Le history sono stack in memoria con nome. I valori vengono copiati in profondità sia in scrittura sia in lettura, quindi modifiche successive ai dati del chiamante non possono cambiare uno stato salvato. I nomi ammettono lettere, cifre, `_`, `.`, `:` e `-` (max 128).

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `history.push(name, value, [limit])` | string, any, int? | (int, string) | Salva uno snapshot; limite predefinito 50, massimo 1000 |
| `history.undo(name)` | string | (any, string) | Torna indietro di un passo e restituisce lo snapshot |
| `history.redo(name)` | string | (any, string) | Avanza di un passo e restituisce lo snapshot |
| `history.current(name)` | string | (any, string) | Restituisce lo snapshot corrente senza spostarsi |
| `history.canUndo(name)` | string | (bool, string) | Se esiste un passo di undo |
| `history.canRedo(name)` | string | (bool, string) | Se esiste un passo di redo |
| `history.clear(name)` | string | (null, string) | Rimuove tutti gli snapshot |
| `history.list()` | none | (array[string], string) | Nomi history ordinati |
| `history.size(name)` | string | (int, string) | Numero di snapshot |
| `history.jump(name, index)` | string, int | (any, string) | Si sposta a uno snapshot tramite indice |

Valori ciclici e valori annidati oltre 64 livelli vengono rifiutati.

## Modulo ZIP (`zip`)

Gli archivi ZIP vengono creati in memoria e usano solo il metodo **stored**. Questo mantiene piccolo il runtime ed elimina le bombe di decompressione. Vengono accettate fino a 1.000 entry, 10 MB per entry e 50 MB per archivio.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `zip.create()` | none | (any, string) | Crea un nuovo writer handle |
| `zip.add(writer, name, bytes)` | any, string, any | (int, string) | Aggiunge un'entry binaria; restituisce il numero di entry |
| `zip.addText(writer, name, text)` | any, string, string | (int, string) | Aggiunge un'entry di testo UTF-8 |
| `zip.finish(writer)` | any | (array[int], string) | Emette i byte dell'archivio e chiude il writer |
| `zip.open(bytes)` | any | (any, string) | Analizza e valida un archivio |
| `zip.list(archive)` | any | (array[any], string) | Mappe entry: `name`, `size`, `crc`, `method` |
| `zip.read(archive, name)` | any, string | (array[int], string) | Legge i byte dopo verifica CRC |
| `zip.readText(archive, name)` | any, string | (string, string) | Decodifica un'entry come UTF-8 |
| `zip.has(archive, name)` | any, string | (bool, string) | Se un'entry esiste |
| `zip.count(archive)` | any | (int, string) | Numero di entry |

I nomi entry devono essere percorsi UTF-8 relativi. Vengono rifiutati percorsi assoluti, lettere di unità Windows, backslash, segmenti `.`/`..`, entry di directory, byte NUL, entry cifrate, metodi compressi e collegamenti simbolici. I CRC vengono verificati in lettura.

## Modulo Archive (`archive`)

`archive` è un wrapper orientato alle mappe sullo stesso writer/reader ZIP.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `archive.pack(files, [format])` | any, string? | (array[int], string) | Mappa nome → array di byte in byte ZIP |
| `archive.unpack(bytes, [format])` | any, string? | (map[string]any, string) | Byte ZIP in mappa nome → array di byte |
| `archive.packText(files, [format])` | any, string? | (array[int], string) | Mappa nome → testo in byte ZIP |
| `archive.unpackText(bytes, [format])` | any, string? | (map[string]string, string) | Byte ZIP in mappa nome → testo |
| `archive.info(bytes)` | any | (array[any], string) | Analizza ed elenca i metadati delle entry |

È supportato solo `format = "zip"`; qualsiasi altro valore restituisce un errore. Le chiavi object chiamate `__proto__` vengono rifiutate durante la costruzione delle mappe.

## Modulo CSV (`csv`)

Il parsing CSV segue lo stile di quoting RFC 4180: le virgolette raddoppiate escapescono una virgoletta e i campi citati possono contenere delimitatori o nuove righe. Vengono accettate fino a 10.000 righe e 1.000 colonne. Le opzioni sono `delimiter` (un carattere), `quote`, `newline`, `columns` e `header` (predefinito true per le esportazioni).

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `csv.parse(text, [options])` | string, any? | (array[any], string) | Righe come array di stringhe; BOM iniziale ignorato |
| `csv.parseObjects(text, [options])` | string, any? | (array[any], string) | La prima riga diventa chiavi object; duplicati rifiutati |
| `csv.stringify(rows, [options])` | any, any? | (string, string) | Serializza array di righe |
| `csv.stringifyObjects(rows, [options])` | any, any? | (string, string) | Serializza mappe usando `columns` o le chiavi della prima riga |
| `csv.validate(text, [options])` | string, any? | (bool, string) | Analizza e richiede conteggio colonne coerente |

Le esportazioni escapano sempre le formule dei fogli elettronici: una cella che inizia con `=`, `+`, `-`, `@`, tab o ritorno carrello viene prefissata con un apostrofo. Non è configurabile, perché i CSV vengono spesso aperti da applicazioni foglio. I dati importati in una tabella devono comunque passare da `table.display`.

## Modulo Environment (`env`)

`env` espone solo i seguenti fatti di sola lettura; non legge mai variabili di processo e non scrive mai su di esse:

| Chiave | Significato |
|---|---|
| `SUI_TARGET` | Target di compilazione, ad esempio `web` o `desktop` |
| `SUI_PLATFORM` | Piattaforma riportata dal browser, oppure `unknown` |
| `SUI_LANGUAGE` | Lingua del browser, oppure `en` |
| `SUI_TIMEZONE` | Time zone IANA, oppure `UTC` |
| `SUI_USER_AGENT` | Stringa user-agent del browser, oppure vuota |
| `SUI_DESKTOP` | `"true"` per target desktop |

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `env.target()` | none | (string, string) | Uguale a `SUI_TARGET` |
| `env.platform()` | none | (string, string) | Uguale a `SUI_PLATFORM` |
| `env.get(name)` | string | (string, string) | Legge una chiave allow-list; chiavi sconosciute restituiscono errore |
| `env.has(name)` | string | (bool, string) | Se la chiave è esposta |
| `env.all()` | none | (map[string]string, string) | Copia di tutte le chiavi esposte |
| `env.language()` | none | (string, string) | Uguale a `SUI_LANGUAGE` |
| `env.timezone()` | none | (string, string) | Uguale a `SUI_TIMEZONE` |

I nomi sono case-insensitive per `get` e `has`. I valori sono calcolati dal runtime e dal browser correnti; sono informativi e non vanno usati per decisioni di autorizzazione.

## Modulo Capabilities (`capabilities`)

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `capabilities.has(name)` | string | (bool, string) | Se la capability è nota in questa build |
| `capabilities.check(name)` | string | (bool, string) | Come `has`, ma restituisce errore se non disponibile |
| `capabilities.list()` | none | (array[string], string) | Nomi capability ordinati |
| `capabilities.target()` | none | (string, string) | Uguale a `env.target()` |

I nomi noti sono `storage`, `clipboard`, `crypto`, `fs`, `exec`, `dialog`, `notifications`, `intl`, `typedArrays` e `textEncoder`. `fs`, `exec` e `dialog` sono true solo per target desktop; per funzionare richiedono comunque i permessi di build `--allow-dir` / `--allow-exec`.

## Modulo Events (`events`)

Il bus è locale all'applicazione in esecuzione. I listener sono valori callback JavaScript passati per nome, ad esempio `events.on("data", onData)`. I payload vengono copiati in profondità prima della consegna dove possibile. Gli errori lanciati da un listener vengono restituiti da `emit` dopo che tutti i listener sono stati eseguiti.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `events.on(name, fn)` | string, any | (int, string) | Registra un listener; restituisce il numero di listener |
| `events.once(name, fn)` | string, any | (int, string) | Registra un listener rimosso dopo prima chiamata |
| `events.off(name, fn)` | string, any | (int, string) | Rimuove i listener corrispondenti; restituisce il numero rimosso |
| `events.emit(name, payload)` | string, any | (int, string) | Notifica i listener; restituisce il numero chiamato |
| `events.clear(name)` | string | (null, string) | Rimuove tutti i listener per l'evento |
| `events.count(name)` | string | (int, string) | Numero di listener |
| `events.names()` | none | (array[string], string) | Nomi evento ordinati con listener |

## Modulo Event Bus (`bus`)

`bus` è un alias di `events` con le stesse firme e comportamento. Esiste per consentire alle applicazioni di scegliere la terminologia più adatta al proprio design:

`bus.on`, `bus.once`, `bus.off`, `bus.emit`, `bus.clear`, `bus.count` e `bus.names`.

## Modulo Log (`log`)

I log restano in memoria: buffer circolare di 500 voci, messaggi di 2.000 caratteri e nessun trasporto di rete. Chiavi che sembrano segreti e pattern token comuni vengono redatti prima che una voce venga memorizzata. La soglia di livello è `debug` < `info` < `warn` < `error`; il predefinito è `info`.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `log.debug(message, [context])` | string, any? | (null, string) | Aggiunge una voce debug |
| `log.info(message, [context])` | string, any? | (null, string) | Aggiunge una voce info |
| `log.warn(message, [context])` | string, any? | (null, string) | Aggiunge una voce warning |
| `log.error(message, [context])` | string, any? | (null, string) | Aggiunge una voce error |
| `log.setLevel(level)` | string | (null, string) | Imposta la soglia |
| `log.level()` | none | (string, string) | Soglia corrente |
| `log.clear()` | none | (null, string) | Cancella le voci memorizzate |
| `log.entries([limit])` | int? | (array[any], string) | Ultime voci, fino a 500; predefinito 100 |
| `log.export()` | none | (string, string) | Testo JSON delle voci memorizzate |
| `log.redact(text)` | string | (string, string) | Redatta i pattern segreti nel testo |
| `log.redactObject(value)` | any | (any, string) | Copia profonda con chiavi segrete redatte |

I nomi chiave redatti includono `password`, `token`, `secret`, `apiKey`, `api_key`, `authorization`, `cookie`, `session`, `creditCard`, `cardNumber`, `cvv` e `pin`. Anche i token `Bearer` vengono sostituiti.

## Modulo Telemetry (`telemetry`)

La telemetria è un registro metriche locale. Non esporta né invia dati.

| Funzione | Parametri | Ritorno | Descrizione |
|---|---|---|---|
| `telemetry.counter(name, [delta])` | string, float? | (float, string) | Aumenta un counter; delta predefinito 1 |
| `telemetry.gauge(name, value)` | string, float | (float, string) | Imposta un valore gauge |
| `telemetry.timing(name, ms)` | string, float | (float, string) | Registra una durata in millisecondi |
| `telemetry.snapshot()` | none | (map[string]any, string) | Copia di tutte le metriche |
| `telemetry.reset()` | none | (null, string) | Rimuove tutte le metriche |
| `telemetry.names()` | none | (array[string], string) | Nomi metrica ordinati |

I nomi metrica ammettono lettere, cifre, `_`, `.`, `:` e `-` (max 128). Ogni voce snapshot contiene `name`, `type`, `value` e `updated`.

## Note di sicurezza

- Non viene introdotto injection di HTML o CSS grezzo. Questi moduli restituiscono solo dati.
- `files` e `mime` validano nomi non fidati ma non concedono accesso filesystem; le capability desktop valgono comunque per `fs`, `exec` e `dialog`.
- La lettura ZIP accetta solo entry stored, rifiuta nomi non sicuri e collegamenti simbolici, applica limiti di dimensione e verifica i CRC.
- L'escaping delle formule CSV è sempre attivo in esportazione.
- `env` è una vista di sola lettura con allow-list di fatti runtime, non un archivio segreti.
- I log redatta i segreti comuni, ma le applicazioni devono comunque evitare di loggare dati sensibili con nomi chiave insoliti.

## Esempio

Vedere `examples/stdlib_data.sui` per un'app eseguibile che esercita ogni modulo di questo documento, incluso round-trip ZIP, escaping formule CSV, event bus e log in memoria. Validare e compilare con:

```bash
./simpleui.exe check examples/stdlib_data.sui --json
./simpleui.exe build examples/stdlib_data.sui --out dist
```
