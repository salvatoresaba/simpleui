# Build the animated GIF spinner example and open it in the browser.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$repo = Join-Path $root "..\.."
$sui = Join-Path $repo "simpleui.exe"
$src = Join-Path $root "app.sui"
$out = Join-Path $root "dist"

if (!(Test-Path $sui)) {
    throw "simpleui.exe not found in $repo. Build it first with: go build -o simpleui.exe ./cmd/simpleui"
}

New-Item -ItemType Directory -Force $out | Out-Null
$html = Join-Path $out "app.html"
& $sui build $src --single --out $html
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Copy-Item -Force $src (Join-Path $out "app.sui")
Start-Process $html
