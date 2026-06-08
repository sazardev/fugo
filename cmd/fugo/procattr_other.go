//go:build !windows

package main

import "os/exec"

// setNewProcessGroup is a no-op outside Windows; the default process handling
// is sufficient on Unix-like systems.
func setNewProcessGroup(_ *exec.Cmd) {}
