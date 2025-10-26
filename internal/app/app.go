package app

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williajm/k8s-tui/internal/k8s"
	"github.com/williajm/k8s-tui/internal/models"
	"github.com/williajm/k8s-tui/internal/ui/components"
	"github.com/williajm/k8s-tui/internal/ui/keys"
)

// Model represents the application state
type Model struct {
	client    *k8s.Client
	keyMap    keys.KeyMap
	header    *components.Header
	footer    *components.Footer
	podList   *components.PodList
	width     int
	height    int
	err       error
	loading   bool
	showHelp  bool
	connected bool
}

// Message types
type podsLoadedMsg struct {
	pods []models.PodInfo
	err  error
}

type tickMsg time.Time

type errMsg struct{ err error }

// NewModel creates a new application model
func NewModel(client *k8s.Client) Model {
	header := components.NewHeader(
		client.GetCurrentContext(),
		client.GetNamespace(),
		false, // Will be set to true after first successful load
	)

	return Model{
		client:    client,
		keyMap:    keys.DefaultKeyMap(),
		header:    header,
		footer:    components.NewFooter(),
		podList:   components.NewPodList(),
		connected: false,
		loading:   true,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadPods(),
		tickCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update component sizes
		m.header.SetWidth(m.width)
		m.footer.SetWidth(m.width)

		// Pod list gets remaining height (minus header and footer)
		listHeight := m.height - 4 // 1 for header, 2 for footer, 1 for padding
		m.podList.SetSize(m.width, listHeight)

	case podsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.connected = false
		} else {
			m.podList.SetPods(msg.pods)
			m.connected = true
			m.err = nil
		}
		m.header.SetConnected(m.connected)

	case tickMsg:
		// Auto-refresh every 5 seconds
		return m, tea.Batch(
			m.loadPods(),
			tickCmd(),
		)

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keyMap.Help):
		m.showHelp = !m.showHelp
		return m, nil

	case key.Matches(msg, m.keyMap.Refresh):
		m.loading = true
		return m, m.loadPods()
	}

	// Don't process other keys if help is shown
	if m.showHelp {
		return m, nil
	}

	// Navigation keys
	switch {
	case key.Matches(msg, m.keyMap.Up):
		m.podList.MoveUp()

	case key.Matches(msg, m.keyMap.Down):
		m.podList.MoveDown()

	case key.Matches(msg, m.keyMap.PageUp):
		m.podList.PageUp()

	case key.Matches(msg, m.keyMap.PageDown):
		m.podList.PageDown()

	case key.Matches(msg, m.keyMap.Home):
		m.podList.Home()

	case key.Matches(msg, m.keyMap.End):
		m.podList.End()
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show help if requested
	if m.showHelp {
		return m.viewHelp()
	}

	// Show error if present
	if m.err != nil {
		return m.viewError()
	}

	// Build the main view
	header := m.header.View()
	footer := m.footer.View()
	podList := m.podList.View()

	// Stack components vertically
	view := header + "\n" + podList + "\n" + footer

	return view
}

// viewHelp renders the help screen
func (m Model) viewHelp() string {
	header := m.header.View()
	help := m.footer.ViewDetailed()
	helpPrompt := "\nPress ? or ESC to close help"

	view := header + "\n" + help + helpPrompt

	return view
}

// viewError renders an error message
func (m Model) viewError() string {
	header := m.header.View()
	footer := m.footer.View()

	errorMsg := fmt.Sprintf(
		"Error: %v\n\nPress 'r' to retry or 'q' to quit",
		m.err,
	)

	view := header + "\n" + errorMsg + "\n" + footer

	return view
}

// loadPods fetches pods from Kubernetes
func (m Model) loadPods() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		podList, err := m.client.GetPods(ctx, m.client.GetNamespace())
		if err != nil {
			return podsLoadedMsg{err: err}
		}

		// Convert to PodInfo
		pods := make([]models.PodInfo, len(podList.Items))
		for i, pod := range podList.Items {
			pods[i] = models.NewPodInfo(&pod)
		}

		return podsLoadedMsg{pods: pods}
	}
}

// tickCmd creates a tick command for auto-refresh
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
