package app

import (
	"strings"
	"testing"
	"time"
	"wslc-tui-ms/internal/commands"
)

func TestOutputShowsSuccessfulSilentCommandStatus(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.outputCmd = "wslc start mycontainer"
	m.outputResult = &commands.ExecutionResult{
		ExitCode: 0,
		Duration: 12 * time.Millisecond,
	}

	m.setViewportOutput()
	plain := strings.ToLower(stripAnsi(m.viewportContent))
	if !strings.Contains(plain, "exit code 0") {
		t.Fatalf("output should show successful exit code, got:\n%s", m.viewportContent)
	}
	if !strings.Contains(plain, "no stdout/stderr") {
		t.Fatalf("output should explain silent command result, got:\n%s", m.viewportContent)
	}
}
