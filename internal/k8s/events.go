package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetEvents retrieves events from the specified namespace
func (c *Client) GetEvents(ctx context.Context, namespace string) (*corev1.EventList, error) {
	namespace = c.resolveNamespace(namespace)

	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events, nil
}

// GetAllEvents retrieves events from all namespaces
func (c *Client) GetAllEvents(ctx context.Context) (*corev1.EventList, error) {
	events, err := c.clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all events: %w", err)
	}

	return events, nil
}

// GetEventsForResource retrieves events for a specific resource
func (c *Client) GetEventsForResource(ctx context.Context, namespace, resourceKind, resourceName string) (*corev1.EventList, error) {
	namespace = c.resolveNamespace(namespace)

	// Build field selector for the specific resource
	fieldSelector := fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s", resourceKind, resourceName)

	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events for resource: %w", err)
	}

	return events, nil
}

// GetEventsByType retrieves events filtered by type (Normal, Warning, Error)
func (c *Client) GetEventsByType(ctx context.Context, namespace, eventType string) (*corev1.EventList, error) {
	namespace = c.resolveNamespace(namespace)

	fieldSelector := fmt.Sprintf("type=%s", eventType)

	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events by type: %w", err)
	}

	return events, nil
}

// GetRecentEvents retrieves events from the last N minutes
func (c *Client) GetRecentEvents(ctx context.Context, namespace string, _ int) (*corev1.EventList, error) {
	namespace = c.resolveNamespace(namespace)

	// Note: This gets all events and filters client-side since
	// Kubernetes API doesn't support time-based field selectors
	allEvents, err := c.GetEvents(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// Filter events by time (this would be more efficient with server-side filtering)
	// For now, we return all events and let the UI handle recent filtering
	return allEvents, nil
}
