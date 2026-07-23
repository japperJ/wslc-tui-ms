# Phase 1 Summary

## Implementation

Release supply-chain work is implemented across three task commits:

- `57cfbda feat: embed release build metadata`
  - `main.go`
  - `internal/buildinfo/buildinfo.go`
  - `internal/buildinfo/buildinfo_test.go`
- `ccf558a feat: add Windows packaging contracts`
  - `update-policy.json`
  - `packaging/update-policy.schema.json`
  - `packaging/checksums.schema.json`
  - `packaging/tests/fixtures/update-policy.json`
  - `packaging/tests/fixtures/checksums.json`
  - `packaging/wix/Product.wxs`
  - `packaging/bootstrapper/Bundle.wxs`
  - `packaging/assets/README.txt`
  - `packaging/assets/LICENSE.txt`
  - `packaging/build.ps1`
  - `packaging/tests/test-contracts.ps1`
  - `packaging/tests/test-package.ps1`
  - `packaging/tests/smoke-msi.ps1`
  - `docs/releases.md`
  - `README.md`
- `e5cf01e ci: add reproducible Windows release workflow`
  - `.github/workflows/release.yml`
  - `scripts/test-release-workflow.ps1`
  - `packaging/tests/smoke-bootstrapper.ps1`
  - final WiX/build-script adjustments in `packaging/bootstrapper/Bundle.wxs` and `packaging/build.ps1`

The release contract is now explicit: tagged binaries report version, channel, commit, build date, and `installer` or `portable` distribution through `--version`; policy and checksum schemas are version-controlled; assets use deterministic names; the MSI is WiX v4 `perUser`; the bootstrapper carries only the generated MSI; and the workflow creates unsigned draft releases with Beta as the initial channel. Clearing GitHub Release `prerelease` is modeled as a state-only Beta-to-Stable promotion with no rebuild or upload.

## Gap Closure

- `7a202e1 fix: close phase 1 verification gaps`
  - Sets the installed WiX file identity to exactly `wslc-tui.exe`.
  - Adds a locked Ajv 8.17.1 plus ajv-formats 3.0.1 validator for draft 2020-12 policy, fixture, checksum fixture, and generated-manifest validation.
  - Makes missing WiX, x64, UAC, and standard-user prerequisites explicitly `blocked`, with `-AllowDeferred` as the only non-pass local mode.
  - Gates draft release creation on the disposable-VM smoke job and publishes result markers and logs as evidence.

## Tests

- `go test ./...` passed.
- `go vet ./...` passed.
- `pwsh -NoProfile -File packaging/tests/test-contracts.ps1` passed.
- `pwsh -NoProfile -File scripts/test-release-workflow.ps1 -Tag v1.2.3` passed, including two local portable builds, tag/distribution metadata assertions, deterministic artifact-name checks, checksum/promotion fixture checks, and unset-signing-hook behavior.
- `packaging/tests/test-package.ps1` passed for the portable ZIP and exact root contents.
- `git diff --check` passed for the implementation changes.
- `pwsh -NoProfile -File packaging/tests/test-contracts.ps1` passed with real JSON Schema validation and intentional invalid-instance rejection.
- `pwsh -NoProfile -File scripts/test-release-workflow.ps1 -Tag v1.2.3 -AllowDeferred` passed twice through its reproducibility loop and recorded `RELEASE_VERIFICATION=deferred`.

## Blockers

- WiX Toolset v4 is not installed in this environment. The build script creates and validates the portable ZIP, emits an explicit warning, and does not substitute another installer technology. On Windows CI, the workflow installs pinned WiX v4 and builds the MSI/bootstrapper.
- A disposable Windows 10/11 x64 VM with a fresh non-admin account was not available, so `smoke-msi.ps1` and `smoke-bootstrapper.ps1` were not executed. They are included for the required medium-integrity, no-elevation, HKCU-only, MSI-payload, and installed-metadata assertions.

## Remaining Human-only Verification

- Run the pinned WiX v4.0.5 build as an administrator on the disposable Windows 10/11 x64 `clean-vm` runner.
- Run both smoke scripts as the fresh standard user, preserve MSI/bootstrapper logs and result markers, and confirm the `windows-installer-smoke-*` evidence artifact is published.
- Release creation remains gated on that VM job; the current environment is explicitly blocked and is not counted as Phase 1 verification.

Unrelated pre-existing modifications to `.planning/ROADMAP.md`, `.planning/phases/1/PLAN.md`, `.planning/phases/2/PLAN.md`, and `.planning/phases/3/PLAN.md` were preserved.
