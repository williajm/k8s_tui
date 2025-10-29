package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/williajm/k8s-tui/internal/app"
	"github.com/williajm/k8s-tui/internal/config"
	"github.com/williajm/k8s-tui/internal/debug"
	"github.com/williajm/k8s-tui/internal/k8s"
	"k8s.io/klog/v2"
)

var (
	kubeconfigPath string
	contextName    string
	namespace      string
	configPath     string
	debugMode      bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "k8s-tui",
		Short: "A terminal UI for Kubernetes cluster management",
		Long: `k8s-tui is a fast, keyboard-driven terminal user interface for managing
Kubernetes clusters. It provides real-time monitoring and navigation of your cluster resources.`,
		RunE: run,
	}

	// Define flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file (default: ~/.k8s-tui/config.yaml)")
	rootCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.Flags().StringVar(&contextName, "context", "", "Kubernetes context to use")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace to use")
	rootCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug logging to ~/.k8s-tui/debug.log")

	// Add init-config subcommand
	initConfigCmd := &cobra.Command{
		Use:   "init-config",
		Short: "Initialize configuration file with default values",
		Long:  `Creates a configuration file at ~/.k8s-tui/config.yaml with default values.`,
		RunE:  initConfig,
	}
	rootCmd.AddCommand(initConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(_ *cobra.Command, _ []string) error {
	// Initialize debug logging
	if err := debug.InitLogger(debugMode); err != nil {
		return fmt.Errorf("failed to initialize debug logger: %w", err)
	}
	defer debug.CloseLogger()

	// Suppress klog output to prevent Kubernetes client-go from corrupting TUI
	// This is CRITICAL - without this, k8s client-go writes to stderr and corrupts the terminal
	if debugMode {
		debug.GetLogger().Log("Debug mode enabled")
		debug.GetLogger().Log("Suppressing klog output to prevent TUI corruption")
	}
	klog.SetOutput(os.NewFile(0, os.DevNull))
	klogFlags := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(klogFlags)
	_ = klogFlags.Set("logtostderr", "false")
	_ = klogFlags.Set("v", "-1")

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(kubeconfigPath, contextName, namespace)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Test connection with configured timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Create the Bubble Tea program with configuration
	p := tea.NewProgram(
		app.NewModelWithConfig(client, cfg),
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running application: %w", err)
	}

	return nil
}

func initConfig(_ *cobra.Command, _ []string) error {
	// Determine config path
	cfgPath := configPath
	if cfgPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		cfgPath = filepath.Join(homeDir, ".k8s-tui", "config.yaml")
	}

	// Check if file already exists
	if _, err := os.Stat(cfgPath); err == nil {
		fmt.Printf("Configuration file already exists at: %s\n", cfgPath)
		fmt.Println("To overwrite, delete the existing file first.")
		return nil
	}

	// Create default config
	cfg := config.DefaultConfig()

	// Save config
	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration file created successfully at: %s\n", cfgPath)
	fmt.Println("\nDefault configuration:")
	fmt.Printf("  Theme: %s\n", cfg.UI.Theme)
	fmt.Printf("  Refresh Interval: %s\n", cfg.UI.RefreshInterval)
	fmt.Printf("  Show System Pods: %v\n", cfg.UI.ShowSystemPods)
	fmt.Printf("  Sidebar Width: %d%%\n", cfg.UI.SidebarWidth)
	fmt.Printf("  Max List Items: %d\n", cfg.Performance.MaxListItems)
	fmt.Printf("  Cache TTL: %s\n", cfg.Performance.CacheTTL)
	fmt.Println("\nEdit the file to customize your configuration.")

	return nil
}
