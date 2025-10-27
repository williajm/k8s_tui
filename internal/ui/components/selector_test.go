package components

import (
	"testing"
)

func TestNewSelector(t *testing.T) {
	selector := NewSelector("Test Selector")

	if selector == nil {
		t.Fatal("NewSelector() returned nil")
	}

	if selector.title != "Test Selector" {
		t.Errorf("NewSelector().title = %s, want Test Selector", selector.title)
	}

	if selector.selectedIdx != 0 {
		t.Errorf("NewSelector().selectedIdx = %d, want 0", selector.selectedIdx)
	}

	if selector.width != 40 {
		t.Errorf("NewSelector().width = %d, want 40", selector.width)
	}

	if selector.height != 15 {
		t.Errorf("NewSelector().height = %d, want 15", selector.height)
	}

	if selector.visible {
		t.Error("NewSelector().visible = true, want false")
	}

	if len(selector.options) != 0 {
		t.Errorf("NewSelector().options length = %d, want 0", len(selector.options))
	}
}

func TestSelector_SetOptions(t *testing.T) {
	selector := NewSelector("Test")

	options := []string{"option1", "option2", "option3"}
	selector.SetOptions(options)

	if len(selector.options) != 3 {
		t.Errorf("SetOptions() resulted in %d options, want 3", len(selector.options))
	}

	for i, option := range options {
		if selector.options[i] != option {
			t.Errorf("SetOptions() options[%d] = %s, want %s", i, selector.options[i], option)
		}
	}

	// Test that selectedIdx is reset when options change and index is out of bounds
	selector.selectedIdx = 5
	selector.SetOptions([]string{"a", "b"})
	if selector.selectedIdx != 0 {
		t.Errorf("SetOptions() with shorter list, selectedIdx = %d, want 0", selector.selectedIdx)
	}
}

func TestSelector_SetSize(t *testing.T) {
	selector := NewSelector("Test")

	selector.SetSize(80, 25)
	if selector.width != 80 {
		t.Errorf("SetSize(80, 25) width = %d, want 80", selector.width)
	}
	if selector.height != 25 {
		t.Errorf("SetSize(80, 25) height = %d, want 25", selector.height)
	}
}

func TestSelector_ShowHide(t *testing.T) {
	selector := NewSelector("Test")

	// Initially not visible
	if selector.IsVisible() {
		t.Error("New selector should not be visible")
	}

	// Show
	selector.Show()
	if !selector.IsVisible() {
		t.Error("After Show(), selector should be visible")
	}

	// Hide
	selector.Hide()
	if selector.IsVisible() {
		t.Error("After Hide(), selector should not be visible")
	}
}

func TestSelector_MoveUp(t *testing.T) {
	selector := NewSelector("Test")
	selector.SetOptions([]string{"a", "b", "c", "d"})

	// Start at 0, move up should stay at 0
	selector.MoveUp()
	if selector.selectedIdx != 0 {
		t.Errorf("MoveUp() from 0, selectedIdx = %d, want 0", selector.selectedIdx)
	}

	// Move to index 2, then move up
	selector.selectedIdx = 2
	selector.MoveUp()
	if selector.selectedIdx != 1 {
		t.Errorf("MoveUp() from 2, selectedIdx = %d, want 1", selector.selectedIdx)
	}

	// Move up again
	selector.MoveUp()
	if selector.selectedIdx != 0 {
		t.Errorf("MoveUp() from 1, selectedIdx = %d, want 0", selector.selectedIdx)
	}
}

func TestSelector_MoveDown(t *testing.T) {
	selector := NewSelector("Test")
	selector.SetOptions([]string{"a", "b", "c", "d"})

	// Start at 0, move down
	selector.MoveDown()
	if selector.selectedIdx != 1 {
		t.Errorf("MoveDown() from 0, selectedIdx = %d, want 1", selector.selectedIdx)
	}

	// Move down again
	selector.MoveDown()
	if selector.selectedIdx != 2 {
		t.Errorf("MoveDown() from 1, selectedIdx = %d, want 2", selector.selectedIdx)
	}

	// Move to last item
	selector.selectedIdx = 3
	selector.MoveDown()
	if selector.selectedIdx != 3 {
		t.Errorf("MoveDown() from last item, selectedIdx = %d, want 3", selector.selectedIdx)
	}
}

func TestSelector_GetSelected(t *testing.T) {
	selector := NewSelector("Test")

	// No options
	selected := selector.GetSelected()
	if selected != "" {
		t.Errorf("GetSelected() with no options = %s, want empty string", selected)
	}

	// With options
	options := []string{"default", "kube-system", "production"}
	selector.SetOptions(options)

	selected = selector.GetSelected()
	if selected != "default" {
		t.Errorf("GetSelected() at index 0 = %s, want default", selected)
	}

	selector.MoveDown()
	selected = selector.GetSelected()
	if selected != "kube-system" {
		t.Errorf("GetSelected() at index 1 = %s, want kube-system", selected)
	}

	selector.MoveDown()
	selected = selector.GetSelected()
	if selected != "production" {
		t.Errorf("GetSelected() at index 2 = %s, want production", selected)
	}
}

func TestSelector_View(t *testing.T) {
	selector := NewSelector("Select Namespace")

	// Not visible, should return empty
	view := selector.View()
	if view != "" {
		t.Error("View() when not visible should return empty string")
	}

	// Show but no options, should return empty
	selector.Show()
	view = selector.View()
	if view != "" {
		t.Error("View() with no options should return empty string")
	}

	// With options and visible
	selector.SetOptions([]string{"default", "kube-system", "production"})
	view = selector.View()
	if view == "" {
		t.Error("View() with options and visible should return non-empty string")
	}

	// Hide again
	selector.Hide()
	view = selector.View()
	if view != "" {
		t.Error("View() when hidden should return empty string")
	}
}

func TestSelector_NavigationCycle(t *testing.T) {
	selector := NewSelector("Test")
	selector.SetOptions([]string{"a", "b", "c"})

	// Navigate down through all options
	expected := []string{"a", "b", "c", "c"} // Last one stays at "c"
	for i, want := range expected {
		got := selector.GetSelected()
		if got != want {
			t.Errorf("Navigation step %d: GetSelected() = %s, want %s", i, got, want)
		}
		selector.MoveDown()
	}

	// Navigate up through all options
	selector.selectedIdx = 2                // Start at "c"
	expected = []string{"c", "b", "a", "a"} // First one stays at "a"
	for i, want := range expected {
		got := selector.GetSelected()
		if got != want {
			t.Errorf("Reverse navigation step %d: GetSelected() = %s, want %s", i, got, want)
		}
		selector.MoveUp()
	}
}

func TestSelector_BoundaryConditions(t *testing.T) {
	selector := NewSelector("Test")

	// Empty options
	selector.SetOptions([]string{})
	selector.MoveUp()
	selector.MoveDown()
	selected := selector.GetSelected()
	if selected != "" {
		t.Errorf("GetSelected() with empty options = %s, want empty string", selected)
	}

	// Single option
	selector.SetOptions([]string{"only-one"})
	selector.MoveUp()
	if selector.selectedIdx != 0 {
		t.Errorf("MoveUp() with single option, selectedIdx = %d, want 0", selector.selectedIdx)
	}
	selector.MoveDown()
	if selector.selectedIdx != 0 {
		t.Errorf("MoveDown() with single option, selectedIdx = %d, want 0", selector.selectedIdx)
	}
	selected = selector.GetSelected()
	if selected != "only-one" {
		t.Errorf("GetSelected() with single option = %s, want only-one", selected)
	}
}

func TestSelector_OptionsUpdate(t *testing.T) {
	selector := NewSelector("Test")

	// Set initial options and navigate
	selector.SetOptions([]string{"a", "b", "c"})
	selector.selectedIdx = 2

	// Update with fewer options - should reset index
	selector.SetOptions([]string{"x", "y"})
	if selector.selectedIdx != 0 {
		t.Errorf("After updating to fewer options, selectedIdx = %d, want 0", selector.selectedIdx)
	}

	// Update with more options - index should stay valid
	selector.selectedIdx = 1
	selector.SetOptions([]string{"p", "q", "r", "s", "t"})
	if selector.selectedIdx != 1 {
		t.Errorf("After updating to more options, selectedIdx = %d, want 1", selector.selectedIdx)
	}
}
