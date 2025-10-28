package components

import (
	"testing"

	"github.com/williajm/k8s-tui/internal/models"
)

func TestNewPodList(t *testing.T) {
	list := NewPodList()

	if list == nil {
		t.Fatal("NewPodList() returned nil")
	}

	if list.selectedIdx != 0 {
		t.Errorf("NewPodList().selectedIdx = %d, want 0", list.selectedIdx)
	}

	if list.viewportTop != 0 {
		t.Errorf("NewPodList().viewportTop = %d, want 0", list.viewportTop)
	}

	if len(list.pods) != 0 {
		t.Errorf("NewPodList().pods length = %d, want 0", len(list.pods))
	}
}

func TestPodList_SetPods(t *testing.T) {
	list := NewPodList()

	pods := []models.PodInfo{
		{Name: "pod1", Status: "Running"},
		{Name: "pod2", Status: "Running"},
		{Name: "pod3", Status: "Pending"},
	}

	list.SetPods(pods)

	if len(list.pods) != 3 {
		t.Errorf("SetPods() pods length = %d, want 3", len(list.pods))
	}

	if list.pods[0].Name != "pod1" {
		t.Errorf("SetPods() pods[0].Name = %s, want pod1", list.pods[0].Name)
	}
}

func TestPodList_Navigation(t *testing.T) {
	list := NewPodList()
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

	// Test Home
	list.Home()
	if list.selectedIdx != 0 {
		t.Errorf("After Home(), selectedIdx = %d, want 0", list.selectedIdx)
	}

	if list.viewportTop != 0 {
		t.Errorf("After Home(), viewportTop = %d, want 0", list.viewportTop)
	}
}

func TestPodList_GetSelected(t *testing.T) {
	list := NewPodList()

	// Test with no pods
	selected := list.GetSelected()
	if selected != nil {
		t.Error("GetSelected() with no pods should return nil")
	}

	// Test with pods
	pods := []models.PodInfo{
		{Name: "pod1", Status: "Running"},
		{Name: "pod2", Status: "Running"},
	}
	list.SetPods(pods)

	selected = list.GetSelected()
	if selected == nil {
		t.Fatal("GetSelected() returned nil")
	}

	if selected.Name != "pod1" {
		t.Errorf("GetSelected().Name = %s, want pod1", selected.Name)
	}

	// Move down and test again
	list.MoveDown()
	selected = list.GetSelected()
	if selected.Name != "pod2" {
		t.Errorf("After MoveDown(), GetSelected().Name = %s, want pod2", selected.Name)
	}
}

func TestPodList_SearchFilter(t *testing.T) {
	list := NewPodList()

	pods := []models.PodInfo{
		{Name: "nginx-pod", Namespace: "default", Status: "Running"},
		{Name: "redis-pod", Namespace: "default", Status: "Running"},
		{Name: "postgres-pod", Namespace: "database", Status: "Pending"},
	}
	list.SetPods(pods)

	// Test filter by name
	list.SetSearchFilter("nginx")
	filtered := list.getFilteredPods()
	if len(filtered) != 1 {
		t.Errorf("Filter 'nginx' returned %d pods, want 1", len(filtered))
	}
	if filtered[0].Name != "nginx-pod" {
		t.Errorf("Filtered pod name = %s, want nginx-pod", filtered[0].Name)
	}

	// Test filter by namespace
	list.SetSearchFilter("database")
	filtered = list.getFilteredPods()
	if len(filtered) != 1 {
		t.Errorf("Filter 'database' returned %d pods, want 1", len(filtered))
	}

	// Test filter by status
	list.SetSearchFilter("pending")
	filtered = list.getFilteredPods()
	if len(filtered) != 1 {
		t.Errorf("Filter 'pending' returned %d pods, want 1", len(filtered))
	}

	// Test no matches
	list.SetSearchFilter("nonexistent")
	filtered = list.getFilteredPods()
	if len(filtered) != 0 {
		t.Errorf("Filter 'nonexistent' returned %d pods, want 0", len(filtered))
	}

	// Test empty filter returns all
	list.SetSearchFilter("")
	filtered = list.getFilteredPods()
	if len(filtered) != 3 {
		t.Errorf("Empty filter returned %d pods, want 3", len(filtered))
	}
}

func TestPodList_PageNavigation(t *testing.T) {
	list := NewPodList()
	list.SetSize(80, 10) // Small height for testing pagination

	// Create many pods
	pods := make([]models.PodInfo, 50)
	for i := 0; i < 50; i++ {
		pods[i] = models.PodInfo{Name: "pod" + string(rune(i)), Status: "Running"}
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
}

func TestPodList_View_EmptyList(t *testing.T) {
	list := NewPodList()
	list.SetSize(80, 20)

	view := list.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}

func TestPodList_View_WithPods(t *testing.T) {
	list := NewPodList()
	list.SetSize(100, 25)

	pods := []models.PodInfo{
		{Name: "test-pod-1", Namespace: "default", Status: "Running", Ready: "1/1", Restarts: 0, Age: "5m"},
		{Name: "test-pod-2", Namespace: "default", Status: "Pending", Ready: "0/1", Restarts: 2, Age: "2m"},
		{Name: "test-pod-3", Namespace: "kube-system", Status: "Running", Ready: "1/1", Restarts: 0, Age: "10h"},
	}
	list.SetPods(pods)

	view := list.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}

func TestPodList_View_WithSearchFilter(t *testing.T) {
	list := NewPodList()
	list.SetSize(100, 25)

	pods := []models.PodInfo{
		{Name: "nginx-pod", Namespace: "default", Status: "Running"},
		{Name: "redis-pod", Namespace: "default", Status: "Running"},
		{Name: "postgres-pod", Namespace: "database", Status: "Pending"},
	}
	list.SetPods(pods)
	list.SetSearchFilter("nginx")

	view := list.View()
	if view == "" {
		t.Fatal("View() with filter returned empty string")
	}
}

func TestPodList_SetSize(t *testing.T) {
	list := NewPodList()

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
			list.SetSize(tt.width, tt.height)

			if list.width != tt.width {
				t.Errorf("SetSize() width = %d, want %d", list.width, tt.width)
			}

			if list.height != tt.height {
				t.Errorf("SetSize() height = %d, want %d", list.height, tt.height)
			}
		})
	}
}

func TestPodList_SetSearchFilter(t *testing.T) {
	list := NewPodList()

	tests := []struct {
		name   string
		filter string
	}{
		{
			name:   "simple filter",
			filter: "nginx",
		},
		{
			name:   "empty filter",
			filter: "",
		},
		{
			name:   "complex filter",
			filter: "kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list.SetSearchFilter(tt.filter)

			if list.searchFilter != tt.filter {
				t.Errorf("SetSearchFilter() searchFilter = %s, want %s", list.searchFilter, tt.filter)
			}
		})
	}
}
