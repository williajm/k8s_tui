package models

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewEventInfo(t *testing.T) {
	now := time.Now()
	firstTime := metav1.Time{Time: now.Add(-10 * time.Minute)}
	lastTime := metav1.Time{Time: now.Add(-5 * time.Minute)}

	tests := []struct {
		name      string
		event     *corev1.Event
		wantType  string
		wantCount int32
	}{
		{
			name: "normal event with involved object",
			event: &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-event",
					Namespace: "default",
				},
				Type:    "Normal",
				Reason:  "Created",
				Message: "Pod created successfully",
				InvolvedObject: corev1.ObjectReference{
					Kind: "Pod",
					Name: "nginx",
				},
				Count:          5,
				FirstTimestamp: firstTime,
				LastTimestamp:  lastTime,
			},
			wantType:  "Normal",
			wantCount: 5,
		},
		{
			name: "warning event",
			event: &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "warning-event",
					Namespace: "kube-system",
				},
				Type:    "Warning",
				Reason:  "BackOff",
				Message: "Back-off restarting failed container",
				InvolvedObject: corev1.ObjectReference{
					Kind: "Pod",
					Name: "api-server",
				},
				Count:          10,
				FirstTimestamp: firstTime,
				LastTimestamp:  lastTime,
			},
			wantType:  "Warning",
			wantCount: 10,
		},
		{
			name: "event with EventTime",
			event: &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "event-time",
					Namespace: "default",
				},
				Type:      "Normal",
				Reason:    "Started",
				Message:   "Container started",
				EventTime: metav1.MicroTime{Time: now},
				InvolvedObject: corev1.ObjectReference{
					Kind: "Pod",
					Name: "test",
				},
			},
			wantType:  "Normal",
			wantCount: 0,
		},
		{
			name: "event without involved object",
			event: &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-object",
					Namespace: "default",
				},
				Type:           "Normal",
				Reason:         "NodeReady",
				Message:        "Node is ready",
				LastTimestamp:  lastTime,
				FirstTimestamp: firstTime,
			},
			wantType:  "Normal",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewEventInfo(tt.event)

			if info.Name != tt.event.Name {
				t.Errorf("NewEventInfo().Name = %v, want %v", info.Name, tt.event.Name)
			}

			if info.Namespace != tt.event.Namespace {
				t.Errorf("NewEventInfo().Namespace = %v, want %v", info.Namespace, tt.event.Namespace)
			}

			if info.Type != tt.wantType {
				t.Errorf("NewEventInfo().Type = %v, want %v", info.Type, tt.wantType)
			}

			if info.Reason != tt.event.Reason {
				t.Errorf("NewEventInfo().Reason = %v, want %v", info.Reason, tt.event.Reason)
			}

			if info.Message != tt.event.Message {
				t.Errorf("NewEventInfo().Message = %v, want %v", info.Message, tt.event.Message)
			}

			if info.Count != tt.wantCount {
				t.Errorf("NewEventInfo().Count = %v, want %v", info.Count, tt.wantCount)
			}

			if info.Event != tt.event {
				t.Error("NewEventInfo().Event should reference original event")
			}

			// Check involved object formatting
			if tt.event.InvolvedObject.Kind != "" && tt.event.InvolvedObject.Name != "" {
				expectedObject := tt.event.InvolvedObject.Kind + "/" + tt.event.InvolvedObject.Name
				if info.Object != expectedObject {
					t.Errorf("NewEventInfo().Object = %v, want %v", info.Object, expectedObject)
				}
			}
		})
	}
}

func TestEventInfo_GetTypeSymbol(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		want      string
	}{
		{
			name:      "normal event",
			eventType: "Normal",
			want:      "ℹ",
		},
		{
			name:      "warning event",
			eventType: "Warning",
			want:      "⚠",
		},
		{
			name:      "error event",
			eventType: "Error",
			want:      "✖",
		},
		{
			name:      "unknown event type",
			eventType: "Unknown",
			want:      "•",
		},
		{
			name:      "empty event type",
			eventType: "",
			want:      "•",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &EventInfo{Type: tt.eventType}
			got := info.GetTypeSymbol()
			if got != tt.want {
				t.Errorf("GetTypeSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventInfo_FormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		lastTime     time.Time
		wantContains string
	}{
		{
			name:         "seconds ago",
			lastTime:     now.Add(-30 * time.Second),
			wantContains: "s",
		},
		{
			name:         "minutes ago",
			lastTime:     now.Add(-5 * time.Minute),
			wantContains: "m",
		},
		{
			name:         "hours ago",
			lastTime:     now.Add(-2 * time.Hour),
			wantContains: "h",
		},
		{
			name:         "days ago",
			lastTime:     now.Add(-3 * 24 * time.Hour),
			wantContains: "d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &EventInfo{
				LastTimestamp: tt.lastTime,
			}
			got := info.FormatAge()
			if got == "" {
				t.Error("FormatAge() returned empty string")
			}
			// Just verify it returns a non-empty string
			// The exact format depends on formatAge implementation
		})
	}
}

func TestEventInfo_GetAgeSeconds(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		lastTime time.Time
		wantMin  int64
		wantMax  int64
	}{
		{
			name:     "30 seconds ago",
			lastTime: now.Add(-30 * time.Second),
			wantMin:  25,
			wantMax:  35,
		},
		{
			name:     "5 minutes ago",
			lastTime: now.Add(-5 * time.Minute),
			wantMin:  295,
			wantMax:  305,
		},
		{
			name:     "1 hour ago",
			lastTime: now.Add(-1 * time.Hour),
			wantMin:  3595,
			wantMax:  3605,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &EventInfo{
				LastTimestamp: tt.lastTime,
			}
			got := info.GetAgeSeconds()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetAgeSeconds() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestEventInfo_IsRecent(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		lastTime time.Time
		want     bool
	}{
		{
			name:     "recent - 1 minute ago",
			lastTime: now.Add(-1 * time.Minute),
			want:     true,
		},
		{
			name:     "recent - 4 minutes ago",
			lastTime: now.Add(-4 * time.Minute),
			want:     true,
		},
		{
			name:     "not recent - 6 minutes ago",
			lastTime: now.Add(-6 * time.Minute),
			want:     false,
		},
		{
			name:     "not recent - 1 hour ago",
			lastTime: now.Add(-1 * time.Hour),
			want:     false,
		},
		{
			name:     "recent - just now",
			lastTime: now,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &EventInfo{
				LastTimestamp: tt.lastTime,
			}
			got := info.IsRecent()
			if got != tt.want {
				t.Errorf("IsRecent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventInfo_GetMessagePreview(t *testing.T) {
	tests := []struct {
		name    string
		message string
		maxLen  int
		want    string
	}{
		{
			name:    "short message",
			message: "Pod created",
			maxLen:  50,
			want:    "Pod created",
		},
		{
			name:    "exact length",
			message: "This is exactly 20c",
			maxLen:  20,
			want:    "This is exactly 20c",
		},
		{
			name:    "long message truncated",
			message: "This is a very long message that should be truncated",
			maxLen:  20,
			want:    "This is a very lo...",
		},
		{
			name:    "very long message",
			message: "Back-off restarting failed container nginx in pod default/nginx-deployment-7d64f9c8d5-abc12",
			maxLen:  40,
			want:    "Back-off restarting failed container ...",
		},
		{
			name:    "empty message",
			message: "",
			maxLen:  20,
			want:    "",
		},
		{
			name:    "single character",
			message: "X",
			maxLen:  20,
			want:    "X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &EventInfo{Message: tt.message}
			got := info.GetMessagePreview(tt.maxLen)
			if got != tt.want {
				t.Errorf("GetMessagePreview(%d) = %q, want %q", tt.maxLen, got, tt.want)
			}

			// Verify length constraint
			if len(got) > tt.maxLen {
				t.Errorf("GetMessagePreview(%d) returned string longer than maxLen: %d > %d", tt.maxLen, len(got), tt.maxLen)
			}
		})
	}
}

func TestEventInfo_TimestampHandling(t *testing.T) {
	now := time.Now()
	eventTime := metav1.MicroTime{Time: now}
	firstTime := metav1.Time{Time: now.Add(-10 * time.Minute)}
	lastTime := metav1.Time{Time: now.Add(-5 * time.Minute)}

	t.Run("EventTime takes precedence", func(t *testing.T) {
		event := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			EventTime:      eventTime,
			FirstTimestamp: firstTime,
			LastTimestamp:  lastTime,
		}

		info := NewEventInfo(event)

		// EventTime should be used for both first and last
		if !info.FirstTimestamp.Equal(eventTime.Time) {
			t.Errorf("FirstTimestamp = %v, want %v (from EventTime)", info.FirstTimestamp, eventTime.Time)
		}
		if !info.LastTimestamp.Equal(eventTime.Time) {
			t.Errorf("LastTimestamp = %v, want %v (from EventTime)", info.LastTimestamp, eventTime.Time)
		}
	})

	t.Run("Uses LastTimestamp and FirstTimestamp when EventTime is zero", func(t *testing.T) {
		event := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			FirstTimestamp: firstTime,
			LastTimestamp:  lastTime,
		}

		info := NewEventInfo(event)

		if !info.FirstTimestamp.Equal(firstTime.Time) {
			t.Errorf("FirstTimestamp = %v, want %v", info.FirstTimestamp, firstTime.Time)
		}
		if !info.LastTimestamp.Equal(lastTime.Time) {
			t.Errorf("LastTimestamp = %v, want %v", info.LastTimestamp, lastTime.Time)
		}
	})

	t.Run("Uses LastTimestamp when FirstTimestamp is zero", func(t *testing.T) {
		event := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			LastTimestamp: lastTime,
		}

		info := NewEventInfo(event)

		if !info.FirstTimestamp.Equal(lastTime.Time) {
			t.Errorf("FirstTimestamp = %v, want %v (fallback to LastTimestamp)", info.FirstTimestamp, lastTime.Time)
		}
		if !info.LastTimestamp.Equal(lastTime.Time) {
			t.Errorf("LastTimestamp = %v, want %v", info.LastTimestamp, lastTime.Time)
		}
	})
}
