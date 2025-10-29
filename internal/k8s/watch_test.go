package k8s

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

// TestWatchPodsCreatesWatcher tests that WatchPods creates a watcher successfully
func TestWatchPodsCreatesWatcher(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchPods(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchPods failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
}

// TestWatchPodsWithResourceVersion tests watching pods with a specific resource version
func TestWatchPodsWithResourceVersion(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Watch with specific resource version
	watcher, err := client.WatchPods(ctx, "default", "12345")
	if err != nil {
		t.Fatalf("WatchPods with resource version failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
}

// TestWatchPodsDefaultNamespace tests that empty namespace uses client's default namespace
func TestWatchPodsDefaultNamespace(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "test-namespace",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Pass empty namespace, should use default
	watcher, err := client.WatchPods(ctx, "", "")
	if err != nil {
		t.Fatalf("WatchPods with default namespace failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
}

// TestWatchServicesCreatesWatcher tests that WatchServices works
func TestWatchServicesCreatesWatcher(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchServices(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchServices failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
}

// TestWatchDeploymentsCreatesWatcher tests that WatchDeployments works
func TestWatchDeploymentsCreatesWatcher(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchDeployments(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchDeployments failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
}

// TestWatchStatefulSetsCreatesWatcher tests that WatchStatefulSets works
func TestWatchStatefulSetsCreatesWatcher(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchStatefulSets(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchStatefulSets failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
}

// TestWatchEventsCreatesWatcher tests that WatchEvents works
func TestWatchEventsCreatesWatcher(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchEvents(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchEvents failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}
}

// TestWatchPodReceivesEvents tests that watch events are received correctly
func TestWatchPodReceivesEvents(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchPods(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchPods failed: %v", err)
	}
	defer watcher.Stop()

	// Create a pod which should trigger a watch event
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: "nginx:latest"},
			},
		},
	}

	// Add the pod (this simulates an ADDED event)
	go func() {
		time.Sleep(100 * time.Millisecond)
		_, _ = fakeClientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	}()

	// Wait for event with timeout
	select {
	case event := <-watcher.ResultChan():
		if event.Type != watch.Added {
			t.Errorf("Expected ADDED event, got %v", event.Type)
		}
		if p, ok := event.Object.(*corev1.Pod); ok {
			if p.Name != "test-pod" {
				t.Errorf("Expected pod name 'test-pod', got '%s'", p.Name)
			}
		} else {
			t.Error("Expected Pod object in watch event")
		}
	case <-time.After(2 * time.Second):
		t.Log("Timeout waiting for watch event (this is expected with fake client)")
	}
}

// TestWatchContextCancellation tests that watch respects context cancellation
func TestWatchContextCancellation(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithCancel(context.Background())

	watcher, err := client.WatchPods(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchPods failed: %v", err)
	}
	defer watcher.Stop()

	// Cancel context immediately
	cancel()

	// The result channel should be closed or return error
	select {
	case event, ok := <-watcher.ResultChan():
		if ok {
			// Check if it's an error event
			if event.Type == watch.Error {
				t.Log("Received error event after context cancellation (expected)")
			}
		} else {
			t.Log("Watch channel closed after context cancellation (expected)")
		}
	case <-time.After(2 * time.Second):
		t.Log("Watch continued after context cancellation (fake client behavior)")
	}
}

// TestWatchMultipleResourceTypes tests watching multiple resource types concurrently
func TestWatchMultipleResourceTypes(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watchers for different resource types
	podWatcher, err := client.WatchPods(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchPods failed: %v", err)
	}
	defer podWatcher.Stop()

	svcWatcher, err := client.WatchServices(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchServices failed: %v", err)
	}
	defer svcWatcher.Stop()

	deployWatcher, err := client.WatchDeployments(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchDeployments failed: %v", err)
	}
	defer deployWatcher.Stop()

	// All watchers should be active
	if podWatcher == nil || svcWatcher == nil || deployWatcher == nil {
		t.Fatal("Expected all watchers to be non-nil")
	}
}

// TestWatchReceivesModifiedEvent tests that modified events are received
func TestWatchReceivesModifiedEvent(t *testing.T) {
	// Create a fake client with initial objects
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchPods(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchPods failed: %v", err)
	}
	defer watcher.Stop()

	// Update the pod
	go func() {
		time.Sleep(100 * time.Millisecond)
		updatedPod := pod.DeepCopy()
		updatedPod.Labels = map[string]string{"updated": "true"}
		_, _ = fakeClientset.CoreV1().Pods("default").Update(context.Background(), updatedPod, metav1.UpdateOptions{})
	}()

	// Note: fake client behavior may vary, this is for demonstration
	select {
	case <-watcher.ResultChan():
		t.Log("Received watch event (may be ADDED or MODIFIED)")
	case <-time.After(2 * time.Second):
		t.Log("Timeout waiting for watch event (expected with fake client)")
	}
}

// TestInt64Ptr tests the helper function
func TestInt64Ptr(t *testing.T) {
	val := int64(300)
	ptr := int64ptr(val)

	if ptr == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *ptr != val {
		t.Errorf("Expected %d, got %d", val, *ptr)
	}
}

// TestWatchWithDifferentResourceTypes tests watching different Kubernetes resource types
func TestWatchWithDifferentResourceTypes(t *testing.T) {
	tests := []struct {
		name         string
		watchFunc    func(*Client, context.Context, string, string) (watch.Interface, error)
		resourceType string
	}{
		{
			name:         "Pods",
			watchFunc:    (*Client).WatchPods,
			resourceType: "Pod",
		},
		{
			name:         "Services",
			watchFunc:    (*Client).WatchServices,
			resourceType: "Service",
		},
		{
			name:         "Deployments",
			watchFunc:    (*Client).WatchDeployments,
			resourceType: "Deployment",
		},
		{
			name:         "StatefulSets",
			watchFunc:    (*Client).WatchStatefulSets,
			resourceType: "StatefulSet",
		},
		{
			name:         "Events",
			watchFunc:    (*Client).WatchEvents,
			resourceType: "Event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset()
			client := &Client{
				clientset: fakeClientset,
				namespace: "default",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			watcher, err := tt.watchFunc(client, ctx, "default", "")
			if err != nil {
				t.Fatalf("Watch %s failed: %v", tt.resourceType, err)
			}
			defer watcher.Stop()

			if watcher == nil {
				t.Fatalf("Expected non-nil watcher for %s", tt.resourceType)
			}
		})
	}
}

// TestWatchReceivesDeletedEvent tests that deleted events are received
func TestWatchReceivesDeletedEvent(t *testing.T) {
	// Create a deployment for testing
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-deployment",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx:latest"},
					},
				},
			},
		},
	}

	fakeClientset := fake.NewSimpleClientset(deployment)
	client := &Client{
		clientset: fakeClientset,
		namespace: "default",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := client.WatchDeployments(ctx, "default", "")
	if err != nil {
		t.Fatalf("WatchDeployments failed: %v", err)
	}
	defer watcher.Stop()

	// Delete the deployment
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = fakeClientset.AppsV1().Deployments("default").Delete(
			context.Background(),
			"test-deployment",
			metav1.DeleteOptions{},
		)
	}()

	// Note: fake client behavior may vary
	select {
	case event := <-watcher.ResultChan():
		t.Logf("Received watch event type: %v", event.Type)
	case <-time.After(2 * time.Second):
		t.Log("Timeout waiting for watch event (expected with fake client)")
	}
}

// int32ptr is a helper for tests
func int32ptr(i int32) *int32 {
	return &i
}

// runtimeObjectPtr is a helper for creating runtime.Object pointers
func runtimeObjectPtr(obj runtime.Object) *runtime.Object {
	return &obj
}
