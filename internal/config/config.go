package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	UI          UIConfig          `yaml:"ui"`
	Performance PerformanceConfig `yaml:"performance"`
	KeyBindings KeyBindingsConfig `yaml:"keybindings"`
}

// UIConfig holds UI-related configuration
type UIConfig struct {
	Theme           string `yaml:"theme"`            // Options: dark, light, auto
	RefreshInterval string `yaml:"refresh_interval"` // Auto-refresh interval (e.g., "5s")
	ShowSystemPods  bool   `yaml:"show_system_pods"` // Show kube-system pods
	SidebarWidth    int    `yaml:"sidebar_width"`    // Sidebar width percentage
}

// PerformanceConfig holds performance-related configuration
type PerformanceConfig struct {
	MaxListItems int    `yaml:"max_list_items"` // Maximum items in lists
	CacheTTL     string `yaml:"cache_ttl"`      // Resource cache duration (e.g., "30s")
}

// KeyBindingsConfig holds customizable key bindings
type KeyBindingsConfig struct {
	Quit   []string `yaml:"quit"`
	Help   []string `yaml:"help"`
	Search []string `yaml:"search"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		UI: UIConfig{
			Theme:           "dark",
			RefreshInterval: "5s",
			ShowSystemPods:  false,
			SidebarWidth:    30,
		},
		Performance: PerformanceConfig{
			MaxListItems: 500,
			CacheTTL:     "30s",
		},
		KeyBindings: KeyBindingsConfig{
			Quit:   []string{"q", "ctrl+c"},
			Help:   []string{"?"},
			Search: []string{"/"},
		},
	}
}

// Load reads the configuration file and returns a Config struct
// If the file doesn't exist, it returns the default configuration
func Load(configPath string) (*Config, error) {
	// If no path specified, use default
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return DefaultConfig(), nil
		}
		configPath = filepath.Join(homeDir, ".k8s-tui", "config.yaml")
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and set defaults for missing fields
	if cfg.UI.Theme == "" {
		cfg.UI.Theme = "dark"
	}
	if cfg.UI.RefreshInterval == "" {
		cfg.UI.RefreshInterval = "5s"
	}
	if cfg.UI.SidebarWidth == 0 {
		cfg.UI.SidebarWidth = 30
	}
	if cfg.Performance.MaxListItems == 0 {
		cfg.Performance.MaxListItems = 500
	}
	if cfg.Performance.CacheTTL == "" {
		cfg.Performance.CacheTTL = "30s"
	}
	if len(cfg.KeyBindings.Quit) == 0 {
		cfg.KeyBindings.Quit = []string{"q", "ctrl+c"}
	}
	if len(cfg.KeyBindings.Help) == 0 {
		cfg.KeyBindings.Help = []string{"?"}
	}
	if len(cfg.KeyBindings.Search) == 0 {
		cfg.KeyBindings.Search = []string{"/"}
	}

	return &cfg, nil
}

// Save writes the configuration to a file
func (c *Config) Save(configPath string) error {
	// If no path specified, use default
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".k8s-tui", "config.yaml")
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetRefreshInterval parses and returns the refresh interval as time.Duration
func (c *Config) GetRefreshInterval() time.Duration {
	duration, err := time.ParseDuration(c.UI.RefreshInterval)
	if err != nil {
		return 5 * time.Second // Default fallback
	}
	return duration
}

// GetCacheTTL parses and returns the cache TTL as time.Duration
func (c *Config) GetCacheTTL() time.Duration {
	duration, err := time.ParseDuration(c.Performance.CacheTTL)
	if err != nil {
		return 30 * time.Second // Default fallback
	}
	return duration
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate theme
	validThemes := map[string]bool{"dark": true, "light": true, "auto": true}
	if !validThemes[c.UI.Theme] {
		return fmt.Errorf("invalid theme: %s (must be dark, light, or auto)", c.UI.Theme)
	}

	// Validate refresh interval
	if _, err := time.ParseDuration(c.UI.RefreshInterval); err != nil {
		return fmt.Errorf("invalid refresh_interval: %s", c.UI.RefreshInterval)
	}

	// Validate cache TTL
	if _, err := time.ParseDuration(c.Performance.CacheTTL); err != nil {
		return fmt.Errorf("invalid cache_ttl: %s", c.Performance.CacheTTL)
	}

	// Validate sidebar width
	if c.UI.SidebarWidth < 10 || c.UI.SidebarWidth > 90 {
		return fmt.Errorf("invalid sidebar_width: %d (must be between 10 and 90)", c.UI.SidebarWidth)
	}

	// Validate max list items
	if c.Performance.MaxListItems < 10 || c.Performance.MaxListItems > 10000 {
		return fmt.Errorf("invalid max_list_items: %d (must be between 10 and 10000)", c.Performance.MaxListItems)
	}

	return nil
}
