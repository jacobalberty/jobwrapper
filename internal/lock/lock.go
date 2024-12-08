package lock

import (
	"context"

	"github.com/jacobalberty/jobwrapper/internal/config"
	"github.com/jacobalberty/jobwrapper/internal/filesystem"
)

// Locker defines the interface for a locking mechanism
type Locker interface {
	Acquire(ctx context.Context, lockName string) error
	Release(lockName string) error
}

type LockFactory func(*config.Config, filesystem.FileSystem) (Locker, error)
