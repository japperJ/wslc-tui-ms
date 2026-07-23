[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$Tag,
  [ValidateSet('Stable', 'Beta')][string]$Channel = 'Beta',
  [string]$Commit = 'iso-test',
  [string]$BuildDate = 'iso-test'
)
$ErrorActionPreference = 'Stop'

$root = (Resolve-Path $PSScriptRoot).Path
$repo = Join-Path $root 'repository'
$artifacts = Join-Path $root 'artifacts'
$results = Join-Path $root 'results'
$required = @(
  "wslc-tui-$Tag-windows-amd64.msi",
  "wslc-tui-$Tag-windows-amd64.exe",
  "wslc-tui-$Tag-windows-amd64-portable.zip",
  'update-policy.json',
  "wslc-tui-$Tag-checksums.json",
  "wslc-tui-$Tag-metadata.json"
)
$missing = @($required | Where-Object { -not (Test-Path (Join-Path $artifacts $_)) })
if ($missing.Count -gt 0) {
  throw "Missing packaged artifacts: $($missing -join ', '). Build the artifacts separately and rebuild this ISO."
}
if (-not $env:windir -or -not [Environment]::Is64BitOperatingSystem) { throw 'Run this bundle on Windows x64.' }
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]$identity
if ($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Sign in as a fresh standard user; do not run this entry point elevated.' }
$uac = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System' -Name EnableLUA -ErrorAction SilentlyContinue
if ($uac.EnableLUA -ne 1) { throw 'UAC must be enabled.' }

Remove-Item $results -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $results | Out-Null
$manifest = Get-Content (Join-Path $artifacts "wslc-tui-$Tag-checksums.json") -Raw | ConvertFrom-Json
if ($manifest.releaseTag -ne $Tag -or $manifest.algorithm -ne 'sha256') { throw 'Packaged checksum metadata does not match the requested tag.' }
$metadata = Get-Content (Join-Path $artifacts "wslc-tui-$Tag-metadata.json") -Raw | ConvertFrom-Json
if ($metadata.releaseTag -ne $Tag -or (@($metadata.assets | Sort-Object) -join '|') -ne (@($required | Select-Object -First 5 | Sort-Object) -join '|')) {
  throw 'Packaged artifact metadata does not match the requested artifact names.'
}
foreach ($name in $required[0..4]) {
  $entry = @($manifest.assets | Where-Object name -eq $name)
  if ($entry.Count -ne 1) { throw "Checksum manifest does not cover exactly one $name artifact." }
  if ((Get-FileHash (Join-Path $artifacts $name) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $entry[0].sha256) { throw "Checksum mismatch: $name" }
}

$smokeRoot = Join-Path $root 'tests'
$installRoot = Join-Path $env:LOCALAPPDATA 'wslc-tui-ms'
$msi = Join-Path $artifacts "wslc-tui-$Tag-windows-amd64.msi"
$bootstrapper = Join-Path $artifacts "wslc-tui-$Tag-windows-amd64.exe"
$portable = Join-Path $artifacts "wslc-tui-$Tag-windows-amd64-portable.zip"
& pwsh -NoProfile -File (Join-Path $smokeRoot 'smoke-portable.ps1') -Portable $portable -Tag $Tag -ExtractRoot (Join-Path $results 'portable') *>&1 |
  Tee-Object (Join-Path $results 'portable-result.txt')
if ($LASTEXITCODE -ne 0) { throw 'Portable smoke test failed.' }
& pwsh -NoProfile -File (Join-Path $smokeRoot 'smoke-msi.ps1') -Msi $msi -Bootstrapper $bootstrapper -InstallRoot $installRoot *>&1 |
  Tee-Object (Join-Path $results 'msi-result.txt')
if ($LASTEXITCODE -ne 0) { throw 'MSI smoke test failed.' }
& pwsh -NoProfile -File (Join-Path $smokeRoot 'smoke-bootstrapper.ps1') -Bootstrapper $bootstrapper -InstallRoot $installRoot *>&1 |
  Tee-Object (Join-Path $results 'bootstrapper-result.txt')
if ($LASTEXITCODE -ne 0) { throw 'Bootstrapper smoke test failed.' }

Copy-Item (Join-Path $env:TEMP 'wslc-tui-msi-smoke.log') $results -ErrorAction Stop
Copy-Item (Join-Path $env:TEMP 'wslc-tui-bootstrapper-smoke.log') $results -ErrorAction Stop
'SMOKE_RESULT=passed' | Set-Content (Join-Path $results 'result-marker.txt') -Encoding utf8NoBOM
Write-Output "SMOKE_RESULT=passed; evidence=$results"
