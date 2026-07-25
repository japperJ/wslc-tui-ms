# Key Input Routing Design

## Goal

Allow users to type arbitrary command names such as `wslc` in the command
search field without the `c` key triggering the Stable/Beta update-channel
toggle. Preserve `c` as a shortcut everywhere that no text editor is active.

## Architecture

Focused text input is an exclusive editing context. Printable characters are
sent to the active editor before global printable shortcuts are considered.
This applies to command search, guided form arguments and options, and preview
placeholder fields. The active view continues to own control keys such as
`Enter`, `Tab`, `Esc`, arrows, and form navigation.

The change remains in the existing Bubble Tea routing in
`internal/app/app.go`. The command browser's focused-input branch will stop
special-casing `c` and `C` as update actions. The non-focused browser branch
will retain the channel shortcut. The same routing rule should prevent other
printable shortcuts, including `u`, `l`, and `q`, from interrupting text entry.

No update service, settings, command execution, or persistence behavior
changes. When editing ends with `Esc`, normal browser shortcuts become active
again. On the update screen, `c` continues to switch channels; on the output
screen, it continues to copy the command.

## Event Flow

1. Handle application-wide termination only where it is already safe, such as
   `Ctrl+C`.
2. Route the key event to the active view.
3. If a text editor is focused, let it consume printable input.
4. Otherwise, process the active view's shortcuts normally.

The command browser will retain explicit handling for `Esc`, `Enter`, and
`Tab`, because those controls manage search focus, command selection, and
completion. Other keys will be passed through to the focused text input.
When no editor is focused, `c` will toggle Stable/Beta and start the existing
manual update check.

## Tests

Add regression coverage in `internal/app` using synthetic `tea.KeyMsg` values:

- Typing `wslc` into focused command search produces exactly `wslc`.
- Lowercase and uppercase `c` do not change the update channel while editing.
- `Esc` restores browser shortcuts, and `c` then toggles the channel.
- Focused form and preview editors accept printable shortcut characters.
- Update-screen channel switching and output-screen command copying remain
  unchanged.

Update the existing focused-search tests that currently expect `c` to toggle
the channel. Run `gofmt` on modified Go files, then verify with:

```powershell
go test ./internal/app
go test ./...
```

Success means `wslc` can be typed and submitted as a search value or ad-hoc
command, while channel switching still works outside text editing.
