//go:build windows

package update

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

const processStillActive = 259

func WaitForProcessExit(pid int, timeout time.Duration) error {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(h)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var code uint32
		if err := windows.GetExitCodeProcess(h, &code); err != nil {
			return fmt.Errorf("query parent process: %w", err)
		}
		if code != processStillActive {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for parent process %d to exit", pid)
}
