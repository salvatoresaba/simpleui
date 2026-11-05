# CLI-table demo launcher (PowerShell): an executable answers with the table
# as JSON, and the desktop SUI client launches it through the bridge.
$ErrorActionPreference = "Stop"
$root  = Split-Path -Parent $MyInvocation.MyCommand.Path
$sui   = Join-Path $root "..\..\simpleui.exe"

# 1. build the "server executable" that prints the JSON table
Push-Location (Join-Path $root "userdata")
go build -o (Join-Path $root "dist\userdata.exe") .
Pop-Location

# 2. compile the SUI client to a desktop app, granting the command
& $sui build (Join-Path $root "app.sui") --target desktop --out (Join-Path $root "dist") --allow-exec userdata.exe
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# 3. run it (opens the system browser with the app)
& (Join-Path $root "dist\app.exe")