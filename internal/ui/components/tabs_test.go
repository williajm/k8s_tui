package components

import (
	"testing"
)

func TestNewTabs(t *testing.T) {
	tabs := NewTabs()

	if tabs == nil {
		t.Fatal("NewTabs() returned nil")
	}

	if tabs.activeTab != 0 {
		t.Errorf("NewTabs().activeTab = %d, want 0", tabs.activeTab)
	}

	if tabs.width != 80 {
		t.Errorf("NewTabs().width = %d, want 80", tabs.width)
	}

	if len(tabs.tabs) != 4 {
		t.Errorf("NewTabs() has %d tabs, want 4", len(tabs.tabs))
	}

	expectedTitles := []string{"Pods", "Services", "Deployments", "StatefulSets"}
	for i, expectedTitle := range expectedTitles {
		if tabs.tabs[i].Title != expectedTitle {
			t.Errorf("NewTabs().tabs[%d].Title = %s, want %s", i, tabs.tabs[i].Title, expectedTitle)
		}
		if tabs.tabs[i].ID != i {
			t.Errorf("NewTabs().tabs[%d].ID = %d, want %d", i, tabs.tabs[i].ID, i)
		}
	}
}

func TestTabs_SetWidth(t *testing.T) {
	tabs := NewTabs()

	tabs.SetWidth(120)
	if tabs.width != 120 {
		t.Errorf("SetWidth(120) resulted in width = %d, want 120", tabs.width)
	}

	tabs.SetWidth(60)
	if tabs.width != 60 {
		t.Errorf("SetWidth(60) resulted in width = %d, want 60", tabs.width)
	}
}

func TestTabs_GetActiveTab(t *testing.T) {
	tabs := NewTabs()

	active := tabs.GetActiveTab()
	if active != 0 {
		t.Errorf("GetActiveTab() = %d, want 0", active)
	}

	tabs.activeTab = 2
	active = tabs.GetActiveTab()
	if active != 2 {
		t.Errorf("After setting activeTab to 2, GetActiveTab() = %d, want 2", active)
	}
}

func TestTabs_SetActiveTab(t *testing.T) {
	tests := []struct {
		name     string
		tabID    int
		expected int
	}{
		{
			name:     "valid tab 0",
			tabID:    0,
			expected: 0,
		},
		{
			name:     "valid tab 2",
			tabID:    2,
			expected: 2,
		},
		{
			name:     "valid tab 3",
			tabID:    3,
			expected: 3,
		},
		{
			name:     "invalid negative tab",
			tabID:    -1,
			expected: 0, // Should not change from initial value
		},
		{
			name:     "invalid tab beyond range",
			tabID:    10,
			expected: 0, // Should not change from initial value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabs := NewTabs()
			tabs.SetActiveTab(tt.tabID)

			if tabs.activeTab != tt.expected {
				t.Errorf("SetActiveTab(%d) resulted in activeTab = %d, want %d", tt.tabID, tabs.activeTab, tt.expected)
			}
		})
	}
}

func TestTabs_NextTab(t *testing.T) {
	tabs := NewTabs()

	// Start at 0, next should be 1
	tabs.NextTab()
	if tabs.activeTab != 1 {
		t.Errorf("After NextTab() from 0, activeTab = %d, want 1", tabs.activeTab)
	}

	// Next should be 2
	tabs.NextTab()
	if tabs.activeTab != 2 {
		t.Errorf("After NextTab() from 1, activeTab = %d, want 2", tabs.activeTab)
	}

	// Next should be 3
	tabs.NextTab()
	if tabs.activeTab != 3 {
		t.Errorf("After NextTab() from 2, activeTab = %d, want 3", tabs.activeTab)
	}

	// Next should wrap around to 0
	tabs.NextTab()
	if tabs.activeTab != 0 {
		t.Errorf("After NextTab() from 3, activeTab = %d, want 0 (wrap around)", tabs.activeTab)
	}
}

func TestTabs_PrevTab(t *testing.T) {
	tabs := NewTabs()

	// Start at 0, prev should wrap to 3
	tabs.PrevTab()
	if tabs.activeTab != 3 {
		t.Errorf("After PrevTab() from 0, activeTab = %d, want 3 (wrap around)", tabs.activeTab)
	}

	// Prev should be 2
	tabs.PrevTab()
	if tabs.activeTab != 2 {
		t.Errorf("After PrevTab() from 3, activeTab = %d, want 2", tabs.activeTab)
	}

	// Prev should be 1
	tabs.PrevTab()
	if tabs.activeTab != 1 {
		t.Errorf("After PrevTab() from 2, activeTab = %d, want 1", tabs.activeTab)
	}

	// Prev should be 0
	tabs.PrevTab()
	if tabs.activeTab != 0 {
		t.Errorf("After PrevTab() from 1, activeTab = %d, want 0", tabs.activeTab)
	}
}

func TestTabs_View(t *testing.T) {
	tabs := NewTabs()

	// Just verify that View() returns a non-empty string
	view := tabs.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Test with different active tabs
	for i := 0; i < 4; i++ {
		tabs.SetActiveTab(i)
		view = tabs.View()
		if view == "" {
			t.Errorf("View() returned empty string with activeTab = %d", i)
		}
	}

	// Test with different widths
	tabs.SetWidth(100)
	view = tabs.View()
	if view == "" {
		t.Error("View() returned empty string with width = 100")
	}
}

func TestTabs_NavigationCycle(t *testing.T) {
	tabs := NewTabs()

	// Test full forward cycle
	for i := 0; i < 4; i++ {
		if tabs.GetActiveTab() != i {
			t.Errorf("Forward cycle iteration %d: activeTab = %d, want %d", i, tabs.GetActiveTab(), i)
		}
		tabs.NextTab()
	}

	// Should be back at 0
	if tabs.GetActiveTab() != 0 {
		t.Errorf("After full forward cycle, activeTab = %d, want 0", tabs.GetActiveTab())
	}

	// Test full backward cycle
	for i := 0; i < 4; i++ {
		expectedTab := (4 - i) % 4
		if tabs.GetActiveTab() != expectedTab {
			t.Errorf("Backward cycle iteration %d: activeTab = %d, want %d", i, tabs.GetActiveTab(), expectedTab)
		}
		tabs.PrevTab()
	}

	// Should be back at 0
	if tabs.GetActiveTab() != 0 {
		t.Errorf("After full backward cycle, activeTab = %d, want 0", tabs.GetActiveTab())
	}
}
