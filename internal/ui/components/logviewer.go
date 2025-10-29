package components

import (
	"container/ring"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/models"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

const (
	maxLogBufferSize = 10000 // Maximum number of log lines to keep in memory
)

// LogViewer displays streaming pod logs with search and filtering capabilities
type LogViewer struct {
	logs           *ring.Ring // Circular buffer for log storage
	logCount       int        // Current number of logs in buffer
	viewport       viewport.Model
	searchMode     bool
	searchTerm     string
	following      bool
	container      string
	podName        string
	showTimestamps bool
	width          int
	height         int
	mu             sync.RWMutex
	isPrevious     bool
}

// NewLogViewer creates a new log viewer component
func NewLogViewer(podName, container string) *LogViewer {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	return &LogViewer{
		logs:           ring.New(maxLogBufferSize),
		logCount:       0,
		viewport:       vp,
		following:      true,
		container:      container,
		podName:        podName,
		showTimestamps: true,
		width:          80,
		height:         20,
	}
}

// SetSize sets the dimensions of the log viewer
func (l *LogViewer) SetSize(width, height int) {
	l.width = width
	l.height = height
	// Width: total - border sides (2) - horizontal padding (2) = 4
	l.viewport.Width = width - 4
	// Height: total - footer (2 lines outside box) - border (2) - header with border (2) = 6
	// Viewport gets remaining height inside the bordered container
	l.viewport.Height = height - 6
}

// AddLogEntry adds a new log entry to the buffer
func (l *LogViewer) AddLogEntry(entry models.LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Store the entry in the ring buffer
	l.logs.Value = entry
	l.logs = l.logs.Next()

	if l.logCount < maxLogBufferSize {
		l.logCount++
	}

	// Update viewport content if following
	if l.following {
		l.updateViewportContent()
	}
}

// AddLogEntries adds multiple log entries at once
func (l *LogViewer) AddLogEntries(entries []models.LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, entry := range entries {
		l.logs.Value = entry
		l.logs = l.logs.Next()

		if l.logCount < maxLogBufferSize {
			l.logCount++
		}
	}

	l.updateViewportContent()
}

// ToggleFollow toggles the follow mode
func (l *LogViewer) ToggleFollow() {
	l.following = !l.following
	if l.following {
		l.viewport.GotoBottom()
	}
}

// SetSearchMode enables or disables search mode
func (l *LogViewer) SetSearchMode(enabled bool) {
	l.searchMode = enabled
	if !enabled {
		l.searchTerm = ""
		l.updateViewportContent()
	}
}

// SetSearchTerm sets the search term and filters logs
func (l *LogViewer) SetSearchTerm(term string) {
	l.searchTerm = term
	l.updateViewportContent()
}

// ToggleTimestamps toggles timestamp display
func (l *LogViewer) ToggleTimestamps() {
	l.showTimestamps = !l.showTimestamps
	l.updateViewportContent()
}

// SetPreviousMode sets whether viewing previous container logs
func (l *LogViewer) SetPreviousMode(previous bool) {
	l.isPrevious = previous
}

// Update handles viewport updates
func (l *LogViewer) Update(msg tea.Msg) (*LogViewer, tea.Cmd) {
	var cmd tea.Cmd
	l.viewport, cmd = l.viewport.Update(msg)
	return l, cmd
}

// View renders the log viewer
func (l *LogViewer) View() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Build header
	header := l.renderHeader()

	// Render viewport
	viewportContent := l.viewport.View()

	// Combine header and viewport in bordered container
	logContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportContent,
	)

	borderedContent := styles.BorderStyle.
		Width(l.width).
		Height(l.height - 2). // Reserve 2 lines for footer outside border
		Render(logContent)

	// Build footer with status (rendered outside the bordered container)
	footer := l.renderFooter()

	// Combine bordered content with footer below it
	return lipgloss.JoinVertical(
		lipgloss.Left,
		borderedContent,
		footer,
	)
}

// renderHeader renders the log viewer header
func (l *LogViewer) renderHeader() string {
	title := fmt.Sprintf("Logs: %s", l.podName)
	if l.container != "" {
		title += fmt.Sprintf(" [%s]", l.container)
	}
	if l.isPrevious {
		title += " (Previous)"
	}

	headerStyle := styles.TableHeaderStyle.Width(l.width - 4)
	return headerStyle.Render(title)
}

// renderFooter renders the log viewer footer with status
func (l *LogViewer) renderFooter() string {
	var statusParts []string

	// Follow status
	if l.following {
		statusParts = append(statusParts, "Following")
	} else {
		statusParts = append(statusParts, "Paused")
	}

	// Log count
	statusParts = append(statusParts, fmt.Sprintf("Lines: %d", l.logCount))

	// Search indicator
	if l.searchMode {
		statusParts = append(statusParts, fmt.Sprintf("Search: %s_", l.searchTerm))
	} else if l.searchTerm != "" {
		statusParts = append(statusParts, fmt.Sprintf("Filter: %s", l.searchTerm))
	}

	// Status line
	statusLine := strings.Join(statusParts, " | ")

	// Keyboard shortcuts with proper styling
	shortcuts := []string{
		styles.RenderKeyHelp("[↑↓]", "Scroll"),
		styles.RenderKeyHelp("[f]", "Follow"),
		styles.RenderKeyHelp("[/]", "Search"),
		styles.RenderKeyHelp("[t]", "Timestamps"),
		styles.RenderKeyHelp("[q/Esc]", "Back"),
		styles.RenderKeyHelp("[Ctrl+C]", "Quit"),
	}
	shortcutsLine := strings.Join(shortcuts, "  ")

	// Combine status and shortcuts
	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.FooterStyle.Width(l.width).Render(statusLine),
		styles.FooterStyle.Width(l.width).Render(shortcutsLine),
	)

	return footer
}

// updateViewportContent updates the viewport with current logs
func (l *LogViewer) updateViewportContent() {
	var lines []string

	// Collect logs from ring buffer
	var startRing *ring.Ring
	if l.logCount < maxLogBufferSize {
		// If buffer not full, start from the beginning
		startRing = l.logs.Move(-l.logCount)
	} else {
		// If buffer full, we're at the oldest entry
		startRing = l.logs
	}

	// Iterate through logs
	count := l.logCount
	if count > maxLogBufferSize {
		count = maxLogBufferSize
	}

	for i := 0; i < count; i++ {
		if startRing.Value != nil {
			entry := startRing.Value.(models.LogEntry)

			// Apply search filter if active
			if l.searchTerm != "" {
				if !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(l.searchTerm)) {
					startRing = startRing.Next()
					continue
				}
			}

			// Format log line
			line := models.FormatLogEntry(entry, l.showTimestamps)

			// Highlight search term if present
			if l.searchTerm != "" && l.searchTerm != "_" {
				line = highlightText(line, l.searchTerm)
			}

			// Color by log level
			line = colorizeLogLevel(line, entry.Level)

			lines = append(lines, line)
		}
		startRing = startRing.Next()
	}

	// Set viewport content
	content := strings.Join(lines, "\n")
	l.viewport.SetContent(content)

	// Auto-scroll to bottom if following
	if l.following {
		l.viewport.GotoBottom()
	}
}

// highlightText highlights search terms in the text
func highlightText(text, term string) string {
	if term == "" {
		return text
	}

	// Simple case-insensitive highlighting
	lower := strings.ToLower(text)
	lowerTerm := strings.ToLower(term)

	var result strings.Builder
	lastIdx := 0

	for {
		idx := strings.Index(lower[lastIdx:], lowerTerm)
		if idx == -1 {
			result.WriteString(text[lastIdx:])
			break
		}

		idx += lastIdx
		result.WriteString(text[lastIdx:idx])

		// Highlight the matched term
		highlightStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("0"))
		result.WriteString(highlightStyle.Render(text[idx : idx+len(term)]))

		lastIdx = idx + len(term)
	}

	return result.String()
}

// colorizeLogLevel applies color based on log level
func colorizeLogLevel(line string, level models.LogLevel) string {
	var style lipgloss.Style

	switch level {
	case models.LogLevelError:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	case models.LogLevelWarn:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	case models.LogLevelDebug:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray
	default:
		return line // Info level - no color
	}

	return style.Render(line)
}

// GetViewport returns the viewport for external updates
func (l *LogViewer) GetViewport() *viewport.Model {
	return &l.viewport
}

// Clear clears all logs from the buffer
func (l *LogViewer) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logs = ring.New(maxLogBufferSize)
	l.logCount = 0
	l.updateViewportContent()
}
