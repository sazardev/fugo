//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// setNewProcessGroup starts the child in a new process group so the parent can
// signal it independently of its own console (Windows-specific).
func setNewProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
