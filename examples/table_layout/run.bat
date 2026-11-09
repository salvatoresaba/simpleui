@echo off
setlocal
set "ROOT=%~dp0"
set "REPO=%ROOT%..\.."
set "SUI=%REPO%\simpleui.exe"
set "SRC=%REPO%\examples\table_layout\app.sui"
set "OUT=%ROOT%dist"

"%SUI%" build "%SRC%" --target desktop --out "%OUT%"
if errorlevel 1 exit /b %errorlevel%
copy /Y "%SRC%" "%OUT%\app.sui" >nul
start "" "%OUT%\app.exe"
