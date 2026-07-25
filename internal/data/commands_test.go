package data

import (
	"reflect"
	"strings"
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
	removedCommands := []string{"restart", "healthcheck", "diff", "commit", "rename", "update", "wait", "pause", "unpause", "port", "convert", "encrypt", "decrypt"}
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

	expectedCmds := map[string]bool{"list": false, "enter": false, "run": false, "shell": false, "terminate": false}
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

func TestSessionCommandsUseSystemNamespace(t *testing.T) {
	for _, command := range GetCommandsByCategory("Session") {
		wantPrefix := "wslc system session " + command.Name
		if !strings.HasPrefix(command.Full, wantPrefix) {
			t.Errorf("session command %q = %q, want prefix %q", command.Name, command.Full, wantPrefix)
		}
	}
}

func TestSessionListUsesSupportedVerboseOption(t *testing.T) {
	list := catalogCommand(t, "Session", "list")
	if len(list.Flags) != 1 || list.Flags[0].Long != "--verbose" {
		t.Fatalf("session list flags = %#v, want only --verbose", list.Flags)
	}
	if list.Schema == nil || len(list.Schema.Options) != 1 || list.Schema.Options[0].Flag != "--verbose" {
		t.Fatalf("session list schema options = %#v, want only --verbose", list.Schema)
	}
}

func TestSystemCategoryContainsSupportedCommands(t *testing.T) {
	commandsByCategory := GetCommandsByCategory("System")
	want := map[string]bool{"version": false}
	for _, command := range commandsByCategory {
		if _, ok := want[command.Name]; ok {
			want[command.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("system category missing command %q", name)
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

func TestInstalledWSLCCommandsAreCataloged(t *testing.T) {
	expected := map[string][]string{
		"Container": {"create", "start", "stop", "kill", "run", "exec", "attach", "ls", "inspect", "logs", "stats", "prune", "rm", "cp", "export"},
		"Image":     {"build", "import", "load", "pull", "push", "ls", "inspect", "rm", "prune"},
		"Network":   {"create", "rm", "prune", "ls", "inspect", "connect", "disconnect"},
		"Volume":    {"create", "rm", "prune", "ls", "inspect"},
		"Session":   {"list", "enter", "run", "shell", "terminate"},
		"System":    {"version"},
		"Registry":  {"login", "logout"},
	}

	for category, names := range expected {
		catalog := GetCommandsByCategory(category)
		found := make(map[string]bool, len(catalog))
		for _, command := range catalog {
			found[command.Name] = true
		}
		for _, name := range names {
			if !found[name] {
				t.Errorf("category %q missing WSLC reference command %q", category, name)
			}
		}
	}
}

func TestUnsupportedWSLCCommandsAreNotCataloged(t *testing.T) {
	unsupported := map[string][]string{
		"Image":    {"history"},
		"Session":  {"start", "stop", "attach"},
		"System":   {"info", "df", "prune", "events"},
		"Registry": {"list"},
	}

	for category, names := range unsupported {
		for _, command := range GetCommandsByCategory(category) {
			for _, name := range names {
				if command.Name == name {
					t.Errorf("category %q contains unsupported command %q", category, name)
				}
			}
		}
	}
}

func TestStatsSchemaUsesSupportedOptions(t *testing.T) {
	stats := catalogCommand(t, "Container", "stats")
	if stats.Schema == nil {
		t.Fatal("stats command has no schema")
	}

	if len(stats.Schema.Arguments) != 1 || !stats.Schema.Arguments[0].Repeatable || stats.Schema.Arguments[0].Required {
		t.Fatalf("stats should accept zero or more containers: %#v", stats.Schema.Arguments)
	}
	if got := stats.Schema.Options[0]; got.Flag != "--format" || got.Kind != commands.OptionKindSelect || !reflect.DeepEqual(got.Choices, []string{"table", "json"}) {
		t.Fatalf("stats format option = %#v", got)
	}
	for _, option := range stats.Schema.Options {
		if option.Flag == "--no-stream" {
			t.Fatal("stats schema contains unsupported --no-stream option")
		}
	}
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

func TestCatalogPickerMetadataIsExplicitAndSupported(t *testing.T) {
	validTypes := map[commands.ResourceType]bool{
		commands.ResourceTypeContainer: true,
		commands.ResourceTypeImage:     true,
		commands.ResourceTypeNetwork:   true,
		commands.ResourceTypeVolume:    true,
		commands.ResourceTypeSession:   true,
	}
	for _, command := range GetAllCommands() {
		for _, argument := range command.Schema.Arguments {
			if argument.PickerEnabled && (!validTypes[argument.ResourceType] || !argument.PickerAvailable()) {
				t.Errorf("%s argument %q has invalid picker metadata: %#v", command.Full, argument.Name, argument)
			}
		}
	}
}

func TestCatalogPickerMetadataRepresentativeFields(t *testing.T) {
	tests := []struct {
		category   string
		command    string
		argument   string
		resource   commands.ResourceType
		repeatable bool
		picker     bool
	}{
		{"Container", "exec", "container", commands.ResourceTypeContainer, false, true},
		{"Container", "start", "containers", commands.ResourceTypeContainer, true, true},
		{"Image", "inspect", "images", commands.ResourceTypeImage, true, true},
		{"Network", "connect", "network", commands.ResourceTypeNetwork, false, true},
		{"Volume", "inspect", "volumes", commands.ResourceTypeVolume, true, true},
		{"Session", "enter", "session", commands.ResourceTypeSession, false, true},
		{"Image", "pull", "image", "", false, false},
		{"Image", "build", "path", "", false, false},
		{"Session", "run", "command", "", true, false},
		{"Registry", "login", "server", "", false, false},
	}
	for _, test := range tests {
		command := catalogCommand(t, test.category, test.command)
		var found *commands.Argument
		for index := range command.Schema.Arguments {
			if command.Schema.Arguments[index].Name == test.argument {
				found = &command.Schema.Arguments[index]
				break
			}
		}
		if found == nil {
			t.Fatalf("%s/%s argument %q not found", test.category, test.command, test.argument)
		}
		if found.ResourceType != test.resource || found.Repeatable != test.repeatable || found.PickerEnabled != test.picker {
			t.Errorf("%s/%s argument %q = %#v", test.category, test.command, test.argument, *found)
		}
	}
}

func TestCatalogSchemasHaveDisplayMetadata(t *testing.T) {
	for _, command := range GetAllCommands() {
		if command.Schema == nil {
			continue
		}
		for _, argument := range command.Schema.Arguments {
			if strings.TrimSpace(argument.Label) == "" || strings.TrimSpace(argument.Placeholder) == "" {
				t.Errorf("%s argument %q lacks display metadata: %#v", command.Full, argument.Name, argument)
			}
		}
		for _, option := range command.Schema.Options {
			if strings.TrimSpace(option.Flag) == "" || strings.TrimSpace(option.Description) == "" {
				t.Errorf("%s option lacks display metadata: %#v", command.Full, option)
			}
		}
	}
}

func TestCatalogSchemaDefaultsPreserveLegacyCommands(t *testing.T) {
	for _, command := range GetAllCommands() {
		t.Run(command.Category+"/"+command.Name, func(t *testing.T) {
			if exception, ok := catalogMigrationExceptions[command.Category+"/"+command.Name]; ok {
				t.Skip(exception)
			}
			values := make(map[string]string)
			for _, placeholder := range commands.ExtractPlaceholders(command.Full) {
				values[placeholder] = "value-" + placeholder
			}
			expected := commands.ParseCommand(commands.SubstitutePlaceholders(command.Full, values))
			var rows [][]string
			for _, argument := range command.Schema.Arguments {
				value := values[argument.Placeholder]
				if value != "" {
					rows = append(rows, []string{value})
				}
			}
			result := commands.Build(legacyBase(command), *command.Schema, rows, nil)
			if len(result.Errors) != 0 {
				t.Fatalf("Build returned errors: %v", result.Errors)
			}
			if !reflect.DeepEqual(result.Args, expected) {
				t.Fatalf("schema args = %#v, legacy Full = %#v", result.Args, expected)
			}
		})
	}
}

func legacyBase(command commands.Command) []string {
	fields := strings.Fields(command.Full)
	for index, field := range fields {
		if strings.HasPrefix(field, "-") || strings.HasPrefix(field, "{") {
			return fields[:index]
		}
	}
	return fields
}

// These entries intentionally use legacy examples or aliases that are not a
// one-to-one representation of the schema's canonical invocation.
var catalogMigrationExceptions = map[string]string{
	"Container/ls":     "legacy Full selects JSON and --all; schema defaults are table and no --all",
	"Container/run":    "legacy Full is an example with -d and --name values not represented by defaults",
	"Container/create": "legacy Full is an example with a --name value not represented by defaults",
	"Container/exec":   "legacy Full supplies literal bash for the required repeatable command",
	"Container/stop":   "legacy Full omits the schema's default --time value",
	"Container/kill":   "legacy Full omits the schema's default --signal value",
	"Container/export": "legacy Full supplies an -o value while the schema has no output default",
	"Image/ls":         "legacy Full selects JSON while the schema defaults to table",
	"Image/save":       "legacy Full supplies an -o value while the schema has no output default",
	"Image/load":       "legacy Full supplies an -i value while the schema has no input default",
	"Image/build":      "legacy Full supplies a tag and literal path while the schema has no tag default",
	"Network/ls":       "legacy Full selects JSON while the schema defaults to table",
	"Network/create":   "legacy Full omits the schema's default bridge driver",
	"Volume/ls":        "legacy Full selects JSON while the schema defaults to table",
	"Session/run":      "legacy Full supplies only the session while the schema requires a command",
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

func TestListFormatAcceptsTemplatesAndWide(t *testing.T) {
	for _, test := range []struct {
		category string
		format   string
	}{
		{category: "Container", format: "{{json .}}"},
		{category: "Image", format: "wide"},
		{category: "Network", format: "{{.Name}}"},
		{category: "Volume", format: "wide"},
	} {
		command := catalogCommand(t, test.category, "ls")
		result := commands.Build(
			[]string{"wslc", strings.ToLower(test.category), "ls"},
			*command.Schema,
			nil,
			map[string]string{"--format": test.format},
		)
		if len(result.Errors) != 0 {
			t.Errorf("%s format %q returned errors: %v", test.category, test.format, result.Errors)
		}
		want := []string{"wslc", strings.ToLower(test.category), "ls", "--format", test.format}
		if !reflect.DeepEqual(result.Args, want) {
			t.Errorf("%s Args = %#v, want %#v", test.category, result.Args, want)
		}
	}
}

func TestStatsFormatRejectsUnsupportedValues(t *testing.T) {
	command := catalogCommand(t, "Container", "stats")
	result := commands.Build(
		[]string{"wslc", "stats"},
		*command.Schema,
		nil,
		map[string]string{"--format": "wide"},
	)
	if !containsCommandError(result.Errors, `option "--format" has invalid value "wide"`) {
		t.Fatalf("stats accepted unsupported format: %v", result.Errors)
	}
}

func TestStatsBuildsAllAcceptedFormatsAndFlags(t *testing.T) {
	command := catalogCommand(t, "Container", "stats")
	for _, format := range []string{"table", "json"} {
		result := commands.Build(
			[]string{"wslc", "stats"},
			*command.Schema,
			[][]string{{"web"}, {"worker"}},
			map[string]string{"--format": format, "--all": "true", "--no-trunc": "true"},
		)
		if len(result.Errors) != 0 {
			t.Fatalf("format %q returned errors: %v", format, result.Errors)
		}
		want := []string{"wslc", "stats", "web", "worker", "--format", format, "--all", "--no-trunc"}
		if !reflect.DeepEqual(result.Args, want) {
			t.Errorf("format %q Args = %#v, want %#v", format, result.Args, want)
		}
	}
}

func TestStatsBuildsWithZeroContainerRows(t *testing.T) {
	command := catalogCommand(t, "Container", "stats")
	result := commands.Build([]string{"wslc", "stats"}, *command.Schema, nil, nil)
	if len(result.Errors) != 0 {
		t.Fatalf("zero rows returned errors: %v", result.Errors)
	}
	if want := []string{"wslc", "stats", "--format", "table"}; !reflect.DeepEqual(result.Args, want) {
		t.Fatalf("zero rows Args = %#v, want %#v", result.Args, want)
	}
}

func TestVariableLengthStopAndRemoveBuildAllRows(t *testing.T) {
	for _, test := range []struct {
		name     string
		category string
		command  string
		base     []string
		want     []string
	}{
		{name: "stop", category: "Container", command: "stop", base: []string{"wslc", "stop"}, want: []string{"wslc", "stop", "one", "two", "--time", "10"}},
		{name: "remove", category: "Container", command: "rm", base: []string{"wslc", "remove"}, want: []string{"wslc", "remove", "one", "two"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			command := catalogCommand(t, test.category, test.command)
			result := commands.Build(test.base, *command.Schema, [][]string{{"one"}, {"two"}}, nil)
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

func containsCommandError(errors []error, want string) bool {
	for _, err := range errors {
		if err.Error() == want {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
