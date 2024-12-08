package lock

import (
	"context"
	"fmt"
	"sync"

	"github.com/jacobalberty/jobwrapper/internal/config"
	"github.com/jacobalberty/jobwrapper/internal/filesystem"
)

// MockLocker provides a mock implementation of the Locker interface.
type MockLocker struct {
	locks       map[string]bool
	mu          sync.Mutex
	AcquireFunc func(lockName string) error // Customizable Acquire function for mocking
	ReleaseFunc func(lockName string) error // Customizable Release function for mocking
}

// NewMockLocker creates a mock Locker instance
func NewMockLocker() *MockLocker {
	return &MockLocker{
		locks: make(map[string]bool),
	}
}

// WrapFactory wraps a MockLocker in a LockFactory function
func WrapFactory(mockLocker *MockLocker) LockFactory {
	return func(cfg *config.Config, fs filesystem.FileSystem) (Locker, error) {
		// Return the mockLocker wrapped as a LockFactory
		return mockLocker, nil
	}
}

// Acquire simulates acquiring a lock.
// It uses the AcquireFunc for mockable behavior.
func (ml *MockLocker) Acquire(ctx context.Context, lockName string) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.AcquireFunc != nil {
		return ml.AcquireFunc(lockName)
	}

	// Default behavior if AcquireFunc isn't set
	if ml.locks[lockName] {
		return fmt.Errorf("lock %s already exists", lockName)
	}
	ml.locks[lockName] = true
	return nil
}

// Release simulates releasing a lock.
// It uses the ReleaseFunc for mockable behavior.
func (ml *MockLocker) Release(lockName string) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.ReleaseFunc != nil {
		return ml.ReleaseFunc(lockName)
	}

	// Default behavior if ReleaseFunc isn't set
	if !ml.locks[lockName] {
		return fmt.Errorf("lock %s does not exist", lockName)
	}
	delete(ml.locks, lockName)
	return nil
}
