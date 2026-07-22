package app

import (
	"reflect"
	"testing"

	"wslc-tui-ms/internal/commands"
)

func testFormCommand() commands.Command {
	return commands.Command{
		Full: "wslc container run",
		Schema: &commands.CommandSchema{
			Arguments: []commands.Argument{
				{Name: "image", Required: true},
				{Name: "command", Repeatable: true},
			},
			Options: []commands.Option{
				{Flag: "--detach", Kind: commands.OptionKindBoolean},
				{Flag: "--format", Kind: commands.OptionKindSelect, Default: "table", Choices: []string{"table", "json"}},
				{Flag: "--timeout", Kind: commands.OptionKindNumeric, Default: "10"},
			},
		},
	}
}

func TestFormInitializesRowsAndOptionsFromSchema(t *testing.T) {
	form := newFormState(testFormCommand(), nil)

	if got, want := form.argumentRows, [][]string{{""}, {""}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argument rows = %#v, want %#v", got, want)
	}
	if got, want := form.optionValues(), map[string]string{"--detach": "", "--format": "table", "--timeout": "10"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("option values = %#v, want %#v", got, want)
	}
	if got, want := form.optionFlags(), []string{"--detach", "--format", "--timeout"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("option order = %#v, want %#v", got, want)
	}
}

func TestFormRemembersOptionsPerCommand(t *testing.T) {
	m := NewModel()
	command := testFormCommand()
	m.openCommandForm(command)
	m.form.setOption("--format", "json")
	m.form.setOption("--detach", "true")
	m.rememberFormOptions()

	other := testFormCommand()
	other.Full = "wslc container ls"
	m.openCommandForm(other)
	if got := m.form.optionValue("--format"); got != "table" {
		t.Fatalf("other command inherited format %q", got)
	}

	m.openCommandForm(command)
	if got := m.form.optionValue("--format"); got != "json" {
		t.Fatalf("remembered format = %q, want json", got)
	}
	if got := m.form.optionValue("--detach"); got != "true" {
		t.Fatalf("remembered detach = %q, want true", got)
	}
}

func TestFormReopeningClearsPositionalValues(t *testing.T) {
	m := NewModel()
	command := testFormCommand()
	m.openCommandForm(command)
	m.form.argumentRows[0][0] = "ubuntu"
	m.form.argumentRows[1][0] = "echo"
	m.rememberFormOptions()
	m.openCommandForm(command)

	if got, want := m.form.argumentRows, [][]string{{""}, {""}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reopened rows = %#v, want %#v", got, want)
	}
}

func TestFormAddsAndRemovesRepeatableRows(t *testing.T) {
	form := newFormState(testFormCommand(), nil)
	form.argumentRows[1][0] = "echo"
	form.addRepeatableRow()
	form.argumentRows[2][0] = "hello"

	if !form.removeRepeatableRow(1) {
		t.Fatal("removeRepeatableRow returned false")
	}
	if got, want := form.argumentRows, [][]string{{""}, {"hello"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rows after remove = %#v, want %#v", got, want)
	}
	if form.removeRepeatableRow(0) {
		t.Fatal("removed non-repeatable row")
	}
}

func TestFormMovesFocusAcrossFields(t *testing.T) {
	form := newFormState(testFormCommand(), nil)
	if got := form.moveFocus(1); got != 1 {
		t.Fatalf("focus after next = %d, want 1", got)
	}
	if got := form.moveFocus(1); got != 2 {
		t.Fatalf("focus after next = %d, want 2", got)
	}
	if got := form.moveFocus(10); got != 4 {
		t.Fatalf("focus at end = %d, want 4", got)
	}
	if got := form.moveFocus(-10); got != 0 {
		t.Fatalf("focus at start = %d, want 0", got)
	}
}

func TestFormRetainsValidationError(t *testing.T) {
	form := newFormState(testFormCommand(), nil)
	result := form.build([]string{"wslc", "container", "run"})
	if len(result.Errors) == 0 {
		t.Fatal("expected required argument validation error")
	}
	form.moveFocus(1)
	if form.validationError == nil || form.validationError.Error() != result.Errors[0].Error() {
		t.Fatalf("validation error = %v, want %v", form.validationError, result.Errors[0])
	}
}

func TestFormStoresBuilderCommandState(t *testing.T) {
	form := newFormState(testFormCommand(), nil)
	form.argumentRows[0][0] = "ubuntu"
	form.argumentRows[1][0] = "echo"
	form.setOption("--format", "json")

	result := form.build([]string{"wslc", "container", "run"})
	if len(result.Errors) != 0 {
		t.Fatalf("build returned errors: %v", result.Errors)
	}
	want := []string{"wslc", "container", "run", "ubuntu", "echo", "--format", "json", "--timeout", "10"}
	if !reflect.DeepEqual(form.buildResult.Args, want) {
		t.Fatalf("stored command args = %#v, want %#v", form.buildResult.Args, want)
	}
}
