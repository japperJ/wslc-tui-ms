package commands

import "regexp"

var placeholderRe = regexp.MustCompile(`\{[a-zA-Z_][a-zA-Z0-9_]*\}`)

// ExtractPlaceholders returns the ordered, de-duplicated placeholder names
// found in a command string. A placeholder looks like "{name}"; the returned
// names have the braces stripped (e.g. "name"). Returns an empty slice when the
// command contains no placeholders.
func ExtractPlaceholders(full string) []string {
	matches := placeholderRe.FindAllString(full, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(matches))
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		name := m[1 : len(m)-1] // strip { }
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

// FindPlaceholderIndex returns the [start, end) byte offsets of the first
// placeholder token in s, or nil if none is found.
func FindPlaceholderIndex(s string) []int {
	return placeholderRe.FindStringIndex(s)
}

// SubstitutePlaceholders replaces each "{name}" token in full with the value
// from values[name]. Placeholders without a corresponding value are left
// unchanged.
func SubstitutePlaceholders(full string, values map[string]string) string {
	return placeholderRe.ReplaceAllStringFunc(full, func(token string) string {
		name := token[1 : len(token)-1]
		if v, ok := values[name]; ok {
			return v
		}
		return token
	})
}
