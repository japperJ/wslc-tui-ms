# Windows Releases

## Contracts

Tag releases with `vX.Y.Z` SemVer. The workflow produces these deterministic assets:

- `wslc-tui-vX.Y.Z-windows-amd64.msi`
- `wslc-tui-vX.Y.Z-windows-amd64.exe`
- `wslc-tui-vX.Y.Z-windows-amd64-portable.zip`
- `wslc-tui-vX.Y.Z-checksums.json`
- `update-policy.json`

The checksum manifest uses SHA-256 and covers the MSI, bootstrapper, portable ZIP, and policy file, but intentionally excludes itself. The policy schema is [`packaging/update-policy.schema.json`](../packaging/update-policy.schema.json).

## Local Build

Run on Windows with Go 1.24.5 and WiX Toolset v4 installed:

```powershell
./packaging/build.ps1 -Tag v1.2.3 -Channel Stable -Output ./dist
./packaging/tests/test-package.ps1 -Dist ./dist -Tag v1.2.3
```

WiX v4.0.5 is required for MSI and bootstrapper output. The build script still creates and validates the portable ZIP when `wix` is unavailable, but release verification is `blocked` rather than passed. Use `-AllowDeferred` only for an explicit local deferred check; it records that the matrix was not executed.

## VM Smoke Test

Use a clean Windows 10/11 x64 snapshot with UAC enabled. Build as the VM administrator, then run `packaging/tests/smoke-msi.ps1` as a fresh standard user with a medium-integrity token and writable `%LOCALAPPDATA%`. The test captures an MSI log and verifies the exact per-user path, HKCU values, installer metadata, and absence of HKLM, service, task, or Program Files state. Run the bootstrapper from the same standard account and record its result in the release evidence; its only package payload must be the generated MSI.

The default smoke preflight requires pinned WiX v4.0.5, Windows x64 tools, UAC, and a non-admin standard-user context. Missing prerequisites emit `SMOKE_RESULT=blocked` and fail release verification. The `verify-installer-vm` workflow job runs on the disposable `clean-vm` runner label and publishes MSI/bootstrapper logs and result markers under the `windows-installer-smoke-*` evidence artifact.

No signing certificate or signing secret is needed for the initial unsigned release. `SIGNING_COMMAND` is an optional future boundary in the workflow and is skipped when unset. Unsigned artifacts can trigger a Windows SmartScreen unknown-publisher prompt.

## Channel Promotion

The release workflow creates a draft with the selected prerelease state. GitHub release `prerelease=true` maps to Beta and `false` maps to Stable. Promoting a draft changes only that state: do not rebuild, rename, replace assets, or regenerate checksums.
