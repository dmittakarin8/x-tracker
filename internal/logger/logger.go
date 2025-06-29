package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	enabled  bool
	logDir   string
	mu       sync.Mutex
	file     *os.File
	filename string
}

var (
	instance *Logger
	once     sync.Once
)

// Initialize creates a new logger instance
func Initialize(enabled bool, logDir string) error {
	var err error
	once.Do(func() {
		instance = &Logger{
			enabled: enabled,
			logDir:  logDir,
		}
		err = instance.rotateFile()
	})
	return err
}

// Info logs an info level message
func Info(format string, args ...interface{}) {
	if instance == nil || !instance.enabled {
		return
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	// Check if we need to rotate to a new day's file
	currentFile := time.Now().Format("2006-01-02") + ".log"
	if currentFile != instance.filename {
		if err := instance.rotateFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Error rotating log file: %v\n", err)
			return
		}
	}

	// Format the log message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [INFO] %s\n", timestamp, msg)

	// Write to file
	if _, err := instance.file.WriteString(logLine); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
	}
}

// rotateFile creates a new log file for the current day
func (l *Logger) rotateFile() error {
	// Close existing file if open
	if l.file != nil {
		l.file.Close()
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("creating log directory: %w", err)
	}

	// Open new file
	l.filename = time.Now().Format("2006-01-02") + ".log"
	filepath := filepath.Join(l.logDir, l.filename)
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}

	l.file = file
	return nil
}

// Close closes the current log file
func Close() error {
	if instance != nil && instance.file != nil {
		return instance.file.Close()
	}
	return nil
} 