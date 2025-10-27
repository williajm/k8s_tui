package config

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDarkColorScheme(t *testing.T) {
	scheme := DarkColorScheme()

	if scheme.Primary != lipgloss.Color("#00ADD8") {
		t.Errorf("expected primary color #00ADD8, got %s", scheme.Primary)
	}

	if scheme.Background != lipgloss.Color("#1A1A1A") {
		t.Errorf("expected background color #1A1A1A, got %s", scheme.Background)
	}

	if scheme.Text != lipgloss.Color("#FFFFFF") {
		t.Errorf("expected text color #FFFFFF, got %s", scheme.Text)
	}
}

func TestLightColorScheme(t *testing.T) {
	scheme := LightColorScheme()

	if scheme.Primary != lipgloss.Color("#0066CC") {
		t.Errorf("expected primary color #0066CC, got %s", scheme.Primary)
	}

	if scheme.Background != lipgloss.Color("#FFFFFF") {
		t.Errorf("expected background color #FFFFFF, got %s", scheme.Background)
	}

	if scheme.Text != lipgloss.Color("#000000") {
		t.Errorf("expected text color #000000, got %s", scheme.Text)
	}
}

func TestGetColorScheme(t *testing.T) {
	tests := []struct {
		name        string
		themeType   ThemeType
		expectError bool
	}{
		{
			name:        "dark theme",
			themeType:   ThemeDark,
			expectError: false,
		},
		{
			name:        "light theme",
			themeType:   ThemeLight,
			expectError: false,
		},
		{
			name:        "auto theme",
			themeType:   ThemeAuto,
			expectError: false,
		},
		{
			name:        "invalid theme",
			themeType:   ThemeType("invalid"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, err := GetColorScheme(tt.themeType)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}

				// Verify scheme has all required colors
				if string(scheme.Primary) == "" {
					t.Error("primary color is empty")
				}
				if string(scheme.Background) == "" {
					t.Error("background color is empty")
				}
				if string(scheme.Text) == "" {
					t.Error("text color is empty")
				}
			}
		})
	}
}

func TestApplyColorScheme(t *testing.T) {
	scheme := DarkColorScheme()
	styles := scheme.ApplyColorScheme()

	// Verify styles are created
	if styles.Base.GetForeground() != scheme.Text {
		t.Errorf("base style text color doesn't match scheme")
	}

	if styles.Base.GetBackground() != scheme.Background {
		t.Errorf("base style background color doesn't match scheme")
	}

	// Verify header style
	if styles.Header.GetBackground() != scheme.Primary {
		t.Errorf("header style background doesn't match primary color")
	}

	// Verify error style
	if styles.Error.GetForeground() != scheme.Error {
		t.Errorf("error style color doesn't match error color")
	}
}

func TestStylesStatusStyle(t *testing.T) {
	scheme := DarkColorScheme()
	styles := scheme.ApplyColorScheme()

	tests := []struct {
		status       string
		expectedType string
	}{
		{"Running", "running"},
		{"Succeeded", "running"},
		{"Active", "running"},
		{"Ready", "running"},
		{"Pending", "pending"},
		{"Creating", "pending"},
		{"Waiting", "pending"},
		{"Failed", "error"},
		{"Error", "error"},
		{"CrashLoopBackOff", "error"},
		{"ImagePullBackOff", "error"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			// Just verify that calling StatusStyle doesn't panic
			_ = styles.StatusStyle(tt.status)
		})
	}
}

func TestStylesRenderKeyHelp(t *testing.T) {
	scheme := DarkColorScheme()
	styles := scheme.ApplyColorScheme()

	result := styles.RenderKeyHelp("q", "quit")

	// Verify result is not empty
	if result == "" {
		t.Error("RenderKeyHelp returned empty string")
	}

	// Verify it contains both key and description
	if len(result) < 6 { // "q quit" at minimum
		t.Errorf("RenderKeyHelp result too short: %s", result)
	}
}

func TestStylesRenderDetailRow(t *testing.T) {
	scheme := DarkColorScheme()
	styles := scheme.ApplyColorScheme()

	result := styles.RenderDetailRow("Name", "test-pod")

	// Verify result is not empty
	if result == "" {
		t.Error("RenderDetailRow returned empty string")
	}

	// Verify it contains both label and value
	if len(result) < 13 { // "Name: test-pod" at minimum
		t.Errorf("RenderDetailRow result too short: %s", result)
	}
}

func TestColorSchemeConsistency(t *testing.T) {
	darkScheme := DarkColorScheme()
	lightScheme := LightColorScheme()

	// Verify both schemes have all required colors
	schemes := []struct {
		name   string
		scheme ColorScheme
	}{
		{"dark", darkScheme},
		{"light", lightScheme},
	}

	for _, s := range schemes {
		t.Run(s.name, func(t *testing.T) {
			if string(s.scheme.Primary) == "" {
				t.Error("primary color is empty")
			}
			if string(s.scheme.Secondary) == "" {
				t.Error("secondary color is empty")
			}
			if string(s.scheme.Accent) == "" {
				t.Error("accent color is empty")
			}
			if string(s.scheme.Success) == "" {
				t.Error("success color is empty")
			}
			if string(s.scheme.Warning) == "" {
				t.Error("warning color is empty")
			}
			if string(s.scheme.Error) == "" {
				t.Error("error color is empty")
			}
			if string(s.scheme.Info) == "" {
				t.Error("info color is empty")
			}
			if string(s.scheme.Border) == "" {
				t.Error("border color is empty")
			}
			if string(s.scheme.Text) == "" {
				t.Error("text color is empty")
			}
			if string(s.scheme.TextDim) == "" {
				t.Error("text dim color is empty")
			}
			if string(s.scheme.Background) == "" {
				t.Error("background color is empty")
			}
			if string(s.scheme.Selected) == "" {
				t.Error("selected color is empty")
			}
			if string(s.scheme.Highlight) == "" {
				t.Error("highlight color is empty")
			}
		})
	}
}

func TestThemeTypeConstants(t *testing.T) {
	if ThemeDark != "dark" {
		t.Errorf("expected ThemeDark to be 'dark', got %s", ThemeDark)
	}

	if ThemeLight != "light" {
		t.Errorf("expected ThemeLight to be 'light', got %s", ThemeLight)
	}

	if ThemeAuto != "auto" {
		t.Errorf("expected ThemeAuto to be 'auto', got %s", ThemeAuto)
	}
}
