package filesystem

import (
	"bytes"
	"errors"
	"io"
	"os"
)

// MockFileSystem provides a mock implementation of the FileSystem interface.
type MockFileSystem struct {
	// Maps file paths to their contents for mocking file reads
	Files map[string]*string

	// Functions to mock behavior
	MkdirAllFunc func(path string, perm os.FileMode) error
	OpenFunc     func(name string) (io.ReadCloser, error)
	OpenFileFunc func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)
	RemoveFunc   func(name string) error
}

// NewMockFileSystem creates a new MockFileSystem with optional file content for mocking.
func NewMockFileSystem(files map[string]*string) *MockFileSystem {
	return &MockFileSystem{
		Files: files,
		// Provide default implementations for each function.
		// These functions are used if no custom function is provided.
		MkdirAllFunc: nil,
		OpenFunc:     nil,
		OpenFileFunc: nil,
		RemoveFunc:   nil,
	}
}

// ReadWriteCloserBuffer provides a mock WriteCloser implementation for our mock file system.
type ReadWriteCloserBuffer struct {
	*bytes.Buffer
	dest *string
}

// Close simulates the closing of the file, returning no error.
func (w *ReadWriteCloserBuffer) Close() error {
	if w.dest != nil {
		*w.dest = w.String()
	}
	w.dest = nil
	w.Buffer.Reset()
	// No resources to release, so just return nil
	return nil
}

// MkdirAll mimics creating directories. Returns error if the custom MkdirAllFunc is not provided.
func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(path, perm)
	}
	// Call default method if no custom function is provided
	return m.MkdirAllDefault(path, perm)
}

// MkdirAllDefault provides the default behavior for MkdirAll.
func (m *MockFileSystem) MkdirAllDefault(path string, perm os.FileMode) error {
	// Default behavior: simulate success (directory creation always succeeds)
	return nil
}

// Open mimics opening a file. Returns error if custom OpenFunc is not provided and file is not in the map.
func (m *MockFileSystem) Open(name string) (io.ReadCloser, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(name)
	}
	// Call default method if no custom function is provided
	return m.OpenDefault(name)
}

// OpenDefault provides the default behavior for Open.
func (m *MockFileSystem) OpenDefault(name string) (io.ReadCloser, error) {
	// Default behavior: check the Files map
	if content, ok := m.Files[name]; ok {
		return io.NopCloser(bytes.NewReader([]byte(*content))), nil
	}
	return nil, errors.New("file not found")
}

// OpenFile mimics opening a file for writing. Returns error if custom OpenFileFunc is not provided.
func (m *MockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	if m.OpenFileFunc != nil {
		return m.OpenFileFunc(name, flag, perm)
	}
	// Call default method if no custom function is provided
	return m.OpenFileDefault(name, flag, perm)
}

// OpenFileDefault provides the default behavior for OpenFile.
func (m *MockFileSystem) OpenFileDefault(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	// Default behavior: check the Files map and handle lockfile
	if content, ok := m.Files[name]; ok {
		return &ReadWriteCloserBuffer{Buffer: bytes.NewBufferString(*content), dest: content}, nil
	}
	// Simulate lock file creation
	if name == "/mock/lockdir/.mocklock" {
		m.Files[name] = new(string)
		return &ReadWriteCloserBuffer{Buffer: new(bytes.Buffer), dest: m.Files[name]}, nil
	}
	// If O_CREATE flag is set, create the file in the map
	if flag&os.O_CREATE != 0 {
		m.Files[name] = new(string)
		return &ReadWriteCloserBuffer{Buffer: new(bytes.Buffer), dest: m.Files[name]}, nil
	}
	return nil, errors.New("file not found")
}

// Remove mimics removing a file. Returns error if custom RemoveFunc is not provided.
func (m *MockFileSystem) Remove(name string) error {
	if m.RemoveFunc != nil {
		return m.RemoveFunc(name)
	}
	// Call default method if no custom function is provided
	return m.RemoveDefault(name)
}

// RemoveDefault provides the default behavior for Remove.
func (m *MockFileSystem) RemoveDefault(name string) error {
	// Default behavior: check the Files map
	if _, ok := m.Files[name]; ok {
		delete(m.Files, name) // Remove the file from the map
		return nil
	}
	return errors.New("file not found")
}
