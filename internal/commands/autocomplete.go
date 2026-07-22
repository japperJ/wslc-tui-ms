package commands

import (
	"strings"

	"github.com/sahilm/fuzzy"
)

type MatchResult struct {
	Command   Command
	MatchedAt []int
}

func Autocomplete(input string, commands []Command) []MatchResult {
	if input == "" {
		return allAsResults(commands)
	}

	lower := strings.ToLower(input)

	// Build search targets
	targets := make([]string, len(commands))
	for i, cmd := range commands {
		targets[i] = strings.ToLower(cmd.Full)
	}

	matches := fuzzy.Find(lower, targets)

	results := make([]MatchResult, 0, len(matches))
	for _, match := range matches {
		results = append(results, MatchResult{
			Command:   commands[match.Index],
			MatchedAt: match.MatchedIndexes,
		})
	}

	return results
}

func allAsResults(commands []Command) []MatchResult {
	results := make([]MatchResult, len(commands))
	for i, cmd := range commands {
		results[i] = MatchResult{Command: cmd}
	}
	return results
}

func FilterByPrefix(input string, commands []Command) []Command {
	if input == "" {
		return commands
	}

	lower := strings.ToLower(input)
	var filtered []Command

	for _, cmd := range commands {
		fullLower := strings.ToLower(cmd.Full)
		if strings.HasPrefix(fullLower, lower) {
			filtered = append(filtered, cmd)
		}
	}

	return filtered
}
