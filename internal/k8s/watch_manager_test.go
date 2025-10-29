package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewWatchManager(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	if wm == nil {
		t.Fatal("Expected non-nil watch manager")
	}
	if wm.client != client {
		t.Error("Expected client to be set")
	}
	if wm.watchers == nil {
		t.Error("Expected watchers map to be initialized")
	}
	if wm.eventChan == nil {
		t.Error("Expected event channel to be initialized")
	}
	if wm.errorChan == nil {
		t.Error("Expected error channel to be initialized")
	}
}

func TestWatchManagerSetDebugMode(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	// Initially false
	if wm.debugMode {
		t.Error("Expected debugMode to be false initially")
	}

	// Enable debug mode
	wm.SetDebugMode(true)
	if !wm.debugMode {
		t.Error("Expected debugMode to be true after enabling")
	}

	// Disable debug mode
	wm.SetDebugMode(false)
	if wm.debugMode {
		t.Error("Expected debugMode to be false after disabling")
	}
}

func TestWatchManagerStartAndStop(t *testing.T) {
	// Create test pod
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

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching pods
	resourceTypes := []ResourceType{ResourceTypePod}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watchers time to initialize
	time.Sleep(100 * time.Millisecond)

	// Check watcher count
	if wm.GetWatcherCount() != 1 {
		t.Errorf("Expected 1 watcher, got %d", wm.GetWatcherCount())
	}

	// Check if watching pods
	if !wm.IsWatching(ResourceTypePod) {
		t.Error("Expected to be watching pods")
	}

	// Stop
	wm.Stop()

	// Check watcher count after stop
	if wm.GetWatcherCount() != 0 {
		t.Errorf("Expected 0 watchers after stop, got %d", wm.GetWatcherCount())
	}
}

func TestWatchManagerStartMultipleResourceTypes(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching multiple resource types
	resourceTypes := []ResourceType{
		ResourceTypePod,
		ResourceTypeService,
		ResourceTypeDeployment,
	}

	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watchers time to initialize
	time.Sleep(100 * time.Millisecond)

	// Check watcher count
	if wm.GetWatcherCount() != 3 {
		t.Errorf("Expected 3 watchers, got %d", wm.GetWatcherCount())
	}

	// Check each resource type
	for _, rt := range resourceTypes {
		if !wm.IsWatching(rt) {
			t.Errorf("Expected to be watching %s", rt)
		}
	}

	// Stop
	wm.Stop()
}

func TestWatchManagerRestartWatcher(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching pods
	resourceTypes := []ResourceType{ResourceTypePod}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watchers time to initialize
	time.Sleep(100 * time.Millisecond)

	// Get initial resource version
	versions := wm.GetResourceVersions()
	initialVersion := versions[ResourceTypePod]

	// Restart pod watcher
	err = wm.RestartWatcher(ResourceTypePod)
	if err != nil {
		t.Fatalf("RestartWatcher failed: %v", err)
	}

	// Give watcher time to restart
	time.Sleep(100 * time.Millisecond)

	// Should still be watching
	if !wm.IsWatching(ResourceTypePod) {
		t.Error("Expected to still be watching pods after restart")
	}

	// Resource version may have changed (or stayed the same with fake client)
	versionsAfter := wm.GetResourceVersions()
	versionAfter := versionsAfter[ResourceTypePod]
	t.Logf("Resource version before: %s, after: %s", initialVersion, versionAfter)

	// Stop
	wm.Stop()
}

func TestWatchManagerGetConnectionStates(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching
	resourceTypes := []ResourceType{ResourceTypePod, ResourceTypeService}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watchers time to initialize
	time.Sleep(200 * time.Millisecond)

	// Get connection states
	states := wm.GetConnectionStates()

	if len(states) != 2 {
		t.Errorf("Expected 2 connection states, got %d", len(states))
	}

	// Check that we have states for both resource types
	if _, exists := states[ResourceTypePod]; !exists {
		t.Error("Expected connection state for Pods")
	}
	if _, exists := states[ResourceTypeService]; !exists {
		t.Error("Expected connection state for Services")
	}

	// Log the states for debugging
	for rt, state := range states {
		t.Logf("%s: %s", rt, state)
	}

	// Stop
	wm.Stop()
}

func TestWatchManagerGetOverallConnectionState(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	// Before starting, should be disconnected
	if state := wm.GetOverallConnectionState(); state != StateDisconnected {
		t.Errorf("Expected StateDisconnected before start, got %s", state)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching
	resourceTypes := []ResourceType{ResourceTypePod}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watcher time to initialize
	time.Sleep(200 * time.Millisecond)

	// Get overall state
	overallState := wm.GetOverallConnectionState()
	t.Logf("Overall connection state: %s", overallState)

	// Should be one of the valid states (not Unknown)
	validStates := map[ConnectionState]bool{
		StateDisconnected:  true,
		StateConnecting:    true,
		StateConnected:     true,
		StateReconnecting:  true,
		StateError:         true,
	}

	if !validStates[overallState] {
		t.Errorf("Expected valid connection state, got %s", overallState)
	}

	// Stop
	wm.Stop()
}

func TestWatchManagerGetEventAndErrorChannels(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	eventChan := wm.GetEventChannel()
	errorChan := wm.GetErrorChannel()

	if eventChan == nil {
		t.Error("Expected non-nil event channel")
	}
	if errorChan == nil {
		t.Error("Expected non-nil error channel")
	}
}

func TestWatchManagerIsWatching(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Before starting
	if wm.IsWatching(ResourceTypePod) {
		t.Error("Expected not to be watching pods before start")
	}

	// Start watching pods only
	resourceTypes := []ResourceType{ResourceTypePod}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// After starting
	if !wm.IsWatching(ResourceTypePod) {
		t.Error("Expected to be watching pods after start")
	}
	if wm.IsWatching(ResourceTypeService) {
		t.Error("Expected not to be watching services")
	}

	// Stop
	wm.Stop()

	// After stopping
	if wm.IsWatching(ResourceTypePod) {
		t.Error("Expected not to be watching pods after stop")
	}
}

func TestWatchManagerUpdateNamespace(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching
	resourceTypes := []ResourceType{ResourceTypePod}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Update namespace
	err = wm.UpdateNamespace("kube-system")
	if err != nil {
		t.Fatalf("UpdateNamespace failed: %v", err)
	}

	// Give watchers time to restart
	time.Sleep(100 * time.Millisecond)

	// Should still be watching
	if !wm.IsWatching(ResourceTypePod) {
		t.Error("Expected to still be watching pods after namespace change")
	}

	// Check namespace was updated
	if client.GetNamespace() != "kube-system" {
		t.Errorf("Expected namespace 'kube-system', got '%s'", client.GetNamespace())
	}

	// Stop
	wm.Stop()
}

func TestWatchManagerGetResourceVersions(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching
	resourceTypes := []ResourceType{ResourceTypePod, ResourceTypeService}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Get resource versions
	versions := wm.GetResourceVersions()

	if len(versions) != 2 {
		t.Errorf("Expected 2 resource versions, got %d", len(versions))
	}

	// Check that we have versions for both resource types
	if _, exists := versions[ResourceTypePod]; !exists {
		t.Error("Expected resource version for Pods")
	}
	if _, exists := versions[ResourceTypeService]; !exists {
		t.Error("Expected resource version for Services")
	}

	// Log versions for debugging
	for rt, version := range versions {
		t.Logf("%s: version=%s", rt, version)
	}

	// Stop
	wm.Stop()
}

func TestWatchManagerRestartAll(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching multiple resource types
	resourceTypes := []ResourceType{ResourceTypePod, ResourceTypeService}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Restart all
	err = wm.RestartAll()
	if err != nil {
		t.Fatalf("RestartAll failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Should still be watching both
	if wm.GetWatcherCount() != 2 {
		t.Errorf("Expected 2 watchers after RestartAll, got %d", wm.GetWatcherCount())
	}

	if !wm.IsWatching(ResourceTypePod) {
		t.Error("Expected to still be watching pods after RestartAll")
	}
	if !wm.IsWatching(ResourceTypeService) {
		t.Error("Expected to still be watching services after RestartAll")
	}

	// Stop
	wm.Stop()
}

func TestWatchManagerConcurrentAccess(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	wm := NewWatchManager(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching
	resourceTypes := []ResourceType{ResourceTypePod}
	err := wm.Start(ctx, resourceTypes)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Concurrent access to read operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = wm.GetWatcherCount()
			_ = wm.GetConnectionStates()
			_ = wm.GetResourceVersions()
			_ = wm.IsWatching(ResourceTypePod)
			_ = wm.GetOverallConnectionState()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Stop
	wm.Stop()
}
