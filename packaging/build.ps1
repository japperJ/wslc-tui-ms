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
$msiVersion = ($version -split '-')[0]
$versionParts = $msiVersion.Split('.')
$versionMajor = [int]$versionParts[0]
$versionMinor = [int]$versionParts[1]
$versionPatch = [int]$versionParts[2]
$root = Resolve-Path (Join-Path $PSScriptRoot '..')
$bin = Join-Path $Output 'wslc-tui.exe'
$updaterBin = Join-Path $Output 'wslc-tui-updater.exe'
$installerBin = Join-Path $Output 'wslc-tui-installer.exe'
$portable = Join-Path $Output "wslc-tui-$Tag-windows-amd64-portable"
$msi = Join-Path $Output "wslc-tui-$Tag-windows-amd64.msi"
$bundle = Join-Path $Output "wslc-tui-$Tag-windows-amd64.exe"
$zip = Join-Path $Output "wslc-tui-$Tag-windows-amd64-portable.zip"
New-Item -ItemType Directory -Force -Path $Output, $portable | Out-Null
Copy-Item (Join-Path $root 'update-policy.json') (Join-Path $Output 'update-policy.json') -Force

function New-VersionResource([string]$Destination, [string]$OriginalName, [string]$InternalName) {
  $template = Get-Content (Join-Path $root 'packaging/versioninfo.json') -Raw | ConvertFrom-Json
  $template.FixedFileInfo.FileVersion.Major = $versionMajor
  $template.FixedFileInfo.FileVersion.Minor = $versionMinor
  $template.FixedFileInfo.FileVersion.Patch = $versionPatch
  $template.FixedFileInfo.ProductVersion.Major = $versionMajor
  $template.FixedFileInfo.ProductVersion.Minor = $versionMinor
  $template.FixedFileInfo.ProductVersion.Patch = $versionPatch
  $template.StringFileInfo.FileVersion = $version
  $template.StringFileInfo.ProductVersion = $version
  $template.StringFileInfo.OriginalFilename = $OriginalName
  $template.StringFileInfo.InternalName = $InternalName
  $jsonPath = "$Destination.json"
  $destinationDir = Split-Path $Destination -Parent
  $destinationName = Split-Path $Destination -Leaf
  try {
    $template | ConvertTo-Json -Depth 10 | Set-Content $jsonPath -Encoding utf8NoBOM
    Push-Location $destinationDir
    try { & go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.7.0 $jsonPath -64 -o $destinationName -propagate-ver-strings }
    finally { Pop-Location }
    if ($LASTEXITCODE -ne 0) { throw "Windows version resource generation failed for $OriginalName" }
  }
  finally { Remove-Item $jsonPath -Force -ErrorAction SilentlyContinue }
}

$ldflags = "-s -w -X wslc-tui-ms/internal/buildinfo.Version=$Tag -X wslc-tui-ms/internal/buildinfo.Commit=$Commit -X wslc-tui-ms/internal/buildinfo.BuildDate=$BuildDate -X wslc-tui-ms/internal/buildinfo.Channel=$Channel -X wslc-tui-ms/internal/buildinfo.Distribution="
$env:GOOS = 'windows'; $env:GOARCH = 'amd64'
 $rootResource = Join-Path $root 'resource.syso'
New-VersionResource $rootResource 'wslc-tui.exe' 'wslc-tui'
try { & go build -trimpath -ldflags "$ldflags`portable" -o $bin $root }
finally { Remove-Item $rootResource -Force -ErrorAction SilentlyContinue }
if ($LASTEXITCODE -ne 0) { throw 'Portable Go build failed' }
Copy-Item $bin (Join-Path $portable 'wslc-tui.exe') -Force
$updaterLdflags = "-s -w -X wslc-tui-ms/internal/buildinfo.Version=$Tag -X wslc-tui-ms/internal/buildinfo.Commit=$Commit -X wslc-tui-ms/internal/buildinfo.BuildDate=$BuildDate -X wslc-tui-ms/internal/buildinfo.Channel=$Channel -X wslc-tui-ms/internal/buildinfo.Distribution=updater"
$updaterRoot = Join-Path $root 'cmd/updater'
$updaterResource = Join-Path $updaterRoot 'resource.syso'
New-VersionResource $updaterResource 'wslc-tui-updater.exe' 'wslc-tui-updater'
try { & go build -trimpath -ldflags $updaterLdflags -o $updaterBin $updaterRoot }
finally { Remove-Item $updaterResource -Force -ErrorAction SilentlyContinue }
if ($LASTEXITCODE -ne 0) { throw 'Updater helper build failed' }
Copy-Item $updaterBin (Join-Path $portable 'wslc-tui-updater.exe') -Force
Copy-Item (Join-Path $root 'packaging/assets/README.txt') (Join-Path $portable 'README.txt') -Force
Copy-Item (Join-Path $root 'packaging/assets/LICENSE.txt') (Join-Path $portable 'LICENSE.txt') -Force
Compress-Archive -Path (Join-Path $portable '*') -DestinationPath $zip -Force

$wix = Get-Command wix -ErrorAction SilentlyContinue
if ($null -eq $wix) { Write-Warning 'WiX v4 (wix) is unavailable; MSI and bootstrapper were not built.'; exit 0 }
$installerLdflags = "-s -w -X wslc-tui-ms/internal/buildinfo.Version=$Tag -X wslc-tui-ms/internal/buildinfo.Commit=$Commit -X wslc-tui-ms/internal/buildinfo.BuildDate=$BuildDate -X wslc-tui-ms/internal/buildinfo.Channel=$Channel -X wslc-tui-ms/internal/buildinfo.Distribution=installer"
New-VersionResource $rootResource 'wslc-tui.exe' 'wslc-tui'
try { & go build -trimpath -ldflags $installerLdflags -o $installerBin $root }
finally { Remove-Item $rootResource -Force -ErrorAction SilentlyContinue }
if ($LASTEXITCODE -ne 0) { throw 'Installer Go build failed' }
& wix build -arch x64 -ext WixToolset.Util.wixext -d ProductVersion=$msiVersion -d ApplicationExecutable=$installerBin -d UpdaterExecutable=$updaterBin -o $msi (Join-Path $root 'packaging/wix/Product.wxs')
if ($LASTEXITCODE -ne 0) { throw 'WiX MSI build failed' }
& wix build -arch x64 -ext WixToolset.Bal.wixext -d ProductVersion=$msiVersion -d MsiPath=$msi -o $bundle (Join-Path $root 'packaging/bootstrapper/Bundle.wxs')
if ($LASTEXITCODE -ne 0) { throw 'WiX bootstrapper build failed' }
Write-Output "Built $zip, $msi, and $bundle"
