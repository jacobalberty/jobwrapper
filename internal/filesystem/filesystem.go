package filesystem

import (
	"io"
	"os"
)

// FileSystem defines an interface for abstracting filesystem operations
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	Open(name string) (io.ReadCloser, error)
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)
	Remove(name string) error
}
