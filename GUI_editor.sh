#!/bin/bash
# SimpleUI GUI Editor Launcher for Linux/macOS
# Avvia l'editor grafico di SimpleUI

echo "Avvio di SimpleUI GUI Editor..."
echo

# Verifica se simpleui esiste
if [ ! -f "./simpleui" ]; then
    echo "Errore: simpleui non trovato nella directory corrente."
    echo "Assicurati di aver compilato il progetto con: go build -o simpleui"
    exit 1
fi

# Rendi eseguibile se necessario
chmod +x ./simpleui

# Avvia l'editor GUI
./simpleui gui &
GUI_PID=$!

echo "Editor GUI avviato (PID: $GUI_PID). Il browser dovrebbe aprirsi automaticamente."
echo "Se il browser non si apre, visita l'URL mostrato nel terminale."
echo
echo "Per terminare: kill $GUI_PID"
wait $GUI_PID
