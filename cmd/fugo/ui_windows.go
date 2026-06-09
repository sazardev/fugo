//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableVirtualTerminal turns on ANSI escape processing for the given console
// handle so colored output renders on legacy Windows consoles (conhost).
// Windows Terminal enables it by default; this makes older hosts work too.
func enableVirtualTerminal(f *os.File) {
	h := windows.Handle(f.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}

	_ = windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}
