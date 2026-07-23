[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$Portable,
  [Parameter(Mandatory = $true)][string]$Tag,
  [Parameter(Mandatory = $true)][string]$ExtractRoot
)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path $Portable -PathType Leaf)) { throw "Portable artifact missing: $Portable" }
Remove-Item $ExtractRoot -Recurse -Force -ErrorAction SilentlyContinue
Expand-Archive -LiteralPath $Portable -DestinationPath $ExtractRoot -Force
$exe = Join-Path $ExtractRoot 'wslc-tui.exe'
if (-not (Test-Path $exe)) { throw 'Portable archive did not contain wslc-tui.exe at its root.' }
$metadata = & $exe --version
if ($metadata -notmatch "wslc-tui $([regex]::Escape($Tag)) .*distribution=portable") { throw "Portable metadata mismatch: $metadata" }
Write-Output 'Portable smoke passed.'
