# Install And Update WSLC TUI

## Install From An MSI

1. Download the Windows x64 MSI for the desired release.
2. Verify the release asset with the published checksum manifest.
3. Run the MSI as the intended Windows user.
4. Follow the installer prompts.
5. Launch `wslc-tui.exe` from the installed application location or the Start menu entry provided by the release.
6. Verify the application:

```powershell
.\wslc-tui.exe --version
```

The MSI is designed for per-user installation under `%LOCALAPPDATA%\wslc-tui-ms`. It does not require an administrator account.

## Install The Portable ZIP

1. Download the Windows x64 portable ZIP.
2. Verify its checksum.
3. Extract it to a directory writable by the current user.
4. Run `wslc-tui.exe` from the extracted directory.

The portable package contains the application, updater helper, README, and license files. It does not register a machine-wide installation.

## Check For Updates In The Application

1. Press `u` in the command browser.
2. Select Stable or Beta with `c`.
3. Review the available release information.
4. Press `d` to review and confirm the update handoff.
5. Press `l` to defer a non-mandatory update.

Automatic checks are silent when the network is unavailable and are limited to one request every 24 hours. The current update phase creates metadata for a later handoff; it does not download or install a release itself.

## Remove The Application

Remove an MSI installation through Windows Settings or the installed product's uninstall entry. Remove a portable installation by deleting the extracted directory. Per-user application data may remain in the user's configuration directories.
