package commands

type OptionKind string

const (
	OptionKindBoolean OptionKind = "boolean"
	OptionKindText    OptionKind = "text"
	OptionKindSelect  OptionKind = "select"
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

type Argument struct {
	Name        string
	Label       string
	Required    bool
	Repeatable  bool
	Placeholder string
	Validation  Validation
}

type Option struct {
	Flag        string
	Description string
	Kind        OptionKind
	Default     string
	Choices     []string
	Required    bool
	Validation  Validation
}

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
