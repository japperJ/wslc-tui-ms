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
	source := schema.Arguments[0]
	if source.Label != "Source image" || source.Placeholder != "image:tag" ||
		source.Validation.Pattern != `^[a-z0-9/:.-]+$` {
		t.Fatalf("source display and validation metadata was not preserved: %#v", source)
	}
	if !source.Required || source.Repeatable {
		t.Fatalf("source metadata was not preserved: %#v", schema.Arguments[0])
	}
	target := schema.Arguments[1]
	if target.Label != "Target image" || !target.Repeatable || target.Required ||
		target.Validation.MinLength != 1 || target.Validation.MaxLength != 64 {
		t.Fatalf("target metadata was not preserved: %#v", schema.Arguments[1])
	}
}

func TestResourceMetadataPreservesPickerContract(t *testing.T) {
	schema := CommandSchema{Arguments: []Argument{
		{Name: "container", ResourceType: ResourceTypeContainer, PickerEnabled: true},
		{Name: "containers", Repeatable: true, ResourceType: ResourceTypeContainer, PickerEnabled: true},
	}}

	if schema.Arguments[0].ResourceType != ResourceTypeContainer || !schema.Arguments[0].PickerEnabled || schema.Arguments[0].Repeatable {
		t.Fatalf("scalar resource metadata = %#v", schema.Arguments[0])
	}
	if schema.Arguments[1].ResourceType != ResourceTypeContainer || !schema.Arguments[1].PickerEnabled || !schema.Arguments[1].Repeatable {
		t.Fatalf("repeatable resource metadata = %#v", schema.Arguments[1])
	}
}

func TestUnknownOrEmptyResourceTypeIsTextOnly(t *testing.T) {
	arguments := []Argument{
		{Name: "free-text", PickerEnabled: true},
		{Name: "unknown", ResourceType: ResourceType("workspace"), PickerEnabled: true},
	}

	for _, argument := range arguments {
		if argument.PickerAvailable() {
			t.Errorf("%q must be text-only: %#v", argument.Name, argument)
		}
	}
}

func TestCommandSchemaPreservesOrderedOptions(t *testing.T) {
	schema := CommandSchema{
		Options: []Option{
			{Flag: "--all", Description: "Show all", Kind: OptionKindBoolean},
			{Flag: "--format", Description: "Output format", Kind: OptionKindSelect, Default: "table", Choices: []string{"table", "json"}},
			{Flag: "--name", Description: "Container name", Kind: OptionKindText, Required: true, Validation: Validation{MinLength: 1}},
			{Flag: "--timeout", Description: "Timeout in seconds", Kind: OptionKindNumeric, Validation: Validation{Min: 0, Max: 300}},
		},
	}

	if len(schema.Options) != 4 {
		t.Fatalf("expected 4 options, got %d", len(schema.Options))
	}
	if schema.Options[0].Flag != "--all" || schema.Options[1].Flag != "--format" ||
		schema.Options[2].Flag != "--name" || schema.Options[3].Flag != "--timeout" {
		t.Fatalf("options are not ordered: %#v", schema.Options)
	}
	all := schema.Options[0]
	if all.Description != "Show all" || all.Kind != OptionKindBoolean || all.Required || all.Default != "" {
		t.Fatalf("boolean metadata was not preserved: %#v", all)
	}
	format := schema.Options[1]
	if format.Description != "Output format" || format.Kind != OptionKindSelect || format.Required ||
		format.Default != "table" || len(format.Choices) != 2 ||
		format.Choices[0] != "table" || format.Choices[1] != "json" {
		t.Fatalf("select metadata was not preserved: %#v", schema.Options[1])
	}
	name := schema.Options[2]
	if name.Description != "Container name" || name.Kind != OptionKindText || !name.Required ||
		name.Default != "" || name.Validation.MinLength != 1 {
		t.Fatalf("required text metadata was not preserved: %#v", schema.Options[2])
	}
	timeout := schema.Options[3]
	if timeout.Description != "Timeout in seconds" || timeout.Kind != OptionKindNumeric || timeout.Required ||
		timeout.Default != "" || timeout.Validation.Min != 0 || timeout.Validation.Max != 300 {
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
