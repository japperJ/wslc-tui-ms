[CmdletBinding()]
param([string]$Root = (Resolve-Path (Join-Path $PSScriptRoot '../..')))
$ErrorActionPreference = 'Stop'

$policy = Get-Content (Join-Path $Root 'update-policy.json') -Raw | ConvertFrom-Json
$fixture = Get-Content (Join-Path $Root 'packaging/tests/fixtures/update-policy.json') -Raw | ConvertFrom-Json
if ($policy.schemaVersion -ne 1 -or $policy.minimumSupportedVersion -notmatch '^[0-9]+\.[0-9]+\.[0-9]+') { throw 'Invalid update policy contract' }
if (($policy | ConvertTo-Json -Compress) -ne ($fixture | ConvertTo-Json -Compress)) { throw 'Policy fixture differs from public policy' }

$schema = Get-Content (Join-Path $Root 'packaging/update-policy.schema.json') -Raw | ConvertFrom-Json
if ($schema.properties.schemaVersion.const -ne 1 -or $schema.required -notcontains 'minimumSupportedVersion') { throw 'Policy schema is incomplete' }
$checksums = Get-Content (Join-Path $Root 'packaging/tests/fixtures/checksums.json') -Raw | ConvertFrom-Json
if ($checksums.algorithm -ne 'sha256' -or $checksums.assets.Count -ne 4) { throw 'Checksum fixture coverage is incomplete' }
if ($checksums.assets.name -contains $checksums.assets[0].name -and ($checksums.assets.name | Where-Object { $_ -eq $checksums.assets[0].name }).Count -ne 1) { throw 'Duplicate checksum asset' }
if (($checksums.assets.name -join '|') -match 'checksums\.json') { throw 'Checksum manifest must not checksum itself' }
Write-Output 'Release contracts valid.'
