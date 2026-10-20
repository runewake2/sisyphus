// Package shell runs a command line in the shell of the platform, so that every runner of hephaestus
// stops a command the same way: when the context ends, the whole command stops, also the programs
// that the shell started, for example a model runner or the children of the checks, so that a
// stopped task spends no more tokens or CPU.
package shell

import (
	"context"
	"os/exec"
	"runtime"
	"time"
)

// Command returns the command that runs line in the shell of the platform. When ctx ends, it stops
// the whole command, also the programs that the shell started.
func Command(ctx context.Context, line string) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", line)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", line)
	}
	stopTree(cmd)
	cmd.WaitDelay = 5 * time.Second
	return cmd
}
