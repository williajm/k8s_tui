package app

import (
	"testing"

	"github.com/williajm/k8s-tui/internal/config"
	"github.com/williajm/k8s-tui/internal/k8s"
	"github.com/williajm/k8s-tui/internal/models"
	"github.com/williajm/k8s-tui/internal/ui/components"
	"k8s.io/client-go/kubernetes/fake"
)

// TestViewModeConstants verifies ViewMode constants are correct
func TestViewModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		mode     ViewMode
		expected int
	}{
		{"ViewModeList", ViewModeList, 0},
		{"ViewModeDetail", ViewModeDetail, 1},
		{"ViewModeLogStream", ViewModeLogStream, 2},
		{"ViewModeDescribe", ViewModeDescribe, 3},
		{"ViewModeContainerSelect", ViewModeContainerSelect, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.mode) != tt.expected {
				t.Errorf("ViewMode %s = %d, want %d", tt.name, tt.mode, tt.expected)
			}
		})
	}
}

// TestNewModelWithConfig tests model creation with configuration
func TestNewModelWithConfig(t *testing.T) {
	// Create a fake kubernetes client
	fakeClientset := fake.NewSimpleClientset()
	client := &k8s.Client{}
	client.SetClientsetForTesting(fakeClientset)

	cfg := config.DefaultConfig()
	model := NewModelWithConfig(client, cfg)

	// Verify initial state
	if model.viewMode != ViewModeList {
		t.Errorf("Initial viewMode = %v, want ViewModeList", model.viewMode)
	}

	if model.loading != true {
		t.Errorf("Initial loading = %v, want true", model.loading)
	}

	if model.searchMode != false {
		t.Errorf("Initial searchMode = %v, want false", model.searchMode)
	}

	if model.showHelp != false {
		t.Errorf("Initial showHelp = %v, want false", model.showHelp)
	}

	if model.useWatchAPI != true {
		t.Errorf("Initial useWatchAPI = %v, want true", model.useWatchAPI)
	}
}

// TestResourcesLoadedMsg tests the resourcesLoadedMsg struct
func TestResourcesLoadedMsg(t *testing.T) {
	msg := resourcesLoadedMsg{
		resourceType: components.ResourceTypePod,
		pods: []models.PodInfo{
			{Name: "test-pod", Namespace: "default"},
		},
		err: nil,
	}

	if msg.resourceType != components.ResourceTypePod {
		t.Errorf("resourceType = %v, want ResourceTypePod", msg.resourceType)
	}

	if len(msg.pods) != 1 {
		t.Errorf("len(pods) = %d, want 1", len(msg.pods))
	}

	if msg.pods[0].Name != "test-pod" {
		t.Errorf("pods[0].Name = %s, want test-pod", msg.pods[0].Name)
	}
}

// TestNamespacesLoadedMsg tests the namespacesLoadedMsg struct
func TestNamespacesLoadedMsg(t *testing.T) {
	msg := namespacesLoadedMsg{
		namespaces: []string{"default", "kube-system", "production"},
		err:        nil,
	}

	if len(msg.namespaces) != 3 {
		t.Errorf("len(namespaces) = %d, want 3", len(msg.namespaces))
	}

	if msg.namespaces[0] != "default" {
		t.Errorf("namespaces[0] = %s, want default", msg.namespaces[0])
	}
}

// TestDescribeLoadedMsg tests the describeLoadedMsg struct
func TestDescribeLoadedMsg(t *testing.T) {
	data := models.NewDescribeData("Pod", "test-pod", "default")
	msg := describeLoadedMsg{
		data: data,
		yaml: "apiVersion: v1\nkind: Pod",
		json: `{"apiVersion": "v1", "kind": "Pod"}`,
	}

	if msg.data.Kind != "Pod" {
		t.Errorf("data.Kind = %s, want Pod", msg.data.Kind)
	}

	if msg.yaml == "" {
		t.Error("yaml should not be empty")
	}

	if msg.json == "" {
		t.Error("json should not be empty")
	}
}

// TestContainersLoadedMsg tests the containersLoadedMsg struct
func TestContainersLoadedMsg(t *testing.T) {
	msg := containersLoadedMsg{
		containers: []string{"main", "sidecar", "init (init)"},
		err:        nil,
	}

	if len(msg.containers) != 3 {
		t.Errorf("len(containers) = %d, want 3", len(msg.containers))
	}
}

// TestErrMsg tests the errMsg struct
func TestErrMsg(t *testing.T) {
	testErr := errMsg{err: nil}
	if testErr.err != nil {
		t.Errorf("err = %v, want nil", testErr.err)
	}
}

// TestWatchEventMsg tests the watchEventMsg struct
func TestWatchEventMsg(t *testing.T) {
	event := k8s.WatchEvent{
		ResourceType: k8s.ResourceTypePod,
		EventType:    "ADDED",
	}
	msg := watchEventMsg{event: event}

	if msg.event.ResourceType != k8s.ResourceTypePod {
		t.Errorf("ResourceType = %v, want ResourceTypePod", msg.event.ResourceType)
	}

	if msg.event.EventType != "ADDED" {
		t.Errorf("EventType = %s, want ADDED", msg.event.EventType)
	}
}

// TestWatchErrorMsg tests the watchErrorMsg struct
func TestWatchErrorMsg(t *testing.T) {
	watchErr := k8s.WatchError{
		ResourceType: k8s.ResourceTypeService,
		Fatal:        false,
	}
	msg := watchErrorMsg{error: watchErr}

	if msg.error.ResourceType != k8s.ResourceTypeService {
		t.Errorf("ResourceType = %v, want ResourceTypeService", msg.error.ResourceType)
	}

	if msg.error.Fatal {
		t.Error("Fatal should be false")
	}
}

// TestConnectionStateMsg tests the connectionStateMsg struct
func TestConnectionStateMsg(t *testing.T) {
	msg := connectionStateMsg{state: k8s.StateConnected}

	if msg.state != k8s.StateConnected {
		t.Errorf("state = %v, want StateConnected", msg.state)
	}
}
