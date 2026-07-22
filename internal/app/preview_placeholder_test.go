package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"wslc-tui-ms/internal/commands"
)

func enterKey() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }

// Preview of a command with a {name} placeholder must register a ph:0 click
// region on its VALUES row after rendering.
func TestPreviewRegistersPlaceholderRegion(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.enterPreview(commands.Command{
		Name:        "run",
		Full:        "wslc container run -d --name {name} {image}",
		Description: "Run a container",
	})
	if len(m.placeholders) != 2 {
		t.Fatalf("expected 2 placeholders, got %d (%v)", len(m.placeholders), m.placeholders)
	}

	_ = m.View() // rendering registers ph: regions

	for _, action := range []string{"ph:0", "ph:1"} {
		if _, ok := regionForAction(m, action); !ok {
			t.Errorf("region %q not registered after preview render", action)
		}
	}
}

// A no-placeholder command executes immediately on Enter, switching to the
// output view with the original command unchanged.
func TestNoPlaceholderEnterExecutesDirectly(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.enterPreview(commands.Command{
		Name: "ls",
		Full: "wslc container ls",
	})
	if len(m.placeholders) != 0 {
		t.Fatalf("expected 0 placeholders, got %v", m.placeholders)
	}

	updated, _ := m.handlePreviewKey(enterKey())
	got := updated.(model)
	if got.currentView != viewOutput {
		t.Errorf("expected viewOutput after Enter, got %v", got.currentView)
	}
	if got.outputCmd != "wslc container ls" {
		t.Errorf("expected unchanged command, got %q", got.outputCmd)
	}
}

func TestYellowCommandRequiresConfirmation(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.enterPreview(commands.Command{
		Name:       "run",
		Full:       "wslc run ubuntu:latest",
		Difficulty: "intermediate",
	})

	updated, _ := m.handlePreviewKey(enterKey())
	got := updated.(model)
	if got.currentView != viewConfirm {
		t.Fatalf("expected confirmation view, got %v", got.currentView)
	}
	if got.pendingCommand != "wslc run ubuntu:latest" {
		t.Errorf("expected pending command, got %q", got.pendingCommand)
	}
	if got.running {
		t.Error("yellow command started before confirmation")
	}
}

func TestRedPlaceholderCommandRequiresConfirmation(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.enterPreview(commands.Command{
		Name:       "kill",
		Full:       "wslc kill {name}",
		Difficulty: "advanced",
	})
	m.phValues["name"] = "web"

	updated, _ := m.handlePreviewKey(enterKey())
	got := updated.(model)
	if got.currentView != viewConfirm {
		t.Fatalf("expected confirmation view, got %v", got.currentView)
	}
	if got.pendingCommand != "wslc kill web" {
		t.Errorf("expected substituted pending command, got %q", got.pendingCommand)
	}
}

func TestConfirmationEnterExecutesPendingCommand(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.currentView = viewConfirm
	m.pendingCommand = "wslc version"
	m.pendingDifficulty = "intermediate"

	updated, _ := m.handleConfirmationKey(enterKey())
	got := updated.(model)
	if got.currentView != viewOutput {
		t.Fatalf("expected output view after confirmation, got %v", got.currentView)
	}
	if got.outputCmd != "wslc version" {
		t.Errorf("expected pending command to execute, got %q", got.outputCmd)
	}
}

func TestConfirmationEscCancelsPendingCommand(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.currentView = viewConfirm
	m.pendingCommand = "wslc system prune"
	m.pendingDifficulty = "advanced"

	updated, _ := m.handleConfirmationKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(model)
	if got.currentView != viewPreview {
		t.Fatalf("expected preview view after cancellation, got %v", got.currentView)
	}
	if got.pendingCommand != "" {
		t.Errorf("expected pending command cleared, got %q", got.pendingCommand)
	}
}

// Enter with an empty placeholder must be blocked (stay in preview, warn, and
// focus the first empty field) rather than executing.
func TestEmptyPlaceholderEnterBlocks(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.enterPreview(commands.Command{
		Name: "run",
		Full: "wslc container run {image}",
	})

	updated, _ := m.handlePreviewKey(enterKey())
	got := updated.(model)
	if got.currentView != viewPreview {
		t.Errorf("expected to stay in viewPreview, got %v", got.currentView)
	}
	if !got.phWarn {
		t.Error("expected phWarn to be set")
	}
	if got.phActiveIndex != 0 {
		t.Errorf("expected first empty field focused (0), got %d", got.phActiveIndex)
	}
}

// clicking a VALUES row focuses that placeholder field.
func TestPlaceholderRegionClickFocuses(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.enterPreview(commands.Command{
		Name: "run",
		Full: "wslc container run {image}",
	})
	_ = m.View()

	updated, _ := m.handleRegionClick("ph:0")
	got := updated.(model)
	if got.phActiveIndex != 0 {
		t.Errorf("expected field 0 focused after ph:0 click, got %d", got.phActiveIndex)
	}
}
