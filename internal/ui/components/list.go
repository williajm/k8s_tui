package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/models"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// PodList represents a list of pods
type PodList struct {
	pods         []models.PodInfo
	selectedIdx  int
	viewportTop  int
	width        int
	height       int
	searchFilter string
}

// NewPodList creates a new pod list component
func NewPodList() *PodList {
	return &PodList{
		pods:        []models.PodInfo{},
		selectedIdx: 0,
		viewportTop: 0,
		width:       80,
		height:      20,
	}
}

// SetPods updates the list of pods
func (l *PodList) SetPods(pods []models.PodInfo) {
	l.pods = pods
	// Reset selection if out of bounds
	if l.selectedIdx >= len(l.pods) {
		l.selectedIdx = 0
	}
}

// SetSize sets the dimensions
func (l *PodList) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// SetSearchFilter sets the search filter
func (l *PodList) SetSearchFilter(filter string) {
	l.searchFilter = filter
}

// MoveUp moves the selection up
func (l *PodList) MoveUp() {
	if l.selectedIdx > 0 {
		l.selectedIdx--
		l.adjustViewport()
	}
}

// MoveDown moves the selection down
func (l *PodList) MoveDown() {
	if l.selectedIdx < len(l.pods)-1 {
		l.selectedIdx++
		l.adjustViewport()
	}
}

// PageUp moves up by one page
func (l *PodList) PageUp() {
	l.selectedIdx -= l.height - 3 // Account for header
	if l.selectedIdx < 0 {
		l.selectedIdx = 0
	}
	l.adjustViewport()
}

// PageDown moves down by one page
func (l *PodList) PageDown() {
	l.selectedIdx += l.height - 3
	if l.selectedIdx >= len(l.pods) {
		l.selectedIdx = len(l.pods) - 1
	}
	l.adjustViewport()
}

// Home moves to the top
func (l *PodList) Home() {
	l.selectedIdx = 0
	l.viewportTop = 0
}

// End moves to the bottom
func (l *PodList) End() {
	l.selectedIdx = len(l.pods) - 1
	l.adjustViewport()
}

// GetSelected returns the currently selected pod
func (l *PodList) GetSelected() *models.PodInfo {
	if l.selectedIdx >= 0 && l.selectedIdx < len(l.pods) {
		return &l.pods[l.selectedIdx]
	}
	return nil
}

// adjustViewport ensures the selected item is visible
func (l *PodList) adjustViewport() {
	visibleHeight := l.height - 3 // Account for header and borders

	// Scroll down if needed
	if l.selectedIdx >= l.viewportTop+visibleHeight {
		l.viewportTop = l.selectedIdx - visibleHeight + 1
	}

	// Scroll up if needed
	if l.selectedIdx < l.viewportTop {
		l.viewportTop = l.selectedIdx
	}

	// Ensure viewport is in bounds
	if l.viewportTop < 0 {
		l.viewportTop = 0
	}
}

// getFilteredPods returns pods matching the search filter
func (l *PodList) getFilteredPods() []models.PodInfo {
	if l.searchFilter == "" {
		return l.pods
	}

	filtered := []models.PodInfo{}
	searchLower := strings.ToLower(l.searchFilter)

	for _, pod := range l.pods {
		if strings.Contains(strings.ToLower(pod.Name), searchLower) ||
			strings.Contains(strings.ToLower(pod.Namespace), searchLower) ||
			strings.Contains(strings.ToLower(pod.Status), searchLower) {
			filtered = append(filtered, pod)
		}
	}

	return filtered
}

// View renders the pod list
func (l *PodList) View() string {
	if len(l.pods) == 0 {
		emptyMsg := styles.InfoBoxStyle.
			Width(l.width - 4).
			Render("No pods found")
		return emptyMsg
	}

	// Get filtered pods
	filteredPods := l.getFilteredPods()

	// Build header
	header := l.renderHeader()

	// Build rows
	var rows []string
	visibleHeight := l.height - 3

	endIdx := l.viewportTop + visibleHeight
	if endIdx > len(filteredPods) {
		endIdx = len(filteredPods)
	}

	for i := l.viewportTop; i < endIdx; i++ {
		pod := filteredPods[i]
		isSelected := i == l.selectedIdx
		row := l.renderRow(pod, isSelected)
		rows = append(rows, row)
	}

	// Join everything together
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		strings.Join(rows, "\n"),
	)

	// Add border
	return styles.BorderStyle.
		Width(l.width).
		Height(l.height).
		Render(content)
}

// renderHeader renders the table header
func (l *PodList) renderHeader() string {
	nameWidth := 30
	readyWidth := 8
	statusWidth := 15
	restartsWidth := 10
	ageWidth := 8

	header := fmt.Sprintf(
		"%-3s %-*s %-*s %-*s %-*s %-*s",
		"",
		nameWidth, "NAME",
		readyWidth, "READY",
		statusWidth, "STATUS",
		restartsWidth, "RESTARTS",
		ageWidth, "AGE",
	)

	return styles.TableHeaderStyle.
		Width(l.width - 4).
		Render(header)
}

// renderRow renders a single pod row
func (l *PodList) renderRow(pod models.PodInfo, selected bool) string {
	nameWidth := 30
	readyWidth := 8
	statusWidth := 15
	restartsWidth := 10
	ageWidth := 8

	// Truncate name if too long
	name := pod.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// Status symbol
	symbol := pod.GetStatusSymbol()

	// Status with color
	statusText := pod.Status
	statusStyle := styles.StatusStyle(pod.Status)

	row := fmt.Sprintf(
		"%s %-*s %-*s %-*s %-*d %-*s",
		symbol,
		nameWidth, name,
		readyWidth, pod.Ready,
		statusWidth, statusStyle.Render(statusText),
		restartsWidth, pod.Restarts,
		ageWidth, pod.Age,
	)

	// Apply selection style
	if selected {
		return styles.SelectedListItemStyle.
			Width(l.width - 4).
			Render(row)
	}

	return styles.ListItemStyle.
		Width(l.width - 4).
		Render(row)
}
