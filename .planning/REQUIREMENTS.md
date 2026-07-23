# Requirements

| ID | Requirement | Phase | Priority |
|---|---|---|---|
| REQ-001 | Build Windows x64 `.exe`, `.msi`, and portable `.zip` assets in GitHub Actions | Phase 1 | Must-have |
| REQ-002 | Use SemVer tags and GitHub Release prerelease state for Stable/Beta channels | Phase 1 | Must-have |
| REQ-003 | Publish SHA-256 checksums for every release asset | Phase 1 | Must-have |
| REQ-004 | Provide a public `update-policy.json` minimum-supported-version control | Phase 1 | Must-have |
| REQ-005 | Detect installed versus portable distribution through explicit build metadata | Phase 1 | Must-have |
| REQ-006 | Check GitHub Releases once per 24 hours and on manual request | Phase 2 | Must-have |
| REQ-007 | Support per-installation Stable/Beta channel selection | Phase 2 | Must-have |
| REQ-008 | Parse and display GitHub release notes in the TUI | Phase 2 | Must-have |
| REQ-009 | Download only after user confirmation and allow Later deferral | Phase 2 | Must-have |
| REQ-010 | Fail quietly for automatic checks and show actionable errors for manual checks | Phase 2 | Must-have |
| REQ-011 | Download and verify the matching asset using SHA-256 | Phase 3 | Must-have |
| REQ-012 | Update through an external helper that can replace the running executable safely | Phase 3 | Must-have |
| REQ-013 | Back up, atomically replace, validate startup, and roll back on failure | Phase 3 | Must-have |
| REQ-014 | Relaunch the TUI with original arguments and working directory | Phase 3 | Must-have |
| REQ-015 | Support per-user MSI installation and an EXE/bootstrapper path | Phase 1 | Must-have |
| REQ-016 | Accept unsigned releases initially while keeping signing integration possible later | Phase 1 | Must-have |
| REQ-017 | Preserve user configuration/data with a backup and versioned migration boundary | Phase 3 | Must-have |
