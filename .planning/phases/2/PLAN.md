---
phase: 2
plan: 1
type: implement
wave: 1
files_modified: [internal/update/, internal/settings/, internal/app/app.go, internal/ui/, README.md]
autonomous: true
must_haves:
  observable_truths:
    - "Stable and Beta channels resolve different GitHub release sets."
    - "The TUI presents release notes and explicit update actions."
    - "Automatic checks are throttled and non-blocking."
    - "A minimum-supported-version policy can force an update state."
  artifacts:
    - path: internal/update/
      has: [release client, SemVer filtering, policy parsing, check scheduler]
    - path: internal/settings/
      has: [per-installation channel, last-check timestamp, deferred version]
    - path: internal/app/app.go
      has: [update state, banner/view, manual check action]
  key_links:
    - from: "startup/manual command"
      to: "non-blocking release check"
      verify: "TUI remains usable while GitHub is unavailable or slow"
    - from: "release result"
      to: "TUI update view"
      verify: "version, channel, notes, and actions are rendered"
---

# Phase 2, Plan 1: Update Discovery And TUI Experience

## Objective
Add a testable update service and integrate it into the Bubble Tea model without blocking command discovery or silently installing anything.

## Context
- The current model has command, learn, preview, form, confirmation, output, and splash views.
- There is no persistent settings layer.
- Release data comes from public GitHub Releases and `update-policy.json`.
- Automatic checks occur once per 24 hours; manual checks bypass the cooldown.

## Tasks

### Task 1: Implement release discovery and local settings
- **files:** internal/update/, internal/settings/
- **action:** Add GitHub release fetching, SemVer parsing, channel filtering, checksum/asset metadata parsing, policy loading, 24-hour throttling, per-installation channel persistence, and deferred-version persistence. Automatic checks must return silent background errors while manual checks retain actionable errors.
- **verify:** Unit tests cover stable/Beta filtering, prerelease promotion, malformed releases, policy minimum version, cooldown, deferral, and network failure behavior.
- **done:** The update service returns a deterministic update decision without touching TUI state directly.

### Task 2: Add update state and TUI flows
- **files:** internal/app/app.go, internal/ui/, related app tests, README.md
- **action:** Add a non-blocking update check command, persistent banner/status notice, release-notes view, Stable/Beta selector, Check for Updates action, Later action, mandatory update state, and Download and Install handoff state. Preserve existing keyboard and terminal layout conventions.
- **verify:** Model tests assert update messages transition correctly; manual terminal smoke test confirms the command browser remains usable during slow/failing checks and that release notes are readable.
- **done:** Users can inspect notes, change channel per installation, defer an update, or explicitly begin download.

## Verification
Run `go test ./...`, `go vet ./...`, and a manual TUI test with a fake GitHub client. Confirm no automatic check starts an installer and no network failure breaks startup.

## Success Criteria
- Stable/Beta behavior matches GitHub prerelease state.
- The update prompt is discoverable but user-controlled.
