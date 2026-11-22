@echo off
REM SimpleUI GUI Editor Launcher for Windows
REM Avvia l'editor grafico di SimpleUI.
REM Compila automaticamente il compilatore se manca, dalla root del repository.

setlocal
cd /d "%~dp0"

set "BIN=simpleui.exe"

echo Avvio di SimpleUI GUI Editor...
echo.

if not exist "%BIN%" (
    echo Il compilatore %BIN% non e' presente: compilazione in corso...
    where go >nul 2>nul
    if errorlevel 1 (
        echo Errore: Go non e' nel PATH. Installa Go 1.24.2+ e riprova.
        pause
        exit /b 1
    )
    REM La root del repository e' il package importabile "sui": il binario
    REM avviabile si compila da ./cmd/simpleui.
    go build -o "%BIN%" ./cmd/simpleui
    if errorlevel 1 (
        echo Errore: compilazione fallita.
        pause
        exit /b 1
    )
    echo Compilato %BIN%.
    echo.
)

REM Avvia l'editor GUI. Gli argomenti extra sono passati al comando gui,
REM per esempio: GUI_editor.bat percorso\file.sui
start "" "%BIN%" gui %*

echo Editor GUI avviato. Il browser dovrebbe aprirsi automaticamente.
echo Se il browser non si apre, visita l'URL mostrato nel terminale.
echo.
pause
