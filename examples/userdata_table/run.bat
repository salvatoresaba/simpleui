@echo off
rem CLI-table demo (Windows): an executable answers with the table as JSON,
rem and the desktop SUI client launches it through the bridge (exec.run).

set ROOT=%~dp0

rem 1. build the "server executable" that prints the JSON table
cd "%ROOT%userdata"
go build -o "%ROOT%dist\userdata.exe" .

rem 2. compile the SUI client to a desktop app, granting the command
"%ROOT%..\..\simpleui.exe" build "%ROOT%app.sui" --target desktop --out "%ROOT%dist" --allow-exec userdata.exe
if errorlevel 1 exit /b 1

rem 3. run it (opens the system browser with the app)
"%ROOT%dist\app.exe"