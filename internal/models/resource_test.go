package models

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPodInfo(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		wantName string
		wantNS   string
	}{
		{
			name: "basic pod",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "10.0.0.1",
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "nginx",
							Ready:        true,
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			wantName: "test-pod",
			wantNS:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewPodInfo(tt.pod)

			if info.Name != tt.wantName {
				t.Errorf("NewPodInfo().Name = %v, want %v", info.Name, tt.wantName)
			}

			if info.Namespace != tt.wantNS {
				t.Errorf("NewPodInfo().Namespace = %v, want %v", info.Namespace, tt.wantNS)
			}

			if info.Status != "Running" {
				t.Errorf("NewPodInfo().Status = %v, want Running", info.Status)
			}

			if info.Ready != "1/1" {
				t.Errorf("NewPodInfo().Ready = %v, want 1/1", info.Ready)
			}

			if info.IP != "10.0.0.1" {
				t.Errorf("NewPodInfo().IP = %v, want 10.0.0.1", info.IP)
			}
		})
	}
}

func TestPodInfo_GetStatusSymbol(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		ready      string
		containers int
		want       string
	}{
		{
			name:       "running and ready",
			status:     "Running",
			ready:      "1/1",
			containers: 1,
			want:       "✓",
		},
		{
			name:       "succeeded",
			status:     "Succeeded",
			ready:      "0/1",
			containers: 1,
			want:       "✓",
		},
		{
			name:       "failed",
			status:     "Failed",
			ready:      "0/1",
			containers: 1,
			want:       "✗",
		},
		{
			name:       "error",
			status:     "Error",
			ready:      "0/1",
			containers: 1,
			want:       "✗",
		},
		{
			name:       "pending",
			status:     "Pending",
			ready:      "0/1",
			containers: 1,
			want:       "○",
		},
		{
			name:       "terminating",
			status:     "Terminating",
			ready:      "0/1",
			containers: 1,
			want:       "⊗",
		},
		{
			name:       "unknown",
			status:     "Unknown",
			ready:      "0/1",
			containers: 1,
			want:       "⚠",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PodInfo{
				Status:     tt.status,
				Ready:      tt.ready,
				Containers: make([]ContainerInfo, tt.containers),
			}

			got := p.GetStatusSymbol()
			if got != tt.want {
				t.Errorf("GetStatusSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodInfo_IsHealthy(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		ready      string
		containers int
		want       bool
	}{
		{
			name:       "healthy pod",
			status:     "Running",
			ready:      "1/1",
			containers: 1,
			want:       true,
		},
		{
			name:       "not ready",
			status:     "Running",
			ready:      "0/1",
			containers: 1,
			want:       false,
		},
		{
			name:       "pending",
			status:     "Pending",
			ready:      "0/1",
			containers: 1,
			want:       false,
		},
		{
			name:       "multi-container healthy",
			status:     "Running",
			ready:      "2/2",
			containers: 2,
			want:       true,
		},
		{
			name:       "multi-container partial",
			status:     "Running",
			ready:      "1/2",
			containers: 2,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PodInfo{
				Status:     tt.status,
				Ready:      tt.ready,
				Containers: make([]ContainerInfo, tt.containers),
			}

			got := p.IsHealthy()
			if got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name      string
		timestamp metav1.Time
		want      string
	}{
		{
			name: "seconds",
			timestamp: metav1.Time{
				Time: time.Now().Add(-30 * time.Second),
			},
			want: "30s",
		},
		{
			name: "minutes",
			timestamp: metav1.Time{
				Time: time.Now().Add(-5 * time.Minute),
			},
			want: "5m",
		},
		{
			name: "hours",
			timestamp: metav1.Time{
				Time: time.Now().Add(-3 * time.Hour),
			},
			want: "3h",
		},
		{
			name: "days",
			timestamp: metav1.Time{
				Time: time.Now().Add(-48 * time.Hour),
			},
			want: "2d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.timestamp)
			if got != tt.want {
				t.Errorf("formatAge() = %v, want %v", got, tt.want)
			}
		})
	}
}
