package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jacobalberty/jobwrapper/internal/command"
	"github.com/jacobalberty/jobwrapper/internal/lock"
)

func TestRun_WithMocks(t *testing.T) {
	testCases := []struct {
		name              string
		mockConfig        string
		expectedCmdStdout string
		expectedCmdStderr string
		lockExists        bool
		expectError       bool
		setupMocks        func(context.Context) TestMocks                             // The test setup function can return specific mocks if needed
		expectHandler     func(t *testing.T, stdout, stderr *bytes.Buffer, err error) // Optional custom handler for specific tests
	}{
		{
			name: "Successful Command Execution",
			mockConfig: `
                lock_dir = "/mock/lockdir"
                timeout = 60
                lock_filename = ".mocklock"
                history_lines = 5
            `,
			expectedCmdStdout: "Hello, world!\n", // Expected stdout content
			expectedCmdStderr: "",
			lockExists:        false,
			expectError:       false,
			setupMocks: func(ctx context.Context) TestMocks {
				return testSetup(t, nil, nil, func(ctx context.Context, name string, args ...string) command.Command {
					return &command.MockCommand{
						StdoutContent: "Hello, world!\n",
						StderrContent: "",
						RunFunc:       func() error { return nil },
					}
				})
			},
			expectHandler: nil, // No custom handler, use default
		},
		{
			name: "Command Fails with Error",
			mockConfig: `
                lock_dir = "/mock/lockdir"
                timeout = 60
                lock_filename = ".mocklock"
                history_lines = 5
            `,
			expectedCmdStdout: "",
			expectedCmdStderr: "Command failed\n", // Expected stderr content
			lockExists:        false,
			expectError:       true,
			setupMocks: func(ctx context.Context) TestMocks {
				return testSetup(t, nil, nil, func(ctx context.Context, name string, args ...string) command.Command {
					return &command.MockCommand{
						StdoutContent: "",
						StderrContent: "Command failed\n",
						RunFunc:       func() error { return fmt.Errorf("mock command error") }, // Simulate command failure
					}
				})
			},
			expectHandler: nil, // No custom handler, use default
		},
		{
			name: "Lock Already Exists",
			mockConfig: `
                lock_dir = "/mock/lockdir"
                timeout = 60
                lock_filename = ".mocklock"
                history_lines = 5
            `,
			expectedCmdStdout: "",
			expectedCmdStderr: "",
			lockExists:        true, // Simulate lock already exists
			expectError:       true,
			setupMocks: func(ctx context.Context) TestMocks {
				mockLocker := lock.NewMockLocker()
				if err := mockLocker.Acquire(ctx, "backup"); err != nil {
					t.Fatalf("Failed to acquire lock: %v", err)
				}
				return testSetup(t, nil, mockLocker, nil)
			},
			expectHandler: func(t *testing.T, stdout, stderr *bytes.Buffer, err error) {
				if err == nil {
					t.Fatalf("Expected an error but got none")
				}
				if !strings.Contains(err.Error(), "error acquiring lock") {
					t.Errorf("Expected lock acquisition error, got '%v'", err)
				}
				if stdout.String() != "" {
					t.Errorf("Expected empty stdout, got '%s'", stdout.String())
				}
				if stderr.String() != "" {
					t.Errorf("Expected empty stderr, got '%s'", stderr.String())
				}
			},
		},
		{
			name: "Custom Config with Timeout",
			mockConfig: `
                lock_dir = "/custom/lockdir"
                timeout = 120
                lock_filename = ".customlock"
                history_lines = 7
            `,
			expectedCmdStdout: "Delayed output\n", // Simulate delayed output
			expectedCmdStderr: "",
			lockExists:        false,
			expectError:       false,
			setupMocks: func(ctx context.Context) TestMocks {
				return testSetup(t, nil, nil, func(ctx context.Context, name string, args ...string) command.Command {
					return &command.MockCommand{
						StdoutContent: "Delayed output\n",
						StderrContent: "",
						RunFunc: func() error {
							time.Sleep(500 * time.Millisecond) // Simulate a delay
							return nil
						},
					}
				})
			},
			expectHandler: nil, // No custom handler, use default
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mocks := tc.setupMocks(ctx)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			args := []string{"backup", "/mock/script.sh"}

			err := run(ctx, args, stdout, stderr, mocks.FileSystem, mocks.Locker, mocks.CommandContext)

			// Default checks for stdout, stderr, and error
			if stdout.String() != tc.expectedCmdStdout {
				t.Errorf("Expected stdout to be '%s', got '%s'", tc.expectedCmdStdout, stdout.String())
			}

			if stderr.String() != tc.expectedCmdStderr {
				t.Errorf("Expected stderr to be '%s', got '%s'", tc.expectedCmdStderr, stderr.String())
			}

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error but got: %v", err)
				}
			}

			// Run custom handler if provided
			if tc.expectHandler != nil {
				tc.expectHandler(t, stdout, stderr, err)
			}
		})
	}
}
