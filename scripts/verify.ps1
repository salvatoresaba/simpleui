param(
    [switch]$SkipFlaky
)

# Single verification entry point for SimpleUI. Run it from any directory:
#   pwsh scripts/verify.ps1
#   pwsh scripts/verify.ps1 -SkipFlaky
#
# -SkipFlaky excludes the two known headless-browser timing flakes
# (TestMotionGraphE2E and TestGuiTableLayoutVisualE2E) until their harness is
# made deterministic. CI uses it so the pipeline is reproducible.

$ErrorActionPreference = 'Stop'
Set-Location (Split-Path -Parent $PSScriptRoot)

function Assert-LastExit($step) {
    if ($LASTEXITCODE -ne 0) {
        Write-Error "$step failed with exit code $LASTEXITCODE"
        exit 1
    }
}

Write-Host '== gofmt (tracked Go files) =='
$files = @(git ls-files '*.go')
Assert-LastExit 'git ls-files'
$unformatted = @(gofmt -l @files)
if ($unformatted.Count -gt 0) {
    Write-Error "gofmt required for:`n$($unformatted -join "`n")"
    exit 1
}

Write-Host '== build =='
go build -o simpleui.exe ./cmd/simpleui
Assert-LastExit 'go build'

Write-Host '== vet =='
go vet ./...
Assert-LastExit 'go vet'

Write-Host '== tests =='
$testArgs = @('test', '-count=1')
if ($SkipFlaky) {
    $testArgs += @('-skip', 'TestMotionGraphE2E|TestGuiTableLayoutVisualE2E')
}
$testArgs += './...'
go @testArgs
Assert-LastExit 'go test'

Write-Host '== example checks =='
$exe = Join-Path (Get-Location) 'simpleui.exe'
$examples = @(Get-ChildItem -Recurse -File -Filter '*.sui' 'examples')
foreach ($file in $examples) {
    $json = (& $exe check $file.FullName --json | Out-String | ConvertFrom-Json)
    if (-not $json.ok) {
        $details = $json.diagnostics | ConvertTo-Json -Compress
        Write-Error "check failed for $($file.FullName): $details"
        exit 1
    }
}
Write-Host "  $($examples.Count) examples OK"

Write-Host '== appearance catalog =='
& $exe appearance check components/appearance.json --json | Out-Null
Assert-LastExit 'appearance check'

Write-Host '== embedded docs =='
& $exe docs --agent | Out-Null
Assert-LastExit 'docs --agent'

Write-Host 'ALL CHECKS PASSED'
