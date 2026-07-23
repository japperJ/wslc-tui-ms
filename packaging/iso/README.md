# Phase 1 Windows Packaging Smoke ISO

This media is a repository-independent test bundle for the Phase 1 Windows x64
packaging smoke matrix. It contains prebuilt artifacts under `artifacts\`, the
VM runner at `RUN-TESTS.ps1`, and a writable `results\` directory. The smoke
run does not use `.git`, repository access, Go, .NET, or WiX.

## VM Prerequisites

- Disposable Windows 10 or Windows 11 x64 VM with a clean snapshot.
- UAC enabled and a local administrator account for setup only.
- At least 4 GB RAM, 10 GB free disk, and PowerShell 5.1 or PowerShell 7.
- Internet access is not required: all release payloads and verification files
  are staged before the VM is disconnected.

`-DependencyRoot` is optional developer setup input and is not read by the
standard-user runner.

## Setup And Run

1. Mount or extract the bundle without copying it into the application install
   directory.
2. Create a fresh local standard user with UAC enabled. Sign in as that user;
   do not use `Run as administrator` for the smoke run.
4. Open PowerShell in the bundle root and run:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\RUN-TESTS.ps1 -Tag v1.2.3
```

The script validates artifact names, checksums, metadata, and portable layout,
then launches the portable, MSI, and bootstrapper smoke tests as the current
standard user. Missing artifacts fail early with one concise message.

## Expected Evidence

`results\` contains the generated artifacts and:

- `portable-result.txt`, `msi-result.txt`, and `bootstrapper-result.txt` with command output.
- `wslc-tui-msi-smoke.log` and `wslc-tui-bootstrapper-smoke.log`.
- `result-marker.txt` containing `SMOKE_RESULT=passed`.
- `bundle-manifest.json` containing SHA-256 hashes for the staged inputs.

The MSI check proves the LocalAppData install path, exact HKCU values,
installer metadata, medium-integrity/non-admin execution, and absence of
HKLM/service/task/Program Files state. The bootstrapper check proves the
installed binary is the installer distribution and that its log identifies one
MSI payload.

## Export Results

Copy `results\` to a host-only folder or attach it to the test record:

```powershell
$out = Join-Path $env:USERPROFILE 'Desktop\wslc-tui-phase1-results'
Copy-Item .\results $out -Recurse -Force
Compress-Archive -Path .\results\* -DestinationPath "$out.zip" -Force
```

Do not include user profiles, registry hives, credentials, or unrelated VM
files in the exported evidence.

## Build Artifacts Separately

Maintainers build the release artifacts on a developer/build VM, then pass that
output directory to the bundle generator. The standard-user runner never
installs or resolves developer tools:

```powershell
./packaging/build.ps1 -Tag v1.2.3 -Channel Beta -Output ./dist
./scripts/build-phase1-test-iso.ps1 -Tag v1.2.3 -ArtifactDirectory ./dist -OutputDirectory ./artifacts
```
