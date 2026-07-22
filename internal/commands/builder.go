package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BuildResult contains the arguments for execution, a human-readable command,
// and any validation errors found while building it.
type BuildResult struct {
	Args    []string
	Display string
	Errors  []error
}

// Build validates form values and builds a command. Display is for review only;
// it is never reparsed or executed.
// Each positional row contains exactly one value. Non-repeatable arguments
// consume one row in schema order; a repeatable argument consumes the rest.
func Build(command []string, schema CommandSchema, argumentRows [][]string, optionValues map[string]string) BuildResult {
	result := BuildResult{Args: append([]string(nil), command...)}
	rowIndex := 0
	hasRepeatable := false

	for argumentIndex, argument := range schema.Arguments {
		if argument.Repeatable {
			hasRepeatable = true
			if argumentIndex != len(schema.Arguments)-1 {
				result.Errors = append(result.Errors, fmt.Errorf("repeatable argument %q must be final", argument.Name))
			}
			startRow := rowIndex
			for ; rowIndex < len(argumentRows); rowIndex++ {
				value, ok := positionalValue(argumentRows[rowIndex])
				if !ok {
					result.Errors = append(result.Errors, fmt.Errorf("argument %q has an incomplete row", argument.Name))
					continue
				}
				validationErrors := appendValidationErrors(nil, argument.Name, value, argument.Validation)
				result.Errors = append(result.Errors, validationErrors...)
				if len(validationErrors) == 0 {
					result.Args = append(result.Args, value)
				}
			}
			if argument.Required && rowIndex == startRow {
				result.Errors = append(result.Errors, fmt.Errorf("required argument %q is missing", argument.Name))
			}
			continue
		}

		if rowIndex >= len(argumentRows) {
			if argument.Required {
				result.Errors = append(result.Errors, fmt.Errorf("required argument %q is missing", argument.Name))
			}
			continue
		}
		value, ok := positionalValue(argumentRows[rowIndex])
		rowIndex++
		if !ok {
			if len(argumentRows[rowIndex-1]) > 1 || argument.Required {
				result.Errors = append(result.Errors, fmt.Errorf("argument %q has an incomplete row", argument.Name))
			}
			continue
		}
		validationErrors := appendValidationErrors(nil, argument.Name, value, argument.Validation)
		result.Errors = append(result.Errors, validationErrors...)
		if len(validationErrors) == 0 {
			result.Args = append(result.Args, value)
		}
	}

	if rowIndex < len(argumentRows) && !hasRepeatable {
		result.Errors = append(result.Errors, fmt.Errorf("unexpected positional argument row"))
	}

	for _, option := range schema.Options {
		value, present := optionValues[option.Flag]
		if option.Kind == OptionKindBoolean && option.Required && (!present || value != "true") {
			result.Errors = append(result.Errors, fmt.Errorf("required option %q must be true", option.Flag))
			continue
		}
		if !present {
			value = option.Default
		}
		switch option.Kind {
		case OptionKindBoolean:
			if value == "" || value == "false" {
				continue
			}
			if value != "true" {
				result.Errors = append(result.Errors, fmt.Errorf("option %q must be boolean", option.Flag))
				continue
			}
			result.Args = append(result.Args, option.Flag)
		case OptionKindText:
			if value == "" {
				if option.Required {
					result.Errors = append(result.Errors, fmt.Errorf("required option %q is missing", option.Flag))
				}
				continue
			}
			validationErrors := appendValidationErrors(nil, option.Flag, value, option.Validation)
			result.Errors = append(result.Errors, validationErrors...)
			if len(validationErrors) == 0 {
				result.Args = append(result.Args, option.Flag, value)
			}
		case OptionKindSelect:
			if value == "" {
				if option.Required {
					result.Errors = append(result.Errors, fmt.Errorf("required option %q is missing", option.Flag))
				}
				continue
			}
			if !contains(option.Choices, value) {
				result.Errors = append(result.Errors, fmt.Errorf("option %q has invalid value %q", option.Flag, value))
				continue
			}
			result.Args = append(result.Args, option.Flag, value)
		case OptionKindNumeric:
			if value == "" {
				if option.Required {
					result.Errors = append(result.Errors, fmt.Errorf("required option %q is missing", option.Flag))
				}
				continue
			}
			number, err := strconv.Atoi(value)
			if err != nil || number < option.Validation.Min || (option.Validation.Max != 0 && number > option.Validation.Max) {
				result.Errors = append(result.Errors, fmt.Errorf("option %q has invalid numeric value %q", option.Flag, value))
				continue
			}
			result.Args = append(result.Args, option.Flag, value)
		default:
			result.Errors = append(result.Errors, fmt.Errorf("option %q has unsupported kind %q", option.Flag, option.Kind))
		}
	}

	result.Display = displayCommand(result.Args)
	return result
}

func positionalValue(row []string) (string, bool) {
	return func() (string, bool) {
		if len(row) != 1 || row[0] == "" {
			return "", false
		}
		return row[0], true
	}()
}

func appendValidationErrors(errors []error, name, value string, validation Validation) []error {
	if validation.MinLength > 0 && len(value) < validation.MinLength {
		errors = append(errors, fmt.Errorf("value for %q is too short", name))
	}
	if validation.MaxLength > 0 && len(value) > validation.MaxLength {
		errors = append(errors, fmt.Errorf("value for %q is too long", name))
	}
	if validation.Pattern != "" {
		pattern, err := regexp.Compile(validation.Pattern)
		if err != nil || !pattern.MatchString(value) {
			errors = append(errors, fmt.Errorf("value for %q is invalid", name))
		}
	}
	return errors
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func displayCommand(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		if arg != "" && strings.IndexFunc(arg, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune("_./:@%+=,-", r))
		}) == -1 {
			quoted[i] = arg
		} else {
			quoted[i] = strconv.Quote(arg)
		}
	}
	return strings.Join(quoted, " ")
}
