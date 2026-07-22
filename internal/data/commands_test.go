package data

import (
	"reflect"
	"testing"

	"wslc-tui-ms/internal/commands"
)

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

func TestEveryCatalogCommandHasSchema(t *testing.T) {
	for _, command := range GetAllCommands() {
		if command.Schema == nil {
			t.Errorf("%s has no schema", command.Full)
		}
	}
}

func TestCatalogSchemasHaveUniqueOrderedArguments(t *testing.T) {
	for _, command := range GetAllCommands() {
		if command.Schema == nil {
			continue
		}
		seen := make(map[string]bool)
		for index, argument := range command.Schema.Arguments {
			if argument.Name == "" {
				t.Errorf("%s has an unnamed argument", command.Full)
			}
			if seen[argument.Name] {
				t.Errorf("%s repeats argument name %q", command.Full, argument.Name)
			}
			seen[argument.Name] = true
			if argument.Repeatable && index != len(command.Schema.Arguments)-1 {
				t.Errorf("%s has non-terminal repeatable argument %q", command.Full, argument.Name)
			}
		}
	}
}

func TestCatalogRepresentativeCommandGeneration(t *testing.T) {
	tests := []struct {
		name    string
		command commands.Command
		rows    [][]string
		options map[string]string
		want    []string
	}{
		{
			name:    "stats defaults",
			command: catalogCommand(t, "Container", "stats"),
			rows:    [][]string{{"one"}, {"two"}},
			want:    []string{"wslc", "stats", "one", "two", "--format", "table"},
		},
		{
			name:    "tag positional order",
			command: catalogCommand(t, "Image", "tag"),
			rows:    [][]string{{"source:latest"}, {"target:v1"}},
			want:    []string{"wslc", "tag", "source:latest", "target:v1"},
		},
		{
			name:    "network connect positional order",
			command: catalogCommand(t, "Network", "connect"),
			rows:    [][]string{{"net"}, {"container"}},
			want:    []string{"wslc", "network", "connect", "net", "container"},
		},
		{
			name:    "stop timeout default",
			command: catalogCommand(t, "Container", "stop"),
			rows:    [][]string{{"one"}},
			want:    []string{"wslc", "stop", "one", "--time", "10"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := commands.Build(commandTokens(test.command), *test.command.Schema, test.rows, test.options)
			if len(result.Errors) != 0 {
				t.Fatalf("Build returned errors: %v", result.Errors)
			}
			if !reflect.DeepEqual(result.Args, test.want) {
				t.Errorf("Args = %#v, want %#v", result.Args, test.want)
			}
		})
	}
}

func catalogCommand(t *testing.T, category, name string) commands.Command {
	t.Helper()
	for _, command := range GetCommandsByCategory(category) {
		if command.Name == name {
			return command
		}
	}
	t.Fatalf("catalog command %s/%s not found", category, name)
	return commands.Command{}
}

func commandTokens(command commands.Command) []string {
	if command.Name == "connect" {
		return []string{"wslc", "network", "connect"}
	}
	switch command.Category {
	case "Container":
		return []string{"wslc", command.Name}
	case "Image":
		if command.Name == "tag" {
			return []string{"wslc", "tag"}
		}
		return []string{"wslc", command.Name}
	default:
		return []string{"wslc", command.Name}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
