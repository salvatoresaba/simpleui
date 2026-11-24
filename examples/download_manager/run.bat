@echo off
rem download manager launcher (Windows batch)
rem 1. compiles the SUI app to a web build (examples/download_manager/public)
rem 2. starts the download manager (file server + chunked downloader) and the app
rem 3. opens the app in the browser

set ROOT=%~dp0
"%~dp0..\..\simpleui.exe" build "%ROOT%app.sui" --target web --out "%ROOT%public"
if errorlevel 1 exit /b 1

echo Starting the download manager ...
echo Open this in your browser:  http://127.0.0.1:8090/
cd "%ROOT%server"
go run . --port 8090 --web-dir "%ROOT%public" --files-dir "%ROOT%files" --work-dir "%ROOT%downloads"
