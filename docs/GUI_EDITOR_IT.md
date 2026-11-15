# Editor GUI: drag and drop

Questa guida spiega i due sistemi di drag and drop indipendenti di SimpleUI:

1. **Drag and drop del designer** - spostare elementi, voci della palette e
   sezioni della toolbar dentro l'editor visuale.
2. **Drag and drop dell'applicazione** - trascinare elementi, file o cartelle
   dentro l'applicazione generata.

Condividono le primitive del browser ma hanno regole diverse. Abilitare uno non
abilita l'altro.

Avviare il designer con un file esistente tramite `simpleui gui path/app.sui`:
l'editor apre subito quel sorgente e **Save** riscrive lo stesso percorso.
Eseguire `simpleui gui` senza file per partire dall'ultimo progetto o da uno
vuoto. Su Windows, `simpleui associate` associa i file `.sui` all'editor sotto
`HKEY_CURRENT_USER`; `--uninstall` annulla e `--dry-run` mostra i comandi di
registro.

## Risposta rapida

| Obiettivo | Cosa impostare |
|---|---|
| Spostare componenti nel designer | Niente; il DnD del designer è sempre attivo |
| Rendere trascinabile un elemento dell'app | `draggable: true` e/o `dragKey` |
| Rendere un elemento target di drop | `dropKey` più `onDrop` |
| Riordinare i dati dopo un drop | Leggere `drag.source()`/`drag.target()` e chiamare `moveItem` |
| Trascinare file/cartelle su un campo | `dropFiles: state`, più `value: state` per mostrare i percorsi |
| Aprire un dialogo file/cartella | `pickFiles: state`, con `multiple`, `accept`, `directory`, `pickMode` |
| Percorsi assoluti nativi | Build desktop più il modulo std separato `dialog` |
| Creare una barra dei menù | Palette **Layout → Navigation → Menu bar**; modificare menu e voci |
| Creare contenuti a schede | Palette **Layout → Navigation → Tabs**; tasto destro per aggiungere/rimuovere schede |
| Creare un dialogo modale | Palette **Overlay → Modal**; aggiungere/rimuovere pulsanti dall'albero |

Nell'inspector a destra la sezione **Drag & drop** usa queste etichette:
`draggable`, `dragKey` (drag key expression), `dropKey` (drop key expression),
`dropFiles` (drop files into state), `pickFiles` (choose files into state) e
`pickMode` (picker API mode). Il gestore `onDrop` si seleziona in **Events**.

## Preferenze dell'editor

Aprire **Preferences → Editor preferences.** Il pannello è diviso in tab, così
la shell dell'editor, il code editor, le scorciatoie e i file
del progetto restano separati. I tab con molti controlli usano una griglia a due
colonne (etichetta e controllo in ogni colonna), i valori numerici usano uno
slider con un riquadro valore e i tab con gruppi usano accordion che mostrano una
sezione alla volta.

Tutti i pannelli delle impostazioni sono spostabili dalla maniglia
nell'intestazione e ripartono centrati a ogni apertura. I pannelli con preview
interna mantengono il backdrop modale scuro (Preferences, Component settings,
Hints, Insert, Icon picker, File browser, New page e Desktop build). I pannelli
operativi senza preview sono flottanti, senza backdrop, così le modifiche
restano visibili nell'editor (Project preferences, Tooltip, Table settings,
Chart settings, Graph style, Structure table, Tree data, Media queries, Project
history e Chart data).

### Interface

Aspetto e layout della shell dell'editor.

| Impostazione | Default | Effetto |
|---|---|---|
| Theme | Dark | Interfaccia Dark, Light o High contrast |
| Interface font | System UI | Font di menu, pannelli, etichette e controlli |
| Custom interface font | - | Famiglia CSS usata quando è selezionato **Custom** |
| Font size | 14 | Scala l'intera interfaccia dell'editor |
| Density | Comfortable | Spaziatura comfortable o compact |
| Accent color | #4c5cf7 | Focus, selezione e azioni principali |
| Panels | tutti visibili tranne history | Mostra o nasconde Structure/Tokens, Inspector e Problems/History |
| Toolbar layout | affiancate | Le due barre superiori condividono una riga quando c'è spazio; trascina la maniglia per riordinarle |
| Reset toolbar order | - | Ripristina il layout predefinito delle toolbar |

### Code editor

Tutto ciò che riguarda la scrittura in SUI: tipografia, comportamento di editing,
colori della sintassi e tooltip dell'editor. Le impostazioni sono raggruppate in
accordion e un'unica anteprima in fondo al tab riflette tutte le scelte.
L'**Interface font** nel tab Interface e il **Code font** qui sono indipendenti:
modificare uno non cambia l'altro.

#### Font & metrics

| Impostazione | Default | Effetto |
|---|---|---|
| Code font | Monospace di sistema | Combo box con stack integrati o famiglia CSS personalizzata, solo per l'editor del codice SUI |
| Custom code font | - | Famiglia CSS usata quando è selezionato **Custom** |
| Editor size | 16px | Dimensione del codice, indipendente dallo zoom generale |
| Line height | 1.6 | Spaziatura verticale e allineamento dei numeri di riga |
| Tab size | 4 | Larghezza dei tab nell'overlay e nella textarea |
| Word wrap | off | Va a capo nelle righe lunghe in entrambi i livelli |
| Line numbers | on | Mostra il gutter |

#### Editing

| Impostazione | Default | Effetto |
|---|---|---|
| Auto-close pairs | on | Inserisce la parentesi o la virgoletta di chiusura |
| Smart indent | on | Mantiene l'indentazione su Enter e dopo una parentesi aperta |
| Auto suggestions | on | Apre il popup di completamento mentre si scrive; Ctrl+Space funziona sempre |

#### Syntax highlighting

| Impostazione | Default | Effetto |
|---|---|---|
| Syntax colors | on | Colora keyword, stringhe, numeri e commenti SUI |
| Bracket match | on | Evidenzia la parentesi corrispondente |
| Selected line | off | Sfondo opzionale per la riga dell'elemento selezionato |

Con **Syntax colors** attivo, le funzioni dichiarate e le loro chiamate sono
mostrate in giallo, mentre stati, parametri di funzione, binding `let` e
variabili dei loop sono mostrati in azzurro. Gli stessi colori vengono usati
nelle dichiarazioni, nei riferimenti e nelle interpolazioni come
`"Count {count}"`, così un nome è facile da trovare e riutilizzare. Cambia solo
il colore, quindi overlay e textarea mantengono metriche identiche.

L'editor usa font, padding, line-height e metriche dei caratteri identici
nell'overlay colorato e nella textarea trasparente, così cursore e caratteri
visibili restano allineati. I token cambiano solo colore: non vengono applicati
grassetto o corsivo al codice.

#### Editor tooltips

Tooltip che descrivono pulsanti, campi e azioni delle toolbar dell'editor. Fanno
parte dell'interfaccia dell'editor e non modificano mai l'app generata.

| Impostazione | Default | Effetto |
|---|---|---|
| Show tooltips | on | Abilita il fumetto personalizzato dell'editor |
| Delay | 450 ms | Ritardo hover prima di mostrare il fumetto |
| Place | Auto | Top, right, bottom, left o posizionamento automatico |
| Max width | 320px | Larghezza massima del fumetto |
| Font size | 12px | Dimensione del testo |
| Background / text / border | palette scura | Colori del fumetto |
| Radius | 6px | Raggio del bordo |

#### Preview

L'anteprima in fondo al tab mostra codice SUI di esempio con font, dimensione,
line-height, tab size, wrapping, colori della sintassi e riga selezionata
correnti. Mostra anche un fumetto con lo stile tooltip dell'editor. Si aggiorna
mentre modifichi i controlli, prima di **Apply**.
### Shortcuts

La tab **Shortcuts** è un riferimento in sola lettura. Il campo di ricerca filtra
per azione, tasto o gruppo. Elenca le scorciatoie di designer, editor SUI, menu,
struttura e dati dei chart; le scorciatoie non sono ancora riassegnabili.

### Files & history

Snapshot e percorsi dei file sono raggruppati in accordion.

| Impostazione | Default | Effetto |
|---|---|---|
| Automatic snapshots | on | Salva una versione del progetto durante il salvataggio |
| Interval | 60 s | Tempo minimo tra snapshot automatici, con slider e riquadro valore |
| Versions kept | 30 | Limite di conservazione automatica, con slider e riquadro valore |
| Snapshot now | - | Crea subito una versione |
| Open project history | - | Apre il pannello di ripristino |
| Clear history | - | Elimina tutte le versioni salvate dopo conferma |

La sezione **Files** mostra la cartella dati, il file delle preferenze e il file
del progetto, con un pulsante Copia per ogni percorso. Le impostazioni degli
snapshot automatici sono salvate nelle preferenze dell'editor; i file delle
versioni restano nella cartella dati.

Le preferenze vengono salvate con le altre impostazioni del designer e restano
dopo il reload. **Font size** generale scala l'interfaccia, mentre **Editor size**
controlla il codice in modo indipendente.
## Completamento e indentazione

Premere **Ctrl+Space** nell'editor SUI per aprire l'elenco dei suggerimenti
vicino al cursore. I suggerimenti cambiano in base al contesto:

| Posizione | Voci suggerite |
|---|---|
| Livello superiore | `import`, `css`, `state`, `type`, `struct`, `fn`, `async`, `view` |
| Dentro una view o un elemento | `title`, `onLoad`, tag degli elementi e view componente |
| Dentro una funzione o un blocco | `let`, `if`, `else`, `loop`, `while`, `return`, stati, funzioni e variabili locali |
| Dopo il nome di un elemento | attributi ed eventi dell'elemento |
| Dopo un attributo evento (`onClick:`) | funzioni dichiarate e **+ new function…** |
| In un'espressione o interpolazione | stati, funzioni, variabili locali, item dei loop, builtin e moduli std |
| Dopo `class:` | classi CSS dichiarate nel sorgente |
| Dopo `state nome:` o `->` | tipi SUI |

Le lettere digitate filtrano l'elenco per prefisso. Le frecce spostano la
selezione, **Invio** o **Tab** inserisce la voce, **Esc** chiude. Il popup
automatico usa lo stesso contesto e compare dopo le prime lettere; Ctrl+Space
funziona anche con la parola vuota.

Quando il designer rigenera il sorgente, il corpo delle funzioni viene indentato
di un livello dentro le parentesi graffe. Le istruzioni annidate mantengono la
loro indentazione relativa, quindi il round-trip attraverso l'editor visuale non
appiattisce il codice delle funzioni.

## Menu e navigazione del code editor

### Menu contestuale

Il tasto destro nell'editor SUI apre azioni raggruppate:

- **Undo / Redo**
- **Cut / Copy / Paste**
- **Select all / Select line**
- **Duplicate line, Move line up/down, Delete line, Indent, Outdent**
- **Go to definition, Find references, Rename symbol**
- **Find, Replace, Go to line, Command palette**
- **Toggle comment**

Il menu è contestuale: Cut e Copy sono attivi solo con una selezione, mentre le
voci di navigazione richiedono un simbolo sotto il cursore. Ogni voce mostra la
scorciatoia, si raggiunge con le frecce e si esegue con Invio; Esc chiude il menu.

### Operazioni sulle righe

| Azione | Scorciatoia |
|---|---|
| Seleziona riga | Ctrl+L |
| Duplica riga o selezione | Shift+Alt+Giù |
| Sposta riga o selezione su / giù | Alt+Su / Alt+Giù |
| Elimina riga o selezione | Ctrl+Shift+K |
| Indenta / riduci indentazione | Tab / Shift+Tab |

### Trova e sostituisci

Premere **Ctrl+F** per aprire la barra di ricerca e **Ctrl+H** per mostrare la
riga di sostituzione.

| Azione | Scorciatoia |
|---|---|
| Corrispondenza successiva / precedente | Invio / Shift+Invio, F3 / Shift+F3 |
| Distinzione maiuscole/minuscole | pulsante **Aa** |
| Parola intera | pulsante **W** |
| Espressione regolare | pulsante **Regex** (`.*`) |
| Cerca nella selezione | pulsante **Sel** |
| Sostituisci la corrispondenza corrente | **Replace** |
| Sostituisci tutte le corrispondenze | **All** |
| Chiudi la barra | Esc |

Le corrispondenze sono evidenziate nell'editor e quella corrente ha un colore più
forte; il contatore mostra posizione e totale. La ricerca parte dalla selezione
corrente quando è una selezione su una sola riga.

Le opzioni di ricerca restano memorizzate tra le sessioni. Il campo di ricerca
mantiene una cronologia delle query recenti. Con la sostituzione aperta,
un'anteprima elenca le prime sostituzioni con il numero di riga prima di usare
**Replace** o **All**. In modalità espressione regolare la sostituzione supporta
i gruppi di cattura come `$1`.

### Quick Fix e problemi

| Azione | Scorciatoia |
|---|---|
| Quick Fix sul cursore | Ctrl+. |
| Problema successivo | F8 |
| Problema precedente | Shift+F8 |

Quick Fix propone azioni per il diagnostico sotto il cursore, tra cui dichiarare
uno stato mancante, definire un handler/funzione mancante e rendere `async` la
funzione che contiene la chiamata. I diagnostici restano cliccabili nel pannello
Problems. L'analisi CSS filtra il rumore generato: classi runtime (`sui-*`,
`suitbl` e classi di stato), passi delle keyframe, blocchi `:root` stratificati e
regole hover/focus/checked o pseudo-elementi non vengono segnalati come
problemi. Le classi dell'autore realmente inutilizzate restano elencate.

### Status bar

In fondo al code editor compaiono riga e colonna del cursore, dimensione della
selezione, numero di errori e warning, dimensione dei tab e stato **Applied** o
**Modified**. Il conteggio dei problemi apre il pannello Problems; le frecce
spostano al problema precedente o successivo.

### Command palette

Premere **Ctrl+Shift+P** per cercare comandi dell'editor e simboli dichiarati.

- Digitare per filtrare per nome comando, scorciatoia, stato, funzione o view.
- Le frecce spostano la selezione; Invio esegue; Esc chiude.
- Le voci simbolo saltano direttamente alla riga di dichiarazione.
- I comandi includono Find, Replace, Go to line, Quick Fix, navigazione dei
  problemi, operazioni sulle righe, toggle comment e Apply SUI source.
### Command palette

Premere **Ctrl+Shift+P** per cercare comandi dell'editor e simboli dichiarati.

- Digitare per filtrare per nome comando, scorciatoia, stato, funzione o view.
- Le frecce spostano la selezione; Invio esegue; Esc chiude.
- Le voci simbolo saltano direttamente alla riga di dichiarazione.
- I comandi includono Find, Replace, Go to line, navigazione, operazioni sulle
  righe, toggle comment e Apply SUI source.

## Undo e redo

La toolbar e **Edit → Undo / Redo** usano un'unica cronologia per tutto il
designer, non solo per l'albero degli elementi. La stessa cronologia risponde a
**Ctrl+Z** e **Ctrl+U** per l'undo e a **Ctrl+Y** / **Ctrl+Shift+Z** per il redo.

Vengono registrate:

- le modifiche al testo SUI (una raffica di digitazione è un solo passo finché
  non viene eseguito l'apply automatico);
- testo, classi ed espressioni reattive degli elementi;
- le proprietà dell'inspector, comprese le regole CSS e le opzioni delle tabelle;
- inserimento, eliminazione, duplicazione, spostamento e incolla di componenti;
- pagine, menu, tab, modali e modifiche drag-and-drop;
- regole CSS/media e sorgente applicato dall'editor SUI.

L'undo ripristina sia il modello sia il sorgente SUI, quindi aggiorna albero,
preview e inspector. Il pannello della cronologia elenca i passi e permette di
tornare direttamente a uno stato precedente. Quando il focus è in un campo
dell'inspector, Ctrl+Z/Ctrl+U usano la cronologia del designer; i campi dei
prompt e delle modali mantengono l'undo nativo.

### Autosave

L'editor scrive una copia di recupero in `localStorage` ogni 2 secondi. La
topbar mostra **Autosaved HH:MM:SS**, **Unsaved changes**, **Saving…** oppure
**Autosave failed**. L'autosave non sovrascrive mai il file del progetto:
**Save** / **Save As** restano espliciti. Quando la copia differisce dal
contenuto corrente, la topbar mostra **Restore autosave**; il ripristino
sostituisce il contenuto e ricarica il modello dopo conferma.

### Theme toggle

Il pulsante tema nella toolbar alterna l'editor tra dark e light. La scelta
viene salvata nelle preferenze dell'editor; il select Preferences → Interface →
Theme resta sincronizzato. L'alto contrasto resta disponibile nello stesso
select delle preferenze.

## Pannello sinistro e design token

### Pannelli ridimensionabili

Trascinando il bordo destro del pannello sinistro lo si ridimensiona, esattamente
come l'inspector. Il divisore ha un'area di presa più ampia (12 px) e una linea
visibile, quindi è più facile da afferrare. La larghezza resta tra 190 e 600 px
e viene ricordata tra i reload. Il trascinamento usa il pointer capture, quindi
il ridimensionamento resta continuo anche quando il puntatore attraversa
l'iframe di anteprima.

Il pannello **Problems/History** in basso si ridimensiona trascinando il bordo
superiore. L'altezza resta tra 90 px e il 65% della finestra, viene ricordata tra
i reload e un doppio click sul divisore ripristina i 190 px predefiniti.

### Etichette dell'albero elementi

Ogni riga dell'albero mantiene il chip del tipo e aggiunge un nome leggibile e un
chip con il codice. Il nome viene dal testo o dal titolo dell'elemento, dal
placeholder, dall'attributo `name` o dalla prima classe. Il chip del codice mostra
il `motionKey` letterale quando presente, altrimenti il percorso nell'albero (per
esempio `0.2.3`). Passando sulla riga si vedono nome, key e percorso completi.

Usa il `motionKey` o il percorso per distinguere card, righe e controlli ripetuti.
Lo stesso codice è disponibile nel picker **event source** dei trigger.

### Design token

La tab **Tokens** elenca i design token presenti nel CSS del progetto. Un design
token è una custom property CSS come:

```css
:root {
  --primary: #4c5cf7;
  --radius: 10px;
  --page-gap: 16px;
}
```

Non sono i token di sintassi usati per colorare il codice SUI. I design token
sono valori riutilizzabili per colori, spaziature, raggi, font e scelte di
design simili. Si usano con `var(--primary)` nelle regole di classe, in
`style: cssSafe(...)` o negli override del tema. Modificando un token si
aggiornano tutte le regole che lo usano.

La tab raggruppa i token in **Colors**, **Typography**, **Spacing**, **Radius**
e **Other**, con un campo di ricerca. Ogni riga è modificabile inline:

- cambiare il nome rinomina il token; la dichiarazione e tutti i riferimenti
  `var(--old)` nel CSS vengono aggiornati;
- cambiare il valore direttamente, oppure usare il color picker per i token
  colore;
- **C** copia `var(--name)`;
- **U** applica `var(--name)` a una proprietà CSS dell'elemento selezionato;
- **D** duplica il token;
- **x** elimina il token (se usato, le occorrenze vengono prima sostituite con
  il suo valore).

Il modulo in alto aggiunge una nuova custom property a `:root` senza prompt del
browser. Nell'inspector, il piccolo pulsante **T** accanto ai campi colore e
lunghezza salva il valore corrente come token e sostituisce il campo con
`var(--name)` in un unico passo annullabile. Usare i token per valori condivisi
nell'interfaccia e valori espliciti per le dichiarazioni una tantum.

## Default dei componenti di progetto

Aprire **Preferences → Project preferences.** per impostare i default di ogni
istanza di un componente nel progetto. Il pannello ha un combo **Theme** e un
campo **Name** sopra le tab delle categorie; entrambi seguono il tema
dell'editor. Le tab sono **All**, **Feedback**, **Layout**, **Form**,
**Navigation**, **Data**, **Overlay**, **Charts** e **General**. I componenti
sono raggruppati per categoria in card: la tab **All** dispone le card in una
griglia a due colonne, mentre una singola categoria usa una card a larghezza
piena. Ogni componente resta un pulsante normale dell'editor con il solo nome
del controllo e apre il suo sottopannello. Il combo Theme applica uno stile
condiviso a tutti i controlli in una sola volta, incluse le scrollbar, invece di
configurare ogni componente a mano. Scegliendo **Custom** i valori correnti restano invariati.

Ogni componente apre lo stesso pannello di configurazione usato dalle
impostazioni dell'elemento. Il pannello segue il layout delle preferenze
dell'editor:

- categorie in tab quando ci sono più gruppi;
- layout a due colonne etichetta/controllo;
- slider con campo numerico per i valori numerici;
- switch per le opzioni on/off;
- combo box per font e scelte;
- icona di chiusura nell'intestazione e azioni Reset/Cancel/Apply.

I valori di progetto sono salvati come custom properties CSS in un blocco
gestito `/* sui-component-defaults */` e consumati da un blocco di regole per
componente. Applicando un componente dalle preferenze di progetto vengono
scritti entrambi i blocchi e la preview viene aggiornata. **Tooltips** nello
stesso pannello apre le impostazioni globali dei tooltip; chiudendole si torna
alla lista di progetto. Il pannello dei tooltip segue il tema di progetto: un
tema chiaro usa un fumetto chiaro e un tema scuro un fumetto scuro, e offre
anche un campo **Letter spacing**.

### Temi di progetto

Ogni stile condiviso ha una versione **Light** e una **Dark**. Il combo Theme le
elenca raggruppate per famiglia, per esempio Sketch Light/Dark, Terminal / CRT
Light/Dark e Geometric Light/Dark. Applicando un tema vengono scritti tutti i
default di tutti i controlli, aggiornati i blocchi CSS gestiti, applicate le
regole delle scrollbar e impostato il **punto di partenza della pagina**
(`--sui-bg`, `--sui-fg`, bordi e hover di base). Un tema chiaro parte quindi da
un body chiaro e un tema scuro da un body scuro. Anche gli elementi base seguono
le stesse variabili: superficie della view, titoli, link, blocchi di codice,
intestazioni delle tabelle, menu bar e pulsanti semplici cambiano con il tema.
Anche tooltip, immagini, video, audio e iframe usano la palette della pagina, e i
preset chart/graph dei temi usano gli stessi nomi delle famiglie. La
famiglia **Standard** offre **Standard Light** e **Standard Dark**, che mantengono
la forma di default di ogni controllo. Modificando poi un componente il combo
Theme passa a **Custom**, quindi e sempre chiaro se il progetto segue ancora un
unico tema. Un nuovo
progetto parte da **Standard Dark**; scegliendo **Standard Light** passano allo
stile chiaro sia la pagina sia i controlli.

I temi condivisi vengono controllati anche per leggibilita: testo di pagina e
pannelli, testo attenuato, link, pulsanti, badge, alert, card, tab, accordion,
label dei form, tree, tabelle, toast, modali e menu rispettano soglie di
contrasto rispetto alle proprie superfici. I link usano il token dedicato
`--sui-link`, derivato dall'accent e scurito o schiarito quando serve, cosi un
accent adatto come riempimento non diventa illeggibile come testo. Anche le tab
attive e i colori hover derivati vengono corretti allo stesso modo, mentre i
bordi dei componenti restano visibili sulle superfici chiare di Glassmorphism e
Standard.

### Esportare e importare un tema

**Export theme** nelle preferenze di progetto scarica un file
`simpleui-theme.json` con la palette della pagina, tutti i default dei
componenti, le impostazioni dei tooltip e i default di progetto di chart e
graph. Il campo **Name** imposta il nome del tema e del file scaricato.
**Import theme** ricarica quel file dopo una validazione rigorosa: campi
sconosciuti, nomi di custom property non validi e valori CSS che potrebbero
aggiungere dichiarazioni o regole vengono rifiutati. L'import sostituisce il tema
corrente, riapplica i default chart/graph importati ai chart e graph gia
esistenti, ripristina il nome del tema e aggiorna la preview; l'undo ripristina i
valori precedenti.

### Scrollbar

Le scrollbar non si inseriscono come elementi normali: compaiono quando un
contenitore scorre. L'aspetto si configura da **Project preferences → General →
Scrollbars**. Il pannello controlla dimensione, raggio del thumb, larghezza
Firefox, track, thumb, hover del thumb e bordo. I temi condivisi impostano la
palette delle scrollbar in modo coerente con gli altri componenti e l'anteprima
del pannello contiene un esempio scorrevole.

### Default di progetto per chart e graph

**Project preferences → Charts** contiene **Chart defaults** e **Graph
defaults**. Aprono lo stesso pannello unificato di chart/graph usato
dall'inspector, con ambito di progetto. Applicando vengono scritti aspetto e
interazione su ogni chart o graph del progetto e su quelli inseriti in seguito:
tema, preset, dimensione punti, tooltip, tabella, Gantt, link e interazioni. I
dati e il tipo restano per singolo elemento, e la tab Data è
nascosta in questa modalità. I default sono salvati in un commento gestito
`/* sui-chart-defaults ... */` nel CSS di progetto.

### Override dell'elemento

Selezionare un elemento e premere **Settings...** nel banner dell'inspector. Si
apre lo stesso pannello con ambito **This element only**. I valori vengono
scritti come custom properties inline su quell'elemento, quindi tutte le altre
istanze mantengono i default di progetto. **Clear overrides** rimuove i valori
inline e ripristina i default di progetto per quell'elemento. Chiudendo il
pannello si torna all'inspector.

Pannelli componente implementati: **Progress bar**, **Badge**, **Alert**,
**Spinner**, **Card**, **Divider**, **Images & media**, **Button**,
**Input / Textarea / Select**, **Combobox**, **Date / Time picker**,
**Checkbox / Radio**, **Switch**, **Slider**, **Tabs**, **Accordion**,
**Menu**, **Tree**, **Table**, **Toast**, **Modal** e **Scrollbars** (globale).
Chart e Graph mantengono i loro inspector specializzati.

Ogni pannello mostra un'anteprima dal vivo sopra i campi. L'anteprima usa i
valori in modifica, non quelli salvati, quindi colori, dimensioni e stati si
possono verificare prima di Apply.

I controlli con testo espongono un campo **Letter spacing**. I controlli
interattivi hanno una tab **Hover** con i colori usati mentre il puntatore è
sopra il controllo: sfondo, testo, bordo, opzione, riga o thumb a seconda del
componente. In questo modo lo stato hover fa parte dello stesso tema.
Anche **Progress bar**, **Spinner** e **Divider** espongono una tab **Hover**,
quindi i controlli non interattivi possono comunque reagire al puntatore.

Il pannello **Images & media** applica il tema a immagini, video, audio e
iframe: sfondo, bordo, raggio, ombra, object fit e padding. Partecipa ai temi di
progetto e agli override del singolo elemento come gli altri pannelli.

### Preset grafici e animazioni

Ogni pannello componente inizia con un selettore **Style**. **Custom** e il
valore predefinito grezzo, mentre ogni preset con nome applica in un solo passo
un insieme coordinato di colori, forme ed effetti. I preset sono opzionali: non
vengono mai usati se non selezionati, e modificando un campo il selettore passa
a **Custom** per rendere esplicita la personalizzazione. Ogni famiglia condivisa
esiste in versione **Light** e **Dark**, per 24 varianti condivise. Il selettore
Style le elenca sotto **Shared styles**; il combo Theme di progetto seleziona le
stesse varianti per tutti i componenti in una volta.

Sopra l'anteprima, **Preview background** passa tra **Dark** e **Light**. La
scelta viene ricordata e permette di controllare il componente sia su una shell
scura sia su una pagina chiara o simile alla carta. Selezionando uno stile
condiviso viene scelto automaticamente anche uno sfondo consigliato (chiaro per
Sketch e Brutalist, scuro per gli altri); i pulsanti permettono comunque di
controllare l'altro sfondo.

I preset coprono tutti e ventuno i pannelli componente, per esempio **Neon glow**
e **Barber pole** per Progress, **Dual ring**, **Gradient ring** e **Coin flip**
per Spinner, **Glass** per Card e Toast, **Pill** per Button e Input, e
**Striped**, **Borderless** e **Rounded** per Table. Diversi preset includono
un effetto di movimento; la scheda Animation/Effects espone gli effetti anche
direttamente:

- Progress: pulse, strisce scorrevoli, shimmer e glow.
- Spinner: rotazione classica, pulse e flip verticale.
- Badge: pulse, glow e float.
- Alert: fade in, slide in, shake e pulse.
- Card: fade in e float.
- Button: pulse, float e glow.
- Toast: discesa, fade, zoom, salita e rimbalzo.

La famiglia **Sketch** ha versioni Light e Dark, condivise da tutti e venti i
pannelli, cosi i componenti sulla stessa pagina restano visivamente coerenti. La
versione chiara usa carta calda; quella scura usa inchiostro crema su carta
scura. Entrambe usano bordi disegnati, ombre offset, un font calligrafico e una
selezione evidenziatore dove serve. Ogni colore, font, raggio e ombra resta
modificabile, quindi la famiglia si puo adattare senza perdere il carattere
disegnato a mano.

Anche la famiglia **Wireframe** ha versioni Light e Dark. Rimuove i riempimenti e
disegna ogni controllo solo con linee sottili, bordi tratteggiati e ombre morbide.
Gli stati attivi usano una linea ciano; colori, font, raggio e ombre restano
modificabili, quindi una pagina puo combinare piu componenti Wireframe mantenendo
lo stesso linguaggio a contorno.

Il selettore Style raggruppa queste famiglie sotto **Shared styles** e lascia gli
stili specifici di un componente sotto **Component styles**. Le altre famiglie
condivise, ciascuna con varianti Light e Dark, sono:

- **Terminal / CRT**: fosfori verdi su nero, etichette monospace e glow.
- **Glassmorphism**: superfici traslucide fumose con blur e bordi luminosi.
- **Brutalist**: contorni neri spessi, blocchi gialli e ombre offset dure.
- **Blueprint**: carta blu profonda, linee ciano tratteggiate e font monospace.
- **Neumorphic**: superfici morbide con ombre chiare e scure accoppiate.
- **High contrast**: nero, bianco, giallo e focus ciano per l'accessibilita.
- **Editorial / Luxury**: navy profondo, crema, filetti oro e carattere serif.
- **Cyberpunk**: nero cromato, linee neon magenta e ciano con glow.
- **Geometric**: quadrati netti, angoli tagliati e motivi a triangoli/quadretti.

Nelle impostazioni dell'elemento il selettore offre anche **Project default**.
Applicandolo si rimuovono gli override componente dell'elemento, che torna a
ereditare i default di progetto. I preset con nome, invece, fissano i loro
valori su quell'elemento. I preset di progetto vengono scritti nei blocchi
gestiti `/* sui-component-defaults */` e `/* sui-component-rules:<id> */`,
mentre i preset dell'elemento usano custom properties inline.

### Inserire qualsiasi componente

La toolbar in alto ha un pulsante **Insert component**. Apre un elenco
ricercabile di tutti i tag componente supportati dal designer, inclusi quelli
annidati nei sottomenu (Progress, Badge, Alert, Spinner, Combobox, Date/Time
picker, Tree, Chart, Graph e così via). Selezionando un componente viene
inserito nella posizione corrente.

## Inspector e stili effettivi

Tutte le sezioni dell'inspector partono chiuse. Aprire una sezione cliccando
sull'intestazione; resta aperta tra cambi di selezione e reload della preview
finché non viene richiusa.

Ogni intestazione di sezione ha una piccola maniglia a sinistra. Trascinala per
riordinare le sezioni dell'inspector; l'ordine viene salvato nel local storage
dell'editor e ripristinato al reload. Le sezioni che non si applicano
all'elemento corrente scompaiono temporaneamente e mantengono la posizione
salvata quando ricompaiono. Il riordino cambia solo la UI dell'editor, non
l'applicazione generata. Anche la sezione **View** (titolo pagina, nome pagina,
onLoad e azioni pagina) è una normale sezione riordinabile e, come tutte le
altre, parte chiusa.

Quando una modifica ridisegna l'inspector vengono ripristinati la sezione
aperta, la posizione di scorrimento e il campo con il focus. Puoi continuare a
modificare lo stesso parametro senza dover ritrovare la sezione.

Ogni riga dell'albero **Structure** a sinistra inizia con un badge del tipo
esatto (`div`, `tree`, `treeitem`, `chart`, `input`, ...). Il badge e separato
dall'etichetta descrittiva, quindi il tipo dell'elemento selezionato e sempre
inequivocabile. L'inspector ripete il tipo esatto in un banner **Selected
element** sopra le sezioni.

I campi di stile mostrano il valore che raggiunge realmente l'elemento
selezionato, non solo una dichiarazione della prima classe. La preview riporta i
valori dichiarati e quelli calcolati, quindi l'inspector popola i campi anche
per:

- regole su selettori composti come `table.suitbl` o `div.card`;
- valori ereditati, incluso il colore del testo;
- attributi `style` inline;
- variabili CSS e regole base delle tabelle.

I valori dichiarati vengono mantenuti quando disponibili, quindi
`width: 100%` resta `100%`; quando il valore deriva solo dal layout viene usato
quello calcolato (per esempio `400px`). Larghezza, altezza, colori, tipografia,
spaziature, bordi, layout e gli altri controlli seguono la stessa regola.
Modificare un campo continua a scrivere la regola appropriata per la classe
selezionata (o per il contenitore padre per le proprietà di layout dei
contenitori).
La proprietà **required** è disponibile per `input`, `textarea`, `select`,
`combobox`, `radio`, `number`, `date`, `time` e `datepicker`. I controlli nativi
usano l'attributo `required` del browser; il combobox aggiunge un errore inline
quando non è selezionato un valore.

### Ombre, filtri ed effetti di fusione

La sezione **Effects** usa card compatte invece di un'unica riga affollata:

- **Box shadow**: preset (`none`, `soft`, `raised`, `crisp`, `inset`), spin box
  X/Y/blur/spread in `px`, alpha in `%`, toggle inset, color picker del progetto
  e campo **CSS** raw. L'alpha di `rgba()` viene conservato.
- **Filter**: blur (`px`), brightness/contrast/saturation/grayscale/sepia/invert
  (`%`) e hue (`°`) con preset e campo CSS raw. Le funzioni sconosciute come
  `drop-shadow()` o `url()` restano nel valore raw e vengono preservate quando
  si modifica un controllo noto.
- **Backdrop filter**: preset glass, frost e tint con controlli blur,
  saturazione, luminosità e contrasto; l'editor scrive sia la proprietà standard
  sia `-webkit-`.
- **Blend mode**: `mix-blend-mode`, da `normal` a `luminosity`.

La sezione **Typography** usa la stessa card per `text-shadow` (senza spread e
inset). Il campo CSS raw è la via per ombre a più livelli o filtri avanzati; un
preset o un controllo numerico sostituisce un'ombra multi-livello con un solo
livello.

### Overlay per stili avanzati

Le proprietà meno comuni sono raggruppate dietro pulsanti nella sezione a cui
appartengono. Ogni pulsante apre un overlay trascinabile; l'overlay resta aperto
mentre si modificano i valori e resta sincronizzato con l'editor CSS.

- **Background** → **Advanced background...**: image, size, position, repeat,
  attachment, clip, origin e blend.
- **Layout** → **Advanced flex...**: flex-wrap, order e grow/shrink/basis.
- **Grid** → **Advanced grid...**: auto flow/rows/columns e justify-items/self.
- **Typography** → **Advanced text...**: text-overflow, overflow-wrap,
  word-break, writing-mode e direction.
- **Size** → **Overflow and sizing...**: overflow, box-sizing, aspect-ratio,
  min-width/min-height e object-fit.
- **Effects** → **Transform and clip path...**: transform, transform-origin,
  clip-path e isolation.
- **Border** → **Outline...**: outline style, width, colore e offset.
- **Interaction**: cursor, pointer-events, user-select, visibility, float,
  clear e scroll-behavior. Queste proprietà agiscono sulla regola di classe
  selezionata. Cursor, pointer-events, user-select e visibility sono ereditate,
  quindi su un contenitore influenzano anche i discendenti; per un solo elemento
  usa **Appearance → style expr**.

L'overlay si chiude con **×** o **Escape**; selezionando un altro elemento si
chiude automaticamente.

### Advanced CSS

Apri **Appearance → CSS declarations → Edit...**. L'overlay elenca le
dichiarazioni della regola di classe selezionata. Ogni riga ha proprietà in sola
lettura, valore modificabile e pulsante di eliminazione. La riga di aggiunta
accetta qualsiasi proprietà CSS valida, comprese le custom properties
(`--accent: #4c5cf7`), con suggerimenti e validazione. Esempi:

```css
overflow: hidden;
aspect-ratio: 16 / 9;
transform: rotate(2deg);
--accent: #4c5cf7;
```

È una seconda vista dello stesso modello CSS usato dall'editor CSS: le modifiche
aggiornano subito `root.css`, il sorgente generato e l'editor CSS. L'overlay
modifica la regola di classe, non i valori inline effettivi; per dichiarazioni
inline reattive usa **style expr**. Poiché un valore arbitrario può contenere
`url(...)`, immagini e font possono caricare risorse esterne. Le proprietà con
prefisso vendor sono ammesse; modificando una proprietà con una entry di
compatibilità nota vengono aggiornate anche le varianti prefissate.

## Drag and drop del designer

Il drag and drop interno del designer è sempre attivo. Non esiste un'impostazione,
una proprietà o un flag di build per abilitarlo. La sezione **Drag & drop**
dell'inspector configura l'applicazione generata; non controlla lo spostamento
dei componenti nell'editor.

| Azione | Come |
|---|---|
| Inserire un componente | Trascinare una voce della palette nel preview o nell'albero |
| Riordinare un componente | Trascinare una riga dell'albero sopra/sotto un'altra |
| Cambiare padre a un componente | Trascinare una riga dentro una riga contenitore |
| Riordinare le sezioni della toolbar | Trascinare la maniglia; le due barre stanno affiancate quando c'è spazio |
| Ridimensionare righe/colonne di una tabella | Trascinare il bordo della tabella |

L'editor usa il modello interno degli elementi come fonte di verità. Un drop
riuscito crea uno snapshot per l'undo, aggiorna l'albero, ridisegna il preview e
rigenera il sorgente `.sui`.

### Drop nell'albero

Le righe dell'albero trasportano il percorso dell'elemento nel `dataTransfer`
come testo. Durante `dragover` l'editor evidenzia il possibile target. Al
`drop`:

1. rifiuta un drop dentro il sottoalbero dell'elemento stesso;
2. rimuove il nodo sorgente dal vecchio padre;
3. lo inserisce nella posizione del target o dentro il contenitore target;
4. registra un passo di undo.

Le righe che contengono altri elementi mostrano una freccia di espansione.
Cliccarla per comprimere o espandere il sottoalbero senza cambiare la
selezione; lo stato resta finché il progetto è aperto. Le righe foglia mostrano
uno spazio vuoto equivalente, così le etichette restano allineate.

Se il drag è ambiguo, usare l'**albero**, non il preview. L'albero mostra in modo
esplicito le relazioni padre/figlio ed è il target più affidabile.

### Drop dalla palette

La palette memorizza un tipo temporaneo come `palette:button`. Il preview
intercetta il drop, invia all'editor l'id dell'elemento target e l'editor
inserisce il nuovo componente:

- dentro un contenitore, come figlio;
- dopo un elemento foglia, come fratello successivo;
- nel body della pagina selezionata quando non c'è un target.

Il nuovo elemento viene selezionato dopo l'inserimento.

### Limiti del designer

- Il drop di un nodo dentro il proprio sottoalbero è bloccato intenzionalmente.
- Alcune foglie (`button`, `input`, `img`, `text`, `br`, ...) non possono
  contenere figli; il drop viene fatto accanto a loro.
- Il preview è un iframe. Se il browser non consegna i drop al suo interno,
  usare l'albero.
- Il drag and drop del designer è una funzione di editing. Non rende
  trascinabile l'applicazione generata.

## Drag and drop dell'applicazione

Il drag and drop dell'applicazione si configura per singolo elemento. Non esiste
una proprietà globale "abilita drag and drop". Un'interazione funzionante di
solito richiede:

- un attributo sul **sorgente**;
- un attributo sul **target**;
- un **gestore evento** sul target;
- stati/funzioni dichiarati per l'operazione.

### Proprietà necessarie

| Proprietà | Dove | Scopo |
|---|---|---|
| `draggable: true` | Sorgente | Marca l'elemento come sorgente HTML |
| `dragKey: expr` | Sorgente | Chiave/payload esposto da `drag.source()` |
| `dropKey: expr` | Target | Chiave esposta da `drag.target()` |
| `onDrop: handler` | Target | Funzione chiamata su un drop valido |
| `dropFiles: state` | Target file | Riceve automaticamente i percorsi di file/cartelle trascinati |
| `pickFiles: state` | Trigger dialogo | Apre il dialogo nativo per file/cartelle |
| `multiple: true` | Picker | Permette più di un file |
| `accept: ".csv,.txt"` | Picker | Filtra i file selezionabili |
| `pickMode: "native" \| "fs" \| "auto"` | Picker | Sceglie input del browser, File System Access API o fallback automatico |
| `directory: true` | Picker | Seleziona una cartella invece di singoli file |

`dropKey`/`onDrop` servono per trascinare elementi e dati dentro la pagina.
`dropFiles` serve per file e cartelle trascinati dal sistema operativo e non
richiede `onDrop`.

Tutti i parametri drag and drop sono raggruppati nella sezione comprimibile
**Drag & drop** in fondo all'inspector, dopo le sezioni di stile. Fare clic
sull'intestazione per espanderla o comprimerla. I gestori `onDrag*` e `onDrop`
restano nella sezione **Events**.

Eventi supportati dal runtime dell'applicazione:

```text
onDragStart  onDragEnter  onDragOver  onDragLeave  onDrop  onDragEnd
```

I gestori non hanno argomenti. Lo stato del drag si legge con i builtin:

```text
drag.source()  drag.target()  drag.data()
drag.files()   drag.folder()  drag.error()
drag.active()  drag.clear()
```

### Esempio minimo di riordino

```sui
state items = ["Design", "Build", "Ship"]

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
}
```

`dragKey: item` da solo rende la riga trascinabile. Il `draggable: true`
esplicito rende l'intenzione visibile nel sorgente e nel designer.
`draggable: true` senza `dragKey` non invia una chiave utile, quindi il target
non può identificare il sorgente. Il target deve avere sia `dropKey` sia
`onDrop`; una chiave senza gestore non fa nulla.

### Passi nel designer per l'esempio di riordino

1. Selezionare la riga ripetuta nell'albero o aprire il suo sorgente nel code
   view. Le proprietà sono nell'inspector a destra.
2. In **Drag & drop** impostare:
   - `draggable` attivo;
   - `dragKey` a `item`;
   - `dropKey` a `item`.
3. In **Events**, scegliere `onDrop` e creare/selezionare `reorder`.
4. Implementare `reorder` nel pannello funzioni/stati.
5. Nel preview, trascinare una riga su un'altra.

Per le righe ripetute di una `list` e per le celle delle tabelle il designer
potrebbe non esporre ogni attributo per riga. In quel caso aggiungere gli
attributi direttamente nel sorgente `.sui`. Il runtime emette le callback DnD
inline per le righe ripetute, quindi funziona ogni riga, non solo la prima.

## Drop di file e cartelle

`dropFiles` è il target automatico per file e cartelle. Richiede uno stato
stringa:

```sui
state attachment = ""

view main {
	input "Allegato" value: attachment dropFiles: attachment
}
```

Lo stato riceve i percorsi trascinati. Abbinare un binding `value:` se il
controllo deve mostrarli. Per più file un input usa percorsi separati da virgola;
una textarea usa righe separate. L'inspector crea automaticamente lo stato
stringa quando si inserisce un nuovo nome in `dropFiles` o `pickFiles`; nel code
editor bisogna dichiararlo.

Il drop di cartelle è ricorsivo dove il browser espone le directory. La
selezione popola anche:

- `drag.files()` - metadati (`name`, `path`, `size`, `type`, `isDirectory`);
- `drag.folder()` - nome della cartella comune di primo livello, se disponibile;
- `drag.error()` - motivo del rifiuto;
- `drag.active()` - se un drag è in corso.

Limiti e sicurezza:

- 2.000 file, 16 livelli di directory e 4.096 caratteri per percorso;
- i nomi sono dati non fidati e vengono renderizzati con l'escaping normale;
- il gestore del drop non legge i byte dei file;
- `http.postFiles` può caricare i byte selezionati verso un backend;
- un percorso trascinato non è il permesso di leggere il percorso locale
  originale.

Vedere [Drag and drop nell'API reference](API_REFERENCE_IT.md#drag-and-drop) e
l'[esempio file analyzer](../examples/file_analyzer/README.md).

## Menu bar control

La palette contiene un controllo pronto per la barra dei menù in
**Layout → Navigation**:

- **Menu bar** inserisce un `menubar` con i menu File, Edit e Help e voci di
  esempio;
- **Menu** inserisce un menu di primo livello con due voci di esempio;
- **Menu item** inserisce una voce dentro il menu selezionato.

### Modifica dall'albero degli elementi

Fare clic con il tasto destro su una riga dell'albero. Il menu contestuale è
l'editor dei menu:

| Riga | Azioni |
|---|---|
| `body` | **Add menu bar** |
| `menubar` | **Add menu**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `menu` | **Add menu item**, **Add submenu**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `menuitem` | **Add item below**, **Add submenu below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| altro elemento | **Duplicate**, **Move up/down**, **Delete** |

**Add submenu** inserisce un `menu` annidato dentro il menu selezionato.
L'albero lo mostra indentato; il suo menu contestuale offre ancora **Add menu
item** e **Add submenu**, quindi l'annidamento non ha limiti. **Add menu** è
disponibile sul `menubar` e crea un menu di primo livello con due voci iniziali.

### Configurare etichette e azioni

1. Inserire **Menu bar**. L'albero mostra `menubar` con figli `menu` e
   `menuitem`.
2. Selezionare un `menu` e modificare **Text** per cambiare l'etichetta
   visibile.
3. Selezionare un `menuitem` e modificare **Text** per la sua etichetta.
4. Per un'azione, impostare **onClick** in **Events** su una funzione senza
   argomenti.
5. Per la navigazione, impostare **link** nell'inspector dell'elemento. I link
   passano dal normale URL guard; gli URL dinamici non fidati richiedono
   comunque validazione.
6. Aggiungere/eliminare/riordinare le voci con il menu contestuale dell'albero.
   Funziona anche il trascinamento di **Menu** o **Menu item** dalla palette.
7. Usare `divider` dentro un menu per un separatore.
8. Personalizzare le parti generate con le classi `.sui-menubar`, `.sui-menu`,
   `.sui-menu-button`, `.sui-menu-popup` e `.sui-menuitem`.

Il campo **Text** del `menubar` è il nome accessibile (`aria-label`), non un
titolo visibile. Il **Text** del `menu` è l'etichetta del pulsante e il **Text**
del `menuitem` è l'etichetta della voce.

### Comportamento

- Il click su un pulsante menu/sottomenu apre/chiude il popup; il click fuori o
  Escape chiude l'intero albero dei menu.
- Le frecce spostano il focus tra le voci di un menu aperto; destra/sinistra
  entrano ed escono da un sottomenu.
- Un `menuitem` può avere `onClick`, `link` o entrambi.
- Un `menu` annidato nel popup di un altro `menu` è un sottomenu; l'annidamento
  non ha limiti.
- Le voci di un popup chiuso sono nascoste nel preview. Selezionarle
  nell'albero; fare clic sul pulsante del menu nel preview per vedere il popup.
- Le etichette sono escapate e gli URL dei link sono sanitizzati dal runtime.
- La struttura generata è semantica: `nav[role=menubar]`, pulsanti menu e voci
  `[role=menuitem]`.

Vedere [API reference: menu bar](API_REFERENCE_IT.md#menu-bar) e l'[esempio menu
bar](../examples/menu_bar.sui).

## Controlli dropdown e popover

**Controls > Dropdown menu** inserisce un pulsante con un elenco ancorato di
figli `dropdownitem`. Imposta l'etichetta da **Text**, l'allineamento e il
`placement` preferito nella sezione **Dropdown**; aggiungi le voci dall'albero.
Le voci supportano `onClick`, `disabled` e `separator`. Il menu si ribalta e
resta dentro la viewport.

**Controls > Popover** inserisce un trigger con contenuto ricco in un pannello
ancorato. Imposta l'etichetta del trigger da **Text** e il lato preferito nella
sezione **Popover**. Si chiude con un click fuori o con Esc. Il posizionamento
auto mantiene il pannello nella finestra.

**Controls > Drawer** inserisce un pannello laterale. Imposta il titolo da
**Text**; la sezione **Drawer** controlla il lato (`left`/`right`), la larghezza,
lo stato `open`, una funzione `onClose` opzionale e `duration`/`easing` della
transizione. Il campo dello stato crea lo stato se manca.

### Impostazioni animazione

Gli inspector di dropdown, popover, combobox e command palette espongono
l'effetto **animation** (`fade`, `scale`, `slide`, `none`), `duration`, `easing` e
`delay`. Drawer e carousel espongono `duration`/`easing`; skeleton espone
`duration` per lo shimmer. Gli stessi campi sono disponibili come attributi nel
sorgente e sono mostrati in anteprima dal designer.

### Struttura, tema e motion dei toggle

Selezionando **checkbox**, **switch** o **radio** si apre **structure & motion**.
Il selettore **structure** sceglie un preset di forma, **theme** un preset di
colore e **animation** sceglie `auto`, `morph`, `fade`, `scale`, `slide` o `none`.
`duration`, `easing`, `delay` e `stagger` regolano la transizione. Forma e tema
sono preset di variabili CSS; il runtime di motion interpola ogni parte tra i due
stati, quindi thumb, track, spunta e bordi morfano invece di saltare. Questi
attributi possono essere legati a stati e cambiati a runtime.

### Motion di layout (FLIP)

Selezionando una card, una voce di lista, un pannello o un contenitore si apre
**Motion (FLIP / enter / exit)**. Imposta l'effetto (`layout`, `fade`, `scale`,
`slide` o off), una **chiave** stabile, durata, easing, ritardo e se animare
l'uscita. Il runtime abbina gli elementi per chiave tra i render, li FLIPpa nelle
nuove posizioni e dissolve elementi mancanti/nuovi. Le voci di lista e i
componenti ripetuti devono impostare `motionKey` (ad esempio `item.id` o il
parametro del componente); `data-sui-id` non è univoco dentro una lista. Tabelle,
tree, chart e toggle hanno sezioni motion dedicate.

Nelle app multi-pagina, `navigate(page)` cattura la pagina visibile prima del
cambio e applica le chiavi condivise dopo il render della nuova pagina. Lo stesso
`motionKey` su due pagine produce una transizione shared-element, incluse
dimensioni e posizione. Il runtime misura i rettangoli iniziale e finale della
card e delle sue parti visive dirette, ingrandisce un livello di superficie senza
contenuto e trasla ogni parte in linea retta verso la posizione finale misurata;
il contenuto viene ritagliato al box che cresce invece di essere riflussato.

### Animation panel

Seleziona un elemento e apri **Motion > Open animation panel**. Il pannello
contiene solo il grafo; i parametri si aprono in pannelli overlay trascinabili:

- **Layout & shared** nella toolbar apre l'overlay per FLIP/liste/pagine.
  **Exit animation** apre l'overlay di uscita: scegli lo stato di destinazione,
  durata, easing e delay. **Spring easing** aggiunge stiffness, damping e mass;
  lo spring sostituisce l'easing e la durata ne controlla la velocità.
  **Preview exit** riproduce la transizione verso lo stato di uscita senza
  rimuovere l'elemento; **Stop preview** la annulla. Il runtime anima un clone
  fisso quando SUI rimuove l'elemento.
- Il pannello è compatto e ha un handle di ridimensionamento nell'angolo in
  basso a destra. Il canvas centra e adatta il diagramma; la toolbar ha i
  controlli **&#8722;**, **+** e **Fit** con la percentuale di zoom. Trascina lo
  spazio vuoto del canvas per il pan e usa la rotella per lo zoom attorno al
  puntatore. Pannello e overlay restano dentro il viewport e possono essere
  trascinati o ridimensionati.

- I blocchi sono piccoli ma crescono in altezza quando si aggiungono trigger
  uscenti, così i punti di connessione restano leggibili. Trascina un blocco per
  spostarlo; la posizione viene conservata nel grafo. Ogni blocco ha **+** per
  aggiungere un link, l'icona a matita per modificarne l'aspetto e **×** per
  eliminare lo stato.
- **Create state** nel pannello animazioni crea un nuovo blocco senza link. Muovi
  il mouse sul grafo e clicca per posizionarlo dove c'è spazio libero, poi usa
  **Confirm position**. Il nuovo stato è selezionato e può essere modificato con
  l'icona a matita.
- Tasto destro su uno stato: **Parameters**, **Duplicate state** e **Delete
  state**; anche l'overlay dello stato ha **Duplicate state**. La duplicazione
  copia CSS e parametri dello stato; modificali prima di salvare. Il normale
  **Duplicate** dell'elemento clona in profondità anche `motion`, `motionKey`,
  durata/easing/delay e `motionGraph`.
- L'icona **+** su uno stato crea un link da quello stato verso una destinazione
  esistente. Il chooser si apre con **Cancel** e **Confirm**; la destinazione è
  preselezionata allo stato creato più di recente. Confirm crea il link, chiude il
  pannello e lo disegna con il colore di completamento. Per tornare allo stato
  precedente, premi **+** sul nuovo stato e scegli il vecchio. La partenza è
  sempre sul lato destro e l'arrivo al centro del lato sinistro. Un link vecchio
  senza trigger può ancora essere completato con doppio clic sullo stato sorgente
  o sulla sua riga rossa.
- Il trigger appartiene allo stato di partenza. Ogni link uscente è numerato
  nel punto di partenza; il numero è l'indice del trigger. Passando il mouse sul
  numero compare l'hint del trigger (ad esempio `click`). Più link possono
  uscire dallo stesso stato verso destinazioni diverse.
- Un trigger può ascoltare questo elemento o un altro elemento: il chooser ha
  **event source**. Per una sorgente esterna scegli un evento DOM oppure
  **external element state**, poi uno degli stati del motion graph della sorgente.
  I trigger di stato esterno scattano quando il grafo della sorgente entra nello
  stato selezionato.
- I cambi di stato partono subito. Un evento durante il tween viene valutato nello
  stato di destinazione e avvia un nuovo tween dai valori correnti (vince l'ultimo
  evento); gli eventi non vengono accodati. Una nuova transizione annulla il link
  `after` e i timer precedenti e un tween interrotto non lascia proprietà
  intermedie. I timer partono all'ingresso; i link `after` aspettano la fine della
  transizione.
- I link partono sempre dal lato destro dello stato sorgente e finiscono al
  centro del lato sinistro della destinazione, orizzontalmente. Il percorso è
  ortogonale con una distanza minima di **15 px** tra box, tra link e tra link e
  box. Il primo e l'ultimo tratto sono perpendicolari al lato del box e lunghi
  almeno 15 px. I punti di connessione sono assegnati in ordine verticale di
  destinazione, così un ventaglio di link si annida senza incroci; ogni modifica
  ricalcola tutti i link e prova più ordini prima di scegliere il percorso. I
  segmenti non attraversano mai un box, usano angoli retti e preferiscono il
  percorso più breve con il minor numero di segmenti. I link all'indietro
  vengono instradati attorno ai box. Il percorso viene ridisegnato in tempo reale
  mentre si trascina un box. Se non esiste un layout senza incroci, i link
  possono incrociarsi: il link successivo viene disegnato sopra il precedente con
  un piccolo distacco sullo sfondo. Quando un piccolo spostamento automatico di
  uno stato può ripristinare un percorso pulito, l'editor muove quello stato e
  ricalcola tutti i link; le posizioni sono fissate per mantenere il layout
  stabile. Blocchi
  e link usano colori sensibili al tema e restano leggibili in modalità chiara
  e scura.
- Click e gli altri eventi DOM singoli si possono usare una sola volta per stato
  sorgente; variabili e timer possono essere ripetuti. Il chooser disabilita gli
  eventi già usati e la validazione rifiuta i duplicati. Per un hover reversibile
  usa `pointerenter` + `pointerleave` (oppure `mouseenter` + `mouseleave`);
  `mouseover`/`mouseout` fanno bubbling, quindi scattano anche attraversando gli
  elementi figli e possono ripetersi più volte.
- Il chooser espone anche `keydown`, `keyup`, `input`, `submit`, `pointerdown`,
  `pointerup`, **wheel**, **swipe**, **long press**, **enters viewport**,
  **leaves viewport** ed eventi **custom**. I link tastiera/custom possono avere
  filtri key, mouse button e Ctrl/Alt/Shift/Meta; se più link corrispondono allo
  stesso stato/evento vince il filtro più specifico. Gli eventi custom usano il
  nome DOM inserito nel chooser.
- I link wheel possono filtrare `down` o `up`. I link swipe scelgono un asse e
  una distanza minima (default 40 px); quelli long press scelgono il tempo di
  pressione (default 500 ms). I link di visibilità scelgono una frazione tra 0 e 1
  (`0.5` per entra, `0.2` per esce) e l'opzione **fire once**. Gesti e viewport
  sono ripetibili, quindi non sono nell'elenco degli eventi singoli.
- L'overlay del trigger ha un campo **guard** opzionale. Accetta nomi di stati
  SUI, `true`/`false`/`null`, numeri, stringhe, `!`, `&&`, `||`, confronti e
  parentesi (ad esempio `open && mode == "edit"`). L'editor valida l'espressione
  prima di salvare e il runtime non usa mai `eval`; una guard non valida blocca la
  transizione.
- Le transizioni usano il motore di animazione del browser con valori iniziali e
  finali calcolati: `transform`, filtri, box shadow, clip path, gradienti,
  colori, lunghezze e custom properties vengono animati dove supportato; i valori
  non animabili cambiano in modo discreto. I valori finali sono applicati alla
  fine.
- Un trigger **variabile** elenca gli stati SUI disponibili e chiede il valore
  che lo attiva (`true`/`false`, testo o numero). Un trigger **timer** chiede
  una durata e avvia la transizione automaticamente dopo quel tempo nello stato
  sorgente.
- Un trigger **after transition** parte automaticamente quando termina la
  transizione che entra nello stato sorgente. Concatenando più link `after`
  (A → B → C) si costruiscono animazioni multi-step. Ogni transizione espone
  `duration`, un **combo box easing** con le funzioni CSS standard e i preset
  `cubic-bezier` comuni, e `delay`.
- L'editor del link espone anche **iterations** (un numero o `infinite`),
  **direction** (`normal`, `reverse`, `alternate`, `alternate-reverse`),
  **playback rate**, un toggle **spring easing** con stiffness/damping/mass e un
  nome **emit event** con **emit when** (`start`, `end`, `cancel`). Lo spring
  sostituisce la curva easing; la durata ne controlla la velocità. I loop
  `alternate` si fermano sullo stato di destinazione, quindi usa un numero
  dispari di iterazioni per terminare sul lato di destinazione.
- Un link può emettere un evento custom; il runtime emette anche
  `sui:motionstate`, `sui:motionstart`, `sui:motionend` e `sui:motioncancel`. Un
  altro grafo può ascoltarli con un trigger **custom**; scegli l'elemento
  sorgente e usa il suo `motionKey` per un riferimento stabile.
- L'overlay del trigger dispone i campi di timing e spring in una griglia compatta
  a due colonne. Duration, delay e timer sono spin box in millisecondi; playback
  rate mostra `×`; stiffness/damping/mass dello spring mostrano `N/m`, `N·s/m` e
  `kg`. **OK** chiude l'overlay dopo aver applicato le modifiche live.
- Il picker della sorgente etichetta ogni elemento con nome, `motionKey` o
  `data-sui-id` e percorso nell'albero. Una riga **Source code** ripete
  l'identificatore esatto da usare nel sorgente. Imposta una layout key sulla
  sorgente per un codice stabile e leggibile; i nodi di solo testo sono esclusi.
- **Preview** riproduce il link selezionato con delay, spring, iterations,
  direction e playback rate. **Stop preview** annulla la preview in corso; una
  preview infinita viene fermata anche aprendo un altro stato, link o pannello.
- I loop infiniti continuano finché un altro evento non li annulla e non
  attivano mai link `after`. La transizione successiva parte dai valori correnti.
- Doppio clic su uno stato: overlay con nome, CSS e lista dei trigger attivi con
  la destinazione. Non ci sono campi per i trigger in ingresso. Se lo stato ha un
  collegamento uscente senza trigger, si apre direttamente il chooser:
  **Cancel** torna all'overlay senza modificare il link, **Confirm** applica il
  trigger, chiude l'overlay, ripristina il grafo e disegna il link con il colore
  normale di completamento. **Add trigger** usa lo stesso chooser per un
  nuovo link. Doppio clic su una freccia: trigger, durata, easing e delay; una
  freccia senza trigger apre il chooser per quel link. Gli overlay sono
  trascinabili e sopra il grafo; aprirli non minimizza il pannello animazioni.
- Il clic semplice su uno stato lo seleziona e mostra i suoi parametri nella
  preview senza lasciare il pannello.
- L'icona a matita sul blocco stato entra in editing grafico: il pannello
  animazioni viene nascosto, l'elemento è isolato e centrato nella preview e
  compare solo la barra con **Save state**, **Cancel** e **Close**. Save cattura
  lo stile della preview, lo associa allo stato e riapre il pannello; Cancel
  scarta le modifiche e riapre il pannello; Close le scarta, chiude il pannello
  animazioni e torna all'editor normale.
- Il salvataggio rifiuta un trigger vuoto, un trigger variabile senza variabile,
  due stati con gli stessi parametri o lo stesso evento singolo usato due volte
  dallo stesso stato; l'errore compare nella notifica e lo stato resta aperto
  per la correzione.
- Aprendo il pannello viene selezionato lo stato corrente o iniziale.
  Cliccando nella preview il pannello si minimizza; **–**, trascinamento,
  ridimensionamento e chiusura restano disponibili. Il pulsante flottante
  **Animation panel** riapre un pannello chiuso.
- **Imposta iniziale** segna lo stato applicato al caricamento. **Anteprima**
  riproduce lo stato o il link selezionato nella preview.
- **Remove states** elimina il grafo degli stati ma conserva Layout & shared.

L'editor scrive un attributo JSON letterale `motionGraph` più gli attributi
motion classici di Layout & shared. Il runtime conserva lo stato corrente
sull'elemento e usa il motore di animazione del browser con valori iniziali e
finali calcolati; rispetta il movimento ridotto. Il pannello è per singolo
elemento; Layout & shared usa ancora `motionKey` per pagine condivise e liste.

**Controls > Split panes** inserisce due figli `pane` e un divisore
trascinabile. Imposta `direction`, `initial` e `min` nella sezione **Splitter**.
Uno splitter deve mantenere esattamente due pannelli.

**Controls > File dropzone** inserisce un'area di drop che apre anche il picker.
La sezione **Dropzone** imposta lo stato di destinazione (`dropFiles`), un
eventuale stato separato `pickFiles`, `accept`, `multiple`, `directory`,
`pickMode` e il suggerimento.

## Controlli contenuto e feedback

Il gruppo **Content** della palette fornisce timeline, stato vuoto, skeleton,
carousel e stepper; la command palette è sotto **Overlay**.

- **Timeline** inserisce tre eventi `timelineitem`. Ogni voce espone **date** e
  **status** (pending, active, done, error).
- **Empty state** inserisce icona, titolo, suggerimento e corpo. La sezione
  **Empty state** imposta icona, hint e modalità compatta.
- **Skeleton loader** inserisce tre righe animate. Imposta `lines`, `width`,
  `height`, `gap`, `circle` e `animated`.
- **Carousel** inserisce slide con pulsanti precedente/successivo e dots. Lega
  `value` a uno stato intero; la sezione **Carousel** imposta altezza, autoplay,
  intervallo, loop, controlli e indicatori.
- **Stepper** inserisce tre passi. Lega `value` a uno stato intero; ogni passo
  espone una descrizione e un'icona opzionale.
- **Command palette** inserisce un overlay legato a uno stato `open` con hotkey
  `k`. Ogni `command` espone una scorciatoia e il suo handler `onClick`. Senza
  `open` lo stesso controllo resta inline.

## Pagine

Ogni progetto ha una o più pagine. I **chip pagina sopra il preview** sono il
selettore: fare clic su un chip per modificare quella pagina. Nel designer è
modificabile/visibile solo la pagina attiva; è una vista dell'editor, non un
limite dell'app.

- **Insert → New page…** crea una pagina vuota o da template.
- La sezione **View** dell'inspector modifica `title` e `name` della pagina
  attiva.
- **Duplicate**, **Delete** e **Save as template** sono in quella sezione.
- Ogni pagina viene emessa come `view` SUI; la prima pagina è quella iniziale
  dell'app generata.

L'app generata nasconde le altre pagine finché un handler chiama il builtin
`navigate("pageName")`:

```sui
fn openSettings() -> null {
	navigate("settings")
}

button "Settings" onClick: openSettings
```

`navigate` è un no-op nelle app a pagina singola e con nomi sconosciuti. Usare
pulsanti, voci di menu o qualsiasi handler normale per esporre la navigazione.
Vedere [API reference: pagine](API_REFERENCE_IT.md#pagine-e-navigazione) e
l'[esempio pages](../examples/pages/README.md).

## Modal control

La voce **Overlay → Modal** inserisce un dialogo nascosto con titolo, contenuto
e due pulsanti predefiniti: **Cancel** e **OK** (primary). Il dialogo usa
`modal`; ogni pulsante del footer usa `modalbutton`.

### Configurazione

1. Selezionare `modal`; nella sezione **Modal** dell'inspector modificare:
   - **name** — l'id stabile usato da `modal.open(name)` / `modal.close(name)`;
   - **Open in preview** — controlla il dialogo live senza uscire dal designer.
2. Selezionare un `modalbutton`:
   - **Text** = etichetta;
   - **variant** = default, primary o danger;
   - **close** = chiude il dialogo dopo il click (attivo di default);
   - **onClick** = azione opzionale, eseguita prima della chiusura.
3. Tasto destro sull'albero:
   - `modal` → **Add modal button**, **Edit label**, **Duplicate**, **Move up/down**, **Delete**;
   - `modalbutton` → **Add button below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete**.

Il contenuto del body è ogni figlio diretto che non è un `modalbutton`;
trascinare paragrafi, input, tabelle, ecc. dentro il `modal`.

### Comportamento

- Il dialogo resta nascosto finché non viene chiamato `modal.open("name")`.
- Si chiude con click sul backdrop, `Escape` o un `modalbutton` predefinito.
- Il focus entra nel dialogo; l'opener viene ripristinato quando è disponibile.
- Aprire un altro modale chiude quello corrente.
- I pulsanti con `close: false` mantengono il dialogo aperto.
- Il runtime ripristina un dialogo aperto dopo un normale re-render.
- Le etichette sono escapate; il pannello usa `role=dialog` / `aria-modal=true`.

Vedere [API reference: dialoghi modali](API_REFERENCE_IT.md#dialoghi-modali) e
l'[esempio modal](../examples/modal.sui).

## Tabs control

La voce **Tabs** in **Layout → Navigation** inserisce un set di schede
funzionante con due schede e contenuto di esempio. Un nodo `tabs` è il set; ogni
figlio diretto `tab` è una scheda:

- **Text** di `tabs` = nome accessibile (`aria-label`), non un titolo visibile;
- **Text** di `tab` = etichetta visibile del pulsante;
- figli di `tab` = contenuto del pannello.

### Modifica dall'albero degli elementi

Fare clic con il tasto destro su una riga:

| Riga | Azioni |
|---|---|
| `tabs` | **Add tab**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `tab` | **Add tab below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| altro elemento | **Duplicate**, **Move up/down**, **Delete** |

L'albero le etichetta esplicitamente come `tabs 'Settings'` e
`tab 'General'`. Ogni nuova scheda riceve un paragrafo iniziale; è possibile
trascinare o inserire qualsiasi componente al suo interno.

### Comportamento

- La prima scheda è attiva inizialmente.
- Il click su una scheda cambia pannello.
- Sinistra/destra/Home/End spostano tra le schede.
- La scheda selezionata viene ripristinata dopo un re-render della view per i
  tab set a livello pagina.
- Solo i figli diretti `tab` sono schede; il contenuto del pannello va dentro
  un `tab`.
- Classi generate: `.sui-tabs`, `.sui-tablist`, `.sui-tab`, `.sui-tabpanels`,
  `.sui-tabpanel`.
- Le etichette sono escapate e la struttura usa i ruoli `tablist`, `tab` e
  `tabpanel`.

Vedere [API reference: tabs](API_REFERENCE_IT.md#tabs) e l'[esempio
tabs](../examples/tabs.sui).

## Controllo Accordion

La voce **Accordion** in **Layout → Navigation** inserisce una pila di sezioni
espandibili con due elementi:

- **Text** di `accordion` = nome accessibile (`aria-label`);
- **Text** di `accordionitem` = intestazione visibile della sezione;
- figli di `accordionitem` = contenuto del pannello.

Fare clic con il tasto destro su una riga:

| Riga | Azioni |
|---|---|
| `accordion` | **Add accordion item**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |
| `accordionitem` | **Add item below**, **Edit label**, **Duplicate**, **Move up/down**, **Delete** |

La prima sezione è aperta. Cliccando un'intestazione si apre e si chiudono le
altre; cliccando quella aperta si chiude. La sezione aperta viene ripristinata
dopo un normale re-render della view. Classi generate: `.sui-accordion`,
`.sui-accordion-item`, `.sui-accordion-head`, `.sui-accordion-panel`.

Vedere [API reference: accordion](API_REFERENCE_IT.md#accordion) e
l'[esempio accordion](../examples/accordion.sui).
## Controllo Combobox

Inserire **Combobox** dalla sezione **Controls**. Un nuovo controllo parte con
uno stato opzioni vuoto (`options1`, `options2`, ...) e un campo ricercabile.
L'etichetta nell'albero è `combobox 'etichetta'`.

La sezione inspector **Combobox** contiene:

| Campo | Significato |
|---|---|
| options state | stato stringa con la lista separata da virgole; vuoto per una lista literal |
| Options in ... | righe di design lette dal valore iniziale dello stato; aggiungi o rimuovi opzioni qui |
| value state | nome dello stato che riceve il valore selezionato |
| placeholder | testo del campo di ricerca |
| empty text | messaggio quando nessuna opzione corrisponde |
| max option length | caratteri massimi accettati per etichetta/valore; oltre = errore |
| max options | numero massimo di elementi accettati quando si forma la lista |
| required | richiede un valore selezionato |
| disabled | disabilita il controllo |
| popup placement | lato preferito (`auto`, `bottom`, `top`, `left`, `right`) |
| animation | effetto di apertura (`fade`, `scale`, `slide`, `none`) |
| duration / easing / delay | tempi ed easing dell'effetto |

Aggiungere una riga aggiorna il **valore iniziale** dello stato collegato. Il
sorgente generato mantiene `options: nomeStato`, quindi il codice applicativo può
assegnare una nuova stringa separata da virgole a runtime e la lista si aggiorna
al render successivo. Rimuovendo tutte le righe lo stato resta vuoto.
`Etichetta=valore` separa il testo mostrato dal valore salvato. Usare `\,` nel
campo di aggiunta per una virgola dentro un'etichetta; il sorgente dello stato
viene escapato automaticamente. Se una riga supera **max option length** o la
lista supera **max options**, la lista candidata viene rifiutata con un errore e
non vengono mostrate opzioni parziali. Anche le righe aggiunte dall'editor
vengono controllate prima dell'inserimento.

La ricerca supporta `*` (qualsiasi sequenza) e `?` (un carattere); `\*` cerca un
asterisco letterale. Dati runtime non validi mostrano un errore sotto il
controllo. Quando una lista sostitutiva non contiene più il valore selezionato,
il binding di stato generato dall'editor azzera automaticamente il valore; un
valore ancora presente viene mantenuto.

Uno stato con inizializzatore non literal non può essere modificato come lista
di righe; modificarne l'inizializzatore nel pannello States o gestire la lista
runtime nel codice. Lasciando vuoto **options state** il controllo torna a una
lista literal `options: "..."`.

La sezione Events espone `onChange`; la funzione viene chiamata dopo che il
valore selezionato è stato scritto nello stato collegato.

La preview supporta filtro digitando, `ArrowDown`/`ArrowUp`, `Enter`, `Escape`
e il pulsante di clear.

Vedere [API reference: combobox](API_REFERENCE_IT.md#combobox) e
l'[esempio combobox](../examples/combobox/README.md).

## Controllo Tree

Inserire **Tree** dalla sezione **Controls**. Un nuovo albero parte con un
elemento genitore e un figlio. **Tree item** non è più offerto nel menu Insert o
nella palette: gli elementi si creano e modificano da **Tree data** o dal menu
contestuale del tree. La struttura a sinistra mostra solo la riga **tree**, non
gli item interni.

L'inspector **Tree** espone solo:

| Campo | Significato |
|---|---|
| settings | apre **Tree settings** |
| data | apre **Edit tree data** |

**Tree settings** apre il pannello componente con queste tab:

- **General** - `Title` (titolo visibile), `Runtime name` (target delle funzioni
  `tree.*`; non è il titolo visibile) e switch: indent guides, truncate labels,
  hierarchical checkboxes, multi-select, search box, drag and drop, context menu,
  virtualize offscreen rows, persist selection ed edit mode.
- **Bindings** - binding di stato opzionali: value state, selection state
  (mostrato con multi-select), load key state, edit depth e rename key/value.
  I campi di stato suggeriscono gli stati dichiarati
  del progetto tramite datalist; lasciarli vuoti se non servono.
- **Columns** - `Show columns` ed editor colonne: aggiungi, rinomina,
  ridimensiona, rimuovi e riordina. Le key sono libere; `size`, `date` e `status`
  hanno attributi dedicati, le altre usano i valori `cells` dell'elemento. Le
  colonne restano nascoste finché **Show columns** è off. L'ordinamento figli è
  nella stessa tab.
- **Colors** - colori di testo, titolo, riga, hover, selezione, toggle, bordo,
  guida, badge, metadata e checkbox.
- **Title** - dimensione, peso, font, spaziatura, allineamento, posizione
  alto/basso e padding del titolo. Si applicano al campo **Title** in General.
- **Shape** - indent, padding/altezza riga, raggio, bordo, dimensione toggle,
  dimensione/peso testo, font step, letter spacing e font family, più icone
  curate dei nodi e simboli di ramo: **Item icon (no children)**, **Parent icon
  (closed)**, **Parent icon (open)**, **Closed branch symbol** e **Open branch
  symbol**. **Font step** riduce il font per ogni livello sotto la radice;
  default `0`. I campi icona aprono un overlay categorizzato con solo le famiglie
  pertinenti: icone file/documento per le foglie, cartelle/struttura per i padri,
  frecce di espansione per i rami e strumenti di modifica per il pulsante rinomina.
  **Add from file…** accetta ancora liste `.txt`, `.json` o `.csv`.
- **Extra** - due sezioni: **Edit button** (visibilità, icona, sfondo, colore
  testo, font, dimensione, bordo, raggio, padding e colori hover) e **Child
  counts** (visibilità, sfondo, colore testo, font, dimensione, peso, padding,
  raggio, bordo e larghezza minima). Entrambi compaiono accanto al nome del nodo.
- **Tooltip** - switch tooltip personalizzati più sfondo, testo, bordo, raggio,
  padding, larghezza massima, dimensione font, ritardo e posizione.

Il selettore **Style** in alto in Tree settings contiene gli stili condivisi più
i 16 temi Tree (Default, Compact, Comfortable, Dense, Bordered, Soft rounded,
Minimal, Neon, Forest, Amber, Ocean, High contrast, Outline, Mono, Candy,
Cards). Selezionando uno stile si vede l'anteprima; **Apply** lo scrive nel tree
e aggiorna la preview dell'editor. Tree settings non ha una tab **Hover**
separata: i colori hover sono in **Colors**. I preset con nome sono nello stesso
pannello.

**Edit tree data** apre un editor dedicato con una riga per elemento: etichetta,
key opzionale, icona, badge, metadata, celle grid size/date/status, stato
open/check/edit e azioni compatte per aggiungere figlio, aggiungere fratello,
spostare su/giù ed eliminare. La toolbar aggiunge elementi radice e
importa/esporta JSON o XML. **Apply** scrive gli elementi nel tree; Cancel non
modifica il modello. Le stesse azioni sono nel menu contestuale del tree
(**Edit tree data**, **Tree settings**, **Add root item**,
**Import JSON/XML**, **Export JSON/XML**).

Selezionando un nodo nella preview viene sempre selezionato l'**intero tree**:
un tree item non è mai selezionabile da solo e l'inspector non apre un pannello
per-item. Per modificare la struttura, fare clic destro su un nodo nella preview.
Il menu offre **Add child item**, **Add item below**, **Edit item label...**,
**Duplicate**, **Move up**, **Move down**, **Indent (under previous)**,
**Outdent** e **Delete**. Il clic destro sullo spazio vuoto del tree (o sul tree
nella lista elementi) offre **Add root item**, **Edit tree data**,
**Tree settings** e le azioni di import/export JSON/XML.

I campi solo-per-item (key, initially open, node style preset, rename icon,
icon/open icon, badge, metadata, checked, lazy children, size/date/status e celle
tree-grid) restano nel sorgente `treeitem` generato e si modificano da **Edit
tree data** o direttamente nel SUI. A runtime la stessa struttura è disponibile
dal menu contestuale del tree quando **Context menu** è attivo: aggiungi
figlio/fratello, rinomina, duplica, sposta, indent/outdent, elimina,
espandi/comprimi e copia key.

`onChange`, `onRename`, `onCheck`, `onMove`, `onLoad`, `onContextMenu` e
`onTreeChange` sono
disponibili in **Tree settings > Events** (la sezione Events generica
dell'inspector è volutamente nascosta per i tree). `onTreeChange` viene chiamata
quando i dati dell'albero cambiano: aggiunta/eliminazione, spostamento,
indent/outdent, import/clear, rinomina, checkbox, figli lazy o celle grid. Usare
`tree.lastChange(name)` per leggere operazione, key, parent e indice. Il
sorgente generato mantiene i `treeitem` annidati con key, open e rename.

Il menu **style preset** del tree scrive lo stile visivo nel campo **style expr**
dell'elemento (sezione **Appearance**). I preset sono **Default**, **Compact**,
**Comfortable**, **Dense**, **Bordered**, **Soft rounded**, **Minimal**,
**Neon**, **Forest**, **Amber**, **Ocean**, **High contrast**, **Outline**,
**Mono**, **Candy** e **Cards**. Modificando direttamente l'espressione
`cssSafe("...")` si ottiene il controllo completo e il menu passa a **Custom**.

Un tree item può comunque avere il proprio **node style preset** (`Inherit`,
`Muted`, `Accent`, `Success`, `Warning`, `Danger`, `Folder`, `Compact row` o
`Custom`) tramite `styleExpr` nel sorgente SUI. I preset di nodo scrivono
variabili `--sui-tree-*` di riga nello style expression del singolo item, quindi
sovrascrivono solo quel nodo.

Il campo **rename icon** del tree cambia il simbolo della matita; **show rename
icon** la nasconde. Un tree item può ereditare, mostrare o nascondere l'icona per
nodo tramite gli attributi generati `renameIcon` / `showRenameIcon`. Nascondere
il pulsante non disabilita F2 o doppio click.

Il campo **tree name** rende il tree un target runtime. Il pannello **Edit tree
data** importa/esporta i tree item nel sorgente di design: l'import sostituisce
i figli `treeitem` del tree selezionato, l'export scarica il modello corrente.
L'import di design non memorizza dati a runtime; per i dati
dell'applicazione usare il modulo core `tree` (`tree.importJson`,
`tree.importXml`, `tree.importJsonFile`, `tree.importXmlFile`,
`tree.exportJson`, `tree.exportXml`, `tree.downloadJson`,
`tree.downloadXml`). L'esempio combina un pulsante `pickFiles` con gli helper
file e un pulsante di download. Schema JSON/XML in
[`STDLIB_IT.md`](STDLIB_IT.md#tree-module-tree-core).

La preview supporta selezione riga, multi-selezione, toggle di espansione,
guide, icone, badge, metadata, colonne tree-grid, ricerca/filtro, checkbox
tristato gerarchico, drag and drop, menu contestuale, navigazione con frecce,
Home/End e rinomina inline (doppio click, F2 o matita; Enter/Escape). Espansione,
stato checked ed etichette rinominate vengono ripristinati dopo un re-render;
termini di ricerca e ancore di selezione restano per la sessione della pagina. Il
runtime usa `localStorage` per stato aperto/checked e override locali delle
etichette; etichette, key, icone, badge, metadata e celle grid sono escapati e
resi come testo.

Vedere [API reference: tree](API_REFERENCE_IT.md#tree) e l'[esempio
tree](../examples/tree/README.md).

## Controllo Table layout

**Layout > Table layout** inserisce una tabella HTML strutturale, distinta dal
controllo dati **Table**. Non ha nome dati, ordinamento, filtri, editing o
persistenza: dispone altri elementi in una griglia esplicita di righe e colonne.

La sezione **Table layout** dell'inspector controlla:

- **preset** (2 x 2, sidebar + contenuto, header + 2 colonne, dashboard 3 x 2);
- **rows** (1-50) e **columns** (1-12);
- **Structure table settings**, che apre un pannello con un campo per la
  larghezza di ogni colonna e per l'altezza di ogni riga; le due liste
  scorrono separatamente per gestire griglie grandi, e un campo vuoto
  significa `auto`. All'apertura il pannello allinea i contatori righe/colonne
  alla griglia effettiva; modificando una misura oltre righe/colonne dichiarate,
  il contatore corrispondente viene alzato;
- **column gap** e **row gap**;
- **empty cell height** per mantenere visibili e utilizzabili le celle vuote;
- **stack on small screens** (impilamento responsive sotto 640px);
- **semantic table** (`role="table"` invece di `role="presentation"`) e
  **aria label**.

La **width** / **height** della tabella si modificano nella sezione generica
**Size** e la cornice della tabella nella sezione generica **Border**; seleziona
una cella per modificarne il bordo nella stessa sezione **Border**. Il bordo
globale delle celle (`borderStyle`/`borderWidth`/`borderColor`) resta
source-level.

Una nuova tabella layout parte come un'unica cella vuota. Usa **Split vertical** /
**Split horizontal** nell'inspector per dividere a metà le celle selezionate;
Ctrl+click aggiunge celle alla selezione e **Merge selected** unisce le celle
solo se le prime due condividono un lato intero; quel lato stabilisce la
direzione del merge e ogni cella successiva deve continuare nella stessa
direzione. Cambiando **rows** si aggiungono/rimuovono righe complete
in basso (una cella per colonna); cambiando **columns** si aggiungono/rimuovono
celle a destra. Se la riduzione eliminerebbe celle con elementi, un dialogo
chiede se **rimuovere** il contenuto o **spostarlo** nella cella rimanente più
vicina; Cancel lascia la griglia invariata. Il **preset** compare solo quando la
tabella è vuota. Trascina le sottili maniglie verticali/orizzontali nella preview
per ridimensionare colonne o righe: le dimensioni si aggiornano e si modificano
in **Structure table settings**. Clicca una maniglia senza trascinarla per selezionare lo
split; Ctrl+click aggiunge altri split e **Delete split** unisce le celle
adiacenti mantenendone il contenuto. Se uno split renderebbe una colonna più
stretta di 40px o una riga più bassa di 24px, un dialogo chiede se continuare. I segmenti pieni separano due celle; quelli
tratteggiati sono solo un prolungamento attraverso una cella unita. Un click
altrove deseleziona cella o split; Ctrl+click è l'unico modo per aggiungere una
seconda cella.

Ogni figlio diretto occupa la successiva cella libera, riga per riga. Clicca una
cella vuota nella preview per selezionarla e vedere riga/colonna; la cella
selezionata è il contenitore in cui viene inserito il prossimo elemento dalla
palette o dal menu Insert. **Alt+click** su una cella occupata, il triangolino
in basso a sinistra sempre visibile o la riga del contenuto nell'albero
permettono di gestire il contenuto; cliccare un elemento interno seleziona
l'elemento e deseleziona
la cella. Cliccare un edit box cancella la selezione delle celle; selezionare
una cella o un altro elemento toglie il focus all'edit box. L'albero degli
elementi segue la selezione: la riga della tabella quando è selezionata la
tabella, la riga della cella quando è selezionata una cella (Ctrl+click mantiene
evidenziate le altre celle selezionate) e la riga del contenuto per un elemento
dentro una cella. Il triangolo verde
appena fuori dall'angolo in alto a sinistra seleziona l'intera tabella; un click
dentro la prima cella seleziona la cella o il suo contenuto. Seleziona una
cella per aprire **Layout cell** e impostare `column span` / `row span`,
unendo la cella con quelle a destra o in basso; **Reset to 1 x 1** la divide di
nuovo. Aumentando uno span vengono unite le celle interamente coperte e il loro
contenuto viene accodato; una sovrapposizione parziale viene rifiutata con un
messaggio. Riducendo uno span il contenuto resta nella cella e l'area liberata
torna a essere composta da celle vuote. Padding, sfondo, bordo, effetti e dimensioni si modificano nelle sezioni
generiche dell'inspector, che agiscono direttamente sulla cella. L'allineamento
riguarda gli elementi dentro la cella: selezionali e usa la loro sezione
**Alignment**. La cella resta una vera cella di tabella: l'allineamento verticale
usa `vertical-align`, quello orizzontale usa i margini automatici; **Clear**
ripristina la posizione in alto a sinistra. Con una cella selezionata
l'inspector nasconde **Position**, **Size**, **Responsive**, **Layout**,
**Grid** e **Bindings**, e **Box model** mostra solo **padding** perché i margini non si
applicano alle celle di tabella. **Tooltip**, **Events** e **Drag & drop**
funzionano sul `<td>`: scattano sulle parti vuote della cella, mentre un evento
su un componente figlio ha la precedenza. Con più celle selezionate, padding,
sfondo, bordo e le altre proprietà CSS generiche vengono applicati a tutte le
celle selezionate.


Le frecce spostano la selezione tra le celle vuote. Nella preview del designer l'impilamento responsive è temporaneamente disattivato per mantenere visibili griglia e maniglie; l'app compilata impila comunque sotto 640px.
Le celle allineate diventano flex-column quando sono impilate, quindi mantengono
l'allineamento orizzontale/verticale; le altre restano block. Le celle sono evidenziate, comprese quelle vuote e
unite, così la griglia è visibile anche senza bordi applicativi. Raggruppa più
elementi in un `col`, `div` o `card` per tenerli nella stessa cella. Usa il
controllo dati **Table** quando servono `table.display`, `table.importText`,
ordinamento, filtri o modifiche persistenti.
Vedi [API reference: table layout](API_REFERENCE_IT.md#table-layout) e
[l'esempio table layout](../examples/table_layout/README.md).

## Controllo Table

Inserire **Table** da **Controls**: crea un workbook con un foglio e una tab **+**
per impostazione predefinita. Il pannello degli elementi mostra un solo nodo
**table** per l'intero workbook; tab, wrapper di scroll e celle sono dettagli
interni. Un click su quel nodo seleziona il foglio visibile e mostra gli handle di
ridimensionamento attorno al controllo tabella. L'inspector mostra subito i
pulsanti **Structure** (`+ Col`, `− Col`, `+ Row`, `− Row`). La selezione di
celle, righe e colonne avviene direttamente nella preview: click su una cella,
gutter esterne per righe/colonne, oppure angolo in alto a sinistra per tutta la
tabella.

L'inspector mantiene **Structure** e aggiunge **+ Table** / **− Table** per più
tabelle. Il pulsante **Table settings...** apre un overlay più stretto con tutti
i gruppi in tab; dentro ogni tab i controlli sono disposti su due colonne con la
stessa dimensione compatta del resto dell'inspector:

**Appearance**, **Borders**, **Header**, **Caption**, **Behavior**, **Data**,
**Columns**, **Data view**, **Events** e **Tooltip**. Ogni campo aggiorna subito
la preview; l'overlay resta aperto durante le modifiche.

Il pulsante **− Table** e il selettore fogli compaiono solo con più tabelle.
**+ Table** avvolge una tabella singola in un workbook con sheet tab in basso
(oppure aggiunge un nuovo sheet) e seleziona la nuova tabella. Un nuovo foglio
copia dalla tabella di origine aspetto e comportamento (sortable, filterable,
eventi), mantiene una propria intestazione generica (`A`, `B`, `C`) con colonne
di testo generiche e parte con due righe corpo vuote. Il nuovo foglio è sempre
editable, anche se la tabella di origine era di sola lettura o usava colonne
numeriche. La preview mantiene il foglio selezionato mentre si modificano le
impostazioni.

La tab **+** nella preview del designer è sempre visibile e aggiunge un foglio
reale al modello del workbook, non una tabella runtime temporanea: sopravvive ai
rebuild della preview e viene scritto nel sorgente `.sui`. In un'app compilata la
**+** compare solo se il workbook ha un foglio editable. Un workbook con un solo
foglio non mostra la tab del foglio, ma solo la **+** nel designer (a runtime,
senza tab strip se il workbook è di sola lettura); le tab dei fogli compaiono dal
secondo foglio in poi. Doppio click su una tab, oppure menu contestuale e
**Rename workbook**, per cambiare il nome del workbook scritto sulla tab. Ogni
tab ha anche una piccola **×**: un click, oppure **Delete workbook** dal menu
contestuale della tab, rimuove quel workbook dal modello e dal sorgente `.sui`.
Nell'app compilata la **×** viene aggiunta dal runtime alle tab dei fogli
editable e rimuove quel foglio dal workbook runtime per la sessione della pagina;
il menu contestuale runtime della tab offre la stessa azione **Delete workbook**. Il controllo tabella occupa il contenitore e ha una
dimensione minima di **640 × 300 px**: se il contenitore non ha una dimensione
esplicita, vale quel minimo. L'area del workbook prende tutto lo spazio interno
al bordo del controllo al netto della riga titolo/ingranaggio e della barra delle
tab; segue ogni ridimensionamento del controllo. È il controllo, non il singolo
foglio, a determinare la dimensione condivisa. I fogli senza una modalità
`selection` esplicita sono editable; un click su cella, riga, colonna o tab del
foglio mantiene l'inspector sui parametri della tabella. Un foglio che richiede
la selezione riga/colonna/cella resta di sola lettura per non rompere
l'interazione di selezione. Un click sulla tab di un foglio cambia solo il foglio
visibile: non disegna né un riquadro di selezione né gli handle. Un click sul nodo
**table** nel pannello elementi evidenzia l'intero controllo (titolo,
ingranaggio, area del foglio e tab) e mostra gli handle per ridimensionarlo. Se le
colonne superano la larghezza dell'area, la scrollbar orizzontale compare in
basso, appena sopra le tab; se le righe superano l'altezza, scorre solo il corpo e
la riga di intestazione resta fissa.

Le nuove tabelle sono nominate automaticamente (`table1`, `table2`, ...), quindi
l'ingranaggio delle impostazioni runtime e gli helper `table.*` sono subito
disponibili. L'ingranaggio apre il pannello in stile Excel descritto in
**API reference → Impostazioni tabella runtime**: ordinamento da header, livelli
multi-colonna e un pulsante a imbuto in ogni header che apre una checklist
ricercabile dei valori della colonna.

### Selezione tabella

- Click su una cella per selezionarla.
- Click nella gutter appena **sopra l'header di una colonna** (fuori dalla tabella) per selezionare la colonna.
- Click nella gutter appena **a sinistra di una riga** (fuori dalla tabella) per selezionare la riga.
- Al passaggio del mouse su queste gutter compare una sottile striscia blu **a
  cavallo del bordo della cella** (metà nel padding del controllo, metà sopra il
  bordo della cella). È esattamente la zona cliccabile e non si sovrappone agli
  handle del controllo. Le strisce stanno fuori dal contenuto scrollabile del
  foglio.
- Il piccolo **triangolo blu nell'angolo in alto a sinistra della prima cella**
  seleziona tutte le celle (tutte le righe e colonne insieme), non il controllo
  esterno.
- Selezionare il controllo esterno non ruba il click sulle gutter: la striscia
  sotto il mouse ha la precedenza su una cella vicina.
- Un click sulla tab di un foglio cambia il workbook senza riquadro né handle;
  un click sul nodo **table** nel pannello elementi seleziona l'intero controllo
  e mostra gli handle di ridimensionamento.
- I nuovi fogli partono con una riga di intestazione etichettata e due righe
  vuote.
- **Shift+click** estende un intervallo di celle/righe/colonne dall'ancora precedente.
- **Ctrl/Command+click** attiva/disattiva singole celle, righe o colonne.
- Trascinare nella gutter sinistra aggiunge righe contigue; trascinare nella
  gutter superiore aggiunge colonne contigue; trascinare sulle celle seleziona
  un'area rettangolare.
- Selezionando la tabella stessa (lista elementi o selettore **Table**) viene
  evidenziato l'intero controllo. Gli handle ridimensionano il controllo; righe e
  colonne che superano il viewport producono le scrollbar interne.
- Un click su una sheet tab passa a quel foglio e aggiorna le impostazioni della
  tabella senza disegnare riquadri o handle.

Gruppi del pannello:

- **Appearance** - **tema** (Inherit global, oppure una variante Light/Dark
  delle stesse famiglie dei temi generali del progetto), padding celle,
  allineamento orizzontale/verticale, righe zebrate, hover e controlli
  scroll/scale. **Inherit global** lascia la tabella sul tema attivo; un tema
  forzato scrive i token `--sui-table-*` solo su questa tabella. **Set default**
  salva la scelta per le nuove tabelle. I colori zebra e hover sono applicati
  come tinta leggera, così restano leggibili su temi chiari e scuri. **Scroll**
  avvolge la tabella in un contenitore `tscroll` con altezza massima, header
  sticky e colori scrollbar; **scale** applica lo zoom all'intera tabella. In un
  workbook, **control padding** (gruppo Scroll & scale, default 10 px) allontana
  il viewport dal bordo del controllo e crea lo spazio per le strisce di
  selezione riga/colonna.
- **Borders** - preset, griglia celle, bordo esterno e raggio.
- **Header** - modalità riga header, colori, peso, maiuscole e header sticky.
- **Caption** - didascalia opzionale, dimensione, peso, allineamento e colore. A
  runtime la didascalia è mostrata nella riga titolo appena sopra la tabella,
  sulla stessa riga dell'ingranaggio impostazioni.
- **Behavior** - **edit mode** e modalità **cell text**. La select edit mode ha
  tre stati: **non-editable** (l'utente può leggere, ordinare, filtrare e
  selezionare ma non può modificare celle, importare dati o cambiare la
  struttura), **editable** (modifiche, import e dati creati durano per la
  sessione della pagina e i dati del sorgente tornano al reload) ed **editable
  persistent** (modifiche, dati importati e fogli creati a runtime vengono
  salvati in `localStorage` e sopravvivono a reload/riavvio; se la tabella parte
  vuota i dati creati dall'utente restano). La modalità **cell text** è
  **truncate** (default; la larghezza della colonna non cresce con il testo, una
  riga con ellissi e testo completo in overlay al passaggio del mouse) oppure
  **wrap** (larghezza colonna fissa, testo a capo e altezza riga che cresce). Le modifiche runtime a celle e intestazioni restano nello
  store della tabella tra re-render e cambio foglio per la sessione della pagina;
  la digitazione non ridisegna la tabella, il valore è finalizzato al blur o dopo
  una breve pausa e gli eventi di cambio scattano al blur. Gli eventi `input`
  delle celle restano interni alla tabella, quindi gli handler `onInput` degli
  antenati non vengono attivati. Le celle editable si possono digitare
  direttamente nella preview: click sulla cella per focalizzarla, poi scrivere.
  Nel designer il valore confermato viene anche scritto nel modello e nel
  sorgente `.sui` al blur o dopo la pausa di idle, senza ricostruire la preview,
  quindi sopravvive a un ridimensionamento o a un reload successivo. Anche le
  intestazioni sono modificabili: il click porta il cursore nell'etichetta,
  mentre l'ordinamento usa il pulsante dedicato nell'header.
- **Data** - **table name** usato da `table.display`, `table.data`, `table.json`
  e `table.exportText`.
- **Columns** - `columns`, labels, larghezze, allineamento, tipi, formati,
  opzioni select e **Use first row** per ricavare chiavi/label dall'header
  disegnato. Le colonne restano opzionali.
- **Data view** - `sortable`, `filterable`, `multi-select`, `virtual rows`
  (riservato), `show add table button` (runtime **+** tab), `runtime structure
  editing` (menu contestuale righe/colonne), `page size` e modalità `selection`
  (row/column/cell). In modalità `column` il click sull'header seleziona la
  colonna; **Alt+click** la ordina quando `sortable` è attivo. Nelle tabelle
  editable il click sull'header modifica l'etichetta e l'ordinamento usa solo il
  pulsante di sort nell'header.

I nuovi eventi sono nella tab **Events** dell'overlay:
`onCellClick`, `onRowClick`, `onCellChange`, `onRowChange`, `onTableChange`,
`onRowSelect` e `onSort`. I dettagli dell'ultima modifica sono in
`table.lastChange(name)`.

Le celle editable confermano su blur/Enter. `columnTypes` valida `number`,
`date`, `boolean` e `select`; i valori non validi vengono ripristinati e
marcati con `aria-invalid`. `columnFormats` visualizza `number`, `currency:EUR`,
`percent` e `date` mantenendo il valore raw per ordinamento, editing ed export.

### Sheet tabs in stile Excel

Per un workbook con più tabelle usare il normale controllo **Tabs**:

```sui
tabs "Workbook" position: "bottom" variant: "sheets" {
	tab "Users" {
		table name: "users" columns: "id,name" columnLabels: "ID,Name" sortable: true {
		}
	}
	tab "Orders" {
		table name: "orders" columns: "id,total" columnLabels: "ID,Total" sortable: true {
		}
	}
}
```

L'inspector **Tabs** offre **tab position** (Top / Bottom) e **style**
(Standard / Sheet tabs). Bottom + Sheet tabs creano una barra compatta sotto i
pannelli, simile a un foglio di calcolo. Ogni tabella mantiene `name`, dati,
stato sort/filter/page ed eventi indipendenti.

Vedere [API reference: tabelle](API_REFERENCE_IT.md#tabelle-e-dati-non-fidati) e
l'[esempio sheet tabs](../examples/table_sheets/README.md).

## Controllo Chart

Inserire **Chart** da **Controls**. Un nuovo grafico parte **vuoto** come barre:
i controlli dati sono contenitori, quindi nessun dataset dimostrativo viene
scritto a meno che non si usi esplicitamente **Insert sample** o si colleghi uno
stato/espressione. Riempie il contenitore (`width: "100%"`, `height: "100%"`) e
usa `overflow: "scroll"`. La tabella resta nascosta finché **Mode** non viene
impostata su `table` o `both`.

L'inspector dedicato **Chart** è ordinato per accesso rapido:

| Campo | Significato |
|---|---|
| type | bar, horizontal bar, stacked bar, grouped bar, line, step, smooth, area, combo, pie, donut, scatter, histogram, waterfall, radar, gauge, funnel, heatmap, candlestick, treemap, box plot, bullet, dumbbell o Gantt; nel sorgente l'attributo è `chartType` perché `type` è riservato. I diagrammi a blocchi usano il controllo separato **Graph** |
| Mode | solo grafico, tabella accessibile, oppure entrambi; ex **table view**, disponibile per ogni tipo |
| settings | apre il pannello unificato **Chart settings** (aspetto, punti, tooltip, tabella, interazioni, eventi, element hints e campi dati) |
| bins | numero bucket dell'istogramma (1-100) |
| min / max | intervallo gauge (default 0-100) |
| edit mode | non-editable, editable per la sessione o editable persistent tra reload/riavvio |
| show grid / show legend / show values | toggle di presentazione |
| preset | 14 preset colore classici più un preset per ogni famiglia di tema (Light/Dark) che scrivono variabili CSS in `style expr` |
| edit data | pulsante in fondo che apre il pannello overlay spostabile e ridimensionabile dei dati. Colonne e serie numeriche sono rilevate automaticamente |

**Point size** e **point width** sono in **Chart settings → Points**.
**Element hints** è anche in **Chart settings → Data**.
Il pulsante separato **Chart attributes** è stato rimosso; gli attributi
sconosciuti già presenti nel sorgente vengono comunque preservati quando il
designer riscrive il file. Non ci sono campi X/Y nell'inspector: si modificano i
dati e il grafico rileva automaticamente colonne label e numeriche.
`overflow` e `initial zoom` non sono più campi
dell'inspector: le scrollbar compaiono normalmente quando la cornice è più
piccola del grafico.

L'inspector **Chart** dipende dal tipo: i campi non utilizzabili vengono nascosti.

- **bins** solo per histogram;
- **min/max** solo per gauge;
- **show grid** nascosto per pie, donut, radar, gauge, funnel, heatmap, treemap, bullet e state;
- **show legend** nascosto per gauge, funnel e state;
- **show values** nascosto per pie, donut, scatter, gauge, radar e state.
Cambiando **type** i campi visibili si aggiornano subito. Se `chartType` �
un'espressione, tutti i campi restano visibili e la compatibilit� viene
controllata a runtime.

Se i dati correnti non sono compatibili con il tipo scelto, una finestra di
conferma avvisa che i dati esistenti verranno ignorati e sostituiti da un dataset
di esempio adatto al nuovo tipo; annullando si torna al tipo precedente. Un
grafico vuoto resta vuoto quando cambia tipo: nessun campione viene inserito
implicitamente. I dati compatibili vengono mantenuti.

Il Gantt usa lo stesso overlay ma legge righe pi� ricche: `task`, `start`, `end`
(o `duration`), `progress`, `milestone`, `group`, `parent`, `dependsOn`,
`dependencyType`, `lag`, `constraint`, `constraintDate`, `status`,
`color`, `cost`, `actualStart`, `actualEnd`, `remaining`, `work`, `workDone` e
`resources` (array o lista separata da virgole/punto e virgola). Date, barre di avanzamento, frecce di dipendenza,
weekend e marcatore oggi non richiedono campi inspector aggiuntivi. `group` è diverso: è una fase visiva che lascia il task come task reale, con la
sua barra e i suoi link. Usa la cella **Group** della tabella per un singolo task,
oppure il menu contestuale **Assign to group...** per assegnare una o più righe
selezionate senza cambiare il parent WBS. Per creare un vero contenitore WBS,
seleziona i task con Ctrl/Cmd + click nel grafico o le righe in tabella, poi
scegli **Create group from selection**: viene inserito un summary **New group**
sopra il primo task selezionato e diventa il parent dei task scelti. `parent` costruisce una struttura WBS: un summary espanso è una riga
di intestazione e solo il summary collassato mostra la barra di riepilogo, con
date, progresso e costo aggregati. Per espanderlo di nuovo basta cliccare la
barra collassata, il nome, il triangolino o il toggle in tabella. L'intestazione Gantt espone i controlli
**timeline zoom**; cambiano la larghezza dell'asse temporale senza scalare
etichette o testo delle righe. Quando la timeline zoomata supera la cornice,
compare una barra di scorrimento orizzontale nel viewport. Ctrl/Cmd + rotella
sulla timeline esegue lo zoom sulla posizione del puntatore; il pinch touch con
due dita esegue lo zoom al centro del gesto senza zoomare la pagina. La cornice del grafico mantiene la dimensione configurata. La colonna sinistra
con i nomi ha larghezza fissa, resta allineata a sinistra anche quando la
timeline è più stretta della cornice e rimane fissa durante scroll e zoom. Le
etichette lunghe vanno a capo fino a tre righe; anche la colonna Task della
tabella manda a capo il testo. **Progress % column** aggiunge una colonna
percentuale fissa dopo i nomi; le deadline sono trascinabili orizzontalmente. Le
dipendenze usano la corsia ortogonale libera più corta e vengono ricalcolate live
durante spostamento o ridimensionamento; anche warning e tabella si aggiornano.
La punta della freccia resta sul bordo sinistro del task e un tratto orizzontale
di almeno 4px la collega al punto di ingresso condiviso, quindi nessun link
termina con un segmento verticale contro la barra. I percorsi verticali non
invertono mai la direzione verso la riga di destinazione; lo stage Gantt cresce
con `linkGap` (in altezza per i corridoi tra righe, in larghezza per le corsie
parallele) e il viewport scorre per mantenere la distanza.
Ogni modifica Gantt ricontrolla tipo/lag delle dipendenze, vincoli, deadline ed
effort; i percorsi evitano barre e segmenti orizzontali/verticali già usati.
**Time scale** in Chart settings sceglie la granularità dei tick (auto, giorno,
settimana, mese o anno). Bande verticali alternate seguono la scala attiva
(giorni, settimane con inizio lunedì o mesi), con colori adattati al tema
dark/light; i tick settimanali iniziano di lunedì e alla scala settimanale le
bande dei giorni non lavorativi non vengono mai disegnate. Alla scala
giornaliera i giorni non lavorativi mantengono il colore weekend/festivo quando
sono visibili. Trascinando il bordo destro della colonna dei nomi la si
ridimensiona; la larghezza viene salvata per grafico e le etichette della scala
si fermano al separatore mentre la timeline scorre. Il menu del grafico può
mostrare/nascondere la colonna progressi e le bande dei giorni non lavorativi
senza cambiare schedulazione o dati della tabella. Mostrando la colonna
progressi, le percentuali accanto alle barre vengono nascoste. `dependsOn` accetta id o nomi dei task; la
tabella mostra chip con i nomi e un combo con tutti i task per aggiungere più
predecessori; un click sul chip apre l'editor del tipo e del lag del singolo
link. Il piccolo connettore sotto il bordo destro della barra crea una dipendenza
FS trascinandolo su un'altra barra o sul nome del task. I cicli di dipendenza
vengono rifiutati. I gruppi collassati
nascondono anche le righe corrispondenti in tabella; una barra di toggle sopra
la tabella permette di espanderli o collassarli senza usare il grafico. Critical path, auto scheduling, festivi, weekend e timezone si configurano
in **Chart settings → Interactions**: i festivi accettano date esatte, intervalli
e ricorrenze annuali; il weekend accetta nomi/numeri dei giorni o `none`; il
timezone accetta un fuso IANA come `Europe/Rome` (predefinito `UTC`). **Show
non-working days** è attivo per default e mostra/nasconde solo le bande
weekend/festivi; la validazione delle date resta attiva. L'overlay
**Calendar...** nel menu del grafico aggiunge festività e intervalli di chiusura;
il calendario è salvato per grafico e ricalcola subito la durata effettiva.

**Mode** è disponibile per ogni grafico e mostra una tabella nativa
accessibile, oppure la affianca al grafico. Il Gantt usa Task, WBS parent, Start,
End, Deadline, Actual start/end, Duration, Effective duration, Remaining,
Progress, Work, Work done, Milestone, Status, Group, Resources, Depends on, Link,
Lag, Constraint, Constraint date, Slack e Cost. Le risorse multiple sono chip
rimovibili; il piccolo pulsante **+** apre una tendina dei tipi di risorsa con il
campo quantità e la conferma aggiunge l'assegnazione. Le schede della tabella passano
tra **Tasks** e **Resources**: la barra delle schede resta visibile durante lo
scroll. La scheda Resources modifica Name, Role, Max people, Daily cost e Notes,
aggiunge righe ed elimina le risorse non usate. Rinominare una
risorsa aggiorna le assegnazioni; l'eliminazione di una risorsa ancora assegnata
viene rifiutata. Ogni chip risorsa ha una quantità (`×N`) per le persone richieste
dall'assegnazione. Max people limita la somma delle quantità dei task
contemporanei e le modifiche che lo superano vengono rifiutate; Daily cost
alimenta il calcolo automatico del Cost dei task. La
voce **Resource workload...** del menu del grafico riepiloga task, ruolo, carico,
lavoro, completamento, picco di persone, max people e costo giornaliero per
risorsa. Il registro viene salvato in `localStorage` e può essere replicato
nell'app con `onResourceChange`; **Reset saved** cancella la copia locale e la
ricostruisce dal sorgente del chart.
Actual/remaining/work guidano
esecuzione e progresso basato sull'effort; il grafico mostra le barre actual.
Duration è la differenza di calendario end-start; Effective duration conta
i giorni lavorativi usando weekend, festivi e timezone. Il
combo Constraint e il suo filtro mostrano sempre tutti i sette tipi con etichette
descrittive, anche quando tutte le righe usano `ASAP`; gli altri tipi usano Label più le serie o colonne
specifiche.

Con **Mode** impostata su **both**, ogni grafico aggancia una tabella di
altezza fissa in basso: il grafico si riduce, la tabella occupa tutta la
larghezza e resta allineata al bordo inferiore del controllo. Trascinando il
separatore sopra la tabella si imposta l'altezza; il valore è salvato per chart.
Il separatore è raggiungibile da tastiera: Su/Giù di 8 px (Shift+Su/Giù di
32 px), PageUp/PageDown del 10% dell'altezza del controllo, Home riporta a 80 px
ed End usa il massimo. Aggiungendo o scorrendo righe si muove solo il corpo della tabella, mentre
l'intestazione e i pulsanti sopra di essa restano fissi. Ridimensionando il
contenitore l'altezza della tabella resta costante e lo spazio in più va al
grafico. Il controllo Graph segue le stesse regole.

I **diagrammi a blocchi** usano il controllo dedicato **Graph**
(`graph "..." data: ...`), inserito da **Controls**: un modello generico
forme/link con `initial`, `states` e `transitions`, array di stati con transizioni
annidate, edge list CSV o XML. Le forme hanno `shape` (rectangle, square, circle,
stadium, ellipse, rhombus), fill, bordo singolo con stile/spessore/colore,
raggio angoli, dimensioni e tag/nota. Nel diagramma sono visibili solo i `tag`
espliciti: nomi, etichette e descrizioni restano nella tabella e nell'inspector.
I link hanno tipo linea (straight, arc, bezier, orthogonal,
circular), spessore, tratteggio, colore, marker iniziale/finale e tag/nota. Il
layout è automatico per livelli, con `x`/`y` espliciti per posizioni manuali.
Usa il controllo dedicato **Graph**: `chartType: "state"` (o `statemachine`)
viene rifiutato con una diagnostica; il menu dei tipi non offre più il
diagramma a blocchi.

Il pulsante a ingranaggio nell'intestazione apre un pannello **Settings** a tab:
**General** (titolo, colori bordo forme/evidenziazione, tema, switch ancore e
maniglie, una riga di pulsanti GRID, SHAPES, LINKS e TAGS, salvataggio come default e
reset), **Versions** (modifiche pendenti, commit e cronologia versioni)
e **Data** (import e export JSON/CSV/SVG/PNG). Alla chiusura di un sotto-pannello
GRID/SHAPES/LINKS/TAGS ricompare il pannello General. I default SHAPES usano due
colonne allineate con righe di pari altezza. Il grafo mantiene solo diagramma,
tabella forme/link e controlli di aggiunta. **Export JSON** include forme, link e
impostazioni complete: reimportandolo il diagramma viene ricostruito identico.
La select **Edit mode** nella tab Interactions è l'impostazione a 3 stati:
**non-editable** (sola visualizzazione, con tooltip, selezione e coda eventi),
**editable** (modifiche per la sessione della pagina, dati del sorgente
ripristinati al reload) ed **editable persistent** (dati, geometria, stile e
attributi runtime salvati in `localStorage` e mantenuti dopo reload/riavvio,
compresi dati importati e controlli inizialmente vuoti riempiti dall'utente).
In modalità non-editable ogni interazione viene accodata dal chart; SUI consuma la
lista con `chart.pollEvents` / `chart.nextEvent` e decide quale funzione esterna
chiamare. L'elemento della lista contiene `event`, `kind`, `part`, `id`,
`label`, `tag`, coordinate e modificatori. I callback diretti non sono
supportati: la lista eventi è l'unico modello di interazione. Gli id delle forme
sono obbligatori e univoci; il designer segnala id mancanti o duplicati come
errore di import.
**Hover hints** nella tab Interactions corrisponde a `stateHints`; disattivarlo
nasconde tutti i tooltip. Il pulsante **HINTS...** apre il pannello globale dello
stile tooltip con due gruppi: switch/colori a sinistra e slider di
dimensione/spaziatura a destra. Include anche **Text align** (left, center,
right) per i tooltip su più righe. Il campione dello sfondo include il proprio
slider di opacità, quindi non serve un pulsante trasparente separato; scrive gli
attributi chart `stateHint*`. Ogni forma/link
accetta un campo `hint`, modificabile dalla colonna **Hint** della tabella o dal
menu contestuale **Hint settings...**; usare `|` per andare a capo e `\|` per una
pipe letterale. **Element hints** in **Chart settings → Data** elenca tutti gli stati e link nei
dati JSON, CSV o XML (incluso un documento letterale contenuto in uno state
semplice come `data`) e modifica tag e hint con campi a riga singola (`|` va a
capo, `\|` è una pipe letterale). Riscrive i dati; il CSV richiede una riga di
intestazione con `id`/`label` oppure `from`/`to` più le colonne `tag`/`hint`, che
vengono aggiunte se mancanti. Le espressioni calcolate usano ancora l'editor dati.
Gli attributi chart non esposti dall'inspector (griglia, routing, stili tag/link,
...) vengono preservati alla lettera quando il documento viene aperto e salvato.
Il pannello separato **Chart attributes** è stato rimosso; usare le opzioni
supportate del pannello chart settings.Con la tabella visibile, selezionando un elemento nel diagramma la
riga corrispondente riceve focus ed evidenziazione (tab **Links** per i link).

La griglia usa `stateGridGap` come lato cella e distanza minima tra linee, con
`stateGridColor` per il colore; è visibile di default (24 px) e si nasconde dal
pannello Settings (GRID). Centri dei
cerchi e vertici delle forme sono agganciati ai vertici della griglia; i box
rettangolo/quadrato/stadio sono allineati e dimensionati sul passo, rombo/ellisse
seguono centro ed estremi. I segmenti orizzontali/verticali, gli stub e le corsie
di fallback dei link ortogonali giacciono sulle linee della griglia. Gli anchor sono intersezioni griglia/perimetro nominate
`V<col>T/B`, `H<row>L/R` (fallback `P:<side>`), visibili solo con l'elemento
selezionato e le maniglie di modifica attive; le proprietà link mostrano nomi e coordinate
in sola lettura. Selezionando un link compaiono le maniglie inizio/fine:
trascinandole su un altro anchor il percorso viene ricalcolato con il router che
evita gli ostacoli e il tipo link scelto.
X/Y, dimensioni e anchor completi sono nei metadati nascosti della riga e nel
pannello proprietà, senza colonne aggiuntive. Il tooltip di forme e link mostra sempre nome e
descrizione, anche senza tag; i tag nel grafico usano lo stesso tooltip.
Description e Tag / note nel pannello proprietà sono campi a riga singola; nei tag
`|` va a capo e `\|` è una pipe letterale. I tratti ortogonali escono/entrano
perpendicolari per almeno `stateGridGap`, poi seguono A* o le corsie esterne; i
self-loop sono normali link tra due punti di attacco distinti della stessa forma
(mai lo stesso punto): si crea in modalità link facendo doppio clic sulla stessa
forma e trascinando una delle due maniglie si cambia quel punto di attacco in modo
indipendente. Il sottotipo scelto determina anche il self-link
(`straight`/`arc`/`circular`/`orthogonal`/`curved`) e le Bézier espongono due
maniglie per i punti di controllo. I link circolari espongono una maniglia del
raggio che ribalta anche l'arco se trascinata oltre la corda; il link selezionato
evidenzia tutto il tracciato, punte comprese. I livelli delle maniglie stanno sopra
le forme, così le maniglie di raggio/vertici restano raggiungibili anche sui
cerchi; durante il trascinamento la lavagna resta fissa e al rilascio l'elemento
spostato viene riportato in vista. Le maniglie possono essere rilasciate anche su un'altra
forma per riagganciare quell'estremità. Le proprietà del link hanno **Direction**
(inversione sorgente/destinazione) e il pannello **LINK** può disegnare la freccia a
entrambe le estremità. Se il router non
trova un percorso pulito il diagramma resta invariato e un avviso non bloccante
spiega il problema. Spostando/ridimensionando una
forma gli anchor sono ricalcolati e i link non più agganciati vengono riancorati.
Lo spostamento e il ridimensionamento agganciano sempre la forma alla griglia e
salvano la geometria arrotondata. Zoom con Ctrl/Cmd + rotella o pinch a due dita (50%-400%). La tabella usa due
schede, **Shapes** e **Links**: Shapes mostra `Element`, `Tag`, `Type` e azioni;
Links mostra `Element`, `Tag`, `Type`, `From`, `To` e azioni. Element è la prima
e contiene il color picker. Selezionando un elemento nel grafico si apre la
scheda corrispondente e la riga viene portata nell'area visibile. Cliccare
l'elemento o il suo tag produce la stessa selezione. Con una forma selezionata,
le frecce la spostano di una cella, Shift+frecce la ridimensionano (Left/Right
larghezza, Up/Down altezza) mantenendo l'angolo superiore sinistro e
Delete/Backspace la elimina. La barra dei tab della tabella contiene un solo
pulsante **Add row** allineato a destra, che apre la scelta Shape/Link. La
tabella non ha pulsanti propri di impostazioni o undo/redo: import/export sono in
**Graph settings** e i pulsanti undo/redo del grafico coprono sia il diagramma
sia la tabella. Il pulsante compatto **+**
nell'intestazione apre un unico menu con i sei tipi e **New link**, senza coprire
i nodi; il tooltip indica `+`/`N` e `L` crea direttamente un link.
La selezione è sincronizzata grafico/tabella con colore ambra.
Passando il mouse su una cella della tabella appare un tooltip con tutto il
testo; la combo della colonna Type è esclusa perché il valore è già leggibile.
I `|` vengono mostrati su righe separate, come nel diagramma.
**Ctrl/Cmd/Shift+click** aggiunge o toglie elementi dalla selezione; il
**trascinamento con il tasto sinistro** sullo spazio vuoto disegna un rettangolo
di selezione, mentre il **trascinamento con il tasto destro** sposta l'intero
grafico. **Alt+trascinamento** e il tasto centrale spostano anch'essi la lavagna.
Un click sinistro sullo spazio vuoto azzera la selezione. Un gruppo selezionato si trascina come un'unica
unità mantenendo invariate le interdistanze; la traslazione finale è arrotondata
a passi interi di griglia su entrambi gli assi. Trascinando oltre il viewport la
lavagna si espande invece di bloccare le forme. **Ctrl+C**/**Ctrl+V** copiano e incollano,
**Ctrl+D** o **Duplicate** duplicano; duplicare le forme non copia i link
collegati se non sono selezionati. Con almeno due forme selezionate compare la
toolbar di allineamento (sinistra/centro/destra, alto/centro/basso,
distribuzione orizzontale/verticale) con guide magnetiche blu durante il
trascinamento. Il menu contestuale della forma contiene **Shape properties...**,
**Tag settings...**, **Duplicate** e **Delete**: lo stile del testo è in
**Tag settings...**, la rotazione si modifica nella tab **Position** e
l'eliminazione rimuove anche i link collegati. Undo/Redo sono icone con hint. Il pulsante
`…` apre il pannello proprietà per forme/link (stili, dimensioni, testo, tag,
marker e porte). Le proprietà del link sono a schede (**Line**, **Arrows**,
**Anchors**): colori linea/evidenziazione impilati, spessore con slider + campo
numerico, tipo di tratteggio e marker iniziale/finale con anteprima grafica,
anchor in sola lettura. Il campo label resta nella tabella; i campi description,
event, guard, action e tag/nota non fanno parte del pannello. Anche le proprietà
della forma sono a schede (**Shape**, **Position**): in **Shape** le impostazioni
bordo (bordi, stile, spessore, corner radius, lunghezza/gap per dashed/dotted)
stanno a sinistra, subtype e colori fill/bordo/evidenziazione a destra, con
un'anteprima della forma in basso; **Position** contiene rotazione, X/Y, larghezza
e altezza su due colonne allineate in alto. La label resta nella tabella e lo stile
del testo è nei Tag settings. Ogni cella Element ha un color picker. I due tipi di tag
(forme/link) hanno box, linea, font e dimensione separati e si configurano dal
sotto-pannello **TAGS** nella tab **General** delle impostazioni; il pannello
imposta anche la dimensione delle frecce dei tag e dei link tra forme (default
`stateGridGap / 4`). I tag evitano forme, altri tag e segmenti. La tab **General**
espone anche i sotto-pannelli GRID, SHAPES e LINKS, il tema chiaro/scuro,
**Save settings as default** per i nuovi diagrammi e **Reset settings** per
ripristinare i valori `.sui`. Le impostazioni del diagramma includono anche
l'interruttore **Show anchors and handles**: disattivato rende il diagramma di
sola lettura (`stateHandles: false`), nascondendo ancore, grip di
ridimensionamento/rotazione, maniglie di tag e percorso e disabilitando il
trascinamento delle forme. La griglia è opzionale dal pannello **GRID**
(`stateGrid: false`). I percorsi ortogonali vengono verificati su griglia,
altre forme e link già instradati; i candidati non validi riprovano con corsie
esterne e il riepilogo è in `data-sui-chart-route-audit`. Il tasto destro su un
link apre Link properties, Tag settings e i comandi per i waypoint: **Add
waypoint here** inserisce un vertice trascinabile agganciato alla griglia,
**Clear manual waypoints** li rimuove. Le modifiche sono autosalvate e supportano Undo/Redo;
**Commit changes…** nella tab **Versions** delle impostazioni crea versioni con
audit per campo, **Version history…** le ispeziona e `onCommit` gestisce la
sincronizzazione server. Vedi
[`examples/state_machine/`](../examples/state_machine/README.md).
Ogni intestazione ha un pulsante filtro in stile Excel con popup:
ricerca testuale, elenco dei valori distinti e campi min/max per colonne numeriche
o date ISO. Un click ordina; Shift+click aggiunge un ordinamento multi-colonna.
Trascinando le intestazioni si riordinano le colonne; trascinando il bordo
destro di un'intestazione si ridimensiona la colonna; un doppio click la adatta
al contenuto più largo. La maniglia è accessibile da tastiera con
Left/Right/Home. Le larghezze sono salvate con la tabella e
**Columns → Reset column widths** ripristina il dimensionamento automatico. La
prima colonna identificativa resta fissa durante lo scroll orizzontale. L'ingranaggio
**Settings** apre i sottomenu Columns, Filters ed Export. Export scrive le righe
visibili in CSV, TSV, JSON o JSONL, copia le righe visibili come TSV oppure apre
una vista di stampa. L'opzione **Add row button** in **Chart settings → Table**
abilita **+ Row** per ogni tipo di grafico; le righe Gantt sono modificabili,
gli altri grafici ricevono una riga locale a valori zero e possono notificare
`onDataChange`. Lo stesso pannello controlla **Delete row button**, **Number
steppers**, **Context menu** e **Selection** (`none`, `row`, `column`, `cell`).
Gli stepper numerici abilitati stanno sul bordo destro delle celle e compaiono
solo dopo un click che porta il focus sulla cella, non al solo passaggio del
mouse, e sono nascosti durante un trascinamento di selezione. Iniziare una
selezione annulla la modifica in sospeso e rimuove il focus. Le selezioni di
righe, colonne e celle usano lo stesso colore e un perimetro da 2px
`--sui-chart-table-selection-border`. In modalità cella viene riempita solo la
cella selezionata; la riga resta collegata al grafico senza tinta completa. Il focus
da hover usa sfondi neutri attenuati (`--sui-chart-table-focus-row-bg` /
`--sui-chart-table-focus-cell-bg`); clic e trascinamento rimuovono l'evidenza
hover e lasciano solo la selezione. Il focus da tastiera è un bordo sottile
`--sui-chart-table-focus-color`, lasciando visibili testo e griglia; righe e
celle selezionate mantengono il colore di selezione anche sotto il mouse. Le
modifiche inline e gli stepper aggiornano grafico e tabella in place; i refresh
generati della view preservano il DOM di grafico e tabella, evitando flicker e
ricostruzioni.
Usa la gutter dei numeri di riga per selezionare/trascinare righe, la striscia
sopra le intestazioni per le colonne, oppure trascina in diagonale nel corpo per
disegnare un rettangolo di celle. Ctrl/Cmd+click attiva/disattiva, Shift+click
estende e Alt+click ordina un'intestazione. Il menu
**Rows** e il menu contestuale copiano, incollano o eliminano la selezione;
l'accesso agli appunti viene richiesto dal browser e il risultato appare nella
status line della tabella. **Export selection** esporta solo righe, colonne o celle
selezionate in CSV, TSV, JSON, JSONL, XLSX o HTML. **Export options** controlla
intestazioni, colonne nascoste e ordine originale/corrente. **Print / Save PDF
(chart + table)...** stampa l'SVG del grafico con la tabella filtrata. Le tabelle
oltre 300 righe usano lo scrolling virtuale a finestra, mentre filtri ed export
coprono l'intero insieme filtrato. Filtri, ordinamento, ordine, larghezze e visibilità
restano in `localStorage`. Le intestazioni ordinano al click o con Enter; righe e header
sono raggiungibili da tastiera. Filtro, ordinamento e colonne restano durante i
rerender. La didascalia appare solo in modalità tabella.

Click destro sul grafico, o pulsante **⚙**, per export SVG, PNG, CSV, TSV, JSON o
JSONL, raggruppati nel sottomenu **Export** del menu del grafico. L'export dal
grafico include tutte le righe; Settings esporta le righe e colonne visibili.

**Chart settings** apre il pannello unificato con tab Appearance, Points, Tooltip,
Table, Interactions, Events e Data. Controlla tema dark/light, **Editable colors**
(color picker on/off), preset, point size e point width, tooltip, tabella, **Column filters**
on/off, **Edit mode** (**non-editable**, **editable** o **editable persistent**), **Critical path**, **Auto schedule**, **Holidays**, **Weekend**,
**Timezone**, **Link color**, **Link width** (default `1px`), **Min link gap**
(default `4px`), **Line style**,
**Arrow**, **Arrow size**, **Bar border** e **Bar border color**; la tab Data
contiene **Element hints**. Apply scrive attributi e variabili
`--sui-chart-tooltip-*` / `--sui-chart-table-*` insieme. I temi dark e light
usano token espliciti; i colori di selezione sono condivisi e le tonalità focus
si adattano allo sfondo del tema. Un click sul campione colore di una serie
nell'intestazione o nella legenda cambia il colore della serie e salva
l'override per chart. Gli aggiornamenti del grafico sfumano il vecchio SVG nel
nuovo, salvo richiesta di movimento ridotto.

I grafici ricevono il focus da tastiera: frecce per spostarsi tra gli elementi,
Home/End per saltare, Enter/Spazio per selezionare, Escape per chiudere il tooltip.

Con **edit mode** editable o persistent, le celle Task, WBS parent, Start, End, Progress, Group e Depends on sono modificabili;
anche le celle Label/X/serie non Gantt sono modificabili e chiamano `onDataChange`
(i bin dell'istogramma restano read-only).
Con **tableAddRow** attivo, **+ Row** aggiunge un task, lo seleziona, cancella i
filtri attivi, ridisegna la timeline e chiama `onTaskChange`.
Le modifiche alle tabelle Gantt e non Gantt si annullano/ripristinano con
`Ctrl+Z`, `Ctrl+Shift+Z`/`Ctrl+Y`, dai controlli della tabella o dal menu
contestuale della tabella; con un input focalizzato resta attivo l'undo nativo del browser.
Una modifica ridisegna il Gantt e chiama `onTaskChange`; drag/resize sulla
timeline aggiornano la tabella con gli stessi valori. Selezionando un'attività
nella timeline, il nome del task a sinistra o la riga si evidenziano tutti e tre e la riga viene portata in vista. La selezione vale per ogni coppia
grafico/tabella: un elemento seleziona la riga e una riga evidenzia tutti gli
elementi della label, comprese tutte le serie nei grafici multi-serie. La
selezione è salvata in `localStorage` e ripristinata dopo il reload. Una modifica
valida fa lampeggiare per qualche istante le celle e l'attività grafica.

Tooltip, evidenziazione hover, navigazione da tastiera e selezione sono
disponibili su ogni grafico. Le righe Gantt sono target da tastiera: Left/Right
spostano, Shift+Left/Right ridimensionano la fine, Alt+Left/Right l'inizio,
Enter apre il menu attività e il connettore dipendenze apre con Enter il
selettore destinazione. Una dipendenza focalizzata supporta Enter/Shift+F10 per
la modifica e Delete per la rimozione. **edit mode** editable/persistent abilita inoltre
drag/resize Gantt, gruppi collassabili e celle Gantt modificabili. Le modifiche
sono autosalvate in una bozza locale e non diventano una versione finché non si
preme **Commit** nell'intestazione. **Commit changes...** e **Version history...**
nel menu del grafico creano e ispezionano versioni con audit per campo. Se il
grafico ha `onTaskChange` in **Chart settings > Events**, l'handler riceve il
JSON aggiornato;
`onCommit` riceve la versione committata per la sincronizzazione host/server con
`revision` e `baseRevision`.


**Overlay dati.** Il pulsante **data** sostituisce la vecchia textarea inline.
Apre un pannello flottante sopra il designer con un grande editor monospace.
Trascinare l'intestazione per spostarlo e la maniglia in basso a destra per
ridimensionarlo; l'editor ricorda posizione e dimensioni. Il pannello offre:

- modalita **Literal data** e **State expression**;
- selettore formato auto, JSON, XML o CSV;
- validazione live e conteggio righe/caratteri;
- **Load file** per file `.json`, `.xml`, `.csv` o `.txt`;
- **Insert sample** per il formato selezionato;
- **Pretty JSON** per JSON valido;
- **Apply**, **Cancel**, chiusura con click sullo sfondo o Escape e Ctrl+Enter
  per applicare.

JSON/XML non validi vengono segnalati prima di Apply e bloccano l'aggiornamento.
L'overlay controlla anche la **compatibilita dati/tipo**: i grafici a serie
singola richiedono una serie numerica, stacked/grouped/combo almeno due, radar
piu serie, heatmap y+valore, candlestick OHLC, box plot cinque statistiche,
bullet valore+target, dumbbell/Gantt due serie, pie/donut/funnel/treemap un
valore positivo e histogram valori numerici. Il selettore tipo
nell'overlay segnala la differenza e disabilita Apply finche non viene risolta.
Applicando dati letterali il pannello scrive automaticamente la stringa SUI con
gli escape. La modalita espressione salva lo stato/espressione e rende il
grafico reattivo. Il pannello resta uno strumento del designer: l'app generata
mantiene il normale runtime chart e i suoi limiti.

Larghezza e altezza del chart si impostano con i campi standard della sezione
**Size** (o con CSS in **style expr**), non con parametri chart dedicati. Se la
dimensione è omessa il chart riempie il contenitore e cresce con esso; la
tabella segue la larghezza del contenitore mantenendo l'altezza fissa. In un
contenitore senza altezza definita il chart mantiene almeno 240 px
(`--sui-chart-min-height` sovrascrive il fallback).

Con **Scrollbars** il grafico mantiene la cornice e mostra una scrollbar
orizzontale quando i punti richiedono piu spazio. **Zoom controls** aggiunge `-`, `+`, `Reset` e la
percentuale; ingrandendo compaiono le scrollbar. **Fit to frame** comprime il
disegno nella cornice senza controlli.

Il menu **preset** scrive variabili `--sui-chart-*` nel campo **style expr**
dell'elemento (sezione **Appearance**). I preset **Classic** sono **Default**,
**Ocean**, **Forest**, **Sunset**, **Neon**, **Mono**, **Pastel**,
**High contrast**, **Amber**, **Candy**, **Slate**, **Light**, **Outline** e
**Minimal**. I **Theme presets** aggiungono una palette chart coordinata per ogni
famiglia di stile condivisa, in variante **Light** e **Dark** (Standard, Sketch,
Wireframe, Terminal, Glassmorphism, Brutalist, Blueprint, Neumorphic, High
contrast, Editorial, Cyberpunk e Geometric). Scegliendo **Custom**
l'espressione resta invariata e si puo modificare direttamente. Il nome del
preset selezionato resta visibile dopo **Apply**, perche il pannello riconosce
un preset anche quando tooltip e tabella aggiungono override alla sua style
expression.

Il combo **Theme** del chart e indipendente dal tema di progetto. Con il preset
**Default**, un chart scuro su una pagina chiara mantiene una superficie scura
solida e testo chiaro, e i pulsanti di intestazione, zoom e impostazioni usano
colori chart-specifici per testo, sfondo e bordo invece della palette chiara
della pagina. Scegliendo il tema chiaro con **Default** si passa alla palette
chiara predefinita. Le vecchie style expression Default vengono riconosciute
come **Default** e aggiornate con Apply, quindi un chart scuro non trasforma piu
i pulsanti in un blocco nero illeggibile.

Un grafico in modalita espressione e reattivo: cambiando lo stato collegato
viene ridisegnato. Dati non validi o vuoti mostrano un messaggio testuale dentro
il grafico. Le graffe JSON letterali vanno scritte `\{` e `\}`; l'overlay dati
lo fa automaticamente quando applica un valore letterale.

Vedere [API reference: chart](API_REFERENCE_IT.md#chart) e l'[esempio
chart](../examples/chart/README.md).

## Controllo Graph

Inserire **Graph** da **Controls**. Un nuovo graph parte come diagramma a blocchi
**vuoto**: stati e transizioni sono dati e non vengono inseriti come contenuto
dimostrativo; usa **Edit data** o collega uno stato/espressione. Riempie il
contenitore (`width: "100%"`, `height: "100%"`). È un controllo foglia: non può contenere
elementi figli. La select **Edit mode** in **Graph settings → Interactions** rende
il diagramma non-editable, editable per la sessione o editable persistent
(`localStorage` tra reload/riavvio). Controlla anche la modifica della tabella e
le maniglie: non c'è più uno switch separato.

L'inspector del graph è ordinato così:

| Campo | Significato |
|---|---|
| Graph settings | primo pulsante; apre lo stesso pannello runtime dell'ingranaggio (Generale più GRID, SHAPES, LINKS, TAGS e HINTS) |
| Styles | preset colore e **Edit styles...**, che apre il Graph style editor con i tab Title, Frame, Buttons, Table e Scrollbars |
| show table | mostra o nasconde la tabella sincronizzata |
| settings button | mostra o nasconde l'ingranaggio che apre Graph settings |
| zoom buttons | mostra o nasconde i pulsanti zoom, reset e fit |
| add element button | mostra o nasconde il pulsante + Add per forme e link |
| undo/redo buttons | mostra o nasconde i pulsanti di cronologia |
| edit data | ultimo controllo; apre l'editor dati JSON/XML/CSV o espressione |

Il selettore del tipo non compare: `graph` disegna sempre la geometria a blocchi
e non scrive `chartType` nel sorgente. Per lo stesso motivo il tipo è disabilitato
nell'editor dati. L'inspector non mostra più zoom iniziale, element hints o
attributi chart per Graph.

Il pannello runtime **Graph settings** aperto dall'ingranaggio (o dal pulsante
Graph settings dell'inspector) non ha più un tab **Style**. Il tab **General**
contiene un'area **Theme** con un combo dei preset (Midnight, Graphite, Ocean,
Forest, Sunset, Candy, Neon, Slate, Paper, Sand, Ice, High contrast, Plum, Mono e
Custom) e un pulsante **Edit theme...**. Il pulsante apre il pannello editor del
tema, diviso nei tab **Control**, **Title**, **Buttons**, **Table** e
**Scrollbars**. Mostra i valori attuali di ogni componente non diagramma:
sfondo, bordo, raggio e padding del controllo; background e altezza della barra
titolo; tipografia, abbellimenti e ombra del titolo; colori dei pulsanti per
ogni stato, spessore bordo/raggio/font; colori, bordi, tipografia, header e toni
degli input della tabella, inclusi hover e riga selezionata; dimensione, colori e
raggio delle scrollbar. Font, font style, decorazione e ombra usano combo box. I
preset definiscono anche default leggibili per axis, grid, legenda ed etichette valori,
tooltip/hint e accenti error/critical. Forme, link, tag, hint e griglia hanno i
loro pannelli dedicati, che sovrascrivono i default del tema. L'editor parte dai
valori attualmente applicati: dopo aver scelto un preset i campi mostrano quel
preset, non i default. I campi sono disposti su due colonne, ciascuna con
l'etichetta sulla stessa riga del controllo; i campi colore mostrano il valore
hex, mentre il pannello di selezione contiene lo slider alfa e il campo alfa
testuale. I campi colore sono raggruppati nella stessa colonna e gli slider/numeri
nell'altra; le combo mostrano una freccia a tendina per non essere confuse con i
campi di testo. Le modifiche sono
immediate, salvate per chart, e gli
stessi stili di controllo e pulsanti sono usati dai pannelli interni al grafico.
**Reset theme** nell'editor (o **Reset settings**) li cancella.

Il **Graph style editor** del designer usa le stesse categorie Title, Frame,
Buttons, Table e Scrollbars. Le modifiche sono mostrate in anteprima dal vivo e
diventano valori del sorgente solo con **Apply**.

Il graph riempie il contenitore e parte con la tabella nascosta. Con **show
table** attivo la tabella occupa tutta la larghezza in basso nel controllo e il
diagramma si riduce, così grafico e tabella si dividono l'altezza disponibile.
Trascinando il separatore tra loro si imposta l'altezza della tabella; l'altezza
viene salvata per chart. Il separatore supporta gli stessi comandi da tastiera
(Su/Giù, Shift+Su/Giù, PageUp/PageDown, Home ed End). Aggiungendo righe scorre solo il corpo della tabella,
mentre l'intestazione e i pulsanti sopra di essa restano fissi. Ridimensionando il
contenitore l'altezza della tabella resta costante e lo spazio in più va al
grafico.

Con **edit mode** editable o persistent la tabella è modificabile: la cella
Element mostra color picker e campo testo, Tag e Hint sono modificabili, Type
diventa una select e l'ultima colonna espone impostazioni testo, proprietà
elemento e cancellazione riga. In modalità non-editable la tabella è in sola
lettura: le celle mostrano testo semplice, la colonna azioni non compare e il
pulsante Add row sparisce. Mentre si scrive in una cella il suo tooltip viene
nascosto. I nuovi Graph partono non-editable; la modalità si sceglie in Graph
settings.

## Controllo Date/Time picker

Inserire **Date/Time picker** dalla sezione **Controls**. Un nuovo controllo
parte in modalità `date`, con valore vuoto ed etichetta `Date`.

La sezione inspector **Date/Time picker** contiene:

| Campo | Significato |
|---|---|
| value state | stato stringa che riceve il valore simil-ISO |
| mode | date, time o datetime |
| min | valore minimo nativo nel formato della modalità |
| max | valore massimo nativo nel formato della modalità |
| step | secondi opzionali per time/datetime |
| disabled | disabilita input e clear |
| required | flag required nativo |

`onChange` è disponibile in **Events**. Il sorgente generato mantiene modalità e
intervallo; il valore è collegato tramite un input nascosto sincronizzato dal
runtime.

Formati:

- date: `YYYY-MM-DD`
- time: `HH:MM` oppure `HH:MM:SS`
- datetime: `YYYY-MM-DDTHH:MM` oppure con secondi

La preview usa il picker nativo del browser. Il pulsante clear azzera valore e
stato. Valori non validi o fuori intervallo mostrano un errore inline.

Vedere [API reference: date/time picker](API_REFERENCE_IT.md#datetime-picker) e
l'[esempio datepicker](../examples/datepicker/README.md).

## Controllo Toast

La voce **Toast host** nella sezione **Overlay** inserisce un contenitore vuoto
con live region. È opzionale: senza host il runtime crea automaticamente un host
fisso in alto a destra.

Usare un handler per mostrare il feedback:

```sui
fn save() -> null {
	toast.show("Salvato", "success", 3000)
}

fn clear() -> null {
	toast.clear()
}
```

- `toast.show(message [, kind [, duration]])`;
- tipi: `info` (default), `success`, `warning`, `error`;
- durata predefinita: `4000` ms; `0` mantiene il toast aperto;
- massimo quattro toast per host per default (`max` accetta 1–8); il più vecchio viene rimosso;
- l'utente può chiudere ogni toast con il pulsante X;
- `toast.clear()` rimuove tutti i toast visibili;
- i messaggi usano `textContent`, mai HTML raw;
- l'host usa `aria-live="polite"` e ogni toast usa `role="status"`.

Selezionare il **toast host** nell'albero per aprire il pannello dedicato
**Toast**:

| Sezione | Controlli |
|---|---|
| Posizionamento | orizzontale `left`/`center`/`right` e verticale `top`/`center`/`bottom` |
| Offset | offset orizzontale solo per left/right; offset verticale solo per top/bottom |
| Comportamento | durata predefinita, massimo visibili, mostra icone |
| Anteprima | tipo anteprima, messaggio, **Show preview**, **Clear preview** |

Ogni asse è indipendente. Un asse centrato non ha campo offset perché è
centrato nel viewport; un asse ancorato a un bordo mostra il proprio campo
offset (default `16px`). Un numero senza unità viene interpretato come pixel;
scrivere esplicitamente `%`, `em`, `rem`, `vw` o un'altra unità quando serve. Il
sorgente generato usa `hAlign`, `vAlign`, `hOffset` e `vOffset`.

Con più notifiche, un'ancora superiore aggiunge ogni nuovo toast sotto il
precedente; un'ancora inferiore inverte lo stack, quindi ogni nuovo toast appare
sopra il precedente. Il toast più vecchio resta il primo a essere rimosso quando
si raggiunge il massimo.

Le normali sezioni **Size**, **Border**, **Box model**, **Typography**,
**Background** ed **Effects** sotto il pannello scrivono la regola della card.
Larghezza e altezza usano quindi i campi **Size** standard; bordo, raggio,
sfondo e tipografia si applicano direttamente alla card del toast.

Le stesse dichiarazioni compaiono nel campo **style expr** dell'host Toast.
Quel campo è l'elenco CSS modificabile per `.sui-toast` (ad esempio
`width: 420px; border-width: 3px;`). Modificarlo aggiorna la card; le custom
properties come `--sui-toast-icon-bg` vengono ereditate da icona, messaggio e
pulsante di chiusura.
Le modifiche aggiornano subito sorgente e CSS generati. La preview mostra un
toast campione con il vero CSS dell'app, così posizione, dimensioni e bordo
sono visibili prima di uscire dal designer. Se il progetto contiene più host
Toast, il campione appare nell'host selezionato; a runtime `toast.show` usa il
primo host nel documento.

L'etichetta nell'albero è `toast host`; il campo Text non viene usato. Senza host
il runtime crea un host fisso in alto a destra.

Vedere [API reference: toast](API_REFERENCE_IT.md#notifiche-toast) e
l'[esempio toast](../examples/toast/README.md).
## Proprietà Tooltip

Il Tooltip è una **proprietà di elementi/controlli**, non un controllo separato
nella palette. Selezionare un elemento qualsiasi (ad esempio un pulsante, un
link, un input o una card) e aprire la sezione collassabile **Tooltip**
nell'inspector. Il campo `text` imposta il messaggio hover/focus; svuotarlo
rimuove la proprietà.

Il target riceve `aria-describedby` mentre il fumetto è visibile; il fumetto usa
`role="tooltip"` e testo escapato. Focus da tastiera e chiusura con `Escape`
funzionano automaticamente per i target focusabili.

Lo stile dei tooltip dell'app generata si gestisce da **Preferences → Project
preferences.** Il pannello scrive `--sui-tooltip-*` nel CSS del progetto. I
tooltip dell'editor sono separati e si trovano in **Preferences → Editor
preferences → Code editor**.

| Impostazione | Significato |
|---|---|
| Place | `top`, `right`, `bottom` o `left` |
| Delay | ritardo hover in millisecondi |
| Background / text / border | colori del fumetto |
| Radius | raggio del bordo |
| Max width | larghezza massima del fumetto |

Il pannello globale scrive le custom properties `--sui-tooltip-*` in `:root`,
quindi le impostazioni restano nel sorgente e valgono per tutti i tooltip. Il
runtime posiziona il fumetto nel viewport, anche per controlli void come
`input`.

Vedere [API reference: tooltip](API_REFERENCE_IT.md#proprietà-tooltip) e
l'[esempio tooltip](../examples/tooltip/README.md).

## Finestra di build desktop

**File → Build desktop executable.** apre la finestra di build. Il selettore
**Target** sceglie il tipo di finestra:

- **Desktop WebView (native window)** - WebView2 incorporata su Windows,
  WKWebView su macOS e WebKitGTK su Linux; nessuna finestra browser separata.
- **Desktop (default browser)** - l'eseguibile avvia il server loopback e apre
  il browser di sistema.

La finestra verifica Go e la directory di output, incorpora i permessi
`--allow-dir` e `--allow-exec`, mostra il log e permette di scaricare
l'eseguibile. Le build WebView richiedono anche CGO e un compilatore C; su Linux
installare i pacchetti di sviluppo GTK 3 e WebKitGTK 4.0. La prima build scarica
`github.com/webview/webview_go`. Le opzioni runtime sono `--no-open` e `--port N`.

## Dialoghi nativi per file e cartelle

Usare `pickFiles` per un dialogo invece del trascinamento:

```sui
button "Scegli file..." pickFiles: attachment multiple: true
button "Scegli cartella..." pickFiles: attachment directory: true pickMode: "auto"
```

Valori di `pickMode` (default `native`):

- `native` - `<input type=file>`/`webkitdirectory` del browser;
- `fs` - File System Access API (`showDirectoryPicker`);
- `auto` - `fs` quando disponibile, altrimenti `native`.

`fs` e `auto` sono orientati alle cartelle: usano `showDirectoryPicker()` per
restituire un handle di directory e i suoi file. Per la normale selezione di
file lasciare `pickMode: "native"`. Usare `directory: true` quando il dialogo
deve selezionare una cartella.

La FS API richiede un browser Chromium, un contesto sicuro e il permesso
dell'utente. Può restituire il nome di una cartella vuota; il picker nativo
`webkitdirectory` no. Sul target web non esiste comunque un percorso locale
assoluto: sono disponibili solo nomi e percorsi relativi, più i byte tramite
upload.

Nelle **build desktop**, il modulo separato `dialog` restituisce veri percorsi
assoluti:

```sui
import { dialog } from "std"

async fn choose() -> null {
	path, err = dialog.pickFile()
	if err != null { status = err; return }
	if path == "" { status = "annullato"; return }
	status = path
}
```

`dialog.pickFile`, `dialog.pickFiles` e `dialog.pickDirectory` usano i dialoghi
nativi del sistema operativo e sono documentati in
[STDLIB](STDLIB_IT.md#dialog-module-dialog-desktop). Il percorso non è una
capability: `fs` richiede `--allow-dir` e `exec` richiede `--allow-exec`.

## Checklist dell'inspector

Quando un'interazione drag dell'applicazione non funziona, controllare:

1. Il **sorgente** ha `dragKey` o `draggable: true`.
2. Il **target** ha `dropKey`.
3. Il **target** ha `onDrop` che nomina una funzione dichiarata.
4. Il **target file** ha `dropFiles` verso uno stato stringa e, per la vista, un
   binding `value:` corrispondente.
5. La funzione evento esiste ed è disponibile nella view corrente.
6. L'elemento non è una tabella se ci si aspetta l'inspector DnD generico; gli
   interni delle tabelle vanno modificati nel sorgente.
7. Le righe di una `list` e gli interni dei componenti possono richiedere la
   modifica del sorgente.
8. Il folder picking con `fs` richiede Chromium, localhost/HTTPS e permesso.

## Risoluzione problemi

| Sintomo | Causa probabile | Correzione |
|---|---|---|
| Un componente del designer non si trascina | Contenitore non valido, drop nel proprio sottoalbero o drop nell'iframe | Usare l'albero, rilasciare accanto/dentro un contenitore valido |
| La riga generata non è trascinabile | Manca `dragKey` o `draggable: false` | Impostare `dragKey` e `draggable: true` |
| Appare il contorno di drop ma non succede nulla | Manca il gestore `onDrop` | Creare/selezionare una funzione `onDrop` |
| Reagisce solo la prima riga di una lista | Il wiring custom punta a un solo elemento invece che alla riga ripetuta | Mettere `dragKey`/`dropKey`/`onDrop` sulla riga ripetuta; il runtime li emette per ogni riga |
| Il drop file non fa nulla | Lo stato `dropFiles` manca o non è una stringa | Aggiungere uno stato stringa e collegare `dropFiles` |
| Il campo non mostra i percorsi | Manca il binding `value:` | Collegare lo stesso stato con `value:` |
| Il folder picker non ha il nome di una cartella vuota | È stato scelto il picker nativo `webkitdirectory` | Provare `pickMode: "auto"`/`"fs"` su Chromium |
| Il picker FS non è disponibile | Firefox/Safari o contesto non sicuro | Usare `native` o servire su localhost/HTTPS con Chromium |
| Il picker desktop restituisce errore | Build web o ambiente headless | Compilare con `--target desktop` ed eseguire in una sessione desktop |

## Note di sicurezza

- Trattare nomi e percorsi trascinati come testo non fidato.
- Non costruire HTML, CSS o URL da questi dati senza validazione.
- Caricare solo verso un backend sotto il proprio controllo.
- Il picker desktop restituisce percorsi solo dopo una selezione esplicita
  dell'utente; non concede da solo accesso al filesystem.
- Tenere `fs` dietro `--allow-dir` e i comandi dietro `--allow-exec`.
- Il runtime limita numero di file, profondità, percorsi e dati delle tabelle.

## Documenti correlati

- [API reference: drag and drop](API_REFERENCE_IT.md#drag-and-drop)
- [Builtins: drag and drop](BUILTINS_IT.md#drag-and-drop-functions)
- [Agent recipes: drag and drop](AGENT_RECIPES_IT.md#drag-and-drop)
- [Standard library: dialog desktop](STDLIB_IT.md#dialog-module-dialog-desktop)
- [Esempio file picker](../examples/file_picker.sui)
- [Esempio menu bar](../examples/menu_bar.sui)
- [Esempio file analyzer](../examples/file_analyzer/README.md)
- [Esempio desktop picker](../examples/desktop_picker.sui)


