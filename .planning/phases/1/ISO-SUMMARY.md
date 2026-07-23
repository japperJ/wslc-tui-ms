# Phase 1 ISO Workflow Summary

## Files

- `scripts/build-phase1-test-iso.ps1` creates a new tagged staging directory and, when `oscdimg.exe` is available, a new ISO. It refuses to overwrite either output.
- `packaging/tests/RUN-TESTS.ps1` is the VM entry point. It derives the bundle root from `$PSScriptRoot`, infers a single release tag from tagged metadata/checksum files when `-Tag` is omitted, validates prebuilt artifacts under `artifacts\`, and requires only Windows x64, UAC, a non-admin account, and Windows PowerShell 5.1. Child smoke scripts run through `powershell.exe`.
- `packaging/iso/README.md` documents VM prerequisites, admin-only setup, standard-user execution, evidence, and export.
- `packaging/iso/dependencies.json` records that the standard-user smoke run has no developer-tool dependency.
- `docs/releases.md` and `README.md` link the workflow.

## Commands

Create a runnable staging bundle:

```powershell
pwsh -NoProfile -File ./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -OutputDirectory ./artifacts
```

The generated runnable bundle is under `artifacts/phase1-runnable-v1.2.3-checksum-fix-host-fix/` and
contains the prebuilt MSI, bootstrapper, portable ZIP, checksums, and metadata.

Create an ISO with a Windows ADK ISO authoring tool:

```powershell
pwsh -NoProfile -File ./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -OutputDirectory ./artifacts -OscdimgPath 'C:\Program Files (x86)\Windows Kits\10\Assessment and Deployment Kit\Deployment Tools\amd64\Oscdimg\oscdimg.exe'
```

Run inside the mounted/extracted bundle as the fresh standard user:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\RUN-TESTS.ps1
```

The runner accepts `-Tag vX.Y.Z` for explicit selection, or accepts
`.\RUN-TESTS.ps1` without `-Tag` by inferring the single staged tag. A bundle
with no tagged metadata/checksum files or with more than one tag fails before
the Windows/prerequisite checks with a concise remediation message.

## Validation

The runnable tagged build was run on the current Windows host with
`oscdimg.exe` unavailable. The script used the built-in
`IMAPI2FS.MsftFileSystemImage` fallback and produced a mountable ISO with a
manifest-backed bundle and prebuilt packaging artifacts.

- Exact ISO: `C:\REP\wslc-tui-ms\artifacts\phase1-runnable-v1.2.3-checksum-fix-host-fix\wslc-tui-ms-phase1-packaging-smoke-v1.2.3.iso`
- Size: `6,750,208` bytes
- SHA-256: `56311EAC22CC3EC287C6EDABFB3DD1F11FC1559B3A07F155C1D85064F27D38CE`
- `Mount-DiskImage` attached it as `E:\`; Windows reported filesystem `UDF`.
- `E:\README.md` was read successfully (`3,045` bytes).
- `E:\RUN-TESTS.ps1` was read successfully and contains the payload-only checksum validation plus Windows PowerShell child-script selection.
- `E:\bundle-manifest.json` was read successfully (`10,846` bytes).
- `Dismount-DiskImage` completed successfully.
- The existing `wslc-tui-ms.iso` was not modified. The generated ISO is intentionally not committed.

## New Windows PowerShell 5.1 ISO

- Source artifacts: `artifacts/dist-v1.2.3`
- Exact ISO: `C:\REP\wslc-tui-ms\artifacts\phase1-powershell51-v1.2.3-rerun\wslc-tui-ms-phase1-packaging-smoke-v1.2.3.iso`
- Repository-relative path: `artifacts/phase1-powershell51-v1.2.3-rerun/wslc-tui-ms-phase1-packaging-smoke-v1.2.3.iso`
- Size: `6,750,208` bytes
- SHA-256: `da3488022b2c93af06b4564205ab277db1a7b018085397f67e3e62794315922d`
- Authoring: Windows IMAPI2FS fallback; `oscdimg.exe` was unavailable.
- Mounted drive: `E:\`; `E:\RUN-TESTS.ps1` was inspected and then the image was dismounted.
- Mounted runner assertions passed: no `pwsh` string and contains `powershell.exe`.
- The six existing ISO files under `artifacts/` were preserved; the new ISO used a unique output directory.

## Windows PowerShell Runner ISO

- Source artifacts: `artifacts/dist-v1.2.3`
- ISO path: `C:\REP\wslc-tui-ms\artifacts\phase1-powershell51-v1.2.3-rerun\wslc-tui-ms-phase1-packaging-smoke-v1.2.3.iso`
- Repository-relative path: `artifacts/phase1-powershell51-v1.2.3-rerun/wslc-tui-ms-phase1-packaging-smoke-v1.2.3.iso`
- Size: `6,750,208` bytes
- SHA-256: `da3488022b2c93af06b4564205ab277db1a7b018085397f67e3e62794315922d`
- Authoring: Windows IMAPI2FS fallback; `oscdimg.exe` was unavailable.

## Mounted Inspection

- Mounted drive: `E:\`
- Inspected file: `E:\RUN-TESTS.ps1`
- Assertion: passed, mounted runner contains no `pwsh` string.
- Assertion: passed, mounted runner contains `powershell.exe`.
- Dismounted after inspection: yes.

## Preservation

The six existing ISO files under `artifacts/` were preserved. The new ISO was written in the uniquely named `artifacts/phase1-powershell51-v1.2.3-rerun/` directory and did not overwrite an existing image.
