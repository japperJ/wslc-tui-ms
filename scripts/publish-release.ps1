[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$Tag,
  [string]$OutputDirectory = (Join-Path (Get-Location) "dist\release-$Tag"),
  [switch]$Publish,
  [switch]$Force
)

$ErrorActionPreference = 'Stop'

function Write-Utf8NoBom([string]$Path, [string]$Content) {
  $encoding = New-Object System.Text.UTF8Encoding -ArgumentList $false
  [System.IO.File]::WriteAllText($Path, $Content, $encoding)
}

if ($Tag -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$') {
  throw "Tag must be a SemVer tag such as v1.2.3 or v1.2.3-beta.1: $Tag"
}

$channel = if ($Tag -match '-') { 'Beta' } else { 'Stable' }
$isPrerelease = $channel -eq 'Beta'
$root = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$output = [System.IO.Path]::GetFullPath($OutputDirectory)

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) { throw 'GitHub CLI (gh) is required.' }
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw 'Go is required.' }
if (-not (Get-Command wix -ErrorAction SilentlyContinue)) { throw 'WiX v4 (wix) is required to publish MSI and bootstrapper assets.' }
& gh auth status *> $null
if ($LASTEXITCODE -ne 0) { throw 'GitHub CLI authentication is required. Run gh auth login first.' }

$previousErrorActionPreference = $ErrorActionPreference
try {
  $ErrorActionPreference = 'Continue'
  & gh release view $Tag *> $null
  $releaseLookupExitCode = $LASTEXITCODE
} finally {
  $ErrorActionPreference = $previousErrorActionPreference
}
if ($releaseLookupExitCode -eq 0) { throw "A GitHub release already exists for $Tag." }

if (Test-Path -LiteralPath $output) {
  $existingItems = @(Get-ChildItem -LiteralPath $output -Force)
  if ($existingItems.Count -gt 0 -and -not $Force) {
    throw "Output directory already contains files. Remove them or pass a new -OutputDirectory: $output"
  }
  if ($Force) { Remove-Item -LiteralPath $output -Recurse -Force }
  New-Item -ItemType Directory -Force -Path $output | Out-Null
} else {
  New-Item -ItemType Directory -Force -Path $output | Out-Null
}

Write-Output "Building $Tag ($channel) into $output"
& (Join-Path $root 'packaging/build.ps1') -Tag $Tag -Channel $channel -Output $output -Commit (& git rev-parse HEAD) -BuildDate ([DateTime]::UtcNow.ToString('o'))
if ($LASTEXITCODE -ne 0) { throw 'Packaging build failed.' }

$assetNames = @(
  "wslc-tui-$Tag-windows-amd64.msi",
  "wslc-tui-$Tag-windows-amd64.exe",
  "wslc-tui-$Tag-windows-amd64-portable.zip",
  'update-policy.json'
)
$assets = foreach ($name in $assetNames) {
  $path = Join-Path $output $name
  if (-not (Test-Path -LiteralPath $path -PathType Leaf)) { throw "Missing release asset: $name" }
  $hash = (Get-FileHash $path -Algorithm SHA256).Hash.ToLowerInvariant()
  [ordered]@{ name = $name; sizeBytes = (Get-Item -LiteralPath $path).Length; sha256 = $hash }
}

$checksumPath = Join-Path $output "wslc-tui-$Tag-checksums.json"
$checksumJson = [ordered]@{
  schemaVersion = 1
  releaseTag = $Tag
  algorithm = 'sha256'
  assets = @($assets)
} | ConvertTo-Json -Depth 5
Write-Utf8NoBom $checksumPath $checksumJson

& (Join-Path $root 'packaging/tests/test-package.ps1') -Dist $output -Tag $Tag
if ($LASTEXITCODE -ne 0) { throw 'Package validation failed.' }
& (Join-Path $root 'packaging/tests/test-contracts.ps1') -ChecksumManifest $checksumPath
if ($LASTEXITCODE -ne 0) { throw 'Release contract validation failed.' }

$remoteTagLookup = $null
$previousErrorActionPreference = $ErrorActionPreference
try {
  $ErrorActionPreference = 'Continue'
  & git ls-remote --exit-code --quiet origin "refs/tags/$Tag" *> $null
  $remoteTagLookup = $LASTEXITCODE
} finally {
  $ErrorActionPreference = $previousErrorActionPreference
}
if ($remoteTagLookup -ne 0) {
  $head = (& git rev-parse HEAD).Trim()
  if ($LASTEXITCODE -ne 0) { throw 'Unable to resolve the current Git commit.' }
  & git show-ref --verify --quiet "refs/tags/$Tag" *> $null
  if ($LASTEXITCODE -eq 0) {
    $localTagCommit = (& git rev-list -n 1 $Tag).Trim()
    if ($localTagCommit -ne $head) { throw "Local tag $Tag does not point to the current commit." }
  } else {
    & git tag -a $Tag -m "Release $Tag" $head
    if ($LASTEXITCODE -ne 0) { throw "Unable to create local tag $Tag." }
  }
  Write-Output "Pushing tag $Tag"
  & git push origin "refs/tags/$Tag"
  if ($LASTEXITCODE -ne 0) { throw "Unable to push tag $Tag." }
}

$releaseArgs = @(
  'release', 'create', $Tag,
  '--verify-tag',
  '--title', "$Tag $channel",
  '--generate-notes'
)
if (-not $Publish) { $releaseArgs += '--draft' }
if ($isPrerelease) { $releaseArgs += '--prerelease' }
$releaseArgs += @(
  (Join-Path $output "wslc-tui-$Tag-windows-amd64.msi"),
  (Join-Path $output "wslc-tui-$Tag-windows-amd64.exe"),
  (Join-Path $output "wslc-tui-$Tag-windows-amd64-portable.zip"),
  $checksumPath,
  (Join-Path $output 'update-policy.json')
)

Write-Output "Creating $channel GitHub release for $Tag"
& gh @releaseArgs
if ($LASTEXITCODE -ne 0) { throw 'GitHub release creation failed.' }
