# Installation Model

The application supports two user-facing distribution models.

## Per-User Installer

The MSI installs beneath the current user's `%LOCALAPPDATA%\wslc-tui-ms` directory and stores product registration in `HKCU`. The bootstrapper launches the generated MSI rather than carrying a second independent application payload.

This model avoids machine-wide changes. It does not add `wslc` to `PATH`; WSLC must already be installed and discoverable.

## Portable Package

The portable ZIP contains the files needed to run the application from an extracted directory. It is useful when the user does not want an installer or does not have administrator privileges.

The portable package still depends on the host's configured `wslc` executable. Portable means the WSLC TUI files are portable; it does not mean the WSLC runtime is bundled.

## Unsigned Releases

Initial release artifacts are unsigned. Windows SmartScreen may therefore display an unknown-publisher warning. Verify the source and SHA-256 checksum before deciding whether to continue.
