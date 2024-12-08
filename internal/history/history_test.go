package history

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jacobalberty/jobwrapper/internal/filesystem"
)

func TestWriteHistory(t *testing.T) {
	mockFS := &filesystem.MockFileSystem{Files: make(map[string]*string)}
	filePath := "test_history.log"
	maxLines := 5

	writeFunc, err := WriteHistory(mockFS, filePath, maxLines, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	time.Sleep(1 * time.Second) // Simulate some duration

	err = writeFunc()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	file, err := mockFS.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	lines := bytes.Split(content, []byte("\n"))
	if len(lines) > maxLines+1 { // +1 because of the trailing newline
		t.Fatalf("expected lines to be <= %d, got %d", maxLines+1, len(lines))
	}

	expectedPrefix := time.Now().Format("2006-01-02")
	if !bytes.Contains(content, []byte(expectedPrefix)) {
		t.Fatalf("expected content to contain prefix %s, got %s", expectedPrefix, string(content))
	}
}
