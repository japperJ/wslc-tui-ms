//go:build !windows

package update

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func processIsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func WaitForProcessExit(pid int, timeout time.Duration) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for parent process %d to exit", pid)
}
