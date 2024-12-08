package command

import (
	"context"
	"io"
)

// Command defines an interface for wrapping command execution
type Command interface {
	Run() error
	SetStdout(io.Writer)
	SetStderr(io.Writer)
}

// CommandContextFunc abstracts the creation of commands
type CommandContextFunc func(ctx context.Context, name string, args ...string) Command
