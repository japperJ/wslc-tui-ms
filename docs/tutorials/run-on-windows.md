# Run WSLC TUI On Windows

This tutorial takes a Windows user from a working WSLC installation to a first command run in WSLC TUI.

## Prerequisites

You need:

- Windows with WSLC installed and configured.
- The `wslc` executable available on `PATH`.
- A terminal with ANSI color support.
- Clipboard access only if you want to use copy actions.

Open PowerShell and verify the underlying CLI:

```powershell
wslc --help
```

If this command fails, fix WSLC or its `PATH` entry before continuing.

## Start The Application

For a downloaded release, launch the executable from PowerShell:

```powershell
.\wslc-tui.exe
```

For a source checkout, use:

```powershell
```

The application opens in the command browser.

## Find And Run A Command

1. Press `/` to focus search.
2. Type a command name or prefix.
3. Press `Tab` to complete a matching command when available.
4. Press `Enter` to open the selected command.
5. Complete the guided form.
6. Review the generated command preview.
7. Press `Enter` or `Y` to run it.
8. Review the combined standard output and error output.

Intermediate and advanced commands display an additional confirmation screen. Treat those commands as potentially destructive.

## Explore The Interface

- Press `1` through `7` to switch command categories.
- Press `l` to open built-in learning topics.
- Press `?` to toggle help.
- Press `Esc` to go back or unfocus search.
- Press `Ctrl+C` to quit.

After a command finishes, press `r` to run it again, `c` to copy the command, or `y` to copy output.

## Finish Safely

Before running cleanup, removal, stop, or prune commands, verify the preview and target resources. Use non-production resources while learning the interface.
