package data

import "testing"

func TestGetCategoriesCount(t *testing.T) {
	cats := GetCategories()
	if len(cats) != 7 {
		t.Errorf("expected 7 categories, got %d: %v", len(cats), cats)
	}
}

func TestAllCategoriesHaveCommands(t *testing.T) {
	for _, cat := range GetCategories() {
		cmds := GetCommandsByCategory(cat)
		if len(cmds) == 0 {
			t.Errorf("category %q has no commands", cat)
		}
	}
}

func TestGetAllCommandsNonEmpty(t *testing.T) {
	all := GetAllCommands()
	if len(all) == 0 {
		t.Fatal("GetAllCommands returned empty")
	}
	if len(all) < 30 {
		t.Errorf("expected at least 30 total commands, got %d", len(all))
	}
}

func TestNoNerdctlSpecificCommands(t *testing.T) {
	removedCommands := []string{"restart", "healthcheck", "diff", "commit", "rename", "update", "wait", "pause", "unpause", "port", "history", "convert", "encrypt", "decrypt"}
	removedCategories := []string{"Builder", "Namespace", "Compose"}

	all := GetAllCommands()

	for _, cmd := range all {
		for _, removed := range removedCommands {
			if cmd.Name == removed {
				t.Errorf("found removed command %q in catalog: %s", removed, cmd.Full)
			}
		}
	}

	for _, cat := range GetCategories() {
		for _, removedCat := range removedCategories {
			if cat == removedCat {
				t.Errorf("found removed category %q in catalog", removedCat)
			}
		}
	}
}

func TestSessionCategoryExists(t *testing.T) {
	cats := GetCategories()
	found := false
	for _, cat := range cats {
		if cat == "Session" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Session category not found (Microsoft WSLC-specific)")
	}

	sessionCmds := GetCommandsByCategory("Session")
	if len(sessionCmds) == 0 {
		t.Error("Session category has no commands")
	}

	expectedCmds := map[string]bool{"ls": false, "enter": false, "run": false, "shell": false, "terminate": false}
	for _, cmd := range sessionCmds {
		if _, ok := expectedCmds[cmd.Name]; ok {
			expectedCmds[cmd.Name] = true
		}
	}
	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Session category missing command %q", name)
		}
	}
}

func TestGPUFlagOnRun(t *testing.T) {
	runCmds := GetCommandsByCategory("Container")
	for _, cmd := range runCmds {
		if cmd.Name == "run" {
			found := false
			for _, flag := range cmd.Flags {
				if flag.Long == "--gpus" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Container run command missing --gpus flag (Microsoft WSLC feature)")
			}
			return
		}
	}
	t.Error("Container run command not found")
}

func TestStatsUsesMicrosoftWSLCFlags(t *testing.T) {
	for _, cmd := range GetCommandsByCategory("Container") {
		if cmd.Name != "stats" {
			continue
		}

		if cmd.Full != "wslc stats --format table" {
			t.Errorf("stats command has incorrect invocation: %q", cmd.Full)
		}

		for _, flag := range cmd.Flags {
			if flag.Long == "--no-stream" {
				t.Error("stats command contains unsupported --no-stream flag")
			}
		}
		return
	}

	t.Fatal("stats command not found")
}

func TestReadOnlyCommandsAreBeginner(t *testing.T) {
	expected := map[string]bool{
		"wslc inspect {name}":         false,
		"wslc tag {source} {target}":  false,
		"wslc network inspect {name}": false,
		"wslc volume inspect {name}":  false,
	}

	for _, cmd := range GetAllCommands() {
		if _, ok := expected[cmd.Full]; ok {
			expected[cmd.Full] = cmd.Difficulty == "beginner"
		}
	}
	for full, beginner := range expected {
		if !beginner {
			t.Errorf("read-only command should be green/beginner: %s", full)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
