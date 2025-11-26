package models

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventInfo represents simplified event information for display
type EventInfo struct {
	Name           string
	Namespace      string
	Type           string // Normal, Warning, Error
	Reason         string
	Message        string
	Object         string // pod/nginx, deployment/api, etc.
	FirstTimestamp time.Time
	LastTimestamp  time.Time
	Count          int32
	Event          *corev1.Event // Keep reference to full event
}

// NewEventInfo creates an EventInfo from a Kubernetes Event
func NewEventInfo(event *corev1.Event) EventInfo {
	info := EventInfo{
		Name:      event.Name,
		Namespace: event.Namespace,
		Type:      event.Type,
		Reason:    event.Reason,
		Message:   event.Message,
		Count:     event.Count,
		Event:     event,
	}

	// Format object reference
	if event.InvolvedObject.Kind != "" && event.InvolvedObject.Name != "" {
		info.Object = fmt.Sprintf("%s/%s",
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
		)
	}

	// Handle timestamp variations
	if !event.EventTime.IsZero() {
		info.LastTimestamp = event.EventTime.Time
		info.FirstTimestamp = event.EventTime.Time
	} else {
		info.LastTimestamp = event.LastTimestamp.Time
		info.FirstTimestamp = event.FirstTimestamp.Time
	}

	// Use last timestamp if first is zero
	if info.FirstTimestamp.IsZero() {
		info.FirstTimestamp = info.LastTimestamp
	}

	return info
}

// GetTypeSymbol returns a visual indicator for event type
func (e *EventInfo) GetTypeSymbol() string {
	switch e.Type {
	case "Normal":
		return "ℹ" // Info symbol for normal events
	case "Warning":
		return "⚠" // Warning triangle
	case "Error":
		return "✖" // X for errors
	default:
		return "•" // Bullet for unknown
	}
}

// FormatAge returns a human-readable age string
func (e *EventInfo) FormatAge() string {
	return formatAge(metav1.Time{Time: e.LastTimestamp})
}

// GetAgeSeconds returns the age in seconds for sorting
func (e *EventInfo) GetAgeSeconds() int64 {
	return int64(time.Since(e.LastTimestamp).Seconds())
}

// IsRecent returns true if the event occurred within the last 5 minutes
func (e *EventInfo) IsRecent() bool {
	return time.Since(e.LastTimestamp) < 5*time.Minute
}

// GetMessagePreview returns a truncated message suitable for list display
func (e *EventInfo) GetMessagePreview(maxLen int) string {
	if len(e.Message) <= maxLen {
		return e.Message
	}
	return e.Message[:maxLen-3] + "..."
}
