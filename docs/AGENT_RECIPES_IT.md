# Agent recipes: local and remote interfaces

Pattern operativi per coding agent. Queste sono **recipe**, non un riferimento
API separato: le firme canoniche sono in [`API_REFERENCE_IT.md`](API_REFERENCE_IT.md),
[`STDLIB_IT.md`](STDLIB_IT.md) e [`BUILTINS_IT.md`](BUILTINS_IT.md). Verifica
ogni esempio con `simpleui check`.

## Workflow

1. Leggi [`AGENTS_IT.md`](../AGENTS_IT.md) e [`docs/manifest.json`](manifest.json).
2. Scegli la recipe più vicina e il file più vicino in `examples/`.
3. Scrivi la più piccola app `.sui` funzionante, con stati loading/error.
4. Esegui `simpleui check app.sui --json`; correggi tutti gli errori.
5. Compila o servi l'app, poi verifica gli stati nel browser.
6. Esegui `go test -count=1 ./...` se hai modificato compilatore/runtime.

Non generare codice a memoria. Se una funzione o un attributo non è documentato,
cerca nei docs (`simpleui docs --search <term>`) o ispeziona il compilatore.

## Common skeleton

Uno schermo robusto per dati remoti ha tre stati: loading, errore e dati
(incluso il caso vuoto). Tienili separati dal payload.

```sui
import { http, json } from "std"

state status = "loading"
state errorMsg = ""

async fn refresh() -> null {
	status = "loading"
	errorMsg = ""
	body, err = http.get("https://api.example.com/items")
	if err != null {
		errorMsg = err
		status = "error"
		return
	}
	data, perr = json.parse(body)
	if perr != null {
		errorMsg = perr
		status = "error"
		return
	}
	n, terr = table.display("items", data)
	if terr != null {
		errorMsg = terr
		status = "error"
		return
	}
	status = "ok"
}

view main {
	title "Items"
	onLoad: refresh
	h1 "Items"
	if status == "error" {
		alert "{errorMsg}"
		button "Retry" onClick: refresh
	} else {
		table name: "items" {
			thead {
				trow {
					tcell "ID" head: true
					tcell "Name" head: true
				}
			}
			tbody {
				trow {
					tcell ""
					tcell ""
				}
			}
		}
	}
}
```

- `table.display` sostituisce il corpo e conserva lo stile della tabella.
- `json.parse` può restituire oggetto, array o scalare: valida la forma.
- Mostra l'errore reale: l'utente deve distinguere rete, JSON e tabella.

## Drag and drop

Usa una chiave stabile per ogni elemento. Per una lista di stringhe va bene la
stringa stessa; per record usare `item.id`.

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
	button "Scegli file..." pickFiles: attachment accept: ".csv,.txt" multiple: true
}
```

- `dragKey` + `dropKey` identificano la riga come sorgente e target; `onDrop`
  esegue il riordino vero e proprio.
- `dropFiles: state` e il target automatico per file/cartelle; abbinare
  `value: state` su input/textarea per mostrare i percorsi.
- `directory: true` sul controllo `file` o su un pulsante picker apre il
  selettore di cartelle. Il browser restituisce tutti i file ricorsivamente e
  `drag.folder()` espone il nome della cartella comune. Per un sottoinsieme
  usare il selettore multiplo o filtrare nell'app.
- `pickFiles: state` apre il dialogo nativo da un `file`, `button` o altro
  elemento cliccabile; `accept`, `multiple` e `directory` configurano il dialogo.
- `drag.files()` in `onDrop` o dopo una selezione fornisce i metadati immediati;
  il contenuto delle directory viene raccolto in modo asincrono per `dropFiles`.
- Il browser non espone percorsi assoluti per sicurezza: mostrare nome/percorso
  relativo restituito dal browser; non assumere che sia un percorso filesystem.
- Trattare i nomi come testo non fidato. Validare eventuali URL e mantenere
  l'accesso desktop dietro `--allow-dir`.

Vedi [`API_REFERENCE_IT.md`](API_REFERENCE_IT.md#drag-and-drop),
[`BUILTINS_IT.md`](BUILTINS_IT.md#drag-and-drop-functions) e la
[guida all'editor GUI](GUI_EDITOR_IT.md) per i passi nell'inspector e la
differenza tra drag and drop del designer e dell'applicazione. Esempio eseguibile:
[`examples/drag_drop.sui`](../examples/drag_drop.sui) e
[`examples/file_picker.sui`](../examples/file_picker.sui).

### Analisi dei tipi di file con backend Go

`http.postFiles` invia la selezione corrente di `pickFiles`/`dropFiles` come
multipart; il backend restituisce `{columns, rows}` per `table.display`.
L'esempio completo, con limiti di upload e server Go, e in
[`examples/file_analyzer/README.md`](../examples/file_analyzer/README.md).

```sui
import { http, json } from "std"

state files = ""
state status = "idle"
state errorMsg = ""

async fn analyze() -> null {
	status = "loading"
	errorMsg = ""
	body, err = http.postFiles("/api/analyze")
	if err != null {
		status = "error"
		errorMsg = err
		return
	}
	data, perr = json.parse(body)
	if perr != null {
		status = "error"
		errorMsg = perr
		return
	}
	status = "ok"
	n, terr = table.display("analysis", data)
	if terr != null {
		status = "error"
		errorMsg = terr
	}
}

view main {
	button "Scegli file..." pickFiles: files multiple: true
	button "Analizza" onClick: analyze
	input "Selezionati" value: files dropFiles: files
	if status == "ok" {
		table name: "analysis" {
			thead { trow { tcell "File" head: true tcell "MIME" head: true } }
			tbody { trow { tcell "" tcell "" } }
		}
	}
}
```

## Local application

### Local form + reactive table

```sui
state name = ""
state items = []
state msg = ""

fn add() -> null {
	let n = trim(name)
	if n == "" {
		msg = "Enter a name"
		return
	}
	items = append(items, n)
	name = ""
	msg = ""
	nrows, err = table.display("items", items)
	if err != null { msg = err } else { msg = str(nrows) + " rows" }
}

view main {
	title "Local items"
	h1 "Local items"
	row {
		input value: name placeholder: "Name"
		button "Add" onClick: add
	}
	if msg != "" { p "{msg}" }
	table name: "items" {
		thead { trow { tcell "Name" head: true } }
		tbody { trow { tcell "" } }
	}
}
```

`let` per valori locali, `state` per valori che devono ri-renderizzare. Per la
persistenza locale usa `storage.*`; non assumere `fs` sul web.

## Remote API

```sui
# Destinazione fissa: URL letterale consentito.
body, err = http.get("https://api.example.com/items")

# Destinazione dinamica: *Safe + allow-list letterale.
body, err = http.getSafe(endpoint, "https://api.example.com")
```

L'allow-list confronta schema/host/porta e un eventuale prefisso di path;
credenziali e redirect sono rifiutati. Non costruire l'allow-list dall'input.

### Remote JSON API + table

Vedi [`examples/users/app.sui`](../examples/users/app.sui). Forma minima:

```sui
import { http, json } from "std"

state status = "loading"
state errText = ""

async fn loadUsers() -> null {
	status = "loading"
	body, herr = http.get("/api/users")
	if herr != null { errText = herr; status = "error"; return }
	data, jerr = json.parse(body)
	if jerr != null { errText = jerr; status = "error"; return }
	n, terr = table.display("users", data)
	if terr != null { errText = terr; status = "error"; return }
	status = "ok"
}

view main {
	title "Users"
	onLoad: loadUsers
	h1 "Users"
	if status == "error" {
		alert "{errText}"
		button "Retry" onClick: loadUsers
	} else {
		table name: "users" {
			thead {
				trow {
					tcell "ID" head: true
					tcell "Name" head: true
					tcell "Email" head: true
				}
			}
			tbody {
				trow {
					tcell ""
					tcell ""
					tcell ""
				}
			}
		}
	}
}
```

Un risultato vuoto è uno stato `ok` valido. Aggiungi un messaggio esplicito se
la tabella vuota sembra un errore.

## CRUD form

### Remote CRUD form

Valida localmente, invia con `postSafe`, poi ricarica o aggiorna il modello
locale. Non concatenare JSON.

```sui
import { http, json } from "std"

state name = ""
state status = ""
state errText = ""

async fn create() -> null {
	let n = trim(name)
	if n == "" { errText = "Name is required"; return }
	payload, jerr = json.stringify({"name": n})
	body, herr = http.postSafe("https://api.example.com/items", payload, "https://api.example.com")
	if herr != null { errText = herr; status = "error"; return }
	name = ""
	status = "created"
}

view main {
	title "Create item"
	h1 "Create item"
	input value: name placeholder: "Name"
	button "Create" onClick: create
	if status == "error" { alert "{errText}" } else { p "{status}" }
}
```

Per update/delete controlla lo status con `http.getStatusSafe` quando l'API non
restituisce un body. Ogni status non di successo è un errore.

## Polling dashboard

```sui
import { http, json } from "std"

state status = "loading"
state rows = []
state busy = false

async fn refresh() -> null {
	if busy { return }
	busy = true
	body, err = http.get("/api/metrics")
	if err != null {
		status = "error"
	} else {
		data, perr = json.parse(body)
		if perr != null { status = "error" } else { rows = data; status = "ok" }
	}
	busy = false
}

fn boot() -> null { interval(2000, refresh) }

view main {
	title "Metrics"
	onLoad: boot
	h1 "Metrics"
	p "Status: {status}"
	if status == "ok" {
		list m in rows { p "{m}" }
	} else {
		p "Unavailable"
	}
	button "Refresh" onClick: refresh
}
```

`interval` riceve il nome di una funzione, non una lambda. Evita richieste
sovrapposte. Per WebSocket reali usa `ws.connectSafe` con allow-list letterale.

## Long job with a Go backend

Un'operazione lunga (un download, un'importazione, un batch) sta nel backend:
l'app SUI e' il pannello di controllo, non il motore. Il backend possiede lo
stato del lavoro; l'app lo interroga e mostra cosa sta succedendo.

```sui
import { http, json } from "std"

state jobId = ""
state jobState = "nessun download"
state percent = 0
state message = "Pronto."

async fn aggiorna() -> null {
	if jobId == "" { return }
	body, err = http.get("http://127.0.0.1:8090/api/downloads")
	if err != null {
		message = "backend non raggiungibile: " + err
		return
	}
	data, jerr = json.parse(body)
	if jerr != null {
		message = "risposta non valida: " + jerr
		return
	}
	let jobs = data["jobs"]
	if len(jobs) == 0 { return }
	let job = jobs[0]
	jobId = str(job["id"])
	jobState = str(job["state"])
	percent = job["percent"]
	n, terr = table.display("parti", job["parts"])
	if terr != null {
		message = "tabella parti non aggiornata: " + terr
	}
}

# L'URL dell'azione contiene l'id del lavoro, quindi e' dinamico: uso la
# variante *Safe con allow-list letterale.
async fn pausa() -> null {
	if jobId == "" { return }
	body, err = http.postSafe("http://127.0.0.1:8090/api/downloads/" + jobId + "/pause", "", "http://127.0.0.1:8090")
	if err != null {
		message = "pausa fallita: " + err
		return
	}
	aggiorna()
}

fn boot() -> null {
	interval(500, aggiorna)
}
```

- Tieni l'identificativo del lavoro in uno stato e costruisci da li' gli URL
  delle azioni: un URL dinamico richiede `http.postSafe` con allow-list
  **letterale**.
- Mostra l'avanzamento per unita' (le parti, i passi, gli elementi), non solo
  una percentuale: e' quello che dice all'operatore che il lavoro avanza.
- Conta anche le unita': quante sono, quante stanno girando adesso e quante
  sono finite. I numeri dicono con un colpo d'occhio se il lavoro avanza o se
  e' fermo.
- Lascia che l'operatore scelga il nome del file da salvare: compili la casella
  dalla URL, tieni il nome che scrive e mandalo con la richiesta. Il backend lo
  pulisce (un nome di file non e' mai un percorso) e lo usa per il file che
  scrive.
- Lascia che un download contenga piu' file. I file condividono la coda delle
  unita' e i worker del lavoro, quindi con sei worker e due file le unita'
  vengono prese tre per tre. Dai a ogni file la sua priorita' (bassa, normale,
  alta) e prendi prima le unita' dei file con la priorita' piu' alta: lo
  scheduler gira a round-robin sui file con la stessa priorita'.
- Cancella le unita' quando il file e' stato ricomposto: le parti sul disco
  sono spazio di lavoro, il file ricomposto e' il risultato.
- Il backend possiede la verita'. Pausa, ripresa e annullamento sono richieste
  al backend, mai flag locali: solo il backend sa quali unita' sono gia' fatte.
- Mostra gli errori del backend come sono. Una parte fallita e' un errore
  vero, non una riga vuota.
- Lascia ripartire da sola un'unita' fallita. L'endpoint riceve il numero
  dell'unita' e rimette in coda solo quella; l'operatore la seleziona nella
  tabella, quindi la tabella deve conservare la selezione (un refresh la
  azzera).
- Separa la vista live dai dati: barre e contatori per l'avanzamento, le
  tabelle in un'altra tab, cosi' lo schermo resta leggibile.
- Un solo pulsante per tutto il download nell'intestazione e **una riga per
  ogni file** sotto: la riga contiene l'URL, il nome del file, la priorita'
  come icona e i pulsanti di quel file (play/pausa, stop/riavvia, annulla).
  I gestori dichiarati dentro una `list` vengono collegati a **tutte** le righe
  e ricevono la chiave della riga a cui appartengono: ogni riga agisce su di
  se' e mai solo sulla prima.
- Blocca l'URL di un file mentre viene scaricato. Il pulsante di stop
  trasforma di nuovo l'URL in una casella modificabile: l'operatore scrive il
  nuovo URL e riavvia il file, e il backend lo scarica dalla nuova URL. Anche
  un file completato puo' essere fermato, cosi' il suo URL si cambia allo
  stesso modo.
- Mostra gli errori come toast (`toast.show(messaggio, "error", 5000)`) e non
  solo come riga di testo: una richiesta fallita deve essere visibile anche
  mentre l'operatore guarda le barre di avanzamento. Deduplica il messaggio
  ripetuto.
- Riproduci un suono alla fine del lavoro con `sound.play(url)`: l'URL e'
  letterale, un URL dinamico richiede `sound.playSafe` con allow-list
  letterale. Il browser puo' rifiutare la riproduzione (policy autoplay):
  mostra l'errore, non far fallire il lavoro.
- Tieni le impostazioni in un pannello sovrapposto (`modal`): modalita',
  dimensione delle parti e parallelismo appartengono alla configurazione, non
  alla vista principale.

Vedi [`examples/download_manager/`](../examples/download_manager/README.md) per
l'esempio completo: un backend Go che scarica l'URL indicato sulla riga di
comando, divide il file in parti (o lo scarica in un pezzo solo), scarica ogni
parte con una richiesta
`Range` propria, conserva i byte di una parte in pausa, riavvia da solo una
parte fallita e ricompone il file con un controllo della dimensione.

## Desktop files and commands

```bash
simpleui build app.sui --target desktop --out dist \
  --allow-dir C:/data --allow-exec helper.exe
```

```sui
import { fs, exec, json } from "std"

state path = "C:/data/input.json"
state status = ""

async fn read() -> null {
	body, err = fs.read(path)
	if err != null { status = err; return }
	data, jerr = json.parse(body)
	if jerr != null { status = jerr; return }
	n, terr = table.display("data", data)
	if terr != null { status = terr } else { status = str(n) + " rows" }
}

async fn runHelper() -> null {
	out, err = exec.run("helper.exe", [path])
	if err != null { status = err; return }
	status = out
}
```

Le build desktop espongono anche i dialoghi nativi:

```sui
import { dialog } from "std"

async fn choose() -> null {
	path, err = dialog.pickFile()
	if err != null { status = err; return }
	if path == "" { status = "annullato"; return }
	status = path
}
```

`dialog.pickFile()`, `dialog.pickFiles()` e `dialog.pickDirectory()` restituiscono
percorsi assoluti sul target desktop; sul web restituiscono un errore chiaro. Il
percorso non e una capability: `fs` richiede `--allow-dir` e `exec` richiede
`--allow-exec`. Il dialogo e guidato dall'utente; l'annullamento restituisce un
valore vuoto. Vedi [`examples/desktop_picker.sui`](../examples/desktop_picker.sui).

Valida i nomi forniti dall'utente; l'allow-list del bridge è il secondo livello,
non un sostituto. Fornisci un messaggio chiaro per il web.

## Documenti JSON

Usa JSON Pointer per indirizzare un valore, JSON Patch per modificare una copia
e `json.diff`/`json.merge3` per la sincronizzazione. I valori JSON sono mappe e
array nativi, quindi non serve un DOM.

```sui
import { json } from "std"

state status = ""

async fn update() -> null {
	let doc = {"items": [{"id": 1, "done": false}, {"id": 2, "done": true}]}
	first, perr = json.pointer(doc, "/items/0/id")
	if perr != null { status = perr; return }

	patched, err = json.patch(doc, [
		{"op": "replace", "path": "/items/0/done", "value": true},
		{"op": "add", "path": "/items/-", "value": {"id": 3, "done": false}},
		{"op": "test", "path": "/items/1/done", "value": true}
	])
	if err != null { status = err; return }

	errors, verr = json.validateAll(patched, {"type": "object", "required": ["items"]})
	if verr != null { status = verr; return }
	text, terr = json.stringify(patched, {"indent": 2, "sortKeys": true})
	status = str(first) + " / " + str(len(errors)) + " errori\n" + text
}

async fn syncFiles() -> null {
	local, lerr = json.loadFile("C:/data/local.json")
	if lerr != null { status = lerr; return }
	remote, rerr = json.loadFile("C:/data/remote.json")
	if rerr != null { status = rerr; return }
	conflicts, cerr = json.merge3(local, local, remote, {"key": "id"})
	if cerr != null { status = cerr; return }
	ok, serr = json.saveFile("C:/data/local.json", local, {"indent": 2})
	status = serr == null ? str(len(conflicts)) + " conflitti" : serr
}
```

`json.diff(a, b, {"key": "id"})` emette operazioni RFC 6902 standard (`add`,
`remove`, `replace`, `move`) applicabili direttamente con `json.patch`.
`json.merge` e RFC 7386 Merge Patch; `json.merge3` modifica `ours` e restituisce
i conflitti. `json.loadFile`/`saveFile` e `loadBytes`/`saveBytes` sono solo
desktop e richiedono `--allow-dir`. Vedi
[`examples/json_library.sui`](../examples/json_library.sui).

## Documenti binari

PDF, immagini e archivi non devono mai viaggiare come testo: `http.get` decodifica il body come UTF-8 e corrompe i byte. Tre funzioni mantengono intatto un payload binario, una per sorgente.

```sui
import { http, files, bytes } from "std"

state picked = ""
state size = 0
state magic = ""

async fn loadRemote() -> null {
	data, err = http.getBytes("https://example.com/report.pdf")
	if err != null { magic = err; return }
	size = len(data)
	head, herr = bytes.slice(data, 0, 4)
	if herr != null { magic = herr; return }
	hex, xerr = bytes.toHex(head)
	if xerr != null { magic = xerr; return }
	magic = hex
}

async fn loadPicked() -> null {
	data, err = files.readBytes(picked)
	if err != null { magic = err; return }
	size = len(data)
}

view main {
	row {
		button "Remote" onClick: loadRemote
		file accept: ".pdf" pickFiles: picked
	}
	p "bytes: {size} magic: {magic}"
}
```

- Remoto: `http.getBytes(url)`, oppure `http.getBytesSafe(url, allow)` per URL dinamici. Entrambe rifiutano uno status non 2xx e un body oltre 10 MB.
- Cartella che l'app puo leggere (desktop): `fs.readBytes(path)`; vale sempre la capability `--allow-dir` concessa a build time.
- File scelto o trascinato dall'utente (web): `files.readBytes(name)`, dove `name` corrisponde al nome o al percorso relativo del file selezionato: non legge mai un percorso.
- Validare i magic byte prima di interpretare il file e trattare ogni byte come dato non fidato.

Esempio eseguibile: [`examples/binary_bytes.sui`](../examples/binary_bytes.sui).

## Spreadsheet data

SimpleUI non legge XLSX/ODS nel browser. Converti a monte con `sheet2json`:

```bash
sheet2json data.xlsx
sheet2json data.ods -sheet 2 -format csv
sheet2json data.xlsx -list
```

```sui
import { http } from "std"

state status = "loading"

async fn load() -> null {
	body, err = http.getSafe("/api/sheet?file=data.xlsx&format=json", "/api/")
	if err != null { status = err; return }
	n, terr = table.display("data", body)
	if terr != null { status = terr } else { status = str(n) + " rows" }
}
```

Vedi [`examples/spreadsheet_table/README.md`](../examples/spreadsheet_table/README.md).

## Iframe

```sui
iframe src: "https://example.com/embed"

iframe src: videoUrl srcAllow: "https://example.com" sandbox: "allow-scripts"
```

- Il default `sandbox=""` blocca script, form, popup e same-origin.
- `allow-scripts` + `allow-same-origin` è un errore di default.
- URL malformati o non in allow-list non vengono caricati; compare
  `Cannot accept the data: ...`.
- Non usare `srcdoc` né markup di fallback.

## Visualizzatore PDF

```sui
import { fs } from "std"

state docUrl = ""

fn open() -> null {
	url, err = fs.documentUrl("docs/report.pdf")
	if err == null {
		docUrl = url
	}
}

view main {
	button "Open" onClick: open

	# Un documento remoto servito dal tuo backend.
	pdfviewer src: "https://example.com/manual.pdf"

	# Un file locale (desktop): l'URL viene dal bridge locale.
	pdfviewer src: docUrl srcAllow: "/" width: 900 height: 620
}
```

- Usa `pdfviewer`, non `iframe`: il visualizzatore PDF nativo del browser non
  parte dentro un frame sandboxato, quindi il frame del visualizzatore non è
  mai sandboxato.
- La sorgente è il confine di sicurezza: un `src` dinamico richiede una
  allow-list letterale `srcAllow`; un URL letterale non richiede nulla.
- Su desktop autorizza la cartella in fase di build (`--allow-dir`) e ottieni
  l'URL con `fs.documentUrl(path)`; il bridge serve solo file PDF veri.
- Il controllo mostra soltanto il documento: non leggerne testo, pagine o
  annotazioni e non passare i byte a un altro strumento.

Vedi [`examples/pdfviewer.sui`](../examples/pdfviewer.sui).

## Security checklist

- [ ] Ogni valore non fidato è trattato come dato, non markup.
- [ ] `style:` dinamico usa `cssSafe`; le classi non hanno significato di
      sicurezza.
- [ ] URL HTTP/WS dinamici usano `*Safe` con allow-list letterale.
- [ ] Ogni iframe ha `src` fidato o `srcAllow`; sandbox minimale.
- [ ] Ogni `pdfviewer` ha `src` fidato o `srcAllow`: il suo frame non è mai
      sandboxato.
- [ ] Errori di table/CSV/JSON/bridge controllati e mostrati.
- [ ] CSV per Excel/LibreOffice con `;formula=escape`.
- [ ] Capability desktop minimali e input validato.
- [ ] Nessun segreto in URL, log o testo della view.
- [ ] `simpleui check` pulito; build realmente eseguito.

## Verification

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --single --out app.html
simpleui build app.sui --single --minify --out app.html
simpleui serve app.sui --watch --sourcemap
```

Verifica nel browser: loading, successo, vuoto, errore di rete/JSON, URL iframe
rifiutato, validazione form e azione riuscita. Per modifiche al compilatore:
`go build -o simpleui.exe ./cmd/simpleui`, `go vet ./...`, `go test -count=1 ./...`.

