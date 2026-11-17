Set-Location $PSScriptRoot
Write-Host "Rebuilding SimpleUI Networking Suite..."
& ..\..\simpleui.exe build app.sui --single --out dist\app.html
Write-Host "Starting server with live watch..."
& ..\..\simpleui.exe serve app.sui --watch
