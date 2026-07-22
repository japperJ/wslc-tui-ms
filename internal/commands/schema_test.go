package commands

import "testing"

func TestCommandSchemaPreservesOrderedArguments(t *testing.T) {
	schema := CommandSchema{
		Arguments: []Argument{
			{
				Name:        "source",
				Label:       "Source image",
				Required:    true,
				Repeatable:  false,
				Placeholder: "image:tag",
				Validation:  Validation{Pattern: `^[a-z0-9/:.-]+$`},
			},
			{
				Name:       "target",
				Label:      "Target image",
				Repeatable: true,
				Validation: Validation{MinLength: 1, MaxLength: 64},
			},
		},
	}

	if len(schema.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(schema.Arguments))
	}
	if schema.Arguments[0].Name != "source" || schema.Arguments[1].Name != "target" {
		t.Fatalf("arguments are not ordered: %#v", schema.Arguments)
	}
	if !schema.Arguments[0].Required || schema.Arguments[0].Repeatable {
		t.Fatalf("source metadata was not preserved: %#v", schema.Arguments[0])
	}
	if !schema.Arguments[1].Repeatable || schema.Arguments[1].Validation.MaxLength != 64 {
		t.Fatalf("target metadata was not preserved: %#v", schema.Arguments[1])
	}
}

func TestCommandSchemaPreservesOrderedOptions(t *testing.T) {
	schema := CommandSchema{
		Options: []Option{
			{Flag: "--all", Description: "Show all", Kind: OptionKindBoolean},
			{Flag: "--format", Kind: OptionKindSelect, Default: "table", Choices: []string{"table", "json"}},
			{Flag: "--name", Kind: OptionKindText, Required: true, Validation: Validation{MinLength: 1}},
			{Flag: "--timeout", Kind: OptionKindNumeric, Validation: Validation{Min: 0, Max: 300}},
		},
	}

	if len(schema.Options) != 4 {
		t.Fatalf("expected 4 options, got %d", len(schema.Options))
	}
	if schema.Options[0].Flag != "--all" || schema.Options[1].Flag != "--format" ||
		schema.Options[2].Flag != "--name" || schema.Options[3].Flag != "--timeout" {
		t.Fatalf("options are not ordered: %#v", schema.Options)
	}
	if schema.Options[1].Default != "table" || len(schema.Options[1].Choices) != 2 {
		t.Fatalf("select metadata was not preserved: %#v", schema.Options[1])
	}
	if !schema.Options[2].Required || schema.Options[2].Validation.MinLength != 1 {
		t.Fatalf("required text metadata was not preserved: %#v", schema.Options[2])
	}
	if schema.Options[3].Kind != OptionKindNumeric || schema.Options[3].Validation.Max != 300 {
		t.Fatalf("numeric metadata was not preserved: %#v", schema.Options[3])
	}
}

func TestCommandHasSchemaWithoutRemovingFull(t *testing.T) {
	schema := &CommandSchema{Arguments: []Argument{{Name: "container"}}}
	command := Command{Full: "wslc ps {container}", Schema: schema}

	if command.Schema != schema {
		t.Fatal("command schema was not retained")
	}
	if command.Full != "wslc ps {container}" {
		t.Fatalf("Full changed during schema migration: %q", command.Full)
	}
}

func TestOptionKindValues(t *testing.T) {
	want := []OptionKind{"boolean", "text", "select", "numeric"}
	got := []OptionKind{OptionKindBoolean, OptionKindText, OptionKindSelect, OptionKindNumeric}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("option kind %d = %q, want %q", i, got[i], want[i])
		}
	}
}
