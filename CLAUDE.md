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
  - **components/**: Reusable UI components (Header, Footer, ResourceList, Tabs, Selector, DetailView)
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
   - Manages auto-refresh with 5-second tick

3. **View** (`app.View`): Renders the current state:
   - Switches between list view and detail view based on `viewMode`
   - Overlays namespace selector when visible
   - Shows help screen or error messages as needed

**Key Message types**:
- `resourcesLoadedMsg`: Contains loaded resources for a specific ResourceType (pods, services, deployments, statefulsets)
- `namespacesLoadedMsg`: List of available namespaces for selector
- `tickMsg`: Timer for 5-second auto-refresh
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
  - Pods: `GetPods`, `GetAllPods`, `GetPod`, `GetPodLogs`
  - Services: `GetServices`, `GetAllServices`, `GetService`
  - Deployments: `GetDeployments`, `GetAllDeployments`, `GetDeployment`
  - StatefulSets: `GetStatefulSets`, `GetAllStatefulSets`, `GetStatefulSet`
  - Namespaces: `GetNamespaces`

### Resource Models

Each resource type has a corresponding `*Info` model in `internal/models/`:
- **PodInfo**: Includes containers array, ready status, restart counts
- **ServiceInfo**: Includes type, ClusterIP, ExternalIP, ports, selectors
- **DeploymentInfo**: Includes replicas, ready, up-to-date, available counts
- **StatefulSetInfo**: Includes replicas, ready count, update strategy

All models provide:
- `GetStatusSymbol()`: Returns visual indicator (✓, ✗, ○, ⚠, ⊗)
- `formatAge()`: Converts timestamp to human-readable age (e.g., "5m", "2h", "3d")

### UI Component Design

**ResourceList** (generic component):
- Maintains state for all resource types in separate slices
- Switches display mode via `ResourceType` enum (Pod, Service, Deployment, StatefulSet)
- Handles navigation (up/down, page up/down, home/end)
- Supports search filtering
- Renders appropriate table columns per resource type

**Tabs**:
- Manages active tab selection (0=Pods, 1=Services, 2=Deployments, 3=StatefulSets)
- Provides `NextTab()` and `PrevTab()` for keyboard navigation
- Renders with active/inactive styling

**Selector** (namespace selector):
- Modal-style component for choosing from a list of options
- Shows/hides via `IsVisible()` state
- Supports up/down navigation within options
- Returns selected value via `GetSelected()`

**DetailView**:
- Renders detailed information for selected resource
- Type-specific methods: `ViewPod()`, `ViewService()`, `ViewDeployment()`, `ViewStatefulSet()`
- Displays formatted key-value pairs using `styles.RenderDetailRow()`

### Component Interaction Pattern

The main app model (`app.Model`) orchestrates all components:
1. Tabs component determines which `ResourceType` is active
2. ResourceList displays resources for the current type
3. When Enter is pressed, app switches to `ViewModeDetail`
4. DetailView renders the selected resource from ResourceList
5. Namespace selector overlays everything when visible (triggered by 'n' key)

This pattern avoids tight coupling—components don't know about each other, only the main model coordinates them.

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
- **Test**: Go 1.21, 1.22, 1.23 on Ubuntu, macOS, Windows with race detector
- **Lint**: golangci-lint with 5m timeout
- **Build**: Cross-platform build verification

The CI configuration uses:
- Test coverage reporting to Codecov (Ubuntu + Go 1.23 only)
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

### Phase 2 - Core Features (v0.2.0) ✅ **READY FOR PR** 🚀
**Status**: Complete on `dev` branch, awaiting merge to `main`

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

**Branch**: Currently on `dev` (commit d9bb503), needs PR to `main`

**Next Steps**:
1. Run linters and tests before creating PR
2. Create PR: `dev` → `main`
3. Merge after CI passes

---

### Phase 3 - Observability & Logs (v0.3.0) 📋 **PLANNED**
**Status**: Not started

**Goal**: Add log viewing, events, and resource inspection capabilities

Planned Features:
- [ ] Pod log streaming view ('l' key from pod list/detail)
  - [ ] Container selection for multi-container pods
  - [ ] Follow mode (live streaming)
  - [ ] Tail line count configuration
  - [ ] Log filtering/search within logs
  - [ ] Previous container logs (for restarted pods)
- [ ] Kubernetes events display
  - [ ] Event list view (new tab)
  - [ ] Resource-specific events in detail view
  - [ ] Event filtering by type (Normal, Warning, Error)
  - [ ] Age and reason display
- [ ] Describe functionality ('d' key from detail view)
  - [ ] Full resource YAML/JSON view
  - [ ] Formatted describe output (like `kubectl describe`)
  - [ ] Syntax highlighting for YAML/JSON
  - [ ] Copy-to-clipboard support

**Branch**: Will be developed on `dev` branch

---

### Phase 4 - Real-time Watch & Performance (v0.4.0) 📋 **PLANNED**
**Status**: Not started

**Goal**: Replace polling with efficient Kubernetes Watch API and improve performance

Planned Features:
- [ ] Kubernetes Watch API integration
  - [ ] Real-time resource updates via watch streams
  - [ ] Replace 5-second polling with event-driven updates
  - [ ] Watch reconnection on failure
  - [ ] Resource version tracking
- [ ] Performance optimizations
  - [ ] Efficient diff-based UI updates
  - [ ] Lazy loading for large resource lists
  - [ ] Virtual scrolling for 1000+ items
  - [ ] Memory usage optimizations
- [ ] Connection management
  - [ ] Better connection error handling
  - [ ] Auto-reconnect with backoff
  - [ ] Cluster connection status indicator

**Branch**: Will be developed on `dev` branch

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
| Phase 2 - Core Features | v0.2.0 | ✅ Complete | `dev` | ❌ No (ready for PR) |
| Phase 3 - Observability | v0.3.0 | 📋 Planned | - | ❌ No |
| Phase 4 - Watch API | v0.4.0 | 📋 Planned | - | ❌ No |
| Phase 5 - Configuration | v0.5.0 | 📋 Planned | - | ❌ No |
| Phase 6 - More Resources | v0.6.0 | 📋 Planned | - | ❌ No |
| Phase 7 - Write Ops | v0.7.0 | 📋 Planned | - | ❌ No |

**Current Focus**: Merging Phase 2 to `main` branch
