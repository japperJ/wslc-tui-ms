package commands

// OptionKind identifies the control used to edit an option value.
type OptionKind string

// ResourceType identifies a discoverable WSLC resource.
type ResourceType string

const (
	ResourceTypeContainer ResourceType = "container"
	ResourceTypeImage     ResourceType = "image"
	ResourceTypeNetwork   ResourceType = "network"
	ResourceTypeVolume    ResourceType = "volume"
	ResourceTypeSession   ResourceType = "session"
)

const (
	// OptionKindBoolean identifies a boolean option.
	OptionKindBoolean OptionKind = "boolean"
	// OptionKindText identifies a free-form text option.
	OptionKindText OptionKind = "text"
	// OptionKindSelect identifies an option with a fixed set of choices.
	OptionKindSelect OptionKind = "select"
	// OptionKindNumeric identifies a numeric option.
	OptionKindNumeric OptionKind = "numeric"
)

// Validation describes constraints applied to an argument or option value.
type Validation struct {
	Pattern   string
	MinLength int
	MaxLength int
	Min       int
	Max       int
}

// Argument describes an ordered positional command argument.
type Argument struct {
	Name          string
	Label         string
	Required      bool
	Repeatable    bool
	ResourceType  ResourceType
	PickerEnabled bool
	Placeholder   string
	Validation    Validation
}

// PickerAvailable reports whether this argument has a supported resource binding.
func (a Argument) PickerAvailable() bool {
	switch a.ResourceType {
	case ResourceTypeContainer, ResourceTypeImage, ResourceTypeNetwork, ResourceTypeVolume, ResourceTypeSession:
		return a.PickerEnabled
	default:
		return false
	}
}

// Option describes an ordered command option.
type Option struct {
	Flag        string
	Description string
	Kind        OptionKind
	Default     string
	Choices     []string
	Required    bool
	Validation  Validation
}

// CommandSchema describes the positional arguments and options for a command.
type CommandSchema struct {
	Arguments []Argument
	Options   []Option
}

type Command struct {
	Name        string
	Full        string
	Category    string
	Description string
	Usage       string
	Examples    []string
	Flags       []Flag
	Difficulty  string
	Tags        []string
	Schema      *CommandSchema
}

type Flag struct {
	Short       string
	Long        string
	Description string
	Default     string
	Required    bool
}
