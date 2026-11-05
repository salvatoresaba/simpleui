# file_analyzer release build (PowerShell)
# Produces self-contained executables with the SUI frontend embedded.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$sui = Join-Path $root "..\..\simpleui.exe"
$server = Join-Path $root "server"
$bin = Join-Path $root "bin"

Write-Host "1/3 building the SimpleUI frontend into server/web ..."
& $sui build (Join-Path $root "app.sui") --target web --out (Join-Path $server "web")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "2/3 building the Windows executable ..."
New-Item -ItemType Directory -Force -Path $bin | Out-Null
Push-Location $server
try {
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags "-s -w" -o (Join-Path $bin "file_analyzer.exe") .
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    Write-Host "3/3 building the Linux and macOS executables ..."
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags "-s -w" -o (Join-Path $bin "file_analyzer_linux_amd64") .
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    $env:GOOS = "darwin"
    $env:GOARCH = "arm64"
    go build -trimpath -ldflags "-s -w" -o (Join-Path $bin "file_analyzer_darwin_arm64") .
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} finally {
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Pop-Location
}

Get-ChildItem $bin | Select-Object Name, Length
Write-Host "Run: .\bin\file_analyzer.exe --port 8097"
