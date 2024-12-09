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

type HistoryWriter interface {
	MarkExecutionStart()
	MarkExecutionEnd()
	WriteHistory(err error) error
}

type historyJsonFileWriter struct {
	fs                 filesystem.FileSystem
	cfg                *config.Config
	exePath            string
	args               []string
	startTime          time.Time
	startExecutionTime *time.Time
	endExecutionTime   *time.Time
}

func (h *historyJsonFileWriter) MarkExecutionStart() {
	startTime := time.Now()
	h.startExecutionTime = &startTime
}

func (h *historyJsonFileWriter) MarkExecutionEnd() {
	endTime := time.Now()
	h.endExecutionTime = &endTime
}

func (h *historyJsonFileWriter) WriteHistory(err error) error {

	history := h.createLogEntry(err)

	if err := appendHistory(h.fs, filepath.Join(h.cfg.LockDir, filepath.Base(h.exePath)+".log"), history, h.cfg.HistoryLines); err != nil {
		return err
	}

	return nil
}

func NewHistoryWriter(fs filesystem.FileSystem, cfg *config.Config, exePath string, args []string) (HistoryWriter, error) {
	startTime := time.Now()
	exeName := filepath.Base(exePath)
	logPath := filepath.Join(cfg.LockDir, exeName+".log")

	// Ensure the log file exists
	file, err := fs.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	file.Close()

	return &historyJsonFileWriter{
		fs:        fs,
		cfg:       cfg,
		exePath:   exePath,
		args:      args,
		startTime: startTime,
	}, nil
}

func truncateArgs(args []string) []string {
	truncatedArgs := make([]string, len(args))
	for i, arg := range args {
		if len(arg) > maxArgLength {
			truncatedArgs[i] = arg[:maxArgLength]
		} else {
			truncatedArgs[i] = arg
		}
	}
	return truncatedArgs
}

func (h *historyJsonFileWriter) createLogEntry(err error) string {
	var (
		exeName       = filepath.Base(h.exePath)
		logBuffer     strings.Builder
		logArgs       []any
		truncatedArgs = truncateArgs(h.args)
	)
	logArgs = append(logArgs,
		"start", h.startTime.Format("2006-01-02 15:04:05"),
	)

	if h.startExecutionTime != nil {
		logArgs = append(logArgs,
			"wait_duration", h.startExecutionTime.Sub(h.startTime).String(),
			"start_execution", h.startExecutionTime.Format("2006-01-02 15:04:05"),
		)
	}
	if h.startExecutionTime != nil && h.endExecutionTime != nil {
		logArgs = append(logArgs,
			"end_execution", h.endExecutionTime.Format("2006-01-02 15:04:05"),
			"execution_duration", h.endExecutionTime.Sub(*h.startExecutionTime).String(),
		)
	}

	logArgs = append(logArgs,
		"executable", exeName,
		"args", strings.Join(truncatedArgs, " "),
		"executable_path", h.exePath,
		"error", err,
	)

	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	logger.Info("script execution",
		logArgs...,
	)
	return logBuffer.String()
}

func appendHistory(fs filesystem.FileSystem, logPath, history string, maxLines int) error {
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
	if err := scanner.Err(); err != nil {
		return err
	}
	lines = append(lines, history)
	if err := file.Close(); err != nil {
		return err
	}

	file, err = fs.OpenFile(logPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}
