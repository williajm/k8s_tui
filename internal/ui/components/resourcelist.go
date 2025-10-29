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
	ResourceTypeEvent
)

// ResourceList represents a generic list of resources
type ResourceList struct {
	resourceType ResourceType
	pods         []models.PodInfo
	services     []models.ServiceInfo
	deployments  []models.DeploymentInfo
	statefulSets []models.StatefulSetInfo
	events       []models.EventInfo
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
		events:       []models.EventInfo{},
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

// SetEvents updates the list of events
func (l *ResourceList) SetEvents(events []models.EventInfo) {
	l.events = events
	if l.selectedIdx >= len(l.events) {
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

// GetSelectedEvent returns the currently selected event
func (l *ResourceList) GetSelectedEvent() *models.EventInfo {
	if l.resourceType == ResourceTypeEvent && l.selectedIdx >= 0 && l.selectedIdx < len(l.events) {
		return &l.events[l.selectedIdx]
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
	case ResourceTypeEvent:
		return len(l.events)
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
		statusWidth := 20
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

	case ResourceTypeEvent:
		typeWidth := 8
		ageWidth := 8
		reasonWidth := 20
		objectWidth := 25
		messageWidth := 30

		header = fmt.Sprintf(
			"%-3s %-*s %-*s %-*s %-*s %-*s",
			"",
			typeWidth, "TYPE",
			ageWidth, "AGE",
			reasonWidth, "REASON",
			objectWidth, "OBJECT",
			messageWidth, "MESSAGE",
		)
	}

	return styles.TableHeaderStyle.
		Width(l.width - 4).
		Render(header)
}

// renderRow renders a single row based on resource type
func (l *ResourceList) renderRow(idx int, selected bool) string {
	var row string

	switch l.resourceType {
	case ResourceTypePod:
		row = l.renderPodRow(idx)
	case ResourceTypeService:
		row = l.renderServiceRow(idx)
	case ResourceTypeDeployment:
		row = l.renderDeploymentRow(idx)
	case ResourceTypeStatefulSet:
		row = l.renderStatefulSetRow(idx)
	case ResourceTypeEvent:
		row = l.renderEventRow(idx)
	}

	if row == "" {
		return ""
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

func (l *ResourceList) renderPodRow(idx int) string {
	if idx >= len(l.pods) {
		return ""
	}
	pod := l.pods[idx]
	symbol := pod.GetStatusSymbol()

	nameWidth := 30
	readyWidth := 8
	statusWidth := 20
	restartsWidth := 10
	ageWidth := 8

	name := pod.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	statusText := pod.Status
	statusStyle := styles.StatusStyle(pod.Status)
	statusRendered := statusStyle.Width(statusWidth).Render(statusText)

	return fmt.Sprintf(
		"%s %-*s %-*s %s %*d %-*s",
		symbol,
		nameWidth, name,
		readyWidth, pod.Ready,
		statusRendered,
		restartsWidth, pod.Restarts,
		ageWidth, pod.Age,
	)
}

func (l *ResourceList) renderServiceRow(idx int) string {
	if idx >= len(l.services) {
		return ""
	}
	svc := l.services[idx]
	symbol := svc.GetStatusSymbol()

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

	return fmt.Sprintf(
		"%s %-*s %-*s %-*s %-*s %-*s %-*s",
		symbol,
		nameWidth, name,
		typeWidth, svc.Type,
		clusterIPWidth, svc.ClusterIP,
		externalIPWidth, svc.ExternalIP,
		portsWidth, svc.Ports,
		ageWidth, svc.Age,
	)
}

func (l *ResourceList) renderDeploymentRow(idx int) string {
	if idx >= len(l.deployments) {
		return ""
	}
	dep := l.deployments[idx]
	symbol := dep.GetStatusSymbol()

	nameWidth := 30
	readyWidth := 10
	upToDateWidth := 12
	availableWidth := 12
	ageWidth := 8

	name := dep.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	return fmt.Sprintf(
		"%s %-*s %-*s %-*d %-*d %-*s",
		symbol,
		nameWidth, name,
		readyWidth, dep.Ready,
		upToDateWidth, dep.UpToDate,
		availableWidth, dep.Available,
		ageWidth, dep.Age,
	)
}

func (l *ResourceList) renderStatefulSetRow(idx int) string {
	if idx >= len(l.statefulSets) {
		return ""
	}
	sts := l.statefulSets[idx]
	symbol := sts.GetStatusSymbol()

	nameWidth := 30
	readyWidth := 10
	ageWidth := 8

	name := sts.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	return fmt.Sprintf(
		"%s %-*s %-*s %-*s",
		symbol,
		nameWidth, name,
		readyWidth, sts.Ready,
		ageWidth, sts.Age,
	)
}

func (l *ResourceList) renderEventRow(idx int) string {
	if idx >= len(l.events) {
		return ""
	}
	event := l.events[idx]
	symbol := event.GetTypeSymbol()

	typeWidth := 8
	ageWidth := 8
	reasonWidth := 20
	objectWidth := 25
	messageWidth := 30

	eventType := event.Type
	if len(eventType) > typeWidth {
		eventType = eventType[:typeWidth-3] + "..."
	}

	reason := event.Reason
	if len(reason) > reasonWidth {
		reason = reason[:reasonWidth-3] + "..."
	}

	object := event.Object
	if len(object) > objectWidth {
		object = object[:objectWidth-3] + "..."
	}

	message := event.GetMessagePreview(messageWidth)

	return fmt.Sprintf(
		"%s %-*s %-*s %-*s %-*s %-*s",
		symbol,
		typeWidth, eventType,
		ageWidth, event.FormatAge(),
		reasonWidth, reason,
		objectWidth, object,
		messageWidth, message,
	)
}

// AddOrUpdatePod adds a new pod or updates an existing one
func (l *ResourceList) AddOrUpdatePod(pod models.PodInfo) {
	// Find if pod already exists
	for i, existing := range l.pods {
		if existing.Namespace == pod.Namespace && existing.Name == pod.Name {
			// Update existing pod
			l.pods[i] = pod
			return
		}
	}
	// Add new pod
	l.pods = append(l.pods, pod)
}

// RemovePod removes a pod by namespace and name
func (l *ResourceList) RemovePod(namespace, name string) {
	for i, pod := range l.pods {
		if pod.Namespace == namespace && pod.Name == name {
			// Remove pod from slice
			l.pods = append(l.pods[:i], l.pods[i+1:]...)
			// Adjust selection if needed
			if l.selectedIdx >= len(l.pods) && len(l.pods) > 0 {
				l.selectedIdx = len(l.pods) - 1
			}
			if len(l.pods) == 0 {
				l.selectedIdx = 0
			}
			return
		}
	}
}

// AddOrUpdateService adds a new service or updates an existing one
func (l *ResourceList) AddOrUpdateService(service models.ServiceInfo) {
	for i, existing := range l.services {
		if existing.Namespace == service.Namespace && existing.Name == service.Name {
			l.services[i] = service
			return
		}
	}
	l.services = append(l.services, service)
}

// RemoveService removes a service by namespace and name
func (l *ResourceList) RemoveService(namespace, name string) {
	for i, svc := range l.services {
		if svc.Namespace == namespace && svc.Name == name {
			l.services = append(l.services[:i], l.services[i+1:]...)
			if l.selectedIdx >= len(l.services) && len(l.services) > 0 {
				l.selectedIdx = len(l.services) - 1
			}
			if len(l.services) == 0 {
				l.selectedIdx = 0
			}
			return
		}
	}
}

// AddOrUpdateDeployment adds a new deployment or updates an existing one
func (l *ResourceList) AddOrUpdateDeployment(deployment models.DeploymentInfo) {
	for i, existing := range l.deployments {
		if existing.Namespace == deployment.Namespace && existing.Name == deployment.Name {
			l.deployments[i] = deployment
			return
		}
	}
	l.deployments = append(l.deployments, deployment)
}

// RemoveDeployment removes a deployment by namespace and name
func (l *ResourceList) RemoveDeployment(namespace, name string) {
	for i, dep := range l.deployments {
		if dep.Namespace == namespace && dep.Name == name {
			l.deployments = append(l.deployments[:i], l.deployments[i+1:]...)
			if l.selectedIdx >= len(l.deployments) && len(l.deployments) > 0 {
				l.selectedIdx = len(l.deployments) - 1
			}
			if len(l.deployments) == 0 {
				l.selectedIdx = 0
			}
			return
		}
	}
}

// AddOrUpdateStatefulSet adds a new statefulset or updates an existing one
func (l *ResourceList) AddOrUpdateStatefulSet(statefulSet models.StatefulSetInfo) {
	for i, existing := range l.statefulSets {
		if existing.Namespace == statefulSet.Namespace && existing.Name == statefulSet.Name {
			l.statefulSets[i] = statefulSet
			return
		}
	}
	l.statefulSets = append(l.statefulSets, statefulSet)
}

// RemoveStatefulSet removes a statefulset by namespace and name
func (l *ResourceList) RemoveStatefulSet(namespace, name string) {
	for i, sts := range l.statefulSets {
		if sts.Namespace == namespace && sts.Name == name {
			l.statefulSets = append(l.statefulSets[:i], l.statefulSets[i+1:]...)
			if l.selectedIdx >= len(l.statefulSets) && len(l.statefulSets) > 0 {
				l.selectedIdx = len(l.statefulSets) - 1
			}
			if len(l.statefulSets) == 0 {
				l.selectedIdx = 0
			}
			return
		}
	}
}

// AddOrUpdateEvent adds a new event or updates an existing one
func (l *ResourceList) AddOrUpdateEvent(event models.EventInfo) {
	for i, existing := range l.events {
		if existing.Namespace == event.Namespace && existing.Name == event.Name {
			l.events[i] = event
			return
		}
	}
	l.events = append(l.events, event)
}

// RemoveEvent removes an event by namespace and name
func (l *ResourceList) RemoveEvent(namespace, name string) {
	for i, evt := range l.events {
		if evt.Namespace == namespace && evt.Name == name {
			l.events = append(l.events[:i], l.events[i+1:]...)
			if l.selectedIdx >= len(l.events) && len(l.events) > 0 {
				l.selectedIdx = len(l.events) - 1
			}
			if len(l.events) == 0 {
				l.selectedIdx = 0
			}
			return
		}
	}
}
