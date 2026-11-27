@echo off
rem download_manager release build (Windows batch)
rem Produces self-contained executables with the app AND the download manager
rem inside: the compiler mounts the Go backend of the example in front of the
rem compiled app, so the single file works with no other file next to it.
setlocal
set "ROOT=%~dp0"
set "SUI=%ROOT%..\..\simpleui.exe"
set "SERVER=%ROOT%server"
set "BIN=%ROOT%bin"

if not exist "%BIN%" mkdir "%BIN%"

echo building the desktop app for windows/amd64 ...
"%SUI%" build "%ROOT%app.sui" --target desktop --backend "%SERVER%" --out "%ROOT%dist\windows"
if errorlevel 1 exit /b 1
copy /y "%ROOT%dist\windows\app.exe" "%BIN%\download_manager.exe" >nul

echo building the desktop app for linux/amd64 ...
set GOOS=linux
set GOARCH=amd64
"%SUI%" build "%ROOT%app.sui" --target desktop --backend "%SERVER%" --out "%ROOT%dist\linux"
if errorlevel 1 exit /b 1
copy /y "%ROOT%dist\linux\app.exe" "%BIN%\download_manager_linux_amd64" >nul

echo building the desktop app for darwin/arm64 ...
set GOOS=darwin
set GOARCH=arm64
"%SUI%" build "%ROOT%app.sui" --target desktop --backend "%SERVER%" --out "%ROOT%dist\darwin"
if errorlevel 1 exit /b 1
copy /y "%ROOT%dist\darwin\app.exe" "%BIN%\download_manager_darwin_arm64" >nul

dir "%BIN%"
echo Run: %BIN%\download_manager.exe
endlocal
