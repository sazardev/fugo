//go:build !windows

package main

import "os"

// enableVirtualTerminal is a no-op on non-Windows platforms, where terminals
// support ANSI escape sequences natively.
func enableVirtualTerminal(*os.File) {}
