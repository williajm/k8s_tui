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
//
//nolint:dupl // Theme definitions intentionally have similar structure with different values
func DarkColorScheme() ColorScheme {
	return ColorScheme{
		Primary:    lipgloss.Color("#326CE5"), // Kubernetes blue
		Secondary:  lipgloss.Color("#5B8DEF"), // Lighter blue
		Accent:     lipgloss.Color("#FFCC00"), // Bright yellow/gold
		Success:    lipgloss.Color("#50FA7B"), // Bright green (Dracula-inspired)
		Warning:    lipgloss.Color("#FFB86C"), // Orange (Dracula-inspired)
		Error:      lipgloss.Color("#FF5555"), // Red (Dracula-inspired)
		Info:       lipgloss.Color("#8BE9FD"), // Cyan (Dracula-inspired)
		Border:     lipgloss.Color("#6272A4"), // Muted blue-gray
		Text:       lipgloss.Color("#F8F8F2"), // Off-white
		TextDim:    lipgloss.Color("#6272A4"), // Muted blue-gray
		Background: lipgloss.Color("#282A36"), // Dark purple-gray
		Selected:   lipgloss.Color("#44475A"), // Lighter purple-gray
		Highlight:  lipgloss.Color("#BD93F9"), // Purple (Dracula-inspired)
	}
}

// LightColorScheme returns the light color scheme
//
//nolint:dupl // Theme definitions intentionally have similar structure with different values
func LightColorScheme() ColorScheme {
	return ColorScheme{
		Primary:    lipgloss.Color("#326CE5"), // Kubernetes blue
		Secondary:  lipgloss.Color("#1A4FC9"), // Darker blue
		Accent:     lipgloss.Color("#B45309"), // Amber/brown
		Success:    lipgloss.Color("#16A34A"), // Green
		Warning:    lipgloss.Color("#EA580C"), // Orange
		Error:      lipgloss.Color("#DC2626"), // Red
		Info:       lipgloss.Color("#0891B2"), // Cyan
		Border:     lipgloss.Color("#CBD5E1"), // Light slate
		Text:       lipgloss.Color("#1E293B"), // Dark slate
		TextDim:    lipgloss.Color("#64748B"), // Slate
		Background: lipgloss.Color("#F8FAFC"), // Very light gray
		Selected:   lipgloss.Color("#E2E8F0"), // Light slate
		Highlight:  lipgloss.Color("#7C3AED"), // Purple
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
