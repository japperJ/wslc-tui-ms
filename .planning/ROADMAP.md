# Roadmap

## Phase 1: Release Supply Chain
**Goal:** Produce trustworthy, correctly classified Windows release assets from maintainer-controlled GitHub Releases.
**Requirements:** REQ-001, REQ-002, REQ-003, REQ-004, REQ-005, REQ-015, REQ-016
**Success Criteria:**
1. A tagged GitHub release produces x64 MSI, EXE, and portable ZIP assets with deterministic naming.
2. Draft/prerelease/stable promotion follows the agreed workflow and exposes the correct channel metadata.
3. Every asset has a published SHA-256 checksum, and build metadata identifies its distribution type.
4. MSI installation is per-user and does not require elevation for normal updates.
**Depends on:** None

## Phase 2: Update Discovery And TUI Experience
**Goal:** Let users discover, understand, and opt into updates from the terminal UI without unexpected installation.
**Requirements:** REQ-006, REQ-007, REQ-008, REQ-009, REQ-010
**Success Criteria:**
1. Stable installations ignore prereleases; Beta installations can receive prereleases and stable releases.
2. Automatic checks are throttled to 24 hours, while manual checks are always available.
3. An available release appears with version, channel, release notes, Later, and Download and Install actions.
4. Network failures during automatic checks do not disrupt the command browser.
5. `update-policy.json` can mark older versions unsupported and trigger a mandatory update state.
**Depends on:** Phase 1

## Phase 3: Safe Installation And Recovery
**Goal:** Install verified updates without corrupting the app or user data, then return the user to the updated TUI.
**Requirements:** REQ-011, REQ-012, REQ-013, REQ-014, REQ-017
**Success Criteria:**
1. The updater selects the correct installer or portable asset and rejects checksum mismatches.
2. The running TUI exits cleanly, the helper updates it atomically, and the helper relaunches it with original arguments and working directory.
3. A failed replacement or startup validation restores the previous executable and reports the failure.
4. User data is backed up before migrations and remains recoverable after a failed update.
**Depends on:** Phase 1 and Phase 2
