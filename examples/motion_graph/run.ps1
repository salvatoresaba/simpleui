# Rebuild the complete motion graph example and launch the desktop demo.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$repo = Join-Path $root "..\.."
$sui = Join-Path $repo "simpleui.exe"
$src = Join-Path $repo "examples\motion_graph\app.sui"
$out = Join-Path $root "dist"

& $sui build $src --target desktop --out $out
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Copy-Item -Force $src (Join-Path $out "app.sui")

$app = Join-Path $out "app.exe"
if (!(Test-Path $app)) {
    throw "Build succeeded but app.exe was not created in $out"
}
Start-Process $app