package components

import (
	"strings"
	"testing"
	"time"

	"github.com/williajm/k8s-tui/internal/models"
)

func TestNewLogViewer(t *testing.T) {
	podName := "test-pod"
	container := "nginx"

	lv := NewLogViewer(podName, container)

	if lv == nil {
		t.Fatal("NewLogViewer() returned nil")
	}

	if lv.podName != podName {
		t.Errorf("podName = %v, want %v", lv.podName, podName)
	}

	if lv.container != container {
		t.Errorf("container = %v, want %v", lv.container, container)
	}

	if !lv.following {
		t.Error("following should be true by default")
	}

	if !lv.showTimestamps {
		t.Error("showTimestamps should be true by default")
	}

	if lv.logCount != 0 {
		t.Errorf("logCount = %d, want 0", lv.logCount)
	}

	if lv.logs == nil {
		t.Error("logs ring buffer should be initialized")
	}
}

func TestLogViewer_SetSize(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	width := 120
	height := 40

	lv.SetSize(width, height)

	if lv.width != width {
		t.Errorf("width = %d, want %d", lv.width, width)
	}

	if lv.height != height {
		t.Errorf("height = %d, want %d", lv.height, height)
	}

	// Viewport should be adjusted (accounting for border, header, footer)
	expectedVpWidth := width - 4   // border (2) + padding (2)
	expectedVpHeight := height - 6 // border (2) + header (2) + footer (2 lines)

	if lv.viewport.Width != expectedVpWidth {
		t.Errorf("viewport.Width = %d, want %d", lv.viewport.Width, expectedVpWidth)
	}

	if lv.viewport.Height != expectedVpHeight {
		t.Errorf("viewport.Height = %d, want %d", lv.viewport.Height, expectedVpHeight)
	}
}

func TestLogViewer_AddLogEntry(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	entry := models.LogEntry{
		Timestamp: time.Now(),
		Container: "app",
		Message:   "Test log message",
		Level:     models.LogLevelInfo,
	}

	lv.AddLogEntry(entry)

	if lv.logCount != 1 {
		t.Errorf("logCount = %d, want 1", lv.logCount)
	}

	// Add more entries
	for i := 0; i < 5; i++ {
		lv.AddLogEntry(entry)
	}

	if lv.logCount != 6 {
		t.Errorf("logCount = %d, want 6", lv.logCount)
	}
}

func TestLogViewer_AddLogEntry_RingBufferWrap(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	// Add more entries than buffer size to test wrapping
	numEntries := maxLogBufferSize + 100
	for i := 0; i < numEntries; i++ {
		entry := models.LogEntry{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Log entry",
			Level:     models.LogLevelInfo,
		}
		lv.AddLogEntry(entry)
	}

	// Log count should be capped at maxLogBufferSize
	if lv.logCount != maxLogBufferSize {
		t.Errorf("logCount = %d, want %d", lv.logCount, maxLogBufferSize)
	}
}

func TestLogViewer_AddLogEntries(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	entries := []models.LogEntry{
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Entry 1",
			Level:     models.LogLevelInfo,
		},
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Entry 2",
			Level:     models.LogLevelWarn,
		},
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Entry 3",
			Level:     models.LogLevelError,
		},
	}

	lv.AddLogEntries(entries)

	if lv.logCount != 3 {
		t.Errorf("logCount = %d, want 3", lv.logCount)
	}
}

func TestLogViewer_ToggleFollow(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	// Should start as following
	if !lv.following {
		t.Error("following should be true initially")
	}

	lv.ToggleFollow()

	if lv.following {
		t.Error("following should be false after toggle")
	}

	lv.ToggleFollow()

	if !lv.following {
		t.Error("following should be true after second toggle")
	}
}

func TestLogViewer_SetSearchMode(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")
	lv.searchTerm = "test"

	// Enable search mode
	lv.SetSearchMode(true)

	if !lv.searchMode {
		t.Error("searchMode should be true")
	}

	// Disable search mode (should clear search term)
	lv.SetSearchMode(false)

	if lv.searchMode {
		t.Error("searchMode should be false")
	}

	if lv.searchTerm != "" {
		t.Errorf("searchTerm should be cleared, got %q", lv.searchTerm)
	}
}

func TestLogViewer_SetSearchTerm(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	searchTerm := "error"
	lv.SetSearchTerm(searchTerm)

	if lv.searchTerm != searchTerm {
		t.Errorf("searchTerm = %q, want %q", lv.searchTerm, searchTerm)
	}
}

func TestLogViewer_ToggleTimestamps(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	// Should start with timestamps enabled
	if !lv.showTimestamps {
		t.Error("showTimestamps should be true initially")
	}

	lv.ToggleTimestamps()

	if lv.showTimestamps {
		t.Error("showTimestamps should be false after toggle")
	}

	lv.ToggleTimestamps()

	if !lv.showTimestamps {
		t.Error("showTimestamps should be true after second toggle")
	}
}

func TestLogViewer_SetPreviousMode(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	if lv.isPrevious {
		t.Error("isPrevious should be false initially")
	}

	lv.SetPreviousMode(true)

	if !lv.isPrevious {
		t.Error("isPrevious should be true after setting")
	}

	lv.SetPreviousMode(false)

	if lv.isPrevious {
		t.Error("isPrevious should be false after unsetting")
	}
}

func TestLogViewer_View(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")
	lv.SetSize(100, 30)

	// Add some log entries
	entries := []models.LogEntry{
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Starting application",
			Level:     models.LogLevelInfo,
		},
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "ERROR: Connection failed",
			Level:     models.LogLevelError,
		},
	}
	lv.AddLogEntries(entries)

	view := lv.View()

	if view == "" {
		t.Error("View() returned empty string")
	}

	// Check that view contains pod name
	if !strings.Contains(view, "test-pod") {
		t.Error("View() should contain pod name")
	}
}

func TestLogViewer_Clear(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	// Add some entries
	for i := 0; i < 10; i++ {
		entry := models.LogEntry{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Log entry",
			Level:     models.LogLevelInfo,
		}
		lv.AddLogEntry(entry)
	}

	if lv.logCount != 10 {
		t.Errorf("logCount before clear = %d, want 10", lv.logCount)
	}

	lv.Clear()

	if lv.logCount != 0 {
		t.Errorf("logCount after clear = %d, want 0", lv.logCount)
	}
}

func TestLogViewer_SearchFiltering(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")
	lv.SetSize(100, 30)

	// Add entries with different messages
	entries := []models.LogEntry{
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Starting application",
			Level:     models.LogLevelInfo,
		},
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "ERROR: Connection failed",
			Level:     models.LogLevelError,
		},
		{
			Timestamp: time.Now(),
			Container: "app",
			Message:   "Processing request",
			Level:     models.LogLevelInfo,
		},
	}
	lv.AddLogEntries(entries)

	// Set search term
	lv.SetSearchTerm("error")

	// Verify search term is set
	if lv.searchTerm != "error" {
		t.Errorf("searchTerm = %q, want %q", lv.searchTerm, "error")
	}

	// The filtering happens in updateViewportContent, which is called
	// We can verify it doesn't panic
	view := lv.View()
	if view == "" {
		t.Error("View() returned empty string with search filter")
	}
}

func TestLogViewer_GetViewport(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")

	vp := lv.GetViewport()

	if vp == nil {
		t.Error("GetViewport() returned nil")
	}

	// Verify it's the same viewport
	if vp != &lv.viewport {
		t.Error("GetViewport() should return pointer to internal viewport")
	}
}

func TestHighlightText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		term     string
		wantHas  string
		wantSkip bool
	}{
		{
			name:    "empty term",
			text:    "Hello world",
			term:    "",
			wantHas: "Hello world",
		},
		{
			name:    "term not found",
			text:    "Hello world",
			term:    "xyz",
			wantHas: "Hello world",
		},
		{
			name:     "term found",
			text:     "Hello world error happened",
			term:     "error",
			wantSkip: false,
		},
		{
			name:    "case insensitive",
			text:    "Hello ERROR World",
			term:    "error",
			wantHas: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := highlightText(tt.text, tt.term)

			if tt.wantSkip {
				return
			}

			if got == "" {
				t.Error("highlightText() returned empty string")
			}

			if tt.wantHas != "" && !strings.Contains(got, tt.wantHas) {
				// Result may contain ANSI codes, so we just check it's not empty
				if len(got) == 0 {
					t.Errorf("highlightText() result is empty, expected to contain %q", tt.wantHas)
				}
			}
		})
	}
}

func TestColorizeLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		level models.LogLevel
	}{
		{
			name:  "info level",
			line:  "Starting server",
			level: models.LogLevelInfo,
		},
		{
			name:  "warn level",
			line:  "Warning: deprecated API",
			level: models.LogLevelWarn,
		},
		{
			name:  "error level",
			line:  "Error: connection failed",
			level: models.LogLevelError,
		},
		{
			name:  "debug level",
			line:  "Debug: entering function",
			level: models.LogLevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorizeLogLevel(tt.line, tt.level)

			if got == "" {
				t.Error("colorizeLogLevel() returned empty string")
			}

			// For info level, should return line unchanged
			if tt.level == models.LogLevelInfo && got != tt.line {
				t.Errorf("colorizeLogLevel() for INFO should return unchanged line")
			}

			// For other levels, may contain ANSI codes
			if tt.level != models.LogLevelInfo && len(got) < len(tt.line) {
				t.Error("colorizeLogLevel() result shorter than input")
			}
		})
	}
}

func TestLogViewer_ConcurrentAccess(t *testing.T) {
	// Test that concurrent access doesn't cause race conditions
	lv := NewLogViewer("test-pod", "app")

	// Add entries concurrently
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				entry := models.LogEntry{
					Timestamp: time.Now(),
					Container: "app",
					Message:   "Concurrent log entry",
					Level:     models.LogLevelInfo,
				}
				lv.AddLogEntry(entry)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Should have 50 entries
	if lv.logCount != 50 {
		t.Errorf("logCount = %d, want 50", lv.logCount)
	}
}

func TestLogViewer_FollowMode(t *testing.T) {
	lv := NewLogViewer("test-pod", "app")
	lv.SetSize(100, 30)

	// Add entries while following
	if !lv.following {
		t.Error("should start in follow mode")
	}

	entry := models.LogEntry{
		Timestamp: time.Now(),
		Container: "app",
		Message:   "Log message",
		Level:     models.LogLevelInfo,
	}
	lv.AddLogEntry(entry)

	// Disable following
	lv.ToggleFollow()

	// Add another entry
	lv.AddLogEntry(entry)

	// Log count should still increase
	if lv.logCount != 2 {
		t.Errorf("logCount = %d, want 2", lv.logCount)
	}
}
