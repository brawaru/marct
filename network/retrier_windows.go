//go:build windows

package network

import (
	"runtime"
	"syscall"

	"golang.org/x/sys/windows"
)

func init() {
	defaultCheck := isNetworkReset

	isNetworkReset = func(errno syscall.Errno) bool {
		if runtime.GOOS == "windows" {
			// on windows the code is different
			if errno == windows.WSAECONNRESET {
				return true
			}
		}

		return defaultCheck(errno)
	}
}
