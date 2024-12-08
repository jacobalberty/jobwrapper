package history

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jacobalberty/jobwrapper/internal/filesystem"
)

const maxArgLength = 15

func WriteHistory(fs filesystem.FileSystem, filePath string, maxLines int, args []string) (func() error, error) {
	startTime := time.Now()

	file, err := fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
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

		history := fmt.Sprintf("%s - Duration: %s - Args: %s\n", startTime.Format("2006-01-02 15:04:05"), time.Since(startTime), argsStr)

		file, err := fs.OpenFile(filePath, os.O_RDWR, 0644)
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

		file, err = fs.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
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
