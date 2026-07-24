# Windows Releases

## Contracts

Tag releases with `vX.Y.Z` SemVer. The workflow produces these deterministic assets:

- `wslc-tui-vX.Y.Z-windows-amd64.msi`
- `wslc-tui-vX.Y.Z-windows-amd64.exe`
- `wslc-tui-vX.Y.Z-windows-amd64-portable.zip`
- `wslc-tui-vX.Y.Z-checksums.json`
- `update-policy.json`

The checksum manifest uses SHA-256 and covers the MSI, bootstrapper, portable ZIP, and policy file, but intentionally excludes itself. The policy schema is [`packaging/update-policy.schema.json`](../packaging/update-policy.schema.json).

The release workflow derives the channel from the tag:

- `v1.2.3` creates a Stable release.
- `v1.2.3-beta.1` creates a Beta release.

Any tag with a prerelease suffix is treated as Beta. The same channel is embedded in the binaries and applied to the GitHub release's `prerelease` flag.

The updater helper is embedded in the portable ZIP and per-user MSI installation as `wslc-tui-updater.exe`; it is not a separate release asset.

## Local Build

Build artifacts on a Windows maintainer/build VM with Go 1.24.5 and WiX
Toolset v4 installed:

```powershell
./packaging/build.ps1 -Tag v1.2.3 -Channel Stable -Output ./dist
./packaging/tests/test-package.ps1 -Dist ./dist -Tag v1.2.3
```

WiX v4.0.5 is required for MSI and bootstrapper output. The build script still creates and validates the portable ZIP when `wix` is unavailable, but release verification is `blocked` rather than passed. Use `-AllowDeferred` only for an explicit local deferred check; it records that the matrix was not executed.

## Manual Publishing

Release publishing is maintainer-controlled and does not wait for a self-hosted VM. Use [`scripts/publish-release.ps1`](../scripts/publish-release.ps1), which builds the tag, validates package contracts, generates checksums, and creates a draft release. See [`docs/how-to/publish-release.md`](how-to/publish-release.md) for the full procedure.

The repository still contains optional MSI, bootstrapper, portable, and offline ISO smoke-test scripts for maintainers who want additional local verification. They are not a release gate.

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

The workflow creates a draft whose channel matches the tag. GitHub release `prerelease=true` maps to Beta and `false` maps to Stable. If a draft is promoted without changing its channel, do not rebuild, rename, replace assets, or regenerate checksums.
