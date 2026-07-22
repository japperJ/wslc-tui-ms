package commands

import (
	"testing"
	"time"
)

func TestAutocomplete(t *testing.T) {
	cmds := []Command{
		{Full: "wslc container ps --all", Category: "Container"},
		{Full: "wslc container exec -it mycontainer bash", Category: "Container"},
		{Full: "wslc distro list", Category: "Distro"},
	}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty input returns all", "", 3},
		{"prefix match container", "wslc container", 2},
		{"fuzzy match ps", "ps", 1},
		{"fuzzy match exec", "exec", 1},
		{"no match", "xyz", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Autocomplete(tt.input, cmds)
			if len(results) != tt.expected {
				t.Errorf("Autocomplete(%q) returned %d results, expected %d", tt.input, len(results), tt.expected)
			}
		})
	}
}

func TestAutocompleteEmptyCommands(t *testing.T) {
	results := Autocomplete("anything", nil)
	if len(results) != 0 {
		t.Errorf("Autocomplete with nil commands returned %d results, expected 0", len(results))
	}
}

func TestFilterByPrefix(t *testing.T) {
	cmds := []Command{
		{Full: "wslc container ps --all", Category: "Container"},
		{Full: "wslc container exec -it mycontainer bash", Category: "Container"},
		{Full: "wslc distro list", Category: "Distro"},
	}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty input returns all", "", 3},
		{"prefix match container", "wslc container", 2},
		{"exact prefix match", "wslc container ps", 1},
		{"no prefix match", "ps", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := FilterByPrefix(tt.input, cmds)
			if len(results) != tt.expected {
				t.Errorf("FilterByPrefix(%q) returned %d results, expected %d", tt.input, len(results), tt.expected)
			}
		})
	}
}

func TestFilterByPrefixEmptyCommands(t *testing.T) {
	results := FilterByPrefix("anything", nil)
	if len(results) != 0 {
		t.Errorf("FilterByPrefix with nil commands returned %d results, expected 0", len(results))
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"simple command", "wslc container ps", []string{"wslc", "container", "ps"}},
		{"command with quotes", `wslc exec "my container" bash`, []string{"wslc", "exec", "my container", "bash"}},
		{"command with single quotes", "wslc exec 'my container' bash", []string{"wslc", "exec", "my container", "bash"}},
		{"empty command", "", nil},
		{"multiple spaces", "wslc   container   ps", []string{"wslc", "container", "ps"}},
		{"tabs as whitespace", "wslc\tcontainer\tps", []string{"wslc", "container", "ps"}},
		{"escaped double quote inside double quotes", `wslc exec "hello \"world\""`, []string{"wslc", "exec", `hello "world"`}},
		{"escaped single quote inside single quotes", `wslc exec 'it\'s'`, []string{"wslc", "exec", "it's"}},
		{"escaped backslash inside quotes", `wslc exec "path\\to\\file"`, []string{"wslc", "exec", `path\to\file`}},
		{"escaped backslash outside quotes", `wslc exec path\\to\\file`, []string{"wslc", "exec", `path\to\file`}},
		{"escaped space outside quotes", `wslc exec hello\ world`, []string{"wslc", "exec", "hello world"}},
		{"backslash before non-special char preserved", `wslc exec C:\path`, []string{"wslc", "exec", `C:\path`}},
		{"unclosed double quote consumes rest", `wslc exec "hello world`, []string{"wslc", "exec", "hello world"}},
		{"unclosed single quote consumes rest", `wslc exec 'hello world`, []string{"wslc", "exec", "hello world"}},
		{"only spaces", "   ", nil},
		{"only quotes", `""`, nil},
		{"mixed quotes", `wslc exec "hello" 'world'`, []string{"wslc", "exec", "hello", "world"}},
		{"backslash at end of input", `wslc\`, []string{"wslc\\"}},
		{"escaped closing quote keeps it literal", `wslc exec "\"`, []string{"wslc", "exec", `"`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCommand(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseCommand(%q) returned %d args, expected %d: got %v", tt.input, len(result), len(tt.expected), result)
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("ParseCommand(%q)[%d] = %q, expected %q", tt.input, i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestExecuteEmptyCommand(t *testing.T) {
	result := Execute("", 5*time.Second)
	if result.Error == nil {
		t.Error("Execute(\"\") should return error for empty command")
	}
	if result.ExitCode != 0 {
		t.Errorf("Execute(\"\") exitCode = %d, expected 0", result.ExitCode)
	}
}

func TestExecuteCommandNotFound(t *testing.T) {
	result := Execute("nonexistent_command_xyz_12345", 5*time.Second)
	if result.Error == nil {
		t.Error("Execute with nonexistent command should return error")
	}
	if result.ExitCode == 0 {
		t.Error("Execute with nonexistent command should have non-zero exit code")
	}
}

func TestExecuteSuccessfulCommand(t *testing.T) {
	result := Execute("cmd /c echo hello", 5*time.Second)
	if result.Error != nil {
		t.Errorf("Execute(\"cmd /c echo hello\") returned error: %v", result.Error)
	}
	if result.ExitCode != 0 {
		t.Errorf("Execute(\"cmd /c echo hello\") exitCode = %d, expected 0", result.ExitCode)
	}
}

func TestExecuteTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	result := Execute("ping -n 30 127.0.0.1", 200*time.Millisecond)
	if result.Error == nil {
		t.Error("Execute with timeout should return error")
	}
	if result.ExitCode == 0 {
		t.Error("Execute with timeout should have non-zero exit code")
	}
	if result.Duration >= 5*time.Second {
		t.Errorf("Execute with timeout took too long: %v", result.Duration)
	}
}
