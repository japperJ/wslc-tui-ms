[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'

$root = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$builder = Join-Path $root 'scripts/build-phase1-test-iso.ps1'
$runner = Get-Content (Join-Path $root 'packaging/tests/RUN-TESTS.ps1') -Raw
$windowsPowerShell = Join-Path $env:windir 'System32\WindowsPowerShell\v1.0\powershell.exe'
function Invoke-BundledRunner([string]$Path, [string[]]$Arguments) {
  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try { @(& $windowsPowerShell -NoProfile -File $Path @Arguments 2>&1 | ForEach-Object { $_.ToString() }) }
  finally { $ErrorActionPreference = $previousErrorActionPreference }
}

if ($runner -notmatch '\$PSScriptRoot') { throw 'ISO runner must derive its bundle root from $PSScriptRoot.' }
if ($runner -match '\[Parameter\(Mandatory\s*=\s*\$true\)\]\[string\]\$Tag') { throw 'ISO runner tag must be optional.' }
if ($runner -match 'Get-Command\s+go|Get-Command\s+wix|packaging\\build\.ps1') {
  throw 'ISO runner must not require developer packaging tools or build from source.'
}
if ($runner -notmatch '\$payloadAssets\s*=') { throw 'ISO runner must define the canonical payload asset list.' }
if ($runner -notmatch 'foreach \(\$name in \$payloadAssets\)') {
  throw 'ISO runner checksum validation must cover payload assets only.'
}
if ($runner -match 'foreach \(\$name in \$required') {
  throw 'ISO runner must not checksum metadata files.'
}
if ($runner -notmatch '\$metadataAssets\s*=') { throw 'ISO runner must define the metadata asset list.' }
if ($runner -notmatch '\$metadataAssets\s*\|\s*Sort-Object') {
  throw 'ISO runner metadata validation must use the canonical metadata asset list.'
}
if ($runner -match '&\s*pwsh\b') {
  throw 'ISO runner must not invoke the PowerShell 7 pwsh executable.'
}
if ($runner -match '(?i)pwsh\.exe') {
  throw 'ISO runner must not contain a pwsh.exe executable path.'
}
if ($runner -notmatch '\$hostExecutable\s*=') {
  throw 'ISO runner must select the current PowerShell host executable.'
}
if ($runner -notmatch '\$env:windir') {
  throw 'ISO runner host selection must use the Windows installation path.'
}
if ($runner -notmatch 'powershell\.exe') {
  throw 'ISO runner must support the Windows PowerShell executable.'
}

$temp = Join-Path ([IO.Path]::GetTempPath()) "wslc-iso-test-$PID"
Remove-Item $temp -Recurse -Force -ErrorAction SilentlyContinue
try {
  & pwsh -NoProfile -File $builder -Tag v1.2.3 -OutputDirectory $temp -SkipIso
  if ($LASTEXITCODE -ne 0) { throw 'Source-only ISO bundle creation failed.' }
  $bundle = Join-Path $temp 'wslc-tui-ms-phase1-packaging-smoke-v1.2.3'
  Push-Location ([IO.Path]::GetTempPath())
  try { $output = Invoke-BundledRunner (Join-Path $bundle 'RUN-TESTS.ps1') @('-Tag', 'v1.2.3') } finally { Pop-Location }
  if (($output -join "`n") -notmatch 'Missing packaged artifacts') {
    throw "Missing-artifact smoke failure was not concise: $($output -join ' | ')"
  }
  if (($output -join "`n") -match '(?im)(?:^|[\s:])(?:Go|WiX|wix)(?:[\s:]|$)') {
    throw "Missing-artifact failure resolved developer prerequisites: $($output -join ' | ')"
  }

  Push-Location ([IO.Path]::GetTempPath())
  try { $output = Invoke-BundledRunner (Join-Path $bundle 'RUN-TESTS.ps1') @() } finally { Pop-Location }
  if (($output -join "`n") -notmatch 'Could not infer a single release tag') {
    throw "Zero-tag failure was not actionable: $($output -join ' | ')"
  }

  Set-Content (Join-Path $bundle 'artifacts\wslc-tui-v1.2.3-checksums.json') '{"releaseTag":"v1.2.3"}'
  Push-Location ([IO.Path]::GetTempPath())
  try { $output = Invoke-BundledRunner (Join-Path $bundle 'RUN-TESTS.ps1') @() } finally { Pop-Location }
  if (($output -join "`n") -match 'Could not infer a single release tag') {
    throw "Single-tag inference failed: $($output -join ' | ')"
  }

  Set-Content (Join-Path $bundle 'artifacts\wslc-tui-v2.0.0-metadata.json') '{"releaseTag":"v2.0.0"}'
  Push-Location ([IO.Path]::GetTempPath())
  try { $output = Invoke-BundledRunner (Join-Path $bundle 'RUN-TESTS.ps1') @() } finally { Pop-Location }
  if (($output -join "`n") -notmatch 'Multiple packaged release tags found') {
    throw "Multiple-tag failure was not actionable: $($output -join ' | ')"
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
