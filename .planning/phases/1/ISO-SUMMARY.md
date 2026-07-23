# Phase 1 ISO Workflow Summary

## Files

- `scripts/build-phase1-test-iso.ps1` creates a new tagged staging directory and, when `oscdimg.exe` is available, a new ISO. It refuses to overwrite either output.
- `packaging/tests/RUN-TESTS.ps1` is the VM entry point. It requires Windows x64, UAC, a non-admin account, Go 1.24.5, and WiX 4.0.5 before building and running both smoke tests.
- `packaging/iso/README.md` documents VM prerequisites, admin-only setup, standard-user execution, evidence, and export.
- `packaging/iso/dependencies.json` documents the exact staged Go MSI and WiX v4.0.5 packages. `TO_BE_STAGED` hashes must be filled during trusted dependency acquisition.
- `docs/releases.md` and `README.md` link the workflow.

## Commands

Create a staging bundle:

```powershell
pwsh -NoProfile -File ./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -OutputDirectory ./artifacts
```

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

The staging script was run on the current host and produced a manifest-backed
bundle. The host has PowerShell and Go but does not have `wix` 4.0.5 or
`oscdimg.exe`, so the Windows MSI/bootstrapper matrix and ISO authoring could
not be executed here. The resulting staging path is the reproducible input to
the Windows ADK `oscdimg` command above.

## Limitations

- The repository does not contain redistributable Go/WiX installers. Supply
  them through `-DependencyRoot` and verify hashes before disconnecting the VM.
- Installing Go, .NET, the WiX dotnet tool, and the WiX Bal extension is an
  administrator setup step; the smoke run itself must be standard-user.
- No credentials, tokens, signing certificates, or GitHub access are needed
  or copied. The bundle does not fetch repository content at runtime.
