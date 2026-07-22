package commands

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildUsesDefaultsAndPreservesOrder(t *testing.T) {
	schema := CommandSchema{
		Arguments: []Argument{{Name: "container", Repeatable: true}},
		Options: []Option{
			{Flag: "--format", Kind: OptionKindSelect, Default: "table", Choices: []string{"table", "json"}},
			{Flag: "--all", Kind: OptionKindBoolean},
		},
	}

	result := Build([]string{"wslc", "stats"}, schema, [][]string{{"one"}, {"two"}}, map[string]string{})

	if len(result.Errors) != 0 {
		t.Fatalf("Build returned errors: %v", result.Errors)
	}
	if want := []string{"wslc", "stats", "one", "two", "--format", "table"}; !reflect.DeepEqual(result.Args, want) {
		t.Errorf("Args = %#v, want %#v", result.Args, want)
	}
	if result.Display != "wslc stats one two --format table" {
		t.Errorf("Display = %q", result.Display)
	}
}

func TestBuildSerializesOptionKindsAndOmitsEmptyValues(t *testing.T) {
	schema := CommandSchema{Options: []Option{
		{Flag: "--all", Kind: OptionKindBoolean},
		{Flag: "--name", Kind: OptionKindText},
		{Flag: "--format", Kind: OptionKindSelect, Choices: []string{"table", "json"}},
		{Flag: "--timeout", Kind: OptionKindNumeric},
		{Flag: "--empty", Kind: OptionKindText},
	}}

	result := Build([]string{"wslc", "run"}, schema, nil, map[string]string{
		"--all":     "true",
		"--name":    "hello world",
		"--format":  "json",
		"--timeout": "30",
		"--empty":   "",
	})

	if len(result.Errors) != 0 {
		t.Fatalf("Build returned errors: %v", result.Errors)
	}
	want := []string{"wslc", "run", "--all", "--name", "hello world", "--format", "json", "--timeout", "30"}
	if !reflect.DeepEqual(result.Args, want) {
		t.Errorf("Args = %#v, want %#v", result.Args, want)
	}
	if !strings.Contains(result.Display, `"hello world"`) {
		t.Errorf("Display does not quote text value: %q", result.Display)
	}
}

func TestBuildRejectsInvalidValues(t *testing.T) {
	schema := CommandSchema{
		Arguments: []Argument{
			{Name: "source", Required: true},
			{Name: "target", Required: true},
			{Name: "container", Repeatable: true},
		},
		Options: []Option{
			{Flag: "--format", Kind: OptionKindSelect, Choices: []string{"table", "json"}},
			{Flag: "--timeout", Kind: OptionKindNumeric, Validation: Validation{Min: 1, Max: 60}},
		},
	}

	result := Build([]string{"wslc", "tag"}, schema, [][]string{{"source"}, {}}, map[string]string{
		"--format":  "yaml",
		"--timeout": "not-a-number",
	})

	if len(result.Errors) != 3 {
		t.Fatalf("got %d errors, want 3: %v", len(result.Errors), result.Errors)
	}
}

func TestBuildRejectsIncompleteRows(t *testing.T) {
	schema := CommandSchema{Arguments: []Argument{{Name: "container", Repeatable: true}}}
	result := Build([]string{"wslc", "stop"}, schema, [][]string{{"one", "extra"}, {}}, nil)

	if len(result.Errors) != 2 {
		t.Fatalf("got %d errors, want 2: %v", len(result.Errors), result.Errors)
	}
}

func TestBuildOmitsEmptyOptionalArgument(t *testing.T) {
	schema := CommandSchema{Arguments: []Argument{{Name: "optional"}}}
	result := Build([]string{"wslc", "list"}, schema, [][]string{{}}, nil)

	if len(result.Errors) != 0 {
		t.Fatalf("Build returned errors: %v", result.Errors)
	}
	if !reflect.DeepEqual(result.Args, []string{"wslc", "list"}) {
		t.Errorf("Args = %#v", result.Args)
	}
}

func TestBuildRejectsRequiredRepeatableArgumentWithoutRows(t *testing.T) {
	schema := CommandSchema{Arguments: []Argument{{Name: "container", Required: true, Repeatable: true}}}
	result := Build([]string{"wslc", "stop"}, schema, nil, nil)

	if len(result.Errors) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(result.Errors), result.Errors)
	}
}

func TestBuildRejectsMissingOrFalseRequiredBoolean(t *testing.T) {
	schema := CommandSchema{Options: []Option{{Flag: "--all", Kind: OptionKindBoolean, Required: true, Default: "true"}}}

	for _, values := range []map[string]string{nil, {"--all": "false"}} {
		result := Build([]string{"wslc", "list"}, schema, nil, values)
		if len(result.Errors) != 1 {
			t.Errorf("values %#v: got %d errors, want 1: %v", values, len(result.Errors), result.Errors)
		}
	}
}

func TestBuildAcceptsExplicitTrueRequiredBoolean(t *testing.T) {
	schema := CommandSchema{Options: []Option{{Flag: "--all", Kind: OptionKindBoolean, Required: true}}}
	result := Build([]string{"wslc", "list"}, schema, nil, map[string]string{"--all": "true"})

	if len(result.Errors) != 0 {
		t.Fatalf("Build returned errors: %v", result.Errors)
	}
	if !reflect.DeepEqual(result.Args, []string{"wslc", "list", "--all"}) {
		t.Errorf("Args = %#v", result.Args)
	}
}

func TestBuildKeepsExecutableArgumentsSeparateFromDisplay(t *testing.T) {
	schema := CommandSchema{Arguments: []Argument{{Name: "value", Required: true}}}
	result := Build([]string{"wslc", "echo"}, schema, [][]string{{`a"b`}}, nil)

	if len(result.Errors) != 0 {
		t.Fatalf("Build returned errors: %v", result.Errors)
	}
	if result.Args[2] != `a"b` {
		t.Errorf("executable value = %q", result.Args[2])
	}
	if result.Display == strings.Join(result.Args, " ") {
		t.Errorf("display command was not quoted separately: %q", result.Display)
	}
}
