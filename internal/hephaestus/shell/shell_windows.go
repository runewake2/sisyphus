//go:build windows

package shell

import "os/exec"

// stopTree leaves the default: a cancel stops the shell.
func stopTree(cmd *exec.Cmd) {}
