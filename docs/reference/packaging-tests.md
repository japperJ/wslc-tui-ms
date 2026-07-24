# Packaging And Test Reference

## Build A Release

On a Windows build machine with Go 1.24.5 and WiX v4 installed:

```powershell
./packaging/build.ps1 -Tag v1.2.3 -Channel Stable -Output ./dist
```

The tag must match `vX.Y.Z` with an optional prerelease suffix. The script builds the portable package first. If WiX is available, it also builds the MSI and bootstrapper.

## Validate Package Contracts

```powershell
./packaging/tests/test-package.ps1 -Dist ./dist -Tag v1.2.3
./packaging/tests/test-contracts.ps1
```

Contract validation checks JSON schemas, policy fixtures, checksum coverage, and the per-user WiX structure. The contract test installs pinned Node dependencies under `packaging/tests/node_modules` when needed; that directory is generated and must not be committed.

## Run Go Tests

```powershell
```

## Build The Offline Phase 1 Bundle

```powershell
./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -ArtifactDirectory ./dist -OutputDirectory ./artifacts
```

The command creates an ISO when `oscdimg.exe` is available. Otherwise it creates a complete staging directory. It does not overwrite the repository's existing ISO.

## VM Smoke Tests

Use a clean Windows 10/11 x64 snapshot with UAC enabled. Build artifacts as an administrator, then run the MSI and bootstrapper smoke tests as a fresh standard user. The tests verify per-user installation state, installer metadata, and the absence of machine-wide state.
