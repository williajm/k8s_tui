package components

import (
	"fmt"

	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// ConnectionState represents the current connection state
type ConnectionState string

const (
	ConnectionStateDisconnected ConnectionState = "Disconnected"
	ConnectionStateConnecting   ConnectionState = "Connecting"
	ConnectionStateConnected    ConnectionState = "Connected"
	ConnectionStateReconnecting ConnectionState = "Reconnecting"
	ConnectionStateError        ConnectionState = "Error"
)

// Header represents the application header
type Header struct {
	context         string
	namespace       string
	connected       bool
	connectionState ConnectionState
	width           int
}

// NewHeader creates a new header component
func NewHeader(context, namespace string, connected bool) *Header {
	state := ConnectionStateConnected
	if !connected {
		state = ConnectionStateDisconnected
	}
	return &Header{
		context:         context,
		namespace:       namespace,
		connected:       connected,
		connectionState: state,
		width:           80,
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
	if connected {
		h.connectionState = ConnectionStateConnected
	} else {
		h.connectionState = ConnectionStateDisconnected
	}
}

// SetConnectionState updates the connection state with more detail
func (h *Header) SetConnectionState(state ConnectionState) {
	h.connectionState = state
	h.connected = (state == ConnectionStateConnected || state == ConnectionStateConnecting || state == ConnectionStateReconnecting)
}

// SetWidth sets the width of the header
func (h *Header) SetWidth(width int) {
	h.width = width
}

// View renders the header
func (h *Header) View() string {
	// Create connection indicator based on state
	var connIndicator string
	var connStatus string

	switch h.connectionState {
	case ConnectionStateConnected:
		connIndicator = "◉"
		connStatus = "Connected"
	case ConnectionStateConnecting:
		connIndicator = "◌"
		connStatus = "Connecting..."
	case ConnectionStateReconnecting:
		connIndicator = "◎"
		connStatus = "Reconnecting..."
	case ConnectionStateError:
		connIndicator = "⊗"
		connStatus = "Error"
	default: // ConnectionStateDisconnected
		connIndicator = "○"
		connStatus = "Disconnected"
	}

	// Build header sections - use plain strings, let HeaderStyle handle all styling
	title := "K8S-TUI"
	contextInfo := fmt.Sprintf("Context: %s", h.context)
	nsInfo := fmt.Sprintf("NS: %s", h.namespace)
	connInfo := fmt.Sprintf("%s %s", connIndicator, connStatus)

	// Calculate spacing
	separator := " | "
	contentWidth := len(title) + len(contextInfo) + len(nsInfo) + len(connInfo) + (len(separator) * 3)

	padding := h.width - contentWidth - 2 // Subtract 2 for HeaderStyle padding
	if padding < 0 {
		padding = 0
	}

	// Build the header line as plain text
	headerContent := title + separator + contextInfo + separator + nsInfo + separator + connInfo

	// Add padding spaces
	for i := 0; i < padding; i++ {
		headerContent += " "
	}

	// Apply header style with explicit height - this handles ALL the styling consistently
	return styles.HeaderStyle.
		Width(h.width).
		Height(1).
		Render(headerContent)
}
