package lock

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/jacobalberty/jobwrapper/internal/config"
	"github.com/jacobalberty/jobwrapper/internal/filesystem"
)

// FileLocker implements the Locker interface using lock files
type FileLocker struct {
	cfg       *config.Config
	fs        filesystem.FileSystem
	fileLocks map[string]*flock.Flock
}

// NewFileLocker creates a new FileLocker with the given base path for lock files
func NewFileLocker(cfg *config.Config, fs filesystem.FileSystem) (Locker, error) {
	return &FileLocker{cfg: cfg, fs: fs, fileLocks: make(map[string]*flock.Flock)}, nil
}

func (fl *FileLocker) lockFilename(lockname string) string {
	return filepath.Join(fl.cfg.LockDir, lockname, fl.cfg.LockFileName)
}

// Acquire creates a lock file with the process ID
func (fl *FileLocker) Acquire(ctx context.Context, lockName string) error {
	var (
		groupLockDir = filepath.Join(fl.cfg.LockDir, lockName)
		err          error
		fileLock     *flock.Flock
		ok           bool
	)
	// Ensure lock directory exists
	if err := fl.fs.MkdirAll(groupLockDir, 0755); err != nil {
		return fmt.Errorf("error creating lock directory for group '%s': %w", lockName, err)
	}

	if fileLock, ok = fl.fileLocks[lockName]; ok && fileLock.Locked() {
		return fmt.Errorf("lock %s already exists", lockName)
	}

	if !ok {
		fileLock = flock.New(fl.lockFilename(lockName))
	}

	locked, err := fileLock.TryLockContext(ctx, 0)
	if err != nil {
		return fmt.Errorf("failed to acquire lock %s: %w", lockName, err)
	}
	if !locked {
		return fmt.Errorf("lock %s already exists", lockName)
	}

	fl.fileLocks[lockName] = fileLock

	return nil
}

// Release removes the lock file
func (fl *FileLocker) Release(lockName string) error {
	fileLock, ok := fl.fileLocks[lockName]
	if !ok {
		return fmt.Errorf("lock %s does not exist", lockName)
	}

	if err := fileLock.Unlock(); err != nil {
		return fmt.Errorf("failed to release lock %s: %w", lockName, err)
	}

	return nil
}
