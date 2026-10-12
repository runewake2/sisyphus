//go:build !windows

package toolagent

import (
	"os/exec"
	"syscall"
)

// stopTree runs cmd in a process group of its own, and makes a cancel stop the whole group.
func stopTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
