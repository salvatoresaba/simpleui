# sync-readme-sui.ps1 — keep the SUI block of a README in sync with its source.
#
# The examples embed the complete SUI source in their README (the agent
# checklist of AGENTS.md). When the source changes, the block must change
# with it: this script rewrites the block from the file.
#
#   pwsh tools/sync-readme-sui.ps1 `
#     -Readme examples/download_manager/README.md `
#     -Sui    examples/download_manager/app.sui
#
# The block is the one fenced with "sui"; it is replaced with the content of
# the SUI file and the line endings of the README are preserved.
param(
    [Parameter(Mandatory = $true)]
    [string]$Readme,
    [Parameter(Mandatory = $true)]
    [string]$Sui
)

$ErrorActionPreference = 'Stop'

$readmePath = (Resolve-Path $Readme).Path
$suiPath = (Resolve-Path $Sui).Path

$readme = [System.IO.File]::ReadAllText($readmePath)
$src = [System.IO.File]::ReadAllText($suiPath)

$fence = ([string][char]96) * 3
$pattern = [regex]::Escape($fence) + 'sui\r?\n.*?\r?\n' + [regex]::Escape($fence)
$match = [regex]::Match($readme, '(?s)' + $pattern)
if (-not $match.Success) { throw "no sui block found in $readmePath" }

# The block keeps the line endings of the text it replaces.
$nl = if ($match.Value.Contains("`r`n")) { "`r`n" } else { "`n" }
$body = $src.TrimEnd("`r", "`n")
$new = $fence + 'sui' + $nl + $body + $nl + $fence
$readme = $readme.Substring(0, $match.Index) + $new + $readme.Substring($match.Index + $match.Length)
[System.IO.File]::WriteAllText($readmePath, $readme, (New-Object System.Text.UTF8Encoding($false)))

Write-Output "synced the sui block of $readmePath with $suiPath"
