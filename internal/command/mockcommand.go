package command

import (
	"fmt"
	"io"
)

// MockCommand provides a mock implementation of the Command interface
type MockCommand struct {
	RunFunc       func() error
	SetStdoutFunc func(io.Writer)
	SetStderrFunc func(io.Writer)
	StdoutContent string // Mock stdout output
	StderrContent string // Mock stderr output
	stdout        io.Writer
	stderr        io.Writer
}

func (mc *MockCommand) Run() error {
	// Write mock content to stdout and stderr
	if mc.stdout != nil && mc.StdoutContent != "" {
		fmt.Fprint(mc.stdout, mc.StdoutContent)
	}
	if mc.stderr != nil && mc.StderrContent != "" {
		fmt.Fprint(mc.stderr, mc.StderrContent)
	}
	if mc.RunFunc != nil {
		return mc.RunFunc()
	}
	return nil
}

func (mc *MockCommand) SetStdout(w io.Writer) {
	mc.stdout = w
	if mc.SetStdoutFunc != nil {
		mc.SetStdoutFunc(w)
	}
}

func (mc *MockCommand) SetStderr(w io.Writer) {
	mc.stderr = w
	if mc.SetStderrFunc != nil {
		mc.SetStderrFunc(w)
	}
}
