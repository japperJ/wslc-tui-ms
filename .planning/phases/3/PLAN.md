---
phase: 3
plan: 1
type: implement
wave: 1
files_modified: [cmd/updater/, internal/update/, internal/app/, internal/platform/, packaging/, internal/data/]
autonomous: true
must_haves:
  observable_truths:
    - "The correct installer or portable asset is downloaded only after confirmation."
    - "Checksum mismatch prevents installation."
    - "The running app exits, updates atomically, validates startup, and relaunches."
    - "Failed updates restore the previous executable and preserve user data."
  artifacts:
    - path: cmd/updater/
      has: [staging, checksum verification, backup, atomic replacement, rollback, relaunch]
    - path: internal/update/
      has: [updater handoff, distribution-aware asset selection]
    - path: internal/platform/
      has: [Windows path and process helpers]
  key_links:
    - from: "TUI Download and Install action"
      to: "external updater helper"
      verify: "helper receives executable path, asset URL, checksum, args, cwd, and relaunch target"
    - from: "helper startup validation"
      to: "rollback path"
      verify: "failed validation restores the prior executable"
---

# Phase 3, Plan 1: Safe Installation And Recovery

## Objective
Implement the external Windows updater and wire it to the TUI so installed and portable distributions update safely, preserve invocation context, and recover from failed replacement or migration.

## Context
- The TUI cannot replace its own running executable reliably.
- MSI is the primary per-user installer; EXE is a bootstrapper path; portable uses ZIP.
- Users accepted unsigned SmartScreen warnings for now.
- Updates must preserve command-line arguments and working directory.
- User data must be backed up before versioned migrations.

## Tasks

### Task 1: Build the external updater helper
- **files:** cmd/updater/, internal/platform/, packaging/
- **action:** Add a Windows x64 helper with a validated handoff contract, staged download, size/error checks, SHA-256 verification, backup, atomic replacement, installer invocation for installed builds, portable ZIP replacement, startup validation, rollback, cleanup, and relaunch.
- **verify:** Unit-test path/argument validation and checksum failures; integration-test successful update, interrupted download, locked executable, failed startup validation, rollback, and relaunch in a temporary directory.
- **done:** The helper never installs an unverified asset and can recover to the prior version.

### Task 2: Wire TUI handoff and data protection
- **files:** internal/update/, internal/app/, internal/platform/, internal/data/, related tests
- **action:** Launch the helper only after user confirmation, pass original args/cwd and distribution metadata, exit the TUI cleanly, back up user data/config before migrations, and expose success/failure status after relaunch or on the next startup.
- **verify:** End-to-end Windows smoke test covers installed and portable flows, argument/cwd preservation, data backup, migration failure, and normal relaunch.
- **done:** Users return to the updated TUI without losing context or data.

## Verification
Run `go test ./...`, `go vet ./...`, build the updater with Windows target settings, and execute a disposable-directory update matrix for MSI, EXE, and portable ZIP. Do not test replacement against the real development executable.

## Success Criteria
- Verified updates install through a helper, not the running TUI.
- Failure paths are recoverable and observable.
