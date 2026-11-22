#!/bin/bash
# SimpleUI GUI Editor Launcher for Linux/macOS
# Avvia l'editor grafico di SimpleUI.
# Compila automaticamente il compilatore se manca, dalla root del repository.

cd "$(dirname "$0")" || exit 1

BIN="./simpleui"

echo "Avvio di SimpleUI GUI Editor..."
echo

# Se il compilatore non c'è, lo compila.
if [ ! -f "$BIN" ]; then
    if ! command -v go >/dev/null 2>&1; then
        echo "Errore: Go non è nel PATH. Installa Go 1.24.2+ e riprova."
        exit 1
    fi
    echo "Il compilatore $BIN non è presente: compilazione in corso..."
    # La root del repository è il package importabile "sui": il binario
    # avviabile si compila da ./cmd/simpleui.
    go build -o "$BIN" ./cmd/simpleui || exit 1
    echo "Compilato $BIN."
    echo
fi

# Rendi eseguibile se necessario
chmod +x "$BIN"

# Avvia l'editor GUI. Gli argomenti extra sono passati al comando gui,
# per esempio: ./GUI_editor.sh percorso/file.sui
"$BIN" gui "$@" &
GUI_PID=$!

echo "Editor GUI avviato (PID: $GUI_PID). Il browser dovrebbe aprirsi automaticamente."
echo "Se il browser non si apre, visita l'URL mostrato nel terminale."
echo
echo "Per terminare: kill $GUI_PID"
wait $GUI_PID
