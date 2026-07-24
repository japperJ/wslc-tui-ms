# Architecture

WSLC TUI has three practical layers.

## Terminal Application

`main.go` handles the process entry point, version requests, and Bubble Tea program startup. The interactive model and views live under `internal/app`.

## Command Definition And Execution

`internal/data` defines the built-in command catalog and form schemas. `internal/commands` handles searching, autocomplete, command construction, validation, and execution. The application combines standard output and standard error so users can inspect the complete result in the output view.

Unknown text entered through search can be passed to the host command runner as an ad-hoc command. This is why users must review previews before execution.

## Presentation

`internal/ui` contains Lip Gloss styles and visual helpers. Bubble Tea provides the event loop and terminal rendering; Bubbles provides reusable terminal components.

## External Boundary

WSLC TUI does not implement container runtime operations itself. It invokes the `wslc` executable available on the host. WSLC remains responsible for command semantics, resources, permissions, and runtime errors.
