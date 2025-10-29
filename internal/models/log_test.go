package models

import (
	"testing"
	"time"
)

func TestDetectLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    LogLevel
	}{
		{
			name:    "error level - error keyword",
			message: "Error: failed to connect to database",
			want:    LogLevelError,
		},
		{
			name:    "error level - err: prefix",
			message: "err: connection timeout",
			want:    LogLevelError,
		},
		{
			name:    "error level - fatal keyword",
			message: "FATAL: system shutdown",
			want:    LogLevelError,
		},
		{
			name:    "error level - panic keyword",
			message: "panic: runtime error",
			want:    LogLevelError,
		},
		{
			name:    "warn level - warn keyword",
			message: "Warning: deprecated API usage",
			want:    LogLevelWarn,
		},
		{
			name:    "warn level - warning keyword",
			message: "WARNING: high memory usage",
			want:    LogLevelWarn,
		},
		{
			name:    "debug level - debug keyword",
			message: "DEBUG: entering function foo()",
			want:    LogLevelDebug,
		},
		{
			name:    "debug level - trace keyword",
			message: "trace: processing request",
			want:    LogLevelDebug,
		},
		{
			name:    "info level - no special keywords",
			message: "Request processed successfully",
			want:    LogLevelInfo,
		},
		{
			name:    "info level - empty message",
			message: "",
			want:    LogLevelInfo,
		},
		{
			name:    "case insensitive - mixed case error",
			message: "ErRoR: something went wrong",
			want:    LogLevelError,
		},
		{
			name:    "case insensitive - uppercase warn",
			message: "WARN: check this out",
			want:    LogLevelWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLogLevel(tt.message)
			if got != tt.want {
				t.Errorf("DetectLogLevel(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		want  string
	}{
		{
			name:  "info level",
			level: LogLevelInfo,
			want:  "INFO",
		},
		{
			name:  "warn level",
			level: LogLevelWarn,
			want:  "WARN",
		},
		{
			name:  "error level",
			level: LogLevelError,
			want:  "ERROR",
		},
		{
			name:  "debug level",
			level: LogLevelDebug,
			want:  "DEBUG",
		},
		{
			name:  "unknown level",
			level: LogLevel(999),
			want:  "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("LogLevel(%d).String() = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestFormatLogEntry(t *testing.T) {
	fixedTime := time.Date(2024, 10, 28, 15, 30, 45, 123456789, time.UTC)

	tests := []struct {
		name          string
		entry         LogEntry
		showTimestamp bool
		want          string
	}{
		{
			name: "with timestamp and container",
			entry: LogEntry{
				Timestamp: fixedTime,
				Container: "nginx",
				Message:   "Starting server",
				Level:     LogLevelInfo,
			},
			showTimestamp: true,
			want:          "15:30:45.123 [nginx] Starting server",
		},
		{
			name: "without timestamp",
			entry: LogEntry{
				Timestamp: fixedTime,
				Container: "nginx",
				Message:   "Starting server",
				Level:     LogLevelInfo,
			},
			showTimestamp: false,
			want:          "[nginx] Starting server",
		},
		{
			name: "without container",
			entry: LogEntry{
				Timestamp: fixedTime,
				Container: "",
				Message:   "Log message",
				Level:     LogLevelInfo,
			},
			showTimestamp: true,
			want:          "15:30:45.123 Log message",
		},
		{
			name: "zero timestamp ignored",
			entry: LogEntry{
				Timestamp: time.Time{},
				Container: "app",
				Message:   "Message",
				Level:     LogLevelInfo,
			},
			showTimestamp: true,
			want:          "[app] Message",
		},
		{
			name: "message only",
			entry: LogEntry{
				Timestamp: time.Time{},
				Container: "",
				Message:   "Simple message",
				Level:     LogLevelInfo,
			},
			showTimestamp: false,
			want:          "Simple message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatLogEntry(tt.entry, tt.showTimestamp)
			if got != tt.want {
				t.Errorf("FormatLogEntry() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultLogOptions(t *testing.T) {
	opts := DefaultLogOptions()

	if !opts.Follow {
		t.Error("DefaultLogOptions().Follow should be true")
	}

	if opts.TailLines != 100 {
		t.Errorf("DefaultLogOptions().TailLines = %d, want 100", opts.TailLines)
	}

	if !opts.Timestamps {
		t.Error("DefaultLogOptions().Timestamps should be true")
	}

	if opts.Previous {
		t.Error("DefaultLogOptions().Previous should be false")
	}

	if opts.Container != "" {
		t.Errorf("DefaultLogOptions().Container = %q, want empty string", opts.Container)
	}

	if opts.SinceTime != nil {
		t.Error("DefaultLogOptions().SinceTime should be nil")
	}
}

func TestLogEntry_FieldValidation(t *testing.T) {
	// Test that LogEntry can be created with all valid types
	entry := LogEntry{
		Timestamp: time.Now(),
		Container: "test-container",
		Message:   "test message",
		Level:     LogLevelInfo,
	}

	if entry.Container != "test-container" {
		t.Errorf("LogEntry.Container = %q, want %q", entry.Container, "test-container")
	}

	if entry.Message != "test message" {
		t.Errorf("LogEntry.Message = %q, want %q", entry.Message, "test message")
	}

	if entry.Level != LogLevelInfo {
		t.Errorf("LogEntry.Level = %v, want %v", entry.Level, LogLevelInfo)
	}
}

func TestLogOptions_FieldValidation(t *testing.T) {
	// Test that LogOptions can be created with all valid types
	sinceTime := time.Now().Add(-1 * time.Hour)

	opts := LogOptions{
		Follow:     true,
		TailLines:  50,
		Timestamps: false,
		Previous:   true,
		Container:  "app",
		SinceTime:  &sinceTime,
	}

	if !opts.Follow {
		t.Error("LogOptions.Follow should be true")
	}

	if opts.TailLines != 50 {
		t.Errorf("LogOptions.TailLines = %d, want 50", opts.TailLines)
	}

	if opts.Timestamps {
		t.Error("LogOptions.Timestamps should be false")
	}

	if !opts.Previous {
		t.Error("LogOptions.Previous should be true")
	}

	if opts.Container != "app" {
		t.Errorf("LogOptions.Container = %q, want %q", opts.Container, "app")
	}

	if opts.SinceTime == nil {
		t.Error("LogOptions.SinceTime should not be nil")
	}
}

func TestLogLevel_Coverage(t *testing.T) {
	// Test all log levels for completeness
	levels := []struct {
		level LogLevel
		str   string
	}{
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelDebug, "DEBUG"},
	}

	for _, tc := range levels {
		t.Run(tc.str, func(t *testing.T) {
			if tc.level.String() != tc.str {
				t.Errorf("LogLevel(%d).String() = %v, want %v", tc.level, tc.level.String(), tc.str)
			}
		})
	}
}
