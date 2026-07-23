package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
	"wslc-tui-ms/internal/buildinfo"
	"wslc-tui-ms/internal/settings"
	"wslc-tui-ms/internal/update"

	tea "github.com/charmbracelet/bubbletea"
)

type appUpdateClient struct{ err error }

func (c appUpdateClient) Releases(context.Context) ([]update.Release, error) { return nil, c.err }
func (c appUpdateClient) Policy(context.Context) (update.Policy, error) {
	return update.Policy{MinimumSupportedVersion: "0.0.0"}, nil
}
func (c appUpdateClient) Checksums(context.Context, update.Release) (map[string]update.Asset, error) {
	return nil, nil
}

func appWithDecision(t *testing.T) model {
	t.Helper()
	m := NewModelForTest(120, 30)
	m.updateService = update.Service{Store: settings.NewStore(filepath.Join(t.TempDir(), "settings.json")), CurrentVersion: "v1.0.0", Distribution: "portable", Now: func() time.Time { return time.Unix(10, 0) }}
	m.updateDecision = &update.Decision{Available: true, Version: "v2.0.0", Channel: "stable", Notes: "security fixes", Asset: update.Asset{Name: "payload.zip", URL: "https://example/payload", SHA256: "hash"}}
	m.currentView = viewUpdate
	return m
}

func TestUpdateConfirmationDoesNotStartDownload(t *testing.T) {
	m := appWithDecision(t)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(model)
	if cmd != nil || !m.updateConfirm {
		t.Fatal("download action should first require confirmation")
	}
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if cmd != nil || !m.updateHandoff {
		t.Fatal("confirmation should only create a handoff state")
	}
	if m.updateDecision.Asset.SHA256 != "hash" {
		t.Fatal("handoff lost checksum metadata")
	}
}

func TestUpdateLaterPersistsAndLeavesCommandBrowser(t *testing.T) {
	m := appWithDecision(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(model)
	if m.currentView != viewCommands || m.updateDecision != nil {
		t.Fatal("Later should dismiss a non-mandatory update")
	}
	state, err := m.updateService.Store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if state.Deferred != "v2.0.0" {
		t.Fatalf("deferred version=%q", state.Deferred)
	}
}

func TestMandatoryUpdateCannotBeDeferred(t *testing.T) {
	m := appWithDecision(t)
	m.updateDecision.Mandatory = true
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(model)
	if m.currentView != viewUpdate || m.updateDecision == nil {
		t.Fatal("mandatory update must remain visible")
	}
}

func TestAutomaticUpdateFailureIsNonBlocking(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.updateService.Client = appUpdateClient{err: errors.New("offline")}
	updated, cmd := m.Update(updateResultMsg{err: errors.New("offline")})
	m = updated.(model)
	if cmd != nil || m.currentView != viewCommands || m.updateChecking {
		t.Fatal("background failure should not change command usability")
	}
	if m.updateError == nil {
		t.Fatal("failure should remain observable as a status notice")
	}
}

func TestFocusedCommandSearchUStartsManualUpdateCheck(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.updateService = update.Service{Store: settings.NewStore(filepath.Join(t.TempDir(), "settings.json"))}
	inputBefore := m.inputValue

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(model)
	if cmd == nil || !m.updateChecking {
		t.Fatal("focused u should start a manual update check")
	}
	if m.inputValue != inputBefore {
		t.Fatalf("focused u changed search input to %q", m.inputValue)
	}
}

func TestInitialUpdateChannelUsesBetaBuildWhenUnconfigured(t *testing.T) {
	old := buildinfo.Channel
	buildinfo.Channel = "Beta"
	t.Cleanup(func() { buildinfo.Channel = old })

	if got := initialUpdateChannel(buildinfo.Channel, settings.Settings{}); got != update.Beta {
		t.Fatalf("initial channel = %q, want %q", got, update.Beta)
	}
}

func TestInitialUpdateChannelPreservesPersistedChoice(t *testing.T) {
	if got := initialUpdateChannel("Beta", settings.Settings{Channel: settings.Stable, ChannelSet: true}); got != update.Stable {
		t.Fatalf("persisted channel = %q, want %q", got, update.Stable)
	}
}
