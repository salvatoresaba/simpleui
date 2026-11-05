# Spreadsheet-table demo (PowerShell)
# 1. compiles the SUI app to a web build (examples/spreadsheet_table/web)
# 2. starts sheet2json as an HTTP service (static app + /api/sheet)
# 3. opens the app in the browser
#
# Put your .xlsx / .ods files in examples/spreadsheet_table/data/ first.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$repo = Join-Path $root "..\.."
$sui  = Join-Path $repo "simpleui.exe"

New-Item -ItemType Directory -Force (Join-Path $root "data") | Out-Null

& $sui build (Join-Path $root "app.sui") --target web --out (Join-Path $root "web")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Push-Location $repo
go build -o (Join-Path $root "sheet2json.exe") ./examples/spreadsheet_table/sheet2json
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Write-Host "Serving on http://127.0.0.1:8099/"
Write-Host "Put .xlsx/.ods files in: $(Join-Path $root 'data')"
Start-Process "http://127.0.0.1:8099/"
& (Join-Path $root "sheet2json.exe") -serve -port 8099 -dir (Join-Path $root "data") -web-dir (Join-Path $root "web")
