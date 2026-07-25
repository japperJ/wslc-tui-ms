# WSLC TUI

An interactive terminal UI for discovering and running [WSLC](https://github.com/microsoft/wsl) container commands from Windows.

WSLC TUI gives command-line users a searchable command catalog, guided forms for command arguments and options, safety confirmations, and readable command output without requiring every command to be memorized.

> **Status:** This is an early-stage project. The command catalog and UI are actively evolving.

For task-oriented Windows documentation, start with the [documentation hub](docs/README.md).

## What It Does

- Browse WSLC commands by category: Container, Image, Network, Volume, Session, System, and Registry.
- Search commands with prefix filtering and autocomplete.
- Open known commands in guided forms instead of manually constructing every flag.
- Edit positional arguments, repeatable argument rows, boolean flags, text values, numeric values, and select options.
- Preview the generated command before it runs.
- Validate required and structured form inputs locally.
- Require confirmation before intermediate and advanced commands that can change system state.
- Run ad-hoc commands that are not in the catalog.
- Display stdout and stderr together with exit status and execution duration.
- Pretty-print JSON output and provide scrollable output views.
- Re-run the last command or copy the command/output to the clipboard.
- Read built-in learning topics covering common WSLC workflows.
- Check GitHub Releases without blocking the command browser, with Stable/Beta channels, release notes, deferral, and explicit download confirmation.
- Support keyboard navigation and mouse clicks/wheel scrolling in supported terminals.

## Requirements

### Runtime

- Windows with WSLC installed and configured.
- The `wslc` executable available on `PATH`.
- Go `1.24.5` or a compatible Go 1.24 toolchain when running from source.
- A terminal emulator with ANSI color support. Mouse features are optional; keyboard controls work without them.
- Clipboard access if you want to use the copy actions.

WSLC TUI is a front end for WSLC. It does not install WSLC, manage its runtime, or replace the WSLC CLI. Verify the underlying CLI first:

```powershell
wslc --help
```

## Quick Start

Clone the repository and run the application from the project directory:

```powershell
git clone <repository-url>
cd wslc-tui-ms
go run .
```

The application starts in the command browser. Select a command with `↑`/`↓` and press `Enter` to preview or open its guided form.

### Build A Standalone Binary

```powershell
go build -o wslc-tui-ms.exe .
.\wslc-tui-ms.exe
```

The generated executable can be placed in a directory on your `PATH` if you want to launch it from anywhere.

Release binaries expose their embedded version and distribution metadata:

```powershell
.\wslc-tui-ms.exe --version
```

## Windows Releases

Maintainers publish Windows x64 MSI, EXE/bootstrapper, and portable ZIP assets from SemVer tags such as `v1.2.3`. The MSI and bootstrapper install per-user beneath `%LOCALAPPDATA%\wslc-tui-ms` and register only `HKCU\Software\japperJ\wslc-tui-ms`; they do not modify `PATH`, install services, or write machine-wide state. Portable ZIPs contain `wslc-tui.exe`, `wslc-tui-updater.exe`, `README.txt`, and `LICENSE.txt`.

Release policy is the version-controlled [`update-policy.json`](update-policy.json), consumed publicly at `https://raw.githubusercontent.com/japperJ/wslc-tui-ms/main/update-policy.json`. Release tags select the channel: `v1.2.3` is Stable and `v1.2.3-beta.1` is Beta. GitHub's release `prerelease` flag is set from that channel: `true` is Beta and `false` is Stable. Promotion changes release state only and does not rebuild assets. Initial artifacts are unsigned, so Windows SmartScreen may show an unknown-publisher warning; signing is an optional future workflow hook.

The maintainer packaging and manual publishing procedure is documented in [`docs/releases.md`](docs/releases.md).
The offline disposable-VM bundle is built with [`scripts/build-phase1-test-iso.ps1`](scripts/build-phase1-test-iso.ps1); it never overwrites the checked-in `wslc-tui-ms.iso`.

## Basic Workflow

1. Start the application with `go run .` or the built executable.
2. Search for a command with `/`, or browse a category with `1` through `7`.
3. Press `Enter` to open the selected command.
4. Complete the guided form, or edit placeholders in the preview where applicable.
5. Review the generated command. Intermediate and advanced commands show an additional confirmation screen.
6. Press `Enter` or `Y` to run the command.
7. Review output, scroll through longer results, copy the command/output, or press `r` to run it again.

Unknown commands typed into the search field are passed through to the host shell command runner. Review ad-hoc commands carefully before submitting them.

## Keyboard Controls

### Command Browser

| Key | Action |
| --- | --- |
| `↑` / `↓`, `j` / `k` | Navigate commands |
| `Enter` | Preview or run the selected command |
| `/` | Focus search |
| `Tab` | Complete a matching command |
| `1` - `7` | Switch command category |
| `l` | Open Learn |
| `?` | Toggle help |
| `Esc` | Unfocus search or go back |
| `Ctrl+C` | Quit |

While any text field is focused, printable characters are entered as text;
browser shortcuts such as update actions do not interrupt editing.

### Forms And Output

| Key | Action |
| --- | --- |
| `↑` / `↓`, `Tab` | Move between form fields |
| `Space` | Toggle a boolean option |
| `←` / `→` | Change a select option |
| `+` / `=` | Add a repeatable argument row |
| `-` | Remove the focused repeatable argument row |
| `Enter` | Submit a form or execute a preview |
| `Esc` | Go back or cancel |
| `r` | Re-run the last command from output |
| `y` | Copy output |
| `c` | Copy the command |
| `g` / `G` | Jump to the top/end of output |
| Mouse wheel | Scroll forms and output |

### Updates

| Key | Action |
| --- | --- |
| `u` | Check for updates manually |
| `d` | Review and confirm the selected update handoff |
| `l` | Defer a non-mandatory update |
| `c` | Switch between Stable and Beta |
| `r` | Bypass the automatic 24-hour check cooldown |

Automatic checks are silent when the network is unavailable and are throttled to one request per 24 hours. Update settings are stored in the operating system's per-user configuration directory. Phase 2 never downloads or installs a release; `Download and Install` only creates the metadata handoff consumed by Phase 3.

## Supported Command Areas

The built-in catalog currently covers:

- Container lifecycle, execution, logs, inspection, stats, export, and cleanup
- Image listing, pulling, pushing, tagging, building, saving/loading, and cleanup
- Network creation, inspection, connection, disconnection, and cleanup
- Volume creation, inspection, removal, and cleanup
- Session listing, entering, running, shell access, and termination
- WSLC version information
- Registry login and logout

The catalog is defined in [`internal/data/commands.go`](internal/data/commands.go), with command schemas used to build and validate guided forms.

## Development

Run the complete test suite:

```powershell
go test ./...
```

Run static checks and format the source:

```powershell
gofmt -w main.go (Get-ChildItem -Recurse -Filter *.go internal | ForEach-Object FullName)
go vet ./...
```

Run focused tests while working on a package:

```powershell
go test ./internal/commands -v
go test ./internal/app -v
go test ./internal/data -v
```

### Project Layout

```text
.
├── main.go                    # Bubble Tea program entry point
├── internal/app/              # TUI model, views, navigation, and execution flow
├── internal/commands/         # Command schemas, autocomplete, building, and execution helpers
├── internal/data/             # Built-in WSLC command catalog
├── internal/ui/               # Lip Gloss styles and visual helpers
└── docs/plans/                # Design and implementation notes
```

## Safety Notes

WSLC TUI executes commands on the local machine using the same `wslc` executable available in your environment. It can start, stop, remove, prune, and otherwise modify containers, images, networks, volumes, sessions, and related resources.

- Treat advanced commands as destructive until verified.
- Review the generated command and confirmation screen before running it.
- Test commands against non-production resources first.
- Do not paste secrets into command arguments or terminal recordings.

## Contributing

Bug reports, command catalog improvements, UI feedback, and pull requests are welcome. Before opening a pull request, run:

```powershell
gofmt -w main.go (Get-ChildItem -Recurse -Filter *.go internal | ForEach-Object FullName)
go vet ./...
go test ./...
```

When adding a catalog command, update its metadata and schema together, include representative examples, and add tests for command building and validation where appropriate.

## License

No license file is currently included in this repository. Add or confirm the project license before distributing the software.
