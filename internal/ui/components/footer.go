package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// Footer represents the help/shortcut bar at the bottom
type Footer struct {
	width int
}

// NewFooter creates a new footer component
func NewFooter() *Footer {
	return &Footer{
		width: 80,
	}
}

// SetWidth sets the width of the footer
func (f *Footer) SetWidth(width int) {
	f.width = width
}

// View renders the footer with keyboard shortcuts
func (f *Footer) View() string {
	shortcuts := []string{
		styles.RenderKeyHelp("[↑↓]", "Navigate"),
		styles.RenderKeyHelp("[Enter]", "Select"),
		styles.RenderKeyHelp("[Esc]", "Back"),
		styles.RenderKeyHelp("[Tab]", "Switch"),
		styles.RenderKeyHelp("[n]", "Namespace"),
		styles.RenderKeyHelp("[l]", "Logs"),
		styles.RenderKeyHelp("[/]", "Search"),
		styles.RenderKeyHelp("[r]", "Refresh"),
		styles.RenderKeyHelp("[?]", "Help"),
		styles.RenderKeyHelp("[q]", "Quit"),
	}

	line1 := strings.Join(shortcuts[:5], "  ")
	line2 := strings.Join(shortcuts[5:], "  ")

	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.FooterStyle.Width(f.width).Render(line1),
		styles.FooterStyle.Width(f.width).Render(line2),
	)

	return footer
}

// ViewDetailed renders an extended help view with all shortcuts
func (f *Footer) ViewDetailed() string {
	categories := []struct {
		title     string
		shortcuts []string
	}{
		{
			title: "Navigation",
			shortcuts: []string{
				styles.RenderKeyHelp("↑/k", "Move up"),
				styles.RenderKeyHelp("↓/j", "Move down"),
				styles.RenderKeyHelp("PgUp/Ctrl+U", "Page up"),
				styles.RenderKeyHelp("PgDn/Ctrl+D", "Page down"),
				styles.RenderKeyHelp("g", "Go to top"),
				styles.RenderKeyHelp("G", "Go to bottom"),
			},
		},
		{
			title: "Selection",
			shortcuts: []string{
				styles.RenderKeyHelp("Enter/→/l", "Open detail/expand"),
				styles.RenderKeyHelp("←/h/Backspace", "Go back/collapse"),
				styles.RenderKeyHelp("Tab", "Switch panes"),
				styles.RenderKeyHelp("Shift+Tab", "Previous pane"),
			},
		},
		{
			title: "Actions",
			shortcuts: []string{
				styles.RenderKeyHelp("n", "Change namespace"),
				styles.RenderKeyHelp("c", "Change context"),
				styles.RenderKeyHelp("/", "Search/filter"),
				styles.RenderKeyHelp("r/F5", "Refresh"),
			},
		},
		{
			title: "Resource Actions",
			shortcuts: []string{
				styles.RenderKeyHelp("l", "View logs (pods)"),
				styles.RenderKeyHelp("d", "Describe resource"),
				styles.RenderKeyHelp("5", "Jump to Events tab"),
			},
		},
		{
			title: "Global",
			shortcuts: []string{
				styles.RenderKeyHelp("?", "Toggle help"),
				styles.RenderKeyHelp("q/Ctrl+C", "Quit application"),
			},
		},
	}

	var sections []string
	for _, cat := range categories {
		header := styles.DetailHeaderStyle.Render(cat.title)
		items := strings.Join(cat.shortcuts, "\n")
		section := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			items,
			"",
		)
		sections = append(sections, section)
	}

	content := strings.Join(sections, "\n")

	helpBox := styles.InfoBoxStyle.
		Width(f.width - 4).
		Render(content)

	return helpBox
}
