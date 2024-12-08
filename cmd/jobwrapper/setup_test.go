package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/jacobalberty/jobwrapper/internal/command"
	"github.com/jacobalberty/jobwrapper/internal/filesystem"
	"github.com/jacobalberty/jobwrapper/internal/lock"
)

// TestMocks is a struct that contains all the mocks used in the test
type TestMocks struct {
	FileSystem     *filesystem.MockFileSystem
	Locker         lock.LockFactory // Change to LockFactory type
	CommandContext func(ctx context.Context, name string, args ...string) command.Command
}

// testSetup prepares the necessary mocks or returns default implementations
func testSetup(t *testing.T, mockFileSystem *filesystem.MockFileSystem, mockLocker *lock.MockLocker, mockCmdCtx func(ctx context.Context, name string, args ...string) command.Command) TestMocks {
	t.Helper()
	// If mockFileSystem is not provided, use the default mock
	if mockFileSystem == nil {
		mockFileSystem = &filesystem.MockFileSystem{
			OpenFunc: func(name string) (io.ReadCloser, error) {
				return nil, os.ErrNotExist
			},
			MkdirAllFunc: func(path string, perm os.FileMode) error {
				return nil
			},
			OpenFileFunc: func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
				return &filesystem.ReadWriteCloserBuffer{Buffer: &bytes.Buffer{}}, nil
			},
			RemoveFunc: func(name string) error { return nil },
		}
	}

	// If mockLocker is not provided, use a default mock Locker
	if mockLocker == nil {
		mockLocker = lock.NewMockLocker()
	}

	// Create a LockFactory that returns the mock locker
	lockFactory := lock.WrapFactory(mockLocker)

	// If mockCmdCtx is not provided, use a default mock CommandContext
	if mockCmdCtx == nil {
		mockCmdCtx = func(ctx context.Context, name string, args ...string) command.Command {
			return &command.MockCommand{
				StdoutContent: "Executed command",
				StderrContent: "",
				RunFunc:       func() error { return nil },
			}
		}
	}

	return TestMocks{
		FileSystem:     mockFileSystem,
		Locker:         lockFactory, // Return the LockFactory function here
		CommandContext: mockCmdCtx,
	}
}
