[CmdletBinding()]
param(
  [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot '../..')),
  [string]$ChecksumManifest
)
$ErrorActionPreference = 'Stop'

$validatorRoot = $PSScriptRoot
$validatorNodeModules = Join-Path $validatorRoot 'node_modules'
if (-not (Get-Command node -ErrorAction SilentlyContinue) -or -not (Get-Command npm.cmd -ErrorAction SilentlyContinue)) { throw 'Pinned JSON Schema validator requires Node.js and npm.' }
if (-not (Test-Path (Join-Path $validatorNodeModules 'ajv')) -or -not (Test-Path (Join-Path $validatorNodeModules 'ajv-formats'))) {
  Push-Location $validatorRoot
  try { & npm.cmd ci --ignore-scripts --no-audit --no-fund } finally { Pop-Location }
  if ($LASTEXITCODE -ne 0) { throw 'Unable to install the pinned JSON Schema validator.' }
}

$schemaPath = Join-Path $Root 'packaging/update-policy.schema.json'
$checksumSchemaPath = Join-Path $Root 'packaging/checksums.schema.json'
$validator = Join-Path $validatorRoot 'validate-schema.mjs'
function Assert-SchemaValid([string]$Schema, [string]$Instance) {
  & node $validator $Schema $Instance
  if ($LASTEXITCODE -ne 0) { throw "JSON Schema validation failed: $Instance" }
}
function Assert-SchemaRejects([string]$Schema, $Instance, [string]$Reason) {
  $path = Join-Path ([System.IO.Path]::GetTempPath()) "wslc-invalid-$PID-$([guid]::NewGuid()).json"
  try {
    $Instance | ConvertTo-Json -Depth 10 | Set-Content $path -Encoding utf8NoBOM
    & node $validator $Schema $path *> $null
    if ($LASTEXITCODE -eq 0) { throw "Schema accepted invalid instance: $Reason" }
  } finally { Remove-Item $path -Force -ErrorAction SilentlyContinue }
}

$policyPath = Join-Path $Root 'update-policy.json'
$fixturePath = Join-Path $Root 'packaging/tests/fixtures/update-policy.json'
$policy = Get-Content $policyPath -Raw | ConvertFrom-Json
$fixture = Get-Content $fixturePath -Raw | ConvertFrom-Json
Assert-SchemaValid $schemaPath $policyPath
Assert-SchemaValid $schemaPath $fixturePath
Assert-SchemaRejects $schemaPath ([ordered]@{ schemaVersion = 1; minimumSupportedVersion = '1.2.3'; unexpected = $true }) 'additional properties'
Assert-SchemaRejects $schemaPath ([ordered]@{ schemaVersion = 1; minimumSupportedVersion = '1.2' }) 'invalid SemVer'
Assert-SchemaRejects $schemaPath ([ordered]@{ schemaVersion = 1; minimumSupportedVersion = '1.2.3'; updatedAt = 'not-a-date' }) 'invalid date-time'
if ($policy.schemaVersion -ne 1 -or $policy.minimumSupportedVersion -notmatch '^[0-9]+\.[0-9]+\.[0-9]+') { throw 'Invalid update policy contract' }
if (($policy | ConvertTo-Json -Compress) -ne ($fixture | ConvertTo-Json -Compress)) { throw 'Policy fixture differs from public policy' }

$checksumsPath = Join-Path $Root 'packaging/tests/fixtures/checksums.json'
$checksums = Get-Content $checksumsPath -Raw | ConvertFrom-Json
Assert-SchemaValid $checksumSchemaPath $checksumsPath
if ($checksums.algorithm -ne 'sha256' -or $checksums.assets.Count -ne 4) { throw 'Checksum fixture coverage is incomplete' }
if (($checksums.assets.name | Sort-Object -Unique).Count -ne $checksums.assets.Count) { throw 'Duplicate checksum asset' }
if (($checksums.assets.name -join '|') -match 'checksums\.json') { throw 'Checksum manifest must not checksum itself' }

if ($ChecksumManifest) {
  if (-not (Test-Path $ChecksumManifest)) { throw "Checksum manifest not found: $ChecksumManifest" }
  Assert-SchemaValid $checksumSchemaPath $ChecksumManifest
  $generated = Get-Content $ChecksumManifest -Raw | ConvertFrom-Json
  $tag = $generated.releaseTag
  $expected = @(
    "wslc-tui-$tag-windows-amd64.msi",
    "wslc-tui-$tag-windows-amd64.exe",
    "wslc-tui-$tag-windows-amd64-portable.zip",
    'update-policy.json'
  ) | Sort-Object
  $actual = @($generated.assets.name) | Sort-Object
  if (($actual -join '|') -ne ($expected -join '|')) { throw 'Generated checksum manifest payload coverage is incomplete' }
  if ($actual -contains (Split-Path $ChecksumManifest -Leaf)) { throw 'Generated checksum manifest must not checksum itself' }
}
Write-Output 'Release contracts valid.'
