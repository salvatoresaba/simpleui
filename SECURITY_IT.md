# Sicurezza di SimpleUI

Questo documento descrive il modello di minaccia, le protezioni del linguaggio
e del runtime quando i dati provengono da **fonti non fidate** (API, WebSocket,
file, CSV/JSON, fogli XLSX/ODS, clipboard, `localStorage`, input utente).

Le protezioni valgono per le app generate da SimpleUI. Il codice SUI stesso è
considerato fidato: chi può scriverlo può già eseguire qualunque azione
consentita all'app.

## Modello di minaccia

| Attore | Controlla | Non controlla |
|---|---|---|
| Sorgente dati remota | testo/JSON/CSV, URL, messaggi WS | il codice SUI, il DOM, il token del bridge |
| File scelto dall'utente | contenuto del file | il codice dell'app |
| Sito esterno | risposte HTTP, redirect, iframe | l'origine dell'app (same-origin policy) |
| Processo locale | file locali, clipboard, rete | il token del bridge (salvo leggerlo dall'app) |

Obiettivi: impedire XSS/HTML injection, esecuzione di codice, path traversal,
comandi arbitrari, formula injection nei CSV, esaurimento di memoria banale.

Fuori ambito: compromissione del sistema operativo o del browser, supply chain
di Go/WebView, codice SUI malevolo, accesso fisico, esfiltrazione tramite canali
legittimi (es. un `http.post` voluto dall'app). Anche il designer visuale e il
compilatore trattano il sorgente `.sui` come **codice fidato**: aprire un `.sui`
di terzi equivale a eseguire il loro codice.

## Protezioni del linguaggio

### 1. Escaping automatico delle interpolazioni

Ogni espressione interpolata in una vista passa da `__sui_esc`:

```sui
state evil = "<img src=x onerror=alert(1)>"
p "{evil}"          # mostra il testo, non crea un <img>
```

- Testo e contenuto delle celle: `&`, `<`, `>`, `"`, `'` vengono escapati.
- Attributi (`value`, `placeholder`, `class`, `style`, `name`, ...): stesso
  escaping, quindi un valore non può chiudere l'attributo e iniettarne altri.
- Anche il testo statico scritto nella vista viene escapato: non esiste un
  meccanismo di "HTML grezzo" tramite interpolazione.
- Componenti: i parametri interpolati sono escapati allo stesso modo.

### 2. Filtro degli URL

Gli attributi `href`, `src` e `link` passano da `__sui_url`:

- consentiti: URL relativi, `http:`, `https:`, `mailto:`, `tel:`;
- bloccati: `javascript:`, `vbscript:`, `data:`, `file:` e altri schemi attivi
  (sostituiti con `#`); i caratteri di controllo vengono ignorati nel controllo
  dello schema.
- eccezione: `data:image/...` è accettato solo nello `src` di `img` e `spinner`,
  dove il browser lo tratta come immagine e non esegue script; resta bloccato in
  `href`, `link` e nei media non immagine.

Resta responsabilità dell'app validare la **destinazione** (allow-list), non solo
lo schema: un URL `https://` può puntare a un endpoint interno (SSRF).

### 3. CSS dinamico: `cssSafe(...)` obbligatorio

Un `style:` **statico** è scritto dallo sviluppatore ed è considerato fidato.
Un `style:` **dinamico** può contenere dati non fidati, quindi il compilatore
lo accetta solo dentro il sanitizer:

```sui
div "ok" style: cssSafe(dynamicCss)
```

`div "x" style: dynamicCss` produce un errore di compilazione. Anche
l'interpolazione in un attributo style richiede il wrapper:

```sui
div "x" style: cssSafe("color: {userColor}")
```

`cssSafe` (runtime `__sui_css`) accetta dichiarazioni `proprietà: valore` e
scarta:

- `url(...)`, `image-set(...)`, `cross-fade(...)`, `paint(...)`, `element(...)`,
  `attr(...)`, `var(...)`, `expression(...)`, `@import`, escape CSS (`\`);
- `position`, `inset`, `top/right/bottom/left`, `z-index` (overlay/redressing);
- `content`, `pointer-events`, `behavior`, `-moz-binding` e proprietà con
  prefisso vendor. Le custom properties CSS sicure (`--nome`) sono ammesse, ma i
  loro valori passano gli stessi filtri.

Il risultato è una stringa CSS con solo le dichiarazioni ammesse. Il designer
avvolge automaticamente le espressioni di stile reattivo in `cssSafe(...)`.

### 3b. Import dati tree inerte

`tree.importJson`/`tree.importXml` trattano il payload come dati dei nodi.
Etichette, key e icone sono scritte con `textContent` e attributi, mai HTML. Le
dichiarazioni DTD/entità XML vengono rifiutate, gli errori di parsing lasciano
l'albero corrente intatto e l'import è limitato per numero di nodi, profondità,
etichetta, key e testo. L'import sostituisce solo la struttura dei nodi; binding
e stile dell'applicazione restano invariati.

### 3c. Dati chart inerti

`chart` tratta l'attributo `data` come testo non fidato. JSON, XML e CSV vengono
convertiti in punti numerici; i nodi SVG sono creati con `createElementNS` e ogni
etichetta/valore viene scritto con `textContent`. Nessuna stringa del payload
raggiunge `innerHTML`. Le dichiarazioni DTD/entità XML vengono rifiutate, i dati
non validi mostrano un breve errore visibile e il parsing è limitato a 5.000
punti, a una dimensione massima del testo e a una lunghezza massima delle
etichette. I dati chart non caricano URL, script o fogli di stile.

### 4. Classi dinamiche: allow-list dell'app

Le classi **statiche** scritte nella vista sono fidate. Per le classi
**dinamiche** il compilatore costruisce una allow-list con:

- tutte le classi definite nel CSS dell'app (`.active`, `.danger`, ...);
- i token letterali nelle `class:` statiche e nelle parti statiche delle
  espressioni (`"btn {state}"`, `cond ? "on" : "off"`, `"btn " + state`).

A runtime `__sui_class(...)` filtra i token: `"active evil"` diventa `"active"`
se `evil` non è dichiarata. Il filtro è automatico, non serve cambiare il
codice:

```sui
div "x" class: dynamicClasses   # -> __sui_class(dynamicClasses)
```

Classi provenienti da CSS esterno non dichiarate nell'app vengono scartate:
dichiarale nel `css` dell'app se devono essere dinamiche.

### 5. Controlli a compile time

Oltre al type checking normale, il compilatore rifiuta:

- `style:` dinamico senza `cssSafe(...)` (sopra);
- classi dinamiche: il compilatore genera la allow-list (CSS + token statici) e
  fa passare ogni espressione da `__sui_class(...)`; le classi sconosciute sono
  scartate a runtime;
- URL **statici** con schema attivo (`javascript:`, `vbscript:`, `data:`) in
  `href`/`src`/`link`;
- `http.*`/`ws.connect` con URL **dinamico** richiedono la variante `*Safe`
  (`getSafe`, `getStatusSafe`, `postSafe`, `connectSafe`) e un’allow-list
  letterale;
- handler evento che non nominano una funzione dichiarata;
- interpolazioni HTML grezze (non esistono: il testo è sempre escapato).

I controlli che dipendono dai dati (host di destinazione, path, comandi, nomi
file) restano a runtime, perché il compilatore non può conoscerne il valore.

### 6. Dati tabellari inerti

`table.display`, `table.importText` e le celle generate usano `textContent`, mai
`innerHTML`: una cella `<img onerror=...>` viene mostrata come testo. Lo stesso
vale per JSON, CSV, TSV, JSON Lines e per i dati XLSX/ODS convertiti.

Limiti applicati dal runtime:

| Limite | Valore |
|---|---|
| righe per tabella | 100.000 |
| colonne per tabella | 10.000 |
| testo importato | 20.000.000 caratteri |

Oltre il limite la chiamata restituisce un errore, senza bloccare la pagina.

### 7. Export CSV e formula injection

`table.exportText` fa il quoting RFC 4180. Per i CSV che verranno aperti in
Excel/LibreOffice è disponibile la neutralizzazione opzionale delle formule:

```sui
csv, err = table.exportText("utenti", "csv;formula=escape")
```

Le celle che iniziano con `=`, `+`, `-`, `@`, TAB o CR vengono prefissate con
`'`. È **opt-in** perché cambia il valore (ad esempio `-5` diventa testo). Il
formato predefinito è `raw`.

### 8. Nessuna valutazione dinamica

Le app generate non usano `eval`, `new Function` o `setTimeout(string)`; gli
handler devono essere funzioni dichiarate (`onClick: fn`). `json.parse` usa
`JSON.parse`.

### 8b. Documenti XML

- Il parser XML rifiuta le dichiarazioni `DOCTYPE`/`ENTITY` prima di `DOMParser`,
  prevenendo XXE ed espansione di entità. I documenti sono limitati a 5.000.000
  caratteri e 200.000 elementi.
- `xml.loadUrlSafe` usa lo stesso guard SSRF e la stessa allow-list letterale di
  `http.getSafe`; si applica il limite di dimensione della risposta.
- Espressioni XPath, schemi XSD e fogli XSLT sono codice fidato dello
  sviluppatore. `xml.validateXsd` implementa un sottoinsieme pratico (niente
  `$ref`, import di schema o substitution group); la validazione DTD non è
  supportata.
- Il serializer emette solo testo XML. I dati XML non vengono mai interpolati
  nell'HTML; la conversione JSON/CSV tratta i valori come dati.
- `xml.xslt` dipende dal `XSLTProcessor` del browser; se non è disponibile
  restituisce un errore esplicito.

### 9. URL dinamici e SSRF

**SSRF** (Server-Side Request Forgery, qui lato browser) significa che l’app
viene indotta a fare una richiesta verso una destinazione scelta dai dati:
localhost, rete locale, metadata cloud, endpoint interni dell’azienda o della
stessa app. CORS spesso impedisce di leggere la risposta, ma **la richiesta
parte comunque** e può avere effetti collaterali.

Politica:

- `http.get`, `http.getStatus`, `http.post` e `ws.connect` accettano solo URL
  **letterali** (scritti dallo sviluppatore).
- Un URL **dinamico** deve passare dalle varianti safe:

```sui
body, err = http.getSafe(userUrl, "https://api.example.com")
st, serr = http.getStatusSafe(userUrl, "https://api.example.com")
out, e2 = http.postSafe(userUrl, payload, "https://api.example.com")
conn, e3 = ws.connectSafe(wsUrl, "wss://socket.example.com")
```

- L’allow-list è un **letterale** (il compilatore rifiuta una variabile): i dati
  non possono allargarla. Voci separate da spazi/virgole:
  - origine assoluta, es. `https://api.example.com` (tutti i path) o
    `https://api.example.com/v1` (solo il prefisso);
  - path della stessa origine, es. `/api/` (richiede lo slash finale per il
    prefisso);
  - per il WebSocket, `ws://host:porta` / `wss://host:porta`.
- Il runtime rifiuta schemi diversi da http(s)/ws(s), credenziali nell'URL,
  caratteri di controllo, host/porta non in allow-list.
- `http.postFiles` carica i byte reali dei file selezionati/drop come multipart
  verso un URL letterale; `http.postFilesSafe` richiede la stessa allow-list
  letterale per destinazioni dinamiche. Il servizio ricevente deve imporre
  limiti di dimensione, numero, nome e contenuto.
- I **redirect sono rifiutati** (`redirect: "error"`), così un URL consentito
  non può rimbalzare su una destinazione interna.
- Limiti: timeout 30 s e risposta max 10 MB.
- `http.get`/`post` con URL letterali restano disponibili per endpoint fissi
  scritti dallo sviluppatore; il compilatore verifica comunque lo schema.

Rischi residui: **DNS rebinding** (un hostname in allow-list che risolve a un IP
privato) e host consentiti ma compromessi o con CORS permissivo. CORS non
impedisce l’invio della richiesta: l’allow-list è il controllo principale.

### 10. Drag-and-drop metadata

I file trascinati o scelti sono **metadati**, non contenuto: il runtime legge
`name`, `path`, `size`, `type` e la struttura delle directory, mai i byte. I
target `dropFiles:` e `pickFiles:` richiedono uno state stringa; i valori
passano dal normale percorso di rendering con escaping. La scansione ricorsiva
ha limiti (2.000 file, profondità 16, 4.096 caratteri per percorso) e gli errori
sono mostrati in un box visibile `Cannot accept the files: ...`. `pickFiles`
crea un input file nascosto per i trigger diversi da `file` e lo rimuove dopo
la selezione o l'annullamento. `pickMode: "fs"` usa la File System Access API
solo dopo gesto utente e permesso, in sola lettura, su Chromium in contesto
sicuro; gli handle non vengono persistiti tra reload. `auto` ripiega sul picker
nativo se l'API non e disponibile.

Il browser non espone percorsi assoluti: un nome può essere testo non fidato e
non dimostra che un percorso desktop sia sicuro da leggere. `fs` resta dietro
`--allow-dir`; validare ogni URL costruito da un nome trascinato.

### 11. Visualizzatore PDF

- Il controllo `pdfviewer` è l'unico frame **non sandboxato** del linguaggio: il
  visualizzatore PDF nativo del browser non parte dentro un frame sandboxato
  (misurato su Chromium ed Electron).
- La sorgente del documento è quindi il confine di sicurezza: un `src`
  letterale è validato come la sorgente di un iframe e un `src` dinamico
  richiede una allow-list letterale `srcAllow` (`pdf.dynamicSrc`, default
  `allowlist`).
- `pdf.unsandboxedFrame: "deny"` toglie il controllo dal linguaggio nei progetti
  che non vogliono un frame non sandboxato.
- Il frame porta sempre `type="application/pdf"` e `referrerpolicy="no-referrer"`,
  così l'URL del documento non trapela nell'header `Referer`.
- Il visualizzatore mostra soltanto il documento: l'app non legge mai i byte del
  PDF, quindi non può estrarne testo, pagine o annotazioni.
- Sul target desktop il documento è servito dal bridge locale:
  - un **secret separato** per ogni esecuzione (non il token del bridge) viene
    consegnato in un cookie `SameSite=Strict` e `HttpOnly` limitato
    all'endpoint dei documenti, quindi non è riutilizzabile sugli altri
    endpoint del bridge;
  - l'endpoint è una `GET` (un iframe non può fare POST), accetta solo i file
    dentro le cartelle autorizzate con `--allow-dir` e serve solo i file i cui
    primi byte sono il magic PDF `%PDF-`;
  - risponde alle richieste di range, così il visualizzatore nativo legge
    documenti grandi;
  - la risposta del documento non porta `X-Frame-Options`/`frame-ancestors`
    (deve essere incorporabile) ma mantiene `nosniff` e `no-referrer`.

## Configurazione della sicurezza

I default sono severi. Un progetto può modificare le policy **di
compilazione** con `simpleui.security.json` (accanto al file `.sui`) oppure con
`--security <file>`:

```bash
simpleui build app.sui --security security.json
simpleui serve app.sui --security security.json
simpleui gui --security security.json       # altrimenti ./simpleui.security.json
# anche lsp/mcp usano ./simpleui.security.json
```

Precedenza: `--security` → `simpleui.security.json` accanto all'entry `.sui` →
default. Un file JSON malformato, un campo sconosciuto o un valore non valido
**bloccano** la compilazione con un messaggio esplicito: nessun dato di
configurazione viene accettato se non è valido.

```json
{
  "style": { "dynamic": "safe" },
  "class": { "dynamic": "filter" },
  "urls": {
    "dynamic": "safe",
    "staticActiveSchemes": "error"
  },
  "table": {
    "maxRows": 100000,
    "maxCols": 10000,
    "maxText": 20000000
  },
  "csv": { "formulaEscapeByDefault": false },
  "iframe": {
    "sandbox": "default",
    "referrerPolicy": "no-referrer",
    "scriptsSameOrigin": "error",
    "dynamicSrc": "allowlist"
  },
  "pdf": {
    "dynamicSrc": "allowlist",
    "unsandboxedFrame": "allow"
  }
}
```

| Campo | Valori | Default | Effetto |
|---|---|---|---|
| `style.dynamic` | `safe`, `warn`, `allow` | `safe` | `safe`: `style:` dinamico richiede `cssSafe`; `warn`: warning ma compila (CSS non sanificato); `allow`: nessun avviso |
| `class.dynamic` | `filter`, `allow` | `filter` | `filter`: allow-list `__sui_class`; `allow`: nessun filtro (solo escaping) |
| `urls.dynamic` | `safe`, `warn`, `allow` | `safe` | `safe`: URL dinamico richiede `*Safe` + allow-list letterale; `warn`/`allow`: consente `http.get(urlDinamico)` |
| `urls.staticActiveSchemes` | `error`, `warn`, `allow` | `error` | Schemi attivi statici in `href`/`src`/`link` o `http.get` letterale |
| `table.maxRows/maxCols/maxText` | interi > 0 | 100000 / 10000 / 20000000 | Limiti runtime del widget tabella |
| `csv.formulaEscapeByDefault` | `true`/`false` | `false` | Default di `;formula=escape` in `table.exportText` |
| `iframe.sandbox` | `default`, `require`, `allow` | `default` | `default`: aggiunge `sandbox=""` se assente; `require`: errore se assente; `allow`: non aggiunge nulla |
| `iframe.referrerPolicy` | policy valida, `unset` | `no-referrer` | `referrerpolicy` aggiunto quando assente; `unset` non aggiunge nulla |
| `iframe.scriptsSameOrigin` | `error`, `warn`, `allow` | `error` | Combinazione `allow-scripts allow-same-origin` |
| `iframe.dynamicSrc` | `allowlist`, `sandbox`, `allow` | `allowlist` | `allowlist`: `src` dinamico richiede `srcAllow`; gli altri valori no |
| `pdf.dynamicSrc` | `allowlist`, `allow` | `allowlist` | `allowlist`: un `src` dinamico di `pdfviewer` richiede `srcAllow`; `allow`: non richiede nulla |
| `pdf.unsandboxedFrame` | `allow`, `deny` | `allow` | `deny`: il controllo `pdfviewer` viene tolto dal linguaggio (il suo frame non è mai sandboxato) |

**Non configurabili** (invarianti di sicurezza):

- escaping HTML di testo e attributi;
- filtro runtime degli schemi per `href`/`src`/`link` (anche con
  `staticActiveSchemes: "allow"` un `javascript:` resta neutralizzato a runtime);
- validazione runtime dell'URL di un iframe dinamico e messaggio visibile
  `Cannot accept the data: ...` (nessun `src` ostile viene caricato);
- celle tabella scritte con `textContent`;
- token del bridge, Host loopback, `X-Frame-Options`, allow-dir con symlink;
- quoting CSV.

`warn` è pensato per la migrazione: il build continua e stampa l’avviso.
`allow` disattiva il controllo, non la protezione runtime sottostante.

## Bridge desktop

Le app desktop sono servite su `127.0.0.1` e possono usare `fs`/`exec` solo
tramite un bridge protetto. Gli endpoint `dialog.pickFile`/`pickFiles`/
`pickDirectory` aprono il dialogo nativo (PowerShell/Windows Forms, `osascript`
oppure `zenity`/`kdialog`) protetti dallo stesso token e richiedono una selezione
esplicita dell'utente. Il percorso restituito non concede accesso: `fs`/`exec`
applicano comunque `--allow-dir`/`--allow-exec`. L'annullamento restituisce un
valore vuoto, non un percorso.

- **token casuale** di 32 byte, confrontato in tempo costante; `app.js` non è
  cacheabile;
- **Host loopback** obbligatorio (`127.0.0.1`, `localhost`, `::1`): blocca il
  DNS rebinding;
- **X-Frame-Options: DENY** e rifiuto delle richieste cross-site;
- `--allow-exec` è una allow-list esatta; il comando è eseguito con
  `exec.Command` (nessuna shell), quindi gli argomenti non vengono interpretati
  da `cmd`/`sh`;
- `--allow-dir` limita lettura/scrittura e risolve i **symlink**: un link dentro
  la directory consentita non può puntare fuori. `fs.copy` e `fs.move`
  controllano sia sorgente sia destinazione; `fs.copy` copia solo file regolari
  (nessuna copia ricorsiva) e `fs.size` rifiuta le directory.
- **Crypto**: `crypto.hash` usa Web Crypto `subtle.digest` con SHA-256/384/512;
  `crypto.uuid`/`crypto.randomBytes` usano il CSPRNG della piattaforma. Non
  esiste un fallback a `Math.random`. `crypto.status()` restituisce `ready`,
  `random-only` o `unavailable`; le funzioni crypto sollevano un errore
  descrittivo invece di restituire valori vuoti, e `crypto.hash` richiede un
  contesto sicuro (HTTPS o localhost). Gli hash non sono una KDF per password.
- **Validazione JSON Schema**: `json.validate` tratta lo schema come codice
  fidato dello sviluppatore e i dati come non fidati. Supporta un sottoinsieme
  pratico; `$ref` è rifiutato. `pattern` è una RegExp dello sviluppatore: evita
  pattern catastrofici su stringhe enormi. La validazione non è un sanitizer:
  continua a usare l'escaping per i dati visualizzati.
- **Quota/sync storage**: `storage.quota` è un dato advisory del browser;
  `storage.sync` ascolta eventi storage locali tra schede e ri-renderizza le
  view. Non è un cloud sync né un servizio di risoluzione conflitti; i valori
  salvati restano dati non fidati.
- nessun bridge nel target web: `serve` espone solo i file statici.

Il token è visibile a un processo locale che legge `app.js`: il bridge non
difende da un processo locale malevolo, solo dal contenuto web.

## Convertitore XLSX/ODS

`sheet2json`/`sheetconv` trattano il foglio come input non fidato:

- nessuna estrazione su disco: si leggono solo le parti ZIP note;
- limite di 64 MiB per parte XML e limiti su righe/colonne/celle;
- `encoding/xml` non risolve entità esterne (niente XXE) e non segue DTD;
- il server ascolta su `127.0.0.1`, accetta solo nomi di file semplici (niente
  path traversal, drive o separatori) e solo `.xlsx`/`.ods`;
- l'output predefinito è JSON, che non esegue formule; per il CSV usare
  `-safe`/`safe=1` se verrà aperto in un foglio di calcolo;
- le formule vengono lette dal valore in cache, non eseguite.

## Iframe: rischi dettagliati

Un `<iframe>` apre un **contesto di navigazione separato** (di solito un’altra
origine) dentro la pagina dell’app. È utile per video, mappe, documenti, ma è
anche una delle superfici più rischiose quando l’URL non è fidato.

Rischi principali:

1. **Phishing e UI redressing**: la pagina incorniciata può imitare l’interfaccia
   dell’app (login, pulsanti, banner) e l’utente può inserire credenziali o
   cliccare azioni credendo che siano dell’app. Senza `sandbox` la pagina può
   anche coprire l’intera area visibile.
2. **Clickjacking dell'app**: un sito esterno può incorporare l'app in un iframe
   e sovrapporre elementi per far compiere azioni all'utente. `serve` e il
   target desktop inviano `X-Frame-Options: DENY` e
   `Content-Security-Policy: frame-ancestors 'none'`; se gli asset vengono
   ospitati altrove, quegli header vanno configurati sul server.
3. **Navigazione del top window**: un iframe non sandboxed può cambiare la
   pagina principale (`top.location`) e portare l’utente su un sito di phishing;
   `sandbox` senza `allow-top-navigation` lo impedisce.
4. **Popup e download**: contenuti ostili possono aprire finestre, richiedere
   permessi o avviare download; servono `sandbox`, `allow-popups`/
   `allow-downloads` solo se necessari e CSP.
5. **Cookie, storage e same-origin**: un iframe ha il proprio origin. Se ha
   `allow-same-origin` ed è **stessa origine** dell’app, può leggere DOM,
   `localStorage`, cookie e token dell’app. La combinazione
   `allow-scripts allow-same-origin` su contenuto della stessa origine può
   annullare l’isolamento del sandbox.
6. **Permessi del browser**: fotocamera, microfono, geolocalizzazione,
   clipboard, pagamenti; vanno limitati con `allow`/Permissions Policy e
   `sandbox`, non concessi per default.
7. **`postMessage`**: la pagina incorniciata può inviare messaggi all’app. Senza
   controlli su `event.origin`/`event.source` e senza validare il payload,
   l’app può eseguire azioni o mostrare dati arbitrari.
8. **Fuga di dati nel referrer**: l’URL dell’app (query, path, token) può
   finire nel `Referer` verso il sito incorniciato; usare
   `referrerpolicy="no-referrer"` e non mettere segreti nell’URL.
9. **SSRF/richieste interne**: anche un iframe è una richiesta. Il filtro
   blocca gli schemi attivi, ma un `https://` può puntare a localhost o alla
   rete interna; per gli URL dinamici `srcAllow` limita le origini, come
   l'allow-list di `http.getSafe`.
10. **Spoofing e fiducia visiva**: un contenuto esterno può sembrare parte
    dell’app; l’utente non vede l’origine. Indicare l’origine e non incorniciare
    contenuti non affidabili.
11. **Malvertising e vulnerabilità del browser**: un iframe può caricare annunci
    o contenuti compromessi; l’app non può sanificare l’HTML/CSS/JS interno.
12. **DoS e risorse**: una pagina pesante o in loop consuma CPU/memoria e può
    rendere l’app non reattiva; `loading="lazy"` e CSP aiutano solo in parte.
13. **`srcdoc`**: non è supportato dal linguaggio; se venisse aggiunto,
    richiederebbe una sanitizzazione HTML esplicita (non basta l’escaping).
14. **CSP**: `simpleui serve` e il target desktop inviano
    `frame-ancestors 'none'` (e `X-Frame-Options: DENY`), quindi l'app non può
    essere incorniciata. Se gli asset sono ospitati altrove, configurare
    `frame-src` per le origini incorporate e `frame-ancestors 'none'` sul
    server.

Mitigazioni implementate:

- attributi `src`, `srcAllow`, `sandbox`, `allow`, `referrerpolicy`,
  `loading`, `width`, `height`;
- **default `sandbox=""`** quando `sandbox` è assente: niente script, form,
  popup o same-origin; `referrerpolicy="no-referrer"` quando assente;
- validazione di compilazione di URL, token `sandbox`, Permissions Policy,
  `referrerpolicy`, `loading` e dimensioni; un dato malformato viene rifiutato
  con un errore esplicito;
- `src` dinamico validato a runtime: schema, credenziali, caratteri di
  controllo, lunghezza e allow-list; se non è accettabile **non viene caricato**
  e compare `Cannot accept the data: ...` (`role="alert"`);
- `srcAllow` obbligatorio per `src` dinamico (default), con allow-list
  letterale di origini/prefix di path;
- `allow-scripts` + `allow-same-origin` è un errore per default;
- `srcAllow`, `allow`, `sandbox`, `referrerpolicy` e dimensioni non possono
  essere espressioni dinamiche;
- `srcdoc` e contenuto di fallback sono rifiutati.

Restano a carico del deployment:

- `frame-src` (e `frame-ancestors` se gli asset sono ospitati fuori da
  `simpleui serve`) per il web;
- un helper per validare `event.origin`/`event.source` in `postMessage`;
- non concedere `allow-scripts`/`allow-same-origin`/`allow-popups` se non
  indispensabili;
- considerare comunque ostile il contenuto incorniciato: sandbox e allow-list
  riducono la superficie, non rendono sicuro un sito compromesso.

## Visualizzatore PDF: rischi dettagliati

Il frame del visualizzatore non è sandboxato, quindi il documento ha tutti i
poteri del frame: può navigare la finestra superiore, aprire popup, scaricare
file e leggere i cookie della propria origine. Le mitigazioni sono:

- il frame è un contesto **documento**: l'endpoint desktop serve
  `application/pdf` con `nosniff`, quindi i byte non vengono mai eseguiti come
  markup;
- la sorgente è limitata da `srcAllow` per gli URL dinamici;
- la referrer policy tiene l'URL del documento fuori dall'header `Referer`.

Rischi residui:

- un documento **della stessa origine** (per esempio il bridge desktop, che
  serve da `127.0.0.1`) condivide l'origine dell'app: contenuti che non sono
  davvero PDF potrebbero raggiungere il DOM dell'app. L'endpoint desktop
  rifiuta tutto ciò che non è un PDF; un URL remoto della stessa origine è
  responsabilità dell'app;
- l'utente non vede l'origine incorporata: mostra la sorgente del documento
  nella UI;
- `pdf.unsandboxedFrame: "deny"` è l'interruttore per i progetti che non
  accettano un frame non sandboxato.

## Rischi residui

Alcuni rischi non sono eliminabili a livello di linguaggio:

- **CSS**: il `style:` statico è codice dello sviluppatore ed è fidato; quello
  dinamico richiede `cssSafe`, che è volutamente conservativo (può scartare
  proprietà/valori legittimi, es. `transition`, `url(...)`). Il sanitizer non è
  un editor CSS completo: non usarlo per layout complessi.
- **Classi**: il filtro elimina le classi sconosciute, ma una classe **nota**
  dell'app può comunque cambiarne l'aspetto (spoofing visivo). Non usare nomi di
  classe che implichino privilegi (es. `admin`) per la sola presentazione.
- **SSRF residuo**: `*Safe` blocca host/redirect/schemi, ma un hostname in
  allow-list può risolvere a un IP privato (DNS rebinding) o essere compromesso;
  CORS limita la lettura, non l'invio. Usa allow-list ristrette e endpoint
  affidabili.
- **Iframe**: il sandbox è vuoto per default e l'URL dinamico è validato a
  runtime, ma la pagina incorporata resta di un'altra origine e può essere
  ostile o compromessa. `srcAllow` controlla l'URL iniziale, non i redirect
  interni all'iframe né il DNS rebinding: usa endpoint fidati. Se si abilitano
  `allow-scripts`, `allow-same-origin`, `allow-popups` o permessi, i rischi
  tornano; `postMessage` e `frame-src` restano responsabilità
  dell'app/deployment.
- **ReDoS**: espressioni regolari costruite da input non fidato possono bloccare
  il thread; limita lunghezza e pattern.
- **Prototype pollution**: evita di fondere oggetti JSON non fidati in oggetti
  condivisi; `merge` crea un nuovo oggetto ma non è un sanitizer.
- **DoS generico**: `json.parse`, loop e regex non hanno timeout; i limiti della
  tabella coprono solo i payload tabellari.
- **CSV**: il quoting evita injection strutturale; `formula=escape` è opt-in.
- **Symlink TOCTOU**: il bridge risolve i symlink prima dell'accesso, ma un
  processo locale potrebbe cambiarli nel frattempo.
- **CSP**: le app generate non impostano una Content-Security-Policy; in
  produzione valuta header aggiuntivi lato server.

## Esempi

La cartella [`examples/security`](examples/security) contiene esempi eseguibili:

| File | Caso |
|---|---|
| `untrusted_text.sui` | testo e attributi ostili, URL `javascript:` |
| `untrusted_style.sui` | CSS dinamico: `cssSafe` obbligatorio e sanitizzazione |
| `untrusted_class.sui` | classi dinamiche filtrate dalla allow-list |
| `untrusted_table.sui` | celle ostili, export CSV grezzo vs sicuro |
| `untrusted_payload.sui` | JSON/CSV arbitrario, errori e limiti |
| `url_guard.sui` | `getSafe`/`getStatusSafe` con allow-list letterale (SSRF) |
| `desktop_files.sui` | nomi file e comandi con capability desktop |

## Segnalazioni

Per una vulnerabilità: non aprire una issue pubblica con un exploit; contatta i
maintainer con una descrizione, la versione e i passi di riproduzione.

