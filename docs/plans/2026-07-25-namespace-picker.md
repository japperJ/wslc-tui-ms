# Reusable Resource Namespace Pickers Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add reusable, session-scoped resource pickers to schema-driven command forms while preserving command-input-only free text, all existing command categories, legacy placeholder preview behavior, and safe revalidation before execution.

**Architecture:** Add explicit resource metadata to command schemas instead of inferring picker behavior from argument names or command strings. A small discovery package owns per-resource-type list definitions, parsing, refresh errors, and the in-memory session cache; the app model owns picker state and opens the picker only from a focused form argument. Picker selections serialize back into the existing ordered argument rows, including repeatable rows as multi-select tokens. Discovery is asynchronous and best-effort: a failed refresh disables picker selection but leaves the focused text field editable. The form validates metadata-backed values when the picker opens and again immediately before build/confirmation/execution.

**Tech Stack:** Go 1.24, Bubble Tea 1.3, Bubbles textinput/viewport, Lip Gloss, standard-library tests, host `wslc` CLI.

---

## Assumptions And Boundaries

- “Command-input-only picker” means there is no standalone resource-browser screen and no picker for raw commands; the picker is an optional affordance attached to a schema form’s focused resource argument. Typing arbitrary text remains the fallback in that same field.
- “Global by resource type” means one session cache keyed by resource type (`container`, `image`, `network`, `volume`, `session`), shared by every command category and command that declares the same type.
- The existing seven command categories remain unchanged: Container, Image, Network, Volume, Session, System, and Registry. Picker metadata is added only to fields that represent discoverable resources; paths, image references that may be remote, registry servers, commands, and free-form options remain text-only unless explicitly classified.
- The picker is current-session only: no settings, disk cache, persisted selections, or cross-launch state.
- “Category open” is interpreted as opening a resource-type section inside the picker. If it instead means opening the main TUI command category in the sidebar, confirm that before implementation; the plan includes the narrower picker interpretation.
- Discovery command paths and output format must be confirmed against the installed WSLC CLI before implementation. The implementation should prefer stable machine-readable list commands and fail closed when parsing or command execution fails.

## Existing Code Map

- `internal/commands/commands.go`: schema types and command metadata; currently has arguments/options but no resource binding metadata.
- `internal/data/commands.go`: all catalog commands, seven categories, schema construction, and legacy `Full` strings used by preview/migration behavior.
- `internal/app/form.go`: session option memory, ordered argument rows, form field focus, and schema-based build state.
- `internal/app/app.go`: Bubble Tea model, form navigation/rendering, command confirmation, structured execution, legacy placeholder preview, and command category browser.
- `internal/commands/builder.go`: validation and structured argument construction; must remain the single serialization path for forms.
- `internal/app/preview_placeholder_test.go`: regression coverage for legacy commands whose `Full` values contain `{placeholder}` tokens.

## Implementation Tasks

### Task 1: Define explicit resource metadata

**Files:**
- Modify: `internal/commands/commands.go`
- Test: `internal/commands/schema_test.go`

**Steps:**
1. Add a `ResourceType` string enum/constants for the supported discoverable types: container, image, network, volume, and session.
2. Add explicit resource metadata to `commands.Argument`, including resource type and whether the field is picker-enabled; keep `Repeatable` as the source of truth for multi-selection cardinality.
3. Define the metadata contract for non-repeatable fields versus repeatable fields: one selected token for a scalar field, ordered selected tokens for a repeatable field.
4. Add tests proving metadata is copied as schema data, unknown/empty resource types are treated as text-only, and repeatability is not silently changed by picker metadata.
5. Run `go test ./internal/commands -run 'TestCommandSchema|TestResource' -v`; expect the new tests to pass after the type additions.

### Task 2: Add the discovery boundary and session cache

**Files:**
- Create: `internal/discovery/discovery.go`
- Create: `internal/discovery/discovery_test.go`

**Steps:**
1. Define a discovery client interface accepting `context.Context` and `ResourceType`, returning ordered resource tokens and an error. Keep the injectable command runner inside `internal/discovery`; do not alter the general command executor.
2. Define explicit per-resource definitions containing the exact `wslc` argv and parser/normalizer; do not derive list commands from `Command.Category`, argument names, or `Full` strings.
3. Cover the five discoverable resource types and document which existing WSLC command provides each list. Keep System and Registry available as catalog categories but do not invent resource discovery for them.
4. Add a session cache keyed only by `ResourceType`, storing values, last refresh result, and a disabled/error state. Cache values must be copied on read/write so picker edits cannot mutate shared state.
5. Make refresh return an error without erasing a previously successful list unless the chosen UX explicitly requires stale data to be hidden; expose enough state for the UI to disable selection after a failed current refresh.
6. Test command argv construction, machine-readable parsing, malformed output, non-zero exit, timeout/cancellation, empty results, cache isolation between resource types, and recovery after a later successful refresh.
7. Run `go test ./internal/discovery -v` and `go test ./internal/commands ./internal/discovery`.

### Task 3: Annotate the complete catalog explicitly

**Files:**
- Modify: `internal/data/commands.go`
- Modify: `internal/data/commands_test.go`

**Steps:**
1. Extend the schema helper constructors so resource metadata is written at the catalog definition site, not inferred later by `enrichCatalogSchema`.
2. Annotate every applicable resource argument across Container, Image, Network, Volume, and Session commands, including repeatable arguments such as `containers`, `images`, `networks`, and `volumes`.
3. Leave command arguments that are intentionally free text unannotated: image build paths, arbitrary commands, tags/targets where discovery is unsafe, session commands, registry servers, and System/Registry-only values unless the CLI contract proves otherwise.
4. Preserve all seven existing categories, command names, legacy `Full` strings, schema defaults, and legacy placeholder derivation. Do not replace `Full` with picker output.
5. Add catalog tests that enumerate all categories and assert every picker-enabled field has a valid resource type, repeatable resource fields are terminal, metadata is explicit, and no unsupported category is omitted.
6. Add representative metadata assertions for container, image, network, volume, and session fields plus negative assertions for free-text fields.
7. Run `go test ./internal/data -v` and the existing schema migration/representative command-generation tests.

### Task 4: Extend form state for picker tokens and revalidation

**Files:**
- Modify: `internal/app/form.go`
- Modify: `internal/app/app.go`
- Test: `internal/app/form_test.go`

**Steps:**
1. Represent picker-backed values using the existing ordered `argumentRows`; do not add a second command serialization format. For repeatable fields, each token occupies one row in selection order.
2. Add form helpers to identify whether the focused field is picker-enabled, read its resource type, replace a scalar value, replace repeatable rows, and preserve typed free text when no discovery result is available.
3. Add a form-level revalidation function that refreshes the declared resource type when the picker opens and verifies selected tokens against the latest successful session list before building the command.
4. Decide and test the stale-token policy: selected tokens that disappeared must produce a visible validation error and block execution; typed free text is allowed when picker discovery is unavailable, but should not be silently rewritten.
5. Preserve current option memory behavior and current-session scope: option values may continue to be remembered in `formOptionMemory`, but picker values and discovery results must not be persisted or reused across process launches.
6. Add tests for scalar selection, repeatable multi-selection, token ordering, clearing/replacing selections, free-text fallback after discovery failure, stale selection rejection, and no cross-command/category leakage.
7. Run `go test ./internal/app -run 'TestCommandForm|TestResourcePicker|TestForm' -v`.

### Task 5: Implement the reusable picker interaction

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/ui/styles.go`
- Test: `internal/app/form_view_test.go`
- Test: `internal/app/click_regions_test.go`

**Steps:**
1. Add picker state to the model: open/closed, owning form field, resource type, filter text, focused candidate, selected tokens, loading state, refresh error, and whether selection is disabled.
2. Add a Bubble Tea command/message pair for background refresh. Opening a picker must immediately show loading state, start refresh in the background, and never block the main event loop.
3. Render a reusable picker panel/section with resource-type heading, filter input, candidate list, selected multi-select tokens, empty/loading/error states, and an explicit “type a value instead” affordance. Do not create a new standalone navigation category.
4. Allow picker open only for a focused schema argument with explicit metadata. Keep ordinary textinput behavior, including arbitrary free text, when the field has no picker metadata or when discovery failed.
5. Define keyboard/mouse behavior in tests: open picker, filter, move focus, toggle one or more tokens, commit selection, cancel without mutation, and continue typing after a failed refresh. Use existing form click-region and viewport conventions.
6. Refresh on every picker open, including reopening after cancel; do not rely only on the session cache. “Background refresh on category open” applies to the picker resource-type section under the assumption above.
7. Run `go test ./internal/app -run 'TestFormView|TestResourcePicker|TestClick' -v` and inspect rendered output at narrow and normal test dimensions.

### Task 6: Wire discovery results and picker commits into form execution

**Files:**
- Modify: `internal/app/app.go`
- Test: `internal/app/execution_test.go`
- Test: `internal/app/preview_placeholder_test.go`

**Steps:**
1. Route picker commits into form rows, then call the existing `commands.Build` path so generated display text and structured `[]string` arguments remain consistent. `internal/commands/builder.go` is not expected to change.
2. Revalidate when the picker opens, after an asynchronous refresh completes, and immediately before `handleFormKey(Enter)` submits. Never trust a previously rendered candidate list at execution time.
3. Ensure validation failure keeps the user in `viewForm`, identifies the stale/missing resource field, and starts no process or confirmation flow.
4. Ensure successful form execution still passes structured args to `exec.CommandContext`, while raw command input continues through the existing parser/compatibility path.
5. Preserve intermediate/advanced confirmation and revalidate again when confirmation is accepted, because the resource list may have changed while the confirmation screen was open.
6. Add tests for picker-selected execution, multi-token execution order, refresh failure with free-text execution, stale selection blocked before confirmation, stale selection blocked after confirmation wait, and raw command execution unaffected.
7. Run `go test ./internal/app -run 'TestStructuredExecution|TestConfirmation|TestResourcePicker' -v`.

### Task 7: Preserve and clarify legacy placeholder preview

**Files:**
- Modify: `internal/app/app.go` only where routing distinguishes schema forms from legacy commands
- Modify: `internal/app/preview_placeholder_test.go`
- Modify: `internal/commands/placeholders.go` only if required by regression behavior

**Steps:**
1. Keep commands without usable schemas on the existing `viewPreview` path, including `{placeholder}` extraction, click regions, empty-placeholder blocking, and substituted display command behavior.
2. Do not open the resource picker for legacy placeholders because they lack explicit metadata; their fields remain free text.
3. Add a regression test proving a legacy placeholder named like a resource (`{name}`, `{image}`, etc.) does not activate discovery implicitly.
4. Retain tests for no-placeholder preview execution, confirmation cancel-back behavior, and placeholder rendering/click focus.
5. Run `go test ./internal/app -run 'Test.*Placeholder|Test.*Preview' -v`.

### Task 8: Full regression, documentation, and manual acceptance

**Files:**
- Modify: `internal/app/app_test.go`
- Modify: `internal/data/commands_test.go`
- Modify: `internal/app/form_test.go`
- Modify: `internal/app/form_view_test.go`
- Modify: `internal/app/execution_test.go`
- Modify: `docs/explanation/architecture.md` to document session-only discovery and free-text fallback

**Steps:**
1. Add an end-to-end test matrix covering each existing category, at least one picker-enabled scalar field, one repeatable multi-select field, one text-only field, and the legacy preview path.
2. Add model-message tests proving category/picker open starts refresh asynchronously, successful refresh enables selection, failed refresh disables only selection, and a later refresh recovers.
3. Add tests proving cache sharing by resource type across commands/categories and isolation between resource types, with no persistence after constructing a new model.
4. Run `gofmt -w internal/commands internal/discovery internal/data internal/app internal/ui`.
5. Run `go vet ./...`.
6. Run `go test ./...`.
7. Run `go build -o wslc-tui-ms.exe .`.
8. Manually verify: open each picker-backed form field, observe a background refresh, select one and several tokens, type an arbitrary value after simulated discovery failure, cancel/reopen the picker, and attempt execution after removing a selected resource. Confirm stale values are rejected and legacy placeholder commands still behave as before.

## Verification Checklist

- All seven existing command categories remain present and selectable.
- Picker behavior is driven only by explicit schema metadata.
- Picker state and discovery cache last for the current process only.
- Resource lists are shared globally by resource type, not duplicated per command category.
- Every picker open starts a background refresh.
- Discovery failure disables picker selection but never disables free-text editing.
- Repeatable fields produce ordered multi-select tokens and structured execution args.
- Picker values are revalidated on open and immediately before execution/confirmation acceptance.
- Raw commands and legacy placeholder previews remain free-text/legacy paths.
- No command executes from an invalid or stale picker selection.

## Ambiguities To Resolve Before Implementation

1. Does “category open” mean opening a resource-type section in the picker, or selecting one of the seven main command-browser categories? This plan assumes the former.
2. Which exact WSLC list commands and machine-readable output schemas are guaranteed by the target CLI version for containers, images, networks, volumes, and sessions? Confirm with `wslc --help` and representative list invocations before locking discovery definitions.
3. Should stale selected tokens be rejected, or retained as typed free text after a refresh? This plan recommends rejecting them as picker selections while allowing a user to explicitly replace the field with free text.
4. Should a scalar picker field permit a typed value that is not in the discovered list even when discovery succeeds? This plan recommends yes for the stated free-text requirement, with picker validation applying only to tokens selected through the picker.
