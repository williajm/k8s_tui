package k8s

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/williajm/k8s-tui/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPodLogsStream streams logs from a pod container
// Returns a channel that receives log entries and an error channel
func (c *Client) GetPodLogsStream(
	ctx context.Context, namespace, podName, containerName string, options models.LogOptions,
) (<-chan models.LogEntry, <-chan error) {
	logChan := make(chan models.LogEntry, 1000)
	errChan := make(chan error, 1)

	if namespace == "" {
		namespace = c.namespace
	}

	go func() {
		defer close(logChan)
		defer close(errChan)

		// Build pod log options
		podLogOpts := &corev1.PodLogOptions{
			Container:  containerName,
			Follow:     options.Follow,
			Timestamps: options.Timestamps,
			Previous:   options.Previous,
		}

		if options.TailLines > 0 {
			podLogOpts.TailLines = &options.TailLines
		}

		if options.SinceTime != nil {
			podLogOpts.SinceTime = &metav1.Time{Time: *options.SinceTime}
		}

		// Get log stream
		req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
		stream, err := req.Stream(ctx)
		if err != nil {
			errChan <- fmt.Errorf("failed to open log stream: %w", err)
			return
		}
		defer stream.Close()

		// Read logs line by line
		reader := bufio.NewReader(stream)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						errChan <- fmt.Errorf("error reading log stream: %w", err)
					}
					return
				}

				// Parse log entry
				entry := parseLogLine(line, containerName, options.Timestamps)
				logChan <- entry
			}
		}
	}()

	return logChan, errChan
}

// GetPodLogsStatic retrieves static logs (non-streaming) from a pod container
func (c *Client) GetPodLogsStatic(
	ctx context.Context, namespace, podName, containerName string, options models.LogOptions,
) ([]models.LogEntry, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	podLogOpts := &corev1.PodLogOptions{
		Container:  containerName,
		Follow:     false,
		Timestamps: options.Timestamps,
		Previous:   options.Previous,
	}

	if options.TailLines > 0 {
		podLogOpts.TailLines = &options.TailLines
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}

	// Parse logs into entries
	lines := strings.Split(string(logs), "\n")
	entries := make([]models.LogEntry, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		entry := parseLogLine(line, containerName, options.Timestamps)
		entries = append(entries, entry)
	}

	return entries, nil
}

// parseLogLine parses a single log line into a LogEntry
func parseLogLine(line, container string, hasTimestamp bool) models.LogEntry {
	line = strings.TrimSpace(line)
	entry := models.LogEntry{
		Container: container,
		Message:   line,
	}

	// Parse timestamp if present
	if hasTimestamp && len(line) > 30 {
		// Kubernetes log format: 2024-01-15T10:30:45.123456789Z message
		timestampStr := line[:30]
		if timestamp, err := time.Parse(time.RFC3339Nano, timestampStr); err == nil {
			entry.Timestamp = timestamp
			entry.Message = strings.TrimSpace(line[30:])
		}
	}

	// If no timestamp was parsed, use current time
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Detect log level from message
	entry.Level = models.DetectLogLevel(entry.Message)

	return entry
}

// GetPodContainers returns a list of container names for a pod
func (c *Client) GetPodContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	pod, err := c.GetPod(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}

	containers := make([]string, 0, len(pod.Spec.Containers)+len(pod.Spec.InitContainers))

	for _, container := range pod.Spec.InitContainers {
		containers = append(containers, container.Name+" (init)")
	}

	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}

	return containers, nil
}

// HasPodRestartedRecently checks if a pod has restarted recently
func (c *Client) HasPodRestartedRecently(ctx context.Context, namespace, podName string) (bool, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	pod, err := c.GetPod(ctx, namespace, podName)
	if err != nil {
		return false, err
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.RestartCount > 0 {
			return true, nil
		}
		if containerStatus.LastTerminationState.Terminated != nil {
			return true, nil
		}
	}

	return false, nil
}
