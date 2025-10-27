package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// Tab represents a single tab
type Tab struct {
	Title string
	ID    int
}

// Tabs represents a tab navigation component
type Tabs struct {
	tabs      []Tab
	activeTab int
	width     int
}

// NewTabs creates a new tabs component
func NewTabs() *Tabs {
	return &Tabs{
		tabs: []Tab{
			{Title: "Pods", ID: 0},
			{Title: "Services", ID: 1},
			{Title: "Deployments", ID: 2},
			{Title: "StatefulSets", ID: 3},
		},
		activeTab: 0,
		width:     80,
	}
}

// SetWidth sets the width of the tabs component
func (t *Tabs) SetWidth(width int) {
	t.width = width
}

// GetActiveTab returns the currently active tab ID
func (t *Tabs) GetActiveTab() int {
	return t.activeTab
}

// SetActiveTab sets the active tab
func (t *Tabs) SetActiveTab(tabID int) {
	if tabID >= 0 && tabID < len(t.tabs) {
		t.activeTab = tabID
	}
}

// NextTab switches to the next tab
func (t *Tabs) NextTab() {
	t.activeTab = (t.activeTab + 1) % len(t.tabs)
}

// PrevTab switches to the previous tab
func (t *Tabs) PrevTab() {
	t.activeTab = (t.activeTab - 1 + len(t.tabs)) % len(t.tabs)
}

// View renders the tabs
func (t *Tabs) View() string {
	var renderedTabs []string

	for _, tab := range t.tabs {
		var style lipgloss.Style

		if tab.ID == t.activeTab {
			style = styles.ActiveTabStyle
		} else {
			style = styles.InactiveTabStyle
		}

		renderedTabs = append(renderedTabs, style.Render(tab.Title))
	}

	// Join tabs horizontally
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Add border
	return styles.TabBorderStyle.
		Width(t.width).
		Render(tabsRow)
}
