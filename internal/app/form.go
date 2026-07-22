package app

import "wslc-tui-ms/internal/commands"

// formOption is an editable copy of one schema option. The schema remains the
// source of truth for ordering and validation metadata.
type formOption struct {
	flag       string
	kind       commands.OptionKind
	value      string
	defaultVal string
	choices    []string
	required   bool
	validation commands.Validation
}

type formState struct {
	commandKey      string
	commandSchema   commands.CommandSchema
	argumentRows    [][]string
	options         []formOption
	focusedField    int
	validationError error
	buildResult     commands.BuildResult
}

func newFormState(command commands.Command, remembered map[string]string) *formState {
	form := &formState{commandKey: commandIdentity(command)}
	if command.Schema == nil {
		return form
	}

	form.commandSchema.Arguments = append([]commands.Argument(nil), command.Schema.Arguments...)
	form.commandSchema.Options = append([]commands.Option(nil), command.Schema.Options...)
	for index := range form.commandSchema.Options {
		form.commandSchema.Options[index].Choices = append([]string(nil), command.Schema.Options[index].Choices...)
	}
	for range form.commandSchema.Arguments {
		form.argumentRows = append(form.argumentRows, []string{""})
	}
	for _, option := range form.commandSchema.Options {
		value := option.Default
		if rememberedValue, ok := remembered[option.Flag]; ok {
			value = rememberedValue
		}
		form.options = append(form.options, formOption{
			flag:       option.Flag,
			kind:       option.Kind,
			value:      value,
			defaultVal: option.Default,
			choices:    append([]string(nil), option.Choices...),
			required:   option.Required,
			validation: option.Validation,
		})
	}
	return form
}

func commandIdentity(command commands.Command) string {
	if command.Full != "" {
		return command.Full
	}
	return command.Category + "/" + command.Name
}

func (f *formState) optionValues() map[string]string {
	values := make(map[string]string, len(f.options))
	for _, option := range f.options {
		values[option.flag] = option.value
	}
	return values
}

func (f *formState) optionFlags() []string {
	flags := make([]string, 0, len(f.options))
	for _, option := range f.options {
		flags = append(flags, option.flag)
	}
	return flags
}

func (f *formState) optionValue(flag string) string {
	for _, option := range f.options {
		if option.flag == flag {
			return option.value
		}
	}
	return ""
}

func (f *formState) setOption(flag, value string) bool {
	for index := range f.options {
		if f.options[index].flag == flag {
			f.options[index].value = value
			return true
		}
	}
	return false
}

func (f *formState) addRepeatableRow() bool {
	if len(f.commandSchema.Arguments) == 0 {
		return false
	}
	argument := f.commandSchema.Arguments[len(f.commandSchema.Arguments)-1]
	if !argument.Repeatable {
		return false
	}
	f.argumentRows = append(f.argumentRows, []string{""})
	return true
}

func (f *formState) removeRepeatableRow(index int) bool {
	if len(f.commandSchema.Arguments) == 0 || index < len(f.commandSchema.Arguments)-1 || index >= len(f.argumentRows) {
		return false
	}
	argument := f.commandSchema.Arguments[len(f.commandSchema.Arguments)-1]
	if !argument.Repeatable {
		return false
	}
	f.argumentRows = append(f.argumentRows[:index], f.argumentRows[index+1:]...)
	if f.focusedField >= f.fieldCount() && f.fieldCount() > 0 {
		f.focusedField = f.fieldCount() - 1
	}
	return true
}

func (f *formState) fieldCount() int {
	return len(f.argumentRows) + len(f.options)
}

func (f *formState) moveFocus(delta int) int {
	count := f.fieldCount()
	if count == 0 {
		f.focusedField = 0
		return 0
	}
	f.focusedField += delta
	if f.focusedField < 0 {
		f.focusedField = 0
	}
	if f.focusedField >= count {
		f.focusedField = count - 1
	}
	return f.focusedField
}

func (f *formState) build(command []string) commands.BuildResult {
	f.buildResult = commands.Build(command, f.commandSchema, f.argumentRows, f.optionValues())
	if len(f.buildResult.Errors) > 0 {
		f.validationError = f.buildResult.Errors[0]
	}
	return f.buildResult
}

func (m *model) openCommandForm(command commands.Command) {
	if m.formOptionMemory == nil {
		m.formOptionMemory = make(map[string]map[string]string)
	}
	remembered := m.formOptionMemory[commandIdentity(command)]
	m.form = newFormState(command, remembered)
}

func (m *model) rememberFormOptions() {
	if m.form == nil {
		return
	}
	if m.formOptionMemory == nil {
		m.formOptionMemory = make(map[string]map[string]string)
	}
	values := m.form.optionValues()
	remembered := make(map[string]string, len(values))
	for flag, value := range values {
		remembered[flag] = value
	}
	m.formOptionMemory[m.form.commandKey] = remembered
}
