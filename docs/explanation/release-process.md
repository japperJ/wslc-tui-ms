# Release Process

The release process separates building, validation, and Windows smoke testing.

## Build

The packaging script embeds version, commit, build date, channel, and distribution metadata into the Windows binaries. It produces a portable ZIP and, when WiX is installed, an MSI and bootstrapper.

## Validate

PowerShell contract tests validate the update policy, checksum schema, release payload names, and WiX installation boundaries. The checksum manifest covers the release assets but excludes itself.

## Smoke Test

The smoke matrix runs on a clean Windows x64 environment as a standard user. It verifies that the MSI and bootstrapper install in the intended per-user location and do not create services, scheduled tasks, HKLM state, or Program Files state.

## Publish And Promote

Release assets are associated with a SemVer tag. Stable and Beta are represented by the GitHub release `prerelease` state. Promotion changes that state only; assets and checksums must not be rebuilt during promotion.

## Offline Verification

The Phase 1 bundle packages prebuilt artifacts and smoke scripts for a disposable VM. It can produce an ISO with `oscdimg.exe` or leave a complete staging directory when that tool is unavailable.
