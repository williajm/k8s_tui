package components

import (
	"strings"
	"testing"
)

func TestNewContainerSelector(t *testing.T) {
	containers := []string{"nginx", "sidecar", "init-db (init)"}

	cs := NewContainerSelector(containers)

	if cs == nil {
		t.Fatal("NewContainerSelector() returned nil")
	}

	if cs.Selector == nil {
		t.Fatal("NewContainerSelector().Selector is nil")
	}

	// Verify options are set
	if len(cs.options) != len(containers) {
		t.Errorf("NewContainerSelector() set %d options, want %d", len(cs.options), len(containers))
	}
}

func TestContainerSelector_GetSelectedContainerName(t *testing.T) {
	tests := []struct {
		name       string
		containers []string
		selectIdx  int
		want       string
	}{
		{
			name:       "regular container",
			containers: []string{"nginx", "sidecar"},
			selectIdx:  0,
			want:       "nginx",
		},
		{
			name:       "init container with suffix",
			containers: []string{"init-db (init)", "app"},
			selectIdx:  0,
			want:       "init-db",
		},
		{
			name:       "multiple init containers",
			containers: []string{"init-config (init)", "init-db (init)", "app"},
			selectIdx:  1,
			want:       "init-db",
		},
		{
			name:       "container name contains init but not suffix",
			containers: []string{"initial-setup", "app"},
			selectIdx:  0,
			want:       "initial-setup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewContainerSelector(tt.containers)
			cs.selectedIdx = tt.selectIdx

			got := cs.GetSelectedContainerName()

			if got != tt.want {
				t.Errorf("GetSelectedContainerName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContainerSelector_GetSelectedContainerName_Empty(t *testing.T) {
	cs := NewContainerSelector([]string{})

	got := cs.GetSelectedContainerName()

	if got != "" {
		t.Errorf("GetSelectedContainerName() with no containers = %q, want empty string", got)
	}
}

func TestContainerSelector_ViewWithInfo(t *testing.T) {
	containers := []string{"nginx", "sidecar"}
	cs := NewContainerSelector(containers)

	t.Run("hidden selector returns empty", func(t *testing.T) {
		cs.Hide()
		view := cs.ViewWithInfo("Select a container")

		if view != "" {
			t.Error("ViewWithInfo() should return empty string when hidden")
		}
	})

	t.Run("visible selector with info", func(t *testing.T) {
		cs.Show()
		info := "Pod has multiple containers"
		view := cs.ViewWithInfo(info)

		if view == "" {
			t.Error("ViewWithInfo() returned empty string when visible")
		}

		if !strings.Contains(view, info) {
			t.Errorf("ViewWithInfo() should contain info message %q", info)
		}
	})

	t.Run("visible selector without info", func(t *testing.T) {
		cs.Show()
		view := cs.ViewWithInfo("")

		if view == "" {
			t.Error("ViewWithInfo() returned empty string when visible")
		}
	})
}

func TestContainerSelector_InheritsSelector(t *testing.T) {
	containers := []string{"app", "sidecar", "init (init)"}
	cs := NewContainerSelector(containers)

	// Test that ContainerSelector inherits Selector methods

	t.Run("Show/Hide", func(t *testing.T) {
		cs.Hide()
		if cs.IsVisible() {
			t.Error("IsVisible() should be false after Hide()")
		}

		cs.Show()
		if !cs.IsVisible() {
			t.Error("IsVisible() should be true after Show()")
		}
	})

	t.Run("MoveUp/MoveDown", func(t *testing.T) {
		cs.selectedIdx = 0
		cs.MoveDown()

		if cs.selectedIdx != 1 {
			t.Errorf("selectedIdx = %d, want 1 after MoveDown()", cs.selectedIdx)
		}

		cs.MoveUp()
		if cs.selectedIdx != 0 {
			t.Errorf("selectedIdx = %d, want 0 after MoveUp()", cs.selectedIdx)
		}
	})

	t.Run("GetSelected returns with suffix", func(t *testing.T) {
		cs.selectedIdx = 2 // "init (init)"
		selected := cs.GetSelected()

		if selected != "init (init)" {
			t.Errorf("GetSelected() = %q, want %q", selected, "init (init)")
		}

		// GetSelectedContainerName should strip suffix
		containerName := cs.GetSelectedContainerName()
		if containerName != "init" {
			t.Errorf("GetSelectedContainerName() = %q, want %q", containerName, "init")
		}
	})
}

func TestContainerSelector_SetOptions(t *testing.T) {
	cs := NewContainerSelector([]string{"old1", "old2"})

	newContainers := []string{"new1", "new2", "new3"}
	cs.SetOptions(newContainers)

	if len(cs.options) != len(newContainers) {
		t.Errorf("SetOptions() set %d options, want %d", len(cs.options), len(newContainers))
	}

	for i, expected := range newContainers {
		if cs.options[i] != expected {
			t.Errorf("options[%d] = %q, want %q", i, cs.options[i], expected)
		}
	}
}

func TestContainerSelector_EmptySuffix(t *testing.T) {
	// Test edge case where container name is exactly " (init)"
	cs := NewContainerSelector([]string{" (init)"})
	cs.selectedIdx = 0

	got := cs.GetSelectedContainerName()

	// Should return " (init)" because length is exactly 7, not > 7
	// (the check requires len(selected) > 7 to strip the suffix)
	if got != " (init)" {
		t.Errorf("GetSelectedContainerName() for ' (init)' = %q, want ' (init)'", got)
	}
}

func TestContainerSelector_MultipleContainers(t *testing.T) {
	// Test with realistic multi-container pod scenario
	containers := []string{
		"init-migrate (init)",
		"init-config (init)",
		"app",
		"nginx-proxy",
		"log-collector",
	}

	cs := NewContainerSelector(containers)
	cs.Show()

	// Select each container and verify name stripping
	tests := []struct {
		idx      int
		wantName string
	}{
		{0, "init-migrate"},
		{1, "init-config"},
		{2, "app"},
		{3, "nginx-proxy"},
		{4, "log-collector"},
	}

	for _, tt := range tests {
		cs.selectedIdx = tt.idx
		got := cs.GetSelectedContainerName()
		if got != tt.wantName {
			t.Errorf("Container at index %d: GetSelectedContainerName() = %q, want %q", tt.idx, got, tt.wantName)
		}
	}
}

func TestContainerSelector_View(t *testing.T) {
	containers := []string{"nginx", "app", "sidecar"}
	cs := NewContainerSelector(containers)

	t.Run("hidden returns empty", func(t *testing.T) {
		cs.Hide()
		view := cs.View()
		if view != "" {
			t.Error("View() should return empty when hidden")
		}
	})

	t.Run("visible returns content", func(t *testing.T) {
		cs.Show()
		view := cs.View()
		if view == "" {
			t.Error("View() should return content when visible")
		}
	})
}

func TestContainerSelector_SuffixEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		want          string
	}{
		{
			name:          "exactly 7 chars ending in (init)",
			containerName: "x (init)",
			want:          "x",
		},
		{
			name:          "less than 7 chars",
			containerName: "short",
			want:          "short",
		},
		{
			name:          "contains (init) but not at end",
			containerName: "(init) container",
			want:          "(init) container",
		},
		{
			name:          "multiple (init) occurrences",
			containerName: "init (init) (init)",
			want:          "init (init)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewContainerSelector([]string{tt.containerName})
			cs.selectedIdx = 0

			got := cs.GetSelectedContainerName()

			if got != tt.want {
				t.Errorf("GetSelectedContainerName() = %q, want %q", got, tt.want)
			}
		})
	}
}
