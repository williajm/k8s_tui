package components

import (
	"strings"
	"testing"
)

func TestNewFooter(t *testing.T) {
	f := NewFooter()

	if f == nil {
		t.Fatal("NewFooter returned nil")
	}

	if f.width != 80 {
		t.Errorf("expected default width 80, got %d", f.width)
	}
}

func TestFooter_SetWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard width",
			width: 120,
		},
		{
			name:  "small width",
			width: 40,
		},
		{
			name:  "large width",
			width: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFooter()
			f.SetWidth(tt.width)

			if f.width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, f.width)
			}
		})
	}
}

func TestFooter_View(t *testing.T) {
	f := NewFooter()
	view := f.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}

	// Check for all expected shortcuts in the basic view
	expectedShortcuts := []string{
		"Navigate",
		"Select",
		"Switch",
		"Search",
		"Refresh",
		"Help",
		"Quit",
	}

	for _, shortcut := range expectedShortcuts {
		if !strings.Contains(view, shortcut) {
			t.Errorf("expected view to contain %q, but it didn't", shortcut)
		}
	}

	// Check for key bindings
	expectedKeys := []string{
		"↑↓",
		"Enter",
		"Tab",
		"/",
		"r",
		"?",
		"q",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(view, key) {
			t.Errorf("expected view to contain key %q, but it didn't", key)
		}
	}
}

func TestFooter_ViewWithWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard width",
			width: 100,
		},
		{
			name:  "large width",
			width: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFooter()
			f.SetWidth(tt.width)

			view := f.View()

			if view == "" {
				t.Fatal("View returned empty string")
			}

			// Should still contain shortcuts regardless of width
			if !strings.Contains(view, "Navigate") {
				t.Error("expected view to contain shortcuts")
			}
		})
	}
}

func TestFooter_ViewDetailed(t *testing.T) {
	f := NewFooter()
	view := f.ViewDetailed()

	if view == "" {
		t.Fatal("ViewDetailed returned empty string")
	}

	// Check for category headers
	expectedCategories := []string{
		"Navigation",
		"Selection",
		"Actions",
		"Resource Actions",
		"Global",
	}

	for _, category := range expectedCategories {
		if !strings.Contains(view, category) {
			t.Errorf("expected detailed view to contain category %q", category)
		}
	}

	// Check for detailed shortcuts
	expectedShortcuts := []string{
		"Move up",
		"Move down",
		"Page up",
		"Page down",
		"Go to top",
		"Go to bottom",
		"Open detail/expand",
		"Go back/collapse",
		"Switch panes",
		"Previous pane",
		"Change namespace",
		"Change context",
		"Search/filter",
		"Refresh",
		"View logs",
		"View events",
		"View YAML",
		"Describe resource",
		"Toggle help",
		"Quit application",
	}

	for _, shortcut := range expectedShortcuts {
		if !strings.Contains(view, shortcut) {
			t.Errorf("expected detailed view to contain %q", shortcut)
		}
	}

	// Check for key bindings in detailed view
	expectedKeys := []string{
		"↑/k",
		"↓/j",
		"PgUp/Ctrl+U",
		"PgDn/Ctrl+D",
		"g",
		"G",
		"Enter/→/l",
		"←/h/Backspace",
		"Tab",
		"Shift+Tab",
		"n",
		"c",
		"/",
		"r/F5",
		"L",
		"E",
		"Y",
		"D",
		"?",
		"q/Ctrl+C",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(view, key) {
			t.Errorf("expected detailed view to contain key %q", key)
		}
	}
}

func TestFooter_ViewDetailedWithWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard width",
			width: 80,
		},
		{
			name:  "large width",
			width: 150,
		},
		{
			name:  "small width",
			width: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFooter()
			f.SetWidth(tt.width)

			view := f.ViewDetailed()

			if view == "" {
				t.Fatal("ViewDetailed returned empty string")
			}

			// Should contain all categories regardless of width
			if !strings.Contains(view, "Navigation") {
				t.Error("expected detailed view to contain categories")
			}
		})
	}
}

func TestFooter_BothViews(t *testing.T) {
	f := NewFooter()

	basicView := f.View()
	detailedView := f.ViewDetailed()

	if basicView == "" {
		t.Fatal("View returned empty string")
	}

	if detailedView == "" {
		t.Fatal("ViewDetailed returned empty string")
	}

	// Detailed view should be longer than basic view
	if len(detailedView) <= len(basicView) {
		t.Error("expected detailed view to be longer than basic view")
	}

	// Both should contain help text
	if !strings.Contains(basicView, "Help") {
		t.Error("basic view should contain Help")
	}

	if !strings.Contains(detailedView, "help") {
		t.Error("detailed view should contain help text")
	}
}
