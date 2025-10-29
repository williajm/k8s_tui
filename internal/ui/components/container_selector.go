package components

import (
	"github.com/charmbracelet/lipgloss"
)

// ContainerSelector is a specialized selector for choosing pod containers
type ContainerSelector struct {
	*Selector
}

// NewContainerSelector creates a new container selector
func NewContainerSelector(containers []string) *ContainerSelector {
	selector := NewSelector("Select Container")
	selector.SetOptions(containers)

	return &ContainerSelector{
		Selector: selector,
	}
}

// ViewWithInfo renders the container selector with additional information
func (c *ContainerSelector) ViewWithInfo(info string) string {
	if !c.visible {
		return ""
	}

	// Render the basic selector
	selectorView := c.Selector.View()

	// Add info message above the selector if provided
	if info != "" {
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
		infoView := infoStyle.Render(info)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			infoView,
			selectorView,
		)
	}

	return selectorView
}

// GetSelectedContainerName returns the selected container name
// Strips the "(init)" suffix if present
func (c *ContainerSelector) GetSelectedContainerName() string {
	selected := c.GetSelected()
	if selected == "" {
		return ""
	}

	// Remove "(init)" suffix if present
	if len(selected) > 7 && selected[len(selected)-7:] == " (init)" {
		return selected[:len(selected)-7]
	}

	return selected
}
