# download_manager release build (PowerShell)
# Produces self-contained executables with the app AND the download manager
# inside: the compiler mounts the Go backend of the example in front of the
# compiled app, so the single file works with no other file next to it.
#
#   pwsh build.ps1
#
# The executables land in bin/:
#
#   bin/download_manager.exe                Windows amd64
#   bin/download_manager_linux_amd64        Linux amd64
#   bin/download_manager_darwin_arm64       macOS Apple Silicon
#
# Run it with no parameter and the app opens in the browser with the sample
# file of the manager as the file to download.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$sui = Join-Path $root "..\..\simpleui.exe"
$server = Join-Path $root "server"
$bin = Join-Path $root "bin"

New-Item -ItemType Directory -Force -Path $bin | Out-Null

function Build-Desktop($label, $goos, $goarch, $name) {
    $out = Join-Path $root ("dist\" + $label)
    Write-Host "building the desktop app for $goos/$goarch ..."
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    & $sui build (Join-Path $root "app.sui") --target desktop --backend $server --out $out
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    # The compiler writes the executable as app(.exe): the name of the
    # release is the one of the platform.
    $built = Join-Path $out "app.exe"
    if (-not (Test-Path $built)) { $built = Join-Path $out "app" }
    Copy-Item -Force $built (Join-Path $bin $name)
}

try {
    Build-Desktop "windows" "windows" "amd64" "download_manager.exe"
    Build-Desktop "linux" "linux" "amd64" "download_manager_linux_amd64"
    Build-Desktop "darwin" "darwin" "arm64" "download_manager_darwin_arm64"
} finally {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
}

Get-ChildItem $bin | Select-Object Name, Length
Write-Host "Run: .\bin\download_manager.exe"
