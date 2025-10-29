package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme colors
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#00ADD8") // Kubernetes blue
	ColorSecondary = lipgloss.Color("#326CE5") // Darker blue
	ColorAccent    = lipgloss.Color("#FFA500") // Orange

	// Status colors
	ColorSuccess = lipgloss.Color("#00C853") // Green
	ColorWarning = lipgloss.Color("#FFC107") // Yellow
	ColorError   = lipgloss.Color("#F44336") // Red
	ColorInfo    = lipgloss.Color("#2196F3") // Blue

	// UI colors
	ColorBorder     = lipgloss.Color("#555555")
	ColorText       = lipgloss.Color("#FFFFFF")
	ColorTextDim    = lipgloss.Color("#888888")
	ColorBackground = lipgloss.Color("#1A1A1A")
	ColorSelected   = lipgloss.Color("#2A2A2A")
	ColorHighlight  = lipgloss.Color("#00ADD8")
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBackground)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// List item styles
	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Bold(true).
				Padding(0, 1)

	// Status styles
	StatusRunningStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StatusPendingStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	StatusUnknownStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim)

	// Footer/help styles
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Padding(0, 1)

	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim)

	// Detail view styles
	DetailHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Underline(true)

	DetailLabelStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim).
				Width(15).
				Align(lipgloss.Right)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(ColorText)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorBorder)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	TableSelectedRowStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Bold(true)

	// Error styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError)

	// Info box styles
	InfoBoxStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim).
				Padding(0, 2)

	TabBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorBorder).
			Padding(0, 1)
)

// Helper functions

// StatusStyle returns the appropriate style for a given status
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "Running", "Succeeded", "Active", "Ready":
		return StatusRunningStyle
	case "Pending", "Creating", "Waiting":
		return StatusPendingStyle
	case "Failed", "Error", "CrashLoopBackOff", "ImagePullBackOff":
		return StatusErrorStyle
	default:
		return StatusUnknownStyle
	}
}

// RenderKeyHelp renders a key-description pair for help text
func RenderKeyHelp(key, description string) string {
	return KeyStyle.Render(key) + " " + DescStyle.Render(description)
}

// RenderDetailRow renders a label-value pair for detail views
func RenderDetailRow(label, value string) string {
	return DetailLabelStyle.Render(label+":") + " " + DetailValueStyle.Render(value)
}

// Width and height helpers
func SetWidth(style lipgloss.Style, width int) lipgloss.Style {
	return style.Width(width)
}

func SetHeight(style lipgloss.Style, height int) lipgloss.Style {
	return style.Height(height)
}
