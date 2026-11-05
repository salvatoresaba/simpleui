@echo off
setlocal
set "ROOT=%~dp0"
set "REPO=%ROOT%..\.."
set "SUI=%REPO%\simpleui.exe"
set "SRC=%REPO%\examples\gantt_calendar\app.sui"
set "OUT=%ROOT%dist"

"%SUI%" build "%SRC%" --target desktop --out "%OUT%"
if errorlevel 1 exit /b %errorlevel%
copy /Y "%SRC%" "%OUT%\app.sui" >nul

if not exist "%OUT%\app.exe" (
  echo Build succeeded but app.exe was not created in %OUT%
  exit /b 1
)
start "" "%OUT%\app.exe"
