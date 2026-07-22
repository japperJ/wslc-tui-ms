package commands

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
}

type Flag struct {
	Short       string
	Long        string
	Description string
	Default     string
	Required    bool
}
