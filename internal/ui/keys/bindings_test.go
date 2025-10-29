package keys

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that all key bindings are initialized
	tests := []struct {
		name    string
		binding key.Binding
	}{
		{"Quit", km.Quit},
		{"Help", km.Help},
		{"Refresh", km.Refresh},
		{"Up", km.Up},
		{"Down", km.Down},
		{"PageUp", km.PageUp},
		{"PageDown", km.PageDown},
		{"Home", km.Home},
		{"End", km.End},
		{"Enter", km.Enter},
		{"Back", km.Back},
		{"Tab", km.Tab},
		{"ShiftTab", km.ShiftTab},
		{"Namespace", km.Namespace},
		{"Context", km.Context},
		{"Search", km.Search},
		{"Logs", km.Logs},
		{"Events", km.Events},
		{"YAML", km.YAML},
		{"Describe", km.Describe},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.binding.Keys()) == 0 {
				t.Errorf("%s binding has no keys", tt.name)
			}
		})
	}
}

func TestKeyMap_Quit(t *testing.T) {
	km := DefaultKeyMap()

	expectedKeys := []string{"q", "ctrl+c"}
	keys := km.Quit.Keys()

	if len(keys) != len(expectedKeys) {
		t.Errorf("expected %d keys for Quit, got %d", len(expectedKeys), len(keys))
	}

	for i, expected := range expectedKeys {
		if i < len(keys) && keys[i] != expected {
			t.Errorf("expected key %q at position %d, got %q", expected, i, keys[i])
		}
	}
}

func TestKeyMap_Navigation(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name         string
		binding      key.Binding
		expectedKeys []string
	}{
		{
			name:         "Up",
			binding:      km.Up,
			expectedKeys: []string{"up", "k"},
		},
		{
			name:         "Down",
			binding:      km.Down,
			expectedKeys: []string{"down", "j"},
		},
		{
			name:         "PageUp",
			binding:      km.PageUp,
			expectedKeys: []string{"pgup", "ctrl+u"},
		},
		{
			name:         "PageDown",
			binding:      km.PageDown,
			expectedKeys: []string{"pgdown", "ctrl+d"},
		},
		{
			name:         "Home",
			binding:      km.Home,
			expectedKeys: []string{"home", "g"},
		},
		{
			name:         "End",
			binding:      km.End,
			expectedKeys: []string{"end", "G"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()

			if len(keys) != len(tt.expectedKeys) {
				t.Errorf("expected %d keys, got %d", len(tt.expectedKeys), len(keys))
			}

			for i, expected := range tt.expectedKeys {
				if i < len(keys) && keys[i] != expected {
					t.Errorf("expected key %q at position %d, got %q", expected, i, keys[i])
				}
			}
		})
	}
}

func TestKeyMap_Selection(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name         string
		binding      key.Binding
		expectedKeys []string
	}{
		{
			name:         "Enter",
			binding:      km.Enter,
			expectedKeys: []string{"enter", "right"},
		},
		{
			name:         "Back",
			binding:      km.Back,
			expectedKeys: []string{"left", "h", "backspace", "esc"},
		},
		{
			name:         "Tab",
			binding:      km.Tab,
			expectedKeys: []string{"tab"},
		},
		{
			name:         "ShiftTab",
			binding:      km.ShiftTab,
			expectedKeys: []string{"shift+tab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()

			if len(keys) != len(tt.expectedKeys) {
				t.Errorf("expected %d keys, got %d", len(tt.expectedKeys), len(keys))
			}

			for i, expected := range tt.expectedKeys {
				if i < len(keys) && keys[i] != expected {
					t.Errorf("expected key %q at position %d, got %q", expected, i, keys[i])
				}
			}
		})
	}
}

func TestKeyMap_ResourceActions(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name         string
		binding      key.Binding
		expectedKeys []string
	}{
		{
			name:         "Namespace",
			binding:      km.Namespace,
			expectedKeys: []string{"n"},
		},
		{
			name:         "Context",
			binding:      km.Context,
			expectedKeys: []string{"c"},
		},
		{
			name:         "Search",
			binding:      km.Search,
			expectedKeys: []string{"/"},
		},
		{
			name:         "Logs",
			binding:      km.Logs,
			expectedKeys: []string{"l"},
		},
		{
			name:         "Events",
			binding:      km.Events,
			expectedKeys: []string{"5"},
		},
		{
			name:         "YAML",
			binding:      km.YAML,
			expectedKeys: []string{"y"},
		},
		{
			name:         "Describe",
			binding:      km.Describe,
			expectedKeys: []string{"d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()

			if len(keys) != len(tt.expectedKeys) {
				t.Errorf("expected %d keys, got %d", len(tt.expectedKeys), len(keys))
			}

			for i, expected := range tt.expectedKeys {
				if i < len(keys) && keys[i] != expected {
					t.Errorf("expected key %q at position %d, got %q", expected, i, keys[i])
				}
			}
		})
	}
}

func TestKeyMap_ShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	shortHelp := km.ShortHelp()

	expectedCount := 7
	if len(shortHelp) != expectedCount {
		t.Errorf("expected %d short help bindings, got %d", expectedCount, len(shortHelp))
	}

	// Verify the short help contains key bindings
	expectedBindings := []key.Binding{
		km.Up,
		km.Down,
		km.Enter,
		km.Search,
		km.Refresh,
		km.Help,
		km.Quit,
	}

	for i, expected := range expectedBindings {
		if i < len(shortHelp) {
			// Compare by checking if keys match
			expectedKeys := expected.Keys()
			actualKeys := shortHelp[i].Keys()

			if len(expectedKeys) != len(actualKeys) {
				t.Errorf("binding %d: expected %d keys, got %d", i, len(expectedKeys), len(actualKeys))
			}
		}
	}
}

func TestKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	expectedCategories := 6
	if len(fullHelp) != expectedCategories {
		t.Errorf("expected %d help categories, got %d", expectedCategories, len(fullHelp))
	}

	// Test navigation category (first category)
	if len(fullHelp) > 0 {
		navigationBindings := fullHelp[0]
		expectedNavCount := 6
		if len(navigationBindings) != expectedNavCount {
			t.Errorf("expected %d navigation bindings, got %d", expectedNavCount, len(navigationBindings))
		}
	}

	// Test selection category (second category)
	if len(fullHelp) > 1 {
		selectionBindings := fullHelp[1]
		expectedSelCount := 4
		if len(selectionBindings) != expectedSelCount {
			t.Errorf("expected %d selection bindings, got %d", expectedSelCount, len(selectionBindings))
		}
	}

	// Test actions category (third category)
	if len(fullHelp) > 2 {
		actionsBindings := fullHelp[2]
		expectedActCount := 4
		if len(actionsBindings) != expectedActCount {
			t.Errorf("expected %d action bindings, got %d", expectedActCount, len(actionsBindings))
		}
	}

	// Test resource actions category (fourth category)
	if len(fullHelp) > 3 {
		resourceBindings := fullHelp[3]
		expectedResCount := 3
		if len(resourceBindings) != expectedResCount {
			t.Errorf("expected %d resource action bindings, got %d", expectedResCount, len(resourceBindings))
		}
	}

	// Test view actions category (fifth category)
	if len(fullHelp) > 4 {
		viewBindings := fullHelp[4]
		expectedViewCount := 5
		if len(viewBindings) != expectedViewCount {
			t.Errorf("expected %d view action bindings, got %d", expectedViewCount, len(viewBindings))
		}
	}

	// Test global category (sixth category)
	if len(fullHelp) > 5 {
		globalBindings := fullHelp[5]
		expectedGlobalCount := 2
		if len(globalBindings) != expectedGlobalCount {
			t.Errorf("expected %d global bindings, got %d", expectedGlobalCount, len(globalBindings))
		}
	}
}

func TestKeyMap_AllBindingsHaveKeys(t *testing.T) {
	km := DefaultKeyMap()

	// Get all bindings from FullHelp
	fullHelp := km.FullHelp()

	for categoryIdx, category := range fullHelp {
		for bindingIdx, binding := range category {
			keys := binding.Keys()
			if len(keys) == 0 {
				t.Errorf("binding at category %d, index %d has no keys", categoryIdx, bindingIdx)
			}
		}
	}
}

func TestKeyMap_Consistency(t *testing.T) {
	// Create two instances and verify they're the same
	km1 := DefaultKeyMap()
	km2 := DefaultKeyMap()

	// Compare a few key bindings
	if len(km1.Quit.Keys()) != len(km2.Quit.Keys()) {
		t.Error("DefaultKeyMap is not consistent across calls")
	}

	if len(km1.Up.Keys()) != len(km2.Up.Keys()) {
		t.Error("DefaultKeyMap is not consistent across calls")
	}

	if len(km1.Enter.Keys()) != len(km2.Enter.Keys()) {
		t.Error("DefaultKeyMap is not consistent across calls")
	}
}
