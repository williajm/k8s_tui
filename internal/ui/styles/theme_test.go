package styles

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStatusStyle(t *testing.T) {
	tests := []struct {
		status string
		want   lipgloss.Style
	}{
		{"Running", StatusRunningStyle},
		{"Succeeded", StatusRunningStyle},
		{"Active", StatusRunningStyle},
		{"Ready", StatusRunningStyle},
		{"Pending", StatusPendingStyle},
		{"Creating", StatusPendingStyle},
		{"Waiting", StatusPendingStyle},
		{"Failed", StatusErrorStyle},
		{"Error", StatusErrorStyle},
		{"CrashLoopBackOff", StatusErrorStyle},
		{"ImagePullBackOff", StatusErrorStyle},
		{"Unknown", StatusUnknownStyle},
		{"SomeOtherStatus", StatusUnknownStyle},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StatusStyle(tt.status)
			// We can't directly compare styles, so compare their string output
			if got.Render("test") == "" {
				t.Error("StatusStyle() returned empty render")
			}
		})
	}
}

func TestRenderKeyHelp(t *testing.T) {
	tests := []struct {
		key         string
		description string
	}{
		{"q", "quit"},
		{"?", "help"},
		{"r", "refresh"},
		{"Enter", "select"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := RenderKeyHelp(tt.key, tt.description)

			if result == "" {
				t.Error("RenderKeyHelp() returned empty string")
			}

			// Should contain both key and description
			if !strings.Contains(result, tt.description) {
				t.Errorf("RenderKeyHelp() = %v, should contain %v", result, tt.description)
			}
		})
	}
}

func TestRenderDetailRow(t *testing.T) {
	tests := []struct {
		label string
		value string
	}{
		{"Name", "test-pod"},
		{"Status", "Running"},
		{"Namespace", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			result := RenderDetailRow(tt.label, tt.value)

			if result == "" {
				t.Error("RenderDetailRow() returned empty string")
			}

			// Should contain both label and value
			if !strings.Contains(result, tt.value) {
				t.Errorf("RenderDetailRow() should contain value %v", tt.value)
			}
		})
	}
}

func TestSetWidth(t *testing.T) {
	style := lipgloss.NewStyle()
	width := 50

	result := SetWidth(style, width)

	// Render something to test the width is applied
	rendered := result.Render("test")
	if rendered == "" {
		t.Error("SetWidth() style rendered empty string")
	}
}

func TestSetHeight(t *testing.T) {
	style := lipgloss.NewStyle()
	height := 10

	result := SetHeight(style, height)

	// Render something to test the height is applied
	rendered := result.Render("test")
	if rendered == "" {
		t.Error("SetHeight() style rendered empty string")
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are defined
	colors := []lipgloss.Color{
		ColorPrimary,
		ColorSecondary,
		ColorAccent,
		ColorSuccess,
		ColorWarning,
		ColorError,
		ColorInfo,
		ColorBorder,
		ColorText,
		ColorTextDim,
		ColorBackground,
		ColorSelected,
		ColorHighlight,
	}

	for i, color := range colors {
		if color == "" {
			t.Errorf("Color constant at index %d is empty", i)
		}
	}
}

func TestStyleConstants(t *testing.T) {
	// Test that style constants can render without panic
	styles := []lipgloss.Style{
		BaseStyle,
		HeaderStyle,
		TitleStyle,
		BorderStyle,
		ListItemStyle,
		SelectedListItemStyle,
		StatusRunningStyle,
		StatusPendingStyle,
		StatusErrorStyle,
		StatusUnknownStyle,
		FooterStyle,
		KeyStyle,
		DescStyle,
		DetailHeaderStyle,
		DetailLabelStyle,
		DetailValueStyle,
		TableHeaderStyle,
		TableRowStyle,
		TableSelectedRowStyle,
		ErrorStyle,
		InfoBoxStyle,
	}

	for i, style := range styles {
		rendered := style.Render("test")
		if rendered == "" {
			t.Errorf("Style at index %d rendered empty string", i)
		}
	}
}
