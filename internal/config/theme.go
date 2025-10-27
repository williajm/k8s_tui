package config

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ThemeType represents the theme type
type ThemeType string

const (
	// ThemeDark is the dark theme
	ThemeDark ThemeType = "dark"
	// ThemeLight is the light theme
	ThemeLight ThemeType = "light"
	// ThemeAuto automatically selects theme based on terminal
	ThemeAuto ThemeType = "auto"
)

// ColorScheme represents a color scheme for the application
type ColorScheme struct {
	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// UI colors
	Border     lipgloss.Color
	Text       lipgloss.Color
	TextDim    lipgloss.Color
	Background lipgloss.Color
	Selected   lipgloss.Color
	Highlight  lipgloss.Color
}

// DarkColorScheme returns the dark color scheme
func DarkColorScheme() ColorScheme {
	return ColorScheme{
		Primary:    lipgloss.Color("#00ADD8"), // Kubernetes blue
		Secondary:  lipgloss.Color("#326CE5"), // Darker blue
		Accent:     lipgloss.Color("#FFA500"), // Orange
		Success:    lipgloss.Color("#00C853"), // Green
		Warning:    lipgloss.Color("#FFC107"), // Yellow
		Error:      lipgloss.Color("#F44336"), // Red
		Info:       lipgloss.Color("#2196F3"), // Blue
		Border:     lipgloss.Color("#555555"),
		Text:       lipgloss.Color("#FFFFFF"),
		TextDim:    lipgloss.Color("#888888"),
		Background: lipgloss.Color("#1A1A1A"),
		Selected:   lipgloss.Color("#2A2A2A"),
		Highlight:  lipgloss.Color("#00ADD8"),
	}
}

// LightColorScheme returns the light color scheme
func LightColorScheme() ColorScheme {
	return ColorScheme{
		Primary:    lipgloss.Color("#0066CC"), // Darker blue for light background
		Secondary:  lipgloss.Color("#0052A3"), // Even darker blue
		Accent:     lipgloss.Color("#FF8C00"), // Dark orange
		Success:    lipgloss.Color("#00A040"), // Darker green
		Warning:    lipgloss.Color("#F57C00"), // Darker yellow/orange
		Error:      lipgloss.Color("#D32F2F"), // Darker red
		Info:       lipgloss.Color("#1976D2"), // Darker blue
		Border:     lipgloss.Color("#CCCCCC"),
		Text:       lipgloss.Color("#000000"),
		TextDim:    lipgloss.Color("#666666"),
		Background: lipgloss.Color("#FFFFFF"),
		Selected:   lipgloss.Color("#E0E0E0"),
		Highlight:  lipgloss.Color("#0066CC"),
	}
}

// GetColorScheme returns the appropriate color scheme based on theme type
func GetColorScheme(themeType ThemeType) (ColorScheme, error) {
	switch themeType {
	case ThemeDark:
		return DarkColorScheme(), nil
	case ThemeLight:
		return LightColorScheme(), nil
	case ThemeAuto:
		// For auto mode, we default to dark theme
		// In a real implementation, this would detect terminal background
		return DarkColorScheme(), nil
	default:
		return ColorScheme{}, fmt.Errorf("invalid theme type: %s", themeType)
	}
}

// ApplyColorScheme applies a color scheme to create styled components
//
//nolint:funlen // Style initialization requires many assignments
func (cs ColorScheme) ApplyColorScheme() Styles {
	return Styles{
		Base: lipgloss.NewStyle().
			Foreground(cs.Text).
			Background(cs.Background),

		Header: lipgloss.NewStyle().
			Foreground(cs.Text).
			Background(cs.Primary).
			Bold(true).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(cs.Primary).
			Bold(true),

		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(cs.Border).
			Padding(0, 1),

		ListItem: lipgloss.NewStyle().
			Padding(0, 1),

		SelectedListItem: lipgloss.NewStyle().
			Background(cs.Selected).
			Foreground(cs.Highlight).
			Bold(true).
			Padding(0, 1),

		StatusRunning: lipgloss.NewStyle().
			Foreground(cs.Success).
			Bold(true),

		StatusPending: lipgloss.NewStyle().
			Foreground(cs.Warning),

		StatusError: lipgloss.NewStyle().
			Foreground(cs.Error).
			Bold(true),

		StatusUnknown: lipgloss.NewStyle().
			Foreground(cs.TextDim),

		Footer: lipgloss.NewStyle().
			Foreground(cs.TextDim).
			Padding(0, 1),

		Key: lipgloss.NewStyle().
			Foreground(cs.Accent).
			Bold(true),

		Desc: lipgloss.NewStyle().
			Foreground(cs.TextDim),

		DetailHeader: lipgloss.NewStyle().
			Foreground(cs.Primary).
			Bold(true).
			Underline(true),

		DetailLabel: lipgloss.NewStyle().
			Foreground(cs.TextDim).
			Width(15).
			Align(lipgloss.Right),

		DetailValue: lipgloss.NewStyle().
			Foreground(cs.Text),

		TableHeader: lipgloss.NewStyle().
			Foreground(cs.Primary).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(cs.Border),

		TableRow: lipgloss.NewStyle().
			Foreground(cs.Text),

		TableSelectedRow: lipgloss.NewStyle().
			Background(cs.Selected).
			Foreground(cs.Highlight).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(cs.Error).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.Error),

		InfoBox: lipgloss.NewStyle().
			Foreground(cs.Info).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.Info),

		ActiveTab: lipgloss.NewStyle().
			Foreground(cs.Text).
			Background(cs.Primary).
			Bold(true).
			Padding(0, 2),

		InactiveTab: lipgloss.NewStyle().
			Foreground(cs.TextDim).
			Background(cs.Background).
			Padding(0, 2),

		TabBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(cs.Border).
			Padding(0, 1),

		ColorScheme: cs,
	}
}

// Styles contains all styled components for the application
type Styles struct {
	Base             lipgloss.Style
	Header           lipgloss.Style
	Title            lipgloss.Style
	Border           lipgloss.Style
	ListItem         lipgloss.Style
	SelectedListItem lipgloss.Style
	StatusRunning    lipgloss.Style
	StatusPending    lipgloss.Style
	StatusError      lipgloss.Style
	StatusUnknown    lipgloss.Style
	Footer           lipgloss.Style
	Key              lipgloss.Style
	Desc             lipgloss.Style
	DetailHeader     lipgloss.Style
	DetailLabel      lipgloss.Style
	DetailValue      lipgloss.Style
	TableHeader      lipgloss.Style
	TableRow         lipgloss.Style
	TableSelectedRow lipgloss.Style
	Error            lipgloss.Style
	InfoBox          lipgloss.Style
	ActiveTab        lipgloss.Style
	InactiveTab      lipgloss.Style
	TabBorder        lipgloss.Style
	ColorScheme      ColorScheme
}

// StatusStyle returns the appropriate style for a given status
func (s Styles) StatusStyle(status string) lipgloss.Style {
	switch status {
	case "Running", "Succeeded", "Active", "Ready":
		return s.StatusRunning
	case "Pending", "Creating", "Waiting":
		return s.StatusPending
	case "Failed", "Error", "CrashLoopBackOff", "ImagePullBackOff":
		return s.StatusError
	default:
		return s.StatusUnknown
	}
}

// RenderKeyHelp renders a key-description pair for help text
func (s Styles) RenderKeyHelp(key, description string) string {
	return s.Key.Render(key) + " " + s.Desc.Render(description)
}

// RenderDetailRow renders a label-value pair for detail views
func (s Styles) RenderDetailRow(label, value string) string {
	return s.DetailLabel.Render(label+":") + " " + s.DetailValue.Render(value)
}
