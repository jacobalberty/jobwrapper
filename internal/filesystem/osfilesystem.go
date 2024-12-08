package filesystem

import (
	"io"
	"os"
)

// OSFileSystem implements the FileSystem interface using the os package
type OSFileSystem struct{}

func (fs OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs OSFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (fs OSFileSystem) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	return os.OpenFile(name, flag, perm)
}

func (fs OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}
