package app

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/williajm/k8s-tui/internal/config"
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
	ViewModeLogStream
	ViewModeDescribe
	ViewModeContainerSelect
)

// Model represents the application state
type Model struct {
	client            *k8s.Client
	config            *config.Config
	styles            config.Styles
	keyMap            keys.KeyMap
	header            *components.Header
	footer            *components.Footer
	tabs              *components.Tabs
	resourceList      *components.ResourceList
	detailView        *components.DetailView
	namespaceSelector *components.Selector
	logViewer         *components.LogViewer
	describeViewer    *components.DescribeViewer
	containerSelector *components.ContainerSelector
	watchManager      *k8s.WatchManager
	width             int
	height            int
	err               error
	loading           bool
	showHelp          bool
	connected         bool
	viewMode          ViewMode
	searchMode        bool
	searchQuery       string
	refreshInterval   time.Duration
	logStreamCancel   context.CancelFunc
	logStreamActive   bool
	previousViewMode  ViewMode
	useWatchAPI       bool
}

// Message types
type resourcesLoadedMsg struct {
	resourceType components.ResourceType
	pods         []models.PodInfo
	services     []models.ServiceInfo
	deployments  []models.DeploymentInfo
	statefulSets []models.StatefulSetInfo
	events       []models.EventInfo
	err          error
}

type namespacesLoadedMsg struct {
	namespaces []string
	err        error
}

type logEntryMsg struct {
	entry   models.LogEntry
	nextCmd tea.Cmd
}

type logStreamStartedMsg struct {
	cancel context.CancelFunc
}

type logStreamErrorMsg struct {
	err error
}

type logStreamStoppedMsg struct{}

type describeLoadedMsg struct {
	data *models.DescribeData
	yaml string
	json string
}

type containersLoadedMsg struct {
	containers []string
	err        error
}

type tickMsg time.Time

type errMsg struct{ err error }

type watchEventMsg struct {
	event k8s.WatchEvent
}

type watchErrorMsg struct {
	error k8s.WatchError
}

type connectionStateMsg struct {
	state k8s.ConnectionState
}

// NewModel creates a new application model with default configuration
func NewModel(client *k8s.Client) Model {
	return NewModelWithConfig(client, config.DefaultConfig())
}

// NewModelWithConfig creates a new application model with the given configuration
func NewModelWithConfig(client *k8s.Client, cfg *config.Config) Model {
	// Get color scheme based on theme
	themeType := config.ThemeType(cfg.UI.Theme)
	colorScheme, err := config.GetColorScheme(themeType)
	if err != nil {
		// Fallback to dark theme if invalid
		colorScheme = config.DarkColorScheme()
	}

	// Apply color scheme to create styles
	styles := colorScheme.ApplyColorScheme()

	header := components.NewHeader(
		client.GetCurrentContext(),
		client.GetNamespace(),
		false, // Will be set to true after first successful load
	)

	// Create watch manager
	watchManager := k8s.NewWatchManager(client)

	// Enable watch API by default (can be overridden by config or flag later)
	useWatchAPI := true

	return Model{
		client:            client,
		config:            cfg,
		styles:            styles,
		keyMap:            keys.DefaultKeyMap(),
		header:            header,
		footer:            components.NewFooter(),
		tabs:              components.NewTabs(),
		resourceList:      components.NewResourceList(components.ResourceTypePod),
		detailView:        components.NewDetailView(),
		namespaceSelector: components.NewSelector("Select Namespace"),
		logViewer:         nil, // Created on demand
		describeViewer:    components.NewDescribeViewer(),
		containerSelector: nil, // Created on demand
		watchManager:      watchManager,
		connected:         false,
		loading:           true,
		viewMode:          ViewModeList,
		searchMode:        false,
		refreshInterval:   cfg.GetRefreshInterval(),
		logStreamActive:   false,
		previousViewMode:  ViewModeList,
		useWatchAPI:       useWatchAPI,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	if m.useWatchAPI {
		// Start watch-based updates
		return tea.Batch(
			m.startWatchMode(),
			m.waitForWatchEvents(),
		)
	}

	// Fallback to polling mode
	return tea.Batch(
		m.loadResources(),
		m.tickCmd(),
	)
}

// Update handles messages and updates the model
//
//nolint:gocyclo,funlen // Complex state machine, acceptable for main update function
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
		selectorWidth := minInt(m.width-10, 50)
		selectorHeight := minInt(m.height-6, 20)
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
			case components.ResourceTypeEvent:
				m.resourceList.SetEvents(msg.events)
			}
		}
		m.header.SetConnected(m.connected)

	case logStreamStartedMsg:
		m.logStreamActive = true
		m.logStreamCancel = msg.cancel

	case logEntryMsg:
		if m.logViewer != nil {
			m.logViewer.AddLogEntry(msg.entry)
		}
		// Chain to next log entry if stream is active
		if m.logStreamActive && msg.nextCmd != nil {
			return m, msg.nextCmd
		}

	case logStreamErrorMsg:
		m.logStreamActive = false
		if m.logStreamCancel != nil {
			m.logStreamCancel()
		}
		m.err = msg.err
		m.viewMode = m.previousViewMode

	case logStreamStoppedMsg:
		m.logStreamActive = false

	case describeLoadedMsg:
		m.describeViewer.SetData(msg.data)
		m.describeViewer.SetYAML(msg.yaml)
		m.describeViewer.SetJSON(msg.json)

	case containersLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.viewMode = m.previousViewMode
		} else if len(msg.containers) == 1 {
			// Single container, go directly to logs
			pod := m.resourceList.GetSelectedPod()
			if pod != nil {
				m.logViewer = components.NewLogViewer(pod.Name, msg.containers[0])
				m.logViewer.SetSize(m.width, m.height-6)
				m.viewMode = ViewModeLogStream
				return m, m.startLogStream(msg.containers[0])
			}
		} else {
			// Multiple containers, show selector
			m.containerSelector = components.NewContainerSelector(msg.containers)
			m.containerSelector.Show()
			m.viewMode = ViewModeContainerSelect
		}

	case namespacesLoadedMsg:
		if msg.err == nil && len(msg.namespaces) > 0 {
			m.namespaceSelector.SetOptions(msg.namespaces)
		}

	case watchEventMsg:
		// Handle watch events (ADDED, MODIFIED, DELETED)
		m.handleWatchEvent(msg.event)
		// Continue waiting for more events
		return m, m.waitForWatchEvents()

	case watchErrorMsg:
		// Handle watch errors
		if msg.error.Fatal {
			m.err = msg.error.Err
			m.connected = false
		}
		// Continue waiting for more events even on non-fatal errors
		return m, m.waitForWatchEvents()

	case connectionStateMsg:
		// Update connection state in header
		m.updateConnectionState(msg.state)
		return m, nil

	case tickMsg:
		// In watch mode, tick is used for connection state updates
		if m.useWatchAPI {
			return m, tea.Batch(
				m.checkConnectionState(),
				m.tickCmd(),
			)
		}
		// In polling mode, tick triggers resource refresh
		return m, tea.Batch(
			m.loadResources(),
			m.tickCmd(),
		)

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
//
//nolint:gocyclo,funlen // Handles many keyboard commands, complexity and length are acceptable
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if 'q' should act as Back in special view modes (not Quit)
	if msg.String() == "q" {
		switch m.viewMode {
		case ViewModeLogStream:
			// Stop log streaming
			if m.logStreamCancel != nil {
				m.logStreamCancel()
			}
			m.viewMode = m.previousViewMode
			return m, nil
		case ViewModeDescribe:
			m.viewMode = m.previousViewMode
			return m, nil
		case ViewModeContainerSelect:
			if m.containerSelector != nil {
				m.containerSelector.Hide()
			}
			m.viewMode = m.previousViewMode
			return m, nil
		}
	}

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
		// Enter detail view (but not if we're in special modes that handle Enter themselves)
		if m.viewMode == ViewModeList {
			m.viewMode = ViewModeDetail
			return m, nil
		}
		// Don't handle Enter here for other view modes - let them handle it
		// (ViewModeContainerSelect, ViewModeDescribe, etc.)

	case key.Matches(msg, m.keyMap.Back):
		// Handle back based on view mode
		switch m.viewMode {
		case ViewModeDetail:
			m.viewMode = ViewModeList
			return m, nil
		case ViewModeLogStream:
			// Stop log streaming
			if m.logStreamCancel != nil {
				m.logStreamCancel()
			}
			m.viewMode = m.previousViewMode
			return m, nil
		case ViewModeDescribe:
			m.viewMode = m.previousViewMode
			return m, nil
		}

	case key.Matches(msg, m.keyMap.Logs):
		// View logs for selected pod
		if m.viewMode == ViewModeList || m.viewMode == ViewModeDetail {
			if m.tabs.GetActiveTab() == int(components.ResourceTypePod) {
				pod := m.resourceList.GetSelectedPod()
				if pod != nil {
					m.previousViewMode = m.viewMode
					return m, m.loadContainers(pod.Namespace, pod.Name)
				}
			}
		}

	case key.Matches(msg, m.keyMap.Describe):
		// Show describe view for selected resource
		if m.viewMode == ViewModeDetail {
			m.previousViewMode = m.viewMode
			m.viewMode = ViewModeDescribe
			return m, m.loadDescribe()
		}

	case key.Matches(msg, m.keyMap.Events):
		// Jump to Events tab
		m.tabs.SetActiveTab(4)
		m.resourceList.SetResourceType(components.ResourceTypeEvent)
		m.viewMode = ViewModeList
		m.loading = true
		return m, m.loadResources()
	}

	// Handle view-specific keys
	switch m.viewMode {
	case ViewModeLogStream:
		return m.handleLogViewerKeys(msg)
	case ViewModeDescribe:
		return m.handleDescribeViewerKeys(msg)
	case ViewModeContainerSelect:
		return m.handleContainerSelectorKeys(msg)
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

				// If using watch mode, restart watchers with new namespace
				if m.useWatchAPI {
					// Restart watch manager with new namespace
					if err := m.watchManager.UpdateNamespace(selectedNS); err != nil {
						m.err = fmt.Errorf("failed to switch namespace: %w", err)
					}
					return m, nil
				}

				// Otherwise use polling mode to reload resources
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
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
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
			m.searchQuery += string(keyMsg.Runes)
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

	// Show container selector if visible
	if m.viewMode == ViewModeContainerSelect && m.containerSelector != nil && m.containerSelector.IsVisible() {
		return m.viewContainerSelector()
	}

	// Show help if requested
	if m.showHelp {
		return m.viewHelp()
	}

	// Show error if present
	if m.err != nil {
		return m.viewError()
	}

	// Handle special view modes
	switch m.viewMode {
	case ViewModeLogStream:
		if m.logViewer != nil {
			return m.logViewer.View()
		}
		return "Log viewer not initialized"

	case ViewModeDescribe:
		return m.describeViewer.View()
	}

	// Build the main view for list and detail modes
	return m.renderMainView()
}

// renderMainView renders the main view with header, tabs, content, and footer
func (m Model) renderMainView() string {
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

	case components.ResourceTypeEvent:
		event := m.resourceList.GetSelectedEvent()
		return m.detailView.ViewEvent(event)

	default:
		return "Unknown resource type"
	}
}

// viewNamespaceSelector renders the namespace selector
func (m Model) viewNamespaceSelector() string {
	// Render the selector centered on screen
	selector := m.namespaceSelector.View()

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		selector,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}),
		lipgloss.WithWhitespaceBackground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}),
	)
}

// viewContainerSelector renders the container selector
func (m Model) viewContainerSelector() string {
	// Render the selector centered on screen
	selector := m.containerSelector.View()

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		selector,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}),
		lipgloss.WithWhitespaceBackground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}),
	)
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
			msg.pods, msg.err = m.loadPods(ctx, namespace)
		case components.ResourceTypeService:
			msg.services, msg.err = m.loadServices(ctx, namespace)
		case components.ResourceTypeDeployment:
			msg.deployments, msg.err = m.loadDeployments(ctx, namespace)
		case components.ResourceTypeStatefulSet:
			msg.statefulSets, msg.err = m.loadStatefulSets(ctx, namespace)
		case components.ResourceTypeEvent:
			msg.events, msg.err = m.loadEvents(ctx, namespace)
		}

		return msg
	}
}

func (m Model) loadPods(ctx context.Context, namespace string) ([]models.PodInfo, error) {
	podList, err := m.client.GetPods(ctx, namespace)
	if err != nil {
		return nil, err
	}
	pods := make([]models.PodInfo, len(podList.Items))
	for i, pod := range podList.Items {
		pods[i] = models.NewPodInfo(&pod)
	}
	return pods, nil
}

func (m Model) loadServices(ctx context.Context, namespace string) ([]models.ServiceInfo, error) {
	serviceList, err := m.client.GetServices(ctx, namespace)
	if err != nil {
		return nil, err
	}
	services := make([]models.ServiceInfo, len(serviceList.Items))
	for i, svc := range serviceList.Items {
		services[i] = models.NewServiceInfo(&svc)
	}
	return services, nil
}

func (m Model) loadDeployments(ctx context.Context, namespace string) ([]models.DeploymentInfo, error) {
	deploymentList, err := m.client.GetDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}
	deployments := make([]models.DeploymentInfo, len(deploymentList.Items))
	for i, dep := range deploymentList.Items {
		deployments[i] = models.NewDeploymentInfo(&dep)
	}
	return deployments, nil
}

func (m Model) loadStatefulSets(ctx context.Context, namespace string) ([]models.StatefulSetInfo, error) {
	statefulSetList, err := m.client.GetStatefulSets(ctx, namespace)
	if err != nil {
		return nil, err
	}
	statefulSets := make([]models.StatefulSetInfo, len(statefulSetList.Items))
	for i, sts := range statefulSetList.Items {
		statefulSets[i] = models.NewStatefulSetInfo(&sts)
	}
	return statefulSets, nil
}

func (m Model) loadEvents(ctx context.Context, namespace string) ([]models.EventInfo, error) {
	eventList, err := m.client.GetEvents(ctx, namespace)
	if err != nil {
		return nil, err
	}
	events := make([]models.EventInfo, len(eventList.Items))
	for i, evt := range eventList.Items {
		events[i] = models.NewEventInfo(&evt)
	}
	return events, nil
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

// tickCmd creates a tick command for auto-refresh using configured interval
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// loadContainers fetches containers for a pod
func (m Model) loadContainers(namespace, podName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		containers, err := m.client.GetPodContainers(ctx, namespace, podName)
		if err != nil {
			return containersLoadedMsg{err: err}
		}

		return containersLoadedMsg{containers: containers}
	}
}

// startLogStream initiates log streaming for a pod container
func (m Model) startLogStream(containerName string) tea.Cmd {
	pod := m.resourceList.GetSelectedPod()
	if pod == nil {
		return func() tea.Msg {
			return logStreamErrorMsg{err: fmt.Errorf("no pod selected")}
		}
	}

	// streamLogs will handle sending logStreamStartedMsg and starting the read chain
	return m.streamLogs(pod.Namespace, pod.Name, containerName)
}

// streamLogs streams logs from a pod container
func (m Model) streamLogs(namespace, podName, containerName string) tea.Cmd {
	// Create context with cancel for this stream
	ctx, cancel := context.WithCancel(context.Background())

	// Get log options
	opts := models.DefaultLogOptions()
	opts.Container = containerName

	// Start streaming
	logChan, errChan := m.client.GetPodLogsStream(ctx, namespace, podName, containerName, opts)

	// Create recursive reader function
	var readNext func() tea.Cmd
	readNext = func() tea.Cmd {
		return func() tea.Msg {
			select {
			case entry, ok := <-logChan:
				if !ok {
					return logStreamStoppedMsg{}
				}
				// Return entry with command for next read
				return logEntryMsg{
					entry:   entry,
					nextCmd: readNext(),
				}
			case err, ok := <-errChan:
				if ok && err != nil {
					return logStreamErrorMsg{err: err}
				}
				return logStreamStoppedMsg{}
			case <-ctx.Done():
				return logStreamStoppedMsg{}
			}
		}
	}

	// Return batch: first send started message with cancel, then start reading
	return tea.Batch(
		func() tea.Msg {
			return logStreamStartedMsg{cancel: cancel}
		},
		readNext(),
	)
}

// loadDescribe loads describe data for the selected resource
func (m Model) loadDescribe() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var data *models.DescribeData
		var yaml, json string
		var err error

		resourceType := components.ResourceType(m.tabs.GetActiveTab())

		switch resourceType {
		case components.ResourceTypePod:
			pod := m.resourceList.GetSelectedPod()
			if pod != nil {
				data, err = m.client.DescribePod(ctx, pod.Namespace, pod.Name)
				if err == nil {
					yaml, _ = m.client.GetResourceYAML(ctx, "Pod", pod.Namespace, pod.Name)
					json, _ = m.client.GetResourceJSON(ctx, "Pod", pod.Namespace, pod.Name)
				}
			}

		case components.ResourceTypeService:
			svc := m.resourceList.GetSelectedService()
			if svc != nil {
				data, err = m.client.DescribeService(ctx, svc.Namespace, svc.Name)
				if err == nil {
					yaml, _ = m.client.GetResourceYAML(ctx, "Service", svc.Namespace, svc.Name)
					json, _ = m.client.GetResourceJSON(ctx, "Service", svc.Namespace, svc.Name)
				}
			}

		case components.ResourceTypeDeployment:
			dep := m.resourceList.GetSelectedDeployment()
			if dep != nil {
				data, err = m.client.DescribeDeployment(ctx, dep.Namespace, dep.Name)
				if err == nil {
					yaml, _ = m.client.GetResourceYAML(ctx, "Deployment", dep.Namespace, dep.Name)
					json, _ = m.client.GetResourceJSON(ctx, "Deployment", dep.Namespace, dep.Name)
				}
			}

		case components.ResourceTypeStatefulSet:
			sts := m.resourceList.GetSelectedStatefulSet()
			if sts != nil {
				data, err = m.client.DescribeStatefulSet(ctx, sts.Namespace, sts.Name)
				if err == nil {
					yaml, _ = m.client.GetResourceYAML(ctx, "StatefulSet", sts.Namespace, sts.Name)
					json, _ = m.client.GetResourceJSON(ctx, "StatefulSet", sts.Namespace, sts.Name)
				}
			}
		}

		if err != nil {
			return errMsg{err: err}
		}

		return describeLoadedMsg{data: data, yaml: yaml, json: json}
	}
}

// handleLogViewerKeys handles key presses in log viewer mode
func (m Model) handleLogViewerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.logViewer == nil {
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keyMap.Follow):
		m.logViewer.ToggleFollow()
	case key.Matches(msg, m.keyMap.Timestamps):
		m.logViewer.ToggleTimestamps()
	case key.Matches(msg, m.keyMap.Search):
		m.logViewer.SetSearchMode(true)
	default:
		// Pass to viewport for scrolling
		var cmd tea.Cmd
		m.logViewer, cmd = m.logViewer.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleDescribeViewerKeys handles key presses in describe viewer mode
func (m Model) handleDescribeViewerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.YAML):
		m.describeViewer.SetFormat(models.FormatYAML)
	case key.Matches(msg, m.keyMap.JSON):
		m.describeViewer.SetFormat(models.FormatJSON)
	case key.Matches(msg, m.keyMap.Describe):
		m.describeViewer.SetFormat(models.FormatDescribe)
	default:
		// Pass to viewport for scrolling
		var cmd tea.Cmd
		m.describeViewer, cmd = m.describeViewer.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleContainerSelectorKeys handles key presses in container selector mode
func (m Model) handleContainerSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.containerSelector == nil {
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keyMap.Up):
		m.containerSelector.MoveUp()
	case key.Matches(msg, m.keyMap.Down):
		m.containerSelector.MoveDown()
	case key.Matches(msg, m.keyMap.Enter):
		// Get selected container and start log stream
		containerName := m.containerSelector.GetSelectedContainerName()
		if containerName != "" {
			pod := m.resourceList.GetSelectedPod()
			if pod != nil {
				// Hide the selector
				m.containerSelector.Hide()
				// Create log viewer
				m.logViewer = components.NewLogViewer(pod.Name, containerName)
				m.logViewer.SetSize(m.width, m.height-6)
				m.viewMode = ViewModeLogStream
				return m, m.startLogStream(containerName)
			}
		}
	case key.Matches(msg, m.keyMap.Back):
		m.containerSelector.Hide()
		m.viewMode = m.previousViewMode
	}

	return m, nil
}

// startWatchMode initializes and starts the watch manager
func (m Model) startWatchMode() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Start watching all resource types
		resourceTypes := []k8s.ResourceType{
			k8s.ResourceTypePod,
			k8s.ResourceTypeService,
			k8s.ResourceTypeDeployment,
			k8s.ResourceTypeStatefulSet,
			k8s.ResourceTypeEvent,
		}

		err := m.watchManager.Start(ctx, resourceTypes)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to start watch manager: %w", err)}
		}

		return nil
	}
}

// waitForWatchEvents waits for the next watch event from any resource watcher
func (m Model) waitForWatchEvents() tea.Cmd {
	return func() tea.Msg {
		select {
		case event := <-m.watchManager.GetEventChannel():
			return watchEventMsg{event: event}
		case err := <-m.watchManager.GetErrorChannel():
			return watchErrorMsg{error: err}
		}
	}
}

// handleWatchEvent processes a watch event and updates the resource list
func (m *Model) handleWatchEvent(event k8s.WatchEvent) {
	m.loading = false
	m.connected = true
	m.err = nil
	m.header.SetConnected(true)

	// Convert k8s.ResourceType to components.ResourceType
	// They have the same values but different types
	componentResourceType := components.ResourceType(event.ResourceType)

	switch event.EventType {
	case "ADDED":
		m.handleResourceAdded(componentResourceType, event.Object)
	case "MODIFIED":
		m.handleResourceModified(componentResourceType, event.Object)
	case "DELETED":
		m.handleResourceDeleted(componentResourceType, event.Object)
	}
}

// handleResourceAdded adds or updates a resource in the list
func (m *Model) handleResourceAdded(resourceType components.ResourceType, obj interface{}) {
	switch resourceType {
	case components.ResourceTypePod:
		if pod, ok := obj.(*corev1.Pod); ok {
			podInfo := models.NewPodInfo(pod)
			m.resourceList.AddOrUpdatePod(podInfo)
		}
	case components.ResourceTypeService:
		if svc, ok := obj.(*corev1.Service); ok {
			svcInfo := models.NewServiceInfo(svc)
			m.resourceList.AddOrUpdateService(svcInfo)
		}
	case components.ResourceTypeDeployment:
		if dep, ok := obj.(*appsv1.Deployment); ok {
			depInfo := models.NewDeploymentInfo(dep)
			m.resourceList.AddOrUpdateDeployment(depInfo)
		}
	case components.ResourceTypeStatefulSet:
		if sts, ok := obj.(*appsv1.StatefulSet); ok {
			stsInfo := models.NewStatefulSetInfo(sts)
			m.resourceList.AddOrUpdateStatefulSet(stsInfo)
		}
	case components.ResourceTypeEvent:
		if evt, ok := obj.(*corev1.Event); ok {
			evtInfo := models.NewEventInfo(evt)
			m.resourceList.AddOrUpdateEvent(evtInfo)
		}
	}
}

// handleResourceModified updates a resource in the list
func (m *Model) handleResourceModified(resourceType components.ResourceType, obj interface{}) {
	// Modified is handled the same as added for our use case
	m.handleResourceAdded(resourceType, obj)
}

// handleResourceDeleted removes a resource from the list
func (m *Model) handleResourceDeleted(resourceType components.ResourceType, obj interface{}) {
	switch resourceType {
	case components.ResourceTypePod:
		if pod, ok := obj.(*corev1.Pod); ok {
			m.resourceList.RemovePod(pod.Namespace, pod.Name)
		}
	case components.ResourceTypeService:
		if svc, ok := obj.(*corev1.Service); ok {
			m.resourceList.RemoveService(svc.Namespace, svc.Name)
		}
	case components.ResourceTypeDeployment:
		if dep, ok := obj.(*appsv1.Deployment); ok {
			m.resourceList.RemoveDeployment(dep.Namespace, dep.Name)
		}
	case components.ResourceTypeStatefulSet:
		if sts, ok := obj.(*appsv1.StatefulSet); ok {
			m.resourceList.RemoveStatefulSet(sts.Namespace, sts.Name)
		}
	case components.ResourceTypeEvent:
		if evt, ok := obj.(*corev1.Event); ok {
			m.resourceList.RemoveEvent(evt.Namespace, evt.Name)
		}
	}
}

// checkConnectionState checks the connection state of all watchers
func (m Model) checkConnectionState() tea.Cmd {
	return func() tea.Msg {
		state := m.watchManager.GetOverallConnectionState()
		return connectionStateMsg{state: state}
	}
}

// updateConnectionState updates the UI based on connection state
func (m *Model) updateConnectionState(state k8s.ConnectionState) {
	// Map k8s.ConnectionState to components.ConnectionState
	var headerState components.ConnectionState
	switch state {
	case k8s.StateConnected:
		m.connected = true
		headerState = components.ConnectionStateConnected
	case k8s.StateConnecting:
		m.connected = true
		headerState = components.ConnectionStateConnecting
	case k8s.StateReconnecting:
		m.connected = true
		headerState = components.ConnectionStateReconnecting
	case k8s.StateError:
		m.connected = false
		headerState = components.ConnectionStateError
	case k8s.StateDisconnected:
		m.connected = false
		headerState = components.ConnectionStateDisconnected
	}
	m.header.SetConnectionState(headerState)
}
