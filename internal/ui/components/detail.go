package components

import (
	"fmt"
	"strings"

	"github.com/williajm/k8s-tui/internal/models"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// DetailView represents a detail view for a resource
type DetailView struct {
	width  int
	height int
}

// NewDetailView creates a new detail view component
func NewDetailView() *DetailView {
	return &DetailView{
		width:  80,
		height: 20,
	}
}

// SetSize sets the dimensions
func (d *DetailView) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// ViewPod renders pod details
func (d *DetailView) ViewPod(pod *models.PodInfo) string {
	if pod == nil {
		return d.emptyView("No pod selected")
	}

	var lines []string

	// Header
	lines = append(lines, styles.DetailHeaderStyle.Render("Pod Details"))
	lines = append(lines, "")

	// Basic info
	lines = append(lines, styles.RenderDetailRow("Name", pod.Name))
	lines = append(lines, styles.RenderDetailRow("Namespace", pod.Namespace))
	lines = append(lines, styles.RenderDetailRow("Status", pod.Status))
	lines = append(lines, styles.RenderDetailRow("Ready", pod.Ready))
	lines = append(lines, styles.RenderDetailRow("Restarts", fmt.Sprintf("%d", pod.Restarts)))
	lines = append(lines, styles.RenderDetailRow("Age", pod.Age))
	lines = append(lines, styles.RenderDetailRow("IP", pod.IP))
	lines = append(lines, styles.RenderDetailRow("Node", pod.Node))
	lines = append(lines, "")

	// Containers
	lines = append(lines, styles.DetailHeaderStyle.Render("Containers"))
	lines = append(lines, "")
	for _, container := range pod.Containers {
		readySymbol := "✗"
		if container.Ready {
			readySymbol = "✓"
		}
		lines = append(lines, fmt.Sprintf("  %s %s", readySymbol, container.Name))
		lines = append(lines, fmt.Sprintf("    Image: %s", container.Image))
		lines = append(lines, fmt.Sprintf("    State: %s", container.State))
		lines = append(lines, fmt.Sprintf("    Restarts: %d", container.RestartCount))
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return styles.BorderStyle.
		Width(d.width).
		Height(d.height).
		Render(content)
}

// ViewService renders service details
func (d *DetailView) ViewService(service *models.ServiceInfo) string {
	if service == nil {
		return d.emptyView("No service selected")
	}

	var lines []string

	// Header
	lines = append(lines, styles.DetailHeaderStyle.Render("Service Details"))
	lines = append(lines, "")

	// Basic info
	lines = append(lines, styles.RenderDetailRow("Name", service.Name))
	lines = append(lines, styles.RenderDetailRow("Namespace", service.Namespace))
	lines = append(lines, styles.RenderDetailRow("Type", service.Type))
	lines = append(lines, styles.RenderDetailRow("Cluster IP", service.ClusterIP))
	lines = append(lines, styles.RenderDetailRow("External IP", service.ExternalIP))
	lines = append(lines, styles.RenderDetailRow("Ports", service.Ports))
	lines = append(lines, styles.RenderDetailRow("Age", service.Age))
	lines = append(lines, "")

	// Selectors
	if len(service.Selector) > 0 {
		lines = append(lines, styles.DetailHeaderStyle.Render("Selector"))
		lines = append(lines, "")
		for key, value := range service.Selector {
			lines = append(lines, fmt.Sprintf("  %s: %s", key, value))
		}
	}

	content := strings.Join(lines, "\n")

	return styles.BorderStyle.
		Width(d.width).
		Height(d.height).
		Render(content)
}

// ViewDeployment renders deployment details
func (d *DetailView) ViewDeployment(deployment *models.DeploymentInfo) string {
	if deployment == nil {
		return d.emptyView("No deployment selected")
	}

	var lines []string

	// Header
	lines = append(lines, styles.DetailHeaderStyle.Render("Deployment Details"))
	lines = append(lines, "")

	// Basic info
	lines = append(lines, styles.RenderDetailRow("Name", deployment.Name))
	lines = append(lines, styles.RenderDetailRow("Namespace", deployment.Namespace))
	lines = append(lines, styles.RenderDetailRow("Replicas", fmt.Sprintf("%d", deployment.Replicas)))
	lines = append(lines, styles.RenderDetailRow("Ready", deployment.Ready))
	lines = append(lines, styles.RenderDetailRow("Up-to-date", fmt.Sprintf("%d", deployment.UpToDate)))
	lines = append(lines, styles.RenderDetailRow("Available", fmt.Sprintf("%d", deployment.Available)))
	lines = append(lines, styles.RenderDetailRow("Strategy", deployment.Strategy))
	lines = append(lines, styles.RenderDetailRow("Age", deployment.Age))

	content := strings.Join(lines, "\n")

	return styles.BorderStyle.
		Width(d.width).
		Height(d.height).
		Render(content)
}

// ViewStatefulSet renders statefulset details
func (d *DetailView) ViewStatefulSet(statefulSet *models.StatefulSetInfo) string {
	if statefulSet == nil {
		return d.emptyView("No statefulset selected")
	}

	var lines []string

	// Header
	lines = append(lines, styles.DetailHeaderStyle.Render("StatefulSet Details"))
	lines = append(lines, "")

	// Basic info
	lines = append(lines, styles.RenderDetailRow("Name", statefulSet.Name))
	lines = append(lines, styles.RenderDetailRow("Namespace", statefulSet.Namespace))
	lines = append(lines, styles.RenderDetailRow("Replicas", fmt.Sprintf("%d", statefulSet.Replicas)))
	lines = append(lines, styles.RenderDetailRow("Ready", statefulSet.Ready))
	lines = append(lines, styles.RenderDetailRow("Strategy", statefulSet.Strategy))
	lines = append(lines, styles.RenderDetailRow("Age", statefulSet.Age))

	content := strings.Join(lines, "\n")

	return styles.BorderStyle.
		Width(d.width).
		Height(d.height).
		Render(content)
}

// emptyView renders an empty state message
func (d *DetailView) emptyView(message string) string {
	return styles.InfoBoxStyle.
		Width(d.width - 4).
		Height(d.height - 4).
		Render(message)
}
