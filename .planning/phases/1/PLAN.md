---
phase: 1
plan: 1
type: implement
wave: 1
files_modified: [.github/workflows/release.yml, packaging/, internal/buildinfo/]
autonomous: true
must_haves:
  observable_truths:
    - "A SemVer tag builds Windows x64 MSI, EXE, and portable ZIP assets."
    - "The release workflow publishes checksums and distribution metadata."
    - "Draft prereleases can be promoted to Stable without rebuilding the artifact."
  artifacts:
    - path: .github/workflows/release.yml
      has: [tag-triggered-build, asset-upload, checksum-generation]
    - path: internal/buildinfo/
      has: [version, channel, distribution metadata]
    - path: packaging/
      has: [per-user MSI definition, EXE/bootstrapper definition]
  key_links:
    - from: "release tag"
      to: "embedded build metadata"
      verify: "the binary reports the tagged version and distribution type"
    - from: "release assets"
      to: "published checksums"
      verify: "each downloadable asset has a matching checksum"
---

# Phase 1, Plan 1: Release Supply Chain

## Objective
Create the maintainer-controlled Windows release pipeline and packaging metadata needed by the updater. Keep unsigned artifacts for now, but make the build metadata and asset layout compatible with future signing.

## Context
- The project is a Go 1.24.5 Bubble Tea TUI.
- `main.go` currently has no version or command-line metadata.
- No GitHub Actions workflow or installer configuration exists.
- Stable users must not see GitHub prereleases; Beta users must receive them.
- The same GitHub Release is promoted by clearing its prerelease flag.

## Tasks

### Task 1: Add build and distribution metadata
- **files:** internal/buildinfo/, main.go or startup wiring, packaging metadata
- **action:** Add compile-time version, channel, commit/build date, and distribution identifiers with safe development defaults and a human-readable version accessor.
- **verify:** Unit tests cover development defaults and injected release values; a local build prints or exposes the expected version metadata.
- **done:** Every release binary can identify its version and whether it is installer or portable.

### Task 2: Add Windows packaging
- **files:** packaging/ and documentation
- **action:** Define a per-user MSI as the primary installer, an EXE/bootstrapper path, and a portable ZIP layout. Document install paths, PATH behavior, update ownership, and unsigned SmartScreen expectations.
- **verify:** Build each package on Windows and install the MSI per-user without elevation; inspect that the portable ZIP contains the expected executable and helper payload.
- **done:** All three distributions have stable asset names and explicit update metadata.

### Task 3: Add GitHub Actions release workflow
- **files:** .github/workflows/release.yml
- **action:** Build on SemVer tags, produce all x64 assets, generate SHA-256 checksums, and create a draft release that preserves the tag artifact for later prerelease/stable promotion. Add validation for release asset naming and `update-policy.json`.
- **verify:** Run the workflow against a test tag or workflow dispatch; confirm assets, checksums, release notes, and prerelease state.
- **done:** Maintainers can move from draft to Beta to Stable through GitHub without rebuilding.

## Verification
Run `go test ./...`, `go vet ./...`, and a Windows package smoke test. Verify that no credential or signing secret is required for the initial unsigned workflow.

## Success Criteria
- A maintainer can produce and promote a complete x64 release from GitHub.
- Asset identity and checksum information are machine-readable by the future updater.
