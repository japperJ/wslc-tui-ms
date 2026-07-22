# Guided Command Forms Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace fixed catalog command strings with schema-driven guided forms that support defaults, selectable flags, multiple positional arguments, repeatable rows, validation, and safe execution.

**Architecture:** Add declarative argument and option metadata to `internal/commands`, then use a pure builder to validate form state and produce both structured executable arguments and a display command. Adapt the existing preview TUI into a single scrollable form for known commands while preserving raw execution for unknown commands and the existing confirmation flow.

**Tech Stack:** Go, Bubble Tea, Bubbles `textinput`/`viewport`, Lip Gloss, standard-library testing.

---

### Task 1: Add schema types

**Files:**
- Modify: `internal/commands/commands.go`
- Test: `internal/commands/schema_test.go`

**Step 1: Write the failing tests**

Add tests for ordered positional arguments and options with control types, defaults, choices, required status, repeatability, and validation metadata.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/commands -run TestCommandSchema -v`
Expected: FAIL because schema types do not exist.

**Step 3: Implement the minimal schema types**

Add types for:

- `Argument`: name, label, required, repeatable, placeholder, and value validation metadata.
- `Option`: flag, description, kind, default, choices, required, and value validation metadata.
- `OptionKind`: boolean, text, select, and numeric.
- `CommandSchema`: ordered arguments and ordered options.

Add `Schema *CommandSchema` to `Command`. Keep `Full` temporarily for migration comparison.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/commands -run TestCommandSchema -v`
Expected: PASS.

---

### Task 2: Build and validate commands

**Files:**
- Create: `internal/commands/builder.go`
- Test: `internal/commands/builder_test.go`

**Step 1: Write failing builder tests**

Cover:

- `stats` default output: `wslc stats --format table`.
- Select value `json`.
- Boolean flags included only when enabled.
- Empty optional values omitted.
- Required values rejected.
- Multiple positional values preserved in order.
- Repeatable arguments serialized in row order.
- Invalid select and numeric values rejected.
- Structured `[]string` arguments returned separately from display text.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/commands -run TestBuild -v`
Expected: FAIL because the builder does not exist.

**Step 3: Implement the builder**

Define form values using ordered argument rows and option values. Implement a pure `Build` function that returns executable arguments, display command, and validation errors. Do not shell-parse the display command. Quote only for display; pass the structured argument slice to `exec.CommandContext`.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/commands -run TestBuild -v`
Expected: PASS.

---

### Task 3: Migrate command schemas

**Files:**
- Modify: `internal/data/commands.go`
- Modify: `internal/data/commands_test.go`

**Step 1: Add catalog schema tests**

Require every catalog command to have a schema, require schema argument names to be unique, and verify each default command against its expected `Full` value during migration.

**Step 2: Run the tests to identify missing schemas**

Run: `go test ./internal/data -run TestCommandSchemas -v`
Expected: FAIL listing unmigrated commands.

**Step 3: Migrate definitions by category**

Add schemas for Container, Image, Network, Volume, Session, System, and Registry commands. Model fixed positional arguments explicitly, and model variable-length operands as repeatable arguments. Convert `stats --format` to a select with `table` default and `json` choice. Preserve descriptions, examples, difficulty, and tags.

**Step 4: Run catalog tests**

Run: `go test ./internal/data -v`
Expected: PASS.

---

### Task 4: Add form state and session option memory

**Files:**
- Modify: `internal/app/app.go`
- Create: `internal/app/form.go`
- Test: `internal/app/form_test.go`

**Step 1: Write failing form-state tests**

Test initialization from schema defaults, option memory per command, clearing positional values on reopen, add/remove repeatable rows, focus movement, and form validation error retention.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run TestCommandForm -v`
Expected: FAIL because form state does not exist.

**Step 3: Implement form state helpers**

Create a model for ordered argument rows, typed option values, focused field, validation error, and generated command. Store remembered options in the Bubble Tea model keyed by command identity. Do not store positional values between openings.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/app -run TestCommandForm -v`
Expected: PASS.

---

### Task 5: Render and navigate the guided form

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/ui/styles.go`
- Test: `internal/app/form_view_test.go`

**Step 1: Write failing view and key-handling tests**

Verify `stats` renders its default command, select values can be changed, boolean options toggle, repeatable rows can be added/removed, required-field errors render, and Enter routes through validation.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run TestFormView -v`
Expected: FAIL because the current preview is placeholder-driven.

**Step 3: Implement the single scrollable form**

Replace known-command placeholder editing with sections for Arguments, Options, Generated Command, and Examples/Help. Preserve clickable rows, keyboard navigation, viewport scrolling, and existing status-bar conventions. Render checkboxes for booleans, select controls for choices, and dynamic text rows for arguments.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/app -run 'TestFormView|TestCommandForm' -v`
Expected: PASS.

---

### Task 6: Execute structured arguments safely

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/commands/executor.go`
- Test: `internal/app/execution_test.go`

**Step 1: Write failing execution tests**

Verify a known command executes using structured arguments, the displayed command matches the executed arguments, unresolved or invalid forms never execute, confirmation receives the final command, and unknown raw commands retain their existing path.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run TestStructuredExecution -v`
Expected: FAIL because execution currently reparses a command string.

**Step 3: Update execution flow**

Add an execution entry point accepting executable name and argument slice. Keep the display command for output headers, clipboard, confirmation, and rerun. Route known forms through validation, confirmation, and structured execution. Keep raw commands on the existing compatibility path.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/app -run 'TestStructuredExecution|TestConfirmation' -v`
Expected: PASS.

---

### Task 7: Complete regression coverage and cleanup

**Files:**
- Modify: `internal/data/commands_test.go`
- Modify: `internal/commands/commands_test.go`
- Modify: `internal/app/preview_placeholder_test.go`
- Modify: `internal/app/form_test.go`

**Step 1: Add end-to-end acceptance tests**

Cover `stats` from initialization through command generation, format change, toggles, repeatable containers, validation, confirmation, and execution. Cover `tag`, `network connect`, `run`, and a variable-length stop/remove command.

**Step 2: Remove obsolete known-command placeholder assumptions**

Keep placeholder utility tests only for raw or legacy behavior still needed. Update tests that expect fixed `Full` execution to assert schema-generated commands instead.

**Step 3: Run the full verification suite**

Run: `gofmt -w internal/commands internal/data internal/app internal/ui`

Run: `go vet ./...`

Run: `go test ./...`

Expected: formatting completes, vet passes, and all tests pass.

**Step 4: Build the application**

Run: `go build -o wslc-tui-ms.exe .`

Expected: build succeeds and the executable launches with the guided form flow.

**Step 5: Manual acceptance check**

Launch `./wslc-tui-ms.exe`, select `stats`, confirm the initial command is `wslc stats --format table`, change format to `json`, add two container rows, and verify the final command and confirmation screen match the intended structured arguments.
