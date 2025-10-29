package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	// DefaultWatchTimeout is the default timeout for watch operations
	// Kubernetes API server typically times out watches after 5 minutes
	DefaultWatchTimeout = 5 * time.Minute
)

// WatchPods creates a watch for pods in the specified namespace.
// If resourceVersion is empty, the watch starts from the current state.
// If resourceVersion is provided, the watch resumes from that version.
func (c *Client) WatchPods(ctx context.Context, namespace string, resourceVersion string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	opts := metav1.ListOptions{
		ResourceVersion: resourceVersion,
		TimeoutSeconds:  int64ptr(int64(DefaultWatchTimeout.Seconds())),
		Watch:           true,
	}

	watcher, err := c.clientset.CoreV1().Pods(namespace).Watch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to watch pods: %w", err)
	}

	return watcher, nil
}

// WatchServices creates a watch for services in the specified namespace.
func (c *Client) WatchServices(ctx context.Context, namespace string, resourceVersion string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	opts := metav1.ListOptions{
		ResourceVersion: resourceVersion,
		TimeoutSeconds:  int64ptr(int64(DefaultWatchTimeout.Seconds())),
		Watch:           true,
	}

	watcher, err := c.clientset.CoreV1().Services(namespace).Watch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to watch services: %w", err)
	}

	return watcher, nil
}

// WatchDeployments creates a watch for deployments in the specified namespace.
func (c *Client) WatchDeployments(ctx context.Context, namespace string, resourceVersion string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	opts := metav1.ListOptions{
		ResourceVersion: resourceVersion,
		TimeoutSeconds:  int64ptr(int64(DefaultWatchTimeout.Seconds())),
		Watch:           true,
	}

	watcher, err := c.clientset.AppsV1().Deployments(namespace).Watch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to watch deployments: %w", err)
	}

	return watcher, nil
}

// WatchStatefulSets creates a watch for statefulsets in the specified namespace.
func (c *Client) WatchStatefulSets(ctx context.Context, namespace string, resourceVersion string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	opts := metav1.ListOptions{
		ResourceVersion: resourceVersion,
		TimeoutSeconds:  int64ptr(int64(DefaultWatchTimeout.Seconds())),
		Watch:           true,
	}

	watcher, err := c.clientset.AppsV1().StatefulSets(namespace).Watch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to watch statefulsets: %w", err)
	}

	return watcher, nil
}

// WatchEvents creates a watch for events in the specified namespace.
func (c *Client) WatchEvents(ctx context.Context, namespace string, resourceVersion string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	opts := metav1.ListOptions{
		ResourceVersion: resourceVersion,
		TimeoutSeconds:  int64ptr(int64(DefaultWatchTimeout.Seconds())),
		Watch:           true,
	}

	watcher, err := c.clientset.CoreV1().Events(namespace).Watch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to watch events: %w", err)
	}

	return watcher, nil
}

// int64ptr is a helper function to convert int64 to *int64
func int64ptr(i int64) *int64 {
	return &i
}
