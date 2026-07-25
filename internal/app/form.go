package app

import (
	"fmt"
	"wslc-tui-ms/internal/commands"
)

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
	pickerRows      map[int]bool
}

func newFormState(command commands.Command, remembered map[string]string) *formState {
	form := &formState{commandKey: commandIdentity(command), pickerRows: make(map[int]bool)}
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

func (f *formState) pickerArgument(field int) (commands.Argument, bool) {
	if field < 0 || field >= len(f.argumentRows) || len(f.commandSchema.Arguments) == 0 {
		return commands.Argument{}, false
	}
	if field >= len(f.commandSchema.Arguments) {
		argument := f.commandSchema.Arguments[len(f.commandSchema.Arguments)-1]
		return argument, argument.Repeatable && argument.PickerAvailable()
	}
	argument := f.commandSchema.Arguments[field]
	return argument, argument.PickerAvailable()
}

func (f *formState) pickerArgumentStart(field int) int {
	if field >= 0 && field < len(f.commandSchema.Arguments) {
		return field
	}
	if len(f.commandSchema.Arguments) > 0 && f.commandSchema.Arguments[len(f.commandSchema.Arguments)-1].Repeatable {
		return len(f.commandSchema.Arguments) - 1
	}
	return field
}

func (f *formState) pickerValues(field int) []string {
	argument, ok := f.pickerArgument(field)
	if !ok {
		return nil
	}
	values := make([]string, 0)
	for row := field; row < len(f.argumentRows); row++ {
		if !argument.Repeatable && row > field {
			break
		}
		if f.pickerRows[row] && len(f.argumentRows[row]) == 1 && f.argumentRows[row][0] != "" {
			values = append(values, f.argumentRows[row][0])
		}
	}
	return values
}

func (f *formState) clearPickerSelection(field int) {
	argument, ok := f.pickerArgument(field)
	if !ok {
		return
	}
	if argument.Repeatable {
		for row := field; row < len(f.argumentRows); row++ {
			delete(f.pickerRows, row)
		}
		return
	}
	delete(f.pickerRows, field)
}

func (f *formState) clearPickerRow(row int) {
	delete(f.pickerRows, row)
}

func (f *formState) setPickerValues(field int, values []string) bool {
	argument, ok := f.pickerArgument(field)
	if !ok {
		return false
	}
	f.clearPickerSelection(field)
	if argument.Repeatable {
		f.argumentRows = append(f.argumentRows[:field], make([][]string, 0, len(values))...)
		for _, value := range values {
			f.argumentRows = append(f.argumentRows, []string{value})
			f.pickerRows[len(f.argumentRows)-1] = true
		}
		return true
	}
	if len(f.argumentRows) <= field {
		return false
	}
	if len(values) == 0 {
		f.argumentRows[field] = []string{""}
		return true
	}
	f.argumentRows[field] = []string{values[0]}
	f.pickerRows[field] = true
	return true
}

func (f *formState) validatePickerValues(field int, available []string) error {
	allowed := make(map[string]bool, len(available))
	for _, value := range available {
		allowed[value] = true
	}
	for _, value := range f.pickerValues(field) {
		if !allowed[value] {
			return fmt.Errorf("selected %q for argument %q is no longer available", value, f.commandSchema.Arguments[field].Name)
		}
	}
	return nil
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
