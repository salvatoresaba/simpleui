# download manager launcher (PowerShell)
# 1. compiles the SUI app to a web build (examples/download_manager/public)
# 2. starts the download manager (file server + chunked downloader) and the app
# 3. opens the app in the browser
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$sui  = Join-Path $root "..\..\simpleui.exe"

& $sui build (Join-Path $root "app.sui") --target web --out (Join-Path $root "public")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Starting the download manager ..."
Write-Host "Open this in your browser:  http://127.0.0.1:8090/"
Push-Location (Join-Path $root "server")
go run . --port 8090 --web-dir (Join-Path $root "public") --files-dir (Join-Path $root "files") --work-dir (Join-Path $root "downloads")
Pop-Location
