package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme colors (Dracula-inspired for better aesthetics)
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#326CE5") // Kubernetes blue
	ColorSecondary = lipgloss.Color("#5B8DEF") // Lighter blue
	ColorAccent    = lipgloss.Color("#FFCC00") // Bright yellow/gold

	// Status colors
	ColorSuccess = lipgloss.Color("#50FA7B") // Bright green
	ColorWarning = lipgloss.Color("#FFB86C") // Orange
	ColorError   = lipgloss.Color("#FF5555") // Red
	ColorInfo    = lipgloss.Color("#8BE9FD") // Cyan

	// UI colors
	ColorBorder     = lipgloss.Color("#6272A4") // Muted blue-gray
	ColorText       = lipgloss.Color("#F8F8F2") // Off-white
	ColorTextDim    = lipgloss.Color("#6272A4") // Muted blue-gray
	ColorBackground = lipgloss.Color("#282A36") // Dark purple-gray
	ColorSelected   = lipgloss.Color("#44475A") // Lighter purple-gray
	ColorHighlight  = lipgloss.Color("#BD93F9") // Purple
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
