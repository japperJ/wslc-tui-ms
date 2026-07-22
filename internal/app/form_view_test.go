package app

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"wslc-tui-ms/internal/commands"
	"wslc-tui-ms/internal/data"
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

func catalogCommandForApp(t *testing.T, category, name string) commands.Command {
	t.Helper()
	for _, command := range data.GetCommandsByCategory(category) {
		if command.Name == name {
			return command
		}
	}
	t.Fatalf("catalog command %s/%s not found", category, name)
	return commands.Command{}
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

func TestCatalogFormsOpenAndBuildAcceptanceCommands(t *testing.T) {
	tests := []struct {
		name        string
		category    string
		command     string
		values      [][]string
		options     map[string]string
		wantView    view
		wantDisplay string
		wantArgs    []string
	}{
		{
			name:     "stats defaults",
			category: "Container", command: "stats",
			wantView:    viewForm,
			wantDisplay: "wslc stats --format table",
			wantArgs:    []string{"wslc", "stats", "--format", "table"},
		},
		{
			name:     "tag source and target",
			category: "Image", command: "tag",
			values:      [][]string{{"source:latest"}, {"target:v1"}},
			wantView:    viewForm,
			wantDisplay: "wslc tag source:latest target:v1",
			wantArgs:    []string{"wslc", "tag", "source:latest", "target:v1"},
		},
		{
			name:     "network connect",
			category: "Network", command: "connect",
			values:      [][]string{{"frontend"}, {"web"}},
			wantView:    viewForm,
			wantDisplay: "wslc network connect frontend web",
			wantArgs:    []string{"wslc", "network", "connect", "frontend", "web"},
		},
		{
			name:     "run image and command",
			category: "Container", command: "run",
			values:      [][]string{{"ubuntu:latest"}, {"echo"}, {"hello"}},
			options:     map[string]string{"--detach": "true", "--name": "worker"},
			wantView:    viewForm,
			wantDisplay: "wslc run ubuntu:latest echo hello --detach --name worker",
			wantArgs:    []string{"wslc", "run", "ubuntu:latest", "echo", "hello", "--detach", "--name", "worker"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := NewModelForTest(120, 30)
			command := catalogCommandForApp(t, test.category, test.command)
			m.enterPreview(command)
			if m.currentView != test.wantView {
				t.Fatalf("opened view = %v, want form", m.currentView)
			}
			for len(m.form.argumentRows) < len(test.values) {
				if !m.form.addRepeatableRow() {
					t.Fatal("could not add repeatable argument row")
				}
			}
			for index, row := range test.values {
				m.form.argumentRows[index] = row
			}
			for flag, value := range test.options {
				m.form.setOption(flag, value)
			}
			result := m.formBuild()
			if len(result.Errors) != 0 {
				t.Fatalf("form build returned errors: %v", result.Errors)
			}
			if result.Display != test.wantDisplay {
				t.Errorf("Display = %q, want %q", result.Display, test.wantDisplay)
			}
			if !reflect.DeepEqual(result.Args, test.wantArgs) {
				t.Errorf("Args = %#v, want %#v", result.Args, test.wantArgs)
			}
		})
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

func TestFormRegistersAndHandlesFieldClickRegions(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.openForm(testFormCommand())
	m.View()

	for _, field := range []int{0, 1, 2, 4} {
		if _, ok := regionForAction(m, "form:"+strconv.Itoa(field)); !ok {
			t.Fatalf("form field %d has no clickable region", field)
		}
	}
	updated, _ := m.handleRegionClick("form:2")
	got := updated.(model)
	if got.form.focusedField != 2 {
		t.Fatalf("clicked field focus = %d, want 2", got.form.focusedField)
	}

	m.form.focusedField = 0
	r, _ := regionForAction(m, "form:4")
	updated, _ = m.handleMouse(tea.MouseMsg{Type: tea.MouseLeft, X: r.x1, Y: r.y1})
	got = updated.(model)
	if got.form.focusedField != 4 {
		t.Fatalf("mouse click focus = %d, want 4", got.form.focusedField)
	}
}

func TestFormNavigationKeepsFocusedFieldVisible(t *testing.T) {
	m := NewModelForTest(80, 12)
	m.openForm(testFormCommand())
	m.formViewport.Width = 76
	m.formViewport.Height = 3
	m.View()
	if m.formViewport.TotalLineCount() <= m.formViewport.Height {
		t.Fatalf("form content did not overflow viewport: %d lines, height %d", m.formViewport.TotalLineCount(), m.formViewport.Height)
	}

	for i := 0; i < m.form.fieldCount()-1; i++ {
		updated, _ := m.handleFormKey(formKey(tea.KeyDown))
		m = updated.(model)
	}
	if m.form.focusedField != m.form.fieldCount()-1 {
		t.Fatalf("focused field = %d, want final field", m.form.focusedField)
	}
	if m.formViewport.YOffset == 0 {
		t.Fatal("viewport did not scroll to keep final field visible")
	}

	updated, _ := m.handleFormKey(formKey(tea.KeyUp))
	got := updated.(model)
	line := got.formFieldLines[got.form.focusedField]
	if line < got.formViewport.YOffset || line >= got.formViewport.YOffset+got.formViewport.Height {
		t.Fatalf("focused line %d is outside viewport [%d,%d)", line, got.formViewport.YOffset, got.formViewport.YOffset+got.formViewport.Height)
	}
}
