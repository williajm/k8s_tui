package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/williajm/k8s-tui/internal/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// GetResourceYAML retrieves a resource and returns it as YAML
func (c *Client) GetResourceYAML(ctx context.Context, resourceType, namespace, name string) (string, error) {
	namespace = c.resolveNamespace(namespace)

	var obj interface{}
	var err error

	switch resourceType {
	case "Pod":
		obj, err = c.GetPod(ctx, namespace, name)
	case "Service":
		obj, err = c.GetService(ctx, namespace, name)
	case "Deployment":
		obj, err = c.GetDeployment(ctx, namespace, name)
	case "StatefulSet":
		obj, err = c.GetStatefulSet(ctx, namespace, name)
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return "", err
	}

	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// GetResourceJSON retrieves a resource and returns it as JSON
func (c *Client) GetResourceJSON(ctx context.Context, resourceType, namespace, name string) (string, error) {
	namespace = c.resolveNamespace(namespace)

	var obj interface{}
	var err error

	switch resourceType {
	case "Pod":
		obj, err = c.GetPod(ctx, namespace, name)
	case "Service":
		obj, err = c.GetService(ctx, namespace, name)
	case "Deployment":
		obj, err = c.GetDeployment(ctx, namespace, name)
	case "StatefulSet":
		obj, err = c.GetStatefulSet(ctx, namespace, name)
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// DescribePod generates a kubectl-style describe output for a pod
func (c *Client) DescribePod(ctx context.Context, namespace, name string) (*models.DescribeData, error) {
	namespace = c.resolveNamespace(namespace)

	pod, err := c.GetPod(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	desc := models.NewDescribeData("Pod", name, namespace)

	// Metadata section
	metadata := desc.AddSection("Metadata")
	metadata.AddField("Name", pod.Name, 0)
	metadata.AddField("Namespace", pod.Namespace, 0)
	metadata.AddField("Labels", formatMap(pod.Labels), 0)
	metadata.AddField("Annotations", formatMap(pod.Annotations), 0)
	metadata.AddField("Status", string(pod.Status.Phase), 0)
	metadata.AddField("IP", pod.Status.PodIP, 0)
	metadata.AddField("Node", pod.Spec.NodeName, 0)

	// Containers section
	containers := desc.AddSection("Containers")
	for _, container := range pod.Spec.Containers {
		containers.AddField(container.Name, "", 0)
		containers.AddField("Image", container.Image, 1)
		containers.AddField("Ports", formatPorts(container.Ports), 1)

		// Find container status
		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == container.Name {
				containers.AddField("Ready", fmt.Sprintf("%v", status.Ready), 1)
				containers.AddField("Restart Count", fmt.Sprintf("%d", status.RestartCount), 1)
				containers.AddField("State", formatContainerState(status.State), 1)
			}
		}
	}

	// Conditions section
	if len(pod.Status.Conditions) > 0 {
		conditions := desc.AddSection("Conditions")
		for _, condition := range pod.Status.Conditions {
			conditions.AddField(string(condition.Type), string(condition.Status), 0)
			if condition.Reason != "" {
				conditions.AddField("Reason", condition.Reason, 1)
			}
			if condition.Message != "" {
				conditions.AddField("Message", condition.Message, 1)
			}
		}
	}

	return desc, nil
}

// DescribeService generates a kubectl-style describe output for a service
func (c *Client) DescribeService(ctx context.Context, namespace, name string) (*models.DescribeData, error) {
	namespace = c.resolveNamespace(namespace)

	service, err := c.GetService(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	desc := models.NewDescribeData("Service", name, namespace)

	// Metadata section
	metadata := desc.AddSection("Metadata")
	metadata.AddField("Name", service.Name, 0)
	metadata.AddField("Namespace", service.Namespace, 0)
	metadata.AddField("Labels", formatMap(service.Labels), 0)
	metadata.AddField("Annotations", formatMap(service.Annotations), 0)

	// Spec section
	spec := desc.AddSection("Spec")
	spec.AddField("Type", string(service.Spec.Type), 0)
	spec.AddField("Cluster IP", service.Spec.ClusterIP, 0)
	spec.AddField("External IPs", formatStringSlice(service.Spec.ExternalIPs), 0)
	spec.AddField("Selector", formatMap(service.Spec.Selector), 0)

	// Ports
	if len(service.Spec.Ports) > 0 {
		spec.AddField("Ports", "", 0)
		for _, port := range service.Spec.Ports {
			portStr := fmt.Sprintf("%s:%d/%s", port.Name, port.Port, port.Protocol)
			if port.NodePort != 0 {
				portStr += fmt.Sprintf(" (NodePort: %d)", port.NodePort)
			}
			spec.AddField("", portStr, 1)
		}
	}

	// Load Balancer section
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		lb := desc.AddSection("LoadBalancer")
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			for _, ingress := range service.Status.LoadBalancer.Ingress {
				if ingress.IP != "" {
					lb.AddField("Ingress", ingress.IP, 0)
				} else if ingress.Hostname != "" {
					lb.AddField("Ingress", ingress.Hostname, 0)
				}
			}
		} else {
			lb.AddField("Ingress", "Pending", 0)
		}
	}

	return desc, nil
}

// DescribeDeployment generates a kubectl-style describe output for a deployment
func (c *Client) DescribeDeployment(ctx context.Context, namespace, name string) (*models.DescribeData, error) {
	namespace = c.resolveNamespace(namespace)

	deployment, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	desc := models.NewDescribeData("Deployment", name, namespace)

	// Metadata section
	metadata := desc.AddSection("Metadata")
	metadata.AddField("Name", deployment.Name, 0)
	metadata.AddField("Namespace", deployment.Namespace, 0)
	metadata.AddField("Labels", formatMap(deployment.Labels), 0)
	metadata.AddField("Annotations", formatMap(deployment.Annotations), 0)

	// Strategy section
	strategy := desc.AddSection("Strategy")
	strategy.AddField("Type", string(deployment.Spec.Strategy.Type), 0)
	if deployment.Spec.Strategy.Type == appsv1.RollingUpdateDeploymentStrategyType &&
		deployment.Spec.Strategy.RollingUpdate != nil {
		ru := deployment.Spec.Strategy.RollingUpdate
		if ru.MaxUnavailable != nil {
			strategy.AddField("Max Unavailable", ru.MaxUnavailable.String(), 1)
		}
		if ru.MaxSurge != nil {
			strategy.AddField("Max Surge", ru.MaxSurge.String(), 1)
		}
	}

	// Replicas section
	replicas := desc.AddSection("Replicas")
	if deployment.Spec.Replicas != nil {
		replicas.AddField("Desired", fmt.Sprintf("%d", *deployment.Spec.Replicas), 0)
	}
	replicas.AddField("Current", fmt.Sprintf("%d", deployment.Status.Replicas), 0)
	replicas.AddField("Updated", fmt.Sprintf("%d", deployment.Status.UpdatedReplicas), 0)
	replicas.AddField("Ready", fmt.Sprintf("%d", deployment.Status.ReadyReplicas), 0)
	replicas.AddField("Available", fmt.Sprintf("%d", deployment.Status.AvailableReplicas), 0)

	// Conditions section
	if len(deployment.Status.Conditions) > 0 {
		conditions := desc.AddSection("Conditions")
		for _, condition := range deployment.Status.Conditions {
			conditions.AddField(string(condition.Type), string(condition.Status), 0)
			if condition.Reason != "" {
				conditions.AddField("Reason", condition.Reason, 1)
			}
			if condition.Message != "" {
				conditions.AddField("Message", condition.Message, 1)
			}
		}
	}

	return desc, nil
}

// DescribeStatefulSet generates a kubectl-style describe output for a statefulset
func (c *Client) DescribeStatefulSet(ctx context.Context, namespace, name string) (*models.DescribeData, error) {
	namespace = c.resolveNamespace(namespace)

	sts, err := c.GetStatefulSet(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	desc := models.NewDescribeData("StatefulSet", name, namespace)

	// Metadata section
	metadata := desc.AddSection("Metadata")
	metadata.AddField("Name", sts.Name, 0)
	metadata.AddField("Namespace", sts.Namespace, 0)
	metadata.AddField("Labels", formatMap(sts.Labels), 0)
	metadata.AddField("Annotations", formatMap(sts.Annotations), 0)

	// Strategy section
	strategy := desc.AddSection("Update Strategy")
	strategy.AddField("Type", string(sts.Spec.UpdateStrategy.Type), 0)
	if sts.Spec.UpdateStrategy.RollingUpdate != nil &&
		sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		strategy.AddField("Partition", fmt.Sprintf("%d", *sts.Spec.UpdateStrategy.RollingUpdate.Partition), 1)
	}

	// Replicas section
	replicas := desc.AddSection("Replicas")
	if sts.Spec.Replicas != nil {
		replicas.AddField("Desired", fmt.Sprintf("%d", *sts.Spec.Replicas), 0)
	}
	replicas.AddField("Current", fmt.Sprintf("%d", sts.Status.Replicas), 0)
	replicas.AddField("Ready", fmt.Sprintf("%d", sts.Status.ReadyReplicas), 0)

	// Service Name
	spec := desc.AddSection("Spec")
	spec.AddField("Service Name", sts.Spec.ServiceName, 0)
	spec.AddField("Pod Management Policy", string(sts.Spec.PodManagementPolicy), 0)

	return desc, nil
}

// Helper functions for formatting

func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return "<none>"
	}
	result := ""
	for k, v := range m {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("%s=%s", k, v)
	}
	return result
}

func formatStringSlice(s []string) string {
	if len(s) == 0 {
		return "<none>"
	}
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}

func formatPorts(ports []corev1.ContainerPort) string {
	if len(ports) == 0 {
		return "<none>"
	}
	result := ""
	for i, port := range ports {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol)
	}
	return result
}

func formatContainerState(state corev1.ContainerState) string {
	if state.Running != nil {
		return "Running"
	}
	if state.Waiting != nil {
		return fmt.Sprintf("Waiting: %s", state.Waiting.Reason)
	}
	if state.Terminated != nil {
		return fmt.Sprintf("Terminated: %s (exit code %d)", state.Terminated.Reason, state.Terminated.ExitCode)
	}
	return "Unknown"
}
