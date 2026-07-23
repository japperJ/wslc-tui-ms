# Windows Releases

## Contracts

Tag releases with `vX.Y.Z` SemVer. The workflow produces these deterministic assets:

- `wslc-tui-vX.Y.Z-windows-amd64.msi`
- `wslc-tui-vX.Y.Z-windows-amd64.exe`
- `wslc-tui-vX.Y.Z-windows-amd64-portable.zip`
- `wslc-tui-vX.Y.Z-checksums.json`
- `update-policy.json`

The checksum manifest uses SHA-256 and covers the MSI, bootstrapper, portable ZIP, and policy file, but intentionally excludes itself. The policy schema is [`packaging/update-policy.schema.json`](../packaging/update-policy.schema.json).

The updater helper is embedded in the portable ZIP and per-user MSI installation as `wslc-tui-updater.exe`; it is not a separate release asset.

## Local Build

Build artifacts on a Windows maintainer/build VM with Go 1.24.5 and WiX
Toolset v4 installed:

```powershell
./packaging/build.ps1 -Tag v1.2.3 -Channel Stable -Output ./dist
./packaging/tests/test-package.ps1 -Dist ./dist -Tag v1.2.3
```

WiX v4.0.5 is required for MSI and bootstrapper output. The build script still creates and validates the portable ZIP when `wix` is unavailable, but release verification is `blocked` rather than passed. Use `-AllowDeferred` only for an explicit local deferred check; it records that the matrix was not executed.

## VM Smoke Test

Use a clean Windows 10/11 x64 snapshot with UAC enabled. Build as the VM administrator, then run `packaging/tests/smoke-msi.ps1` as a fresh standard user with a medium-integrity token and writable `%LOCALAPPDATA%`. The test captures an MSI log and verifies the exact per-user path, HKCU values, installer metadata, and absence of HKLM, service, task, or Program Files state. Run the bootstrapper from the same standard account and record its result in the release evidence; its only package payload must be the generated MSI.

The ISO smoke runner requires only Windows x64, UAC, a non-admin standard-user
context, PowerShell, and prebuilt artifacts. It does not require Go, .NET, WiX,
`.git`, repository access, or internet. The separate release build job owns the
pinned Go/WiX prerequisites.

No signing certificate or signing secret is needed for the initial unsigned release. `SIGNING_COMMAND` is an optional future boundary in the workflow and is skipped when unset. Unsigned artifacts can trigger a Windows SmartScreen unknown-publisher prompt.

## Offline Phase 1 VM Bundle

Build a repository-independent test bundle from prebuilt artifacts without
touching the existing ISO:

```powershell
./packaging/build.ps1 -Tag v1.2.3 -Channel Beta -Output ./dist
./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -ArtifactDirectory ./dist -OutputDirectory ./artifacts
```

The script creates `wslc-tui-ms-phase1-packaging-smoke-v1.2.3.iso` when
`oscdimg.exe` is installed, or a complete staging directory when it is not.
Pass `-OscdimgPath` with the Windows ADK `oscdimg.exe` to make the ISO later.
The bundle contains `RUN-TESTS.ps1`, prebuilt payloads, smoke scripts, and
checksums/metadata. Omitting `-ArtifactDirectory` creates a clearly labeled
source-only bundle that fails early with a missing-artifacts message.

## Channel Promotion

The release workflow creates a draft with the selected prerelease state. GitHub release `prerelease=true` maps to Beta and `false` maps to Stable. Promoting a draft changes only that state: do not rebuild, rename, replace assets, or regenerate checksums.
