package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/models"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// ResourceType represents the type of resource being displayed
type ResourceType int

const (
	ResourceTypePod ResourceType = iota
	ResourceTypeService
	ResourceTypeDeployment
	ResourceTypeStatefulSet
)

// ResourceList represents a generic list of resources
type ResourceList struct {
	resourceType ResourceType
	pods         []models.PodInfo
	services     []models.ServiceInfo
	deployments  []models.DeploymentInfo
	statefulSets []models.StatefulSetInfo
	selectedIdx  int
	viewportTop  int
	width        int
	height       int
	searchFilter string
}

// NewResourceList creates a new resource list component
func NewResourceList(resourceType ResourceType) *ResourceList {
	return &ResourceList{
		resourceType: resourceType,
		pods:         []models.PodInfo{},
		services:     []models.ServiceInfo{},
		deployments:  []models.DeploymentInfo{},
		statefulSets: []models.StatefulSetInfo{},
		selectedIdx:  0,
		viewportTop:  0,
		width:        80,
		height:       20,
	}
}

// SetResourceType changes the resource type
func (l *ResourceList) SetResourceType(resourceType ResourceType) {
	l.resourceType = resourceType
	l.selectedIdx = 0
	l.viewportTop = 0
}

// SetPods updates the list of pods
func (l *ResourceList) SetPods(pods []models.PodInfo) {
	l.pods = pods
	if l.selectedIdx >= len(l.pods) {
		l.selectedIdx = 0
	}
}

// SetServices updates the list of services
func (l *ResourceList) SetServices(services []models.ServiceInfo) {
	l.services = services
	if l.selectedIdx >= len(l.services) {
		l.selectedIdx = 0
	}
}

// SetDeployments updates the list of deployments
func (l *ResourceList) SetDeployments(deployments []models.DeploymentInfo) {
	l.deployments = deployments
	if l.selectedIdx >= len(l.deployments) {
		l.selectedIdx = 0
	}
}

// SetStatefulSets updates the list of statefulsets
func (l *ResourceList) SetStatefulSets(statefulSets []models.StatefulSetInfo) {
	l.statefulSets = statefulSets
	if l.selectedIdx >= len(l.statefulSets) {
		l.selectedIdx = 0
	}
}

// SetSize sets the dimensions
func (l *ResourceList) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// SetSearchFilter sets the search filter
func (l *ResourceList) SetSearchFilter(filter string) {
	l.searchFilter = filter
}

// MoveUp moves the selection up
func (l *ResourceList) MoveUp() {
	if l.selectedIdx > 0 {
		l.selectedIdx--
		l.adjustViewport()
	}
}

// MoveDown moves the selection down
func (l *ResourceList) MoveDown() {
	maxIdx := l.getItemCount() - 1
	if l.selectedIdx < maxIdx {
		l.selectedIdx++
		l.adjustViewport()
	}
}

// PageUp moves up by one page
func (l *ResourceList) PageUp() {
	l.selectedIdx -= l.height - 3 // Account for header
	if l.selectedIdx < 0 {
		l.selectedIdx = 0
	}
	l.adjustViewport()
}

// PageDown moves down by one page
func (l *ResourceList) PageDown() {
	l.selectedIdx += l.height - 3
	maxIdx := l.getItemCount() - 1
	if l.selectedIdx > maxIdx {
		l.selectedIdx = maxIdx
	}
	l.adjustViewport()
}

// Home moves to the top
func (l *ResourceList) Home() {
	l.selectedIdx = 0
	l.viewportTop = 0
}

// End moves to the bottom
func (l *ResourceList) End() {
	l.selectedIdx = l.getItemCount() - 1
	l.adjustViewport()
}

// GetSelectedPod returns the currently selected pod
func (l *ResourceList) GetSelectedPod() *models.PodInfo {
	if l.resourceType == ResourceTypePod && l.selectedIdx >= 0 && l.selectedIdx < len(l.pods) {
		return &l.pods[l.selectedIdx]
	}
	return nil
}

// GetSelectedService returns the currently selected service
func (l *ResourceList) GetSelectedService() *models.ServiceInfo {
	if l.resourceType == ResourceTypeService && l.selectedIdx >= 0 && l.selectedIdx < len(l.services) {
		return &l.services[l.selectedIdx]
	}
	return nil
}

// GetSelectedDeployment returns the currently selected deployment
func (l *ResourceList) GetSelectedDeployment() *models.DeploymentInfo {
	if l.resourceType == ResourceTypeDeployment && l.selectedIdx >= 0 && l.selectedIdx < len(l.deployments) {
		return &l.deployments[l.selectedIdx]
	}
	return nil
}

// GetSelectedStatefulSet returns the currently selected statefulset
func (l *ResourceList) GetSelectedStatefulSet() *models.StatefulSetInfo {
	if l.resourceType == ResourceTypeStatefulSet && l.selectedIdx >= 0 && l.selectedIdx < len(l.statefulSets) {
		return &l.statefulSets[l.selectedIdx]
	}
	return nil
}

// getItemCount returns the number of items in the current resource list
func (l *ResourceList) getItemCount() int {
	switch l.resourceType {
	case ResourceTypePod:
		return len(l.pods)
	case ResourceTypeService:
		return len(l.services)
	case ResourceTypeDeployment:
		return len(l.deployments)
	case ResourceTypeStatefulSet:
		return len(l.statefulSets)
	default:
		return 0
	}
}

// adjustViewport ensures the selected item is visible
func (l *ResourceList) adjustViewport() {
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

// View renders the resource list
func (l *ResourceList) View() string {
	itemCount := l.getItemCount()
	if itemCount == 0 {
		emptyMsg := styles.InfoBoxStyle.
			Width(l.width - 4).
			Render("No resources found")
		return emptyMsg
	}

	// Build header
	header := l.renderHeader()

	// Build rows
	var rows []string
	visibleHeight := l.height - 3

	endIdx := l.viewportTop + visibleHeight
	if endIdx > itemCount {
		endIdx = itemCount
	}

	for i := l.viewportTop; i < endIdx; i++ {
		isSelected := i == l.selectedIdx
		row := l.renderRow(i, isSelected)
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

// renderHeader renders the table header based on resource type
func (l *ResourceList) renderHeader() string {
	var header string

	switch l.resourceType {
	case ResourceTypePod:
		nameWidth := 30
		readyWidth := 8
		statusWidth := 15
		restartsWidth := 10
		ageWidth := 8

		header = fmt.Sprintf(
			"%-3s %-*s %-*s %-*s %-*s %-*s",
			"",
			nameWidth, "NAME",
			readyWidth, "READY",
			statusWidth, "STATUS",
			restartsWidth, "RESTARTS",
			ageWidth, "AGE",
		)

	case ResourceTypeService:
		nameWidth := 25
		typeWidth := 12
		clusterIPWidth := 16
		externalIPWidth := 16
		portsWidth := 15
		ageWidth := 8

		header = fmt.Sprintf(
			"%-3s %-*s %-*s %-*s %-*s %-*s %-*s",
			"",
			nameWidth, "NAME",
			typeWidth, "TYPE",
			clusterIPWidth, "CLUSTER-IP",
			externalIPWidth, "EXTERNAL-IP",
			portsWidth, "PORT(S)",
			ageWidth, "AGE",
		)

	case ResourceTypeDeployment:
		nameWidth := 30
		readyWidth := 10
		upToDateWidth := 12
		availableWidth := 12
		ageWidth := 8

		header = fmt.Sprintf(
			"%-3s %-*s %-*s %-*s %-*s %-*s",
			"",
			nameWidth, "NAME",
			readyWidth, "READY",
			upToDateWidth, "UP-TO-DATE",
			availableWidth, "AVAILABLE",
			ageWidth, "AGE",
		)

	case ResourceTypeStatefulSet:
		nameWidth := 30
		readyWidth := 10
		ageWidth := 8

		header = fmt.Sprintf(
			"%-3s %-*s %-*s %-*s",
			"",
			nameWidth, "NAME",
			readyWidth, "READY",
			ageWidth, "AGE",
		)
	}

	return styles.TableHeaderStyle.
		Width(l.width - 4).
		Render(header)
}

// renderRow renders a single row based on resource type
//
//nolint:funlen // Handles rendering for multiple resource types
func (l *ResourceList) renderRow(idx int, selected bool) string {
	var row string
	var symbol string

	switch l.resourceType {
	case ResourceTypePod:
		if idx >= len(l.pods) {
			return ""
		}
		pod := l.pods[idx]
		symbol = pod.GetStatusSymbol()

		nameWidth := 30
		readyWidth := 8
		statusWidth := 15
		restartsWidth := 10
		ageWidth := 8

		name := pod.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		statusText := pod.Status
		statusStyle := styles.StatusStyle(pod.Status)

		row = fmt.Sprintf(
			"%s %-*s %-*s %-*s %-*d %-*s",
			symbol,
			nameWidth, name,
			readyWidth, pod.Ready,
			statusWidth, statusStyle.Render(statusText),
			restartsWidth, pod.Restarts,
			ageWidth, pod.Age,
		)

	case ResourceTypeService:
		if idx >= len(l.services) {
			return ""
		}
		svc := l.services[idx]
		symbol = svc.GetStatusSymbol()

		nameWidth := 25
		typeWidth := 12
		clusterIPWidth := 16
		externalIPWidth := 16
		portsWidth := 15
		ageWidth := 8

		name := svc.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		row = fmt.Sprintf(
			"%s %-*s %-*s %-*s %-*s %-*s %-*s",
			symbol,
			nameWidth, name,
			typeWidth, svc.Type,
			clusterIPWidth, svc.ClusterIP,
			externalIPWidth, svc.ExternalIP,
			portsWidth, svc.Ports,
			ageWidth, svc.Age,
		)

	case ResourceTypeDeployment:
		if idx >= len(l.deployments) {
			return ""
		}
		dep := l.deployments[idx]
		symbol = dep.GetStatusSymbol()

		nameWidth := 30
		readyWidth := 10
		upToDateWidth := 12
		availableWidth := 12
		ageWidth := 8

		name := dep.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		row = fmt.Sprintf(
			"%s %-*s %-*s %-*d %-*d %-*s",
			symbol,
			nameWidth, name,
			readyWidth, dep.Ready,
			upToDateWidth, dep.UpToDate,
			availableWidth, dep.Available,
			ageWidth, dep.Age,
		)

	case ResourceTypeStatefulSet:
		if idx >= len(l.statefulSets) {
			return ""
		}
		sts := l.statefulSets[idx]
		symbol = sts.GetStatusSymbol()

		nameWidth := 30
		readyWidth := 10
		ageWidth := 8

		name := sts.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		row = fmt.Sprintf(
			"%s %-*s %-*s %-*s",
			symbol,
			nameWidth, name,
			readyWidth, sts.Ready,
			ageWidth, sts.Age,
		)
	}

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
