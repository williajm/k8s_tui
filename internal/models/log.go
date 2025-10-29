package models

import (
	"strings"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
	LogLevelDebug
)

// LogEntry represents a single log line
type LogEntry struct {
	Timestamp time.Time
	Container string
	Message   string
	Level     LogLevel
}

// LogOptions configures how logs are fetched and displayed
type LogOptions struct {
	Follow     bool
	TailLines  int64
	Timestamps bool
	Previous   bool
	Container  string
	SinceTime  *time.Time
}

// DefaultLogOptions returns sensible defaults for log viewing
func DefaultLogOptions() LogOptions {
	return LogOptions{
		Follow:     true,
		TailLines:  100,
		Timestamps: true,
		Previous:   false,
		Container:  "",
	}
}

// DetectLogLevel attempts to determine the log level from the message content
func DetectLogLevel(message string) LogLevel {
	lower := strings.ToLower(message)

	// Check for error indicators
	if strings.Contains(lower, "error") || strings.Contains(lower, "err:") ||
		strings.Contains(lower, "fatal") || strings.Contains(lower, "panic") {
		return LogLevelError
	}

	// Check for warning indicators
	if strings.Contains(lower, "warn") || strings.Contains(lower, "warning") {
		return LogLevelWarn
	}

	// Check for debug indicators
	if strings.Contains(lower, "debug") || strings.Contains(lower, "trace") {
		return LogLevelDebug
	}

	// Default to info
	return LogLevelInfo
}

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// FormatLogEntry formats a log entry for display
func FormatLogEntry(entry LogEntry, showTimestamp bool) string {
	var parts []string

	if showTimestamp && !entry.Timestamp.IsZero() {
		parts = append(parts, entry.Timestamp.Format("15:04:05.000"))
	}

	if entry.Container != "" {
		parts = append(parts, "["+entry.Container+"]")
	}

	parts = append(parts, entry.Message)

	return strings.Join(parts, " ")
}
