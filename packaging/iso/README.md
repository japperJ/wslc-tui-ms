# Phase 1 Windows Packaging Smoke ISO

This media is a repository-independent test bundle for the Phase 1 Windows x64
packaging smoke matrix. It contains the source and packaging inputs under
`repository\`, the VM runner at `RUN-TESTS.ps1`, and a writable `results\`
working-tree `.git` directory.

## VM Prerequisites

- Disposable Windows 10 or Windows 11 x64 VM with a clean snapshot.
- UAC enabled and a local administrator account for setup only.
- At least 4 GB RAM, 10 GB free disk, and PowerShell 5.1 or PowerShell 7.
- Internet access is not required when the files listed in `tools\` are staged.
- The exact dependency staging contract is in `tools\dependencies.json`.

The bundle generator copies any files supplied with `-DependencyRoot` into
`tools\`. If the directory is empty, stage the pinned installers/packages before
the VM is disconnected. `sha256` values marked `TO_BE_STAGED` must be replaced
with hashes from the trusted download process; the runner does not silently
trust unverified files.

## Setup And Run

1. Mount or extract the bundle without copying it into the application install
   directory.
2. As the VM administrator, install Go 1.24.5 from the staged MSI, or install
   it from the documented official source. Install the WiX v4.0.5 dotnet tool
   and `WixToolset.Bal.wixext` from the staged NuGet packages or an approved
   offline package source. Do not use a different WiX version.
3. Create a fresh local standard user with UAC enabled. Sign in as that user;
   do not use `Run as administrator` for the smoke run.
4. Open PowerShell in the bundle root and run:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\RUN-TESTS.ps1 -Tag v1.2.3
```

The script builds the portable, MSI, and bootstrapper artifacts from
`repository\`, validates the portable layout, then launches the MSI and
bootstrapper smoke tests as the current standard user. It stops on missing or
wrong-version tools and records `SMOKE_RESULT=passed` only after both tests
complete.

## Expected Evidence

`results\` contains the generated artifacts and:

- `msi-result.txt` and `bootstrapper-result.txt` with command output.
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
