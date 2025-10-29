package components

import (
	"strings"
	"testing"
)

func TestNewHeader(t *testing.T) {
	tests := []struct {
		name      string
		context   string
		namespace string
		connected bool
	}{
		{
			name:      "connected header",
			context:   "production",
			namespace: "default",
			connected: true,
		},
		{
			name:      "disconnected header",
			context:   "dev",
			namespace: "kube-system",
			connected: false,
		},
		{
			name:      "empty values",
			context:   "",
			namespace: "",
			connected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeader(tt.context, tt.namespace, tt.connected)

			if h == nil {
				t.Fatal("NewHeader returned nil")
			}

			if h.context != tt.context {
				t.Errorf("expected context %q, got %q", tt.context, h.context)
			}

			if h.namespace != tt.namespace {
				t.Errorf("expected namespace %q, got %q", tt.namespace, h.namespace)
			}

			if h.connected != tt.connected {
				t.Errorf("expected connected %v, got %v", tt.connected, h.connected)
			}

			if h.width != 80 {
				t.Errorf("expected default width 80, got %d", h.width)
			}
		})
	}
}

func TestHeader_SetContext(t *testing.T) {
	h := NewHeader("initial", "default", true)

	testContext := "new-context"
	h.SetContext(testContext)

	if h.context != testContext {
		t.Errorf("expected context %q, got %q", testContext, h.context)
	}
}

func TestHeader_SetNamespace(t *testing.T) {
	h := NewHeader("context", "initial", true)

	testNamespace := "kube-system"
	h.SetNamespace(testNamespace)

	if h.namespace != testNamespace {
		t.Errorf("expected namespace %q, got %q", testNamespace, h.namespace)
	}
}

func TestHeader_SetConnected(t *testing.T) {
	tests := []struct {
		name      string
		initial   bool
		newStatus bool
	}{
		{
			name:      "connected to disconnected",
			initial:   true,
			newStatus: false,
		},
		{
			name:      "disconnected to connected",
			initial:   false,
			newStatus: true,
		},
		{
			name:      "remain connected",
			initial:   true,
			newStatus: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeader("ctx", "ns", tt.initial)
			h.SetConnected(tt.newStatus)

			if h.connected != tt.newStatus {
				t.Errorf("expected connected %v, got %v", tt.newStatus, h.connected)
			}
		})
	}
}

func TestHeader_SetWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard width",
			width: 120,
		},
		{
			name:  "small width",
			width: 40,
		},
		{
			name:  "large width",
			width: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeader("ctx", "ns", true)
			h.SetWidth(tt.width)

			if h.width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, h.width)
			}
		})
	}
}

func TestHeader_View(t *testing.T) {
	tests := []struct {
		name            string
		context         string
		namespace       string
		connected       bool
		width           int
		expectedStrings []string
	}{
		{
			name:      "connected header",
			context:   "production",
			namespace: "default",
			connected: true,
			width:     120,
			expectedStrings: []string{
				"K8S-TUI",
				"Context: production",
				"NS: default",
				"Connected",
				"◉",
			},
		},
		{
			name:      "disconnected header",
			context:   "dev",
			namespace: "kube-system",
			connected: false,
			width:     120,
			expectedStrings: []string{
				"K8S-TUI",
				"Context: dev",
				"NS: kube-system",
				"Disconnected",
				"○",
			},
		},
		{
			name:      "small width",
			context:   "ctx",
			namespace: "ns",
			connected: true,
			width:     40,
			expectedStrings: []string{
				"K8S-TUI",
				"Context: ctx",
				"NS: ns",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeader(tt.context, tt.namespace, tt.connected)
			h.SetWidth(tt.width)

			view := h.View()

			if view == "" {
				t.Fatal("View returned empty string")
			}

			for _, expected := range tt.expectedStrings {
				if !strings.Contains(view, expected) {
					t.Errorf("expected view to contain %q, but it didn't\nView: %s", expected, view)
				}
			}
		})
	}
}

func TestHeader_ViewConnectionStatus(t *testing.T) {
	tests := []struct {
		name              string
		connected         bool
		expectedIndicator string
		expectedStatus    string
	}{
		{
			name:              "connected shows filled circle",
			connected:         true,
			expectedIndicator: "◉",
			expectedStatus:    "Connected",
		},
		{
			name:              "disconnected shows empty circle",
			connected:         false,
			expectedIndicator: "○",
			expectedStatus:    "Disconnected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeader("ctx", "ns", tt.connected)
			view := h.View()

			if !strings.Contains(view, tt.expectedIndicator) {
				t.Errorf("expected indicator %q in view", tt.expectedIndicator)
			}

			if !strings.Contains(view, tt.expectedStatus) {
				t.Errorf("expected status %q in view", tt.expectedStatus)
			}
		})
	}
}

func TestHeader_UpdateAndView(t *testing.T) {
	h := NewHeader("initial", "default", true)

	// Update all fields
	h.SetContext("new-context")
	h.SetNamespace("new-namespace")
	h.SetConnected(false)
	h.SetWidth(150)

	view := h.View()

	expectedStrings := []string{
		"new-context",
		"new-namespace",
		"Disconnected",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(view, expected) {
			t.Errorf("expected updated view to contain %q", expected)
		}
	}
}

func TestHeader_SetConnectionState(t *testing.T) {
	tests := []struct {
		name              string
		state             ConnectionState
		expectedConnected bool
		expectedIndicator string
		expectedStatus    string
	}{
		{
			name:              "Connected state",
			state:             ConnectionStateConnected,
			expectedConnected: true,
			expectedIndicator: "◉",
			expectedStatus:    "Connected",
		},
		{
			name:              "Connecting state",
			state:             ConnectionStateConnecting,
			expectedConnected: true,
			expectedIndicator: "◌",
			expectedStatus:    "Connecting...",
		},
		{
			name:              "Reconnecting state",
			state:             ConnectionStateReconnecting,
			expectedConnected: true,
			expectedIndicator: "◎",
			expectedStatus:    "Reconnecting...",
		},
		{
			name:              "Error state",
			state:             ConnectionStateError,
			expectedConnected: false,
			expectedIndicator: "⊗",
			expectedStatus:    "Error",
		},
		{
			name:              "Disconnected state",
			state:             ConnectionStateDisconnected,
			expectedConnected: false,
			expectedIndicator: "○",
			expectedStatus:    "Disconnected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeader("ctx", "ns", true)
			h.SetConnectionState(tt.state)

			// Check connected field
			if h.connected != tt.expectedConnected {
				t.Errorf("Expected connected=%v, got %v", tt.expectedConnected, h.connected)
			}

			// Check state field
			if h.connectionState != tt.state {
				t.Errorf("Expected connectionState=%v, got %v", tt.state, h.connectionState)
			}

			// Check view rendering
			view := h.View()
			if !strings.Contains(view, tt.expectedIndicator) {
				t.Errorf("Expected indicator %q in view", tt.expectedIndicator)
			}
			if !strings.Contains(view, tt.expectedStatus) {
				t.Errorf("Expected status %q in view", tt.expectedStatus)
			}
		})
	}
}

func TestHeader_ConnectionStateTransitions(t *testing.T) {
	h := NewHeader("ctx", "ns", true)

	// Test transition from Connected to Reconnecting
	h.SetConnectionState(ConnectionStateConnected)
	if h.connectionState != ConnectionStateConnected {
		t.Errorf("Expected ConnectionStateConnected, got %v", h.connectionState)
	}

	h.SetConnectionState(ConnectionStateReconnecting)
	if h.connectionState != ConnectionStateReconnecting {
		t.Errorf("Expected ConnectionStateReconnecting, got %v", h.connectionState)
	}
	if !h.connected {
		t.Error("Expected connected=true for reconnecting state")
	}

	// Test transition to Error
	h.SetConnectionState(ConnectionStateError)
	if h.connectionState != ConnectionStateError {
		t.Errorf("Expected ConnectionStateError, got %v", h.connectionState)
	}
	if h.connected {
		t.Error("Expected connected=false for error state")
	}

	// Test recovery to Connecting
	h.SetConnectionState(ConnectionStateConnecting)
	if h.connectionState != ConnectionStateConnecting {
		t.Errorf("Expected ConnectionStateConnecting, got %v", h.connectionState)
	}
	if !h.connected {
		t.Error("Expected connected=true for connecting state")
	}
}

func TestHeader_AllConnectionStatesRender(t *testing.T) {
	// Ensure all connection states render without panicking
	states := []ConnectionState{
		ConnectionStateConnected,
		ConnectionStateConnecting,
		ConnectionStateReconnecting,
		ConnectionStateError,
		ConnectionStateDisconnected,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			h := NewHeader("ctx", "ns", true)
			h.SetConnectionState(state)

			view := h.View()
			if view == "" {
				t.Error("View() returned empty string")
			}

			// Check that view contains expected elements
			if !strings.Contains(view, "K8S-TUI") {
				t.Error("View missing K8S-TUI title")
			}
			if !strings.Contains(view, "Context: ctx") {
				t.Error("View missing context")
			}
			if !strings.Contains(view, "NS: ns") {
				t.Error("View missing namespace")
			}
		})
	}
}
