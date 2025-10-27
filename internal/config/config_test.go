package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.UI.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %s", cfg.UI.Theme)
	}

	if cfg.UI.RefreshInterval != "5s" {
		t.Errorf("expected refresh interval '5s', got %s", cfg.UI.RefreshInterval)
	}

	if cfg.UI.ShowSystemPods != false {
		t.Errorf("expected show_system_pods false, got %v", cfg.UI.ShowSystemPods)
	}

	if cfg.UI.SidebarWidth != 30 {
		t.Errorf("expected sidebar_width 30, got %d", cfg.UI.SidebarWidth)
	}

	if cfg.Performance.MaxListItems != 500 {
		t.Errorf("expected max_list_items 500, got %d", cfg.Performance.MaxListItems)
	}

	if cfg.Performance.CacheTTL != "30s" {
		t.Errorf("expected cache_ttl '30s', got %s", cfg.Performance.CacheTTL)
	}

	if len(cfg.KeyBindings.Quit) != 2 {
		t.Errorf("expected 2 quit keybindings, got %d", len(cfg.KeyBindings.Quit))
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("expected no error for non-existent file, got %v", err)
	}

	// Should return default config
	if cfg.UI.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got %s", cfg.UI.Theme)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config
	cfg := DefaultConfig()
	cfg.UI.Theme = "light"
	cfg.UI.RefreshInterval = "10s"

	// Save config
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify values
	if loadedCfg.UI.Theme != "light" {
		t.Errorf("expected theme 'light', got %s", loadedCfg.UI.Theme)
	}

	if loadedCfg.UI.RefreshInterval != "10s" {
		t.Errorf("expected refresh interval '10s', got %s", loadedCfg.UI.RefreshInterval)
	}
}

func TestGetRefreshInterval(t *testing.T) {
	cfg := DefaultConfig()

	duration := cfg.GetRefreshInterval()
	expected := 5 * time.Second

	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestGetRefreshIntervalInvalid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.UI.RefreshInterval = "invalid"

	duration := cfg.GetRefreshInterval()
	expected := 5 * time.Second // Default fallback

	if duration != expected {
		t.Errorf("expected fallback %v, got %v", expected, duration)
	}
}

func TestGetCacheTTL(t *testing.T) {
	cfg := DefaultConfig()

	duration := cfg.GetCacheTTL()
	expected := 30 * time.Second

	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		modifyFn  func(*Config)
		expectErr bool
	}{
		{
			name:      "valid config",
			modifyFn:  func(_ *Config) {},
			expectErr: false,
		},
		{
			name: "invalid theme",
			modifyFn: func(c *Config) {
				c.UI.Theme = "invalid"
			},
			expectErr: true,
		},
		{
			name: "invalid refresh interval",
			modifyFn: func(c *Config) {
				c.UI.RefreshInterval = "not-a-duration"
			},
			expectErr: true,
		},
		{
			name: "invalid cache ttl",
			modifyFn: func(c *Config) {
				c.Performance.CacheTTL = "not-a-duration"
			},
			expectErr: true,
		},
		{
			name: "sidebar width too small",
			modifyFn: func(c *Config) {
				c.UI.SidebarWidth = 5
			},
			expectErr: true,
		},
		{
			name: "sidebar width too large",
			modifyFn: func(c *Config) {
				c.UI.SidebarWidth = 95
			},
			expectErr: true,
		},
		{
			name: "max list items too small",
			modifyFn: func(c *Config) {
				c.Performance.MaxListItems = 5
			},
			expectErr: true,
		},
		{
			name: "max list items too large",
			modifyFn: func(c *Config) {
				c.Performance.MaxListItems = 20000
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modifyFn(cfg)

			err := cfg.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := DefaultConfig()
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestLoadWithMissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write minimal config
	content := `ui:
  theme: light
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify defaults are applied for missing fields
	if cfg.UI.RefreshInterval != "5s" {
		t.Errorf("expected default refresh_interval '5s', got %s", cfg.UI.RefreshInterval)
	}

	if cfg.UI.SidebarWidth != 30 {
		t.Errorf("expected default sidebar_width 30, got %d", cfg.UI.SidebarWidth)
	}
}
