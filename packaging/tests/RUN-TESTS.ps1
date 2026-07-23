[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$Tag,
  [ValidateSet('Stable', 'Beta')][string]$Channel = 'Beta',
  [string]$BundleRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')),
  [string]$Commit = 'iso-test',
  [string]$BuildDate = 'iso-test'
)
$ErrorActionPreference = 'Stop'

$root = (Resolve-Path $BundleRoot).Path
$repo = Join-Path $root 'repository'
$results = Join-Path $root 'results'
$dist = Join-Path $results 'dist'
if (-not $env:windir -or -not [Environment]::Is64BitOperatingSystem) { throw 'Run this bundle on Windows x64.' }
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]$identity
if ($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Sign in as a fresh standard user; do not run this entry point elevated.' }
$uac = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System' -Name EnableLUA -ErrorAction SilentlyContinue
if ($uac.EnableLUA -ne 1) { throw 'UAC must be enabled.' }

$go = Get-Command go -ErrorAction SilentlyContinue
if (-not $go) { throw 'Go 1.24.5 is required. Install it from the staged tools or documented source.' }
if ((& go version) -notmatch 'go1\.24\.5') { throw 'The active Go toolchain is not 1.24.5.' }
$wix = Get-Command wix -ErrorAction SilentlyContinue
if (-not $wix -or ((& wix --version 2>$null) -notmatch '4\.0\.5')) { throw 'WiX Toolset 4.0.5 is required.' }
if (-not (Test-Path (Join-Path $repo 'packaging\build.ps1'))) { throw "Bundle repository is incomplete: $repo" }

Remove-Item $results -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $results | Out-Null
& pwsh -NoProfile -File (Join-Path $repo 'packaging\build.ps1') -Tag $Tag -Channel $Channel -Output $dist -Commit $Commit -BuildDate $BuildDate *>&1 |
  Tee-Object (Join-Path $results 'build-result.txt')
if ($LASTEXITCODE -ne 0) { throw 'Packaging build failed.' }
& pwsh -NoProfile -File (Join-Path $repo 'packaging\tests\test-package.ps1') -Dist $dist -Tag $Tag *>&1 |
  Tee-Object (Join-Path $results 'package-result.txt')
if ($LASTEXITCODE -ne 0) { throw 'Package contract test failed.' }

$installRoot = Join-Path $env:LOCALAPPDATA 'wslc-tui-ms'
$msi = Join-Path $dist "wslc-tui-$Tag-windows-amd64.msi"
$bootstrapper = Join-Path $dist "wslc-tui-$Tag-windows-amd64.exe"
& pwsh -NoProfile -File (Join-Path $repo 'packaging\tests\smoke-msi.ps1') -Msi $msi -Bootstrapper $bootstrapper -InstallRoot $installRoot *>&1 |
  Tee-Object (Join-Path $results 'msi-result.txt')
if ($LASTEXITCODE -ne 0) { throw 'MSI smoke test failed.' }
& pwsh -NoProfile -File (Join-Path $repo 'packaging\tests\smoke-bootstrapper.ps1') -Bootstrapper $bootstrapper -InstallRoot $installRoot *>&1 |
  Tee-Object (Join-Path $results 'bootstrapper-result.txt')
if ($LASTEXITCODE -ne 0) { throw 'Bootstrapper smoke test failed.' }

Copy-Item (Join-Path $env:TEMP 'wslc-tui-msi-smoke.log') $results -ErrorAction Stop
Copy-Item (Join-Path $env:TEMP 'wslc-tui-bootstrapper-smoke.log') $results -ErrorAction Stop
'SMOKE_RESULT=passed' | Set-Content (Join-Path $results 'result-marker.txt') -Encoding utf8NoBOM
Write-Output "SMOKE_RESULT=passed; evidence=$results"
