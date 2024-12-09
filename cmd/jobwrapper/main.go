package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/jacobalberty/jobwrapper/internal/command"
	"github.com/jacobalberty/jobwrapper/internal/config"
	"github.com/jacobalberty/jobwrapper/internal/filesystem"
	"github.com/jacobalberty/jobwrapper/internal/history"
	"github.com/jacobalberty/jobwrapper/internal/lock"
)

func main() {
	if err := run(
		context.Background(),
		os.Args[1:],
		os.Stdout,
		os.Stderr,
		filesystem.OSFileSystem{},
		lock.NewFileLocker,
		command.NewRealCommandContext,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	fs filesystem.FileSystem,
	lockFactory lock.LockFactory,
	commandCtx command.CommandContextFunc,
) error {
	var (
		historyWriter history.HistoryWriter
		locker        lock.Locker
		err           error
	)
	if len(args) < 2 {
		return fmt.Errorf("usage: jobwrapper <group> <script> [args...]")
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	group := args[0]
	cmd := args[1]
	cmdArgs := args[2:]

	// Load configuration
	cfg := config.LoadConfig(fs)

	// Create the locker using the LockFactory function
	locker, err = lockFactory(&cfg, fs)
	if err != nil {
		return err
	}

	historyWriter, err = history.NewHistoryWriter(fs, &cfg, cmd, cmdArgs)
	if err != nil {
		return fmt.Errorf("error creating history writer: %w", err)
	}
	defer func() {
		if historyErr := historyWriter.WriteHistory(err); historyErr != nil {
			fmt.Fprintf(stderr, "Error writing history: %v\n", historyErr)
		}
	}()

	// Set up a timeout for the lock acquisition
	lockCtx, lockCancel := context.WithTimeout(ctx, cfg.Timeout)
	defer lockCancel()

	// Acquire lock
	if err = locker.Acquire(lockCtx, group); err != nil {
		return fmt.Errorf("error acquiring lock for group '%s': %w", group, err)
	}
	defer func() {
		if err = locker.Release(group); err != nil {
			fmt.Fprintf(stderr, "Error releasing lock for group '%s': %v\n", group, err)
		}
	}()

	historyWriter.MarkExecutionStart()

	// Execute job
	cmdCtx := commandCtx(ctx, cmd, cmdArgs...)
	cmdCtx.SetStdout(stdout)
	cmdCtx.SetStderr(stderr)

	if err = cmdCtx.Run(); err != nil {
		return fmt.Errorf("job execution for script '%s' failed: %w", cmd, err)
	}

	historyWriter.MarkExecutionEnd()

	return nil
}
