# Troubleshoot Installation

## MSI Does Not Install

1. Confirm that the downloaded MSI hash matches the release checksum manifest.
2. Confirm that the Windows account can write to `%LOCALAPPDATA%`.
3. Close another running copy of WSLC TUI.
4. Retry the installation as the intended user.
5. Check whether Windows SmartScreen is showing an unknown-publisher warning.

Initial release artifacts are unsigned, so SmartScreen may display that warning. Do not bypass it until the asset source and checksum are verified.

## Bootstrapper Does Not Start

Verify the bootstrapper and its MSI came from the same release and that both checksums match. The bootstrapper's package payload is the generated MSI; it is not an independent application distribution.

## Portable ZIP Does Not Run

Extract the complete ZIP to a user-writable directory. Do not run the executable while it is still inside the archive. Verify `wslc --help` separately.

## Installation Appears Machine-Wide

The release packaging is per-user. It is expected to install beneath `%LOCALAPPDATA%\wslc-tui-ms` and use `HKCU\Software\japperJ\wslc-tui-ms`. It should not create services, scheduled tasks, or Program Files state.

If an installation changes machine-wide state, record the release tag and inspect the installer log before reporting the issue.
