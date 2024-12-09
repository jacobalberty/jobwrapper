package history

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jacobalberty/jobwrapper/internal/config"
	"github.com/jacobalberty/jobwrapper/internal/filesystem"
)

const maxArgLength = 15

func WriteHistory(fs filesystem.FileSystem, cfg *config.Config, exePath string, maxLines int, args []string) (func() error, error) {
	startTime := time.Now()
	exeName := filepath.Base(exePath)
	logPath := filepath.Join(cfg.LockDir, exeName+".log")

	file, err := fs.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	file.Close()

	return func() error {
		truncatedArgs := make([]string, len(args))
		for i, arg := range args {
			if len(arg) > maxArgLength {
				truncatedArgs[i] = arg[:maxArgLength]
			} else {
				truncatedArgs[i] = arg
			}
		}
		argsStr := strings.Join(truncatedArgs, " ")

		stopTime := time.Now()
		duration := stopTime.Sub(startTime)
		var logBuffer strings.Builder
		logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
		logger.Info("script execution",
			"start", startTime.Format("2006-01-02 15:04:05"),
			"stop", stopTime.Format("2006-01-02 15:04:05"),
			"duration", duration.String(),
			"executable", exeName,
			"args", argsStr,
			"executable_path", exePath,
		)
		history := logBuffer.String()

		file, err := fs.OpenFile(logPath, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		lines := []string{}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text()+"\n")
			if len(lines) > maxLines {
				lines = lines[1:]
			}
		}
		lines = append(lines, history)
		file.Close()

		file, err = fs.OpenFile(logPath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		for _, line := range lines {
			_, err := file.Write([]byte(line))
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}
