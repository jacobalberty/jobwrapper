package command

import (
	"context"
	"io"
	"os/exec"
)

// RealCommand wraps exec.Cmd for actual command execution
type RealCommand struct {
	cmd *exec.Cmd
}

func (rc *RealCommand) Run() error {
	return rc.cmd.Run()
}

func (rc *RealCommand) SetStdout(w io.Writer) {
	rc.cmd.Stdout = w
}

func (rc *RealCommand) SetStderr(w io.Writer) {
	rc.cmd.Stderr = w
}

// NewRealCommandContext creates a RealCommand from exec.CommandContext
func NewRealCommandContext(ctx context.Context, name string, args ...string) Command {
	return &RealCommand{cmd: exec.CommandContext(ctx, name, args...)}
}
