# Phase 1 ISO Workflow Summary

## Files

- `scripts/build-phase1-test-iso.ps1` creates a new tagged staging directory and, when `oscdimg.exe` is available, a new ISO. It refuses to overwrite either output.
- `packaging/tests/RUN-TESTS.ps1` is the VM entry point. It derives the bundle root from `$PSScriptRoot`, validates prebuilt artifacts under `artifacts\`, and requires only Windows x64, UAC, a non-admin account, and PowerShell.
- `packaging/iso/README.md` documents VM prerequisites, admin-only setup, standard-user execution, evidence, and export.
- `packaging/iso/dependencies.json` records that the standard-user smoke run has no developer-tool dependency.
- `docs/releases.md` and `README.md` link the workflow.

## Commands

Create a staging bundle:

```powershell
pwsh -NoProfile -File ./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -OutputDirectory ./artifacts
```

For a runnable bundle, build artifacts separately and pass them with
`-ArtifactDirectory ./dist`. Without that option the output is source-only and
the runner fails early with `Missing packaged artifacts`.

Create an ISO with a Windows ADK ISO authoring tool:

```powershell
pwsh -NoProfile -File ./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -OutputDirectory ./artifacts -OscdimgPath 'C:\Program Files (x86)\Windows Kits\10\Assessment and Deployment Kit\Deployment Tools\amd64\Oscdimg\oscdimg.exe'
```

Run inside the mounted/extracted bundle as the fresh standard user:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\RUN-TESTS.ps1 -Tag v1.2.3
```

## Validation

The tagged build was run on the current Windows host with `oscdimg.exe`
unavailable. The script used the built-in
`IMAPI2FS.MsftFileSystemImage` fallback and produced a mountable ISO with a
manifest-backed bundle.

- Exact ISO: `C:\REP\wslc-tui-ms\artifacts\phase1-v1.2.3-imapi2fs-final2\wslc-tui-ms-phase1-packaging-smoke-v1.2.3.iso`
- Size: `1,835,008` bytes
- `Mount-DiskImage` attached it as `E:\`; Windows reported filesystem `UDF`.
- `E:\README.md` was read successfully (`3,045` bytes).
- `E:\bundle-manifest.json` was read successfully (`10,846` bytes).
- `Dismount-DiskImage` completed successfully.
- The existing `wslc-tui-ms.iso` was not modified. The generated ISO is intentionally not committed.

## Limitations

- The repository does not contain redistributable Go/WiX installers. Supply
  them through `-DependencyRoot` and verify hashes before disconnecting the VM.
- Go, .NET, and WiX are administrator/developer build-VM prerequisites only;
  the final standard-user setup test does not install or resolve them.
- No credentials, tokens, signing certificates, or GitHub access are needed
  or copied. The bundle does not fetch repository content at runtime.
