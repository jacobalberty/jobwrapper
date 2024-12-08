package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/jacobalberty/jobwrapper/internal/command"
	"github.com/jacobalberty/jobwrapper/internal/config"
	"github.com/jacobalberty/jobwrapper/internal/filesystem"
	"github.com/jacobalberty/jobwrapper/internal/history"
	"github.com/jacobalberty/jobwrapper/internal/lock"
)

func main() {
	fs := filesystem.OSFileSystem{}
	cmdCtx := command.NewRealCommandContext

	// Use the NewFileLockFactory function to create the LockFactory
	lockFactory := lock.NewFileLocker

	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, fs, lockFactory, cmdCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	stdout, stderr io.Writer,
	fs filesystem.FileSystem,
	lockFactory lock.LockFactory,
	commandCtx command.CommandContextFunc,
) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: jobwrapper <group> <script> [args...]")
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	group := args[0]
	script := args[1]
	cmdArgs := args[2:]

	// Load configuration
	cfg := config.LoadConfig(fs)

	// Create the locker using the LockFactory function
	locker, err := lockFactory(&cfg, fs)
	if err != nil {
		return fmt.Errorf("error creating locker: %w", err)
	}

	// Set up a timeout for the lock acquisition
	lockCtx, lockCancel := context.WithTimeout(ctx, cfg.Timeout)
	defer lockCancel()

	// Acquire lock
	if err := locker.Acquire(lockCtx, group); err != nil {
		return fmt.Errorf("error acquiring lock for group '%s': %w", group, err)
	}
	defer func() {
		if err := locker.Release(group); err != nil {
			fmt.Fprintf(stderr, "Error releasing lock for group '%s': %v\n", group, err)
		}
	}()

	scriptName := filepath.Base(script)

	historyWriter, err := history.WriteHistory(fs, filepath.Join(cfg.LockDir, group, scriptName), cfg.HistoryLines, cmdArgs)
	if err != nil {
		return fmt.Errorf("error creating history writer: %w", err)
	}

	// Execute job
	cmd := commandCtx(ctx, script, cmdArgs...)
	cmd.SetStdout(stdout)
	cmd.SetStderr(stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("job execution for script '%s' failed: %w", script, err)
	}

	if err := historyWriter(); err != nil {
		return fmt.Errorf("error writing history: %w", err)
	}

	return nil
}
