# Phase 2 Summary

Implemented update discovery and the user-controlled TUI update flow.

## Delivered

- Added `internal/update` with an injectable `ReleaseClient`, GitHub Releases client, SemVer filtering, Stable/Beta channel handling, policy minimum-version enforcement, checksum and distribution asset mapping, and a 24-hour automatic-check cooldown.
- Added `internal/settings` with atomic per-user JSON persistence for channel, last check time, and deferred version.
- Added non-blocking startup/manual checks, persistent status/banner messaging, release notes, channel selection, manual check, Later, mandatory update handling, and explicit confirmation before the Phase 3 handoff state.
- Preserved release version, notes, URL, selected asset, distribution mapping, checksum, and size in the update decision.
- Added client, service, and app tests covering filtering, malformed releases, policy, cooldown, deferral, errors, metadata, confirmation, and TUI state transitions.
- Documented update behavior and keyboard controls in `README.md`.

## Ambiguity Decisions

- Settings use `os.UserConfigDir()/wslc-tui-ms/settings.json`, with Stable as the default channel.
- Stable selects the newest non-prerelease release; Beta selects the newest release including prereleases.
- Asset mapping uses Phase 1 names: installer distributions select the MSI, `exe` selects the bootstrapper EXE, and all other distributions select the portable ZIP.
- Phase 2's Download and Install confirmation intentionally creates only an in-memory handoff-ready state. It performs no network asset download and launches no installer; Phase 3 owns that handoff.
- Automatic failures are returned as silent service results and do not interrupt startup. Manual failures remain actionable and are rendered as a TUI status notice.
- Fixed focused command-search handling so `u` starts a manual update check rather than inserting text, and initialize an unconfigured update channel from embedded Beta build metadata while preserving an explicitly persisted Stable/Beta choice.
- Added regression coverage for focused `u` handling, build-channel initialization, persisted channel precedence, and Beta prerelease `v1.2.3` discovery through a fake release client.

## Verification

- `go test ./...`
- `go vet ./...`

Packaging, ISO scripts, `packaging/`, and `wslc-tui-ms.iso` were not changed by this phase.
