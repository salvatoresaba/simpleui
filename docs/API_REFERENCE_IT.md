# SimpleUI API Reference

Guida completa alle API di SimpleUI: linguaggio, builtin, stdlib, CLI.

## Sommario
- [Linguaggio SUI](#linguaggio-sui-sintassi-essenziale)
- [Enum e switch](#enum-e-switch)
- [Tipi union e type guard `is`](#tipi-union-e-type-guard-is)
- [Literal types](#literal-types)
- [Generici e parametri rest](#generici-e-parametri-rest)
- [Costanti, valori computed e campi readonly](#costanti-valori-computed-e-campi-readonly)
- [Aggiornamenti reattivi, batching, error boundary e chiavi delle liste](#aggiornamenti-reattivi-batching-error-boundary-e-chiavi-delle-liste)
- [Destructuring, spread e argomenti nominati](#destructuring-spread-e-argomenti-nominati)
- [Menu bar](#menu-bar)
- [Spinner](#spinner)
- [Internazionalizzazione (i18n)](#internazionalizzazione-i18n)
- [Formattazione locale (Intl)](#formattazione-locale-intl)
- [Tabs](#tabs)
- [Pagine e navigazione](#pagine-e-navigazione)
- [Dialoghi modali](#dialoghi-modali)
- [Table layout](#table-layout)
- [Tabelle e dati non fidati](#tabelle-e-dati-non-fidati)
- [Sicurezza](#sicurezza)
- [Builtins](BUILTINS_IT.md)
- [Stdlib](STDLIB_IT.md)
- [CLI](#cli)

## Linguaggio SUI (Sintassi Essenziale)

### Tipi
- `int`, `float`, `string`, `bool`, `null`, `any`
- `array[T]`, `map[string]T`
- Tipi strutturati: `type Nome struct { campo: Tipo }`
- `xmlDocument`, `xmlNode`: handle opachi usati dal modulo std XML

### Dichiarazioni
```sui
state count = 0
let name = "Alice"
type User struct { id: int, name: string }
view App { ... }
fn inc() { count++ }
async fn load() { ... }
```

### Controllo di flusso
```sui
if count > 0 { ... } else { ... }
loop i in 0..5 { ... }
while count < 10 { ... }
loop k, v in m { ... }  // iterazione mappa
break, continue
```

### Componenti & Views
```sui
view Card(title: string) {
  children  // slot
}

view App {
  Card title: "Demo" {
    text "Content"
  }
}
```

### Assegnazioni & Operatori
- Compound: `+=`, `-=`, `*=`, `/=`
- Post/pre: `x++`, `x--`
- Logici: `and`, `or`, `not`
- Confronto: `==`, `!=`, `<`, `<=`, `>`, `>=`
- Aritmetici: `+`, `-`, `*`, `/`, `%`

### Controlli form

`required: true` è disponibile su `input`, `textarea`, `select`, `combobox`,
`radio`, `number`, `date`, `time` e `datepicker`. I controlli nativi usano
l'attributo `required` del browser; il combobox mostra anche un errore inline
quando non è selezionato un valore.

### Tipi union e type guard `is`

Un tipo può elencare alternative con `|`; il suffisso `?` significa `| null`:

```sui
state value: string | int = "hello"
state maybe: string? = null
state mixed: array[string | int] = ["a", 1]
```

Assegnazioni, parametri, tipi di ritorno, campi di struct, elementi di array e
valori di mappe accettano una union quando il valore rientra in un membro. Gli
array e le mappe con valori misti compatibili inferiscono un tipo union per
elemento/valore.

```sui
fn show(v: string | int) -> string {
	if v is string {
		return "string: " + v
	}
	return "number: " + str(v)
}
```

`value is Tipo` è un'espressione booleana. Supporta `string`, `int`, `float`,
`bool`, `null`, `fn`, tipi array ed enum. I guard su struct e mappe sono
rifiutati perché a runtime sono entrambi oggetti JavaScript; usa un campo
discriminante enum, oppure `is null`/`is array[...]` per separare i casi. Nel
ramo corrispondente la variabile viene ristretta:

- `if v is string { ... }`: `v` è `string` nel ramo then e i membri
  rimanenti nel ramo else.
- `not (v is T)` inverte il ramo.
- `A and B` restringe entrambe nel ramo then; `A or B` restringe entrambe
  negate nel ramo else.
- `while v is string { ... }` restringe il corpo.
- Un guard che non può mai corrispondere è un errore di compilazione.
- Riassegnare uno state ristretto usa il tipo union dichiarato, quindi passare a
  un altro membro è permesso.

`is int` usa `Number.isInteger`; `is float` richiede un numero non intero.
La distinzione int/float è statica, mentre a runtime sono entrambi numeri
JavaScript. `map[string]int | null` è una mappa nullable; valori di mappa nullable non
sono ancora esprimibili (usa `map[string]any`). La restrizione di espressioni
member/index e l'esaustività per switch non letterali non sono ancora
disponibili.

## Literal types

Un literal type limita un valore a una stringa, un numero o un booleano esatto.
Spesso si scrive come union:

```sui
state mode: "idle" | "loading" | "error" = "idle"
state level: 1 | 2 | 3 = 1
state mixed: array["idle" | "loading"] = ["idle", "loading"]
```

- Assegnazioni, argomenti di funzione, valori di ritorno, campi di struct ed
  elementi di array sono confrontati con il valore letterale esatto.
  `mode = "run"` viene rifiutato.
- Un valore letterale è accettato ovunque sia atteso il suo tipo base, ma un
  valore `string`, `int` o `bool` generico non è accettato dove serve un literal
  type (il compilatore non può dimostrare il valore esatto).
- Gli array e le mappe letterali sono controllati elemento per elemento, anche
  con union di literal.
- I literal negativi sono supportati: `state delta: -1 | 1 = -1`.
- I literal numerici mantengono la distinzione int/float: `1` e `1.0` sono
  membri diversi.
- Operatori e builtin usano il tipo base, quindi `mode == "idle"`,
  `upper(mode)` e `"mode " + mode` funzionano normalmente.
- `value is "idle"` confronta con il literal a runtime. Un guard che non può
  mai corrispondere è un errore di compilazione.
- Uno `switch` su una union composta solo da literal deve coprire ogni membro,
  a meno che ci sia un blocco `default`. Il compilatore segnala i membri
  mancanti.

```sui
fn describe() -> null {
	switch mode {
	case "idle" { message = "ready" }
	case "loading" { message = "working" }
	case "error" { message = "failed" }
	}
}
```

Limiti: i valori di una mappa non possono ancora essere una union, quindi
`map[string]"ok" | "bad"` viene letto come `map[string]"ok"` oppure `"bad"`;
usa `map[string]string`. La restrizione di espressioni member/index e
l'esaustività per switch `string`/`int` ordinari non sono disponibili.

## Generici e parametri rest

Funzioni e componenti possono dichiarare parametri di tipo dopo il nome. Le
chiamate inferiscono gli argomenti di tipo dai valori ricevuti; non esiste una
sintassi esplicita `<int>` nel punto di chiamata.

```sui
fn identity<T>(value: T) -> T {
	return value
}

fn first<T>(items: array[T]) -> T {
	return items[0]
}

view List<T> (items: array[T]) {
	ul {
		list item in items key: item {
			li "{str(item)}"
		}
	}
}
```

- Un parametro di tipo può comparire nei parametri, nei valori di ritorno, nei
  tipi array/mappa e nei parametri dei componenti. Forme annidate come
  `map[string]array[T]` funzionano.
- Ogni parametro di tipo deve essere inferibile da un parametro. Una chiamata
  che non riesce a legarlo segnala `cannot infer generic type parameter`.
- Argomenti in conflitto (`both<string>("a", 1)`) segnalano il tipo sostituito
  sull'argomento problematico.
- Le struct generiche non sono ancora supportate. I parametri di tipo vengono
  cancellati nel JavaScript generato.
- Gli argomenti di tipo dei componenti sono inferiti dalle prop fornite, come
  per le chiamate a funzione.

Un **parametro rest** raccoglie gli argomenti rimanenti in un array:

```sui
fn sum(base: int, ...values: array[int]) -> int {
	let total = base
	loop v in values {
		total = total + v
	}
	return total
}

let x = sum(10, 1, 2, 3)   # 16
let y = sum(5)             # 5
```

Il parametro rest deve essere l'ultimo, deve avere un tipo array e non può avere
un valore predefinito. Il tipo dell'elemento viene controllato per ogni
argomento extra. Gli argomenti nominati non possono essere combinati con un
parametro rest.

## Costanti, valori computed e campi readonly

`const` dichiara un valore immutabile di primo livello:

```sui
const LIMIT: int = 3
const TAGS: array[string] = ["a", "b"]

state count = LIMIT
```

L'inizializzatore deve essere un valore compile-time: literal, const
precedenti, membri enum, array/mappe/struct costruiti da questi e operatori o
condizionali su di essi. State runtime e chiamate a funzione sono rifiutati,
quindi le costanti non dipendono dall'ordine di esecuzione. Assegnare a una
costante o mutare uno dei suoi contenitori è un errore.

`computed` dichiara un valore in sola lettura derivato dagli state (e da
computed precedenti):

```sui
state price: int = 10
state qty: int = 2

computed total: int = price * qty
computed label = "total " + str(total)
```

I valori computed vengono ricalcolati dopo ogni cambio di state e prima di ogni
render. Non sono assegnabili né mutabili. Dichiarali nell'ordine delle
dipendenze; ogni computed può usare state, costanti, computed precedenti e
chiamate a funzioni pure. Mantieni l'espressione priva di effetti collaterali.

I campi di una struct possono essere marcati `readonly`. La costruzione resta
invariata, ma un'assegnazione successiva attraverso quel campo viene rifiutata:

```sui
type Job struct {
	readonly id: string
	mode: string
}

state job: Job = Job { id: "a", mode: "idle" }
# job.id = "b"   -> errore di compilazione
```

## Aggiornamenti reattivi, batching, error boundary e chiavi delle liste

I setter degli state aggiornano subito i valori computed, poi raggruppano il
lavoro DOM in una microtask. Più assegnazioni nello stesso handler causano
quindi un solo render. Vengono ri-renderizzate solo le view legate allo state
modificato (direttamente o tramite un computed), così i controlli non correlati
mantengono il loro stato DOM nativo.

`errorboundary` intercetta gli errori sincroni sollevati durante il render del
suo sottoalbero e mostra un fallback:

```sui
errorboundary fallback: "Questa sezione non può essere renderizzata" {
	p "{maybeBroken()}"
}
p "Il resto della pagina continua a funzionare."
```

Il messaggio d'errore viene salvato in `data-sui-error`, il testo del fallback
viene escapato e l'errore viene registrato anche con `console.error`. Gli errori
sollevati più tardi da event handler o lavoro asincrono non vengono intercettati
dall'error boundary. `assert(cond, "messaggio")` solleva un `Error` quando
`cond` è falso; è utile per validare invarianti e per provare un error boundary.

Una lista in una view può dichiarare una chiave stabile:

```sui
ul {
	list item in items key: item.id {
		li "{item.name}"
	}
}
```

La chiave viene applicata al primo elemento renderizzato per ogni item come
`data-sui-key`; il ripristino del focus segue la chiave quando la lista viene
riordinata. Le chiavi devono essere stringhe, numeri, booleani o valori enum.
Non cambiano i dati né la struttura del markup generato. Nel designer i blocchi dinamici `if`/`list` vengono conservati come sorgente raw e si modificano dalla scheda SUI.

## Destructuring, spread e argomenti nominati

```sui
type Person struct { name: string, age: int }

let person = Person{ name: "Anna", age: 30 }
let {name, age} = person          # campi dello struct
let [first, second] = [10, 20]   # posizioni dell'array
let {x, y} = {"x": 7, "y": 8}    # valori della mappa

let numbers = [...[1, 2], ...[3, 4]]
let defaults = {"theme": "dark", "lang": "it"}
let config = {...defaults, "lang": "en"}

fn greet(name: string, excited: bool = false) -> string { ... }
greet(excited: true, name: "Anna")
```

- I pattern a oggetto leggono campi di uno struct o valori di una mappa; i
  pattern ad array leggono le posizioni. Ogni nome diventa un binding `let`
  locale.
- Gli array e le mappe accettano lo spread `...valore`. Uno spread di mappa
  richiede una mappa; una mappa mantiene un solo tipo comune per i valori.
- Gli argomenti nominati possono essere in qualsiasi ordine. Argomenti
  posizionali e nominati non si possono mescolare nella stessa chiamata; nomi
  sconosciuti o duplicati sono errori. I parametri opzionali finali omessi usano
  il default; un parametro opzionale intermedio viene riempito col default se
  viene fornito un argomento nominato successivo.
- Non ancora supportati: rinomina/default/buchi nei pattern, spread nei struct
  literal o negli argomenti delle chiamate, argomenti nominati per builtin e
  funzioni std. `simpleui fmt` preserva queste forme.

## Enum e switch

Gli enum sono tipi nominali:

```sui
enum Step { Idle, Loading, Done, Failed }

state step: Step = Step.Idle
state steps: array[Step] = [Step.Idle, Step.Loading]

fn label(s: Step) -> string {
	switch s {
		case Step.Idle { return "Ready" }
		case Step.Loading { return "Loading..." }
		case Step.Done { return "Done" }
		case Step.Failed { return "Failed" }
	}
	return "Unknown"
}
```

- I membri si scrivono `Step.Idle` e hanno il tipo dell'enum. Una stringa non è
  assegnabile a `Step`; assegnazioni e confronti sono controllati
  nominalmente.
- `switch` è un'istruzione dentro le funzioni. Valuta il valore una volta, non
  ha fallthrough e ogni corpo `case` termina automaticamente. `default` è
  opzionale. `break` è ammesso dentro uno switch; `continue` richiede ancora un
  ciclo.
- I case possono essere membri enum, stringhe, numeri o booleani. Valori
  duplicati sono errori; un membro enum sconosciuto mostra un hint con i membri
  disponibili.
- Uno switch su enum senza `default` deve coprire tutti i membri. Il compilatore
  elenca i mancanti, quindi aggiungere un membro evidenzia ogni switch da
  aggiornare. Usa `default` solo quando ignorare i nuovi membri è intenzionale.
- A runtime un enum è un oggetto congelato di stringhe, quindi `str(Step.Idle)`
  vale `"Idle"` e il JSON produce il nome del membro.
- Non ancora: literal types, payload/valori associati agli enum e
  destructuring nei pattern dei `case`.

## Drag and drop

Il runtime supporta il riordino di elementi e il drop di file o directory dal
sistema operativo. Gli handler non ricevono argomenti: usare i builtin
`drag.*` per leggere il payload. Per l'inspector, la differenza tra DnD del
designer e dell'applicazione e i limiti di liste/tabelle, vedere
[Editor GUI: drag and drop](GUI_EDITOR_IT.md).

```sui
state items = ["Design", "Build", "Ship"]
state attachment = ""

fn reorder() -> null {
	let src = indexOf(items, drag.source())
	let dst = indexOf(items, drag.target())
	if src < 0 or dst < 0 or src == dst {
		return
	}
	items = moveItem(items, src, dst)
}

view main {
	list item in items {
		row draggable: true dragKey: item dropKey: item onDrop: reorder {
			text "{item}"
		}
	}
	input "Allegato" value: attachment dropFiles: attachment
}
```

### Attributi

| Attributo | Significato |
|---|---|
| `draggable: true` | Rende l'elemento una sorgente drag. `dragKey` da solo lo rende già trascinabile. |
| `dragKey: expr` | Chiave della sorgente, letta con `drag.source()`. |
| `dropKey: expr` | Chiave del target, letta con `drag.target()`. |
| `dropFiles: state` | Target automatico per file/directory. Lo **state stringa** riceve i percorsi; un `input`/`textarea` con `value:` li mostra. |
| `pickFiles: state` | Apre il **dialogo file nativo**. Lo state stringa riceve i percorsi scelti. Su un controllo `file` usa l'input nativo; su `button`/contenitore/campo apre un input nascosto. |
| `pickMode: "native" \| "fs" \| "auto"` | Implementazione del picker. `native` usa input file/`webkitdirectory`; `fs` usa il folder picker della File System Access API (`showDirectoryPicker`); `auto` usa `fs` se disponibile, altrimenti native. Per la normale selezione di file usare `native`. |
| `directory: true` | Sul controllo `file` abilita il selettore di directory; su un trigger `pickFiles` usa un selettore nascosto (`webkitdirectory`). |
| `onDragStart`, `onDragEnter`, `onDragOver`, `onDragLeave`, `onDrop`, `onDragEnd` | Handler evento; usare `drag.*`. |

`draggable`, `dragKey`, `dropKey`, `dropFiles` funzionano su contenitori,
righe, elementi di lista e controlli. `dropFiles` non e limitato agli input:
puo riempire uno state usato da qualsiasi elemento visibile. Impostare sempre
`dragKey` sulla sorgente: un semplice `draggable: true` non invia una chiave
utile a `onDrop`.

### File e directory

- Il browser espone **nomi e percorsi relativi**, non percorsi assoluti. Un file
  singolo di solito mostra solo il nome; una cartella mostra percorsi come
  `foto/2026/a.jpg`. Host desktop che espongono `File.path` possono mostrare il
  percorso assoluto, ma non è garantito.
- La scansione ricorsiva delle directory è asincrona: `dropFiles` aggiorna lo
  state al termine. `drag.files()` contiene i metadati della selezione corrente/ultima. La
  ricorsione usa `webkitGetAsEntry` (Chromium/Edge); browser senza questa API
  ripiegano sui file esposti dal browser.
- Limiti: 2.000 file, profondità 16, 4.096 caratteri per percorso. Input non
  valido produce un messaggio visibile `Cannot accept the files: ...`.
- Su textarea i percorsi sono separati da newline; su input da `, `.
- Il runtime legge solo metadati, mai il contenuto dei file.

I nomi dei file sono dati non fidati: visualizzarli tramite state (escaping
automatico) o `textContent`; non usarli come HTML, CSS o URL non validati. Il
drop non concede accesso al filesystem: su desktop `fs` richiede sempre
`--allow-dir`.

### Picker cartella con File System Access API

Su browser Chromium e contesti sicuri (localhost/HTTPS), `pickMode: "fs"` usa
`showDirectoryPicker()`:

```sui
button "Scegli cartella..." pickFiles: attachment directory: true pickMode: "fs"
```

`pickMode: "auto"` usa la File System Access API quando disponibile e ripiega
sul picker nativo `webkitdirectory` altrove. A differenza del picker nativo, la
FS API restituisce un handle di directory: una cartella vuota mantiene quindi il
nome. Il runtime la percorre in sola lettura, produce gli stessi percorsi/metadati
e imposta esplicitamente `drag.folder()`.

Vincoli:

- permesso concesso dall'utente tramite gesto nel browser;
- solo Chromium/Edge nelle versioni attuali (non Firefox/Safari);
- richiede un contesto sicuro;
- sola lettura: nessuna scrittura/cancellazione esposta;
- il riuso dell'handle dopo un reload non è implementato;
- `pickMode: "fs"` mostra un errore esplicito se l'API non c'è, `auto` ripiega
  sul picker nativo.

### Selezione cartella

`pickFiles: state directory: true` apre il selettore di cartelle del browser.
Il browser restituisce ricorsivamente i file contenuti; `drag.folder()` restituisce
il nome della cartella top-level comune, se i percorsi condividono lo stesso
primo segmento.

Il selettore nativo sceglie una **intera cartella**, non un sottoinsieme
arbitrario di file al suo interno. Per analizzare solo alcuni file usare il
selettore multiplo (`multiple: true` senza `directory`) oppure filtrare la
selezione nel codice dell'app. Una cartella vuota scelta dal dialogo nativo non
produce file ne nome; usare `pickMode: "fs"` quando serve il nome di una
cartella vuota. Una cartella vuota trascinata puo produrre una entry sintetica in
Chromium (`drag.files()` con `isDirectory: true`), ma non ha byte da caricare.

L'accesso cross-browser alle cartelle resta volutamente limitato a questo
modello basato sui file. `pickMode: "fs"` espone un handle di directory in sola
lettura, ma il runtime seleziona comunque tutti i file. Una UI ad albero con
checkbox per file andrebbe costruita sopra l'handle e non fa parte del runtime
attuale.

### Finestra di dialogo file

```sui
state attachment = ""

view main {
	button "Scegli file..." pickFiles: attachment accept: ".csv,.txt" multiple: true
	input "Selezionati" value: attachment
}
```

Oppure direttamente su un controllo `file`:

```sui
file accept: ".csv,.txt" multiple: true pickFiles: attachment
input "Selezionati" value: attachment
```

- `pickFiles` funziona su `button`, `file`, contenitori e campi. Su elementi
  diversi da `file` il runtime crea un input file nascosto e lo apre al click.
- `accept`, `multiple` e `directory` configurano il dialogo.
- I percorsi seguono le stesse regole e gli stessi limiti del drop.
  `drag.files()` espone la selezione; `drag.error()` l'eventuale errore.
- Un campo di testo con `pickFiles` apre il dialogo al click: usarlo solo per
  campi percorso di sola lettura; preferire un `button` separato.
- Il browser espone nomi e percorsi relativi, non percorsi assoluti. Il controllo
  `file` nativo mostra i nomi selezionati, ma un re-render guidato dallo state
  puo azzerare quella visualizzazione: usare un campo collegato
  (`value: state`) come vista stabile della selezione.

## Menu bar

`menubar`, `menu` e `menuitem` costruiscono una barra di navigazione orizzontale
con menu a discesa cliccabili. Etichette, link e handler sono normali valori
SUI: non esiste HTML raw.

```sui
menubar "Main menu" {
	menu "File" {
		menuitem "New" onClick: newFile
		menuitem "Open" onClick: openFile
		divider
		menuitem "Project site" link: "https://example.com"
	}
	menu "Edit" {
		menuitem "Undo" onClick: undo
		menuitem "Redo" onClick: redo
		menu "More" {
			menuitem "Preferences" onClick: preferences
		}
	}
}
```

| Elemento | Testo | Attributi utili | Figli |
|---|---|---|---|
| `menubar "Nome accessibile"` | Nome accessibile (`aria-label`) | `class`, `style` | `menu` |
| `menu "Etichetta"` | Pulsante visibile del menu/sottomenu | `class`, `style` | `menuitem`, `menu`, `divider` |
| `menuitem "Etichetta"` | Etichetta visibile della voce | `class`, `style`, `link`, `onClick` | no |

La barra è un vero `nav` con `role="menubar"`. Un menu è un vero pulsante con
popup nascosto: il click apre/chiude, il click fuori o Escape chiude, le frecce
spostano il focus tra le voci. Un `menuitem` con `link` viene avvolto in un link
sicuro (`javascript:` e altri schemi attivi sono rifiutati dal normale URL
guard). Un `menuitem` con `onClick` chiama la funzione senza argomenti indicata.

Un `menu` annidato nel popup di un altro `menu` è un sottomenu; i sottomenu
possono annidarsi ulteriormente. Usare `divider` per un separatore dentro un
menu. Etichette e link sono escapati; trattare le etichette dinamiche come testo
non fidato.

Nel designer aprire **Layout → Navigation** e inserire **Menu bar**. Parte con
le voci File/Edit/Help. Fare clic con il tasto destro su una riga dell'albero
per l'editor dei menu: **Add menu**, **Add menu item**, **Add submenu**,
**Duplicate**, **Move up/down**, **Delete**. È anche possibile trascinare
**Menu** o **Menu item** nella barra/menu e modificare **Text**, **link** e
**onClick** nell'inspector. Vedere [Editor GUI: menu bar](GUI_EDITOR_IT.md#menu-bar-control).

## Dropdown menu

`dropdown` è un pulsante con un menu ancorato. Il testo è l'etichetta del
pulsante e i figli diretti `dropdownitem` diventano le voci:

```sui
dropdown "Azioni" position: "right" {
	dropdownitem "Modifica" onClick: edit
	dropdownitem "Duplica" onClick: duplicate
	dropdownitem separator: true
	dropdownitem "Elimina" onClick: delete
}
```

`position` è `left` (default) o `right`. Le voci supportano `onClick`,
`disabled` e `separator`; il menu si chiude con un click fuori, Esc o la
selezione di una voce. Il runtime riusa il comportamento tastiera/focus della
menu bar e riposiziona il popup dentro la finestra: si sposta a sinistra/destra
ai bordi, si apre sopra il trigger quando sotto non c'è spazio e diventa
scorrevole se è più alto della finestra. Nel designer scegliere **Controls >
Dropdown menu**; modificare l'etichetta da **Text**, il lato dalla sezione
**Dropdown** e aggiungere le voci dall'albero.

### Posizionamento e animazione degli overlay

Gli overlay ancorati condividono lo stesso motore di posizionamento. Il lato
preferito viene invertito quando non c'è spazio e le coordinate vengono
limitate alla viewport.

| Attributo | Valori | Si applica a |
|---|---|---|
| `placement` | `auto` (default), `bottom`, `top`, `left`, `right` | dropdown, combobox |
| `position` | `auto` (default), `top`, `right`, `bottom`, `left` | popover |
| `animation` | `fade` (default), `scale`, `slide`, `none` | dropdown, popover, combobox, commandbar |
| `duration` | tempo CSS o millisecondi, es. `180ms` | overlay, drawer, carousel, skeleton |
| `easing` | funzione di timing CSS, es. `ease-out` | overlay, drawer, carousel |
| `delay` | tempo CSS o millisecondi | dropdown, popover, combobox, commandbar |

L'effetto `slide` segue il posizionamento finale (down, up, left o right).
`drawer` e `carousel` usano `duration`/`easing` per le transizioni; `skeleton`
usa `duration` per la velocità dello shimmer. I tooltip usano le impostazioni
globali del progetto per posizione, effetto, durata, easing e ritardo.

### Motion tra stati e preset dei toggle

`checkbox`, `switch` e `radio` possono morfare tra due strutture e due temi
invece di saltare. I preset di forma e tema impostano variabili CSS per ogni
parte; il motore di motion cattura i vecchi valori e interpola colori, lunghezze,
raggi e posizioni frame per frame:

```sui
state on = true
state shape = "pill"
state theme = "accent"

switch "Modalità" shape: shape theme: theme checked: on
	animation: "morph" duration: "360ms" easing: "ease-in-out"
	stagger: "0.12" onChange: toggle
```

| Attributo | Valori | Significato |
|---|---|---|
| `shape` | switch: `default`, `rounded`, `square`, `pill`, `ios`, `material`, `minimal`, `soft`; checkbox: `default`, `rounded`, `square`, `circle`, `ios`, `minimal`, `soft`; radio: `default`, `compact`, `large`, `soft`, `minimal` | Preset di struttura. |
| `theme` | `default`/`accent`, `success`, `warning`, `danger`, `mono`, `dark`, `light` | Preset di colore. |
| `animation` | `auto`, `morph`, `fade`, `scale`, `slide`, `none` | Effetto di transizione. |
| `duration` | tempo CSS o ms | Durata del tween. |
| `easing` | funzione di timing CSS | Easing, incluso `cubic-bezier(...)`. |
| `delay` | tempo CSS o ms | Ritardo prima del tween. |
| `stagger` | `0`-`0.5` | Frazione della durata distribuita sulle parti. |

`morph` interpola solo forma e colore; `fade` scende al 35% di opacità e
ritorna; `scale` si riduce all'82% senza dissolvenza; `slide` si sposta di 10px
e scende al 40%. Tutti gli effetti restano leggibili e non fanno sparire il
controllo.

Lo stesso meccanismo vale per qualsiasi elemento che opta in con gli attributi
data del motion: il runtime cattura proprietà CSS e custom properties scelte,
poi genera i frame intermedi. Per i toggle nativi lo stato precedente viene
catturato su `pointerdown`/`keydown`, prima che il browser cambi `checked`.

### Motion di layout con FLIP e chiavi

Qualsiasi elemento può usare il motion di layout con chiave:

```sui
card motion: "layout" motionKey: project.id motionDuration: "420ms" {
	h3 "{project.name}"
}
```

| Attributo | Significato |
|---|---|
| `motion` | `layout` (solo FLIP), `fade`, `scale`, `slide` o `none`. |
| `motionKey` | Identità stabile tra i render. Obbligatoria per elementi di liste e componenti. |
| `motionDuration` | Durata della transizione, es. `420ms`. |
| `motionEasing` | Funzione di timing CSS. |
| `motionDelay` | Ritardo prima della transizione. |
| `motionExit` | `false` disabilita il clone di uscita per le liste grandi. |

Prima del render il runtime salva rettangolo e stili di ogni elemento con chiave.
Dopo il render abbina per `motionKey` e trasla l'elemento dalla vecchia posizione
alla nuova. Le chiavi scomparse vengono clonate in un overlay fisso, dissolte e
rimosse; le nuove chiavi entrano in dissolvenza. `layout` mantiene l'elemento
opaco e lo sposta soltanto; `fade`, `scale` e `slide` aggiungono l'effetto di
entrata/uscita corrispondente.

`data-sui-id` non è univoco dentro un nodo `list`, quindi gli elementi ripetuti
devono impostare `motionKey` (ad esempio `motionKey: item.id`). Card, pannelli e
rami condizionali possono condividere la stessa chiave tra percorsi di render
diversi, abilitando transizioni lista → dettaglio e liste filtrate.

In un'app multi-pagina, `navigate(page)` cattura prima la pagina visibile,
renderizza la nuova pagina e poi applica le chiavi corrispondenti. Lo stesso
`motionKey` su due pagine anima quindi l'elemento tra le pagine, inclusi
posizione e dimensioni. Il runtime misura i rettangoli iniziale e finale della
card e delle sue parti visive dirette (titolo, paragrafo, pulsante), poi
interpola questi stati misurati con un'unica funzione di easing. Un livello di
superficie senza contenuto muove e ridimensiona sfondo, bordo e ombra, mentre un
livello di contenuto ritagliato mantiene il layout finale; ogni parte viaggia in
linea retta verso il proprio rettangolo finale e il testo non viene mai
riflussato o riscritto durante la transizione. L'elemento reale viene mostrato
alla fine esattamente alla stessa geometria, quindi non ci sono cross-fade,
salti o testo che scompare. Il contenuto senza chiave corrispondente appare
semplicemente con la nuova pagina.

### Motion graph (stati e trigger)

`motionGraph` trasforma un elemento in una piccola macchina a stati. L'attributo
è un oggetto JSON letterale generato dal designer:

```sui
card motionGraph: "\{\"initial\":\"closed\",\"states\":[\{\"id\":\"closed\",\"name\":\"Chiuso\",\"css\":\"width: 300px;\"\},\{\"id\":\"open\",\"name\":\"Aperto\",\"css\":\"width: 420px; margin-left: 80px;\"\}],\"links\":[\{\"id\":\"l1\",\"from\":\"closed\",\"to\":\"open\",\"event\":\"click\",\"duration\":\"320ms\",\"easing\":\"ease-out\"\}]\}" { ... }
```

- `states[].css` è la fotografia dello stile per quello stato.
- `links[].event` può essere `click`, `dblclick`, `mouseover`, `mouseout`,
  `mouseenter`, `mouseleave`, `pointerenter`, `pointerleave`, `focus`, `blur`,
  `change`, `input`, `keydown`, `keyup`, `pointerdown`, `pointerup`, `submit`,
  `wheel`, `swipe`, `longpress`, `visible`, `hidden`, `custom`, `state`,
  `elementState` o `after`.
- `links[].sourceId` collega un evento DOM a un altro elemento (il suo
  `data-sui-id` o `data-sui-motion-key`); `whenState` seleziona lo stato del
  motion graph di quell'elemento per `elementState`. `sourceLabel` è l'etichetta
  del designer e non viene usata a runtime. Imposta `motionKey` sull'elemento
  sorgente per un riferimento stabile e leggibile; le chiavi usano lettere,
  cifre, `_`, `.`, `:` e `-`.
- Gli eventi `custom` usano `customEvent` come nome DOM. I link DOM possono avere
  filtri: `key` per gli eventi tastiera, `ctrl`/`alt`/`shift`/`meta`, `button`
  per il puntatore e `wheel` (`any`, `down`, `up`) per la rotella. Se più link
  corrispondono allo stesso stato/evento vince quello con i filtri più specifici.
- `links[].guard` è una condizione opzionale valutata prima di eseguire il
  trigger. È un piccolo linguaggio sicuro, non JavaScript: nomi di stati SUI
  (booleani, numeri, stringhe), `true`/`false`/`null`, `!`, `&&`, `||`, confronti
  (`==`, `===`, `!=`, `!==`, `<`, `<=`, `>`, `>=`) e parentesi. Nessun `eval` e
  nessuna chiamata a funzioni. I nomi sconosciuti valgono `undefined`; una guard
  non valida non scatta mai. Le guard filtrano eventi DOM, trigger di stato ed
  elemento esterno, timer, link `after` e `run`. Le chiamate esplicite a
  `SimpleUIMotion.to()` sono programmatiche e non vengono bloccate da una guard.
- `simpleui check` valida il JSON letterale di `motionGraph`: sintassi JSON, id
  degli stati, riferimenti a `initial` e `exit`, nomi dei trigger e sintassi delle
  guard. Le espressioni dinamiche di `motionGraph` non vengono valutate in fase
  di compilazione.
- Link gesto: `swipe` usa `axis` (`left`, `right`, `up`, `down`) e `distance`
  (pixel minimi, default 40); `longpress` usa `hold` (millisecondi, default 500,
  annullato dopo circa 10 px di movimento). Entrambi usano i pointer events e
  sono ripetibili, quindi possono scattare di nuovo dopo il ritorno allo stato
  sorgente. Per l'accesso da tastiera, un grafo con link gesto rende l'elemento
  focalizzabile (`tabindex="0"` quando serve) e mappa le frecce su swipe,
  PageUp/PageDown su wheel e Spazio/Invio su long press; l'elemento riceve anche
  `aria-keyshortcuts` e un outline di focus visibile. I controlli di form non
  vengono intercettati.
- Link di visibilità: `visible` scatta quando l'elemento entra nel viewport e
  `hidden` quando esce. `threshold` è la frazione visibile tra `0` e `1`
  (default `0.5` per `visible`, `0.2` per `hidden`); `once: true` scatta solo la
  prima volta. Usano `IntersectionObserver` con un controllo iniziale immediato,
  così le animazioni di ingresso partono appena l'elemento viene montato. Senza
  `IntersectionObserver`, il runtime misura anche al timer iniziale, allo scroll
  del documento (inclusi gli scroll annidati), al resize della finestra e con un
  controllo periodico, quindi funzionano entrambe le direzioni.
- `links[].iterations` è un intero positivo o `"infinite"`; `direction` è
  `normal`, `reverse`, `alternate` o `alternate-reverse`; `playbackRate` è tra
  `0.1` e `10`. `spring: {stiffness, damping, mass}` sostituisce `easing` e usa
  la durata del link come velocità. I link infiniti non terminano, non attivano
  link `after` e vengono annullati dalla transizione successiva partendo dai
  valori correnti. I loop finiti `alternate` si fermano sullo stato di
  destinazione.
- `links[].emit` è un nome di evento oppure
  `{ "name": "...", "on": "start" | "end" | "cancel" }`. Il runtime emette anche
  `sui:motionstate`, `sui:motionstart`, `sui:motionend` e `sui:motioncancel`
  sull'elemento con `{from,to,state,link}`. Gli eventi fanno bubbling e sono
  composed, quindi un altro grafo può ascoltarli con un trigger `custom` e un
  `sourceId`.
- `window.SimpleUIMotion.get(elementoOSelettore)` restituisce l'API
  dell'elemento: `to(state, opts)`, `pause()`, `resume()`, `seek(ms)`,
  `progress`, `state` e `paused`. `to` accetta gli stessi campi di timing di un
  link. L'API è disponibile anche come `element.__suiMotion`.
- Un cambio di stato parte subito: durante il tween l'elemento è considerato
  nello stato di destinazione. Un evento che arriva durante il tween viene
  valutato nel nuovo stato e avvia un nuovo tween dai valori correnti (vince
  l'ultimo evento); gli eventi non vengono accodati. Una nuova transizione annulla
  i link `after` e i timer dello stato precedente e tutte le proprietà gestite dal
  grafo sono tracciate, quindi un tween interrotto non lascia valori intermedi.
  I timer partono quando lo stato di destinazione viene attivato (all'inizio
  della transizione); i link `after` scattano al termine.
- I link `state` valutano lo stato booleano indicato in `when` (il prefisso `!`
  lo nega) e scattano sulla transizione da falso a vero.
- I link `after` continuano una catena quando l'animazione precedente termina.
  Partono dalla callback di fine, quindi non gareggiano con i valori finali.
- Le transizioni usano il motore di animazione del browser con valori iniziali e
  finali calcolati: transform, filtri, box shadow, clip path, gradienti, colori,
  lunghezze e custom properties vengono interpolati dove supportato; i valori non
  animabili cambiano in modo discreto. I valori finali sono applicati alla fine e
  `prefers-reduced-motion` o durata zero saltano subito, loop inclusi.

Il pulsante **Motion > Open animation panel** del designer modifica lo stesso
JSON: i blocchi sono gli stati, le frecce sono i link. Il pannello è compatto e
ridimensionabile dall'handle in basso a destra; il canvas centra e scala il
diagramma. I parametri si aprono in overlay trascinabili ancorati a una
posizione fissa. Pannelli e overlay restano dentro il viewport visibile.
**Layout & shared** nella toolbar apre le impostazioni classiche `motion`,
`motionKey`, `motionDuration`, `motionEasing`, `motionDelay` e `motionExit`
usate da FLIP, riordino liste e transizioni tra pagine.
**Exit animation** nella toolbar modifica `motionGraph.exit`:

```json
{"state":"gone","duration":"240ms","easing":"ease-out","delay":"0ms"}
```

Quando SUI rimuove un elemento gestito dal grafo, il runtime ne clona lo stato
attuale nella posizione corrente, rimuove identità e attributi motion e anima il
clone verso i valori CSS di `exit.state` (anche con `spring`). Il clone cattura i
valori calcolati correnti, conserva il transform 2D e il box visivo e risolve le
percentuali di `width`/`height`/`min-*`/`max-*` rispetto al parent originale. Il
clone è inerte, ignora i pointer events e viene rimosso al termine. Gli elementi
che sopravvivono al render non producono un clone di uscita. Con
`prefers-reduced-motion` l'elemento viene rimosso senza clone.

I blocchi crescono in altezza con i trigger uscenti e possono essere trascinati
in posizioni personalizzate. Il canvas supporta zoom con la rotella, pan
trascinando lo spazio vuoto e i controlli meno/più/Fit con la percentuale di
zoom. Ogni blocco ha **+** per aggiungere un link, l'icona a matita per
modificarne l'aspetto e **×** per eliminarlo. **Create
state** nel pannello crea un nuovo blocco senza link; muovi il mouse e clicca su
uno spazio libero, poi usa **Confirm position**. Il nuovo stato è selezionato. Usa
**+** su uno stato per creare un link verso una destinazione esistente: il
chooser si apre con **Cancel** e **Confirm** e preseleziona lo stato creato più di
recente. Confirm crea il link, chiude il pannello e lo disegna con il colore di
completamento. Per collegare all'indietro, premi **+** sul nuovo stato e scegli il
vecchio. Un link vecchio senza trigger può ancora essere completato con doppio
clic sullo stato sorgente o sulla sua riga rossa. Il trigger appartiene allo stato
di partenza; i link uscenti sono numerati all'inizio e il tooltip del numero
mostra il tipo di trigger. I link partono dal lato destro e arrivano
orizzontalmente al centro del lato sinistro della destinazione. Il percorso è
ortogonale con una distanza minima di **15 px** tra box, tra link e tra link e
box; il primo e l'ultimo tratto sono perpendicolari al lato e lunghi almeno
15 px. I punti di connessione sono assegnati in ordine verticale di destinazione
e tutti i link vengono ricalcolati a ogni modifica, provando più ordini; il router
evita i box, usa angoli retti e preferisce il percorso più breve con il minor
numero di segmenti. I link all'indietro girano attorno ai box e i
percorsi si aggiornano in tempo reale durante il trascinamento. Se non esiste un
layout senza incroci, i link possono incrociarsi: il link successivo viene
disegnato sopra il precedente con un piccolo distacco sullo sfondo. Quando un
piccolo spostamento automatico di uno stato può ripristinare un percorso pulito,
l'editor muove quello stato e ricalcola tutti i link; le posizioni sono fissate
per mantenere il layout stabile. Blocchi e link sono sensibili al tema
e restano leggibili in modalità chiara e scura.

Il doppio clic su uno stato apre l'overlay con nome, CSS e lista dei trigger
attivi con la destinazione (nessun campo per i trigger in ingresso). Se lo stato
ha un collegamento uscente senza trigger, si apre direttamente il chooser:
**Cancel** torna all'overlay senza modificare il link, **Confirm** applica il
trigger, chiude l'overlay, ripristina il grafo e disegna il link con il colore
normale di completamento. **Add trigger** riusa lo stesso chooser per un
nuovo link. Il doppio clic su una freccia apre trigger, durata, easing e delay;
una freccia senza trigger apre il chooser per quel link. Il chooser supporta
eventi DOM (una volta per stato sorgente), trigger variabile con valore
attivatore e timer con durata.

Il clic semplice su uno stato lo seleziona e mostra il suo CSS nella preview
senza lasciare il pannello. L'icona a matita sul blocco stato entra in editing
grafico: il pannello animazioni viene nascosto, l'elemento è isolato e resta solo
la barra con **Save state**, **Cancel** e **Close**. Save cattura lo stile della
preview, lo associa allo stato e riapre il pannello; Cancel scarta le modifiche e
riapre il pannello; Close le scarta, chiude il pannello animazioni e torna
all'editor normale. Il salvataggio rifiuta trigger vuoto, variabile mancante,
parametri uguali o evento singolo ripetuto e mostra l'errore. **Remove states**
elimina il grafo degli stati ma conserva Layout & shared.

`motionGraph` è per singolo elemento. Usa `motionKey` per le transizioni
condivise tra pagine e per il FLIP delle liste; i due sistemi possono convivere
sullo stesso elemento.

## Popover

`popover` è un trigger con contenuto ricco in un pannello ancorato. Il testo è
l'etichetta del trigger; i figli diventano il contenuto:

```sui
popover "Info" position: "bottom" {
	h3 "Dettagli"
	p "Contenuto ancorato."
	button "Azione"
}
```

`position` è `auto` (default), `top`, `right`, `bottom` o `left`. Auto sceglie
il lato con spazio sufficiente e limita il pannello nella viewport. Si apre al
click e si chiude con un click fuori o Esc. Aggiungere `animation`, `duration`,
`easing` e `delay` per l'effetto di apertura. Nel designer scegliere **Controls
> Popover**; la sezione **Popover** imposta lato e animazione.

## Drawer

`drawer` è un pannello laterale controllato da un binding `open`. Il testo è il
titolo e i figli diventano il corpo scorrevole:

```sui
state showFilters = false

fn openDrawer() -> null {
	showFilters = true
}

drawer "Filtri" side: "right" width: "320px" open: showFilters {
	h3 "Filtri"
	p "Scegli uno stato."
}
```

`side` è `right` (default) o `left`. Quando `open` nomina uno stato, lo sfondo,
il pulsante di chiusura ed Esc chiamano il setter con `false`. Altrimenti usare
`onClose: <fn>`. Il pannello è fissato alla finestra con uno sfondo scurito e un
`aside` con `role=dialog`. `duration` ed `easing` controllano la transizione.
Nel designer scegliere **Controls > Drawer**; la sezione **Drawer** espone lato,
larghezza, open, onClose, duration ed easing.

## Split panes

`splitter` dispone esattamente due figli `pane` con un divisore trascinabile:

```sui
splitter direction: "horizontal" initial: "40%" min: "120px" {
	pane {
		p "Lista"
	}
	pane {
		p "Dettaglio"
	}
}
```

`direction` è `horizontal` (default) o `vertical`; `initial` è la dimensione del
primo pannello e `min` quella minima di entrambi. Trascinare la barra ridimensiona
i pannelli; la barra è focalizzabile e le frecce, `Home` e `End` la ridimensionano
senza puntatore. `simpleui check` rifiuta gli splitter senza esattamente due figli
`pane`. Nel designer scegliere **Controls > Split panes**.

## File dropzone

`dropzone` è un'area di drop stilizzata che riusa il runtime drag-and-drop e il
picker:

```sui
state files = ""

dropzone "File CSV" dropFiles: files accept: ".csv,text/csv" multiple: true {
	p "Solo file CSV."
}
```

| Attributo | Significato |
|---|---|
| `dropFiles` | Stato stringa che riceve i percorsi trascinati. Apre anche il picker al click. |
| `pickFiles` | Stato separato opzionale per il picker; default `dropFiles`. |
| `accept` | Filtro del picker, ad esempio `.csv` o `image/*`. |
| `multiple` | Consente più file. |
| `directory` | Abilita il picker cartelle e il drop di cartelle. |
| `pickMode` | `native`, `fs` o `auto`. |
| `hint` | Riga secondaria sotto il titolo. |

Il titolo è il testo dell'elemento, i figli vengono renderizzati dentro l'area e
una riga di stato mostra il valore dello stato finché non è vuoto. Nomi e
percorsi sono escapati come testo dal runtime. Nel designer scegliere
**Controls > File dropzone**.

## Timeline

`timeline` impila figli `timelineitem`. Ogni voce mostra un titolo (il suo
testo), una `date` opzionale, un punto di stato e contenuto ricco opzionale:

```sui
timeline {
	timelineitem "Progetto creato" date: "10 nov, 09:12" status: "done"
	timelineitem "Build completata" date: "10 nov, 09:30" status: "active" {
		p "Gli artefatti sono pronti."
	}
}
```

`status` è `pending` (default), `active`, `done` o `error`. Vengono renderizzati
solo i figli diretti `timelineitem`. Nel designer scegliere **Content >
Timeline**.

## Empty state

`emptystate` è un segnaposto centrato con icona, titolo, suggerimento e azioni
figlie opzionali:

```sui
emptystate "Nessun progetto" icon: "▢" hint: "Crea il primo progetto." {
	button "Nuovo progetto" onClick: newProject
}
```

`icon` e `hint` sono opzionali; `compact: true` riduce il padding. Nel designer
scegliere **Content > Empty state**.

## Skeleton

`skeleton` mostra righe segnaposto animate o un cerchio durante il caricamento:

```sui
skeleton lines: 3 width: "80%" height: "12px" gap: "8px"
skeleton circle: true height: "48px"
```

| Attributo | Significato |
|---|---|
| `lines` | Numero di righe, default 3 (1-20). |
| `width` | Larghezza del blocco, default `100%`. |
| `height` | Dimensione riga/cerchio, default `12px`. |
| `gap` | Spazio tra le righe, default `8px`. |
| `circle` | Mostra un cerchio invece delle righe. |
| `animated` | `false` disabilita l'animazione. |
| `duration` | Ciclo dello shimmer, default `1.4s`. |

`simpleui check` rifiuta i figli di skeleton. Nel designer scegliere **Content >
Skeleton loader**.

## Spinner

`spinner "Etichetta"` mostra l'anello CSS. Aggiungi `src` per usare invece
un'immagine o una GIF animata:

```sui
spinner "Loading" src: "spinner.gif"
spinner "Loading" src: "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
```

L'etichetta diventa il testo `alt` dell'immagine. Gli spinner con immagine
generano un wrapper `.sui-spinner.sui-spinner-image` con dentro un'immagine
`.sui-spinner-img`; l'anello CSS, il bordo e la rotazione sono disattivati.
L'immagine mantiene le proporzioni ed è limitata da `--sui-spinner-size`
(default `96px`). `src` può essere un'espressione reattiva; il runtime usa la
guardia URL per immagini, che accetta `data:image/...`, `http(s)` e percorsi
relativi e rifiuta `javascript:` e altri schemi attivi. `alt` sovrascrive
l'etichetta.

Quando `src` è presente, `variant` e le variabili di struttura, stile e
animazione vengono ignorate. Nel designer la sezione **Spinner** ha una
casella **Image / GIF** con selettore degli asset e **Upload image**; i campi
struttura, stile e animazione sono disabilitati mentre è attiva. Vedi
[GUI editor: modalità immagine/GIF dello spinner](GUI_EDITOR_IT.md#modalità-immaginegif-dello-spinner).

## Carousel

`carousel` mostra una `slide` alla volta. La slide attiva può essere legata a
uno stato intero con `value`; frecce, dots e tasti sinistra/destra lo aggiornano:

```sui
state slide = 0

carousel value: slide height: "180px" {
	slide label: "Benvenuto" {
		h2 "Benvenuto"
		p "La prima slide."
	}
	slide label: "Layout" {
		h2 "Layout"
		p "La seconda slide."
	}
}
```

| Attributo | Significato |
|---|---|
| `value` | Indice della slide attiva (stato o intero), default 0. |
| `height` | Altezza minima della slide. |
| `controls` | Pulsanti precedente/successivo, default true. |
| `indicators` | Dots, default true. |
| `loop` | Ripartizione circolare, default true. |
| `autoplay` | Avanzamento automatico, default false. |
| `interval` | Ritardo autoplay in ms, default 4000. |
| `duration` | Durata della transizione slide, default `.28s`. |
| `easing` | Easing della transizione, default `ease`. |

La `label` di ogni slide è il nome accessibile. Un carousel richiede almeno una
slide e rifiuta altri figli. Nel designer scegliere **Content > Carousel**.

## Stepper

`stepper` mostra passi numerati e un indice corrente. Legare `value` a uno stato
intero; con `clickable: true` (default quando è legato) il click su un passo o
Invio/Spazio lo aggiorna:

```sui
state step = 1

stepper value: step {
	step "Dettagli" description: "Nome e proprietario"
	step "Pagamento" description: "Metodo"
	step "Conferma"
}
```

I passi completati mostrano una spunta, quello corrente è evidenziato e la
`description` opzionale appare sotto l'etichetta. `vertical: true` impila i
passi. Uno stepper richiede almeno un figlio `step` e rifiuta gli altri. Nel
designer scegliere **Content > Stepper**.

## Command palette

`commandbar` è un elenco di comandi ricercabile. Senza `open` resta inline; con
uno stato `open` diventa un overlay che Ctrl/Cmd + `hotkey` apre, Esc chiude, le
frecce scorrono e Invio esegue:

```sui
state showPalette = false

fn newProject() -> null {
}

commandbar "Cerca comandi" open: showPalette hotkey: "k" {
	command "Nuovo progetto" shortcut: "Ctrl N" onClick: newProject
	command "Apri impostazioni" shortcut: "Ctrl ," onClick: openSettings
}
```

| Attributo | Significato |
|---|---|
| `open` | Stato opzionale che controlla l'overlay. |
| `hotkey` | Singolo tasto usato con Ctrl/Cmd, ad esempio `k`. |
| `placeholder` | Placeholder di ricerca; default al testo dell'elemento. |
| `animation` | `fade`, `scale`, `slide` o `none`, default `fade`. |
| `duration` | Durata dell'effetto, default `140ms`. |
| `easing` | Easing dell'effetto, default `ease-out`. |
| `delay` | Ritardo dell'effetto, default `0ms`. |

Ogni `command` richiede il testo; `shortcut` è solo visuale e `onClick` esegue
l'azione. La digitazione filtra per etichetta e scorciatoia, e appare "No
matching commands" quando non c'è corrispondenza. Nel designer scegliere
**Overlay > Command palette**.

## Tabs

`tabs` crea una barra di schede; ogni figlio diretto `tab` fornisce l'etichetta
del pulsante (il suo testo) e il contenuto del pannello (i suoi figli):

```sui
tabs "Settings" {
	tab "General" {
		p "General settings"
		input "Name" value: name
	}
	tab "Advanced" {
		p "Advanced settings"
	}
	tab "About" {
		p "SimpleUI tabs example."
	}
}
```

| Elemento | Testo | Figli |
|---|---|---|
| `tabs "Nome accessibile"` | Nome accessibile (`aria-label`) | elementi `tab` diretti |
| `tab "Etichetta pulsante"` | Pulsante visibile della scheda | contenuto del pannello |

La prima scheda è attiva inizialmente. Il click o i tasti sinistra/destra/Home/End
cambiano pannello; la scheda selezionata viene ripristinata dopo un normale
re-render della view. Solo i figli diretti `tab` diventano schede. Il markup
generato usa `role=tablist`, pulsanti `role=tab` e pannelli `role=tabpanel`; le
etichette sono escapate normalmente.

Nel designer inserire **Tabs** da **Layout → Navigation**. Parte con due schede.
Fare clic con il tasto destro su una riga `tabs` per **Add tab**, oppure su una
riga `tab` per **Add tab below**, **Edit label**, **Duplicate**, **Move up/down**
e **Delete**. Ogni scheda è un contenitore: trascinare il contenuto del pannello
al suo interno. Vedere [Editor GUI: tabs](GUI_EDITOR_IT.md#tabs-control).

## Accordion

`accordion` crea sezioni espandibili impilate. Ogni figlio diretto
`accordionitem` fornisce l'etichetta della sezione (il suo testo) e il contenuto
del pannello (i suoi figli):

```sui
accordion "FAQ" {
	accordionitem "Che cos'è SimpleUI?" {
		p "Un piccolo linguaggio UI reattivo."
	}
	accordionitem "Posso usare il mio CSS?" {
		p "Sì, con un blocco css oppure dal designer."
	}
}
```

| Elemento | Testo | Figli |
|---|---|---|
| `accordion "Nome accessibile"` | Nome accessibile (`aria-label`) | elementi `accordionitem` diretti |
| `accordionitem "Etichetta sezione"` | Pulsante di espansione visibile | contenuto del pannello |

La prima sezione è aperta inizialmente. Cliccando l'intestazione di una sezione
si apre e si chiudono le altre; cliccando l'intestazione già aperta si chiude.
La sezione aperta viene ripristinata dopo un normale re-render della view. Le
intestazioni sono pulsanti nativi con `aria-expanded`; etichette e contenuti
sono escapati normalmente.

Nel designer inserire **Accordion** da **Layout → Navigation**. Parte con due
elementi. Fare clic con il tasto destro su una riga `accordion` per **Add
accordion item**, oppure su una riga `accordionitem` per **Add item below**,
**Edit label**, **Duplicate**, **Move up/down** e **Delete**. Ogni elemento è un
contenitore per il contenuto del pannello.
## Notifiche toast

I toast sono messaggi temporanei mostrati dal codice:

```sui
toast class: "my-toast-host"

fn save() -> null {
	toast.show("Salvato", "success", 3000)
}

fn clear() -> null {
	toast.clear()
}
```

`toast.show(message [, kind [, duration]])`:

| Argomento | Significato |
|---|---|
| `message` | Testo mostrato nella notifica; viene sempre trattato come testo |
| `kind` | `info` (default), `success`, `warning` oppure `error` |
| `duration` | Millisecondi prima della chiusura automatica; default `4000`; `0` la mantiene aperta |

`toast.clear()` rimuove tutte le notifiche visibili. L'utente può chiudere ogni
toast con il pulsante X. Un host conserva al massimo quattro toast; quando ne
arriva un quinto, il più vecchio viene rimosso.

L'elemento opzionale `toast` è un host:

```sui
toast class: "my-toast-host"
```

Senza host il runtime crea automaticamente un host fisso in alto a destra. Un
host inserito nel designer usa ancore orizzontali e verticali indipendenti:

```sui
toast hAlign: "left" vAlign: "bottom" hOffset: "24px" vOffset: "16px"
```

`hAlign` è `left`, `center` o `right`; `vAlign` è `top`, `center` o `bottom`.
Un asse ancorato a un bordo usa il proprio `hOffset`/`vOffset` come distanza dal
bordo. Un asse centrato ignora l'offset e non ne ha bisogno. L'host accetta
anche `duration`, `max` e `icons` per i default runtime. Un host ancorato in
alto impila i nuovi messaggi sotto i precedenti; un host ancorato in basso li
impila sopra i precedenti. Usa `aria-live="polite"`, ogni toast usa
`role="status"` e il contenuto viene assegnato tramite `textContent` (mai HTML
raw).
L'aspetto del toast è normale CSS sulle parti interne: `.sui-toast` (card),
`.sui-toast-icon`, `.sui-toast-message` e `.sui-toast-close`. La classe dell'host
le limita, ad esempio `.my-host .sui-toast { width: 420px; }`. Il runtime legge
anche `--sui-toast-accent`, `--sui-toast-icon-bg`, `--sui-toast-icon-fg`,
`--sui-toast-message-fg` e `--sui-toast-close-fg`.

Nel designer inserire **Toast host** nella sezione **Overlay**. Selezionandolo si
apre il pannello dedicato **Toast** con i due selettori di asse, i campi offset
condizionali, i default runtime e l'anteprima live. Le normali sezioni **Size**,
**Border**, **Box model**, **Typography**, **Background** ed **Effects** sotto
il pannello aggiornano la regola della card; le stesse dichiarazioni compaiono
nel campo **style expr** dell'host, dove il CSS di `.sui-toast` può essere
valutato e modificato. Usare un handler come `onClick: save` per chiamare
`toast.show(...)`. Il campo **Text** non viene usato da questo host. Con più
host, il designer mostra l'anteprima nell'host selezionato mentre `toast.show`
usa il primo host nel documento. Vedere
[Editor GUI: toast](GUI_EDITOR_IT.md#controllo-toast).

## Proprietà Tooltip

Qualsiasi elemento o controllo può avere l'attributo `tooltip`. Il testo viene
mostrato al passaggio del mouse o al focus da tastiera; non serve un controllo
wrapper.

```sui
button "Salva" tooltip: "Salva il documento" onClick: save
button "Elimina" tooltip: "Elimina definitivamente" onClick: remove
```

- `tooltip` è disponibile su tutti gli elementi e controlli consentiti;
- per l'accesso da tastiera il target dovrebbe essere focusabile (ad esempio un
  pulsante, un link o un input);
- il testo è solo testo e viene escapato, mai HTML raw;
- uscita/blur lo nascondono; `Escape` chiude il fumetto aperto;
- il target riceve `aria-describedby` mentre il fumetto è visibile;
- il fumetto usa `role="tooltip"` ed è posizionato nel viewport, quindi funziona
  anche con controlli void come `input`;
- la posizione preferita può essere `auto` (default) o `top`, `right`, `bottom`,
  `left`; auto sceglie il lato con spazio e limita il fumetto;
- i tooltip usano le animazioni condivise (fade, scale, slide, none) con durata,
  easing e ritardo configurabili.

### Impostazioni tooltip del progetto

Il designer offre **Preferences → Project preferences** per i default globali:
posizione (`auto`, `top`, `right`, `bottom`, `left`), ritardo hover, effetto,
durata, easing, ritardo animazione, sfondo, colore testo, colore bordo, raggio e
larghezza massima. Sono scritti nel CSS del progetto come custom properties su
`:root`, quindi valgono per tutti gli elementi con tooltip:

```css
:root {
  --sui-tooltip-place: auto;
  --sui-tooltip-delay: 140ms;
  --sui-tooltip-animation: fade;
  --sui-tooltip-duration: 140ms;
  --sui-tooltip-easing: ease-out;
  --sui-tooltip-anim-delay: 0ms;
  --sui-tooltip-bg: #1d2129;
  --sui-tooltip-fg: #e6e6ef;
  --sui-tooltip-border: #3a4152;
  --sui-tooltip-radius: 6px;
  --sui-tooltip-max-width: 260px;
}
```

## Combobox

Controllo a scelta singola con ricerca. Il testo dell'elemento è la label,
`options` è una lista separata da virgole e `value` collega il valore salvato.

```sui
state country = ""
state countryOptions = ""

combobox "Paese"
	options: countryOptions
	value: country
	onChange: countryChanged
	placeholder: "Cerca un paese..."
	empty: "Nessun risultato"
	maxOptionLength: 24
	maxOptions: 8
```

- `options`: etichette separate da virgole. Può essere una lista literal come
  `"Italia,Francia,Germania"` oppure uno stato stringa come `countryOptions`.
  Usare `Etichetta=valore` quando il valore salvato è diverso dal testo
  mostrato.
- Uno stato `options` dinamico viene valutato a ogni render della view:
  assegnando una nuova stringa separata da virgole da un handler, la lista si
  aggiorna automaticamente. È il percorso supportato per liste caricate a
  runtime o da remoto.
- Usare il backslash per proteggere una virgola dentro un'etichetta. A runtime
  i dati contengono `Washington\, D.C.`; nel sorgente SUI il backslash va
  scritto due volte: `"Washington\\, D.C."`.
- `maxOptionLength`: numero massimo di caratteri accettati per ogni etichetta o
  valore salvato quando la lista viene caricata. Se un solo elemento supera il
  limite, l'intera lista viene rifiutata e viene mostrato un errore; non viene
  troncata silenziosamente. `0` o assenza significano nessun limite applicativo
  (il limite massimo del runtime è 10.000 caratteri).
- `maxOptions`: numero massimo di elementi accettati quando la lista viene
  formata. Se la lista caricata è più grande, l'intera lista viene rifiutata e
  viene mostrato un errore. `0` o assenza significano nessun limite applicativo;
  il limite massimo del runtime è 2.000 elementi.
- `value`: binding di stato; selezionando un'opzione viene aggiornato
  automaticamente.
- `onChange`: funzione senza argomenti opzionale, chiamata dopo l'aggiornamento
  del valore. Non viene chiamata quando una sostituzione runtime azzera
  automaticamente un valore non più presente.
- `required`: rende il controllo obbligatorio. Il campo di ricerca riceve
  `aria-required`; se non viene selezionato un valore, alla perdita del focus
  viene mostrato un errore inline.
- `placeholder`: testo del campo di ricerca; default `Search...`.
- `empty`: testo mostrato quando il filtro non trova risultati; default
  `No results`.
- `disabled`: disabilita input e pulsante.
- `placement` (`auto`, `bottom`, `top`, `left`, `right`) e `animation`
  (`fade`, `scale`, `slide`, `none`), `duration`, `easing`, `delay`: il popup
  si ribalta e resta nella viewport e usa l'effetto condiviso. La larghezza
  segue il controllo.
- La ricerca è case-insensitive. Senza wildcard cerca una sottostringa. `*`
  corrisponde a qualsiasi sequenza, `?` a un carattere e `\` protegge una
  wildcard per cercarla letteralmente (ad esempio `\*`).
- Dati opzioni runtime non validi mostrano un errore sotto il controllo e non
  visualizzano una lista parziale. Etichette/valori vuoti, caratteri di
  controllo, escape incompleti, opzioni oltre `maxOptionLength` e liste oltre
  `maxOptions` vengono rifiutati nel loro insieme.
- Se una lista sostitutiva valida non contiene più il valore selezionato, il
  valore e lo stato collegato vengono azzerati automaticamente. Se il valore è
  ancora presente viene mantenuto e cambia solo la lista.
- Tastiera: `ArrowDown`/`ArrowUp`, `Enter`, `Escape`, `Tab`.
- L'input usa `role="combobox"`, il popup `role="listbox"` e le righe
  `role="option"` con `aria-selected`.
- Etichette e testo vuoto sono nodi di testo, mai HTML raw.

```sui
state countryOptions = ""

fn loadCountries() -> null {
	# Valore runtime: Italia,Francia,Germania,Spagna=es,Washington\, D.C.
	countryOptions = "Italia,Francia,Germania,Spagna=es,Washington\\, D.C."
}

fn clearCountries() -> null {
	countryOptions = ""
}
```

Il valore selezionato è in un input nascosto, mentre l'input visibile serve alla
ricerca. Se una sostituzione runtime rimuove il valore selezionato, il valore e
lo stato collegato vengono azzerati automaticamente; se è ancora presente cambia
solo la lista. Per dati remoti, validare prima la risposta e poi assegnare la
stringa separata da virgole allo stato.

### Sicurezza degli aggiornamenti runtime

- Considerare la stringa opzioni non attendibile quando arriva da API, file o
  storage. Il controllo scrive le etichette con `textContent` e gli attributi
  dinamici generati sono escapati, quindi non può iniettare markup.
- Il parser rifiuta escape incompleti, etichette/valori vuoti, caratteri di
  controllo, opzioni oltre 10.000 caratteri, liste oltre 500.000 caratteri e
  più di 2.000 opzioni. Il messaggio d'errore è generico e non mostra il
  payload.
- La ricerca wildcard usa un matcher glob iterativo, non espressioni regolari
  fornite dall'utente, evitando regex injection e backtracking catastrofico. Le
  query sono limitate a 256 caratteri.
- `maxOptionLength` e `maxOptions` sono limiti di validazione: l'intera lista
  candidata viene controllata prima di essere pubblicata. Non sostituiscono i
  limiti di dimensione e schema sulla risposta remota.
- L'azzeramento automatico di un valore rimosso è una mutazione di stato:
  chiama il setter generato con `""` e **non** chiama `onChange`. Evitare di
  pubblicare liste vuote o parziali durante il caricamento; usare uno stato di
  loading e pubblicare solo la lista completa.

## Tree

Albero annidato accessibile da tastiera. `tree` fornisce titolo visibile e
binding del valore selezionato; ogni `treeitem` fornisce etichetta, `key` stabile
e stato iniziale `open` opzionale.

```sui
state selected = "main"
state renameKey = ""
state renameValue = ""

tree "File progetto"
	name: "files"
	value: selected
	onChange: selectedChanged
	editMode: "editable"
	editDepth: 2
	renameKey: renameKey
	renameValue: renameValue
	onRename: renamed {
	treeitem "src" key: "src" open: true {
		treeitem "main.go" key: "main"
		treeitem "components" key: "components" {
			treeitem "tree.go" key: "tree"
		}
	}
	treeitem "README.md" key: "readme"
	treeitem "generated" key: "generated" rename: false
}
```

- Il testo dell'elemento tree è il titolo visibile. `name` è il target runtime
  usato dal modulo core `tree`; usa il titolo come default ma non viene mostrato.
  Impostare un nome univoco per import/export senza cambiare il titolo.
- `value` collega la key dell'elemento selezionato a uno stato stringa.
- `onChange` viene chiamata dopo l'aggiornamento dello stato.
- `treeitem key`: valore stabile salvato nello stato. Se assente viene usato il
  percorso dell'elemento.
- `treeitem open: true`: stato di espansione iniziale per un genitore.
- `guides: false` disattiva le guide sottili di indentazione (default attive).
- `truncate: false` manda a capo le etichette lunghe invece di troncarle. Le
  etichette troncate espongono sempre il testo completo come tooltip nativo;
  `treeitem tooltip: "..."` imposta un testo personalizzato.
- `tooltips: true` usa il tooltip stilizzato del tree invece del title nativo.
  Stile tramite `--sui-tree-tooltip-bg`, `-fg`, `-border`, `-radius`, `-padding`,
  `-max-width`, `-font-size`, `-delay` e `-placement`.
- `rowHeight: 32` imposta l'altezza minima di riga (12-120). Default `28`.
- `iconLeaf`, `iconParent` e `iconParentOpen` scelgono il glifo predefinito per
  elementi senza figli, padri chiusi e padri aperti. Usa
  `iconLeaf`/`iconParent`/`iconParentOpen`; i vecchi nomi `icon`/`iconOpen`
  vengono rifiutati con una diagnostica. `toggleClosed` e
  `toggleOpen` sostituiscono i simboli di espansione del ramo. `treeitem
  icon`/`iconOpen` sovrascrivono i default. Icone e simboli di ramo sono testo o emoji inerte,
  mai markup.
- `treeitem badge: "12"` mostra una pill; `badgeType` è `neutral`, `info`,
  `success`, `warning` o `danger`. `meta` mostra testo allineato a destra.
  `counts: true` aggiunge il numero di figli alle righe padre.
- `checkbox: true` mostra checkbox gerarchiche. `treeitem checked: true` imposta
  lo stato iniziale; un padre deriva true, false o mixed dai figli. Il click
  aggiorna i discendenti e ricalcola gli antenati. Lo stato è persistito in
  `localStorage`.
- `onCheck` viene chiamata dopo il click su una checkbox; non riceve argomenti.
  Usare `tree.exportJson(name)` per leggere i campi `checked` risultanti.
- `onTreeChange` viene chiamata quando i dati dell'albero cambiano:
  aggiunta/eliminazione, spostamento, indent/outdent, import/clear, rinomina,
  checkbox, figli lazy o celle grid. Non riceve argomenti e scatta dopo
  `onMove`/`onRename`/`onCheck` per la stessa operazione.
- `multiSelect: true` abilita Ctrl/Command click per il toggle e Shift click per
  l'intervallo. `selection` collega uno stato stringa che riceve un array JSON
  delle key selezionate; `value` conserva l'ultima key. Lo stato `selection`
  fornisce anche la selezione iniziale.
- `search: true` aggiunge una casella di ricerca inline. Filtra le righe, espande
  i rami con match, evidenzia tutte le occorrenze e mostra conteggio o empty
  state. La ricerca include anche badge, metadata, size, date e status.
  Ctrl/Command+F la focalizza; Escape la pulisce. Il termine sopravvive ai
  re-render reattivi nella sessione della pagina.
- `drag: true` abilita drag and drop con indicatori sopra/sotto/dentro. I drop
  nel nodo trascinato o in un suo discendente vengono rifiutati. `onMove` viene
  chiamata dopo uno spostamento valido; usare `tree.exportJson(name)` per
  persistere il nuovo ordine.
- `contextMenu: true` aggiunge il menu contestuale integrato (aggiungi
  figlio/fratello, rinomina, duplica, sposta su/giù, indent/outdent, elimina,
  espandi/comprimi sottoalbero, copia key, espandi/comprimi tutto).
  `onContextMenu` viene chiamata al click destro dopo la selezione, per menu
  applicativi.
- `columns: "owner,size"` mostra intestazione tree-grid e celle per key di
  colonna arbitrarie. `columnsEnabled: false` mantiene la configurazione ma
  nasconde intestazione e celle; il designer lo espone come **Show columns**.
  `columnLabels: "Owner,Size"` imposta le intestazioni in ordine e
  `columnWidths: "120,90"` le larghezze. `treeitem size`, `date` e `status` sono
  campi di comodo; qualsiasi altra colonna legge `treeitem cells`, una lista
  `key=value` come `cells: "owner=Anna; role=Admin"` (il runtime import accetta
  anche JSON). `sort: "asc"`/`"desc"` ordina i figli per etichetta; ometterlo per
  mantenere l'ordine manuale.
- API colonne runtime: `tree.setColumns(name, json)` crea, rinomina, ridimensiona
  o riordina le colonne; `tree.getColumns(name)` restituisce la lista corrente;
  `tree.moveColumn(name, from, to)` riordina; `tree.setCell(name, key, column,
  value)` aggiorna una cella. Le override runtime sopravvivono ai re-render
  reattivi nella sessione della pagina.
- `fontStep: 1.5` riduce la dimensione del font di 1.5px per ogni livello sotto
  la radice. Il default `0` mantiene la stessa dimensione per tutti i livelli.
- `persistSelection: true` ripristina le key selezionate da `localStorage` dopo
  un reload quando non è collegato uno stato `selection`.
- `treeitem lazy: true` mostra la freccia senza figli. L'espansione chiama
  `onLoad` e scrive la key nello stato collegato a `loadKey`. Caricare i dati con
  `tree.setChildren(name, key, jsonText)`, oppure mostrare un errore con
  `tree.setChildrenError(name, key, message)`. Un indicatore di caricamento resta
  visibile finché i figli non arrivano; la riga di errore ha **Retry** e
  `tree.retryChildren(name, key)` riprova da codice.
- `tree.lastMove(name)` restituisce JSON `{key, parent, index}` dell'ultimo
  drag/riordino/duplica/elimina; `tree.undoMove(name)` ripristina l'albero
  precedente da uno stack in memoria di 20 passi.
- `tree.lastChange(name)` restituisce il JSON dell'ultima modifica dati, con
  `op` (`add`, `delete`, `duplicate`, `move`, `indent`, `outdent`, `rename`,
  `check`, `load`, `cell`, `import`, `clear` o `undo`), `key`, `parent`,
  `index` e campi specifici come `column`, `label`, `checked`, `count`,
  `target` o `mode`.
- Il menu contestuale integrato offre anche **Add child item**, **Add item
  below**, **Move up/down**, **Indent/Outdent** e **Delete item** (tutti
  annullabili con `tree.undoMove`).
- `tree.expandAll`, `tree.collapseAll` e `tree.expandTo(name, level)` controllano
  l'espansione da codice. `tree.selectedKeys(name)` e `tree.checkedKeys(name)`
  restituiscono array JSON. `tree.search(name, term)` pilota il filtro da codice.
- `virtual: true` applica `content-visibility: auto` alle righe per far saltare
  al browser layout/paint fuori schermo. È un'ottimizzazione di rendering, non
  virtualizzazione del DOM: alberi molto grandi costruiscono comunque il DOM una
  volta.
- Le interazioni del tree (espansione, selezione, check, ricerca, riordino)
  aggiornano il DOM esistente in modo incrementale; un cambio di stato SUI
  ri-renderizza comunque la view dal sorgente, con lo stato del tree ripristinato
  da binding e localStorage.
- `editMode: "none" | "editable" | "persistent"` controlla la rinomina inline;
  `"persistent"` salva le rinomine in `localStorage` tra reload/riavvio.
  `editDepth: 2` consente i livelli 1-2. `0` significa tutti.
- `treeitem rename: true|false` sovrascrive il default. `false` blocca il nodo
  e i suoi discendenti. Attributo assente = eredita.
- `renameIcon: "✏️"` cambia il simbolo della matita; `showRenameIcon: false`
  la nasconde. Sono testo, mai HTML. `treeitem renameIcon`/`showRenameIcon`
  sovrascrivono il valore del tree; ometterli per ereditare. Nascondere il
  pulsante lascia comunque disponibili F2 e doppio click.
- Rinomina: doppio click, F2 o pulsante matita; Enter conferma, Escape annulla; etichette vuote
  rifiutate. La key non cambia.
- `renameKey`/`renameValue` ricevono key e nuova etichetta prima di `onRename`.
- Il click sulla riga seleziona; il click sulla freccia espande/comprime
  soltanto.
- Espansione e selezione sono indipendenti; lo stato aperto è persistito per
  percorso in `localStorage`. Anche le etichette rinominate sono salvate
  localmente per albero/percorso.
- Tastiera: Up/Down sui visibili; Right espande o entra; Left comprime o torna
  al genitore; Enter seleziona; Spazio seleziona oppure, con `checkbox: true`,
  attiva la checkbox; Shift+Up/Down e Shift+Home/End estendono una
  multi-selezione; Ctrl/Command+A seleziona i visibili; Ctrl/Command+F focalizza
  la ricerca; F2 rinomina; Home/End saltano.
- ARIA: `role="tree"`, `role="treeitem"`, `role="group"`, `aria-level`,
  `aria-expanded`, `aria-selected`, `aria-readonly` e, con `checkbox: true`,
  `aria-multiselectable` più `aria-checked="true|false|mixed"`. La
  multi-selezione usa `aria-selected` su ogni elemento selezionato. Un solo
  elemento è nel tab order.
- Lo stile visivo è trasportato da `style expr` come proprietà CSS custom sul
  nodo radice dell'albero. Preset del tree: **Default**, **Compact**,
  **Comfortable**, **Dense**, **Bordered**, **Soft rounded**, **Minimal**,
  **Neon**, **Forest**, **Amber**, **Ocean**, **High contrast**, **Outline**,
  **Mono**, **Candy** e **Cards**; selezionandone uno viene riscritto
  `style expr`.
- Ogni `treeitem` espone preset di nodo: **Inherit tree style**, **Muted**,
  **Accent**, **Success**, **Warning**, **Danger**, **Folder** e **Compact row**.
  Il preset scrive variabili di riga nello `style expr` del solo nodo; **Custom**
  permette di modificare direttamente l'espressione.
- Le variabili includono `--sui-tree-gap`, `--sui-tree-row-padding`,
  `--sui-tree-row-radius`, `--sui-tree-row-bg`, `--sui-tree-hover-bg`,
  `--sui-tree-selected-bg`, `--sui-tree-selected-fg`, `--sui-tree-title-color`,
  `--sui-tree-title-size`, `--sui-tree-panel-bg`,
  `--sui-tree-panel-border`, `--sui-tree-panel-radius`,
  `--sui-tree-panel-padding`, `--sui-tree-indent`, `--sui-tree-guide-color`,
  `--sui-tree-icon-color`, `--sui-tree-badge-bg`, `--sui-tree-badge-fg`,
  `--sui-tree-meta-color`, `--sui-tree-check-on`, `--sui-tree-check-border`,
  `--sui-tree-check-bg`, `--sui-tree-row-height`, `--sui-tree-tooltip-bg`,
  `--sui-tree-tooltip-fg`, `--sui-tree-tooltip-border`,
  `--sui-tree-tooltip-radius`, `--sui-tree-tooltip-padding`,
  `--sui-tree-tooltip-max-width`, `--sui-tree-tooltip-font-size`,
  `--sui-tree-tooltip-delay` e `--sui-tree-tooltip-placement`. I controlli Extra
  usano `--sui-tree-rename-bg`, `-fg`, `-font`, `-size`, `-border`,
  `-border-width`, `-radius`, `-padding`, `-hover-bg`, `-hover`,
  `--sui-tree-count-bg`, `-fg`, `-font`, `-size`, `-weight`, `-padding`,
  `-radius`, `-border`, `-border-width` e `-min-width`.
- Modificare direttamente l'espressione `style: cssSafe("...")` per valori
  personalizzati. Lo stile dinamico resta soggetto alla policy `cssSafe`.
  Poiché `cssSafe` accetta dichiarazioni senza virgole, nello `style expr` usare
  colori esadecimali a 8 cifre come `#4c5cf761` invece di `rgba(...)`.
- Dati: `tree.importJson`/`tree.importXml` sostituiscono i nodi da testo;
  `tree.importJsonFile`/`tree.importXmlFile` leggono un file selezionato;
  `tree.exportJson`/`tree.exportXml` restituiscono il testo; `tree.downloadJson`/
  `tree.downloadXml` scaricano un file; `tree.clear` rimuove tutti i nodi.
  Espansione: `tree.expandAll`, `tree.collapseAll`, `tree.expandTo`. Dati lazy:
  `tree.setChildren`, `tree.setChildrenError`. Selezione:
  `tree.selectedKeys`, `tree.checkedKeys`; filtro: `tree.search`. Cronologia:
  `tree.lastMove`, `tree.lastChange`, `tree.undoMove`; retry lazy: `tree.retryChildren`; colonne:
  `tree.setColumns`, `tree.getColumns`, `tree.moveColumn`, `tree.setCell`.
  Schema e limiti in [`STDLIB_IT.md`](STDLIB_IT.md#tree-module-tree-core).
- I dati importati vengono applicati al tree con il `name` corrispondente;
  sopravvivono ai re-render e sostituiscono solo la struttura dei nodi (titolo,
  stile e binding restano). Etichette, key, icone, badge, metadata e celle grid
  sono testo, mai HTML; `checked`, `lazy` e le modifiche da selezione/drag sono
  preservate nel modello runtime; le dichiarazioni DTD/entità XML vengono
  rifiutate.
- Etichette, key e testo sono escapati; nessun HTML raw.

Le key dovrebbero essere uniche nell'albero. La key selezionata è dato
applicativo normale: validarla prima di usarla per accedere a risorse server.
`editMode` è un'indicazione UI, non un'autorizzazione. In modalità `persistent`
le etichette rinominate sono in `localStorage`; in modalità `editable` durano per
la sessione della pagina. Usare `renameKey`/`renameValue` e `onRename` per
persistenza applicativa/server.

## Editor di testo avanzato

`richtext` è un controllo core (senza import) per documenti testuali
strutturati. Il documento è un albero JSON, mai una stringa HTML; l'HTML
generato e una rappresentazione `contenteditable` di quell'albero.

```sui
state doc = ""
richtext "Articolo" name: "article" value: doc height: "420px" theme: "dark"
  toolbarItems: "undo,redo|bold,italic,underline|link,image,table"
  onChange: changed onInput: typing
```

### Attributi

| Attributo | Significato |
|---|---|
| `name` | Nome runtime usato da `richtext.*` (univoco per editor) |
| `value` | Stato collegato che contiene il documento JSON |
| `onChange` / `onInput` | Handler per modifica e digitazione |
| `onReady` / `onSelectionChange` | Callback di pronto e selezione |
| `height` | Altezza minima dell'editor (lunghezza CSS) |
| `placeholder` | Testo per documento vuoto |
| `readonly` | Disabilita la modifica |
| `menu`, `toolbar`, `contextToolbar`, `statusbar` | Mostra/nascondi aree |
| `toolbarItems` | Comandi separati da virgola, `|` separa i gruppi |
| `theme` | Tema dell'interfaccia: `dark` (default), `light` o `auto`. Un unico pulsante nella toolbar alterna dark/light e cambia solo la chrome dell'editor |
| `contentTheme` | Tema della pagina: `light` (default, foglio bianco) o `dark`. Indipendente da `theme` |
| `header` / `footer` | Testo ripetuto su ogni pagina in PDF, DOCX, ODT e stampa. `\{page\}` e `\{pages\}` diventano numero e totale pagine (escape delle graffe nelle stringhe SUI) |
| `template` | `article`, `letter`, `report` o `compact` |

I gruppi della toolbar (separati da `|`) sono toolbar indipendenti e riposizionabili:
trascina la maniglia ⣿ per sganciarle, poi rilasciale sulla barra superiore per
reinserirle nello slot più vicino alla fine di una toolbar già presente (o al bordo
sinistro se la barra è vuota); le altre toolbar si spostano. Le posizioni flottanti
vengono ricordate per ogni editor. Un unico pulsante a destra alterna il tema dell'interfaccia.

**Trova e sostituisci:** `Ctrl+F` (o Modifica > Trova e sostituisci) apre la barra;
Invio/F3 successivo, Shift+F3 precedente, Esc chiude. **Incolla** da Word/Excel/HTML
viene sanitizzato nel modello e conserva grassetto/corsivo, colori, link, tabelle e
immagini (le immagini `blob:` vengono convertite in data URL). Il correttore
ortografico del browser è attivo e l'editor espone ruoli ARIA textbox/toolbar.
**Anteprima di stampa / PDF:** `Ctrl+P` (File > Stampa / Salva come PDF)
apre la stampa del browser su pagine A4 con header/footer ripetuti.
**Liste:** elenchi puntati e numerati si annidano con Tab e si riportano indietro con
Shift+Tab; l'export DOCX/ODT usa la numerazione nativa.

**Vista paginata:** il pulsante **Pages** nella barra di stato (o Visualizza >
Vista paginata) mostra fogli A4 con margini, header/footer e righe di tabella
spezzate tra le pagine; l'intestazione della tabella si ripete su ogni pagina.
Premi **Edit** per tornare all'editor continuo.

Token toolbar: `undo`, `redo`, `copy`, `cut`, `paste`, `styleSelect`,
`fontFamily`, `fontSize`, `bold`, `italic`, `underline`, `strike`, `textColor`,
`highlight`, `superscript`, `subscript`, `clearFormat`, `alignLeft`,
`alignCenter`, `alignRight`, `justify`, `indent`, `outdent`, `lineHeight`,
`spaceBefore`, `spaceAfter`, `bulletList`, `orderedList`, `blockquote`,
`codeBlock`, `link`, `image`, `table`, `symbol`, `hr`, `pageBreak`.

Quando il cursore è dentro una tabella compare una **toolbar della tabella**
subito sotto la toolbar principale. Ha pulsanti con icone per aggiungere o
eliminare righe e colonne, attivare l'intestazione, unire o dividere celle,
allineare le celle ed eliminare la tabella. Clicca o fai Shift+clic sulle celle
per selezionare un intervallo rettangolare. Trascina la maniglia ⣿ per
spostare la toolbar dove vuoi; doppio clic sulla maniglia o il pulsante dock la
riporta sotto la toolbar principale. Le modifiche entrano nella cronologia undo/redo. Clicca fuori da un menu per chiuderlo. Trascina i bordi
destra/inferiore di una tabella per ridimensionare colonne o righe, oppure
l'angolo in basso a destra per ridimensionare l'intera tabella; le dimensioni
sono salvate nel documento JSON. File > Import ed Export sono sottomenu.
L'azione **Seleziona tutta la tabella** fa applicare l'allineamento alla
tabella stessa. Seleziona due o più celle e copiale: incollando in un'altra
cella viene riempito un blocco corrispondente (funziona anche il testo TSV di
altre applicazioni).

### Modello del documento

```json
{
  "type": "doc",
  "children": [
    { "type": "heading", "level": 1, "children": [
      { "type": "text", "text": "Titolo", "marks": [] }
    ]},
    { "type": "paragraph", "children": [
      { "type": "text", "text": "Grassetto", "marks": [{ "type": "bold" }] }
    ]},
    { "type": "bullet_list", "children": [
      { "type": "list_item", "children": [
        { "type": "paragraph", "children": [
          { "type": "text", "text": "Voce", "marks": [] }
        ]}
      ]}
    ]}
  ]
}
```

Tipi blocco: `paragraph`, `heading`, `bullet_list`, `ordered_list`,
`list_item`, `blockquote`, `code_block`, `image`, `table`, `table_row`,
`table_cell`, `hr`. Tipi inline: `text` con mark (`bold`, `italic`,
`underline`, `strike`, `code`, `sup`, `sub`, `link`, `color`, `background`,
`font`, `size`), più `break` e `image` inline.

### API

`richtext` è sempre disponibile; le funzioni ricevono il `name` dell'editor.

| Funzione | Descrizione |
|---|---|
| `richtext.get(name)` | Documento come JSON |
| `richtext.set(name, jsonOrMap)` | Sostituisce il documento |
| `richtext.command(name, command, [value])` | Esegue un comando di formattazione/blocco/plugin |
| `richtext.format(name, format, [value])` | Alias di `command` |
| `richtext.insertText(name, text)` | Inserisce testo al cursore |
| `richtext.insertImage(name, src, [alt])` | Inserisce un'immagine |
| `richtext.insertLink(name, href, [text])` | Crea o avvolge un collegamento |
| `richtext.insertTable(name, rows, cols)` | Inserisce una tabella |
| `richtext.undo(name)` / `richtext.redo(name)` | Cronologia |
| `richtext.exportFormat(name, format)` | JSON, HTML, Markdown, XML o testo semplice |
| `richtext.exportBytes(name, format)` | Byte TXT, RTF, DOC, DOCX, ODT, PDF, HTML, Markdown, JSON o XML |
| `richtext.importFormat(name, text, format)` | Importa testo nel documento |
| `richtext.openFile(name)` | Selettore file del browser; importa per estensione |
| `richtext.importBytes(name, bytes, filename)` | Importa byte; formato rilevato dal nome file |
| `richtext.saveAs(name, format)` | Scarica il documento nel formato scelto |
| `richtext.registerPlugin(name, plugin)` | Registra comandi e voci toolbar |
| `richtext.plugins()` | Nomi dei plugin registrati |
| `richtext.status(name)` | JSON con parole/caratteri/paragrafi/cursore |
| `richtext.selection(name)` | JSON della selezione |
| `richtext.setToolbar(name, items)` | Riconfigura i gruppi toolbar |
| `richtext.setTemplate(name, template)` | Applica un modello documento |
| `richtext.setTheme(name, theme)` | Cambia tema `dark`, `light` o `auto` |
| `richtext.destroy(name)` | Distrugge l'istanza runtime |

Tutti i comandi sono centralizzati: lo stesso comando `bold` può essere
richiamato da toolbar, menu, scorciatoia, plugin o codice host. Undo/redo è una
cronologia di snapshot JSON (fino a 120 passi).

### File del documento

`File > Open file...` apre il selettore file del browser; `File > Save as`
scarica il documento. Formati supportati:

| Formato | Import | Export | Note |
|---|---|---|---|
| TXT | sì | sì | Paragrafi di testo semplice |
| RTF | sottoinsieme pratico | sì | Paragrafi, formattazione di carattere e paragrafo, font, allineamento, tabelle con celle unite e sfondo, caselle di controllo, immagini di intestazione/piede pagina e immagini PNG/JPEG incorporate |
| DOC | best effort | contenuto RTF con estensione `.doc` | Mantiene font, dimensioni, colori, tabelle, immagini, liste e interruzioni di pagina. Il DOC binario OLE legacy è rifiutato; sono accettati file DOC RTF/HTML/testo |
| DOCX | sì | sì | Paragrafi, run, liste native puntate/numerate, tabelle con intestazione ripetuta, formattazione, font, sfondo celle, immagini, header/footer e interruzioni di pagina WordprocessingML |
| ODT | sì | sì | Testo OpenDocument, paragrafi, formattazione, liste native, tabelle, immagini incorporate, header/footer e interruzioni di pagina |
| PDF | - | sì | Pagine A4 con stili del testo, font, colori, tabelle (celle unite e intestazione ripetuta), immagini PNG/JPEG, liste e header/footer ripetuti con numeri di pagina |
| HTML, Markdown, JSON, XML | sì | sì | Formati già disponibili nell'editor |

I file importati non sono fidati: il parsing ZIP/XML resta nel browser e
l'HTML non sicuro viene comunque convertito nel modello del documento. Le immagini
incorporate vengono estratte da RTF (comprese le immagini di intestazione e piede
pagina), DOCX e ODT come data URL e reinserite
esportando. Selettore
file e download sono azioni dell'utente e non richiedono permessi del bridge
desktop. L'export DOCX/ODT crea il pacchetto nel browser, senza servizi esterni.
L'import DOCX/ODT recupera font, dimensioni, colori, evidenziazioni, link,
spaziatura/rientro dei paragrafi, liste puntate/numerate, testo di
intestazione/piede e colonne/sfondi delle tabelle.

### Import, export e sicurezza

JSON è il formato nativo; HTML, Markdown e XML passano da parser/serializer
dedicati. L'HTML importato viene convertito nel modello e ricostruito con API
DOM: `<script>`, `<style>`, attributi evento e elementi sconosciuti vengono
scartati. Le immagini ammettono `http(s)`, `blob:` e `data:image/...;base64`;
i link ammettono `http(s)`, `mailto:`, `tel:`, frammenti e percorsi relativi.
Gli schemi non sicuri vengono rifiutati. L'editor usa `contenteditable` e le
API Selection del browser, quindi il comportamento del cursore può variare
leggermente tra Chromium, Firefox e WebKit.

### Plugin

Un plugin dichiarativo può aggiungere un pulsante che esegue un comando
integrato:

```sui
let plugin = {
  "commands": {"shout": {"title": "Shout", "command": "bold"}},
  "toolbar": [{"id": "shout", "label": "Shout"}]
}
ok, err = richtext.registerPlugin("demo", plugin)
```

I plugin JavaScript possono registrare comandi personalizzati e renderer di
blocchi tramite `window.__SUI_RICH.registerPlugin`. Vedi
[`examples/richtext_editor/README.md`](../examples/richtext_editor/README.md)
per un esempio eseguibile completo.

## Chart

`chart` visualizza dati JSON, XML o CSV come grafico SVG a barre, linee, area,
torta, ciambella o scatter. Il testo dell'elemento è il titolo visibile. I dati
sono testo, mai markup, e ogni etichetta SVG viene scritta con `textContent`.
Senza dati il controllo resta vuoto ("No data"): i dati arrivano di solito da uno
stato, un'espressione o un'API. Se un sorgente SUI preimposta dei dati è una
scelta intenzionale e vengono mostrati di nuovo a ogni caricamento dell'app; le
modifiche runtime durano per la sessione della pagina salvo persistenza esplicita.

`editMode` è l'impostazione a 3 stati dei controlli dati: `"none"` rende tabella
e diagramma di sola lettura ma continua a mostrare dati, tooltip e selezione;
`"editable"` consente modifiche, import, drag/resize Gantt e editing del diagramma
per la sessione della pagina; `"persistent"` salva le modifiche in `localStorage`
e le mantiene dopo reload e riavvio. In `"persistent"` vengono salvati insieme ai
dati anche le modifiche runtime agli attributi del chart, alle variabili CSS,
al layout della tabella, ai colori delle serie e ai dati del diagramma. Un
`editMode` assente equivale a `"none"`. Tooltip e selezione restano funzioni di
visualizzazione e non vengono disabilitati da `editMode: "none"`.

```sui
state sales = "[\{\"month\":\"Jan\",\"value\":12\},\{\"month\":\"Feb\",\"value\":18\}]"

chart "Monthly sales"
	name: "sales"
	data: sales
	chartType: "bar"
	format: "auto"
	x: "month"
	y: "value"
	overflow: "scroll"
	pointWidth: 48
	showGrid: true
	showLegend: true
	showValues: false
```

`width` e `height` sono opzionali. Se omessi, vuoti, `auto` o percentuali, il
chart riempie il contenitore e lo segue quando cresce; la tabella mantiene
l'altezza fissa in pixel mentre l'area grafica cresce in entrambe le direzioni.
Una dimensione esplicita in pixel mantiene una cornice fissa. In un contenitore
senza altezza definita il chart usa un minimo di 240 px
(`--sui-chart-min-height` sovrascrive il fallback), quindi non collassa.

`theme: "dark"` (predefinito) o `theme: "light"` seleziona il set di token
integrato. Il tema chiaro definisce valori espliciti `--sui-chart-*` per sfondo,
testo, griglia, controlli, tooltip, tabella e scrollbar. Overlay, menu del grafico,
menu contestuali, popup di impostazioni/filtri, selettore risorse e finestra
workload copiano i token del tema attivo, restando leggibili sia in dark sia in
light. I token di focus e selezione (`--sui-chart-table-selection-bg`, `--sui-chart-table-selection-border`,
`--sui-chart-table-focus-row-bg`, `--sui-chart-table-focus-cell-bg` e
`--sui-chart-table-focus-color`) possono essere sovrascritti indipendentemente:
la selezione è condivisa, mentre le tonalità focus hanno valori adattati al
tema chiaro. Un click sul campione colore di una serie nell'intestazione della
tabella o nella legenda apre il selettore colore; il campione pieno nella legenda
mostra il colore esatto usato dalla serie e la scelta aggiorna subito il token
`--sui-chart-series-N`, viene salvata per chart in `localStorage` e non
ricostruisce la tabella né ridisegna l'SVG. `colorPicker: false` lascia visibili
i campioni ma disabilita la modifica (predefinito `true`). Per i diagrammi a blocchi, `settingsButton: false` nasconde l'ingranaggio, `showZoom: false` nasconde i controlli zoom, `showAdd: false` nasconde il pulsante + Add e `showHistory: false` nasconde undo/redo. Un grafico con più
colori in legenda mostra sempre la legenda, anche con `showLegend: false`, così
ogni colore resta identificabile. Le righe Gantt mostrano un campione colore
accanto al nome del task; modificarlo aggiorna barra e tabella. Le voci legenda
Gantt rappresentano gruppi o singoli task quando i colori differiscono. Un task
può anche avere un campo iniziale `color` nei dati JSON, XML o CSV; le scelte
utente vengono salvate separatamente per id task. Gli aggiornamenti del grafico sfumano
l'SVG precedente nel nuovo, salvo `prefers-reduced-motion`; i refresh reattivi
della view preservano il DOM di grafico e tabella e ridisegnano solo i chart con
attributi modificati.

`type` è una parola riservata SUI, quindi l'attributo della geometria si chiama
**`chartType`**: `bar` (default), `barhorizontal`, `barstacked`, `bargrouped`,
`line`, `step`, `smooth`, `area`, `combo`, `pie`, `donut`, `scatter`,
`histogram`, `waterfall`, `radar`, `gauge`, `funnel`, `heatmap`, `candlestick`,
`treemap`, `boxplot`, `bullet`, `dumbbell`, `gantt`, `state` (la geometria
a blocchi si dichiara solo con il controllo `graph`).

I tipi `barstacked`, `bargrouped` e `combo` richiedono almeno due serie
numeriche; `radar` ne accetta una o più. `heatmap` richiede due serie (y e
valore), `candlestick` quattro (open, high, low, close), `boxplot` cinque,
`bullet` due o tre, `dumbbell`/`gantt` due. `histogram` usa `bins`,
`gauge` usa `min`/`max`.

### Dati

`data` è una stringa letterale o un'espressione reattiva (di solito uno stato
`str`). `format` può essere `auto` (default), `json`, `xml` o `csv`. Il rilevamento
automatico usa il primo carattere non spazio: `{`/`[` JSON, `<` XML, altro CSV.

JSON accetta array di oggetti, array di array (con eventuale riga di intestazione),
`{ "labels": [...], "values": [...] }`, `{ "rows": [...], "columns": [...] }` o
una mappa semplice. `x` e `y` scelgono le colonne label e numerica; se omessi si
usano la prima colonna testuale e la prima numerica.

```sui
chart "Trend" data: "Month,Value\nJan,12\nFeb,18\nMar,15" format: "csv" x: "Month" y: "Value"
chart "Rows" data: "<data><row label=\"Jan\" value=\"12\"/><row label=\"Feb\" value=\"18\"/></data>" format: "xml"
```

CSV supporta delimitatore virgola, punto e virgola, tab e pipe, con celle tra
virgolette. L'intestazione viene rilevata da celle testuali sopra celle numeriche.
XML accetta `<row label="..." value="..."/>`,
`<row><label>...</label><value>...</value></row>`, attributi `x`/`y` ed elementi
annidati. I grafici torta/ciambella usano label e valori positivi; DTD ed entità
XML vengono rifiutati.

La heatmap usa la colonna label come x, la prima serie numerica come riga y e la
seconda come valore della cella. Il candlestick richiede le quattro colonne
numeriche nell'ordine `open`, `high`, `low`, `close`.

Il Gantt accetta solo date ISO o date-time e i campi `id`, `task`, `start`,
`end`, `duration`, `progress`, `milestone`, `group`,
`parent`, `dependsOn`, `dependencyType`, `lag`, `constraint`,
`constraintDate`, `deadline`, `resources`, `status`, `color`, `cost`, `actualStart`,
`actualEnd`, `remaining`, `work`, `workDone` e `resourceUnits`. `resources` accetta
un array o una lista separata da virgole/punto e virgola; gli elementi dell'array
possono essere stringhe o oggetti `{ "name": "...", "units": N }`. `resourceUnits`
è una mappa opzionale nome → quantità. L'asse data
sceglie automaticamente tick orari, giornalieri,
settimanali, mensili o annuali; weekend e festivi sono evidenziati e la linea
`Today` appare se la data corrente rientra nell'intervallo. `progress` disegna
la barra di completamento, `milestone` un diamante e `dependsOn` le frecce di
dipendenza. Le linee usano un grigio neutro; i link critici sono più spessi ma
non gialli. Il runtime verifica ogni percorso contro tutte le barre e usa corsie
esterne quando un segmento attraverserebbe una barra. Con
`overflow: "scroll"` i Gantt densi aumentano in altezza e scorrono, mantenendo
spazio sufficiente tra le righe per barre e connessioni leggibili.
I Gantt espongono anche i controlli **timeline zoom** (`−`, `Timeline N%`, `+`,
`Reset`) nell'intestazione del grafico. Lo zoom modifica solo la larghezza
dell'asse temporale; etichette, righe e testo mantengono la stessa dimensione.
Quando la timeline risultante è più larga della cornice, il viewport dedicato
mostra una barra di scorrimento orizzontale. `zoom` imposta la scala iniziale
(0.5–24); `ganttScale` (`auto`, `day`, `week`, `month`, `year`) sceglie
indipendentemente la granularità dei tick. Tick e bande settimanali iniziano di
lunedì e alla scala settimanale le bande dei giorni non lavorativi non vengono
mostrate. Con zoom elevato la scala automatica mostra tick giornalieri con
giorno della settimana ed etichetta dell'anno. La cornice del grafico mantiene la
dimensione assegnata. La colonna sinistra con i nomi ha larghezza fissa e resta
allineata a sinistra anche quando la timeline ridotta è più stretta della
cornice; rimane inoltre fissa durante scroll e zoom. Le etichette lunghe di task
e risorse vanno a capo fino a tre righe nella colonna fissa; anche la colonna
Task della tabella sincronizzata manda a capo il testo. Le risorse multiple sono
mostrate separate da virgole nel grafico/tooltip e come chip in tabella.
`progressColumn: true`
aggiunge una colonna percentuale fissa tra nomi e timeline e nasconde le
percentuali numeriche accanto alle barre. Il bordo destro della colonna dei
nomi è trascinabile per ridimensionarla; la larghezza viene salvata per grafico.
Quando la timeline scorre orizzontalmente, le etichette della scala si fermano
al separatore della colonna nomi invece di scorrere sotto i nomi.
Le deadline sono più grandi e trascinabili orizzontalmente; il campo Deadline
in tabella si
aggiorna al rilascio.
Il routing delle dipendenze valuta più corsie ortogonali, scarta i percorsi che
attraversano barre, evita la sovrapposizione con segmenti orizzontali/verticali
già disegnati e sceglie quello più corto misurato come lunghezza orizzontale +
verticale. I tratti verticali rispettano il `linkGap` configurato rispetto ai
bordi dei task e agli altri tratti verticali; solo il punto condiviso di
uscita/ingresso può coincidere. Tutti i link escono dal bordo destro del task e entrano dal bordo
sinistro del task di destinazione, indipendentemente dalla semantica SS/SF/FF.
L'uscita laterale parte dal centro del lato con un tratto orizzontale di almeno
4px, che può essere condiviso da più collegamenti in uscita.
Un'eccezione vale per un gruppo collassato: se un task interno termina prima
della fine del gruppo, il link esce dal **bordo inferiore della barra di
riepilogo, in corrispondenza della fine del task**, mai dalla fine del gruppo.
Tutti i link in uscita da un task condividono lo stesso punto di uscita e tutti
quelli in ingresso condividono lo stesso punto di entrata; questa coincidenza è
l'unica sovrapposizione ammessa. Un breve ramo separa le uscite; la punta della
freccia è centrata sul bordo sinistro del task, dal retro della freccia parte un
tratto orizzontale di almeno 4px e tutti i link in ingresso si incontrano
all'estremità sinistra di quel tratto. Il routing resta confinato nell'area del
grafico, quindi le linee non passano mai sotto la colonna fissa dei nomi. Le
corsie possono attraversare l'area del grafico, ma ogni tratto verticale rispetta
il `linkGap` rispetto ai bordi dei task e agli altri tratti verticali; l'ingresso
della coda della freccia resta sul bordo sinistro della destinazione. I
tratti verticali sono monotoni nella direzione della riga di destinazione: un
link verso un task più in basso non sale mai, uno verso l'alto non scende mai.
Il primo tratto verso il basso di un'uscita dal bordo inferiore di un gruppo
collassato è esente da questa regola. Se il `linkGap` richiesto necessita di più spazio, lo stage Gantt cresce in
altezza e, quando servono più corsie verticali, anche in larghezza; il viewport
scorre preservando il gap invece di comprimere il percorso. I tratti
orizzontali restano vicini a origine e destinazione invece di
usare subito il bordo estremo del grafico.
Spostare o ridimensionare un task esegue un ridisegno live throttled: i link
vengono ricalcolati, i warning di schedulazione rivalutati e la tabella resta
sincronizzata; la voce di cronologia viene salvata al rilascio.
Ogni ridisegno Gantt valida date, actual, remaining/work, vincoli, deadline e
tipo/lag di ogni dipendenza. Le violazioni appaiono nel banner e sulle righe
della tabella, quindi ogni modifica viene controllata subito. Il drag/resize
porta inizio, fine o milestone a una data lavorativa; scrivere una data non
lavorativa nella tabella viene rifiutato con errore inline e messaggio di
stato. Con il puntatore sulla timeline,
**Ctrl/Cmd + rotella** (o il pinch del trackpad, che il browser segnala come
Ctrl + rotella) modifica lo zoom dell'asse temporale. Sui dispositivi touch è
disponibile il **pinch con due dita**, senza attivare lo zoom della pagina.
`parent` punta all'id di un'altra attività: le attività con figli
diventano summary task WBS con inizio, fine, progresso e costo aggregati dai
discendenti. Questo trasforma il genitore in un contenitore summary. Per
mantenere ogni task come task reale, con la propria barra e i propri link, usa
`group`: è solo un'etichetta di fase e non modifica il parent WBS. I task con lo
stesso `group` condividono un'intestazione di gruppo; dal menu contestuale della
tabella puoi assegnare più righe selezionate con **Assign to group...**.
Seleziona i task con Ctrl/Cmd + click nel grafico, oppure seleziona le righe
della tabella, quindi scegli **Create group from selection**: viene inserito un
nuovo nodo summary chiamato **New group** subito sopra il primo task selezionato,
impostato come parent WBS di tutti i task selezionati; barre e link restano
invariati.
I figli sono indentati a destra del genitore sia nel grafico sia
nella tabella; summary e gruppi usano una banda di intestazione evidenziata e
testo in grassetto. Un summary espanso è una riga di intestazione senza barra; quando
viene collassato, i figli vengono sostituiti da una sola barra di riepilogo e le
dipendenze esterne restano collegate. Per riespanderlo basta cliccare la barra
collassata, il nome, il triangolino o il toggle in tabella. `dependsOn` può
contenere id o nomi di
attività e può elencare più predecessori. Ogni voce può essere una stringa
(`"A"`, `"A[SS+2]"`) o un oggetto (`{"id":"A","type":"SS","lag":2}`), così ogni
link può avere tipo e lag propri; `dependencyType`/`lag` restano il fallback.
Il runtime risolve i nomi in id, rifiuta i cicli durante la modifica e mostra un
errore esplicito per dati importati già ciclici. Con `autoSchedule: true` il runtime ricalcola gli
inizi dal tipo/lag di ogni link e da constraint/data. In modalità data, lag e
spostamenti usano i giorni lavorativi del calendario attivo (weekend, festivi e
chiusure); le durate dei task sono interpretate in giorni lavorativi e le date
di fine vengono ricalcolate. La durata effettiva mostra l'equivalente lavorativo. I cicli vengono segnalati.

Gli attributi calendario del Gantt sono a livello grafico:

- `weekend: "sat,sun"` (predefinito) sceglie i giorni non lavorativi per nome o
  numero (`0` = domenica, `6` = sabato); `weekend: "none"` disattiva le bande
  weekend.
- `timezone: "Europe/Rome"` usa un fuso IANA per tick, date in tabella, tooltip e
  calcolo dei giorni di calendario quando i dati contengono date con orario. Le
  date senza orario restano date di calendario stabili. Il valore predefinito è
  `UTC`.
- `showNonWorkingDays`: default `true`. Mostra o nasconde le bande weekend e
  festivi nella timeline, senza cambiare schedulazione, durata effettiva o
  validazione delle date non lavorative. Alla scala settimanale queste bande
  sono sempre omesse; restano visibili le bande alternate delle settimane.
- `holidays` accetta date esatte (`2026-12-25`), intervalli inclusivi
  (`2026-12-24..2026-12-26`) e date annuali ricorrenti (`*-12-25` o `12-25`).
  Gli intervalli ricorrenti come `*-12-24..*-01-06` attraversano il capodanno.

Il menu del Gantt include **Calendar...**: un overlay dedicato permette di
scegliere i giorni non lavorativi e aggiungere festività o intervalli di
chiusura tramite un unico date-range picker (un click per un singolo giorno, due
per un intervallo). Il calendario è salvato per grafico in `localStorage`
(`sui.chart.calendar.<chart-key>`), può essere fornito nei dati JSON come
`{"tasks": [...], "calendar": {"weekend": "sat,sun", "holidays":
["2026-12-25", "2026-12-24..2026-12-26"]}}` e aggiorna subito bande weekend,
festivi e **Effective duration**. Un handler opzionale `onCalendarChange`
riceve il JSON del calendario salvato. Lo stesso menu del grafico espone
**Show/Hide progress column** e **Show/Hide non-working days** per la sessione.

Le bande verticali alternate seguono la scala attiva: mesi alla scala mensile
(`ganttScale: "month"` oppure auto), settimane con inizio lunedì alla scala
settimanale e giorni di calendario alla scala giornaliera. I colori seguono il
tema dark/light e restano dietro weekend, festivi e barre. Alla scala
giornaliera un giorno non lavorativo usa il colore weekend/festivo invece del
colore alternato dei giorni, quando i giorni non lavorativi sono visibili.

L'aspetto dei link Gantt si configura con attributi a livello grafico:

- `linkColor`: colore CSS delle dipendenze; default neutro `#94a3b8`.
- `linkWidth`: spessore in pixel (`0.5`-`8`), default `1`; i link critici sono
  leggermente più spessi.
- `linkGap`: distanza minima in pixel tra i segmenti di dipendenza e altri link,
  bordi dei task/milestone o punte delle frecce; i tratti verticali rispettano
  sempre questa distanza dalle barre e dagli altri tratti verticali. Default
  `4`, intervallo `0`-`24`.
- `linkStyle`: `auto` (tratteggio per tipo di link), `solid`, `dashed` o
  `dotted`.
- `arrowType`: `triangle`, `open` o `none`.
- `arrowSize`: dimensione della punta in pixel (`2`-`20`).
- `progressColumn`: default `false`; aggiunge una colonna percentuale fissa tra
  nomi dei task e timeline. Quando è visibile, sostituisce le percentuali
  numeriche accanto alle barre.
- `barBorderColor`: colore CSS opzionale per bordo di task, summary e milestone;
  vuoto/omesso mantiene il bordo del tema.

Un esempio completo ed eseguibile è in
[`examples/gantt_calendar/`](../examples/gantt_calendar/README.md): cambia ogni
attributo a runtime e mostra date e bande risultanti.
[`examples/gantt_wbs/`](../examples/gantt_wbs/README.md) mostra summary task
padre/figlio, barre di riepilogo collassate, auto scheduling e cicli di
dipendenza.

Le graffe JSON letterali nel sorgente SUI vanno scritte `\{` e `\}`; l'overlay
dati aggiunge automaticamente gli escape. Per payload lunghi preferire uno stato
stringa.

### Compatibilit? dati/tipo

### Controllo Graph (diagrammi a blocchi)

Il controllo dedicato `graph` disegna un diagramma a blocchi/stati. Usa sempre
la geometria forme/link, quindi non scrive `chartType`. Un graph vuoto è un
contenitore valido; stati e transizioni preimpostati nel sorgente vengono mostrati
di nuovo a ogni caricamento dell'app, mentre le modifiche runtime durano per la
sessione della pagina. Usa `editMode: "none" | "editable" | "persistent"` per
governare l'editing del diagramma: `"none"` è sola visualizzazione, `"editable"`
mantiene le modifiche per la sessione, `"persistent"` le mantiene dopo
reload/riavvio (stati, transizioni, geometria e task Gantt).

```sui
state flow = "\{\"states\":[\{\"id\":\"input\",\"label\":\"Input\"\},\{\"id\":\"process\",\"label\":\"Process\"\},\{\"id\":\"output\",\"label\":\"Output\"\}],\"transitions\":[\{\"from\":\"input\",\"to\":\"process\"\},\{\"from\":\"process\",\"to\":\"output\"\}]\}"

view main {
  graph "Pipeline" name: "pipeline" data: flow
}
```

`graph` accetta gli stessi attributi di dati, layout, griglia, stile tag/link,
tabella e runtime della geometria a blocchi. Usa il tag dedicato `graph`:
`chart "..." chartType: "state"` (e l'alias `statemachine`) viene rifiutato con
una diagnostica. Il menu dei tipi del designer offre solo i tipi
numerici/geometrici.
Quando si cambia un chart verso un tipo con struttura dati incompatibile appare
una conferma e, se accettata, i dati vengono sostituiti con un esempio adatto al
nuovo tipo.

Il JSON principale usa `states` e `transitions`; ogni stato deve avere un `id` esplicito
e univoco all'interno del chart (id mancanti o duplicati bloccano il parsing) e
può avere `label`, `description`, `hint`, `color`, `shape`, proprietà bordo/testo e
`x`/`y`/`w`/`h` opzionali. Un `hint` non vuoto sostituisce il tooltip predefinito;
`"none"`, `"off"` o `"-"` lo nasconde per quel singolo elemento. In `hint` e
`tag`, `|` inizia una nuova riga e `\|` è una pipe letterale (`\\` un backslash
letterale); i campi di modifica sono a riga singola. I ritorni a capo reali sono
normalizzati in `|`.
L'attributo chart `stateHints: false` nasconde tutti gli hint. La colonna **Hint**
della tabella e il menu contestuale **Hint settings...** modificano lo stesso
campo. Il pannello impostazioni del chart ha un pannello autonomo **HINTS** che
stila globalmente tutti i tooltip (attiva/disattiva, sfondo e opacità, colore
testo/bordo, font e dimensione, spessore bordo, padding, raggio) con anteprima dal
vivo. I default sorgente usano `stateHintBg`, `stateHintOpacity`,
`stateHintTextColor`, `stateHintBorderColor`, `stateHintFont`,
`stateHintFontSize`, `stateHintBorderWidth`, `stateHintPadding` e
`stateHintRadius`. `tableView` seleziona `chart`, `table` o `both`; con la tabella visibile,
selezionando un elemento nel diagramma la riga corrispondente riceve focus ed
evidenziazione. Gli id dei link sono opzionali e diventano `edge_1`,
`edge_2`, ... con verifica di non collisione con gli id delle forme. Le funzioni
runtime accettano solo l'`id` esplicito. I valori `type` `initial`, `final`,
`choice` e `normal` sono metadati ma non cambiano la geometria né aggiungono indicatori di
ruolo. I campi forma sono
`shape` (`rectangle`, `square`, `circle`, `stadium`, `ellipse`, `rhombus`,
`triangle`, `parallelogram`, `hexagon`, `trapezoid`, `pentagon`, `note`),
`borderStyle`, `borderWidth`, `borderColor`,
`cornerRadius` (arrotonda anche i vertici dei poligoni come rombo, triangolo,
parallelogramma, esagono, trapezio, pentagono e note), `rotation` in gradi (per le forme con orientamento
significativo; selezionando la forma compare una maniglia di rotazione e il
pannello proprietà espone l'angolo esatto), `tag` e `tagSide`.
Le transizioni hanno `id`, `from`, `to`, `label`, `event`, `guard`, `action`,
`color`, `routing`/`linkType` (`straight`, `arc`, `bezier`, `orthogonal`,
`circular`), `width`, `dashStyle`, `startMarker`, `endMarker`, `fromPort`,
`toPort`, `waypoints` e `tag` opzionali; gli id mancanti vengono generati e
conservati. Nel diagramma è visibile solo il `tag` del link; label, evento, guard
e action restano nella tabella, nell'inspector e nel tooltip. JSON è il formato
completo; il CSV esporta i campi link incluso `hint`. Il tooltip mostra sempre nome,
descrizione e tag anche quando il tag non è disegnato; i tag nel grafico usano
lo stesso tooltip. Nel pannello proprietà Description e Tag / note sono campi a
riga singola; nei tag `|` va a capo e `\|` è una pipe letterale.

Sono accettati anche array di stati con transizioni annidate, edge list CSV e XML
con elementi `state`/`transition`. Le forme possono essere rettangoli, quadrati,
cerchi, stadi con due lati a semicirconferenza, ellissi e rombi; il riempimento,
lo stile/colore/spessore del bordo (sempre un singolo tratto), il raggio degli
angoli e le dimensioni sono proprietà. Nel diagramma non vengono disegnati nomi,
etichette o descrizioni: è visibile solo il `tag`/nota esplicito. Il nome resta
nella tabella, nell'inspector, nel tooltip e nell'etichetta di accessibilità.
Il tipo `triangle` è un triangolo isoscele con la punta in alto; gli anchor sono
calcolati sui tre lati e ai vertici la direzione di uscita è la bisettrice
dell'angolo. `parallelogram` (input/output), `hexagon` (preparazione),
`trapezoid` (operazione manuale), `pentagon` (connettore off-page) e `note`
(angolo ripiegato in alto a destra) sono forme poligonali. I loro anchor sono
distribuiti in modo uniforme su ogni lato (vertici più punti intermedi regolari)
e ruotano insieme alla forma, evitando addensamenti irregolari sui lati
obliqui.

La lavagna del diagramma è almeno grande quanto il viewport visibile, quindi la
griglia e il canvas utilizzabile riempiono tutta la vista. La griglia è disegnata
sul viewport e segue scorrimento e pan, quindi resta continua anche quando la
vista va oltre la lavagna. Allargando il viewport
la lavagna si ingrandisce automaticamente; si espande inoltre quando forme, tag o
waypoint escono dal bordo corrente. Il contenuto mantiene 20 px di margine.
**Fit width** adatta il contenuto in orizzontale lasciando lo scorrimento
verticale, mentre **Fit all** adatta entrambe le dimensioni e centra il grafo;
nessuno dei due taglia il contenuto.

La griglia usa `stateGridGap` come lato della cella e come distanza minima tra
linee; `stateGridColor` ne imposta il colore. È opzionale: si mostra o nasconde
con `stateGrid: false` o dal pannello unico **GRID** (visibilità, dimensione e
colore); senza griglia resta il canvas uniforme. Il gap predefinito è `24` px.
Anche i punti di attacco e le maniglie di modifica sono opzionali: con
`stateHandles: false` (o dal pannello **Diagram settings**) non vengono disegnati
ancore, grip di ridimensionamento/rotazione, maniglie di tag e di percorso e il
pulsante flottante **+ Add**, e le forme non si trascinano: il diagramma è solo da
visualizzare. Le maniglie sono attive di default; impostare `stateHandles: false`
per nasconderle.
I centri dei cerchi e tutti i vertici
delle forme vengono agganciati ai vertici della griglia; rettangoli/quadrati/
stadi hanno box allineato e dimensioni multiple del passo, rombo/ellisse hanno
centro ed estremi allineati, anche dopo il resize. Ogni segmento orizzontale o
verticale interno dei link ortogonali giace su una linea della griglia ed è lungo
almeno `stateGridGap`; il primo/ultimo stub di attacco può partire da un punto
esatto (anche fuori griglia) del bordo curvo. I punti di aggancio sono calcolati
sul bordo reale della forma: per cerchi, ellissi, lati del rombo e estremità
dello stadio una coordinata viene dalla griglia e l'altra è risolta sulla curva,
quindi il link parte esattamente sulla circonferenza e non su
un'approssimazione arrotondata. Sono nominati `V<col>T/B`, `H<row>L/R`
(fallback `P:<side>`) e visibili solo con l'elemento selezionato e le maniglie
di modifica attive.
I link non-self partono/arrivano su un punto di aggancio; i tratti
ortogonali escono e entrano perpendicolari per almeno `stateGridGap` pixel.
Quando l'anchor si trova su una superficie curva o obliqua (cerchio, ellisse,
estremità dello stadio o lato del rombo), il primo/ultimo tratto segue la normale
reale alla superficie: può essere diagonale, non viene forzato sulla griglia
ortogonale ed è escluso dalla regola dei segmenti ortogonali. Il resto della
spezzata resta ortogonale. I self-loop sono normali link tra due punti di attacco
**distinti** della stessa forma (mai lo stesso punto). Si creano dal canvas (in
modalità link fare doppio clic sulla stessa forma) e dalla tabella. Vale anche qui
il subtype scelto: `straight`, `arc`, `circular` (arco maggiore circolare attorno
alla forma, ridimensionabile con la maniglia del raggio), `orthogonal` (loop
rettangolare attorno alla forma) o `curved` (Bézier che esce/entra lungo le normali
degli anchor). Le due estremità sono i veri anchor e si spostano in modo
indipendente con le maniglie. Non serve una configurazione dedicata per
lato/raggio/gap. Il layout
automatico usa livelli e non amplifica i cicli; `x`/`y` espliciti consentono il
posizionamento manuale.

Le maniglie di estremità possono essere rilasciate su **qualsiasi** forma: trascinando
l'inizio o la fine su un'altra forma il link viene riagganciato a quella forma. Le
proprietà del link includono **Direction**: scegliendo il verso inverso vengono
scambiati sorgente/destinazione, anchor e porte, i waypoint manuali vengono invertiti
e la posizione del tag specchiata, così la freccia segue il nuovo target. Il pannello
**LINK** rispecchia quello del singolo link con due schede: **Line** (subtype,
colori linea/evidenziazione, spessore, tipo di tratteggio con lunghezza/gap, raggio
arco circolare e percorsi modificabili) e **Arrows** (dimensione freccia, marker
iniziale/finale e anteprima grafica). Sono i default ereditati da ogni link; i
marker iniziale/finale del singolo link possono sovrascriverli (**default** eredita
l'impostazione del grafico).

Se il router non riesce a trovare un percorso pulito (un link attraverserebbe una
forma, percorso vuoto/acuto/fuori griglia) il diagramma resta invariato, la modifica
viene annullata e un avviso non bloccante segnala il problema. Gli avvisi di qualità
(segmenti corti, distanza dagli altri link) restano in
`data-sui-chart-route-audit` (`failures`, `hard`) e non riscrivono mai ancore,
waypoint o percorsi in automatico.

I tag hanno una configurazione indipendente per forme e link (`stateShapeTag*`
e `stateLinkTag*`). Ogni elemento può sovrascrivere i default con `tagKind`,
`tagBox`, `tagAlign`, `tagLeader`, `tagArrow`, `tagArrowSize`, `tagFont`,
`tagFontSize` e un punto di attacco (`tagAnchor` per le forme, `tagAnchorT` per i
link):
- i tag **esterni** stanno fuori dall'elemento e possono essere collegati con una
  linea retta o una curva Bézier (`tagLeader`: `line`, `bezier`) con freccia
  opzionale (`tagArrow`). Gli stili sono `plain` (senza bordo), `rectangle`,
  `rounded`, `square`, `circle`, `ellipse` e `line` (testo sopra una linea);
- i tag **interni** delle forme sono centrati nella forma, non ruotano con essa e
  restano selezionabili/modificabili;
- i tag **interni** dei link interrompono la linea con una maschera di sfondo,
  usano gli stessi stili e possono seguire la direzione locale della linea o
  restare orizzontali (`tagAlign`: `line`, `horizontal`);
- il punto arancione mostrato sull'elemento selezionato sposta il contatto del
  leader/linea. Per le forme è memorizzato come punto normalizzato sul bordo
  (`tagAnchor: {rx, ry}`) e sopravvive a resize e rotazione; per i link è una
  frazione del percorso (`tagAnchorT`);
- il leader è sempre allineato alla normale uscente nel punto di contatto: parte
  dal bordo reale del box scelto (lato del rettangolo, circonferenza di
  cerchio/ellisse) oppure dalla linea che sottolinea il testo per lo stile
  `line`, e termina esattamente sul perimetro della forma o sulla linea del link.
  Il box viene orientato in modo che il bordo affacciato sia perpendicolare al
  leader (sulle forme oblique il contenuto del tag ruota con la normale; i box
  circolari mantengono il testo orizzontale perché ogni leader radiale è
  perpendicolare). La punta della freccia coincide con quel punto di contatto e
  gli anchor delle forme provengono dal bordo reale ruotato, quindi un triangolo
  o un poligono ruotato riceve la punta esattamente sul lato. Il tag può essere
  trascinato ovunque. Un leader retto usa il numero minimo di segmenti: uno
  quando il tag giace sulla normale dell'elemento (è anche la posizione
  predefinita di un tag non ancora posizionato), altrimenti tre (uscita
  perpendicolare dall'elemento, un connettore, arrivo perpendicolare al box). Un
  leader Bézier viene spezzato in due tratti cubici raccordati con tangenti
  finali perpendicolari al box e all'elemento;
- `tagRotation` ruota l'intero tag (box e testo). Se è `null`/assente il tag
  segue l'orientamento automatico; un angolo manuale seleziona il lato del box
  affacciato all'elemento e il leader viene re-instradato mantenendo la
  perpendicolarità ai due contatti. I default di famiglia sono
  `stateShapeTagRotation` e `stateLinkTagRotation`; i pannelli TAG per elemento e
  per famiglia espongono **Tag rotation** (Auto/Manual) e **Rotation (deg)**.
  Selezionando un elemento compare anche una maniglia di rotazione sul tag:
  trascinandola il tag ruota attorno al centro, passa in manuale e `Shift`
  aggancia a 15°;
- aggancio automatico dei tag esterni: le forme usano il punto centrale del lato
  più alto; cerchio/ellisse usano il punto più alto; un poligono obliquo usa il
  punto centrale del lato più orizzontale tra i due superiori, oppure il vertice
  alto se le inclinazioni coincidono. I link usano il punto medio del segmento
  orizzontale più alto; un link verticale usa il punto medio con il tag a lato.
  Il testo parte a due altezze di riga lungo la normale uscente e si allontana
  solo se rischia di sovrapporsi a un'altra forma, link o tag. L'anchor automatico
  viene scelto una sola volta e memorizzato (normalizzato per le forme, frazione
  del percorso per i link), quindi i ridisegni successivi mantengono lo stesso
  punto relativo e la lavagna non oscilla;
- le polilinee dei leader dei tag non hanno mai angoli interni inferiori a 90°:
  i candidati con pieghe acute vengono scartati e il router aggiunge un détour
  ortogonale finché ogni svolta è almeno 90°. I leader Bézier mantengono i bracci
  di controllo entro il 35% della corda e sono spezzati in due tratti raccordati,
  quindi non possono formare pieghe più acute di 90°;
- le frecce sono proporzionali al gap di griglia: lunghezza e base valgono
  `stateGridGap / 4`, quindi con gap 20 px la freccia è 5 px. `arrowSize`
  sovrascrive le frecce dei link tra forme, `stateShapeTagArrowSize` /
  `stateLinkTagArrowSize` quelle delle famiglie di tag e ogni elemento può usare
  `tagArrowSize`. Tutti questi valori si modificano dal pannello TAG;
- il sotto-pannello **TAGS** nella tab **General** delle impostazioni apre un unico pannello per le famiglie
  forme/link (tipo predefinito, box, direzione del testo, leader, dimensione
  freccia, font e tag spostabili) con anteprima dal vivo. Il pulsante **Tag
  settings…** nelle proprietà di forma/link (anche nel menu contestuale) apre il
  pannello del singolo elemento; i vecchi sottomenu dei tag sono stati rimossi.
  I controlli sono organizzati in tab (**Tag**, **Text**, **Box** e
  **Leader & arrow**), quindi l'altezza del pannello segue il tab più grande: le
  opzioni on/off sono switch, i valori numerici usano uno slider con campo
  numerico modificabile e il font è un combo box ricercabile. Il campo font
  mostra il font corrente o quello predefinito e la lista contiene sempre tutte
  le famiglie disponibili, inclusa quella selezionata. Applicando il pannello del
  singolo elemento vengono salvati solo i campi modificati, quindi le proprietà
  non toccate continuano a ereditare i default della famiglia. Il tab **Text** espone
  anche il testo del tag multilinea, dimensione, peso, stile, decorazione,
  allineamento, spaziatura tra lettere, interlinea, trasformazione, ombra del
  testo e colore dell'ombra. Il comando **Text settings…** in ogni riga della
  tabella e nel menu contestuale di ogni elemento apre direttamente questo
  pannello sul tab **Text**. I controlli di rotazione sono nel tab **Tag**. I colori usano un campione compatto con campo hex e lo
  sfondo del box offre anche un pulsante **Transparent**. Il tab **Box** contiene
  stile, sfondo, colore bordo, spessore bordo (minimo 1 px), stile bordo e padding
  tra testo e bordo. **No border** mantiene comunque selezionabili stile del box,
  sfondo e padding e disegna un box senza bordo con lo sfondo scelto; solo colore,
  spessore e stile del bordo vengono disabilitati. Lo switch **Arrow** abilita o
  disabilita forma e dimensione della freccia. **Movable tags** è nel tab **Tag**:
  con un tag interno di forma il tag si muove solo dentro la forma, con un tag
  interno di link solo lungo il percorso del link. Nell'anteprima il tag si
  trascina e la maniglia rotonda lo ruota: la rotazione aggancia a
  0/90/180/270° entro 6° e viene riportata nel campo Rotation, mentre la posizione
  trascinata resta locale all'anteprima. **Text direction** è attivo solo con
  rotazione manuale: **Always horizontal** mantiene il testo orizzontale e
  disabilita il valore di rotazione, mentre **Follow line** usa l'angolo manuale.
  Un tag interno di forma usa per default sfondo trasparente e nessun tag box e ha
  i controlli leader/freccia disabilitati; un tag interno di link resta sulla
  linea, la interrompe visivamente e non sta mai sopra o sotto. Il link di esempio
  è sempre una linea
  retta e usa colore, spessore, tratteggio e punta uguali al grafico. Il leader usa
  le stesse funzioni polilinea/Bézier, le stesse frecce e le stesse normali del
  grafico;

La tabella usa due schede, **Shapes** e **Links**. Shapes usa `Element`, `Tag`,
`Type` e azioni; Links usa `Element`, `Tag`, `Type`, `From`, `To` e azioni.
Element è la prima colonna e inizia con il color picker; `Type` indica la
geometria o il tipo di linea. Grafico e tabelle sono due viste dello stesso
modello: aggiungere, modificare, spostare, ridimensionare, cambiare il tag o
eliminare in una vista aggiorna l'elemento condiviso e si riflette subito
nell'altra. Eliminando una forma vengono eliminati anche i link collegati.
Selezionando un elemento nel grafico si apre
automaticamente la scheda corrispondente e la riga viene portata nell'area
visibile della tabella. L'area del grafico e quella delle tabelle sono separate
da 10 px. La barra delle schede Shapes/Links resta fissa; scorrono solo le
righe dei dati, con le intestazioni sticky. Il pulsante **Add row** è nella
barra delle schede, allineato a destra, e apre la scelta Shape/Link. La tabella
non ha pulsanti propri di impostazioni o undo/redo: import/export sono in **Graph
settings** e i pulsanti undo/redo del grafico coprono diagramma e tabella.
Passando il mouse su una cella appare un tooltip con tutto il testo; la combo
Type è esclusa. I `|` vengono mostrati su righe separate, come nel diagramma.
Event/guard/action restano nel pannello
proprietà del link. Il pulsante compatto **+** nell'intestazione del grafico apre un unico
menu con le sette forme e **New link**, quindi non copre i nodi. Tooltip e
`aria-keyshortcuts` indicano `+`/`N`; `L` crea direttamente un link.
La creazione di un link è un gesto a due click: il primo sceglie il punto di
contatto di partenza, il secondo quello di arrivo. Al passaggio del mouse su un
elemento vengono evidenziati tutti i suoi punti di contatto e il più vicino è
il candidato; cliccando in un punto qualsiasi dell'elemento viene scelto
l'anchor più vicino. I due anchor scelti vengono salvati come
`fromAnchor`/`toAnchor` e vincolano il disegno del link, anche quando il router
deve ripiegare su port allineati agli assi.
Undo/redo sono icone con hint accessibili nell'intestazione del grafico e nella
toolbar della tabella; le scorciatoie sono `Ctrl+Z` e `Ctrl+U`. Il pulsante `…`
apre il pannello
proprietà per forma/link e per i tag individuali. Ogni cella Element inizia con
un color picker che aggiorna subito forma o linea. Le proprietà della forma sono a
schede: **Shape** (impostazioni bordo a sinistra: bordi, stile, spessore, corner
radius e, per bordi dashed/dotted, lunghezza/gap del tratteggio; subtype e i tre
colori fill/bordo/evidenziazione nella colonna destra; anteprima della forma in
basso) e **Position** (rotazione, X/Y, larghezza e altezza su due colonne allineate
in alto). La label si modifica dalla tabella e lo stile del testo dai Tag settings;
la vecchia sezione Text e il campo State role sono stati rimossi. Le
proprietà del link sono organizzate in schede: **Line** su due colonne (entrambi i
selettori di colore impilati nella stessa colonna; subtype, spessore con slider +
campo numerico, tipo di tratteggio, lunghezza/gap dei tratti e pulsante direzione
che inverte il link quando premuto), **Arrows** (dimensione freccia e marker
iniziale/finale, ciascuno con esempio grafico) e **Anchors** (nomi e coordinate
degli anchor di partenza/arrivo, in sola lettura). Il campo label resta nella
tabella; i vecchi campi description, event, guard, action e tag/nota sono stati
rimossi. Lo stile del link eredita la configurazione
generale **LINK**: un nuovo link non salva campi di stile; modificando un campo
nel pannello del singolo link viene salvato un override solo per quel link
(**default**/**auto** torna a ereditare) e **Reset style to defaults** cancella
gli override. Spostamento e ridimensionamento agganciano
sempre la forma alla griglia e salvano X/Y/larghezza/altezza arrotondati.
Il ridimensionamento mantiene fisso l'angolo superiore sinistro; cerchio e
quadrato mantengono le proporzioni e la maniglia resta delle stesse dimensioni
a schermo a ogni livello di zoom. Le forme possono essere ridotte fino a una
cella della griglia; il diametro di un cerchio può scendere fino a un
`stateGridGap` (il testo dell'etichetta non impone più un minimo maggiore,
perché i nomi non vengono disegnati sulla forma). Ridimensionando una forma
ruotata la maniglia resta raggiungibile: il canvas scorre automaticamente per
mantenerla nella vista.
Spostando/ridimensionando una forma o cambiando
`stateGridGap`, gli anchor vengono ricalcolati e i link senza ancoraggio valido sono
riancorati prima del salvataggio. Selezionando dal grafico, la scheda e la riga
corrispondenti vengono aperte e portate nell'area visibile. La selezione è
sincronizzata grafico/tabella e usa il colore ambra
`--sui-chart-state-selection`; **Ctrl/Cmd/Shift+click** aggiunge o rimuove
elementi dalla selezione. Il **trascinamento con il tasto sinistro** sullo spazio
vuoto disegna il rettangolo di selezione, mentre il **trascinamento con il tasto
destro** sposta l'intero grafico; **Alt+trascinamento** e il tasto centrale
spostano anch'essi la lavagna.
Un click sullo spazio vuoto azzera l'intera selezione. Un gruppo selezionato si
sposta come un'unica unità mantenendo invariate le interdistanze; la traslazione
finale viene arrotondata a un multiplo intero di `stateGridGap` su entrambi gli
assi. Se le forme vengono trascinate oltre il viewport la lavagna si espande
invece di bloccarle. **Ctrl+C** copia, **Ctrl+V**
incolla e **Ctrl+D** (o **Duplicate** nel menu contestuale) duplica la selezione;
duplicare una forma non duplica i link collegati se non sono selezionati. Con due
o più forme selezionate compare la toolbar di allineamento
sinistra/centro/destra, alto/centro/basso e distribuzione orizzontale/verticale,
con guide magnetiche blu durante il trascinamento. Le frecce spostano tutte le
forme selezionate; Shift+frecce ridimensionano solo una singola forma
(Left/Right larghezza, Up/Down altezza) mantenendo ancorato l'angolo superiore
sinistro. Delete/Backspace elimina tutte le forme e i link selezionati.
`editMode: "editable"` mantiene l'editor; `"persistent"` salva anche geometria,
stile e attributi runtime tra reload/riavvio; `editMode: "none"` rende il grafo
una superficie di presentazione: niente selezione, drag/resize o modifica di
route/tag, mentre pan e zoom restano disponibili. Tooltip e coda eventi restano
attivi.
Le interazioni non editabili vengono accodate dal chart e consumate con
`chart.pollEvents` / `chart.nextEvent`. Ogni evento contiene `event`, `kind`,
`part`, `id`, `label`, `tag`, `from`, `to`, coordinate e modificatori; il
codice SUI decide quale funzione esterna chiamare. Usare `chart.clearEvents` e
`chart.eventCount` per gestire la lista. `chart.focusRow` evidenzia, porta in
vista e focalizza una riga tabella (anche non-editable). I callback diretti non
sono disponibili: il chart è solo una sorgente di eventi e il codice applicativo
decide cosa fare con ciascun evento.

`stateRouting` sceglie `curved`, `arc`, `circular`, `orthogonal` o `straight`;
i link ortogonali usano sempre il router a percorso minimo che evita gli ostacoli
(con ritentativi automatici sulle corsie esterne), quindi non c'è un algoritmo da
scegliere; `stateGridGap` imposta sia la cella della griglia sia la distanza
minima dei tracciati ortogonali (default 24, range 0-64). Ogni percorso ortogonale
viene verificato: i segmenti devono stare sulle linee della griglia e mantenere
`stateGridGap` dalle altre forme e dai link gia instradati. Sono esclusi solo lo
stub normale curvo/obliquo e il suo breve raccordo; i waypoint manuali sono
autoritativi anche se la ricerca ostacoli non
trova un corridoio. Segmenti consecutivi non formano mai un angolo interno
inferiore a 90 gradi: se uno stub normale diagonale creerebbe un angolo acuto,
il router usa il port allineato all'asse più vicino; un waypoint manuale che
richiederebbe un'inversione viene raggiunto con un dogleg rettangolare. Ogni
segmento orizzontale o verticale è lungo almeno `stateGridGap`; un candidato con
tratti più corti viene ricostruito con port allineati agli assi. I diagrammi di
stato non disegnano assi, tacche o etichette di scala lungo il bordo del
grafico: sono visibili solo i tag degli elementi. Un
candidato non valido viene ritentato con corsie esterne sempre
piu ampie; il risultato è esposto in `__suiStateRouteAudit` e in
`data-sui-chart-route-audit` (`gap`, `checked`, `failures`, `minNode`, `minLink`).
Le impostazioni del diagramma sono salvate per chart-key in localStorage
(`sui.chart.state.settings.<chart-key>`): **Apply** salva il grafico corrente,
**Save as default** imposta anche il fallback globale, **Reset settings** ripristina
i valori del sorgente `.sui`; gli stessi default restano disponibili come attributi.
Il pannello modifica anche il **titolo del grafico** (`data-sui-chart-title`) e i
colori generali di **bordo forme**, **link** ed **evidenziazione**; `borderColor`,
`color` e `highlightColor` per singolo elemento hanno precedenza e
l'evidenziazione colora anche la riga selezionata in tabella.
Le nuove forme usano una palette pastello tenue. Le maniglie di rotazione
agganciano al valore più vicino tra 0/90/180/270° entro 6°, facilitando
l'allineamento orizzontale o verticale di testo e forme.
**Tag spostabili** (`stateTagsMovable`, anche nel menu del grafico e in Diagram
settings) rende trascinabili tutti i tag/note, compresi quelli dei link. L'offset
viene applicato all'intero gruppo del tag (testo, sfondo e linea guida) e salvato
come `tagOffset` nel payload, quindi sopravvive a reload, undo/redo e commit.
Se disattivato, i tag restano nella posizione automatica.
**Percorsi modificabili** (`stateRoutesEditable`, anche nel menu e in Diagram
settings) mostra le maniglie di modifica su ogni link. Il trascinamento modella il
percorso in base al sottotipo: `orthogonal` crea/sposta waypoint agganciati alla
griglia e ricostruisce il percorso, `straight` crea una polilinea retta attraverso
i waypoint, una Bézier `curved` espone **due maniglie indipendenti** per i punti di
controllo (`bend1X`/`bend1Y`, `bend2X`/`bend2Y`), `arc` espone il suo punto di
controllo (`bendX`/`bendY`), `circular` espone il raggio (`arcRadius`) e gli
estremi restano sempre trascinabili su un altro anchor o su un'altra forma;
trascinando la maniglia del raggio oltre la corda l'arco si ribalta dall'altro lato
(`arcSweep`) e oltre il semicerchio diventa un arco maggiore (`arcLarge`), così il
loop continua a crescere invece di bloccarsi. I self-link usano un arco maggiore di
default; la maniglia del raggio segue il puntatore e sta sopra le forme, quindi
resta raggiungibile anche sui nodi circolari. Il link selezionato
evidenzia tutto il tracciato, incluse le punte delle frecce. Durante il
trascinamento di una forma o di una maniglia la lavagna resta fissa (le coordinate
non saltano) e al rilascio l'elemento spostato viene riportato in vista. Se
disattivato, i percorsi restano automatici.
I tag hanno stile e link separati per forme e link (`stateShapeTagStyle`,
`stateLinkTagStyle`, `stateShapeTagLink`, `stateLinkTagLink`): gli stili sono
`plain` (senza bordo), `rectangle`, `rounded`, `square`, `circle`, `ellipse`,
`line` (testo sopra una linea centrata), `line-start` e `line-end` (il leader si
aggancia all'inizio o alla fine della linea sotto il testo); i collegamenti sono
`none`, `line`, `bezier`, `arrow` o `bezier-arrow`. `stateShapeTagKind`/
`stateLinkTagKind` impostano il tipo predefinito (`external`/`internal`) e
`stateShapeTagAlign`/`stateLinkTagAlign` selezionano `line` o `horizontal`.
`stateShapeTagArrowSize`/`stateLinkTagArrowSize` impostano la dimensione freccia
della famiglia (0 o assente = `stateGridGap / 4`), come `arrowSize` per i link
del diagramma; `dashLength`/`dashGap` impostano il tratteggio del singolo link,
mentre i default del grafico sono `data-sui-chart-dash-length`/`dash-gap` e
`data-sui-chart-state-start-marker`. `stateShapeTagArrowShape`/`stateLinkTagArrowShape` scelgono
`triangle`, `open`, `circle` o `square`, mentre
`stateShapeTagTextColor`/`stateLinkTagTextColor`,
`stateShapeTagBoxColor`/`stateLinkTagBoxColor` e
`stateShapeTagLeaderColor`/`stateLinkTagLeaderColor` colorano testo, box e leader
in modo indipendente (i fallback comuni sono `stateTagTextColor`,
`stateTagBoxColor`, `stateTagLeaderColor` e `stateTagArrowShape`).
`stateShapeTagBorderColor`/`stateLinkTagBorderColor`,
`stateShapeTagBorderWidth`/`stateLinkTagBorderWidth`,
`stateShapeTagBorderStyle`/`stateLinkTagBorderStyle` e
`stateShapeTagPadding`/`stateLinkTagPadding` definiscono colore, spessore, stile
del bordo del tag box e padding tra testo e bordo (`solid`, `dashed`, `dotted`).
Il link usa `linkStyle` (`auto`/`solid`/`dashed`/`dotted`) e `linkWidth` quando
impostati dal sorgente o dalle proprietà del link. La tipografia del tag è
configurabile per famiglia con `stateShapeTagFont`/`stateLinkTagFont` e varianti
`FontSize`, `FontWeight`, `FontStyle`, `TextDecoration`, `TextAlign`,
`LetterSpacing`, `LineSpacing`, `TextTransform`, `TextShadow` e
`TextShadowColor` (le varianti
generali `stateTagFont*` sono fallback). Le
sovrascritture per elemento nel payload (`tagKind`, `tagBox`, `tagAlign`,
`tagLeader`, `tagArrow`, `tagArrowShape`, `tagArrowSize`, `tagFont`,
`tagFontSize`, `tagFontWeight`, `tagFontStyle`, `tagTextDecoration`,
`tagTextAlign`, `tagLetterSpacing`, `tagLineSpacing`, `tagTextTransform`,
`tagTextShadow`, `tagTextShadowColor`, `tagTextColor`, `tagBoxColor`, `tagBorderColor`,
`tagBorderWidth`, `tagBorderStyle`, `tagPadding`, `tagLeaderColor`, `tagAnchor`,
`tagAnchorT`) hanno precedenza. Font/dimensione dedicati (`stateShapeTagFont`,
`stateLinkTagFont` e varianti `FontSize`) completano la configurazione, mentre
`stateTagFont`/`stateTagFontSize` restano i fallback comuni. A runtime
`chart.setTag(chart, id, text, color, textColor)` aggiorna un tag,
`chart.update(chart, id, patch)` applica una patch e
`chart.updateMany(chart, updates)` aggiorna più elementi con un solo redraw; gli
id sono solo gli `id` espliciti di forma/link; prefissi, label e `from->to` sono
rifiutati. Gli helper tipizzati (`chart.setText`, `chart.setTagColor`,
`chart.setHint`, `chart.setColor`, ...) coprono i singoli attributi; le patch
accettano gli alias `text`, `textColor`, `tagColor`/`backgroundColor`. Vedi
[`STDLIB_IT.md`](STDLIB_IT.md#chart-module-chart-core). Le forme misurano il
testo e non possono
essere più piccole dell'etichetta; nodi e link si spostano/ridimensionano e il
percorso dei link si modifica con waypoint/maniglie. Selezionando un link
compaiono le maniglie iniziale/finale: sono **route-aware**, cioè durante il
trascinamento valutano gli anchor vicini e scelgono quello più vicino al
puntatore il cui percorso non risulti sensibilmente più lungo o complesso
(lunghezza, numero di pieghe, segmenti corti e angoli acuti). Tenendo premuto
**Alt** si forza l'anchor più vicino; i nomi anchor restano in sola lettura
nelle proprietà.
I vertici manuali del percorso sono trascinabili sul link selezionato: ogni
vertice si aggancia alla griglia e viene conservato durante la modifica (solo i
punti consecutivi identici vengono uniti). Si possono afferrare anche le pieghe
generate automaticamente: il primo trascinamento le converte in waypoint manuali.
**Add waypoint here** inserisce un vertice trascinabile agganciato alla griglia
nel punto cliccato; trascina la maniglia per piegare il percorso e usa **Clear
manual waypoints** per rimuoverli. Lo state diagram usa un canvas dinamico: la lavagna è l'area
grafica effettivamente usata più 20 px per lato, si estende quando forme, tag o
waypoint escono dal bordo e mostra le scrollbar quando serve. **Fit** misura il
contenuto renderizzato e non taglia mai il grafo.
Lo zoom si controlla con **Ctrl/Cmd + rotella** o pinch a due dita (5%–400%),
mantenendo stabile il centro del puntatore/pinch; i controlli mostrano la
percentuale. Lo zoom scala l'intero disegno: passo della griglia, geometria di
forme e link, testo e tag. **Fit** e **All** adattano il contenuto renderizzato
alla finestra visibile. Trascinando lo spazio libero con il pulsante principale
si sposta l'intero diagramma (forme, link e tag insieme);
partendo da forma, link, tag o maniglia resta il normale comportamento di
modifica. Il router ortogonale usa
A* attorno a forme, bordi e linee già tracciate; i tag evitano forme, altri tag e
segmenti. Con i tag spostabili attivi il trascinamento muove testo e sfondo del
tag, mentre la linea guida e la freccia restano agganciate alla forma o al link.
Il pulsante a ingranaggio nell'intestazione apre un pannello **Settings** a tab:
**General** contiene titolo del grafico, colore bordo forme, colore di
evidenziazione, tema chiaro/scuro e switch **Show anchors and handles**, oltre a
una riga di pulsanti **GRID**, **SHAPES**, **LINKS** e **TAGS**, salvataggio come default e reset; **Versions**
contiene modifiche pendenti, commit e cronologia; **Data** contiene import dei
dati di stato ed export JSON/CSV/SVG/PNG. Il grafo mantiene solo diagramma,
tabella forme/link e controlli di aggiunta. Alla chiusura di un sotto-pannello
GRID/SHAPES/LINKS/TAGS ricompare il pannello General. SHAPES
imposta i default per le nuove forme (stile/spessore bordo, lunghezza/gap del
tratteggio, corner radius, colore bordo ed evidenziazione, più lo switch **Random
fill color** o un colore fisso) con anteprima; i controlli sono disposti su due
colonne allineate con righe di pari altezza. Le proprietà del singolo
elemento hanno la precedenza. LINK
raccoglie aspetto, routing, percorsi modificabili e frecce; GRID raccoglie
visibilità, dimensione e colore della griglia. **Export JSON** include forme,
link e impostazioni complete, quindi reimportando il file il diagramma viene
ricostruito con geometria, tag, percorsi, stili e impostazioni identici. La toolbar della
tabella e l'intestazione offrono aggiunta/eliminazione, undo/redo (`Ctrl+Z` /
`Ctrl+U`), import ed export JSON/CSV.
Le modifiche sono autosalvate in `sui.chart.state.<chart-key>` e usano 50 passi
di undo/redo (spostamenti, resize, rotazioni, trascinamento tag, waypoint,
campi, aggiunte e cancellazioni). **Commit changes…** nella tab **Versions** delle
impostazioni crea una versione immutabile con audit in
`sui.chart.state.versions.<chart-key>`; **Version history…** è nella stessa tab.
Il ripristino carica lo snapshot senza creare una nuova versione. `onCommit` e l'evento `sui:state-commit` ricevono `{ id,
revision, baseRevision, at, author, message, chart, kind: "state", payload,
changes }`; il controllo non esegue rete, l'host gestisce autenticazione,
concorrenza e conflitti `409`. Vedi
[`examples/state_machine/`](../examples/state_machine/README.md).

Il chart valida i dati analizzati rispetto a `chartType` a runtime, e l'overlay
dati mostra lo stesso controllo prima di Apply:

- i tipi a serie singola accettano una sola serie numerica;
- `barstacked`, `bargrouped` e `combo` richiedono almeno due serie;
- `radar` accetta una o più serie e almeno tre categorie;
- `heatmap` richiede esattamente due serie: riga y e valore;
- `candlestick` richiede esattamente quattro serie: open, high, low, close;
- `treemap` richiede una serie positiva;
- `boxplot` richiede cinque serie: min, q1, median, q3, max;
- `bullet` richiede due o tre serie: value, target, max opzionale;
- `dumbbell` e `gantt` richiedono esattamente due serie;
- `state` richiede almeno uno stato o una transizione e usa lo schema
  stati/transizioni;
- `gauge` usa il primo valore e `min`/`max` opzionali (default 0-100);
- `pie`/`donut`/`funnel`/`treemap` richiedono almeno un valore positivo;
- `histogram` richiede valori numerici e accetta `bins` (1-100);
- `waterfall` accetta delta positivi e negativi.

Se il dato contiene pi? colonne numeriche, scegliere un tipo multi-serie oppure
impostare `y` su una sola colonna. Il messaggio di errore indica il tipo scelto e
la differenza. L'overlay disabilita Apply finché dati e tipo non sono compatibili;
il runtime rifiuta di disegnare una combinazione incompatibile.

### Dimensioni, scroll e zoom

Il chart non ha parametri width/height separati: usare la dimensione normale del
componente (`width` e `height` nella sezione **Size** dell'inspector, oppure CSS
in `style expr`). Il grafico generato riempie quella cornice. La dimensione
predefinita è 640 x 320 px.

- `overflow: "scroll"` (default) disegna alla dimensione naturale e mostra
  scrollbar nella viewport quando il disegno supera la cornice.
- `overflow: "zoom"` mantiene il disegno nella cornice e mostra i controlli
  `-`, `+`, `Reset` con la percentuale. Intervallo 50%-400% a passi del 25%;
  ingrandendo compaiono le scrollbar.
- `overflow: "fit"` comprime il disegno nella cornice senza controlli.
- `pointWidth` riserva spazio orizzontale per punto in modalità scroll (default
  48). `zoom` è lo zoom iniziale della modalità zoom (default 1).
- `pointSize` è il raggio dei punti dati in pixel (default 3, intervallo 1-24).
  Vale per line, step, smooth, area, scatter, combo, radar e dumbbell; gli altri
  tipi lo ignorano.

`showGrid`, `showLegend` e `showValues` sono booleani. Anche `pointWidth`,
`pointSize` e `zoom` accettano espressioni reattive.

`tableView` è disponibile per ogni grafico: `"chart"` (default), `"table"` o
`"both"`. La tabella è
nativa e accessibile. Il Gantt usa Task, WBS parent, Start, End, Deadline, Actual
start/end, Duration, Effective duration, Remaining, Progress, Work, Work done,
Milestone, Status, Group, Resources, Depends on, Link, Lag, Constraint,
Constraint date, Slack e Cost. La cella Resources mostra chip rimovibili e una
quantità `×N` di persone per assegnazione. Un piccolo pulsante **+** apre un
pannello con i tipi di risorsa della scheda Resources e un campo quantità; la
conferma aggiunge l'assegnazione. Per i Gantt il menu del grafico
espone anche **Resource workload...**, un riepilogo per risorsa di task assegnati
ed effort. Il carico usa il valore `work` del task quando presente, altrimenti la
durata in giorni lavorativi; un task assegnato a più risorse conta per intero su
ciascuna. La tabella ha due schede: **Tasks** e **Resources**. La barra delle
schede resta fissa in alto durante lo scroll. La scheda Resources memorizza un
registro con Name, Role, Max people, Daily cost e Notes. Il nome della risorsa è
la chiave usata dalle assegnazioni: rinominarla aggiorna tutti i task, mentre
l'eliminazione è rifiutata finché resta assegnata. Una risorsa digitata nella
scheda Tasks viene aggiunta automaticamente al registro. Un oggetto JSON Gantt
può fornire anche il registro iniziale:

```json
{
  "tasks": [ { "task": "Design", "start": "2027-03-01", "end": "2027-03-05", "resources": [ { "name": "Marta", "units": 2 } ] } ],
  "resources": [ { "name": "Marta", "role": "Design", "maxPeople": 3, "dailyCost": 450, "notes": "Part time" } ]
}
```

Ogni assegnazione risorsa ha una quantità: usa `{ "name": "...", "units": N }`
dentro `resources`, oppure un oggetto separato `resourceUnits` come
`{ "Marta": 2 }`. `units` vale `1` per impostazione predefinita e si modifica
direttamente nel chip della risorsa nella tabella Tasks. `maxPeople` è il numero
massimo di persone disponibili per la risorsa: la somma delle `units` di tutti i
task contemporanei assegnati non deve superarlo. Le modifiche interattive di
date, risorse, quantità e Max people, gli incolla e le variazioni che creano un
nuovo sovraccarico vengono rifiutati e ripristinati; i sovraccarichi già presenti
mostrano comunque un avviso. `0` o un `maxPeople` vuoto significa senza limite.
`dailyCost` è il costo di una persona per un giorno. Quando tutte le risorse
assegnate a un task hanno un costo giornaliero, la cella **Cost** del task
diventa di sola lettura e viene calcolata come
`somma(dailyCost × units) × durata in giorni lavorativi`; altrimenti resta il
costo manuale. La modifica di una tariffa, di una quantità o del calendario
ricalcola i costi dei task coinvolti.
**Deadline** è un vincolo
soft: il task viene comunque schedulato, ma una fine pianificata successiva alla
deadline viene evidenziata nel grafico e in tabella; il combo Constraint e il suo
filtro elencano sempre tutti i sette tipi (`ASAP`, `SNET`, `SNLT`, `FNET`,
`FNLT`, `MSO`, `MFO`) con etichette esplicative, anche quando i dati usano solo
`ASAP`. Duration è la differenza di calendario `end - start`; **Effective
duration** è un conteggio read-only dei giorni lavorativi nello stesso periodo,
calcolato con `weekend`, `holidays` e `timezone`, quindi utile per effort e costo
(per dati numerici mostra `-`). **Actual start/end** registrano l'esecuzione.
**Remaining** è modificabile: se vuoto viene derivato da durata e progresso in
giorni lavorativi. **Work** e **Work done** abilitano il progresso basato
sull'effort: quando entrambi sono presenti, Progress deriva da
`workDone / work`; modificare Progress aggiorna Work done, mentre modificare
Work done ricalcola Progress. Il grafico mostra il lavoro effettivo come barra
sottile superiore. **Depends on** mostra i nomi
dei task come chip rimovibili e offre un combo con tutti i nomi per aggiungere
uno o più predecessori; un click sul chip apre l'editor del tipo e del lag del
singolo link. Ogni barra interattiva ha anche un piccolo connettore sotto il
bordo destro: trascinalo su un'altra barra o sul nome del task per creare una
dipendenza FS, poi usa il chip per modificarne tipo e lag. Una dipendenza che
creerebbe un ciclo viene rifiutata con un messaggio nella tabella. Collassando
un summary si nascondono anche le righe
dei discendenti, così tabella e timeline mostrano gli stessi task. I task sono resi come blocchi
contigui sotto il rispettivo `group` sia nel grafico sia nella tabella, anche
quando l'ordine interno dei task è diverso; una nuova riga eredita il gruppo
dell'ultimo task e viene mostrata dentro quel gruppo. Lo stesso vale per i
`group` collassati: la tabella mostra una barra di toggle dei gruppi per
espanderli o collassarli senza usare il grafico.
gli altri tipi espongono Label più le serie, oppure colonne specifiche come
Open/High/Low/Close, Min/Q1/Median/Q3/Max e X/Y/Value per la heatmap.
`"table"` nasconde il grafico; `"both"` mostra grafico e tabella insieme. In
modalità `"both"` la tabella è agganciata al bordo inferiore con altezza fissa
per ogni tipo di grafico: il grafico si riduce, un separatore trascinabile sopra
la tabella ne imposta l'altezza (salvata per chart), scorre solo il corpo mentre
intestazione e pulsanti restano fissi, e ridimensionando il contenitore
l'altezza della tabella resta costante mentre il grafico prende lo spazio extra.
Il separatore è raggiungibile da tastiera: Su/Giù ridimensionano di 8 px
(Shift+Su/Giù di 32 px), PageUp/PageDown del 10% dell'altezza del chart, Home
riporta a 80 px ed End usa l'altezza massima. Ogni cella di intestazione contiene un pulsante filtro che apre un popup in stile
Excel, indipendente per colonna. Il popup combina ricerca testuale, elenco dei
valori (fino a 40 valori distinti) e, per colonne numeriche o date ISO, campi
minimo/massimo. Un click sull'intestazione (non sul filtro) ordina in modo
crescente o decrescente; Shift+click aggiunge un ordinamento multi-colonna.
Trascinando le intestazioni si riordinano le colonne; trascinando il bordo
destro di un'intestazione si ridimensiona la colonna (48-900 px), oppure si
sposta il focus sulla maniglia e si usano Left/Right. Un doppio click sulla
maniglia adatta la colonna al contenuto più largo. Le larghezze restano salvate
con lo stato della tabella e **Reset column widths** ripristina il
dimensionamento automatico. Righe e intestazioni sono raggiungibili da tastiera
e la prima colonna identificativa resta fissa durante lo scroll orizzontale. Il menu **Settings** della tabella ha i sottomenu Columns,
Filters ed Export; Export scrive le righe visibili in CSV, TSV, JSON o JSONL,
copia le righe visibili come TSV oppure apre una vista di stampa.
`tableFilters: false` rimuove i pulsanti filtro di colonna e il sottomenu
Filters. Filtri, ordinamento, ordine, larghezze e visibilità delle colonne restano in
`localStorage` per quel grafico. La didascalia appare in modalità solo tabella ed
è nascosta quando grafico e tabella sono entrambi visibili.

`tableAddRow: true` (predefinito `false`) mostra il pulsante **+ Row** su ogni
tabella. Nel Gantt aggiunge un task con id univoco, intervallo di un giorno dopo
l'ultimo task, gruppo ereditato e nessuna dipendenza. Negli altri tipi aggiunge
una riga con valori zero ai dati locali del grafico, cancella i filtri attivi,
ridisegna grafico e tabella e chiama `onDataChange` se configurato. Le tabelle
derivate come l'istogramma ricalcolano i bin dal valore aggiunto. Senza
`onDataChange`, la riga non Gantt resta locale fino al successivo rerender dei
dati sorgente.

`tableSelection` sceglie l'unità di selezione del corpo: `"row"` (predefinito),
`"column"`, `"cell"` o `"none"`. La gutter con i numeri di riga seleziona e
trascina le righe; una striscia sopra ogni intestazione seleziona e trascina le
colonne; trascinando nel corpo si disegna un rettangolo di celle. Ctrl/Cmd+click
attiva/disattiva e Shift+click estende dall'ancora. Alt+click su un'intestazione
ordina invece di selezionare. `tableDeleteRow: true` aggiunge un pulsante
visibile per eliminare la riga in fondo alla riga. `tableSteppers: true`
nasconde gli spinner nativi e mette due pulsanti affiancati meno/più sul bordo
destro delle celle numeriche; compaiono solo dopo un click che porta il focus
sulla cella (o via focus da tastiera), non al solo passaggio del mouse, e sono
nascosti mentre è attivo un trascinamento di selezione di righe, colonne o celle.
Senza questa
opzione le celle numeriche non hanno controlli spinner. Le selezioni di righe, colonne e
celle usano lo stesso colore `--sui-chart-table-selection-bg`, con un perimetro
da 2px `--sui-chart-table-selection-border` attorno al blocco selezionato. In
modalità cella viene riempita solo la cella selezionata; la riga resta collegata
alla selezione del grafico senza colorare l'intera riga.
Il focus da hover usa tonalità neutre attenuate tramite
`--sui-chart-table-focus-row-bg` e `--sui-chart-table-focus-cell-bg`; clic e
trascinamento rimuovono il focus hover e lasciano solo la selezione. Il focus
da tastiera è un bordo da 1px `--sui-chart-table-focus-color` (predefinito
`#64748b`, con variante più chiara per il tema chiaro). Righe e celle selezionate
mantengono il colore di selezione anche sotto il mouse. Iniziare una selezione o
cliccare un'altra cella annulla la modifica in sospeso e rimuove il relativo
focus/highlight.
`tableContextMenu: true` abilita il menu contestuale con copia, incolla,
eliminazione delle righe selezionate ed export CSV/TSV/JSON/JSONL della
selezione. Copia scrive TSV negli appunti; Incolla richiede il permesso di
lettura e riempie celle, righe o colonne selezionate. Le tabelle Gantt e grafico
validano i valori incollati e chiamano `onTaskChange` / `onDataChange`.

L'applicazione richiede al browser il permesso di scrittura/lettura degli appunti
durante copia e incolla e mostra il risultato in una status line `aria-live`;
se l'API non è disponibile usa un textarea nascosto per la copia.

Righe, colonne e celle supportano la selezione multipla. Il menu **Rows** può
selezionare tutte le celle visibili nella modalità corrente, cancellare la
selezione, copiare o incollare la selezione come TSV per Excel/Word ed eliminare
le righe selezionate. **Export selection** scrive CSV, TSV, JSON, JSONL, XLSX o
HTML per righe, colonne o celle.

Le **Export options** della tabella controllano riga di intestazione, colonne
nascoste e ordine originale o corrente delle righe. **Export visible rows** usa
l'ordine e i filtri correnti; **Export selection** limita righe e colonne alla
selezione attiva. XLSX e HTML conservano stile delle intestazioni e allineamento
numerico; i valori simili a formule restano testo inerte in CSV/TSV/XLSX, quindi
non possono essere eseguiti come formule. **Print table...** stampa la tabella
filtrata; **Print / Save PDF (chart + table)...** apre la stampa del browser con
l'SVG corrente del grafico e la tabella filtrata, da cui salvare un PDF.

Tutte le tabelle chart modificabili e le tabelle Gantt mantengono una cronologia
undo/redo in memoria di 50 passi per modifiche alle celle, stepper numerici,
incolla, inserimento ed eliminazione di righe. Usa **Undo**/**Redo** dalla toolbar
della tabella, dal menu del grafico Gantt o dal menu contestuale della tabella, oppure `Ctrl+Z` e
`Ctrl+Shift+Z` / `Ctrl+Y`. Con un input focalizzato resta attivo l'undo nativo
del browser. Le modifiche alle celle e gli stepper aggiornano il grafico e la
tabella in place; quando `onDataChange` aggiorna lo stato reattivo e la view
viene ridisegnata, il DOM di grafico e tabella viene preservato per evitare
flicker e ricostruzioni. I bin dell'istogramma sono derivati, quindi read-only e
non eliminabili singolarmente. Le tabelle oltre 300 righe usano lo scrolling virtuale
a finestra, mentre filtri ed export continuano a coprire tutte le righe filtrate,
non solo quelle visibili.

Con `editMode: "editable"` o `"persistent"` le celle Gantt Task, WBS parent, Start, End, Progress, Group e
Depends on sono modificabili, e anche le celle Label/X/serie non Gantt sono
modificabili in place (i bin dell'istogramma restano read-only). Le modifiche
non Gantt aggiornano i dati locali e chiamano `onDataChange`. Il pulsante
**+ Row** aggiunge un task con id
univoco, intervallo predefinito di un giorno dopo l'ultimo task, gruppo
dell'ultimo task e nessuna dipendenza. La nuova riga viene selezionata e
focusata per la modifica immediata; l'aggiunta cancella i filtri attivi,
ridisegna la timeline e chiama `onTaskChange`. Una modifica valida ridisegna il
grafico e chiama `onTaskChange`, mentre drag/resize aggiornano la tabella. La
selezione è sincronizzata nei due sensi e le modifiche fanno lampeggiare
brevemente celle e attività Gantt.

La selezione è sincronizzata per ogni coppia grafico/tabella: in un Gantt,
selezionando una barra, il nome del task a sinistra, la riga o la gutter con il
numero di riga si evidenziano tutti e tre; per gli
altri tipi, selezionando un elemento si seleziona la riga, selezionando una riga
si evidenziano tutti gli elementi di quella label; nei grafici multi-serie la
riga evidenzia tutte le serie.
La selezione è salvata in `localStorage` e ripristinata dopo il reload; anche
filtri, ordinamento, ordine e visibilità delle colonne usano uno stato persistente.

Click destro sul grafico, oppure il pulsante **⋮**, apre il menu del grafico:
tutti i formati di export (SVG, PNG, CSV, TSV, JSON e JSONL) sono raggruppati
nel sottomenu **Export**. L'export dal grafico contiene tutti i dati; l'export da Settings
della tabella contiene righe e colonne visibili.

Tooltip, hover e selezione sono disponibili su ogni grafico. `criticalPath: true`
evidenzia le attività Gantt con slack zero e le relative dipendenze. Il calcolo
usa tipo/lag di ogni link, vincoli, durata in giorni lavorativi e calendario
weekend/festivi attivo;
`holidays: "*-11-16, 2026-11-20..2026-11-21"` colora festivi esatti, intervalli
e ricorrenze annuali. `weekend` e `timezone` configurano il calendario usato per
bande, tick e date in tabella.
`editMode: "editable"` o `"persistent"` abilita anche drag/resize Gantt, gruppi
collassabili e celle Gantt modificabili.
Le modifiche Gantt si annullano/ripristinano con `Ctrl+Z`, `Ctrl+Shift+Z` /
`Ctrl+Y` o dal menu del grafico.
L'intera riga attività è anche un target da tastiera: `ArrowLeft`/`ArrowRight`
sposta di un giorno lavorativo, `Shift+Arrow` ridimensiona la fine,
`Alt+Arrow` ridimensiona l'inizio e `Enter` apre il menu attività. Il
connettore delle dipendenze è focalizzabile: `Enter` apre il selettore della
destinazione; una dipendenza focalizzata supporta `Enter`/`Shift+F10` per la
modifica e `Delete` per la rimozione. Le righe di
intestazione dei gruppi mostrano triangolo e numero di attività; un click
collassa il gruppo in una barra di riepilogo. La barra copre l'intervallo dei
task figli, usa il progresso ponderato per durata e il colore del gruppo; le
dipendenze esterne restano collegate alla barra, quelle interne sono nascoste.
I tooltip sono disponibili su ogni tipo di
grafico e su ogni elemento dati: barre, punti di linea/area/step/smooth, fette
pie/donut, punti radar, arco gauge, stadi funnel, celle heatmap, corpi candlestick,
celle treemap, box plot, barre bullet, punti dumbbell e attività Gantt. Il Gantt può
chiamare un handler quando un'attività cambia:

```sui
chart "Plan"
	data: ganttData
	chartType: "gantt"
	editMode: "persistent"
	onTaskChange: ganttChanged
```

L'handler riceve una stringa JSON con le attività correnti (`id`, `task`,
`start`, `end`, `progress`, `milestone`, `group`, `dependsOn`). L'handler è
opzionale: ogni modifica viene salvata automaticamente in una bozza locale sotto
`sui.chart.gantt.draft.<chart-key>`, anche dopo un reload. La bozza non è una
versione committata.

L'intestazione Gantt ha un pulsante **Commit**. Raggruppa le modifiche fatte
dall'ultimo commit in una versione locale immutabile e in una voce di audit,
invece di creare cronologia a ogni piccola modifica. È disabilitato quando non
ci sono modifiche; il menu del grafico espone anche **Commit changes...** e
**Version history...**. Versioni e dettagli di audit sono in
`sui.chart.gantt.versions.<chart-key>`; ripristinare una versione carica il suo
snapshot di attività e risorse senza creare un'altra versione. Usa `onCommit`
per inviare la versione a un'applicazione o a un server:

```sui
chart "Plan"
	data: ganttData
	chartType: "gantt"
	editMode: "persistent"
	onCommit: ganttCommitted
```

Il callback e l'evento DOM `sui:gantt-commit` ricevono `{ id, revision,
baseRevision, at, author, message, chart, payload, resources, changes }`.
`baseRevision` supporta la concorrenza ottimistica multiutente: l'host può
inviare il payload al proprio endpoint con `If-Match`/`baseRevision` e
ricaricare o riconciliare in caso di `409`. Il controllo non esegue I/O di rete
autonomo; autenticazione, autorizzazione e server condiviso sono forniti
dall'host. `onResourceChange` resta disponibile per replicare anche il registro
risorse nel modello applicativo.

Il registro risorse viene salvato in `localStorage` nella chiave
`sui.chart.resources.<chart-key>`. Aggiungi `onResourceChange` per salvarlo anche
nell'applicazione ospite:

```sui
chart "Plan"
	data: ganttData
	chartType: "gantt"
	editMode: "persistent"
	onResourceChange: resourcesChanged
```

L'handler riceve un array JSON di `{ name, role, maxPeople, dailyCost, notes }`.
Il pulsante **Reset saved** nella scheda Resources cancella il registro locale e
lo ricostruisce dal sorgente del chart e dai nomi usati nei task.

L'inserimento di righe non Gantt può notificare l'app con `onDataChange`:

```sui
chart "Revenue"
	data: chartData
	chartType: "bar"
	tableView: "both"
	tableAddRow: true
	onDataChange: chartDataChanged
```

L'handler riceve un array JSON di oggetti con `label`, `x` e una proprietà per
ogni serie. Assegnandolo allo stato sorgente la nuova riga sopravvive ai
successivi rerender.

### Colori e preset

Le variabili CSS supportate via `style expr` includono
`--sui-chart-bg`, `--sui-chart-fg`, `--sui-chart-title-color`,
`--sui-chart-muted`, `--sui-chart-grid`, `--sui-chart-axis`,
`--sui-chart-series-1` ... `--sui-chart-series-8`, `--sui-chart-border`,
`--sui-chart-radius`, `--sui-chart-padding`, `--sui-chart-title-size`,
`--sui-chart-label-size`, `--sui-chart-value-size`, `--sui-chart-line-width`,
`--sui-chart-bar-opacity`, `--sui-chart-area-opacity`,
`--sui-chart-slice-stroke`, `--sui-chart-scroll-thumb`,
`--sui-chart-scroll-track`, `--sui-chart-control-bg`,
`--sui-chart-control-fg`, `--sui-chart-control-border`,
`--sui-chart-control-muted`, `--sui-chart-control-hover-bg`,
`--sui-chart-control-hover-fg`, `--sui-chart-control-hover-border`,
`--sui-chart-control-radius`, `--sui-chart-legend-color`,
`--sui-chart-legend-bg`, `--sui-chart-legend-size`,
`--sui-chart-tooltip-bg`, `--sui-chart-tooltip-fg`,
`--sui-chart-tooltip-border`, `--sui-chart-tooltip-radius`,
`--sui-chart-tooltip-max-width`, `--sui-chart-tooltip-font-size`,
`--sui-chart-tooltip-delay`, `--sui-chart-table-font-size`,
`--sui-chart-table-cell-padding`, `--sui-chart-table-head-bg`,
`--sui-chart-table-stripe-opacity`, `--sui-chart-table-caption-display` e
`--sui-chart-table-header-position`.

Il designer offre 14 preset: **Default**, **Ocean**, **Forest**, **Sunset**,
**Neon**, **Mono**, **Pastel**, **High contrast**, **Amber**, **Candy**,
**Slate**, **Light**, **Outline**, **Minimal** e **Custom**. Un preset riscrive `style expr`
con le variabili; modificare l'espressione per colori personalizzati. Pulsanti
zoom, legenda, etichette assi e scrollbar usano `--sui-chart-control-*`,
`--sui-chart-legend-*` e le variabili scrollbar con fallback ai colori
foreground/axis/muted del preset, quindi restano leggibili con ogni preset.
Dentro `cssSafe` usare colori esadecimali a 8 cifre invece di `rgba(...)`.

### Accessibilità e sicurezza

La radice del grafico è un gruppo etichettato e l'SVG ha `role="img"` con
`aria-label` dal titolo. Un grafico è un riepilogo visivo: se i numeri devono
essere letti con precisione, affiancare una tabella.

`data` è testo non fidato. Il parsing ha un limite di 5.000 punti e un limite di
testo; formati sconosciuti o righe non valide producono un messaggio visibile.
I nodi SVG sono creati con `createElementNS` e le etichette con `textContent`,
quindi il payload non può iniettare HTML o script. DTD/entità XML sono rifiutati.
Vedere [`SECURITY_IT.md`](../SECURITY_IT.md#3c-dati-chart-inerti).

## Date/Time picker

`datepicker` usa l'input nativo del browser per data, ora o data/ora. Supporta
label visibile, binding di stato, clear, intervallo min/max e stati
required/disabled.

```sui
state when = ""

datepicker "Appuntamento"
	mode: "datetime"
	value: when
	min: "2026-01-01T08:00"
	max: "2026-12-31T20:00"
	step: 60
	onChange: appointmentChanged
	required: true
	disabled: false
```

- `mode`: `"date"` (default), `"time"` o `"datetime"`. L'input generato usa
  `type=date`, `type=time` o `type=datetime-local`.
- `value`: collega uno stato stringa. Il formato salvato è simile a ISO:
  - data: `YYYY-MM-DD`
  - ora: `HH:MM` (oppure `HH:MM:SS`)
  - data/ora: `YYYY-MM-DDTHH:MM` (oppure con secondi)
- `min` / `max`: limiti nativi opzionali nello stesso formato della modalità.
- `step`: passo opzionale in secondi per time/datetime (passato all'input
  nativo).
- `onChange`: funzione senza argomenti opzionale, chiamata dopo il setter dello
  stato.
- `required` e `disabled`: passati all'input nativo. Il clear è disabilitato
  quando il controllo è disabilitato.
- Il pulsante clear imposta valore e stato a `""` e chiama `onChange`.
- Un valore dello stato non valido o fuori intervallo mostra un errore inline con
  `role="alert"` e imposta `aria-invalid`; il picker nativo mantiene la propria
  validazione.
- La label è associata all'input e il pulsante clear ha un nome accessibile.

### Calendario personalizzato

`variant` seleziona l'interfaccia: `native` (predefinito) mantiene il selettore
data/ora del browser; `calendar` mostra un calendario SimpleUI per i modi `date`
e `datetime`. Nel modo `datetime` il calendario include anche un campo orario.
Il calendario supporta navigazione tra mesi, spostamento dei giorni con le
frecce, Home / End, Page Up / Page Down ed Escape. I giorni rispettano `min` e
`max` e il valore ISO collegato allo stato non cambia. Nel menu Struttura del
designer sono disponibili **Native picker**, **SimpleUI calendar**, **Calendar
circles** e **Calendar tiles**. Il modo `time` resta un input orario nativo.

Con `variant: "calendar"` il campo data è modificabile: digitare le cifre
inserisce automaticamente `/`, spazi e `:`. Le date usano `DD/MM/YYYY`; i
valori data/ora usano `DD/MM/YYYY HH:MM` (i secondi sono accettati se inclusi).
È possibile incollare un valore ISO. Il valore associato allo stato resta ISO
(`YYYY-MM-DD` oppure `YYYY-MM-DDTHH:MM[:SS]`). Gli input incompleti o non validi
mostrano un errore inline e non vengono mantenuti nello stato. Fare clic sul
pulsante calendario, oppure premere Invio o Freccia giù con il campo attivo,
per aprire il calendario.

Il comportamento del selettore nativo è gestito dal browser. Il calendario
personalizzato non aggiunge capability di rete, file o comandi. Validare valori
malformati o fuori intervallo prima di assegnare dati remoti allo stato
collegato.

## Pagine e navigazione

Ogni pagina è una `view` root non invocata. L'app generata mostra inizialmente
la prima pagina; le altre restano nel DOM ma sono nascoste. Usare il builtin
`navigate(pageName)` da un handler per mostrare un'altra pagina:

```sui
fn openSettings() -> null {
	navigate("settings")
}

fn backHome() -> null {
	navigate("main")
}

view main {
	title "Home"
	h1 "Home"
	button "Settings" onClick: openSettings
}

view settings {
	title "Settings"
	h1 "Settings"
	button "Back" onClick: backHome
}
```

`navigate` accetta una stringa con il nome della pagina, restituisce `null`, non
richiede import ed è un no-op nelle app a pagina singola o con un nome
sconosciuto. Lo stato è globale: i valori impostati in una pagina restano
disponibili nelle altre.

Nel **designer** i chip pagina sopra il preview sono il selettore: fare clic su
un chip per modificare quella pagina; **Insert → New page…** aggiunge pagine.
La sezione **View** dell'inspector modifica titolo/nome della pagina attiva e
offre duplicate/delete/save-as-template. La navigazione nell'app generata è
separata: aggiungere pulsanti o voci di menu con handler che chiamano `navigate`.

Vedere l'[esempio pages](../examples/pages/README.md).

## Internazionalizzazione (i18n)

I cataloghi di traduzione sono dichiarazioni di primo livello. La prima lingua
dichiarata è il fallback:

```sui
i18n "it" {
	"save": "Salva"
	"welcome": "Ciao, {name}!"
	"items": { one: "{count} elemento", other: "{count} elementi" }
}
i18n "en" {
	"save": "Save"
	"welcome": "Hello, {name}!"
	"items": { one: "{count} item", other: "{count} items" }
}
```

Un valore può essere una stringa o un blocco plurale. I blocchi plurali usano le
categorie CLDR `zero`, `one`, `two`, `few`, `many` e `other`; `other` è
obbligatoria. I segnaposto come `{name}` e `{count}` restano letterali e vengono
sostituiti a runtime.

```sui
state keySave = "save"
state player = "Anna"
state count = 1

fn welcomeText() -> string { return t("welcome", {"name": player}) }
fn itemsText(n: int) -> string { return t("items", {"count": n}) }

view main {
	p "{t(keySave)}"
	p "{welcomeText()}"
	p "{itemsText(count)}"
	button "English" onClick: setEn
}

fn setEn() -> null { i18n.setLocale("en") }
```

`t(key, vars?)` è disponibile anche come `i18n.t` e non richiede import. Le
stringhe letterali dentro un'interpolazione sono supportate, quindi funziona
anche la forma diretta `p "{t("save")}"`. La lingua attiva viene scelta in
quest'ordine: parametro `?locale=`, chiave
`localStorage` `sui.locale`, lingua del browser, prima lingua dichiarata.
`i18n.setLocale(locale)` cambia e salva la scelta; `"auto"` riesegue il rilevamento.
Le chiavi mancanti passano al catalogo di fallback, poi alla chiave stessa.

| Chiamata | Risultato |
|---|---|
| `i18n.locale()` | Codice della lingua attiva. |
| `i18n.setLocale(code)` | Cambia lingua; restituisce `true` se il codice è caricato. |
| `i18n.detect()` | Migliore corrispondenza caricata per le lingue del browser. |
| `i18n.available()` | Array dei codici lingua caricati. |
| `i18n.dir()` | `"ltr"` o `"rtl"`; `setLocale` aggiorna anche `lang`/`dir` del documento. |
| `i18n.has(key)` | Se la chiave esiste nel catalogo attivo o di fallback. |
| `i18n.missing()` | Chiavi presenti nel fallback ma assenti nella lingua attiva. |

`Intl.PluralRules` sceglie la categoria plurale. I valori dei segnaposto sono
inseriti come testo e passano dall'escaping normale delle view: traduzioni e
variabili non possono iniettare HTML. Nel designer **Preferenze → Lingue...**
permette di modificare i cataloghi, importare/esportare JSON e scegliere la
lingua dell'anteprima. Vedi [Editor GUI: lingue](GUI_EDITOR_IT.md#lingue) e
`examples/i18n/README.md`.

## Formattazione locale (Intl)

Questi builtin formattano date, numeri, valute, percentuali e tempi relativi con
le API `Intl` del browser. Una locale vuota o omessa usa la locale i18n attiva,
poi la lingua del browser.

| Chiamata | Risultato |
|---|---|
| `intl.date(value, style?, locale?, timeZone?)` | Data; `value` è una stringa ISO o millisecondi epoch. `style` è `short`, `medium`, `long`, `full` o `iso`. |
| `intl.time(value, style?, locale?, timeZone?)` | Solo orario. |
| `intl.datetime(value, style?, locale?, timeZone?)` | Data e orario (`iso` restituisce il timestamp ISO completo). |
| `intl.number(value, decimals?, locale?)` | Numero decimale con separatori; `decimals = -1` usa il default della locale. |
| `intl.currency(value, currency, locale?)` | Formattazione valuta con codice ISO 4217 (`EUR`, `USD`, ...). |
| `intl.percent(value, decimals?, locale?)` | Percentuale da una frazione (`0.256` diventa `25.6%`). |
| `intl.relative(offset, unit, locale?, numeric?)` | Tempo relativo; unità `second(s)`, `minute(s)`, `hour(s)`, `day(s)`, `week(s)`, `month(s)`, `quarter(s)`, `year(s)`. `numeric` è `auto` (default) o `always`. |
| `intl.timezone()` | Il fuso orario IANA risolto, per esempio `Europe/Rome`. |

```sui
p "{intl.date("2024-01-02T15:04:05Z", "short", "en-US", "UTC")}"
p "{intl.number(1234.56, 2, "en-US")}"
p "{intl.currency(1234.5, "EUR", "de-DE")}"
p "{intl.relative(-2, "day")}"
```

Le date non valide restituiscono `Invalid date`; i numeri non validi
restituiscono una stringa vuota. Tutti i risultati sono testo normale e passano
dall'escaping standard delle view.

## Dialoghi modali

Un `modal` è un dialogo nascosto con titolo, contenuto e un footer di pulsanti
`modalbutton`. Un `name` lo identifica nel runtime:

```sui
modal "Edit name" name: "edit-name" {
	p "Update the name."
	input "Name" value: name
	modalbutton "Cancel"
	modalbutton "Save" variant: "primary" onClick: saveName
}
```

Aprire e chiudere da un handler:

```sui
fn openName() -> null {
	modal.open("edit-name")
}

fn closeName() -> null {
	modal.close("edit-name")
}

button "Edit name" onClick: openName
```

| Elemento | Testo | Attributi | Figli |
|---|---|---|---|
| `modal "Titolo"` | Titolo/etichetta accessibile | `name`, `class`, `style` | contenuto + `modalbutton` |
| `modalbutton "Etichetta"` | Etichetta del pulsante | `variant`, `close`, `onClick` | no |

`modal.open(name)` e `modal.close(name)` sono builtin, non richiedono import e
restituiscono `null`. I nomi sconosciuti sono ignorati. Aprire un altro modale
chiude quello corrente. Il dialogo si chiude con click sul backdrop, `Escape` o
un pulsante modale predefinito. Il focus entra nel dialogo; l'elemento che lo ha
aperto viene ripristinato quando è ancora disponibile dopo un re-render. Il
runtime compensa la scrollbar nascosta e ripristina il focus con
`preventScroll`, quindi aprire o chiudere un dialogo non sposta né scorre la
pagina sottostante (incluso un chart statico).

`modalbutton` chiude il dialogo dopo l'handler `onClick` per impostazione
predefinita. Usare `close: false` se l'azione deve lasciarlo aperto. `variant`
può essere `primary`, `danger` o vuoto.

Nel designer usare **Overlay → Modal**. Parte con titolo, contenuto e pulsanti
Cancel/OK. Fare clic con il tasto destro sulla riga `modal` per **Add modal
button**, oppure su una riga `modalbutton` per **Add button below**,
**Edit label**, **Duplicate**, **Move up/down** e **Delete**. La sezione **Modal**
dell'inspector modifica il `name` stabile; selezionando un pulsante si modificano
variant e comportamento di chiusura. Vedere l'[esempio
modal](../examples/modal.sui).

## Table layout

`tablelayout` è un contenitore **strutturale**, non il controllo dati `table`.
Genera una vera `<table>` HTML con una griglia esplicita di righe, colonne, celle
unite, dimensioni e bordi:

```sui
tablelayout rows: 3 columns: 2 width: "100%" columnWidths: "55%, 45%"
  gapX: "18px" gapY: "14px"
  borderStyle: "dashed" borderWidth: "1px" borderColor: "#475569" {
  card "Header" colspan: 2
  card "Actions"
  col {
    input "Name" value: name
    button "Save" onClick: save
  }
  p "Ogni figlio diretto occupa una cella."
}
```

- `rows` va da `1` a `50` e `columns` da `1` a `12`. Se `rows` è omesso viene
  ricavato dai figli; le righe esplicite sono un minimo perché figli aggiuntivi
  espandono la griglia.
- Ogni figlio diretto occupa la successiva cella libera, riga per riga.
- `colspan` e `rowspan` su un figlio diretto uniscono la cella verso destra o
  verso il basso. Gli span vengono limitati alle righe/colonne rimanenti.
- `width` / `height` impostano la dimensione della tabella (`"80%"`, `"640px"`).
- `columnWidths` e `rowHeights` accettano misure separate da virgola in `px`,
  `%`, `em`/`rem`, `vw`/`vh`, `pt` o `auto`.
- `gapX` / `gapY` impostano lo spazio tra colonne e righe (default `12px`).
- `borderStyle` (`none`, `solid`, `dashed`, `dotted`, `double`),
  `borderWidth` e `borderColor` stilano ogni cella del layout.
- `cellMinHeight` imposta l'altezza visibile delle celle vuote (default `72px`).
- Attributi per cella: `cellPadding`, `cellBackground` e
  `cellBorderStyle`/`cellBorderWidth`/`cellBorderColor`.
- Per allineare il contenuto, applica lo stile all'elemento figlio (sezione
  Alignment o CSS), non alla cella.
- `responsive: true` impila le celle sotto i 640px di larghezza; le celle
  allineate diventano flex-column quando impilate e mantengono l'allineamento.
- `semantic: true` usa `role="table"` invece di `role="presentation"`; `label`
  aggiunge un `aria-label` accessibile.
- I figli mantengono comportamento e stile normali. Usa `col`/`div`/`card` per
  tenere più elementi nella stessa cella.
- Non ha `name`, `editMode`, API dati, ordinamento, filtri o persistenza. Per le
  griglie dati usa il controllo `table` e le funzioni `table.*`.

Le tabelle generate dal designer usano celle esplicite:

```sui
tablelayout rows: 2 columns: 2 {
  layoutcell row: 0 col: 0 colspan: 2 { card "Header" }
  layoutcell row: 1 col: 0 { card "Left" }
  layoutcell row: 1 col: 1
}
```

`layoutcell` accetta `row`, `col`, `rowspan`, `colspan` e gli attributi di stile
per cella indicati sopra. I figli diretti senza `layoutcell` mantengono il
comportamento automatico precedente.

Nel designer **Split vertical** / **Split horizontal** dividono a metà le celle
selezionate, aggiungono i confini globali necessari e aggiornano i contatori.
Ctrl+click aggiunge celle alla selezione; **Merge selected** unisce le celle solo
se le prime due condividono un lato intero, e ogni cella successiva deve
continuare nella stessa direzione. Uno split che renderebbe una colonna più
stretta di 40px o una riga più bassa di 24px chiede conferma. I contatori di
righe/colonne aggiungono o
rimuovono celle sul bordo destro/inferiore; riducendoli viene chiesto se
rimuovere o spostare il contenuto delle celle coinvolte. Trascina le sottili
maniglie dei separatori nella preview per ridimensionare colonne o righe:
**Structure table settings** modifica le misure in pixel risultanti. Le celle
vuote si selezionano con un click normale, quelle occupate con Alt+click o con
il triangolo in basso a sinistra sempre visibile; cliccando un figlio si
seleziona il figlio. Il triangolo verde appena fuori dall'angolo in alto a
sinistra seleziona l'intera tabella; un click dentro la prima cella seleziona la
cella o il suo contenuto.
Nel designer è in **Layout > Table layout**. L'inspector include righe e
colonne, split/merge, preset (solo a tabella vuota), **Structure table
settings**, spazi, altezza delle celle vuote, impilamento responsive e opzioni
semantiche/accessibilità. Larghezza/altezza della tabella sono in **Size** e la
cornice della tabella in **Border**; seleziona una cella per modificarne il
bordo. Gli attributi globali del bordo delle celle restano source-level.
Seleziona una cella per modificarne span e
posizione; padding, sfondo, bordo ed effetti usano le sezioni generiche
dell'inspector, mentre l'allineamento riguarda gli elementi dentro la cella.
Con una cella selezionata sono nascosti
Position/Size/Responsive/Layout/Grid/Bindings; Tooltip, Events e Drag & drop
funzionano sul `<td>` e scattano sulle aree vuote, mentre l'evento di un
componente figlio ha la precedenza. Con più celle selezionate le sezioni CSS
generiche si applicano a tutte.
Le celle sono evidenziate nella preview durante la progettazione. Vedi
[examples/table_layout](../examples/table_layout/README.md).

## Tabelle e dati non fidati

Una tabella con `name:` è un bersaglio runtime:

```sui
table name: "utenti" {
  thead { trow { tcell "ID" head: true ... } }
  tbody { trow { tcell "" ... } }   # modello grafico
}
```

```sui
fn load() -> null {
  n, e = table.display("utenti", data)     # JSON, CSV/TSV, array, oggetti
  if e != null { out = e }
}

fn save() -> null {
  rows, e = table.data("utenti")           # righe visibili (incluse modifiche)
  part, e2 = table.select("utenti", [0, 2])
  body, e3 = table.json("utenti")
  csv, e4 = table.exportText("utenti", "csv;formula=escape")
}
```

- `table` è un modulo core, sempre disponibile, sincrono: niente `import`.
- Le celle sono scritte come testo: HTML nei dati non viene interpretato.
- Limiti: 100.000 righe, 10.000 colonne, 20 MB di testo per chiamata.
- `table.importText`/`table.exportText` supportano `csv`, `tsv`, `ssv`, `psv`,
  `json`, `jsonl` e le opzioni `header`, `sep`, `quote`, `trim`, `formula`.
- Dettagli: [`STDLIB.md`](STDLIB.md#table-module-table-core).

### Colonne dichiarative e vista dati

```sui
table name: "utenti" columns: "id,nome,stato,importo"
	columnLabels: "ID,Nome,Stato,Importo"
	columnWidths: "60,160,100,110"
	columnAlign: "right,left,center,right"
	columnTypes: "number,text,select,number"
	columnFormats: "text,text,text,currency:EUR"
	columnOptions: "stato=attivo|sospeso"
	sortable: true filterable: true pageSize: 20
	selection: "row" multiSelect: true {
}
```

- `columns`, `columnLabels`, `columnWidths`, `columnAlign`: chiavi, testi,
  larghezze e allineamenti posizionali.
- `columnTypes`: `text`, `number`, `date`, `boolean`, `select`; le celle editable
  vengono validate al commit e i valori non validi sono ripristinati con
  `aria-invalid`.
- `columnFormats`: `text`, `number`, `currency:EUR`, `percent`, `date`; il valore
  raw resta disponibile per ordinamento, editing ed export.
- `columnOptions`: `colonna=valore1|valore2;...` per `select`.
- `cellText`: `truncate` (default; la larghezza della colonna non cresce con il
  testo, una riga con ellissi e testo completo in overlay al passaggio del
  mouse) oppure `wrap` (larghezza colonna fissa, testo a capo e altezza riga che
  cresce). Cambiando la larghezza della colonna il testo si ridispone secondo la
  modalità scelta. In modalità `persistent` vengono salvati con i dati anche
  larghezze colonne/righe ed etichette header.
- Tema: la tabella eredita il tema attivo del progetto tramite `--sui-table-*`
  (`bg`, `fg`, `head-bg`, `head-fg`, `row-bg`, `striped-bg`, `hover`, `border`,
  radius, padding, font e shadow). **Table settings → Appearance → theme** può
  forzare una variante Light/Dark di una famiglia generale solo su quella
  tabella; **Set default** salva la scelta per le nuove tabelle. I preset
  strutturali (Lines, Grid, Card, …) usano lo stesso token `--sui-table-border`,
  quindi seguono il tema attivo o forzato.
- `sortable` abilita l'ordinamento dell'header (e inizializza il toggle runtime).
  Nelle tabelle editable il click sull'header porta il cursore nell'etichetta;
  per ordinare si usa il pulsante di sort nell'header, così l'edit non perde il
  fuoco. `filterable` aggiunge il filtro testo globale e inizializza i filtri
  per colonna nell'header; `pageSize` pagina le righe; `selection` è `row`,
  `column` o `cell` e `multiSelect` abilita la selezione multipla.
- `editMode` è l'impostazione a 3 stati dei controlli dati:
  - `"none"`: l'utente non può modificare celle, inserire righe, importare o
    cambiare la struttura; la tabella continua a caricare e mostrare i dati.
  - `"editable"`: l'utente può modificare, importare o creare dati per la
    sessione della pagina. Un reload ripristina i dati del sorgente.
  - `"persistent"`: modifiche, dati importati, fogli creati a runtime, larghezze
    colonne/righe, etichette header, ordinamento/filtri/pagina e modalità cell
    text vengono salvati in `localStorage` e sopravvivono a reload e riavvio. Se
    il controllo parte vuoto, i dati creati dall'utente restano. I dati salvati
    hanno la precedenza su `table.display`/`table.importText` eseguiti durante il
    boot iniziale, così il dataset preimpostato non sovrascrive il lavoro
    dell'utente. `table.clearPersisted(name)` cancella lo stato salvato e
    ripristina la tabella di design.

`editMode` governa editing e inserimento dati; `showAddTable` e
`structureEditable` restano le opzioni esplicite del workbook e vengono ignorate
con `editMode: "none"`.

Helper runtime:

```sui
fn vista() -> null {
  ok, e = table.sort("utenti", "nome", "asc")
  ok2, e2 = table.filter("utenti", "stato", "attivo")
  ok3, e3 = table.filterText("utenti", "anna")
  ok4, e4 = table.page("utenti", 2)
  ok5, e5 = table.setPageSize("utenti", 50)
  righe, e6 = table.selectedRows("utenti")
  celle, e7 = table.selectedCells("utenti")
  colonne, e8 = table.selectedColumns("utenti")
  info, e9 = table.lastChange("utenti")
  n, e10 = table.clearPersisted("utenti")
}
```

`chart.clearPersisted(name)` cancella dati, layout, stile e stato Gantt salvati
di un chart o graph persistente e ripristina il sorgente di design.

Eventi: `onCellClick`, `onRowClick`, `onCellChange`, `onRowChange`,
`onTableChange`, `onRowSelect`, `onSort`. Gli handler non ricevono argomenti; i
dettagli sono in `table.lastChange(name)`.

### Impostazioni tabella runtime (stile Excel)

Una tabella con `name` mostra un piccolo ingranaggio in una riga titolo appena
sopra la tabella, allineato a destra sulla stessa riga del titolo/caption. Un
click apre il pannello delle impostazioni:

- **Filter on column headers**: ogni intestazione ha un piccolo pulsante a
  imbuto allineato a destra, sulla stessa riga del testo. Apre un pannello in
  stile Excel con search box, checkbox **(Select All)**, i valori unici della
  colonna in ordine alfabetico con checkbox e i pulsanti **Clear Filter** /
  **Apply**. Più filtri di colonna si combinano in AND.
- **Sort by clicking headers**: ogni intestazione ha un piccolo pulsante di
  ordinamento allineato a destra, sulla stessa riga del testo. Cicla crescente,
  decrescente e ordine originale; Ctrl/Command/Shift aggiunge o attiva/disattiva
  un livello di ordinamento. Nelle tabelle non editable anche il click sul testo
  dell'header ordina; in quelle editable il testo è il campo di rinomina, quindi
  si usa solo il pulsante.
- **Sort levels**: elenco ordinato di coppie colonna/direzione. Le righe sono
  ordinate per il primo livello, i pareggi sono risolti dal successivo, come un
  ordinamento multi-colonna di Excel.
- **Clear filters** e **Clear sort** azzerano lo stato runtime.

I toggle partono da `sortable` e `filterable`. Il pannello è identificato dal
`name` della tabella, quindi il nome è necessario; il designer assegna
automaticamente `table1`, `table2`, ... alle nuove tabelle. Lo stato di filtro e
ordinamento vive per la sessione della pagina e non viene scritto nel sorgente
`.sui`. `table.sort` imposta un solo livello; il pannello può aggiungerne altri.
Un cambio di ordinamento registra
`{ op: "sort", column, dir, keys: [{ column, dir }, ...] }` in
`table.lastChange(name)`.

### Sheet tabs in stile Excel

```sui
tabs "Workbook" position: "bottom" variant: "sheets" {
	tab "Utenti" {
		table name: "utenti" columns: "id,nome" columnLabels: "ID,Nome" sortable: true { }
	}
	tab "Ordini" {
		table name: "ordini" columns: "id,totale" columnLabels: "ID,Totale" sortable: true { }
	}
}
```

La barra delle tab è sotto i pannelli e usa lo stile compatto “sheet”; ogni
tabella mantiene dati, ordinamento, filtri, pagina ed eventi indipendenti. Con un
solo foglio la tab del foglio è nascosta; la **+** è sempre visibile nella
preview del designer e a runtime solo se il workbook ha un foglio editable. Dal
secondo foglio in poi compaiono le tab del workbook. Doppio click su una tab o
menu contestuale per rinominare il workbook; ogni tab ha una piccola **×** e il
menu contestuale permette anche di eliminare il workbook. Nell'app compilata il
runtime aggiunge la **×** alle tab la cui tabella è editable e il menu contestuale
della tab offre **Delete workbook**; l'eliminazione rimuove il foglio
sorgente/runtime dal workbook per la sessione della pagina, anche attraverso i
re-render reattivi. Il padding del
controllo (default 10 px) allontana il viewport del foglio dal bordo; le strisce
di selezione riga/colonna sono a cavallo del bordo cella in questo padding e il
triangolo nella prima cella seleziona tutte le celle. Nel
designer visivo il workbook è un unico nodo `table`: i fogli senza una modalità
`selection` esplicita sono editable, tutti condividono la dimensione visibile del
controllo tabella e la tab **+** aggiunge un foglio reale al sorgente `.sui`, non
una tabella temporanea della preview. Un foglio che richiede la selezione
riga/colonna/cella resta di sola lettura. Il controllo occupa il contenitore con
un minimo di 640 × 300 px; l'area del foglio prende lo spazio rimasto dopo la riga
titolo e le tab. Le colonne larghe scorrono in orizzontale in basso, appena sopra
le tab, e le righe in eccesso scorrono in verticale con l'intestazione fissa.

### Tabelle runtime e modifica struttura

```sui
table name: "utenti" columns: "id,nome" columnLabels: "ID,Nome"
	showAddTable: true structureEditable: true {
}
```

- `showAddTable` mostra una **tab +** alla fine della striscia degli sheet quando
  la tabella è dentro un controllo `tabs`. Il click aggiunge un nuovo
  sheet/tabella con aspetto e comportamento della tabella di origine, una propria
  intestazione generica (`A`, `B`, ...), colonne di testo generiche e due righe
  corpo vuote. Il nuovo foglio è sempre editable, anche se la tabella di origine
  era di sola lettura o usava colonne numeriche, così l'inserimento manuale non
  eredita la validazione della sorgente. La nuova tabella riceve un nome generato
  (`table-1`, `table-2`, …) e resta disponibile tra i re-render reattivi della
  sessione; il foglio selezionato resta attivo quando la vista si aggiorna.
- `structureEditable` abilita il menu contestuale runtime di riga/colonna. Clic
  destro su una cella o sull’header permette di:
  - inserire riga sopra/sotto o colonna a sinistra/destra;
  - eliminare le righe/colonne selezionate;
  - copiare o tagliare righe/colonne;
  - incollare la selezione precedente;
  - aggiungere riga o colonna in fondo.
- Il menu agisce sulla multi-selezione corrente quando sono selezionate righe o
  colonne (`selection: "row"`/`"column"` + `multiSelect: true`), altrimenti sulla
  riga/colonna cliccata.
- Le operazioni strutturali aggiornano subito la griglia, chiamano
  `onTableChange` e registrano `row-insert`, `row-delete`, `column-insert` o
  `column-delete` in `table.lastChange`.
- Una tabella creata a runtime è sempre editable, quindi si può riempire senza
  uscire dall’app; `structureEditable` abilita in più il menu riga/colonna.
- Le modifiche runtime a celle del corpo e etichette di intestazione restano
  nello store della tabella tra re-render e cambio foglio per la sessione della
  pagina. Le etichette usano lo span del testo, quindi i pulsanti sort/filtro
  continuano a funzionare. La digitazione non ridisegna la tabella: il testo è
  salvato subito e il valore è finalizzato al blur o dopo una breve pausa. Gli eventi di cambio e la
  validazione scattano al blur. Gli eventi `input` delle celle restano interni
  alla tabella: gli handler `onInput` applicativi sulla tabella o sui suoi
  antenati non vengono chiamati per le modifiche alle celle; usa `onCellChange`.
  Usa `storage.*` se i valori devono sopravvivere a un reload completo.

## Sicurezza

Interpolazioni e attributi sono **escapati automaticamente**; `href`/`src`
sono filtrati per schema (`javascript:`/`data:` bloccati) e gli schemi attivi
statici sono rifiutati in compilazione. Il `style:` dinamico richiede
`cssSafe(...)`; le classi dinamiche sono filtrate sulla allow-list delle classi
dell'app. Gli URL dinamici richiedono `http.getSafe`/`postSafe`/
`getStatusSafe`/`getWithHeadersSafe`/`postWithHeadersSafe`/`requestSafe`/`uploadSafe`
(o `ws.connectSafe`/`ws.connectWithOptionsSafe`) con allow-list letterale: niente SSRF,
redirect o credenziali. Non concatenare dati non fidati in `style` o in URL non
validati. Gli iframe hanno `sandbox=""` e `referrerpolicy="no-referrer"` per
default; `src` dinamico richiede `srcAllow` (allow-list letterale) e viene
validato a runtime, con un messaggio visibile se il dato non è accettabile.
Le policy (style/class/URL, limiti tabella, default CSV, iframe) sono
modificabili con `simpleui.security.json` o `--security <file>`; i default sono
severi. Modello completo, rischi residui e pattern sicuri:
[`SECURITY.md`](../SECURITY_IT.md); esempi eseguibili:
[`examples/security`](../examples/security).

### Iframe

```simpleui
iframe src: "https://example.com/embed"
   sandbox: "allow-scripts allow-forms"
   allow: "camera; microphone"
   referrerpolicy: "no-referrer"
   loading: "lazy"
   width: 640
   height: 360

iframe src: videoUrl srcAllow: "https://example.com"
```

- `src`: `http(s)` o relativo; un URL dinamico richiede `srcAllow` (default).
- `srcAllow`: allow-list **letterale** di origini/prefix di path, separati da
  spazio o virgola. Non viene emessa nell'HTML.
- `sandbox`: token letterali; se assente il compilatore aggiunge `sandbox=""`.
- `referrerpolicy`: valore letterale; se assente aggiunge `no-referrer`.
- `allow`, `loading`, `width`, `height`: letterali e validati.
- Un URL dinamico rifiutato non viene caricato: al suo posto compare
  `Cannot accept the data: ...`.

## CLI

```bash
simpleui version
simpleui help
simpleui check examples/counter.sui --json
simpleui parse examples/counter.sui
simpleui fmt examples/counter.sui --write
simpleui docs --agent
simpleui docs --search table
simpleui build examples/counter.sui --target web --out dist
simpleui build examples/counter.sui --single --out app.html
simpleui build examples/counter.sui --single --minify --out app.html
simpleui build examples/counter.sui --target desktop --out dist \
  --allow-exec cmd --allow-dir "C:/logs"
simpleui build examples/counter.sui --target desktop-webview --out dist \
  --allow-exec cmd --allow-dir "C:/logs"
simpleui serve examples/counter.sui --watch
simpleui gui
simpleui gui examples/counter.sui
simpleui associate
simpleui lsp
simpleui mcp
```

## Target

- **web**: genera `index.html`, `style.css`, `app.js`; con `--single` genera un solo HTML con CSS/JS/runtime inline. `--out app.html` indica il file, `--out dist` produce `dist/app.html`.
- **web `--sourcemap`**: source map JavaScript e CSS opt-in. Le build multi-file
  scrivono `app.js.map` e `style.css.map` e aggiungono i commenti
  `sourceMappingURL`; `--single` incorpora entrambe le mappe come URL `data:`.
  Le mappe includono `sourcesContent`: non pubblicarle. Non combinabile con
  `--minify`. I browser non hanno source map native per HTML: il percorso
  SUI → HTML è coperto dalla mappa JavaScript perché `app.js` costruisce il DOM.
- **single HTML e CSP**: CSS/JS sono inline; se servi il file con una Content-Security-Policy restrittiva devi consentire `'unsafe-inline'` (o usare hash/nonce). La build multi-file non ha questo vincolo.
- **desktop**: genera eseguibile single-file con bridge nativo (exec/fs) + micro-server localhost + apertura browser predefinito. Supporta `--no-open`, `--port N`. Non combinabile con `--single`.
- **desktop-webview**: genera eseguibile single-file con lo stesso bridge e server localhost, ma apre una finestra WebView nativa invece del browser: WebView2 su Windows, WKWebView su macOS, WebKitGTK su Linux. Supporta `--no-open`, `--port N`. Non combinabile con `--single`. La prima build scarica `github.com/webview/webview_go`; il target richiede CGO e un toolchain C (Xcode Command Line Tools su macOS; su Linux installare i pacchetti di sviluppo GTK 3 e WebKitGTK 4.0).
- **gui [file.sui]**: apre il designer visuale. Con un file lo carica subito e **Save** riscrive lo stesso percorso. `--port N` fissa la porta dell'editor; `--security <file>` seleziona la policy.
- **associate**: su Windows associa i file `.sui` all'editor sotto `HKEY_CURRENT_USER` (senza diritti amministratore). `--uninstall` rimuove l'associazione e `--dry-run` stampa solo i comandi di registro. Un doppio clic su un file `.sui` apre `simpleui gui <file>`.


