package config

import (
	"fmt"
	"os"
	"time"

	"github.com/jacobalberty/jobwrapper/internal/filesystem"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	LockDir      string        `toml:"lock_dir"`
	Timeout      time.Duration `toml:"timeout"`
	LockFileName string        `toml:"lock_filename"`
	HistoryLines int           `toml:"history_lines"`
}

var DefaultConfig = Config{
	Timeout:      30 * time.Minute,
	LockFileName: ".lockfile",
	HistoryLines: 5,
}

func LoadConfig(fs filesystem.FileSystem) Config {
	config := DefaultConfig

	home, err := os.UserHomeDir()
	if err != nil {
		// Handle error if needed
		return config
	}

	path := fmt.Sprintf("%s/.jobwrapper/jobwrapper.conf", home)

	file, err := fs.Open(path)
	if err == nil {
		defer file.Close()

		decoder := toml.NewDecoder(file)
		_ = decoder.Decode(&config)
	}

	// Default lock directory if not provided
	if config.LockDir == "" {
		config.LockDir = ".jobwrapper"
	}
	if !isAbsPath(config.LockDir) {
		// If lock_dir is relative, prepend the home directory
		config.LockDir = fmt.Sprintf("%s/%s", home, config.LockDir)
	}
	return config
}

// isAbsPath checks if a path is absolute.
func isAbsPath(path string) bool {
	return path[0] == '/'
}
