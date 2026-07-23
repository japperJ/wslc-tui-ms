//go:build !windows

package platform

// IsElevated is only meaningful on Windows, where WSLC runs.
func IsElevated() bool {
	return true
}
