package app

import (
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"wslc-tui-ms/internal/commands"
)

func TestCanceledExecutionIgnoresLateCompletion(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.currentView = viewOutput
	m.running = true
	m.executionID = 7

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	canceled := updated.(model)
	if canceled.currentView != viewCommands {
		t.Fatalf("canceled view = %v, want commands", canceled.currentView)
	}
	if canceled.executionID == 7 {
		t.Fatal("cancel did not invalidate the active execution generation")
	}

	updated, _ = canceled.Update(execDoneMsg{
		id: 7,
		result: commands.ExecutionResult{
			Output:   "late output",
			Duration: time.Second,
		},
	})
	got := updated.(model)
	if got.outputResult != nil {
		t.Fatalf("late completion populated output: %#v", got.outputResult)
	}
}

func TestExecutionStreamChunkAppearsBeforeCompletion(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.currentView = viewOutput
	m.running = true
	m.executionID = 3

	updated, _ := m.Update(execStreamMsg{id: 3, chunk: "first log line\n"})
	got := updated.(model)

	if got.executionOutput != "first log line\n" {
		t.Fatalf("streamed output = %q, want %q", got.executionOutput, "first log line\n")
	}
	if !strings.Contains(stripAnsi(got.renderRunningView()), "first log line") {
		t.Fatalf("running view omitted streamed output:\n%s", got.renderRunningView())
	}

	updated, _ = got.Update(execStreamMsg{id: 3, done: true, result: commands.ExecutionResult{ExitCode: 0}})
	got = updated.(model)
	if got.running || got.outputResult == nil || got.outputResult.Output != "first log line\n" {
		t.Fatalf("stream completion did not preserve output: running=%v result=%#v", got.running, got.outputResult)
	}
}

func TestValidFormExecutesBuilderArgsAndKeepsDisplay(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.openForm(testFormCommand())
	m.form.argumentRows[0][0] = "ubuntu"
	m.form.argumentRows[1][0] = `echo "quoted"`

	updated, _ := m.handleFormKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(model)

	wantArgs := []string{"wslc", "container", "run", "ubuntu", `echo "quoted"`, "--format", "table", "--timeout", "10"}
	wantDisplay := `wslc container run ubuntu "echo \"quoted\"" --format table --timeout 10`
	if got.currentView != viewOutput {
		t.Fatalf("current view = %v, want output", got.currentView)
	}
	if got.outputCmd != wantDisplay {
		t.Fatalf("output command = %q, want %q", got.outputCmd, wantDisplay)
	}
	if !reflect.DeepEqual(got.outputArgs, wantArgs) {
		t.Fatalf("output args = %#v, want %#v", got.outputArgs, wantArgs)
	}
}

func TestFormConfirmationKeepsBuilderArgs(t *testing.T) {
	m := NewModelForTest(120, 30)
	command := testFormCommand()
	command.Difficulty = "advanced"
	m.openForm(command)
	m.form.argumentRows[0][0] = "ubuntu"
	m.form.argumentRows[1][0] = "echo"

	updated, _ := m.handleFormKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(model)
	if got.currentView != viewConfirm {
		t.Fatalf("current view = %v, want confirmation", got.currentView)
	}
	if got.pendingCommand != "wslc container run ubuntu echo --format table --timeout 10" {
		t.Fatalf("pending command = %q", got.pendingCommand)
	}
	if got.pendingDifficulty != "advanced" {
		t.Fatalf("pending difficulty = %q", got.pendingDifficulty)
	}
	if !reflect.DeepEqual(got.pendingArgs, []string{"wslc", "container", "run", "ubuntu", "echo", "--format", "table", "--timeout", "10"}) {
		t.Fatalf("pending args = %#v", got.pendingArgs)
	}

	wantArgs := append([]string(nil), got.pendingArgs...)
	updated, _ = got.handleConfirmationKey(tea.KeyMsg{Type: tea.KeyEnter})
	got = updated.(model)
	if got.outputDifficulty != "advanced" {
		t.Fatalf("output difficulty = %q, want advanced", got.outputDifficulty)
	}
	if !reflect.DeepEqual(got.outputArgs, wantArgs) {
		t.Fatalf("confirmed args = %#v, want %#v", got.outputArgs, wantArgs)
	}
}

func TestInvalidFormNeverStartsExecution(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.openForm(testFormCommand())

	updated, _ := m.handleFormKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(model)
	if got.currentView != viewForm {
		t.Fatalf("current view = %v, want form", got.currentView)
	}
	if got.running {
		t.Fatal("invalid form started execution")
	}
	if got.outputCmd != "" {
		t.Fatalf("invalid form set output command %q", got.outputCmd)
	}
}

func TestUnknownTypedCommandKeepsRawExecutionPath(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.inputFocused = true
	m.inputValue = `wslc custom "raw value"`

	updated, _ := m.handleCommandsKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(model)
	if got.outputCmd != m.inputValue {
		t.Fatalf("output command = %q, want raw %q", got.outputCmd, m.inputValue)
	}
	if got.outputArgs != nil {
		t.Fatalf("unknown command unexpectedly received structured args: %#v", got.outputArgs)
	}
}
