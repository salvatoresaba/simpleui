@echo off
rem Spreadsheet-table demo (Windows)
rem 1. compiles the SUI app to a web build (examples\spreadsheet_table\web)
rem 2. starts sheet2json as an HTTP service (static app + /api/sheet)
rem 3. opens the app in the browser
rem
rem Put your .xlsx / .ods files in examples\spreadsheet_table\data\ first.

set ROOT=%~dp0
set REPO=%ROOT%..\..\
if not exist "%ROOT%data" mkdir "%ROOT%data"

"%REPO%simpleui.exe" build "%ROOT%app.sui" --target web --out "%ROOT%web"
if errorlevel 1 exit /b 1

pushd "%REPO%"
go build -o "%ROOT%sheet2json.exe" ./examples/spreadsheet_table/sheet2json
if errorlevel 1 (popd & exit /b 1)
popd

echo Serving on http://127.0.0.1:8099/
echo Put .xlsx/.ods files in: %ROOT%data
start "" http://127.0.0.1:8099/
"%ROOT%sheet2json.exe" -serve -port 8099 -dir "%ROOT%data" -web-dir "%ROOT%web"
