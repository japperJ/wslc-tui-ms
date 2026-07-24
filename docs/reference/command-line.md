# Command-Line Reference

WSLC TUI is primarily an interactive terminal application.

## Start The Interactive UI

From a release directory:

```powershell
.\wslc-tui.exe
```

From the source repository:

```powershell
```

## Show Version Information

Use either option:

```powershell
.\wslc-tui.exe --version
.\wslc-tui.exe -v
```

The output is embedded into release binaries during packaging. A normal local `go build` may contain development build information.

## Error Behavior

If the terminal program exits with an error, the application writes the error to standard error and exits with status `1`. Missing arguments are not currently a documented command-line interface; use the interactive UI instead.

## Underlying WSLC Command

Catalog commands and ad-hoc commands are executed through the `wslc` executable found on the host `PATH`. Verify it independently with:

```powershell
wslc --help
```
