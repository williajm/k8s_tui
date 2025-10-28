package components

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/williajm/k8s-tui/internal/models"
)

func TestNewResourceList(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
	}{
		{
			name:         "Pod resource type",
			resourceType: ResourceTypePod,
		},
		{
			name:         "Service resource type",
			resourceType: ResourceTypeService,
		},
		{
			name:         "Deployment resource type",
			resourceType: ResourceTypeDeployment,
		},
		{
			name:         "StatefulSet resource type",
			resourceType: ResourceTypeStatefulSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewResourceList(tt.resourceType)

			if list == nil {
				t.Fatal("NewResourceList() returned nil")
			}

			if list.resourceType != tt.resourceType {
				t.Errorf("NewResourceList().resourceType = %v, want %v", list.resourceType, tt.resourceType)
			}

			if list.selectedIdx != 0 {
				t.Errorf("NewResourceList().selectedIdx = %d, want 0", list.selectedIdx)
			}

			if list.viewportTop != 0 {
				t.Errorf("NewResourceList().viewportTop = %d, want 0", list.viewportTop)
			}

			if list.width != 80 {
				t.Errorf("NewResourceList().width = %d, want 80", list.width)
			}

			if list.height != 20 {
				t.Errorf("NewResourceList().height = %d, want 20", list.height)
			}
		})
	}
}

func TestResourceList_SetResourceType(t *testing.T) {
	list := NewResourceList(ResourceTypePod)
	list.selectedIdx = 5
	list.viewportTop = 3

	list.SetResourceType(ResourceTypeService)

	if list.resourceType != ResourceTypeService {
		t.Errorf("SetResourceType() resourceType = %v, want %v", list.resourceType, ResourceTypeService)
	}

	if list.selectedIdx != 0 {
		t.Errorf("SetResourceType() should reset selectedIdx to 0, got %d", list.selectedIdx)
	}

	if list.viewportTop != 0 {
		t.Errorf("SetResourceType() should reset viewportTop to 0, got %d", list.viewportTop)
	}
}

func TestResourceList_SetPods(t *testing.T) {
	list := NewResourceList(ResourceTypePod)

	pods := []models.PodInfo{
		{Name: "pod1", Namespace: "default", Status: "Running"},
		{Name: "pod2", Namespace: "default", Status: "Running"},
		{Name: "pod3", Namespace: "default", Status: "Pending"},
	}

	list.SetPods(pods)

	if len(list.pods) != 3 {
		t.Errorf("SetPods() resulted in %d pods, want 3", len(list.pods))
	}

	// Test that selectedIdx is reset when out of bounds
	list.selectedIdx = 10
	list.SetPods([]models.PodInfo{{Name: "pod1"}})
	if list.selectedIdx != 0 {
		t.Errorf("SetPods() with out of bounds selectedIdx, got %d, want 0", list.selectedIdx)
	}
}

func TestResourceList_SetServices(t *testing.T) {
	list := NewResourceList(ResourceTypeService)

	services := []models.ServiceInfo{
		{Name: "service1", Namespace: "default", Type: "ClusterIP"},
		{Name: "service2", Namespace: "default", Type: "NodePort"},
	}

	list.SetServices(services)

	if len(list.services) != 2 {
		t.Errorf("SetServices() resulted in %d services, want 2", len(list.services))
	}

	// Test that selectedIdx is reset when out of bounds
	list.selectedIdx = 10
	list.SetServices([]models.ServiceInfo{{Name: "service1"}})
	if list.selectedIdx != 0 {
		t.Errorf("SetServices() with out of bounds selectedIdx, got %d, want 0", list.selectedIdx)
	}
}

func TestResourceList_SetDeployments(t *testing.T) {
	list := NewResourceList(ResourceTypeDeployment)

	deployments := []models.DeploymentInfo{
		{Name: "deploy1", Namespace: "default", Replicas: 3},
		{Name: "deploy2", Namespace: "default", Replicas: 2},
	}

	list.SetDeployments(deployments)

	if len(list.deployments) != 2 {
		t.Errorf("SetDeployments() resulted in %d deployments, want 2", len(list.deployments))
	}

	// Test that selectedIdx is reset when out of bounds
	list.selectedIdx = 10
	list.SetDeployments([]models.DeploymentInfo{{Name: "deploy1"}})
	if list.selectedIdx != 0 {
		t.Errorf("SetDeployments() with out of bounds selectedIdx, got %d, want 0", list.selectedIdx)
	}
}

func TestResourceList_SetStatefulSets(t *testing.T) {
	list := NewResourceList(ResourceTypeStatefulSet)

	statefulSets := []models.StatefulSetInfo{
		{Name: "sts1", Namespace: "default", Replicas: 3},
	}

	list.SetStatefulSets(statefulSets)

	if len(list.statefulSets) != 1 {
		t.Errorf("SetStatefulSets() resulted in %d statefulsets, want 1", len(list.statefulSets))
	}

	// Test that selectedIdx is reset when out of bounds
	list.selectedIdx = 10
	list.SetStatefulSets([]models.StatefulSetInfo{{Name: "sts1"}})
	if list.selectedIdx != 0 {
		t.Errorf("SetStatefulSets() with out of bounds selectedIdx, got %d, want 0", list.selectedIdx)
	}
}

func TestResourceList_SetSize(t *testing.T) {
	list := NewResourceList(ResourceTypePod)

	list.SetSize(120, 40)

	if list.width != 120 {
		t.Errorf("SetSize(120, 40) width = %d, want 120", list.width)
	}

	if list.height != 40 {
		t.Errorf("SetSize(120, 40) height = %d, want 40", list.height)
	}
}

func TestResourceList_SetSearchFilter(t *testing.T) {
	list := NewResourceList(ResourceTypePod)

	list.SetSearchFilter("test-filter")

	if list.searchFilter != "test-filter" {
		t.Errorf("SetSearchFilter() searchFilter = %s, want test-filter", list.searchFilter)
	}
}

func TestResourceList_Navigation(t *testing.T) {
	list := NewResourceList(ResourceTypePod)
	list.SetSize(80, 20)

	pods := []models.PodInfo{
		{Name: "pod1", Status: "Running"},
		{Name: "pod2", Status: "Running"},
		{Name: "pod3", Status: "Pending"},
		{Name: "pod4", Status: "Running"},
		{Name: "pod5", Status: "Failed"},
	}
	list.SetPods(pods)

	// Test MoveDown
	list.MoveDown()
	if list.selectedIdx != 1 {
		t.Errorf("After MoveDown(), selectedIdx = %d, want 1", list.selectedIdx)
	}

	// Test MoveUp
	list.MoveUp()
	if list.selectedIdx != 0 {
		t.Errorf("After MoveUp(), selectedIdx = %d, want 0", list.selectedIdx)
	}

	// Test MoveUp at boundary
	list.MoveUp()
	if list.selectedIdx != 0 {
		t.Errorf("After MoveUp() at top, selectedIdx = %d, want 0", list.selectedIdx)
	}

	// Test End
	list.End()
	if list.selectedIdx != 4 {
		t.Errorf("After End(), selectedIdx = %d, want 4", list.selectedIdx)
	}

	// Test MoveDown at boundary
	list.MoveDown()
	if list.selectedIdx != 4 {
		t.Errorf("After MoveDown() at bottom, selectedIdx = %d, want 4", list.selectedIdx)
	}

	// Test Home
	list.Home()
	if list.selectedIdx != 0 {
		t.Errorf("After Home(), selectedIdx = %d, want 0", list.selectedIdx)
	}

	if list.viewportTop != 0 {
		t.Errorf("After Home(), viewportTop = %d, want 0", list.viewportTop)
	}
}

func TestResourceList_GetSelectedPod(t *testing.T) {
	list := NewResourceList(ResourceTypePod)

	// No pods
	pod := list.GetSelectedPod()
	if pod != nil {
		t.Error("GetSelectedPod() with no pods should return nil")
	}

	// With pods
	pods := []models.PodInfo{
		{Name: "pod1", Status: "Running"},
		{Name: "pod2", Status: "Running"},
	}
	list.SetPods(pods)

	pod = list.GetSelectedPod()
	if pod == nil {
		t.Fatal("GetSelectedPod() returned nil")
	}

	if pod.Name != "pod1" {
		t.Errorf("GetSelectedPod().Name = %s, want pod1", pod.Name)
	}

	// Move down and test
	list.MoveDown()
	pod = list.GetSelectedPod()
	if pod.Name != "pod2" {
		t.Errorf("After MoveDown(), GetSelectedPod().Name = %s, want pod2", pod.Name)
	}

	// Wrong resource type
	list.SetResourceType(ResourceTypeService)
	pod = list.GetSelectedPod()
	if pod != nil {
		t.Error("GetSelectedPod() with wrong resource type should return nil")
	}
}

func TestResourceList_GetSelectedService(t *testing.T) {
	list := NewResourceList(ResourceTypeService)

	services := []models.ServiceInfo{
		{Name: "service1", Type: "ClusterIP"},
		{Name: "service2", Type: "NodePort"},
	}
	list.SetServices(services)

	service := list.GetSelectedService()
	if service == nil {
		t.Fatal("GetSelectedService() returned nil")
	}

	if service.Name != "service1" {
		t.Errorf("GetSelectedService().Name = %s, want service1", service.Name)
	}
}

func TestResourceList_GetSelectedDeployment(t *testing.T) {
	list := NewResourceList(ResourceTypeDeployment)

	deployments := []models.DeploymentInfo{
		{Name: "deploy1", Replicas: 3},
		{Name: "deploy2", Replicas: 2},
	}
	list.SetDeployments(deployments)

	deployment := list.GetSelectedDeployment()
	if deployment == nil {
		t.Fatal("GetSelectedDeployment() returned nil")
	}

	if deployment.Name != "deploy1" {
		t.Errorf("GetSelectedDeployment().Name = %s, want deploy1", deployment.Name)
	}
}

func TestResourceList_GetSelectedStatefulSet(t *testing.T) {
	list := NewResourceList(ResourceTypeStatefulSet)

	statefulSets := []models.StatefulSetInfo{
		{Name: "sts1", Replicas: 3},
	}
	list.SetStatefulSets(statefulSets)

	sts := list.GetSelectedStatefulSet()
	if sts == nil {
		t.Fatal("GetSelectedStatefulSet() returned nil")
	}

	if sts.Name != "sts1" {
		t.Errorf("GetSelectedStatefulSet().Name = %s, want sts1", sts.Name)
	}
}

func TestResourceList_PageNavigation(t *testing.T) {
	list := NewResourceList(ResourceTypePod)
	list.SetSize(80, 10) // Small height for testing pagination

	// Create many pods
	pods := make([]models.PodInfo, 50)
	for i := 0; i < 50; i++ {
		pods[i] = models.PodInfo{Name: "pod", Status: "Running"}
	}
	list.SetPods(pods)

	// Test PageDown
	list.PageDown()
	expectedIdx := 7 // height - 3 (for header and borders)
	if list.selectedIdx != expectedIdx {
		t.Errorf("After PageDown(), selectedIdx = %d, want %d", list.selectedIdx, expectedIdx)
	}

	// Test PageUp
	list.PageUp()
	if list.selectedIdx != 0 {
		t.Errorf("After PageUp(), selectedIdx = %d, want 0", list.selectedIdx)
	}

	// Test PageDown to end
	list.End()
	lastIdx := list.selectedIdx
	list.PageDown() // Should stay at end
	if list.selectedIdx != lastIdx {
		t.Errorf("PageDown() at end, selectedIdx changed from %d to %d", lastIdx, list.selectedIdx)
	}
}

func TestResourceList_SearchFilter(t *testing.T) {
	list := NewResourceList(ResourceTypePod)

	pods := []models.PodInfo{
		{Name: "nginx-pod", Namespace: "default", Status: "Running"},
		{Name: "redis-pod", Namespace: "default", Status: "Running"},
		{Name: "postgres-pod", Namespace: "database", Status: "Pending"},
	}
	list.SetPods(pods)

	// Set filter
	list.SetSearchFilter("nginx")
	if list.searchFilter != "nginx" {
		t.Errorf("SetSearchFilter() searchFilter = %s, want nginx", list.searchFilter)
	}

	// Test with different resource types
	services := []models.ServiceInfo{
		{Name: "nginx-service", Namespace: "default"},
		{Name: "redis-service", Namespace: "cache"},
	}
	list.SetResourceType(ResourceTypeService)
	list.SetServices(services)
	list.SetSearchFilter("redis")
	if list.searchFilter != "redis" {
		t.Errorf("SetSearchFilter() for services, searchFilter = %s, want redis", list.searchFilter)
	}
}

func TestResourceList_EmptyLists(t *testing.T) {
	list := NewResourceList(ResourceTypePod)

	// Test navigation with empty list - these operations should not crash
	// but selectedIdx may end up at -1 (which is getItemCount() - 1 when empty)
	list.MoveDown()
	// MoveDown on empty list stays at 0 (doesn't go past maxIdx which is -1)
	if list.selectedIdx != 0 {
		t.Errorf("MoveDown() with empty list, selectedIdx = %d, want 0", list.selectedIdx)
	}

	list.MoveUp()
	if list.selectedIdx != 0 {
		t.Errorf("MoveUp() with empty list, selectedIdx = %d, want 0", list.selectedIdx)
	}

	// PageDown with empty list will set to maxIdx which is -1 for empty list
	list.PageDown()
	if list.selectedIdx != -1 {
		t.Errorf("PageDown() with empty list, selectedIdx = %d, want -1", list.selectedIdx)
	}

	// Reset to 0 and test PageUp
	list.selectedIdx = 0
	list.PageUp()
	if list.selectedIdx != 0 {
		t.Errorf("PageUp() with empty list, selectedIdx = %d, want 0", list.selectedIdx)
	}

	// Test GetSelected* methods with empty lists
	if list.GetSelectedPod() != nil {
		t.Error("GetSelectedPod() with empty list should return nil")
	}

	list.SetResourceType(ResourceTypeService)
	if list.GetSelectedService() != nil {
		t.Error("GetSelectedService() with empty list should return nil")
	}

	list.SetResourceType(ResourceTypeDeployment)
	if list.GetSelectedDeployment() != nil {
		t.Error("GetSelectedDeployment() with empty list should return nil")
	}

	list.SetResourceType(ResourceTypeStatefulSet)
	if list.GetSelectedStatefulSet() != nil {
		t.Error("GetSelectedStatefulSet() with empty list should return nil")
	}
}

func TestResourceList_MultipleResourceTypes(t *testing.T) {
	list := NewResourceList(ResourceTypePod)

	// Add pods
	pods := []models.PodInfo{
		{Name: "pod1", Status: "Running"},
		{Name: "pod2", Status: "Running"},
	}
	list.SetPods(pods)

	// Add services
	services := []models.ServiceInfo{
		{Name: "service1", Type: "ClusterIP"},
	}
	list.SetServices(services)

	// Verify pod is selected initially
	pod := list.GetSelectedPod()
	if pod == nil || pod.Name != "pod1" {
		t.Error("Initial resource type should be Pod")
	}

	// Switch to services
	list.SetResourceType(ResourceTypeService)
	service := list.GetSelectedService()
	if service == nil || service.Name != "service1" {
		t.Error("After switching to Service, should select service1")
	}

	// Switch back to pods
	list.SetResourceType(ResourceTypePod)
	pod = list.GetSelectedPod()
	if pod == nil || pod.Name != "pod1" {
		t.Error("After switching back to Pod, should select pod1")
	}
}

func TestResourceList_View_EmptyList(t *testing.T) {
	list := NewResourceList(ResourceTypePod)
	list.SetSize(80, 20)

	view := list.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}

	// Should show "No resources found" message
	// We can't check exact content due to styling, but view should not be empty
}

func TestResourceList_View_WithPods(t *testing.T) {
	list := NewResourceList(ResourceTypePod)
	list.SetSize(100, 25)

	pods := []models.PodInfo{
		{Name: "test-pod-1", Namespace: "default", Status: "Running", Ready: "1/1", Restarts: 0, Age: "5m"},
		{Name: "test-pod-2", Namespace: "default", Status: "Pending", Ready: "0/1", Restarts: 2, Age: "2m"},
	}
	list.SetPods(pods)

	view := list.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}

	// View should contain some content (we can't test exact rendering due to styling)
}

func TestResourceList_View_WithServices(t *testing.T) {
	list := NewResourceList(ResourceTypeService)
	list.SetSize(120, 30)

	services := []models.ServiceInfo{
		{Name: "nginx-svc", Namespace: "default", Type: "ClusterIP", ClusterIP: "10.0.1.1", ExternalIP: "<none>", Ports: "80/TCP", Age: "1h"},
	}
	list.SetServices(services)

	view := list.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}

func TestResourceList_View_WithDeployments(t *testing.T) {
	list := NewResourceList(ResourceTypeDeployment)
	list.SetSize(120, 30)

	deployments := []models.DeploymentInfo{
		{Name: "nginx-deploy", Namespace: "default", Replicas: 3, Ready: "3/3", UpToDate: 3, Available: 3, Age: "7d"},
	}
	list.SetDeployments(deployments)

	view := list.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}

func TestResourceList_View_WithStatefulSets(t *testing.T) {
	list := NewResourceList(ResourceTypeStatefulSet)
	list.SetSize(120, 30)

	// Create a proper StatefulSet with the required fields
	replicas := int32(3)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "redis-sts",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas: 3,
		},
	}

	statefulSets := []models.StatefulSetInfo{
		models.NewStatefulSetInfo(sts),
	}
	list.SetStatefulSets(statefulSets)

	view := list.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}

func TestResourceList_View_AllResourceTypes(t *testing.T) {
	resourceTypes := []ResourceType{
		ResourceTypePod,
		ResourceTypeService,
		ResourceTypeDeployment,
		ResourceTypeStatefulSet,
	}

	resourceNames := []string{"Pod", "Service", "Deployment", "StatefulSet"}
	for i, rt := range resourceTypes {
		t.Run(resourceNames[i], func(t *testing.T) {
			list := NewResourceList(rt)
			list.SetSize(120, 30)

			// Add sample data for each type
			switch rt {
			case ResourceTypePod:
				list.SetPods([]models.PodInfo{{Name: "test-pod", Status: "Running"}})
			case ResourceTypeService:
				list.SetServices([]models.ServiceInfo{{Name: "test-svc", Type: "ClusterIP"}})
			case ResourceTypeDeployment:
				list.SetDeployments([]models.DeploymentInfo{{Name: "test-deploy", Replicas: 1}})
			case ResourceTypeStatefulSet:
				replicas := int32(1)
				sts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "test-sts",
						Namespace:         "default",
						CreationTimestamp: metav1.Now(),
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas,
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				}
				list.SetStatefulSets([]models.StatefulSetInfo{models.NewStatefulSetInfo(sts)})
			}

			view := list.View()
			if view == "" {
				t.Errorf("View() for %s returned empty string", resourceNames[i])
			}
		})
	}
}
