package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"wslc-tui-ms/internal/commands"
)

func formKey(key tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: key} }

func TestKnownCommandSelectionOpensGuidedForm(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.inputFocused = false
	m.textInput.Blur()
	m.selectedIndex = 0

	updated, _ := m.handleCommandsKey(formKey(tea.KeyEnter))
	got := updated.(model)
	if got.currentView != viewForm {
		t.Fatalf("known command opened view %v, want form", got.currentView)
	}
	if got.form == nil {
		t.Fatal("known command did not initialize form state")
	}
}

func TestFormViewRendersSectionsAndStatsDefault(t *testing.T) {
	m := NewModelForTest(120, 30)
	var stats commands.Command
	for _, command := range m.allCommands["Container"] {
		if command.Name == "stats" {
			stats = command
		}
	}
	m.openForm(stats)
	view := stripAnsiTest(m.View())
	for _, section := range []string{"ARGUMENTS", "OPTIONS", "GENERATED COMMAND", "EXAMPLES / HELP"} {
		if !strings.Contains(view, section) {
			t.Errorf("form view missing section %q", section)
		}
	}
	if !strings.Contains(view, "wslc stats --format table") {
		t.Errorf("form view missing stats default command: %q", view)
	}
	if !strings.Contains(view, "[ ]") {
		t.Error("boolean option was not rendered as an unchecked checkbox")
	}
}

func TestFormKeyboardEditsFieldsAndNavigates(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.openForm(testFormCommand())

	updated, _ := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ubuntu")})
	got := updated.(model)
	if got.form.argumentRows[0][0] != "ubuntu" {
		t.Fatalf("first argument = %q, want ubuntu", got.form.argumentRows[0][0])
	}
	updated, _ = got.handleFormKey(formKey(tea.KeyTab))
	got = updated.(model)
	if got.form.focusedField != 1 {
		t.Fatalf("focused field = %d, want 1 after tab", got.form.focusedField)
	}
}

func TestFormTogglesSelectsAndAddsRepeatableRows(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.openForm(testFormCommand())
	m.form.focusedField = len(m.form.argumentRows)

	updated, _ := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	got := updated.(model)
	if got.form.optionValue("--detach") != "true" {
		t.Fatalf("boolean option = %q, want true", got.form.optionValue("--detach"))
	}
	got.form.focusedField = len(got.form.argumentRows) + 1
	updated, _ = got.handleFormKey(formKey(tea.KeyRight))
	got = updated.(model)
	if got.form.optionValue("--format") != "json" {
		t.Fatalf("select option = %q, want json", got.form.optionValue("--format"))
	}
	if !got.form.addRepeatableRow() || len(got.form.argumentRows) != 3 {
		t.Fatalf("repeatable row was not added: %#v", got.form.argumentRows)
	}
}

func TestFormEnterBlocksInvalidValuesLocally(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.openForm(testFormCommand())
	updated, _ := m.handleFormKey(formKey(tea.KeyEnter))
	got := updated.(model)
	if got.currentView != viewForm {
		t.Fatalf("invalid form left view %v", got.currentView)
	}
	if got.form.validationError == nil {
		t.Fatal("invalid form did not retain validation error")
	}
	if !strings.Contains(stripAnsiTest(got.View()), "argument \"image\"") {
		t.Fatalf("validation error was not rendered locally: %q", stripAnsiTest(got.View()))
	}
}
