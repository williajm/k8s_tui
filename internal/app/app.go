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

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeDetail
)

// Model represents the application state
type Model struct {
	client            *k8s.Client
	keyMap            keys.KeyMap
	header            *components.Header
	footer            *components.Footer
	tabs              *components.Tabs
	resourceList      *components.ResourceList
	detailView        *components.DetailView
	namespaceSelector *components.Selector
	width             int
	height            int
	err               error
	loading           bool
	showHelp          bool
	connected         bool
	viewMode          ViewMode
	searchMode        bool
	searchQuery       string
}

// Message types
type resourcesLoadedMsg struct {
	resourceType components.ResourceType
	pods         []models.PodInfo
	services     []models.ServiceInfo
	deployments  []models.DeploymentInfo
	statefulSets []models.StatefulSetInfo
	err          error
}

type namespacesLoadedMsg struct {
	namespaces []string
	err        error
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
		client:            client,
		keyMap:            keys.DefaultKeyMap(),
		header:            header,
		footer:            components.NewFooter(),
		tabs:              components.NewTabs(),
		resourceList:      components.NewResourceList(components.ResourceTypePod),
		detailView:        components.NewDetailView(),
		namespaceSelector: components.NewSelector("Select Namespace"),
		connected:         false,
		loading:           true,
		viewMode:          ViewModeList,
		searchMode:        false,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadResources(),
		tickCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle namespace selector first if visible
	if m.namespaceSelector.IsVisible() {
		return m.handleNamespaceSelector(msg)
	}

	// Handle search mode
	if m.searchMode {
		return m.handleSearchMode(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update component sizes
		m.header.SetWidth(m.width)
		m.footer.SetWidth(m.width)
		m.tabs.SetWidth(m.width)

		// Resource list and detail view get remaining height
		remainingHeight := m.height - 6 // Header, tabs, footer, padding
		m.resourceList.SetSize(m.width, remainingHeight)
		m.detailView.SetSize(m.width, remainingHeight)

		// Selector size
		selectorWidth := min(m.width-10, 50)
		selectorHeight := min(m.height-6, 20)
		m.namespaceSelector.SetSize(selectorWidth, selectorHeight)

	case resourcesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.connected = false
		} else {
			m.connected = true
			m.err = nil

			// Update the appropriate resource list
			switch msg.resourceType {
			case components.ResourceTypePod:
				m.resourceList.SetPods(msg.pods)
			case components.ResourceTypeService:
				m.resourceList.SetServices(msg.services)
			case components.ResourceTypeDeployment:
				m.resourceList.SetDeployments(msg.deployments)
			case components.ResourceTypeStatefulSet:
				m.resourceList.SetStatefulSets(msg.statefulSets)
			}
		}
		m.header.SetConnected(m.connected)

	case namespacesLoadedMsg:
		if msg.err == nil && len(msg.namespaces) > 0 {
			m.namespaceSelector.SetOptions(msg.namespaces)
		}

	case tickMsg:
		// Auto-refresh every 5 seconds
		return m, tea.Batch(
			m.loadResources(),
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
		return m, m.loadResources()

	case key.Matches(msg, m.keyMap.Namespace):
		// Show namespace selector
		m.namespaceSelector.Show()
		return m, m.loadNamespaces()

	case key.Matches(msg, m.keyMap.Search):
		// Enter search mode
		m.searchMode = true
		m.searchQuery = ""
		return m, nil

	case key.Matches(msg, m.keyMap.Tab):
		// Next tab
		m.tabs.NextTab()
		m.resourceList.SetResourceType(components.ResourceType(m.tabs.GetActiveTab()))
		m.viewMode = ViewModeList // Reset to list view when switching tabs
		m.loading = true
		return m, m.loadResources()

	case key.Matches(msg, m.keyMap.ShiftTab):
		// Previous tab
		m.tabs.PrevTab()
		m.resourceList.SetResourceType(components.ResourceType(m.tabs.GetActiveTab()))
		m.viewMode = ViewModeList // Reset to list view when switching tabs
		m.loading = true
		return m, m.loadResources()

	case key.Matches(msg, m.keyMap.Enter):
		// Enter detail view
		if m.viewMode == ViewModeList {
			m.viewMode = ViewModeDetail
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Back):
		// Exit detail view
		if m.viewMode == ViewModeDetail {
			m.viewMode = ViewModeList
			return m, nil
		}
	}

	// Don't process other keys if help is shown
	if m.showHelp {
		return m, nil
	}

	// Navigation keys (only in list view)
	if m.viewMode == ViewModeList {
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.resourceList.MoveUp()

		case key.Matches(msg, m.keyMap.Down):
			m.resourceList.MoveDown()

		case key.Matches(msg, m.keyMap.PageUp):
			m.resourceList.PageUp()

		case key.Matches(msg, m.keyMap.PageDown):
			m.resourceList.PageDown()

		case key.Matches(msg, m.keyMap.Home):
			m.resourceList.Home()

		case key.Matches(msg, m.keyMap.End):
			m.resourceList.End()
		}
	}

	return m, nil
}

// handleNamespaceSelector handles input when namespace selector is visible
func (m Model) handleNamespaceSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.namespaceSelector.MoveUp()

		case key.Matches(msg, m.keyMap.Down):
			m.namespaceSelector.MoveDown()

		case key.Matches(msg, m.keyMap.Enter):
			// Select namespace and reload resources
			selectedNS := m.namespaceSelector.GetSelected()
			if selectedNS != "" {
				m.client.SetNamespace(selectedNS)
				m.header = components.NewHeader(
					m.client.GetCurrentContext(),
					m.client.GetNamespace(),
					m.connected,
				)
				m.namespaceSelector.Hide()
				m.loading = true
				return m, m.loadResources()
			}

		case key.Matches(msg, m.keyMap.Back), key.Matches(msg, m.keyMap.Quit):
			// Cancel namespace selection
			m.namespaceSelector.Hide()
		}

	case namespacesLoadedMsg:
		if msg.err == nil && len(msg.namespaces) > 0 {
			m.namespaceSelector.SetOptions(msg.namespaces)
		}
	}

	return m, nil
}

// handleSearchMode handles input when in search mode
func (m Model) handleSearchMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Exit search mode
			m.searchMode = false
			m.searchQuery = ""
			m.resourceList.SetSearchFilter("")

		case tea.KeyEnter:
			// Apply search and exit search mode
			m.searchMode = false
			m.resourceList.SetSearchFilter(m.searchQuery)

		case tea.KeyBackspace:
			// Remove last character
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}

		case tea.KeyRunes:
			// Add character to search query
			m.searchQuery += string(msg.Runes)
		}
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show namespace selector if visible
	if m.namespaceSelector.IsVisible() {
		return m.viewNamespaceSelector()
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
	tabs := m.tabs.View()
	footer := m.footer.View()

	var mainContent string
	if m.viewMode == ViewModeDetail {
		mainContent = m.viewDetail()
	} else {
		mainContent = m.resourceList.View()
	}

	// Add search indicator if in search mode
	if m.searchMode {
		searchPrompt := fmt.Sprintf("Search: %s_", m.searchQuery)
		mainContent = mainContent + "\n" + searchPrompt
	}

	// Stack components vertically
	view := header + "\n" + tabs + "\n" + mainContent + "\n" + footer

	return view
}

// viewDetail renders the detail view based on current resource type
func (m Model) viewDetail() string {
	switch components.ResourceType(m.tabs.GetActiveTab()) {
	case components.ResourceTypePod:
		pod := m.resourceList.GetSelectedPod()
		return m.detailView.ViewPod(pod)

	case components.ResourceTypeService:
		service := m.resourceList.GetSelectedService()
		return m.detailView.ViewService(service)

	case components.ResourceTypeDeployment:
		deployment := m.resourceList.GetSelectedDeployment()
		return m.detailView.ViewDeployment(deployment)

	case components.ResourceTypeStatefulSet:
		statefulSet := m.resourceList.GetSelectedStatefulSet()
		return m.detailView.ViewStatefulSet(statefulSet)

	default:
		return "Unknown resource type"
	}
}

// viewNamespaceSelector renders the namespace selector
func (m Model) viewNamespaceSelector() string {
	header := m.header.View()
	footer := m.footer.View()
	selector := m.namespaceSelector.View()

	// Simple overlay (center the selector)
	return header + "\n\n" + selector + "\n\n" + footer
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

// loadResources fetches resources from Kubernetes based on current tab
func (m Model) loadResources() tea.Cmd {
	resourceType := components.ResourceType(m.tabs.GetActiveTab())

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		msg := resourcesLoadedMsg{
			resourceType: resourceType,
		}

		namespace := m.client.GetNamespace()

		switch resourceType {
		case components.ResourceTypePod:
			podList, err := m.client.GetPods(ctx, namespace)
			if err != nil {
				msg.err = err
				return msg
			}
			pods := make([]models.PodInfo, len(podList.Items))
			for i, pod := range podList.Items {
				pods[i] = models.NewPodInfo(&pod)
			}
			msg.pods = pods

		case components.ResourceTypeService:
			serviceList, err := m.client.GetServices(ctx, namespace)
			if err != nil {
				msg.err = err
				return msg
			}
			services := make([]models.ServiceInfo, len(serviceList.Items))
			for i, svc := range serviceList.Items {
				services[i] = models.NewServiceInfo(&svc)
			}
			msg.services = services

		case components.ResourceTypeDeployment:
			deploymentList, err := m.client.GetDeployments(ctx, namespace)
			if err != nil {
				msg.err = err
				return msg
			}
			deployments := make([]models.DeploymentInfo, len(deploymentList.Items))
			for i, dep := range deploymentList.Items {
				deployments[i] = models.NewDeploymentInfo(&dep)
			}
			msg.deployments = deployments

		case components.ResourceTypeStatefulSet:
			statefulSetList, err := m.client.GetStatefulSets(ctx, namespace)
			if err != nil {
				msg.err = err
				return msg
			}
			statefulSets := make([]models.StatefulSetInfo, len(statefulSetList.Items))
			for i, sts := range statefulSetList.Items {
				statefulSets[i] = models.NewStatefulSetInfo(&sts)
			}
			msg.statefulSets = statefulSets
		}

		return msg
	}
}

// loadNamespaces fetches all namespaces
func (m Model) loadNamespaces() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		nsList, err := m.client.GetNamespaces(ctx)
		if err != nil {
			return namespacesLoadedMsg{err: err}
		}

		namespaces := make([]string, len(nsList.Items))
		for i, ns := range nsList.Items {
			namespaces[i] = ns.Name
		}

		return namespacesLoadedMsg{namespaces: namespaces}
	}
}

// tickCmd creates a tick command for auto-refresh
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
