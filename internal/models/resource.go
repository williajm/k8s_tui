package models

import (
	"fmt"
	"strings"
	"time"

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
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
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
