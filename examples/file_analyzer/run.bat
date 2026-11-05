@echo off
rem file_analyzer launcher (Windows batch)
rem 1. compiles the SUI app into public/
rem 2. starts the Go backend + static server
rem 3. opens the browser

set ROOT=%~dp0
"%~dp0..\..\simpleui.exe" build "%ROOT%app.sui" --target web --out "%ROOT%public"
if errorlevel 1 exit /b 1

echo Starting file analyzer (Go backend + app) ...
cd "%ROOT%server"
go run . --port 8097 --web-dir "%ROOT%public"
