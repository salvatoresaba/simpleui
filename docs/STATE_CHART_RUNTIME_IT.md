# Guida runtime del grafico di stato

Questa guida descrive il lato runtime del chart `state`: aggiornamento di tag e
colori da codice SUI, passaggio tra modalità editing e static, ed eventi su
forme, link e tag. L’esempio completo è
[`examples/process_monitor`](../examples/process_monitor/README.md).

## Panoramica

Un grafico di stato è identificato dall’attributo `name:`. Il codice runtime può
usare il modulo core `chart` (nessun import) per modificare il grafico, oppure
passare alla modalità static e ricevere eventi quando l’utente interagisce con
forme e tag.

Il modulo è sincrono e restituisce `(valore, errore)`.

```sui
graph "Process monitor" name: "monitor" data: data
```

## Identità del chart

Le chiamate runtime usano come primo argomento il `name:` (o il titolo) del
grafico:

```sui
ok, err = chart.setTag("monitor", "process", "RUNNING", "#f59e0b", "#111827")
```

## Identità degli elementi

Ogni forma deve dichiarare un `id` esplicito e ogni id deve essere univoco
all’interno del grafico. Il parsing si interrompe con un errore se un id manca o
è duplicato. I link usano l’`id` esplicito quando presente; altrimenti ricevono
`edge_1`, `edge_2`, ... e il runtime verifica che non entrino in collisione con
gli id delle forme.

L’id è l’unico riferimento pubblico a un elemento:

```json
{
  "states": [
    { "id": "process", "label": "Process", "onClick": "processClicked" }
  ],
  "transitions": [
    { "id": "edge_1", "from": "process", "to": "output" }
  ]
}
```

- dati sorgente: `"id": "process"`;
- API runtime: `chart.get("monitor", "process")`;
- payload eventi: `event["id"]`.

Usare id stabili: rinominare un id rompe tutte le chiamate runtime. L’API
accetta solo l’id esplicito; chiavi con prefisso, label e coppie di endpoint non
sono valide.

## Hint e tabella

Ogni forma e link accetta un campo `hint`:

```json
{ "id": "process", "label": "Process", "hint": "Elaborazione ordine" }
```

- se `hint` è una stringa non vuota sostituisce il tooltip predefinito;
- se è vuoto o assente viene mostrato il tooltip predefinito;
- `hint: "none"`, `"off"` o `"-"` nasconde il tooltip solo per quell’elemento;
- `stateHints: false` sul chart nasconde tutti gli hint.

Il campo `hint` viene importato da tutti i formati:

- JSON: proprietà `hint`;
- CSV: colonna `hint`;
- XML: attributo `hint` o figlio `<hint>` su uno stato, e attributo
  `hint` su una transizione.

L’hint si può modificare:

- dalla colonna **Hint** della tabella (tab forme e link);
- dal menu contestuale del grafico → **Hint settings...** (il dialogo si apre
  sulla tab **Hint**).

### Testo multiriga per tag e hint

I campi tag e hint sono a riga singola. Scrivere `|` dove serve l’interruzione:

```json
{ "id": "process", "tag": "STEP 1|STEP 2", "hint": "Prima riga|Seconda riga" }
```

- `|` inizia una nuova riga;
- `\|` è una pipe letterale e `\\` un backslash letterale;
- i ritorni a capo reali nei dati esistenti sono ancora accettati e normalizzati
  in `|`;
- le celle tabella, il campo **Tooltip hint** e **Tag text** sono a riga singola:
  digitare `|` per andare a capo;
- i tag mostrano al massimo 100 righe; il resto viene unito all’ultima riga con
  un’ellissi per mantenere il DOM limitato;
- i vecchi export JSON (version 1) vengono migrati all’import: le pipe esistenti
  restano letterali e i ritorni a capo reali diventano interruzioni. Questa
  versione esporta `version: 2` e usa `|` come marcatore di riga;
- i JSON senza versione e il sorgente `.sui` usano direttamente il nuovo modello;
  le pipe letterali vanno scritte `\|` (JSON: `\\|`).

Solo `tag` e `hint` interpretano `|`; label, description, id e colori trattano
la pipe come un carattere normale.

L’escape dipende dal contenitore:

| Contenitore | Interruzione | Pipe letterale |
|---|---|---|
| Campi UI | `|` | `\|` |
| File JSON / oggetto API | `|` | `\\|` |
| CSV | `|` | `\|` |
| XML | `|` | `\|` |
| Stringa JSON inline in `.sui` | `|` | `\\\\|` |

I backslash extra servono perché lo strato stringa SUI e lo strato stringa JSON
decriptano un livello ciascuno.

### Stile globale degli hint

Il pannello impostazioni del grafico ha un pulsante dedicato **HINTS** (accanto a
GRID, SHAPES, LINKS e TAGS) che apre un pannello autonomo. Lo stile vale per tutti
i tooltip del grafico; per singolo elemento restano personalizzabili solo il
testo `hint` e la sua soppressione (`"none"`). Il pannello elenca gli attributi
come una tabella a due colonne e termina con **Cancel** / **Apply**.

| Controllo | Effetto |
|---|---|
| Show hover hints | abilita o disabilita tutti i tooltip |
| Hint background | riempimento del tooltip (`#rrggbb`, `transparent`) |
| Background opacity | `0`–`1`, moltiplica l’alpha dello sfondo |
| Hint text color | colore del testo |
| Hint border color | colore del bordo |
| Hint font | famiglia del font |
| Font size | dimensione del testo in pixel |
| Border width | spessore del bordo in pixel (`0` lo rimuove) |
| Text padding | distanza tra testo e bordo |
| Corner radius | arrotondamento del contenitore |

Il pannello include una **Hint preview** dal vivo che si aggiorna mentre si
spostano gli slider; il tooltip reale usa le stesse variabili CSS
(`--sui-chart-tooltip-bg`, `-fg`, `-border`, `-border-width`, `-padding`,
`-radius`, `-font-family`, `-font-size`, `-shadow`).

- opacity `0` o sfondo `transparent` rimuovono riempimento e ombra;
- i valori sono salvati in `localStorage` per chart e sopravvivono al reload;
- **Save settings as default** (tab General) li salva come fallback per ogni
  chart che non ha ancora impostazioni proprie;
- i valori non impostati usano i default del tema.

Gli stessi default si possono scrivere nel sorgente `.sui` come attributi del
chart: `stateHintBg`, `stateHintOpacity`, `stateHintTextColor`,
`stateHintBorderColor`, `stateHintFont`, `stateHintFontSize`,
`stateHintBorderWidth`, `stateHintPadding`, `stateHintRadius`. Il designer li
espone da **Chart settings → Interactions → HINTS...**; le modifiche runtime
restano in `localStorage`, mentre **Reset settings** torna ai valori sorgente.

Anche la tab General è disposta come una tabella a due colonne: etichette a
sinistra e input a destra. Non c'è più uno switch static separato: scegli
`editMode` (`none`, `editable` o `persistent`) nel sorgente o nel designer.

La tabella è controllata da `tableView`:

| Valore | Superficie |
|---|---|
| `chart` | solo diagramma |
| `table` | solo tabella |
| `both` | diagramma + tabella |

Il designer espone la stessa scelta come **Table view**. Quando la tabella è
visibile e si seleziona un elemento nel diagramma, la riga corrispondente viene
evidenziata, portata in vista e riceve il focus. Selezionando un link si attiva
automaticamente la tab **Links**.

## Modalità di editing

`editMode` accetta tre valori:

| Valore | Comportamento |
|---|---|
| `none` | Superficie di presentazione: niente selezione, drag/resize, chrome di editing o modifiche tabella; pan/zoom e coda eventi restano attivi |
| `editable` | Editor: selezione, drag/resize, maniglie route/tag, strumenti add/delete e tabella modificabile, per la sessione della pagina |
| `persistent` | Come editable, ma geometria/stile/attributi sopravvivono a reload e riavvio |

Imposta la modalità dal sorgente o da uno stato reattivo:

```sui
state chartMode = "editable"

graph "Monitor" name: "monitor" data: data editMode: chartMode
```

La modalità non-editable mantiene pan e zoom, disabilita strumenti, maniglie,
selezione, editing route/tag e tabella, e lascia attiva la coda eventi.
## Lista eventi

Il chart è una **sorgente di eventi**, non un dispatcher. Non deve chiamare
funzioni che modificano lo stato SUI esterno. Ogni interazione in modalità static
viene accodata in una lista eventi per chart, con il tipo di evento e l’elemento
che l’ha prodotto. Il codice SUI svuota la lista e decide quali funzioni esterne
chiamare.

```sui
fn startEvents() -> null {
	interval(120, pumpChartEvents)
}

async fn pumpChartEvents() -> null {
	events, err = chart.pollEvents("monitor")
	if err != null {
		return
	}
	items, jerr = json.parse(events)
	if jerr != null {
		return
	}
	let i = 0
	while i < len(items) {
		let ev = items[i]
		if str(ev["event"]) == "click" and str(ev["part"]) == "body" {
			if str(ev["kind"]) == "shape" and str(ev["id"]) == "process" {
				processClicked(ev)
			}
		}
		i = i + 1
	}
}

fn processClicked(ev: any) -> null {
	ok, err = chart.update("monitor", "output", "\{\"text\":\"DONE\",\"tagColor\":\"#22c55e\",\"textColor\":\"#ffffff\"\}")
	modal.open("process-details")
}

view main {
	onLoad: startEvents
	graph "Monitor" name: "monitor" data: data
		editMode: "none"
}
```

### Funzioni della lista eventi

| Funzione | Risultato |
|---|---|
| `chart.pollEvents(name)` | tutti gli eventi pendenti in JSON e svuota la lista |
| `chart.peekEvents(name)` | tutti gli eventi pendenti senza svuotare |
| `chart.nextEvent(name)` | il più vecchio evento in JSON e lo rimuove (`""` se vuota) |
| `chart.clearEvents(name)` | svuota la lista |
| `chart.eventCount(name)` | numero di eventi pendenti |

Ogni evento ha questo payload:

```json
{
  "chart": "monitor",
  "kind": "shape",
  "part": "body",
  "target": "shape",
  "event": "click",
  "id": "process",
  "label": "Process",
  "tag": "RUNNING",
  "from": "",
  "to": "",
  "clientX": 420,
  "clientY": 260,
  "button": 0,
  "altKey": false,
  "ctrlKey": false,
  "metaKey": false,
  "shiftKey": false,
  "time": 1730000000000
}
```

- `kind`: `shape` o `link`;
- `part`: `body` o `tag`;
- `event`: `click`, `dblclick`, `mouseover`, `mouseout`, `contextmenu`;
- `id`: l'elemento che ha prodotto l'evento;
- `label`, `tag`, `from`, `to`, coordinate e modificatori.

La lista conserva al massimo gli ultimi 200 eventi. Il polling è il modello
consigliato perché il chart produce solo dati; SUI possiede tutte le modifiche
di stato.

### Callback diretti rimossi

Il modello a callback diretti (`chart.on`, campi handler sugli elementi e
attributi chart-level `onShape*`/`onLink*`/`onTag*`) e stato rimosso. Il chart e
solo una sorgente di eventi: usare `chart.pollEvents` / `chart.peekEvents` /
`chart.nextEvent` e smistare l'evento alle funzioni applicative.

## API runtime del chart

Le funzioni del chart di stato inizializzano i dati su richiesta, quindi si
possono usare da `onLoad` anche prima che il primo disegno sia completato. Ogni
funzione restituisce `(valore, errore)`.

### `chart.setTag`

```sui
ok, err = chart.setTag(name, id, text, tagColor, textColor)
```

Imposta testo, colore sfondo e colore testo del tag di un elemento. Una stringa
vuota per un colore lascia il valore invariato.

```sui
ok, err = chart.setTag("monitor", "process", "RUNNING", "#f59e0b", "#111827")
```

### `chart.update`

```sui
ok, err = chart.update(name, id, patch)
```

Applica una patch a un elemento. La patch può essere un oggetto SUI o una stringa
JSON.

```sui
ok, err = chart.update("monitor", "process", "\{\"text\":\"RUNNING\",\"tagColor\":\"#f59e0b\",\"textColor\":\"#111827\",\"color\":\"#fffbeb\"\}")
```

### `chart.updateMany`

```sui
count, err = chart.updateMany(name, updates)
```

`updates` può essere una mappa `{ id: patch }` o un array con `id` e patch. Il
grafico viene aggiornato una sola volta.

```sui
n, err = chart.updateMany("monitor", "\{\"intake\":\{\"text\":\"READY\"\},\"process\":\{\"text\":\"RUNNING\",\"tagColor\":\"#f59e0b\"\}\}")
```

### `chart.get`

```sui
jsonText, err = chart.get(name, id)
```

Restituisce gli attributi **correnti** di una forma/link come oggetto JSON.
Include le modifiche runtime, non solo i dati sorgente. Il risultato contiene
`id`, `label`, `tag`, geometria e colori.

```sui
info, err = chart.get("monitor", "process")
if err == null {
	state, jerr = json.parse(info)
}
```

### `chart.elements`

```sui
jsonText, err = chart.elements(name)
```

Restituisce un array JSON con un riepilogo per forma/link: `kind`, `id`,
`label`, `tag`, `hint`, `shape` (forme), `from`/`to`/`routing` (link), `color`,
`tagTextColor`, `tagBoxColor` e i nomi dei gestori `onClick`/`onTagClick`
associati. Serve per enumerare il grafico e scoprire gli id da aggiornare.

```sui
all, err = chart.elements("monitor")
items, jerr = json.parse(all)
```

Ogni elemento può essere passato direttamente alle API runtime:

```sui
id = str(item["id"])            // "process" o "edge_1"
ok, err = chart.setTag("monitor", id, "RUNNING", "#f59e0b", "#111827")
```

### `chart.find`

```sui
jsonText, err = chart.find(name, query)
```

Cerca nel grafico e restituisce un array JSON con gli stessi riepiloghi di
`chart.elements`. Una query vuota restituisce tutti gli elementi. Una query
`campo:valore` filtra un campo:

```sui
all, err = chart.find("monitor", "")
running, err2 = chart.find("monitor", "tag:RUNNING")
shapes, err3 = chart.find("monitor", "kind:shape")
process, err4 = chart.find("monitor", "id:process")
```

Campi supportati: `id`, `kind`, `label`, `tag`, `hint`, `shape`, `from`, `to`,
`routing`. Senza due punti la query viene cercata in tutti i campi.

### `chart.getAttr`

```sui
value, err = chart.getAttr(name, id, field)
```

Legge un attributo come stringa. `kind` è calcolato; gli altri nomi
sono cercati sull’elemento corrente. Un campo vuoto o sconosciuto restituisce
`("", null)`.

```sui
tag, err = chart.getAttr("monitor", "process", "tag")
kind, err2 = chart.getAttr("monitor", "process", "kind")   // shape
```

### `chart.setAttr`

```sui
ok, err = chart.setAttr(name, id, field, value)
```

Scrive un attributo con la stessa validazione di `chart.update`. I campi non
supportati restituiscono errore.

```sui
ok, err = chart.setAttr("monitor", "process", "text", "RUNNING")
ok2, err2 = chart.setAttr("monitor", "process", "tagColor", "#f59e0b")
```

### `chart.focusRow`

```sui
ok, err = chart.focusRow(name, id)
```

Evidenzia, porta in vista e focalizza la riga corrispondente della tabella.
Funziona anche in modalità static, a differenza della selezione del grafico.
Passare `""` come `id` per rimuovere l’evidenziazione. Selezionando un link si
attiva la tab **Links**.

```sui
ok, err = chart.focusRow("monitor", "output")
ok2, err2 = chart.focusRow("monitor", "edge_1")
```

### `chart.pollEvents`

```sui
jsonText, err = chart.pollEvents(name)
```

Restituisce tutti gli eventi pendenti come array JSON e svuota la lista. È il
modo consigliato per consumare gli eventi del chart.

### `chart.peekEvents`

```sui
jsonText, err = chart.peekEvents(name)
```

Restituisce gli eventi pendenti senza svuotare la lista.

### `chart.nextEvent`

```sui
jsonText, err = chart.nextEvent(name)
```

Restituisce il più vecchio evento e lo rimuove. Restituisce `""` se la lista è
vuota.

### `chart.clearEvents`

```sui
ok, err = chart.clearEvents(name)
```

Svuota la lista eventi.

### `chart.eventCount`

```sui
count, err = chart.eventCount(name)
```

Restituisce il numero di eventi pendenti.

### Identificatori elemento

| Target | Id accettato |
|---|---|
| Forma | il suo `id`, per esempio `process` |
| Link | il suo `id`, per esempio `edge_1` |

Tutte le funzioni runtime accettano **solo l’id esplicito**. Prefissi (`shape:`,
`state:`, `link:`, `edge:`), label e `from->to` non sono accettati. Gli id sono
univoci tra forme e link.

### Getter e setter tipizzati

Per modificare un singolo attributo usare gli helper tipizzati. Ogni getter
restituisce `(string, error)`, ogni setter `(bool, error)`.

| Attributo | Getter | Setter |
|---|---|---|
| label | `chart.getLabel(name, id)` | `chart.setLabel(name, id, value)` |
| testo tag | `chart.getText(name, id)` | `chart.setText(name, id, value)` |
| hint | `chart.getHint(name, id)` | `chart.setHint(name, id, value)` |
| description | `chart.getDescription(name, id)` | `chart.setDescription(name, id, value)` |
| colore forma/link | `chart.getColor(name, id)` | `chart.setColor(name, id, value)` |
| sfondo tag | `chart.getTagColor(name, id)` | `chart.setTagColor(name, id, value)` |
| colore testo tag | `chart.getTextColor(name, id)` | `chart.setTextColor(name, id, value)` |
| tipo forma | `chart.getShape(name, id)` | `chart.setShape(name, id, value)` |
| routing link | `chart.getRouting(name, id)` | `chart.setRouting(name, id, value)` |
| colore bordo | `chart.getBorderColor(name, id)` | `chart.setBorderColor(name, id, value)` |
| stile bordo | `chart.getBorderStyle(name, id)` | `chart.setBorderStyle(name, id, value)` |
| spessore bordo | `chart.getBorderWidth(name, id)` | `chart.setBorderWidth(name, id, value)` |
| raggio angoli | `chart.getCornerRadius(name, id)` | `chart.setCornerRadius(name, id, value)` |

Accessor aggiuntivi:

| Attributo | Getter | Setter |
|---|---|---|
| larghezza (shape `w` / linea link) | `chart.getWidth` | `chart.setWidth` |
| altezza (shape) | `chart.getHeight` | `chart.setHeight` |
| posizione x | `chart.getX` | `chart.setX` |
| posizione y | `chart.getY` | `chart.setY` |
| rotazione | `chart.getRotation` | `chart.setRotation` |
| stile tratteggio link | `chart.getDashStyle` | `chart.setDashStyle` |
| marker link | `chart.getStartMarker` / `chart.getEndMarker` | `chart.setStartMarker` / `chart.setEndMarker` |
| estremi link | `chart.getFrom` / `chart.getTo` | `chart.setFrom` / `chart.setTo` |
| waypoint link | `chart.getWaypoints` | `chart.setWaypoints` |
| tag kind/box/align/leader | `chart.getTagKind` / `getTagBox` / `getTagAlign` / `getTagLeader` | `chart.setTagKind` / `setTagBox` / `setTagAlign` / `setTagLeader` |
| text placement / tag side | `chart.getTextPlacement` / `chart.getTagSide` | `chart.setTextPlacement` / `chart.setTagSide` |
| dash bordo | `chart.getBorderDashLength` / `chart.getBorderDashGap` | `chart.setBorderDashLength` / `chart.setBorderDashGap` |
| snap | `chart.getSnap` | `chart.setSnap` |
| kind | `chart.getKind` | sola lettura |

`setWidth` distingue forma e link: sulla forma scrive `w`, sul link la larghezza
della linea. `setHeight` vale solo per le forme. `setWaypoints` accetta un array
o una stringa JSON.

```sui
ok1, e1 = chart.setText("monitor", "output", "DONE")
ok2, e2 = chart.setTagColor("monitor", "output", "#22c55e")
ok3, e3 = chart.setTextColor("monitor", "output", "#ffffff")

label, e4 = chart.getLabel("monitor", "process")
hint, e5 = chart.getHint("monitor", "process")
kind, e6 = chart.getKind("monitor", "edge_1")
```

Per gli altri attributi usare gli accessor generici:

```sui
value, err = chart.getAttr("monitor", "process", "tagKind")
ok, uerr = chart.setAttr("monitor", "process", "tagKind", "external")
```

`chart.update` resta disponibile per modifiche multi-attributo o batch.

### Campi patch

Tag:

- `tag` / `text` / `tagText`
- `tagTextColor` / `textColor`
- `tagBoxColor` / `tagColor` / `backgroundColor` / `boxColor`
- `tagBorderColor`, `tagLeaderColor`
- `tagKind`, `tagAlign`, `tagBox`, `tagLeader`
- `tagArrow`, `tagArrowSize`, `tagArrowShape`
- `tagFont`, `tagFontSize`, `tagFontWeight`, `tagFontStyle`
- `tagTextDecoration`, `tagTextAlign`, `tagLetterSpacing`, `tagLineSpacing`
- `tagTextTransform`, `tagTextShadow`, `tagTextShadowColor`
- `tagPadding`, `tagRotation`, `tagOffset`

Elemento:

- `label`, `description`
- `color` (fill forma o colore linea)
- `borderColor`, `highlightColor`
- `borderStyle`, `borderWidth`, `cornerRadius`
- `shape` (`rectangle`, `square`, `circle`, `stadium`, `ellipse`, `rhombus`,
  `triangle`, `parallelogram`, `hexagon`, `trapezoid`, `pentagon`, `note`)
- `routing` per i link (`auto`, `straight`, `arc`, `bezier`, `curved`,
  `orthogonal`, `circular`)
- `textPlacement`, `tagSide`, `snap`
- `borderDashLength`, `borderDashGap`

Geometria forme:

- `x`, `y` (coordinate del centro), `w`, `h`, `rotation`
- `width` / `height` sono alias dei link; la geometria delle forme usa `w` / `h`

Campi link:

- `width` (spessore linea), `dashStyle`, `dashLength`, `dashGap`, `arrowSize`
- `startMarker`, `endMarker` (`none`, `triangle`, `open`, `circle`, `square`)
- `from`, `to` (id espliciti delle forme, validati)
- `fromPort`, `toPort`, `fromAnchor`, `toAnchor`
- `waypoints` (array o stringa JSON di `{x, y}`)
- `arcRadius`, `arcSweep`, `arcLarge`
- `bendX`, `bendY`, `bend1X`, `bend1Y`, `bend2X`, `bend2Y`

Colori: `#rgb`, `#rrggbb`, `rgb()/rgba()`, `transparent`, `none`. Valori non
validi restituiscono errore e non vengono applicati.

## Esempi

### Pulsanti che aggiornano tag e colore forma

```sui
state data = "\{\"states\":[\{\"id\":\"intake\",\"label\":\"Intake\",\"shape\":\"stadium\",\"w\":150,\"h\":64,\"tag\":\"WAITING\"\},\{\"id\":\"process\",\"label\":\"Process\",\"shape\":\"rectangle\",\"w\":160,\"h\":76,\"tag\":\"IDLE\"\},\{\"id\":\"output\",\"label\":\"Output\",\"shape\":\"circle\",\"w\":96,\"h\":96,\"tag\":\"OUTPUT\"\}],\"transitions\":[\{\"id\":\"edge_1\",\"from\":\"intake\",\"to\":\"process\",\"tag\":\"start\"\},\{\"id\":\"edge_2\",\"from\":\"process\",\"to\":\"output\",\"tag\":\"done\"\}]\}"

fn setRunning() -> null {
	ok1, e1 = chart.update("monitor", "intake", "\{\"text\":\"READY\",\"tagColor\":\"#dbeafe\",\"textColor\":\"#1e3a8a\"\}")
	ok2, e2 = chart.update("monitor", "process", "\{\"text\":\"RUNNING\",\"tagColor\":\"#f59e0b\",\"textColor\":\"#111827\",\"color\":\"#fffbeb\"\}")
	ok3, e3 = chart.update("monitor", "output", "\{\"text\":\"WAITING\",\"tagColor\":\"#e5e7eb\",\"textColor\":\"#374151\"\}")
}

view main {
	col {
		button "Running" onClick: setRunning
		graph "Process monitor" name: "monitor" data: data
	}
}
```

Gli aggiornamenti vengono applicati a tag, box e forma esistenti: il grafico non
viene ricostruito, quindi non ci sono flicker o salti di viewport.

### Consumare la lista eventi e aggiornare un’altra forma

Il chart accoda il click; SUI legge la lista e chiama la propria funzione:

```sui
import { json } from "std"

state data = "\{\"states\":[\{\"id\":\"process\",\"label\":\"Process\"\},\{\"id\":\"output\",\"label\":\"Output\"\}],\"transitions\":[\{\"id\":\"edge_1\",\"from\":\"process\",\"to\":\"output\"\}]\}"

fn startEvents() -> null {
	interval(120, pumpChartEvents)
}

async fn pumpChartEvents() -> null {
	events, err = chart.pollEvents("monitor")
	if err != null {
		return
	}
	items, jerr = json.parse(events)
	if jerr != null {
		return
	}
	let i = 0
	while i < len(items) {
		let ev = items[i]
		if str(ev["event"]) == "click" and str(ev["part"]) == "body" and str(ev["id"]) == "process" {
			processClicked(ev)
		}
		i = i + 1
	}
}

fn processClicked(ev: any) -> null {
	ok, uerr = chart.update("monitor", "output",
		"\{\"text\":\"PROCESS CLICKED\",\"tagColor\":\"#22c55e\",\"textColor\":\"#ffffff\"\}")
	if uerr != null {
		status = "Update failed: " + uerr
	}
}

view main {
	onLoad: startEvents
	graph "Monitor" name: "monitor" data: data
		editMode: "none"
}
```

Il click su `process` produce un elemento nella lista simile a:

```json
{ "event": "click", "kind": "shape", "part": "body", "id": "process" }
```

`pumpChartEvents` associa quell’evento a `processClicked`; il chart non chiama
mai codice SUI e non tocca lo stato SUI. Aggiungere altri rami allo stesso ciclo
per altre forme, tag o link.

### Identificare elementi e attributi

### Identificare elementi e attributi

Gli identificatori arrivano dai dati del chart:

```json
{
  "states": [
    { "id": "intake", "label": "Intake", "shape": "stadium", "tag": "WAITING" },
    { "id": "process", "label": "Process", "shape": "rectangle", "tag": "RUNNING" },
    { "id": "output", "label": "Output", "shape": "circle", "tag": "OUTPUT" }
  ],
  "transitions": [
    { "id": "edge_1", "from": "intake", "to": "process", "tag": "start" },
    { "id": "edge_2", "from": "process", "to": "output", "tag": "done" }
  ]
}
```

- id forma: `id` (`intake`, `process`, `output`), obbligatorio e univoco;
- label forma: `label`;
- geometria: `shape`, `w`, `h`, `x`, `y`, `rotation`;
- tag forma: `tag`, `tagTextColor`, `tagBoxColor`, `tagBorderColor`;
- colore forma: `color`;
- gestori forma: `onClick`, `onTagClick`, ...;
- id link: `id` (`edge_1`);
- tag link: `tag`, colori, `routing`, `waypoints`.

Dal payload leggere `event["id"]`, `event["label"]`,
`event["tag"]`, `event["kind"]` (`shape`/`link`) e `event["part"]`
(`body`/`tag`). Le definizioni originali sono nello stato dati del chart (per
esempio `data`). Per i valori **correnti** dopo modifiche/import usare
`chart.get(chart, id)` per un elemento e `chart.elements(chart)` per enumerarli
tutti.

Per ispezionare un elemento durante lo sviluppo, mostrare il payload o la forma
parsata:

```sui
async fn shapeClicked(eventJson: string) -> null {
	event, err = json.parse(eventJson)
	if err == null {
		status = "id=" + str(event["id"]) + " label=" + str(event["label"]) +
			" tag=" + str(event["tag"]) + " kind=" + str(event["kind"]) +
			" part=" + str(event["part"])
	}
}
```

### Switch editable/non-editable

```sui
state staticMode = false
state chartMode = "editable"

fn toggleMode() -> null {
	staticMode = staticMode == true ? false : true
	chartMode = staticMode ? "none" : "editable"
}

view main {
	row {
		switch checked: staticMode onChange: toggleMode
		text "Non-editable mode"
	}
	graph "Monitor" name: "monitor" data: data editMode: chartMode
}
```

### Click forma che apre un modal

```sui
import { json } from "std"

state data = "\{\"states\":[\{\"id\":\"process\",\"label\":\"Process\"\}],\"transitions\":[]\}"
state selectedLabel = ""
state selectedTag = ""

fn startEvents() -> null {
	interval(120, pumpChartEvents)
}

async fn pumpChartEvents() -> null {
	events, err = chart.pollEvents("monitor")
	if err != null {
		return
	}
	items, jerr = json.parse(events)
	if jerr != null {
		return
	}
	let i = 0
	while i < len(items) {
		let ev = items[i]
		if str(ev["event"]) == "click" and str(ev["part"]) == "body" and str(ev["id"]) == "process" {
			selectedLabel = str(ev["label"])
			selectedTag = str(ev["tag"])
			modal.open("process-details")
		}
		i = i + 1
	}
}

view main {
	onLoad: startEvents
	graph "Monitor" name: "monitor" data: data
		editMode: "none"

	modal "Process details" name: "process-details" {
		p "Shape: {selectedLabel}"
		p "Tag: {selectedTag}"
		modalbutton "Close"
	}
}
```

La finestra è aperta da `pumpChartEvents` quando trova il click su `process`
nella lista eventi. Il runtime del chart non chiama `modal.open` e non tocca lo
stato SUI. Lo stesso ciclo può gestire altri id, kind o eventi tag.

### Click tag che aggiorna stato esterno

In `pumpChartEvents`, discriminare `part == "tag"`:

```sui
if str(ev["event"]) == "click" and str(ev["part"]) == "tag" {
	status = "Tag " + str(ev["tag"]) + " su " + str(ev["id"])
}
```

L’elemento della lista contiene `kind`, `part`, `id`, `label` e `tag`,
quindi lo stesso ciclo gestisce click su forme, tag e link.

### Aggiornamento batch da un feed

```sui
fn applyStatuses(payload: string) -> null {
	n, err = chart.updateMany("monitor", payload)
	if err != null {
		status = "Errore aggiornamento: " + err
	}
}
```

`payload` può essere:

```json
{
  "intake": { "text": "READY", "tagColor": "#dbeafe", "textColor": "#1e3a8a" },
  "process": { "text": "RUNNING", "tagColor": "#f59e0b", "textColor": "#111827" },
  "output": { "text": "WAITING", "tagColor": "#e5e7eb", "textColor": "#374151" }
}
```

## Esempio process monitor

[`examples/process_monitor/app.sui`](../examples/process_monitor/app.sui)
combina tutte le funzionalità:

- viewport ampio (`width: "100%"`, `height: "620px"`);
- id forma obbligatori e univoci (`intake`, `process`, `output`);
- `pumpChartEvents` che interroga `chart.pollEvents("monitor")` ogni 120 ms;
- il click su **Process** nella lista eventi chiama `processClicked`, che aggiorna
  la forma **Output** e apre il modal **Process details**;
- pulsanti Waiting / Running / Done con `chart.update`;
- switch collegato a `editMode`;
- click su tag e hover gestiti dalla stessa lista eventi;
- hint personalizzati su forme e link mostrati al posto del tooltip predefinito;
- `tableView: "both"` e `chart.focusRow`, quindi un click evidenzia la riga
  tabella corrispondente anche in modalità non-editable.

## Sicurezza e prestazioni

- Il testo dei tag è inserito come testo, mai come markup.
- I colori sono validati; quelli non validi restituiscono errore.
- Il chart accoda solo dati; non deve chiamare funzioni SUI. È il codice SUI a
  decidere quale funzione esterna riceve ogni evento.
- Gli id delle forme sono validati per presenza e unicità prima del disegno.
- La modalità non-editable rende la tabella in sola lettura e rimuove i controlli di editing.
- Gli aggiornamenti tag/colori sono applicati in place all’SVG esistente; le
  modifiche strutturali (forma, routing, tipo/box/font tag, geometria) usano un
  ridisegno completo.
- Una breve finestra di soppressione del draw evita che rerender SUI o
  ResizeObserver ricostruiscano subito il chart dopo un aggiornamento in place.
- La modalità non-editable disabilita il refit automatico e ignora piccole variazioni del viewport
  (per esempio la scrollbar nascosta da un modal aperto), così le forme non
  saltano quando una finestra appare o si chiude.
- La lista eventi conserva al massimo gli ultimi 200 eventi; interrogarla
  regolarmente.

## Risoluzione problemi

- **Nessun evento nella lista:** il chart deve essere in `editMode: "none"`.
  La modalità editing è un editor e non emette eventi static.
- **La lista cresce:** chiamare regolarmente `chart.pollEvents` (per esempio con
  `interval`) oppure `chart.clearEvents`.
- **Chart sbagliato:** controllare `name:` e il primo argomento.
- **Elemento non trovato:** usare l’`id` esplicito di forma/link dai dati.
  Prefissi, label e coppie di endpoint non sono accettati.
- **Id mancante o duplicato:** aggiungere `"id"` a ogni forma; gli id devono
  essere univoci nel chart. Gli id dei link possono essere omessi, ma è
  consigliato dichiararli.
- **La funzione non viene chiamata:** controllare il ramo in `pumpChartEvents`;
  la lista fornisce `event`, `kind`, `part` e `id`.
- **Dopo il reload non succede nulla:** `interval` parte da `onLoad`, quindi il
  poller viene ricreato a ogni caricamento. Se serve, svuotare la vecchia lista
  con `chart.clearEvents`.
- **Testo aggiornato ma box non perfetto:** l’updater in place ricostruisce il
  box; per testi molto lunghi o modifiche font usare una patch strutturale
  (`tagFontSize`, `tagBox`, ecc.).
- **La modalità non cambia:** `editMode` deve essere `none`, `editable` o `persistent`.



