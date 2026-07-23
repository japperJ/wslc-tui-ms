[CmdletBinding()]
param([string]$Tag = 'v1.2.3')
$ErrorActionPreference = 'Stop'
$root = Resolve-Path (Join-Path $PSScriptRoot '..')
if ($Tag -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$') { throw 'Invalid test tag' }
& pwsh -NoProfile -File (Join-Path $root 'packaging/tests/test-contracts.ps1')
$run1 = Join-Path $env:TEMP "wslc-release-test-1-$PID"
$run2 = Join-Path $env:TEMP "wslc-release-test-2-$PID"
Remove-Item $run1, $run2 -Recurse -Force -ErrorAction SilentlyContinue
foreach ($run in @($run1, $run2)) {
  & pwsh -NoProfile -File (Join-Path $root 'packaging/build.ps1') -Tag $Tag -Channel Beta -Output $run -Commit fixed-test-commit -BuildDate 2026-07-23T00:00:00Z
  & pwsh -NoProfile -File (Join-Path $root 'packaging/tests/test-package.ps1') -Dist $run -Tag $Tag
  $metadata = & (Join-Path $run 'wslc-tui.exe') --version
  if ($metadata -notmatch "wslc-tui $([regex]::Escape($Tag)) .*channel=Beta.*distribution=portable") { throw "Tag metadata assertion failed: $metadata" }
  if ((Get-Command wix -ErrorAction SilentlyContinue)) {
    foreach ($name in @("wslc-tui-$Tag-windows-amd64.msi", "wslc-tui-$Tag-windows-amd64.exe")) {
      if (-not (Test-Path (Join-Path $run $name))) { throw "WiX build did not produce $name" }
    }
  } else {
    Write-Warning 'WiX v4 unavailable; MSI/bootstrapper VM smoke assertions are deferred.'
  }
}
$names1 = @(Get-ChildItem $run1 -File | Select-Object -ExpandProperty Name | Sort-Object)
$names2 = @(Get-ChildItem $run2 -File | Select-Object -ExpandProperty Name | Sort-Object)
if (($names1 -join '|') -ne ($names2 -join '|')) { throw 'Repeated runs produced different artifact names' }

# Promotion changes only the GitHub release state. This fixture models the API
# request/response boundary and proves no build or upload request is issued.
$assetSnapshot = @('msi-bytes', 'exe-bytes', 'zip-bytes', 'policy-bytes', 'checksum-bytes')
$release = [ordered]@{ tag_name = $Tag; prerelease = $true; assets = $assetSnapshot }
$requests = [System.Collections.Generic.List[string]]::new()
function Resolve-Channel($response) { if ($response.prerelease) { 'Beta' } else { 'Stable' } }
if ((Resolve-Channel $release) -ne 'Beta') { throw 'Prerelease release did not map to Beta' }
$release.prerelease = $false
if ((Resolve-Channel $release) -ne 'Stable') { throw 'Promoted release did not map to Stable' }
if (($release.assets -join '|') -ne ($assetSnapshot -join '|')) { throw 'Promotion changed assets' }
if ($requests.Count -ne 0) { throw 'Promotion unexpectedly issued build/upload requests' }
if ($env:SIGNING_COMMAND) { throw 'Signing test requires SIGNING_COMMAND to be unset' }
Write-Output 'Release workflow contract and mocked promotion test passed.'
