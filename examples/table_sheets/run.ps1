# Rebuild the sheet-tabs workbook and launch the desktop demo.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$repo = Join-Path $root "..\.."
$sui = Join-Path $repo "simpleui.exe"
$src = Join-Path $repo "examples\table_sheets\app.sui"
$out = Join-Path $root "dist"

& $sui build $src --target desktop --out $out
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Copy-Item -Force $src (Join-Path $out "app.sui")
Start-Process (Join-Path $out "app.exe")
