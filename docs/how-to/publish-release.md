# Publish A Release Manually

Use `scripts/publish-release.ps1` to build, validate, and publish a Windows release from a maintainer workstation. The script is draft-first and does not require the repository's VM smoke-test runner.

## Prerequisites

Install and authenticate:

- Go 1.24.5.
- WiX Toolset v4.0.5.
- GitHub CLI (`gh`) authenticated with permission to create releases.
- Node.js and npm for the pinned JSON Schema contract validator.

Run the script from the repository root in Windows PowerShell 5.1 or PowerShell 7.

PowerShell requires `./` or `.\` before a script path. From the repository root, use:

```powershell
.\scripts\publish-release.ps1 -Tag v1.2.15-beta.1
```

## Create A Beta Draft

Prerelease tags create Beta builds automatically:

```powershell
.\scripts\publish-release.ps1 -Tag v1.2.15-beta.1
```

The script builds the binaries, generates the SHA-256 checksum manifest, validates the package contracts, and creates a draft GitHub release marked as a prerelease.

## Create A Stable Draft

Plain SemVer tags create Stable builds:

```powershell
.\scripts\publish-release.ps1 -Tag v1.2.15
```

The script creates a draft release without the prerelease flag.

## Publish Immediately

Drafts are recommended so the maintainer can inspect assets and release notes first. To publish immediately, add `-Publish`:

```powershell
.\scripts\publish-release.ps1 -Tag v1.2.15 -Publish
```

## Output Directory

By default, artifacts are written to `dist\release-<tag>`. Use a separate directory when repeating a build:

```powershell
.\scripts\publish-release.ps1 -Tag v1.2.15 -OutputDirectory .\dist\release-v1.2.15
```

The script refuses to reuse an existing output directory or overwrite an existing GitHub release.

If a previous build failed and left partial files, retry with `-Force` to remove that output directory before rebuilding:

```powershell
.\scripts\publish-release.ps1 -Tag v1.2.16-beta.1 -Publish -Force
```

## Release Checklist

1. Confirm the current commit is the intended release commit.
2. Run the script without `-Publish`.
3. Inspect the generated release assets and notes on GitHub.
4. Confirm the release channel matches the tag.
5. Publish the draft from GitHub.

If the tag does not exist remotely, the script creates and pushes an annotated tag at the current commit after build and validation succeed.

The workflow in `.github/workflows/release.yml` builds and uploads CI artifacts only. It does not publish releases or wait for a VM smoke-test runner.
