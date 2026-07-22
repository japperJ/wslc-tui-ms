package commands

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type ExecutionResult struct {
	Output   string
	Error    error
	ExitCode int
	Duration time.Duration
}

func Execute(command string, timeout time.Duration) ExecutionResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()

	// Parse command and arguments
	args := ParseCommand(command)
	if len(args) == 0 {
		return ExecutionResult{Error: fmt.Errorf("empty command")}
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return ExecutionResult{
		Output:   output,
		Error:    err,
		ExitCode: exitCode,
		Duration: duration,
	}
}

func ParseCommand(command string) []string {
	var args []string
	var current string
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if inQuote {
			if ch == '\\' && i+1 < len(command) {
				next := command[i+1]
				if next == quoteChar || next == '\\' {
					current += string(next)
					i++
					continue
				}
				current += string(ch)
			} else if ch == quoteChar {
				inQuote = false
			} else {
				current += string(ch)
			}
		} else {
			if ch == '\\' && i+1 < len(command) {
				next := command[i+1]
				if next == '"' || next == '\'' || next == '\\' || next == ' ' {
					current += string(next)
					i++
					continue
				}
				current += string(ch)
			} else if ch == '"' || ch == '\'' {
				inQuote = true
				quoteChar = ch
			} else if ch == ' ' || ch == '\t' {
				if current != "" {
					args = append(args, current)
					current = ""
				}
			} else {
				current += string(ch)
			}
		}
	}

	if inQuote {
		args = append(args, current)
	} else if current != "" {
		args = append(args, current)
	}

	return args
}
