package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// Header represents the application header
type Header struct {
	context   string
	namespace string
	connected bool
	width     int
}

// NewHeader creates a new header component
func NewHeader(context, namespace string, connected bool) *Header {
	return &Header{
		context:   context,
		namespace: namespace,
		connected: connected,
		width:     80,
	}
}

// SetContext updates the context
func (h *Header) SetContext(context string) {
	h.context = context
}

// SetNamespace updates the namespace
func (h *Header) SetNamespace(namespace string) {
	h.namespace = namespace
}

// SetConnected updates the connection status
func (h *Header) SetConnected(connected bool) {
	h.connected = connected
}

// SetWidth sets the width of the header
func (h *Header) SetWidth(width int) {
	h.width = width
}

// View renders the header
func (h *Header) View() string {
	// Create connection indicator
	connIndicator := "◉"
	connStatus := "Connected"
	connStyle := styles.StatusRunningStyle

	if !h.connected {
		connIndicator = "○"
		connStatus = "Disconnected"
		connStyle = styles.StatusErrorStyle
	}

	// Build header sections
	title := styles.TitleStyle.Render("K8S-TUI")
	contextInfo := fmt.Sprintf("Context: %s", h.context)
	nsInfo := fmt.Sprintf("NS: %s", h.namespace)
	connInfo := connStyle.Render(fmt.Sprintf("%s %s", connIndicator, connStatus))

	// Calculate spacing
	separator := " | "
	contentWidth := lipgloss.Width(title) +
		lipgloss.Width(contextInfo) +
		lipgloss.Width(nsInfo) +
		lipgloss.Width(connInfo) +
		(len(separator) * 3)

	padding := h.width - contentWidth
	if padding < 0 {
		padding = 0
	}

	// Build the header line
	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		separator,
		contextInfo,
		separator,
		nsInfo,
		separator,
		connInfo,
		lipgloss.NewStyle().Width(padding).Render(""),
	)

	// Apply header style
	return styles.HeaderStyle.
		Width(h.width).
		Render(headerContent)
}
