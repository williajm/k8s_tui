package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Logger interface for debug logging
type Logger interface {
	Log(format string, args ...interface{})
	LogWithSample(message string, content string)
	LogKeyPress(key string, runes string)
	LogResize(component string, oldW, oldH, newW, newH int)
}

// FileLogger writes debug logs to a file
type FileLogger struct {
	file  *os.File
	mutex sync.Mutex
}

// NoOpLogger is a no-op implementation
type NoOpLogger struct{}

var (
	globalLogger Logger = &NoOpLogger{}
	ansiPattern  = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// InitLogger initializes the global logger
func InitLogger(enabled bool) error {
	if !enabled {
		globalLogger = &NoOpLogger{}
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".k8s-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, "debug.log")
	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	globalLogger = &FileLogger{file: file}
	globalLogger.Log("=== Debug logging started ===")
	return nil
}

// GetLogger returns the global logger
func GetLogger() Logger {
	return globalLogger
}

// CloseLogger closes the logger
func CloseLogger() error {
	if fl, ok := globalLogger.(*FileLogger); ok {
		fl.mutex.Lock()
		defer fl.mutex.Unlock()
		return fl.file.Close()
	}
	return nil
}

// Log writes a formatted log message with timestamp
func (l *FileLogger) Log(format string, args ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.file, "[%s] %s\n", timestamp, message)
	l.file.Sync()
}

// LogWithSample writes a log message with ANSI analysis and content samples
func (l *FileLogger) LogWithSample(message string, content string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	timestamp := time.Now().Format("15:04:05.000")

	// Count ANSI escape sequences
	ansiCount := len(ansiPattern.FindAllString(content, -1))

	// Get length
	length := len(content)

	// Sample first and last 200 characters
	sampleSize := 200
	var firstSample, lastSample string

	if length <= sampleSize*2 {
		firstSample = content
		lastSample = ""
	} else {
		firstSample = content[:sampleSize]
		lastSample = content[length-sampleSize:]
	}

	// Escape samples for readability
	firstSample = strings.ReplaceAll(firstSample, "\n", "\\n")
	firstSample = strings.ReplaceAll(firstSample, "\t", "\\t")
	lastSample = strings.ReplaceAll(lastSample, "\n", "\\n")
	lastSample = strings.ReplaceAll(lastSample, "\t", "\\t")

	fmt.Fprintf(l.file, "[%s] %s\n", timestamp, message)
	fmt.Fprintf(l.file, "  Length: %d bytes, ANSI sequences: %d\n", length, ansiCount)
	fmt.Fprintf(l.file, "  First 200 chars: \"%s\"\n", firstSample)
	if lastSample != "" {
		fmt.Fprintf(l.file, "  Last 200 chars: \"%s\"\n", lastSample)
	}
	l.file.Sync()
}

// LogKeyPress logs a key press event
func (l *FileLogger) LogKeyPress(key string, runes string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	timestamp := time.Now().Format("15:04:05.000")
	if runes != "" {
		fmt.Fprintf(l.file, "[%s] KeyPress: key=%s, runes=%s\n", timestamp, key, runes)
	} else {
		fmt.Fprintf(l.file, "[%s] KeyPress: key=%s\n", timestamp, key)
	}
	l.file.Sync()
}

// LogResize logs a component resize event
func (l *FileLogger) LogResize(component string, oldW, oldH, newW, newH int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(l.file, "[%s] Resize %s: (%dx%d) -> (%dx%d)\n",
		timestamp, component, oldW, oldH, newW, newH)
	l.file.Sync()
}

// NoOpLogger implementations (do nothing)
func (l *NoOpLogger) Log(format string, args ...interface{})                      {}
func (l *NoOpLogger) LogWithSample(message string, content string)                {}
func (l *NoOpLogger) LogKeyPress(key string, runes string)                        {}
func (l *NoOpLogger) LogResize(component string, oldW, oldH, newW, newH int)     {}
