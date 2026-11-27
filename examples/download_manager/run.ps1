# download manager launcher (PowerShell)
# 1. compiles the SUI app to a web build (examples/download_manager/public)
# 2. starts the download manager (file server + chunked downloader) and the app
# 3. opens the app in the browser
#
# The URL of the file to download is a parameter: it is the default of the
# GUI, and the operator can change it there before pressing Start.
#
#   pwsh run.ps1
#   pwsh run.ps1 -Url "https://example.com/file.bin" -PartSize "2 MiB"
#   pwsh run.ps1 -Url "https://example.com/file.bin" -AllowHost "example.com"
#   pwsh run.ps1 -NoOpen
param(
    [string]$Url = "http://127.0.0.1:8090/files/sample.bin",
    [string]$PartSize = "1 MiB",
    [int]$Workers = 4,
    [int]$Throttle = 0,
    [string]$OutDir = "",
    [string[]]$AllowHost = @(),
    [switch]$NoOpen
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$sui  = Join-Path $root "..\..\simpleui.exe"

& $sui build (Join-Path $root "app.sui") --target web --out (Join-Path $root "public")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if ($OutDir -eq "") { $OutDir = Join-Path $root "downloads" }

$backendArgs = @(
    "--port", "8090",
    "--web-dir", (Join-Path $root "public"),
    "--files-dir", (Join-Path $root "files"),
    "--out-dir", $OutDir,
    "--url", $Url,
    "--part-size", $PartSize,
    "--workers", "$Workers",
    "--throttle", "$Throttle"
)
foreach ($allowed in $AllowHost) { $backendArgs += @("--allow-host", $allowed) }
if ($NoOpen) { $backendArgs += "--no-open" }

Write-Host "Starting the download manager ..."
Write-Host "Open this in your browser:  http://127.0.0.1:8090/"
Write-Host "File to download:  $Url"
Push-Location (Join-Path $root "server")
go run ./cmd/dm @backendArgs
$code = $LASTEXITCODE
Pop-Location
exit $code
