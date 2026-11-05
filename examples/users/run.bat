@echo off
rem users demo launcher (Windows batch)
rem 1. compiles the SUI app to a web build (examples/users/public)
rem 2. starts the JSON API + static server
rem 3. opens the app in the browser

set ROOT=%~dp0
"%~dp0..\..\simpleui.exe" build "%ROOT%app.sui" --target web --out "%ROOT%public"
if errorlevel 1 exit /b 1

echo Starting local server (JSON API + app) ...
echo Open this in your browser:  http://127.0.0.1:8090/
cd "%ROOT%server"
go run . --port 8090 --web-dir "%ROOT%public"