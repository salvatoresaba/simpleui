@echo off
rem download manager launcher (Windows batch)
rem 1. compiles the SUI app to a web build (examples/download_manager/public)
rem 2. starts the download manager (file server + chunked downloader) and the app
rem 3. opens the app in the browser
rem
rem The URL of the file to download is the first parameter: it is the default
rem of the GUI, and the operator can change it there before pressing Start.
rem
rem   run.bat
rem   run.bat "https://example.com/file.bin" "2 MiB"

setlocal
set "ROOT=%~dp0"
set "URL=%~1"
if "%URL%"=="" set "URL=http://127.0.0.1:8090/files/sample.bin"
set "PARTSIZE=%~2"
if "%PARTSIZE%"=="" set "PARTSIZE=1 MiB"
set "WORKERS=%~3"
if "%WORKERS%"=="" set "WORKERS=4"
set "THROTTLE=%~4"
if "%THROTTLE%"=="" set "THROTTLE=0"

"%~dp0..\..\simpleui.exe" build "%ROOT%app.sui" --target web --out "%ROOT%public"
if errorlevel 1 exit /b 1

echo Starting the download manager ...
echo Open this in your browser:  http://127.0.0.1:8090/
echo File to download:  %URL%
cd "%ROOT%server"
go run ./cmd/dm --port 8090 --web-dir "%ROOT%public" --files-dir "%ROOT%files" --out-dir "%ROOT%downloads" --url "%URL%" --part-size "%PARTSIZE%" --workers "%WORKERS%" --throttle "%THROTTLE%"
