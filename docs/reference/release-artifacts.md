# Release Artifacts Reference

Releases are Windows x64 artifacts produced from SemVer tags such as `v1.2.3`.

## Asset Names

| Asset | Purpose |
| --- | --- |
| `wslc-tui-vX.Y.Z-windows-amd64.msi` | Per-user Windows Installer package |
| `wslc-tui-vX.Y.Z-windows-amd64.exe` | Bootstrapper that hands off to the MSI |
| `wslc-tui-vX.Y.Z-windows-amd64-portable.zip` | Extract-and-run portable package |
| `wslc-tui-vX.Y.Z-checksums.json` | SHA-256 manifest for release assets |
| `update-policy.json` | Public update-policy contract |

The updater helper is embedded in the portable ZIP and per-user MSI installation. It is not a separate release asset.

## Channels

The GitHub release `prerelease` flag is the channel source of truth:

- `true` means Beta.
- `false` means Stable.

Promoting a release changes its channel state; it does not rebuild or rename assets.

## Installation State

The MSI and bootstrapper install per-user under `%LOCALAPPDATA%\wslc-tui-ms` and register application state below `HKCU\Software\japperJ\wslc-tui-ms`. They do not modify `PATH`, install services, create scheduled tasks, or write machine-wide Program Files state.

## Checksum Coverage

The checksum manifest covers the MSI, bootstrapper, portable ZIP, and `update-policy.json`. It intentionally excludes itself.
