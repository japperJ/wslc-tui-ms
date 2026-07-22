# Guided Command Forms Design

## Goal

Allow catalog commands to be configured through a guided form instead of relying on one fixed command string. Users should be able to choose supported options, provide multiple positional values, review the generated command, and then execute it safely.

## Decisions

- Known catalog commands use guided forms.
- Unknown typed commands keep the existing raw-command execution path.
- The command catalog becomes schema-first and declarative.
- Forms use one scrollable view with live command generation.
- Defaults are visible and included in generated commands. For example, `stats` starts as `wslc stats --format table`.
- Repeatable arguments use dynamic rows with add/remove controls.
- Option choices are remembered per command for the current app session.
- Positional values are cleared when reopening a command to avoid repeating an operation against the wrong resource.
- Existing difficulty levels and confirmation behavior remain in place.

## Command Schema

`commands.Command` will describe how to build a command, not only how to display it. The schema will include ordered positional arguments and ordered options.

Positional argument metadata includes:

- Name and display label
- Required or optional status
- Repeatable status
- Placeholder/help text
- Validation rules

Option metadata includes:

- Flag name and description
- Control type: boolean, text, select, or numeric
- Default value
- Allowed values for select controls
- Required status and validation rules

Examples:

- `stats`: optional repeatable containers; `--format` select with `table` default and `json` option; `--all`; `--no-trunc`.
- `tag`: required `source` and `target` arguments.
- `network connect`: required `network` and `container` arguments.
- `stop`: repeatable containers and optional timeout value.

The existing `Full` field may remain temporarily during migration, but the form and builder will use the schema rather than parsing `Usage` or inferring behavior from `Full`.

## Form Flow

Selecting a known command initializes a mutable form state from schema defaults and remembered option values. The form contains Arguments, Options, Generated Command, and Examples/Help sections.

The generated command is rebuilt after every edit. Optional empty fields and disabled boolean flags are omitted. Defaults are included. Required fields, invalid select values, invalid numeric values, and incomplete repeatable rows block execution with local actionable errors.

The builder returns both the executable argument array and the display command. Execution should use the structured argument array directly instead of reparsing the display string. The exact display command is used in the review/confirmation screen and output history.

Changing form values invalidates pending confirmation. Cancelling returns to the command list without executing or saving incomplete positional values.

## Migration

1. Add schema types and a command builder without changing the current catalog behavior.
2. Add builder and validation tests.
3. Migrate command definitions category by category.
4. Replace the known-command preview flow with the guided form.
5. Keep raw execution for unknown commands.
6. Remove the legacy fixed-command dependency after all catalog commands are migrated.

## Testing

Tests must cover:

- Boolean, text, select, and numeric option serialization
- Default inclusion, especially `--format table`
- Omission of disabled and empty optional fields
- Required argument validation
- Repeatable row add/remove and serialization
- Multiple positional arguments such as `tag source target`
- Variable-length arguments such as multiple containers
- Stable option and argument ordering
- Per-command session option memory
- Positional-value reset when reopening a command
- Confirmation and raw-command compatibility

The first acceptance case is `stats`: it opens as `wslc stats --format table`, supports format selection, toggles `--all` and `--no-trunc`, accepts zero or more container rows, and shows the exact final command before execution.
