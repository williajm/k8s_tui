package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewResourceWatcher(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
	if watcher.resourceType != ResourceTypePod {
		t.Errorf("Expected ResourceTypePod, got %v", watcher.resourceType)
	}
	if watcher.namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", watcher.namespace)
	}
	if watcher.state != StateDisconnected {
		t.Errorf("Expected StateDisconnected, got %v", watcher.state)
	}
}

func TestResourceWatcherSetDebugMode(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	// Initially false
	if watcher.debugMode {
		t.Error("Expected debugMode to be false initially")
	}

	// Enable debug mode
	watcher.SetDebugMode(true)
	if !watcher.debugMode {
		t.Error("Expected debugMode to be true after enabling")
	}

	// Disable debug mode
	watcher.SetDebugMode(false)
	if watcher.debugMode {
		t.Error("Expected debugMode to be false after disabling")
	}
}

func TestResourceWatcherGetState(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	// Initial state
	if watcher.GetState() != StateDisconnected {
		t.Errorf("Expected StateDisconnected, got %v", watcher.GetState())
	}

	// Change state
	watcher.setState(StateConnected)
	if watcher.GetState() != StateConnected {
		t.Errorf("Expected StateConnected, got %v", watcher.GetState())
	}
}

func TestResourceWatcherGetResourceVersion(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	// Initially empty
	if watcher.GetResourceVersion() != "" {
		t.Errorf("Expected empty resource version, got '%s'", watcher.GetResourceVersion())
	}

	// Set resource version
	watcher.setResourceVersion("12345")
	if watcher.GetResourceVersion() != "12345" {
		t.Errorf("Expected resource version '12345', got '%s'", watcher.GetResourceVersion())
	}
}

func TestResourceWatcherStop(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	// Set up a cancel func
	ctx, cancel := context.WithCancel(context.Background())
	watcher.mu.Lock()
	watcher.cancelFunc = cancel
	watcher.state = StateConnected
	watcher.mu.Unlock()

	// Stop should cancel context and set state to disconnected
	watcher.Stop()

	if watcher.GetState() != StateDisconnected {
		t.Errorf("Expected StateDisconnected after stop, got %v", watcher.GetState())
	}
	if watcher.cancelFunc != nil {
		t.Error("Expected cancelFunc to be nil after stop")
	}

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Expected context to be cancelled")
	}
}

func TestResourceWatcherStartAndStop(t *testing.T) {
	// Create a pod for testing
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-pod",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: "nginx:latest"},
			},
		},
	}

	fakeClientset := fake.NewSimpleClientset(pod)
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventChan := make(chan WatchEvent, 10)
	errorChan := make(chan WatchError, 10)

	// Start the watcher
	watcher.Start(ctx, eventChan, errorChan)

	// Give it time to initialize
	time.Sleep(100 * time.Millisecond)

	// Stop the watcher
	watcher.Stop()

	// Verify state is disconnected
	if watcher.GetState() != StateDisconnected {
		t.Errorf("Expected StateDisconnected after stop, got %v", watcher.GetState())
	}
}

func TestResourceTypeString(t *testing.T) {
	tests := []struct {
		resourceType ResourceType
		expected     string
	}{
		{ResourceTypePod, "Pod"},
		{ResourceTypeService, "Service"},
		{ResourceTypeDeployment, "Deployment"},
		{ResourceTypeStatefulSet, "StatefulSet"},
		{ResourceTypeEvent, "Event"},
		{ResourceType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.resourceType.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConnectionStateString(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateDisconnected, "Disconnected"},
		{StateConnecting, "Connecting"},
		{StateConnected, "Connected"},
		{StateReconnecting, "Reconnecting"},
		{StateError, "Error"},
		{ConnectionState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestUpdateResourceVersionFromObject(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	tests := []struct {
		name     string
		obj      runtime.Object
		expected string
	}{
		{
			name: "Pod",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "123"},
			},
			expected: "123",
		},
		{
			name: "Service",
			obj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "456"},
			},
			expected: "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watcher.updateResourceVersionFromObject(tt.obj)
			if err != nil {
				t.Fatalf("updateResourceVersionFromObject failed: %v", err)
			}
			if watcher.GetResourceVersion() != tt.expected {
				t.Errorf("Expected resource version '%s', got '%s'", tt.expected, watcher.GetResourceVersion())
			}
		})
	}
}

func TestGetObjectName(t *testing.T) {
	tests := []struct {
		name     string
		obj      runtime.Object
		expected string
	}{
		{
			name: "Pod",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			},
			expected: "default/test-pod",
		},
		{
			name: "Service",
			obj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "kube-system",
				},
			},
			expected: "kube-system/test-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getObjectName(tt.obj)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPerformInitialList(t *testing.T) {
	// Create test pods
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod1",
			Namespace:       "default",
			ResourceVersion: "100",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: "nginx:latest"},
			},
		},
	}
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod2",
			Namespace:       "default",
			ResourceVersion: "101",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: "nginx:latest"},
			},
		},
	}

	fakeClientset := fake.NewSimpleClientset(pod1, pod2)
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	ctx := context.Background()
	eventChan := make(chan WatchEvent, 10)

	err := watcher.performInitialList(ctx, eventChan)
	if err != nil {
		t.Fatalf("performInitialList failed: %v", err)
	}

	// Should have received ADDED events for both pods
	receivedEvents := 0
	timeout := time.After(1 * time.Second)
	for {
		select {
		case event := <-eventChan:
			if event.EventType != watch.Added {
				t.Errorf("Expected ADDED event, got %v", event.EventType)
			}
			receivedEvents++
			if receivedEvents == 2 {
				// Got both events
				return
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for events, received %d/2", receivedEvents)
		}
	}
}

func TestHandleWatchEventError410Gone(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")
	watcher.setResourceVersion("12345")

	eventChan := make(chan WatchEvent, 10)
	errorChan := make(chan WatchError, 10)

	// Create a 410 Gone error event
	errorEvent := watch.Event{
		Type: watch.Error,
		Object: &metav1.Status{
			Code:    410,
			Message: "too old resource version",
		},
	}

	err := watcher.handleWatchEvent(errorEvent, eventChan, errorChan)

	// Should return error
	if err == nil {
		t.Error("Expected error for 410 Gone")
	}

	// Resource version should be reset
	if watcher.GetResourceVersion() != "" {
		t.Errorf("Expected resource version to be reset, got '%s'", watcher.GetResourceVersion())
	}
}

func TestHandleWatchEventBookmark(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")
	watcher.setResourceVersion("100")

	eventChan := make(chan WatchEvent, 10)
	errorChan := make(chan WatchError, 10)

	// Create a bookmark event
	bookmarkEvent := watch.Event{
		Type: watch.Bookmark,
		Object: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "200",
			},
		},
	}

	err := watcher.handleWatchEvent(bookmarkEvent, eventChan, errorChan)
	if err != nil {
		t.Fatalf("handleWatchEvent failed: %v", err)
	}

	// Resource version should be updated
	if watcher.GetResourceVersion() != "200" {
		t.Errorf("Expected resource version '200', got '%s'", watcher.GetResourceVersion())
	}
}

func TestHandleWatchEventAddedModifiedDeleted(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	eventChan := make(chan WatchEvent, 10)
	errorChan := make(chan WatchError, 10)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-pod",
			Namespace:       "default",
			ResourceVersion: "123",
		},
	}

	// Test ADDED event
	addedEvent := watch.Event{
		Type:   watch.Added,
		Object: pod,
	}

	err := watcher.handleWatchEvent(addedEvent, eventChan, errorChan)
	if err != nil {
		t.Fatalf("handleWatchEvent for ADDED failed: %v", err)
	}

	// Should have received the event
	select {
	case event := <-eventChan:
		if event.EventType != watch.Added {
			t.Errorf("Expected ADDED event, got %v", event.EventType)
		}
		if event.ResourceType != ResourceTypePod {
			t.Errorf("Expected ResourceTypePod, got %v", event.ResourceType)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for ADDED event")
	}

	// Resource version should be updated
	if watcher.GetResourceVersion() != "123" {
		t.Errorf("Expected resource version '123', got '%s'", watcher.GetResourceVersion())
	}
}

func TestCreateWatcherForDifferentResourceTypes(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
	}{
		{"Pod", ResourceTypePod},
		{"Service", ResourceTypeService},
		{"Deployment", ResourceTypeDeployment},
		{"StatefulSet", ResourceTypeStatefulSet},
		{"Event", ResourceTypeEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset()
			client := &Client{
				clientset: fakeClientset,
				namespace: "default",
			}

			watcher := NewResourceWatcher(client, tt.resourceType, "default")
			watcher.setResourceVersion("123")

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			w, err := watcher.createWatcher(ctx)
			if err != nil {
				t.Fatalf("createWatcher for %s failed: %v", tt.name, err)
			}
			defer w.Stop()

			if w == nil {
				t.Fatalf("Expected non-nil watcher for %s", tt.name)
			}
		})
	}
}
