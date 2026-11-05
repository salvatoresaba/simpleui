@echo off
REM SimpleUI GUI Editor Launcher for Windows
REM Avvia l'editor grafico di SimpleUI

echo Avvio di SimpleUI GUI Editor...
echo.

REM Verifica se simpleui.exe esiste
if not exist "simpleui.exe" (
    echo Errore: simpleui.exe non trovato nella directory corrente.
    echo Assicurati di aver compilato il progetto con: go build -o simpleui.exe
    pause
    exit /b 1
)

REM Avvia l'editor GUI
start "" simpleui.exe gui

echo Editor GUI avviato. Il browser dovrebbe aprirsi automaticamente.
echo Se il browser non si apre, visita l'URL mostrato nel terminale.
echo.
pause
