package components

import (
	"strings"
	"testing"

	"github.com/williajm/k8s-tui/internal/models"
)

func TestNewDetailView(t *testing.T) {
	d := NewDetailView()

	if d == nil {
		t.Fatal("NewDetailView returned nil")
	}

	if d.width != 80 {
		t.Errorf("expected default width 80, got %d", d.width)
	}

	if d.height != 20 {
		t.Errorf("expected default height 20, got %d", d.height)
	}
}

func TestDetailView_SetSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "standard size",
			width:  120,
			height: 30,
		},
		{
			name:   "small size",
			width:  40,
			height: 10,
		},
		{
			name:   "large size",
			width:  200,
			height: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetailView()
			d.SetSize(tt.width, tt.height)

			if d.width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, d.width)
			}

			if d.height != tt.height {
				t.Errorf("expected height %d, got %d", tt.height, d.height)
			}
		})
	}
}

func TestDetailView_ViewPod(t *testing.T) {
	tests := []struct {
		name            string
		pod             *models.PodInfo
		expectedStrings []string
		nilPod          bool
	}{
		{
			name: "valid pod",
			pod: &models.PodInfo{
				Name:      "test-pod",
				Namespace: "default",
				Status:    "Running",
				Ready:     "1/1",
				Restarts:  0,
				Age:       "5m",
				IP:        "10.0.0.1",
				Node:      "node-1",
				Containers: []models.ContainerInfo{
					{
						Name:         "main",
						Image:        "nginx:latest",
						Ready:        true,
						State:        "Running",
						RestartCount: 0,
					},
				},
			},
			expectedStrings: []string{
				"Pod Details",
				"Name",
				"test-pod",
				"Namespace",
				"default",
				"Status",
				"Running",
				"Ready",
				"1/1",
				"Restarts",
				"Age",
				"5m",
				"IP",
				"10.0.0.1",
				"Node",
				"node-1",
				"Containers",
				"main",
				"nginx:latest",
			},
		},
		{
			name:            "nil pod",
			pod:             nil,
			nilPod:          true,
			expectedStrings: []string{"No pod selected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetailView()
			view := d.ViewPod(tt.pod)

			if view == "" {
				t.Fatal("ViewPod returned empty string")
			}

			for _, expected := range tt.expectedStrings {
				if !strings.Contains(view, expected) {
					t.Errorf("expected view to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

func TestDetailView_ViewPodWithMultipleContainers(t *testing.T) {
	pod := &models.PodInfo{
		Name:      "multi-container-pod",
		Namespace: "default",
		Status:    "Running",
		Ready:     "2/2",
		Restarts:  1,
		Age:       "10m",
		IP:        "10.0.0.2",
		Node:      "node-1",
		Containers: []models.ContainerInfo{
			{
				Name:         "app",
				Image:        "app:v1",
				Ready:        true,
				State:        "Running",
				RestartCount: 0,
			},
			{
				Name:         "sidecar",
				Image:        "sidecar:v2",
				Ready:        true,
				State:        "Running",
				RestartCount: 1,
			},
		},
	}

	d := NewDetailView()
	view := d.ViewPod(pod)

	// Check both containers are displayed
	if !strings.Contains(view, "app") {
		t.Error("expected view to contain first container name")
	}

	if !strings.Contains(view, "sidecar") {
		t.Error("expected view to contain second container name")
	}

	if !strings.Contains(view, "app:v1") {
		t.Error("expected view to contain first container image")
	}

	if !strings.Contains(view, "sidecar:v2") {
		t.Error("expected view to contain second container image")
	}
}

func TestDetailView_ViewService(t *testing.T) {
	tests := []struct {
		name            string
		service         *models.ServiceInfo
		expectedStrings []string
		nilService      bool
	}{
		{
			name: "valid ClusterIP service",
			service: &models.ServiceInfo{
				Name:       "test-service",
				Namespace:  "default",
				Type:       "ClusterIP",
				ClusterIP:  "10.0.1.1",
				ExternalIP: "<none>",
				Ports:      "80/TCP",
				Age:        "1h",
				Selector: map[string]string{
					"app": "test",
				},
			},
			expectedStrings: []string{
				"Service Details",
				"Name",
				"test-service",
				"Namespace",
				"default",
				"Type",
				"ClusterIP",
				"Cluster IP",
				"10.0.1.1",
				"Ports",
				"80/TCP",
				"Age",
				"1h",
				"Selector",
				"app",
				"test",
			},
		},
		{
			name:            "nil service",
			service:         nil,
			nilService:      true,
			expectedStrings: []string{"No service selected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetailView()
			view := d.ViewService(tt.service)

			if view == "" {
				t.Fatal("ViewService returned empty string")
			}

			for _, expected := range tt.expectedStrings {
				if !strings.Contains(view, expected) {
					t.Errorf("expected view to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

func TestDetailView_ViewServiceWithMultipleSelectors(t *testing.T) {
	service := &models.ServiceInfo{
		Name:       "multi-selector-service",
		Namespace:  "default",
		Type:       "LoadBalancer",
		ClusterIP:  "10.0.1.2",
		ExternalIP: "203.0.113.1",
		Ports:      "80/TCP,443/TCP",
		Age:        "2d",
		Selector: map[string]string{
			"app":     "myapp",
			"version": "v1",
			"tier":    "frontend",
		},
	}

	d := NewDetailView()
	view := d.ViewService(service)

	// Check all selectors are displayed
	expectedKeys := []string{"app", "version", "tier"}
	for _, key := range expectedKeys {
		if !strings.Contains(view, key) {
			t.Errorf("expected view to contain selector key %q", key)
		}
	}
}

func TestDetailView_ViewDeployment(t *testing.T) {
	tests := []struct {
		name            string
		deployment      *models.DeploymentInfo
		expectedStrings []string
		nilDeployment   bool
	}{
		{
			name: "valid deployment",
			deployment: &models.DeploymentInfo{
				Name:      "test-deployment",
				Namespace: "default",
				Replicas:  3,
				Ready:     "3/3",
				UpToDate:  3,
				Available: 3,
				Strategy:  "RollingUpdate",
				Age:       "7d",
			},
			expectedStrings: []string{
				"Deployment Details",
				"Name",
				"test-deployment",
				"Namespace",
				"default",
				"Replicas",
				"3",
				"Ready",
				"3/3",
				"Up-to-date",
				"Available",
				"Strategy",
				"RollingUpdate",
				"Age",
				"7d",
			},
		},
		{
			name:            "nil deployment",
			deployment:      nil,
			nilDeployment:   true,
			expectedStrings: []string{"No deployment selected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetailView()
			view := d.ViewDeployment(tt.deployment)

			if view == "" {
				t.Fatal("ViewDeployment returned empty string")
			}

			for _, expected := range tt.expectedStrings {
				if !strings.Contains(view, expected) {
					t.Errorf("expected view to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

func TestDetailView_ViewStatefulSet(t *testing.T) {
	tests := []struct {
		name            string
		statefulSet     *models.StatefulSetInfo
		expectedStrings []string
		nilStatefulSet  bool
	}{
		{
			name: "valid statefulset",
			statefulSet: &models.StatefulSetInfo{
				Name:      "test-statefulset",
				Namespace: "default",
				Replicas:  5,
				Ready:     "5/5",
				Strategy:  "RollingUpdate",
				Age:       "30d",
			},
			expectedStrings: []string{
				"StatefulSet Details",
				"Name",
				"test-statefulset",
				"Namespace",
				"default",
				"Replicas",
				"5",
				"Ready",
				"5/5",
				"Strategy",
				"RollingUpdate",
				"Age",
				"30d",
			},
		},
		{
			name:            "nil statefulset",
			statefulSet:     nil,
			nilStatefulSet:  true,
			expectedStrings: []string{"No statefulset selected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetailView()
			view := d.ViewStatefulSet(tt.statefulSet)

			if view == "" {
				t.Fatal("ViewStatefulSet returned empty string")
			}

			for _, expected := range tt.expectedStrings {
				if !strings.Contains(view, expected) {
					t.Errorf("expected view to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

func TestDetailView_AllResourceTypes(t *testing.T) {
	d := NewDetailView()
	d.SetSize(120, 40)

	// Create test resources
	pod := &models.PodInfo{
		Name:      "test-pod",
		Namespace: "default",
		Status:    "Running",
		Ready:     "1/1",
		Restarts:  0,
		Age:       "5m",
		IP:        "10.0.0.1",
		Node:      "node-1",
		Containers: []models.ContainerInfo{
			{
				Name:         "main",
				Image:        "nginx:latest",
				Ready:        true,
				State:        "Running",
				RestartCount: 0,
			},
		},
	}

	service := &models.ServiceInfo{
		Name:       "test-service",
		Namespace:  "default",
		Type:       "ClusterIP",
		ClusterIP:  "10.0.1.1",
		ExternalIP: "<none>",
		Ports:      "80/TCP",
		Age:        "1h",
		Selector:   map[string]string{"app": "test"},
	}

	deployment := &models.DeploymentInfo{
		Name:      "test-deployment",
		Namespace: "default",
		Replicas:  3,
		Ready:     "3/3",
		UpToDate:  3,
		Available: 3,
		Strategy:  "RollingUpdate",
		Age:       "7d",
	}

	statefulSet := &models.StatefulSetInfo{
		Name:      "test-statefulset",
		Namespace: "default",
		Replicas:  5,
		Ready:     "5/5",
		Strategy:  "RollingUpdate",
		Age:       "30d",
	}

	// Test all view methods
	views := map[string]string{
		"pod":         d.ViewPod(pod),
		"service":     d.ViewService(service),
		"deployment":  d.ViewDeployment(deployment),
		"statefulset": d.ViewStatefulSet(statefulSet),
	}

	for resourceType, view := range views {
		if view == "" {
			t.Errorf("%s view returned empty string", resourceType)
		}

		if !strings.Contains(view, "Details") {
			t.Errorf("%s view should contain 'Details' header", resourceType)
		}
	}
}
