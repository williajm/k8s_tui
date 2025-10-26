package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/williajm/k8s-tui/internal/app"
	"github.com/williajm/k8s-tui/internal/k8s"
)

var (
	kubeconfigPath string
	contextName    string
	namespace      string
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
	rootCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.Flags().StringVar(&contextName, "context", "", "Kubernetes context to use")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace to use")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Create Kubernetes client
	client, err := k8s.NewClient(kubeconfigPath, contextName, namespace)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(
		app.NewModel(client),
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running application: %w", err)
	}

	return nil
}
