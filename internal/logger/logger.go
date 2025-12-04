package logger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ProcessLogger handles logging for a process
type ProcessLogger struct {
	mu       sync.Mutex
	id       string
	name     string
	logDir   string
	stdout   *os.File
	stderr   *os.File
	combined *os.File
}

// NewProcessLogger creates a new process logger
func NewProcessLogger(id, name, logDir string) (*ProcessLogger, error) {
	processLogDir := filepath.Join(logDir, fmt.Sprintf("%s-%s", name, id))
	if err := os.MkdirAll(processLogDir, 0755); err != nil {
		return nil, err
	}

	stdout, err := os.OpenFile(
		filepath.Join(processLogDir, "stdout.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	stderr, err := os.OpenFile(
		filepath.Join(processLogDir, "stderr.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		stdout.Close()
		return nil, err
	}

	combined, err := os.OpenFile(
		filepath.Join(processLogDir, "combined.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		stdout.Close()
		stderr.Close()
		return nil, err
	}

	return &ProcessLogger{
		id:       id,
		name:     name,
		logDir:   processLogDir,
		stdout:   stdout,
		stderr:   stderr,
		combined: combined,
	}, nil
}

// Log writes a log entry
func (l *ProcessLogger) Log(logType, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("[%s] %s\n", timestamp, message)

	switch logType {
	case "stdout":
		l.stdout.WriteString(line)
		l.combined.WriteString(fmt.Sprintf("[%s] [OUT] %s\n", timestamp, message))
	case "stderr":
		l.stderr.WriteString(line)
		l.combined.WriteString(fmt.Sprintf("[%s] [ERR] %s\n", timestamp, message))
	}
}

// GetLogs reads recent log entries
func (l *ProcessLogger) GetLogs(lines int, logType string) ([]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var logFile string
	switch logType {
	case "stdout":
		logFile = filepath.Join(l.logDir, "stdout.log")
	case "stderr":
		logFile = filepath.Join(l.logDir, "stderr.log")
	default:
		logFile = filepath.Join(l.logDir, "combined.log")
	}

	return readLastLines(logFile, lines)
}

// Close closes all log files
func (l *ProcessLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var err error
	if l.stdout != nil {
		if e := l.stdout.Close(); e != nil {
			err = e
		}
	}
	if l.stderr != nil {
		if e := l.stderr.Close(); e != nil {
			err = e
		}
	}
	if l.combined != nil {
		if e := l.combined.Close(); e != nil {
			err = e
		}
	}
	return err
}

// readLastLines reads the last n lines from a file
func readLastLines(path string, n int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if n <= 0 || n >= len(lines) {
		return lines, nil
	}

	return lines[len(lines)-n:], nil
}

// RotateLogs rotates log files if they exceed the size limit
func (l *ProcessLogger) RotateLogs(maxSizeMB int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	maxSize := int64(maxSizeMB * 1024 * 1024)

	for _, logName := range []string{"stdout.log", "stderr.log", "combined.log"} {
		logPath := filepath.Join(l.logDir, logName)
		info, err := os.Stat(logPath)
		if err != nil {
			continue
		}

		if info.Size() > maxSize {
			// Rotate the file
			rotatedPath := fmt.Sprintf("%s.%s", logPath, time.Now().Format("20060102-150405"))
			if err := os.Rename(logPath, rotatedPath); err != nil {
				return err
			}

			// Recreate the log file
			newFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}

			switch logName {
			case "stdout.log":
				l.stdout.Close()
				l.stdout = newFile
			case "stderr.log":
				l.stderr.Close()
				l.stderr = newFile
			case "combined.log":
				l.combined.Close()
				l.combined = newFile
			}
		}
	}

	return nil
}
