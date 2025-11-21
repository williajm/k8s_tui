# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

K8S-TUI is a fast, keyboard-driven terminal user interface for Kubernetes cluster management built with Go. It uses the Bubble Tea TUI framework for rendering and the official Kubernetes client-go library for cluster interaction.

## Build, Test, and Lint Commands

### Building
```bash
# Build the main binary
go build -o k8s-tui cmd/k8s-tui/main.go

# Build for specific platform
go build -v -o k8s-tui.exe ./cmd/k8s-tui  # Windows
go build -v -o k8s-tui ./cmd/k8s-tui     # Linux/macOS

# Run directly
go run cmd/k8s-tui/main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector (important for concurrency bugs)
go test -race ./...

# Run with coverage in atomic mode (CI uses this)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run specific package tests
go test ./internal/models -v
go test ./internal/ui/components -v
go test ./internal/ui/styles -v
```

### Linting
```bash
# Run golangci-lint (what CI uses)
golangci-lint run ./...

# Run with timeout (CI uses 5m)
golangci-lint run --timeout=5m ./...

# Format code
go fmt ./...
gofmt -w -s .
```

### Dependency Management
```bash
# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
go mod tidy
```

## Architecture

### Application Structure

The codebase follows a clean architecture pattern:

- **cmd/k8s-tui/**: Application entry point using Cobra for CLI handling
- **internal/app/**: Core Bubble Tea application model (Model-Update-View pattern)
- **internal/k8s/**: Kubernetes client wrapper with connection management
- **internal/models/**: Data models for resources (PodInfo, ServiceInfo, DeploymentInfo, StatefulSetInfo)
- **internal/ui/**: UI layer separated into:
  - **components/**: Reusable UI components (Header, Footer, ResourceList, Tabs, Selector, DetailView, LogViewer, DescribeViewer, ContainerSelector)
  - **keys/**: Keyboard bindings
  - **styles/**: Lipgloss theme and styling

### Bubble Tea Architecture

The app uses the Elm Architecture (TEA) pattern:

1. **Model** (`app.Model`): Holds application state including:
   - UI components (header, footer, tabs, resourceList, detailView, namespaceSelector)
   - View state (viewMode, searchMode, loading, connected)
   - Kubernetes client and current namespace/context

2. **Update** (`app.Update`): Processes messages and returns new model + commands:
   - Routes to specialized handlers: `handleKeyPress`, `handleNamespaceSelector`, `handleSearchMode`
   - Handles resource loading messages per resource type
   - Processes watch events for real-time updates (with fallback to 5-second polling)

3. **View** (`app.View`): Renders the current state:
   - Switches between list view and detail view based on `viewMode`
   - Overlays namespace selector when visible
   - Shows help screen or error messages as needed

**Key Message types**:
- `resourcesLoadedMsg`: Contains loaded resources for a specific ResourceType (pods, services, deployments, statefulsets)
- `watchEventMsg`: Watch API event (ADDED, MODIFIED, DELETED) for real-time updates
- `watchErrorMsg`: Watch connection errors and reconnection notifications
- `connectionStateMsg`: Connection state changes (Connected, Reconnecting, Disconnected, Error)
- `namespacesLoadedMsg`: List of available namespaces for selector
- `tickMsg`: Timer for fallback polling mode (5-second interval when watch unavailable)
- `tea.KeyMsg`: Keyboard input
- `tea.WindowSizeMsg`: Terminal resize events

### ViewMode State Management

The app has two primary view modes:
- **ViewModeList**: Shows the resource list (default view)
- **ViewModeDetail**: Shows detailed information for the selected resource

Pressing Enter switches to detail view, Esc/Back returns to list view. View mode resets when switching tabs.

### Kubernetes Client Pattern

The `k8s.Client` wraps `kubernetes.Clientset` and provides:
- Kubeconfig loading priority: in-cluster → KUBECONFIG env → ~/.kube/config
- Context and namespace management (mutable via SetNamespace)
- Connection testing with timeout
- Resource fetching methods for multiple resource types:
  - Pods: `GetPods`, `GetAllPods`, `GetPod`, `GetPodLogs`, `StreamPodLogs`, `WatchPods`
  - Services: `GetServices`, `GetAllServices`, `GetService`, `WatchServices`
  - Deployments: `GetDeployments`, `GetAllDeployments`, `GetDeployment`, `WatchDeployments`
  - StatefulSets: `GetStatefulSets`, `GetAllStatefulSets`, `GetStatefulSet`, `WatchStatefulSets`
  - Events: `GetEvents`, `GetAllEvents`, `GetEventsForResource`, `WatchEvents`
  - Describe: `DescribePod`, `DescribeService`, `DescribeDeployment`, `DescribeStatefulSet`
  - Namespaces: `GetNamespaces`
- Watch API infrastructure:
  - `WatchManager`: Orchestrates multiple resource watchers
  - `ResourceWatcher`: Manages watch streams for individual resource types
  - `ExponentialBackoff`: Reconnection strategy with jitter

### Resource Models

Each resource type has a corresponding `*Info` model in `internal/models/`:
- **PodInfo**: Includes containers array, ready status, restart counts
- **ServiceInfo**: Includes type, ClusterIP, ExternalIP, ports, selectors
- **DeploymentInfo**: Includes replicas, ready, up-to-date, available counts
- **StatefulSetInfo**: Includes replicas, ready count, update strategy
- **EventInfo**: Includes type, reason, message, timestamp, involved object
- **LogEntry**: Includes timestamp, level, message with automatic log level detection

All resource models provide:
- `GetStatusSymbol()`: Returns visual indicator (✓, ✗, ○, ⚠, ⊗)
- `formatAge()`: Converts timestamp to human-readable age (e.g., "5m", "2h", "3d")

### UI Component Design

**ResourceList** (generic component):
- Maintains state for all resource types in separate slices
- Switches display mode via `ResourceType` enum (Pod, Service, Deployment, StatefulSet, Event)
- Handles navigation (up/down, page up/down, home/end)
- Supports search filtering
- Renders appropriate table columns per resource type
- Diff-based updates for watch events: `AddOrUpdatePod()`, `RemovePod()`, etc. for all resource types
- Preserves cursor position and selection during incremental updates

**Tabs**:
- Manages active tab selection (0=Pods, 1=Services, 2=Deployments, 3=StatefulSets, 4=Events)
- Provides `NextTab()` and `PrevTab()` for keyboard navigation
- Renders with active/inactive styling

**Selector** (namespace selector):
- Modal-style component for choosing from a list of options
- Shows/hides via `IsVisible()` state
- Supports up/down navigation within options
- Returns selected value via `GetSelected()`

**ContainerSelector**:
- Modal dialog for selecting containers in multi-container pods
- Shows init containers with differentiation
- Keyboard navigation (up/down/enter/esc)
- Returns selected container name

**DetailView**:
- Renders detailed information for selected resource
- Type-specific methods: `ViewPod()`, `ViewService()`, `ViewDeployment()`, `ViewStatefulSet()`
- Displays formatted key-value pairs using `styles.RenderDetailRow()`

**LogViewer**:
- Real-time log streaming with follow mode
- Search/filter within logs
- Timestamp toggle
- Log level detection and color coding
- Circular buffer (10,000 lines) for memory efficiency
- Previous container logs support

**DescribeViewer**:
- Multiple format support (Describe, YAML, JSON)
- Format switching with keyboard shortcuts (d/y/j)
- Structured describe output with sections
- Scrollable viewport for large resources

### Component Interaction Pattern

The main app model (`app.Model`) orchestrates all components:
1. Tabs component determines which `ResourceType` is active
2. ResourceList displays resources for the current type
3. WatchManager sends real-time updates via watch events (when enabled)
4. App processes watch events and calls ResourceList diff-based update methods
5. When Enter is pressed, app switches to `ViewModeDetail`
6. DetailView renders the selected resource from ResourceList
7. Namespace selector overlays everything when visible (triggered by 'n' key)

This pattern avoids tight coupling—components don't know about each other, only the main model coordinates them. Watch mode operates transparently; the UI receives updates via message passing regardless of whether they come from watch events or polling.

## Development Workflow

### Branch Strategy
- **main**: Protected branch requiring PR and passing CI
- **dev**: Experimental work branch for ongoing development
- **feature/*** or **hotfix/***: Feature branches for PRs

### Before Creating PR
```bash
# Always run these locally first
go test ./...
go fmt ./...
golangci-lint run --timeout=5m ./...
```

### CI Requirements
All PRs must pass:
- **Test**: Go 1.25 on Ubuntu, macOS, Windows with race detector
- **Lint**: golangci-lint with 5m timeout
- **Build**: Cross-platform build verification
- **Coverage**: Minimum 30% test coverage threshold (enforced on Ubuntu only)

The CI configuration uses:
- Test coverage reporting to Codecov (Ubuntu + Go 1.25 only)
- Artifact uploads for built binaries (7-day retention)
- Parallel matrix strategy for efficiency

### Commit Message Convention
Follow Conventional Commits format:
```
<type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, perf, test, chore, ci
Examples:
  feat(ui): Add pod log streaming view
  fix(k8s): Handle connection timeout gracefully
  test: Add unit tests for pod list component
```

## Linting Configuration

The project uses golangci-lint with specific settings:
- **Line length**: 140 characters
- **Function length**: 100 lines, 50 statements
- **Complexity**: Max cyclomatic complexity of 15
- **Import ordering**: Local prefix `github.com/williajm/k8s-tui`
- **Test exclusions**: funlen and dupl disabled for `_test.go` files

Key enabled linters: bodyclose, errcheck, gosec, gosimple, ineffassign, staticcheck, unused, revive

## Running the Application

```bash
# Use default kubeconfig
./k8s-tui

# Specify custom kubeconfig
./k8s-tui --kubeconfig ~/.kube/custom-config

# Start in specific namespace
./k8s-tui --namespace production
./k8s-tui -n production

# Use specific context
./k8s-tui --context staging-cluster
```

## Key Dependencies

- **Bubble Tea** (v1.3.10): TUI framework for Model-Update-View architecture
- **Lipgloss** (v1.1.0): Terminal styling and layout
- **Bubbles** (v0.21.0): Pre-built TUI components
- **Cobra** (v1.10.1): CLI argument parsing
- **client-go** (v0.34.1): Official Kubernetes Go client
- **k8s.io/api** (v0.34.1): Kubernetes API types
- **k8s.io/apimachinery** (v0.34.1): Kubernetes meta types

## Development Roadmap

### Phase 1 - Foundation (v0.1.0) ✅ **MERGED TO MAIN**
**Status**: Complete and in production on `main` branch

Features:
- ✅ Basic Bubble Tea TUI framework setup
- ✅ Kubernetes client integration with kubeconfig support
- ✅ Pod listing and navigation (up/down, page up/down, home/end)
- ✅ Basic status indicators and formatting
- ✅ Header and footer components
- ✅ Keyboard shortcuts and help screen
- ✅ CI/CD pipeline (test, lint, build on multiple platforms)
- ✅ Unit testing infrastructure with race detector
- ✅ golangci-lint configuration
- ✅ Cross-platform build support (Linux, macOS, Windows)
- ✅ Codecov integration for test coverage

**Branch**: Merged via PR #1 (`feature/phase1-foundation` → `main`)

---

### Phase 2 - Core Features (v0.2.0) ✅ **MERGED TO MAIN**
**Status**: Complete and in production on `main` branch

Features:
- ✅ Multi-resource support (Pods, Services, Deployments, StatefulSets)
- ✅ Generic ResourceList component for all resource types
- ✅ Tab navigation between resource types (Tab/Shift+Tab, 1-4 keys)
- ✅ Detail views for all resources (Enter to view, Esc to return)
- ✅ Namespace switching with selector dialog ('n' key)
- ✅ Search/filter functionality ('/' key for search mode)
- ✅ ViewMode state management (list vs detail view)
- ✅ Real-time updates with 5-second auto-refresh
- ✅ Resource-specific data models with status symbols
- ✅ Component architecture (Tabs, Selector, DetailView)

**Branch**: Merged to `main` branch

---

### Phase 3 - Observability & Logs (v0.3.0) ✅ **MERGED TO MAIN**
**Status**: Complete and in production on `main` branch

**Goal**: Add log viewing, events, and resource inspection capabilities

Features:
- ✅ Pod log streaming view ('l' key from pod list/detail)
  - ✅ Container selection for multi-container pods
  - ✅ Follow mode (live streaming)
  - ✅ Log filtering/search within logs
  - ✅ Timestamp toggle
  - ✅ Log level detection and color coding
  - ✅ Circular buffer (10,000 lines) for memory efficiency
  - ✅ Previous container logs support
- ✅ Kubernetes events display
  - ✅ Event list view (new tab - 5th tab)
  - ✅ Event type filtering (Normal, Warning, Error)
  - ✅ Time-based sorting with age display
  - ✅ Event type symbols and color coding
  - ✅ Resource-specific event support
- ✅ Describe functionality ('d' key from detail view)
  - ✅ Multiple format support (Describe, YAML, JSON)
  - ✅ Format switching with keyboard shortcuts (d/y/j)
  - ✅ Structured describe output with sections
  - ✅ Full resource inspection capabilities
  - ✅ Scrollable viewport for large resources

**UI Improvements**:
- ✅ Comprehensive footer with all keyboard shortcuts (10+ shortcuts across 2 lines)
- ✅ View-specific footers (log viewer, describe viewer) with context-aware shortcuts
- ✅ Standardized 'q' behavior: quit in main views, back in special views
- ✅ Consistent [Esc]/[Backspace] for back navigation

**Bug Fixes**:
- ✅ Fixed log viewer footer scrolling issue
- ✅ Fixed 'q' key behavior in special views
- ✅ Improved namespace selector overlay rendering
- ✅ Fixed golangci-lint configuration warnings

**Technical Implementation**:
- Added `internal/k8s/logs.go` - Log streaming with context management
- Added `internal/k8s/events.go` - Event fetching and filtering
- Added `internal/k8s/describe.go` - Resource describe operations
- Added `internal/models/log.go` - Log entry models with level detection
- Added `internal/models/event.go` - Event info models
- Added `internal/models/describe.go` - Describe data structures
- Added `internal/ui/components/logviewer.go` - Log viewer component
- Added `internal/ui/components/describe.go` - Describe viewer component
- Added `internal/ui/components/container_selector.go` - Container selection modal
- Comprehensive unit tests for all new components (51 new tests)

**Branch**: Merged to `main` branch

---

### Phase 4 - Real-time Watch & Performance (v0.4.0) ✅ **COMPLETE**
**Status**: Implementation complete on `feature/phase4-watch-performance` branch

**Goal**: Replace polling with efficient Kubernetes Watch API and improve performance

Implemented Features:
- ✅ Kubernetes Watch API integration
  - ✅ Real-time resource updates via watch streams for all resource types
  - ✅ Replace 5-second polling with event-driven updates
  - ✅ Watch reconnection on failure with exponential backoff
  - ✅ Resource version tracking and handling (including 410 Gone errors)
  - ✅ Bookmark support for resource version updates
- ✅ Performance optimizations
  - ✅ Efficient diff-based UI updates (AddOrUpdate/Remove methods)
  - ✅ Event-driven rendering (only updates when resources change)
  - 🔄 Lazy loading for large resource lists (deferred to future)
  - 🔄 Virtual scrolling for 1000+ items (deferred to future)
  - ✅ Memory-efficient watch stream management
- ✅ Connection management
  - ✅ Comprehensive connection error handling
  - ✅ Auto-reconnect with exponential backoff (1s → 2s → 4s → 8s → 16s → 30s max)
  - ✅ Enhanced cluster connection status indicator (Connected, Connecting, Reconnecting, Error, Disconnected)
  - ✅ Namespace switching properly restarts watchers

**Technical Implementation**:
- Added `internal/k8s/backoff.go` - Exponential backoff with jitter for reconnection
- Added `internal/k8s/watch.go` - Low-level watch API wrappers for all resource types
- Added `internal/k8s/resource_watcher.go` - Single resource type watcher with state management
- Added `internal/k8s/watch_manager.go` - Multi-resource watch orchestration
- Enhanced `internal/ui/components/resourcelist.go` - Diff-based update methods (AddOrUpdate*/Remove*)
- Enhanced `internal/ui/components/header.go` - Detailed connection states (5 states)
- Enhanced `internal/app/app.go` - Watch mode integration with fallback to polling
- Comprehensive unit tests for all watch components (35+ new tests)
- Full golangci-lint compliance

**Branch**: `feature/phase4-watch-performance`

**Testing**:
- ✅ All 287+ unit tests passing with race detector
- ✅ Golangci-lint clean (no warnings or errors)
- ✅ Test coverage maintained at 60%+ overall
- ✅ Manual testing completed with real Kubernetes cluster
- ✅ Comprehensive test report in PHASE4_TEST_REPORT.md

**Notes**:
- Virtual scrolling and lazy loading deferred to future optimization phase
- Watch API provides significant performance improvement over 5-second polling
- Memory usage is efficient with proper cleanup on namespace/tab switches
- Backward compatible - falls back to polling if watch mode disabled

---

### Phase 5 - Configuration & Customization (v0.5.0) 📋 **PLANNED**
**Status**: Not started

**Goal**: Add persistent configuration and theme customization

Planned Features:
- [ ] Configuration file support (`~/.config/k8s-tui/config.yaml`)
  - [ ] Default namespace preference
  - [ ] Default context preference
  - [ ] Refresh interval configuration
  - [ ] Keyboard shortcut customization
  - [ ] Column visibility/ordering
- [ ] Theme system
  - [ ] Multiple built-in themes (dark, light, high-contrast)
  - [ ] Custom color schemes
  - [ ] Theme preview and switching ('t' key)
  - [ ] Per-resource-type color customization
- [ ] UI preferences
  - [ ] Timestamp format options (relative vs absolute)
  - [ ] Table layout preferences
  - [ ] Font/Unicode symbol fallbacks

**Branch**: Will be developed on `dev` branch

---

### Phase 6 - Additional Resources (v0.6.0) 📋 **PLANNED**
**Status**: Not started

**Goal**: Support more Kubernetes resource types

Planned Features:
- [ ] ConfigMaps and Secrets (read-only view, no secret values)
- [ ] Jobs and CronJobs
- [ ] DaemonSets and ReplicaSets
- [ ] Ingresses and NetworkPolicies
- [ ] PersistentVolumes and PersistentVolumeClaims
- [ ] Nodes (cluster-level view)
- [ ] Resource filtering by labels/annotations

**Branch**: Will be developed on `dev` branch

---

### Phase 7 - Write Operations (v0.7.0) 🔒 **FUTURE**
**Status**: Not started - **High-risk phase requiring careful design**

**Goal**: Add controlled write operations for resource management

**IMPORTANT**: This phase requires:
- Confirmation dialogs for all destructive operations
- Dry-run mode
- Audit logging
- Optional write-protection mode
- Extensive testing

Planned Features:
- [ ] Pod operations
  - [ ] Delete pod (with confirmation)
  - [ ] Restart pod (delete and wait for recreation)
  - [ ] Port-forward setup
- [ ] Deployment operations
  - [ ] Scale replicas up/down
  - [ ] Restart rollout
  - [ ] Pause/resume rollout
- [ ] Safety features
  - [ ] Confirmation prompts for all write operations
  - [ ] Dry-run preview mode
  - [ ] Write operation audit log
  - [ ] Read-only mode flag (`--read-only`)

**Branch**: Will be developed on `dev` branch with extra caution

---

## Current Status Summary

| Phase | Version | Status | Branch | In Main |
|-------|---------|--------|--------|---------|
| Phase 1 - Foundation | v0.1.0 | ✅ Complete | `main` | ✅ Yes |
| Phase 2 - Core Features | v0.2.0 | ✅ Complete | `main` | ✅ Yes |
| Phase 3 - Observability | v0.3.0 | ✅ Complete | `main` | ✅ Yes |
| Phase 4 - Watch API | v0.4.0 | ✅ Complete | `feature/phase4-watch-performance` | 🔄 PR Pending |
| Phase 5 - Configuration | v0.5.0 | 📋 Planned | - | ❌ No |
| Phase 6 - More Resources | v0.6.0 | 📋 Planned | - | ❌ No |
| Phase 7 - Write Ops | v0.7.0 | 📋 Planned | - | ❌ No |

**Current Focus**: Phase 4 complete and tested, ready for PR to main. Phase 5 or 6 next.

**Performance Impact**:
- Network traffic reduction: ~80-90% compared to polling
- UI responsiveness: Instant updates (sub-second) instead of up to 5-second delay
- Memory efficiency: Maintained with proper watch stream cleanup
- Backward compatible: Falls back to polling if watch mode disabled or unavailable
