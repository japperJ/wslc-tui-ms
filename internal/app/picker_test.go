package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"wslc-tui-ms/internal/commands"
)

type pickerDiscovery struct {
	values []string
	err    error
	calls  int
}

func (d *pickerDiscovery) Discover(context.Context, commands.ResourceType) ([]string, error) {
	d.calls++
	return append([]string(nil), d.values...), d.err
}

func pickerCommand(repeatable bool) commands.Command {
	return commands.Command{
		Full: "wslc container start",
		Schema: &commands.CommandSchema{Arguments: []commands.Argument{{
			Name: "containers", Required: true, Repeatable: repeatable,
			ResourceType: commands.ResourceTypeContainer, PickerEnabled: true,
		}}},
	}
}

func TestPickerOpensRefreshesAndCommitsScalarSelection(t *testing.T) {
	discovery := &pickerDiscovery{values: []string{"web", "worker"}}
	m := NewModelForTest(120, 30)
	m.discovery = discovery
	m.openForm(pickerCommand(false))

	updated, refresh := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(model)
	if m.currentView != viewPicker || !m.pickerLoading || refresh == nil {
		t.Fatalf("picker did not open asynchronously: view=%v loading=%v cmd=%v", m.currentView, m.pickerLoading, refresh != nil)
	}
	updated, _ = m.Update(refresh())
	m = updated.(model)
	if m.pickerLoading || discovery.calls != 1 {
		t.Fatalf("refresh state: loading=%v calls=%d", m.pickerLoading, discovery.calls)
	}

	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.currentView != viewForm || m.form.argumentRows[0][0] != "web" {
		t.Fatalf("scalar commit: view=%v rows=%#v", m.currentView, m.form.argumentRows)
	}
	if !m.form.pickerRows[0] {
		t.Fatal("scalar commit was not marked as picker-originated")
	}
}

func TestPickerFiltersAndPreservesOrderedRepeatableSelection(t *testing.T) {
	discovery := &pickerDiscovery{values: []string{"web", "worker", "db"}}
	m := NewModelForTest(120, 30)
	m.discovery = discovery
	command := pickerCommand(true)
	m.openForm(command)
	m.form.addRepeatableRow()

	updated, refresh := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(model)
	updated, _ = m.Update(refresh())
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(model)
	if got := m.filteredPickerCandidates(); !reflect.DeepEqual(got, []string{"db"}) {
		t.Fatalf("filtered candidates=%#v", got)
	}
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	// Reset the filter and selection for the ordered multi-select assertion.
	m.pickerFilter = ""
	m.formInput.SetValue("")
	m.pickerSelected = nil
	m.pickerIndex = 0
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if got, want := m.form.argumentRows, [][]string{{"web"}, {"worker"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("repeatable rows=%#v, want %#v", got, want)
	}
}

func TestPickerFailureLeavesFreeTextEditable(t *testing.T) {
	discovery := &pickerDiscovery{err: errors.New("offline")}
	m := NewModelForTest(120, 30)
	m.discovery = discovery
	m.openForm(pickerCommand(false))
	updated, refresh := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(model)
	updated, _ = m.Update(refresh())
	m = updated.(model)
	if !m.pickerDisabled || m.pickerError == nil {
		t.Fatal("failed discovery did not disable picker selection")
	}
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	updated, _ = m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("typed-value")})
	m = updated.(model)
	if m.form.argumentRows[0][0] != "typed-value" {
		t.Fatalf("typed fallback=%q", m.form.argumentRows[0][0])
	}
	updated, _ = m.handleFormKey(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(model).currentView != viewOutput {
		t.Fatal("typed fallback was blocked after discovery failure")
	}
}

func TestStalePickerSelectionBlockedBeforeExecution(t *testing.T) {
	discovery := &pickerDiscovery{values: []string{"web"}}
	m := NewModelForTest(120, 30)
	m.discovery = discovery
	m.openForm(pickerCommand(false))
	updated, refresh := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(model)
	updated, _ = m.Update(refresh())
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	discovery.values = []string{"worker"}
	updated, _ = m.handleFormKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.currentView != viewForm || m.form.validationError == nil {
		t.Fatal("stale picker value was not rejected")
	}
}

func TestStalePickerSelectionBlockedWhenConfirmationIsAccepted(t *testing.T) {
	discovery := &pickerDiscovery{values: []string{"web"}}
	m := NewModelForTest(120, 30)
	m.discovery = discovery
	command := pickerCommand(false)
	command.Difficulty = "advanced"
	m.openForm(command)
	updated, refresh := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(model)
	updated, _ = m.Update(refresh())
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	updated, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.handleFormKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.currentView != viewConfirm {
		t.Fatalf("view=%v, want confirmation", m.currentView)
	}
	discovery.values = []string{"worker"}
	updated, _ = m.handleConfirmationKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.currentView != viewForm || m.form.validationError == nil {
		t.Fatal("stale picker value was not rejected at confirmation")
	}
}

func TestEditingPickerValueConvertsOnlyEditedRowToFreeText(t *testing.T) {
	discovery := &pickerDiscovery{values: []string{"web", "worker"}}
	m := NewModelForTest(120, 30)
	m.discovery = discovery
	m.openForm(pickerCommand(true))
	m.form.addRepeatableRow()
	m.form.setPickerValues(0, []string{"web", "worker"})
	m.form.focusedField = 0
	m.syncFormInput()

	updated, _ := m.handleFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updated.(model)

	if m.form.pickerRows[0] {
		t.Fatal("edited row remained marked as picker-originated")
	}
	if !m.form.pickerRows[1] {
		t.Fatal("un-edited row lost picker-origin marker")
	}
}
