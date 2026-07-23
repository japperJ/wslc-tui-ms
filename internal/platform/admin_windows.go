//go:build windows

package platform

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// IsElevated reports whether the current Windows process has an elevated token.
func IsElevated() bool {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false
	}
	defer token.Close()

	var elevation struct {
		TokenIsElevated uint32
	}
	var returned uint32
	if err := windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&returned,
	); err != nil {
		return false
	}

	return elevation.TokenIsElevated != 0
}
