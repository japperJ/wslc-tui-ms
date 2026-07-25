package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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

type browserUpdateClient struct{}

func (browserUpdateClient) Releases(context.Context) ([]update.Release, error) {
	return []update.Release{{
		Tag:        "v2.0.0-beta.1",
		Prerelease: true,
		Assets:     []update.Asset{{Name: "wslc-tui-v2.0.0-beta.1-windows-amd64-portable.zip", URL: "https://example/payload"}},
	}}, nil
}
func (browserUpdateClient) Policy(context.Context) (update.Policy, error) {
	return update.Policy{MinimumSupportedVersion: "0.0.0"}, nil
}
func (browserUpdateClient) Checksums(context.Context, update.Release) (map[string]update.Asset, error) {
	return map[string]update.Asset{"wslc-tui-v2.0.0-beta.1-windows-amd64-portable.zip": {SHA256: "hash"}}, nil
}

func appWithDecision(t *testing.T) model {
	t.Helper()
	m := NewModelForTest(120, 30)
	m.updateService = update.Service{Store: settings.NewStore(filepath.Join(t.TempDir(), "settings.json")), CurrentVersion: "v1.0.0", Distribution: "portable", Now: func() time.Time { return time.Unix(10, 0) }}
	m.updateChannel = update.Stable
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

func TestManualUpdateCheckLeavesReadableCompletionStatus(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.updateChecking = true
	updated, cmd := m.Update(updateResultMsg{manual: true})
	m = updated.(model)
	if cmd != nil || m.updateChecking {
		t.Fatal("manual check should finish without another command")
	}
	if m.updateStatus != "No newer update found." {
		t.Fatalf("completion status = %q", m.updateStatus)
	}
}

func TestFocusedCommandSearchAcceptsLowercaseUpdateShortcutCharacter(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.updateService = update.Service{Store: settings.NewStore(filepath.Join(t.TempDir(), "settings.json"))}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(model)
	if cmd != nil || m.updateChecking {
		t.Fatal("focused u should remain text input")
	}
	if m.inputValue != "u" || m.textInput.Value() != "u" {
		t.Fatalf("focused u was not entered into search input: input=%q text=%q", m.inputValue, m.textInput.Value())
	}
}

func TestFocusedCommandSearchAcceptsUppercaseUpdateShortcutCharacter(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.updateService = update.Service{Store: settings.NewStore(filepath.Join(t.TempDir(), "settings.json"))}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'U'}})
	m = updated.(model)
	if cmd != nil || m.updateChecking {
		t.Fatal("focused U should remain text input")
	}
	if m.inputValue != "U" || m.textInput.Value() != "U" {
		t.Fatalf("focused U was not entered into search input: input=%q text=%q", m.inputValue, m.textInput.Value())
	}
}

func TestFocusedCommandFooterShowsUpdateShortcut(t *testing.T) {
	m := NewModelForTest(120, 30)
	if !strings.Contains(m.View(), "u Updates") {
		t.Fatalf("focused command footer omitted update shortcut:\n%s", m.View())
	}
}

func TestHelpIncludesUpdateShortcut(t *testing.T) {
	m := NewModelForTest(120, 30)
	if !strings.Contains(m.getHelpTooltip(), "u         Check for updates") {
		t.Fatalf("help omitted update shortcut:\n%s", m.getHelpTooltip())
	}
}

func TestCategorySidebarUsesKeyboardShortcutNumbers(t *testing.T) {
	m := NewModelForTest(120, 30)
	sidebar := stripAnsi(m.renderSidebar(32))
	for i, category := range m.categories {
		want := fmt.Sprintf("%d  %s", i+1, category)
		if !strings.Contains(sidebar, want) {
			t.Fatalf("sidebar omitted shortcut %q:\n%s", want, sidebar)
		}
	}
}

func TestHeaderShowsBuildVersion(t *testing.T) {
	old := buildinfo.Version
	buildinfo.Version = "v9.9.9-test"
	t.Cleanup(func() { buildinfo.Version = old })
	m := NewModelForTest(120, 30)
	if !strings.Contains(m.renderHeader(), "v9.9.9-test") {
		t.Fatalf("header omitted build version: %s", m.renderHeader())
	}
}

func TestCommandBrowserChannelShortcutTogglesAndFindsBetaPrerelease(t *testing.T) {
	m := NewModelForTest(120, 30)
	store := settings.NewStore(filepath.Join(t.TempDir(), "settings.json"))
	if err := store.Save(settings.Settings{Channel: settings.Stable}); err != nil {
		t.Fatal(err)
	}
	m.updateService = update.Service{
		Client:         browserUpdateClient{},
		Store:          store,
		CurrentVersion: "v1.0.0",
		Distribution:   "portable",
	}
	m.updateChannel = update.Stable
	m.inputFocused = false
	m.textInput.Blur()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(model)
	if cmd == nil || m.updateChannel != update.Beta {
		t.Fatalf("channel shortcut did not start beta check: channel=%q cmd=%v", m.updateChannel, cmd != nil)
	}
	if !strings.Contains(m.channelStatus, "Beta") || !strings.Contains(m.renderCommandsView(), "Beta") {
		t.Fatalf("active channel status is not visible: %q", m.channelStatus)
	}
	state, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if state.Channel != settings.Beta {
		t.Fatalf("persisted channel=%q, want %q", state.Channel, settings.Beta)
	}

	updated, _ = m.Update(cmd())
	m = updated.(model)
	if m.currentView != viewUpdate || m.updateDecision == nil || m.updateDecision.Version != "v2.0.0-beta.1" {
		t.Fatalf("beta-only prerelease was not surfaced: view=%v decision=%+v", m.currentView, m.updateDecision)
	}
}

func TestFocusedCommandSearchAcceptsChannelShortcutCharacters(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.updateService = update.Service{Store: settings.NewStore(filepath.Join(t.TempDir(), "settings.json"))}
	m.updateChannel = update.Beta
	m.inputFocused = true
	m.textInput.Focus()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(model)
	if cmd != nil || m.updateChannel != update.Beta {
		t.Fatalf("focused c triggered channel action: channel=%q cmd=%v", m.updateChannel, cmd != nil)
	}
	if m.inputValue != "c" || m.textInput.Value() != "c" {
		t.Fatalf("focused c was not entered into search input: input=%q text=%q", m.inputValue, m.textInput.Value())
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})
	m = updated.(model)
	if cmd != nil || m.updateChannel != update.Beta {
		t.Fatalf("focused C triggered channel action: channel=%q cmd=%v", m.updateChannel, cmd != nil)
	}
	if m.inputValue != "cC" || m.textInput.Value() != "cC" {
		t.Fatalf("focused C was not entered into search input: input=%q text=%q", m.inputValue, m.textInput.Value())
	}
}

func TestUpdateChannelAcceptsUppercaseC(t *testing.T) {
	m := appWithDecision(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})
	m = updated.(model)
	if m.updateChannel != update.Beta {
		t.Fatalf("uppercase C channel = %q, want %q", m.updateChannel, update.Beta)
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
