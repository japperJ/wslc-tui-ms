# Troubleshoot Startup

## Confirm WSLC Works

Run:

```powershell
wslc --help
Get-Command wslc
```

If `wslc` is not found, add the directory containing it to `PATH` or repair the WSLC installation. WSLC TUI cannot run catalog commands until the underlying executable works.

## Confirm The Binary

Check the application version:

```powershell
.\wslc-tui.exe --version
```

If this fails, confirm that you are running the Windows x64 release executable and that it was extracted completely.

## Run From PowerShell

Use the explicit relative path when the executable is in the current directory:

```powershell
.\wslc-tui.exe
```

For a source checkout:

```powershell
```

## Check Terminal Behavior

Use a terminal with ANSI color support. Mouse input is optional; keyboard controls work without mouse support. If the screen is unreadable, resize the terminal or use a different Windows terminal emulator.

## Capture A Useful Error

Run the application from PowerShell so the error remains visible:

```powershell
.\wslc-tui.exe *> .\wslc-tui-error.log
```

Record the application version, Windows version, terminal name, the command that failed, and whether `wslc --help` succeeds. Remove secrets and sensitive resource names before sharing logs.
