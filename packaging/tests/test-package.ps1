[CmdletBinding()]
param([Parameter(Mandatory = $true)][string]$Dist, [Parameter(Mandatory = $true)][string]$Tag)
$ErrorActionPreference = 'Stop'
$expectedZip = "wslc-tui-$Tag-windows-amd64-portable.zip"
$zip = Join-Path $Dist $expectedZip
if (-not (Test-Path $zip)) { throw "Missing portable archive: $zip" }
$entries = @(tar -tf $zip | Where-Object { $_ -and $_ -notmatch '/$' } | ForEach-Object { $_.Replace('\', '/') })
if (($entries | Sort-Object) -join '|' -ne 'LICENSE.txt|README.txt|wslc-tui-updater.exe|wslc-tui.exe') { throw "Unexpected portable ZIP contents: $($entries -join ', ')" }
$version = & (Join-Path $Dist 'wslc-tui.exe') --version
if ($version -notmatch "wslc-tui $([regex]::Escape($Tag)) .*distribution=portable") { throw "Portable metadata mismatch: $version" }
$msi = Join-Path $Dist "wslc-tui-$Tag-windows-amd64.msi"
$bundle = Join-Path $Dist "wslc-tui-$Tag-windows-amd64.exe"
if ((Test-Path $msi) -and (Test-Path $bundle)) {
  $installerBinary = Join-Path $Dist 'wslc-tui-installer.exe'
  if (-not (Test-Path $installerBinary)) { throw 'Installer metadata binary missing from build staging directory' }
  $installerMetadata = & $installerBinary --version
  if ($installerMetadata -notmatch "wslc-tui $([regex]::Escape($Tag)) .*distribution=installer") { throw "Installer metadata mismatch: $installerMetadata" }
  Write-Output 'MSI and bootstrapper artifacts are present; run smoke-msi.ps1 on a clean standard-user VM.'
} else {
  Write-Warning 'WiX artifacts unavailable; portable package checks passed.'
}
