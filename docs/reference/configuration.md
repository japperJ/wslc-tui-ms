# Configuration Reference

## Runtime Dependencies

WSLC TUI requires:

- Windows.
- A configured WSLC installation.
- `wslc.exe` available on `PATH`.
- A terminal with ANSI color support.

Mouse support is optional. Clipboard access is required only for copy actions.

## Update Settings

Update settings are stored in the operating system's per-user configuration directory. The exact location is selected by the application configuration layer and should be treated as implementation detail rather than a file to edit manually.

Automatic update checks are throttled to one request per 24 hours. Network failures are silent during automatic checks.

## Command Catalog

The built-in catalog is defined in `internal/data/commands.go`. It contains schemas used to render guided forms and validate inputs. The catalog covers containers, images, networks, volumes, sessions, system operations, and registry operations.

## Local Source Build

The repository declares Go `1.24.5` in `go.mod`. Build or run from the repository root:

```powershell
```
