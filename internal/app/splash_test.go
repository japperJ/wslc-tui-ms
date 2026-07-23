package app

import (
	"strings"
	"testing"
	"wslc-tui-ms/internal/update"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelStartsWithAnimatedSplash(t *testing.T) {
	m := NewModel()

	if !m.splashActive {
		t.Fatal("new model should start with the splash animation active")
	}
	if len(splashFrames) < 8 {
		t.Fatalf("splash should have at least 8 frames, got %d", len(splashFrames))
	}
	if !strings.Contains(m.View(), "W S L C") {
		t.Fatalf("splash view should contain the WSLC logo, got:\n%s", m.View())
	}
	if !strings.Contains(m.View(), "NATIVE WSL CONTROL") {
		t.Fatalf("splash view should contain the enlarged logo, got:\n%s", m.View())
	}
	if !strings.Contains(strings.Join(splashFrames, "\n"), "Are we there") {
		t.Fatal("splash frames should include a humorous startup line")
	}
	if !strings.Contains(m.View(), "Press Enter to continue") {
		t.Fatal("splash view should tell the user how to continue")
	}
}

func TestSplashWaitsForEnter(t *testing.T) {
	m := NewModel()

	updated, _ := m.Update(splashTickMsg{})
	got := updated.(model)
	if !got.splashActive {
		t.Fatal("splash should remain active until Enter is pressed")
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = updated.(model)
	if got.splashActive {
		t.Fatal("splash should end when Enter is pressed")
	}
}

func TestSplashDismissalPreservesStartupUpdateView(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.splashActive = true
	m.updateDecision = &update.Decision{Available: true, Version: "v2.0.0"}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(model)
	if got.currentView != viewUpdate {
		t.Fatalf("startup update view was lost on splash dismissal: %v", got.currentView)
	}
}
