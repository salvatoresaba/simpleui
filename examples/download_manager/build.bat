@echo off
rem download_manager release build (Windows batch)
rem Produces self-contained executables with the SUI frontend embedded: the
rem single file carries the app and the download manager.
setlocal
set ROOT=%~dp0
set SUI=%ROOT%..\..\simpleui.exe
set SERVER=%ROOT%server
set BIN=%ROOT%bin

echo 1/3 building the SimpleUI frontend into server\web ...
"%SUI%" build "%ROOT%app.sui" --target web --out "%SERVER%\web"
if errorlevel 1 exit /b 1

echo 2/3 building the Windows executable ...
if not exist "%BIN%" mkdir "%BIN%"
cd /d "%SERVER%"

set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BIN%\download_manager.exe" .
if errorlevel 1 exit /b 1

echo 3/3 building the Linux and macOS executables ...
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BIN%\download_manager_linux_amd64" .
if errorlevel 1 exit /b 1

set GOOS=darwin
set GOARCH=arm64
go build -trimpath -ldflags "-s -w" -o "%BIN%\download_manager_darwin_arm64" .
if errorlevel 1 exit /b 1

dir "%BIN%"
echo Run: %BIN%\download_manager.exe --port 8090
endlocal
