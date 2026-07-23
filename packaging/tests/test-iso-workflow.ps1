[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'

$root = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$builder = Join-Path $root 'scripts/build-phase1-test-iso.ps1'
$runner = Get-Content (Join-Path $root 'packaging/tests/RUN-TESTS.ps1') -Raw

if ($runner -notmatch '\$PSScriptRoot') { throw 'ISO runner must derive its bundle root from $PSScriptRoot.' }
if ($runner -match 'Get-Command\s+go|Get-Command\s+wix|packaging\\build\.ps1') {
  throw 'ISO runner must not require developer packaging tools or build from source.'
}

$temp = Join-Path ([IO.Path]::GetTempPath()) "wslc-iso-test-$PID"
Remove-Item $temp -Recurse -Force -ErrorAction SilentlyContinue
try {
  & pwsh -NoProfile -File $builder -Tag v1.2.3 -OutputDirectory $temp -SkipIso
  if ($LASTEXITCODE -ne 0) { throw 'Source-only ISO bundle creation failed.' }
  $bundle = Join-Path $temp 'wslc-tui-ms-phase1-packaging-smoke-v1.2.3'
  Push-Location ([IO.Path]::GetTempPath())
  try { $output = & pwsh -NoProfile -File (Join-Path $bundle 'RUN-TESTS.ps1') -Tag v1.2.3 2>&1 } finally { Pop-Location }
  if (($output -join "`n") -notmatch 'Missing packaged artifacts') {
    throw "Missing-artifact smoke failure was not concise: $($output -join ' | ')"
  }
  if (($output -join "`n") -match 'Go|WiX|wix') {
    throw "Missing-artifact failure resolved developer prerequisites: $($output -join ' | ')"
  }

  $sourceArtifacts = Join-Path $temp 'prebuilt'
  New-Item -ItemType Directory -Force -Path $sourceArtifacts | Out-Null
  foreach ($name in @(
    'wslc-tui-v1.2.3-windows-amd64.msi',
    'wslc-tui-v1.2.3-windows-amd64.exe',
    'wslc-tui-v1.2.3-windows-amd64-portable.zip',
    'update-policy.json'
  )) { Set-Content (Join-Path $sourceArtifacts $name) 'fixture' }
  $stagedRoot = Join-Path $temp 'staged'
  & pwsh -NoProfile -File $builder -Tag v1.2.3 -OutputDirectory $stagedRoot -ArtifactDirectory $sourceArtifacts -SkipIso
  if ($LASTEXITCODE -ne 0) { throw 'Prebuilt artifact bundle creation failed.' }
  $staged = Join-Path $stagedRoot 'wslc-tui-ms-phase1-packaging-smoke-v1.2.3\artifacts'
  foreach ($name in @(
    'wslc-tui-v1.2.3-windows-amd64.msi',
    'wslc-tui-v1.2.3-windows-amd64.exe',
    'wslc-tui-v1.2.3-windows-amd64-portable.zip',
    'update-policy.json',
    'wslc-tui-v1.2.3-checksums.json',
    'wslc-tui-v1.2.3-metadata.json'
  )) { if (-not (Test-Path (Join-Path $staged $name))) { throw "Staged artifact missing: $name" } }
} finally {
  Remove-Item $temp -Recurse -Force -ErrorAction SilentlyContinue
}
Write-Output 'ISO workflow tests passed.'
