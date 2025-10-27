package models

import (
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodInfo represents simplified pod information for display
type PodInfo struct {
	Name       string
	Namespace  string
	Status     string
	Ready      string
	Restarts   int32
	Age        string
	IP         string
	Node       string
	Containers []ContainerInfo
	Pod        *corev1.Pod // Keep reference to full pod
}

// ContainerInfo represents container information
type ContainerInfo struct {
	Name         string
	Image        string
	Ready        bool
	RestartCount int32
	State        string
}

// NewPodInfo creates a PodInfo from a Kubernetes Pod
func NewPodInfo(pod *corev1.Pod) PodInfo {
	info := PodInfo{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Status:     string(pod.Status.Phase),
		IP:         pod.Status.PodIP,
		Node:       pod.Spec.NodeName,
		Age:        formatAge(pod.CreationTimestamp),
		Pod:        pod,
		Containers: make([]ContainerInfo, 0, len(pod.Spec.Containers)),
	}

	// Calculate ready status
	readyCount := 0
	totalCount := len(pod.Spec.Containers)
	for i, container := range pod.Spec.Containers {
		containerInfo := ContainerInfo{
			Name:  container.Name,
			Image: container.Image,
		}

		// Get container status if available
		if i < len(pod.Status.ContainerStatuses) {
			status := pod.Status.ContainerStatuses[i]
			containerInfo.Ready = status.Ready
			containerInfo.RestartCount = status.RestartCount
			info.Restarts += status.RestartCount

			if status.Ready {
				readyCount++
			}

			// Determine container state
			if status.State.Running != nil {
				containerInfo.State = "Running"
			} else if status.State.Waiting != nil {
				containerInfo.State = fmt.Sprintf("Waiting: %s", status.State.Waiting.Reason)
			} else if status.State.Terminated != nil {
				containerInfo.State = fmt.Sprintf("Terminated: %s", status.State.Terminated.Reason)
			}
		}

		info.Containers = append(info.Containers, containerInfo)
	}

	info.Ready = fmt.Sprintf("%d/%d", readyCount, totalCount)

	// Refine status based on container states
	if pod.Status.Phase == corev1.PodRunning {
		if readyCount < totalCount {
			info.Status = "NotReady"
		}
	} else if pod.Status.Phase == corev1.PodPending {
		// Check for more specific pending reasons
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
				info.Status = "Pending: " + condition.Reason
				break
			}
		}
		// Check container statuses for image pull issues
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				info.Status = cs.State.Waiting.Reason
				break
			}
		}
	}

	return info
}

// formatAge formats a timestamp as a human-readable age
func formatAge(timestamp metav1.Time) string {
	duration := time.Since(timestamp.Time)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	return fmt.Sprintf("%dd", int(duration.Hours()/24))
}

// GetStatusSymbol returns a visual indicator for pod status
func (p *PodInfo) GetStatusSymbol() string {
	switch {
	case p.Status == "Running" && p.Ready == fmt.Sprintf("%d/%d", len(p.Containers), len(p.Containers)):
		return "✓"
	case p.Status == "Succeeded":
		return "✓"
	case strings.Contains(p.Status, "Error") || strings.Contains(p.Status, "Failed"):
		return "✗"
	case strings.Contains(p.Status, "Pending") || strings.Contains(p.Status, "Creating"):
		return "○"
	case strings.Contains(p.Status, "Terminating"):
		return "⊗"
	default:
		return "⚠"
	}
}

// IsHealthy returns true if the pod is running and ready
func (p *PodInfo) IsHealthy() bool {
	return p.Status == "Running" && p.Ready == fmt.Sprintf("%d/%d", len(p.Containers), len(p.Containers))
}

// GetResourceVersion returns the resource version
func (p *PodInfo) GetResourceVersion() string {
	if p.Pod != nil {
		return p.Pod.ResourceVersion
	}
	return ""
}

// ServiceInfo represents simplified service information for display
type ServiceInfo struct {
	Name       string
	Namespace  string
	Type       string
	ClusterIP  string
	ExternalIP string
	Ports      string
	Age        string
	Selector   map[string]string
	Service    *corev1.Service // Keep reference to full service
}

// NewServiceInfo creates a ServiceInfo from a Kubernetes Service
func NewServiceInfo(service *corev1.Service) ServiceInfo {
	info := ServiceInfo{
		Name:      service.Name,
		Namespace: service.Namespace,
		Type:      string(service.Spec.Type),
		ClusterIP: service.Spec.ClusterIP,
		Age:       formatAge(service.CreationTimestamp),
		Selector:  service.Spec.Selector,
		Service:   service,
	}

	// Format ports
	var ports []string
	for _, port := range service.Spec.Ports {
		if port.NodePort != 0 {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
		}
	}
	info.Ports = strings.Join(ports, ",")

	// Get external IPs
	externalIPs := make([]string, 0, len(service.Spec.ExternalIPs)+len(service.Status.LoadBalancer.Ingress))
	externalIPs = append(externalIPs, service.Spec.ExternalIPs...)
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				externalIPs = append(externalIPs, ingress.IP)
			} else if ingress.Hostname != "" {
				externalIPs = append(externalIPs, ingress.Hostname)
			}
		}
	}
	info.ExternalIP = strings.Join(externalIPs, ",")
	if info.ExternalIP == "" {
		info.ExternalIP = "<none>"
	}

	return info
}

// GetStatusSymbol returns a visual indicator for service status
func (s *ServiceInfo) GetStatusSymbol() string {
	// Services don't have a running status like pods, so we indicate if they have endpoints
	if s.Type == "LoadBalancer" && s.ExternalIP == "<none>" {
		return "○" // Pending external IP
	}
	return "✓" // Service exists and is configured
}

// DeploymentInfo represents simplified deployment information for display
type DeploymentInfo struct {
	Name       string
	Namespace  string
	Ready      string
	UpToDate   int32
	Available  int32
	Age        string
	Replicas   int32
	Strategy   string
	Deployment *appsv1.Deployment // Keep reference to full deployment
}

// NewDeploymentInfo creates a DeploymentInfo from a Kubernetes Deployment
func NewDeploymentInfo(deployment *appsv1.Deployment) DeploymentInfo {
	replicas := int32(0)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}

	info := DeploymentInfo{
		Name:       deployment.Name,
		Namespace:  deployment.Namespace,
		UpToDate:   deployment.Status.UpdatedReplicas,
		Available:  deployment.Status.AvailableReplicas,
		Age:        formatAge(deployment.CreationTimestamp),
		Replicas:   replicas,
		Strategy:   string(deployment.Spec.Strategy.Type),
		Deployment: deployment,
	}

	// Calculate ready status
	info.Ready = fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, replicas)

	return info
}

// GetStatusSymbol returns a visual indicator for deployment status
func (d *DeploymentInfo) GetStatusSymbol() string {
	if d.Available == d.Replicas && d.UpToDate == d.Replicas {
		return "✓"
	}
	if d.Available == 0 {
		return "✗"
	}
	return "○" // Partially ready
}

// IsHealthy returns true if the deployment is fully ready
func (d *DeploymentInfo) IsHealthy() bool {
	return d.Available == d.Replicas && d.UpToDate == d.Replicas
}

// StatefulSetInfo represents simplified statefulset information for display
type StatefulSetInfo struct {
	Name        string
	Namespace   string
	Ready       string
	Age         string
	Replicas    int32
	Strategy    string
	StatefulSet *appsv1.StatefulSet // Keep reference to full statefulset
}

// NewStatefulSetInfo creates a StatefulSetInfo from a Kubernetes StatefulSet
func NewStatefulSetInfo(statefulSet *appsv1.StatefulSet) StatefulSetInfo {
	replicas := int32(0)
	if statefulSet.Spec.Replicas != nil {
		replicas = *statefulSet.Spec.Replicas
	}

	info := StatefulSetInfo{
		Name:        statefulSet.Name,
		Namespace:   statefulSet.Namespace,
		Age:         formatAge(statefulSet.CreationTimestamp),
		Replicas:    replicas,
		Strategy:    string(statefulSet.Spec.UpdateStrategy.Type),
		StatefulSet: statefulSet,
	}

	// Calculate ready status
	info.Ready = fmt.Sprintf("%d/%d", statefulSet.Status.ReadyReplicas, replicas)

	return info
}

// GetStatusSymbol returns a visual indicator for statefulset status
func (s *StatefulSetInfo) GetStatusSymbol() string {
	if s.StatefulSet.Status.ReadyReplicas == s.Replicas {
		return "✓"
	}
	if s.StatefulSet.Status.ReadyReplicas == 0 {
		return "✗"
	}
	return "○" // Partially ready
}

// IsHealthy returns true if the statefulset is fully ready
func (s *StatefulSetInfo) IsHealthy() bool {
	return s.StatefulSet.Status.ReadyReplicas == s.Replicas
}
