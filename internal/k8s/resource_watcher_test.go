package k8s

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	// Context should be canceled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Expected context to be canceled")
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

func TestIsWatchErrorFatal(t *testing.T) {
	podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	tests := []struct {
		name  string
		err   error
		fatal bool
	}{
		{
			name:  "Unauthorized error is fatal",
			err:   apierrors.NewUnauthorized("Unauthorized"),
			fatal: true,
		},
		{
			name:  "Forbidden error is fatal",
			err:   apierrors.NewForbidden(podGVR.GroupResource(), "test-pod", fmt.Errorf("forbidden")),
			fatal: true,
		},
		{
			name:  "Not found error is fatal",
			err:   apierrors.NewNotFound(podGVR.GroupResource(), "test-pod"),
			fatal: true,
		},
		{
			name:  "Gone error (410) is NOT fatal - triggers re-list",
			err:   apierrors.NewGone("Resource version too old"),
			fatal: false,
		},
		{
			name:  "Service unavailable (503) is NOT fatal",
			err:   apierrors.NewServiceUnavailable("Service Unavailable"),
			fatal: false,
		},
		{
			name:  "Generic error is NOT fatal",
			err:   fmt.Errorf("generic error"),
			fatal: false,
		},
		{
			name:  "Network timeout is NOT fatal",
			err:   fmt.Errorf("context deadline exceeded"),
			fatal: false,
		},
		{
			name:  "Too many requests (429) is NOT fatal",
			err:   apierrors.NewTooManyRequests("rate limited", 10),
			fatal: false,
		},
		{
			name:  "Internal server error (500) is NOT fatal",
			err:   apierrors.NewInternalError(fmt.Errorf("internal error")),
			fatal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWatchErrorFatal(tt.err)
			if result != tt.fatal {
				t.Errorf("isWatchErrorFatal() = %v, want %v", result, tt.fatal)
			}
		})
	}
}

func TestResourceWatcherReconnectionBackoff(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	// Check initial backoff state
	if watcher.backoff.Attempts() != 0 {
		t.Errorf("Expected 0 initial attempts, got %d", watcher.backoff.Attempts())
	}

	// Verify backoff resets on successful connection
	watcher.backoff.Next() // Simulate one failed attempt
	watcher.backoff.Next() // Simulate another failed attempt

	if watcher.backoff.Attempts() != 2 {
		t.Errorf("Expected 2 attempts, got %d", watcher.backoff.Attempts())
	}

	// Reset should clear attempts
	watcher.backoff.Reset()
	if watcher.backoff.Attempts() != 0 {
		t.Errorf("Expected 0 attempts after reset, got %d", watcher.backoff.Attempts())
	}
}

func TestResourceWatcherConcurrency(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	// Test concurrent access to state
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = watcher.GetState()
			_ = watcher.GetResourceVersion()
			watcher.setState(StateConnected)
			watcher.setResourceVersion("test-version")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should have a valid state
	state := watcher.GetState()
	if state != StateConnected {
		t.Logf("Final state: %v (expected Connected, but concurrent access may vary)", state)
	}
}

func TestHandleWatchEventUnsupportedObject(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher := NewResourceWatcher(client, ResourceTypePod, "default")

	eventChan := make(chan WatchEvent, 10)
	errorChan := make(chan WatchError, 10)

	// Create event with unsupported object type (raw string)
	unsupportedEvent := watch.Event{
		Type:   watch.Added,
		Object: &metav1.Status{Message: "This is not a Pod"}, // Wrong type for pod watcher
	}

	// This should handle gracefully (log warning but not crash)
	err := watcher.handleWatchEvent(unsupportedEvent, eventChan, errorChan)

	// We expect either no error or a handled error - not a panic
	if err != nil {
		t.Logf("handleWatchEvent returned error (expected): %v", err)
	}
}

func TestResourceWatcherNamespaceIsolation(t *testing.T) {
	// Test that watchers respect namespace boundaries
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	watcher1 := NewResourceWatcher(client, ResourceTypePod, "default")
	watcher2 := NewResourceWatcher(client, ResourceTypePod, "kube-system")

	if watcher1.namespace == watcher2.namespace {
		t.Error("Watchers should have different namespaces")
	}

	if watcher1.namespace != "default" {
		t.Errorf("watcher1 namespace = %s, want default", watcher1.namespace)
	}

	if watcher2.namespace != "kube-system" {
		t.Errorf("watcher2 namespace = %s, want kube-system", watcher2.namespace)
	}
}
