package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// Selector represents a selection dialog for choosing from a list of options
type Selector struct {
	title       string
	options     []string
	selectedIdx int
	width       int
	height      int
	visible     bool
}

// NewSelector creates a new selector component
func NewSelector(title string) *Selector {
	return &Selector{
		title:       title,
		options:     []string{},
		selectedIdx: 0,
		width:       40,
		height:      15,
		visible:     false,
	}
}

// SetOptions sets the available options
func (s *Selector) SetOptions(options []string) {
	s.options = options
	if s.selectedIdx >= len(s.options) {
		s.selectedIdx = 0
	}
}

// SetSize sets the dimensions
func (s *Selector) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// Show shows the selector
func (s *Selector) Show() {
	s.visible = true
}

// Hide hides the selector
func (s *Selector) Hide() {
	s.visible = false
}

// IsVisible returns whether the selector is visible
func (s *Selector) IsVisible() bool {
	return s.visible
}

// MoveUp moves the selection up
func (s *Selector) MoveUp() {
	if s.selectedIdx > 0 {
		s.selectedIdx--
	}
}

// MoveDown moves the selection down
func (s *Selector) MoveDown() {
	if s.selectedIdx < len(s.options)-1 {
		s.selectedIdx++
	}
}

// GetSelected returns the currently selected option
func (s *Selector) GetSelected() string {
	if s.selectedIdx >= 0 && s.selectedIdx < len(s.options) {
		return s.options[s.selectedIdx]
	}
	return ""
}

// View renders the selector
func (s *Selector) View() string {
	if !s.visible || len(s.options) == 0 {
		return ""
	}

	// Build title
	title := styles.DetailHeaderStyle.Render(s.title)

	// Build options list
	var optionRows []string
	maxVisible := s.height - 4 // Account for title, borders, padding

	startIdx := s.selectedIdx - maxVisible/2
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(s.options) {
		endIdx = len(s.options)
		startIdx = endIdx - maxVisible
		if startIdx < 0 {
			startIdx = 0
		}
	}

	for i := startIdx; i < endIdx; i++ {
		option := s.options[i]
		if len(option) > s.width-10 {
			option = option[:s.width-13] + "..."
		}

		if i == s.selectedIdx {
			optionRows = append(optionRows, styles.SelectedListItemStyle.Render("▸ "+option))
		} else {
			optionRows = append(optionRows, styles.ListItemStyle.Render("  "+option))
		}
	}

	optionsList := strings.Join(optionRows, "\n")

	// Build help text
	helpText := styles.FooterStyle.Render("↑↓ navigate • enter select • esc cancel")

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		optionsList,
		"",
		helpText,
	)

	// Add border and center
	box := styles.BorderStyle.
		Width(s.width).
		Height(s.height).
		Render(content)

	// Center the box on screen (this will be handled by the parent component)
	return box
}
