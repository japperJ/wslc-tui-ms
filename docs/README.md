# WSLC TUI Documentation

WSLC TUI is a Windows terminal user interface for discovering and running WSLC container commands. It provides searchable commands, guided forms, previews, safety confirmations, and readable output.

## Choose A Starting Point

- [Run WSLC TUI on Windows](tutorials/run-on-windows.md): Follow a complete first-use walkthrough.
- [Install and update](how-to/install-and-update.md): Install a release or update an existing installation.
- [Publish a release](how-to/publish-release.md): Build and publish Stable or Beta releases manually.
- [Troubleshoot startup](how-to/troubleshoot-startup.md): Diagnose failures before or during launch.
- [Troubleshoot installation](how-to/troubleshoot-installation.md): Diagnose MSI, bootstrapper, and portable-package issues.
- [Command-line reference](reference/command-line.md): Find supported command-line options.
- [Release artifacts](reference/release-artifacts.md): Understand MSI, bootstrapper, portable ZIP, and checksum files.
- [Architecture](explanation/architecture.md): Understand how the application connects the TUI to WSLC.

## Documentation Areas

### Tutorials

Learning-oriented lessons for completing a first successful task.

### How-To Guides

Problem-oriented recipes for installation, verification, diagnostics, and support.

### Reference

Technical descriptions of commands, files, paths, release assets, and maintenance procedures.

### Explanation

Conceptual background about the application, installation model, and release flow.

## Important Boundary

WSLC TUI is a front end for the `wslc` executable. It does not install WSLC, manage the WSLC runtime, or replace the WSLC command-line interface. Confirm that `wslc --help` works before troubleshooting WSLC TUI itself.
