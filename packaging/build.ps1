[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$Tag,
  [ValidateSet('Stable', 'Beta')][string]$Channel = 'Stable',
  [string]$Output = (Join-Path (Get-Location) 'dist'),
  [string]$Commit = 'local',
  [string]$BuildDate = 'unknown'
)
$ErrorActionPreference = 'Stop'
if ($Tag -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$') { throw "Tag must be a SemVer tag: $Tag" }
$version = $Tag.Substring(1)
$root = Resolve-Path (Join-Path $PSScriptRoot '..')
$bin = Join-Path $Output 'wslc-tui.exe'
$installerBin = Join-Path $Output 'wslc-tui-installer.exe'
$portable = Join-Path $Output "wslc-tui-$Tag-windows-amd64-portable"
$msi = Join-Path $Output "wslc-tui-$Tag-windows-amd64.msi"
$bundle = Join-Path $Output "wslc-tui-$Tag-windows-amd64.exe"
$zip = Join-Path $Output "wslc-tui-$Tag-windows-amd64-portable.zip"
New-Item -ItemType Directory -Force -Path $Output, $portable | Out-Null

$ldflags = "-s -w -X wslc-tui-ms/internal/buildinfo.Version=$Tag -X wslc-tui-ms/internal/buildinfo.Commit=$Commit -X wslc-tui-ms/internal/buildinfo.BuildDate=$BuildDate -X wslc-tui-ms/internal/buildinfo.Channel=$Channel -X wslc-tui-ms/internal/buildinfo.Distribution="
$env:GOOS = 'windows'; $env:GOARCH = 'amd64'
& go build -trimpath -ldflags "$ldflags`portable" -o $bin $root
if ($LASTEXITCODE -ne 0) { throw 'Portable Go build failed' }
Copy-Item $bin (Join-Path $portable 'wslc-tui.exe') -Force
Copy-Item (Join-Path $root 'packaging/assets/README.txt') (Join-Path $portable 'README.txt') -Force
Copy-Item (Join-Path $root 'packaging/assets/LICENSE.txt') (Join-Path $portable 'LICENSE.txt') -Force
Compress-Archive -Path (Join-Path $portable '*') -DestinationPath $zip -Force

$wix = Get-Command wix -ErrorAction SilentlyContinue
if ($null -eq $wix) { Write-Warning 'WiX v4 (wix) is unavailable; MSI and bootstrapper were not built.'; exit 0 }
$installerLdflags = "-s -w -X wslc-tui-ms/internal/buildinfo.Version=$Tag -X wslc-tui-ms/internal/buildinfo.Commit=$Commit -X wslc-tui-ms/internal/buildinfo.BuildDate=$BuildDate -X wslc-tui-ms/internal/buildinfo.Channel=$Channel -X wslc-tui-ms/internal/buildinfo.Distribution=installer"
& go build -trimpath -ldflags $installerLdflags -o $installerBin $root
if ($LASTEXITCODE -ne 0) { throw 'Installer Go build failed' }
& wix build -arch x64 -d ProductVersion=$version -d ApplicationExecutable=$installerBin -o $msi (Join-Path $root 'packaging/wix/Product.wxs')
if ($LASTEXITCODE -ne 0) { throw 'WiX MSI build failed' }
& wix build -arch x64 -d ProductVersion=$version -d MsiPath=$msi -o $bundle (Join-Path $root 'packaging/bootstrapper/Bundle.wxs')
if ($LASTEXITCODE -ne 0) { throw 'WiX bootstrapper build failed' }
Write-Output "Built $zip, $msi, and $bundle"
