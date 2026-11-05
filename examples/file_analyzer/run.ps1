# file_analyzer launcher (PowerShell)
# 1. compiles the SUI app into public/
# 2. starts the Go backend + static server
# 3. opens the browser
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$sui = Join-Path $root "..\..\simpleui.exe"

& $sui build (Join-Path $root "app.sui") --target web --out (Join-Path $root "public")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Starting file analyzer (Go backend + app) ..."
Push-Location (Join-Path $root "server")
go run . --port 8097 --web-dir (Join-Path $root "public")
Pop-Location
